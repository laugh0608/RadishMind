package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestWorkflowHTTPToolRegistryMatchesCommittedContractFixture(t *testing.T) {
	payload, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "scripts", "checks", "fixtures", "workflow-http-tool-contracts-v1.json"))
	if err != nil {
		t.Fatalf("read committed HTTP tool contract fixture: %v", err)
	}
	var fixture struct {
		Positive struct {
			Definition       WorkflowHTTPToolDefinition       `json:"definition"`
			ExecutionProfile WorkflowHTTPToolExecutionProfile `json:"execution_profile"`
			ActionPlan       WorkflowHTTPToolActionPlan       `json:"action_plan"`
		} `json:"positive"`
	}
	if err := json.Unmarshal(payload, &fixture); err != nil {
		t.Fatalf("decode committed HTTP tool contract fixture: %v", err)
	}
	registry, err := newWorkflowHTTPToolRegistry()
	if err != nil {
		t.Fatalf("build HTTP tool registry: %v", err)
	}
	if !reflect.DeepEqual(registry.definition, fixture.Positive.Definition) {
		t.Fatalf("runtime definition drifted from committed contract:\nruntime=%#v\nfixture=%#v", registry.definition, fixture.Positive.Definition)
	}
	if !reflect.DeepEqual(registry.profile, fixture.Positive.ExecutionProfile) {
		t.Fatalf("runtime execution profile drifted from committed contract:\nruntime=%#v\nfixture=%#v", registry.profile, fixture.Positive.ExecutionProfile)
	}
	plan := fixture.Positive.ActionPlan
	computedPlanDigest, err := workflowHTTPToolPlanDigest(plan)
	if err != nil {
		t.Fatalf("canonicalize committed action plan: %v", err)
	}
	if plan.DefinitionDigest != registry.definitionDigest || plan.ProfileDigest != registry.profileDigest ||
		plan.OutputSchemaDigest != registry.outputSchemaDigest || plan.ToolPlanDigest != computedPlanDigest {
		t.Fatalf("committed digest chain drifted: plan=%#v registry=%#v computed_plan=%s", plan, registry, computedPlanDigest)
	}
}

func TestRestrictedJCSCanonicalizationDomain(t *testing.T) {
	left, err := canonicalSHA256(map[string]any{"z": 2, "a": []any{"safe", true, nil}})
	if err != nil {
		t.Fatalf("canonicalize ordered input: %v", err)
	}
	right, err := canonicalSHA256(map[string]any{"a": []any{"safe", true, nil}, "z": 2})
	if err != nil || left != right {
		t.Fatalf("object insertion order changed canonical digest: left=%s right=%s err=%v", left, right, err)
	}
	for name, input := range map[string]any{
		"floating point": map[string]any{"value": 1.5},
		"unicode":        map[string]any{"value": "非 ASCII"},
		"html sensitive": map[string]any{"value": "a&b"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := canonicalSHA256(input); err == nil {
				t.Fatalf("restricted JCS must reject %s input", name)
			}
		})
	}
}

