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

func TestWorkflowEvaluationRevisionHistoryPromotionConflictAndHistoricalReview(t *testing.T) {
	ctx := workflowExecutorTestContext()
	runs := newMemoryWorkflowRunStore(20)
	store := newMemoryWorkflowEvaluationStore(20)
	base := time.Now().UTC().Add(-time.Minute)
	seedEvaluationRun(t, runs, ctx, "run_revision_base", WorkflowRunStatusSucceeded, base)
	seedEvaluationRun(t, runs, ctx, "run_revision_failed", WorkflowRunStatusFailed, base.Add(time.Second))
	seedEvaluationRun(t, runs, ctx, "run_revision_next_base", WorkflowRunStatusSucceeded, base.Add(2*time.Second))
	service := newWorkflowEvaluationService(store, runs)
	service.newCaseID = func() (string, error) { return "eval_revision", nil }
	now := base.Add(time.Minute)
	service.now = func() time.Time { now = now.Add(time.Millisecond); return now }
	created := service.Create(ctx, WorkflowEvaluationCreateRequest{Name: "release regression", BaselineRunID: "run_revision_base", Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "run_revision_failed", ExpectedClassification: WorkflowRunComparisonRegression}}})
	if created.Case == nil || created.Case.Version != 1 || created.Case.RevisionKind != WorkflowEvaluationRevisionCreated {
		t.Fatalf("unexpected created case: %#v", created)
	}
	service.newCaseID = func() (string, error) { return "eval_newer_family", nil }
	newerFamily := service.Create(ctx, WorkflowEvaluationCreateRequest{Name: "newer family", BaselineRunID: "run_revision_base", Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "run_revision_failed", ExpectedClassification: WorkflowRunComparisonRegression}}})
	if newerFamily.Case == nil {
		t.Fatalf("create newer family: %#v", newerFamily)
	}
	revised := service.Revise(ctx, created.Case.CaseID, WorkflowEvaluationRevisionRequest{ExpectedVersion: 1, RevisionKind: WorkflowEvaluationRevisionCase, Name: "release expectation revised", BaselineRunID: "run_revision_base", Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "run_revision_failed", ExpectedClassification: WorkflowRunComparisonChanged}}})
	if revised.Case == nil || revised.Case.Version != 2 || revised.Case.PreviousVersion != 1 || strings.Join(revised.Case.ChangeCodes, ",") != "expectation_changed,name_changed" {
		t.Fatalf("unexpected revision: %#v", revised)
	}
	conflict := service.Revise(ctx, created.Case.CaseID, WorkflowEvaluationRevisionRequest{ExpectedVersion: 1, RevisionKind: WorkflowEvaluationRevisionCase, Name: "stale write", BaselineRunID: "run_revision_base", Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "run_revision_failed", ExpectedClassification: WorkflowRunComparisonRegression}}})
	if conflict.FailureCode != WorkflowEvaluationFailureVersionConflict || conflict.Case == nil || conflict.Case.Version != 2 {
		t.Fatalf("stale revision did not expose current version: %#v", conflict)
	}
	promoted := service.Revise(ctx, created.Case.CaseID, WorkflowEvaluationRevisionRequest{ExpectedVersion: 2, RevisionKind: WorkflowEvaluationRevisionBaselinePromotion, Name: "promoted baseline", BaselineRunID: "run_revision_next_base", Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "run_revision_failed", ExpectedClassification: WorkflowRunComparisonRegression}}})
	if promoted.Case == nil || promoted.Case.Version != 3 || promoted.Case.RevisionKind != WorkflowEvaluationRevisionBaselinePromotion || !containsString(promoted.Case.ChangeCodes, "baseline_changed") {
		t.Fatalf("promotion failed: %#v", promoted)
	}
	families := service.List(ctx, WorkflowEvaluationListRequest{Limit: 10})
	if len(families.Cases) != 2 || families.Cases[0].CaseID != newerFamily.Case.CaseID || families.Cases[1].CreatedAt != created.Case.CreatedAt {
		t.Fatalf("revision reordered family list: %#v", families)
	}
	page := service.ListRevisions(ctx, created.Case.CaseID, WorkflowEvaluationRevisionListRequest{Limit: 2})
	if len(page.Revisions) != 2 || !page.HasMore || page.Revisions[0].Version != 3 || page.NextCursor == "" {
		t.Fatalf("revision page failed: %#v", page)
	}
	next := service.ListRevisions(ctx, created.Case.CaseID, WorkflowEvaluationRevisionListRequest{Limit: 2, Cursor: page.NextCursor})
	if len(next.Revisions) != 1 || next.Revisions[0].Version != 1 {
		t.Fatalf("revision cursor failed: %#v", next)
	}
	if changed := service.ListRevisions(ctx, created.Case.CaseID, WorkflowEvaluationRevisionListRequest{Limit: 3, Cursor: page.NextCursor}); changed.FailureCode != WorkflowEvaluationFailureRevisionCursor {
		t.Fatalf("revision cursor was not bound: %#v", changed)
	}
	oldReview := service.ReviewVersion(ctx, created.Case.CaseID, 1)
	newReview := service.ReviewVersion(ctx, created.Case.CaseID, 2)
	if oldReview.Review == nil || oldReview.Review.Outcome != "passed" || oldReview.Review.Version != 1 || newReview.Review == nil || newReview.Review.Outcome != "mismatch" || newReview.Review.Version != 2 {
		t.Fatalf("historical review changed: old=%#v new=%#v", oldReview, newReview)
	}
}

