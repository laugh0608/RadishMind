package httpapi

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"sync"
	"time"

	modernsqlite "modernc.org/sqlite"
)

type sqliteLocalIdentityRepository struct {
	database      *sql.DB
	oidcConsumeMu sync.Mutex
}

type localIdentitySQLRow interface {
	Scan(...any) error
}

func newSQLiteLocalIdentityRepository(database *sql.DB) *sqliteLocalIdentityRepository {
	return &sqliteLocalIdentityRepository{database: database}
}

func (repository *sqliteLocalIdentityRepository) CreateAccount(ctx context.Context, account UserAccount, credential LocalCredential) error {
	if repository == nil || repository.database == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validUserAccount(account) || !validLocalCredential(credential) || account.LifecycleState != localIdentityStateActive ||
		credential.LifecycleState != localIdentityStateActive || account.UserID != credential.UserID || !account.CreatedAt.Equal(credential.CreatedAt) {
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
		return sqliteConflictOrUnavailable(err, errLocalIdentityIdentifierConflict)
	}
	credentialCreatedAt, _ := localIdentityUnixNano(credential.CreatedAt)
	credentialUpdatedAt, _ := localIdentityUnixNano(credential.UpdatedAt)
	if _, err := transaction.ExecContext(identityContext(ctx), `INSERT INTO local_credentials
        (credential_id, user_id, schema_version, algorithm, policy_version, iterations, key_length, salt,
         derived_key, lifecycle_state, record_version, created_at_unix_nano, updated_at_unix_nano, audit_ref)
        VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, credential.CredentialID, credential.UserID, credential.SchemaVersion,
		credential.Algorithm, credential.PolicyVersion, credential.Iterations, credential.KeyLength, credential.salt,
		credential.derivedKey, credential.LifecycleState, credential.RecordVersion, credentialCreatedAt,
		credentialUpdatedAt, credential.AuditRef); err != nil {
		return sqliteConflictOrUnavailable(err, errLocalIdentityIdentifierConflict)
	}
	if err := transaction.Commit(); err != nil {
		return errLocalIdentityStoreUnavailable
	}
	return nil
}

func (repository *sqliteLocalIdentityRepository) CreateAccountAndWebSession(
	ctx context.Context,
	account UserAccount,
	credential LocalCredential,
	session WebSession,
) error {
	if repository == nil || repository.database == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validLocalAccountAndSessionRegistration(account, credential, session) {
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
		return sqliteConflictOrUnavailable(err, errLocalIdentityIdentifierConflict)
	}
	credentialCreatedAt, _ := localIdentityUnixNano(credential.CreatedAt)
	credentialUpdatedAt, _ := localIdentityUnixNano(credential.UpdatedAt)
	if _, err := transaction.ExecContext(identityContext(ctx), `INSERT INTO local_credentials
        (credential_id, user_id, schema_version, algorithm, policy_version, iterations, key_length, salt,
         derived_key, lifecycle_state, record_version, created_at_unix_nano, updated_at_unix_nano, audit_ref)
        VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, credential.CredentialID, credential.UserID, credential.SchemaVersion,
		credential.Algorithm, credential.PolicyVersion, credential.Iterations, credential.KeyLength, credential.salt,
		credential.derivedKey, credential.LifecycleState, credential.RecordVersion, credentialCreatedAt,
		credentialUpdatedAt, credential.AuditRef); err != nil {
		return sqliteConflictOrUnavailable(err, errLocalIdentityIdentifierConflict)
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
		return sqliteConflictOrUnavailable(err, errLocalIdentityIdentifierConflict)
	}
	if err := transaction.Commit(); err != nil {
		return errLocalIdentityStoreUnavailable
	}
	return nil
}

func (repository *sqliteLocalIdentityRepository) ReadAccount(ctx context.Context, userID string) (UserAccount, error) {
	if repository == nil || repository.database == nil {
		return UserAccount{}, errLocalIdentityStoreUnavailable
	}
	return scanSQLiteUserAccount(repository.database.QueryRowContext(identityContext(ctx), sqliteAccountSelect+` WHERE user_id=?`, strings.TrimSpace(userID)))
}

func (repository *sqliteLocalIdentityRepository) FindAccountByLoginIdentifier(ctx context.Context, rawIdentifier string) (UserAccount, error) {
	identifier, err := NormalizeLocalLoginIdentifier(rawIdentifier)
	if err != nil {
		return UserAccount{}, err
	}
	if repository == nil || repository.database == nil {
		return UserAccount{}, errLocalIdentityStoreUnavailable
	}
	return scanSQLiteUserAccount(repository.database.QueryRowContext(identityContext(ctx), sqliteAccountSelect+` WHERE normalized_login_identifier=?`, identifier))
}

