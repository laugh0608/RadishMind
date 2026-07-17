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

func TestWorkflowRAGSnapshotHTTPRoutesRemainPreExecution(t *testing.T) {
	server := newWorkflowRAGSnapshotHTTPTestServer(t, true)
	created := createWorkflowRAGSnapshotOverHTTP(t, server)
	if created.Record == nil || created.Record.SnapshotVersion != 1 || len(created.Record.Fragments) != 2 {
		t.Fatalf("snapshot HTTP create response drifted: %#v", created)
	}

	listRequest := httptest.NewRequest(http.MethodGet,
		"/v1/user-workspace/workflow-retrieval-snapshots?workspace_id=workspace_demo&application_id=app_flow_copilot&lifecycle_state=active", nil)
	setSavedWorkflowDraftDevHeaders(listRequest, "workflow_rag_snapshots:read")
	listResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(listResponse, listRequest)
	listed := decodeWorkflowRAGSnapshotListEnvelope(t, listResponse, http.StatusOK)
	if listed.FailureCode != nil || len(listed.Items) != 1 || listed.Items[0].SnapshotID != created.Record.SnapshotID {
		t.Fatalf("snapshot HTTP list response drifted: %#v", listed)
	}
	if strings.Contains(listResponse.Body.String(), "official retrieval guidance") || strings.Contains(listResponse.Body.String(), "internal operator notes") {
		t.Fatalf("snapshot list leaked fragment content: %s", listResponse.Body.String())
	}

	readRequest := httptest.NewRequest(http.MethodGet,
		"/v1/user-workspace/workflow-retrieval-snapshots/"+created.Record.SnapshotID+"?workspace_id=workspace_demo&application_id=app_flow_copilot&snapshot_version=1", nil)
	setSavedWorkflowDraftDevHeaders(readRequest, "workflow_rag_snapshots:read")
	readResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(readResponse, readRequest)
	read := decodeWorkflowRAGSnapshotEnvelope(t, readResponse, http.StatusOK)
	if read.FailureCode != nil || read.Record == nil || len(read.Record.Fragments) != 2 {
		t.Fatalf("snapshot HTTP detail did not return the authorized exact version: %#v", read)
	}

	versionRequest := httptest.NewRequest(http.MethodPost,
		"/v1/user-workspace/workflow-retrieval-snapshots/"+created.Record.SnapshotID+"/versions",
		bytes.NewReader(mustWorkflowRAGJSON(t, workflowRAGSnapshotVersionBody{
			WorkspaceID: "workspace_demo", ApplicationID: "app_flow_copilot", ExpectedLatestVersion: 1,
			DisplayName: "Operator manual v2", ContentClassification: "workspace_internal",
			Fragments: []WorkflowRAGFragmentInput{{FragmentRef: "replacement", SourceType: "manual", SourceRef: "manual.v2", PageSlug: "manual/v2", Content: "replacement content"}},
		})))
	setSavedWorkflowDraftDevHeaders(versionRequest, "workflow_rag_snapshots:write")
	versionResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(versionResponse, versionRequest)
	versioned := decodeWorkflowRAGSnapshotEnvelope(t, versionResponse, http.StatusOK)
	if versioned.FailureCode != nil || versioned.Record == nil || versioned.Record.SnapshotVersion != 2 {
		t.Fatalf("snapshot HTTP version response drifted: %#v", versioned)
	}

	archiveRequest := httptest.NewRequest(http.MethodPost,
		"/v1/user-workspace/workflow-retrieval-snapshots/"+created.Record.SnapshotID+"/archive",
		strings.NewReader(`{"workspace_id":"workspace_demo","application_id":"app_flow_copilot","expected_latest_version":2}`))
	setSavedWorkflowDraftDevHeaders(archiveRequest, "workflow_rag_snapshots:archive")
	archiveResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(archiveResponse, archiveRequest)
	archived := decodeWorkflowRAGSnapshotEnvelope(t, archiveResponse, http.StatusOK)
	if archived.FailureCode != nil || archived.Record == nil || archived.Record.LifecycleState != workflowRAGSnapshotArchived {
		t.Fatalf("snapshot HTTP archive response drifted: %#v", archived)
	}

	runStore := server.workflowRunStore.(*memoryWorkflowRunStore)
	runStore.mu.RLock()
	defer runStore.mu.RUnlock()
	if len(runStore.records) != 0 {
		t.Fatalf("snapshot lifecycle created %d workflow runs", len(runStore.records))
	}
}

