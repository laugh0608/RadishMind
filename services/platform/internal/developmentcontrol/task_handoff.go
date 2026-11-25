package developmentcontrol

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Prepare creates private directories and a bound sandbox only. The runtime
// owns process creation, model dispatch and delivery of the credential on stdin.
type TaskWorkerEnvironment interface {
	Prepare(context.Context, TaskResourceOwner, TaskPlan, Delegation) (WorkerLaunch, func() error, error)
}

type taskWorkerTracker struct {
	runtime *Runtime
	owner   TaskResourceOwner
	host    TaskHost
	cancel  context.CancelFunc
}

// StartWorker is the host entry for asynchronous task execution. The caller
// receives the durable receipt even if the subsequent handoff fails. Only
// the first submission has a completion channel. No credential leaves here.
func (c *TaskCoordinator) StartWorker(ctx context.Context, s TaskScope, task, request string, version int64, preparation, confirmation string, environment TaskWorkerEnvironment) (TaskReceipt, <-chan error, error) {
	if environment == nil {
		return TaskReceipt{}, nil, ErrUnavailable
	}
	// Account for the gap between commit and RunWorker registration. Close
	// waits for final evidence and slot reconciliation as well as the child.
	if err := c.runtime.beginIssue(); err != nil {
		return TaskReceipt{}, nil, err
	}
	receipt, issued, err := c.start(ctx, s, task, request, version, preparation, confirmation, true)
	if err != nil || issued == nil {
		c.runtime.work.Done()
		return receipt, nil, err
	}
	done := make(chan error, 1)
	go func() {
		defer c.runtime.work.Done()
		defer close(done)
		err := c.handoff(s, receipt, *issued, environment)
		// An unacknowledged handoff must fence the credential even when the
		// error precedes RunWorker's claim. Never issue a replacement.
		finish, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err != nil {
			err = errors.Join(err, c.runtime.stopWorkerRecord(finish, s.Identity, receipt.AttemptID, false))
		}
		err = errors.Join(err, c.runtime.finishTaskResources(finish, s, receipt.AttemptID, false))
		done <- err
	}()
	return receipt, done, nil
}

func (c *TaskCoordinator) handoff(s TaskScope, receipt TaskReceipt, issued Issued, environment TaskWorkerEnvironment) (returned error) {
	r := c.runtime
	attempt := receipt.AttemptID
	ctx, cancel := context.WithDeadline(context.Background(), issued.deadline)
	defer cancel()
	plan, err := readTaskPlan(ctx, r.store.db, s, receipt.TaskID, receipt.Revision)
	if err != nil {
		return err
	}
	if digest, err := plan.AttemptPlan.Digest(); err != nil || digest != issued.Delegation.Scope.PlanDigest {
		return ErrTaskCorrupt
	}
	owner := TaskResourceOwner{s, receipt.TaskID, r.instanceID, SandboxBindingFor(issued.Delegation), plan.SDKImageDigest, plan.GoImageDigest, plan.ExecutorDigest}
	tracker := &taskWorkerTracker{r, owner, c.host, cancel}
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return ErrUnavailable
	}
	if r.taskWorkers == nil {
		r.taskWorkers = make(map[string]*taskWorkerTracker)
	}
	r.taskWorkers[attempt] = tracker
	r.mu.Unlock()
	defer func() { r.mu.Lock(); delete(r.taskWorkers, attempt); r.mu.Unlock() }()
	live, err := r.store.authenticate(ctx, r.instanceID, attempt, issued.Credential)
	if err != nil {
		return err
	}
	if live.Revoked || live.Control.phase != "active" {
		return AttemptStopped
	}
	if err := tracker.host.AuthorizeTask(ctx, s, TaskStart); err != nil {
		return err
	}
	if err := tracker.observe(ctx, "code_environment", "pending", 0); err != nil {
		return err
	}
	launch, release, err := environment.Prepare(ctx, owner, plan, issued.Delegation)
	if err != nil {
		// A factory can prepare files, but cannot dispatch code or start the SDK.
		finish, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		return errors.Join(err, tracker.observe(finish, "code_environment", "confirmed", 0))
	}
	if release != nil {
		defer func() {
			finish, stop := context.WithTimeout(context.Background(), 5*time.Second)
			defer stop()
			resources, err := readTaskResources(finish, r.store.db, s, attempt)
			if err != nil || !resourceEnded(resources.SDKWorker) {
				returned = errors.Join(returned, err, ErrTaskTerminationUnconfirmed)
				return
			}
			returned = errors.Join(returned, release())
		}()
	}
	if launch.Sandbox == nil || launch.Sandbox.Binding() != owner.Sandbox {
		return ErrIneligible
	}
	owned := &trackedTaskSandbox{WorkerSandbox: launch.Sandbox, tracker: tracker}
	launch.Sandbox = owned
	defer func() {
		finish, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		_, err := owned.Close(finish)
		returned = errors.Join(returned, err)
	}()
	if err := tracker.observe(ctx, "code_environment", "active", 0); err != nil {
		return err
	}
	_, err = r.RunWorker(ctx, s.Identity, issued, launch, nil)
	return err
}

func (r *Runtime) taskWorker(id string) *taskWorkerTracker {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.taskWorkers[id]
}

// RunWorker and its pre-claim error path share one cleanup result.
type trackedTaskSandbox struct {
	WorkerSandbox
	tracker *taskWorkerTracker
	once    sync.Once
	empty   bool
	err     error
}

func (s *trackedTaskSandbox) Close(ctx context.Context) (bool, error) {
	s.once.Do(func() {
		s.empty, s.err = s.WorkerSandbox.Close(ctx)
		state := "confirmed"
		if !s.empty || s.err != nil {
			state = "unconfirmed"
			s.err = errors.Join(s.err, ErrTaskTerminationUnconfirmed)
		}
		s.err = errors.Join(s.err, s.tracker.observe(ctx, "code_environment", state, 0))
	})
	return s.empty && s.err == nil, s.err
}
