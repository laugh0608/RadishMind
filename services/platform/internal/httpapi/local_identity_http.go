package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"mime"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"radishmind.local/services/platform/internal/config"
)

const (
	localIdentityRegisterRoute       = "/v1/auth/local/register"
	localIdentityLoginRoute          = "/v1/auth/local/login"
	localIdentityCurrentSessionRoute = "/v1/auth/session"
	localIdentityLogoutRoute         = "/v1/auth/logout"
	localIdentitySessionRevokeRoute  = "/v1/auth/sessions/{session_id}/revoke"

	localIdentityActiveTenantHeader = "X-RadishMind-Active-Tenant"
	localIdentityCSRFHeader         = "X-RadishMind-CSRF-Token"
	localIdentityBootstrapCSRF      = "bootstrap"
	localIdentityAuthMode           = controlPlaneReadAuthModeLocalSessionDevTest
)

const (
	localIdentityHTTPDisabled           = "LOCAL_IDENTITY_HTTP_DISABLED"
	localIdentityOriginForbidden        = "LOCAL_IDENTITY_ORIGIN_FORBIDDEN"
	localIdentityCSRFInvalid            = "LOCAL_IDENTITY_CSRF_INVALID"
	localIdentityPayloadInvalid         = "LOCAL_IDENTITY_PAYLOAD_INVALID"
	localIdentityReturnTargetInvalid    = "LOCAL_IDENTITY_RETURN_TARGET_INVALID"
	localIdentityAccountConflict        = "LOCAL_IDENTITY_ACCOUNT_CONFLICT"
	localIdentityAuthenticationFailed   = "LOCAL_IDENTITY_AUTHENTICATION_FAILED"
	localIdentityAuthenticationRequired = "LOCAL_IDENTITY_AUTHENTICATION_REQUIRED"
	localIdentityAlreadyAuthenticated   = "LOCAL_IDENTITY_ALREADY_AUTHENTICATED"
	localIdentitySessionOwnershipDenied = "LOCAL_IDENTITY_SESSION_OWNERSHIP_DENIED"
	localIdentityServiceUnavailable     = "LOCAL_IDENTITY_SERVICE_UNAVAILABLE"
)

type localIdentityHTTPService struct {
	repository    localIdentityRepository
	enabled       bool
	allowedOrigin string
	cookieSecure  bool
	sessionTTL    time.Duration
	now           func() time.Time
}

type localIdentityRequestSession struct {
	sessionID            string
	sessionVersion       int
	userID               string
	displayName          string
	accountLifecycle     string
	authenticationMethod string
	expiresAt            time.Time
	csrfToken            string
	csrfCookieValid      bool
	failure              string
}

type localIdentityRequestSessionContextKey struct{}

type localIdentityRegistrationRequest struct {
	LoginIdentifier string `json:"login_identifier"`
	DisplayName     string `json:"display_name"`
	Password        string `json:"password"`
	ReturnTo        string `json:"return_to,omitempty"`
}

type localIdentityLoginRequest struct {
	LoginIdentifier string `json:"login_identifier"`
	Password        string `json:"password"`
	ReturnTo        string `json:"return_to,omitempty"`
}

type localIdentityAccountDocument struct {
	UserID         string `json:"user_id"`
	DisplayName    string `json:"display_name"`
	LifecycleState string `json:"lifecycle_state"`
}

type localIdentitySessionDocument struct {
	SessionID            string    `json:"session_id"`
	AuthenticationMethod string    `json:"authentication_method"`
	ExpiresAt            time.Time `json:"expires_at"`
}

type localIdentityAuthenticationDocument struct {
	Account  localIdentityAccountDocument `json:"account"`
	Session  localIdentitySessionDocument `json:"session"`
	ReturnTo string                       `json:"return_to,omitempty"`
}

