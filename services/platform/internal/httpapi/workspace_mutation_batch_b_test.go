package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
)

func TestBatchBMutationAuthorizationDenialsDoNotReachOwners(t *testing.T) {
	now := time.Now().UTC()
	operations := batchBMutationAuthorizationOperations(t)
	tests := []struct {
		name          string
		bodyMismatch  bool
		malformedBody bool
		active        string
		mutate        func(*controlPlaneReadAuthContext)
		expectedCode  int
		expectedError string
	}{
		{
			name: "identity permission denied", active: "workspace_demo",
			mutate:       func(auth *controlPlaneReadAuthContext) { auth.ScopeGrants = []string{"applications:read"} },
			expectedCode: http.StatusForbidden, expectedError: "scope_denied",
		},
		{
			name: "selection missing before malformed payload", malformedBody: true,
			expectedCode: http.StatusBadRequest, expectedError: "workspace_selection_missing",
		},
		{
			name: "membership missing", active: "workspace_demo",
			mutate:       func(auth *controlPlaneReadAuthContext) { auth.WorkspaceMemberships = nil },
			expectedCode: http.StatusForbidden, expectedError: "workspace_membership_denied",
		},
		{
			name: "membership expired", active: "workspace_demo",
			mutate: func(auth *controlPlaneReadAuthContext) {
				auth.WorkspaceMemberships[0].ExpiresAt = now.Add(-time.Second)
			},
			expectedCode: http.StatusForbidden, expectedError: "workspace_membership_expired",
		},
		{
			name: "membership permission denied", active: "workspace_demo",
			mutate: func(auth *controlPlaneReadAuthContext) {
				auth.WorkspaceMemberships[0].PermissionGrants = []string{"applications:read"}
			},
			expectedCode: http.StatusForbidden, expectedError: "workspace_permission_denied",
		},
		{
			name: "payload workspace mismatch", active: "workspace_demo", bodyMismatch: true,
			expectedCode: http.StatusForbidden, expectedError: "workspace_binding_mismatch",
		},
		{
			name: "OIDC membership unavailable", active: "workspace_demo",
			mutate: func(auth *controlPlaneReadAuthContext) {
				auth.AuthMode = controlPlaneReadAuthModeRadishOIDCIntegrationTest
			},
			expectedCode: http.StatusServiceUnavailable, expectedError: "workspace_membership_unavailable",
		},
	}

	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			for _, test := range tests {
				t.Run(test.name, func(t *testing.T) {
					server, ownerCalls := operation.server()
					body := operation.body(test.bodyMismatch)
					if test.malformedBody {
						body = `{`
					}
					request := httptest.NewRequest(http.MethodPost, operation.target, strings.NewReader(body))
					if test.active != "" {
						request.Header.Set(activeWorkspaceHeader, test.active)
					}
					auth := batchBMutationAuth(now, operation.permission)
					if test.mutate != nil {
						test.mutate(&auth)
					}
					request = request.WithContext(withControlPlaneReadFakeAuthContext(request.Context(), auth))
					recorder := httptest.NewRecorder()

					operation.handle(server, recorder, request)

					if recorder.Code != test.expectedCode || !strings.Contains(recorder.Body.String(), test.expectedError) {
						t.Fatalf("unexpected denial: status=%d body=%s", recorder.Code, recorder.Body.String())
					}
					if ownerCalls.Load() != 0 {
						t.Fatalf("authorization denial reached %s owner %d times", operation.name, ownerCalls.Load())
					}
				})
			}
		})
	}
}

