package httpapi

import (
	"context"
	"errors"
	"strings"
	"time"
)

type workspaceInvitationRepository interface {
	CreateWorkspaceInvitation(context.Context, string, WorkspaceInvitation, workspaceInvitationSecretDigest, time.Time) error
	ListWorkspaceInvitations(context.Context, string, WorkspaceInvitationListQuery) (WorkspaceInvitationPage, error)
	RevokeWorkspaceInvitation(context.Context, string, WorkspaceInvitationRevokeInput, time.Time) (WorkspaceInvitation, error)
	PreviewWorkspaceInvitation(context.Context, string, string, string, workspaceInvitationSecretDigest, time.Time) (WorkspaceInvitation, error)
	ClaimWorkspaceInvitation(
		context.Context,
		string,
		string,
		string,
		workspaceInvitationSecretDigest,
		int,
		string,
		string,
		time.Time,
		string,
		string,
	) (WorkspaceInvitation, WorkspaceMembership, LocalRoleAssignment, error)
}

type workspaceInvitationService struct {
	repository workspaceInvitationRepository
	now        func() time.Time
	newID      func(string) (string, error)
	newCode    func(string) (string, workspaceInvitationSecretDigest, error)
}

func newWorkspaceInvitationService(repository workspaceInvitationRepository) *workspaceInvitationService {
	return &workspaceInvitationService{
		repository: repository,
		now:        time.Now,
		newID:      randomLocalIdentityID,
		newCode:    newWorkspaceInvitationCode,
	}
}

func (service *workspaceInvitationService) Create(
	ctx context.Context,
	actor LocalIdentityAdministrationActor,
	input WorkspaceInvitationCreateInput,
) (WorkspaceInvitationCreation, error) {
	now := service.currentTime()
	if err := service.validateAdministratorActor(actor, input.TenantRef, input.WorkspaceID, true, now); err != nil {
		return WorkspaceInvitationCreation{}, err
	}
	input.TenantRef = strings.TrimSpace(input.TenantRef)
	input.WorkspaceID = strings.TrimSpace(input.WorkspaceID)
	input.RoleKey = strings.TrimSpace(input.RoleKey)
	input.ExpectedCatalogVersion = strings.TrimSpace(input.ExpectedCatalogVersion)
	input.ExpectedRoleDefinitionDigest = strings.TrimSpace(input.ExpectedRoleDefinitionDigest)
	input.TTLPolicy = strings.TrimSpace(input.TTLPolicy)
	input.RequestRef = strings.TrimSpace(input.RequestRef)
	input.AuditRef = strings.TrimSpace(input.AuditRef)
	duration, ttlValid := workspaceInvitationTTLDuration(input.TTLPolicy)
	if !input.Confirmed || !ttlValid || !validAuditRef(input.RequestRef) || !validAuditRef(input.AuditRef) {
		return WorkspaceInvitationCreation{}, errWorkspaceInvitationTransitionInvalid
	}
	definition, exists := builtInLocalIdentityRole(input.RoleKey)
	if !exists || !workspaceInvitationRoleEligible(definition) {
		return WorkspaceInvitationCreation{}, errWorkspaceInvitationRoleIneligible
	}
	if input.ExpectedCatalogVersion != definition.CatalogVersion ||
		input.ExpectedRoleDefinitionDigest != definition.DefinitionDigest {
		return WorkspaceInvitationCreation{}, errWorkspaceInvitationRoleCatalogMismatch
	}
	invitationID, err := service.newIdentityID("wsi_")
	if err != nil {
		return WorkspaceInvitationCreation{}, errWorkspaceInvitationAdminUnavailable
	}
	invitationCode, secretDigest, err := service.newInvitationCode(invitationID)
	if err != nil {
		return WorkspaceInvitationCreation{}, errWorkspaceInvitationAdminUnavailable
	}
	actorRef, err := LocalUserActorRef(strings.TrimSpace(actor.UserID))
	if err != nil {
		return WorkspaceInvitationCreation{}, errWorkspaceInvitationAdminUnavailable
	}
	invitation := WorkspaceInvitation{
		SchemaVersion: workspaceInvitationSchemaVersion, InvitationID: invitationID, RecordVersion: 1,
		TenantRef: input.TenantRef, WorkspaceID: input.WorkspaceID, RoleKey: definition.RoleKey,
		RoleCatalogVersion: definition.CatalogVersion, RoleDefinitionDigest: definition.DefinitionDigest,
		TTLPolicy: input.TTLPolicy, LifecycleState: workspaceInvitationLifecyclePending,
		ExpiresAt: now.Add(duration), CreatedAt: now, UpdatedAt: now, CreatedByActorRef: actorRef,
		CreatedRequestRef: input.RequestRef, CreatedAuditRef: input.AuditRef,
		UpdatedRequestRef: input.RequestRef, UpdatedAuditRef: input.AuditRef,
	}
	if !validWorkspaceInvitation(invitation) {
		return WorkspaceInvitationCreation{}, errWorkspaceInvitationAdminUnavailable
	}
	if err := service.repository.CreateWorkspaceInvitation(ctx, actor.UserID, invitation, secretDigest, now); err != nil {
		return WorkspaceInvitationCreation{}, normalizeWorkspaceInvitationAdminError(err)
	}
	return WorkspaceInvitationCreation{
		SchemaVersion:  workspaceInvitationCreationSchemaVersion,
		Invitation:     projectWorkspaceInvitation(invitation, now),
		InvitationCode: invitationCode,
	}, nil
}

