//go:build postgres_integration

package httpapi

import (
	"context"
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

	state, err := gatewayrequestmigrations.Apply(ctx, adminPool)
	if err != nil || state.MigrationState != gatewayrequestmigrations.MigrationStateApplied {
		t.Fatalf("apply Gateway request migration: %#v %v", state, err)
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
		if err = store.CreateRequest(requestContext, &record); err != nil {
			t.Fatal(err)
		}
		record.Status = GatewayRequestStatusSucceeded
		record.CompletedAt = base.Add(time.Duration(index)*time.Second + time.Millisecond).Format(time.RFC3339Nano)
		record.DurationMS = 1
		record.HTTPStatusCode = 200
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
	if err != nil || !found || recovered.Status != GatewayRequestStatusSucceeded || recovered.StoreMode != gatewayRequestStoreModePostgresDevTest {
		t.Fatalf("Gateway request restart recovery failed: found=%v record=%#v err=%v", found, recovered, err)
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
