package httpapi

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/sqlitedev"
	applicationcatalogmigrations "radishmind.local/services/platform/migrations/sqlite/application_catalog_records"
)

func TestApplicationCatalogSQLiteRepositoryContract(t *testing.T) {
	t.Run("lifecycle and owner isolation", func(t *testing.T) {
		runtime := openApplicationCatalogSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
		runApplicationCatalogLifecycleAndOwnerIsolation(t, newSQLiteApplicationCatalogRepository(runtime.DB()))
	})
	t.Run("validation pagination and CAS", func(t *testing.T) {
		runtime := openApplicationCatalogSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
		runApplicationCatalogValidationPaginationAndCAS(t, newSQLiteApplicationCatalogRepository(runtime.DB()))
	})
}

func TestApplicationCatalogSQLiteRestartRecoveryAndNoFallback(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "radishmind.db")
	firstRuntime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   applicationcatalogmigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open first application catalog SQLite runtime: %v", err)
	}
	firstRepository := newSQLiteApplicationCatalogRepository(firstRuntime.DB())
	service := newApplicationCatalogService(firstRepository)
	service.newID = func() (string, error) { return "app_aaaaaaaaaaaaaaaa", nil }
	service.now = func() time.Time { return time.Date(2026, 7, 14, 9, 0, 0, 0, time.UTC) }
	requestContext := applicationCatalogTestContext("subject_owner")
	created := service.Create(requestContext, ApplicationCatalogCreateInput{
		DisplayName: "SQLite App", Description: "Persistent application catalog", ApplicationKind: "agent",
	})
	if created.Record == nil || created.FailureCode != "" {
		t.Fatalf("create persistent application catalog record: %#v", created)
	}
	if err := firstRuntime.Close(); err != nil {
		t.Fatalf("close first application catalog SQLite runtime: %v", err)
	}
	if result := service.Read(requestContext, created.Record.ApplicationID); result.FailureCode != ApplicationCatalogFailureStoreUnavailable {
		t.Fatalf("closed SQLite store must fail without memory fallback: %#v", result)
	}

	secondRuntime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   applicationcatalogmigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("reopen application catalog SQLite runtime: %v", err)
	}
	t.Cleanup(func() { _ = secondRuntime.Close() })
	restarted := newApplicationCatalogService(newSQLiteApplicationCatalogRepository(secondRuntime.DB()))
	restored := restarted.Read(requestContext, created.Record.ApplicationID)
	if restored.Record == nil || restored.FailureCode != "" || restored.Record.DisplayName != "SQLite App" || restored.Record.RecordVersion != 1 {
		t.Fatalf("restore application catalog record after SQLite restart: %#v", restored)
	}
	otherOwner := restarted.Read(applicationCatalogTestContext("subject_other"), created.Record.ApplicationID)
	if otherOwner.FailureCode != ApplicationCatalogFailureNotFound {
		t.Fatalf("restart must preserve owner isolation: %#v", otherOwner)
	}
}

func openApplicationCatalogSQLiteRuntime(t *testing.T, databasePath string) *sqlitedev.Runtime {
	t.Helper()
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   applicationcatalogmigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open application catalog SQLite runtime: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })
	return runtime
}
