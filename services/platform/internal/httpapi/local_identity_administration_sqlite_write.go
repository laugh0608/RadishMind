package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"
)

func (repository *sqliteLocalIdentityRepository) CreateWorkspaceMembershipForAdministration(
	ctx context.Context,
	actorUserID string,
	membership WorkspaceMembership,
	now time.Time,
) error {
	if repository == nil || repository.database == nil {
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
		func(snapshot *memoryLocalIdentityRepository, connection *sql.Conn) error {
			if err := snapshot.CreateWorkspaceMembershipForAdministration(ctx, actorUserID, membership, now); err != nil {
				return err
			}
			return insertSQLiteAdministrationMembership(ctx, connection, membership, errLocalIdentityMembershipConflict)
		},
	)
}

func (repository *sqliteLocalIdentityRepository) CreateCatalogRoleAssignment(
	ctx context.Context,
	actorUserID string,
	assignment LocalRoleAssignment,
	now time.Time,
) error {
	if repository == nil || repository.database == nil {
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
		func(snapshot *memoryLocalIdentityRepository, connection *sql.Conn) error {
			if err := snapshot.CreateCatalogRoleAssignment(ctx, actorUserID, assignment, now); err != nil {
				return err
			}
			return insertSQLiteAdministrationRoleAssignment(
				ctx, connection, assignment, errLocalIdentityRoleAssignmentConflict,
			)
		},
	)
}

