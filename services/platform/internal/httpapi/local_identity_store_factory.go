package httpapi

import (
	"context"
	"errors"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/sqlitedev"
	localidentitymigrations "radishmind.local/services/platform/migrations/local_identity_records"
	sqlitelocalidentitymigrations "radishmind.local/services/platform/migrations/sqlite/local_identity_records"
)

const (
	localIdentityStoreModeMemoryDev       = "memory_dev"
	localIdentityStoreModeSQLiteDev       = "sqlite_dev"
	localIdentityStoreModePostgresDevTest = "postgres_dev_test"
)

type localIdentityStoreOptions struct {
	Mode                string
	SQLiteRuntime       *sqlitedev.Runtime
	PostgresDatabaseURL string
	DatabaseTimeout     time.Duration
}

func newLocalIdentityRepositoryFromOptions(options localIdentityStoreOptions) (localIdentityRepository, func(), error) {
	mode := strings.TrimSpace(options.Mode)
	if mode == "" || mode == localIdentityStoreModeMemoryDev {
		return newMemoryLocalIdentityRepository(), func() {}, nil
	}
	timeout := options.DatabaseTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if mode == localIdentityStoreModeSQLiteDev {
		if options.SQLiteRuntime == nil || options.SQLiteRuntime.DB() == nil {
			return nil, nil, errors.New("sqlite_dev local identity requires the shared SQLite runtime")
		}
		if err := options.SQLiteRuntime.VerifyMigrations(ctx, sqlitelocalidentitymigrations.Migrations()); err != nil {
			return nil, nil, err
		}
		return newSQLiteLocalIdentityRepository(options.SQLiteRuntime.DB()), func() {}, nil
	}
	if mode != localIdentityStoreModePostgresDevTest {
		return nil, nil, errors.New("unsupported local identity store mode")
	}
	if strings.TrimSpace(options.PostgresDatabaseURL) == "" {
		return nil, nil, errors.New("postgres_dev_test local identity database URL is missing")
	}
	pool, err := localidentitymigrations.OpenPool(ctx, options.PostgresDatabaseURL)
	if err != nil {
		return nil, nil, err
	}
	closePool := func() { pool.Close() }
	state, err := localidentitymigrations.Inspect(ctx, pool)
	if err != nil {
		closePool()
		return nil, nil, err
	}
	if state.MigrationState != localidentitymigrations.MigrationStateApplied ||
		state.StoreSchemaVersion != localidentitymigrations.StoreSchemaVersion {
		closePool()
		return nil, nil, errors.New("local identity PostgreSQL migration is not applied or incompatible")
	}
	return newPostgresLocalIdentityRepository(pool), closePool, nil
}
