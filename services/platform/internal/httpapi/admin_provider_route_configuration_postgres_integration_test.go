//go:build postgres_integration

package httpapi

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
	adminproviderroutemigrations "radishmind.local/services/platform/migrations/admin_provider_routes"
)

func TestAdminProviderRoutePostgresLifecycleRestartCASAndRuntimeRole(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeURL := postgresIntegrationDatabaseURLForCredentials(
		t,
		runtimeUser,
		os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	adminPool, err := adminproviderroutemigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open admin provider route migration pool: %v", err)
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, adminPool)
	if _, err := adminproviderroutemigrations.RollbackForDevTest(ctx, adminPool); err != nil {
		t.Fatalf("reset admin provider route migration: %v", err)
	}
	preparePostgresIntegrationRuntimeRole(t, ctx, adminPool, runtimeUser)
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		_, _ = adminproviderroutemigrations.RollbackForDevTest(cleanupContext, adminPool)
		adminPool.Close()
	})
	state, err := adminproviderroutemigrations.Apply(ctx, adminPool)
	if err != nil || state.MigrationState != adminproviderroutemigrations.MigrationStateApplied {
		t.Fatalf("apply admin provider route migration: %#v %v", state, err)
	}

	runtimePool, err := adminproviderroutemigrations.OpenPool(ctx, runtimeURL)
	if err != nil {
		t.Fatalf("open admin provider route runtime pool: %v", err)
	}
	repository := newPostgresAdminProviderRouteRepository(runtimePool)
	primaryProfile := adminProviderRouteBridgeProfile()
	secondaryProfile := primaryProfile
	secondaryProfile.Profile = "mock-secondary"
	secondaryProfile.NormalizedProfile = "mock-secondary"
	routeBridge := &fakeBridge{inventory: bridge.ProviderInventory{
		Profiles: []bridge.ProviderProfileDescription{primaryProfile, secondaryProfile},
	}}
	resolver := bridgeAdminProviderInventoryResolver{bridge: routeBridge}
	service := newAdminProviderRouteService(repository, resolver)
	requestContext := adminProviderRouteTestContext()
	requestContext.RequestContext = ctx

	adminProviderRoutePrepareCandidate(t, service, requestContext, "candidate-postgres-one")
	if reviewed := service.ReviewCandidate(
		requestContext,
		"gateway-default",
		"candidate-postgres-one",
		AdminProviderRouteReviewInput{
			ExpectedReviewVersion: 0,
			Decision:              adminProviderRouteDecisionApprove,
			Reason:                "Approve the first PostgreSQL provider route candidate.",
		},
	); reviewed.FailureCode != "" {
		t.Fatalf("approve first PostgreSQL candidate: %#v", reviewed)
	}
	if activated := service.Activate(requestContext, AdminProviderRouteActivationInput{
		ConfigurationID:    "gateway-default",
		CandidateID:        "candidate-postgres-one",
		ExpectedGeneration: 0,
		Action:             adminProviderRouteActionActivate,
		Reason:             "Activate the first PostgreSQL provider route candidate.",
	}); activated.FailureCode != "" || activated.Snapshot == nil || activated.Snapshot.Generation != 1 {
		t.Fatalf("activate first PostgreSQL candidate: %#v", activated)
	}

	secondDraftInput := adminProviderRouteTestDraftInput(1, "mock-secondary")
	secondDraftInput.DisplayName = "PostgreSQL replacement routing"
	secondDraft := service.PutDraft(requestContext, secondDraftInput)
	if secondDraft.FailureCode != "" || secondDraft.Draft == nil || secondDraft.Draft.DraftRevision != 2 {
		t.Fatalf("save second PostgreSQL draft: %#v", secondDraft)
	}
	adminProviderRoutePrepareCandidateWithRevision(
		t,
		service,
		requestContext,
		"candidate-postgres-two",
		secondDraft.Draft.DraftRevision,
	)
	if reviewed := service.ReviewCandidate(
		requestContext,
		"gateway-default",
		"candidate-postgres-two",
		AdminProviderRouteReviewInput{
			ExpectedReviewVersion: 0,
			Decision:              adminProviderRouteDecisionApprove,
			Reason:                "Approve the replacement PostgreSQL provider route candidate.",
		},
	); reviewed.FailureCode != "" {
		t.Fatalf("approve second PostgreSQL candidate: %#v", reviewed)
	}

	const writers = 8
	start := make(chan struct{})
	results := make(chan AdminProviderRouteResult, writers)
	var waitGroup sync.WaitGroup
	for index := 0; index < writers; index++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			results <- service.Activate(requestContext, AdminProviderRouteActivationInput{
				ConfigurationID:    "gateway-default",
				CandidateID:        "candidate-postgres-two",
				ExpectedGeneration: 1,
				Action:             adminProviderRouteActionActivate,
				Reason:             "Concurrently activate the replacement PostgreSQL route.",
			})
		}()
	}
	close(start)
	waitGroup.Wait()
	close(results)
	successes := 0
	conflicts := 0
	for result := range results {
		switch result.FailureCode {
		case "":
			successes++
		case AdminProviderRouteFailureGenerationConflict:
			conflicts++
		default:
			t.Fatalf("unexpected PostgreSQL activation result: %#v", result)
		}
	}
	if successes != 1 || conflicts != writers-1 {
		t.Fatalf("PostgreSQL activation CAS: successes=%d conflicts=%d", successes, conflicts)
	}
	if _, err := runtimePool.Exec(ctx, "CREATE TABLE admin_provider_route_runtime_ddl_denied(id integer)"); err == nil {
		t.Fatal("admin provider route runtime role unexpectedly received DDL permission")
	}
	if _, err := runtimePool.Exec(ctx, `UPDATE admin_provider_route_activation_records
        SET action=action WHERE tenant_ref=$1 AND workspace_id=$2 AND environment=$3 AND configuration_id=$4`,
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.Environment, "gateway-default"); err == nil {
		t.Fatal("PostgreSQL append-only activation history accepted update")
	}

	var persistedPayloads string
	if err := runtimePool.QueryRow(ctx, `SELECT
        COALESCE(string_agg(sanitized_candidate_payload::text, ''), '') ||
        COALESCE((SELECT string_agg(sanitized_snapshot_payload::text, '') FROM admin_provider_route_active_snapshots), '') ||
        COALESCE((SELECT string_agg(sanitized_activation_payload::text, '') FROM admin_provider_route_activation_records), '')
        FROM admin_provider_route_candidates`).Scan(&persistedPayloads); err != nil {
		t.Fatalf("read PostgreSQL payloads for privacy scan: %v", err)
	}
	for _, forbidden := range []string{
		"sk-admin-provider-route-secret",
		"https://provider.example.invalid/v1",
		"Authorization: Bearer",
	} {
		if strings.Contains(persistedPayloads, forbidden) {
			t.Fatalf("PostgreSQL payload contains forbidden material %q", forbidden)
		}
	}

	runtimePool.Close()
	if _, err := repository.ReadDraft(requestContext, "gateway-default"); !errors.Is(err, errAdminProviderRouteStoreUnavailable) {
		t.Fatalf("closed PostgreSQL repository must not fall back: %v", err)
	}
	reopened, err := adminproviderroutemigrations.OpenPool(ctx, runtimeURL)
	if err != nil {
		t.Fatalf("reopen admin provider route runtime pool: %v", err)
	}
	defer reopened.Close()
	restartedService := newAdminProviderRouteService(
		newPostgresAdminProviderRouteRepository(reopened),
		resolver,
	)
	snapshot := restartedService.ReadActiveSnapshot(requestContext, "gateway-default")
	history := restartedService.ListActivations(requestContext, "gateway-default")
	candidate := restartedService.ReadCandidate(
		requestContext,
		"gateway-default",
		"candidate-postgres-two",
	)
	if snapshot.FailureCode != "" || snapshot.Snapshot == nil ||
		snapshot.Snapshot.Generation != 2 || snapshot.Snapshot.CandidateID != "candidate-postgres-two" {
		t.Fatalf("restore PostgreSQL active snapshot: %#v", snapshot)
	}
	if history.FailureCode != "" || len(history.ActivationHistory) != 2 {
		t.Fatalf("restore PostgreSQL activation history: %#v", history)
	}
	if candidate.FailureCode != "" || candidate.Candidate == nil ||
		candidate.Candidate.CandidateState != adminProviderRouteCandidateApproved ||
		candidate.Candidate.Review == nil {
		t.Fatalf("restore PostgreSQL candidate and review: %#v", candidate)
	}
	gatewayServer := &Server{
		bridge: routeBridge,
		config: config.Config{
			GatewayProviderRouteSource:          "admin_snapshot_dev_test",
			GatewayProviderRouteEnvironment:     "test",
			GatewayProviderRouteConfigurationID: "gateway-default",
		},
		providerRouteSnapshotProvider: adminProviderRouteSnapshotProvider{
			repository: newPostgresAdminProviderRouteRepository(reopened),
		},
	}
	gatewayContext := GatewayRequestContext{
		RequestContext: ctx, TenantRef: requestContext.TenantRef, WorkspaceID: requestContext.WorkspaceID,
		ConsumerRef: "api_key:key_aaaaaaaaaaaaaaaa", ApplicationID: "app-postgres-route",
		SubjectRef: requestContext.ActorRef, AuditContext: "api-key-dev-test",
		Source: gatewayAPIKeyAuthenticationSource, RequestID: "request-postgres-route-consumer",
		AuditRef: "audit-postgres-route-consumer",
	}
	selection, failure := gatewayServer.resolveGatewayNorthboundSelection(
		ctx, gatewayContext, northboundProtocolChatCompletions, "mock-model", nil,
	)
	if failure != "" || selection.routeGeneration != 2 ||
		selection.providerProfile != "mock-secondary" ||
		selection.routeSnapshotDigest != snapshot.Snapshot.SnapshotDigest {
		t.Fatalf("consume PostgreSQL active snapshot through Gateway: selection=%#v failure=%s", selection, failure)
	}

	httpFixture := newAdminProviderRouteHTTPFixture()
	httpFixture.server.adminProviderRouteRepository = newPostgresAdminProviderRouteRepository(reopened)
	httpDraftInput := adminProviderRouteTestDraftInput(0, "mock-primary")
	httpDraft := httpFixture.serve(
		t,
		"PUT",
		"/v1/admin/provider-route-configurations/gateway-http",
		adminProviderRouteDraftPutBody{
			ExpectedRevision: httpDraftInput.ExpectedRevision,
			DisplayName:      "PostgreSQL HTTP routing",
			ProviderProfiles: httpDraftInput.ProviderProfiles,
			ModelRoutes:      httpDraftInput.ModelRoutes,
		},
		httpFixture.auth,
		200,
	)
	if httpDraft.Draft == nil || httpDraft.Draft.DraftRevision != 1 {
		t.Fatalf("create PostgreSQL draft through Admin HTTP: %#v", httpDraft)
	}
	httpFixture.serve(
		t,
		"POST",
		"/v1/admin/provider-route-configurations/gateway-http/candidates",
		adminProviderRouteCandidateCreateBody{
			CandidateID:           "candidate-http",
			ExpectedDraftRevision: 1,
		},
		httpFixture.auth,
		201,
	)
	httpFixture.serve(
		t,
		"POST",
		"/v1/admin/provider-route-configurations/gateway-http/candidates/candidate-http/reviews",
		adminProviderRouteReviewBody{
			ExpectedReviewVersion: 0,
			Decision:              "approve",
			Reason:                "Approve the PostgreSQL Admin HTTP candidate.",
		},
		httpFixture.auth,
		200,
	)
	httpActivation := httpFixture.serve(
		t,
		"POST",
		"/v1/admin/provider-route-configurations/gateway-http/candidates/candidate-http/activations",
		adminProviderRouteActivationBody{
			ExpectedGeneration: 0,
			Action:             "activate",
			Reason:             "Activate the PostgreSQL Admin HTTP candidate.",
		},
		httpFixture.auth,
		200,
	)
	if httpActivation.Snapshot == nil || httpActivation.Snapshot.Generation != 1 {
		t.Fatalf("activate PostgreSQL candidate through Admin HTTP: %#v", httpActivation)
	}
	httpHistory := httpFixture.serve(
		t,
		"GET",
		"/v1/admin/provider-route-configurations/gateway-http/activation-history",
		nil,
		httpFixture.auth,
		200,
	)
	if len(httpHistory.ActivationHistory) != 1 {
		t.Fatalf("restore PostgreSQL Admin HTTP activation history: %#v", httpHistory)
	}
}