func (service *workspaceInvitationService) List(
	ctx context.Context,
	actor LocalIdentityAdministrationActor,
	query WorkspaceInvitationListQuery,
) (WorkspaceInvitationPage, error) {
	now := service.currentTime()
	if err := service.validateAdministratorActor(actor, query.TenantRef, query.WorkspaceID, false, now); err != nil {
		return WorkspaceInvitationPage{}, err
	}
	query.TenantRef = strings.TrimSpace(query.TenantRef)
	query.WorkspaceID = strings.TrimSpace(query.WorkspaceID)
	query.EffectiveState = strings.TrimSpace(query.EffectiveState)
	query.Cursor = strings.TrimSpace(query.Cursor)
	query.asOf = now
	query.authorizedAt = now
	page, err := service.repository.ListWorkspaceInvitations(ctx, actor.UserID, query)
	if err != nil {
		return WorkspaceInvitationPage{}, normalizeWorkspaceInvitationAdminError(err)
	}
	return page, nil
}

func (service *workspaceInvitationService) Revoke(
	ctx context.Context,
	actor LocalIdentityAdministrationActor,
	input WorkspaceInvitationRevokeInput,
) (WorkspaceInvitationMutation, error) {
	now := service.currentTime()
	if err := service.validateAdministratorActor(actor, input.TenantRef, input.WorkspaceID, true, now); err != nil {
		return WorkspaceInvitationMutation{}, err
	}
	input.TenantRef = strings.TrimSpace(input.TenantRef)
	input.WorkspaceID = strings.TrimSpace(input.WorkspaceID)
	input.InvitationID = strings.TrimSpace(input.InvitationID)
	input.RequestRef = strings.TrimSpace(input.RequestRef)
	input.AuditRef = strings.TrimSpace(input.AuditRef)
	if !input.Confirmed || !workspaceInvitationIDPattern.MatchString(input.InvitationID) ||
		input.ExpectedVersion < 1 || !validAuditRef(input.RequestRef) || !validAuditRef(input.AuditRef) {
		return WorkspaceInvitationMutation{}, errWorkspaceInvitationTransitionInvalid
	}
	invitation, err := service.repository.RevokeWorkspaceInvitation(ctx, actor.UserID, input, now)
	if err != nil {
		return WorkspaceInvitationMutation{}, normalizeWorkspaceInvitationAdminError(err)
	}
	return WorkspaceInvitationMutation{
		SchemaVersion: workspaceInvitationMutationSchemaVersion,
		Invitation:    projectWorkspaceInvitation(invitation, now),
	}, nil
}

