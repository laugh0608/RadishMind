package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
)

func TestApplicationEvaluationScheduleHTTPStrictControlPlane(t *testing.T) {
	server := NewServer(config.Config{
		ControlPlaneReadDevAuthEnabled: true, ApplicationCatalogDevHTTPEnabled: true,
		WorkflowRAGEvaluationDevEnabled: true, ApplicationEvaluationCampaignDevEnabled: true,
		ApplicationEvaluationCampaignEnvironment: applicationEvaluationEnvironmentTest, Provider: "mock",
		APIKeyLifecycleDevHTTPEnabled: true, GatewayAuthMode: gatewayAPIKeyAuthenticationSource,
		GatewayRequestHistoryDevEnabled: true, GatewayRequestQuotaEnforcementDevEnabled: true,
		GatewayRequestQuotaEnvironment: applicationEvaluationEnvironmentTest,
	}, Options{BuildVersion: "test"})
	t.Cleanup(server.Close)
	applicationID := "app_aaaaaaaaaaaaaaaa"
	catalog := server.applicationCatalogService()
	catalog.newID = func() (string, error) { return applicationID, nil }
	createdApplication := catalog.Create(ApplicationCatalogContext{
		RequestContext: context.Background(), RequestID: "request-schedule-app", TenantRef: "tenant_demo",
		WorkspaceID: "workspace_demo", ActorRef: "subject_demo_user", OwnerSubjectRef: "subject_demo_user",
		AuditRef: "audit-schedule-app", WriteEnabled: true,
	}, ApplicationCatalogCreateInput{DisplayName: "Scheduled evaluation app", ApplicationKind: "prompt_application"})
	if createdApplication.FailureCode != "" || createdApplication.Record == nil {
		t.Fatalf("create Prompt application: %+v", createdApplication)
	}
	ctx := ApplicationEvaluationContext{
		RequestContext: context.Background(), RequestID: "request-schedule-plan", TenantRef: "tenant_demo",
		WorkspaceID: "workspace_demo", Environment: applicationEvaluationEnvironmentTest, ApplicationID: applicationID,
		ActorRef: "subject_demo_user", AuditRef: "audit-schedule-plan", WriteEnabled: true,
	}
	planService := server.applicationEvaluationPlanService()
	planService.newPlanID = func() (string, error) { return "aeplan_aaaaaaaaaaaaaaaa", nil }
	plan := planService.Create(ctx, applicationEvaluationPromptPlanInput("HTTP scheduled regression", WorkflowRunComparisonUnchanged))
	if plan.FailureCode != "" || plan.Plan == nil || plan.Version == nil {
		t.Fatalf("create Prompt plan: %+v", plan)
	}
	keyService := newAPIKeyService(server.apiKeyRepository, server.applicationCatalogRepository)
	keyService.newID = func() (string, error) { return "key_aaaaaaaaaaaaaaaa", nil }
	issued := keyService.Create(APIKeyContext{
		RequestContext: context.Background(), RequestID: "request-schedule-key", TenantRef: "tenant_demo",
		WorkspaceID: "workspace_demo", ActorRef: "subject_demo_user", OwnerSubjectRef: "subject_demo_user",
		AuditRef: "audit-schedule-key", WriteEnabled: true,
	}, APIKeyCreateInput{
		ApplicationID: applicationID, DisplayName: "Scheduled evaluation quota consumer",
		Scopes: []string{"prompt_application:invoke"}, ExpiresInDays: 30,
	})
	if issued.FailureCode != "" || issued.Record == nil {
		t.Fatalf("issue owned quota API key: %+v", issued)
	}

	createPath := "/v1/user-workspace/applications/" + applicationID + "/evaluation-schedules"
	createBody := applicationEvaluationScheduleCreateBody{
		WorkspaceID: "workspace_demo", Environment: applicationEvaluationEnvironmentTest,
		PlanID: plan.Plan.PlanID, PlanVersion: plan.Version.PlanVersion, PlanDigest: plan.Version.PlanDigest,
		ExpectedPlanRecordVersion: plan.Plan.RecordVersion, QuotaAPIKeyID: issued.Record.APIKeyID,
		Schedule:                       ApplicationEvaluationScheduleDailyUTC{Rule: applicationEvaluationScheduleRuleDailyUTC, Hour: 9, Minute: 30},
		AcknowledgeProviderConsumption: true,
	}
	permissions := "application_evaluations:execute,workflow_runs:execute"
	createdResponse := serveApplicationEvaluationRequest(t, server, http.MethodPost, createPath, createBody, applicationID, permissions)
	created := decodeApplicationEvaluationScheduleHTTPEnvelope(t, createdResponse, http.StatusOK)
	if created.FailureCode != nil || created.Schedule == nil || created.Version == nil ||
		created.Schedule.LifecycleState != applicationEvaluationScheduleStateDraft || created.Version.ScheduleVersion != 1 {
		t.Fatalf("create schedule envelope: %+v", created)
	}

	duplicate := httptest.NewRequest(http.MethodPost, createPath, strings.NewReader(`{
		"workspace_id":"workspace_demo","workspace_id":"workspace_demo","environment":"test",
		"plan_id":"aeplan_aaaaaaaaaaaaaaaa","plan_version":1,"plan_digest":"`+plan.Version.PlanDigest+`",
		"expected_plan_record_version":1,"quota_api_key_id":"key_aaaaaaaaaaaaaaaa",
		"schedule":{"rule":"daily_utc","hour":9,"minute":30},"acknowledge_provider_consumption":true}`))
	setApplicationEvaluationHTTPHeaders(duplicate, applicationID, permissions)
	duplicateResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(duplicateResponse, duplicate)
	if duplicateResponse.Code != http.StatusBadRequest {
		t.Fatalf("duplicate schedule field was accepted: status=%d body=%s", duplicateResponse.Code, duplicateResponse.Body.String())
	}
	unknown := httptest.NewRequest(http.MethodPost, createPath, strings.NewReader(`{"workspace_id":"workspace_demo","environment":"test","credential":"forbidden"}`))
	setApplicationEvaluationHTTPHeaders(unknown, applicationID, permissions)
	unknownResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(unknownResponse, unknown)
	if unknownResponse.Code != http.StatusBadRequest {
		t.Fatalf("unknown schedule field was accepted: status=%d body=%s", unknownResponse.Code, unknownResponse.Body.String())
	}
	denied := serveApplicationEvaluationRequest(t, server, http.MethodPost, createPath, createBody, applicationID, "application_evaluations:execute")
	if denied.Code != http.StatusForbidden {
		t.Fatalf("missing workflow execution permission was accepted: status=%d body=%s", denied.Code, denied.Body.String())
	}

	revisePath := createPath + "/" + created.Schedule.ScheduleID + "/revisions"
	reviseResponse := serveApplicationEvaluationRequest(t, server, http.MethodPost, revisePath, applicationEvaluationScheduleReviseBody{
		WorkspaceID: "workspace_demo", Environment: applicationEvaluationEnvironmentTest, ExpectedVersion: 1,
		PlanID: plan.Plan.PlanID, PlanVersion: plan.Version.PlanVersion, PlanDigest: plan.Version.PlanDigest,
		ExpectedPlanRecordVersion: plan.Plan.RecordVersion, QuotaAPIKeyID: issued.Record.APIKeyID,
		Schedule:                       ApplicationEvaluationScheduleDailyUTC{Rule: applicationEvaluationScheduleRuleDailyUTC, Hour: 10, Minute: 15},
		AcknowledgeProviderConsumption: true,
	}, applicationID, permissions)
	revised := decodeApplicationEvaluationScheduleHTTPEnvelope(t, reviseResponse, http.StatusOK)
	if revised.Schedule == nil || revised.Version == nil || revised.Schedule.RecordVersion != 2 || revised.Version.ScheduleVersion != 2 {
		t.Fatalf("revise schedule envelope: %+v", revised)
	}

	readPath := createPath + "/" + created.Schedule.ScheduleID + "?workspace_id=workspace_demo&environment=test"
	read := decodeApplicationEvaluationScheduleHTTPEnvelope(t,
		serveApplicationEvaluationRequest(t, server, http.MethodGet, readPath, nil, applicationID, "application_evaluations:read"), http.StatusOK)
	if read.Schedule == nil || read.Schedule.RecordVersion != 2 || read.Version != nil {
		t.Fatalf("read current schedule envelope: %+v", read)
	}
	versionPath := createPath + "/" + created.Schedule.ScheduleID + "/versions/1?workspace_id=workspace_demo&environment=test"
	exact := decodeApplicationEvaluationScheduleHTTPEnvelope(t,
		serveApplicationEvaluationRequest(t, server, http.MethodGet, versionPath, nil, applicationID, "application_evaluations:read"), http.StatusOK)
	if exact.Version == nil || exact.Version.ScheduleVersion != 1 || exact.Version.Schedule.Minute != 30 || exact.Schedule != nil {
		t.Fatalf("read exact schedule version envelope: %+v", exact)
	}
	listPath := createPath + "?workspace_id=workspace_demo&environment=test&lifecycle_state=draft&limit=1"
	listed := decodeApplicationEvaluationScheduleListHTTPEnvelope(t,
		serveApplicationEvaluationRequest(t, server, http.MethodGet, listPath, nil, applicationID, "application_evaluations:read"), http.StatusOK)
	if len(listed.Schedules) != 1 || listed.Schedules[0].ScheduleID != created.Schedule.ScheduleID || listed.HasMore {
		t.Fatalf("list schedule envelope: %+v", listed)
	}
	strictQuery := serveApplicationEvaluationRequest(t, server, http.MethodGet, listPath+"&unexpected=true", nil, applicationID, "application_evaluations:read")
	if strictQuery.Code != http.StatusBadRequest {
		t.Fatalf("unknown schedule query was accepted: status=%d body=%s", strictQuery.Code, strictQuery.Body.String())
	}

	current, found, err := server.applicationEvaluationScheduleRepository.ReadSchedule(ctx, created.Schedule.ScheduleID)
	if err != nil || !found {
		t.Fatalf("read schedule before HTTP lifecycle: found=%v err=%v", found, err)
	}
	currentVersion, found, err := server.applicationEvaluationScheduleRepository.ReadScheduleVersion(ctx, current.ScheduleID, current.LatestScheduleVersion)
	if err != nil || !found {
		t.Fatalf("read schedule version before HTTP lifecycle: found=%v err=%v", found, err)
	}
	updatedAt, ok := parseApplicationEvaluationScheduleUTCTimestamp(current.UpdatedAt)
	if !ok {
		t.Fatalf("parse schedule update time: %s", current.UpdatedAt)
	}
	updatedAt = updatedAt.Add(time.Nanosecond)
	nextDue, err := applicationEvaluationScheduleNextDue(updatedAt, currentVersion.Schedule)
	if err != nil {
		t.Fatal(err)
	}
	active := current
	active.RecordVersion++
	active.LifecycleState = applicationEvaluationScheduleStateActive
	active.UpdatedAt = updatedAt.Format(time.RFC3339Nano)
	dueText := nextDue.Format(time.RFC3339Nano)
	active.NextDueAt = &dueText
	active.RequestID, active.AuditRef = ctx.RequestID, ctx.AuditRef
	if _, updated, updateErr := server.applicationEvaluationScheduleRepository.UpdateSchedule(ctx, current.RecordVersion, active); updateErr != nil || !updated {
		t.Fatalf("seed active schedule for HTTP pause: updated=%v err=%v", updated, updateErr)
	}
	pausePath := createPath + "/" + current.ScheduleID + "/pause"
	paused := decodeApplicationEvaluationScheduleHTTPEnvelope(t,
		serveApplicationEvaluationRequest(t, server, http.MethodPost, pausePath, applicationEvaluationScheduleLifecycleBody{
			WorkspaceID: "workspace_demo", Environment: applicationEvaluationEnvironmentTest, ExpectedVersion: active.RecordVersion,
		}, applicationID, permissions), http.StatusOK)
	if paused.Schedule == nil || paused.Schedule.LifecycleState != applicationEvaluationScheduleStatePaused || paused.Schedule.NextDueAt != nil {
		t.Fatalf("pause schedule envelope: %+v", paused)
	}
	resumePath := createPath + "/" + current.ScheduleID + "/resume"
	resumeDenied := decodeApplicationEvaluationScheduleHTTPEnvelope(t,
		serveApplicationEvaluationRequest(t, server, http.MethodPost, resumePath, applicationEvaluationScheduleLifecycleBody{
			WorkspaceID: "workspace_demo", Environment: applicationEvaluationEnvironmentTest, ExpectedVersion: paused.Schedule.RecordVersion,
			AcknowledgeProviderConsumption: true,
		}, applicationID, permissions), http.StatusConflict)
	if resumeDenied.FailureCode == nil || *resumeDenied.FailureCode != ApplicationEvaluationScheduleFailureAuthorityChanged {
		t.Fatalf("resume without current Prompt authority did not fail closed: %+v", resumeDenied)
	}
	archivePath := createPath + "/" + current.ScheduleID + "/archive"
	archived := decodeApplicationEvaluationScheduleHTTPEnvelope(t,
		serveApplicationEvaluationRequest(t, server, http.MethodPost, archivePath, applicationEvaluationScheduleLifecycleBody{
			WorkspaceID: "workspace_demo", Environment: applicationEvaluationEnvironmentTest, ExpectedVersion: paused.Schedule.RecordVersion,
			AcknowledgeNoFutureOccurrences: true,
		}, applicationID, permissions), http.StatusOK)
	if archived.Schedule == nil || archived.Schedule.LifecycleState != applicationEvaluationScheduleStateArchived {
		t.Fatalf("archive schedule envelope: %+v", archived)
	}

	occurrencePath := createPath + "/" + current.ScheduleID + "/occurrences/2/" + dueText + "?workspace_id=workspace_demo&environment=test"
	missingOccurrence := decodeApplicationEvaluationScheduleHTTPEnvelope(t,
		serveApplicationEvaluationRequest(t, server, http.MethodGet, occurrencePath, nil, applicationID, "application_evaluations:read"), http.StatusNotFound)
	if missingOccurrence.FailureCode == nil || *missingOccurrence.FailureCode != ApplicationEvaluationScheduleFailureNotFound {
		t.Fatalf("missing occurrence read did not return stable not-found: %+v", missingOccurrence)
	}
}

func decodeApplicationEvaluationScheduleHTTPEnvelope(
	t *testing.T,
	response *httptest.ResponseRecorder,
	expectedStatus int,
) applicationEvaluationScheduleEnvelope {
	t.Helper()
	if response.Code != expectedStatus {
		t.Fatalf("unexpected schedule status: got=%d want=%d body=%s", response.Code, expectedStatus, response.Body.String())
	}
	var envelope applicationEvaluationScheduleEnvelope
	decoder := json.NewDecoder(response.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&envelope); err != nil {
		t.Fatal(err)
	}
	return envelope
}

func decodeApplicationEvaluationScheduleListHTTPEnvelope(
	t *testing.T,
	response *httptest.ResponseRecorder,
	expectedStatus int,
) applicationEvaluationScheduleListEnvelope {
	t.Helper()
	if response.Code != expectedStatus {
		t.Fatalf("unexpected schedule list status: got=%d want=%d body=%s", response.Code, expectedStatus, response.Body.String())
	}
	var envelope applicationEvaluationScheduleListEnvelope
	decoder := json.NewDecoder(response.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&envelope); err != nil {
		t.Fatal(err)
	}
	return envelope
}
