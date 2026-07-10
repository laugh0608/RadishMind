package bridge

import (
	"strings"
	"testing"
	"time"
)

func TestBridgeCredentialTransport(t *testing.T) {
	const apiKey = "provider-secret-token"
	client := NewClient("python3", defaultScriptPath)
	options := EnvelopeOptions{
		Provider:        "openai-compatible",
		ProviderProfile: "default",
		Model:           "test-model",
		BaseURL:         "https://provider.invalid/v1",
		APIKey:          apiKey,
		Temperature:     0.2,
		RequestTimeout:  5 * time.Second,
	}

	args := client.buildEnvelopeArgs("envelope", options)
	joinedArgs := strings.Join(args, "\x00")
	if strings.Contains(joinedArgs, "--api-key") || strings.Contains(joinedArgs, apiKey) {
		t.Fatalf("bridge command arguments must not expose credentials: %q", args)
	}

	environment := bridgeCommandEnvironment(apiKey)
	if value, found := bridgeEnvironmentValue(environment, bridgeAPIKeyEnvironmentVariable); !found || value != apiKey {
		t.Fatalf("bridge credential was not scoped to the child environment: found=%v value=%q", found, value)
	}
}

func TestBridgeCredentialEnvironmentDoesNotLeakAcrossRequests(t *testing.T) {
	t.Setenv(bridgeAPIKeyEnvironmentVariable, "inherited-secret")

	environment := bridgeCommandEnvironment("")
	if value, found := bridgeEnvironmentValue(environment, bridgeAPIKeyEnvironmentVariable); found {
		t.Fatalf("request without a credential inherited stale bridge secret %q", value)
	}
}

func bridgeEnvironmentValue(environment []string, key string) (string, bool) {
	prefix := key + "="
	for _, entry := range environment {
		if strings.HasPrefix(entry, prefix) {
			return strings.TrimPrefix(entry, prefix), true
		}
	}
	return "", false
}
