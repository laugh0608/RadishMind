package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWorkflowRAGRunV3ComparisonHTTPRouteIsMetadataOnly(t *testing.T) {
	server, _, _ := newWorkflowExecutorHTTPTestServer(t)
	runStore, ok := server.workflowRunStore.(*memoryWorkflowRunStore)
	if !ok {
		t.Fatalf("unexpected workflow run store: %T", server.workflowRunStore)
	}
	snapshotRepository := newMemoryWorkflowRAGSnapshotRepository(&runStore.mu)
	baselineService, _, runContext, _, draft := workflowRAGExecutionStoreFixture(t, runStore, snapshotRepository, "run_httpragbase00001")
	baseline := baselineService.Execute(runContext, WorkflowRAGExecutionRequest{
		DraftID: draft.DraftID, DraftVersion: draft.DraftVersion, InputText: "official retrieval guidance", Model: "mock-rag",
	})
	candidateService, _, _, _, _ := workflowRAGExecutionStoreFixtureFromExisting(t, runStore, snapshotRepository, draft, "run_httpragcand00001")
	candidate := candidateService.Execute(runContext, WorkflowRAGExecutionRequest{
		DraftID: draft.DraftID, DraftVersion: draft.DraftVersion, InputText: "official retrieval guidance", Model: "mock-rag",
	})
	if baseline.Record == nil || candidate.Record == nil {
		t.Fatalf("seed HTTP RAG comparison runs: baseline=%#v candidate=%#v", baseline, candidate)
	}
	target := "/v1/user-workspace/workflow-runs/" + candidate.Record.RunID + "/comparison?baseline_run_id=" + baseline.Record.RunID +
		"&workspace_id=" + draft.WorkspaceID + "&application_id=" + draft.ApplicationID
	request := httptest.NewRequest(http.MethodGet, target, nil)
	setSavedWorkflowDraftDevHeaders(request, "workflow_runs:read")
	response := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(response, request)
	var envelope workflowRunComparisonEnvelope
	if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &envelope) != nil || envelope.FailureCode != nil ||
		envelope.Comparison == nil || envelope.Comparison.SchemaVersion != workflowRAGRunComparisonSchemaVersion || envelope.Comparison.Retrieval == nil {
		t.Fatalf("unexpected RAG comparison HTTP response: %d %s", response.Code, response.Body.String())
	}
	for _, forbidden := range []string{"official retrieval guidance", "fragment_content", "prompt_packet", "raw_response", "credential", "\"answer\""} {
		if strings.Contains(response.Body.String(), forbidden) {
			t.Fatalf("RAG comparison HTTP response leaked %q: %s", forbidden, response.Body.String())
		}
	}
}

