package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresAdminProviderRouteRepository struct {
	pool *pgxpool.Pool
}

type postgresAdminProviderRouteQueryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func newPostgresAdminProviderRouteRepository(pool *pgxpool.Pool) *postgresAdminProviderRouteRepository {
	return &postgresAdminProviderRouteRepository{pool: pool}
}

func (repository *postgresAdminProviderRouteRepository) PutDraft(
	ctx AdminProviderRouteContext,
	expectedRevision int,
	draft AdminProviderRouteConfigurationDraft,
	now time.Time,
) (AdminProviderRouteConfigurationDraft, error) {
	var output AdminProviderRouteConfigurationDraft
	err := repository.mutate(ctx, func(transaction pgx.Tx) error {
		store := newMemoryAdminProviderRouteRepository()
		current, readErr := readPostgresAdminProviderRouteDraft(transaction, ctx, draft.ConfigurationID)
		if readErr == nil {
			store.drafts[adminProviderRouteConfigurationKey(ctx, draft.ConfigurationID)] = current
		} else if !errors.Is(readErr, errAdminProviderRouteDraftNotFound) {
			return readErr
		}
		updated, operationErr := store.PutDraft(ctx, expectedRevision, draft, now)
		if operationErr != nil {
			return operationErr
		}
		payload, encodeErr := encodeAdminProviderRouteStored(updated)
		if encodeErr != nil {
			return encodeErr
		}
		var command pgconn.CommandTag
		if expectedRevision == 0 {
			command, operationErr = transaction.Exec(adminProviderRouteDatabaseContext(ctx), `INSERT INTO admin_provider_route_drafts
                (tenant_ref,workspace_id,environment,configuration_id,draft_revision,draft_digest,sanitized_draft_payload,updated_at)
                VALUES ($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT DO NOTHING`,
				ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, updated.ConfigurationID, updated.DraftRevision,
				updated.DraftDigest, payload, updated.UpdatedAt)
		} else {
			command, operationErr = transaction.Exec(adminProviderRouteDatabaseContext(ctx), `UPDATE admin_provider_route_drafts SET
                draft_revision=$1,draft_digest=$2,sanitized_draft_payload=$3,updated_at=$4
                WHERE tenant_ref=$5 AND workspace_id=$6 AND environment=$7 AND configuration_id=$8 AND draft_revision=$9`,
				updated.DraftRevision, updated.DraftDigest, payload, updated.UpdatedAt,
				ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, updated.ConfigurationID, expectedRevision)
		}
		if operationErr != nil {
			return errAdminProviderRouteStoreUnavailable
		}
		if command.RowsAffected() != 1 {
			return adminProviderRouteDraftConflictError{CurrentRevision: expectedRevision}
		}
		output = updated
		return nil
	})
	return output, adminProviderRouteStorageError(err)
}

func (repository *postgresAdminProviderRouteRepository) ReadDraft(
	ctx AdminProviderRouteContext,
	configurationID string,
) (AdminProviderRouteConfigurationDraft, error) {
	if !repository.ready(ctx) {
		return AdminProviderRouteConfigurationDraft{}, errAdminProviderRouteStoreUnavailable
	}
	value, err := readPostgresAdminProviderRouteDraft(repository.pool, ctx, configurationID)
	return value, adminProviderRouteStorageError(err)
}

func (repository *postgresAdminProviderRouteRepository) CreateCandidate(
	ctx AdminProviderRouteContext,
	candidate AdminProviderRouteCandidate,
) (AdminProviderRouteCandidate, error) {
	if !repository.ready(ctx) || candidate.TenantRef != ctx.TenantRef || candidate.WorkspaceID != ctx.WorkspaceID ||
		candidate.Environment != ctx.Environment || !validAdminProviderRouteCandidateIntegrity(candidate) {
		return AdminProviderRouteCandidate{}, errAdminProviderRouteStoreUnavailable
	}
	payload, err := encodeAdminProviderRouteStored(candidate)
	if err != nil {
		return AdminProviderRouteCandidate{}, err
	}
	command, err := repository.pool.Exec(adminProviderRouteDatabaseContext(ctx), `INSERT INTO admin_provider_route_candidates
        (tenant_ref,workspace_id,environment,configuration_id,candidate_id,source_draft_revision,source_draft_digest,
         candidate_digest,candidate_state,review_version,sanitized_candidate_payload,created_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) ON CONFLICT DO NOTHING`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, candidate.ConfigurationID, candidate.CandidateID,
		candidate.SourceDraftRevision, candidate.SourceDraftDigest, candidate.CandidateDigest, candidate.CandidateState,
		candidate.ReviewVersion, payload, candidate.CreatedAt)
	if err != nil {
		return AdminProviderRouteCandidate{}, errAdminProviderRouteStoreUnavailable
	}
	if command.RowsAffected() != 1 {
		return AdminProviderRouteCandidate{}, errAdminProviderRouteCandidateConflict
	}
	return cloneAdminProviderRouteCandidate(candidate), nil
}

