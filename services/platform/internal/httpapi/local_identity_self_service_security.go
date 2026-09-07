package httpapi

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	localIdentitySelfServiceSessionSummarySchemaVersion     = "local_identity_self_service_session_summary.v1"
	localIdentitySelfServiceSessionCursorSchemaVersion      = "local_identity_self_service_session_cursor.v1"
	localIdentitySelfServiceSessionRevocationSchemaVersion  = "local_identity_self_service_session_revocation.v1"
	localIdentitySelfServiceBulkRevocationSchemaVersion     = "local_identity_self_service_bulk_session_revocation.v1"
	localIdentitySelfServiceCredentialRotationSchemaVersion = "local_identity_self_service_credential_rotation.v1"
	localIdentitySelfServiceSessionDefaultLimit             = 50
	localIdentitySelfServiceSessionMaximumLimit             = 100
	localIdentitySelfServiceRecentAuthenticationAge         = 10 * time.Minute

	localIdentitySessionEffectiveStateExpired = "expired"
	localIdentitySessionStateFilterAll        = "all"
)

const (
	LocalIdentityFailureSessionCursorInvalid        = "local_identity_session_cursor_invalid"
	LocalIdentityFailureSessionScopeDenied          = "local_identity_session_scope_denied"
	LocalIdentityFailureSessionVersionConflict      = "local_identity_session_version_conflict"
	LocalIdentityFailureSessionRecentAuthentication = "local_identity_session_recent_authentication_required"
	LocalIdentityFailureSessionBulkRevokeConflict   = "local_identity_session_bulk_revoke_conflict"
	LocalIdentityFailureCredentialUnavailable       = "local_identity_credential_unavailable"
	LocalIdentityFailureCredentialCurrentInvalid    = "local_identity_credential_current_invalid"
	LocalIdentityFailureCredentialPolicyRejected    = "local_identity_credential_policy_rejected"
	LocalIdentityFailureCredentialReuseDenied       = "local_identity_credential_reuse_denied"
	LocalIdentityFailureCredentialRotationConflict  = "local_identity_credential_rotation_conflict"
	LocalIdentityFailureSelfServiceUnavailable      = "local_identity_service_unavailable"
)

var (
	errLocalIdentitySessionCursorInvalid        = errors.New(LocalIdentityFailureSessionCursorInvalid)
	errLocalIdentitySessionScopeDenied          = errors.New(LocalIdentityFailureSessionScopeDenied)
	errLocalIdentitySessionVersionConflict      = errors.New(LocalIdentityFailureSessionVersionConflict)
	errLocalIdentitySessionRecentAuthentication = errors.New(LocalIdentityFailureSessionRecentAuthentication)
	errLocalIdentitySessionBulkRevokeConflict   = errors.New(LocalIdentityFailureSessionBulkRevokeConflict)
	errLocalIdentityCredentialUnavailable       = errors.New(LocalIdentityFailureCredentialUnavailable)
	errLocalIdentityCredentialCurrentInvalid    = errors.New(LocalIdentityFailureCredentialCurrentInvalid)
	errLocalIdentityCredentialPolicyRejected    = errors.New(LocalIdentityFailureCredentialPolicyRejected)
	errLocalIdentityCredentialReuseDenied       = errors.New(LocalIdentityFailureCredentialReuseDenied)
	errLocalIdentityCredentialRotationConflict  = errors.New(LocalIdentityFailureCredentialRotationConflict)
	errLocalIdentitySelfServiceUnavailable      = errors.New(LocalIdentityFailureSelfServiceUnavailable)
)

type LocalIdentitySelfServiceActor struct {
	UserID           string
	CurrentSessionID string
	AuthenticatedAt  time.Time
}

