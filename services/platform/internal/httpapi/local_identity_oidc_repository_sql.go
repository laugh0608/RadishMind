package httpapi

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

const sqliteOIDCAuthorizationTransactionSelect = `SELECT
    schema_version, transaction_id, intent, user_id, session_id, session_version, return_to, state_digest, nonce_digest, policy_digest,
    code_verifier, lifecycle_state, record_version, created_at_unix_nano, updated_at_unix_nano,
    expires_at_unix_nano, consumed_at_unix_nano, audit_ref
    FROM local_identity_oidc_authorization_transactions`

const postgresOIDCAuthorizationTransactionSelect = `SELECT
    schema_version, transaction_id, intent, user_id, session_id, session_version, return_to, state_digest, nonce_digest, policy_digest,
    code_verifier, lifecycle_state, record_version, created_at, updated_at, expires_at, consumed_at, audit_ref
    FROM local_identity_oidc_authorization_transactions`

func (repository *sqliteLocalIdentityRepository) CreateOIDCAccountAndWebSession(
	ctx context.Context,
	account UserAccount,
	binding ExternalIdentityBinding,
	session WebSession,
) error {
	if repository == nil || repository.database == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validLocalOIDCAccountAndSessionRegistration(account, binding, session) {
		return errLocalIdentityContractMismatch
	}
	transaction, err := repository.database.BeginTx(identityContext(ctx), nil)
	if err != nil {
		return errLocalIdentityStoreUnavailable
	}
	defer func() { _ = transaction.Rollback() }()
	createdAt, _ := localIdentityUnixNano(account.CreatedAt)
	updatedAt, _ := localIdentityUnixNano(account.UpdatedAt)
	if _, err := transaction.ExecContext(identityContext(ctx), `INSERT INTO local_user_accounts
        (user_id, schema_version, login_identifier, normalized_login_identifier, display_name, lifecycle_state,
         record_version, created_at_unix_nano, updated_at_unix_nano, disabled_at_unix_nano, audit_ref)
        VALUES (?,?,?,?,?,?,?,?,?,?,?)`, account.UserID, account.SchemaVersion, account.LoginIdentifier,
		account.NormalizedLoginIdentifier, account.DisplayName, account.LifecycleState, account.RecordVersion,
		createdAt, updatedAt, nil, account.AuditRef); err != nil {
		return sqliteConflictOrUnavailable(err, errLocalIdentityExternalConflict)
	}
	bindingCreatedAt, _ := localIdentityUnixNano(binding.CreatedAt)
	bindingUpdatedAt, _ := localIdentityUnixNano(binding.UpdatedAt)
	if _, err := transaction.ExecContext(identityContext(ctx), `INSERT INTO external_identity_bindings
        (binding_id, user_id, schema_version, issuer, subject_value, lifecycle_state, record_version,
         created_at_unix_nano, updated_at_unix_nano, revoked_at_unix_nano, audit_ref)
        VALUES (?,?,?,?,?,?,?,?,?,?,?)`, binding.BindingID, binding.UserID, binding.SchemaVersion, binding.Issuer,
		binding.Subject, binding.LifecycleState, binding.RecordVersion, bindingCreatedAt, bindingUpdatedAt, nil,
		binding.AuditRef); err != nil {
		return sqliteConflictOrUnavailable(err, errLocalIdentityExternalConflict)
	}
	sessionCreatedAt, _ := localIdentityUnixNano(session.CreatedAt)
	sessionUpdatedAt, _ := localIdentityUnixNano(session.UpdatedAt)
	lastVerifiedAt, _ := localIdentityUnixNano(session.LastVerifiedAt)
	expiresAt, _ := localIdentityUnixNano(session.ExpiresAt)
	if _, err := transaction.ExecContext(identityContext(ctx), `INSERT INTO local_web_sessions
        (session_id, user_id, schema_version, credential_digest, authentication_method, authentication_source_ref,
         policy_version, lifecycle_state, record_version, created_at_unix_nano, updated_at_unix_nano,
         last_verified_at_unix_nano, expires_at_unix_nano, revoked_at_unix_nano, audit_ref)
        VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, session.SessionID, session.UserID, session.SchemaVersion,
		session.credentialDigest, session.AuthenticationMethod, session.AuthenticationSourceRef, session.PolicyVersion,
		session.LifecycleState, session.RecordVersion, sessionCreatedAt, sessionUpdatedAt, lastVerifiedAt, expiresAt,
		nil, session.AuditRef); err != nil {
		return sqliteConflictOrUnavailable(err, errLocalIdentityExternalConflict)
	}
	if err := transaction.Commit(); err != nil {
		return errLocalIdentityStoreUnavailable
	}
	return nil
}

func (repository *postgresLocalIdentityRepository) CreateOIDCAccountAndWebSession(
	ctx context.Context,
	account UserAccount,
	binding ExternalIdentityBinding,
	session WebSession,
) error {
	if repository == nil || repository.pool == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validLocalOIDCAccountAndSessionRegistration(account, binding, session) {
		return errLocalIdentityContractMismatch
	}
	transaction, err := repository.pool.Begin(identityContext(ctx))
	if err != nil {
		return errLocalIdentityStoreUnavailable
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	if _, err := transaction.Exec(identityContext(ctx), `INSERT INTO local_user_accounts
        (user_id, schema_version, login_identifier, normalized_login_identifier, display_name, lifecycle_state,
         record_version, created_at, updated_at, disabled_at, audit_ref)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, account.UserID, account.SchemaVersion,
		account.LoginIdentifier, account.NormalizedLoginIdentifier, account.DisplayName, account.LifecycleState,
		account.RecordVersion, account.CreatedAt, account.UpdatedAt, account.DisabledAt, account.AuditRef); err != nil {
		return postgresIdentityConflictOrUnavailable(err, errLocalIdentityExternalConflict)
	}
	if _, err := transaction.Exec(identityContext(ctx), `INSERT INTO external_identity_bindings
        (binding_id, user_id, schema_version, issuer, subject_value, lifecycle_state, record_version,
         created_at, updated_at, revoked_at, audit_ref) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		binding.BindingID, binding.UserID, binding.SchemaVersion, binding.Issuer, binding.Subject,
		binding.LifecycleState, binding.RecordVersion, binding.CreatedAt, binding.UpdatedAt, binding.RevokedAt,
		binding.AuditRef); err != nil {
		return postgresIdentityConflictOrUnavailable(err, errLocalIdentityExternalConflict)
	}
	if _, err := transaction.Exec(identityContext(ctx), `INSERT INTO local_web_sessions
        (session_id, user_id, schema_version, credential_digest, authentication_method, authentication_source_ref,
         policy_version, lifecycle_state, record_version, created_at, updated_at, last_verified_at, expires_at,
         revoked_at, audit_ref) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		session.SessionID, session.UserID, session.SchemaVersion, session.credentialDigest,
		session.AuthenticationMethod, session.AuthenticationSourceRef, session.PolicyVersion, session.LifecycleState,
		session.RecordVersion, session.CreatedAt, session.UpdatedAt, session.LastVerifiedAt, session.ExpiresAt,
		session.RevokedAt, session.AuditRef); err != nil {
		return postgresIdentityConflictOrUnavailable(err, errLocalIdentityExternalConflict)
	}
	if err := transaction.Commit(identityContext(ctx)); err != nil {
		return errLocalIdentityStoreUnavailable
	}
	return nil
}

