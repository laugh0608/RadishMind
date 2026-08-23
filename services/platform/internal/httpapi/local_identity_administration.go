package httpapi

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	localIdentityWorkspaceMemberSummarySchemaVersion      = "local_identity_workspace_member_summary.v1"
	localIdentityWorkspaceMemberDetailSchemaVersion       = "local_identity_workspace_member_detail.v1"
	localIdentityWorkspaceMembershipViewSchemaVersion     = "local_identity_workspace_membership_view.v1"
	localIdentityWorkspaceRoleAssignmentViewSchemaVersion = "local_identity_workspace_role_assignment_view.v1"
	localIdentityWorkspaceMemberCursorSchemaVersion       = "local_identity_workspace_member_cursor.v1"
	localIdentityWorkspaceMemberDefaultLimit              = 50
	localIdentityWorkspaceMemberMaximumLimit              = 100
	localIdentityAdministrationRecentAuthenticationAge    = 10 * time.Minute
)

const (
	LocalIdentityFailureAdminUnavailable       = "local_identity_admin_unavailable"
	LocalIdentityFailureAdminScopeMismatch     = "local_identity_admin_scope_mismatch"
	LocalIdentityFailureMemberUnavailable      = "local_identity_member_unavailable"
	LocalIdentityFailureMemberCursorInvalid    = "local_identity_member_cursor_invalid"
	LocalIdentityFailureRoleCatalogMismatch    = "local_identity_role_catalog_mismatch"
	LocalIdentityFailureMembershipConflict     = "local_identity_membership_conflict"
	LocalIdentityFailureRoleAssignmentConflict = "local_identity_role_assignment_conflict"
	LocalIdentityFailureSelfMembershipRevoke   = "local_identity_self_membership_revoke_denied"
	LocalIdentityFailureLastAdminRemoval       = "local_identity_last_admin_removal_denied"
	LocalIdentityFailureRecentAuthentication   = "local_identity_recent_authentication_required"
	LocalIdentityFailureAdminBootstrapDenied   = "local_identity_admin_bootstrap_denied"
)

var (
	errLocalIdentityAdminUnavailable       = errors.New(LocalIdentityFailureAdminUnavailable)
	errLocalIdentityAdminScopeMismatch     = errors.New(LocalIdentityFailureAdminScopeMismatch)
	errLocalIdentityMemberUnavailable      = errors.New(LocalIdentityFailureMemberUnavailable)
	errLocalIdentityMemberCursorInvalid    = errors.New(LocalIdentityFailureMemberCursorInvalid)
	errLocalIdentityRoleCatalogMismatch    = errors.New(LocalIdentityFailureRoleCatalogMismatch)
	errLocalIdentityMembershipConflict     = errors.New(LocalIdentityFailureMembershipConflict)
	errLocalIdentityRoleAssignmentConflict = errors.New(LocalIdentityFailureRoleAssignmentConflict)
	errLocalIdentitySelfMembershipRevoke   = errors.New(LocalIdentityFailureSelfMembershipRevoke)
	errLocalIdentityLastAdminRemoval       = errors.New(LocalIdentityFailureLastAdminRemoval)
	errLocalIdentityRecentAuthentication   = errors.New(LocalIdentityFailureRecentAuthentication)
	errLocalIdentityAdminBootstrapDenied   = errors.New(LocalIdentityFailureAdminBootstrapDenied)
)

type LocalIdentityAdministrationActor struct {
	UserID          string
	TenantRef       string
	WorkspaceID     string
	AuthenticatedAt time.Time
}

type LocalIdentityWorkspaceMemberListQuery struct {
	TenantRef       string `json:"tenant_ref"`
	WorkspaceID     string `json:"workspace_id"`
	MembershipState string `json:"membership_state,omitempty"`
	Limit           int    `json:"limit,omitempty"`
	Cursor          string `json:"cursor,omitempty"`
	asOf            time.Time
}

