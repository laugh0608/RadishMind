package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"radishmind.local/services/platform/internal/bridge"
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

	t.Run("execution route consumes one approved plan and rejects repeated execution", func(t *testing.T) {
		server, testBridge, draft := newWorkflowHTTPToolActionHTTPTestServer(t)
		plan := createWorkflowHTTPToolActionPlanOverHTTP(t, server, draft)
		plan = approveWorkflowHTTPToolActionPlanOverHTTP(t, server, draft, plan)
		networkAttempts := 0
		transport := workflowHTTPToolTestTransport(func(*http.Request) (*http.Response, error) {
			networkAttempts++
			return workflowHTTPToolJSONResponse(http.StatusOK, `{"resource_key":"docs/radishflow/overview","title":"RadishFlow","summary":"Reviewed resource","updated_at":"2026-07-17T02:00:00Z"}`), nil
		})
		server.workflowHTTPToolExecutionTransport = &transport
		request := httptest.NewRequest(
			http.MethodPost,
			"/v1/user-workspace/workflow-tool-action-plans/"+plan.PlanID+"/executions",
			strings.NewReader(`{"workspace_id":"workspace_demo","application_id":"app_flow_copilot","expected_record_version":2,"input_text":"Review the approved resource.","model":"mock","temperature":0}`),
		)
		setWorkflowHTTPToolActionDevHeaders(request, "workflow_tool_actions:execute,workflow_runs:execute,workflow_drafts:read")
		response := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(response, request)

		executed := decodeWorkflowHTTPToolExecutionEnvelope(t, response, http.StatusOK)
		if executed.FailureCode != nil || executed.ActionPlan == nil || executed.Run == nil ||
			executed.ActionPlan.Status != WorkflowHTTPToolActionStatusConsumed ||
			executed.Run.SchemaVersion != workflowRunRecordToolSchemaVersion ||
			executed.Run.Status != WorkflowRunStatusSucceeded ||
			executed.Run.SideEffects.ToolCalls != 1 || executed.Run.SideEffects.ConfirmationCalls != 1 ||
			executed.Run.SideEffects.BusinessWrites != 0 || executed.Run.SideEffects.ReplayWrites != 0 ||
			networkAttempts != 1 || testBridge.callCount() != 1 {
			t.Fatalf("approved plan did not execute exactly once: %#v network=%d bridge=%d", executed, networkAttempts, testBridge.callCount())
		}

		repeatedResponse := httptest.NewRecorder()
		repeatedRequest := httptest.NewRequest(http.MethodPost, request.URL.String(), strings.NewReader(`{"workspace_id":"workspace_demo","application_id":"app_flow_copilot","expected_record_version":2,"input_text":"Do not repeat.","model":"mock","temperature":0}`))
		setWorkflowHTTPToolActionDevHeaders(repeatedRequest, "workflow_tool_actions:execute,workflow_runs:execute,workflow_drafts:read")
		server.httpServer.Handler.ServeHTTP(repeatedResponse, repeatedRequest)
		repeated := decodeWorkflowHTTPToolExecutionEnvelope(t, repeatedResponse, http.StatusOK)
		if repeated.FailureCode == nil || *repeated.FailureCode != string(WorkflowRunFailureToolConfirmation) ||
			repeated.Run != nil || networkAttempts != 1 || testBridge.callCount() != 1 {
			t.Fatalf("repeated execution did not fail before side effects: %#v network=%d bridge=%d", repeated, networkAttempts, testBridge.callCount())
		}
	})

	t.Run("execution route requires all three scopes and strict JSON", func(t *testing.T) {
		server, testBridge, draft := newWorkflowHTTPToolActionHTTPTestServer(t)
		plan := approveWorkflowHTTPToolActionPlanOverHTTP(t, server, draft, createWorkflowHTTPToolActionPlanOverHTTP(t, server, draft))
		networkAttempts := 0
		transport := workflowHTTPToolTestTransport(func(*http.Request) (*http.Response, error) {
			networkAttempts++
			return workflowHTTPToolJSONResponse(http.StatusOK, `{}`), nil
		})
		server.workflowHTTPToolExecutionTransport = &transport

		missingScopeRequest := httptest.NewRequest(
			http.MethodPost,
			"/v1/user-workspace/workflow-tool-action-plans/"+plan.PlanID+"/executions",
			strings.NewReader(`{"workspace_id":"workspace_demo","application_id":"app_flow_copilot","expected_record_version":2,"input_text":"Review.","model":"mock","temperature":0}`),
		)
		setWorkflowHTTPToolActionDevHeaders(missingScopeRequest, "workflow_tool_actions:execute,workflow_drafts:read")
		missingScopeResponse := httptest.NewRecorder()
		server.httpServer.Handler.ServeHTTP(missingScopeResponse, missingScopeRequest)
		missingScope := decodeWorkflowHTTPToolExecutionEnvelope(t, missingScopeResponse, http.StatusOK)
		if missingScope.FailureCode == nil || *missingScope.FailureCode != string(WorkflowRunFailureScopeDenied) {
			t.Fatalf("missing workflow_runs:execute scope was accepted: %#v", missingScope)
		}

		unknownFieldRequest := httptest.NewRequest(
			http.MethodPost,
			"/v1/user-workspace/workflow-tool-action-plans/"+plan.PlanID+"/executions",
			strings.NewReader(`{"workspace_id":"workspace_demo","application_id":"app_flow_copilot","expected_record_version":2,"input_text":"Review.","model":"mock","temperature":0,"retry":true}`),
		)
		setWorkflowHTTPToolActionDevHeaders(unknownFieldRequest, "workflow_tool_actions:execute,workflow_runs:execute,workflow_drafts:read")
		unknownFieldResponse := httptest.NewRecorder()
		server.httpServer.Handler.ServeHTTP(unknownFieldResponse, unknownFieldRequest)
		assertWorkflowHTTPToolActionStrictJSONError(t, unknownFieldResponse)
		if networkAttempts != 0 || testBridge.callCount() != 0 {
			t.Fatalf("pre-claim HTTP failures crossed execution boundary: network=%d bridge=%d", networkAttempts, testBridge.callCount())
		}
	})
}

