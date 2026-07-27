package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"
)

const (
	activeWorkspaceHeader                   = "X-RadishMind-Active-Workspace"
	controlPlaneReadDevMembershipHeader     = "X-RadishMind-Dev-Read-Membership-Workspace"
	controlPlaneReadDevMembershipPermHeader = "X-RadishMind-Dev-Read-Membership-Permissions"
	workspaceMembershipPolicyVersion        = "workspace_membership_dev_test_v1"
)

var workspacePermissionAllowlist = map[string]struct{}{
	"applications:read":    {},
	"applications:write":   {},
	"applications:archive": {},
	"api_keys:read":        {},
	"api_keys:write":       {},
	"api_keys:revoke":      {},
	"usage:read":           {},
	"runs:read":            {},
}

type VerifiedWorkspaceMembershipAssertion struct {
	TenantRef        string
	SubjectRef       string
	WorkspaceID      string
	PermissionGrants []string
	SourceRef        string
	PolicyVersion    string
	ExpiresAt        time.Time
}

type WorkspaceMembershipRequest struct {
	Auth               controlPlaneReadAuthContext
	ActiveWorkspaceID  string
	RequiredPermission string
}

type WorkspaceMembershipDecision struct {
	Binding     ControlPlaneResourceBinding
	FailureCode string
	HTTPStatus  int
}

type WorkspaceMembershipProvider interface {
	AuthorizeWorkspace(context.Context, WorkspaceMembershipRequest) WorkspaceMembershipDecision
}

type deterministicDevTestWorkspaceMembershipProvider struct {
	now func() time.Time
}

func newDeterministicDevTestWorkspaceMembershipProvider() WorkspaceMembershipProvider {
	return deterministicDevTestWorkspaceMembershipProvider{now: time.Now}
}

func (provider deterministicDevTestWorkspaceMembershipProvider) AuthorizeWorkspace(
	_ context.Context,
	request WorkspaceMembershipRequest,
) WorkspaceMembershipDecision {
	auth := request.Auth
	if auth.AuthMode == controlPlaneReadAuthModeRadishOIDCIntegrationTest {
		return workspaceMembershipFailure("workspace_membership_unavailable", http.StatusServiceUnavailable)
	}
	if auth.AuthMode != controlPlaneReadAuthModeDevHeaders && auth.AuthMode != controlPlaneReadAuthModeSignedTestToken {
		return workspaceMembershipFailure("workspace_membership_unavailable", http.StatusServiceUnavailable)
	}
	workspaceID := strings.TrimSpace(request.ActiveWorkspaceID)
	if workspaceID == "" {
		return workspaceMembershipFailure("workspace_selection_missing", http.StatusBadRequest)
	}
	if !validControlPlaneReadAuthReference(workspaceID, false) {
		return workspaceMembershipFailure("workspace_binding_mismatch", http.StatusForbidden)
	}
	permission := strings.TrimSpace(request.RequiredPermission)
	if _, allowed := workspacePermissionAllowlist[permission]; !allowed {
		return workspaceMembershipFailure("workspace_permission_denied", http.StatusForbidden)
	}
	now := provider.now().UTC()
	for _, assertion := range auth.WorkspaceMemberships {
		if strings.TrimSpace(assertion.WorkspaceID) != workspaceID {
			continue
		}
		if strings.TrimSpace(assertion.TenantRef) != strings.TrimSpace(auth.TenantBinding) ||
			strings.TrimSpace(assertion.SubjectRef) != strings.TrimSpace(auth.SubjectBinding) {
			return workspaceMembershipFailure("workspace_binding_mismatch", http.StatusForbidden)
		}
		if !assertion.ExpiresAt.IsZero() && !assertion.ExpiresAt.After(now) {
			return workspaceMembershipFailure("workspace_membership_expired", http.StatusForbidden)
		}
		if !controlPlaneReadHasScope(assertion.PermissionGrants, permission) {
			return workspaceMembershipFailure("workspace_permission_denied", http.StatusForbidden)
		}
		binding := auth.ResourceBinding
		binding.WorkspaceID = workspaceID
		binding.WorkspaceMembershipVerified = true
		binding.WorkspacePermissionGrants = append([]string{}, assertion.PermissionGrants...)
		binding.WorkspaceSourceRef = strings.TrimSpace(assertion.SourceRef)
		binding.WorkspacePolicyVersion = strings.TrimSpace(assertion.PolicyVersion)
		binding.WorkspaceExpiresAt = assertion.ExpiresAt
		return WorkspaceMembershipDecision{Binding: binding, HTTPStatus: http.StatusOK}
	}
	if len(auth.WorkspaceMemberships) > 0 {
		return workspaceMembershipFailure("workspace_binding_mismatch", http.StatusForbidden)
	}
	return workspaceMembershipFailure("workspace_membership_denied", http.StatusForbidden)
}

func workspaceMembershipFailure(code string, status int) WorkspaceMembershipDecision {
	return WorkspaceMembershipDecision{FailureCode: code, HTTPStatus: status}
}

func validWorkspacePermissionGrants(grants []string) bool {
	if len(grants) == 0 {
		return false
	}
	seen := make(map[string]struct{}, len(grants))
	for _, raw := range grants {
		grant := strings.TrimSpace(raw)
		if _, allowed := workspacePermissionAllowlist[grant]; !allowed {
			return false
		}
		if _, duplicate := seen[grant]; duplicate {
			return false
		}
		seen[grant] = struct{}{}
	}
	return true
}

func workspaceMembershipsFromDevHeaders(
	request *http.Request,
	tenantRef string,
	subjectRef string,
) []VerifiedWorkspaceMembershipAssertion {
	workspaceID := strings.TrimSpace(request.Header.Get(controlPlaneReadDevMembershipHeader))
	permissions := splitControlPlaneReadDevScopes(request.Header.Get(controlPlaneReadDevMembershipPermHeader))
	if workspaceID == "" || !validControlPlaneReadAuthReference(workspaceID, false) || !validWorkspacePermissionGrants(permissions) {
		return nil
	}
	return []VerifiedWorkspaceMembershipAssertion{{
		TenantRef: tenantRef, SubjectRef: subjectRef, WorkspaceID: workspaceID,
		PermissionGrants: append([]string{}, permissions...),
		SourceRef:        "membership:dev-headers", PolicyVersion: workspaceMembershipPolicyVersion,
	}}
}
