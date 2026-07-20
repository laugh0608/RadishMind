package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/bridge"
)

func TestWorkflowHTTPToolExecutionConsumesApprovalOnceAndCompletesRunV2(t *testing.T) {
	service, ctx, plan, runStore, testBridge, networkAttempts := newWorkflowHTTPToolExecutionServiceForTest(t, nil)
	result := service.Execute(ctx, WorkflowHTTPToolExecutionRequest{
		PlanID: plan.PlanID, ExpectedRecordVersion: plan.RecordVersion,
		InputText: "Review the approved resource.", Model: "mock",
	})
	if result.FailureCode != "" || result.Record == nil || result.ActionPlan == nil {
		t.Fatalf("execute approved workflow HTTP tool: %#v", result)
	}
	if result.ActionPlan.Status != WorkflowHTTPToolActionStatusConsumed || result.ActionPlan.RecordVersion != plan.RecordVersion+1 {
		t.Fatalf("approval was not consumed exactly once: %#v", result.ActionPlan)
	}
	if result.Record.SchemaVersion != workflowRunRecordToolSchemaVersion || result.Record.Status != WorkflowRunStatusSucceeded ||
		result.Record.ToolAttempt == nil || result.Record.ToolAttempt.Status != WorkflowHTTPToolAttemptSucceeded ||
		result.Record.SideEffects.ToolCalls != 1 || result.Record.SideEffects.ConfirmationCalls != 1 ||
		result.Record.SideEffects.ProviderCalls != 1 || result.Record.SideEffects.BusinessWrites != 0 ||
		result.Record.SideEffects.ReplayWrites != 0 || *networkAttempts != 1 || testBridge.callCount() != 1 {
		t.Fatalf("unexpected successful run v2: %#v network=%d bridge=%d", result.Record, *networkAttempts, testBridge.callCount())
	}
	if result.Record.ToolAttempt.OutputProjection["resource_key"] != "docs/radishflow/overview" ||
		!strings.Contains(result.Record.Output, "reviewable workflow answer") {
		t.Fatalf("reviewed projection did not reach the model/output chain: %#v", result.Record)
	}
	stored, found, err := runStore.ReadRun(workflowRunContextFromToolAction(ctx), result.Record.RunID)
	if err != nil || !found || stored.Status != WorkflowRunStatusSucceeded || stored.RecordVersion != 2 {
		t.Fatalf("terminal run v2 was not durably stored: %#v found=%t err=%v", stored, found, err)
	}
	encoded := fmt.Sprintf("%#v", stored)
	for _, forbidden := range []string{"api.dev.example.invalid", "Authorization", "raw_response", "8.8.8.8"} {
		if strings.Contains(encoded, forbidden) {
			t.Fatalf("stored run exposed forbidden transport detail %q: %s", forbidden, encoded)
		}
	}

	repeated := service.Execute(ctx, WorkflowHTTPToolExecutionRequest{
		PlanID: plan.PlanID, ExpectedRecordVersion: plan.RecordVersion,
		InputText: "Do not execute twice.", Model: "mock",
	})
	if repeated.FailureCode != WorkflowRunFailureToolConfirmation || *networkAttempts != 1 || testBridge.callCount() != 1 {
		t.Fatalf("repeated execution was not rejected before side effects: %#v network=%d bridge=%d", repeated, *networkAttempts, testBridge.callCount())
	}
}

