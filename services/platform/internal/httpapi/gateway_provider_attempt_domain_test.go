package httpapi

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/bridge"
)

func TestAdminProviderRouteV2LifecycleAndV1Compatibility(t *testing.T) {
	service, _, _ := newAdminProviderRouteTestService()
	ctx := adminProviderRouteTestContext()
	v2Input := adminProviderRouteV2TestDraftInput(0)
	draft := service.PutDraft(ctx, v2Input)
	if draft.FailureCode != "" || draft.Draft == nil || draft.Draft.SchemaVersion != adminProviderRouteDraftSchemaVersionV2 {
		t.Fatalf("create v2 draft: %#v", draft)
	}
	candidate := service.CreateCandidate(ctx, AdminProviderRouteCandidateInput{
		ConfigurationID: v2Input.ConfigurationID, CandidateID: "candidate-fallback", ExpectedDraftRevision: 1,
	})
	if candidate.FailureCode != "" || candidate.Candidate == nil ||
		candidate.Candidate.SchemaVersion != adminProviderRouteCandidateSchemaVersionV2 {
		t.Fatalf("create v2 candidate: %#v", candidate)
	}
	approved := service.ReviewCandidate(ctx, v2Input.ConfigurationID, "candidate-fallback", AdminProviderRouteReviewInput{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: "Reviewed both provider attempt targets.",
	})
	if approved.FailureCode != "" {
		t.Fatalf("approve v2 candidate: %#v", approved)
	}
	active := service.Activate(ctx, AdminProviderRouteActivationInput{
		ConfigurationID: v2Input.ConfigurationID, CandidateID: "candidate-fallback", ExpectedGeneration: 0,
		Action: "activate", Reason: "Activate the reviewed sequential fallback route.",
	})
	if active.FailureCode != "" || active.Snapshot == nil ||
		active.Snapshot.SchemaVersion != adminProviderRouteSnapshotSchemaVersionV2 ||
		active.Snapshot.Configuration.ModelRoutes[0].ExecutionMode != AdminProviderRouteExecutionSequentialFallback {
		t.Fatalf("activate v2 route: %#v", active)
	}

	v1Service, _, _ := newAdminProviderRouteTestService()
	v1Draft := v1Service.PutDraft(ctx, adminProviderRouteTestDraftInput(0, "mock-primary"))
	if v1Draft.FailureCode != "" || v1Draft.Draft.SchemaVersion != adminProviderRouteDraftSchemaVersion ||
		adminProviderRoutePrimaryProfileID(v1Draft.Draft.ModelRoutes[0]) != "primary" {
		t.Fatalf("v1 route compatibility drifted: %#v", v1Draft)
	}
}

func TestAdminProviderRouteV2RejectsUnsafeTargetShapes(t *testing.T) {
	for name, mutate := range map[string]func(*AdminProviderRouteDraftInput){
		"duplicate profile": func(input *AdminProviderRouteDraftInput) {
			input.ModelRoutes[0].AttemptTargets[1].ProviderProfileID = "primary"
		},
		"ordinal gap": func(input *AdminProviderRouteDraftInput) {
			input.ModelRoutes[0].AttemptTargets[1].Ordinal = 3
		},
		"missing backup": func(input *AdminProviderRouteDraftInput) {
			input.ModelRoutes[0].AttemptTargets = input.ModelRoutes[0].AttemptTargets[:1]
		},
		"mixed v1 and v2": func(input *AdminProviderRouteDraftInput) {
			input.ModelRoutes[0].ProviderProfileID = "primary"
		},
		"mixed contract versions": func(input *AdminProviderRouteDraftInput) {
			input.ModelRoutes = append(input.ModelRoutes, AdminModelRouteDefinition{
				RouteID: "route-responses", Protocol: "responses", ModelID: "mock-model", ProviderProfileID: "primary",
			})
		},
	} {
		t.Run(name, func(t *testing.T) {
			service, _, _ := newAdminProviderRouteTestService()
			input := adminProviderRouteV2TestDraftInput(0)
			mutate(&input)
			result := service.PutDraft(adminProviderRouteTestContext(), input)
			if result.FailureCode != AdminProviderRouteFailurePayloadInvalid || result.Draft != nil {
				t.Fatalf("unsafe v2 route accepted: %#v", result)
			}
		})
	}
}

