package httpapi

import (
	"context"
	"errors"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	gatewayrequestquotamigrations "radishmind.local/services/platform/migrations/gateway_request_quotas"
	sqlitegatewayrequestquotamigrations "radishmind.local/services/platform/migrations/sqlite/gateway_request_quotas"
)

const (
	gatewayRequestQuotaStoreModeMemoryDev       = "memory_dev"
	gatewayRequestQuotaStoreModeSQLiteDev       = "sqlite_dev"
	gatewayRequestQuotaStoreModePostgresDevTest = "postgres_dev_test"
)

func newGatewayRequestQuotaRepositoryFromConfigWithSQLiteRuntime(
	cfg config.Config,
	sqliteRuntime *sqlitedev.Runtime,
) (GatewayRequestQuotaRepository, func(), error) {
	mode := strings.TrimSpace(cfg.GatewayRequestQuotaStoreMode)
	if mode == "" || mode == gatewayRequestQuotaStoreModeMemoryDev {
		return newMemoryGatewayRequestQuotaRepository(), func() {}, nil
	}
	timeout := cfg.GatewayRequestQuotaDatabaseTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if mode == gatewayRequestQuotaStoreModeSQLiteDev {
		if sqliteRuntime == nil || sqliteRuntime.DB() == nil {
			return nil, nil, errors.New("sqlite_dev gateway request quota requires the shared SQLite runtime")
		}
		if err := sqliteRuntime.VerifyMigrations(ctx, sqlitegatewayrequestquotamigrations.Migrations()); err != nil {
			return nil, nil, err
		}
		return newSQLiteGatewayRequestQuotaRepository(sqliteRuntime.DB()), func() {}, nil
	}
	if mode != gatewayRequestQuotaStoreModePostgresDevTest {
		return nil, nil, errors.New("unsupported gateway request quota store mode")
	}
	if strings.TrimSpace(cfg.GatewayRequestQuotaDatabaseURL) == "" {
		return nil, nil, errors.New("postgres_dev_test gateway request quota database URL is missing")
	}
	pool, err := gatewayrequestquotamigrations.OpenPool(ctx, cfg.GatewayRequestQuotaDatabaseURL)
	if err != nil {
		return nil, nil, err
	}
	closePool := func() { pool.Close() }
	state, err := gatewayrequestquotamigrations.Inspect(ctx, pool)
	if err != nil {
		closePool()
		return nil, nil, err
	}
	if state.MigrationState != gatewayrequestquotamigrations.MigrationStateApplied ||
		state.StoreSchemaVersion != gatewayrequestquotamigrations.StoreSchemaVersion {
		closePool()
		return nil, nil, errors.New("gateway request quota PostgreSQL migration is not applied")
	}
	return newPostgresGatewayRequestQuotaRepository(pool), closePool, nil
}
