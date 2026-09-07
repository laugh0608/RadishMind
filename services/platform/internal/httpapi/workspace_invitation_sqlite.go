package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

const sqliteWorkspaceInvitationSelect = `SELECT
    schema_version, invitation_id, record_version, tenant_ref, workspace_id, role_key,
    role_catalog_version, role_definition_digest, ttl_policy, lifecycle_state, secret_digest,
    expires_at_unix_nano, created_at_unix_nano, updated_at_unix_nano, created_by_actor_ref,
    created_request_ref, created_audit_ref, updated_request_ref, updated_audit_ref,
    claimed_at_unix_nano, claimed_by_user_id, membership_id, assignment_id,
    revoked_at_unix_nano, revoked_by_actor_ref
    FROM local_workspace_invitations`

func (repository *sqliteLocalIdentityRepository) CreateWorkspaceInvitation(
	ctx context.Context,
	actorUserID string,
	invitation WorkspaceInvitation,
	secretDigest workspaceInvitationSecretDigest,
	now time.Time,
) error {
	if repository == nil || repository.database == nil {
		return errWorkspaceInvitationAdminUnavailable
	}
	actorUserID = strings.TrimSpace(actorUserID)
	now = now.UTC()
	if !localUserIDPattern.MatchString(actorUserID) || !validWorkspaceInvitation(invitation) ||
		!validWorkspaceInvitationSecretDigest(secretDigest) || now.IsZero() || !invitation.CreatedAt.Equal(now) {
		return errWorkspaceInvitationAdminUnavailable
	}
	return repository.mutateLocalIdentityAdministrationScope(
		ctx,
		invitation.TenantRef,
		invitation.WorkspaceID,
		[]string{actorUserID},
		func(snapshot *memoryLocalIdentityRepository, connection *sql.Conn) error {
			if err := snapshot.CreateWorkspaceInvitation(ctx, actorUserID, invitation, secretDigest, now); err != nil {
				return err
			}
			if err := insertSQLiteWorkspaceInvitation(ctx, connection, invitation, secretDigest); err != nil {
				return errWorkspaceInvitationAdminUnavailable
			}
			return nil
		},
	)
}

