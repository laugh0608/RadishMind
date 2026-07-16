package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWorkflowRunComparisonClassificationAndSanitizedNodeDiff(t *testing.T) {
	ctx := workflowExecutorTestContext()
	now := time.Now().UTC()
	baseline := terminalComparisonTestRun(ctx, "run_baseline", WorkflowRunStatusSucceeded, now)
	baseline.Nodes = []WorkflowRunNodeRecord{{NodeID: "node_prompt", NodeType: "prompt", Status: WorkflowRunNodeStatusSucceeded, DurationMS: 10, OutputPreview: "private baseline"}}
	candidate := terminalComparisonTestRun(ctx, "run_candidate", WorkflowRunStatusFailed, now.Add(time.Second))
	candidate.FailureCode = WorkflowRunFailureGatewayFailed
	setWorkflowRunFailureDiagnostic(&candidate, candidate.FailureCode, "node_model", WorkflowRunGatewayFailureTimeout)
	candidate.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
	candidate.Nodes = []WorkflowRunNodeRecord{{NodeID: "node_prompt", NodeType: "prompt", Status: WorkflowRunNodeStatusSucceeded, DurationMS: 15, OutputPreview: "private candidate"}, {NodeID: "node_model", NodeType: "llm", Status: WorkflowRunNodeStatusFailed, DurationMS: 20, FailureCode: WorkflowRunFailureGatewayFailed}}

	comparison := buildWorkflowRunComparison(baseline, candidate, now.Add(2*time.Second))
	if comparison.Classification != WorkflowRunComparisonRegression || comparison.ComparisonState != WorkflowRunComparisonComparable || len(comparison.Nodes) != 2 || comparison.Nodes[0].Change != "changed" || comparison.Nodes[1].Change != "added" {
		t.Fatalf("unexpected regression comparison: %#v", comparison)
	}
	document, err := json.Marshal(comparison)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"private baseline", "private candidate", "output_preview", "input_bytes", "condition_node_ids", "actor_ref"} {
		if strings.Contains(string(document), forbidden) {
			t.Fatalf("comparison leaked %q: %s", forbidden, document)
		}
	}

	legacy := baseline
	legacy.SchemaVersion = workflowRunRecordLegacySchemaVersion
	legacy.Diagnostic = nil
	partial := buildWorkflowRunComparison(legacy, candidate, now.Add(2*time.Second))
	if partial.ComparisonState != WorkflowRunComparisonLegacyPartial {
		t.Fatalf("legacy comparison was not partial: %#v", partial)
	}

	running := candidate
	running.Status, running.CompletedAt = WorkflowRunStatusRunning, ""
	running.StartedAt = workflowRunTimestamp(now)
	running.Diagnostic.TerminalWriteState = WorkflowRunTerminalWritePending
	inconclusive := buildWorkflowRunComparison(baseline, running, now.Add(time.Second))
	if inconclusive.Classification != WorkflowRunComparisonInconclusive || inconclusive.ComparisonState != WorkflowRunComparisonRunningInconclusive {
		t.Fatalf("running comparison should be inconclusive: %#v", inconclusive)
	}
}

func TestWorkflowRunComparisonServiceScopeValidationAndConcurrentReads(t *testing.T) {
	ctx := workflowExecutorTestContext()
	store := newMemoryWorkflowRunStore(10)
	for _, record := range []WorkflowRunRecord{terminalComparisonTestRun(ctx, "run_a", WorkflowRunStatusFailed, time.Now().UTC()), terminalComparisonTestRun(ctx, "run_b", WorkflowRunStatusSucceeded, time.Now().UTC().Add(time.Second))} {
		value := record
		storeTerminalComparisonTestRun(t, store, ctx, &value)
	}
	service := newWorkflowExecutorService(nil, nil, store)
	results := make(chan WorkflowRunComparisonResult, 8)
	for index := 0; index < 8; index++ {
		go func() { results <- service.CompareRuns(ctx, "run_a", "run_b") }()
	}
	for index := 0; index < 8; index++ {
		if result := <-results; result.FailureCode != "" || result.Comparison == nil || result.Comparison.Classification != WorkflowRunComparisonImprovement {
			t.Fatalf("concurrent comparison failed: %#v", result)
		}
	}
	if result := service.CompareRuns(ctx, "run_a", "run_a"); result.FailureCode != WorkflowRunFailureComparisonInvalid {
		t.Fatalf("same id accepted: %#v", result)
	}
	other := ctx
	other.TenantRef = "tenant_other"
	if result := service.CompareRuns(other, "run_a", "run_b"); result.FailureCode != WorkflowRunFailureRecordNotFound {
		t.Fatalf("cross scope comparison leaked: %#v", result)
	}
}

