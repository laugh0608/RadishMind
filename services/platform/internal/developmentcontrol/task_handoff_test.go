//go:build darwin || linux

package developmentcontrol

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

type taskEnvironmentFunc func(context.Context, TaskResourceOwner, TaskPlan, Delegation) (WorkerLaunch, func() error, error)

func (f taskEnvironmentFunc) Prepare(ctx context.Context, o TaskResourceOwner, p TaskPlan, d Delegation) (WorkerLaunch, func() error, error) {
	return f(ctx, o, p, d)
}

func TestTaskWorkerHandoffAndDurableReplay(t *testing.T) {
	t.Run("completed", func(t *testing.T) { testTaskWorkerHandoff(t, false) })
	t.Run("cancelled with outstanding model call", func(t *testing.T) { testTaskWorkerHandoff(t, true) })
}

func testTaskWorkerHandoff(t *testing.T, cancelWhileActive bool) {
	f := newTaskFixture(t)
	messages := initialRequest(t).Messages
	raw, _ := json.Marshal(messages)
	f.profile.AttemptPlan.InitialMessagesDigest = PayloadDigest(raw)
	entered, release := make(chan struct{}), make(chan struct{})
	var calls, opens atomic.Int32
	f.runtime.backend = &fixedNative{complete: func(ctx context.Context, _ NativeRequest) (NativeResult, error) {
		calls.Add(1)
		close(entered)
		if cancelWhileActive {
			<-release // Cancellation alone cannot prove that this call has returned.
			return NativeResult{}, ctx.Err()
		}
		select {
		case <-release:
			return stopResult(), nil
		case <-ctx.Done():
			return NativeResult{}, ctx.Err()
		}
	}}
	server := httptest.NewServer(f.runtime.Handler())
	defer server.Close()
	finishCall := sync.OnceFunc(func() { close(release) })
	defer finishCall()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err = os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	env := taskEnvironmentFunc(func(ctx context.Context, o TaskResourceOwner, p TaskPlan, d Delegation) (WorkerLaunch, func() error, error) {
		opens.Add(1)
		if o.Sandbox != SandboxBindingFor(d) || p.SDKImageDigest != o.SDKImageDigest || o.Scope != f.scope {
			return WorkerLaunch{}, nil, ErrIneligible
		}
		return WorkerLaunch{WorkerCommand{exe, []string{"-test.run=^TestSDKWorkerHelper$", "--", "--worker-helper", "success"}, dir}, server.URL, messages, &ownedTestSandbox{o.Sandbox, true}}, nil, nil
	})
	task := f.create(t, "create_worker")
	p := f.preview(t, task.TaskID, "prepare_worker")
	submit := func(ctx context.Context) (TaskReceipt, <-chan error, error) {
		return f.StartWorker(ctx, f.scope, task.TaskID, "start_worker", p.TaskVersion, p.PreparationID, p.ConfirmationDigest, env)
	}
	request, end := context.WithCancel(t.Context())
	receipt, done, err := submit(request)
	end()
	if err != nil || done == nil {
		t.Fatal(receipt, err)
	}
	select {
	case <-entered:
	case err := <-done:
		t.Fatalf("worker stopped early: %v", err)
	case <-time.After(8 * time.Second):
		t.Fatal("model not called")
	}
	status, err := f.ReadTaskResources(t.Context(), f.scope, task.TaskID, receipt.AttemptID)
	if err != nil || status != (TaskResourceStatus{"active", "active", "active"}) {
		t.Fatal(status, err)
	}
	replayed, again, err := submit(t.Context())
	if err != nil || again != nil || replayed != receipt {
		t.Fatal("live replay", replayed, err)
	}
	if cancelWhileActive {
		rec, err := f.ReadAttempt(t.Context(), f.scope, task.TaskID, receipt.AttemptID)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Cancel(t.Context(), f.scope, task.TaskID, receipt.AttemptID, "cancel_model", rec.Version); err != nil {
			t.Fatal(err)
		}
	} else {
		finishCall()
	}
	select {
	case err := <-done:
		if (err != nil) != cancelWhileActive {
			t.Fatal(err)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("handoff did not finish")
	}
	phase := "completed"
	if cancelWhileActive {
		phase = "outcome_unknown"
		status, err := f.ReadTaskResources(t.Context(), f.scope, task.TaskID, receipt.AttemptID)
		if err != nil || status != (TaskResourceStatus{"stopping", "confirmed", "confirmed"}) || f.count(t, "SELECT count(*) FROM development_tasks WHERE execution_slot=1") != 1 {
			t.Fatal("released slot with outstanding model call", status, err)
		}
		finishCall()
		server.Close() // Wait for late model accounting before reopening the store.
	}
	record, err := f.ReadAttempt(t.Context(), f.scope, task.TaskID, receipt.AttemptID)
	if err != nil || record.Control.phase != phase || !taskResourcesEnded(record) || opens.Load() != 1 || calls.Load() != 1 || f.count(t, "SELECT count(*) FROM development_tasks WHERE execution_slot=1") != 0 {
		t.Fatal("incomplete worker", err)
	}
	var events []string
	for _, e := range record.Control.Events() {
		events = append(events, e.Type)
	}
	if !strings.Contains(strings.Join(events, " "), "worker_claimed worker_started worker_ready") {
		t.Fatal(events)
	}
	if err := f.db.Close(); err != nil {
		t.Fatal(err)
	}
	f.open(t)
	replayed, again, err = submit(t.Context())
	if err != nil || replayed != receipt || again != nil || opens.Load() != 1 {
		t.Fatal("restarted replay", err)
	}
	status, err = f.ReadTaskResources(t.Context(), f.scope, task.TaskID, receipt.AttemptID)
	if err != nil || status != (TaskResourceStatus{"confirmed", "confirmed", "confirmed"}) {
		t.Fatal(status, err)
	}
}

func TestTaskWorkerPreparationFailureAndShutdown(t *testing.T) {
	for _, mode := range []string{"failed", "shutdown", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			f := newTaskFixture(t)
			task := f.create(t, "create_setup")
			p := f.preview(t, task.TaskID, "prepare_setup")
			entered := make(chan struct{})
			env := taskEnvironmentFunc(func(ctx context.Context, _ TaskResourceOwner, _ TaskPlan, _ Delegation) (WorkerLaunch, func() error, error) {
				close(entered)
				if mode != "failed" {
					<-ctx.Done()
				}
				return WorkerLaunch{}, nil, ErrWorker
			})
			receipt, done, err := f.StartWorker(t.Context(), f.scope, task.TaskID, "start_setup", p.TaskVersion, p.PreparationID, p.ConfirmationDigest, env)
			if err != nil || done == nil {
				t.Fatal(err)
			}
			<-entered
			switch mode {
			case "shutdown":
				ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
				defer cancel()
				if err := f.runtime.Close(ctx); err != nil {
					t.Fatal(err)
				}
			case "cancel":
				record, err := f.ReadAttempt(t.Context(), f.scope, task.TaskID, receipt.AttemptID)
				if err != nil {
					t.Fatal(err)
				}
				if _, err = f.Cancel(t.Context(), f.scope, task.TaskID, receipt.AttemptID, "cancel_setup", record.Version); err != nil {
					t.Fatal(err)
				}
			}
			if err = <-done; !errors.Is(err, ErrWorker) {
				t.Fatal(err)
			}
			if r, err := f.ReadReceipt(t.Context(), f.scope, "start_setup"); err != nil || r != receipt {
				t.Fatal("lost committed receipt", err)
			}
			r, err := f.ReadAttempt(t.Context(), f.scope, task.TaskID, receipt.AttemptID)
			if err != nil || !r.Revoked || r.Control.terminationStatus != "confirmed" || hasControlEvent(r, "worker_started") {
				t.Fatal("false launch", err)
			}
			if mode == "cancel" && r.Control.phase != "cancelled" {
				t.Fatal(r.Control.phase)
			}
			if f.count(t, "SELECT count(*) FROM development_tasks WHERE execution_slot=1") != 0 {
				t.Fatal("slot held after known no-launch")
			}
		})
	}
}

