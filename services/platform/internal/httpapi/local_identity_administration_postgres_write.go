package httpapi

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func (repository *postgresLocalIdentityRepository) CreateWorkspaceMembershipForAdministration(
	ctx context.Context,
	actorUserID string,
	membership WorkspaceMembership,
	now time.Time,
) error {
	if repository == nil || repository.pool == nil {
		return errLocalIdentityStoreUnavailable
	}
	actorUserID = strings.TrimSpace(actorUserID)
	now = now.UTC()
	if !localUserIDPattern.MatchString(actorUserID) || now.IsZero() ||
		!validWorkspaceMembership(membership) || membership.LifecycleState != localIdentityStateActive ||
		!membership.CreatedAt.Equal(now) || !membership.UpdatedAt.Equal(now) {
		return errLocalIdentityContractMismatch
	}
	return repository.mutateLocalIdentityAdministrationScope(
		ctx,
		membership.TenantRef,
		membership.WorkspaceID,
		[]string{actorUserID, membership.UserID},
		func(snapshot *memoryLocalIdentityRepository, transaction pgx.Tx) error {
			if err := snapshot.CreateWorkspaceMembershipForAdministration(ctx, actorUserID, membership, now); err != nil {
				return err
			}
			return insertPostgresAdministrationMembership(
				ctx, transaction, membership, errLocalIdentityMembershipConflict,
			)
		},
	)
}

func (repository *postgresLocalIdentityRepository) CreateCatalogRoleAssignment(
	ctx context.Context,
	actorUserID string,
	assignment LocalRoleAssignment,
	now time.Time,
) error {
	if repository == nil || repository.pool == nil {
		return errLocalIdentityStoreUnavailable
	}
	actorUserID = strings.TrimSpace(actorUserID)
	now = now.UTC()
	grants, ok := normalizedPermissionGrants(assignment.PermissionGrants)
	assignment.PermissionGrants = grants
	definition, exists := builtInLocalIdentityRole(assignment.RoleKey)
	if !localUserIDPattern.MatchString(actorUserID) || now.IsZero() || !ok || !exists ||
		!validLocalRoleAssignment(assignment) || assignment.LifecycleState != localIdentityStateActive ||
		!assignment.CreatedAt.Equal(now) || !assignment.UpdatedAt.Equal(now) ||
		!localIdentityRoleDefinitionMatchesAssignment(definition, assignment) {
		return errLocalIdentityRoleCatalogMismatch
	}
	return repository.mutateLocalIdentityAdministrationScope(
		ctx,
		assignment.TenantRef,
		assignment.WorkspaceID,
		[]string{actorUserID, assignment.UserID},
		func(snapshot *memoryLocalIdentityRepository, transaction pgx.Tx) error {
			if err := snapshot.CreateCatalogRoleAssignment(ctx, actorUserID, assignment, now); err != nil {
				return err
			}
			return insertPostgresAdministrationRoleAssignment(
				ctx, transaction, assignment, errLocalIdentityRoleAssignmentConflict,
			)
		},
	)
}

func (repository *postgresLocalIdentityRepository) RevokeCatalogRoleAssignment(
	ctx context.Context,
	actorUserID string,
	tenantRef string,
	workspaceID string,
	assignmentID string,
	expectedVersion int,
	revokedAt time.Time,
	auditRef string,
) (LocalRoleAssignment, error) {
	if repository == nil || repository.pool == nil {
		return LocalRoleAssignment{}, errLocalIdentityStoreUnavailable
	}
	actorUserID = strings.TrimSpace(actorUserID)
	tenantRef = strings.TrimSpace(tenantRef)
	workspaceID = strings.TrimSpace(workspaceID)
	assignmentID = strings.TrimSpace(assignmentID)
	revokedAt = revokedAt.UTC()
	if !localUserIDPattern.MatchString(actorUserID) || !validControlPlaneReadAuthReference(tenantRef, false) ||
		!validControlPlaneReadAuthReference(workspaceID, false) ||
		!localRoleAssignmentIDPattern.MatchString(assignmentID) || expectedVersion < 1 || revokedAt.IsZero() ||
		!validAuditRef(auditRef) {
		return LocalRoleAssignment{}, errLocalIdentityContractMismatch
	}
	var revoked LocalRoleAssignment
	err := repository.mutateLocalIdentityAdministrationScope(
		ctx,
		tenantRef,
		workspaceID,
		[]string{actorUserID},
		func(snapshot *memoryLocalIdentityRepository, transaction pgx.Tx) error {
			var err error
			revoked, err = snapshot.RevokeCatalogRoleAssignment(
				ctx, actorUserID, tenantRef, workspaceID, assignmentID, expectedVersion, revokedAt, auditRef,
			)
			if err != nil {
				return err
			}
			command, err := transaction.Exec(identityContext(ctx), `UPDATE local_role_assignments SET
                lifecycle_state='revoked', record_version=$1, updated_at=$2, revoked_at=$2, audit_ref=$3
                WHERE assignment_id=$4 AND record_version=$5 AND lifecycle_state='active'`,
				revoked.RecordVersion, revokedAt, strings.TrimSpace(auditRef), assignmentID, expectedVersion,
			)
			if err != nil || command.RowsAffected() != 1 {
				return errLocalIdentityStoreUnavailable
			}
			return nil
		},
	)
	if err != nil {
		return LocalRoleAssignment{}, err
	}
	return cloneLocalRoleAssignment(revoked), nil
}

