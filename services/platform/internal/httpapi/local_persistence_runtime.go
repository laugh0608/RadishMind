package httpapi

import (
	"context"
	"time"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	localmigrations "radishmind.local/services/platform/migrations/sqlite"
)

const localPersistenceStartupTimeout = 30 * time.Second

func openLocalPersistenceRuntime(cfg config.Config) (*sqlitedev.Runtime, error) {
	if config.EffectiveLocalPersistenceMode(cfg) != "sqlite_dev" {
		return nil, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), localPersistenceStartupTimeout)
	defer cancel()
	return sqlitedev.Open(ctx, sqlitedev.Options{
		DatabasePath: cfg.SQLiteDevDatabasePath,
		Migrations:   localmigrations.Migrations(),
	})
}

func closeServerStartupResources(closers ...func()) {
	for index := len(closers) - 1; index >= 0; index-- {
		if closers[index] != nil {
			closers[index]()
		}
	}
}

func sqliteRuntimeCloser(runtime *sqlitedev.Runtime) func() {
	if runtime == nil {
		return nil
	}
	return func() {
		_ = runtime.Close()
	}
}