func newLocalIdentityHTTPService(cfg config.Config, repository localIdentityRepository) *localIdentityHTTPService {
	ttl := cfg.LocalIdentitySessionTTL
	if ttl <= 0 {
		ttl = 12 * time.Hour
	}
	return &localIdentityHTTPService{
		repository: repository, enabled: cfg.LocalIdentityDevHTTPEnabled,
		allowedOrigin: strings.TrimSpace(cfg.LocalIdentityAllowedOrigin),
		cookieSecure:  cfg.LocalIdentityCookieSecure, sessionTTL: ttl, now: time.Now,
	}
}

func registerLocalIdentityHTTPRoutes(mux *http.ServeMux, server *Server) {
	mux.HandleFunc("POST "+localIdentityRegisterRoute, server.handleLocalIdentityRegister)
	mux.HandleFunc("POST "+localIdentityLoginRoute, server.handleLocalIdentityLogin)
	mux.HandleFunc("GET "+localIdentityCurrentSessionRoute, server.handleLocalIdentityCurrentSession)
	mux.HandleFunc("POST "+localIdentityLogoutRoute, server.handleLocalIdentityLogout)
	mux.HandleFunc("POST "+localIdentitySessionRevokeRoute, server.handleLocalIdentitySessionRevoke)
}

func withLocalIdentitySessionAuthentication(next http.Handler, service *localIdentityHTTPService) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if service == nil || !service.enabled || isGatewayNorthboundRequest(request) {
			next.ServeHTTP(writer, request)
			return
		}
		requestSession, auth := service.authenticateRequest(request)
		ctx := context.WithValue(request.Context(), localIdentityRequestSessionContextKey{}, requestSession)
		ctx = withControlPlaneReadFakeAuthContext(ctx, auth)
		next.ServeHTTP(writer, sanitizedLocalIdentityCookieRequest(request, ctx, service))
	})
}

func sanitizedLocalIdentityCookieRequest(request *http.Request, ctx context.Context, service *localIdentityHTTPService) *http.Request {
	cloned := request.Clone(ctx)
	cloned.Header.Del("Cookie")
	for _, cookie := range request.Cookies() {
		if cookie.Name == service.sessionCookieName() || cookie.Name == service.csrfCookieName() {
			continue
		}
		cloned.AddCookie(cookie)
	}
	return cloned
}

