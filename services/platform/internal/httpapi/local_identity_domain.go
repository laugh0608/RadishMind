package httpapi

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"net"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"
)

const (
	localIdentitySchemaVersion         = "local_identity_v1"
	localPasswordAlgorithmPBKDF2SHA256 = "pbkdf2_sha256"
	localPasswordPolicyVersion         = "local_password_policy_v1"
	localPasswordIterations            = 600_000
	localPasswordSaltLength            = 32
	localPasswordKeyLength             = 32
	localSessionPolicyVersion          = "local_web_session_policy_v1"

	localIdentityStateActive     = "active"
	localIdentityStateDisabled   = "disabled"
	localIdentityStateRevoked    = "revoked"
	localIdentityStateSuperseded = "superseded"

	localAuthenticationMethodPassword = "local_password"
	localAuthenticationMethodOIDC     = "oidc"
)

const (
	LocalIdentityFailureStoreUnavailable   = "local_identity_store_unavailable"
	LocalIdentityFailureContractMismatch   = "local_identity_contract_mismatch"
	LocalIdentityFailureIdentifierConflict = "account_identifier_conflict"
	LocalIdentityFailureExternalConflict   = "external_identity_conflict"
	LocalIdentityFailureVersionConflict    = "local_identity_version_conflict"
	LocalIdentityFailureNotFound           = "local_identity_not_found"
	LocalIdentityFailureAccountInactive    = "account_inactive"
	LocalIdentityFailureCredentialInvalid  = "local_credential_invalid"
	LocalIdentityFailureSessionInvalid     = "session_invalid"
	LocalIdentityFailureSessionExpired     = "session_expired"
	LocalIdentityFailureMembershipDenied   = "workspace_membership_denied"
	LocalIdentityFailurePermissionDenied   = "workspace_permission_denied"
)

var (
	errLocalIdentityStoreUnavailable   = errors.New(LocalIdentityFailureStoreUnavailable)
	errLocalIdentityContractMismatch   = errors.New(LocalIdentityFailureContractMismatch)
	errLocalIdentityIdentifierConflict = errors.New(LocalIdentityFailureIdentifierConflict)
	errLocalIdentityExternalConflict   = errors.New(LocalIdentityFailureExternalConflict)
	errLocalIdentityVersionConflict    = errors.New(LocalIdentityFailureVersionConflict)
	errLocalIdentityNotFound           = errors.New(LocalIdentityFailureNotFound)
	errLocalIdentityAccountInactive    = errors.New(LocalIdentityFailureAccountInactive)
	errLocalIdentityCredentialInvalid  = errors.New(LocalIdentityFailureCredentialInvalid)
	errLocalIdentitySessionInvalid     = errors.New(LocalIdentityFailureSessionInvalid)
	errLocalIdentitySessionExpired     = errors.New(LocalIdentityFailureSessionExpired)
	errLocalIdentityMembershipDenied   = errors.New(LocalIdentityFailureMembershipDenied)
	errLocalIdentityPermissionDenied   = errors.New(LocalIdentityFailurePermissionDenied)
)

var (
	localUserIDPattern           = regexp.MustCompile(`^usr_[a-z0-9]{16,64}$`)
	localCredentialIDPattern     = regexp.MustCompile(`^cred_[a-z0-9]{16,64}$`)
	localBindingIDPattern        = regexp.MustCompile(`^xid_[a-z0-9]{16,64}$`)
	localSessionIDPattern        = regexp.MustCompile(`^ses_[a-z0-9]{16,64}$`)
	localRoleAssignmentIDPattern = regexp.MustCompile(`^rla_[a-z0-9]{16,64}$`)
	localMembershipIDPattern     = regexp.MustCompile(`^mbr_[a-z0-9]{16,64}$`)
	localLoginIdentifierPattern  = regexp.MustCompile(`^[a-z0-9][a-z0-9._+@-]{2,253}$`)
	localRoleKeyPattern          = regexp.MustCompile(`^[a-z][a-z0-9_]{2,63}$`)
)

type UserAccount struct {
	SchemaVersion             string     `json:"schema_version"`
	UserID                    string     `json:"user_id"`
	LoginIdentifier           string     `json:"-"`
	NormalizedLoginIdentifier string     `json:"-"`
	DisplayName               string     `json:"display_name"`
	LifecycleState            string     `json:"lifecycle_state"`
	RecordVersion             int        `json:"record_version"`
	CreatedAt                 time.Time  `json:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at"`
	DisabledAt                *time.Time `json:"disabled_at,omitempty"`
	AuditRef                  string     `json:"audit_ref"`
}

