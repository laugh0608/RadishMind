package config

import (
	"errors"
	"strings"
)

func ValidateServerStart(cfg Config) error {
	if EffectiveLocalPersistenceMode(cfg) != "sqlite_dev" {
		return nil
	}
	return validateBridgeRuntimeConfig(EffectiveLocalPersistenceConfig(cfg))
}

func EffectiveLocalPersistenceMode(cfg Config) string {
	mode := strings.TrimSpace(cfg.LocalPersistenceMode)
	if mode == "" {
		return defaultLocalPersistenceMode
	}
	return mode
}

func EffectiveLocalPersistenceConfig(cfg Config) Config {
	if EffectiveLocalPersistenceMode(cfg) != "sqlite_dev" {
		return cfg
	}
	cfg.ApplicationCatalogStoreMode = "sqlite_dev"
	cfg.ApplicationDraftStoreMode = "sqlite_dev"
	cfg.ApplicationPublishStoreMode = "sqlite_dev"
	cfg.APIKeyStoreMode = "sqlite_dev"
	cfg.GatewayRequestStoreMode = "sqlite_dev"
	cfg.WorkflowSavedDraftStoreMode = "sqlite_dev"
	cfg.WorkflowRunStoreMode = "sqlite_dev"
	return cfg
}

func validateLocalPersistenceConfig(cfg Config) error {
	switch EffectiveLocalPersistenceMode(cfg) {
	case "memory_dev":
		if strings.TrimSpace(cfg.LocalPersistenceMode) != "" && !localPersistenceComponentsConsistent(cfg) {
			return errors.New("memory_dev local persistence conflicts with an explicit component store mode")
		}
	case "sqlite_dev":
		if strings.TrimSpace(cfg.SQLiteDevDatabasePath) == "" {
			return errors.New("sqlite_dev local persistence requires a database path")
		}
		if !localPersistenceComponentsConsistent(cfg) {
			return errors.New("sqlite_dev local persistence conflicts with an explicit component store mode")
		}
	default:
		return errors.New("local persistence mode must be memory_dev or sqlite_dev")
	}
	return nil
}

func localPersistenceComponentsConsistent(cfg Config) bool {
	componentStoreFields := []struct {
		name string
		mode string
	}{
		{name: "application_catalog_store", mode: cfg.ApplicationCatalogStoreMode},
		{name: "application_draft_store", mode: cfg.ApplicationDraftStoreMode},
		{name: "application_publish_store", mode: cfg.ApplicationPublishStoreMode},
		{name: "api_key_store", mode: cfg.APIKeyStoreMode},
		{name: "gateway_request_store", mode: cfg.GatewayRequestStoreMode},
		{name: "workflow_saved_draft_store", mode: cfg.WorkflowSavedDraftStoreMode},
		{name: "workflow_run_store", mode: cfg.WorkflowRunStoreMode},
	}
	if EffectiveLocalPersistenceMode(cfg) == "sqlite_dev" {
		for _, component := range componentStoreFields {
			if fieldSource(cfg.FieldSources, component.name) != configSourceDefault {
				return false
			}
		}
		return true
	}
	for _, component := range componentStoreFields {
		mode := strings.TrimSpace(component.mode)
		if mode != "" && mode != "memory_dev" {
			return false
		}
	}
	return true
}

func sqliteDevSchemaStatus(localPersistenceMode string) string {
	if localPersistenceMode == "sqlite_dev" {
		return "startup_migrations_configured"
	}
	return "not_selected"
}
