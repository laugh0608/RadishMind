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

func TestBatchEPermissionProjectionCoversEveryMutationGrant(t *testing.T) {
	grants := projectControlPlaneReadPermissions([]string{
		"radishmind.workflow-runs.read",
		"radishmind.workflow-runs.execute",
		"radishmind.workflow-drafts.read",
		"radishmind.workflow-rag-snapshots.read",
		"radishmind.workflow-rag-snapshots.write",
		"radishmind.workflow-rag-snapshots.archive",
		"radishmind.workflow-rag.execute",
		"radishmind.workflow-rag-evaluation-datasets.read",
		"radishmind.workflow-rag-evaluation-datasets.write",
		"radishmind.workflow-rag-evaluation-datasets.review",
		"radishmind.workflow-rag-evaluation-datasets.archive",
		"radishmind.workflow-tool-actions.plan",
		"radishmind.workflow-tool-actions.confirm",
		"radishmind.workflow-tool-actions.execute",
		"radishmind.workflow-definitions.read",
		"radishmind.workflow-evaluations.write",
	})
	for _, expected := range []string{
		"workflow_runs:read",
		"workflow_runs:execute",
		"workflow_drafts:read",
		"workflow_rag_snapshots:read",
		"workflow_rag_snapshots:write",
		"workflow_rag_snapshots:archive",
		"workflow_rag:execute",
		"workflow_rag_evaluation_datasets:read",
		"workflow_rag_evaluation_datasets:write",
		"workflow_rag_evaluation_datasets:review",
		"workflow_rag_evaluation_datasets:archive",
		"workflow_tool_actions:plan",
		"workflow_tool_actions:confirm",
		"workflow_tool_actions:execute",
		"workflow_definitions:read",
		"workflow_evaluations:write",
	} {
		if !controlPlaneReadHasScope(grants, expected) {
			t.Fatalf("Batch E upstream permission did not project %q: %#v", expected, grants)
		}
	}
}