func TestMemoryWorkflowHTTPToolActionStoreCASAndScopeIsolation(t *testing.T) {
	t.Run("one concurrent decision wins the expected record version", func(t *testing.T) {
		owner := &sync.RWMutex{}
		store := newMemoryWorkflowHTTPToolActionStore(owner)
		ctx := workflowHTTPToolActionTestContext()
		plan := workflowHTTPToolActionPlanForStoreTest(t, ctx, "wtap_0000000000000001")
		if err := store.CreatePlan(ctx, &plan, workflowHTTPToolAuditForStoreTest(plan, "wtae_0000000000000001", "confirmation_requested")); err != nil {
			t.Fatalf("create action plan: %v", err)
		}

		outcomes := []WorkflowHTTPToolConfirmationOutcome{
			WorkflowHTTPToolConfirmationApprove,
			WorkflowHTTPToolConfirmationReject,
		}
		start := make(chan struct{})
		results := make(chan error, len(outcomes))
		for index, outcome := range outcomes {
			index, outcome := index, outcome
			go func() {
				<-start
				updated := cloneWorkflowHTTPToolActionPlan(plan)
				updated.RecordVersion = 2
				updated.Status = workflowHTTPToolStatusForOutcome(outcome)
				actor := fmt.Sprintf("subject_decider_%d", index)
				decidedAt := "2026-07-16T09:01:00Z"
				updated.LastDecisionByActorRef = &actor
				updated.LastDecisionAt = &decidedAt
				decision := workflowHTTPToolDecisionForStoreTest(updated, fmt.Sprintf("wtcd_%016d", index+1), outcome, actor)
				audit := workflowHTTPToolAuditForStoreTest(updated, fmt.Sprintf("wtae_%016d", index+2), workflowHTTPToolAuditEventForOutcome(outcome), decision.ConfirmationID)
				decisionContext := ctx
				decisionContext.ActorRef = actor
				results <- store.DecidePlan(decisionContext, &updated, decision, audit)
			}()
		}
		close(start)

		var succeeded, conflicted int
		for range outcomes {
			switch err := <-results; {
			case err == nil:
				succeeded++
			case errors.Is(err, errWorkflowHTTPToolActionConflict):
				conflicted++
			default:
				t.Fatalf("unexpected concurrent decision result: %v", err)
			}
		}
		if succeeded != 1 || conflicted != 1 {
			t.Fatalf("expected one successful CAS and one conflict, got success=%d conflict=%d", succeeded, conflicted)
		}

		stored, found, err := store.ReadPlan(ctx, plan.PlanID)
		if err != nil || !found || stored.RecordVersion != 2 {
			t.Fatalf("read winning plan: found=%v err=%v plan=%#v", found, err, stored)
		}
		if stored.Status != WorkflowHTTPToolActionStatusApproved && stored.Status != WorkflowHTTPToolActionStatusRejected {
			t.Fatalf("unexpected winning status: %s", stored.Status)
		}
		owner.RLock()
		decisionCount, auditCount := len(store.decisions), len(store.audits)
		owner.RUnlock()
		if decisionCount != 1 || auditCount != 2 {
			t.Fatalf("failed CAS must not append partial evidence: decisions=%d audits=%d", decisionCount, auditCount)
		}
	})

	t.Run("tenant workspace and application are part of the plan key", func(t *testing.T) {
		store := newMemoryWorkflowHTTPToolActionStore(nil)
		base := workflowHTTPToolActionTestContext()
		contexts := []WorkflowHTTPToolActionContext{
			base,
			workflowHTTPToolActionContextWithScope(base, "tenant_other", base.WorkspaceID, base.ApplicationID),
			workflowHTTPToolActionContextWithScope(base, base.TenantRef, "workspace_other", base.ApplicationID),
			workflowHTTPToolActionContextWithScope(base, base.TenantRef, base.WorkspaceID, "application_other"),
		}
		for index, scoped := range contexts {
			plan := workflowHTTPToolActionPlanForStoreTest(t, scoped, "wtap_0000000000000002")
			audit := workflowHTTPToolAuditForStoreTest(plan, fmt.Sprintf("wtae_%016d", index+10), "confirmation_requested")
			if err := store.CreatePlan(scoped, &plan, audit); err != nil {
				t.Fatalf("create scoped plan %d: %v", index, err)
			}
		}
		for index, scoped := range contexts {
			plan, found, err := store.ReadPlan(scoped, "wtap_0000000000000002")
			if err != nil || !found {
				t.Fatalf("read scoped plan %d: found=%v err=%v", index, found, err)
			}
			if plan.TenantRef != scoped.TenantRef || plan.WorkspaceID != scoped.WorkspaceID || plan.ApplicationID != scoped.ApplicationID {
				t.Fatalf("scope %d leaked a different plan: %#v", index, plan)
			}
		}
	})

	t.Run("store rejects mismatched audit evidence and illegal approved transition", func(t *testing.T) {
		store := newMemoryWorkflowHTTPToolActionStore(nil)
		ctx := workflowHTTPToolActionTestContext()
		plan := workflowHTTPToolActionPlanForStoreTest(t, ctx, "wtap_0000000000000003")
		mismatchedAudit := workflowHTTPToolAuditForStoreTest(plan, "wtae_0000000000000020", "confirmation_requested")
		mismatchedAudit.RequestID = "request_mismatched"
		if err := store.CreatePlan(ctx, &plan, mismatchedAudit); !errors.Is(err, errWorkflowHTTPToolActionContract) {
			t.Fatalf("mismatched create audit must fail closed, got %v", err)
		}
		if len(store.plans) != 0 || len(store.audits) != 0 {
			t.Fatalf("mismatched create audit left partial state: plans=%d audits=%d", len(store.plans), len(store.audits))
		}

		if err := store.CreatePlan(ctx, &plan, workflowHTTPToolAuditForStoreTest(plan, "wtae_0000000000000021", "confirmation_requested")); err != nil {
			t.Fatalf("create plan for transition check: %v", err)
		}
		tampered := cloneWorkflowHTTPToolActionPlan(plan)
		tampered.RecordVersion = 2
		tampered.Status = WorkflowHTTPToolActionStatusApproved
		tampered.PlannedByActorRef = "subject_tampered_planner"
		tamperedActor, tamperedAt := ctx.ActorRef, "2026-07-16T09:01:00Z"
		tampered.LastDecisionByActorRef, tampered.LastDecisionAt = &tamperedActor, &tamperedAt
		tampered.ToolPlanDigest, _ = workflowHTTPToolPlanDigest(tampered)
		tamperedDecision := workflowHTTPToolDecisionForStoreTest(tampered, "wtcd_0000000000000019", WorkflowHTTPToolConfirmationApprove, tamperedActor)
		tamperedAudit := workflowHTTPToolAuditForStoreTest(tampered, "wtae_0000000000000019", "confirmation_recorded", tamperedDecision.ConfirmationID)
		if err := store.DecidePlan(ctx, &tampered, tamperedDecision, tamperedAudit); !errors.Is(err, errWorkflowHTTPToolActionConflict) {
			t.Fatalf("immutable plan mutation must conflict, got %v", err)
		}
		if len(store.decisions) != 0 || len(store.audits) != 1 {
			t.Fatalf("immutable plan mutation left partial evidence: decisions=%d audits=%d", len(store.decisions), len(store.audits))
		}
		approved := cloneWorkflowHTTPToolActionPlan(plan)
		approved.RecordVersion = 2
		approved.Status = WorkflowHTTPToolActionStatusApproved
		actor, decidedAt := ctx.ActorRef, "2026-07-16T09:01:00Z"
		approved.LastDecisionByActorRef, approved.LastDecisionAt = &actor, &decidedAt
		approval := workflowHTTPToolDecisionForStoreTest(approved, "wtcd_0000000000000020", WorkflowHTTPToolConfirmationApprove, actor)
		approvalAudit := workflowHTTPToolAuditForStoreTest(approved, "wtae_0000000000000022", "confirmation_recorded", approval.ConfirmationID)
		if err := store.DecidePlan(ctx, &approved, approval, approvalAudit); err != nil {
			t.Fatalf("approve plan for transition check: %v", err)
		}

		illegal := cloneWorkflowHTTPToolActionPlan(approved)
		illegal.RecordVersion = 3
		illegal.Status = WorkflowHTTPToolActionStatusRejected
		illegalDecision := workflowHTTPToolDecisionForStoreTest(illegal, "wtcd_0000000000000021", WorkflowHTTPToolConfirmationReject, actor)
		illegalAudit := workflowHTTPToolAuditForStoreTest(illegal, "wtae_0000000000000023", "confirmation_rejected", illegalDecision.ConfirmationID)
		if err := store.DecidePlan(ctx, &illegal, illegalDecision, illegalAudit); !errors.Is(err, errWorkflowHTTPToolActionConflict) {
			t.Fatalf("approved plan must reject a reject transition, got %v", err)
		}
		if len(store.decisions) != 1 || len(store.audits) != 2 {
			t.Fatalf("illegal transition left partial evidence: decisions=%d audits=%d", len(store.decisions), len(store.audits))
		}
	})
}

