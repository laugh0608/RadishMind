package httpapi

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type postgresLocalIdentitySelfServiceQuery interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func (repository *postgresLocalIdentityRepository) ListSelfServiceSessions(
	ctx context.Context,
	userID string,
	currentSessionID string,
	query LocalIdentitySelfServiceSessionListQuery,
) (LocalIdentitySelfServiceSessionPage, error) {
	if repository == nil || repository.pool == nil {
		return LocalIdentitySelfServiceSessionPage{}, errLocalIdentityStoreUnavailable
	}
	filter, cursor, err := normalizeLocalIdentitySelfServiceSessionQuery(userID, query)
	if err != nil {
		return LocalIdentitySelfServiceSessionPage{}, err
	}
	userID = strings.TrimSpace(userID)
	currentSessionID = strings.TrimSpace(currentSessionID)
	if !localSessionIDPattern.MatchString(currentSessionID) {
		return LocalIdentitySelfServiceSessionPage{}, errLocalIdentitySessionScopeDenied
	}
	transaction, err := repository.pool.BeginTx(identityContext(ctx), pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return LocalIdentitySelfServiceSessionPage{}, errLocalIdentityStoreUnavailable
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	if err = validatePostgresLocalIdentitySelfServiceActor(
		identityContext(ctx), transaction, userID, currentSessionID, filter.requestedAt,
	); err != nil {
		return LocalIdentitySelfServiceSessionPage{}, err
	}

	arguments := make([]any, 0, 8)
	bind := func(value any) string {
		arguments = append(arguments, value)
		return "$" + strconv.Itoa(len(arguments))
	}
	statement := postgresSessionSelect + ` WHERE user_id=` + bind(userID) + ` AND created_at<=` + bind(filter.snapshotAt)
	switch filter.State {
	case localIdentityStateActive:
		snapshotBinding := bind(filter.snapshotAt)
		statement += ` AND (revoked_at IS NULL OR revoked_at>` + snapshotBinding + `) AND expires_at>` + snapshotBinding
	case localIdentitySessionEffectiveStateExpired:
		snapshotBinding := bind(filter.snapshotAt)
		statement += ` AND (revoked_at IS NULL OR revoked_at>` + snapshotBinding + `) AND expires_at<=` + snapshotBinding
	case localIdentityStateRevoked:
		statement += ` AND revoked_at IS NOT NULL AND revoked_at<=` + bind(filter.snapshotAt)
	}
	if cursor.SessionID != "" {
		anchor, parseErr := localIdentitySelfServiceCursorCreatedAt(cursor)
		if parseErr != nil {
			return LocalIdentitySelfServiceSessionPage{}, errLocalIdentitySessionCursorInvalid
		}
		anchorBinding := bind(anchor)
		statement += ` AND (created_at<` + anchorBinding + ` OR (created_at=` + anchorBinding +
			` AND session_id<` + bind(cursor.SessionID) + `))`
	}
	statement += ` ORDER BY created_at DESC, session_id DESC LIMIT ` + bind(filter.Limit+1)
	rows, err := transaction.Query(identityContext(ctx), statement, arguments...)
	if err != nil {
		return LocalIdentitySelfServiceSessionPage{}, errLocalIdentityStoreUnavailable
	}
	sessions := make([]WebSession, 0, filter.Limit+1)
	for rows.Next() {
		session, scanErr := scanPostgresWebSession(rows)
		if scanErr != nil {
			rows.Close()
			return LocalIdentitySelfServiceSessionPage{}, scanErr
		}
		sessions = append(sessions, session)
	}
	if rows.Err() != nil {
		rows.Close()
		return LocalIdentitySelfServiceSessionPage{}, errLocalIdentityStoreUnavailable
	}
	rows.Close()
	if err = transaction.Commit(identityContext(ctx)); err != nil {
		return LocalIdentitySelfServiceSessionPage{}, errLocalIdentityStoreUnavailable
	}
	return buildLocalIdentitySelfServiceSessionPage(userID, currentSessionID, filter, sessions)
}

func (repository *postgresLocalIdentityRepository) RevokeOwnedWebSession(
	ctx context.Context,
	mutation localIdentityOwnedSessionRevocation,
) (LocalIdentitySelfServiceSessionRevocation, error) {
	if repository == nil || repository.pool == nil {
		return LocalIdentitySelfServiceSessionRevocation{}, errLocalIdentityStoreUnavailable
	}
	var result LocalIdentitySelfServiceSessionRevocation
	err := repository.mutateLocalIdentitySelfServiceScope(ctx, mutation.UserID, func(
		snapshot *memoryLocalIdentityRepository,
		transaction pgx.Tx,
	) error {
		var mutationErr error
		result, mutationErr = snapshot.RevokeOwnedWebSession(ctx, mutation)
		if mutationErr != nil {
			return mutationErr
		}
		revoked := snapshot.sessions[strings.TrimSpace(mutation.TargetSessionID)]
		command, execErr := transaction.Exec(identityContext(ctx), `UPDATE local_web_sessions SET
            lifecycle_state='revoked', record_version=record_version+1, updated_at=$1, revoked_at=$1, audit_ref=$2
            WHERE session_id=$3 AND user_id=$4 AND record_version=$5 AND lifecycle_state='active'
            AND expires_at>$1`, revoked.UpdatedAt, revoked.AuditRef, revoked.SessionID, revoked.UserID,
			mutation.ExpectedVersion)
		if execErr != nil {
			return errLocalIdentityStoreUnavailable
		}
		if command.RowsAffected() != 1 {
			return errLocalIdentitySessionVersionConflict
		}
		return nil
	})
	if err != nil {
		return LocalIdentitySelfServiceSessionRevocation{}, err
	}
	return result, nil
}

func (repository *postgresLocalIdentityRepository) RevokeOtherWebSessions(
	ctx context.Context,
	mutation localIdentityOtherSessionRevocation,
) (LocalIdentitySelfServiceBulkSessionRevocation, error) {
	if repository == nil || repository.pool == nil {
		return LocalIdentitySelfServiceBulkSessionRevocation{}, errLocalIdentityStoreUnavailable
	}
	var result LocalIdentitySelfServiceBulkSessionRevocation
	err := repository.mutateLocalIdentitySelfServiceScope(ctx, mutation.UserID, func(
		snapshot *memoryLocalIdentityRepository,
		transaction pgx.Tx,
	) error {
		var mutationErr error
		result, mutationErr = snapshot.RevokeOtherWebSessions(ctx, mutation)
		if mutationErr != nil {
			return mutationErr
		}
		command, execErr := transaction.Exec(identityContext(ctx), `UPDATE local_web_sessions SET
            lifecycle_state='revoked', record_version=record_version+1, updated_at=$1, revoked_at=$1, audit_ref=$2
            WHERE user_id=$3 AND session_id<>$4 AND lifecycle_state='active' AND expires_at>$1`,
			mutation.RevokedAt.UTC(), strings.TrimSpace(mutation.AuditRef), strings.TrimSpace(mutation.UserID),
			strings.TrimSpace(mutation.CurrentSessionID))
		if execErr != nil {
			return errLocalIdentityStoreUnavailable
		}
		if command.RowsAffected() != int64(result.RevokedCount) {
			return errLocalIdentitySessionBulkRevokeConflict
		}
		return nil
	})
	if err != nil {
		return LocalIdentitySelfServiceBulkSessionRevocation{}, err
	}
	return result, nil
}

func (repository *postgresLocalIdentityRepository) RotateLocalCredentialAndRevokeSessions(
	ctx context.Context,
	mutation localIdentityCredentialRotation,
) (LocalIdentitySelfServiceCredentialRotation, error) {
	if repository == nil || repository.pool == nil {
		return LocalIdentitySelfServiceCredentialRotation{}, errLocalIdentityStoreUnavailable
	}
	var result LocalIdentitySelfServiceCredentialRotation
	err := repository.mutateLocalIdentitySelfServiceScope(ctx, mutation.UserID, func(
		snapshot *memoryLocalIdentityRepository,
		transaction pgx.Tx,
	) error {
		userID := strings.TrimSpace(mutation.UserID)
		currentCredentialID := snapshot.activeCredentialByUser[userID]
		currentCredential := snapshot.credentials[currentCredentialID]
		var mutationErr error
		result, mutationErr = snapshot.RotateLocalCredentialAndRevokeSessions(ctx, mutation)
		if mutationErr != nil {
			return mutationErr
		}
		command, execErr := transaction.Exec(identityContext(ctx), `UPDATE local_credentials SET
            lifecycle_state='superseded', record_version=record_version+1, updated_at=$1, audit_ref=$2
            WHERE credential_id=$3 AND user_id=$4 AND record_version=$5 AND lifecycle_state='active'
            AND created_at<$1`, mutation.RotatedAt.UTC(), strings.TrimSpace(mutation.AuditRef),
			currentCredentialID, userID, currentCredential.RecordVersion)
		if execErr != nil {
			return errLocalIdentityStoreUnavailable
		}
		if command.RowsAffected() != 1 {
			return errLocalIdentityCredentialRotationConflict
		}
		replacement := mutation.Replacement
		_, execErr = transaction.Exec(identityContext(ctx), `INSERT INTO local_credentials
            (credential_id, user_id, schema_version, algorithm, policy_version, iterations, key_length, salt,
             derived_key, lifecycle_state, record_version, created_at, updated_at, audit_ref)
            VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`, replacement.CredentialID,
			replacement.UserID, replacement.SchemaVersion, replacement.Algorithm, replacement.PolicyVersion,
			replacement.Iterations, replacement.KeyLength, replacement.salt, replacement.derivedKey,
			replacement.LifecycleState, replacement.RecordVersion, replacement.CreatedAt, replacement.UpdatedAt,
			replacement.AuditRef)
		if execErr != nil {
			return postgresIdentityConflictOrUnavailable(execErr, errLocalIdentityCredentialRotationConflict)
		}
		command, execErr = transaction.Exec(identityContext(ctx), `UPDATE local_web_sessions SET
            lifecycle_state='revoked', record_version=record_version+1, updated_at=$1, revoked_at=$1, audit_ref=$2
            WHERE user_id=$3 AND lifecycle_state='active' AND authentication_method='local_password'
            AND authentication_source_ref=$4`, mutation.RotatedAt.UTC(), strings.TrimSpace(mutation.AuditRef),
			userID, "credential:"+currentCredentialID)
		if execErr != nil {
			return errLocalIdentityStoreUnavailable
		}
		if command.RowsAffected() != int64(result.RevokedSessionCount) {
			return errLocalIdentityCredentialRotationConflict
		}
		return nil
	})
	if err != nil {
		return LocalIdentitySelfServiceCredentialRotation{}, err
	}
	return result, nil
}

func (repository *postgresLocalIdentityRepository) mutateLocalIdentitySelfServiceScope(
	ctx context.Context,
	userID string,
	operation func(*memoryLocalIdentityRepository, pgx.Tx) error,
) error {
	transaction, err := repository.pool.Begin(identityContext(ctx))
	if err != nil {
		return errLocalIdentityStoreUnavailable
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	snapshot, err := loadPostgresLocalIdentitySelfServiceSnapshot(identityContext(ctx), transaction, userID)
	if err != nil {
		return err
	}
	if err = operation(snapshot, transaction); err != nil {
		return err
	}
	if err = transaction.Commit(identityContext(ctx)); err != nil {
		return errLocalIdentityStoreUnavailable
	}
	return nil
}

func loadPostgresLocalIdentitySelfServiceSnapshot(
	ctx context.Context,
	query postgresLocalIdentitySelfServiceQuery,
	userID string,
) (*memoryLocalIdentityRepository, error) {
	userID = strings.TrimSpace(userID)
	if !localUserIDPattern.MatchString(userID) {
		return nil, errLocalIdentityContractMismatch
	}
	snapshot := newMemoryLocalIdentityRepository()
	account, err := scanPostgresUserAccount(query.QueryRow(
		ctx, postgresAccountSelect+` WHERE user_id=$1 FOR UPDATE`, userID,
	))
	if errors.Is(err, errLocalIdentityNotFound) {
		return nil, errLocalIdentitySessionScopeDenied
	}
	if err != nil {
		return nil, err
	}
	snapshot.accounts[userID] = account
	snapshot.accountByLoginIdentifier[account.NormalizedLoginIdentifier] = userID
	credentialRows, err := query.Query(
		ctx, postgresCredentialSelect+` WHERE user_id=$1 ORDER BY credential_id FOR UPDATE`, userID,
	)
	if err != nil {
		return nil, errLocalIdentityStoreUnavailable
	}
	for credentialRows.Next() {
		credential, scanErr := scanPostgresLocalCredential(credentialRows)
		if scanErr != nil {
			credentialRows.Close()
			return nil, scanErr
		}
		snapshot.credentials[credential.CredentialID] = credential
		if credential.LifecycleState == localIdentityStateActive {
			snapshot.activeCredentialByUser[userID] = credential.CredentialID
		}
	}
	if credentialRows.Err() != nil {
		credentialRows.Close()
		return nil, errLocalIdentityStoreUnavailable
	}
	credentialRows.Close()
	sessionRows, err := query.Query(
		ctx, postgresSessionSelect+` WHERE user_id=$1 ORDER BY session_id FOR UPDATE`, userID,
	)
	if err != nil {
		return nil, errLocalIdentityStoreUnavailable
	}
	for sessionRows.Next() {
		session, scanErr := scanPostgresWebSession(sessionRows)
		if scanErr != nil {
			sessionRows.Close()
			return nil, scanErr
		}
		snapshot.sessions[session.SessionID] = session
		snapshot.sessionByCredentialDigest[string(session.credentialDigest)] = session.SessionID
	}
	if sessionRows.Err() != nil {
		sessionRows.Close()
		return nil, errLocalIdentityStoreUnavailable
	}
	sessionRows.Close()
	return snapshot, nil
}

func validatePostgresLocalIdentitySelfServiceActor(
	ctx context.Context,
	query postgresLocalIdentitySelfServiceQuery,
	userID string,
	currentSessionID string,
	asOf time.Time,
) error {
	account, err := scanPostgresUserAccount(query.QueryRow(ctx, postgresAccountSelect+` WHERE user_id=$1`, userID))
	if err != nil {
		if errors.Is(err, errLocalIdentityNotFound) {
			return errLocalIdentitySessionScopeDenied
		}
		return err
	}
	session, err := scanPostgresWebSession(query.QueryRow(
		ctx, postgresSessionSelect+` WHERE session_id=$1`, currentSessionID,
	))
	if err != nil {
		if errors.Is(err, errLocalIdentityNotFound) {
			return errLocalIdentitySessionScopeDenied
		}
		return err
	}
	if account.LifecycleState != localIdentityStateActive || session.UserID != userID ||
		session.LifecycleState != localIdentityStateActive || !session.ExpiresAt.After(asOf.UTC()) {
		return errLocalIdentitySessionScopeDenied
	}
	return nil
}
