package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresPromptApplicationRuntimeRepository struct {
	pool *pgxpool.Pool
}

func newPostgresPromptApplicationRuntimeRepository(pool *pgxpool.Pool) *postgresPromptApplicationRuntimeRepository {
	return &postgresPromptApplicationRuntimeRepository{pool: pool}
}

type postgresPromptApplicationRuntimeQueryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func (repository *postgresPromptApplicationRuntimeRepository) Read(ctx PromptApplicationRuntimeContext) (PromptApplicationRuntimeAssignment, []PromptApplicationRuntimeAssignmentEvent, error) {
	if repository == nil || repository.pool == nil || validatePromptApplicationRuntimeContext(ctx) != nil {
		return PromptApplicationRuntimeAssignment{}, nil, errPromptApplicationRuntimeStore
	}
	tx, err := repository.pool.BeginTx(ctx.RequestContext, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return PromptApplicationRuntimeAssignment{}, nil, errPromptApplicationRuntimeStore
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	assignment, events, err := readPostgresPromptApplicationRuntimeEntry(ctx, tx)
	if err != nil {
		return PromptApplicationRuntimeAssignment{}, nil, err
	}
	if tx.Commit(ctx.RequestContext) != nil {
		return PromptApplicationRuntimeAssignment{}, nil, errPromptApplicationRuntimeStore
	}
	return assignment, events, nil
}

func (repository *postgresPromptApplicationRuntimeRepository) Apply(ctx PromptApplicationRuntimeContext, expectedVersion int, assignment PromptApplicationRuntimeAssignment, event PromptApplicationRuntimeAssignmentEvent) error {
	if repository == nil || repository.pool == nil || validatePromptApplicationRuntimeContext(ctx) != nil {
		return errPromptApplicationRuntimeStore
	}
	tx, err := repository.pool.BeginTx(ctx.RequestContext, pgx.TxOptions{})
	if err != nil {
		return errPromptApplicationRuntimeStore
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var lockedVersion int
	lockErr := tx.QueryRow(ctx.RequestContext, `SELECT assignment_version FROM prompt_application_runtime_assignments
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 FOR UPDATE`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef).Scan(&lockedVersion)
	exists := lockErr == nil
	if lockErr != nil && !errors.Is(lockErr, pgx.ErrNoRows) {
		return errPromptApplicationRuntimeStore
	}
	var current promptApplicationRuntimeMemoryEntry
	if exists {
		currentAssignment, currentEvents, readErr := readPostgresPromptApplicationRuntimeEntry(ctx, tx)
		if readErr != nil {
			return readErr
		}
		current = promptApplicationRuntimeMemoryEntry{assignment: currentAssignment, events: currentEvents}
	}
	if exists && lockedVersion != expectedVersion || !exists && expectedVersion != 0 {
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
	updatedAt, err := time.Parse(time.RFC3339Nano, assignment.UpdatedAt)
	if err != nil {
		return errPromptApplicationRuntimeContract
	}
	occurredAt, err := time.Parse(time.RFC3339Nano, event.OccurredAt)
	if err != nil {
		return errPromptApplicationRuntimeContract
	}
	var command pgconnCommandTag
	if exists {
		command, err = tx.Exec(ctx.RequestContext, `UPDATE prompt_application_runtime_assignments SET
assignment_version=$1,assignment_state=$2,assignment_digest=$3,updated_at=$4,sanitized_assignment_payload=$5
WHERE tenant_ref=$6 AND workspace_id=$7 AND application_id=$8 AND owner_subject_ref=$9 AND assignment_id=$10 AND assignment_version=$11`,
			assignment.AssignmentVersion, assignment.State, assignment.AssignmentDigest, updatedAt, assignmentPayload,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, assignment.AssignmentID, expectedVersion)
	} else {
		command, err = tx.Exec(ctx.RequestContext, `INSERT INTO prompt_application_runtime_assignments
(tenant_ref,workspace_id,application_id,owner_subject_ref,assignment_id,assignment_version,assignment_state,assignment_digest,updated_at,sanitized_assignment_payload)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT DO NOTHING`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef,
			assignment.AssignmentID, assignment.AssignmentVersion, assignment.State, assignment.AssignmentDigest, updatedAt, assignmentPayload)
	}
	if err != nil {
		return errPromptApplicationRuntimeStore
	}
	if command.RowsAffected() != 1 {
		return errPromptApplicationRuntimeVersionConflict
	}
	if _, err = tx.Exec(ctx.RequestContext, `INSERT INTO prompt_application_runtime_assignment_events
(tenant_ref,workspace_id,application_id,owner_subject_ref,event_id,assignment_id,event_sequence,resulting_assignment_version,occurred_at,sanitized_event_payload)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef,
		event.EventID, event.AssignmentID, event.EventSequence, event.ResultingAssignmentVersion, occurredAt, eventPayload); err != nil {
		return errPromptApplicationRuntimeStore
	}
	if tx.Commit(ctx.RequestContext) != nil {
		return errPromptApplicationRuntimeStore
	}
	return nil
}

func readPostgresPromptApplicationRuntimeEntry(ctx PromptApplicationRuntimeContext, query postgresPromptApplicationRuntimeQueryer) (PromptApplicationRuntimeAssignment, []PromptApplicationRuntimeAssignmentEvent, error) {
	var assignmentID, state, digest string
	var version int
	var updatedAt time.Time
	var payload []byte
	err := query.QueryRow(ctx.RequestContext, `SELECT assignment_id,assignment_version,assignment_state,assignment_digest,updated_at,sanitized_assignment_payload
FROM prompt_application_runtime_assignments WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef).Scan(&assignmentID, &version, &state, &digest, &updatedAt, &payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return PromptApplicationRuntimeAssignment{}, nil, errPromptApplicationRuntimeNotFound
	}
	if err != nil {
		return PromptApplicationRuntimeAssignment{}, nil, errPromptApplicationRuntimeStore
	}
	var assignment PromptApplicationRuntimeAssignment
	if json.Unmarshal(payload, &assignment) != nil || validatePromptApplicationRuntimeAssignment(assignment) != nil {
		return PromptApplicationRuntimeAssignment{}, nil, errPromptApplicationRuntimeContract
	}
	decodedAt, err := time.Parse(time.RFC3339Nano, assignment.UpdatedAt)
	if err != nil || assignmentID != assignment.AssignmentID || version != assignment.AssignmentVersion || state != assignment.State || digest != assignment.AssignmentDigest || !postgresWorkflowRAGApplicationRuntimeTimeMatches(updatedAt, decodedAt) {
		return PromptApplicationRuntimeAssignment{}, nil, errPromptApplicationRuntimeContract
	}
	events, err := readPostgresPromptApplicationRuntimeEvents(ctx, query, assignment.AssignmentID)
	if err != nil {
		return PromptApplicationRuntimeAssignment{}, nil, err
	}
	if validatePromptApplicationRuntimeEntry(ctx, promptApplicationRuntimeMemoryEntry{assignment: assignment, events: events}) != nil {
		return PromptApplicationRuntimeAssignment{}, nil, errPromptApplicationRuntimeContract
	}
	return assignment, events, nil
}

func readPostgresPromptApplicationRuntimeEvents(ctx PromptApplicationRuntimeContext, query postgresPromptApplicationRuntimeQueryer, assignmentID string) ([]PromptApplicationRuntimeAssignmentEvent, error) {
	rows, err := query.Query(ctx.RequestContext, `SELECT event_id,event_sequence,resulting_assignment_version,occurred_at,sanitized_event_payload
FROM prompt_application_runtime_assignment_events
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND assignment_id=$5 ORDER BY event_sequence`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, assignmentID)
	if err != nil {
		return nil, errPromptApplicationRuntimeStore
	}
	defer rows.Close()
	events := make([]PromptApplicationRuntimeAssignmentEvent, 0)
	for rows.Next() {
		var eventID string
		var sequence, version int
		var occurredAt time.Time
		var payload []byte
		if rows.Scan(&eventID, &sequence, &version, &occurredAt, &payload) != nil {
			return nil, errPromptApplicationRuntimeStore
		}
		var event PromptApplicationRuntimeAssignmentEvent
		if json.Unmarshal(payload, &event) != nil || validatePromptApplicationRuntimeAssignmentEvent(event) != nil {
			return nil, errPromptApplicationRuntimeContract
		}
		decodedAt, decodeErr := time.Parse(time.RFC3339Nano, event.OccurredAt)
		if decodeErr != nil || eventID != event.EventID || sequence != event.EventSequence || version != event.ResultingAssignmentVersion || !postgresWorkflowRAGApplicationRuntimeTimeMatches(occurredAt, decodedAt) {
			return nil, errPromptApplicationRuntimeContract
		}
		events = append(events, event)
	}
	if rows.Err() != nil {
		return nil, errPromptApplicationRuntimeStore
	}
	return events, nil
}
