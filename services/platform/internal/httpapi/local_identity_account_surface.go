package httpapi

import (
	"context"
	"slices"
	"strings"
)

// LocalIdentityAccountAccessProfile is the single repository-owned projection
// used by the authenticated account surface and the S7 User/Role read model.
// Secret credentials and upstream identity claims are deliberately excluded.
type LocalIdentityAccountAccessProfile struct {
	Account                  UserAccount
	ExternalIdentities       []ExternalIdentityBinding
	RoleAssignments          []LocalRoleAssignment
	WorkspaceMemberships     []WorkspaceMembership
	HasActiveLocalCredential bool
}

func (repository *memoryLocalIdentityRepository) ReadAccountAccessProfile(
	_ context.Context,
	userID string,
) (LocalIdentityAccountAccessProfile, error) {
	if repository == nil {
		return LocalIdentityAccountAccessProfile{}, errLocalIdentityStoreUnavailable
	}
	userID = strings.TrimSpace(userID)
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	account, exists := repository.accounts[userID]
	if !exists {
		return LocalIdentityAccountAccessProfile{}, errLocalIdentityNotFound
	}
	profile := LocalIdentityAccountAccessProfile{
		Account:                  cloneUserAccount(account),
		ExternalIdentities:       make([]ExternalIdentityBinding, 0),
		RoleAssignments:          make([]LocalRoleAssignment, 0),
		WorkspaceMemberships:     make([]WorkspaceMembership, 0),
		HasActiveLocalCredential: repository.activeCredentialByUser[userID] != "",
	}
	for _, binding := range repository.externalBindings {
		if binding.UserID == userID {
			profile.ExternalIdentities = append(profile.ExternalIdentities, cloneExternalIdentityBinding(binding))
		}
	}
	for _, assignment := range repository.roleAssignments {
		if assignment.UserID == userID {
			profile.RoleAssignments = append(profile.RoleAssignments, cloneLocalRoleAssignment(assignment))
		}
	}
	for _, membership := range repository.memberships {
		if membership.UserID == userID {
			profile.WorkspaceMemberships = append(profile.WorkspaceMemberships, cloneWorkspaceMembership(membership))
		}
	}
	sortLocalIdentityAccountAccessProfile(&profile)
	return profile, nil
}