type LocalIdentityWorkspaceMemberSummary struct {
	SchemaVersion            string     `json:"schema_version"`
	TenantRef                string     `json:"tenant_ref"`
	WorkspaceID              string     `json:"workspace_id"`
	UserID                   string     `json:"user_id"`
	DisplayName              string     `json:"display_name"`
	AccountLifecycleState    string     `json:"account_lifecycle_state"`
	MembershipID             string     `json:"membership_id"`
	MembershipLifecycleState string     `json:"membership_lifecycle_state"`
	MembershipRecordVersion  int        `json:"membership_record_version"`
	MembershipExpiresAt      *time.Time `json:"membership_expires_at,omitempty"`
	MembershipEffective      bool       `json:"membership_effective"`
	RoleKeys                 []string   `json:"role_keys"`
	CanManageLocalIdentity   bool       `json:"can_manage_local_identity"`
	RoleCatalogDrift         bool       `json:"role_catalog_drift"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

type LocalIdentityWorkspaceMemberPage struct {
	Members    []LocalIdentityWorkspaceMemberSummary `json:"members"`
	NextCursor string                                `json:"next_cursor,omitempty"`
}

type LocalIdentityWorkspaceMembershipView struct {
	SchemaVersion  string     `json:"schema_version"`
	MembershipID   string     `json:"membership_id"`
	LifecycleState string     `json:"lifecycle_state"`
	RecordVersion  int        `json:"record_version"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	RevokedAt      *time.Time `json:"revoked_at,omitempty"`
	Effective      bool       `json:"effective"`
}

