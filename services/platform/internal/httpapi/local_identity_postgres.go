package httpapi

import (
	"context"
	"crypto/sha256"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresLocalIdentityRepository struct {
	pool *pgxpool.Pool
}

func newPostgresLocalIdentityRepository(pool *pgxpool.Pool) *postgresLocalIdentityRepository {
	return &postgresLocalIdentityRepository{pool: pool}
}

func (repository *postgresLocalIdentityRepository) CreateAccount(ctx context.Context, account UserAccount, credential LocalCredential) error {
	if repository == nil || repository.pool == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validUserAccount(account) || !validLocalCredential(credential) || account.LifecycleState != localIdentityStateActive ||
		credential.LifecycleState != localIdentityStateActive || account.UserID != credential.UserID || !account.CreatedAt.Equal(credential.CreatedAt) {
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
		return postgresIdentityConflictOrUnavailable(err, errLocalIdentityIdentifierConflict)
	}
	if _, err := transaction.Exec(identityContext(ctx), `INSERT INTO local_credentials
        (credential_id, user_id, schema_version, algorithm, policy_version, iterations, key_length, salt,
         derived_key, lifecycle_state, record_version, created_at, updated_at, audit_ref)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`, credential.CredentialID,
		credential.UserID, credential.SchemaVersion, credential.Algorithm, credential.PolicyVersion,
		credential.Iterations, credential.KeyLength, credential.salt, credential.derivedKey,
		credential.LifecycleState, credential.RecordVersion, credential.CreatedAt, credential.UpdatedAt,
		credential.AuditRef); err != nil {
		return postgresIdentityConflictOrUnavailable(err, errLocalIdentityIdentifierConflict)
	}
	if err := transaction.Commit(identityContext(ctx)); err != nil {
		return errLocalIdentityStoreUnavailable
	}
	return nil
}

func (repository *postgresLocalIdentityRepository) CreateAccountAndWebSession(
	ctx context.Context,
	account UserAccount,
	credential LocalCredential,
	session WebSession,
) error {
	if repository == nil || repository.pool == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validLocalAccountAndSessionRegistration(account, credential, session) {
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
		return postgresIdentityConflictOrUnavailable(err, errLocalIdentityIdentifierConflict)
	}
	if _, err := transaction.Exec(identityContext(ctx), `INSERT INTO local_credentials
        (credential_id, user_id, schema_version, algorithm, policy_version, iterations, key_length, salt,
         derived_key, lifecycle_state, record_version, created_at, updated_at, audit_ref)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`, credential.CredentialID,
		credential.UserID, credential.SchemaVersion, credential.Algorithm, credential.PolicyVersion,
		credential.Iterations, credential.KeyLength, credential.salt, credential.derivedKey,
		credential.LifecycleState, credential.RecordVersion, credential.CreatedAt, credential.UpdatedAt,
		credential.AuditRef); err != nil {
		return postgresIdentityConflictOrUnavailable(err, errLocalIdentityIdentifierConflict)
	}
	if _, err := transaction.Exec(identityContext(ctx), `INSERT INTO local_web_sessions
        (session_id, user_id, schema_version, credential_digest, authentication_method, authentication_source_ref,
         policy_version, lifecycle_state, record_version, created_at, updated_at, last_verified_at, expires_at,
         revoked_at, audit_ref) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		session.SessionID, session.UserID, session.SchemaVersion, session.credentialDigest,
		session.AuthenticationMethod, session.AuthenticationSourceRef, session.PolicyVersion, session.LifecycleState,
		session.RecordVersion, session.CreatedAt, session.UpdatedAt, session.LastVerifiedAt, session.ExpiresAt,
		session.RevokedAt, session.AuditRef); err != nil {
		return postgresIdentityConflictOrUnavailable(err, errLocalIdentityIdentifierConflict)
	}
	if err := transaction.Commit(identityContext(ctx)); err != nil {
		return errLocalIdentityStoreUnavailable
	}
	return nil
}

func (repository *postgresLocalIdentityRepository) ReadAccount(ctx context.Context, userID string) (UserAccount, error) {
	if repository == nil || repository.pool == nil {
		return UserAccount{}, errLocalIdentityStoreUnavailable
	}
	return scanPostgresUserAccount(repository.pool.QueryRow(identityContext(ctx), postgresAccountSelect+` WHERE user_id=$1`, strings.TrimSpace(userID)))
}

func (repository *postgresLocalIdentityRepository) FindAccountByLoginIdentifier(ctx context.Context, rawIdentifier string) (UserAccount, error) {
	identifier, err := NormalizeLocalLoginIdentifier(rawIdentifier)
	if err != nil {
		return UserAccount{}, err
	}
	if repository == nil || repository.pool == nil {
		return UserAccount{}, errLocalIdentityStoreUnavailable
	}
	return scanPostgresUserAccount(repository.pool.QueryRow(identityContext(ctx), postgresAccountSelect+`
        WHERE normalized_login_identifier=$1`, identifier))
}

func (repository *postgresLocalIdentityRepository) DisableAccount(ctx context.Context, userID string, expectedVersion int, disabledAt time.Time, auditRef string) (UserAccount, error) {
	if repository == nil || repository.pool == nil {
		return UserAccount{}, errLocalIdentityStoreUnavailable
	}
	disabledAt = disabledAt.UTC()
	if !localUserIDPattern.MatchString(userID) || expectedVersion < 1 || disabledAt.IsZero() || !validAuditRef(auditRef) {
		return UserAccount{}, errLocalIdentityContractMismatch
	}
	transaction, err := repository.pool.Begin(identityContext(ctx))
	if err != nil {
		return UserAccount{}, errLocalIdentityStoreUnavailable
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	account, err := scanPostgresUserAccount(transaction.QueryRow(identityContext(ctx), `UPDATE local_user_accounts SET
        lifecycle_state='disabled', record_version=record_version+1, updated_at=$1, disabled_at=$1, audit_ref=$2
        WHERE user_id=$3 AND record_version=$4 AND lifecycle_state='active' RETURNING
        schema_version, user_id, login_identifier, normalized_login_identifier, display_name, lifecycle_state,
        record_version, created_at, updated_at, disabled_at, audit_ref`, disabledAt, auditRef, userID, expectedVersion))
	if err != nil {
		if !errors.Is(err, errLocalIdentityNotFound) {
			return UserAccount{}, errLocalIdentityStoreUnavailable
		}
		var currentVersion int
		var currentState string
		readErr := transaction.QueryRow(identityContext(ctx), `SELECT record_version, lifecycle_state
            FROM local_user_accounts WHERE user_id=$1`, userID).Scan(&currentVersion, &currentState)
		if errors.Is(readErr, pgx.ErrNoRows) {
			return UserAccount{}, errLocalIdentityNotFound
		}
		if readErr != nil {
			return UserAccount{}, errLocalIdentityStoreUnavailable
		}
		return UserAccount{}, errLocalIdentityVersionConflict
	}
	if _, err := transaction.Exec(identityContext(ctx), `UPDATE local_web_sessions SET
        lifecycle_state='revoked', record_version=record_version+1, updated_at=$1, revoked_at=$1, audit_ref=$2
        WHERE user_id=$3 AND lifecycle_state='active'`, disabledAt, auditRef, userID); err != nil {
		return UserAccount{}, errLocalIdentityStoreUnavailable
	}
	if err := transaction.Commit(identityContext(ctx)); err != nil {
		return UserAccount{}, errLocalIdentityStoreUnavailable
	}
	return account, nil
}

func (repository *postgresLocalIdentityRepository) ReadActiveCredential(ctx context.Context, userID string) (LocalCredential, error) {
	if repository == nil || repository.pool == nil {
		return LocalCredential{}, errLocalIdentityStoreUnavailable
	}
	return scanPostgresLocalCredential(repository.pool.QueryRow(identityContext(ctx), postgresCredentialSelect+`
        WHERE user_id=$1 AND lifecycle_state='active'`, strings.TrimSpace(userID)))
}

func (repository *postgresLocalIdentityRepository) ReplaceCredential(ctx context.Context, userID string, expectedCredentialID string, expectedVersion int, replacement LocalCredential) error {
	if repository == nil || repository.pool == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validLocalCredential(replacement) || replacement.UserID != userID || replacement.LifecycleState != localIdentityStateActive ||
		!localCredentialIDPattern.MatchString(expectedCredentialID) || expectedVersion < 1 {
		return errLocalIdentityContractMismatch
	}
	transaction, err := repository.pool.Begin(identityContext(ctx))
	if err != nil {
		return errLocalIdentityStoreUnavailable
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	var accountState string
	if err := transaction.QueryRow(identityContext(ctx), `SELECT lifecycle_state FROM local_user_accounts
        WHERE user_id=$1 FOR UPDATE`, userID).Scan(&accountState); errors.Is(err, pgx.ErrNoRows) {
		return errLocalIdentityNotFound
	} else if err != nil {
		return errLocalIdentityStoreUnavailable
	} else if accountState != localIdentityStateActive {
		return errLocalIdentityAccountInactive
	}
	commandTag, err := transaction.Exec(identityContext(ctx), `UPDATE local_credentials SET lifecycle_state='superseded',
        record_version=record_version+1, updated_at=$1, audit_ref=$2
		WHERE credential_id=$3 AND user_id=$4 AND record_version=$5 AND lifecycle_state='active'
		AND created_at < $1`, replacement.CreatedAt, replacement.AuditRef, expectedCredentialID, userID, expectedVersion)
	if err != nil {
		return errLocalIdentityStoreUnavailable
	}
	if commandTag.RowsAffected() != 1 {
		return errLocalIdentityVersionConflict
	}
	if _, err := transaction.Exec(identityContext(ctx), `INSERT INTO local_credentials
        (credential_id, user_id, schema_version, algorithm, policy_version, iterations, key_length, salt,
         derived_key, lifecycle_state, record_version, created_at, updated_at, audit_ref)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`, replacement.CredentialID,
		replacement.UserID, replacement.SchemaVersion, replacement.Algorithm, replacement.PolicyVersion,
		replacement.Iterations, replacement.KeyLength, replacement.salt, replacement.derivedKey,
		replacement.LifecycleState, replacement.RecordVersion, replacement.CreatedAt, replacement.UpdatedAt,
		replacement.AuditRef); err != nil {
		return postgresIdentityConflictOrUnavailable(err, errLocalIdentityIdentifierConflict)
	}
	if err := transaction.Commit(identityContext(ctx)); err != nil {
		return errLocalIdentityStoreUnavailable
	}
	return nil
}

func (repository *postgresLocalIdentityRepository) BindExternalIdentity(ctx context.Context, binding ExternalIdentityBinding) error {
	if repository == nil || repository.pool == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validExternalIdentityBinding(binding) || binding.LifecycleState != localIdentityStateActive {
		return errLocalIdentityContractMismatch
	}
	transaction, err := repository.pool.Begin(identityContext(ctx))
	if err != nil {
		return errLocalIdentityStoreUnavailable
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	var accountState string
	if err := transaction.QueryRow(identityContext(ctx), `SELECT lifecycle_state FROM local_user_accounts
		WHERE user_id=$1 FOR UPDATE`, binding.UserID).Scan(&accountState); errors.Is(err, pgx.ErrNoRows) {
		return errLocalIdentityNotFound
	} else if err != nil {
		return errLocalIdentityStoreUnavailable
	} else if accountState != localIdentityStateActive {
		return errLocalIdentityAccountInactive
	}
	_, err = transaction.Exec(identityContext(ctx), `INSERT INTO external_identity_bindings
        (binding_id, user_id, schema_version, issuer, subject_value, lifecycle_state, record_version,
         created_at, updated_at, revoked_at, audit_ref) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		binding.BindingID, binding.UserID, binding.SchemaVersion, binding.Issuer, binding.Subject,
		binding.LifecycleState, binding.RecordVersion, binding.CreatedAt, binding.UpdatedAt, binding.RevokedAt,
		binding.AuditRef)
	if err != nil {
		return postgresIdentityConflictOrUnavailable(err, errLocalIdentityExternalConflict)
	}
	if err := transaction.Commit(identityContext(ctx)); err != nil {
		return errLocalIdentityStoreUnavailable
	}
	return nil
}

func (repository *postgresLocalIdentityRepository) ResolveExternalIdentity(ctx context.Context, rawIssuer string, rawSubject string) (ExternalIdentityBinding, error) {
	issuer, err := NormalizeExternalIssuer(rawIssuer)
	if err != nil {
		return ExternalIdentityBinding{}, err
	}
	subject, err := NormalizeExternalSubject(rawSubject)
	if err != nil {
		return ExternalIdentityBinding{}, err
	}
	if repository == nil || repository.pool == nil {
		return ExternalIdentityBinding{}, errLocalIdentityStoreUnavailable
	}
	return scanPostgresExternalIdentityBinding(repository.pool.QueryRow(identityContext(ctx), postgresBindingSelect+`
        WHERE issuer=$1 AND subject_value=$2 AND lifecycle_state='active'`, issuer, subject))
}

func (repository *postgresLocalIdentityRepository) RevokeExternalIdentity(ctx context.Context, bindingID string, expectedVersion int, revokedAt time.Time, auditRef string) (ExternalIdentityBinding, error) {
	if repository == nil || repository.pool == nil {
		return ExternalIdentityBinding{}, errLocalIdentityStoreUnavailable
	}
	revokedAt = revokedAt.UTC()
	if !localBindingIDPattern.MatchString(bindingID) || expectedVersion < 1 || revokedAt.IsZero() || !validAuditRef(auditRef) {
		return ExternalIdentityBinding{}, errLocalIdentityContractMismatch
	}
	transaction, err := repository.pool.Begin(identityContext(ctx))
	if err != nil {
		return ExternalIdentityBinding{}, errLocalIdentityStoreUnavailable
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	var userID, lifecycleState string
	var currentVersion int
	if err := transaction.QueryRow(identityContext(ctx), `SELECT user_id, record_version, lifecycle_state
        FROM external_identity_bindings WHERE binding_id=$1 FOR UPDATE`, bindingID).Scan(&userID, &currentVersion, &lifecycleState); errors.Is(err, pgx.ErrNoRows) {
		return ExternalIdentityBinding{}, errLocalIdentityVersionConflict
	} else if err != nil {
		return ExternalIdentityBinding{}, errLocalIdentityStoreUnavailable
	}
	if currentVersion != expectedVersion || lifecycleState != localIdentityStateActive {
		return ExternalIdentityBinding{}, errLocalIdentityVersionConflict
	}
	var lockedUserID string
	if err := transaction.QueryRow(identityContext(ctx), `SELECT user_id FROM local_user_accounts
		WHERE user_id=$1 FOR UPDATE`, userID).Scan(&lockedUserID); err != nil {
		return ExternalIdentityBinding{}, errLocalIdentityStoreUnavailable
	}
	var activeCredentialCount, activeBindingCount int
	if err := transaction.QueryRow(identityContext(ctx), `SELECT COUNT(*) FROM local_credentials
        WHERE user_id=$1 AND lifecycle_state='active'`, userID).Scan(&activeCredentialCount); err != nil {
		return ExternalIdentityBinding{}, errLocalIdentityStoreUnavailable
	}
	if err := transaction.QueryRow(identityContext(ctx), `SELECT COUNT(*) FROM external_identity_bindings
        WHERE user_id=$1 AND lifecycle_state='active'`, userID).Scan(&activeBindingCount); err != nil {
		return ExternalIdentityBinding{}, errLocalIdentityStoreUnavailable
	}
	if activeCredentialCount == 0 && activeBindingCount <= 1 {
		return ExternalIdentityBinding{}, errLocalIdentityLastLoginMethodRemoval
	}
	binding, err := scanPostgresExternalIdentityBinding(transaction.QueryRow(identityContext(ctx), `UPDATE external_identity_bindings SET
        lifecycle_state='revoked', record_version=record_version+1, updated_at=$1, revoked_at=$1, audit_ref=$2
        WHERE binding_id=$3 AND record_version=$4 AND lifecycle_state='active' RETURNING
        schema_version, binding_id, user_id, issuer, subject_value, lifecycle_state, record_version,
        created_at, updated_at, revoked_at, audit_ref`, revokedAt, auditRef, bindingID, expectedVersion))
	if err != nil {
		return ExternalIdentityBinding{}, classifyPostgresMutationFailure(err)
	}
	if err := transaction.Commit(identityContext(ctx)); err != nil {
		return ExternalIdentityBinding{}, errLocalIdentityStoreUnavailable
	}
	return binding, nil
}

func (repository *postgresLocalIdentityRepository) CreateWebSession(ctx context.Context, session WebSession) error {
	if repository == nil || repository.pool == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validWebSession(session) || session.LifecycleState != localIdentityStateActive {
		return errLocalIdentityContractMismatch
	}
	transaction, err := repository.pool.Begin(identityContext(ctx))
	if err != nil {
		return errLocalIdentityStoreUnavailable
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	var accountState string
	if err := transaction.QueryRow(identityContext(ctx), `SELECT lifecycle_state FROM local_user_accounts
        WHERE user_id=$1 FOR UPDATE`, session.UserID).Scan(&accountState); errors.Is(err, pgx.ErrNoRows) {
		return errLocalIdentityNotFound
	} else if err != nil {
		return errLocalIdentityStoreUnavailable
	} else if accountState != localIdentityStateActive {
		return errLocalIdentityAccountInactive
	}
	_, err = transaction.Exec(identityContext(ctx), `INSERT INTO local_web_sessions
        (session_id, user_id, schema_version, credential_digest, authentication_method, authentication_source_ref,
         policy_version, lifecycle_state, record_version, created_at, updated_at, last_verified_at, expires_at,
         revoked_at, audit_ref) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		session.SessionID, session.UserID, session.SchemaVersion, session.credentialDigest,
		session.AuthenticationMethod, session.AuthenticationSourceRef, session.PolicyVersion, session.LifecycleState,
		session.RecordVersion, session.CreatedAt, session.UpdatedAt, session.LastVerifiedAt, session.ExpiresAt,
		session.RevokedAt, session.AuditRef)
	if err != nil {
		return postgresIdentityConflictOrUnavailable(err, errLocalIdentityIdentifierConflict)
	}
	if err := transaction.Commit(identityContext(ctx)); err != nil {
		return errLocalIdentityStoreUnavailable
	}
	return nil
}

func (repository *postgresLocalIdentityRepository) ResolveWebSession(ctx context.Context, digest [sha256.Size]byte, now time.Time) (WebSession, UserAccount, error) {
	if repository == nil || repository.pool == nil {
		return WebSession{}, UserAccount{}, errLocalIdentityStoreUnavailable
	}
	session, err := scanPostgresWebSession(repository.pool.QueryRow(identityContext(ctx), postgresSessionSelect+`
        WHERE credential_digest=$1`, digest[:]))
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

func (repository *postgresLocalIdentityRepository) ReadWebSession(ctx context.Context, sessionID string, now time.Time) (WebSession, UserAccount, error) {
	if repository == nil || repository.pool == nil {
		return WebSession{}, UserAccount{}, errLocalIdentityStoreUnavailable
	}
	now = now.UTC()
	session, err := scanPostgresWebSession(repository.pool.QueryRow(identityContext(ctx), postgresSessionSelect+`
		WHERE session_id=$1`, strings.TrimSpace(sessionID)))
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

func (repository *postgresLocalIdentityRepository) RevokeWebSession(ctx context.Context, sessionID string, expectedVersion int, revokedAt time.Time, auditRef string) (WebSession, error) {
	if repository == nil || repository.pool == nil {
		return WebSession{}, errLocalIdentityStoreUnavailable
	}
	revokedAt = revokedAt.UTC()
	if !localSessionIDPattern.MatchString(sessionID) || expectedVersion < 1 || revokedAt.IsZero() || !validAuditRef(auditRef) {
		return WebSession{}, errLocalIdentityContractMismatch
	}
	session, err := scanPostgresWebSession(repository.pool.QueryRow(identityContext(ctx), `UPDATE local_web_sessions SET
        lifecycle_state='revoked', record_version=record_version+1, updated_at=$1, revoked_at=$1, audit_ref=$2
        WHERE session_id=$3 AND record_version=$4 AND lifecycle_state='active' RETURNING
        schema_version, session_id, user_id, credential_digest, authentication_method, authentication_source_ref,
        policy_version, lifecycle_state, record_version, created_at, updated_at, last_verified_at, expires_at,
        revoked_at, audit_ref`, revokedAt, auditRef, sessionID, expectedVersion))
	if err != nil {
		return WebSession{}, classifyPostgresMutationFailure(err)
	}
	return session, nil
}

func (repository *postgresLocalIdentityRepository) CreateRoleAssignment(ctx context.Context, assignment LocalRoleAssignment) error {
	grants, ok := normalizedPermissionGrants(assignment.PermissionGrants)
	assignment.PermissionGrants = grants
	if repository == nil || repository.pool == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !ok || localIdentityContainsManagementPermission(grants) ||
		!validLocalRoleAssignment(assignment) || assignment.LifecycleState != localIdentityStateActive {
		return errLocalIdentityContractMismatch
	}
	if err := repository.requireActivePostgresAccount(ctx, assignment.UserID); err != nil {
		return err
	}
	_, err := repository.pool.Exec(identityContext(ctx), `INSERT INTO local_role_assignments
        (assignment_id, user_id, schema_version, tenant_ref, workspace_id, role_key, permission_grants,
         lifecycle_state, record_version, created_at, updated_at, expires_at, revoked_at, audit_ref)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`, assignment.AssignmentID,
		assignment.UserID, assignment.SchemaVersion, assignment.TenantRef, assignment.WorkspaceID, assignment.RoleKey,
		assignment.PermissionGrants, assignment.LifecycleState, assignment.RecordVersion, assignment.CreatedAt,
		assignment.UpdatedAt, assignment.ExpiresAt, assignment.RevokedAt, assignment.AuditRef)
	if err != nil {
		return postgresIdentityConflictOrUnavailable(err, errLocalIdentityIdentifierConflict)
	}
	return nil
}

func (repository *postgresLocalIdentityRepository) RevokeRoleAssignment(ctx context.Context, assignmentID string, expectedVersion int, revokedAt time.Time, auditRef string) (LocalRoleAssignment, error) {
	if repository == nil || repository.pool == nil {
		return LocalRoleAssignment{}, errLocalIdentityStoreUnavailable
	}
	revokedAt = revokedAt.UTC()
	if !localRoleAssignmentIDPattern.MatchString(assignmentID) || expectedVersion < 1 || revokedAt.IsZero() || !validAuditRef(auditRef) {
		return LocalRoleAssignment{}, errLocalIdentityContractMismatch
	}
	assignment, err := scanPostgresLocalRoleAssignment(repository.pool.QueryRow(identityContext(ctx), `UPDATE local_role_assignments SET
        lifecycle_state='revoked', record_version=record_version+1, updated_at=$1, revoked_at=$1, audit_ref=$2
        WHERE assignment_id=$3 AND record_version=$4 AND lifecycle_state='active' RETURNING
        schema_version, assignment_id, user_id, tenant_ref, workspace_id, role_key, permission_grants,
        lifecycle_state, record_version, created_at, updated_at, expires_at, revoked_at, audit_ref`,
		revokedAt, auditRef, assignmentID, expectedVersion))
	if err != nil {
		return LocalRoleAssignment{}, classifyPostgresMutationFailure(err)
	}
	return assignment, nil
}

func (repository *postgresLocalIdentityRepository) CreateWorkspaceMembership(ctx context.Context, membership WorkspaceMembership) error {
	if repository == nil || repository.pool == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validWorkspaceMembership(membership) || membership.LifecycleState != localIdentityStateActive {
		return errLocalIdentityContractMismatch
	}
	if err := repository.requireActivePostgresAccount(ctx, membership.UserID); err != nil {
		return err
	}
	_, err := repository.pool.Exec(identityContext(ctx), `INSERT INTO local_workspace_memberships
        (membership_id, user_id, schema_version, tenant_ref, workspace_id, lifecycle_state, record_version,
         created_at, updated_at, expires_at, revoked_at, audit_ref)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, membership.MembershipID, membership.UserID,
		membership.SchemaVersion, membership.TenantRef, membership.WorkspaceID, membership.LifecycleState,
		membership.RecordVersion, membership.CreatedAt, membership.UpdatedAt, membership.ExpiresAt,
		membership.RevokedAt, membership.AuditRef)
	if err != nil {
		return postgresIdentityConflictOrUnavailable(err, errLocalIdentityIdentifierConflict)
	}
	return nil
}

func (repository *postgresLocalIdentityRepository) RevokeWorkspaceMembership(ctx context.Context, membershipID string, expectedVersion int, revokedAt time.Time, auditRef string) (WorkspaceMembership, error) {
	if repository == nil || repository.pool == nil {
		return WorkspaceMembership{}, errLocalIdentityStoreUnavailable
	}
	revokedAt = revokedAt.UTC()
	if !localMembershipIDPattern.MatchString(membershipID) || expectedVersion < 1 || revokedAt.IsZero() || !validAuditRef(auditRef) {
		return WorkspaceMembership{}, errLocalIdentityContractMismatch
	}
	membership, err := scanPostgresWorkspaceMembership(repository.pool.QueryRow(identityContext(ctx), `UPDATE local_workspace_memberships SET
        lifecycle_state='revoked', record_version=record_version+1, updated_at=$1, revoked_at=$1, audit_ref=$2
        WHERE membership_id=$3 AND record_version=$4 AND lifecycle_state='active' RETURNING
        schema_version, membership_id, user_id, tenant_ref, workspace_id, lifecycle_state, record_version,
        created_at, updated_at, expires_at, revoked_at, audit_ref`, revokedAt, auditRef, membershipID, expectedVersion))
	if err != nil {
		return WorkspaceMembership{}, classifyPostgresMutationFailure(err)
	}
	return membership, nil
}

func (repository *postgresLocalIdentityRepository) AuthorizeWorkspace(ctx context.Context, userID string, tenantRef string, workspaceID string, required []string, now time.Time) (LocalWorkspaceAuthorization, error) {
	required, ok := normalizedPermissionGrants(required)
	if repository == nil || repository.pool == nil {
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
	membership, err := scanPostgresWorkspaceMembership(repository.pool.QueryRow(identityContext(ctx), postgresMembershipSelect+`
        WHERE user_id=$1 AND tenant_ref=$2 AND workspace_id=$3 AND lifecycle_state='active'`, userID, tenantRef, workspaceID))
	if err != nil {
		if errors.Is(err, errLocalIdentityNotFound) {
			return LocalWorkspaceAuthorization{}, errLocalIdentityMembershipDenied
		}
		return LocalWorkspaceAuthorization{}, err
	}
	if membership.ExpiresAt != nil && !membership.ExpiresAt.After(now.UTC()) {
		return LocalWorkspaceAuthorization{}, errLocalIdentityMembershipDenied
	}
	rows, err := repository.pool.Query(identityContext(ctx), postgresRoleAssignmentSelect+`
        WHERE user_id=$1 AND tenant_ref=$2 AND (workspace_id='' OR workspace_id=$3) AND lifecycle_state='active'
        ORDER BY assignment_id`, userID, tenantRef, workspaceID)
	if err != nil {
		return LocalWorkspaceAuthorization{}, errLocalIdentityStoreUnavailable
	}
	defer rows.Close()
	assignments := make([]LocalRoleAssignment, 0)
	grantSet := make(map[string]struct{})
	for rows.Next() {
		assignment, scanErr := scanPostgresLocalRoleAssignment(rows)
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
	slices.SortFunc(assignments, func(left, right LocalRoleAssignment) int {
		return strings.Compare(left.AssignmentID, right.AssignmentID)
	})
	return buildLocalWorkspaceAuthorization(account, membership, assignments, grantSet, required)
}

const postgresAccountSelect = `SELECT schema_version, user_id, login_identifier, normalized_login_identifier,
    display_name, lifecycle_state, record_version, created_at, updated_at, disabled_at, audit_ref
    FROM local_user_accounts`

const postgresCredentialSelect = `SELECT schema_version, credential_id, user_id, algorithm, policy_version,
    iterations, key_length, salt, derived_key, lifecycle_state, record_version, created_at, updated_at, audit_ref
    FROM local_credentials`

const postgresBindingSelect = `SELECT schema_version, binding_id, user_id, issuer, subject_value, lifecycle_state,
    record_version, created_at, updated_at, revoked_at, audit_ref FROM external_identity_bindings`

const postgresSessionSelect = `SELECT schema_version, session_id, user_id, credential_digest, authentication_method,
    authentication_source_ref, policy_version, lifecycle_state, record_version, created_at, updated_at,
    last_verified_at, expires_at, revoked_at, audit_ref FROM local_web_sessions`

const postgresRoleAssignmentSelect = `SELECT schema_version, assignment_id, user_id, tenant_ref, workspace_id,
    role_key, permission_grants, lifecycle_state, record_version, created_at, updated_at, expires_at,
    revoked_at, audit_ref FROM local_role_assignments`

const postgresMembershipSelect = `SELECT schema_version, membership_id, user_id, tenant_ref, workspace_id,
    lifecycle_state, record_version, created_at, updated_at, expires_at, revoked_at, audit_ref
    FROM local_workspace_memberships`

func scanPostgresUserAccount(row localIdentitySQLRow) (UserAccount, error) {
	var account UserAccount
	err := row.Scan(&account.SchemaVersion, &account.UserID, &account.LoginIdentifier, &account.NormalizedLoginIdentifier,
		&account.DisplayName, &account.LifecycleState, &account.RecordVersion, &account.CreatedAt, &account.UpdatedAt,
		&account.DisabledAt, &account.AuditRef)
	if err != nil {
		return UserAccount{}, normalizePostgresReadError(err)
	}
	account.CreatedAt = account.CreatedAt.UTC()
	account.UpdatedAt = account.UpdatedAt.UTC()
	account.DisabledAt = cloneTimePointer(account.DisabledAt)
	if !validUserAccount(account) {
		return UserAccount{}, errLocalIdentityStoreUnavailable
	}
	return account, nil
}

func scanPostgresLocalCredential(row localIdentitySQLRow) (LocalCredential, error) {
	var credential LocalCredential
	err := row.Scan(&credential.SchemaVersion, &credential.CredentialID, &credential.UserID, &credential.Algorithm,
		&credential.PolicyVersion, &credential.Iterations, &credential.KeyLength, &credential.salt,
		&credential.derivedKey, &credential.LifecycleState, &credential.RecordVersion, &credential.CreatedAt,
		&credential.UpdatedAt, &credential.AuditRef)
	if err != nil {
		return LocalCredential{}, normalizePostgresReadError(err)
	}
	credential.CreatedAt = credential.CreatedAt.UTC()
	credential.UpdatedAt = credential.UpdatedAt.UTC()
	if !validLocalCredential(credential) {
		return LocalCredential{}, errLocalIdentityStoreUnavailable
	}
	return credential, nil
}

func scanPostgresExternalIdentityBinding(row localIdentitySQLRow) (ExternalIdentityBinding, error) {
	var binding ExternalIdentityBinding
	err := row.Scan(&binding.SchemaVersion, &binding.BindingID, &binding.UserID, &binding.Issuer, &binding.Subject,
		&binding.LifecycleState, &binding.RecordVersion, &binding.CreatedAt, &binding.UpdatedAt, &binding.RevokedAt,
		&binding.AuditRef)
	if err != nil {
		return ExternalIdentityBinding{}, normalizePostgresReadError(err)
	}
	binding.CreatedAt = binding.CreatedAt.UTC()
	binding.UpdatedAt = binding.UpdatedAt.UTC()
	binding.RevokedAt = cloneTimePointer(binding.RevokedAt)
	if !validExternalIdentityBinding(binding) {
		return ExternalIdentityBinding{}, errLocalIdentityStoreUnavailable
	}
	return binding, nil
}

func scanPostgresWebSession(row localIdentitySQLRow) (WebSession, error) {
	var session WebSession
	err := row.Scan(&session.SchemaVersion, &session.SessionID, &session.UserID, &session.credentialDigest,
		&session.AuthenticationMethod, &session.AuthenticationSourceRef, &session.PolicyVersion, &session.LifecycleState,
		&session.RecordVersion, &session.CreatedAt, &session.UpdatedAt, &session.LastVerifiedAt, &session.ExpiresAt,
		&session.RevokedAt, &session.AuditRef)
	if err != nil {
		return WebSession{}, normalizePostgresReadError(err)
	}
	session.CreatedAt = session.CreatedAt.UTC()
	session.UpdatedAt = session.UpdatedAt.UTC()
	session.LastVerifiedAt = session.LastVerifiedAt.UTC()
	session.ExpiresAt = session.ExpiresAt.UTC()
	session.RevokedAt = cloneTimePointer(session.RevokedAt)
	if !validWebSession(session) {
		return WebSession{}, errLocalIdentityStoreUnavailable
	}
	return session, nil
}

func scanPostgresLocalRoleAssignment(row localIdentitySQLRow) (LocalRoleAssignment, error) {
	var assignment LocalRoleAssignment
	err := row.Scan(&assignment.SchemaVersion, &assignment.AssignmentID, &assignment.UserID, &assignment.TenantRef,
		&assignment.WorkspaceID, &assignment.RoleKey, &assignment.PermissionGrants, &assignment.LifecycleState,
		&assignment.RecordVersion, &assignment.CreatedAt, &assignment.UpdatedAt, &assignment.ExpiresAt,
		&assignment.RevokedAt, &assignment.AuditRef)
	if err != nil {
		return LocalRoleAssignment{}, normalizePostgresReadError(err)
	}
	assignment.CreatedAt = assignment.CreatedAt.UTC()
	assignment.UpdatedAt = assignment.UpdatedAt.UTC()
	assignment.ExpiresAt = cloneTimePointer(assignment.ExpiresAt)
	assignment.RevokedAt = cloneTimePointer(assignment.RevokedAt)
	if !validLocalRoleAssignment(assignment) {
		return LocalRoleAssignment{}, errLocalIdentityStoreUnavailable
	}
	return assignment, nil
}

func scanPostgresWorkspaceMembership(row localIdentitySQLRow) (WorkspaceMembership, error) {
	var membership WorkspaceMembership
	err := row.Scan(&membership.SchemaVersion, &membership.MembershipID, &membership.UserID, &membership.TenantRef,
		&membership.WorkspaceID, &membership.LifecycleState, &membership.RecordVersion, &membership.CreatedAt,
		&membership.UpdatedAt, &membership.ExpiresAt, &membership.RevokedAt, &membership.AuditRef)
	if err != nil {
		return WorkspaceMembership{}, normalizePostgresReadError(err)
	}
	membership.CreatedAt = membership.CreatedAt.UTC()
	membership.UpdatedAt = membership.UpdatedAt.UTC()
	membership.ExpiresAt = cloneTimePointer(membership.ExpiresAt)
	membership.RevokedAt = cloneTimePointer(membership.RevokedAt)
	if !validWorkspaceMembership(membership) {
		return WorkspaceMembership{}, errLocalIdentityStoreUnavailable
	}
	return membership, nil
}

func (repository *postgresLocalIdentityRepository) requireActivePostgresAccount(ctx context.Context, userID string) error {
	account, err := repository.ReadAccount(ctx, userID)
	if err != nil {
		return err
	}
	if account.LifecycleState != localIdentityStateActive {
		return errLocalIdentityAccountInactive
	}
	return nil
}

func postgresIdentityConflictOrUnavailable(err error, conflict error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23505":
			return conflict
		case "23503":
			return errLocalIdentityNotFound
		}
	}
	return errLocalIdentityStoreUnavailable
}

func classifyPostgresMutationFailure(err error) error {
	if errors.Is(err, errLocalIdentityNotFound) {
		return errLocalIdentityVersionConflict
	}
	return errLocalIdentityStoreUnavailable
}

func normalizePostgresReadError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return errLocalIdentityNotFound
	}
	return errLocalIdentityStoreUnavailable
}