func (repository *sqliteLocalIdentityRepository) DisableAccount(ctx context.Context, userID string, expectedVersion int, disabledAt time.Time, auditRef string) (UserAccount, error) {
	if repository == nil || repository.database == nil {
		return UserAccount{}, errLocalIdentityStoreUnavailable
	}
	disabledAt = disabledAt.UTC()
	if !localUserIDPattern.MatchString(userID) || expectedVersion < 1 || disabledAt.IsZero() || !validAuditRef(auditRef) {
		return UserAccount{}, errLocalIdentityContractMismatch
	}
	disabledAtUnixNano, _ := localIdentityUnixNano(disabledAt)
	transaction, err := repository.database.BeginTx(identityContext(ctx), nil)
	if err != nil {
		return UserAccount{}, errLocalIdentityStoreUnavailable
	}
	defer func() { _ = transaction.Rollback() }()
	account, err := scanSQLiteUserAccount(transaction.QueryRowContext(identityContext(ctx), `UPDATE local_user_accounts SET
        lifecycle_state='disabled', record_version=record_version+1, updated_at_unix_nano=?, disabled_at_unix_nano=?, audit_ref=?
        WHERE user_id=? AND record_version=? AND lifecycle_state='active' RETURNING
        schema_version, user_id, login_identifier, normalized_login_identifier, display_name, lifecycle_state,
        record_version, created_at_unix_nano, updated_at_unix_nano, disabled_at_unix_nano, audit_ref`,
		disabledAtUnixNano, disabledAtUnixNano, auditRef, userID, expectedVersion))
	if err != nil {
		if !errors.Is(err, errLocalIdentityNotFound) {
			return UserAccount{}, errLocalIdentityStoreUnavailable
		}
		var currentVersion int
		var currentState string
		readErr := transaction.QueryRowContext(identityContext(ctx), `SELECT record_version, lifecycle_state
            FROM local_user_accounts WHERE user_id=?`, userID).Scan(&currentVersion, &currentState)
		if errors.Is(readErr, sql.ErrNoRows) {
			return UserAccount{}, errLocalIdentityNotFound
		}
		if readErr != nil {
			return UserAccount{}, errLocalIdentityStoreUnavailable
		}
		return UserAccount{}, errLocalIdentityVersionConflict
	}
	if _, err := transaction.ExecContext(identityContext(ctx), `UPDATE local_web_sessions SET
        lifecycle_state='revoked', record_version=record_version+1, updated_at_unix_nano=?, revoked_at_unix_nano=?, audit_ref=?
        WHERE user_id=? AND lifecycle_state='active'`, disabledAtUnixNano, disabledAtUnixNano, auditRef, userID); err != nil {
		return UserAccount{}, errLocalIdentityStoreUnavailable
	}
	if err := transaction.Commit(); err != nil {
		return UserAccount{}, errLocalIdentityStoreUnavailable
	}
	return account, nil
}

func (repository *sqliteLocalIdentityRepository) ReadActiveCredential(ctx context.Context, userID string) (LocalCredential, error) {
	if repository == nil || repository.database == nil {
		return LocalCredential{}, errLocalIdentityStoreUnavailable
	}
	return scanSQLiteLocalCredential(repository.database.QueryRowContext(identityContext(ctx), sqliteCredentialSelect+`
        WHERE user_id=? AND lifecycle_state='active'`, strings.TrimSpace(userID)))
}

