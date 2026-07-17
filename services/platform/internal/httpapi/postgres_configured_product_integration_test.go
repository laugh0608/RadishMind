//go:build postgres_integration

package httpapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"radishmind.local/services/platform/internal/config"
	apikeymigrations "radishmind.local/services/platform/migrations/api_key_records"
	applicationcatalogmigrations "radishmind.local/services/platform/migrations/application_catalog_records"
	applicationdraftmigrations "radishmind.local/services/platform/migrations/application_configuration_drafts"
	applicationpublishmigrations "radishmind.local/services/platform/migrations/application_publish_candidates"
	gatewayrequestmigrations "radishmind.local/services/platform/migrations/gateway_requests"
	workflowrunmigrations "radishmind.local/services/platform/migrations/workflow_runs"
	workflowdraftmigrations "radishmind.local/services/platform/migrations/workflow_saved_drafts"
)

type postgresMigrationGate struct {
	name     string
	table    string
	apply    func(context.Context) (string, string, error)
	inspect  func(context.Context) (string, error)
	rollback func(context.Context) (string, error)
}

type postgresMigrationResult struct {
	state    string
	checksum string
	err      error
}

func TestPostgresConfiguredMigrationRoleSchemaAndConcurrencyGate(t *testing.T) {
	adminPool, runtimeDatabaseURL, ctx := newConfiguredPostgresTestDatabase(t)
	gates := configuredPostgresMigrationGates(adminPool)
	for _, gate := range gates {
		t.Run(gate.name, func(t *testing.T) {
			runConfiguredPostgresMigrationGate(t, ctx, adminPool, gate)
		})
	}

	assertConfiguredPostgresSchema(t, ctx, adminPool)
	assertPostgresRuntimeIsolationAndConnections(t, ctx, adminPool, runtimeDatabaseURL)
}

