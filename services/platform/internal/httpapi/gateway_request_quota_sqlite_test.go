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
	gatewayrequestquotamigrations "radishmind.local/services/platform/migrations/sqlite/gateway_request_quotas"
)

func TestGatewayRequestQuotaSQLiteRepositoryPersistsAndAdmitsAtomically(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "radishmind.db")
	runtime := openGatewayRequestQuotaSQLiteRuntime(t, databasePath)
	repository := newSQLiteGatewayRequestQuotaRepository(runtime.DB())
	configuredRepository, closeConfiguredRepository, err := newGatewayRequestQuotaRepositoryFromConfigWithSQLiteRuntime(config.Config{
		GatewayRequestQuotaStoreMode:       gatewayRequestQuotaStoreModeSQLiteDev,
		GatewayRequestQuotaDatabaseTimeout: 5 * time.Second,
	}, runtime)
	if err != nil {
		t.Fatalf("select configured SQLite quota repository: %v", err)
	}
	defer closeConfiguredRepository()
	if _, ok := configuredRepository.(*sqliteGatewayRequestQuotaRepository); !ok {
		t.Fatalf("SQLite quota selector returned %T", configuredRepository)
	}
	quotaContext := testGatewayRequestQuotaContext("app-sqlite")
	now := time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC)
	policy, err := repository.PutPolicy(quotaContext, 0, 1, now)
	if err != nil || policy.RecordVersion != 1 {
		t.Fatalf("put SQLite quota policy: policy=%+v err=%v", policy, err)
	}

	start := make(chan struct{})
	results := make(chan error, 2)
	var waitGroup sync.WaitGroup
	for attempt := 0; attempt < 2; attempt++ {
		waitGroup.Add(1)
		go func(attempt int) {
			defer waitGroup.Done()
			<-start
			input := testQuotaAdmission("request-sqlite-"+string(rune('a'+attempt)), now)
			_, attemptErr := repository.AdmitProviderAttempt(quotaContext, input)
			results <- attemptErr
		}(attempt)
	}
	close(start)
	waitGroup.Wait()
	close(results)
	admitted, exceeded := 0, 0
	for attemptErr := range results {
		switch {
		case attemptErr == nil:
			admitted++
		case errors.Is(attemptErr, errGatewayRequestQuotaExceeded):
			exceeded++
		default:
			t.Fatalf("unexpected SQLite quota admission error: %v", attemptErr)
		}
	}
	if admitted != 1 || exceeded != 1 {
		t.Fatalf("expected one SQLite admission and one rejection, got %d/%d", admitted, exceeded)
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("close first SQLite quota runtime: %v", err)
	}

	restarted := openGatewayRequestQuotaSQLiteRuntime(t, databasePath)
	restored := newSQLiteGatewayRequestQuotaRepository(restarted.DB())
	usage, found, err := restored.ReadUsage(quotaContext, "2026-08-09")
	if err != nil || !found || usage.AdmittedRequestCount != 1 || usage.RemainingRequestCount != 0 {
		t.Fatalf("restore SQLite quota usage: found=%t usage=%+v err=%v", found, usage, err)
	}
	if _, err := restored.AdmitProviderAttempt(quotaContext, testQuotaAdmission("request-after-restart", now)); !errors.Is(err, errGatewayRequestQuotaExceeded) {
		t.Fatalf("restarted SQLite quota did not fail closed: %v", err)
	}
}

func TestGatewayRequestQuotaStoreSelectorRejectsUnknownModeWithoutFallback(t *testing.T) {
	repository, closeRepository, err := newGatewayRequestQuotaRepositoryFromConfigWithSQLiteRuntime(config.Config{
		GatewayRequestQuotaStoreMode:       "unknown",
		GatewayRequestQuotaDatabaseTimeout: 5 * time.Second,
	}, nil)
	if err == nil || repository != nil || closeRepository != nil {
		t.Fatalf("unknown quota store mode did not fail closed: repository=%T close=%v err=%v", repository, closeRepository != nil, err)
	}
}

func openGatewayRequestQuotaSQLiteRuntime(t *testing.T, databasePath string) *sqlitedev.Runtime {
	t.Helper()
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   gatewayrequestquotamigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open gateway request quota SQLite runtime: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })
	return runtime
}