func (repository *sqliteLocalIdentityRepository) ReplaceCredential(ctx context.Context, userID string, expectedCredentialID string, expectedVersion int, replacement LocalCredential) error {
	if repository == nil || repository.database == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validLocalCredential(replacement) || replacement.UserID != userID || replacement.LifecycleState != localIdentityStateActive ||
		!localCredentialIDPattern.MatchString(expectedCredentialID) || expectedVersion < 1 {
		return errLocalIdentityContractMismatch
	}
	createdAt, _ := localIdentityUnixNano(replacement.CreatedAt)
	updatedAt, _ := localIdentityUnixNano(replacement.UpdatedAt)
	transaction, err := repository.database.BeginTx(identityContext(ctx), nil)
	if err != nil {
		return errLocalIdentityStoreUnavailable
	}
	defer func() { _ = transaction.Rollback() }()
	result, err := transaction.ExecContext(identityContext(ctx), `UPDATE local_credentials SET lifecycle_state='superseded',
        record_version=record_version+1, updated_at_unix_nano=?, audit_ref=?
		WHERE credential_id=? AND user_id=? AND record_version=? AND lifecycle_state='active'
		AND created_at_unix_nano < ? AND EXISTS (
			SELECT 1 FROM local_user_accounts WHERE user_id=? AND lifecycle_state='active'
		)`, createdAt, replacement.AuditRef, expectedCredentialID, userID, expectedVersion, createdAt, userID)
	if err != nil {
		return errLocalIdentityStoreUnavailable
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return errLocalIdentityStoreUnavailable
	}
	if affected != 1 {
		var accountState string
		if err := transaction.QueryRowContext(identityContext(ctx), `SELECT lifecycle_state FROM local_user_accounts
			WHERE user_id=?`, userID).Scan(&accountState); errors.Is(err, sql.ErrNoRows) {
			return errLocalIdentityNotFound
		} else if err != nil {
			return errLocalIdentityStoreUnavailable
		} else if accountState != localIdentityStateActive {
			return errLocalIdentityAccountInactive
		}
		return errLocalIdentityVersionConflict
	}
	if _, err := transaction.ExecContext(identityContext(ctx), `INSERT INTO local_credentials
        (credential_id, user_id, schema_version, algorithm, policy_version, iterations, key_length, salt,
         derived_key, lifecycle_state, record_version, created_at_unix_nano, updated_at_unix_nano, audit_ref)
        VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, replacement.CredentialID, replacement.UserID, replacement.SchemaVersion,
		replacement.Algorithm, replacement.PolicyVersion, replacement.Iterations, replacement.KeyLength, replacement.salt,
		replacement.derivedKey, replacement.LifecycleState, replacement.RecordVersion, createdAt, updatedAt,
		replacement.AuditRef); err != nil {
		return sqliteConflictOrUnavailable(err, errLocalIdentityIdentifierConflict)
	}
	if err := transaction.Commit(); err != nil {
		return errLocalIdentityStoreUnavailable
	}
	return nil
}

func (repository *sqliteLocalIdentityRepository) BindExternalIdentity(ctx context.Context, binding ExternalIdentityBinding) error {
	if repository == nil || repository.database == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validExternalIdentityBinding(binding) || binding.LifecycleState != localIdentityStateActive {
		return errLocalIdentityContractMismatch
	}
	if err := repository.requireActiveSQLiteAccount(ctx, binding.UserID); err != nil {
		return err
	}
	createdAt, _ := localIdentityUnixNano(binding.CreatedAt)
	updatedAt, _ := localIdentityUnixNano(binding.UpdatedAt)
	_, err := repository.database.ExecContext(identityContext(ctx), `INSERT INTO external_identity_bindings
        (binding_id, user_id, schema_version, issuer, subject_value, lifecycle_state, record_version,
         created_at_unix_nano, updated_at_unix_nano, revoked_at_unix_nano, audit_ref)
        VALUES (?,?,?,?,?,?,?,?,?,?,?)`, binding.BindingID, binding.UserID, binding.SchemaVersion, binding.Issuer,
		binding.Subject, binding.LifecycleState, binding.RecordVersion, createdAt, updatedAt, nil, binding.AuditRef)
	if err != nil {
		return sqliteConflictOrUnavailable(err, errLocalIdentityExternalConflict)
	}
	return nil
}

func (repository *sqliteLocalIdentityRepository) ResolveExternalIdentity(ctx context.Context, rawIssuer string, rawSubject string) (ExternalIdentityBinding, error) {
	issuer, err := NormalizeExternalIssuer(rawIssuer)
	if err != nil {
		return ExternalIdentityBinding{}, err
	}
	subject, err := NormalizeExternalSubject(rawSubject)
	if err != nil {
		return ExternalIdentityBinding{}, err
	}
	if repository == nil || repository.database == nil {
		return ExternalIdentityBinding{}, errLocalIdentityStoreUnavailable
	}
	return scanSQLiteExternalIdentityBinding(repository.database.QueryRowContext(identityContext(ctx), sqliteBindingSelect+`
        WHERE issuer=? AND subject_value=? AND lifecycle_state='active'`, issuer, subject))
}

func (repository *sqliteLocalIdentityRepository) RevokeExternalIdentity(ctx context.Context, bindingID string, expectedVersion int, revokedAt time.Time, auditRef string) (ExternalIdentityBinding, error) {
	if repository == nil || repository.database == nil {
		return ExternalIdentityBinding{}, errLocalIdentityStoreUnavailable
	}
	revokedAt = revokedAt.UTC()
	if !localBindingIDPattern.MatchString(bindingID) || expectedVersion < 1 || revokedAt.IsZero() || !validAuditRef(auditRef) {
		return ExternalIdentityBinding{}, errLocalIdentityContractMismatch
	}
	revokedAtUnixNano, _ := localIdentityUnixNano(revokedAt)
	binding, err := scanSQLiteExternalIdentityBinding(repository.database.QueryRowContext(identityContext(ctx), `UPDATE external_identity_bindings SET
		lifecycle_state='revoked', record_version=record_version+1, updated_at_unix_nano=?, revoked_at_unix_nano=?, audit_ref=?
		WHERE binding_id=? AND record_version=? AND lifecycle_state='active'
		AND (
			EXISTS (SELECT 1 FROM local_credentials WHERE user_id=external_identity_bindings.user_id AND lifecycle_state='active')
			OR (SELECT COUNT(*) FROM external_identity_bindings AS active_bindings
				WHERE active_bindings.user_id=external_identity_bindings.user_id AND active_bindings.lifecycle_state='active') > 1
		) RETURNING
		schema_version, binding_id, user_id, issuer, subject_value, lifecycle_state, record_version,
		created_at_unix_nano, updated_at_unix_nano, revoked_at_unix_nano, audit_ref`, revokedAtUnixNano,
		revokedAtUnixNano, auditRef, bindingID, expectedVersion))
	if !errors.Is(err, errLocalIdentityNotFound) {
		if err != nil {
			return ExternalIdentityBinding{}, classifySQLiteMutationFailure(err)
		}
		return binding, nil
	}
	var userID, lifecycleState string
	var currentVersion int
	if readErr := repository.database.QueryRowContext(identityContext(ctx), `SELECT user_id, record_version, lifecycle_state
		FROM external_identity_bindings WHERE binding_id=?`, bindingID).Scan(&userID, &currentVersion, &lifecycleState); errors.Is(readErr, sql.ErrNoRows) {
		return ExternalIdentityBinding{}, errLocalIdentityVersionConflict
	} else if readErr != nil {
		return ExternalIdentityBinding{}, errLocalIdentityStoreUnavailable
	}
	if currentVersion == expectedVersion && lifecycleState == localIdentityStateActive {
		return ExternalIdentityBinding{}, errLocalIdentityLastLoginMethodRemoval
	}
	return ExternalIdentityBinding{}, errLocalIdentityVersionConflict
}

func (repository *sqliteLocalIdentityRepository) CreateWebSession(ctx context.Context, session WebSession) error {
	if repository == nil || repository.database == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validWebSession(session) || session.LifecycleState != localIdentityStateActive {
		return errLocalIdentityContractMismatch
	}
	createdAt, _ := localIdentityUnixNano(session.CreatedAt)
	updatedAt, _ := localIdentityUnixNano(session.UpdatedAt)
	lastVerifiedAt, _ := localIdentityUnixNano(session.LastVerifiedAt)
	expiresAt, _ := localIdentityUnixNano(session.ExpiresAt)
	transaction, err := repository.database.BeginTx(identityContext(ctx), nil)
	if err != nil {
		return errLocalIdentityStoreUnavailable
	}
	defer func() { _ = transaction.Rollback() }()
	var accountState string
	if err := transaction.QueryRowContext(identityContext(ctx), `SELECT lifecycle_state FROM local_user_accounts
        WHERE user_id=?`, session.UserID).Scan(&accountState); errors.Is(err, sql.ErrNoRows) {
		return errLocalIdentityNotFound
	} else if err != nil {
		return errLocalIdentityStoreUnavailable
	} else if accountState != localIdentityStateActive {
		return errLocalIdentityAccountInactive
	}
	_, err = transaction.ExecContext(identityContext(ctx), `INSERT INTO local_web_sessions
        (session_id, user_id, schema_version, credential_digest, authentication_method, authentication_source_ref,
         policy_version, lifecycle_state, record_version, created_at_unix_nano, updated_at_unix_nano,
         last_verified_at_unix_nano, expires_at_unix_nano, revoked_at_unix_nano, audit_ref)
        VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, session.SessionID, session.UserID, session.SchemaVersion,
		session.credentialDigest, session.AuthenticationMethod, session.AuthenticationSourceRef, session.PolicyVersion,
		session.LifecycleState, session.RecordVersion, createdAt, updatedAt, lastVerifiedAt, expiresAt, nil,
		session.AuditRef)
	if err != nil {
		return sqliteConflictOrUnavailable(err, errLocalIdentityIdentifierConflict)
	}
	if err := transaction.Commit(); err != nil {
		return errLocalIdentityStoreUnavailable
	}
	return nil
}

