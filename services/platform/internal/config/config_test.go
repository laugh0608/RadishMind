package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestSanitizedSummaryDoesNotExposeSecrets(t *testing.T) {
	cfg := Config{
		ListenAddr:                        "127.0.0.1:7000",
		ReadHeaderTimeout:                 5 * time.Second,
		WriteTimeout:                      30 * time.Second,
		BridgeTimeout:                     45 * time.Second,
		BridgeMode:                        "stdio_pool",
		BridgeWorkerCount:                 3,
		BridgeQueueCapacity:               12,
		BridgeHandshakeTimeout:            4 * time.Second,
		PythonBinary:                      "python3",
		BridgeScript:                      "scripts/run-platform-bridge.py",
		Provider:                          "openai-compatible",
		ProviderProfile:                   "anyrouter",
		Model:                             "deepseek-chat",
		BaseURL:                           "https://example.invalid/v1",
		APIKey:                            "secret-token",
		WorkflowSavedDraftDatabaseURL:     "postgresql://database.invalid/secret",
		WorkflowSavedDraftDatabaseTimeout: 8 * time.Second,
		Temperature:                       0.2,
	}

	summary := cfg.SanitizedSummary()

	if summary.Provider != "openai-compatible" {
		t.Fatalf("unexpected provider: %s", summary.Provider)
	}
	if summary.Profile != "anyrouter" {
		t.Fatalf("unexpected profile: %s", summary.Profile)
	}
	if summary.Model != "deepseek-chat" || !summary.ModelConfigured {
		t.Fatalf("unexpected model summary: %#v", summary)
	}
	if !summary.BaseURLConfigured {
		t.Fatalf("expected base_url_configured")
	}
	if summary.CredentialState != "configured" {
		t.Fatalf("unexpected credential state: %s", summary.CredentialState)
	}
	if summary.Timeouts["bridge"] != "45s" {
		t.Fatalf("unexpected bridge timeout: %#v", summary.Timeouts)
	}
	if summary.PythonBridge.Mode != "stdio_pool" || summary.PythonBridge.WorkerCount != 3 ||
		summary.PythonBridge.QueueCapacity != 12 || summary.PythonBridge.HandshakeTimeout != "4s" {
		t.Fatalf("unexpected persistent bridge summary: %#v", summary.PythonBridge)
	}
	if summary.WorkflowSavedDraftStoreMode != "memory_dev" {
		t.Fatalf("unexpected workflow saved draft store mode: %s", summary.WorkflowSavedDraftStoreMode)
	}
	if !summary.WorkflowSavedDraftDatabaseConfigured || summary.Timeouts["workflow_saved_draft_database"] != "8s" {
		t.Fatalf("unexpected workflow saved draft database summary: %#v", summary)
	}
	if !reflect.DeepEqual(summary.MissingRequiredFields, []string{}) {
		t.Fatalf("unexpected missing required fields: %#v", summary.MissingRequiredFields)
	}
	if !reflect.DeepEqual(summary.SecretFields, []string{
		"RADISHMIND_PLATFORM_API_KEY",
		"RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_DATABASE_URL",
		"RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_MIGRATION_DATABASE_URL",
		"RADISHMIND_APPLICATION_DRAFT_DEV_TEST_DATABASE_URL",
		"RADISHMIND_APPLICATION_DRAFT_DEV_TEST_MIGRATION_DATABASE_URL",
		"RADISHMIND_APPLICATION_PUBLISH_DEV_TEST_DATABASE_URL",
		"RADISHMIND_APPLICATION_PUBLISH_DEV_TEST_MIGRATION_DATABASE_URL",
		"RADISHMIND_APPLICATION_CATALOG_DEV_TEST_DATABASE_URL",
		"RADISHMIND_APPLICATION_CATALOG_DEV_TEST_MIGRATION_DATABASE_URL",
		"RADISHMIND_API_KEY_DEV_TEST_DATABASE_URL",
		"RADISHMIND_API_KEY_DEV_TEST_MIGRATION_DATABASE_URL",
		"RADISHMIND_WORKFLOW_RUN_DEV_TEST_DATABASE_URL",
		"RADISHMIND_WORKFLOW_RUN_DEV_TEST_MIGRATION_DATABASE_URL",
		"RADISHMIND_GATEWAY_REQUEST_DEV_TEST_DATABASE_URL",
		"RADISHMIND_GATEWAY_REQUEST_DEV_TEST_MIGRATION_DATABASE_URL",
		"RADISHMIND_CONTROL_PLANE_READ_DEV_TEST_DATABASE_URL",
		"RADISHMIND_CONTROL_PLANE_READ_DEV_TEST_MIGRATION_DATABASE_URL",
	}) {
		t.Fatalf("unexpected secret fields: %#v", summary.SecretFields)
	}
}

func TestWorkflowDefinitionReleaseDevRequiresSavedDraftAuthorityGates(t *testing.T) {
	cfg := defaultConfig()
	cfg.WorkflowDefinitionReleaseDevEnabled = true
	if err := ValidateServerStart(cfg); err == nil || !strings.Contains(err.Error(), "saved workflow draft HTTP/write gates") {
		t.Fatalf("expected workflow definition release prerequisite failure, got %v", err)
	}
	cfg.ControlPlaneReadDevAuthEnabled = true
	cfg.WorkflowSavedDraftDevHTTPEnabled = true
	cfg.WorkflowSavedDraftDevWriteEnabled = true
	if err := ValidateServerStart(cfg); err != nil {
		t.Fatalf("complete workflow definition release dev gates should validate: %v", err)
	}
}

