package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"radishmind.local/services/platform/internal/config"
)

func TestSavedWorkflowDraftHTTPRoutes(t *testing.T) {
	t.Run("dev route disabled by default", func(t *testing.T) {
		server := NewServer(config.Config{}, Options{BuildVersion: "test"})
		body := mustSavedWorkflowDraftJSON(t, savedWorkflowDraftSaveHTTPBody{
			Draft: savedWorkflowDraftPayloadDocumentFromDraftPayload(validSavedWorkflowDraftPayload()),
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-drafts", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected disabled dev route to fail with 403, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("save read validate over dev-only route", func(t *testing.T) {
		server := newSavedWorkflowDraftHTTPTestServer(true)
		payload := validSavedWorkflowDraftPayload()
		saveBody := mustSavedWorkflowDraftJSON(t, savedWorkflowDraftSaveHTTPBody{
			ExpectedDraftVersion: 0,
			Draft:                savedWorkflowDraftPayloadDocumentFromDraftPayload(payload),
		})
		saveReq := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-drafts", bytes.NewReader(saveBody))
		setSavedWorkflowDraftDevHeaders(saveReq, "workflow_drafts:read,workflow_drafts:write")
		saveRec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(saveRec, saveReq)

		saveEnvelope := decodeSavedWorkflowDraftEnvelope(t, saveRec, http.StatusOK)
		if saveEnvelope.FailureCode != nil || saveEnvelope.Draft == nil {
			t.Fatalf("save should succeed: %#v", saveEnvelope)
		}
		if saveEnvelope.Draft.SampleOrUnsavedDraftStatus != "saved_draft_record" || saveEnvelope.CurrentDraftVersion != 1 {
			t.Fatalf("save did not return saved record metadata: %#v", saveEnvelope)
		}

		readReq := httptest.NewRequest(
			http.MethodGet,
			"/v1/user-workspace/workflow-drafts/"+payload.DraftID+"?workspace_id="+payload.WorkspaceID+"&application_id="+payload.ApplicationID,
			nil,
		)
		setSavedWorkflowDraftDevHeaders(readReq, "workflow_drafts:read,workflow_drafts:write")
		readRec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(readRec, readReq)

		readEnvelope := decodeSavedWorkflowDraftEnvelope(t, readRec, http.StatusOK)
		if readEnvelope.FailureCode != nil || readEnvelope.Draft == nil || readEnvelope.Draft.DraftVersion != 1 {
			t.Fatalf("read should return saved draft record: %#v", readEnvelope)
		}

		validateBody := mustSavedWorkflowDraftJSON(t, savedWorkflowDraftValidateHTTPBody{
			Draft: savedWorkflowDraftPayloadDocumentFromDraftPayload(payload),
		})
		validateReq := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-drafts/validate", bytes.NewReader(validateBody))
		setSavedWorkflowDraftDevHeaders(validateReq, "workflow_drafts:read,workflow_drafts:write")
		validateRec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(validateRec, validateReq)

		validateEnvelope := decodeSavedWorkflowDraftEnvelope(t, validateRec, http.StatusOK)
		if validateEnvelope.FailureCode != nil || validateEnvelope.Draft != nil {
			t.Fatalf("validate should return summary without materializing draft: %#v", validateEnvelope)
		}
		if !validateEnvelope.ValidationSummary.ValidForReview {
			t.Fatalf("validate should preserve valid_for_review summary: %#v", validateEnvelope.ValidationSummary)
		}
	})

	t.Run("write disabled returns draft_write_disabled without partial write", func(t *testing.T) {
		server := newSavedWorkflowDraftHTTPTestServer(false)
		payload := validSavedWorkflowDraftPayload()
		saveBody := mustSavedWorkflowDraftJSON(t, savedWorkflowDraftSaveHTTPBody{
			Draft: savedWorkflowDraftPayloadDocumentFromDraftPayload(payload),
		})
		saveReq := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-drafts", bytes.NewReader(saveBody))
		setSavedWorkflowDraftDevHeaders(saveReq, "workflow_drafts:read,workflow_drafts:write")
		saveRec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(saveRec, saveReq)

		saveEnvelope := decodeSavedWorkflowDraftEnvelope(t, saveRec, http.StatusOK)
		if saveEnvelope.FailureCode == nil || *saveEnvelope.FailureCode != string(SavedWorkflowDraftFailureWriteDisabled) {
			t.Fatalf("save should fail with write disabled: %#v", saveEnvelope)
		}

		readReq := httptest.NewRequest(
			http.MethodGet,
			"/v1/user-workspace/workflow-drafts/"+payload.DraftID+"?workspace_id="+payload.WorkspaceID+"&application_id="+payload.ApplicationID,
			nil,
		)
		setSavedWorkflowDraftDevHeaders(readReq, "workflow_drafts:read,workflow_drafts:write")
		readRec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(readRec, readReq)

		readEnvelope := decodeSavedWorkflowDraftEnvelope(t, readRec, http.StatusOK)
		if readEnvelope.FailureCode == nil || *readEnvelope.FailureCode != string(SavedWorkflowDraftFailureNotFound) {
			t.Fatalf("write disabled save must not leave partial draft: %#v", readEnvelope)
		}
	})

	t.Run("stale version and scope mismatch fail closed", func(t *testing.T) {
		server := newSavedWorkflowDraftHTTPTestServer(true)
		payload := validSavedWorkflowDraftPayload()
		saveBody := mustSavedWorkflowDraftJSON(t, savedWorkflowDraftSaveHTTPBody{
			Draft: savedWorkflowDraftPayloadDocumentFromDraftPayload(payload),
		})
		saveReq := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-drafts", bytes.NewReader(saveBody))
		setSavedWorkflowDraftDevHeaders(saveReq, "workflow_drafts:read,workflow_drafts:write")
		server.httpServer.Handler.ServeHTTP(httptest.NewRecorder(), saveReq)

		staleBody := mustSavedWorkflowDraftJSON(t, savedWorkflowDraftSaveHTTPBody{
			ExpectedDraftVersion: 0,
			Draft:                savedWorkflowDraftPayloadDocumentFromDraftPayload(payload),
		})
		staleReq := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-drafts", bytes.NewReader(staleBody))
		setSavedWorkflowDraftDevHeaders(staleReq, "workflow_drafts:read,workflow_drafts:write")
		staleRec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(staleRec, staleReq)

		staleEnvelope := decodeSavedWorkflowDraftEnvelope(t, staleRec, http.StatusOK)
		if staleEnvelope.FailureCode == nil ||
			*staleEnvelope.FailureCode != string(SavedWorkflowDraftFailureVersionConflict) ||
			staleEnvelope.CurrentDraftVersion != 1 {
			t.Fatalf("stale save should return version conflict: %#v", staleEnvelope)
		}

		scopeReq := httptest.NewRequest(
			http.MethodGet,
			"/v1/user-workspace/workflow-drafts/"+payload.DraftID+"?workspace_id="+payload.WorkspaceID+"&application_id=app_other",
			nil,
		)
		setSavedWorkflowDraftDevHeaders(scopeReq, "workflow_drafts:read,workflow_drafts:write")
		scopeReq.Header.Set(savedWorkflowDraftDevApplicationHeader, "app_other")
		scopeRec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(scopeRec, scopeReq)

		scopeEnvelope := decodeSavedWorkflowDraftEnvelope(t, scopeRec, http.StatusOK)
		if scopeEnvelope.FailureCode == nil || *scopeEnvelope.FailureCode != string(SavedWorkflowDraftFailureScopeDenied) {
			t.Fatalf("scope mismatch should fail closed: %#v", scopeEnvelope)
		}
	})
}

func newSavedWorkflowDraftHTTPTestServer(writeEnabled bool) *Server {
	return NewServer(config.Config{
		ControlPlaneReadDevAuthEnabled:    true,
		WorkflowSavedDraftDevHTTPEnabled:  true,
		WorkflowSavedDraftDevWriteEnabled: writeEnabled,
	}, Options{BuildVersion: "test"})
}

func setSavedWorkflowDraftDevHeaders(req *http.Request, scopes string) {
	setControlPlaneReadDevAuthHeaders(req)
	req.Header.Set(controlPlaneReadDevScopesHeader, scopes)
	req.Header.Set(savedWorkflowDraftDevWorkspaceHeader, "workspace_demo")
	req.Header.Set(savedWorkflowDraftDevApplicationHeader, "app_flow_copilot")
}

func mustSavedWorkflowDraftJSON(t *testing.T, document any) []byte {
	t.Helper()
	body, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("marshal saved workflow draft test body: %v", err)
	}
	return body
}

func decodeSavedWorkflowDraftEnvelope(
	t *testing.T,
	rec *httptest.ResponseRecorder,
	expectedStatus int,
) savedWorkflowDraftEnvelope {
	t.Helper()
	if rec.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d: %s", expectedStatus, rec.Code, rec.Body.String())
	}
	var envelope savedWorkflowDraftEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode saved workflow draft envelope: %v\n%s", err, rec.Body.String())
	}
	return envelope
}
