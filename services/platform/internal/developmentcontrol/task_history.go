package developmentcontrol

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"slices"

	"radishmind.local/services/platform/internal/strictjson"
)

type TaskPage struct {
	Tasks      []Task `json:"tasks"`
	NextCursor string `json:"next_cursor"`
	HasMore    bool   `json:"has_more"`
}

// The cursor carries no authority. Each page reads under the current scope.
type HistoryPosition struct {
	Binding     string `json:"binding"`
	CreatedAtMS int64  `json:"created_at_ms"`
	ID          string `json:"id"`
}

func historyBinding(scope TaskScope, kind string) (string, error) {
	return taskDigest(struct {
		Scope TaskScope `json:"scope"`
		Kind  string    `json:"kind"`
	}{scope, kind})
}
func EncodeHistoryPosition(scope TaskScope, kind string, ms int64, id string) (string, error) {
	binding, err := historyBinding(scope, kind)
	if err != nil {
		return "", err
	}
	raw, err := json.Marshal(HistoryPosition{binding, ms, id})
	if err != nil {
		return "", ErrTaskInvalid
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
func DecodeHistoryPosition(scope TaskScope, kind, encoded string) (HistoryPosition, error) {
	var p HistoryPosition
	if len(encoded) == 0 || len(encoded) > 1024 {
		return p, ErrTaskInvalid
	}
	raw, err := base64.RawURLEncoding.Strict().DecodeString(encoded)
	if err != nil || strictjson.DecodeBounded(raw, &p, 768) != nil || !idPattern.MatchString(p.ID) || !validTaskVersion(p.CreatedAtMS) {
		return HistoryPosition{}, ErrTaskInvalid
	}
	expected, err := EncodeHistoryPosition(scope, kind, p.CreatedAtMS, p.ID)
	if err != nil || expected != encoded {
		return HistoryPosition{}, ErrTaskInvalid
	}
	return p, nil
}

func (m *TaskMetadata) List(ctx context.Context, s TaskScope, cursor string, limit int) (TaskPage, error) {
	if !s.valid() || limit < 1 || limit > 50 {
		return TaskPage{}, ErrTaskInvalid
	}
	if err := m.host.AuthorizeTask(ctx, s, TaskRead); err != nil {
		return TaskPage{}, err
	}
	sqlText := `SELECT task_id,record_version,created_at_ms,updated_at_ms,plan_revision,coalesce(attempt_id,'') FROM development_tasks
 WHERE tenant_ref=? AND workspace_id=? AND account_id=? AND application_id=?`
	args := []any{s.TenantRef, s.WorkspaceID, s.AccountID, s.ApplicationID}
	if cursor != "" {
		p, err := DecodeHistoryPosition(s, "tasks", cursor)
		if err != nil {
			return TaskPage{}, err
		}
		sqlText += " AND (created_at_ms,task_id)<(?,?)"
		args = append(args, p.CreatedAtMS, p.ID)
	}
	rows, err := m.store.db.QueryContext(ctx, sqlText+" ORDER BY created_at_ms DESC,task_id DESC LIMIT ?", append(args, limit+1)...)
	if err != nil {
		return TaskPage{}, ErrUnavailable
	}
	defer rows.Close()
	page := TaskPage{Tasks: []Task{}}
	for rows.Next() {
		if len(page.Tasks) == limit {
			last := page.Tasks[limit-1]
			page.NextCursor, err = EncodeHistoryPosition(s, "tasks", last.CreatedAtMS, last.TaskID)
			if err != nil {
				return TaskPage{}, err
			}
			page.HasMore = true
			break
		}
		t := Task{SourceProfileID: taskSource}
		if err := rows.Scan(&t.TaskID, &t.Version, &t.CreatedAtMS, &t.UpdatedAtMS, &t.Revision, &t.AttemptID); err != nil {
			return TaskPage{}, ErrUnavailable
		}
		page.Tasks = append(page.Tasks, t)
	}
	if rows.Err() != nil {
		return TaskPage{}, ErrUnavailable
	}
	return page, nil
}

type TaskSnapshot struct {
	Task             Task     `json:"task"`
	Plan             TaskPlan `json:"plan"`
	SavedArtifactIDs []string `json:"saved_artifact_ids"`
	ReceiptCount     int      `json:"-"`
}

// Snapshot reads the current task, revision and saved references at one SQLite
// version. Readers cannot observe a mix of two competing plan revisions.
func (m *TaskMetadata) Snapshot(ctx context.Context, s TaskScope, id string) (TaskSnapshot, error) {
	if !s.valid() || !idPattern.MatchString(id) {
		return TaskSnapshot{}, ErrTaskInvalid
	}
	if err := m.host.AuthorizeTask(ctx, s, TaskRead); err != nil {
		return TaskSnapshot{}, err
	}
	tx, err := m.store.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return TaskSnapshot{}, ErrUnavailable
	}
	defer tx.Rollback()
	t, err := readTask(ctx, tx, s, id)
	if err != nil {
		return TaskSnapshot{}, err
	}
	p, err := readTaskPlan(ctx, tx, s, id, t.Revision)
	if err != nil {
		return TaskSnapshot{}, err
	}
	result := TaskSnapshot{Task: t, Plan: p, SavedArtifactIDs: []string{}}
	rows, err := tx.QueryContext(ctx, `SELECT coalesce(artifact_id,'') FROM development_task_receipts WHERE task_id=? ORDER BY request_id LIMIT 65`, id)
	if err != nil {
		return TaskSnapshot{}, ErrUnavailable
	}
	defer rows.Close()
	for rows.Next() {
		var artifact string
		if err := rows.Scan(&artifact); err != nil {
			return TaskSnapshot{}, ErrUnavailable
		}
		result.ReceiptCount++
		if artifact != "" && !slices.Contains(result.SavedArtifactIDs, artifact) {
			result.SavedArtifactIDs = append(result.SavedArtifactIDs, artifact)
		}
	}
	if rows.Err() != nil {
		return TaskSnapshot{}, ErrUnavailable
	}
	if result.ReceiptCount > 64 || result.ReceiptCount < 1 {
		return TaskSnapshot{}, ErrTaskCorrupt
	}
	if err := tx.Commit(); err != nil {
		return TaskSnapshot{}, ErrUnavailable
	}
	return result, nil
}

func (m *TaskMetadata) HasCreateCapacity(ctx context.Context, s TaskScope) (bool, error) {
	if !s.valid() {
		return false, ErrTaskInvalid
	}
	if err := m.host.AuthorizeTask(ctx, s, TaskRead); err != nil {
		return false, err
	}
	return hasTaskCapacity(ctx, m.store.db, s)
}

func hasTaskCapacity(ctx context.Context, q queryer, s TaskScope) (bool, error) {
	var count int
	if err := q.QueryRowContext(ctx, `SELECT count(*) FROM development_tasks WHERE tenant_ref=? AND workspace_id=? AND account_id=?`, s.TenantRef, s.WorkspaceID, s.AccountID).Scan(&count); err != nil {
		return false, ErrUnavailable
	}
	return count < 256, nil
}
