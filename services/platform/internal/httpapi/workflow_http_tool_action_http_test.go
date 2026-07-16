package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"radishmind.local/services/platform/internal/config"
)

func TestWorkflowHTTPToolActionHTTPRoutes(t *testing.T) {
	t.Run("create read and approve remain pre-run only", func(t *testing.T) {
		server, testBridge, draft := newWorkflowHTTPToolActionHTTPTestServer(t)
		createRequest := httptest.NewRequest(
			http.MethodPost,
			"/v1/user-workspace/workflow-drafts/"+draft.DraftID+"/tool-action-plans",
			bytes.NewReader(mustWorkflowHTTPToolActionJSON(t, workflowHTTPToolCreatePlanBody{
				WorkspaceID: draft.WorkspaceID, ApplicationID: draft.ApplicationID,
				DraftVersion: draft.DraftVersion, NodeID: "node_http_tool",
				PublicArguments: map[string]any{"resource_key": "docs/radishflow/overview", "locale": "zh-CN"},
			})),
		)
		setWorkflowHTTPToolActionDevHeaders(createRequest, "workflow_drafts:read,workflow_tool_actions:plan")
		createResponse := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(createResponse, createRequest)

		created := decodeWorkflowHTTPToolActionEnvelope(t, createResponse, http.StatusOK)
		if created.FailureCode != nil || created.ActionPlan == nil || created.ConfirmationDecision != nil {
			t.Fatalf("create plan should succeed without a decision: %#v", created)
		}
		if created.ActionPlan.DraftID != draft.DraftID || created.ActionPlan.DraftVersion != draft.DraftVersion ||
			created.ActionPlan.NodeID != "node_http_tool" || created.ActionPlan.Status != WorkflowHTTPToolActionStatusPending ||
			created.ActionPlan.RecordVersion != 1 {
			t.Fatalf("create response drifted from the exact saved draft: %#v", created.ActionPlan)
		}

		readRequest := httptest.NewRequest(
			http.MethodGet,
			"/v1/user-workspace/workflow-tool-action-plans/"+created.ActionPlan.PlanID+
				"?workspace_id="+draft.WorkspaceID+"&application_id="+draft.ApplicationID,
			nil,
		)
		setWorkflowHTTPToolActionDevHeaders(readRequest, "workflow_tool_actions:read")
		readResponse := httptest.NewRecorder()
		server.httpServer.Handler.ServeHTTP(readResponse, readRequest)
		read := decodeWorkflowHTTPToolActionEnvelope(t, readResponse, http.StatusOK)
		if read.FailureCode != nil || read.ActionPlan == nil || read.ActionPlan.PlanID != created.ActionPlan.PlanID ||
			read.ActionPlan.RecordVersion != 1 {
			t.Fatalf("scoped plan read should succeed: %#v", read)
		}

		decisionRequest := httptest.NewRequest(
			http.MethodPost,
			"/v1/user-workspace/workflow-tool-action-plans/"+created.ActionPlan.PlanID+"/decisions",
			bytes.NewReader(mustWorkflowHTTPToolActionJSON(t, workflowHTTPToolDecisionBody{
				WorkspaceID: draft.WorkspaceID, ApplicationID: draft.ApplicationID,
				ExpectedRecordVersion: 1, Decision: WorkflowHTTPToolConfirmationApprove,
			})),
		)
		setWorkflowHTTPToolActionDevHeaders(decisionRequest, "workflow_tool_actions:confirm")
		decisionResponse := httptest.NewRecorder()
		server.httpServer.Handler.ServeHTTP(decisionResponse, decisionRequest)
		approved := decodeWorkflowHTTPToolActionEnvelope(t, decisionResponse, http.StatusOK)
		if approved.FailureCode != nil || approved.ActionPlan == nil || approved.ConfirmationDecision == nil ||
			approved.ActionPlan.Status != WorkflowHTTPToolActionStatusApproved || approved.ActionPlan.RecordVersion != 2 ||
			approved.ConfirmationDecision.Outcome != WorkflowHTTPToolConfirmationApprove {
			t.Fatalf("approval should be persisted as a separate CAS decision: %#v", approved)
		}

		assertWorkflowHTTPToolBatchAHasNoExecutionSideEffects(t, server, testBridge)
	})

	t.Run("each route requires its dedicated scope", func(t *testing.T) {
		server, testBridge, draft := newWorkflowHTTPToolActionHTTPTestServer(t)
		plan := createWorkflowHTTPToolActionPlanOverHTTP(t, server, draft)
		tests := []struct {
			name   string
			method string
			path   string
			body   string
			scopes string
		}{
			{
				name: "create requires plan scope", method: http.MethodPost,
				path: "/v1/user-workspace/workflow-drafts/" + draft.DraftID + "/tool-action-plans",
				body: string(mustWorkflowHTTPToolActionJSON(t, workflowHTTPToolCreatePlanBody{
					WorkspaceID: draft.WorkspaceID, ApplicationID: draft.ApplicationID,
					DraftVersion: draft.DraftVersion, NodeID: "node_http_tool",
					PublicArguments: map[string]any{"resource_key": "docs/radishflow/overview"},
				})),
				scopes: "workflow_drafts:read",
			},
			{
				name: "read requires read scope", method: http.MethodGet,
				path:   "/v1/user-workspace/workflow-tool-action-plans/" + plan.PlanID + "?workspace_id=" + draft.WorkspaceID + "&application_id=" + draft.ApplicationID,
				scopes: "workflow_tool_actions:plan",
			},
			{
				name: "decision requires confirm scope", method: http.MethodPost,
				path: "/v1/user-workspace/workflow-tool-action-plans/" + plan.PlanID + "/decisions",
				body: string(mustWorkflowHTTPToolActionJSON(t, workflowHTTPToolDecisionBody{
					WorkspaceID: draft.WorkspaceID, ApplicationID: draft.ApplicationID,
					ExpectedRecordVersion: 1, Decision: WorkflowHTTPToolConfirmationApprove,
				})),
				scopes: "workflow_tool_actions:read",
			},
		}
		for _, testCase := range tests {
			t.Run(testCase.name, func(t *testing.T) {
				request := httptest.NewRequest(testCase.method, testCase.path, strings.NewReader(testCase.body))
				setWorkflowHTTPToolActionDevHeaders(request, testCase.scopes)
				response := httptest.NewRecorder()
				server.httpServer.Handler.ServeHTTP(response, request)
				envelope := decodeWorkflowHTTPToolActionEnvelope(t, response, http.StatusOK)
				if envelope.FailureCode == nil || *envelope.FailureCode != string(WorkflowHTTPToolActionFailureScopeDenied) || envelope.ActionPlan != nil {
					t.Fatalf("missing dedicated scope must fail closed: %#v", envelope)
				}
			})
		}
		assertWorkflowHTTPToolBatchAHasNoExecutionSideEffects(t, server, testBridge)
	})

	t.Run("request bodies and read query reject unknown fields", func(t *testing.T) {
		server, testBridge, draft := newWorkflowHTTPToolActionHTTPTestServer(t)
		plan := createWorkflowHTTPToolActionPlanOverHTTP(t, server, draft)

		createRequest := httptest.NewRequest(
			http.MethodPost,
			"/v1/user-workspace/workflow-drafts/"+draft.DraftID+"/tool-action-plans",
			strings.NewReader(`{"workspace_id":"workspace_demo","application_id":"app_flow_copilot","draft_version":1,"node_id":"node_http_tool","public_arguments":{"resource_key":"docs/radishflow/overview"},"tool_id":"client-controlled"}`),
		)
		setWorkflowHTTPToolActionDevHeaders(createRequest, "workflow_drafts:read,workflow_tool_actions:plan")
		createResponse := httptest.NewRecorder()
		server.httpServer.Handler.ServeHTTP(createResponse, createRequest)
		assertWorkflowHTTPToolActionStrictJSONError(t, createResponse)

		decisionRequest := httptest.NewRequest(
			http.MethodPost,
			"/v1/user-workspace/workflow-tool-action-plans/"+plan.PlanID+"/decisions",
			strings.NewReader(`{"workspace_id":"workspace_demo","application_id":"app_flow_copilot","expected_record_version":1,"decision":"approve","execute":true}`),
		)
		setWorkflowHTTPToolActionDevHeaders(decisionRequest, "workflow_tool_actions:confirm")
		decisionResponse := httptest.NewRecorder()
		server.httpServer.Handler.ServeHTTP(decisionResponse, decisionRequest)
		assertWorkflowHTTPToolActionStrictJSONError(t, decisionResponse)

		readRequest := httptest.NewRequest(
			http.MethodGet,
			"/v1/user-workspace/workflow-tool-action-plans/"+plan.PlanID+
				"?workspace_id="+draft.WorkspaceID+"&application_id="+draft.ApplicationID+"&include_execution=true",
			nil,
		)
		setWorkflowHTTPToolActionDevHeaders(readRequest, "workflow_tool_actions:read")
		readResponse := httptest.NewRecorder()
		server.httpServer.Handler.ServeHTTP(readResponse, readRequest)
		readEnvelope := decodeWorkflowHTTPToolActionEnvelope(t, readResponse, http.StatusOK)
		if readEnvelope.FailureCode == nil || *readEnvelope.FailureCode != string(WorkflowHTTPToolActionFailureInputInvalid) || readEnvelope.ActionPlan != nil {
			t.Fatalf("unknown read query must fail closed: %#v", readEnvelope)
		}
		assertWorkflowHTTPToolBatchAHasNoExecutionSideEffects(t, server, testBridge)
	})

	t.Run("execution route is not registered in batch A", func(t *testing.T) {
		server, testBridge, draft := newWorkflowHTTPToolActionHTTPTestServer(t)
		plan := createWorkflowHTTPToolActionPlanOverHTTP(t, server, draft)
		request := httptest.NewRequest(
			http.MethodPost,
			"/v1/user-workspace/workflow-tool-action-plans/"+plan.PlanID+"/executions",
			strings.NewReader(`{"workspace_id":"workspace_demo","application_id":"app_flow_copilot","expected_record_version":1}`),
		)
		setWorkflowHTTPToolActionDevHeaders(request, "workflow_tool_actions:execute,workflow_runs:execute,workflow_drafts:read")
		response := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(response, request)

		if response.Code != http.StatusNotFound {
			t.Fatalf("batch A must not register /executions, got %d: %s", response.Code, response.Body.String())
		}
		assertWorkflowHTTPToolBatchAHasNoExecutionSideEffects(t, server, testBridge)
	})
}

