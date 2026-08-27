package httpapi

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresWorkflowTemplateCatalogRepository struct{ pool *pgxpool.Pool }

func newPostgresWorkflowTemplateCatalogRepository(pool *pgxpool.Pool) *postgresWorkflowTemplateCatalogRepository {
	return &postgresWorkflowTemplateCatalogRepository{pool: pool}
}

type postgresWorkflowTemplateQueryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func (repository *postgresWorkflowTemplateCatalogRepository) CreateCandidate(ctx WorkflowTemplateCatalogContext, candidate WorkflowTemplateCandidate, now time.Time) (WorkflowTemplateCandidate, error) {
	var output WorkflowTemplateCandidate
	err := repository.mutate(ctx, func(tx pgx.Tx, store *memoryWorkflowTemplateCatalogRepository) error {
		beforeAuditCount := len(store.audits[workflowTemplateScopeKey(ctx, "audits")])
		created, err := store.CreateCandidate(ctx, candidate, now)
		if err != nil {
			return err
		}
		payload, err := encodeWorkflowTemplateRecord(created)
		if err != nil {
			return err
		}
		createdAt, _ := time.Parse(time.RFC3339Nano, created.CreatedAt)
		if _, err = tx.Exec(workflowTemplateRequestContext(ctx), `INSERT INTO workflow_template_candidates
(tenant_ref,workspace_id,candidate_id,template_id,candidate_state,review_version,source_application_id,source_owner_subject_ref,source_definition_id,source_definition_version,source_definition_digest,created_at,updated_at,sanitized_candidate_payload)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$12,$13)`, ctx.TenantRef, ctx.WorkspaceID, created.CandidateID, created.TemplateID, created.State, created.ReviewVersion, created.SourceApplicationID, created.SourceOwnerSubjectRef, created.SourceDefinitionID, created.SourceDefinitionVersion, created.SourceDefinitionDigest, createdAt, string(payload)); err != nil {
			return postgresWorkflowTemplateMutationError(err, errWorkflowTemplateCandidateConflict)
		}
		if err = insertPostgresWorkflowTemplateNewAudits(tx, ctx, store, beforeAuditCount); err != nil {
			return err
		}
		output = created
		return nil
	})
	return output, err
}