func TestWorkflowHTTPToolActionServicePreRunBoundaries(t *testing.T) {
	t.Run("exact draft creates and approves a plan without execution side effects", func(t *testing.T) {
		draft := workflowHTTPToolEligibleDraftForTest()
		service, store, runStore, readCalls, _ := newWorkflowHTTPToolActionServiceForTest(t, &draft)
		ctx := workflowHTTPToolActionTestContext()

		created := service.CreatePlan(ctx, WorkflowHTTPToolCreatePlanRequest{
			DraftID: draft.DraftID, DraftVersion: draft.DraftVersion, NodeID: "node_http_tool",
			PublicArguments: map[string]any{"resource_key": "docs/radishflow/overview", "locale": "zh-CN"},
		})
		if created.FailureCode != "" || created.ActionPlan == nil {
			t.Fatalf("create exact action plan: %#v", created)
		}
		if created.ActionPlan.Status != WorkflowHTTPToolActionStatusPending || created.ActionPlan.RecordVersion != 1 || created.ActionPlan.ToolPlanDigest == "" {
			t.Fatalf("created plan is incomplete: %#v", created.ActionPlan)
		}

		approved := service.DecidePlan(ctx, WorkflowHTTPToolDecisionRequest{
			PlanID: created.ActionPlan.PlanID, ExpectedRecordVersion: 1, Decision: WorkflowHTTPToolConfirmationApprove,
		})
		if approved.FailureCode != "" || approved.ActionPlan == nil || approved.ConfirmationDecision == nil {
			t.Fatalf("approve exact action plan: %#v", approved)
		}
		if approved.ActionPlan.Status != WorkflowHTTPToolActionStatusApproved || approved.ActionPlan.RecordVersion != 2 ||
			approved.ConfirmationDecision.Outcome != WorkflowHTTPToolConfirmationApprove {
			t.Fatalf("approval did not preserve CAS state: %#v", approved)
		}
		if *readCalls != 2 {
			t.Fatalf("create and decision must each re-read the exact saved draft, got %d reads", *readCalls)
		}

		runStore.mu.RLock()
		runCount := len(runStore.records)
		runStore.mu.RUnlock()
		if runCount != 0 {
			t.Fatalf("batch A plan and approval must not create workflow runs, got %d", runCount)
		}
		store.ownerLock.RLock()
		defer store.ownerLock.RUnlock()
		if len(store.decisions) != 1 || len(store.audits) != 2 {
			t.Fatalf("expected one append-only decision and two audits: decisions=%d audits=%d", len(store.decisions), len(store.audits))
		}
		for _, audit := range store.audits {
			if audit.ExecutionAttemptID != nil || audit.RunID != nil || audit.SideEffects.NetworkAttempts != 0 ||
				audit.SideEffects.ToolCalls != 0 || audit.SideEffects.ProviderCalls != 0 ||
				audit.SideEffects.BusinessWrites != 0 || audit.SideEffects.ReplayWrites != 0 {
				t.Fatalf("batch A audit claimed an execution side effect: %#v", audit)
			}
		}
	})

	t.Run("draft version must match exactly before persistence", func(t *testing.T) {
		draft := workflowHTTPToolEligibleDraftForTest()
		service, store, _, _, _ := newWorkflowHTTPToolActionServiceForTest(t, &draft)
		result := service.CreatePlan(workflowHTTPToolActionTestContext(), WorkflowHTTPToolCreatePlanRequest{
			DraftID: draft.DraftID, DraftVersion: draft.DraftVersion + 1, NodeID: "node_http_tool",
			PublicArguments: map[string]any{"resource_key": "docs/radishflow/overview"},
		})
		if result.FailureCode != WorkflowHTTPToolActionFailureDraftVersion || result.ActionPlan != nil {
			t.Fatalf("version drift must fail closed: %#v", result)
		}
		if len(store.plans) != 0 || len(store.audits) != 0 {
			t.Fatalf("version mismatch left partial persistence: plans=%d audits=%d", len(store.plans), len(store.audits))
		}
	})

	t.Run("registry profile and public argument gates fail before persistence", func(t *testing.T) {
		tests := []struct {
			name     string
			mutate   func(*workflowHTTPToolActionService)
			argument string
			expected WorkflowHTTPToolActionFailureCode
		}{
			{
				name: "definition missing", mutate: func(service *workflowHTTPToolActionService) {
					service.registry.definition.ToolID = ""
				}, argument: "docs/radishflow/overview", expected: WorkflowHTTPToolActionFailureDefinitionMissing,
			},
			{
				name: "tool version missing", mutate: func(service *workflowHTTPToolActionService) {
					service.registry.definition.ToolVersion = 0
				}, argument: "docs/radishflow/overview", expected: WorkflowHTTPToolActionFailureDefinitionMissing,
			},
			{
				name: "profile disabled", mutate: func(service *workflowHTTPToolActionService) {
					service.registry.profile.Enabled = false
				}, argument: "docs/radishflow/overview", expected: WorkflowHTTPToolActionFailureProfileUnavailable,
			},
			{
				name: "profile binding drift", mutate: func(service *workflowHTTPToolActionService) {
					service.registry.profile.DefinitionDigest = "sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
				}, argument: "docs/radishflow/overview", expected: WorkflowHTTPToolActionFailureProfileUnavailable,
			},
			{
				name: "absolute URL argument", mutate: func(_ *workflowHTTPToolActionService) {},
				argument: "https://service.invalid/resource", expected: WorkflowHTTPToolActionFailureInputInvalid,
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				draft := workflowHTTPToolEligibleDraftForTest()
				service, store, _, _, _ := newWorkflowHTTPToolActionServiceForTest(t, &draft)
				test.mutate(&service)
				result := service.CreatePlan(workflowHTTPToolActionTestContext(), WorkflowHTTPToolCreatePlanRequest{
					DraftID: draft.DraftID, DraftVersion: draft.DraftVersion, NodeID: "node_http_tool",
					PublicArguments: map[string]any{"resource_key": test.argument},
				})
				if result.FailureCode != test.expected || result.ActionPlan != nil {
					t.Fatalf("gate returned an unexpected result: %#v", result)
				}
				if len(store.plans) != 0 || len(store.decisions) != 0 || len(store.audits) != 0 {
					t.Fatalf("failed gate left persistence: plans=%d decisions=%d audits=%d", len(store.plans), len(store.decisions), len(store.audits))
				}
			})
		}
	})

	t.Run("expired plan transitions through a system decision", func(t *testing.T) {
		draft := workflowHTTPToolEligibleDraftForTest()
		service, _, _, _, clock := newWorkflowHTTPToolActionServiceForTest(t, &draft)
		ctx := workflowHTTPToolActionTestContext()
		created := service.CreatePlan(ctx, WorkflowHTTPToolCreatePlanRequest{
			DraftID: draft.DraftID, DraftVersion: draft.DraftVersion, NodeID: "node_http_tool",
			PublicArguments: map[string]any{"resource_key": "docs/radishflow/overview"},
		})
		*clock = clock.Add(workflowHTTPToolDefaultExpiry + time.Second)
		read := service.ReadPlan(ctx, created.ActionPlan.PlanID)
		if read.FailureCode != "" || read.ActionPlan == nil || read.ConfirmationDecision == nil ||
			read.ActionPlan.Status != WorkflowHTTPToolActionStatusExpired ||
			read.ConfirmationDecision.Outcome != workflowHTTPToolConfirmationExpire ||
			read.ConfirmationDecision.ActorSource != "system" {
			t.Fatalf("expired plan did not record a system transition: %#v", read)
		}
	})

	t.Run("saved draft drift invalidates the existing digest", func(t *testing.T) {
		draft := workflowHTTPToolEligibleDraftForTest()
		service, _, _, _, _ := newWorkflowHTTPToolActionServiceForTest(t, &draft)
		ctx := workflowHTTPToolActionTestContext()
		created := service.CreatePlan(ctx, WorkflowHTTPToolCreatePlanRequest{
			DraftID: draft.DraftID, DraftVersion: draft.DraftVersion, NodeID: "node_http_tool",
			PublicArguments: map[string]any{"resource_key": "docs/radishflow/overview"},
		})
		draft.Nodes[1].ToolRef = "workflow.http.changed-definition.v1"
		read := service.ReadPlan(ctx, created.ActionPlan.PlanID)
		if read.FailureCode != "" || read.ActionPlan == nil || read.ConfirmationDecision == nil ||
			read.ActionPlan.Status != WorkflowHTTPToolActionStatusInvalidated ||
			read.ConfirmationDecision.Outcome != workflowHTTPToolConfirmationInvalidate ||
			read.ConfirmationDecision.ActorSource != "system" {
			t.Fatalf("source drift did not invalidate the plan: %#v", read)
		}
	})

	t.Run("current registry drift invalidates a structurally valid stored plan", func(t *testing.T) {
		draft := workflowHTTPToolEligibleDraftForTest()
		service, _, _, _, _ := newWorkflowHTTPToolActionServiceForTest(t, &draft)
		ctx := workflowHTTPToolActionTestContext()
		created := service.CreatePlan(ctx, WorkflowHTTPToolCreatePlanRequest{
			DraftID: draft.DraftID, DraftVersion: draft.DraftVersion, NodeID: "node_http_tool",
			PublicArguments: map[string]any{"resource_key": "docs/radishflow/overview"},
		})
		if created.FailureCode != "" || created.ActionPlan == nil {
			t.Fatalf("create plan before registry drift: %#v", created)
		}
		service.registry.profile.TimeoutMS--
		var err error
		service.registry.profileDigest, err = canonicalSHA256(service.registry.profile)
		if err != nil {
			t.Fatalf("digest drifted registry profile: %v", err)
		}
		read := service.ReadPlan(ctx, created.ActionPlan.PlanID)
		if read.FailureCode != "" || read.ActionPlan == nil || read.ConfirmationDecision == nil ||
			read.ActionPlan.Status != WorkflowHTTPToolActionStatusInvalidated ||
			read.ConfirmationDecision.Outcome != workflowHTTPToolConfirmationInvalidate {
			t.Fatalf("registry drift did not record an invalidation: %#v", read)
		}
	})

	t.Run("independent confirmer rechecks the original draft owner", func(t *testing.T) {
		draft := workflowHTTPToolEligibleDraftForTest()
		service, _, _, _, _ := newWorkflowHTTPToolActionServiceForTest(t, &draft)
		plannerContext := workflowHTTPToolActionTestContext()
		created := service.CreatePlan(plannerContext, WorkflowHTTPToolCreatePlanRequest{
			DraftID: draft.DraftID, DraftVersion: draft.DraftVersion, NodeID: "node_http_tool",
			PublicArguments: map[string]any{"resource_key": "docs/radishflow/overview"},
		})
		if created.FailureCode != "" || created.ActionPlan == nil {
			t.Fatalf("planner could not create action plan: %#v", created)
		}
		confirmerContext := plannerContext
		confirmerContext.ActorRef = "subject_independent_confirmer"
		confirmerContext.ScopeGrants = []string{"workflow_tool_actions:confirm"}
		approved := service.DecidePlan(confirmerContext, WorkflowHTTPToolDecisionRequest{
			PlanID: created.ActionPlan.PlanID, ExpectedRecordVersion: 1, Decision: WorkflowHTTPToolConfirmationApprove,
		})
		if approved.FailureCode != "" || approved.ActionPlan == nil || approved.ConfirmationDecision == nil ||
			approved.ActionPlan.Status != WorkflowHTTPToolActionStatusApproved ||
			approved.ConfirmationDecision.DecidedByActorRef != confirmerContext.ActorRef {
			t.Fatalf("independent confirmer must approve against the original owner source: %#v", approved)
		}
	})

	t.Run("transient source read failure preserves the healthy plan", func(t *testing.T) {
		draft := workflowHTTPToolEligibleDraftForTest()
		service, _, _, _, _ := newWorkflowHTTPToolActionServiceForTest(t, &draft)
		ctx := workflowHTTPToolActionTestContext()
		created := service.CreatePlan(ctx, WorkflowHTTPToolCreatePlanRequest{
			DraftID: draft.DraftID, DraftVersion: draft.DraftVersion, NodeID: "node_http_tool",
			PublicArguments: map[string]any{"resource_key": "docs/radishflow/overview"},
		})
		originalRead := service.readDraft
		service.readDraft = func(_ SavedWorkflowDraftContext, _ ReadWorkflowDraftRequest) SavedWorkflowDraftResult {
			return SavedWorkflowDraftResult{FailureCode: SavedWorkflowDraftFailureStoreUnavailable}
		}
		failedRead := service.ReadPlan(ctx, created.ActionPlan.PlanID)
		if failedRead.FailureCode != WorkflowHTTPToolActionFailureStoreUnavailable || failedRead.ActionPlan != nil {
			t.Fatalf("transient source failure must fail closed without a projection: %#v", failedRead)
		}
		service.readDraft = originalRead
		recovered := service.ReadPlan(ctx, created.ActionPlan.PlanID)
		if recovered.FailureCode != "" || recovered.ActionPlan == nil ||
			recovered.ActionPlan.Status != WorkflowHTTPToolActionStatusPending || recovered.ActionPlan.RecordVersion != 1 {
			t.Fatalf("transient source failure must not invalidate the plan: %#v", recovered)
		}
	})

	t.Run("source configuration failures never become permanent drift", func(t *testing.T) {
		tests := []struct {
			name     string
			source   SavedWorkflowDraftFailureCode
			expected WorkflowHTTPToolActionFailureCode
		}{
			{"repository disabled", SavedWorkflowDraftFailureRepositoryStoreDisabled, WorkflowHTTPToolActionFailureStoreUnavailable},
			{"write gate disabled", SavedWorkflowDraftFailureWriteDisabled, WorkflowHTTPToolActionFailureStoreUnavailable},
			{"invalid store mode", SavedWorkflowDraftFailureInvalidStoreMode, WorkflowHTTPToolActionFailureStoreContract},
			{"unknown source contract", SavedWorkflowDraftFailureCode("draft_unknown_failure"), WorkflowHTTPToolActionFailureStoreContract},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				draft := workflowHTTPToolEligibleDraftForTest()
				service, store, _, _, _ := newWorkflowHTTPToolActionServiceForTest(t, &draft)
				ctx := workflowHTTPToolActionTestContext()
				created := service.CreatePlan(ctx, WorkflowHTTPToolCreatePlanRequest{
					DraftID: draft.DraftID, DraftVersion: draft.DraftVersion, NodeID: "node_http_tool",
					PublicArguments: map[string]any{"resource_key": "docs/radishflow/overview"},
				})
				originalRead := service.readDraft
				service.readDraft = func(_ SavedWorkflowDraftContext, _ ReadWorkflowDraftRequest) SavedWorkflowDraftResult {
					return SavedWorkflowDraftResult{FailureCode: test.source}
				}
				failedRead := service.ReadPlan(ctx, created.ActionPlan.PlanID)
				if failedRead.FailureCode != test.expected || failedRead.ActionPlan != nil {
					t.Fatalf("source configuration failure was misclassified: %#v", failedRead)
				}
				if len(store.decisions) != 0 || len(store.audits) != 1 {
					t.Fatalf("source configuration failure appended drift evidence: decisions=%d audits=%d", len(store.decisions), len(store.audits))
				}
				service.readDraft = originalRead
				recovered := service.ReadPlan(ctx, created.ActionPlan.PlanID)
				if recovered.FailureCode != "" || recovered.ActionPlan == nil ||
					recovered.ActionPlan.Status != WorkflowHTTPToolActionStatusPending || recovered.ActionPlan.RecordVersion != 1 {
					t.Fatalf("source configuration failure invalidated the plan: %#v", recovered)
				}
			})
		}
	})

	t.Run("a deferred plan cannot be deferred repeatedly", func(t *testing.T) {
		draft := workflowHTTPToolEligibleDraftForTest()
		service, store, _, _, _ := newWorkflowHTTPToolActionServiceForTest(t, &draft)
		ctx := workflowHTTPToolActionTestContext()
		created := service.CreatePlan(ctx, WorkflowHTTPToolCreatePlanRequest{
			DraftID: draft.DraftID, DraftVersion: draft.DraftVersion, NodeID: "node_http_tool",
			PublicArguments: map[string]any{"resource_key": "docs/radishflow/overview"},
		})
		deferred := service.DecidePlan(ctx, WorkflowHTTPToolDecisionRequest{
			PlanID: created.ActionPlan.PlanID, ExpectedRecordVersion: 1, Decision: WorkflowHTTPToolConfirmationDefer,
		})
		if deferred.FailureCode != "" || deferred.ActionPlan == nil || deferred.ActionPlan.Status != WorkflowHTTPToolActionStatusDeferred {
			t.Fatalf("defer pending plan: %#v", deferred)
		}
		repeated := service.DecidePlan(ctx, WorkflowHTTPToolDecisionRequest{
			PlanID: deferred.ActionPlan.PlanID, ExpectedRecordVersion: 2, Decision: WorkflowHTTPToolConfirmationDefer,
		})
		if repeated.FailureCode != WorkflowHTTPToolActionFailureConflict || repeated.ActionPlan != nil {
			t.Fatalf("repeated defer must conflict: %#v", repeated)
		}
		if len(store.decisions) != 1 || len(store.audits) != 2 {
			t.Fatalf("repeated defer appended evidence: decisions=%d audits=%d", len(store.decisions), len(store.audits))
		}
	})

	t.Run("decision uses one timestamp across expiry validation and persistence", func(t *testing.T) {
		draft := workflowHTTPToolEligibleDraftForTest()
		service, _, _, _, clock := newWorkflowHTTPToolActionServiceForTest(t, &draft)
		ctx := workflowHTTPToolActionTestContext()
		created := service.CreatePlan(ctx, WorkflowHTTPToolCreatePlanRequest{
			DraftID: draft.DraftID, DraftVersion: draft.DraftVersion, NodeID: "node_http_tool",
			PublicArguments: map[string]any{"resource_key": "docs/radishflow/overview"},
		})
		expiresAt, err := time.Parse(time.RFC3339Nano, created.ActionPlan.ExpiresAt)
		if err != nil {
			t.Fatalf("parse plan expiry: %v", err)
		}
		calls := 0
		service.now = func() time.Time {
			calls++
			if calls == 1 {
				return expiresAt.Add(-time.Nanosecond)
			}
			return expiresAt.Add(time.Second)
		}
		approved := service.DecidePlan(ctx, WorkflowHTTPToolDecisionRequest{
			PlanID: created.ActionPlan.PlanID, ExpectedRecordVersion: 1, Decision: WorkflowHTTPToolConfirmationApprove,
		})
		if approved.FailureCode != "" || approved.ActionPlan == nil || approved.ActionPlan.Status != WorkflowHTTPToolActionStatusApproved {
			t.Fatalf("single pre-expiry timestamp should approve: %#v", approved)
		}
		if calls != 1 {
			t.Fatalf("decision path sampled time %d times; expected one", calls)
		}
		*clock = expiresAt
	})
}

