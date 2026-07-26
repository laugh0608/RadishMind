package httpapi

import (
	"context"
	"sync"
	"testing"
	"time"
)

type fakeAdminProviderInventoryResolver struct {
	mu          sync.RWMutex
	bindings    map[string]AdminProviderInventoryBinding
	unavailable bool
}

func newFakeAdminProviderInventoryResolver() *fakeAdminProviderInventoryResolver {
	resolver := &fakeAdminProviderInventoryResolver{bindings: make(map[string]AdminProviderInventoryBinding)}
	resolver.put(AdminProviderInventoryBinding{
		ProviderID: "mock", RuntimeProfileRef: "ref:radishmind/test/provider-profiles/mock-primary",
		Environment: "test", Capabilities: []string{"messages", "responses", "chat_completions"},
		InventoryDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Enabled: true,
	})
	resolver.put(AdminProviderInventoryBinding{
		ProviderID: "mock", RuntimeProfileRef: "ref:radishmind/test/provider-profiles/mock-secondary",
		Environment: "test", Capabilities: []string{"chat_completions", "messages", "responses"},
		InventoryDigest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Enabled: true,
	})
	return resolver
}

func (resolver *fakeAdminProviderInventoryResolver) ResolveProviderProfile(
	_ context.Context,
	environment string,
	providerID string,
	runtimeProfileRef string,
) (AdminProviderInventoryBinding, error) {
	resolver.mu.RLock()
	defer resolver.mu.RUnlock()
	if resolver.unavailable {
		return AdminProviderInventoryBinding{}, errAdminProviderRouteInventoryUnavailable
	}
	binding, exists := resolver.bindings[environment+"\x00"+providerID+"\x00"+runtimeProfileRef]
	if !exists {
		return AdminProviderInventoryBinding{}, errAdminProviderRouteInventoryNotFound
	}
	return binding, nil
}

func (resolver *fakeAdminProviderInventoryResolver) put(binding AdminProviderInventoryBinding) {
	resolver.mu.Lock()
	defer resolver.mu.Unlock()
	key := binding.Environment + "\x00" + binding.ProviderID + "\x00" + binding.RuntimeProfileRef
	resolver.bindings[key] = binding
}

func (resolver *fakeAdminProviderInventoryResolver) setUnavailable(unavailable bool) {
	resolver.mu.Lock()
	defer resolver.mu.Unlock()
	resolver.unavailable = unavailable
}

