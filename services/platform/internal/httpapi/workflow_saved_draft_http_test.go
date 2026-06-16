package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
		assertSavedWorkflowDraftEnvelopeContract(t, saveRec)
		if saveEnvelope.FailureCode != nil || saveEnvelope.Draft == nil {
			t.Fatalf("save should succeed: %#v", saveEnvelope)
		}
		if saveEnvelope.Draft.SampleOrUnsavedDraftStatus != "saved_draft_record" || saveEnvelope.CurrentDraftVersion != 1 {
			t.Fatalf("save did not return saved record metadata: %#v", saveEnvelope)
		}
		if saveEnvelope.WorkspaceID != payload.WorkspaceID || saveEnvelope.ApplicationID != payload.ApplicationID {
			t.Fatalf("save envelope scope drifted: %#v", saveEnvelope)
		}
		if saveEnvelope.AuditRef == "" || saveEnvelope.Draft.RequestAuditMetadata.ActorRef == "" {
			t.Fatalf("save envelope must preserve audit metadata: %#v", saveEnvelope)
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
		assertSavedWorkflowDraftEnvelopeContract(t, readRec)
		if readEnvelope.FailureCode != nil || readEnvelope.Draft == nil || readEnvelope.Draft.DraftVersion != 1 {
			t.Fatalf("read should return saved draft record: %#v", readEnvelope)
		}

		listReq := httptest.NewRequest(
			http.MethodGet,
			"/v1/user-workspace/workflow-drafts?workspace_id="+payload.WorkspaceID+"&application_id="+payload.ApplicationID,
			nil,
		)
		setSavedWorkflowDraftDevHeaders(listReq, "workflow_drafts:read,workflow_drafts:write")
		listRec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(listRec, listReq)

		listEnvelope := decodeSavedWorkflowDraftListEnvelope(t, listRec, http.StatusOK)
		assertSavedWorkflowDraftListEnvelopeContract(t, listRec)
		if listEnvelope.FailureCode != nil || len(listEnvelope.DraftSummaries) != 1 {
			t.Fatalf("list should return one saved draft summary: %#v", listEnvelope)
		}
		if listEnvelope.DraftSummaries[0].DraftID != payload.DraftID ||
			listEnvelope.DraftSummaries[0].SampleOrUnsavedDraftStatus != "saved_draft_record" ||
			listEnvelope.DraftSummaries[0].NodeCount == 0 {
			t.Fatalf("list summary should stay sanitized and scoped: %#v", listEnvelope.DraftSummaries[0])
		}

		validateBody := mustSavedWorkflowDraftJSON(t, savedWorkflowDraftValidateHTTPBody{
			Draft: savedWorkflowDraftPayloadDocumentFromDraftPayload(payload),
		})
		validateReq := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-drafts/validate", bytes.NewReader(validateBody))
		setSavedWorkflowDraftDevHeaders(validateReq, "workflow_drafts:read,workflow_drafts:write")
		validateRec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(validateRec, validateReq)

		validateEnvelope := decodeSavedWorkflowDraftEnvelope(t, validateRec, http.StatusOK)
		assertSavedWorkflowDraftEnvelopeContract(t, validateRec)
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

		listReq := httptest.NewRequest(
			http.MethodGet,
			"/v1/user-workspace/workflow-drafts?workspace_id="+payload.WorkspaceID+"&application_id="+payload.ApplicationID,
			nil,
		)
		setSavedWorkflowDraftDevHeaders(listReq, "workflow_drafts:read,workflow_drafts:write")
		listRec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(listRec, listReq)

		listEnvelope := decodeSavedWorkflowDraftListEnvelope(t, listRec, http.StatusOK)
		if listEnvelope.FailureCode != nil || len(listEnvelope.DraftSummaries) != 0 {
			t.Fatalf("write disabled save must not leave list fallback: %#v", listEnvelope)
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

		scopeListReq := httptest.NewRequest(
			http.MethodGet,
			"/v1/user-workspace/workflow-drafts?workspace_id="+payload.WorkspaceID+"&application_id=app_other",
			nil,
		)
		setSavedWorkflowDraftDevHeaders(scopeListReq, "workflow_drafts:read,workflow_drafts:write")
		scopeListReq.Header.Set(savedWorkflowDraftDevApplicationHeader, "app_other")
		scopeListRec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(scopeListRec, scopeListReq)

		scopeListEnvelope := decodeSavedWorkflowDraftListEnvelope(t, scopeListRec, http.StatusOK)
		if scopeListEnvelope.FailureCode != nil || len(scopeListEnvelope.DraftSummaries) != 0 {
			t.Fatalf("list for wrong application must not fallback to saved sample: %#v", scopeListEnvelope)
		}
	})

	t.Run("route contract keeps failure envelope fail closed", func(t *testing.T) {
		server := newSavedWorkflowDraftHTTPTestServer(true)
		payload := validSavedWorkflowDraftPayload()
		readReq := httptest.NewRequest(
			http.MethodGet,
			"/v1/user-workspace/workflow-drafts/"+payload.DraftID+"?workspace_id="+payload.WorkspaceID+"&application_id="+payload.ApplicationID,
			nil,
		)
		setSavedWorkflowDraftDevHeaders(readReq, "workflow_drafts:read,workflow_drafts:write")
		readRec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(readRec, readReq)

		readEnvelope := decodeSavedWorkflowDraftEnvelope(t, readRec, http.StatusOK)
		assertSavedWorkflowDraftEnvelopeContract(t, readRec)
		if readEnvelope.FailureCode == nil ||
			*readEnvelope.FailureCode != string(SavedWorkflowDraftFailureNotFound) ||
			readEnvelope.Draft != nil {
			t.Fatalf("not found must stay fail closed without sample fallback: %#v", readEnvelope)
		}

		store := newMemorySavedWorkflowDraftStore()
		store.unavailable = true
		server.savedWorkflowDraftStore = store
		unavailableReq := httptest.NewRequest(
			http.MethodGet,
			"/v1/user-workspace/workflow-drafts/"+payload.DraftID+"?workspace_id="+payload.WorkspaceID+"&application_id="+payload.ApplicationID,
			nil,
		)
		setSavedWorkflowDraftDevHeaders(unavailableReq, "workflow_drafts:read,workflow_drafts:write")
		unavailableRec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(unavailableRec, unavailableReq)

		unavailableEnvelope := decodeSavedWorkflowDraftEnvelope(t, unavailableRec, http.StatusOK)
		assertSavedWorkflowDraftEnvelopeContract(t, unavailableRec)
		if unavailableEnvelope.FailureCode == nil ||
			*unavailableEnvelope.FailureCode != string(SavedWorkflowDraftFailureStoreUnavailable) ||
			unavailableEnvelope.Draft != nil {
			t.Fatalf("store unavailable must stay fail closed without sample fallback: %#v", unavailableEnvelope)
		}

		unavailableListReq := httptest.NewRequest(
			http.MethodGet,
			"/v1/user-workspace/workflow-drafts?workspace_id="+payload.WorkspaceID+"&application_id="+payload.ApplicationID,
			nil,
		)
		setSavedWorkflowDraftDevHeaders(unavailableListReq, "workflow_drafts:read,workflow_drafts:write")
		unavailableListRec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(unavailableListRec, unavailableListReq)

		unavailableListEnvelope := decodeSavedWorkflowDraftListEnvelope(t, unavailableListRec, http.StatusOK)
		assertSavedWorkflowDraftListEnvelopeContract(t, unavailableListRec)
		if unavailableListEnvelope.FailureCode == nil ||
			*unavailableListEnvelope.FailureCode != string(SavedWorkflowDraftFailureStoreUnavailable) ||
			len(unavailableListEnvelope.DraftSummaries) != 0 {
			t.Fatalf("store unavailable list must stay fail closed without sample fallback: %#v", unavailableListEnvelope)
		}
	})

	t.Run("cors preflight exposes saved draft dev headers", func(t *testing.T) {
		server := newSavedWorkflowDraftHTTPTestServer(true)
		req := httptest.NewRequest(http.MethodOptions, "/v1/user-workspace/workflow-drafts", nil)
		req.Header.Set("Origin", "http://127.0.0.1:4100")
		req.Header.Set("Access-Control-Request-Method", http.MethodPost)
		req.Header.Set("Access-Control-Request-Headers", savedWorkflowDraftDevWorkspaceHeader)
		rec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected CORS preflight to return 204, got %d: %s", rec.Code, rec.Body.String())
		}
		allowedHeaders := rec.Header().Get("Access-Control-Allow-Headers")
		for _, header := range []string{savedWorkflowDraftDevWorkspaceHeader, savedWorkflowDraftDevApplicationHeader} {
			if !strings.Contains(allowedHeaders, header) {
				t.Fatalf("CORS allowed headers missing %s: %s", header, allowedHeaders)
			}
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

func decodeSavedWorkflowDraftListEnvelope(
	t *testing.T,
	rec *httptest.ResponseRecorder,
	expectedStatus int,
) savedWorkflowDraftListEnvelope {
	t.Helper()
	if rec.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d: %s", expectedStatus, rec.Code, rec.Body.String())
	}
	var envelope savedWorkflowDraftListEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode saved workflow draft list envelope: %v\n%s", err, rec.Body.String())
	}
	return envelope
}

