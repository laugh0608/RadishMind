package httpapi

import (
	"context"
	"errors"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	adminproviderroutemigrations "radishmind.local/services/platform/migrations/admin_provider_routes"
	sqliteadminproviderroutemigrations "radishmind.local/services/platform/migrations/sqlite/admin_provider_routes"
)

const (
	adminProviderRouteStoreModeMemoryDev       = "memory_dev"
	adminProviderRouteStoreModeSQLiteDev       = "sqlite_dev"
	adminProviderRouteStoreModePostgresDevTest = "postgres_dev_test"
)

func newAdminProviderRouteRepositoryFromConfig(
	cfg config.Config,
) (adminProviderRouteRepository, func(), error) {
	return newAdminProviderRouteRepositoryFromConfigWithSQLiteRuntime(cfg, nil)
}

func newAdminProviderRouteRepositoryFromConfigWithSQLiteRuntime(
	cfg config.Config,
	sqliteRuntime *sqlitedev.Runtime,
) (adminProviderRouteRepository, func(), error) {
	mode := strings.TrimSpace(cfg.AdminProviderRouteStoreMode)
	if mode == "" || mode == adminProviderRouteStoreModeMemoryDev {
		return newMemoryAdminProviderRouteRepository(), func() {}, nil
	}
	timeout := cfg.AdminProviderRouteDatabaseTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if mode == adminProviderRouteStoreModeSQLiteDev {
		if sqliteRuntime == nil || sqliteRuntime.DB() == nil {
			return nil, nil, errors.New("sqlite_dev admin provider route requires the shared SQLite runtime")
		}
		if err := sqliteRuntime.VerifyMigrations(ctx, sqliteadminproviderroutemigrations.Migrations()); err != nil {
			return nil, nil, err
		}
		return newSQLiteAdminProviderRouteRepository(sqliteRuntime.DB()), func() {}, nil
	}
	if mode != adminProviderRouteStoreModePostgresDevTest {
		return nil, nil, errors.New("unsupported admin provider route store mode")
	}
	if strings.TrimSpace(cfg.AdminProviderRouteDatabaseURL) == "" {
		return nil, nil, errors.New("postgres_dev_test admin provider route database URL is missing")
	}
	pool, err := adminproviderroutemigrations.OpenPool(ctx, cfg.AdminProviderRouteDatabaseURL)
	if err != nil {
		return nil, nil, err
	}
	closePool := func() { pool.Close() }
	state, err := adminproviderroutemigrations.Inspect(ctx, pool)
	if err != nil {
		closePool()
		return nil, nil, err
	}
	if state.MigrationState != adminproviderroutemigrations.MigrationStateApplied {
		closePool()
		return nil, nil, errors.New("admin provider route PostgreSQL migration is not applied")
	}
	return newPostgresAdminProviderRouteRepository(pool), closePool, nil
}
