package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	sqlitelocalidentitymigrations "radishmind.local/services/platform/migrations/sqlite/local_identity_records"
)

const localIdentityHTTPTestOrigin = "http://127.0.0.1:4000"

type localIdentityHTTPTestFixture struct {
	server     *Server
	repository localIdentityRepository
	service    *localIdentityHTTPService
	handler    http.Handler
	now        *time.Time
}

func TestServerWiresLocalIdentityHTTPModeAndCredentialedCORS(t *testing.T) {
	cfg := config.Config{
		ControlPlaneReadDevAuthEnabled: true, ControlPlaneReadAuthMode: localIdentityAuthMode,
		LocalIdentityDevHTTPEnabled: true, LocalIdentityAllowedOrigin: localIdentityHTTPTestOrigin,
		LocalIdentityCookieSecure: false, LocalIdentitySessionTTL: time.Hour, Provider: "mock",
	}
	server, err := NewServerWithError(cfg, Options{BuildVersion: "local-identity-http-test", TestOnly: true})
	if err != nil {
		t.Fatalf("construct local identity HTTP server: %v", err)
	}
	defer server.Close()

	preflight := httptest.NewRequest(http.MethodOptions, localIdentityRegisterRoute, nil)
	preflight.Header.Set("Origin", localIdentityHTTPTestOrigin)
	preflight.Header.Set("Access-Control-Request-Method", http.MethodPost)
	preflight.Header.Set("Access-Control-Request-Headers", "Content-Type, "+localIdentityCSRFHeader)
	preflightResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(preflightResponse, preflight)
	if preflightResponse.Code != http.StatusNoContent ||
		preflightResponse.Header().Get("Access-Control-Allow-Credentials") != "true" ||
		!strings.Contains(preflightResponse.Header().Get("Access-Control-Allow-Headers"), localIdentityCSRFHeader) {
		t.Fatalf("credentialed local identity CORS drifted: status=%d headers=%v", preflightResponse.Code, preflightResponse.Header())
	}
	legacyOriginPreflight := httptest.NewRequest(http.MethodOptions, localIdentityCurrentSessionRoute, nil)
	legacyOriginPreflight.Header.Set("Origin", "http://localhost:4000")
	legacyOriginPreflight.Header.Set("Access-Control-Request-Method", http.MethodGet)
	legacyOriginResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(legacyOriginResponse, legacyOriginPreflight)
	if legacyOriginResponse.Code != http.StatusNoContent ||
		legacyOriginResponse.Header().Get("Access-Control-Allow-Origin") != "http://localhost:4000" ||
		legacyOriginResponse.Header().Get("Access-Control-Allow-Credentials") != "" {
		t.Fatalf("legacy console origin must remain non-credentialed: status=%d headers=%v", legacyOriginResponse.Code, legacyOriginResponse.Header())
	}

	register := localIdentityHTTPJSONRequest(t, http.MethodPost, localIdentityRegisterRoute, map[string]any{
		"login_identifier": "wired@example.com", "display_name": "Wired User", "password": "a sufficiently long password",
	}, nil)
	registerResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(registerResponse, register)
	if registerResponse.Code != http.StatusCreated || server.localIdentityHTTPService == nil ||
		server.localIdentityAdministrationService == nil || server.workspaceMembershipProvider == nil {
		t.Fatalf("server local identity wiring failed: status=%d body=%s", registerResponse.Code, registerResponse.Body.String())
	}
}

func TestLocalIdentityHTTPSCookieUsesHostPrefix(t *testing.T) {
	service := newLocalIdentityHTTPService(config.Config{
		LocalIdentityDevHTTPEnabled: true, LocalIdentityAllowedOrigin: "https://mind.example.test",
		LocalIdentityCookieSecure: true, LocalIdentitySessionTTL: time.Hour,
	}, newMemoryLocalIdentityRepository())
	response := httptest.NewRecorder()
	service.setAuthenticationCookies(response, "a-session-credential-with-at-least-thirty-two-bytes", time.Now().Add(time.Hour))
	cookies := response.Result().Cookies()
	assertLocalIdentityCookiePolicy(t, cookies, true)
	for _, cookie := range cookies {
		if !strings.HasPrefix(cookie.Name, "__Host-") {
			t.Fatalf("secure identity cookie lacks __Host- prefix: %#v", cookie)
		}
	}
}