func TestBatchBCompositeBindingPermissionsAreAtomic(t *testing.T) {
	tests := []struct {
		name                  string
		identityPermissions   []string
		membershipPermissions []string
		expectedError         string
		expectedProviderCalls int64
	}{
		{
			name:                  "identity second permission missing",
			identityPermissions:   []string{"application_drafts:write"},
			membershipPermissions: []string{"application_drafts:write", "prompt_application_templates:bind"},
			expectedError:         "scope_denied", expectedProviderCalls: 0,
		},
		{
			name:                  "membership second permission missing",
			identityPermissions:   []string{"application_drafts:write", "prompt_application_templates:bind"},
			membershipPermissions: []string{"application_drafts:write"},
			expectedError:         "workspace_permission_denied", expectedProviderCalls: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			owner := &countingApplicationConfigurationDraftRepository{}
			provider := &countingWorkspaceMembershipProvider{delegate: newDeterministicDevTestWorkspaceMembershipProvider()}
			server := &Server{
				config:                     config.Config{ApplicationDraftDevHTTPEnabled: true, ApplicationDraftDevWriteEnabled: true},
				applicationDraftRepository: owner, workspaceMembershipProvider: provider,
			}
			body := applicationConfigurationDraftPromptTemplateBindingBody{
				WorkspaceID: "workspace_demo", ApplicationID: "app_flow_copilot", ExpectedDraftVersion: 1,
				TemplateID: "ptpl_aaaaaaaaaaaaaaaa", TemplateVersion: 1,
			}
			raw, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			request := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/application-configuration-drafts/draft_demo/prompt-template-binding", strings.NewReader(string(raw)))
			request.SetPathValue("draft_id", "draft_demo")
			request.Header.Set(activeWorkspaceHeader, "workspace_demo")
			auth := batchBMutationAuth(time.Now().UTC(), test.identityPermissions...)
			auth.WorkspaceMemberships[0].PermissionGrants = append([]string{}, test.membershipPermissions...)
			request = request.WithContext(withControlPlaneReadFakeAuthContext(request.Context(), auth))
			recorder := httptest.NewRecorder()

			server.handleBindApplicationConfigurationDraftPromptTemplate(recorder, request)

			if recorder.Code != http.StatusForbidden || !strings.Contains(recorder.Body.String(), test.expectedError) {
				t.Fatalf("unexpected composite denial: status=%d body=%s", recorder.Code, recorder.Body.String())
			}
			if provider.calls.Load() != test.expectedProviderCalls || owner.calls.Load() != 0 {
				t.Fatalf("composite authorization was not atomic: provider=%d owner=%d", provider.calls.Load(), owner.calls.Load())
			}
		})
	}
}

