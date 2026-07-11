package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestWorkflowEvaluationSuiteReviewDecisionPaginationAndScope(t *testing.T) {
	ctx := workflowExecutorTestContext()
	runs := newMemoryWorkflowRunStore(20)
	cases := newMemoryWorkflowEvaluationStore(20)
	base := time.Now().UTC().Add(-time.Minute)
	seedEvaluationRun(t, runs, ctx, "suite_base", WorkflowRunStatusSucceeded, base)
	seedEvaluationRun(t, runs, ctx, "suite_failed", WorkflowRunStatusFailed, base.Add(time.Second))
	evaluation := newWorkflowEvaluationService(cases, runs)
	caseCounter := 0
	evaluation.newCaseID = func() (string, error) { caseCounter++; return "eval_suite_" + string(rune('a'+caseCounter-1)), nil }
	passed := evaluation.Create(ctx, WorkflowEvaluationCreateRequest{Name: "known regression", BaselineRunID: "suite_base", Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "suite_failed", ExpectedClassification: WorkflowRunComparisonRegression}}})
	mismatch := evaluation.Create(ctx, WorkflowEvaluationCreateRequest{Name: "unexpected improvement", BaselineRunID: "suite_base", Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "suite_failed", ExpectedClassification: WorkflowRunComparisonImprovement}}})
	if passed.Case == nil || mismatch.Case == nil {
		t.Fatalf("seed cases failed: passed=%#v mismatch=%#v", passed, mismatch)
	}

	service := newWorkflowEvaluationSuiteService(newMemoryWorkflowEvaluationSuiteStore(10), evaluation)
	suiteCounter, decisionCounter := 0, 0
	service.newSuiteID = func() (string, error) { suiteCounter++; return "suite_" + string(rune('a'+suiteCounter-1)), nil }
	service.newDecisionID = func() (string, error) {
		decisionCounter++
		return "decision_" + string(rune('a'+decisionCounter-1)), nil
	}
	now := base.Add(time.Minute)
	service.now = func() time.Time { now = now.Add(time.Millisecond); return now }
	first := service.Create(ctx, WorkflowEvaluationSuiteCreateRequest{Name: "release review", CaseRefs: []WorkflowEvaluationSuiteCaseRef{{CaseID: passed.Case.CaseID, Version: 1}, {CaseID: mismatch.Case.CaseID, Version: 1}}})
	if first.Suite == nil || first.FailureCode != "" {
		t.Fatalf("create suite: %#v", first)
	}
	review := service.Review(ctx, first.Suite.SuiteID)
	if review.Review == nil || review.Review.Outcome != "mismatch" || review.Review.Passed != 1 || review.Review.Mismatch != 1 || len(review.Review.ReviewDigest) != 64 {
		t.Fatalf("review: %#v", review)
	}
	if changed := service.Decide(ctx, first.Suite.SuiteID, WorkflowEvaluationDecisionRequest{Decision: "needs_review", ReviewDigest: strings.Repeat("0", 64)}); changed.FailureCode != WorkflowEvaluationSuiteFailureReviewChanged {
		t.Fatalf("stale digest accepted: %#v", changed)
	}
	if approved := service.Decide(ctx, first.Suite.SuiteID, WorkflowEvaluationDecisionRequest{Decision: "approved", ReviewDigest: review.Review.ReviewDigest}); approved.FailureCode != WorkflowEvaluationSuiteFailureApprovalBlocked {
		t.Fatalf("unsafe approval accepted: %#v", approved)
	}
	decided := service.Decide(ctx, first.Suite.SuiteID, WorkflowEvaluationDecisionRequest{Decision: "needs_review", ReviewDigest: review.Review.ReviewDigest})
	if decided.Decision == nil || decided.Suite == nil || decided.Suite.CurrentDecisionVersion != 1 || decided.Suite.CurrentDecision != "needs_review" {
		t.Fatalf("decision: %#v", decided)
	}
	if history := service.ListDecisions(ctx, first.Suite.SuiteID, WorkflowEvaluationDecisionListRequest{}); len(history.Decisions) != 1 || history.Decisions[0].ReviewDigest != review.Review.ReviewDigest {
		t.Fatalf("decision history: %#v", history)
	}

	second := service.Create(ctx, WorkflowEvaluationSuiteCreateRequest{Name: "passed release", CaseRefs: []WorkflowEvaluationSuiteCaseRef{{CaseID: passed.Case.CaseID, Version: 1}}})
	page := service.List(ctx, WorkflowEvaluationSuiteListRequest{Limit: 1})
	if second.Suite == nil || len(page.Suites) != 1 || !page.HasMore || page.Suites[0].SuiteID != second.Suite.SuiteID {
		t.Fatalf("suite page: %#v", page)
	}
	next := service.List(ctx, WorkflowEvaluationSuiteListRequest{Limit: 1, Cursor: page.NextCursor})
	if len(next.Suites) != 1 || next.Suites[0].SuiteID != first.Suite.SuiteID {
		t.Fatalf("suite next page: %#v", next)
	}
	if rebound := service.List(ctx, WorkflowEvaluationSuiteListRequest{Limit: 2, Cursor: page.NextCursor}); rebound.FailureCode != WorkflowEvaluationSuiteFailureCursor {
		t.Fatalf("cursor was not bound: %#v", rebound)
	}
	other := ctx
	other.WorkspaceID = "other_workspace"
	if leaked := service.Read(other, first.Suite.SuiteID); leaked.FailureCode != WorkflowEvaluationSuiteFailureNotFound {
		t.Fatalf("cross-scope suite leaked: %#v", leaked)
	}
}