func newLocalIdentityHTTPTestFixture(repository localIdentityRepository, now time.Time) *localIdentityHTTPTestFixture {
	cfg := config.Config{
		ControlPlaneReadDevAuthEnabled: true, ControlPlaneReadAuthMode: localIdentityAuthMode,
		LocalIdentityDevHTTPEnabled: true, LocalIdentityStoreMode: localIdentityStoreModeMemoryDev,
		LocalIdentityAllowedOrigin: localIdentityHTTPTestOrigin, LocalIdentityCookieSecure: false,
		LocalIdentitySessionTTL: time.Hour,
	}
	service := newLocalIdentityHTTPService(cfg, repository)
	service.now = func() time.Time { return now }
	var administration *localIdentityAdministrationService
	if administrationRepository, ok := repository.(localIdentityAdministrationRepository); ok {
		administration = newLocalIdentityAdministrationService(administrationRepository)
		administration.now = func() time.Time { return now }
	}
	server := &Server{
		config: cfg, localIdentityHTTPService: service, localIdentityAdministrationService: administration,
		workspaceMembershipProvider: newLocalWorkspaceMembershipProvider(repository),
	}
	mux := http.NewServeMux()
	registerLocalIdentityHTTPRoutes(mux, server)
	handler := withLocalIdentitySessionAuthentication(
		withControlPlaneReadAuthenticator(mux, &controlPlaneReadAuthenticator{mode: localIdentityAuthMode}),
		service,
	)
	return &localIdentityHTTPTestFixture{server: server, repository: repository, service: service, handler: handler, now: &now}
}

func (fixture *localIdentityHTTPTestFixture) setNow(now time.Time) {
	*fixture.now = now
	fixture.service.now = func() time.Time { return *fixture.now }
	if fixture.server.localIdentityAdministrationService != nil {
		fixture.server.localIdentityAdministrationService.now = func() time.Time { return *fixture.now }
	}
}

func TestLocalIdentityHTTPRegisterCurrentLogoutAndRelogin(t *testing.T) {
	now := time.Date(2026, 8, 19, 13, 0, 0, 0, time.UTC)
	fixture := newLocalIdentityHTTPTestFixture(newMemoryLocalIdentityRepository(), now)
	password := "correct horse battery staple"
	register := localIdentityHTTPJSONRequest(t, http.MethodPost, localIdentityRegisterRoute, map[string]any{
		"login_identifier": " Alice.Example@Example.COM ", "display_name": "Alice", "password": password,
		"return_to": "/workspace?tab=apps",
	}, nil)
	registerResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(registerResponse, register)
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("register failed: status=%d body=%s", registerResponse.Code, registerResponse.Body.String())
	}
	if strings.Contains(registerResponse.Body.String(), password) || strings.Contains(registerResponse.Body.String(), "Alice.Example@Example.COM") {
		t.Fatalf("registration response leaked credential or login identifier: %s", registerResponse.Body.String())
	}
	var registered localIdentityAuthenticationDocument
	decodeLocalIdentityHTTPResponse(t, registerResponse, &registered)
	if registered.Account.UserID == "" || registered.Account.DisplayName != "Alice" || registered.ReturnTo != "/workspace?tab=apps" {
		t.Fatalf("unexpected registration response: %#v", registered)
	}
	cookies := activeLocalIdentityCookies(registerResponse.Result().Cookies())
	assertLocalIdentityCookiePolicy(t, cookies, false)

	current := httptest.NewRequest(http.MethodGet, localIdentityCurrentSessionRoute, nil)
	addLocalIdentityCookies(current, cookies)
	currentResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(currentResponse, current)
	if currentResponse.Code != http.StatusOK || !strings.Contains(currentResponse.Body.String(), registered.Account.UserID) {
		t.Fatalf("current session failed: status=%d body=%s", currentResponse.Code, currentResponse.Body.String())
	}

	csrf := localIdentityCookieValue(t, cookies, fixture.service.csrfCookieName())
	logout := localIdentityHTTPJSONRequest(t, http.MethodPost, localIdentityLogoutRoute, map[string]any{}, cookies)
	logout.Header.Set(localIdentityCSRFHeader, csrf)
	logoutResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(logoutResponse, logout)
	if logoutResponse.Code != http.StatusNoContent {
		t.Fatalf("logout failed: status=%d body=%s", logoutResponse.Code, logoutResponse.Body.String())
	}

	afterLogout := httptest.NewRequest(http.MethodGet, localIdentityCurrentSessionRoute, nil)
	addLocalIdentityCookies(afterLogout, cookies)
	afterLogoutResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(afterLogoutResponse, afterLogout)
	assertLocalIdentityError(t, afterLogoutResponse, http.StatusUnauthorized, localIdentityAuthenticationRequired)

	login := localIdentityHTTPJSONRequest(t, http.MethodPost, localIdentityLoginRoute, map[string]any{
		"login_identifier": "alice.example@example.com", "password": password, "return_to": "/",
	}, nil)
	loginResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(loginResponse, login)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("relogin failed: status=%d body=%s", loginResponse.Code, loginResponse.Body.String())
	}
	if strings.Contains(loginResponse.Body.String(), password) || len(activeLocalIdentityCookies(loginResponse.Result().Cookies())) != 2 {
		t.Fatalf("login response or cookie contract drifted: %s", loginResponse.Body.String())
	}
}