func (service *localIdentityHTTPService) authenticateRequest(request *http.Request) (localIdentityRequestSession, controlPlaneReadAuthContext) {
	failure := localIdentityAuthenticationRequired
	if controlPlaneReadHasAnyDevHeader(request) || strings.TrimSpace(request.Header.Get("Authorization")) != "" {
		return localIdentityRequestSession{failure: "auth_context_contract_mismatch"}, controlPlaneReadAuthContext{
			AuthMode: localIdentityAuthMode, FailureCode: "auth_context_contract_mismatch",
		}
	}
	rawCredential, ok := readSingleCookieValue(request, service.sessionCookieName())
	if !ok {
		return localIdentityRequestSession{failure: failure}, controlPlaneReadAuthContext{AuthMode: localIdentityAuthMode, FailureCode: "identity_context_missing"}
	}
	digest, err := DigestWebSessionCredential(rawCredential)
	if err != nil || service.repository == nil {
		return localIdentityRequestSession{failure: failure}, controlPlaneReadAuthContext{AuthMode: localIdentityAuthMode, FailureCode: "identity_context_missing"}
	}
	now := service.nowUTC()
	session, account, err := service.repository.ResolveWebSession(request.Context(), digest, now)
	if err != nil {
		failureCode := "identity_context_missing"
		if errors.Is(err, errLocalIdentityStoreUnavailable) {
			failure = localIdentityServiceUnavailable
			failureCode = "identity_provider_unavailable"
		}
		return localIdentityRequestSession{failure: failure}, controlPlaneReadAuthContext{AuthMode: localIdentityAuthMode, FailureCode: failureCode}
	}
	tenantRef, tenantValid := selectedLocalIdentityTenant(request)
	actorRef, _ := LocalUserActorRef(account.UserID)
	scopes := localIdentityControlPlaneScopeCandidates()
	identity := &VerifiedControlPlaneIdentity{
		AuthSource: localIdentityAuthMode, IssuerRef: "issuer:radishmind-local", SubjectRef: actorRef,
		TenantRef: tenantRef, ScopeGrants: append([]string(nil), scopes...), IssuedAt: session.CreatedAt,
		ExpiresAt: session.ExpiresAt, AuthTime: session.CreatedAt, PolicyVersion: localSessionPolicyVersion,
		SessionRef: "session:" + session.SessionID,
	}
	auth := controlPlaneReadAuthContext{
		AuthMode: localIdentityAuthMode, IdentityContext: "verified:local-web-session", TenantBinding: tenantRef,
		SubjectBinding: actorRef, ScopeGrants: scopes, AuditContext: "audit:local-web-session",
		IssuerRef: identity.IssuerRef, SessionRef: identity.SessionRef, VerifiedIdentity: identity,
		ResourceBinding: ControlPlaneResourceBinding{
			TenantRef: tenantRef, TenantVerified: tenantValid, PermissionGrants: append([]string(nil), scopes...),
			SourceRef: "binding:local-session-selection", PolicyVersion: localSessionPolicyVersion,
			ExpiresAt: session.ExpiresAt,
		},
	}
	if !tenantValid && tenantRef != "" {
		auth.FailureCode = "tenant_binding_missing"
		auth.FailureStatus = http.StatusForbidden
	}
	csrfToken := deriveLocalIdentityCSRFToken(rawCredential)
	csrfCookie, csrfCookieFound := readSingleCookieValue(request, service.csrfCookieName())
	return localIdentityRequestSession{
		sessionID: session.SessionID, sessionVersion: session.RecordVersion, userID: account.UserID,
		displayName: account.DisplayName, accountLifecycle: account.LifecycleState,
		authenticationMethod: session.AuthenticationMethod, expiresAt: session.ExpiresAt,
		csrfToken:       csrfToken,
		csrfCookieValid: csrfCookieFound && subtle.ConstantTimeCompare([]byte(csrfCookie), []byte(csrfToken)) == 1,
	}, auth
}

func (server *Server) handleLocalIdentityRegister(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, localIdentityRegisterRoute)
	service, ok := server.requireLocalIdentityHTTP(writer, trace)
	if !ok || !service.requireBootstrapWriteRequest(writer, request, trace) || !server.requireUnauthenticatedLocalIdentityRequest(writer, request, trace) {
		return
	}
	var input localIdentityRegistrationRequest
	if !server.decodeJSONRequestBody(writer, request, trace, &input, jsonRequestBodyOptions{
		maxBytes: 16 << 10, rejectUnknownFields: true, rejectDuplicateFields: true,
	}) {
		return
	}
	returnTo, err := normalizeLocalIdentityReturnTarget(input.ReturnTo)
	if err != nil {
		server.writeLocalIdentityError(writer, trace, localIdentityReturnTargetInvalid)
		return
	}
	account, credential, session, rawCredential, err := service.buildRegistration(input, trace.requestID)
	if err != nil {
		server.writeLocalIdentityError(writer, trace, localIdentityPayloadInvalid)
		return
	}
	if err := service.repository.CreateAccountAndWebSession(request.Context(), account, credential, session); err != nil {
		if errors.Is(err, errLocalIdentityIdentifierConflict) {
			server.writeLocalIdentityError(writer, trace, localIdentityAccountConflict)
			return
		}
		server.writeLocalIdentityError(writer, trace, localIdentityServiceUnavailable)
		return
	}
	service.setAuthenticationCookies(writer, rawCredential, session.ExpiresAt)
	writeObservedJSON(writer, http.StatusCreated, trace, localIdentityAuthenticationResponse(account, session, returnTo))
}

