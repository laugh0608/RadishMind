package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"radishmind.local/services/platform/internal/config"
)

func TestApplicationEvaluationQuotaConsumerRequiresActiveOwnedApplicationKey(t *testing.T) {
	applicationRepository := newMemoryApplicationCatalogRepository()
	apiKeys := newMemoryAPIKeyRepository()
	managementContext := APIKeyContext{
		RequestContext: context.Background(), RequestID: "request-key", TenantRef: "tenant-one", WorkspaceID: "workspace-one",
		ActorRef: "subject-owner", OwnerSubjectRef: "subject-owner", AuditRef: "audit-key", WriteEnabled: true,
	}
	seedAPIKeyTestApplication(t, applicationRepository, managementContext, "app_aaaaaaaaaaaaaaaa", applicationCatalogLifecycleActive)
	keyService := newAPIKeyService(apiKeys, applicationRepository)
	keyService.newID = func() (string, error) { return "key_aaaaaaaaaaaaaaaa", nil }
	issued := keyService.Create(managementContext, APIKeyCreateInput{
		ApplicationID: "app_aaaaaaaaaaaaaaaa", DisplayName: "Evaluation quota consumer", Scopes: []string{"prompt_application:invoke"}, ExpiresInDays: 30,
	})
	if issued.FailureCode != "" || issued.Record == nil {
		t.Fatalf("issue evaluation quota consumer: %+v", issued)
	}
	server := &Server{apiKeyRepository: apiKeys}
	ctx := applicationEvaluationTestContext()
	if failure := server.validateApplicationEvaluationQuotaConsumer(ctx, issued.Record.APIKeyID); failure != "" {
		t.Fatalf("active owned key was rejected: %s", failure)
	}
	otherOwner := ctx
	otherOwner.ActorRef = "subject-other"
	if failure := server.validateApplicationEvaluationQuotaConsumer(otherOwner, issued.Record.APIKeyID); failure != ApplicationEvaluationFailureQuotaConsumerInvalid {
		t.Fatalf("cross-owner key did not fail closed: %s", failure)
	}
	apiKeys.unavailable = true
	if failure := server.validateApplicationEvaluationQuotaConsumer(ctx, issued.Record.APIKeyID); failure != ApplicationEvaluationFailureStoreUnavailable {
		t.Fatalf("quota consumer store failure was hidden: %s", failure)
	}
}

func TestApplicationEvaluationPairHTTPStrictPermissionsAndMissingPair(t *testing.T) {
	server := NewServer(config.Config{
		ControlPlaneReadDevAuthEnabled: true, ApplicationCatalogDevHTTPEnabled: true,
		WorkflowRAGEvaluationDevEnabled: true, ApplicationEvaluationCampaignDevEnabled: true,
		ApplicationEvaluationCampaignEnvironment: applicationEvaluationEnvironmentTest, Provider: "mock",
		APIKeyLifecycleDevHTTPEnabled: true, GatewayAuthMode: gatewayAPIKeyAuthenticationSource,
		GatewayRequestHistoryDevEnabled: true, GatewayRequestQuotaEnforcementDevEnabled: true, GatewayRequestQuotaEnvironment: applicationEvaluationEnvironmentTest,
	}, Options{BuildVersion: "test"})
	t.Cleanup(server.Close)
	applicationID := "app_aaaaaaaaaaaaaaaa"

	unknown := httptest.NewRequest(http.MethodPost,
		"/v1/user-workspace/applications/"+applicationID+"/evaluation-campaign-pairs/preview",
		strings.NewReader(`{"workspace_id":"workspace_demo","environment":"test","baseline_campaign_id":"aecamp_aaaaaaaaaaaaaaaa","candidate_campaign_id":"aecamp_bbbbbbbbbbbbbbbb","credential":"forbidden"}`))
	setApplicationEvaluationHTTPHeaders(unknown, applicationID, "application_evaluations:read,workflow_runs:read")
	unknownResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(unknownResponse, unknown)
	if unknownResponse.Code != http.StatusBadRequest {
		t.Fatalf("pair preview accepted an unknown field: status=%d body=%s", unknownResponse.Code, unknownResponse.Body.String())
	}

	body := applicationEvaluationPairPreviewBody{
		WorkspaceID: "workspace_demo", Environment: applicationEvaluationEnvironmentTest,
		BaselineCampaignID: "aecamp_aaaaaaaaaaaaaaaa", CandidateCampaignID: "aecamp_bbbbbbbbbbbbbbbb",
	}
	missing := serveApplicationEvaluationRequest(t, server, http.MethodPost,
		"/v1/user-workspace/applications/"+applicationID+"/evaluation-campaign-pairs/preview",
		body, applicationID, "application_evaluations:read,workflow_runs:read")
	var missingEnvelope applicationEvaluationPairEnvelope
	decodeStrictHTTPEnvelope(t, missing, http.StatusNotFound, &missingEnvelope)
	if missingEnvelope.FailureCode == nil || *missingEnvelope.FailureCode != ApplicationEvaluationFailureNotFound {
		t.Fatalf("missing pair did not fail closed: %+v", missingEnvelope)
	}

	handoff := applicationEvaluationHandoffBody{
		WorkspaceID: "workspace_demo", Environment: applicationEvaluationEnvironmentTest,
		BaselineCampaignID: "aecamp_aaaaaaaaaaaaaaaa", CandidateCampaignID: "aecamp_bbbbbbbbbbbbbbbb",
		ExpectedBaselineCampaignRecordVersion: 1, ExpectedCandidateCampaignRecordVersion: 1,
		AcknowledgeEvidenceMaterializing: true,
	}
	denied := serveApplicationEvaluationRequest(t, server, http.MethodPost,
		"/v1/user-workspace/applications/"+applicationID+"/evaluation-campaign-pairs/handoff",
		handoff, applicationID, "application_evaluations:read,workflow_runs:read")
	if denied.Code != http.StatusForbidden {
		t.Fatalf("handoff without workflow evaluation permission was accepted: status=%d body=%s", denied.Code, denied.Body.String())
	}
}

