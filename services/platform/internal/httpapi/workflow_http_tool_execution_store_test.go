package httpapi

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestSQLiteWorkflowHTTPToolExecutionRestartReconciliationIsMetadataOnly(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "workflow-http-tool-reconciliation.db")
	runtime := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	actionStore := newSQLiteWorkflowHTTPToolActionStore(runtime.DB())
	executionStore := newSQLiteWorkflowHTTPToolExecutionStore(runtime.DB())
	ctx, approvedPlan, confirmation := seedApprovedWorkflowHTTPToolPlan(t, actionStore)
	ctx.ScopeGrants = append(ctx.ScopeGrants, workflowHTTPToolExecutionRequiredScopes...)
	plan := cloneWorkflowHTTPToolActionPlan(approvedPlan)
	plan.Status = WorkflowHTTPToolActionStatusConsumed
	plan.RecordVersion++
	attempt, run, claimAudit := workflowHTTPToolClaimFixture(ctx, plan, confirmation, 0)
	if err := executionStore.ClaimExecution(ctx, &plan, confirmation, &attempt, &run, claimAudit); err != nil {
		t.Fatalf("claim SQLite execution: %v", err)
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("close SQLite runtime before reconciliation: %v", err)
	}

	reopened := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	t.Cleanup(func() { _ = reopened.Close() })
	service := workflowHTTPToolExecutionService{store: newSQLiteWorkflowHTTPToolExecutionStore(reopened.DB())}
	service.now = func() time.Time { return time.Date(2026, 7, 16, 9, 2, 31, 0, time.UTC) }
	service.newID = func(string) (string, error) { return "wtae_3333333333333333", nil }
	result := service.ReconcileStale(ctx, plan.PlanID)
	if result.FailureCode != WorkflowRunFailureToolOutcomeUnknown || result.Record == nil ||
		result.Record.Status != WorkflowRunStatusOutcomeUnknown || result.Record.RecordVersion != 2 {
		t.Fatalf("restart reconciliation did not close stale SQLite metadata: %#v", result)
	}
	var attemptStatus, runStatus, actorSource string
	var auditCount int
	if err := reopened.DB().QueryRow(`SELECT status FROM workflow_http_tool_execution_attempts WHERE plan_id=?`, plan.PlanID).Scan(&attemptStatus); err != nil {
		t.Fatal(err)
	}
	if err := reopened.DB().QueryRow(`SELECT run_status FROM workflow_run_records WHERE run_id=?`, run.RunID).Scan(&runStatus); err != nil {
		t.Fatal(err)
	}
	if err := reopened.DB().QueryRow(`SELECT count(*) FROM workflow_http_tool_execution_audits WHERE plan_id=?`, plan.PlanID).Scan(&auditCount); err != nil {
		t.Fatal(err)
	}
	if err := reopened.DB().QueryRow(`SELECT json_extract(sanitized_execution_audit,'$.actor_source')
	 FROM workflow_http_tool_execution_audits WHERE audit_id=?`, "wtae_3333333333333333").Scan(&actorSource); err != nil {
		t.Fatal(err)
	}
	if attemptStatus != "outcome_unknown" || runStatus != "outcome_unknown" || auditCount != 4 || actorSource != "system" {
		t.Fatalf("unexpected restart reconciliation evidence: attempt=%s run=%s audits=%d actor=%s", attemptStatus, runStatus, auditCount, actorSource)
	}
}