func TestPostgresConfiguredProductChainRestartRacesAndNoFallback(t *testing.T) {
	adminPool, runtimeDatabaseURL, ctx := newConfiguredPostgresTestDatabase(t)
	for _, gate := range configuredPostgresMigrationGates(adminPool) {
		state, _, err := gate.apply(ctx)
		if err != nil || state != "applied" {
			t.Fatalf("apply %s migration for configured product chain: state=%s err=%v", gate.name, state, err)
		}
	}

	cfg := configuredPostgresProductConfig(runtimeDatabaseURL)
	server, err := NewServerWithError(cfg, Options{BuildVersion: "postgres-configured-first"})
	if err != nil {
		t.Fatalf("start configured PostgreSQL platform: %v", err)
	}
	server.bridge = &workflowExecutorTestBridge{}
	assertConfiguredPostgresRepositorySelection(t, server)

	owner := "subject_platform_ops"
	catalogContext := applicationCatalogTestContext(owner)
	catalogService := newApplicationCatalogService(server.applicationCatalogRepository)
	catalogService.now = func() time.Time { return time.Date(2026, 7, 14, 9, 0, 0, 0, time.UTC) }
	applicationIDs := []string{"app_aaaaaaaaaaaaaaaa", "app_bbbbbbbbbbbbbbbb", "app_cccccccccccccccc"}
	nextApplicationID := 0
	catalogService.newID = func() (string, error) {
		identifier := applicationIDs[nextApplicationID]
		nextApplicationID++
		return identifier, nil
	}
	applications := make([]ApplicationCatalogRecord, 0, len(applicationIDs))
	for index := range applicationIDs {
		created := catalogService.Create(catalogContext, ApplicationCatalogCreateInput{
			DisplayName: fmt.Sprintf("PostgreSQL configured app %d", index+1),
			Description: "Configured PostgreSQL product gate application.", ApplicationKind: "workflow_copilot",
		})
		if created.FailureCode != "" || created.Record == nil {
			t.Fatalf("create configured PostgreSQL application %d: %#v", index, created)
		}
		applications = append(applications, *created.Record)
	}
	firstApplications := catalogService.List(catalogContext, ApplicationCatalogListInput{Limit: 2, ApplicationKind: "workflow_copilot"})
	if firstApplications.FailureCode != "" || len(firstApplications.Records) != 2 || firstApplications.NextCursor == nil ||
		firstApplications.Records[0].ApplicationID != applicationIDs[2] || firstApplications.Records[1].ApplicationID != applicationIDs[1] {
		t.Fatalf("application catalog PostgreSQL tuple pagination first page drifted: %#v", firstApplications)
	}
	secondApplications := catalogService.List(catalogContext, ApplicationCatalogListInput{
		Limit: 2, ApplicationKind: "workflow_copilot", Cursor: *firstApplications.NextCursor,
	})
	if secondApplications.FailureCode != "" || len(secondApplications.Records) != 1 || secondApplications.NextCursor != nil ||
		secondApplications.Records[0].ApplicationID != applicationIDs[0] {
		t.Fatalf("application catalog PostgreSQL tuple pagination second page drifted: %#v", secondApplications)
	}
	primaryApplication := applications[0]

	draftContext := validApplicationDraftContext()
	draftContext.ApplicationID = primaryApplication.ApplicationID
	draftContext.TenantRef = catalogContext.TenantRef
	draftContext.WorkspaceID = catalogContext.WorkspaceID
	draftContext.ActorRef = owner
	draftContext.OwnerSubjectRef = owner
	draftPayload := validApplicationDraftPayload()
	draftPayload.DraftID = "app-config-postgres-gate"
	draftPayload.ApplicationID = primaryApplication.ApplicationID
	draftPayload.WorkspaceID = catalogContext.WorkspaceID
	draftPayload.BaseApplicationUpdatedAt = primaryApplication.UpdatedAt
	applicationDraft := server.applicationConfigurationDraftService().Save(draftContext, draftPayload, 0)
	if applicationDraft.FailureCode != "" || applicationDraft.Draft == nil {
		t.Fatalf("save configured PostgreSQL application draft: %#v", applicationDraft)
	}

	publishContext := validApplicationPublishContext()
	publishContext.ApplicationID = primaryApplication.ApplicationID
	publishContext.TenantRef = catalogContext.TenantRef
	publishContext.WorkspaceID = catalogContext.WorkspaceID
	publishContext.ActorRef = owner
	publishContext.OwnerSubjectRef = owner
	candidateID := "candidate-postgres-configured-gate"
	publishService := server.applicationPublishCandidateService()
	publishCandidate := publishService.Create(publishContext, ApplicationPublishCreateInput{
		CandidateID: candidateID, DraftID: draftPayload.DraftID, ExpectedDraftVersion: 1,
	})
	if publishCandidate.FailureCode != "" || publishCandidate.Candidate == nil {
		t.Fatalf("create configured PostgreSQL publish candidate: %#v", publishCandidate)
	}
	approvedCandidate := publishService.Review(publishContext, candidateID, ApplicationPublishReviewInput{
		ExpectedReviewVersion: 0, Decision: applicationPublishDecisionApprove,
		Reason: "PostgreSQL configured product evidence reviewed and approved.",
	})
	if approvedCandidate.FailureCode != "" || approvedCandidate.Candidate == nil ||
		approvedCandidate.Candidate.CandidateState != applicationPublishStateApproved {
		t.Fatalf("approve configured PostgreSQL publish candidate: %#v", approvedCandidate)
	}

	apiKeyContext := apiKeyTestContext(owner)
	apiKeyContext.TenantRef = catalogContext.TenantRef
	apiKeyContext.WorkspaceID = catalogContext.WorkspaceID
	keyService := newAPIKeyService(server.apiKeyRepository, server.applicationCatalogRepository)
	keyService.now = func() time.Time { return time.Date(2026, 7, 14, 9, 5, 0, 0, time.UTC) }
	apiKeyIDs := []string{"key_aaaaaaaaaaaaaaaa", "key_bbbbbbbbbbbbbbbb", "key_cccccccccccccccc"}
	nextAPIKeyID := 0
	keyService.newID = func() (string, error) {
		identifier := apiKeyIDs[nextAPIKeyID]
		nextAPIKeyID++
		return identifier, nil
	}
	issuedKeys := make([]APIKeyResult, 0, len(apiKeyIDs))
	for index := range apiKeyIDs {
		issued := keyService.Create(apiKeyContext, APIKeyCreateInput{
			ApplicationID: primaryApplication.ApplicationID,
			DisplayName:   fmt.Sprintf("PostgreSQL configured key %d", index+1),
			Scopes:        []string{"chat:invoke", "models:read"}, ExpiresInDays: 30,
		})
		if issued.FailureCode != "" || issued.Record == nil || issued.CredentialToken == "" {
			t.Fatalf("issue configured PostgreSQL API key %d: failure=%s record_present=%t token_present=%t",
				index, issued.FailureCode, issued.Record != nil, issued.CredentialToken != "")
		}
		issuedKeys = append(issuedKeys, issued)
	}
	firstKeys := keyService.List(apiKeyContext, APIKeyListInput{ApplicationID: primaryApplication.ApplicationID, Limit: 2})
	if firstKeys.FailureCode != "" || len(firstKeys.Records) != 2 || firstKeys.NextCursor == nil ||
		firstKeys.Records[0].APIKeyID != apiKeyIDs[2] || firstKeys.Records[1].APIKeyID != apiKeyIDs[1] {
		t.Fatalf("API key PostgreSQL tuple pagination first page drifted: %#v", firstKeys)
	}
	secondKeys := keyService.List(apiKeyContext, APIKeyListInput{
		ApplicationID: primaryApplication.ApplicationID, Limit: 2, Cursor: *firstKeys.NextCursor,
	})
	if secondKeys.FailureCode != "" || len(secondKeys.Records) != 1 || secondKeys.NextCursor != nil ||
		secondKeys.Records[0].APIKeyID != apiKeyIDs[0] {
		t.Fatalf("API key PostgreSQL tuple pagination second page drifted: %#v", secondKeys)
	}

	rawGatewayInput := "private configured PostgreSQL Gateway prompt must not persist"
	gatewayRequestID := "request-postgres-configured-gate"
	gatewayRequest := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(mustLocalProductJSON(t, map[string]any{
		"model": "platform-model", "messages": []map[string]string{{"role": "user", "content": rawGatewayInput}},
	})))
	gatewayRequest.Header.Set("Authorization", "Bearer "+issuedKeys[0].CredentialToken)
	gatewayRequest.Header.Set("X-Request-Id", gatewayRequestID)
	gatewayResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(gatewayResponse, gatewayRequest)
	if gatewayResponse.Code != http.StatusOK {
		t.Fatalf("call configured PostgreSQL Gateway: status=%d body=%s", gatewayResponse.Code, gatewayResponse.Body.String())
	}
	usedKey := keyService.Read(apiKeyContext, apiKeyIDs[0])
	if usedKey.FailureCode != "" || usedKey.Record == nil || usedKey.Record.LastUsedAt == nil || usedKey.Record.RecordVersion != 1 {
		t.Fatalf("configured PostgreSQL authentication did not update last_used_at: %#v", usedKey)
	}
	lastUsedAt, err := time.Parse(time.RFC3339Nano, *usedKey.Record.LastUsedAt)
	if err != nil {
		t.Fatalf("parse configured PostgreSQL last_used_at: %v", err)
	}
	monotonic, err := server.apiKeyRepository.RecordSuccessfulAuthentication(ctx, apiKeyIDs[0], 1, lastUsedAt.Add(-time.Minute))
	if err != nil || monotonic.LastUsedAt == nil || *monotonic.LastUsedAt != *usedKey.Record.LastUsedAt {
		t.Fatalf("configured PostgreSQL last_used_at regressed: record=%#v err=%v", monotonic, err)
	}

	gatewayContext := GatewayRequestContext{
		RequestContext: ctx, TenantRef: catalogContext.TenantRef, WorkspaceID: catalogContext.WorkspaceID,
		ConsumerRef: "api_key:" + apiKeyIDs[0], ApplicationID: primaryApplication.ApplicationID,
		SubjectRef: owner, ScopeGrants: []string{"gateway_requests:read"}, AuditContext: "api-key-dev-test",
		Source: gatewayAPIKeyAuthenticationSource, RequestID: "request-postgres-history-read", AuditRef: "audit-postgres-history-read",
	}
	storedGateway, found, err := server.gatewayRequestHistoryStore.ReadRequest(gatewayContext, gatewayRequestID)
	if err != nil || !found || storedGateway.Status != GatewayRequestStatusSucceeded || storedGateway.ConsumerRef != gatewayContext.ConsumerRef {
		t.Fatalf("configured PostgreSQL Gateway history association failed: found=%t record=%#v err=%v", found, storedGateway, err)
	}
	assertPostgresGatewayTuplePagination(t, server.gatewayRequestHistoryStore, ctx, primaryApplication.ApplicationID)

	savedDraftContext := savedWorkflowDraftSQLiteContext()
	savedDraftContext.TenantRef = catalogContext.TenantRef
	savedDraftContext.WorkspaceID = catalogContext.WorkspaceID
	savedDraftContext.ApplicationID = primaryApplication.ApplicationID
	savedDraftContext.ActorRef = owner
	savedDraftContext.OwnerSubjectRef = owner
	savedDraftPayload := validSavedWorkflowDraftPayload()
	savedDraftPayload.DraftID = "draft_postgres_configured_gate"
	savedDraftPayload.WorkspaceID = catalogContext.WorkspaceID
	savedDraftPayload.ApplicationID = primaryApplication.ApplicationID
	savedDraft := newSavedWorkflowDraftService(server.savedWorkflowDraftStore).SaveDraft(
		savedDraftContext, SaveWorkflowDraftRequest{Payload: savedDraftPayload},
	)
	if savedDraft.FailureCode != "" || savedDraft.Draft == nil {
		t.Fatalf("save configured PostgreSQL workflow draft: %#v", savedDraft)
	}

	runContext := workflowExecutorTestContext()
	runContext.RequestContext = ctx
	runContext.TenantRef = catalogContext.TenantRef
	runContext.WorkspaceID = catalogContext.WorkspaceID
	runContext.ApplicationID = primaryApplication.ApplicationID
	runContext.ActorRef = owner
	runID := "run_postgres_configured_gate"
	runRecord := workflowRunHistoryTestRecord(runContext, runID, savedDraftPayload.DraftID, time.Now().UTC())
	if err := server.workflowRunStore.UpsertRun(runContext, &runRecord); err != nil {
		t.Fatalf("create configured PostgreSQL workflow run: %v", err)
	}
	runRecord.Status = WorkflowRunStatusSucceeded
	runRecord.CompletedAt = workflowRunTimestamp(time.Now().UTC())
	if runRecord.Diagnostic != nil {
		runRecord.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
	}
	if err := server.workflowRunStore.UpsertRun(runContext, &runRecord); err != nil {
		t.Fatalf("complete configured PostgreSQL workflow run: %v", err)
	}

	assertConfiguredPostgresSecretBoundary(t, ctx, adminPool, apiKeyIDs[0], gatewayRequestID,
		issuedKeys[0].CredentialToken, rawGatewayInput)
	assertPostgresAuthenticationRevokeRace(t, server, keyService, apiKeyContext, issuedKeys[1])
	assertPostgresApplicationArchiveRace(t, server, catalogService, catalogContext, primaryApplication.ApplicationID, issuedKeys[2])

	closedCatalogRepository := server.applicationCatalogRepository
	closedDraftRepository := server.applicationDraftRepository
	closedPublishRepository := server.applicationPublishCandidateRepository
	closedAPIKeyRepository := server.apiKeyRepository
	closedSavedDraftStore := server.savedWorkflowDraftStore
	closedRunStore := server.workflowRunStore
	closedGatewayStore := server.gatewayRequestHistoryStore
	server.Close()
	assertConfiguredPostgresStoresFailClosed(t, catalogContext, draftContext, publishContext, apiKeyContext,
		savedDraftContext, runContext, gatewayContext, primaryApplication.ApplicationID, draftPayload.DraftID,
		candidateID, apiKeyIDs[0], savedDraftPayload.DraftID, runID, closedCatalogRepository, closedDraftRepository,
		closedPublishRepository, closedAPIKeyRepository, closedSavedDraftStore, closedRunStore, closedGatewayStore)

	restarted, err := NewServerWithError(cfg, Options{BuildVersion: "postgres-configured-restarted"})
	if err != nil {
		t.Fatalf("restart configured PostgreSQL platform: %v", err)
	}
	t.Cleanup(restarted.Close)
	restarted.bridge = &workflowExecutorTestBridge{}
	assertConfiguredPostgresRepositorySelection(t, restarted)
	if restored := newApplicationCatalogService(restarted.applicationCatalogRepository).Read(catalogContext, primaryApplication.ApplicationID); restored.FailureCode != "" || restored.Record == nil || restored.Record.LifecycleState != applicationCatalogLifecycleArchived {
		t.Fatalf("restore archived configured PostgreSQL application: %#v", restored)
	}
	if restored := restarted.applicationConfigurationDraftService().Read(draftContext, draftPayload.DraftID); restored.FailureCode != "" || restored.Draft == nil || restored.Draft.DraftVersion != 1 {
		t.Fatalf("restore configured PostgreSQL application draft: %#v", restored)
	}
	if restored := restarted.applicationPublishCandidateService().Read(publishContext, candidateID); restored.FailureCode != "" || restored.Candidate == nil || restored.Candidate.CandidateState != applicationPublishStateApproved {
		t.Fatalf("restore configured PostgreSQL publish candidate: %#v", restored)
	}
	if restored := newAPIKeyService(restarted.apiKeyRepository, restarted.applicationCatalogRepository).Read(apiKeyContext, apiKeyIDs[0]); restored.FailureCode != "" || restored.Record == nil || restored.Record.LastUsedAt == nil {
		t.Fatalf("restore configured PostgreSQL API key: %#v", restored)
	}
	if restored, found, err := restarted.gatewayRequestHistoryStore.ReadRequest(gatewayContext, gatewayRequestID); err != nil || !found || restored.Status != GatewayRequestStatusSucceeded {
		t.Fatalf("restore configured PostgreSQL Gateway history: found=%t record=%#v err=%v", found, restored, err)
	}
	if restored := newSavedWorkflowDraftService(restarted.savedWorkflowDraftStore).ReadDraft(
		savedDraftContext, ReadWorkflowDraftRequest{DraftID: savedDraftPayload.DraftID}); restored.FailureCode != "" || restored.Draft == nil || restored.Draft.DraftVersion != 1 {
		t.Fatalf("restore configured PostgreSQL workflow draft: %#v", restored)
	}
	if restored, found, err := restarted.workflowRunStore.ReadRun(runContext, runID); err != nil || !found || restored.Status != WorkflowRunStatusSucceeded {
		t.Fatalf("restore configured PostgreSQL workflow run: found=%t record=%#v err=%v", found, restored, err)
	}
}

