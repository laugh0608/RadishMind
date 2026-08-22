package httpapi

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
)

func TestBatchCPermissionProjectionCoversEveryMutationGrant(t *testing.T) {
	upstreamPermissions := []string{
		"radishmind.application-publish-candidates.write",
		"radishmind.application-publish-candidates.review",
		"radishmind.workflow-definitions.write",
		"radishmind.workflow-definitions.review",
		"radishmind.workflow-definitions.activate",
		"radishmind.workflow-rag-promotions.read",
		"radishmind.workflow-rag-promotions.write",
		"radishmind.workflow-rag-promotions.review",
		"radishmind.workflow-rag-evaluation-datasets.read",
		"radishmind.workflow-rag-snapshots.read",
		"radishmind.workflow-rag-runtime.write",
		"radishmind.prompt-application-templates.read-source",
		"radishmind.prompt-application-runtime.write",
		"radishmind.agent-copilot-profiles.read-source",
		"radishmind.agent-copilot-runtime.write",
	}
	expectedGrants := []string{
		"application_publish_candidates:write",
		"application_publish_candidates:review",
		"workflow_definitions:write",
		"workflow_definitions:review",
		"workflow_definitions:activate",
		"workflow_rag_promotions:read",
		"workflow_rag_promotions:write",
		"workflow_rag_promotions:review",
		"workflow_rag_evaluation_datasets:read",
		"workflow_rag_snapshots:read",
		"workflow_rag_runtime:write",
		"prompt_application_templates:read_source",
		"prompt_application_runtime:write",
		"agent_copilot_profiles:read_source",
		"agent_copilot_runtime:write",
	}

	grants := projectControlPlaneReadPermissions(upstreamPermissions)
	for _, expected := range expectedGrants {
		if !controlPlaneReadHasScope(grants, expected) {
			t.Fatalf("Batch C upstream permission did not project %q: %#v", expected, grants)
		}
	}
}

func TestBatchCMutationAuthorizationDenialsDoNotReachPrimaryOwners(t *testing.T) {
	now := time.Now().UTC()
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

	for _, operation := range batchCMutationAuthorizationOperations() {
		t.Run(operation.name, func(t *testing.T) {
			for _, test := range tests {
				t.Run(test.name, func(t *testing.T) {
					server, ownerCalls := operation.server()
					body := operation.body(test.mismatch)
					if test.malformedBody {
						body = `{`
					}
					request := httptest.NewRequest(http.MethodPost, strings.TrimPrefix(operation.target, "POST "), strings.NewReader(body))
					operation.prepare(request, test.mismatch)
					if test.active != "" {
						request.Header.Set(activeWorkspaceHeader, test.active)
					}
					auth := batchBMutationAuth(now, operation.permissions...)
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
						t.Fatalf("authorization denial reached %s primary owner %d times", operation.name, ownerCalls.Load())
					}
				})
			}
		})
	}
}

func TestBatchCRAGPromotionPermissionsUseSingleMembershipDecision(t *testing.T) {
	required := []string{
		"workflow_rag_promotions:write",
		"workflow_rag_evaluation_datasets:read",
		"workflow_rag_snapshots:read",
		"application_drafts:read",
	}
	tests := []struct {
		name                  string
		identityPermissions   []string
		membershipPermissions []string
		expectedError         string
		expectedProviderCalls int64
	}{
		{
			name:                "identity fourth permission missing",
			identityPermissions: required[:3], membershipPermissions: required,
			expectedError: "scope_denied", expectedProviderCalls: 0,
		},
		{
			name:                "membership fourth permission missing",
			identityPermissions: required, membershipPermissions: required[:3],
			expectedError: "workspace_permission_denied", expectedProviderCalls: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			owner := &countingBatchCPromotionRepository{}
			provider := &countingWorkspaceMembershipProvider{delegate: newDeterministicDevTestWorkspaceMembershipProvider()}
			server := &Server{
				config:                         config.Config{WorkflowRAGPromotionDevEnabled: true},
				workflowRAGPromotionRepository: owner, workspaceMembershipProvider: provider,
			}
			request := httptest.NewRequest(http.MethodPost, strings.TrimPrefix(workflowRAGPromotionCandidateCreateRoute, "POST "), strings.NewReader(`{}`))
			request.Header.Set(activeWorkspaceHeader, "workspace_demo")
			request.Header.Set(savedWorkflowDraftDevWorkspaceHeader, "workspace_demo")
			request.Header.Set(savedWorkflowDraftDevApplicationHeader, "app_flow_copilot")
			auth := batchBMutationAuth(time.Now().UTC(), test.identityPermissions...)
			auth.WorkspaceMemberships[0].PermissionGrants = append([]string{}, test.membershipPermissions...)
			request = request.WithContext(withControlPlaneReadFakeAuthContext(request.Context(), auth))
			recorder := httptest.NewRecorder()

			server.handleCreateWorkflowRAGPromotionCandidate(recorder, request)

			if recorder.Code != http.StatusForbidden || !strings.Contains(recorder.Body.String(), test.expectedError) {
				t.Fatalf("unexpected composite denial: status=%d body=%s", recorder.Code, recorder.Body.String())
			}
			if provider.calls.Load() != test.expectedProviderCalls || owner.calls.Load() != 0 {
				t.Fatalf("RAG promotion authorization was not atomic: provider=%d owner=%d", provider.calls.Load(), owner.calls.Load())
			}
		})
	}
}