func TestGatewayProviderAttemptPlanFreezesTargetsAndExplicitFallback(t *testing.T) {
	snapshot, modelRoute := gatewayProviderAttemptV2Snapshot(t)
	targets := gatewayProviderAttemptPlanTestTargets()
	plan, err := buildGatewayProviderAttemptPlan(
		"request-attempt-plan", "/v1/chat/completions", northboundProtocolChatCompletions,
		"mock-model", snapshot, modelRoute, GatewayProviderFallbackAllowConfigured, targets,
	)
	if err != nil || plan.MaxAttempts != 2 || !plan.FallbackAllowed || len(plan.Targets) != 2 ||
		plan.Targets[0].AttemptID != "request-attempt-plan.pa1" || plan.Targets[1].AttemptID != "request-attempt-plan.pa2" {
		t.Fatalf("build fallback plan: plan=%#v err=%v", plan, err)
	}
	targets[0].ProviderID = "mutated"
	if plan.Targets[0].ProviderID != "mock" {
		t.Fatal("plan retained mutable input")
	}
	clone := cloneGatewayProviderAttemptPlan(plan)
	clone.Targets[0].ProviderID = "changed"
	if plan.Targets[0].ProviderID == "changed" {
		t.Fatal("plan clone shared target slice")
	}
	payload, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip GatewayProviderAttemptPlan
	if err := json.Unmarshal(payload, &roundTrip); err != nil || !validGatewayProviderAttemptPlan(roundTrip) {
		t.Fatalf("frozen plan did not survive JSON round trip: plan=%#v err=%v", roundTrip, err)
	}

	disabled, err := buildGatewayProviderAttemptPlan(
		"request-attempt-disabled", "/v1/chat/completions", northboundProtocolChatCompletions,
		"mock-model", snapshot, modelRoute, GatewayProviderFallbackDisabled, gatewayProviderAttemptPlanTestTargets(),
	)
	if err != nil || disabled.FallbackAllowed || disabled.MaxAttempts != 1 || len(disabled.Targets) != 2 {
		t.Fatalf("disabled fallback must retain reviewed plan but execute once: %#v err=%v", disabled, err)
	}
}

