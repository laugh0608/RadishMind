//go:build postgres_integration

package httpapi

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
	gatewayrequestmigrations "radishmind.local/services/platform/migrations/gateway_requests"
)

func TestPostgresGatewayRequestStoreIntegration(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeDatabaseURL := postgresIntegrationDatabaseURLForCredentials(t, runtimeUser, os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	adminPool, err := gatewayrequestmigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, adminPool)
	resetPostgresGatewayRequestSchema(t, ctx, adminPool)
	preparePostgresIntegrationRuntimeRole(t, ctx, adminPool, runtimeUser)
	t.Cleanup(func() {
		cleanup, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		resetPostgresGatewayRequestSchema(t, cleanup, adminPool)
		adminPool.Close()
	})

	installPostgresGatewayRequestV2Schema(t, ctx, adminPool)
	legacyState, err := gatewayrequestmigrations.Inspect(ctx, adminPool)
	if err != nil || legacyState.MigrationState != gatewayrequestmigrations.MigrationStateUpgradeRequired {
		t.Fatalf("inspect Gateway request v2 migration: %#v %v", legacyState, err)
	}
	state, err := gatewayrequestmigrations.Apply(ctx, adminPool)
	if err != nil || state.MigrationState != gatewayrequestmigrations.MigrationStateApplied {
		t.Fatalf("apply Gateway request migration: %#v %v", state, err)
	}
	var legacySchema string
	var legacyAttempts int
	if err = adminPool.QueryRow(ctx, `SELECT schema_version, provider_attempt_count
        FROM gateway_request_records WHERE request_id='request_postgres_legacy_v2'`).
		Scan(&legacySchema, &legacyAttempts); err != nil ||
		legacySchema != gatewayRequestRecordSchemaVersionV2 || legacyAttempts != 0 {
		t.Fatalf("upgrade Gateway request v2 row: schema=%s attempts=%d err=%v",
			legacySchema, legacyAttempts, err)
	}
	if repeated, repeatErr := gatewayrequestmigrations.Apply(ctx, adminPool); repeatErr != nil || repeated.MigrationChecksum != state.MigrationChecksum {
		t.Fatalf("repeat Gateway request migration: %#v %v", repeated, repeatErr)
	}

	runtimePool, err := gatewayrequestmigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	if _, ddlErr := runtimePool.Exec(ctx, "CREATE TABLE gateway_request_runtime_must_not_create (id integer)"); ddlErr == nil {
		t.Fatal("Gateway request runtime role must not have schema DDL permission")
	}
	store := newPostgresGatewayRequestStore(runtimePool)
	requestContext := gatewayRequestTestContext()
	requestContext.RequestContext = ctx
	base := time.Now().UTC().Add(-time.Minute)
	for index := 0; index < 3; index++ {
		record := gatewayRequestTestRecord(requestContext, "request_pg_"+string(rune('a'+index)), base.Add(time.Duration(index)*time.Second))
		record.SelectedProvider = "mock"
		record.SelectedProfile = "profile_pg"
		record.SelectedModel = "model_pg"
		record.ProviderRouteConfigurationID = "gateway-default"
		record.ProviderRouteGeneration = index + 1
		record.ProviderRouteSnapshotDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		if err = store.CreateRequest(requestContext, &record); err != nil {
			t.Fatal(err)
		}
		record.Status = GatewayRequestStatusSucceeded
		record.CompletedAt = base.Add(time.Duration(index)*time.Second + time.Millisecond).Format(time.RFC3339Nano)
		record.DurationMS = 1
		record.HTTPStatusCode = 200
		if index == 2 {
			record.Usage = GatewayRequestUsage{
				Availability: GatewayRequestUsageReported,
				Source:       "openai_compatible_usage",
				InputTokens:  17,
				OutputTokens: 6,
				TotalTokens:  23,
			}
		}
		if err = store.UpdateRequest(requestContext, &record); err != nil {
			t.Fatal(err)
		}
	}
	page, err := store.ListRequests(requestContext, GatewayRequestListFilter{Limit: 2, Provider: "mock", Profile: "profile_pg", Model: "model_pg"})
	if err != nil || len(page.Records) != 2 || !page.HasMore || page.Records[0].RequestID != "request_pg_c" {
		t.Fatalf("unexpected Gateway request PostgreSQL page: %#v %v", page, err)
	}
	other := requestContext
	other.TenantRef = "tenant_other"
	if scoped, listErr := store.ListRequests(other, GatewayRequestListFilter{Limit: 10}); listErr != nil || len(scoped.Records) != 0 {
		t.Fatalf("Gateway request scope leaked: %#v %v", scoped, listErr)
	}

	runtimePool.Close()
	reopened, err := gatewayrequestmigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	store = newPostgresGatewayRequestStore(reopened)
	recovered, found, err := store.ReadRequest(requestContext, "request_pg_c")
	if err != nil || !found || recovered.Status != GatewayRequestStatusSucceeded ||
		recovered.StoreMode != gatewayRequestStoreModePostgresDevTest ||
		recovered.ProviderRouteConfigurationID != "gateway-default" ||
		recovered.ProviderRouteGeneration != 3 ||
		recovered.ProviderRouteSnapshotDigest != "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" ||
		recovered.Usage.Availability != GatewayRequestUsageReported ||
		recovered.Usage.Source != "openai_compatible_usage" || recovered.Usage.TotalTokens != 23 {
		t.Fatalf("Gateway request restart recovery failed: found=%v record=%#v err=%v", found, recovered, err)
	}
	plan := gatewayProviderAttemptTestPlan(t, "request_pg_fallback")
	root := gatewayRequestTestRecord(requestContext, plan.RootRequestID, time.Now().UTC())
	attemptRecord, err := newGatewayProviderAttemptHistoryRecord(root, plan)
	if err != nil || store.CreateRequest(requestContext, &attemptRecord) != nil {
		t.Fatalf("create PostgreSQL v3 root: record=%#v err=%v", attemptRecord, err)
	}
	attemptService := newGatewayProviderAttemptHistoryService(store)
	attemptTime := time.Now().UTC().Add(time.Second)
	if _, err = attemptService.StartAttempt(
		requestContext, plan.RootRequestID, plan.Targets[0], "quota-primary", attemptTime,
	); err != nil {
		t.Fatalf("start PostgreSQL primary attempt: %v", err)
	}
	providerFailure := gatewayProviderAttemptTestFailure(
		t, bridge.ProviderFailureTemporarilyUnavailable, bridge.ProviderFallbackEligible, bridge.ProviderAttemptFailed,
	)
	if _, err = attemptService.CompleteAttempt(
		requestContext, plan.RootRequestID, plan.Targets[0].AttemptID,
		GatewayRequestUsage{Availability: GatewayRequestUsageNotReported},
		gatewayRequestCostUnavailable(GatewayRequestCostUsageNotReported, "provider_usage_not_reported"),
		&providerFailure, true, attemptTime.Add(time.Second),
	); err != nil {
		t.Fatalf("checkpoint PostgreSQL fallback pending: %v", err)
	}
	if _, err = attemptService.StartAttempt(
		requestContext, plan.RootRequestID, plan.Targets[1], "quota-fallback", attemptTime.Add(2*time.Second),
	); err != nil {
		t.Fatalf("start PostgreSQL fallback attempt: %v", err)
	}
	usage := GatewayRequestUsage{
		Availability: GatewayRequestUsageReported,
		Source:       "openai_compatible_usage",
		InputTokens:  2,
		OutputTokens: 3,
		TotalTokens:  5,
	}
	cost := buildGatewayRequestCostEstimate(
		true, usage, gatewayModelPricingSnapshotFromAttempt(plan.Targets[1].PricingSnapshot),
	)
	if _, err = attemptService.CompleteAttempt(
		requestContext, plan.RootRequestID, plan.Targets[1].AttemptID,
		usage, cost, nil, false, attemptTime.Add(3*time.Second),
	); err != nil {
		t.Fatalf("complete PostgreSQL fallback attempt: %v", err)
	}
	finished, err := attemptService.Finalize(
		requestContext, plan.RootRequestID, GatewayRequestStatusSucceeded, http.StatusOK, "", "",
		attemptTime.Add(4*time.Second),
	)
	if err != nil || !finished.FallbackUsed || finished.ProviderAttemptCount != 2 ||
		finished.TerminalAttemptID != plan.Targets[1].AttemptID {
		t.Fatalf("finalize PostgreSQL fallback attempt: record=%#v err=%v", finished, err)
	}
	var attemptCount int
	var fallbackUsed bool
	var terminalProvider string
	var terminalProfile string
	if err = reopened.QueryRow(ctx, `SELECT provider_attempt_count, fallback_used, terminal_provider, terminal_profile
        FROM gateway_request_records WHERE request_id=$1`, plan.RootRequestID).
		Scan(&attemptCount, &fallbackUsed, &terminalProvider, &terminalProfile); err != nil {
		t.Fatalf("inspect PostgreSQL attempt summary columns: %v", err)
	}
	if attemptCount != 2 || !fallbackUsed || terminalProvider != plan.Targets[1].ProviderID ||
		terminalProfile != plan.Targets[1].RuntimeProfile {
		t.Fatalf("PostgreSQL attempt summary drifted: count=%d fallback=%v provider=%s profile=%s",
			attemptCount, fallbackUsed, terminalProvider, terminalProfile)
	}

	concurrentPlan := gatewayProviderAttemptTestPlan(t, "request_pg_attempt_cas")
	concurrentRoot := gatewayRequestTestRecord(requestContext, concurrentPlan.RootRequestID, time.Now().UTC())
	concurrentRecord, err := newGatewayProviderAttemptHistoryRecord(concurrentRoot, concurrentPlan)
	if err != nil || store.CreateRequest(requestContext, &concurrentRecord) != nil {
		t.Fatalf("create concurrent PostgreSQL v3 root: %v", err)
	}
	concurrentService := newGatewayProviderAttemptHistoryService(store)
	startAttempt := make(chan struct{})
	attemptResults := make(chan error, 2)
	var attemptWait sync.WaitGroup
	for index := 0; index < 2; index++ {
		attemptWait.Add(1)
		go func() {
			defer attemptWait.Done()
			<-startAttempt
			_, checkpointErr := concurrentService.StartAttempt(
				requestContext, concurrentPlan.RootRequestID, concurrentPlan.Targets[0],
				"quota-primary", time.Now().UTC(),
			)
			attemptResults <- checkpointErr
		}()
	}
	close(startAttempt)
	attemptWait.Wait()
	close(attemptResults)
	attemptWinners := 0
	for result := range attemptResults {
		if result == nil {
			attemptWinners++
		} else if result != errGatewayRequestStoreConflict && result != errGatewayRequestStoreContract {
			t.Fatalf("unexpected PostgreSQL attempt CAS result: %v", result)
		}
	}
	if attemptWinners != 1 {
		t.Fatalf("expected one PostgreSQL attempt checkpoint winner, got %d", attemptWinners)
	}
	if _, err = reopened.Exec(ctx, `UPDATE gateway_request_records
        SET sanitized_request_record='{"unexpected":true}'::jsonb WHERE request_id=$1`, plan.RootRequestID); err != nil {
		t.Fatalf("prepare PostgreSQL corrupted v3 payload: %v", err)
	}
	if _, _, readErr := store.ReadRequest(requestContext, plan.RootRequestID); readErr != errGatewayRequestStoreContract {
		t.Fatalf("PostgreSQL corrupted v3 payload did not fail closed: %v", readErr)
	}

	canceledContext := requestContext
	canceledRequestContext, cancelRequest := context.WithCancel(ctx)
	canceledContext.RequestContext = canceledRequestContext
	canceledRecord := gatewayRequestTestRecord(canceledContext, "request_pg_canceled", time.Now().UTC())
	if err = store.CreateRequest(canceledContext, &canceledRecord); err != nil {
		t.Fatal(err)
	}
	cancelRequest()
	canceledRecord.Status = GatewayRequestStatusCanceled
	canceledRecord.CompletedAt = time.Now().UTC().Format(time.RFC3339Nano)
	canceledRecord.HTTPStatusCode = http.StatusRequestTimeout
	canceledRecord.FailureCode = bridge.ErrorCodeWorkerCanceled
	canceledRecord.FailureBoundary = errorBoundaryPythonBridge
	directUpdate := canceledRecord
	if updateErr := store.UpdateRequest(canceledContext, &directUpdate); updateErr == nil {
		t.Fatal("canceled request context unexpectedly persisted the terminal record")
	}
	terminalServer := &Server{config: config.Config{GatewayRequestDatabaseTimeout: time.Second}}
	terminalContext, terminalCancel := terminalServer.gatewayRequestTerminalStoreContext(canceledContext)
	if err = store.UpdateRequest(terminalContext, &canceledRecord); err != nil {
		terminalCancel()
		t.Fatalf("detached terminal context did not persist cancellation: %v", err)
	}
	terminalCancel()
	persistedCancellation, found, err := store.ReadRequest(requestContext, canceledRecord.RequestID)
	if err != nil || !found || persistedCancellation.Status != GatewayRequestStatusCanceled || persistedCancellation.RecordVersion != 2 {
		t.Fatalf("canceled Gateway request terminal state was not durable: found=%v record=%#v err=%v", found, persistedCancellation, err)
	}

	running := gatewayRequestTestRecord(requestContext, "request_pg_concurrent", time.Now().UTC())
	if err = store.CreateRequest(requestContext, &running); err != nil {
		t.Fatal(err)
	}
	left, right := running, running
	var wait sync.WaitGroup
	results := make(chan error, 2)
	for _, candidate := range []*GatewayRequestRecord{&left, &right} {
		wait.Add(1)
		go func(value *GatewayRequestRecord) {
			defer wait.Done()
			value.Status = GatewayRequestStatusFailed
			value.CompletedAt = time.Now().UTC().Format(time.RFC3339Nano)
			value.HTTPStatusCode = 502
			value.FailureCode = "PLATFORM_GATEWAY_FAILED"
			value.FailureBoundary = errorBoundaryPythonBridge
			results <- store.UpdateRequest(requestContext, value)
		}(candidate)
	}
	wait.Wait()
	close(results)
	winners := 0
	for result := range results {
		if result == nil {
			winners++
		} else if result != errGatewayRequestStoreConflict {
			t.Fatalf("unexpected Gateway request CAS result: %v", result)
		}
	}
	if winners != 1 {
		t.Fatalf("expected one Gateway request CAS winner, got %d", winners)
	}
	reopened.Close()

	cfg := config.Config{
		ControlPlaneReadDevAuthEnabled:  true,
		GatewayRequestHistoryDevEnabled: true,
		GatewayRequestStoreMode:         gatewayRequestStoreModePostgresDevTest,
		GatewayRequestDatabaseURL:       runtimeDatabaseURL,
		GatewayRequestDatabaseTimeout:   5 * time.Second,
	}
	selected, selectedMode, closeSelected, err := newGatewayRequestStoreFromConfig(cfg)
	if err != nil || selected == nil || selectedMode != gatewayRequestStoreModePostgresDevTest || closeSelected == nil {
		t.Fatalf("Gateway request factory did not select PostgreSQL: store=%T mode=%s err=%v", selected, selectedMode, err)
	}
	closeSelected()

	if _, err = adminPool.Exec(ctx, "UPDATE gateway_request_schema_versions SET migration_checksum='sha256:mismatch' WHERE component=$1", gatewayrequestmigrations.Component); err != nil {
		t.Fatal(err)
	}
	if mismatch, inspectErr := gatewayrequestmigrations.Inspect(ctx, adminPool); inspectErr != nil || mismatch.MigrationState != gatewayrequestmigrations.MigrationStateMismatch {
		t.Fatalf("Gateway request marker mismatch not detected: %#v %v", mismatch, inspectErr)
	}
	if selected, _, closeStore, factoryErr := newGatewayRequestStoreFromConfig(cfg); factoryErr == nil || selected != nil || closeStore != nil {
		t.Fatalf("Gateway request factory accepted marker mismatch: store=%T close=%v err=%v", selected, closeStore != nil, factoryErr)
	}
	if _, err = adminPool.Exec(ctx, "UPDATE gateway_request_schema_versions SET migration_checksum=$1 WHERE component=$2", gatewayrequestmigrations.ExpectedChecksum(), gatewayrequestmigrations.Component); err != nil {
		t.Fatal(err)
	}
	if _, err = gatewayrequestmigrations.RollbackForDevTest(ctx, adminPool); err != nil {
		t.Fatal(err)
	}
	if notApplied, inspectErr := gatewayrequestmigrations.Inspect(ctx, adminPool); inspectErr != nil || notApplied.MigrationState != gatewayrequestmigrations.MigrationStateNotApplied {
		t.Fatalf("Gateway request rollback state invalid: %#v %v", notApplied, inspectErr)
	}
	if reapplied, reapplyErr := gatewayrequestmigrations.Apply(ctx, adminPool); reapplyErr != nil || reapplied.MigrationState != gatewayrequestmigrations.MigrationStateApplied {
		t.Fatalf("Gateway request reapply failed: %#v %v", reapplied, reapplyErr)
	}
}

