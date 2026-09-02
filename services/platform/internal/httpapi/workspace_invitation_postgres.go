package httpapi

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const postgresWorkspaceInvitationSelect = `SELECT
    schema_version, invitation_id, record_version, tenant_ref, workspace_id, role_key,
    role_catalog_version, role_definition_digest, ttl_policy, lifecycle_state, secret_digest,
    expires_at, created_at, updated_at, created_by_actor_ref, created_request_ref, created_audit_ref,
    updated_request_ref, updated_audit_ref, claimed_at, claimed_by_user_id, membership_id, assignment_id,
    revoked_at, revoked_by_actor_ref
    FROM local_workspace_invitations`

func (repository *postgresLocalIdentityRepository) CreateWorkspaceInvitation(
	ctx context.Context,
	actorUserID string,
	invitation WorkspaceInvitation,
	secretDigest workspaceInvitationSecretDigest,
	now time.Time,
) error {
	if repository == nil || repository.pool == nil {
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
		func(snapshot *memoryLocalIdentityRepository, transaction pgx.Tx) error {
			if err := snapshot.CreateWorkspaceInvitation(ctx, actorUserID, invitation, secretDigest, now); err != nil {
				return err
			}
			if err := insertPostgresWorkspaceInvitation(ctx, transaction, invitation, secretDigest); err != nil {
				return errWorkspaceInvitationAdminUnavailable
			}
			return nil
		},
	)
}

