package httpapi

import (
	"context"
	"database/sql"
	"slices"
	"strings"
	"time"
)

type sqliteLocalIdentityAdministrationQuery interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (repository *sqliteLocalIdentityRepository) ListWorkspaceMembers(
	ctx context.Context,
	query LocalIdentityWorkspaceMemberListQuery,
) (LocalIdentityWorkspaceMemberPage, error) {
	if repository == nil || repository.database == nil {
		return LocalIdentityWorkspaceMemberPage{}, errLocalIdentityStoreUnavailable
	}
	filter, cursor, err := normalizeLocalIdentityWorkspaceMemberQuery(query)
	if err != nil {
		return LocalIdentityWorkspaceMemberPage{}, err
	}
	transaction, err := repository.database.BeginTx(identityContext(ctx), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return LocalIdentityWorkspaceMemberPage{}, errLocalIdentityStoreUnavailable
	}
	defer func() { _ = transaction.Rollback() }()

	statement := sqliteWorkspaceMemberDirectorySelect + `
        WHERE m.tenant_ref=? AND m.workspace_id=? AND m.lifecycle_state=?`
	arguments := []any{filter.TenantRef, filter.WorkspaceID, filter.MembershipState}
	if cursor.MembershipID != "" {
		anchor, _ := localIdentityUnixNano(mustParseLocalIdentityCursorTime(cursor.UpdatedAt))
		statement += ` AND (m.updated_at_unix_nano < ? OR
            (m.updated_at_unix_nano = ? AND m.membership_id < ?))`
		arguments = append(arguments, anchor, anchor, cursor.MembershipID)
	}
	statement += ` ORDER BY m.updated_at_unix_nano DESC, m.membership_id DESC LIMIT ?`
	arguments = append(arguments, filter.Limit+1)
	rows, err := transaction.QueryContext(identityContext(ctx), statement, arguments...)
	if err != nil {
		return LocalIdentityWorkspaceMemberPage{}, errLocalIdentityStoreUnavailable
	}
	memberships := make([]WorkspaceMembership, 0, filter.Limit+1)
	accounts := make(map[string]UserAccount, filter.Limit+1)
	for rows.Next() {
		membership, account, scanErr := scanSQLiteWorkspaceMemberDirectoryRow(rows)
		if scanErr != nil {
			rows.Close()
			return LocalIdentityWorkspaceMemberPage{}, scanErr
		}
		memberships = append(memberships, membership)
		accounts[account.UserID] = account
	}
	if rows.Close() != nil || rows.Err() != nil {
		return LocalIdentityWorkspaceMemberPage{}, errLocalIdentityStoreUnavailable
	}
	hasNext := len(memberships) > filter.Limit
	if hasNext {
		memberships = memberships[:filter.Limit]
	}
	userIDs := make([]string, 0, len(memberships))
	for _, membership := range memberships {
		userIDs = append(userIDs, membership.UserID)
	}
	assignments, err := readSQLiteLocalIdentityAssignmentsForUsers(
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

func (repository *sqliteLocalIdentityRepository) ReadWorkspaceMember(
	ctx context.Context,
	tenantRef string,
	workspaceID string,
	userID string,
	asOf time.Time,
) (LocalIdentityWorkspaceMemberDetail, error) {
	if repository == nil || repository.database == nil {
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
	transaction, err := repository.database.BeginTx(identityContext(ctx), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return LocalIdentityWorkspaceMemberDetail{}, errLocalIdentityStoreUnavailable
	}
	defer func() { _ = transaction.Rollback() }()
	account, err := scanSQLiteUserAccount(transaction.QueryRowContext(
		identityContext(ctx), sqliteAccountSelect+` WHERE user_id=?`, userID,
	))
	if err != nil {
		return LocalIdentityWorkspaceMemberDetail{}, err
	}
	memberships, err := readSQLiteLocalIdentityMemberships(
		identityContext(ctx), transaction, tenantRef, workspaceID, userID, false,
	)
	if err != nil {
		return LocalIdentityWorkspaceMemberDetail{}, err
	}
	if len(memberships) == 0 {
		return LocalIdentityWorkspaceMemberDetail{}, errLocalIdentityMemberUnavailable
	}
	assignmentMap, err := readSQLiteLocalIdentityAssignmentsForUsers(
		identityContext(ctx), transaction, tenantRef, workspaceID, []string{userID}, false,
	)
	if err != nil {
		return LocalIdentityWorkspaceMemberDetail{}, err
	}
	return buildLocalIdentityWorkspaceMemberDetail(
		account, tenantRef, workspaceID, memberships, assignmentMap[userID], asOf,
	), nil
}

func loadSQLiteLocalIdentityAdministrationScope(
	ctx context.Context,
	query sqliteLocalIdentityAdministrationQuery,
	tenantRef string,
	workspaceID string,
	additionalUserIDs ...string,
) (*memoryLocalIdentityRepository, error) {
	memberships, err := readSQLiteLocalIdentityMemberships(ctx, query, tenantRef, workspaceID, "", true)
	if err != nil {
		return nil, err
	}
	assignmentMap, err := readSQLiteLocalIdentityAssignmentsForUsers(ctx, query, tenantRef, workspaceID, nil, true)
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
	accounts, err := readSQLiteLocalIdentityAccounts(ctx, query, userIDs)
	if err != nil {
		return nil, err
	}
	return newLocalIdentityAdministrationScopeSnapshot(accounts, memberships, assignments)
}

func readSQLiteLocalIdentityMemberships(
	ctx context.Context,
	query sqliteLocalIdentityAdministrationQuery,
	tenantRef string,
	workspaceID string,
	userID string,
	activeOnly bool,
) ([]WorkspaceMembership, error) {
	statement := sqliteMembershipSelect + ` WHERE tenant_ref=? AND workspace_id=?`
	arguments := []any{tenantRef, workspaceID}
	if userID != "" {
		statement += ` AND user_id=?`
		arguments = append(arguments, userID)
	}
	if activeOnly {
		statement += ` AND lifecycle_state='active'`
	}
	statement += ` ORDER BY updated_at_unix_nano DESC, membership_id DESC`
	rows, err := query.QueryContext(ctx, statement, arguments...)
	if err != nil {
		return nil, errLocalIdentityStoreUnavailable
	}
	memberships := make([]WorkspaceMembership, 0)
	for rows.Next() {
		membership, scanErr := scanSQLiteWorkspaceMembership(rows)
		if scanErr != nil {
			rows.Close()
			return nil, scanErr
		}
		memberships = append(memberships, membership)
	}
	if rows.Close() != nil || rows.Err() != nil {
		return nil, errLocalIdentityStoreUnavailable
	}
	return memberships, nil
}

func readSQLiteLocalIdentityAssignmentsForUsers(
	ctx context.Context,
	query sqliteLocalIdentityAdministrationQuery,
	tenantRef string,
	workspaceID string,
	userIDs []string,
	activeOnly bool,
) (map[string][]LocalRoleAssignment, error) {
	if userIDs != nil && len(userIDs) == 0 {
		return map[string][]LocalRoleAssignment{}, nil
	}
	statement := sqliteRoleAssignmentSelect + ` WHERE tenant_ref=? AND (workspace_id='' OR workspace_id=?)`
	arguments := []any{tenantRef, workspaceID}
	if activeOnly {
		statement += ` AND lifecycle_state='active'`
	}
	if len(userIDs) > 0 {
		statement += ` AND user_id IN (` + strings.TrimSuffix(strings.Repeat("?,", len(userIDs)), ",") + `)`
		for _, userID := range userIDs {
			arguments = append(arguments, userID)
		}
	}
	statement += ` ORDER BY assignment_id`
	rows, err := query.QueryContext(ctx, statement, arguments...)
	if err != nil {
		return nil, errLocalIdentityStoreUnavailable
	}
	assignments := make(map[string][]LocalRoleAssignment)
	for rows.Next() {
		assignment, scanErr := scanSQLiteLocalRoleAssignment(rows)
		if scanErr != nil {
			rows.Close()
			return nil, scanErr
		}
		assignments[assignment.UserID] = append(assignments[assignment.UserID], assignment)
	}
	if rows.Close() != nil || rows.Err() != nil {
		return nil, errLocalIdentityStoreUnavailable
	}
	return assignments, nil
}

func readSQLiteLocalIdentityAccounts(
	ctx context.Context,
	query sqliteLocalIdentityAdministrationQuery,
	userIDs []string,
) ([]UserAccount, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	statement := sqliteAccountSelect + ` WHERE user_id IN (` + strings.TrimSuffix(strings.Repeat("?,", len(userIDs)), ",") + `)`
	arguments := make([]any, 0, len(userIDs))
	for _, userID := range userIDs {
		arguments = append(arguments, userID)
	}
	rows, err := query.QueryContext(ctx, statement, arguments...)
	if err != nil {
		return nil, errLocalIdentityStoreUnavailable
	}
	accounts := make([]UserAccount, 0, len(userIDs))
	for rows.Next() {
		account, scanErr := scanSQLiteUserAccount(rows)
		if scanErr != nil {
			rows.Close()
			return nil, scanErr
		}
		accounts = append(accounts, account)
	}
	if rows.Close() != nil || rows.Err() != nil {
		return nil, errLocalIdentityStoreUnavailable
	}
	return accounts, nil
}

const sqliteWorkspaceMemberDirectorySelect = `SELECT
    m.schema_version, m.membership_id, m.user_id, m.tenant_ref, m.workspace_id, m.lifecycle_state,
    m.record_version, m.created_at_unix_nano, m.updated_at_unix_nano, m.expires_at_unix_nano,
    m.revoked_at_unix_nano, m.audit_ref,
    a.schema_version, a.user_id, a.login_identifier, a.normalized_login_identifier, a.display_name,
    a.lifecycle_state, a.record_version, a.created_at_unix_nano, a.updated_at_unix_nano,
    a.disabled_at_unix_nano, a.audit_ref
    FROM local_workspace_memberships m
    JOIN local_user_accounts a ON a.user_id=m.user_id`

func scanSQLiteWorkspaceMemberDirectoryRow(row localIdentitySQLRow) (WorkspaceMembership, UserAccount, error) {
	var membership WorkspaceMembership
	var account UserAccount
	var membershipCreatedAt, membershipUpdatedAt, accountCreatedAt, accountUpdatedAt int64
	var membershipExpiresAt, membershipRevokedAt, accountDisabledAt sql.NullInt64
	err := row.Scan(
		&membership.SchemaVersion, &membership.MembershipID, &membership.UserID, &membership.TenantRef,
		&membership.WorkspaceID, &membership.LifecycleState, &membership.RecordVersion,
		&membershipCreatedAt, &membershipUpdatedAt, &membershipExpiresAt, &membershipRevokedAt, &membership.AuditRef,
		&account.SchemaVersion, &account.UserID, &account.LoginIdentifier, &account.NormalizedLoginIdentifier,
		&account.DisplayName, &account.LifecycleState, &account.RecordVersion, &accountCreatedAt, &accountUpdatedAt,
		&accountDisabledAt, &account.AuditRef,
	)
	if err != nil {
		return WorkspaceMembership{}, UserAccount{}, normalizeSQLiteReadError(err)
	}
	membership.CreatedAt = time.Unix(0, membershipCreatedAt).UTC()
	membership.UpdatedAt = time.Unix(0, membershipUpdatedAt).UTC()
	membership.ExpiresAt = sqliteOptionalTime(membershipExpiresAt)
	membership.RevokedAt = sqliteOptionalTime(membershipRevokedAt)
	account.CreatedAt = time.Unix(0, accountCreatedAt).UTC()
	account.UpdatedAt = time.Unix(0, accountUpdatedAt).UTC()
	account.DisabledAt = sqliteOptionalTime(accountDisabledAt)
	if !validWorkspaceMembership(membership) || !validUserAccount(account) || membership.UserID != account.UserID {
		return WorkspaceMembership{}, UserAccount{}, errLocalIdentityStoreUnavailable
	}
	return membership, account, nil
}

func mustParseLocalIdentityCursorTime(value string) time.Time {
	parsed, _ := time.Parse(time.RFC3339Nano, value)
	return parsed.UTC()
}