func (repository *postgresLocalIdentityRepository) RevokeWorkspaceMembershipAndAssignments(
	ctx context.Context,
	tenantRef string,
	workspaceID string,
	membershipID string,
	expectedVersion int,
	actorUserID string,
	revokedAt time.Time,
	auditRef string,
) (LocalIdentityWorkspaceMembershipRevocation, error) {
	if repository == nil || repository.pool == nil {
		return LocalIdentityWorkspaceMembershipRevocation{}, errLocalIdentityStoreUnavailable
	}
	tenantRef = strings.TrimSpace(tenantRef)
	workspaceID = strings.TrimSpace(workspaceID)
	membershipID = strings.TrimSpace(membershipID)
	actorUserID = strings.TrimSpace(actorUserID)
	revokedAt = revokedAt.UTC()
	if !validControlPlaneReadAuthReference(tenantRef, false) ||
		!validControlPlaneReadAuthReference(workspaceID, false) ||
		!localMembershipIDPattern.MatchString(membershipID) || !localUserIDPattern.MatchString(actorUserID) ||
		expectedVersion < 1 || revokedAt.IsZero() || !validAuditRef(auditRef) {
		return LocalIdentityWorkspaceMembershipRevocation{}, errLocalIdentityContractMismatch
	}
	var revocation LocalIdentityWorkspaceMembershipRevocation
	err := repository.mutateLocalIdentityAdministrationScope(
		ctx,
		tenantRef,
		workspaceID,
		[]string{actorUserID},
		func(snapshot *memoryLocalIdentityRepository, transaction pgx.Tx) error {
			var err error
			revocation, err = snapshot.RevokeWorkspaceMembershipAndAssignments(
				ctx, tenantRef, workspaceID, membershipID, expectedVersion, actorUserID, revokedAt, auditRef,
			)
			if err != nil {
				return err
			}
			command, err := transaction.Exec(identityContext(ctx), `UPDATE local_workspace_memberships SET
                lifecycle_state='revoked', record_version=$1, updated_at=$2, revoked_at=$2, audit_ref=$3
                WHERE membership_id=$4 AND record_version=$5 AND lifecycle_state='active'`,
				revocation.Membership.RecordVersion, revokedAt, strings.TrimSpace(auditRef), membershipID, expectedVersion,
			)
			if err != nil || command.RowsAffected() != 1 {
				return errLocalIdentityStoreUnavailable
			}
			command, err = transaction.Exec(identityContext(ctx), `UPDATE local_role_assignments SET
                lifecycle_state='revoked', record_version=record_version+1, updated_at=$1,
                revoked_at=$1, audit_ref=$2
                WHERE user_id=$3 AND tenant_ref=$4 AND workspace_id=$5 AND lifecycle_state='active'`,
				revokedAt, strings.TrimSpace(auditRef), revocation.Membership.UserID, tenantRef, workspaceID,
			)
			if err != nil || command.RowsAffected() != int64(len(revocation.RevokedRoleAssignments)) {
				return errLocalIdentityStoreUnavailable
			}
			return nil
		},
	)
	if err != nil {
		return LocalIdentityWorkspaceMembershipRevocation{}, err
	}
	return revocation, nil
}

