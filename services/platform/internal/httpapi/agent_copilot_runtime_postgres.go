package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresAgentCopilotRuntimeRepository struct {
	pool *pgxpool.Pool
}

func newPostgresAgentCopilotRuntimeRepository(pool *pgxpool.Pool) *postgresAgentCopilotRuntimeRepository {
	return &postgresAgentCopilotRuntimeRepository{pool: pool}
}

type postgresAgentCopilotRuntimeQueryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func (repository *postgresAgentCopilotRuntimeRepository) Read(ctx AgentCopilotRuntimeContext) (AgentCopilotRuntimeAssignmentV1, []AgentCopilotRuntimeAssignmentEventV1, error) {
	if repository == nil || repository.pool == nil || validateAgentCopilotRuntimeContext(ctx) != nil {
		return AgentCopilotRuntimeAssignmentV1{}, nil, errAgentCopilotRuntimeStore
	}
	tx, err := repository.pool.BeginTx(ctx.RequestContext, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return AgentCopilotRuntimeAssignmentV1{}, nil, errAgentCopilotRuntimeStore
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	assignment, events, err := readPostgresAgentCopilotRuntimeEntry(ctx, tx)
	if err != nil {
		return AgentCopilotRuntimeAssignmentV1{}, nil, err
	}
	if tx.Commit(ctx.RequestContext) != nil {
		return AgentCopilotRuntimeAssignmentV1{}, nil, errAgentCopilotRuntimeStore
	}
	return assignment, events, nil
}

func (repository *postgresAgentCopilotRuntimeRepository) Apply(
	ctx AgentCopilotRuntimeContext,
	expectedVersion int,
	assignment AgentCopilotRuntimeAssignmentV1,
	event AgentCopilotRuntimeAssignmentEventV1,
) error {
	if repository == nil || repository.pool == nil || validateAgentCopilotRuntimeContext(ctx) != nil {
		return errAgentCopilotRuntimeStore
	}
	tx, err := repository.pool.BeginTx(ctx.RequestContext, pgx.TxOptions{})
	if err != nil {
		return errAgentCopilotRuntimeStore
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var lockedVersion int
	lockErr := tx.QueryRow(ctx.RequestContext, `SELECT assignment_version FROM agent_copilot_runtime_assignments
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 FOR UPDATE`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef,
	).Scan(&lockedVersion)
	exists := lockErr == nil
	if lockErr != nil && !errors.Is(lockErr, pgx.ErrNoRows) {
		return errAgentCopilotRuntimeStore
	}
	var current agentCopilotRuntimeMemoryEntry
	if exists {
		currentAssignment, currentEvents, readErr := readPostgresAgentCopilotRuntimeEntry(ctx, tx)
		if readErr != nil {
			return readErr
		}
		current = agentCopilotRuntimeMemoryEntry{assignment: currentAssignment, events: currentEvents}
	}
	if exists && lockedVersion != expectedVersion || !exists && expectedVersion != 0 {
		return errAgentCopilotRuntimeVersionConflict
	}
	if validateAgentCopilotRuntimeMutation(ctx, current, exists, assignment, event) != nil {
		return errAgentCopilotRuntimeContract
	}
	assignmentPayload, err := json.Marshal(assignment)
	if err != nil {
		return errAgentCopilotRuntimeContract
	}
	eventPayload, err := json.Marshal(event)
	if err != nil {
		return errAgentCopilotRuntimeContract
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, assignment.UpdatedAt)
	if err != nil {
		return errAgentCopilotRuntimeContract
	}
	occurredAt, err := time.Parse(time.RFC3339Nano, event.OccurredAt)
	if err != nil {
		return errAgentCopilotRuntimeContract
	}
	var command pgconnCommandTag
	if exists {
		command, err = tx.Exec(ctx.RequestContext, `UPDATE agent_copilot_runtime_assignments SET
assignment_version=$1,assignment_state=$2,assignment_digest=$3,updated_at=$4,sanitized_assignment_payload=$5
WHERE tenant_ref=$6 AND workspace_id=$7 AND application_id=$8 AND owner_subject_ref=$9 AND assignment_id=$10 AND assignment_version=$11`,
			assignment.AssignmentVersion, assignment.State, assignment.AssignmentDigest, updatedAt, assignmentPayload,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, assignment.AssignmentID, expectedVersion)
	} else {
		command, err = tx.Exec(ctx.RequestContext, `INSERT INTO agent_copilot_runtime_assignments
(tenant_ref,workspace_id,application_id,owner_subject_ref,assignment_id,assignment_version,assignment_state,assignment_digest,updated_at,sanitized_assignment_payload)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT DO NOTHING`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef,
			assignment.AssignmentID, assignment.AssignmentVersion, assignment.State,
			assignment.AssignmentDigest, updatedAt, assignmentPayload)
	}
	if err != nil {
		return errAgentCopilotRuntimeStore
	}
	if command.RowsAffected() != 1 {
		return errAgentCopilotRuntimeVersionConflict
	}
	if _, err = tx.Exec(ctx.RequestContext, `INSERT INTO agent_copilot_runtime_assignment_events
(tenant_ref,workspace_id,application_id,owner_subject_ref,event_id,assignment_id,event_sequence,resulting_assignment_version,occurred_at,sanitized_event_payload)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef,
		event.EventID, event.AssignmentID, event.EventSequence, event.ResultingAssignmentVersion,
		occurredAt, eventPayload,
	); err != nil {
		return errAgentCopilotRuntimeStore
	}
	if tx.Commit(ctx.RequestContext) != nil {
		return errAgentCopilotRuntimeStore
	}
	return nil
}

func readPostgresAgentCopilotRuntimeEntry(
	ctx AgentCopilotRuntimeContext,
	query postgresAgentCopilotRuntimeQueryer,
) (AgentCopilotRuntimeAssignmentV1, []AgentCopilotRuntimeAssignmentEventV1, error) {
	var assignmentID, state, digest string
	var version int
	var updatedAt time.Time
	var payload []byte
	err := query.QueryRow(ctx.RequestContext, `SELECT assignment_id,assignment_version,assignment_state,assignment_digest,updated_at,sanitized_assignment_payload
FROM agent_copilot_runtime_assignments WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef,
	).Scan(&assignmentID, &version, &state, &digest, &updatedAt, &payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return AgentCopilotRuntimeAssignmentV1{}, nil, errAgentCopilotRuntimeNotFound
	}
	if err != nil {
		return AgentCopilotRuntimeAssignmentV1{}, nil, errAgentCopilotRuntimeStore
	}
	decodedAssignment, decodeErr := decodeAgentCopilotContract(agentCopilotRuntimeAssignmentSchema, payload)
	if decodeErr != nil {
		return AgentCopilotRuntimeAssignmentV1{}, nil, errAgentCopilotRuntimeContract
	}
	assignment := *decodedAssignment.(*AgentCopilotRuntimeAssignmentV1)
	decodedAt, err := time.Parse(time.RFC3339Nano, assignment.UpdatedAt)
	if err != nil || assignmentID != assignment.AssignmentID || version != assignment.AssignmentVersion ||
		state != assignment.State || digest != assignment.AssignmentDigest ||
		!postgresWorkflowRAGApplicationRuntimeTimeMatches(updatedAt, decodedAt) {
		return AgentCopilotRuntimeAssignmentV1{}, nil, errAgentCopilotRuntimeContract
	}
	events, err := readPostgresAgentCopilotRuntimeEvents(ctx, query, assignment.AssignmentID)
	if err != nil {
		return AgentCopilotRuntimeAssignmentV1{}, nil, err
	}
	if validateAgentCopilotRuntimeEntry(ctx, agentCopilotRuntimeMemoryEntry{assignment: assignment, events: events}) != nil {
		return AgentCopilotRuntimeAssignmentV1{}, nil, errAgentCopilotRuntimeContract
	}
	return assignment, events, nil
}

func readPostgresAgentCopilotRuntimeEvents(
	ctx AgentCopilotRuntimeContext,
	query postgresAgentCopilotRuntimeQueryer,
	assignmentID string,
) ([]AgentCopilotRuntimeAssignmentEventV1, error) {
	rows, err := query.Query(ctx.RequestContext, `SELECT event_id,event_sequence,resulting_assignment_version,occurred_at,sanitized_event_payload
FROM agent_copilot_runtime_assignment_events
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND assignment_id=$5 ORDER BY event_sequence`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, assignmentID)
	if err != nil {
		return nil, errAgentCopilotRuntimeStore
	}
	defer rows.Close()
	events := make([]AgentCopilotRuntimeAssignmentEventV1, 0)
	for rows.Next() {
		var eventID string
		var sequence, version int
		var occurredAt time.Time
		var payload []byte
		if rows.Scan(&eventID, &sequence, &version, &occurredAt, &payload) != nil {
			return nil, errAgentCopilotRuntimeStore
		}
		decodedEvent, decodeErr := decodeAgentCopilotContract(agentCopilotRuntimeAssignmentEventSchema, payload)
		if decodeErr != nil {
			return nil, errAgentCopilotRuntimeContract
		}
		event := *decodedEvent.(*AgentCopilotRuntimeAssignmentEventV1)
		decodedAt, decodeErr := time.Parse(time.RFC3339Nano, event.OccurredAt)
		if decodeErr != nil || eventID != event.EventID || sequence != event.EventSequence ||
			version != event.ResultingAssignmentVersion ||
			!postgresWorkflowRAGApplicationRuntimeTimeMatches(occurredAt, decodedAt) {
			return nil, errAgentCopilotRuntimeContract
		}
		events = append(events, event)
	}
	if rows.Err() != nil {
		return nil, errAgentCopilotRuntimeStore
	}
	return events, nil
}
