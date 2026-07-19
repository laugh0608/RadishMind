package httpapi

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresWorkflowDefinitionReleaseRepository struct{ pool *pgxpool.Pool }

func newPostgresWorkflowDefinitionReleaseRepository(pool *pgxpool.Pool) *postgresWorkflowDefinitionReleaseRepository {
	return &postgresWorkflowDefinitionReleaseRepository{pool: pool}
}

type postgresWorkflowDefinitionQueryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func (repository *postgresWorkflowDefinitionReleaseRepository) CreateCandidate(ctx WorkflowDefinitionReleaseContext, candidateID, definitionID string, draft SavedWorkflowDraft, now time.Time) (WorkflowDefinitionReleaseCandidate, error) {
	var output WorkflowDefinitionReleaseCandidate
	err := repository.mutate(ctx, func(tx pgx.Tx, store *workflowDefinitionReleaseStore) error {
		candidate, err := store.CreateCandidate(ctx, candidateID, definitionID, draft, now)
		if err != nil {
			return err
		}
		payload, err := encodeWorkflowDefinitionRecord(candidate)
		if err != nil {
			return err
		}
		createdAt, _ := time.Parse(time.RFC3339Nano, candidate.CreatedAt)
		_, err = tx.Exec(workflowDefinitionRequestContext(ctx), `INSERT INTO workflow_definition_release_candidates
(tenant_ref,workspace_id,application_id,owner_subject_ref,candidate_id,definition_id,candidate_state,review_version,source_draft_id,source_draft_version,source_draft_digest,definition_digest,created_at,updated_at,sanitized_candidate_payload)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$13,$14)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, candidate.CandidateID, candidate.DefinitionID, candidate.State, candidate.ReviewVersion, candidate.SourceDraftID, candidate.SourceDraftVersion, candidate.SourceDraftDigest, candidate.DefinitionDigest, createdAt, payload)
		if err != nil {
			return postgresWorkflowDefinitionMutationError(err)
		}
		audit := store.audits[workflowDefinitionScopeKey(ctx, "candidate")][0]
		if err = insertPostgresWorkflowDefinitionAudit(tx, ctx, audit); err != nil {
			return err
		}
		output = candidate
		return nil
	})
	return output, err
}

func (repository *postgresWorkflowDefinitionReleaseRepository) Review(ctx WorkflowDefinitionReleaseContext, candidateID string, expected int, decision, reason, sourceDigest string, now time.Time) (WorkflowDefinitionReleaseCandidate, *WorkflowDefinitionVersion, error) {
	var output WorkflowDefinitionReleaseCandidate
	var outputVersion *WorkflowDefinitionVersion
	err := repository.mutate(ctx, func(tx pgx.Tx, store *workflowDefinitionReleaseStore) error {
		candidate, version, err := store.Review(ctx, candidateID, expected, decision, reason, sourceDigest, now)
		if err != nil {
			return err
		}
		payload, err := encodeWorkflowDefinitionRecord(candidate)
		if err != nil {
			return err
		}
		updatedAt, _ := time.Parse(time.RFC3339Nano, candidate.UpdatedAt)
		command, err := tx.Exec(workflowDefinitionRequestContext(ctx), `UPDATE workflow_definition_release_candidates SET candidate_state=$1,review_version=$2,updated_at=$3,sanitized_candidate_payload=$4 WHERE tenant_ref=$5 AND workspace_id=$6 AND application_id=$7 AND owner_subject_ref=$8 AND candidate_id=$9 AND review_version=$10`, candidate.State, candidate.ReviewVersion, updatedAt, payload, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, candidateID, expected)
		if err != nil {
			return postgresWorkflowDefinitionMutationError(err)
		}
		if command.RowsAffected() != 1 {
			return errWorkflowDefinitionConflict
		}
		review := candidate.Reviews[len(candidate.Reviews)-1]
		reviewPayload, err := encodeWorkflowDefinitionRecord(review)
		if err != nil {
			return err
		}
		reviewedAt, _ := time.Parse(time.RFC3339Nano, review.ReviewedAt)
		if _, err = tx.Exec(workflowDefinitionRequestContext(ctx), `INSERT INTO workflow_definition_release_decisions (tenant_ref,workspace_id,application_id,owner_subject_ref,candidate_id,review_version,decision,reviewed_at,sanitized_decision_payload) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, candidateID, review.ReviewVersion, review.Decision, reviewedAt, reviewPayload); err != nil {
			return postgresWorkflowDefinitionMutationError(err)
		}
		candidateAudits := store.audits[workflowDefinitionScopeKey(ctx, "candidate")]
		if err = insertPostgresWorkflowDefinitionAudit(tx, ctx, candidateAudits[len(candidateAudits)-1]); err != nil {
			return err
		}
		if version != nil {
			versionPayload, encodeErr := encodeWorkflowDefinitionRecord(*version)
			if encodeErr != nil {
				return encodeErr
			}
			createdAt, _ := time.Parse(time.RFC3339Nano, version.CreatedAt)
			if _, err = tx.Exec(workflowDefinitionRequestContext(ctx), `INSERT INTO workflow_definition_versions (tenant_ref,workspace_id,application_id,owner_subject_ref,definition_id,definition_version,definition_digest,candidate_id,candidate_review_version,created_at,sanitized_version_payload) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, version.DefinitionID, version.Version, version.DefinitionDigest, version.CandidateID, version.CandidateReviewVersion, createdAt, versionPayload); err != nil {
				return postgresWorkflowDefinitionMutationError(err)
			}
			versionAudits := store.audits[workflowDefinitionScopeKey(ctx, "version")]
			if err = insertPostgresWorkflowDefinitionAudit(tx, ctx, versionAudits[len(versionAudits)-1]); err != nil {
				return err
			}
			copyVersion := *version
			outputVersion = &copyVersion
		}
		output = candidate
		return nil
	})
	return output, outputVersion, err
}

func (repository *postgresWorkflowDefinitionReleaseRepository) DecideActivation(ctx WorkflowDefinitionReleaseContext, definitionID string, expected int, decision string, version int, reason string, now time.Time) (WorkflowDefinitionActivation, error) {
	var output WorkflowDefinitionActivation
	err := repository.mutate(ctx, func(tx pgx.Tx, store *workflowDefinitionReleaseStore) error {
		activation, err := store.DecideActivation(ctx, definitionID, expected, decision, version, reason, now)
		if err != nil {
			return err
		}
		payload, err := encodeWorkflowDefinitionRecord(activation)
		if err != nil {
			return err
		}
		updatedAt, _ := time.Parse(time.RFC3339Nano, activation.UpdatedAt)
		var command pgconn.CommandTag
		if expected == 0 {
			command, err = tx.Exec(workflowDefinitionRequestContext(ctx), `INSERT INTO workflow_definition_activations (tenant_ref,workspace_id,application_id,owner_subject_ref,definition_id,pointer_version,activation_state,active_version,active_definition_digest,updated_at,sanitized_activation_payload) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT DO NOTHING`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, definitionID, activation.PointerVersion, activation.State, activation.ActiveVersion, activation.ActiveDefinitionDigest, updatedAt, payload)
		} else {
			command, err = tx.Exec(workflowDefinitionRequestContext(ctx), `UPDATE workflow_definition_activations SET pointer_version=$1,activation_state=$2,active_version=$3,active_definition_digest=$4,updated_at=$5,sanitized_activation_payload=$6 WHERE tenant_ref=$7 AND workspace_id=$8 AND application_id=$9 AND owner_subject_ref=$10 AND definition_id=$11 AND pointer_version=$12`, activation.PointerVersion, activation.State, activation.ActiveVersion, activation.ActiveDefinitionDigest, updatedAt, payload, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, definitionID, expected)
		}
		if err != nil {
			return postgresWorkflowDefinitionMutationError(err)
		}
		if command.RowsAffected() != 1 {
			return errWorkflowDefinitionConflict
		}
		event := activation.Events[len(activation.Events)-1]
		eventPayload, _ := encodeWorkflowDefinitionRecord(event)
		eventAt, _ := time.Parse(time.RFC3339Nano, event.CreatedAt)
		if _, err = tx.Exec(workflowDefinitionRequestContext(ctx), `INSERT INTO workflow_definition_activation_events (tenant_ref,workspace_id,application_id,owner_subject_ref,definition_id,event_id,after_pointer_version,occurred_at,sanitized_event_payload) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, definitionID, event.EventID, event.AfterPointerVersion, eventAt, eventPayload); err != nil {
			return postgresWorkflowDefinitionMutationError(err)
		}
		audits := store.audits[workflowDefinitionScopeKey(ctx, "activation")]
		if err = insertPostgresWorkflowDefinitionAudit(tx, ctx, audits[len(audits)-1]); err != nil {
			return err
		}
		output = activation
		return nil
	})
	return output, err
}

func (repository *postgresWorkflowDefinitionReleaseRepository) ReadCandidate(ctx WorkflowDefinitionReleaseContext, candidateID string) (WorkflowDefinitionReleaseCandidate, error) {
	store, done, err := repository.readStore(ctx)
	if err != nil {
		return WorkflowDefinitionReleaseCandidate{}, err
	}
	defer done()
	return store.ReadCandidate(ctx, candidateID)
}
func (repository *postgresWorkflowDefinitionReleaseRepository) ListCandidates(ctx WorkflowDefinitionReleaseContext) ([]WorkflowDefinitionReleaseCandidate, error) {
	store, done, err := repository.readStore(ctx)
	if err != nil {
		return nil, err
	}
	defer done()
	return store.ListCandidates(ctx)
}
func (repository *postgresWorkflowDefinitionReleaseRepository) ListVersions(ctx WorkflowDefinitionReleaseContext, definitionID string) ([]WorkflowDefinitionVersion, error) {
	store, done, err := repository.readStore(ctx)
	if err != nil {
		return nil, err
	}
	defer done()
	return store.ListVersions(ctx, definitionID)
}
func (repository *postgresWorkflowDefinitionReleaseRepository) ReadVersion(ctx WorkflowDefinitionReleaseContext, definitionID string, version int) (WorkflowDefinitionVersion, error) {
	store, done, err := repository.readStore(ctx)
	if err != nil {
		return WorkflowDefinitionVersion{}, err
	}
	defer done()
	return store.ReadVersion(ctx, definitionID, version)
}
func (repository *postgresWorkflowDefinitionReleaseRepository) ReadActivation(ctx WorkflowDefinitionReleaseContext, definitionID string) (WorkflowDefinitionActivation, error) {
	store, done, err := repository.readStore(ctx)
	if err != nil {
		return WorkflowDefinitionActivation{}, err
	}
	defer done()
	return store.ReadActivation(ctx, definitionID)
}

func (repository *postgresWorkflowDefinitionReleaseRepository) ListSummaries(ctx ReadRepositoryContext, request ListWorkflowDefinitionSummariesRequest) ListWorkflowDefinitionSummariesResult {
	if repository == nil || repository.pool == nil || strings.TrimSpace(ctx.TenantRef) == "" || strings.TrimSpace(ctx.SubjectRef) == "" {
		return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureStoreUnavailable)
	}
	requestContext := ctx.RequestContext
	if requestContext == nil {
		requestContext = context.Background()
	}
	tx, err := repository.pool.BeginTx(requestContext, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureStoreUnavailable)
	}
	defer tx.Rollback(context.Background())
	versions := map[string][]WorkflowDefinitionVersion{}
	meta := map[string]workflowDefinitionSummarySource{}
	rows, err := tx.Query(requestContext, `SELECT workspace_id,application_id,owner_subject_ref,definition_id,definition_version,definition_digest,candidate_id,candidate_review_version,created_at,sanitized_version_payload FROM workflow_definition_versions WHERE tenant_ref=$1 AND owner_subject_ref=$2 ORDER BY workspace_id,application_id,definition_id,definition_version`, ctx.TenantRef, ctx.SubjectRef)
	if err != nil {
		return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureStoreUnavailable)
	}
	for rows.Next() {
		var workspaceID, applicationID, owner, definitionID, digest, candidateID string
		var version, reviewVersion int
		var createdAt time.Time
		var payload []byte
		if rows.Scan(&workspaceID, &applicationID, &owner, &definitionID, &version, &digest, &candidateID, &reviewVersion, &createdAt, &payload) != nil {
			rows.Close()
			return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureStoreUnavailable)
		}
		var value WorkflowDefinitionVersion
		if decodeWorkflowDefinitionRecord(payload, &value) != nil || validateStoredWorkflowDefinitionVersion(value) != nil {
			rows.Close()
			return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureContractMismatch)
		}
		decodedAt, _ := time.Parse(time.RFC3339Nano, value.CreatedAt)
		if value.DefinitionID != definitionID || value.Version != version || value.DefinitionDigest != digest || value.CandidateID != candidateID || value.CandidateReviewVersion != reviewVersion || !postgresWorkflowDefinitionTimeMatches(createdAt, decodedAt) {
			rows.Close()
			return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureContractMismatch)
		}
		key := strings.Join([]string{ctx.TenantRef, workspaceID, applicationID, owner, definitionID}, "\x00")
		if value.Version != len(versions[key])+1 {
			rows.Close()
			return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureContractMismatch)
		}
		versions[key] = append(versions[key], value)
		meta[key] = workflowDefinitionSummarySource{TenantRef: ctx.TenantRef, WorkspaceID: workspaceID, ApplicationID: applicationID, OwnerSubjectRef: owner}
	}
	if rows.Err() != nil {
		rows.Close()
		return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureStoreUnavailable)
	}
	rows.Close()
	activations := map[string]WorkflowDefinitionActivation{}
	rows, err = tx.Query(requestContext, `SELECT workspace_id,application_id,owner_subject_ref,definition_id,pointer_version,activation_state,active_version,active_definition_digest,updated_at,sanitized_activation_payload FROM workflow_definition_activations WHERE tenant_ref=$1 AND owner_subject_ref=$2 ORDER BY workspace_id,application_id,definition_id`, ctx.TenantRef, ctx.SubjectRef)
	if err != nil {
		return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureStoreUnavailable)
	}
	for rows.Next() {
		var workspaceID, applicationID, owner, definitionID, state, digest string
		var pointer, active int
		var updatedAt time.Time
		var payload []byte
		if rows.Scan(&workspaceID, &applicationID, &owner, &definitionID, &pointer, &state, &active, &digest, &updatedAt, &payload) != nil {
			rows.Close()
			return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureStoreUnavailable)
		}
		var value WorkflowDefinitionActivation
		if decodeWorkflowDefinitionRecord(payload, &value) != nil || validateStoredWorkflowDefinitionActivation(value) != nil {
			rows.Close()
			return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureContractMismatch)
		}
		decodedAt, _ := time.Parse(time.RFC3339Nano, value.UpdatedAt)
		if value.DefinitionID != definitionID || value.PointerVersion != pointer || value.State != state || value.ActiveVersion != active || value.ActiveDefinitionDigest != digest || !postgresWorkflowDefinitionTimeMatches(updatedAt, decodedAt) {
			rows.Close()
			return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureContractMismatch)
		}
		key := strings.Join([]string{ctx.TenantRef, workspaceID, applicationID, owner, definitionID}, "\x00")
		activations[key] = value
	}
	if rows.Err() != nil {
		rows.Close()
		return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureStoreUnavailable)
	}
	rows.Close()
	sources := make([]workflowDefinitionSummarySource, 0, len(versions))
	for key, list := range versions {
		source := meta[key]
		selected := list[len(list)-1]
		if activation, ok := activations[key]; ok {
			copied := activation
			source.Activation = &copied
			if activation.State == workflowDefinitionActivationActive {
				if activation.ActiveVersion < 1 || activation.ActiveVersion > len(list) {
					return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureContractMismatch)
				}
				selected = list[activation.ActiveVersion-1]
			}
		}
		source.Version = selected
		sources = append(sources, source)
	}
	return projectWorkflowDefinitionSummaries(ctx, request, sources)
}

