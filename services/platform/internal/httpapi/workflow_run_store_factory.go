package httpapi

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	sqliteworkflowrunmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_runs"
	workflowrunmigrations "radishmind.local/services/platform/migrations/workflow_runs"
)

const (
	workflowRunStoreModeMemoryDev       = "memory_dev"
	workflowRunStoreModeSQLiteDev       = "sqlite_dev"
	workflowRunStoreModePostgresDevTest = "postgres_dev_test"
)

func newWorkflowRunStoreFromConfig(cfg config.Config) (workflowRunStore, func(), error) {
	return newWorkflowRunStoreFromConfigWithSQLiteRuntime(cfg, nil)
}

func newWorkflowRunStoreFromConfigWithSQLiteRuntime(
	cfg config.Config,
	sqliteRuntime *sqlitedev.Runtime,
) (workflowRunStore, func(), error) {
	mode := strings.TrimSpace(cfg.WorkflowRunStoreMode)
	if mode == "" || mode == workflowRunStoreModeMemoryDev {
		return newMemoryWorkflowRunStore(defaultWorkflowRunStoreCapacity), func() {}, nil
	}
	if mode == workflowRunStoreModeSQLiteDev {
		if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.WorkflowSavedDraftDevHTTPEnabled ||
			(!cfg.WorkflowExecutorDevEnabled && !cfg.WorkflowRAGExecutionDevEnabled) {
			return nil, nil, errors.New("sqlite_dev workflow run store config is incomplete")
		}
		if sqliteRuntime == nil || sqliteRuntime.DB() == nil {
			return nil, nil, errors.New("sqlite_dev workflow run store requires the shared SQLite runtime")
		}
		timeout := cfg.WorkflowRunDatabaseTimeout
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := sqliteRuntime.VerifyMigrations(ctx, sqliteworkflowrunmigrations.Migrations()); err != nil {
			return nil, nil, err
		}
		return newSQLiteWorkflowRunStore(sqliteRuntime.DB()), func() {}, nil
	}
	if mode != workflowRunStoreModePostgresDevTest {
		if mode == "repository" || mode == "repository_disabled" {
			return nil, nil, fmt.Errorf("%s", WorkflowRunFailureStoreModeDisabled)
		}
		return nil, nil, fmt.Errorf("%s", WorkflowRunFailureStoreModeInvalid)
	}
	if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.WorkflowSavedDraftDevHTTPEnabled ||
		(!cfg.WorkflowExecutorDevEnabled && !cfg.WorkflowRAGExecutionDevEnabled) || strings.TrimSpace(cfg.WorkflowRunDatabaseURL) == "" {
		return nil, nil, errors.New("postgres_dev_test workflow run store config is incomplete")
	}
	timeout := cfg.WorkflowRunDatabaseTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	pool, err := workflowrunmigrations.OpenPool(ctx, cfg.WorkflowRunDatabaseURL)
	if err != nil {
		return nil, nil, err
	}
	closePool := func() { pool.Close() }
	state, err := workflowrunmigrations.Inspect(ctx, pool)
	if err != nil {
		closePool()
		return nil, nil, err
	}
	if state.MigrationState != workflowrunmigrations.MigrationStateApplied || state.StoreSchemaVersion != workflowrunmigrations.StoreSchemaVersion {
		closePool()
		return nil, nil, errors.New("workflow run PostgreSQL migration is not applied or incompatible")
	}
	return newPostgresWorkflowRunStore(pool), closePool, nil
}

func newWorkflowEvaluationStoreForRunStore(store workflowRunStore) workflowEvaluationStore {
	if postgres, ok := store.(*postgresWorkflowRunStore); ok {
		return newPostgresWorkflowEvaluationStore(postgres.pool)
	}
	return newMemoryWorkflowEvaluationStore(defaultWorkflowEvaluationCapacity)
}

func newWorkflowHTTPToolActionStoreForRunStore(store workflowRunStore) workflowHTTPToolActionStore {
	switch typed := store.(type) {
	case *memoryWorkflowRunStore:
		return newMemoryWorkflowHTTPToolActionStore(&typed.mu)
	case *sqliteWorkflowRunStore:
		return newSQLiteWorkflowHTTPToolActionStore(typed.database)
	case *postgresWorkflowRunStore:
		return newPostgresWorkflowHTTPToolActionStore(typed.pool)
	default:
		return nil
	}
}

func newWorkflowHTTPToolExecutionStoreForRunStore(
	store workflowRunStore,
	actionStore workflowHTTPToolActionStore,
) workflowHTTPToolExecutionStore {
	switch typed := store.(type) {
	case *memoryWorkflowRunStore:
		actions, ok := actionStore.(*memoryWorkflowHTTPToolActionStore)
		if !ok {
			return nil
		}
		return newMemoryWorkflowHTTPToolExecutionStore(actions, typed)
	case *sqliteWorkflowRunStore:
		return newSQLiteWorkflowHTTPToolExecutionStore(typed.database)
	case *postgresWorkflowRunStore:
		return newPostgresWorkflowHTTPToolExecutionStore(typed.pool)
	default:
		return nil
	}
}

func newWorkflowEvaluationSuiteStoreForRunStore(store workflowRunStore) workflowEvaluationSuiteStore {
	if postgres, ok := store.(*postgresWorkflowRunStore); ok {
		return newPostgresWorkflowEvaluationSuiteStore(postgres.pool)
	}
	return newMemoryWorkflowEvaluationSuiteStore(defaultWorkflowEvaluationCapacity)
}

func newWorkflowRAGSnapshotRepositoryForRunStore(store workflowRunStore) (workflowRAGSnapshotRepository, error) {
	switch typed := store.(type) {
	case *memoryWorkflowRunStore:
		return newMemoryWorkflowRAGSnapshotRepository(&typed.mu), nil
	case *sqliteWorkflowRunStore:
		if typed.database == nil {
			return nil, errors.New("workflow RAG snapshot store requires the shared SQLite database")
		}
		return newSQLiteWorkflowRAGSnapshotRepository(typed.database), nil
	case *postgresWorkflowRunStore:
		if typed.pool == nil {
			return nil, errors.New("workflow RAG snapshot store requires the workflow PostgreSQL pool")
		}
		return newPostgresWorkflowRAGSnapshotRepository(typed.pool), nil
	default:
		return nil, errors.New("workflow RAG snapshot store requires a supported workflow runtime backend")
	}
}
