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

func TestWorkflowExecutorHTTPRoutes(t *testing.T) {
	t.Run("dev route disabled by default", func(t *testing.T) {
		server := NewServer(config.Config{}, Options{BuildVersion: "test"})
		request := httptest.NewRequest(
			http.MethodPost,
			"/v1/user-workspace/workflow-drafts/draft_missing/runs",
			strings.NewReader(`{"workspace_id":"workspace_demo","application_id":"app_flow_copilot","input_text":"test"}`),
		)
		response := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(response, request)

		if response.Code != http.StatusForbidden {
			t.Fatalf("disabled executor route should return 403, got %d: %s", response.Code, response.Body.String())
		}
	})

	t.Run("start and scoped read", func(t *testing.T) {
		server, testBridge, draft := newWorkflowExecutorHTTPTestServer(t)
		rawInput := "private workflow input should not be persisted"
		body := mustWorkflowRunJSON(t, workflowRunStartHTTPBody{
			WorkspaceID:   draft.WorkspaceID,
			ApplicationID: draft.ApplicationID,
			InputText:     rawInput,
			Model:         "mock",
		})
		request := httptest.NewRequest(
			http.MethodPost,
			"/v1/user-workspace/workflow-drafts/"+draft.DraftID+"/runs",
			bytes.NewReader(body),
		)
		setSavedWorkflowDraftDevHeaders(
			request,
			"workflow_drafts:read,workflow_runs:execute,workflow_runs:read",
		)
		response := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(response, request)

		envelope := decodeWorkflowRunEnvelope(t, response, http.StatusOK)
		if envelope.FailureCode != nil || envelope.Run == nil || envelope.Run.Status != WorkflowRunStatusSucceeded {
			t.Fatalf("workflow run should succeed: %#v", envelope)
		}
		if envelope.Run.DraftID != draft.DraftID || envelope.Run.DraftVersion != draft.DraftVersion {
			t.Fatalf("workflow run must consume the stored draft version: %#v", envelope.Run)
		}
		if testBridge.callCount() != 1 {
			t.Fatalf("expected one Gateway call, got %d", testBridge.callCount())
		}
		if strings.Contains(response.Body.String(), rawInput) {
			t.Fatalf("HTTP run envelope must not retain raw input: %s", response.Body.String())
		}

		readRequest := httptest.NewRequest(
			http.MethodGet,
			"/v1/user-workspace/workflow-runs/"+envelope.Run.RunID+
				"?workspace_id="+draft.WorkspaceID+"&application_id="+draft.ApplicationID,
			nil,
		)
		setSavedWorkflowDraftDevHeaders(readRequest, "workflow_runs:read")
		readResponse := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(readResponse, readRequest)

		readEnvelope := decodeWorkflowRunEnvelope(t, readResponse, http.StatusOK)
		if readEnvelope.FailureCode != nil || readEnvelope.Run == nil ||
			readEnvelope.Run.RunID != envelope.Run.RunID || readEnvelope.Run.Output != envelope.Run.Output {
			t.Fatalf("workflow run record should be readable from scoped store: %#v", readEnvelope)
		}

		otherTenantRequest := httptest.NewRequest(
			http.MethodGet,
			"/v1/user-workspace/workflow-runs/"+envelope.Run.RunID+
				"?workspace_id="+draft.WorkspaceID+"&application_id="+draft.ApplicationID,
			nil,
		)
		setSavedWorkflowDraftDevHeaders(otherTenantRequest, "workflow_runs:read")
		otherTenantRequest.Header.Set(controlPlaneReadDevTenantHeader, "tenant_other")
		otherTenantResponse := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(otherTenantResponse, otherTenantRequest)

		otherTenantEnvelope := decodeWorkflowRunEnvelope(t, otherTenantResponse, http.StatusOK)
		if otherTenantEnvelope.FailureCode == nil ||
			*otherTenantEnvelope.FailureCode != string(WorkflowRunFailureRecordNotFound) ||
			otherTenantEnvelope.Run != nil {
			t.Fatalf("cross-tenant run read must fail closed: %#v", otherTenantEnvelope)
		}
	})

	t.Run("missing execute scope fails before Gateway", func(t *testing.T) {
		server, testBridge, draft := newWorkflowExecutorHTTPTestServer(t)
		body := mustWorkflowRunJSON(t, workflowRunStartHTTPBody{
			WorkspaceID:   draft.WorkspaceID,
			ApplicationID: draft.ApplicationID,
			InputText:     "safe input",
		})
		request := httptest.NewRequest(
			http.MethodPost,
			"/v1/user-workspace/workflow-drafts/"+draft.DraftID+"/runs",
			bytes.NewReader(body),
		)
		setSavedWorkflowDraftDevHeaders(request, "workflow_drafts:read,workflow_runs:read")
		response := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(response, request)

		envelope := decodeWorkflowRunEnvelope(t, response, http.StatusOK)
		if envelope.FailureCode == nil || *envelope.FailureCode != string(WorkflowRunFailureScopeDenied) || envelope.Run != nil {
			t.Fatalf("missing execute scope must fail closed: %#v", envelope)
		}
		if testBridge.callCount() != 0 {
			t.Fatalf("scope denial must not call Gateway")
		}
	})

	t.Run("strict body and unknown draft fail closed", func(t *testing.T) {
		server, testBridge, draft := newWorkflowExecutorHTTPTestServer(t)
		unknownFieldRequest := httptest.NewRequest(
			http.MethodPost,
			"/v1/user-workspace/workflow-drafts/"+draft.DraftID+"/runs",
			strings.NewReader(`{"workspace_id":"workspace_demo","application_id":"app_flow_copilot","input_text":"test","execute_tool":true}`),
		)
		setSavedWorkflowDraftDevHeaders(
			unknownFieldRequest,
			"workflow_drafts:read,workflow_runs:execute,workflow_runs:read",
		)
		unknownFieldResponse := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(unknownFieldResponse, unknownFieldRequest)

		if unknownFieldResponse.Code != http.StatusBadRequest {
			t.Fatalf("unknown executor request field should return 400: %s", unknownFieldResponse.Body.String())
		}

		notFoundBody := mustWorkflowRunJSON(t, workflowRunStartHTTPBody{
			WorkspaceID:   draft.WorkspaceID,
			ApplicationID: draft.ApplicationID,
			InputText:     "safe input",
		})
		notFoundRequest := httptest.NewRequest(
			http.MethodPost,
			"/v1/user-workspace/workflow-drafts/draft_unknown/runs",
			bytes.NewReader(notFoundBody),
		)
		setSavedWorkflowDraftDevHeaders(
			notFoundRequest,
			"workflow_drafts:read,workflow_runs:execute,workflow_runs:read",
		)
		notFoundResponse := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(notFoundResponse, notFoundRequest)

		notFoundEnvelope := decodeWorkflowRunEnvelope(t, notFoundResponse, http.StatusOK)
		if notFoundEnvelope.FailureCode == nil ||
			*notFoundEnvelope.FailureCode != string(WorkflowRunFailureDraftNotFound) ||
			notFoundEnvelope.Run != nil {
			t.Fatalf("unknown draft should fail without a run record: %#v", notFoundEnvelope)
		}
		if testBridge.callCount() != 0 {
			t.Fatalf("strict or unknown draft failures must not call Gateway")
		}
	})

	t.Run("local console preflight reuses scoped workflow headers", func(t *testing.T) {
		server, _, draft := newWorkflowExecutorHTTPTestServer(t)
		request := httptest.NewRequest(
			http.MethodOptions,
			"/v1/user-workspace/workflow-drafts/"+draft.DraftID+"/runs",
			nil,
		)
		request.Header.Set("Origin", "http://127.0.0.1:4100")
		request.Header.Set("Access-Control-Request-Method", http.MethodPost)
		request.Header.Set("Access-Control-Request-Headers", savedWorkflowDraftDevWorkspaceHeader)
		response := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(response, request)

		if response.Code != http.StatusNoContent {
			t.Fatalf("executor CORS preflight should return 204, got %d: %s", response.Code, response.Body.String())
		}
		allowedHeaders := response.Header().Get("Access-Control-Allow-Headers")
		if !strings.Contains(allowedHeaders, savedWorkflowDraftDevWorkspaceHeader) ||
			!strings.Contains(allowedHeaders, savedWorkflowDraftDevApplicationHeader) {
			t.Fatalf("executor preflight is missing workflow scope headers: %s", allowedHeaders)
		}
	})
}

