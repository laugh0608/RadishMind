package developmentcontrol

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// Only the explicitly authorized isolated SDK image sets these paths. Normal
// repository tests never install the SDK, start Docker or substitute a fake Agent.
func sdkLaunch(t *testing.T, f *runtimeFixture) (WorkerLaunch, *atomic.Int32) {
	t.Helper()
	python, script := os.Getenv("RADISHMIND_SDK_PYTHON"), os.Getenv("RADISHMIND_SDK_SCRIPT")
	if python == "" || script == "" {
		t.Skip("real SDK worker requires explicit isolated runtime window")
	}
	count := &atomic.Int32{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/chat/completions") {
			count.Add(1)
		}
		f.handler.ServeHTTP(w, r)
	}))
	t.Cleanup(server.Close)
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	return WorkerLaunch{WorkerCommand{python, []string{script}, dir}, server.URL, initialRequest(t).Messages}, count
}
func TestSDKWorkerAgentFixedResponseRoundTrip(t *testing.T) {
	f := newRuntimeFixture(t, func(_ context.Context, input NativeRequest) (NativeResult, error) {
		if len(input.Messages) == 2 {
			return toolResult(), nil
		}
		return stopResult(), nil
	})
	launch, count := sdkLaunch(t, f)
	result, err := f.runtime.RunWorker(context.Background(), f.identity, f.issued, launch, func(_ context.Context, call ToolCall) (string, error) {
		if call.Arguments != readAction() {
			return "", ErrContext
		}
		return "synthetic SDK observation", nil
	})
	if err != nil || !result.Finished || !result.ProcessExited || !result.GroupEmpty {
		t.Fatalf("SDK roundtrip: %+v %v", result, err)
	}
	if count.Load() != 2 || f.backend.calls.Load() != 2 {
		t.Fatal("SDK call bypass or hidden retry", count.Load(), f.backend.calls.Load())
	}
	record := f.read(t)
	kinds := map[string]bool{}
	for _, event := range record.Control.Events() {
		kinds[event.Type] = true
	}
	for _, kind := range []string{"sdk_system_prompt_observed", "sdk_user_message_observed", "sdk_action_observed", "sdk_observation_observed", "sdk_assistant_message_observed", "worker_waiting_tool", "worker_exited"} {
		if !kinds[kind] {
			t.Fatal("missing SDK event", kind)
		}
	}
	if record.Control.phase != "completed" || record.Control.budget.SpentTokens != 70 {
		t.Fatal("SDK did not reconcile through Go")
	}
}

func TestSDKWorkerFailureDoesNotRetry(t *testing.T) {
	for _, mode := range []string{"native_failure", "tool_failure", "missing_usage"} {
		t.Run(mode, func(t *testing.T) {
			f := newRuntimeFixture(t, func(_ context.Context, request NativeRequest) (NativeResult, error) {
				if mode == "native_failure" {
					return NativeResult{}, errors.New("synthetic native failure")
				}
				if len(request.Messages) == 2 {
					r := toolResult()
					if mode == "missing_usage" {
						r.Usage = nil
					}
					return r, nil
				}
				return stopResult(), nil
			})
			launch, count := sdkLaunch(t, f)
			result, err := f.runtime.RunWorker(context.Background(), f.identity, f.issued, launch, func(context.Context, ToolCall) (string, error) {
				if mode == "tool_failure" {
					return "", errors.New("synthetic tool failure")
				}
				return "synthetic SDK observation", nil
			})
			record := f.read(t)
			if mode == "missing_usage" {
				if err != nil || !result.Finished || count.Load() != 2 || record.Control.budget.HeldTokens != 150 || record.Control.budget.SpentTokens != 40 || record.Control.budget.Unresolved != 1 {
					t.Fatal("SDK missing usage became zero", result, err, record.Control.budget, count.Load())
				}
			} else {
				if err == nil || result.Finished || !result.ProcessExited || !result.GroupEmpty {
					t.Fatal(result, err)
				}
				if count.Load() != 1 || f.backend.calls.Load() != 1 {
					t.Fatal("hidden SDK retry", count.Load())
				}
				if record.Control.budget.Unresolved != 1 {
					t.Fatal("unknown effect lost")
				}
				if mode == "native_failure" && record.Control.budget.HeldTokens != 150 {
					t.Fatal("unknown model usage lost")
				}
				if mode == "tool_failure" && (record.Control.budget.HeldTokens != 0 || record.Control.budget.InFlight != 0) {
					t.Fatal("unknown tool forged model usage")
				}
			}
		})
	}
}

