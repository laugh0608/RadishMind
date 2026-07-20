package httpapi

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestSQLiteWorkflowEvaluationStoresPersistAcrossRestartWithoutFallback(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "workflow-evaluation.db")
	runtime := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	runContext := workflowExecutorTestContext()
	caseStore := newSQLiteWorkflowEvaluationStore(runtime.DB())
	suiteStore := newSQLiteWorkflowEvaluationSuiteStore(runtime.DB())

	created := WorkflowEvaluationCase{
		SchemaVersion:   workflowEvaluationCaseSchema,
		CaseID:          "eval_sqlite_restart_001",
		Version:         1,
		PreviousVersion: 0,
		RevisionKind:    WorkflowEvaluationRevisionCreated,
		ChangeCodes:     []string{"created"},
		Name:            "SQLite RAG regression case",
		WorkspaceID:     runContext.WorkspaceID,
		ApplicationID:   runContext.ApplicationID,
		BaselineRunID:   "run_rag_baseline_001",
		Expectations: []WorkflowEvaluationExpectation{{
			CandidateRunID:         "run_rag_candidate_001",
			ExpectedClassification: WorkflowRunComparisonUnchanged,
		}},
		CreatedAt: "2026-07-18T05:00:00Z",
		RevisedAt: "2026-07-18T05:00:00Z",
		ActorRef:  "actor_sqlite",
		RequestID: "request_sqlite_case_001",
		AuditRef:  "audit_sqlite_case_001",
	}
	if err := caseStore.CreateCase(runContext, created); err != nil {
		t.Fatalf("create SQLite evaluation case: %v", err)
	}
	if err := caseStore.CreateCase(runContext, created); !errors.Is(err, errWorkflowEvaluationStoreContract) {
		t.Fatalf("duplicate SQLite evaluation case did not fail closed: %v", err)
	}
	revised := cloneWorkflowEvaluationCase(created)
	revised.Version = 2
	revised.PreviousVersion = 1
	revised.RevisionKind = WorkflowEvaluationRevisionCase
	revised.ChangeCodes = []string{"name_changed"}
	revised.Name = "SQLite RAG regression case revised"
	revised.RevisedAt = "2026-07-18T05:01:00Z"
	revised.RequestID = "request_sqlite_case_002"
	revised.AuditRef = "audit_sqlite_case_002"
	if current, changed, err := caseStore.ReviseCase(runContext, 1, revised); err != nil || !changed || current.Version != 2 {
		t.Fatalf("revise SQLite evaluation case: current=%#v changed=%v err=%v", current, changed, err)
	}
	if current, changed, err := caseStore.ReviseCase(runContext, 1, revised); err != nil || changed || current.Version != 2 {
		t.Fatalf("stale SQLite evaluation case revision was not rejected: current=%#v changed=%v err=%v", current, changed, err)
	}

	suite := WorkflowEvaluationSuite{
		SchemaVersion: workflowEvaluationSuiteSchema,
		SuiteID:       "suite_sqlite_restart_001",
		Name:          "SQLite RAG regression suite",
		WorkspaceID:   runContext.WorkspaceID,
		ApplicationID: runContext.ApplicationID,
		CaseRefs: []WorkflowEvaluationSuiteCaseRef{{
			CaseID:  revised.CaseID,
			Version: revised.Version,
		}},
		CreatedAt: "2026-07-18T05:02:00Z",
		ActorRef:  "actor_sqlite",
		RequestID: "request_sqlite_suite_001",
		AuditRef:  "audit_sqlite_suite_001",
	}
	if err := suiteStore.CreateSuite(runContext, suite); err != nil {
		t.Fatalf("create SQLite evaluation suite: %v", err)
	}
	decision := WorkflowEvaluationReleaseDecision{
		SchemaVersion: workflowEvaluationDecisionSchema,
		DecisionID:    "decision_sqlite_001",
		SuiteID:       suite.SuiteID,
		Version:       1,
		Decision:      "approved",
		ReviewDigest:  "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		ReviewOutcome: "passed",
		Passed:        1,
		CreatedAt:     "2026-07-18T05:03:00Z",
		ActorRef:      "actor_sqlite",
		RequestID:     "request_sqlite_decision_001",
		AuditRef:      "audit_sqlite_decision_001",
	}
	if current, appended, err := suiteStore.AppendDecision(runContext, 0, decision); err != nil || !appended || current.CurrentDecisionVersion != 1 {
		t.Fatalf("append SQLite evaluation decision: current=%#v appended=%v err=%v", current, appended, err)
	}

	if err := runtime.Close(); err != nil {
		t.Fatalf("close first SQLite runtime: %v", err)
	}
	runtime = openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	caseStore = newSQLiteWorkflowEvaluationStore(runtime.DB())
	suiteStore = newSQLiteWorkflowEvaluationSuiteStore(runtime.DB())

	storedCase, found, err := caseStore.ReadCase(runContext, created.CaseID)
	if err != nil || !found || storedCase.Version != 2 || storedCase.Name != revised.Name {
		t.Fatalf("read restarted SQLite evaluation case: case=%#v found=%v err=%v", storedCase, found, err)
	}
	revisions, err := caseStore.ListRevisions(runContext, created.CaseID, WorkflowEvaluationRevisionListFilter{Limit: 10})
	if err != nil || len(revisions.Revisions) != 2 || revisions.Revisions[0].Version != 2 || revisions.Revisions[1].Version != 1 {
		t.Fatalf("list restarted SQLite evaluation revisions: page=%#v err=%v", revisions, err)
	}
	storedSuite, found, err := suiteStore.ReadSuite(runContext, suite.SuiteID)
	if err != nil || !found || storedSuite.CurrentDecisionVersion != 1 || storedSuite.CurrentDecision != "approved" {
		t.Fatalf("read restarted SQLite evaluation suite: suite=%#v found=%v err=%v", storedSuite, found, err)
	}
	decisions, err := suiteStore.ListDecisions(runContext, suite.SuiteID, workflowEvaluationDecisionListFilter{Limit: 10})
	if err != nil || len(decisions.Decisions) != 1 || decisions.Decisions[0].DecisionID != decision.DecisionID {
		t.Fatalf("list restarted SQLite evaluation decisions: page=%#v err=%v", decisions, err)
	}
	if _, err := runtime.DB().ExecContext(context.Background(), `UPDATE workflow_evaluation_case_revisions SET sanitized_revision_record=sanitized_revision_record WHERE case_id=?`, created.CaseID); err == nil {
		t.Fatal("SQLite evaluation revisions are not append-only")
	}
	if _, err := runtime.DB().ExecContext(context.Background(), `UPDATE workflow_evaluation_suite_decisions SET sanitized_decision_record=sanitized_decision_record WHERE suite_id=?`, suite.SuiteID); err == nil {
		t.Fatal("SQLite evaluation decisions are not append-only")
	}
	if _, err := runtime.DB().ExecContext(context.Background(), `UPDATE workflow_evaluation_cases SET baseline_run_id=? WHERE case_id=?`, "run_corrupted", created.CaseID); err != nil {
		t.Fatalf("inject SQLite evaluation projection corruption: %v", err)
	}
	if _, _, err := caseStore.ReadCase(runContext, created.CaseID); !errors.Is(err, errWorkflowEvaluationStoreContract) {
		t.Fatalf("SQLite evaluation projection corruption was accepted: %v", err)
	}

	if err := runtime.Close(); err != nil {
		t.Fatalf("close restarted SQLite runtime: %v", err)
	}
	if _, _, err := caseStore.ReadCase(runContext, created.CaseID); err == nil {
		t.Fatal("closed SQLite case store silently fell back")
	}
	if _, _, err := suiteStore.ReadSuite(runContext, suite.SuiteID); err == nil {
		t.Fatal("closed SQLite suite store silently fell back")
	}
}