func TestLoadFromEnvAppliesConfigFileThenEnvOverride(t *testing.T) {
	clearPlatformEnv(t)
	configPath := filepath.Join(t.TempDir(), "platform-config.json")
	configDocument := []byte(`{
  "listen_addr": "127.0.0.1:9000",
  "read_header_timeout": "7s",
  "write_timeout": "35s",
  "bridge_timeout": "50s",
  "bridge_mode": "process_per_request",
  "bridge_worker_count": 2,
  "bridge_queue_capacity": 8,
  "bridge_handshake_timeout": "3s",
  "python_binary": "python-from-file",
  "bridge_script": "scripts/from-file.py",
  "provider": "openai-compatible",
  "provider_profile": "file-profile",
  "model": "file-model",
  "base_url": "https://file.example.invalid/v1",
  "api_key": "file-secret",
  "temperature": 0.4,
  "workflow_saved_draft_store": "memory_dev"
}`)
	if err := os.WriteFile(configPath, configDocument, 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}
	t.Setenv("RADISHMIND_PLATFORM_CONFIG", configPath)
	t.Setenv("RADISHMIND_PLATFORM_MODEL", "env-model")
	t.Setenv("RADISHMIND_PLATFORM_BRIDGE_TIMEOUT", "12s")
	t.Setenv("RADISHMIND_PLATFORM_BRIDGE_MODE", "stdio_pool")
	t.Setenv("RADISHMIND_PLATFORM_BRIDGE_WORKER_COUNT", "3")
	t.Setenv("RADISHMIND_PLATFORM_BRIDGE_QUEUE_CAPACITY", "9")
	t.Setenv("RADISHMIND_PLATFORM_BRIDGE_HANDSHAKE_TIMEOUT", "4s")
	t.Setenv("RADISHMIND_PLATFORM_API_KEY", "env-secret")
	t.Setenv("RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH", "1")
	t.Setenv("RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP", "1")
	t.Setenv("RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE", "true")
	t.Setenv("RADISHMIND_WORKFLOW_DEFINITION_RELEASE_DEV", "1")
	t.Setenv("RADISHMIND_WORKFLOW_EXECUTOR_DEV", "1")
	t.Setenv("RADISHMIND_GATEWAY_REQUEST_HISTORY_DEV", "1")
	t.Setenv("RADISHMIND_GATEWAY_REQUEST_STORE", "postgres_dev_test")
	t.Setenv("RADISHMIND_GATEWAY_REQUEST_DEV_TEST_DATABASE_URL", "postgresql://gateway-runtime.invalid/secret")
	t.Setenv("RADISHMIND_GATEWAY_REQUEST_DATABASE_TIMEOUT", "8s")
	t.Setenv("RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE", "repository_disabled")
	t.Setenv("RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_DATABASE_URL", "postgresql://runtime.invalid/secret")
	t.Setenv("RADISHMIND_WORKFLOW_SAVED_DRAFT_DATABASE_TIMEOUT", "9s")
	t.Setenv("RADISHMIND_APPLICATION_DRAFT_DEV_HTTP", "1")
	t.Setenv("RADISHMIND_APPLICATION_DRAFT_DEV_WRITE", "true")
	t.Setenv("RADISHMIND_APPLICATION_DRAFT_STORE", "memory_dev")
	t.Setenv("RADISHMIND_APPLICATION_DRAFT_DEV_TEST_DATABASE_URL", "postgresql://application-draft.invalid/secret")
	t.Setenv("RADISHMIND_APPLICATION_DRAFT_DATABASE_TIMEOUT", "11s")
	t.Setenv("RADISHMIND_APPLICATION_PUBLISH_DEV_HTTP", "1")
	t.Setenv("RADISHMIND_APPLICATION_PUBLISH_DEV_WRITE", "true")
	t.Setenv("RADISHMIND_APPLICATION_PUBLISH_STORE", "memory_dev")
	t.Setenv("RADISHMIND_APPLICATION_PUBLISH_DEV_TEST_DATABASE_URL", "postgresql://application-publish.invalid/secret")
	t.Setenv("RADISHMIND_APPLICATION_PUBLISH_DATABASE_TIMEOUT", "13s")
	t.Setenv("RADISHMIND_APPLICATION_CATALOG_DEV_HTTP", "1")
	t.Setenv("RADISHMIND_APPLICATION_CATALOG_DEV_WRITE", "true")
	t.Setenv("RADISHMIND_APPLICATION_CATALOG_STORE", "memory_dev")
	t.Setenv("RADISHMIND_APPLICATION_CATALOG_DEV_TEST_DATABASE_URL", "postgresql://application-catalog.invalid/secret")
	t.Setenv("RADISHMIND_APPLICATION_CATALOG_DATABASE_TIMEOUT", "14s")
	t.Setenv("RADISHMIND_API_KEY_LIFECYCLE_DEV_HTTP", "1")
	t.Setenv("RADISHMIND_API_KEY_LIFECYCLE_DEV_WRITE", "true")
	t.Setenv("RADISHMIND_API_KEY_STORE", "memory_dev")
	t.Setenv("RADISHMIND_API_KEY_DEV_TEST_DATABASE_URL", "postgresql://api-key.invalid/secret")
	t.Setenv("RADISHMIND_API_KEY_DATABASE_TIMEOUT", "15s")
	t.Setenv("RADISHMIND_GATEWAY_AUTH_MODE", "api_key_dev_test")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.ListenAddr != "127.0.0.1:9000" {
		t.Fatalf("unexpected listen addr: %s", cfg.ListenAddr)
	}
	if cfg.Model != "env-model" {
		t.Fatalf("expected env model override, got %s", cfg.Model)
	}
	if cfg.BridgeTimeout != 12*time.Second {
		t.Fatalf("expected env bridge timeout override, got %s", cfg.BridgeTimeout)
	}
	if cfg.BridgeMode != "stdio_pool" || cfg.BridgeWorkerCount != 3 || cfg.BridgeQueueCapacity != 9 ||
		cfg.BridgeHandshakeTimeout != 4*time.Second {
		t.Fatalf("expected env bridge runtime overrides, got %#v", cfg)
	}
	if cfg.APIKey != "env-secret" {
		t.Fatalf("expected env credential override")
	}
	if !cfg.ControlPlaneReadDevAuthEnabled {
		t.Fatalf("expected control plane read dev auth env override")
	}
	if !cfg.WorkflowSavedDraftDevHTTPEnabled {
		t.Fatalf("expected workflow saved draft dev http env override")
	}
	if !cfg.WorkflowSavedDraftDevWriteEnabled {
		t.Fatalf("expected workflow saved draft dev write env override")
	}
	if !cfg.WorkflowDefinitionReleaseDevEnabled || cfg.FieldSources["workflow_definition_release_dev"] != configSourceEnv {
		t.Fatalf("expected workflow definition release dev env override")
	}
	if !cfg.WorkflowExecutorDevEnabled {
		t.Fatalf("expected workflow executor dev env override")
	}
	if !cfg.GatewayRequestHistoryDevEnabled {
		t.Fatalf("expected gateway request history dev env override")
	}
	if cfg.GatewayRequestStoreMode != "postgres_dev_test" || cfg.GatewayRequestDatabaseURL == "" || cfg.GatewayRequestDatabaseTimeout != 8*time.Second {
		t.Fatalf("expected Gateway request store env overrides: %#v", cfg)
	}
	if cfg.WorkflowSavedDraftStoreMode != "repository_disabled" {
		t.Fatalf("expected workflow saved draft store env override, got %s", cfg.WorkflowSavedDraftStoreMode)
	}
	if cfg.WorkflowSavedDraftDatabaseURL == "" || cfg.WorkflowSavedDraftDatabaseTimeout != 9*time.Second {
		t.Fatalf("expected workflow saved draft database env overrides")
	}
	if !cfg.ApplicationDraftDevHTTPEnabled || !cfg.ApplicationDraftDevWriteEnabled || cfg.ApplicationDraftStoreMode != "memory_dev" ||
		cfg.ApplicationDraftDatabaseURL == "" || cfg.ApplicationDraftDatabaseTimeout != 11*time.Second {
		t.Fatalf("expected application draft env overrides: %#v", cfg)
	}
	if !cfg.ApplicationPublishDevHTTPEnabled || !cfg.ApplicationPublishDevWriteEnabled || cfg.ApplicationPublishStoreMode != "memory_dev" ||
		cfg.ApplicationPublishDatabaseURL == "" || cfg.ApplicationPublishDatabaseTimeout != 13*time.Second {
		t.Fatalf("expected application publish env overrides: %#v", cfg)
	}
	if !cfg.ApplicationCatalogDevHTTPEnabled || !cfg.ApplicationCatalogDevWriteEnabled || cfg.ApplicationCatalogStoreMode != "memory_dev" ||
		cfg.ApplicationCatalogDatabaseURL == "" || cfg.ApplicationCatalogDatabaseTimeout != 14*time.Second {
		t.Fatalf("expected application catalog env overrides: %#v", cfg)
	}
	if !cfg.APIKeyLifecycleDevHTTPEnabled || !cfg.APIKeyLifecycleDevWriteEnabled {
		t.Fatalf("expected API key lifecycle env overrides: %#v", cfg)
	}
	if cfg.APIKeyStoreMode != "memory_dev" || cfg.APIKeyDatabaseURL == "" || cfg.APIKeyDatabaseTimeout != 15*time.Second || cfg.GatewayAuthMode != "api_key_dev_test" {
		t.Fatalf("expected API key store and Gateway auth env overrides: %#v", cfg)
	}

	summary := cfg.SanitizedSummary()
	if summary.FieldSources["listen_addr"] != "file" {
		t.Fatalf("expected listen_addr source=file, got %#v", summary.FieldSources)
	}
	if summary.FieldSources["model"] != "env" {
		t.Fatalf("expected model source=env, got %#v", summary.FieldSources)
	}
	if summary.FieldSources["bridge_timeout"] != "env" {
		t.Fatalf("expected bridge_timeout source=env, got %#v", summary.FieldSources)
	}
	for _, field := range []string{
		"bridge_mode",
		"bridge_worker_count",
		"bridge_queue_capacity",
		"bridge_handshake_timeout",
	} {
		if summary.FieldSources[field] != "env" {
			t.Fatalf("expected %s source=env, got %#v", field, summary.FieldSources)
		}
	}
	if summary.FieldSources["credential"] != "env" {
		t.Fatalf("expected credential source=env, got %#v", summary.FieldSources)
	}
	if summary.FieldSources["control_plane_read_dev_auth"] != "env" {
		t.Fatalf("expected control_plane_read_dev_auth source=env, got %#v", summary.FieldSources)
	}
	if summary.FieldSources["workflow_saved_draft_dev_http"] != "env" {
		t.Fatalf("expected workflow_saved_draft_dev_http source=env, got %#v", summary.FieldSources)
	}
	if summary.FieldSources["workflow_saved_draft_dev_write"] != "env" {
		t.Fatalf("expected workflow_saved_draft_dev_write source=env, got %#v", summary.FieldSources)
	}
	if summary.FieldSources["workflow_executor_dev"] != "env" {
		t.Fatalf("expected workflow_executor_dev source=env, got %#v", summary.FieldSources)
	}
	if summary.FieldSources["gateway_request_history_dev"] != "env" {
		t.Fatalf("expected gateway_request_history_dev source=env, got %#v", summary.FieldSources)
	}
	if summary.FieldSources["gateway_request_store"] != "env" || summary.FieldSources["gateway_request_database"] != "env" || summary.FieldSources["gateway_request_database_timeout"] != "env" {
		t.Fatalf("expected Gateway request store sources=env, got %#v", summary.FieldSources)
	}
	if summary.FieldSources["workflow_saved_draft_store"] != "env" {
		t.Fatalf("expected workflow_saved_draft_store source=env, got %#v", summary.FieldSources)
	}
	if summary.FieldSources["workflow_saved_draft_database"] != "env" ||
		summary.FieldSources["workflow_saved_draft_database_timeout"] != "env" {
		t.Fatalf("expected workflow saved draft database sources=env, got %#v", summary.FieldSources)
	}
	if summary.WorkflowSavedDraftStoreMode != "repository_disabled" {
		t.Fatalf("unexpected workflow saved draft store summary: %#v", summary)
	}
	if summary.FieldSources["application_draft_dev_http"] != "env" || summary.FieldSources["application_draft_dev_write"] != "env" ||
		summary.FieldSources["application_draft_store"] != "env" || summary.FieldSources["application_draft_database"] != "env" ||
		summary.FieldSources["application_draft_database_timeout"] != "env" {
		t.Fatalf("expected application draft sources=env, got %#v", summary.FieldSources)
	}
	if summary.FieldSources["application_publish_dev_http"] != "env" || summary.FieldSources["application_publish_dev_write"] != "env" ||
		summary.FieldSources["application_publish_store"] != "env" || summary.FieldSources["application_publish_database"] != "env" ||
		summary.FieldSources["application_publish_database_timeout"] != "env" {
		t.Fatalf("expected application publish sources=env, got %#v", summary.FieldSources)
	}
	if summary.ConfigFile.Path != configPath || !summary.ConfigFile.Configured || !summary.ConfigFile.Loaded {
		t.Fatalf("unexpected config file summary: %#v", summary.ConfigFile)
	}
}