func (repository *sqliteLocalIdentityRepository) ListWorkspaceInvitations(
	ctx context.Context,
	actorUserID string,
	query WorkspaceInvitationListQuery,
) (WorkspaceInvitationPage, error) {
	if repository == nil || repository.database == nil {
		return WorkspaceInvitationPage{}, errWorkspaceInvitationAdminUnavailable
	}
	filter, cursor, err := normalizeWorkspaceInvitationListQuery(query)
	if err != nil {
		return WorkspaceInvitationPage{}, err
	}
	transaction, err := repository.database.BeginTx(identityContext(ctx), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return WorkspaceInvitationPage{}, errWorkspaceInvitationAdminUnavailable
	}
	defer func() { _ = transaction.Rollback() }()
	snapshot, err := loadSQLiteLocalIdentityAdministrationScope(
		identityContext(ctx), transaction, filter.TenantRef, filter.WorkspaceID, strings.TrimSpace(actorUserID),
	)
	if err != nil {
		return WorkspaceInvitationPage{}, errWorkspaceInvitationAdminUnavailable
	}
	if err = snapshot.authorizeWorkspaceInvitationAdministratorLocked(
		strings.TrimSpace(actorUserID), filter.TenantRef, filter.WorkspaceID, filter.authorizedAt,
		localIdentityPermissionMembersRead, localIdentityPermissionRolesRead,
	); err != nil {
		return WorkspaceInvitationPage{}, err
	}
	statement := sqliteWorkspaceInvitationSelect + ` WHERE tenant_ref=? AND workspace_id=?`
	arguments := []any{filter.TenantRef, filter.WorkspaceID}
	switch filter.EffectiveState {
	case workspaceInvitationEffectivePending:
		statement += ` AND lifecycle_state='pending' AND expires_at_unix_nano>?`
		arguments = append(arguments, filter.asOf.UnixNano())
	case workspaceInvitationEffectiveExpired:
		statement += ` AND lifecycle_state='pending' AND expires_at_unix_nano<=?`
		arguments = append(arguments, filter.asOf.UnixNano())
	case workspaceInvitationEffectiveClaimed, workspaceInvitationEffectiveRevoked:
		statement += ` AND lifecycle_state=?`
		arguments = append(arguments, filter.EffectiveState)
	default:
		return WorkspaceInvitationPage{}, errWorkspaceInvitationCursorInvalid
	}
	if cursor.InvitationID != "" {
		anchor := mustParseLocalIdentityCursorTime(cursor.UpdatedAt).UnixNano()
		statement += ` AND (updated_at_unix_nano<? OR (updated_at_unix_nano=? AND invitation_id<?))`
		arguments = append(arguments, anchor, anchor, cursor.InvitationID)
	}
	statement += ` ORDER BY updated_at_unix_nano DESC, invitation_id DESC LIMIT ?`
	arguments = append(arguments, filter.Limit+1)
	rows, err := transaction.QueryContext(identityContext(ctx), statement, arguments...)
	if err != nil {
		return WorkspaceInvitationPage{}, errWorkspaceInvitationAdminUnavailable
	}
	stored := make([]memoryWorkspaceInvitation, 0, filter.Limit+1)
	for rows.Next() {
		item, scanErr := scanSQLiteWorkspaceInvitation(rows)
		if scanErr != nil {
			rows.Close()
			return WorkspaceInvitationPage{}, errWorkspaceInvitationAdminUnavailable
		}
		stored = append(stored, item)
	}
	if rows.Close() != nil || rows.Err() != nil {
		return WorkspaceInvitationPage{}, errWorkspaceInvitationAdminUnavailable
	}
	hasNext := len(stored) > filter.Limit
	if hasNext {
		stored = stored[:filter.Limit]
	}
	page := WorkspaceInvitationPage{
		SchemaVersion: workspaceInvitationPageSchemaVersion,
		AsOf:          filter.asOf,
		Invitations:   make([]WorkspaceInvitation, 0, len(stored)),
	}
	for _, item := range stored {
		page.Invitations = append(page.Invitations, projectWorkspaceInvitation(item.invitation, filter.asOf))
	}
	if hasNext && len(stored) > 0 {
		page.NextCursor, err = encodeWorkspaceInvitationCursor(filter, stored[len(stored)-1].invitation)
		if err != nil {
			return WorkspaceInvitationPage{}, errWorkspaceInvitationAdminUnavailable
		}
	}
	return page, nil
}

