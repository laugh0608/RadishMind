package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"
)

const localWorkspaceMembershipPolicyVersion = "local_workspace_membership_v1"

type localWorkspaceMembershipProvider struct {
	repository localIdentityRepository
	now        func() time.Time
}

func newLocalWorkspaceMembershipProvider(repository localIdentityRepository) WorkspaceMembershipProvider {
	return &localWorkspaceMembershipProvider{repository: repository, now: time.Now}
}

func (provider *localWorkspaceMembershipProvider) AuthorizeWorkspace(
	ctx context.Context,
	request WorkspaceMembershipRequest,
) WorkspaceMembershipDecision {
	if provider == nil || provider.repository == nil {
		return workspaceMembershipFailure("workspace_membership_unavailable", http.StatusServiceUnavailable)
	}
	workspaceID := strings.TrimSpace(request.ActiveWorkspaceID)
	if workspaceID == "" {
		return workspaceMembershipFailure("workspace_selection_missing", http.StatusBadRequest)
	}
	userID, ok := localUserIDFromActorRef(request.Auth.SubjectBinding)
	if !ok || !validControlPlaneReadAuthReference(request.Auth.TenantBinding, false) ||
		!validControlPlaneReadAuthReference(workspaceID, false) || !request.Auth.ResourceBinding.TenantVerified ||
		strings.TrimSpace(request.Auth.ResourceBinding.TenantRef) != strings.TrimSpace(request.Auth.TenantBinding) {
		return workspaceMembershipFailure("workspace_binding_mismatch", http.StatusForbidden)
	}
	now := time.Now().UTC()
	if provider.now != nil {
		now = provider.now().UTC()
	}
	authorization, err := provider.repository.AuthorizeWorkspace(
		ctx,
		userID,
		strings.TrimSpace(request.Auth.TenantBinding),
		workspaceID,
		request.RequiredPermissions,
		now,
	)
	if err != nil {
		switch localIdentityRepositoryError(err) {
		case LocalIdentityFailureStoreUnavailable:
			return workspaceMembershipFailure("workspace_membership_unavailable", http.StatusServiceUnavailable)
		case LocalIdentityFailurePermissionDenied:
			return workspaceMembershipFailure("workspace_permission_denied", http.StatusForbidden)
		default:
			return workspaceMembershipFailure("workspace_membership_denied", http.StatusForbidden)
		}
	}
	binding := request.Auth.ResourceBinding
	binding.TenantRef = authorization.Membership.TenantRef
	binding.WorkspaceID = authorization.Membership.WorkspaceID
	binding.WorkspaceMembershipVerified = true
	binding.WorkspacePermissionGrants = append([]string(nil), authorization.PermissionGrants...)
	binding.WorkspaceSourceRef = "membership:local:" + authorization.Membership.MembershipID
	binding.WorkspacePolicyVersion = localWorkspaceMembershipPolicyVersion
	if authorization.Membership.ExpiresAt != nil {
		binding.WorkspaceExpiresAt = authorization.Membership.ExpiresAt.UTC()
	}
	return WorkspaceMembershipDecision{Binding: binding, HTTPStatus: http.StatusOK}
}

func localUserIDFromActorRef(actorRef string) (string, bool) {
	const prefix = "user:"
	actorRef = strings.TrimSpace(actorRef)
	if !strings.HasPrefix(actorRef, prefix) {
		return "", false
	}
	userID := strings.TrimPrefix(actorRef, prefix)
	return userID, localUserIDPattern.MatchString(userID)
}
