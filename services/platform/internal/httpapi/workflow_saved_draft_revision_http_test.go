package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func TestSavedWorkflowDraftRevisionHTTPRoutes(t *testing.T) {
	server := newSavedWorkflowDraftHTTPTestServer(true)
	t.Cleanup(server.Close)
	payload := validSavedWorkflowDraftPayload()
	payload.Name = "HTTP revision one"
	saveWorkflowDraftRevisionHTTPFixture(t, server, payload, 0, 1)
	payload.Name = "HTTP revision two"
	saveWorkflowDraftRevisionHTTPFixture(t, server, payload, 1, 2)

	query := "?workspace_id=" + url.QueryEscape(payload.WorkspaceID) +
		"&application_id=" + url.QueryEscape(payload.ApplicationID)
	listRequest := httptest.NewRequest(
		http.MethodGet,
		"/v1/user-workspace/workflow-drafts/"+payload.DraftID+"/revisions"+
			query+"&limit=1",
		nil,
	)
	setSavedWorkflowDraftDevHeaders(listRequest, "workflow_drafts:read,workflow_drafts:write")
	listRecorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(listRecorder, listRequest)
	listEnvelope := decodeSavedWorkflowDraftRevisionListEnvelope(t, listRecorder, http.StatusOK)
	if listEnvelope.FailureCode != nil || len(listEnvelope.Revisions) != 1 ||
		listEnvelope.Revisions[0].DraftVersion != 2 || !listEnvelope.HasMore ||
		listEnvelope.NextCursor == "" {
		t.Fatalf("list revision history: %#v", listEnvelope)
	}

	readRequest := httptest.NewRequest(
		http.MethodGet,
		"/v1/user-workspace/workflow-drafts/"+payload.DraftID+"/revisions/1"+query,
		nil,
	)
	setSavedWorkflowDraftDevHeaders(readRequest, "workflow_drafts:read")
	readRecorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(readRecorder, readRequest)
	readEnvelope := decodeSavedWorkflowDraftRevisionEnvelope(t, readRecorder, http.StatusOK)
	if readEnvelope.FailureCode != nil || readEnvelope.Revision == nil ||
		readEnvelope.Revision.Draft == nil ||
		readEnvelope.Revision.Draft.Name != "HTTP revision one" ||
		readEnvelope.Revision.RevisionKind != string(SavedWorkflowDraftRevisionKindSaved) {
		t.Fatalf("read revision: %#v", readEnvelope)
	}

	restoreBody := mustSavedWorkflowDraftJSON(t, savedWorkflowDraftRevisionRestoreHTTPBody{
		ExpectedCurrentDraftVersion: 2,
	})
	restoreRequest := httptest.NewRequest(
		http.MethodPost,
		"/v1/user-workspace/workflow-drafts/"+payload.DraftID+"/revisions/1/restore",
		bytes.NewReader(restoreBody),
	)
	setSavedWorkflowDraftDevHeaders(restoreRequest, "workflow_drafts:read,workflow_drafts:write")
	restoreRecorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(restoreRecorder, restoreRequest)
	restoreEnvelope := decodeSavedWorkflowDraftRevisionEnvelope(t, restoreRecorder, http.StatusOK)
	if restoreEnvelope.FailureCode != nil || restoreEnvelope.Draft == nil ||
		restoreEnvelope.Draft.DraftVersion != 3 ||
		restoreEnvelope.Draft.Name != "HTTP revision one" {
		t.Fatalf("restore revision: %#v", restoreEnvelope)
	}

	restoredRead := httptest.NewRequest(
		http.MethodGet,
		"/v1/user-workspace/workflow-drafts/"+payload.DraftID+"/revisions/3"+query,
		nil,
	)
	setSavedWorkflowDraftDevHeaders(restoredRead, "workflow_drafts:read")
	restoredRecorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(restoredRecorder, restoredRead)
	restoredEnvelope := decodeSavedWorkflowDraftRevisionEnvelope(t, restoredRecorder, http.StatusOK)
	if restoredEnvelope.Revision == nil ||
		restoredEnvelope.Revision.RevisionKind != string(SavedWorkflowDraftRevisionKindRestored) ||
		restoredEnvelope.Revision.RestoredFromVersion != 1 {
		t.Fatalf("read restored revision metadata: %#v", restoredEnvelope)
	}

	unauthorized := httptest.NewRequest(
		http.MethodPost,
		"/v1/user-workspace/workflow-drafts/"+payload.DraftID+"/revisions/2/restore",
		bytes.NewBufferString(`{"malformed"`),
	)
	setSavedWorkflowDraftDevHeaders(unauthorized, "workflow_drafts:read")
	unauthorizedRecorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(unauthorizedRecorder, unauthorized)
	unauthorizedEnvelope := decodeSavedWorkflowDraftRevisionEnvelope(
		t,
		unauthorizedRecorder,
		http.StatusForbidden,
	)
	if unauthorizedEnvelope.FailureCode == nil ||
		*unauthorizedEnvelope.FailureCode != "scope_denied" {
		t.Fatalf("restore must authorize before decoding body: %#v", unauthorizedEnvelope)
	}
}

func saveWorkflowDraftRevisionHTTPFixture(
	t *testing.T,
	server *Server,
	payload SavedWorkflowDraftPayload,
	expectedVersion int,
	wantedVersion int,
) {
	t.Helper()
	body := mustSavedWorkflowDraftJSON(t, savedWorkflowDraftSaveHTTPBody{
		ExpectedDraftVersion: expectedVersion,
		Draft:                savedWorkflowDraftPayloadDocumentFromDraftPayload(payload),
	})
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/user-workspace/workflow-drafts",
		bytes.NewReader(body),
	)
	setSavedWorkflowDraftDevHeaders(request, "workflow_drafts:read,workflow_drafts:write")
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	envelope := decodeSavedWorkflowDraftEnvelope(t, recorder, http.StatusOK)
	if envelope.FailureCode != nil || envelope.CurrentDraftVersion != wantedVersion {
		t.Fatalf("save HTTP revision %s: %#v", strconv.Itoa(wantedVersion), envelope)
	}
}

func decodeSavedWorkflowDraftRevisionEnvelope(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
	expectedStatus int,
) savedWorkflowDraftRevisionEnvelope {
	t.Helper()
	if recorder.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d: %s", expectedStatus, recorder.Code, recorder.Body.String())
	}
	var envelope savedWorkflowDraftRevisionEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode saved workflow draft revision envelope: %v", err)
	}
	return envelope
}

func decodeSavedWorkflowDraftRevisionListEnvelope(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
	expectedStatus int,
) savedWorkflowDraftRevisionListEnvelope {
	t.Helper()
	if recorder.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d: %s", expectedStatus, recorder.Code, recorder.Body.String())
	}
	var envelope savedWorkflowDraftRevisionListEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode saved workflow draft revision list envelope: %v", err)
	}
	return envelope
}
