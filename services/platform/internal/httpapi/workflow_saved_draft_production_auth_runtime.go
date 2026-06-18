package httpapi

import "strings"

const savedWorkflowDraftProductionAuthSourceRadishOIDC = "radish_oidc_verified_context"

type SavedWorkflowDraftVerifiedAuthContext struct {
	RequestID       string
	AuthSource      string
	TenantRef       string
	ActorSubjectRef string
	ScopeGrants     []string
	AuditRef        string
}

type SavedWorkflowDraftWorkspaceBinding struct {
	TenantRef                   string
	WorkspaceID                 string
	ApplicationID               string
	OwnerSubjectRef             string
	WorkspaceMembershipVerified bool
	ApplicationScopeVerified    bool
	OwnerScopeVerified          bool
}

type SavedWorkflowDraftProductionAuthRuntimeResult struct {
	ActorContext SavedWorkflowDraftRepositoryActorContext
	FailureCode  SavedWorkflowDraftFailureCode
}

func BuildSavedWorkflowDraftRepositoryActorContextForOperation(
	auth SavedWorkflowDraftVerifiedAuthContext,
	binding SavedWorkflowDraftWorkspaceBinding,
	operation string,
) SavedWorkflowDraftProductionAuthRuntimeResult {
	requiredScope, ok := savedWorkflowDraftRepositoryRequiredScopeForOperation(operation)
	if !ok {
		return SavedWorkflowDraftProductionAuthRuntimeResult{
			FailureCode: SavedWorkflowDraftFailureAuthContextMismatch,
		}
	}
	return BuildSavedWorkflowDraftRepositoryActorContextFromProductionAuth(
		auth,
		binding,
		requiredScope,
	)
}

func BuildSavedWorkflowDraftRepositoryActorContextFromProductionAuth(
	auth SavedWorkflowDraftVerifiedAuthContext,
	binding SavedWorkflowDraftWorkspaceBinding,
	requiredScope string,
) SavedWorkflowDraftProductionAuthRuntimeResult {
	auth = normalizeSavedWorkflowDraftVerifiedAuthContext(auth)
	binding = normalizeSavedWorkflowDraftWorkspaceBinding(binding)
	requiredScope = strings.TrimSpace(requiredScope)

	actor := SavedWorkflowDraftRepositoryActorContext{
		RequestID:       auth.RequestID,
		TenantRef:       auth.TenantRef,
		WorkspaceID:     binding.WorkspaceID,
		ApplicationID:   binding.ApplicationID,
		ActorSubjectRef: auth.ActorSubjectRef,
		OwnerSubjectRef: binding.OwnerSubjectRef,
		ScopeGrants:     append([]string{}, auth.ScopeGrants...),
		AuditRef:        auth.AuditRef,
	}
	if failureCode := savedWorkflowDraftProductionAuthRuntimeFailure(auth, binding, requiredScope); failureCode != "" {
		return SavedWorkflowDraftProductionAuthRuntimeResult{
			ActorContext: actor,
			FailureCode:  failureCode,
		}
	}
	return SavedWorkflowDraftProductionAuthRuntimeResult{ActorContext: actor}
}

func normalizeSavedWorkflowDraftVerifiedAuthContext(
	auth SavedWorkflowDraftVerifiedAuthContext,
) SavedWorkflowDraftVerifiedAuthContext {
	auth.RequestID = strings.TrimSpace(auth.RequestID)
	auth.AuthSource = strings.TrimSpace(auth.AuthSource)
	auth.TenantRef = strings.TrimSpace(auth.TenantRef)
	auth.ActorSubjectRef = strings.TrimSpace(auth.ActorSubjectRef)
	auth.ScopeGrants = normalizedStringSet(auth.ScopeGrants)
	auth.AuditRef = strings.TrimSpace(auth.AuditRef)
	return auth
}

func normalizeSavedWorkflowDraftWorkspaceBinding(
	binding SavedWorkflowDraftWorkspaceBinding,
) SavedWorkflowDraftWorkspaceBinding {
	binding.TenantRef = strings.TrimSpace(binding.TenantRef)
	binding.WorkspaceID = strings.TrimSpace(binding.WorkspaceID)
	binding.ApplicationID = strings.TrimSpace(binding.ApplicationID)
	binding.OwnerSubjectRef = strings.TrimSpace(binding.OwnerSubjectRef)
	return binding
}

func savedWorkflowDraftProductionAuthRuntimeFailure(
	auth SavedWorkflowDraftVerifiedAuthContext,
	binding SavedWorkflowDraftWorkspaceBinding,
	requiredScope string,
) SavedWorkflowDraftFailureCode {
	if auth.AuthSource != savedWorkflowDraftProductionAuthSourceRadishOIDC {
		return SavedWorkflowDraftFailureAuthContextMismatch
	}
	if auth.RequestID == "" {
		return SavedWorkflowDraftFailureAuthContextMismatch
	}
	if auth.ActorSubjectRef == "" {
		return SavedWorkflowDraftFailureIdentityContextMissing
	}
	if auth.AuditRef == "" {
		return SavedWorkflowDraftFailureAuditContextMissing
	}
	if auth.TenantRef == "" || binding.TenantRef == "" || auth.TenantRef != binding.TenantRef {
		return SavedWorkflowDraftFailureTenantBindingMissing
	}
	if binding.WorkspaceID == "" || !binding.WorkspaceMembershipVerified {
		return SavedWorkflowDraftFailureWorkspaceMembershipDenied
	}
	if binding.ApplicationID == "" || !binding.ApplicationScopeVerified {
		return SavedWorkflowDraftFailureApplicationScopeDenied
	}
	if binding.OwnerSubjectRef == "" || !binding.OwnerScopeVerified {
		return SavedWorkflowDraftFailureOwnerScopeDenied
	}
	if requiredScope == "" {
		return SavedWorkflowDraftFailureAuthContextMismatch
	}
	if !controlPlaneReadHasScope(auth.ScopeGrants, requiredScope) {
		return SavedWorkflowDraftFailureScopeGrantMissing
	}
	return ""
}

func savedWorkflowDraftRepositoryRequiredScopeForOperation(operation string) (string, bool) {
	switch strings.TrimSpace(operation) {
	case "SaveWorkflowDraftRecord":
		return "workflow_drafts:write", true
	case "ReadWorkflowDraftRecord", "ListWorkflowDraftRecords":
		return "workflow_drafts:read", true
	default:
		return "", false
	}
}
