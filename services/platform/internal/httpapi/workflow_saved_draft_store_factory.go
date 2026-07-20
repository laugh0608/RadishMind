package httpapi

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	sqliteworkflowdraftmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_saved_drafts"
	workflowdraftmigrations "radishmind.local/services/platform/migrations/workflow_saved_drafts"
)

func newSavedWorkflowDraftStoreFromConfig(
	cfg config.Config,
) (savedWorkflowDraftStore, func(), error) {
	return newSavedWorkflowDraftStoreFromConfigWithSQLiteRuntime(cfg, nil)
}

func newSavedWorkflowDraftStoreFromConfigWithSQLiteRuntime(
	cfg config.Config,
	sqliteRuntime *sqlitedev.Runtime,
) (savedWorkflowDraftStore, func(), error) {
	mode := strings.TrimSpace(cfg.WorkflowSavedDraftStoreMode)
	if mode == "" {
		mode = string(WorkflowSavedDraftStoreModeMemoryDev)
	}
	if mode == string(WorkflowSavedDraftStoreModeSQLiteDev) {
		if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.WorkflowSavedDraftDevHTTPEnabled ||
			!cfg.WorkflowSavedDraftDevWriteEnabled {
			return nil, nil, errors.New("sqlite_dev saved workflow draft config is incomplete")
		}
		if sqliteRuntime == nil || sqliteRuntime.DB() == nil {
			return nil, nil, errors.New("sqlite_dev saved workflow draft store requires the shared SQLite runtime")
		}
		timeout := cfg.WorkflowSavedDraftDatabaseTimeout
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := sqliteRuntime.VerifyMigrations(ctx, sqliteworkflowdraftmigrations.Migrations()); err != nil {
			return nil, nil, err
		}
		if sqliteworkflowdraftmigrations.StoreSchemaVersion != savedWorkflowDraftRepositoryStoreSchemaVersion {
			return nil, nil, errors.New("saved workflow draft SQLite store schema version is incompatible")
		}
		selection := SelectWorkflowSavedDraftStore(mode, WorkflowSavedDraftStoreSelector{
			SQLiteDevStore: newSQLiteSavedWorkflowDraftStore(sqliteRuntime.DB()),
		})
		if selection.FailureCode != "" {
			return nil, nil, errors.New("sqlite_dev saved workflow draft store selection failed")
		}
		return selection.Store, func() {}, nil
	}
	if mode != string(WorkflowSavedDraftStoreModePostgresDevTest) {
		selection := SelectWorkflowSavedDraftStore(mode, WorkflowSavedDraftStoreSelector{})
		return selection.Store, func() {}, nil
	}

	missing := cfg.Check()
	if len(missing) > 0 {
		return nil, nil, fmt.Errorf(
			"postgres_dev_test saved workflow draft config is incomplete: %s",
			strings.Join(missing, ", "),
		)
	}
	timeout := cfg.WorkflowSavedDraftDatabaseTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	pool, err := workflowdraftmigrations.OpenPool(ctx, cfg.WorkflowSavedDraftDatabaseURL)
	if err != nil {
		return nil, nil, err
	}
	closePool := func() { pool.Close() }
	state, err := workflowdraftmigrations.Inspect(ctx, pool)
	if err != nil {
		closePool()
		return nil, nil, err
	}
	switch state.MigrationState {
	case workflowdraftmigrations.MigrationStateNotApplied:
		closePool()
		return nil, nil, errors.New("saved workflow draft PostgreSQL migration is not applied")
	case workflowdraftmigrations.MigrationStateMismatch:
		closePool()
		return nil, nil, errors.New("saved workflow draft PostgreSQL migration marker is incompatible")
	case workflowdraftmigrations.MigrationStateApplied:
	default:
		closePool()
		return nil, nil, errors.New("saved workflow draft PostgreSQL migration state is unavailable")
	}
	if savedWorkflowDraftRepositoryStoreSchemaVersion != workflowdraftmigrations.StoreSchemaVersion ||
		state.StoreSchemaVersion != savedWorkflowDraftRepositoryStoreSchemaVersion {
		closePool()
		return nil, nil, errors.New("saved workflow draft PostgreSQL store schema version is incompatible")
	}

	executor := newPostgresSavedWorkflowDraftQueryExecutor(pool)
	repository := NewSavedWorkflowDraftRepositoryAdapter(SavedWorkflowDraftRepositoryAdapterConfig{
		QueryExecutor: executor,
		SchemaPreflight: SavedWorkflowDraftRepositorySchemaPreflight{
			StoreSchemaVersion: state.StoreSchemaVersion,
			MigrationState:     savedWorkflowDraftRepositoryMigrationApplied,
		},
	})
	postgresStore := newRepositorySavedWorkflowDraftStore(repository)
	selection := SelectWorkflowSavedDraftStore(mode, WorkflowSavedDraftStoreSelector{
		PostgresDevTestStore: postgresStore,
	})
	if selection.FailureCode != "" {
		closePool()
		return nil, nil, errors.New("postgres_dev_test saved workflow draft store selection failed")
	}
	return selection.Store, closePool, nil
}
