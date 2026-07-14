package httpapi

import (
	"context"
	"path/filepath"
	"testing"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	apikeymigrations "radishmind.local/services/platform/migrations/sqlite/api_key_records"
	applicationcatalogmigrations "radishmind.local/services/platform/migrations/sqlite/application_catalog_records"
)

func TestAPIKeyStoreFactoryModes(t *testing.T) {
	repository, closeRepository, err := newAPIKeyRepositoryFromConfig(config.Config{APIKeyStoreMode: "memory_dev"})
	if err != nil || repository == nil || closeRepository == nil {
		t.Fatalf("memory_dev API key repository selection failed: repository=%T err=%v", repository, err)
	}
	closeRepository()
	if _, _, err := newAPIKeyRepositoryFromConfig(config.Config{APIKeyStoreMode: "unknown"}); err == nil {
		t.Fatal("unknown API key store mode must fail")
	}
	if _, _, err := newAPIKeyRepositoryFromConfig(config.Config{APIKeyStoreMode: "postgres_dev_test"}); err == nil {
		t.Fatal("incomplete postgres_dev_test API key configuration must fail before connecting")
	}
}

func TestAPIKeySQLiteFactoryRequiresSharedRuntimeAndDependencyMigrations(t *testing.T) {
	cfg := apiKeySQLiteFactoryConfig()
	if _, _, err := newAPIKeyRepositoryFromConfig(cfg); err == nil || err.Error() != "sqlite_dev API key requires the shared SQLite runtime" {
		t.Fatalf("API key must reject a missing shared SQLite runtime, got %v", err)
	}

	withoutMigrations, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: filepath.Join(t.TempDir(), "without-migrations.db"),
	})
	if err != nil {
		t.Fatalf("open SQLite runtime without component migrations: %v", err)
	}
	t.Cleanup(func() { _ = withoutMigrations.Close() })
	if _, _, err := newAPIKeyRepositoryFromConfigWithSQLiteRuntime(cfg, withoutMigrations); err == nil || err.Error() != "SQLite development component migration is not applied" {
		t.Fatalf("API key must reject missing migrations, got %v", err)
	}

	catalogOnly, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: filepath.Join(t.TempDir(), "catalog-only.db"),
		Migrations:   applicationcatalogmigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open SQLite runtime with only application catalog migration: %v", err)
	}
	t.Cleanup(func() { _ = catalogOnly.Close() })
	if _, _, err := newAPIKeyRepositoryFromConfigWithSQLiteRuntime(cfg, catalogOnly); err == nil || err.Error() != "SQLite development component migration is not applied" {
		t.Fatalf("API key must reject missing API key migration, got %v", err)
	}

	apiKeyOnly, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: filepath.Join(t.TempDir(), "api-key-only.db"),
		Migrations:   apikeymigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open SQLite runtime with only API key migration: %v", err)
	}
	t.Cleanup(func() { _ = apiKeyOnly.Close() })
	if _, _, err := newAPIKeyRepositoryFromConfigWithSQLiteRuntime(cfg, apiKeyOnly); err == nil || err.Error() != "SQLite development component migration is not applied" {
		t.Fatalf("API key must reject missing application catalog migration, got %v", err)
	}
}

func TestAPIKeySQLiteFactorySharesRuntimeOwnership(t *testing.T) {
	runtime := openAPIKeySQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
	repository, closeRepository, err := newAPIKeyRepositoryFromConfigWithSQLiteRuntime(apiKeySQLiteFactoryConfig(), runtime)
	if err != nil || closeRepository == nil {
		t.Fatalf("construct API key SQLite repository: repository=%T err=%v", repository, err)
	}
	sqliteRepository, ok := repository.(*sqliteAPIKeyRepository)
	if !ok || sqliteRepository.database != runtime.DB() {
		t.Fatalf("API key repository did not receive the shared runtime: %T", repository)
	}
	closeRepository()
	if err := runtime.DB().PingContext(context.Background()); err != nil {
		t.Fatalf("API key close callback must not close the shared SQLite runtime: %v", err)
	}
}

func TestAPIKeySQLiteFactoryRequiresSQLiteApplicationCatalog(t *testing.T) {
	cfg := apiKeySQLiteFactoryConfig()
	cfg.ApplicationCatalogStoreMode = "memory_dev"
	if _, _, err := newAPIKeyRepositoryFromConfigWithSQLiteRuntime(cfg, nil); err == nil || err.Error() != "sqlite_dev API key config is incomplete" {
		t.Fatalf("API key must reject a different application catalog store mode, got %v", err)
	}
}

func apiKeySQLiteFactoryConfig() config.Config {
	return config.Config{
		ControlPlaneReadDevAuthEnabled:    true,
		ApplicationCatalogDevHTTPEnabled:  true,
		ApplicationCatalogDevWriteEnabled: true,
		ApplicationCatalogStoreMode:       "sqlite_dev",
		APIKeyLifecycleDevHTTPEnabled:     true,
		APIKeyLifecycleDevWriteEnabled:    true,
		APIKeyStoreMode:                   "sqlite_dev",
	}
}
