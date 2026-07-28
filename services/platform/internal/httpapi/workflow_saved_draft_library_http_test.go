package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSavedWorkflowDraftLibraryHTTPListPaginationAndFilters(t *testing.T) {
	server := newSavedWorkflowDraftHTTPTestServer(true)
	t.Cleanup(server.Close)
	service := server.savedWorkflowDraftService()
	requestContext := savedWorkflowDraftTestContext()

	sourcePayload := validSavedWorkflowDraftPayload()
	sourcePayload.DraftID = "draft_library_http_source"
	sourcePayload.Name = "Alpha source"
	source := service.SaveDraft(requestContext, SaveWorkflowDraftRequest{Payload: sourcePayload})
	if source.FailureCode != "" || source.Draft == nil {
		t.Fatalf("save source: %#v", source)
	}

	definitionPayload := validSavedWorkflowDraftPayload()
	definitionPayload.DraftID = "draft_library_http_definition"
	definitionPayload.Name = "Beta definition"
	definitionPayload.SourceDefinitionID = "definition_library_http"
	definitionPayload.BaseDefinitionVersion = 2
	definition := service.SaveDraft(
		requestContext,
		SaveWorkflowDraftRequest{Payload: definitionPayload},
	)
	if definition.FailureCode != "" || definition.Draft == nil {
		t.Fatalf("save definition provenance draft: %#v", definition)
	}

	derivedPayload := validSavedWorkflowDraftPayload()
	derivedPayload.DraftID = "draft_library_http_derived"
	derivedPayload.Name = "Alpha derived"
	derivedPayload.AdditionalFields = map[string]any{
		savedWorkflowDraftDerivationAdditionalField: map[string]any{
			"version":              1,
			"source_kind":          savedWorkflowDraftDerivationSourceKind,
			"source_draft_id":      source.Draft.DraftID,
			"source_draft_version": source.Draft.DraftVersion,
		},
	}
	derived := service.SaveDraft(requestContext, SaveWorkflowDraftRequest{
		Payload:                  derivedPayload,
		ExpectedLifecycleVersion: source.Draft.LifecycleVersion,
	})
	if derived.FailureCode != "" || derived.Draft == nil {
		t.Fatalf("save derived draft: %#v", derived)
	}

	archived := service.ArchiveDraft(requestContext, TransitionSavedWorkflowDraftLifecycleRequest{
		DraftID:                  definition.Draft.DraftID,
		ExpectedDraftVersion:     definition.Draft.DraftVersion,
		ExpectedLifecycleVersion: definition.Draft.LifecycleVersion,
	})
	if archived.FailureCode != "" {
		t.Fatalf("archive definition provenance draft: %#v", archived)
	}

	first := requestSavedWorkflowDraftLibraryPage(
		t,
		server,
		"?workspace_id=workspace_demo&application_id=app_flow_copilot&limit=1&name_prefix=Alpha",
		http.StatusOK,
	)
	if first.FailureCode != nil || len(first.DraftSummaries) != 1 ||
		!first.HasMore || first.NextCursor == "" {
		t.Fatalf("first active page drifted: %#v", first)
	}
	second := requestSavedWorkflowDraftLibraryPage(
		t,
		server,
		"?workspace_id=workspace_demo&application_id=app_flow_copilot&limit=1&name_prefix=Alpha&cursor="+first.NextCursor,
		http.StatusOK,
	)
	if second.FailureCode != nil || len(second.DraftSummaries) != 1 ||
		second.HasMore || second.NextCursor != "" ||
		second.DraftSummaries[0].DraftID == first.DraftSummaries[0].DraftID {
		t.Fatalf("second active page drifted: %#v", second)
	}

	archivedPage := requestSavedWorkflowDraftLibraryPage(
		t,
		server,
		"?workspace_id=workspace_demo&application_id=app_flow_copilot&lifecycle_state=archived&provenance_kind=workflow_definition",
		http.StatusOK,
	)
	if archivedPage.FailureCode != nil || len(archivedPage.DraftSummaries) != 1 ||
		archivedPage.DraftSummaries[0].DraftID != definition.Draft.DraftID ||
		archivedPage.DraftSummaries[0].LifecycleState != string(SavedWorkflowDraftLifecycleArchived) ||
		archivedPage.DraftSummaries[0].ArchivedAt == nil {
		t.Fatalf("archived filtered page drifted: %#v", archivedPage)
	}

	for _, rawQuery := range []string{
		"?workspace_id=workspace_demo&application_id=app_flow_copilot&unknown=value",
		"?workspace_id=workspace_demo&application_id=app_flow_copilot&limit=0",
		"?workspace_id=workspace_demo&application_id=app_flow_copilot&lifecycle_state=deleted",
		"?workspace_id=workspace_demo&application_id=app_flow_copilot&cursor=not-a-cursor",
	} {
		failed := requestSavedWorkflowDraftLibraryPage(t, server, rawQuery, http.StatusOK)
		if failed.FailureCode == nil ||
			(*failed.FailureCode != string(SavedWorkflowDraftFailureListFilterInvalid) &&
				*failed.FailureCode != string(SavedWorkflowDraftFailureListCursorInvalid)) ||
			len(failed.DraftSummaries) != 0 || failed.NextCursor != "" {
			t.Fatalf("invalid list query did not fail closed: query=%s envelope=%#v", rawQuery, failed)
		}
	}
}