func TestAdminProviderRouteFullLifecycleAndRollback(t *testing.T) {
	service, _, _ := newAdminProviderRouteTestService()
	ctx := adminProviderRouteTestContext()

	firstDraft := service.PutDraft(ctx, adminProviderRouteTestDraftInput(0, "mock-primary"))
	if firstDraft.FailureCode != "" || firstDraft.Draft == nil || firstDraft.Draft.DraftRevision != 1 {
		t.Fatalf("create draft: %#v", firstDraft)
	}
	firstCandidate := service.CreateCandidate(ctx, AdminProviderRouteCandidateInput{
		ConfigurationID: "gateway-default", CandidateID: "candidate-one", ExpectedDraftRevision: 1,
	})
	if firstCandidate.FailureCode != "" || firstCandidate.Candidate == nil ||
		firstCandidate.Candidate.CandidateState != adminProviderRouteCandidatePending {
		t.Fatalf("create first candidate: %#v", firstCandidate)
	}
	firstApproval := service.ReviewCandidate(ctx, "gateway-default", "candidate-one", AdminProviderRouteReviewInput{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: "Reviewed provider assignment and route binding.",
	})
	if firstApproval.FailureCode != "" || firstApproval.Candidate == nil ||
		firstApproval.Candidate.CandidateState != adminProviderRouteCandidateApproved {
		t.Fatalf("approve first candidate: %#v", firstApproval)
	}
	beforeActivation := service.ReadActiveSnapshot(ctx, "gateway-default")
	if beforeActivation.FailureCode != "" || beforeActivation.Snapshot != nil || beforeActivation.CurrentGeneration != 0 {
		t.Fatalf("approval changed runtime state: %#v", beforeActivation)
	}
	firstActivation := service.Activate(ctx, AdminProviderRouteActivationInput{
		ConfigurationID: "gateway-default", CandidateID: "candidate-one", ExpectedGeneration: 0,
		Action: "activate", Reason: "Enable the reviewed test route configuration.",
	})
	if firstActivation.FailureCode != "" || firstActivation.Snapshot == nil ||
		firstActivation.Snapshot.Generation != 1 || firstActivation.Activation == nil ||
		firstActivation.Activation.Action != adminProviderRouteActionActivate {
		t.Fatalf("activate first candidate: %#v", firstActivation)
	}

	secondDraftInput := adminProviderRouteTestDraftInput(1, "mock-secondary")
	secondDraftInput.DisplayName = "Secondary test routing"
	secondDraft := service.PutDraft(ctx, secondDraftInput)
	if secondDraft.FailureCode != "" || secondDraft.Draft.DraftRevision != 2 {
		t.Fatalf("update draft: %#v", secondDraft)
	}
	secondCandidate := service.CreateCandidate(ctx, AdminProviderRouteCandidateInput{
		ConfigurationID: "gateway-default", CandidateID: "candidate-two", ExpectedDraftRevision: 2,
	})
	if secondCandidate.FailureCode != "" {
		t.Fatalf("create second candidate: %#v", secondCandidate)
	}
	secondApproval := service.ReviewCandidate(ctx, "gateway-default", "candidate-two", AdminProviderRouteReviewInput{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: "Reviewed the replacement profile and route.",
	})
	if secondApproval.FailureCode != "" {
		t.Fatalf("approve second candidate: %#v", secondApproval)
	}
	secondActivation := service.Activate(ctx, AdminProviderRouteActivationInput{
		ConfigurationID: "gateway-default", CandidateID: "candidate-two", ExpectedGeneration: 1,
		Action: "activate", Reason: "Switch new requests to the reviewed replacement.",
	})
	if secondActivation.FailureCode != "" || secondActivation.Snapshot.Generation != 2 {
		t.Fatalf("activate second candidate: %#v", secondActivation)
	}
	rollback := service.Activate(ctx, AdminProviderRouteActivationInput{
		ConfigurationID: "gateway-default", CandidateID: "candidate-one", ExpectedGeneration: 2,
		Action: "rollback", Reason: "Restore the previously reviewed test configuration.",
	})
	if rollback.FailureCode != "" || rollback.Snapshot == nil || rollback.Snapshot.Generation != 3 ||
		rollback.Snapshot.CandidateID != "candidate-one" || rollback.Activation.Action != adminProviderRouteActionRollback {
		t.Fatalf("rollback first candidate: %#v", rollback)
	}
	history := service.ListActivations(ctx, "gateway-default")
	if history.FailureCode != "" || len(history.ActivationHistory) != 3 {
		t.Fatalf("activation history: %#v", history)
	}
	for index, expected := range []string{"candidate-one", "candidate-two", "candidate-one"} {
		record := history.ActivationHistory[index]
		if record.AfterGeneration != index+1 || record.AfterCandidateID != expected {
			t.Fatalf("history[%d]: %#v", index, record)
		}
	}

	rollback.Snapshot.Configuration.ProviderProfiles[0].DisplayName = "mutated outside repository"
	readBack := service.ReadActiveSnapshot(ctx, "gateway-default")
	if readBack.Snapshot.Configuration.ProviderProfiles[0].DisplayName == "mutated outside repository" {
		t.Fatal("active snapshot leaked mutable repository state")
	}
}