func newWorkflowHTTPToolActionServiceForTest(
	t *testing.T,
	draft *SavedWorkflowDraft,
) (workflowHTTPToolActionService, *memoryWorkflowHTTPToolActionStore, *memoryWorkflowRunStore, *int, *time.Time) {
	t.Helper()
	runStore := newMemoryWorkflowRunStore(8)
	store := newMemoryWorkflowHTTPToolActionStore(&runStore.mu)
	readCalls := 0
	service, err := newWorkflowHTTPToolActionService(
		func(context SavedWorkflowDraftContext, request ReadWorkflowDraftRequest) SavedWorkflowDraftResult {
			readCalls++
			if request.DraftID != draft.DraftID || context.OwnerSubjectRef != "subject_demo_user" ||
				!controlPlaneReadHasScope(context.ScopeGrants, "workflow_drafts:read") {
				return SavedWorkflowDraftResult{FailureCode: SavedWorkflowDraftFailureNotFound}
			}
			return SavedWorkflowDraftResult{Draft: cloneSavedWorkflowDraftPointer(*draft)}
		},
		store,
	)
	if err != nil {
		t.Fatalf("create workflow HTTP tool action service: %v", err)
	}
	clock := time.Date(2026, 7, 16, 9, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return clock }
	nextID := 0
	service.newID = func(prefix string) (string, error) {
		nextID++
		return fmt.Sprintf("%s%024d", prefix, nextID), nil
	}
	return service, store, runStore, &readCalls, &clock
}