func TestWorkflowEvaluationSuiteDecisionCASAllowsOneWriter(t *testing.T) {
	ctx, evaluation, caseRef := newWorkflowEvaluationSuiteTestFixture(t)
	service := newWorkflowEvaluationSuiteService(newMemoryWorkflowEvaluationSuiteStore(10), evaluation)
	created := service.Create(ctx, WorkflowEvaluationSuiteCreateRequest{Name: "concurrent decision", CaseRefs: []WorkflowEvaluationSuiteCaseRef{caseRef}})
	review := service.Review(ctx, created.Suite.SuiteID)
	var wait sync.WaitGroup
	var mu sync.Mutex
	successes, conflicts := 0, 0
	for index := 0; index < 16; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result := service.Decide(ctx, created.Suite.SuiteID, WorkflowEvaluationDecisionRequest{ExpectedDecisionVersion: 0, Decision: "approved", ReviewDigest: review.Review.ReviewDigest})
			mu.Lock()
			defer mu.Unlock()
			if result.FailureCode == "" {
				successes++
			} else if result.FailureCode == WorkflowEvaluationSuiteFailureDecisionConflict {
				conflicts++
			}
		}()
	}
	wait.Wait()
	if successes != 1 || conflicts != 15 {
		t.Fatalf("CAS successes=%d conflicts=%d", successes, conflicts)
	}
}

