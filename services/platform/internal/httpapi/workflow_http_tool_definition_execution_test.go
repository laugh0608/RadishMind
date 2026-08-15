package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/bridge"
)

func TestWorkflowDefinitionHTTPToolExecutionPersistsStrictV9Evidence(t *testing.T) {
	service, ctx, plan, runStore, bridgeClient, networkAttempts := newWorkflowDefinitionHTTPToolExecutionServiceForTest(t)
	input := "private definition tool input"
	result := service.Execute(ctx, WorkflowHTTPToolExecutionRequest{
		PlanID: plan.PlanID, ExpectedRecordVersion: plan.RecordVersion, InputText: input,
	})
	if result.FailureCode != "" || result.Record == nil || result.Record.Status != WorkflowRunStatusSucceeded {
		t.Fatalf("execute Definition-bound HTTP tool plan: %#v", result)
	}
	record := result.Record
	if record.SchemaVersion != workflowRunRecordDefinitionToolSchemaVersion ||
		record.ExecutionKind != workflowDefinitionHTTPToolExecutionKind ||
		record.ExecutionSourceKind != workflowDefinitionExecutionSourceKind ||
		record.ExecutionSourceID != plan.WorkflowDefinitionID ||
		record.ExecutionSourceVersion != plan.WorkflowDefinitionVersion ||
		record.ExecutionProfile != workflowDefinitionHTTPToolProfile || record.DefinitionAuthority == nil ||
		record.DefinitionAuthority.ActivationPointerVersion != plan.ActivationPointerVersion {
		t.Fatalf("v9 Definition authority projection drifted: %#v", record)
	}
	if record.DraftID != "" || record.DraftVersion != 0 || record.DraftDigest != "" || record.Output != "" ||
		record.InputDigest != workflowDefinitionInputDigest(input) || record.InputBytes != len([]byte(input)) {
		t.Fatalf("v9 privacy projection is incomplete: %#v", record)
	}
	for _, node := range record.Nodes {
		if node.OutputPreview != "" {
			t.Fatalf("v9 persisted node output preview: %#v", node)
		}
	}
	if *networkAttempts != 1 || bridgeClient.callCount() != 1 || record.SideEffects.ToolCalls != 1 ||
		record.SideEffects.ConfirmationCalls != 1 || record.SideEffects.ProviderCalls != 1 {
		t.Fatalf("v9 side-effect evidence drifted: side_effects=%#v network=%d provider=%d", record.SideEffects, *networkAttempts, bridgeClient.callCount())
	}
	payload, err := json.Marshal(record)
	if err != nil || strings.Contains(string(payload), input) || strings.Contains(string(payload), "reviewable workflow answer") {
		t.Fatalf("v9 persisted sensitive execution material: err=%v payload=%s", err, payload)
	}
	stored, found, err := runStore.ReadRun(workflowRunContextFromToolAction(ctx), record.RunID)
	if err != nil || !found || stored.SchemaVersion != workflowRunRecordDefinitionToolSchemaVersion || stored.Output != "" {
		t.Fatalf("read stored v9 run: found=%t err=%v record=%#v", found, err, stored)
	}
	withOutput := cloneWorkflowRunRecord(stored)
	withOutput.Output = "forbidden persisted answer"
	if validateWorkflowRunStoreRecord(workflowRunContextFromToolAction(ctx), &withOutput) == nil {
		t.Fatal("v9 accepted a persisted final output")
	}
	withPreview := cloneWorkflowRunRecord(stored)
	withPreview.Nodes[0].OutputPreview = "forbidden persisted preview"
	if validateWorkflowRunStoreRecord(workflowRunContextFromToolAction(ctx), &withPreview) == nil {
		t.Fatal("v9 accepted a persisted node output preview")
	}
	runContext := workflowRunContextFromToolAction(ctx)
	comparisonRecord := terminalComparisonTestRun(runContext, "run_0000000000009002", WorkflowRunStatusSucceeded, time.Now().UTC())
	storeTerminalComparisonTestRun(t, runStore, runContext, &comparisonRecord)
	comparison := service.executor.CompareRuns(runContext, comparisonRecord.RunID, stored.RunID)
	if comparison.FailureCode != WorkflowRunFailureSideEffectUnsupported || comparison.Comparison != nil {
		t.Fatalf("v9 comparison did not preserve side-effect rejection: %#v", comparison)
	}
	evaluation := newWorkflowEvaluationService(newMemoryWorkflowEvaluationStore(4), runStore).Create(
		runContext,
		WorkflowEvaluationCreateRequest{
			Name: "Definition tool side effects", BaselineRunID: comparisonRecord.RunID,
			Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: stored.RunID, ExpectedClassification: WorkflowRunComparisonChanged}},
		},
	)
	if evaluation.FailureCode != WorkflowEvaluationFailureSideEffectProfile || evaluation.Case != nil {
		t.Fatalf("v9 evaluation did not preserve side-effect rejection: %#v", evaluation)
	}
	repeated := service.Execute(ctx, WorkflowHTTPToolExecutionRequest{
		PlanID: plan.PlanID, ExpectedRecordVersion: plan.RecordVersion, InputText: input,
	})
	if repeated.FailureCode != WorkflowRunFailureToolConfirmation || *networkAttempts != 1 || bridgeClient.callCount() != 1 {
		t.Fatalf("consumed Definition plan executed twice: result=%#v network=%d provider=%d", repeated, *networkAttempts, bridgeClient.callCount())
	}
}

