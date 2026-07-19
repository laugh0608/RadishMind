package httpapi

import (
	"context"
	"database/sql"
	"errors"
)

type sqliteWorkflowRAGApplicationRuntimeRepository struct {
	database *sql.DB
}

func newSQLiteWorkflowRAGApplicationRuntimeRepository(database *sql.DB) *sqliteWorkflowRAGApplicationRuntimeRepository {
	return &sqliteWorkflowRAGApplicationRuntimeRepository{database: database}
}

type sqliteWorkflowRAGApplicationRuntimeQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func (repository *sqliteWorkflowRAGApplicationRuntimeRepository) Read(ctx WorkflowRAGApplicationRuntimeContext) (WorkflowRAGApplicationRuntimeAssignment, []WorkflowRAGApplicationRuntimeEvent, []WorkflowRAGApplicationRuntimeAudit, error) {
	if repository == nil || repository.database == nil || validateWorkflowRAGApplicationRuntimeContext(ctx) != nil {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, errWorkflowRAGApplicationStore
	}
	tx, err := repository.database.BeginTx(ctx.RequestContext, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, errWorkflowRAGApplicationStore
	}
	defer func() { _ = tx.Rollback() }()
	assignment, events, audits, err := readSQLiteWorkflowRAGApplicationRuntimeEntry(ctx, tx)
	if err != nil {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, err
	}
	if err = tx.Commit(); err != nil {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, errWorkflowRAGApplicationStore
	}
	return assignment, events, audits, nil
}