func workflowHTTPToolEligibleDraftForTest() SavedWorkflowDraft {
	node := func(id, nodeType string) SavedWorkflowDraftNode {
		return SavedWorkflowDraftNode{
			NodeID: id, NodeType: nodeType, Label: id,
			InputSummary: "sanitized input", OutputSummary: "sanitized output",
			InputContractRef: "contract_input", OutputContractRef: "contract_output",
			InputContractFields: []string{"input"}, OutputContractFields: []string{"output"},
			OutputMappingSummary: "map sanitized fields", RiskLevel: "low",
		}
	}
	prompt := node("node_prompt", "prompt")
	httpTool := node("node_http_tool", "http_tool")
	httpTool.ToolRef = workflowHTTPToolID
	httpTool.RiskLevel = "medium"
	httpTool.RequiresConfirmation = true
	model := node("node_model", "llm")
	model.ProviderRef = "profile:mock-workflow"
	output := node("node_output", "output")
	return SavedWorkflowDraft{
		DraftID: "draft_workflow_http_tool_test", WorkspaceID: "workspace_demo", ApplicationID: "app_flow_copilot",
		SourceDefinitionID: "workflow_http_tool_test", BaseDefinitionVersion: 1, DraftVersion: 1,
		SchemaVersion: savedWorkflowDraftSchemaVersion, DraftStatus: SavedWorkflowDraftStatusValidForReview,
		CreatedAt: "2026-07-16T09:00:00Z", UpdatedAt: "2026-07-16T09:00:00Z",
		CreatedByActorRef: "subject_platform_ops", UpdatedByActorRef: "subject_platform_ops",
		Name: "Reviewed HTTP tool workflow", Description: "A linear, reviewable HTTP tool workflow.",
		Nodes: []SavedWorkflowDraftNode{prompt, httpTool, model, output},
		Edges: []SavedWorkflowDraftEdge{
			{EdgeID: "edge_prompt_tool", FromNodeID: prompt.NodeID, ToNodeID: httpTool.NodeID, ConditionSummary: "prompt output is ready"},
			{EdgeID: "edge_tool_model", FromNodeID: httpTool.NodeID, ToNodeID: model.NodeID, ConditionSummary: "reviewed tool projection is ready"},
			{EdgeID: "edge_model_output", FromNodeID: model.NodeID, ToNodeID: output.NodeID, ConditionSummary: "model output is ready"},
		},
		InputContract:         SavedWorkflowDraftContract{ContractID: "contract_input", RequiredFields: []string{"input"}, Summary: "sanitized input"},
		OutputContract:        SavedWorkflowDraftContract{ContractID: "contract_output", RequiredFields: []string{"output"}, Summary: "reviewable output"},
		ProviderRefs:          []string{"profile:mock-workflow"},
		ToolRefs:              []string{workflowHTTPToolID},
		RequestedCapabilities: []string{"publish", "run", "confirmation_decision", "writeback", "replay"},
		ValidationSummary: SavedWorkflowDraftValidationSummary{
			ValidationState: SavedWorkflowDraftStatusBlockedCapability,
			ValidForReview:  false,
			Findings: []SavedWorkflowDraftValidationFinding{
				{Code: SavedWorkflowDraftFailureBlockedCapability, Severity: SavedWorkflowDraftValidationBlocking, Field: "requested_capabilities", Summary: "review-only capability", EvidenceID: "publish"},
				{Code: SavedWorkflowDraftFailureBlockedCapability, Severity: SavedWorkflowDraftValidationBlocking, Field: "requested_capabilities", Summary: "review-only capability", EvidenceID: "run"},
				{Code: SavedWorkflowDraftFailureBlockedCapability, Severity: SavedWorkflowDraftValidationBlocking, Field: "requested_capabilities", Summary: "review-only capability", EvidenceID: "confirmation_decision"},
				{Code: SavedWorkflowDraftFailureBlockedCapability, Severity: SavedWorkflowDraftValidationBlocking, Field: "requested_capabilities", Summary: "review-only capability", EvidenceID: "writeback"},
				{Code: SavedWorkflowDraftFailureBlockedCapability, Severity: SavedWorkflowDraftValidationBlocking, Field: "requested_capabilities", Summary: "review-only capability", EvidenceID: "replay"},
			},
		},
		SampleOrUnsavedDraftStatus: "saved_draft_record",
	}
}

