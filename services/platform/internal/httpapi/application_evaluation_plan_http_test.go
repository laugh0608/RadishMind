package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"radishmind.local/services/platform/internal/config"
)

func TestApplicationEvaluationPlanHTTPStrictLifecycle(t *testing.T) {
	server := NewServer(config.Config{
		ControlPlaneReadDevAuthEnabled: true, ApplicationCatalogDevHTTPEnabled: true,
		WorkflowRAGEvaluationDevEnabled: true, ApplicationEvaluationCampaignDevEnabled: true,
		ApplicationEvaluationCampaignEnvironment: applicationEvaluationEnvironmentTest, Provider: "mock",
		APIKeyLifecycleDevHTTPEnabled: true, GatewayAuthMode: gatewayAPIKeyAuthenticationSource,
		GatewayRequestHistoryDevEnabled: true, GatewayRequestQuotaEnforcementDevEnabled: true, GatewayRequestQuotaEnvironment: applicationEvaluationEnvironmentTest,
	}, Options{BuildVersion: "test"})
	t.Cleanup(server.Close)
	applicationID := "app_aaaaaaaaaaaaaaaa"
	catalog := server.applicationCatalogService()
	catalog.newID = func() (string, error) { return applicationID, nil }
	if created := catalog.Create(ApplicationCatalogContext{
		RequestContext: context.Background(), RequestID: "request-app", TenantRef: "tenant_demo",
		WorkspaceID: "workspace_demo", ActorRef: "subject_demo_user", OwnerSubjectRef: "subject_demo_user",
		AuditRef: "audit-app", WriteEnabled: true,
	}, ApplicationCatalogCreateInput{DisplayName: "Evaluation HTTP app", ApplicationKind: "prompt_application"}); created.FailureCode != "" {
		t.Fatalf("create application: %+v", created)
	}

	createBody := applicationEvaluationPlanCreateBody{
		WorkspaceID: "workspace_demo", Environment: applicationEvaluationEnvironmentTest,
		Name: "HTTP prompt regression", ExecutionProfile: applicationInteractionProfilePrompt,
		Items: []ApplicationEvaluationPlanItem{applicationEvaluationPromptItem("first", "First prompt", WorkflowRunComparisonUnchanged)},
	}
	createResponse := serveApplicationEvaluationRequest(t, server, http.MethodPost,
		"/v1/user-workspace/applications/"+applicationID+"/evaluation-plans", createBody,
		applicationID, "application_evaluations:write")
	createEnvelope := decodeApplicationEvaluationPlanHTTPEnvelope(t, createResponse, http.StatusOK)
	if createEnvelope.FailureCode != nil || createEnvelope.Plan == nil || createEnvelope.Version == nil ||
		createEnvelope.Plan.RecordVersion != 1 || createEnvelope.Version.PlanVersion != 1 {
		t.Fatalf("unexpected create envelope: %+v body=%s", createEnvelope, createResponse.Body.String())
	}

	readResponse := serveApplicationEvaluationRequest(t, server, http.MethodGet,
		"/v1/user-workspace/applications/"+applicationID+"/evaluation-plans/"+createEnvelope.Plan.PlanID+"?workspace_id=workspace_demo&environment=test",
		nil, applicationID, "application_evaluations:read")
	readEnvelope := decodeApplicationEvaluationPlanHTTPEnvelope(t, readResponse, http.StatusOK)
	if readEnvelope.Plan == nil || readEnvelope.Plan.PlanID != createEnvelope.Plan.PlanID || readEnvelope.Version != nil {
		t.Fatalf("unexpected read envelope: %+v", readEnvelope)
	}

	reviseBody := applicationEvaluationPlanReviseBody{
		WorkspaceID: "workspace_demo", Environment: applicationEvaluationEnvironmentTest, ExpectedVersion: 1,
		Name: "HTTP prompt regression revised", ExecutionProfile: applicationInteractionProfilePrompt,
		Items: []ApplicationEvaluationPlanItem{
			applicationEvaluationPromptItem("first", "First prompt", WorkflowRunComparisonChanged),
			applicationEvaluationPromptItem("second", "Second prompt", WorkflowRunComparisonUnchanged),
		},
	}
	revisePath := "/v1/user-workspace/applications/" + applicationID + "/evaluation-plans/" + createEnvelope.Plan.PlanID + "/revisions"
	reviseResponse := serveApplicationEvaluationRequest(t, server, http.MethodPost, revisePath, reviseBody, applicationID, "application_evaluations:write")
	reviseEnvelope := decodeApplicationEvaluationPlanHTTPEnvelope(t, reviseResponse, http.StatusOK)
	if reviseEnvelope.Plan == nil || reviseEnvelope.Version == nil || reviseEnvelope.Plan.RecordVersion != 2 || reviseEnvelope.Version.PlanVersion != 2 {
		t.Fatalf("unexpected revise envelope: %+v", reviseEnvelope)
	}
	staleResponse := serveApplicationEvaluationRequest(t, server, http.MethodPost, revisePath, reviseBody, applicationID, "application_evaluations:write")
	staleEnvelope := decodeApplicationEvaluationPlanHTTPEnvelope(t, staleResponse, http.StatusConflict)
	if staleEnvelope.FailureCode == nil || *staleEnvelope.FailureCode != ApplicationEvaluationFailureVersionConflict || staleEnvelope.CurrentRecordVersion != 2 {
		t.Fatalf("unexpected stale envelope: %+v", staleEnvelope)
	}

	archiveBody := applicationEvaluationPlanArchiveBody{
		WorkspaceID: "workspace_demo", Environment: applicationEvaluationEnvironmentTest,
		ExpectedVersion: 2, AcknowledgeNoNewCampaigns: true,
	}
	archiveResponse := serveApplicationEvaluationRequest(t, server, http.MethodPost,
		"/v1/user-workspace/applications/"+applicationID+"/evaluation-plans/"+createEnvelope.Plan.PlanID+"/archive",
		archiveBody, applicationID, "application_evaluations:write")
	archiveEnvelope := decodeApplicationEvaluationPlanHTTPEnvelope(t, archiveResponse, http.StatusOK)
	if archiveEnvelope.Plan == nil || archiveEnvelope.Plan.RecordVersion != 3 || archiveEnvelope.Plan.LifecycleState != applicationEvaluationPlanStateArchived {
		t.Fatalf("unexpected archive envelope: %+v", archiveEnvelope)
	}
}