func (server *Server) handleLocalIdentityLogin(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, localIdentityLoginRoute)
	service, ok := server.requireLocalIdentityHTTP(writer, trace)
	if !ok || !service.requireBootstrapWriteRequest(writer, request, trace) || !server.requireUnauthenticatedLocalIdentityRequest(writer, request, trace) {
		return
	}
	var input localIdentityLoginRequest
	if !server.decodeJSONRequestBody(writer, request, trace, &input, jsonRequestBodyOptions{
		maxBytes: 16 << 10, rejectUnknownFields: true, rejectDuplicateFields: true,
	}) {
		return
	}
	returnTo, err := normalizeLocalIdentityReturnTarget(input.ReturnTo)
	if err != nil {
		server.writeLocalIdentityError(writer, trace, localIdentityReturnTargetInvalid)
		return
	}
	account, credential, authenticated, storeFailure := service.authenticatePassword(request.Context(), input.LoginIdentifier, input.Password)
	if storeFailure {
		server.writeLocalIdentityError(writer, trace, localIdentityServiceUnavailable)
		return
	}
	if !authenticated {
		server.writeLocalIdentityError(writer, trace, localIdentityAuthenticationFailed)
		return
	}
	session, rawCredential, err := service.buildWebSession(account.UserID, credential.CredentialID, trace.requestID)
	if err != nil || service.repository.CreateWebSession(request.Context(), session) != nil {
		server.writeLocalIdentityError(writer, trace, localIdentityServiceUnavailable)
		return
	}
	service.setAuthenticationCookies(writer, rawCredential, session.ExpiresAt)
	writeObservedJSON(writer, http.StatusOK, trace, localIdentityAuthenticationResponse(account, session, returnTo))
}

func (server *Server) handleLocalIdentityCurrentSession(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, localIdentityCurrentSessionRoute)
	service, ok := server.requireLocalIdentityHTTP(writer, trace)
	if !ok {
		return
	}
	requestSession, ok := requireLocalIdentityRequestSession(request)
	if !ok {
		server.writeLocalIdentityError(writer, trace, localIdentityAuthenticationRequired)
		return
	}
	service.setCSRFCookie(writer, requestSession.csrfToken, requestSession.expiresAt)
	writeObservedJSON(writer, http.StatusOK, trace, localIdentityAuthenticationDocument{
		Account: localIdentityAccountDocument{
			UserID: requestSession.userID, DisplayName: requestSession.displayName, LifecycleState: requestSession.accountLifecycle,
		},
		Session: localIdentitySessionDocument{
			SessionID: requestSession.sessionID, AuthenticationMethod: requestSession.authenticationMethod, ExpiresAt: requestSession.expiresAt,
		},
	})
}

func (server *Server) handleLocalIdentityLogout(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, localIdentityLogoutRoute)
	server.revokeCurrentLocalIdentitySession(writer, request, trace, "logout")
}

func (server *Server) handleLocalIdentitySessionRevoke(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, localIdentitySessionRevokeRoute)
	server.revokeTargetLocalIdentitySession(writer, request, trace, "revoke", strings.TrimSpace(request.PathValue("session_id")))
}

func (server *Server) revokeCurrentLocalIdentitySession(writer http.ResponseWriter, request *http.Request, trace requestTrace, action string) {
	server.revokeTargetLocalIdentitySession(writer, request, trace, action, "")
}

func (server *Server) revokeTargetLocalIdentitySession(
	writer http.ResponseWriter,
	request *http.Request,
	trace requestTrace,
	action string,
	targetSessionID string,
) {
	service, ok := server.requireLocalIdentityHTTP(writer, trace)
	if !ok {
		return
	}
	requestSession, ok := requireLocalIdentityRequestSession(request)
	if !ok {
		server.writeLocalIdentityError(writer, trace, localIdentityAuthenticationRequired)
		return
	}
	if !service.requireAuthenticatedWriteRequest(writer, request, trace, requestSession.csrfToken, requestSession.csrfCookieValid) {
		return
	}
	if targetSessionID != "" && targetSessionID != requestSession.sessionID {
		server.writeLocalIdentityError(writer, trace, localIdentitySessionOwnershipDenied)
		return
	}
	var input struct{}
	if !server.decodeJSONRequestBody(writer, request, trace, &input, jsonRequestBodyOptions{
		maxBytes: 1024, rejectUnknownFields: true, rejectDuplicateFields: true,
	}) {
		return
	}
	auditRef := "auth:" + action + ":" + trace.requestID
	_, err := service.repository.RevokeWebSession(request.Context(), requestSession.sessionID, requestSession.sessionVersion, service.nowUTC(), auditRef)
	if err != nil && !errors.Is(err, errLocalIdentityVersionConflict) {
		server.writeLocalIdentityError(writer, trace, localIdentityServiceUnavailable)
		return
	}
	service.clearAuthenticationCookies(writer)
	writeTraceHeaders(writer, trace)
	writer.WriteHeader(http.StatusNoContent)
	logRequestTrace(trace, http.StatusNoContent, "", "")
}