func TestWorkflowDefinitionHTTPToolExecutionConcurrentClaimRunsOnce(t *testing.T) {
	service, ctx, plan, runStore, bridgeClient, networkAttempts := newWorkflowDefinitionHTTPToolExecutionServiceForTest(t)
	fixedNow := time.Date(2026, 8, 15, 9, 4, 0, 0, time.UTC)
	service.now = func() time.Time { return fixedNow }
	service.transport.now = service.now
	service.newID = func(prefix string) (string, error) { return prefix + "0000000000009999", nil }
	service.newRunID = func() (string, error) { return "run_0000000000009999", nil }
	results := make(chan WorkflowHTTPToolExecutionResult, 2)
	var group sync.WaitGroup
	for range 2 {
		group.Add(1)
		go func() {
			defer group.Done()
			results <- service.Execute(ctx, WorkflowHTTPToolExecutionRequest{
				PlanID: plan.PlanID, ExpectedRecordVersion: plan.RecordVersion, InputText: "bounded concurrent input",
			})
		}()
	}
	group.Wait()
	close(results)
	succeeded, rejected := 0, 0
	for result := range results {
		switch result.FailureCode {
		case "":
			succeeded++
		case WorkflowRunFailureToolConfirmation:
			rejected++
		default:
			t.Fatalf("unexpected concurrent execution result: %#v", result)
		}
	}
	if succeeded != 1 || rejected != 1 || *networkAttempts != 1 || bridgeClient.callCount() != 1 || len(runStore.records) != 1 {
		t.Fatalf("Definition plan was not claimed exactly once: succeeded=%d rejected=%d network=%d provider=%d runs=%d", succeeded, rejected, *networkAttempts, bridgeClient.callCount(), len(runStore.records))
	}
}

func TestWorkflowDefinitionHTTPToolExecutionRequiresDefinitionReadScope(t *testing.T) {
	service, ctx, plan, runStore, bridgeClient, networkAttempts := newWorkflowDefinitionHTTPToolExecutionServiceForTest(t)
	ctx.ScopeGrants = append([]string(nil), workflowHTTPToolExecutionRequiredScopes...)
	result := service.Execute(ctx, WorkflowHTTPToolExecutionRequest{
		PlanID: plan.PlanID, ExpectedRecordVersion: plan.RecordVersion, InputText: "bounded input",
	})
	if result.FailureCode != WorkflowRunFailureScopeDenied || *networkAttempts != 0 || bridgeClient.callCount() != 0 || len(runStore.records) != 0 {
		t.Fatalf("Definition source scope did not fail before claim: result=%#v network=%d provider=%d runs=%d", result, *networkAttempts, bridgeClient.callCount(), len(runStore.records))
	}
}

