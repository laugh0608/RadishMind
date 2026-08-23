package httpapi

import (
	"crypto/sha256"
	"encoding/binary"
	"slices"
	"strings"
	"time"
)

func localIdentityAdministrationScopeLockKey(tenantRef string, workspaceID string) int64 {
	digest := sha256.Sum256([]byte(strings.TrimSpace(tenantRef) + "\x00" + strings.TrimSpace(workspaceID)))
	return int64(binary.BigEndian.Uint64(digest[:8]))
}

func buildLocalIdentityWorkspaceMemberSummary(
	account UserAccount,
	membership WorkspaceMembership,
	assignments []LocalRoleAssignment,
	asOf time.Time,
) LocalIdentityWorkspaceMemberSummary {
	membershipEffective := account.LifecycleState == localIdentityStateActive &&
		localIdentityMembershipEffective(membership, asOf)
	roleSet := make(map[string]struct{})
	canManage := false
	drift := false
	if membershipEffective {
		for _, assignment := range assignments {
			if !localIdentityRoleAssignmentEffective(assignment, asOf) {
				continue
			}
			roleSet[assignment.RoleKey] = struct{}{}
			definition, known := builtInLocalIdentityRole(assignment.RoleKey)
			if !known || !localIdentityRoleDefinitionMatchesAssignment(definition, assignment) {
				drift = true
			}
			canManage = canManage || localIdentityAssignmentCanManage(assignment, asOf)
		}
	}
	roleKeys := make([]string, 0, len(roleSet))
	for roleKey := range roleSet {
		roleKeys = append(roleKeys, roleKey)
	}
	slices.Sort(roleKeys)
	return LocalIdentityWorkspaceMemberSummary{
		SchemaVersion:            localIdentityWorkspaceMemberSummarySchemaVersion,
		TenantRef:                membership.TenantRef,
		WorkspaceID:              membership.WorkspaceID,
		UserID:                   account.UserID,
		DisplayName:              account.DisplayName,
		AccountLifecycleState:    account.LifecycleState,
		MembershipID:             membership.MembershipID,
		MembershipLifecycleState: membership.LifecycleState,
		MembershipRecordVersion:  membership.RecordVersion,
		MembershipExpiresAt:      cloneTimePointer(membership.ExpiresAt),
		MembershipEffective:      membershipEffective,
		RoleKeys:                 roleKeys,
		CanManageLocalIdentity:   canManage,
		RoleCatalogDrift:         drift,
		UpdatedAt:                membership.UpdatedAt,
	}
}

func buildLocalIdentityWorkspaceMemberDetail(
	account UserAccount,
	tenantRef string,
	workspaceID string,
	memberships []WorkspaceMembership,
	assignments []LocalRoleAssignment,
	asOf time.Time,
) LocalIdentityWorkspaceMemberDetail {
	slices.SortFunc(memberships, func(left, right WorkspaceMembership) int {
		if !left.UpdatedAt.Equal(right.UpdatedAt) {
			return right.UpdatedAt.Compare(left.UpdatedAt)
		}
		return strings.Compare(right.MembershipID, left.MembershipID)
	})
	slices.SortFunc(assignments, func(left, right LocalRoleAssignment) int {
		return strings.Compare(left.AssignmentID, right.AssignmentID)
	})
	hasEffectiveMembership := false
	membershipViews := make([]LocalIdentityWorkspaceMembershipView, 0, len(memberships))
	for _, membership := range memberships {
		effective := account.LifecycleState == localIdentityStateActive && localIdentityMembershipEffective(membership, asOf)
		hasEffectiveMembership = hasEffectiveMembership || effective
		membershipViews = append(membershipViews, localIdentityMembershipView(membership, effective))
	}
	assignmentViews := make([]LocalIdentityWorkspaceRoleAssignmentView, 0, len(assignments))
	canManage := false
	for _, assignment := range assignments {
		effective := hasEffectiveMembership && localIdentityRoleAssignmentEffective(assignment, asOf)
		view := localIdentityRoleAssignmentView(assignment, effective, asOf)
		canManage = canManage || view.CanManageLocalIdentity
		assignmentViews = append(assignmentViews, view)
	}
	return LocalIdentityWorkspaceMemberDetail{
		SchemaVersion:          localIdentityWorkspaceMemberDetailSchemaVersion,
		TenantRef:              tenantRef,
		WorkspaceID:            workspaceID,
		UserID:                 account.UserID,
		DisplayName:            account.DisplayName,
		AccountLifecycleState:  account.LifecycleState,
		AccountRecordVersion:   account.RecordVersion,
		Memberships:            membershipViews,
		RoleAssignments:        assignmentViews,
		CanManageLocalIdentity: canManage,
	}
}

func newLocalIdentityAdministrationScopeSnapshot(
	accounts []UserAccount,
	memberships []WorkspaceMembership,
	assignments []LocalRoleAssignment,
) (*memoryLocalIdentityRepository, error) {
	repository := newMemoryLocalIdentityRepository()
	for _, account := range accounts {
		if !validUserAccount(account) || repository.accounts[account.UserID].UserID != "" {
			return nil, errLocalIdentityStoreUnavailable
		}
		repository.accounts[account.UserID] = cloneUserAccount(account)
		repository.accountByLoginIdentifier[account.NormalizedLoginIdentifier] = account.UserID
	}
	for _, membership := range memberships {
		if !validWorkspaceMembership(membership) || repository.memberships[membership.MembershipID].MembershipID != "" {
			return nil, errLocalIdentityStoreUnavailable
		}
		repository.memberships[membership.MembershipID] = cloneWorkspaceMembership(membership)
		if membership.LifecycleState == localIdentityStateActive {
			key := localMembershipScopeKey(membership.UserID, membership.TenantRef, membership.WorkspaceID)
			if repository.activeMembershipByScope[key] != "" {
				return nil, errLocalIdentityStoreUnavailable
			}
			repository.activeMembershipByScope[key] = membership.MembershipID
		}
	}
	for _, assignment := range assignments {
		if !validLocalRoleAssignment(assignment) || repository.roleAssignments[assignment.AssignmentID].AssignmentID != "" {
			return nil, errLocalIdentityStoreUnavailable
		}
		repository.roleAssignments[assignment.AssignmentID] = cloneLocalRoleAssignment(assignment)
		if assignment.LifecycleState == localIdentityStateActive {
			key := localRoleScopeKey(assignment.UserID, assignment.TenantRef, assignment.WorkspaceID, assignment.RoleKey)
			if repository.activeRoleByScope[key] != "" {
				return nil, errLocalIdentityStoreUnavailable
			}
			repository.activeRoleByScope[key] = assignment.AssignmentID
		}
	}
	return repository, nil
}
