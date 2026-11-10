//go:build linux || darwin

package developmentcontrol

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestSDKWorkerHelper(t *testing.T) {
	marker := -1
	for i, arg := range os.Args {
		if arg == "--worker-helper" {
			marker = i
			break
		}
	}
	if marker < 0 {
		return
	}
	mode := os.Args[marker+1]
	if mode == "child" {
		signal.Ignore(syscall.SIGTERM)
		for {
			time.Sleep(time.Second)
		}
	}
	var input WorkerInput
	decoder := json.NewDecoder(os.Stdin)
	if decoder.Decode(&input) != nil {
		os.Exit(3)
	}
	if os.Getenv("RADISHMIND_TEST_SECRET") != "" {
		os.Exit(4)
	}
	encoder := json.NewEncoder(os.Stdout)
	send := func(seq int, kind string) {
		if encoder.Encode(map[string]any{"schema_version": "development_worker_message.v1", "sequence": seq, "kind": kind, "event_type": "", "request_id": ""}) != nil {
			os.Exit(5)
		}
		var reply workerReply
		if decoder.Decode(&reply) != nil || reply.Sequence != int64(seq) {
			os.Exit(10)
		}
	}
	send(1, "ready")
	if mode == "malformed" {
		fmt.Fprintln(os.Stdout, `{"schema_version":"development_worker_message.v1","sequence":2,"kind":"event","event_type":"user_message","request_id":"","raw_text":"forbidden"}`)
		for {
			time.Sleep(time.Second)
		}
	}
	if mode == "wait" || mode == "spawn" {
		signal.Ignore(syscall.SIGTERM)
		if mode == "spawn" {
			exe, _ := os.Executable()
			child := exec.Command(exe, "-test.run=^TestSDKWorkerHelper$", "--", "--worker-helper", "child")
			if child.Start() != nil {
				os.Exit(6)
			}
		}
		for {
			time.Sleep(time.Second)
		}
	}
	body, _ := json.Marshal(NativeRequest{Model: input.Model, Messages: input.InitialMessages, MaxCompletionTokens: input.MaxOutputTokens, Tools: input.Tools})
	req, _ := http.NewRequest("POST", input.ChannelURL+"/internal/development/v1/attempts/"+input.AttemptID+"/chat/completions", strings.NewReader(string(body)))
	req.Header.Set("Authorization", "Bearer "+input.Credential)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-RadishMind-Development-Request-ID", "model_helper")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		os.Exit(7)
	}
	io.Copy(io.Discard, response.Body)
	response.Body.Close()
	if response.StatusCode != 200 {
		os.Exit(8)
	}
	if mode != "no_finished" {
		send(2, "finished")
	}
	if mode == "exit_nonzero" {
		os.Exit(9)
	}
	os.Exit(0)
}
func helperLaunch(t *testing.T, f *runtimeFixture, mode string) (WorkerLaunch, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(f.handler)
	t.Cleanup(server.Close)
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	return WorkerLaunch{WorkerCommand{exe, []string{"-test.run=^TestSDKWorkerHelper$", "--", "--worker-helper", mode}, dir}, server.URL, initialRequest(t).Messages, nil}, server
}
func TestWorkerOwnedProcessAndSingleClaim(t *testing.T) {
	t.Setenv("RADISHMIND_TEST_SECRET", "must-not-be-inherited")
	f := newRuntimeFixture(t, func(context.Context, NativeRequest) (NativeResult, error) { return stopResult(), nil })
	launch, _ := helperLaunch(t, f, "success")
	result, err := f.runtime.RunWorker(context.Background(), f.identity, f.issued, launch, nil)
	if err != nil || !result.Finished || !result.ProcessExited || !result.GroupEmpty {
		t.Fatal(result, err)
	}
	if f.backend.calls.Load() != 1 {
		t.Fatal("unexpected calls")
	}
	record := f.read(t)
	kinds := map[string]bool{}
	for _, event := range record.Control.Events() {
		kinds[event.Type] = true
	}
	for _, kind := range []string{"worker_claimed", "worker_started", "worker_ready", "sdk_finished_observed", "worker_exited"} {
		if !kinds[kind] {
			t.Fatal("missing event", kind)
		}
	}
	if _, err = f.runtime.RunWorker(context.Background(), f.identity, f.issued, launch, nil); err == nil {
		t.Fatal("worker was replayed")
	}
}
func TestWorkerFailureIsNotCompletion(t *testing.T) {
	for _, mode := range []string{"malformed", "exit_nonzero", "no_finished"} {
		t.Run(mode, func(t *testing.T) {
			f := newRuntimeFixture(t, func(context.Context, NativeRequest) (NativeResult, error) { return stopResult(), nil })
			launch, _ := helperLaunch(t, f, mode)
			result, err := f.runtime.RunWorker(context.Background(), f.identity, f.issued, launch, nil)
			if err == nil || result.Finished || !result.ProcessExited || !result.GroupEmpty {
				t.Fatal(result, err)
			}
			if f.read(t).Control.phase == "completed" {
				t.Fatal("failed SDK was recorded completed")
			}
			if !f.read(t).Revoked {
				t.Fatal("failed worker credential remains live")
			}
		})
	}
}
func TestWorkerCancelKillsOwnedGroup(t *testing.T) {
	for _, mode := range []string{"wait", "spawn"} {
		t.Run(mode, func(t *testing.T) {
			f := newRuntimeFixture(t, func(context.Context, NativeRequest) (NativeResult, error) { return stopResult(), nil })
			launch, _ := helperLaunch(t, f, mode)
			type outcome struct {
				r WorkerResult
				e error
			}
			done := make(chan outcome, 1)
			go func() {
				r, e := f.runtime.RunWorker(context.Background(), f.identity, f.issued, launch, nil)
				done <- outcome{r, e}
			}()
			until := time.Now().Add(4 * time.Second)
			for {
				record := f.read(t)
				ready := false
				for _, e := range record.Control.Events() {
					ready = ready || e.Type == "worker_ready"
				}
				if ready {
					break
				}
				if time.Now().After(until) {
					t.Fatal("worker did not start")
				}
				time.Sleep(10 * time.Millisecond)
			}
			// Give the helper time to install its signal handler and spawn its child.
			time.Sleep(100 * time.Millisecond)
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
				t.Fatal("owned process group did not stop")
			}
			if f.read(t).Control.phase == "cancelled" {
				t.Fatal("SDK exit forged full sandbox termination")
			}
		})
	}
}
func TestWorkerProtocolRejectsSpoofedReports(t *testing.T) {
	for _, body := range []string{
		`{"schema_version":"development_worker_message.v1","sequence":2,"kind":"ready","event_type":"","request_id":""}`,
		`{"schema_version":"development_worker_message.v1","sequence":1,"kind":"event","event_type":"cancelled","request_id":""}`,
		`{"schema_version":"development_worker_message.v1","sequence":1,"kind":"ready","event_type":"","request_id":"","credential":"leak"}`,
	} {
		if _, err := parseWorkerFrame([]byte(body), 1); err == nil {
			t.Fatal("spoof accepted")
		}
	}
}

