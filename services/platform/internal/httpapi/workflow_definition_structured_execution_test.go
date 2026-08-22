package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
)

func TestWorkflowDefinitionStructuredExecutionUsesExactContractAndMetadataOnlyV8(t *testing.T) {
	service, runContext, request, bridgeClient, store := workflowDefinitionStructuredExecutionFixture(t)
	privateValue := "private-customer-Ada"
	request.Inputs = map[string]any{"retry_count": 2, "customer_name": privateValue}
	result := service.StartRun(runContext, request)
	if result.FailureCode != "" || result.Record == nil || result.Record.Status != WorkflowRunStatusSucceeded || result.AdvisoryOutput == "" {
		t.Fatalf("structured definition execution failed: %#v", result)
	}
	record := result.Record
	if record.SchemaVersion != workflowRunRecordDefinitionStructuredSchemaVersion || record.ExecutionProfile != workflowDefinitionStructuredExecutorProfile ||
		record.InputContractID == "" || record.InputContractDigest == "" || record.InputDigest == "" || record.InputBytes < 2 || len(record.InputFields) != 2 {
		t.Fatalf("Run v8 identity drifted: %#v", record)
	}
	if record.InputFields[0].Name != "customer_name" || record.InputFields[1].Name != "retry_count" || record.Output != "" || record.DefinitionAuthority == nil {
		t.Fatalf("Run v8 metadata projection drifted: %#v", record)
	}
	persisted, err := json.Marshal(record)
	if err != nil || bytes.Contains(persisted, []byte(privateValue)) || bytes.Contains(persisted, []byte(result.AdvisoryOutput)) {
		t.Fatalf("Run v8 persisted private values: err=%v payload=%s", err, persisted)
	}
	providerRequest := bridgeClient.lastRequest()
	if bridgeClient.callCount() != 1 || !bytes.Contains(providerRequest, []byte(privateValue)) ||
		!bytes.Contains(providerRequest, []byte("value_type")) || !bytes.Contains(providerRequest, []byte(record.InputContractDigest)) {
		t.Fatalf("executor v2 did not send the bounded typed packet: bridge=%d request=%s", bridgeClient.callCount(), providerRequest)
	}
	stored, found, err := store.ReadRun(runContext, record.RunID)
	if err != nil || !found || stored.SchemaVersion != workflowRunRecordDefinitionStructuredSchemaVersion || stored.InputDigest != record.InputDigest {
		t.Fatalf("memory Run v8 history read failed: found=%t err=%v record=%#v", found, err, stored)
	}
	page, err := store.ListRuns(runContext, WorkflowRunListFilter{ExecutionSourceKind: workflowDefinitionExecutionSourceKind, ExecutionSourceID: record.ExecutionSourceID, ExecutionSourceVersion: record.ExecutionSourceVersion, Limit: 10})
	if err != nil || len(page.Records) != 1 || page.Records[0].RunID != record.RunID || page.Records[0].InputContractDigest != record.InputContractDigest {
		t.Fatalf("memory Run v8 history list failed: page=%#v err=%v", page, err)
	}
}