func TestAdminProviderRouteCanonicalDraftIsDeterministic(t *testing.T) {
	service, _, _ := newAdminProviderRouteTestService()
	ctx := adminProviderRouteTestContext()
	firstInput := AdminProviderRouteDraftInput{
		ConfigurationID: "gateway-deterministic", ExpectedRevision: 0, DisplayName: "Deterministic routing",
		ProviderProfiles: []AdminProviderProfileAssignment{
			{
				ProfileID: "secondary", DisplayName: "Secondary profile", ProviderID: "mock",
				RuntimeProfileRef: "ref:radishmind/test/provider-profiles/mock-secondary",
				Capabilities:      []string{"responses", "chat_completions"},
			},
			adminProviderRouteTestProfile("primary", "mock-primary"),
		},
		ModelRoutes: []AdminModelRouteDefinition{
			{RouteID: "route-responses", Protocol: "responses", ModelID: "mock-model", ProviderProfileID: "secondary"},
			{RouteID: "route-chat", Protocol: "chat_completions", ModelID: "mock-model", ProviderProfileID: "primary"},
		},
	}
	first := service.PutDraft(ctx, firstInput)
	if first.FailureCode != "" {
		t.Fatalf("first deterministic draft: %#v", first)
	}
	secondInput := firstInput
	secondInput.ExpectedRevision = 1
	secondInput.ProviderProfiles = []AdminProviderProfileAssignment{firstInput.ProviderProfiles[1], firstInput.ProviderProfiles[0]}
	secondInput.ProviderProfiles[0].Capabilities = []string{"messages", "chat_completions", "responses"}
	secondInput.ModelRoutes = []AdminModelRouteDefinition{firstInput.ModelRoutes[1], firstInput.ModelRoutes[0]}
	second := service.PutDraft(ctx, secondInput)
	if second.FailureCode != "" || second.Draft.DraftDigest != first.Draft.DraftDigest {
		t.Fatalf("canonical digest drifted: first=%#v second=%#v", first, second)
	}
	if second.Draft.ProviderProfiles[0].ProfileID != "primary" ||
		second.Draft.ModelRoutes[0].Protocol != "chat_completions" {
		t.Fatalf("canonical ordering not applied: %#v", second.Draft)
	}
}

func TestAdminProviderRouteDraftCASAllowsOneConcurrentUpdate(t *testing.T) {
	service, _, _ := newAdminProviderRouteTestService()
	ctx := adminProviderRouteTestContext()
	if result := service.PutDraft(ctx, adminProviderRouteTestDraftInput(0, "mock-primary")); result.FailureCode != "" {
		t.Fatalf("seed draft: %#v", result)
	}
	const workers = 16
	start := make(chan struct{})
	results := make(chan AdminProviderRouteResult, workers)
	var waitGroup sync.WaitGroup
	for index := 0; index < workers; index++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			results <- service.PutDraft(ctx, adminProviderRouteTestDraftInput(1, "mock-secondary"))
		}()
	}
	close(start)
	waitGroup.Wait()
	close(results)
	successes := 0
	conflicts := 0
	for result := range results {
		switch result.FailureCode {
		case "":
			successes++
		case AdminProviderRouteFailureDraftRevisionConflict:
			conflicts++
			if result.CurrentDraftRevision != 2 {
				t.Fatalf("unexpected current revision: %#v", result)
			}
		default:
			t.Fatalf("unexpected concurrent draft result: %#v", result)
		}
	}
	if successes != 1 || conflicts != workers-1 {
		t.Fatalf("successes=%d conflicts=%d", successes, conflicts)
	}
}