type LocalCredential struct {
	SchemaVersion  string `json:"schema_version"`
	CredentialID   string `json:"credential_id"`
	UserID         string `json:"user_id"`
	Algorithm      string `json:"algorithm"`
	PolicyVersion  string `json:"policy_version"`
	Iterations     int    `json:"iterations"`
	KeyLength      int    `json:"key_length"`
	salt           []byte
	derivedKey     []byte
	LifecycleState string    `json:"lifecycle_state"`
	RecordVersion  int       `json:"record_version"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	AuditRef       string    `json:"audit_ref"`
}

type ExternalIdentityBinding struct {
	SchemaVersion  string     `json:"schema_version"`
	BindingID      string     `json:"binding_id"`
	UserID         string     `json:"user_id"`
	Issuer         string     `json:"-"`
	Subject        string     `json:"-"`
	LifecycleState string     `json:"lifecycle_state"`
	RecordVersion  int        `json:"record_version"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	RevokedAt      *time.Time `json:"revoked_at,omitempty"`
	AuditRef       string     `json:"audit_ref"`
}

type WebSession struct {
	SchemaVersion           string `json:"schema_version"`
	SessionID               string `json:"session_id"`
	UserID                  string `json:"user_id"`
	credentialDigest        []byte
	AuthenticationMethod    string     `json:"authentication_method"`
	AuthenticationSourceRef string     `json:"authentication_source_ref"`
	PolicyVersion           string     `json:"policy_version"`
	LifecycleState          string     `json:"lifecycle_state"`
	RecordVersion           int        `json:"record_version"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
	LastVerifiedAt          time.Time  `json:"last_verified_at"`
	ExpiresAt               time.Time  `json:"expires_at"`
	RevokedAt               *time.Time `json:"revoked_at,omitempty"`
	AuditRef                string     `json:"audit_ref"`
}

type LocalRoleAssignment struct {
	SchemaVersion    string     `json:"schema_version"`
	AssignmentID     string     `json:"assignment_id"`
	UserID           string     `json:"user_id"`
	TenantRef        string     `json:"tenant_ref"`
	WorkspaceID      string     `json:"workspace_id,omitempty"`
	RoleKey          string     `json:"role_key"`
	PermissionGrants []string   `json:"permission_grants"`
	LifecycleState   string     `json:"lifecycle_state"`
	RecordVersion    int        `json:"record_version"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	AuditRef         string     `json:"audit_ref"`
}