func TestTaskWorkerRestartInspectsRecordedGroup(t *testing.T) {
	f := newTaskFixture(t)
	task := f.create(t, "create_crash")
	p := f.preview(t, task.TaskID, "prepare_crash")
	r, _, err := f.startManagedForRecovery(p)
	if err != nil {
		t.Fatal(err)
	}
	// A real process group, created and reaped by this test. Recovery may
	// check its absence but must never signal a numeric PID from the database.
	cmd := exec.Command("/usr/bin/true")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err = cmd.Run(); err != nil {
		t.Fatal(err)
	}
	_, err = f.db.DB().Exec("UPDATE development_task_resources SET worker_pgid=?,sdk_worker='active',code_environment='active',model_connection='confirmed' WHERE attempt_id=?", cmd.Process.Pid, r.AttemptID)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.db.Close(); err != nil {
		t.Fatal(err)
	}
	f.open(t)
	checked := 0
	check := func(_ context.Context, o TaskResourceOwner) (bool, error) {
		checked++
		if o.Scope != f.scope || o.TaskID != task.TaskID || o.Sandbox.AttemptID != r.AttemptID || o.OwnerInstanceID == f.runtime.instanceID {
			t.Fatal(o)
		}
		return checked > 1, nil
	}
	err = RecoverWorkbenchTasks(t.Context(), f.store, check)
	if !errors.Is(err, ErrTaskTerminationUnconfirmed) || f.count(t, "SELECT count(*) FROM development_tasks WHERE execution_slot=1") != 1 {
		t.Fatal(err)
	}
	status, err := f.ReadTaskResources(t.Context(), f.scope, task.TaskID, r.AttemptID)
	if err != nil || status.CodeEnvironment != "unconfirmed" {
		t.Fatal(status, err)
	}
	err = RecoverWorkbenchTasks(t.Context(), f.store, check)
	if processNamespace() == "" {
		if !errors.Is(err, ErrTaskTerminationUnconfirmed) {
			t.Fatal("missing namespace released group", err)
		}
		return
	}
	if err != nil || checked != 2 || f.count(t, "SELECT count(*) FROM development_tasks WHERE execution_slot=1") != 0 {
		t.Fatal(checked, err, status)
	}
	record, err := f.ReadAttempt(t.Context(), f.scope, task.TaskID, r.AttemptID)
	if err != nil || record.Control.phase != "outcome_unknown" || record.Control.terminationStatus != "confirmed" {
		t.Fatal(err)
	}
}

