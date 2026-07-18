package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
)

type workflowRAGPromotionHTTPAuthority struct {
	ctx         WorkflowRAGPromotionContext
	createBody  workflowRAGPromotionCandidateCreateBody
	datasetID   string
	reviewID    string
	draftID     string
	datasetText string
	fragment    string
}

func TestWorkflowRAGPromotionHTTPStrictPermissionsLifecycleAndZeroExecution(t *testing.T) {
	server := newWorkflowRAGPromotionHTTPTestServer(t, true)
	authority := seedWorkflowRAGPromotionHTTPAuthority(t, server)
	bridge := &workflowExecutorTestBridge{}
	server.bridge = bridge

	deniedRequest := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-rag-knowledge-promotion-candidates", bytes.NewReader(mustWorkflowRAGJSON(t, authority.createBody)))
	setWorkflowRAGPromotionHTTPHeaders(deniedRequest, "workflow_rag_promotions:write,workflow_rag_evaluation_datasets:read,application_drafts:read")
	deniedResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(deniedResponse, deniedRequest)
	denied := decodeWorkflowRAGPromotionEnvelope(t, deniedResponse, http.StatusOK)
	if denied.FailureCode == nil || *denied.FailureCode != WorkflowRAGPromotionFailureScopeDenied || denied.Candidate != nil {
		t.Fatalf("create accepted a missing snapshot read permission: %#v", denied)
	}

	unknownRequest := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-rag-knowledge-promotion-candidates",
		strings.NewReader(`{"workspace_id":"workspace_demo","application_id":"app_flow_copilot","unknown":true}`))
	setWorkflowRAGPromotionHTTPHeaders(unknownRequest, "workflow_rag_promotions:write,workflow_rag_evaluation_datasets:read,workflow_rag_snapshots:read,application_drafts:read")
	unknownResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(unknownResponse, unknownRequest)
	if unknownResponse.Code != http.StatusBadRequest || !strings.Contains(unknownResponse.Body.String(), "INVALID_JSON") {
		t.Fatalf("strict promotion JSON accepted unknown field: %d %s", unknownResponse.Code, unknownResponse.Body.String())
	}

	createRequest := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-rag-knowledge-promotion-candidates", bytes.NewReader(mustWorkflowRAGJSON(t, authority.createBody)))
	setWorkflowRAGPromotionHTTPHeaders(createRequest, "workflow_rag_promotions:write,workflow_rag_evaluation_datasets:read,workflow_rag_snapshots:read,application_drafts:read")
	createResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(createResponse, createRequest)
	created := decodeWorkflowRAGPromotionEnvelope(t, createResponse, http.StatusOK)
	if created.FailureCode != nil || created.Candidate == nil || created.Candidate.CandidateState != workflowRAGPromotionStatePending || created.CurrentRecordVersion != 1 || created.Binding != nil {
		t.Fatalf("create promotion candidate over HTTP: %#v", created)
	}
	for _, forbidden := range []string{authority.datasetText, authority.fragment, "promotion HTTP review note", "prompt", "Authorization:"} {
		if strings.Contains(createResponse.Body.String(), forbidden) {
			t.Fatalf("promotion create leaked %q: %s", forbidden, createResponse.Body.String())
		}
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/workflow-rag-knowledge-promotion-candidates?workspace_id=workspace_demo&application_id=app_flow_copilot", nil)
	setWorkflowRAGPromotionHTTPHeaders(listRequest, "workflow_rag_promotions:read")
	listResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(listResponse, listRequest)
	listed := decodeWorkflowRAGPromotionListEnvelope(t, listResponse, http.StatusOK)
	if listed.FailureCode != nil || len(listed.Items) != 1 || listed.Items[0].CandidateID != created.Candidate.CandidateID ||
		strings.Contains(listResponse.Body.String(), authority.datasetText) || strings.Contains(listResponse.Body.String(), authority.fragment) {
		t.Fatalf("metadata-only promotion list drifted: %#v body=%s", listed, listResponse.Body.String())
	}

	approveBody := workflowRAGPromotionDecisionBody{
		WorkspaceID: "workspace_demo", ApplicationID: "app_flow_copilot", ExpectedRecordVersion: 1,
		Decision: workflowRAGPromotionDecisionApprove, Reason: "人工确认 HTTP 权威证据",
	}
	approveRequest := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-rag-knowledge-promotion-candidates/"+created.Candidate.CandidateID+"/decisions", bytes.NewReader(mustWorkflowRAGJSON(t, approveBody)))
	setWorkflowRAGPromotionHTTPHeaders(approveRequest, "workflow_rag_promotions:review")
	approveResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(approveResponse, approveRequest)
	approved := decodeWorkflowRAGPromotionEnvelope(t, approveResponse, http.StatusOK)
	if approved.FailureCode != nil || approved.Candidate == nil || approved.Candidate.CandidateState != workflowRAGPromotionStateApproved ||
		approved.Binding == nil || !approved.Eligibility.Eligible || approved.CurrentRecordVersion != 2 {
		t.Fatalf("approve promotion candidate over HTTP: %#v", approved)
	}

	staleRequest := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-rag-knowledge-promotion-candidates/"+created.Candidate.CandidateID+"/decisions", bytes.NewReader(mustWorkflowRAGJSON(t, approveBody)))
	setWorkflowRAGPromotionHTTPHeaders(staleRequest, "workflow_rag_promotions:review")
	staleResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(staleResponse, staleRequest)
	stale := decodeWorkflowRAGPromotionEnvelope(t, staleResponse, http.StatusOK)
	if stale.FailureCode == nil || *stale.FailureCode != WorkflowRAGPromotionFailureRecordConflict || stale.CurrentRecordVersion != 2 || stale.CurrentState != workflowRAGPromotionStateApproved || stale.Candidate != nil || stale.Binding != nil {
		t.Fatalf("stale decision conflict was not metadata-only: %#v", stale)
	}

	detailRequest := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/workflow-rag-knowledge-promotion-candidates/"+created.Candidate.CandidateID+"?workspace_id=workspace_demo&application_id=app_flow_copilot", nil)
	setWorkflowRAGPromotionHTTPHeaders(detailRequest, "workflow_rag_promotions:read")
	detailResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(detailResponse, detailRequest)
	detail := decodeWorkflowRAGPromotionEnvelope(t, detailResponse, http.StatusOK)
	if detail.FailureCode != nil || detail.Candidate == nil || detail.Binding == nil || len(detail.Decisions) != 1 || !detail.Eligibility.Eligible {
		t.Fatalf("read approved promotion candidate over HTTP: %#v", detail)
	}

	if bridge.callCount() != 0 {
		t.Fatalf("promotion HTTP path called Gateway %d times", bridge.callCount())
	}
	runStore := server.workflowRunStore.(*memoryWorkflowRunStore)
	runStore.mu.RLock()
	defer runStore.mu.RUnlock()
	if len(runStore.records) != 0 {
		t.Fatalf("promotion HTTP path created %d workflow runs", len(runStore.records))
	}
}

