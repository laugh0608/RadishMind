package httpapi

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
)

func TestApplicationResultArtifactApplicationLibraryFiltersPagesAndBindsCursor(t *testing.T) {
	repository := newMemoryApplicationResultArtifactRepository()
	ctx, artifacts := applicationResultArtifactPersistenceProfileFixtures()
	artifacts[len(artifacts)-1].ContentType = "application/json"
	artifacts[len(artifacts)-1].Content = `{"result":"agent copilot"}`
	artifacts[len(artifacts)-1].ContentBytes = len([]byte(artifacts[len(artifacts)-1].Content))
	artifacts[len(artifacts)-1].ContentDigest = applicationResultArtifactContentDigest(
		artifacts[len(artifacts)-1].ContentType,
		artifacts[len(artifacts)-1].Content,
	)
	for _, artifact := range artifacts {
		if _, replay, err := repository.Create(ctx, artifact); err != nil || replay {
			t.Fatalf("create application library fixture %s: replay=%v err=%v", artifact.ArtifactID, replay, err)
		}
	}
	service := newApplicationResultArtifactService(repository)
	service.now = func() time.Time { return time.Date(2026, 8, 17, 10, 0, 0, 123456789, time.UTC) }
	archived := service.Archive(ctx, ApplicationResultArtifactLifecycleTransitionInput{
		SessionID: artifacts[2].SessionID, ArtifactID: artifacts[2].ArtifactID, ExpectedLifecycleVersion: 1,
	})
	if archived.FailureCode != "" {
		t.Fatalf("archive application library fixture: %#v", archived)
	}

	first := service.ListApplication(ctx, ApplicationResultArtifactListInput{Limit: 2})
	if first.FailureCode != "" || len(first.Items) != 2 || first.NextCursor == nil ||
		first.Items[0].ArtifactID != artifacts[4].ArtifactID || first.Items[1].ArtifactID != artifacts[3].ArtifactID {
		t.Fatalf("application library first page drifted: %#v", first)
	}
	second := service.ListApplication(ctx, ApplicationResultArtifactListInput{Limit: 2, Cursor: *first.NextCursor})
	if second.FailureCode != "" || len(second.Items) != 2 || second.NextCursor != nil ||
		second.Items[0].ArtifactID != artifacts[1].ArtifactID || second.Items[1].ArtifactID != artifacts[0].ArtifactID {
		t.Fatalf("application library second page drifted: %#v", second)
	}

	profile := service.ListApplication(ctx, ApplicationResultArtifactListInput{
		ExecutionProfile: applicationInteractionProfilePrompt,
	})
	if profile.FailureCode != "" || len(profile.Items) != 1 || profile.Items[0].ArtifactID != artifacts[3].ArtifactID {
		t.Fatalf("application library profile filter drifted: %#v", profile)
	}
	jsonResults := service.ListApplication(ctx, ApplicationResultArtifactListInput{ContentType: "application/json"})
	if jsonResults.FailureCode != "" || len(jsonResults.Items) != 1 || jsonResults.Items[0].ArtifactID != artifacts[4].ArtifactID {
		t.Fatalf("application library content type filter drifted: %#v", jsonResults)
	}
	archivedResults := service.ListApplication(ctx, ApplicationResultArtifactListInput{
		LifecycleState: ApplicationResultArtifactLifecycleArchived,
	})
	if archivedResults.FailureCode != "" || len(archivedResults.Items) != 1 ||
		archivedResults.Items[0].ArtifactID != artifacts[2].ArtifactID || archivedResults.Items[0].LifecycleVersion != 2 {
		t.Fatalf("application library lifecycle filter drifted: %#v", archivedResults)
	}

	drifted := service.ListApplication(ctx, ApplicationResultArtifactListInput{
		Limit: 2, Cursor: *first.NextCursor, ContentType: "application/json",
	})
	if drifted.FailureCode != ApplicationResultArtifactFailurePayloadInvalid || len(drifted.Items) != 0 {
		t.Fatalf("application library cursor filter drift did not fail closed: %#v", drifted)
	}
	cursorPayload, err := base64.RawURLEncoding.DecodeString(*first.NextCursor)
	if err != nil {
		t.Fatalf("decode application library cursor fixture: %v", err)
	}
	trailingCursor := base64.RawURLEncoding.EncodeToString(append(cursorPayload, []byte(`{}`)...))
	if trailing := service.ListApplication(ctx, ApplicationResultArtifactListInput{Limit: 2, Cursor: trailingCursor}); trailing.FailureCode != ApplicationResultArtifactFailurePayloadInvalid {
		t.Fatalf("application library cursor accepted trailing JSON: %#v", trailing)
	}
	if invalid := service.ListApplication(ctx, ApplicationResultArtifactListInput{ExecutionProfile: "unknown"}); invalid.FailureCode != ApplicationResultArtifactFailurePayloadInvalid {
		t.Fatalf("application library accepted unknown profile: %#v", invalid)
	}
	if sessionFilter := service.List(ctx, ApplicationResultArtifactListInput{
		SessionID: artifacts[0].SessionID, ContentType: "text/markdown",
	}); sessionFilter.FailureCode != ApplicationResultArtifactFailurePayloadInvalid {
		t.Fatalf("session list unexpectedly accepted application-only filters: %#v", sessionFilter)
	}
	otherApplication := ctx
	otherApplication.ApplicationID = "application_other"
	if other := service.ListApplication(otherApplication, ApplicationResultArtifactListInput{}); other.FailureCode != "" || len(other.Items) != 0 {
		t.Fatalf("application library leaked across application scope: %#v", other)
	}
}