func newConfiguredPostgresTestDatabase(t *testing.T) (*pgxpool.Pool, string, context.Context) {
	t.Helper()
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	if runtimeUser == strings.TrimSpace(os.Getenv("PGUSER")) {
		t.Fatal("PostgreSQL configured gate requires distinct migration and runtime users")
	}
	runtimeDatabaseURL := postgresIntegrationDatabaseURLForCredentials(
		t, runtimeUser, os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	adminPool, err := workflowdraftmigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		cancel()
		t.Fatalf("open configured PostgreSQL migration pool: %v", err)
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, adminPool)
	resetConfiguredPostgresSchemas(t, ctx, adminPool)
	preparePostgresIntegrationRuntimeRole(t, ctx, adminPool, runtimeUser)
	t.Cleanup(func() {
		cleanup, cleanupCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cleanupCancel()
		resetConfiguredPostgresSchemas(t, cleanup, adminPool)
		adminPool.Close()
		cancel()
	})
	return adminPool, runtimeDatabaseURL, ctx
}

func configuredPostgresMigrationGates(pool *pgxpool.Pool) []postgresMigrationGate {
	return []postgresMigrationGate{
		postgresMigrationGateForAPIKeys(pool),
		postgresMigrationGateForApplicationCatalog(pool),
		postgresMigrationGateForApplicationDrafts(pool),
		postgresMigrationGateForApplicationPublish(pool),
		postgresMigrationGateForGatewayRequests(pool),
		postgresMigrationGateForWorkflowDrafts(pool),
		postgresMigrationGateForWorkflowRuns(pool),
	}
}