func TestWorkflowHTTPToolExecutionKnownResponseFailureIsTerminalWithoutProviderCall(t *testing.T) {
	service, ctx, plan, runStore, testBridge, networkAttempts := newWorkflowHTTPToolExecutionServiceForTest(
		t,
		func(*http.Request) (*http.Response, error) {
			return workflowHTTPToolJSONResponse(http.StatusServiceUnavailable, `{"error":"unavailable"}`), nil
		},
	)
	result := service.Execute(ctx, WorkflowHTTPToolExecutionRequest{
		PlanID: plan.PlanID, ExpectedRecordVersion: plan.RecordVersion, InputText: "Review.", Model: "mock",
	})
	if result.FailureCode != WorkflowRunFailureToolResponseStatus || result.Record == nil ||
		result.Record.Status != WorkflowRunStatusFailed || result.Record.ToolAttempt == nil ||
		result.Record.ToolAttempt.Status != WorkflowHTTPToolAttemptFailed ||
		result.Record.Diagnostic == nil || result.Record.Diagnostic.ToolFailureCategory != WorkflowHTTPToolFailureResponseStatus ||
		result.Record.Diagnostic.FailureBoundary != WorkflowRunFailureBoundaryToolResponse ||
		*networkAttempts != 1 || testBridge.callCount() != 0 {
		t.Fatalf("known response failure did not stop at the tool boundary: %#v network=%d bridge=%d", result, *networkAttempts, testBridge.callCount())
	}
	stored, found, err := runStore.ReadRun(workflowRunContextFromToolAction(ctx), result.Record.RunID)
	if err != nil || !found || stored.Status != WorkflowRunStatusFailed || stored.RecordVersion != 2 {
		t.Fatalf("known response failure was not stored: %#v found=%t err=%v", stored, found, err)
	}
}

func TestWorkflowHTTPToolExecutionAmbiguousTransportFailureBecomesOutcomeUnknown(t *testing.T) {
	service, ctx, plan, runStore, testBridge, networkAttempts := newWorkflowHTTPToolExecutionServiceForTest(
		t,
		func(*http.Request) (*http.Response, error) {
			return nil, context.DeadlineExceeded
		},
	)
	result := service.Execute(ctx, WorkflowHTTPToolExecutionRequest{
		PlanID: plan.PlanID, ExpectedRecordVersion: plan.RecordVersion, InputText: "Review.", Model: "mock",
	})
	if result.FailureCode != WorkflowRunFailureToolOutcomeUnknown || result.Record == nil ||
		result.Record.Status != WorkflowRunStatusOutcomeUnknown || result.Record.ToolAttempt == nil ||
		result.Record.ToolAttempt.Status != WorkflowHTTPToolAttemptOutcomeUnknown ||
		result.Record.Diagnostic == nil || result.Record.Diagnostic.ToolFailureCategory != WorkflowHTTPToolFailureOutcomeUnknown ||
		result.Record.Diagnostic.RecommendedReviewAction != WorkflowRunReviewToolOutcome ||
		*networkAttempts != 1 || testBridge.callCount() != 0 {
		t.Fatalf("ambiguous transport did not become outcome_unknown: %#v network=%d bridge=%d", result, *networkAttempts, testBridge.callCount())
	}
	stored, found, err := runStore.ReadRun(workflowRunContextFromToolAction(ctx), result.Record.RunID)
	if err != nil || !found || stored.Status != WorkflowRunStatusOutcomeUnknown || stored.RecordVersion != 2 {
		t.Fatalf("outcome_unknown was not stored: %#v found=%t err=%v", stored, found, err)
	}
}

func TestWorkflowHTTPToolExecutionPreservesSuccessfulAttemptWhenModelFails(t *testing.T) {
	service, ctx, plan, runStore, testBridge, networkAttempts := newWorkflowHTTPToolExecutionServiceForTest(t, nil)
	var completionAudit WorkflowHTTPToolExecutionAudit
	service.store = observingWorkflowHTTPToolExecutionStore{
		workflowHTTPToolExecutionStore: service.store,
		beforeComplete: func(_ *WorkflowHTTPToolExecutionAttempt, _ *WorkflowRunRecord, audit WorkflowHTTPToolExecutionAudit) {
			completionAudit = audit
		},
	}
	testBridge.handle = func(context.Context, []byte, bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
		return bridge.GatewayEnvelope{}, errors.New("injected gateway failure")
	}
	result := service.Execute(ctx, WorkflowHTTPToolExecutionRequest{
		PlanID: plan.PlanID, ExpectedRecordVersion: plan.RecordVersion, InputText: "Review.", Model: "mock",
	})
	if result.FailureCode != WorkflowRunFailureGatewayFailed || result.Record == nil ||
		result.Record.Status != WorkflowRunStatusFailed || result.Record.ToolAttempt == nil ||
		result.Record.ToolAttempt.Status != WorkflowHTTPToolAttemptSucceeded ||
		len(result.Record.ToolAttempt.OutputProjection) == 0 || result.Record.SideEffects.ProviderCalls != 1 ||
		result.Record.Diagnostic == nil || result.Record.Diagnostic.FailureBoundary != WorkflowRunFailureBoundaryGateway ||
		result.Record.Diagnostic.ToolFailureCategory != WorkflowHTTPToolFailureNone ||
		completionAudit.FailureBoundary != nil || completionAudit.EventKind != "tool_execution_succeeded" ||
		*networkAttempts != 1 || testBridge.callCount() != 1 {
		t.Fatalf("downstream model failure corrupted the successful tool attempt: %#v audit=%#v network=%d bridge=%d", result, completionAudit, *networkAttempts, testBridge.callCount())
	}
	stored, found, err := runStore.ReadRun(workflowRunContextFromToolAction(ctx), result.Record.RunID)
	if err != nil || !found || stored.Status != WorkflowRunStatusFailed || stored.RecordVersion != 2 ||
		stored.ToolAttempt == nil || stored.ToolAttempt.Status != WorkflowHTTPToolAttemptSucceeded {
		t.Fatalf("downstream model failure was not stored with its successful tool attempt: %#v found=%t err=%v", stored, found, err)
	}
}

