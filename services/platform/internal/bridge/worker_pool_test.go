package bridge

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const bridgeWorkerHelperEnvironment = "RADISHMIND_GO_BRIDGE_WORKER_HELPER"

type helperCanonicalRequest struct {
	RequestID string `json:"request_id"`
	Control   string `json:"control"`
	SleepMS   int    `json:"sleep_ms"`
}

func TestWorkerPoolConcurrentRequestsRemainIsolated(t *testing.T) {
	pool, _ := newTestWorkerPool(2, 2, "")
	client := &Client{mode: ModeStdioPool, pool: pool}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Start(ctx); err != nil {
		t.Fatalf("start worker pool: %v", err)
	}
	defer client.Close()

	requestIDs := []string{"canonical-001", "canonical-002", "canonical-003", "canonical-004"}
	start := make(chan struct{})
	errorsByRequest := make(chan error, len(requestIDs))
	var waitGroup sync.WaitGroup
	for _, requestID := range requestIDs {
		requestID := requestID
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			payload := helperRequestPayload(t, requestID, "", 20)
			envelope, err := client.HandleEnvelope(ctx, payload, testEnvelopeOptions())
			if err != nil {
				errorsByRequest <- err
				return
			}
			if envelope.RequestID != requestID || envelope.Response["summary"] != requestID {
				errorsByRequest <- newBridgeError(ErrorCodeWorkerProtocol, "concurrent response crossed request boundary")
			}
		}()
	}
	close(start)
	waitGroup.Wait()
	close(errorsByRequest)
	for err := range errorsByRequest {
		t.Fatalf("concurrent worker request failed: %v", err)
	}
}

func TestWorkerPoolStreamEventsRemainCorrelated(t *testing.T) {
	pool, _ := newTestWorkerPool(2, 2, "")
	client := &Client{mode: ModeStdioPool, pool: pool}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Start(ctx); err != nil {
		t.Fatalf("start worker pool: %v", err)
	}
	defer client.Close()

	requestIDs := []string{"stream-001", "stream-002"}
	var waitGroup sync.WaitGroup
	errorsByRequest := make(chan error, len(requestIDs))
	for _, requestID := range requestIDs {
		requestID := requestID
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			events := make([]StreamEvent, 0, 2)
			err := client.StreamEnvelope(
				ctx,
				helperRequestPayload(t, requestID, "", 20),
				testEnvelopeOptions(),
				func(event StreamEvent) error {
					events = append(events, event)
					return nil
				},
			)
			if err != nil {
				errorsByRequest <- err
				return
			}
			if len(events) != 2 || events[0].Delta != requestID || events[1].Type != "completed" ||
				events[1].Envelope == nil || events[1].Envelope.RequestID != requestID {
				errorsByRequest <- newBridgeError(ErrorCodeWorkerProtocol, "stream event crossed request boundary")
			}
		}()
	}
	waitGroup.Wait()
	close(errorsByRequest)
	for err := range errorsByRequest {
		t.Fatalf("concurrent stream failed: %v", err)
	}
}

func TestWorkerPoolSupportsProviderMetadataOperations(t *testing.T) {
	pool, _ := newTestWorkerPool(1, 1, "")
	client := &Client{mode: ModeStdioPool, pool: pool}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.Start(ctx); err != nil {
		t.Fatalf("start worker pool: %v", err)
	}
	defer client.Close()

	providers, err := client.DescribeProviders(ctx)
	if err != nil || len(providers) != 1 || providers[0].ProviderID != "mock" {
		t.Fatalf("unexpected provider registry: providers=%#v err=%v", providers, err)
	}
	inventory, err := client.DescribeInventory(ctx)
	if err != nil || len(inventory.Providers) != 1 || inventory.Providers[0].ProviderID != "mock" {
		t.Fatalf("unexpected provider inventory: inventory=%#v err=%v", inventory, err)
	}
}