func postgresMigrationGateForAPIKeys(pool *pgxpool.Pool) postgresMigrationGate {
	return postgresMigrationGate{name: "api_keys", table: "api_key_records",
		apply: func(ctx context.Context) (string, string, error) {
			state, err := apikeymigrations.Apply(ctx, pool)
			return state.MigrationState, state.MigrationChecksum, err
		}, inspect: func(ctx context.Context) (string, error) {
			state, err := apikeymigrations.Inspect(ctx, pool)
			return state.MigrationState, err
		}, rollback: func(ctx context.Context) (string, error) {
			state, err := apikeymigrations.RollbackForDevTest(ctx, pool)
			return state.MigrationState, err
		}}
}

func postgresMigrationGateForApplicationCatalog(pool *pgxpool.Pool) postgresMigrationGate {
	return postgresMigrationGate{name: "application_catalog", table: "application_catalog_records",
		apply: func(ctx context.Context) (string, string, error) {
			state, err := applicationcatalogmigrations.Apply(ctx, pool)
			return state.MigrationState, state.MigrationChecksum, err
		}, inspect: func(ctx context.Context) (string, error) {
			state, err := applicationcatalogmigrations.Inspect(ctx, pool)
			return state.MigrationState, err
		}, rollback: func(ctx context.Context) (string, error) {
			state, err := applicationcatalogmigrations.RollbackForDevTest(ctx, pool)
			return state.MigrationState, err
		}}
}

func postgresMigrationGateForApplicationDrafts(pool *pgxpool.Pool) postgresMigrationGate {
	return postgresMigrationGate{name: "application_drafts", table: "application_configuration_drafts",
		apply: func(ctx context.Context) (string, string, error) {
			state, err := applicationdraftmigrations.Apply(ctx, pool)
			return state.MigrationState, state.MigrationChecksum, err
		}, inspect: func(ctx context.Context) (string, error) {
			state, err := applicationdraftmigrations.Inspect(ctx, pool)
			return state.MigrationState, err
		}, rollback: func(ctx context.Context) (string, error) {
			state, err := applicationdraftmigrations.RollbackForDevTest(ctx, pool)
			return state.MigrationState, err
		}}
}

func postgresMigrationGateForApplicationPublish(pool *pgxpool.Pool) postgresMigrationGate {
	return postgresMigrationGate{name: "application_publish", table: "application_publish_candidates",
		apply: func(ctx context.Context) (string, string, error) {
			state, err := applicationpublishmigrations.Apply(ctx, pool)
			return state.MigrationState, state.MigrationChecksum, err
		}, inspect: func(ctx context.Context) (string, error) {
			state, err := applicationpublishmigrations.Inspect(ctx, pool)
			return state.MigrationState, err
		}, rollback: func(ctx context.Context) (string, error) {
			state, err := applicationpublishmigrations.RollbackForDevTest(ctx, pool)
			return state.MigrationState, err
		}}
}

func postgresMigrationGateForGatewayRequests(pool *pgxpool.Pool) postgresMigrationGate {
	return postgresMigrationGate{name: "gateway_requests", table: "gateway_request_records",
		apply: func(ctx context.Context) (string, string, error) {
			state, err := gatewayrequestmigrations.Apply(ctx, pool)
			return state.MigrationState, state.MigrationChecksum, err
		}, inspect: func(ctx context.Context) (string, error) {
			state, err := gatewayrequestmigrations.Inspect(ctx, pool)
			return state.MigrationState, err
		}, rollback: func(ctx context.Context) (string, error) {
			state, err := gatewayrequestmigrations.RollbackForDevTest(ctx, pool)
			return state.MigrationState, err
		}}
}

func postgresMigrationGateForWorkflowDrafts(pool *pgxpool.Pool) postgresMigrationGate {
	return postgresMigrationGate{name: "workflow_drafts", table: "saved_workflow_drafts",
		apply: func(ctx context.Context) (string, string, error) {
			state, err := workflowdraftmigrations.Apply(ctx, pool)
			return state.MigrationState, state.MigrationChecksum, err
		}, inspect: func(ctx context.Context) (string, error) {
			state, err := workflowdraftmigrations.Inspect(ctx, pool)
			return state.MigrationState, err
		}, rollback: func(ctx context.Context) (string, error) {
			state, err := workflowdraftmigrations.RollbackForDevTest(ctx, pool)
			return state.MigrationState, err
		}}
}

func postgresMigrationGateForWorkflowRuns(pool *pgxpool.Pool) postgresMigrationGate {
	return postgresMigrationGate{name: "workflow_runs", table: "workflow_run_records",
		apply: func(ctx context.Context) (string, string, error) {
			state, err := workflowrunmigrations.Apply(ctx, pool)
			return state.MigrationState, state.MigrationChecksum, err
		}, inspect: func(ctx context.Context) (string, error) {
			state, err := workflowrunmigrations.Inspect(ctx, pool)
			return state.MigrationState, err
		}, rollback: func(ctx context.Context) (string, error) {
			state, err := workflowrunmigrations.RollbackForDevTest(ctx, pool)
			return state.MigrationState, err
		}}
}

