package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
)

func TestApplicationInteractionStructuredSessionV4DelegatesV8AndKeepsValuesTransient(t *testing.T) {
	coordinator, ctx, session, bridgeClient, repository := structuredWorkflowApplicationInteractionCoordinatorFixture(t)
	artifactService := newApplicationResultArtifactService(newMemoryApplicationResultArtifactRepository(), repository)
	artifactService.newID = applicationResultArtifactStableIDGenerator()
	coordinator = coordinator.withResultArtifacts(artifactService)
	if session.SchemaVersion != applicationStructuredSessionSchemaVersion || session.Authority.SchemaVersion != applicationInteractionStructuredAuthoritySchemaVersion ||
		session.Authority.WorkflowDefinition == nil || session.Authority.WorkflowDefinition.InputContract == nil ||
		session.Authority.WorkflowDefinition.InputContract.ContractDigest == "" {
		t.Fatalf("structured session authority was not frozen: %#v", session)
	}

	privateValue := "private-structured-session-customer"
	result := coordinator.Execute(ctx, session.SessionID, ApplicationInteractionTurnExecutionInput{
		ExpectedSessionVersion: session.RecordVersion,
		ClientTurnKey:          "turn_structured_001",
		SaveResult:             true,
		Inputs:                 map[string]any{"customer_name": privateValue, "retry_count": 2},
		ConditionValues:        map[string]bool{},
	})
	if result.FailureCode != "" || result.Turn == nil || result.Turn.SchemaVersion != applicationStructuredSessionTurnSchemaVersion ||
		result.Turn.Status != string(WorkflowRunStatusSucceeded) || result.Turn.RunRef == nil ||
		result.Turn.RunRef.SchemaVersion != workflowRunRecordDefinitionStructuredSchemaVersion || result.AdvisoryOutput == "" ||
		result.ResultArtifact == nil || result.ResultArtifactFailureCode != "" || bridgeClient.callCount() != 1 {
		t.Fatalf("structured session turn failed: result=%#v bridge=%d", result, bridgeClient.callCount())
	}
	turn := result.Turn
	if turn.InputContractID != session.Authority.WorkflowDefinition.InputContract.ContractID ||
		turn.InputContractDigest != session.Authority.WorkflowDefinition.InputContract.ContractDigest || len(turn.InputFields) != 2 ||
		turn.InputFields[0].Name != "customer_name" || turn.InputFields[1].Name != "retry_count" {
		t.Fatalf("structured turn metadata drifted: %#v", turn)
	}
	sessionPayload, sessionMarshalErr := json.Marshal(session)
	turnPayload, turnMarshalErr := json.Marshal(turn)
	if sessionMarshalErr != nil || turnMarshalErr != nil ||
		validateApplicationInteractionContractJSON(applicationStructuredSessionSchemaVersion, sessionPayload) != nil ||
		validateApplicationInteractionContractJSON(applicationStructuredSessionTurnSchemaVersion, turnPayload) != nil {
		t.Fatalf("structured Session v4 contract serialization drifted: session_err=%v turn_err=%v session=%s turn=%s", sessionMarshalErr, turnMarshalErr, sessionPayload, turnPayload)
	}
	stored, err := repository.ListTurns(ctx, session.SessionID)
	payload, marshalErr := json.Marshal(stored)
	if err != nil || marshalErr != nil || len(stored) != 1 || strings.Contains(string(payload), privateValue) || strings.Contains(string(payload), result.AdvisoryOutput) {
		t.Fatalf("structured session persisted transient values: turns=%#v err=%v marshal=%v payload=%s", stored, err, marshalErr, payload)
	}
}

func TestApplicationInteractionStructuredSessionV4RejectsContractDriftBeforeReservation(t *testing.T) {
	coordinator, ctx, session, bridgeClient, repository := structuredWorkflowApplicationInteractionCoordinatorFixture(t)
	drifted := cloneApplicationInteractionAuthority(session.Authority)
	driftedContract := drifted.WorkflowDefinition.InputContract
	driftedContract.Summary += " changed"
	driftedContract.ContractDigest = ""
	normalized, code, summary := normalizeWorkflowStructuredInputContract(savedWorkflowDraftContractFromDefinition(*driftedContract))
	if code != "" {
		t.Fatalf("normalize drifted contract: code=%s summary=%s", code, summary)
	}
	driftedContract.ContractDigest = normalized.ContractDigest
	drifted.AuthorityDigest, _ = applicationInteractionAuthorityDigest(drifted)
	coordinator.sessions.resolver = applicationInteractionAuthorityResolverFunc(func(ApplicationInteractionContext, ApplicationInteractionProfileBinding) (ApplicationInteractionAuthoritySnapshot, string) {
		return drifted, ""
	})

	result := coordinator.Execute(ctx, session.SessionID, ApplicationInteractionTurnExecutionInput{
		ExpectedSessionVersion: session.RecordVersion, ClientTurnKey: "turn_structured_drift",
		Inputs: map[string]any{"customer_name": "Ada", "retry_count": 2}, ConditionValues: map[string]bool{},
	})
	turns, err := repository.ListTurns(ctx, session.SessionID)
	if result.FailureCode != ApplicationInteractionFailureAuthorityChanged || result.Turn != nil || err != nil || len(turns) != 0 || bridgeClient.callCount() != 0 {
		t.Fatalf("structured contract drift escaped pre-reservation checkpoint: result=%#v turns=%#v err=%v bridge=%d", result, turns, err, bridgeClient.callCount())
	}
}

