package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/developmentcontrol"
	"radishmind.local/services/platform/internal/sqlitedev"
	"radishmind.local/services/platform/internal/workspacepolicy"
	migrations "radishmind.local/services/platform/migrations/sqlite"
)

type developmentFixtureBackend struct{}

func (developmentFixtureBackend) Policy(context.Context, string) (developmentcontrol.NativePolicy, error) {
	return developmentcontrol.NativePolicy{InputTokenCeiling: 100, MaxOutputTokens: 50, Timeout: time.Second}, nil
}
func (developmentFixtureBackend) Complete(context.Context, string, developmentcontrol.NativeRequest) (developmentcontrol.NativeResult, error) {
	return developmentcontrol.NativeResult{Message: developmentcontrol.NativeMessage{Role: "assistant", Content: "fixed native transport response"}, Usage: &developmentcontrol.NativeUsage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15}, FinishReason: "stop"}, nil
}

type developmentServerFixture struct {
	server             *Server
	identity           developmentcontrol.Identity
	plan               developmentcontrol.Plan
	request            *http.Request
	identityRepository *sqliteLocalIdentityRepository
	pricingContext     GatewayModelPricingContext
	role               LocalRoleAssignment
	membershipID       string
	runtime            *developmentcontrol.Runtime
	bridge             *fakeBridge
}

func newDevelopmentServerFixture(t *testing.T) *developmentServerFixture {
	return newDevelopmentServerFixtureWithConnector(t, func(s *Server) (*developmentcontrol.Runtime, error) {
		return s.ConnectDevelopmentControl(developmentFixtureBackend{})
	})
}