func resetPostgresGatewayRequestSchema(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	if _, err := pool.Exec(ctx, "DROP TABLE IF EXISTS gateway_request_records"); err != nil {
		t.Fatalf("reset Gateway request integration table: %v", err)
	}
	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS gateway_request_schema_versions (component text PRIMARY KEY, migration_id text NOT NULL, store_schema_version text NOT NULL, migration_checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		t.Fatalf("prepare Gateway request integration marker: %v", err)
	}
	if _, err := pool.Exec(ctx, "DELETE FROM gateway_request_schema_versions WHERE component=$1", gatewayrequestmigrations.Component); err != nil {
		t.Fatalf("reset Gateway request integration marker: %v", err)
	}
}

func installPostgresGatewayRequestV2Schema(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	initial, err := os.ReadFile("../../migrations/gateway_requests/0001_gateway_requests.up.sql")
	if err != nil {
		t.Fatalf("read Gateway request v1 migration: %v", err)
	}
	cost, err := os.ReadFile("../../migrations/gateway_requests/0002_gateway_request_cost_estimate.up.sql")
	if err != nil {
		t.Fatalf("read Gateway request v2 migration: %v", err)
	}
	if _, err = pool.Exec(ctx, string(initial)); err != nil {
		t.Fatalf("install Gateway request v1 schema: %v", err)
	}
	if _, err = pool.Exec(ctx, string(cost)); err != nil {
		t.Fatalf("install Gateway request v2 schema: %v", err)
	}
	checksum := sha256.Sum256([]byte(string(initial) + "\n" + string(cost)))
	if _, err = pool.Exec(ctx, `INSERT INTO gateway_request_schema_versions
        (component, migration_id, store_schema_version, migration_checksum)
        VALUES ($1, $2, $3, $4)`, gatewayrequestmigrations.Component,
		"0002_gateway_request_cost_estimate", "gateway_request_store_v2",
		fmt.Sprintf("sha256:%x", checksum)); err != nil {
		t.Fatalf("install Gateway request v2 marker: %v", err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO gateway_request_records (
        tenant_ref, workspace_id, consumer_ref, application_id, request_id, record_version,
        schema_version, store_mode, request_route, protocol, request_status, started_at,
        selected_provider, selected_profile, selected_model, failure_boundary,
        usage_availability, cost_availability, sanitized_request_record
    ) VALUES (
        'tenant_legacy', 'workspace_legacy', 'consumer_legacy', '', 'request_postgres_legacy_v2', 1,
        'gateway_request_record.v2', 'postgres_dev_test', '/v1/chat/completions',
        'openai-chat-completions', 'started', now(), 'mock', 'profile_legacy', 'model_legacy', '',
        'not_reported', 'usage_not_reported', '{"schema_version":"gateway_request_record.v2"}'::jsonb
    )`); err != nil {
		t.Fatalf("seed Gateway request v2 row: %v", err)
	}
}
