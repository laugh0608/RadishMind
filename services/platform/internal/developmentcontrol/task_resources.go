package developmentcontrol

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type TaskResourceStatus struct {
	ModelConnection string `json:"model_connection"`
	SDKWorker       string `json:"sdk_worker"`
	CodeEnvironment string `json:"code_environment"`
}

type taskResources struct {
	TaskResourceStatus
	hostPID, workerPGID int
	namespace           string
}

func readTaskResources(ctx context.Context, q queryer, s TaskScope, attempt string) (taskResources, error) {
	var p taskResources
	err := q.QueryRowContext(ctx, `SELECT m.model_connection,m.sdk_worker,m.code_environment,m.host_pid,m.worker_pgid,m.process_namespace
 FROM development_task_resources m JOIN development_tasks t USING(attempt_id) JOIN development_attempts a USING(attempt_id)
 WHERE t.tenant_ref=? AND t.workspace_id=? AND t.account_id=? AND t.application_id=? AND t.attempt_id=? AND m.owner_instance_id=a.owner_instance_id`, s.TenantRef, s.WorkspaceID, s.AccountID, s.ApplicationID, attempt).Scan(&p.ModelConnection, &p.SDKWorker, &p.CodeEnvironment, &p.hostPID, &p.workerPGID, &p.namespace)
	return p, taskReadError(err)
}

func (c *TaskMetadata) ReadTaskResources(ctx context.Context, s TaskScope, task, attempt string) (TaskResourceStatus, error) {
	t, err := c.ReadTask(ctx, s, task)
	if err != nil {
		return TaskResourceStatus{}, err
	}
	if t.AttemptID == "" || t.AttemptID != attempt {
		return TaskResourceStatus{}, ErrNotFound
	}
	p, err := readTaskResources(ctx, c.store.db, s, attempt)
	if errors.Is(err, ErrNotFound) {
		return TaskResourceStatus{"unconfirmed", "unconfirmed", "unconfirmed"}, nil
	}
	return p.TaskResourceStatus, err
}

// The column is selected by the trusted observer, never by SDK/HTTP input.
func (t *taskWorkerTracker) observe(ctx context.Context, column, state string, pgid int) error {
	return t.runtime.store.transaction(ctx, func(conn *sql.Conn) error { return t.writeObservation(ctx, conn, column, state, pgid) })
}

func (t *taskWorkerTracker) writeObservation(ctx context.Context, conn *sql.Conn, column, state string, pgid int) error {
	var update string
	switch column {
	case "model_connection":
		update = "model_connection=?"
	case "code_environment":
		update = "code_environment=?"
	case "sdk_worker":
		update = "sdk_worker=?"
	default:
		return ErrTaskInvalid
	}
	if state == "active" && column != "model_connection" {
		update = column + "=CASE WHEN " + column + "='stopping' THEN 'stopping' ELSE ? END"
	}
	args := []any{state}
	if pgid > 0 {
		update += ",worker_pgid=?"
		args = append(args, pgid)
	}
	res, err := conn.ExecContext(ctx, "UPDATE development_task_resources SET "+update+" WHERE attempt_id=? AND owner_instance_id=?", append(args, t.owner.Sandbox.AttemptID, t.owner.OwnerInstanceID)...)
	if err != nil {
		return ErrUnavailable
	}
	n, err := res.RowsAffected()
	if err != nil || n != 1 {
		return ErrUnavailable
	}
	return nil
}

func (t *taskWorkerTracker) recordSDKExit(ended bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	state := "unconfirmed"
	if ended {
		state = "confirmed"
	}
	return t.observe(ctx, "sdk_worker", state, 0)
}

func resourceEnded(state string) bool { return state == "not_requested" || state == "confirmed" }

func (r *Runtime) finishTaskResources(ctx context.Context, s TaskScope, attempt string, lateModel bool) error {
	return r.store.transaction(ctx, func(conn *sql.Conn) error {
		p, err := readTaskResources(ctx, conn, s, attempt)
		if err != nil {
			return err
		}
		record, err := changeRecord(ctx, conn, s.Identity, attempt, 0, r.instanceID, func(rec *Record) error {
			if taskResourcesEnded(rec) || !rec.Revoked || rec.Control.phase == "active" || lateModel && rec.Control.terminationStatus != "unconfirmed" {
				return nil
			}
			proof := TerminationEvidence{attempt, rec.Control.delegation.ExecutorInstanceID, resourceEnded(p.SDKWorker), resourceEnded(p.CodeEnvironment), resourceEnded(p.ModelConnection)}
			if rec.Control.terminationStatus == "unconfirmed" && !(proof.ExecutorExited && proof.SandboxExited && proof.ModelConnectionClosed) {
				return nil
			}
			_, err := rec.Control.ConfirmTermination(proof, r.now().UnixMilli())
			return err
		})
		if err != nil {
			return err
		}
		if !taskResourcesEnded(record) {
			return nil
		}
		return releaseTaskSlot(ctx, conn, record.Control.delegation.Scope.TaskID, attempt)
	})
}

// Called by the exclusive workbench host. Previous credentials are fenced
// before resource inspection. Missing pre-migration evidence stays unknown.
func (c *TaskCoordinator) RecoverTaskWorkers(ctx context.Context, checkCode func(context.Context, TaskResourceOwner) (bool, error)) error {
	return c.slotRecovery().recoverTaskWorkers(ctx, checkCode)
}
func (c taskSlotRecovery) recoverTaskWorkers(ctx context.Context, checkCode func(context.Context, TaskResourceOwner) (bool, error)) error {
	return c.recoverPreviousSlot(ctx, func(ctx context.Context, owner TaskResourceOwner) (TaskResourceReport, error) {
		p, err := readTaskResources(ctx, c.store.db, owner.Scope, owner.Sandbox.AttemptID)
		if errors.Is(err, ErrNotFound) {
			return TaskResourceReport{Owner: owner}, nil
		}
		if err != nil {
			return TaskResourceReport{}, err
		}
		sameNamespace := p.namespace != "" && p.namespace == processNamespace()
		report := TaskResourceReport{Owner: owner, SDKExited: resourceEnded(p.SDKWorker), CodeEnvironmentExited: resourceEnded(p.CodeEnvironment), ModelConnectionClosed: resourceEnded(p.ModelConnection)}
		if sameNamespace {
			if p.workerPGID > 0 && workerGroupEmpty(p.workerPGID) {
				report.SDKExited = true
			}
			if processAbsent(p.hostPID) {
				report.ModelConnectionClosed = true
			}
		}
		if !report.CodeEnvironmentExited && checkCode != nil {
			empty, err := checkCode(ctx, owner)
			report.CodeEnvironmentExited = empty && err == nil
		}
		return report, nil
	})
}