func newWorkflowHTTPToolActionHTTPTestServer(t *testing.T) (*Server, *workflowExecutorTestBridge, SavedWorkflowDraft) {
	t.Helper()
	server := NewServer(config.Config{
		ControlPlaneReadDevAuthEnabled:      true,
		WorkflowSavedDraftDevHTTPEnabled:    true,
		WorkflowSavedDraftDevWriteEnabled:   true,
		WorkflowExecutorDevEnabled:          true,
		WorkflowToolActionDevEnabled:        true,
		WorkflowHTTPToolExecutionDevEnabled: true,
		Provider:                            "mock",
	}, Options{BuildVersion: "test"})
	t.Cleanup(server.Close)
	testBridge := &workflowExecutorTestBridge{handle: func(context.Context, []byte, bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
		return successfulWorkflowExecutorEnvelope("reviewable workflow answer"), nil
	}}
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

func approveWorkflowHTTPToolActionPlanOverHTTP(
	t *testing.T,
	server *Server,
	draft SavedWorkflowDraft,
	plan WorkflowHTTPToolActionPlan,
) WorkflowHTTPToolActionPlan {
	t.Helper()
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/user-workspace/workflow-tool-action-plans/"+plan.PlanID+"/decisions",
		bytes.NewReader(mustWorkflowHTTPToolActionJSON(t, workflowHTTPToolDecisionBody{
			WorkspaceID: draft.WorkspaceID, ApplicationID: draft.ApplicationID,
			ExpectedRecordVersion: plan.RecordVersion, Decision: WorkflowHTTPToolConfirmationApprove,
		})),
	)
	setWorkflowHTTPToolActionDevHeaders(request, "workflow_tool_actions:confirm")
	response := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(response, request)
	envelope := decodeWorkflowHTTPToolActionEnvelope(t, response, http.StatusOK)
	if envelope.FailureCode != nil || envelope.ActionPlan == nil || envelope.ActionPlan.Status != WorkflowHTTPToolActionStatusApproved {
		t.Fatalf("approve workflow HTTP tool action plan: %#v", envelope)
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

func decodeWorkflowHTTPToolExecutionEnvelope(
	t *testing.T,
	response *httptest.ResponseRecorder,
	expectedStatus int,
) workflowHTTPToolExecutionEnvelope {
	t.Helper()
	if response.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d: %s", expectedStatus, response.Code, response.Body.String())
	}
	var envelope workflowHTTPToolExecutionEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode workflow HTTP tool execution envelope: %v\n%s", err, response.Body.String())
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