func TestBatchBMutationSignedIdentityAndMembershipReachValidationOwners(t *testing.T) {
	privateKey := generateSignedTestPrivateKey(t)
	server := newSignedTestControlPlaneReadServer(t, privateKey)
	server.config.WorkflowSavedDraftDevHTTPEnabled = true
	server.config.ApplicationDraftDevHTTPEnabled = true
	server.config.PromptTemplateDevHTTPEnabled = true
	server.config.AgentCopilotProfileDevHTTPEnabled = true

	workflowPayload := validSavedWorkflowDraftPayload()
	applicationPayload := validApplicationDraftPayload()
	promptPayload := validPromptApplicationTemplateDraftInput()
	agentPayload := agentCopilotProfileTestInput("acpf_aaaaaaaaaaaaaaaa")
	agentPayload.WorkspaceID = "workspace_demo"
	agentPayload.ApplicationID = "app_aaaaaaaaaaaaaaaa"
	tests := []struct {
		name               string
		target             string
		upstreamPermission string
		permission         string
		body               any
	}{
		{
			name: "workflow draft", target: "/v1/user-workspace/workflow-drafts/validate",
			upstreamPermission: "radishmind.workflow-drafts.write", permission: "workflow_drafts:write",
			body: savedWorkflowDraftValidateHTTPBody{Draft: savedWorkflowDraftPayloadDocumentFromDraftPayload(workflowPayload)},
		},
		{
			name: "application configuration draft", target: "/v1/user-workspace/application-drafts/validate",
			upstreamPermission: "radishmind.application-drafts.write", permission: "application_drafts:write",
			body: applicationConfigurationDraftValidateBody{Draft: applicationPayload},
		},
		{
			name: "prompt template", target: "/v1/user-workspace/prompt-application-templates/validate",
			upstreamPermission: "radishmind.prompt-application-templates.write", permission: "prompt_application_templates:write",
			body: promptApplicationTemplateValidateBody{Template: promptPayload},
		},
		{
			name: "agent profile", target: "/v1/user-workspace/agent-copilot-profiles/validate",
			upstreamPermission: "radishmind.agent-copilot-profiles.write", permission: "agent_copilot_profiles:write",
			body: agentCopilotProfileValidateBody{Profile: agentPayload},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw, err := json.Marshal(test.body)
			if err != nil {
				t.Fatal(err)
			}
			claims := validSignedTestClaims()
			claims["permissions"] = []string{test.upstreamPermission}
			claims["workspace_memberships"] = []map[string]any{{
				"workspace_id": "workspace_demo",
				"permissions":  []string{test.permission},
			}}
			request := httptest.NewRequest(http.MethodPost, test.target, strings.NewReader(string(raw)))
			request.Header.Set(activeWorkspaceHeader, "workspace_demo")
			request.Header.Set("Authorization", "Bearer "+signControlPlaneReadTestToken(t, privateKey, "RS256", claims))
			recorder := httptest.NewRecorder()

			server.httpServer.Handler.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK ||
				strings.Contains(recorder.Body.String(), "scope_denied") ||
				strings.Contains(recorder.Body.String(), "workspace_membership_") {
				t.Fatalf("signed Batch B authorization did not reach validation owner: status=%d body=%s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

type batchBMutationAuthorizationOperation struct {
	name       string
	target     string
	permission string
	body       func(bool) string
	server     func() (*Server, *atomic.Int64)
	handle     func(*Server, http.ResponseWriter, *http.Request)
}

func batchBMutationAuthorizationOperations(t *testing.T) []batchBMutationAuthorizationOperation {
	t.Helper()
	workflowBody := func(mismatch bool) string {
		payload := validSavedWorkflowDraftPayload()
		if mismatch {
			payload.WorkspaceID = "workspace_other"
		}
		raw, err := json.Marshal(savedWorkflowDraftSaveHTTPBody{
			Draft: savedWorkflowDraftPayloadDocumentFromDraftPayload(payload),
		})
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	applicationBody := func(mismatch bool) string {
		payload := validApplicationDraftPayload()
		if mismatch {
			payload.WorkspaceID = "workspace_other"
		}
		raw, err := json.Marshal(applicationConfigurationDraftSaveBody{Draft: payload})
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	promptBody := func(mismatch bool) string {
		payload := validPromptApplicationTemplateDraftInput()
		if mismatch {
			payload.WorkspaceID = "workspace_other"
		}
		raw, err := json.Marshal(promptApplicationTemplateSaveBody{Template: payload})
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	agentBody := func(mismatch bool) string {
		payload := agentCopilotProfileTestInput("acpf_aaaaaaaaaaaaaaaa")
		payload.WorkspaceID = "workspace_demo"
		payload.ApplicationID = "app_aaaaaaaaaaaaaaaa"
		if mismatch {
			payload.WorkspaceID = "workspace_other"
		}
		raw, err := json.Marshal(agentCopilotProfileSaveBody{Profile: payload})
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	return []batchBMutationAuthorizationOperation{
		{
			name: "workflow draft", target: "/v1/user-workspace/workflow-drafts", permission: "workflow_drafts:write", body: workflowBody,
			server: func() (*Server, *atomic.Int64) {
				owner := &countingSavedWorkflowDraftStore{}
				return &Server{
					config:                  config.Config{WorkflowSavedDraftDevHTTPEnabled: true, WorkflowSavedDraftDevWriteEnabled: true},
					savedWorkflowDraftStore: owner, workspaceMembershipProvider: newDeterministicDevTestWorkspaceMembershipProvider(),
				}, &owner.calls
			},
			handle: (*Server).handleSaveWorkflowDraft,
		},
		{
			name: "application configuration draft", target: "/v1/user-workspace/application-drafts", permission: "application_drafts:write", body: applicationBody,
			server: func() (*Server, *atomic.Int64) {
				owner := &countingApplicationConfigurationDraftRepository{}
				return &Server{
					config:                     config.Config{ApplicationDraftDevHTTPEnabled: true, ApplicationDraftDevWriteEnabled: true},
					applicationDraftRepository: owner, workspaceMembershipProvider: newDeterministicDevTestWorkspaceMembershipProvider(),
				}, &owner.calls
			},
			handle: (*Server).handleSaveApplicationConfigurationDraft,
		},
		{
			name: "prompt template", target: "/v1/user-workspace/prompt-application-templates", permission: "prompt_application_templates:write", body: promptBody,
			server: func() (*Server, *atomic.Int64) {
				owner := &countingPromptApplicationTemplateRepository{}
				return &Server{
					config:                              config.Config{PromptTemplateDevHTTPEnabled: true, PromptTemplateDevWriteEnabled: true},
					promptApplicationTemplateRepository: owner, workspaceMembershipProvider: newDeterministicDevTestWorkspaceMembershipProvider(),
				}, &owner.calls
			},
			handle: (*Server).handleSavePromptApplicationTemplate,
		},
		{
			name: "agent profile", target: "/v1/user-workspace/agent-copilot-profiles", permission: "agent_copilot_profiles:write", body: agentBody,
			server: func() (*Server, *atomic.Int64) {
				owner := &countingAgentCopilotProfileRepository{}
				return &Server{
					config:                        config.Config{AgentCopilotProfileDevHTTPEnabled: true, AgentCopilotProfileDevWriteEnabled: true},
					agentCopilotProfileRepository: owner, workspaceMembershipProvider: newDeterministicDevTestWorkspaceMembershipProvider(),
				}, &owner.calls
			},
			handle: (*Server).handleSaveAgentCopilotProfile,
		},
	}
}

func batchBMutationAuth(now time.Time, permissions ...string) controlPlaneReadAuthContext {
	return controlPlaneReadAuthContext{
		AuthMode: controlPlaneReadAuthModeDevHeaders, IdentityContext: "verified:batch-b",
		TenantBinding: "tenant_demo", SubjectBinding: "subject_owner",
		ScopeGrants:     append([]string{}, permissions...),
		ResourceBinding: ControlPlaneResourceBinding{TenantRef: "tenant_demo", TenantVerified: true},
		WorkspaceMemberships: []VerifiedWorkspaceMembershipAssertion{{
			TenantRef: "tenant_demo", SubjectRef: "subject_owner", WorkspaceID: "workspace_demo",
			PermissionGrants: append([]string{}, permissions...), SourceRef: "membership:test",
			PolicyVersion: workspaceMembershipPolicyVersion, ExpiresAt: now.Add(time.Hour),
		}},
	}
}

type countingWorkspaceMembershipProvider struct {
	delegate WorkspaceMembershipProvider
	calls    atomic.Int64
}

func (provider *countingWorkspaceMembershipProvider) AuthorizeWorkspace(ctx context.Context, request WorkspaceMembershipRequest) WorkspaceMembershipDecision {
	provider.calls.Add(1)
	return provider.delegate.AuthorizeWorkspace(ctx, request)
}

type countingSavedWorkflowDraftStore struct{ calls atomic.Int64 }

func (store *countingSavedWorkflowDraftStore) ReadDraftByID(SavedWorkflowDraftContext, string) (SavedWorkflowDraft, bool, error) {
	store.calls.Add(1)
	return SavedWorkflowDraft{}, false, errors.New("unexpected owner call")
}
func (store *countingSavedWorkflowDraftStore) ListDraftSummariesByScope(SavedWorkflowDraftContext) ([]SavedWorkflowDraftSummary, error) {
	store.calls.Add(1)
	return nil, errors.New("unexpected owner call")
}
func (store *countingSavedWorkflowDraftStore) WriteDraft(SavedWorkflowDraftContext, SavedWorkflowDraft, int) (int, error) {
	store.calls.Add(1)
	return 0, errors.New("unexpected owner call")
}
func (store *countingSavedWorkflowDraftStore) SideEffects() SavedWorkflowDraftSideEffects {
	return SavedWorkflowDraftSideEffects{}
}

type countingApplicationConfigurationDraftRepository struct{ calls atomic.Int64 }

func (repository *countingApplicationConfigurationDraftRepository) Save(ApplicationConfigurationDraftContext, ApplicationConfigurationDraft, int) (ApplicationConfigurationDraft, error) {
	repository.calls.Add(1)
	return ApplicationConfigurationDraft{}, errors.New("unexpected owner call")
}
func (repository *countingApplicationConfigurationDraftRepository) Read(ApplicationConfigurationDraftContext, string) (ApplicationConfigurationDraft, error) {
	repository.calls.Add(1)
	return ApplicationConfigurationDraft{}, errors.New("unexpected owner call")
}
func (repository *countingApplicationConfigurationDraftRepository) List(ApplicationConfigurationDraftContext) ([]ApplicationConfigurationDraftSummary, error) {
	repository.calls.Add(1)
	return nil, errors.New("unexpected owner call")
}

type countingPromptApplicationTemplateRepository struct{ calls atomic.Int64 }

func (repository *countingPromptApplicationTemplateRepository) SaveDraft(PromptApplicationTemplateContext, PromptApplicationTemplateDraft, int) (PromptApplicationTemplateDraft, error) {
	repository.calls.Add(1)
	return PromptApplicationTemplateDraft{}, errors.New("unexpected owner call")
}
func (repository *countingPromptApplicationTemplateRepository) ReadDraft(PromptApplicationTemplateContext, string) (PromptApplicationTemplateDraft, error) {
	repository.calls.Add(1)
	return PromptApplicationTemplateDraft{}, errors.New("unexpected owner call")
}
func (repository *countingPromptApplicationTemplateRepository) ListDrafts(PromptApplicationTemplateContext) ([]PromptApplicationTemplateDraftSummary, error) {
	repository.calls.Add(1)
	return nil, errors.New("unexpected owner call")
}
func (repository *countingPromptApplicationTemplateRepository) CreateVersion(PromptApplicationTemplateContext, PromptApplicationTemplateVersion) (PromptApplicationTemplateVersion, error) {
	repository.calls.Add(1)
	return PromptApplicationTemplateVersion{}, errors.New("unexpected owner call")
}
func (repository *countingPromptApplicationTemplateRepository) ReadVersion(PromptApplicationTemplateContext, string, int) (PromptApplicationTemplateVersion, error) {
	repository.calls.Add(1)
	return PromptApplicationTemplateVersion{}, errors.New("unexpected owner call")
}
func (repository *countingPromptApplicationTemplateRepository) ListVersions(PromptApplicationTemplateContext, string) ([]PromptApplicationTemplateVersionSummary, error) {
	repository.calls.Add(1)
	return nil, errors.New("unexpected owner call")
}

type countingAgentCopilotProfileRepository struct{ calls atomic.Int64 }

func (repository *countingAgentCopilotProfileRepository) SaveDraft(AgentCopilotProfileContext, AgentCopilotProfileDraftV1, int) (AgentCopilotProfileDraftV1, error) {
	repository.calls.Add(1)
	return AgentCopilotProfileDraftV1{}, errors.New("unexpected owner call")
}
func (repository *countingAgentCopilotProfileRepository) ReadDraft(AgentCopilotProfileContext, string) (AgentCopilotProfileDraftV1, error) {
	repository.calls.Add(1)
	return AgentCopilotProfileDraftV1{}, errors.New("unexpected owner call")
}
func (repository *countingAgentCopilotProfileRepository) ListDrafts(AgentCopilotProfileContext) ([]AgentCopilotProfileDraftSummary, error) {
	repository.calls.Add(1)
	return nil, errors.New("unexpected owner call")
}
func (repository *countingAgentCopilotProfileRepository) CreateVersion(AgentCopilotProfileContext, AgentCopilotProfileVersionV1) (AgentCopilotProfileVersionV1, error) {
	repository.calls.Add(1)
	return AgentCopilotProfileVersionV1{}, errors.New("unexpected owner call")
}
func (repository *countingAgentCopilotProfileRepository) ReadVersion(AgentCopilotProfileContext, string, int) (AgentCopilotProfileVersionV1, error) {
	repository.calls.Add(1)
	return AgentCopilotProfileVersionV1{}, errors.New("unexpected owner call")
}
func (repository *countingAgentCopilotProfileRepository) ListVersions(AgentCopilotProfileContext, string) ([]AgentCopilotProfileVersionSummary, error) {
	repository.calls.Add(1)
	return nil, errors.New("unexpected owner call")
}
