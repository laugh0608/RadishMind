package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"
)

type localIdentityWorkspaceMemberCursor struct {
	SchemaVersion   string `json:"schema_version"`
	MembershipState string `json:"membership_state"`
	Limit           int    `json:"limit"`
	UpdatedAt       string `json:"updated_at"`
	MembershipID    string `json:"membership_id"`
	BindingDigest   string `json:"binding_digest"`
}

func (repository *memoryLocalIdentityRepository) ListWorkspaceMembers(
	_ context.Context,
	query LocalIdentityWorkspaceMemberListQuery,
) (LocalIdentityWorkspaceMemberPage, error) {
	if repository == nil {
		return LocalIdentityWorkspaceMemberPage{}, errLocalIdentityStoreUnavailable
	}
	filter, cursor, err := normalizeLocalIdentityWorkspaceMemberQuery(query)
	if err != nil {
		return LocalIdentityWorkspaceMemberPage{}, err
	}
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	memberships := make([]WorkspaceMembership, 0)
	for _, membership := range repository.memberships {
		if membership.TenantRef != filter.TenantRef || membership.WorkspaceID != filter.WorkspaceID ||
			membership.LifecycleState != filter.MembershipState {
			continue
		}
		if cursor.MembershipID != "" && !localIdentityMembershipComesAfterCursor(membership, cursor) {
			continue
		}
		memberships = append(memberships, cloneWorkspaceMembership(membership))
	}
	slices.SortFunc(memberships, func(left, right WorkspaceMembership) int {
		if !left.UpdatedAt.Equal(right.UpdatedAt) {
			return right.UpdatedAt.Compare(left.UpdatedAt)
		}
		return strings.Compare(right.MembershipID, left.MembershipID)
	})
	hasNext := len(memberships) > filter.Limit
	if hasNext {
		memberships = memberships[:filter.Limit]
	}
	page := LocalIdentityWorkspaceMemberPage{
		Members: make([]LocalIdentityWorkspaceMemberSummary, 0, len(memberships)),
	}
	for _, membership := range memberships {
		summary, buildErr := repository.workspaceMemberSummaryLocked(membership, filter.asOf)
		if buildErr != nil {
			return LocalIdentityWorkspaceMemberPage{}, buildErr
		}
		page.Members = append(page.Members, summary)
	}
	if hasNext && len(memberships) > 0 {
		page.NextCursor, err = encodeLocalIdentityWorkspaceMemberCursor(filter, memberships[len(memberships)-1])
		if err != nil {
			return LocalIdentityWorkspaceMemberPage{}, errLocalIdentityStoreUnavailable
		}
	}
	return page, nil
}

func (repository *memoryLocalIdentityRepository) ReadWorkspaceMember(
	_ context.Context,
	tenantRef string,
	workspaceID string,
	userID string,
	asOf time.Time,
) (LocalIdentityWorkspaceMemberDetail, error) {
	if repository == nil {
		return LocalIdentityWorkspaceMemberDetail{}, errLocalIdentityStoreUnavailable
	}
	tenantRef = strings.TrimSpace(tenantRef)
	workspaceID = strings.TrimSpace(workspaceID)
	userID = strings.TrimSpace(userID)
	asOf = asOf.UTC()
	if !validControlPlaneReadAuthReference(tenantRef, false) ||
		!validControlPlaneReadAuthReference(workspaceID, false) ||
		!localUserIDPattern.MatchString(userID) || asOf.IsZero() {
		return LocalIdentityWorkspaceMemberDetail{}, errLocalIdentityContractMismatch
	}
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	account, exists := repository.accounts[userID]
	if !exists {
		return LocalIdentityWorkspaceMemberDetail{}, errLocalIdentityNotFound
	}
	memberships := make([]WorkspaceMembership, 0)
	for _, membership := range repository.memberships {
		if membership.UserID == userID && membership.TenantRef == tenantRef && membership.WorkspaceID == workspaceID {
			memberships = append(memberships, cloneWorkspaceMembership(membership))
		}
	}
	if len(memberships) == 0 {
		return LocalIdentityWorkspaceMemberDetail{}, errLocalIdentityMemberUnavailable
	}
	slices.SortFunc(memberships, func(left, right WorkspaceMembership) int {
		if !left.UpdatedAt.Equal(right.UpdatedAt) {
			return right.UpdatedAt.Compare(left.UpdatedAt)
		}
		return strings.Compare(right.MembershipID, left.MembershipID)
	})
	hasEffectiveMembership := false
	membershipViews := make([]LocalIdentityWorkspaceMembershipView, 0, len(memberships))
	for _, membership := range memberships {
		effective := account.LifecycleState == localIdentityStateActive && localIdentityMembershipEffective(membership, asOf)
		hasEffectiveMembership = hasEffectiveMembership || effective
		membershipViews = append(membershipViews, localIdentityMembershipView(membership, effective))
	}
	assignments := repository.workspaceRoleAssignmentsLocked(userID, tenantRef, workspaceID)
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
	}, nil
}

