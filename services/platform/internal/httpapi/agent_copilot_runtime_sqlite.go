package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

type sqliteAgentCopilotRuntimeRepository struct {
	database *sql.DB
}

func newSQLiteAgentCopilotRuntimeRepository(database *sql.DB) *sqliteAgentCopilotRuntimeRepository {
	return &sqliteAgentCopilotRuntimeRepository{database: database}
}

type sqliteAgentCopilotRuntimeQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func (repository *sqliteAgentCopilotRuntimeRepository) Read(ctx AgentCopilotRuntimeContext) (AgentCopilotRuntimeAssignmentV1, []AgentCopilotRuntimeAssignmentEventV1, error) {
	if repository == nil || repository.database == nil || validateAgentCopilotRuntimeContext(ctx) != nil {
		return AgentCopilotRuntimeAssignmentV1{}, nil, errAgentCopilotRuntimeStore
	}
	tx, err := repository.database.BeginTx(ctx.RequestContext, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return AgentCopilotRuntimeAssignmentV1{}, nil, errAgentCopilotRuntimeStore
	}
	defer func() { _ = tx.Rollback() }()
	assignment, events, err := readSQLiteAgentCopilotRuntimeEntry(ctx, tx)
	if err != nil {
		return AgentCopilotRuntimeAssignmentV1{}, nil, err
	}
	if tx.Commit() != nil {
		return AgentCopilotRuntimeAssignmentV1{}, nil, errAgentCopilotRuntimeStore
	}
	return assignment, events, nil
}

