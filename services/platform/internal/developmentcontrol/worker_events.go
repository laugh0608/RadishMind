package developmentcontrol

import (
	"context"
	"strings"
	"time"
)

func (r *Runtime) workerEvent(ctx context.Context, identity Identity, attemptID, kind, requestID string) (*Record, error) {
	return r.store.change(ctx, identity, attemptID, 0, r.instanceID, func(record *Record) error {
		if len(record.Control.events) >= 200 || len(record.Control.events) >= 180 && strings.HasPrefix(kind, "sdk_") {
			return ErrWorker
		}
		if requestID != "" {
			if _, exists := record.Control.operations[requestID]; !exists {
				return ErrWorker
			}
		}
		record.Control.emit(kind, requestID, r.now().UnixMilli())
		return nil
	})
}
func (r *Runtime) stopWorkerRecord(ctx context.Context, identity Identity, attemptID string, failed bool) error {
	_, err := r.store.change(ctx, identity, attemptID, 0, r.instanceID, func(record *Record) error {
		if record.Control.phase == "active" || record.Control.phase == "cancel_requested" {
			if failed && record.Control.phase == "active" {
				if _, err := record.Control.Fail("sdk_worker_failed", r.now().UnixMilli()); err != nil {
					return err
				}
			} else {
				if _, err := record.Control.Interrupt(r.now().UnixMilli()); err != nil {
					return err
				}
			}
		}
		record.Revoked = true
		return nil
	})
	return err
}

func hasControlEvent(record *Record, kind string) bool {
	for _, event := range record.Control.events {
		if event.Type == kind {
			return true
		}
	}
	return false
}
func (r *Runtime) workerFinalEligibility(ctx context.Context, record *Record) error {
	selected, err := r.reader.Resolve(ctx, record.Binding.Identity, record.Binding.Plan)
	if err != nil {
		return err
	}
	d := record.Control.delegation
	deadline, live := r.attemptDeadline(d.Scope.AttemptID)
	now := r.now()
	if selected != record.Binding.Selection || !live || !now.Before(deadline) || now.UnixMilli() < d.IssuedAtMS || now.UnixMilli() >= d.ExpiresAtMS {
		return ErrIneligible
	}
	return nil
}
func (r *Runtime) completeWorker(ctx context.Context, identity Identity, attemptID string) error {
	_, err := r.store.change(ctx, identity, attemptID, 0, r.instanceID, func(record *Record) error {
		if record.Control.phase != "active" || !hasControlEvent(record, "model_final_response") {
			return ErrWorker
		}
		if err := r.workerFinalEligibility(ctx, record); err != nil {
			return err
		}
		record.Control.emit("worker_exited", "", r.now().UnixMilli())
		_, err := record.Control.Complete(r.now().UnixMilli())
		return err
	})
	return err
}

func (r *Runtime) unknownWorkerTool(identity Identity, call ToolCall) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.store.change(ctx, identity, call.AttemptID, 0, r.instanceID, func(record *Record) error {
		op := record.Control.operations[call.RequestID]
		if op == nil {
			return ErrWorker
		}
		if op.state == "dispatched" {
			if _, err := record.Control.RecordReceipt(Receipt{"development_attempt_receipt.v1", call.AttemptID, call.RequestID, "unknown", 0, 0}, r.now().UnixMilli()); err != nil {
				return err
			}
		}
		record.Revoked = true
		return nil
	})
	return err
}
