package httpapi

import (
	"testing"

	"radishmind.local/services/platform/internal/config"
)

func TestAPIKeyStoreFactoryModes(t *testing.T) {
	repository, closeRepository, err := newAPIKeyRepositoryFromConfig(config.Config{APIKeyStoreMode: "memory_dev"})
	if err != nil || repository == nil || closeRepository == nil {
		t.Fatalf("memory_dev API key repository selection failed: repository=%T err=%v", repository, err)
	}
	closeRepository()
	if _, _, err := newAPIKeyRepositoryFromConfig(config.Config{APIKeyStoreMode: "unknown"}); err == nil {
		t.Fatal("unknown API key store mode must fail")
	}
	if _, _, err := newAPIKeyRepositoryFromConfig(config.Config{APIKeyStoreMode: "postgres_dev_test"}); err == nil {
		t.Fatal("incomplete postgres_dev_test API key configuration must fail before connecting")
	}
}