func (repository *sqliteLocalIdentityRepository) ResolveWebSession(ctx context.Context, digest [sha256.Size]byte, now time.Time) (WebSession, UserAccount, error) {
	if repository == nil || repository.database == nil {
		return WebSession{}, UserAccount{}, errLocalIdentityStoreUnavailable
	}
	session, err := scanSQLiteWebSession(repository.database.QueryRowContext(identityContext(ctx), sqliteSessionSelect+`
        WHERE credential_digest=?`, digest[:]))
	if errors.Is(err, errLocalIdentityNotFound) {
		return WebSession{}, UserAccount{}, errLocalIdentitySessionInvalid
	}
	if err != nil {
		return WebSession{}, UserAccount{}, err
	}
	if session.LifecycleState != localIdentityStateActive {
		return WebSession{}, UserAccount{}, errLocalIdentitySessionInvalid
	}
	if !session.ExpiresAt.After(now.UTC()) {
		return WebSession{}, UserAccount{}, errLocalIdentitySessionExpired
	}
	account, err := repository.ReadAccount(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, errLocalIdentityNotFound) {
			return WebSession{}, UserAccount{}, errLocalIdentityAccountInactive
		}
		return WebSession{}, UserAccount{}, err
	}
	if account.LifecycleState != localIdentityStateActive {
		return WebSession{}, UserAccount{}, errLocalIdentityAccountInactive
	}
	return session, account, nil
}

func (repository *sqliteLocalIdentityRepository) ReadWebSession(ctx context.Context, sessionID string, now time.Time) (WebSession, UserAccount, error) {
	if repository == nil || repository.database == nil {
		return WebSession{}, UserAccount{}, errLocalIdentityStoreUnavailable
	}
	now = now.UTC()
	session, err := scanSQLiteWebSession(repository.database.QueryRowContext(identityContext(ctx), sqliteSessionSelect+`
		WHERE session_id=?`, strings.TrimSpace(sessionID)))
	if errors.Is(err, errLocalIdentityNotFound) || err == nil && session.LifecycleState != localIdentityStateActive {
		return WebSession{}, UserAccount{}, errLocalIdentitySessionInvalid
	}
	if err != nil {
		return WebSession{}, UserAccount{}, err
	}
	if !session.ExpiresAt.After(now) {
		return WebSession{}, UserAccount{}, errLocalIdentitySessionExpired
	}
	account, err := repository.ReadAccount(ctx, session.UserID)
	if err != nil || account.LifecycleState != localIdentityStateActive {
		return WebSession{}, UserAccount{}, errLocalIdentityAccountInactive
	}
	return session, account, nil
}

func (repository *sqliteLocalIdentityRepository) RevokeWebSession(ctx context.Context, sessionID string, expectedVersion int, revokedAt time.Time, auditRef string) (WebSession, error) {
	if repository == nil || repository.database == nil {
		return WebSession{}, errLocalIdentityStoreUnavailable
	}
	revokedAt = revokedAt.UTC()
	if !localSessionIDPattern.MatchString(sessionID) || expectedVersion < 1 || revokedAt.IsZero() || !validAuditRef(auditRef) {
		return WebSession{}, errLocalIdentityContractMismatch
	}
	revokedAtUnixNano, _ := localIdentityUnixNano(revokedAt)
	session, err := scanSQLiteWebSession(repository.database.QueryRowContext(identityContext(ctx), `UPDATE local_web_sessions SET
        lifecycle_state='revoked', record_version=record_version+1, updated_at_unix_nano=?, revoked_at_unix_nano=?, audit_ref=?
        WHERE session_id=? AND record_version=? AND lifecycle_state='active' RETURNING
        schema_version, session_id, user_id, credential_digest, authentication_method, authentication_source_ref,
        policy_version, lifecycle_state, record_version, created_at_unix_nano, updated_at_unix_nano,
        last_verified_at_unix_nano, expires_at_unix_nano, revoked_at_unix_nano, audit_ref`, revokedAtUnixNano,
		revokedAtUnixNano, auditRef, sessionID, expectedVersion))
	if err != nil {
		return WebSession{}, classifySQLiteMutationFailure(err)
	}
	return session, nil
}