func TestWorkflowRAGRunV3ComparisonEvaluationBaselineAndSuite(t *testing.T) {
	fixture := newWorkflowRAGExecutionFixture(t)
	runCounter := 0
	fixture.service.newRunID = func() (string, error) {
		runCounter++
		if runCounter == 1 {
			return "run_compare000000001", nil
		}
		return "run_compare000000002", nil
	}
	now := time.Date(2026, 7, 18, 8, 0, 0, 0, time.UTC)
	fixture.service.now = func() time.Time {
		now = now.Add(time.Millisecond)
		return now
	}
	baseline := fixture.service.Execute(fixture.runContext, workflowRAGExecutionFixtureRequest(fixture, "official retrieval guidance", fixture.draft.DraftVersion))
	candidate := fixture.service.Execute(fixture.runContext, workflowRAGExecutionFixtureRequest(fixture, "official retrieval guidance", fixture.draft.DraftVersion))
	if baseline.FailureCode != "" || baseline.Record == nil || candidate.FailureCode != "" || candidate.Record == nil {
		t.Fatalf("create comparable v3 fixtures: baseline=%#v candidate=%#v", baseline, candidate)
	}

	comparison := newWorkflowExecutorService(nil, nil, fixture.runStore).CompareRuns(fixture.runContext, baseline.Record.RunID, candidate.Record.RunID)
	if comparison.FailureCode != "" || comparison.Comparison == nil || comparison.Comparison.SchemaVersion != workflowRAGRunComparisonSchemaVersion ||
		comparison.Comparison.Retrieval == nil || comparison.Comparison.Retrieval.RunProfile != workflowRAGComparisonProfile ||
		comparison.Comparison.Classification != WorkflowRunComparisonUnchanged {
		t.Fatalf("run comparison did not accept matching v3 records: %#v", comparison)
	}
	serialized, err := json.Marshal(comparison.Comparison)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"official retrieval guidance", "Use the official guidance", "fragment_content", "prompt_packet", "raw_response", "credential"} {
		if strings.Contains(string(serialized), forbidden) {
			t.Fatalf("RAG comparison leaked %q: %s", forbidden, serialized)
		}
	}

	evaluationStore := newMemoryWorkflowEvaluationStore(10)
	evaluation := newWorkflowEvaluationService(evaluationStore, fixture.runStore)
	evaluation.newCaseID = func() (string, error) { return "eval_rag_profile", nil }
	created := evaluation.Create(fixture.runContext, WorkflowEvaluationCreateRequest{
		Name: "retrieval regression profile", BaselineRunID: baseline.Record.RunID,
		Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: candidate.Record.RunID, ExpectedClassification: WorkflowRunComparisonUnchanged}},
	})
	if created.FailureCode != "" || created.Case == nil {
		t.Fatalf("evaluation case rejected matching v3 records: %#v", created)
	}
	review := evaluation.Review(fixture.runContext, created.Case.CaseID)
	if review.FailureCode != "" || review.Review == nil || review.Review.Outcome != "passed" ||
		review.Review.RunProfile != workflowRAGComparisonProfile || review.Review.Items[0].ComparisonSchemaVersion != workflowRAGRunComparisonSchemaVersion {
		t.Fatalf("evaluation review did not preserve retrieval profile: %#v", review)
	}

	suite := newWorkflowEvaluationSuiteService(newMemoryWorkflowEvaluationSuiteStore(10), evaluation)
	suite.newSuiteID = func() (string, error) { return "suite_rag_profile", nil }
	suiteResult := suite.Create(fixture.runContext, WorkflowEvaluationSuiteCreateRequest{
		Name: "retrieval suite", CaseRefs: []WorkflowEvaluationSuiteCaseRef{{CaseID: created.Case.CaseID, Version: 1}},
	})
	if suiteResult.FailureCode != "" || suiteResult.Suite == nil {
		t.Fatalf("evaluation suite rejected retrieval case: %#v", suiteResult)
	}
	suiteReview := suite.Review(fixture.runContext, suiteResult.Suite.SuiteID)
	if suiteReview.FailureCode != "" || suiteReview.Review == nil || suiteReview.Review.Outcome != "passed" ||
		suiteReview.Review.Items[0].RunProfile != workflowRAGComparisonProfile || len(suiteReview.Review.ReviewDigest) != 64 {
		t.Fatalf("suite review did not include retrieval profile: %#v", suiteReview)
	}
}