func newWorkflowExecutorHTTPTestServer(
	t *testing.T,
) (*Server, *workflowExecutorTestBridge, SavedWorkflowDraft) {
	t.Helper()
	server := NewServer(config.Config{
		ControlPlaneReadDevAuthEnabled:    true,
		WorkflowSavedDraftDevHTTPEnabled:  true,
		WorkflowSavedDraftDevWriteEnabled: true,
		WorkflowExecutorDevEnabled:        true,
		Provider:                          "mock",
	}, Options{BuildVersion: "test"})
	testBridge := &workflowExecutorTestBridge{}
	server.bridge = testBridge
	payload := validSavedWorkflowDraftPayload()
	result := server.savedWorkflowDraftService().SaveDraft(
		savedWorkflowDraftTestContext(),
		SaveWorkflowDraftRequest{ExpectedDraftVersion: 0, Payload: payload},
	)
	if result.FailureCode != "" || result.Draft == nil {
		t.Fatalf("prepare saved draft for executor HTTP test: %#v", result)
	}
	return server, testBridge, *result.Draft
}

func mustWorkflowRunJSON(t *testing.T, document any) []byte {
	t.Helper()
	body, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("marshal workflow run request: %v", err)
	}
	return body
}

func decodeWorkflowRunEnvelope(
	t *testing.T,
	response *httptest.ResponseRecorder,
	expectedStatus int,
) workflowRunEnvelope {
	t.Helper()
	if response.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d: %s", expectedStatus, response.Code, response.Body.String())
	}
	var envelope workflowRunEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode workflow run envelope: %v\n%s", err, response.Body.String())
	}
	return envelope
}