func (repository *sqliteLocalIdentityRepository) ReadAccountAccessProfile(
	ctx context.Context,
	userID string,
) (LocalIdentityAccountAccessProfile, error) {
	if repository == nil || repository.database == nil {
		return LocalIdentityAccountAccessProfile{}, errLocalIdentityStoreUnavailable
	}
	userID = strings.TrimSpace(userID)
	transaction, err := repository.database.BeginTx(identityContext(ctx), nil)
	if err != nil {
		return LocalIdentityAccountAccessProfile{}, errLocalIdentityStoreUnavailable
	}
	defer func() { _ = transaction.Rollback() }()
	profile := LocalIdentityAccountAccessProfile{
		ExternalIdentities:   make([]ExternalIdentityBinding, 0),
		RoleAssignments:      make([]LocalRoleAssignment, 0),
		WorkspaceMemberships: make([]WorkspaceMembership, 0),
	}
	profile.Account, err = scanSQLiteUserAccount(transaction.QueryRowContext(
		identityContext(ctx), sqliteAccountSelect+` WHERE user_id=?`, userID,
	))
	if err != nil {
		return LocalIdentityAccountAccessProfile{}, err
	}
	var activeCredentialCount int
	if err := transaction.QueryRowContext(identityContext(ctx), `SELECT COUNT(*) FROM local_credentials
		WHERE user_id=? AND lifecycle_state='active'`, userID).Scan(&activeCredentialCount); err != nil {
		return LocalIdentityAccountAccessProfile{}, errLocalIdentityStoreUnavailable
	}
	profile.HasActiveLocalCredential = activeCredentialCount > 0
	bindingRows, err := transaction.QueryContext(identityContext(ctx), sqliteBindingSelect+`
		WHERE user_id=? ORDER BY binding_id`, userID)
	if err != nil {
		return LocalIdentityAccountAccessProfile{}, errLocalIdentityStoreUnavailable
	}
	for bindingRows.Next() {
		binding, scanErr := scanSQLiteExternalIdentityBinding(bindingRows)
		if scanErr != nil {
			_ = bindingRows.Close()
			return LocalIdentityAccountAccessProfile{}, scanErr
		}
		profile.ExternalIdentities = append(profile.ExternalIdentities, binding)
	}
	if rowsErr := bindingRows.Err(); rowsErr != nil {
		_ = bindingRows.Close()
		return LocalIdentityAccountAccessProfile{}, errLocalIdentityStoreUnavailable
	}
	_ = bindingRows.Close()
	assignmentRows, err := transaction.QueryContext(identityContext(ctx), sqliteRoleAssignmentSelect+`
		WHERE user_id=? ORDER BY assignment_id`, userID)
	if err != nil {
		return LocalIdentityAccountAccessProfile{}, errLocalIdentityStoreUnavailable
	}
	for assignmentRows.Next() {
		assignment, scanErr := scanSQLiteLocalRoleAssignment(assignmentRows)
		if scanErr != nil {
			_ = assignmentRows.Close()
			return LocalIdentityAccountAccessProfile{}, scanErr
		}
		profile.RoleAssignments = append(profile.RoleAssignments, assignment)
	}
	if rowsErr := assignmentRows.Err(); rowsErr != nil {
		_ = assignmentRows.Close()
		return LocalIdentityAccountAccessProfile{}, errLocalIdentityStoreUnavailable
	}
	_ = assignmentRows.Close()
	membershipRows, err := transaction.QueryContext(identityContext(ctx), sqliteMembershipSelect+`
		WHERE user_id=? ORDER BY membership_id`, userID)
	if err != nil {
		return LocalIdentityAccountAccessProfile{}, errLocalIdentityStoreUnavailable
	}
	for membershipRows.Next() {
		membership, scanErr := scanSQLiteWorkspaceMembership(membershipRows)
		if scanErr != nil {
			_ = membershipRows.Close()
			return LocalIdentityAccountAccessProfile{}, scanErr
		}
		profile.WorkspaceMemberships = append(profile.WorkspaceMemberships, membership)
	}
	if rowsErr := membershipRows.Err(); rowsErr != nil {
		_ = membershipRows.Close()
		return LocalIdentityAccountAccessProfile{}, errLocalIdentityStoreUnavailable
	}
	_ = membershipRows.Close()
	if err := transaction.Commit(); err != nil {
		return LocalIdentityAccountAccessProfile{}, errLocalIdentityStoreUnavailable
	}
	return profile, nil
}