func (server *Server) requireLocalIdentityHTTP(writer http.ResponseWriter, trace requestTrace) (*localIdentityHTTPService, bool) {
	service := server.localIdentityHTTPService
	if service == nil || !service.enabled || service.repository == nil {
		server.writeLocalIdentityError(writer, trace, localIdentityHTTPDisabled)
		return nil, false
	}
	return service, true
}

func (server *Server) requireUnauthenticatedLocalIdentityRequest(writer http.ResponseWriter, request *http.Request, trace requestTrace) bool {
	requestSession, found := request.Context().Value(localIdentityRequestSessionContextKey{}).(localIdentityRequestSession)
	if found && requestSession.failure == "auth_context_contract_mismatch" {
		server.writeLocalIdentityError(writer, trace, localIdentityAuthenticationFailed)
		return false
	}
	if found && requestSession.sessionID != "" {
		server.writeLocalIdentityError(writer, trace, localIdentityAlreadyAuthenticated)
		return false
	}
	return true
}

func requireLocalIdentityRequestSession(request *http.Request) (localIdentityRequestSession, bool) {
	requestSession, ok := request.Context().Value(localIdentityRequestSessionContextKey{}).(localIdentityRequestSession)
	return requestSession, ok && requestSession.failure == "" && requestSession.sessionID != "" && requestSession.userID != ""
}

func (service *localIdentityHTTPService) requireBootstrapWriteRequest(writer http.ResponseWriter, request *http.Request, trace requestTrace) bool {
	if !service.requireSameOriginJSONRequest(writer, request, trace) {
		return false
	}
	values := request.Header.Values(localIdentityCSRFHeader)
	if len(values) != 1 || subtle.ConstantTimeCompare([]byte(values[0]), []byte(localIdentityBootstrapCSRF)) != 1 {
		writeLocalIdentityHTTPError(writer, trace, localIdentityCSRFInvalid)
		return false
	}
	return true
}

func (service *localIdentityHTTPService) requireAuthenticatedWriteRequest(
	writer http.ResponseWriter,
	request *http.Request,
	trace requestTrace,
	expected string,
	csrfCookieValid bool,
) bool {
	if !service.requireSameOriginJSONRequest(writer, request, trace) {
		return false
	}
	headerValues := request.Header.Values(localIdentityCSRFHeader)
	if len(headerValues) != 1 || !csrfCookieValid || expected == "" ||
		subtle.ConstantTimeCompare([]byte(headerValues[0]), []byte(expected)) != 1 {
		writeLocalIdentityHTTPError(writer, trace, localIdentityCSRFInvalid)
		return false
	}
	return true
}

func (service *localIdentityHTTPService) requireSameOriginJSONRequest(writer http.ResponseWriter, request *http.Request, trace requestTrace) bool {
	origins := request.Header.Values("Origin")
	if len(origins) != 1 || strings.TrimSpace(origins[0]) != service.allowedOrigin {
		writeLocalIdentityHTTPError(writer, trace, localIdentityOriginForbidden)
		return false
	}
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeLocalIdentityHTTPError(writer, trace, localIdentityPayloadInvalid)
		return false
	}
	return true
}