func TestApplicationResultArtifactExportIsCanonicalAndHasNoRepositorySideEffect(t *testing.T) {
	repository := newMemoryApplicationResultArtifactRepository()
	ctx, artifact := applicationResultArtifactPersistenceFixture()
	if _, _, err := repository.Create(ctx, artifact); err != nil {
		t.Fatalf("create export fixture: %v", err)
	}
	service := newApplicationResultArtifactService(repository)
	service.now = func() time.Time { return time.Date(2026, 8, 17, 10, 15, 0, 987654321, time.UTC) }

	result := service.Export(ctx, artifact.ArtifactID)
	if result.FailureCode != "" || result.Export == nil ||
		result.Export.SchemaVersion != applicationResultArtifactExportSchemaVersion ||
		result.Export.Artifact.Content != artifact.Content ||
		result.Export.Lifecycle.LifecycleState != ApplicationResultArtifactLifecycleActive ||
		result.Export.ExportedAt != "2026-08-17T10:15:00.987654321Z" ||
		result.Export.ExportedByActorRef != ctx.ActorRef ||
		result.Export.ExportDigest != applicationResultArtifactExportDigest(*result.Export) ||
		validateApplicationResultArtifactExport(ctx, *result.Export) != nil {
		t.Fatalf("canonical artifact export drifted: %#v", result)
	}
	result.Export.Artifact.Content = "mutated export copy"
	restored, err := repository.Read(ctx, artifact.ArtifactID)
	if err != nil || restored.Content != artifact.Content {
		t.Fatalf("export mutated repository artifact: restored=%#v err=%v", restored, err)
	}
	lifecycle, err := repository.ReadLifecycle(ctx, artifact.ArtifactID)
	if err != nil || lifecycle.LifecycleVersion != 1 || lifecycle.LifecycleState != ApplicationResultArtifactLifecycleActive {
		t.Fatalf("export mutated repository lifecycle: lifecycle=%#v err=%v", lifecycle, err)
	}

	otherApplication := ctx
	otherApplication.ApplicationID = "application_other"
	if missing := service.Export(otherApplication, artifact.ArtifactID); missing.FailureCode != ApplicationResultArtifactFailureNotFound {
		t.Fatalf("export leaked across application scope: %#v", missing)
	}
	repository.lifecycleByID[artifact.ArtifactID] = ApplicationResultArtifactLifecycle{
		SchemaVersion: applicationResultArtifactLifecycleSchemaVersion,
		TenantRef:     ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		OwnerSubjectRef: ctx.OwnerSubjectRef, ArtifactID: artifact.ArtifactID,
		LifecycleState: ApplicationResultArtifactLifecycleActive,
	}
	if corrupt := service.Export(ctx, artifact.ArtifactID); corrupt.FailureCode != ApplicationResultArtifactFailureStoreContract {
		t.Fatalf("export accepted corrupt lifecycle: %#v", corrupt)
	}
}

