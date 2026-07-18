package httpapi

import (
	"context"
	"path/filepath"
	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	sqliteworkflowrunmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_runs"
	"testing"
	"time"
)

func TestWorkflowRunStoreFactoryNoFallback(t *testing.T) {
	cfg := config.Config{ControlPlaneReadDevAuthEnabled: true, WorkflowSavedDraftDevHTTPEnabled: true, WorkflowExecutorDevEnabled: true, WorkflowRunStoreMode: "postgres_dev_test", WorkflowRunDatabaseURL: "postgresql://invalid:invalid@127.0.0.1:1/invalid", WorkflowRunDatabaseTimeout: 50 * time.Millisecond}
	store, closeStore, err := newWorkflowRunStoreFromConfig(cfg)
	if err == nil || store != nil || closeStore != nil {
		t.Fatalf("database failure must not select memory: store=%T err=%v", store, err)
	}
}

func TestWorkflowRunStoreFactoryModesFailClosed(t *testing.T) {
	for _, mode := range []string{"repository", "repository_disabled", "unknown"} {
		store, closeStore, err := newWorkflowRunStoreFromConfig(config.Config{WorkflowRunStoreMode: mode})
		if err == nil || store != nil || closeStore != nil {
			t.Fatalf("mode %s must fail closed", mode)
		}
	}
}

func TestWorkflowRunSQLiteStoreFactory(t *testing.T) {
	cfg := sqliteWorkflowRunConfig()
	if store, closeStore, err := newWorkflowRunStoreFromConfigWithSQLiteRuntime(cfg, nil); err == nil ||
		store != nil || closeStore != nil || err.Error() != "sqlite_dev workflow run store requires the shared SQLite runtime" {
		t.Fatalf("missing shared SQLite runtime did not fail closed: store=%#v close_set=%v err=%v", store, closeStore != nil, err)
	}

	withoutMigration, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: filepath.Join(t.TempDir(), "without-workflow-run-migration.db"),
	})
	if err != nil {
		t.Fatalf("open SQLite runtime without workflow run migration: %v", err)
	}
	t.Cleanup(func() { _ = withoutMigration.Close() })
	if store, closeStore, factoryErr := newWorkflowRunStoreFromConfigWithSQLiteRuntime(cfg, withoutMigration); factoryErr == nil ||
		store != nil || closeStore != nil || factoryErr.Error() != "SQLite development component migration is not applied" {
		t.Fatalf("missing workflow run marker did not fail closed: store=%#v close_set=%v err=%v", store, closeStore != nil, factoryErr)
	}

	runtime := openWorkflowRunSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
	store, closeStore, err := newWorkflowRunStoreFromConfigWithSQLiteRuntime(cfg, runtime)
	if err != nil || store == nil || closeStore == nil {
		t.Fatalf("construct SQLite workflow run store: store=%#v close_set=%v err=%v", store, closeStore != nil, err)
	}
	sqliteStore, ok := store.(*sqliteWorkflowRunStore)
	if !ok || sqliteStore.database != runtime.DB() {
		t.Fatalf("SQLite workflow run store did not reuse the shared database: %T", store)
	}
	if evaluationStore, ok := newWorkflowEvaluationStoreForRunStore(store).(*sqliteWorkflowEvaluationStore); !ok || evaluationStore.database != runtime.DB() {
		t.Fatalf("SQLite evaluation case store did not share the workflow run database: %T", newWorkflowEvaluationStoreForRunStore(store))
	}
	if suiteStore, ok := newWorkflowEvaluationSuiteStoreForRunStore(store).(*sqliteWorkflowEvaluationSuiteStore); !ok || suiteStore.database != runtime.DB() {
		t.Fatalf("SQLite evaluation suite store did not share the workflow run database: %T", newWorkflowEvaluationSuiteStoreForRunStore(store))
	}
	if actionStore, ok := newWorkflowHTTPToolActionStoreForRunStore(store).(*sqliteWorkflowHTTPToolActionStore); !ok || actionStore.database != runtime.DB() {
		t.Fatalf("SQLite tool action store did not share the workflow run database: %T", actionStore)
	}
	closeStore()
	if err := runtime.DB().PingContext(context.Background()); err != nil {
		t.Fatalf("component close must not own the shared SQLite runtime: %v", err)
	}
	if _, err := runtime.DB().ExecContext(context.Background(), `UPDATE radishmind_schema_migrations
        SET migration_checksum=? WHERE component=? AND migration_id=?`,
		"sha256:changed",
		sqliteworkflowrunmigrations.Component,
		sqliteworkflowrunmigrations.MigrationID,
	); err != nil {
		t.Fatalf("corrupt workflow run SQLite migration marker: %v", err)
	}
	if store, closeStore, err := newWorkflowRunStoreFromConfigWithSQLiteRuntime(cfg, runtime); err == nil ||
		store != nil || closeStore != nil || err.Error() != "SQLite development component migration marker mismatch" {
		t.Fatalf("incompatible workflow run marker did not fail closed: store=%#v close_set=%v err=%v", store, closeStore != nil, err)
	}
}

