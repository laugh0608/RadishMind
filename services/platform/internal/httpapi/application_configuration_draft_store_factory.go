package httpapi

import (
	"context"
	"errors"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
	applicationdraftmigrations "radishmind.local/services/platform/migrations/application_configuration_drafts"
)

func newApplicationConfigurationDraftRepositoryFromConfig(cfg config.Config) (applicationConfigurationDraftRepository, func(), error) {
	mode := strings.TrimSpace(cfg.ApplicationDraftStoreMode)
	if mode == "" || mode == "memory_dev" {
		return newMemoryApplicationConfigurationDraftRepository(), func() {}, nil
	}
	if mode != "postgres_dev_test" {
		return nil, nil, errors.New("unsupported application draft store mode")
	}
	if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.ApplicationDraftDevHTTPEnabled ||
		!cfg.ApplicationDraftDevWriteEnabled || strings.TrimSpace(cfg.ApplicationDraftDatabaseURL) == "" {
		return nil, nil, errors.New("postgres_dev_test application draft config is incomplete")
	}
	timeout := cfg.ApplicationDraftDatabaseTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	pool, err := applicationdraftmigrations.OpenPool(ctx, cfg.ApplicationDraftDatabaseURL)
	if err != nil {
		return nil, nil, err
	}
	closePool := func() { pool.Close() }
	state, err := applicationdraftmigrations.Inspect(ctx, pool)
	if err != nil {
		closePool()
		return nil, nil, err
	}
	if state.MigrationState != applicationdraftmigrations.MigrationStateApplied {
		closePool()
		return nil, nil, errors.New("application draft PostgreSQL migration is not applied")
	}
	return newPostgresApplicationConfigurationDraftRepository(pool), closePool, nil
}