func TestApplicationInteractionStructuredSessionV4RejectsLegacyAndInvalidInputsBeforeReservation(t *testing.T) {
	tests := []struct {
		name  string
		input ApplicationInteractionTurnExecutionInput
		code  string
	}{
		{name: "legacy input text", input: ApplicationInteractionTurnExecutionInput{InputText: "legacy input"}, code: ApplicationInteractionFailurePayloadInvalid},
		{name: "missing required field", input: ApplicationInteractionTurnExecutionInput{Inputs: map[string]any{}}, code: string(WorkflowRunFailureInputRequiredFieldMissing)},
		{name: "unknown field", input: ApplicationInteractionTurnExecutionInput{Inputs: map[string]any{"customer_name": "Ada", "unknown": true}}, code: string(WorkflowRunFailureInputUnknownField)},
		{name: "secret value", input: ApplicationInteractionTurnExecutionInput{Inputs: map[string]any{"customer_name": "password=hunter2"}}, code: string(WorkflowRunFailureInputSecretMaterialForbidden)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			coordinator, ctx, session, bridgeClient, repository := structuredWorkflowApplicationInteractionCoordinatorFixture(t)
			test.input.ExpectedSessionVersion = session.RecordVersion
			test.input.ClientTurnKey = "turn_structured_invalid"
			test.input.ConditionValues = map[string]bool{}
			result := coordinator.Execute(ctx, session.SessionID, test.input)
			turns, err := repository.ListTurns(ctx, session.SessionID)
			if result.FailureCode != test.code || result.Turn != nil || err != nil || len(turns) != 0 || bridgeClient.callCount() != 0 {
				t.Fatalf("invalid structured session input escaped boundary: result=%#v turns=%#v err=%v bridge=%d", result, turns, err, bridgeClient.callCount())
			}
		})
	}
}

