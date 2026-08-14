package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type sqliteAdminProviderRouteRepository struct {
	database *sql.DB
}

type sqliteAdminProviderRouteQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func newSQLiteAdminProviderRouteRepository(database *sql.DB) *sqliteAdminProviderRouteRepository {
	return &sqliteAdminProviderRouteRepository{database: database}
}

func (repository *sqliteAdminProviderRouteRepository) PutDraft(
	ctx AdminProviderRouteContext,
	expectedRevision int,
	draft AdminProviderRouteConfigurationDraft,
	now time.Time,
) (AdminProviderRouteConfigurationDraft, error) {
	var output AdminProviderRouteConfigurationDraft
	err := repository.mutate(ctx, func(connection *sql.Conn) error {
		store := newMemoryAdminProviderRouteRepository()
		current, readErr := readSQLiteAdminProviderRouteDraft(connection, ctx, draft.ConfigurationID)
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
		var result sql.Result
		if expectedRevision == 0 {
			result, operationErr = connection.ExecContext(adminProviderRouteDatabaseContext(ctx), `INSERT INTO admin_provider_route_drafts
                (tenant_ref,workspace_id,environment,configuration_id,draft_revision,draft_digest,sanitized_draft_payload,updated_at)
                VALUES (?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING`,
				ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, updated.ConfigurationID, updated.DraftRevision,
				updated.DraftDigest, string(payload), updated.UpdatedAt)
		} else {
			result, operationErr = connection.ExecContext(adminProviderRouteDatabaseContext(ctx), `UPDATE admin_provider_route_drafts SET
                draft_revision=?,draft_digest=?,sanitized_draft_payload=?,updated_at=?
                WHERE tenant_ref=? AND workspace_id=? AND environment=? AND configuration_id=? AND draft_revision=?`,
				updated.DraftRevision, updated.DraftDigest, string(payload), updated.UpdatedAt,
				ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, updated.ConfigurationID, expectedRevision)
		}
		if operationErr != nil {
			return errAdminProviderRouteStoreUnavailable
		}
		if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
			return adminProviderRouteDraftConflictError{CurrentRevision: expectedRevision}
		}
		output = updated
		return nil
	})
	return output, adminProviderRouteStorageError(err)
}

func (repository *sqliteAdminProviderRouteRepository) ReadDraft(
	ctx AdminProviderRouteContext,
	configurationID string,
) (AdminProviderRouteConfigurationDraft, error) {
	if !repository.ready(ctx) {
		return AdminProviderRouteConfigurationDraft{}, errAdminProviderRouteStoreUnavailable
	}
	value, err := readSQLiteAdminProviderRouteDraft(repository.database, ctx, configurationID)
	return value, adminProviderRouteStorageError(err)
}

