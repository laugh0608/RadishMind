package bridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const (
	bridgeProcessHelperEnvironment     = "RADISHMIND_GO_BRIDGE_PROCESS_HELPER"
	bridgeProcessHelperModeEnvironment = "RADISHMIND_GO_BRIDGE_PROCESS_HELPER_MODE"
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

func TestBridgeClientDefaultsOptionsAndLifecycle(t *testing.T) {
	var nilClient *Client
	if nilClient.Mode() != ModeProcessPerRequest || nilClient.ProcessStarts() != 0 {
		t.Fatal("nil bridge client must expose safe process defaults")
	}
	if err := nilClient.Start(context.Background()); err != nil {
		t.Fatalf("nil bridge client start failed: %v", err)
	}
	nilClient.Close()

	client := NewClient("", "")
	if client.pythonBinary != defaultPythonBinary || client.scriptPath != defaultScriptPath || client.Mode() != ModeProcessPerRequest {
		t.Fatalf("unexpected bridge defaults: %#v", client)
	}
	if err := client.Start(context.Background()); err != nil {
		t.Fatalf("process bridge start must be a no-op: %v", err)
	}
	client.Close()

	processClient, err := NewClientWithOptions("python3", defaultScriptPath, ClientOptions{})
	if err != nil || processClient.Mode() != ModeProcessPerRequest || processClient.pool != nil {
		t.Fatalf("unexpected normalized process client: mode=%q pool_present=%t err=%v", processClient.Mode(), processClient.pool != nil, err)
	}
	poolClient, err := NewClientWithOptions("python3", defaultScriptPath, ClientOptions{
		Mode: ModeStdioPool, WorkerCount: 1, QueueCapacity: 1, HandshakeTimeout: time.Second,
	})
	if err != nil || poolClient.Mode() != ModeStdioPool || poolClient.pool == nil {
		t.Fatalf("unexpected stdio pool client: mode=%q pool_present=%t err=%v", poolClient.Mode(), poolClient.pool != nil, err)
	}
	poolClient.Close()

	invalidOptions := []ClientOptions{
		{Mode: ModeStdioPool, WorkerCount: -1},
		{Mode: ModeStdioPool, QueueCapacity: -1},
		{Mode: ModeStdioPool, QueueCapacity: maximumQueueCapacity + 1},
		{Mode: ModeStdioPool, HandshakeTimeout: -time.Second},
	}
	for _, options := range invalidOptions {
		if _, err := NewClientWithOptions("python3", defaultScriptPath, options); err == nil {
			t.Fatalf("unsafe bridge options were accepted: %#v", options)
		}
	}
}

func TestProcessBridgeSupportsMetadataEnvelopeAndStream(t *testing.T) {
	client := newTestProcessClient(t, "")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	providers, err := client.DescribeProviders(ctx)
	if err != nil || len(providers) != 1 || providers[0].ProviderID != "mock" {
		t.Fatalf("unexpected process provider registry: providers=%#v err=%v", providers, err)
	}
	inventory, err := client.DescribeInventory(ctx)
	if err != nil || len(inventory.Providers) != 1 || inventory.Providers[0].ProviderID != "mock" {
		t.Fatalf("unexpected process inventory: inventory=%#v err=%v", inventory, err)
	}

	request := processHelperRequest{RequestID: "process-envelope-001"}
	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("encode process helper request: %v", err)
	}
	envelope, err := client.HandleEnvelope(ctx, payload, EnvelopeOptions{Provider: "mock", APIKey: "temporary-provider-secret"})
	if err != nil || envelope.RequestID != request.RequestID || envelope.Metadata["credential_present"] != true {
		t.Fatalf("unexpected process envelope: envelope=%#v err=%v", envelope, err)
	}

	events := make([]StreamEvent, 0, 2)
	err = client.StreamEnvelope(ctx, payload, EnvelopeOptions{Provider: "mock"}, func(event StreamEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil || len(events) != 2 || events[0].Delta != request.RequestID || events[1].Envelope == nil || events[1].Envelope.RequestID != request.RequestID {
		t.Fatalf("unexpected process stream: events=%#v err=%v", events, err)
	}
	if client.ProcessStarts() != 4 {
		t.Fatalf("unexpected process start count: %d", client.ProcessStarts())
	}
}

func TestProcessBridgeCancellationAndConsumerStopAreStable(t *testing.T) {
	client := newTestProcessClient(t, "")
	payload, err := json.Marshal(processHelperRequest{RequestID: "process-timeout-001", SleepMS: 500})
	if err != nil {
		t.Fatalf("encode timeout helper request: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := client.HandleEnvelope(ctx, payload, EnvelopeOptions{Provider: "mock"}); ErrorCode(err) != ErrorCodeWorkerTimeout {
		t.Fatalf("unexpected process timeout: code=%q err=%v", ErrorCode(err), err)
	}

	streamPayload, err := json.Marshal(processHelperRequest{RequestID: "process-consumer-stop-001"})
	if err != nil {
		t.Fatalf("encode stream helper request: %v", err)
	}
	err = client.StreamEnvelope(context.Background(), streamPayload, EnvelopeOptions{Provider: "mock"}, func(StreamEvent) error {
		return errors.New("stop without exposing local detail")
	})
	if ErrorCode(err) != ErrorCodeWorkerCanceled || strings.Contains(err.Error(), "local detail") {
		t.Fatalf("unexpected process consumer stop: code=%q err=%v", ErrorCode(err), err)
	}
}

func TestProcessBridgeFailuresRemainSanitized(t *testing.T) {
	t.Run("process failure", func(t *testing.T) {
		client := newTestProcessClient(t, "failure")
		_, err := client.DescribeProviders(context.Background())
		if ErrorCode(err) != ErrorCodeProcessFailed || strings.Contains(err.Error(), "provider-secret") {
			t.Fatalf("process failure was not sanitized: code=%q err=%v", ErrorCode(err), err)
		}
	})
	t.Run("invalid registry", func(t *testing.T) {
		client := newTestProcessClient(t, "invalid_json")
		_, err := client.DescribeProviders(context.Background())
		if err == nil || ErrorCode(err) != "" || !strings.Contains(err.Error(), "decode provider registry") {
			t.Fatalf("invalid provider registry was not rejected: code=%q err=%v", ErrorCode(err), err)
		}
	})
	t.Run("invalid stream", func(t *testing.T) {
		client := newTestProcessClient(t, "invalid_json")
		err := client.StreamEnvelope(context.Background(), nil, EnvelopeOptions{Provider: "mock"}, nil)
		if err == nil || ErrorCode(err) != "" || !strings.Contains(err.Error(), "decode bridge stream event") {
			t.Fatalf("invalid stream event was not rejected: code=%q err=%v", ErrorCode(err), err)
		}
	})
}

type processHelperRequest struct {
	RequestID string `json:"request_id"`
	SleepMS   int    `json:"sleep_ms"`
}

func newTestProcessClient(t *testing.T, mode string) *Client {
	t.Helper()
	t.Setenv(bridgeProcessHelperEnvironment, "1")
	t.Setenv(bridgeProcessHelperModeEnvironment, mode)
	client := NewClient("test-process-helper", os.Args[0])
	client.processCommand = func(ctx context.Context, _ string, args ...string) *exec.Cmd {
		helperArgs := append([]string{"-test.run=TestBridgeProcessHelperProcess", "--"}, args...)
		return exec.CommandContext(ctx, os.Args[0], helperArgs...)
	}
	return client
}

func TestBridgeProcessHelperProcess(t *testing.T) {
	if os.Getenv(bridgeProcessHelperEnvironment) != "1" {
		return
	}
	args := os.Args
	separator := -1
	for index, argument := range args {
		if argument == "--" {
			separator = index
			break
		}
	}
	if separator < 0 || len(args) <= separator+2 {
		os.Exit(2)
	}
	operation := args[separator+2]
	mode := os.Getenv(bridgeProcessHelperModeEnvironment)
	if mode == "failure" {
		_, _ = fmt.Fprintln(os.Stderr, "provider-secret must stay private")
		os.Exit(7)
	}
	if mode == "invalid_json" {
		_, _ = fmt.Fprintln(os.Stdout, "{")
		os.Exit(0)
	}

	var request processHelperRequest
	_ = json.NewDecoder(os.Stdin).Decode(&request)
	if request.SleepMS > 0 {
		time.Sleep(time.Duration(request.SleepMS) * time.Millisecond)
	}
	encoder := json.NewEncoder(os.Stdout)
	switch operation {
	case "providers":
		_ = encoder.Encode([]ProviderDescription{{ProviderID: "mock"}})
	case "inventory":
		_ = encoder.Encode(ProviderInventory{Providers: []ProviderDescription{{ProviderID: "mock"}}})
	case "envelope":
		_ = encoder.Encode(processHelperEnvelope(request.RequestID))
	case "stream":
		_ = encoder.Encode(StreamEvent{Type: "delta", Delta: request.RequestID})
		envelope := processHelperEnvelope(request.RequestID)
		_ = encoder.Encode(StreamEvent{Type: "completed", Envelope: &envelope})
	default:
		os.Exit(3)
	}
	os.Exit(0)
}

func processHelperEnvelope(requestID string) GatewayEnvelope {
	_, credentialPresent := os.LookupEnv(bridgeAPIKeyEnvironmentVariable)
	return GatewayEnvelope{
		SchemaVersion: 1,
		Status:        "ok",
		RequestID:     requestID,
		Project:       "radish",
		Task:          "answer_docs_question",
		Response:      map[string]any{"summary": requestID},
		Metadata:      map[string]any{"credential_present": credentialPresent},
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
