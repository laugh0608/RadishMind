package developmentcontrol

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

func taskReadError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return ErrUnavailable
}

func readTask(ctx context.Context, q queryer, scope TaskScope, id string) (Task, error) {
	t := Task{TaskID: id, SourceProfileID: taskSource}
	err := q.QueryRowContext(ctx, `SELECT record_version,created_at_ms,updated_at_ms,plan_revision,coalesce(attempt_id,'') FROM development_tasks
 WHERE tenant_ref=? AND workspace_id=? AND account_id=? AND application_id=? AND task_id=?`, scope.TenantRef, scope.WorkspaceID, scope.AccountID, scope.ApplicationID, id).Scan(&t.Version, &t.CreatedAtMS, &t.UpdatedAtMS, &t.Revision, &t.AttemptID)
	return t, taskReadError(err)
}

func readTaskPlan(ctx context.Context, q queryer, scope TaskScope, task string, revision int) (TaskPlan, error) {
	p := TaskPlan{Revision: revision}
	var body string
	err := q.QueryRowContext(ctx, `SELECT p.profile_record,p.plan_digest FROM development_task_plans p JOIN development_tasks t USING(task_id)
 WHERE t.tenant_ref=? AND t.workspace_id=? AND t.account_id=? AND t.application_id=? AND p.task_id=? AND p.plan_revision=?`, scope.TenantRef, scope.WorkspaceID, scope.AccountID, scope.ApplicationID, task, revision).Scan(&body, &p.Digest)
	if err != nil {
		return TaskPlan{}, taskReadError(err)
	}
	if PayloadDigest([]byte(body)) != p.Digest || json.Unmarshal([]byte(body), &p.TaskProfile) != nil || !p.valid(scope, TaskChoice{p.AttemptPlan.ModelRef, p.AttemptPlan.Limits}) {
		return TaskPlan{}, ErrTaskCorrupt
	}
	return p, nil
}

func readTaskReceipt(ctx context.Context, q queryer, scope TaskScope, request string) (TaskReceipt, error) {
	r := TaskReceipt{RequestID: request}
	err := q.QueryRowContext(ctx, `SELECT operation,request_digest,task_id,plan_revision,coalesce(preparation_id,''),coalesce(attempt_id,''),coalesce(artifact_id,'')
 FROM development_task_receipts WHERE tenant_ref=? AND workspace_id=? AND account_id=? AND application_id=? AND request_id=?`, scope.TenantRef, scope.WorkspaceID, scope.AccountID, scope.ApplicationID, request).Scan(&r.Operation, &r.RequestDigest, &r.TaskID, &r.Revision, &r.PreparationID, &r.AttemptID, &r.ArtifactID)
	return r, taskReadError(err)
}

func (c *TaskCoordinator) ReadTask(ctx context.Context, scope TaskScope, id string) (Task, error) {
	if !scope.valid() || !idPattern.MatchString(id) {
		return Task{}, ErrTaskInvalid
	}
	if err := c.host.AuthorizeTask(ctx, scope, TaskRead); err != nil {
		return Task{}, err
	}
	return readTask(ctx, c.runtime.store.db, scope, id)
}

// ReadAttempt returns a private control record for projection by the host.
func (c *TaskCoordinator) ReadAttempt(ctx context.Context, s TaskScope, task, attempt string) (*Record, error) {
	if !idPattern.MatchString(attempt) {
		return nil, ErrTaskInvalid
	}
	t, err := c.ReadTask(ctx, s, task)
	if err != nil {
		return nil, err
	}
	if t.AttemptID != attempt {
		return nil, ErrNotFound
	}
	record, err := c.runtime.Read(ctx, s.Identity, attempt)
	if err != nil {
		return nil, err
	}
	if !belongsToTask(record, s, task, attempt) {
		return nil, ErrTaskCorrupt
	}
	return record, nil
}

