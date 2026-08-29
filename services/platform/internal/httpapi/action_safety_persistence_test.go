package httpapi

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestActionSafetyStorageSnapshotStrictCodecAndLegacyStatus(t *testing.T) {
	_, _, plan, _, _, _ := newWorkflowDefinitionHTTPToolExecutionServiceForTest(t)
	if plan.ActionSafety == nil {
		t.Fatal("Action Safety plan fixture is missing its projection")
	}
	snapshot, err := encodeActionSafetyPlanSnapshot(plan.ActionSafety)
	if err != nil {
		t.Fatalf("encode Action Safety plan snapshot: %v", err)
	}
	restored, err := decodeActionSafetyPlanSnapshot(snapshot)
	if err != nil || restored == nil || restored.ProjectionDigest != plan.ActionSafety.ProjectionDigest {
		t.Fatalf("round-trip Action Safety plan snapshot: projection=%#v err=%v", restored, err)
	}

	legacy := actionSafetyStorageSnapshot{}
	legacyProjection, err := decodeActionSafetyPlanSnapshot(legacy)
	if err != nil || legacyProjection != nil || actionSafetySnapshotStatus(legacy) != actionSafetySnapshotStatusNotRecordedLegacy {
		t.Fatalf("legacy Action Safety snapshot was not explicit: projection=%#v status=%s err=%v",
			legacyProjection, actionSafetySnapshotStatus(legacy), err)
	}
	if _, err = decodeActionSafetyPlanSnapshot(actionSafetyStorageSnapshot{SchemaVersion: actionSafetyPlanProjectionSchema}); !errors.Is(err, errActionSafetyProjectionContract) {
		t.Fatalf("partial Action Safety snapshot was accepted: %v", err)
	}

	unknownFieldPayload := append([]byte(nil), snapshot.Payload[:len(snapshot.Payload)-1]...)
	unknownFieldPayload = append(unknownFieldPayload, []byte(`,"unexpected_field":true}`)...)
	if _, err = decodeActionSafetyPlanSnapshot(actionSafetyStorageSnapshot{
		SchemaVersion: snapshot.SchemaVersion, ProjectionDigest: snapshot.ProjectionDigest, Payload: unknownFieldPayload,
	}); !errors.Is(err, errActionSafetyProjectionContract) {
		t.Fatalf("unknown Action Safety snapshot field was accepted: %v", err)
	}
	duplicateFieldPayload := append([]byte(nil), snapshot.Payload[:len(snapshot.Payload)-1]...)
	duplicateFieldPayload = append(duplicateFieldPayload, []byte(`,"schema_version":"action_safety_plan_projection.v1"}`)...)
	if _, err = decodeActionSafetyPlanSnapshot(actionSafetyStorageSnapshot{
		SchemaVersion: snapshot.SchemaVersion, ProjectionDigest: snapshot.ProjectionDigest, Payload: duplicateFieldPayload,
	}); !errors.Is(err, errActionSafetyProjectionContract) {
		t.Fatalf("duplicate Action Safety snapshot field was accepted: %v", err)
	}

	corruptedDigest := "sha256:" + strings.Repeat("f", 64)
	if _, err = decodeActionSafetyPlanSnapshot(actionSafetyStorageSnapshot{
		SchemaVersion: snapshot.SchemaVersion, ProjectionDigest: corruptedDigest, Payload: snapshot.Payload,
	}); !errors.Is(err, errActionSafetyProjectionContract) {
		t.Fatalf("Action Safety snapshot digest drift was accepted: %v", err)
	}
}