func TestAdminProviderRouteReviewAndActivationCAS(t *testing.T) {
	service, _, _ := newAdminProviderRouteTestService()
	ctx := adminProviderRouteTestContext()
	adminProviderRoutePrepareCandidate(t, service, ctx, "candidate-cas")

	startReview := make(chan struct{})
	reviewResults := make(chan AdminProviderRouteResult, 2)
	var reviewWaitGroup sync.WaitGroup
	for _, decision := range []string{"approve", "reject"} {
		reviewWaitGroup.Add(1)
		go func(decision string) {
			defer reviewWaitGroup.Done()
			<-startReview
			reviewResults <- service.ReviewCandidate(ctx, "gateway-default", "candidate-cas", AdminProviderRouteReviewInput{
				ExpectedReviewVersion: 0, Decision: decision, Reason: "Concurrent reviewer decision is deliberate.",
			})
		}(decision)
	}
	close(startReview)
	reviewWaitGroup.Wait()
	close(reviewResults)
	reviewSuccesses := 0
	reviewConflicts := 0
	for result := range reviewResults {
		if result.FailureCode == "" {
			reviewSuccesses++
		} else if result.FailureCode == AdminProviderRouteFailureReviewVersionConflict {
			reviewConflicts++
		} else {
			t.Fatalf("unexpected review result: %#v", result)
		}
	}
	if reviewSuccesses != 1 || reviewConflicts != 1 {
		t.Fatalf("review successes=%d conflicts=%d", reviewSuccesses, reviewConflicts)
	}
	activationCandidate := service.CreateCandidate(ctx, AdminProviderRouteCandidateInput{
		ConfigurationID: "gateway-default", CandidateID: "candidate-activation-cas", ExpectedDraftRevision: 1,
	})
	if activationCandidate.FailureCode != "" {
		t.Fatalf("create activation candidate: %#v", activationCandidate)
	}
	activationApproval := service.ReviewCandidate(ctx, "gateway-default", "candidate-activation-cas", AdminProviderRouteReviewInput{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: "Dedicated candidate for activation generation CAS.",
	})
	if activationApproval.FailureCode != "" {
		t.Fatalf("approve activation candidate: %#v", activationApproval)
	}

	const workers = 12
	startActivation := make(chan struct{})
	activationResults := make(chan AdminProviderRouteResult, workers)
	var activationWaitGroup sync.WaitGroup
	for index := 0; index < workers; index++ {
		activationWaitGroup.Add(1)
		go func() {
			defer activationWaitGroup.Done()
			<-startActivation
			activationResults <- service.Activate(ctx, AdminProviderRouteActivationInput{
				ConfigurationID: "gateway-default", CandidateID: "candidate-activation-cas", ExpectedGeneration: 0,
				Action: "activate", Reason: "Concurrent activation validates generation CAS.",
			})
		}()
	}
	close(startActivation)
	activationWaitGroup.Wait()
	close(activationResults)
	activationSuccesses := 0
	activationConflicts := 0
	for result := range activationResults {
		if result.FailureCode == "" {
			activationSuccesses++
		} else if result.FailureCode == AdminProviderRouteFailureGenerationConflict {
			activationConflicts++
		} else {
			t.Fatalf("unexpected activation result: %#v", result)
		}
	}
	if activationSuccesses != 1 || activationConflicts != workers-1 {
		t.Fatalf("activation successes=%d conflicts=%d", activationSuccesses, activationConflicts)
	}
	history := service.ListActivations(ctx, "gateway-default")
	if len(history.ActivationHistory) != 1 {
		t.Fatalf("concurrent activation appended partial history: %#v", history)
	}
}

func TestAdminProviderRouteValidationFailsClosed(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*AdminProviderRouteContext, *AdminProviderRouteDraftInput)
		failureCode string
	}{
		{
			name: "production environment",
			mutate: func(ctx *AdminProviderRouteContext, _ *AdminProviderRouteDraftInput) {
				ctx.Environment = "production"
			},
			failureCode: AdminProviderRouteFailureEnvironmentForbidden,
		},
		{
			name: "raw endpoint",
			mutate: func(_ *AdminProviderRouteContext, input *AdminProviderRouteDraftInput) {
				input.ProviderProfiles[0].RuntimeProfileRef = "https://provider.example/v1"
			},
			failureCode: AdminProviderRouteFailurePayloadInvalid,
		},
		{
			name: "cross environment ref",
			mutate: func(_ *AdminProviderRouteContext, input *AdminProviderRouteDraftInput) {
				input.ProviderProfiles[0].RuntimeProfileRef = "ref:radishmind/development/provider-profiles/mock-primary"
			},
			failureCode: AdminProviderRouteFailureEnvironmentForbidden,
		},
		{
			name: "sensitive display name",
			mutate: func(_ *AdminProviderRouteContext, input *AdminProviderRouteDraftInput) {
				input.DisplayName = "Authorization: Bearer private"
			},
			failureCode: AdminProviderRouteFailureSensitiveForbidden,
		},
		{
			name: "capability mismatch",
			mutate: func(_ *AdminProviderRouteContext, input *AdminProviderRouteDraftInput) {
				input.ProviderProfiles[0].Capabilities = []string{"responses"}
			},
			failureCode: AdminProviderRouteFailurePayloadInvalid,
		},
		{
			name: "duplicate route binding",
			mutate: func(_ *AdminProviderRouteContext, input *AdminProviderRouteDraftInput) {
				input.ModelRoutes = append(input.ModelRoutes, AdminModelRouteDefinition{
					RouteID: "route-duplicate", Protocol: "chat_completions",
					ModelID: "mock-model", ProviderProfileID: "primary",
				})
			},
			failureCode: AdminProviderRouteFailurePayloadInvalid,
		},
		{
			name: "unknown protocol",
			mutate: func(_ *AdminProviderRouteContext, input *AdminProviderRouteDraftInput) {
				input.ModelRoutes[0].Protocol = "embeddings"
			},
			failureCode: AdminProviderRouteFailurePayloadInvalid,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service, _, _ := newAdminProviderRouteTestService()
			ctx := adminProviderRouteTestContext()
			input := adminProviderRouteTestDraftInput(0, "mock-primary")
			test.mutate(&ctx, &input)
			result := service.PutDraft(ctx, input)
			if result.FailureCode != test.failureCode || result.Draft != nil {
				t.Fatalf("result=%#v", result)
			}
		})
	}
}