func (c *TaskCoordinator) ReadPlan(ctx context.Context, scope TaskScope, task string, revision int) (TaskPlan, error) {
	if !scope.valid() || !idPattern.MatchString(task) || revision < 1 || revision > 16 {
		return TaskPlan{}, ErrTaskInvalid
	}
	if err := c.host.AuthorizeTask(ctx, scope, TaskRead); err != nil {
		return TaskPlan{}, err
	}
	return readTaskPlan(ctx, c.runtime.store.db, scope, task, revision)
}

func (c *TaskCoordinator) ReadReceipt(ctx context.Context, scope TaskScope, request string) (TaskReceipt, error) {
	if !scope.valid() || !idPattern.MatchString(request) {
		return TaskReceipt{}, ErrTaskInvalid
	}
	r, err := readTaskReceipt(ctx, c.runtime.store.db, scope, request)
	if err != nil {
		return TaskReceipt{}, err
	}
	op := TaskRead
	if r.Operation == TaskSaveArtifact {
		op = TaskReadArtifact
	}
	if err := c.host.AuthorizeTask(ctx, scope, op); err != nil {
		if errors.Is(err, ErrIneligible) {
			return TaskReceipt{}, ErrNotFound
		}
		return TaskReceipt{}, err
	}
	return r, nil
}

// A write reads the current owner and operation permission before consulting a
// receipt. Only first submissions evaluate time-sensitive prerequisites.
func (c *TaskCoordinator) write(ctx context.Context, s TaskScope, op TaskOperation, task, request string, input any, apply func(*sql.Conn, Task, *TaskReceipt) error) (TaskReceipt, error) {
	if !s.valid() || !idPattern.MatchString(request) || op != TaskCreate && !idPattern.MatchString(task) {
		return TaskReceipt{}, ErrTaskInvalid
	}
	digest, err := taskDigest(struct {
		TaskID string `json:"task_id"`
		Input  any    `json:"input"`
	}{task, input})
	if err != nil {
		return TaskReceipt{}, err
	}
	r := TaskReceipt{RequestID: request, Operation: op, RequestDigest: digest, TaskID: task}
	err = c.runtime.store.transaction(ctx, func(conn *sql.Conn) error {
		if err := c.host.AuthorizeTask(ctx, s, op); err != nil {
			return err
		}
		var current Task
		if op != TaskCreate {
			var err error
			current, err = readTask(ctx, conn, s, task)
			if err != nil {
				return err
			}
			r.Revision = current.Revision
		}
		old, err := readTaskReceipt(ctx, conn, s, request)
		if err == nil {
			if old.Operation != op || old.RequestDigest != digest {
				return ErrConflict
			}
			r = old
			return nil
		}
		if !errors.Is(err, ErrNotFound) {
			return err
		}
		if op != TaskCreate {
			var count int
			if err := conn.QueryRowContext(ctx, `SELECT count(*) FROM development_task_receipts WHERE task_id=?`, task).Scan(&count); err != nil {
				return ErrUnavailable
			}
			// A committed start must leave room for the owner to cancel.
			if count >= 64 || op == TaskStart && count >= 63 {
				return ErrTaskCapacity
			}
		}
		if err := apply(conn, current, &r); err != nil {
			return err
		}
		_, err = conn.ExecContext(ctx, `INSERT INTO development_task_receipts(tenant_ref,workspace_id,account_id,application_id,request_id,operation,request_digest,task_id,plan_revision,preparation_id,attempt_id,artifact_id)
 VALUES(?,?,?,?,?,?,?,?,?,nullif(?,''),nullif(?,''),nullif(?,''))`, s.TenantRef, s.WorkspaceID, s.AccountID, s.ApplicationID, request, op, digest, r.TaskID, r.Revision, r.PreparationID, r.AttemptID, r.ArtifactID)
		if err != nil {
			return ErrUnavailable
		}
		return nil
	})
	if err != nil {
		return TaskReceipt{}, err
	}
	return r, nil
}