func workflowHTTPToolActionTestContext() WorkflowHTTPToolActionContext {
	return WorkflowHTTPToolActionContext{
		RequestContext: context.Background(), RequestID: "request_workflow_http_tool_test",
		TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", ApplicationID: "app_flow_copilot",
		ActorRef: "subject_demo_user", ScopeGrants: []string{"workflow_tool_actions:plan", "workflow_tool_actions:read", "workflow_tool_actions:confirm"},
		AuditRef: "audit_workflow_http_tool_test",
	}
}

func workflowHTTPToolActionContextWithScope(base WorkflowHTTPToolActionContext, tenant, workspace, application string) WorkflowHTTPToolActionContext {
	base.TenantRef = tenant
	base.WorkspaceID = workspace
	base.ApplicationID = application
	return base
}

func workflowHTTPToolActionPlanForStoreTest(t *testing.T, ctx WorkflowHTTPToolActionContext, planID string) WorkflowHTTPToolActionPlan {
	t.Helper()
	registry, err := newWorkflowHTTPToolRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	plan := WorkflowHTTPToolActionPlan{
		SchemaVersion: workflowHTTPToolPlanSchema, PlanID: planID, RecordVersion: 1,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		DraftID: "draft_workflow_http_tool_test", DraftVersion: 1, NodeID: "node_http_tool",
		ToolID: workflowHTTPToolID, ToolVersion: workflowHTTPToolVersion, DefinitionDigest: registry.definitionDigest,
		ProfileID: registry.profile.ProfileID, ProfileVersion: registry.profile.ProfileVersion, ProfileDigest: registry.profileDigest,
		Method: registry.profile.Method, TargetPolicyKey: registry.profile.TargetPolicyKey,
		PublicArguments: WorkflowHTTPToolPublicArguments{ResourceKey: "docs/radishflow/overview", Locale: "zh-CN"},
		OutputFields:    append([]string(nil), registry.definition.OutputFields...), OutputSchemaDigest: registry.outputSchemaDigest,
		CredentialPolicy: registry.profile.CredentialPolicy,
		TimeoutMS:        registry.profile.TimeoutMS, MaxResponseBytes: registry.profile.MaxResponseBytes, MaxOutputBytes: registry.profile.MaxOutputBytes,
		PlannedByActorRef: ctx.ActorRef, CreatedAt: "2026-07-16T09:00:00Z", ExpiresAt: "2026-07-16T09:15:00Z",
		Status: WorkflowHTTPToolActionStatusPending, AuditRef: ctx.AuditRef,
	}
	plan.ToolPlanDigest, err = workflowHTTPToolPlanDigest(plan)
	if err != nil {
		t.Fatalf("digest test plan: %v", err)
	}
	return plan
}

