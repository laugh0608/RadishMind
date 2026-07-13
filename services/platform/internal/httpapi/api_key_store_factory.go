package httpapi

import (
	"context"
	"errors"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
	apikeymigrations "radishmind.local/services/platform/migrations/api_key_records"
)

func newAPIKeyRepositoryFromConfig(cfg config.Config) (apiKeyRepository, func(), error) {
	mode := strings.TrimSpace(cfg.APIKeyStoreMode)
	if mode == "" || mode == "memory_dev" {
		return newMemoryAPIKeyRepository(), func() {}, nil
	}
	if mode != "postgres_dev_test" {
		return nil, nil, errors.New("unsupported API key store mode")
	}
	if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.APIKeyLifecycleDevHTTPEnabled || !cfg.APIKeyLifecycleDevWriteEnabled ||
		strings.TrimSpace(cfg.APIKeyDatabaseURL) == "" {
		return nil, nil, errors.New("postgres_dev_test API key config is incomplete")
	}
	timeout := cfg.APIKeyDatabaseTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	pool, err := apikeymigrations.OpenPool(ctx, cfg.APIKeyDatabaseURL)
	if err != nil {
		return nil, nil, err
	}
	closePool := func() { pool.Close() }
	state, err := apikeymigrations.Inspect(ctx, pool)
	if err != nil {
		closePool()
		return nil, nil, err
	}
	if state.MigrationState != apikeymigrations.MigrationStateApplied || state.StoreSchemaVersion != apikeymigrations.StoreSchemaVersion {
		closePool()
		return nil, nil, errors.New("API key PostgreSQL migration is not applied or incompatible")
	}
	return newPostgresAPIKeyRepository(pool), closePool, nil
}