func runConfiguredPostgresMigrationGate(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	gate postgresMigrationGate,
) {
	t.Helper()
	if state, err := gate.rollback(ctx); err != nil || state != "not_applied" {
		t.Fatalf("reset %s migration: state=%s err=%v", gate.name, state, err)
	}
	start := make(chan struct{})
	results := make(chan postgresMigrationResult, 2)
	var wait sync.WaitGroup
	for index := 0; index < 2; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			state, checksum, err := gate.apply(ctx)
			results <- postgresMigrationResult{state: state, checksum: checksum, err: err}
		}()
	}
	close(start)
	wait.Wait()
	close(results)
	checksum := ""
	for result := range results {
		if result.err != nil || result.state != "applied" || result.checksum == "" {
			t.Fatalf("concurrent %s migration did not serialize: %#v", gate.name, result)
		}
		if checksum != "" && result.checksum != checksum {
			t.Fatalf("concurrent %s migration checksum drifted: %s != %s", gate.name, checksum, result.checksum)
		}
		checksum = result.checksum
	}
	if state, repeatedChecksum, err := gate.apply(ctx); err != nil || state != "applied" || repeatedChecksum != checksum {
		t.Fatalf("repeat %s migration: state=%s checksum=%s err=%v", gate.name, state, repeatedChecksum, err)
	}

	tableName := pgx.Identifier{gate.table}.Sanitize()
	missingName := pgx.Identifier{gate.table + "_preflight_missing"}.Sanitize()
	if _, err := pool.Exec(ctx, "ALTER TABLE "+tableName+" RENAME TO "+missingName); err != nil {
		t.Fatalf("hide %s table for preflight negative test: %v", gate.name, err)
	}
	mismatchState, mismatchErr := gate.inspect(ctx)
	_, restoreErr := pool.Exec(ctx, "ALTER TABLE "+missingName+" RENAME TO "+tableName)
	if restoreErr != nil {
		t.Fatalf("restore %s table after preflight negative test: %v", gate.name, restoreErr)
	}
	if mismatchErr != nil || mismatchState != "mismatch" {
		t.Fatalf("%s marker accepted a missing physical table: state=%s err=%v", gate.name, mismatchState, mismatchErr)
	}

	if state, err := gate.rollback(ctx); err != nil || state != "not_applied" {
		t.Fatalf("rollback %s migration: state=%s err=%v", gate.name, state, err)
	}
	if state, err := gate.inspect(ctx); err != nil || state != "not_applied" {
		t.Fatalf("inspect rolled back %s migration: state=%s err=%v", gate.name, state, err)
	}
	if state, _, err := gate.apply(ctx); err != nil || state != "applied" {
		t.Fatalf("reapply %s migration: state=%s err=%v", gate.name, state, err)
	}
}

func resetConfiguredPostgresSchemas(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(ctx, `DROP TABLE IF EXISTS
		workflow_rag_execution_audits,
		workflow_rag_snapshot_fragments,
		workflow_rag_snapshot_versions,
		workflow_rag_snapshot_resources,
		workflow_http_tool_execution_attempts,
		workflow_http_tool_confirmation_decisions,
		workflow_http_tool_execution_audits,
		workflow_http_tool_action_plans,
		workflow_evaluation_suite_decisions,
		workflow_evaluation_suites,
		workflow_evaluation_case_revisions,
		workflow_evaluation_cases,
		workflow_run_records,
		gateway_request_records,
		api_key_records,
		application_publish_candidates,
		application_configuration_drafts,
		application_catalog_records,
		saved_workflow_drafts,
		workflow_run_schema_versions,
		gateway_request_schema_versions,
		api_key_schema_versions,
		application_publish_candidate_schema_versions,
		application_configuration_draft_schema_versions,
		application_catalog_schema_versions,
		workflow_saved_draft_schema_versions
		CASCADE`)
	if err != nil {
		t.Fatalf("reset configured PostgreSQL schemas: %v", err)
	}
	if _, err := pool.Exec(ctx, `DROP FUNCTION IF EXISTS reject_workflow_http_tool_append_only_mutation()`); err != nil {
		t.Fatalf("reset configured PostgreSQL append-only guard: %v", err)
	}
	if _, err := pool.Exec(ctx, `DROP FUNCTION IF EXISTS reject_workflow_rag_snapshot_append_only_mutation()`); err != nil {
		t.Fatalf("reset configured PostgreSQL RAG append-only guard: %v", err)
	}
}