func TestApplicationResultArtifactApplicationLibraryPaginatesMoreThanMaximumWithTimestampTieBreak(t *testing.T) {
	repository := newMemoryApplicationResultArtifactRepository()
	ctx, template := applicationResultArtifactPersistenceFixture()
	alphabet := "abcdefghijklmnopqrstuvwxyz234567"
	artifactIDs := make([]string, 0, 101)
	for index := range 101 {
		suffix := strings.Repeat("a", 14) + string(alphabet[index/32]) + string(alphabet[index%32])
		artifact := template
		artifact.ArtifactID = "appres_" + suffix
		artifact.SessionID = "appsess_" + suffix
		artifact.TurnID = "appturn_" + suffix
		artifact.ClientTurnKey = "application_library_" + suffix
		artifact.RunRef.RunID = "run_" + suffix
		artifact.Content = "application library pagination " + suffix
		artifact.ContentBytes = len([]byte(artifact.Content))
		artifact.ContentDigest = applicationResultArtifactContentDigest(artifact.ContentType, artifact.Content)
		if _, replay, err := repository.Create(ctx, artifact); err != nil || replay {
			t.Fatalf("create application library pagination fixture %d: replay=%v err=%v", index, replay, err)
		}
		artifactIDs = append(artifactIDs, artifact.ArtifactID)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(artifactIDs)))
	service := newApplicationResultArtifactService(repository)
	first := service.ListApplication(ctx, ApplicationResultArtifactListInput{Limit: 100})
	if first.FailureCode != "" || len(first.Items) != 100 || first.NextCursor == nil ||
		first.Items[0].ArtifactID != artifactIDs[0] || first.Items[99].ArtifactID != artifactIDs[99] {
		t.Fatalf("application library maximum page drifted: first=%#v last=%#v failure=%s cursor=%v", first.Items[0], first.Items[len(first.Items)-1], first.FailureCode, first.NextCursor)
	}
	second := service.ListApplication(ctx, ApplicationResultArtifactListInput{Limit: 100, Cursor: *first.NextCursor})
	if second.FailureCode != "" || len(second.Items) != 1 || second.NextCursor != nil || second.Items[0].ArtifactID != artifactIDs[100] {
		t.Fatalf("application library timestamp tie-break page drifted: %#v", second)
	}
}