func TestSQLiteWorkflowHTTPToolExecutionClaimCompletionAndRestart(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "workflow-http-tool-execution.db")
	runtime := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	actionStore := newSQLiteWorkflowHTTPToolActionStore(runtime.DB())
	executionStore := newSQLiteWorkflowHTTPToolExecutionStore(runtime.DB())
	ctx, approvedPlan, confirmation := seedApprovedWorkflowHTTPToolPlan(t, actionStore)
	plan := cloneWorkflowHTTPToolActionPlan(approvedPlan)
	plan.Status = WorkflowHTTPToolActionStatusConsumed
	plan.RecordVersion++
	attempt, run, claimAudit := workflowHTTPToolClaimFixture(ctx, plan, confirmation, 0)
	if err := executionStore.ClaimExecution(ctx, &plan, confirmation, &attempt, &run, claimAudit); err != nil {
		t.Fatalf("claim SQLite execution: %v", err)
	}
	losingPlan := cloneWorkflowHTTPToolActionPlan(plan)
	losingAttempt, losingRun, losingAudit := workflowHTTPToolClaimFixture(ctx, losingPlan, confirmation, 1)
	if err := executionStore.ClaimExecution(ctx, &losingPlan, confirmation, &losingAttempt, &losingRun, losingAudit); !errors.Is(err, errWorkflowHTTPToolExecutionConflict) {
		t.Fatalf("expected duplicate SQLite claim conflict, got %v", err)
	}
	completeWorkflowHTTPToolFixture(ctx, &attempt, &run, WorkflowHTTPToolAttemptOutcomeUnknown)
	completionAudit := workflowHTTPToolCompletionAuditFixture(ctx, attempt, run, "wtae_bbbbbbbbbbbbbbbb")
	if err := executionStore.CompleteExecution(ctx, &attempt, &run, completionAudit); err != nil {
		t.Fatalf("complete SQLite execution: %v", err)
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("close SQLite execution runtime: %v", err)
	}

	reopened := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	t.Cleanup(func() { _ = reopened.Close() })
	storedRun, found, err := newSQLiteWorkflowRunStore(reopened.DB()).ReadRun(workflowRunContextFromToolAction(ctx), run.RunID)
	if err != nil || !found || storedRun.Status != WorkflowRunStatusOutcomeUnknown ||
		storedRun.ToolAttempt == nil || storedRun.ToolAttempt.Status != WorkflowHTTPToolAttemptOutcomeUnknown {
		t.Fatalf("unexpected restarted SQLite run: %#v found=%t err=%v", storedRun, found, err)
	}
	storedPlan, found, err := newSQLiteWorkflowHTTPToolActionStore(reopened.DB()).ReadPlan(ctx, plan.PlanID)
	if err != nil || !found || storedPlan.Status != WorkflowHTTPToolActionStatusConsumed || storedPlan.RecordVersion != 3 {
		t.Fatalf("unexpected restarted SQLite plan: %#v found=%t err=%v", storedPlan, found, err)
	}
	var attemptStatus, runStatus string
	var attemptCount, auditCount int
	if err := reopened.DB().QueryRow(`SELECT count(*),min(status) FROM workflow_http_tool_execution_attempts WHERE plan_id=?`, plan.PlanID).Scan(&attemptCount, &attemptStatus); err != nil {
		t.Fatal(err)
	}
	if err := reopened.DB().QueryRow(`SELECT run_status FROM workflow_run_records WHERE run_id=?`, run.RunID).Scan(&runStatus); err != nil {
		t.Fatal(err)
	}
	if err := reopened.DB().QueryRow(`SELECT count(*) FROM workflow_http_tool_execution_audits WHERE plan_id=?`, plan.PlanID).Scan(&auditCount); err != nil {
		t.Fatal(err)
	}
	if attemptCount != 1 || attemptStatus != "outcome_unknown" || runStatus != "outcome_unknown" || auditCount != 4 {
		t.Fatalf("SQLite execution evidence mismatch: attempts=%d attempt_status=%s run_status=%s audits=%d", attemptCount, attemptStatus, runStatus, auditCount)
	}
}

