package httpapi

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	gatewaymodelpricingmigrations "radishmind.local/services/platform/migrations/sqlite/gateway_model_pricing"
)

func TestGatewayModelPricingSQLiteRepositoryPersistsImmutableRevisions(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "radishmind.db")
	runtime := openGatewayModelPricingSQLiteRuntime(t, databasePath)
	repository := newSQLiteGatewayModelPricingRepository(runtime.DB())
	pricingContext := gatewayModelPricingTestContext("model_sqlite")
	first := putGatewayModelPricingTestPolicy(t, repository, pricingContext, 0, 10, 20, "first SQLite evidence")
	second := putGatewayModelPricingTestPolicy(t, repository, pricingContext, 1, 30, 40, "second SQLite evidence")
	if first.RecordVersion != 1 || second.RecordVersion != 2 {
		t.Fatalf("unexpected SQLite revisions: first=%+v second=%+v", first, second)
	}
	if _, err := runtime.DB().Exec(`UPDATE gateway_model_pricing_revisions SET policy_digest='sha256:tampered'
        WHERE tenant_ref=? AND policy_id=? AND record_version=1`, pricingContext.TenantRef, first.PolicyID); err == nil {
		t.Fatal("SQLite append-only revision accepted an update")
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("close first SQLite pricing runtime: %v", err)
	}

	restarted := openGatewayModelPricingSQLiteRuntime(t, databasePath)
	restored := newSQLiteGatewayModelPricingRepository(restarted.DB())
	current, found, err := restored.ReadCurrent(pricingContext)
	if err != nil || !found || current != second {
		t.Fatalf("restore SQLite current policy: policy=%+v found=%v err=%v", current, found, err)
	}
	retained, found, err := restored.ReadRevision(pricingContext, 1)
	if err != nil || !found || retained != first {
		t.Fatalf("restore SQLite first revision: policy=%+v found=%v err=%v", retained, found, err)
	}
}

func TestGatewayModelPricingSQLiteRepositoryCASHasSingleWinner(t *testing.T) {
	runtime := openGatewayModelPricingSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
	repository := newSQLiteGatewayModelPricingRepository(runtime.DB())
	pricingContext := gatewayModelPricingTestContext("model_sqlite_cas")
	putGatewayModelPricingTestPolicy(t, repository, pricingContext, 0, 10, 20, "initial SQLite CAS evidence")

	start := make(chan struct{})
	results := make(chan error, 8)
	var waitGroup sync.WaitGroup
	for index := 0; index < 8; index++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			_, err := repository.PutRevision(pricingContext, GatewayModelPricingPolicyInput{
				ExpectedVersion: 1, Currency: GatewayModelPricingCurrency,
				InputPriceMicrosPerTokenUnit: 30, OutputPriceMicrosPerTokenUnit: 40,
				Reason: "concurrent SQLite CAS evidence",
			}, time.Now().UTC())
			results <- err
		}()
	}
	close(start)
	waitGroup.Wait()
	close(results)
	winners, conflicts := 0, 0
	for result := range results {
		switch {
		case result == nil:
			winners++
		case errors.Is(result, errGatewayModelPricingVersionConflict):
			conflicts++
		default:
			t.Fatalf("unexpected SQLite CAS result: %v", result)
		}
	}
	if winners != 1 || conflicts != 7 {
		t.Fatalf("SQLite CAS did not produce one winner: winners=%d conflicts=%d", winners, conflicts)
	}
}

func TestGatewayModelPricingStoreSelectorUsesSharedSQLiteAndRejectsUnknownMode(t *testing.T) {
	runtime := openGatewayModelPricingSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
	repository, closeRepository, err := newGatewayModelPricingRepositoryFromConfigWithSQLiteRuntime(config.Config{
		GatewayModelPricingStoreMode:       gatewayModelPricingStoreModeSQLiteDev,
		GatewayModelPricingDatabaseTimeout: 5 * time.Second,
	}, runtime)
	if err != nil {
		t.Fatalf("select configured SQLite pricing repository: %v", err)
	}
	defer closeRepository()
	if _, ok := repository.(*sqliteGatewayModelPricingRepository); !ok {
		t.Fatalf("SQLite pricing selector returned %T", repository)
	}
	unknown, closeUnknown, err := newGatewayModelPricingRepositoryFromConfigWithSQLiteRuntime(config.Config{
		GatewayModelPricingStoreMode: "unknown",
	}, nil)
	if err == nil || unknown != nil || closeUnknown != nil {
		t.Fatalf("unknown pricing store mode did not fail closed: repository=%T close=%v err=%v", unknown, closeUnknown != nil, err)
	}
}

func openGatewayModelPricingSQLiteRuntime(t *testing.T, databasePath string) *sqlitedev.Runtime {
	t.Helper()
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   gatewaymodelpricingmigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open gateway model pricing SQLite runtime: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })
	return runtime
}
