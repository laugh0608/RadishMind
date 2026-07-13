package httpapi

import (
	"testing"

	"radishmind.local/services/platform/internal/config"
)

func TestApplicationCatalogStoreFactoryModes(t *testing.T) {
	repository, closeRepository, err := newApplicationCatalogRepositoryFromConfig(config.Config{ApplicationCatalogStoreMode: "memory_dev"})
	if err != nil || repository == nil || closeRepository == nil {
		t.Fatalf("memory_dev store must be available: repository=%T err=%v", repository, err)
	}
	closeRepository()
	if _, ok := repository.(*memoryApplicationCatalogRepository); !ok {
		t.Fatalf("unexpected memory repository type: %T", repository)
	}

	if _, _, err := newApplicationCatalogRepositoryFromConfig(config.Config{ApplicationCatalogStoreMode: "unknown"}); err == nil {
		t.Fatal("unknown application catalog store mode must fail")
	}
	if _, _, err := newApplicationCatalogRepositoryFromConfig(config.Config{ApplicationCatalogStoreMode: "postgres_dev_test"}); err == nil {
		t.Fatal("incomplete postgres_dev_test configuration must fail before connecting")
	}
}
