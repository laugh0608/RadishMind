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

func TestWorkflowRAGEvaluationHTTPStrictPermissionsAndCandidateReview(t *testing.T) {
	server := newWorkflowRAGEvaluationHTTPTestServer(t, true)
	baseline := seedWorkflowRAGEvaluationHTTPSnapshot(t, server, "baseline_http", "public", []WorkflowRAGFragmentInput{{FragmentRef: "official_http", SourceType: "manual", SourceRef: "manual.http", PageSlug: "http/official", Title: "HTTP official", IsOfficial: true, Content: "http candidate review evidence"}})
	candidate := seedWorkflowRAGEvaluationHTTPSnapshot(t, server, "candidate_http", "public", []WorkflowRAGFragmentInput{{FragmentRef: "candidate_http", SourceType: "wiki", SourceRef: "wiki.http", PageSlug: "http/candidate", Title: "HTTP candidate", Content: "unrelated candidate content"}})
	body := workflowRAGEvaluationDatasetCreateBody{
		WorkspaceID: "workspace_demo", ApplicationID: "app_flow_copilot", DatasetKey: "http_review", DisplayName: "HTTP review", ContentClassification: "synthetic_public",
		BaselineSnapshot: workflowRAGEvaluationSnapshotBinding(baseline), Thresholds: workflowRAGEvaluationPerfectThresholds(), ReviewSummary: "Reviewed HTTP evidence.",
		Samples: []WorkflowRAGEvaluationSample{{SampleID: "http_evidence", QueryText: "http candidate review evidence", Expectation: "evidence_required", ExpectedCitationRefs: []string{"official_http"}, RequiredOfficialRefs: []string{"official_http"}, TopK: 1, ReviewNote: "Official HTTP evidence required."}},
	}

	deniedRequest := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-rag-evaluation-datasets", bytes.NewReader(mustWorkflowRAGJSON(t, body)))
	setSavedWorkflowDraftDevHeaders(deniedRequest, "workflow_rag_evaluation_datasets:write")
	deniedResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(deniedResponse, deniedRequest)
	denied := decodeWorkflowRAGEvaluationEnvelope(t, deniedResponse, http.StatusOK)
	if denied.FailureCode == nil || *denied.FailureCode != WorkflowRAGEvaluationFailureScopeDenied {
		t.Fatalf("dataset create accepted missing snapshot read permission: %#v", denied)
	}

	unknownRequest := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-rag-evaluation-datasets", strings.NewReader(`{"workspace_id":"workspace_demo","application_id":"app_flow_copilot","unknown":true}`))
	setSavedWorkflowDraftDevHeaders(unknownRequest, "workflow_rag_evaluation_datasets:write,workflow_rag_snapshots:read")
	unknownResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(unknownResponse, unknownRequest)
	if unknownResponse.Code != http.StatusBadRequest || !strings.Contains(unknownResponse.Body.String(), "INVALID_JSON") {
		t.Fatalf("strict JSON accepted unknown field: %d %s", unknownResponse.Code, unknownResponse.Body.String())
	}

	createRequest := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-rag-evaluation-datasets", bytes.NewReader(mustWorkflowRAGJSON(t, body)))
	setSavedWorkflowDraftDevHeaders(createRequest, "workflow_rag_evaluation_datasets:write,workflow_rag_snapshots:read")
	createResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(createResponse, createRequest)
	created := decodeWorkflowRAGEvaluationEnvelope(t, createResponse, http.StatusOK)
	if created.FailureCode != nil || created.Resource == nil || created.Version == nil {
		t.Fatalf("create dataset over HTTP: %#v", created)
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/workflow-rag-evaluation-datasets?workspace_id=workspace_demo&application_id=app_flow_copilot&lifecycle_state=active", nil)
	setSavedWorkflowDraftDevHeaders(listRequest, "workflow_rag_evaluation_datasets:read")
	listResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(listResponse, listRequest)
	if listResponse.Code != http.StatusOK || strings.Contains(listResponse.Body.String(), "http candidate review evidence") {
		t.Fatalf("dataset list failed or leaked query: %d %s", listResponse.Code, listResponse.Body.String())
	}

	readDeniedRequest := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/workflow-rag-evaluation-datasets/"+created.Resource.DatasetID+"?workspace_id=workspace_demo&application_id=app_flow_copilot&dataset_version=1", nil)
	setSavedWorkflowDraftDevHeaders(readDeniedRequest, "workflow_rag_evaluation_datasets:read")
	readDeniedResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(readDeniedResponse, readDeniedRequest)
	readDenied := decodeWorkflowRAGEvaluationEnvelope(t, readDeniedResponse, http.StatusOK)
	if readDenied.FailureCode == nil || *readDenied.FailureCode != WorkflowRAGEvaluationFailureScopeDenied || strings.Contains(readDeniedResponse.Body.String(), "http candidate review evidence") {
		t.Fatalf("content detail permission did not fail closed: %#v body=%s", readDenied, readDeniedResponse.Body.String())
	}

	reviewBody := workflowRAGCandidateReviewCreateBody{WorkspaceID: "workspace_demo", ApplicationID: "app_flow_copilot", DatasetVersion: 1, DatasetDigest: created.Version.Dataset.DatasetDigest, CandidateSnapshot: workflowRAGEvaluationSnapshotBinding(candidate)}
	reviewRequest := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-rag-evaluation-datasets/"+created.Resource.DatasetID+"/candidate-reviews", bytes.NewReader(mustWorkflowRAGJSON(t, reviewBody)))
	setSavedWorkflowDraftDevHeaders(reviewRequest, "workflow_rag_evaluation_datasets:review,workflow_rag_evaluation_datasets:read,workflow_rag_snapshots:read")
	reviewResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(reviewResponse, reviewRequest)
	reviewed := decodeWorkflowRAGEvaluationEnvelope(t, reviewResponse, http.StatusOK)
	if reviewed.FailureCode != nil || reviewed.Review == nil || reviewed.Review.Conclusion != "regressed" {
		t.Fatalf("candidate review over HTTP: %#v", reviewed)
	}
	for _, forbidden := range []string{"http candidate review evidence", "unrelated candidate content", "Official HTTP evidence required."} {
		if strings.Contains(reviewResponse.Body.String(), forbidden) {
			t.Fatalf("candidate review HTTP response leaked %q: %s", forbidden, reviewResponse.Body.String())
		}
	}
}

func TestWorkflowRAGEvaluationHTTPGateIsIndependent(t *testing.T) {
	server := newWorkflowRAGEvaluationHTTPTestServer(t, false)
	request := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/workflow-rag-evaluation-datasets?workspace_id=workspace_demo&application_id=app_flow_copilot", nil)
	setSavedWorkflowDraftDevHeaders(request, "workflow_rag_evaluation_datasets:read")
	response := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), "WORKFLOW_RAG_EVALUATION_DEV_DISABLED") {
		t.Fatalf("disabled evaluation gate did not fail closed: %d %s", response.Code, response.Body.String())
	}
}

