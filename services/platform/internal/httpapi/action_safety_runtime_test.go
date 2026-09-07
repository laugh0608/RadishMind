package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/bridge"
)

func TestActionSafetyRuntimeNormalizesResponseAndProjectsCandidate(t *testing.T) {
	fixture := newAgentCopilotInvocationFixture(t)
	authority, authorityFailure := fixture.service.resolveAuthority(fixture.ctx)
	if authorityFailure != "" {
		t.Fatalf("resolve Agent Copilot authority: %s", authorityFailure)
	}
	runtime := newActionSafetyRuntimeV1("development")
	runtime.now = func() time.Time { return time.Date(2026, 8, 29, 10, 0, 0, 0, time.UTC) }
	response := AgentCopilotResponse{
		SchemaVersion: 1, Status: "partial", Project: "radishflow", Task: "explain_diagnostics",
		Summary: "Candidate available.", ProposedActions: []AgentCopilotResponseAction{{
			Kind: "candidate_edit", Title: "Review edit", Rationale: "Address the diagnostic.",
			RiskLevel: "medium", RequiresConfirmation: true,
		}},
		Answers: []AgentCopilotResponseAnswer{}, Issues: []AgentCopilotResponseIssue{},
		Citations: []AgentCopilotResponseCitation{}, Confidence: 0.7, RiskLevel: "medium", RequiresConfirmation: true,
	}
	projection, failure := runtime.NormalizeAgentCopilotResponse(
		fixture.ctx, "run_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", response, authority,
	)
	if failure != "" || validateActionSafetyResponseProjection(projection) != nil || len(projection.Candidates) != 1 ||
		projection.Candidates[0].Decision.EffectiveLevel != ActionSafetyLevelProposalOnly {
		t.Fatalf("normalize response: failure=%s projection=%#v", failure, projection)
	}
	runProjection, err := runtime.ProjectRun(
		"response_normalization", actionSafetyResponseDecisions(projection), WorkflowRunSideEffects{ProviderCalls: 1},
	)
	if err != nil || validateActionSafetyRunProjection(runProjection) != nil {
		t.Fatalf("project response Run: err=%v projection=%#v", err, runProjection)
	}
}

func TestActionSafetyRuntimeReviewAndAssignmentRevalidateAuthority(t *testing.T) {
	fixture := newAgentCopilotInvocationFixture(t)
	authority, authorityFailure := fixture.service.resolveAuthority(fixture.ctx)
	if authorityFailure != "" {
		t.Fatalf("resolve Agent Copilot authority: %s", authorityFailure)
	}
	runtime := newActionSafetyRuntimeV1("development")
	runtime.now = func() time.Time { return time.Date(2026, 8, 29, 10, 30, 0, 0, time.UTC) }
	response := AgentCopilotResponse{
		SchemaVersion: 1, Status: "partial", Project: "radishflow", Task: "explain_diagnostics",
		Summary: "Candidate available.", ProposedActions: []AgentCopilotResponseAction{{
			Kind: "candidate_edit", Title: "Review edit", Rationale: "Address the diagnostic.",
			RiskLevel: "medium", RequiresConfirmation: true,
		}},
		Answers: []AgentCopilotResponseAnswer{}, Issues: []AgentCopilotResponseIssue{},
		Citations: []AgentCopilotResponseCitation{}, Confidence: 0.7, RiskLevel: "medium", RequiresConfirmation: true,
	}
	normalized, failure := runtime.NormalizeAgentCopilotResponse(
		fixture.ctx, "run_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", response, authority,
	)
	if failure != "" || len(normalized.Candidates) != 1 {
		t.Fatalf("normalize candidate: failure=%s projection=%#v", failure, normalized)
	}
	candidate := normalized.Candidates[0]
	var err error
	driftedSource := candidate.Source
	driftedSource.Digest = workflowRAGSHA256("changed candidate source")
	if _, failure, err = runtime.ReviewCandidate(
		fixture.ctx, candidate, candidate.RecordVersion, driftedSource, authority.Resolved.Profile, true, true, true,
	); failure != ActionSafetyFailureSourceChanged || err != nil {
		t.Fatalf("source drift did not fail closed: failure=%s err=%v", failure, err)
	}
	if _, failure, err = runtime.ReviewCandidate(
		fixture.ctx, candidate, candidate.RecordVersion, candidate.Source, authority.Resolved.Profile, true, false, true,
	); failure != ActionSafetyFailureScopeDenied || err != nil {
		t.Fatalf("membership denial did not fail closed: failure=%s err=%v", failure, err)
	}
	if _, failure, err = runtime.ReviewCandidate(
		fixture.ctx, candidate, candidate.RecordVersion, candidate.Source, authority.Resolved.Profile, true, true, false,
	); failure != ActionSafetyFailureLevelEscalationDenied || err != nil {
		t.Fatalf("human approval escalated without an accepted target: failure=%s err=%v", failure, err)
	}
	reviewed, failure, err := runtime.ReviewCandidate(
		fixture.ctx, candidate, candidate.RecordVersion, candidate.Source, authority.Resolved.Profile, true, true, true,
	)
	if failure != "" || err != nil || reviewed.RecordVersion != 2 || reviewed.ReviewState != actionSafetyCandidateApproved ||
		reviewed.Decision.EffectiveLevel != ActionSafetyLevelHandoffReady {
		t.Fatalf("review candidate: failure=%s err=%v projection=%#v", failure, err, reviewed)
	}
	if _, _, err = runtime.ReviewCandidate(
		fixture.ctx, reviewed, 1, reviewed.Source, authority.Resolved.Profile, true, true, true,
	); !errors.Is(err, errActionSafetyOwnerConflict) {
		t.Fatalf("stale candidate review did not fail CAS: %v", err)
	}
	assignment, failure := runtime.ActivateCandidate(fixture.ctx, reviewed, authority.Resolved.Profile, 1, true)
	if failure != "" || validateActionSafetyAssignmentProjection(assignment) != nil ||
		assignment.Decision.EffectiveLevel != ActionSafetyLevelHandoffReady {
		t.Fatalf("activate reviewed candidate: failure=%s projection=%#v", failure, assignment)
	}

	driftedProfile := authority.Resolved.Profile
	driftedProfile.PolicyDigest = workflowRAGSHA256("changed action safety owner policy")
	if _, failure = runtime.ActivateCandidate(fixture.ctx, reviewed, driftedProfile, 1, true); failure != ActionSafetyFailurePolicyChanged {
		t.Fatalf("policy drift did not fail closed: %s", failure)
	}
}

