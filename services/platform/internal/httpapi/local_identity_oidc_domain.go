package httpapi

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"regexp"
	"strings"
	"time"
)

const (
	localIdentityOIDCPolicyVersion = "local_oidc_client_policy_v1"

	localIdentityOIDCIntentLogin = "login"
	localIdentityOIDCIntentLink  = "link"

	localIdentityOIDCTransactionPending  = "pending"
	localIdentityOIDCTransactionConsumed = "consumed"
	localIdentityOIDCTransactionExpired  = "expired"
)

const (
	LocalIdentityFailureOIDCStateInvalid       = "oidc_authorization_state_invalid"
	LocalIdentityFailureOIDCStateExpired       = "oidc_authorization_state_expired"
	LocalIdentityFailureLastLoginMethodRemoval = "last_login_method_removal_denied"
)

var (
	errLocalIdentityOIDCStateInvalid       = errors.New(LocalIdentityFailureOIDCStateInvalid)
	errLocalIdentityOIDCStateExpired       = errors.New(LocalIdentityFailureOIDCStateExpired)
	errLocalIdentityLastLoginMethodRemoval = errors.New(LocalIdentityFailureLastLoginMethodRemoval)

	localIdentityOIDCTransactionIDPattern = regexp.MustCompile(`^oat_[a-z0-9]{16,64}$`)
)

// LocalIdentityOIDCAuthorizationTransaction is the single-use server-side owner
// for a browser authorization request. Credential material is deliberately kept
// out of JSON projections and is cleared by repositories when the record is
// consumed or expires.
type LocalIdentityOIDCAuthorizationTransaction struct {
	SchemaVersion  string `json:"schema_version"`
	TransactionID  string `json:"transaction_id"`
	Intent         string `json:"intent"`
	UserID         string `json:"user_id,omitempty"`
	SessionID      string `json:"session_id,omitempty"`
	SessionVersion int    `json:"session_version,omitempty"`
	ReturnTo       string `json:"return_to"`

	stateDigest  []byte
	nonceDigest  []byte
	policyDigest []byte
	codeVerifier string

	LifecycleState string     `json:"lifecycle_state"`
	RecordVersion  int        `json:"record_version"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	ExpiresAt      time.Time  `json:"expires_at"`
	ConsumedAt     *time.Time `json:"consumed_at,omitempty"`
	AuditRef       string     `json:"audit_ref"`
}

func validLocalIdentityOIDCAuthorizationTransaction(transaction LocalIdentityOIDCAuthorizationTransaction) bool {
	if transaction.SchemaVersion != localIdentitySchemaVersion ||
		!localIdentityOIDCTransactionIDPattern.MatchString(transaction.TransactionID) ||
		(transaction.Intent != localIdentityOIDCIntentLogin && transaction.Intent != localIdentityOIDCIntentLink) ||
		transaction.Intent == localIdentityOIDCIntentLogin && transaction.UserID != "" ||
		transaction.Intent == localIdentityOIDCIntentLogin && (transaction.SessionID != "" || transaction.SessionVersion != 0) ||
		transaction.Intent == localIdentityOIDCIntentLink && (!localUserIDPattern.MatchString(transaction.UserID) ||
			!localSessionIDPattern.MatchString(transaction.SessionID) || transaction.SessionVersion < 1) ||
		len(transaction.stateDigest) != sha256.Size || len(transaction.nonceDigest) != sha256.Size ||
		len(transaction.policyDigest) != sha256.Size || transaction.RecordVersion < 1 ||
		!validAuditRef(transaction.AuditRef) || transaction.CreatedAt.Location() != time.UTC ||
		transaction.UpdatedAt.Location() != time.UTC || transaction.ExpiresAt.Location() != time.UTC ||
		transaction.UpdatedAt.Before(transaction.CreatedAt) || !transaction.ExpiresAt.After(transaction.CreatedAt) ||
		transaction.ExpiresAt.Sub(transaction.CreatedAt) > 15*time.Minute {
		return false
	}
	if _, err := normalizeLocalIdentityReturnTarget(transaction.ReturnTo); err != nil {
		return false
	}
	zeroDigest := make([]byte, sha256.Size)
	if subtle.ConstantTimeCompare(transaction.stateDigest, zeroDigest) == 1 ||
		subtle.ConstantTimeCompare(transaction.nonceDigest, zeroDigest) == 1 ||
		subtle.ConstantTimeCompare(transaction.policyDigest, zeroDigest) == 1 {
		return false
	}
	switch transaction.LifecycleState {
	case localIdentityOIDCTransactionPending:
		return transaction.ConsumedAt == nil && validOIDCCodeVerifier(transaction.codeVerifier)
	case localIdentityOIDCTransactionConsumed, localIdentityOIDCTransactionExpired:
		return transaction.ConsumedAt != nil && transaction.ConsumedAt.Location() == time.UTC &&
			transaction.ConsumedAt.Equal(transaction.UpdatedAt) && !transaction.ConsumedAt.Before(transaction.CreatedAt) &&
			transaction.codeVerifier == ""
	default:
		return false
	}
}

func validOIDCCodeVerifier(value string) bool {
	if len(value) < 43 || len(value) > 128 {
		return false
	}
	for _, character := range value {
		if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' ||
			character >= '0' && character <= '9' || strings.ContainsRune("-._~", character) {
			continue
		}
		return false
	}
	return true
}

func validLocalOIDCAccountAndSessionRegistration(account UserAccount, binding ExternalIdentityBinding, session WebSession) bool {
	return validUserAccount(account) && validExternalIdentityBinding(binding) && validWebSession(session) &&
		account.LifecycleState == localIdentityStateActive && binding.LifecycleState == localIdentityStateActive &&
		session.LifecycleState == localIdentityStateActive && account.UserID == binding.UserID && account.UserID == session.UserID &&
		account.CreatedAt.Equal(binding.CreatedAt) && account.CreatedAt.Equal(session.CreatedAt) &&
		session.AuthenticationMethod == localAuthenticationMethodOIDC &&
		session.AuthenticationSourceRef == "external:"+binding.BindingID
}

func localIdentityOIDCStateDigest(rawState string) ([sha256.Size]byte, error) {
	if len(rawState) < 32 || len(rawState) > 256 || strings.ContainsAny(rawState, "\x00\r\n") {
		return [sha256.Size]byte{}, errLocalIdentityOIDCStateInvalid
	}
	return sha256.Sum256([]byte(rawState)), nil
}

func cloneLocalIdentityOIDCAuthorizationTransaction(transaction LocalIdentityOIDCAuthorizationTransaction) LocalIdentityOIDCAuthorizationTransaction {
	transaction.stateDigest = append([]byte(nil), transaction.stateDigest...)
	transaction.nonceDigest = append([]byte(nil), transaction.nonceDigest...)
	transaction.policyDigest = append([]byte(nil), transaction.policyDigest...)
	transaction.ConsumedAt = cloneTimePointer(transaction.ConsumedAt)
	return transaction
}
