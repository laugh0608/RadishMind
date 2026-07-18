package httpapi

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresWorkflowRAGPromotionRepository struct{ pool *pgxpool.Pool }

func newPostgresWorkflowRAGPromotionRepository(pool *pgxpool.Pool) *postgresWorkflowRAGPromotionRepository {
	return &postgresWorkflowRAGPromotionRepository{pool: pool}
}

func (repository *postgresWorkflowRAGPromotionRepository) Create(ctx WorkflowRAGPromotionContext, candidate WorkflowRAGKnowledgePromotionCandidate, audit WorkflowRAGPromotionAudit) error {
	if repository == nil || repository.pool == nil {
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
	createdAt, err := workflowRAGPromotionTime(candidate.CreatedAt)
	if err != nil {
		return err
	}
	databaseContext := workflowRAGPromotionDatabaseContext(ctx)
	tx, err := repository.pool.Begin(databaseContext)
	if err != nil {
		return errWorkflowRAGPromotionStore
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	result, err := tx.Exec(databaseContext, `INSERT INTO workflow_rag_knowledge_promotion_candidates (
	 tenant_ref,workspace_id,application_id,owner_subject_ref,candidate_id,candidate_digest,candidate_state,record_version,
	 binding_id,binding_version,binding_digest,created_at,updated_at,sanitized_candidate_payload
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NULL,NULL,NULL,$9,$9,$10) ON CONFLICT DO NOTHING`, candidate.TenantRef,
		candidate.WorkspaceID, candidate.ApplicationID, candidate.OwnerSubjectRef, candidate.CandidateID, candidate.CandidateDigest,
		candidate.CandidateState, candidate.RecordVersion, createdAt, candidatePayload)
	if err != nil {
		return errWorkflowRAGPromotionStore
	}
	if result.RowsAffected() != 1 {
		return errWorkflowRAGPromotionStoreContract
	}
	if err = insertPostgresWorkflowRAGPromotionAudit(databaseContext, tx, audit, 1, auditPayload); err != nil {
		return err
	}
	if err = tx.Commit(databaseContext); err != nil {
		return errWorkflowRAGPromotionStore
	}
	return nil
}

func (repository *postgresWorkflowRAGPromotionRepository) Read(ctx WorkflowRAGPromotionContext, candidateID string) (WorkflowRAGKnowledgePromotionCandidate, []WorkflowRAGKnowledgePromotionDecision, *WorkflowRAGApplicationBinding, []WorkflowRAGPromotionAudit, error) {
	if repository == nil || repository.pool == nil {
		return WorkflowRAGKnowledgePromotionCandidate{}, nil, nil, nil, errWorkflowRAGPromotionStore
	}
	databaseContext := workflowRAGPromotionDatabaseContext(ctx)
	tx, err := repository.pool.BeginTx(databaseContext, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return WorkflowRAGKnowledgePromotionCandidate{}, nil, nil, nil, errWorkflowRAGPromotionStore
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	entry, err := readPostgresWorkflowRAGPromotionEntry(databaseContext, tx, ctx, candidateID)
	if err != nil {
		return WorkflowRAGKnowledgePromotionCandidate{}, nil, nil, nil, err
	}
	if err = tx.Commit(databaseContext); err != nil {
		return WorkflowRAGKnowledgePromotionCandidate{}, nil, nil, nil, errWorkflowRAGPromotionStore
	}
	cloned := cloneWorkflowRAGPromotionMemoryEntry(entry)
	return cloned.candidate, cloned.decisions, cloned.binding, cloned.audits, nil
}

func (repository *postgresWorkflowRAGPromotionRepository) List(ctx WorkflowRAGPromotionContext, query workflowRAGPromotionListQuery) ([]WorkflowRAGKnowledgePromotionCandidate, error) {
	if repository == nil || repository.pool == nil || query.Limit < 1 {
		return nil, errWorkflowRAGPromotionStore
	}
	databaseContext := workflowRAGPromotionDatabaseContext(ctx)
	var rows pgx.Rows
	var err error
	if query.BeforeCreatedAt == "" {
		rows, err = repository.pool.Query(databaseContext, `SELECT sanitized_candidate_payload
		 FROM workflow_rag_knowledge_promotion_candidates
		 WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4
		 ORDER BY created_at DESC,candidate_id DESC LIMIT $5`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, query.Limit)
	} else {
		beforeAt, parseErr := time.Parse(time.RFC3339Nano, query.BeforeCreatedAt)
		if parseErr != nil {
			return nil, errWorkflowRAGPromotionStoreContract
		}
		rows, err = repository.pool.Query(databaseContext, `SELECT sanitized_candidate_payload
		 FROM workflow_rag_knowledge_promotion_candidates
		 WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4
		 AND (created_at<$5 OR (created_at=$5 AND candidate_id<$6))
		 ORDER BY created_at DESC,candidate_id DESC LIMIT $7`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID,
			ctx.OwnerSubjectRef, beforeAt, query.BeforeCandidateID, query.Limit)
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

func (repository *postgresWorkflowRAGPromotionRepository) AppendDecision(
	ctx WorkflowRAGPromotionContext,
	candidateID string,
	expectedVersion int,
	updated WorkflowRAGKnowledgePromotionCandidate,
	decision WorkflowRAGKnowledgePromotionDecision,
	binding *WorkflowRAGApplicationBinding,
	audits []WorkflowRAGPromotionAudit,
) error {
	if repository == nil || repository.pool == nil {
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
	updatedAt, err := workflowRAGPromotionTime(updated.UpdatedAt)
	if err != nil {
		return err
	}
	decisionAt, err := workflowRAGPromotionTime(decision.OccurredAt)
	if err != nil {
		return err
	}
	var bindingPayload []byte
	var issuedAt time.Time
	if binding != nil {
		bindingPayload, err = encodeWorkflowRAGApplicationBinding(*binding)
		if err != nil {
			return err
		}
		issuedAt, err = workflowRAGPromotionTime(binding.IssuedAt)
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
	tx, err := repository.pool.Begin(databaseContext)
	if err != nil {
		return errWorkflowRAGPromotionStore
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var storedVersion int
	err = tx.QueryRow(databaseContext, `UPDATE workflow_rag_knowledge_promotion_candidates SET
	 candidate_state=$1,record_version=$2,binding_id=$3,binding_version=$4,binding_digest=$5,updated_at=$6,sanitized_candidate_payload=$7
	 WHERE tenant_ref=$8 AND workspace_id=$9 AND application_id=$10 AND owner_subject_ref=$11 AND candidate_id=$12 AND record_version=$13
	 RETURNING record_version`, updated.CandidateState, updated.RecordVersion, bindingID, bindingVersion, bindingDigest, updatedAt, candidatePayload,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, candidateID, expectedVersion).Scan(&storedVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return postgresWorkflowRAGPromotionConflict(databaseContext, tx, ctx, candidateID)
	}
	if err != nil || storedVersion != updated.RecordVersion {
		return errWorkflowRAGPromotionStore
	}
	if _, err = tx.Exec(databaseContext, `INSERT INTO workflow_rag_knowledge_promotion_decisions (
	 tenant_ref,workspace_id,application_id,owner_subject_ref,candidate_id,decision_id,decision,before_record_version,
	 after_record_version,occurred_at,sanitized_decision_payload
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef,
		candidateID, decision.DecisionID, decision.Decision, decision.BeforeRecordVersion, decision.AfterRecordVersion, decisionAt, decisionPayload); err != nil {
		return errWorkflowRAGPromotionStore
	}
	if binding != nil {
		if _, err = tx.Exec(databaseContext, `INSERT INTO workflow_rag_application_bindings (
		 tenant_ref,workspace_id,application_id,owner_subject_ref,candidate_id,binding_id,binding_version,binding_digest,
		 approved_decision_id,approved_record_version,issued_at,sanitized_binding_payload
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID,
			ctx.OwnerSubjectRef, candidateID, binding.BindingID, binding.BindingVersion, binding.BindingDigest, binding.ApprovedDecisionID,
			binding.ApprovedRecordVersion, issuedAt, bindingPayload); err != nil {
			return errWorkflowRAGPromotionStore
		}
	}
	sequence := len(currentAudits)
	for index, audit := range audits {
		if err = insertPostgresWorkflowRAGPromotionAudit(databaseContext, tx, audit, sequence+index+1, auditPayloads[index]); err != nil {
			return err
		}
	}
	if err = tx.Commit(databaseContext); err != nil {
		return errWorkflowRAGPromotionStore
	}
	return nil
}

type postgresWorkflowRAGPromotionReader interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func readPostgresWorkflowRAGPromotionEntry(databaseContext context.Context, reader postgresWorkflowRAGPromotionReader, ctx WorkflowRAGPromotionContext, candidateID string) (workflowRAGPromotionMemoryEntry, error) {
	var candidatePayload []byte
	err := reader.QueryRow(databaseContext, `SELECT sanitized_candidate_payload FROM workflow_rag_knowledge_promotion_candidates
	 WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND candidate_id=$5`, ctx.TenantRef,
		ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, candidateID).Scan(&candidatePayload)
	if errors.Is(err, pgx.ErrNoRows) {
		return workflowRAGPromotionMemoryEntry{}, postgresWorkflowRAGPromotionMissing(databaseContext, reader, candidateID)
	}
	if err != nil {
		return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionStore
	}
	candidate, err := decodeWorkflowRAGPromotionCandidate(candidatePayload)
	if err != nil || validateStoredWorkflowRAGPromotionCandidate(candidate, ctx) != nil {
		return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionStoreContract
	}
	entry := workflowRAGPromotionMemoryEntry{candidate: candidate, decisions: []WorkflowRAGKnowledgePromotionDecision{}, audits: []WorkflowRAGPromotionAudit{}}
	rows, err := reader.Query(databaseContext, `SELECT sanitized_decision_payload FROM workflow_rag_knowledge_promotion_decisions
	 WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND candidate_id=$5 ORDER BY after_record_version ASC`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, candidateID)
	if err != nil {
		return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionStore
	}
	for rows.Next() {
		var payload []byte
		if err = rows.Scan(&payload); err != nil {
			rows.Close()
			return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionStore
		}
		decision, decodeErr := decodeWorkflowRAGPromotionDecision(payload)
		if decodeErr != nil {
			rows.Close()
			return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionStoreContract
		}
		entry.decisions = append(entry.decisions, decision)
	}
	if rows.Err() != nil {
		rows.Close()
		return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionStore
	}
	rows.Close()
	var bindingPayload []byte
	err = reader.QueryRow(databaseContext, `SELECT sanitized_binding_payload FROM workflow_rag_application_bindings
	 WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND candidate_id=$5`, ctx.TenantRef,
		ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, candidateID).Scan(&bindingPayload)
	if err == nil {
		binding, decodeErr := decodeWorkflowRAGApplicationBinding(bindingPayload)
		if decodeErr != nil {
			return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionStoreContract
		}
		entry.binding = &binding
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionStore
	}
	rows, err = reader.Query(databaseContext, `SELECT sanitized_audit_payload FROM workflow_rag_knowledge_promotion_audits
	 WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND candidate_id=$5 ORDER BY event_sequence ASC`,
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

func insertPostgresWorkflowRAGPromotionAudit(databaseContext context.Context, tx pgx.Tx, audit WorkflowRAGPromotionAudit, sequence int, payload []byte) error {
	occurredAt, err := workflowRAGPromotionTime(audit.OccurredAt)
	if err != nil {
		return err
	}
	_, err = tx.Exec(databaseContext, `INSERT INTO workflow_rag_knowledge_promotion_audits (
	 tenant_ref,workspace_id,application_id,owner_subject_ref,candidate_id,event_id,event_sequence,event_kind,record_version,
	 occurred_at,sanitized_audit_payload
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, audit.TenantRef, audit.WorkspaceID, audit.ApplicationID,
		audit.OwnerSubjectRef, audit.CandidateID, audit.EventID, sequence, audit.EventKind, audit.RecordVersion, occurredAt, payload)
	if err != nil {
		return errWorkflowRAGPromotionStore
	}
	return nil
}

func workflowRAGPromotionTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, errWorkflowRAGPromotionStoreContract
	}
	return parsed, nil
}

func postgresWorkflowRAGPromotionMissing(databaseContext context.Context, reader postgresWorkflowRAGPromotionReader, candidateID string) error {
	var count int
	if err := reader.QueryRow(databaseContext, `SELECT count(*) FROM workflow_rag_knowledge_promotion_candidates WHERE candidate_id=$1`, candidateID).Scan(&count); err != nil {
		return errWorkflowRAGPromotionStore
	}
	if count > 0 {
		return errWorkflowRAGPromotionScopeDenied
	}
	return errWorkflowRAGPromotionNotFound
}

func postgresWorkflowRAGPromotionConflict(databaseContext context.Context, reader postgresWorkflowRAGPromotionReader, ctx WorkflowRAGPromotionContext, candidateID string) error {
	var version int
	var state string
	err := reader.QueryRow(databaseContext, `SELECT record_version,candidate_state FROM workflow_rag_knowledge_promotion_candidates
	 WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND candidate_id=$5`, ctx.TenantRef,
		ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, candidateID).Scan(&version, &state)
	if errors.Is(err, pgx.ErrNoRows) {
		return postgresWorkflowRAGPromotionMissing(databaseContext, reader, candidateID)
	}
	if err != nil {
		return errWorkflowRAGPromotionStore
	}
	return workflowRAGPromotionConflictError{CurrentVersion: version, CurrentState: state}
}