func TestMemoryWorkflowHTTPToolExecutionClaimIsAtomicAndSingleUse(t *testing.T) {
	runStore := newMemoryWorkflowRunStore(100)
	actionStore := newMemoryWorkflowHTTPToolActionStore(&runStore.mu)
	executionStore := newMemoryWorkflowHTTPToolExecutionStore(actionStore, runStore)
	ctx, approvedPlan, confirmation := seedApprovedMemoryWorkflowHTTPToolPlan(t, actionStore)

	const contenders = 12
	type winner struct {
		attempt WorkflowHTTPToolExecutionAttempt
		run     WorkflowRunRecord
	}
	winners := make(chan winner, contenders)
	errorsFound := make(chan error, contenders)
	var wait sync.WaitGroup
	for index := 0; index < contenders; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			plan := cloneWorkflowHTTPToolActionPlan(approvedPlan)
			plan.Status = WorkflowHTTPToolActionStatusConsumed
			plan.RecordVersion++
			attempt, run, audit := workflowHTTPToolClaimFixture(ctx, plan, confirmation, index)
			err := executionStore.ClaimExecution(ctx, &plan, confirmation, &attempt, &run, audit)
			if err == nil {
				winners <- winner{attempt: attempt, run: run}
				return
			}
			if !errors.Is(err, errWorkflowHTTPToolExecutionConflict) {
				errorsFound <- err
			}
		}(index)
	}
	wait.Wait()
	close(winners)
	close(errorsFound)
	for err := range errorsFound {
		t.Fatalf("unexpected claim error: %v", err)
	}
	claimed := make([]winner, 0, 1)
	for value := range winners {
		claimed = append(claimed, value)
	}
	if len(claimed) != 1 {
		t.Fatalf("expected exactly one successful claim, got %d", len(claimed))
	}
	storedPlan, found, err := actionStore.ReadPlan(ctx, approvedPlan.PlanID)
	if err != nil || !found || storedPlan.Status != WorkflowHTTPToolActionStatusConsumed || storedPlan.RecordVersion != 3 {
		t.Fatalf("unexpected consumed plan: %#v found=%t err=%v", storedPlan, found, err)
	}
	storedRun, found, err := runStore.ReadRun(workflowRunContextFromToolAction(ctx), claimed[0].run.RunID)
	if err != nil || !found || storedRun.RecordVersion != 1 || storedRun.Status != WorkflowRunStatusRunning {
		t.Fatalf("unexpected claimed run: %#v found=%t err=%v", storedRun, found, err)
	}
}

func TestMemoryWorkflowHTTPToolExecutionCompletionIsAtomic(t *testing.T) {
	runStore := newMemoryWorkflowRunStore(100)
	actionStore := newMemoryWorkflowHTTPToolActionStore(&runStore.mu)
	executionStore := newMemoryWorkflowHTTPToolExecutionStore(actionStore, runStore)
	ctx, approvedPlan, confirmation := seedApprovedMemoryWorkflowHTTPToolPlan(t, actionStore)
	plan := cloneWorkflowHTTPToolActionPlan(approvedPlan)
	plan.Status = WorkflowHTTPToolActionStatusConsumed
	plan.RecordVersion++
	attempt, run, claimAudit := workflowHTTPToolClaimFixture(ctx, plan, confirmation, 0)
	if err := executionStore.ClaimExecution(ctx, &plan, confirmation, &attempt, &run, claimAudit); err != nil {
		t.Fatalf("claim execution: %v", err)
	}
	completeWorkflowHTTPToolFixture(ctx, &attempt, &run, WorkflowHTTPToolAttemptSucceeded)
	completionAudit := workflowHTTPToolCompletionAuditFixture(ctx, attempt, run, "wtae_aaaaaaaaaaaaaaaa")
	if err := executionStore.CompleteExecution(ctx, &attempt, &run, completionAudit); err != nil {
		t.Fatalf("complete execution: %v", err)
	}
	if run.RecordVersion != 2 {
		t.Fatalf("expected terminal run record version 2, got %d", run.RecordVersion)
	}
	stored, found, err := runStore.ReadRun(workflowRunContextFromToolAction(ctx), run.RunID)
	if err != nil || !found || stored.Status != WorkflowRunStatusSucceeded ||
		stored.ToolAttempt == nil || stored.ToolAttempt.Status != WorkflowHTTPToolAttemptSucceeded {
		t.Fatalf("unexpected terminal run: %#v found=%t err=%v", stored, found, err)
	}
	if err := executionStore.CompleteExecution(ctx, &attempt, &run, completionAudit); !errors.Is(err, errWorkflowHTTPToolExecutionConflict) {
		t.Fatalf("expected repeated completion conflict, got %v", err)
	}
}