func assertSavedWorkflowDraftEnvelopeContract(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json; charset=utf-8" {
		t.Fatalf("saved draft route must return JSON content type, got %q", contentType)
	}
	var document map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &document); err != nil {
		t.Fatalf("decode saved workflow draft contract envelope: %v\n%s", err, rec.Body.String())
	}
	for _, key := range []string{
		"request_id",
		"workspace_id",
		"application_id",
		"draft",
		"failure_code",
		"current_draft_version",
		"validation_summary",
		"blocked_capabilities",
		"audit_ref",
	} {
		if _, ok := document[key]; !ok {
			t.Fatalf("saved draft envelope missing required key %q: %#v", key, document)
		}
	}
	validationSummary, ok := document["validation_summary"].(map[string]any)
	if !ok {
		t.Fatalf("saved draft validation_summary must be an object: %#v", document["validation_summary"])
	}
	for _, key := range []string{"validation_state", "valid_for_review", "findings"} {
		if _, ok := validationSummary[key]; !ok {
			t.Fatalf("saved draft validation_summary missing required key %q: %#v", key, validationSummary)
		}
	}
}

func assertSavedWorkflowDraftListEnvelopeContract(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json; charset=utf-8" {
		t.Fatalf("saved draft list route must return JSON content type, got %q", contentType)
	}
	var document map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &document); err != nil {
		t.Fatalf("decode saved workflow draft list contract envelope: %v\n%s", err, rec.Body.String())
	}
	for _, key := range []string{
		"request_id",
		"workspace_id",
		"application_id",
		"draft_summaries",
		"failure_code",
		"audit_ref",
	} {
		if _, ok := document[key]; !ok {
			t.Fatalf("saved draft list envelope missing required key %q: %#v", key, document)
		}
	}
	summaries, ok := document["draft_summaries"].([]any)
	if !ok {
		t.Fatalf("saved draft list envelope must expose draft_summaries array: %#v", document["draft_summaries"])
	}
	for _, rawSummary := range summaries {
		summary, ok := rawSummary.(map[string]any)
		if !ok {
			t.Fatalf("saved draft list summary must be an object: %#v", rawSummary)
		}
		for _, key := range []string{
			"draft_id",
			"workspace_id",
			"application_id",
			"source_definition_id",
			"draft_version",
			"schema_version",
			"draft_status",
			"name",
			"updated_at",
			"node_count",
			"edge_count",
			"validation_state",
			"valid_for_review",
			"sample_or_unsaved_draft_status",
		} {
			if _, ok := summary[key]; !ok {
				t.Fatalf("saved draft list summary missing key %q: %#v", key, summary)
			}
		}
	}
}
