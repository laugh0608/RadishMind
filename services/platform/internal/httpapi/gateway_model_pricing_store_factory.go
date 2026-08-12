package httpapi

import (
	"context"
	"errors"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	gatewaymodelpricingmigrations "radishmind.local/services/platform/migrations/gateway_model_pricing"
	sqlitegatewaymodelpricingmigrations "radishmind.local/services/platform/migrations/sqlite/gateway_model_pricing"
)

const (
	gatewayModelPricingStoreModeMemoryDev       = "memory_dev"
	gatewayModelPricingStoreModeSQLiteDev       = "sqlite_dev"
	gatewayModelPricingStoreModePostgresDevTest = "postgres_dev_test"
)

func newGatewayModelPricingRepositoryFromConfigWithSQLiteRuntime(
	cfg config.Config,
	sqliteRuntime *sqlitedev.Runtime,
) (GatewayModelPricingRepository, func(), error) {
	mode := strings.TrimSpace(cfg.GatewayModelPricingStoreMode)
	if mode == "" || mode == gatewayModelPricingStoreModeMemoryDev {
		return newMemoryGatewayModelPricingRepository(), func() {}, nil
	}
	timeout := cfg.GatewayModelPricingDatabaseTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if mode == gatewayModelPricingStoreModeSQLiteDev {
		if sqliteRuntime == nil || sqliteRuntime.DB() == nil {
			return nil, nil, errors.New("sqlite_dev gateway model pricing requires the shared SQLite runtime")
		}
		if err := sqliteRuntime.VerifyMigrations(ctx, sqlitegatewaymodelpricingmigrations.Migrations()); err != nil {
			return nil, nil, err
		}
		return newSQLiteGatewayModelPricingRepository(sqliteRuntime.DB()), func() {}, nil
	}
	if mode != gatewayModelPricingStoreModePostgresDevTest {
		return nil, nil, errors.New("unsupported gateway model pricing store mode")
	}
	if strings.TrimSpace(cfg.GatewayModelPricingDatabaseURL) == "" {
		return nil, nil, errors.New("postgres_dev_test gateway model pricing database URL is missing")
	}
	pool, err := gatewaymodelpricingmigrations.OpenPool(ctx, cfg.GatewayModelPricingDatabaseURL)
	if err != nil {
		return nil, nil, err
	}
	closePool := func() { pool.Close() }
	state, err := gatewaymodelpricingmigrations.Inspect(ctx, pool)
	if err != nil {
		closePool()
		return nil, nil, err
	}
	if state.MigrationState != gatewaymodelpricingmigrations.MigrationStateApplied ||
		state.StoreSchemaVersion != gatewaymodelpricingmigrations.StoreSchemaVersion {
		closePool()
		return nil, nil, errors.New("gateway model pricing PostgreSQL migration is not applied")
	}
	return newPostgresGatewayModelPricingRepository(pool), closePool, nil
}