func TestSavedWorkflowDraftLibraryHTTPLifecycleAuthorizationAndSanitizedResponse(t *testing.T) {
	server := newSavedWorkflowDraftHTTPTestServer(true)
	t.Cleanup(server.Close)
	provider := &countingWorkspaceMembershipProvider{
		delegate: server.workspaceMembershipProvider,
	}
	server.workspaceMembershipProvider = provider

	service := server.savedWorkflowDraftService()
	requestContext := savedWorkflowDraftTestContext()
	payload := validSavedWorkflowDraftPayload()
	payload.DraftID = "draft_library_http_lifecycle"
	saved := service.SaveDraft(requestContext, SaveWorkflowDraftRequest{Payload: payload})
	if saved.FailureCode != "" || saved.Draft == nil {
		t.Fatalf("save lifecycle fixture: %#v", saved)
	}
	baseline := server.savedWorkflowDraftStore.SideEffects()
	if grants := projectControlPlaneReadPermissions(
		[]string{"radishmind.workflow-drafts.archive"},
	); !controlPlaneReadHasScope(grants, "workflow_drafts:archive") {
		t.Fatalf("archive permission projection drifted: %#v", grants)
	}

	mismatch := requestSavedWorkflowDraftLifecycle(
		t,
		server,
		saved.Draft.DraftID,
		"archive",
		savedWorkflowDraftLifecycleHTTPBody{
			WorkspaceID:              "workspace_other",
			ApplicationID:            payload.ApplicationID,
			ExpectedDraftVersion:     saved.Draft.DraftVersion,
			ExpectedLifecycleVersion: saved.Draft.LifecycleVersion,
		},
		"workflow_drafts:read,workflow_drafts:archive",
		"",
		http.StatusForbidden,
	)
	if mismatch.FailureCode == nil ||
		*mismatch.FailureCode != string(SavedWorkflowDraftFailureScopeDenied) ||
		server.savedWorkflowDraftStore.SideEffects() != baseline {
		t.Fatalf("body/header binding mismatch reached lifecycle owner: %#v", mismatch)
	}

	unknownRequest := httptest.NewRequest(
		http.MethodPost,
		"/v1/user-workspace/workflow-drafts/"+saved.Draft.DraftID+"/archive",
		strings.NewReader(`{
			"workspace_id":"workspace_demo",
			"application_id":"app_flow_copilot",
			"expected_draft_version":1,
			"expected_lifecycle_version":1,
			"unknown":true
		}`),
	)
	setSavedWorkflowDraftDevHeaders(
		unknownRequest,
		"workflow_drafts:read,workflow_drafts:archive",
	)
	unknownRecorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(unknownRecorder, unknownRequest)
	if unknownRecorder.Code != http.StatusBadRequest ||
		!strings.Contains(unknownRecorder.Body.String(), "INVALID_JSON") ||
		server.savedWorkflowDraftStore.SideEffects() != baseline {
		t.Fatalf(
			"unknown lifecycle body field reached owner: status=%d body=%s",
			unknownRecorder.Code,
			unknownRecorder.Body.String(),
		)
	}

	provider.calls.Store(0)
	activeUnarchive := requestSavedWorkflowDraftLifecycle(
		t,
		server,
		saved.Draft.DraftID,
		"unarchive",
		savedWorkflowDraftLifecycleHTTPBody{
			WorkspaceID:              payload.WorkspaceID,
			ApplicationID:            payload.ApplicationID,
			ExpectedDraftVersion:     saved.Draft.DraftVersion,
			ExpectedLifecycleVersion: saved.Draft.LifecycleVersion,
		},
		"workflow_drafts:read,workflow_drafts:archive",
		"",
		http.StatusOK,
	)
	if activeUnarchive.FailureCode == nil ||
		*activeUnarchive.FailureCode != string(SavedWorkflowDraftFailureLifecycleStateConflict) ||
		activeUnarchive.CurrentLifecycleState != string(SavedWorkflowDraftLifecycleActive) ||
		server.savedWorkflowDraftStore.SideEffects() != baseline {
		t.Fatalf("unarchive mutated an active draft: %#v", activeUnarchive)
	}

	provider.calls.Store(0)
	permissionDenied := requestSavedWorkflowDraftLifecycle(
		t,
		server,
		saved.Draft.DraftID,
		"archive",
		savedWorkflowDraftLifecycleHTTPBody{
			WorkspaceID:              payload.WorkspaceID,
			ApplicationID:            payload.ApplicationID,
			ExpectedDraftVersion:     saved.Draft.DraftVersion,
			ExpectedLifecycleVersion: saved.Draft.LifecycleVersion,
		},
		"workflow_drafts:read,workflow_drafts:archive",
		"workflow_drafts:read",
		http.StatusForbidden,
	)
	if permissionDenied.FailureCode == nil ||
		*permissionDenied.FailureCode != "workspace_permission_denied" ||
		permissionDenied.Lifecycle != nil || provider.calls.Load() != 1 ||
		server.savedWorkflowDraftStore.SideEffects() != baseline {
		t.Fatalf("membership denial reached lifecycle owner: %#v", permissionDenied)
	}

	provider.calls.Store(0)
	oidcRequest := httptest.NewRequest(
		http.MethodPost,
		"/v1/user-workspace/workflow-drafts/"+saved.Draft.DraftID+"/archive",
		bytes.NewReader(mustSavedWorkflowDraftJSON(t, savedWorkflowDraftLifecycleHTTPBody{
			WorkspaceID:              payload.WorkspaceID,
			ApplicationID:            payload.ApplicationID,
			ExpectedDraftVersion:     saved.Draft.DraftVersion,
			ExpectedLifecycleVersion: saved.Draft.LifecycleVersion,
		})),
	)
	oidcRequest.SetPathValue("draft_id", saved.Draft.DraftID)
	oidcRequest.Header.Set(activeWorkspaceHeader, payload.WorkspaceID)
	oidcAuth := batchBMutationAuth(
		fixedSavedWorkflowDraftClock()(),
		"workflow_drafts:read",
		"workflow_drafts:archive",
	)
	oidcAuth.AuthMode = controlPlaneReadAuthModeRadishOIDCIntegrationTest
	oidcRequest = oidcRequest.WithContext(
		withControlPlaneReadFakeAuthContext(oidcRequest.Context(), oidcAuth),
	)
	oidcRecorder := httptest.NewRecorder()
	server.handleArchiveWorkflowDraft(oidcRecorder, oidcRequest)
	if oidcRecorder.Code != http.StatusServiceUnavailable ||
		!strings.Contains(oidcRecorder.Body.String(), "workspace_membership_unavailable") ||
		provider.calls.Load() != 0 ||
		server.savedWorkflowDraftStore.SideEffects() != baseline {
		t.Fatalf(
			"OIDC membership unavailability reached lifecycle owner: status=%d body=%s calls=%d",
			oidcRecorder.Code,
			oidcRecorder.Body.String(),
			provider.calls.Load(),
		)
	}

	provider.calls.Store(0)
	archived := requestSavedWorkflowDraftLifecycle(
		t,
		server,
		saved.Draft.DraftID,
		"archive",
		savedWorkflowDraftLifecycleHTTPBody{
			WorkspaceID:              payload.WorkspaceID,
			ApplicationID:            payload.ApplicationID,
			ExpectedDraftVersion:     saved.Draft.DraftVersion,
			ExpectedLifecycleVersion: saved.Draft.LifecycleVersion,
		},
		"workflow_drafts:read,workflow_drafts:archive",
		"",
		http.StatusOK,
	)
	if archived.FailureCode != nil || archived.Lifecycle == nil ||
		archived.Lifecycle.LifecycleState != string(SavedWorkflowDraftLifecycleArchived) ||
		archived.Lifecycle.LifecycleVersion != 2 ||
		archived.CurrentDraftVersion != saved.Draft.DraftVersion ||
		archived.CurrentLifecycleVersion != 2 ||
		archived.CurrentLifecycleState != string(SavedWorkflowDraftLifecycleArchived) ||
		provider.calls.Load() != 1 {
		t.Fatalf("archive response drifted: %#v provider_calls=%d", archived, provider.calls.Load())
	}
	sideEffects := server.savedWorkflowDraftStore.SideEffects()
	if sideEffects.DraftWriteCount != baseline.DraftWriteCount ||
		sideEffects.LifecycleTransitionCount != baseline.LifecycleTransitionCount+1 ||
		sideEffects.LifecycleEventWriteCount != baseline.LifecycleEventWriteCount+1 ||
		hasSavedWorkflowDraftRuntimeSideEffect(sideEffects) {
		t.Fatalf("archive side effects drifted: before=%#v after=%#v", baseline, sideEffects)
	}

	stale := requestSavedWorkflowDraftLifecycle(
		t,
		server,
		saved.Draft.DraftID,
		"unarchive",
		savedWorkflowDraftLifecycleHTTPBody{
			WorkspaceID:              payload.WorkspaceID,
			ApplicationID:            payload.ApplicationID,
			ExpectedDraftVersion:     saved.Draft.DraftVersion,
			ExpectedLifecycleVersion: 1,
		},
		"workflow_drafts:read,workflow_drafts:archive",
		"",
		http.StatusOK,
	)
	if stale.FailureCode == nil ||
		*stale.FailureCode != string(SavedWorkflowDraftFailureLifecycleVersionConflict) ||
		stale.Lifecycle != nil || stale.CurrentLifecycleVersion != 2 ||
		stale.CurrentLifecycleState != string(SavedWorkflowDraftLifecycleArchived) {
		t.Fatalf("stale unarchive did not return sanitized current metadata: %#v", stale)
	}

	unarchived := requestSavedWorkflowDraftLifecycle(
		t,
		server,
		saved.Draft.DraftID,
		"unarchive",
		savedWorkflowDraftLifecycleHTTPBody{
			WorkspaceID:              payload.WorkspaceID,
			ApplicationID:            payload.ApplicationID,
			ExpectedDraftVersion:     saved.Draft.DraftVersion,
			ExpectedLifecycleVersion: 2,
		},
		"workflow_drafts:read,workflow_drafts:archive",
		"",
		http.StatusOK,
	)
	if unarchived.FailureCode != nil || unarchived.Lifecycle == nil ||
		unarchived.Lifecycle.LifecycleState != string(SavedWorkflowDraftLifecycleActive) ||
		unarchived.Lifecycle.LifecycleVersion != 3 ||
		unarchived.Lifecycle.ArchivedAt != nil {
		t.Fatalf("unarchive response drifted: %#v", unarchived)
	}
}

