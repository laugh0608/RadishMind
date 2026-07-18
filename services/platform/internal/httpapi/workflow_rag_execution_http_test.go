package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
)

func TestWorkflowRAGExecutionHTTPUsesIndependentGateAndStrictVerifiedScope(t *testing.T) {
	server, testBridge, draft := newWorkflowRAGExecutionHTTPTestServer(t, true)
	body := workflowRAGExecutionHTTPBody{
		WorkspaceID: draft.WorkspaceID, ApplicationID: draft.ApplicationID, DraftVersion: draft.DraftVersion,
		InputText: "official retrieval guidance", Model: "mock-rag",
	}
	request := workflowRAGExecutionHTTPRequest(t, draft.DraftID, body)
	setSavedWorkflowDraftDevHeaders(request, strings.Join(workflowRAGExecutionRequiredScopes, ","))
	response := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(response, request)
	envelope := decodeWorkflowRAGExecutionEnvelope(t, response, http.StatusOK)
	if server.config.WorkflowExecutorDevEnabled || envelope.FailureCode != nil || envelope.Run == nil ||
		envelope.Run.Status != WorkflowRunStatusSucceeded || envelope.RetrievalAnswer == nil || testBridge.handleCalls != 1 {
		t.Fatalf("independent retrieval execution gate did not complete one execution: %#v calls=%d", envelope, testBridge.handleCalls)
	}
	for _, forbidden := range []string{body.InputText, "official retrieval guidance", "用户问题"} {
		if strings.Contains(response.Body.String(), forbidden) {
			t.Fatalf("execution response leaked query or prompt packet %q: %s", forbidden, response.Body.String())
		}
	}

	for _, missing := range workflowRAGExecutionRequiredScopes {
		t.Run("missing "+missing, func(t *testing.T) {
			before := testBridge.handleCalls
			request := workflowRAGExecutionHTTPRequest(t, draft.DraftID, body)
			scopes := make([]string, 0, len(workflowRAGExecutionRequiredScopes)-1)
			for _, scope := range workflowRAGExecutionRequiredScopes {
				if scope != missing {
					scopes = append(scopes, scope)
				}
			}
			setSavedWorkflowDraftDevHeaders(request, strings.Join(scopes, ","))
			response := httptest.NewRecorder()
			server.httpServer.Handler.ServeHTTP(response, request)
			envelope := decodeWorkflowRAGExecutionEnvelope(t, response, http.StatusOK)
			if envelope.FailureCode == nil || *envelope.FailureCode != WorkflowRAGFailureScopeDenied ||
				envelope.Run != nil || testBridge.handleCalls != before {
				t.Fatalf("missing scope %s was not denied before execution: %#v", missing, envelope)
			}
		})
	}

	t.Run("three-level binding mismatch", func(t *testing.T) {
		before := testBridge.handleCalls
		request := workflowRAGExecutionHTTPRequest(t, draft.DraftID, body)
		setSavedWorkflowDraftDevHeaders(request, strings.Join(workflowRAGExecutionRequiredScopes, ","))
		request.Header.Set(savedWorkflowDraftDevApplicationHeader, "application_other")
		response := httptest.NewRecorder()
		server.httpServer.Handler.ServeHTTP(response, request)
		envelope := decodeWorkflowRAGExecutionEnvelope(t, response, http.StatusOK)
		if envelope.FailureCode == nil || *envelope.FailureCode != WorkflowRAGFailureScopeDenied || testBridge.handleCalls != before {
			t.Fatalf("application binding mismatch was not denied: %#v", envelope)
		}
	})

	t.Run("actor must be verified", func(t *testing.T) {
		before := testBridge.handleCalls
		request := workflowRAGExecutionHTTPRequest(t, draft.DraftID, body)
		response := httptest.NewRecorder()
		server.httpServer.Handler.ServeHTTP(response, request)
		envelope := decodeWorkflowRAGExecutionEnvelope(t, response, http.StatusOK)
		if envelope.FailureCode == nil || *envelope.FailureCode != WorkflowRAGFailureScopeDenied || testBridge.handleCalls != before {
			t.Fatalf("unverified actor reached retrieval execution: %d %s", response.Code, response.Body.String())
		}
	})
}

