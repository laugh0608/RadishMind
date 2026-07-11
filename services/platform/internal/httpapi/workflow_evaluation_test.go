package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWorkflowEvaluationCaseLifecyclePaginationAndBatchReview(t *testing.T) {
	ctx := workflowExecutorTestContext()
	runStore := newMemoryWorkflowRunStore(20)
	evaluationStore := newMemoryWorkflowEvaluationStore(20)
	base := time.Now().UTC().Add(-time.Minute)
	seedEvaluationRun(t, runStore, ctx, "run_base", WorkflowRunStatusSucceeded, base)
	seedEvaluationRun(t, runStore, ctx, "run_failed", WorkflowRunStatusFailed, base.Add(time.Second))
	seedEvaluationRun(t, runStore, ctx, "run_success", WorkflowRunStatusSucceeded, base.Add(2*time.Second))
	service := newWorkflowEvaluationService(evaluationStore, runStore)
	counter := 0
	service.newCaseID = func() (string, error) { counter++; return "eval_" + string(rune('a'+counter-1)), nil }
	now := base.Add(time.Minute)
	service.now = func() time.Time { now = now.Add(time.Millisecond); return now }
	first := service.Create(ctx, WorkflowEvaluationCreateRequest{Name: "release candidate review", BaselineRunID: "run_base", Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "run_failed", ExpectedClassification: WorkflowRunComparisonRegression}, {CandidateRunID: "run_success", ExpectedClassification: WorkflowRunComparisonUnchanged}}})
	if first.FailureCode != "" || first.Case == nil {
		t.Fatalf("create failed: %#v", first)
	}
	review := service.Review(ctx, first.Case.CaseID)
	if review.FailureCode != "" || review.Review == nil || review.Review.Outcome != "passed" || review.Review.Matched != 2 {
		t.Fatalf("unexpected batch review: %#v", review)
	}
	second := service.Create(ctx, WorkflowEvaluationCreateRequest{Name: "expected mismatch", BaselineRunID: "run_base", Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "run_failed", ExpectedClassification: WorkflowRunComparisonImprovement}}})
	if second.FailureCode != "" {
		t.Fatalf("second create failed: %#v", second)
	}
	page := service.List(ctx, WorkflowEvaluationListRequest{Limit: 1})
	if len(page.Cases) != 1 || !page.HasMore || page.NextCursor == "" || page.Cases[0].CaseID != second.Case.CaseID {
		t.Fatalf("bad first page: %#v", page)
	}
	next := service.List(ctx, WorkflowEvaluationListRequest{Limit: 1, Cursor: page.NextCursor})
	if len(next.Cases) != 1 || next.Cases[0].CaseID != first.Case.CaseID {
		t.Fatalf("bad next page: %#v", next)
	}
	if changed := service.List(ctx, WorkflowEvaluationListRequest{Limit: 2, Cursor: page.NextCursor}); changed.FailureCode != WorkflowEvaluationFailureCursorInvalid {
		t.Fatalf("cursor was not filter-bound: %#v", changed)
	}
	if mismatch := service.Review(ctx, second.Case.CaseID); mismatch.Review == nil || mismatch.Review.Outcome != "mismatch" {
		t.Fatalf("mismatch not reported: %#v", mismatch)
	}
	other := ctx
	other.TenantRef = "tenant_other"
	if leaked := service.Read(other, first.Case.CaseID); leaked.FailureCode != WorkflowEvaluationFailureNotFound {
		t.Fatalf("cross-scope case leaked: %#v", leaked)
	}
}

func TestWorkflowEvaluationRejectsUnsafeOrIneligibleDefinitions(t *testing.T) {
	ctx := workflowExecutorTestContext()
	runs := newMemoryWorkflowRunStore(10)
	store := newMemoryWorkflowEvaluationStore(10)
	running := workflowRunHistoryTestRecord(ctx, "run_running", "draft", time.Now().UTC())
	if err := runs.UpsertRun(ctx, &running); err != nil {
		t.Fatal(err)
	}
	seedEvaluationRun(t, runs, ctx, "run_done", WorkflowRunStatusSucceeded, time.Now().UTC().Add(-time.Second))
	service := newWorkflowEvaluationService(store, runs)
	requests := []WorkflowEvaluationCreateRequest{{Name: "secret=private", BaselineRunID: "run_done", Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "run_running", ExpectedClassification: WorkflowRunComparisonChanged}}}, {Name: "duplicate", BaselineRunID: "run_done", Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "run_done", ExpectedClassification: WorkflowRunComparisonChanged}}}, {Name: "fresh running", BaselineRunID: "run_done", Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "run_running", ExpectedClassification: WorkflowRunComparisonChanged}}}}
	for index, request := range requests {
		result := service.Create(ctx, request)
		if index < 2 && result.FailureCode != WorkflowEvaluationFailureInvalid {
			t.Fatalf("invalid request accepted: %#v", result)
		}
		if index == 2 && result.FailureCode != WorkflowEvaluationFailureRunNotEligible {
			t.Fatalf("running request accepted: %#v", result)
		}
	}
}

