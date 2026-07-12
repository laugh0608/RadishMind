package httpapi

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
	gatewayrequestmigrations "radishmind.local/services/platform/migrations/gateway_requests"
)

func newGatewayRequestStoreFromConfig(cfg config.Config) (gatewayRequestStore, string, func(), error) {
	mode := strings.TrimSpace(cfg.GatewayRequestStoreMode)
	if mode == "" || mode == gatewayRequestStoreModeMemoryDev {
		return newMemoryGatewayRequestStore(defaultGatewayRequestStoreCapacity), gatewayRequestStoreModeMemoryDev, func() {}, nil
	}
	if mode != gatewayRequestStoreModePostgresDevTest {
		if mode == "repository" || mode == "repository_disabled" {
			return nil, mode, nil, fmt.Errorf("%s", GatewayRequestHistoryFailureStoreModeDisabled)
		}
		return nil, mode, nil, fmt.Errorf("%s", GatewayRequestHistoryFailureStoreModeInvalid)
	}
	if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.GatewayRequestHistoryDevEnabled || strings.TrimSpace(cfg.GatewayRequestDatabaseURL) == "" {
		return nil, mode, nil, errors.New("postgres_dev_test Gateway request store config is incomplete")
	}
	timeout := cfg.GatewayRequestDatabaseTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	pool, err := gatewayrequestmigrations.OpenPool(ctx, cfg.GatewayRequestDatabaseURL)
	if err != nil {
		return nil, mode, nil, err
	}
	closePool := func() { pool.Close() }
	state, err := gatewayrequestmigrations.Inspect(ctx, pool)
	if err != nil {
		closePool()
		return nil, mode, nil, err
	}
	if state.MigrationState != gatewayrequestmigrations.MigrationStateApplied || state.StoreSchemaVersion != gatewayrequestmigrations.StoreSchemaVersion {
		closePool()
		return nil, mode, nil, errors.New("Gateway request PostgreSQL migration is not applied or incompatible")
	}
	return newPostgresGatewayRequestStore(pool), mode, closePool, nil
}