type ownedTestSandbox struct {
	binding SandboxBinding
	empty   bool
}

func (s *ownedTestSandbox) Binding() SandboxBinding { return s.binding }
func (s *ownedTestSandbox) HandleTool(context.Context, ToolCall) (string, error) {
	return "", ErrWorker
}
func (s *ownedTestSandbox) Close(context.Context) (bool, error) { return s.empty, nil }
func TestWorkerSandboxCleanupGatesCompletion(t *testing.T) {
	for _, empty := range []bool{true, false} {
		t.Run(fmt.Sprint(empty), func(t *testing.T) {
			f := newRuntimeFixture(t, func(context.Context, NativeRequest) (NativeResult, error) { return stopResult(), nil })
			launch, _ := helperLaunch(t, f, "success")
			launch.Sandbox = &ownedTestSandbox{SandboxBindingFor(f.issued.Delegation), empty}
			result, err := f.runtime.RunWorker(context.Background(), f.identity, f.issued, launch, nil)
			record := f.read(t)
			if empty {
				if err != nil || !result.Finished || record.Control.phase != "completed" {
					t.Fatal(result, err)
				}
			} else {
				if err == nil || result.Finished || record.Control.phase == "completed" || record.Control.terminationStatus != "unconfirmed" {
					t.Fatal("unconfirmed sandbox completed", result, err, record.Control.phase, record.Control.terminationStatus)
				}
			}
		})
	}
}
