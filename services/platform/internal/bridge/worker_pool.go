package bridge

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	workerProtocolVersion  = 1
	workerMaximumFrameSize = 8 * 1024 * 1024
)

type workerCommandFactory func() (*exec.Cmd, error)

type workerPoolConfig struct {
	workerCount      int
	queueCapacity    int
	handshakeTimeout time.Duration
	shutdownTimeout  time.Duration
	commandFactory   workerCommandFactory
}

type workerPool struct {
	config workerPoolConfig

	idle      chan *workerProcess
	admission chan struct{}
	closed    chan struct{}
	closeOnce sync.Once

	startMu sync.Mutex
	stateMu sync.Mutex
	workers map[*workerProcess]struct{}
	nextID  atomic.Uint64
	starts  atomic.Int64
}

type workerProcess struct {
	command  *exec.Cmd
	rawStdin io.WriteCloser
	scanner  *bufio.Scanner
	waitDone chan struct{}

	mu            sync.Mutex
	closed        atomic.Bool
	terminateOnce sync.Once
}

type workerRequestFrame struct {
	ProtocolVersion int                   `json:"protocol_version"`
	Type            string                `json:"type"`
	RequestID       string                `json:"request_id"`
	Operation       string                `json:"operation"`
	Request         json.RawMessage       `json:"request,omitempty"`
	Options         *workerRequestOptions `json:"options,omitempty"`
}

type workerRequestOptions struct {
	Provider              string  `json:"provider"`
	ProviderProfile       string  `json:"provider_profile"`
	Model                 string  `json:"model"`
	BaseURL               string  `json:"base_url"`
	APIKey                string  `json:"api_key,omitempty"`
	Temperature           float64 `json:"temperature"`
	RequestTimeoutSeconds float64 `json:"request_timeout_seconds"`
}

type workerResponseFrame struct {
	ProtocolVersion int               `json:"protocol_version"`
	Type            string            `json:"type"`
	RequestID       string            `json:"request_id"`
	Operations      []string          `json:"operations"`
	Payload         json.RawMessage   `json:"payload"`
	Event           json.RawMessage   `json:"event"`
	Error           *workerFrameError `json:"error"`
}

type workerFrameError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type workerScanResult struct {
	line string
	err  error
}

func newWorkerPool(config workerPoolConfig) *workerPool {
	return &workerPool{
		config:    config,
		idle:      make(chan *workerProcess, config.workerCount),
		admission: make(chan struct{}, config.workerCount+config.queueCapacity),
		closed:    make(chan struct{}),
		workers:   make(map[*workerProcess]struct{}, config.workerCount),
	}
}

func newPythonWorkerCommandFactory(pythonBinary string, scriptPath string) workerCommandFactory {
	return func() (*exec.Cmd, error) {
		resolvedScriptPath, err := resolveScriptPath(scriptPath)
		if err != nil {
			return nil, newBridgeError(ErrorCodeWorkerUnavailable, "bridge worker script is unavailable")
		}
		command := exec.Command(pythonBinary, resolvedScriptPath, "worker")
		command.Env = bridgeWorkerEnvironment()
		return command, nil
	}
}

func bridgeWorkerEnvironment() []string {
	return bridgeCommandEnvironment("")
}

func (p *workerPool) ensureCapacity(ctx context.Context) error {
	if p == nil {
		return newBridgeError(ErrorCodeWorkerUnavailable, "bridge worker pool is unavailable")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if p.isClosed() {
		return newBridgeError(ErrorCodeClientClosed, "bridge client is closed")
	}
	if ctx.Err() != nil {
		return bridgeContextError(ctx.Err())
	}

	p.startMu.Lock()
	defer p.startMu.Unlock()
	for {
		if p.isClosed() {
			return newBridgeError(ErrorCodeClientClosed, "bridge client is closed")
		}
		p.stateMu.Lock()
		missingWorkers := p.config.workerCount - len(p.workers)
		p.stateMu.Unlock()
		if missingWorkers <= 0 {
			return nil
		}

		worker, err := startWorkerProcess(ctx, p.config)
		if err != nil {
			return err
		}
		p.starts.Add(1)
		p.stateMu.Lock()
		if p.isClosed() {
			p.stateMu.Unlock()
			worker.stop(true, p.config.shutdownTimeout)
			return newBridgeError(ErrorCodeClientClosed, "bridge client is closed")
		}
		p.workers[worker] = struct{}{}
		p.stateMu.Unlock()

		select {
		case p.idle <- worker:
		case <-p.closed:
			p.removeWorker(worker)
			worker.stop(true, p.config.shutdownTimeout)
			return newBridgeError(ErrorCodeClientClosed, "bridge client is closed")
		case <-ctx.Done():
			p.removeWorker(worker)
			worker.stop(true, p.config.shutdownTimeout)
			return bridgeContextError(ctx.Err())
		}
	}
}

func (p *workerPool) execute(
	ctx context.Context,
	operation string,
	canonicalRequest []byte,
	options EnvelopeOptions,
	handleEvent func(StreamEvent) error,
) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := p.ensureCapacity(ctx); err != nil {
		return nil, err
	}
	if err := p.acquireAdmission(ctx); err != nil {
		return nil, err
	}

	worker, err := p.acquireWorker(ctx)
	if err != nil {
		p.releaseAdmission()
		return nil, err
	}
	request := p.buildRequest(operation, canonicalRequest, options)
	payload, healthy, requestErr := worker.execute(ctx, request, handleEvent)
	if healthy {
		p.returnWorker(worker)
	} else {
		p.removeWorker(worker)
		worker.stop(false, p.config.shutdownTimeout)
		go p.replenish()
	}
	p.releaseAdmission()
	return payload, requestErr
}