func (repository *sqliteLocalIdentityRepository) CreateRoleAssignment(ctx context.Context, assignment LocalRoleAssignment) error {
	grants, ok := normalizedPermissionGrants(assignment.PermissionGrants)
	assignment.PermissionGrants = grants
	if repository == nil || repository.database == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !ok || assignment.RoleCatalogVersion != "" || assignment.RoleDefinitionDigest != "" ||
		localIdentityContainsManagementPermission(grants) ||
		!validLocalRoleAssignment(assignment) || assignment.LifecycleState != localIdentityStateActive {
		return errLocalIdentityContractMismatch
	}
	if err := repository.requireActiveSQLiteAccount(ctx, assignment.UserID); err != nil {
		return err
	}
	grantsJSON, _ := json.Marshal(assignment.PermissionGrants)
	createdAt, _ := localIdentityUnixNano(assignment.CreatedAt)
	updatedAt, _ := localIdentityUnixNano(assignment.UpdatedAt)
	expiresAt, _ := optionalLocalIdentityUnixNano(assignment.ExpiresAt)
	_, err := repository.database.ExecContext(identityContext(ctx), `INSERT INTO local_role_assignments
        (assignment_id, user_id, schema_version, tenant_ref, workspace_id, role_key, role_catalog_version,
         role_definition_digest, permission_grants_json,
         lifecycle_state, record_version, created_at_unix_nano, updated_at_unix_nano, expires_at_unix_nano,
         revoked_at_unix_nano, audit_ref) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, assignment.AssignmentID,
		assignment.UserID, assignment.SchemaVersion, assignment.TenantRef, assignment.WorkspaceID, assignment.RoleKey,
		nullableLocalIdentityString(assignment.RoleCatalogVersion), nullableLocalIdentityString(assignment.RoleDefinitionDigest),
		string(grantsJSON), assignment.LifecycleState, assignment.RecordVersion, createdAt, updatedAt, expiresAt, nil, assignment.AuditRef)
	if err != nil {
		return sqliteConflictOrUnavailable(err, errLocalIdentityIdentifierConflict)
	}
	return nil
}

func (repository *sqliteLocalIdentityRepository) RevokeRoleAssignment(ctx context.Context, assignmentID string, expectedVersion int, revokedAt time.Time, auditRef string) (LocalRoleAssignment, error) {
	if repository == nil || repository.database == nil {
		return LocalRoleAssignment{}, errLocalIdentityStoreUnavailable
	}
	revokedAt = revokedAt.UTC()
	if !localRoleAssignmentIDPattern.MatchString(assignmentID) || expectedVersion < 1 || revokedAt.IsZero() || !validAuditRef(auditRef) {
		return LocalRoleAssignment{}, errLocalIdentityContractMismatch
	}
	revokedAtUnixNano, _ := localIdentityUnixNano(revokedAt)
	assignment, err := scanSQLiteLocalRoleAssignment(repository.database.QueryRowContext(identityContext(ctx), `UPDATE local_role_assignments SET
        lifecycle_state='revoked', record_version=record_version+1, updated_at_unix_nano=?, revoked_at_unix_nano=?, audit_ref=?
		WHERE assignment_id=? AND record_version=? AND lifecycle_state='active'
		AND role_catalog_version IS NULL AND role_definition_digest IS NULL RETURNING
        schema_version, assignment_id, user_id, tenant_ref, workspace_id, role_key, role_catalog_version,
        role_definition_digest, permission_grants_json,
        lifecycle_state, record_version, created_at_unix_nano, updated_at_unix_nano, expires_at_unix_nano,
        revoked_at_unix_nano, audit_ref`, revokedAtUnixNano, revokedAtUnixNano, auditRef, assignmentID, expectedVersion))
	if err != nil {
		return LocalRoleAssignment{}, classifySQLiteMutationFailure(err)
	}
	return assignment, nil
}

func (repository *sqliteLocalIdentityRepository) CreateWorkspaceMembership(ctx context.Context, membership WorkspaceMembership) error {
	if repository == nil || repository.database == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validWorkspaceMembership(membership) || membership.LifecycleState != localIdentityStateActive {
		return errLocalIdentityContractMismatch
	}
	if err := repository.requireActiveSQLiteAccount(ctx, membership.UserID); err != nil {
		return err
	}
	createdAt, _ := localIdentityUnixNano(membership.CreatedAt)
	updatedAt, _ := localIdentityUnixNano(membership.UpdatedAt)
	expiresAt, _ := optionalLocalIdentityUnixNano(membership.ExpiresAt)
	_, err := repository.database.ExecContext(identityContext(ctx), `INSERT INTO local_workspace_memberships
        (membership_id, user_id, schema_version, tenant_ref, workspace_id, lifecycle_state, record_version,
         created_at_unix_nano, updated_at_unix_nano, expires_at_unix_nano, revoked_at_unix_nano, audit_ref)
        VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, membership.MembershipID, membership.UserID, membership.SchemaVersion,
		membership.TenantRef, membership.WorkspaceID, membership.LifecycleState, membership.RecordVersion,
		createdAt, updatedAt, expiresAt, nil, membership.AuditRef)
	if err != nil {
		return sqliteConflictOrUnavailable(err, errLocalIdentityIdentifierConflict)
	}
	return nil
}