func TestApplicationEvaluationPlanHTTPPermissionEnvironmentAndUnknownFieldClose(t *testing.T) {
	server := NewServer(config.Config{
		ControlPlaneReadDevAuthEnabled: true, ApplicationCatalogDevHTTPEnabled: true,
		WorkflowRAGEvaluationDevEnabled: true, ApplicationEvaluationCampaignDevEnabled: true,
		ApplicationEvaluationCampaignEnvironment: applicationEvaluationEnvironmentTest, Provider: "mock",
		APIKeyLifecycleDevHTTPEnabled: true, GatewayAuthMode: gatewayAPIKeyAuthenticationSource,
		GatewayRequestHistoryDevEnabled: true, GatewayRequestQuotaEnforcementDevEnabled: true, GatewayRequestQuotaEnvironment: applicationEvaluationEnvironmentTest,
	}, Options{BuildVersion: "test"})
	t.Cleanup(server.Close)
	applicationID := "app_aaaaaaaaaaaaaaaa"
	body := `{"workspace_id":"workspace_demo","environment":"test","name":"Unknown field","execution_profile":"prompt_application_invocation_v1","target":{"workflow_definition":null},"items":[],"credential":"forbidden"}`
	unknown := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/applications/"+applicationID+"/evaluation-plans", strings.NewReader(body))
	setApplicationEvaluationHTTPHeaders(unknown, applicationID, "application_evaluations:write")
	unknownResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(unknownResponse, unknown)
	if unknownResponse.Code != http.StatusBadRequest {
		t.Fatalf("unknown field was accepted: status=%d body=%s", unknownResponse.Code, unknownResponse.Body.String())
	}

	validBody := applicationEvaluationPlanCreateBody{
		WorkspaceID: "workspace_demo", Environment: applicationEvaluationEnvironmentTest,
		Name: "Permission denied", ExecutionProfile: applicationInteractionProfilePrompt,
		Items: []ApplicationEvaluationPlanItem{applicationEvaluationPromptItem("first", "First prompt", WorkflowRunComparisonUnchanged)},
	}
	denied := serveApplicationEvaluationRequest(t, server, http.MethodPost,
		"/v1/user-workspace/applications/"+applicationID+"/evaluation-plans", validBody,
		applicationID, "application_evaluations:read")
	if denied.Code != http.StatusForbidden {
		t.Fatalf("missing write permission was accepted: status=%d body=%s", denied.Code, denied.Body.String())
	}

	environmentPayload, _ := json.Marshal(validBody)
	environmentRequest := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/applications/"+applicationID+"/evaluation-plans", bytes.NewReader(environmentPayload))
	setApplicationEvaluationHTTPHeaders(environmentRequest, applicationID, "application_evaluations:write")
	environmentRequest.Header.Set(applicationEvaluationEnvironmentHeader, applicationEvaluationEnvironmentDevelopment)
	environmentResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(environmentResponse, environmentRequest)
	if environmentResponse.Code != http.StatusForbidden {
		t.Fatalf("environment mismatch was accepted: status=%d body=%s", environmentResponse.Code, environmentResponse.Body.String())
	}
}