func TestBatchEMutationAuthorizationDenialsHaveZeroOwnerAndGatewayCalls(t *testing.T) {
	tests := []struct {
		name          string
		active        string
		malformedBody bool
		bodyWorkspace string
		legacyApp     string
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
			name: "body workspace mismatch", active: "workspace_demo", bodyWorkspace: "workspace_other",
			expectedCode: http.StatusForbidden, expectedError: "workspace_binding_mismatch",
		},
		{
			name: "legacy application mismatch", active: "workspace_demo", legacyApp: "app_other",
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

	for _, operation := range batchEMutationAuthorizationOperations() {
		t.Run(operation.name, func(t *testing.T) {
			for _, test := range tests {
				t.Run(test.name, func(t *testing.T) {
					server, ownerCalls, gateway := newBatchEMutationAuthorizationServer()
					workspace := test.bodyWorkspace
					if workspace == "" {
						workspace = "workspace_demo"
					}
					body := `{"workspace_id":"` + workspace + `","application_id":"app_flow_copilot"}`
					if test.malformedBody {
						body = `{`
					}
					request := httptest.NewRequest(http.MethodPost, strings.TrimPrefix(operation.route, "POST "), strings.NewReader(body))
					operation.prepare(request)
					if test.legacyApp != "" {
						request.Header.Set(savedWorkflowDraftDevApplicationHeader, test.legacyApp)
					}
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

func TestBatchECompositePermissionsUseSingleMembershipDecision(t *testing.T) {
	for _, operation := range batchEMutationAuthorizationOperations() {
		if len(operation.permissions) < 2 {
			continue
		}
		t.Run(operation.name, func(t *testing.T) {
			for _, test := range []struct {
				name                  string
				identityPermissions   []string
				membershipPermissions []string
				expectedError         string
				expectedProviderCalls int64
			}{
				{
					name:                  "identity final permission missing",
					identityPermissions:   operation.permissions[:len(operation.permissions)-1],
					membershipPermissions: operation.permissions,
					expectedError:         "scope_denied", expectedProviderCalls: 0,
				},
				{
					name:                  "membership final permission missing",
					identityPermissions:   operation.permissions,
					membershipPermissions: operation.permissions[:len(operation.permissions)-1],
					expectedError:         "workspace_permission_denied", expectedProviderCalls: 1,
				},
			} {
				t.Run(test.name, func(t *testing.T) {
					server, ownerCalls, gateway := newBatchEMutationAuthorizationServer()
					provider := &countingWorkspaceMembershipProvider{delegate: newDeterministicDevTestWorkspaceMembershipProvider()}
					server.workspaceMembershipProvider = provider
					request := httptest.NewRequest(http.MethodPost, strings.TrimPrefix(operation.route, "POST "),
						strings.NewReader(`{"workspace_id":"workspace_demo","application_id":"app_flow_copilot"}`))
					request.Header.Set(activeWorkspaceHeader, "workspace_demo")
					operation.prepare(request)
					auth := batchBMutationAuth(time.Now().UTC(), test.identityPermissions...)
					auth.WorkspaceMemberships[0].PermissionGrants = append([]string{}, test.membershipPermissions...)
					request = request.WithContext(withControlPlaneReadFakeAuthContext(request.Context(), auth))
					recorder := httptest.NewRecorder()

					operation.handle(server, recorder, request)

					if recorder.Code != http.StatusForbidden || !strings.Contains(recorder.Body.String(), test.expectedError) {
						t.Fatalf("unexpected composite denial: status=%d body=%s", recorder.Code, recorder.Body.String())
					}
					if provider.calls.Load() != test.expectedProviderCalls || ownerCalls.Load() != 0 || gateway.callCount() != 0 {
						t.Fatalf("authorization was not atomic: provider=%d owner=%d gateway=%d", provider.calls.Load(), ownerCalls.Load(), gateway.callCount())
					}
				})
			}
		})
	}
}

type batchEMutationAuthorizationOperation struct {
	name        string
	route       string
	permissions []string
	prepare     func(*http.Request)
	handle      func(*Server, http.ResponseWriter, *http.Request)
}

func batchEMutationAuthorizationOperations() []batchEMutationAuthorizationOperation {
	prepare := func(pathKey, pathValue string) func(*http.Request) {
		return func(request *http.Request) {
			request.Header.Set(savedWorkflowDraftDevWorkspaceHeader, "workspace_demo")
			request.Header.Set(savedWorkflowDraftDevApplicationHeader, "app_flow_copilot")
			if pathKey != "" {
				request.SetPathValue(pathKey, pathValue)
			}
		}
	}
	return []batchEMutationAuthorizationOperation{
		{name: "RAG retrieval execution", route: "POST /v1/user-workspace/workflow-drafts/{draft_id}/retrieval-executions", permissions: []string{"workflow_rag:execute", "workflow_runs:execute", "workflow_drafts:read", "workflow_rag_snapshots:read"}, prepare: prepare("draft_id", "draft_demo"), handle: func(s *Server, w http.ResponseWriter, r *http.Request) { s.handleWorkflowRAGExecution(w, r) }},
		{name: "RAG snapshot create", route: workflowRAGSnapshotCreateRoute, permissions: []string{"workflow_rag_snapshots:write"}, prepare: prepare("", ""), handle: func(s *Server, w http.ResponseWriter, r *http.Request) { s.handleCreateWorkflowRAGSnapshot(w, r) }},
		{name: "RAG snapshot version", route: workflowRAGSnapshotVersionRoute, permissions: []string{"workflow_rag_snapshots:write"}, prepare: prepare("snapshot_id", "rags_aaaaaaaaaaaaaaaa"), handle: func(s *Server, w http.ResponseWriter, r *http.Request) { s.handleVersionWorkflowRAGSnapshot(w, r) }},
		{name: "RAG snapshot archive", route: workflowRAGSnapshotArchiveRoute, permissions: []string{"workflow_rag_snapshots:archive"}, prepare: prepare("snapshot_id", "rags_aaaaaaaaaaaaaaaa"), handle: func(s *Server, w http.ResponseWriter, r *http.Request) { s.handleArchiveWorkflowRAGSnapshot(w, r) }},
		{name: "RAG evaluation dataset create", route: workflowRAGEvaluationDatasetCreateRoute, permissions: []string{"workflow_rag_evaluation_datasets:write", "workflow_rag_snapshots:read"}, prepare: prepare("", ""), handle: func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.handleCreateWorkflowRAGEvaluationDataset(w, r)
		}},
		{name: "RAG evaluation dataset version", route: workflowRAGEvaluationDatasetVersionRoute, permissions: []string{"workflow_rag_evaluation_datasets:write", "workflow_rag_snapshots:read"}, prepare: prepare("dataset_id", "raged_aaaaaaaaaaaaaaaa"), handle: func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.handleVersionWorkflowRAGEvaluationDataset(w, r)
		}},
		{name: "RAG evaluation dataset archive", route: workflowRAGEvaluationDatasetArchiveRoute, permissions: []string{"workflow_rag_evaluation_datasets:archive"}, prepare: prepare("dataset_id", "raged_aaaaaaaaaaaaaaaa"), handle: func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.handleArchiveWorkflowRAGEvaluationDataset(w, r)
		}},
		{name: "RAG candidate review create", route: workflowRAGCandidateReviewCreateRoute, permissions: []string{"workflow_rag_evaluation_datasets:review", "workflow_rag_evaluation_datasets:read", "workflow_rag_snapshots:read"}, prepare: prepare("dataset_id", "raged_aaaaaaaaaaaaaaaa"), handle: func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.handleCreateWorkflowRAGCandidateReview(w, r)
		}},
		{name: "HTTP tool plan", route: workflowHTTPToolPlanCreateRoute, permissions: []string{"workflow_drafts:read", "workflow_tool_actions:plan"}, prepare: prepare("draft_id", "draft_demo"), handle: func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.handleCreateWorkflowHTTPToolActionPlan(w, r)
		}},
		{name: "Definition HTTP tool plan", route: workflowDefinitionHTTPToolPlanCreateRoute, permissions: []string{"workflow_definitions:read", "workflow_tool_actions:plan"}, prepare: prepare("definition_id", "wfd_aaaaaaaaaaaaaaaa"), handle: func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.handleCreateWorkflowDefinitionHTTPToolActionPlan(w, r)
		}},
		{name: "HTTP tool decision", route: workflowHTTPToolDecisionRoute, permissions: []string{"workflow_tool_actions:confirm"}, prepare: prepare("plan_id", "wtap_aaaaaaaaaaaaaaaa"), handle: func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.handleDecideWorkflowHTTPToolActionPlan(w, r)
		}},
		{name: "HTTP tool execution", route: workflowHTTPToolExecutionRoute, permissions: []string{"workflow_tool_actions:execute", "workflow_runs:execute", "workflow_drafts:read"}, prepare: prepare("plan_id", "wtap_aaaaaaaaaaaaaaaa"), handle: func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.handleExecuteWorkflowHTTPToolActionPlan(w, r)
		}},
		{name: "workflow evaluation case create", route: workflowEvaluationCreateRoute, permissions: []string{"workflow_evaluations:write", "workflow_runs:read"}, prepare: prepare("", ""), handle: func(s *Server, w http.ResponseWriter, r *http.Request) { s.handleCreateWorkflowEvaluation(w, r) }},
		{name: "workflow evaluation case revision", route: workflowEvaluationRevisionCreateRoute, permissions: []string{"workflow_evaluations:write", "workflow_runs:read"}, prepare: prepare("case_id", "eval_aaaaaaaaaaaaaaaa"), handle: func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.handleCreateWorkflowEvaluationRevision(w, r)
		}},
		{name: "workflow evaluation suite create", route: workflowEvaluationSuiteCreateRoute, permissions: []string{"workflow_evaluations:write", "workflow_runs:read"}, prepare: prepare("", ""), handle: func(s *Server, w http.ResponseWriter, r *http.Request) { s.handleCreateWorkflowEvaluationSuite(w, r) }},
		{name: "workflow evaluation suite decision", route: workflowEvaluationDecisionCreateRoute, permissions: []string{"workflow_evaluations:write", "workflow_runs:read"}, prepare: prepare("suite_id", "suite_aaaaaaaaaaaaaaaa"), handle: func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.handleCreateWorkflowEvaluationDecision(w, r)
		}},
	}
}

