package httpapi

import (
	"context"
	"time"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	sqliteapikeymigrations "radishmind.local/services/platform/migrations/sqlite/api_key_records"
	sqliteapplicationcatalogmigrations "radishmind.local/services/platform/migrations/sqlite/application_catalog_records"
	sqliteapplicationdraftmigrations "radishmind.local/services/platform/migrations/sqlite/application_configuration_drafts"
	sqliteapplicationpublishmigrations "radishmind.local/services/platform/migrations/sqlite/application_publish_candidates"
	sqlitegatewayrequestmigrations "radishmind.local/services/platform/migrations/sqlite/gateway_requests"
	sqliteworkflowrunmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_runs"
	sqliteworkflowdraftmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_saved_drafts"
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
		Migrations:   localPersistenceSQLiteMigrations(),
	})
}

func localPersistenceSQLiteMigrations() []sqlitedev.Migration {
	migrations := make([]sqlitedev.Migration, 0, 7)
	migrations = append(migrations, sqliteapplicationcatalogmigrations.Migrations()...)
	migrations = append(migrations, sqliteapplicationdraftmigrations.Migrations()...)
	migrations = append(migrations, sqliteapplicationpublishmigrations.Migrations()...)
	migrations = append(migrations, sqliteapikeymigrations.Migrations()...)
	migrations = append(migrations, sqlitegatewayrequestmigrations.Migrations()...)
	migrations = append(migrations, sqliteworkflowdraftmigrations.Migrations()...)
	migrations = append(migrations, sqliteworkflowrunmigrations.Migrations()...)
	return migrations
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