type batchCMutationAuthorizationOperation struct {
	name        string
	target      string
	permissions []string
	body        func(bool) string
	prepare     func(*http.Request, bool)
	server      func() (*Server, *atomic.Int64)
	handle      func(*Server, http.ResponseWriter, *http.Request)
}

func batchCMutationAuthorizationOperations() []batchCMutationAuthorizationOperation {
	applicationHeaders := func(request *http.Request, mismatch bool) {
		request.Header.Set(applicationPublishDevWorkspaceHeader, batchCWorkspace(mismatch))
		request.Header.Set(applicationPublishDevApplicationHeader, "app_flow_copilot")
	}
	workflowHeaders := func(request *http.Request, mismatch bool) {
		request.Header.Set(savedWorkflowDraftDevWorkspaceHeader, batchCWorkspace(mismatch))
		request.Header.Set(savedWorkflowDraftDevApplicationHeader, "app_flow_copilot")
	}
	promotionBody := func(mismatch bool) string {
		return `{"workspace_id":"` + batchCWorkspace(mismatch) + `","application_id":"app_flow_copilot","dataset_id":"dataset_demo","dataset_version":1,"dataset_digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","candidate_review_id":"review_demo","draft_id":"draft_demo","expected_draft_version":1}`
	}
	return []batchCMutationAuthorizationOperation{
		{
			name: "application publish create", target: applicationPublishCandidateCreateRoute,
			permissions: []string{"application_publish_candidates:write"},
			body: func(bool) string {
				return `{"candidate_id":"candidate_demo","draft_id":"draft_demo","expected_draft_version":1,"evidence_request_ids":[]}`
			},
			prepare: applicationHeaders,
			server: func() (*Server, *atomic.Int64) {
				owner := &countingApplicationConfigurationDraftRepository{}
				return &Server{
					config:                     config.Config{ApplicationPublishDevHTTPEnabled: true, ApplicationPublishDevWriteEnabled: true},
					applicationDraftRepository: owner, workspaceMembershipProvider: newDeterministicDevTestWorkspaceMembershipProvider(),
				}, &owner.calls
			},
			handle: (*Server).handleCreateApplicationPublishCandidate,
		},
		{
			name: "application publish review", target: "/v1/user-workspace/application-publish-candidates/candidate_demo/reviews",
			permissions: []string{"application_publish_candidates:review"},
			body: func(bool) string {
				return `{"expected_review_version":0,"decision":"reject","reason":"Reviewed candidate evidence."}`
			},
			prepare: func(request *http.Request, mismatch bool) {
				applicationHeaders(request, mismatch)
				request.SetPathValue("candidate_id", "candidate_demo")
			},
			server: func() (*Server, *atomic.Int64) {
				owner := &countingBatchCApplicationPublishRepository{}
				return &Server{
					config:                                config.Config{ApplicationPublishDevHTTPEnabled: true, ApplicationPublishDevWriteEnabled: true},
					applicationPublishCandidateRepository: owner, workspaceMembershipProvider: newDeterministicDevTestWorkspaceMembershipProvider(),
				}, &owner.calls
			},
			handle: (*Server).handleReviewApplicationPublishCandidate,
		},
		{
			name: "workflow definition create", target: workflowDefinitionCandidateCreateRoute,
			permissions: []string{"workflow_definitions:write"},
			body: func(bool) string {
				return `{"candidate_id":"candidate_demo","definition_id":"definition_demo","draft_id":"draft_demo","expected_draft_version":1}`
			},
			prepare: workflowHeaders,
			server: func() (*Server, *atomic.Int64) {
				owner := &countingSavedWorkflowDraftStore{}
				return &Server{
					config:                  config.Config{WorkflowDefinitionReleaseDevEnabled: true},
					savedWorkflowDraftStore: owner, workspaceMembershipProvider: newDeterministicDevTestWorkspaceMembershipProvider(),
				}, &owner.calls
			},
			handle: (*Server).handleCreateWorkflowDefinitionCandidate,
		},
		{
			name: "workflow definition review", target: "/v1/user-workspace/workflow-definition-candidates/candidate_demo/decisions",
			permissions: []string{"workflow_definitions:review"},
			body: func(bool) string {
				return `{"expected_review_version":0,"decision":"reject","reason":"Reviewed definition candidate."}`
			},
			prepare: func(request *http.Request, mismatch bool) {
				workflowHeaders(request, mismatch)
				request.SetPathValue("candidate_id", "candidate_demo")
			},
			server: batchCWorkflowDefinitionServer,
			handle: (*Server).handleDecideWorkflowDefinitionCandidate,
		},
		{
			name: "workflow definition activation", target: "/v1/user-workspace/workflow-definitions/definition_demo/activation-decisions",
			permissions: []string{"workflow_definitions:activate"},
			body: func(bool) string {
				return `{"expected_pointer_version":0,"decision":"activate","version":1,"reason":"Activate reviewed definition."}`
			},
			prepare: func(request *http.Request, mismatch bool) {
				workflowHeaders(request, mismatch)
				request.SetPathValue("definition_id", "definition_demo")
			},
			server: batchCWorkflowDefinitionServer,
			handle: (*Server).handleDecideWorkflowDefinitionActivation,
		},
		{
			name: "workflow RAG promotion create", target: workflowRAGPromotionCandidateCreateRoute,
			permissions: []string{"workflow_rag_promotions:write", "workflow_rag_evaluation_datasets:read", "workflow_rag_snapshots:read", "application_drafts:read"},
			body:        promotionBody, prepare: workflowHeaders,
			server: batchCWorkflowRAGPromotionServer,
			handle: (*Server).handleCreateWorkflowRAGPromotionCandidate,
		},
		{
			name: "workflow RAG promotion review", target: "/v1/user-workspace/workflow-rag-knowledge-promotion-candidates/candidate_demo/decisions",
			permissions: []string{"workflow_rag_promotions:review"},
			body: func(mismatch bool) string {
				return `{"workspace_id":"` + batchCWorkspace(mismatch) + `","application_id":"app_flow_copilot","expected_record_version":1,"decision":"reject","reason":"Reviewed promotion evidence."}`
			},
			prepare: func(request *http.Request, mismatch bool) {
				workflowHeaders(request, false)
				request.SetPathValue("candidate_id", "candidate_demo")
			},
			server: batchCWorkflowRAGPromotionServer,
			handle: (*Server).handleDecideWorkflowRAGPromotionCandidate,
		},
		batchCRuntimeOperation("workflow RAG assignment", "workflow_rag_runtime:write", workflowRAGApplicationRuntimeAssignmentDecisionRoute),
		batchCRuntimeOperation("Prompt assignment", "prompt_application_runtime:write", promptApplicationRuntimeDecisionRoute),
		batchCRuntimeOperation("Agent Copilot assignment", "agent_copilot_runtime:write", agentCopilotRuntimeDecisionRoute),
	}
}

