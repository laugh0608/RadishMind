package httpapi

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	gatewayrequestmigrations "radishmind.local/services/platform/migrations/sqlite/gateway_requests"
)

func TestGatewayRequestSQLiteRepositoryContract(t *testing.T) {
	t.Run("scope filters and cursor", func(t *testing.T) {
		runtime := openGatewayRequestSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
		runGatewayRequestHistoryScopeFilterAndCursor(t, newSQLiteGatewayRequestStore(runtime.DB()))
	})
	t.Run("concurrent terminal update has one winner", func(t *testing.T) {
		runtime := openGatewayRequestSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
		runGatewayRequestStoreRejectsConcurrentTerminalRewrite(t, newSQLiteGatewayRequestStore(runtime.DB()))
	})
	t.Run("sensitive references and invalid usage are rejected", func(t *testing.T) {
		runtime := openGatewayRequestSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
		runGatewayRequestStoreRejectsSensitiveAndInvalidUsage(t, newSQLiteGatewayRequestStore(runtime.DB()))
	})
}

func TestGatewayRequestSQLiteRecordsNorthboundAndExcludesPayloads(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "radishmind.db")
	runtime := openGatewayRequestSQLiteRuntime(t, databasePath)
	server := newGatewayRequestHistoryHTTPTestServer()
	t.Cleanup(server.Close)
	server.gatewayRequestHistoryStore = newSQLiteGatewayRequestStore(runtime.DB())
	server.gatewayRequestHistoryStoreMode = gatewayRequestStoreModeSQLiteDev

	runGatewayRequestHistoryRecordsNorthboundAndReadsScopedHistory(t, server)

	var storeMode string
	var payload string
	if err := runtime.DB().QueryRowContext(context.Background(), `SELECT store_mode, sanitized_request_record
        FROM gateway_request_records WHERE request_id=?`, "request_gateway_success").Scan(&storeMode, &payload); err != nil {
		t.Fatalf("inspect Gateway request SQLite payload: %v", err)
	}
	if storeMode != gatewayRequestStoreModeSQLiteDev {
		t.Fatalf("Gateway request SQLite store mode drifted: %s", storeMode)
	}
	for _, forbidden := range []string{"private prompt must not persist", "bridge summary", "messages", "authorization"} {
		if strings.Contains(strings.ToLower(payload), strings.ToLower(forbidden)) {
			t.Fatalf("Gateway request SQLite payload contains forbidden material %q: %s", forbidden, payload)
		}
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("close Gateway request SQLite runtime: %v", err)
	}
	assertSQLiteFilesExcludeMarker(t, databasePath, "private prompt must not persist")
	assertSQLiteFilesExcludeMarker(t, databasePath, "bridge summary")
}