func newWorkflowHTTPToolActionHTTPTestServer(t *testing.T) (*Server, *workflowExecutorTestBridge, SavedWorkflowDraft) {
	t.Helper()
	server := NewServer(config.Config{
		ControlPlaneReadDevAuthEnabled:    true,
		WorkflowSavedDraftDevHTTPEnabled:  true,
		WorkflowSavedDraftDevWriteEnabled: true,
		WorkflowExecutorDevEnabled:        true,
		WorkflowToolActionDevEnabled:      true,
		Provider:                          "mock",
	}, Options{BuildVersion: "test"})
	t.Cleanup(server.Close)
	testBridge := &workflowExecutorTestBridge{}
	server.bridge = testBridge
	payload := savedWorkflowDraftPayloadFromDraft(workflowHTTPToolEligibleDraftForTest())
	result := server.savedWorkflowDraftService().SaveDraft(
		savedWorkflowDraftTestContext(),
		SaveWorkflowDraftRequest{ExpectedDraftVersion: 0, Payload: payload},
	)
	if result.FailureCode != "" || result.Draft == nil {
		t.Fatalf("prepare exact saved draft for workflow HTTP tool action test: %#v", result)
	}
	return server, testBridge, *result.Draft
}

func createWorkflowHTTPToolActionPlanOverHTTP(t *testing.T, server *Server, draft SavedWorkflowDraft) WorkflowHTTPToolActionPlan {
	t.Helper()
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/user-workspace/workflow-drafts/"+draft.DraftID+"/tool-action-plans",
		bytes.NewReader(mustWorkflowHTTPToolActionJSON(t, workflowHTTPToolCreatePlanBody{
			WorkspaceID: draft.WorkspaceID, ApplicationID: draft.ApplicationID,
			DraftVersion: draft.DraftVersion, NodeID: "node_http_tool",
			PublicArguments: map[string]any{"resource_key": "docs/radishflow/overview"},
		})),
	)
	setWorkflowHTTPToolActionDevHeaders(request, "workflow_drafts:read,workflow_tool_actions:plan")
	response := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(response, request)
	envelope := decodeWorkflowHTTPToolActionEnvelope(t, response, http.StatusOK)
	if envelope.FailureCode != nil || envelope.ActionPlan == nil {
		t.Fatalf("prepare workflow HTTP tool action plan: %#v", envelope)
	}
	return *envelope.ActionPlan
}