func batchCRuntimeOperation(name, permission, target string) batchCMutationAuthorizationOperation {
	return batchCMutationAuthorizationOperation{
		name: name, target: target, permissions: []string{permission},
		body: func(mismatch bool) string {
			if permission == "workflow_rag_runtime:write" {
				return `{"workspace_id":"` + batchCWorkspace(mismatch) + `","expected_record_version":0,"decision":"activate","publish_candidate_id":"candidate_demo","reason":"Activate reviewed candidate."}`
			}
			return `{"workspace_id":"` + batchCWorkspace(mismatch) + `","expected_assignment_version":0,"action":"activate","candidate_id":"candidate_demo"}`
		},
		prepare: func(request *http.Request, _ bool) {
			request.SetPathValue("application_id", "app_flow_copilot")
			switch permission {
			case "workflow_rag_runtime:write":
				request.Header.Set(savedWorkflowDraftDevWorkspaceHeader, "workspace_demo")
				request.Header.Set(savedWorkflowDraftDevApplicationHeader, "app_flow_copilot")
			case "prompt_application_runtime:write":
				request.Header.Set(promptApplicationRuntimeWorkspaceHeader, "workspace_demo")
				request.Header.Set(promptApplicationRuntimeApplicationHeader, "app_flow_copilot")
			case "agent_copilot_runtime:write":
				request.Header.Set(agentCopilotRuntimeWorkspaceHeader, "workspace_demo")
				request.Header.Set(agentCopilotRuntimeApplicationHeader, "app_flow_copilot")
			}
		},
		server: func() (*Server, *atomic.Int64) {
			switch permission {
			case "workflow_rag_runtime:write":
				owner := &countingBatchCWorkflowRAGRuntimeRepository{}
				return &Server{
					config:                          config.Config{WorkflowRAGAppInvocationDevEnabled: true},
					workflowRAGAppRuntimeRepository: owner, workspaceMembershipProvider: newDeterministicDevTestWorkspaceMembershipProvider(),
				}, &owner.calls
			case "prompt_application_runtime:write":
				owner := &countingBatchCPromptRuntimeRepository{}
				return &Server{
					config:                             config.Config{PromptApplicationRuntimeDevHTTPEnabled: true, PromptApplicationRuntimeDevWriteEnabled: true},
					promptApplicationRuntimeRepository: owner, workspaceMembershipProvider: newDeterministicDevTestWorkspaceMembershipProvider(),
				}, &owner.calls
			default:
				owner := &countingBatchCAgentRuntimeRepository{}
				return &Server{
					config:                        config.Config{AgentCopilotRuntimeDevHTTPEnabled: true, AgentCopilotRuntimeDevWriteEnabled: true},
					agentCopilotRuntimeRepository: owner, workspaceMembershipProvider: newDeterministicDevTestWorkspaceMembershipProvider(),
				}, &owner.calls
			}
		},
		handle: func(server *Server, writer http.ResponseWriter, request *http.Request) {
			switch permission {
			case "workflow_rag_runtime:write":
				server.handleDecideWorkflowRAGApplicationRuntimeAssignment(writer, request)
			case "prompt_application_runtime:write":
				server.handleDecidePromptApplicationRuntimeAssignment(writer, request)
			default:
				server.handleDecideAgentCopilotRuntimeAssignment(writer, request)
			}
		},
	}
}