func newBatchEMutationAuthorizationServer() (*Server, *atomic.Int64, *workflowExecutorTestBridge) {
	calls := &atomic.Int64{}
	gateway := &workflowExecutorTestBridge{}
	return &Server{
		config: config.Config{
			WorkflowExecutorDevEnabled:          true,
			WorkflowDefinitionReleaseDevEnabled: true,
			WorkflowRAGExecutionDevEnabled:      true,
			WorkflowRAGSnapshotDevEnabled:       true,
			WorkflowRAGEvaluationDevEnabled:     true,
			WorkflowToolActionDevEnabled:        true,
			WorkflowHTTPToolExecutionDevEnabled: true,
		},
		savedWorkflowDraftStore:                &countingBatchEDraftStore{calls: calls},
		workflowRAGSnapshotRepository:          &countingBatchESnapshotRepository{calls: calls},
		workflowRAGEvaluationDatasetRepository: &countingBatchEDatasetRepository{calls: calls},
		workflowHTTPToolActionStore:            &countingBatchEToolActionStore{calls: calls},
		workflowEvaluationStore:                &countingBatchEEvaluationStore{calls: calls},
		workflowEvaluationSuiteStore:           &countingBatchEEvaluationSuiteStore{calls: calls},
		workspaceMembershipProvider:            newDeterministicDevTestWorkspaceMembershipProvider(),
		bridge:                                 gateway,
	}, calls, gateway
}