func TestWorkflowRAGSnapshotDevGateRequiresVerifiedDevAuth(t *testing.T) {
	clearPlatformEnv(t)
	t.Setenv("RADISHMIND_WORKFLOW_RAG_SNAPSHOT_DEV", "1")
	if _, err := LoadFromEnv(); err == nil || !strings.Contains(err.Error(), "requires control plane read dev auth") {
		t.Fatalf("workflow RAG snapshot dev gate did not reject missing verified dev auth: %v", err)
	}
	t.Setenv("RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH", "1")
	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("load workflow RAG snapshot dev gate with auth: %v", err)
	}
	if validateErr := ValidateServerStart(cfg); validateErr != nil {
		t.Fatalf("workflow RAG snapshot dev gate rejected complete dev auth: %v", validateErr)
	}
	if summary := cfg.SanitizedSummary(); !summary.WorkflowRAGSnapshotDevEnabled {
		t.Fatalf("workflow RAG snapshot dev gate missing from sanitized summary: %#v", summary)
	}
}

func TestWorkflowRAGExecutionDevGateIsIndependentFromSnapshotAndExecutorGates(t *testing.T) {
	clearPlatformEnv(t)
	t.Setenv("RADISHMIND_WORKFLOW_RAG_EXECUTION_DEV", "1")
	if _, err := LoadFromEnv(); err == nil || !strings.Contains(err.Error(), "requires control plane read dev auth and saved workflow draft dev HTTP") {
		t.Fatalf("workflow RAG execution gate accepted missing verified draft-read prerequisites: %v", err)
	}
	t.Setenv("RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH", "1")
	t.Setenv("RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP", "1")
	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("load independent workflow RAG execution gate: %v", err)
	}
	if cfg.WorkflowExecutorDevEnabled || cfg.WorkflowRAGSnapshotDevEnabled || !cfg.WorkflowRAGExecutionDevEnabled {
		t.Fatalf("workflow RAG execution gate reused another execution or snapshot gate: %#v", cfg.SanitizedSummary())
	}
	if validateErr := ValidateServerStart(cfg); validateErr != nil {
		t.Fatalf("independent workflow RAG execution gate was rejected: %v", validateErr)
	}
	if missing := cfg.Check(); len(missing) != 0 {
		t.Fatalf("independent workflow RAG execution gate reported missing fields: %#v", missing)
	}
	if cfg.FieldSources["workflow_rag_execution_dev"] != configSourceEnv || !cfg.SanitizedSummary().WorkflowRAGExecutionDevEnabled {
		t.Fatalf("workflow RAG execution gate source or summary is incomplete: %#v", cfg.SanitizedSummary())
	}
}

func TestWorkflowRAGEvaluationDevGateIsIndependentAndRequiresVerifiedAuth(t *testing.T) {
	clearPlatformEnv(t)
	t.Setenv("RADISHMIND_WORKFLOW_RAG_EVALUATION_DEV", "1")
	if _, err := LoadFromEnv(); err == nil || !strings.Contains(err.Error(), "requires control plane read dev auth") {
		t.Fatalf("workflow RAG evaluation gate accepted missing verified auth: %v", err)
	}
	t.Setenv("RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH", "1")
	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("load workflow RAG evaluation gate: %v", err)
	}
	if cfg.WorkflowExecutorDevEnabled || cfg.WorkflowRAGExecutionDevEnabled || cfg.WorkflowRAGSnapshotDevEnabled || !cfg.WorkflowRAGEvaluationDevEnabled {
		t.Fatalf("workflow RAG evaluation gate reused another gate: %#v", cfg.SanitizedSummary())
	}
	if cfg.FieldSources["workflow_rag_evaluation_dev"] != configSourceEnv || !cfg.SanitizedSummary().WorkflowRAGEvaluationDevEnabled {
		t.Fatalf("workflow RAG evaluation gate source or summary is incomplete: %#v", cfg.SanitizedSummary())
	}
	if validateErr := ValidateServerStart(cfg); validateErr != nil {
		t.Fatalf("workflow RAG evaluation gate rejected complete verified auth: %v", validateErr)
	}
}

func TestWorkflowRAGPromotionDevGateIsIndependentAndRequiresVerifiedAuth(t *testing.T) {
	clearPlatformEnv(t)
	t.Setenv("RADISHMIND_WORKFLOW_RAG_PROMOTION_DEV", "1")
	if _, err := LoadFromEnv(); err == nil || !strings.Contains(err.Error(), "requires control plane read dev auth") {
		t.Fatalf("workflow RAG promotion gate accepted missing verified auth: %v", err)
	}
	t.Setenv("RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH", "1")
	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("load workflow RAG promotion gate: %v", err)
	}
	if cfg.WorkflowExecutorDevEnabled || cfg.WorkflowRAGExecutionDevEnabled || cfg.WorkflowRAGSnapshotDevEnabled || cfg.WorkflowRAGEvaluationDevEnabled || !cfg.WorkflowRAGPromotionDevEnabled {
		t.Fatalf("workflow RAG promotion gate reused another gate: %#v", cfg.SanitizedSummary())
	}
	if cfg.FieldSources["workflow_rag_promotion_dev"] != configSourceEnv || !cfg.SanitizedSummary().WorkflowRAGPromotionDevEnabled {
		t.Fatalf("workflow RAG promotion gate source or summary is incomplete: %#v", cfg.SanitizedSummary())
	}
	if validateErr := ValidateServerStart(cfg); validateErr != nil {
		t.Fatalf("workflow RAG promotion gate rejected complete verified auth: %v", validateErr)
	}
}

func TestWorkflowRAGApplicationInvocationDevGateRequiresCompleteDevAuthorities(t *testing.T) {
	clearPlatformEnv(t)
	t.Setenv("RADISHMIND_WORKFLOW_RAG_APPLICATION_INVOCATION_DEV", "1")
	if _, err := LoadFromEnv(); err == nil || !strings.Contains(err.Error(), "workflow RAG application invocation dev requires") {
		t.Fatalf("incomplete application RAG invocation dependencies were accepted: %v", err)
	}
	t.Setenv("RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH", "1")
	t.Setenv("RADISHMIND_APPLICATION_DRAFT_DEV_HTTP", "1")
	t.Setenv("RADISHMIND_APPLICATION_PUBLISH_DEV_HTTP", "1")
	t.Setenv("RADISHMIND_APPLICATION_CATALOG_DEV_HTTP", "1")
	t.Setenv("RADISHMIND_API_KEY_LIFECYCLE_DEV_HTTP", "1")
	t.Setenv("RADISHMIND_WORKFLOW_RAG_PROMOTION_DEV", "1")
	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("load complete application RAG invocation gates: %v", err)
	}
	if !cfg.WorkflowRAGAppInvocationDevEnabled || cfg.FieldSources["workflow_rag_application_invocation_dev"] != configSourceEnv || !cfg.SanitizedSummary().WorkflowRAGAppInvocationDevEnabled {
		t.Fatal("sanitized summary omitted the application RAG invocation gate")
	}
}

func TestLoadFromEnvRejectsInvalidConfigFileDuration(t *testing.T) {
	clearPlatformEnv(t)
	configPath := filepath.Join(t.TempDir(), "platform-config.json")
	if err := os.WriteFile(configPath, []byte(`{"bridge_timeout":"invalid"}`), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}
	t.Setenv("RADISHMIND_PLATFORM_CONFIG", configPath)

	_, err := LoadFromEnv()
	if err == nil {
		t.Fatalf("expected invalid duration error")
	}
}

func TestLoadFromEnvUsesPersistentBridgeDefaults(t *testing.T) {
	clearPlatformEnv(t)
	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("load default config: %v", err)
	}
	if cfg.BridgeMode != "stdio_pool" || cfg.BridgeWorkerCount != 4 || cfg.BridgeQueueCapacity != 64 ||
		cfg.BridgeHandshakeTimeout != 5*time.Second {
		t.Fatalf("unexpected default bridge runtime: %#v", cfg)
	}
}

func TestLoadFromEnvRejectsInvalidBridgeRuntimeConfig(t *testing.T) {
	testCases := []struct {
		name  string
		key   string
		value string
	}{
		{name: "unknown mode", key: "RADISHMIND_PLATFORM_BRIDGE_MODE", value: "unknown"},
		{name: "zero workers", key: "RADISHMIND_PLATFORM_BRIDGE_WORKER_COUNT", value: "0"},
		{name: "queue too large", key: "RADISHMIND_PLATFORM_BRIDGE_QUEUE_CAPACITY", value: "1025"},
		{name: "zero handshake", key: "RADISHMIND_PLATFORM_BRIDGE_HANDSHAKE_TIMEOUT", value: "0s"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			clearPlatformEnv(t)
			t.Setenv(testCase.key, testCase.value)
			if _, err := LoadFromEnv(); err == nil {
				t.Fatalf("expected invalid bridge runtime config to fail: %s=%s", testCase.key, testCase.value)
			}
		})
	}
}

func TestSanitizedSummaryCredentialStates(t *testing.T) {
	cases := []struct {
		name            string
		config          Config
		expectedState   string
		expectedMissing []string
	}{
		{
			name: "mock does not require credential",
			config: Config{
				ListenAddr: ":7000",
				Provider:   "mock",
			},
			expectedState:   "not_required",
			expectedMissing: []string{},
		},
		{
			name: "remote provider requires model base url and credential",
			config: Config{
				ListenAddr: ":7000",
				Provider:   "huggingface",
			},
			expectedState:   "missing",
			expectedMissing: []string{"model", "base_url", "credential"},
		},
		{
			name: "ollama credential is optional but model and base url are required",
			config: Config{
				ListenAddr: ":7000",
				Provider:   "ollama",
				BaseURL:    "http://localhost:11434",
			},
			expectedState:   "optional_missing",
			expectedMissing: []string{"model"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			summary := testCase.config.SanitizedSummary()
			if summary.CredentialState != testCase.expectedState {
				t.Fatalf("unexpected credential state: %s", summary.CredentialState)
			}
			if !reflect.DeepEqual(summary.MissingRequiredFields, testCase.expectedMissing) {
				t.Fatalf("unexpected missing required fields: %#v", summary.MissingRequiredFields)
			}
		})
	}
}