func TestWorkflowDefinitionStructuredExecutionStrictUnionAndFailuresBeforeProvider(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*WorkflowDefinitionRunRequest)
		code   WorkflowRunFailureCode
	}{
		{name: "missing required", mutate: func(request *WorkflowDefinitionRunRequest) { request.Inputs = map[string]any{} }, code: WorkflowRunFailureInputRequiredFieldMissing},
		{name: "unknown field", mutate: func(request *WorkflowDefinitionRunRequest) {
			request.Inputs = map[string]any{"customer_name": "Ada", "unknown": true}
		}, code: WorkflowRunFailureInputUnknownField},
		{name: "invalid type", mutate: func(request *WorkflowDefinitionRunRequest) { request.Inputs = map[string]any{"customer_name": 12} }, code: WorkflowRunFailureInputValueTypeInvalid},
		{name: "secret", mutate: func(request *WorkflowDefinitionRunRequest) {
			request.Inputs = map[string]any{"customer_name": "password=hunter2"}
		}, code: WorkflowRunFailureInputSecretMaterialForbidden},
		{name: "v2 with input text", mutate: func(request *WorkflowDefinitionRunRequest) { request.Inputs = nil; request.InputText = "legacy input" }, code: WorkflowRunFailureInputSchemaUnsupported},
		{name: "mixed union", mutate: func(request *WorkflowDefinitionRunRequest) { request.InputText = "legacy input" }, code: WorkflowRunFailureInputSchemaUnsupported},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service, runContext, request, bridgeClient, store := workflowDefinitionStructuredExecutionFixture(t)
			test.mutate(&request)
			result := service.StartRun(runContext, request)
			if result.FailureCode != test.code || result.Record != nil || bridgeClient.callCount() != 0 {
				t.Fatalf("failure did not close before provider: result=%#v bridge=%d", result, bridgeClient.callCount())
			}
			page, err := store.ListRuns(runContext, WorkflowRunListFilter{Limit: 10})
			if err != nil || len(page.Records) != 0 {
				t.Fatalf("rejected input created Run evidence: page=%#v err=%v", page, err)
			}
		})
	}

	legacyService, legacyContext, legacyRequest, legacyBridge, _ := workflowDefinitionExecutionFixture(t)
	legacyRequest.InputText = ""
	legacyRequest.Inputs = map[string]any{"customer_name": "Ada"}
	if result := legacyService.StartRun(legacyContext, legacyRequest); result.FailureCode != WorkflowRunFailureInputSchemaUnsupported || result.Record != nil || legacyBridge.callCount() != 0 {
		t.Fatalf("v1 accepted v2 inputs: %#v bridge=%d", result, legacyBridge.callCount())
	}
}

func TestWorkflowDefinitionStructuredAuthorityDriftUsesStableInputFailure(t *testing.T) {
	service, runContext, request, bridgeClient, store := workflowDefinitionStructuredExecutionFixture(t)
	service.repository = &driftingWorkflowDefinitionRepository{workflowDefinitionReleaseRepository: service.repository}
	result := service.StartRun(runContext, request)
	if result.FailureCode != WorkflowRunFailureInputAuthorityDrift || result.Record == nil || result.Record.SchemaVersion != workflowRunRecordDefinitionStructuredSchemaVersion ||
		result.Record.Status != WorkflowRunStatusFailed || result.Record.SideEffects.ProviderCalls != 0 || bridgeClient.callCount() != 0 {
		t.Fatalf("structured authority drift did not fail closed: %#v bridge=%d", result, bridgeClient.callCount())
	}
	stored, found, err := store.ReadRun(runContext, result.Record.RunID)
	if err != nil || !found || stored.FailureCode != WorkflowRunFailureInputAuthorityDrift {
		t.Fatalf("structured authority drift evidence missing: %#v found=%t err=%v", stored, found, err)
	}
}