func (service *workspaceInvitationService) Preview(
	ctx context.Context,
	actor WorkspaceInvitationClaimantActor,
	invitationCode string,
) (WorkspaceInvitationPreview, error) {
	now := service.currentTime()
	if err := validateWorkspaceInvitationClaimantActor(actor, now); err != nil {
		return WorkspaceInvitationPreview{}, err
	}
	parsed, err := parseWorkspaceInvitationCode(invitationCode)
	if err != nil {
		return WorkspaceInvitationPreview{}, errWorkspaceInvitationInvalid
	}
	invitation, err := service.repository.PreviewWorkspaceInvitation(
		ctx,
		strings.TrimSpace(actor.UserID),
		strings.TrimSpace(actor.TenantRef),
		parsed.InvitationID,
		digestWorkspaceInvitationSecret(parsed.secret),
		now,
	)
	if err != nil {
		return WorkspaceInvitationPreview{}, normalizeWorkspaceInvitationClaimError(err)
	}
	if invitation.InvitationID != parsed.InvitationID {
		return WorkspaceInvitationPreview{}, errWorkspaceInvitationInvalid
	}
	definition, exists := builtInLocalIdentityRole(invitation.RoleKey)
	if !exists || !workspaceInvitationRoleEligible(definition) ||
		definition.CatalogVersion != invitation.RoleCatalogVersion ||
		definition.DefinitionDigest != invitation.RoleDefinitionDigest {
		return WorkspaceInvitationPreview{}, errWorkspaceInvitationNotClaimable
	}
	return WorkspaceInvitationPreview{
		SchemaVersion: workspaceInvitationPreviewSchemaVersion,
		InvitationID:  invitation.InvitationID, RecordVersion: invitation.RecordVersion,
		TenantRef: invitation.TenantRef, WorkspaceID: invitation.WorkspaceID,
		Role:           workspaceInvitationRoleSummary(definition),
		EffectiveState: workspaceInvitationEffectiveState(invitation, now), ExpiresAt: invitation.ExpiresAt,
	}, nil
}

func (service *workspaceInvitationService) Claim(
	ctx context.Context,
	actor WorkspaceInvitationClaimantActor,
	input WorkspaceInvitationClaimInput,
) (WorkspaceInvitationMutation, error) {
	now := service.currentTime()
	if err := validateWorkspaceInvitationClaimantActor(actor, now); err != nil {
		return WorkspaceInvitationMutation{}, err
	}
	if !input.Confirmed {
		return WorkspaceInvitationMutation{}, errWorkspaceInvitationNotClaimable
	}
	input.RequestRef = strings.TrimSpace(input.RequestRef)
	input.AuditRef = strings.TrimSpace(input.AuditRef)
	if input.ExpectedVersion < 1 || !validAuditRef(input.RequestRef) || !validAuditRef(input.AuditRef) {
		return WorkspaceInvitationMutation{}, errWorkspaceInvitationVersionConflict
	}
	parsed, err := parseWorkspaceInvitationCode(input.InvitationCode)
	if err != nil {
		return WorkspaceInvitationMutation{}, errWorkspaceInvitationInvalid
	}
	membershipID, err := service.newIdentityID("mbr_")
	if err != nil {
		return WorkspaceInvitationMutation{}, errWorkspaceInvitationStoreUnavailable
	}
	assignmentID, err := service.newIdentityID("rla_")
	if err != nil {
		return WorkspaceInvitationMutation{}, errWorkspaceInvitationStoreUnavailable
	}
	invitation, membership, assignment, err := service.repository.ClaimWorkspaceInvitation(
		ctx,
		strings.TrimSpace(actor.UserID),
		strings.TrimSpace(actor.TenantRef),
		parsed.InvitationID,
		digestWorkspaceInvitationSecret(parsed.secret),
		input.ExpectedVersion,
		membershipID,
		assignmentID,
		now,
		input.RequestRef,
		input.AuditRef,
	)
	if err != nil {
		return WorkspaceInvitationMutation{}, normalizeWorkspaceInvitationClaimError(err)
	}
	if invitation.InvitationID != parsed.InvitationID {
		return WorkspaceInvitationMutation{}, errWorkspaceInvitationStoreUnavailable
	}
	return WorkspaceInvitationMutation{
		SchemaVersion:  workspaceInvitationMutationSchemaVersion,
		Invitation:     projectWorkspaceInvitation(invitation, now),
		Membership:     workspaceInvitationMembershipRef(membership),
		RoleAssignment: workspaceInvitationAssignmentRef(assignment),
	}, nil
}

func (service *workspaceInvitationService) validateAdministratorActor(
	actor LocalIdentityAdministrationActor,
	tenantRef string,
	workspaceID string,
	requireRecentAuthentication bool,
	now time.Time,
) error {
	if service == nil || service.repository == nil || now.IsZero() ||
		!localUserIDPattern.MatchString(strings.TrimSpace(actor.UserID)) ||
		!validControlPlaneReadAuthReference(strings.TrimSpace(actor.TenantRef), false) ||
		!validControlPlaneReadAuthReference(strings.TrimSpace(actor.WorkspaceID), false) ||
		strings.TrimSpace(actor.TenantRef) != strings.TrimSpace(tenantRef) ||
		strings.TrimSpace(actor.WorkspaceID) != strings.TrimSpace(workspaceID) {
		return errWorkspaceInvitationAdminUnavailable
	}
	if requireRecentAuthentication && !workspaceInvitationAuthenticationRecent(actor.AuthenticatedAt, now) {
		return errLocalIdentityRecentAuthentication
	}
	return nil
}