func TestLocalIdentityHTTPNegativeSecurityMatrix(t *testing.T) {
	now := time.Date(2026, 8, 19, 14, 0, 0, 0, time.UTC)
	fixture := newLocalIdentityHTTPTestFixture(newMemoryLocalIdentityRepository(), now)
	password := "a sufficiently long password"

	missingOrigin := localIdentityHTTPJSONRequest(t, http.MethodPost, localIdentityRegisterRoute, map[string]any{
		"login_identifier": "alice@example.com", "display_name": "Alice", "password": password,
	}, nil)
	missingOrigin.Header.Del("Origin")
	missingOriginResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(missingOriginResponse, missingOrigin)
	assertLocalIdentityError(t, missingOriginResponse, http.StatusForbidden, localIdentityOriginForbidden)

	missingCSRF := localIdentityHTTPJSONRequest(t, http.MethodPost, localIdentityRegisterRoute, map[string]any{
		"login_identifier": "alice@example.com", "display_name": "Alice", "password": password,
	}, nil)
	missingCSRF.Header.Del(localIdentityCSRFHeader)
	missingCSRFResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(missingCSRFResponse, missingCSRF)
	assertLocalIdentityError(t, missingCSRFResponse, http.StatusForbidden, localIdentityCSRFInvalid)

	unsafeReturn := localIdentityHTTPJSONRequest(t, http.MethodPost, localIdentityRegisterRoute, map[string]any{
		"login_identifier": "alice@example.com", "display_name": "Alice", "password": password,
		"return_to": "//evil.example/steal",
	}, nil)
	unsafeReturnResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(unsafeReturnResponse, unsafeReturn)
	assertLocalIdentityError(t, unsafeReturnResponse, http.StatusBadRequest, localIdentityReturnTargetInvalid)

	registerResponse, cookies, registered := registerLocalIdentityHTTPTestAccount(t, fixture, "alice@example.com", password)
	if registerResponse.Code != http.StatusCreated {
		t.Fatal("registration fixture did not succeed")
	}
	duplicate := localIdentityHTTPJSONRequest(t, http.MethodPost, localIdentityRegisterRoute, map[string]any{
		"login_identifier": "ALICE@example.com", "display_name": "Other", "password": "another long password",
	}, nil)
	duplicateResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(duplicateResponse, duplicate)
	assertLocalIdentityError(t, duplicateResponse, http.StatusConflict, localIdentityAccountConflict)

	unknownLoginResponse := performLocalIdentityLogin(t, fixture, "nobody@example.com", "wrong but long password")
	wrongPasswordResponse := performLocalIdentityLogin(t, fixture, "alice@example.com", "wrong but long password")
	assertLocalIdentityError(t, unknownLoginResponse, http.StatusUnauthorized, localIdentityAuthenticationFailed)
	assertLocalIdentityError(t, wrongPasswordResponse, http.StatusUnauthorized, localIdentityAuthenticationFailed)

	csrf := localIdentityCookieValue(t, cookies, fixture.service.csrfCookieName())
	missingMutationCSRF := localIdentityHTTPJSONRequest(t, http.MethodPost, localIdentityLogoutRoute, map[string]any{}, cookies)
	missingMutationCSRF.Header.Del(localIdentityCSRFHeader)
	missingMutationCSRFResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(missingMutationCSRFResponse, missingMutationCSRF)
	assertLocalIdentityError(t, missingMutationCSRFResponse, http.StatusForbidden, localIdentityCSRFInvalid)

	wrongSessionRevoke := localIdentityHTTPJSONRequest(t, http.MethodPost, "/v1/auth/sessions/ses_0000000000000000/revoke", map[string]any{}, cookies)
	wrongSessionRevoke.Header.Set(localIdentityCSRFHeader, csrf)
	wrongSessionRevokeResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(wrongSessionRevokeResponse, wrongSessionRevoke)
	assertLocalIdentityError(t, wrongSessionRevokeResponse, http.StatusForbidden, localIdentitySessionOwnershipDenied)

	account, err := fixture.repository.ReadAccount(context.Background(), registered.Account.UserID)
	if err != nil {
		t.Fatalf("read registered account: %v", err)
	}
	if _, err := fixture.repository.DisableAccount(context.Background(), account.UserID, account.RecordVersion, now.Add(time.Minute), "auth:test-disable"); err != nil {
		t.Fatalf("disable account: %v", err)
	}
	disabledLoginResponse := performLocalIdentityLogin(t, fixture, "alice@example.com", password)
	assertLocalIdentityError(t, disabledLoginResponse, http.StatusUnauthorized, localIdentityAuthenticationFailed)

	disabledSession := httptest.NewRequest(http.MethodGet, localIdentityCurrentSessionRoute, nil)
	addLocalIdentityCookies(disabledSession, cookies)
	disabledSessionResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(disabledSessionResponse, disabledSession)
	assertLocalIdentityError(t, disabledSessionResponse, http.StatusUnauthorized, localIdentityAuthenticationRequired)
}