func TestWorkflowRAGExecutionHTTPRejectsForbiddenClientAuthorityAndDisabledGate(t *testing.T) {
	server, testBridge, draft := newWorkflowRAGExecutionHTTPTestServer(t, true)
	for _, forbidden := range []string{"selected_fragments", "ranking", "citation_whitelist", "snapshot_digest", "profile_digest"} {
		t.Run(forbidden, func(t *testing.T) {
			payload := map[string]any{
				"workspace_id": draft.WorkspaceID, "application_id": draft.ApplicationID,
				"draft_version": draft.DraftVersion, "input_text": "official retrieval guidance", "model": "mock-rag",
				forbidden: "client-authority-is-forbidden",
			}
			request := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-drafts/"+draft.DraftID+"/retrieval-executions", bytes.NewReader(mustWorkflowRAGJSON(t, payload)))
			setSavedWorkflowDraftDevHeaders(request, strings.Join(workflowRAGExecutionRequiredScopes, ","))
			response := httptest.NewRecorder()
			server.httpServer.Handler.ServeHTTP(response, request)
			if response.Code != http.StatusBadRequest || testBridge.handleCalls != 0 {
				t.Fatalf("strict JSON accepted forbidden field %s: %d %s", forbidden, response.Code, response.Body.String())
			}
		})
	}

	disabled, disabledBridge, disabledDraft := newWorkflowRAGExecutionHTTPTestServer(t, false)
	request := workflowRAGExecutionHTTPRequest(t, disabledDraft.DraftID, workflowRAGExecutionHTTPBody{
		WorkspaceID: disabledDraft.WorkspaceID, ApplicationID: disabledDraft.ApplicationID,
		DraftVersion: disabledDraft.DraftVersion, InputText: "official retrieval guidance", Model: "mock-rag",
	})
	setSavedWorkflowDraftDevHeaders(request, strings.Join(workflowRAGExecutionRequiredScopes, ","))
	response := httptest.NewRecorder()
	disabled.httpServer.Handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), "WORKFLOW_RAG_EXECUTION_DEV_DISABLED") || disabledBridge.handleCalls != 0 {
		t.Fatalf("disabled retrieval execution gate did not fail closed: %d %s", response.Code, response.Body.String())
	}
}

func TestWorkflowRAGRunHistoryV3ReadsMetadataAndAuthorizedRepositoryPreview(t *testing.T) {
	server, _, draft := newWorkflowRAGExecutionHTTPTestServer(t, true)
	execute := workflowRAGExecutionHTTPRequest(t, draft.DraftID, workflowRAGExecutionHTTPBody{
		WorkspaceID: draft.WorkspaceID, ApplicationID: draft.ApplicationID, DraftVersion: draft.DraftVersion,
		InputText: "official retrieval guidance", Model: "mock-rag",
	})
	setSavedWorkflowDraftDevHeaders(execute, strings.Join(workflowRAGExecutionRequiredScopes, ","))
	executeResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(executeResponse, execute)
	execution := decodeWorkflowRAGExecutionEnvelope(t, executeResponse, http.StatusOK)
	if execution.Run == nil {
		t.Fatalf("execution fixture did not return a run: %#v", execution)
	}

	detailURL := "/v1/user-workspace/workflow-runs/" + execution.Run.RunID + "?workspace_id=workspace_demo&application_id=app_flow_copilot"
	detail := httptest.NewRequest(http.MethodGet, detailURL, nil)
	setSavedWorkflowDraftDevHeaders(detail, "workflow_runs:read")
	detailResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(detailResponse, detail)
	detailEnvelope := decodeWorkflowRAGExecutionEnvelope(t, detailResponse, http.StatusOK)
	if detailEnvelope.Run == nil || detailEnvelope.Run.SchemaVersion != workflowRunRecordRAGSchemaVersion ||
		detailEnvelope.Run.RAGAnswer != nil || len(detailEnvelope.RetrievalFragmentPreviews) != 0 || detailEnvelope.RetrievalAnswer != nil {
		t.Fatalf("ordinary v3 history detail is not metadata-only: %#v", detailEnvelope)
	}
	if strings.Contains(detailResponse.Body.String(), "official retrieval guidance") {
		t.Fatalf("ordinary history leaked query or fragment body: %s", detailResponse.Body.String())
	}

	list := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/workflow-runs?workspace_id=workspace_demo&application_id=app_flow_copilot", nil)
	setSavedWorkflowDraftDevHeaders(list, "workflow_runs:read")
	listResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(listResponse, list)
	listEnvelope := decodeWorkflowRunListEnvelope(t, listResponse, http.StatusOK)
	if len(listEnvelope.Runs) != 1 || listEnvelope.Runs[0].SchemaVersion != workflowRunRecordRAGSchemaVersion ||
		listEnvelope.Runs[0].SnapshotDigest == "" || listEnvelope.Runs[0].RetrievalProfileDigest == "" ||
		listEnvelope.Runs[0].QueryDigest == "" || len(listEnvelope.Runs[0].SelectedFragments) != 1 ||
		len(listEnvelope.Runs[0].CitationRefs) != 1 {
		t.Fatalf("ordinary v3 history list did not expose stable metadata: %#v", listEnvelope.Runs)
	}
	if strings.Contains(listResponse.Body.String(), "official retrieval guidance") {
		t.Fatalf("ordinary history list leaked query or fragment body: %s", listResponse.Body.String())
	}

	preview := httptest.NewRequest(http.MethodGet, detailURL+"&include_retrieval_fragment_previews=true", nil)
	setSavedWorkflowDraftDevHeaders(preview, "workflow_runs:read,workflow_rag_snapshots:read")
	previewResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(previewResponse, preview)
	previewEnvelope := decodeWorkflowRAGExecutionEnvelope(t, previewResponse, http.StatusOK)
	if len(previewEnvelope.RetrievalFragmentPreviews) != 1 ||
		len([]rune(previewEnvelope.RetrievalFragmentPreviews[0].Preview)) > workflowRAGHistoryPreviewCharacters {
		t.Fatalf("authorized immutable snapshot preview is invalid: %#v", previewEnvelope.RetrievalFragmentPreviews)
	}

	denied := httptest.NewRequest(http.MethodGet, detailURL+"&include_retrieval_fragment_previews=true", nil)
	setSavedWorkflowDraftDevHeaders(denied, "workflow_runs:read")
	deniedResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(deniedResponse, denied)
	deniedEnvelope := decodeWorkflowRAGExecutionEnvelope(t, deniedResponse, http.StatusOK)
	if deniedEnvelope.FailureCode == nil || *deniedEnvelope.FailureCode != string(WorkflowRunFailureScopeDenied) || len(deniedEnvelope.RetrievalFragmentPreviews) != 0 {
		t.Fatalf("fragment preview without snapshot read scope was accepted: %#v", deniedEnvelope)
	}
}