func (repository *sqliteLocalIdentityRepository) RevokeCatalogRoleAssignment(
	ctx context.Context,
	actorUserID string,
	tenantRef string,
	workspaceID string,
	assignmentID string,
	expectedVersion int,
	revokedAt time.Time,
	auditRef string,
) (LocalRoleAssignment, error) {
	if repository == nil || repository.database == nil {
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
		func(snapshot *memoryLocalIdentityRepository, connection *sql.Conn) error {
			var err error
			revoked, err = snapshot.RevokeCatalogRoleAssignment(
				ctx, actorUserID, tenantRef, workspaceID, assignmentID, expectedVersion, revokedAt, auditRef,
			)
			if err != nil {
				return err
			}
			result, err := connection.ExecContext(identityContext(ctx), `UPDATE local_role_assignments SET
                lifecycle_state='revoked', record_version=?, updated_at_unix_nano=?, revoked_at_unix_nano=?, audit_ref=?
                WHERE assignment_id=? AND record_version=? AND lifecycle_state='active'`,
				revoked.RecordVersion, revokedAt.UnixNano(), revokedAt.UnixNano(), strings.TrimSpace(auditRef),
				assignmentID, expectedVersion,
			)
			if err != nil {
				return errLocalIdentityStoreUnavailable
			}
			if affected, affectedErr := result.RowsAffected(); affectedErr != nil || affected != 1 {
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

func (repository *sqliteLocalIdentityRepository) RevokeWorkspaceMembershipAndAssignments(
	ctx context.Context,
	tenantRef string,
	workspaceID string,
	membershipID string,
	expectedVersion int,
	actorUserID string,
	revokedAt time.Time,
	auditRef string,
) (LocalIdentityWorkspaceMembershipRevocation, error) {
	if repository == nil || repository.database == nil {
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
		func(snapshot *memoryLocalIdentityRepository, connection *sql.Conn) error {
			var err error
			revocation, err = snapshot.RevokeWorkspaceMembershipAndAssignments(
				ctx, tenantRef, workspaceID, membershipID, expectedVersion, actorUserID, revokedAt, auditRef,
			)
			if err != nil {
				return err
			}
			result, err := connection.ExecContext(identityContext(ctx), `UPDATE local_workspace_memberships SET
                lifecycle_state='revoked', record_version=?, updated_at_unix_nano=?, revoked_at_unix_nano=?, audit_ref=?
                WHERE membership_id=? AND record_version=? AND lifecycle_state='active'`,
				revocation.Membership.RecordVersion, revokedAt.UnixNano(), revokedAt.UnixNano(), strings.TrimSpace(auditRef),
				membershipID, expectedVersion,
			)
			if err != nil {
				return errLocalIdentityStoreUnavailable
			}
			if affected, affectedErr := result.RowsAffected(); affectedErr != nil || affected != 1 {
				return errLocalIdentityStoreUnavailable
			}
			result, err = connection.ExecContext(identityContext(ctx), `UPDATE local_role_assignments SET
                lifecycle_state='revoked', record_version=record_version+1, updated_at_unix_nano=?,
                revoked_at_unix_nano=?, audit_ref=?
                WHERE user_id=? AND tenant_ref=? AND workspace_id=? AND lifecycle_state='active'`,
				revokedAt.UnixNano(), revokedAt.UnixNano(), strings.TrimSpace(auditRef),
				revocation.Membership.UserID, tenantRef, workspaceID,
			)
			if err != nil {
				return errLocalIdentityStoreUnavailable
			}
			if affected, affectedErr := result.RowsAffected(); affectedErr != nil ||
				affected != int64(len(revocation.RevokedRoleAssignments)) {
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

func (repository *sqliteLocalIdentityRepository) BootstrapWorkspaceAdministrator(
	ctx context.Context,
	bootstrap LocalIdentityWorkspaceAdministratorBootstrap,
	now time.Time,
) (LocalIdentityWorkspaceAdministratorBootstrap, error) {
	if repository == nil || repository.database == nil {
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
		func(snapshot *memoryLocalIdentityRepository, connection *sql.Conn) error {
			var err error
			created, err = snapshot.BootstrapWorkspaceAdministrator(ctx, bootstrap, now)
			if err != nil {
				return err
			}
			if err = insertSQLiteAdministrationMembership(
				ctx, connection, created.Membership, errLocalIdentityAdminBootstrapDenied,
			); err != nil {
				return err
			}
			return insertSQLiteAdministrationRoleAssignment(
				ctx, connection, created.RoleAssignment, errLocalIdentityAdminBootstrapDenied,
			)
		},
	)
	if err != nil {
		return LocalIdentityWorkspaceAdministratorBootstrap{}, err
	}
	return created, nil
}

func (repository *sqliteLocalIdentityRepository) mutateLocalIdentityAdministrationScope(
	ctx context.Context,
	tenantRef string,
	workspaceID string,
	additionalUserIDs []string,
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
	snapshot, err := loadSQLiteLocalIdentityAdministrationScope(
		identityContext(ctx), connection, tenantRef, workspaceID, additionalUserIDs...,
	)
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

func insertSQLiteAdministrationMembership(
	ctx context.Context,
	connection *sql.Conn,
	membership WorkspaceMembership,
	conflict error,
) error {
	createdAt, _ := localIdentityUnixNano(membership.CreatedAt)
	updatedAt, _ := localIdentityUnixNano(membership.UpdatedAt)
	expiresAt, _ := optionalLocalIdentityUnixNano(membership.ExpiresAt)
	_, err := connection.ExecContext(identityContext(ctx), `INSERT INTO local_workspace_memberships
        (membership_id, user_id, schema_version, tenant_ref, workspace_id, lifecycle_state, record_version,
         created_at_unix_nano, updated_at_unix_nano, expires_at_unix_nano, revoked_at_unix_nano, audit_ref)
        VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, membership.MembershipID, membership.UserID, membership.SchemaVersion,
		membership.TenantRef, membership.WorkspaceID, membership.LifecycleState, membership.RecordVersion,
		createdAt, updatedAt, expiresAt, nil, membership.AuditRef)
	if err != nil {
		return sqliteConflictOrUnavailable(err, conflict)
	}
	return nil
}

func insertSQLiteAdministrationRoleAssignment(
	ctx context.Context,
	connection *sql.Conn,
	assignment LocalRoleAssignment,
	conflict error,
) error {
	grantsJSON, _ := json.Marshal(assignment.PermissionGrants)
	createdAt, _ := localIdentityUnixNano(assignment.CreatedAt)
	updatedAt, _ := localIdentityUnixNano(assignment.UpdatedAt)
	expiresAt, _ := optionalLocalIdentityUnixNano(assignment.ExpiresAt)
	_, err := connection.ExecContext(identityContext(ctx), `INSERT INTO local_role_assignments
        (assignment_id, user_id, schema_version, tenant_ref, workspace_id, role_key, role_catalog_version,
         role_definition_digest, permission_grants_json, lifecycle_state, record_version, created_at_unix_nano,
         updated_at_unix_nano, expires_at_unix_nano, revoked_at_unix_nano, audit_ref)
        VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, assignment.AssignmentID, assignment.UserID,
		assignment.SchemaVersion, assignment.TenantRef, assignment.WorkspaceID, assignment.RoleKey,
		assignment.RoleCatalogVersion, assignment.RoleDefinitionDigest, string(grantsJSON), assignment.LifecycleState,
		assignment.RecordVersion, createdAt, updatedAt, expiresAt, nil, assignment.AuditRef)
	if err != nil {
		return sqliteConflictOrUnavailable(err, conflict)
	}
	return nil
}