func TestSQLiteActionSafetyAssignmentSnapshotRestartCASAndCorruption(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "action-safety-assignment.db")
	firstRuntime := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	ctx, assignment, event := actionSafetyAssignmentForPersistenceTest(t)
	repository := newSQLiteAgentCopilotRuntimeRepository(firstRuntime.DB())
	if err := repository.Apply(ctx, 0, assignment, event); err != nil {
		t.Fatalf("persist SQLite Action Safety assignment: %v", err)
	}
	if err := repository.Apply(ctx, 0, assignment, event); !errorsIsAgentCopilotRuntimeVersionConflict(err) {
		t.Fatalf("stale Action Safety assignment bypassed SQLite CAS: %v", err)
	}
	if err := firstRuntime.Close(); err != nil {
		t.Fatalf("close first Action Safety assignment runtime: %v", err)
	}

	secondRuntime := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	t.Cleanup(func() { _ = secondRuntime.Close() })
	restarted := newSQLiteAgentCopilotRuntimeRepository(secondRuntime.DB())
	restored, events, err := restarted.Read(ctx)
	if err != nil || restored.ActionSafety == nil || len(events) != 1 || events[0].ActionSafety == nil ||
		restored.ActionSafety.ProjectionDigest != assignment.ActionSafety.ProjectionDigest ||
		events[0].ActionSafety.ProjectionDigest != event.ActionSafety.ProjectionDigest ||
		restored.ActionSafety.Decision.PolicyVersion != assignment.ActionSafety.Decision.PolicyVersion ||
		restored.ActionSafety.Decision.PolicyDigest != assignment.ActionSafety.Decision.PolicyDigest {
		t.Fatalf("restore SQLite Action Safety assignment: assignment=%#v events=%#v err=%v", restored, events, err)
	}

	corruptedDigest := "sha256:" + strings.Repeat("e", 64)
	// Simulate physical corruption below the normal lifecycle guard. The Action
	// Safety trigger still verifies the three-column storage shape; the Go owner
	// must independently reject a digest that no longer seals the projection.
	if _, err = secondRuntime.DB().ExecContext(ctx.RequestContext, `DROP TRIGGER agent_copilot_assignments_controlled_update`); err != nil {
		t.Fatalf("disable assignment lifecycle guard for corruption injection: %v", err)
	}
	if _, err = secondRuntime.DB().ExecContext(ctx.RequestContext, `UPDATE agent_copilot_runtime_assignments
SET action_safety_projection_digest=?,
    sanitized_action_safety_snapshot=json_set(sanitized_action_safety_snapshot,'$.projection_digest',?)
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=?`,
		corruptedDigest, corruptedDigest, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef,
	); err != nil {
		t.Fatalf("inject SQLite Action Safety assignment corruption: %v", err)
	}
	if _, _, err = restarted.Read(ctx); !errorsIsAgentCopilotRuntime(err, errAgentCopilotRuntimeContract) {
		t.Fatalf("corrupted SQLite Action Safety assignment did not fail closed: %v", err)
	}
}

func TestSQLiteActionSafetyAgentRunSnapshotRestartNoFallbackAndCorruption(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "action-safety-agent-run.db")
	firstRuntime := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	fixture := newAgentCopilotInvocationFixture(t)
	store := &sqlitePromptApplicationRunStore{
		database: firstRuntime.DB(), table: "agent_copilot_run_records", schema: agentCopilotRunV7Schema,
	}
	fixture.service.runStore = store
	input := validAgentCopilotInvocationInput()
	input.ClientInvocationKey = "client-agent-action-safety-sqlite-restart"
	result := fixture.service.Invoke(fixture.ctx, input)
	if result.FailureCode != "" || result.Run == nil || result.Run.ActionSafety == nil {
		t.Fatalf("persist SQLite Action Safety Agent Run: %#v", result)
	}
	runID := result.Run.RunID
	projectionDigest := result.Run.ActionSafety.ProjectionDigest
	policyVersion := result.Run.ActionSafety.Decisions[0].PolicyVersion
	policyDigest := result.Run.ActionSafety.Decisions[0].PolicyDigest
	providerCalls := fixture.bridge.callCount()
	if err := firstRuntime.Close(); err != nil {
		t.Fatalf("close first Action Safety Agent Run runtime: %v", err)
	}

	secondRuntime := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	t.Cleanup(func() { _ = secondRuntime.Close() })
	restarted := &sqlitePromptApplicationRunStore{
		database: secondRuntime.DB(), table: "agent_copilot_run_records", schema: agentCopilotRunV7Schema,
	}
	restored, found, err := restarted.ReadRun(agentCopilotWorkflowRunContext(fixture.ctx), runID)
	if err != nil || !found || restored.ActionSafety == nil || restored.ActionSafety.ProjectionDigest != projectionDigest ||
		restored.ActionSafety.Decisions[0].PolicyVersion != policyVersion ||
		restored.ActionSafety.Decisions[0].PolicyDigest != policyDigest {
		t.Fatalf("restore SQLite Action Safety Agent Run: found=%t run=%#v err=%v", found, restored, err)
	}

	corruptedDigest := "sha256:" + strings.Repeat("d", 64)
	if _, err = secondRuntime.DB().ExecContext(fixture.ctx.RequestContext, `DROP TRIGGER agent_copilot_runs_controlled_update`); err != nil {
		t.Fatalf("disable Agent Run lifecycle guard for corruption injection: %v", err)
	}
	if _, err = secondRuntime.DB().ExecContext(fixture.ctx.RequestContext, `UPDATE agent_copilot_run_records
SET action_safety_projection_digest=?,
    sanitized_action_safety_snapshot=json_set(sanitized_action_safety_snapshot,'$.projection_digest',?)
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND run_id=?`,
		corruptedDigest, corruptedDigest, fixture.ctx.TenantRef, fixture.ctx.WorkspaceID, fixture.ctx.ApplicationID, runID,
	); err != nil {
		t.Fatalf("inject SQLite Action Safety Agent Run corruption: %v", err)
	}
	if _, _, err = restarted.ReadRun(agentCopilotWorkflowRunContext(fixture.ctx), runID); !errors.Is(err, errWorkflowRunStoreContract) {
		t.Fatalf("corrupted SQLite Action Safety Agent Run did not fail closed: %v", err)
	}
	fixture.service.runStore = restarted
	replayed := fixture.service.Invoke(fixture.ctx, input)
	if replayed.FailureCode != AgentCopilotRuntimeFailureStoreUnavailable || fixture.bridge.callCount() != providerCalls {
		t.Fatalf("corrupted Agent Run fell back or crossed provider boundary: result=%#v calls=%d", replayed, fixture.bridge.callCount())
	}
}

