package developmentcontrol

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

type WorkerLaunch struct {
	Command         WorkerCommand
	ChannelURL      string
	InitialMessages []NativeMessage
	Sandbox         WorkerSandbox
}

// RunWorker starts one owned SDK process after a durable, non-replayable claim.
// The host supplies an explicit command; SDK/model data never chooses a program.
func (r *Runtime) RunWorker(ctx context.Context, identity Identity, issued Issued, launch WorkerLaunch, handler ToolHandler) (result WorkerResult, returned error) {
	if launch.Command.validate() != nil || !validWorkerURL(launch.ChannelURL) {
		return result, ErrWorker
	}
	if err := validateWorkerDirectory(launch.Command.Dir); err != nil {
		return result, err
	}
	record, err := r.store.authenticate(ctx, r.instanceID, issued.Delegation.Scope.AttemptID, issued.Credential)
	if err != nil || record.Binding.Identity != identity || record.Control.delegation != issued.Delegation {
		return result, ErrCredential
	}
	if len(launch.InitialMessages) != 2 || launch.InitialMessages[0].Role != "system" || launch.InitialMessages[1].Role != "user" || checkContext(record, launch.InitialMessages) != nil {
		return result, ErrContext
	}
	if launch.Sandbox != nil {
		if handler != nil || launch.Sandbox.Binding() != SandboxBindingFor(issued.Delegation) {
			return result, ErrWorker
		}
		handler = launch.Sandbox.HandleTool
	}
	deadline, live := r.attemptDeadline(issued.Delegation.Scope.AttemptID)
	if !live {
		return result, ErrCredential
	}
	workerCtx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()
	attempt := issued.Delegation.Scope.AttemptID
	r.mu.Lock()
	if r.closed || r.workers[attempt] != nil || len(r.workers) >= 8 {
		r.mu.Unlock()
		return result, ErrConflict
	}
	r.workers[attempt] = cancel
	r.work.Add(1)
	r.mu.Unlock()
	defer func() { r.mu.Lock(); delete(r.workers, attempt); r.mu.Unlock(); r.work.Done() }()
	var refusal error
	_, err = r.store.change(workerCtx, identity, attempt, 0, r.instanceID, func(current *Record) error {
		refusal = r.revalidate(workerCtx, current)
		if refusal != nil {
			return nil
		}
		for _, e := range current.Control.events {
			if e.Type == "worker_claimed" {
				return ErrConflict
			}
		}
		if current.Control.phase != "active" {
			return AttemptStopped
		}
		current.Control.emit("worker_claimed", "", r.now().UnixMilli())
		return nil
	})
	if err != nil {
		return result, err
	}
	if refusal != nil {
		return result, refusal
	}
	sandboxHandled, sandboxEmpty := false, false
	defer func() {
		if launch.Sandbox != nil && !sandboxHandled {
			cleanup, stop := context.WithTimeout(context.Background(), 10*time.Second)
			empty, cleanupErr := launch.Sandbox.Close(cleanup)
			stop()
			returned = errors.Join(returned, cleanupErr)
			if !empty {
				returned = errors.Join(returned, ErrWorker)
			}
		}
	}()
	defer func() {
		if returned != nil {
			finish, stop := context.WithTimeout(context.Background(), 5*time.Second)
			defer stop()
			returned = errors.Join(returned, r.stopWorkerRecord(finish, identity, attempt, workerCtx.Err() == nil))
		}
	}()
	tools, err := NativeTools()
	if err != nil {
		return result, ErrWorker
	}
	input := WorkerInput{"development_worker_input.v1", attempt, issued.Delegation.ExecutorInstanceID, launch.ChannelURL, issued.Credential, record.Binding.Selection.WireModel, launch.InitialMessages, tools, record.Binding.Plan.MaxOutputTokens, record.Binding.Plan.Limits.MaxModelCalls}
	command := exec.Command(launch.Command.Path, launch.Command.Args...)
	command.Dir = launch.Command.Dir
	command.Env = workerEnvironment(launch.Command.Dir)
	command.Stderr = io.Discard // SDK logs may contain messages/headers; never publish them.
	if err = prepareWorkerProcess(command); err != nil {
		return result, err
	}
	parentRead, parentWrite, err := os.Pipe()
	if err != nil {
		return result, ErrWorker
	}
	defer parentRead.Close()
	defer parentWrite.Close()
	command.ExtraFiles = []*os.File{parentRead}
	inputPipe, err := command.StdinPipe()
	if err != nil {
		return result, ErrWorker
	}
	defer inputPipe.Close()
	output, childOutput, err := os.Pipe()
	if err != nil {
		return result, ErrWorker
	}
	defer output.Close()
	defer childOutput.Close()
	command.Stdout = childOutput
	if workerCtx.Err() != nil {
		return result, ErrWorker
	}
	if err = command.Start(); err != nil {
		return result, ErrWorker
	}
	// Signal/kill and reap run independently of a blocked SDK or tool wait.
	processDone := make(chan WorkerResult, 1)
	childOutput.Close()
	parentRead.Close()
	go func() {
		proof := watchWorkerProcess(workerCtx, command)
		if !proof.ProcessExited {
			output.Close()
		}
		processDone <- proof
	}()
	finishedProcess := false
	defer func() {
		failed := workerCtx.Err() == nil
		r.mu.Lock()
		r.fenced[attempt] = true
		r.mu.Unlock()
		cancel()
		parentWrite.Close()
		r.cancelActive(attempt)
		if returned != nil {
			fctx, fcancel := context.WithTimeout(context.Background(), 5*time.Second)
			returned = errors.Join(returned, r.stopWorkerRecord(fctx, identity, attempt, failed))
			if failed {
				_, e := r.workerEvent(fctx, identity, attempt, "worker_failed", "")
				returned = errors.Join(returned, e)
			}
			fcancel()
		}
		inputPipe.Close()
		output.Close()
		if !finishedProcess {
			proof := <-processDone
			result.ProcessExited = proof.ProcessExited
			result.GroupEmpty = proof.GroupEmpty
		}
		r.cancelActive(attempt)
		if launch.Sandbox != nil {
			if !sandboxHandled {
				cleanup, stop := context.WithTimeout(context.Background(), 10*time.Second)
				empty, cleanupErr := launch.Sandbox.Close(cleanup)
				stop()
				sandboxHandled = true
				sandboxEmpty = empty && cleanupErr == nil
				returned = errors.Join(returned, cleanupErr)
				if !sandboxEmpty {
					returned = errors.Join(returned, ErrWorker)
				}
			}
			if returned != nil {
				returned = errors.Join(returned, r.recordSandboxTermination(identity, attempt, result, sandboxEmpty))
			}
		}
		finish, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		if result.Finished && returned == nil {
			return
		}
		if result.ProcessExited && result.GroupEmpty {
			_, err := r.workerEvent(finish, identity, attempt, "worker_exited", "")
			returned = errors.Join(returned, err)
		} else {
			_, err := r.workerEvent(finish, identity, attempt, "worker_termination_unconfirmed", "")
			returned = errors.Join(returned, ErrWorker, err)
		}
	}()
	if _, err = r.workerEvent(workerCtx, identity, attempt, "worker_started", ""); err != nil {
		return result, err
	}
	watchStop := make(chan struct{})
	watchDone := make(chan struct{})
	go r.watchWorkerAuthority(workerCtx, cancel, identity, attempt, watchStop, watchDone)
	defer func() { close(watchStop); <-watchDone }()
	// Closing the pipe on cancellation also interrupts an oversized blocked write.
	pipeStop := make(chan struct{})
	defer close(pipeStop)
	go func() {
		select {
		case <-workerCtx.Done():
			inputPipe.Close()
		case <-pipeStop:
		}
	}()
	if err = json.NewEncoder(inputPipe).Encode(input); err != nil {
		return result, ErrWorker
	}
	scanner := bufio.NewScanner(output)
	scanner.Buffer(make([]byte, 4096), 262144)
	ready, finished := false, false
	for scanner.Scan() {
		if result.Frames >= 128 || finished {
			return result, ErrWorker
		}
		frame, err := parseWorkerFrame(scanner.Bytes(), result.Frames+1)
		if err != nil {
			return result, err
		}
		result.Frames++
		switch frame.Kind {
		case "ready":
			if ready || result.Frames != 1 {
				return result, ErrWorker
			}
			ready = true
			if _, err = r.workerEvent(workerCtx, identity, attempt, "worker_ready", ""); err != nil {
				return result, err
			}
			if err = ackWorker(inputPipe, frame.Sequence, "worker_ready"); err != nil {
				return result, ErrWorker
			}
		case "event":
			if !ready {
				return result, ErrWorker
			}
			if _, err = r.workerEvent(workerCtx, identity, attempt, sdkEventTypes[frame.EventType], frame.RequestID); err != nil {
				return result, err
			}
			if frame.EventType == "error" {
				return result, fmt.Errorf("%w: %s", ErrWorker, frame.ErrorCode)
			}
			if err = ackWorker(inputPipe, frame.Sequence, "sdk_event"); err != nil {
				return result, ErrWorker
			}
		case "tool":
			if !ready || frame.Call.AttemptID != attempt {
				return result, ErrWorker
			}
			if err = r.workerTool(workerCtx, identity, *frame.Call, handler, inputPipe, frame.Sequence); err != nil {
				return result, err
			}
		case "finished":
			if !ready {
				return result, ErrWorker
			}
			current, err := r.Read(workerCtx, identity, attempt)
			if err != nil || current.Control.phase != "active" || !hasControlEvent(current, "model_final_response") {
				return result, ErrWorker
			}
			if _, err = r.workerEvent(workerCtx, identity, attempt, "sdk_finished_observed", ""); err != nil {
				return result, err
			}
			if err = ackWorker(inputPipe, frame.Sequence, "worker_finished"); err != nil {
				return result, ErrWorker
			}
			finished = true
		case "failed":
			return result, fmt.Errorf("%w: %s", ErrWorker, frame.EventType)
		}
	}
	if scanner.Err() != nil || !finished {
		return result, ErrWorker
	}
	// A success frame cannot stand in for OS exit and an empty owned process group.
	select {
	case proof := <-processDone:
		result.ProcessExited = proof.ProcessExited
		result.GroupEmpty = proof.GroupEmpty
		finishedProcess = true
		if !proof.Finished {
			return result, ErrWorker
		}
	case <-time.After(2 * time.Second):
		return result, ErrWorker
	}
	if !result.ProcessExited || !result.GroupEmpty {
		return result, ErrWorker
	}
	if launch.Sandbox != nil {
		cleanup, cancelCleanup := context.WithTimeout(context.Background(), 10*time.Second)
		empty, cleanupErr := launch.Sandbox.Close(cleanup)
		cancelCleanup()
		sandboxHandled = true
		sandboxEmpty = empty && cleanupErr == nil
		if cleanupErr != nil || !empty {
			return result, errors.Join(ErrWorker, cleanupErr)
		}
		if _, err = r.workerEvent(workerCtx, identity, attempt, "sandbox_closed", ""); err != nil {
			return result, err
		}
	}
	if err = r.completeWorker(workerCtx, identity, attempt); err != nil {
		return result, err
	}
	result.Finished = true
	return result, nil
}