func workflowHTTPToolDecisionForStoreTest(
	plan WorkflowHTTPToolActionPlan,
	confirmationID string,
	outcome WorkflowHTTPToolConfirmationOutcome,
	actor string,
) WorkflowHTTPToolConfirmationDecision {
	return WorkflowHTTPToolConfirmationDecision{
		SchemaVersion: workflowHTTPToolDecisionSchema, ConfirmationID: confirmationID, PlanID: plan.PlanID,
		TenantRef: plan.TenantRef, WorkspaceID: plan.WorkspaceID, ApplicationID: plan.ApplicationID,
		DraftID: plan.DraftID, DraftVersion: plan.DraftVersion, NodeID: plan.NodeID, ToolID: plan.ToolID, ToolVersion: plan.ToolVersion,
		ToolPlanDigest: plan.ToolPlanDigest, Outcome: outcome, DecidedByActorRef: actor, ActorSource: "human",
		DecidedAt: "2026-07-16T09:01:00Z", ReasonCode: workflowHTTPToolConfirmationReason(outcome),
		ExpectedRecordVersion: plan.RecordVersion - 1, ResultingRecordVersion: plan.RecordVersion, AuditRef: plan.AuditRef,
	}
}

func workflowHTTPToolSystemDecisionForStoreTest(
	plan WorkflowHTTPToolActionPlan,
	confirmationID string,
	outcome WorkflowHTTPToolConfirmationOutcome,
) WorkflowHTTPToolConfirmationDecision {
	decision := workflowHTTPToolDecisionForStoreTest(plan, confirmationID, outcome, "system:workflow_tool_action_policy")
	decision.ActorSource = "system"
	return decision
}