func TestSQLiteActionSafetyToolPlanAndWorkflowRunSnapshotsRestartLegacyAndCorruption(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "action-safety-tool-and-run.db")
	firstRuntime := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	_, ctx, approvedPlan, _, _, _ := newWorkflowDefinitionHTTPToolExecutionServiceForTest(t)
	plan := cloneWorkflowHTTPToolActionPlan(approvedPlan)
	plan.RecordVersion = 1
	plan.Status = WorkflowHTTPToolActionStatusPending
	plan.LastDecisionByActorRef = nil
	plan.LastDecisionAt = nil
	planStore := newSQLiteWorkflowHTTPToolActionStore(firstRuntime.DB())
	if err := planStore.CreatePlan(ctx, &plan, workflowHTTPToolAuditForStoreTest(
		plan, "wtae_0000000000000900", "confirmation_requested",
	)); err != nil {
		t.Fatalf("persist SQLite Action Safety Tool plan: %v", err)
	}

	runContext := workflowExecutorTestContext()
	run := workflowRunHistoryTestRecord(runContext, "run_action_safety_sqlite", "draft_action_safety", time.Now().UTC())
	runtime := newActionSafetyRuntimeV1("development")
	runProjection, err := runtime.ProjectRun("terminal", []ActionSafetyDecision{plan.ActionSafety.Decision}, run.SideEffects)
	if err != nil {
		t.Fatalf("project Action Safety Workflow Run: %v", err)
	}
	run.ActionSafety = &runProjection
	runStore := newSQLiteWorkflowRunStore(firstRuntime.DB())
	if err = runStore.UpsertRun(runContext, &run); err != nil {
		t.Fatalf("persist SQLite Action Safety Workflow Run: %v", err)
	}
	legacy := workflowRunHistoryTestRecord(runContext, "run_action_safety_legacy", "draft_action_safety", time.Now().UTC().Add(time.Second))
	if err = runStore.UpsertRun(runContext, &legacy); err != nil {
		t.Fatalf("persist legacy SQLite Workflow Run: %v", err)
	}
	if err = firstRuntime.Close(); err != nil {
		t.Fatalf("close first Action Safety Tool/Run runtime: %v", err)
	}

	secondRuntime := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	t.Cleanup(func() { _ = secondRuntime.Close() })
	planStore = newSQLiteWorkflowHTTPToolActionStore(secondRuntime.DB())
	restoredPlan, found, err := planStore.ReadPlan(ctx, plan.PlanID)
	if err != nil || !found || restoredPlan.ActionSafety == nil ||
		restoredPlan.ActionSafety.ProjectionDigest != plan.ActionSafety.ProjectionDigest ||
		restoredPlan.ActionSafety.Decision.PolicyVersion != plan.ActionSafety.Decision.PolicyVersion ||
		restoredPlan.ActionSafety.Decision.PolicyDigest != plan.ActionSafety.Decision.PolicyDigest {
		t.Fatalf("restore SQLite Action Safety Tool plan: found=%t plan=%#v err=%v", found, restoredPlan, err)
	}
	if _, err = secondRuntime.DB().ExecContext(ctx.RequestContext, `UPDATE workflow_http_tool_action_plans
SET action_safety_projection_digest=NULL WHERE plan_id=?`, plan.PlanID); err == nil {
		t.Fatal("SQLite accepted a partial Action Safety snapshot triplet")
	}
	runStore = newSQLiteWorkflowRunStore(secondRuntime.DB())
	restoredRun, found, err := runStore.ReadRun(runContext, run.RunID)
	if err != nil || !found || restoredRun.ActionSafety == nil ||
		restoredRun.ActionSafety.ProjectionDigest != run.ActionSafety.ProjectionDigest ||
		restoredRun.ActionSafety.Decisions[0].PolicyDigest != run.ActionSafety.Decisions[0].PolicyDigest {
		t.Fatalf("restore SQLite Action Safety Workflow Run: found=%t run=%#v err=%v", found, restoredRun, err)
	}
	restoredLegacy, found, err := runStore.ReadRun(runContext, legacy.RunID)
	if err != nil || !found || restoredLegacy.ActionSafety != nil {
		t.Fatalf("legacy SQLite Workflow Run was not explicit: found=%t run=%#v err=%v", found, restoredLegacy, err)
	}

	corruptedDigest := "sha256:" + strings.Repeat("c", 64)
	if _, err = secondRuntime.DB().ExecContext(ctx.RequestContext, `UPDATE workflow_http_tool_action_plans
SET action_safety_projection_digest=?,
    sanitized_action_safety_snapshot=json_set(sanitized_action_safety_snapshot,'$.projection_digest',?)
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND plan_id=?`,
		corruptedDigest, corruptedDigest, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, plan.PlanID,
	); err != nil {
		t.Fatalf("inject SQLite Action Safety Tool plan corruption: %v", err)
	}
	if _, _, err = planStore.ReadPlan(ctx, plan.PlanID); !errors.Is(err, errWorkflowHTTPToolActionContract) {
		t.Fatalf("corrupted SQLite Action Safety Tool plan did not fail closed: %v", err)
	}
	if _, err = secondRuntime.DB().ExecContext(runContext.RequestContext, `UPDATE workflow_run_records
SET action_safety_projection_digest=?,
    sanitized_action_safety_snapshot=json_set(sanitized_action_safety_snapshot,'$.projection_digest',?)
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND run_id=?`,
		corruptedDigest, corruptedDigest, runContext.TenantRef, runContext.WorkspaceID, runContext.ApplicationID, run.RunID,
	); err != nil {
		t.Fatalf("inject SQLite Action Safety Workflow Run corruption: %v", err)
	}
	if _, _, err = runStore.ReadRun(runContext, run.RunID); !errors.Is(err, errWorkflowRunStoreContract) {
		t.Fatalf("corrupted SQLite Action Safety Workflow Run did not fail closed: %v", err)
	}
}