func TestWorkflowDefinitionStructuredHTTPAndRunHistoryMetadataOnly(t *testing.T) {
	server := NewServer(config.Config{ControlPlaneReadDevAuthEnabled: true, WorkflowSavedDraftDevHTTPEnabled: true, WorkflowSavedDraftDevWriteEnabled: true, WorkflowDefinitionReleaseDevEnabled: true, WorkflowExecutorDevEnabled: true, Provider: "mock"}, Options{BuildVersion: "test"})
	t.Cleanup(server.Close)
	bridgeClient := &workflowExecutorTestBridge{}
	server.bridge = bridgeClient
	applicationID := "app_cccccccccccccccc"
	actor := "subject_demo_user"
	catalog := server.applicationCatalogService()
	catalog.newID = func() (string, error) { return applicationID, nil }
	applicationContext := ApplicationCatalogContext{RequestContext: context.Background(), RequestID: "request_structured_http_app", TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", ActorRef: actor, OwnerSubjectRef: actor, AuditRef: "audit_structured_http_app", WriteEnabled: true}
	if created := catalog.Create(applicationContext, ApplicationCatalogCreateInput{DisplayName: "Structured HTTP definition app", ApplicationKind: "workflow_copilot"}); created.FailureCode != "" {
		t.Fatalf("create application fixture: %#v", created)
	}
	releaseContext := WorkflowDefinitionReleaseContext{RequestContext: context.Background(), TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", ApplicationID: applicationID, OwnerSubjectRef: actor, ActorRef: actor, RequestID: "request_structured_http_release", AuditRef: "audit_structured_http_release"}
	draft := executableWorkflowStructuredDraftForTest(applicationID)
	candidate, err := server.workflowDefinitionReleaseRepository.CreateCandidate(releaseContext, "candidate_structured_http", "definition_structured_http", "", draft, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	_, version, err := server.workflowDefinitionReleaseRepository.Review(releaseContext, candidate.CandidateID, 0, "approve", "approve structured HTTP definition", candidate.SourceDraftDigest, time.Now().UTC())
	if err != nil || version == nil {
		t.Fatalf("approve structured definition: %#v err=%v", version, err)
	}
	activation, err := server.workflowDefinitionReleaseRepository.DecideActivation(releaseContext, version.DefinitionID, 0, "activate", version.Version, "activate structured HTTP definition", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	privateValue := "private-http-customer"
	body := map[string]any{
		"workspace_id": "workspace_demo", "application_id": applicationID, "definition_id": version.DefinitionID,
		"expected_pointer_version": activation.PointerVersion, "expected_definition_version": version.Version,
		"expected_definition_digest": version.DefinitionDigest, "inputs": map[string]any{"customer_name": privateValue, "dry_run": true},
		"condition_values": map[string]bool{}, "model": "", "temperature": nil,
	}
	response := performWorkflowDefinitionStructuredHTTPRequest(t, server, body, applicationID)
	envelope := decodeWorkflowRunEnvelope(t, response, http.StatusOK)
	if envelope.FailureCode != nil || envelope.Run == nil || envelope.Run.SchemaVersion != workflowRunRecordDefinitionStructuredSchemaVersion ||
		envelope.Run.InputContractDigest == "" || len(envelope.Run.InputFields) != 2 || envelope.AdvisoryOutput == "" || bridgeClient.callCount() != 1 {
		t.Fatalf("structured HTTP execution failed: %#v body=%s", envelope, response.Body.String())
	}
	if strings.Contains(response.Body.String(), privateValue) {
		t.Fatalf("structured HTTP response persisted private input: %s", response.Body.String())
	}
	runPayload, err := json.Marshal(envelope.Run)
	if err != nil || strings.Contains(string(runPayload), envelope.AdvisoryOutput) {
		t.Fatalf("Run v8 persisted advisory output: err=%v run=%s", err, runPayload)
	}

	readRequest := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/workflow-runs/"+envelope.Run.RunID+"?workspace_id=workspace_demo&application_id="+applicationID, nil)
	setControlPlaneReadDevAuthHeaders(readRequest)
	readRequest.Header.Set(controlPlaneReadDevScopesHeader, "workflow_runs:read")
	readRequest.Header.Set(controlPlaneReadDevMembershipPermHeader, "workflow_runs:read")
	readRequest.Header.Set(savedWorkflowDraftDevWorkspaceHeader, "workspace_demo")
	readRequest.Header.Set(savedWorkflowDraftDevApplicationHeader, applicationID)
	readResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(readResponse, readRequest)
	readEnvelope := decodeWorkflowRunEnvelope(t, readResponse, http.StatusOK)
	if readEnvelope.Run == nil || readEnvelope.Run.InputContractDigest != envelope.Run.InputContractDigest || strings.Contains(readResponse.Body.String(), privateValue) {
		t.Fatalf("Run History lost metadata-only v8 contract: %#v body=%s", readEnvelope, readResponse.Body.String())
	}

	wrongShape := cloneStructuredHTTPBody(body)
	delete(wrongShape, "inputs")
	wrongShape["input_text"] = "legacy input"
	wrongResponse := performWorkflowDefinitionStructuredHTTPRequest(t, server, wrongShape, applicationID)
	wrongEnvelope := decodeWorkflowRunEnvelope(t, wrongResponse, http.StatusOK)
	if wrongEnvelope.FailureCode == nil || *wrongEnvelope.FailureCode != string(WorkflowRunFailureInputSchemaUnsupported) || wrongEnvelope.Run != nil || bridgeClient.callCount() != 1 {
		t.Fatalf("v2 HTTP accepted input_text: %#v bridge=%d", wrongEnvelope, bridgeClient.callCount())
	}
}

func workflowDefinitionStructuredExecutionFixture(t *testing.T) (workflowDefinitionExecutionService, WorkflowRunContext, WorkflowDefinitionRunRequest, *workflowExecutorTestBridge, *memoryWorkflowRunStore) {
	t.Helper()
	applicationID := "app_bbbbbbbbbbbbbbbb"
	owner := "subject_structured"
	applicationRepository := newMemoryApplicationCatalogRepository()
	applicationService := newApplicationCatalogService(applicationRepository)
	applicationService.newID = func() (string, error) { return applicationID, nil }
	applicationContext := ApplicationCatalogContext{RequestContext: context.Background(), RequestID: "request_structured_app", TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", ActorRef: owner, OwnerSubjectRef: owner, AuditRef: "audit_structured_app", WriteEnabled: true}
	if result := applicationService.Create(applicationContext, ApplicationCatalogCreateInput{DisplayName: "Structured definition app", ApplicationKind: "workflow_copilot"}); result.FailureCode != "" {
		t.Fatalf("create application fixture: %#v", result)
	}
	releaseContext := WorkflowDefinitionReleaseContext{RequestContext: context.Background(), TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", ApplicationID: applicationID, OwnerSubjectRef: owner, ActorRef: owner, RequestID: "request_structured_release", AuditRef: "audit_structured_release"}
	repository := newWorkflowDefinitionReleaseStore()
	candidate, err := repository.CreateCandidate(releaseContext, "candidate_structured_runtime", "definition_structured_runtime", "", executableWorkflowStructuredDraftForTest(applicationID), time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	_, version, err := repository.Review(releaseContext, candidate.CandidateID, 0, "approve", "approve exact structured runtime", candidate.SourceDraftDigest, time.Now().UTC())
	if err != nil || version == nil {
		t.Fatalf("approve structured definition: %#v err=%v", version, err)
	}
	activation, err := repository.DecideActivation(releaseContext, version.DefinitionID, 0, "activate", version.Version, "activate exact structured runtime", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	store := newMemoryWorkflowRunStore(20)
	bridgeClient := &workflowExecutorTestBridge{}
	executor := newWorkflowExecutorService(nil, bridgeClient, store)
	runContext := WorkflowRunContext{RequestContext: context.Background(), RequestID: "request_structured_run", TenantRef: releaseContext.TenantRef, WorkspaceID: releaseContext.WorkspaceID, ApplicationID: applicationID, ActorRef: owner, AuditRef: "audit_structured_run"}
	request := WorkflowDefinitionRunRequest{DefinitionID: version.DefinitionID, ExpectedPointerVersion: activation.PointerVersion, ExpectedDefinitionVersion: version.Version, ExpectedDefinitionDigest: version.DefinitionDigest, Inputs: map[string]any{"customer_name": "Ada"}, ConditionValues: map[string]bool{}}
	return newWorkflowDefinitionExecutionService(repository, applicationRepository, executor), runContext, request, bridgeClient, store
}

func executableWorkflowStructuredDraftForTest(applicationID string) SavedWorkflowDraft {
	draft := executableWorkflowDraftForTest()
	draft.ApplicationID = applicationID
	draft.SchemaVersion = savedWorkflowDraftStructuredSchemaVersion
	draft.InputContract = workflowStructuredInputContractForTest()
	draft.ToolRefs = []string{}
	draft.RAGRefs = []string{}
	draft.RequestedCapabilities = []string{}
	return draft
}

func performWorkflowDefinitionStructuredHTTPRequest(t *testing.T, server *Server, body map[string]any, applicationID string) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-definition-runs", bytes.NewReader(payload))
	setControlPlaneReadDevAuthHeaders(request)
	request.Header.Set(controlPlaneReadDevScopesHeader, "workflow_runs:execute,workflow_definitions:read")
	request.Header.Set(controlPlaneReadDevMembershipPermHeader, "workflow_runs:execute,workflow_definitions:read")
	request.Header.Set(savedWorkflowDraftDevWorkspaceHeader, "workspace_demo")
	request.Header.Set(savedWorkflowDraftDevApplicationHeader, applicationID)
	response := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(response, request)
	return response
}

func cloneStructuredHTTPBody(value map[string]any) map[string]any {
	cloned := make(map[string]any, len(value))
	for key, item := range value {
		cloned[key] = item
	}
	return cloned
}
