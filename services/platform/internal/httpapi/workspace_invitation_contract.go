package httpapi

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strings"
	"time"
)

const (
	workspaceInvitationSchemaVersion         = "workspace_invitation.v1"
	workspaceInvitationPreviewSchemaVersion  = "workspace_invitation_preview.v1"
	workspaceInvitationMutationSchemaVersion = "workspace_invitation_mutation.v1"
	workspaceInvitationCreationSchemaVersion = "workspace_invitation_creation.v1"
	workspaceInvitationPageSchemaVersion     = "workspace_invitation_page.v1"
	workspaceInvitationCursorSchemaVersion   = "workspace_invitation_cursor.v1"

	workspaceInvitationLifecyclePending = "pending"
	workspaceInvitationLifecycleClaimed = "claimed"
	workspaceInvitationLifecycleRevoked = "revoked"

	workspaceInvitationEffectivePending = "pending"
	workspaceInvitationEffectiveClaimed = "claimed"
	workspaceInvitationEffectiveRevoked = "revoked"
	workspaceInvitationEffectiveExpired = "expired"

	workspaceInvitationTTL1Hour   = "1h"
	workspaceInvitationTTL24Hours = "24h"
	workspaceInvitationTTL72Hours = "72h"
	workspaceInvitationTTL7Days   = "7d"

	workspaceInvitationDefaultListLimit = 50
	workspaceInvitationMaximumListLimit = 100
	workspaceInvitationRecentAuthAge    = 10 * time.Minute
)

const (
	WorkspaceInvitationFailureAdminUnavailable    = "workspace_invitation_admin_unavailable"
	WorkspaceInvitationFailureCursorInvalid       = "workspace_invitation_cursor_invalid"
	WorkspaceInvitationFailureRoleIneligible      = "workspace_invitation_role_ineligible"
	WorkspaceInvitationFailureRoleCatalogMismatch = "workspace_invitation_role_catalog_mismatch"
	WorkspaceInvitationFailureVersionConflict     = "workspace_invitation_version_conflict"
	WorkspaceInvitationFailureTransitionInvalid   = "workspace_invitation_transition_invalid"
	WorkspaceInvitationFailureInvalid             = "workspace_invitation_invalid"
	WorkspaceInvitationFailureNotClaimable        = "workspace_invitation_not_claimable"
	WorkspaceInvitationFailureAccountIneligible   = "workspace_invitation_account_ineligible"
	WorkspaceInvitationFailureMembershipConflict  = "workspace_invitation_membership_conflict"
	WorkspaceInvitationFailureStoreUnavailable    = "workspace_invitation_store_unavailable"
)

var (
	errWorkspaceInvitationAdminUnavailable    = errors.New(WorkspaceInvitationFailureAdminUnavailable)
	errWorkspaceInvitationCursorInvalid       = errors.New(WorkspaceInvitationFailureCursorInvalid)
	errWorkspaceInvitationRoleIneligible      = errors.New(WorkspaceInvitationFailureRoleIneligible)
	errWorkspaceInvitationRoleCatalogMismatch = errors.New(WorkspaceInvitationFailureRoleCatalogMismatch)
	errWorkspaceInvitationVersionConflict     = errors.New(WorkspaceInvitationFailureVersionConflict)
	errWorkspaceInvitationTransitionInvalid   = errors.New(WorkspaceInvitationFailureTransitionInvalid)
	errWorkspaceInvitationInvalid             = errors.New(WorkspaceInvitationFailureInvalid)
	errWorkspaceInvitationNotClaimable        = errors.New(WorkspaceInvitationFailureNotClaimable)
	errWorkspaceInvitationAccountIneligible   = errors.New(WorkspaceInvitationFailureAccountIneligible)
	errWorkspaceInvitationMembershipConflict  = errors.New(WorkspaceInvitationFailureMembershipConflict)
	errWorkspaceInvitationStoreUnavailable    = errors.New(WorkspaceInvitationFailureStoreUnavailable)
)

var workspaceInvitationIDPattern = regexp.MustCompile(`^wsi_[a-f0-9]{32}$`)
var workspaceInvitationLocatorPattern = regexp.MustCompile(`^rmi_[a-f0-9]{32}$`)

