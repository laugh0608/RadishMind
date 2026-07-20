package httpapi

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	gatewayrequestmigrations "radishmind.local/services/platform/migrations/gateway_requests"
	sqlitegatewayrequestmigrations "radishmind.local/services/platform/migrations/sqlite/gateway_requests"
)

func newGatewayRequestStoreFromConfig(cfg config.Config) (gatewayRequestStore, string, func(), error) {
	return newGatewayRequestStoreFromConfigWithSQLiteRuntime(cfg, nil)
}

func newGatewayRequestStoreFromConfigWithSQLiteRuntime(
	cfg config.Config,
	sqliteRuntime *sqlitedev.Runtime,
) (gatewayRequestStore, string, func(), error) {
	mode := strings.TrimSpace(cfg.GatewayRequestStoreMode)
	if mode == "" || mode == gatewayRequestStoreModeMemoryDev {
		return newMemoryGatewayRequestStore(defaultGatewayRequestStoreCapacity), gatewayRequestStoreModeMemoryDev, func() {}, nil
	}
	if mode == gatewayRequestStoreModeSQLiteDev {
		if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.GatewayRequestHistoryDevEnabled {
			return nil, mode, nil, errors.New("sqlite_dev Gateway request store config is incomplete")
		}
		if sqliteRuntime == nil || sqliteRuntime.DB() == nil {
			return nil, mode, nil, errors.New("sqlite_dev Gateway request store requires the shared SQLite runtime")
		}
		timeout := cfg.GatewayRequestDatabaseTimeout
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := sqliteRuntime.VerifyMigrations(ctx, sqlitegatewayrequestmigrations.Migrations()); err != nil {
			return nil, mode, nil, err
		}
		return newSQLiteGatewayRequestStore(sqliteRuntime.DB()), mode, func() {}, nil
	}
	if mode != gatewayRequestStoreModePostgresDevTest {
		if mode == "repository" || mode == "repository_disabled" {
			return nil, mode, nil, fmt.Errorf("%s", GatewayRequestHistoryFailureStoreModeDisabled)
		}
		return nil, mode, nil, fmt.Errorf("%s", GatewayRequestHistoryFailureStoreModeInvalid)
	}
	if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.GatewayRequestHistoryDevEnabled || strings.TrimSpace(cfg.GatewayRequestDatabaseURL) == "" {
		return nil, mode, nil, errors.New("postgres_dev_test Gateway request store config is incomplete")
	}
	timeout := cfg.GatewayRequestDatabaseTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	pool, err := gatewayrequestmigrations.OpenPool(ctx, cfg.GatewayRequestDatabaseURL)
	if err != nil {
		return nil, mode, nil, err
	}
	closePool := func() { pool.Close() }
	state, err := gatewayrequestmigrations.Inspect(ctx, pool)
	if err != nil {
		closePool()
		return nil, mode, nil, err
	}
	if state.MigrationState != gatewayrequestmigrations.MigrationStateApplied || state.StoreSchemaVersion != gatewayrequestmigrations.StoreSchemaVersion {
		closePool()
		return nil, mode, nil, errors.New("Gateway request PostgreSQL migration is not applied or incompatible")
	}
	return newPostgresGatewayRequestStore(pool), mode, closePool, nil
}
