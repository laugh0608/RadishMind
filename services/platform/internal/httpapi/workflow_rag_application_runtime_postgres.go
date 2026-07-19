package httpapi

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresWorkflowRAGApplicationRuntimeRepository struct {
	pool *pgxpool.Pool
}

func newPostgresWorkflowRAGApplicationRuntimeRepository(pool *pgxpool.Pool) *postgresWorkflowRAGApplicationRuntimeRepository {
	return &postgresWorkflowRAGApplicationRuntimeRepository{pool: pool}
}

type postgresWorkflowRAGApplicationRuntimeQueryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func (repository *postgresWorkflowRAGApplicationRuntimeRepository) Read(ctx WorkflowRAGApplicationRuntimeContext) (WorkflowRAGApplicationRuntimeAssignment, []WorkflowRAGApplicationRuntimeEvent, []WorkflowRAGApplicationRuntimeAudit, error) {
	if repository == nil || repository.pool == nil || validateWorkflowRAGApplicationRuntimeContext(ctx) != nil {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, errWorkflowRAGApplicationStore
	}
	tx, err := repository.pool.BeginTx(ctx.RequestContext, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, errWorkflowRAGApplicationStore
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	assignment, events, audits, err := readPostgresWorkflowRAGApplicationRuntimeEntry(ctx, tx)
	if err != nil {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, err
	}
	if err = tx.Commit(ctx.RequestContext); err != nil {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, errWorkflowRAGApplicationStore
	}
	return assignment, events, audits, nil
}

func (repository *postgresWorkflowRAGApplicationRuntimeRepository) Apply(ctx WorkflowRAGApplicationRuntimeContext, expectedVersion int, assignment WorkflowRAGApplicationRuntimeAssignment, event WorkflowRAGApplicationRuntimeEvent, audit WorkflowRAGApplicationRuntimeAudit) error {
	if repository == nil || repository.pool == nil || validateWorkflowRAGApplicationRuntimeContext(ctx) != nil {
		return errWorkflowRAGApplicationStore
	}
	tx, err := repository.pool.BeginTx(ctx.RequestContext, pgx.TxOptions{})
	if err != nil {
		return errWorkflowRAGApplicationStore
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	var lockedVersion int
	lockErr := tx.QueryRow(ctx.RequestContext, `SELECT record_version FROM workflow_rag_application_runtime_assignments
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 FOR UPDATE`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef).Scan(&lockedVersion)
	exists := lockErr == nil
	if lockErr != nil && !errors.Is(lockErr, pgx.ErrNoRows) {
		return errWorkflowRAGApplicationStore
	}
	var current workflowRAGApplicationRuntimeMemoryEntry
	if exists {
		currentAssignment, currentEvents, currentAudits, readErr := readPostgresWorkflowRAGApplicationRuntimeEntry(ctx, tx)
		if readErr != nil {
			return readErr
		}
		current = workflowRAGApplicationRuntimeMemoryEntry{assignment: currentAssignment, events: currentEvents, audits: currentAudits}
	}
	if validateWorkflowRAGApplicationRuntimeMutation(ctx, current, exists, assignment, event, audit) != nil {
		return errWorkflowRAGApplicationStoreContract
	}
	if (exists && lockedVersion != expectedVersion) || (!exists && expectedVersion != 0) {
		return errWorkflowRAGApplicationVersionConflict
	}
	assignmentPayload, err := encodeWorkflowRAGApplicationRuntimeRecord(assignment)
	if err != nil {
		return err
	}
	eventPayload, err := encodeWorkflowRAGApplicationRuntimeRecord(event)
	if err != nil {
		return err
	}
	auditPayload, err := encodeWorkflowRAGApplicationRuntimeRecord(audit)
	if err != nil {
		return err
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, assignment.UpdatedAt)
	if err != nil {
		return errWorkflowRAGApplicationStoreContract
	}
	eventAt, err := time.Parse(time.RFC3339Nano, event.OccurredAt)
	if err != nil {
		return errWorkflowRAGApplicationStoreContract
	}
	auditAt, err := time.Parse(time.RFC3339Nano, audit.OccurredAt)
	if err != nil {
		return errWorkflowRAGApplicationStoreContract
	}
	var command pgconnCommandTag
	if exists {
		command, err = tx.Exec(ctx.RequestContext, `UPDATE workflow_rag_application_runtime_assignments SET
record_version=$1,assignment_state=$2,assignment_digest=$3,updated_at=$4,sanitized_assignment_payload=$5
WHERE tenant_ref=$6 AND workspace_id=$7 AND application_id=$8 AND owner_subject_ref=$9 AND assignment_id=$10 AND record_version=$11`,
			assignment.RecordVersion, assignment.State, assignment.AssignmentDigest, updatedAt, assignmentPayload,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, assignment.AssignmentID, expectedVersion)
	} else {
		command, err = tx.Exec(ctx.RequestContext, `INSERT INTO workflow_rag_application_runtime_assignments
(tenant_ref,workspace_id,application_id,owner_subject_ref,assignment_id,record_version,assignment_state,assignment_digest,updated_at,sanitized_assignment_payload)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT DO NOTHING`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, assignment.AssignmentID,
			assignment.RecordVersion, assignment.State, assignment.AssignmentDigest, updatedAt, assignmentPayload)
	}
	if err != nil {
		return errWorkflowRAGApplicationStore
	}
	if command.RowsAffected() != 1 {
		return errWorkflowRAGApplicationVersionConflict
	}
	if _, err = tx.Exec(ctx.RequestContext, `INSERT INTO workflow_rag_application_runtime_events
(tenant_ref,workspace_id,application_id,owner_subject_ref,assignment_id,event_id,after_record_version,occurred_at,sanitized_event_payload)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, assignment.AssignmentID, event.EventID, event.AfterRecordVersion, eventAt, eventPayload); err != nil {
		return errWorkflowRAGApplicationStore
	}
	if _, err = tx.Exec(ctx.RequestContext, `INSERT INTO workflow_rag_application_runtime_audits
(tenant_ref,workspace_id,application_id,owner_subject_ref,assignment_id,audit_event_id,record_version,occurred_at,sanitized_audit_payload)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, assignment.AssignmentID, audit.AuditEventID, audit.RecordVersion, auditAt, auditPayload); err != nil {
		return errWorkflowRAGApplicationStore
	}
	if err = tx.Commit(ctx.RequestContext); err != nil {
		return errWorkflowRAGApplicationStore
	}
	return nil
}

type pgconnCommandTag interface {
	RowsAffected() int64
}

func readPostgresWorkflowRAGApplicationRuntimeEntry(ctx WorkflowRAGApplicationRuntimeContext, query postgresWorkflowRAGApplicationRuntimeQueryer) (WorkflowRAGApplicationRuntimeAssignment, []WorkflowRAGApplicationRuntimeEvent, []WorkflowRAGApplicationRuntimeAudit, error) {
	var assignmentID, state, digest string
	var recordVersion int
	var updatedAt time.Time
	var payload []byte
	err := query.QueryRow(ctx.RequestContext, `SELECT assignment_id,record_version,assignment_state,assignment_digest,updated_at,sanitized_assignment_payload
FROM workflow_rag_application_runtime_assignments
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef).
		Scan(&assignmentID, &recordVersion, &state, &digest, &updatedAt, &payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, errWorkflowRAGApplicationAssignmentNotFound
	}
	if err != nil {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, errWorkflowRAGApplicationStore
	}
	var assignment WorkflowRAGApplicationRuntimeAssignment
	if decodeWorkflowRAGApplicationRuntimeRecord(payload, &assignment) != nil || validateStoredWorkflowRAGApplicationAssignment(assignment, ctx) != nil {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, errWorkflowRAGApplicationStoreContract
	}
	decodedUpdatedAt, err := time.Parse(time.RFC3339Nano, assignment.UpdatedAt)
	if err != nil || assignmentID != assignment.AssignmentID || recordVersion != assignment.RecordVersion || state != assignment.State || digest != assignment.AssignmentDigest || !postgresWorkflowRAGApplicationRuntimeTimeMatches(updatedAt, decodedUpdatedAt) {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, errWorkflowRAGApplicationStoreContract
	}
	events, err := readPostgresWorkflowRAGApplicationRuntimeEvents(ctx, query, assignment.AssignmentID)
	if err != nil {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, err
	}
	audits, err := readPostgresWorkflowRAGApplicationRuntimeAudits(ctx, query, assignment.AssignmentID)
	if err != nil {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, err
	}
	entry := workflowRAGApplicationRuntimeMemoryEntry{assignment: assignment, events: events, audits: audits}
	if validateWorkflowRAGApplicationRuntimeEntry(ctx, entry) != nil {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, errWorkflowRAGApplicationStoreContract
	}
	return assignment, events, audits, nil
}

func readPostgresWorkflowRAGApplicationRuntimeEvents(ctx WorkflowRAGApplicationRuntimeContext, query postgresWorkflowRAGApplicationRuntimeQueryer, assignmentID string) ([]WorkflowRAGApplicationRuntimeEvent, error) {
	rows, err := query.Query(ctx.RequestContext, `SELECT event_id,after_record_version,occurred_at,sanitized_event_payload
FROM workflow_rag_application_runtime_events WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND assignment_id=$5 ORDER BY after_record_version`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, assignmentID)
	if err != nil {
		return nil, errWorkflowRAGApplicationStore
	}
	defer rows.Close()
	events := make([]WorkflowRAGApplicationRuntimeEvent, 0)
	for rows.Next() {
		var eventID string
		var version int
		var occurredAt time.Time
		var payload []byte
		if rows.Scan(&eventID, &version, &occurredAt, &payload) != nil {
			return nil, errWorkflowRAGApplicationStore
		}
		var event WorkflowRAGApplicationRuntimeEvent
		if decodeWorkflowRAGApplicationRuntimeRecord(payload, &event) != nil || validateStoredWorkflowRAGApplicationEvent(event) != nil {
			return nil, errWorkflowRAGApplicationStoreContract
		}
		decodedAt, decodeErr := time.Parse(time.RFC3339Nano, event.OccurredAt)
		if decodeErr != nil || eventID != event.EventID || version != event.AfterRecordVersion || !postgresWorkflowRAGApplicationRuntimeTimeMatches(occurredAt, decodedAt) {
			return nil, errWorkflowRAGApplicationStoreContract
		}
		events = append(events, event)
	}
	if rows.Err() != nil {
		return nil, errWorkflowRAGApplicationStore
	}
	return events, nil
}

func readPostgresWorkflowRAGApplicationRuntimeAudits(ctx WorkflowRAGApplicationRuntimeContext, query postgresWorkflowRAGApplicationRuntimeQueryer, assignmentID string) ([]WorkflowRAGApplicationRuntimeAudit, error) {
	rows, err := query.Query(ctx.RequestContext, `SELECT audit_event_id,record_version,occurred_at,sanitized_audit_payload
FROM workflow_rag_application_runtime_audits WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND assignment_id=$5 ORDER BY record_version`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, assignmentID)
	if err != nil {
		return nil, errWorkflowRAGApplicationStore
	}
	defer rows.Close()
	audits := make([]WorkflowRAGApplicationRuntimeAudit, 0)
	for rows.Next() {
		var auditID string
		var version int
		var occurredAt time.Time
		var payload []byte
		if rows.Scan(&auditID, &version, &occurredAt, &payload) != nil {
			return nil, errWorkflowRAGApplicationStore
		}
		var audit WorkflowRAGApplicationRuntimeAudit
		if decodeWorkflowRAGApplicationRuntimeRecord(payload, &audit) != nil || validateStoredWorkflowRAGApplicationAudit(audit, ctx) != nil {
			return nil, errWorkflowRAGApplicationStoreContract
		}
		decodedAt, decodeErr := time.Parse(time.RFC3339Nano, audit.OccurredAt)
		if decodeErr != nil || auditID != audit.AuditEventID || version != audit.RecordVersion || !postgresWorkflowRAGApplicationRuntimeTimeMatches(occurredAt, decodedAt) {
			return nil, errWorkflowRAGApplicationStoreContract
		}
		audits = append(audits, audit)
	}
	if rows.Err() != nil {
		return nil, errWorkflowRAGApplicationStore
	}
	return audits, nil
}

func postgresWorkflowRAGApplicationRuntimeTimeMatches(stored, decoded time.Time) bool {
	return stored.UTC().UnixMicro() == decoded.UTC().UnixMicro()
}

var _ workflowRAGApplicationRuntimeRepository = (*postgresWorkflowRAGApplicationRuntimeRepository)(nil)
