package httpapi

import (
	"context"
	"errors"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
	applicationpublishmigrations "radishmind.local/services/platform/migrations/application_publish_candidates"
)

func newApplicationPublishCandidateRepositoryFromConfig(cfg config.Config) (applicationPublishCandidateRepository, func(), error) {
	mode := strings.TrimSpace(cfg.ApplicationPublishStoreMode)
	if mode == "" || mode == "memory_dev" {
		return newMemoryApplicationPublishCandidateRepository(), func() {}, nil
	}
	if mode != "postgres_dev_test" {
		return nil, nil, errors.New("unsupported application publish candidate store mode")
	}
	if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.ApplicationDraftDevHTTPEnabled || !cfg.ApplicationDraftDevWriteEnabled ||
		strings.TrimSpace(cfg.ApplicationDraftStoreMode) != "postgres_dev_test" || strings.TrimSpace(cfg.ApplicationDraftDatabaseURL) == "" ||
		!cfg.ApplicationPublishDevHTTPEnabled || !cfg.ApplicationPublishDevWriteEnabled || strings.TrimSpace(cfg.ApplicationPublishDatabaseURL) == "" {
		return nil, nil, errors.New("postgres_dev_test application publish candidate config is incomplete")
	}
	timeout := cfg.ApplicationPublishDatabaseTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	pool, err := applicationpublishmigrations.OpenPool(ctx, cfg.ApplicationPublishDatabaseURL)
	if err != nil {
		return nil, nil, err
	}
	closePool := func() { pool.Close() }
	state, err := applicationpublishmigrations.Inspect(ctx, pool)
	if err != nil {
		closePool()
		return nil, nil, err
	}
	if state.MigrationState != applicationpublishmigrations.MigrationStateApplied {
		closePool()
		return nil, nil, errors.New("application publish candidate PostgreSQL migration is not applied")
	}
	return newPostgresApplicationPublishCandidateRepository(pool), closePool, nil
}