func (repository *sqliteWorkflowRAGApplicationRuntimeRepository) Apply(ctx WorkflowRAGApplicationRuntimeContext, expectedVersion int, assignment WorkflowRAGApplicationRuntimeAssignment, event WorkflowRAGApplicationRuntimeEvent, audit WorkflowRAGApplicationRuntimeAudit) error {
	if repository == nil || repository.database == nil || validateWorkflowRAGApplicationRuntimeContext(ctx) != nil {
		return errWorkflowRAGApplicationStore
	}
	connection, err := repository.database.Conn(ctx.RequestContext)
	if err != nil {
		return errWorkflowRAGApplicationStore
	}
	defer connection.Close()
	if _, err = connection.ExecContext(ctx.RequestContext, "BEGIN IMMEDIATE"); err != nil {
		return errWorkflowRAGApplicationStore
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = connection.ExecContext(context.Background(), "ROLLBACK")
		}
	}()

	currentAssignment, currentEvents, currentAudits, readErr := readSQLiteWorkflowRAGApplicationRuntimeEntry(ctx, connection)
	exists := readErr == nil
	if readErr != nil && !errors.Is(readErr, errWorkflowRAGApplicationAssignmentNotFound) {
		return readErr
	}
	current := workflowRAGApplicationRuntimeMemoryEntry{assignment: currentAssignment, events: currentEvents, audits: currentAudits}
	if validateWorkflowRAGApplicationRuntimeMutation(ctx, current, exists, assignment, event, audit) != nil {
		return errWorkflowRAGApplicationStoreContract
	}
	if (exists && currentAssignment.RecordVersion != expectedVersion) || (!exists && expectedVersion != 0) {
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
	updatedAt, err := workflowRAGApplicationRuntimeUnixNano(assignment.UpdatedAt)
	if err != nil {
		return err
	}
	eventAt, err := workflowRAGApplicationRuntimeUnixNano(event.OccurredAt)
	if err != nil {
		return err
	}
	auditAt, err := workflowRAGApplicationRuntimeUnixNano(audit.OccurredAt)
	if err != nil {
		return err
	}
	var result sql.Result
	if exists {
		result, err = connection.ExecContext(ctx.RequestContext, `UPDATE workflow_rag_application_runtime_assignments
SET record_version=?,assignment_state=?,assignment_digest=?,updated_at_unix_nano=?,sanitized_assignment_payload=?
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND assignment_id=? AND record_version=?`,
			assignment.RecordVersion, assignment.State, assignment.AssignmentDigest, updatedAt, string(assignmentPayload),
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, assignment.AssignmentID, expectedVersion)
	} else {
		result, err = connection.ExecContext(ctx.RequestContext, `INSERT INTO workflow_rag_application_runtime_assignments
(tenant_ref,workspace_id,application_id,owner_subject_ref,assignment_id,record_version,assignment_state,assignment_digest,updated_at_unix_nano,sanitized_assignment_payload)
VALUES (?,?,?,?,?,?,?,?,?,?) ON CONFLICT (tenant_ref,workspace_id,application_id,owner_subject_ref) DO NOTHING`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, assignment.AssignmentID,
			assignment.RecordVersion, assignment.State, assignment.AssignmentDigest, updatedAt, string(assignmentPayload))
	}
	if err != nil {
		return errWorkflowRAGApplicationStore
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errWorkflowRAGApplicationStore
	}
	if rowsAffected != 1 {
		return errWorkflowRAGApplicationVersionConflict
	}
	if _, err = connection.ExecContext(ctx.RequestContext, `INSERT INTO workflow_rag_application_runtime_events
(tenant_ref,workspace_id,application_id,owner_subject_ref,assignment_id,event_id,after_record_version,occurred_at_unix_nano,sanitized_event_payload)
VALUES (?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, assignment.AssignmentID, event.EventID, event.AfterRecordVersion, eventAt, string(eventPayload)); err != nil {
		return errWorkflowRAGApplicationStore
	}
	if _, err = connection.ExecContext(ctx.RequestContext, `INSERT INTO workflow_rag_application_runtime_audits
(tenant_ref,workspace_id,application_id,owner_subject_ref,assignment_id,audit_event_id,record_version,occurred_at_unix_nano,sanitized_audit_payload)
VALUES (?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, assignment.AssignmentID, audit.AuditEventID, audit.RecordVersion, auditAt, string(auditPayload)); err != nil {
		return errWorkflowRAGApplicationStore
	}
	if _, err = connection.ExecContext(ctx.RequestContext, "COMMIT"); err != nil {
		return errWorkflowRAGApplicationStore
	}
	committed = true
	return nil
}

func readSQLiteWorkflowRAGApplicationRuntimeEntry(ctx WorkflowRAGApplicationRuntimeContext, query sqliteWorkflowRAGApplicationRuntimeQueryer) (WorkflowRAGApplicationRuntimeAssignment, []WorkflowRAGApplicationRuntimeEvent, []WorkflowRAGApplicationRuntimeAudit, error) {
	var assignmentID, state, digest string
	var recordVersion int
	var updatedAt int64
	var payload []byte
	err := query.QueryRowContext(ctx.RequestContext, `SELECT assignment_id,record_version,assignment_state,assignment_digest,updated_at_unix_nano,sanitized_assignment_payload
FROM workflow_rag_application_runtime_assignments
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=?`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef).
		Scan(&assignmentID, &recordVersion, &state, &digest, &updatedAt, &payload)
	if errors.Is(err, sql.ErrNoRows) {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, errWorkflowRAGApplicationAssignmentNotFound
	}
	if err != nil {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, errWorkflowRAGApplicationStore
	}
	var assignment WorkflowRAGApplicationRuntimeAssignment
	if decodeWorkflowRAGApplicationRuntimeRecord(payload, &assignment) != nil || validateStoredWorkflowRAGApplicationAssignment(assignment, ctx) != nil {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, errWorkflowRAGApplicationStoreContract
	}
	decodedUpdatedAt, err := workflowRAGApplicationRuntimeUnixNano(assignment.UpdatedAt)
	if err != nil || assignmentID != assignment.AssignmentID || recordVersion != assignment.RecordVersion || state != assignment.State || digest != assignment.AssignmentDigest || updatedAt != decodedUpdatedAt {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, errWorkflowRAGApplicationStoreContract
	}
	events, err := readSQLiteWorkflowRAGApplicationRuntimeEvents(ctx, query, assignment.AssignmentID)
	if err != nil {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, err
	}
	audits, err := readSQLiteWorkflowRAGApplicationRuntimeAudits(ctx, query, assignment.AssignmentID)
	if err != nil {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, err
	}
	entry := workflowRAGApplicationRuntimeMemoryEntry{assignment: assignment, events: events, audits: audits}
	if validateWorkflowRAGApplicationRuntimeEntry(ctx, entry) != nil {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, errWorkflowRAGApplicationStoreContract
	}
	return assignment, events, audits, nil
}

func readSQLiteWorkflowRAGApplicationRuntimeEvents(ctx WorkflowRAGApplicationRuntimeContext, query sqliteWorkflowRAGApplicationRuntimeQueryer, assignmentID string) ([]WorkflowRAGApplicationRuntimeEvent, error) {
	rows, err := query.QueryContext(ctx.RequestContext, `SELECT event_id,after_record_version,occurred_at_unix_nano,sanitized_event_payload
FROM workflow_rag_application_runtime_events
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND assignment_id=?
ORDER BY after_record_version`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, assignmentID)
	if err != nil {
		return nil, errWorkflowRAGApplicationStore
	}
	defer rows.Close()
	events := make([]WorkflowRAGApplicationRuntimeEvent, 0)
	for rows.Next() {
		var eventID string
		var version int
		var occurredAt int64
		var payload []byte
		if rows.Scan(&eventID, &version, &occurredAt, &payload) != nil {
			return nil, errWorkflowRAGApplicationStore
		}
		var event WorkflowRAGApplicationRuntimeEvent
		if decodeWorkflowRAGApplicationRuntimeRecord(payload, &event) != nil || validateStoredWorkflowRAGApplicationEvent(event) != nil {
			return nil, errWorkflowRAGApplicationStoreContract
		}
		decodedAt, decodeErr := workflowRAGApplicationRuntimeUnixNano(event.OccurredAt)
		if decodeErr != nil || eventID != event.EventID || version != event.AfterRecordVersion || occurredAt != decodedAt {
			return nil, errWorkflowRAGApplicationStoreContract
		}
		events = append(events, event)
	}
	if rows.Err() != nil {
		return nil, errWorkflowRAGApplicationStore
	}
	return events, nil
}

func readSQLiteWorkflowRAGApplicationRuntimeAudits(ctx WorkflowRAGApplicationRuntimeContext, query sqliteWorkflowRAGApplicationRuntimeQueryer, assignmentID string) ([]WorkflowRAGApplicationRuntimeAudit, error) {
	rows, err := query.QueryContext(ctx.RequestContext, `SELECT audit_event_id,record_version,occurred_at_unix_nano,sanitized_audit_payload
FROM workflow_rag_application_runtime_audits
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND assignment_id=?
ORDER BY record_version`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, assignmentID)
	if err != nil {
		return nil, errWorkflowRAGApplicationStore
	}
	defer rows.Close()
	audits := make([]WorkflowRAGApplicationRuntimeAudit, 0)
	for rows.Next() {
		var auditID string
		var version int
		var occurredAt int64
		var payload []byte
		if rows.Scan(&auditID, &version, &occurredAt, &payload) != nil {
			return nil, errWorkflowRAGApplicationStore
		}
		var audit WorkflowRAGApplicationRuntimeAudit
		if decodeWorkflowRAGApplicationRuntimeRecord(payload, &audit) != nil || validateStoredWorkflowRAGApplicationAudit(audit, ctx) != nil {
			return nil, errWorkflowRAGApplicationStoreContract
		}
		decodedAt, decodeErr := workflowRAGApplicationRuntimeUnixNano(audit.OccurredAt)
		if decodeErr != nil || auditID != audit.AuditEventID || version != audit.RecordVersion || occurredAt != decodedAt {
			return nil, errWorkflowRAGApplicationStoreContract
		}
		audits = append(audits, audit)
	}
	if rows.Err() != nil {
		return nil, errWorkflowRAGApplicationStore
	}
	return audits, nil
}

var _ workflowRAGApplicationRuntimeRepository = (*sqliteWorkflowRAGApplicationRuntimeRepository)(nil)