func TestLocalIdentityHTTPSessionExpiryRevokeAndNoFallback(t *testing.T) {
	now := time.Date(2026, 8, 19, 15, 0, 0, 0, time.UTC)
	fixture := newLocalIdentityHTTPTestFixture(newMemoryLocalIdentityRepository(), now)
	_, cookies, registered := registerLocalIdentityHTTPTestAccount(t, fixture, "expiry@example.com", "a sufficiently long password")

	fixture.setNow(now.Add(2 * time.Hour))
	expired := httptest.NewRequest(http.MethodGet, localIdentityCurrentSessionRoute, nil)
	addLocalIdentityCookies(expired, cookies)
	expiredResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(expiredResponse, expired)
	assertLocalIdentityError(t, expiredResponse, http.StatusUnauthorized, localIdentityAuthenticationRequired)

	fixture = newLocalIdentityHTTPTestFixture(newMemoryLocalIdentityRepository(), now)
	_, cookies, registered = registerLocalIdentityHTTPTestAccount(t, fixture, "revoke@example.com", "a sufficiently long password")
	rawSession := localIdentityCookieValue(t, cookies, fixture.service.sessionCookieName())
	digest, _ := DigestWebSessionCredential(rawSession)
	session, _, err := fixture.repository.ResolveWebSession(context.Background(), digest, now)
	if err != nil {
		t.Fatalf("resolve issued session: %v", err)
	}
	if _, err := fixture.repository.RevokeWebSession(context.Background(), session.SessionID, session.RecordVersion, now.Add(time.Minute), "auth:test-revoke"); err != nil {
		t.Fatalf("revoke issued session: %v", err)
	}
	revoked := httptest.NewRequest(http.MethodGet, localIdentityCurrentSessionRoute, nil)
	addLocalIdentityCookies(revoked, cookies)
	revokedResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(revokedResponse, revoked)
	assertLocalIdentityError(t, revokedResponse, http.StatusUnauthorized, localIdentityAuthenticationRequired)

	protected := withLocalIdentitySessionAuthentication(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		auth, failure, _ := authorizeControlPlaneReadRequest(request, "", "applications:read")
		writeJSON(writer, http.StatusOK, map[string]string{"failure": failure, "subject": auth.SubjectBinding})
	}), fixture.service)
	conflicting := httptest.NewRequest(http.MethodGet, "/protected", nil)
	addLocalIdentityCookies(conflicting, cookies)
	conflicting.Header.Set(controlPlaneReadDevIdentityHeader, "dev:test")
	conflicting.Header.Set(controlPlaneReadDevTenantHeader, "tenant_demo")
	conflicting.Header.Set(controlPlaneReadDevSubjectHeader, "user:dev")
	conflicting.Header.Set(controlPlaneReadDevScopesHeader, "applications:read")
	conflictingResponse := httptest.NewRecorder()
	protected.ServeHTTP(conflictingResponse, conflicting)
	if !strings.Contains(conflictingResponse.Body.String(), "auth_context_contract_mismatch") || strings.Contains(conflictingResponse.Body.String(), registered.Account.UserID) {
		t.Fatalf("dev headers became a local session fallback: %s", conflictingResponse.Body.String())
	}
}