func (repository *postgresWorkflowDefinitionReleaseRepository) mutate(ctx WorkflowDefinitionReleaseContext, operation func(pgx.Tx, *workflowDefinitionReleaseStore) error) error {
	if repository == nil || repository.pool == nil || !validWorkflowDefinitionContext(ctx) {
		return errWorkflowDefinitionStore
	}
	requestContext := workflowDefinitionRequestContext(ctx)
	tx, err := repository.pool.BeginTx(requestContext, pgx.TxOptions{})
	if err != nil {
		return errWorkflowDefinitionStore
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	lockScope := sha256.Sum256([]byte(workflowDefinitionScopeKey(ctx, "definition-release")))
	if _, err = tx.Exec(requestContext, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, fmt.Sprintf("%x", lockScope)); err != nil {
		return postgresWorkflowDefinitionMutationError(err)
	}
	store, err := loadPostgresWorkflowDefinitionStore(ctx, tx)
	if err != nil {
		return err
	}
	if err = operation(tx, store); err != nil {
		return err
	}
	if err = tx.Commit(requestContext); err != nil {
		return errWorkflowDefinitionStore
	}
	return nil
}

func (repository *postgresWorkflowDefinitionReleaseRepository) readStore(ctx WorkflowDefinitionReleaseContext) (*workflowDefinitionReleaseStore, func(), error) {
	if repository == nil || repository.pool == nil || !validWorkflowDefinitionContext(ctx) {
		return nil, func() {}, errWorkflowDefinitionStore
	}
	tx, err := repository.pool.BeginTx(workflowDefinitionRequestContext(ctx), pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, func() {}, errWorkflowDefinitionStore
	}
	store, err := loadPostgresWorkflowDefinitionStore(ctx, tx)
	if err != nil {
		_ = tx.Rollback(context.Background())
		return nil, func() {}, err
	}
	return store, func() { _ = tx.Rollback(context.Background()) }, nil
}

func loadPostgresWorkflowDefinitionStore(ctx WorkflowDefinitionReleaseContext, query postgresWorkflowDefinitionQueryer) (*workflowDefinitionReleaseStore, error) {
	store := newWorkflowDefinitionReleaseStore()
	requestContext := workflowDefinitionRequestContext(ctx)
	scope := []any{ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef}
	rows, err := query.Query(requestContext, `SELECT candidate_id,definition_id,candidate_state,review_version,source_draft_id,source_draft_version,source_draft_digest,definition_digest,created_at,updated_at,sanitized_candidate_payload FROM workflow_definition_release_candidates WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 ORDER BY created_at,candidate_id`, scope...)
	if err != nil {
		return nil, postgresWorkflowDefinitionMutationError(err)
	}
	for rows.Next() {
		var candidateID, definitionID, state, draftID, draftDigest, definitionDigest string
		var reviewVersion, draftVersion int
		var createdAt, updatedAt time.Time
		var payload []byte
		if rows.Scan(&candidateID, &definitionID, &state, &reviewVersion, &draftID, &draftVersion, &draftDigest, &definitionDigest, &createdAt, &updatedAt, &payload) != nil {
			rows.Close()
			return nil, errWorkflowDefinitionStore
		}
		value, loadErr := loadWorkflowDefinitionCandidateIntoStore(store, ctx, payload)
		decodedCreated, _ := time.Parse(time.RFC3339Nano, value.CreatedAt)
		decodedUpdated, _ := time.Parse(time.RFC3339Nano, value.UpdatedAt)
		if loadErr != nil || value.CandidateID != candidateID || value.DefinitionID != definitionID || value.State != state || value.ReviewVersion != reviewVersion || value.SourceDraftID != draftID || value.SourceDraftVersion != draftVersion || value.SourceDraftDigest != draftDigest || value.DefinitionDigest != definitionDigest || !postgresWorkflowDefinitionTimeMatches(createdAt, decodedCreated) || !postgresWorkflowDefinitionTimeMatches(updatedAt, decodedUpdated) {
			rows.Close()
			return nil, errWorkflowDefinitionStore
		}
	}
	if rows.Err() != nil {
		rows.Close()
		return nil, errWorkflowDefinitionStore
	}
	rows.Close()
	rows, err = query.Query(requestContext, `SELECT definition_id,definition_version,definition_digest,candidate_id,candidate_review_version,created_at,sanitized_version_payload FROM workflow_definition_versions WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 ORDER BY definition_id,definition_version`, scope...)
	if err != nil {
		return nil, postgresWorkflowDefinitionMutationError(err)
	}
	for rows.Next() {
		var definitionID, digest, candidateID string
		var version, reviewVersion int
		var createdAt time.Time
		var payload []byte
		if rows.Scan(&definitionID, &version, &digest, &candidateID, &reviewVersion, &createdAt, &payload) != nil {
			rows.Close()
			return nil, errWorkflowDefinitionStore
		}
		value, loadErr := loadWorkflowDefinitionVersionIntoStore(store, ctx, payload)
		decodedAt, _ := time.Parse(time.RFC3339Nano, value.CreatedAt)
		if loadErr != nil || value.DefinitionID != definitionID || value.Version != version || value.DefinitionDigest != digest || value.CandidateID != candidateID || value.CandidateReviewVersion != reviewVersion || !postgresWorkflowDefinitionTimeMatches(createdAt, decodedAt) {
			rows.Close()
			return nil, errWorkflowDefinitionStore
		}
	}
	if rows.Err() != nil {
		rows.Close()
		return nil, errWorkflowDefinitionStore
	}
	rows.Close()
	rows, err = query.Query(requestContext, `SELECT definition_id,pointer_version,activation_state,active_version,active_definition_digest,updated_at,sanitized_activation_payload FROM workflow_definition_activations WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 ORDER BY definition_id`, scope...)
	if err != nil {
		return nil, postgresWorkflowDefinitionMutationError(err)
	}
	for rows.Next() {
		var definitionID, state, digest string
		var pointer, active int
		var updatedAt time.Time
		var payload []byte
		if rows.Scan(&definitionID, &pointer, &state, &active, &digest, &updatedAt, &payload) != nil {
			rows.Close()
			return nil, errWorkflowDefinitionStore
		}
		value, loadErr := loadWorkflowDefinitionActivationIntoStore(store, ctx, payload)
		decodedAt, _ := time.Parse(time.RFC3339Nano, value.UpdatedAt)
		if loadErr != nil || value.DefinitionID != definitionID || value.PointerVersion != pointer || value.State != state || value.ActiveVersion != active || value.ActiveDefinitionDigest != digest || !postgresWorkflowDefinitionTimeMatches(updatedAt, decodedAt) {
			rows.Close()
			return nil, errWorkflowDefinitionStore
		}
	}
	if rows.Err() != nil {
		rows.Close()
		return nil, errWorkflowDefinitionStore
	}
	rows.Close()
	rows, err = query.Query(requestContext, `SELECT audit_id,resource_kind,resource_id,action,occurred_at,sanitized_audit_payload FROM workflow_definition_release_audits WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 ORDER BY occurred_at,audit_id`, scope...)
	if err != nil {
		return nil, postgresWorkflowDefinitionMutationError(err)
	}
	for rows.Next() {
		var auditID, kind, resourceID, action string
		var occurredAt time.Time
		var payload []byte
		if rows.Scan(&auditID, &kind, &resourceID, &action, &occurredAt, &payload) != nil {
			rows.Close()
			return nil, errWorkflowDefinitionStore
		}
		value, loadErr := loadWorkflowDefinitionAuditIntoStore(store, ctx, payload)
		decodedAt, _ := time.Parse(time.RFC3339Nano, value.CreatedAt)
		if loadErr != nil || value.AuditID != auditID || value.ResourceKind != kind || value.ResourceID != resourceID || value.Action != action || !postgresWorkflowDefinitionTimeMatches(occurredAt, decodedAt) {
			rows.Close()
			return nil, errWorkflowDefinitionStore
		}
	}
	if rows.Err() != nil {
		rows.Close()
		return nil, errWorkflowDefinitionStore
	}
	rows.Close()
	if err = validatePostgresWorkflowDefinitionEvidence(ctx, query, store); err != nil {
		return nil, err
	}
	return store, nil
}

func validatePostgresWorkflowDefinitionEvidence(ctx WorkflowDefinitionReleaseContext, query postgresWorkflowDefinitionQueryer, store *workflowDefinitionReleaseStore) error {
	requestContext := workflowDefinitionRequestContext(ctx)
	scope := []any{ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef}
	decisionCounts := map[string]int{}
	rows, err := query.Query(requestContext, `SELECT candidate_id,review_version,sanitized_decision_payload FROM workflow_definition_release_decisions WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 ORDER BY candidate_id,review_version`, scope...)
	if err != nil {
		return postgresWorkflowDefinitionMutationError(err)
	}
	for rows.Next() {
		var candidateID string
		var version int
		var payload []byte
		if rows.Scan(&candidateID, &version, &payload) != nil {
			rows.Close()
			return errWorkflowDefinitionStore
		}
		var review WorkflowDefinitionReview
		if decodeWorkflowDefinitionRecord(payload, &review) != nil {
			rows.Close()
			return errWorkflowDefinitionStore
		}
		candidate, ok := store.candidates[workflowDefinitionScopeKey(ctx, candidateID)]
		if !ok || version < 1 || version > len(candidate.Reviews) || !reflect.DeepEqual(review, candidate.Reviews[version-1]) {
			rows.Close()
			return errWorkflowDefinitionStore
		}
		decisionCounts[candidateID]++
	}
	if rows.Err() != nil {
		rows.Close()
		return errWorkflowDefinitionStore
	}
	rows.Close()
	for _, candidate := range store.candidates {
		if decisionCounts[candidate.CandidateID] != len(candidate.Reviews) {
			return errWorkflowDefinitionStore
		}
	}
	eventCounts := map[string]int{}
	rows, err = query.Query(requestContext, `SELECT definition_id,after_pointer_version,sanitized_event_payload FROM workflow_definition_activation_events WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 ORDER BY definition_id,after_pointer_version`, scope...)
	if err != nil {
		return postgresWorkflowDefinitionMutationError(err)
	}
	for rows.Next() {
		var definitionID string
		var version int
		var payload []byte
		if rows.Scan(&definitionID, &version, &payload) != nil {
			rows.Close()
			return errWorkflowDefinitionStore
		}
		var event WorkflowDefinitionActivationEvent
		if decodeWorkflowDefinitionRecord(payload, &event) != nil {
			rows.Close()
			return errWorkflowDefinitionStore
		}
		activation, ok := store.activations[workflowDefinitionScopeKey(ctx, definitionID)]
		if !ok || version < 1 || version > len(activation.Events) || !reflect.DeepEqual(event, activation.Events[version-1]) {
			rows.Close()
			return errWorkflowDefinitionStore
		}
		eventCounts[definitionID]++
	}
	if rows.Err() != nil {
		rows.Close()
		return errWorkflowDefinitionStore
	}
	rows.Close()
	for _, activation := range store.activations {
		if eventCounts[activation.DefinitionID] != len(activation.Events) {
			return errWorkflowDefinitionStore
		}
	}
	return nil
}

func insertPostgresWorkflowDefinitionAudit(tx pgx.Tx, ctx WorkflowDefinitionReleaseContext, audit WorkflowDefinitionReleaseAudit) error {
	payload, err := encodeWorkflowDefinitionRecord(audit)
	if err != nil {
		return err
	}
	occurredAt, _ := time.Parse(time.RFC3339Nano, audit.CreatedAt)
	if _, err = tx.Exec(workflowDefinitionRequestContext(ctx), `INSERT INTO workflow_definition_release_audits (tenant_ref,workspace_id,application_id,owner_subject_ref,audit_id,resource_kind,resource_id,action,occurred_at,sanitized_audit_payload) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, audit.AuditID, audit.ResourceKind, audit.ResourceID, audit.Action, occurredAt, payload); err != nil {
		return postgresWorkflowDefinitionMutationError(err)
	}
	return nil
}
func postgresWorkflowDefinitionMutationError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && (pgErr.Code == "23505" || pgErr.Code == "40001") {
		return errWorkflowDefinitionConflict
	}
	if errors.As(err, &pgErr) {
		return fmt.Errorf("%w (SQLSTATE %s)", errWorkflowDefinitionStore, pgErr.Code)
	}
	return errWorkflowDefinitionStore
}
func postgresWorkflowDefinitionTimeMatches(stored, decoded time.Time) bool {
	return stored.UTC().UnixMicro() == decoded.UTC().UnixMicro()
}

var _ workflowDefinitionReleaseRepository = (*postgresWorkflowDefinitionReleaseRepository)(nil)
