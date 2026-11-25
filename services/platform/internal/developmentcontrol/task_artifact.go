package developmentcontrol

import (
	"context"
	"database/sql"
)

// The artifact owner validates the trusted snapshot and writes the immutable
// package on the same transaction connection as the public request receipt.
type TaskArtifactWriter interface {
	SaveTaskArtifact(context.Context, *sql.Conn, TaskScope, Scope, string, string, string) (string, error)
}

func (c *TaskCoordinator) SaveArtifact(ctx context.Context, s TaskScope, task, attempt, request, digest string, writer TaskArtifactWriter) (TaskReceipt, error) {
	if !idPattern.MatchString(attempt) || !digestPattern.MatchString(digest) {
		return TaskReceipt{}, ErrTaskInvalid
	}
	input := struct {
		Attempt string `json:"attempt_id"`
		Digest  string `json:"expected_digest"`
	}{attempt, digest}
	return c.write(ctx, s, TaskSaveArtifact, task, request, input, func(conn *sql.Conn, t Task, r *TaskReceipt) error {
		if t.AttemptID != attempt {
			return ErrNotFound
		}
		record, err := readRecord(ctx, conn, s.Identity, attempt)
		if err != nil {
			return err
		}
		if !belongsToTask(record, s, task, attempt) {
			return ErrTaskCorrupt
		}
		if !taskResourcesEnded(record) {
			return ErrConflict
		}
		if writer == nil {
			return ErrUnavailable
		}
		id, err := writer.SaveTaskArtifact(ctx, conn, s, record.Control.Delegation().Scope, record.Control.phase, request, digest)
		if err != nil {
			return err
		}
		r.AttemptID, r.ArtifactID = attempt, id
		return nil
	})
}

func taskResourcesEnded(record *Record) bool {
	if !record.Revoked {
		return false
	}
	return record.Control.terminationStatus == "confirmed" || record.Control.phase == "completed" && hasControlEvent(record, "sandbox_closed") && hasControlEvent(record, "worker_exited")
}
