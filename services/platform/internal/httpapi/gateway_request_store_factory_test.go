package httpapi

import (
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
)

func TestGatewayRequestStoreFactoryModesFailClosed(t *testing.T) {
	for _, mode := range []string{"repository", "repository_disabled", "unknown"} {
		store, _, closeStore, err := newGatewayRequestStoreFromConfig(config.Config{GatewayRequestStoreMode: mode})
		if err == nil || store != nil || closeStore != nil {
			t.Fatalf("mode %s must fail closed: store=%T close=%v err=%v", mode, store, closeStore != nil, err)
		}
	}
	store, mode, closeStore, err := newGatewayRequestStoreFromConfig(config.Config{})
	if err != nil || mode != gatewayRequestStoreModeMemoryDev || closeStore == nil {
		t.Fatalf("default memory mode failed: store=%T mode=%s err=%v", store, mode, err)
	}
	closeStore()
}

func TestGatewayRequestStoreFactoryPostgresNoFallback(t *testing.T) {
	cfg := config.Config{
		ControlPlaneReadDevAuthEnabled:  true,
		GatewayRequestHistoryDevEnabled: true,
		GatewayRequestStoreMode:         gatewayRequestStoreModePostgresDevTest,
		GatewayRequestDatabaseURL:       "postgresql://invalid:invalid@127.0.0.1:1/invalid",
		GatewayRequestDatabaseTimeout:   50 * time.Millisecond,
	}
	store, mode, closeStore, err := newGatewayRequestStoreFromConfig(cfg)
	if err == nil || store != nil || closeStore != nil || mode != gatewayRequestStoreModePostgresDevTest {
		t.Fatalf("PostgreSQL connection failure must not fall back: store=%T mode=%s close=%v err=%v", store, mode, closeStore != nil, err)
	}
	if strings.Contains(err.Error(), "invalid:invalid") {
		t.Fatalf("database error leaked credentials: %v", err)
	}
}

func TestGatewayRequestStoreFactoryPostgresRequiresAllDevInputs(t *testing.T) {
	configurations := []config.Config{
		{GatewayRequestStoreMode: gatewayRequestStoreModePostgresDevTest},
		{GatewayRequestStoreMode: gatewayRequestStoreModePostgresDevTest, ControlPlaneReadDevAuthEnabled: true},
		{GatewayRequestStoreMode: gatewayRequestStoreModePostgresDevTest, ControlPlaneReadDevAuthEnabled: true, GatewayRequestHistoryDevEnabled: true},
	}
	for _, cfg := range configurations {
		store, _, closeStore, err := newGatewayRequestStoreFromConfig(cfg)
		if err == nil || store != nil || closeStore != nil {
			t.Fatalf("incomplete PostgreSQL config was accepted: %#v", cfg)
		}
	}
}