func TestLocalIdentityHTTPAccountProfileAndExternalIdentityRevoke(t *testing.T) {
	now := time.Date(2026, 8, 19, 16, 30, 0, 0, time.UTC)
	fixture := newLocalIdentityHTTPTestFixture(newMemoryLocalIdentityRepository(), now)
	_, cookies, registered := registerLocalIdentityHTTPTestAccount(
		t, fixture, "profile@example.com", "a sufficiently long password",
	)
	binding := localIdentityTestBinding(
		"xid_profileaaaaaaaaa", registered.Account.UserID,
		"https://issuer.secret.example.com/oidc", "upstream-subject-must-not-leak",
	)
	if err := fixture.repository.BindExternalIdentity(t.Context(), binding); err != nil {
		t.Fatalf("bind profile external identity: %v", err)
	}
	membership := localIdentityTestMembership(registered.Account.UserID)
	role := localIdentityTestRoleAssignment(registered.Account.UserID)
	if err := fixture.repository.CreateWorkspaceMembership(t.Context(), membership); err != nil {
		t.Fatalf("create profile membership: %v", err)
	}
	if err := fixture.repository.CreateRoleAssignment(t.Context(), role); err != nil {
		t.Fatalf("create profile role: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, localIdentityCurrentAccountRoute, nil)
	addLocalIdentityCookies(request, cookies)
	response := httptest.NewRecorder()
	fixture.handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("read account profile: status=%d body=%s", response.Code, response.Body.String())
	}
	var document localIdentityAccountProfileDocument
	decodeLocalIdentityHTTPResponse(t, response, &document)
	if document.Account.UserID != registered.Account.UserID || len(document.ExternalIdentities) != 1 ||
		document.ExternalIdentities[0].ProviderRef != "radish_oidc" || !document.ExternalIdentities[0].CanRevoke ||
		len(document.RoleAssignments) != 1 || len(document.WorkspaceMemberships) != 1 ||
		!document.Capabilities.HasActiveLocalCredential || !document.Capabilities.RecentAuthentication {
		t.Fatalf("unexpected account profile: %#v", document)
	}
	for _, forbidden := range []string{
		"profile@example.com", binding.Issuer, binding.Subject, binding.AuditRef, "login_identifier", "subject", "issuer",
	} {
		if strings.Contains(response.Body.String(), forbidden) {
			t.Fatalf("account profile leaked %q: %s", forbidden, response.Body.String())
		}
	}

	csrf := localIdentityCookieValue(t, cookies, fixture.service.csrfCookieName())
	wrongOwner := localIdentityHTTPJSONRequest(
		t, http.MethodPost, "/v1/auth/external-identities/xid_otheraaaaaaaaaaa/revoke",
		map[string]any{"expected_record_version": 1}, cookies,
	)
	wrongOwner.Header.Set(localIdentityCSRFHeader, csrf)
	wrongOwnerResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(wrongOwnerResponse, wrongOwner)
	assertLocalIdentityError(t, wrongOwnerResponse, http.StatusForbidden, localIdentityBindingOwnershipDenied)

	revoke := localIdentityHTTPJSONRequest(
		t, http.MethodPost, "/v1/auth/external-identities/"+binding.BindingID+"/revoke",
		map[string]any{"expected_record_version": 1}, cookies,
	)
	revoke.Header.Set(localIdentityCSRFHeader, csrf)
	revokeResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(revokeResponse, revoke)
	if revokeResponse.Code != http.StatusNoContent {
		t.Fatalf("revoke external identity: status=%d body=%s", revokeResponse.Code, revokeResponse.Body.String())
	}
	if _, err := fixture.repository.ResolveExternalIdentity(t.Context(), binding.Issuer, binding.Subject); !errors.Is(err, errLocalIdentityNotFound) {
		t.Fatalf("revoked external identity still resolves: %v", err)
	}
}

func TestSQLiteLocalIdentityHTTPSessionSurvivesRestart(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "identity", "radishmind.db")
	now := time.Date(2026, 8, 19, 16, 0, 0, 0, time.UTC)
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath, Migrations: sqlitelocalidentitymigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open SQLite identity runtime: %v", err)
	}
	fixture := newLocalIdentityHTTPTestFixture(newSQLiteLocalIdentityRepository(runtime.DB()), now)
	_, cookies, _ := registerLocalIdentityHTTPTestAccount(t, fixture, "restart@example.com", "a sufficiently long password")
	if err := runtime.Close(); err != nil {
		t.Fatalf("close first SQLite runtime: %v", err)
	}

	restarted, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath, Migrations: sqlitelocalidentitymigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("restart SQLite identity runtime: %v", err)
	}
	defer func() { _ = restarted.Close() }()
	restartedFixture := newLocalIdentityHTTPTestFixture(newSQLiteLocalIdentityRepository(restarted.DB()), now.Add(10*time.Minute))
	current := httptest.NewRequest(http.MethodGet, localIdentityCurrentSessionRoute, nil)
	addLocalIdentityCookies(current, cookies)
	currentResponse := httptest.NewRecorder()
	restartedFixture.handler.ServeHTTP(currentResponse, current)
	if currentResponse.Code != http.StatusOK {
		t.Fatalf("SQLite restart did not restore session: status=%d body=%s", currentResponse.Code, currentResponse.Body.String())
	}
}