func workflowHTTPToolAuditForStoreTest(plan WorkflowHTTPToolActionPlan, eventID, eventKind string, confirmationIDs ...string) WorkflowHTTPToolExecutionAudit {
	actorRef, actorSource, occurredAt := plan.PlannedByActorRef, "human", plan.CreatedAt
	if plan.LastDecisionByActorRef != nil && plan.LastDecisionAt != nil {
		actorRef, occurredAt = *plan.LastDecisionByActorRef, *plan.LastDecisionAt
		if actorRef == "system:workflow_tool_action_policy" {
			actorSource = "system"
		}
	}
	var confirmationID *string
	if len(confirmationIDs) == 1 {
		value := confirmationIDs[0]
		confirmationID = &value
	}
	return WorkflowHTTPToolExecutionAudit{
		SchemaVersion: workflowHTTPToolAuditSchema, EventID: eventID, EventKind: eventKind,
		OccurredAt: occurredAt, TenantRef: plan.TenantRef, WorkspaceID: plan.WorkspaceID,
		ApplicationID: plan.ApplicationID, DraftID: plan.DraftID, DraftVersion: plan.DraftVersion,
		NodeID: plan.NodeID, PlanID: plan.PlanID, ToolID: plan.ToolID, ToolVersion: plan.ToolVersion, DefinitionDigest: plan.DefinitionDigest,
		ProfileID: plan.ProfileID, ProfileDigest: plan.ProfileDigest, ToolPlanDigest: plan.ToolPlanDigest,
		ConfirmationID: confirmationID, ActorRef: actorRef, ActorSource: actorSource, RequestID: "request_workflow_http_tool_test",
		AuditRef: plan.AuditRef, AttemptStatus: "not_claimed", SideEffects: WorkflowHTTPToolAuditSideEffects{},
	}
}
