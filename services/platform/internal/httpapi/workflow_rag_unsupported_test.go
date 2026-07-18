package httpapi

import (
	"testing"
	"time"
)

func TestWorkflowRAGRunV3IsExplicitlyUnsupportedByComparisonEvaluationBaselineAndSuite(t *testing.T) {
	fixture := newWorkflowRAGExecutionFixture(t)
	executed := fixture.service.Execute(fixture.runContext, workflowRAGExecutionFixtureRequest(
		fixture, "official retrieval guidance", fixture.draft.DraftVersion,
	))
	if executed.FailureCode != "" || executed.Record == nil {
		t.Fatalf("create v3 unsupported fixture: %#v", executed)
	}
	v1 := terminalComparisonTestRun(fixture.runContext, "run_nonrag_baseline", WorkflowRunStatusSucceeded, time.Now().UTC().Add(-time.Minute))
	storeTerminalComparisonTestRun(t, fixture.runStore, fixture.runContext, &v1)

	comparison := newWorkflowExecutorService(nil, nil, fixture.runStore).CompareRuns(fixture.runContext, v1.RunID, executed.Record.RunID)
	if comparison.FailureCode != WorkflowRunFailureRetrievalUnsupported || comparison.Comparison != nil {
		t.Fatalf("run comparison did not reject v3 explicitly: %#v", comparison)
	}

	evaluationStore := newMemoryWorkflowEvaluationStore(10)
	evaluation := newWorkflowEvaluationService(evaluationStore, fixture.runStore)
	evaluation.newCaseID = func() (string, error) { return "eval_rag_unsupported", nil }
	created := evaluation.Create(fixture.runContext, WorkflowEvaluationCreateRequest{
		Name: "retrieval baseline is unsupported", BaselineRunID: executed.Record.RunID,
		Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: v1.RunID, ExpectedClassification: WorkflowRunComparisonUnchanged}},
	})
	if created.FailureCode != WorkflowEvaluationFailureRetrievalProfile || created.Case != nil {
		t.Fatalf("evaluation case accepted v3 baseline: %#v", created)
	}

	evaluation.newCaseID = func() (string, error) { return "eval_nonrag_case", nil }
	seedEvaluationRun(t, fixture.runStore, fixture.runContext, "run_nonrag_candidate", WorkflowRunStatusSucceeded, time.Now().UTC().Add(-30*time.Second))
	nonRAG := evaluation.Create(fixture.runContext, WorkflowEvaluationCreateRequest{
		Name: "non retrieval baseline", BaselineRunID: v1.RunID,
		Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "run_nonrag_candidate", ExpectedClassification: WorkflowRunComparisonUnchanged}},
	})
	if nonRAG.Case == nil || nonRAG.FailureCode != "" {
		t.Fatalf("create non-RAG evaluation fixture: %#v", nonRAG)
	}
	promotion := evaluation.Revise(fixture.runContext, nonRAG.Case.CaseID, WorkflowEvaluationRevisionRequest{
		ExpectedVersion: 1, RevisionKind: WorkflowEvaluationRevisionBaselinePromotion,
		Name: "retrieval baseline promotion", BaselineRunID: executed.Record.RunID,
		Expectations: nonRAG.Case.Expectations,
	})
	if promotion.FailureCode != WorkflowEvaluationFailureRetrievalProfile || promotion.Case != nil {
		t.Fatalf("baseline promotion accepted v3: %#v", promotion)
	}

	historical := WorkflowEvaluationCase{
		SchemaVersion: workflowEvaluationCaseSchema, CaseID: "eval_historical_rag", Version: 1,
		RevisionKind: WorkflowEvaluationRevisionCreated, ChangeCodes: []string{"created"}, Name: "historical retrieval case",
		WorkspaceID: fixture.runContext.WorkspaceID, ApplicationID: fixture.runContext.ApplicationID,
		BaselineRunID: executed.Record.RunID,
		Expectations:  []WorkflowEvaluationExpectation{{CandidateRunID: v1.RunID, ExpectedClassification: WorkflowRunComparisonUnchanged}},
		CreatedAt:     workflowRunTimestamp(time.Now().UTC()), RevisedAt: workflowRunTimestamp(time.Now().UTC()),
		ActorRef: fixture.runContext.ActorRef, RequestID: fixture.runContext.RequestID, AuditRef: fixture.runContext.AuditRef,
	}
	if err := evaluationStore.CreateCase(fixture.runContext, historical); err != nil {
		t.Fatalf("seed historical v3 evaluation case: %v", err)
	}
	suite := newWorkflowEvaluationSuiteService(newMemoryWorkflowEvaluationSuiteStore(10), evaluation)
	suite.newSuiteID = func() (string, error) { return "suite_rag_unsupported", nil }
	suiteResult := suite.Create(fixture.runContext, WorkflowEvaluationSuiteCreateRequest{
		Name: "retrieval suite is unsupported", CaseRefs: []WorkflowEvaluationSuiteCaseRef{{CaseID: historical.CaseID, Version: 1}},
	})
	if suiteResult.FailureCode != WorkflowEvaluationSuiteFailureRetrievalProfile || suiteResult.Suite != nil {
		t.Fatalf("evaluation suite accepted a v3 case: %#v", suiteResult)
	}
}