func TestPostgresDevTestModeRequiresExplicitDevelopmentGates(t *testing.T) {
	config := Config{
		ListenAddr:                  ":7000",
		Provider:                    "mock",
		WorkflowSavedDraftStoreMode: "postgres_dev_test",
	}

	if got := config.Check(); !reflect.DeepEqual(got, []string{
		"control_plane_read_dev_auth",
		"workflow_saved_draft_dev_http",
		"workflow_saved_draft_dev_write",
		"workflow_saved_draft_database",
	}) {
		t.Fatalf("unexpected postgres dev/test missing fields: %#v", got)
	}

	config.ControlPlaneReadDevAuthEnabled = true
	config.WorkflowSavedDraftDevHTTPEnabled = true
	config.WorkflowSavedDraftDevWriteEnabled = true
	config.WorkflowSavedDraftDatabaseURL = "postgresql://runtime.invalid/secret"
	if got := config.Check(); len(got) != 0 {
		t.Fatalf("complete postgres dev/test config should pass: %#v", got)
	}
	if summary := config.SanitizedSummary(); !summary.WorkflowSavedDraftDatabaseConfigured {
		t.Fatalf("database presence must be visible without exposing its value: %#v", summary)
	}
}

func TestSQLiteWorkflowSavedDraftModeRequiresExplicitDevelopmentGates(t *testing.T) {
	config := defaultConfig()
	config.WorkflowSavedDraftStoreMode = "sqlite_dev"
	if got := config.Check(); !reflect.DeepEqual(got, []string{
		"control_plane_read_dev_auth",
		"workflow_saved_draft_dev_http",
		"workflow_saved_draft_dev_write",
	}) {
		t.Fatalf("unexpected SQLite saved draft missing fields: %#v", got)
	}
	if summary := config.SanitizedSummary(); summary.WorkflowSavedDraftStoreMode != "sqlite_dev" ||
		summary.WorkflowSavedDraftDatabaseConfigured {
		t.Fatalf("SQLite saved draft summary drifted: %#v", summary)
	}

	config.ControlPlaneReadDevAuthEnabled = true
	config.WorkflowSavedDraftDevHTTPEnabled = true
	config.WorkflowSavedDraftDevWriteEnabled = true
	if got := config.Check(); len(got) != 0 {
		t.Fatalf("complete SQLite saved draft config should pass: %#v", got)
	}
	if err := validateBridgeRuntimeConfig(config); err != nil {
		t.Fatalf("complete SQLite saved draft runtime config was rejected: %v", err)
	}

	config.WorkflowSavedDraftStoreMode = "future_backend"
	if err := validateBridgeRuntimeConfig(config); err == nil ||
		err.Error() != "saved workflow draft store must be memory_dev, sqlite_dev, postgres_dev_test, repository_disabled, or repository" {
		t.Fatalf("unknown saved draft store mode did not fail closed: %v", err)
	}
	config.WorkflowSavedDraftStoreMode = "repository_disabled"
	if err := validateBridgeRuntimeConfig(config); err != nil {
		t.Fatalf("reserved saved draft store mode should remain request-time disabled: %v", err)
	}
	config.WorkflowSavedDraftDatabaseTimeout = 0
	if err := validateBridgeRuntimeConfig(config); err == nil ||
		err.Error() != "saved workflow draft database timeout must be positive" {
		t.Fatalf("non-positive saved draft database timeout was accepted: %v", err)
	}
}

func TestApplicationDraftPostgresDevTestModeRequiresExplicitDevelopmentGates(t *testing.T) {
	config := defaultConfig()
	config.ApplicationDraftStoreMode = "postgres_dev_test"
	if got := config.Check(); !reflect.DeepEqual(got, []string{
		"control_plane_read_dev_auth",
		"application_draft_dev_http",
		"application_draft_dev_write",
		"application_draft_database",
	}) {
		t.Fatalf("unexpected application draft postgres dev/test missing fields: %#v", got)
	}
	config.ControlPlaneReadDevAuthEnabled = true
	config.ApplicationDraftDevHTTPEnabled = true
	config.ApplicationDraftDevWriteEnabled = true
	config.ApplicationDraftStoreMode = "postgres_dev_test"
	config.ApplicationDraftDatabaseURL = "postgresql://runtime.invalid/secret"
	if got := config.Check(); len(got) != 0 {
		t.Fatalf("complete application draft postgres config should pass: %#v", got)
	}
	if summary := config.SanitizedSummary(); !summary.ApplicationDraftDatabaseConfigured || summary.ApplicationDraftStoreMode != "postgres_dev_test" {
		t.Fatalf("application draft database summary must stay sanitized: %#v", summary)
	}
}

func TestApplicationPublishPostgresDevTestModeRequiresDraftAndPublishGates(t *testing.T) {
	config := defaultConfig()
	config.ApplicationPublishStoreMode = "postgres_dev_test"
	if got := config.Check(); !reflect.DeepEqual(got, []string{
		"control_plane_read_dev_auth",
		"application_draft_dev_http",
		"application_draft_dev_write",
		"application_draft_store_postgres_dev_test",
		"application_draft_database",
		"application_publish_dev_http",
		"application_publish_dev_write",
		"application_publish_database",
	}) {
		t.Fatalf("unexpected application publish postgres dev/test missing fields: %#v", got)
	}
	config.ControlPlaneReadDevAuthEnabled = true
	config.ApplicationDraftDevHTTPEnabled = true
	config.ApplicationDraftDevWriteEnabled = true
	config.ApplicationDraftStoreMode = "postgres_dev_test"
	config.ApplicationDraftDatabaseURL = "postgresql://draft-runtime.invalid/secret"
	config.ApplicationPublishDevHTTPEnabled = true
	config.ApplicationPublishDevWriteEnabled = true
	config.ApplicationPublishDatabaseURL = "postgresql://runtime.invalid/secret"
	if got := config.Check(); len(got) != 0 {
		t.Fatalf("complete application publish postgres config should pass: %#v", got)
	}
	if summary := config.SanitizedSummary(); !summary.ApplicationPublishDatabaseConfigured || summary.ApplicationPublishStoreMode != "postgres_dev_test" {
		t.Fatalf("application publish database summary must stay sanitized: %#v", summary)
	}
}

func TestApplicationCatalogPostgresDevTestModeRequiresExplicitDevelopmentGates(t *testing.T) {
	config := defaultConfig()
	config.ApplicationCatalogStoreMode = "postgres_dev_test"
	if got := config.Check(); !reflect.DeepEqual(got, []string{
		"control_plane_read_dev_auth",
		"application_catalog_dev_http",
		"application_catalog_dev_write",
		"application_catalog_database",
	}) {
		t.Fatalf("unexpected application catalog postgres dev/test missing fields: %#v", got)
	}
	config.ControlPlaneReadDevAuthEnabled = true
	config.ApplicationCatalogDevHTTPEnabled = true
	config.ApplicationCatalogDevWriteEnabled = true
	config.ApplicationCatalogDatabaseURL = "postgresql://runtime.invalid/secret"
	if got := config.Check(); len(got) != 0 {
		t.Fatalf("complete application catalog postgres config should pass: %#v", got)
	}
	if summary := config.SanitizedSummary(); !summary.ApplicationCatalogDatabaseConfigured || summary.ApplicationCatalogStoreMode != "postgres_dev_test" {
		t.Fatalf("application catalog database summary must stay sanitized: %#v", summary)
	}
}

func TestAPIKeyLifecycleRequiresExplicitApplicationCatalogGates(t *testing.T) {
	config := defaultConfig()
	config.APIKeyLifecycleDevHTTPEnabled = true
	if err := validateBridgeRuntimeConfig(config); err == nil || err.Error() != "API key lifecycle dev HTTP requires control plane read dev auth" {
		t.Fatalf("unexpected missing auth validation: %v", err)
	}
	config.ControlPlaneReadDevAuthEnabled = true
	if err := validateBridgeRuntimeConfig(config); err == nil || err.Error() != "API key lifecycle dev HTTP requires application catalog dev HTTP" {
		t.Fatalf("unexpected missing catalog validation: %v", err)
	}
	config.ApplicationCatalogDevHTTPEnabled = true
	config.APIKeyLifecycleDevWriteEnabled = true
	if err := validateBridgeRuntimeConfig(config); err == nil || err.Error() != "API key lifecycle dev write requires application catalog dev write" {
		t.Fatalf("unexpected missing catalog write validation: %v", err)
	}
	config.ApplicationCatalogDevWriteEnabled = true
	if err := validateBridgeRuntimeConfig(config); err != nil {
		t.Fatalf("complete API key lifecycle development gates should pass: %v", err)
	}
	if summary := config.SanitizedSummary(); !summary.APIKeyLifecycleDevHTTPEnabled || !summary.APIKeyLifecycleDevWriteEnabled {
		t.Fatalf("API key lifecycle summary must expose only enablement state: %#v", summary)
	}
}

