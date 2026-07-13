package httpapi

import (
	"context"
	"errors"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
	applicationcatalogmigrations "radishmind.local/services/platform/migrations/application_catalog_records"
)

func newApplicationCatalogRepositoryFromConfig(cfg config.Config) (applicationCatalogRepository, func(), error) {
	mode := strings.TrimSpace(cfg.ApplicationCatalogStoreMode)
	if mode == "" || mode == "memory_dev" {
		return newMemoryApplicationCatalogRepository(), func() {}, nil
	}
	if mode != "postgres_dev_test" {
		return nil, nil, errors.New("unsupported application catalog store mode")
	}
	if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.ApplicationCatalogDevHTTPEnabled || !cfg.ApplicationCatalogDevWriteEnabled ||
		strings.TrimSpace(cfg.ApplicationCatalogDatabaseURL) == "" {
		return nil, nil, errors.New("postgres_dev_test application catalog config is incomplete")
	}
	timeout := cfg.ApplicationCatalogDatabaseTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	pool, err := applicationcatalogmigrations.OpenPool(ctx, cfg.ApplicationCatalogDatabaseURL)
	if err != nil {
		return nil, nil, err
	}
	closePool := func() { pool.Close() }
	state, err := applicationcatalogmigrations.Inspect(ctx, pool)
	if err != nil {
		closePool()
		return nil, nil, err
	}
	if state.MigrationState != applicationcatalogmigrations.MigrationStateApplied {
		closePool()
		return nil, nil, errors.New("application catalog PostgreSQL migration is not applied")
	}
	return newPostgresApplicationCatalogRepository(pool), closePool, nil
}