func (repository *sqliteLocalIdentityRepository) CreateOIDCAuthorizationTransaction(
	ctx context.Context,
	transaction LocalIdentityOIDCAuthorizationTransaction,
) error {
	if repository == nil || repository.database == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validLocalIdentityOIDCAuthorizationTransaction(transaction) || transaction.LifecycleState != localIdentityOIDCTransactionPending {
		return errLocalIdentityContractMismatch
	}
	createdAt, _ := localIdentityUnixNano(transaction.CreatedAt)
	updatedAt, _ := localIdentityUnixNano(transaction.UpdatedAt)
	expiresAt, _ := localIdentityUnixNano(transaction.ExpiresAt)
	_, err := repository.database.ExecContext(identityContext(ctx), `INSERT INTO local_identity_oidc_authorization_transactions
        (transaction_id, schema_version, intent, user_id, session_id, session_version, return_to, state_digest, nonce_digest, policy_digest,
         code_verifier, lifecycle_state, record_version, created_at_unix_nano, updated_at_unix_nano,
         expires_at_unix_nano, consumed_at_unix_nano, audit_ref)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, transaction.TransactionID, transaction.SchemaVersion,
		transaction.Intent, transaction.UserID, transaction.SessionID, transaction.SessionVersion, transaction.ReturnTo, transaction.stateDigest, transaction.nonceDigest,
		transaction.policyDigest, transaction.codeVerifier, transaction.LifecycleState, transaction.RecordVersion,
		createdAt, updatedAt, expiresAt, nil, transaction.AuditRef)
	if err != nil {
		return sqliteConflictOrUnavailable(err, errLocalIdentityIdentifierConflict)
	}
	return nil
}

func (repository *postgresLocalIdentityRepository) CreateOIDCAuthorizationTransaction(
	ctx context.Context,
	transaction LocalIdentityOIDCAuthorizationTransaction,
) error {
	if repository == nil || repository.pool == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validLocalIdentityOIDCAuthorizationTransaction(transaction) || transaction.LifecycleState != localIdentityOIDCTransactionPending {
		return errLocalIdentityContractMismatch
	}
	_, err := repository.pool.Exec(identityContext(ctx), `INSERT INTO local_identity_oidc_authorization_transactions
        (transaction_id, schema_version, intent, user_id, session_id, session_version, return_to, state_digest, nonce_digest, policy_digest,
         code_verifier, lifecycle_state, record_version, created_at, updated_at, expires_at, consumed_at, audit_ref)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`, transaction.TransactionID,
		transaction.SchemaVersion, transaction.Intent, transaction.UserID, transaction.SessionID, transaction.SessionVersion,
		transaction.ReturnTo, transaction.stateDigest, transaction.nonceDigest, transaction.policyDigest, transaction.codeVerifier,
		transaction.LifecycleState, transaction.RecordVersion, transaction.CreatedAt, transaction.UpdatedAt,
		transaction.ExpiresAt, transaction.ConsumedAt, transaction.AuditRef)
	if err != nil {
		return postgresIdentityConflictOrUnavailable(err, errLocalIdentityIdentifierConflict)
	}
	return nil
}

func (repository *sqliteLocalIdentityRepository) ConsumeOIDCAuthorizationTransaction(
	ctx context.Context,
	stateDigest [sha256.Size]byte,
	now time.Time,
) (LocalIdentityOIDCAuthorizationTransaction, error) {
	if repository == nil || repository.database == nil {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityStoreUnavailable
	}
	repository.oidcConsumeMu.Lock()
	defer repository.oidcConsumeMu.Unlock()
	now = now.UTC()
	if now.IsZero() {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityContractMismatch
	}
	databaseTransaction, err := repository.database.BeginTx(identityContext(ctx), nil)
	if err != nil {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityStoreUnavailable
	}
	defer func() { _ = databaseTransaction.Rollback() }()
	transaction, err := scanSQLiteOIDCAuthorizationTransaction(databaseTransaction.QueryRowContext(
		identityContext(ctx), sqliteOIDCAuthorizationTransactionSelect+` WHERE state_digest=?`, stateDigest[:],
	))
	if errors.Is(err, errLocalIdentityNotFound) || err == nil && transaction.LifecycleState != localIdentityOIDCTransactionPending {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityOIDCStateInvalid
	}
	if err != nil {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityStoreUnavailable
	}
	newState := localIdentityOIDCTransactionConsumed
	resultErr := error(nil)
	if !transaction.ExpiresAt.After(now) {
		newState = localIdentityOIDCTransactionExpired
		resultErr = errLocalIdentityOIDCStateExpired
	}
	nowNano, _ := localIdentityUnixNano(now)
	command, err := databaseTransaction.ExecContext(identityContext(ctx), `UPDATE local_identity_oidc_authorization_transactions SET
        lifecycle_state=?, record_version=record_version+1, updated_at_unix_nano=?, consumed_at_unix_nano=?, code_verifier=''
        WHERE transaction_id=? AND record_version=? AND lifecycle_state='pending'`, newState, nowNano, nowNano,
		transaction.TransactionID, transaction.RecordVersion)
	if err != nil {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityStoreUnavailable
	}
	if affected, err := command.RowsAffected(); err != nil || affected != 1 {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityOIDCStateInvalid
	}
	if err := databaseTransaction.Commit(); err != nil {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityStoreUnavailable
	}
	if resultErr != nil {
		return LocalIdentityOIDCAuthorizationTransaction{}, resultErr
	}
	transaction.LifecycleState = newState
	transaction.RecordVersion++
	transaction.UpdatedAt = now
	transaction.ConsumedAt = timePointer(now)
	return transaction, nil
}

func (repository *postgresLocalIdentityRepository) ConsumeOIDCAuthorizationTransaction(
	ctx context.Context,
	stateDigest [sha256.Size]byte,
	now time.Time,
) (LocalIdentityOIDCAuthorizationTransaction, error) {
	if repository == nil || repository.pool == nil {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityStoreUnavailable
	}
	now = now.UTC()
	if now.IsZero() {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityContractMismatch
	}
	databaseTransaction, err := repository.pool.Begin(identityContext(ctx))
	if err != nil {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityStoreUnavailable
	}
	defer func() { _ = databaseTransaction.Rollback(context.Background()) }()
	transaction, err := scanPostgresOIDCAuthorizationTransaction(databaseTransaction.QueryRow(
		identityContext(ctx), postgresOIDCAuthorizationTransactionSelect+` WHERE state_digest=$1 FOR UPDATE`, stateDigest[:],
	))
	if errors.Is(err, errLocalIdentityNotFound) || err == nil && transaction.LifecycleState != localIdentityOIDCTransactionPending {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityOIDCStateInvalid
	}
	if err != nil {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityStoreUnavailable
	}
	newState := localIdentityOIDCTransactionConsumed
	resultErr := error(nil)
	if !transaction.ExpiresAt.After(now) {
		newState = localIdentityOIDCTransactionExpired
		resultErr = errLocalIdentityOIDCStateExpired
	}
	command, err := databaseTransaction.Exec(identityContext(ctx), `UPDATE local_identity_oidc_authorization_transactions SET
        lifecycle_state=$1, record_version=record_version+1, updated_at=$2, consumed_at=$2, code_verifier=''
        WHERE transaction_id=$3 AND record_version=$4 AND lifecycle_state='pending'`, newState, now,
		transaction.TransactionID, transaction.RecordVersion)
	if err != nil {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityStoreUnavailable
	}
	if command.RowsAffected() != 1 {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityOIDCStateInvalid
	}
	if err := databaseTransaction.Commit(identityContext(ctx)); err != nil {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityStoreUnavailable
	}
	if resultErr != nil {
		return LocalIdentityOIDCAuthorizationTransaction{}, resultErr
	}
	transaction.LifecycleState = newState
	transaction.RecordVersion++
	transaction.UpdatedAt = now
	transaction.ConsumedAt = timePointer(now)
	return transaction, nil
}

func scanSQLiteOIDCAuthorizationTransaction(row localIdentitySQLRow) (LocalIdentityOIDCAuthorizationTransaction, error) {
	var transaction LocalIdentityOIDCAuthorizationTransaction
	var createdAtNano, updatedAtNano, expiresAtNano int64
	var consumedAtNano sql.NullInt64
	err := row.Scan(&transaction.SchemaVersion, &transaction.TransactionID, &transaction.Intent, &transaction.UserID,
		&transaction.SessionID, &transaction.SessionVersion, &transaction.ReturnTo, &transaction.stateDigest, &transaction.nonceDigest, &transaction.policyDigest,
		&transaction.codeVerifier, &transaction.LifecycleState, &transaction.RecordVersion, &createdAtNano,
		&updatedAtNano, &expiresAtNano, &consumedAtNano, &transaction.AuditRef)
	if err != nil {
		return LocalIdentityOIDCAuthorizationTransaction{}, normalizeSQLiteReadError(err)
	}
	transaction.CreatedAt = time.Unix(0, createdAtNano).UTC()
	transaction.UpdatedAt = time.Unix(0, updatedAtNano).UTC()
	transaction.ExpiresAt = time.Unix(0, expiresAtNano).UTC()
	if consumedAtNano.Valid {
		value := time.Unix(0, consumedAtNano.Int64).UTC()
		transaction.ConsumedAt = &value
	}
	if !validLocalIdentityOIDCAuthorizationTransaction(transaction) {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityStoreUnavailable
	}
	return transaction, nil
}

func scanPostgresOIDCAuthorizationTransaction(row pgx.Row) (LocalIdentityOIDCAuthorizationTransaction, error) {
	var transaction LocalIdentityOIDCAuthorizationTransaction
	var consumedAt sql.NullTime
	err := row.Scan(&transaction.SchemaVersion, &transaction.TransactionID, &transaction.Intent, &transaction.UserID,
		&transaction.SessionID, &transaction.SessionVersion, &transaction.ReturnTo, &transaction.stateDigest, &transaction.nonceDigest, &transaction.policyDigest,
		&transaction.codeVerifier, &transaction.LifecycleState, &transaction.RecordVersion, &transaction.CreatedAt,
		&transaction.UpdatedAt, &transaction.ExpiresAt, &consumedAt, &transaction.AuditRef)
	if errors.Is(err, pgx.ErrNoRows) {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityNotFound
	}
	if err != nil {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityStoreUnavailable
	}
	transaction.CreatedAt = transaction.CreatedAt.UTC()
	transaction.UpdatedAt = transaction.UpdatedAt.UTC()
	transaction.ExpiresAt = transaction.ExpiresAt.UTC()
	if consumedAt.Valid {
		value := consumedAt.Time.UTC()
		transaction.ConsumedAt = &value
	}
	if !validLocalIdentityOIDCAuthorizationTransaction(transaction) {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityStoreUnavailable
	}
	return transaction, nil
}
