package httpapi

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestSQLiteWorkflowRAGApplicationRuntimeRestartTransactionAndV4RunSource(t *testing.T) {
	fixture := newWorkflowRAGApplicationRuntimeTestFixture(t)
	databasePath := filepath.Join(t.TempDir(), "workflow-rag-application-runtime.db")
	firstRuntime := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	runStore := newSQLiteWorkflowRunStore(firstRuntime.DB())
	repository, err := newWorkflowRAGApplicationRuntimeRepositoryForRunStore(runStore)
	if err != nil {
		t.Fatalf("create shared SQLite application runtime repository: %v", err)
	}
	sqliteRepository, ok := repository.(*sqliteWorkflowRAGApplicationRuntimeRepository)
	if !ok || sqliteRepository.database != firstRuntime.DB() {
		t.Fatalf("application runtime repository did not reuse workflow SQLite database: %T", repository)
	}
	service := newWorkflowRAGApplicationRuntimeService(repository, fixture.resolver)
	service.newID = workflowRAGEvaluationTestIDGenerator()
	service.now = fixture.runtimeService.now
	activated := service.Decide(fixture.runtimeContext, WorkflowRAGApplicationRuntimeDecisionInput{
		ExpectedRecordVersion: 0, Decision: workflowRAGApplicationRuntimeDecisionActivate,
		PublishCandidateID: fixture.publishCandidate.CandidateID, Reason: "SQLite 持久化激活受控调用候选",
	})
	if activated.FailureCode != "" || activated.Assignment == nil || activated.Assignment.RecordVersion != 1 {
		t.Fatalf("activate SQLite application runtime assignment: %#v", activated)
	}

	invocation := newWorkflowRAGApplicationInvocationService(repository, fixture.resolver, runStore, fixture.bridge)
	invocation.resolveSelection = func(context.Context, string) northboundSelection { return workflowRAGTestSelection() }
	invocation.newRunID = func() (string, error) { return "run_applicationragsqlite01", nil }
	invocation.now = fixture.runtimeService.now
	result := invocation.Invoke(fixture.runtimeContext, WorkflowRAGApplicationInvocationInput{Input: "SQLite restart authority evidence"})
	if result.FailureCode != "" || result.Run == nil || result.Run.SchemaVersion != workflowRunRecordAppRAGSchemaVersion {
		t.Fatalf("invoke SQLite application RAG runtime: %#v", result)
	}
	if filtered, listErr := runStore.ListRuns(workflowRAGApplicationRunContext(fixture.runtimeContext), WorkflowRunListFilter{DraftID: fixture.publishCandidate.DraftID}); listErr != nil || len(filtered.Records) != 0 {
		t.Fatalf("legacy draft filter returned application configuration source: %#v err=%v", filtered, listErr)
	}
	invocation.newRunID = func() (string, error) { return "run_applicationragsqlite02", nil }
	candidateResult := invocation.Invoke(fixture.runtimeContext, WorkflowRAGApplicationInvocationInput{Input: "SQLite restart authority evidence"})
	if candidateResult.FailureCode != "" || candidateResult.Run == nil {
		t.Fatalf("invoke comparable SQLite application RAG runtime: %#v", candidateResult)
	}
	historyService := workflowExecutorService{store: runStore}
	comparison := historyService.CompareRuns(workflowRAGApplicationRunContext(fixture.runtimeContext), result.Run.RunID, candidateResult.Run.RunID)
	if comparison.FailureCode != "" || comparison.Comparison == nil || comparison.Comparison.Retrieval == nil || comparison.Comparison.Retrieval.RunProfile != workflowRAGApplicationComparisonProfile || comparison.Comparison.ExecutionSourceChanged {
		t.Fatalf("compare SQLite application RAG v4 runs: %#v", comparison)
	}
	history := historyService.ListRuns(workflowRAGApplicationRunContext(fixture.runtimeContext), WorkflowRunListRequest{Limit: 10})
	if history.FailureCode != "" || len(history.Runs) != 2 || history.Runs[0].ExecutionSourceKind != workflowRAGApplicationExecutionSourceKind || history.Runs[0].RuntimeAssignmentID != activated.Assignment.AssignmentID || history.Runs[0].PublishCandidateID != fixture.publishCandidate.CandidateID {
		t.Fatalf("list SQLite application RAG v4 history: %#v", history)
	}
	evaluationService := newWorkflowEvaluationService(newWorkflowEvaluationStoreForRunStore(runStore), runStore)
	evaluation := evaluationService.Create(workflowRAGApplicationRunContext(fixture.runtimeContext), WorkflowEvaluationCreateRequest{
		Name: "application RAG SQLite metadata comparison", BaselineRunID: result.Run.RunID,
		Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: candidateResult.Run.RunID, ExpectedClassification: WorkflowRunComparisonUnchanged}},
	})
	if evaluation.FailureCode != "" || evaluation.Case == nil {
		t.Fatalf("create SQLite application RAG v4 evaluation: %#v", evaluation)
	}
	review := evaluationService.Review(workflowRAGApplicationRunContext(fixture.runtimeContext), evaluation.Case.CaseID)
	if review.FailureCode != "" || review.Review == nil || review.Review.Outcome != "passed" {
		t.Fatalf("review SQLite application RAG v4 evaluation: %#v", review)
	}

	assignment, authority, failure := invocation.loadAuthority(fixture.runtimeContext)
	if failure != "" {
		t.Fatalf("load authority for stale SQLite v4 run: %s", failure)
	}
	stale := newWorkflowRAGApplicationRunRecord(fixture.runtimeContext, "stale metadata only", assignment, authority, workflowRAGTestSelection(), "run_applicationragstale001", fixture.now.Add(-time.Hour))
	if err = runStore.UpsertRun(workflowRAGApplicationRunContext(fixture.runtimeContext), &stale); err != nil {
		t.Fatalf("create stale SQLite application RAG v4 run: %v", err)
	}
	bridgeCalls := fixture.bridge.callCount()
	reconciled := invocation.ReconcileStale(workflowRAGApplicationRunContext(fixture.runtimeContext))
	if reconciled.FailureCode != "" || reconciled.Reconciled != 1 || fixture.bridge.callCount() != bridgeCalls {
		t.Fatalf("reconcile stale SQLite application RAG v4 run without replay: %#v bridge=%d/%d", reconciled, bridgeCalls, fixture.bridge.callCount())
	}
	recoveredStale, found, err := runStore.ReadRun(workflowRAGApplicationRunContext(fixture.runtimeContext), stale.RunID)
	if err != nil || !found || recoveredStale.Status != WorkflowRunStatusFailed || recoveredStale.FailureCode != WorkflowRunFailureCode(WorkflowRAGFailureInterrupted) || recoveredStale.SideEffects.ProviderCalls != 0 {
		t.Fatalf("read reconciled SQLite application RAG v4 run: found=%t run=%#v err=%v", found, recoveredStale, err)
	}
	var sourceKind, sourceID string
	var sourceVersion int
	if err = firstRuntime.DB().QueryRowContext(context.Background(), `SELECT execution_source_kind,execution_source_id,execution_source_version
FROM workflow_run_records WHERE run_id=?`, result.Run.RunID).Scan(&sourceKind, &sourceID, &sourceVersion); err != nil || sourceKind != workflowRAGApplicationExecutionSourceKind || sourceID != fixture.publishCandidate.DraftID || sourceVersion != fixture.publishCandidate.DraftVersion {
		t.Fatalf("stored v4 execution source drifted: kind=%q id=%q version=%d err=%v", sourceKind, sourceID, sourceVersion, err)
	}

	if _, err = firstRuntime.DB().ExecContext(context.Background(), `CREATE TRIGGER workflow_rag_application_runtime_test_reject_audit
BEFORE INSERT ON workflow_rag_application_runtime_audits BEGIN
	SELECT RAISE(ABORT, 'injected application runtime audit failure');
END`); err != nil {
		t.Fatalf("install SQLite application runtime failure trigger: %v", err)
	}
	fixture.advanceRuntimeRequest("sqlite_rollback")
	failed := service.Decide(fixture.runtimeContext, WorkflowRAGApplicationRuntimeDecisionInput{ExpectedRecordVersion: 1, Decision: workflowRAGApplicationRuntimeDecisionRevoke, Reason: "事务失败必须回滚全部运行时记录"})
	if failed.FailureCode != WorkflowRAGApplicationFailureStoreUnavailable {
		t.Fatalf("injected SQLite transaction failure did not fail closed: %#v", failed)
	}
	if _, err = firstRuntime.DB().ExecContext(context.Background(), `DROP TRIGGER workflow_rag_application_runtime_test_reject_audit`); err != nil {
		t.Fatalf("remove SQLite application runtime failure trigger: %v", err)
	}
	stable := service.Read(fixture.runtimeContext)
	if stable.FailureCode != "" || stable.Assignment == nil || stable.Assignment.RecordVersion != 1 || len(stable.Events) != 1 || len(stable.Audits) != 1 {
		t.Fatalf("SQLite application runtime transaction partially committed: %#v", stable)
	}
	if _, err = firstRuntime.DB().ExecContext(context.Background(), `UPDATE workflow_rag_application_runtime_events SET after_record_version=after_record_version WHERE event_id=?`, stable.Events[0].EventID); err == nil {
		t.Fatal("SQLite application runtime event accepted an update")
	}
	if _, err = firstRuntime.DB().ExecContext(context.Background(), `DELETE FROM workflow_rag_application_runtime_audits WHERE audit_event_id=?`, stable.Audits[0].AuditEventID); err == nil {
		t.Fatal("SQLite application runtime audit accepted a delete")
	}
	fixture.advanceRuntimeRequest("sqlite_revoke")
	revoked := service.Decide(fixture.runtimeContext, WorkflowRAGApplicationRuntimeDecisionInput{ExpectedRecordVersion: 1, Decision: workflowRAGApplicationRuntimeDecisionRevoke, Reason: "连续链人工撤销后必须关闭调用"})
	if revoked.FailureCode != "" || revoked.Assignment == nil || revoked.Assignment.RecordVersion != 2 || revoked.Assignment.State != workflowRAGApplicationRuntimeStateRevoked {
		t.Fatalf("revoke SQLite application runtime assignment: %#v", revoked)
	}
	bridgeCalls = fixture.bridge.callCount()
	blocked := invocation.Invoke(fixture.runtimeContext, WorkflowRAGApplicationInvocationInput{Input: "SQLite invocation after revoke"})
	if blocked.FailureCode != WorkflowRAGApplicationFailureAssignmentRevoked || blocked.Run != nil || fixture.bridge.callCount() != bridgeCalls {
		t.Fatalf("revoked SQLite assignment reached Gateway: result=%#v calls=%d/%d", blocked, fixture.bridge.callCount(), bridgeCalls)
	}

	if err = firstRuntime.Close(); err != nil {
		t.Fatalf("close first SQLite application runtime: %v", err)
	}
	if _, _, _, err = repository.Read(fixture.runtimeContext); !errors.Is(err, errWorkflowRAGApplicationStore) {
		t.Fatalf("closed SQLite application runtime silently fell back: %v", err)
	}
	secondRuntime := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	t.Cleanup(func() { _ = secondRuntime.Close() })
	restartedRepository := newSQLiteWorkflowRAGApplicationRuntimeRepository(secondRuntime.DB())
	recoveredAssignment, recoveredEvents, recoveredAudits, err := restartedRepository.Read(fixture.runtimeContext)
	if err != nil || recoveredAssignment.AssignmentID != stable.Assignment.AssignmentID || recoveredAssignment.RecordVersion != 2 || recoveredAssignment.State != workflowRAGApplicationRuntimeStateRevoked || len(recoveredEvents) != 2 || len(recoveredAudits) != 2 {
		t.Fatalf("recover SQLite application runtime assignment: assignment=%#v events=%d audits=%d err=%v", recoveredAssignment, len(recoveredEvents), len(recoveredAudits), err)
	}
	restartedRunStore := newSQLiteWorkflowRunStore(secondRuntime.DB())
	recoveredRun, found, err := restartedRunStore.ReadRun(workflowRAGApplicationRunContext(fixture.runtimeContext), result.Run.RunID)
	if err != nil || !found || recoveredRun.ExecutionSource == nil || recoveredRun.ExecutionSource.SourceKind != workflowRAGApplicationExecutionSourceKind || recoveredRun.RAGApplication == nil {
		t.Fatalf("recover SQLite workflow run v4: found=%t run=%#v err=%v", found, recoveredRun, err)
	}
	if _, err = secondRuntime.DB().ExecContext(context.Background(), `UPDATE workflow_rag_application_runtime_assignments
SET sanitized_assignment_payload=json_set(sanitized_assignment_payload, '$.unexpected_field', 1)
WHERE assignment_id=?`, recoveredAssignment.AssignmentID); err != nil {
		t.Fatalf("inject SQLite application runtime corruption: %v", err)
	}
	if _, _, _, err = restartedRepository.Read(fixture.runtimeContext); !errors.Is(err, errWorkflowRAGApplicationStoreContract) {
		t.Fatalf("corrupted SQLite application runtime assignment was accepted: %v", err)
	}
}
