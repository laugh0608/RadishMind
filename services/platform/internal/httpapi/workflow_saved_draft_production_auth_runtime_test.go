package httpapi

import "testing"

func TestSavedWorkflowDraftProductionAuthRuntimeBuildsRepositoryActorContext(t *testing.T) {
	result := BuildSavedWorkflowDraftRepositoryActorContextForOperation(
		validSavedWorkflowDraftVerifiedAuthContext(),
		validSavedWorkflowDraftWorkspaceBinding(),
		"SaveWorkflowDraftRecord",
	)
	if result.FailureCode != "" {
		t.Fatalf("production auth runtime bridge should succeed: %#v", result)
	}
	actor := result.ActorContext
	if actor.RequestID != "req:saved-draft-production-auth-runtime" ||
		actor.TenantRef != "tenant:demo" ||
		actor.WorkspaceID != "workspace_demo" ||
		actor.ApplicationID != "app_flow_copilot" ||
		actor.ActorSubjectRef != "subject_platform_ops" ||
		actor.OwnerSubjectRef != "subject_platform_ops" ||
		actor.AuditRef != "audit:saved-draft-production-auth-runtime" {
		t.Fatalf("actor context projection drifted: %#v", actor)
	}
	if len(actor.ScopeGrants) != 2 ||
		actor.ScopeGrants[0] != "workflow_drafts:read" ||
		actor.ScopeGrants[1] != "workflow_drafts:write" {
		t.Fatalf("scope grants must be normalized and projected: %#v", actor.ScopeGrants)
	}
}

func TestSavedWorkflowDraftProductionAuthRuntimeOperationScopes(t *testing.T) {
	for _, tc := range []struct {
		operation string
		scopes    []string
	}{
		{operation: "SaveWorkflowDraftRecord", scopes: []string{"workflow_drafts:write"}},
		{operation: "ReadWorkflowDraftRecord", scopes: []string{"workflow_drafts:read"}},
		{operation: "ListWorkflowDraftRecords", scopes: []string{"workflow_drafts:read"}},
	} {
		t.Run(tc.operation, func(t *testing.T) {
			auth := validSavedWorkflowDraftVerifiedAuthContext()
			auth.ScopeGrants = tc.scopes
			result := BuildSavedWorkflowDraftRepositoryActorContextForOperation(
				auth,
				validSavedWorkflowDraftWorkspaceBinding(),
				tc.operation,
			)
			if result.FailureCode != "" {
				t.Fatalf("%s should accept its required scope: %#v", tc.operation, result)
			}
		})
	}
}