type countingBatchEDraftStore struct{ calls *atomic.Int64 }

func (s *countingBatchEDraftStore) ReadDraftByID(SavedWorkflowDraftContext, string) (SavedWorkflowDraft, bool, error) {
	s.calls.Add(1)
	return SavedWorkflowDraft{}, false, nil
}
func (s *countingBatchEDraftStore) ListDraftSummariesByScope(SavedWorkflowDraftContext) ([]SavedWorkflowDraftSummary, error) {
	s.calls.Add(1)
	return nil, nil
}
func (s *countingBatchEDraftStore) WriteDraft(SavedWorkflowDraftContext, SavedWorkflowDraft, int) (int, error) {
	s.calls.Add(1)
	return 0, nil
}
func (s *countingBatchEDraftStore) SideEffects() SavedWorkflowDraftSideEffects {
	s.calls.Add(1)
	return SavedWorkflowDraftSideEffects{}
}

type countingBatchESnapshotRepository struct{ calls *atomic.Int64 }

func (r *countingBatchESnapshotRepository) Create(WorkflowRAGSnapshotContext, WorkflowRAGSnapshotResource, WorkflowRAGSnapshotRecord, WorkflowRAGExecutionAudit) error {
	r.calls.Add(1)
	return nil
}
func (r *countingBatchESnapshotRepository) List(WorkflowRAGSnapshotContext, workflowRAGSnapshotListQuery) ([]WorkflowRAGSnapshotResource, error) {
	r.calls.Add(1)
	return nil, nil
}
func (r *countingBatchESnapshotRepository) ReadLatest(WorkflowRAGSnapshotContext, string) (WorkflowRAGSnapshotResource, WorkflowRAGSnapshotRecord, error) {
	r.calls.Add(1)
	return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, nil
}
func (r *countingBatchESnapshotRepository) ReadVersion(WorkflowRAGSnapshotContext, string, int) (WorkflowRAGSnapshotResource, WorkflowRAGSnapshotRecord, error) {
	r.calls.Add(1)
	return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, nil
}
func (r *countingBatchESnapshotRepository) ReadByRAGRef(WorkflowRAGSnapshotContext, string) (WorkflowRAGSnapshotResource, WorkflowRAGSnapshotRecord, error) {
	r.calls.Add(1)
	return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, nil
}
func (r *countingBatchESnapshotRepository) CreateVersion(WorkflowRAGSnapshotContext, string, int, WorkflowRAGSnapshotResource, WorkflowRAGSnapshotRecord, WorkflowRAGExecutionAudit) error {
	r.calls.Add(1)
	return nil
}
func (r *countingBatchESnapshotRepository) Archive(WorkflowRAGSnapshotContext, string, int, WorkflowRAGSnapshotResource, WorkflowRAGExecutionAudit) error {
	r.calls.Add(1)
	return nil
}
func (r *countingBatchESnapshotRepository) AppendAudit(WorkflowRAGSnapshotContext, WorkflowRAGExecutionAudit) error {
	r.calls.Add(1)
	return nil
}