func TestSDKWorkerModelWaitCancellation(t *testing.T) {
	started := make(chan struct{})
	f := newRuntimeFixture(t, func(ctx context.Context, _ NativeRequest) (NativeResult, error) {
		close(started)
		<-ctx.Done()
		return NativeResult{}, ctx.Err()
	})
	launch, _ := sdkLaunch(t, f)
	type outcome struct {
		r WorkerResult
		e error
	}
	done := make(chan outcome, 1)
	go func() {
		r, e := f.runtime.RunWorker(context.Background(), f.identity, f.issued, launch, nil)
		done <- outcome{r, e}
	}()
	select {
	case <-started:
	case early := <-done:
		t.Fatalf("SDK ended before native call: %+v %v", early.r, early.e)
	case <-time.After(25 * time.Second):
		t.Fatal("SDK did not reach Go native backend")
	}
	record := f.read(t)
	if _, err := f.runtime.Cancel(context.Background(), f.identity, f.issued.Delegation.Scope.AttemptID, record.Version); err != nil {
		t.Fatal(err)
	}
	select {
	case result := <-done:
		if result.e == nil || !result.r.ProcessExited || !result.r.GroupEmpty {
			t.Fatal(result)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("SDK process remained alive")
	}
	record = f.read(t)
	if record.Control.phase != "outcome_unknown" || record.Control.budget.HeldTokens != 150 {
		t.Fatal("cancel lost uncertainty")
	}
	t.Run("parent_liveness_pipe", sdkParentDisconnectCase)
}

func sdkParentDisconnectCase(t *testing.T) {
	started := make(chan struct{})
	f := newRuntimeFixture(t, func(ctx context.Context, _ NativeRequest) (NativeResult, error) {
		close(started)
		<-ctx.Done()
		return NativeResult{}, ctx.Err()
	})
	launch, _ := sdkLaunch(t, f)
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer read.Close()
	defer write.Close()
	out, childOut, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	defer childOut.Close()
	cmd := exec.Command(launch.Command.Path, launch.Command.Args...)
	cmd.Dir = launch.Command.Dir
	cmd.Env = workerEnvironment(cmd.Dir)
	cmd.ExtraFiles = []*os.File{read}
	cmd.Stdout = childOut
	if err = prepareWorkerProcess(cmd); err != nil {
		t.Fatal(err)
	}
	input, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	read.Close()
	childOut.Close()
	waited := make(chan error, 1)
	go func() { waited <- cmd.Wait() }()
	// Cleanup is only a backstop; the assertion below observes exit before it runs.
	t.Cleanup(func() { _ = signalWorkerGroup(cmd.Process.Pid, true) })
	tools, err := NativeTools()
	if err != nil {
		t.Fatal(err)
	}
	in := WorkerInput{"development_worker_input.v1", f.issued.Delegation.Scope.AttemptID, f.issued.Delegation.ExecutorInstanceID, launch.ChannelURL, f.issued.Credential, "fixture-model", launch.InitialMessages, tools, 50, 4}
	if err = json.NewEncoder(input).Encode(in); err != nil {
		t.Fatal(err)
	}
	framesDone := make(chan struct{})
	go func() {
		defer close(framesDone)
		scanner := bufio.NewScanner(out)
		for scanner.Scan() {
			var frame workerFrame
			if json.Unmarshal(scanner.Bytes(), &frame) != nil {
				return
			}
			id := "sdk_event"
			if frame.Kind == "ready" {
				id = "worker_ready"
			}
			if ackWorker(input, frame.Sequence, id) != nil {
				return
			}
		}
	}()
	select {
	case <-started:
	case early := <-waited:
		t.Fatalf("SDK stopped before liveness test: %v", early)
	case <-time.After(20 * time.Second):
		_ = signalWorkerGroup(cmd.Process.Pid, true)
		t.Fatal("SDK did not start")
	}
	// No Go cancel/signal: EOF alone must make Python stop its owned group.
	write.Close()
	select {
	case <-waited:
	case <-time.After(3 * time.Second):
		t.Fatal("SDK survived lost Go liveness pipe")
	}
	<-framesDone
	if !workerGroupEmpty(cmd.Process.Pid) {
		t.Fatal("SDK owned group remains after parent loss")
	}
}