func (r *Runtime) workerTool(ctx context.Context, identity Identity, call ToolCall, handler ToolHandler, writer io.Writer, sequence int64) error {
	record, err := r.Read(ctx, identity, call.AttemptID)
	if err != nil {
		return err
	}
	offer, ok := record.Offers[call.RequestID]
	body, err := json.Marshal(call)
	if err != nil || !ok || offer.State != "queued" || offer.PayloadDigest != PayloadDigest(body) {
		return ErrWorker
	}
	if _, err = r.workerEvent(ctx, identity, call.AttemptID, "worker_waiting_tool", call.RequestID); err != nil {
		return err
	}
	if handler == nil {
		return ErrUnavailable
	}
	if _, err = r.DispatchTool(ctx, identity, call.AttemptID, call.RequestID); err != nil {
		return err
	}
	observation, err := handler(ctx, call)
	if err != nil {
		return errors.Join(ErrWorker, r.unknownWorkerTool(identity, call))
	}
	finish, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	if _, err = r.CompleteTool(finish, identity, call.AttemptID, call.RequestID, observation); err != nil {
		return errors.Join(err, r.unknownWorkerTool(identity, call))
	}
	return json.NewEncoder(writer).Encode(workerReply{"development_worker_reply.v1", sequence, call.RequestID, observation})
}
func workerEnvironment(dir string) []string {
	return []string{"PATH=/usr/local/bin:/usr/bin:/bin", "HOME=" + dir, "XDG_CACHE_HOME=" + dir, "PYTHONUNBUFFERED=1", "PYTHONDONTWRITEBYTECODE=1", "OPENHANDS_SUPPRESS_BANNER=1", "LITELLM_LOCAL_MODEL_COST_MAP=True", "HF_HUB_OFFLINE=1", "TRANSFORMERS_OFFLINE=1", "NO_PROXY=*", "OTEL_SDK_DISABLED=true"}
}
func (r *Runtime) watchWorkerAuthority(ctx context.Context, cancel context.CancelFunc, identity Identity, attempt string, stop <-chan struct{}, done chan<- struct{}) {
	defer close(done)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	started := r.now()
	var completedAt time.Time
	for {
		select {
		case <-ctx.Done():
			return
		case <-stop:
			return
		case <-ticker.C:
			var refusal error
			_, err := r.store.change(ctx, identity, attempt, 0, r.instanceID, func(record *Record) error {
				if record.Control.phase == "active" && hasControlEvent(record, "model_final_response") {
					refusal = r.workerFinalEligibility(ctx, record)
					if refusal != nil {
						return nil
					}
					if completedAt.IsZero() {
						completedAt = r.now()
					}
					if r.now().Sub(completedAt) > 5*time.Second {
						refusal = AttemptStopped
					}
					return nil
				}
				if record.Control.phase != "active" {
					refusal = AttemptStopped
					return nil
				}
				refusal = r.revalidate(ctx, record)
				if r.now().Sub(started) > 20*time.Second {
					ready := false
					for _, e := range record.Control.events {
						if e.Type == "worker_ready" {
							ready = true
							break
						}
					}
					if !ready {
						refusal = ErrWorker
					}
				}
				return nil
			})
			if err != nil || refusal != nil {
				cancel()
				return
			}
		}
	}
}
func watchWorkerProcess(ctx context.Context, command *exec.Cmd) WorkerResult {
	wait := make(chan error, 1)
	go func() { wait <- command.Wait() }()
	var waitErr error
	select {
	case waitErr = <-wait:
	case <-ctx.Done():
		_ = signalWorkerGroup(command.Process.Pid, false)
		select {
		case waitErr = <-wait:
		case <-time.After(2 * time.Second):
			_ = signalWorkerGroup(command.Process.Pid, true)
			select {
			case waitErr = <-wait:
			case <-time.After(5 * time.Second):
				return WorkerResult{}
			}
		}
	}
	// A normal parent exit can leave owned children behind; terminate them too.
	_ = signalWorkerGroup(command.Process.Pid, true)
	until := time.Now().Add(2 * time.Second)
	for !workerGroupEmpty(command.Process.Pid) && time.Now().Before(until) {
		reapWorkerGroup(command.Process.Pid)
		time.Sleep(10 * time.Millisecond)
	}
	_, failed := waitErr.(*exec.ExitError)
	return WorkerResult{ProcessExited: command.ProcessState != nil, GroupEmpty: workerGroupEmpty(command.Process.Pid), Finished: waitErr == nil && !failed}
}

func ackWorker(writer io.Writer, sequence int64, requestID string) error {
	return json.NewEncoder(writer).Encode(workerReply{"development_worker_reply.v1", sequence, requestID, ""})
}