func (repository *sqliteAgentCopilotRuntimeRepository) Apply(
	ctx AgentCopilotRuntimeContext,
	expectedVersion int,
	assignment AgentCopilotRuntimeAssignmentV1,
	event AgentCopilotRuntimeAssignmentEventV1,
) error {
	if repository == nil || repository.database == nil || validateAgentCopilotRuntimeContext(ctx) != nil {
		return errAgentCopilotRuntimeStore
	}
	connection, err := repository.database.Conn(ctx.RequestContext)
	if err != nil {
		return errAgentCopilotRuntimeStore
	}
	defer connection.Close()
	if _, err = connection.ExecContext(ctx.RequestContext, "BEGIN IMMEDIATE"); err != nil {
		return errAgentCopilotRuntimeStore
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = connection.ExecContext(context.Background(), "ROLLBACK")
		}
	}()
	currentAssignment, currentEvents, readErr := readSQLiteAgentCopilotRuntimeEntry(ctx, connection)
	exists := readErr == nil
	if readErr != nil && !errors.Is(readErr, errAgentCopilotRuntimeNotFound) {
		return readErr
	}
	current := agentCopilotRuntimeMemoryEntry{assignment: currentAssignment, events: currentEvents}
	if exists && currentAssignment.AssignmentVersion != expectedVersion || !exists && expectedVersion != 0 {
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
	updatedAt, err := promptApplicationRuntimeUnixNano(assignment.UpdatedAt)
	if err != nil {
		return errAgentCopilotRuntimeContract
	}
	occurredAt, err := promptApplicationRuntimeUnixNano(event.OccurredAt)
	if err != nil {
		return errAgentCopilotRuntimeContract
	}
	var result sql.Result
	if exists {
		result, err = connection.ExecContext(ctx.RequestContext, `UPDATE agent_copilot_runtime_assignments
SET assignment_version=?,assignment_state=?,assignment_digest=?,updated_at_unix_nano=?,sanitized_assignment_payload=?
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND assignment_id=? AND assignment_version=?`,
			assignment.AssignmentVersion, assignment.State, assignment.AssignmentDigest, updatedAt, string(assignmentPayload),
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, assignment.AssignmentID, expectedVersion)
	} else {
		result, err = connection.ExecContext(ctx.RequestContext, `INSERT INTO agent_copilot_runtime_assignments
(tenant_ref,workspace_id,application_id,owner_subject_ref,assignment_id,assignment_version,assignment_state,assignment_digest,updated_at_unix_nano,sanitized_assignment_payload)
VALUES (?,?,?,?,?,?,?,?,?,?) ON CONFLICT (tenant_ref,workspace_id,application_id,owner_subject_ref) DO NOTHING`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, assignment.AssignmentID,
			assignment.AssignmentVersion, assignment.State, assignment.AssignmentDigest, updatedAt, string(assignmentPayload))
	}
	if err != nil {
		return errAgentCopilotRuntimeStore
	}
	rows, err := result.RowsAffected()
	if err != nil || rows != 1 {
		return errAgentCopilotRuntimeVersionConflict
	}
	if _, err = connection.ExecContext(ctx.RequestContext, `INSERT INTO agent_copilot_runtime_assignment_events
(tenant_ref,workspace_id,application_id,owner_subject_ref,event_id,assignment_id,event_sequence,resulting_assignment_version,occurred_at_unix_nano,sanitized_event_payload)
VALUES (?,?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef,
		event.EventID, event.AssignmentID, event.EventSequence, event.ResultingAssignmentVersion,
		occurredAt, string(eventPayload)); err != nil {
		return errAgentCopilotRuntimeStore
	}
	if _, err = connection.ExecContext(ctx.RequestContext, "COMMIT"); err != nil {
		return errAgentCopilotRuntimeStore
	}
	committed = true
	return nil
}

func readSQLiteAgentCopilotRuntimeEntry(
	ctx AgentCopilotRuntimeContext,
	query sqliteAgentCopilotRuntimeQueryer,
) (AgentCopilotRuntimeAssignmentV1, []AgentCopilotRuntimeAssignmentEventV1, error) {
	var assignmentID, state, digest string
	var assignmentVersion int
	var updatedAt int64
	var payload []byte
	err := query.QueryRowContext(ctx.RequestContext, `SELECT assignment_id,assignment_version,assignment_state,assignment_digest,updated_at_unix_nano,sanitized_assignment_payload
FROM agent_copilot_runtime_assignments WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef,
	).Scan(&assignmentID, &assignmentVersion, &state, &digest, &updatedAt, &payload)
	if errors.Is(err, sql.ErrNoRows) {
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
	decodedUpdatedAt, err := promptApplicationRuntimeUnixNano(assignment.UpdatedAt)
	if err != nil || assignmentID != assignment.AssignmentID || assignmentVersion != assignment.AssignmentVersion ||
		state != assignment.State || digest != assignment.AssignmentDigest || updatedAt != decodedUpdatedAt {
		return AgentCopilotRuntimeAssignmentV1{}, nil, errAgentCopilotRuntimeContract
	}
	events, err := readSQLiteAgentCopilotRuntimeEvents(ctx, query, assignment.AssignmentID)
	if err != nil {
		return AgentCopilotRuntimeAssignmentV1{}, nil, err
	}
	if validateAgentCopilotRuntimeEntry(ctx, agentCopilotRuntimeMemoryEntry{assignment: assignment, events: events}) != nil {
		return AgentCopilotRuntimeAssignmentV1{}, nil, errAgentCopilotRuntimeContract
	}
	return assignment, events, nil
}

func readSQLiteAgentCopilotRuntimeEvents(
	ctx AgentCopilotRuntimeContext,
	query sqliteAgentCopilotRuntimeQueryer,
	assignmentID string,
) ([]AgentCopilotRuntimeAssignmentEventV1, error) {
	rows, err := query.QueryContext(ctx.RequestContext, `SELECT event_id,event_sequence,resulting_assignment_version,occurred_at_unix_nano,sanitized_event_payload
FROM agent_copilot_runtime_assignment_events
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND assignment_id=? ORDER BY event_sequence`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, assignmentID)
	if err != nil {
		return nil, errAgentCopilotRuntimeStore
	}
	defer rows.Close()
	events := make([]AgentCopilotRuntimeAssignmentEventV1, 0)
	for rows.Next() {
		var eventID string
		var sequence, version int
		var occurredAt int64
		var payload []byte
		if rows.Scan(&eventID, &sequence, &version, &occurredAt, &payload) != nil {
			return nil, errAgentCopilotRuntimeStore
		}
		decodedEvent, decodeErr := decodeAgentCopilotContract(agentCopilotRuntimeAssignmentEventSchema, payload)
		if decodeErr != nil {
			return nil, errAgentCopilotRuntimeContract
		}
		event := *decodedEvent.(*AgentCopilotRuntimeAssignmentEventV1)
		decodedAt, decodeErr := promptApplicationRuntimeUnixNano(event.OccurredAt)
		if decodeErr != nil || eventID != event.EventID || sequence != event.EventSequence ||
			version != event.ResultingAssignmentVersion || occurredAt != decodedAt {
			return nil, errAgentCopilotRuntimeContract
		}
		events = append(events, event)
	}
	if rows.Err() != nil {
		return nil, errAgentCopilotRuntimeStore
	}
	return events, nil
}