func (repository *postgresWorkflowTemplateCatalogRepository) ReviewCandidate(ctx WorkflowTemplateCatalogContext, candidateID string, expected int, decision, reason, sourceDigest string, now time.Time) (WorkflowTemplateCandidate, *WorkflowTemplateVersion, error) {
	var output WorkflowTemplateCandidate
	var outputVersion *WorkflowTemplateVersion
	err := repository.mutate(ctx, func(tx pgx.Tx, store *memoryWorkflowTemplateCatalogRepository) error {
		beforeAuditCount := len(store.audits[workflowTemplateScopeKey(ctx, "audits")])
		updated, version, err := store.ReviewCandidate(ctx, candidateID, expected, decision, reason, sourceDigest, now)
		if err != nil {
			return err
		}
		payload, err := encodeWorkflowTemplateRecord(updated)
		if err != nil {
			return err
		}
		updatedAt, _ := time.Parse(time.RFC3339Nano, updated.UpdatedAt)
		command, err := tx.Exec(workflowTemplateRequestContext(ctx), `UPDATE workflow_template_candidates
SET candidate_state=$1,review_version=$2,updated_at=$3,sanitized_candidate_payload=$4
WHERE tenant_ref=$5 AND workspace_id=$6 AND candidate_id=$7 AND review_version=$8 AND candidate_state='pending'`, updated.State, updated.ReviewVersion, updatedAt, string(payload), ctx.TenantRef, ctx.WorkspaceID, candidateID, expected)
		if err != nil {
			return postgresWorkflowTemplateMutationError(err, errWorkflowTemplateCandidateConflict)
		}
		if command.RowsAffected() != 1 {
			return errWorkflowTemplateCandidateConflict
		}
		review := updated.Decisions[len(updated.Decisions)-1]
		reviewPayload, err := encodeWorkflowTemplateRecord(review)
		if err != nil {
			return err
		}
		decidedAt, _ := time.Parse(time.RFC3339Nano, review.DecidedAt)
		if _, err = tx.Exec(workflowTemplateRequestContext(ctx), `INSERT INTO workflow_template_decisions
(tenant_ref,workspace_id,candidate_id,review_version,decision,decided_at,sanitized_decision_payload)
VALUES ($1,$2,$3,$4,$5,$6,$7)`, ctx.TenantRef, ctx.WorkspaceID, candidateID, review.ReviewVersion, review.Decision, decidedAt, string(reviewPayload)); err != nil {
			return postgresWorkflowTemplateMutationError(err, errWorkflowTemplateCandidateConflict)
		}
		if version != nil {
			versionPayload, encodeErr := encodeWorkflowTemplateRecord(*version)
			if encodeErr != nil {
				return encodeErr
			}
			createdAt, _ := time.Parse(time.RFC3339Nano, version.CreatedAt)
			if _, err = tx.Exec(workflowTemplateRequestContext(ctx), `INSERT INTO workflow_template_versions
(tenant_ref,workspace_id,template_id,template_version,template_digest,candidate_id,candidate_review_version,source_definition_id,source_definition_version,source_definition_digest,created_at,sanitized_version_payload)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, ctx.TenantRef, ctx.WorkspaceID, version.TemplateID, version.Version, version.TemplateDigest, version.CandidateID, version.CandidateReviewVersion, version.SourceDefinitionID, version.SourceDefinitionVersion, version.SourceDefinitionDigest, createdAt, string(versionPayload)); err != nil {
				return postgresWorkflowTemplateMutationError(err, errWorkflowTemplateCandidateConflict)
			}
			if version.Version == 1 {
				lineage := store.lineages[workflowTemplateScopeKey(ctx, version.TemplateID)]
				lineagePayload, encodeErr := encodeWorkflowTemplateRecord(lineage)
				if encodeErr != nil {
					return encodeErr
				}
				lineageCreatedAt, _ := time.Parse(time.RFC3339Nano, lineage.CreatedAt)
				if _, err = tx.Exec(workflowTemplateRequestContext(ctx), `INSERT INTO workflow_template_lineages
(tenant_ref,workspace_id,template_id,pointer_version,lifecycle,listed_version,listed_digest,created_at,updated_at,sanitized_lineage_payload)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$8,$9)`, ctx.TenantRef, ctx.WorkspaceID, lineage.TemplateID, lineage.PointerVersion, lineage.Lifecycle, lineage.ListedVersion, lineage.ListedDigest, lineageCreatedAt, string(lineagePayload)); err != nil {
					return postgresWorkflowTemplateMutationError(err, errWorkflowTemplateCandidateConflict)
				}
			}
			outputVersion = cloneWorkflowTemplateVersionPointer(*version)
		}
		if err = insertPostgresWorkflowTemplateNewAudits(tx, ctx, store, beforeAuditCount); err != nil {
			return err
		}
		output = updated
		return nil
	})
	return output, outputVersion, err
}

func (repository *postgresWorkflowTemplateCatalogRepository) DecideListing(ctx WorkflowTemplateCatalogContext, templateID string, expected int, decision string, version int, reason string, now time.Time) (WorkflowTemplateLineage, error) {
	var output WorkflowTemplateLineage
	err := repository.mutate(ctx, func(tx pgx.Tx, store *memoryWorkflowTemplateCatalogRepository) error {
		beforeAuditCount := len(store.audits[workflowTemplateScopeKey(ctx, "audits")])
		lineage, err := store.DecideListing(ctx, templateID, expected, decision, version, reason, now)
		if err != nil {
			return err
		}
		payload, err := encodeWorkflowTemplateRecord(lineage)
		if err != nil {
			return err
		}
		updatedAt, _ := time.Parse(time.RFC3339Nano, lineage.UpdatedAt)
		command, err := tx.Exec(workflowTemplateRequestContext(ctx), `UPDATE workflow_template_lineages
SET pointer_version=$1,lifecycle=$2,listed_version=$3,listed_digest=$4,updated_at=$5,sanitized_lineage_payload=$6
WHERE tenant_ref=$7 AND workspace_id=$8 AND template_id=$9 AND pointer_version=$10`, lineage.PointerVersion, lineage.Lifecycle, lineage.ListedVersion, lineage.ListedDigest, updatedAt, string(payload), ctx.TenantRef, ctx.WorkspaceID, templateID, expected)
		if err != nil {
			return postgresWorkflowTemplateMutationError(err, errWorkflowTemplatePointerConflict)
		}
		if command.RowsAffected() != 1 {
			return errWorkflowTemplatePointerConflict
		}
		event := lineage.Events[len(lineage.Events)-1]
		eventPayload, err := encodeWorkflowTemplateRecord(event)
		if err != nil {
			return err
		}
		occurredAt, _ := time.Parse(time.RFC3339Nano, event.CreatedAt)
		if _, err = tx.Exec(workflowTemplateRequestContext(ctx), `INSERT INTO workflow_template_listing_events
(tenant_ref,workspace_id,template_id,event_id,after_pointer_version,decision,occurred_at,sanitized_event_payload)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, ctx.TenantRef, ctx.WorkspaceID, templateID, event.EventID, event.AfterPointerVersion, event.Decision, occurredAt, string(eventPayload)); err != nil {
			return postgresWorkflowTemplateMutationError(err, errWorkflowTemplatePointerConflict)
		}
		if err = insertPostgresWorkflowTemplateNewAudits(tx, ctx, store, beforeAuditCount); err != nil {
			return err
		}
		output = lineage
		return nil
	})
	return output, err
}