func TestWorkflowDefinitionHTTPToolExecutionRechecksAuthorityAfterClaim(t *testing.T) {
	service, ctx, plan, runStore, bridgeClient, networkAttempts := newWorkflowDefinitionHTTPToolExecutionServiceForTest(t)
	service.applications = &driftingApplicationCatalogRepository{applicationCatalogRepository: service.applications}
	result := service.Execute(ctx, WorkflowHTTPToolExecutionRequest{
		PlanID: plan.PlanID, ExpectedRecordVersion: plan.RecordVersion, InputText: "bounded input",
	})
	if result.FailureCode != WorkflowRunFailureDefinitionAuthority || result.Record == nil ||
		result.Record.Status != WorkflowRunStatusFailed || result.Record.SchemaVersion != workflowRunRecordDefinitionToolSchemaVersion ||
		*networkAttempts != 0 || bridgeClient.callCount() != 0 || result.Record.SideEffects.ProviderCalls != 0 {
		t.Fatalf("post-claim authority drift did not fail before side effects: result=%#v network=%d provider=%d", result, *networkAttempts, bridgeClient.callCount())
	}
	stored, found, err := runStore.ReadRun(workflowRunContextFromToolAction(ctx), result.Record.RunID)
	if err != nil || !found || stored.FailureCode != WorkflowRunFailureDefinitionAuthority || stored.ToolAttempt == nil ||
		stored.ToolAttempt.Status != WorkflowHTTPToolAttemptFailed {
		t.Fatalf("post-claim authority failure evidence missing: found=%t err=%v run=%#v", found, err, stored)
	}
}