func (repository *sqliteLocalIdentityRepository) RevokeWorkspaceMembership(ctx context.Context, membershipID string, expectedVersion int, revokedAt time.Time, auditRef string) (WorkspaceMembership, error) {
	if repository == nil || repository.database == nil {
		return WorkspaceMembership{}, errLocalIdentityStoreUnavailable
	}
	revokedAt = revokedAt.UTC()
	if !localMembershipIDPattern.MatchString(membershipID) || expectedVersion < 1 || revokedAt.IsZero() || !validAuditRef(auditRef) {
		return WorkspaceMembership{}, errLocalIdentityContractMismatch
	}
	revokedAtUnixNano, _ := localIdentityUnixNano(revokedAt)
	membership, err := scanSQLiteWorkspaceMembership(repository.database.QueryRowContext(identityContext(ctx), `UPDATE local_workspace_memberships SET
        lifecycle_state='revoked', record_version=record_version+1, updated_at_unix_nano=?, revoked_at_unix_nano=?, audit_ref=?
		WHERE membership_id=? AND record_version=? AND lifecycle_state='active'
		AND NOT EXISTS (
			SELECT 1 FROM local_role_assignments r
			WHERE r.user_id=local_workspace_memberships.user_id
			AND r.tenant_ref=local_workspace_memberships.tenant_ref
			AND r.workspace_id=local_workspace_memberships.workspace_id
			AND r.lifecycle_state='active'
		) RETURNING
        schema_version, membership_id, user_id, tenant_ref, workspace_id, lifecycle_state, record_version,
        created_at_unix_nano, updated_at_unix_nano, expires_at_unix_nano, revoked_at_unix_nano, audit_ref`,
		revokedAtUnixNano, revokedAtUnixNano, auditRef, membershipID, expectedVersion))
	if err != nil {
		return WorkspaceMembership{}, classifySQLiteMutationFailure(err)
	}
	return membership, nil
}

func (repository *sqliteLocalIdentityRepository) AuthorizeWorkspace(ctx context.Context, userID string, tenantRef string, workspaceID string, required []string, now time.Time) (LocalWorkspaceAuthorization, error) {
	required, ok := normalizedPermissionGrants(required)
	if repository == nil || repository.database == nil {
		return LocalWorkspaceAuthorization{}, errLocalIdentityStoreUnavailable
	}
	if !ok || !localUserIDPattern.MatchString(userID) || !validControlPlaneReadAuthReference(tenantRef, false) ||
		!validControlPlaneReadAuthReference(workspaceID, false) {
		return LocalWorkspaceAuthorization{}, errLocalIdentityContractMismatch
	}
	account, err := repository.ReadAccount(ctx, userID)
	if err != nil {
		if errors.Is(err, errLocalIdentityNotFound) {
			return LocalWorkspaceAuthorization{}, errLocalIdentityAccountInactive
		}
		return LocalWorkspaceAuthorization{}, err
	}
	if account.LifecycleState != localIdentityStateActive {
		return LocalWorkspaceAuthorization{}, errLocalIdentityAccountInactive
	}
	membership, err := scanSQLiteWorkspaceMembership(repository.database.QueryRowContext(identityContext(ctx), sqliteMembershipSelect+`
        WHERE user_id=? AND tenant_ref=? AND workspace_id=? AND lifecycle_state='active'`, userID, tenantRef, workspaceID))
	if err != nil {
		if errors.Is(err, errLocalIdentityNotFound) {
			return LocalWorkspaceAuthorization{}, errLocalIdentityMembershipDenied
		}
		return LocalWorkspaceAuthorization{}, err
	}
	if membership.ExpiresAt != nil && !membership.ExpiresAt.After(now.UTC()) {
		return LocalWorkspaceAuthorization{}, errLocalIdentityMembershipDenied
	}
	rows, err := repository.database.QueryContext(identityContext(ctx), sqliteRoleAssignmentSelect+`
        WHERE user_id=? AND tenant_ref=? AND (workspace_id='' OR workspace_id=?) AND lifecycle_state='active'
        ORDER BY assignment_id`, userID, tenantRef, workspaceID)
	if err != nil {
		return LocalWorkspaceAuthorization{}, errLocalIdentityStoreUnavailable
	}
	defer rows.Close()
	assignments := make([]LocalRoleAssignment, 0)
	grantSet := make(map[string]struct{})
	for rows.Next() {
		assignment, scanErr := scanSQLiteLocalRoleAssignment(rows)
		if scanErr != nil {
			return LocalWorkspaceAuthorization{}, scanErr
		}
		if assignment.ExpiresAt != nil && !assignment.ExpiresAt.After(now.UTC()) {
			continue
		}
		assignments = append(assignments, assignment)
		for _, grant := range assignment.PermissionGrants {
			grantSet[grant] = struct{}{}
		}
	}
	if rows.Err() != nil {
		return LocalWorkspaceAuthorization{}, errLocalIdentityStoreUnavailable
	}
	return buildLocalWorkspaceAuthorization(account, membership, assignments, grantSet, required)
}

const sqliteAccountSelect = `SELECT schema_version, user_id, login_identifier, normalized_login_identifier,
    display_name, lifecycle_state, record_version, created_at_unix_nano, updated_at_unix_nano,
    disabled_at_unix_nano, audit_ref FROM local_user_accounts`

