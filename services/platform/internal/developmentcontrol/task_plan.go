package developmentcontrol

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

func (c *TaskMetadata) storePlan(ctx context.Context, conn *sql.Conn, s TaskScope, choice TaskChoice, r *TaskReceipt) error {
	p, _, err := c.host.FixedProfile(ctx, s, choice)
	if err != nil {
		return err
	}
	if !p.valid(s, choice) {
		return ErrIneligible
	}
	body, err := json.Marshal(p)
	if err != nil {
		return ErrTaskInvalid
	}
	_, err = conn.ExecContext(ctx, `INSERT INTO development_task_plans(task_id,plan_revision,profile_record,plan_digest) VALUES(?,?,?,?)`, r.TaskID, r.Revision, string(body), PayloadDigest(body))
	if err != nil {
		return ErrUnavailable
	}
	return nil
}

func (c *TaskMetadata) Create(ctx context.Context, s TaskScope, request, source string, choice TaskChoice) (TaskReceipt, error) {
	if source != taskSource || !choice.valid() {
		return TaskReceipt{}, ErrTaskInvalid
	}
	input := struct {
		Source string `json:"source_profile_id"`
		TaskChoice
	}{source, choice}
	return c.write(ctx, s, TaskCreate, "", request, input, func(conn *sql.Conn, _ Task, r *TaskReceipt) error {
		available, err := hasTaskCapacity(ctx, conn, s)
		if err != nil {
			return err
		}
		if !available {
			return ErrTaskCapacity
		}
		id, err := randomID("task_")
		if err != nil {
			return err
		}
		r.TaskID, r.Revision = id, 1
		now := c.now().UnixMilli()
		_, err = conn.ExecContext(ctx, `INSERT INTO development_tasks(task_id,tenant_ref,workspace_id,account_id,application_id,record_version,plan_revision,created_at_ms,updated_at_ms) VALUES(?,?,?,?,?,1,1,?,?)`, id, s.TenantRef, s.WorkspaceID, s.AccountID, s.ApplicationID, now, now)
		if err != nil {
			return ErrUnavailable
		}
		return c.storePlan(ctx, conn, s, choice, r)
	})
}

func (c *TaskMetadata) Revise(ctx context.Context, s TaskScope, task, request string, version int64, choice TaskChoice) (TaskReceipt, error) {
	if !validTaskVersion(version) || !choice.valid() {
		return TaskReceipt{}, ErrTaskInvalid
	}
	input := struct {
		Version int64 `json:"expected_version"`
		TaskChoice
	}{version, choice}
	return c.write(ctx, s, TaskRevise, task, request, input, func(conn *sql.Conn, t Task, r *TaskReceipt) error {
		if t.Version != version || t.AttemptID != "" {
			return ErrConflict
		}
		if t.Revision >= 16 {
			return ErrTaskCapacity
		}
		r.Revision = t.Revision + 1
		if err := c.storePlan(ctx, conn, s, choice, r); err != nil {
			return err
		}
		_, err := conn.ExecContext(ctx, `UPDATE development_tasks SET plan_revision=?,record_version=record_version+1,updated_at_ms=max(updated_at_ms,?) WHERE task_id=?`, r.Revision, c.now().UnixMilli(), task)
		if err != nil {
			return ErrUnavailable
		}
		return nil
	})
}

func validTaskVersion(n int64) bool { return n >= 1 && n <= 9007199254740991 }

// Resolve the frozen profile again. A changed image, retained-content policy,
// selection or input reservation ceiling invalidates an outstanding confirmation.
func (c *TaskCoordinator) qualification(ctx context.Context, s TaskScope, p TaskPlan) (Binding, string, int64, error) {
	choice := TaskChoice{p.AttemptPlan.ModelRef, p.AttemptPlan.Limits}
	current, expires, err := c.host.FixedProfile(ctx, s, choice)
	if err != nil {
		return Binding{}, "", 0, err
	}
	digest, err := taskDigest(current)
	if err != nil || !current.valid(s, choice) || digest != p.Digest {
		return Binding{}, "", 0, ErrConflict
	}
	binding, err := c.runtime.currentBinding(ctx, s.Identity, p.AttemptPlan)
	if err != nil {
		return Binding{}, "", 0, err
	}
	eligible, err := taskDigest(binding)
	return binding, eligible, expires.UnixMilli(), err
}