func assertConfiguredPostgresSchema(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	columns := []struct{ table, column, expected string }{
		{"api_key_records", "credential_digest", "bytea"},
		{"api_key_records", "scopes", "text[]"},
		{"api_key_records", "sanitized_record_payload", "jsonb"},
		{"api_key_records", "created_at", "timestamp with time zone"},
		{"api_key_records", "record_version", "bigint"},
		{"application_catalog_records", "sanitized_record_payload", "jsonb"},
		{"application_catalog_records", "updated_at", "timestamp with time zone"},
		{"application_configuration_drafts", "sanitized_draft_payload", "jsonb"},
		{"application_publish_candidates", "sanitized_candidate_payload", "jsonb"},
		{"gateway_request_records", "sanitized_request_record", "jsonb"},
		{"gateway_request_records", "started_at", "timestamp with time zone"},
		{"saved_workflow_drafts", "sanitized_draft_payload", "jsonb"},
		{"workflow_run_records", "sanitized_run_record", "jsonb"},
		{"workflow_run_records", "started_at", "timestamp with time zone"},
		{"workflow_http_tool_action_plans", "tool_version", "integer"},
		{"workflow_http_tool_confirmation_decisions", "tool_version", "integer"},
		{"workflow_http_tool_execution_audits", "tool_version", "integer"},
		{"workflow_http_tool_execution_attempts", "sanitized_execution_attempt", "jsonb"},
		{"workflow_rag_snapshot_resources", "sanitized_resource_payload", "jsonb"},
		{"workflow_rag_snapshot_versions", "sanitized_snapshot_payload", "jsonb"},
		{"workflow_rag_snapshot_fragments", "sanitized_fragment_payload", "jsonb"},
		{"workflow_rag_execution_audits", "sanitized_audit_payload", "jsonb"},
	}
	for _, column := range columns {
		var actual string
		err := pool.QueryRow(ctx, `SELECT format_type(attribute.atttypid, attribute.atttypmod)
			FROM pg_attribute attribute
			JOIN pg_class relation ON relation.oid=attribute.attrelid
			JOIN pg_namespace namespace ON namespace.oid=relation.relnamespace
			WHERE namespace.nspname='public' AND relation.relname=$1 AND attribute.attname=$2
			  AND attribute.attnum > 0 AND NOT attribute.attisdropped`, column.table, column.column).Scan(&actual)
		if err != nil || actual != column.expected {
			t.Fatalf("PostgreSQL column type drifted: %s.%s expected=%s actual=%s err=%v",
				column.table, column.column, column.expected, actual, err)
		}
	}

	indexes := []struct{ name, orderFragment string }{
		{"api_key_records_owner_list_idx", "created_at DESC, api_key_id DESC"},
		{"api_key_records_owner_application_list_idx", "created_at DESC, api_key_id DESC"},
		{"api_key_records_authentication_idx", "api_key_id, lifecycle_state, expires_at"},
		{"application_catalog_records_owner_list_idx", "updated_at DESC, application_id DESC"},
		{"application_catalog_records_owner_kind_list_idx", "updated_at DESC, application_id DESC"},
		{"application_configuration_drafts_scope_list_idx", "updated_at DESC, draft_id"},
		{"application_publish_candidates_scope_list_idx", "created_at DESC, candidate_id DESC"},
		{"gateway_request_records_history_idx", "started_at DESC, request_id DESC"},
		{"gateway_request_records_selection_idx", "started_at DESC, request_id DESC"},
		{"saved_workflow_drafts_owner_list_idx", "updated_at DESC, draft_id"},
		{"workflow_run_records_history_idx", "started_at DESC, run_id DESC"},
		{"workflow_http_tool_execution_attempts_status_idx", "status, claimed_at, attempt_id"},
		{"workflow_rag_snapshot_resources_list_idx", "lifecycle_state, snapshot_key"},
		{"workflow_rag_snapshot_versions_history_idx", "snapshot_version DESC"},
		{"workflow_rag_execution_audits_history_idx", "occurred_at DESC, event_id"},
	}
	for _, index := range indexes {
		var definition string
		err := pool.QueryRow(ctx, `SELECT pg_get_indexdef(to_regclass('public.' || $1))`, index.name).Scan(&definition)
		if err != nil || !strings.Contains(definition, index.orderFragment) {
			t.Fatalf("PostgreSQL index drifted: index=%s fragment=%q definition=%q err=%v",
				index.name, index.orderFragment, definition, err)
		}
	}
}

func assertPostgresRuntimeIsolationAndConnections(
	t *testing.T,
	ctx context.Context,
	adminPool *pgxpool.Pool,
	runtimeDatabaseURL string,
) {
	t.Helper()
	runtimePool, err := workflowdraftmigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatalf("open configured PostgreSQL runtime pool: %v", err)
	}
	defer runtimePool.Close()
	var runtimeUser string
	if err := runtimePool.QueryRow(ctx, "SELECT current_user").Scan(&runtimeUser); err != nil {
		t.Fatalf("read configured PostgreSQL runtime identity: %v", err)
	}
	if runtimeUser == strings.TrimSpace(os.Getenv("PGUSER")) {
		t.Fatalf("configured PostgreSQL runtime identity reused migration role: %s", runtimeUser)
	}
	_, ddlErr := runtimePool.Exec(ctx, "CREATE TABLE configured_runtime_must_not_create (id bigint)")
	if ddlErr == nil {
		_, _ = adminPool.Exec(ctx, "DROP TABLE IF EXISTS configured_runtime_must_not_create")
		t.Fatal("configured PostgreSQL runtime role unexpectedly executed DDL")
	}
	var databaseError *pgconn.PgError
	if !errors.As(ddlErr, &databaseError) || databaseError.Code != "42501" {
		t.Fatalf("configured PostgreSQL runtime DDL denial returned unexpected error: %v", ddlErr)
	}

	connections := make([]*pgxpool.Conn, 0, 4)
	backendIDs := make(map[int32]struct{}, 4)
	for index := 0; index < 4; index++ {
		connection, err := runtimePool.Acquire(ctx)
		if err != nil {
			t.Fatalf("acquire configured PostgreSQL runtime connection %d: %v", index, err)
		}
		connections = append(connections, connection)
		var backendID int32
		if err := connection.QueryRow(ctx, "SELECT pg_backend_pid()").Scan(&backendID); err != nil {
			t.Fatalf("read configured PostgreSQL backend id %d: %v", index, err)
		}
		backendIDs[backendID] = struct{}{}
	}
	for _, connection := range connections {
		connection.Release()
	}
	if len(backendIDs) != len(connections) {
		t.Fatalf("configured PostgreSQL pool did not establish independent connections: backends=%v", backendIDs)
	}
}

func configuredPostgresProductConfig(databaseURL string) config.Config {
	return config.Config{
		ListenAddr: ":0", ReadHeaderTimeout: time.Second, WriteTimeout: time.Second,
		BridgeTimeout: time.Second, BridgeMode: "process_per_request", BridgeWorkerCount: 1,
		BridgeQueueCapacity: 1, BridgeHandshakeTimeout: time.Second, PythonBinary: "python3",
		BridgeScript: "scripts/run-platform-bridge.py", Provider: "mock", Model: "platform-model",
		ControlPlaneReadDevAuthEnabled: true, ControlPlaneReadStoreMode: "fake_store_dev",
		ControlPlaneReadDatabaseTimeout:  time.Second,
		WorkflowSavedDraftDevHTTPEnabled: true, WorkflowSavedDraftDevWriteEnabled: true,
		WorkflowSavedDraftStoreMode: "postgres_dev_test", WorkflowSavedDraftDatabaseURL: databaseURL,
		WorkflowSavedDraftDatabaseTimeout: time.Second,
		ApplicationDraftDevHTTPEnabled:    true, ApplicationDraftDevWriteEnabled: true,
		ApplicationDraftStoreMode: "postgres_dev_test", ApplicationDraftDatabaseURL: databaseURL,
		ApplicationDraftDatabaseTimeout:  time.Second,
		ApplicationPublishDevHTTPEnabled: true, ApplicationPublishDevWriteEnabled: true,
		ApplicationPublishStoreMode: "postgres_dev_test", ApplicationPublishDatabaseURL: databaseURL,
		ApplicationPublishDatabaseTimeout: time.Second,
		ApplicationCatalogDevHTTPEnabled:  true, ApplicationCatalogDevWriteEnabled: true,
		ApplicationCatalogStoreMode: "postgres_dev_test", ApplicationCatalogDatabaseURL: databaseURL,
		ApplicationCatalogDatabaseTimeout: time.Second,
		APIKeyLifecycleDevHTTPEnabled:     true, APIKeyLifecycleDevWriteEnabled: true,
		APIKeyStoreMode: "postgres_dev_test", APIKeyDatabaseURL: databaseURL, APIKeyDatabaseTimeout: time.Second,
		GatewayAuthMode:            gatewayAPIKeyAuthenticationSource,
		WorkflowExecutorDevEnabled: true, WorkflowRunStoreMode: "postgres_dev_test",
		WorkflowRunDatabaseURL: databaseURL, WorkflowRunDatabaseTimeout: time.Second,
		GatewayRequestHistoryDevEnabled: true, GatewayRequestStoreMode: "postgres_dev_test",
		GatewayRequestDatabaseURL: databaseURL, GatewayRequestDatabaseTimeout: time.Second,
	}
}

