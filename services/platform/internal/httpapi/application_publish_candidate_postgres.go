package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresApplicationPublishCandidateRepository struct {
	pool *pgxpool.Pool
}

type applicationPublishCandidateRow interface {
	Scan(...any) error
}

func newPostgresApplicationPublishCandidateRepository(pool *pgxpool.Pool) *postgresApplicationPublishCandidateRepository {
	return &postgresApplicationPublishCandidateRepository{pool: pool}
}

func (repository *postgresApplicationPublishCandidateRepository) Create(requestContext ApplicationPublishContext, candidate ApplicationPublishCandidate) (ApplicationPublishCandidate, error) {
	if repository == nil || repository.pool == nil {
		return ApplicationPublishCandidate{}, errApplicationPublishStoreUnavailable
	}
	payload, err := json.Marshal(candidate)
	if err != nil {
		return ApplicationPublishCandidate{}, errApplicationPublishStoreUnavailable
	}
	created, err := scanApplicationPublishCandidate(repository.pool.QueryRow(applicationPublishDatabaseContext(requestContext), `INSERT INTO application_publish_candidates
        (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id, schema_version,
         draft_id, draft_version, draft_digest, candidate_state, review_version, sanitized_candidate_payload,
         created_at, updated_at, created_by_actor_ref, updated_by_actor_ref, request_id, audit_ref)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
        ON CONFLICT DO NOTHING RETURNING sanitized_candidate_payload`,
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ApplicationID, requestContext.OwnerSubjectRef,
		candidate.CandidateID, candidate.SchemaVersion, candidate.DraftID, candidate.DraftVersion, candidate.DraftDigest,
		candidate.CandidateState, candidate.ReviewVersion, payload, candidate.CreatedAt, candidate.UpdatedAt,
		candidate.CreatedByActorRef, candidate.UpdatedByActorRef, candidate.RequestID, candidate.AuditRef,
	))
	if err == nil {
		return created, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ApplicationPublishCandidate{}, errApplicationPublishImmutableConflict
	}
	return ApplicationPublishCandidate{}, errApplicationPublishStoreUnavailable
}

func (repository *postgresApplicationPublishCandidateRepository) Read(requestContext ApplicationPublishContext, candidateID string) (ApplicationPublishCandidate, error) {
	if repository == nil || repository.pool == nil {
		return ApplicationPublishCandidate{}, errApplicationPublishStoreUnavailable
	}
	candidate, err := scanApplicationPublishCandidate(repository.pool.QueryRow(applicationPublishDatabaseContext(requestContext), `SELECT sanitized_candidate_payload
        FROM application_publish_candidates WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3
          AND owner_subject_ref=$4 AND candidate_id=$5`, requestContext.TenantRef, requestContext.WorkspaceID,
		requestContext.ApplicationID, requestContext.OwnerSubjectRef, candidateID))
	if errors.Is(err, pgx.ErrNoRows) {
		return ApplicationPublishCandidate{}, errApplicationPublishNotFound
	}
	if err != nil {
		return ApplicationPublishCandidate{}, errApplicationPublishStoreUnavailable
	}
	return candidate, nil
}

func (repository *postgresApplicationPublishCandidateRepository) List(requestContext ApplicationPublishContext) ([]ApplicationPublishCandidate, error) {
	if repository == nil || repository.pool == nil {
		return nil, errApplicationPublishStoreUnavailable
	}
	rows, err := repository.pool.Query(applicationPublishDatabaseContext(requestContext), `SELECT sanitized_candidate_payload
        FROM application_publish_candidates WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3
          AND owner_subject_ref=$4 ORDER BY created_at DESC, candidate_id DESC LIMIT 200`, requestContext.TenantRef,
		requestContext.WorkspaceID, requestContext.ApplicationID, requestContext.OwnerSubjectRef)
	if err != nil {
		return nil, errApplicationPublishStoreUnavailable
	}
	defer rows.Close()
	candidates := make([]ApplicationPublishCandidate, 0)
	for rows.Next() {
		candidate, scanErr := scanApplicationPublishCandidate(rows)
		if scanErr != nil {
			return nil, errApplicationPublishStoreUnavailable
		}
		candidates = append(candidates, candidate)
	}
	if rows.Err() != nil {
		return nil, errApplicationPublishStoreUnavailable
	}
	return candidates, nil
}