func TestApplicationInteractionStructuredSessionV4HTTPStrictUnionAndV8Response(t *testing.T) {
	definitionService, runContext, runRequest, bridgeClient, runStore := workflowDefinitionStructuredExecutionFixture(t)
	server := &Server{
		config: config.Config{ApplicationSessionDevEnabled: true, WorkflowDefinitionReleaseDevEnabled: true, WorkflowExecutorDevEnabled: true},
		bridge: bridgeClient, applicationCatalogRepository: definitionService.applications,
		workflowDefinitionReleaseRepository:     definitionService.repository,
		applicationInteractionSessionRepository: newMemoryApplicationInteractionSessionRepository(), workflowRunStore: runStore,
		workspaceMembershipProvider: newDeterministicDevTestWorkspaceMembershipProvider(),
	}
	auth := applicationInteractionSessionHTTPAuth(runContext, "application_sessions:write", "application_sessions:read", "application_sessions:execute")
	createBody := `{"workspace_id":"` + runContext.WorkspaceID + `","application_id":"` + runContext.ApplicationID + `","execution_profile":"workflow_definition_executor_v2","definition_id":"` + runRequest.DefinitionID + `"}`
	createRequest := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/application-sessions", strings.NewReader(createBody))
	createRequest.Header.Set(activeWorkspaceHeader, runContext.WorkspaceID)
	createRequest = createRequest.WithContext(withControlPlaneReadFakeAuthContext(createRequest.Context(), auth))
	createResponse := httptest.NewRecorder()
	server.handleCreateApplicationInteractionSession(createResponse, createRequest)
	var created applicationInteractionSessionEnvelope
	if createResponse.Code != http.StatusOK || json.Unmarshal(createResponse.Body.Bytes(), &created) != nil || created.FailureCode != nil || created.Session == nil ||
		created.Session.SchemaVersion != applicationStructuredSessionSchemaVersion || created.Session.Authority.SchemaVersion != applicationInteractionStructuredAuthoritySchemaVersion {
		t.Fatalf("create structured HTTP session: status=%d body=%s", createResponse.Code, createResponse.Body.String())
	}

	execute := func(body string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/application-sessions/"+created.Session.SessionID+"/turns", strings.NewReader(body))
		request.SetPathValue("session_id", created.Session.SessionID)
		request.Header.Set(activeWorkspaceHeader, runContext.WorkspaceID)
		request = request.WithContext(withControlPlaneReadFakeAuthContext(request.Context(), auth))
		response := httptest.NewRecorder()
		server.handleExecuteApplicationInteractionTurn(response, request)
		return response
	}
	base := `{"workspace_id":"` + runContext.WorkspaceID + `","application_id":"` + runContext.ApplicationID + `","expected_session_version":1,"client_turn_key":"turn_http_structured_001","condition_values":{},"model":"","temperature":null,`
	mixed := execute(base + `"input_text":"legacy","inputs":{"customer_name":"Ada","retry_count":2}}`)
	if mixed.Code != http.StatusOK || !strings.Contains(mixed.Body.String(), ApplicationInteractionFailurePayloadInvalid) || bridgeClient.callCount() != 0 {
		t.Fatalf("structured HTTP session accepted mixed union: status=%d body=%s bridge=%d", mixed.Code, mixed.Body.String(), bridgeClient.callCount())
	}
	privateValue := "private-http-structured-customer"
	response := execute(base + `"inputs":{"customer_name":"` + privateValue + `","retry_count":2}}`)
	var executed applicationInteractionTurnEnvelope
	if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &executed) != nil || executed.FailureCode != nil || executed.Turn == nil ||
		executed.Turn.SchemaVersion != applicationStructuredSessionTurnSchemaVersion || executed.Turn.RunRef == nil ||
		executed.Turn.RunRef.SchemaVersion != workflowRunRecordDefinitionStructuredSchemaVersion || len(executed.Turn.InputFields) != 2 || bridgeClient.callCount() != 1 {
		t.Fatalf("execute structured HTTP session: status=%d body=%s bridge=%d", response.Code, response.Body.String(), bridgeClient.callCount())
	}
	if strings.Contains(response.Body.String(), privateValue) || strings.Contains(response.Body.String(), `"input_text"`) || strings.Contains(response.Body.String(), `"inputs"`) {
		t.Fatalf("structured HTTP session response exposed request values: %s", response.Body.String())
	}
}

