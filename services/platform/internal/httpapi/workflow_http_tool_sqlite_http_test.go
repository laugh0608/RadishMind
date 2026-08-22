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
	"time"

	"radishmind.local/services/platform/internal/bridge"
)

func TestSQLiteDevWorkflowDefinitionHTTPToolPlanSurvivesRestart(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "radishmind.db")
	cfg := aggregateSQLiteDevServerConfig(databasePath)
	cfg.WorkflowToolActionDevEnabled = true
	cfg.WorkflowDefinitionReleaseDevEnabled = true
	cfg.WorkflowHTTPToolExecutionDevEnabled = true
	cfg.Model = "mock"

	firstServer, err := NewServerWithError(cfg, Options{BuildVersion: "sqlite-workflow-definition-http-tool-first"})
	if err != nil {
		t.Fatalf("start first SQLite Workflow Definition HTTP Tool server: %v", err)
	}
	firstBridge := &workflowExecutorTestBridge{}
	firstServer.bridge = firstBridge
	draft := workflowHTTPToolEligibleDraftForTest()
	applicationID := "app_aaaaaaaaaaaaaaaa"
	draft.ApplicationID = applicationID
	applicationService := firstServer.applicationCatalogService()
	applicationService.newID = func() (string, error) { return applicationID, nil }
	applicationContext := ApplicationCatalogContext{
		RequestContext: context.Background(), RequestID: "request_sqlite_definition_tool_app", TenantRef: "tenant_demo",
		WorkspaceID: "workspace_demo", OwnerSubjectRef: "subject_demo_user", ActorRef: "subject_demo_user",
		AuditRef: "audit_sqlite_definition_tool_app", WriteEnabled: true,
	}
	if created := applicationService.Create(applicationContext, ApplicationCatalogCreateInput{DisplayName: "SQLite Definition tool app", ApplicationKind: "workflow_copilot"}); created.FailureCode != "" || created.Record == nil {
		firstServer.Close()
		t.Fatalf("create SQLite Definition tool application: %#v", created)
	}
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
		t.Fatalf("save exact Workflow Definition HTTP Tool draft over HTTP: %#v", saved)
	}

	actionContext := workflowHTTPToolActionTestContext()
	actionContext.ApplicationID = applicationID
	releaseContext := WorkflowDefinitionReleaseContext{
		RequestContext:  actionContext.RequestContext,
		TenantRef:       actionContext.TenantRef,
		WorkspaceID:     actionContext.WorkspaceID,
		ApplicationID:   actionContext.ApplicationID,
		OwnerSubjectRef: actionContext.ActorRef,
		ActorRef:        actionContext.ActorRef,
		RequestID:       actionContext.RequestID,
		AuditRef:        actionContext.AuditRef,
	}
	candidate, err := firstServer.workflowDefinitionReleaseRepository.CreateCandidate(
		releaseContext,
		"candidate-sqlite-http-tool-restart",
		"definition-sqlite-http-tool-restart",
		workflowDefinitionHTTPToolProfile,
		draft,
		time.Date(2026, 8, 15, 11, 0, 0, 0, time.UTC),
	)
	if err != nil {
		firstServer.Close()
		t.Fatalf("create SQLite Workflow Definition HTTP Tool candidate: %v", err)
	}
	_, version, err := firstServer.workflowDefinitionReleaseRepository.Review(
		releaseContext,
		candidate.CandidateID,
		0,
		"approve",
		"reviewed SQLite Workflow Definition HTTP Tool",
		candidate.SourceDraftDigest,
		time.Date(2026, 8, 15, 11, 1, 0, 0, time.UTC),
	)
	if err != nil || version == nil {
		firstServer.Close()
		t.Fatalf("materialize SQLite Workflow Definition HTTP Tool version: version=%#v err=%v", version, err)
	}
	activation, err := firstServer.workflowDefinitionReleaseRepository.DecideActivation(
		releaseContext,
		version.DefinitionID,
		0,
		"activate",
		version.Version,
		"activate SQLite Workflow Definition HTTP Tool",
		time.Date(2026, 8, 15, 11, 2, 0, 0, time.UTC),
	)
	if err != nil {
		firstServer.Close()
		t.Fatalf("activate SQLite Workflow Definition HTTP Tool: %v", err)
	}

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/v1/user-workspace/workflow-definitions/"+version.DefinitionID+"/tool-action-plans",
		bytes.NewReader(mustWorkflowHTTPToolActionJSON(t, workflowDefinitionHTTPToolCreatePlanBody{
			WorkspaceID: saved.Draft.WorkspaceID, ApplicationID: saved.Draft.ApplicationID, NodeID: "node_http_tool",
			PublicArguments: map[string]any{"resource_key": "docs/radishflow/overview"},
		})),
	)
	setWorkflowHTTPToolActionDevHeaders(createRequest, "workflow_definitions:read,workflow_tool_actions:plan")
	createRequest.Header.Set(savedWorkflowDraftDevApplicationHeader, applicationID)
	createResponse := httptest.NewRecorder()
	firstServer.httpServer.Handler.ServeHTTP(createResponse, createRequest)
	created := decodeWorkflowHTTPToolActionEnvelope(t, createResponse, http.StatusOK)
	if created.FailureCode != nil || created.ActionPlan == nil ||
		created.ActionPlan.SchemaVersion != workflowHTTPToolPlanSchemaV2 ||
		created.ActionPlan.SourceKind != workflowHTTPToolSourceDefinition ||
		created.ActionPlan.WorkflowDefinitionID != version.DefinitionID ||
		created.ActionPlan.WorkflowDefinitionVersion != version.Version ||
		created.ActionPlan.WorkflowDefinitionDigest != version.DefinitionDigest ||
		created.ActionPlan.ActivationPointerVersion != activation.PointerVersion ||
		created.ActionPlan.DraftID != "" || created.ActionPlan.DraftVersion != 0 {
		firstServer.Close()
		t.Fatalf("create SQLite Definition-bound action plan: %#v", created)
	}
	approved := approveWorkflowHTTPToolActionPlanOverHTTP(t, firstServer, draft, *created.ActionPlan)
	if firstBridge.callCount() != 0 {
		firstServer.Close()
		t.Fatalf("pre-run SQLite Definition plan crossed Gateway/provider boundary: %d", firstBridge.callCount())
	}
	firstServer.Close()

	restartedServer, err := NewServerWithError(cfg, Options{BuildVersion: "sqlite-workflow-definition-http-tool-restarted"})
	if err != nil {
		t.Fatalf("restart SQLite Workflow Definition HTTP Tool server: %v", err)
	}
	t.Cleanup(restartedServer.Close)
	restartedBridge := &workflowExecutorTestBridge{handle: func(_ context.Context, _ []byte, _ bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
		return successfulWorkflowExecutorEnvelope("reviewable Definition tool answer"), nil
	}}
	restartedServer.bridge = restartedBridge

	restoredVersion, err := restartedServer.workflowDefinitionReleaseRepository.ReadVersion(releaseContext, version.DefinitionID, version.Version)
	if err != nil || restoredVersion.DefinitionDigest != version.DefinitionDigest || restoredVersion.Snapshot.ExecutionProfile != workflowDefinitionHTTPToolProfile {
		t.Fatalf("restore active SQLite Workflow Definition source: version=%#v err=%v", restoredVersion, err)
	}
	restoredActivation, err := restartedServer.workflowDefinitionReleaseRepository.ReadActivation(releaseContext, version.DefinitionID)
	if err != nil || restoredActivation.PointerVersion != activation.PointerVersion || restoredActivation.ActiveVersion != version.Version {
		t.Fatalf("restore SQLite Workflow Definition activation: activation=%#v err=%v", restoredActivation, err)
	}

	readPlanRequest := httptest.NewRequest(
		http.MethodGet,
		"/v1/user-workspace/workflow-tool-action-plans/"+approved.PlanID+
			"?workspace_id="+url.QueryEscape(approved.WorkspaceID)+"&application_id="+url.QueryEscape(approved.ApplicationID),
		nil,
	)
	setWorkflowHTTPToolActionDevHeaders(readPlanRequest, "workflow_tool_actions:read")
	readPlanRequest.Header.Set(savedWorkflowDraftDevApplicationHeader, applicationID)
	readPlanResponse := httptest.NewRecorder()
	restartedServer.httpServer.Handler.ServeHTTP(readPlanResponse, readPlanRequest)
	restored := decodeWorkflowHTTPToolActionEnvelope(t, readPlanResponse, http.StatusOK)
	if restored.FailureCode != nil || restored.ActionPlan == nil ||
		restored.ActionPlan.Status != WorkflowHTTPToolActionStatusApproved || restored.ActionPlan.RecordVersion != 2 ||
		restored.ActionPlan.SchemaVersion != workflowHTTPToolPlanSchemaV2 ||
		restored.ActionPlan.SourceKind != workflowHTTPToolSourceDefinition ||
		restored.ActionPlan.WorkflowDefinitionID != version.DefinitionID ||
		restored.ActionPlan.WorkflowDefinitionVersion != version.Version ||
		restored.ActionPlan.WorkflowDefinitionDigest != version.DefinitionDigest ||
		restored.ActionPlan.ActivationPointerVersion != activation.PointerVersion ||
		restored.ActionPlan.DraftID != "" || restored.ActionPlan.DraftVersion != 0 || restartedBridge.callCount() != 0 {
		t.Fatalf("restore approved SQLite Definition-bound action plan: %#v bridge=%d", restored, restartedBridge.callCount())
	}

	networkAttempts := 0
	transport := workflowHTTPToolTestTransport(func(*http.Request) (*http.Response, error) {
		networkAttempts++
		return workflowHTTPToolJSONResponse(http.StatusOK, `{"resource_key":"docs/radishflow/overview","title":"RadishFlow","summary":"Reviewed resource","updated_at":"2026-08-15T11:03:00Z"}`), nil
	})
	restartedServer.workflowHTTPToolExecutionTransport = &transport
	rawInput := "private SQLite Definition HTTP Tool input must not persist"
	executeRequest := workflowHTTPToolSQLiteExecutionRequest(t, *restored.ActionPlan, rawInput)
	executeResponse := httptest.NewRecorder()
	restartedServer.httpServer.Handler.ServeHTTP(executeResponse, executeRequest)
	executed := decodeWorkflowHTTPToolExecutionEnvelope(t, executeResponse, http.StatusOK)
	if executed.FailureCode != nil || executed.ActionPlan == nil || executed.Run == nil ||
		executed.ActionPlan.Status != WorkflowHTTPToolActionStatusConsumed ||
		executed.Run.SchemaVersion != workflowRunRecordDefinitionToolSchemaVersion ||
		executed.Run.ExecutionSourceKind != workflowDefinitionExecutionSourceKind ||
		executed.Run.ExecutionSourceID != version.DefinitionID || executed.Run.ExecutionSourceVersion != version.Version ||
		executed.Run.DefinitionAuthority == nil || executed.Run.Output != "" || executed.Run.DraftID != "" ||
		executed.Run.SideEffects.ToolCalls != 1 || executed.Run.SideEffects.ConfirmationCalls != 1 ||
		networkAttempts != 1 || restartedBridge.callCount() != 1 {
		t.Fatalf("execute restored SQLite Definition-bound plan exactly once: %#v network=%d bridge=%d", executed, networkAttempts, restartedBridge.callCount())
	}
	runID := executed.Run.RunID
	restartedServer.Close()

	readServer, err := NewServerWithError(cfg, Options{BuildVersion: "sqlite-workflow-definition-http-tool-run-restored"})
	if err != nil {
		t.Fatalf("restart SQLite Definition HTTP Tool run server: %v", err)
	}
	t.Cleanup(readServer.Close)
	readBridge := &workflowExecutorTestBridge{}
	readServer.bridge = readBridge
	readNetworkAttempts := 0
	readTransport := workflowHTTPToolTestTransport(func(*http.Request) (*http.Response, error) {
		readNetworkAttempts++
		return workflowHTTPToolJSONResponse(http.StatusOK, `{}`), nil
	})
	readServer.workflowHTTPToolExecutionTransport = &readTransport

	readRunRequest := httptest.NewRequest(
		http.MethodGet,
		"/v1/user-workspace/workflow-runs/"+runID+"?workspace_id="+url.QueryEscape(draft.WorkspaceID)+"&application_id="+url.QueryEscape(applicationID),
		nil,
	)
	setLocalProductWorkflowHeaders(readRunRequest, "workflow_runs:read", applicationID)
	readRunResponse := httptest.NewRecorder()
	readServer.httpServer.Handler.ServeHTTP(readRunResponse, readRunRequest)
	restoredRun := decodeWorkflowRunEnvelope(t, readRunResponse, http.StatusOK)
	if restoredRun.FailureCode != nil || restoredRun.Run == nil ||
		restoredRun.Run.SchemaVersion != workflowRunRecordDefinitionToolSchemaVersion || restoredRun.Run.RunID != runID ||
		restoredRun.Run.DefinitionAuthority == nil || restoredRun.Run.Output != "" {
		t.Fatalf("restore SQLite Definition HTTP Tool run v9: %#v", restoredRun)
	}
	repeatedResponse := httptest.NewRecorder()
	readServer.httpServer.Handler.ServeHTTP(repeatedResponse, workflowHTTPToolSQLiteExecutionRequest(t, *restored.ActionPlan, "do not repeat"))
	repeated := decodeWorkflowHTTPToolExecutionEnvelope(t, repeatedResponse, http.StatusOK)
	if repeated.FailureCode == nil || *repeated.FailureCode != string(WorkflowRunFailureToolConfirmation) ||
		repeated.Run != nil || readNetworkAttempts != 0 || readBridge.callCount() != 0 {
		t.Fatalf("restart allowed repeated Definition HTTP Tool execution: %#v network=%d bridge=%d", repeated, readNetworkAttempts, readBridge.callCount())
	}
	readServer.Close()
	assertLocalProductSQLiteFilesExclude(t, databasePath, rawInput, "reviewable Definition tool answer", "Authorization", "raw_response")
}

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
	requiredScopes := workflowHTTPToolExecutionRequiredScopes
	if plan.SchemaVersion == workflowHTTPToolPlanSchemaV2 && plan.SourceKind == workflowHTTPToolSourceDefinition {
		requiredScopes = workflowDefinitionHTTPToolExecutionRequiredScopes
	}
	setWorkflowHTTPToolActionDevHeaders(request, strings.Join(requiredScopes, ","))
	request.Header.Set(savedWorkflowDraftDevApplicationHeader, plan.ApplicationID)
	return request
}