func TestAdminProviderRouteInventoryFailureAndDriftHaveNoSideEffects(t *testing.T) {
	service, repository, resolver := newAdminProviderRouteTestService()
	ctx := adminProviderRouteTestContext()
	if result := service.PutDraft(ctx, adminProviderRouteTestDraftInput(0, "mock-primary")); result.FailureCode != "" {
		t.Fatalf("seed draft: %#v", result)
	}
	resolver.setUnavailable(true)
	unavailable := service.CreateCandidate(ctx, AdminProviderRouteCandidateInput{
		ConfigurationID: "gateway-default", CandidateID: "candidate-unavailable", ExpectedDraftRevision: 1,
	})
	if unavailable.FailureCode != AdminProviderRouteFailureInventoryUnavailable {
		t.Fatalf("inventory unavailable: %#v", unavailable)
	}
	resolver.setUnavailable(false)
	missing := service.PutDraft(ctx, AdminProviderRouteDraftInput{
		ConfigurationID: "gateway-missing", ExpectedRevision: 0, DisplayName: "Missing profile routing",
		ProviderProfiles: []AdminProviderProfileAssignment{{
			ProfileID: "missing", DisplayName: "Missing profile", ProviderID: "mock",
			RuntimeProfileRef: "ref:radishmind/test/provider-profiles/missing",
			Capabilities:      []string{"chat_completions"},
		}},
		ModelRoutes: []AdminModelRouteDefinition{{
			RouteID: "route-missing", Protocol: "chat_completions", ModelID: "mock-model", ProviderProfileID: "missing",
		}},
	})
	if missing.FailureCode != "" {
		t.Fatalf("missing inventory draft should remain reviewable: %#v", missing)
	}
	missingCandidate := service.CreateCandidate(ctx, AdminProviderRouteCandidateInput{
		ConfigurationID: "gateway-missing", CandidateID: "candidate-missing", ExpectedDraftRevision: 1,
	})
	if missingCandidate.FailureCode != AdminProviderRouteFailureInventoryNotFound {
		t.Fatalf("missing inventory candidate: %#v", missingCandidate)
	}

	candidate := service.CreateCandidate(ctx, AdminProviderRouteCandidateInput{
		ConfigurationID: "gateway-default", CandidateID: "candidate-drift", ExpectedDraftRevision: 1,
	})
	if candidate.FailureCode != "" {
		t.Fatalf("create drift candidate: %#v", candidate)
	}
	approval := service.ReviewCandidate(ctx, "gateway-default", "candidate-drift", AdminProviderRouteReviewInput{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: "Review completed before inventory drift.",
	})
	if approval.FailureCode != "" {
		t.Fatalf("approve drift candidate: %#v", approval)
	}
	drifted := resolver.bindings["test\x00mock\x00ref:radishmind/test/provider-profiles/mock-primary"]
	drifted.InventoryDigest = "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	resolver.put(drifted)
	activation := service.Activate(ctx, AdminProviderRouteActivationInput{
		ConfigurationID: "gateway-default", CandidateID: "candidate-drift", ExpectedGeneration: 0,
		Action: "activate", Reason: "This activation must be blocked by inventory drift.",
	})
	if activation.FailureCode != AdminProviderRouteFailureInventoryMismatch || activation.Snapshot != nil {
		t.Fatalf("inventory drift activation: %#v", activation)
	}
	history := service.ListActivations(ctx, "gateway-default")
	if len(history.ActivationHistory) != 0 {
		t.Fatalf("inventory drift appended activation history: %#v", history)
	}
	if _, exists := repository.snapshots[adminProviderRouteConfigurationKey(ctx, "gateway-default")]; exists {
		t.Fatal("inventory drift wrote an active snapshot")
	}
}