// WorkspaceInvitation is the secret-free canonical current record. The secret
// digest exists only in repository-private state and can never be serialized
// through this contract.
type WorkspaceInvitation struct {
	SchemaVersion        string     `json:"schema_version"`
	InvitationID         string     `json:"invitation_id"`
	RecordVersion        int        `json:"record_version"`
	TenantRef            string     `json:"tenant_ref"`
	WorkspaceID          string     `json:"workspace_id"`
	RoleKey              string     `json:"role_key"`
	RoleCatalogVersion   string     `json:"role_catalog_version"`
	RoleDefinitionDigest string     `json:"role_definition_digest"`
	TTLPolicy            string     `json:"ttl_policy"`
	LifecycleState       string     `json:"lifecycle_state"`
	EffectiveState       string     `json:"effective_state"`
	ExpiresAt            time.Time  `json:"expires_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	CreatedByActorRef    string     `json:"created_by_actor_ref"`
	CreatedRequestRef    string     `json:"created_request_ref"`
	CreatedAuditRef      string     `json:"created_audit_ref"`
	UpdatedRequestRef    string     `json:"updated_request_ref"`
	UpdatedAuditRef      string     `json:"updated_audit_ref"`
	ClaimedAt            *time.Time `json:"claimed_at,omitempty"`
	ClaimedByUserID      string     `json:"claimed_by_user_id,omitempty"`
	MembershipID         string     `json:"membership_id,omitempty"`
	AssignmentID         string     `json:"assignment_id,omitempty"`
	RevokedAt            *time.Time `json:"revoked_at,omitempty"`
	RevokedByActorRef    string     `json:"revoked_by_actor_ref,omitempty"`
}

type WorkspaceInvitationRoleSummary struct {
	RoleKey          string `json:"role_key"`
	DisplayName      string `json:"display_name"`
	Summary          string `json:"summary"`
	CatalogVersion   string `json:"catalog_version"`
	DefinitionDigest string `json:"definition_digest"`
}

type WorkspaceInvitationPreview struct {
	SchemaVersion  string                         `json:"schema_version"`
	InvitationID   string                         `json:"invitation_id"`
	RecordVersion  int                            `json:"record_version"`
	TenantRef      string                         `json:"tenant_ref"`
	WorkspaceID    string                         `json:"workspace_id"`
	Role           WorkspaceInvitationRoleSummary `json:"role"`
	EffectiveState string                         `json:"effective_state"`
	ExpiresAt      time.Time                      `json:"expires_at"`
}

type WorkspaceInvitationMembershipRef struct {
	MembershipID  string `json:"membership_id"`
	UserID        string `json:"user_id"`
	TenantRef     string `json:"tenant_ref"`
	WorkspaceID   string `json:"workspace_id"`
	RecordVersion int    `json:"record_version"`
}

type WorkspaceInvitationAssignmentRef struct {
	AssignmentID  string `json:"assignment_id"`
	UserID        string `json:"user_id"`
	TenantRef     string `json:"tenant_ref"`
	WorkspaceID   string `json:"workspace_id"`
	RoleKey       string `json:"role_key"`
	RecordVersion int    `json:"record_version"`
}

type WorkspaceInvitationMutation struct {
	SchemaVersion  string                            `json:"schema_version"`
	Invitation     WorkspaceInvitation               `json:"invitation"`
	Membership     *WorkspaceInvitationMembershipRef `json:"membership,omitempty"`
	RoleAssignment *WorkspaceInvitationAssignmentRef `json:"role_assignment,omitempty"`
}

type WorkspaceInvitationCreation struct {
	SchemaVersion  string              `json:"schema_version"`
	Invitation     WorkspaceInvitation `json:"invitation"`
	InvitationCode string              `json:"invitation_code"`
}

type WorkspaceInvitationPage struct {
	SchemaVersion string                `json:"schema_version"`
	AsOf          time.Time             `json:"as_of"`
	Invitations   []WorkspaceInvitation `json:"invitations"`
	NextCursor    string                `json:"next_cursor,omitempty"`
}

type WorkspaceInvitationCreateInput struct {
	TenantRef                    string
	WorkspaceID                  string
	RoleKey                      string
	ExpectedCatalogVersion       string
	ExpectedRoleDefinitionDigest string
	TTLPolicy                    string
	Confirmed                    bool
	RequestRef                   string
	AuditRef                     string
}

type WorkspaceInvitationListQuery struct {
	TenantRef      string
	WorkspaceID    string
	EffectiveState string
	Limit          int
	Cursor         string
	asOf           time.Time
}

type WorkspaceInvitationRevokeInput struct {
	TenantRef       string
	WorkspaceID     string
	InvitationID    string
	ExpectedVersion int
	Confirmed       bool
	RequestRef      string
	AuditRef        string
}

type WorkspaceInvitationClaimantActor struct {
	UserID          string
	TenantRef       string
	AuthenticatedAt time.Time
}

type WorkspaceInvitationClaimInput struct {
	InvitationCode  string
	ExpectedVersion int
	Confirmed       bool
	RequestRef      string
	AuditRef        string
}

type workspaceInvitationSecretDigest [sha256.Size]byte

type parsedWorkspaceInvitationCode struct {
	InvitationID string
	secret       [32]byte
}

type workspaceInvitationCursor struct {
	SchemaVersion  string `json:"schema_version"`
	TenantRef      string `json:"tenant_ref"`
	WorkspaceID    string `json:"workspace_id"`
	EffectiveState string `json:"effective_state"`
	Limit          int    `json:"limit"`
	AsOf           string `json:"as_of"`
	UpdatedAt      string `json:"updated_at"`
	InvitationID   string `json:"invitation_id"`
	BindingDigest  string `json:"binding_digest"`
}

var workspaceInvitationDummySecretDigest = func() workspaceInvitationSecretDigest {
	return sha256.Sum256([]byte("radishmind-workspace-invitation-v1:unknown-locator"))
}()

func newWorkspaceInvitationCode(invitationID string) (string, workspaceInvitationSecretDigest, error) {
	if !workspaceInvitationIDPattern.MatchString(strings.TrimSpace(invitationID)) {
		return "", workspaceInvitationSecretDigest{}, errWorkspaceInvitationAdminUnavailable
	}
	secret := [32]byte{}
	if _, err := rand.Read(secret[:]); err != nil {
		return "", workspaceInvitationSecretDigest{}, errWorkspaceInvitationAdminUnavailable
	}
	encodedSecret := base64.RawURLEncoding.EncodeToString(secret[:])
	locator := "rmi_" + strings.TrimPrefix(invitationID, "wsi_")
	return locator + "." + encodedSecret, digestWorkspaceInvitationSecret(secret), nil
}

func parseWorkspaceInvitationCode(raw string) (parsedWorkspaceInvitationCode, error) {
	if raw != strings.TrimSpace(raw) || len(raw) > 128 {
		return parsedWorkspaceInvitationCode{}, errWorkspaceInvitationInvalid
	}
	parts := strings.Split(raw, ".")
	if len(parts) != 2 || !workspaceInvitationLocatorPattern.MatchString(parts[0]) || len(parts[1]) != 43 {
		return parsedWorkspaceInvitationCode{}, errWorkspaceInvitationInvalid
	}
	decoded, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || len(decoded) != 32 || base64.RawURLEncoding.EncodeToString(decoded) != parts[1] {
		return parsedWorkspaceInvitationCode{}, errWorkspaceInvitationInvalid
	}
	parsed := parsedWorkspaceInvitationCode{InvitationID: "wsi_" + strings.TrimPrefix(parts[0], "rmi_")}
	copy(parsed.secret[:], decoded)
	return parsed, nil
}

func digestWorkspaceInvitationSecret(secret [32]byte) workspaceInvitationSecretDigest {
	material := make([]byte, 0, len("radishmind-workspace-invitation-v1\x00")+len(secret))
	material = append(material, "radishmind-workspace-invitation-v1\x00"...)
	material = append(material, secret[:]...)
	return sha256.Sum256(material)
}

func workspaceInvitationSecretMatches(actual, expected workspaceInvitationSecretDigest) bool {
	return subtle.ConstantTimeCompare(actual[:], expected[:]) == 1
}

func validWorkspaceInvitationSecretDigest(digest workspaceInvitationSecretDigest) bool {
	return !workspaceInvitationSecretMatches(digest, workspaceInvitationSecretDigest{})
}

func workspaceInvitationTTLDuration(policy string) (time.Duration, bool) {
	switch strings.TrimSpace(policy) {
	case workspaceInvitationTTL1Hour:
		return time.Hour, true
	case workspaceInvitationTTL24Hours:
		return 24 * time.Hour, true
	case workspaceInvitationTTL72Hours:
		return 72 * time.Hour, true
	case workspaceInvitationTTL7Days:
		return 7 * 24 * time.Hour, true
	default:
		return 0, false
	}
}

func workspaceInvitationRoleEligible(definition LocalIdentityRoleDefinition) bool {
	if definition.CanManageLocalIdentity {
		return false
	}
	return workspaceInvitationRoleKeyEligible(definition.RoleKey)
}

func workspaceInvitationRoleKeyEligible(roleKey string) bool {
	switch roleKey {
	case localIdentityRoleWorkspaceReader, localIdentityRoleWorkspaceBuilder, localIdentityRoleWorkspaceReviewer:
		return true
	default:
		return false
	}
}

func workspaceInvitationEffectiveState(invitation WorkspaceInvitation, asOf time.Time) string {
	switch invitation.LifecycleState {
	case workspaceInvitationLifecycleClaimed:
		return workspaceInvitationEffectiveClaimed
	case workspaceInvitationLifecycleRevoked:
		return workspaceInvitationEffectiveRevoked
	case workspaceInvitationLifecyclePending:
		if !invitation.ExpiresAt.After(asOf.UTC()) {
			return workspaceInvitationEffectiveExpired
		}
		return workspaceInvitationEffectivePending
	default:
		return ""
	}
}

func projectWorkspaceInvitation(invitation WorkspaceInvitation, asOf time.Time) WorkspaceInvitation {
	projected := cloneWorkspaceInvitation(invitation)
	projected.EffectiveState = workspaceInvitationEffectiveState(projected, asOf)
	return projected
}

func validWorkspaceInvitation(invitation WorkspaceInvitation) bool {
	duration, ttlValid := workspaceInvitationTTLDuration(invitation.TTLPolicy)
	metadataValid := strings.TrimSpace(invitation.RoleCatalogVersion) == invitation.RoleCatalogVersion &&
		validControlPlaneReadAuthReference(invitation.RoleCatalogVersion, false) &&
		len(invitation.RoleDefinitionDigest) == len("sha256:")+sha256.Size*2 &&
		strings.HasPrefix(invitation.RoleDefinitionDigest, "sha256:") &&
		isLowerHex(strings.TrimPrefix(invitation.RoleDefinitionDigest, "sha256:"))
	if invitation.SchemaVersion != workspaceInvitationSchemaVersion ||
		!workspaceInvitationIDPattern.MatchString(invitation.InvitationID) || invitation.RecordVersion < 1 ||
		!validControlPlaneReadAuthReference(invitation.TenantRef, false) ||
		!validControlPlaneReadAuthReference(invitation.WorkspaceID, false) ||
		!workspaceInvitationRoleKeyEligible(invitation.RoleKey) || !metadataValid || !ttlValid || invitation.EffectiveState != "" ||
		!validRequiredTimes(invitation.CreatedAt, invitation.UpdatedAt) || invitation.UpdatedAt.Before(invitation.CreatedAt) ||
		!invitation.ExpiresAt.Equal(invitation.CreatedAt.Add(duration)) || invitation.ExpiresAt.Location() != time.UTC ||
		!validWorkspaceInvitationActorRef(invitation.CreatedByActorRef) ||
		!validAuditRef(invitation.CreatedRequestRef) || !validAuditRef(invitation.CreatedAuditRef) ||
		!validAuditRef(invitation.UpdatedRequestRef) || !validAuditRef(invitation.UpdatedAuditRef) {
		return false
	}
	switch invitation.LifecycleState {
	case workspaceInvitationLifecyclePending:
		return invitation.RecordVersion == 1 && invitation.UpdatedAt.Equal(invitation.CreatedAt) &&
			invitation.ClaimedAt == nil && invitation.ClaimedByUserID == "" && invitation.MembershipID == "" &&
			invitation.AssignmentID == "" && invitation.RevokedAt == nil && invitation.RevokedByActorRef == ""
	case workspaceInvitationLifecycleClaimed:
		return invitation.RecordVersion == 2 && invitation.ClaimedAt != nil && invitation.ClaimedAt.Location() == time.UTC &&
			invitation.ClaimedAt.Equal(invitation.UpdatedAt) && invitation.ClaimedAt.Before(invitation.ExpiresAt) &&
			localUserIDPattern.MatchString(invitation.ClaimedByUserID) && localMembershipIDPattern.MatchString(invitation.MembershipID) &&
			localRoleAssignmentIDPattern.MatchString(invitation.AssignmentID) &&
			invitation.RevokedAt == nil && invitation.RevokedByActorRef == ""
	case workspaceInvitationLifecycleRevoked:
		return invitation.RecordVersion == 2 && invitation.RevokedAt != nil && invitation.RevokedAt.Location() == time.UTC &&
			invitation.RevokedAt.Equal(invitation.UpdatedAt) && invitation.RevokedAt.Before(invitation.ExpiresAt) &&
			validWorkspaceInvitationActorRef(invitation.RevokedByActorRef) &&
			invitation.ClaimedAt == nil && invitation.ClaimedByUserID == "" && invitation.MembershipID == "" && invitation.AssignmentID == ""
	default:
		return false
	}
}

func validWorkspaceInvitationActorRef(actorRef string) bool {
	return strings.HasPrefix(actorRef, "user:") && localUserIDPattern.MatchString(strings.TrimPrefix(actorRef, "user:"))
}

func cloneWorkspaceInvitation(invitation WorkspaceInvitation) WorkspaceInvitation {
	invitation.ClaimedAt = cloneTimePointer(invitation.ClaimedAt)
	invitation.RevokedAt = cloneTimePointer(invitation.RevokedAt)
	return invitation
}

func workspaceInvitationRoleSummary(definition LocalIdentityRoleDefinition) WorkspaceInvitationRoleSummary {
	return WorkspaceInvitationRoleSummary{
		RoleKey: definition.RoleKey, DisplayName: definition.DisplayName, Summary: definition.Summary,
		CatalogVersion: definition.CatalogVersion, DefinitionDigest: definition.DefinitionDigest,
	}
}

func workspaceInvitationMembershipRef(membership WorkspaceMembership) *WorkspaceInvitationMembershipRef {
	return &WorkspaceInvitationMembershipRef{
		MembershipID: membership.MembershipID, UserID: membership.UserID, TenantRef: membership.TenantRef,
		WorkspaceID: membership.WorkspaceID, RecordVersion: membership.RecordVersion,
	}
}

func workspaceInvitationAssignmentRef(assignment LocalRoleAssignment) *WorkspaceInvitationAssignmentRef {
	return &WorkspaceInvitationAssignmentRef{
		AssignmentID: assignment.AssignmentID, UserID: assignment.UserID, TenantRef: assignment.TenantRef,
		WorkspaceID: assignment.WorkspaceID, RoleKey: assignment.RoleKey, RecordVersion: assignment.RecordVersion,
	}
}

func workspaceInvitationFailureCode(err error) string {
	switch {
	case errors.Is(err, errWorkspaceInvitationAdminUnavailable):
		return WorkspaceInvitationFailureAdminUnavailable
	case errors.Is(err, errWorkspaceInvitationCursorInvalid):
		return WorkspaceInvitationFailureCursorInvalid
	case errors.Is(err, errWorkspaceInvitationRoleIneligible):
		return WorkspaceInvitationFailureRoleIneligible
	case errors.Is(err, errWorkspaceInvitationRoleCatalogMismatch):
		return WorkspaceInvitationFailureRoleCatalogMismatch
	case errors.Is(err, errWorkspaceInvitationVersionConflict):
		return WorkspaceInvitationFailureVersionConflict
	case errors.Is(err, errWorkspaceInvitationTransitionInvalid):
		return WorkspaceInvitationFailureTransitionInvalid
	case errors.Is(err, errWorkspaceInvitationInvalid):
		return WorkspaceInvitationFailureInvalid
	case errors.Is(err, errWorkspaceInvitationNotClaimable):
		return WorkspaceInvitationFailureNotClaimable
	case errors.Is(err, errWorkspaceInvitationAccountIneligible):
		return WorkspaceInvitationFailureAccountIneligible
	case errors.Is(err, errWorkspaceInvitationMembershipConflict):
		return WorkspaceInvitationFailureMembershipConflict
	case errors.Is(err, errWorkspaceInvitationStoreUnavailable):
		return WorkspaceInvitationFailureStoreUnavailable
	case errors.Is(err, errLocalIdentityRecentAuthentication):
		return LocalIdentityFailureRecentAuthentication
	case errors.Is(err, errLocalIdentityMembershipDenied):
		return LocalIdentityFailureMembershipDenied
	case errors.Is(err, errLocalIdentityPermissionDenied):
		return LocalIdentityFailurePermissionDenied
	default:
		return WorkspaceInvitationFailureStoreUnavailable
	}
}

func decodeWorkspaceInvitationCursor(raw string) (workspaceInvitationCursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil || len(decoded) == 0 || len(decoded) > 4096 {
		return workspaceInvitationCursor{}, errWorkspaceInvitationCursorInvalid
	}
	decoder := json.NewDecoder(bytes.NewReader(decoded))
	decoder.DisallowUnknownFields()
	var cursor workspaceInvitationCursor
	if err := decoder.Decode(&cursor); err != nil {
		return workspaceInvitationCursor{}, errWorkspaceInvitationCursorInvalid
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return workspaceInvitationCursor{}, errWorkspaceInvitationCursorInvalid
	}
	return cursor, nil
}
