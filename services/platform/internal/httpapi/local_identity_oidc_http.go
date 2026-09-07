package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
)

const (
	localIdentityOIDCStartRoute    = "/v1/auth/oidc/start"
	localIdentityOIDCCallbackRoute = "/v1/auth/oidc/callback"
)

const (
	localIdentityOIDCDisabled             = "LOCAL_IDENTITY_OIDC_DISABLED"
	localIdentityOIDCStateInvalid         = "LOCAL_IDENTITY_OIDC_STATE_INVALID"
	localIdentityOIDCCallbackMismatch     = "LOCAL_IDENTITY_OIDC_CALLBACK_MISMATCH"
	localIdentityOIDCTokenExchangeFailed  = "LOCAL_IDENTITY_OIDC_TOKEN_EXCHANGE_FAILED"
	localIdentityOIDCIdentityMismatch     = "LOCAL_IDENTITY_OIDC_IDENTITY_MISMATCH"
	localIdentityOIDCProviderUnavailable  = "LOCAL_IDENTITY_OIDC_PROVIDER_UNAVAILABLE"
	localIdentityExternalIdentityUnbound  = "LOCAL_IDENTITY_EXTERNAL_IDENTITY_UNBOUND"
	localIdentityExternalIdentityConflict = "LOCAL_IDENTITY_EXTERNAL_IDENTITY_CONFLICT"
	localIdentityOIDCAdmissionDenied      = "LOCAL_IDENTITY_OIDC_ADMISSION_DENIED"
	localIdentityLinkRecentAuthRequired   = "LOCAL_IDENTITY_ACCOUNT_LINK_REQUIRES_RECENT_AUTHENTICATION"
)

type localIdentityOIDCStartRequest struct {
	Intent   string `json:"intent"`
	ReturnTo string `json:"return_to,omitempty"`
}

type localIdentityOIDCStartDocument struct {
	AuthorizationURL string    `json:"authorization_url"`
	ExpiresAt        time.Time `json:"expires_at"`
}

func (service *localIdentityHTTPService) configureOIDC(ctx context.Context, cfg config.Config) error {
	client, err := newLocalIdentityOIDCClient(ctx, cfg)
	if err != nil {
		return localIdentityOIDCSanitizedStartupError(err)
	}
	service.oidcClient = client
	return nil
}

func registerLocalIdentityOIDCHTTPRoutes(mux *http.ServeMux, server *Server) {
	mux.HandleFunc("POST "+localIdentityOIDCStartRoute, server.handleLocalIdentityOIDCStart)
	mux.HandleFunc("GET "+localIdentityOIDCCallbackRoute, server.handleLocalIdentityOIDCCallback)
}

func (server *Server) handleLocalIdentityOIDCStart(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, localIdentityOIDCStartRoute)
	service, ok := server.requireLocalIdentityOIDC(writer, trace)
	if !ok || !service.requireSameOriginJSONRequest(writer, request, trace) {
		return
	}
	var input localIdentityOIDCStartRequest
	if !server.decodeJSONRequestBody(writer, request, trace, &input, jsonRequestBodyOptions{
		maxBytes: 4096, rejectUnknownFields: true, rejectDuplicateFields: true,
	}) {
		return
	}
	intent := strings.TrimSpace(input.Intent)
	returnTo, err := normalizeLocalIdentityReturnTarget(input.ReturnTo)
	if err != nil {
		server.writeLocalIdentityError(writer, trace, localIdentityReturnTargetInvalid)
		return
	}
	requestSession, authenticated := requireLocalIdentityRequestSession(request)
	switch intent {
	case localIdentityOIDCIntentLogin:
		if authenticated {
			server.writeLocalIdentityError(writer, trace, localIdentityAlreadyAuthenticated)
			return
		}
		values := request.Header.Values(localIdentityCSRFHeader)
		if len(values) != 1 || subtle.ConstantTimeCompare([]byte(values[0]), []byte(localIdentityBootstrapCSRF)) != 1 {
			server.writeLocalIdentityError(writer, trace, localIdentityCSRFInvalid)
			return
		}
	case localIdentityOIDCIntentLink:
		if !authenticated {
			server.writeLocalIdentityError(writer, trace, localIdentityAuthenticationRequired)
			return
		}
		if !localIdentitySessionAuthenticationIsRecent(requestSession.lastVerifiedAt, service.nowUTC()) {
			server.writeLocalIdentityError(writer, trace, localIdentityLinkRecentAuthRequired)
			return
		}
		if !service.requireAuthenticatedWriteRequest(writer, request, trace, requestSession.csrfToken, requestSession.csrfCookieValid) {
			return
		}
	default:
		server.writeLocalIdentityError(writer, trace, localIdentityPayloadInvalid)
		return
	}
	transaction, rawState, rawNonce, challenge, err := service.buildOIDCAuthorizationTransaction(
		intent, requestSession.userID, requestSession.sessionID, requestSession.sessionVersion, returnTo, trace.requestID,
	)
	if err != nil {
		server.writeLocalIdentityError(writer, trace, localIdentityServiceUnavailable)
		return
	}
	authorizationURL, err := service.oidcClient.authorizationURL(rawState, rawNonce, challenge)
	if err != nil {
		server.writeLocalIdentityError(writer, trace, localIdentityOIDCCallbackMismatch)
		return
	}
	if service.repository.CreateOIDCAuthorizationTransaction(request.Context(), transaction) != nil {
		server.writeLocalIdentityError(writer, trace, localIdentityServiceUnavailable)
		return
	}
	writeObservedJSON(writer, http.StatusCreated, trace, localIdentityOIDCStartDocument{
		AuthorizationURL: authorizationURL, ExpiresAt: transaction.ExpiresAt,
	})
}

