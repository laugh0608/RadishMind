package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
)

type sqliteApplicationPublishCandidateRepository struct {
	database *sql.DB
}

func newSQLiteApplicationPublishCandidateRepository(database *sql.DB) *sqliteApplicationPublishCandidateRepository {
	return &sqliteApplicationPublishCandidateRepository{database: database}
}

func (repository *sqliteApplicationPublishCandidateRepository) Create(requestContext ApplicationPublishContext, candidate ApplicationPublishCandidate) (ApplicationPublishCandidate, error) {
	if repository == nil || repository.database == nil {
		return ApplicationPublishCandidate{}, errApplicationPublishStoreUnavailable
	}
	payload, err := json.Marshal(candidate)
	if err != nil {
		return ApplicationPublishCandidate{}, errApplicationPublishStoreUnavailable
	}
	created, err := scanApplicationPublishCandidate(repository.database.QueryRowContext(applicationPublishDatabaseContext(requestContext), `INSERT INTO application_publish_candidates
        (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id, schema_version,
         draft_id, draft_version, draft_digest, candidate_state, review_version, sanitized_candidate_payload,
         created_at, updated_at, created_by_actor_ref, updated_by_actor_ref, request_id, audit_ref)
        VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
        ON CONFLICT DO NOTHING RETURNING sanitized_candidate_payload`,
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ApplicationID, requestContext.OwnerSubjectRef,
		candidate.CandidateID, candidate.SchemaVersion, candidate.DraftID, candidate.DraftVersion, candidate.DraftDigest,
		candidate.CandidateState, candidate.ReviewVersion, string(payload), candidate.CreatedAt, candidate.UpdatedAt,
		candidate.CreatedByActorRef, candidate.UpdatedByActorRef, candidate.RequestID, candidate.AuditRef,
	))
	if err == nil {
		return created, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ApplicationPublishCandidate{}, errApplicationPublishImmutableConflict
	}
	return ApplicationPublishCandidate{}, errApplicationPublishStoreUnavailable
}

func (repository *sqliteApplicationPublishCandidateRepository) Read(requestContext ApplicationPublishContext, candidateID string) (ApplicationPublishCandidate, error) {
	if repository == nil || repository.database == nil {
		return ApplicationPublishCandidate{}, errApplicationPublishStoreUnavailable
	}
	candidate, err := scanApplicationPublishCandidate(repository.database.QueryRowContext(applicationPublishDatabaseContext(requestContext), `SELECT sanitized_candidate_payload
        FROM application_publish_candidates WHERE tenant_ref=? AND workspace_id=? AND application_id=?
          AND owner_subject_ref=? AND candidate_id=?`, requestContext.TenantRef, requestContext.WorkspaceID,
		requestContext.ApplicationID, requestContext.OwnerSubjectRef, candidateID))
	if errors.Is(err, sql.ErrNoRows) {
		return ApplicationPublishCandidate{}, errApplicationPublishNotFound
	}
	if err != nil {
		return ApplicationPublishCandidate{}, errApplicationPublishStoreUnavailable
	}
	return candidate, nil
}

func (repository *sqliteApplicationPublishCandidateRepository) List(requestContext ApplicationPublishContext) ([]ApplicationPublishCandidate, error) {
	if repository == nil || repository.database == nil {
		return nil, errApplicationPublishStoreUnavailable
	}
	rows, err := repository.database.QueryContext(applicationPublishDatabaseContext(requestContext), `SELECT sanitized_candidate_payload
        FROM application_publish_candidates WHERE tenant_ref=? AND workspace_id=? AND application_id=?
          AND owner_subject_ref=? ORDER BY created_at DESC, candidate_id DESC LIMIT 200`, requestContext.TenantRef,
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

func (repository *sqliteApplicationPublishCandidateRepository) AppendReview(requestContext ApplicationPublishContext, candidateID string, expectedVersion int, review ApplicationPublishReviewRecord) (ApplicationPublishCandidate, error) {
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
	updated, err := scanApplicationPublishCandidate(repository.database.QueryRowContext(applicationPublishDatabaseContext(requestContext), `UPDATE application_publish_candidates SET
        candidate_state=?, review_version=?, sanitized_candidate_payload=?, updated_at=?,
        updated_by_actor_ref=?, request_id=?, audit_ref=?
        WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=?
          AND candidate_id=? AND review_version=? RETURNING sanitized_candidate_payload`, current.CandidateState,
		current.ReviewVersion, string(payload), current.UpdatedAt, current.UpdatedByActorRef, current.RequestID, current.AuditRef,
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ApplicationID, requestContext.OwnerSubjectRef,
		candidateID, expectedVersion))
	if err == nil {
		return updated, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return ApplicationPublishCandidate{}, errApplicationPublishStoreUnavailable
	}
	latest, readErr := repository.Read(requestContext, candidateID)
	if readErr != nil {
		return ApplicationPublishCandidate{}, readErr
	}
	return ApplicationPublishCandidate{}, applicationPublishReviewConflictError{CurrentVersion: latest.ReviewVersion, CurrentState: latest.CandidateState}
}