func batchCWorkspace(mismatch bool) string {
	if mismatch {
		return "workspace_other"
	}
	return "workspace_demo"
}

func batchCWorkflowDefinitionServer() (*Server, *atomic.Int64) {
	owner := &countingBatchCWorkflowDefinitionRepository{}
	return &Server{
		config:                              config.Config{WorkflowDefinitionReleaseDevEnabled: true},
		workflowDefinitionReleaseRepository: owner, workspaceMembershipProvider: newDeterministicDevTestWorkspaceMembershipProvider(),
	}, &owner.calls
}

func batchCWorkflowRAGPromotionServer() (*Server, *atomic.Int64) {
	owner := &countingBatchCPromotionRepository{}
	return &Server{
		config:                         config.Config{WorkflowRAGPromotionDevEnabled: true},
		workflowRAGPromotionRepository: owner, workspaceMembershipProvider: newDeterministicDevTestWorkspaceMembershipProvider(),
	}, &owner.calls
}

type countingBatchCApplicationPublishRepository struct{ calls atomic.Int64 }

func (repository *countingBatchCApplicationPublishRepository) Create(ApplicationPublishContext, ApplicationPublishCandidate) (ApplicationPublishCandidate, error) {
	repository.calls.Add(1)
	return ApplicationPublishCandidate{}, errors.New("unexpected owner call")
}
func (repository *countingBatchCApplicationPublishRepository) Read(ApplicationPublishContext, string) (ApplicationPublishCandidate, error) {
	repository.calls.Add(1)
	return ApplicationPublishCandidate{}, errors.New("unexpected owner call")
}
func (repository *countingBatchCApplicationPublishRepository) List(ApplicationPublishContext) ([]ApplicationPublishCandidate, error) {
	repository.calls.Add(1)
	return nil, errors.New("unexpected owner call")
}
func (repository *countingBatchCApplicationPublishRepository) AppendReview(ApplicationPublishContext, string, int, ApplicationPublishReviewRecord) (ApplicationPublishCandidate, error) {
	repository.calls.Add(1)
	return ApplicationPublishCandidate{}, errors.New("unexpected owner call")
}