func (f *taskFixture) startManagedForRecovery(p TaskPreparation) (TaskReceipt, *Issued, error) {
	return f.TaskCoordinator.start(context.Background(), f.scope, p.TaskID, "start_crash", p.TaskVersion, p.PreparationID, p.ConfirmationDigest, true)
}

type revocableWorkerHost struct {
	TaskHost
	revoked atomic.Bool
}

func (h *revocableWorkerHost) AuthorizeTask(ctx context.Context, s TaskScope, op TaskOperation) error {
	if op == TaskStart && h.revoked.Load() {
		return ErrIneligible
	}
	return h.TaskHost.AuthorizeTask(ctx, s, op)
}

func TestTaskWorkerIdleRevocationStopsOwnedProcess(t *testing.T) {
	f := newTaskFixture(t)
	host := &revocableWorkerHost{TaskHost: f}
	f.TaskMetadata.host = host
	messages := initialRequest(t).Messages
	raw, _ := json.Marshal(messages)
	f.profile.AttemptPlan.InitialMessagesDigest = PayloadDigest(raw)
	task := f.create(t, "create_idle")
	p := f.preview(t, task.TaskID, "prepare_idle")
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err = os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	env := taskEnvironmentFunc(func(_ context.Context, o TaskResourceOwner, _ TaskPlan, _ Delegation) (WorkerLaunch, func() error, error) {
		return WorkerLaunch{WorkerCommand{exe, []string{"-test.run=^TestSDKWorkerHelper$", "--", "--worker-helper", "wait"}, dir}, "http://127.0.0.1:12345", messages, &ownedTestSandbox{o.Sandbox, true}}, nil, nil
	})
	r, done, err := f.StartWorker(t.Context(), f.scope, p.TaskID, "start_idle", p.TaskVersion, p.PreparationID, p.ConfirmationDigest, env)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		if err := f.runtime.Close(ctx); err != nil {
			t.Error(err)
		}
	}()
	timeout := time.After(8 * time.Second)
	for {
		rec, err := f.ReadAttempt(t.Context(), f.scope, p.TaskID, r.AttemptID)
		if err != nil {
			t.Fatal(err)
		}
		if hasControlEvent(rec, "worker_ready") {
			break
		}
		select {
		case <-timeout:
			t.Fatal("SDK did not become ready")
		case <-time.After(10 * time.Millisecond):
		}
	}
	host.revoked.Store(true)
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("revoked SDK continued")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("idle SDK ignored lost authority")
	}
	rec, err := f.ReadAttempt(t.Context(), f.scope, p.TaskID, r.AttemptID)
	if err != nil || !rec.Revoked || rec.Control.terminationStatus != "confirmed" || rec.Control.phase == "completed" || f.count(t, "SELECT count(*) FROM development_tasks WHERE execution_slot=1") != 0 {
		t.Fatal("unfinished revocation", err)
	}
}
