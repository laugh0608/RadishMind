package httpapi

import (
	"errors"
	"strings"

	"radishmind.local/services/platform/internal/config"
)

func newApplicationCatalogRepositoryFromConfig(cfg config.Config) (applicationCatalogRepository, func(), error) {
	mode := strings.TrimSpace(cfg.ApplicationCatalogStoreMode)
	if mode == "" || mode == "memory_dev" {
		return newMemoryApplicationCatalogRepository(), func() {}, nil
	}
	return nil, nil, errors.New("unsupported application catalog store mode")
}