func (repository *postgresWorkflowTemplateCatalogRepository) ReadCandidate(ctx WorkflowTemplateCatalogContext, candidateID string) (WorkflowTemplateCandidate, error) {
	store, done, err := repository.readStore(ctx)
	if err != nil {
		return WorkflowTemplateCandidate{}, err
	}
	defer done()
	return store.ReadCandidate(ctx, candidateID)
}
func (repository *postgresWorkflowTemplateCatalogRepository) ListCandidates(ctx WorkflowTemplateCatalogContext) ([]WorkflowTemplateCandidate, error) {
	store, done, err := repository.readStore(ctx)
	if err != nil {
		return nil, err
	}
	defer done()
	return store.ListCandidates(ctx)
}
func (repository *postgresWorkflowTemplateCatalogRepository) ReadLineage(ctx WorkflowTemplateCatalogContext, templateID string) (WorkflowTemplateLineage, error) {
	store, done, err := repository.readStore(ctx)
	if err != nil {
		return WorkflowTemplateLineage{}, err
	}
	defer done()
	return store.ReadLineage(ctx, templateID)
}
func (repository *postgresWorkflowTemplateCatalogRepository) ListLineages(ctx WorkflowTemplateCatalogContext) ([]WorkflowTemplateLineage, error) {
	store, done, err := repository.readStore(ctx)
	if err != nil {
		return nil, err
	}
	defer done()
	return store.ListLineages(ctx)
}
func (repository *postgresWorkflowTemplateCatalogRepository) ReadVersion(ctx WorkflowTemplateCatalogContext, templateID string, version int) (WorkflowTemplateVersion, error) {
	store, done, err := repository.readStore(ctx)
	if err != nil {
		return WorkflowTemplateVersion{}, err
	}
	defer done()
	return store.ReadVersion(ctx, templateID, version)
}
func (repository *postgresWorkflowTemplateCatalogRepository) ListVersions(ctx WorkflowTemplateCatalogContext, templateID string) ([]WorkflowTemplateVersion, error) {
	store, done, err := repository.readStore(ctx)
	if err != nil {
		return nil, err
	}
	defer done()
	return store.ListVersions(ctx, templateID)
}

