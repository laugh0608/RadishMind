package config

import (
	"errors"
	"strings"
)

func ValidateServerStart(cfg Config) error {
	if err := validateLocalIdentityConfig(EffectiveLocalPersistenceConfig(cfg)); err != nil {
		return err
	}
	if cfg.ApplicationEvaluationCampaignDevEnabled {
		environment := strings.TrimSpace(cfg.ApplicationEvaluationCampaignEnvironment)
		if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.ApplicationCatalogDevHTTPEnabled || !cfg.WorkflowRAGEvaluationDevEnabled ||
			!cfg.APIKeyLifecycleDevHTTPEnabled || !cfg.GatewayRequestQuotaEnforcementDevEnabled {
			return errors.New("application evaluation campaign dev requires control plane read dev auth, application catalog dev HTTP, workflow RAG evaluation dev, API key lifecycle, and Gateway quota enforcement")
		}
		if environment != "development" && environment != "test" {
			return errors.New("application evaluation campaign environment must be development or test")
		}
		if strings.TrimSpace(cfg.GatewayProviderRouteSource) == "admin_snapshot_dev_test" && environment != strings.TrimSpace(cfg.GatewayProviderRouteEnvironment) {
			return errors.New("application evaluation campaign environment must match the Gateway provider route environment")
		}
		if environment != strings.TrimSpace(cfg.GatewayRequestQuotaEnvironment) {
			return errors.New("application evaluation campaign environment must match the Gateway quota environment")
		}
	}
	if cfg.ApplicationEvaluationScheduleRunnerDevEnabled {
		if !cfg.ApplicationEvaluationCampaignDevEnabled || !cfg.LocalIdentityDevHTTPEnabled || EffectiveControlPlaneReadAuthMode(cfg) != "local_session_dev_test" {
			return errors.New("application evaluation schedule runner dev requires campaign dev, local identity HTTP, and local session auth")
		}
	}
	if cfg.PromptTemplateDevHTTPEnabled && !cfg.ControlPlaneReadDevAuthEnabled {
		return errors.New("prompt application template dev HTTP requires control plane read dev auth")
	}
	if cfg.PromptTemplateDevWriteEnabled && !cfg.PromptTemplateDevHTTPEnabled {
		return errors.New("prompt application template dev write requires its HTTP gate")
	}
	if cfg.AgentCopilotProfileDevHTTPEnabled && !cfg.ControlPlaneReadDevAuthEnabled {
		return errors.New("agent copilot profile dev HTTP requires control plane read dev auth")
	}
	if cfg.AgentCopilotProfileDevWriteEnabled && !cfg.AgentCopilotProfileDevHTTPEnabled {
		return errors.New("agent copilot profile dev write requires its HTTP gate")
	}
	if cfg.AgentCopilotRuntimeDevHTTPEnabled && (!cfg.ControlPlaneReadDevAuthEnabled || !cfg.ApplicationCatalogDevHTTPEnabled ||
		!cfg.ApplicationDraftDevHTTPEnabled || !cfg.ApplicationPublishDevHTTPEnabled || !cfg.AgentCopilotProfileDevHTTPEnabled) {
		return errors.New("agent copilot runtime dev HTTP requires auth, catalog, draft, publish, and profile HTTP gates")
	}
	if cfg.AgentCopilotRuntimeDevWriteEnabled && !cfg.AgentCopilotRuntimeDevHTTPEnabled {
		return errors.New("agent copilot runtime dev write requires its HTTP gate")
	}
	if cfg.PromptApplicationRuntimeDevHTTPEnabled && (!cfg.ControlPlaneReadDevAuthEnabled || !cfg.ApplicationDraftDevHTTPEnabled || !cfg.ApplicationPublishDevHTTPEnabled || !cfg.PromptTemplateDevHTTPEnabled) {
		return errors.New("prompt application runtime dev HTTP requires auth, draft, publish, and template HTTP gates")
	}
	if cfg.PromptApplicationRuntimeDevWriteEnabled && !cfg.PromptApplicationRuntimeDevHTTPEnabled {
		return errors.New("prompt application runtime dev write requires its HTTP gate")
	}
	if cfg.WorkflowDefinitionReleaseDevEnabled && (!cfg.ControlPlaneReadDevAuthEnabled || !cfg.WorkflowSavedDraftDevHTTPEnabled || !cfg.WorkflowSavedDraftDevWriteEnabled) {
		return errors.New("workflow definition release dev requires control plane auth and saved workflow draft HTTP/write gates")
	}
	if cfg.WorkflowTemplateCatalogDevEnabled && (!cfg.ControlPlaneReadDevAuthEnabled || !cfg.WorkflowDefinitionReleaseDevEnabled ||
		!cfg.WorkflowSavedDraftDevHTTPEnabled || !cfg.WorkflowSavedDraftDevWriteEnabled || !cfg.ApplicationCatalogDevHTTPEnabled) {
		return errors.New("workflow template catalog dev requires control plane auth, workflow definition release, application catalog HTTP, and saved workflow draft HTTP/write gates")
	}
	if cfg.ApplicationSessionDevEnabled && (!cfg.ControlPlaneReadDevAuthEnabled || !cfg.ApplicationCatalogDevHTTPEnabled ||
		(!cfg.WorkflowDefinitionReleaseDevEnabled && !cfg.WorkflowRAGAppInvocationDevEnabled &&
			!cfg.PromptApplicationRuntimeDevHTTPEnabled && !cfg.AgentCopilotRuntimeDevHTTPEnabled)) {
		return errors.New("application session dev requires control plane auth, application catalog HTTP, and at least one supported runtime authority")
	}
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
	cfg.PromptTemplateStoreMode = "sqlite_dev"
	cfg.AgentCopilotProfileStoreMode = "sqlite_dev"
	cfg.AdminProviderRouteStoreMode = "sqlite_dev"
	cfg.APIKeyStoreMode = "sqlite_dev"
	cfg.GatewayRequestStoreMode = "sqlite_dev"
	cfg.GatewayRequestQuotaStoreMode = "sqlite_dev"
	cfg.GatewayModelPricingStoreMode = "sqlite_dev"
	cfg.WorkflowSavedDraftStoreMode = "sqlite_dev"
	cfg.WorkflowRunStoreMode = "sqlite_dev"
	cfg.LocalIdentityStoreMode = "sqlite_dev"
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
		{name: "prompt_application_template_store", mode: cfg.PromptTemplateStoreMode},
		{name: "agent_copilot_profile_store", mode: cfg.AgentCopilotProfileStoreMode},
		{name: "admin_provider_route_store", mode: cfg.AdminProviderRouteStoreMode},
		{name: "api_key_store", mode: cfg.APIKeyStoreMode},
		{name: "gateway_request_store", mode: cfg.GatewayRequestStoreMode},
		{name: "gateway_request_quota_store", mode: cfg.GatewayRequestQuotaStoreMode},
		{name: "gateway_model_pricing_store", mode: cfg.GatewayModelPricingStoreMode},
		{name: "workflow_saved_draft_store", mode: cfg.WorkflowSavedDraftStoreMode},
		{name: "workflow_run_store", mode: cfg.WorkflowRunStoreMode},
		{name: "local_identity_store", mode: cfg.LocalIdentityStoreMode},
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