type WorkspaceMembership struct {
	SchemaVersion  string     `json:"schema_version"`
	MembershipID   string     `json:"membership_id"`
	UserID         string     `json:"user_id"`
	TenantRef      string     `json:"tenant_ref"`
	WorkspaceID    string     `json:"workspace_id"`
	LifecycleState string     `json:"lifecycle_state"`
	RecordVersion  int        `json:"record_version"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	RevokedAt      *time.Time `json:"revoked_at,omitempty"`
	AuditRef       string     `json:"audit_ref"`
}

type LocalWorkspaceAuthorization struct {
	Account          UserAccount
	Membership       WorkspaceMembership
	RoleAssignments  []LocalRoleAssignment
	PermissionGrants []string
}

func NormalizeLocalLoginIdentifier(raw string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if !localLoginIdentifierPattern.MatchString(normalized) {
		return "", errLocalIdentityContractMismatch
	}
	return normalized, nil
}

func NormalizeExternalIssuer(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Host == "" {
		return "", errLocalIdentityContractMismatch
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	if parsed.Scheme != "https" && !(parsed.Scheme == "http" && isLoopbackIssuerHost(parsed.Hostname())) {
		return "", errLocalIdentityContractMismatch
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	parsed.RawPath = ""
	return parsed.String(), nil
}

func isLoopbackIssuerHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	address := net.ParseIP(host)
	return address != nil && address.IsLoopback()
}

func NormalizeExternalSubject(raw string) (string, error) {
	subject := strings.TrimSpace(raw)
	if subject == "" || len(subject) > 255 || strings.ContainsAny(subject, "\x00\r\n") {
		return "", errLocalIdentityContractMismatch
	}
	return subject, nil
}

func DeriveLocalCredential(password string, credentialID string, userID string, now time.Time, auditRef string) (LocalCredential, error) {
	if password == "" || len(password) > 1024 {
		return LocalCredential{}, errLocalIdentityCredentialInvalid
	}
	salt := make([]byte, localPasswordSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return LocalCredential{}, errLocalIdentityStoreUnavailable
	}
	derivedKey, err := pbkdf2.Key(sha256.New, password, salt, localPasswordIterations, localPasswordKeyLength)
	if err != nil {
		return LocalCredential{}, errLocalIdentityCredentialInvalid
	}
	now = now.UTC()
	credential := LocalCredential{
		SchemaVersion: localIdentitySchemaVersion, CredentialID: credentialID, UserID: userID,
		Algorithm: localPasswordAlgorithmPBKDF2SHA256, PolicyVersion: localPasswordPolicyVersion,
		Iterations: localPasswordIterations, KeyLength: localPasswordKeyLength, salt: salt, derivedKey: derivedKey,
		LifecycleState: localIdentityStateActive, RecordVersion: 1, CreatedAt: now, UpdatedAt: now,
		AuditRef: strings.TrimSpace(auditRef),
	}
	if !validLocalCredential(credential) {
		return LocalCredential{}, errLocalIdentityContractMismatch
	}
	return credential, nil
}

func VerifyLocalPassword(password string, credential LocalCredential) bool {
	if password == "" || len(password) > 1024 || !validLocalCredential(credential) || credential.LifecycleState != localIdentityStateActive {
		return false
	}
	derivedKey, err := pbkdf2.Key(sha256.New, password, credential.salt, credential.Iterations, credential.KeyLength)
	if err != nil || len(derivedKey) != len(credential.derivedKey) {
		return false
	}
	return subtle.ConstantTimeCompare(derivedKey, credential.derivedKey) == 1
}

func DigestWebSessionCredential(rawCredential string) ([sha256.Size]byte, error) {
	if len(rawCredential) < 32 || len(rawCredential) > 512 {
		return [sha256.Size]byte{}, errLocalIdentitySessionInvalid
	}
	return sha256.Sum256([]byte(rawCredential)), nil
}

func LocalUserActorRef(userID string) (string, error) {
	if !localUserIDPattern.MatchString(strings.TrimSpace(userID)) {
		return "", errLocalIdentityContractMismatch
	}
	return "user:" + strings.TrimSpace(userID), nil
}

func validUserAccount(account UserAccount) bool {
	identifier, err := NormalizeLocalLoginIdentifier(account.LoginIdentifier)
	return err == nil && identifier == account.NormalizedLoginIdentifier && account.SchemaVersion == localIdentitySchemaVersion &&
		localUserIDPattern.MatchString(account.UserID) && strings.TrimSpace(account.DisplayName) != "" &&
		len(account.DisplayName) <= 120 && account.RecordVersion > 0 && validAuditRef(account.AuditRef) &&
		validLifecycleTimes(account.LifecycleState, account.CreatedAt, account.UpdatedAt, account.DisabledAt, localIdentityStateDisabled)
}

func validLocalCredential(credential LocalCredential) bool {
	return credential.SchemaVersion == localIdentitySchemaVersion && localCredentialIDPattern.MatchString(credential.CredentialID) &&
		localUserIDPattern.MatchString(credential.UserID) && credential.Algorithm == localPasswordAlgorithmPBKDF2SHA256 &&
		credential.PolicyVersion == localPasswordPolicyVersion && credential.Iterations >= localPasswordIterations &&
		credential.KeyLength == localPasswordKeyLength && len(credential.salt) >= 16 && len(credential.derivedKey) == credential.KeyLength &&
		(credential.LifecycleState == localIdentityStateActive || credential.LifecycleState == localIdentityStateSuperseded || credential.LifecycleState == localIdentityStateRevoked) &&
		credential.RecordVersion > 0 && validRequiredTimes(credential.CreatedAt, credential.UpdatedAt) && validAuditRef(credential.AuditRef)
}

func validExternalIdentityBinding(binding ExternalIdentityBinding) bool {
	issuer, issuerErr := NormalizeExternalIssuer(binding.Issuer)
	subject, subjectErr := NormalizeExternalSubject(binding.Subject)
	return issuerErr == nil && subjectErr == nil && issuer == binding.Issuer && subject == binding.Subject &&
		binding.SchemaVersion == localIdentitySchemaVersion && localBindingIDPattern.MatchString(binding.BindingID) &&
		localUserIDPattern.MatchString(binding.UserID) && binding.RecordVersion > 0 && validAuditRef(binding.AuditRef) &&
		validLifecycleTimes(binding.LifecycleState, binding.CreatedAt, binding.UpdatedAt, binding.RevokedAt, localIdentityStateRevoked)
}

func validWebSession(session WebSession) bool {
	zeroDigest := make([]byte, sha256.Size)
	return session.SchemaVersion == localIdentitySchemaVersion && localSessionIDPattern.MatchString(session.SessionID) &&
		localUserIDPattern.MatchString(session.UserID) && len(session.credentialDigest) == sha256.Size &&
		subtle.ConstantTimeCompare(session.credentialDigest, zeroDigest) == 0 &&
		(session.AuthenticationMethod == localAuthenticationMethodPassword || session.AuthenticationMethod == localAuthenticationMethodOIDC) &&
		strings.TrimSpace(session.AuthenticationSourceRef) != "" && len(session.AuthenticationSourceRef) <= 160 &&
		session.PolicyVersion == localSessionPolicyVersion && session.RecordVersion > 0 && validAuditRef(session.AuditRef) &&
		validLifecycleTimes(session.LifecycleState, session.CreatedAt, session.UpdatedAt, session.RevokedAt, localIdentityStateRevoked) &&
		session.ExpiresAt.UTC().After(session.CreatedAt.UTC()) && !session.LastVerifiedAt.UTC().Before(session.CreatedAt.UTC())
}

func validLocalRoleAssignment(assignment LocalRoleAssignment) bool {
	return assignment.SchemaVersion == localIdentitySchemaVersion && localRoleAssignmentIDPattern.MatchString(assignment.AssignmentID) &&
		localUserIDPattern.MatchString(assignment.UserID) && validControlPlaneReadAuthReference(assignment.TenantRef, false) &&
		(assignment.WorkspaceID == "" || validControlPlaneReadAuthReference(assignment.WorkspaceID, false)) &&
		localRoleKeyPattern.MatchString(assignment.RoleKey) && validWorkspacePermissionGrants(assignment.PermissionGrants) &&
		assignment.RecordVersion > 0 && validAuditRef(assignment.AuditRef) && validOptionalExpiry(assignment.CreatedAt, assignment.ExpiresAt) &&
		validLifecycleTimes(assignment.LifecycleState, assignment.CreatedAt, assignment.UpdatedAt, assignment.RevokedAt, localIdentityStateRevoked)
}

func validWorkspaceMembership(membership WorkspaceMembership) bool {
	return membership.SchemaVersion == localIdentitySchemaVersion && localMembershipIDPattern.MatchString(membership.MembershipID) &&
		localUserIDPattern.MatchString(membership.UserID) && validControlPlaneReadAuthReference(membership.TenantRef, false) &&
		validControlPlaneReadAuthReference(membership.WorkspaceID, false) && membership.RecordVersion > 0 &&
		validAuditRef(membership.AuditRef) && validOptionalExpiry(membership.CreatedAt, membership.ExpiresAt) &&
		validLifecycleTimes(membership.LifecycleState, membership.CreatedAt, membership.UpdatedAt, membership.RevokedAt, localIdentityStateRevoked)
}

func validLifecycleTimes(state string, createdAt time.Time, updatedAt time.Time, terminalAt *time.Time, terminalState string) bool {
	if !validRequiredTimes(createdAt, updatedAt) || updatedAt.UTC().Before(createdAt.UTC()) {
		return false
	}
	if state == localIdentityStateActive {
		return terminalAt == nil
	}
	return state == terminalState && terminalAt != nil && !terminalAt.UTC().Before(createdAt.UTC()) && updatedAt.UTC().Equal(terminalAt.UTC())
}

func validRequiredTimes(createdAt time.Time, updatedAt time.Time) bool {
	return !createdAt.IsZero() && !updatedAt.IsZero() && createdAt.Location() == time.UTC && updatedAt.Location() == time.UTC
}

func validOptionalExpiry(createdAt time.Time, expiresAt *time.Time) bool {
	return expiresAt == nil || expiresAt.Location() == time.UTC && expiresAt.After(createdAt)
}

func validAuditRef(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && len(value) <= 160 && !strings.ContainsAny(value, "\x00\r\n")
}

func normalizedPermissionGrants(grants []string) ([]string, bool) {
	normalized, ok := normalizeRequiredWorkspacePermissions(grants)
	if !ok {
		return nil, false
	}
	slices.Sort(normalized)
	return normalized, true
}