func TestWorkflowRAGRunV3RejectsMixedAndDifferentBindings(t *testing.T) {
	fixture := newWorkflowRAGExecutionFixture(t)
	executed := fixture.service.Execute(fixture.runContext, workflowRAGExecutionFixtureRequest(fixture, "official retrieval guidance", fixture.draft.DraftVersion))
	if executed.FailureCode != "" || executed.Record == nil {
		t.Fatalf("create v3 fixture: %#v", executed)
	}
	v1 := terminalComparisonTestRun(fixture.runContext, "run_nonrag_baseline", WorkflowRunStatusSucceeded, time.Now().UTC().Add(-time.Minute))
	storeTerminalComparisonTestRun(t, fixture.runStore, fixture.runContext, &v1)
	comparisonService := newWorkflowExecutorService(nil, nil, fixture.runStore)
	if result := comparisonService.CompareRuns(fixture.runContext, v1.RunID, executed.Record.RunID); result.FailureCode != WorkflowRunFailureRetrievalIncompatible || result.Comparison != nil {
		t.Fatalf("mixed run profiles were accepted: %#v", result)
	}

	incompatible := cloneWorkflowRunRecord(*executed.Record)
	incompatible.RunID = "run_incompat00000000"
	incompatible.RecordVersion = 0
	incompatible.Status = WorkflowRunStatusRunning
	incompatible.CompletedAt = ""
	incompatible.RetrievalAttempt.QueryDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	incompatible.Diagnostic.TerminalWriteState = WorkflowRunTerminalWritePending
	if err := fixture.runStore.UpsertRun(fixture.runContext, &incompatible); err != nil {
		t.Fatalf("store incompatible running record: %v", err)
	}
	incompatible.Status = WorkflowRunStatusSucceeded
	startedAt, _ := time.Parse(time.RFC3339Nano, incompatible.StartedAt)
	incompatible.CompletedAt = workflowRunTimestamp(startedAt.Add(time.Second))
	incompatible.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
	if err := fixture.runStore.UpsertRun(fixture.runContext, &incompatible); err != nil {
		t.Fatalf("store incompatible terminal record: %v", err)
	}
	if result := comparisonService.CompareRuns(fixture.runContext, executed.Record.RunID, incompatible.RunID); result.FailureCode != WorkflowRunFailureRetrievalIncompatible {
		t.Fatalf("different query binding was accepted: %#v", result)
	}

	evaluation := newWorkflowEvaluationService(newMemoryWorkflowEvaluationStore(10), fixture.runStore)
	created := evaluation.Create(fixture.runContext, WorkflowEvaluationCreateRequest{
		Name: "incompatible retrieval bindings", BaselineRunID: executed.Record.RunID,
		Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: incompatible.RunID, ExpectedClassification: WorkflowRunComparisonChanged}},
	})
	if created.FailureCode != WorkflowEvaluationFailureRetrievalIncompatible || created.Case != nil {
		t.Fatalf("evaluation accepted incompatible v3 records: %#v", created)
	}
}

func TestWorkflowRAGComparisonTreatsCitationLossAsRegression(t *testing.T) {
	fixture := newWorkflowRAGExecutionFixture(t)
	executed := fixture.service.Execute(fixture.runContext, workflowRAGExecutionFixtureRequest(fixture, "official retrieval guidance", fixture.draft.DraftVersion))
	if executed.Record == nil {
		t.Fatalf("create v3 fixture: %#v", executed)
	}
	baseline := cloneWorkflowRunRecord(*executed.Record)
	candidate := cloneWorkflowRunRecord(*executed.Record)
	second := workflowRAGRunSelectedFragment{FragmentRef: "internal_notes", ContentDigest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Rank: 2, SourceType: "wiki"}
	baseline.RetrievalAttempt.SelectedFragments = append(baseline.RetrievalAttempt.SelectedFragments, second)
	candidate.RetrievalAttempt.SelectedFragments = append(candidate.RetrievalAttempt.SelectedFragments, second)
	baseline.RetrievalAttempt.CitationRefs = []string{"official_guide", "internal_notes"}
	candidate.RetrievalAttempt.CitationRefs = []string{"official_guide"}
	comparison := buildWorkflowRAGRunComparison(baseline, candidate, time.Now().UTC())
	if comparison.Classification != WorkflowRunComparisonRegression || comparison.Retrieval == nil ||
		len(comparison.Retrieval.CitationRemovedRefs) != 1 || comparison.Retrieval.CitationRemovedRefs[0] != "internal_notes" ||
		comparison.RecommendedReviewAction != WorkflowRunReviewRetrievalEvidence {
		t.Fatalf("citation loss was not classified as regression: %#v", comparison)
	}
}