func (repository *sqliteLocalIdentityRepository) RevokeWorkspaceInvitation(
	ctx context.Context,
	actorUserID string,
	input WorkspaceInvitationRevokeInput,
	now time.Time,
) (WorkspaceInvitation, error) {
	if repository == nil || repository.database == nil {
		return WorkspaceInvitation{}, errWorkspaceInvitationAdminUnavailable
	}
	actorUserID = strings.TrimSpace(actorUserID)
	now = now.UTC()
	if !localUserIDPattern.MatchString(actorUserID) || now.IsZero() {
		return WorkspaceInvitation{}, errWorkspaceInvitationTransitionInvalid
	}
	var revoked WorkspaceInvitation
	err := repository.mutateLocalIdentityAdministrationScope(
		ctx,
		strings.TrimSpace(input.TenantRef),
		strings.TrimSpace(input.WorkspaceID),
		[]string{actorUserID},
		func(snapshot *memoryLocalIdentityRepository, connection *sql.Conn) error {
			stored, readErr := scanSQLiteWorkspaceInvitation(connection.QueryRowContext(
				identityContext(ctx), sqliteWorkspaceInvitationSelect+` WHERE invitation_id=?`, strings.TrimSpace(input.InvitationID),
			))
			if errors.Is(readErr, errLocalIdentityNotFound) {
				return errWorkspaceInvitationTransitionInvalid
			}
			if readErr != nil {
				return errWorkspaceInvitationAdminUnavailable
			}
			snapshot.workspaceInvitations[stored.invitation.InvitationID] = stored
			revoked, readErr = snapshot.RevokeWorkspaceInvitation(ctx, actorUserID, input, now)
			if readErr != nil {
				return readErr
			}
			result, updateErr := connection.ExecContext(identityContext(ctx), `UPDATE local_workspace_invitations SET
                record_version=?, lifecycle_state='revoked', updated_at_unix_nano=?, updated_request_ref=?,
                updated_audit_ref=?, revoked_at_unix_nano=?, revoked_by_actor_ref=?
                WHERE invitation_id=? AND record_version=? AND lifecycle_state='pending'`,
				revoked.RecordVersion, revoked.UpdatedAt.UnixNano(), revoked.UpdatedRequestRef, revoked.UpdatedAuditRef,
				revoked.RevokedAt.UnixNano(), revoked.RevokedByActorRef, revoked.InvitationID, input.ExpectedVersion,
			)
			if updateErr != nil {
				return errWorkspaceInvitationAdminUnavailable
			}
			affected, affectedErr := result.RowsAffected()
			if affectedErr != nil || affected != 1 {
				return errWorkspaceInvitationAdminUnavailable
			}
			return nil
		},
	)
	if err != nil {
		return WorkspaceInvitation{}, err
	}
	return cloneWorkspaceInvitation(revoked), nil
}

func (repository *sqliteLocalIdentityRepository) PreviewWorkspaceInvitation(
	ctx context.Context,
	claimantUserID string,
	claimantTenantRef string,
	invitationID string,
	secretDigest workspaceInvitationSecretDigest,
	now time.Time,
) (WorkspaceInvitation, error) {
	if repository == nil || repository.database == nil {
		return WorkspaceInvitation{}, errWorkspaceInvitationStoreUnavailable
	}
	transaction, err := repository.database.BeginTx(identityContext(ctx), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return WorkspaceInvitation{}, errWorkspaceInvitationStoreUnavailable
	}
	defer func() { _ = transaction.Rollback() }()
	stored, err := scanSQLiteWorkspaceInvitation(transaction.QueryRowContext(
		identityContext(ctx), sqliteWorkspaceInvitationSelect+` WHERE invitation_id=?`, strings.TrimSpace(invitationID),
	))
	if errors.Is(err, errLocalIdentityNotFound) {
		_ = workspaceInvitationSecretMatches(secretDigest, workspaceInvitationDummySecretDigest)
		return WorkspaceInvitation{}, errWorkspaceInvitationInvalid
	}
	if err != nil {
		return WorkspaceInvitation{}, errWorkspaceInvitationStoreUnavailable
	}
	accounts, err := readSQLiteLocalIdentityAccounts(identityContext(ctx), transaction, []string{strings.TrimSpace(claimantUserID)})
	if err != nil {
		return WorkspaceInvitation{}, errWorkspaceInvitationStoreUnavailable
	}
	snapshot, err := newLocalIdentityAdministrationScopeSnapshot(accounts, nil, nil)
	if err != nil {
		return WorkspaceInvitation{}, errWorkspaceInvitationStoreUnavailable
	}
	snapshot.workspaceInvitations[stored.invitation.InvitationID] = stored
	return snapshot.PreviewWorkspaceInvitation(
		ctx, claimantUserID, claimantTenantRef, invitationID, secretDigest, now,
	)
}