func setWorkflowHTTPToolActionDevHeaders(request *http.Request, scopes string) {
	setSavedWorkflowDraftDevHeaders(request, scopes)
}

func mustWorkflowHTTPToolActionJSON(t *testing.T, document any) []byte {
	t.Helper()
	payload, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("marshal workflow HTTP tool action document: %v", err)
	}
	return payload
}

func decodeWorkflowHTTPToolActionEnvelope(
	t *testing.T,
	response *httptest.ResponseRecorder,
	expectedStatus int,
) workflowHTTPToolActionEnvelope {
	t.Helper()
	if response.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d: %s", expectedStatus, response.Code, response.Body.String())
	}
	var envelope workflowHTTPToolActionEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode workflow HTTP tool action envelope: %v\n%s", err, response.Body.String())
	}
	return envelope
}

func assertWorkflowHTTPToolActionStrictJSONError(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	if response.Code != http.StatusBadRequest {
		t.Fatalf("unknown request field should return 400, got %d: %s", response.Code, response.Body.String())
	}
	var document errorDocument
	if err := json.Unmarshal(response.Body.Bytes(), &document); err != nil {
		t.Fatalf("decode strict JSON error: %v", err)
	}
	if document.Error.Code != "INVALID_JSON" {
		t.Fatalf("unknown request field should fail as INVALID_JSON: %#v", document.Error)
	}
}