func TestAdminProviderRoutePermissionsAndScopeIsolation(t *testing.T) {
	service, _, _ := newAdminProviderRouteTestService()
	ctx := adminProviderRouteTestContext()
	denied := ctx
	denied.DraftEnabled = false
	if result := service.PutDraft(denied, adminProviderRouteTestDraftInput(0, "mock-primary")); result.FailureCode != AdminProviderRouteFailureScopeDenied {
		t.Fatalf("draft permission: %#v", result)
	}
	created := service.PutDraft(ctx, adminProviderRouteTestDraftInput(0, "mock-primary"))
	if created.FailureCode != "" {
		t.Fatalf("create draft: %#v", created)
	}
	otherTenant := ctx
	otherTenant.TenantRef = "tenant-other"
	read := service.ReadDraft(otherTenant, "gateway-default")
	if read.FailureCode != AdminProviderRouteFailureDraftNotFound || read.Draft != nil {
		t.Fatalf("cross-tenant draft leaked: %#v", read)
	}
	otherWorkspace := ctx
	otherWorkspace.WorkspaceID = "workspace-other"
	candidate := service.CreateCandidate(otherWorkspace, AdminProviderRouteCandidateInput{
		ConfigurationID: "gateway-default", CandidateID: "candidate-other", ExpectedDraftRevision: 1,
	})
	if candidate.FailureCode != AdminProviderRouteFailureDraftNotFound {
		t.Fatalf("cross-workspace candidate leaked: %#v", candidate)
	}
	otherEnvironment := ctx
	otherEnvironment.Environment = "development"
	environmentRead := service.ReadDraft(otherEnvironment, "gateway-default")
	if environmentRead.FailureCode != AdminProviderRouteFailureDraftNotFound {
		t.Fatalf("cross-environment draft leaked: %#v", environmentRead)
	}
}

func TestAdminProviderRouteRejectedCandidateAndRollbackRules(t *testing.T) {
	service, _, _ := newAdminProviderRouteTestService()
	ctx := adminProviderRouteTestContext()
	adminProviderRoutePrepareCandidate(t, service, ctx, "candidate-rejected")
	rejected := service.ReviewCandidate(ctx, "gateway-default", "candidate-rejected", AdminProviderRouteReviewInput{
		ExpectedReviewVersion: 0, Decision: "reject", Reason: "Route evidence does not satisfy the review.",
	})
	if rejected.FailureCode != "" {
		t.Fatalf("reject candidate: %#v", rejected)
	}
	activation := service.Activate(ctx, AdminProviderRouteActivationInput{
		ConfigurationID: "gateway-default", CandidateID: "candidate-rejected", ExpectedGeneration: 0,
		Action: "activate", Reason: "Rejected candidate must not activate.",
	})
	if activation.FailureCode != AdminProviderRouteFailureCandidateNotApproved {
		t.Fatalf("rejected candidate activated: %#v", activation)
	}

	adminProviderRoutePrepareCandidateWithRevision(t, service, ctx, "candidate-approved", 1)
	approved := service.ReviewCandidate(ctx, "gateway-default", "candidate-approved", AdminProviderRouteReviewInput{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: "Candidate is valid but has never been active.",
	})
	if approved.FailureCode != "" {
		t.Fatalf("approve candidate: %#v", approved)
	}
	rollback := service.Activate(ctx, AdminProviderRouteActivationInput{
		ConfigurationID: "gateway-default", CandidateID: "candidate-approved", ExpectedGeneration: 0,
		Action: "rollback", Reason: "Never-active candidate is not a rollback target.",
	})
	if rollback.FailureCode != AdminProviderRouteFailureRollbackTargetInvalid {
		t.Fatalf("invalid rollback target: %#v", rollback)
	}
	secondReview := service.ReviewCandidate(ctx, "gateway-default", "candidate-approved", AdminProviderRouteReviewInput{
		ExpectedReviewVersion: 1, Decision: "reject", Reason: "Terminal review cannot be overwritten.",
	})
	if secondReview.FailureCode != AdminProviderRouteFailureReviewTransitionInvalid {
		t.Fatalf("terminal review changed: %#v", secondReview)
	}
}

