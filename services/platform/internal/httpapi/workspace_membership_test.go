package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
)

func TestWorkspaceScopedReadAuthorizationDenialsDoNotQueryRepository(t *testing.T) {
	now := time.Now().UTC()
	baseAuth := controlPlaneReadAuthContext{
		AuthMode: controlPlaneReadAuthModeDevHeaders, IdentityContext: "verified:workspace-test",
		TenantBinding: "tenant_demo", SubjectBinding: "subject_demo", ScopeGrants: []string{"applications:read"},
		ResourceBinding: ControlPlaneResourceBinding{TenantRef: "tenant_demo", TenantVerified: true},
		WorkspaceMemberships: []VerifiedWorkspaceMembershipAssertion{{
			TenantRef: "tenant_demo", SubjectRef: "subject_demo", WorkspaceID: "workspace_demo",
			PermissionGrants: []string{"applications:read"}, SourceRef: "membership:test",
			PolicyVersion: workspaceMembershipPolicyVersion, ExpiresAt: now.Add(time.Hour),
		}},
	}
	tests := []struct {
		name          string
		active        string
		mutate        func(*controlPlaneReadAuthContext)
		expectedCode  int
		expectedError string
	}{
		{name: "selection missing", expectedCode: http.StatusBadRequest, expectedError: "workspace_selection_missing"},
		{name: "cross tenant", active: "workspace_demo", mutate: func(auth *controlPlaneReadAuthContext) {
			auth.WorkspaceMemberships[0].TenantRef = "tenant_other"
		}, expectedCode: http.StatusForbidden, expectedError: "workspace_binding_mismatch"},
		{name: "subject mismatch", active: "workspace_demo", mutate: func(auth *controlPlaneReadAuthContext) {
			auth.WorkspaceMemberships[0].SubjectRef = "subject_other"
		}, expectedCode: http.StatusForbidden, expectedError: "workspace_binding_mismatch"},
		{name: "non member", active: "workspace_demo", mutate: func(auth *controlPlaneReadAuthContext) {
			auth.WorkspaceMemberships = nil
		}, expectedCode: http.StatusForbidden, expectedError: "workspace_membership_denied"},
		{name: "workspace mismatch", active: "workspace_other", expectedCode: http.StatusForbidden, expectedError: "workspace_binding_mismatch"},
		{name: "membership expired", active: "workspace_demo", mutate: func(auth *controlPlaneReadAuthContext) {
			auth.WorkspaceMemberships[0].ExpiresAt = now.Add(-time.Second)
		}, expectedCode: http.StatusForbidden, expectedError: "workspace_membership_expired"},
		{name: "permission insufficient", active: "workspace_demo", mutate: func(auth *controlPlaneReadAuthContext) {
			auth.WorkspaceMemberships[0].PermissionGrants = []string{"runs:read"}
		}, expectedCode: http.StatusForbidden, expectedError: "workspace_permission_denied"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := NewServer(config.Config{}, Options{BuildVersion: "test"})
			repository := &recordingControlPlaneReadRepository{}
			server.workspaceControlPlaneReadRepo = repository
			auth := baseAuth
			auth.ScopeGrants = append([]string{}, baseAuth.ScopeGrants...)
			auth.WorkspaceMemberships = append([]VerifiedWorkspaceMembershipAssertion{}, baseAuth.WorkspaceMemberships...)
			if test.mutate != nil {
				test.mutate(&auth)
			}
			request := newControlPlaneReadRequest(http.MethodGet, "/v1/user-workspace/applications", auth)
			request.Header.Del(activeWorkspaceHeader)
			if test.active != "" {
				request.Header.Set(activeWorkspaceHeader, test.active)
			}
			recorder := httptest.NewRecorder()

			server.httpServer.Handler.ServeHTTP(recorder, request)

			envelope := decodeControlPlaneReadEnvelope(t, recorder, test.expectedCode)
			assertControlPlaneReadFailure(t, envelope, test.expectedError)
			if repository.totalCalls != 0 {
				t.Fatalf("workspace denial reached repository %d times", repository.totalCalls)
			}
		})
	}
}

