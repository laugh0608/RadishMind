package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type sqlitePromptApplicationRuntimeRepository struct {
	database *sql.DB
}

func newSQLitePromptApplicationRuntimeRepository(database *sql.DB) *sqlitePromptApplicationRuntimeRepository {
	return &sqlitePromptApplicationRuntimeRepository{database: database}
}

type sqlitePromptApplicationRuntimeQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func (repository *sqlitePromptApplicationRuntimeRepository) Read(ctx PromptApplicationRuntimeContext) (PromptApplicationRuntimeAssignment, []PromptApplicationRuntimeAssignmentEvent, error) {
	if repository == nil || repository.database == nil || validatePromptApplicationRuntimeContext(ctx) != nil {
		return PromptApplicationRuntimeAssignment{}, nil, errPromptApplicationRuntimeStore
	}
	tx, err := repository.database.BeginTx(ctx.RequestContext, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return PromptApplicationRuntimeAssignment{}, nil, errPromptApplicationRuntimeStore
	}
	defer func() { _ = tx.Rollback() }()
	assignment, events, err := readSQLitePromptApplicationRuntimeEntry(ctx, tx)
	if err != nil {
		return PromptApplicationRuntimeAssignment{}, nil, err
	}
	if tx.Commit() != nil {
		return PromptApplicationRuntimeAssignment{}, nil, errPromptApplicationRuntimeStore
	}
	return assignment, events, nil
}

func (repository *sqlitePromptApplicationRuntimeRepository) Apply(ctx PromptApplicationRuntimeContext, expectedVersion int, assignment PromptApplicationRuntimeAssignment, event PromptApplicationRuntimeAssignmentEvent) error {
	if repository == nil || repository.database == nil || validatePromptApplicationRuntimeContext(ctx) != nil {
		return errPromptApplicationRuntimeStore
	}
	connection, err := repository.database.Conn(ctx.RequestContext)
	if err != nil {
		return errPromptApplicationRuntimeStore
	}
	defer connection.Close()
	if _, err = connection.ExecContext(ctx.RequestContext, "BEGIN IMMEDIATE"); err != nil {
		return errPromptApplicationRuntimeStore
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = connection.ExecContext(context.Background(), "ROLLBACK")
		}
	}()
	currentAssignment, currentEvents, readErr := readSQLitePromptApplicationRuntimeEntry(ctx, connection)
	exists := readErr == nil
	if readErr != nil && !errors.Is(readErr, errPromptApplicationRuntimeNotFound) {
		return readErr
	}
	current := promptApplicationRuntimeMemoryEntry{assignment: currentAssignment, events: currentEvents}
	if exists && currentAssignment.AssignmentVersion != expectedVersion || !exists && expectedVersion != 0 {
		return errPromptApplicationRuntimeVersionConflict
	}
	if validatePromptApplicationRuntimeMutation(ctx, current, exists, assignment, event) != nil {
		return errPromptApplicationRuntimeContract
	}
	assignmentPayload, err := json.Marshal(assignment)
	if err != nil {
		return errPromptApplicationRuntimeContract
	}
	eventPayload, err := json.Marshal(event)
	if err != nil {
		return errPromptApplicationRuntimeContract
	}
	updatedAt, err := promptApplicationRuntimeUnixNano(assignment.UpdatedAt)
	if err != nil {
		return errPromptApplicationRuntimeContract
	}
	occurredAt, err := promptApplicationRuntimeUnixNano(event.OccurredAt)
	if err != nil {
		return errPromptApplicationRuntimeContract
	}
	var result sql.Result
	if exists {
		result, err = connection.ExecContext(ctx.RequestContext, `UPDATE prompt_application_runtime_assignments
SET assignment_version=?,assignment_state=?,assignment_digest=?,updated_at_unix_nano=?,sanitized_assignment_payload=?
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND assignment_id=? AND assignment_version=?`,
			assignment.AssignmentVersion, assignment.State, assignment.AssignmentDigest, updatedAt, string(assignmentPayload),
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, assignment.AssignmentID, expectedVersion)
	} else {
		result, err = connection.ExecContext(ctx.RequestContext, `INSERT INTO prompt_application_runtime_assignments
(tenant_ref,workspace_id,application_id,owner_subject_ref,assignment_id,assignment_version,assignment_state,assignment_digest,updated_at_unix_nano,sanitized_assignment_payload)
VALUES (?,?,?,?,?,?,?,?,?,?) ON CONFLICT (tenant_ref,workspace_id,application_id,owner_subject_ref) DO NOTHING`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, assignment.AssignmentID, assignment.AssignmentVersion,
			assignment.State, assignment.AssignmentDigest, updatedAt, string(assignmentPayload))
	}
	if err != nil {
		return errPromptApplicationRuntimeStore
	}
	rows, err := result.RowsAffected()
	if err != nil || rows != 1 {
		return errPromptApplicationRuntimeVersionConflict
	}
	if _, err = connection.ExecContext(ctx.RequestContext, `INSERT INTO prompt_application_runtime_assignment_events
(tenant_ref,workspace_id,application_id,owner_subject_ref,event_id,assignment_id,event_sequence,resulting_assignment_version,occurred_at_unix_nano,sanitized_event_payload)
VALUES (?,?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, event.EventID, event.AssignmentID,
		event.EventSequence, event.ResultingAssignmentVersion, occurredAt, string(eventPayload)); err != nil {
		return errPromptApplicationRuntimeStore
	}
	if _, err = connection.ExecContext(ctx.RequestContext, "COMMIT"); err != nil {
		return errPromptApplicationRuntimeStore
	}
	committed = true
	return nil
}

func readSQLitePromptApplicationRuntimeEntry(ctx PromptApplicationRuntimeContext, query sqlitePromptApplicationRuntimeQueryer) (PromptApplicationRuntimeAssignment, []PromptApplicationRuntimeAssignmentEvent, error) {
	var assignmentID, state, digest string
	var assignmentVersion int
	var updatedAt int64
	var payload []byte
	err := query.QueryRowContext(ctx.RequestContext, `SELECT assignment_id,assignment_version,assignment_state,assignment_digest,updated_at_unix_nano,sanitized_assignment_payload
FROM prompt_application_runtime_assignments WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef).Scan(&assignmentID, &assignmentVersion, &state, &digest, &updatedAt, &payload)
	if errors.Is(err, sql.ErrNoRows) {
		return PromptApplicationRuntimeAssignment{}, nil, errPromptApplicationRuntimeNotFound
	}
	if err != nil {
		return PromptApplicationRuntimeAssignment{}, nil, errPromptApplicationRuntimeStore
	}
	var assignment PromptApplicationRuntimeAssignment
	if json.Unmarshal(payload, &assignment) != nil || validatePromptApplicationRuntimeAssignment(assignment) != nil {
		return PromptApplicationRuntimeAssignment{}, nil, errPromptApplicationRuntimeContract
	}
	decodedUpdatedAt, err := promptApplicationRuntimeUnixNano(assignment.UpdatedAt)
	if err != nil || assignmentID != assignment.AssignmentID || assignmentVersion != assignment.AssignmentVersion || state != assignment.State || digest != assignment.AssignmentDigest || updatedAt != decodedUpdatedAt {
		return PromptApplicationRuntimeAssignment{}, nil, errPromptApplicationRuntimeContract
	}
	events, err := readSQLitePromptApplicationRuntimeEvents(ctx, query, assignment.AssignmentID)
	if err != nil {
		return PromptApplicationRuntimeAssignment{}, nil, err
	}
	if validatePromptApplicationRuntimeEntry(ctx, promptApplicationRuntimeMemoryEntry{assignment: assignment, events: events}) != nil {
		return PromptApplicationRuntimeAssignment{}, nil, errPromptApplicationRuntimeContract
	}
	return assignment, events, nil
}