func TestAdminProviderRouteStoreUnavailableIsStable(t *testing.T) {
	service, repository, _ := newAdminProviderRouteTestService()
	ctx := adminProviderRouteTestContext()
	repository.setUnavailableForTest(true)
	for _, result := range []AdminProviderRouteResult{
		service.PutDraft(ctx, adminProviderRouteTestDraftInput(0, "mock-primary")),
		service.ReadDraft(ctx, "gateway-default"),
		service.ReadActiveSnapshot(ctx, "gateway-default"),
		service.ListActivations(ctx, "gateway-default"),
	} {
		if result.FailureCode != AdminProviderRouteFailureStoreUnavailable {
			t.Fatalf("unstable unavailable result: %#v", result)
		}
	}
}

func TestAdminProviderRouteSensitiveReviewAndActivationReasonsAreRejected(t *testing.T) {
	service, _, _ := newAdminProviderRouteTestService()
	ctx := adminProviderRouteTestContext()
	adminProviderRoutePrepareCandidate(t, service, ctx, "candidate-sensitive")
	review := service.ReviewCandidate(ctx, "gateway-default", "candidate-sensitive", AdminProviderRouteReviewInput{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: "Authorization: Bearer private-token",
	})
	if review.FailureCode != AdminProviderRouteFailureSensitiveForbidden {
		t.Fatalf("sensitive review reason: %#v", review)
	}
	approved := service.ReviewCandidate(ctx, "gateway-default", "candidate-sensitive", AdminProviderRouteReviewInput{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: "Reviewed without sensitive runtime material.",
	})
	if approved.FailureCode != "" {
		t.Fatalf("approve candidate: %#v", approved)
	}
	activation := service.Activate(ctx, AdminProviderRouteActivationInput{
		ConfigurationID: "gateway-default", CandidateID: "candidate-sensitive", ExpectedGeneration: 0,
		Action: "activate", Reason: "credential=private-value",
	})
	if activation.FailureCode != AdminProviderRouteFailureSensitiveForbidden {
		t.Fatalf("sensitive activation reason: %#v", activation)
	}
	if active := service.ReadActiveSnapshot(ctx, "gateway-default"); active.Snapshot != nil || active.CurrentGeneration != 0 {
		t.Fatalf("sensitive activation produced side effects: %#v", active)
	}
}

func TestAdminProviderRouteReadIntegrityRejectsStoredDrift(t *testing.T) {
	service, repository, _ := newAdminProviderRouteTestService()
	ctx := adminProviderRouteTestContext()
	adminProviderRoutePrepareCandidate(t, service, ctx, "candidate-integrity")
	approval := service.ReviewCandidate(ctx, "gateway-default", "candidate-integrity", AdminProviderRouteReviewInput{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: "Reviewed before deliberate repository drift.",
	})
	if approval.FailureCode != "" {
		t.Fatalf("approve candidate: %#v", approval)
	}
	activation := service.Activate(ctx, AdminProviderRouteActivationInput{
		ConfigurationID: "gateway-default", CandidateID: "candidate-integrity", ExpectedGeneration: 0,
		Action: "activate", Reason: "Activate before deliberate repository drift.",
	})
	if activation.FailureCode != "" {
		t.Fatalf("activate candidate: %#v", activation)
	}

	repository.mu.Lock()
	draftKey := adminProviderRouteConfigurationKey(ctx, "gateway-default")
	draft := repository.drafts[draftKey]
	draft.DisplayName = "Stored draft drift"
	repository.drafts[draftKey] = draft
	candidateKey := adminProviderRouteCandidateKey(ctx, "gateway-default", "candidate-integrity")
	candidate := repository.candidates[candidateKey]
	candidate.Configuration.DisplayName = "Stored candidate drift"
	repository.candidates[candidateKey] = candidate
	snapshot := repository.snapshots[draftKey]
	snapshot.Configuration.DisplayName = "Stored snapshot drift"
	repository.snapshots[draftKey] = snapshot
	history := repository.activations[draftKey]
	history[0].AfterSnapshotDigest = "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
	repository.activations[draftKey] = history
	repository.mu.Unlock()

	for _, result := range []AdminProviderRouteResult{
		service.ReadDraft(ctx, "gateway-default"),
		service.ReadCandidate(ctx, "gateway-default", "candidate-integrity"),
		service.ReadActiveSnapshot(ctx, "gateway-default"),
		service.ListActivations(ctx, "gateway-default"),
	} {
		if result.FailureCode != AdminProviderRouteFailureStoreUnavailable {
			t.Fatalf("stored drift escaped integrity guard: %#v", result)
		}
	}
}