func TestWorkspaceScopedReadExpiredSignedIdentityDoesNotQueryRepository(t *testing.T) {
	privateKey := generateSignedTestPrivateKey(t)
	server := newSignedTestControlPlaneReadServer(t, privateKey)
	repository := &recordingControlPlaneReadRepository{}
	server.workspaceControlPlaneReadRepo = repository
	claims := validSignedTestClaims()
	claims["exp"] = time.Now().Add(-time.Minute).Unix()
	request := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/applications", nil)
	request.Header.Set(activeWorkspaceHeader, "workspace_demo")
	request.Header.Set("Authorization", "Bearer "+signControlPlaneReadTestToken(t, privateKey, "RS256", claims))
	recorder := httptest.NewRecorder()

	server.httpServer.Handler.ServeHTTP(recorder, request)

	envelope := decodeControlPlaneReadEnvelope(t, recorder, http.StatusUnauthorized)
	assertControlPlaneReadFailure(t, envelope, "auth_context_contract_mismatch")
	if repository.totalCalls != 0 {
		t.Fatalf("expired identity reached repository %d times", repository.totalCalls)
	}
}

func TestWorkspaceScopedReadSignedMembershipProjectsWorkspaceToRepository(t *testing.T) {
	privateKey := generateSignedTestPrivateKey(t)
	server := newSignedTestControlPlaneReadServer(t, privateKey)
	repository := &recordingControlPlaneReadRepository{}
	server.workspaceControlPlaneReadRepo = repository
	claims := validSignedTestClaims()
	claims["permissions"] = []string{"radishmind.applications.read"}
	request := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/applications", nil)
	request.Header.Set(activeWorkspaceHeader, "workspace_demo")
	request.Header.Set("Authorization", "Bearer "+signControlPlaneReadTestToken(t, privateKey, "RS256", claims))
	recorder := httptest.NewRecorder()

	server.httpServer.Handler.ServeHTTP(recorder, request)

	envelope := decodeControlPlaneReadEnvelope(t, recorder, http.StatusOK)
	if envelope.FailureCode != nil || repository.applicationCalls != 1 ||
		repository.lastContext.WorkspaceID != "workspace_demo" {
		t.Fatalf("signed membership was not projected to repository: envelope=%#v context=%#v", envelope, repository.lastContext)
	}
}

func TestWorkspaceScopedReadDevMembershipProjectsWorkspaceToRepository(t *testing.T) {
	server := NewServer(
		config.Config{ControlPlaneReadDevAuthEnabled: true},
		Options{BuildVersion: "test"},
	)
	repository := &recordingControlPlaneReadRepository{}
	server.workspaceControlPlaneReadRepo = repository
	request := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/applications", nil)
	setControlPlaneReadDevAuthHeaders(request)
	recorder := httptest.NewRecorder()

	server.httpServer.Handler.ServeHTTP(recorder, request)

	envelope := decodeControlPlaneReadEnvelope(t, recorder, http.StatusOK)
	if envelope.FailureCode != nil || repository.applicationCalls != 1 ||
		repository.lastContext.WorkspaceID != "workspace_demo" {
		t.Fatalf("dev membership was not projected to repository: envelope=%#v context=%#v", envelope, repository.lastContext)
	}
}

func TestWorkspaceQuotaPolicyUnavailableDoesNotQueryRepository(t *testing.T) {
	server := NewServer(config.Config{}, Options{BuildVersion: "test"})
	repository := &recordingControlPlaneReadRepository{}
	server.workspaceControlPlaneReadRepo = repository
	request := newControlPlaneReadRequest(
		http.MethodGet,
		"/v1/user-workspace/usage/quota-summary",
		controlPlaneReadTestAuth("tenant_demo", "usage:read"),
	)
	recorder := httptest.NewRecorder()

	server.httpServer.Handler.ServeHTTP(recorder, request)

	envelope := decodeControlPlaneReadEnvelope(t, recorder, http.StatusServiceUnavailable)
	assertControlPlaneReadFailure(t, envelope, "quota_policy_unavailable")
	if repository.totalCalls != 0 {
		t.Fatalf("quota without policy owner reached repository %d times", repository.totalCalls)
	}
}