func (repository *postgresWorkflowTemplateCatalogRepository) mutate(ctx WorkflowTemplateCatalogContext, operation func(pgx.Tx, *memoryWorkflowTemplateCatalogRepository) error) error {
	if repository == nil || repository.pool == nil || !validWorkflowTemplateContext(ctx) {
		return errWorkflowTemplateStoreUnavailable
	}
	requestContext := workflowTemplateRequestContext(ctx)
	tx, err := repository.pool.BeginTx(requestContext, pgx.TxOptions{})
	if err != nil {
		return errWorkflowTemplateStoreUnavailable
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err = tx.Exec(requestContext, `SELECT pg_advisory_xact_lock(hashtextextended($1,hashtextextended($2,0)))`, ctx.TenantRef, ctx.WorkspaceID); err != nil {
		return postgresWorkflowTemplateMutationError(err, nil)
	}
	store, err := loadPostgresWorkflowTemplateCatalogStore(ctx, tx)
	if err != nil {
		return err
	}
	if err = operation(tx, store); err != nil {
		return err
	}
	if err = tx.Commit(requestContext); err != nil {
		return errWorkflowTemplateStoreUnavailable
	}
	return nil
}

func (repository *postgresWorkflowTemplateCatalogRepository) readStore(ctx WorkflowTemplateCatalogContext) (*memoryWorkflowTemplateCatalogRepository, func(), error) {
	if repository == nil || repository.pool == nil || !validWorkflowTemplateContext(ctx) {
		return nil, func() {}, errWorkflowTemplateStoreUnavailable
	}
	tx, err := repository.pool.BeginTx(workflowTemplateRequestContext(ctx), pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, func() {}, errWorkflowTemplateStoreUnavailable
	}
	store, err := loadPostgresWorkflowTemplateCatalogStore(ctx, tx)
	if err != nil {
		_ = tx.Rollback(context.Background())
		return nil, func() {}, err
	}
	return store, func() { _ = tx.Rollback(context.Background()) }, nil
}

func loadPostgresWorkflowTemplateCatalogStore(ctx WorkflowTemplateCatalogContext, query postgresWorkflowTemplateQueryer) (*memoryWorkflowTemplateCatalogRepository, error) {
	store := newMemoryWorkflowTemplateCatalogRepository()
	requestContext := workflowTemplateRequestContext(ctx)
	scope := []any{ctx.TenantRef, ctx.WorkspaceID}
	rows, err := query.Query(requestContext, `SELECT candidate_id,template_id,candidate_state,review_version,source_application_id,source_owner_subject_ref,source_definition_id,source_definition_version,source_definition_digest,created_at,updated_at,sanitized_candidate_payload
FROM workflow_template_candidates WHERE tenant_ref=$1 AND workspace_id=$2 ORDER BY created_at,candidate_id`, scope...)
	if err != nil {
		return nil, postgresWorkflowTemplateMutationError(err, nil)
	}
	for rows.Next() {
		var candidateID, templateID, state, applicationID, owner, definitionID, definitionDigest string
		var reviewVersion, definitionVersion int
		var createdAt, updatedAt time.Time
		var payload []byte
		if rows.Scan(&candidateID, &templateID, &state, &reviewVersion, &applicationID, &owner, &definitionID, &definitionVersion, &definitionDigest, &createdAt, &updatedAt, &payload) != nil {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
		value, loadErr := loadWorkflowTemplateCandidateRecord(store, ctx, payload)
		if loadErr != nil || workflowTemplateStoredFieldMismatch(value.CandidateID == candidateID, value.TemplateID == templateID, value.State == state, value.ReviewVersion == reviewVersion, value.SourceApplicationID == applicationID, value.SourceOwnerSubjectRef == owner, value.SourceDefinitionID == definitionID, value.SourceDefinitionVersion == definitionVersion, value.SourceDefinitionDigest == definitionDigest, workflowTemplatePostgresTimeMatches(createdAt, value.CreatedAt), workflowTemplatePostgresTimeMatches(updatedAt, value.UpdatedAt)) != nil {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
	}
	if rows.Err() != nil {
		rows.Close()
		return nil, errWorkflowTemplateStoreUnavailable
	}
	rows.Close()
	rows, err = query.Query(requestContext, `SELECT template_id,template_version,template_digest,candidate_id,candidate_review_version,source_definition_id,source_definition_version,source_definition_digest,created_at,sanitized_version_payload
FROM workflow_template_versions WHERE tenant_ref=$1 AND workspace_id=$2 ORDER BY template_id,template_version`, scope...)
	if err != nil {
		return nil, postgresWorkflowTemplateMutationError(err, nil)
	}
	for rows.Next() {
		var templateID, digest, candidateID, definitionID, definitionDigest string
		var version, reviewVersion, definitionVersion int
		var createdAt time.Time
		var payload []byte
		if rows.Scan(&templateID, &version, &digest, &candidateID, &reviewVersion, &definitionID, &definitionVersion, &definitionDigest, &createdAt, &payload) != nil {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
		value, loadErr := loadWorkflowTemplateVersionRecord(store, ctx, payload)
		if loadErr != nil || workflowTemplateStoredFieldMismatch(value.TemplateID == templateID, value.Version == version, value.TemplateDigest == digest, value.CandidateID == candidateID, value.CandidateReviewVersion == reviewVersion, value.SourceDefinitionID == definitionID, value.SourceDefinitionVersion == definitionVersion, value.SourceDefinitionDigest == definitionDigest, workflowTemplatePostgresTimeMatches(createdAt, value.CreatedAt)) != nil {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
	}
	if rows.Err() != nil {
		rows.Close()
		return nil, errWorkflowTemplateStoreUnavailable
	}
	rows.Close()
	rows, err = query.Query(requestContext, `SELECT template_id,pointer_version,lifecycle,listed_version,listed_digest,created_at,updated_at,sanitized_lineage_payload
FROM workflow_template_lineages WHERE tenant_ref=$1 AND workspace_id=$2 ORDER BY template_id`, scope...)
	if err != nil {
		return nil, postgresWorkflowTemplateMutationError(err, nil)
	}
	for rows.Next() {
		var templateID, lifecycle, digest string
		var pointerVersion, listedVersion int
		var createdAt, updatedAt time.Time
		var payload []byte
		if rows.Scan(&templateID, &pointerVersion, &lifecycle, &listedVersion, &digest, &createdAt, &updatedAt, &payload) != nil {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
		value, loadErr := loadWorkflowTemplateLineageRecord(store, ctx, payload)
		if loadErr != nil || workflowTemplateStoredFieldMismatch(value.TemplateID == templateID, value.PointerVersion == pointerVersion, value.Lifecycle == lifecycle, value.ListedVersion == listedVersion, value.ListedDigest == digest, workflowTemplatePostgresTimeMatches(createdAt, value.CreatedAt), workflowTemplatePostgresTimeMatches(updatedAt, value.UpdatedAt)) != nil {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
	}
	if rows.Err() != nil {
		rows.Close()
		return nil, errWorkflowTemplateStoreUnavailable
	}
	rows.Close()
	rows, err = query.Query(requestContext, `SELECT audit_id,audit_sequence,resource_kind,resource_id,action,occurred_at,sanitized_audit_payload
FROM workflow_template_audits WHERE tenant_ref=$1 AND workspace_id=$2 ORDER BY audit_sequence`, scope...)
	if err != nil {
		return nil, postgresWorkflowTemplateMutationError(err, nil)
	}
	for rows.Next() {
		var auditID, kind, resourceID, action string
		var sequence int
		var occurredAt time.Time
		var payload []byte
		if rows.Scan(&auditID, &sequence, &kind, &resourceID, &action, &occurredAt, &payload) != nil {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
		value, loadErr := loadWorkflowTemplateAuditRecord(store, ctx, sequence, payload)
		if loadErr != nil || workflowTemplateStoredFieldMismatch(value.AuditID == auditID, value.ResourceKind == kind, value.ResourceID == resourceID, value.Action == action, workflowTemplatePostgresTimeMatches(occurredAt, value.CreatedAt)) != nil {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
	}
	if rows.Err() != nil {
		rows.Close()
		return nil, errWorkflowTemplateStoreUnavailable
	}
	rows.Close()
	decisions := map[string]int{}
	rows, err = query.Query(requestContext, `SELECT candidate_id,review_version,decision,decided_at,sanitized_decision_payload
FROM workflow_template_decisions WHERE tenant_ref=$1 AND workspace_id=$2 ORDER BY candidate_id,review_version`, scope...)
	if err != nil {
		return nil, postgresWorkflowTemplateMutationError(err, nil)
	}
	for rows.Next() {
		var candidateID, decision string
		var reviewVersion int
		var decidedAt time.Time
		var payload []byte
		if rows.Scan(&candidateID, &reviewVersion, &decision, &decidedAt, &payload) != nil || validateWorkflowTemplateDecisionEvidence(store, ctx, candidateID, reviewVersion, payload) != nil {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
		var decoded WorkflowTemplateReviewDecision
		_ = decodeWorkflowTemplateRecord(payload, &decoded)
		if decoded.Decision != decision || !workflowTemplatePostgresTimeMatches(decidedAt, decoded.DecidedAt) {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
		decisions[candidateID]++
	}
	if rows.Err() != nil {
		rows.Close()
		return nil, errWorkflowTemplateStoreUnavailable
	}
	rows.Close()
	events := map[string]int{}
	rows, err = query.Query(requestContext, `SELECT template_id,event_id,after_pointer_version,decision,occurred_at,sanitized_event_payload
FROM workflow_template_listing_events WHERE tenant_ref=$1 AND workspace_id=$2 ORDER BY template_id,after_pointer_version`, scope...)
	if err != nil {
		return nil, postgresWorkflowTemplateMutationError(err, nil)
	}
	for rows.Next() {
		var templateID, eventID, decision string
		var pointerVersion int
		var occurredAt time.Time
		var payload []byte
		if rows.Scan(&templateID, &eventID, &pointerVersion, &decision, &occurredAt, &payload) != nil || validateWorkflowTemplateListingEventEvidence(store, ctx, templateID, pointerVersion, payload) != nil {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
		var decoded WorkflowTemplateListingEvent
		_ = decodeWorkflowTemplateRecord(payload, &decoded)
		if decoded.EventID != eventID || decoded.Decision != decision || !workflowTemplatePostgresTimeMatches(occurredAt, decoded.CreatedAt) {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
		events[templateID]++
	}
	if rows.Err() != nil {
		rows.Close()
		return nil, errWorkflowTemplateStoreUnavailable
	}
	rows.Close()
	if validateWorkflowTemplateEvidenceCounts(store, ctx, decisions, events) != nil {
		return nil, errWorkflowTemplateStoreUnavailable
	}
	return store, nil
}

func insertPostgresWorkflowTemplateNewAudits(tx pgx.Tx, ctx WorkflowTemplateCatalogContext, store *memoryWorkflowTemplateCatalogRepository, before int) error {
	values := store.audits[workflowTemplateScopeKey(ctx, "audits")]
	for index := before; index < len(values); index++ {
		value := values[index]
		payload, err := encodeWorkflowTemplateRecord(value)
		if err != nil {
			return err
		}
		occurredAt, _ := time.Parse(time.RFC3339Nano, value.CreatedAt)
		sequence, err := workflowTemplateAuditSequence(value)
		if err != nil || sequence != index+1 {
			return errWorkflowTemplateStoreUnavailable
		}
		if _, err = tx.Exec(workflowTemplateRequestContext(ctx), `INSERT INTO workflow_template_audits
(tenant_ref,workspace_id,audit_id,audit_sequence,resource_kind,resource_id,action,occurred_at,sanitized_audit_payload)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, ctx.TenantRef, ctx.WorkspaceID, value.AuditID, sequence, value.ResourceKind, value.ResourceID, value.Action, occurredAt, string(payload)); err != nil {
			return postgresWorkflowTemplateMutationError(err, errWorkflowTemplateStoreUnavailable)
		}
	}
	return nil
}

func postgresWorkflowTemplateMutationError(err error, conflict error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && (pgErr.Code == "23505" || pgErr.Code == "40001") && conflict != nil {
		return conflict
	}
	if errors.As(err, &pgErr) {
		return fmt.Errorf("%w (SQLSTATE %s)", errWorkflowTemplateStoreUnavailable, pgErr.Code)
	}
	return errWorkflowTemplateStoreUnavailable
}

var _ workflowTemplateCatalogRepository = (*postgresWorkflowTemplateCatalogRepository)(nil)
