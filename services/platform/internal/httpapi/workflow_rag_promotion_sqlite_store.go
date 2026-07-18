package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type sqliteWorkflowRAGPromotionRepository struct{ database *sql.DB }

func newSQLiteWorkflowRAGPromotionRepository(database *sql.DB) *sqliteWorkflowRAGPromotionRepository {
	return &sqliteWorkflowRAGPromotionRepository{database: database}
}

func (repository *sqliteWorkflowRAGPromotionRepository) Create(ctx WorkflowRAGPromotionContext, candidate WorkflowRAGKnowledgePromotionCandidate, audit WorkflowRAGPromotionAudit) error {
	if repository == nil || repository.database == nil {
		return errWorkflowRAGPromotionStore
	}
	if validateWorkflowRAGPromotionCreate(ctx, candidate, audit) != nil {
		return errWorkflowRAGPromotionStoreContract
	}
	candidatePayload, err := encodeWorkflowRAGPromotionCandidate(candidate)
	if err != nil {
		return err
	}
	auditPayload, err := encodeWorkflowRAGPromotionAudit(audit)
	if err != nil {
		return err
	}
	createdAt, err := workflowRAGPromotionUnixNano(candidate.CreatedAt)
	if err != nil {
		return err
	}
	databaseContext := workflowRAGPromotionDatabaseContext(ctx)
	tx, err := repository.database.BeginTx(databaseContext, nil)
	if err != nil {
		return errWorkflowRAGPromotionStore
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(databaseContext, `INSERT INTO workflow_rag_knowledge_promotion_candidates (
	 tenant_ref,workspace_id,application_id,owner_subject_ref,candidate_id,candidate_digest,candidate_state,record_version,
	 binding_id,binding_version,binding_digest,created_at_unix_nano,updated_at_unix_nano,sanitized_candidate_payload
	) VALUES (?,?,?,?,?,?,?,?,NULL,NULL,NULL,?,?,?) ON CONFLICT DO NOTHING`, candidate.TenantRef, candidate.WorkspaceID,
		candidate.ApplicationID, candidate.OwnerSubjectRef, candidate.CandidateID, candidate.CandidateDigest, candidate.CandidateState,
		candidate.RecordVersion, createdAt, createdAt, string(candidatePayload))
	if err != nil {
		return errWorkflowRAGPromotionStore
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
		return errWorkflowRAGPromotionStoreContract
	}
	if err = insertSQLiteWorkflowRAGPromotionAudit(databaseContext, tx, audit, 1, auditPayload); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return errWorkflowRAGPromotionStore
	}
	return nil
}

func (repository *sqliteWorkflowRAGPromotionRepository) Read(ctx WorkflowRAGPromotionContext, candidateID string) (WorkflowRAGKnowledgePromotionCandidate, []WorkflowRAGKnowledgePromotionDecision, *WorkflowRAGApplicationBinding, []WorkflowRAGPromotionAudit, error) {
	if repository == nil || repository.database == nil {
		return WorkflowRAGKnowledgePromotionCandidate{}, nil, nil, nil, errWorkflowRAGPromotionStore
	}
	databaseContext := workflowRAGPromotionDatabaseContext(ctx)
	tx, err := repository.database.BeginTx(databaseContext, nil)
	if err != nil {
		return WorkflowRAGKnowledgePromotionCandidate{}, nil, nil, nil, errWorkflowRAGPromotionStore
	}
	defer func() { _ = tx.Rollback() }()
	entry, err := readSQLiteWorkflowRAGPromotionEntry(databaseContext, tx, ctx, candidateID)
	if err != nil {
		return WorkflowRAGKnowledgePromotionCandidate{}, nil, nil, nil, err
	}
	if err = tx.Commit(); err != nil {
		return WorkflowRAGKnowledgePromotionCandidate{}, nil, nil, nil, errWorkflowRAGPromotionStore
	}
	cloned := cloneWorkflowRAGPromotionMemoryEntry(entry)
	return cloned.candidate, cloned.decisions, cloned.binding, cloned.audits, nil
}

