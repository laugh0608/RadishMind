//go:build postgres_integration

package httpapi

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	workflowrunmigrations "radishmind.local/services/platform/migrations/workflow_runs"
)

func TestPostgresActionSafetyDurableOwnersRestartCorruptionAndMigrationReplay(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeDatabaseURL := postgresIntegrationDatabaseURLForCredentials(
		t, runtimeUser, os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"),
	)
	requestContext, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	adminPool, err := workflowrunmigrations.OpenPool(requestContext, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, requestContext, adminPool)
	resetPostgresWorkflowRunSchema(t, requestContext, adminPool)
	preparePostgresIntegrationRuntimeRole(t, requestContext, adminPool, runtimeUser)
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		resetPostgresWorkflowRunSchema(t, cleanupContext, adminPool)
		adminPool.Close()
	})
	state, err := workflowrunmigrations.Apply(requestContext, adminPool)
	if err != nil || state.MigrationState != workflowrunmigrations.MigrationStateApplied {
		t.Fatalf("apply Action Safety PostgreSQL migration: state=%#v err=%v", state, err)
	}

	runtimePool, err := workflowrunmigrations.OpenPool(requestContext, runtimeDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	assignmentContext, assignment, event := actionSafetyAssignmentForPersistenceTest(t)
	assignmentContext.RequestContext = requestContext
	assignmentStore := newPostgresAgentCopilotRuntimeRepository(runtimePool)
	if err = assignmentStore.Apply(assignmentContext, 0, assignment, event); err != nil {
		t.Fatalf("persist PostgreSQL Action Safety assignment: %v", err)
	}
	if err = assignmentStore.Apply(assignmentContext, 0, assignment, event); !errorsIsAgentCopilotRuntimeVersionConflict(err) {
		t.Fatalf("stale PostgreSQL Action Safety assignment bypassed CAS: %v", err)
	}

	_, planContext, approvedPlan, _, _, _ := newWorkflowDefinitionHTTPToolExecutionServiceForTest(t)
	planContext.RequestContext = requestContext
	plan := cloneWorkflowHTTPToolActionPlan(approvedPlan)
	plan.RecordVersion = 1
	plan.Status = WorkflowHTTPToolActionStatusPending
	plan.LastDecisionByActorRef = nil
	plan.LastDecisionAt = nil
	planStore := newPostgresWorkflowHTTPToolActionStore(runtimePool)
	if err = planStore.CreatePlan(planContext, &plan, workflowHTTPToolAuditForStoreTest(
		plan, "wtae_0000000000000910", "confirmation_requested",
	)); err != nil {
		t.Fatalf("persist PostgreSQL Action Safety Tool plan: %v", err)
	}

	runContext := workflowExecutorTestContext()
	runContext.RequestContext = requestContext
	run := workflowRunHistoryTestRecord(runContext, "run_action_safety_postgres", "draft_action_safety", time.Now().UTC())
	runProjection, err := newActionSafetyRuntimeV1("development").ProjectRun(
		"terminal", []ActionSafetyDecision{plan.ActionSafety.Decision}, run.SideEffects,
	)
	if err != nil {
		t.Fatalf("project PostgreSQL Action Safety Workflow Run: %v", err)
	}
	run.ActionSafety = &runProjection
	runStore := newPostgresWorkflowRunStore(runtimePool)
	if err = runStore.UpsertRun(runContext, &run); err != nil {
		t.Fatalf("persist PostgreSQL Action Safety Workflow Run: %v", err)
	}
	legacy := workflowRunHistoryTestRecord(runContext, "run_action_safety_pg_legacy", "draft_action_safety", time.Now().UTC().Add(time.Second))
	if err = runStore.UpsertRun(runContext, &legacy); err != nil {
		t.Fatalf("persist legacy PostgreSQL Workflow Run: %v", err)
	}

	agentFixture := newAgentCopilotInvocationFixture(t)
	agentFixture.ctx.RequestContext = requestContext
	agentStore := &postgresPromptApplicationRunStore{
		pool: runtimePool, table: "agent_copilot_run_records", schema: agentCopilotRunV7Schema,
	}
	agentFixture.service.runStore = agentStore
	agentInput := validAgentCopilotInvocationInput()
	agentInput.ClientInvocationKey = "client-agent-action-safety-postgres-restart"
	agentResult := agentFixture.service.Invoke(agentFixture.ctx, agentInput)
	if agentResult.FailureCode != "" || agentResult.Run == nil || agentResult.Run.ActionSafety == nil {
		t.Fatalf("persist PostgreSQL Action Safety Agent Run: %#v", agentResult)
	}
	agentProviderCalls := agentFixture.bridge.callCount()
	runtimePool.Close()

	reopened, err := workflowrunmigrations.OpenPool(requestContext, runtimeDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	reopenedClosed := false
	t.Cleanup(func() {
		if !reopenedClosed {
			reopened.Close()
		}
	})
	assignmentStore = newPostgresAgentCopilotRuntimeRepository(reopened)
	restoredAssignment, events, err := assignmentStore.Read(assignmentContext)
	if err != nil || restoredAssignment.ActionSafety == nil || len(events) != 1 || events[0].ActionSafety == nil ||
		restoredAssignment.ActionSafety.ProjectionDigest != assignment.ActionSafety.ProjectionDigest ||
		restoredAssignment.ActionSafety.Decision.PolicyVersion != assignment.ActionSafety.Decision.PolicyVersion ||
		restoredAssignment.ActionSafety.Decision.PolicyDigest != assignment.ActionSafety.Decision.PolicyDigest {
		t.Fatalf("restore PostgreSQL Action Safety assignment: assignment=%#v events=%#v err=%v", restoredAssignment, events, err)
	}
	planStore = newPostgresWorkflowHTTPToolActionStore(reopened)
	restoredPlan, found, err := planStore.ReadPlan(planContext, plan.PlanID)
	if err != nil || !found || restoredPlan.ActionSafety == nil ||
		restoredPlan.ActionSafety.ProjectionDigest != plan.ActionSafety.ProjectionDigest ||
		restoredPlan.ActionSafety.Decision.PolicyVersion != plan.ActionSafety.Decision.PolicyVersion ||
		restoredPlan.ActionSafety.Decision.PolicyDigest != plan.ActionSafety.Decision.PolicyDigest {
		t.Fatalf("restore PostgreSQL Action Safety Tool plan: found=%t plan=%#v err=%v", found, restoredPlan, err)
	}
	runStore = newPostgresWorkflowRunStore(reopened)
	restoredRun, found, err := runStore.ReadRun(runContext, run.RunID)
	if err != nil || !found || restoredRun.ActionSafety == nil ||
		restoredRun.ActionSafety.ProjectionDigest != run.ActionSafety.ProjectionDigest ||
		restoredRun.ActionSafety.Decisions[0].PolicyDigest != run.ActionSafety.Decisions[0].PolicyDigest {
		t.Fatalf("restore PostgreSQL Action Safety Workflow Run: found=%t run=%#v err=%v", found, restoredRun, err)
	}
	restoredLegacy, found, err := runStore.ReadRun(runContext, legacy.RunID)
	if err != nil || !found || restoredLegacy.ActionSafety != nil {
		t.Fatalf("legacy PostgreSQL Workflow Run was not explicit: found=%t run=%#v err=%v", found, restoredLegacy, err)
	}
	agentStore = &postgresPromptApplicationRunStore{
		pool: reopened, table: "agent_copilot_run_records", schema: agentCopilotRunV7Schema,
	}
	restoredAgentRun, found, err := agentStore.ReadRun(agentCopilotWorkflowRunContext(agentFixture.ctx), agentResult.Run.RunID)
	if err != nil || !found || restoredAgentRun.ActionSafety == nil ||
		restoredAgentRun.ActionSafety.ProjectionDigest != agentResult.Run.ActionSafety.ProjectionDigest ||
		restoredAgentRun.ActionSafety.Decisions[0].PolicyVersion != agentResult.Run.ActionSafety.Decisions[0].PolicyVersion ||
		restoredAgentRun.ActionSafety.Decisions[0].PolicyDigest != agentResult.Run.ActionSafety.Decisions[0].PolicyDigest {
		t.Fatalf("restore PostgreSQL Action Safety Agent Run: found=%t run=%#v err=%v", found, restoredAgentRun, err)
	}

	if _, err = reopened.Exec(requestContext, `UPDATE workflow_http_tool_action_plans
SET action_safety_projection_digest=NULL WHERE plan_id=$1`, plan.PlanID); err == nil {
		t.Fatal("PostgreSQL runtime role bypassed the Action Safety three-column constraint")
	}
	corruptedDigest := "sha256:" + strings.Repeat("b", 64)
	if _, err = adminPool.Exec(requestContext, `ALTER TABLE agent_copilot_runtime_assignments
DISABLE TRIGGER agent_copilot_assignments_controlled_update`); err != nil {
		t.Fatalf("disable PostgreSQL assignment lifecycle guard for corruption injection: %v", err)
	}
	if _, err = adminPool.Exec(requestContext, `UPDATE agent_copilot_runtime_assignments
SET action_safety_projection_digest=$1,
    sanitized_action_safety_snapshot=jsonb_set(sanitized_action_safety_snapshot,'{projection_digest}',to_jsonb($1::text),false)
WHERE assignment_id=$2`, corruptedDigest, assignment.AssignmentID); err != nil {
		t.Fatalf("inject PostgreSQL Action Safety assignment corruption: %v", err)
	}
	if _, err = adminPool.Exec(requestContext, `ALTER TABLE agent_copilot_runtime_assignments
ENABLE TRIGGER agent_copilot_assignments_controlled_update`); err != nil {
		t.Fatalf("restore PostgreSQL assignment lifecycle guard: %v", err)
	}
	if _, _, err = assignmentStore.Read(assignmentContext); !errorsIsAgentCopilotRuntime(err, errAgentCopilotRuntimeContract) {
		t.Fatalf("corrupted PostgreSQL Action Safety assignment did not fail closed: %v", err)
	}

	if _, err = adminPool.Exec(requestContext, `UPDATE workflow_http_tool_action_plans
SET action_safety_projection_digest=$1,
    sanitized_action_safety_snapshot=jsonb_set(sanitized_action_safety_snapshot,'{projection_digest}',to_jsonb($1::text),false)
WHERE plan_id=$2`, corruptedDigest, plan.PlanID); err != nil {
		t.Fatalf("inject PostgreSQL Action Safety Tool plan corruption: %v", err)
	}
	if _, _, err = planStore.ReadPlan(planContext, plan.PlanID); !errors.Is(err, errWorkflowHTTPToolActionContract) {
		t.Fatalf("corrupted PostgreSQL Action Safety Tool plan did not fail closed: %v", err)
	}
	if _, err = adminPool.Exec(requestContext, `UPDATE workflow_run_records
SET action_safety_projection_digest=$1,
    sanitized_action_safety_snapshot=jsonb_set(sanitized_action_safety_snapshot,'{projection_digest}',to_jsonb($1::text),false)
WHERE run_id=$2`, corruptedDigest, run.RunID); err != nil {
		t.Fatalf("inject PostgreSQL Action Safety Workflow Run corruption: %v", err)
	}
	if _, _, err = runStore.ReadRun(runContext, run.RunID); !errors.Is(err, errWorkflowRunStoreContract) {
		t.Fatalf("corrupted PostgreSQL Action Safety Workflow Run did not fail closed: %v", err)
	}

	if _, err = adminPool.Exec(requestContext, `ALTER TABLE agent_copilot_run_records
DISABLE TRIGGER agent_copilot_runs_controlled_update`); err != nil {
		t.Fatalf("disable PostgreSQL Agent Run lifecycle guard for corruption injection: %v", err)
	}
	if _, err = adminPool.Exec(requestContext, `UPDATE agent_copilot_run_records
SET action_safety_projection_digest=$1,
    sanitized_action_safety_snapshot=jsonb_set(sanitized_action_safety_snapshot,'{projection_digest}',to_jsonb($1::text),false)
WHERE run_id=$2`, corruptedDigest, agentResult.Run.RunID); err != nil {
		t.Fatalf("inject PostgreSQL Action Safety Agent Run corruption: %v", err)
	}
	if _, err = adminPool.Exec(requestContext, `ALTER TABLE agent_copilot_run_records
ENABLE TRIGGER agent_copilot_runs_controlled_update`); err != nil {
		t.Fatalf("restore PostgreSQL Agent Run lifecycle guard: %v", err)
	}
	if _, _, err = agentStore.ReadRun(agentCopilotWorkflowRunContext(agentFixture.ctx), agentResult.Run.RunID); !errors.Is(err, errWorkflowRunStoreContract) {
		t.Fatalf("corrupted PostgreSQL Action Safety Agent Run did not fail closed: %v", err)
	}
	agentFixture.service.runStore = agentStore
	replayed := agentFixture.service.Invoke(agentFixture.ctx, agentInput)
	if replayed.FailureCode != AgentCopilotRuntimeFailureStoreUnavailable || agentFixture.bridge.callCount() != agentProviderCalls {
		t.Fatalf("corrupted PostgreSQL Agent Run fell back or crossed provider boundary: result=%#v calls=%d",
			replayed, agentFixture.bridge.callCount())
	}

	reopened.Close()
	reopenedClosed = true
	if _, err = workflowrunmigrations.RollbackForDevTest(requestContext, adminPool); err != nil {
		t.Fatalf("rollback Action Safety PostgreSQL migration family: %v", err)
	}
	pending, err := workflowrunmigrations.Inspect(requestContext, adminPool)
	if err != nil || pending.MigrationState != workflowrunmigrations.MigrationStateNotApplied {
		t.Fatalf("inspect rolled-back Action Safety PostgreSQL migration: state=%#v err=%v", pending, err)
	}
	reapplied, err := workflowrunmigrations.Apply(requestContext, adminPool)
	if err != nil || reapplied.MigrationState != workflowrunmigrations.MigrationStateApplied ||
		reapplied.MigrationChecksum != workflowrunmigrations.ExpectedChecksum() {
		t.Fatalf("reapply Action Safety PostgreSQL migration: state=%#v err=%v", reapplied, err)
	}
	var actionSafetyColumnCount int
	if err = adminPool.QueryRow(requestContext, `SELECT count(*) FROM information_schema.columns
WHERE table_schema='public'
  AND table_name IN ('agent_copilot_runtime_assignments','agent_copilot_runtime_assignment_events',
                     'agent_copilot_run_records','workflow_http_tool_action_plans','workflow_run_records')
  AND column_name IN ('action_safety_schema_version','action_safety_projection_digest','sanitized_action_safety_snapshot')`).Scan(
		&actionSafetyColumnCount,
	); err != nil || actionSafetyColumnCount != 15 {
		t.Fatalf("Action Safety PostgreSQL reapply did not restore all owner columns: count=%d err=%v",
			actionSafetyColumnCount, err)
	}
}