func newWorkflowRAGEvaluationHTTPTestServer(t *testing.T, enabled bool) *Server {
	t.Helper()
	server := NewServer(config.Config{ControlPlaneReadDevAuthEnabled: true, WorkflowRAGEvaluationDevEnabled: enabled, Provider: "mock"}, Options{BuildVersion: "test"})
	t.Cleanup(server.Close)
	return server
}

func seedWorkflowRAGEvaluationHTTPSnapshot(t *testing.T, server *Server, key, classification string, fragments []WorkflowRAGFragmentInput) WorkflowRAGSnapshotRecord {
	t.Helper()
	ctx := workflowRAGTestContext()
	ctx.RequestID, ctx.AuditRef = "request_"+key, "audit_"+key
	result := newWorkflowRAGSnapshotService(server.workflowRAGSnapshotRepository).Create(ctx, WorkflowRAGSnapshotCreateInput{SnapshotKey: key, DisplayName: strings.ReplaceAll(key, "_", " "), ContentClassification: classification, Fragments: fragments})
	if result.FailureCode != "" || result.Record == nil {
		t.Fatalf("seed HTTP snapshot: %#v", result)
	}
	return *result.Record
}

func decodeWorkflowRAGEvaluationEnvelope(t *testing.T, response *httptest.ResponseRecorder, status int) workflowRAGEvaluationEnvelope {
	t.Helper()
	if response.Code != status {
		t.Fatalf("expected status %d, got %d: %s", status, response.Code, response.Body.String())
	}
	var envelope workflowRAGEvaluationEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode workflow RAG evaluation envelope: %v\n%s", err, response.Body.String())
	}
	return envelope
}