func TestWorkflowEvaluationSuiteHTTPHistoryReviewAndDecision(t *testing.T) {
	server, _, draft := newWorkflowExecutorHTTPTestServer(t)
	ctx := workflowExecutorTestContext()
	seedEvaluationRun(t, server.workflowRunStore, ctx, "suite_http_base", WorkflowRunStatusSucceeded, time.Now().UTC().Add(-time.Minute))
	seedEvaluationRun(t, server.workflowRunStore, ctx, "suite_http_failed", WorkflowRunStatusFailed, time.Now().UTC().Add(-time.Minute+time.Second))
	caseResult := server.workflowEvaluationService().Create(ctx, WorkflowEvaluationCreateRequest{Name: "HTTP suite case", BaselineRunID: "suite_http_base", Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "suite_http_failed", ExpectedClassification: WorkflowRunComparisonRegression}}})
	createBody := workflowEvaluationSuiteCreateHTTPBody{WorkspaceID: draft.WorkspaceID, ApplicationID: draft.ApplicationID, Name: "HTTP release review", CaseRefs: []WorkflowEvaluationSuiteCaseRef{{CaseID: caseResult.Case.CaseID, Version: 1}}}
	created := workflowEvaluationSuiteRequest(t, server, http.MethodPost, "/v1/user-workspace/workflow-evaluation-suites", createBody, "workflow_evaluations:write,workflow_evaluations:read,workflow_runs:read")
	if created.Suite == nil || created.FailureCode != nil {
		t.Fatalf("create suite HTTP: %#v", created)
	}
	query := "?workspace_id=" + draft.WorkspaceID + "&application_id=" + draft.ApplicationID
	reviewed := workflowEvaluationSuiteRequest(t, server, http.MethodGet, "/v1/user-workspace/workflow-evaluation-suites/"+created.Suite.SuiteID+"/review"+query, nil, "workflow_evaluations:read,workflow_runs:read")
	if reviewed.Review == nil || reviewed.Review.Outcome != "passed" {
		t.Fatalf("review suite HTTP: %#v", reviewed)
	}
	decisionBody := workflowEvaluationDecisionHTTPBody{WorkspaceID: draft.WorkspaceID, ApplicationID: draft.ApplicationID, ExpectedDecisionVersion: 0, Decision: "approved", ReviewDigest: reviewed.Review.ReviewDigest}
	decided := workflowEvaluationSuiteRequest(t, server, http.MethodPost, "/v1/user-workspace/workflow-evaluation-suites/"+created.Suite.SuiteID+"/decisions", decisionBody, "workflow_evaluations:write,workflow_runs:read")
	if decided.Decision == nil || decided.Decision.Decision != "approved" {
		t.Fatalf("decision HTTP: %#v", decided)
	}
	list := workflowEvaluationSuiteListRequest(t, server, "/v1/user-workspace/workflow-evaluation-suites"+query)
	if len(list.Suites) != 1 || list.Suites[0].CurrentDecision != "approved" {
		t.Fatalf("suite history HTTP: %#v", list)
	}
	for _, forbidden := range []string{"input_text", "condition_values", "credential", "endpoint", "provider_raw_envelope"} {
		raw, _ := json.Marshal([]any{created, reviewed, decided, list})
		if strings.Contains(string(raw), forbidden) {
			t.Fatalf("suite HTTP leaked %s", forbidden)
		}
	}
	invalid := workflowEvaluationSuiteRequest(t, server, http.MethodGet, "/v1/user-workspace/workflow-evaluation-suites/"+created.Suite.SuiteID+"/review"+query+"&execute=true", nil, "workflow_evaluations:read,workflow_runs:read")
	if invalid.FailureCode == nil || *invalid.FailureCode != string(WorkflowEvaluationSuiteFailureInvalid) {
		t.Fatalf("unknown query accepted: %#v", invalid)
	}
}

func newWorkflowEvaluationSuiteTestFixture(t *testing.T) (WorkflowRunContext, workflowEvaluationService, WorkflowEvaluationSuiteCaseRef) {
	t.Helper()
	ctx := workflowExecutorTestContext()
	runs := newMemoryWorkflowRunStore(10)
	seedEvaluationRun(t, runs, ctx, "suite_fixture_base", WorkflowRunStatusSucceeded, time.Now().UTC().Add(-time.Minute))
	seedEvaluationRun(t, runs, ctx, "suite_fixture_failed", WorkflowRunStatusFailed, time.Now().UTC().Add(-time.Minute+time.Second))
	evaluation := newWorkflowEvaluationService(newMemoryWorkflowEvaluationStore(10), runs)
	created := evaluation.Create(ctx, WorkflowEvaluationCreateRequest{Name: "fixture regression", BaselineRunID: "suite_fixture_base", Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "suite_fixture_failed", ExpectedClassification: WorkflowRunComparisonRegression}}})
	return ctx, evaluation, WorkflowEvaluationSuiteCaseRef{CaseID: created.Case.CaseID, Version: created.Case.Version}
}

func workflowEvaluationSuiteRequest(t *testing.T, server *Server, method, target string, body any, scopes string) workflowEvaluationSuiteEnvelope {
	t.Helper()
	var request *http.Request
	if body == nil {
		request = httptest.NewRequest(method, target, nil)
	} else {
		raw, _ := json.Marshal(body)
		request = httptest.NewRequest(method, target, bytes.NewReader(raw))
	}
	setSavedWorkflowDraftDevHeaders(request, scopes)
	response := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(response, request)
	var result workflowEvaluationSuiteEnvelope
	if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &result) != nil {
		t.Fatalf("suite request %s: %d %s", target, response.Code, response.Body.String())
	}
	return result
}

func workflowEvaluationSuiteListRequest(t *testing.T, server *Server, target string) workflowEvaluationSuiteListEnvelope {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, target, nil)
	setSavedWorkflowDraftDevHeaders(request, "workflow_evaluations:read,workflow_runs:read")
	response := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(response, request)
	var result workflowEvaluationSuiteListEnvelope
	if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &result) != nil {
		t.Fatalf("suite list %s: %d %s", target, response.Code, response.Body.String())
	}
	return result
}