func TestWorkflowRAGPromotionHTTPGateAndPermissionMappingAreIndependent(t *testing.T) {
	server := newWorkflowRAGPromotionHTTPTestServer(t, false)
	request := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/workflow-rag-knowledge-promotion-candidates?workspace_id=workspace_demo&application_id=app_flow_copilot", nil)
	setWorkflowRAGPromotionHTTPHeaders(request, "workflow_rag_promotions:read")
	response := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), "WORKFLOW_RAG_PROMOTION_DEV_DISABLED") {
		t.Fatalf("disabled promotion gate did not fail closed: %d %s", response.Code, response.Body.String())
	}
	for permission, scope := range map[string]string{
		"radishmind.workflow-rag-promotions.read":   "workflow_rag_promotions:read",
		"radishmind.workflow-rag-promotions.write":  "workflow_rag_promotions:write",
		"radishmind.workflow-rag-promotions.review": "workflow_rag_promotions:review",
		"radishmind.workflow-rag-promotions.bind":   "workflow_rag_promotions:bind",
	} {
		if controlPlaneReadPermissionGrants[permission] != scope {
			t.Fatalf("promotion permission mapping drifted: %s=%q", permission, controlPlaneReadPermissionGrants[permission])
		}
	}
}

func newWorkflowRAGPromotionHTTPTestServer(t *testing.T, enabled bool) *Server {
	t.Helper()
	server := NewServer(config.Config{
		ControlPlaneReadDevAuthEnabled: true, WorkflowRAGPromotionDevEnabled: enabled, Provider: "mock",
	}, Options{BuildVersion: "test"})
	t.Cleanup(server.Close)
	return server
}