func TestGatewayProviderAttemptMemoryHistoryFallbackSuccess(t *testing.T) {
	store := newMemoryGatewayRequestStore(10)
	requestContext := gatewayRequestTestContext()
	plan := gatewayProviderAttemptTestPlan(t, "request-fallback-success")
	record := gatewayRequestTestRecord(requestContext, plan.RootRequestID, time.Date(2026, 8, 13, 8, 0, 0, 0, time.UTC))
	v3, err := newGatewayProviderAttemptHistoryRecord(record, plan)
	if err != nil || store.CreateRequest(requestContext, &v3) != nil {
		t.Fatalf("create v3 root: record=%#v err=%v", v3, err)
	}
	service := newGatewayProviderAttemptHistoryService(store)
	clock := time.Date(2026, 8, 13, 8, 0, 1, 0, time.UTC)
	if _, err := service.StartAttempt(requestContext, plan.RootRequestID, plan.Targets[0], "quota-primary", clock); err != nil {
		t.Fatalf("start primary: %v", err)
	}
	eligible, ok := bridge.NewProviderAttemptFailure(
		bridge.ProviderFailureTemporarilyUnavailable, bridge.ProviderFallbackEligible,
		false, bridge.ProviderAttemptFailed, "PROVIDER_TEMPORARILY_UNAVAILABLE", "5xx",
	)
	if !ok {
		t.Fatal("eligible failure fixture invalid")
	}
	usageMissing := GatewayRequestUsage{Availability: GatewayRequestUsageNotReported}
	costMissing := gatewayRequestCostUnavailable(GatewayRequestCostUsageNotReported, "provider_usage_not_reported")
	primaryDone, err := service.CompleteAttempt(
		requestContext, plan.RootRequestID, plan.Targets[0].AttemptID,
		usageMissing, costMissing, &eligible, true, clock.Add(time.Second),
	)
	if err != nil || primaryDone.ProviderAttemptPhase != GatewayProviderAttemptPhaseFallbackPending {
		t.Fatalf("checkpoint fallback pending: record=%#v err=%v", primaryDone, err)
	}
	substitutedTarget := plan.Targets[1]
	substitutedTarget.ProviderID = "unreviewed-provider"
	if _, err := service.StartAttempt(
		requestContext, plan.RootRequestID, substitutedTarget, "quota-backup", clock.Add(2*time.Second),
	); err != errGatewayRequestStoreContract {
		t.Fatalf("unreviewed fallback target accepted: %v", err)
	}
	if _, err := service.StartAttempt(requestContext, plan.RootRequestID, plan.Targets[1], "quota-backup", clock.Add(2*time.Second)); err != nil {
		t.Fatalf("start backup: %v", err)
	}
	usage := GatewayRequestUsage{
		Availability: GatewayRequestUsageReported, Source: "openai_compatible_usage",
		InputTokens: 2, OutputTokens: 3, TotalTokens: 5,
	}
	cost := buildGatewayRequestCostEstimate(
		true, usage, gatewayModelPricingSnapshotFromAttempt(plan.Targets[1].PricingSnapshot),
	)
	backupDone, err := service.CompleteAttempt(
		requestContext, plan.RootRequestID, plan.Targets[1].AttemptID,
		usage, cost, nil, false, clock.Add(3*time.Second),
	)
	if err != nil || backupDone.ProviderAttemptPhase != GatewayProviderAttemptPhaseTerminalPending ||
		backupDone.ProviderAttemptCostSummary.Coverage != GatewayProviderAttemptCostCoveragePartial {
		t.Fatalf("complete backup: record=%#v err=%v", backupDone, err)
	}
	finished, err := service.Finalize(
		requestContext, plan.RootRequestID, GatewayRequestStatusSucceeded, 200, "", "", clock.Add(4*time.Second),
	)
	if err != nil || finished.Status != GatewayRequestStatusSucceeded || !finished.FallbackUsed ||
		finished.ProviderAttemptCount != 2 || finished.TerminalAttemptID != plan.Targets[1].AttemptID ||
		finished.ProviderAttempts[0].Failure == nil || finished.ProviderAttempts[1].Status != GatewayProviderAttemptStatusSucceeded {
		t.Fatalf("finalize fallback success: record=%#v err=%v", finished, err)
	}

	finished.ProviderAttempts[0].Failure.Code = "MUTATED"
	finished.ProviderAttemptPlan.Targets[0].ProviderID = "MUTATED"
	readBack, found, readErr := store.ReadRequest(requestContext, plan.RootRequestID)
	if readErr != nil || !found || readBack.ProviderAttempts[0].Failure.Code == "MUTATED" ||
		readBack.ProviderAttemptPlan.Targets[0].ProviderID == "MUTATED" {
		t.Fatal("memory history leaked mutable attempt evidence")
	}
}

