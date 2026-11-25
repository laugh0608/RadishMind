package developmentcontrol

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type taskSlotRecovery struct {
	store      *SQLiteStore
	instanceID string
	now        func() time.Time
}

func (c *TaskCoordinator) slotRecovery() taskSlotRecovery {
	return taskSlotRecovery{c.store, c.runtime.instanceID, c.runtime.now}
}

// RecoverWorkbenchTasks runs under the exclusive workbench lock before any
// executor is connected. All occupied slots at this point are previous owners.
func RecoverWorkbenchTasks(ctx context.Context, store *SQLiteStore, checkCode func(context.Context, TaskResourceOwner) (bool, error)) error {
	if store == nil {
		return ErrUnavailable
	}
	return (taskSlotRecovery{store: store, now: time.Now}).recoverTaskWorkers(ctx, checkCode)
}

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
	return c.slotRecovery().recoverPreviousSlot(ctx, check)
}
func (c taskSlotRecovery) recoverPreviousSlot(ctx context.Context, check func(context.Context, TaskResourceOwner) (TaskResourceReport, error)) error {
	if err := c.reconcileSlot(ctx); err != nil {
		return err
	}
	s, task, attempt, err := occupiedTask(ctx, c.store.db)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	r, err := c.store.Read(ctx, s.Identity, attempt)
	if err != nil {
		return err
	}
	if !belongsToTask(r, s, task, attempt) {
		return ErrTaskCorrupt
	}
	if r.OwnerInstanceID == c.instanceID {
		return nil
	}
	if !r.Revoked || check == nil {
		return ErrTaskTerminationUnconfirmed
	}
	t, err := readTask(ctx, c.store.db, s, task)
	if err != nil {
		return err
	}
	p, err := readTaskPlan(ctx, c.store.db, s, task, t.Revision)
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
	err = c.store.transaction(ctx, func(conn *sql.Conn) error {
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
			_, err := record.Control.ConfirmTermination(TerminationEvidence{attempt, d.ExecutorInstanceID, report.SDKExited, report.CodeEnvironmentExited, report.ModelConnectionClosed}, c.now().UnixMilli())
			return err
		})
		if err != nil {
			return err
		}
		// Persist partial observations in the same transaction as the slot decision.
		_, err = conn.ExecContext(ctx, `UPDATE development_task_resources SET
 sdk_worker=CASE WHEN sdk_worker='not_requested' THEN sdk_worker WHEN ? THEN 'confirmed' ELSE 'unconfirmed' END,
 code_environment=CASE WHEN code_environment='not_requested' THEN code_environment WHEN ? THEN 'confirmed' ELSE 'unconfirmed' END,
 model_connection=CASE WHEN model_connection='not_requested' THEN model_connection WHEN ? THEN 'confirmed' ELSE 'unconfirmed' END
 WHERE attempt_id=? AND owner_instance_id=?`, report.SDKExited, report.CodeEnvironmentExited, report.ModelConnectionClosed, attempt, owner.OwnerInstanceID)
		if err != nil {
			return ErrUnavailable
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