type LocalIdentityWorkspaceRoleAssignmentView struct {
	SchemaVersion          string     `json:"schema_version"`
	AssignmentID           string     `json:"assignment_id"`
	Scope                  string     `json:"scope"`
	WorkspaceID            string     `json:"workspace_id,omitempty"`
	RoleKey                string     `json:"role_key"`
	RoleCatalogVersion     string     `json:"role_catalog_version,omitempty"`
	RoleDefinitionDigest   string     `json:"role_definition_digest,omitempty"`
	PermissionGrants       []string   `json:"permission_grants"`
	LifecycleState         string     `json:"lifecycle_state"`
	RecordVersion          int        `json:"record_version"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
	ExpiresAt              *time.Time `json:"expires_at,omitempty"`
	RevokedAt              *time.Time `json:"revoked_at,omitempty"`
	Effective              bool       `json:"effective"`
	CatalogDrift           bool       `json:"catalog_drift"`
	CanManageLocalIdentity bool       `json:"can_manage_local_identity"`
}

type LocalIdentityWorkspaceMemberDetail struct {
	SchemaVersion          string                                     `json:"schema_version"`
	TenantRef              string                                     `json:"tenant_ref"`
	WorkspaceID            string                                     `json:"workspace_id"`
	UserID                 string                                     `json:"user_id"`
	DisplayName            string                                     `json:"display_name"`
	AccountLifecycleState  string                                     `json:"account_lifecycle_state"`
	AccountRecordVersion   int                                        `json:"account_record_version"`
	Memberships            []LocalIdentityWorkspaceMembershipView     `json:"memberships"`
	RoleAssignments        []LocalIdentityWorkspaceRoleAssignmentView `json:"role_assignments"`
	CanManageLocalIdentity bool                                       `json:"can_manage_local_identity"`
}

type LocalIdentityCreateWorkspaceMembershipInput struct {
	TenantRef   string
	WorkspaceID string
	UserID      string
	ExpiresAt   *time.Time
	AuditRef    string
}

type LocalIdentityAssignWorkspaceRoleInput struct {
	TenantRef                    string
	WorkspaceID                  string
	UserID                       string
	RoleKey                      string
	ExpectedCatalogVersion       string
	ExpectedRoleDefinitionDigest string
	ExpiresAt                    *time.Time
	AuditRef                     string
}

type LocalIdentityRevokeWorkspaceRoleInput struct {
	TenantRef       string
	WorkspaceID     string
	AssignmentID    string
	ExpectedVersion int
	Confirmed       bool
	AuditRef        string
}

type LocalIdentityRevokeWorkspaceMembershipInput struct {
	TenantRef       string
	WorkspaceID     string
	MembershipID    string
	ExpectedVersion int
	Confirmed       bool
	AuditRef        string
}

type LocalIdentityBootstrapWorkspaceAdministratorInput struct {
	TenantRef   string
	WorkspaceID string
	UserID      string
	AuditRef    string
}

type LocalIdentityWorkspaceMembershipRevocation struct {
	Membership             WorkspaceMembership
	RevokedRoleAssignments []LocalRoleAssignment
}

type LocalIdentityWorkspaceAdministratorBootstrap struct {
	Membership     WorkspaceMembership
	RoleAssignment LocalRoleAssignment
}

type localIdentityAdministrationRepository interface {
	AuthorizeWorkspace(context.Context, string, string, string, []string, time.Time) (LocalWorkspaceAuthorization, error)
	CreateWorkspaceMembershipForAdministration(context.Context, string, WorkspaceMembership, time.Time) error
	ListWorkspaceMembers(context.Context, LocalIdentityWorkspaceMemberListQuery) (LocalIdentityWorkspaceMemberPage, error)
	ReadWorkspaceMember(context.Context, string, string, string, time.Time) (LocalIdentityWorkspaceMemberDetail, error)
	CreateCatalogRoleAssignment(context.Context, string, LocalRoleAssignment, time.Time) error
	RevokeCatalogRoleAssignment(context.Context, string, string, string, string, int, time.Time, string) (LocalRoleAssignment, error)
	RevokeWorkspaceMembershipAndAssignments(context.Context, string, string, string, int, string, time.Time, string) (LocalIdentityWorkspaceMembershipRevocation, error)
	BootstrapWorkspaceAdministrator(context.Context, LocalIdentityWorkspaceAdministratorBootstrap, time.Time) (LocalIdentityWorkspaceAdministratorBootstrap, error)
}

type localIdentityAdministrationService struct {
	repository localIdentityAdministrationRepository
	now        func() time.Time
	newID      func(string) (string, error)
}

func newLocalIdentityAdministrationService(repository localIdentityAdministrationRepository) *localIdentityAdministrationService {
	return &localIdentityAdministrationService{
		repository: repository,
		now:        time.Now,
		newID:      randomLocalIdentityID,
	}
}

func (service *localIdentityAdministrationService) ListWorkspaceMembers(
	ctx context.Context,
	actor LocalIdentityAdministrationActor,
	query LocalIdentityWorkspaceMemberListQuery,
) (LocalIdentityWorkspaceMemberPage, error) {
	now := service.currentTime()
	if err := service.authorize(ctx, actor, query.TenantRef, query.WorkspaceID, false, localIdentityPermissionMembersRead); err != nil {
		return LocalIdentityWorkspaceMemberPage{}, err
	}
	query.asOf = now
	page, err := service.repository.ListWorkspaceMembers(ctx, query)
	if err != nil {
		return LocalIdentityWorkspaceMemberPage{}, normalizeLocalIdentityAdministrationError(err)
	}
	return page, nil
}

func (service *localIdentityAdministrationService) ReadWorkspaceMember(
	ctx context.Context,
	actor LocalIdentityAdministrationActor,
	tenantRef string,
	workspaceID string,
	userID string,
) (LocalIdentityWorkspaceMemberDetail, error) {
	now := service.currentTime()
	if err := service.authorize(ctx, actor, tenantRef, workspaceID, false, localIdentityPermissionMembersRead); err != nil {
		return LocalIdentityWorkspaceMemberDetail{}, err
	}
	if !localUserIDPattern.MatchString(strings.TrimSpace(userID)) {
		return LocalIdentityWorkspaceMemberDetail{}, errLocalIdentityContractMismatch
	}
	detail, err := service.repository.ReadWorkspaceMember(ctx, tenantRef, workspaceID, strings.TrimSpace(userID), now)
	if err != nil {
		return LocalIdentityWorkspaceMemberDetail{}, normalizeLocalIdentityAdministrationError(err)
	}
	return detail, nil
}

func (service *localIdentityAdministrationService) ReadRoleCatalog(
	ctx context.Context,
	actor LocalIdentityAdministrationActor,
) (LocalIdentityRoleCatalog, error) {
	if err := service.authorize(ctx, actor, actor.TenantRef, actor.WorkspaceID, false, localIdentityPermissionRolesRead); err != nil {
		return LocalIdentityRoleCatalog{}, err
	}
	return LocalIdentityBuiltInRoleCatalog(), nil
}

func (service *localIdentityAdministrationService) CreateWorkspaceMembership(
	ctx context.Context,
	actor LocalIdentityAdministrationActor,
	input LocalIdentityCreateWorkspaceMembershipInput,
) (WorkspaceMembership, error) {
	now := service.currentTime()
	if err := service.authorize(ctx, actor, input.TenantRef, input.WorkspaceID, true, localIdentityPermissionMembershipsWrite); err != nil {
		return WorkspaceMembership{}, err
	}
	if !localUserIDPattern.MatchString(strings.TrimSpace(input.UserID)) ||
		!validAuditRef(input.AuditRef) || !validAdministrationExpiry(now, input.ExpiresAt) {
		return WorkspaceMembership{}, errLocalIdentityContractMismatch
	}
	membershipID, err := service.newIdentityID("mbr_")
	if err != nil {
		return WorkspaceMembership{}, err
	}
	membership := WorkspaceMembership{
		SchemaVersion:  localIdentitySchemaVersion,
		MembershipID:   membershipID,
		UserID:         strings.TrimSpace(input.UserID),
		TenantRef:      strings.TrimSpace(input.TenantRef),
		WorkspaceID:    strings.TrimSpace(input.WorkspaceID),
		LifecycleState: localIdentityStateActive,
		RecordVersion:  1,
		CreatedAt:      now,
		UpdatedAt:      now,
		ExpiresAt:      cloneTimePointer(input.ExpiresAt),
		AuditRef:       strings.TrimSpace(input.AuditRef),
	}
	if err := service.repository.CreateWorkspaceMembershipForAdministration(ctx, strings.TrimSpace(actor.UserID), membership, now); err != nil {
		return WorkspaceMembership{}, normalizeMembershipCreationError(err)
	}
	return cloneWorkspaceMembership(membership), nil
}

func (service *localIdentityAdministrationService) AssignWorkspaceRole(
	ctx context.Context,
	actor LocalIdentityAdministrationActor,
	input LocalIdentityAssignWorkspaceRoleInput,
) (LocalRoleAssignment, error) {
	now := service.currentTime()
	if err := service.authorize(ctx, actor, input.TenantRef, input.WorkspaceID, true, localIdentityPermissionRolesAssign); err != nil {
		return LocalRoleAssignment{}, err
	}
	definition, exists := builtInLocalIdentityRole(input.RoleKey)
	if !exists || input.ExpectedCatalogVersion != builtInLocalIdentityRoleCatalog.CatalogVersion ||
		input.ExpectedRoleDefinitionDigest != definition.DefinitionDigest {
		return LocalRoleAssignment{}, errLocalIdentityRoleCatalogMismatch
	}
	if !localUserIDPattern.MatchString(strings.TrimSpace(input.UserID)) ||
		!validAuditRef(input.AuditRef) || !validAdministrationExpiry(now, input.ExpiresAt) {
		return LocalRoleAssignment{}, errLocalIdentityContractMismatch
	}
	assignmentID, err := service.newIdentityID("rla_")
	if err != nil {
		return LocalRoleAssignment{}, err
	}
	assignment := LocalRoleAssignment{
		SchemaVersion:        localIdentitySchemaVersion,
		AssignmentID:         assignmentID,
		UserID:               strings.TrimSpace(input.UserID),
		TenantRef:            strings.TrimSpace(input.TenantRef),
		WorkspaceID:          strings.TrimSpace(input.WorkspaceID),
		RoleKey:              definition.RoleKey,
		RoleCatalogVersion:   definition.CatalogVersion,
		RoleDefinitionDigest: definition.DefinitionDigest,
		PermissionGrants:     append([]string(nil), definition.PermissionGrants...),
		LifecycleState:       localIdentityStateActive,
		RecordVersion:        1,
		CreatedAt:            now,
		UpdatedAt:            now,
		ExpiresAt:            cloneTimePointer(input.ExpiresAt),
		AuditRef:             strings.TrimSpace(input.AuditRef),
	}
	if err := service.repository.CreateCatalogRoleAssignment(ctx, strings.TrimSpace(actor.UserID), assignment, now); err != nil {
		return LocalRoleAssignment{}, normalizeRoleAssignmentMutationError(err)
	}
	return cloneLocalRoleAssignment(assignment), nil
}

func (service *localIdentityAdministrationService) RevokeWorkspaceRole(
	ctx context.Context,
	actor LocalIdentityAdministrationActor,
	input LocalIdentityRevokeWorkspaceRoleInput,
) (LocalRoleAssignment, error) {
	now := service.currentTime()
	if err := service.authorize(ctx, actor, input.TenantRef, input.WorkspaceID, true, localIdentityPermissionRolesAssign); err != nil {
		return LocalRoleAssignment{}, err
	}
	if !input.Confirmed || !localRoleAssignmentIDPattern.MatchString(strings.TrimSpace(input.AssignmentID)) ||
		input.ExpectedVersion < 1 || !validAuditRef(input.AuditRef) {
		return LocalRoleAssignment{}, errLocalIdentityContractMismatch
	}
	assignment, err := service.repository.RevokeCatalogRoleAssignment(
		ctx,
		strings.TrimSpace(actor.UserID),
		strings.TrimSpace(input.TenantRef),
		strings.TrimSpace(input.WorkspaceID),
		strings.TrimSpace(input.AssignmentID),
		input.ExpectedVersion,
		now,
		strings.TrimSpace(input.AuditRef),
	)
	if err != nil {
		return LocalRoleAssignment{}, normalizeRoleAssignmentMutationError(err)
	}
	return assignment, nil
}

func (service *localIdentityAdministrationService) RevokeWorkspaceMembership(
	ctx context.Context,
	actor LocalIdentityAdministrationActor,
	input LocalIdentityRevokeWorkspaceMembershipInput,
) (LocalIdentityWorkspaceMembershipRevocation, error) {
	now := service.currentTime()
	if err := service.authorize(ctx, actor, input.TenantRef, input.WorkspaceID, true, localIdentityPermissionMembershipsWrite); err != nil {
		return LocalIdentityWorkspaceMembershipRevocation{}, err
	}
	if !input.Confirmed || !localMembershipIDPattern.MatchString(strings.TrimSpace(input.MembershipID)) ||
		input.ExpectedVersion < 1 || !validAuditRef(input.AuditRef) {
		return LocalIdentityWorkspaceMembershipRevocation{}, errLocalIdentityContractMismatch
	}
	revocation, err := service.repository.RevokeWorkspaceMembershipAndAssignments(
		ctx,
		strings.TrimSpace(input.TenantRef),
		strings.TrimSpace(input.WorkspaceID),
		strings.TrimSpace(input.MembershipID),
		input.ExpectedVersion,
		strings.TrimSpace(actor.UserID),
		now,
		strings.TrimSpace(input.AuditRef),
	)
	if err != nil {
		return LocalIdentityWorkspaceMembershipRevocation{}, normalizeMembershipMutationError(err)
	}
	return revocation, nil
}

func (service *localIdentityAdministrationService) BootstrapWorkspaceAdministrator(
	ctx context.Context,
	input LocalIdentityBootstrapWorkspaceAdministratorInput,
) (LocalIdentityWorkspaceAdministratorBootstrap, error) {
	now := service.currentTime()
	if !validControlPlaneReadAuthReference(strings.TrimSpace(input.TenantRef), false) ||
		!validControlPlaneReadAuthReference(strings.TrimSpace(input.WorkspaceID), false) ||
		!localUserIDPattern.MatchString(strings.TrimSpace(input.UserID)) || !validAuditRef(input.AuditRef) {
		return LocalIdentityWorkspaceAdministratorBootstrap{}, errLocalIdentityContractMismatch
	}
	membershipID, err := service.newIdentityID("mbr_")
	if err != nil {
		return LocalIdentityWorkspaceAdministratorBootstrap{}, err
	}
	assignmentID, err := service.newIdentityID("rla_")
	if err != nil {
		return LocalIdentityWorkspaceAdministratorBootstrap{}, err
	}
	definition, _ := builtInLocalIdentityRole(localIdentityRoleWorkspaceAdmin)
	bootstrap := LocalIdentityWorkspaceAdministratorBootstrap{
		Membership: WorkspaceMembership{
			SchemaVersion:  localIdentitySchemaVersion,
			MembershipID:   membershipID,
			UserID:         strings.TrimSpace(input.UserID),
			TenantRef:      strings.TrimSpace(input.TenantRef),
			WorkspaceID:    strings.TrimSpace(input.WorkspaceID),
			LifecycleState: localIdentityStateActive,
			RecordVersion:  1,
			CreatedAt:      now,
			UpdatedAt:      now,
			AuditRef:       strings.TrimSpace(input.AuditRef),
		},
		RoleAssignment: LocalRoleAssignment{
			SchemaVersion:        localIdentitySchemaVersion,
			AssignmentID:         assignmentID,
			UserID:               strings.TrimSpace(input.UserID),
			TenantRef:            strings.TrimSpace(input.TenantRef),
			WorkspaceID:          strings.TrimSpace(input.WorkspaceID),
			RoleKey:              definition.RoleKey,
			RoleCatalogVersion:   definition.CatalogVersion,
			RoleDefinitionDigest: definition.DefinitionDigest,
			PermissionGrants:     append([]string(nil), definition.PermissionGrants...),
			LifecycleState:       localIdentityStateActive,
			RecordVersion:        1,
			CreatedAt:            now,
			UpdatedAt:            now,
			AuditRef:             strings.TrimSpace(input.AuditRef),
		},
	}
	created, err := service.repository.BootstrapWorkspaceAdministrator(ctx, bootstrap, now)
	if err != nil {
		return LocalIdentityWorkspaceAdministratorBootstrap{}, normalizeBootstrapError(err)
	}
	return created, nil
}

func (service *localIdentityAdministrationService) authorize(
	ctx context.Context,
	actor LocalIdentityAdministrationActor,
	tenantRef string,
	workspaceID string,
	requireRecentAuthentication bool,
	permissions ...string,
) error {
	if service == nil || service.repository == nil {
		return errLocalIdentityAdminUnavailable
	}
	tenantRef = strings.TrimSpace(tenantRef)
	workspaceID = strings.TrimSpace(workspaceID)
	if !localUserIDPattern.MatchString(strings.TrimSpace(actor.UserID)) ||
		!validControlPlaneReadAuthReference(strings.TrimSpace(actor.TenantRef), false) ||
		!validControlPlaneReadAuthReference(strings.TrimSpace(actor.WorkspaceID), false) {
		return errLocalIdentityAdminUnavailable
	}
	if tenantRef != strings.TrimSpace(actor.TenantRef) || workspaceID != strings.TrimSpace(actor.WorkspaceID) ||
		!validControlPlaneReadAuthReference(tenantRef, false) || !validControlPlaneReadAuthReference(workspaceID, false) {
		return errLocalIdentityAdminScopeMismatch
	}
	now := service.currentTime()
	if _, err := service.repository.AuthorizeWorkspace(
		ctx,
		strings.TrimSpace(actor.UserID),
		tenantRef,
		workspaceID,
		permissions,
		now,
	); err != nil {
		return normalizeAdministrationAuthorizationError(err)
	}
	if requireRecentAuthentication &&
		(actor.AuthenticatedAt.IsZero() || actor.AuthenticatedAt.Location() != time.UTC ||
			actor.AuthenticatedAt.After(now) || now.Sub(actor.AuthenticatedAt) > localIdentityAdministrationRecentAuthenticationAge) {
		return errLocalIdentityRecentAuthentication
	}
	return nil
}

func (service *localIdentityAdministrationService) currentTime() time.Time {
	if service == nil || service.now == nil {
		return time.Time{}
	}
	return service.now().UTC()
}

func (service *localIdentityAdministrationService) newIdentityID(prefix string) (string, error) {
	if service == nil || service.newID == nil {
		return "", errLocalIdentityAdminUnavailable
	}
	identifier, err := service.newID(prefix)
	if err != nil {
		return "", errLocalIdentityAdminUnavailable
	}
	identifier = strings.TrimSpace(identifier)
	valid := prefix == "mbr_" && localMembershipIDPattern.MatchString(identifier) ||
		prefix == "rla_" && localRoleAssignmentIDPattern.MatchString(identifier)
	if !valid {
		return "", errLocalIdentityAdminUnavailable
	}
	return identifier, nil
}

func validAdministrationExpiry(now time.Time, expiresAt *time.Time) bool {
	return expiresAt == nil || expiresAt.Location() == time.UTC && expiresAt.After(now)
}

func normalizeAdministrationAuthorizationError(err error) error {
	switch {
	case errors.Is(err, errLocalIdentityMembershipDenied), errors.Is(err, errLocalIdentityPermissionDenied):
		return err
	default:
		return errLocalIdentityAdminUnavailable
	}
}

func normalizeLocalIdentityAdministrationError(err error) error {
	switch {
	case errors.Is(err, errLocalIdentityMemberCursorInvalid):
		return errLocalIdentityMemberCursorInvalid
	case errors.Is(err, errLocalIdentityNotFound), errors.Is(err, errLocalIdentityAccountInactive), errors.Is(err, errLocalIdentityMemberUnavailable):
		return errLocalIdentityMemberUnavailable
	default:
		return errLocalIdentityAdminUnavailable
	}
}

func normalizeMembershipCreationError(err error) error {
	switch {
	case errors.Is(err, errLocalIdentityMembershipDenied), errors.Is(err, errLocalIdentityPermissionDenied):
		return err
	case errors.Is(err, errLocalIdentityNotFound), errors.Is(err, errLocalIdentityAccountInactive):
		return errLocalIdentityMemberUnavailable
	case errors.Is(err, errLocalIdentityIdentifierConflict), errors.Is(err, errLocalIdentityVersionConflict):
		return errLocalIdentityMembershipConflict
	default:
		return normalizeLocalIdentityAdministrationError(err)
	}
}

func normalizeRoleAssignmentMutationError(err error) error {
	switch {
	case errors.Is(err, errLocalIdentityMembershipDenied), errors.Is(err, errLocalIdentityPermissionDenied):
		return err
	case errors.Is(err, errLocalIdentityMemberUnavailable), errors.Is(err, errLocalIdentityNotFound), errors.Is(err, errLocalIdentityAccountInactive):
		return errLocalIdentityMemberUnavailable
	case errors.Is(err, errLocalIdentityRoleCatalogMismatch):
		return errLocalIdentityRoleCatalogMismatch
	case errors.Is(err, errLocalIdentityLastAdminRemoval):
		return errLocalIdentityLastAdminRemoval
	case errors.Is(err, errLocalIdentityIdentifierConflict), errors.Is(err, errLocalIdentityVersionConflict), errors.Is(err, errLocalIdentityRoleAssignmentConflict):
		return errLocalIdentityRoleAssignmentConflict
	default:
		return errLocalIdentityAdminUnavailable
	}
}

func normalizeMembershipMutationError(err error) error {
	switch {
	case errors.Is(err, errLocalIdentityMembershipDenied), errors.Is(err, errLocalIdentityPermissionDenied):
		return err
	case errors.Is(err, errLocalIdentitySelfMembershipRevoke):
		return errLocalIdentitySelfMembershipRevoke
	case errors.Is(err, errLocalIdentityLastAdminRemoval):
		return errLocalIdentityLastAdminRemoval
	case errors.Is(err, errLocalIdentityMemberUnavailable):
		return errLocalIdentityMemberUnavailable
	case errors.Is(err, errLocalIdentityVersionConflict), errors.Is(err, errLocalIdentityMembershipConflict):
		return errLocalIdentityMembershipConflict
	default:
		return errLocalIdentityAdminUnavailable
	}
}

func normalizeBootstrapError(err error) error {
	switch {
	case errors.Is(err, errLocalIdentityNotFound), errors.Is(err, errLocalIdentityAccountInactive):
		return errLocalIdentityMemberUnavailable
	case errors.Is(err, errLocalIdentityAdminBootstrapDenied), errors.Is(err, errLocalIdentityIdentifierConflict):
		return errLocalIdentityAdminBootstrapDenied
	default:
		return errLocalIdentityAdminUnavailable
	}
}
