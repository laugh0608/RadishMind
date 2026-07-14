package httpapi

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	sqliteworkflowdraftmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_saved_drafts"
)

func TestSavedWorkflowDraftSQLiteStoreFactory(t *testing.T) {
	cfg := sqliteSavedWorkflowDraftConfig()

	if store, closeStore, err := newSavedWorkflowDraftStoreFromConfigWithSQLiteRuntime(cfg, nil); err == nil ||
		store != nil || closeStore != nil || err.Error() != "sqlite_dev saved workflow draft store requires the shared SQLite runtime" {
		t.Fatalf("missing shared SQLite runtime did not fail closed: store=%#v close_set=%v err=%v", store, closeStore != nil, err)
	}

	withoutMigration, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: filepath.Join(t.TempDir(), "without-migration.db"),
	})
	if err != nil {
		t.Fatalf("open SQLite runtime without saved draft migration: %v", err)
	}
	t.Cleanup(func() { _ = withoutMigration.Close() })
	if store, closeStore, factoryErr := newSavedWorkflowDraftStoreFromConfigWithSQLiteRuntime(cfg, withoutMigration); factoryErr == nil ||
		store != nil || closeStore != nil || factoryErr.Error() != "SQLite development component migration is not applied" {
		t.Fatalf("missing saved draft marker did not fail closed: store=%#v close_set=%v err=%v", store, closeStore != nil, factoryErr)
	}

	runtime := openSavedWorkflowDraftSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
	store, closeStore, err := newSavedWorkflowDraftStoreFromConfigWithSQLiteRuntime(cfg, runtime)
	if err != nil || store == nil || closeStore == nil {
		t.Fatalf("construct SQLite saved draft store: store=%#v close_set=%v err=%v", store, closeStore != nil, err)
	}
	repositoryStore, ok := store.(*repositorySavedWorkflowDraftStore)
	if !ok {
		t.Fatalf("SQLite selector returned unexpected store type: %T", store)
	}
	adapter, ok := repositoryStore.repository.(SavedWorkflowDraftRepositoryAdapter)
	if !ok {
		t.Fatalf("SQLite store did not use the repository adapter: %T", repositoryStore.repository)
	}
	executor, ok := adapter.queryExecutor.(*sqliteSavedWorkflowDraftQueryExecutor)
	if !ok || executor.database != runtime.DB() {
		t.Fatalf("SQLite saved draft store did not reuse the shared database: %#v", adapter.queryExecutor)
	}
	closeStore()
	if err := runtime.DB().PingContext(context.Background()); err != nil {
		t.Fatalf("component close must not own the shared SQLite runtime: %v", err)
	}
	if _, err := runtime.DB().ExecContext(context.Background(), `UPDATE radishmind_schema_migrations
        SET migration_checksum=? WHERE component=? AND migration_id=?`,
		"sha256:changed",
		sqliteworkflowdraftmigrations.Component,
		sqliteworkflowdraftmigrations.MigrationID,
	); err != nil {
		t.Fatalf("corrupt saved draft SQLite migration marker: %v", err)
	}
	if store, closeStore, err := newSavedWorkflowDraftStoreFromConfigWithSQLiteRuntime(cfg, runtime); err == nil ||
		store != nil || closeStore != nil || err.Error() != "SQLite development component migration marker mismatch" {
		t.Fatalf("incompatible saved draft marker did not fail closed: store=%#v close_set=%v err=%v", store, closeStore != nil, err)
	}
}

func TestSavedWorkflowDraftSQLiteStoreFactoryRequiresDevelopmentGates(t *testing.T) {
	runtime := openSavedWorkflowDraftSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
	cfg := sqliteSavedWorkflowDraftConfig()
	for _, mutate := range []func(*config.Config){
		func(cfg *config.Config) { cfg.ControlPlaneReadDevAuthEnabled = false },
		func(cfg *config.Config) { cfg.WorkflowSavedDraftDevHTTPEnabled = false },
		func(cfg *config.Config) { cfg.WorkflowSavedDraftDevWriteEnabled = false },
	} {
		candidate := cfg
		mutate(&candidate)
		if store, closeStore, err := newSavedWorkflowDraftStoreFromConfigWithSQLiteRuntime(candidate, runtime); err == nil ||
			store != nil || closeStore != nil || err.Error() != "sqlite_dev saved workflow draft config is incomplete" {
			t.Fatalf("incomplete SQLite saved draft gates were accepted: store=%#v close_set=%v err=%v", store, closeStore != nil, err)
		}
	}
}

func sqliteSavedWorkflowDraftConfig() config.Config {
	return config.Config{
		ControlPlaneReadDevAuthEnabled:    true,
		WorkflowSavedDraftDevHTTPEnabled:  true,
		WorkflowSavedDraftDevWriteEnabled: true,
		WorkflowSavedDraftStoreMode:       string(WorkflowSavedDraftStoreModeSQLiteDev),
		WorkflowSavedDraftDatabaseTimeout: time.Second,
	}
}

func savedWorkflowDraftSQLiteMigrations() []sqlitedev.Migration {
	return sqliteworkflowdraftmigrations.Migrations()
}