func TestGatewayRequestSQLiteSharesRuntimeWithAPIKeyAuthentication(t *testing.T) {
	migrations := apiKeySQLiteMigrations()
	migrations = append(migrations, gatewayrequestmigrations.Migrations()...)
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: filepath.Join(t.TempDir(), "radishmind.db"),
		Migrations:   migrations,
	})
	if err != nil {
		t.Fatalf("open shared API key and Gateway request SQLite runtime: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })
	historyStore := newSQLiteGatewayRequestStore(runtime.DB())
	runGatewayAPIKeyAuthenticationProtectsNorthboundWithHistoryStore(
		t,
		newSQLiteAPIKeyRepository(runtime.DB()),
		newSQLiteApplicationCatalogRepository(runtime.DB()),
		historyStore,
		gatewayRequestStoreModeSQLiteDev,
	)

	var count int
	var storeMode string
	var consumerRef string
	if err := runtime.DB().QueryRowContext(context.Background(), `SELECT count(*), min(store_mode), min(consumer_ref)
        FROM gateway_request_records`).Scan(&count, &storeMode, &consumerRef); err != nil {
		t.Fatalf("inspect shared SQLite Gateway request history: %v", err)
	}
	if count != 1 || storeMode != gatewayRequestStoreModeSQLiteDev || consumerRef != "api_key:key_aaaaaaaaaaaaaaaa" {
		t.Fatalf("shared SQLite API key history handoff drifted: count=%d mode=%s consumer=%s", count, storeMode, consumerRef)
	}
}

func TestGatewayRequestSQLiteCheckpointTerminalRestartAndNoFallback(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "radishmind.db")
	firstRuntime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   gatewayrequestmigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open first Gateway request SQLite runtime: %v", err)
	}
	store := newSQLiteGatewayRequestStore(firstRuntime.DB())
	requestContext := gatewayRequestTestContext()
	startedAt := time.Date(2026, 7, 14, 9, 30, 0, 123456789, time.UTC)
	record := gatewayRequestTestRecord(requestContext, "request_sqlite_restart", startedAt)
	if err := store.CreateRequest(requestContext, &record); err != nil || record.RecordVersion != 1 ||
		record.StoreMode != gatewayRequestStoreModeSQLiteDev {
		t.Fatalf("create Gateway request SQLite record: record=%#v err=%v", record, err)
	}
	record.Stream = true
	record.SelectionSource = "config"
	record.SelectedProvider = "mock"
	record.SelectedProfile = "profile_sqlite"
	record.SelectedModel = "model_sqlite"
	record.ProviderRouteConfigurationID = "gateway-default"
	record.ProviderRouteGeneration = 4
	record.ProviderRouteSnapshotDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if err := store.UpdateRequest(requestContext, &record); err != nil || record.RecordVersion != 2 ||
		record.Status != GatewayRequestStatusStarted {
		t.Fatalf("checkpoint Gateway request SQLite record: record=%#v err=%v", record, err)
	}

	record.Status = GatewayRequestStatusCanceled
	record.CompletedAt = startedAt.Add(2 * time.Second).Format(time.RFC3339Nano)
	record.DurationMS = 2000
	record.HTTPStatusCode = http.StatusRequestTimeout
	record.FailureCode = bridge.ErrorCodeWorkerCanceled
	record.FailureBoundary = errorBoundaryPythonBridge
	canceledParent, cancelParent := context.WithCancel(context.Background())
	cancelParent()
	canceledContext := requestContext
	canceledContext.RequestContext = canceledParent
	if updateErr := store.UpdateRequest(canceledContext, &record); !errors.Is(updateErr, errGatewayRequestStoreUnavailable) {
		t.Fatalf("canceled northbound context must not persist terminal state directly: %v", updateErr)
	}
	terminalServer := &Server{config: config.Config{GatewayRequestDatabaseTimeout: time.Second}}
	terminalContext, cancelTerminal := terminalServer.gatewayRequestTerminalStoreContext(canceledContext)
	if err := store.UpdateRequest(terminalContext, &record); err != nil {
		cancelTerminal()
		t.Fatalf("detached terminal context did not persist SQLite cancellation: %v", err)
	}
	cancelTerminal()
	if record.RecordVersion != 3 {
		t.Fatalf("Gateway request terminal update version drifted: %#v", record)
	}

	sensitiveMarker := "secret=gateway-history-sensitive-marker"
	rejected := gatewayRequestTestRecord(requestContext, "request_sqlite_rejected", startedAt.Add(time.Minute))
	rejected.SelectedProvider = sensitiveMarker
	if err := store.CreateRequest(requestContext, &rejected); !errors.Is(err, errGatewayRequestStoreContract) {
		t.Fatalf("sensitive Gateway request reference was accepted: %v", err)
	}
	var storedStartedAt int64
	var storedCompletedAt int64
	var storedPayload string
	if err := firstRuntime.DB().QueryRowContext(context.Background(), `SELECT started_at_unix_nano,
        completed_at_unix_nano, sanitized_request_record FROM gateway_request_records WHERE request_id=?`,
		record.RequestID).Scan(&storedStartedAt, &storedCompletedAt, &storedPayload); err != nil {
		t.Fatalf("inspect Gateway request SQLite physical record: %v", err)
	}
	if storedStartedAt != startedAt.UnixNano() || storedCompletedAt != startedAt.Add(2*time.Second).UnixNano() ||
		strings.Contains(storedPayload, sensitiveMarker) {
		t.Fatalf("Gateway request SQLite physical record drifted: started=%d completed=%d payload=%s", storedStartedAt, storedCompletedAt, storedPayload)
	}

	if err := firstRuntime.Close(); err != nil {
		t.Fatalf("close first Gateway request SQLite runtime: %v", err)
	}
	service := newGatewayRequestHistoryService(store)
	if result := service.Read(requestContext, record.RequestID); result.FailureCode != GatewayRequestHistoryFailureStore {
		t.Fatalf("closed Gateway request SQLite read must not fall back: %#v", result)
	}
	if result := service.List(requestContext, GatewayRequestListRequest{}); result.FailureCode != GatewayRequestHistoryFailureStore {
		t.Fatalf("closed Gateway request SQLite list must not fall back: %#v", result)
	}
	assertSQLiteFilesExcludeMarker(t, databasePath, sensitiveMarker)

	secondRuntime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   gatewayrequestmigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("reopen Gateway request SQLite runtime: %v", err)
	}
	t.Cleanup(func() { _ = secondRuntime.Close() })
	restartedStore := newSQLiteGatewayRequestStore(secondRuntime.DB())
	restored, found, err := restartedStore.ReadRequest(requestContext, record.RequestID)
	if err != nil || !found || restored.Status != GatewayRequestStatusCanceled || restored.RecordVersion != 3 ||
		restored.StoreMode != gatewayRequestStoreModeSQLiteDev || restored.SelectedProfile != "profile_sqlite" ||
		restored.ProviderRouteConfigurationID != "gateway-default" || restored.ProviderRouteGeneration != 4 ||
		restored.ProviderRouteSnapshotDigest != record.ProviderRouteSnapshotDigest {
		t.Fatalf("restore Gateway request SQLite record: found=%v record=%#v err=%v", found, restored, err)
	}
	page, err := restartedStore.ListRequests(requestContext, GatewayRequestListFilter{
		Limit: 10, Status: GatewayRequestStatusCanceled, FailureBoundary: errorBoundaryPythonBridge,
	})
	if err != nil || len(page.Records) != 1 || page.Records[0].RequestID != record.RequestID {
		t.Fatalf("list restored Gateway request SQLite record: page=%#v err=%v", page, err)
	}
	otherScope := requestContext
	otherScope.ConsumerRef = "consumer_other"
	if page, err := restartedStore.ListRequests(otherScope, GatewayRequestListFilter{Limit: 10}); err != nil || len(page.Records) != 0 {
		t.Fatalf("Gateway request SQLite restart leaked caller scope: page=%#v err=%v", page, err)
	}
	restored.Status = GatewayRequestStatusStarted
	restored.CompletedAt = ""
	restored.HTTPStatusCode = 0
	restored.FailureCode = ""
	restored.FailureBoundary = ""
	if err := restartedStore.UpdateRequest(requestContext, &restored); !errors.Is(err, errGatewayRequestStoreConflict) {
		t.Fatalf("terminal Gateway request record was rewritten after restart: %v", err)
	}
	if _, err := secondRuntime.DB().ExecContext(context.Background(), `UPDATE gateway_request_records
        SET sanitized_request_record='{"unexpected":true}' WHERE request_id=?`, record.RequestID); err != nil {
		t.Fatalf("prepare stored Gateway request contract mismatch: %v", err)
	}
	restartedService := newGatewayRequestHistoryService(restartedStore)
	if result := restartedService.Read(requestContext, record.RequestID); result.FailureCode != GatewayRequestHistoryFailureContract ||
		result.FailureSummary != "Gateway request history storage is unavailable." {
		t.Fatalf("stored Gateway request contract mismatch did not fail closed: %#v", result)
	}
	if result := restartedService.List(requestContext, GatewayRequestListRequest{}); result.FailureCode != GatewayRequestHistoryFailureContract ||
		len(result.Records) != 0 {
		t.Fatalf("stored Gateway request list contract mismatch returned partial data: %#v", result)
	}
}