func (service *localIdentityHTTPService) buildRegistration(input localIdentityRegistrationRequest, requestID string) (UserAccount, LocalCredential, WebSession, string, error) {
	identifier := strings.TrimSpace(input.LoginIdentifier)
	normalized, err := NormalizeLocalLoginIdentifier(identifier)
	displayName := strings.TrimSpace(input.DisplayName)
	if err != nil || displayName == "" || utf8.RuneCountInString(displayName) > 120 || !validLocalIdentityPassword(input.Password) {
		return UserAccount{}, LocalCredential{}, WebSession{}, "", errLocalIdentityContractMismatch
	}
	now := service.nowUTC()
	userID, err := randomLocalIdentityID("usr_")
	if err != nil {
		return UserAccount{}, LocalCredential{}, WebSession{}, "", err
	}
	credentialID, err := randomLocalIdentityID("cred_")
	if err != nil {
		return UserAccount{}, LocalCredential{}, WebSession{}, "", err
	}
	auditRef := "auth:register:" + requestID
	account := UserAccount{
		SchemaVersion: localIdentitySchemaVersion, UserID: userID, LoginIdentifier: identifier,
		NormalizedLoginIdentifier: normalized, DisplayName: displayName, LifecycleState: localIdentityStateActive,
		RecordVersion: 1, CreatedAt: now, UpdatedAt: now, AuditRef: auditRef,
	}
	credential, err := DeriveLocalCredential(input.Password, credentialID, userID, now, auditRef)
	if err != nil {
		return UserAccount{}, LocalCredential{}, WebSession{}, "", err
	}
	session, rawCredential, err := service.buildWebSessionAt(userID, credentialID, requestID, now)
	return account, credential, session, rawCredential, err
}

func (service *localIdentityHTTPService) authenticatePassword(ctx context.Context, identifier string, password string) (UserAccount, LocalCredential, bool, bool) {
	account, accountErr := service.repository.FindAccountByLoginIdentifier(ctx, identifier)
	if accountErr != nil && !errors.Is(accountErr, errLocalIdentityNotFound) && !errors.Is(accountErr, errLocalIdentityContractMismatch) {
		return UserAccount{}, LocalCredential{}, false, errors.Is(accountErr, errLocalIdentityStoreUnavailable)
	}
	credential := dummyLocalIdentityCredential()
	credentialErr := error(nil)
	if accountErr == nil {
		credential, credentialErr = service.repository.ReadActiveCredential(ctx, account.UserID)
		if credentialErr != nil && !errors.Is(credentialErr, errLocalIdentityNotFound) {
			return UserAccount{}, LocalCredential{}, false, errors.Is(credentialErr, errLocalIdentityStoreUnavailable)
		}
	}
	verified := VerifyLocalPassword(password, credential)
	if accountErr != nil || credentialErr != nil || account.LifecycleState != localIdentityStateActive || !verified {
		return UserAccount{}, LocalCredential{}, false, false
	}
	return account, credential, true, false
}

func (service *localIdentityHTTPService) buildWebSession(userID string, credentialID string, requestID string) (WebSession, string, error) {
	return service.buildWebSessionAt(userID, credentialID, requestID, service.nowUTC())
}

func (service *localIdentityHTTPService) buildWebSessionAt(userID string, credentialID string, requestID string, now time.Time) (WebSession, string, error) {
	sessionID, err := randomLocalIdentityID("ses_")
	if err != nil {
		return WebSession{}, "", err
	}
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return WebSession{}, "", errLocalIdentityStoreUnavailable
	}
	rawCredential := base64.RawURLEncoding.EncodeToString(rawBytes)
	digest, err := DigestWebSessionCredential(rawCredential)
	if err != nil {
		return WebSession{}, "", err
	}
	session := WebSession{
		SchemaVersion: localIdentitySchemaVersion, SessionID: sessionID, UserID: userID,
		credentialDigest: digest[:], AuthenticationMethod: localAuthenticationMethodPassword,
		AuthenticationSourceRef: "credential:" + credentialID, PolicyVersion: localSessionPolicyVersion,
		LifecycleState: localIdentityStateActive, RecordVersion: 1, CreatedAt: now, UpdatedAt: now,
		LastVerifiedAt: now, ExpiresAt: now.Add(service.sessionTTL), AuditRef: "auth:session:" + requestID,
	}
	if !validWebSession(session) {
		return WebSession{}, "", errLocalIdentityContractMismatch
	}
	return session, rawCredential, nil
}