type countingBatchCWorkflowDefinitionRepository struct{ calls atomic.Int64 }

func (repository *countingBatchCWorkflowDefinitionRepository) CreateCandidate(WorkflowDefinitionReleaseContext, string, string, string, SavedWorkflowDraft, time.Time) (WorkflowDefinitionReleaseCandidate, error) {
	repository.calls.Add(1)
	return WorkflowDefinitionReleaseCandidate{}, errors.New("unexpected owner call")
}
func (repository *countingBatchCWorkflowDefinitionRepository) Review(WorkflowDefinitionReleaseContext, string, int, string, string, string, time.Time) (WorkflowDefinitionReleaseCandidate, *WorkflowDefinitionVersion, error) {
	repository.calls.Add(1)
	return WorkflowDefinitionReleaseCandidate{}, nil, errors.New("unexpected owner call")
}
func (repository *countingBatchCWorkflowDefinitionRepository) DecideActivation(WorkflowDefinitionReleaseContext, string, int, string, int, string, time.Time) (WorkflowDefinitionActivation, error) {
	repository.calls.Add(1)
	return WorkflowDefinitionActivation{}, errors.New("unexpected owner call")
}
func (repository *countingBatchCWorkflowDefinitionRepository) ReadCandidate(WorkflowDefinitionReleaseContext, string) (WorkflowDefinitionReleaseCandidate, error) {
	repository.calls.Add(1)
	return WorkflowDefinitionReleaseCandidate{}, errors.New("unexpected owner call")
}
func (repository *countingBatchCWorkflowDefinitionRepository) ListCandidates(WorkflowDefinitionReleaseContext) ([]WorkflowDefinitionReleaseCandidate, error) {
	repository.calls.Add(1)
	return nil, errors.New("unexpected owner call")
}
func (repository *countingBatchCWorkflowDefinitionRepository) ListVersions(WorkflowDefinitionReleaseContext, string) ([]WorkflowDefinitionVersion, error) {
	repository.calls.Add(1)
	return nil, errors.New("unexpected owner call")
}
func (repository *countingBatchCWorkflowDefinitionRepository) ReadVersion(WorkflowDefinitionReleaseContext, string, int) (WorkflowDefinitionVersion, error) {
	repository.calls.Add(1)
	return WorkflowDefinitionVersion{}, errors.New("unexpected owner call")
}
func (repository *countingBatchCWorkflowDefinitionRepository) ReadActivation(WorkflowDefinitionReleaseContext, string) (WorkflowDefinitionActivation, error) {
	repository.calls.Add(1)
	return WorkflowDefinitionActivation{}, errors.New("unexpected owner call")
}
func (repository *countingBatchCWorkflowDefinitionRepository) ListSummaries(ReadRepositoryContext, ListWorkflowDefinitionSummariesRequest) ListWorkflowDefinitionSummariesResult {
	repository.calls.Add(1)
	return ListWorkflowDefinitionSummariesResult{FailureCode: "unexpected_owner_call"}
}

type countingBatchCPromotionRepository struct{ calls atomic.Int64 }