func actionSafetyAssignmentForPersistenceTest(
	t *testing.T,
) (AgentCopilotRuntimeContext, AgentCopilotRuntimeAssignmentV1, AgentCopilotRuntimeAssignmentEventV1) {
	t.Helper()
	fixture := newAgentCopilotBatchCFixture(t)
	resolved, failure := fixture.runtimeService.resolver.Resolve(
		fixture.runtimeContext, fixture.firstCandidate.CandidateID, nil,
	)
	if failure != "" {
		t.Fatalf("resolve Action Safety assignment authority: %s", failure)
	}
	runtime := newActionSafetyRuntimeV1("development")
	response := AgentCopilotResponse{
		SchemaVersion: 1, Status: "partial", Project: "radishflow", Task: "explain_diagnostics",
		Summary: "Candidate available.", Answers: []AgentCopilotResponseAnswer{}, Issues: []AgentCopilotResponseIssue{},
		ProposedActions: []AgentCopilotResponseAction{{
			Kind: "candidate_edit", Title: "Review edit", Rationale: "Address the diagnostic.",
			RiskLevel: "medium", RequiresConfirmation: true,
		}},
		Citations: []AgentCopilotResponseCitation{}, Confidence: 0.7, RiskLevel: "medium", RequiresConfirmation: true,
	}
	normalized, safetyFailure := runtime.NormalizeAgentCopilotResponse(
		fixture.runtimeContext, "run_dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd", response,
		agentCopilotInvocationAuthority{Resolved: resolved},
	)
	if safetyFailure != "" || len(normalized.Candidates) != 1 {
		t.Fatalf("normalize Action Safety assignment candidate: failure=%s projection=%#v", safetyFailure, normalized)
	}
	candidate, safetyFailure, err := runtime.ReviewCandidate(
		fixture.runtimeContext, normalized.Candidates[0], 1, normalized.Candidates[0].Source,
		resolved.Profile, true, true, true,
	)
	if safetyFailure != "" || err != nil {
		t.Fatalf("review Action Safety assignment candidate: failure=%s err=%v", safetyFailure, err)
	}
	activated := fixture.runtimeService.Decide(fixture.runtimeContext, AgentCopilotRuntimeDecisionInput{
		ExpectedAssignmentVersion: 0, Action: "activate", CandidateID: fixture.firstCandidate.CandidateID,
		ActionSafetyCandidate: &candidate,
	})
	if activated.FailureCode != "" || activated.Assignment == nil || activated.Assignment.ActionSafety == nil ||
		len(activated.Events) != 1 || activated.Events[0].ActionSafety == nil {
		t.Fatalf("activate Action Safety assignment fixture: %#v", activated)
	}
	return fixture.runtimeContext, *activated.Assignment, activated.Events[0]
}