func newWorkflowDefinitionHTTPToolExecutionServiceForTest(
	t *testing.T,
) (workflowHTTPToolExecutionService, WorkflowHTTPToolActionContext, WorkflowHTTPToolActionPlan, *memoryWorkflowRunStore, *workflowExecutorTestBridge, *int) {
	t.Helper()
	actor := "subject_demo_user"
	applicationID := "app_aaaaaaaaaaaaaaaa"
	draft := workflowHTTPToolEligibleDraftForTest()
	draft.ApplicationID = applicationID
	ctx := workflowHTTPToolActionTestContext()
	ctx.ApplicationID = applicationID
	ctx.ScopeGrants = append(ctx.ScopeGrants, workflowDefinitionHTTPToolExecutionRequiredScopes...)

	applications := newMemoryApplicationCatalogRepository()
	applicationService := newApplicationCatalogService(applications)
	applicationService.newID = func() (string, error) { return applicationID, nil }
	applicationContext := ApplicationCatalogContext{
		RequestContext: context.Background(), RequestID: "request_definition_tool_app", TenantRef: ctx.TenantRef,
		WorkspaceID: ctx.WorkspaceID, OwnerSubjectRef: actor, ActorRef: actor, AuditRef: "audit_definition_tool_app", WriteEnabled: true,
	}
	if created := applicationService.Create(applicationContext, ApplicationCatalogCreateInput{DisplayName: "Definition tool app", ApplicationKind: "workflow_copilot"}); created.FailureCode != "" || created.Record == nil {
		t.Fatalf("create Definition tool application: %#v", created)
	}

	definitions := newWorkflowDefinitionReleaseStore()
	releaseContext := WorkflowDefinitionReleaseContext{
		RequestContext: context.Background(), TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
		ApplicationID: applicationID, OwnerSubjectRef: actor, ActorRef: actor,
		RequestID: "request_definition_tool_release", AuditRef: "audit_definition_tool_release",
	}
	now := time.Date(2026, 8, 15, 9, 0, 0, 0, time.UTC)
	candidate, err := definitions.CreateCandidate(releaseContext, "candidate_definition_tool", "definition_tool", workflowDefinitionHTTPToolProfile, draft, now)
	if err != nil {
		t.Fatalf("create Definition tool candidate: %v", err)
	}
	_, version, err := definitions.Review(releaseContext, candidate.CandidateID, 0, "approve", "approve Definition tool", candidate.SourceDraftDigest, now.Add(time.Minute))
	if err != nil || version == nil {
		t.Fatalf("approve Definition tool version: version=%#v err=%v", version, err)
	}
	if _, err = definitions.DecideActivation(releaseContext, version.DefinitionID, 0, "activate", version.Version, "activate Definition tool", now.Add(2*time.Minute)); err != nil {
		t.Fatalf("activate Definition tool version: %v", err)
	}

	runStore := newMemoryWorkflowRunStore(20)
	actionStore := newMemoryWorkflowHTTPToolActionStore(&runStore.mu)
	actionService, err := newWorkflowHTTPToolActionService(
		func(SavedWorkflowDraftContext, ReadWorkflowDraftRequest) SavedWorkflowDraftResult {
			return SavedWorkflowDraftResult{FailureCode: SavedWorkflowDraftFailureNotFound}
		},
		definitions,
		actionStore,
	)
	if err != nil {
		t.Fatalf("create Definition tool action service: %v", err)
	}
	clock := now.Add(3 * time.Minute)
	actionService.now = func() time.Time { return clock }
	sequence := 0
	actionService.newID = func(prefix string) (string, error) {
		sequence++
		return fmt.Sprintf("%s%024x", prefix, sequence), nil
	}
	created := actionService.CreatePlan(ctx, WorkflowHTTPToolCreatePlanRequest{
		DefinitionID: version.DefinitionID, NodeID: "node_http_tool",
		PublicArguments: map[string]any{"resource_key": "docs/radishflow/overview", "locale": "zh-CN"},
	})
	if created.FailureCode != "" || created.ActionPlan == nil {
		t.Fatalf("create Definition-bound plan: %#v", created)
	}
	approved := actionService.DecidePlan(ctx, WorkflowHTTPToolDecisionRequest{
		PlanID: created.ActionPlan.PlanID, ExpectedRecordVersion: created.ActionPlan.RecordVersion,
		Decision: WorkflowHTTPToolConfirmationApprove,
	})
	if approved.FailureCode != "" || approved.ActionPlan == nil {
		t.Fatalf("approve Definition-bound plan: %#v", approved)
	}
	if readable := actionService.ReadPlan(ctx, approved.ActionPlan.PlanID); readable.FailureCode != "" {
		t.Fatalf("re-read approved Definition-bound plan: %#v", readable)
	}

	bridgeClient := &workflowExecutorTestBridge{handle: func(context.Context, []byte, bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
		return successfulWorkflowExecutorEnvelope("reviewable workflow answer"), nil
	}}
	executor := workflowExecutorTestService(draft, bridgeClient, runStore)
	service := newWorkflowHTTPToolExecutionService(actionService, newMemoryWorkflowHTTPToolExecutionStore(actionStore, runStore), executor, applications)
	service.now = func() time.Time {
		clock = clock.Add(time.Millisecond)
		return clock
	}
	service.newID = func(prefix string) (string, error) {
		sequence++
		return fmt.Sprintf("%s%024x", prefix, sequence), nil
	}
	service.newRunID = func() (string, error) { return "run_0000000000009001", nil }
	networkAttempts := 0
	service.transport = workflowHTTPToolTestTransport(func(*http.Request) (*http.Response, error) {
		networkAttempts++
		return workflowHTTPToolJSONResponse(http.StatusOK, `{"resource_key":"docs/radishflow/overview","title":"RadishFlow","summary":"Reviewed resource","updated_at":"2026-08-15T09:00:00Z"}`), nil
	})
	service.transport.now = service.now
	return service, ctx, *approved.ActionPlan, runStore, bridgeClient, &networkAttempts
}