func TestLocalIdentityMiddlewareRestoresActorThenLocalMembership(t *testing.T) {
	now := time.Date(2026, 8, 19, 17, 0, 0, 0, time.UTC)
	fixture := newLocalIdentityHTTPTestFixture(newMemoryLocalIdentityRepository(), now)
	_, cookies, registered := registerLocalIdentityHTTPTestAccount(t, fixture, "member@example.com", "a sufficiently long password")
	membership := WorkspaceMembership{
		SchemaVersion: localIdentitySchemaVersion, MembershipID: "mbr_aaaaaaaaaaaaaaaa", UserID: registered.Account.UserID,
		TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", LifecycleState: localIdentityStateActive,
		RecordVersion: 1, CreatedAt: now, UpdatedAt: now, AuditRef: "auth:test-membership",
	}
	if err := fixture.repository.CreateWorkspaceMembership(context.Background(), membership); err != nil {
		t.Fatalf("create local membership: %v", err)
	}
	role := LocalRoleAssignment{
		SchemaVersion: localIdentitySchemaVersion, AssignmentID: "rla_aaaaaaaaaaaaaaaa", UserID: registered.Account.UserID,
		TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", RoleKey: "workspace_reader",
		PermissionGrants: []string{"applications:read"}, LifecycleState: localIdentityStateActive,
		RecordVersion: 1, CreatedAt: now, UpdatedAt: now, AuditRef: "auth:test-role",
	}
	if err := fixture.repository.CreateRoleAssignment(context.Background(), role); err != nil {
		t.Fatalf("create local role: %v", err)
	}

	protected := withLocalIdentitySessionAuthentication(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		auth, failure, status := fixture.server.authorizeWorkspaceScopedPermissions(request, "applications:read")
		writeJSON(writer, status, map[string]any{
			"failure": failure, "subject": auth.SubjectBinding,
			"membership_verified": auth.ResourceBinding.WorkspaceMembershipVerified,
			"cookie_header":       request.Header.Get("Cookie"),
		})
	}), fixture.service)
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	addLocalIdentityCookies(request, cookies)
	request.Header.Set(localIdentityActiveTenantHeader, "tenant_demo")
	request.Header.Set(activeWorkspaceHeader, "workspace_demo")
	response := httptest.NewRecorder()
	protected.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "user:"+registered.Account.UserID) ||
		!strings.Contains(response.Body.String(), `"membership_verified":true`) ||
		!strings.Contains(response.Body.String(), `"cookie_header":""`) {
		t.Fatalf("local actor/membership restoration failed: status=%d body=%s", response.Code, response.Body.String())
	}
}

