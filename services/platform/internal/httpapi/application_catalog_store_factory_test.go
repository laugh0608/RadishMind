package httpapi

import (
	"context"
	"path/filepath"
	"testing"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	applicationcatalogmigrations "radishmind.local/services/platform/migrations/sqlite/application_catalog_records"
)

func TestApplicationCatalogStoreFactoryModes(t *testing.T) {
	repository, closeRepository, err := newApplicationCatalogRepositoryFromConfig(config.Config{ApplicationCatalogStoreMode: "memory_dev"})
	if err != nil || repository == nil || closeRepository == nil {
		t.Fatalf("memory_dev store must be available: repository=%T err=%v", repository, err)
	}
	closeRepository()
	if _, ok := repository.(*memoryApplicationCatalogRepository); !ok {
		t.Fatalf("unexpected memory repository type: %T", repository)
	}

	if _, _, err := newApplicationCatalogRepositoryFromConfig(config.Config{ApplicationCatalogStoreMode: "unknown"}); err == nil {
		t.Fatal("unknown application catalog store mode must fail")
	}
	if _, _, err := newApplicationCatalogRepositoryFromConfig(config.Config{ApplicationCatalogStoreMode: "postgres_dev_test"}); err == nil {
		t.Fatal("incomplete postgres_dev_test configuration must fail before connecting")
	}
}

func TestApplicationCatalogStoreFactoryUsesSharedSQLiteRuntime(t *testing.T) {
	cfg := config.Config{
		ControlPlaneReadDevAuthEnabled:    true,
		ApplicationCatalogDevHTTPEnabled:  true,
		ApplicationCatalogDevWriteEnabled: true,
		ApplicationCatalogStoreMode:       "sqlite_dev",
	}
	if _, _, err := newApplicationCatalogRepositoryFromConfig(cfg); err == nil || err.Error() != "sqlite_dev application catalog requires the shared SQLite runtime" {
		t.Fatalf("sqlite_dev must reject a missing shared runtime, got %v", err)
	}

	withoutMigration, err := sqlitedev.Open(context.Background(), sqlitedev.Options{DatabasePath: filepath.Join(t.TempDir(), "without-migration.db")})
	if err != nil {
		t.Fatalf("open SQLite runtime without application catalog migration: %v", err)
	}
	defer func() { _ = withoutMigration.Close() }()
	if _, _, err := newApplicationCatalogRepositoryFromConfigWithSQLiteRuntime(cfg, withoutMigration); err == nil || err.Error() != "SQLite development component migration is not applied" {
		t.Fatalf("sqlite_dev must reject a missing application catalog migration, got %v", err)
	}

	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: filepath.Join(t.TempDir(), "radishmind.db"),
		Migrations:   applicationcatalogmigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open migrated SQLite runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()
	repository, closeRepository, err := newApplicationCatalogRepositoryFromConfigWithSQLiteRuntime(cfg, runtime)
	if err != nil || repository == nil || closeRepository == nil {
		t.Fatalf("construct application catalog SQLite repository: repository=%T err=%v", repository, err)
	}
	if _, ok := repository.(*sqliteApplicationCatalogRepository); !ok {
		t.Fatalf("unexpected SQLite repository type: %T", repository)
	}
	closeRepository()
	if err := runtime.DB().PingContext(context.Background()); err != nil {
		t.Fatalf("component close callback must not close the shared SQLite runtime: %v", err)
	}
}