func serveApplicationEvaluationRequest(t *testing.T, server *Server, method, path string, body any, applicationID string, permissions string) *httptest.ResponseRecorder {
	t.Helper()
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	request := httptest.NewRequest(method, path, bytes.NewReader(payload))
	setApplicationEvaluationHTTPHeaders(request, applicationID, permissions)
	response := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(response, request)
	return response
}

func setApplicationEvaluationHTTPHeaders(request *http.Request, applicationID, permissions string) {
	request.Header.Set(controlPlaneReadDevIdentityHeader, "application-evaluation-test")
	request.Header.Set(controlPlaneReadDevTenantHeader, "tenant_demo")
	request.Header.Set(controlPlaneReadDevSubjectHeader, "subject_demo_user")
	request.Header.Set(controlPlaneReadDevScopesHeader, permissions)
	request.Header.Set(controlPlaneReadDevAuditHeader, "audit-application-evaluation")
	request.Header.Set(activeWorkspaceHeader, "workspace_demo")
	request.Header.Set(controlPlaneReadDevMembershipHeader, "workspace_demo")
	request.Header.Set(controlPlaneReadDevMembershipPermHeader, permissions)
	request.Header.Set(savedWorkflowDraftDevWorkspaceHeader, "workspace_demo")
	request.Header.Set(savedWorkflowDraftDevApplicationHeader, applicationID)
	request.Header.Set(applicationEvaluationEnvironmentHeader, applicationEvaluationEnvironmentTest)
}

func decodeApplicationEvaluationPlanHTTPEnvelope(t *testing.T, response *httptest.ResponseRecorder, expectedStatus int) applicationEvaluationPlanEnvelope {
	t.Helper()
	if response.Code != expectedStatus {
		t.Fatalf("unexpected status: got=%d want=%d body=%s", response.Code, expectedStatus, response.Body.String())
	}
	var envelope applicationEvaluationPlanEnvelope
	decoder := json.NewDecoder(response.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&envelope); err != nil {
		t.Fatal(err)
	}
	return envelope
}