func (repository *sqliteWorkflowRAGPromotionRepository) List(ctx WorkflowRAGPromotionContext, query workflowRAGPromotionListQuery) ([]WorkflowRAGKnowledgePromotionCandidate, error) {
	if repository == nil || repository.database == nil || query.Limit < 1 {
		return nil, errWorkflowRAGPromotionStore
	}
	databaseContext := workflowRAGPromotionDatabaseContext(ctx)
	var rows *sql.Rows
	var err error
	if query.BeforeCreatedAt == "" {
		rows, err = repository.database.QueryContext(databaseContext, `SELECT sanitized_candidate_payload
		 FROM workflow_rag_knowledge_promotion_candidates
		 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=?
		 ORDER BY created_at_unix_nano DESC,candidate_id DESC LIMIT ?`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, query.Limit)
	} else {
		beforeAt, parseErr := workflowRAGPromotionUnixNano(query.BeforeCreatedAt)
		if parseErr != nil {
			return nil, errWorkflowRAGPromotionStoreContract
		}
		rows, err = repository.database.QueryContext(databaseContext, `SELECT sanitized_candidate_payload
		 FROM workflow_rag_knowledge_promotion_candidates
		 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=?
		 AND (created_at_unix_nano<? OR (created_at_unix_nano=? AND candidate_id<?))
		 ORDER BY created_at_unix_nano DESC,candidate_id DESC LIMIT ?`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID,
			ctx.OwnerSubjectRef, beforeAt, beforeAt, query.BeforeCandidateID, query.Limit)
	}
	if err != nil {
		return nil, errWorkflowRAGPromotionStore
	}
	defer rows.Close()
	candidates := make([]WorkflowRAGKnowledgePromotionCandidate, 0)
	for rows.Next() {
		var payload []byte
		if err = rows.Scan(&payload); err != nil {
			return nil, errWorkflowRAGPromotionStore
		}
		candidate, decodeErr := decodeWorkflowRAGPromotionCandidate(payload)
		if decodeErr != nil || validateStoredWorkflowRAGPromotionCandidate(candidate, ctx) != nil {
			return nil, errWorkflowRAGPromotionStoreContract
		}
		candidates = append(candidates, candidate)
	}
	if rows.Err() != nil {
		return nil, errWorkflowRAGPromotionStore
	}
	return candidates, nil
}