func (repository *sqliteAdminProviderRouteRepository) CreateCandidate(
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
	result, err := repository.database.ExecContext(adminProviderRouteDatabaseContext(ctx), `INSERT INTO admin_provider_route_candidates
        (tenant_ref,workspace_id,environment,configuration_id,candidate_id,source_draft_revision,source_draft_digest,
         candidate_digest,candidate_state,review_version,sanitized_candidate_payload,created_at)
        VALUES (?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, candidate.ConfigurationID, candidate.CandidateID,
		candidate.SourceDraftRevision, candidate.SourceDraftDigest, candidate.CandidateDigest, candidate.CandidateState,
		candidate.ReviewVersion, string(payload), candidate.CreatedAt)
	if err != nil {
		return AdminProviderRouteCandidate{}, errAdminProviderRouteStoreUnavailable
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
		return AdminProviderRouteCandidate{}, errAdminProviderRouteCandidateConflict
	}
	return cloneAdminProviderRouteCandidate(candidate), nil
}

func (repository *sqliteAdminProviderRouteRepository) ReadCandidate(
	ctx AdminProviderRouteContext,
	configurationID string,
	candidateID string,
) (AdminProviderRouteCandidate, error) {
	if !repository.ready(ctx) {
		return AdminProviderRouteCandidate{}, errAdminProviderRouteStoreUnavailable
	}
	value, err := readSQLiteAdminProviderRouteCandidate(repository.database, ctx, configurationID, candidateID)
	return value, adminProviderRouteStorageError(err)
}

func (repository *sqliteAdminProviderRouteRepository) ReviewCandidate(
	ctx AdminProviderRouteContext,
	configurationID string,
	candidateID string,
	expectedReviewVersion int,
	review AdminProviderRouteReview,
) (AdminProviderRouteCandidate, error) {
	var output AdminProviderRouteCandidate
	err := repository.mutate(ctx, func(connection *sql.Conn) error {
		current, readErr := readSQLiteAdminProviderRouteCandidate(connection, ctx, configurationID, candidateID)
		if readErr != nil {
			return readErr
		}
		store := newMemoryAdminProviderRouteRepository()
		key := adminProviderRouteCandidateKey(ctx, configurationID, candidateID)
		store.candidates[key] = current
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
		result, operationErr := connection.ExecContext(adminProviderRouteDatabaseContext(ctx), `UPDATE admin_provider_route_candidates SET
            candidate_state=?,review_version=?,sanitized_candidate_payload=?
            WHERE tenant_ref=? AND workspace_id=? AND environment=? AND configuration_id=? AND candidate_id=? AND review_version=?`,
			updated.CandidateState, updated.ReviewVersion, string(candidatePayload),
			ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, configurationID, candidateID, expectedReviewVersion)
		if operationErr != nil {
			return errAdminProviderRouteStoreUnavailable
		}
		if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
			return adminProviderRouteReviewConflictError{
				CurrentReviewVersion: current.ReviewVersion,
				CurrentState:         current.CandidateState,
			}
		}
		_, operationErr = connection.ExecContext(adminProviderRouteDatabaseContext(ctx), `INSERT INTO admin_provider_route_reviews
            (tenant_ref,workspace_id,environment,configuration_id,candidate_id,review_version,decision,resulting_state,
             sanitized_review_payload,reviewed_at) VALUES (?,?,?,?,?,?,?,?,?,?)`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, configurationID, candidateID, review.ReviewVersion,
			review.Decision, review.ResultingState, string(reviewPayload), review.ReviewedAt)
		if operationErr != nil {
			return errAdminProviderRouteStoreUnavailable
		}
		output = updated
		return nil
	})
	return output, adminProviderRouteStorageError(err)
}

func (repository *sqliteAdminProviderRouteRepository) CommitActivation(
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
	err := repository.mutate(ctx, func(connection *sql.Conn) error {
		candidate, readErr := readSQLiteAdminProviderRouteCandidate(connection, ctx, configurationID, candidateID)
		if readErr != nil {
			return readErr
		}
		store := newMemoryAdminProviderRouteRepository()
		store.candidates[adminProviderRouteCandidateKey(ctx, configurationID, candidateID)] = candidate
		current, readErr := readSQLiteAdminProviderRouteSnapshot(connection, ctx, configurationID)
		if readErr == nil {
			store.snapshots[adminProviderRouteConfigurationKey(ctx, configurationID)] = current
		} else if !errors.Is(readErr, errAdminProviderRouteCandidateNotFound) {
			return readErr
		}
		history, readErr := listSQLiteAdminProviderRouteActivations(connection, ctx, configurationID)
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
		var result sql.Result
		if expectedGeneration == 0 {
			result, operationErr = connection.ExecContext(adminProviderRouteDatabaseContext(ctx), `INSERT INTO admin_provider_route_active_snapshots
                (tenant_ref,workspace_id,environment,configuration_id,generation,candidate_id,candidate_digest,
                 snapshot_digest,sanitized_snapshot_payload,activated_at)
                VALUES (?,?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING`,
				ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, configurationID, active.Generation, active.CandidateID,
				active.CandidateDigest, active.SnapshotDigest, string(snapshotPayload), active.ActivatedAt)
		} else {
			result, operationErr = connection.ExecContext(adminProviderRouteDatabaseContext(ctx), `UPDATE admin_provider_route_active_snapshots SET
                generation=?,candidate_id=?,candidate_digest=?,snapshot_digest=?,sanitized_snapshot_payload=?,activated_at=?
                WHERE tenant_ref=? AND workspace_id=? AND environment=? AND configuration_id=? AND generation=?`,
				active.Generation, active.CandidateID, active.CandidateDigest, active.SnapshotDigest, string(snapshotPayload),
				active.ActivatedAt, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, configurationID, expectedGeneration)
		}
		if operationErr != nil {
			return errAdminProviderRouteStoreUnavailable
		}
		if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
			return adminProviderRouteGenerationConflictError{CurrentGeneration: expectedGeneration}
		}
		_, operationErr = connection.ExecContext(adminProviderRouteDatabaseContext(ctx), `INSERT INTO admin_provider_route_activation_records
            (tenant_ref,workspace_id,environment,configuration_id,after_generation,activation_id,action,after_candidate_id,
             after_snapshot_digest,previous_record_digest,record_digest,sanitized_activation_payload,created_at)
            VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, configurationID, activation.AfterGeneration,
			activation.ActivationID, activation.Action, activation.AfterCandidateID, activation.AfterSnapshotDigest,
			activation.PreviousRecordDigest, activation.RecordDigest, string(activationPayload), activation.CreatedAt)
		if operationErr != nil {
			return errAdminProviderRouteStoreUnavailable
		}
		outputSnapshot = active
		outputActivation = activation
		return nil
	})
	return outputSnapshot, outputActivation, adminProviderRouteStorageError(err)
}