func assertConfiguredPostgresRepositorySelection(t *testing.T, server *Server) {
	t.Helper()
	if server == nil || server.localPersistenceRuntime != nil {
		t.Fatal("configured PostgreSQL platform unexpectedly selected local SQLite runtime")
	}
	for component, mode := range map[string]string{
		"application_catalog": server.config.ApplicationCatalogStoreMode,
		"application_draft":   server.config.ApplicationDraftStoreMode,
		"application_publish": server.config.ApplicationPublishStoreMode,
		"api_key":             server.config.APIKeyStoreMode,
		"gateway_request":     server.config.GatewayRequestStoreMode,
		"workflow_draft":      server.config.WorkflowSavedDraftStoreMode,
		"workflow_run":        server.config.WorkflowRunStoreMode,
	} {
		if mode != "postgres_dev_test" {
			t.Fatalf("configured PostgreSQL store selection drifted: component=%s mode=%s", component, mode)
		}
	}
	if _, ok := server.applicationCatalogRepository.(*postgresApplicationCatalogRepository); !ok {
		t.Fatalf("configured application catalog did not select PostgreSQL: %T", server.applicationCatalogRepository)
	}
	if _, ok := server.apiKeyRepository.(*postgresAPIKeyRepository); !ok {
		t.Fatalf("configured API key store did not select PostgreSQL: %T", server.apiKeyRepository)
	}
	if _, ok := server.gatewayRequestHistoryStore.(*postgresGatewayRequestStore); !ok {
		t.Fatalf("configured Gateway request store did not select PostgreSQL: %T", server.gatewayRequestHistoryStore)
	}
	if _, ok := server.workflowRunStore.(*postgresWorkflowRunStore); !ok {
		t.Fatalf("configured workflow run store did not select PostgreSQL: %T", server.workflowRunStore)
	}
	if actionStore, ok := server.workflowHTTPToolActionStore.(*postgresWorkflowHTTPToolActionStore); !ok || actionStore.pool != server.workflowRunStore.(*postgresWorkflowRunStore).pool {
		t.Fatalf("configured workflow tool actions did not share the PostgreSQL pool: %T", server.workflowHTTPToolActionStore)
	}
	if snapshotStore, ok := server.workflowRAGSnapshotRepository.(*postgresWorkflowRAGSnapshotRepository); !ok || snapshotStore.pool != server.workflowRunStore.(*postgresWorkflowRunStore).pool {
		t.Fatalf("configured workflow RAG snapshots did not share the PostgreSQL pool: %T", server.workflowRAGSnapshotRepository)
	}
}

func assertPostgresGatewayTuplePagination(
	t *testing.T,
	store gatewayRequestStore,
	ctx context.Context,
	applicationID string,
) {
	t.Helper()
	requestContext := GatewayRequestContext{
		RequestContext: ctx, TenantRef: "tenant_demo", WorkspaceID: "workspace_demo",
		ConsumerRef: "consumer_postgres_pagination", ApplicationID: applicationID,
		SubjectRef: "subject_platform_ops", ScopeGrants: []string{"gateway_requests:read"},
		AuditContext: "postgres-pagination", Source: "configured", RequestID: "request-postgres-pagination",
		AuditRef: "audit-postgres-pagination",
	}
	startedAt := time.Date(2026, 7, 14, 9, 10, 0, 0, time.UTC)
	requestIDs := []string{"request_pg_page_a", "request_pg_page_b", "request_pg_page_c"}
	for _, requestID := range requestIDs {
		record := gatewayRequestTestRecord(requestContext, requestID, startedAt)
		if err := store.CreateRequest(requestContext, &record); err != nil {
			t.Fatalf("create PostgreSQL pagination Gateway record %s: %v", requestID, err)
		}
	}
	first, err := store.ListRequests(requestContext, GatewayRequestListFilter{Limit: 2})
	if err != nil || len(first.Records) != 2 || !first.HasMore || first.Records[0].RequestID != requestIDs[2] ||
		first.Records[1].RequestID != requestIDs[1] {
		t.Fatalf("Gateway request PostgreSQL tuple pagination first page drifted: %#v err=%v", first, err)
	}
	beforeTime, err := time.Parse(time.RFC3339Nano, first.Records[1].StartedAt)
	if err != nil {
		t.Fatalf("parse PostgreSQL Gateway pagination cursor time: %v", err)
	}
	second, err := store.ListRequests(requestContext, GatewayRequestListFilter{
		Limit: 2, BeforeTime: &beforeTime, BeforeRequestID: first.Records[1].RequestID,
	})
	if err != nil || len(second.Records) != 1 || second.HasMore || second.Records[0].RequestID != requestIDs[0] {
		t.Fatalf("Gateway request PostgreSQL tuple pagination second page drifted: %#v err=%v", second, err)
	}
}

