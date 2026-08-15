package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"radishmind.local/services/platform/internal/config"
)

func TestApplicationInteractionSessionHTTPManagementPathAndStrictBoundary(t *testing.T) {
	definitionService, runContext, definitionRequest, bridgeClient, _ := workflowDefinitionExecutionFixture(t)
	server := &Server{config: config.Config{ApplicationSessionDevEnabled: true}, applicationCatalogRepository: definitionService.applications, workflowDefinitionReleaseRepository: definitionService.repository, applicationInteractionSessionRepository: newMemoryApplicationInteractionSessionRepository(), workspaceMembershipProvider: newDeterministicDevTestWorkspaceMembershipProvider()}
	auth := applicationInteractionSessionHTTPAuth(runContext, "application_sessions:read", "application_sessions:write")

	createBody := `{"workspace_id":"` + runContext.WorkspaceID + `","application_id":"` + runContext.ApplicationID + `","execution_profile":"workflow_definition_executor_v1","definition_id":"` + definitionRequest.DefinitionID + `"}`
	createRequest := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/application-sessions", strings.NewReader(createBody))
	createRequest.Header.Set(activeWorkspaceHeader, runContext.WorkspaceID)
	createRequest = createRequest.WithContext(withControlPlaneReadFakeAuthContext(createRequest.Context(), auth))
	createResponse := httptest.NewRecorder()
	server.handleCreateApplicationInteractionSession(createResponse, createRequest)
	var created applicationInteractionSessionEnvelope
	if createResponse.Code != http.StatusOK || json.Unmarshal(createResponse.Body.Bytes(), &created) != nil || created.FailureCode != nil || created.Session == nil || created.Session.RecordVersion != 1 || created.Session.Authority.WorkflowDefinition == nil {
		t.Fatalf("create application session HTTP: status=%d body=%s", createResponse.Code, createResponse.Body.String())
	}
	if bridgeClient.callCount() != 0 {
		t.Fatalf("session HTTP create called provider: %d", bridgeClient.callCount())
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/application-sessions?workspace_id="+runContext.WorkspaceID+"&application_id="+runContext.ApplicationID+"&execution_profile=workflow_definition_executor_v1", nil)
	listRequest = listRequest.WithContext(withControlPlaneReadFakeAuthContext(listRequest.Context(), auth))
	listResponse := httptest.NewRecorder()
	server.handleListApplicationInteractionSessions(listResponse, listRequest)
	var listed applicationInteractionSessionListEnvelope
	if listResponse.Code != http.StatusOK || json.Unmarshal(listResponse.Body.Bytes(), &listed) != nil || listed.FailureCode != nil || len(listed.Items) != 1 || listed.Items[0].SessionID != created.Session.SessionID {
		t.Fatalf("list application sessions HTTP: status=%d body=%s", listResponse.Code, listResponse.Body.String())
	}

	readRequest := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/application-sessions/"+created.Session.SessionID+"?workspace_id="+runContext.WorkspaceID+"&application_id="+runContext.ApplicationID, nil)
	readRequest.SetPathValue("session_id", created.Session.SessionID)
	readRequest = readRequest.WithContext(withControlPlaneReadFakeAuthContext(readRequest.Context(), auth))
	readResponse := httptest.NewRecorder()
	server.handleReadApplicationInteractionSession(readResponse, readRequest)
	if readResponse.Code != http.StatusOK || !strings.Contains(readResponse.Body.String(), created.Session.SessionID) {
		t.Fatalf("read application session HTTP: status=%d body=%s", readResponse.Code, readResponse.Body.String())
	}

	turnListRequest := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/application-sessions/"+created.Session.SessionID+"/turns?workspace_id="+runContext.WorkspaceID+"&application_id="+runContext.ApplicationID, nil)
	turnListRequest.SetPathValue("session_id", created.Session.SessionID)
	turnListRequest = turnListRequest.WithContext(withControlPlaneReadFakeAuthContext(turnListRequest.Context(), auth))
	turnListResponse := httptest.NewRecorder()
	server.handleListApplicationInteractionTurns(turnListResponse, turnListRequest)
	if turnListResponse.Code != http.StatusOK || !strings.Contains(turnListResponse.Body.String(), `"items":[]`) {
		t.Fatalf("list empty session turns HTTP: status=%d body=%s", turnListResponse.Code, turnListResponse.Body.String())
	}

	closeBody := `{"workspace_id":"` + runContext.WorkspaceID + `","application_id":"` + runContext.ApplicationID + `","expected_version":1}`
	closeRequest := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/application-sessions/"+created.Session.SessionID+"/close", strings.NewReader(closeBody))
	closeRequest.SetPathValue("session_id", created.Session.SessionID)
	closeRequest.Header.Set(activeWorkspaceHeader, runContext.WorkspaceID)
	closeRequest = closeRequest.WithContext(withControlPlaneReadFakeAuthContext(closeRequest.Context(), auth))
	closeResponse := httptest.NewRecorder()
	server.handleCloseApplicationInteractionSession(closeResponse, closeRequest)
	if closeResponse.Code != http.StatusOK || !strings.Contains(closeResponse.Body.String(), `"state":"closed"`) || bridgeClient.callCount() != 0 {
		t.Fatalf("close application session HTTP: status=%d body=%s provider=%d", closeResponse.Code, closeResponse.Body.String(), bridgeClient.callCount())
	}

	unknownRequest := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/application-sessions", strings.NewReader(strings.TrimSuffix(createBody, "}")+`,"input":"forbidden"}`))
	unknownRequest.Header.Set(activeWorkspaceHeader, runContext.WorkspaceID)
	unknownRequest = unknownRequest.WithContext(withControlPlaneReadFakeAuthContext(unknownRequest.Context(), auth))
	unknownResponse := httptest.NewRecorder()
	server.handleCreateApplicationInteractionSession(unknownResponse, unknownRequest)
	if unknownResponse.Code != http.StatusBadRequest || bridgeClient.callCount() != 0 {
		t.Fatalf("session HTTP accepted unknown input: status=%d body=%s provider=%d", unknownResponse.Code, unknownResponse.Body.String(), bridgeClient.callCount())
	}
	for _, forbidden := range []string{"forbidden", "input_text", "answer", "prompt", "credential", "token", "header"} {
		if strings.Contains(createResponse.Body.String()+listResponse.Body.String()+readResponse.Body.String()+closeResponse.Body.String(), `"`+forbidden+`"`) {
			t.Fatalf("session HTTP metadata exposed forbidden field %q", forbidden)
		}
	}
}