type observingWorkflowHTTPToolExecutionStore struct {
	workflowHTTPToolExecutionStore
	beforeComplete func(*WorkflowHTTPToolExecutionAttempt, *WorkflowRunRecord, WorkflowHTTPToolExecutionAudit)
}

func (store observingWorkflowHTTPToolExecutionStore) CompleteExecution(
	ctx WorkflowHTTPToolActionContext,
	attempt *WorkflowHTTPToolExecutionAttempt,
	run *WorkflowRunRecord,
	audit WorkflowHTTPToolExecutionAudit,
) error {
	store.beforeComplete(attempt, run, audit)
	return store.workflowHTTPToolExecutionStore.CompleteExecution(ctx, attempt, run, audit)
}

func TestWorkflowHTTPToolExecutionPreClaimFailuresHaveNoNetworkSideEffects(t *testing.T) {
	service, ctx, plan, _, testBridge, networkAttempts := newWorkflowHTTPToolExecutionServiceForTest(t, nil)
	ctx.ScopeGrants = []string{"workflow_tool_actions:execute", "workflow_drafts:read"}
	result := service.Execute(ctx, WorkflowHTTPToolExecutionRequest{
		PlanID: plan.PlanID, ExpectedRecordVersion: plan.RecordVersion, InputText: "Review.", Model: "mock",
	})
	if result.FailureCode != WorkflowRunFailureScopeDenied || result.Record != nil || *networkAttempts != 0 || testBridge.callCount() != 0 {
		t.Fatalf("scope failure crossed the pre-claim boundary: %#v network=%d bridge=%d", result, *networkAttempts, testBridge.callCount())
	}
	storedPlan, found, err := service.actions.store.ReadPlan(ctx, plan.PlanID)
	if err != nil || !found || storedPlan.Status != WorkflowHTTPToolActionStatusApproved || storedPlan.RecordVersion != plan.RecordVersion {
		t.Fatalf("pre-claim failure changed the approved plan: %#v found=%t err=%v", storedPlan, found, err)
	}
}

func TestWorkflowHTTPToolExecutionTerminalWriteFailureReturnsOutcomeUnknownWithoutRetry(t *testing.T) {
	service, ctx, plan, runStore, testBridge, networkAttempts := newWorkflowHTTPToolExecutionServiceForTest(t, nil)
	service.store = failingWorkflowHTTPToolCompletionStore{delegate: service.store}
	result := service.Execute(ctx, WorkflowHTTPToolExecutionRequest{
		PlanID: plan.PlanID, ExpectedRecordVersion: plan.RecordVersion, InputText: "Review.", Model: "mock",
	})
	if result.FailureCode != WorkflowRunFailureToolOutcomeUnknown || result.Record == nil ||
		result.Record.Status != WorkflowRunStatusOutcomeUnknown || result.Record.Diagnostic == nil ||
		result.Record.Diagnostic.TerminalWriteState != WorkflowRunTerminalWritePending ||
		*networkAttempts != 1 || testBridge.callCount() != 1 {
		t.Fatalf("terminal write failure did not fail closed without retry: %#v network=%d bridge=%d", result, *networkAttempts, testBridge.callCount())
	}
	stored, found, err := runStore.ReadRun(workflowRunContextFromToolAction(ctx), result.Record.RunID)
	if err != nil || !found || stored.Status != WorkflowRunStatusRunning || stored.RecordVersion != 1 ||
		stored.ToolAttempt == nil || stored.ToolAttempt.Status != WorkflowHTTPToolAttemptClaimed {
		t.Fatalf("last durable claimed state was not preserved for reconciliation: %#v found=%t err=%v", stored, found, err)
	}
}