func (server *Server) handleLocalIdentityOIDCCallback(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, localIdentityOIDCCallbackRoute)
	service, ok := server.requireLocalIdentityOIDC(writer, trace)
	if !ok {
		return
	}
	rawState, ok := strictLocalIdentityOIDCCallbackState(request)
	if !ok {
		server.writeLocalIdentityError(writer, trace, localIdentityOIDCCallbackMismatch)
		return
	}
	stateDigest, err := localIdentityOIDCStateDigest(rawState)
	if err != nil {
		server.writeLocalIdentityError(writer, trace, localIdentityOIDCStateInvalid)
		return
	}
	transaction, err := service.repository.ConsumeOIDCAuthorizationTransaction(request.Context(), stateDigest, service.nowUTC())
	if err != nil {
		if errors.Is(err, errLocalIdentityOIDCStateInvalid) || errors.Is(err, errLocalIdentityOIDCStateExpired) {
			server.writeLocalIdentityError(writer, trace, localIdentityOIDCStateInvalid)
			return
		}
		server.writeLocalIdentityError(writer, trace, localIdentityServiceUnavailable)
		return
	}
	currentPolicyDigest := service.oidcClient.policyDigest()
	if subtle.ConstantTimeCompare(transaction.policyDigest, currentPolicyDigest[:]) != 1 {
		server.writeLocalIdentityError(writer, trace, localIdentityOIDCCallbackMismatch)
		return
	}
	code, ok := strictLocalIdentityOIDCCallbackCode(request)
	if !ok {
		server.writeLocalIdentityError(writer, trace, localIdentityOIDCCallbackMismatch)
		return
	}
	idToken, err := service.oidcClient.exchangeCode(request.Context(), code, transaction.codeVerifier)
	if err != nil {
		if errors.Is(err, errLocalIdentityOIDCProviderUnavailable) {
			server.writeLocalIdentityError(writer, trace, localIdentityOIDCProviderUnavailable)
			return
		}
		server.writeLocalIdentityError(writer, trace, localIdentityOIDCTokenExchangeFailed)
		return
	}
	identity, err := service.oidcClient.validateIDToken(request.Context(), idToken, transaction.nonceDigest)
	if err != nil {
		if errors.Is(err, errLocalIdentityOIDCProviderUnavailable) {
			server.writeLocalIdentityError(writer, trace, localIdentityOIDCProviderUnavailable)
			return
		}
		server.writeLocalIdentityError(writer, trace, localIdentityOIDCIdentityMismatch)
		return
	}
	if transaction.Intent == localIdentityOIDCIntentLink {
		if !server.completeLocalIdentityOIDCLink(writer, request, trace, transaction, identity) {
			return
		}
		writeLocalIdentityOIDCRedirect(writer, trace, transaction.ReturnTo)
		return
	}
	_, session, rawCredential, failure := service.completeOIDCLogin(request.Context(), transaction, identity, trace.requestID)
	if failure != "" {
		server.writeLocalIdentityError(writer, trace, failure)
		return
	}
	service.setAuthenticationCookies(writer, rawCredential, session.ExpiresAt)
	writeLocalIdentityOIDCRedirect(writer, trace, transaction.ReturnTo)
}

