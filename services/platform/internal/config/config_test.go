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
		ListenAddr:        "127.0.0.1:7000",
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		BridgeTimeout:     45 * time.Second,
		PythonBinary:      "python3",
		BridgeScript:      "scripts/run-platform-bridge.py",
		Provider:          "openai-compatible",
		ProviderProfile:   "anyrouter",
		Model:             "deepseek-chat",
		BaseURL:           "https://example.invalid/v1",
		APIKey:            "secret-token",
		Temperature:       0.2,
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
	if summary.WorkflowSavedDraftStoreMode != "memory_dev" {
		t.Fatalf("unexpected workflow saved draft store mode: %s", summary.WorkflowSavedDraftStoreMode)
	}
	if !reflect.DeepEqual(summary.MissingRequiredFields, []string{}) {
		t.Fatalf("unexpected missing required fields: %#v", summary.MissingRequiredFields)
	}
	if !reflect.DeepEqual(summary.SecretFields, []string{"RADISHMIND_PLATFORM_API_KEY"}) {
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
	t.Setenv("RADISHMIND_PLATFORM_API_KEY", "env-secret")
	t.Setenv("RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH", "1")
	t.Setenv("RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP", "1")
	t.Setenv("RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE", "true")
	t.Setenv("RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE", "repository_disabled")

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
	if cfg.WorkflowSavedDraftStoreMode != "repository_disabled" {
		t.Fatalf("expected workflow saved draft store env override, got %s", cfg.WorkflowSavedDraftStoreMode)
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
	if summary.FieldSources["workflow_saved_draft_store"] != "env" {
		t.Fatalf("expected workflow_saved_draft_store source=env, got %#v", summary.FieldSources)
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

func clearPlatformEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"RADISHMIND_PLATFORM_CONFIG",
		"RADISHMIND_PLATFORM_LISTEN_ADDR",
		"RADISHMIND_PLATFORM_READ_HEADER_TIMEOUT",
		"RADISHMIND_PLATFORM_WRITE_TIMEOUT",
		"RADISHMIND_PLATFORM_BRIDGE_TIMEOUT",
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
		"RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE",
	} {
		t.Setenv(key, "")
	}
}
