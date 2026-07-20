package httpapi

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	sqlitegatewayrequestmigrations "radishmind.local/services/platform/migrations/sqlite/gateway_requests"
)

func TestGatewayRequestStoreFactoryModesFailClosed(t *testing.T) {
	for _, mode := range []string{"repository", "repository_disabled", "unknown"} {
		store, _, closeStore, err := newGatewayRequestStoreFromConfig(config.Config{GatewayRequestStoreMode: mode})
		if err == nil || store != nil || closeStore != nil {
			t.Fatalf("mode %s must fail closed: store=%T close=%v err=%v", mode, store, closeStore != nil, err)
		}
	}
	store, mode, closeStore, err := newGatewayRequestStoreFromConfig(config.Config{})
	if err != nil || mode != gatewayRequestStoreModeMemoryDev || closeStore == nil {
		t.Fatalf("default memory mode failed: store=%T mode=%s err=%v", store, mode, err)
	}
	closeStore()
}

func TestGatewayRequestSQLiteFactoryRequiresSharedRuntimeAndMigration(t *testing.T) {
	cfg := gatewayRequestSQLiteFactoryConfig()
	if store, mode, closeStore, err := newGatewayRequestStoreFromConfig(cfg); err == nil || store != nil ||
		closeStore != nil || mode != gatewayRequestStoreModeSQLiteDev ||
		err.Error() != "sqlite_dev Gateway request store requires the shared SQLite runtime" {
		t.Fatalf("missing shared SQLite runtime was accepted: store=%T mode=%s close=%v err=%v", store, mode, closeStore != nil, err)
	}

	withoutMigration, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: filepath.Join(t.TempDir(), "without-gateway-request-migration.db"),
	})
	if err != nil {
		t.Fatalf("open SQLite runtime without Gateway request migration: %v", err)
	}
	t.Cleanup(func() { _ = withoutMigration.Close() })
	if store, mode, closeStore, factoryErr := newGatewayRequestStoreFromConfigWithSQLiteRuntime(cfg, withoutMigration); factoryErr == nil || store != nil || closeStore != nil || mode != gatewayRequestStoreModeSQLiteDev ||
		factoryErr.Error() != "SQLite development component migration is not applied" {
		t.Fatalf("missing Gateway request migration was accepted: store=%T mode=%s close=%v err=%v", store, mode, closeStore != nil, factoryErr)
	}
}

func TestGatewayRequestSQLiteFactorySharesRuntimeOwnership(t *testing.T) {
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: filepath.Join(t.TempDir(), "radishmind.db"),
		Migrations:   sqlitegatewayrequestmigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open Gateway request SQLite runtime: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })
	store, mode, closeStore, err := newGatewayRequestStoreFromConfigWithSQLiteRuntime(
		gatewayRequestSQLiteFactoryConfig(), runtime,
	)
	if err != nil || mode != gatewayRequestStoreModeSQLiteDev || closeStore == nil {
		t.Fatalf("construct Gateway request SQLite store: store=%T mode=%s err=%v", store, mode, err)
	}
	sqliteStore, ok := store.(*sqliteGatewayRequestStore)
	if !ok || sqliteStore.database != runtime.DB() {
		t.Fatalf("Gateway request store did not receive shared SQLite runtime: %T", store)
	}
	closeStore()
	if err := runtime.DB().PingContext(context.Background()); err != nil {
		t.Fatalf("Gateway request component close must not close shared SQLite runtime: %v", err)
	}
}

func TestGatewayRequestSQLiteFactoryRequiresDevelopmentGates(t *testing.T) {
	for _, cfg := range []config.Config{
		{GatewayRequestStoreMode: gatewayRequestStoreModeSQLiteDev},
		{GatewayRequestStoreMode: gatewayRequestStoreModeSQLiteDev, ControlPlaneReadDevAuthEnabled: true},
	} {
		store, mode, closeStore, err := newGatewayRequestStoreFromConfigWithSQLiteRuntime(cfg, nil)
		if err == nil || store != nil || closeStore != nil || mode != gatewayRequestStoreModeSQLiteDev ||
			err.Error() != "sqlite_dev Gateway request store config is incomplete" {
			t.Fatalf("incomplete SQLite config was accepted: store=%T mode=%s close=%v err=%v", store, mode, closeStore != nil, err)
		}
	}
}

func TestGatewayRequestStoreFactoryPostgresNoFallback(t *testing.T) {
	cfg := config.Config{
		ControlPlaneReadDevAuthEnabled:  true,
		GatewayRequestHistoryDevEnabled: true,
		GatewayRequestStoreMode:         gatewayRequestStoreModePostgresDevTest,
		GatewayRequestDatabaseURL:       "postgresql://invalid:invalid@127.0.0.1:1/invalid",
		GatewayRequestDatabaseTimeout:   50 * time.Millisecond,
	}
	store, mode, closeStore, err := newGatewayRequestStoreFromConfig(cfg)
	if err == nil || store != nil || closeStore != nil || mode != gatewayRequestStoreModePostgresDevTest {
		t.Fatalf("PostgreSQL connection failure must not fall back: store=%T mode=%s close=%v err=%v", store, mode, closeStore != nil, err)
	}
	if strings.Contains(err.Error(), "invalid:invalid") {
		t.Fatalf("database error leaked credentials: %v", err)
	}
}

func TestGatewayRequestStoreFactoryPostgresRequiresAllDevInputs(t *testing.T) {
	configurations := []config.Config{
		{GatewayRequestStoreMode: gatewayRequestStoreModePostgresDevTest},
		{GatewayRequestStoreMode: gatewayRequestStoreModePostgresDevTest, ControlPlaneReadDevAuthEnabled: true},
		{GatewayRequestStoreMode: gatewayRequestStoreModePostgresDevTest, ControlPlaneReadDevAuthEnabled: true, GatewayRequestHistoryDevEnabled: true},
	}
	for _, cfg := range configurations {
		store, _, closeStore, err := newGatewayRequestStoreFromConfig(cfg)
		if err == nil || store != nil || closeStore != nil {
			t.Fatalf("incomplete PostgreSQL config was accepted: %#v", cfg)
		}
	}
}

func gatewayRequestSQLiteFactoryConfig() config.Config {
	return config.Config{
		ControlPlaneReadDevAuthEnabled:  true,
		GatewayRequestHistoryDevEnabled: true,
		GatewayRequestStoreMode:         gatewayRequestStoreModeSQLiteDev,
		GatewayRequestDatabaseTimeout:   time.Second,
	}
}
