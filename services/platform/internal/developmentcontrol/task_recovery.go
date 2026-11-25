package developmentcontrol

import (
	"context"
	"database/sql"
	"errors"
)

// TaskResourceOwner is derived solely from the occupied slot, the original
// delegation and its immutable plan. A host checker must match these identities
// to its own resource ledger before touching any SDK process or code container.
type TaskResourceOwner struct {
	Scope           TaskScope
	TaskID          string
	OwnerInstanceID string
	Sandbox         SandboxBinding
	SDKImageDigest  string
	GoImageDigest   string
	ExecutorDigest  string
}

type TaskResourceReport struct {
	Owner                 TaskResourceOwner
	SDKExited             bool
	CodeEnvironmentExited bool
	ModelConnectionClosed bool
}

var ErrTaskTerminationUnconfirmed = errors.New("development_task_resource_termination_unconfirmed")

func occupiedTask(ctx context.Context, q queryer) (TaskScope, string, string, error) {
	var s TaskScope
	var task, attempt string
	err := q.QueryRowContext(ctx, `SELECT tenant_ref,workspace_id,account_id,application_id,task_id,attempt_id FROM development_tasks WHERE execution_slot=1`).Scan(&s.TenantRef, &s.WorkspaceID, &s.AccountID, &s.ApplicationID, &task, &attempt)
	return s, task, attempt, taskReadError(err)
}

// RecoverPreviousSlot runs only for a host holding the workbench process lock.
// External cleanup is outside the writer transaction. Late or mismatched proof
// cannot release a different slot, and recovery never issues a credential.
func (c *TaskCoordinator) RecoverPreviousSlot(ctx context.Context, check func(context.Context, TaskResourceOwner) (TaskResourceReport, error)) error {
	if err := c.ReconcileSlot(ctx); err != nil {
		return err
	}
	s, task, attempt, err := occupiedTask(ctx, c.runtime.store.db)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	r, err := c.runtime.Read(ctx, s.Identity, attempt)
	if err != nil {
		return err
	}
	if !belongsToTask(r, s, task, attempt) {
		return ErrTaskCorrupt
	}
	if r.OwnerInstanceID == c.runtime.instanceID {
		return nil
	}
	if !r.Revoked || check == nil {
		return ErrTaskTerminationUnconfirmed
	}
	t, err := readTask(ctx, c.runtime.store.db, s, task)
	if err != nil {
		return err
	}
	p, err := readTaskPlan(ctx, c.runtime.store.db, s, task, t.Revision)
	if err != nil {
		return err
	}
	d := r.Control.Delegation()
	pd, err := p.AttemptPlan.Digest()
	if err != nil || pd != d.Scope.PlanDigest || t.AttemptID != attempt {
		return ErrTaskCorrupt
	}
	owner := TaskResourceOwner{s, task, r.OwnerInstanceID, SandboxBindingFor(d), p.SDKImageDigest, p.GoImageDigest, p.ExecutorDigest}
	report, err := check(ctx, owner)
	if err != nil {
		return err
	}
	if report.Owner != owner {
		return ErrTaskTerminationUnconfirmed
	}
	ended := report.SDKExited && report.CodeEnvironmentExited && report.ModelConnectionClosed
	err = c.runtime.store.transaction(ctx, func(conn *sql.Conn) error {
		currentScope, currentTask, currentAttempt, err := occupiedTask(ctx, conn)
		if err != nil {
			return err
		}
		if currentScope != s || currentTask != task || currentAttempt != attempt {
			return ErrConflict
		}
		current, err := changeRecord(ctx, conn, s.Identity, attempt, 0, owner.OwnerInstanceID, func(record *Record) error {
			if !belongsToTask(record, s, task, attempt) || record.Control.Delegation() != d || !record.Revoked {
				return ErrTaskCorrupt
			}
			if taskResourcesEnded(record) || (!ended && record.Control.terminationStatus == "unconfirmed") {
				return nil
			}
			_, err := record.Control.ConfirmTermination(TerminationEvidence{attempt, d.ExecutorInstanceID, report.SDKExited, report.CodeEnvironmentExited, report.ModelConnectionClosed}, c.runtime.now().UnixMilli())
			return err
		})
		if err != nil {
			return err
		}
		if taskResourcesEnded(current) {
			return releaseTaskSlot(ctx, conn, task, attempt)
		}
		return nil
	})
	if err == nil && !ended {
		return ErrTaskTerminationUnconfirmed
	}
	return err
}

func releaseTaskSlot(ctx context.Context, conn *sql.Conn, task, attempt string) error {
	_, err := conn.ExecContext(ctx, `UPDATE development_tasks SET execution_slot=NULL WHERE task_id=? AND attempt_id=? AND execution_slot=1`, task, attempt)
	if err != nil {
		return ErrUnavailable
	}
	return nil
}