func TestWorkflowRunV2StorageRoundTripAndLegacyIsolation(t *testing.T) {
	ctx := workflowHTTPToolActionTestContext()
	plan := workflowHTTPToolActionPlanForStoreTest(t, ctx, "wtap_1234567890abcdef")
	plan.Status = WorkflowHTTPToolActionStatusApproved
	plan.RecordVersion = 2
	decisionAt := "2026-07-16T09:01:00Z"
	actor := ctx.ActorRef
	plan.LastDecisionAt, plan.LastDecisionByActorRef = &decisionAt, &actor
	confirmation := workflowHTTPToolDecisionForStoreTest(plan, "wtcd_1234567890abcdef", WorkflowHTTPToolConfirmationApprove, actor)
	plan.Status = WorkflowHTTPToolActionStatusConsumed
	plan.RecordVersion++
	attempt, run, _ := workflowHTTPToolClaimFixture(ctx, plan, confirmation, 0)
	run.RecordVersion = 1
	payload, _, _, err := encodeWorkflowRunStorageRecord(run)
	if err != nil {
		t.Fatalf("encode run v2: %v", err)
	}
	decoded, err := decodeWorkflowRunStorageRecord(workflowRunContextFromToolAction(ctx), payload)
	if err != nil || decoded.ToolAttempt == nil || decoded.ToolAttempt.AttemptID != attempt.AttemptID ||
		decoded.SchemaVersion != workflowRunRecordToolSchemaVersion {
		t.Fatalf("unexpected decoded run v2: %#v err=%v", decoded, err)
	}
	legacy := run
	legacy.SchemaVersion = workflowRunRecordSchemaVersion
	legacy.PlanID, legacy.ConfirmationID, legacy.TenantRef, legacy.ToolAttempt = "", "", "", nil
	legacy.SideEffects.ToolCalls, legacy.SideEffects.ConfirmationCalls = 0, 0
	legacy.Diagnostic.ToolFailureCategory = ""
	if err := validateWorkflowRunStoreRecord(workflowRunContextFromToolAction(ctx), &legacy); err != nil {
		t.Fatalf("legacy v1 invariant regressed: %v", err)
	}
	legacy.SideEffects.ToolCalls = 1
	if err := validateWorkflowRunStoreRecord(workflowRunContextFromToolAction(ctx), &legacy); !errors.Is(err, errWorkflowRunStoreContract) {
		t.Fatalf("legacy v1 accepted tool side effect: %v", err)
	}
}

func seedApprovedMemoryWorkflowHTTPToolPlan(
	t *testing.T,
	store *memoryWorkflowHTTPToolActionStore,
) (WorkflowHTTPToolActionContext, WorkflowHTTPToolActionPlan, WorkflowHTTPToolConfirmationDecision) {
	t.Helper()
	ctx := workflowHTTPToolActionTestContext()
	plan := workflowHTTPToolActionPlanForStoreTest(t, ctx, "wtap_1234567890abcdef")
	createAudit := workflowHTTPToolAuditForStoreTest(plan, "wtae_1234567890abcdef", "confirmation_requested")
	if err := store.CreatePlan(ctx, &plan, createAudit); err != nil {
		t.Fatalf("create action plan: %v", err)
	}
	plan.Status = WorkflowHTTPToolActionStatusApproved
	plan.RecordVersion++
	decisionAt := "2026-07-16T09:01:00Z"
	actor := ctx.ActorRef
	plan.LastDecisionAt, plan.LastDecisionByActorRef = &decisionAt, &actor
	confirmation := workflowHTTPToolDecisionForStoreTest(plan, "wtcd_1234567890abcdef", WorkflowHTTPToolConfirmationApprove, actor)
	decisionAudit := workflowHTTPToolAuditForStoreTest(plan, "wtae_abcdef1234567890", "confirmation_recorded", confirmation.ConfirmationID)
	decisionAudit.AuditRef = ctx.AuditRef
	if err := store.DecidePlan(ctx, &plan, confirmation, decisionAudit); err != nil {
		t.Fatalf("approve action plan: %v", err)
	}
	return ctx, plan, confirmation
}