const sqliteCredentialSelect = `SELECT schema_version, credential_id, user_id, algorithm, policy_version,
    iterations, key_length, salt, derived_key, lifecycle_state, record_version, created_at_unix_nano,
    updated_at_unix_nano, audit_ref FROM local_credentials`

const sqliteBindingSelect = `SELECT schema_version, binding_id, user_id, issuer, subject_value, lifecycle_state,
    record_version, created_at_unix_nano, updated_at_unix_nano, revoked_at_unix_nano, audit_ref
    FROM external_identity_bindings`

const sqliteSessionSelect = `SELECT schema_version, session_id, user_id, credential_digest, authentication_method,
    authentication_source_ref, policy_version, lifecycle_state, record_version, created_at_unix_nano,
    updated_at_unix_nano, last_verified_at_unix_nano, expires_at_unix_nano, revoked_at_unix_nano, audit_ref
    FROM local_web_sessions`

const sqliteRoleAssignmentSelect = `SELECT schema_version, assignment_id, user_id, tenant_ref, workspace_id,
    role_key, role_catalog_version, role_definition_digest, permission_grants_json, lifecycle_state,
    record_version, created_at_unix_nano,
    updated_at_unix_nano, expires_at_unix_nano, revoked_at_unix_nano, audit_ref FROM local_role_assignments`

const sqliteMembershipSelect = `SELECT schema_version, membership_id, user_id, tenant_ref, workspace_id,
    lifecycle_state, record_version, created_at_unix_nano, updated_at_unix_nano, expires_at_unix_nano,
    revoked_at_unix_nano, audit_ref FROM local_workspace_memberships`

func scanSQLiteUserAccount(row localIdentitySQLRow) (UserAccount, error) {
	var account UserAccount
	var createdAt, updatedAt int64
	var disabledAt sql.NullInt64
	err := row.Scan(&account.SchemaVersion, &account.UserID, &account.LoginIdentifier, &account.NormalizedLoginIdentifier,
		&account.DisplayName, &account.LifecycleState, &account.RecordVersion, &createdAt, &updatedAt, &disabledAt,
		&account.AuditRef)
	if err != nil {
		return UserAccount{}, normalizeSQLiteReadError(err)
	}
	account.CreatedAt = time.Unix(0, createdAt).UTC()
	account.UpdatedAt = time.Unix(0, updatedAt).UTC()
	account.DisabledAt = sqliteOptionalTime(disabledAt)
	if !validUserAccount(account) {
		return UserAccount{}, errLocalIdentityStoreUnavailable
	}
	return account, nil
}

func scanSQLiteLocalCredential(row localIdentitySQLRow) (LocalCredential, error) {
	var credential LocalCredential
	var createdAt, updatedAt int64
	err := row.Scan(&credential.SchemaVersion, &credential.CredentialID, &credential.UserID, &credential.Algorithm,
		&credential.PolicyVersion, &credential.Iterations, &credential.KeyLength, &credential.salt,
		&credential.derivedKey, &credential.LifecycleState, &credential.RecordVersion, &createdAt, &updatedAt,
		&credential.AuditRef)
	if err != nil {
		return LocalCredential{}, normalizeSQLiteReadError(err)
	}
	credential.CreatedAt = time.Unix(0, createdAt).UTC()
	credential.UpdatedAt = time.Unix(0, updatedAt).UTC()
	if !validLocalCredential(credential) {
		return LocalCredential{}, errLocalIdentityStoreUnavailable
	}
	return credential, nil
}

func scanSQLiteExternalIdentityBinding(row localIdentitySQLRow) (ExternalIdentityBinding, error) {
	var binding ExternalIdentityBinding
	var createdAt, updatedAt int64
	var revokedAt sql.NullInt64
	err := row.Scan(&binding.SchemaVersion, &binding.BindingID, &binding.UserID, &binding.Issuer, &binding.Subject,
		&binding.LifecycleState, &binding.RecordVersion, &createdAt, &updatedAt, &revokedAt, &binding.AuditRef)
	if err != nil {
		return ExternalIdentityBinding{}, normalizeSQLiteReadError(err)
	}
	binding.CreatedAt = time.Unix(0, createdAt).UTC()
	binding.UpdatedAt = time.Unix(0, updatedAt).UTC()
	binding.RevokedAt = sqliteOptionalTime(revokedAt)
	if !validExternalIdentityBinding(binding) {
		return ExternalIdentityBinding{}, errLocalIdentityStoreUnavailable
	}
	return binding, nil
}

func scanSQLiteWebSession(row localIdentitySQLRow) (WebSession, error) {
	var session WebSession
	var createdAt, updatedAt, lastVerifiedAt, expiresAt int64
	var revokedAt sql.NullInt64
	err := row.Scan(&session.SchemaVersion, &session.SessionID, &session.UserID, &session.credentialDigest,
		&session.AuthenticationMethod, &session.AuthenticationSourceRef, &session.PolicyVersion, &session.LifecycleState,
		&session.RecordVersion, &createdAt, &updatedAt, &lastVerifiedAt, &expiresAt, &revokedAt, &session.AuditRef)
	if err != nil {
		return WebSession{}, normalizeSQLiteReadError(err)
	}
	session.CreatedAt = time.Unix(0, createdAt).UTC()
	session.UpdatedAt = time.Unix(0, updatedAt).UTC()
	session.LastVerifiedAt = time.Unix(0, lastVerifiedAt).UTC()
	session.ExpiresAt = time.Unix(0, expiresAt).UTC()
	session.RevokedAt = sqliteOptionalTime(revokedAt)
	if !validWebSession(session) {
		return WebSession{}, errLocalIdentityStoreUnavailable
	}
	return session, nil
}

