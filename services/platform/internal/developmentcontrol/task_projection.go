package developmentcontrol

import (
	"context"
	"database/sql"
	"errors"
)

type TaskActiveTool struct {
	State     string `json:"state"`
	Sequence  int64  `json:"sequence,omitempty"`
	Name      string `json:"name,omitempty"`
	Candidate string `json:"candidate_digest,omitempty"`
	Started   int64  `json:"started_at_ms,omitempty"`
}

type TaskAttemptView struct {
	AttemptID       string             `json:"attempt_id"`
	Version         int64              `json:"record_version"`
	Phase           string             `json:"phase"`
	Termination     string             `json:"termination_status"`
	Usage           string             `json:"usage_status"`
	Budget          Budget             `json:"budget"`
	ResourceCheck   string             `json:"resource_check_status"`
	Resources       TaskResourceStatus `json:"resource_status"`
	Tool            TaskActiveTool     `json:"active_tool"`
	CancelRequested bool               `json:"-"`
}

// State, tool dispatch and resource proof are read at one database version.
func (m *TaskMetadata) AttemptView(ctx context.Context, s TaskScope, task, attempt string) (TaskAttemptView, error) {
	var v TaskAttemptView
	if !s.valid() || !idPattern.MatchString(task) || !idPattern.MatchString(attempt) {
		return v, ErrTaskInvalid
	}
	if err := m.host.AuthorizeTask(ctx, s, TaskRead); err != nil {
		return v, err
	}
	tx, err := m.store.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return v, ErrUnavailable
	}
	defer tx.Rollback()
	t, err := readTask(ctx, tx, s, task)
	if err != nil {
		return v, err
	}
	if t.AttemptID != attempt {
		return v, ErrNotFound
	}
	r, err := readRecord(ctx, tx, s.Identity, attempt)
	if err != nil {
		return v, err
	}
	if !belongsToTask(r, s, task, attempt) || len(r.Control.events) == 0 {
		return v, ErrTaskCorrupt
	}
	last := r.Control.events[len(r.Control.events)-1]
	v = TaskAttemptView{
		AttemptID: attempt, Version: r.Version, Phase: last.Phase,
		Termination: last.TerminationStatus, Usage: last.UsageStatus, Budget: last.Budget,
		ResourceCheck: last.TerminationStatus, Tool: TaskActiveTool{State: "idle"}, CancelRequested: hasControlEvent(r, "cancel_requested"),
	}
	p, err := readTaskResources(ctx, tx, s, attempt)
	switch {
	case errors.Is(err, ErrNotFound):
		v.Resources = TaskResourceStatus{"unconfirmed", "unconfirmed", "unconfirmed"}
		v.ResourceCheck = "unconfirmed"
	case err != nil:
		return v, err
	default:
		v.Resources = p.TaskResourceStatus
	}
	if err == nil && taskResourcesEnded(r) {
		v.ResourceCheck = "confirmed"
	}
	var request string
	var active TaskActiveTool
	err = tx.QueryRowContext(ctx, "SELECT request_id,sequence,tool_name,candidate_digest,started_at_ms FROM development_task_tool_activity WHERE attempt_id=?", attempt).Scan(&request, &active.Sequence, &active.Name, &active.Candidate, &active.Started)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return v, ErrUnavailable
	}
	if op := r.Control.operations[request]; op != nil && op.state == "dispatched" {
		active.State = "active"
		v.Tool = active
	}
	return v, tx.Commit()
}