func (repository *postgresLocalIdentityRepository) BootstrapWorkspaceAdministrator(
	ctx context.Context,
	bootstrap LocalIdentityWorkspaceAdministratorBootstrap,
	now time.Time,
) (LocalIdentityWorkspaceAdministratorBootstrap, error) {
	if repository == nil || repository.pool == nil {
		return LocalIdentityWorkspaceAdministratorBootstrap{}, errLocalIdentityStoreUnavailable
	}
	now = now.UTC()
	membership := bootstrap.Membership
	assignment := bootstrap.RoleAssignment
	definition, exists := builtInLocalIdentityRole(localIdentityRoleWorkspaceAdmin)
	if now.IsZero() || !validWorkspaceMembership(membership) || !validLocalRoleAssignment(assignment) ||
		membership.LifecycleState != localIdentityStateActive || assignment.LifecycleState != localIdentityStateActive ||
		membership.UserID != assignment.UserID || membership.TenantRef != assignment.TenantRef ||
		membership.WorkspaceID != assignment.WorkspaceID || !membership.CreatedAt.Equal(now) ||
		!membership.UpdatedAt.Equal(now) || !assignment.CreatedAt.Equal(now) || !assignment.UpdatedAt.Equal(now) ||
		!exists || !definition.CanManageLocalIdentity || !localIdentityRoleDefinitionMatchesAssignment(definition, assignment) {
		return LocalIdentityWorkspaceAdministratorBootstrap{}, errLocalIdentityContractMismatch
	}
	var created LocalIdentityWorkspaceAdministratorBootstrap
	err := repository.mutateLocalIdentityAdministrationScope(
		ctx,
		membership.TenantRef,
		membership.WorkspaceID,
		[]string{membership.UserID},
		func(snapshot *memoryLocalIdentityRepository, transaction pgx.Tx) error {
			var err error
			created, err = snapshot.BootstrapWorkspaceAdministrator(ctx, bootstrap, now)
			if err != nil {
				return err
			}
			if err = insertPostgresAdministrationMembership(
				ctx, transaction, created.Membership, errLocalIdentityAdminBootstrapDenied,
			); err != nil {
				return err
			}
			return insertPostgresAdministrationRoleAssignment(
				ctx, transaction, created.RoleAssignment, errLocalIdentityAdminBootstrapDenied,
			)
		},
	)
	if err != nil {
		return LocalIdentityWorkspaceAdministratorBootstrap{}, err
	}
	return created, nil
}

func (repository *postgresLocalIdentityRepository) mutateLocalIdentityAdministrationScope(
	ctx context.Context,
	tenantRef string,
	workspaceID string,
	additionalUserIDs []string,
	operation func(*memoryLocalIdentityRepository, pgx.Tx) error,
) error {
	transaction, err := repository.pool.Begin(identityContext(ctx))
	if err != nil {
		return errLocalIdentityStoreUnavailable
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	if err = lockPostgresLocalIdentityAdministrationScope(
		identityContext(ctx), transaction, tenantRef, workspaceID,
	); err != nil {
		return errLocalIdentityStoreUnavailable
	}
	snapshot, err := loadPostgresLocalIdentityAdministrationScope(
		identityContext(ctx), transaction, tenantRef, workspaceID, additionalUserIDs...,
	)
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

func lockPostgresLocalIdentityAdministrationScope(
	ctx context.Context,
	transaction pgx.Tx,
	tenantRef string,
	workspaceID string,
) error {
	if transaction == nil || !validControlPlaneReadAuthReference(strings.TrimSpace(tenantRef), false) ||
		!validControlPlaneReadAuthReference(strings.TrimSpace(workspaceID), false) {
		return errLocalIdentityContractMismatch
	}
	if _, err := transaction.Exec(
		identityContext(ctx), "SELECT pg_advisory_xact_lock($1)",
		localIdentityAdministrationScopeLockKey(tenantRef, workspaceID),
	); err != nil {
		return errLocalIdentityStoreUnavailable
	}
	return nil
}

func insertPostgresAdministrationMembership(
	ctx context.Context,
	transaction pgx.Tx,
	membership WorkspaceMembership,
	conflict error,
) error {
	_, err := transaction.Exec(identityContext(ctx), `INSERT INTO local_workspace_memberships
        (membership_id, user_id, schema_version, tenant_ref, workspace_id, lifecycle_state, record_version,
         created_at, updated_at, expires_at, revoked_at, audit_ref)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, membership.MembershipID, membership.UserID,
		membership.SchemaVersion, membership.TenantRef, membership.WorkspaceID, membership.LifecycleState,
		membership.RecordVersion, membership.CreatedAt, membership.UpdatedAt, membership.ExpiresAt,
		membership.RevokedAt, membership.AuditRef)
	if err != nil {
		return postgresIdentityConflictOrUnavailable(err, conflict)
	}
	return nil
}

func insertPostgresAdministrationRoleAssignment(
	ctx context.Context,
	transaction pgx.Tx,
	assignment LocalRoleAssignment,
	conflict error,
) error {
	_, err := transaction.Exec(identityContext(ctx), `INSERT INTO local_role_assignments
        (assignment_id, user_id, schema_version, tenant_ref, workspace_id, role_key, role_catalog_version,
         role_definition_digest, permission_grants, lifecycle_state, record_version, created_at, updated_at,
         expires_at, revoked_at, audit_ref)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`, assignment.AssignmentID,
		assignment.UserID, assignment.SchemaVersion, assignment.TenantRef, assignment.WorkspaceID, assignment.RoleKey,
		assignment.RoleCatalogVersion, assignment.RoleDefinitionDigest, assignment.PermissionGrants,
		assignment.LifecycleState, assignment.RecordVersion, assignment.CreatedAt, assignment.UpdatedAt,
		assignment.ExpiresAt, assignment.RevokedAt, assignment.AuditRef)
	if err != nil {
		return postgresIdentityConflictOrUnavailable(err, conflict)
	}
	return nil
}