func (p *workerPool) buildRequest(
	operation string,
	canonicalRequest []byte,
	options EnvelopeOptions,
) workerRequestFrame {
	request := workerRequestFrame{
		ProtocolVersion: workerProtocolVersion,
		Type:            "request",
		RequestID:       "bridge-request-" + strconv.FormatUint(p.nextID.Add(1), 10),
		Operation:       operation,
	}
	if operation == "envelope" || operation == "stream" {
		request.Request = append(json.RawMessage(nil), canonicalRequest...)
		request.Options = &workerRequestOptions{
			Provider:              strings.TrimSpace(options.Provider),
			ProviderProfile:       strings.TrimSpace(options.ProviderProfile),
			Model:                 strings.TrimSpace(options.Model),
			BaseURL:               strings.TrimSpace(options.BaseURL),
			APIKey:                strings.TrimSpace(options.APIKey),
			Temperature:           options.Temperature,
			RequestTimeoutSeconds: options.RequestTimeout.Seconds(),
		}
		if request.Options.Provider == "" {
			request.Options.Provider = "mock"
		}
		if request.Options.RequestTimeoutSeconds <= 0 {
			request.Options.RequestTimeoutSeconds = 120
		}
	}
	return request
}

func (p *workerPool) acquireAdmission(ctx context.Context) error {
	select {
	case <-p.closed:
		return newBridgeError(ErrorCodeClientClosed, "bridge client is closed")
	case <-ctx.Done():
		return bridgeContextError(ctx.Err())
	default:
	}
	select {
	case p.admission <- struct{}{}:
		return nil
	default:
		return newBridgeError(ErrorCodeWorkerQueueFull, "bridge worker queue is full")
	}
}

func (p *workerPool) releaseAdmission() {
	select {
	case <-p.admission:
	default:
	}
}

func (p *workerPool) acquireWorker(ctx context.Context) (*workerProcess, error) {
	select {
	case worker := <-p.idle:
		if worker == nil {
			return nil, newBridgeError(ErrorCodeWorkerUnavailable, "bridge worker is unavailable")
		}
		return worker, nil
	case <-p.closed:
		return nil, newBridgeError(ErrorCodeClientClosed, "bridge client is closed")
	case <-ctx.Done():
		return nil, bridgeContextError(ctx.Err())
	}
}

func (p *workerPool) returnWorker(worker *workerProcess) {
	if worker == nil {
		return
	}
	select {
	case p.idle <- worker:
	case <-p.closed:
		p.removeWorker(worker)
		worker.stop(true, p.config.shutdownTimeout)
	}
}

func (p *workerPool) removeWorker(worker *workerProcess) {
	p.stateMu.Lock()
	delete(p.workers, worker)
	p.stateMu.Unlock()
}