func TestWorkflowHTTPToolReconciliationOnlyTransitionsStaleClaimedMetadata(t *testing.T) {
	runStore := newMemoryWorkflowRunStore(8)
	actionStore := newMemoryWorkflowHTTPToolActionStore(&runStore.mu)
	executionStore := newMemoryWorkflowHTTPToolExecutionStore(actionStore, runStore)
	ctx, approvedPlan, confirmation := seedApprovedMemoryWorkflowHTTPToolPlan(t, actionStore)
	ctx.ScopeGrants = append(ctx.ScopeGrants, workflowHTTPToolExecutionRequiredScopes...)
	plan := cloneWorkflowHTTPToolActionPlan(approvedPlan)
	plan.Status = WorkflowHTTPToolActionStatusConsumed
	plan.RecordVersion++
	attempt, run, audit := workflowHTTPToolClaimFixture(ctx, plan, confirmation, 0)
	if err := executionStore.ClaimExecution(ctx, &plan, confirmation, &attempt, &run, audit); err != nil {
		t.Fatalf("claim execution for reconciliation: %v", err)
	}
	service := workflowHTTPToolExecutionService{store: executionStore}
	now := time.Date(2026, 7, 16, 9, 2, 20, 0, time.UTC)
	service.now = func() time.Time { return now }
	service.newID = func(string) (string, error) { return "wtae_2222222222222222", nil }

	notStale := service.ReconcileStale(ctx, plan.PlanID)
	if notStale.FailureCode != "" || notStale.Record == nil || notStale.Record.Status != WorkflowRunStatusRunning {
		t.Fatalf("fresh claim should remain running: %#v", notStale)
	}
	now = time.Date(2026, 7, 16, 9, 2, 31, 0, time.UTC)
	reconciled := service.ReconcileStale(ctx, plan.PlanID)
	if reconciled.FailureCode != WorkflowRunFailureToolOutcomeUnknown || reconciled.Record == nil ||
		reconciled.Record.Status != WorkflowRunStatusOutcomeUnknown || reconciled.Record.ToolAttempt == nil ||
		reconciled.Record.ToolAttempt.Status != WorkflowHTTPToolAttemptOutcomeUnknown ||
		reconciled.Record.Diagnostic == nil || reconciled.Record.Diagnostic.TerminalWriteState != WorkflowRunTerminalWriteStored {
		t.Fatalf("stale metadata was not reconciled to outcome_unknown: %#v", reconciled)
	}
	stored, found, err := runStore.ReadRun(workflowRunContextFromToolAction(ctx), run.RunID)
	if err != nil || !found || stored.Status != WorkflowRunStatusOutcomeUnknown || stored.RecordVersion != 2 {
		t.Fatalf("reconciled run was not stored: %#v found=%t err=%v", stored, found, err)
	}
	repeated := service.ReconcileStale(ctx, plan.PlanID)
	if repeated.FailureCode != WorkflowRunFailureToolConfirmation {
		t.Fatalf("terminal execution should not reconcile twice: %#v", repeated)
	}
}