func seedWorkflowRAGPromotionHTTPAuthority(t *testing.T, server *Server) workflowRAGPromotionHTTPAuthority {
	t.Helper()
	ctx := WorkflowRAGPromotionContext{
		RequestContext: workflowRAGTestContext().RequestContext, RequestID: "request_promotion_http_seed", TenantRef: "tenant_demo",
		WorkspaceID: "workspace_demo", ApplicationID: "app_flow_copilot", ActorRef: "subject_platform_ops",
		OwnerSubjectRef: "subject_platform_ops", AuditRef: "audit_promotion_http_seed", WriteEnabled: true,
	}
	snapshotContext := workflowRAGPromotionSnapshotContext(ctx)
	fragmentText := "promotion HTTP authoritative fragment"
	fragments := []WorkflowRAGFragmentInput{{
		FragmentRef: "promotion_http_authority", SourceType: "manual", SourceRef: "manual.promotion.http", PageSlug: "promotion/http",
		Title: "Promotion HTTP authority", IsOfficial: true, Content: fragmentText,
	}}
	baseline := createWorkflowRAGEvaluationTestSnapshot(t, server.workflowRAGSnapshotRepository, snapshotContext, "rags_cccccccccccccccc", "promotion_http_baseline", "public", fragments)
	candidate := createWorkflowRAGEvaluationTestSnapshot(t, server.workflowRAGSnapshotRepository, snapshotContext, "rags_dddddddddddddddd", "promotion_http_candidate", "public", fragments)

	evaluations := newWorkflowRAGEvaluationDatasetService(server.workflowRAGEvaluationDatasetRepository, server.workflowRAGSnapshotRepository)
	now := time.Date(2026, 7, 18, 10, 0, 0, 0, time.UTC)
	evaluations.now = func() time.Time { return now }
	evaluations.newID = workflowRAGEvaluationTestIDGenerator()
	queryText := "promotion HTTP authority query"
	dataset := evaluations.Create(snapshotContext, WorkflowRAGEvaluationDatasetCreateInput{
		DatasetKey: "promotion_http_authority", DisplayName: "Promotion HTTP authority", ContentClassification: "synthetic_public",
		BaselineSnapshot: workflowRAGEvaluationSnapshotBinding(baseline), Thresholds: workflowRAGEvaluationPerfectThresholds(),
		ReviewSummary: "Metadata-only HTTP promotion evidence.", Samples: []WorkflowRAGEvaluationSample{{
			SampleID: "promotion_http_authority", QueryText: queryText, Expectation: "evidence_required",
			ExpectedCitationRefs: []string{"promotion_http_authority"}, RequiredOfficialRefs: []string{"promotion_http_authority"}, TopK: 1,
			ReviewNote: "promotion HTTP review note",
		}},
	})
	if dataset.FailureCode != "" || dataset.Resource == nil || dataset.Version == nil {
		t.Fatalf("seed promotion HTTP dataset: %#v", dataset)
	}
	now = now.Add(time.Minute)
	review := evaluations.CreateCandidateReview(snapshotContext, dataset.Resource.DatasetID, WorkflowRAGCandidateReviewInput{
		DatasetVersion: 1, DatasetDigest: dataset.Version.Dataset.DatasetDigest, CandidateSnapshot: workflowRAGEvaluationSnapshotBinding(candidate),
	})
	if review.FailureCode != "" || review.Review == nil || review.Review.Candidate.Status != "passed" || review.Review.Conclusion != "unchanged" {
		t.Fatalf("seed promotion HTTP review: %#v", review)
	}

	draftService := newApplicationConfigurationDraftService(server.applicationDraftRepository)
	draftService.now = func() time.Time { return now }
	payload := validApplicationDraftPayload()
	payload.BaseApplicationUpdatedAt = "2026-05-31T10:20:00Z"
	draftContext := workflowRAGPromotionDraftContext(ctx)
	draftContext.WriteEnabled = true
	draft := draftService.Save(draftContext, payload, 0)
	if draft.FailureCode != "" || draft.Draft == nil {
		t.Fatalf("seed promotion HTTP draft: %#v", draft)
	}
	return workflowRAGPromotionHTTPAuthority{
		ctx: ctx, datasetID: dataset.Resource.DatasetID, reviewID: review.Review.ReviewID, draftID: draft.Draft.DraftID,
		datasetText: queryText, fragment: fragmentText,
		createBody: workflowRAGPromotionCandidateCreateBody{
			WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, DatasetID: dataset.Resource.DatasetID,
			DatasetVersion: 1, DatasetDigest: dataset.Version.Dataset.DatasetDigest, CandidateReviewID: review.Review.ReviewID,
			DraftID: draft.Draft.DraftID, ExpectedDraftVersion: 1,
		},
	}
}

func setWorkflowRAGPromotionHTTPHeaders(request *http.Request, scopes string) {
	setSavedWorkflowDraftDevHeaders(request, scopes)
	request.Header.Set(controlPlaneReadDevSubjectHeader, "subject_platform_ops")
}

func decodeWorkflowRAGPromotionEnvelope(t *testing.T, response *httptest.ResponseRecorder, status int) workflowRAGPromotionEnvelope {
	t.Helper()
	if response.Code != status {
		t.Fatalf("expected status %d, got %d: %s", status, response.Code, response.Body.String())
	}
	var envelope workflowRAGPromotionEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode workflow RAG promotion envelope: %v\n%s", err, response.Body.String())
	}
	return envelope
}

func decodeWorkflowRAGPromotionListEnvelope(t *testing.T, response *httptest.ResponseRecorder, status int) workflowRAGPromotionListEnvelope {
	t.Helper()
	if response.Code != status {
		t.Fatalf("expected status %d, got %d: %s", status, response.Code, response.Body.String())
	}
	var envelope workflowRAGPromotionListEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode workflow RAG promotion list envelope: %v\n%s", err, response.Body.String())
	}
	return envelope
}
