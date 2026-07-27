package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
)

func TestBatchDPermissionProjectionCoversEveryMutationGrant(t *testing.T) {
	grants := projectControlPlaneReadPermissions([]string{
		"radishmind.application-sessions.write",
		"radishmind.application-sessions.execute",
		"radishmind.workflow-runs.execute",
		"radishmind.workflow-drafts.read",
		"radishmind.workflow-definitions.read",
	})
	for _, expected := range []string{
		"application_sessions:write",
		"application_sessions:execute",
		"workflow_runs:execute",
		"workflow_drafts:read",
		"workflow_definitions:read",
	} {
		if !controlPlaneReadHasScope(grants, expected) {
			t.Fatalf("Batch D upstream permission did not project %q: %#v", expected, grants)
		}
	}
}

func TestBatchDMutationAuthorizationDenialsHaveZeroOwnerAndGatewayCalls(t *testing.T) {
	tests := []struct {
		name          string
		active        string
		malformedBody bool
		mismatch      bool
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
			name: "membership permission denied", active: "workspace_demo",
			mutate: func(auth *controlPlaneReadAuthContext) {
				auth.WorkspaceMemberships[0].PermissionGrants = []string{"applications:read"}
			},
			expectedCode: http.StatusForbidden, expectedError: "workspace_permission_denied",
		},
		{
			name: "resource workspace mismatch", active: "workspace_demo", mismatch: true,
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

	for _, operation := range batchDMutationAuthorizationOperations() {
		t.Run(operation.name, func(t *testing.T) {
			for _, test := range tests {
				t.Run(test.name, func(t *testing.T) {
					server, ownerCalls, gateway := operation.server()
					body := operation.body(test.mismatch)
					if test.malformedBody {
						body = `{`
					}
					request := httptest.NewRequest(http.MethodPost, strings.TrimPrefix(operation.route, "POST "), strings.NewReader(body))
					operation.prepare(request, test.mismatch)
					if test.active != "" {
						request.Header.Set(activeWorkspaceHeader, test.active)
					}
					auth := batchBMutationAuth(time.Now().UTC(), operation.permissions...)
					if test.mutate != nil {
						test.mutate(&auth)
					}
					request = request.WithContext(withControlPlaneReadFakeAuthContext(request.Context(), auth))
					recorder := httptest.NewRecorder()

					operation.handle(server, recorder, request)

					if recorder.Code != test.expectedCode || !strings.Contains(recorder.Body.String(), test.expectedError) {
						t.Fatalf("unexpected denial: status=%d body=%s", recorder.Code, recorder.Body.String())
					}
					if ownerCalls.Load() != 0 || gateway.callCount() != 0 {
						t.Fatalf("authorization denial reached side effects: owner=%d gateway=%d", ownerCalls.Load(), gateway.callCount())
					}
				})
			}
		})
	}
}

func TestBatchDRunPermissionsUseSingleMembershipDecision(t *testing.T) {
	for _, operation := range batchDMutationAuthorizationOperations()[3:] {
		t.Run(operation.name, func(t *testing.T) {
			for _, test := range []struct {
				name                  string
				identityPermissions   []string
				membershipPermissions []string
				expectedError         string
				expectedProviderCalls int64
			}{
				{
					name:                "identity second permission missing",
					identityPermissions: operation.permissions[:1], membershipPermissions: operation.permissions,
					expectedError: "scope_denied", expectedProviderCalls: 0,
				},
				{
					name:                "membership second permission missing",
					identityPermissions: operation.permissions, membershipPermissions: operation.permissions[:1],
					expectedError: "workspace_permission_denied", expectedProviderCalls: 1,
				},
			} {
				t.Run(test.name, func(t *testing.T) {
					server, ownerCalls, gateway := operation.server()
					provider := &countingWorkspaceMembershipProvider{delegate: newDeterministicDevTestWorkspaceMembershipProvider()}
					server.workspaceMembershipProvider = provider
					request := httptest.NewRequest(http.MethodPost, strings.TrimPrefix(operation.route, "POST "), strings.NewReader(operation.body(false)))
					request.Header.Set(activeWorkspaceHeader, "workspace_demo")
					operation.prepare(request, false)
					auth := batchBMutationAuth(time.Now().UTC(), test.identityPermissions...)
					auth.WorkspaceMemberships[0].PermissionGrants = append([]string{}, test.membershipPermissions...)
					request = request.WithContext(withControlPlaneReadFakeAuthContext(request.Context(), auth))
					recorder := httptest.NewRecorder()

					operation.handle(server, recorder, request)

					if recorder.Code != http.StatusForbidden || !strings.Contains(recorder.Body.String(), test.expectedError) {
						t.Fatalf("unexpected composite denial: status=%d body=%s", recorder.Code, recorder.Body.String())
					}
					if provider.calls.Load() != test.expectedProviderCalls || ownerCalls.Load() != 0 || gateway.callCount() != 0 {
						t.Fatalf("run authorization was not atomic: provider=%d owner=%d gateway=%d", provider.calls.Load(), ownerCalls.Load(), gateway.callCount())
					}
				})
			}
		})
	}
}

type batchDMutationAuthorizationOperation struct {
	name        string
	route       string
	permissions []string
	body        func(bool) string
	prepare     func(*http.Request, bool)
	server      func() (*Server, *atomic.Int64, *workflowExecutorTestBridge)
	handle      func(*Server, http.ResponseWriter, *http.Request)
}

func batchDMutationAuthorizationOperations() []batchDMutationAuthorizationOperation {
	sessionBody := func(mismatch bool) string {
		return `{"workspace_id":"` + batchDWorkspace(mismatch) + `","application_id":"app_flow_copilot","execution_profile":"workflow_definition_executor_v1","definition_id":"definition_demo"}`
	}
	sessionServer := func() (*Server, *atomic.Int64, *workflowExecutorTestBridge) {
		owner := &countingBatchDSessionRepository{}
		gateway := &workflowExecutorTestBridge{}
		return &Server{
			config:                                  config.Config{ApplicationSessionDevEnabled: true},
			applicationInteractionSessionRepository: owner,
			workspaceMembershipProvider:             newDeterministicDevTestWorkspaceMembershipProvider(),
			bridge:                                  gateway,
		}, &owner.calls, gateway
	}
	runHeaders := func(request *http.Request, mismatch bool) {
		request.Header.Set(savedWorkflowDraftDevWorkspaceHeader, batchDWorkspace(mismatch))
		request.Header.Set(savedWorkflowDraftDevApplicationHeader, "app_flow_copilot")
	}
	return []batchDMutationAuthorizationOperation{
		{
			name: "application session create", route: applicationSessionCreateRoute,
			permissions: []string{"application_sessions:write"}, body: sessionBody,
			prepare: func(*http.Request, bool) {}, server: sessionServer,
			handle: func(server *Server, writer http.ResponseWriter, request *http.Request) {
				server.handleCreateApplicationInteractionSession(writer, request)
			},
		},
		{
			name: "application session close", route: applicationSessionCloseRoute,
			permissions: []string{"application_sessions:write"},
			body: func(mismatch bool) string {
				return `{"workspace_id":"` + batchDWorkspace(mismatch) + `","application_id":"app_flow_copilot","expected_version":1}`
			},
			prepare: func(request *http.Request, _ bool) { request.SetPathValue("session_id", "appsess_aaaaaaaaaaaaaaaa") },
			server:  sessionServer,
			handle: func(server *Server, writer http.ResponseWriter, request *http.Request) {
				server.handleCloseApplicationInteractionSession(writer, request)
			},
		},
		{
			name: "application session turn", route: applicationSessionTurnRoute,
			permissions: []string{"application_sessions:execute"},
			body: func(mismatch bool) string {
				return `{"workspace_id":"` + batchDWorkspace(mismatch) + `","application_id":"app_flow_copilot","expected_session_version":1,"client_turn_key":"turn_demo","input_text":"bounded input","condition_values":{},"model":"","temperature":null}`
			},
			prepare: func(request *http.Request, _ bool) { request.SetPathValue("session_id", "appsess_aaaaaaaaaaaaaaaa") },
			server:  sessionServer,
			handle: func(server *Server, writer http.ResponseWriter, request *http.Request) {
				server.handleExecuteApplicationInteractionTurn(writer, request)
			},
		},
		{
			name: "saved workflow draft run", route: workflowExecutorStartRoute,
			permissions: []string{"workflow_runs:execute", "workflow_drafts:read"},
			body: func(mismatch bool) string {
				return `{"workspace_id":"` + batchDWorkspace(mismatch) + `","application_id":"app_flow_copilot","input_text":"bounded input","condition_values":{},"model":"","temperature":null}`
			},
			prepare: func(request *http.Request, mismatch bool) {
				runHeaders(request, mismatch)
				request.SetPathValue("draft_id", "draft_demo")
			},
			server: func() (*Server, *atomic.Int64, *workflowExecutorTestBridge) {
				owner := &countingBatchDDraftStore{}
				gateway := &workflowExecutorTestBridge{}
				return &Server{
					config:                  config.Config{WorkflowExecutorDevEnabled: true},
					savedWorkflowDraftStore: owner, workspaceMembershipProvider: newDeterministicDevTestWorkspaceMembershipProvider(),
					bridge: gateway,
				}, &owner.calls, gateway
			},
			handle: func(server *Server, writer http.ResponseWriter, request *http.Request) {
				server.handleStartWorkflowRun(writer, request)
			},
		},
		{
			name: "workflow definition run", route: workflowDefinitionRunCreateRoute,
			permissions: []string{"workflow_runs:execute", "workflow_definitions:read"},
			body: func(mismatch bool) string {
				return `{"workspace_id":"` + batchDWorkspace(mismatch) + `","application_id":"app_flow_copilot","definition_id":"definition_demo","expected_pointer_version":1,"expected_definition_version":1,"expected_definition_digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","input_text":"bounded input","condition_values":{},"model":"","temperature":null}`
			},
			prepare: runHeaders,
			server: func() (*Server, *atomic.Int64, *workflowExecutorTestBridge) {
				owner := &countingBatchCWorkflowDefinitionRepository{}
				gateway := &workflowExecutorTestBridge{}
				return &Server{
					config:                              config.Config{WorkflowDefinitionReleaseDevEnabled: true, WorkflowExecutorDevEnabled: true},
					workflowDefinitionReleaseRepository: owner, workspaceMembershipProvider: newDeterministicDevTestWorkspaceMembershipProvider(),
					bridge: gateway,
				}, &owner.calls, gateway
			},
			handle: func(server *Server, writer http.ResponseWriter, request *http.Request) {
				server.handleStartWorkflowDefinitionRun(writer, request)
			},
		},
	}
}

func batchDWorkspace(mismatch bool) string {
	if mismatch {
		return "workspace_other"
	}
	return "workspace_demo"
}

type countingBatchDSessionRepository struct{ calls atomic.Int64 }

func (repository *countingBatchDSessionRepository) Create(ApplicationInteractionContext, ApplicationInteractionSession) (ApplicationInteractionSession, error) {
	repository.calls.Add(1)
	return ApplicationInteractionSession{}, nil
}
func (repository *countingBatchDSessionRepository) Read(ApplicationInteractionContext, string) (ApplicationInteractionSession, error) {
	repository.calls.Add(1)
	return ApplicationInteractionSession{}, nil
}
func (repository *countingBatchDSessionRepository) List(ApplicationInteractionContext, applicationInteractionSessionListQuery) ([]ApplicationInteractionSession, error) {
	repository.calls.Add(1)
	return nil, nil
}
func (repository *countingBatchDSessionRepository) Close(ApplicationInteractionContext, string, int, ApplicationInteractionSession) (ApplicationInteractionSession, error) {
	repository.calls.Add(1)
	return ApplicationInteractionSession{}, nil
}
func (repository *countingBatchDSessionRepository) ReserveTurn(ApplicationInteractionContext, int, ApplicationInteractionSession, ApplicationInteractionTurn) (ApplicationInteractionSession, ApplicationInteractionTurn, bool, error) {
	repository.calls.Add(1)
	return ApplicationInteractionSession{}, ApplicationInteractionTurn{}, false, nil
}
func (repository *countingBatchDSessionRepository) ReadTurn(ApplicationInteractionContext, string, string) (ApplicationInteractionTurn, error) {
	repository.calls.Add(1)
	return ApplicationInteractionTurn{}, nil
}
func (repository *countingBatchDSessionRepository) ReadTurnByClientKey(ApplicationInteractionContext, string, string) (ApplicationInteractionTurn, error) {
	repository.calls.Add(1)
	return ApplicationInteractionTurn{}, nil
}
func (repository *countingBatchDSessionRepository) CompleteTurn(ApplicationInteractionContext, ApplicationInteractionTurn) (ApplicationInteractionSession, ApplicationInteractionTurn, bool, error) {
	repository.calls.Add(1)
	return ApplicationInteractionSession{}, ApplicationInteractionTurn{}, false, nil
}
func (repository *countingBatchDSessionRepository) ListTurns(ApplicationInteractionContext, string) ([]ApplicationInteractionTurn, error) {
	repository.calls.Add(1)
	return nil, nil
}

type countingBatchDDraftStore struct{ calls atomic.Int64 }

func (store *countingBatchDDraftStore) ReadDraftByID(SavedWorkflowDraftContext, string) (SavedWorkflowDraft, bool, error) {
	store.calls.Add(1)
	return SavedWorkflowDraft{}, false, nil
}
func (store *countingBatchDDraftStore) ListDraftSummariesByScope(SavedWorkflowDraftContext) ([]SavedWorkflowDraftSummary, error) {
	store.calls.Add(1)
	return nil, nil
}
func (store *countingBatchDDraftStore) WriteDraft(SavedWorkflowDraftContext, SavedWorkflowDraft, int) (int, error) {
	store.calls.Add(1)
	return 0, nil
}
func (store *countingBatchDDraftStore) SideEffects() SavedWorkflowDraftSideEffects {
	store.calls.Add(1)
	return SavedWorkflowDraftSideEffects{}
}