func (repository *postgresLocalIdentityRepository) ReadAccountAccessProfile(
	ctx context.Context,
	userID string,
) (LocalIdentityAccountAccessProfile, error) {
	if repository == nil || repository.pool == nil {
		return LocalIdentityAccountAccessProfile{}, errLocalIdentityStoreUnavailable
	}
	userID = strings.TrimSpace(userID)
	transaction, err := repository.pool.Begin(identityContext(ctx))
	if err != nil {
		return LocalIdentityAccountAccessProfile{}, errLocalIdentityStoreUnavailable
	}
	defer func() { _ = transaction.Rollback(identityContext(ctx)) }()
	profile := LocalIdentityAccountAccessProfile{
		ExternalIdentities:   make([]ExternalIdentityBinding, 0),
		RoleAssignments:      make([]LocalRoleAssignment, 0),
		WorkspaceMemberships: make([]WorkspaceMembership, 0),
	}
	profile.Account, err = scanPostgresUserAccount(transaction.QueryRow(
		identityContext(ctx), postgresAccountSelect+` WHERE user_id=$1`, userID,
	))
	if err != nil {
		return LocalIdentityAccountAccessProfile{}, err
	}
	var activeCredentialCount int
	if err := transaction.QueryRow(identityContext(ctx), `SELECT COUNT(*) FROM local_credentials
		WHERE user_id=$1 AND lifecycle_state='active'`, userID).Scan(&activeCredentialCount); err != nil {
		return LocalIdentityAccountAccessProfile{}, errLocalIdentityStoreUnavailable
	}
	profile.HasActiveLocalCredential = activeCredentialCount > 0
	bindingRows, err := transaction.Query(identityContext(ctx), postgresBindingSelect+`
		WHERE user_id=$1 ORDER BY binding_id`, userID)
	if err != nil {
		return LocalIdentityAccountAccessProfile{}, errLocalIdentityStoreUnavailable
	}
	for bindingRows.Next() {
		binding, scanErr := scanPostgresExternalIdentityBinding(bindingRows)
		if scanErr != nil {
			bindingRows.Close()
			return LocalIdentityAccountAccessProfile{}, scanErr
		}
		profile.ExternalIdentities = append(profile.ExternalIdentities, binding)
	}
	if rowsErr := bindingRows.Err(); rowsErr != nil {
		bindingRows.Close()
		return LocalIdentityAccountAccessProfile{}, errLocalIdentityStoreUnavailable
	}
	bindingRows.Close()
	assignmentRows, err := transaction.Query(identityContext(ctx), postgresRoleAssignmentSelect+`
		WHERE user_id=$1 ORDER BY assignment_id`, userID)
	if err != nil {
		return LocalIdentityAccountAccessProfile{}, errLocalIdentityStoreUnavailable
	}
	for assignmentRows.Next() {
		assignment, scanErr := scanPostgresLocalRoleAssignment(assignmentRows)
		if scanErr != nil {
			assignmentRows.Close()
			return LocalIdentityAccountAccessProfile{}, scanErr
		}
		profile.RoleAssignments = append(profile.RoleAssignments, assignment)
	}
	if rowsErr := assignmentRows.Err(); rowsErr != nil {
		assignmentRows.Close()
		return LocalIdentityAccountAccessProfile{}, errLocalIdentityStoreUnavailable
	}
	assignmentRows.Close()
	membershipRows, err := transaction.Query(identityContext(ctx), postgresMembershipSelect+`
		WHERE user_id=$1 ORDER BY membership_id`, userID)
	if err != nil {
		return LocalIdentityAccountAccessProfile{}, errLocalIdentityStoreUnavailable
	}
	for membershipRows.Next() {
		membership, scanErr := scanPostgresWorkspaceMembership(membershipRows)
		if scanErr != nil {
			membershipRows.Close()
			return LocalIdentityAccountAccessProfile{}, scanErr
		}
		profile.WorkspaceMemberships = append(profile.WorkspaceMemberships, membership)
	}
	if rowsErr := membershipRows.Err(); rowsErr != nil {
		membershipRows.Close()
		return LocalIdentityAccountAccessProfile{}, errLocalIdentityStoreUnavailable
	}
	membershipRows.Close()
	if err := transaction.Commit(identityContext(ctx)); err != nil {
		return LocalIdentityAccountAccessProfile{}, errLocalIdentityStoreUnavailable
	}
	return profile, nil
}

func sortLocalIdentityAccountAccessProfile(profile *LocalIdentityAccountAccessProfile) {
	if profile == nil {
		return
	}
	slices.SortFunc(profile.ExternalIdentities, func(left ExternalIdentityBinding, right ExternalIdentityBinding) int {
		return strings.Compare(left.BindingID, right.BindingID)
	})
	slices.SortFunc(profile.RoleAssignments, func(left LocalRoleAssignment, right LocalRoleAssignment) int {
		return strings.Compare(left.AssignmentID, right.AssignmentID)
	})
	slices.SortFunc(profile.WorkspaceMemberships, func(left WorkspaceMembership, right WorkspaceMembership) int {
		return strings.Compare(left.MembershipID, right.MembershipID)
	})
}