func assertWorkflowHTTPToolBatchAHasNoExecutionSideEffects(
	t *testing.T,
	server *Server,
	testBridge *workflowExecutorTestBridge,
) {
	t.Helper()
	if testBridge.callCount() != 0 {
		t.Fatalf("batch A route called Gateway/provider %d times", testBridge.callCount())
	}
	runStore, ok := server.workflowRunStore.(*memoryWorkflowRunStore)
	if !ok {
		t.Fatalf("test server did not select memory workflow run store: %T", server.workflowRunStore)
	}
	actionStore, ok := server.workflowHTTPToolActionStore.(*memoryWorkflowHTTPToolActionStore)
	if !ok {
		t.Fatalf("test server did not select memory action store: %T", server.workflowHTTPToolActionStore)
	}
	if actionStore.ownerLock != &runStore.mu {
		t.Fatalf("action plans and workflow runs must share the same backend owner lock")
	}
	actionStore.ownerLock.RLock()
	defer actionStore.ownerLock.RUnlock()
	if len(runStore.records) != 0 {
		t.Fatalf("batch A route created %d workflow run records", len(runStore.records))
	}
	for _, audit := range actionStore.audits {
		if audit.ExecutionAttemptID != nil || audit.RunID != nil || audit.SideEffects.NetworkAttempts != 0 ||
			audit.SideEffects.ToolCalls != 0 || audit.SideEffects.ProviderCalls != 0 ||
			audit.SideEffects.BusinessWrites != 0 || audit.SideEffects.ReplayWrites != 0 {
			t.Fatalf("batch A HTTP audit claimed an execution side effect: %#v", audit)
		}
	}
	if sideEffects := server.savedWorkflowDraftStore.SideEffects(); hasSavedWorkflowDraftRuntimeSideEffect(sideEffects) {
		t.Fatalf("batch A HTTP flow crossed a saved-draft execution boundary: %#v", sideEffects)
	}
}