func (server *Server) completeLocalIdentityOIDCLink(
	writer http.ResponseWriter,
	request *http.Request,
	trace requestTrace,
	transaction LocalIdentityOIDCAuthorizationTransaction,
	identity localIdentityOIDCVerifiedIdentity,
) bool {
	service := server.localIdentityHTTPService
	linkSession, account, err := service.repository.ReadWebSession(request.Context(), transaction.SessionID, service.nowUTC())
	if err != nil || linkSession.RecordVersion != transaction.SessionVersion || account.UserID != transaction.UserID {
		server.writeLocalIdentityError(writer, trace, localIdentityAuthenticationRequired)
		return false
	}
	if !localIdentitySessionAuthenticationIsRecent(linkSession.LastVerifiedAt, service.nowUTC()) {
		server.writeLocalIdentityError(writer, trace, localIdentityLinkRecentAuthRequired)
		return false
	}
	existing, err := service.repository.ResolveExternalIdentity(request.Context(), identity.Issuer, identity.Subject)
	if err == nil {
		if existing.UserID == transaction.UserID {
			return true
		}
		server.writeLocalIdentityError(writer, trace, localIdentityExternalIdentityConflict)
		return false
	}
	if !errors.Is(err, errLocalIdentityNotFound) {
		server.writeLocalIdentityError(writer, trace, localIdentityServiceUnavailable)
		return false
	}
	binding, err := service.buildExternalIdentityBinding(account.UserID, identity, trace.requestID)
	if err != nil {
		server.writeLocalIdentityError(writer, trace, localIdentityServiceUnavailable)
		return false
	}
	if err := service.repository.BindExternalIdentity(request.Context(), binding); err != nil {
		if errors.Is(err, errLocalIdentityExternalConflict) {
			server.writeLocalIdentityError(writer, trace, localIdentityExternalIdentityConflict)
			return false
		}
		server.writeLocalIdentityError(writer, trace, localIdentityServiceUnavailable)
		return false
	}
	return true
}

func (service *localIdentityHTTPService) completeOIDCLogin(
	ctx context.Context,
	transaction LocalIdentityOIDCAuthorizationTransaction,
	identity localIdentityOIDCVerifiedIdentity,
	requestID string,
) (UserAccount, WebSession, string, string) {
	binding, err := service.repository.ResolveExternalIdentity(ctx, identity.Issuer, identity.Subject)
	if err == nil {
		account, readErr := service.repository.ReadAccount(ctx, binding.UserID)
		if readErr != nil || account.LifecycleState != localIdentityStateActive {
			return UserAccount{}, WebSession{}, "", localIdentityAuthenticationFailed
		}
		session, rawCredential, buildErr := service.buildOIDCWebSession(account.UserID, binding.BindingID, requestID)
		if buildErr != nil || service.repository.CreateWebSession(ctx, session) != nil {
			return UserAccount{}, WebSession{}, "", localIdentityServiceUnavailable
		}
		return account, session, rawCredential, ""
	}
	if !errors.Is(err, errLocalIdentityNotFound) {
		return UserAccount{}, WebSession{}, "", localIdentityServiceUnavailable
	}
	if !service.oidcClient.policy.firstLoginEnabled {
		return UserAccount{}, WebSession{}, "", localIdentityOIDCAdmissionDenied
	}
	account, binding, session, rawCredential, buildErr := service.buildOIDCRegistration(identity, requestID)
	if buildErr != nil {
		return UserAccount{}, WebSession{}, "", localIdentityServiceUnavailable
	}
	if err := service.repository.CreateOIDCAccountAndWebSession(ctx, account, binding, session); err != nil {
		if errors.Is(err, errLocalIdentityExternalConflict) || errors.Is(err, errLocalIdentityIdentifierConflict) {
			return UserAccount{}, WebSession{}, "", localIdentityExternalIdentityConflict
		}
		return UserAccount{}, WebSession{}, "", localIdentityServiceUnavailable
	}
	return account, session, rawCredential, ""
}