func (p TaskPreparation) digest(s TaskScope) (string, error) {
	p.ConfirmationDigest = ""
	return taskDigest(struct {
		Scope   TaskScope       `json:"scope"`
		Preview TaskPreparation `json:"preview"`
	}{s, p})
}

func (c *TaskCoordinator) Prepare(ctx context.Context, s TaskScope, task, request string, version int64, planDigest string) (TaskReceipt, error) {
	if !validTaskVersion(version) || !digestPattern.MatchString(planDigest) {
		return TaskReceipt{}, ErrTaskInvalid
	}
	input := struct {
		Version int64  `json:"expected_version"`
		Digest  string `json:"plan_digest"`
	}{version, planDigest}
	return c.write(ctx, s, TaskPrepare, task, request, input, func(conn *sql.Conn, t Task, r *TaskReceipt) error {
		if t.Version != version || t.AttemptID != "" {
			return ErrConflict
		}
		p, err := readTaskPlan(ctx, conn, s, task, t.Revision)
		if err != nil {
			return err
		}
		if p.Digest != planDigest {
			return ErrConflict
		}
		_, eligible, validUntil, err := c.qualification(ctx, s, p)
		if err != nil {
			return err
		}
		now := c.runtime.now()
		expires := min(validUntil, now.Add(5*time.Minute).UnixMilli())
		if expires <= now.UnixMilli() {
			return ErrTaskExpired
		}
		id, err := randomID("prep_")
		if err != nil {
			return err
		}
		preview := TaskPreparation{PreparationID: id, TaskID: task, TaskVersion: version, Revision: t.Revision, PreparedAtMS: now.UnixMilli(), ExpiresAtMS: expires, Plan: p, EligibilityDigest: eligible}
		confirmation, err := preview.digest(s)
		if err != nil {
			return err
		}
		r.PreparationID = id
		_, err = conn.ExecContext(ctx, `INSERT INTO development_task_preparations(preparation_id,task_id,task_version,plan_revision,prepared_at_ms,expires_at_ms,eligibility_digest,confirmation_digest) VALUES(?,?,?,?,?,?,?,?)`, id, task, version, t.Revision, preview.PreparedAtMS, expires, eligible, confirmation)
		if err != nil {
			return ErrUnavailable
		}
		return nil
	})
}

func readPreparation(ctx context.Context, q queryer, s TaskScope, task, id string) (TaskPreparation, bool, error) {
	p := TaskPreparation{PreparationID: id, TaskID: task}
	var consumed bool
	err := q.QueryRowContext(ctx, `SELECT p.task_version,p.plan_revision,p.prepared_at_ms,p.expires_at_ms,p.eligibility_digest,p.confirmation_digest,p.consumed
 FROM development_task_preparations p JOIN development_tasks t USING(task_id)
 WHERE t.tenant_ref=? AND t.workspace_id=? AND t.account_id=? AND t.application_id=? AND p.task_id=? AND p.preparation_id=?`, s.TenantRef, s.WorkspaceID, s.AccountID, s.ApplicationID, task, id).Scan(&p.TaskVersion, &p.Revision, &p.PreparedAtMS, &p.ExpiresAtMS, &p.EligibilityDigest, &p.ConfirmationDigest, &consumed)
	if err != nil {
		return TaskPreparation{}, false, taskReadError(err)
	}
	p.Plan, err = readTaskPlan(ctx, q, s, task, p.Revision)
	if err != nil {
		return TaskPreparation{}, false, err
	}
	digest, err := p.digest(s)
	if err != nil || digest != p.ConfirmationDigest {
		return TaskPreparation{}, false, ErrTaskCorrupt
	}
	return p, consumed, nil
}

func (c *TaskCoordinator) ReadPreparation(ctx context.Context, s TaskScope, task, id string) (TaskPreparation, error) {
	if !s.valid() || !idPattern.MatchString(id) || !idPattern.MatchString(task) {
		return TaskPreparation{}, ErrTaskInvalid
	}
	if err := c.host.AuthorizeTask(ctx, s, TaskRead); err != nil {
		return TaskPreparation{}, err
	}
	p, _, err := readPreparation(ctx, c.runtime.store.db, s, task, id)
	return p, err
}