func TestWorkflowRunComparisonAndEvaluationRejectToolSideEffectProfileExplicitly(t *testing.T) {
	runStore := newMemoryWorkflowRunStore(20)
	actionStore := newMemoryWorkflowHTTPToolActionStore(&runStore.mu)
	executionStore := newMemoryWorkflowHTTPToolExecutionStore(actionStore, runStore)
	actionCtx, approvedPlan, confirmation := seedApprovedMemoryWorkflowHTTPToolPlan(t, actionStore)
	consumedPlan := cloneWorkflowHTTPToolActionPlan(approvedPlan)
	consumedPlan.Status = WorkflowHTTPToolActionStatusConsumed
	consumedPlan.RecordVersion++
	attempt, toolRun, claimAudit := workflowHTTPToolClaimFixture(actionCtx, consumedPlan, confirmation, 0)
	if err := executionStore.ClaimExecution(actionCtx, &consumedPlan, confirmation, &attempt, &toolRun, claimAudit); err != nil {
		t.Fatalf("claim tool run: %v", err)
	}
	completeWorkflowHTTPToolFixture(actionCtx, &attempt, &toolRun, WorkflowHTTPToolAttemptSucceeded)
	completionAudit := workflowHTTPToolCompletionAuditFixture(actionCtx, attempt, toolRun, "wtae_4444444444444444")
	if err := executionStore.CompleteExecution(actionCtx, &attempt, &toolRun, completionAudit); err != nil {
		t.Fatalf("complete tool run: %v", err)
	}
	runCtx := workflowRunContextFromToolAction(actionCtx)
	baseline := terminalComparisonTestRun(runCtx, "run_zero_side_effect", WorkflowRunStatusSucceeded, time.Now().UTC().Add(-time.Second))
	storeTerminalComparisonTestRun(t, runStore, runCtx, &baseline)

	comparisonService := newWorkflowExecutorService(nil, nil, runStore)
	comparison := comparisonService.CompareRuns(runCtx, baseline.RunID, toolRun.RunID)
	if comparison.FailureCode != WorkflowRunFailureSideEffectUnsupported || comparison.Comparison != nil {
		t.Fatalf("tool run comparison did not return explicit unsupported profile: %#v", comparison)
	}
	evaluationService := newWorkflowEvaluationService(newMemoryWorkflowEvaluationStore(10), runStore)
	evaluation := evaluationService.Create(runCtx, WorkflowEvaluationCreateRequest{
		Name: "tool side effect isolation", BaselineRunID: baseline.RunID,
		Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: toolRun.RunID, ExpectedClassification: WorkflowRunComparisonChanged}},
	})
	if evaluation.FailureCode != WorkflowEvaluationFailureSideEffectProfile || evaluation.Case != nil {
		t.Fatalf("tool run evaluation did not return explicit unsupported profile: %#v", evaluation)
	}
}

func TestWorkflowRunComparisonHTTPRoute(t *testing.T) {
	server, _, draft := newWorkflowExecutorHTTPTestServer(t)
	ctx := workflowExecutorTestContext()
	baseline := terminalComparisonTestRun(ctx, "run_http_baseline", WorkflowRunStatusSucceeded, time.Now().UTC())
	candidate := terminalComparisonTestRun(ctx, "run_http_candidate", WorkflowRunStatusFailed, time.Now().UTC().Add(time.Second))
	candidate.FailureCode = WorkflowRunFailureGatewayFailed
	setWorkflowRunFailureDiagnostic(&candidate, candidate.FailureCode, "node_model", WorkflowRunGatewayFailureTimeout)
	candidate.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
	for _, record := range []*WorkflowRunRecord{&baseline, &candidate} {
		storeTerminalComparisonTestRun(t, server.workflowRunStore, ctx, record)
	}

	url := "/v1/user-workspace/workflow-runs/run_http_candidate/comparison?baseline_run_id=run_http_baseline&workspace_id=" + draft.WorkspaceID + "&application_id=" + draft.ApplicationID
	request := httptest.NewRequest(http.MethodGet, url, nil)
	setSavedWorkflowDraftDevHeaders(request, "workflow_runs:read")
	response := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(response, request)
	var envelope workflowRunComparisonEnvelope
	if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &envelope) != nil || envelope.FailureCode != nil || envelope.Comparison == nil || envelope.Comparison.Classification != WorkflowRunComparisonRegression {
		t.Fatalf("unexpected comparison response: %d %s", response.Code, response.Body.String())
	}
	for _, forbidden := range []string{"output_preview", "input_bytes", "condition_node_ids", "actor_ref", "credential", "endpoint", "provider_raw_envelope"} {
		if strings.Contains(response.Body.String(), forbidden) {
			t.Fatalf("comparison response leaked %s: %s", forbidden, response.Body.String())
		}
	}

	invalid := httptest.NewRequest(http.MethodGet, url+"&threshold=20", nil)
	setSavedWorkflowDraftDevHeaders(invalid, "workflow_runs:read")
	invalidResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(invalidResponse, invalid)
	var invalidEnvelope workflowRunComparisonEnvelope
	_ = json.Unmarshal(invalidResponse.Body.Bytes(), &invalidEnvelope)
	if invalidEnvelope.FailureCode == nil || *invalidEnvelope.FailureCode != string(WorkflowRunFailureComparisonInvalid) {
		t.Fatalf("unknown query accepted: %s", invalidResponse.Body.String())
	}
}

func terminalComparisonTestRun(ctx WorkflowRunContext, runID string, status WorkflowRunStatus, startedAt time.Time) WorkflowRunRecord {
	record := workflowRunHistoryTestRecord(ctx, runID, "draft_compare", startedAt)
	record.Status = status
	record.CompletedAt = workflowRunTimestamp(startedAt.Add(100 * time.Millisecond))
	record.SelectedProvider, record.SelectedProfile, record.SelectedModel = "mock", "default", "mock"
	record.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
	return record
}

func storeTerminalComparisonTestRun(t *testing.T, store workflowRunStore, ctx WorkflowRunContext, record *WorkflowRunRecord) {
	t.Helper()
	terminalStatus, completedAt := record.Status, record.CompletedAt
	record.Status, record.CompletedAt = WorkflowRunStatusRunning, ""
	record.Diagnostic.TerminalWriteState = WorkflowRunTerminalWritePending
	if err := store.UpsertRun(ctx, record); err != nil {
		t.Fatal(err)
	}
	record.Status, record.CompletedAt = terminalStatus, completedAt
	record.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
	if err := store.UpsertRun(ctx, record); err != nil {
		t.Fatal(err)
	}
}