func TestAPIKeyPostgresAndGatewayAuthRequireCompleteDevelopmentChain(t *testing.T) {
	cfg := defaultConfig()
	cfg.APIKeyStoreMode = "postgres_dev_test"
	if got := cfg.SanitizedSummary().MissingRequiredFields; !reflect.DeepEqual(got, []string{
		"control_plane_read_dev_auth", "api_key_lifecycle_dev_http", "api_key_lifecycle_dev_write", "api_key_database",
	}) {
		t.Fatalf("unexpected API key PostgreSQL requirements: %#v", got)
	}
	if err := validateBridgeRuntimeConfig(cfg); err == nil {
		t.Fatal("incomplete API key PostgreSQL store was accepted")
	}
	cfg.ControlPlaneReadDevAuthEnabled = true
	cfg.ApplicationCatalogDevHTTPEnabled = true
	cfg.ApplicationCatalogDevWriteEnabled = true
	cfg.APIKeyLifecycleDevHTTPEnabled = true
	cfg.APIKeyLifecycleDevWriteEnabled = true
	cfg.APIKeyDatabaseURL = "postgresql://runtime.invalid/secret"
	if err := validateBridgeRuntimeConfig(cfg); err != nil {
		t.Fatalf("complete API key PostgreSQL config was rejected: %v", err)
	}
	summary := cfg.SanitizedSummary()
	if summary.APIKeyStoreMode != "postgres_dev_test" || !summary.APIKeyDatabaseConfigured || summary.GatewayAuthMode != "dev_headers" {
		t.Fatalf("API key store summary is incomplete or exposed a secret: %#v", summary)
	}

	cfg.GatewayAuthMode = "api_key_dev_test"
	if err := validateBridgeRuntimeConfig(cfg); err == nil || err.Error() != "Gateway api_key_dev_test auth requires API key lifecycle HTTP and Gateway request history" {
		t.Fatalf("Gateway API key auth accepted missing request history: %v", err)
	}
	cfg.GatewayRequestHistoryDevEnabled = true
	if err := validateBridgeRuntimeConfig(cfg); err != nil {
		t.Fatalf("complete Gateway API key auth chain was rejected: %v", err)
	}
	cfg.GatewayAuthMode = "unknown"
	if err := validateBridgeRuntimeConfig(cfg); err == nil || err.Error() != "Gateway auth mode must be dev_headers or api_key_dev_test" {
		t.Fatalf("unknown Gateway auth mode was not rejected: %v", err)
	}
}

func TestPostgresWorkflowRunModeRequiresExplicitDevelopmentGates(t *testing.T) {
	cfg := defaultConfig()
	cfg.WorkflowRunStoreMode = "postgres_dev_test"
	summary := cfg.SanitizedSummary()
	if !reflect.DeepEqual(summary.MissingRequiredFields, []string{
		"control_plane_read_dev_auth", "workflow_executor_dev", "workflow_run_database",
	}) {
		t.Fatalf("unexpected workflow run PostgreSQL requirements: %#v", summary.MissingRequiredFields)
	}
	cfg.ControlPlaneReadDevAuthEnabled = true
	cfg.WorkflowExecutorDevEnabled = true
	cfg.WorkflowSavedDraftDevHTTPEnabled = true
	cfg.WorkflowRunDatabaseURL = "postgresql://runtime.invalid/secret"
	summary = cfg.SanitizedSummary()
	if len(summary.MissingRequiredFields) != 0 || !summary.WorkflowRunDatabaseConfigured ||
		summary.WorkflowRunStoreMode != "postgres_dev_test" {
		t.Fatalf("workflow run PostgreSQL config should be ready and sanitized: %#v", summary)
	}
}

func TestSQLiteWorkflowRunModeRequiresExplicitDevelopmentGates(t *testing.T) {
	cfg := defaultConfig()
	cfg.WorkflowRunStoreMode = "sqlite_dev"
	summary := cfg.SanitizedSummary()
	if !reflect.DeepEqual(summary.MissingRequiredFields, []string{
		"control_plane_read_dev_auth", "workflow_executor_dev",
	}) {
		t.Fatalf("unexpected workflow run SQLite requirements: %#v", summary.MissingRequiredFields)
	}
	if err := validateBridgeRuntimeConfig(cfg); err == nil ||
		err.Error() != "workflow run sqlite_dev store requires control plane read dev auth, an enabled workflow runtime product, and saved workflow draft dev HTTP for execution products" {
		t.Fatalf("incomplete workflow run SQLite config was accepted: %v", err)
	}
	cfg.ControlPlaneReadDevAuthEnabled = true
	cfg.WorkflowSavedDraftDevHTTPEnabled = true
	cfg.WorkflowExecutorDevEnabled = true
	if err := validateBridgeRuntimeConfig(cfg); err != nil {
		t.Fatalf("complete workflow run SQLite config was rejected: %v", err)
	}
	summary = cfg.SanitizedSummary()
	if len(summary.MissingRequiredFields) != 0 || summary.WorkflowRunStoreMode != "sqlite_dev" ||
		summary.WorkflowRunDatabaseConfigured {
		t.Fatalf("workflow run SQLite config should be ready and sanitized: %#v", summary)
	}
	cfg.WorkflowExecutorDevEnabled = false
	cfg.WorkflowRAGExecutionDevEnabled = true
	if summary = cfg.SanitizedSummary(); len(summary.MissingRequiredFields) != 0 {
		t.Fatalf("independent RAG execution gate did not satisfy workflow run SQLite requirements: %#v", summary.MissingRequiredFields)
	}
	cfg.WorkflowRAGExecutionDevEnabled = false
	cfg.WorkflowSavedDraftDevHTTPEnabled = false
	cfg.WorkflowRAGEvaluationDevEnabled = true
	if err := validateBridgeRuntimeConfig(cfg); err != nil {
		t.Fatalf("independent RAG evaluation gate did not enable the shared SQLite backend: %v", err)
	}
}

func TestPostgresGatewayRequestModeRequiresExplicitDevelopmentGates(t *testing.T) {
	cfg := defaultConfig()
	cfg.GatewayRequestStoreMode = "postgres_dev_test"
	summary := cfg.SanitizedSummary()
	if !reflect.DeepEqual(summary.MissingRequiredFields, []string{
		"control_plane_read_dev_auth", "gateway_request_history_dev", "gateway_request_database",
	}) {
		t.Fatalf("unexpected Gateway request PostgreSQL requirements: %#v", summary.MissingRequiredFields)
	}
	cfg.ControlPlaneReadDevAuthEnabled = true
	cfg.GatewayRequestHistoryDevEnabled = true
	cfg.GatewayRequestDatabaseURL = "postgresql://runtime.invalid/secret"
	summary = cfg.SanitizedSummary()
	if len(summary.MissingRequiredFields) != 0 || !summary.GatewayRequestDatabaseConfigured || summary.GatewayRequestStoreMode != "postgres_dev_test" {
		t.Fatalf("Gateway request PostgreSQL config should be ready and sanitized: %#v", summary)
	}
}

func TestControlPlaneReadAuthModesFailClosed(t *testing.T) {
	cfg := defaultConfig()
	if got := EffectiveControlPlaneReadAuthMode(cfg); got != "disabled" {
		t.Fatalf("unexpected default auth mode: %s", got)
	}

	cfg.ControlPlaneReadDevAuthEnabled = true
	if got := EffectiveControlPlaneReadAuthMode(cfg); got != "dev_headers" {
		t.Fatalf("legacy dev auth must map to dev_headers, got %s", got)
	}

	cfg = defaultConfig()
	cfg.ControlPlaneReadAuthMode = "signed_test_token"
	if err := validateBridgeRuntimeConfig(cfg); err == nil {
		t.Fatal("signed test auth without verifier material was accepted")
	}
	cfg.ControlPlaneReadTestIssuer = "https://radish.test/oidc"
	cfg.ControlPlaneReadTestAudience = "radishmind-control-plane"
	cfg.ControlPlaneReadTestPublicKeyPEM = "not-a-public-key"
	if err := validateBridgeRuntimeConfig(cfg); err == nil {
		t.Fatal("signed test auth with invalid public key material was accepted")
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate test RSA key: %v", err)
	}
	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("marshal test RSA public key: %v", err)
	}
	cfg.ControlPlaneReadTestPublicKeyPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicKeyDER}))
	if err := validateBridgeRuntimeConfig(cfg); err != nil {
		t.Fatalf("complete signed test auth config was rejected: %v", err)
	}
	summary := cfg.SanitizedSummary()
	if summary.ControlPlaneReadAuthMode != "signed_test_token" || !summary.ControlPlaneReadTestConfigured {
		t.Fatalf("unexpected signed test auth summary: %#v", summary)
	}

	cfg.ControlPlaneReadStoreMode = "postgres_dev_test"
	if err := validateBridgeRuntimeConfig(cfg); err == nil {
		t.Fatal("control plane PostgreSQL store without database URL was accepted")
	}
	cfg.ControlPlaneReadDatabaseURL = "postgresql://runtime.invalid/secret"
	if err := validateBridgeRuntimeConfig(cfg); err != nil {
		t.Fatalf("complete control plane PostgreSQL config was rejected: %v", err)
	}
	summary = cfg.SanitizedSummary()
	if summary.ControlPlaneReadStoreMode != "postgres_dev_test" || !summary.ControlPlaneReadDatabaseConfigured || len(summary.MissingRequiredFields) != 0 {
		t.Fatalf("unexpected control plane PostgreSQL summary: %#v", summary)
	}

	cfg.ControlPlaneReadAuthMode = "production_oidc"
	if err := validateBridgeRuntimeConfig(cfg); err == nil {
		t.Fatal("unknown control plane read auth mode was accepted")
	}
}

