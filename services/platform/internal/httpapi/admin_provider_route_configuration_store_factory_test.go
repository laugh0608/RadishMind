package httpapi

import (
	"context"
	"path/filepath"
	"testing"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	sqliteadminproviderroutemigrations "radishmind.local/services/platform/migrations/sqlite/admin_provider_routes"
)

func TestAdminProviderRouteRepositorySelectorIsMutuallyExclusive(t *testing.T) {
	memory, closeMemory, err := newAdminProviderRouteRepositoryFromConfig(config.Config{
		AdminProviderRouteStoreMode: "memory_dev",
	})
	if err != nil || closeMemory == nil {
		t.Fatalf("select memory repository: repository=%T err=%v", memory, err)
	}
	if _, ok := memory.(*memoryAdminProviderRouteRepository); !ok {
		t.Fatalf("unexpected memory repository: %T", memory)
	}
	closeMemory()

	cfg := config.Config{AdminProviderRouteStoreMode: "sqlite_dev"}
	if _, _, err := newAdminProviderRouteRepositoryFromConfig(cfg); err == nil ||
		err.Error() != "sqlite_dev admin provider route requires the shared SQLite runtime" {
		t.Fatalf("SQLite selector must require shared runtime, got %v", err)
	}
	withoutMigration, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: filepath.Join(t.TempDir(), "without-migration.db"),
	})
	if err != nil {
		t.Fatalf("open SQLite runtime without component migration: %v", err)
	}
	t.Cleanup(func() { _ = withoutMigration.Close() })
	if _, _, err := newAdminProviderRouteRepositoryFromConfigWithSQLiteRuntime(cfg, withoutMigration); err == nil || err.Error() != "SQLite development component migration is not applied" {
		t.Fatalf("SQLite selector must reject missing component migration, got %v", err)
	}

	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: filepath.Join(t.TempDir(), "with-migration.db"),
		Migrations:   sqliteadminproviderroutemigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open migrated SQLite runtime: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })
	sqliteRepository, closeSQLite, err := newAdminProviderRouteRepositoryFromConfigWithSQLiteRuntime(cfg, runtime)
	if err != nil || closeSQLite == nil {
		t.Fatalf("select SQLite repository: repository=%T err=%v", sqliteRepository, err)
	}
	if selected, ok := sqliteRepository.(*sqliteAdminProviderRouteRepository); !ok ||
		selected.database != runtime.DB() {
		t.Fatalf("unexpected SQLite repository: %T", sqliteRepository)
	}
	closeSQLite()
	if err := runtime.DB().PingContext(context.Background()); err != nil {
		t.Fatalf("component close must not close shared SQLite runtime: %v", err)
	}

	if _, _, err := newAdminProviderRouteRepositoryFromConfig(config.Config{
		AdminProviderRouteStoreMode: "repository_disabled",
	}); err == nil || err.Error() != "unsupported admin provider route store mode" {
		t.Fatalf("unknown or disabled mode must fail without memory fallback, got %v", err)
	}
}