func TestGatewayProviderAttemptMemoryHistoryFailsClosed(t *testing.T) {
	for name, failure := range map[string]bridge.ProviderAttemptFailure{
		"ineligible": gatewayProviderAttemptTestFailure(t, bridge.ProviderFailureInvalidRequest, bridge.ProviderFallbackIneligible, bridge.ProviderAttemptFailed),
		"unknown":    gatewayProviderAttemptTestFailure(t, bridge.ProviderFailureTemporarilyUnavailable, bridge.ProviderFallbackIneligible, bridge.ProviderAttemptUnknown),
	} {
		t.Run(name, func(t *testing.T) {
			store := newMemoryGatewayRequestStore(10)
			ctx := gatewayRequestTestContext()
			plan := gatewayProviderAttemptTestPlan(t, "request-"+name)
			record := gatewayRequestTestRecord(ctx, plan.RootRequestID, time.Now().UTC())
			v3, err := newGatewayProviderAttemptHistoryRecord(record, plan)
			if err != nil || store.CreateRequest(ctx, &v3) != nil {
				t.Fatalf("create root: %v", err)
			}
			service := newGatewayProviderAttemptHistoryService(store)
			now := time.Now().UTC().Add(time.Second)
			if _, err := service.StartAttempt(ctx, plan.RootRequestID, plan.Targets[0], "quota-primary", now); err != nil {
				t.Fatal(err)
			}
			_, err = service.CompleteAttempt(
				ctx, plan.RootRequestID, plan.Targets[0].AttemptID,
				GatewayRequestUsage{Availability: GatewayRequestUsageNotReported},
				gatewayRequestCostUnavailable(GatewayRequestCostUsageNotReported, "provider_usage_not_reported"),
				&failure, true, now.Add(time.Second),
			)
			if err != errGatewayRequestStoreContract {
				t.Fatalf("unsafe fallback checkpoint accepted: %v", err)
			}
			stored, _, _ := store.ReadRequest(ctx, plan.RootRequestID)
			if stored.ProviderAttemptPhase != GatewayProviderAttemptPhasePrimaryRunning || len(stored.ProviderAttempts) != 1 {
				t.Fatalf("rejected checkpoint mutated durable record: %#v", stored)
			}
			terminalPending, err := service.CompleteAttempt(
				ctx, plan.RootRequestID, plan.Targets[0].AttemptID,
				GatewayRequestUsage{Availability: GatewayRequestUsageNotReported},
				gatewayRequestCostUnavailable(GatewayRequestCostUsageNotReported, "provider_usage_not_reported"),
				&failure, false, now.Add(2*time.Second),
			)
			if err != nil || terminalPending.ProviderAttemptPhase != GatewayProviderAttemptPhaseTerminalPending {
				t.Fatalf("terminal failure checkpoint rejected: record=%#v err=%v", terminalPending, err)
			}
			if _, err := service.Finalize(
				ctx, plan.RootRequestID, GatewayRequestStatusSucceeded, 200, "", "", now.Add(3*time.Second),
			); err != errGatewayRequestStoreContract {
				t.Fatalf("failed attempt finalized as success: %v", err)
			}
		})
	}
}

func TestGatewayProviderAttemptMemoryCheckpointHasOneConcurrentWinner(t *testing.T) {
	store := newMemoryGatewayRequestStore(10)
	ctx := gatewayRequestTestContext()
	plan := gatewayProviderAttemptTestPlan(t, "request-concurrent-attempt")
	record := gatewayRequestTestRecord(ctx, plan.RootRequestID, time.Now().UTC())
	v3, err := newGatewayProviderAttemptHistoryRecord(record, plan)
	if err != nil || store.CreateRequest(ctx, &v3) != nil {
		t.Fatal(err)
	}
	service := newGatewayProviderAttemptHistoryService(store)
	start := make(chan struct{})
	results := make(chan error, 2)
	var wait sync.WaitGroup
	for index := 0; index < 2; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			_, checkpointErr := service.StartAttempt(ctx, plan.RootRequestID, plan.Targets[0], "quota-primary", time.Now().UTC())
			results <- checkpointErr
		}()
	}
	close(start)
	wait.Wait()
	close(results)
	successes := 0
	for result := range results {
		if result == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("expected one checkpoint winner, got %d", successes)
	}
}

func adminProviderRouteV2TestDraftInput(expectedRevision int) AdminProviderRouteDraftInput {
	return AdminProviderRouteDraftInput{
		ConfigurationID: "gateway-fallback", ExpectedRevision: expectedRevision, DisplayName: "Fallback test routing",
		ProviderProfiles: []AdminProviderProfileAssignment{
			adminProviderRouteTestProfile("primary", "mock-primary"),
			adminProviderRouteTestProfile("secondary", "mock-secondary"),
		},
		ModelRoutes: []AdminModelRouteDefinition{{
			RouteID: "route-chat", Protocol: "chat_completions", ModelID: "mock-model",
			ExecutionMode: AdminProviderRouteExecutionSequentialFallback,
			AttemptTargets: []AdminProviderRouteAttemptTarget{
				{Ordinal: 1, ProviderProfileID: "primary"},
				{Ordinal: 2, ProviderProfileID: "secondary"},
			},
		}},
	}
}