func TestApplicationInteractionSessionHTTPGateAndScopeFailClosed(t *testing.T) {
	definitionService, runContext, definitionRequest, bridgeClient, _ := workflowDefinitionExecutionFixture(t)
	server := &Server{applicationCatalogRepository: definitionService.applications, workflowDefinitionReleaseRepository: definitionService.repository, applicationInteractionSessionRepository: newMemoryApplicationInteractionSessionRepository(), workspaceMembershipProvider: newDeterministicDevTestWorkspaceMembershipProvider()}
	body := `{"workspace_id":"` + runContext.WorkspaceID + `","application_id":"` + runContext.ApplicationID + `","execution_profile":"workflow_definition_executor_v1","definition_id":"` + definitionRequest.DefinitionID + `"}`
	request := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/application-sessions", strings.NewReader(body))
	request = request.WithContext(withControlPlaneReadFakeAuthContext(request.Context(), applicationInteractionSessionHTTPAuth(runContext, "application_sessions:write")))
	response := httptest.NewRecorder()
	server.handleCreateApplicationInteractionSession(response, request)
	if response.Code != http.StatusBadGateway || bridgeClient.callCount() != 0 {
		t.Fatalf("disabled application session gate did not fail closed: status=%d body=%s", response.Code, response.Body.String())
	}

	server.config.ApplicationSessionDevEnabled = true
	denied := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/application-sessions", strings.NewReader(body))
	denied = denied.WithContext(withControlPlaneReadFakeAuthContext(denied.Context(), applicationInteractionSessionHTTPAuth(runContext, "application_sessions:read")))
	deniedResponse := httptest.NewRecorder()
	server.handleCreateApplicationInteractionSession(deniedResponse, denied)
	if deniedResponse.Code != http.StatusForbidden || !strings.Contains(deniedResponse.Body.String(), "scope_denied") || bridgeClient.callCount() != 0 {
		t.Fatalf("application session write scope did not fail closed: status=%d body=%s", deniedResponse.Code, deniedResponse.Body.String())
	}
}

