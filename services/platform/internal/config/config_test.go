package config

import (
	"reflect"
	"testing"
	"time"
)

func TestSanitizedSummaryDoesNotExposeSecrets(t *testing.T) {
	cfg := Config{
		ListenAddr:        "127.0.0.1:8080",
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
	if !reflect.DeepEqual(summary.MissingRequiredFields, []string{}) {
		t.Fatalf("unexpected missing required fields: %#v", summary.MissingRequiredFields)
	}
	if !reflect.DeepEqual(summary.SecretFields, []string{"RADISHMIND_PLATFORM_API_KEY"}) {
		t.Fatalf("unexpected secret fields: %#v", summary.SecretFields)
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
				ListenAddr: ":8080",
				Provider:   "mock",
			},
			expectedState:   "not_required",
			expectedMissing: []string{},
		},
		{
			name: "remote provider requires model base url and credential",
			config: Config{
				ListenAddr: ":8080",
				Provider:   "huggingface",
			},
			expectedState:   "missing",
			expectedMissing: []string{"model", "base_url", "credential"},
		},
		{
			name: "ollama credential is optional but model and base url are required",
			config: Config{
				ListenAddr: ":8080",
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