func TestApplicationEvaluationCampaignHTTPRequiresDefinitionPermissionAndQuotaConsumer(t *testing.T) {
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
	createdApplication := catalog.Create(ApplicationCatalogContext{
		RequestContext: context.Background(), RequestID: "request-app", TenantRef: "tenant_demo", WorkspaceID: "workspace_demo",
		ActorRef: "subject_demo_user", OwnerSubjectRef: "subject_demo_user", AuditRef: "audit-app", WriteEnabled: true,
	}, ApplicationCatalogCreateInput{DisplayName: "Evaluation definition app", ApplicationKind: "workflow_copilot"})
	if createdApplication.FailureCode != "" {
		t.Fatalf("create evaluation application: %+v", createdApplication)
	}
	ctx := ApplicationEvaluationContext{
		RequestContext: context.Background(), RequestID: "request-plan", TenantRef: "tenant_demo", WorkspaceID: "workspace_demo",
		Environment: applicationEvaluationEnvironmentTest, ApplicationID: applicationID, ActorRef: "subject_demo_user", AuditRef: "audit-plan", WriteEnabled: true,
	}
	plan := server.applicationEvaluationPlanService().Create(ctx, applicationEvaluationWorkflowPlanInput("HTTP definition campaign"))
	if plan.FailureCode != "" || plan.Plan == nil || plan.Version == nil {
		t.Fatalf("create evaluation plan: %+v", plan)
	}
	body := applicationEvaluationCampaignExecuteBody{
		WorkspaceID: "workspace_demo", Environment: applicationEvaluationEnvironmentTest,
		PlanID: plan.Plan.PlanID, PlanVersion: plan.Version.PlanVersion, PlanDigest: plan.Version.PlanDigest,
		ExpectedPlanRecordVersion: plan.Plan.RecordVersion, ClientCampaignKey: "campaign_http_definition", QuotaAPIKeyID: "key_aaaaaaaaaaaaaaaa",
		AcknowledgeSequentialExecution: true, AcknowledgeQuotaConsumption: true,
	}
	path := "/v1/user-workspace/applications/" + applicationID + "/evaluation-campaigns"
	missingDefinitionPermission := serveApplicationEvaluationRequest(t, server, http.MethodPost, path, body, applicationID,
		"application_evaluations:execute,workflow_runs:execute")
	if missingDefinitionPermission.Code != http.StatusForbidden {
		t.Fatalf("definition campaign without definition read permission was accepted: status=%d body=%s", missingDefinitionPermission.Code, missingDefinitionPermission.Body.String())
	}
	invalidQuotaConsumer := serveApplicationEvaluationRequest(t, server, http.MethodPost, path, body, applicationID,
		"application_evaluations:execute,workflow_runs:execute,workflow_definitions:read")
	var invalidEnvelope applicationEvaluationCampaignEnvelope
	decodeStrictHTTPEnvelope(t, invalidQuotaConsumer, http.StatusForbidden, &invalidEnvelope)
	if invalidEnvelope.FailureCode == nil || *invalidEnvelope.FailureCode != ApplicationEvaluationFailureQuotaConsumerInvalid {
		t.Fatalf("invalid quota consumer did not fail closed: %+v", invalidEnvelope)
	}
	unknown := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"workspace_id":"workspace_demo","environment":"test","unknown":true}`))
	setApplicationEvaluationHTTPHeaders(unknown, applicationID, "application_evaluations:execute,workflow_runs:execute,workflow_definitions:read")
	unknownResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(unknownResponse, unknown)
	if unknownResponse.Code != http.StatusBadRequest {
		t.Fatalf("campaign execute accepted unknown field: status=%d body=%s", unknownResponse.Code, unknownResponse.Body.String())
	}
}

func decodeStrictHTTPEnvelope(t *testing.T, response *httptest.ResponseRecorder, expectedStatus int, target any) {
	t.Helper()
	if response.Code != expectedStatus {
		t.Fatalf("unexpected status: got=%d want=%d body=%s", response.Code, expectedStatus, response.Body.String())
	}
	decoder := json.NewDecoder(response.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		t.Fatal(err)
	}
}