func (repository *memoryLocalIdentityRepository) CreateWorkspaceMembershipForAdministration(
	_ context.Context,
	actorUserID string,
	membership WorkspaceMembership,
	now time.Time,
) error {
	if repository == nil {
		return errLocalIdentityStoreUnavailable
	}
	actorUserID = strings.TrimSpace(actorUserID)
	now = now.UTC()
	if !localUserIDPattern.MatchString(actorUserID) || now.IsZero() ||
		!validWorkspaceMembership(membership) || membership.LifecycleState != localIdentityStateActive ||
		!membership.CreatedAt.Equal(now) || !membership.UpdatedAt.Equal(now) {
		return errLocalIdentityContractMismatch
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if err := repository.authorizeAdministrationActorLocked(
		actorUserID,
		membership.TenantRef,
		membership.WorkspaceID,
		localIdentityPermissionMembershipsWrite,
		now,
	); err != nil {
		return err
	}
	account, exists := repository.accounts[membership.UserID]
	if !exists {
		return errLocalIdentityNotFound
	}
	if account.LifecycleState != localIdentityStateActive {
		return errLocalIdentityAccountInactive
	}
	key := localMembershipScopeKey(membership.UserID, membership.TenantRef, membership.WorkspaceID)
	if repository.memberships[membership.MembershipID].MembershipID != "" || repository.activeMembershipByScope[key] != "" {
		return errLocalIdentityIdentifierConflict
	}
	for _, assignment := range repository.roleAssignments {
		if assignment.UserID == membership.UserID && assignment.TenantRef == membership.TenantRef &&
			assignment.WorkspaceID == membership.WorkspaceID && assignment.LifecycleState == localIdentityStateActive {
			return errLocalIdentityMembershipConflict
		}
	}
	repository.memberships[membership.MembershipID] = cloneWorkspaceMembership(membership)
	repository.activeMembershipByScope[key] = membership.MembershipID
	return nil
}

func (repository *memoryLocalIdentityRepository) CreateCatalogRoleAssignment(
	_ context.Context,
	actorUserID string,
	assignment LocalRoleAssignment,
	now time.Time,
) error {
	if repository == nil {
		return errLocalIdentityStoreUnavailable
	}
	actorUserID = strings.TrimSpace(actorUserID)
	now = now.UTC()
	grants, ok := normalizedPermissionGrants(assignment.PermissionGrants)
	assignment.PermissionGrants = grants
	definition, exists := builtInLocalIdentityRole(assignment.RoleKey)
	if !localUserIDPattern.MatchString(actorUserID) || now.IsZero() ||
		!ok || !exists || assignment.WorkspaceID == "" || !validLocalRoleAssignment(assignment) ||
		assignment.LifecycleState != localIdentityStateActive ||
		!assignment.CreatedAt.Equal(now) || !assignment.UpdatedAt.Equal(now) ||
		!localIdentityRoleDefinitionMatchesAssignment(definition, assignment) {
		return errLocalIdentityRoleCatalogMismatch
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if err := repository.authorizeAdministrationActorLocked(
		actorUserID,
		assignment.TenantRef,
		assignment.WorkspaceID,
		localIdentityPermissionRolesAssign,
		now,
	); err != nil {
		return err
	}
	account, exists := repository.accounts[assignment.UserID]
	if !exists {
		return errLocalIdentityNotFound
	}
	if account.LifecycleState != localIdentityStateActive {
		return errLocalIdentityAccountInactive
	}
	membership, exists := repository.memberships[repository.activeMembershipByScope[localMembershipScopeKey(assignment.UserID, assignment.TenantRef, assignment.WorkspaceID)]]
	if !exists || !localIdentityMembershipEffective(membership, assignment.CreatedAt) {
		return errLocalIdentityMemberUnavailable
	}
	if membership.ExpiresAt != nil && (assignment.ExpiresAt == nil || assignment.ExpiresAt.After(*membership.ExpiresAt)) {
		return errLocalIdentityRoleAssignmentConflict
	}
	key := localRoleScopeKey(assignment.UserID, assignment.TenantRef, assignment.WorkspaceID, assignment.RoleKey)
	if repository.roleAssignments[assignment.AssignmentID].AssignmentID != "" || repository.activeRoleByScope[key] != "" {
		return errLocalIdentityIdentifierConflict
	}
	repository.roleAssignments[assignment.AssignmentID] = cloneLocalRoleAssignment(assignment)
	repository.activeRoleByScope[key] = assignment.AssignmentID
	return nil
}

func (repository *memoryLocalIdentityRepository) RevokeCatalogRoleAssignment(
	_ context.Context,
	actorUserID string,
	tenantRef string,
	workspaceID string,
	assignmentID string,
	expectedVersion int,
	revokedAt time.Time,
	auditRef string,
) (LocalRoleAssignment, error) {
	if repository == nil {
		return LocalRoleAssignment{}, errLocalIdentityStoreUnavailable
	}
	tenantRef = strings.TrimSpace(tenantRef)
	workspaceID = strings.TrimSpace(workspaceID)
	assignmentID = strings.TrimSpace(assignmentID)
	actorUserID = strings.TrimSpace(actorUserID)
	revokedAt = revokedAt.UTC()
	if !localUserIDPattern.MatchString(actorUserID) || !validControlPlaneReadAuthReference(tenantRef, false) ||
		!validControlPlaneReadAuthReference(workspaceID, false) ||
		!localRoleAssignmentIDPattern.MatchString(assignmentID) || expectedVersion < 1 ||
		revokedAt.IsZero() || !validAuditRef(auditRef) {
		return LocalRoleAssignment{}, errLocalIdentityContractMismatch
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if err := repository.authorizeAdministrationActorLocked(
		actorUserID,
		tenantRef,
		workspaceID,
		localIdentityPermissionRolesAssign,
		revokedAt,
	); err != nil {
		return LocalRoleAssignment{}, err
	}
	assignment, exists := repository.roleAssignments[assignmentID]
	if !exists || assignment.TenantRef != tenantRef || assignment.WorkspaceID != workspaceID ||
		assignment.RecordVersion != expectedVersion || assignment.LifecycleState != localIdentityStateActive {
		return LocalRoleAssignment{}, errLocalIdentityRoleAssignmentConflict
	}
	if localIdentityAssignmentCanManage(assignment, revokedAt) &&
		repository.activeIdentityAdministratorCountLocked(tenantRef, workspaceID, revokedAt) <= 1 {
		return LocalRoleAssignment{}, errLocalIdentityLastAdminRemoval
	}
	assignment = revokeLocalIdentityRoleAssignment(assignment, revokedAt, auditRef)
	delete(repository.activeRoleByScope, localRoleScopeKey(assignment.UserID, assignment.TenantRef, assignment.WorkspaceID, assignment.RoleKey))
	repository.roleAssignments[assignmentID] = assignment
	return cloneLocalRoleAssignment(assignment), nil
}

func (repository *memoryLocalIdentityRepository) RevokeWorkspaceMembershipAndAssignments(
	_ context.Context,
	tenantRef string,
	workspaceID string,
	membershipID string,
	expectedVersion int,
	actorUserID string,
	revokedAt time.Time,
	auditRef string,
) (LocalIdentityWorkspaceMembershipRevocation, error) {
	if repository == nil {
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
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if err := repository.authorizeAdministrationActorLocked(
		actorUserID,
		tenantRef,
		workspaceID,
		localIdentityPermissionMembershipsWrite,
		revokedAt,
	); err != nil {
		return LocalIdentityWorkspaceMembershipRevocation{}, err
	}
	membership, exists := repository.memberships[membershipID]
	if !exists || membership.TenantRef != tenantRef || membership.WorkspaceID != workspaceID ||
		membership.RecordVersion != expectedVersion || membership.LifecycleState != localIdentityStateActive {
		return LocalIdentityWorkspaceMembershipRevocation{}, errLocalIdentityMembershipConflict
	}
	if membership.UserID == actorUserID {
		return LocalIdentityWorkspaceMembershipRevocation{}, errLocalIdentitySelfMembershipRevoke
	}
	if repository.userCanManageLocalIdentityLocked(membership.UserID, tenantRef, workspaceID, revokedAt) &&
		repository.activeIdentityAdministratorCountLocked(tenantRef, workspaceID, revokedAt) <= 1 {
		return LocalIdentityWorkspaceMembershipRevocation{}, errLocalIdentityLastAdminRemoval
	}
	membership.LifecycleState = localIdentityStateRevoked
	membership.RecordVersion++
	membership.UpdatedAt = revokedAt
	membership.RevokedAt = timePointer(revokedAt)
	membership.AuditRef = strings.TrimSpace(auditRef)
	delete(repository.activeMembershipByScope, localMembershipScopeKey(membership.UserID, tenantRef, workspaceID))
	repository.memberships[membershipID] = membership
	revokedAssignments := make([]LocalRoleAssignment, 0)
	for assignmentID, assignment := range repository.roleAssignments {
		if assignment.UserID != membership.UserID || assignment.TenantRef != tenantRef || assignment.WorkspaceID != workspaceID ||
			assignment.LifecycleState != localIdentityStateActive {
			continue
		}
		assignment = revokeLocalIdentityRoleAssignment(assignment, revokedAt, auditRef)
		delete(repository.activeRoleByScope, localRoleScopeKey(assignment.UserID, tenantRef, workspaceID, assignment.RoleKey))
		repository.roleAssignments[assignmentID] = assignment
		revokedAssignments = append(revokedAssignments, cloneLocalRoleAssignment(assignment))
	}
	slices.SortFunc(revokedAssignments, func(left, right LocalRoleAssignment) int {
		return strings.Compare(left.AssignmentID, right.AssignmentID)
	})
	return LocalIdentityWorkspaceMembershipRevocation{
		Membership:             cloneWorkspaceMembership(membership),
		RevokedRoleAssignments: revokedAssignments,
	}, nil
}

func (repository *memoryLocalIdentityRepository) BootstrapWorkspaceAdministrator(
	_ context.Context,
	bootstrap LocalIdentityWorkspaceAdministratorBootstrap,
	now time.Time,
) (LocalIdentityWorkspaceAdministratorBootstrap, error) {
	if repository == nil {
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
		!exists || !definition.CanManageLocalIdentity ||
		!localIdentityRoleDefinitionMatchesAssignment(definition, assignment) {
		return LocalIdentityWorkspaceAdministratorBootstrap{}, errLocalIdentityContractMismatch
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	account, exists := repository.accounts[membership.UserID]
	if !exists {
		return LocalIdentityWorkspaceAdministratorBootstrap{}, errLocalIdentityNotFound
	}
	if account.LifecycleState != localIdentityStateActive {
		return LocalIdentityWorkspaceAdministratorBootstrap{}, errLocalIdentityAccountInactive
	}
	if repository.activeIdentityAdministratorCountLocked(membership.TenantRef, membership.WorkspaceID, now) > 0 ||
		repository.memberships[membership.MembershipID].MembershipID != "" ||
		repository.roleAssignments[assignment.AssignmentID].AssignmentID != "" ||
		repository.activeMembershipByScope[localMembershipScopeKey(membership.UserID, membership.TenantRef, membership.WorkspaceID)] != "" ||
		repository.activeRoleByScope[localRoleScopeKey(assignment.UserID, assignment.TenantRef, assignment.WorkspaceID, assignment.RoleKey)] != "" {
		return LocalIdentityWorkspaceAdministratorBootstrap{}, errLocalIdentityAdminBootstrapDenied
	}
	repository.memberships[membership.MembershipID] = cloneWorkspaceMembership(membership)
	repository.activeMembershipByScope[localMembershipScopeKey(membership.UserID, membership.TenantRef, membership.WorkspaceID)] = membership.MembershipID
	repository.roleAssignments[assignment.AssignmentID] = cloneLocalRoleAssignment(assignment)
	repository.activeRoleByScope[localRoleScopeKey(assignment.UserID, assignment.TenantRef, assignment.WorkspaceID, assignment.RoleKey)] = assignment.AssignmentID
	return LocalIdentityWorkspaceAdministratorBootstrap{
		Membership:     cloneWorkspaceMembership(membership),
		RoleAssignment: cloneLocalRoleAssignment(assignment),
	}, nil
}

func (repository *memoryLocalIdentityRepository) workspaceMemberSummaryLocked(
	membership WorkspaceMembership,
	asOf time.Time,
) (LocalIdentityWorkspaceMemberSummary, error) {
	account, exists := repository.accounts[membership.UserID]
	if !exists {
		return LocalIdentityWorkspaceMemberSummary{}, errLocalIdentityStoreUnavailable
	}
	membershipEffective := account.LifecycleState == localIdentityStateActive && localIdentityMembershipEffective(membership, asOf)
	roleSet := make(map[string]struct{})
	canManage := false
	drift := false
	if membershipEffective {
		for _, assignment := range repository.workspaceRoleAssignmentsLocked(account.UserID, membership.TenantRef, membership.WorkspaceID) {
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
	}, nil
}

func (repository *memoryLocalIdentityRepository) workspaceRoleAssignmentsLocked(
	userID string,
	tenantRef string,
	workspaceID string,
) []LocalRoleAssignment {
	assignments := make([]LocalRoleAssignment, 0)
	for _, assignment := range repository.roleAssignments {
		if assignment.UserID == userID && assignment.TenantRef == tenantRef &&
			(assignment.WorkspaceID == "" || assignment.WorkspaceID == workspaceID) {
			assignments = append(assignments, cloneLocalRoleAssignment(assignment))
		}
	}
	slices.SortFunc(assignments, func(left, right LocalRoleAssignment) int {
		return strings.Compare(left.AssignmentID, right.AssignmentID)
	})
	return assignments
}

func (repository *memoryLocalIdentityRepository) activeIdentityAdministratorCountLocked(
	tenantRef string,
	workspaceID string,
	now time.Time,
) int {
	administrators := make(map[string]struct{})
	for _, membership := range repository.memberships {
		if membership.TenantRef != tenantRef || membership.WorkspaceID != workspaceID ||
			!localIdentityMembershipEffective(membership, now) {
			continue
		}
		account, exists := repository.accounts[membership.UserID]
		if !exists || account.LifecycleState != localIdentityStateActive {
			continue
		}
		if repository.userCanManageLocalIdentityLocked(membership.UserID, tenantRef, workspaceID, now) {
			administrators[membership.UserID] = struct{}{}
		}
	}
	return len(administrators)
}

func (repository *memoryLocalIdentityRepository) userCanManageLocalIdentityLocked(
	userID string,
	tenantRef string,
	workspaceID string,
	now time.Time,
) bool {
	for _, assignment := range repository.roleAssignments {
		if assignment.UserID == userID && assignment.TenantRef == tenantRef &&
			(assignment.WorkspaceID == "" || assignment.WorkspaceID == workspaceID) &&
			localIdentityAssignmentCanManage(assignment, now) {
			return true
		}
	}
	return false
}

func (repository *memoryLocalIdentityRepository) authorizeAdministrationActorLocked(
	userID string,
	tenantRef string,
	workspaceID string,
	requiredPermission string,
	now time.Time,
) error {
	account, exists := repository.accounts[userID]
	if !exists || account.LifecycleState != localIdentityStateActive {
		return errLocalIdentityAdminUnavailable
	}
	membership, exists := repository.memberships[repository.activeMembershipByScope[localMembershipScopeKey(userID, tenantRef, workspaceID)]]
	if !exists || !localIdentityMembershipEffective(membership, now) {
		return errLocalIdentityMembershipDenied
	}
	for _, assignment := range repository.roleAssignments {
		if assignment.UserID != userID || assignment.TenantRef != tenantRef ||
			(assignment.WorkspaceID != "" && assignment.WorkspaceID != workspaceID) ||
			!localIdentityRoleAssignmentEffective(assignment, now) {
			continue
		}
		if slices.Contains(assignment.PermissionGrants, requiredPermission) {
			return nil
		}
	}
	return errLocalIdentityPermissionDenied
}

func localIdentityMembershipEffective(membership WorkspaceMembership, now time.Time) bool {
	return membership.LifecycleState == localIdentityStateActive &&
		(membership.ExpiresAt == nil || membership.ExpiresAt.After(now.UTC()))
}

func localIdentityRoleAssignmentEffective(assignment LocalRoleAssignment, now time.Time) bool {
	return assignment.LifecycleState == localIdentityStateActive &&
		(assignment.ExpiresAt == nil || assignment.ExpiresAt.After(now.UTC()))
}

func localIdentityMembershipView(
	membership WorkspaceMembership,
	effective bool,
) LocalIdentityWorkspaceMembershipView {
	return LocalIdentityWorkspaceMembershipView{
		SchemaVersion:  localIdentityWorkspaceMembershipViewSchemaVersion,
		MembershipID:   membership.MembershipID,
		LifecycleState: membership.LifecycleState,
		RecordVersion:  membership.RecordVersion,
		CreatedAt:      membership.CreatedAt,
		UpdatedAt:      membership.UpdatedAt,
		ExpiresAt:      cloneTimePointer(membership.ExpiresAt),
		RevokedAt:      cloneTimePointer(membership.RevokedAt),
		Effective:      effective,
	}
}

func localIdentityRoleAssignmentView(
	assignment LocalRoleAssignment,
	effective bool,
	now time.Time,
) LocalIdentityWorkspaceRoleAssignmentView {
	definition, known := builtInLocalIdentityRole(assignment.RoleKey)
	drift := !known || !localIdentityRoleDefinitionMatchesAssignment(definition, assignment)
	scope := "workspace"
	if assignment.WorkspaceID == "" {
		scope = "tenant"
	}
	return LocalIdentityWorkspaceRoleAssignmentView{
		SchemaVersion:          localIdentityWorkspaceRoleAssignmentViewSchemaVersion,
		AssignmentID:           assignment.AssignmentID,
		Scope:                  scope,
		WorkspaceID:            assignment.WorkspaceID,
		RoleKey:                assignment.RoleKey,
		RoleCatalogVersion:     assignment.RoleCatalogVersion,
		RoleDefinitionDigest:   assignment.RoleDefinitionDigest,
		PermissionGrants:       append([]string(nil), assignment.PermissionGrants...),
		LifecycleState:         assignment.LifecycleState,
		RecordVersion:          assignment.RecordVersion,
		CreatedAt:              assignment.CreatedAt,
		UpdatedAt:              assignment.UpdatedAt,
		ExpiresAt:              cloneTimePointer(assignment.ExpiresAt),
		RevokedAt:              cloneTimePointer(assignment.RevokedAt),
		Effective:              effective,
		CatalogDrift:           drift,
		CanManageLocalIdentity: effective && localIdentityAssignmentCanManage(assignment, now),
	}
}

func revokeLocalIdentityRoleAssignment(
	assignment LocalRoleAssignment,
	revokedAt time.Time,
	auditRef string,
) LocalRoleAssignment {
	assignment.LifecycleState = localIdentityStateRevoked
	assignment.RecordVersion++
	assignment.UpdatedAt = revokedAt.UTC()
	assignment.RevokedAt = timePointer(revokedAt)
	assignment.AuditRef = strings.TrimSpace(auditRef)
	return assignment
}

func normalizeLocalIdentityWorkspaceMemberQuery(
	query LocalIdentityWorkspaceMemberListQuery,
) (LocalIdentityWorkspaceMemberListQuery, localIdentityWorkspaceMemberCursor, error) {
	query.TenantRef = strings.TrimSpace(query.TenantRef)
	query.WorkspaceID = strings.TrimSpace(query.WorkspaceID)
	query.MembershipState = strings.TrimSpace(query.MembershipState)
	query.Cursor = strings.TrimSpace(query.Cursor)
	query.asOf = query.asOf.UTC()
	if query.MembershipState == "" {
		query.MembershipState = localIdentityStateActive
	}
	if query.Limit == 0 {
		query.Limit = localIdentityWorkspaceMemberDefaultLimit
	}
	if !validControlPlaneReadAuthReference(query.TenantRef, false) ||
		!validControlPlaneReadAuthReference(query.WorkspaceID, false) ||
		(query.MembershipState != localIdentityStateActive && query.MembershipState != localIdentityStateRevoked) ||
		query.Limit < 1 || query.Limit > localIdentityWorkspaceMemberMaximumLimit || query.asOf.IsZero() {
		return LocalIdentityWorkspaceMemberListQuery{}, localIdentityWorkspaceMemberCursor{}, errLocalIdentityMemberCursorInvalid
	}
	if query.Cursor == "" {
		return query, localIdentityWorkspaceMemberCursor{}, nil
	}
	cursor, err := decodeLocalIdentityWorkspaceMemberCursor(query)
	if err != nil {
		return LocalIdentityWorkspaceMemberListQuery{}, localIdentityWorkspaceMemberCursor{}, errLocalIdentityMemberCursorInvalid
	}
	return query, cursor, nil
}

func encodeLocalIdentityWorkspaceMemberCursor(
	query LocalIdentityWorkspaceMemberListQuery,
	last WorkspaceMembership,
) (string, error) {
	document := localIdentityWorkspaceMemberCursor{
		SchemaVersion:   localIdentityWorkspaceMemberCursorSchemaVersion,
		MembershipState: query.MembershipState,
		Limit:           query.Limit,
		UpdatedAt:       last.UpdatedAt.UTC().Format(time.RFC3339Nano),
		MembershipID:    last.MembershipID,
	}
	document.BindingDigest = localIdentityWorkspaceMemberCursorDigest(query, document)
	payload, err := json.Marshal(document)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeLocalIdentityWorkspaceMemberCursor(
	query LocalIdentityWorkspaceMemberListQuery,
) (localIdentityWorkspaceMemberCursor, error) {
	if len(query.Cursor) > 2048 {
		return localIdentityWorkspaceMemberCursor{}, errLocalIdentityMemberCursorInvalid
	}
	payload, err := base64.RawURLEncoding.DecodeString(query.Cursor)
	if err != nil {
		return localIdentityWorkspaceMemberCursor{}, errLocalIdentityMemberCursorInvalid
	}
	var document localIdentityWorkspaceMemberCursor
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&document) != nil || decoder.Decode(&struct{}{}) != io.EOF ||
		document.SchemaVersion != localIdentityWorkspaceMemberCursorSchemaVersion ||
		document.MembershipState != query.MembershipState || document.Limit != query.Limit ||
		!localMembershipIDPattern.MatchString(document.MembershipID) ||
		document.BindingDigest != localIdentityWorkspaceMemberCursorDigest(query, document) {
		return localIdentityWorkspaceMemberCursor{}, errLocalIdentityMemberCursorInvalid
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, document.UpdatedAt)
	if err != nil || updatedAt.Location() != time.UTC || updatedAt.Format(time.RFC3339Nano) != document.UpdatedAt {
		return localIdentityWorkspaceMemberCursor{}, errLocalIdentityMemberCursorInvalid
	}
	return document, nil
}

func localIdentityWorkspaceMemberCursorDigest(
	query LocalIdentityWorkspaceMemberListQuery,
	document localIdentityWorkspaceMemberCursor,
) string {
	return localIdentityDigest(
		localIdentityWorkspaceMemberCursorSchemaVersion,
		query.TenantRef,
		query.WorkspaceID,
		query.MembershipState,
		strconv.Itoa(query.Limit),
		document.UpdatedAt,
		document.MembershipID,
	)
}

func localIdentityMembershipComesAfterCursor(
	membership WorkspaceMembership,
	cursor localIdentityWorkspaceMemberCursor,
) bool {
	anchor, err := time.Parse(time.RFC3339Nano, cursor.UpdatedAt)
	if err != nil {
		return false
	}
	return membership.UpdatedAt.Before(anchor) ||
		membership.UpdatedAt.Equal(anchor) && membership.MembershipID < cursor.MembershipID
}