func seedApprovedWorkflowHTTPToolPlan(
	t *testing.T,
	store workflowHTTPToolActionStore,
) (WorkflowHTTPToolActionContext, WorkflowHTTPToolActionPlan, WorkflowHTTPToolConfirmationDecision) {
	t.Helper()
	ctx := workflowHTTPToolActionTestContext()
	plan := workflowHTTPToolActionPlanForStoreTest(t, ctx, "wtap_1234567890abcdef")
	createAudit := workflowHTTPToolAuditForStoreTest(plan, "wtae_1234567890abcdef", "confirmation_requested")
	if err := store.CreatePlan(ctx, &plan, createAudit); err != nil {
		t.Fatalf("create action plan: %v", err)
	}
	plan.Status = WorkflowHTTPToolActionStatusApproved
	plan.RecordVersion++
	decisionAt := "2026-07-16T09:01:00Z"
	actor := ctx.ActorRef
	plan.LastDecisionAt, plan.LastDecisionByActorRef = &decisionAt, &actor
	confirmation := workflowHTTPToolDecisionForStoreTest(plan, "wtcd_1234567890abcdef", WorkflowHTTPToolConfirmationApprove, actor)
	decisionCtx := ctx
	decisionCtx.AuditRef = "audit_workflow_http_tool_decision"
	confirmation.AuditRef = decisionCtx.AuditRef
	decisionAudit := workflowHTTPToolAuditForStoreTest(plan, "wtae_abcdef1234567890", "confirmation_recorded", confirmation.ConfirmationID)
	decisionAudit.AuditRef = decisionCtx.AuditRef
	if err := store.DecidePlan(decisionCtx, &plan, confirmation, decisionAudit); err != nil {
		t.Fatalf("approve action plan: %v", err)
	}
	ctx.AuditRef = "audit_workflow_http_tool_execution"
	return ctx, plan, confirmation
}

func workflowHTTPToolClaimFixture(
	ctx WorkflowHTTPToolActionContext,
	plan WorkflowHTTPToolActionPlan,
	confirmation WorkflowHTTPToolConfirmationDecision,
	index int,
) (WorkflowHTTPToolExecutionAttempt, WorkflowRunRecord, WorkflowHTTPToolExecutionAudit) {
	claimedAt := "2026-07-16T09:02:00Z"
	attemptID := fmt.Sprintf("wtea_%016x", index+1)
	runID := fmt.Sprintf("run_%016x", index+1)
	attempt := WorkflowHTTPToolExecutionAttempt{
		AttemptID: attemptID, NodeID: plan.NodeID, ToolID: plan.ToolID,
		DefinitionDigest: plan.DefinitionDigest, ProfileID: plan.ProfileID, ProfileDigest: plan.ProfileDigest,
		ToolPlanDigest: plan.ToolPlanDigest, ConfirmationID: confirmation.ConfirmationID,
		Status: WorkflowHTTPToolAttemptClaimed, ClaimedAt: claimedAt, OutputProjection: map[string]any{},
	}
	run := WorkflowRunRecord{
		SchemaVersion: workflowRunRecordToolSchemaVersion, RunID: runID, PlanID: plan.PlanID,
		ConfirmationID: confirmation.ConfirmationID, TenantRef: ctx.TenantRef,
		DraftID: plan.DraftID, DraftVersion: plan.DraftVersion, WorkspaceID: ctx.WorkspaceID,
		ApplicationID: ctx.ApplicationID, Status: WorkflowRunStatusRunning, StartedAt: claimedAt,
		Nodes: []WorkflowRunNodeRecord{
			{NodeID: "node_prompt", NodeType: "prompt", Label: "Prompt", Status: WorkflowRunNodeStatusPending},
			{NodeID: plan.NodeID, NodeType: "http_tool", Label: "HTTP Tool", Status: WorkflowRunNodeStatusPending, PredecessorNodeIDs: []string{"node_prompt"}},
			{NodeID: "node_model", NodeType: "llm", Label: "Model", Status: WorkflowRunNodeStatusPending, PredecessorNodeIDs: []string{plan.NodeID}, ProviderRef: "profile:mock-workflow"},
			{NodeID: "node_output", NodeType: "output", Label: "Output", Status: WorkflowRunNodeStatusPending, PredecessorNodeIDs: []string{"node_model"}},
		},
		ToolAttempt: &attempt, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef, ActorRef: ctx.ActorRef,
		SideEffects: WorkflowRunSideEffects{ToolCalls: 1, ConfirmationCalls: 1},
		Diagnostic: &WorkflowRunDiagnostic{
			TerminalWriteState: WorkflowRunTerminalWritePending, GatewayFailureCategory: WorkflowRunGatewayFailureNone,
			ToolFailureCategory: WorkflowHTTPToolFailureNone, ObservedAt: claimedAt,
		},
	}
	audit := workflowHTTPToolExecutionAuditFixture(ctx, plan, confirmation, attempt, run, "wtae_1111111111111111")
	return attempt, run, audit
}

