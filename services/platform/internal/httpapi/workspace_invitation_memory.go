package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"slices"
	"strconv"
	"strings"
	"time"
)

type memoryWorkspaceInvitation struct {
	invitation   WorkspaceInvitation
	secretDigest workspaceInvitationSecretDigest
}

func (repository *memoryLocalIdentityRepository) CreateWorkspaceInvitation(
	_ context.Context,
	actorUserID string,
	invitation WorkspaceInvitation,
	secretDigest workspaceInvitationSecretDigest,
	now time.Time,
) error {
	if repository == nil {
		return errWorkspaceInvitationAdminUnavailable
	}
	actorUserID = strings.TrimSpace(actorUserID)
	now = now.UTC()
	if !localUserIDPattern.MatchString(actorUserID) || now.IsZero() || !invitation.CreatedAt.Equal(now) ||
		!validWorkspaceInvitation(invitation) || !validWorkspaceInvitationSecretDigest(secretDigest) {
		return errWorkspaceInvitationAdminUnavailable
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.workspaceInvitations == nil {
		return errWorkspaceInvitationAdminUnavailable
	}
	if err := repository.authorizeWorkspaceInvitationAdministratorLocked(
		actorUserID,
		invitation.TenantRef,
		invitation.WorkspaceID,
		now,
		localIdentityPermissionMembershipsWrite,
		localIdentityPermissionRolesAssign,
	); err != nil {
		return err
	}
	if _, exists := repository.workspaceInvitations[invitation.InvitationID]; exists {
		return errWorkspaceInvitationAdminUnavailable
	}
	repository.workspaceInvitations[invitation.InvitationID] = memoryWorkspaceInvitation{
		invitation: cloneWorkspaceInvitation(invitation), secretDigest: secretDigest,
	}
	return nil
}

func (repository *memoryLocalIdentityRepository) ListWorkspaceInvitations(
	_ context.Context,
	actorUserID string,
	query WorkspaceInvitationListQuery,
) (WorkspaceInvitationPage, error) {
	if repository == nil {
		return WorkspaceInvitationPage{}, errWorkspaceInvitationAdminUnavailable
	}
	filter, cursor, err := normalizeWorkspaceInvitationListQuery(query)
	if err != nil {
		return WorkspaceInvitationPage{}, err
	}
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.workspaceInvitations == nil {
		return WorkspaceInvitationPage{}, errWorkspaceInvitationAdminUnavailable
	}
	if err := repository.authorizeWorkspaceInvitationAdministratorLocked(
		strings.TrimSpace(actorUserID),
		filter.TenantRef,
		filter.WorkspaceID,
		filter.authorizedAt,
		localIdentityPermissionMembersRead,
		localIdentityPermissionRolesRead,
	); err != nil {
		return WorkspaceInvitationPage{}, err
	}
	invitations := make([]WorkspaceInvitation, 0)
	for _, stored := range repository.workspaceInvitations {
		if !validWorkspaceInvitation(stored.invitation) || !validWorkspaceInvitationSecretDigest(stored.secretDigest) {
			return WorkspaceInvitationPage{}, errWorkspaceInvitationAdminUnavailable
		}
		invitation := stored.invitation
		if invitation.TenantRef != filter.TenantRef || invitation.WorkspaceID != filter.WorkspaceID ||
			workspaceInvitationEffectiveState(invitation, filter.asOf) != filter.EffectiveState {
			continue
		}
		if cursor.InvitationID != "" && !workspaceInvitationComesAfterCursor(invitation, cursor) {
			continue
		}
		invitations = append(invitations, cloneWorkspaceInvitation(invitation))
	}
	slices.SortFunc(invitations, func(left, right WorkspaceInvitation) int {
		if !left.UpdatedAt.Equal(right.UpdatedAt) {
			return right.UpdatedAt.Compare(left.UpdatedAt)
		}
		return strings.Compare(right.InvitationID, left.InvitationID)
	})
	hasNext := len(invitations) > filter.Limit
	if hasNext {
		invitations = invitations[:filter.Limit]
	}
	page := WorkspaceInvitationPage{
		SchemaVersion: workspaceInvitationPageSchemaVersion,
		AsOf:          filter.asOf,
		Invitations:   make([]WorkspaceInvitation, 0, len(invitations)),
	}
	for _, invitation := range invitations {
		page.Invitations = append(page.Invitations, projectWorkspaceInvitation(invitation, filter.asOf))
	}
	if hasNext && len(invitations) > 0 {
		page.NextCursor, err = encodeWorkspaceInvitationCursor(filter, invitations[len(invitations)-1])
		if err != nil {
			return WorkspaceInvitationPage{}, errWorkspaceInvitationAdminUnavailable
		}
	}
	return page, nil
}

func (repository *memoryLocalIdentityRepository) RevokeWorkspaceInvitation(
	_ context.Context,
	actorUserID string,
	input WorkspaceInvitationRevokeInput,
	now time.Time,
) (WorkspaceInvitation, error) {
	if repository == nil {
		return WorkspaceInvitation{}, errWorkspaceInvitationAdminUnavailable
	}
	actorUserID = strings.TrimSpace(actorUserID)
	now = now.UTC()
	if !localUserIDPattern.MatchString(actorUserID) || now.IsZero() ||
		!validControlPlaneReadAuthReference(input.TenantRef, false) ||
		!validControlPlaneReadAuthReference(input.WorkspaceID, false) ||
		!workspaceInvitationIDPattern.MatchString(input.InvitationID) || input.ExpectedVersion < 1 ||
		!validAuditRef(input.RequestRef) || !validAuditRef(input.AuditRef) {
		return WorkspaceInvitation{}, errWorkspaceInvitationTransitionInvalid
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.workspaceInvitations == nil {
		return WorkspaceInvitation{}, errWorkspaceInvitationAdminUnavailable
	}
	if err := repository.authorizeWorkspaceInvitationAdministratorLocked(
		actorUserID,
		input.TenantRef,
		input.WorkspaceID,
		now,
		localIdentityPermissionMembershipsWrite,
		localIdentityPermissionRolesAssign,
	); err != nil {
		return WorkspaceInvitation{}, err
	}
	stored, exists := repository.workspaceInvitations[input.InvitationID]
	if !exists || stored.invitation.TenantRef != input.TenantRef || stored.invitation.WorkspaceID != input.WorkspaceID {
		return WorkspaceInvitation{}, errWorkspaceInvitationTransitionInvalid
	}
	if !validWorkspaceInvitation(stored.invitation) || !validWorkspaceInvitationSecretDigest(stored.secretDigest) {
		return WorkspaceInvitation{}, errWorkspaceInvitationAdminUnavailable
	}
	if stored.invitation.RecordVersion != input.ExpectedVersion {
		return WorkspaceInvitation{}, errWorkspaceInvitationVersionConflict
	}
	if workspaceInvitationEffectiveState(stored.invitation, now) != workspaceInvitationEffectivePending {
		return WorkspaceInvitation{}, errWorkspaceInvitationTransitionInvalid
	}
	actorRef, err := LocalUserActorRef(actorUserID)
	if err != nil {
		return WorkspaceInvitation{}, errWorkspaceInvitationAdminUnavailable
	}
	invitation := stored.invitation
	invitation.RecordVersion++
	invitation.LifecycleState = workspaceInvitationLifecycleRevoked
	invitation.UpdatedAt = now
	invitation.UpdatedRequestRef = input.RequestRef
	invitation.UpdatedAuditRef = input.AuditRef
	invitation.RevokedAt = timePointer(now)
	invitation.RevokedByActorRef = actorRef
	if !validWorkspaceInvitation(invitation) {
		return WorkspaceInvitation{}, errWorkspaceInvitationAdminUnavailable
	}
	stored.invitation = invitation
	repository.workspaceInvitations[input.InvitationID] = stored
	return cloneWorkspaceInvitation(invitation), nil
}

func (repository *memoryLocalIdentityRepository) PreviewWorkspaceInvitation(
	_ context.Context,
	claimantUserID string,
	claimantTenantRef string,
	invitationID string,
	secretDigest workspaceInvitationSecretDigest,
	now time.Time,
) (WorkspaceInvitation, error) {
	if repository == nil {
		return WorkspaceInvitation{}, errWorkspaceInvitationStoreUnavailable
	}
	claimantUserID = strings.TrimSpace(claimantUserID)
	claimantTenantRef = strings.TrimSpace(claimantTenantRef)
	invitationID = strings.TrimSpace(invitationID)
	now = now.UTC()
	if !localUserIDPattern.MatchString(claimantUserID) ||
		!validControlPlaneReadAuthReference(claimantTenantRef, false) ||
		!workspaceInvitationIDPattern.MatchString(invitationID) || now.IsZero() ||
		!validWorkspaceInvitationSecretDigest(secretDigest) {
		return WorkspaceInvitation{}, errWorkspaceInvitationInvalid
	}
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	stored, exists := repository.workspaceInvitations[invitationID]
	expectedDigest := workspaceInvitationDummySecretDigest
	if exists {
		expectedDigest = stored.secretDigest
	}
	secretMatches := workspaceInvitationSecretMatches(secretDigest, expectedDigest)
	if !exists || !secretMatches {
		return WorkspaceInvitation{}, errWorkspaceInvitationInvalid
	}
	if !validWorkspaceInvitation(stored.invitation) || !validWorkspaceInvitationSecretDigest(stored.secretDigest) {
		return WorkspaceInvitation{}, errWorkspaceInvitationStoreUnavailable
	}
	if err := repository.validateWorkspaceInvitationClaimantLocked(claimantUserID, claimantTenantRef, stored.invitation); err != nil {
		return WorkspaceInvitation{}, err
	}
	if workspaceInvitationEffectiveState(stored.invitation, now) != workspaceInvitationEffectivePending ||
		!workspaceInvitationMatchesCurrentRole(stored.invitation) {
		return WorkspaceInvitation{}, errWorkspaceInvitationNotClaimable
	}
	return cloneWorkspaceInvitation(stored.invitation), nil
}

func (repository *memoryLocalIdentityRepository) ClaimWorkspaceInvitation(
	_ context.Context,
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
	if repository == nil {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationStoreUnavailable
	}
	claimantUserID = strings.TrimSpace(claimantUserID)
	claimantTenantRef = strings.TrimSpace(claimantTenantRef)
	invitationID = strings.TrimSpace(invitationID)
	requestRef = strings.TrimSpace(requestRef)
	auditRef = strings.TrimSpace(auditRef)
	now = now.UTC()
	if !localUserIDPattern.MatchString(claimantUserID) || !validControlPlaneReadAuthReference(claimantTenantRef, false) ||
		!workspaceInvitationIDPattern.MatchString(invitationID) || !validWorkspaceInvitationSecretDigest(secretDigest) ||
		expectedVersion < 1 || !localMembershipIDPattern.MatchString(membershipID) ||
		!localRoleAssignmentIDPattern.MatchString(assignmentID) || now.IsZero() ||
		!validAuditRef(requestRef) || !validAuditRef(auditRef) {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationStoreUnavailable
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	stored, exists := repository.workspaceInvitations[invitationID]
	expectedDigest := workspaceInvitationDummySecretDigest
	if exists {
		expectedDigest = stored.secretDigest
	}
	secretMatches := workspaceInvitationSecretMatches(secretDigest, expectedDigest)
	if !exists || !secretMatches {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationInvalid
	}
	if !validWorkspaceInvitation(stored.invitation) || !validWorkspaceInvitationSecretDigest(stored.secretDigest) {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationStoreUnavailable
	}
	if err := repository.validateWorkspaceInvitationClaimantLocked(claimantUserID, claimantTenantRef, stored.invitation); err != nil {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, err
	}
	if workspaceInvitationEffectiveState(stored.invitation, now) != workspaceInvitationEffectivePending ||
		!workspaceInvitationMatchesCurrentRole(stored.invitation) {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationNotClaimable
	}
	if stored.invitation.RecordVersion != expectedVersion {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationVersionConflict
	}
	if err := repository.validateWorkspaceInvitationMembershipInvariantLocked(
		claimantUserID,
		stored.invitation.TenantRef,
		stored.invitation.WorkspaceID,
	); err != nil {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, err
	}
	if repository.memberships[membershipID].MembershipID != "" ||
		repository.roleAssignments[assignmentID].AssignmentID != "" {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationStoreUnavailable
	}
	definition, exists := builtInLocalIdentityRole(stored.invitation.RoleKey)
	if !exists || !workspaceInvitationRoleEligible(definition) ||
		definition.CatalogVersion != stored.invitation.RoleCatalogVersion ||
		definition.DefinitionDigest != stored.invitation.RoleDefinitionDigest {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationNotClaimable
	}
	membership := WorkspaceMembership{
		SchemaVersion: localIdentitySchemaVersion, MembershipID: membershipID, UserID: claimantUserID,
		TenantRef: stored.invitation.TenantRef, WorkspaceID: stored.invitation.WorkspaceID,
		LifecycleState: localIdentityStateActive, RecordVersion: 1, CreatedAt: now, UpdatedAt: now,
		AuditRef: auditRef,
	}
	assignment := LocalRoleAssignment{
		SchemaVersion: localIdentitySchemaVersion, AssignmentID: assignmentID, UserID: claimantUserID,
		TenantRef: stored.invitation.TenantRef, WorkspaceID: stored.invitation.WorkspaceID,
		RoleKey: definition.RoleKey, RoleCatalogVersion: definition.CatalogVersion,
		RoleDefinitionDigest: definition.DefinitionDigest,
		PermissionGrants:     append([]string(nil), definition.PermissionGrants...),
		LifecycleState:       localIdentityStateActive, RecordVersion: 1, CreatedAt: now, UpdatedAt: now,
		AuditRef: auditRef,
	}
	invitation := stored.invitation
	invitation.RecordVersion++
	invitation.LifecycleState = workspaceInvitationLifecycleClaimed
	invitation.UpdatedAt = now
	invitation.UpdatedRequestRef = requestRef
	invitation.UpdatedAuditRef = auditRef
	invitation.ClaimedAt = timePointer(now)
	invitation.ClaimedByUserID = claimantUserID
	invitation.MembershipID = membershipID
	invitation.AssignmentID = assignmentID
	if !validWorkspaceMembership(membership) || !validLocalRoleAssignment(assignment) ||
		!localIdentityRoleDefinitionMatchesAssignment(definition, assignment) || !validWorkspaceInvitation(invitation) {
		return WorkspaceInvitation{}, WorkspaceMembership{}, LocalRoleAssignment{}, errWorkspaceInvitationStoreUnavailable
	}
	membershipKey := localMembershipScopeKey(claimantUserID, invitation.TenantRef, invitation.WorkspaceID)
	roleKey := localRoleScopeKey(claimantUserID, invitation.TenantRef, invitation.WorkspaceID, assignment.RoleKey)
	repository.memberships[membershipID] = cloneWorkspaceMembership(membership)
	repository.activeMembershipByScope[membershipKey] = membershipID
	repository.roleAssignments[assignmentID] = cloneLocalRoleAssignment(assignment)
	repository.activeRoleByScope[roleKey] = assignmentID
	stored.invitation = invitation
	repository.workspaceInvitations[invitationID] = stored
	return cloneWorkspaceInvitation(invitation), cloneWorkspaceMembership(membership), cloneLocalRoleAssignment(assignment), nil
}

func (repository *memoryLocalIdentityRepository) authorizeWorkspaceInvitationAdministratorLocked(
	actorUserID string,
	tenantRef string,
	workspaceID string,
	now time.Time,
	permissions ...string,
) error {
	for _, permission := range permissions {
		if err := repository.authorizeAdministrationActorLocked(actorUserID, tenantRef, workspaceID, permission, now); err != nil {
			return err
		}
	}
	return nil
}

func (repository *memoryLocalIdentityRepository) validateWorkspaceInvitationClaimantLocked(
	claimantUserID string,
	claimantTenantRef string,
	invitation WorkspaceInvitation,
) error {
	account, exists := repository.accounts[claimantUserID]
	if !exists || !validUserAccount(account) || account.LifecycleState != localIdentityStateActive ||
		claimantTenantRef != invitation.TenantRef {
		return errWorkspaceInvitationAccountIneligible
	}
	return nil
}

func (repository *memoryLocalIdentityRepository) validateWorkspaceInvitationMembershipInvariantLocked(
	userID string,
	tenantRef string,
	workspaceID string,
) error {
	membershipKey := localMembershipScopeKey(userID, tenantRef, workspaceID)
	activeMembershipID := ""
	for _, membership := range repository.memberships {
		if membership.UserID != userID || membership.TenantRef != tenantRef || membership.WorkspaceID != workspaceID {
			continue
		}
		if !validWorkspaceMembership(membership) {
			return errWorkspaceInvitationStoreUnavailable
		}
		if membership.LifecycleState == localIdentityStateActive {
			if activeMembershipID != "" {
				return errWorkspaceInvitationStoreUnavailable
			}
			activeMembershipID = membership.MembershipID
		}
	}
	if activeMembershipID != "" {
		if repository.activeMembershipByScope[membershipKey] != activeMembershipID {
			return errWorkspaceInvitationStoreUnavailable
		}
		return errWorkspaceInvitationMembershipConflict
	}
	if repository.activeMembershipByScope[membershipKey] != "" {
		return errWorkspaceInvitationStoreUnavailable
	}
	for _, assignment := range repository.roleAssignments {
		if assignment.UserID != userID || assignment.TenantRef != tenantRef || assignment.WorkspaceID != workspaceID {
			continue
		}
		if !validLocalRoleAssignment(assignment) {
			return errWorkspaceInvitationStoreUnavailable
		}
		if assignment.LifecycleState == localIdentityStateActive {
			return errWorkspaceInvitationMembershipConflict
		}
	}
	for key, assignmentID := range repository.activeRoleByScope {
		parts := strings.Split(key, "\x00")
		if len(parts) != 4 || parts[0] != userID || parts[1] != tenantRef || parts[2] != workspaceID {
			continue
		}
		assignment, exists := repository.roleAssignments[assignmentID]
		if !exists || assignment.LifecycleState != localIdentityStateActive ||
			key != localRoleScopeKey(assignment.UserID, assignment.TenantRef, assignment.WorkspaceID, assignment.RoleKey) {
			return errWorkspaceInvitationStoreUnavailable
		}
		return errWorkspaceInvitationMembershipConflict
	}
	return nil
}

func workspaceInvitationMatchesCurrentRole(invitation WorkspaceInvitation) bool {
	definition, exists := builtInLocalIdentityRole(invitation.RoleKey)
	return exists && workspaceInvitationRoleEligible(definition) &&
		definition.CatalogVersion == invitation.RoleCatalogVersion &&
		definition.DefinitionDigest == invitation.RoleDefinitionDigest
}

func normalizeWorkspaceInvitationListQuery(
	query WorkspaceInvitationListQuery,
) (WorkspaceInvitationListQuery, workspaceInvitationCursor, error) {
	query.TenantRef = strings.TrimSpace(query.TenantRef)
	query.WorkspaceID = strings.TrimSpace(query.WorkspaceID)
	query.EffectiveState = strings.TrimSpace(query.EffectiveState)
	query.Cursor = strings.TrimSpace(query.Cursor)
	query.asOf = query.asOf.UTC()
	query.authorizedAt = query.authorizedAt.UTC()
	if query.authorizedAt.IsZero() {
		query.authorizedAt = query.asOf
	}
	if query.EffectiveState == "" {
		query.EffectiveState = workspaceInvitationEffectivePending
	}
	if query.Limit == 0 {
		query.Limit = workspaceInvitationDefaultListLimit
	}
	if !validControlPlaneReadAuthReference(query.TenantRef, false) ||
		!validControlPlaneReadAuthReference(query.WorkspaceID, false) || query.asOf.IsZero() ||
		query.Limit < 1 || query.Limit > workspaceInvitationMaximumListLimit ||
		!validWorkspaceInvitationEffectiveState(query.EffectiveState) {
		return WorkspaceInvitationListQuery{}, workspaceInvitationCursor{}, errWorkspaceInvitationCursorInvalid
	}
	if query.Cursor == "" {
		return query, workspaceInvitationCursor{}, nil
	}
	cursor, err := decodeWorkspaceInvitationCursor(query.Cursor)
	if err != nil {
		return WorkspaceInvitationListQuery{}, workspaceInvitationCursor{}, errWorkspaceInvitationCursorInvalid
	}
	expectedDigest := workspaceInvitationCursorBindingDigest(cursor)
	asOf, asOfErr := time.Parse(time.RFC3339Nano, cursor.AsOf)
	updatedAt, updatedAtErr := time.Parse(time.RFC3339Nano, cursor.UpdatedAt)
	if cursor.SchemaVersion != workspaceInvitationCursorSchemaVersion ||
		cursor.TenantRef != query.TenantRef || cursor.WorkspaceID != query.WorkspaceID ||
		cursor.EffectiveState != query.EffectiveState || cursor.Limit != query.Limit ||
		asOfErr != nil || updatedAtErr != nil || asOf.Location() != time.UTC || updatedAt.Location() != time.UTC ||
		asOf.After(query.asOf) || !workspaceInvitationIDPattern.MatchString(cursor.InvitationID) ||
		cursor.BindingDigest != expectedDigest {
		return WorkspaceInvitationListQuery{}, workspaceInvitationCursor{}, errWorkspaceInvitationCursorInvalid
	}
	query.asOf = asOf
	return query, cursor, nil
}

func encodeWorkspaceInvitationCursor(
	query WorkspaceInvitationListQuery,
	last WorkspaceInvitation,
) (string, error) {
	cursor := workspaceInvitationCursor{
		SchemaVersion: workspaceInvitationCursorSchemaVersion,
		TenantRef:     query.TenantRef, WorkspaceID: query.WorkspaceID,
		EffectiveState: query.EffectiveState, Limit: query.Limit,
		AsOf: query.asOf.Format(time.RFC3339Nano), UpdatedAt: last.UpdatedAt.Format(time.RFC3339Nano),
		InvitationID: last.InvitationID,
	}
	cursor.BindingDigest = workspaceInvitationCursorBindingDigest(cursor)
	encoded, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(encoded), nil
}

func workspaceInvitationCursorBindingDigest(cursor workspaceInvitationCursor) string {
	return localIdentityDigest(
		"workspace-invitation-cursor-binding-v1",
		cursor.SchemaVersion,
		cursor.TenantRef,
		cursor.WorkspaceID,
		cursor.EffectiveState,
		cursor.AsOf,
		cursor.UpdatedAt,
		cursor.InvitationID,
		strconv.Itoa(cursor.Limit),
	)
}

func workspaceInvitationComesAfterCursor(invitation WorkspaceInvitation, cursor workspaceInvitationCursor) bool {
	updatedAt, err := time.Parse(time.RFC3339Nano, cursor.UpdatedAt)
	if err != nil {
		return false
	}
	return invitation.UpdatedAt.Before(updatedAt) ||
		invitation.UpdatedAt.Equal(updatedAt) && invitation.InvitationID < cursor.InvitationID
}

func validWorkspaceInvitationEffectiveState(state string) bool {
	return slices.Contains([]string{
		workspaceInvitationEffectivePending,
		workspaceInvitationEffectiveClaimed,
		workspaceInvitationEffectiveRevoked,
		workspaceInvitationEffectiveExpired,
	}, state)
}