func TestGatewayRequestSQLiteRejectsTimeOutsideNanosecondRange(t *testing.T) {
	if _, err := gatewayRequestUnixNano("2500-01-01T00:00:00Z"); err == nil {
		t.Fatal("Gateway request SQLite time outside the nanosecond range must be rejected")
	}
}

func TestGatewayProviderAttemptSQLiteFallbackPersistsAcrossRestart(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "gateway-provider-attempt.db")
	first := openGatewayRequestSQLiteRuntime(t, databasePath)
	ctx := gatewayRequestTestContext()
	plan := gatewayProviderAttemptTestPlan(t, "request-sqlite-fallback")
	root := gatewayRequestTestRecord(ctx, plan.RootRequestID, time.Date(2026, 8, 13, 9, 0, 0, 0, time.UTC))
	record, err := newGatewayProviderAttemptHistoryRecord(root, plan)
	store := newSQLiteGatewayRequestStore(first.DB())
	if err != nil || store.CreateRequest(ctx, &record) != nil {
		t.Fatalf("create SQLite v3 root: record=%#v err=%v", record, err)
	}
	service := newGatewayProviderAttemptHistoryService(store)
	now := time.Date(2026, 8, 13, 9, 0, 1, 0, time.UTC)
	if _, err = service.StartAttempt(ctx, plan.RootRequestID, plan.Targets[0], "quota-primary", now); err != nil {
		t.Fatalf("start SQLite primary attempt: %v", err)
	}
	failure := gatewayProviderAttemptTestFailure(
		t, bridge.ProviderFailureTemporarilyUnavailable, bridge.ProviderFallbackEligible, bridge.ProviderAttemptFailed,
	)
	if checkpoint, checkpointErr := service.CompleteAttempt(
		ctx, plan.RootRequestID, plan.Targets[0].AttemptID,
		GatewayRequestUsage{Availability: GatewayRequestUsageNotReported},
		gatewayRequestCostUnavailable(GatewayRequestCostUsageNotReported, "provider_usage_not_reported"),
		&failure, true, now.Add(time.Second),
	); checkpointErr != nil || checkpoint.ProviderAttemptPhase != GatewayProviderAttemptPhaseFallbackPending {
		t.Fatalf("persist SQLite fallback checkpoint: record=%#v err=%v", checkpoint, checkpointErr)
	}
	if err = first.Close(); err != nil {
		t.Fatalf("close first SQLite attempt runtime: %v", err)
	}

	restarted := openGatewayRequestSQLiteRuntime(t, databasePath)
	t.Cleanup(func() { _ = restarted.Close() })
	restartedStore := newSQLiteGatewayRequestStore(restarted.DB())
	restartedService := newGatewayProviderAttemptHistoryService(restartedStore)
	restored, found, err := restartedStore.ReadRequest(ctx, plan.RootRequestID)
	if err != nil || !found || restored.ProviderAttemptPhase != GatewayProviderAttemptPhaseFallbackPending ||
		len(restored.ProviderAttempts) != 1 || restored.ProviderAttempts[0].Failure == nil {
		t.Fatalf("restore SQLite fallback checkpoint: found=%v record=%#v err=%v", found, restored, err)
	}
	if _, err = restartedService.StartAttempt(
		ctx, plan.RootRequestID, plan.Targets[1], "quota-fallback", now.Add(2*time.Second),
	); err != nil {
		t.Fatalf("start SQLite fallback attempt: %v", err)
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
	if _, err = restartedService.CompleteAttempt(
		ctx, plan.RootRequestID, plan.Targets[1].AttemptID, usage, cost, nil, false, now.Add(3*time.Second),
	); err != nil {
		t.Fatalf("complete SQLite fallback attempt: %v", err)
	}
	finished, err := restartedService.Finalize(
		ctx, plan.RootRequestID, GatewayRequestStatusSucceeded, http.StatusOK, "", "", now.Add(4*time.Second),
	)
	if err != nil || !finished.FallbackUsed || finished.ProviderAttemptCount != 2 ||
		finished.TerminalAttemptID != plan.Targets[1].AttemptID {
		t.Fatalf("finalize SQLite fallback attempt: record=%#v err=%v", finished, err)
	}
	var attemptCount int
	var fallbackUsed bool
	var terminalProvider string
	var terminalProfile string
	if err = restarted.DB().QueryRowContext(context.Background(), `SELECT provider_attempt_count, fallback_used,
        terminal_provider, terminal_profile FROM gateway_request_records WHERE request_id=?`, plan.RootRequestID).
		Scan(&attemptCount, &fallbackUsed, &terminalProvider, &terminalProfile); err != nil {
		t.Fatalf("inspect SQLite attempt summary columns: %v", err)
	}
	if attemptCount != 2 || !fallbackUsed || terminalProvider != plan.Targets[1].ProviderID ||
		terminalProfile != plan.Targets[1].RuntimeProfile {
		t.Fatalf("SQLite attempt summary drifted: count=%d fallback=%v provider=%s profile=%s",
			attemptCount, fallbackUsed, terminalProvider, terminalProfile)
	}
}