func workflowHTTPToolExecutionAuditFixture(
	ctx WorkflowHTTPToolActionContext,
	plan WorkflowHTTPToolActionPlan,
	confirmation WorkflowHTTPToolConfirmationDecision,
	attempt WorkflowHTTPToolExecutionAttempt,
	run WorkflowRunRecord,
	eventID string,
) WorkflowHTTPToolExecutionAudit {
	attemptID, runID, confirmationID := attempt.AttemptID, run.RunID, confirmation.ConfirmationID
	return WorkflowHTTPToolExecutionAudit{
		SchemaVersion: workflowHTTPToolAuditSchema, EventID: eventID, EventKind: "tool_execution_started",
		OccurredAt: attempt.ClaimedAt, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
		ApplicationID: ctx.ApplicationID, DraftID: plan.DraftID, DraftVersion: plan.DraftVersion,
		NodeID: plan.NodeID, PlanID: plan.PlanID, ConfirmationID: &confirmationID,
		ExecutionAttemptID: &attemptID, RunID: &runID, ToolID: plan.ToolID, ToolVersion: plan.ToolVersion,
		DefinitionDigest: plan.DefinitionDigest, ProfileID: plan.ProfileID, ProfileDigest: plan.ProfileDigest,
		ToolPlanDigest: plan.ToolPlanDigest, ActorRef: ctx.ActorRef, ActorSource: "human",
		RequestID: ctx.RequestID, AuditRef: "audit_" + eventID, AttemptStatus: "claimed",
		SideEffects: WorkflowHTTPToolAuditSideEffects{},
	}
}

