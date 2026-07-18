//go:build postgres_integration

package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	workflowrunmigrations "radishmind.local/services/platform/migrations/workflow_runs"
)

func TestPostgresWorkflowRunStoreIntegration(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeDatabaseURL := postgresIntegrationDatabaseURLForCredentials(t, runtimeUser, os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := workflowrunmigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, pool)
	resetPostgresWorkflowRunSchema(t, ctx, pool)
	preparePostgresIntegrationRuntimeRole(t, ctx, pool, runtimeUser)
	t.Cleanup(func() {
		cleanup, cancelCleanup := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelCleanup()
		resetPostgresWorkflowRunSchema(t, cleanup, pool)
		pool.Close()
	})
	state, err := workflowrunmigrations.Apply(ctx, pool)
	if err != nil || state.MigrationState != workflowrunmigrations.MigrationStateApplied {
		t.Fatalf("apply migration: %#v %v", state, err)
	}
	if repeated, err := workflowrunmigrations.Apply(ctx, pool); err != nil || repeated.MigrationChecksum != state.MigrationChecksum {
		t.Fatalf("repeat migration: %#v %v", repeated, err)
	}
	runtimePool, err := workflowrunmigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	if _, ddlErr := runtimePool.Exec(ctx, "CREATE TABLE workflow_run_runtime_must_not_create (id integer)"); ddlErr == nil {
		t.Fatal("workflow run runtime role must not have schema DDL permission")
	}
	store := newPostgresWorkflowRunStore(runtimePool)
	runContext := workflowExecutorTestContext()
	runContext.RequestContext = ctx
	actionContext := workflowHTTPToolActionTestContext()
	actionContext.RequestContext = ctx
	actionStore := newPostgresWorkflowHTTPToolActionStore(runtimePool)
	actionPlan := workflowHTTPToolActionPlanForStoreTest(t, actionContext, "wtap_0000000000000200")
	if err = actionStore.CreatePlan(actionContext, &actionPlan, workflowHTTPToolAuditForStoreTest(actionPlan, "wtae_0000000000000200", "confirmation_requested")); err != nil {
		t.Fatalf("create PostgreSQL workflow HTTP tool action plan: %v", err)
	}
	actionPlan.RecordVersion = 2
	actionPlan.Status = WorkflowHTTPToolActionStatusApproved
	actionActor, actionDecidedAt := actionContext.ActorRef, "2026-07-16T09:01:00Z"
	actionPlan.LastDecisionByActorRef, actionPlan.LastDecisionAt = &actionActor, &actionDecidedAt
	actionDecision := workflowHTTPToolDecisionForStoreTest(actionPlan, "wtcd_0000000000000200", WorkflowHTTPToolConfirmationApprove, actionActor)
	actionDecision.AuditRef = "audit_workflow_http_tool_postgres_decision"
	actionContext.AuditRef = actionDecision.AuditRef
	actionAudit := workflowHTTPToolAuditForStoreTest(actionPlan, "wtae_0000000000000201", "confirmation_recorded", actionDecision.ConfirmationID)
	actionAudit.AuditRef = actionDecision.AuditRef
	if err = actionStore.DecidePlan(actionContext, &actionPlan, actionDecision, actionAudit); err != nil {
		t.Fatalf("approve PostgreSQL workflow HTTP tool action plan: %v", err)
	}
	actionPlan.RecordVersion = 3
	actionPlan.Status = WorkflowHTTPToolActionStatusInvalidated
	systemActor, invalidatedAt := "system:workflow_tool_action_policy", "2026-07-16T09:02:00Z"
	actionPlan.LastDecisionByActorRef, actionPlan.LastDecisionAt = &systemActor, &invalidatedAt
	invalidateDecision := workflowHTTPToolSystemDecisionForStoreTest(actionPlan, "wtcd_0000000000000201", workflowHTTPToolConfirmationInvalidate)
	invalidateDecision.DecidedAt = invalidatedAt
	invalidateDecision.AuditRef = "audit_workflow_http_tool_postgres_invalidate"
	actionContext.AuditRef = invalidateDecision.AuditRef
	invalidateAudit := workflowHTTPToolAuditForStoreTest(actionPlan, "wtae_0000000000000202", "confirmation_invalidated", invalidateDecision.ConfirmationID)
	invalidateAudit.AuditRef = invalidateDecision.AuditRef
	if err = actionStore.DecidePlan(actionContext, &actionPlan, invalidateDecision, invalidateAudit); err != nil {
		t.Fatalf("invalidate PostgreSQL workflow HTTP tool action plan: %v", err)
	}
	expireContext := workflowHTTPToolActionTestContext()
	expireContext.RequestContext = ctx
	expireContext.AuditRef = "audit_workflow_http_tool_postgres_expire_create"
	expiredActionPlan := workflowHTTPToolActionPlanForStoreTest(t, expireContext, "wtap_0000000000000210")
	if err = actionStore.CreatePlan(expireContext, &expiredActionPlan, workflowHTTPToolAuditForStoreTest(expiredActionPlan, "wtae_0000000000000210", "confirmation_requested")); err != nil {
		t.Fatalf("create PostgreSQL workflow HTTP tool expiry plan: %v", err)
	}
	expiredActionPlan.RecordVersion = 2
	expiredActionPlan.Status = WorkflowHTTPToolActionStatusExpired
	expiredAt := "2026-07-16T09:16:00Z"
	expiredActionPlan.LastDecisionByActorRef, expiredActionPlan.LastDecisionAt = &systemActor, &expiredAt
	expireDecision := workflowHTTPToolSystemDecisionForStoreTest(expiredActionPlan, "wtcd_0000000000000210", workflowHTTPToolConfirmationExpire)
	expireDecision.DecidedAt = expiredAt
	expireDecision.AuditRef = "audit_workflow_http_tool_postgres_expire"
	expireContext.AuditRef = expireDecision.AuditRef
	expireAudit := workflowHTTPToolAuditForStoreTest(expiredActionPlan, "wtae_0000000000000211", "confirmation_expired", expireDecision.ConfirmationID)
	expireAudit.AuditRef = expireDecision.AuditRef
	if err = actionStore.DecidePlan(expireContext, &expiredActionPlan, expireDecision, expireAudit); err != nil {
		t.Fatalf("expire PostgreSQL workflow HTTP tool action plan: %v", err)
	}
	deferContext := workflowHTTPToolActionTestContext()
	deferContext.RequestContext = ctx
	deferContext.AuditRef = "audit_workflow_http_tool_postgres_defer_create"
	deferredActionPlan := workflowHTTPToolActionPlanForStoreTest(t, deferContext, "wtap_0000000000000220")
	if err = actionStore.CreatePlan(deferContext, &deferredActionPlan, workflowHTTPToolAuditForStoreTest(deferredActionPlan, "wtae_0000000000000220", "confirmation_requested")); err != nil {
		t.Fatalf("create PostgreSQL workflow HTTP tool defer plan: %v", err)
	}
	deferredActionPlan.RecordVersion = 2
	deferredActionPlan.Status = WorkflowHTTPToolActionStatusDeferred
	deferredAt := "2026-07-16T09:01:00Z"
	deferredActionPlan.LastDecisionByActorRef, deferredActionPlan.LastDecisionAt = &actionActor, &deferredAt
	deferDecision := workflowHTTPToolDecisionForStoreTest(deferredActionPlan, "wtcd_0000000000000220", WorkflowHTTPToolConfirmationDefer, actionActor)
	deferDecision.AuditRef = "audit_workflow_http_tool_postgres_defer"
	deferContext.AuditRef = deferDecision.AuditRef
	deferAudit := workflowHTTPToolAuditForStoreTest(deferredActionPlan, "wtae_0000000000000221", "confirmation_deferred", deferDecision.ConfirmationID)
	deferAudit.AuditRef = deferDecision.AuditRef
	if err = actionStore.DecidePlan(deferContext, &deferredActionPlan, deferDecision, deferAudit); err != nil {
		t.Fatalf("defer pending PostgreSQL workflow HTTP tool action plan: %v", err)
	}
	repeatedDefer := cloneWorkflowHTTPToolActionPlan(deferredActionPlan)
	repeatedDefer.RecordVersion = 3
	repeatedAt := "2026-07-16T09:02:00Z"
	repeatedDefer.LastDecisionAt = &repeatedAt
	repeatedDecision := workflowHTTPToolDecisionForStoreTest(repeatedDefer, "wtcd_0000000000000221", WorkflowHTTPToolConfirmationDefer, actionActor)
	repeatedDecision.DecidedAt = repeatedAt
	repeatedDecision.AuditRef = "audit_workflow_http_tool_postgres_repeat_defer"
	repeatedContext := deferContext
	repeatedContext.AuditRef = repeatedDecision.AuditRef
	repeatedAudit := workflowHTTPToolAuditForStoreTest(repeatedDefer, "wtae_0000000000000222", "confirmation_deferred", repeatedDecision.ConfirmationID)
	repeatedAudit.AuditRef = repeatedDecision.AuditRef
	if err = actionStore.DecidePlan(repeatedContext, &repeatedDefer, repeatedDecision, repeatedAudit); !errors.Is(err, errWorkflowHTTPToolActionConflict) {
		t.Fatalf("repeated PostgreSQL deferred-to-deferred transition must conflict, got %v", err)
	}
	var actionToolVersion, decisionToolVersionMin, decisionToolVersionMax, auditToolVersionMin, auditToolVersionMax int
	if err = runtimePool.QueryRow(ctx, `SELECT tool_version FROM workflow_http_tool_action_plans WHERE plan_id=$1`, actionPlan.PlanID).Scan(&actionToolVersion); err != nil {
		t.Fatalf("read PostgreSQL workflow HTTP tool plan version: %v", err)
	}
	if err = runtimePool.QueryRow(ctx, `SELECT min(tool_version),max(tool_version) FROM workflow_http_tool_confirmation_decisions WHERE plan_id=$1`, actionPlan.PlanID).Scan(&decisionToolVersionMin, &decisionToolVersionMax); err != nil {
		t.Fatalf("read PostgreSQL workflow HTTP tool decision versions: %v", err)
	}
	if err = runtimePool.QueryRow(ctx, `SELECT min(tool_version),max(tool_version) FROM workflow_http_tool_execution_audits WHERE plan_id=$1`, actionPlan.PlanID).Scan(&auditToolVersionMin, &auditToolVersionMax); err != nil {
		t.Fatalf("read PostgreSQL workflow HTTP tool audit versions: %v", err)
	}
	if actionToolVersion != workflowHTTPToolVersion || decisionToolVersionMin != workflowHTTPToolVersion || decisionToolVersionMax != workflowHTTPToolVersion ||
		auditToolVersionMin != workflowHTTPToolVersion || auditToolVersionMax != workflowHTTPToolVersion {
		t.Fatalf("PostgreSQL workflow HTTP tool evidence version drifted: plan=%d decisions=%d..%d audits=%d..%d",
			actionToolVersion, decisionToolVersionMin, decisionToolVersionMax, auditToolVersionMin, auditToolVersionMax)
	}
	var deferredStatus string
	var deferredVersion, deferredDecisionCount, deferredAuditCount int
	if err = runtimePool.QueryRow(ctx, `SELECT status,record_version FROM workflow_http_tool_action_plans WHERE plan_id=$1`, deferredActionPlan.PlanID).Scan(&deferredStatus, &deferredVersion); err != nil {
		t.Fatalf("read PostgreSQL deferred workflow HTTP tool plan: %v", err)
	}
	if err = runtimePool.QueryRow(ctx, `SELECT count(*) FROM workflow_http_tool_confirmation_decisions WHERE plan_id=$1`, deferredActionPlan.PlanID).Scan(&deferredDecisionCount); err != nil {
		t.Fatalf("count PostgreSQL deferred workflow HTTP tool decisions: %v", err)
	}
	if err = runtimePool.QueryRow(ctx, `SELECT count(*) FROM workflow_http_tool_execution_audits WHERE plan_id=$1`, deferredActionPlan.PlanID).Scan(&deferredAuditCount); err != nil {
		t.Fatalf("count PostgreSQL deferred workflow HTTP tool audits: %v", err)
	}
	if deferredStatus != string(WorkflowHTTPToolActionStatusDeferred) || deferredVersion != 2 || deferredDecisionCount != 1 || deferredAuditCount != 2 {
		t.Fatalf("repeated PostgreSQL defer wrote partial evidence: status=%s version=%d decisions=%d audits=%d",
			deferredStatus, deferredVersion, deferredDecisionCount, deferredAuditCount)
	}
	executionContext := workflowHTTPToolActionTestContext()
	executionContext.RequestContext = ctx
	executionContext.ScopeGrants = append(executionContext.ScopeGrants, workflowHTTPToolExecutionRequiredScopes...)
	executionContext.AuditRef = "audit_workflow_http_tool_postgres_execution_create"
	executionPlan := workflowHTTPToolActionPlanForStoreTest(t, executionContext, "wtap_0000000000000230")
	if err = actionStore.CreatePlan(executionContext, &executionPlan, workflowHTTPToolAuditForStoreTest(executionPlan, "wtae_0000000000000230", "confirmation_requested")); err != nil {
		t.Fatalf("create PostgreSQL workflow HTTP tool execution plan: %v", err)
	}
	executionPlan.RecordVersion = 2
	executionPlan.Status = WorkflowHTTPToolActionStatusApproved
	executionActor, executionDecidedAt := executionContext.ActorRef, "2026-07-16T09:01:00Z"
	executionPlan.LastDecisionByActorRef, executionPlan.LastDecisionAt = &executionActor, &executionDecidedAt
	executionDecision := workflowHTTPToolDecisionForStoreTest(executionPlan, "wtcd_0000000000000230", WorkflowHTTPToolConfirmationApprove, executionActor)
	executionDecision.AuditRef = "audit_workflow_http_tool_postgres_execution_decision"
	executionContext.AuditRef = executionDecision.AuditRef
	executionDecisionAudit := workflowHTTPToolAuditForStoreTest(executionPlan, "wtae_0000000000000231", "confirmation_recorded", executionDecision.ConfirmationID)
	executionDecisionAudit.AuditRef = executionDecision.AuditRef
	if err = actionStore.DecidePlan(executionContext, &executionPlan, executionDecision, executionDecisionAudit); err != nil {
		t.Fatalf("approve PostgreSQL workflow HTTP tool execution plan: %v", err)
	}
	executionContext.AuditRef = "audit_workflow_http_tool_postgres_execution_claim"
	executionStore := newPostgresWorkflowHTTPToolExecutionStore(runtimePool)
	type executionWinner struct {
		attempt WorkflowHTTPToolExecutionAttempt
		run     WorkflowRunRecord
	}
	var executionWG sync.WaitGroup
	executionWinners := make(chan executionWinner, 8)
	executionErrors := make(chan error, 8)
	for index := 0; index < 8; index++ {
		executionWG.Add(1)
		go func(index int) {
			defer executionWG.Done()
			candidatePlan := cloneWorkflowHTTPToolActionPlan(executionPlan)
			candidatePlan.Status = WorkflowHTTPToolActionStatusConsumed
			candidatePlan.RecordVersion++
			attempt, run, audit := workflowHTTPToolClaimFixture(executionContext, candidatePlan, executionDecision, index+100)
			claimErr := executionStore.ClaimExecution(executionContext, &candidatePlan, executionDecision, &attempt, &run, audit)
			if claimErr == nil {
				executionWinners <- executionWinner{attempt: attempt, run: run}
				return
			}
			if !errors.Is(claimErr, errWorkflowHTTPToolExecutionConflict) {
				executionErrors <- claimErr
			}
		}(index)
	}
	executionWG.Wait()
	close(executionWinners)
	close(executionErrors)
	for claimErr := range executionErrors {
		t.Fatalf("unexpected PostgreSQL workflow HTTP tool claim error: %v", claimErr)
	}
	claimedExecutions := make([]executionWinner, 0, 1)
	for winner := range executionWinners {
		claimedExecutions = append(claimedExecutions, winner)
	}
	if len(claimedExecutions) != 1 {
		t.Fatalf("expected one PostgreSQL workflow HTTP tool claim winner, got %d", len(claimedExecutions))
	}
	claimedExecution := claimedExecutions[0]
	var executionAttemptCount, executionStartedAuditCount int
	var executionPlanStatus string
	if err = runtimePool.QueryRow(ctx, `SELECT count(*) FROM workflow_http_tool_execution_attempts WHERE plan_id=$1`, executionPlan.PlanID).Scan(&executionAttemptCount); err != nil {
		t.Fatal(err)
	}
	if err = runtimePool.QueryRow(ctx, `SELECT status FROM workflow_http_tool_action_plans WHERE plan_id=$1`, executionPlan.PlanID).Scan(&executionPlanStatus); err != nil {
		t.Fatal(err)
	}
	if err = runtimePool.QueryRow(ctx, `SELECT count(*) FROM workflow_http_tool_execution_audits WHERE plan_id=$1 AND event_kind='tool_execution_started'`, executionPlan.PlanID).Scan(&executionStartedAuditCount); err != nil {
		t.Fatal(err)
	}
	if executionAttemptCount != 1 || executionPlanStatus != string(WorkflowHTTPToolActionStatusConsumed) || executionStartedAuditCount != 1 {
		t.Fatalf("PostgreSQL execution claim evidence drifted: attempts=%d plan=%s started_audits=%d", executionAttemptCount, executionPlanStatus, executionStartedAuditCount)
	}
	base := time.Now().UTC().Add(-time.Minute)
	for index := 0; index < 3; index++ {
		record := workflowRunHistoryTestRecord(runContext, "run_pg_"+string(rune('a'+index)), "draft_pg", base.Add(time.Duration(index)*time.Second))
		if err = store.UpsertRun(runContext, &record); err != nil {
			t.Fatal(err)
		}
		record.Status = WorkflowRunStatusSucceeded
		record.CompletedAt = workflowRunTimestamp(base.Add(time.Duration(index)*time.Second + time.Millisecond))
		record.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
		if err = store.UpsertRun(runContext, &record); err != nil {
			t.Fatal(err)
		}
	}
	page, err := store.ListRuns(runContext, WorkflowRunListFilter{Limit: 2})
	if err != nil || len(page.Records) != 2 || !page.HasMore || page.Records[0].RunID != "run_pg_c" {
		t.Fatalf("unexpected PostgreSQL page: %#v %v", page, err)
	}
	legacy := workflowRunHistoryTestRecord(runContext, "run_pg_legacy", "draft_pg", base.Add(-time.Second))
	legacy.SchemaVersion = workflowRunRecordLegacySchemaVersion
	legacy.Diagnostic = nil
	if err = store.UpsertRun(runContext, &legacy); err != nil {
		t.Fatalf("write legacy run: %v", err)
	}
	if decoded, found, readErr := store.ReadRun(runContext, legacy.RunID); readErr != nil || !found || decoded.SchemaVersion != workflowRunRecordLegacySchemaVersion || decoded.Diagnostic != nil {
		t.Fatalf("legacy v0 read compatibility failed: found=%v record=%#v err=%v", found, decoded, readErr)
	}
	diagnostic := workflowRunHistoryTestRecord(runContext, "run_pg_diagnostic", "draft_pg", base.Add(4*time.Second))
	diagnostic.SelectedProvider = "mock"
	diagnostic.SelectedModel = "model-diagnostic"
	if err = store.UpsertRun(runContext, &diagnostic); err != nil {
		t.Fatal(err)
	}
	diagnostic.Status = WorkflowRunStatusFailed
	diagnostic.FailureCode = WorkflowRunFailureGatewayFailed
	diagnostic.FailureSummary = "Gateway timed out while executing the workflow model node."
	diagnostic.CompletedAt = workflowRunTimestamp(base.Add(5 * time.Second))
	setWorkflowRunFailureDiagnostic(&diagnostic, diagnostic.FailureCode, "node_model", WorkflowRunGatewayFailureTimeout)
	diagnostic.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
	if err = store.UpsertRun(runContext, &diagnostic); err != nil {
		t.Fatal(err)
	}
	diagnosticPage, err := store.ListRuns(runContext, WorkflowRunListFilter{Limit: 10, FailureCode: WorkflowRunFailureGatewayFailed, FailureBoundary: WorkflowRunFailureBoundaryGateway, Provider: "mock", Model: "model-diagnostic"})
	if err != nil || len(diagnosticPage.Records) != 1 || diagnosticPage.Records[0].RunID != diagnostic.RunID {
		t.Fatalf("diagnostic PostgreSQL filter failed: %#v %v", diagnosticPage, err)
	}
	evaluationStore := newPostgresWorkflowEvaluationStore(runtimePool)
	evaluationService := newWorkflowEvaluationService(evaluationStore, store)
	evaluationService.newCaseID = func() (string, error) { return "eval_pg_restart", nil }
	evaluation := evaluationService.Create(runContext, WorkflowEvaluationCreateRequest{Name: "PostgreSQL restart review", BaselineRunID: diagnostic.RunID, Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "run_pg_c", ExpectedClassification: WorkflowRunComparisonImprovement}}})
	if evaluation.FailureCode != "" || evaluation.Case == nil {
		t.Fatalf("create PostgreSQL evaluation case: %#v", evaluation)
	}
	revisedEvaluation := evaluationService.Revise(runContext, evaluation.Case.CaseID, WorkflowEvaluationRevisionRequest{ExpectedVersion: 1, RevisionKind: WorkflowEvaluationRevisionCase, Name: "PostgreSQL revised review", BaselineRunID: diagnostic.RunID, Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "run_pg_c", ExpectedClassification: WorkflowRunComparisonChanged}}})
	if revisedEvaluation.FailureCode != "" || revisedEvaluation.Case == nil || revisedEvaluation.Case.Version != 2 {
		t.Fatalf("revise PostgreSQL evaluation case: %#v", revisedEvaluation)
	}
	evaluationService.newCaseID = func() (string, error) { return "eval_pg_cas", nil }
	casCase := evaluationService.Create(runContext, WorkflowEvaluationCreateRequest{Name: "PostgreSQL evaluation CAS", BaselineRunID: diagnostic.RunID, Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "run_pg_c", ExpectedClassification: WorkflowRunComparisonImprovement}}})
	if casCase.Case == nil {
		t.Fatalf("create evaluation CAS case: %#v", casCase)
	}
	var evaluationWG sync.WaitGroup
	evaluationResults := make(chan WorkflowEvaluationResult, 8)
	for index := 0; index < 8; index++ {
		evaluationWG.Add(1)
		go func(index int) {
			defer evaluationWG.Done()
			evaluationResults <- evaluationService.Revise(runContext, casCase.Case.CaseID, WorkflowEvaluationRevisionRequest{ExpectedVersion: 1, RevisionKind: WorkflowEvaluationRevisionCase, Name: fmt.Sprintf("PostgreSQL evaluation CAS %d", index), BaselineRunID: diagnostic.RunID, Expectations: casCase.Case.Expectations})
		}(index)
	}
	evaluationWG.Wait()
	close(evaluationResults)
	evaluationWinners := 0
	for result := range evaluationResults {
		if result.FailureCode == "" {
			evaluationWinners++
		} else if result.FailureCode != WorkflowEvaluationFailureVersionConflict {
			t.Fatalf("unexpected evaluation CAS result: %#v", result)
		}
	}
	if evaluationWinners != 1 {
		t.Fatalf("expected one evaluation CAS winner, got %d", evaluationWinners)
	}
	suiteStore := newPostgresWorkflowEvaluationSuiteStore(runtimePool)
	suiteService := newWorkflowEvaluationSuiteService(suiteStore, evaluationService)
	suiteService.newSuiteID = func() (string, error) { return "suite_pg_restart", nil }
	suite := suiteService.Create(runContext, WorkflowEvaluationSuiteCreateRequest{Name: "PostgreSQL release review", CaseRefs: []WorkflowEvaluationSuiteCaseRef{{CaseID: evaluation.Case.CaseID, Version: 1}}})
	if suite.Suite == nil || suite.FailureCode != "" {
		t.Fatalf("create PostgreSQL evaluation suite: %#v", suite)
	}
	suiteReview := suiteService.Review(runContext, suite.Suite.SuiteID)
	if suiteReview.Review == nil || suiteReview.Review.Outcome != "passed" {
		t.Fatalf("review PostgreSQL evaluation suite: %#v", suiteReview)
	}
	var suiteWG sync.WaitGroup
	suiteResults := make(chan WorkflowEvaluationSuiteResult, 8)
	for index := 0; index < 8; index++ {
		suiteWG.Add(1)
		go func() {
			defer suiteWG.Done()
			suiteResults <- suiteService.Decide(runContext, suite.Suite.SuiteID, WorkflowEvaluationDecisionRequest{ExpectedDecisionVersion: 0, Decision: "approved", ReviewDigest: suiteReview.Review.ReviewDigest})
		}()
	}
	suiteWG.Wait()
	close(suiteResults)
	suiteWinners := 0
	for result := range suiteResults {
		if result.FailureCode == "" {
			suiteWinners++
		} else if result.FailureCode != WorkflowEvaluationSuiteFailureDecisionConflict {
			t.Fatalf("unexpected suite CAS result: %#v", result)
		}
	}
	if suiteWinners != 1 {
		t.Fatalf("expected one suite decision CAS winner, got %d", suiteWinners)
	}
	other := runContext
	other.TenantRef = "tenant_other"
	if scoped, err := store.ListRuns(other, WorkflowRunListFilter{Limit: 10}); err != nil || len(scoped.Records) != 0 {
		t.Fatalf("scope leaked: %#v %v", scoped, err)
	}
	ragRepository := newPostgresWorkflowRAGSnapshotRepository(runtimePool)
	runWorkflowRAGSnapshotLifecycle(t, ragRepository)
	ragExecutionService, ragExecutionBridge, ragRunContext, ragExecutionSnapshot, ragExecutionDraft := workflowRAGExecutionStoreFixture(
		t, store, ragRepository, "run_pgretrieval00001",
	)
	ragExecution := ragExecutionService.Execute(ragRunContext, WorkflowRAGExecutionRequest{
		DraftID: ragExecutionDraft.DraftID, DraftVersion: ragExecutionDraft.DraftVersion,
		InputText: "official retrieval guidance", Model: "mock-rag",
	})
	if ragExecution.FailureCode != "" || ragExecution.Record == nil || ragExecution.Record.Status != WorkflowRunStatusSucceeded || ragExecutionBridge.handleCalls != 1 {
		t.Fatalf("PostgreSQL workflow RAG execution failed: %#v calls=%d", ragExecution, ragExecutionBridge.handleCalls)
	}
	ragCandidateService, ragCandidateBridge, _, _, _ := workflowRAGExecutionStoreFixtureFromExisting(
		t, store, ragRepository, ragExecutionDraft, "run_pgretrieval00002",
	)
	ragCandidate := ragCandidateService.Execute(ragRunContext, WorkflowRAGExecutionRequest{
		DraftID: ragExecutionDraft.DraftID, DraftVersion: ragExecutionDraft.DraftVersion,
		InputText: "official retrieval guidance", Model: "mock-rag",
	})
	if ragCandidate.FailureCode != "" || ragCandidate.Record == nil || ragCandidate.Record.Status != WorkflowRunStatusSucceeded || ragCandidateBridge.handleCalls != 1 {
		t.Fatalf("PostgreSQL workflow RAG candidate execution failed: %#v calls=%d", ragCandidate, ragCandidateBridge.handleCalls)
	}
	ragComparison := newWorkflowExecutorService(nil, nil, store).CompareRuns(ragRunContext, ragExecution.Record.RunID, ragCandidate.Record.RunID)
	if ragComparison.FailureCode != "" || ragComparison.Comparison == nil ||
		ragComparison.Comparison.SchemaVersion != workflowRAGRunComparisonSchemaVersion ||
		ragComparison.Comparison.Classification != WorkflowRunComparisonUnchanged || ragComparison.Comparison.Retrieval == nil {
		t.Fatalf("PostgreSQL workflow RAG comparison failed: %#v", ragComparison)
	}
	ragEvaluationService := newWorkflowEvaluationService(evaluationStore, store)
	ragEvaluationService.newCaseID = func() (string, error) { return "eval_pg_rag_restart", nil }
	ragEvaluation := ragEvaluationService.Create(ragRunContext, WorkflowEvaluationCreateRequest{
		Name: "PostgreSQL RAG regression review", BaselineRunID: ragExecution.Record.RunID,
		Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: ragCandidate.Record.RunID, ExpectedClassification: WorkflowRunComparisonUnchanged}},
	})
	if ragEvaluation.FailureCode != "" || ragEvaluation.Case == nil {
		t.Fatalf("create PostgreSQL RAG evaluation case: %#v", ragEvaluation)
	}
	ragReview := ragEvaluationService.Review(ragRunContext, ragEvaluation.Case.CaseID)
	if ragReview.FailureCode != "" || ragReview.Review == nil || ragReview.Review.Outcome != "passed" ||
		ragReview.Review.RunProfile != workflowRAGComparisonProfile {
		t.Fatalf("review PostgreSQL RAG evaluation case: %#v", ragReview)
	}
	ragSuiteService := newWorkflowEvaluationSuiteService(suiteStore, ragEvaluationService)
	ragSuiteService.newSuiteID = func() (string, error) { return "suite_pg_rag_restart", nil }
	ragSuite := ragSuiteService.Create(ragRunContext, WorkflowEvaluationSuiteCreateRequest{
		Name:     "PostgreSQL RAG regression suite",
		CaseRefs: []WorkflowEvaluationSuiteCaseRef{{CaseID: ragEvaluation.Case.CaseID, Version: ragEvaluation.Case.Version}},
	})
	if ragSuite.FailureCode != "" || ragSuite.Suite == nil {
		t.Fatalf("create PostgreSQL RAG evaluation suite: %#v", ragSuite)
	}
	ragPlan, ragPlanFailure := buildWorkflowRAGExecutionPlan(ragExecutionDraft, ragExecutionDraft.DraftVersion)
	if ragPlanFailure != "" {
		t.Fatalf("build PostgreSQL stale RAG plan: %s", ragPlanFailure)
	}
	ragDraftDigest, _ := workflowRAGDraftDigest(ragExecutionDraft)
	ragStale := newWorkflowRAGRunRecord(
		ragRunContext,
		WorkflowRAGExecutionRequest{DraftID: ragExecutionDraft.DraftID, DraftVersion: ragExecutionDraft.DraftVersion, InputText: "official retrieval guidance", Model: "mock-rag"},
		ragExecutionDraft, ragDraftDigest, ragPlan, ragExecutionSnapshot, workflowRAGLexicalProfile(), workflowRAGTestSelection(),
		"run_pgragstale000001", time.Now().UTC().Add(-workflowExecutorDefaultMaxRuntime-time.Second),
	)
	if err = store.UpsertRun(ragRunContext, &ragStale); err != nil {
		t.Fatalf("seed PostgreSQL stale workflow RAG run: %v", err)
	}
	if _, mutationErr := runtimePool.Exec(ctx, `UPDATE workflow_rag_snapshot_versions SET snapshot_version=3 WHERE snapshot_id='rags_aaaaaaaaaaaaaaaa' AND snapshot_version=2`); mutationErr == nil {
		t.Fatal("PostgreSQL runtime role mutated an immutable workflow RAG snapshot version")
	}
	if _, mutationErr := runtimePool.Exec(ctx, `DELETE FROM workflow_rag_snapshot_fragments WHERE snapshot_id='rags_aaaaaaaaaaaaaaaa'`); mutationErr == nil {
		t.Fatal("PostgreSQL runtime role deleted immutable workflow RAG snapshot fragments")
	}
	if _, mutationErr := runtimePool.Exec(ctx, `UPDATE workflow_rag_execution_audits SET event_kind='snapshot_created' WHERE snapshot_id='rags_aaaaaaaaaaaaaaaa'`); mutationErr == nil {
		t.Fatal("PostgreSQL runtime role mutated append-only workflow RAG snapshot audits")
	}
	runtimePool.Close()
	reopened, err := workflowrunmigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	store = newPostgresWorkflowRunStore(reopened)
	evaluationStore = newPostgresWorkflowEvaluationStore(reopened)
	actionStore = newPostgresWorkflowHTTPToolActionStore(reopened)
	executionStore = newPostgresWorkflowHTTPToolExecutionStore(reopened)
	restoredRAG := newWorkflowRAGSnapshotService(newPostgresWorkflowRAGSnapshotRepository(reopened)).Read(
		workflowRAGTestContext(), "rags_aaaaaaaaaaaaaaaa", 2,
	)
	if restoredRAG.FailureCode != "" || restoredRAG.Record == nil || restoredRAG.Record.LifecycleState != workflowRAGSnapshotArchived ||
		len(restoredRAG.Record.Fragments) != 1 || restoredRAG.Record.Fragments[0].Content != "version two replacement content" {
		t.Fatalf("restart workflow RAG snapshot recovery failed: %#v", restoredRAG)
	}
	restartedRAGRepository := newPostgresWorkflowRAGSnapshotRepository(reopened)
	restoredRAGRun, ragRunFound, ragRunErr := store.ReadRun(ragRunContext, ragExecution.Record.RunID)
	if ragRunErr != nil || !ragRunFound || restoredRAGRun.Status != WorkflowRunStatusSucceeded || restoredRAGRun.RAGAnswer != nil ||
		restoredRAGRun.RetrievalAttempt == nil || len(restoredRAGRun.RetrievalAttempt.CitationRefs) != 1 {
		t.Fatalf("restart workflow RAG v3 run recovery failed: %#v found=%t err=%v", restoredRAGRun, ragRunFound, ragRunErr)
	}
	ragReconciliation, ragRestartBridge, _, _, _ := workflowRAGExecutionStoreFixtureFromExisting(
		t, store, restartedRAGRepository, ragExecutionDraft, "run_pgragunused00001",
	)
	ragRankCalls := 0
	ragReconciliation.rank = func(string, []WorkflowRAGFragment, int) WorkflowRAGRankingResult {
		ragRankCalls++
		return WorkflowRAGRankingResult{}
	}
	ragReconciled := ragReconciliation.ReconcileStale(ragRunContext)
	if ragReconciled.FailureCode != "" || ragReconciled.Reconciled != 1 || ragRankCalls != 0 || ragRestartBridge.handleCalls != 0 {
		t.Fatalf("PostgreSQL RAG restart reconciliation retried execution: %#v rank=%d gateway=%d", ragReconciled, ragRankCalls, ragRestartBridge.handleCalls)
	}
	restoredRAGStale, ragStaleFound, ragStaleErr := store.ReadRun(ragRunContext, ragStale.RunID)
	if ragStaleErr != nil || !ragStaleFound || restoredRAGStale.Status != WorkflowRunStatusFailed ||
		restoredRAGStale.FailureCode != WorkflowRunFailureCode(WorkflowRAGFailureInterrupted) {
		t.Fatalf("PostgreSQL stale workflow RAG run was not closed: %#v found=%t err=%v", restoredRAGStale, ragStaleFound, ragStaleErr)
	}
	var ragExecutionAuditCount int
	if err = reopened.QueryRow(ctx, `SELECT count(*) FROM workflow_rag_execution_audits WHERE snapshot_id=$1`, ragExecutionSnapshot.SnapshotID).Scan(&ragExecutionAuditCount); err != nil || ragExecutionAuditCount != 6 {
		t.Fatalf("PostgreSQL workflow RAG execution audits did not survive restart: count=%d err=%v", ragExecutionAuditCount, err)
	}
	if recoveredAction, actionFound, actionErr := actionStore.ReadPlan(actionContext, actionPlan.PlanID); actionErr != nil || !actionFound || recoveredAction.Status != WorkflowHTTPToolActionStatusInvalidated || recoveredAction.RecordVersion != 3 {
		t.Fatalf("restart workflow HTTP tool action recovery failed: found=%v plan=%#v err=%v", actionFound, recoveredAction, actionErr)
	}
	if recoveredExpired, actionFound, actionErr := actionStore.ReadPlan(expireContext, expiredActionPlan.PlanID); actionErr != nil || !actionFound || recoveredExpired.Status != WorkflowHTTPToolActionStatusExpired || recoveredExpired.RecordVersion != 2 {
		t.Fatalf("restart expired workflow HTTP tool action recovery failed: found=%v plan=%#v err=%v", actionFound, recoveredExpired, actionErr)
	}
	reconciliationService := workflowHTTPToolExecutionService{store: executionStore}
	reconciliationService.now = func() time.Time { return time.Date(2026, 7, 16, 9, 4, 0, 0, time.UTC) }
	reconciliationService.newID = func(string) (string, error) { return "wtae_0000000000000232", nil }
	reconciledExecution := reconciliationService.ReconcileStale(executionContext, executionPlan.PlanID)
	if reconciledExecution.FailureCode != WorkflowRunFailureToolOutcomeUnknown || reconciledExecution.Record == nil ||
		reconciledExecution.Record.RunID != claimedExecution.run.RunID || reconciledExecution.Record.Status != WorkflowRunStatusOutcomeUnknown ||
		reconciledExecution.Record.RecordVersion != 2 {
		t.Fatalf("restart PostgreSQL execution reconciliation failed: %#v", reconciledExecution)
	}
	if _, _, pendingFound, pendingErr := executionStore.ReadClaimedExecution(executionContext, executionPlan.PlanID); pendingErr != nil || pendingFound {
		t.Fatalf("reconciled PostgreSQL execution remained claimed: found=%t err=%v", pendingFound, pendingErr)
	}
	recovered, found, err := store.ReadRun(runContext, "run_pg_c")
	if err != nil || !found || recovered.Status != WorkflowRunStatusSucceeded {
		t.Fatalf("restart recovery failed: %#v %v", recovered, err)
	}
	comparisonService := newWorkflowExecutorService(nil, nil, store)
	comparison := comparisonService.CompareRuns(runContext, diagnostic.RunID, "run_pg_c")
	if comparison.FailureCode != "" || comparison.Comparison == nil || comparison.Comparison.Classification != WorkflowRunComparisonImprovement {
		t.Fatalf("restart comparison failed: %#v", comparison)
	}
	if scopedComparison := comparisonService.CompareRuns(other, diagnostic.RunID, "run_pg_c"); scopedComparison.FailureCode != WorkflowRunFailureRecordNotFound {
		t.Fatalf("restart comparison leaked cross scope: %#v", scopedComparison)
	}
	restartedEvaluationService := newWorkflowEvaluationService(evaluationStore, store)
	restartedReview := restartedEvaluationService.ReviewVersion(runContext, evaluation.Case.CaseID, 1)
	currentReview := restartedEvaluationService.Review(runContext, evaluation.Case.CaseID)
	if restartedReview.FailureCode != "" || restartedReview.Review == nil || restartedReview.Review.Outcome != "passed" || restartedReview.Review.Version != 1 || currentReview.Review == nil || currentReview.Review.Outcome != "mismatch" || currentReview.Review.Version != 2 {
		t.Fatalf("restart versioned evaluation failed: old=%#v current=%#v", restartedReview, currentReview)
	}
	revisionHistory := restartedEvaluationService.ListRevisions(runContext, evaluation.Case.CaseID, WorkflowEvaluationRevisionListRequest{Limit: 10})
	if revisionHistory.FailureCode != "" || len(revisionHistory.Revisions) != 2 || revisionHistory.Revisions[0].Version != 2 {
		t.Fatalf("restart revision history failed: %#v", revisionHistory)
	}
	if leaked := restartedEvaluationService.Read(other, evaluation.Case.CaseID); leaked.FailureCode != WorkflowEvaluationFailureNotFound {
		t.Fatalf("evaluation case leaked scope: %#v", leaked)
	}
	restartedSuiteService := newWorkflowEvaluationSuiteService(newPostgresWorkflowEvaluationSuiteStore(reopened), restartedEvaluationService)
	recoveredSuite := restartedSuiteService.Read(runContext, suite.Suite.SuiteID)
	recoveredDecisions := restartedSuiteService.ListDecisions(runContext, suite.Suite.SuiteID, WorkflowEvaluationDecisionListRequest{})
	if recoveredSuite.Suite == nil || recoveredSuite.Suite.CurrentDecisionVersion != 1 || recoveredSuite.Suite.CurrentDecision != "approved" || len(recoveredDecisions.Decisions) != 1 {
		t.Fatalf("restart suite recovery failed: suite=%#v decisions=%#v", recoveredSuite, recoveredDecisions)
	}
	if leaked := restartedSuiteService.Read(other, suite.Suite.SuiteID); leaked.FailureCode != WorkflowEvaluationSuiteFailureNotFound {
		t.Fatalf("evaluation suite leaked scope: %#v", leaked)
	}
	restartedRAGEvaluationService := newWorkflowEvaluationService(newPostgresWorkflowEvaluationStore(reopened), store)
	restartedRAGReview := restartedRAGEvaluationService.Review(ragRunContext, ragEvaluation.Case.CaseID)
	if restartedRAGReview.FailureCode != "" || restartedRAGReview.Review == nil || restartedRAGReview.Review.Outcome != "passed" ||
		restartedRAGReview.Review.RunProfile != workflowRAGComparisonProfile ||
		restartedRAGReview.Review.Items[0].ComparisonSchemaVersion != workflowRAGRunComparisonSchemaVersion {
		t.Fatalf("restart PostgreSQL RAG evaluation review failed: %#v", restartedRAGReview)
	}
	restartedRAGSuiteService := newWorkflowEvaluationSuiteService(newPostgresWorkflowEvaluationSuiteStore(reopened), restartedRAGEvaluationService)
	restartedRAGSuiteReview := restartedRAGSuiteService.Review(ragRunContext, ragSuite.Suite.SuiteID)
	if restartedRAGSuiteReview.FailureCode != "" || restartedRAGSuiteReview.Review == nil || restartedRAGSuiteReview.Review.Outcome != "passed" ||
		len(restartedRAGSuiteReview.Review.Items) != 1 || restartedRAGSuiteReview.Review.Items[0].RunProfile != workflowRAGComparisonProfile {
		t.Fatalf("restart PostgreSQL RAG evaluation suite review failed: %#v", restartedRAGSuiteReview)
	}
	running := workflowRunHistoryTestRecord(runContext, "run_pg_concurrent", "draft_pg", time.Now().UTC())
	if err = store.UpsertRun(runContext, &running); err != nil {
		t.Fatal(err)
	}
	left, right := cloneWorkflowRunRecord(running), cloneWorkflowRunRecord(running)
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, candidate := range []*WorkflowRunRecord{&left, &right} {
		wg.Add(1)
		go func(value *WorkflowRunRecord) {
			defer wg.Done()
			value.Status = WorkflowRunStatusFailed
			value.FailureCode = WorkflowRunFailureGatewayFailed
			value.FailureSummary = "Gateway failed."
			value.CompletedAt = workflowRunTimestamp(time.Now())
			setWorkflowRunFailureDiagnostic(value, value.FailureCode, "node_model", WorkflowRunGatewayFailureUnavailable)
			value.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
			results <- store.UpsertRun(runContext, value)
		}(candidate)
	}
	wg.Wait()
	close(results)
	successes := 0
	for result := range results {
		if result == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("expected one PostgreSQL CAS winner, got %d", successes)
	}

	reopened.Close()
	if result := restartedRAGEvaluationService.Review(ragRunContext, ragEvaluation.Case.CaseID); result.FailureCode != WorkflowEvaluationFailureStoreUnavailable {
		t.Fatalf("closed PostgreSQL RAG evaluation store fell back: %#v", result)
	}
	if result := restartedRAGSuiteService.Read(ragRunContext, ragSuite.Suite.SuiteID); result.FailureCode != WorkflowEvaluationSuiteFailureStoreUnavailable {
		t.Fatalf("closed PostgreSQL RAG evaluation suite store fell back: %#v", result)
	}
	noFallbackRAG, noFallbackBridge, _, _, _ := workflowRAGExecutionStoreFixtureFromExisting(
		t, store, restartedRAGRepository, ragExecutionDraft, "run_pgragnofallback1",
	)
	noFallbackResult := noFallbackRAG.Execute(ragRunContext, WorkflowRAGExecutionRequest{
		DraftID: ragExecutionDraft.DraftID, DraftVersion: ragExecutionDraft.DraftVersion,
		InputText: "official retrieval guidance", Model: "mock-rag",
	})
	if noFallbackResult.FailureCode != WorkflowRunFailureCode(WorkflowRAGFailureStoreUnavailable) || noFallbackResult.Record != nil || noFallbackBridge.handleCalls != 0 {
		t.Fatalf("closed PostgreSQL workflow RAG backend fell back or called Gateway: %#v calls=%d", noFallbackResult, noFallbackBridge.handleCalls)
	}
	if _, err = pool.Exec(ctx, "UPDATE workflow_run_schema_versions SET migration_checksum='sha256:mismatch' WHERE component=$1", workflowrunmigrations.Component); err != nil {
		t.Fatal(err)
	}
	if mismatch, inspectErr := workflowrunmigrations.Inspect(ctx, pool); inspectErr != nil || mismatch.MigrationState != workflowrunmigrations.MigrationStateMismatch {
		t.Fatalf("marker mismatch not detected: %#v %v", mismatch, inspectErr)
	}
	if _, err = pool.Exec(ctx, "UPDATE workflow_run_schema_versions SET migration_checksum=$1 WHERE component=$2", workflowrunmigrations.ExpectedChecksum(), workflowrunmigrations.Component); err != nil {
		t.Fatal(err)
	}
	if _, err = workflowrunmigrations.RollbackForDevTest(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if _, err = workflowrunmigrations.Apply(ctx, pool); err != nil {
		t.Fatalf("reapply after rollback: %v", err)
	}
	if _, err = workflowrunmigrations.RollbackForDevTest(ctx, pool); err != nil {
		t.Fatal(err)
	}
	legacySQL, err := os.ReadFile("../../migrations/workflow_runs/0001_workflow_runs.up.sql")
	if err != nil {
		t.Fatalf("read legacy workflow run migration: %v", err)
	}
	if _, err = pool.Exec(ctx, string(legacySQL)); err != nil {
		t.Fatalf("apply legacy workflow run migration: %v", err)
	}
	legacyChecksum := fmt.Sprintf("sha256:%x", sha256.Sum256(legacySQL))
	if _, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS workflow_run_schema_versions (component text PRIMARY KEY, migration_id text NOT NULL, store_schema_version text NOT NULL, migration_checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workflow_run_schema_versions(component,migration_id,store_schema_version,migration_checksum) VALUES($1,'0001_workflow_runs','workflow_run_store_v1',$2)`, workflowrunmigrations.Component, legacyChecksum); err != nil {
		t.Fatal(err)
	}
	if pending, inspectErr := workflowrunmigrations.Inspect(ctx, pool); inspectErr != nil || pending.MigrationState != workflowrunmigrations.MigrationStatePending {
		t.Fatalf("legacy marker was not recognized as pending: %#v %v", pending, inspectErr)
	}
	if upgraded, upgradeErr := workflowrunmigrations.Apply(ctx, pool); upgradeErr != nil || upgraded.MigrationState != workflowrunmigrations.MigrationStateApplied {
		t.Fatalf("legacy migration upgrade failed: %#v %v", upgraded, upgradeErr)
	}
	if _, err = workflowrunmigrations.RollbackForDevTest(ctx, pool); err != nil {
		t.Fatal(err)
	}
	diagnosticsSQL, err := os.ReadFile("../../migrations/workflow_runs/0002_workflow_run_diagnostics.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, string(legacySQL)+"\n"+string(diagnosticsSQL)); err != nil {
		t.Fatalf("apply diagnostics migration: %v", err)
	}
	diagnosticsChecksum := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(string(legacySQL)+"\n"+string(diagnosticsSQL))))
	if _, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS workflow_run_schema_versions (component text PRIMARY KEY, migration_id text NOT NULL, store_schema_version text NOT NULL, migration_checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workflow_run_schema_versions(component,migration_id,store_schema_version,migration_checksum) VALUES($1,'0002_workflow_run_diagnostics','workflow_run_store_v2',$2)`, workflowrunmigrations.Component, diagnosticsChecksum); err != nil {
		t.Fatal(err)
	}
	if pending, inspectErr := workflowrunmigrations.Inspect(ctx, pool); inspectErr != nil || pending.MigrationState != workflowrunmigrations.MigrationStatePending {
		t.Fatalf("diagnostics marker was not pending: %#v %v", pending, inspectErr)
	}
	if upgraded, upgradeErr := workflowrunmigrations.Apply(ctx, pool); upgradeErr != nil || upgraded.MigrationState != workflowrunmigrations.MigrationStateApplied {
		t.Fatalf("diagnostics migration upgrade failed: %#v %v", upgraded, upgradeErr)
	}
	if _, err = workflowrunmigrations.RollbackForDevTest(ctx, pool); err != nil {
		t.Fatal(err)
	}
	evaluationSQL, err := os.ReadFile("../../migrations/workflow_runs/0003_workflow_evaluation_cases.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, string(legacySQL)+"\n"+string(diagnosticsSQL)+"\n"+string(evaluationSQL)); err != nil {
		t.Fatalf("apply evaluation migration: %v", err)
	}
	legacyCase := WorkflowEvaluationCase{SchemaVersion: workflowEvaluationLegacySchema, CaseID: "eval_pg_legacy_backfill", Name: "legacy backfill", WorkspaceID: runContext.WorkspaceID, ApplicationID: runContext.ApplicationID, BaselineRunID: diagnostic.RunID, Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "run_pg_c", ExpectedClassification: WorkflowRunComparisonImprovement}}, CreatedAt: workflowRunTimestamp(time.Now().UTC()), ActorRef: runContext.ActorRef, RequestID: runContext.RequestID, AuditRef: runContext.AuditRef}
	legacyCasePayload, _ := json.Marshal(legacyCase)
	if _, err = pool.Exec(ctx, `INSERT INTO workflow_evaluation_cases(tenant_ref,workspace_id,application_id,case_id,baseline_run_id,created_at,sanitized_case_record) VALUES($1,$2,$3,$4,$5,$6,$7)`, runContext.TenantRef, runContext.WorkspaceID, runContext.ApplicationID, legacyCase.CaseID, legacyCase.BaselineRunID, legacyCase.CreatedAt, legacyCasePayload); err != nil {
		t.Fatalf("seed legacy evaluation case: %v", err)
	}
	evaluationChecksum := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(string(legacySQL)+"\n"+string(diagnosticsSQL)+"\n"+string(evaluationSQL))))
	if _, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS workflow_run_schema_versions (component text PRIMARY KEY, migration_id text NOT NULL, store_schema_version text NOT NULL, migration_checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workflow_run_schema_versions(component,migration_id,store_schema_version,migration_checksum) VALUES($1,'0003_workflow_evaluation_cases','workflow_run_store_v3',$2)`, workflowrunmigrations.Component, evaluationChecksum); err != nil {
		t.Fatal(err)
	}
	if pending, inspectErr := workflowrunmigrations.Inspect(ctx, pool); inspectErr != nil || pending.MigrationState != workflowrunmigrations.MigrationStatePending {
		t.Fatalf("evaluation marker was not pending: %#v %v", pending, inspectErr)
	}
	if upgraded, upgradeErr := workflowrunmigrations.Apply(ctx, pool); upgradeErr != nil || upgraded.MigrationState != workflowrunmigrations.MigrationStateApplied {
		t.Fatalf("evaluation migration upgrade failed: %#v %v", upgraded, upgradeErr)
	}
	var backfilledVersion, backfilledCount int
	if err = pool.QueryRow(ctx, `SELECT current_version,(SELECT count(*) FROM workflow_evaluation_case_revisions WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND case_id=$4) FROM workflow_evaluation_cases WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND case_id=$4`, runContext.TenantRef, runContext.WorkspaceID, runContext.ApplicationID, legacyCase.CaseID).Scan(&backfilledVersion, &backfilledCount); err != nil || backfilledVersion != 1 || backfilledCount != 1 {
		t.Fatalf("legacy evaluation backfill failed: version=%d count=%d err=%v", backfilledVersion, backfilledCount, err)
	}
	if _, err = workflowrunmigrations.RollbackForDevTest(ctx, pool); err != nil {
		t.Fatal(err)
	}
	versioningSQL, err := os.ReadFile("../../migrations/workflow_runs/0004_workflow_evaluation_case_revisions.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	throughV4 := string(legacySQL) + "\n" + string(diagnosticsSQL) + "\n" + string(evaluationSQL) + "\n" + string(versioningSQL)
	if _, err = pool.Exec(ctx, throughV4); err != nil {
		t.Fatalf("apply case versioning migration: %v", err)
	}
	versioningChecksum := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(throughV4)))
	if _, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS workflow_run_schema_versions (component text PRIMARY KEY, migration_id text NOT NULL, store_schema_version text NOT NULL, migration_checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workflow_run_schema_versions(component,migration_id,store_schema_version,migration_checksum) VALUES($1,'0004_workflow_evaluation_case_revisions','workflow_run_store_v4',$2)`, workflowrunmigrations.Component, versioningChecksum); err != nil {
		t.Fatal(err)
	}
	if pending, inspectErr := workflowrunmigrations.Inspect(ctx, pool); inspectErr != nil || pending.MigrationState != workflowrunmigrations.MigrationStatePending {
		t.Fatalf("case versioning marker was not pending: %#v %v", pending, inspectErr)
	}
	if upgraded, upgradeErr := workflowrunmigrations.Apply(ctx, pool); upgradeErr != nil || upgraded.MigrationState != workflowrunmigrations.MigrationStateApplied {
		t.Fatalf("suite migration upgrade failed: %#v %v", upgraded, upgradeErr)
	}
	if _, err = workflowrunmigrations.RollbackForDevTest(ctx, pool); err != nil {
		t.Fatal(err)
	}
	suiteSQL, err := os.ReadFile("../../migrations/workflow_runs/0005_workflow_evaluation_suites.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	throughV5 := throughV4 + "\n" + string(suiteSQL)
	if _, err = pool.Exec(ctx, throughV5); err != nil {
		t.Fatalf("apply evaluation suite migration: %v", err)
	}
	suiteChecksum := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(throughV5)))
	if _, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS workflow_run_schema_versions (component text PRIMARY KEY, migration_id text NOT NULL, store_schema_version text NOT NULL, migration_checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workflow_run_schema_versions(component,migration_id,store_schema_version,migration_checksum) VALUES($1,'0005_workflow_evaluation_suites','workflow_run_store_v5',$2)`, workflowrunmigrations.Component, suiteChecksum); err != nil {
		t.Fatal(err)
	}
	if pending, inspectErr := workflowrunmigrations.Inspect(ctx, pool); inspectErr != nil || pending.MigrationState != workflowrunmigrations.MigrationStatePending {
		t.Fatalf("evaluation suite marker was not pending for tool action migration: %#v %v", pending, inspectErr)
	}
	if upgraded, upgradeErr := workflowrunmigrations.Apply(ctx, pool); upgradeErr != nil || upgraded.MigrationState != workflowrunmigrations.MigrationStateApplied {
		t.Fatalf("tool action migration upgrade failed: %#v %v", upgraded, upgradeErr)
	}
	if _, err = workflowrunmigrations.RollbackForDevTest(ctx, pool); err != nil {
		t.Fatal(err)
	}
	toolActionSQL, err := os.ReadFile("../../migrations/workflow_runs/0006_workflow_http_tool_actions.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	toolExecutionSQL, err := os.ReadFile("../../migrations/workflow_runs/0007_workflow_http_tool_execution.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	ragSnapshotSQL, err := os.ReadFile("../../migrations/workflow_runs/0008_workflow_rag_snapshots.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	throughV8 := throughV5 + "\n" + string(toolActionSQL) + "\n" + string(toolExecutionSQL) + "\n" + string(ragSnapshotSQL)
	if _, err = pool.Exec(ctx, throughV8); err != nil {
		t.Fatalf("apply workflow RAG snapshot migration: %v", err)
	}
	ragSnapshotChecksum := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(throughV8)))
	if _, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS workflow_run_schema_versions (component text PRIMARY KEY, migration_id text NOT NULL, store_schema_version text NOT NULL, migration_checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workflow_run_schema_versions(component,migration_id,store_schema_version,migration_checksum) VALUES($1,'0008_workflow_rag_snapshots','workflow_run_store_v8',$2)`, workflowrunmigrations.Component, ragSnapshotChecksum); err != nil {
		t.Fatal(err)
	}
	if pending, inspectErr := workflowrunmigrations.Inspect(ctx, pool); inspectErr != nil || pending.MigrationState != workflowrunmigrations.MigrationStatePending {
		t.Fatalf("workflow RAG snapshot marker was not pending for execution migration: %#v %v", pending, inspectErr)
	}
	if upgraded, upgradeErr := workflowrunmigrations.Apply(ctx, pool); upgradeErr != nil || upgraded.MigrationState != workflowrunmigrations.MigrationStateApplied {
		t.Fatalf("workflow RAG execution migration upgrade failed: %#v %v", upgraded, upgradeErr)
	}
	var retrievalAuditConstraintCount int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM pg_constraint WHERE conname='workflow_rag_execution_audits_event_kind_check' AND pg_get_constraintdef(oid) LIKE '%retrieval_started%'`).Scan(&retrievalAuditConstraintCount); err != nil || retrievalAuditConstraintCount != 1 {
		t.Fatalf("workflow RAG execution audit constraint was not upgraded: count=%d err=%v", retrievalAuditConstraintCount, err)
	}
}

func resetPostgresWorkflowRunSchema(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	if _, err := pool.Exec(ctx, `DROP TABLE IF EXISTS workflow_rag_execution_audits, workflow_rag_snapshot_fragments, workflow_rag_snapshot_versions, workflow_rag_snapshot_resources, workflow_http_tool_execution_attempts, workflow_http_tool_confirmation_decisions, workflow_http_tool_execution_audits, workflow_http_tool_action_plans, workflow_evaluation_suite_decisions, workflow_evaluation_suites, workflow_evaluation_case_revisions, workflow_evaluation_cases, workflow_run_records`); err != nil {
		t.Fatalf("reset workflow run integration tables: %v", err)
	}
	if _, err := pool.Exec(ctx, `DROP FUNCTION IF EXISTS reject_workflow_rag_snapshot_append_only_mutation()`); err != nil {
		t.Fatalf("reset workflow RAG append-only guard: %v", err)
	}
	if _, err := pool.Exec(ctx, `DROP FUNCTION IF EXISTS reject_workflow_http_tool_append_only_mutation()`); err != nil {
		t.Fatalf("reset workflow HTTP tool append-only guard: %v", err)
	}
	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS workflow_run_schema_versions (component text PRIMARY KEY, migration_id text NOT NULL, store_schema_version text NOT NULL, migration_checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		t.Fatalf("prepare workflow run integration marker: %v", err)
	}
	if _, err := pool.Exec(ctx, `DELETE FROM workflow_run_schema_versions WHERE component=$1`, workflowrunmigrations.Component); err != nil {
		t.Fatalf("reset workflow run integration marker: %v", err)
	}
}