func TestControlPlaneReadOIDCIntegrationConfigFailsClosed(t *testing.T) {
	cfg := defaultConfig()
	cfg.ControlPlaneReadAuthMode = "radish_oidc_integration_test"
	if err := validateBridgeRuntimeConfig(cfg); err == nil {
		t.Fatal("OIDC integration mode without reviewed policy was accepted")
	}
	cfg.ControlPlaneReadStoreMode = "postgres_dev_test"
	cfg.ControlPlaneReadDatabaseURL = "postgresql://runtime.invalid/secret"
	cfg.ControlPlaneReadOIDCIssuer = "http://127.0.0.1:18080/issuer"
	cfg.ControlPlaneReadOIDCDiscoveryURL = "http://127.0.0.1:18080/.well-known/openid-configuration"
	cfg.ControlPlaneReadOIDCAudience = "audience:radishmind-integration"
	cfg.ControlPlaneReadOIDCMappingVersion = "mapping:reviewed-v1"
	cfg.ControlPlaneReadOIDCEvidenceRef = "issuer:reviewed-evidence"
	cfg.ControlPlaneReadOIDCSubjectClaim = "sub"
	cfg.ControlPlaneReadOIDCTenantClaim = "tenant_id"
	cfg.ControlPlaneReadOIDCPermissionClaim = "permissions"
	cfg.ControlPlaneReadOIDCTenantPermission = "permission:tenant-read"
	cfg.ControlPlaneReadOIDCAuditPermission = "permission:audit-read"
	cfg.ControlPlaneReadOIDCAlgorithms = "RS256"
	cfg.ControlPlaneReadOIDCJWKSOrigin = "http://127.0.0.1:18080"
	if err := validateBridgeRuntimeConfig(cfg); err != nil {
		t.Fatalf("complete deterministic OIDC integration policy was rejected: %v", err)
	}
	summary := cfg.SanitizedSummary()
	if !summary.ControlPlaneReadOIDCConfigured || summary.ControlPlaneReadOIDCMappingVersion != "mapping:reviewed-v1" ||
		summary.ControlPlaneReadOIDCEvidenceRef != "issuer:reviewed-evidence" {
		t.Fatalf("unexpected sanitized OIDC summary: %#v", summary)
	}
	cfg.ControlPlaneReadStoreMode = "fake_store_dev"
	if err := validateBridgeRuntimeConfig(cfg); err == nil {
		t.Fatal("OIDC integration mode was allowed with fake store")
	}
	cfg.ControlPlaneReadStoreMode = "postgres_dev_test"
	cfg.ControlPlaneReadOIDCAlgorithms = "RS256,HS256"
	if err := validateBridgeRuntimeConfig(cfg); err == nil {
		t.Fatal("OIDC integration mode accepted an unsafe algorithm")
	}
}

func TestWorkflowExecutorDevModeRequiresDevelopmentAuthAndSavedDraftHTTP(t *testing.T) {
	config := Config{
		ListenAddr:                 ":7000",
		Provider:                   "mock",
		WorkflowExecutorDevEnabled: true,
	}

	if got := config.Check(); !reflect.DeepEqual(got, []string{
		"control_plane_read_dev_auth",
		"workflow_saved_draft_dev_http",
	}) {
		t.Fatalf("unexpected executor dev missing fields: %#v", got)
	}

	config.ControlPlaneReadDevAuthEnabled = true
	config.WorkflowSavedDraftDevHTTPEnabled = true
	if got := config.Check(); len(got) != 0 {
		t.Fatalf("complete executor dev config should pass: %#v", got)
	}
	if summary := config.SanitizedSummary(); !summary.WorkflowExecutorDevEnabled {
		t.Fatalf("executor dev gate should be visible in sanitized summary: %#v", summary)
	}
}

func TestWorkflowToolActionAndExecutionDevRequireLayeredGates(t *testing.T) {
	config := defaultConfig()
	config.WorkflowToolActionDevEnabled = true
	if got := config.Check(); !reflect.DeepEqual(got, []string{
		"control_plane_read_dev_auth",
		"workflow_saved_draft_dev_http",
		"workflow_executor_dev",
	}) {
		t.Fatalf("unexpected workflow tool action missing fields: %#v", got)
	}
	if err := validateBridgeRuntimeConfig(config); err == nil || !strings.Contains(err.Error(), "workflow tool action dev requires") {
		t.Fatalf("incomplete workflow tool action gates were accepted: %v", err)
	}
	config.ControlPlaneReadDevAuthEnabled = true
	config.WorkflowSavedDraftDevHTTPEnabled = true
	config.WorkflowExecutorDevEnabled = true
	if err := validateBridgeRuntimeConfig(config); err != nil {
		t.Fatalf("complete pre-run workflow tool action gates were rejected: %v", err)
	}
	summary := config.SanitizedSummary()
	if !summary.WorkflowToolActionDevEnabled || summary.WorkflowHTTPToolExecutionDevEnabled {
		t.Fatalf("workflow tool gates drifted in sanitized summary: %#v", summary)
	}
	config.WorkflowHTTPToolExecutionDevEnabled = true
	if err := validateBridgeRuntimeConfig(config); err != nil {
		t.Fatalf("complete workflow HTTP tool execution gates were rejected: %v", err)
	}
	config.WorkflowHTTPToolTestLoopbackEnabled = true
	if err := validateBridgeRuntimeConfig(config); err != nil {
		t.Fatalf("complete workflow HTTP tool test loopback gates were rejected: %v", err)
	}
	if !config.SanitizedSummary().WorkflowHTTPToolTestLoopbackEnabled {
		t.Fatal("workflow HTTP tool test loopback gate is missing from sanitized summary")
	}
	config.WorkflowHTTPToolExecutionDevEnabled = false
	if got := config.Check(); !reflect.DeepEqual(got, []string{"workflow_http_tool_execution_dev"}) {
		t.Fatalf("test loopback gate did not require execution dev: %#v", got)
	}
	if err := validateBridgeRuntimeConfig(config); err == nil || !strings.Contains(err.Error(), "test loopback requires") {
		t.Fatalf("test loopback accepted a missing execution gate: %v", err)
	}
	config.WorkflowHTTPToolExecutionDevEnabled = true
	config.WorkflowToolActionDevEnabled = false
	if got := config.Check(); !reflect.DeepEqual(got, []string{"workflow_tool_action_dev"}) {
		t.Fatalf("execution gate did not require the action gate: %#v", got)
	}
	if err := validateBridgeRuntimeConfig(config); err == nil || !strings.Contains(err.Error(), "requires workflow tool action dev") {
		t.Fatalf("execution gate accepted a missing action gate: %v", err)
	}
}

func TestWorkflowToolActionEnvironmentGatesAreParsedExplicitly(t *testing.T) {
	t.Setenv("RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH", "1")
	t.Setenv("RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP", "1")
	t.Setenv("RADISHMIND_WORKFLOW_EXECUTOR_DEV", "1")
	t.Setenv("RADISHMIND_WORKFLOW_TOOL_ACTION_DEV", "1")
	t.Setenv("RADISHMIND_WORKFLOW_HTTP_TOOL_EXECUTION_DEV", "1")
	t.Setenv("RADISHMIND_WORKFLOW_HTTP_TOOL_TEST_LOOPBACK", "1")
	config, err := LoadFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if !config.WorkflowToolActionDevEnabled || !config.WorkflowHTTPToolExecutionDevEnabled ||
		!config.WorkflowHTTPToolTestLoopbackEnabled {
		t.Fatalf("workflow tool action environment gates were not parsed: %#v", config.SanitizedSummary())
	}
	if config.FieldSources["workflow_tool_action_dev"] != configSourceEnv ||
		config.FieldSources["workflow_http_tool_execution_dev"] != configSourceEnv ||
		config.FieldSources["workflow_http_tool_test_loopback"] != configSourceEnv {
		t.Fatalf("workflow tool action gate sources drifted: %#v", config.FieldSources)
	}
}

func TestWorkflowDiagnosticsDevRequiresExecutorAndMockProvider(t *testing.T) {
	base := defaultConfig()
	base.WorkflowDiagnosticsDevEnabled = true
	if err := validateBridgeRuntimeConfig(base); err == nil {
		t.Fatal("workflow diagnostics dev was accepted without executor dev")
	}
	base.WorkflowExecutorDevEnabled = true
	base.Provider = "openai-compatible"
	if err := validateBridgeRuntimeConfig(base); err == nil {
		t.Fatal("workflow diagnostics dev was accepted with a non-mock provider")
	}
	base.Provider = "mock"
	if err := validateBridgeRuntimeConfig(base); err != nil {
		t.Fatalf("explicit mock diagnostics config was rejected: %v", err)
	}
	base = defaultConfig()
	base.GatewayRequestHistoryDevEnabled = true
	if err := validateBridgeRuntimeConfig(base); err == nil {
		t.Fatal("gateway request history dev was accepted without control plane read dev auth")
	}
	base.ControlPlaneReadDevAuthEnabled = true
	if err := validateBridgeRuntimeConfig(base); err != nil {
		t.Fatalf("explicit gateway request history dev config was rejected: %v", err)
	}
}