func (repository *countingBatchCPromotionRepository) Create(WorkflowRAGPromotionContext, WorkflowRAGKnowledgePromotionCandidate, WorkflowRAGPromotionAudit) error {
	repository.calls.Add(1)
	return errors.New("unexpected owner call")
}
func (repository *countingBatchCPromotionRepository) Read(WorkflowRAGPromotionContext, string) (WorkflowRAGKnowledgePromotionCandidate, []WorkflowRAGKnowledgePromotionDecision, *WorkflowRAGApplicationBinding, []WorkflowRAGPromotionAudit, error) {
	repository.calls.Add(1)
	return WorkflowRAGKnowledgePromotionCandidate{}, nil, nil, nil, errors.New("unexpected owner call")
}
func (repository *countingBatchCPromotionRepository) ReadBinding(WorkflowRAGPromotionContext, WorkflowRAGApplicationBindingRef) (WorkflowRAGKnowledgePromotionCandidate, WorkflowRAGApplicationBinding, error) {
	repository.calls.Add(1)
	return WorkflowRAGKnowledgePromotionCandidate{}, WorkflowRAGApplicationBinding{}, errors.New("unexpected owner call")
}
func (repository *countingBatchCPromotionRepository) List(WorkflowRAGPromotionContext, workflowRAGPromotionListQuery) ([]WorkflowRAGKnowledgePromotionCandidate, error) {
	repository.calls.Add(1)
	return nil, errors.New("unexpected owner call")
}
func (repository *countingBatchCPromotionRepository) AppendDecision(WorkflowRAGPromotionContext, string, int, WorkflowRAGKnowledgePromotionCandidate, WorkflowRAGKnowledgePromotionDecision, *WorkflowRAGApplicationBinding, []WorkflowRAGPromotionAudit) error {
	repository.calls.Add(1)
	return errors.New("unexpected owner call")
}

type countingBatchCWorkflowRAGRuntimeRepository struct{ calls atomic.Int64 }

func (repository *countingBatchCWorkflowRAGRuntimeRepository) Read(WorkflowRAGApplicationRuntimeContext) (WorkflowRAGApplicationRuntimeAssignment, []WorkflowRAGApplicationRuntimeEvent, []WorkflowRAGApplicationRuntimeAudit, error) {
	repository.calls.Add(1)
	return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, errors.New("unexpected owner call")
}
func (repository *countingBatchCWorkflowRAGRuntimeRepository) Apply(WorkflowRAGApplicationRuntimeContext, int, WorkflowRAGApplicationRuntimeAssignment, WorkflowRAGApplicationRuntimeEvent, WorkflowRAGApplicationRuntimeAudit) error {
	repository.calls.Add(1)
	return errors.New("unexpected owner call")
}

type countingBatchCPromptRuntimeRepository struct{ calls atomic.Int64 }

func (repository *countingBatchCPromptRuntimeRepository) Read(PromptApplicationRuntimeContext) (PromptApplicationRuntimeAssignment, []PromptApplicationRuntimeAssignmentEvent, error) {
	repository.calls.Add(1)
	return PromptApplicationRuntimeAssignment{}, nil, errors.New("unexpected owner call")
}
func (repository *countingBatchCPromptRuntimeRepository) Apply(PromptApplicationRuntimeContext, int, PromptApplicationRuntimeAssignment, PromptApplicationRuntimeAssignmentEvent) error {
	repository.calls.Add(1)
	return errors.New("unexpected owner call")
}

type countingBatchCAgentRuntimeRepository struct{ calls atomic.Int64 }

func (repository *countingBatchCAgentRuntimeRepository) Read(AgentCopilotRuntimeContext) (AgentCopilotRuntimeAssignmentV1, []AgentCopilotRuntimeAssignmentEventV1, error) {
	repository.calls.Add(1)
	return AgentCopilotRuntimeAssignmentV1{}, nil, errors.New("unexpected owner call")
}
func (repository *countingBatchCAgentRuntimeRepository) Apply(AgentCopilotRuntimeContext, int, AgentCopilotRuntimeAssignmentV1, AgentCopilotRuntimeAssignmentEventV1) error {
	repository.calls.Add(1)
	return errors.New("unexpected owner call")
}