type LocalIdentitySelfServiceSessionListQuery struct {
	State  string `json:"state,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Cursor string `json:"cursor,omitempty"`

	requestedAt time.Time
	snapshotAt  time.Time
}

type LocalIdentitySelfServiceSessionSummary struct {
	SchemaVersion        string     `json:"schema_version"`
	SessionID            string     `json:"session_id"`
	AuthenticationMethod string     `json:"authentication_method"`
	EffectiveState       string     `json:"effective_state"`
	RecordVersion        int        `json:"record_version"`
	CurrentSession       bool       `json:"current_session"`
	CreatedAt            time.Time  `json:"created_at"`
	LastVerifiedAt       time.Time  `json:"last_verified_at"`
	ExpiresAt            time.Time  `json:"expires_at"`
	RevokedAt            *time.Time `json:"revoked_at,omitempty"`
}

type LocalIdentitySelfServiceSessionPage struct {
	Sessions   []LocalIdentitySelfServiceSessionSummary `json:"sessions"`
	SnapshotAt time.Time                                `json:"snapshot_at"`
	NextCursor string                                   `json:"next_cursor,omitempty"`
}

type LocalIdentityRevokeOwnedSessionInput struct {
	SessionID       string
	ExpectedVersion int
	Confirmed       bool
	AuditRef        string
}

type LocalIdentityRevokeOtherSessionsInput struct {
	Confirmed bool
	AuditRef  string
}

type LocalIdentityRotateCredentialInput struct {
	CurrentPassword        string
	NewPassword            string
	SessionImpactConfirmed bool
	AuditRef               string
}

type LocalIdentitySelfServiceSessionRevocation struct {
	SchemaVersion         string                                 `json:"schema_version"`
	Session               LocalIdentitySelfServiceSessionSummary `json:"session"`
	CurrentSessionRevoked bool                                   `json:"current_session_revoked"`
}

type LocalIdentitySelfServiceBulkSessionRevocation struct {
	SchemaVersion string `json:"schema_version"`
	RevokedCount  int    `json:"revoked_count"`
}

type LocalIdentitySelfServiceCredentialRotation struct {
	SchemaVersion         string `json:"schema_version"`
	PolicyVersion         string `json:"policy_version"`
	RevokedSessionCount   int    `json:"revoked_session_count"`
	CurrentSessionRevoked bool   `json:"current_session_revoked"`
}

type localIdentityOwnedSessionRevocation struct {
	UserID           string
	CurrentSessionID string
	TargetSessionID  string
	ExpectedVersion  int
	RevokedAt        time.Time
	AuditRef         string
}

type localIdentityOtherSessionRevocation struct {
	UserID           string
	CurrentSessionID string
	RevokedAt        time.Time
	AuditRef         string
}

type localIdentityCredentialRotation struct {
	UserID           string
	CurrentSessionID string
	CurrentPassword  string
	NewPassword      string
	Replacement      LocalCredential
	RotatedAt        time.Time
	AuditRef         string
}

type localIdentitySelfServiceSecurityRepository interface {
	ListSelfServiceSessions(context.Context, string, string, LocalIdentitySelfServiceSessionListQuery) (LocalIdentitySelfServiceSessionPage, error)
	RevokeOwnedWebSession(context.Context, localIdentityOwnedSessionRevocation) (LocalIdentitySelfServiceSessionRevocation, error)
	RevokeOtherWebSessions(context.Context, localIdentityOtherSessionRevocation) (LocalIdentitySelfServiceBulkSessionRevocation, error)
	RotateLocalCredentialAndRevokeSessions(context.Context, localIdentityCredentialRotation) (LocalIdentitySelfServiceCredentialRotation, error)
}

type localIdentitySelfServiceSecurityService struct {
	repository localIdentitySelfServiceSecurityRepository
	now        func() time.Time
	newID      func(string) (string, error)
}

func newLocalIdentitySelfServiceSecurityService(
	repository localIdentitySelfServiceSecurityRepository,
) *localIdentitySelfServiceSecurityService {
	return &localIdentitySelfServiceSecurityService{
		repository: repository,
		now:        time.Now,
		newID:      randomLocalIdentityID,
	}
}

func (service *localIdentitySelfServiceSecurityService) ListSessions(
	ctx context.Context,
	actor LocalIdentitySelfServiceActor,
	query LocalIdentitySelfServiceSessionListQuery,
) (LocalIdentitySelfServiceSessionPage, error) {
	now, err := service.authorize(actor, false)
	if err != nil {
		return LocalIdentitySelfServiceSessionPage{}, err
	}
	query.requestedAt = now
	query.snapshotAt = now
	page, err := service.repository.ListSelfServiceSessions(
		ctx,
		strings.TrimSpace(actor.UserID),
		strings.TrimSpace(actor.CurrentSessionID),
		query,
	)
	if err != nil {
		return LocalIdentitySelfServiceSessionPage{}, normalizeLocalIdentitySelfServiceListError(err)
	}
	return page, nil
}

func (service *localIdentitySelfServiceSecurityService) RevokeSession(
	ctx context.Context,
	actor LocalIdentitySelfServiceActor,
	input LocalIdentityRevokeOwnedSessionInput,
) (LocalIdentitySelfServiceSessionRevocation, error) {
	now, err := service.authorize(actor, true)
	if err != nil {
		return LocalIdentitySelfServiceSessionRevocation{}, err
	}
	input.SessionID = strings.TrimSpace(input.SessionID)
	input.AuditRef = strings.TrimSpace(input.AuditRef)
	if !input.Confirmed || !localSessionIDPattern.MatchString(input.SessionID) ||
		input.ExpectedVersion < 1 || !validAuditRef(input.AuditRef) {
		return LocalIdentitySelfServiceSessionRevocation{}, errLocalIdentityContractMismatch
	}
	result, err := service.repository.RevokeOwnedWebSession(ctx, localIdentityOwnedSessionRevocation{
		UserID:           strings.TrimSpace(actor.UserID),
		CurrentSessionID: strings.TrimSpace(actor.CurrentSessionID),
		TargetSessionID:  input.SessionID,
		ExpectedVersion:  input.ExpectedVersion,
		RevokedAt:        now,
		AuditRef:         input.AuditRef,
	})
	if err != nil {
		return LocalIdentitySelfServiceSessionRevocation{}, normalizeLocalIdentitySelfServiceSessionMutationError(err)
	}
	return result, nil
}

func (service *localIdentitySelfServiceSecurityService) RevokeOtherSessions(
	ctx context.Context,
	actor LocalIdentitySelfServiceActor,
	input LocalIdentityRevokeOtherSessionsInput,
) (LocalIdentitySelfServiceBulkSessionRevocation, error) {
	now, err := service.authorize(actor, true)
	if err != nil {
		return LocalIdentitySelfServiceBulkSessionRevocation{}, err
	}
	input.AuditRef = strings.TrimSpace(input.AuditRef)
	if !input.Confirmed || !validAuditRef(input.AuditRef) {
		return LocalIdentitySelfServiceBulkSessionRevocation{}, errLocalIdentityContractMismatch
	}
	result, err := service.repository.RevokeOtherWebSessions(ctx, localIdentityOtherSessionRevocation{
		UserID:           strings.TrimSpace(actor.UserID),
		CurrentSessionID: strings.TrimSpace(actor.CurrentSessionID),
		RevokedAt:        now,
		AuditRef:         input.AuditRef,
	})
	if err != nil {
		return LocalIdentitySelfServiceBulkSessionRevocation{}, normalizeLocalIdentitySelfServiceBulkMutationError(err)
	}
	return result, nil
}

func (service *localIdentitySelfServiceSecurityService) RotateCredential(
	ctx context.Context,
	actor LocalIdentitySelfServiceActor,
	input LocalIdentityRotateCredentialInput,
) (LocalIdentitySelfServiceCredentialRotation, error) {
	now, err := service.authorize(actor, true)
	if err != nil {
		return LocalIdentitySelfServiceCredentialRotation{}, err
	}
	input.AuditRef = strings.TrimSpace(input.AuditRef)
	if !input.SessionImpactConfirmed || input.CurrentPassword == "" || len(input.CurrentPassword) > 1024 ||
		!validLocalIdentityPassword(input.NewPassword) || !validAuditRef(input.AuditRef) {
		return LocalIdentitySelfServiceCredentialRotation{}, errLocalIdentityCredentialPolicyRejected
	}
	credentialID, err := service.newCredentialID()
	if err != nil {
		return LocalIdentitySelfServiceCredentialRotation{}, err
	}
	replacement, err := DeriveLocalCredential(
		input.NewPassword,
		credentialID,
		strings.TrimSpace(actor.UserID),
		now,
		input.AuditRef,
	)
	if err != nil {
		return LocalIdentitySelfServiceCredentialRotation{}, errLocalIdentitySelfServiceUnavailable
	}
	result, err := service.repository.RotateLocalCredentialAndRevokeSessions(ctx, localIdentityCredentialRotation{
		UserID:           strings.TrimSpace(actor.UserID),
		CurrentSessionID: strings.TrimSpace(actor.CurrentSessionID),
		CurrentPassword:  input.CurrentPassword,
		NewPassword:      input.NewPassword,
		Replacement:      replacement,
		RotatedAt:        now,
		AuditRef:         input.AuditRef,
	})
	if err != nil {
		return LocalIdentitySelfServiceCredentialRotation{}, normalizeLocalIdentityCredentialRotationError(err)
	}
	return result, nil
}

func (service *localIdentitySelfServiceSecurityService) authorize(
	actor LocalIdentitySelfServiceActor,
	requireRecentAuthentication bool,
) (time.Time, error) {
	if service == nil || service.repository == nil || service.now == nil {
		return time.Time{}, errLocalIdentitySelfServiceUnavailable
	}
	now := service.now().UTC()
	if now.IsZero() || !localUserIDPattern.MatchString(strings.TrimSpace(actor.UserID)) ||
		!localSessionIDPattern.MatchString(strings.TrimSpace(actor.CurrentSessionID)) {
		return time.Time{}, errLocalIdentitySessionScopeDenied
	}
	if requireRecentAuthentication &&
		(actor.AuthenticatedAt.IsZero() || actor.AuthenticatedAt.Location() != time.UTC ||
			actor.AuthenticatedAt.After(now) || now.Sub(actor.AuthenticatedAt) > localIdentitySelfServiceRecentAuthenticationAge) {
		return time.Time{}, errLocalIdentitySessionRecentAuthentication
	}
	return now, nil
}

func (service *localIdentitySelfServiceSecurityService) newCredentialID() (string, error) {
	if service == nil || service.newID == nil {
		return "", errLocalIdentitySelfServiceUnavailable
	}
	identifier, err := service.newID("cred_")
	if err != nil || !localCredentialIDPattern.MatchString(strings.TrimSpace(identifier)) {
		return "", errLocalIdentitySelfServiceUnavailable
	}
	return strings.TrimSpace(identifier), nil
}

func normalizeLocalIdentitySelfServiceListError(err error) error {
	switch {
	case errors.Is(err, errLocalIdentitySessionCursorInvalid):
		return errLocalIdentitySessionCursorInvalid
	case errors.Is(err, errLocalIdentitySessionScopeDenied), errors.Is(err, errLocalIdentityNotFound), errors.Is(err, errLocalIdentityAccountInactive):
		return errLocalIdentitySessionScopeDenied
	default:
		return errLocalIdentitySelfServiceUnavailable
	}
}

func normalizeLocalIdentitySelfServiceSessionMutationError(err error) error {
	switch {
	case errors.Is(err, errLocalIdentitySessionScopeDenied):
		return errLocalIdentitySessionScopeDenied
	case errors.Is(err, errLocalIdentitySessionVersionConflict), errors.Is(err, errLocalIdentityVersionConflict):
		return errLocalIdentitySessionVersionConflict
	default:
		return errLocalIdentitySelfServiceUnavailable
	}
}

func normalizeLocalIdentitySelfServiceBulkMutationError(err error) error {
	if errors.Is(err, errLocalIdentitySessionScopeDenied) {
		return errLocalIdentitySessionScopeDenied
	}
	if errors.Is(err, errLocalIdentitySessionBulkRevokeConflict) {
		return errLocalIdentitySessionBulkRevokeConflict
	}
	return errLocalIdentitySelfServiceUnavailable
}

func normalizeLocalIdentityCredentialRotationError(err error) error {
	switch {
	case errors.Is(err, errLocalIdentitySessionScopeDenied):
		return errLocalIdentitySessionScopeDenied
	case errors.Is(err, errLocalIdentityCredentialUnavailable):
		return errLocalIdentityCredentialUnavailable
	case errors.Is(err, errLocalIdentityCredentialCurrentInvalid):
		return errLocalIdentityCredentialCurrentInvalid
	case errors.Is(err, errLocalIdentityCredentialPolicyRejected):
		return errLocalIdentityCredentialPolicyRejected
	case errors.Is(err, errLocalIdentityCredentialReuseDenied):
		return errLocalIdentityCredentialReuseDenied
	case errors.Is(err, errLocalIdentityCredentialRotationConflict), errors.Is(err, errLocalIdentityVersionConflict):
		return errLocalIdentityCredentialRotationConflict
	default:
		return errLocalIdentitySelfServiceUnavailable
	}
}