func TestApplicationInteractionTurnHTTPExecutesStrictWorkflowV5AndDoesNotReplayProvider(t *testing.T) {
	definitionService, runContext, definitionRequest, bridgeClient, runStore := workflowDefinitionExecutionFixture(t)
	server := &Server{
		config: config.Config{ApplicationSessionDevEnabled: true, WorkflowDefinitionReleaseDevEnabled: true, WorkflowExecutorDevEnabled: true},
		bridge: bridgeClient, applicationCatalogRepository: definitionService.applications,
		workflowDefinitionReleaseRepository:     definitionService.repository,
		applicationInteractionSessionRepository: newMemoryApplicationInteractionSessionRepository(), workflowRunStore: runStore,
		applicationResultArtifactRepository: newMemoryApplicationResultArtifactRepository(),
		workspaceMembershipProvider:         newDeterministicDevTestWorkspaceMembershipProvider(),
	}
	auth := applicationInteractionSessionHTTPAuth(runContext, "application_sessions:write", "application_sessions:read", "application_sessions:execute")
	createBody := `{"workspace_id":"` + runContext.WorkspaceID + `","application_id":"` + runContext.ApplicationID + `","execution_profile":"workflow_definition_executor_v1","definition_id":"` + definitionRequest.DefinitionID + `"}`
	createRequest := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/application-sessions", strings.NewReader(createBody))
	createRequest.Header.Set(activeWorkspaceHeader, runContext.WorkspaceID)
	createRequest = createRequest.WithContext(withControlPlaneReadFakeAuthContext(createRequest.Context(), auth))
	createResponse := httptest.NewRecorder()
	server.handleCreateApplicationInteractionSession(createResponse, createRequest)
	var created applicationInteractionSessionEnvelope
	if createResponse.Code != http.StatusOK || json.Unmarshal(createResponse.Body.Bytes(), &created) != nil || created.Session == nil {
		t.Fatalf("create execution session: status=%d body=%s", createResponse.Code, createResponse.Body.String())
	}

	input := "private HTTP session turn input"
	turnBody := `{"workspace_id":"` + runContext.WorkspaceID + `","application_id":"` + runContext.ApplicationID + `","expected_session_version":1,"client_turn_key":"turn_http_001","save_result":true,"input_text":"` + input + `","condition_values":{},"model":"","temperature":null}`
	execute := func(body string, authContext controlPlaneReadAuthContext) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/application-sessions/"+created.Session.SessionID+"/turns", strings.NewReader(body))
		request.SetPathValue("session_id", created.Session.SessionID)
		request.Header.Set(activeWorkspaceHeader, runContext.WorkspaceID)
		request = request.WithContext(withControlPlaneReadFakeAuthContext(request.Context(), authContext))
		response := httptest.NewRecorder()
		server.handleExecuteApplicationInteractionTurn(response, request)
		return response
	}
	response := execute(turnBody, auth)
	var executed applicationInteractionTurnEnvelope
	if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &executed) != nil || executed.FailureCode != nil || executed.Turn == nil || executed.Turn.Status != string(WorkflowRunStatusSucceeded) || executed.Turn.RunRef == nil || executed.Turn.RunRef.SchemaVersion != workflowRunRecordDefinitionSchemaVersion || executed.AdvisoryOutput == "" || executed.ResultArtifact == nil || executed.ResultArtifactFailureCode != nil || bridgeClient.callCount() != 1 {
		t.Fatalf("execute workflow session turn: status=%d body=%s bridge=%d", response.Code, response.Body.String(), bridgeClient.callCount())
	}
	if strings.Contains(response.Body.String(), input) {
		t.Fatalf("turn response echoed private input: %s", response.Body.String())
	}
	replay := execute(turnBody, auth)
	if replay.Code != http.StatusOK || !strings.Contains(replay.Body.String(), `"idempotent_replay":true`) || strings.Contains(replay.Body.String(), executed.AdvisoryOutput) || !strings.Contains(replay.Body.String(), executed.ResultArtifact.ArtifactID) || bridgeClient.callCount() != 1 {
		t.Fatalf("turn HTTP retry repeated provider: status=%d body=%s bridge=%d", replay.Code, replay.Body.String(), bridgeClient.callCount())
	}

	query := "?workspace_id=" + runContext.WorkspaceID + "&application_id=" + runContext.ApplicationID
	listRequest := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/application-sessions/"+created.Session.SessionID+"/result-artifacts"+query, nil)
	listRequest.SetPathValue("session_id", created.Session.SessionID)
	listRequest.Header.Set(activeWorkspaceHeader, runContext.WorkspaceID)
	listRequest = listRequest.WithContext(withControlPlaneReadFakeAuthContext(listRequest.Context(), auth))
	listResponse := httptest.NewRecorder()
	server.handleListApplicationResultArtifacts(listResponse, listRequest)
	var listed applicationResultArtifactListEnvelope
	if listResponse.Code != http.StatusOK || json.Unmarshal(listResponse.Body.Bytes(), &listed) != nil || listed.FailureCode != nil || len(listed.Items) != 1 ||
		listed.Items[0].ArtifactID != executed.ResultArtifact.ArtifactID || strings.Contains(listResponse.Body.String(), executed.AdvisoryOutput) || listResponse.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("list result artifacts: status=%d body=%s headers=%v", listResponse.Code, listResponse.Body.String(), listResponse.Header())
	}
	readRequest := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/application-sessions/"+created.Session.SessionID+"/result-artifacts/"+executed.ResultArtifact.ArtifactID+query, nil)
	readRequest.SetPathValue("session_id", created.Session.SessionID)
	readRequest.SetPathValue("artifact_id", executed.ResultArtifact.ArtifactID)
	readRequest.Header.Set(activeWorkspaceHeader, runContext.WorkspaceID)
	readRequest = readRequest.WithContext(withControlPlaneReadFakeAuthContext(readRequest.Context(), auth))
	readResponse := httptest.NewRecorder()
	server.handleReadApplicationResultArtifact(readResponse, readRequest)
	var read applicationResultArtifactEnvelope
	if readResponse.Code != http.StatusOK || json.Unmarshal(readResponse.Body.Bytes(), &read) != nil || read.FailureCode != nil || read.Artifact == nil ||
		read.Artifact.Content != executed.AdvisoryOutput || read.Artifact.TurnID != executed.Turn.TurnID || readResponse.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("read result artifact: status=%d body=%s headers=%v", readResponse.Code, readResponse.Body.String(), readResponse.Header())
	}
	deniedRead := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/application-sessions/"+created.Session.SessionID+"/result-artifacts/"+executed.ResultArtifact.ArtifactID+query, nil)
	deniedRead.SetPathValue("session_id", created.Session.SessionID)
	deniedRead.SetPathValue("artifact_id", executed.ResultArtifact.ArtifactID)
	deniedRead.Header.Set(activeWorkspaceHeader, runContext.WorkspaceID)
	deniedRead = deniedRead.WithContext(withControlPlaneReadFakeAuthContext(deniedRead.Context(), applicationInteractionSessionHTTPAuth(runContext, "application_sessions:execute")))
	deniedReadResponse := httptest.NewRecorder()
	server.handleReadApplicationResultArtifact(deniedReadResponse, deniedRead)
	if deniedReadResponse.Code != http.StatusForbidden || !strings.Contains(deniedReadResponse.Body.String(), "scope_denied") {
		t.Fatalf("result artifact read scope did not fail closed: status=%d body=%s", deniedReadResponse.Code, deniedReadResponse.Body.String())
	}

	unknown := execute(strings.TrimSuffix(turnBody, "}")+`,"authority_digest":"sha256:`+strings.Repeat("a", 64)+`"}`, auth)
	if unknown.Code != http.StatusBadRequest || bridgeClient.callCount() != 1 {
		t.Fatalf("turn HTTP accepted client authority: status=%d body=%s bridge=%d", unknown.Code, unknown.Body.String(), bridgeClient.callCount())
	}
	denied := execute(strings.Replace(turnBody, "turn_http_001", "turn_http_002", 1), applicationInteractionSessionHTTPAuth(runContext, "application_sessions:read"))
	if denied.Code != http.StatusForbidden || !strings.Contains(denied.Body.String(), "scope_denied") || bridgeClient.callCount() != 1 {
		t.Fatalf("turn execute scope did not fail closed: status=%d body=%s bridge=%d", denied.Code, denied.Body.String(), bridgeClient.callCount())
	}
}

func applicationInteractionSessionHTTPAuth(ctx WorkflowRunContext, scopes ...string) controlPlaneReadAuthContext {
	return controlPlaneReadAuthContext{
		AuthMode: controlPlaneReadAuthModeDevHeaders, IdentityContext: "dev:application-session-test",
		TenantBinding: ctx.TenantRef, SubjectBinding: ctx.ActorRef, ScopeGrants: scopes,
		AuditContext:     "audit_application_session_http",
		VerifiedIdentity: &VerifiedControlPlaneIdentity{SubjectRef: ctx.ActorRef, TenantRef: ctx.TenantRef},
		ResourceBinding:  ControlPlaneResourceBinding{TenantRef: ctx.TenantRef, TenantVerified: true},
		WorkspaceMemberships: []VerifiedWorkspaceMembershipAssertion{{
			TenantRef: ctx.TenantRef, SubjectRef: ctx.ActorRef, WorkspaceID: ctx.WorkspaceID,
			PermissionGrants: append([]string{}, scopes...),
		}},
	}
}
