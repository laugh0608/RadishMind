package developmentcontrol

import (
	"context"
	"database/sql"
	"errors"
	"os"
)

// Start returns a credential exclusively to the trusted host on first commit.
// Replaying an authorized request returns the same receipt and a nil issuance.
func (c *TaskCoordinator) Start(ctx context.Context, s TaskScope, task, request string, version int64, preparation, confirmation string) (TaskReceipt, *Issued, error) {
	return c.start(ctx, s, task, request, version, preparation, confirmation, false)
}

func (c *TaskCoordinator) start(ctx context.Context, s TaskScope, task, request string, version int64, preparation, confirmation string, managed bool) (TaskReceipt, *Issued, error) {
	if !validTaskVersion(version) || !idPattern.MatchString(preparation) || !digestPattern.MatchString(confirmation) {
		return TaskReceipt{}, nil, ErrTaskInvalid
	}
	input := struct {
		Version     int64  `json:"expected_version"`
		Preparation string `json:"preparation_id"`
		Digest      string `json:"confirmation_digest"`
	}{version, preparation, confirmation}
	issuing := false
	defer func() {
		if issuing {
			c.runtime.work.Done()
		}
	}()
	var credential *Issued
	receipt, err := c.write(ctx, s, TaskStart, task, request, input, func(conn *sql.Conn, t Task, r *TaskReceipt) error {
		if t.Version != version || t.AttemptID != "" {
			return ErrConflict
		}
		if err := c.runtime.beginIssue(); err != nil {
			return err
		}
		issuing = true
		p, consumed, err := readPreparation(ctx, conn, s, task, preparation)
		if err != nil {
			return err
		}
		if consumed || p.TaskVersion != version || p.Revision != t.Revision || p.ConfirmationDigest != confirmation {
			return ErrConflict
		}
		if c.runtime.now().UnixMilli() >= p.ExpiresAtMS {
			return ErrTaskExpired
		}
		b, eligible, validUntil, err := c.qualification(ctx, s, p.Plan)
		if err != nil {
			return err
		}
		if eligible != p.EligibilityDigest || validUntil < p.ExpiresAtMS {
			return ErrConflict
		}
		var busy bool
		if err := conn.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM development_tasks WHERE execution_slot=1)`).Scan(&busy); err != nil {
			return ErrUnavailable
		}
		if busy {
			return ErrSlotBusy
		}
		now := c.runtime.now().UnixMilli()
		if now < p.PreparedAtMS || now >= p.ExpiresAtMS {
			return ErrTaskExpired
		}
		issued, err := c.runtime.issueInTransaction(ctx, conn, b, task, preparation)
		if err != nil {
			return err
		}
		r.AttemptID = issued.Delegation.Scope.AttemptID
		_, err = conn.ExecContext(ctx, `UPDATE development_task_preparations SET consumed=1 WHERE preparation_id=?`, preparation)
		if err != nil {
			return ErrUnavailable
		}
		_, err = conn.ExecContext(ctx, `UPDATE development_tasks SET execution_slot=1,attempt_id=?,record_version=record_version+1,updated_at_ms=max(updated_at_ms,?) WHERE task_id=?`, r.AttemptID, now, task)
		if err != nil {
			return ErrUnavailable
		}
		if managed {
			if _, err := conn.ExecContext(ctx, `INSERT INTO development_task_resources(attempt_id,owner_instance_id,host_pid,process_namespace) VALUES(?,?,?,?)`, r.AttemptID, c.runtime.instanceID, os.Getpid(), processNamespace()); err != nil {
				return ErrUnavailable
			}
		}
		credential = &issued
		return nil
	})
	if err != nil {
		return TaskReceipt{}, nil, err
	}
	if credential != nil {
		c.runtime.registerIssue(*credential)
	}
	return receipt, credential, nil
}

func (c *TaskCoordinator) Cancel(ctx context.Context, s TaskScope, task, attempt, request string, version int64) (TaskReceipt, error) {
	if !validTaskVersion(version) || !idPattern.MatchString(attempt) {
		return TaskReceipt{}, ErrTaskInvalid
	}
	input := struct {
		Attempt string `json:"attempt_id"`
		Version int64  `json:"expected_version"`
	}{attempt, version}
	cancelled := false
	receipt, err := c.write(ctx, s, TaskCancel, task, request, input, func(conn *sql.Conn, t Task, r *TaskReceipt) error {
		if t.AttemptID != attempt {
			return ErrNotFound
		}
		_, err := changeRecord(ctx, conn, s.Identity, attempt, version, "", func(record *Record) error {
			if !belongsToTask(record, s, task, attempt) {
				return ErrTaskCorrupt
			}
			if record.Control.phase != "active" {
				return nil
			}
			if _, err := record.Control.RequestCancel("user_cancelled", c.runtime.now().UnixMilli()); err != nil {
				return err
			}
			record.Revoked = true
			cancelled = true
			return nil
		})
		if err != nil {
			return err
		}
		if cancelled {
			_, err = conn.ExecContext(ctx, `UPDATE development_task_resources SET
 sdk_worker=CASE WHEN sdk_worker IN ('pending','active') THEN 'stopping' ELSE sdk_worker END,
 code_environment=CASE WHEN code_environment IN ('pending','active') THEN 'stopping' ELSE code_environment END,
 model_connection=CASE WHEN model_connection='active' THEN 'stopping' ELSE model_connection END
 WHERE attempt_id=?`, attempt)
			if err != nil {
				return ErrUnavailable
			}
		}
		r.AttemptID = attempt
		return nil
	})
	if err == nil && cancelled {
		c.runtime.cancelActive(attempt)
	}
	return receipt, err
}

func belongsToTask(record *Record, s TaskScope, task, attempt string) bool {
	d := record.Control.Delegation().Scope
	return record.Binding.Identity == s.Identity && d.ApplicationID == s.ApplicationID && d.RunID == task && d.TaskID == task && d.AttemptID == attempt
}

// ReconcileSlot is for the exclusive workbench host at startup or worker exit.
// It fences an old active owner without starting any work. Only the original
// control evidence can free the slot; missing cleanup evidence keeps it held.
func (c *TaskCoordinator) ReconcileSlot(ctx context.Context) error {
	return c.slotRecovery().reconcileSlot(ctx)
}
func (c taskSlotRecovery) reconcileSlot(ctx context.Context) error {
	return c.store.transaction(ctx, func(conn *sql.Conn) error {
		s, task, attempt, err := occupiedTask(ctx, conn)
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		record, err := changeRecord(ctx, conn, s.Identity, attempt, 0, "", func(r *Record) error {
			if !belongsToTask(r, s, task, attempt) {
				return ErrTaskCorrupt
			}
			if r.OwnerInstanceID == c.instanceID {
				return nil
			}
			r.Revoked = true
			if r.Control.phase == "active" || r.Control.phase == "cancel_requested" {
				_, err := r.Control.Interrupt(c.now().UnixMilli())
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
		if !taskResourcesEnded(record) {
			return nil
		}
		return releaseTaskSlot(ctx, conn, task, attempt)
	})
}