func randomLocalIdentityID(prefix string) (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", errLocalIdentityStoreUnavailable
	}
	return prefix + hex.EncodeToString(randomBytes), nil
}

func validLocalIdentityPassword(password string) bool {
	count := utf8.RuneCountInString(password)
	return utf8.ValidString(password) && count >= 12 && len(password) <= 1024 && !strings.ContainsAny(password, "\x00\r\n")
}

func dummyLocalIdentityCredential() LocalCredential {
	createdAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	return LocalCredential{
		SchemaVersion: localIdentitySchemaVersion, CredentialID: "cred_0000000000000000", UserID: "usr_0000000000000000",
		Algorithm: localPasswordAlgorithmPBKDF2SHA256, PolicyVersion: localPasswordPolicyVersion,
		Iterations: localPasswordIterations, KeyLength: localPasswordKeyLength, salt: make([]byte, 32), derivedKey: make([]byte, 32),
		LifecycleState: localIdentityStateActive, RecordVersion: 1, CreatedAt: createdAt, UpdatedAt: createdAt,
		AuditRef: "auth:dummy-credential",
	}
}

func localIdentityAuthenticationResponse(account UserAccount, session WebSession, returnTo string) localIdentityAuthenticationDocument {
	return localIdentityAuthenticationDocument{
		Account:  localIdentityAccountDocument{UserID: account.UserID, DisplayName: account.DisplayName, LifecycleState: account.LifecycleState},
		Session:  localIdentitySessionDocument{SessionID: session.SessionID, AuthenticationMethod: session.AuthenticationMethod, ExpiresAt: session.ExpiresAt},
		ReturnTo: returnTo,
	}
}

func normalizeLocalIdentityReturnTarget(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "/", nil
	}
	if len(raw) > 512 || !strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "//") || strings.ContainsAny(raw, "\\\x00\r\n") {
		return "", errLocalIdentityContractMismatch
	}
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.IsAbs() || parsed.Host != "" || parsed.Fragment != "" || !strings.HasPrefix(parsed.Path, "/") || strings.HasPrefix(parsed.Path, "//") {
		return "", errLocalIdentityContractMismatch
	}
	return raw, nil
}

func selectedLocalIdentityTenant(request *http.Request) (string, bool) {
	values := request.Header.Values(localIdentityActiveTenantHeader)
	if len(values) == 0 {
		return "", true
	}
	value := strings.TrimSpace(values[0])
	return value, len(values) == 1 && validControlPlaneReadAuthReference(value, false)
}

func localIdentityControlPlaneScopeCandidates() []string {
	scopes := make([]string, 0, len(workspacePermissionAllowlist))
	for scope := range workspacePermissionAllowlist {
		scopes = append(scopes, scope)
	}
	slices.Sort(scopes)
	return scopes
}

func deriveLocalIdentityCSRFToken(rawCredential string) string {
	digest := sha256.Sum256([]byte("radishmind-local-csrf-v1:" + rawCredential))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}

func readSingleCookieValue(request *http.Request, name string) (string, bool) {
	value := ""
	count := 0
	for _, cookie := range request.Cookies() {
		if cookie.Name != name {
			continue
		}
		count++
		value = strings.TrimSpace(cookie.Value)
	}
	return value, count == 1 && value != ""
}

func (service *localIdentityHTTPService) sessionCookieName() string {
	if service.cookieSecure {
		return "__Host-radishmind_session"
	}
	return "radishmind_session_dev"
}

func (service *localIdentityHTTPService) csrfCookieName() string {
	if service.cookieSecure {
		return "__Host-radishmind_csrf"
	}
	return "radishmind_csrf_dev"
}