func (repository *sqliteWorkflowRAGPromotionRepository) AppendDecision(
	ctx WorkflowRAGPromotionContext,
	candidateID string,
	expectedVersion int,
	updated WorkflowRAGKnowledgePromotionCandidate,
	decision WorkflowRAGKnowledgePromotionDecision,
	binding *WorkflowRAGApplicationBinding,
	audits []WorkflowRAGPromotionAudit,
) error {
	if repository == nil || repository.database == nil {
		return errWorkflowRAGPromotionStore
	}
	currentCandidate, decisions, currentBinding, currentAudits, err := repository.Read(ctx, candidateID)
	if err != nil {
		return err
	}
	currentEntry := workflowRAGPromotionMemoryEntry{candidate: currentCandidate, decisions: decisions, binding: currentBinding, audits: currentAudits}
	if _, err = validateWorkflowRAGPromotionAppend(ctx, currentEntry, candidateID, expectedVersion, updated, decision, binding, audits); err != nil {
		return err
	}
	candidatePayload, err := encodeWorkflowRAGPromotionCandidate(updated)
	if err != nil {
		return err
	}
	decisionPayload, err := encodeWorkflowRAGPromotionDecision(decision)
	if err != nil {
		return err
	}
	updatedAt, err := workflowRAGPromotionUnixNano(updated.UpdatedAt)
	if err != nil {
		return err
	}
	decisionAt, err := workflowRAGPromotionUnixNano(decision.OccurredAt)
	if err != nil {
		return err
	}
	var bindingPayload []byte
	var issuedAt int64
	if binding != nil {
		bindingPayload, err = encodeWorkflowRAGApplicationBinding(*binding)
		if err != nil {
			return err
		}
		issuedAt, err = workflowRAGPromotionUnixNano(binding.IssuedAt)
		if err != nil {
			return err
		}
	}
	auditPayloads := make([][]byte, len(audits))
	for index, audit := range audits {
		auditPayloads[index], err = encodeWorkflowRAGPromotionAudit(audit)
		if err != nil {
			return err
		}
	}
	var bindingID, bindingVersion, bindingDigest any
	if updated.BindingRef != nil {
		bindingID, bindingVersion, bindingDigest = updated.BindingRef.BindingID, updated.BindingRef.BindingVersion, updated.BindingRef.BindingDigest
	}
	databaseContext := workflowRAGPromotionDatabaseContext(ctx)
	tx, err := repository.database.BeginTx(databaseContext, nil)
	if err != nil {
		return errWorkflowRAGPromotionStore
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(databaseContext, `UPDATE workflow_rag_knowledge_promotion_candidates SET
	 candidate_state=?,record_version=?,binding_id=?,binding_version=?,binding_digest=?,updated_at_unix_nano=?,sanitized_candidate_payload=?
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND candidate_id=? AND record_version=?`,
		updated.CandidateState, updated.RecordVersion, bindingID, bindingVersion, bindingDigest, updatedAt, string(candidatePayload),
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, candidateID, expectedVersion)
	if err != nil {
		return errWorkflowRAGPromotionStore
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
		return sqliteWorkflowRAGPromotionConflict(databaseContext, tx, ctx, candidateID)
	}
	if _, err = tx.ExecContext(databaseContext, `INSERT INTO workflow_rag_knowledge_promotion_decisions (
	 tenant_ref,workspace_id,application_id,owner_subject_ref,candidate_id,decision_id,decision,before_record_version,
	 after_record_version,occurred_at_unix_nano,sanitized_decision_payload
	) VALUES (?,?,?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, candidateID,
		decision.DecisionID, decision.Decision, decision.BeforeRecordVersion, decision.AfterRecordVersion, decisionAt, string(decisionPayload)); err != nil {
		return errWorkflowRAGPromotionStore
	}
	if binding != nil {
		if _, err = tx.ExecContext(databaseContext, `INSERT INTO workflow_rag_application_bindings (
		 tenant_ref,workspace_id,application_id,owner_subject_ref,candidate_id,binding_id,binding_version,binding_digest,
		 approved_decision_id,approved_record_version,issued_at_unix_nano,sanitized_binding_payload
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, candidateID,
			binding.BindingID, binding.BindingVersion, binding.BindingDigest, binding.ApprovedDecisionID, binding.ApprovedRecordVersion,
			issuedAt, string(bindingPayload)); err != nil {
			return errWorkflowRAGPromotionStore
		}
	}
	sequence := len(currentAudits)
	for index, audit := range audits {
		if err = insertSQLiteWorkflowRAGPromotionAudit(databaseContext, tx, audit, sequence+index+1, auditPayloads[index]); err != nil {
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return errWorkflowRAGPromotionStore
	}
	return nil
}

type sqliteWorkflowRAGPromotionReader interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func readSQLiteWorkflowRAGPromotionEntry(databaseContext context.Context, reader sqliteWorkflowRAGPromotionReader, ctx WorkflowRAGPromotionContext, candidateID string) (workflowRAGPromotionMemoryEntry, error) {
	var candidatePayload []byte
	err := reader.QueryRowContext(databaseContext, `SELECT sanitized_candidate_payload FROM workflow_rag_knowledge_promotion_candidates
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND candidate_id=?`, ctx.TenantRef, ctx.WorkspaceID,
		ctx.ApplicationID, ctx.OwnerSubjectRef, candidateID).Scan(&candidatePayload)
	if errors.Is(err, sql.ErrNoRows) {
		return workflowRAGPromotionMemoryEntry{}, sqliteWorkflowRAGPromotionMissing(databaseContext, reader, candidateID)
	}
	if err != nil {
		return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionStore
	}
	candidate, err := decodeWorkflowRAGPromotionCandidate(candidatePayload)
	if err != nil || validateStoredWorkflowRAGPromotionCandidate(candidate, ctx) != nil {
		return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionStoreContract
	}
	entry := workflowRAGPromotionMemoryEntry{candidate: candidate, decisions: []WorkflowRAGKnowledgePromotionDecision{}, audits: []WorkflowRAGPromotionAudit{}}
	rows, err := reader.QueryContext(databaseContext, `SELECT sanitized_decision_payload FROM workflow_rag_knowledge_promotion_decisions
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND candidate_id=? ORDER BY after_record_version ASC`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, candidateID)
	if err != nil {
		return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionStore
	}
	for rows.Next() {
		var payload []byte
		if err = rows.Scan(&payload); err != nil {
			_ = rows.Close()
			return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionStore
		}
		decision, decodeErr := decodeWorkflowRAGPromotionDecision(payload)
		if decodeErr != nil {
			_ = rows.Close()
			return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionStoreContract
		}
		entry.decisions = append(entry.decisions, decision)
	}
	if rows.Err() != nil {
		_ = rows.Close()
		return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionStore
	}
	_ = rows.Close()
	var bindingPayload []byte
	err = reader.QueryRowContext(databaseContext, `SELECT sanitized_binding_payload FROM workflow_rag_application_bindings
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND candidate_id=?`, ctx.TenantRef, ctx.WorkspaceID,
		ctx.ApplicationID, ctx.OwnerSubjectRef, candidateID).Scan(&bindingPayload)
	if err == nil {
		binding, decodeErr := decodeWorkflowRAGApplicationBinding(bindingPayload)
		if decodeErr != nil {
			return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionStoreContract
		}
		entry.binding = &binding
	} else if !errors.Is(err, sql.ErrNoRows) {
		return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionStore
	}
	rows, err = reader.QueryContext(databaseContext, `SELECT sanitized_audit_payload FROM workflow_rag_knowledge_promotion_audits
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND candidate_id=? ORDER BY event_sequence ASC`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, candidateID)
	if err != nil {
		return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionStore
	}
	defer rows.Close()
	for rows.Next() {
		var payload []byte
		if err = rows.Scan(&payload); err != nil {
			return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionStore
		}
		audit, decodeErr := decodeWorkflowRAGPromotionAudit(payload)
		if decodeErr != nil {
			return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionStoreContract
		}
		entry.audits = append(entry.audits, audit)
	}
	if rows.Err() != nil {
		return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionStore
	}
	if validateWorkflowRAGPromotionMemoryEntry(entry) != nil {
		return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionStoreContract
	}
	return entry, nil
}

func insertSQLiteWorkflowRAGPromotionAudit(databaseContext context.Context, tx *sql.Tx, audit WorkflowRAGPromotionAudit, sequence int, payload []byte) error {
	occurredAt, err := workflowRAGPromotionUnixNano(audit.OccurredAt)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(databaseContext, `INSERT INTO workflow_rag_knowledge_promotion_audits (
	 tenant_ref,workspace_id,application_id,owner_subject_ref,candidate_id,event_id,event_sequence,event_kind,record_version,
	 occurred_at_unix_nano,sanitized_audit_payload
	) VALUES (?,?,?,?,?,?,?,?,?,?,?)`, audit.TenantRef, audit.WorkspaceID, audit.ApplicationID, audit.OwnerSubjectRef, audit.CandidateID,
		audit.EventID, sequence, audit.EventKind, audit.RecordVersion, occurredAt, string(payload))
	if err != nil {
		return errWorkflowRAGPromotionStore
	}
	return nil
}

func sqliteWorkflowRAGPromotionMissing(databaseContext context.Context, reader sqliteWorkflowRAGPromotionReader, candidateID string) error {
	var count int
	if err := reader.QueryRowContext(databaseContext, `SELECT count(*) FROM workflow_rag_knowledge_promotion_candidates WHERE candidate_id=?`, candidateID).Scan(&count); err != nil {
		return errWorkflowRAGPromotionStore
	}
	if count > 0 {
		return errWorkflowRAGPromotionScopeDenied
	}
	return errWorkflowRAGPromotionNotFound
}

func sqliteWorkflowRAGPromotionConflict(databaseContext context.Context, reader sqliteWorkflowRAGPromotionReader, ctx WorkflowRAGPromotionContext, candidateID string) error {
	var version int
	var state string
	err := reader.QueryRowContext(databaseContext, `SELECT record_version,candidate_state FROM workflow_rag_knowledge_promotion_candidates
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND candidate_id=?`, ctx.TenantRef, ctx.WorkspaceID,
		ctx.ApplicationID, ctx.OwnerSubjectRef, candidateID).Scan(&version, &state)
	if errors.Is(err, sql.ErrNoRows) {
		return sqliteWorkflowRAGPromotionMissing(databaseContext, reader, candidateID)
	}
	if err != nil {
		return errWorkflowRAGPromotionStore
	}
	return workflowRAGPromotionConflictError{CurrentVersion: version, CurrentState: state}
}

func workflowRAGPromotionDatabaseContext(ctx WorkflowRAGPromotionContext) context.Context {
	if ctx.RequestContext != nil {
		return ctx.RequestContext
	}
	return context.Background()
}

func workflowRAGPromotionUnixNano(value string) (int64, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return 0, errWorkflowRAGPromotionStoreContract
	}
	return parsed.UnixNano(), nil
}