func newWorkflowRAGExecutionHTTPTestServer(t *testing.T, enabled bool) (*Server, *fakeBridge, SavedWorkflowDraft) {
	t.Helper()
	server := NewServer(config.Config{
		ControlPlaneReadDevAuthEnabled: true, WorkflowSavedDraftDevHTTPEnabled: true,
		WorkflowSavedDraftDevWriteEnabled: true, WorkflowRAGExecutionDevEnabled: enabled,
		Provider: "mock", Model: "mock-rag",
	}, Options{BuildVersion: "rag-execution-test", TestOnly: true})
	t.Cleanup(server.Close)
	testBridge := &fakeBridge{envelope: bridge.GatewayEnvelope{
		Status: "ok", Response: map[string]any{"structured_answer": workflowRAGTestAnswer("official_guide", "Official guidance supports the answer.")},
	}}
	server.bridge = testBridge
	snapshotContext := workflowRAGTestContext()
	created := server.workflowRAGSnapshotService().Create(snapshotContext, WorkflowRAGSnapshotCreateInput{
		SnapshotKey: "http_execution_manual", DisplayName: "HTTP execution manual",
		ContentClassification: "workspace_internal", Fragments: workflowRAGTestFragments(),
	})
	if created.FailureCode != "" || created.Record == nil {
		t.Fatalf("create HTTP execution snapshot: %#v", created)
	}
	draft := workflowRAGEligibleDraft(created.Record.RAGRef)
	draftContext := SavedWorkflowDraftContext{
		RequestContext: context.Background(), RequestID: "request_rag_draft_seed", TenantRef: snapshotContext.TenantRef,
		WorkspaceID: draft.WorkspaceID, ApplicationID: draft.ApplicationID, ActorRef: snapshotContext.ActorRef,
		OwnerSubjectRef: snapshotContext.ActorRef, ScopeGrants: []string{"workflow_drafts:read", "workflow_drafts:write"},
		AuditRef: "audit_rag_draft_seed", WriteEnabled: true,
	}
	saved := server.savedWorkflowDraftService().SaveDraft(draftContext, SaveWorkflowDraftRequest{Payload: savedWorkflowDraftPayloadFromDraft(draft)})
	if saved.FailureCode != "" || saved.Draft == nil {
		t.Fatalf("save HTTP execution draft: %#v", saved)
	}
	return server, testBridge, *saved.Draft
}

func workflowRAGExecutionHTTPRequest(t *testing.T, draftID string, body workflowRAGExecutionHTTPBody) *http.Request {
	t.Helper()
	return httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-drafts/"+draftID+"/retrieval-executions", bytes.NewReader(mustWorkflowRAGJSON(t, body)))
}

func decodeWorkflowRAGExecutionEnvelope(t *testing.T, response *httptest.ResponseRecorder, expectedStatus int) workflowRunEnvelope {
	t.Helper()
	if response.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d: %s", expectedStatus, response.Code, response.Body.String())
	}
	var envelope workflowRunEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode workflow RAG execution envelope: %v\n%s", err, response.Body.String())
	}
	return envelope
}