func TestApplicationResultArtifactApplicationLibraryAndExportHTTPBoundaries(t *testing.T) {
	_, runContext, _, _, _ := workflowDefinitionExecutionFixture(t)
	ctx, artifacts := applicationResultArtifactPersistenceProfileFixtures()
	ctx.TenantRef = runContext.TenantRef
	ctx.WorkspaceID = runContext.WorkspaceID
	ctx.ApplicationID = runContext.ApplicationID
	ctx.OwnerSubjectRef = runContext.ActorRef
	ctx.ActorRef = runContext.ActorRef
	ctx.RequestID = "request_result_library_http"
	ctx.AuditRef = "audit_result_library_http"
	repository := newMemoryApplicationResultArtifactRepository()
	for index := range artifacts[:2] {
		artifacts[index].TenantRef = ctx.TenantRef
		artifacts[index].WorkspaceID = ctx.WorkspaceID
		artifacts[index].ApplicationID = ctx.ApplicationID
		artifacts[index].OwnerSubjectRef = ctx.OwnerSubjectRef
		artifacts[index].CreatedByActorRef = ctx.ActorRef
		artifacts[index].RequestID = ctx.RequestID
		artifacts[index].AuditRef = ctx.AuditRef
		if _, _, err := repository.Create(ctx, artifacts[index]); err != nil {
			t.Fatalf("create HTTP application library fixture: %v", err)
		}
	}
	server := &Server{
		config:                              config.Config{ApplicationSessionDevEnabled: true},
		applicationResultArtifactRepository: repository,
		workspaceMembershipProvider:         newDeterministicDevTestWorkspaceMembershipProvider(),
	}
	auth := applicationInteractionSessionHTTPAuth(
		runContext, "application_sessions:read", "application_result_artifacts:export",
	)
	request := func(method, target string, authContext controlPlaneReadAuthContext) *http.Request {
		req := httptest.NewRequest(method, target, nil)
		req.Header.Set(activeWorkspaceHeader, runContext.WorkspaceID)
		return req.WithContext(withControlPlaneReadFakeAuthContext(req.Context(), authContext))
	}

	listTarget := "/v1/user-workspace/applications/" + runContext.ApplicationID + "/result-artifacts?workspace_id=" + runContext.WorkspaceID + "&execution_profile=" + artifacts[1].ExecutionProfile
	listRequest := request(http.MethodGet, listTarget, auth)
	listRequest.SetPathValue("application_id", runContext.ApplicationID)
	listResponse := httptest.NewRecorder()
	server.handleListApplicationResultArtifactsByApplication(listResponse, listRequest)
	var listed applicationResultArtifactApplicationListEnvelope
	if listResponse.Code != http.StatusOK || json.Unmarshal(listResponse.Body.Bytes(), &listed) != nil ||
		listed.FailureCode != nil || len(listed.Items) != 1 || listed.Items[0].ArtifactID != artifacts[1].ArtifactID ||
		strings.Contains(listResponse.Body.String(), artifacts[1].Content) || listResponse.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("application library HTTP list drifted: status=%d body=%s headers=%v", listResponse.Code, listResponse.Body.String(), listResponse.Header())
	}

	exportTarget := "/v1/user-workspace/applications/" + runContext.ApplicationID + "/result-artifacts/" + artifacts[0].ArtifactID + "/export?workspace_id=" + runContext.WorkspaceID
	exportRequest := request(http.MethodGet, exportTarget, auth)
	exportRequest.SetPathValue("application_id", runContext.ApplicationID)
	exportRequest.SetPathValue("artifact_id", artifacts[0].ArtifactID)
	exportResponse := httptest.NewRecorder()
	server.handleExportApplicationResultArtifact(exportResponse, exportRequest)
	var exported applicationResultArtifactExportEnvelope
	if exportResponse.Code != http.StatusOK || json.Unmarshal(exportResponse.Body.Bytes(), &exported) != nil ||
		exported.FailureCode != nil || exported.Export == nil || exported.Export.Artifact.Content != artifacts[0].Content ||
		exported.Export.ExportDigest != applicationResultArtifactExportDigest(*exported.Export) ||
		exportResponse.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("artifact export HTTP drifted: status=%d body=%s headers=%v", exportResponse.Code, exportResponse.Body.String(), exportResponse.Header())
	}

	deniedRequest := request(http.MethodGet, exportTarget, applicationInteractionSessionHTTPAuth(runContext, "application_sessions:read"))
	deniedRequest.SetPathValue("application_id", runContext.ApplicationID)
	deniedRequest.SetPathValue("artifact_id", artifacts[0].ArtifactID)
	deniedResponse := httptest.NewRecorder()
	server.handleExportApplicationResultArtifact(deniedResponse, deniedRequest)
	if deniedResponse.Code != http.StatusForbidden || !strings.Contains(deniedResponse.Body.String(), "scope_denied") {
		t.Fatalf("artifact export permission did not fail closed: status=%d body=%s", deniedResponse.Code, deniedResponse.Body.String())
	}

	unknownRequest := request(http.MethodGet, listTarget+"&unknown=value", auth)
	unknownRequest.SetPathValue("application_id", runContext.ApplicationID)
	unknownResponse := httptest.NewRecorder()
	server.handleListApplicationResultArtifactsByApplication(unknownResponse, unknownRequest)
	if unknownResponse.Code != http.StatusBadRequest || !strings.Contains(unknownResponse.Body.String(), ApplicationResultArtifactFailurePayloadInvalid) {
		t.Fatalf("application library accepted unknown query: status=%d body=%s", unknownResponse.Code, unknownResponse.Body.String())
	}
	duplicateRequest := request(http.MethodGet, listTarget+"&workspace_id="+runContext.WorkspaceID, auth)
	duplicateRequest.SetPathValue("application_id", runContext.ApplicationID)
	duplicateResponse := httptest.NewRecorder()
	server.handleListApplicationResultArtifactsByApplication(duplicateResponse, duplicateRequest)
	if duplicateResponse.Code != http.StatusBadRequest || !strings.Contains(duplicateResponse.Body.String(), ApplicationResultArtifactFailurePayloadInvalid) {
		t.Fatalf("application library accepted duplicate query: status=%d body=%s", duplicateResponse.Code, duplicateResponse.Body.String())
	}
	exportUnknownRequest := request(http.MethodGet, exportTarget+"&unknown=value", auth)
	exportUnknownRequest.SetPathValue("application_id", runContext.ApplicationID)
	exportUnknownRequest.SetPathValue("artifact_id", artifacts[0].ArtifactID)
	exportUnknownResponse := httptest.NewRecorder()
	server.handleExportApplicationResultArtifact(exportUnknownResponse, exportUnknownRequest)
	if exportUnknownResponse.Code != http.StatusBadRequest || !strings.Contains(exportUnknownResponse.Body.String(), ApplicationResultArtifactFailurePayloadInvalid) {
		t.Fatalf("artifact export accepted unknown query: status=%d body=%s", exportUnknownResponse.Code, exportUnknownResponse.Body.String())
	}

	otherApplication := "application_other"
	crossTarget := "/v1/user-workspace/applications/" + otherApplication + "/result-artifacts/" + artifacts[0].ArtifactID + "/export?workspace_id=" + runContext.WorkspaceID
	crossRequest := request(http.MethodGet, crossTarget, auth)
	crossRequest.SetPathValue("application_id", otherApplication)
	crossRequest.SetPathValue("artifact_id", artifacts[0].ArtifactID)
	crossResponse := httptest.NewRecorder()
	server.handleExportApplicationResultArtifact(crossResponse, crossRequest)
	if crossResponse.Code != http.StatusNotFound || !strings.Contains(crossResponse.Body.String(), ApplicationResultArtifactFailureNotFound) {
		t.Fatalf("artifact export leaked across application: status=%d body=%s", crossResponse.Code, crossResponse.Body.String())
	}

	mux := http.NewServeMux()
	mux.HandleFunc(applicationResultArtifactApplicationListRoute, server.handleListApplicationResultArtifactsByApplication)
	mux.HandleFunc(applicationResultArtifactExportRoute, server.handleExportApplicationResultArtifact)
	deleteResponse := httptest.NewRecorder()
	mux.ServeHTTP(deleteResponse, request(http.MethodDelete, exportTarget, auth))
	if deleteResponse.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected permanent delete route: status=%d body=%s", deleteResponse.Code, deleteResponse.Body.String())
	}
}
