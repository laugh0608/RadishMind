package config

import (
	"os"
	"path/filepath"
	"reflect"
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
		"RADISHMIND_WORKFLOW_RUN_DEV_TEST_DATABASE_URL",
		"RADISHMIND_WORKFLOW_RUN_DEV_TEST_MIGRATION_DATABASE_URL",
	}) {
		t.Fatalf("unexpected secret fields: %#v", summary.SecretFields)
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
	t.Setenv("RADISHMIND_WORKFLOW_EXECUTOR_DEV", "1")
	t.Setenv("RADISHMIND_GATEWAY_REQUEST_HISTORY_DEV", "1")
	t.Setenv("RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE", "repository_disabled")
	t.Setenv("RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_DATABASE_URL", "postgresql://runtime.invalid/secret")
	t.Setenv("RADISHMIND_WORKFLOW_SAVED_DRAFT_DATABASE_TIMEOUT", "9s")

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
	if !cfg.WorkflowExecutorDevEnabled {
		t.Fatalf("expected workflow executor dev env override")
	}
	if !cfg.GatewayRequestHistoryDevEnabled {
		t.Fatalf("expected gateway request history dev env override")
	}
	if cfg.WorkflowSavedDraftStoreMode != "repository_disabled" {
		t.Fatalf("expected workflow saved draft store env override, got %s", cfg.WorkflowSavedDraftStoreMode)
	}
	if cfg.WorkflowSavedDraftDatabaseURL == "" || cfg.WorkflowSavedDraftDatabaseTimeout != 9*time.Second {
		t.Fatalf("expected workflow saved draft database env overrides")
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
	if summary.ConfigFile.Path != configPath || !summary.ConfigFile.Configured || !summary.ConfigFile.Loaded {
		t.Fatalf("unexpected config file summary: %#v", summary.ConfigFile)
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
		"RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP",
		"RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE",
		"RADISHMIND_WORKFLOW_EXECUTOR_DEV",
		"RADISHMIND_WORKFLOW_DIAGNOSTICS_DEV",
		"RADISHMIND_GATEWAY_REQUEST_HISTORY_DEV",
		"RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE",
		"RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_DATABASE_URL",
		"RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_MIGRATION_DATABASE_URL",
		"RADISHMIND_WORKFLOW_SAVED_DRAFT_DATABASE_TIMEOUT",
		"RADISHMIND_WORKFLOW_RUN_STORE",
		"RADISHMIND_WORKFLOW_RUN_DEV_TEST_DATABASE_URL",
		"RADISHMIND_WORKFLOW_RUN_DEV_TEST_MIGRATION_DATABASE_URL",
		"RADISHMIND_WORKFLOW_RUN_DATABASE_TIMEOUT",
	} {
		t.Setenv(key, "")
	}
}