func (service *localIdentityHTTPService) buildOIDCAuthorizationTransaction(
	intent string,
	userID string,
	sessionID string,
	sessionVersion int,
	returnTo string,
	requestID string,
) (LocalIdentityOIDCAuthorizationTransaction, string, string, string, error) {
	transactionID, err := randomLocalIdentityID("oat_")
	if err != nil {
		return LocalIdentityOIDCAuthorizationTransaction{}, "", "", "", err
	}
	rawState, err := randomLocalIdentityOIDCValue()
	if err != nil {
		return LocalIdentityOIDCAuthorizationTransaction{}, "", "", "", err
	}
	rawNonce, err := randomLocalIdentityOIDCValue()
	if err != nil {
		return LocalIdentityOIDCAuthorizationTransaction{}, "", "", "", err
	}
	codeVerifier, err := randomLocalIdentityOIDCValue()
	if err != nil {
		return LocalIdentityOIDCAuthorizationTransaction{}, "", "", "", err
	}
	stateDigest := sha256.Sum256([]byte(rawState))
	nonceDigest := sha256.Sum256([]byte(rawNonce))
	policyDigest := service.oidcClient.policyDigest()
	challengeDigest := sha256.Sum256([]byte(codeVerifier))
	now := service.nowUTC()
	transaction := LocalIdentityOIDCAuthorizationTransaction{
		SchemaVersion: localIdentitySchemaVersion, TransactionID: transactionID, Intent: intent, UserID: userID,
		SessionID: sessionID, SessionVersion: sessionVersion,
		ReturnTo: returnTo, stateDigest: stateDigest[:], nonceDigest: nonceDigest[:], policyDigest: policyDigest[:],
		codeVerifier: codeVerifier, LifecycleState: localIdentityOIDCTransactionPending, RecordVersion: 1,
		CreatedAt: now, UpdatedAt: now, ExpiresAt: now.Add(service.oidcClient.policy.transactionTTL),
		AuditRef: "auth:oidc-start:" + requestID,
	}
	if !validLocalIdentityOIDCAuthorizationTransaction(transaction) {
		return LocalIdentityOIDCAuthorizationTransaction{}, "", "", "", errLocalIdentityContractMismatch
	}
	return transaction, rawState, rawNonce, base64.RawURLEncoding.EncodeToString(challengeDigest[:]), nil
}

func (service *localIdentityHTTPService) buildOIDCRegistration(
	identity localIdentityOIDCVerifiedIdentity,
	requestID string,
) (UserAccount, ExternalIdentityBinding, WebSession, string, error) {
	userID, err := randomLocalIdentityID("usr_")
	if err != nil {
		return UserAccount{}, ExternalIdentityBinding{}, WebSession{}, "", err
	}
	now := service.nowUTC()
	loginIdentifier := "oidc." + strings.TrimPrefix(userID, "usr_")
	auditRef := "auth:oidc-register:" + requestID
	account := UserAccount{
		SchemaVersion: localIdentitySchemaVersion, UserID: userID, LoginIdentifier: loginIdentifier,
		NormalizedLoginIdentifier: loginIdentifier, DisplayName: "External user", LifecycleState: localIdentityStateActive,
		RecordVersion: 1, CreatedAt: now, UpdatedAt: now, AuditRef: auditRef,
	}
	binding, err := service.buildExternalIdentityBindingAt(userID, identity, requestID, now)
	if err != nil {
		return UserAccount{}, ExternalIdentityBinding{}, WebSession{}, "", err
	}
	session, rawCredential, err := service.buildOIDCWebSessionAt(userID, binding.BindingID, requestID, now)
	return account, binding, session, rawCredential, err
}

func (service *localIdentityHTTPService) buildExternalIdentityBinding(
	userID string,
	identity localIdentityOIDCVerifiedIdentity,
	requestID string,
) (ExternalIdentityBinding, error) {
	return service.buildExternalIdentityBindingAt(userID, identity, requestID, service.nowUTC())
}