func TestWorkflowEvaluationRevisionCASAllowsOneConcurrentWriter(t *testing.T) {
	ctx := workflowExecutorTestContext()
	runs := newMemoryWorkflowRunStore(10)
	store := newMemoryWorkflowEvaluationStore(10)
	seedEvaluationRun(t, runs, ctx, "run_cas_base", WorkflowRunStatusSucceeded, time.Now().UTC().Add(-time.Minute))
	seedEvaluationRun(t, runs, ctx, "run_cas_candidate", WorkflowRunStatusFailed, time.Now().UTC().Add(-time.Minute+time.Second))
	service := newWorkflowEvaluationService(store, runs)
	service.newCaseID = func() (string, error) { return "eval_cas", nil }
	created := service.Create(ctx, WorkflowEvaluationCreateRequest{Name: "CAS case", BaselineRunID: "run_cas_base", Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "run_cas_candidate", ExpectedClassification: WorkflowRunComparisonRegression}}})
	var wait sync.WaitGroup
	var mu sync.Mutex
	successes, conflicts := 0, 0
	for index := 0; index < 16; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			result := service.Revise(ctx, created.Case.CaseID, WorkflowEvaluationRevisionRequest{ExpectedVersion: 1, RevisionKind: WorkflowEvaluationRevisionCase, Name: "CAS case " + string(rune('a'+index)), BaselineRunID: "run_cas_base", Expectations: created.Case.Expectations})
			mu.Lock()
			defer mu.Unlock()
			if result.FailureCode == "" {
				successes++
			} else if result.FailureCode == WorkflowEvaluationFailureVersionConflict {
				conflicts++
			}
		}(index)
	}
	wait.Wait()
	if successes != 1 || conflicts != 15 {
		t.Fatalf("CAS results successes=%d conflicts=%d", successes, conflicts)
	}
}

func TestWorkflowEvaluationLegacyRecordUpgradesToVersionOne(t *testing.T) {
	ctx := workflowExecutorTestContext()
	legacy := WorkflowEvaluationCase{SchemaVersion: workflowEvaluationLegacySchema, CaseID: "eval_legacy", Name: "legacy case", WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, BaselineRunID: "run_base", Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "run_candidate", ExpectedClassification: WorkflowRunComparisonChanged}}, CreatedAt: workflowRunTimestamp(time.Now().UTC()), ActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef}
	payload, _ := json.Marshal(legacy)
	var decoded WorkflowEvaluationCase
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&decoded) != nil {
		t.Fatal("legacy decode failed")
	}
	decoded = upgradeWorkflowEvaluationCase(decoded)
	if decoded.Version != 1 || decoded.RevisionKind != WorkflowEvaluationRevisionCreated || validateWorkflowEvaluationCase(ctx, decoded) != nil {
		t.Fatalf("legacy upgrade failed: %#v", decoded)
	}
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
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

