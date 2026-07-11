package httpapi

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
	workflowrunmigrations "radishmind.local/services/platform/migrations/workflow_runs"
)

const (
	workflowRunStoreModeMemoryDev       = "memory_dev"
	workflowRunStoreModePostgresDevTest = "postgres_dev_test"
)

func newWorkflowRunStoreFromConfig(cfg config.Config) (workflowRunStore, func(), error) {
	mode := strings.TrimSpace(cfg.WorkflowRunStoreMode)
	if mode == "" || mode == workflowRunStoreModeMemoryDev {
		return newMemoryWorkflowRunStore(defaultWorkflowRunStoreCapacity), func() {}, nil
	}
	if mode != workflowRunStoreModePostgresDevTest {
		if mode == "repository" || mode == "repository_disabled" {
			return nil, nil, fmt.Errorf("%s", WorkflowRunFailureStoreModeDisabled)
		}
		return nil, nil, fmt.Errorf("%s", WorkflowRunFailureStoreModeInvalid)
	}
	if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.WorkflowExecutorDevEnabled || strings.TrimSpace(cfg.WorkflowRunDatabaseURL) == "" {
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