func TestWorkflowEvaluationHTTPCreateListDetailAndReview(t *testing.T) {
	server, _, draft := newWorkflowExecutorHTTPTestServer(t)
	ctx := workflowExecutorTestContext()
	seedEvaluationRun(t, server.workflowRunStore, ctx, "run_http_base", WorkflowRunStatusSucceeded, time.Now().UTC().Add(-time.Minute))
	seedEvaluationRun(t, server.workflowRunStore, ctx, "run_http_candidate", WorkflowRunStatusFailed, time.Now().UTC().Add(-time.Minute+time.Second))
	body := workflowEvaluationCreateHTTPBody{WorkspaceID: draft.WorkspaceID, ApplicationID: draft.ApplicationID, Name: "HTTP regression batch", BaselineRunID: "run_http_base", Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "run_http_candidate", ExpectedClassification: WorkflowRunComparisonRegression}}}
	raw, _ := json.Marshal(body)
	request := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-evaluation-cases", bytes.NewReader(raw))
	setSavedWorkflowDraftDevHeaders(request, "workflow_evaluations:write,workflow_evaluations:read,workflow_runs:read")
	response := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(response, request)
	var created workflowEvaluationEnvelope
	if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &created) != nil || created.FailureCode != nil || created.Case == nil {
		t.Fatalf("create HTTP failed: %d %s", response.Code, response.Body.String())
	}
	headers := func(request *http.Request) {
		setSavedWorkflowDraftDevHeaders(request, "workflow_evaluations:read,workflow_runs:read")
	}
	query := "?workspace_id=" + draft.WorkspaceID + "&application_id=" + draft.ApplicationID
	listRequest := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/workflow-evaluation-cases"+query, nil)
	headers(listRequest)
	listResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(listResponse, listRequest)
	var list workflowEvaluationListEnvelope
	if json.Unmarshal(listResponse.Body.Bytes(), &list) != nil || len(list.Cases) != 1 {
		t.Fatalf("list failed: %s", listResponse.Body.String())
	}
	reviewRequest := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/workflow-evaluation-cases/"+created.Case.CaseID+"/review"+query, nil)
	headers(reviewRequest)
	reviewResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(reviewResponse, reviewRequest)
	var reviewed workflowEvaluationEnvelope
	if json.Unmarshal(reviewResponse.Body.Bytes(), &reviewed) != nil || reviewed.Review == nil || reviewed.Review.Outcome != "passed" {
		t.Fatalf("review failed: %s", reviewResponse.Body.String())
	}
	for _, forbidden := range []string{"input_text", "input_bytes", "condition_values", "output_preview", "credential", "endpoint", "provider_raw_envelope"} {
		if strings.Contains(reviewResponse.Body.String(), forbidden) || strings.Contains(response.Body.String(), forbidden) {
			t.Fatalf("HTTP leaked %s", forbidden)
		}
	}
	strict := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/workflow-evaluation-cases/"+created.Case.CaseID+"/review"+query+"&execute=true", nil)
	headers(strict)
	strictResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(strictResponse, strict)
	var invalid workflowEvaluationEnvelope
	_ = json.Unmarshal(strictResponse.Body.Bytes(), &invalid)
	if invalid.FailureCode == nil || *invalid.FailureCode != string(WorkflowEvaluationFailureInvalid) {
		t.Fatalf("unknown review query accepted: %s", strictResponse.Body.String())
	}
}

func seedEvaluationRun(t *testing.T, store workflowRunStore, ctx WorkflowRunContext, id string, status WorkflowRunStatus, started time.Time) {
	t.Helper()
	record := terminalComparisonTestRun(ctx, id, status, started)
	if status == WorkflowRunStatusFailed {
		record.FailureCode = WorkflowRunFailureGatewayFailed
		setWorkflowRunFailureDiagnostic(&record, record.FailureCode, "node_model", WorkflowRunGatewayFailureTimeout)
		record.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
	}
	storeTerminalComparisonTestRun(t, store, ctx, &record)
}