func (repository *postgresLocalIdentityRepository) ListWorkspaceInvitations(
	ctx context.Context,
	actorUserID string,
	query WorkspaceInvitationListQuery,
) (WorkspaceInvitationPage, error) {
	if repository == nil || repository.pool == nil {
		return WorkspaceInvitationPage{}, errWorkspaceInvitationAdminUnavailable
	}
	filter, cursor, err := normalizeWorkspaceInvitationListQuery(query)
	if err != nil {
		return WorkspaceInvitationPage{}, err
	}
	transaction, err := repository.pool.BeginTx(identityContext(ctx), pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return WorkspaceInvitationPage{}, errWorkspaceInvitationAdminUnavailable
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	snapshot, err := loadPostgresLocalIdentityAdministrationScope(
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
	statement := postgresWorkspaceInvitationSelect + ` WHERE tenant_ref=$1 AND workspace_id=$2`
	arguments := []any{filter.TenantRef, filter.WorkspaceID}
	switch filter.EffectiveState {
	case workspaceInvitationEffectivePending:
		arguments = append(arguments, filter.asOf)
		statement += ` AND lifecycle_state='pending' AND expires_at>$3`
	case workspaceInvitationEffectiveExpired:
		arguments = append(arguments, filter.asOf)
		statement += ` AND lifecycle_state='pending' AND expires_at<=$3`
	case workspaceInvitationEffectiveClaimed, workspaceInvitationEffectiveRevoked:
		arguments = append(arguments, filter.EffectiveState)
		statement += ` AND lifecycle_state=$3`
	default:
		return WorkspaceInvitationPage{}, errWorkspaceInvitationCursorInvalid
	}
	if cursor.InvitationID != "" {
		arguments = append(arguments, mustParseLocalIdentityCursorTime(cursor.UpdatedAt), cursor.InvitationID)
		anchor := len(arguments) - 1
		statement += ` AND (updated_at<$` + strconv.Itoa(anchor) + ` OR (updated_at=$` +
			strconv.Itoa(anchor) + ` AND invitation_id<$` + strconv.Itoa(anchor+1) + `))`
	}
	arguments = append(arguments, filter.Limit+1)
	statement += ` ORDER BY updated_at DESC, invitation_id DESC LIMIT $` + strconv.Itoa(len(arguments))
	rows, err := transaction.Query(identityContext(ctx), statement, arguments...)
	if err != nil {
		return WorkspaceInvitationPage{}, errWorkspaceInvitationAdminUnavailable
	}
	stored := make([]memoryWorkspaceInvitation, 0, filter.Limit+1)
	for rows.Next() {
		item, scanErr := scanPostgresWorkspaceInvitation(rows)
		if scanErr != nil {
			rows.Close()
			return WorkspaceInvitationPage{}, errWorkspaceInvitationAdminUnavailable
		}
		stored = append(stored, item)
	}
	if rows.Err() != nil {
		rows.Close()
		return WorkspaceInvitationPage{}, errWorkspaceInvitationAdminUnavailable
	}
	rows.Close()
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

func (repository *postgresLocalIdentityRepository) RevokeWorkspaceInvitation(
	ctx context.Context,
	actorUserID string,
	input WorkspaceInvitationRevokeInput,
	now time.Time,
) (WorkspaceInvitation, error) {
	if repository == nil || repository.pool == nil {
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
		func(snapshot *memoryLocalIdentityRepository, transaction pgx.Tx) error {
			stored, readErr := scanPostgresWorkspaceInvitation(transaction.QueryRow(
				identityContext(ctx), postgresWorkspaceInvitationSelect+` WHERE invitation_id=$1 FOR UPDATE`,
				strings.TrimSpace(input.InvitationID),
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
			command, updateErr := transaction.Exec(identityContext(ctx), `UPDATE local_workspace_invitations SET
                record_version=$1, lifecycle_state='revoked', updated_at=$2, updated_request_ref=$3,
                updated_audit_ref=$4, revoked_at=$5, revoked_by_actor_ref=$6
                WHERE invitation_id=$7 AND record_version=$8 AND lifecycle_state='pending'`,
				revoked.RecordVersion, revoked.UpdatedAt, revoked.UpdatedRequestRef, revoked.UpdatedAuditRef,
				revoked.RevokedAt, revoked.RevokedByActorRef, revoked.InvitationID, input.ExpectedVersion,
			)
			if updateErr != nil || command.RowsAffected() != 1 {
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

func (repository *postgresLocalIdentityRepository) PreviewWorkspaceInvitation(
	ctx context.Context,
	claimantUserID string,
	claimantTenantRef string,
	invitationID string,
	secretDigest workspaceInvitationSecretDigest,
	now time.Time,
) (WorkspaceInvitation, error) {
	if repository == nil || repository.pool == nil {
		return WorkspaceInvitation{}, errWorkspaceInvitationStoreUnavailable
	}
	transaction, err := repository.pool.BeginTx(identityContext(ctx), pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return WorkspaceInvitation{}, errWorkspaceInvitationStoreUnavailable
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	stored, err := scanPostgresWorkspaceInvitation(transaction.QueryRow(
		identityContext(ctx), postgresWorkspaceInvitationSelect+` WHERE invitation_id=$1`, strings.TrimSpace(invitationID),
	))
	if errors.Is(err, errLocalIdentityNotFound) {
		_ = workspaceInvitationSecretMatches(secretDigest, workspaceInvitationDummySecretDigest)
		return WorkspaceInvitation{}, errWorkspaceInvitationInvalid
	}
	if err != nil {
		return WorkspaceInvitation{}, errWorkspaceInvitationStoreUnavailable
	}
	accounts, err := readPostgresLocalIdentityAccounts(
		identityContext(ctx), transaction, []string{strings.TrimSpace(claimantUserID)},
	)
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

func (repository *postgresLocalIdentityRepository) ClaimWorkspaceInvitation(
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
	if repository == nil || repository.pool == nil {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationStoreUnavailable
	}
	transaction, err := repository.pool.Begin(identityContext(ctx))
	if err != nil {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationStoreUnavailable
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	located, err := scanPostgresWorkspaceInvitation(transaction.QueryRow(
		identityContext(ctx), postgresWorkspaceInvitationSelect+` WHERE invitation_id=$1`, strings.TrimSpace(invitationID),
	))
	if errors.Is(err, errLocalIdentityNotFound) {
		_ = workspaceInvitationSecretMatches(secretDigest, workspaceInvitationDummySecretDigest)
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationInvalid
	}
	if err != nil {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationStoreUnavailable
	}
	if err = lockPostgresLocalIdentityAdministrationScope(
		identityContext(ctx), transaction, located.invitation.TenantRef, located.invitation.WorkspaceID,
	); err != nil {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationStoreUnavailable
	}
	stored, err := scanPostgresWorkspaceInvitation(transaction.QueryRow(
		identityContext(ctx), postgresWorkspaceInvitationSelect+` WHERE invitation_id=$1 FOR UPDATE`, strings.TrimSpace(invitationID),
	))
	if err != nil {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationStoreUnavailable
	}
	snapshot, err := loadPostgresLocalIdentityAdministrationScope(
		identityContext(ctx), transaction, stored.invitation.TenantRef, stored.invitation.WorkspaceID,
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
	if err = insertPostgresAdministrationMembership(
		ctx, transaction, membership, errWorkspaceInvitationMembershipConflict,
	); err != nil {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, normalizeWorkspaceInvitationDurableClaimWriteError(err)
	}
	if err = insertPostgresAdministrationRoleAssignment(
		ctx, transaction, assignment, errWorkspaceInvitationMembershipConflict,
	); err != nil {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, normalizeWorkspaceInvitationDurableClaimWriteError(err)
	}
	command, err := transaction.Exec(identityContext(ctx), `UPDATE local_workspace_invitations SET
        record_version=$1, lifecycle_state='claimed', updated_at=$2, updated_request_ref=$3,
        updated_audit_ref=$4, claimed_at=$5, claimed_by_user_id=$6, membership_id=$7, assignment_id=$8
        WHERE invitation_id=$9 AND record_version=$10 AND lifecycle_state='pending' AND expires_at>$11`,
		invitation.RecordVersion, invitation.UpdatedAt, invitation.UpdatedRequestRef, invitation.UpdatedAuditRef,
		invitation.ClaimedAt, invitation.ClaimedByUserID, invitation.MembershipID, invitation.AssignmentID,
		invitation.InvitationID, expectedVersion, now.UTC(),
	)
	if err != nil || command.RowsAffected() != 1 {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationStoreUnavailable
	}
	if err = transaction.Commit(identityContext(ctx)); err != nil {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationStoreUnavailable
	}
	return invitation, membership, assignment, nil
}

func insertPostgresWorkspaceInvitation(
	ctx context.Context,
	transaction pgx.Tx,
	invitation WorkspaceInvitation,
	secretDigest workspaceInvitationSecretDigest,
) error {
	_, err := transaction.Exec(identityContext(ctx), `INSERT INTO local_workspace_invitations
        (invitation_id, schema_version, record_version, tenant_ref, workspace_id, role_key,
         role_catalog_version, role_definition_digest, ttl_policy, lifecycle_state, secret_digest,
         expires_at, created_at, updated_at, created_by_actor_ref, created_request_ref, created_audit_ref,
         updated_request_ref, updated_audit_ref, claimed_at, claimed_by_user_id, membership_id,
         assignment_id, revoked_at, revoked_by_actor_ref)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25)`,
		invitation.InvitationID, invitation.SchemaVersion, invitation.RecordVersion, invitation.TenantRef,
		invitation.WorkspaceID, invitation.RoleKey, invitation.RoleCatalogVersion, invitation.RoleDefinitionDigest,
		invitation.TTLPolicy, invitation.LifecycleState, secretDigest[:], invitation.ExpiresAt,
		invitation.CreatedAt, invitation.UpdatedAt, invitation.CreatedByActorRef, invitation.CreatedRequestRef,
		invitation.CreatedAuditRef, invitation.UpdatedRequestRef, invitation.UpdatedAuditRef,
		invitation.ClaimedAt, nil, nil, nil, invitation.RevokedAt, nil,
	)
	return err
}

func scanPostgresWorkspaceInvitation(row localIdentitySQLRow) (memoryWorkspaceInvitation, error) {
	var invitation WorkspaceInvitation
	var digest []byte
	var claimedAt, revokedAt *time.Time
	var claimedByUserID, membershipID, assignmentID, revokedByActorRef *string
	err := row.Scan(
		&invitation.SchemaVersion, &invitation.InvitationID, &invitation.RecordVersion,
		&invitation.TenantRef, &invitation.WorkspaceID, &invitation.RoleKey,
		&invitation.RoleCatalogVersion, &invitation.RoleDefinitionDigest, &invitation.TTLPolicy,
		&invitation.LifecycleState, &digest, &invitation.ExpiresAt, &invitation.CreatedAt, &invitation.UpdatedAt,
		&invitation.CreatedByActorRef, &invitation.CreatedRequestRef, &invitation.CreatedAuditRef,
		&invitation.UpdatedRequestRef, &invitation.UpdatedAuditRef, &claimedAt, &claimedByUserID,
		&membershipID, &assignmentID, &revokedAt, &revokedByActorRef,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return memoryWorkspaceInvitation{}, errLocalIdentityNotFound
	}
	if err != nil || len(digest) != len(workspaceInvitationSecretDigest{}) {
		return memoryWorkspaceInvitation{}, errWorkspaceInvitationStoreUnavailable
	}
	invitation.ExpiresAt = invitation.ExpiresAt.UTC()
	invitation.CreatedAt = invitation.CreatedAt.UTC()
	invitation.UpdatedAt = invitation.UpdatedAt.UTC()
	if claimedAt != nil {
		value := claimedAt.UTC()
		invitation.ClaimedAt = &value
	}
	if revokedAt != nil {
		value := revokedAt.UTC()
		invitation.RevokedAt = &value
	}
	if claimedByUserID != nil {
		invitation.ClaimedByUserID = *claimedByUserID
	}
	if membershipID != nil {
		invitation.MembershipID = *membershipID
	}
	if assignmentID != nil {
		invitation.AssignmentID = *assignmentID
	}
	if revokedByActorRef != nil {
		invitation.RevokedByActorRef = *revokedByActorRef
	}
	var secretDigest workspaceInvitationSecretDigest
	copy(secretDigest[:], digest)
	if !validWorkspaceInvitation(invitation) || !validWorkspaceInvitationSecretDigest(secretDigest) {
		return memoryWorkspaceInvitation{}, errWorkspaceInvitationStoreUnavailable
	}
	return memoryWorkspaceInvitation{invitation: invitation, secretDigest: secretDigest}, nil
}