func gatewayProviderAttemptV2Snapshot(t *testing.T) (AdminProviderRouteSnapshot, AdminModelRouteDefinition) {
	t.Helper()
	service, _, _ := newAdminProviderRouteTestService()
	ctx := adminProviderRouteTestContext()
	input := adminProviderRouteV2TestDraftInput(0)
	draft := service.PutDraft(ctx, input)
	if draft.FailureCode != "" {
		t.Fatal(draft.FailureCode)
	}
	if result := service.CreateCandidate(ctx, AdminProviderRouteCandidateInput{
		ConfigurationID: input.ConfigurationID, CandidateID: "candidate-plan", ExpectedDraftRevision: 1,
	}); result.FailureCode != "" {
		t.Fatal(result.FailureCode)
	}
	if result := service.ReviewCandidate(ctx, input.ConfigurationID, "candidate-plan", AdminProviderRouteReviewInput{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: "Reviewed deterministic attempt plan targets.",
	}); result.FailureCode != "" {
		t.Fatal(result.FailureCode)
	}
	active := service.Activate(ctx, AdminProviderRouteActivationInput{
		ConfigurationID: input.ConfigurationID, CandidateID: "candidate-plan", ExpectedGeneration: 0,
		Action: "activate", Reason: "Activate deterministic attempt plan fixture.",
	})
	if active.FailureCode != "" || active.Snapshot == nil {
		t.Fatalf("activate plan fixture: %#v", active)
	}
	return *active.Snapshot, active.Snapshot.Configuration.ModelRoutes[0]
}

func gatewayProviderAttemptPlanTestTargets() []GatewayProviderAttemptPlanTargetInput {
	pricingContext := GatewayModelPricingContext{
		TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", Environment: "test",
		ProviderID: "mock", ProfileID: "profile_primary", ModelID: "mock-model",
		ActorRef: "subject_demo", RequestID: "request_pricing", AuditRef: "audit_pricing",
	}
	policy := buildGatewayModelPricingPolicy(pricingContext, GatewayModelPricingPolicyInput{
		Currency: GatewayModelPricingCurrency, InputPriceMicrosPerTokenUnit: 1_000_000,
		OutputPriceMicrosPerTokenUnit: 2_000_000, Reason: "Test deterministic attempt pricing snapshot.",
	}, 1, time.Date(2026, 8, 13, 8, 0, 0, 0, time.UTC))
	snapshot := gatewayModelPricingSnapshotFromPolicy(policy)
	return []GatewayProviderAttemptPlanTargetInput{
		{
			Ordinal: 1, ProviderProfileID: "primary", ProviderID: "mock", RuntimeProfile: "mock-primary",
			SelectedModel: "mock-model", UpstreamModel: "mock-model-primary",
			InventoryDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			PricingSnapshot: snapshot,
		},
		{
			Ordinal: 2, ProviderProfileID: "secondary", ProviderID: "mock", RuntimeProfile: "mock-secondary",
			SelectedModel: "mock-model", UpstreamModel: "mock-model-secondary",
			InventoryDigest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			PricingSnapshot: snapshot,
		},
	}
}

func gatewayProviderAttemptTestPlan(t *testing.T, requestID string) GatewayProviderAttemptPlan {
	t.Helper()
	snapshot, route := gatewayProviderAttemptV2Snapshot(t)
	plan, err := buildGatewayProviderAttemptPlan(
		requestID, "/v1/chat/completions", northboundProtocolChatCompletions,
		"mock-model", snapshot, route, GatewayProviderFallbackAllowConfigured, gatewayProviderAttemptPlanTestTargets(),
	)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func gatewayProviderAttemptTestFailure(
	t *testing.T,
	failureClass bridge.ProviderAttemptFailureClass,
	disposition bridge.ProviderAttemptFallbackDisposition,
	outcome bridge.ProviderAttemptOutcome,
) bridge.ProviderAttemptFailure {
	t.Helper()
	failure, ok := bridge.NewProviderAttemptFailure(
		failureClass, disposition, false, outcome, "PROVIDER_TYPED_FAILURE", "5xx",
	)
	if !ok {
		t.Fatal("invalid typed failure fixture")
	}
	return failure
}