func TestWorkflowRunSQLiteStoreFactoryRequiresDevelopmentGates(t *testing.T) {
	runtime := openWorkflowRunSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
	cfg := sqliteWorkflowRunConfig()
	for _, mutate := range []func(*config.Config){
		func(cfg *config.Config) { cfg.ControlPlaneReadDevAuthEnabled = false },
		func(cfg *config.Config) { cfg.WorkflowSavedDraftDevHTTPEnabled = false },
		func(cfg *config.Config) { cfg.WorkflowExecutorDevEnabled = false },
	} {
		candidate := cfg
		mutate(&candidate)
		if store, closeStore, err := newWorkflowRunStoreFromConfigWithSQLiteRuntime(candidate, runtime); err == nil ||
			store != nil || closeStore != nil || err.Error() != "sqlite_dev workflow run store config is incomplete" {
			t.Fatalf("incomplete SQLite workflow run gates were accepted: store=%#v close_set=%v err=%v", store, closeStore != nil, err)
		}
	}
}

func TestWorkflowRunSQLiteStoreFactoryAcceptsIndependentRAGExecutionGate(t *testing.T) {
	runtime := openWorkflowRunSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
	cfg := sqliteWorkflowRunConfig()
	cfg.WorkflowExecutorDevEnabled = false
	cfg.WorkflowRAGExecutionDevEnabled = true

	store, closeStore, err := newWorkflowRunStoreFromConfigWithSQLiteRuntime(cfg, runtime)
	if err != nil || store == nil || closeStore == nil {
		t.Fatalf("independent RAG execution gate did not select SQLite run store: store=%#v close_set=%v err=%v", store, closeStore != nil, err)
	}
	closeStore()
}

func sqliteWorkflowRunConfig() config.Config {
	return config.Config{
		ControlPlaneReadDevAuthEnabled:   true,
		WorkflowSavedDraftDevHTTPEnabled: true,
		WorkflowExecutorDevEnabled:       true,
		WorkflowRunStoreMode:             workflowRunStoreModeSQLiteDev,
		WorkflowRunDatabaseTimeout:       time.Second,
	}
}

func TestWorkflowEvaluationStoresFollowExplicitRunStoreSelection(t *testing.T) {
	memoryRunStore := newMemoryWorkflowRunStore(10)
	if _, ok := newWorkflowEvaluationStoreForRunStore(memoryRunStore).(*memoryWorkflowEvaluationStore); !ok {
		t.Fatal("memory_dev case store selection drifted")
	}
	if _, ok := newWorkflowEvaluationSuiteStoreForRunStore(memoryRunStore).(*memoryWorkflowEvaluationSuiteStore); !ok {
		t.Fatal("memory_dev suite store selection drifted")
	}
	if actionStore, ok := newWorkflowHTTPToolActionStoreForRunStore(memoryRunStore).(*memoryWorkflowHTTPToolActionStore); !ok || actionStore.ownerLock != &memoryRunStore.mu {
		t.Fatalf("memory tool action store did not share the run owner lock: %T", actionStore)
	}
	postgresRunStore := newPostgresWorkflowRunStore(nil)
	if _, ok := newWorkflowEvaluationStoreForRunStore(postgresRunStore).(*postgresWorkflowEvaluationStore); !ok {
		t.Fatal("postgres_dev_test case store selection drifted")
	}
	if _, ok := newWorkflowEvaluationSuiteStoreForRunStore(postgresRunStore).(*postgresWorkflowEvaluationSuiteStore); !ok {
		t.Fatal("postgres_dev_test suite store selection drifted")
	}
	if actionStore, ok := newWorkflowHTTPToolActionStoreForRunStore(postgresRunStore).(*postgresWorkflowHTTPToolActionStore); !ok || actionStore.pool != postgresRunStore.pool {
		t.Fatalf("PostgreSQL tool action store did not share the run pool: %T", actionStore)
	}
}