func TestSQLiteDevLocalPersistenceProjectsEffectiveStoresAndRequiresDevelopmentGates(t *testing.T) {
	clearPlatformEnv(t)
	databasePath := filepath.Join(t.TempDir(), "radishmind.db")
	t.Setenv("RADISHMIND_LOCAL_PERSISTENCE_MODE", "sqlite_dev")
	t.Setenv("RADISHMIND_SQLITE_DEV_DATABASE_PATH", databasePath)

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("load sqlite_dev local persistence config: %v", err)
	}
	if cfg.LocalPersistenceMode != "sqlite_dev" || cfg.SQLiteDevDatabasePath != databasePath {
		t.Fatalf("unexpected sqlite_dev config: %#v", cfg)
	}
	summary := cfg.SanitizedSummary()
	if summary.LocalPersistenceMode != "sqlite_dev" || !summary.LocalPersistenceConfigured ||
		!summary.SQLiteDevDatabaseConfigured || !summary.LocalPersistenceComponentsConsistent ||
		summary.SQLiteDevSchemaStatus != "startup_migrations_configured" ||
		summary.ApplicationCatalogStoreMode != "sqlite_dev" || summary.ApplicationDraftStoreMode != "sqlite_dev" ||
		summary.ApplicationPublishStoreMode != "sqlite_dev" || summary.APIKeyStoreMode != "sqlite_dev" ||
		summary.GatewayRequestStoreMode != "sqlite_dev" || summary.WorkflowSavedDraftStoreMode != "sqlite_dev" ||
		summary.WorkflowRunStoreMode != "sqlite_dev" {
		t.Fatalf("unexpected sqlite_dev sanitized summary: %#v", summary)
	}
	if !reflect.DeepEqual(summary.MissingRequiredFields, []string{
		"control_plane_read_dev_auth",
		"workflow_saved_draft_dev_http",
		"workflow_saved_draft_dev_write",
		"application_draft_dev_http",
		"application_draft_dev_write",
		"application_publish_dev_http",
		"application_publish_dev_write",
		"application_catalog_dev_http",
		"application_catalog_dev_write",
		"api_key_lifecycle_dev_http",
		"api_key_lifecycle_dev_write",
		"gateway_request_history_dev",
		"workflow_executor_dev",
	}) {
		t.Fatalf("sqlite_dev must expose incomplete development gates: %#v", summary.MissingRequiredFields)
	}
	encoded, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("encode sqlite_dev summary: %v", err)
	}
	if strings.Contains(string(encoded), databasePath) {
		t.Fatalf("sanitized summary leaked SQLite database path: %s", encoded)
	}
	if err := ValidateServerStart(cfg); err == nil || err.Error() != "saved workflow draft sqlite_dev store requires complete development gates" {
		t.Fatalf("sqlite_dev without development gates must fail before runtime construction, got %v", err)
	}

	cfg.ControlPlaneReadDevAuthEnabled = true
	cfg.WorkflowSavedDraftDevHTTPEnabled = true
	cfg.WorkflowSavedDraftDevWriteEnabled = true
	cfg.ApplicationDraftDevHTTPEnabled = true
	cfg.ApplicationDraftDevWriteEnabled = true
	cfg.ApplicationPublishDevHTTPEnabled = true
	cfg.ApplicationPublishDevWriteEnabled = true
	cfg.ApplicationCatalogDevHTTPEnabled = true
	cfg.ApplicationCatalogDevWriteEnabled = true
	cfg.APIKeyLifecycleDevHTTPEnabled = true
	cfg.APIKeyLifecycleDevWriteEnabled = true
	cfg.GatewayRequestHistoryDevEnabled = true
	cfg.WorkflowExecutorDevEnabled = true
	if err := ValidateServerStart(cfg); err != nil {
		t.Fatalf("complete sqlite_dev aggregate start config was rejected: %v", err)
	}
	if missing := cfg.SanitizedSummary().MissingRequiredFields; len(missing) != 0 {
		t.Fatalf("complete sqlite_dev aggregate config reported missing fields: %#v", missing)
	}
}

func TestLocalPersistenceRejectsUnknownModeAndComponentConflict(t *testing.T) {
	base := defaultConfig()
	base.LocalPersistenceMode = "unknown"
	if err := validateBridgeRuntimeConfig(base); err == nil || err.Error() != "local persistence mode must be memory_dev or sqlite_dev" {
		t.Fatalf("unknown local persistence mode must fail, got %v", err)
	}

	base = defaultConfig()
	base.LocalPersistenceMode = "sqlite_dev"
	base.FieldSources["api_key_store"] = configSourceEnv
	if err := validateBridgeRuntimeConfig(base); err == nil || err.Error() != "sqlite_dev local persistence conflicts with an explicit component store mode" {
		t.Fatalf("sqlite_dev component override must fail, got %v", err)
	}

	base = defaultConfig()
	base.LocalPersistenceMode = "memory_dev"
	base.ApplicationCatalogStoreMode = "postgres_dev_test"
	if err := validateBridgeRuntimeConfig(base); err == nil || err.Error() != "memory_dev local persistence conflicts with an explicit component store mode" {
		t.Fatalf("memory_dev component conflict must fail, got %v", err)
	}
}

func TestApplicationCatalogSQLiteComponentSelectorRequiresDevelopmentGates(t *testing.T) {
	base := defaultConfig()
	base.ApplicationCatalogStoreMode = "sqlite_dev"
	if err := validateBridgeRuntimeConfig(base); err == nil || err.Error() != "application catalog sqlite_dev store requires complete development gates" {
		t.Fatalf("application catalog sqlite_dev without development gates must fail, got %v", err)
	}
	base.ControlPlaneReadDevAuthEnabled = true
	base.ApplicationCatalogDevHTTPEnabled = true
	base.ApplicationCatalogDevWriteEnabled = true
	if err := validateBridgeRuntimeConfig(base); err != nil {
		t.Fatalf("application catalog sqlite_dev component selector was rejected: %v", err)
	}
}

func TestApplicationConfigurationAndPublishSQLiteSelectorsRequireCompleteChain(t *testing.T) {
	base := defaultConfig()
	base.ApplicationDraftStoreMode = "sqlite_dev"
	if summary := base.SanitizedSummary(); !reflect.DeepEqual(summary.MissingRequiredFields, []string{
		"control_plane_read_dev_auth", "application_draft_dev_http", "application_draft_dev_write",
	}) {
		t.Fatalf("unexpected application draft sqlite_dev requirements: %#v", summary.MissingRequiredFields)
	}
	if err := validateBridgeRuntimeConfig(base); err == nil || err.Error() != "application draft sqlite_dev store requires complete development gates" {
		t.Fatalf("application draft sqlite_dev without gates must fail, got %v", err)
	}

	base.ControlPlaneReadDevAuthEnabled = true
	base.ApplicationDraftDevHTTPEnabled = true
	base.ApplicationDraftDevWriteEnabled = true
	base.ApplicationPublishStoreMode = "sqlite_dev"
	if summary := base.SanitizedSummary(); !reflect.DeepEqual(summary.MissingRequiredFields, []string{
		"application_publish_dev_http", "application_publish_dev_write",
	}) {
		t.Fatalf("unexpected application publish sqlite_dev requirements: %#v", summary.MissingRequiredFields)
	}
	if err := validateBridgeRuntimeConfig(base); err == nil || err.Error() != "application publish sqlite_dev store requires complete development gates and sqlite_dev application draft store" {
		t.Fatalf("application publish sqlite_dev without publish gates must fail, got %v", err)
	}

	base.ApplicationPublishDevHTTPEnabled = true
	base.ApplicationPublishDevWriteEnabled = true
	if err := validateBridgeRuntimeConfig(base); err != nil {
		t.Fatalf("complete application draft and publish sqlite_dev chain was rejected: %v", err)
	}
	if missing := base.SanitizedSummary().MissingRequiredFields; len(missing) != 0 {
		t.Fatalf("complete application draft and publish sqlite_dev chain reported missing fields: %#v", missing)
	}

	base.ApplicationDraftStoreMode = "memory_dev"
	if err := validateBridgeRuntimeConfig(base); err == nil || err.Error() != "application publish sqlite_dev store requires complete development gates and sqlite_dev application draft store" {
		t.Fatalf("publish sqlite_dev with memory draft store must fail, got %v", err)
	}
}

func TestApplicationConfigurationAndPublishRejectUnknownStoreModes(t *testing.T) {
	base := defaultConfig()
	base.ApplicationDraftStoreMode = "unknown"
	if err := validateBridgeRuntimeConfig(base); err == nil || err.Error() != "application draft store must be memory_dev, sqlite_dev, or postgres_dev_test" {
		t.Fatalf("unknown application draft store mode must fail, got %v", err)
	}
	base = defaultConfig()
	base.ApplicationPublishStoreMode = "unknown"
	if err := validateBridgeRuntimeConfig(base); err == nil || err.Error() != "application publish store must be memory_dev, sqlite_dev, or postgres_dev_test" {
		t.Fatalf("unknown application publish store mode must fail, got %v", err)
	}
}

func TestAPIKeySQLiteSelectorRequiresApplicationCatalogChain(t *testing.T) {
	base := defaultConfig()
	base.APIKeyStoreMode = "sqlite_dev"
	if summary := base.SanitizedSummary(); !reflect.DeepEqual(summary.MissingRequiredFields, []string{
		"control_plane_read_dev_auth", "application_catalog_dev_http", "application_catalog_dev_write",
		"application_catalog_store_sqlite_dev", "api_key_lifecycle_dev_http", "api_key_lifecycle_dev_write",
	}) {
		t.Fatalf("unexpected API key sqlite_dev requirements: %#v", summary.MissingRequiredFields)
	}
	if err := validateBridgeRuntimeConfig(base); err == nil || err.Error() != "API key sqlite_dev store requires complete development gates and sqlite_dev application catalog store" {
		t.Fatalf("API key sqlite_dev without gates must fail, got %v", err)
	}

	base.ControlPlaneReadDevAuthEnabled = true
	base.ApplicationCatalogDevHTTPEnabled = true
	base.ApplicationCatalogDevWriteEnabled = true
	base.ApplicationCatalogStoreMode = "sqlite_dev"
	base.APIKeyLifecycleDevHTTPEnabled = true
	base.APIKeyLifecycleDevWriteEnabled = true
	if err := validateBridgeRuntimeConfig(base); err != nil {
		t.Fatalf("complete API key sqlite_dev chain was rejected: %v", err)
	}
	if missing := base.SanitizedSummary().MissingRequiredFields; len(missing) != 0 {
		t.Fatalf("complete API key sqlite_dev chain reported missing fields: %#v", missing)
	}

	base.ApplicationCatalogStoreMode = "memory_dev"
	if err := validateBridgeRuntimeConfig(base); err == nil || err.Error() != "API key sqlite_dev store requires complete development gates and sqlite_dev application catalog store" {
		t.Fatalf("API key sqlite_dev with memory application catalog must fail, got %v", err)
	}
}