type countingBatchEDatasetRepository struct{ calls *atomic.Int64 }

func (r *countingBatchEDatasetRepository) Create(WorkflowRAGSnapshotContext, WorkflowRAGEvaluationDatasetResource, WorkflowRAGEvaluationDatasetVersion, WorkflowRAGEvaluationAudit) error {
	r.calls.Add(1)
	return nil
}
func (r *countingBatchEDatasetRepository) List(WorkflowRAGSnapshotContext, workflowRAGEvaluationListQuery) ([]WorkflowRAGEvaluationDatasetResource, error) {
	r.calls.Add(1)
	return nil, nil
}
func (r *countingBatchEDatasetRepository) ReadLatest(WorkflowRAGSnapshotContext, string) (WorkflowRAGEvaluationDatasetResource, WorkflowRAGEvaluationDatasetVersion, error) {
	r.calls.Add(1)
	return WorkflowRAGEvaluationDatasetResource{}, WorkflowRAGEvaluationDatasetVersion{}, nil
}
func (r *countingBatchEDatasetRepository) ReadVersion(WorkflowRAGSnapshotContext, string, int) (WorkflowRAGEvaluationDatasetResource, WorkflowRAGEvaluationDatasetVersion, error) {
	r.calls.Add(1)
	return WorkflowRAGEvaluationDatasetResource{}, WorkflowRAGEvaluationDatasetVersion{}, nil
}
func (r *countingBatchEDatasetRepository) CreateVersion(WorkflowRAGSnapshotContext, string, int, WorkflowRAGEvaluationDatasetResource, WorkflowRAGEvaluationDatasetVersion, WorkflowRAGEvaluationAudit) error {
	r.calls.Add(1)
	return nil
}
func (r *countingBatchEDatasetRepository) Archive(WorkflowRAGSnapshotContext, string, int, WorkflowRAGEvaluationDatasetResource, WorkflowRAGEvaluationAudit) error {
	r.calls.Add(1)
	return nil
}
func (r *countingBatchEDatasetRepository) CreateReview(WorkflowRAGSnapshotContext, WorkflowRAGCandidateSnapshotReview, WorkflowRAGEvaluationAudit) error {
	r.calls.Add(1)
	return nil
}
func (r *countingBatchEDatasetRepository) ReadReview(WorkflowRAGSnapshotContext, string, string) (WorkflowRAGCandidateSnapshotReview, error) {
	r.calls.Add(1)
	return WorkflowRAGCandidateSnapshotReview{}, nil
}
func (r *countingBatchEDatasetRepository) ListReviews(WorkflowRAGSnapshotContext, string, workflowRAGCandidateReviewListQuery) ([]WorkflowRAGCandidateSnapshotReview, error) {
	r.calls.Add(1)
	return nil, nil
}

