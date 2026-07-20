package httpapi

import (
	"context"
	"errors"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
	controlplanereadmigrations "radishmind.local/services/platform/migrations/control_plane_admin_read"
)

func newControlPlaneReadRepositoryFromConfig(cfg config.Config) (ControlPlaneReadRepository, func(), error) {
	fakeRepository := newControlPlaneReadRepository(newControlPlaneReadFakeStore())
	switch config.EffectiveControlPlaneReadStoreMode(cfg) {
	case "fake_store_dev":
		return fakeRepository, func() {}, nil
	case "postgres_dev_test":
		authMode := config.EffectiveControlPlaneReadAuthMode(cfg)
		if (authMode != "signed_test_token" && authMode != "radish_oidc_integration_test") || strings.TrimSpace(cfg.ControlPlaneReadDatabaseURL) == "" {
			return nil, func() {}, errors.New("control plane read postgres_dev_test requires verified test auth and database configuration")
		}
		timeout := cfg.ControlPlaneReadDatabaseTimeout
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		pool, err := controlplanereadmigrations.OpenPool(ctx, cfg.ControlPlaneReadDatabaseURL)
		if err != nil {
			return nil, func() {}, err
		}
		if _, err := controlplanereadmigrations.PreflightRuntime(ctx, pool); err != nil {
			pool.Close()
			return nil, func() {}, err
		}
		return newRoutedControlPlaneReadRepository(pool, timeout, fakeRepository), pool.Close, nil
	default:
		return nil, func() {}, errors.New("unsupported control plane read store mode")
	}
}