func TestAPIKeyRejectsUnknownStoreMode(t *testing.T) {
	base := defaultConfig()
	base.APIKeyStoreMode = "unknown"
	if err := validateBridgeRuntimeConfig(base); err == nil || err.Error() != "API key store must be memory_dev, sqlite_dev, or postgres_dev_test" {
		t.Fatalf("unknown API key store mode must fail, got %v", err)
	}
}

func TestGatewayRequestSQLiteSelectorRequiresDevelopmentGates(t *testing.T) {
	base := defaultConfig()
	base.GatewayRequestStoreMode = "sqlite_dev"
	if summary := base.SanitizedSummary(); !reflect.DeepEqual(summary.MissingRequiredFields, []string{
		"control_plane_read_dev_auth", "gateway_request_history_dev",
	}) {
		t.Fatalf("unexpected Gateway request sqlite_dev requirements: %#v", summary.MissingRequiredFields)
	}
	if err := validateBridgeRuntimeConfig(base); err == nil ||
		err.Error() != "Gateway request sqlite_dev store requires complete development gates" {
		t.Fatalf("Gateway request sqlite_dev without gates must fail, got %v", err)
	}

	base.ControlPlaneReadDevAuthEnabled = true
	base.GatewayRequestHistoryDevEnabled = true
	if err := validateBridgeRuntimeConfig(base); err != nil {
		t.Fatalf("complete Gateway request sqlite_dev selector was rejected: %v", err)
	}
	if missing := base.SanitizedSummary().MissingRequiredFields; len(missing) != 0 {
		t.Fatalf("complete Gateway request sqlite_dev selector reported missing fields: %#v", missing)
	}
	if summary := base.SanitizedSummary(); summary.GatewayRequestStoreMode != "sqlite_dev" ||
		summary.GatewayRequestDatabaseConfigured {
		t.Fatalf("Gateway request sqlite_dev summary drifted: %#v", summary)
	}
}

func TestGatewayRequestStoreRejectsUnknownModeAndInvalidTimeout(t *testing.T) {
	base := defaultConfig()
	base.GatewayRequestStoreMode = "unknown"
	if err := validateBridgeRuntimeConfig(base); err == nil ||
		err.Error() != "Gateway request store must be memory_dev, sqlite_dev, or postgres_dev_test" {
		t.Fatalf("unknown Gateway request store mode must fail, got %v", err)
	}

	base = defaultConfig()
	base.GatewayRequestDatabaseTimeout = 0
	if err := validateBridgeRuntimeConfig(base); err == nil ||
		err.Error() != "Gateway request database timeout must be positive" {
		t.Fatalf("invalid Gateway request database timeout must fail, got %v", err)
	}
}

func clearPlatformEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"RADISHMIND_PLATFORM_CONFIG",
		"RADISHMIND_PLATFORM_LISTEN_ADDR",
		"RADISHMIND_PLATFORM_READ_HEADER_TIMEOUT",
		"RADISHMIND_PLATFORM_WRITE_TIMEOUT",
		"RADISHMIND_PLATFORM_BRIDGE_TIMEOUT",
		"RADISHMIND_PLATFORM_BRIDGE_MODE",
		"RADISHMIND_PLATFORM_BRIDGE_WORKER_COUNT",
		"RADISHMIND_PLATFORM_BRIDGE_QUEUE_CAPACITY",
		"RADISHMIND_PLATFORM_BRIDGE_HANDSHAKE_TIMEOUT",
		"RADISHMIND_PLATFORM_PYTHON_BIN",
		"RADISHMIND_PLATFORM_BRIDGE_SCRIPT",
		"RADISHMIND_PLATFORM_PROVIDER",
		"RADISHMIND_PLATFORM_PROVIDER_PROFILE",
		"RADISHMIND_PLATFORM_MODEL",
		"RADISHMIND_PLATFORM_BASE_URL",
		"RADISHMIND_PLATFORM_API_KEY",
		"RADISHMIND_PLATFORM_TEMPERATURE",
		"RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH",
		"RADISHMIND_CONTROL_PLANE_READ_AUTH_MODE",
		"RADISHMIND_CONTROL_PLANE_READ_SIGNED_TEST_ISSUER",
		"RADISHMIND_CONTROL_PLANE_READ_SIGNED_TEST_AUDIENCE",
		"RADISHMIND_CONTROL_PLANE_READ_SIGNED_TEST_PUBLIC_KEY_PEM",
		"RADISHMIND_CONTROL_PLANE_READ_STORE",
		"RADISHMIND_CONTROL_PLANE_READ_DEV_TEST_DATABASE_URL",
		"RADISHMIND_CONTROL_PLANE_READ_DEV_TEST_MIGRATION_DATABASE_URL",
		"RADISHMIND_CONTROL_PLANE_READ_DATABASE_TIMEOUT",
		"RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP",
		"RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE",
		"RADISHMIND_WORKFLOW_EXECUTOR_DEV",
		"RADISHMIND_WORKFLOW_RAG_SNAPSHOT_DEV",
		"RADISHMIND_WORKFLOW_RAG_EXECUTION_DEV",
		"RADISHMIND_WORKFLOW_RAG_EVALUATION_DEV",
		"RADISHMIND_WORKFLOW_RAG_PROMOTION_DEV",
		"RADISHMIND_WORKFLOW_DIAGNOSTICS_DEV",
		"RADISHMIND_GATEWAY_REQUEST_HISTORY_DEV",
		"RADISHMIND_GATEWAY_REQUEST_STORE",
		"RADISHMIND_GATEWAY_REQUEST_DEV_TEST_DATABASE_URL",
		"RADISHMIND_GATEWAY_REQUEST_DEV_TEST_MIGRATION_DATABASE_URL",
		"RADISHMIND_GATEWAY_REQUEST_DATABASE_TIMEOUT",
		"RADISHMIND_LOCAL_PERSISTENCE_MODE",
		"RADISHMIND_SQLITE_DEV_DATABASE_PATH",
		"RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE",
		"RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_DATABASE_URL",
		"RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_MIGRATION_DATABASE_URL",
		"RADISHMIND_WORKFLOW_SAVED_DRAFT_DATABASE_TIMEOUT",
		"RADISHMIND_APPLICATION_DRAFT_DEV_HTTP",
		"RADISHMIND_APPLICATION_DRAFT_DEV_WRITE",
		"RADISHMIND_APPLICATION_DRAFT_STORE",
		"RADISHMIND_APPLICATION_DRAFT_DEV_TEST_DATABASE_URL",
		"RADISHMIND_APPLICATION_DRAFT_DEV_TEST_MIGRATION_DATABASE_URL",
		"RADISHMIND_APPLICATION_DRAFT_DATABASE_TIMEOUT",
		"RADISHMIND_APPLICATION_PUBLISH_DEV_HTTP",
		"RADISHMIND_APPLICATION_PUBLISH_DEV_WRITE",
		"RADISHMIND_APPLICATION_PUBLISH_STORE",
		"RADISHMIND_APPLICATION_PUBLISH_DEV_TEST_DATABASE_URL",
		"RADISHMIND_APPLICATION_PUBLISH_DEV_TEST_MIGRATION_DATABASE_URL",
		"RADISHMIND_APPLICATION_PUBLISH_DATABASE_TIMEOUT",
		"RADISHMIND_APPLICATION_CATALOG_DEV_HTTP",
		"RADISHMIND_APPLICATION_CATALOG_DEV_WRITE",
		"RADISHMIND_APPLICATION_CATALOG_STORE",
		"RADISHMIND_APPLICATION_CATALOG_DEV_TEST_DATABASE_URL",
		"RADISHMIND_APPLICATION_CATALOG_DEV_TEST_MIGRATION_DATABASE_URL",
		"RADISHMIND_APPLICATION_CATALOG_DATABASE_TIMEOUT",
		"RADISHMIND_API_KEY_LIFECYCLE_DEV_HTTP",
		"RADISHMIND_API_KEY_LIFECYCLE_DEV_WRITE",
		"RADISHMIND_API_KEY_STORE",
		"RADISHMIND_API_KEY_DEV_TEST_DATABASE_URL",
		"RADISHMIND_API_KEY_DEV_TEST_MIGRATION_DATABASE_URL",
		"RADISHMIND_API_KEY_DATABASE_TIMEOUT",
		"RADISHMIND_GATEWAY_AUTH_MODE",
		"RADISHMIND_WORKFLOW_RUN_STORE",
		"RADISHMIND_WORKFLOW_RUN_DEV_TEST_DATABASE_URL",
		"RADISHMIND_WORKFLOW_RUN_DEV_TEST_MIGRATION_DATABASE_URL",
		"RADISHMIND_WORKFLOW_RUN_DATABASE_TIMEOUT",
	} {
		t.Setenv(key, "")
	}
}
