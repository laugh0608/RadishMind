package bridge

import (
	"context"
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

func TestBridgeErrorCodesAreStable(t *testing.T) {
	if code := ErrorCode(context.DeadlineExceeded); code != ErrorCodeWorkerTimeout {
		t.Fatalf("unexpected timeout code: %s", code)
	}
	if code := ErrorCode(context.Canceled); code != ErrorCodeWorkerCanceled {
		t.Fatalf("unexpected cancellation code: %s", code)
	}
	err := newBridgeError(ErrorCodeWorkerExited, "bridge worker exited before completing request")
	if code := ErrorCode(err); code != ErrorCodeWorkerExited {
		t.Fatalf("unexpected bridge error code: %s", code)
	}
}

func TestBridgeModesFailClosed(t *testing.T) {
	if mode, err := ParseMode("stdio_pool"); err != nil || mode != ModeStdioPool {
		t.Fatalf("stdio pool mode was rejected: mode=%q err=%v", mode, err)
	}
	if _, err := ParseMode("unknown"); err == nil {
		t.Fatal("unknown bridge mode must fail closed")
	}
	if _, err := NewClientWithOptions("python3", defaultScriptPath, ClientOptions{
		Mode:        ModeStdioPool,
		WorkerCount: 33,
	}); err == nil {
		t.Fatal("unsafe bridge worker count must be rejected")
	}
}

func TestProcessBridgeFailureDoesNotExposeLocalPath(t *testing.T) {
	client := NewClient("python3", "missing/private/bridge-script.py")
	_, err := client.DescribeProviders(context.Background())
	if ErrorCode(err) != ErrorCodeProcessFailed {
		t.Fatalf("unexpected process bridge error: code=%q err=%v", ErrorCode(err), err)
	}
	if strings.Contains(err.Error(), "missing/private") {
		t.Fatalf("process bridge error exposed local path: %v", err)
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