func TestActionSafetyAgentInvocationProjectsWithoutExposingHTTPContract(t *testing.T) {
	t.Run("answer", func(t *testing.T) {
		fixture := newAgentCopilotInvocationFixture(t)
		result := fixture.service.Invoke(fixture.ctx, validAgentCopilotInvocationInput())
		if result.FailureCode != "" || result.ActionSafety == nil || result.ActionSafety.AnswerDecision == nil ||
			result.ActionSafety.AnswerDecision.EffectiveLevel != ActionSafetyLevelAnswerOnly ||
			result.Run == nil || result.Run.ActionSafety == nil || validateActionSafetyRunProjection(*result.Run.ActionSafety) != nil ||
			result.Run.SideEffects.ProviderCalls != 1 || result.Run.SideEffects.ToolCalls != 0 ||
			result.Run.SideEffects.BusinessWrites != 0 || result.Run.SideEffects.ReplayWrites != 0 {
			t.Fatalf("answer response did not retain a safe internal projection: %#v", result)
		}
		stored, found, err := fixture.runStore.ReadRun(agentCopilotWorkflowRunContext(fixture.ctx), result.Run.RunID)
		if err != nil || !found || stored.ActionSafety == nil || stored.ActionSafety.ProjectionDigest != result.Run.ActionSafety.ProjectionDigest {
			t.Fatalf("answer Run projection was not retained by the memory owner: found=%t err=%v run=%#v", found, err, stored)
		}
		responsePayload, _ := json.Marshal(result.Response)
		runPayload, _ := json.Marshal(result.Run)
		resultPayload, _ := json.Marshal(result)
		if stringContainsActionSafetyProjection(responsePayload) || stringContainsActionSafetyProjection(runPayload) ||
			stringContainsActionSafetyProjection(resultPayload) {
			t.Fatalf("Batch B leaked internal Action Safety projections: response=%s run=%s result=%s", responsePayload, runPayload, resultPayload)
		}
		replayed := fixture.service.Invoke(fixture.ctx, validAgentCopilotInvocationInput())
		if !replayed.IdempotentReplay || replayed.Response != nil || replayed.Run == nil || replayed.Run.ActionSafety == nil || fixture.bridge.callCount() != 1 {
			t.Fatalf("terminal replay crossed the provider boundary or lost its snapshot: %#v calls=%d", replayed, fixture.bridge.callCount())
		}
	})

	t.Run("candidate and blocked write", func(t *testing.T) {
		fixture := newAgentCopilotInvocationFixture(t)
		fixture.bridge.handle = func(context.Context, []byte, bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
			return agentCopilotGatewayEnvelope(map[string]any{
				"schema_version": 1, "status": "partial", "project": "radishflow", "task": "explain_diagnostics",
				"summary": "Unsafe automatic write requested.", "answers": []any{}, "issues": []any{},
				"proposed_actions": []any{map[string]any{
					"kind": "candidate_edit", "title": "Apply immediately", "rationale": "Unsafe automatic path.",
					"apply": map[string]any{"automatic": true}, "risk_level": "medium", "requires_confirmation": true,
				}},
				"citations": []any{}, "confidence": 0.5, "risk_level": "medium", "requires_confirmation": true,
			}), nil
		}
		input := validAgentCopilotInvocationInput()
		input.ClientInvocationKey = "client-agent-action-safety-write-blocked"
		result := fixture.service.Invoke(fixture.ctx, input)
		if result.FailureCode != "" || result.ActionSafety == nil || len(result.ActionSafety.Candidates) != 1 ||
			result.ActionSafety.Candidates[0].Decision.EffectiveLevel != ActionSafetyLevelWriteBlocked ||
			!actionSafetyDecisionHasBlocker(result.ActionSafety.Candidates[0].Decision, ActionSafetyFailureWriteBlocked) ||
			result.Run == nil || result.Run.ActionSafety == nil || result.Run.SideEffects.ProviderCalls != 1 ||
			result.Run.SideEffects.ToolCalls != 0 || result.Run.SideEffects.BusinessWrites != 0 || result.Run.SideEffects.ReplayWrites != 0 ||
			fixture.bridge.callCount() != 1 {
			t.Fatalf("automatic write was not blocked without secondary effects: %#v calls=%d", result, fixture.bridge.callCount())
		}
		authority, authorityFailure := fixture.service.resolveAuthority(fixture.ctx)
		if authorityFailure != "" {
			t.Fatalf("resolve current authority: %s", authorityFailure)
		}
		candidate := result.ActionSafety.Candidates[0]
		if _, failure, err := fixture.service.actionSafety.ReviewCandidate(
			fixture.ctx, candidate, candidate.RecordVersion, candidate.Source, authority.Resolved.Profile, true, true, true,
		); failure != ActionSafetyFailureWriteBlocked || err != nil {
			t.Fatalf("human approval bypassed the write blocker: failure=%s err=%v", failure, err)
		}
	})

	t.Run("late response", func(t *testing.T) {
		fixture := newAgentCopilotInvocationFixture(t)
		requestContext, cancel := context.WithCancel(context.Background())
		fixture.ctx.RequestContext = requestContext
		fixture.bridge.handle = func(context.Context, []byte, bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
			cancel()
			return successfulAgentCopilotGatewayEnvelope(), nil
		}
		input := validAgentCopilotInvocationInput()
		input.ClientInvocationKey = "client-agent-action-safety-late-response"
		result := fixture.service.Invoke(fixture.ctx, input)
		if result.FailureCode != AgentCopilotInvocationFailureCanceled || result.ActionSafety != nil ||
			result.Run == nil || result.Run.ActionSafety != nil || result.Run.Status != WorkflowRunStatusCanceled ||
			result.Run.SideEffects.ProviderCalls != 1 || fixture.bridge.callCount() != 1 {
			t.Fatalf("late provider response was accepted after cancellation: %#v calls=%d", result, fixture.bridge.callCount())
		}
		replayed := fixture.service.Invoke(fixture.ctx, input)
		if !replayed.IdempotentReplay || replayed.FailureCode != AgentCopilotInvocationFailureCanceled || fixture.bridge.callCount() != 1 {
			t.Fatalf("late response replayed the provider: %#v calls=%d", replayed, fixture.bridge.callCount())
		}
	})

	t.Run("authority drift after provider", func(t *testing.T) {
		fixture := newAgentCopilotInvocationFixture(t)
		fixture.bridge.handle = func(context.Context, []byte, bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
			assignment, events, err := fixture.runtime.Read(fixture.ctx)
			if err != nil {
				t.Fatalf("read assignment before drift: %v", err)
			}
			revokedAt := "2026-07-25T16:00:00Z"
			assignment.State, assignment.RevokedAt = "revoked", &revokedAt
			assignment.AssignmentVersion++
			assignment.UpdatedAt = revokedAt
			assignment.ActionSafety = nil
			assignment.AssignmentDigest, _ = agentCopilotRuntimeAssignmentDigest(assignment)
			event := events[len(events)-1]
			event.EventID, event.EventSequence = "acrae_cccccccccccccccc", len(events)+1
			event.Action, event.ExpectedAssignmentVersion = "revoke", assignment.AssignmentVersion-1
			event.ResultingAssignmentVersion, event.AssignmentDigest = assignment.AssignmentVersion, assignment.AssignmentDigest
			event.OccurredAt, event.ActionSafety = revokedAt, nil
			if err = fixture.runtime.Apply(fixture.ctx, event.ExpectedAssignmentVersion, assignment, event); err != nil {
				t.Fatalf("revoke assignment during provider call: %v", err)
			}
			return successfulAgentCopilotGatewayEnvelope(), nil
		}
		input := validAgentCopilotInvocationInput()
		input.ClientInvocationKey = "client-agent-action-safety-authority-drift"
		result := fixture.service.Invoke(fixture.ctx, input)
		if result.FailureCode != AgentCopilotRuntimeFailureAuthorityChanged || result.ActionSafety != nil ||
			result.Run == nil || result.Run.ActionSafety != nil || result.Run.Status != WorkflowRunStatusFailed ||
			result.Run.SideEffects.ProviderCalls != 1 || fixture.bridge.callCount() != 1 {
			t.Fatalf("post-provider authority drift reached response normalization: %#v calls=%d", result, fixture.bridge.callCount())
		}
	})
}