func newWorkflowHTTPToolExecutionServiceForTest(
	t *testing.T,
	roundTrip workflowHTTPToolRoundTripperFunc,
) (workflowHTTPToolExecutionService, WorkflowHTTPToolActionContext, WorkflowHTTPToolActionPlan, *memoryWorkflowRunStore, *workflowExecutorTestBridge, *int) {
	t.Helper()
	draft := workflowHTTPToolEligibleDraftForTest()
	actionService, actionStore, runStore, _, actionClock := newWorkflowHTTPToolActionServiceForTest(t, &draft)
	ctx := workflowHTTPToolActionTestContext()
	created := actionService.CreatePlan(ctx, WorkflowHTTPToolCreatePlanRequest{
		DraftID: draft.DraftID, DraftVersion: draft.DraftVersion, NodeID: "node_http_tool",
		PublicArguments: map[string]any{"resource_key": "docs/radishflow/overview", "locale": "zh-CN"},
	})
	if created.FailureCode != "" || created.ActionPlan == nil {
		t.Fatalf("create action plan: %#v", created)
	}
	approved := actionService.DecidePlan(ctx, WorkflowHTTPToolDecisionRequest{
		PlanID: created.ActionPlan.PlanID, ExpectedRecordVersion: created.ActionPlan.RecordVersion,
		Decision: WorkflowHTTPToolConfirmationApprove,
	})
	if approved.FailureCode != "" || approved.ActionPlan == nil {
		t.Fatalf("approve action plan: %#v", approved)
	}
	ctx.ScopeGrants = append(ctx.ScopeGrants, workflowHTTPToolExecutionRequiredScopes...)
	ctx.AuditRef = "audit_workflow_http_tool_execution_test"
	executionStore := newMemoryWorkflowHTTPToolExecutionStore(actionStore, runStore)
	testBridge := &workflowExecutorTestBridge{handle: func(context.Context, []byte, bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
		return successfulWorkflowExecutorEnvelope("reviewable workflow answer"), nil
	}}
	executor := workflowExecutorTestService(draft, testBridge, runStore)
	service := newWorkflowHTTPToolExecutionService(actionService, executionStore, executor)
	clock := actionClock.Add(2 * time.Minute)
	service.now = func() time.Time {
		clock = clock.Add(time.Millisecond)
		return clock
	}
	sequence := 0
	service.newID = func(prefix string) (string, error) {
		sequence++
		return fmt.Sprintf("%s%024x", prefix, sequence), nil
	}
	service.newRunID = func() (string, error) { return "run_1234567890abcdef", nil }
	networkAttempts := 0
	if roundTrip == nil {
		roundTrip = func(*http.Request) (*http.Response, error) {
			networkAttempts++
			return workflowHTTPToolJSONResponse(http.StatusOK, `{"resource_key":"docs/radishflow/overview","title":"RadishFlow","summary":"Reviewed resource","updated_at":"2026-07-16T09:00:00Z"}`), nil
		}
	} else {
		delegate := roundTrip
		roundTrip = func(request *http.Request) (*http.Response, error) {
			networkAttempts++
			return delegate(request)
		}
	}
	service.transport = workflowHTTPToolTestTransport(roundTrip)
	service.transport.now = service.now
	return service, ctx, *approved.ActionPlan, runStore, testBridge, &networkAttempts
}

type failingWorkflowHTTPToolCompletionStore struct {
	delegate workflowHTTPToolExecutionStore
}

func (store failingWorkflowHTTPToolCompletionStore) ReadApprovedConfirmation(
	ctx WorkflowHTTPToolActionContext,
	planID string,
	recordVersion int,
) (WorkflowHTTPToolConfirmationDecision, bool, error) {
	return store.delegate.ReadApprovedConfirmation(ctx, planID, recordVersion)
}

func (store failingWorkflowHTTPToolCompletionStore) ReadClaimedExecution(
	ctx WorkflowHTTPToolActionContext,
	planID string,
) (WorkflowHTTPToolExecutionAttempt, WorkflowRunRecord, bool, error) {
	return store.delegate.ReadClaimedExecution(ctx, planID)
}

func (store failingWorkflowHTTPToolCompletionStore) ClaimExecution(
	ctx WorkflowHTTPToolActionContext,
	plan *WorkflowHTTPToolActionPlan,
	confirmation WorkflowHTTPToolConfirmationDecision,
	attempt *WorkflowHTTPToolExecutionAttempt,
	run *WorkflowRunRecord,
	audit WorkflowHTTPToolExecutionAudit,
) error {
	return store.delegate.ClaimExecution(ctx, plan, confirmation, attempt, run, audit)
}

func (store failingWorkflowHTTPToolCompletionStore) CompleteExecution(
	WorkflowHTTPToolActionContext,
	*WorkflowHTTPToolExecutionAttempt,
	*WorkflowRunRecord,
	WorkflowHTTPToolExecutionAudit,
) error {
	return errWorkflowHTTPToolExecutionUnavailable
}

var _ workflowHTTPToolExecutionStore = failingWorkflowHTTPToolCompletionStore{}