func readSQLitePromptApplicationRuntimeEvents(ctx PromptApplicationRuntimeContext, query sqlitePromptApplicationRuntimeQueryer, assignmentID string) ([]PromptApplicationRuntimeAssignmentEvent, error) {
	rows, err := query.QueryContext(ctx.RequestContext, `SELECT event_id,event_sequence,resulting_assignment_version,occurred_at_unix_nano,sanitized_event_payload
FROM prompt_application_runtime_assignment_events
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND assignment_id=? ORDER BY event_sequence`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, assignmentID)
	if err != nil {
		return nil, errPromptApplicationRuntimeStore
	}
	defer rows.Close()
	events := make([]PromptApplicationRuntimeAssignmentEvent, 0)
	for rows.Next() {
		var eventID string
		var sequence, version int
		var occurredAt int64
		var payload []byte
		if rows.Scan(&eventID, &sequence, &version, &occurredAt, &payload) != nil {
			return nil, errPromptApplicationRuntimeStore
		}
		var event PromptApplicationRuntimeAssignmentEvent
		if json.Unmarshal(payload, &event) != nil || validatePromptApplicationRuntimeAssignmentEvent(event) != nil {
			return nil, errPromptApplicationRuntimeContract
		}
		decodedAt, decodeErr := promptApplicationRuntimeUnixNano(event.OccurredAt)
		if decodeErr != nil || eventID != event.EventID || sequence != event.EventSequence || version != event.ResultingAssignmentVersion || occurredAt != decodedAt {
			return nil, errPromptApplicationRuntimeContract
		}
		events = append(events, event)
	}
	if rows.Err() != nil {
		return nil, errPromptApplicationRuntimeStore
	}
	return events, nil
}

func promptApplicationRuntimeUnixNano(value string) (int64, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return 0, err
	}
	return parsed.UnixNano(), nil
}
