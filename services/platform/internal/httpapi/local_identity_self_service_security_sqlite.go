package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type sqliteLocalIdentitySelfServiceQuery interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (repository *sqliteLocalIdentityRepository) ListSelfServiceSessions(
	ctx context.Context,
	userID string,
	currentSessionID string,
	query LocalIdentitySelfServiceSessionListQuery,
) (LocalIdentitySelfServiceSessionPage, error) {
	if repository == nil || repository.database == nil {
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
	transaction, err := repository.database.BeginTx(identityContext(ctx), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return LocalIdentitySelfServiceSessionPage{}, errLocalIdentityStoreUnavailable
	}
	defer func() { _ = transaction.Rollback() }()
	if err = validateSQLiteLocalIdentitySelfServiceActor(
		identityContext(ctx), transaction, userID, currentSessionID, filter.requestedAt.UnixNano(),
	); err != nil {
		return LocalIdentitySelfServiceSessionPage{}, err
	}

	statement := sqliteSessionSelect + ` WHERE user_id=? AND created_at_unix_nano<=?`
	arguments := []any{userID, filter.snapshotAt.UnixNano()}
	switch filter.State {
	case localIdentityStateActive:
		statement += ` AND (revoked_at_unix_nano IS NULL OR revoked_at_unix_nano>?) AND expires_at_unix_nano>?`
		arguments = append(arguments, filter.snapshotAt.UnixNano(), filter.snapshotAt.UnixNano())
	case localIdentitySessionEffectiveStateExpired:
		statement += ` AND (revoked_at_unix_nano IS NULL OR revoked_at_unix_nano>?) AND expires_at_unix_nano<=?`
		arguments = append(arguments, filter.snapshotAt.UnixNano(), filter.snapshotAt.UnixNano())
	case localIdentityStateRevoked:
		statement += ` AND revoked_at_unix_nano IS NOT NULL AND revoked_at_unix_nano<=?`
		arguments = append(arguments, filter.snapshotAt.UnixNano())
	}
	if cursor.SessionID != "" {
		anchor, parseErr := localIdentitySelfServiceCursorCreatedAt(cursor)
		if parseErr != nil {
			return LocalIdentitySelfServiceSessionPage{}, errLocalIdentitySessionCursorInvalid
		}
		statement += ` AND (created_at_unix_nano<? OR (created_at_unix_nano=? AND session_id<?))`
		arguments = append(arguments, anchor.UnixNano(), anchor.UnixNano(), cursor.SessionID)
	}
	statement += ` ORDER BY created_at_unix_nano DESC, session_id DESC LIMIT ?`
	arguments = append(arguments, filter.Limit+1)
	rows, err := transaction.QueryContext(identityContext(ctx), statement, arguments...)
	if err != nil {
		return LocalIdentitySelfServiceSessionPage{}, errLocalIdentityStoreUnavailable
	}
	sessions := make([]WebSession, 0, filter.Limit+1)
	for rows.Next() {
		session, scanErr := scanSQLiteWebSession(rows)
		if scanErr != nil {
			_ = rows.Close()
			return LocalIdentitySelfServiceSessionPage{}, scanErr
		}
		sessions = append(sessions, session)
	}
	if rows.Err() != nil || rows.Close() != nil {
		return LocalIdentitySelfServiceSessionPage{}, errLocalIdentityStoreUnavailable
	}
	if err = transaction.Commit(); err != nil {
		return LocalIdentitySelfServiceSessionPage{}, errLocalIdentityStoreUnavailable
	}
	return buildLocalIdentitySelfServiceSessionPage(userID, currentSessionID, filter, sessions)
}

func (repository *sqliteLocalIdentityRepository) RevokeOwnedWebSession(
	ctx context.Context,
	mutation localIdentityOwnedSessionRevocation,
) (LocalIdentitySelfServiceSessionRevocation, error) {
	if repository == nil || repository.database == nil {
		return LocalIdentitySelfServiceSessionRevocation{}, errLocalIdentityStoreUnavailable
	}
	var result LocalIdentitySelfServiceSessionRevocation
	err := repository.mutateLocalIdentitySelfServiceScope(ctx, mutation.UserID, func(
		snapshot *memoryLocalIdentityRepository,
		connection *sql.Conn,
	) error {
		var mutationErr error
		result, mutationErr = snapshot.RevokeOwnedWebSession(ctx, mutation)
		if mutationErr != nil {
			return mutationErr
		}
		revoked := snapshot.sessions[strings.TrimSpace(mutation.TargetSessionID)]
		updatedAt, _ := localIdentityUnixNano(revoked.UpdatedAt)
		resultSet, execErr := connection.ExecContext(identityContext(ctx), `UPDATE local_web_sessions SET
            lifecycle_state='revoked', record_version=record_version+1, updated_at_unix_nano=?,
            revoked_at_unix_nano=?, audit_ref=?
            WHERE session_id=? AND user_id=? AND record_version=? AND lifecycle_state='active'
            AND expires_at_unix_nano>?`, updatedAt, updatedAt, revoked.AuditRef, revoked.SessionID,
			revoked.UserID, mutation.ExpectedVersion, updatedAt)
		if execErr != nil {
			return errLocalIdentityStoreUnavailable
		}
		if affected, affectedErr := resultSet.RowsAffected(); affectedErr != nil || affected != 1 {
			return errLocalIdentitySessionVersionConflict
		}
		return nil
	})
	if err != nil {
		return LocalIdentitySelfServiceSessionRevocation{}, err
	}
	return result, nil
}

func (repository *sqliteLocalIdentityRepository) RevokeOtherWebSessions(
	ctx context.Context,
	mutation localIdentityOtherSessionRevocation,
) (LocalIdentitySelfServiceBulkSessionRevocation, error) {
	if repository == nil || repository.database == nil {
		return LocalIdentitySelfServiceBulkSessionRevocation{}, errLocalIdentityStoreUnavailable
	}
	var result LocalIdentitySelfServiceBulkSessionRevocation
	err := repository.mutateLocalIdentitySelfServiceScope(ctx, mutation.UserID, func(
		snapshot *memoryLocalIdentityRepository,
		connection *sql.Conn,
	) error {
		var mutationErr error
		result, mutationErr = snapshot.RevokeOtherWebSessions(ctx, mutation)
		if mutationErr != nil {
			return mutationErr
		}
		revokedAt := mutation.RevokedAt.UTC().UnixNano()
		resultSet, execErr := connection.ExecContext(identityContext(ctx), `UPDATE local_web_sessions SET
            lifecycle_state='revoked', record_version=record_version+1, updated_at_unix_nano=?,
            revoked_at_unix_nano=?, audit_ref=?
            WHERE user_id=? AND session_id<>? AND lifecycle_state='active' AND expires_at_unix_nano>?`,
			revokedAt, revokedAt, strings.TrimSpace(mutation.AuditRef), strings.TrimSpace(mutation.UserID),
			strings.TrimSpace(mutation.CurrentSessionID), revokedAt)
		if execErr != nil {
			return errLocalIdentityStoreUnavailable
		}
		if affected, affectedErr := resultSet.RowsAffected(); affectedErr != nil || affected != int64(result.RevokedCount) {
			return errLocalIdentitySessionBulkRevokeConflict
		}
		return nil
	})
	if err != nil {
		return LocalIdentitySelfServiceBulkSessionRevocation{}, err
	}
	return result, nil
}

func (repository *sqliteLocalIdentityRepository) RotateLocalCredentialAndRevokeSessions(
	ctx context.Context,
	mutation localIdentityCredentialRotation,
) (LocalIdentitySelfServiceCredentialRotation, error) {
	if repository == nil || repository.database == nil {
		return LocalIdentitySelfServiceCredentialRotation{}, errLocalIdentityStoreUnavailable
	}
	var result LocalIdentitySelfServiceCredentialRotation
	err := repository.mutateLocalIdentitySelfServiceScope(ctx, mutation.UserID, func(
		snapshot *memoryLocalIdentityRepository,
		connection *sql.Conn,
	) error {
		userID := strings.TrimSpace(mutation.UserID)
		currentCredentialID := snapshot.activeCredentialByUser[userID]
		currentCredential := snapshot.credentials[currentCredentialID]
		var mutationErr error
		result, mutationErr = snapshot.RotateLocalCredentialAndRevokeSessions(ctx, mutation)
		if mutationErr != nil {
			return mutationErr
		}
		rotatedAt := mutation.RotatedAt.UTC().UnixNano()
		resultSet, execErr := connection.ExecContext(identityContext(ctx), `UPDATE local_credentials SET
            lifecycle_state='superseded', record_version=record_version+1, updated_at_unix_nano=?, audit_ref=?
            WHERE credential_id=? AND user_id=? AND record_version=? AND lifecycle_state='active'
            AND created_at_unix_nano<?`, rotatedAt, strings.TrimSpace(mutation.AuditRef), currentCredentialID,
			userID, currentCredential.RecordVersion, rotatedAt)
		if execErr != nil {
			return errLocalIdentityStoreUnavailable
		}
		if affected, affectedErr := resultSet.RowsAffected(); affectedErr != nil || affected != 1 {
			return errLocalIdentityCredentialRotationConflict
		}
		replacement := mutation.Replacement
		createdAt, _ := localIdentityUnixNano(replacement.CreatedAt)
		updatedAt, _ := localIdentityUnixNano(replacement.UpdatedAt)
		_, execErr = connection.ExecContext(identityContext(ctx), `INSERT INTO local_credentials
            (credential_id, user_id, schema_version, algorithm, policy_version, iterations, key_length, salt,
             derived_key, lifecycle_state, record_version, created_at_unix_nano, updated_at_unix_nano, audit_ref)
            VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, replacement.CredentialID, replacement.UserID,
			replacement.SchemaVersion, replacement.Algorithm, replacement.PolicyVersion, replacement.Iterations,
			replacement.KeyLength, replacement.salt, replacement.derivedKey, replacement.LifecycleState,
			replacement.RecordVersion, createdAt, updatedAt, replacement.AuditRef)
		if execErr != nil {
			return sqliteConflictOrUnavailable(execErr, errLocalIdentityCredentialRotationConflict)
		}
		sourceRef := "credential:" + currentCredentialID
		resultSet, execErr = connection.ExecContext(identityContext(ctx), `UPDATE local_web_sessions SET
            lifecycle_state='revoked', record_version=record_version+1, updated_at_unix_nano=?,
            revoked_at_unix_nano=?, audit_ref=?
            WHERE user_id=? AND lifecycle_state='active' AND authentication_method='local_password'
            AND authentication_source_ref=?`, rotatedAt, rotatedAt, strings.TrimSpace(mutation.AuditRef),
			userID, sourceRef)
		if execErr != nil {
			return errLocalIdentityStoreUnavailable
		}
		if affected, affectedErr := resultSet.RowsAffected(); affectedErr != nil ||
			affected != int64(result.RevokedSessionCount) {
			return errLocalIdentityCredentialRotationConflict
		}
		return nil
	})
	if err != nil {
		return LocalIdentitySelfServiceCredentialRotation{}, err
	}
	return result, nil
}

func (repository *sqliteLocalIdentityRepository) mutateLocalIdentitySelfServiceScope(
	ctx context.Context,
	userID string,
	operation func(*memoryLocalIdentityRepository, *sql.Conn) error,
) error {
	connection, err := repository.database.Conn(identityContext(ctx))
	if err != nil {
		return errLocalIdentityStoreUnavailable
	}
	defer connection.Close()
	if _, err = connection.ExecContext(identityContext(ctx), "BEGIN IMMEDIATE"); err != nil {
		return errLocalIdentityStoreUnavailable
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = connection.ExecContext(context.Background(), "ROLLBACK")
		}
	}()
	snapshot, err := loadSQLiteLocalIdentitySelfServiceSnapshot(identityContext(ctx), connection, userID)
	if err != nil {
		return err
	}
	if err = operation(snapshot, connection); err != nil {
		return err
	}
	if _, err = connection.ExecContext(identityContext(ctx), "COMMIT"); err != nil {
		return errLocalIdentityStoreUnavailable
	}
	committed = true
	return nil
}

func loadSQLiteLocalIdentitySelfServiceSnapshot(
	ctx context.Context,
	query sqliteLocalIdentitySelfServiceQuery,
	userID string,
) (*memoryLocalIdentityRepository, error) {
	userID = strings.TrimSpace(userID)
	if !localUserIDPattern.MatchString(userID) {
		return nil, errLocalIdentityContractMismatch
	}
	snapshot := newMemoryLocalIdentityRepository()
	account, err := scanSQLiteUserAccount(query.QueryRowContext(ctx, sqliteAccountSelect+` WHERE user_id=?`, userID))
	if errors.Is(err, errLocalIdentityNotFound) {
		return nil, errLocalIdentitySessionScopeDenied
	}
	if err != nil {
		return nil, err
	}
	snapshot.accounts[userID] = account
	snapshot.accountByLoginIdentifier[account.NormalizedLoginIdentifier] = userID
	credentialRows, err := query.QueryContext(ctx, sqliteCredentialSelect+` WHERE user_id=? ORDER BY credential_id`, userID)
	if err != nil {
		return nil, errLocalIdentityStoreUnavailable
	}
	for credentialRows.Next() {
		credential, scanErr := scanSQLiteLocalCredential(credentialRows)
		if scanErr != nil {
			_ = credentialRows.Close()
			return nil, scanErr
		}
		snapshot.credentials[credential.CredentialID] = credential
		if credential.LifecycleState == localIdentityStateActive {
			snapshot.activeCredentialByUser[userID] = credential.CredentialID
		}
	}
	if credentialRows.Err() != nil || credentialRows.Close() != nil {
		return nil, errLocalIdentityStoreUnavailable
	}
	sessionRows, err := query.QueryContext(ctx, sqliteSessionSelect+` WHERE user_id=? ORDER BY session_id`, userID)
	if err != nil {
		return nil, errLocalIdentityStoreUnavailable
	}
	for sessionRows.Next() {
		session, scanErr := scanSQLiteWebSession(sessionRows)
		if scanErr != nil {
			_ = sessionRows.Close()
			return nil, scanErr
		}
		snapshot.sessions[session.SessionID] = session
		snapshot.sessionByCredentialDigest[string(session.credentialDigest)] = session.SessionID
	}
	if sessionRows.Err() != nil || sessionRows.Close() != nil {
		return nil, errLocalIdentityStoreUnavailable
	}
	return snapshot, nil
}

func validateSQLiteLocalIdentitySelfServiceActor(
	ctx context.Context,
	query sqliteLocalIdentitySelfServiceQuery,
	userID string,
	currentSessionID string,
	asOfUnixNano int64,
) error {
	account, err := scanSQLiteUserAccount(query.QueryRowContext(ctx, sqliteAccountSelect+` WHERE user_id=?`, userID))
	if err != nil {
		if errors.Is(err, errLocalIdentityNotFound) {
			return errLocalIdentitySessionScopeDenied
		}
		return err
	}
	session, err := scanSQLiteWebSession(query.QueryRowContext(ctx, sqliteSessionSelect+`
        WHERE session_id=?`, currentSessionID))
	if err != nil {
		if errors.Is(err, errLocalIdentityNotFound) {
			return errLocalIdentitySessionScopeDenied
		}
		return err
	}
	if account.LifecycleState != localIdentityStateActive || session.UserID != userID ||
		session.LifecycleState != localIdentityStateActive || session.ExpiresAt.UnixNano() <= asOfUnixNano {
		return errLocalIdentitySessionScopeDenied
	}
	return nil
}