func TestWorkerPoolQueueIsBounded(t *testing.T) {
	pool, _ := newTestWorkerPool(1, 1, "")
	client := &Client{mode: ModeStdioPool, pool: pool}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Start(ctx); err != nil {
		t.Fatalf("start worker pool: %v", err)
	}
	defer client.Close()

	requestDone := make(chan error, 2)
	go func() {
		_, err := client.HandleEnvelope(ctx, helperRequestPayload(t, "queue-001", "", 250), testEnvelopeOptions())
		requestDone <- err
	}()
	waitForWorkerPoolState(t, pool, 1, 0)
	go func() {
		_, err := client.HandleEnvelope(ctx, helperRequestPayload(t, "queue-002", "", 10), testEnvelopeOptions())
		requestDone <- err
	}()
	waitForWorkerPoolState(t, pool, 2, 0)

	_, err := client.HandleEnvelope(ctx, helperRequestPayload(t, "queue-003", "", 0), testEnvelopeOptions())
	if ErrorCode(err) != ErrorCodeWorkerQueueFull {
		t.Fatalf("expected bounded queue failure, got code=%q err=%v", ErrorCode(err), err)
	}
	for range 2 {
		if err := <-requestDone; err != nil {
			t.Fatalf("admitted queue request failed: %v", err)
		}
	}
}

func TestWorkerPoolCancellationRebuildsBeforeLaterRequest(t *testing.T) {
	pool, commandStarts := newTestWorkerPool(1, 1, "")
	client := &Client{mode: ModeStdioPool, pool: pool}
	startContext, startCancel := context.WithTimeout(context.Background(), time.Second)
	defer startCancel()
	if err := client.Start(startContext); err != nil {
		t.Fatalf("start worker pool: %v", err)
	}
	defer client.Close()

	requestContext, requestCancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer requestCancel()
	_, err := client.HandleEnvelope(
		requestContext,
		helperRequestPayload(t, "timeout-001", "", 500),
		testEnvelopeOptions(),
	)
	if ErrorCode(err) != ErrorCodeWorkerTimeout {
		t.Fatalf("expected worker timeout, got code=%q err=%v", ErrorCode(err), err)
	}

	recoveryContext, recoveryCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer recoveryCancel()
	envelope, err := client.HandleEnvelope(
		recoveryContext,
		helperRequestPayload(t, "timeout-recovery-001", "", 0),
		testEnvelopeOptions(),
	)
	if err != nil || envelope.RequestID != "timeout-recovery-001" {
		t.Fatalf("worker did not recover after timeout: envelope=%#v err=%v", envelope, err)
	}
	if commandStarts.Load() < 2 {
		t.Fatalf("expected a replacement worker start, got %d", commandStarts.Load())
	}
}

func TestWorkerPoolCrashFailsCurrentRequestWithoutRetryAndRecovers(t *testing.T) {
	pool, commandStarts := newTestWorkerPool(1, 1, "")
	client := &Client{mode: ModeStdioPool, pool: pool}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Start(ctx); err != nil {
		t.Fatalf("start worker pool: %v", err)
	}
	defer client.Close()

	_, err := client.HandleEnvelope(ctx, helperRequestPayload(t, "crash-001", "crash", 0), testEnvelopeOptions())
	if ErrorCode(err) != ErrorCodeWorkerExited {
		t.Fatalf("expected current request to fail on crash, got code=%q err=%v", ErrorCode(err), err)
	}
	envelope, err := client.HandleEnvelope(
		ctx,
		helperRequestPayload(t, "crash-recovery-001", "", 0),
		testEnvelopeOptions(),
	)
	if err != nil || envelope.RequestID != "crash-recovery-001" {
		t.Fatalf("worker did not recover after crash: envelope=%#v err=%v", envelope, err)
	}
	if commandStarts.Load() != 2 {
		t.Fatalf("crashed request must not be retried; expected two total worker starts, got %d", commandStarts.Load())
	}
}

func TestWorkerPoolRejectsIncompatibleHandshake(t *testing.T) {
	pool, _ := newTestWorkerPool(1, 1, "bad_handshake")
	defer pool.close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := pool.ensureCapacity(ctx)
	if ErrorCode(err) != ErrorCodeWorkerProtocol {
		t.Fatalf("expected protocol error, got code=%q err=%v", ErrorCode(err), err)
	}
}