func requestSavedWorkflowDraftLibraryPage(
	t *testing.T,
	server *Server,
	rawQuery string,
	expectedStatus int,
) savedWorkflowDraftListEnvelope {
	t.Helper()
	request := httptest.NewRequest(
		http.MethodGet,
		"/v1/user-workspace/workflow-drafts"+rawQuery,
		nil,
	)
	setSavedWorkflowDraftDevHeaders(request, "workflow_drafts:read")
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	return decodeSavedWorkflowDraftListEnvelope(t, recorder, expectedStatus)
}

func requestSavedWorkflowDraftLifecycle(
	t *testing.T,
	server *Server,
	draftID string,
	transition string,
	body savedWorkflowDraftLifecycleHTTPBody,
	scopes string,
	membershipPermissions string,
	expectedStatus int,
) savedWorkflowDraftLifecycleEnvelope {
	t.Helper()
	rawBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal lifecycle body: %v", err)
	}
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/user-workspace/workflow-drafts/"+draftID+"/"+transition,
		bytes.NewReader(rawBody),
	)
	setSavedWorkflowDraftDevHeaders(request, scopes)
	if membershipPermissions != "" {
		request.Header.Set(controlPlaneReadDevMembershipPermHeader, membershipPermissions)
	}
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != expectedStatus {
		t.Fatalf("expected lifecycle status %d, got %d: %s", expectedStatus, recorder.Code, recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), `"nodes"`) ||
		strings.Contains(recorder.Body.String(), `"provider_refs"`) ||
		strings.Contains(recorder.Body.String(), `"additional_fields"`) {
		t.Fatalf("lifecycle response leaked draft payload: %s", recorder.Body.String())
	}
	var envelope savedWorkflowDraftLifecycleEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode lifecycle envelope: %v\n%s", err, recorder.Body.String())
	}
	return envelope
}