func scanSQLiteLocalRoleAssignment(row localIdentitySQLRow) (LocalRoleAssignment, error) {
	var assignment LocalRoleAssignment
	var grantsJSON string
	var roleCatalogVersion, roleDefinitionDigest sql.NullString
	var createdAt, updatedAt int64
	var expiresAt, revokedAt sql.NullInt64
	err := row.Scan(&assignment.SchemaVersion, &assignment.AssignmentID, &assignment.UserID, &assignment.TenantRef,
		&assignment.WorkspaceID, &assignment.RoleKey, &roleCatalogVersion, &roleDefinitionDigest, &grantsJSON, &assignment.LifecycleState,
		&assignment.RecordVersion, &createdAt, &updatedAt, &expiresAt, &revokedAt, &assignment.AuditRef)
	if err != nil {
		return LocalRoleAssignment{}, normalizeSQLiteReadError(err)
	}
	if err := json.Unmarshal([]byte(grantsJSON), &assignment.PermissionGrants); err != nil {
		return LocalRoleAssignment{}, errLocalIdentityStoreUnavailable
	}
	assignment.RoleCatalogVersion = roleCatalogVersion.String
	assignment.RoleDefinitionDigest = roleDefinitionDigest.String
	assignment.CreatedAt = time.Unix(0, createdAt).UTC()
	assignment.UpdatedAt = time.Unix(0, updatedAt).UTC()
	assignment.ExpiresAt = sqliteOptionalTime(expiresAt)
	assignment.RevokedAt = sqliteOptionalTime(revokedAt)
	if !validLocalRoleAssignment(assignment) {
		return LocalRoleAssignment{}, errLocalIdentityStoreUnavailable
	}
	return assignment, nil
}

func scanSQLiteWorkspaceMembership(row localIdentitySQLRow) (WorkspaceMembership, error) {
	var membership WorkspaceMembership
	var createdAt, updatedAt int64
	var expiresAt, revokedAt sql.NullInt64
	err := row.Scan(&membership.SchemaVersion, &membership.MembershipID, &membership.UserID, &membership.TenantRef,
		&membership.WorkspaceID, &membership.LifecycleState, &membership.RecordVersion, &createdAt, &updatedAt,
		&expiresAt, &revokedAt, &membership.AuditRef)
	if err != nil {
		return WorkspaceMembership{}, normalizeSQLiteReadError(err)
	}
	membership.CreatedAt = time.Unix(0, createdAt).UTC()
	membership.UpdatedAt = time.Unix(0, updatedAt).UTC()
	membership.ExpiresAt = sqliteOptionalTime(expiresAt)
	membership.RevokedAt = sqliteOptionalTime(revokedAt)
	if !validWorkspaceMembership(membership) {
		return WorkspaceMembership{}, errLocalIdentityStoreUnavailable
	}
	return membership, nil
}

func (repository *sqliteLocalIdentityRepository) requireActiveSQLiteAccount(ctx context.Context, userID string) error {
	account, err := repository.ReadAccount(ctx, userID)
	if err != nil {
		return err
	}
	if account.LifecycleState != localIdentityStateActive {
		return errLocalIdentityAccountInactive
	}
	return nil
}

func sqliteConflictOrUnavailable(err error, conflict error) error {
	var sqliteError *modernsqlite.Error
	if errors.As(err, &sqliteError) && sqliteError.Code()&0xff == 19 {
		return conflict
	}
	return errLocalIdentityStoreUnavailable
}

func classifySQLiteMutationFailure(err error) error {
	if errors.Is(err, errLocalIdentityNotFound) {
		return errLocalIdentityVersionConflict
	}
	return errLocalIdentityStoreUnavailable
}

func normalizeSQLiteReadError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return errLocalIdentityNotFound
	}
	return errLocalIdentityStoreUnavailable
}

func localIdentityUnixNano(value time.Time) (int64, error) {
	value = value.UTC()
	if value.IsZero() {
		return 0, errLocalIdentityContractMismatch
	}
	unixNano := value.UnixNano()
	if !time.Unix(0, unixNano).UTC().Equal(value) {
		return 0, errLocalIdentityContractMismatch
	}
	return unixNano, nil
}

func optionalLocalIdentityUnixNano(value *time.Time) (any, error) {
	if value == nil {
		return nil, nil
	}
	return localIdentityUnixNano(*value)
}

func sqliteOptionalTime(value sql.NullInt64) *time.Time {
	if !value.Valid {
		return nil
	}
	parsed := time.Unix(0, value.Int64).UTC()
	return &parsed
}

func nullableLocalIdentityString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func buildLocalWorkspaceAuthorization(
	account UserAccount,
	membership WorkspaceMembership,
	assignments []LocalRoleAssignment,
	grantSet map[string]struct{},
	required []string,
) (LocalWorkspaceAuthorization, error) {
	for _, permission := range required {
		if _, granted := grantSet[permission]; !granted {
			return LocalWorkspaceAuthorization{}, errLocalIdentityPermissionDenied
		}
	}
	grants := make([]string, 0, len(grantSet))
	for grant := range grantSet {
		grants = append(grants, grant)
	}
	slices.Sort(grants)
	return LocalWorkspaceAuthorization{
		Account: account, Membership: membership, RoleAssignments: assignments, PermissionGrants: grants,
	}, nil
}

func identityContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
