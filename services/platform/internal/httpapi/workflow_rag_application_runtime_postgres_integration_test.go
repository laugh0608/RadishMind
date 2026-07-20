//go:build postgres_integration

package httpapi

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	workflowrunmigrations "radishmind.local/services/platform/migrations/workflow_runs"
)

func TestPostgresWorkflowRAGApplicationRuntimeRestartAndV4RunIntegration(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeDatabaseURL := postgresIntegrationDatabaseURLForCredentials(t, runtimeUser, os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	adminPool, err := workflowrunmigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, adminPool)
	resetPostgresWorkflowRunSchema(t, ctx, adminPool)
	preparePostgresIntegrationRuntimeRole(t, ctx, adminPool, runtimeUser)
	t.Cleanup(func() {
		cleanup, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		resetPostgresWorkflowRunSchema(t, cleanup, adminPool)
		adminPool.Close()
	})
	state, err := workflowrunmigrations.Apply(ctx, adminPool)
	if err != nil || state.MigrationID != workflowrunmigrations.MigrationID || state.StoreSchemaVersion != workflowrunmigrations.StoreSchemaVersion {
		t.Fatalf("apply PostgreSQL application RAG runtime migration: state=%#v err=%v", state, err)
	}
	runtimePool, err := workflowrunmigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	runStore := newPostgresWorkflowRunStore(runtimePool)
	repository, err := newWorkflowRAGApplicationRuntimeRepositoryForRunStore(runStore)
	if err != nil {
		t.Fatalf("create shared PostgreSQL application runtime repository: %v", err)
	}
	postgresRepository, ok := repository.(*postgresWorkflowRAGApplicationRuntimeRepository)
	if !ok || postgresRepository.pool != runtimePool {
		t.Fatalf("application runtime repository did not reuse workflow PostgreSQL pool: %T", repository)
	}
	fixture := newWorkflowRAGApplicationRuntimeTestFixture(t)
	fixture.runtimeContext.RequestContext = ctx
	service := newWorkflowRAGApplicationRuntimeService(repository, fixture.resolver)
	service.newID = workflowRAGEvaluationTestIDGenerator()
	service.now = fixture.runtimeService.now
	activated := service.Decide(fixture.runtimeContext, WorkflowRAGApplicationRuntimeDecisionInput{
		ExpectedRecordVersion: 0, Decision: workflowRAGApplicationRuntimeDecisionActivate,
		PublishCandidateID: fixture.publishCandidate.CandidateID, Reason: "PostgreSQL 持久化激活受控调用候选",
	})
	if activated.FailureCode != "" || activated.Assignment == nil {
		t.Fatalf("activate PostgreSQL application runtime assignment: %#v", activated)
	}
	invocation := newWorkflowRAGApplicationInvocationService(repository, fixture.resolver, runStore, fixture.bridge)
	invocation.resolveSelection = func(context.Context, string) northboundSelection { return workflowRAGTestSelection() }
	invocation.newRunID = func() (string, error) { return "run_applicationragpostgres01", nil }
	invocation.now = fixture.runtimeService.now
	result := invocation.Invoke(fixture.runtimeContext, WorkflowRAGApplicationInvocationInput{Input: "PostgreSQL restart authority evidence"})
	if result.FailureCode != "" || result.Run == nil {
		t.Fatalf("invoke PostgreSQL application RAG runtime: %#v", result)
	}
	invocation.newRunID = func() (string, error) { return "run_applicationragpostgres02", nil }
	candidateResult := invocation.Invoke(fixture.runtimeContext, WorkflowRAGApplicationInvocationInput{Input: "PostgreSQL restart authority evidence"})
	if candidateResult.FailureCode != "" || candidateResult.Run == nil {
		t.Fatalf("invoke comparable PostgreSQL application RAG runtime: %#v", candidateResult)
	}
	historyService := workflowExecutorService{store: runStore}
	comparison := historyService.CompareRuns(workflowRAGApplicationRunContext(fixture.runtimeContext), result.Run.RunID, candidateResult.Run.RunID)
	if comparison.FailureCode != "" || comparison.Comparison == nil || comparison.Comparison.SchemaVersion != workflowRAGAppRunComparisonSchemaVersion || comparison.Comparison.Retrieval == nil || comparison.Comparison.Retrieval.RunProfile != workflowRAGApplicationComparisonProfile {
		t.Fatalf("compare PostgreSQL application RAG v4 runs: %#v", comparison)
	}
	evaluationService := newWorkflowEvaluationService(newWorkflowEvaluationStoreForRunStore(runStore), runStore)
	evaluation := evaluationService.Create(workflowRAGApplicationRunContext(fixture.runtimeContext), WorkflowEvaluationCreateRequest{
		Name: "application RAG PostgreSQL metadata comparison", BaselineRunID: result.Run.RunID,
		Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: candidateResult.Run.RunID, ExpectedClassification: WorkflowRunComparisonUnchanged}},
	})
	if evaluation.FailureCode != "" || evaluation.Case == nil {
		t.Fatalf("create PostgreSQL application RAG v4 evaluation: %#v", evaluation)
	}
	review := evaluationService.Review(workflowRAGApplicationRunContext(fixture.runtimeContext), evaluation.Case.CaseID)
	if review.FailureCode != "" || review.Review == nil || review.Review.Outcome != "passed" || review.Review.Items[0].ComparisonSchemaVersion != workflowRAGAppRunComparisonSchemaVersion {
		t.Fatalf("review PostgreSQL application RAG v4 evaluation: %#v", review)
	}
	fixture.advanceRuntimeRequest("postgres_revoke")
	revoked := service.Decide(fixture.runtimeContext, WorkflowRAGApplicationRuntimeDecisionInput{ExpectedRecordVersion: 1, Decision: workflowRAGApplicationRuntimeDecisionRevoke, Reason: "连续链人工撤销后必须关闭调用"})
	if revoked.FailureCode != "" || revoked.Assignment == nil || revoked.Assignment.RecordVersion != 2 {
		t.Fatalf("revoke PostgreSQL application RAG runtime: %#v", revoked)
	}
	bridgeCalls := fixture.bridge.callCount()
	blocked := invocation.Invoke(fixture.runtimeContext, WorkflowRAGApplicationInvocationInput{Input: "PostgreSQL invocation after revoke"})
	if blocked.FailureCode != WorkflowRAGApplicationFailureAssignmentRevoked || blocked.Run != nil || fixture.bridge.callCount() != bridgeCalls {
		t.Fatalf("revoked PostgreSQL assignment reached Gateway: result=%#v calls=%d/%d", blocked, fixture.bridge.callCount(), bridgeCalls)
	}
	if _, err = runtimePool.Exec(ctx, `UPDATE workflow_rag_application_runtime_events SET after_record_version=after_record_version WHERE event_id=$1`, activated.Events[0].EventID); err == nil {
		t.Fatal("PostgreSQL application runtime event accepted an update")
	}
	runtimePool.Close()
	if _, _, _, err = repository.Read(fixture.runtimeContext); err == nil {
		t.Fatal("closed PostgreSQL application runtime silently fell back")
	}

	reopened, err := workflowrunmigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	restartedRepository := newPostgresWorkflowRAGApplicationRuntimeRepository(reopened)
	recoveredAssignment, recoveredEvents, recoveredAudits, err := restartedRepository.Read(fixture.runtimeContext)
	if err != nil || recoveredAssignment.AssignmentID != activated.Assignment.AssignmentID || recoveredAssignment.RecordVersion != 2 || recoveredAssignment.State != workflowRAGApplicationRuntimeStateRevoked || len(recoveredEvents) != 2 || len(recoveredAudits) != 2 {
		t.Fatalf("recover PostgreSQL application runtime assignment: assignment=%#v events=%d audits=%d err=%v", recoveredAssignment, len(recoveredEvents), len(recoveredAudits), err)
	}
	recoveredRun, found, err := newPostgresWorkflowRunStore(reopened).ReadRun(workflowRAGApplicationRunContext(fixture.runtimeContext), result.Run.RunID)
	if err != nil || !found || recoveredRun.ExecutionSource == nil || recoveredRun.ExecutionSource.SourceKind != workflowRAGApplicationExecutionSourceKind || recoveredRun.RAGApplication == nil {
		t.Fatalf("recover PostgreSQL application RAG v4 run: found=%t run=%#v err=%v", found, recoveredRun, err)
	}
}