func TestGatewayProviderAttemptSQLiteCheckpointHasOneConcurrentWinner(t *testing.T) {
	runtime := openGatewayRequestSQLiteRuntime(t, filepath.Join(t.TempDir(), "gateway-provider-attempt-cas.db"))
	ctx := gatewayRequestTestContext()
	plan := gatewayProviderAttemptTestPlan(t, "request-sqlite-attempt-cas")
	root := gatewayRequestTestRecord(ctx, plan.RootRequestID, time.Now().UTC())
	record, err := newGatewayProviderAttemptHistoryRecord(root, plan)
	store := newSQLiteGatewayRequestStore(runtime.DB())
	if err != nil || store.CreateRequest(ctx, &record) != nil {
		t.Fatalf("create concurrent SQLite v3 root: %v", err)
	}
	service := newGatewayProviderAttemptHistoryService(store)
	start := make(chan struct{})
	results := make(chan error, 2)
	var wait sync.WaitGroup
	for index := 0; index < 2; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			_, checkpointErr := service.StartAttempt(
				ctx, plan.RootRequestID, plan.Targets[0], "quota-primary", time.Now().UTC(),
			)
			results <- checkpointErr
		}()
	}
	close(start)
	wait.Wait()
	close(results)
	successes := 0
	conflicts := 0
	for result := range results {
		if result == nil {
			successes++
		} else if errors.Is(result, errGatewayRequestStoreConflict) {
			conflicts++
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("SQLite attempt checkpoint CAS drifted: successes=%d conflicts=%d", successes, conflicts)
	}
}

func openGatewayRequestSQLiteRuntime(t *testing.T, databasePath string) *sqlitedev.Runtime {
	t.Helper()
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   gatewayrequestmigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open Gateway request SQLite runtime: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })
	return runtime
}