func TestWorkerPoolCloseReapsAllWorkers(t *testing.T) {
	pool, _ := newTestWorkerPool(2, 1, "")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := pool.ensureCapacity(ctx); err != nil {
		t.Fatalf("start worker pool: %v", err)
	}

	pool.stateMu.Lock()
	workers := make([]*workerProcess, 0, len(pool.workers))
	for worker := range pool.workers {
		workers = append(workers, worker)
	}
	pool.stateMu.Unlock()
	pool.close()

	for _, worker := range workers {
		select {
		case <-worker.waitDone:
		case <-time.After(time.Second):
			t.Fatal("worker process remained after pool close")
		}
	}
	_, err := pool.execute(ctx, "providers", nil, EnvelopeOptions{}, nil)
	if ErrorCode(err) != ErrorCodeClientClosed {
		t.Fatalf("closed pool accepted a request: code=%q err=%v", ErrorCode(err), err)
	}
}

func TestWorkerPoolCloseInterruptsActiveRequestWithinBudget(t *testing.T) {
	pool, _ := newTestWorkerPool(1, 1, "")
	pool.config.shutdownTimeout = 50 * time.Millisecond
	client := &Client{mode: ModeStdioPool, pool: pool}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Start(ctx); err != nil {
		t.Fatalf("start worker pool: %v", err)
	}

	requestDone := make(chan error, 1)
	go func() {
		_, err := client.HandleEnvelope(
			ctx,
			helperRequestPayload(t, "close-active-001", "", 5000),
			testEnvelopeOptions(),
		)
		requestDone <- err
	}()
	waitForWorkerPoolState(t, pool, 1, 0)

	startedAt := time.Now()
	client.Close()
	if elapsed := time.Since(startedAt); elapsed > 500*time.Millisecond {
		t.Fatalf("active worker close exceeded shutdown budget: %s", elapsed)
	}
	if err := <-requestDone; ErrorCode(err) != ErrorCodeClientClosed {
		t.Fatalf("active request did not observe client close: code=%q err=%v", ErrorCode(err), err)
	}
}

func TestBridgeWorkerEnvironmentClearsInheritedCredential(t *testing.T) {
	t.Setenv(bridgeAPIKeyEnvironmentVariable, "stale-worker-secret")
	if value, found := bridgeEnvironmentValue(bridgeWorkerEnvironment(), bridgeAPIKeyEnvironmentVariable); found {
		t.Fatalf("persistent worker inherited stale credential %q", value)
	}
}

func TestWorkerProcessWriteStopsOnContextCancellation(t *testing.T) {
	writer := &blockingWriteCloser{closed: make(chan struct{})}
	waitDone := make(chan struct{})
	close(waitDone)
	worker := &workerProcess{
		command:  &exec.Cmd{},
		rawStdin: writer,
		waitDone: waitDone,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := worker.writeRequestWithContext(ctx, []byte("request-secret-token"))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected write timeout, got %v", err)
	}
	if !worker.closed.Load() {
		t.Fatal("timed out writer did not terminate its worker")
	}
}

type blockingWriteCloser struct {
	closed    chan struct{}
	closeOnce sync.Once
}

func (w *blockingWriteCloser) Write(_ []byte) (int, error) {
	<-w.closed
	return 0, io.ErrClosedPipe
}

func (w *blockingWriteCloser) Close() error {
	w.closeOnce.Do(func() {
		close(w.closed)
	})
	return nil
}

func newTestWorkerPool(workerCount int, queueCapacity int, helperMode string) (*workerPool, *atomic.Int64) {
	commandStarts := &atomic.Int64{}
	pool := newWorkerPool(workerPoolConfig{
		workerCount:      workerCount,
		queueCapacity:    queueCapacity,
		handshakeTimeout: time.Second,
		shutdownTimeout:  time.Second,
		commandFactory: func() (*exec.Cmd, error) {
			commandStarts.Add(1)
			command := exec.Command(os.Args[0], "-test.run=TestBridgeWorkerHelperProcess", "--")
			command.Env = append(
				os.Environ(),
				bridgeWorkerHelperEnvironment+"=1",
				"RADISHMIND_GO_BRIDGE_WORKER_HELPER_MODE="+helperMode,
			)
			return command, nil
		},
	})
	return pool, commandStarts
}