func newDevelopmentServerFixtureWithConnector(t *testing.T, connect func(*Server) (*developmentcontrol.Runtime, error)) *developmentServerFixture {
	t.Helper()
	ctx := context.Background()
	now := time.Now().UTC()
	db, err := sqlitedev.Open(ctx, sqlitedev.Options{DatabasePath: filepath.Join(t.TempDir(), "platform.db"), Migrations: migrations.Migrations()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := newSQLiteLocalIdentityRepository(db.DB())
	httpFixture := newLocalIdentityHTTPTestFixture(repo, now)
	_, cookies, document := registerLocalIdentityHTTPTestAccount(t, httpFixture, "development-owner@example.com", "synthetic test password only")
	userID := document.Account.UserID
	administration := newLocalIdentityAdministrationService(repo)
	bootstrap, err := administration.BootstrapWorkspaceAdministrator(ctx, LocalIdentityBootstrapWorkspaceAdministratorInput{TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", UserID: userID, AuditRef: "audit:development-bootstrap"})
	if err != nil {
		t.Fatal("bootstrap", err)
	}
	member, role := bootstrap.Membership, bootstrap.RoleAssignment
	s := httpFixture.server
	s.localPersistenceRuntime = db
	s.localIdentityRepository = repo
	s.applicationCatalogRepository = newSQLiteApplicationCatalogRepository(db.DB())
	s.gatewayModelPricingRepository = newSQLiteGatewayModelPricingRepository(db.DB())
	actor := "user:" + userID
	appContext := ApplicationCatalogContext{RequestContext: ctx, RequestID: "request_development", TenantRef: member.TenantRef, WorkspaceID: member.WorkspaceID, ActorRef: actor, OwnerSubjectRef: actor, AuditRef: "audit:development", WriteEnabled: true}
	appService := newApplicationCatalogService(s.applicationCatalogRepository)
	appService.newID = func() (string, error) { return "app_aaaaaaaaaaaaaaaa", nil }
	created := appService.Create(appContext, ApplicationCatalogCreateInput{DisplayName: "Synthetic development app", ApplicationKind: "workflow_copilot"})
	if created.FailureCode != "" || created.Record == nil {
		t.Fatal(created)
	}
	appID := created.Record.ApplicationID
	releaseContext := WorkflowDefinitionReleaseContext{RequestContext: ctx, TenantRef: member.TenantRef, WorkspaceID: member.WorkspaceID, ApplicationID: appID, OwnerSubjectRef: actor, ActorRef: actor, RequestID: "request_development", AuditRef: "audit:development"}
	draft := executableWorkflowDraftForTest()
	draft.ApplicationID = appID
	draft.ToolRefs = []string{}
	draft.RAGRefs = []string{}
	draft.RequestedCapabilities = []string{}
	release := newWorkflowDefinitionReleaseStore()
	candidate, err := release.CreateCandidate(releaseContext, "candidate_development", "definition_development", "", draft, now)
	if err != nil {
		t.Fatal(err)
	}
	_, version, err := release.Review(releaseContext, candidate.CandidateID, 0, "approve", "synthetic authority fixture", candidate.SourceDraftDigest, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = release.Activate(releaseContext, version.DefinitionID, 0, version.Version, now); err != nil {
		t.Fatal(err)
	}
	s.workflowDefinitionReleaseRepository = release
	profile := bridge.ProviderProfileDescription{Profile: "development", ProviderID: "openai-compatible", ResolvedModel: "synthetic-model", APIStyle: "openai-compatible", Enabled: true, CredentialState: "configured", HasAPIKey: true, HasBaseURL: true}
	b := &fakeBridge{inventory: bridge.ProviderInventory{Profiles: []bridge.ProviderProfileDescription{profile}}}
	s.bridge = b
	pricingContext := GatewayModelPricingContext{RequestContext: ctx, TenantRef: member.TenantRef, WorkspaceID: member.WorkspaceID, Environment: "development", ProviderID: profile.ProviderID, ProfileID: profile.Profile, ModelID: profile.ResolvedModel, ActorRef: actor, RequestID: "request_development", AuditRef: "audit:development"}
	if _, err = s.gatewayModelPricingRepository.PutRevision(pricingContext, GatewayModelPricingPolicyInput{Currency: "USD", InputPriceMicrosPerTokenUnit: 1000000, OutputPriceMicrosPerTokenUnit: 2000000, Reason: "synthetic fixed price"}, now); err != nil {
		t.Fatal(err)
	}
	runtime, err := connect(s)
	if err != nil {
		t.Fatal(err)
	}
	initialBytes, err := json.Marshal([]developmentcontrol.NativeMessage{{Role: "user", Content: "synthetic task"}})
	if err != nil {
		t.Fatal(err)
	}
	plan := developmentcontrol.Plan{InitialMessagesDigest: developmentcontrol.PayloadDigest(initialBytes), SchemaVersion: "development_attempt_plan.v1", ApplicationID: appID, ExecutionProfile: workflowDefinitionExecutorProfile, DefinitionID: version.DefinitionID, ModelRef: "profile:development", BaselineDigest: developmentcontrol.PayloadDigest([]byte("synthetic baseline")), Tools: developmentcontrol.ToolPolicy{ReadPaths: []string{"main.go"}, WritePaths: []string{"main.go"}, VerificationProfile: "go_test_v1"}, Limits: developmentcontrol.Limits{MaxModelCalls: 2, MaxToolCalls: 2, MaxTokens: 1000, MaxCostMicrousd: 10000}, LifetimeMS: 60000, MaxOutputTokens: 50}
	// Capture only the request already verified by the real local-session middleware.
	var verified *http.Request
	capture := withLocalIdentitySessionAuthentication(withControlPlaneReadAuthenticator(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { verified = r }), &controlPlaneReadAuthenticator{mode: localIdentityAuthMode}), httpFixture.service)
	req := httptest.NewRequest("POST", "http://localhost/trusted-host-confirmation", nil)
	addLocalIdentityCookies(req, cookies)
	req.Header.Set(localIdentityActiveTenantHeader, member.TenantRef)
	req.Header.Set(activeWorkspaceHeader, member.WorkspaceID)
	capture.ServeHTTP(httptest.NewRecorder(), req)
	if verified == nil {
		t.Fatal("verified request missing")
	}
	t.Cleanup(s.Close)
	if plan.Validate() != nil {
		t.Fatalf("invalid fixture plan: %+v", plan)
	}
	if _, err := repo.AuthorizeWorkspace(ctx, userID, member.TenantRef, member.WorkspaceID, []string{"applications:read", workspacepolicy.PermissionRolesAssign}, time.Now()); err != nil {
		t.Fatal("identity qualification", err)
	}
	_, failure := s.applicationInteractionAuthorityResolver().Resolve(ApplicationInteractionContext{RequestContext: ctx, RequestID: "development-qualification", TenantRef: member.TenantRef, WorkspaceID: member.WorkspaceID, ApplicationID: appID, ActorRef: actor, OwnerSubjectRef: actor, AuditRef: "audit:development-qualification", WriteEnabled: true}, ApplicationInteractionProfileBinding{ExecutionProfile: plan.ExecutionProfile, DefinitionID: plan.DefinitionID})
	if failure != "" {
		t.Fatal("application qualification", failure)
	}
	return &developmentServerFixture{s, developmentcontrol.Identity{TenantRef: member.TenantRef, AccountID: userID, WorkspaceID: member.WorkspaceID}, plan, verified, repo, pricingContext, role, member.MembershipID, runtime, b}
}

func TestDevelopmentTrustedIssuerAndDedicatedListener(t *testing.T) {
	f := newDevelopmentServerFixture(t)
	prepared, err := f.server.PrepareDevelopmentAttempt(f.request, f.plan)
	if err != nil {
		t.Fatal(err)
	}
	issued, err := f.server.IssueDevelopmentAttempt(f.request, f.plan, "confirmation_development", prepared)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.server.IssueDevelopmentAttempt(f.request, f.plan, "confirmation_development", prepared); !errors.Is(err, developmentcontrol.ErrConflict) {
		t.Fatalf("confirmation reused: %v", err)
	}
	bare := httptest.NewRequest("POST", "http://localhost", nil)
	bare.Header.Set("Authorization", "Bearer long-lived-api-key")
	if _, err = f.server.PrepareDevelopmentAttempt(bare, f.plan); !errors.Is(err, developmentcontrol.ErrCredential) {
		t.Fatalf("unverified issuer accepted: %v", err)
	}
	url, done, err := f.server.StartDevelopmentChannel(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = f.server.StartDevelopmentChannel(context.Background()); !errors.Is(err, developmentcontrol.ErrConflict) {
		t.Fatal(err)
	}
	tools, err := developmentcontrol.NativeTools()
	if err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(developmentcontrol.NativeRequest{Model: prepared.Selection.WireModel, Messages: []developmentcontrol.NativeMessage{{Role: "user", Content: "synthetic task"}}, MaxCompletionTokens: 50, Tools: tools})
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest("POST", url+"/internal/development/v1/attempts/"+issued.Delegation.Scope.AttemptID+"/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+issued.Credential)
	req.Header.Set("X-RadishMind-Development-Request-ID", "model_fixture")
	client := &http.Client{Timeout: 3 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil || response.StatusCode != 200 {
		t.Fatalf("dedicated native response: %s %s %v", response.Status, payload, err)
	}
	if f.bridge.lastRequest != nil {
		t.Fatal("advisory bridge invoked")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err = f.server.CloseDevelopmentControl(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("listener remained running")
	}
	if _, err = client.Do(req); err == nil {
		t.Fatal("listener still accepts connections")
	}
}

func TestDevelopmentCurrentOwnerQualifications(t *testing.T) {
	for _, mode := range []string{"price", "member", "role", "model", "application", "scope"} {
		t.Run(mode, func(t *testing.T) {
			f := newDevelopmentServerFixture(t)
			reader := developmentQualificationReader{f.server}
			original, err := reader.Resolve(context.Background(), f.identity, f.plan)
			if err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "price":
				_, err = f.server.gatewayModelPricingRepository.PutRevision(f.pricingContext, GatewayModelPricingPolicyInput{ExpectedVersion: 1, Currency: "USD", InputPriceMicrosPerTokenUnit: 2000000, OutputPriceMicrosPerTokenUnit: 2000000, Reason: "changed synthetic price"}, time.Now())
			case "member":
				_, err = revokeDevelopmentFixtureMembership(t, f)
			case "role":
				_, err = revokeDevelopmentFixtureRole(t, f)
			case "model":
				f.bridge.inventory.Profiles[0].Enabled = false
			case "application":
				f.plan.ApplicationID = "app_missingfixture"
			case "scope":
				f.identity.WorkspaceID = "workspace_other"
			}
			if err != nil {
				t.Fatal(err)
			}
			current, err := reader.Resolve(context.Background(), f.identity, f.plan)
			if mode == "price" {
				if err != nil || current == original || current.PricingDigest == original.PricingDigest {
					t.Fatal("price owner revision not observed", err)
				}
			} else if !errors.Is(err, developmentcontrol.ErrIneligible) {
				t.Fatalf("current owner denial missing: %v", err)
			}
		})
	}
}

func developmentFixtureSecondAdministrator(t *testing.T, f *developmentServerFixture) (*localIdentityAdministrationService, LocalIdentityAdministrationActor) {
	t.Helper()
	ctx := context.Background()
	now := time.Now().UTC()
	account, credential := localIdentityTestAccount("usr_bbbbbbbbbbbbbbbb", "cred_bbbbbbbbbbbbbbbb", "second-admin@example.com", now)
	if err := f.identityRepository.CreateAccount(ctx, account, credential); err != nil {
		t.Fatal(err)
	}
	service := newLocalIdentityAdministrationService(f.identityRepository)
	owner := LocalIdentityAdministrationActor{UserID: f.identity.AccountID, TenantRef: f.identity.TenantRef, WorkspaceID: f.identity.WorkspaceID, AuthenticatedAt: now}
	if _, err := service.CreateWorkspaceMembership(ctx, owner, LocalIdentityCreateWorkspaceMembershipInput{TenantRef: f.identity.TenantRef, WorkspaceID: f.identity.WorkspaceID, UserID: account.UserID, AuditRef: "audit:development-second-admin"}); err != nil {
		t.Fatal(err)
	}
	definition, _ := workspacepolicy.BuiltInRole(workspacepolicy.RoleWorkspaceAdmin)
	if _, err := service.AssignWorkspaceRole(ctx, owner, LocalIdentityAssignWorkspaceRoleInput{TenantRef: f.identity.TenantRef, WorkspaceID: f.identity.WorkspaceID, UserID: account.UserID, RoleKey: definition.RoleKey, ExpectedCatalogVersion: definition.CatalogVersion, ExpectedRoleDefinitionDigest: definition.DefinitionDigest, AuditRef: "audit:development-second-admin"}); err != nil {
		t.Fatal(err)
	}
	return service, LocalIdentityAdministrationActor{UserID: account.UserID, TenantRef: f.identity.TenantRef, WorkspaceID: f.identity.WorkspaceID, AuthenticatedAt: now}
}
func revokeDevelopmentFixtureMembership(t *testing.T, f *developmentServerFixture) (LocalIdentityWorkspaceMembershipRevocation, error) {
	service, actor := developmentFixtureSecondAdministrator(t, f)
	return service.RevokeWorkspaceMembership(context.Background(), actor, LocalIdentityRevokeWorkspaceMembershipInput{TenantRef: f.identity.TenantRef, WorkspaceID: f.identity.WorkspaceID, MembershipID: f.membershipID, ExpectedVersion: 1, Confirmed: true, AuditRef: "audit:development-revoke"})
}
func revokeDevelopmentFixtureRole(t *testing.T, f *developmentServerFixture) (LocalRoleAssignment, error) {
	service, actor := developmentFixtureSecondAdministrator(t, f)
	return service.RevokeWorkspaceRole(context.Background(), actor, LocalIdentityRevokeWorkspaceRoleInput{TenantRef: f.identity.TenantRef, WorkspaceID: f.identity.WorkspaceID, AssignmentID: f.role.AssignmentID, ExpectedVersion: 1, Confirmed: true, AuditRef: "audit:development-revoke"})
}