func (service *localIdentityHTTPService) setAuthenticationCookies(writer http.ResponseWriter, rawCredential string, expiresAt time.Time) {
	maxAge := max(1, int(expiresAt.Sub(service.nowUTC()).Seconds()))
	http.SetCookie(writer, &http.Cookie{Name: service.sessionCookieName(), Value: rawCredential, Path: "/", Expires: expiresAt,
		MaxAge: maxAge, HttpOnly: true, Secure: service.cookieSecure, SameSite: http.SameSiteStrictMode})
	service.setCSRFCookie(writer, deriveLocalIdentityCSRFToken(rawCredential), expiresAt)
}

func (service *localIdentityHTTPService) setCSRFCookie(writer http.ResponseWriter, value string, expiresAt time.Time) {
	maxAge := max(1, int(expiresAt.Sub(service.nowUTC()).Seconds()))
	http.SetCookie(writer, &http.Cookie{Name: service.csrfCookieName(), Value: value, Path: "/", Expires: expiresAt,
		MaxAge: maxAge, HttpOnly: false, Secure: service.cookieSecure, SameSite: http.SameSiteStrictMode})
}

func (service *localIdentityHTTPService) clearAuthenticationCookies(writer http.ResponseWriter) {
	expiresAt := time.Unix(1, 0).UTC()
	for _, cookie := range []http.Cookie{
		{Name: service.sessionCookieName(), Path: "/", Expires: expiresAt, MaxAge: -1, HttpOnly: true, Secure: service.cookieSecure, SameSite: http.SameSiteStrictMode},
		{Name: service.csrfCookieName(), Path: "/", Expires: expiresAt, MaxAge: -1, HttpOnly: false, Secure: service.cookieSecure, SameSite: http.SameSiteStrictMode},
	} {
		http.SetCookie(writer, &cookie)
	}
}

func (service *localIdentityHTTPService) nowUTC() time.Time {
	if service.now == nil {
		return time.Now().UTC()
	}
	return service.now().UTC()
}

func (server *Server) writeLocalIdentityError(writer http.ResponseWriter, trace requestTrace, code string) {
	writeLocalIdentityHTTPError(writer, trace, code)
}

func writeLocalIdentityHTTPError(writer http.ResponseWriter, trace requestTrace, code string) {
	status, errorType, message := localIdentityHTTPErrorDefinition(code)
	writeTraceHeaders(writer, trace)
	writeJSON(writer, status, errorDocument{Error: errorBody{
		Message: message, Type: errorType, Code: code, RequestID: trace.requestID,
		Route: trace.route, FailureBoundary: "local_identity",
	}})
	logRequestTrace(trace, status, code, "local_identity")
}

func localIdentityHTTPErrorDefinition(code string) (int, string, string) {
	switch code {
	case localIdentityHTTPDisabled:
		return http.StatusForbidden, "configuration_error", "local identity HTTP is disabled"
	case localIdentityOriginForbidden:
		return http.StatusForbidden, "authentication_error", "request origin is not allowed"
	case localIdentityCSRFInvalid:
		return http.StatusForbidden, "authentication_error", "CSRF validation failed"
	case localIdentityPayloadInvalid:
		return http.StatusBadRequest, "invalid_request_error", "local identity request is invalid"
	case localIdentityReturnTargetInvalid:
		return http.StatusBadRequest, "invalid_request_error", "return target must be a local relative path"
	case localIdentityAccountConflict:
		return http.StatusConflict, "invalid_request_error", "account identifier is already registered"
	case localIdentityAuthenticationFailed:
		return http.StatusUnauthorized, "authentication_error", "login identifier or password is invalid"
	case localIdentityAuthenticationRequired:
		return http.StatusUnauthorized, "authentication_error", "an active local session is required"
	case localIdentityAlreadyAuthenticated:
		return http.StatusConflict, "authentication_error", "the request already has an active local session"
	case localIdentitySessionOwnershipDenied:
		return http.StatusForbidden, "permission_error", "the session cannot be revoked by this actor"
	default:
		return http.StatusServiceUnavailable, "service_unavailable_error", "local identity service is unavailable"
	}
}