func (repository *postgresAdminProviderRouteRepository) ReadCandidate(
	ctx AdminProviderRouteContext,
	configurationID string,
	candidateID string,
) (AdminProviderRouteCandidate, error) {
	if !repository.ready(ctx) {
		return AdminProviderRouteCandidate{}, errAdminProviderRouteStoreUnavailable
	}
	value, err := readPostgresAdminProviderRouteCandidate(repository.pool, ctx, configurationID, candidateID)
	return value, adminProviderRouteStorageError(err)
}

func (repository *postgresAdminProviderRouteRepository) ReviewCandidate(
	ctx AdminProviderRouteContext,
	configurationID string,
	candidateID string,
	expectedReviewVersion int,
	review AdminProviderRouteReview,
) (AdminProviderRouteCandidate, error) {
	var output AdminProviderRouteCandidate
	err := repository.mutate(ctx, func(transaction pgx.Tx) error {
		current, readErr := readPostgresAdminProviderRouteCandidate(transaction, ctx, configurationID, candidateID)
		if readErr != nil {
			return readErr
		}
		store := newMemoryAdminProviderRouteRepository()
		store.candidates[adminProviderRouteCandidateKey(ctx, configurationID, candidateID)] = current
		updated, operationErr := store.ReviewCandidate(
			ctx, configurationID, candidateID, expectedReviewVersion, review,
		)
		if operationErr != nil {
			return operationErr
		}
		candidatePayload, encodeErr := encodeAdminProviderRouteStored(updated)
		if encodeErr != nil {
			return encodeErr
		}
		reviewPayload, encodeErr := encodeAdminProviderRouteStored(review)
		if encodeErr != nil {
			return encodeErr
		}
		command, operationErr := transaction.Exec(adminProviderRouteDatabaseContext(ctx), `UPDATE admin_provider_route_candidates SET
            candidate_state=$1,review_version=$2,sanitized_candidate_payload=$3
            WHERE tenant_ref=$4 AND workspace_id=$5 AND environment=$6 AND configuration_id=$7
              AND candidate_id=$8 AND review_version=$9`,
			updated.CandidateState, updated.ReviewVersion, candidatePayload,
			ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, configurationID, candidateID, expectedReviewVersion)
		if operationErr != nil {
			return errAdminProviderRouteStoreUnavailable
		}
		if command.RowsAffected() != 1 {
			return adminProviderRouteReviewConflictError{
				CurrentReviewVersion: current.ReviewVersion,
				CurrentState:         current.CandidateState,
			}
		}
		_, operationErr = transaction.Exec(adminProviderRouteDatabaseContext(ctx), `INSERT INTO admin_provider_route_reviews
            (tenant_ref,workspace_id,environment,configuration_id,candidate_id,review_version,decision,resulting_state,
             sanitized_review_payload,reviewed_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, configurationID, candidateID, review.ReviewVersion,
			review.Decision, review.ResultingState, reviewPayload, review.ReviewedAt)
		if operationErr != nil {
			return errAdminProviderRouteStoreUnavailable
		}
		output = updated
		return nil
	})
	return output, adminProviderRouteStorageError(err)
}

func (repository *postgresAdminProviderRouteRepository) CommitActivation(
	ctx AdminProviderRouteContext,
	configurationID string,
	candidateID string,
	expectedGeneration int,
	action string,
	reason string,
	snapshot AdminProviderRouteSnapshot,
	now time.Time,
) (AdminProviderRouteSnapshot, AdminProviderRouteActivationRecord, error) {
	var outputSnapshot AdminProviderRouteSnapshot
	var outputActivation AdminProviderRouteActivationRecord
	err := repository.mutate(ctx, func(transaction pgx.Tx) error {
		candidate, readErr := readPostgresAdminProviderRouteCandidate(transaction, ctx, configurationID, candidateID)
		if readErr != nil {
			return readErr
		}
		store := newMemoryAdminProviderRouteRepository()
		store.candidates[adminProviderRouteCandidateKey(ctx, configurationID, candidateID)] = candidate
		current, readErr := readPostgresAdminProviderRouteSnapshot(transaction, ctx, configurationID)
		if readErr == nil {
			store.snapshots[adminProviderRouteConfigurationKey(ctx, configurationID)] = current
		} else if !errors.Is(readErr, errAdminProviderRouteCandidateNotFound) {
			return readErr
		}
		history, readErr := listPostgresAdminProviderRouteActivations(transaction, ctx, configurationID)
		if readErr != nil {
			return readErr
		}
		store.activations[adminProviderRouteConfigurationKey(ctx, configurationID)] = history
		active, activation, operationErr := store.CommitActivation(
			ctx, configurationID, candidateID, expectedGeneration, action, reason, snapshot, now,
		)
		if operationErr != nil {
			return operationErr
		}
		snapshotPayload, encodeErr := encodeAdminProviderRouteStored(active)
		if encodeErr != nil {
			return encodeErr
		}
		activationPayload, encodeErr := encodeAdminProviderRouteStored(activation)
		if encodeErr != nil {
			return encodeErr
		}
		var command pgconn.CommandTag
		if expectedGeneration == 0 {
			command, operationErr = transaction.Exec(adminProviderRouteDatabaseContext(ctx), `INSERT INTO admin_provider_route_active_snapshots
                (tenant_ref,workspace_id,environment,configuration_id,generation,candidate_id,candidate_digest,
                 snapshot_digest,sanitized_snapshot_payload,activated_at)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT DO NOTHING`,
				ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, configurationID, active.Generation, active.CandidateID,
				active.CandidateDigest, active.SnapshotDigest, snapshotPayload, active.ActivatedAt)
		} else {
			command, operationErr = transaction.Exec(adminProviderRouteDatabaseContext(ctx), `UPDATE admin_provider_route_active_snapshots SET
                generation=$1,candidate_id=$2,candidate_digest=$3,snapshot_digest=$4,sanitized_snapshot_payload=$5,activated_at=$6
                WHERE tenant_ref=$7 AND workspace_id=$8 AND environment=$9 AND configuration_id=$10 AND generation=$11`,
				active.Generation, active.CandidateID, active.CandidateDigest, active.SnapshotDigest, snapshotPayload,
				active.ActivatedAt, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, configurationID, expectedGeneration)
		}
		if operationErr != nil {
			return errAdminProviderRouteStoreUnavailable
		}
		if command.RowsAffected() != 1 {
			return adminProviderRouteGenerationConflictError{CurrentGeneration: expectedGeneration}
		}
		_, operationErr = transaction.Exec(adminProviderRouteDatabaseContext(ctx), `INSERT INTO admin_provider_route_activation_records
            (tenant_ref,workspace_id,environment,configuration_id,after_generation,activation_id,action,after_candidate_id,
             after_snapshot_digest,previous_record_digest,record_digest,sanitized_activation_payload,created_at)
            VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, configurationID, activation.AfterGeneration,
			activation.ActivationID, activation.Action, activation.AfterCandidateID, activation.AfterSnapshotDigest,
			activation.PreviousRecordDigest, activation.RecordDigest, activationPayload, activation.CreatedAt)
		if operationErr != nil {
			return errAdminProviderRouteStoreUnavailable
		}
		outputSnapshot = active
		outputActivation = activation
		return nil
	})
	return outputSnapshot, outputActivation, adminProviderRouteStorageError(err)
}