func newAdminProviderRouteTestService() (
	adminProviderRouteService,
	*memoryAdminProviderRouteRepository,
	*fakeAdminProviderInventoryResolver,
) {
	repository := newMemoryAdminProviderRouteRepository()
	resolver := newFakeAdminProviderInventoryResolver()
	service := newAdminProviderRouteService(repository, resolver)
	var clockMu sync.Mutex
	next := time.Date(2026, 7, 26, 8, 0, 0, 0, time.UTC)
	service.now = func() time.Time {
		clockMu.Lock()
		defer clockMu.Unlock()
		current := next
		next = next.Add(time.Second)
		return current
	}
	return service, repository, resolver
}

func adminProviderRouteTestContext() AdminProviderRouteContext {
	return AdminProviderRouteContext{
		RequestContext: context.Background(),
		RequestID:      "request-admin-provider-route", TenantRef: "tenant-alpha",
		WorkspaceID: "workspace-alpha", Environment: "test",
		ActorRef: "subject-admin", AuditRef: "audit-admin-provider-route",
		ReadEnabled: true, DraftEnabled: true, ReviewEnabled: true, ActivateEnabled: true,
	}
}

func adminProviderRouteTestProfile(profileID, runtimeKey string) AdminProviderProfileAssignment {
	return AdminProviderProfileAssignment{
		ProfileID: profileID, DisplayName: "Mock profile", ProviderID: "mock",
		RuntimeProfileRef: "ref:radishmind/test/provider-profiles/" + runtimeKey,
		Capabilities:      []string{"messages", "chat_completions", "responses"},
	}
}

func adminProviderRouteTestDraftInput(expectedRevision int, runtimeKey string) AdminProviderRouteDraftInput {
	return AdminProviderRouteDraftInput{
		ConfigurationID: "gateway-default", ExpectedRevision: expectedRevision, DisplayName: "Default test routing",
		ProviderProfiles: []AdminProviderProfileAssignment{adminProviderRouteTestProfile("primary", runtimeKey)},
		ModelRoutes: []AdminModelRouteDefinition{{
			RouteID: "route-chat", Protocol: "chat_completions", ModelID: "mock-model", ProviderProfileID: "primary",
		}},
	}
}

func adminProviderRoutePrepareCandidate(
	t *testing.T,
	service adminProviderRouteService,
	ctx AdminProviderRouteContext,
	candidateID string,
) {
	t.Helper()
	adminProviderRoutePrepareCandidateWithRevision(t, service, ctx, candidateID, 0)
}

func adminProviderRoutePrepareCandidateWithRevision(
	t *testing.T,
	service adminProviderRouteService,
	ctx AdminProviderRouteContext,
	candidateID string,
	expectedDraftRevision int,
) {
	t.Helper()
	revision := expectedDraftRevision
	if revision == 0 {
		draft := service.PutDraft(ctx, adminProviderRouteTestDraftInput(0, "mock-primary"))
		if draft.FailureCode != "" {
			t.Fatalf("prepare draft: %#v", draft)
		}
		revision = draft.Draft.DraftRevision
	}
	candidate := service.CreateCandidate(ctx, AdminProviderRouteCandidateInput{
		ConfigurationID: "gateway-default", CandidateID: candidateID, ExpectedDraftRevision: revision,
	})
	if candidate.FailureCode != "" {
		t.Fatalf("prepare candidate: %#v", candidate)
	}
}