func (p *workerPool) replenish() {
	startupTimeout := p.config.handshakeTimeout * time.Duration(p.config.workerCount)
	if startupTimeout <= 0 {
		startupTimeout = DefaultHandshakeTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()
	_ = p.ensureCapacity(ctx)
}

func (p *workerPool) processStarts() int64 {
	if p == nil {
		return 0
	}
	return p.starts.Load()
}

func (p *workerPool) close() {
	if p == nil {
		return
	}
	p.closeOnce.Do(func() {
		close(p.closed)
		p.startMu.Lock()
		p.stateMu.Lock()
		workers := make([]*workerProcess, 0, len(p.workers))
		for worker := range p.workers {
			workers = append(workers, worker)
		}
		p.workers = make(map[*workerProcess]struct{})
		p.stateMu.Unlock()
		p.startMu.Unlock()
		var waitGroup sync.WaitGroup
		waitGroup.Add(len(workers))
		for _, worker := range workers {
			go func(worker *workerProcess) {
				defer waitGroup.Done()
				worker.stop(true, p.config.shutdownTimeout)
			}(worker)
		}
		waitGroup.Wait()
	})
}

func (p *workerPool) isClosed() bool {
	select {
	case <-p.closed:
		return true
	default:
		return false
	}
}

func startWorkerProcess(ctx context.Context, config workerPoolConfig) (*workerProcess, error) {
	if config.commandFactory == nil {
		return nil, newBridgeError(ErrorCodeWorkerUnavailable, "bridge worker command is unavailable")
	}
	command, err := config.commandFactory()
	if err != nil {
		return nil, err
	}
	stdin, err := command.StdinPipe()
	if err != nil {
		return nil, newBridgeError(ErrorCodeWorkerUnavailable, "bridge worker stdin is unavailable")
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, newBridgeError(ErrorCodeWorkerUnavailable, "bridge worker stdout is unavailable")
	}
	command.Stderr = io.Discard
	if err := command.Start(); err != nil {
		_ = stdin.Close()
		return nil, newBridgeError(ErrorCodeWorkerUnavailable, "bridge worker could not start")
	}

	worker := &workerProcess{
		command:  command,
		rawStdin: stdin,
		scanner:  bufio.NewScanner(stdout),
		waitDone: make(chan struct{}),
	}
	worker.scanner.Buffer(make([]byte, 0, 64*1024), workerMaximumFrameSize)
	go func() {
		_ = command.Wait()
		close(worker.waitDone)
	}()

	handshakeContext, cancel := context.WithTimeout(ctx, config.handshakeTimeout)
	defer cancel()
	worker.mu.Lock()
	frame, readErr := worker.readFrameLocked(handshakeContext)
	worker.mu.Unlock()
	if readErr != nil {
		worker.terminate(false, config.shutdownTimeout)
		if errors.Is(readErr, context.DeadlineExceeded) || errors.Is(readErr, context.Canceled) {
			return nil, newBridgeError(ErrorCodeWorkerUnavailable, "bridge worker handshake failed")
		}
		return nil, newBridgeError(ErrorCodeWorkerExited, "bridge worker exited during handshake")
	}
	if !validWorkerReadyFrame(frame) {
		worker.terminate(false, config.shutdownTimeout)
		return nil, newBridgeError(ErrorCodeWorkerProtocol, "bridge worker handshake is incompatible")
	}
	return worker, nil
}

func validWorkerReadyFrame(frame workerResponseFrame) bool {
	if frame.ProtocolVersion != workerProtocolVersion || frame.Type != "ready" || frame.RequestID != "" {
		return false
	}
	required := map[string]bool{
		"envelope":  false,
		"stream":    false,
		"providers": false,
		"inventory": false,
	}
	for _, operation := range frame.Operations {
		if _, ok := required[operation]; ok {
			required[operation] = true
		}
	}
	for _, present := range required {
		if !present {
			return false
		}
	}
	return true
}

func (w *workerProcess) execute(
	ctx context.Context,
	request workerRequestFrame,
	handleEvent func(StreamEvent) error,
) ([]byte, bool, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed.Load() {
		return nil, false, newBridgeError(ErrorCodeClientClosed, "bridge client is closed")
	}
	encodedJSON, err := json.Marshal(request)
	if err != nil || len(encodedJSON)+1 > workerMaximumFrameSize {
		clear(encodedJSON)
		return nil, true, newBridgeError(ErrorCodeWorkerProtocol, "bridge worker request frame is invalid")
	}
	encodedRequest := make([]byte, len(encodedJSON)+1)
	copy(encodedRequest, encodedJSON)
	encodedRequest[len(encodedJSON)] = '\n'
	clear(encodedJSON)
	writeErr := w.writeRequestWithContext(ctx, encodedRequest)
	clear(encodedRequest)
	if request.Options != nil {
		request.Options.APIKey = ""
	}
	if writeErr != nil {
		if ctx.Err() != nil {
			return nil, false, bridgeContextError(ctx.Err())
		}
		if w.closed.Load() {
			return nil, false, newBridgeError(ErrorCodeClientClosed, "bridge client is closed")
		}
		w.terminate(false, defaultShutdownTimeout)
		return nil, false, newBridgeError(ErrorCodeWorkerExited, "bridge worker exited before request write")
	}

	for {
		frame, err := w.readFrameLocked(ctx)
		if err != nil {
			if w.closed.Load() {
				return nil, false, newBridgeError(ErrorCodeClientClosed, "bridge client is closed")
			}
			w.terminate(false, defaultShutdownTimeout)
			if ctx.Err() != nil {
				return nil, false, bridgeContextError(ctx.Err())
			}
			return nil, false, newBridgeError(ErrorCodeWorkerExited, "bridge worker exited before completing request")
		}
		if frame.ProtocolVersion != workerProtocolVersion || frame.RequestID != request.RequestID {
			w.terminate(false, defaultShutdownTimeout)
			return nil, false, newBridgeError(ErrorCodeWorkerProtocol, "bridge worker response correlation failed")
		}
		switch frame.Type {
		case "stream_event":
			if request.Operation != "stream" || handleEvent == nil || len(frame.Event) == 0 {
				w.terminate(false, defaultShutdownTimeout)
				return nil, false, newBridgeError(ErrorCodeWorkerProtocol, "bridge worker stream frame is invalid")
			}
			var event StreamEvent
			if err := json.Unmarshal(frame.Event, &event); err != nil || strings.TrimSpace(event.Type) == "" {
				w.terminate(false, defaultShutdownTimeout)
				return nil, false, newBridgeError(ErrorCodeWorkerProtocol, "bridge worker stream frame is invalid")
			}
			if err := handleEvent(event); err != nil {
				w.terminate(false, defaultShutdownTimeout)
				return nil, false, newBridgeError(ErrorCodeWorkerCanceled, "bridge stream consumer stopped")
			}
		case "result":
			if request.Operation == "stream" {
				return nil, true, nil
			}
			if len(frame.Payload) == 0 || string(frame.Payload) == "null" {
				w.terminate(false, defaultShutdownTimeout)
				return nil, false, newBridgeError(ErrorCodeWorkerProtocol, "bridge worker result frame is invalid")
			}
			return append([]byte(nil), frame.Payload...), true, nil
		case "error":
			if frame.Error == nil {
				w.terminate(false, defaultShutdownTimeout)
				return nil, false, newBridgeError(ErrorCodeWorkerProtocol, "bridge worker error frame is invalid")
			}
			switch strings.TrimSpace(frame.Error.Code) {
			case ErrorCodeWorkerProtocol:
				return nil, true, newBridgeError(ErrorCodeWorkerProtocol, "bridge worker rejected the request frame")
			case ErrorCodeWorkerRequestFailed:
				return nil, true, newBridgeError(ErrorCodeWorkerRequestFailed, "bridge worker request failed")
			default:
				w.terminate(false, defaultShutdownTimeout)
				return nil, false, newBridgeError(ErrorCodeWorkerProtocol, "bridge worker error code is invalid")
			}
		default:
			w.terminate(false, defaultShutdownTimeout)
			return nil, false, newBridgeError(ErrorCodeWorkerProtocol, "bridge worker response frame type is invalid")
		}
	}
}

func (w *workerProcess) writeRequestWithContext(ctx context.Context, content []byte) error {
	writeDone := make(chan error, 1)
	go func() {
		writeDone <- writeWorkerRequest(w.rawStdin, content)
	}()

	select {
	case err := <-writeDone:
		return err
	case <-ctx.Done():
		w.terminate(false, 0)
		<-writeDone
		return ctx.Err()
	}
}

func writeWorkerRequest(writer io.Writer, content []byte) error {
	for len(content) > 0 {
		written, err := writer.Write(content)
		if err != nil {
			return err
		}
		if written <= 0 {
			return io.ErrShortWrite
		}
		content = content[written:]
	}
	return nil
}

func (w *workerProcess) readFrameLocked(ctx context.Context) (workerResponseFrame, error) {
	resultChannel := make(chan workerScanResult, 1)
	go func() {
		if w.scanner.Scan() {
			resultChannel <- workerScanResult{line: w.scanner.Text()}
			return
		}
		resultChannel <- workerScanResult{err: w.scanner.Err()}
	}()

	select {
	case <-ctx.Done():
		return workerResponseFrame{}, ctx.Err()
	case result := <-resultChannel:
		if result.err != nil || strings.TrimSpace(result.line) == "" {
			return workerResponseFrame{}, errors.New("bridge worker response stream closed")
		}
		var frame workerResponseFrame
		if err := json.Unmarshal([]byte(result.line), &frame); err != nil {
			return workerResponseFrame{}, errors.New("bridge worker response frame is invalid")
		}
		return frame, nil
	}
}

func (w *workerProcess) stop(graceful bool, timeout time.Duration) {
	if w == nil {
		return
	}
	w.terminate(graceful, timeout)
}

func (w *workerProcess) terminate(graceful bool, timeout time.Duration) {
	w.terminateOnce.Do(func() {
		w.closed.Store(true)
		_ = w.rawStdin.Close()
	})
	if graceful && timeout > 0 {
		if w.waitForExit(timeout) {
			return
		}
	}
	if w.command.Process != nil {
		_ = w.command.Process.Kill()
	}
	if timeout > 0 {
		_ = w.waitForExit(timeout)
	}
}

func (w *workerProcess) waitForExit(timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-w.waitDone:
		return true
	case <-timer.C:
		return false
	}
}