func assertConfiguredPostgresSecretBoundary(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	apiKeyID string,
	requestID string,
	forbidden ...string,
) {
	t.Helper()
	var keyPayload, digestHex, requestPayload string
	if err := pool.QueryRow(ctx, `SELECT sanitized_record_payload::text, encode(credential_digest, 'hex')
		FROM api_key_records WHERE api_key_id=$1`, apiKeyID).Scan(&keyPayload, &digestHex); err != nil {
		t.Fatalf("inspect configured PostgreSQL API key material: %v", err)
	}
	if len(digestHex) != 64 || strings.Contains(keyPayload, "credential_digest") {
		t.Fatalf("configured PostgreSQL API key digest boundary drifted: digest_length=%d payload=%s", len(digestHex), keyPayload)
	}
	if err := pool.QueryRow(ctx, `SELECT sanitized_request_record::text FROM gateway_request_records WHERE request_id=$1`, requestID).Scan(&requestPayload); err != nil {
		t.Fatalf("inspect configured PostgreSQL Gateway request payload: %v", err)
	}
	for _, value := range forbidden {
		if value != "" && (strings.Contains(keyPayload, value) || strings.Contains(requestPayload, value) || strings.Contains(digestHex, value)) {
			t.Fatal("configured PostgreSQL persisted forbidden API key or Gateway request material")
		}
	}
}

func assertPostgresAuthenticationRevokeRace(
	t *testing.T,
	server *Server,
	service apiKeyService,
	requestContext APIKeyContext,
	issued APIKeyResult,
) {
	t.Helper()
	start := make(chan struct{})
	results := make(chan string, 8)
	var wait sync.WaitGroup
	for index := 0; index < 8; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			request := newAPIKeyGatewayIntegrationRequest(issued.CredentialToken)
			results <- server.authenticateGatewayAPIKey(request, newRequestTrace(request, "/v1/models"), "models:read").FailureCode
		}()
	}
	close(start)
	revoked := service.Revoke(requestContext, issued.Record.APIKeyID, 1)
	wait.Wait()
	close(results)
	if revoked.FailureCode != "" || revoked.Record == nil || revoked.Record.RecordVersion != 2 {
		t.Fatalf("revoke configured PostgreSQL API key during authentication race: %#v", revoked)
	}
	for failure := range results {
		if failure != "" && failure != APIKeyFailureRevoked {
			t.Fatalf("configured PostgreSQL authentication/revoke race returned unstable failure: %s", failure)
		}
	}
	request := newAPIKeyGatewayIntegrationRequest(issued.CredentialToken)
	if result := server.authenticateGatewayAPIKey(request, newRequestTrace(request, "/v1/models"), "models:read"); result.FailureCode != APIKeyFailureRevoked {
		t.Fatalf("configured PostgreSQL revoked key remained valid: %#v", result)
	}
}

func assertPostgresApplicationArchiveRace(
	t *testing.T,
	server *Server,
	service applicationCatalogService,
	requestContext ApplicationCatalogContext,
	applicationID string,
	issued APIKeyResult,
) {
	t.Helper()
	start := make(chan struct{})
	results := make(chan string, 8)
	var wait sync.WaitGroup
	for index := 0; index < 8; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			request := newAPIKeyGatewayIntegrationRequest(issued.CredentialToken)
			results <- server.authenticateGatewayAPIKey(request, newRequestTrace(request, "/v1/models"), "models:read").FailureCode
		}()
	}
	close(start)
	archived := service.Archive(requestContext, applicationID, 1)
	wait.Wait()
	close(results)
	if archived.FailureCode != "" || archived.Record == nil || archived.Record.LifecycleState != applicationCatalogLifecycleArchived {
		t.Fatalf("archive configured PostgreSQL application during authentication race: %#v", archived)
	}
	for failure := range results {
		if failure != "" && failure != APIKeyFailureApplicationUnavailable {
			t.Fatalf("configured PostgreSQL authentication/archive race returned unstable failure: %s", failure)
		}
	}
	request := newAPIKeyGatewayIntegrationRequest(issued.CredentialToken)
	if result := server.authenticateGatewayAPIKey(request, newRequestTrace(request, "/v1/models"), "models:read"); result.FailureCode != APIKeyFailureApplicationUnavailable {
		t.Fatalf("configured PostgreSQL archived application still admitted API key: %#v", result)
	}
}

func assertConfiguredPostgresStoresFailClosed(
	t *testing.T,
	catalogContext ApplicationCatalogContext,
	draftContext ApplicationConfigurationDraftContext,
	publishContext ApplicationPublishContext,
	apiKeyContext APIKeyContext,
	savedDraftContext SavedWorkflowDraftContext,
	runContext WorkflowRunContext,
	gatewayContext GatewayRequestContext,
	applicationID, draftID, candidateID, apiKeyID, savedDraftID, runID string,
	catalogRepository applicationCatalogRepository,
	draftRepository applicationConfigurationDraftRepository,
	publishRepository applicationPublishCandidateRepository,
	apiKeyRepository apiKeyRepository,
	savedDraftStore savedWorkflowDraftStore,
	runStore workflowRunStore,
	gatewayStore gatewayRequestStore,
) {
	t.Helper()
	if _, err := catalogRepository.Read(catalogContext, applicationID); !errors.Is(err, errApplicationCatalogStoreUnavailable) {
		t.Fatalf("closed configured PostgreSQL application catalog did not fail closed: %v", err)
	}
	if _, err := draftRepository.Read(draftContext, draftID); !errors.Is(err, errApplicationDraftStoreUnavailable) {
		t.Fatalf("closed configured PostgreSQL application draft store did not fail closed: %v", err)
	}
	if _, err := publishRepository.Read(publishContext, candidateID); !errors.Is(err, errApplicationPublishStoreUnavailable) {
		t.Fatalf("closed configured PostgreSQL application publish store did not fail closed: %v", err)
	}
	if _, err := apiKeyRepository.Read(apiKeyContext, apiKeyID); !errors.Is(err, errAPIKeyStoreUnavailable) {
		t.Fatalf("closed configured PostgreSQL API key store did not fail closed: %v", err)
	}
	if result := newSavedWorkflowDraftService(savedDraftStore).ReadDraft(
		savedDraftContext, ReadWorkflowDraftRequest{DraftID: savedDraftID}); result.FailureCode != SavedWorkflowDraftFailureStoreUnavailable || result.Draft != nil {
		t.Fatalf("closed configured PostgreSQL workflow draft store did not fail closed: %#v", result)
	}
	if _, _, err := runStore.ReadRun(runContext, runID); !errors.Is(err, errWorkflowRunStoreUnavailable) {
		t.Fatalf("closed configured PostgreSQL workflow run store did not fail closed: %v", err)
	}
	if page, err := gatewayStore.ListRequests(gatewayContext, GatewayRequestListFilter{Limit: 10}); !errors.Is(err, errGatewayRequestStoreUnavailable) || len(page.Records) != 0 {
		t.Fatalf("closed configured PostgreSQL Gateway request store did not fail closed: page=%#v err=%v", page, err)
	}
}