func TestSQLiteApplicationInteractionStructuredSessionV4RestartsWithMetadataOnly(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "structured-application-session.db")
	runtime := openApplicationInteractionSQLiteRuntime(t, databasePath)
	definitionService, runContext, runRequest, bridgeClient, _ := workflowDefinitionStructuredExecutionFixture(t)
	resolver := newExactApplicationInteractionAuthorityResolver(definitionService.applications, definitionService.repository, nil, workflowRAGApplicationAuthorityResolver{})
	repository := newSQLiteApplicationInteractionSessionRepository(runtime.DB())
	sessions := newApplicationInteractionSessionService(repository, resolver)
	sessions.newID = applicationInteractionStableIDGenerator()
	baseTime := time.Date(2026, 8, 11, 10, 0, 0, 0, time.UTC)
	sessions.now = func() time.Time { return baseTime }
	ctx := ApplicationInteractionContext{RequestContext: runContext.RequestContext, RequestID: "request_structured_sqlite", TenantRef: runContext.TenantRef, WorkspaceID: runContext.WorkspaceID, ApplicationID: runContext.ApplicationID, ActorRef: runContext.ActorRef, OwnerSubjectRef: runContext.ActorRef, AuditRef: "audit_structured_sqlite", WriteEnabled: true}
	created := sessions.Create(ctx, ApplicationInteractionSessionCreateInput{ProfileBinding: ApplicationInteractionProfileBinding{ExecutionProfile: applicationInteractionProfileWorkflowStructured, DefinitionID: runRequest.DefinitionID}})
	if created.FailureCode != "" || created.Session == nil {
		t.Fatalf("create SQLite structured session: %#v", created)
	}
	coordinator := newApplicationInteractionTurnCoordinator(sessions, resolver, definitionService.StartRun, nil)
	coordinator.now = func() time.Time { return baseTime.Add(time.Second) }
	privateValue := "private-sqlite-structured-customer"
	executed := coordinator.Execute(ctx, created.Session.SessionID, ApplicationInteractionTurnExecutionInput{
		ExpectedSessionVersion: created.Session.RecordVersion, ClientTurnKey: "turn_structured_sqlite",
		Inputs: map[string]any{"customer_name": privateValue, "retry_count": 2}, ConditionValues: map[string]bool{},
	})
	if executed.FailureCode != "" || executed.Turn == nil || bridgeClient.callCount() != 1 {
		t.Fatalf("execute SQLite structured session: result=%#v bridge=%d", executed, bridgeClient.callCount())
	}
	var sessionPayload, turnPayload string
	if err := runtime.DB().QueryRowContext(context.Background(), `SELECT sanitized_session_payload FROM application_interaction_sessions WHERE session_id=?`, created.Session.SessionID).Scan(&sessionPayload); err != nil {
		t.Fatal(err)
	}
	if err := runtime.DB().QueryRowContext(context.Background(), `SELECT sanitized_turn_payload FROM application_interaction_session_turns WHERE turn_id=?`, executed.Turn.TurnID).Scan(&turnPayload); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(sessionPayload+turnPayload, privateValue) || !strings.Contains(turnPayload, `"input_contract_id"`) || !strings.Contains(turnPayload, `"input_fields"`) {
		t.Fatalf("SQLite structured session persistence drifted: session=%s turn=%s", sessionPayload, turnPayload)
	}
	if err := runtime.Close(); err != nil {
		t.Fatal(err)
	}
	restarted := openApplicationInteractionSQLiteRuntime(t, databasePath)
	defer restarted.Close()
	restartedService := newApplicationInteractionSessionService(newSQLiteApplicationInteractionSessionRepository(restarted.DB()), resolver)
	read := restartedService.Read(ctx, created.Session.SessionID)
	turns, failure := restartedService.ListTurns(ctx, created.Session.SessionID)
	if read.FailureCode != "" || read.Session == nil || read.Session.SchemaVersion != applicationStructuredSessionSchemaVersion || failure != "" || len(turns) != 1 ||
		turns[0].SchemaVersion != applicationStructuredSessionTurnSchemaVersion || turns[0].RunRef == nil || turns[0].RunRef.SchemaVersion != workflowRunRecordDefinitionStructuredSchemaVersion || len(turns[0].InputFields) != 2 {
		t.Fatalf("read SQLite structured session after restart: read=%#v turns=%#v failure=%s", read, turns, failure)
	}
}

func structuredWorkflowApplicationInteractionCoordinatorFixture(t *testing.T) (applicationInteractionTurnCoordinator, ApplicationInteractionContext, ApplicationInteractionSession, *workflowExecutorTestBridge, *memoryApplicationInteractionSessionRepository) {
	t.Helper()
	definitionService, runContext, runRequest, bridgeClient, _ := workflowDefinitionStructuredExecutionFixture(t)
	resolver := newExactApplicationInteractionAuthorityResolver(definitionService.applications, definitionService.repository, nil, workflowRAGApplicationAuthorityResolver{})
	repository := newMemoryApplicationInteractionSessionRepository()
	sessions := newApplicationInteractionSessionService(repository, resolver)
	sessions.newID = applicationInteractionStableIDGenerator()
	baseTime := time.Date(2026, 8, 11, 9, 0, 0, 0, time.UTC)
	sessions.now = func() time.Time { return baseTime }
	ctx := ApplicationInteractionContext{RequestContext: runContext.RequestContext, RequestID: "request_structured_session", TenantRef: runContext.TenantRef, WorkspaceID: runContext.WorkspaceID, ApplicationID: runContext.ApplicationID, ActorRef: runContext.ActorRef, OwnerSubjectRef: runContext.ActorRef, AuditRef: "audit_structured_session", WriteEnabled: true}
	if authority, failure := resolver.Resolve(ctx, ApplicationInteractionProfileBinding{ExecutionProfile: applicationInteractionProfileWorkflowStructured, DefinitionID: runRequest.DefinitionID}); failure != "" {
		t.Fatalf("resolve structured session authority: failure=%s authority=%#v", failure, authority)
	}
	created := sessions.Create(ctx, ApplicationInteractionSessionCreateInput{ProfileBinding: ApplicationInteractionProfileBinding{ExecutionProfile: applicationInteractionProfileWorkflowStructured, DefinitionID: runRequest.DefinitionID}})
	if created.FailureCode != "" || created.Session == nil {
		t.Fatalf("create structured application session: %#v", created)
	}
	coordinator := newApplicationInteractionTurnCoordinator(sessions, resolver, definitionService.StartRun, nil)
	coordinator.now = func() time.Time { return baseTime.Add(time.Second) }
	return coordinator, ctx, *created.Session, bridgeClient, repository
}