func TestWorkflowRAGSnapshotHTTPFailClosedBoundaries(t *testing.T) {
	t.Run("dedicated scope", func(t *testing.T) {
		server := newWorkflowRAGSnapshotHTTPTestServer(t, true)
		request := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-retrieval-snapshots",
			bytes.NewReader(mustWorkflowRAGJSON(t, workflowRAGSnapshotCreateBody{
				WorkspaceID: "workspace_demo", ApplicationID: "app_flow_copilot", SnapshotKey: "operator_manual",
				DisplayName: "Operator manual", ContentClassification: "workspace_internal", Fragments: workflowRAGTestFragments(),
			})))
		setSavedWorkflowDraftDevHeaders(request, "workflow_rag_snapshots:read")
		response := httptest.NewRecorder()
		server.httpServer.Handler.ServeHTTP(response, request)
		envelope := decodeWorkflowRAGSnapshotEnvelope(t, response, http.StatusOK)
		if envelope.FailureCode == nil || *envelope.FailureCode != WorkflowRAGFailureScopeDenied || envelope.Record != nil {
			t.Fatalf("snapshot write without dedicated scope did not fail closed: %#v", envelope)
		}
	})

	t.Run("strict JSON and query", func(t *testing.T) {
		server := newWorkflowRAGSnapshotHTTPTestServer(t, true)
		unknownBody := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-retrieval-snapshots",
			strings.NewReader(`{"workspace_id":"workspace_demo","application_id":"app_flow_copilot","snapshot_key":"manual","display_name":"Manual","content_classification":"public","fragments":[],"execute":true}`))
		setSavedWorkflowDraftDevHeaders(unknownBody, "workflow_rag_snapshots:write")
		bodyResponse := httptest.NewRecorder()
		server.httpServer.Handler.ServeHTTP(bodyResponse, unknownBody)
		if bodyResponse.Code != http.StatusBadRequest {
			t.Fatalf("unknown snapshot body field was accepted: %d %s", bodyResponse.Code, bodyResponse.Body.String())
		}
		unknownQuery := httptest.NewRequest(http.MethodGet,
			"/v1/user-workspace/workflow-retrieval-snapshots?workspace_id=workspace_demo&application_id=app_flow_copilot&include_content=true", nil)
		setSavedWorkflowDraftDevHeaders(unknownQuery, "workflow_rag_snapshots:read")
		queryResponse := httptest.NewRecorder()
		server.httpServer.Handler.ServeHTTP(queryResponse, unknownQuery)
		if queryResponse.Code != http.StatusBadRequest || strings.Contains(queryResponse.Body.String(), "fragment") {
			t.Fatalf("unknown snapshot query did not fail before content read: %d %s", queryResponse.Code, queryResponse.Body.String())
		}
	})

	t.Run("gate disabled", func(t *testing.T) {
		server := newWorkflowRAGSnapshotHTTPTestServer(t, false)
		request := httptest.NewRequest(http.MethodGet,
			"/v1/user-workspace/workflow-retrieval-snapshots?workspace_id=workspace_demo&application_id=app_flow_copilot", nil)
		setSavedWorkflowDraftDevHeaders(request, "workflow_rag_snapshots:read")
		response := httptest.NewRecorder()
		server.httpServer.Handler.ServeHTTP(response, request)
		if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), "WORKFLOW_RAG_SNAPSHOT_DEV_DISABLED") {
			t.Fatalf("disabled snapshot gate did not fail closed: %d %s", response.Code, response.Body.String())
		}
	})
}

func newWorkflowRAGSnapshotHTTPTestServer(t *testing.T, enabled bool) *Server {
	t.Helper()
	server := NewServer(config.Config{
		ControlPlaneReadDevAuthEnabled: true, WorkflowRAGSnapshotDevEnabled: enabled, Provider: "mock",
	}, Options{BuildVersion: "test"})
	t.Cleanup(server.Close)
	return server
}

func createWorkflowRAGSnapshotOverHTTP(t *testing.T, server *Server) workflowRAGSnapshotEnvelope {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-retrieval-snapshots",
		bytes.NewReader(mustWorkflowRAGJSON(t, workflowRAGSnapshotCreateBody{
			WorkspaceID: "workspace_demo", ApplicationID: "app_flow_copilot", SnapshotKey: "operator_manual",
			DisplayName: "Operator manual", ContentClassification: "workspace_internal", Fragments: workflowRAGTestFragments(),
		})))
	setSavedWorkflowDraftDevHeaders(request, "workflow_rag_snapshots:write")
	response := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(response, request)
	return decodeWorkflowRAGSnapshotEnvelope(t, response, http.StatusOK)
}

func mustWorkflowRAGJSON(t *testing.T, document any) []byte {
	t.Helper()
	payload, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("marshal workflow RAG document: %v", err)
	}
	return payload
}

func decodeWorkflowRAGSnapshotEnvelope(t *testing.T, response *httptest.ResponseRecorder, status int) workflowRAGSnapshotEnvelope {
	t.Helper()
	if response.Code != status {
		t.Fatalf("expected status %d, got %d: %s", status, response.Code, response.Body.String())
	}
	var envelope workflowRAGSnapshotEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode workflow RAG snapshot envelope: %v\n%s", err, response.Body.String())
	}
	return envelope
}

func decodeWorkflowRAGSnapshotListEnvelope(t *testing.T, response *httptest.ResponseRecorder, status int) workflowRAGSnapshotListEnvelope {
	t.Helper()
	if response.Code != status {
		t.Fatalf("expected status %d, got %d: %s", status, response.Code, response.Body.String())
	}
	var envelope workflowRAGSnapshotListEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode workflow RAG snapshot list envelope: %v\n%s", err, response.Body.String())
	}
	return envelope
}