func validateWorkspaceInvitationClaimantActor(actor WorkspaceInvitationClaimantActor, now time.Time) error {
	if !localUserIDPattern.MatchString(strings.TrimSpace(actor.UserID)) ||
		!validControlPlaneReadAuthReference(strings.TrimSpace(actor.TenantRef), false) {
		return errWorkspaceInvitationAccountIneligible
	}
	if !workspaceInvitationAuthenticationRecent(actor.AuthenticatedAt, now) {
		return errLocalIdentityRecentAuthentication
	}
	return nil
}

func workspaceInvitationAuthenticationRecent(authenticatedAt, now time.Time) bool {
	return !now.IsZero() && !authenticatedAt.IsZero() && authenticatedAt.Location() == time.UTC &&
		!authenticatedAt.After(now) && now.Sub(authenticatedAt) <= workspaceInvitationRecentAuthAge
}

func (service *workspaceInvitationService) currentTime() time.Time {
	if service == nil || service.now == nil {
		return time.Time{}
	}
	return service.now().UTC()
}

func (service *workspaceInvitationService) newIdentityID(prefix string) (string, error) {
	if service == nil || service.newID == nil {
		return "", errWorkspaceInvitationStoreUnavailable
	}
	identifier, err := service.newID(prefix)
	identifier = strings.TrimSpace(identifier)
	valid := prefix == "wsi_" && workspaceInvitationIDPattern.MatchString(identifier) ||
		prefix == "mbr_" && localMembershipIDPattern.MatchString(identifier) ||
		prefix == "rla_" && localRoleAssignmentIDPattern.MatchString(identifier)
	if err != nil || !valid {
		return "", errWorkspaceInvitationStoreUnavailable
	}
	return identifier, nil
}

func (service *workspaceInvitationService) newInvitationCode(
	invitationID string,
) (string, workspaceInvitationSecretDigest, error) {
	if service == nil || service.newCode == nil {
		return "", workspaceInvitationSecretDigest{}, errWorkspaceInvitationAdminUnavailable
	}
	code, digest, err := service.newCode(invitationID)
	if err != nil || !validWorkspaceInvitationSecretDigest(digest) {
		return "", workspaceInvitationSecretDigest{}, errWorkspaceInvitationAdminUnavailable
	}
	parsed, parseErr := parseWorkspaceInvitationCode(code)
	if parseErr != nil || parsed.InvitationID != invitationID ||
		!workspaceInvitationSecretMatches(digestWorkspaceInvitationSecret(parsed.secret), digest) {
		return "", workspaceInvitationSecretDigest{}, errWorkspaceInvitationAdminUnavailable
	}
	return code, digest, nil
}

func normalizeWorkspaceInvitationAdminError(err error) error {
	switch {
	case errors.Is(err, errWorkspaceInvitationCursorInvalid),
		errors.Is(err, errWorkspaceInvitationRoleIneligible),
		errors.Is(err, errWorkspaceInvitationRoleCatalogMismatch),
		errors.Is(err, errWorkspaceInvitationVersionConflict),
		errors.Is(err, errWorkspaceInvitationTransitionInvalid),
		errors.Is(err, errLocalIdentityRecentAuthentication),
		errors.Is(err, errLocalIdentityMembershipDenied),
		errors.Is(err, errLocalIdentityPermissionDenied):
		return err
	default:
		return errWorkspaceInvitationAdminUnavailable
	}
}

func normalizeWorkspaceInvitationClaimError(err error) error {
	switch {
	case errors.Is(err, errWorkspaceInvitationInvalid),
		errors.Is(err, errWorkspaceInvitationNotClaimable),
		errors.Is(err, errWorkspaceInvitationAccountIneligible),
		errors.Is(err, errWorkspaceInvitationMembershipConflict),
		errors.Is(err, errWorkspaceInvitationVersionConflict),
		errors.Is(err, errWorkspaceInvitationStoreUnavailable),
		errors.Is(err, errLocalIdentityRecentAuthentication):
		return err
	default:
		return errWorkspaceInvitationStoreUnavailable
	}
}