func (repository *sqliteLocalIdentityRepository) ClaimWorkspaceInvitation(
	ctx context.Context,
	claimantUserID string,
	claimantTenantRef string,
	invitationID string,
	secretDigest workspaceInvitationSecretDigest,
	expectedVersion int,
	membershipID string,
	assignmentID string,
	now time.Time,
	requestRef string,
	auditRef string,
) (WorkspaceInvitation, WorkspaceMembership, LocalRoleAssignment, error) {
	if repository == nil || repository.database == nil {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationStoreUnavailable
	}
	connection, err := repository.database.Conn(identityContext(ctx))
	if err != nil {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationStoreUnavailable
	}
	defer connection.Close()
	if _, err = connection.ExecContext(identityContext(ctx), "BEGIN IMMEDIATE"); err != nil {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationStoreUnavailable
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = connection.ExecContext(context.Background(), "ROLLBACK")
		}
	}()
	stored, err := scanSQLiteWorkspaceInvitation(connection.QueryRowContext(
		identityContext(ctx), sqliteWorkspaceInvitationSelect+` WHERE invitation_id=?`, strings.TrimSpace(invitationID),
	))
	if errors.Is(err, errLocalIdentityNotFound) {
		_ = workspaceInvitationSecretMatches(secretDigest, workspaceInvitationDummySecretDigest)
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationInvalid
	}
	if err != nil {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationStoreUnavailable
	}
	snapshot, err := loadSQLiteLocalIdentityAdministrationScope(
		identityContext(ctx), connection, stored.invitation.TenantRef, stored.invitation.WorkspaceID,
		strings.TrimSpace(claimantUserID),
	)
	if err != nil {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationStoreUnavailable
	}
	snapshot.workspaceInvitations[stored.invitation.InvitationID] = stored
	invitation, membership, assignment, err := snapshot.ClaimWorkspaceInvitation(
		ctx, claimantUserID, claimantTenantRef, invitationID, secretDigest, expectedVersion,
		membershipID, assignmentID, now, requestRef, auditRef,
	)
	if err != nil {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, err
	}
	if err = insertSQLiteAdministrationMembership(ctx, connection, membership, errWorkspaceInvitationMembershipConflict); err != nil {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, normalizeWorkspaceInvitationDurableClaimWriteError(err)
	}
	if err = insertSQLiteAdministrationRoleAssignment(ctx, connection, assignment, errWorkspaceInvitationMembershipConflict); err != nil {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, normalizeWorkspaceInvitationDurableClaimWriteError(err)
	}
	result, err := connection.ExecContext(identityContext(ctx), `UPDATE local_workspace_invitations SET
        record_version=?, lifecycle_state='claimed', updated_at_unix_nano=?, updated_request_ref=?,
        updated_audit_ref=?, claimed_at_unix_nano=?, claimed_by_user_id=?, membership_id=?, assignment_id=?
        WHERE invitation_id=? AND record_version=? AND lifecycle_state='pending' AND expires_at_unix_nano>?`,
		invitation.RecordVersion, invitation.UpdatedAt.UnixNano(), invitation.UpdatedRequestRef,
		invitation.UpdatedAuditRef, invitation.ClaimedAt.UnixNano(), invitation.ClaimedByUserID,
		invitation.MembershipID, invitation.AssignmentID, invitation.InvitationID, expectedVersion, now.UTC().UnixNano(),
	)
	if err != nil {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationStoreUnavailable
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr != nil || affected != 1 {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationStoreUnavailable
	}
	if _, err = connection.ExecContext(identityContext(ctx), "COMMIT"); err != nil {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationStoreUnavailable
	}
	committed = true
	return invitation, membership, assignment, nil
}

