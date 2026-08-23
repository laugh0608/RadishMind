package httpapi

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type postgresLocalIdentityAdministrationQuery interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func (repository *postgresLocalIdentityRepository) ListWorkspaceMembers(
	ctx context.Context,
	query LocalIdentityWorkspaceMemberListQuery,
) (LocalIdentityWorkspaceMemberPage, error) {
	if repository == nil || repository.pool == nil {
		return LocalIdentityWorkspaceMemberPage{}, errLocalIdentityStoreUnavailable
	}
	filter, cursor, err := normalizeLocalIdentityWorkspaceMemberQuery(query)
	if err != nil {
		return LocalIdentityWorkspaceMemberPage{}, err
	}
	transaction, err := repository.pool.BeginTx(identityContext(ctx), pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return LocalIdentityWorkspaceMemberPage{}, errLocalIdentityStoreUnavailable
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()

	statement := postgresWorkspaceMemberDirectorySelect + `
        WHERE m.tenant_ref=$1 AND m.workspace_id=$2 AND m.lifecycle_state=$3`
	arguments := []any{filter.TenantRef, filter.WorkspaceID, filter.MembershipState}
	if cursor.MembershipID != "" {
		statement += ` AND (m.updated_at < $4 OR (m.updated_at = $4 AND m.membership_id < $5))`
		arguments = append(arguments, mustParseLocalIdentityCursorTime(cursor.UpdatedAt), cursor.MembershipID)
	}
	statement += ` ORDER BY m.updated_at DESC, m.membership_id DESC LIMIT $` + strconv.Itoa(len(arguments)+1)
	arguments = append(arguments, filter.Limit+1)
	rows, err := transaction.Query(identityContext(ctx), statement, arguments...)
	if err != nil {
		return LocalIdentityWorkspaceMemberPage{}, errLocalIdentityStoreUnavailable
	}
	memberships := make([]WorkspaceMembership, 0, filter.Limit+1)
	accounts := make(map[string]UserAccount, filter.Limit+1)
	for rows.Next() {
		membership, account, scanErr := scanPostgresWorkspaceMemberDirectoryRow(rows)
		if scanErr != nil {
			rows.Close()
			return LocalIdentityWorkspaceMemberPage{}, scanErr
		}
		memberships = append(memberships, membership)
		accounts[account.UserID] = account
	}
	if rows.Err() != nil {
		rows.Close()
		return LocalIdentityWorkspaceMemberPage{}, errLocalIdentityStoreUnavailable
	}
	rows.Close()
	hasNext := len(memberships) > filter.Limit
	if hasNext {
		memberships = memberships[:filter.Limit]
	}
	userIDs := make([]string, 0, len(memberships))
	for _, membership := range memberships {
		userIDs = append(userIDs, membership.UserID)
	}
	assignments, err := readPostgresLocalIdentityAssignmentsForUsers(
		identityContext(ctx), transaction, filter.TenantRef, filter.WorkspaceID, userIDs, true,
	)
	if err != nil {
		return LocalIdentityWorkspaceMemberPage{}, err
	}
	page := LocalIdentityWorkspaceMemberPage{Members: make([]LocalIdentityWorkspaceMemberSummary, 0, len(memberships))}
	for _, membership := range memberships {
		page.Members = append(page.Members, buildLocalIdentityWorkspaceMemberSummary(
			accounts[membership.UserID], membership, assignments[membership.UserID], filter.asOf,
		))
	}
	if hasNext && len(memberships) > 0 {
		page.NextCursor, err = encodeLocalIdentityWorkspaceMemberCursor(filter, memberships[len(memberships)-1])
		if err != nil {
			return LocalIdentityWorkspaceMemberPage{}, errLocalIdentityStoreUnavailable
		}
	}
	return page, nil
}

func (repository *postgresLocalIdentityRepository) ReadWorkspaceMember(
	ctx context.Context,
	tenantRef string,
	workspaceID string,
	userID string,
	asOf time.Time,
) (LocalIdentityWorkspaceMemberDetail, error) {
	if repository == nil || repository.pool == nil {
		return LocalIdentityWorkspaceMemberDetail{}, errLocalIdentityStoreUnavailable
	}
	tenantRef = strings.TrimSpace(tenantRef)
	workspaceID = strings.TrimSpace(workspaceID)
	userID = strings.TrimSpace(userID)
	asOf = asOf.UTC()
	if !validControlPlaneReadAuthReference(tenantRef, false) ||
		!validControlPlaneReadAuthReference(workspaceID, false) || !localUserIDPattern.MatchString(userID) || asOf.IsZero() {
		return LocalIdentityWorkspaceMemberDetail{}, errLocalIdentityContractMismatch
	}
	transaction, err := repository.pool.BeginTx(identityContext(ctx), pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return LocalIdentityWorkspaceMemberDetail{}, errLocalIdentityStoreUnavailable
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	account, err := scanPostgresUserAccount(transaction.QueryRow(
		identityContext(ctx), postgresAccountSelect+` WHERE user_id=$1`, userID,
	))
	if err != nil {
		return LocalIdentityWorkspaceMemberDetail{}, err
	}
	memberships, err := readPostgresLocalIdentityMemberships(
		identityContext(ctx), transaction, tenantRef, workspaceID, userID, false,
	)
	if err != nil {
		return LocalIdentityWorkspaceMemberDetail{}, err
	}
	if len(memberships) == 0 {
		return LocalIdentityWorkspaceMemberDetail{}, errLocalIdentityMemberUnavailable
	}
	assignmentMap, err := readPostgresLocalIdentityAssignmentsForUsers(
		identityContext(ctx), transaction, tenantRef, workspaceID, []string{userID}, false,
	)
	if err != nil {
		return LocalIdentityWorkspaceMemberDetail{}, err
	}
	return buildLocalIdentityWorkspaceMemberDetail(
		account, tenantRef, workspaceID, memberships, assignmentMap[userID], asOf,
	), nil
}

func loadPostgresLocalIdentityAdministrationScope(
	ctx context.Context,
	query postgresLocalIdentityAdministrationQuery,
	tenantRef string,
	workspaceID string,
	additionalUserIDs ...string,
) (*memoryLocalIdentityRepository, error) {
	memberships, err := readPostgresLocalIdentityMemberships(ctx, query, tenantRef, workspaceID, "", true)
	if err != nil {
		return nil, err
	}
	assignmentMap, err := readPostgresLocalIdentityAssignmentsForUsers(ctx, query, tenantRef, workspaceID, nil, true)
	if err != nil {
		return nil, err
	}
	assignments := make([]LocalRoleAssignment, 0)
	userSet := make(map[string]struct{})
	for _, membership := range memberships {
		userSet[membership.UserID] = struct{}{}
	}
	for userID, userAssignments := range assignmentMap {
		userSet[userID] = struct{}{}
		assignments = append(assignments, userAssignments...)
	}
	for _, userID := range additionalUserIDs {
		userID = strings.TrimSpace(userID)
		if localUserIDPattern.MatchString(userID) {
			userSet[userID] = struct{}{}
		}
	}
	userIDs := make([]string, 0, len(userSet))
	for userID := range userSet {
		userIDs = append(userIDs, userID)
	}
	slices.Sort(userIDs)
	accounts, err := readPostgresLocalIdentityAccounts(ctx, query, userIDs)
	if err != nil {
		return nil, err
	}
	return newLocalIdentityAdministrationScopeSnapshot(accounts, memberships, assignments)
}

func readPostgresLocalIdentityMemberships(
	ctx context.Context,
	query postgresLocalIdentityAdministrationQuery,
	tenantRef string,
	workspaceID string,
	userID string,
	activeOnly bool,
) ([]WorkspaceMembership, error) {
	statement := postgresMembershipSelect + ` WHERE tenant_ref=$1 AND workspace_id=$2`
	arguments := []any{tenantRef, workspaceID}
	if userID != "" {
		statement += ` AND user_id=$3`
		arguments = append(arguments, userID)
	}
	if activeOnly {
		statement += ` AND lifecycle_state='active'`
	}
	statement += ` ORDER BY updated_at DESC, membership_id DESC`
	rows, err := query.Query(ctx, statement, arguments...)
	if err != nil {
		return nil, errLocalIdentityStoreUnavailable
	}
	memberships := make([]WorkspaceMembership, 0)
	for rows.Next() {
		membership, scanErr := scanPostgresWorkspaceMembership(rows)
		if scanErr != nil {
			rows.Close()
			return nil, scanErr
		}
		memberships = append(memberships, membership)
	}
	if rows.Err() != nil {
		rows.Close()
		return nil, errLocalIdentityStoreUnavailable
	}
	rows.Close()
	return memberships, nil
}

func readPostgresLocalIdentityAssignmentsForUsers(
	ctx context.Context,
	query postgresLocalIdentityAdministrationQuery,
	tenantRef string,
	workspaceID string,
	userIDs []string,
	activeOnly bool,
) (map[string][]LocalRoleAssignment, error) {
	if userIDs != nil && len(userIDs) == 0 {
		return map[string][]LocalRoleAssignment{}, nil
	}
	statement := postgresRoleAssignmentSelect + ` WHERE tenant_ref=$1 AND (workspace_id='' OR workspace_id=$2)`
	arguments := []any{tenantRef, workspaceID}
	if activeOnly {
		statement += ` AND lifecycle_state='active'`
	}
	if userIDs != nil {
		statement += ` AND user_id=ANY($3::text[])`
		arguments = append(arguments, userIDs)
	}
	statement += ` ORDER BY assignment_id`
	rows, err := query.Query(ctx, statement, arguments...)
	if err != nil {
		return nil, errLocalIdentityStoreUnavailable
	}
	assignments := make(map[string][]LocalRoleAssignment)
	for rows.Next() {
		assignment, scanErr := scanPostgresLocalRoleAssignment(rows)
		if scanErr != nil {
			rows.Close()
			return nil, scanErr
		}
		assignments[assignment.UserID] = append(assignments[assignment.UserID], assignment)
	}
	if rows.Err() != nil {
		rows.Close()
		return nil, errLocalIdentityStoreUnavailable
	}
	rows.Close()
	return assignments, nil
}

func readPostgresLocalIdentityAccounts(
	ctx context.Context,
	query postgresLocalIdentityAdministrationQuery,
	userIDs []string,
) ([]UserAccount, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	rows, err := query.Query(ctx, postgresAccountSelect+` WHERE user_id=ANY($1::text[])`, userIDs)
	if err != nil {
		return nil, errLocalIdentityStoreUnavailable
	}
	accounts := make([]UserAccount, 0, len(userIDs))
	for rows.Next() {
		account, scanErr := scanPostgresUserAccount(rows)
		if scanErr != nil {
			rows.Close()
			return nil, scanErr
		}
		accounts = append(accounts, account)
	}
	if rows.Err() != nil {
		rows.Close()
		return nil, errLocalIdentityStoreUnavailable
	}
	rows.Close()
	return accounts, nil
}

const postgresWorkspaceMemberDirectorySelect = `SELECT
    m.schema_version, m.membership_id, m.user_id, m.tenant_ref, m.workspace_id, m.lifecycle_state,
    m.record_version, m.created_at, m.updated_at, m.expires_at, m.revoked_at, m.audit_ref,
    a.schema_version, a.user_id, a.login_identifier, a.normalized_login_identifier, a.display_name,
    a.lifecycle_state, a.record_version, a.created_at, a.updated_at, a.disabled_at, a.audit_ref
    FROM local_workspace_memberships m
    JOIN local_user_accounts a ON a.user_id=m.user_id`

func scanPostgresWorkspaceMemberDirectoryRow(row localIdentitySQLRow) (WorkspaceMembership, UserAccount, error) {
	var membership WorkspaceMembership
	var account UserAccount
	err := row.Scan(
		&membership.SchemaVersion, &membership.MembershipID, &membership.UserID, &membership.TenantRef,
		&membership.WorkspaceID, &membership.LifecycleState, &membership.RecordVersion,
		&membership.CreatedAt, &membership.UpdatedAt, &membership.ExpiresAt, &membership.RevokedAt, &membership.AuditRef,
		&account.SchemaVersion, &account.UserID, &account.LoginIdentifier, &account.NormalizedLoginIdentifier,
		&account.DisplayName, &account.LifecycleState, &account.RecordVersion, &account.CreatedAt, &account.UpdatedAt,
		&account.DisabledAt, &account.AuditRef,
	)
	if err != nil {
		return WorkspaceMembership{}, UserAccount{}, normalizePostgresReadError(err)
	}
	membership.CreatedAt = membership.CreatedAt.UTC()
	membership.UpdatedAt = membership.UpdatedAt.UTC()
	membership.ExpiresAt = cloneTimePointer(membership.ExpiresAt)
	membership.RevokedAt = cloneTimePointer(membership.RevokedAt)
	account.CreatedAt = account.CreatedAt.UTC()
	account.UpdatedAt = account.UpdatedAt.UTC()
	account.DisabledAt = cloneTimePointer(account.DisabledAt)
	if !validWorkspaceMembership(membership) || !validUserAccount(account) || membership.UserID != account.UserID {
		return WorkspaceMembership{}, UserAccount{}, errLocalIdentityStoreUnavailable
	}
	return membership, account, nil
}