func (repository *sqliteAdminProviderRouteRepository) ReadActiveSnapshot(
	ctx AdminProviderRouteContext,
	configurationID string,
) (AdminProviderRouteSnapshot, error) {
	if !repository.ready(ctx) {
		return AdminProviderRouteSnapshot{}, errAdminProviderRouteStoreUnavailable
	}
	value, err := readSQLiteAdminProviderRouteSnapshot(repository.database, ctx, configurationID)
	return value, adminProviderRouteStorageError(err)
}

func (repository *sqliteAdminProviderRouteRepository) ListActivations(
	ctx AdminProviderRouteContext,
	configurationID string,
) ([]AdminProviderRouteActivationRecord, error) {
	if !repository.ready(ctx) {
		return nil, errAdminProviderRouteStoreUnavailable
	}
	values, err := listSQLiteAdminProviderRouteActivations(repository.database, ctx, configurationID)
	return values, adminProviderRouteStorageError(err)
}

func (repository *sqliteAdminProviderRouteRepository) ready(ctx AdminProviderRouteContext) bool {
	return repository != nil && repository.database != nil && validAdminProviderRouteStorageContext(ctx)
}

func (repository *sqliteAdminProviderRouteRepository) mutate(
	ctx AdminProviderRouteContext,
	operation func(*sql.Conn) error,
) error {
	if !repository.ready(ctx) {
		return errAdminProviderRouteStoreUnavailable
	}
	requestContext := adminProviderRouteDatabaseContext(ctx)
	connection, err := repository.database.Conn(requestContext)
	if err != nil {
		return errAdminProviderRouteStoreUnavailable
	}
	defer connection.Close()
	if _, err = connection.ExecContext(requestContext, "BEGIN IMMEDIATE"); err != nil {
		return errAdminProviderRouteStoreUnavailable
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = connection.ExecContext(context.Background(), "ROLLBACK")
		}
	}()
	if err = operation(connection); err != nil {
		return err
	}
	if _, err = connection.ExecContext(requestContext, "COMMIT"); err != nil {
		return errAdminProviderRouteStoreUnavailable
	}
	committed = true
	return nil
}