func (repository *postgresApplicationPublishCandidateRepository) AppendReview(requestContext ApplicationPublishContext, candidateID string, expectedVersion int, review ApplicationPublishReviewRecord) (ApplicationPublishCandidate, error) {
	current, err := repository.Read(requestContext, candidateID)
	if err != nil {
		return ApplicationPublishCandidate{}, err
	}
	if current.ReviewVersion != expectedVersion {
		return ApplicationPublishCandidate{}, applicationPublishReviewConflictError{CurrentVersion: current.ReviewVersion, CurrentState: current.CandidateState}
	}
	if !applicationPublishTransitionAllowed(current.CandidateState, review.Decision) {
		return ApplicationPublishCandidate{}, errApplicationPublishReviewTransition
	}
	current.ReviewVersion = review.ReviewVersion
	current.CandidateState = review.State
	current.Reviews = append(current.Reviews, review)
	current.UpdatedAt = review.ReviewedAt
	current.UpdatedByActorRef = review.ReviewerRef
	current.RequestID = review.RequestID
	current.AuditRef = review.AuditRef
	payload, err := json.Marshal(current)
	if err != nil {
		return ApplicationPublishCandidate{}, errApplicationPublishStoreUnavailable
	}
	updated, err := scanApplicationPublishCandidate(repository.pool.QueryRow(applicationPublishDatabaseContext(requestContext), `UPDATE application_publish_candidates SET
        candidate_state=$1, review_version=$2, sanitized_candidate_payload=$3, updated_at=$4,
        updated_by_actor_ref=$5, request_id=$6, audit_ref=$7
        WHERE tenant_ref=$8 AND workspace_id=$9 AND application_id=$10 AND owner_subject_ref=$11
          AND candidate_id=$12 AND review_version=$13 RETURNING sanitized_candidate_payload`, current.CandidateState,
		current.ReviewVersion, payload, current.UpdatedAt, current.UpdatedByActorRef, current.RequestID, current.AuditRef,
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ApplicationID, requestContext.OwnerSubjectRef,
		candidateID, expectedVersion))
	if err == nil {
		return updated, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return ApplicationPublishCandidate{}, errApplicationPublishStoreUnavailable
	}
	latest, readErr := repository.Read(requestContext, candidateID)
	if readErr != nil {
		return ApplicationPublishCandidate{}, readErr
	}
	return ApplicationPublishCandidate{}, applicationPublishReviewConflictError{CurrentVersion: latest.ReviewVersion, CurrentState: latest.CandidateState}
}

func scanApplicationPublishCandidate(row applicationPublishCandidateRow) (ApplicationPublishCandidate, error) {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		return ApplicationPublishCandidate{}, err
	}
	var candidate ApplicationPublishCandidate
	if err := json.Unmarshal(payload, &candidate); err != nil || strings.TrimSpace(candidate.CandidateID) == "" ||
		!applicationPublishCandidateSchemaSupported(candidate.SchemaVersion) || candidate.DraftVersion < 1 || candidate.ReviewVersion < 0 ||
		(candidate.SchemaVersion == applicationPublishCandidateSchemaVersionV1 && (candidate.Configuration.WorkflowRAGBindingRef != nil || candidate.Configuration.PromptTemplateRef != nil)) ||
		(candidate.SchemaVersion == applicationPublishCandidateSchemaVersionV2 && (candidate.Configuration.WorkflowRAGBindingRef == nil || candidate.Configuration.PromptTemplateRef != nil)) ||
		(candidate.SchemaVersion == applicationPublishCandidateSchemaVersionV3 && (candidate.Configuration.WorkflowRAGBindingRef != nil || candidate.Configuration.PromptTemplateRef == nil || candidate.Configuration.ApplicationKind != "prompt_application")) ||
		(candidate.Configuration.WorkflowRAGBindingRef != nil && !validWorkflowRAGApplicationBindingRef(*candidate.Configuration.WorkflowRAGBindingRef)) ||
		(candidate.Configuration.PromptTemplateRef != nil && !validPromptApplicationTemplateRef(*candidate.Configuration.PromptTemplateRef)) {
		return ApplicationPublishCandidate{}, errors.New("stored application publish candidate contract mismatch")
	}
	digest, err := applicationConfigurationCanonicalDigest(candidate.Configuration)
	if err != nil || digest != candidate.DraftDigest {
		return ApplicationPublishCandidate{}, errors.New("stored application publish candidate contract mismatch")
	}
	if candidate.EvidenceRequestIDs == nil {
		candidate.EvidenceRequestIDs = []string{}
	}
	if candidate.Reviews == nil {
		candidate.Reviews = []ApplicationPublishReviewRecord{}
	}
	return candidate, nil
}

func applicationPublishCandidateSchemaSupported(schemaVersion string) bool {
	return schemaVersion == applicationPublishCandidateSchemaVersionV1 || schemaVersion == applicationPublishCandidateSchemaVersionV2 || schemaVersion == applicationPublishCandidateSchemaVersionV3
}

func applicationPublishDatabaseContext(requestContext ApplicationPublishContext) context.Context {
	if requestContext.RequestContext != nil {
		return requestContext.RequestContext
	}
	return context.Background()
}