func TestSavedWorkflowDraftProductionAuthRuntimeFailureMapping(t *testing.T) {
	for _, tc := range []struct {
		name        string
		mutateAuth  func(SavedWorkflowDraftVerifiedAuthContext) SavedWorkflowDraftVerifiedAuthContext
		mutateScope func(SavedWorkflowDraftWorkspaceBinding) SavedWorkflowDraftWorkspaceBinding
		operation   string
		failureCode SavedWorkflowDraftFailureCode
	}{
		{
			name: "rejects dev fake auth source",
			mutateAuth: func(auth SavedWorkflowDraftVerifiedAuthContext) SavedWorkflowDraftVerifiedAuthContext {
				auth.AuthSource = "dev_fake_auth_header"
				return auth
			},
			failureCode: SavedWorkflowDraftFailureAuthContextMismatch,
		},
		{
			name: "rejects missing request id",
			mutateAuth: func(auth SavedWorkflowDraftVerifiedAuthContext) SavedWorkflowDraftVerifiedAuthContext {
				auth.RequestID = ""
				return auth
			},
			failureCode: SavedWorkflowDraftFailureAuthContextMismatch,
		},
		{
			name: "rejects missing actor subject",
			mutateAuth: func(auth SavedWorkflowDraftVerifiedAuthContext) SavedWorkflowDraftVerifiedAuthContext {
				auth.ActorSubjectRef = ""
				return auth
			},
			failureCode: SavedWorkflowDraftFailureIdentityContextMissing,
		},
		{
			name: "rejects missing audit",
			mutateAuth: func(auth SavedWorkflowDraftVerifiedAuthContext) SavedWorkflowDraftVerifiedAuthContext {
				auth.AuditRef = ""
				return auth
			},
			failureCode: SavedWorkflowDraftFailureAuditContextMissing,
		},
		{
			name: "rejects tenant mismatch",
			mutateScope: func(binding SavedWorkflowDraftWorkspaceBinding) SavedWorkflowDraftWorkspaceBinding {
				binding.TenantRef = "tenant:other"
				return binding
			},
			failureCode: SavedWorkflowDraftFailureTenantBindingMissing,
		},
		{
			name: "rejects missing workspace membership",
			mutateScope: func(binding SavedWorkflowDraftWorkspaceBinding) SavedWorkflowDraftWorkspaceBinding {
				binding.WorkspaceMembershipVerified = false
				return binding
			},
			failureCode: SavedWorkflowDraftFailureWorkspaceMembershipDenied,
		},
		{
			name: "rejects missing application scope",
			mutateScope: func(binding SavedWorkflowDraftWorkspaceBinding) SavedWorkflowDraftWorkspaceBinding {
				binding.ApplicationScopeVerified = false
				return binding
			},
			failureCode: SavedWorkflowDraftFailureApplicationScopeDenied,
		},
		{
			name: "rejects missing owner scope",
			mutateScope: func(binding SavedWorkflowDraftWorkspaceBinding) SavedWorkflowDraftWorkspaceBinding {
				binding.OwnerScopeVerified = false
				return binding
			},
			failureCode: SavedWorkflowDraftFailureOwnerScopeDenied,
		},
		{
			name: "rejects missing operation scope",
			mutateAuth: func(auth SavedWorkflowDraftVerifiedAuthContext) SavedWorkflowDraftVerifiedAuthContext {
				auth.ScopeGrants = []string{"workflow_drafts:read"}
				return auth
			},
			failureCode: SavedWorkflowDraftFailureScopeGrantMissing,
		},
		{
			name:        "rejects unknown repository operation",
			operation:   "DeleteWorkflowDraftRecord",
			failureCode: SavedWorkflowDraftFailureAuthContextMismatch,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			auth := validSavedWorkflowDraftVerifiedAuthContext()
			binding := validSavedWorkflowDraftWorkspaceBinding()
			operation := "SaveWorkflowDraftRecord"
			if tc.operation != "" {
				operation = tc.operation
			}
			if tc.mutateAuth != nil {
				auth = tc.mutateAuth(auth)
			}
			if tc.mutateScope != nil {
				binding = tc.mutateScope(binding)
			}
			result := BuildSavedWorkflowDraftRepositoryActorContextForOperation(auth, binding, operation)
			if result.FailureCode != tc.failureCode {
				t.Fatalf("unexpected failure mapping: got %s want %s result=%#v", result.FailureCode, tc.failureCode, result)
			}
		})
	}
}

func validSavedWorkflowDraftVerifiedAuthContext() SavedWorkflowDraftVerifiedAuthContext {
	return SavedWorkflowDraftVerifiedAuthContext{
		RequestID:       " req:saved-draft-production-auth-runtime ",
		AuthSource:      savedWorkflowDraftProductionAuthSourceRadishOIDC,
		TenantRef:       " tenant:demo ",
		ActorSubjectRef: " subject_platform_ops ",
		ScopeGrants: []string{
			" workflow_drafts:write ",
			"workflow_drafts:read",
			"workflow_drafts:write",
		},
		AuditRef: " audit:saved-draft-production-auth-runtime ",
	}
}

func validSavedWorkflowDraftWorkspaceBinding() SavedWorkflowDraftWorkspaceBinding {
	return SavedWorkflowDraftWorkspaceBinding{
		TenantRef:                   " tenant:demo ",
		WorkspaceID:                 " workspace_demo ",
		ApplicationID:               " app_flow_copilot ",
		OwnerSubjectRef:             " subject_platform_ops ",
		WorkspaceMembershipVerified: true,
		ApplicationScopeVerified:    true,
		OwnerScopeVerified:          true,
	}
}