func waitForWorkerPoolState(t *testing.T, pool *workerPool, admissionCount int, idleCount int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if len(pool.admission) == admissionCount && len(pool.idle) == idleCount {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf(
		"worker pool state was not reached: admission=%d/%d idle=%d/%d",
		len(pool.admission),
		admissionCount,
		len(pool.idle),
		idleCount,
	)
}

func helperRequestPayload(t *testing.T, requestID string, control string, sleepMS int) []byte {
	t.Helper()
	payload, err := json.Marshal(helperCanonicalRequest{
		RequestID: requestID,
		Control:   control,
		SleepMS:   sleepMS,
	})
	if err != nil {
		t.Fatalf("encode helper request: %v", err)
	}
	return payload
}

func testEnvelopeOptions() EnvelopeOptions {
	return EnvelopeOptions{
		Provider:       "mock",
		RequestTimeout: time.Second,
	}
}

func TestBridgeWorkerHelperProcess(t *testing.T) {
	if os.Getenv(bridgeWorkerHelperEnvironment) != "1" {
		return
	}
	encoder := json.NewEncoder(os.Stdout)
	if os.Getenv("RADISHMIND_GO_BRIDGE_WORKER_HELPER_MODE") == "bad_handshake" {
		_ = encoder.Encode(workerResponseFrame{
			ProtocolVersion: workerProtocolVersion + 1,
			Type:            "ready",
			Operations:      []string{"envelope", "stream", "providers", "inventory"},
		})
		os.Exit(0)
	}
	_ = encoder.Encode(workerResponseFrame{
		ProtocolVersion: workerProtocolVersion,
		Type:            "ready",
		Operations:      []string{"envelope", "stream", "providers", "inventory"},
	})

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 64*1024), workerMaximumFrameSize)
	for scanner.Scan() {
		var request workerRequestFrame
		if err := json.Unmarshal(scanner.Bytes(), &request); err != nil {
			os.Exit(3)
		}
		var canonical helperCanonicalRequest
		if len(request.Request) > 0 {
			_ = json.Unmarshal(request.Request, &canonical)
		}
		if canonical.Control == "crash" {
			os.Exit(7)
		}
		if canonical.SleepMS > 0 {
			time.Sleep(time.Duration(canonical.SleepMS) * time.Millisecond)
		}
		switch request.Operation {
		case "envelope":
			writeHelperResult(encoder, request.RequestID, helperGatewayEnvelope(canonical.RequestID))
		case "stream":
			_ = encoder.Encode(workerResponseFrame{
				ProtocolVersion: workerProtocolVersion,
				Type:            "stream_event",
				RequestID:       request.RequestID,
				Event:           mustMarshalHelper(StreamEvent{Type: "delta", Delta: canonical.RequestID}),
			})
			completedEnvelope := helperGatewayEnvelope(canonical.RequestID)
			_ = encoder.Encode(workerResponseFrame{
				ProtocolVersion: workerProtocolVersion,
				Type:            "stream_event",
				RequestID:       request.RequestID,
				Event: mustMarshalHelper(StreamEvent{
					Type:     "completed",
					Envelope: &completedEnvelope,
				}),
			})
			writeHelperResult(encoder, request.RequestID, nil)
		case "providers":
			writeHelperResult(encoder, request.RequestID, []ProviderDescription{{ProviderID: "mock"}})
		case "inventory":
			writeHelperResult(encoder, request.RequestID, ProviderInventory{
				Providers:          []ProviderDescription{{ProviderID: "mock"}},
				Profiles:           []ProviderProfileDescription{},
				ActiveProfileChain: []string{},
			})
		default:
			os.Exit(4)
		}
	}
	os.Exit(0)
}

func helperGatewayEnvelope(requestID string) GatewayEnvelope {
	return GatewayEnvelope{
		SchemaVersion: 1,
		Status:        "ok",
		RequestID:     requestID,
		Project:       "radish",
		Task:          "answer_docs_question",
		Response: map[string]any{
			"summary": requestID,
		},
		Metadata: map[string]any{
			"duration_ms":          1,
			"provider_duration_ms": 0,
		},
	}
}

func writeHelperResult(encoder *json.Encoder, requestID string, payload any) {
	_ = encoder.Encode(workerResponseFrame{
		ProtocolVersion: workerProtocolVersion,
		Type:            "result",
		RequestID:       requestID,
		Payload:         mustMarshalHelper(payload),
	})
}

func mustMarshalHelper(value any) json.RawMessage {
	payload, err := json.Marshal(value)
	if err != nil {
		os.Exit(5)
	}
	return payload
}