func completeWorkflowHTTPToolFixture(
	ctx WorkflowHTTPToolActionContext,
	attempt *WorkflowHTTPToolExecutionAttempt,
	run *WorkflowRunRecord,
	status WorkflowHTTPToolAttemptStatus,
) {
	completedAt := "2026-07-16T09:02:01Z"
	attempt.Status, attempt.CompletedAt, attempt.DurationMS = status, completedAt, 1000
	run.CompletedAt = completedAt
	run.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
	run.Diagnostic.ObservedAt = completedAt
	switch status {
	case WorkflowHTTPToolAttemptSucceeded:
		attempt.HTTPStatusClass = "2xx"
		attempt.ResponseBytes = 120
		attempt.OutputProjection = map[string]any{
			"resource_key": "docs/radishflow/overview", "title": "RadishFlow", "summary": "Reviewed", "updated_at": "2026-07-16T09:00:00Z",
		}
		run.Status = WorkflowRunStatusSucceeded
		run.Output = "Reviewed"
		run.Diagnostic.ToolFailureCategory = WorkflowHTTPToolFailureNone
	case WorkflowHTTPToolAttemptOutcomeUnknown:
		attempt.FailureCode = WorkflowRunFailureToolOutcomeUnknown
		attempt.OutputProjection = map[string]any{}
		run.Status = WorkflowRunStatusOutcomeUnknown
		run.FailureCode = WorkflowRunFailureToolOutcomeUnknown
		run.FailureSummary = "Workflow tool execution outcome requires manual review."
		run.Diagnostic.ToolFailureCategory = WorkflowHTTPToolFailureOutcomeUnknown
		run.Diagnostic.FailureBoundary = WorkflowRunFailureBoundaryToolStore
		run.Diagnostic.FailureStage = "outcome_reconciliation"
		run.Diagnostic.RecommendedReviewAction = WorkflowRunReviewToolOutcome
	case WorkflowHTTPToolAttemptFailed:
		attempt.FailureCode = WorkflowRunFailureToolTransport
		attempt.OutputProjection = map[string]any{}
		run.Status = WorkflowRunStatusFailed
		run.FailureCode = attempt.FailureCode
		run.FailureSummary = "Workflow tool transport failed."
		run.Diagnostic.ToolFailureCategory = WorkflowHTTPToolFailureTransport
		run.Diagnostic.FailureBoundary = WorkflowRunFailureBoundaryToolTransport
		run.Diagnostic.FailureStage = "http_request"
	}
	run.ToolAttempt = attempt
	_ = ctx
}

func workflowHTTPToolCompletionAuditFixture(
	ctx WorkflowHTTPToolActionContext,
	attempt WorkflowHTTPToolExecutionAttempt,
	run WorkflowRunRecord,
	eventID string,
) WorkflowHTTPToolExecutionAudit {
	attemptID, runID, confirmationID := attempt.AttemptID, run.RunID, attempt.ConfirmationID
	audit := WorkflowHTTPToolExecutionAudit{
		SchemaVersion: workflowHTTPToolAuditSchema, EventID: eventID, OccurredAt: attempt.CompletedAt,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		DraftID: run.DraftID, DraftVersion: run.DraftVersion, NodeID: attempt.NodeID, PlanID: run.PlanID,
		ConfirmationID: &confirmationID, ExecutionAttemptID: &attemptID, RunID: &runID,
		ToolID: attempt.ToolID, ToolVersion: workflowHTTPToolVersion, DefinitionDigest: attempt.DefinitionDigest,
		ProfileID: attempt.ProfileID, ProfileDigest: attempt.ProfileDigest, ToolPlanDigest: attempt.ToolPlanDigest,
		ActorRef: ctx.ActorRef, ActorSource: "human", RequestID: ctx.RequestID, AuditRef: "audit_" + eventID,
		HTTPStatusClass: workflowHTTPToolOptionalString(attempt.HTTPStatusClass), ResponseBytes: attempt.ResponseBytes,
		DurationMS:  attempt.DurationMS,
		SideEffects: WorkflowHTTPToolAuditSideEffects{NetworkAttempts: 1, ToolCalls: 1},
	}
	switch attempt.Status {
	case WorkflowHTTPToolAttemptSucceeded:
		audit.EventKind, audit.AttemptStatus = "tool_execution_succeeded", "succeeded"
	case WorkflowHTTPToolAttemptFailed:
		code := string(attempt.FailureCode)
		audit.EventKind, audit.AttemptStatus, audit.FailureCode = "tool_execution_failed", "failed", &code
	case WorkflowHTTPToolAttemptOutcomeUnknown:
		code := string(WorkflowRunFailureToolOutcomeUnknown)
		audit.EventKind, audit.AttemptStatus, audit.FailureCode = "tool_execution_outcome_unknown", "outcome_unknown", &code
	}
	return audit
}