func TestWorkflowEvaluationHTTPRevisionHistoryAndConflict(t *testing.T) {
	server, _, draft := newWorkflowExecutorHTTPTestServer(t)
	ctx := workflowExecutorTestContext()
	seedEvaluationRun(t, server.workflowRunStore, ctx, "run_http_revision_base", WorkflowRunStatusSucceeded, time.Now().UTC().Add(-time.Minute))
	seedEvaluationRun(t, server.workflowRunStore, ctx, "run_http_revision_candidate", WorkflowRunStatusFailed, time.Now().UTC().Add(-time.Minute+time.Second))
	createBody := workflowEvaluationCreateHTTPBody{WorkspaceID: draft.WorkspaceID, ApplicationID: draft.ApplicationID, Name: "HTTP revision", BaselineRunID: "run_http_revision_base", Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "run_http_revision_candidate", ExpectedClassification: WorkflowRunComparisonRegression}}}
	raw, _ := json.Marshal(createBody)
	request := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-evaluation-cases", bytes.NewReader(raw))
	setSavedWorkflowDraftDevHeaders(request, "workflow_evaluations:write,workflow_evaluations:read,workflow_runs:read")
	response := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(response, request)
	var created workflowEvaluationEnvelope
	_ = json.Unmarshal(response.Body.Bytes(), &created)
	if created.Case == nil {
		t.Fatalf("create failed: %s", response.Body.String())
	}
	revisionBody := workflowEvaluationRevisionHTTPBody{WorkspaceID: draft.WorkspaceID, ApplicationID: draft.ApplicationID, ExpectedVersion: 1, RevisionKind: WorkflowEvaluationRevisionCase, Name: "HTTP revision changed", BaselineRunID: "run_http_revision_base", Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "run_http_revision_candidate", ExpectedClassification: WorkflowRunComparisonChanged}}}
	raw, _ = json.Marshal(revisionBody)
	revise := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-evaluation-cases/"+created.Case.CaseID+"/revisions", bytes.NewReader(raw))
	setSavedWorkflowDraftDevHeaders(revise, "workflow_evaluations:write,workflow_runs:read")
	revisedResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(revisedResponse, revise)
	var revised workflowEvaluationEnvelope
	_ = json.Unmarshal(revisedResponse.Body.Bytes(), &revised)
	if revised.Case == nil || revised.Case.Version != 2 {
		t.Fatalf("revise failed: %s", revisedResponse.Body.String())
	}
	raw, _ = json.Marshal(revisionBody)
	stale := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-evaluation-cases/"+created.Case.CaseID+"/revisions", bytes.NewReader(raw))
	setSavedWorkflowDraftDevHeaders(stale, "workflow_evaluations:write,workflow_runs:read")
	staleResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(staleResponse, stale)
	var conflict workflowEvaluationEnvelope
	_ = json.Unmarshal(staleResponse.Body.Bytes(), &conflict)
	if conflict.FailureCode == nil || *conflict.FailureCode != string(WorkflowEvaluationFailureVersionConflict) || conflict.Case == nil || conflict.Case.Version != 2 {
		t.Fatalf("conflict failed: %s", staleResponse.Body.String())
	}
	query := "?workspace_id=" + draft.WorkspaceID + "&application_id=" + draft.ApplicationID
	list := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/workflow-evaluation-cases/"+created.Case.CaseID+"/revisions"+query+"&limit=1", nil)
	setSavedWorkflowDraftDevHeaders(list, "workflow_evaluations:read,workflow_runs:read")
	listResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(listResponse, list)
	var history workflowEvaluationRevisionListEnvelope
	_ = json.Unmarshal(listResponse.Body.Bytes(), &history)
	if len(history.Revisions) != 1 || history.Revisions[0].Version != 2 || !history.HasMore {
		t.Fatalf("history failed: %s", listResponse.Body.String())
	}
	read := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/workflow-evaluation-cases/"+created.Case.CaseID+"/revisions/1"+query, nil)
	setSavedWorkflowDraftDevHeaders(read, "workflow_evaluations:read,workflow_runs:read")
	readResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(readResponse, read)
	var old workflowEvaluationEnvelope
	_ = json.Unmarshal(readResponse.Body.Bytes(), &old)
	if old.Case == nil || old.Case.Version != 1 {
		t.Fatalf("old revision failed: %s", readResponse.Body.String())
	}
	review := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/workflow-evaluation-cases/"+created.Case.CaseID+"/review"+query+"&version=1", nil)
	setSavedWorkflowDraftDevHeaders(review, "workflow_evaluations:read,workflow_runs:read")
	reviewResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(reviewResponse, review)
	var historical workflowEvaluationEnvelope
	_ = json.Unmarshal(reviewResponse.Body.Bytes(), &historical)
	if historical.Review == nil || historical.Review.Version != 1 || historical.Review.Outcome != "passed" {
		t.Fatalf("historical review failed: %s", reviewResponse.Body.String())
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