func insertSQLiteWorkspaceInvitation(
	ctx context.Context,
	connection *sql.Conn,
	invitation WorkspaceInvitation,
	secretDigest workspaceInvitationSecretDigest,
) error {
	_, err := connection.ExecContext(identityContext(ctx), `INSERT INTO local_workspace_invitations
        (invitation_id, schema_version, record_version, tenant_ref, workspace_id, role_key,
         role_catalog_version, role_definition_digest, ttl_policy, lifecycle_state, secret_digest,
         expires_at_unix_nano, created_at_unix_nano, updated_at_unix_nano, created_by_actor_ref,
         created_request_ref, created_audit_ref, updated_request_ref, updated_audit_ref,
         claimed_at_unix_nano, claimed_by_user_id, membership_id, assignment_id,
         revoked_at_unix_nano, revoked_by_actor_ref)
        VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		invitation.InvitationID, invitation.SchemaVersion, invitation.RecordVersion, invitation.TenantRef,
		invitation.WorkspaceID, invitation.RoleKey, invitation.RoleCatalogVersion, invitation.RoleDefinitionDigest,
		invitation.TTLPolicy, invitation.LifecycleState, secretDigest[:], invitation.ExpiresAt.UnixNano(),
		invitation.CreatedAt.UnixNano(), invitation.UpdatedAt.UnixNano(), invitation.CreatedByActorRef,
		invitation.CreatedRequestRef, invitation.CreatedAuditRef, invitation.UpdatedRequestRef,
		invitation.UpdatedAuditRef, nil, nil, nil, nil, nil, nil,
	)
	return err
}

func scanSQLiteWorkspaceInvitation(row localIdentitySQLRow) (memoryWorkspaceInvitation, error) {
	var invitation WorkspaceInvitation
	var digest []byte
	var expiresAt, createdAt, updatedAt int64
	var claimedAt, revokedAt sql.NullInt64
	var claimedByUserID, membershipID, assignmentID, revokedByActorRef sql.NullString
	err := row.Scan(
		&invitation.SchemaVersion, &invitation.InvitationID, &invitation.RecordVersion,
		&invitation.TenantRef, &invitation.WorkspaceID, &invitation.RoleKey,
		&invitation.RoleCatalogVersion, &invitation.RoleDefinitionDigest, &invitation.TTLPolicy,
		&invitation.LifecycleState, &digest, &expiresAt, &createdAt, &updatedAt,
		&invitation.CreatedByActorRef, &invitation.CreatedRequestRef, &invitation.CreatedAuditRef,
		&invitation.UpdatedRequestRef, &invitation.UpdatedAuditRef, &claimedAt, &claimedByUserID,
		&membershipID, &assignmentID, &revokedAt, &revokedByActorRef,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return memoryWorkspaceInvitation{}, errLocalIdentityNotFound
	}
	if err != nil || len(digest) != len(workspaceInvitationSecretDigest{}) {
		return memoryWorkspaceInvitation{}, errWorkspaceInvitationStoreUnavailable
	}
	invitation.ExpiresAt = time.Unix(0, expiresAt).UTC()
	invitation.CreatedAt = time.Unix(0, createdAt).UTC()
	invitation.UpdatedAt = time.Unix(0, updatedAt).UTC()
	invitation.ClaimedAt = sqliteOptionalTime(claimedAt)
	invitation.RevokedAt = sqliteOptionalTime(revokedAt)
	if claimedByUserID.Valid {
		invitation.ClaimedByUserID = claimedByUserID.String
	}
	if membershipID.Valid {
		invitation.MembershipID = membershipID.String
	}
	if assignmentID.Valid {
		invitation.AssignmentID = assignmentID.String
	}
	if revokedByActorRef.Valid {
		invitation.RevokedByActorRef = revokedByActorRef.String
	}
	var secretDigest workspaceInvitationSecretDigest
	copy(secretDigest[:], digest)
	if !validWorkspaceInvitation(invitation) || !validWorkspaceInvitationSecretDigest(secretDigest) {
		return memoryWorkspaceInvitation{}, errWorkspaceInvitationStoreUnavailable
	}
	return memoryWorkspaceInvitation{invitation: invitation, secretDigest: secretDigest}, nil
}

func normalizeWorkspaceInvitationDurableClaimWriteError(err error) error {
	if errors.Is(err, errWorkspaceInvitationMembershipConflict) || errors.Is(err, errLocalIdentityMembershipConflict) ||
		errors.Is(err, errLocalIdentityRoleAssignmentConflict) {
		return errWorkspaceInvitationMembershipConflict
	}
	return errWorkspaceInvitationStoreUnavailable
}