type countingBatchEToolActionStore struct{ calls *atomic.Int64 }

func (s *countingBatchEToolActionStore) CreatePlan(WorkflowHTTPToolActionContext, *WorkflowHTTPToolActionPlan, WorkflowHTTPToolExecutionAudit) error {
	s.calls.Add(1)
	return nil
}
func (s *countingBatchEToolActionStore) ReadPlan(WorkflowHTTPToolActionContext, string) (WorkflowHTTPToolActionPlan, bool, error) {
	s.calls.Add(1)
	return WorkflowHTTPToolActionPlan{}, false, nil
}
func (s *countingBatchEToolActionStore) DecidePlan(WorkflowHTTPToolActionContext, *WorkflowHTTPToolActionPlan, WorkflowHTTPToolConfirmationDecision, WorkflowHTTPToolExecutionAudit) error {
	s.calls.Add(1)
	return nil
}

type countingBatchEEvaluationStore struct{ calls *atomic.Int64 }

func (s *countingBatchEEvaluationStore) CreateCase(WorkflowRunContext, WorkflowEvaluationCase) error {
	s.calls.Add(1)
	return nil
}
func (s *countingBatchEEvaluationStore) ReviseCase(WorkflowRunContext, int, WorkflowEvaluationCase) (WorkflowEvaluationCase, bool, error) {
	s.calls.Add(1)
	return WorkflowEvaluationCase{}, false, nil
}
func (s *countingBatchEEvaluationStore) ReadCase(WorkflowRunContext, string) (WorkflowEvaluationCase, bool, error) {
	s.calls.Add(1)
	return WorkflowEvaluationCase{}, false, nil
}
func (s *countingBatchEEvaluationStore) ReadRevision(WorkflowRunContext, string, int) (WorkflowEvaluationCase, bool, error) {
	s.calls.Add(1)
	return WorkflowEvaluationCase{}, false, nil
}
func (s *countingBatchEEvaluationStore) ListCases(WorkflowRunContext, WorkflowEvaluationListFilter) (WorkflowEvaluationListPage, error) {
	s.calls.Add(1)
	return WorkflowEvaluationListPage{}, nil
}
func (s *countingBatchEEvaluationStore) ListRevisions(WorkflowRunContext, string, WorkflowEvaluationRevisionListFilter) (WorkflowEvaluationRevisionListPage, error) {
	s.calls.Add(1)
	return WorkflowEvaluationRevisionListPage{}, nil
}

type countingBatchEEvaluationSuiteStore struct{ calls *atomic.Int64 }

func (s *countingBatchEEvaluationSuiteStore) CreateSuite(WorkflowRunContext, WorkflowEvaluationSuite) error {
	s.calls.Add(1)
	return nil
}
func (s *countingBatchEEvaluationSuiteStore) ReadSuite(WorkflowRunContext, string) (WorkflowEvaluationSuite, bool, error) {
	s.calls.Add(1)
	return WorkflowEvaluationSuite{}, false, nil
}
func (s *countingBatchEEvaluationSuiteStore) ListSuites(WorkflowRunContext, workflowEvaluationSuiteListFilter) (workflowEvaluationSuiteListPage, error) {
	s.calls.Add(1)
	return workflowEvaluationSuiteListPage{}, nil
}
func (s *countingBatchEEvaluationSuiteStore) AppendDecision(WorkflowRunContext, int, WorkflowEvaluationReleaseDecision) (WorkflowEvaluationSuite, bool, error) {
	s.calls.Add(1)
	return WorkflowEvaluationSuite{}, false, nil
}
func (s *countingBatchEEvaluationSuiteStore) ListDecisions(WorkflowRunContext, string, workflowEvaluationDecisionListFilter) (workflowEvaluationDecisionListPage, error) {
	s.calls.Add(1)
	return workflowEvaluationDecisionListPage{}, nil
}