func TestActionSafetyAssignmentUsesExistingMemoryCASOwner(t *testing.T) {
	fixture := newAgentCopilotBatchCFixture(t)
	resolved, resolveFailure := fixture.runtimeService.resolver.Resolve(fixture.runtimeContext, fixture.firstCandidate.CandidateID, nil)
	if resolveFailure != "" {
		t.Fatalf("resolve assignment authority: %s", resolveFailure)
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
	normalized, failure := runtime.NormalizeAgentCopilotResponse(
		fixture.runtimeContext, "run_cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc", response,
		agentCopilotInvocationAuthority{Resolved: resolved},
	)
	if failure != "" || len(normalized.Candidates) != 1 {
		t.Fatalf("normalize assignment candidate: failure=%s projection=%#v", failure, normalized)
	}
	candidate, failure, err := runtime.ReviewCandidate(
		fixture.runtimeContext, normalized.Candidates[0], 1, normalized.Candidates[0].Source, resolved.Profile, true, true, true,
	)
	if failure != "" || err != nil {
		t.Fatalf("review assignment candidate: failure=%s err=%v", failure, err)
	}
	activated := fixture.runtimeService.Decide(fixture.runtimeContext, AgentCopilotRuntimeDecisionInput{
		ExpectedAssignmentVersion: 0, Action: "activate", CandidateID: fixture.firstCandidate.CandidateID,
		ActionSafetyCandidate: &candidate,
	})
	if activated.FailureCode != "" || activated.Assignment == nil || activated.Assignment.ActionSafety == nil ||
		activated.Assignment.ActionSafety.AssignmentVersion != 1 || len(activated.Events) != 1 ||
		activated.Events[0].ActionSafety == nil ||
		activated.Events[0].ActionSafety.ProjectionDigest != activated.Assignment.ActionSafety.ProjectionDigest {
		t.Fatalf("existing assignment owner did not atomically retain Action Safety: %#v", activated)
	}
	stale := fixture.runtimeService.Decide(fixture.runtimeContext, AgentCopilotRuntimeDecisionInput{
		ExpectedAssignmentVersion: 0, Action: "replace", CandidateID: fixture.firstCandidate.CandidateID,
		ActionSafetyCandidate: &candidate,
	})
	if stale.FailureCode != AgentCopilotRuntimeFailureVersionConflict {
		t.Fatalf("stale assignment Action Safety mutation bypassed CAS: %#v", stale)
	}
	stored, events, err := fixture.runtimeRepository.Read(fixture.runtimeContext)
	if err != nil || stored.ActionSafety == nil || stored.AssignmentVersion != 1 || len(events) != 1 {
		t.Fatalf("failed CAS changed the assignment owner: err=%v assignment=%#v events=%#v", err, stored, events)
	}
}

func TestActionSafetyHTTPToolPlanPreDispatchAndRunProjection(t *testing.T) {
	service, ctx, plan, runStore, bridgeClient, networkAttempts := newWorkflowDefinitionHTTPToolExecutionServiceForTest(t)
	if plan.ActionSafety == nil || validateActionSafetyPlanProjection(*plan.ActionSafety) != nil ||
		plan.ActionSafety.Decision.RequestedLevel != ActionSafetyLevelToolCallable ||
		plan.ActionSafety.Decision.EffectiveLevel != ActionSafetyLevelProposalOnly ||
		!actionSafetyDecisionHasBlocker(plan.ActionSafety.Decision, ActionSafetyFailureConfirmationRequired) {
		t.Fatalf("plan creation did not persist an immutable pre-approval snapshot: %#v", plan.ActionSafety)
	}
	corruptedPlan := cloneWorkflowHTTPToolActionPlan(plan)
	corruptedPlan.ActionSafety.Decision.DecisionDigest = workflowRAGSHA256("corrupted plan safety decision")
	if workflowHTTPToolPlanMatchesContext(corruptedPlan, ctx) {
		t.Fatal("plan owner accepted a corrupted Action Safety projection")
	}
	confirmation, found, err := service.store.ReadApprovedConfirmation(ctx, plan.PlanID, plan.RecordVersion)
	if err != nil || !found {
		t.Fatalf("read approved confirmation: found=%t err=%v", found, err)
	}
	decision, safetyFailure := service.actions.actionSafety.RevalidateToolPreDispatch(
		ctx, plan, &confirmation, service.actions.registry, true, true, true,
	)
	if safetyFailure != "" || decision.EffectiveLevel != ActionSafetyLevelToolCallable ||
		decision.SideEffectBudget.ToolNetworkCalls != 1 || decision.SideEffectBudget.ConfirmationConsumptions != 1 {
		t.Fatalf("approved exact plan was not tool-callable: failure=%s decision=%#v", safetyFailure, decision)
	}
	withoutConfirmation, safetyFailure := service.actions.actionSafety.RevalidateToolPreDispatch(
		ctx, plan, nil, service.actions.registry, true, true, true,
	)
	if safetyFailure != ActionSafetyFailureConfirmationRequired || withoutConfirmation.EffectiveLevel == ActionSafetyLevelToolCallable ||
		!actionSafetyDecisionHasBlocker(withoutConfirmation, ActionSafetyFailureConfirmationRequired) {
		t.Fatalf("missing confirmation did not fail closed: failure=%s decision=%#v", safetyFailure, withoutConfirmation)
	}
	changedConfirmation := confirmation
	changedConfirmation.ToolPlanDigest = workflowRAGSHA256("changed confirmation plan")
	changedDecision, safetyFailure := service.actions.actionSafety.RevalidateToolPreDispatch(
		ctx, plan, &changedConfirmation, service.actions.registry, true, true, true,
	)
	if safetyFailure != ActionSafetyFailureConfirmationChanged ||
		!actionSafetyDecisionHasBlocker(changedDecision, ActionSafetyFailureConfirmationChanged) {
		t.Fatalf("confirmation drift did not fail closed: failure=%s decision=%#v", safetyFailure, changedDecision)
	}
	if oversized, projectionErr := service.actions.actionSafety.ProjectRun(
		"pre_dispatch", []ActionSafetyDecision{decision}, WorkflowRunSideEffects{ProviderCalls: 2, ToolCalls: 1, ConfirmationCalls: 1},
	); projectionErr == nil {
		t.Fatalf("Run projection accepted more provider calls than the tool chain allows: %#v", oversized)
	}

	result := service.Execute(ctx, WorkflowHTTPToolExecutionRequest{
		PlanID: plan.PlanID, ExpectedRecordVersion: plan.RecordVersion, InputText: "bounded input",
	})
	if result.FailureCode != "" || result.ActionSafety == nil || result.ActionSafety.EffectiveLevel != ActionSafetyLevelToolCallable ||
		result.Record == nil || result.Record.ActionSafety == nil || result.Record.ActionSafety.Checkpoint != "terminal" ||
		validateActionSafetyRunProjection(*result.Record.ActionSafety) != nil || *networkAttempts != 1 || bridgeClient.callCount() != 1 ||
		result.Record.SideEffects.ToolCalls != 1 || result.Record.SideEffects.ConfirmationCalls != 1 ||
		result.Record.SideEffects.BusinessWrites != 0 || result.Record.SideEffects.ReplayWrites != 0 {
		t.Fatalf("tool execution did not retain the exact terminal safety projection: result=%#v network=%d provider=%d", result, *networkAttempts, bridgeClient.callCount())
	}
	stored, storedFound, err := runStore.ReadRun(workflowRunContextFromToolAction(ctx), result.Record.RunID)
	if err != nil || !storedFound || stored.ActionSafety == nil || stored.ActionSafety.ProjectionDigest != result.Record.ActionSafety.ProjectionDigest {
		t.Fatalf("terminal Run owner lost the Action Safety projection: found=%t err=%v run=%#v", storedFound, err, stored)
	}
	mismatched := cloneWorkflowRunRecord(stored)
	mismatched.SideEffects.ProviderCalls++
	if validateWorkflowRunStoreRecord(workflowRunContextFromToolAction(ctx), &mismatched) == nil {
		t.Fatal("Run owner accepted side-effect counters that drifted from the Action Safety projection")
	}
	repeated := service.Execute(ctx, WorkflowHTTPToolExecutionRequest{
		PlanID: plan.PlanID, ExpectedRecordVersion: plan.RecordVersion, InputText: "do not execute twice",
	})
	if repeated.FailureCode != WorkflowRunFailureToolConfirmation || *networkAttempts != 1 || bridgeClient.callCount() != 1 || len(runStore.records) != 1 {
		t.Fatalf("duplicate execution crossed a side-effect boundary: result=%#v network=%d provider=%d runs=%d", repeated, *networkAttempts, bridgeClient.callCount(), len(runStore.records))
	}
}

func TestActionSafetyHTTPToolDriftFailsBeforeClaim(t *testing.T) {
	t.Run("policy", func(t *testing.T) {
		service, ctx, plan, runStore, bridgeClient, networkAttempts := newWorkflowDefinitionHTTPToolExecutionServiceForTest(t)
		currentPolicy := service.actions.actionSafety.policyRef
		service.actions.actionSafety.policyRef = func(owner string) (ActionSafetyPolicyReference, bool) {
			ref, available := currentPolicy(owner)
			ref.Digest = workflowRAGSHA256("drifted policy " + owner)
			return ref, available
		}
		result := service.Execute(ctx, WorkflowHTTPToolExecutionRequest{
			PlanID: plan.PlanID, ExpectedRecordVersion: plan.RecordVersion, InputText: "bounded input",
		})
		if result.FailureCode != WorkflowRunFailureToolPolicy || result.ActionSafety == nil ||
			!actionSafetyDecisionHasBlocker(*result.ActionSafety, ActionSafetyFailurePolicyChanged) ||
			*networkAttempts != 0 || bridgeClient.callCount() != 0 || len(runStore.records) != 0 {
			t.Fatalf("policy drift crossed the claim boundary: result=%#v network=%d provider=%d runs=%d", result, *networkAttempts, bridgeClient.callCount(), len(runStore.records))
		}
	})

	t.Run("source and membership", func(t *testing.T) {
		service, ctx, plan, _, _, _ := newWorkflowDefinitionHTTPToolExecutionServiceForTest(t)
		confirmation, found, err := service.store.ReadApprovedConfirmation(ctx, plan.PlanID, plan.RecordVersion)
		if err != nil || !found {
			t.Fatalf("read approved confirmation: found=%t err=%v", found, err)
		}
		drifted := cloneWorkflowHTTPToolActionPlan(plan)
		drifted.ToolPlanDigest = workflowRAGSHA256("drifted tool plan source")
		decision, failure := service.actions.actionSafety.RevalidateToolPreDispatch(
			ctx, drifted, &confirmation, service.actions.registry, true, true, true,
		)
		if failure != ActionSafetyFailureSourceChanged || !actionSafetyDecisionHasBlocker(decision, ActionSafetyFailureSourceChanged) {
			t.Fatalf("source drift did not fail closed: failure=%s decision=%#v", failure, decision)
		}
		decision, failure = service.actions.actionSafety.RevalidateToolPreDispatch(
			ctx, plan, &confirmation, service.actions.registry, true, false, true,
		)
		if failure != ActionSafetyFailureScopeDenied || !actionSafetyDecisionHasBlocker(decision, ActionSafetyFailureScopeDenied) {
			t.Fatalf("membership denial did not fail closed: failure=%s decision=%#v", failure, decision)
		}
	})
}

func actionSafetyDecisionHasBlocker(decision ActionSafetyDecision, blocker ActionSafetyFailureCode) bool {
	for _, current := range decision.Blockers {
		if current == blocker {
			return true
		}
	}
	return false
}

func stringContainsActionSafetyProjection(payload []byte) bool {
	return json.Valid(payload) && bytes.Contains(payload, []byte("action_safety"))
}