func registerLocalIdentityHTTPTestAccount(
	t *testing.T,
	fixture *localIdentityHTTPTestFixture,
	identifier string,
	password string,
) (*httptest.ResponseRecorder, []*http.Cookie, localIdentityAuthenticationDocument) {
	t.Helper()
	request := localIdentityHTTPJSONRequest(t, http.MethodPost, localIdentityRegisterRoute, map[string]any{
		"login_identifier": identifier, "display_name": "Test User", "password": password,
	}, nil)
	response := httptest.NewRecorder()
	fixture.handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("register test account: status=%d body=%s", response.Code, response.Body.String())
	}
	var document localIdentityAuthenticationDocument
	decodeLocalIdentityHTTPResponse(t, response, &document)
	return response, activeLocalIdentityCookies(response.Result().Cookies()), document
}

func performLocalIdentityLogin(t *testing.T, fixture *localIdentityHTTPTestFixture, identifier string, password string) *httptest.ResponseRecorder {
	t.Helper()
	request := localIdentityHTTPJSONRequest(t, http.MethodPost, localIdentityLoginRoute, map[string]any{
		"login_identifier": identifier, "password": password,
	}, nil)
	response := httptest.NewRecorder()
	fixture.handler.ServeHTTP(response, request)
	return response
}

func localIdentityHTTPJSONRequest(t *testing.T, method string, path string, body any, cookies []*http.Cookie) *http.Request {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal HTTP request: %v", err)
	}
	request := httptest.NewRequest(method, path, bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", localIdentityHTTPTestOrigin)
	request.Header.Set(localIdentityCSRFHeader, localIdentityBootstrapCSRF)
	addLocalIdentityCookies(request, cookies)
	return request
}

func addLocalIdentityCookies(request *http.Request, cookies []*http.Cookie) {
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
}

func activeLocalIdentityCookies(cookies []*http.Cookie) []*http.Cookie {
	active := make([]*http.Cookie, 0, len(cookies))
	for _, cookie := range cookies {
		if cookie.MaxAge >= 0 {
			active = append(active, cookie)
		}
	}
	return active
}

func localIdentityCookieValue(t *testing.T, cookies []*http.Cookie, name string) string {
	t.Helper()
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	t.Fatalf("cookie %q is missing", name)
	return ""
}

func assertLocalIdentityCookiePolicy(t *testing.T, cookies []*http.Cookie, secure bool) {
	t.Helper()
	if len(cookies) != 2 {
		t.Fatalf("expected session and CSRF cookies, got %#v", cookies)
	}
	for _, cookie := range cookies {
		if cookie.Path != "/" || cookie.SameSite != http.SameSiteStrictMode || cookie.Secure != secure || cookie.Domain != "" {
			t.Fatalf("cookie policy drifted: %#v", cookie)
		}
		if strings.Contains(cookie.Name, "session") && !cookie.HttpOnly {
			t.Fatalf("session cookie is not HttpOnly: %#v", cookie)
		}
		if strings.Contains(cookie.Name, "csrf") && cookie.HttpOnly {
			t.Fatalf("CSRF cookie must remain readable for double-submit: %#v", cookie)
		}
	}
}

func assertLocalIdentityError(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if response.Code != status {
		t.Fatalf("unexpected error status: got=%d want=%d body=%s", response.Code, status, response.Body.String())
	}
	var document errorDocument
	decodeLocalIdentityHTTPResponse(t, response, &document)
	if document.Error.Code != code {
		t.Fatalf("unexpected error code: got=%q want=%q body=%s", document.Error.Code, code, response.Body.String())
	}
}

func decodeLocalIdentityHTTPResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.Unmarshal(response.Body.Bytes(), target); err != nil {
		t.Fatalf("decode HTTP response: %v body=%s", err, response.Body.String())
	}
}
