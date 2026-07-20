package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"radishmind.local/services/platform/internal/bridge"
)

func TestSQLiteDevWorkflowHTTPToolExecutionHTTPChainSurvivesRestart(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "radishmind.db")
	cfg := aggregateSQLiteDevServerConfig(databasePath)
	cfg.WorkflowToolActionDevEnabled = true
	cfg.WorkflowHTTPToolExecutionDevEnabled = true
	cfg.Model = "mock"

	firstServer, err := NewServerWithError(cfg, Options{BuildVersion: "sqlite-workflow-http-tool-first"})
	if err != nil {
		t.Fatalf("start first SQLite Workflow HTTP Tool server: %v", err)
	}
	firstServer.bridge = &workflowExecutorTestBridge{handle: func(_ context.Context, _ []byte, _ bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
		return successfulWorkflowExecutorEnvelope("reviewable workflow answer"), nil
	}}
	draft := workflowHTTPToolEligibleDraftForTest()
	draftPayload := savedWorkflowDraftPayloadFromDraft(draft)
	saveRequest := httptest.NewRequest(
		http.MethodPost,
		"/v1/user-workspace/workflow-drafts",
		bytes.NewReader(mustSavedWorkflowDraftJSON(t, savedWorkflowDraftSaveHTTPBody{
			Draft: savedWorkflowDraftPayloadDocumentFromDraftPayload(draftPayload),
		})),
	)
	setLocalProductWorkflowHeaders(saveRequest, "workflow_drafts:read,workflow_drafts:write", draftPayload.ApplicationID)
	saveResponse := httptest.NewRecorder()
	firstServer.httpServer.Handler.ServeHTTP(saveResponse, saveRequest)
	saved := decodeSavedWorkflowDraftEnvelope(t, saveResponse, http.StatusOK)
	if saved.FailureCode != nil || saved.Draft == nil {
		firstServer.Close()
		t.Fatalf("save exact Workflow HTTP Tool draft over HTTP: %#v", saved)
	}

	plan := createWorkflowHTTPToolActionPlanOverHTTP(t, firstServer, draft)
	approved := approveWorkflowHTTPToolActionPlanOverHTTP(t, firstServer, draft, plan)
	bridgeCallsBeforeExecution := firstServer.bridge.(*workflowExecutorTestBridge).callCount()
	networkAttempts := 0
	transport := workflowHTTPToolTestTransport(func(*http.Request) (*http.Response, error) {
		networkAttempts++
		return workflowHTTPToolJSONResponse(http.StatusOK, `{"resource_key":"docs/radishflow/overview","title":"RadishFlow","summary":"Reviewed resource","updated_at":"2026-07-17T02:00:00Z"}`), nil
	})
	firstServer.workflowHTTPToolExecutionTransport = &transport
	rawInput := "private SQLite Workflow HTTP Tool input must not persist"
	executeRequest := workflowHTTPToolSQLiteExecutionRequest(t, approved, rawInput)
	executeResponse := httptest.NewRecorder()
	firstServer.httpServer.Handler.ServeHTTP(executeResponse, executeRequest)
	executed := decodeWorkflowHTTPToolExecutionEnvelope(t, executeResponse, http.StatusOK)
	if executed.FailureCode != nil || executed.ActionPlan == nil || executed.Run == nil ||
		executed.ActionPlan.Status != WorkflowHTTPToolActionStatusConsumed ||
		executed.Run.SchemaVersion != workflowRunRecordToolSchemaVersion ||
		executed.Run.Status != WorkflowRunStatusSucceeded || executed.Run.RecordVersion != 2 ||
		executed.Run.SideEffects.ToolCalls != 1 || executed.Run.SideEffects.ConfirmationCalls != 1 ||
		executed.Run.SideEffects.BusinessWrites != 0 || executed.Run.SideEffects.ReplayWrites != 0 ||
		networkAttempts != 1 || firstServer.bridge.(*workflowExecutorTestBridge).callCount() != bridgeCallsBeforeExecution+1 {
		firstServer.Close()
		t.Fatalf("execute SQLite Workflow HTTP Tool exactly once: %#v network=%d", executed, networkAttempts)
	}
	runID := executed.Run.RunID
	firstServer.Close()

	restartedServer, err := NewServerWithError(cfg, Options{BuildVersion: "sqlite-workflow-http-tool-restarted"})
	if err != nil {
		t.Fatalf("restart SQLite Workflow HTTP Tool server: %v", err)
	}
	t.Cleanup(restartedServer.Close)
	restartedBridge := &workflowExecutorTestBridge{}
	restartedServer.bridge = restartedBridge
	restartedNetworkAttempts := 0
	restartedTransport := workflowHTTPToolTestTransport(func(*http.Request) (*http.Response, error) {
		restartedNetworkAttempts++
		return workflowHTTPToolJSONResponse(http.StatusOK, `{}`), nil
	})
	restartedServer.workflowHTTPToolExecutionTransport = &restartedTransport

	readPlanRequest := httptest.NewRequest(
		http.MethodGet,
		"/v1/user-workspace/workflow-tool-action-plans/"+approved.PlanID+
			"?workspace_id="+url.QueryEscape(approved.WorkspaceID)+"&application_id="+url.QueryEscape(approved.ApplicationID),
		nil,
	)
	setWorkflowHTTPToolActionDevHeaders(readPlanRequest, "workflow_tool_actions:read")
	readPlanResponse := httptest.NewRecorder()
	restartedServer.httpServer.Handler.ServeHTTP(readPlanResponse, readPlanRequest)
	restoredPlan := decodeWorkflowHTTPToolActionEnvelope(t, readPlanResponse, http.StatusOK)
	if restoredPlan.FailureCode != nil || restoredPlan.ActionPlan == nil ||
		restoredPlan.ActionPlan.Status != WorkflowHTTPToolActionStatusConsumed || restoredPlan.ActionPlan.RecordVersion != 3 {
		t.Fatalf("restore consumed action plan after restart: %#v", restoredPlan)
	}

	readRunRequest := httptest.NewRequest(
		http.MethodGet,
		"/v1/user-workspace/workflow-runs/"+runID+"?workspace_id=workspace_demo&application_id=app_flow_copilot",
		nil,
	)
	setLocalProductWorkflowHeaders(readRunRequest, "workflow_runs:read", approved.ApplicationID)
	readRunResponse := httptest.NewRecorder()
	restartedServer.httpServer.Handler.ServeHTTP(readRunResponse, readRunRequest)
	restoredRun := decodeWorkflowRunEnvelope(t, readRunResponse, http.StatusOK)
	if restoredRun.FailureCode != nil || restoredRun.Run == nil ||
		restoredRun.Run.SchemaVersion != workflowRunRecordToolSchemaVersion || restoredRun.Run.Status != WorkflowRunStatusSucceeded ||
		restoredRun.Run.PlanID != approved.PlanID || restoredRun.Run.ToolAttempt == nil ||
		restoredRun.Run.ToolAttempt.Status != WorkflowHTTPToolAttemptSucceeded {
		t.Fatalf("restore Workflow HTTP Tool run v2 after restart: %#v", restoredRun)
	}

	historyRequest := httptest.NewRequest(
		http.MethodGet,
		"/v1/user-workspace/workflow-runs?workspace_id=workspace_demo&application_id=app_flow_copilot&status=succeeded",
		nil,
	)
	setLocalProductWorkflowHeaders(historyRequest, "workflow_runs:read", approved.ApplicationID)
	historyResponse := httptest.NewRecorder()
	restartedServer.httpServer.Handler.ServeHTTP(historyResponse, historyRequest)
	history := decodeWorkflowRunListEnvelope(t, historyResponse, http.StatusOK)
	if history.FailureCode != nil || len(history.Runs) != 1 || history.Runs[0].RunID != runID ||
		history.Runs[0].PlanID != approved.PlanID || history.Runs[0].ConfirmationID == "" ||
		history.Runs[0].ToolAttemptStatus != WorkflowHTTPToolAttemptSucceeded ||
		history.Runs[0].SideEffects.ToolCalls != 1 || history.Runs[0].SideEffects.ConfirmationCalls != 1 {
		t.Fatalf("restore Workflow HTTP Tool run v2 history after restart: %#v", history)
	}

	repeatedResponse := httptest.NewRecorder()
	restartedServer.httpServer.Handler.ServeHTTP(repeatedResponse, workflowHTTPToolSQLiteExecutionRequest(t, approved, "do not repeat"))
	repeated := decodeWorkflowHTTPToolExecutionEnvelope(t, repeatedResponse, http.StatusOK)
	if repeated.FailureCode == nil || *repeated.FailureCode != string(WorkflowRunFailureToolConfirmation) ||
		repeated.Run != nil || restartedNetworkAttempts != 0 || restartedBridge.callCount() != 0 {
		t.Fatalf("restart allowed repeated Workflow HTTP Tool execution: %#v network=%d bridge=%d", repeated, restartedNetworkAttempts, restartedBridge.callCount())
	}

	restartedServer.Close()
	assertLocalProductSQLiteFilesExclude(t, databasePath, rawInput, "api.dev.example.invalid", "Authorization", "raw_response")
}

func workflowHTTPToolSQLiteExecutionRequest(
	t *testing.T,
	plan WorkflowHTTPToolActionPlan,
	inputText string,
) *http.Request {
	t.Helper()
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/user-workspace/workflow-tool-action-plans/"+plan.PlanID+"/executions",
		bytes.NewReader(mustWorkflowHTTPToolActionJSON(t, workflowHTTPToolExecutionBody{
			WorkspaceID: plan.WorkspaceID, ApplicationID: plan.ApplicationID,
			ExpectedRecordVersion: plan.RecordVersion, InputText: inputText, Model: "mock",
		})),
	)
	setWorkflowHTTPToolActionDevHeaders(request, strings.Join(workflowHTTPToolExecutionRequiredScopes, ","))
	return request
}
