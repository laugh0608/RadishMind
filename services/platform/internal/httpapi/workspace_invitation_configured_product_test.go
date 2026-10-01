package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
)

func TestWorkspaceInvitationConfiguredSQLiteProductRestart(t *testing.T) {
	cfg := aggregateSQLiteDevServerConfig(filepath.Join(t.TempDir(), "invitation-product.sqlite"))
	cfg.ControlPlaneReadAuthMode = localIdentityAuthMode
	cfg.LocalIdentityDevHTTPEnabled = true
	cfg.LocalIdentityAllowedOrigin = localIdentityHTTPTestOrigin
	cfg.LocalIdentitySessionTTL = time.Hour
	server, fixture := newWorkspaceInvitationConfiguredProduct(t, cfg)
	if _, ok := server.localIdentityRepository.(*sqliteLocalIdentityRepository); !ok {
		t.Fatal("configured product did not select SQLite identity owner")
	}
	creation := runWorkspaceInvitationHTTPVerticalChain(t, fixture)
	server.Close()
	restarted, err := NewServerWithError(cfg, Options{TestOnly: true})
	if err != nil {
		t.Fatalf("restart configured SQLite server: %v", err)
	}
	defer restarted.Close()
	assertWorkspaceInvitationProductRestart(t, restarted, fixture, creation)
}

func workspaceInvitationProductConfig() config.Config {
	return config.Config{
		Provider: "mock", ControlPlaneReadDevAuthEnabled: true, ControlPlaneReadAuthMode: localIdentityAuthMode,
		ControlPlaneReadDatabaseTimeout: time.Second,
		LocalIdentityDevHTTPEnabled:     true, LocalIdentityAllowedOrigin: localIdentityHTTPTestOrigin,
		LocalIdentityCookieSecure: false, LocalIdentitySessionTTL: time.Hour,
	}
}

func newWorkspaceInvitationConfiguredProduct(t *testing.T, cfg config.Config) (*Server, localIdentityAdministrationHTTPFixture) {
	t.Helper()
	server, err := NewServerWithError(cfg, Options{TestOnly: true})
	if err != nil {
		t.Fatalf("construct configured invitation product: %v", err)
	}
	t.Cleanup(server.Close)
	now := time.Now().UTC()
	identity := &localIdentityHTTPTestFixture{server: server, repository: server.localIdentityRepository,
		service: server.localIdentityHTTPService, handler: server.httpServer.Handler, now: &now}
	_, adminCookies, administrator := registerLocalIdentityHTTPTestAccount(t, identity,
		"product-admin@example.test", "administrator password with enough entropy")
	_, targetCookies, target := registerLocalIdentityHTTPTestAccount(t, identity,
		"product-claimant@example.test", "claimant password with enough entropy")
	bootstrap, err := server.localIdentityAdministrationService.BootstrapWorkspaceAdministrator(context.Background(),
		LocalIdentityBootstrapWorkspaceAdministratorInput{TenantRef: "tenant_demo", WorkspaceID: "workspace_demo",
			UserID: administrator.Account.UserID, AuditRef: "audit:invitation-product-bootstrap"})
	if err != nil {
		t.Fatalf("bootstrap product administrator: %v", err)
	}
	logout := localIdentityHTTPJSONRequest(t, http.MethodPost, localIdentityLogoutRoute, map[string]any{}, targetCookies)
	logout.Header.Set(localIdentityCSRFHeader, localIdentityCookieValue(t, targetCookies, identity.service.csrfCookieName()))
	logoutResponse := httptest.NewRecorder()
	identity.handler.ServeHTTP(logoutResponse, logout)
	if logoutResponse.Code != http.StatusNoContent {
		t.Fatalf("claimant logout: %d", logoutResponse.Code)
	}
	login := localIdentityHTTPJSONRequest(t, http.MethodPost, localIdentityLoginRoute, map[string]any{
		"login_identifier": "product-claimant@example.test", "password": "claimant password with enough entropy",
	}, nil)
	loginResponse := httptest.NewRecorder()
	identity.handler.ServeHTTP(loginResponse, login)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("claimant login: %d", loginResponse.Code)
	}
	targetCookies = activeLocalIdentityCookies(loginResponse.Result().Cookies())
	return server, localIdentityAdministrationHTTPFixture{identity: identity, administrator: administrator,
		adminCookies: adminCookies, target: target, targetCookies: targetCookies, bootstrap: bootstrap}
}

func assertWorkspaceInvitationProductRestart(t *testing.T, server *Server,
	fixture localIdentityAdministrationHTTPFixture, creation workspaceInvitationCreationHTTPResponse) {
	t.Helper()
	fixture.identity.server = server
	fixture.identity.repository = server.localIdentityRepository
	fixture.identity.service = server.localIdentityHTTPService
	fixture.identity.handler = server.httpServer.Handler
	if _, err := server.localIdentityRepository.AuthorizeWorkspace(context.Background(), fixture.target.Account.UserID,
		"tenant_demo", "workspace_demo", []string{"applications:write", "workflow_runs:execute"}, time.Now().UTC()); err != nil {
		t.Fatalf("restart lost claimed permission: %v", err)
	}
	path := strings.Replace(workspaceInvitationAdminListRoute, "{workspace_id}", "workspace_demo", 1)
	listed := fixture.request(t, http.MethodGet, path+"?effective_state=claimed", nil, fixture.adminCookies)
	if listed.Code != http.StatusOK {
		t.Fatalf("restart list status: %d", listed.Code)
	}
	var page workspaceInvitationListHTTPResponse
	decodeLocalIdentityHTTPResponse(t, listed, &page)
	if len(page.Invitations) != 1 || page.Invitations[0].InvitationID != creation.Invitation.InvitationID ||
		page.Invitations[0].MembershipID == "" || page.Invitations[0].AssignmentID == "" {
		t.Fatal("restart lost canonical claimed invitation references")
	}
	assertWorkspaceInvitationHTTPSafePayload(t, listed.Body.String(), false, creation.InvitationCode)
	replay := workspaceInvitationClaimantResponse(t, fixture, http.MethodPost, workspaceInvitationClaimRoute,
		map[string]any{"invitation_code": creation.InvitationCode, "expected_record_version": 1, "confirmed": true},
		fixture.targetCookies, "tenant_demo")
	assertWorkspaceInvitationHTTPError(t, replay, http.StatusConflict, WorkspaceInvitationFailureNotClaimable, "discard_terminal_invitation")
}
