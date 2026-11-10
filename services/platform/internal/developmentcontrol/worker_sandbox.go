package developmentcontrol

import (
	"context"
	"time"
)

// WorkerSandbox is an explicitly bound trusted executor. The SDK cannot supply
// an implementation or assert its cleanup result.
type WorkerSandbox interface {
	Binding() SandboxBinding
	HandleTool(context.Context, ToolCall) (string, error)
	Close(context.Context) (bool, error)
}
type SandboxBinding struct {
	AttemptID          string `json:"attempt_id"`
	ExecutorInstanceID string `json:"executor_instance_id"`
	PlanDigest         string `json:"plan_digest"`
	BaselineDigest     string `json:"baseline_digest"`
}

func SandboxBindingFor(d Delegation) SandboxBinding {
	return SandboxBinding{d.Scope.AttemptID, d.ExecutorInstanceID, d.Scope.PlanDigest, d.Scope.BaselineDigest}
}
func (r *Runtime) recordSandboxTermination(identity Identity, attempt string, proof WorkerResult, sandboxEmpty bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Wait for the owned model call to leave, not just a cancel function to run.
	until := time.Now().Add(time.Second)
	closed := false
	for {
		r.mu.Lock()
		closed = r.active[attempt] == nil
		r.mu.Unlock()
		if closed || time.Now().After(until) {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
	_, err := r.store.change(ctx, identity, attempt, 0, r.instanceID, func(record *Record) error {
		_, err := record.Control.ConfirmTermination(TerminationEvidence{attempt, record.Control.delegation.ExecutorInstanceID, proof.ProcessExited && proof.GroupEmpty, sandboxEmpty, closed}, r.now().UnixMilli())
		return err
	})
	return err
}