func (service *localIdentityHTTPService) buildExternalIdentityBindingAt(
	userID string,
	identity localIdentityOIDCVerifiedIdentity,
	requestID string,
	now time.Time,
) (ExternalIdentityBinding, error) {
	bindingID, err := randomLocalIdentityID("xid_")
	if err != nil {
		return ExternalIdentityBinding{}, err
	}
	binding := ExternalIdentityBinding{
		SchemaVersion: localIdentitySchemaVersion, BindingID: bindingID, UserID: userID,
		Issuer: identity.Issuer, Subject: identity.Subject, LifecycleState: localIdentityStateActive,
		RecordVersion: 1, CreatedAt: now, UpdatedAt: now, AuditRef: "auth:oidc-binding:" + requestID,
	}
	if !validExternalIdentityBinding(binding) {
		return ExternalIdentityBinding{}, errLocalIdentityContractMismatch
	}
	return binding, nil
}

func (service *localIdentityHTTPService) buildOIDCWebSession(userID string, bindingID string, requestID string) (WebSession, string, error) {
	return service.buildOIDCWebSessionAt(userID, bindingID, requestID, service.nowUTC())
}

func (service *localIdentityHTTPService) buildOIDCWebSessionAt(
	userID string,
	bindingID string,
	requestID string,
	now time.Time,
) (WebSession, string, error) {
	sessionID, err := randomLocalIdentityID("ses_")
	if err != nil {
		return WebSession{}, "", err
	}
	rawCredential, err := randomLocalIdentityOIDCValue()
	if err != nil {
		return WebSession{}, "", err
	}
	digest, err := DigestWebSessionCredential(rawCredential)
	if err != nil {
		return WebSession{}, "", err
	}
	session := WebSession{
		SchemaVersion: localIdentitySchemaVersion, SessionID: sessionID, UserID: userID,
		credentialDigest: digest[:], AuthenticationMethod: localAuthenticationMethodOIDC,
		AuthenticationSourceRef: "external:" + bindingID, PolicyVersion: localSessionPolicyVersion,
		LifecycleState: localIdentityStateActive, RecordVersion: 1, CreatedAt: now, UpdatedAt: now,
		LastVerifiedAt: now, ExpiresAt: now.Add(service.sessionTTL), AuditRef: "auth:oidc-session:" + requestID,
	}
	if !validWebSession(session) {
		return WebSession{}, "", errLocalIdentityContractMismatch
	}
	return session, rawCredential, nil
}

func randomLocalIdentityOIDCValue() (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", errLocalIdentityStoreUnavailable
	}
	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

func strictLocalIdentityOIDCCallbackState(request *http.Request) (string, bool) {
	query := request.URL.Query()
	if len(query["state"]) != 1 {
		return "", false
	}
	state := strings.TrimSpace(query.Get("state"))
	return state, state != ""
}

func strictLocalIdentityOIDCCallbackCode(request *http.Request) (string, bool) {
	query := request.URL.Query()
	if len(query) != 2 || len(query["state"]) != 1 || len(query["code"]) != 1 {
		return "", false
	}
	code := strings.TrimSpace(query.Get("code"))
	return code, code != ""
}

func localIdentitySessionAuthenticationIsRecent(lastVerifiedAt time.Time, now time.Time) bool {
	lastVerifiedAt = lastVerifiedAt.UTC()
	now = now.UTC()
	return !lastVerifiedAt.IsZero() && !lastVerifiedAt.After(now) &&
		now.Sub(lastVerifiedAt) <= localIdentityLinkRecentAuthenticationMaxAge
}

func (server *Server) requireLocalIdentityOIDC(writer http.ResponseWriter, trace requestTrace) (*localIdentityHTTPService, bool) {
	service, ok := server.requireLocalIdentityHTTP(writer, trace)
	if !ok {
		return nil, false
	}
	if service.oidcClient == nil {
		server.writeLocalIdentityError(writer, trace, localIdentityOIDCDisabled)
		return nil, false
	}
	return service, true
}

func writeLocalIdentityOIDCRedirect(writer http.ResponseWriter, trace requestTrace, returnTo string) {
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Location", returnTo)
	writeTraceHeaders(writer, trace)
	writer.WriteHeader(http.StatusSeeOther)
	logRequestTrace(trace, http.StatusSeeOther, "", "")
}