func readSQLiteAdminProviderRouteDraft(
	query sqliteAdminProviderRouteQueryer,
	ctx AdminProviderRouteContext,
	configurationID string,
) (AdminProviderRouteConfigurationDraft, error) {
	var revision int
	var digest string
	var payload []byte
	err := query.QueryRowContext(adminProviderRouteDatabaseContext(ctx), `SELECT draft_revision,draft_digest,sanitized_draft_payload
        FROM admin_provider_route_drafts WHERE tenant_ref=? AND workspace_id=? AND environment=? AND configuration_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, configurationID).Scan(&revision, &digest, &payload)
	if errors.Is(err, sql.ErrNoRows) {
		return AdminProviderRouteConfigurationDraft{}, errAdminProviderRouteDraftNotFound
	}
	if err != nil {
		return AdminProviderRouteConfigurationDraft{}, errAdminProviderRouteStoreUnavailable
	}
	return decodeAdminProviderRouteDraft(payload, ctx, configurationID, revision, digest)
}

func readSQLiteAdminProviderRouteCandidate(
	query sqliteAdminProviderRouteQueryer,
	ctx AdminProviderRouteContext,
	configurationID string,
	candidateID string,
) (AdminProviderRouteCandidate, error) {
	var sourceDraftRevision, reviewVersion int
	var sourceDraftDigest, candidateDigest, candidateState string
	var payload []byte
	err := query.QueryRowContext(adminProviderRouteDatabaseContext(ctx), `SELECT source_draft_revision,source_draft_digest,
        candidate_digest,candidate_state,review_version,sanitized_candidate_payload
        FROM admin_provider_route_candidates WHERE tenant_ref=? AND workspace_id=? AND environment=?
          AND configuration_id=? AND candidate_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, configurationID, candidateID).
		Scan(&sourceDraftRevision, &sourceDraftDigest, &candidateDigest, &candidateState, &reviewVersion, &payload)
	if errors.Is(err, sql.ErrNoRows) {
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
	review, err := readSQLiteAdminProviderRouteReview(query, ctx, configurationID, candidateID, reviewVersion)
	if err != nil {
		return AdminProviderRouteCandidate{}, err
	}
	if err := verifyAdminProviderRouteCandidateReview(candidate, review); err != nil {
		return AdminProviderRouteCandidate{}, err
	}
	return candidate, nil
}

func readSQLiteAdminProviderRouteReview(
	query sqliteAdminProviderRouteQueryer,
	ctx AdminProviderRouteContext,
	configurationID string,
	candidateID string,
	expectedReviewVersion int,
) (*AdminProviderRouteReview, error) {
	var count int
	if err := query.QueryRowContext(adminProviderRouteDatabaseContext(ctx), `SELECT COUNT(*)
        FROM admin_provider_route_reviews WHERE tenant_ref=? AND workspace_id=? AND environment=?
          AND configuration_id=? AND candidate_id=?`,
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
	err := query.QueryRowContext(adminProviderRouteDatabaseContext(ctx), `SELECT decision,resulting_state,sanitized_review_payload
        FROM admin_provider_route_reviews WHERE tenant_ref=? AND workspace_id=? AND environment=?
          AND configuration_id=? AND candidate_id=? AND review_version=?`,
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

func readSQLiteAdminProviderRouteSnapshot(
	query sqliteAdminProviderRouteQueryer,
	ctx AdminProviderRouteContext,
	configurationID string,
) (AdminProviderRouteSnapshot, error) {
	var generation int
	var candidateID, candidateDigest, snapshotDigest string
	var payload []byte
	err := query.QueryRowContext(adminProviderRouteDatabaseContext(ctx), `SELECT generation,candidate_id,candidate_digest,
        snapshot_digest,sanitized_snapshot_payload FROM admin_provider_route_active_snapshots
        WHERE tenant_ref=? AND workspace_id=? AND environment=? AND configuration_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, configurationID).
		Scan(&generation, &candidateID, &candidateDigest, &snapshotDigest, &payload)
	if errors.Is(err, sql.ErrNoRows) {
		return AdminProviderRouteSnapshot{}, errAdminProviderRouteCandidateNotFound
	}
	if err != nil {
		return AdminProviderRouteSnapshot{}, errAdminProviderRouteStoreUnavailable
	}
	return decodeAdminProviderRouteSnapshot(
		payload, ctx, configurationID, generation, candidateID, candidateDigest, snapshotDigest,
	)
}

func listSQLiteAdminProviderRouteActivations(
	query sqliteAdminProviderRouteQueryer,
	ctx AdminProviderRouteContext,
	configurationID string,
) ([]AdminProviderRouteActivationRecord, error) {
	rows, err := query.QueryContext(adminProviderRouteDatabaseContext(ctx), `SELECT after_generation,activation_id,action,
        after_candidate_id,after_snapshot_digest,previous_record_digest,record_digest,sanitized_activation_payload
        FROM admin_provider_route_activation_records WHERE tenant_ref=? AND workspace_id=? AND environment=?
          AND configuration_id=? ORDER BY after_generation`,
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

var _ adminProviderRouteRepository = (*sqliteAdminProviderRouteRepository)(nil)