func (repository *postgresAdminProviderRouteRepository) ReadActiveSnapshot(
	ctx AdminProviderRouteContext,
	configurationID string,
) (AdminProviderRouteSnapshot, error) {
	if !repository.ready(ctx) {
		return AdminProviderRouteSnapshot{}, errAdminProviderRouteStoreUnavailable
	}
	value, err := readPostgresAdminProviderRouteSnapshot(repository.pool, ctx, configurationID)
	return value, adminProviderRouteStorageError(err)
}

func (repository *postgresAdminProviderRouteRepository) ListActivations(
	ctx AdminProviderRouteContext,
	configurationID string,
) ([]AdminProviderRouteActivationRecord, error) {
	if !repository.ready(ctx) {
		return nil, errAdminProviderRouteStoreUnavailable
	}
	values, err := listPostgresAdminProviderRouteActivations(repository.pool, ctx, configurationID)
	return values, adminProviderRouteStorageError(err)
}

func (repository *postgresAdminProviderRouteRepository) ready(ctx AdminProviderRouteContext) bool {
	return repository != nil && repository.pool != nil && validAdminProviderRouteStorageContext(ctx)
}

func (repository *postgresAdminProviderRouteRepository) mutate(
	ctx AdminProviderRouteContext,
	operation func(pgx.Tx) error,
) error {
	if !repository.ready(ctx) {
		return errAdminProviderRouteStoreUnavailable
	}
	requestContext := adminProviderRouteDatabaseContext(ctx)
	transaction, err := repository.pool.BeginTx(requestContext, pgx.TxOptions{})
	if err != nil {
		return errAdminProviderRouteStoreUnavailable
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	if _, err = transaction.Exec(
		requestContext,
		"SELECT pg_advisory_xact_lock($1)",
		adminProviderRoutePostgresScopeLockKey(ctx),
	); err != nil {
		return errAdminProviderRouteStoreUnavailable
	}
	if err = operation(transaction); err != nil {
		return err
	}
	if err = transaction.Commit(requestContext); err != nil {
		return errAdminProviderRouteStoreUnavailable
	}
	return nil
}

func adminProviderRoutePostgresScopeLockKey(ctx AdminProviderRouteContext) int64 {
	digest := sha256.Sum256([]byte(adminProviderRouteScopePrefix(ctx)))
	return int64(binary.BigEndian.Uint64(digest[:8]))
}

func readPostgresAdminProviderRouteDraft(
	query postgresAdminProviderRouteQueryer,
	ctx AdminProviderRouteContext,
	configurationID string,
) (AdminProviderRouteConfigurationDraft, error) {
	var revision int
	var digest string
	var payload []byte
	err := query.QueryRow(adminProviderRouteDatabaseContext(ctx), `SELECT draft_revision,draft_digest,sanitized_draft_payload
        FROM admin_provider_route_drafts WHERE tenant_ref=$1 AND workspace_id=$2 AND environment=$3 AND configuration_id=$4`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, configurationID).Scan(&revision, &digest, &payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return AdminProviderRouteConfigurationDraft{}, errAdminProviderRouteDraftNotFound
	}
	if err != nil {
		return AdminProviderRouteConfigurationDraft{}, errAdminProviderRouteStoreUnavailable
	}
	return decodeAdminProviderRouteDraft(payload, ctx, configurationID, revision, digest)
}

func readPostgresAdminProviderRouteCandidate(
	query postgresAdminProviderRouteQueryer,
	ctx AdminProviderRouteContext,
	configurationID string,
	candidateID string,
) (AdminProviderRouteCandidate, error) {
	var sourceDraftRevision, reviewVersion int
	var sourceDraftDigest, candidateDigest, candidateState string
	var payload []byte
	err := query.QueryRow(adminProviderRouteDatabaseContext(ctx), `SELECT source_draft_revision,source_draft_digest,
        candidate_digest,candidate_state,review_version,sanitized_candidate_payload
        FROM admin_provider_route_candidates WHERE tenant_ref=$1 AND workspace_id=$2 AND environment=$3
          AND configuration_id=$4 AND candidate_id=$5`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, configurationID, candidateID).
		Scan(&sourceDraftRevision, &sourceDraftDigest, &candidateDigest, &candidateState, &reviewVersion, &payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return AdminProviderRouteCandidate{}, errAdminProviderRouteCandidateNotFound
	}
	if err != nil {
		return AdminProviderRouteCandidate{}, errAdminProviderRouteStoreUnavailable
	}
	candidate, err := decodeAdminProviderRouteCandidate(
		payload, ctx, configurationID, candidateID, sourceDraftRevision, sourceDraftDigest,
		candidateDigest, candidateState, reviewVersion,
	)
	if err != nil {
		return AdminProviderRouteCandidate{}, err
	}
	review, err := readPostgresAdminProviderRouteReview(query, ctx, configurationID, candidateID, reviewVersion)
	if err != nil {
		return AdminProviderRouteCandidate{}, err
	}
	if err := verifyAdminProviderRouteCandidateReview(candidate, review); err != nil {
		return AdminProviderRouteCandidate{}, err
	}
	return candidate, nil
}

func readPostgresAdminProviderRouteReview(
	query postgresAdminProviderRouteQueryer,
	ctx AdminProviderRouteContext,
	configurationID string,
	candidateID string,
	expectedReviewVersion int,
) (*AdminProviderRouteReview, error) {
	var count int
	if err := query.QueryRow(adminProviderRouteDatabaseContext(ctx), `SELECT COUNT(*)
        FROM admin_provider_route_reviews WHERE tenant_ref=$1 AND workspace_id=$2 AND environment=$3
          AND configuration_id=$4 AND candidate_id=$5`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, configurationID, candidateID).Scan(&count); err != nil {
		return nil, errAdminProviderRouteStoreUnavailable
	}
	if count != expectedReviewVersion {
		return nil, errAdminProviderRouteStoreUnavailable
	}
	if expectedReviewVersion == 0 {
		return nil, nil
	}
	var decision, resultingState string
	var payload []byte
	err := query.QueryRow(adminProviderRouteDatabaseContext(ctx), `SELECT decision,resulting_state,sanitized_review_payload
        FROM admin_provider_route_reviews WHERE tenant_ref=$1 AND workspace_id=$2 AND environment=$3
          AND configuration_id=$4 AND candidate_id=$5 AND review_version=$6`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, configurationID, candidateID, expectedReviewVersion).
		Scan(&decision, &resultingState, &payload)
	if err != nil {
		return nil, errAdminProviderRouteStoreUnavailable
	}
	value, err := decodeAdminProviderRouteReview(payload, expectedReviewVersion, decision, resultingState)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func readPostgresAdminProviderRouteSnapshot(
	query postgresAdminProviderRouteQueryer,
	ctx AdminProviderRouteContext,
	configurationID string,
) (AdminProviderRouteSnapshot, error) {
	var generation int
	var candidateID, candidateDigest, snapshotDigest string
	var payload []byte
	err := query.QueryRow(adminProviderRouteDatabaseContext(ctx), `SELECT generation,candidate_id,candidate_digest,
        snapshot_digest,sanitized_snapshot_payload FROM admin_provider_route_active_snapshots
        WHERE tenant_ref=$1 AND workspace_id=$2 AND environment=$3 AND configuration_id=$4`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, configurationID).
		Scan(&generation, &candidateID, &candidateDigest, &snapshotDigest, &payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return AdminProviderRouteSnapshot{}, errAdminProviderRouteCandidateNotFound
	}
	if err != nil {
		return AdminProviderRouteSnapshot{}, errAdminProviderRouteStoreUnavailable
	}
	return decodeAdminProviderRouteSnapshot(
		payload, ctx, configurationID, generation, candidateID, candidateDigest, snapshotDigest,
	)
}

func listPostgresAdminProviderRouteActivations(
	query postgresAdminProviderRouteQueryer,
	ctx AdminProviderRouteContext,
	configurationID string,
) ([]AdminProviderRouteActivationRecord, error) {
	rows, err := query.Query(adminProviderRouteDatabaseContext(ctx), `SELECT after_generation,activation_id,action,
        after_candidate_id,after_snapshot_digest,previous_record_digest,record_digest,sanitized_activation_payload
        FROM admin_provider_route_activation_records WHERE tenant_ref=$1 AND workspace_id=$2 AND environment=$3
          AND configuration_id=$4 ORDER BY after_generation`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, configurationID)
	if err != nil {
		return nil, errAdminProviderRouteStoreUnavailable
	}
	defer rows.Close()
	values := make([]AdminProviderRouteActivationRecord, 0)
	for rows.Next() {
		var generation int
		var activationID, action, candidateID, snapshotDigest, previousDigest, recordDigest string
		var payload []byte
		if rows.Scan(&generation, &activationID, &action, &candidateID, &snapshotDigest,
			&previousDigest, &recordDigest, &payload) != nil {
			return nil, errAdminProviderRouteStoreUnavailable
		}
		value, decodeErr := decodeAdminProviderRouteActivation(
			payload, configurationID, generation, activationID, action, candidateID,
			snapshotDigest, previousDigest, recordDigest,
		)
		if decodeErr != nil {
			return nil, decodeErr
		}
		values = append(values, value)
	}
	if rows.Err() != nil || !validAdminProviderRouteActivationHistory(configurationID, values) {
		return nil, errAdminProviderRouteStoreUnavailable
	}
	return values, nil
}

var _ adminProviderRouteRepository = (*postgresAdminProviderRouteRepository)(nil)
