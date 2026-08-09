package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestApplicationEvaluationPlanVersionsCASAndArchive(t *testing.T) {
	service, ctx := newApplicationEvaluationPlanTestService(t, "prompt_application")
	created := service.Create(ctx, applicationEvaluationPromptPlanInput("Prompt regression", "unchanged"))
	if created.FailureCode != "" || created.Plan == nil || created.Version == nil {
		t.Fatalf("unexpected create result: %+v", created)
	}
	if created.Plan.RecordVersion != 1 || created.Plan.LatestPlanVersion != 1 || created.Version.PlanVersion != 1 ||
		created.Plan.LatestPlanDigest != created.Version.PlanDigest || len(created.Version.Items) != 1 {
		t.Fatalf("unexpected created plan: %+v / %+v", created.Plan, created.Version)
	}

	revisionInput := ApplicationEvaluationPlanReviseInput{
		ExpectedVersion: 1, Name: "Prompt regression revised", ExecutionProfile: applicationInteractionProfilePrompt,
		Items: []ApplicationEvaluationPlanItem{
			applicationEvaluationPromptItem("first", "First prompt", "changed"),
			applicationEvaluationPromptItem("second", "Second prompt", "unchanged"),
		},
	}
	revised := service.Revise(ctx, created.Plan.PlanID, revisionInput)
	if revised.FailureCode != "" || revised.Plan == nil || revised.Version == nil || revised.Plan.RecordVersion != 2 ||
		revised.Version.PlanVersion != 2 || revised.Version.PreviousPlanVersion != 1 || revised.Plan.ItemCount != 2 {
		t.Fatalf("unexpected revision result: %+v", revised)
	}
	firstVersion := service.ReadVersion(ctx, created.Plan.PlanID, 1)
	if firstVersion.FailureCode != "" || firstVersion.Version == nil || firstVersion.Version.Name != "Prompt regression" || len(firstVersion.Version.Items) != 1 {
		t.Fatalf("immutable v1 changed: %+v", firstVersion)
	}

	stale := service.Revise(ctx, created.Plan.PlanID, revisionInput)
	if stale.FailureCode != ApplicationEvaluationFailureVersionConflict || stale.CurrentRecordVersion != 2 || stale.CurrentState != applicationEvaluationPlanStateActive {
		t.Fatalf("expected version conflict, got %+v", stale)
	}
	archive := service.Archive(ctx, created.Plan.PlanID, ApplicationEvaluationPlanArchiveInput{ExpectedVersion: 2, AcknowledgeNoNewCampaigns: true})
	if archive.FailureCode != "" || archive.Plan == nil || archive.Plan.RecordVersion != 3 || archive.Plan.LifecycleState != applicationEvaluationPlanStateArchived {
		t.Fatalf("unexpected archive result: %+v", archive)
	}
	revisionInput.ExpectedVersion = 3
	if result := service.Revise(ctx, created.Plan.PlanID, revisionInput); result.FailureCode != ApplicationEvaluationFailureArchived {
		t.Fatalf("archived plan revision must fail closed: %+v", result)
	}
}

func TestApplicationEvaluationPlanRejectsSecretsProfilesAndDuplicateKeys(t *testing.T) {
	service, ctx := newApplicationEvaluationPlanTestService(t, "prompt_application")
	secret := applicationEvaluationPromptPlanInput("Secret fixtures", "unchanged")
	secret.Items[0].PromptApplication.Variables["token"] = "sk-secret-value"
	if result := service.Create(ctx, secret); result.FailureCode != ApplicationEvaluationFailureSecretForbidden {
		t.Fatalf("secret fixture must be rejected: %+v", result)
	}

	duplicate := applicationEvaluationPromptPlanInput("Duplicate fixtures", "unchanged")
	duplicate.Items = append(duplicate.Items, duplicate.Items[0])
	if result := service.Create(ctx, duplicate); result.FailureCode != ApplicationEvaluationFailurePayloadInvalid {
		t.Fatalf("duplicate item key must be rejected: %+v", result)
	}

	wrongProfile := applicationEvaluationPromptPlanInput("Wrong profile", "unchanged")
	wrongProfile.ExecutionProfile = applicationInteractionProfileAgentCopilot
	if result := service.Create(ctx, wrongProfile); result.FailureCode != ApplicationEvaluationFailurePayloadInvalid {
		t.Fatalf("fixture/profile mismatch must be rejected before application eligibility: %+v", result)
	}

	invalidEnvironment := ctx
	invalidEnvironment.Environment = "production"
	if result := service.Create(invalidEnvironment, applicationEvaluationPromptPlanInput("Production denied", "unchanged")); result.FailureCode != ApplicationEvaluationFailureEnvironmentDenied {
		t.Fatalf("production must be rejected: %+v", result)
	}
}

func TestApplicationEvaluationPlanCursorScopeAndVersionBinding(t *testing.T) {
	service, ctx := newApplicationEvaluationPlanTestService(t, "prompt_application")
	base := time.Date(2026, 8, 9, 8, 0, 0, 0, time.UTC)
	sequence := 0
	service.now = func() time.Time {
		sequence++
		return base.Add(time.Duration(sequence) * time.Second)
	}
	ids := []string{"aeplan_aaaaaaaaaaaaaaaa", "aeplan_bbbbbbbbbbbbbbbb", "aeplan_cccccccccccccccc"}
	service.newPlanID = func() (string, error) {
		value := ids[0]
		ids = ids[1:]
		return value, nil
	}
	for index := 0; index < 3; index++ {
		input := applicationEvaluationPromptPlanInput("Plan "+string(rune('A'+index)), "unchanged")
		if result := service.Create(ctx, input); result.FailureCode != "" {
			t.Fatalf("create plan %d: %+v", index, result)
		}
	}
	first := service.List(ctx, ApplicationEvaluationPlanListInput{Limit: 2})
	if first.FailureCode != "" || len(first.Plans) != 2 || !first.HasMore || first.NextCursor == "" {
		t.Fatalf("unexpected first page: %+v", first)
	}
	second := service.List(ctx, ApplicationEvaluationPlanListInput{Limit: 2, Cursor: first.NextCursor})
	if second.FailureCode != "" || len(second.Plans) != 1 || second.HasMore {
		t.Fatalf("unexpected second page: %+v", second)
	}
	wrongEnvironment := ctx
	wrongEnvironment.Environment = applicationEvaluationEnvironmentDevelopment
	if result := service.List(wrongEnvironment, ApplicationEvaluationPlanListInput{Limit: 2, Cursor: first.NextCursor}); result.FailureCode != ApplicationEvaluationFailureCursorInvalid {
		t.Fatalf("cross-environment cursor must fail: %+v", result)
	}
	if result := service.List(ctx, ApplicationEvaluationPlanListInput{Limit: 3, Cursor: first.NextCursor}); result.FailureCode != ApplicationEvaluationFailureCursorInvalid {
		t.Fatalf("cursor limit drift must fail: %+v", result)
	}
}

func TestMemoryApplicationEvaluationCampaignIdempotencyAndTransitions(t *testing.T) {
	repository := newMemoryApplicationEvaluationRepository(10, 10)
	ctx := applicationEvaluationTestContext()
	now := "2026-08-09T08:00:00Z"
	campaign := applicationEvaluationPendingCampaign(ctx, "campaign-key", now)
	created, inserted, err := repository.CreateCampaign(ctx, campaign)
	if err != nil || !inserted || created.CampaignID != campaign.CampaignID {
		t.Fatalf("create campaign: inserted=%v err=%v value=%+v", inserted, err, created)
	}
	replayed, inserted, err := repository.CreateCampaign(ctx, campaign)
	if err != nil || inserted || replayed.CampaignID != campaign.CampaignID {
		t.Fatalf("idempotent campaign replay: inserted=%v err=%v value=%+v", inserted, err, replayed)
	}
	conflicting := campaign
	conflicting.PlanVersion = 2
	conflicting.PlanDigest = "sha256:" + strings.Repeat("b", 64)
	conflicting.CampaignID = applicationEvaluationDeterministicCampaignID(ctx, campaign.ClientCampaignKey)
	if _, _, err = repository.CreateCampaign(ctx, conflicting); !errors.Is(err, errApplicationEvaluationCampaignConflict) {
		t.Fatalf("expected client campaign key conflict, got %v", err)
	}

	authority := applicationEvaluationWorkflowAuthority(t, ctx)
	running := campaign
	running.RecordVersion = 2
	running.State = applicationEvaluationCampaignStateRunning
	running.StartedAt = now
	running.Authority = &authority
	running.UpdatedByActorRef = "subject-owner"
	stored, updated, err := repository.UpdateCampaign(ctx, 1, running)
	if err != nil || !updated || stored.State != applicationEvaluationCampaignStateRunning {
		t.Fatalf("start campaign: updated=%v err=%v value=%+v", updated, err, stored)
	}

	finished := running
	finished.RecordVersion = 3
	finished.State = applicationEvaluationCampaignStateSucceeded
	finished.CurrentItemIndex = 1
	finished.SucceededItems = 1
	finished.CompletedAt = "2026-08-09T08:00:01Z"
	finished.Items[0].State = applicationEvaluationCampaignItemSucceeded
	finished.Items[0].StartedAt = now
	finished.Items[0].CompletedAt = finished.CompletedAt
	finished.Items[0].RunSchemaVersion = workflowRunRecordDefinitionSchemaVersion
	finished.Items[0].RunProfile = workflowDefinitionExecutorProfile
	finished.Items[0].AuthorityDigest = authority.AuthorityDigest
	stored, updated, err = repository.UpdateCampaign(ctx, 2, finished)
	if err != nil || !updated || stored.State != applicationEvaluationCampaignStateSucceeded {
		t.Fatalf("finish campaign: updated=%v err=%v value=%+v", updated, err, stored)
	}
	invalid := finished
	invalid.RecordVersion = 4
	invalid.State = applicationEvaluationCampaignStateRunning
	invalid.CompletedAt = ""
	if _, _, err = repository.UpdateCampaign(ctx, 3, invalid); !errors.Is(err, errApplicationEvaluationStoreContract) {
		t.Fatalf("terminal campaign must not return to running: %v", err)
	}
}

func TestMemoryApplicationEvaluationRepositoryFailureAndCorruptionClose(t *testing.T) {
	service, ctx := newApplicationEvaluationPlanTestService(t, "prompt_application")
	repository := service.repository.(*memoryApplicationEvaluationRepository)
	repository.unavailable = true
	if result := service.Create(ctx, applicationEvaluationPromptPlanInput("Store failure", "unchanged")); result.FailureCode != ApplicationEvaluationFailureStoreUnavailable {
		t.Fatalf("store failure must close: %+v", result)
	}
	repository.unavailable = false
	created := service.Create(ctx, applicationEvaluationPromptPlanInput("Corruption", "unchanged"))
	if created.FailureCode != "" || created.Plan == nil {
		t.Fatalf("create before corruption: %+v", created)
	}
	key := applicationEvaluationPlanKey(ctx, created.Plan.PlanID)
	corrupted := repository.plans[key]
	corrupted.LatestPlanDigest = "invalid"
	repository.plans[key] = corrupted
	if result := service.Read(ctx, created.Plan.PlanID); result.FailureCode != ApplicationEvaluationFailureStoreContract {
		t.Fatalf("corrupted record must fail closed: %+v", result)
	}
}

func TestApplicationEvaluationCampaignExecutesSequentialDurableRuns(t *testing.T) {
	planService, ctx := newApplicationEvaluationPlanTestService(t, "workflow_copilot")
	created := planService.Create(ctx, applicationEvaluationWorkflowPlanInput("Sequential campaign"))
	if created.FailureCode != "" || created.Plan == nil || created.Version == nil {
		t.Fatalf("create campaign plan: %+v", created)
	}
	revision := planService.Revise(ctx, created.Plan.PlanID, ApplicationEvaluationPlanReviseInput{
		ExpectedVersion: 1, Name: "Sequential campaign", ExecutionProfile: workflowDefinitionExecutorProfile,
		Target: created.Version.Target, Items: append(created.Version.Items, ApplicationEvaluationPlanItem{
			ItemKey: "second", Name: "Second input", ExpectedClassification: WorkflowRunComparisonUnchanged,
			WorkflowDefinition: &ApplicationEvaluationDefinitionFixture{InputText: "second input", ConditionValues: map[string]bool{}},
		}),
	})
	if revision.FailureCode != "" || revision.Plan == nil || revision.Version == nil {
		t.Fatalf("revise campaign plan: %+v", revision)
	}
	authority := applicationEvaluationWorkflowAuthority(t, ctx)
	invocations := make([]string, 0, 2)
	quotaBindings := make([]gatewayRequestQuotaBinding, 0, 2)
	service := newApplicationEvaluationCampaignService(
		planService.repository,
		func(ApplicationEvaluationContext, ApplicationEvaluationPlanVersion) (ApplicationEvaluationCampaignAuthority, string) {
			return authority, ""
		},
		func(callContext ApplicationEvaluationContext, _ ApplicationEvaluationPlanVersion, _ ApplicationEvaluationPlanItem, runID string) (*WorkflowRunRecord, string, string) {
			invocations = append(invocations, runID)
			binding, found := gatewayRequestQuotaBindingFromContext(callContext.RequestContext)
			if !found {
				t.Fatal("campaign invocation did not carry quota binding")
			}
			quotaBindings = append(quotaBindings, binding)
			return applicationEvaluationSucceededDefinitionRun(callContext, authority, runID), "", ""
		},
		func(ApplicationEvaluationContext, string) (WorkflowRunRecord, bool, error) {
			return WorkflowRunRecord{}, false, nil
		},
	)
	result := service.Execute(ctx, ApplicationEvaluationCampaignExecuteInput{
		PlanID: revision.Plan.PlanID, PlanVersion: revision.Version.PlanVersion, PlanDigest: revision.Version.PlanDigest,
		ExpectedPlanRecordVersion: revision.Plan.RecordVersion, ClientCampaignKey: "campaign_sequential",
		QuotaAPIKeyID:                  "key_aaaaaaaaaaaaaaaa",
		AcknowledgeSequentialExecution: true, AcknowledgeQuotaConsumption: true,
	})
	if result.FailureCode != "" || result.Campaign == nil || result.Campaign.State != applicationEvaluationCampaignStateSucceeded ||
		result.Campaign.SucceededItems != 2 || len(invocations) != 2 || invocations[0] == invocations[1] || len(quotaBindings) != 2 {
		t.Fatalf("unexpected sequential campaign: result=%+v invocations=%v", result, invocations)
	}
	for index, binding := range quotaBindings {
		if binding.APIKeyID != "key_aaaaaaaaaaaaaaaa" || binding.RequestID != invocations[index] ||
			binding.QuotaContext.RequestID != invocations[index] || binding.Route != workflowDefinitionRunCreateRoute ||
			binding.QuotaContext.TenantRef != ctx.TenantRef || binding.QuotaContext.WorkspaceID != ctx.WorkspaceID ||
			binding.QuotaContext.Environment != ctx.Environment || binding.QuotaContext.ApplicationID != ctx.ApplicationID {
			t.Fatalf("quota binding does not preserve campaign scope: %+v", binding)
		}
	}
	replay := service.Execute(ctx, ApplicationEvaluationCampaignExecuteInput{
		PlanID: revision.Plan.PlanID, PlanVersion: revision.Version.PlanVersion, PlanDigest: revision.Version.PlanDigest,
		ExpectedPlanRecordVersion: revision.Plan.RecordVersion, ClientCampaignKey: "campaign_sequential",
		QuotaAPIKeyID:                  "key_aaaaaaaaaaaaaaaa",
		AcknowledgeSequentialExecution: true, AcknowledgeQuotaConsumption: true,
	})
	if replay.FailureCode != "" || !replay.IdempotentReplay || replay.Campaign == nil || len(invocations) != 2 {
		t.Fatalf("campaign replay must not invoke provider again: result=%+v invocations=%v", replay, invocations)
	}
}

func TestApplicationEvaluationCampaignStopsOnAuthorityDriftAndQuotaFailure(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		drift       bool
		failureCode string
	}{
		{name: "authority drift", drift: true, failureCode: ApplicationEvaluationFailureAuthorityChanged},
		{name: "quota admission", failureCode: GatewayRequestQuotaFailureExceeded},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			planService, ctx := newApplicationEvaluationPlanTestService(t, "workflow_copilot")
			created := planService.Create(ctx, applicationEvaluationWorkflowPlanInput("Stopped campaign"))
			if created.FailureCode != "" || created.Plan == nil || created.Version == nil {
				t.Fatalf("create plan: %+v", created)
			}
			authority := applicationEvaluationWorkflowAuthority(t, ctx)
			resolveCalls, invokeCalls := 0, 0
			service := newApplicationEvaluationCampaignService(
				planService.repository,
				func(ApplicationEvaluationContext, ApplicationEvaluationPlanVersion) (ApplicationEvaluationCampaignAuthority, string) {
					resolveCalls++
					if testCase.drift && resolveCalls >= 3 {
						changed := authority
						changed.AuthorityDigest = "sha256:" + strings.Repeat("c", 64)
						return changed, ""
					}
					return authority, ""
				},
				func(callContext ApplicationEvaluationContext, _ ApplicationEvaluationPlanVersion, _ ApplicationEvaluationPlanItem, runID string) (*WorkflowRunRecord, string, string) {
					invokeCalls++
					run := applicationEvaluationSucceededDefinitionRun(callContext, authority, runID)
					if testCase.failureCode == GatewayRequestQuotaFailureExceeded {
						run.Status = WorkflowRunStatusFailed
						run.FailureCode = WorkflowRunFailureCode(GatewayRequestQuotaFailureExceeded)
						run.FailureSummary = "quota exhausted before provider"
						run.Diagnostic = &WorkflowRunDiagnostic{FailureBoundary: WorkflowRunFailureBoundary("quota_admission")}
						return run, GatewayRequestQuotaFailureExceeded, run.FailureSummary
					}
					return run, "", ""
				},
				func(ApplicationEvaluationContext, string) (WorkflowRunRecord, bool, error) {
					return WorkflowRunRecord{}, false, nil
				},
			)
			result := service.Execute(ctx, ApplicationEvaluationCampaignExecuteInput{
				PlanID: created.Plan.PlanID, PlanVersion: 1, PlanDigest: created.Version.PlanDigest,
				ExpectedPlanRecordVersion: 1, ClientCampaignKey: "campaign_stopped",
				QuotaAPIKeyID:                  "key_aaaaaaaaaaaaaaaa",
				AcknowledgeSequentialExecution: true, AcknowledgeQuotaConsumption: true,
			})
			if result.Campaign == nil || result.Campaign.State != applicationEvaluationCampaignStateFailed || result.FailureCode != testCase.failureCode || invokeCalls != 1 {
				t.Fatalf("campaign must stop with exact failure: result=%+v invokes=%d", result, invokeCalls)
			}
		})
	}
}

func TestApplicationEvaluationCampaignHandoffAcceptsSemanticAuthorityJSON(t *testing.T) {
	ctx := applicationEvaluationTestContext()
	current := applicationEvaluationPendingCampaign(ctx, "semantic-authority", "2026-08-09T08:00:00Z")
	authority := applicationEvaluationWorkflowAuthority(t, ctx)
	current.RecordVersion = 3
	current.State = applicationEvaluationCampaignStateSucceeded
	current.StartedAt = "2026-08-09T08:00:01Z"
	current.CompletedAt = "2026-08-09T08:00:02Z"
	current.SucceededItems = 1
	current.Authority = &authority
	current.Items[0].State = applicationEvaluationCampaignItemSucceeded
	current.Items[0].RunSchemaVersion = workflowRunRecordDefinitionSchemaVersion
	current.Items[0].RunProfile = workflowDefinitionExecutorProfile
	current.Items[0].AuthorityDigest = authority.AuthorityDigest
	current.Items[0].StartedAt = current.StartedAt
	current.Items[0].CompletedAt = current.CompletedAt

	next := cloneApplicationEvaluationCampaign(current)
	next.RecordVersion = 4
	var normalizedAuthority any
	if err := json.Unmarshal(next.Authority.Snapshot, &normalizedAuthority); err != nil {
		t.Fatal(err)
	}
	next.Authority.Snapshot, _ = json.MarshalIndent(normalizedAuthority, "", "  ")
	next.Handoff = &ApplicationEvaluationHandoffRef{
		BaselineCampaignID:  "aecamp_bbbbbbbbbbbbbbbb",
		CandidateCampaignID: next.CampaignID,
		CaseRefs:            []WorkflowEvaluationSuiteCaseRef{{CaseID: "eval_semantic_authority", Version: 1}},
		State:               "partial",
		AuditRef:            "audit-semantic-authority",
	}
	if !validApplicationEvaluationCampaignUpdate(current, next) {
		t.Fatal("semantically equal authority JSON blocked a handoff checkpoint")
	}
}

func TestApplicationEvaluationCampaignReconcileNeverReplaysProvider(t *testing.T) {
	repository := newMemoryApplicationEvaluationRepository(10, 10)
	ctx := applicationEvaluationTestContext()
	authority := applicationEvaluationWorkflowAuthority(t, ctx)
	campaign := applicationEvaluationPendingCampaign(ctx, "campaign_reconcile", "2026-08-09T08:00:00Z")
	if _, inserted, err := repository.CreateCampaign(ctx, campaign); err != nil || !inserted {
		t.Fatal(err)
	}
	running := campaign
	running.RecordVersion = 2
	running.State = applicationEvaluationCampaignStateRunning
	running.StartedAt = "2026-08-09T08:00:01Z"
	running.Authority = &authority
	running.Items[0].State = applicationEvaluationCampaignItemRunning
	running.Items[0].StartedAt = running.StartedAt
	if _, updated, err := repository.UpdateCampaign(ctx, 1, running); err != nil || !updated {
		t.Fatal(err)
	}
	run := applicationEvaluationSucceededDefinitionRun(ctx, authority, running.Items[0].RunID)
	invokeCalls := 0
	service := newApplicationEvaluationCampaignService(
		repository,
		func(ApplicationEvaluationContext, ApplicationEvaluationPlanVersion) (ApplicationEvaluationCampaignAuthority, string) {
			return authority, ""
		},
		func(ApplicationEvaluationContext, ApplicationEvaluationPlanVersion, ApplicationEvaluationPlanItem, string) (*WorkflowRunRecord, string, string) {
			invokeCalls++
			return nil, "", ""
		},
		func(ApplicationEvaluationContext, string) (WorkflowRunRecord, bool, error) { return *run, true, nil },
	)
	result := service.Reconcile(ctx, campaign.CampaignID, 2)
	if result.Campaign == nil || result.Campaign.State != applicationEvaluationCampaignStateInterrupted || result.FailureCode != ApplicationEvaluationFailureRunUnavailable || invokeCalls != 0 {
		t.Fatalf("reconciliation must close without replay: result=%+v invokes=%d", result, invokeCalls)
	}
}

func TestApplicationEvaluationCampaignPairMaterializesExactCasesAndSuite(t *testing.T) {
	planService, ctx := newApplicationEvaluationPlanTestService(t, "workflow_copilot")
	created := planService.Create(ctx, applicationEvaluationWorkflowPlanInput("Handoff campaign"))
	if created.FailureCode != "" || created.Plan == nil || created.Version == nil {
		t.Fatalf("create handoff plan: %+v", created)
	}
	authority := applicationEvaluationWorkflowAuthority(t, ctx)
	runStore := newMemoryWorkflowRunStore(20)
	campaignService := newApplicationEvaluationCampaignService(
		planService.repository,
		func(ApplicationEvaluationContext, ApplicationEvaluationPlanVersion) (ApplicationEvaluationCampaignAuthority, string) {
			return authority, ""
		},
		func(callContext ApplicationEvaluationContext, _ ApplicationEvaluationPlanVersion, _ ApplicationEvaluationPlanItem, runID string) (*WorkflowRunRecord, string, string) {
			run := applicationEvaluationSucceededDefinitionRun(callContext, authority, runID)
			storeTerminalComparisonTestRun(t, runStore, applicationEvaluationWorkflowRunContext(callContext), run)
			return run, "", ""
		},
		func(callContext ApplicationEvaluationContext, runID string) (WorkflowRunRecord, bool, error) {
			return runStore.ReadRun(applicationEvaluationWorkflowRunContext(callContext), runID)
		},
	)
	execute := func(key string) ApplicationEvaluationCampaign {
		result := campaignService.Execute(ctx, ApplicationEvaluationCampaignExecuteInput{
			PlanID: created.Plan.PlanID, PlanVersion: created.Version.PlanVersion, PlanDigest: created.Version.PlanDigest,
			ExpectedPlanRecordVersion: created.Plan.RecordVersion, ClientCampaignKey: key, QuotaAPIKeyID: "key_aaaaaaaaaaaaaaaa",
			AcknowledgeSequentialExecution: true, AcknowledgeQuotaConsumption: true,
		})
		if result.FailureCode != "" || result.Campaign == nil || result.Campaign.State != applicationEvaluationCampaignStateSucceeded {
			t.Fatalf("execute %s: %+v", key, result)
		}
		return *result.Campaign
	}
	baseline, candidate := execute("campaign_baseline"), execute("campaign_candidate")
	evaluation := newWorkflowEvaluationService(newMemoryWorkflowEvaluationStore(10), runStore)
	evaluation.newCaseID = func() (string, error) { return "eval_handoff_case", nil }
	suite := newWorkflowEvaluationSuiteService(newMemoryWorkflowEvaluationSuiteStore(10), evaluation)
	suite.newSuiteID = func() (string, error) { return "suite_handoff_evidence", nil }
	service := newApplicationEvaluationHandoffService(planService.repository, runStore, evaluation, suite)

	preview := service.Preview(ctx, ApplicationEvaluationPairInput{BaselineCampaignID: baseline.CampaignID, CandidateCampaignID: candidate.CampaignID})
	if preview.FailureCode != "" || preview.Review == nil || len(preview.Review.Items) != 1 ||
		preview.Review.ExpectedMatches != 1 || !preview.Review.Items[0].ExpectationMatched {
		t.Fatalf("unexpected pair preview: %+v", preview)
	}
	materialized := service.Materialize(ctx, ApplicationEvaluationHandoffInput{
		BaselineCampaignID: baseline.CampaignID, CandidateCampaignID: candidate.CampaignID,
		ExpectedBaselineRecordVersion: baseline.RecordVersion, ExpectedCandidateRecordVersion: candidate.RecordVersion,
		AcknowledgeEvidenceMaterializing: true,
	})
	if materialized.FailureCode != "" || materialized.Handoff == nil || materialized.Handoff.State != "complete" ||
		materialized.Handoff.SuiteID != "suite_handoff_evidence" || len(materialized.Handoff.CaseRefs) != 1 ||
		materialized.Handoff.CaseRefs[0] != (WorkflowEvaluationSuiteCaseRef{CaseID: "eval_handoff_case", Version: 1}) ||
		materialized.CandidateCampaign == nil || materialized.CandidateCampaign.RecordVersion != candidate.RecordVersion+2 {
		t.Fatalf("unexpected materialized handoff: %+v", materialized)
	}
	replay := service.Materialize(ctx, ApplicationEvaluationHandoffInput{
		BaselineCampaignID: baseline.CampaignID, CandidateCampaignID: candidate.CampaignID,
		ExpectedBaselineRecordVersion: baseline.RecordVersion, ExpectedCandidateRecordVersion: candidate.RecordVersion,
		AcknowledgeEvidenceMaterializing: true,
	})
	if replay.FailureCode != "" || !replay.IdempotentReplay || replay.Handoff == nil || replay.Handoff.SuiteID != materialized.Handoff.SuiteID {
		t.Fatalf("handoff replay created new evidence or failed: %+v", replay)
	}
}

func TestApplicationEvaluationCampaignHandoffPersistsPartialCaseRefs(t *testing.T) {
	planService, ctx := newApplicationEvaluationPlanTestService(t, "workflow_copilot")
	created := planService.Create(ctx, applicationEvaluationWorkflowPlanInput("Partial handoff"))
	authority := applicationEvaluationWorkflowAuthority(t, ctx)
	runStore := newMemoryWorkflowRunStore(20)
	campaignService := newApplicationEvaluationCampaignService(
		planService.repository,
		func(ApplicationEvaluationContext, ApplicationEvaluationPlanVersion) (ApplicationEvaluationCampaignAuthority, string) {
			return authority, ""
		},
		func(callContext ApplicationEvaluationContext, _ ApplicationEvaluationPlanVersion, _ ApplicationEvaluationPlanItem, runID string) (*WorkflowRunRecord, string, string) {
			run := applicationEvaluationSucceededDefinitionRun(callContext, authority, runID)
			storeTerminalComparisonTestRun(t, runStore, applicationEvaluationWorkflowRunContext(callContext), run)
			return run, "", ""
		},
		func(callContext ApplicationEvaluationContext, runID string) (WorkflowRunRecord, bool, error) {
			return runStore.ReadRun(applicationEvaluationWorkflowRunContext(callContext), runID)
		},
	)
	execute := func(key string) ApplicationEvaluationCampaign {
		result := campaignService.Execute(ctx, ApplicationEvaluationCampaignExecuteInput{
			PlanID: created.Plan.PlanID, PlanVersion: 1, PlanDigest: created.Version.PlanDigest,
			ExpectedPlanRecordVersion: 1, ClientCampaignKey: key, QuotaAPIKeyID: "key_aaaaaaaaaaaaaaaa",
			AcknowledgeSequentialExecution: true, AcknowledgeQuotaConsumption: true,
		})
		return *result.Campaign
	}
	baseline, candidate := execute("partial_baseline"), execute("partial_candidate")
	evaluation := newWorkflowEvaluationService(newMemoryWorkflowEvaluationStore(10), runStore)
	evaluation.newCaseID = func() (string, error) { return "eval_partial_case", nil }
	suiteStore := &failingApplicationEvaluationSuiteStore{workflowEvaluationSuiteStore: newMemoryWorkflowEvaluationSuiteStore(10)}
	suite := newWorkflowEvaluationSuiteService(suiteStore, evaluation)
	service := newApplicationEvaluationHandoffService(planService.repository, runStore, evaluation, suite)
	result := service.Materialize(ctx, ApplicationEvaluationHandoffInput{
		BaselineCampaignID: baseline.CampaignID, CandidateCampaignID: candidate.CampaignID,
		ExpectedBaselineRecordVersion: baseline.RecordVersion, ExpectedCandidateRecordVersion: candidate.RecordVersion,
		AcknowledgeEvidenceMaterializing: true,
	})
	if result.FailureCode != ApplicationEvaluationFailureHandoffPartial || result.Handoff == nil || result.Handoff.State != "partial" ||
		len(result.Handoff.CaseRefs) != 1 || result.Handoff.SuiteID != "" {
		t.Fatalf("partial handoff lost exact refs: %+v", result)
	}
	stored, found, err := planService.repository.ReadCampaign(ctx, candidate.CampaignID)
	if err != nil || !found || stored.Handoff == nil || stored.Handoff.State != "partial" || len(stored.Handoff.CaseRefs) != 1 {
		t.Fatalf("partial handoff was not durable: found=%v err=%v campaign=%+v", found, err, stored)
	}
}

type failingApplicationEvaluationSuiteStore struct {
	workflowEvaluationSuiteStore
}

func (store *failingApplicationEvaluationSuiteStore) CreateSuite(WorkflowRunContext, WorkflowEvaluationSuite) error {
	return errWorkflowEvaluationSuiteStoreContract
}

func applicationEvaluationSucceededDefinitionRun(ctx ApplicationEvaluationContext, authority ApplicationEvaluationCampaignAuthority, runID string) *WorkflowRunRecord {
	var snapshot ApplicationInteractionAuthoritySnapshot
	_ = json.Unmarshal(authority.Snapshot, &snapshot)
	record := workflowDefinitionRunRecordForStoreTest(applicationEvaluationWorkflowRunContext(ctx), runID)
	record.ExecutionSourceID, record.ExecutionSourceVersion = snapshot.WorkflowDefinition.DefinitionID, snapshot.WorkflowDefinition.DefinitionVersion
	record.ExecutionSource.ID, record.ExecutionSource.Version = record.ExecutionSourceID, record.ExecutionSourceVersion
	record.DefinitionAuthority.DefinitionID, record.DefinitionAuthority.DefinitionVersion = record.ExecutionSourceID, record.ExecutionSourceVersion
	record.DefinitionAuthority.DefinitionDigest = snapshot.WorkflowDefinition.DefinitionDigest
	record.DefinitionAuthority.SourceDraftDigest = snapshot.WorkflowDefinition.DefinitionDigest
	record.DefinitionAuthority.ActivationPointerVersion = snapshot.WorkflowDefinition.ActivationPointerVersion
	record.DefinitionAuthority.ApplicationRecordVersion = snapshot.ApplicationRecordVersion
	record.DefinitionAuthority.ApplicationLifecycle = snapshot.ApplicationLifecycle
	record.Status = WorkflowRunStatusSucceeded
	record.CompletedAt = workflowRunTimestamp(time.Now().UTC())
	record.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
	return &record
}

func newApplicationEvaluationPlanTestService(t *testing.T, applicationKind string) (applicationEvaluationPlanService, ApplicationEvaluationContext) {
	t.Helper()
	catalogRepository := newMemoryApplicationCatalogRepository()
	catalogService := newApplicationCatalogService(catalogRepository)
	catalogContext := ApplicationCatalogContext{
		RequestContext: context.Background(), RequestID: "request-catalog", TenantRef: "tenant-one",
		WorkspaceID: "workspace-one", ActorRef: "subject-owner", OwnerSubjectRef: "subject-owner",
		AuditRef: "audit-catalog", WriteEnabled: true,
	}
	created := catalogService.Create(catalogContext, ApplicationCatalogCreateInput{DisplayName: "Evaluation application", ApplicationKind: applicationKind})
	if created.FailureCode != "" || created.Record == nil {
		t.Fatalf("create catalog application: %+v", created)
	}
	ctx := applicationEvaluationTestContext()
	ctx.ApplicationID = created.Record.ApplicationID
	service := newApplicationEvaluationPlanService(newMemoryApplicationEvaluationRepository(20, 20), catalogRepository)
	return service, ctx
}

func applicationEvaluationTestContext() ApplicationEvaluationContext {
	return ApplicationEvaluationContext{
		RequestContext: context.Background(), RequestID: "request-evaluation", TenantRef: "tenant-one",
		WorkspaceID: "workspace-one", Environment: applicationEvaluationEnvironmentTest,
		ApplicationID: "app_aaaaaaaaaaaaaaaa", ActorRef: "subject-owner", AuditRef: "audit-evaluation", WriteEnabled: true,
	}
}

func applicationEvaluationPromptPlanInput(name string, classification WorkflowRunComparisonClassification) ApplicationEvaluationPlanCreateInput {
	return ApplicationEvaluationPlanCreateInput{
		Name: name, ExecutionProfile: applicationInteractionProfilePrompt,
		Items: []ApplicationEvaluationPlanItem{applicationEvaluationPromptItem("first", "First prompt", classification)},
	}
}

func applicationEvaluationPromptItem(key, name string, classification WorkflowRunComparisonClassification) ApplicationEvaluationPlanItem {
	return ApplicationEvaluationPlanItem{
		ItemKey: key, Name: name, ExpectedClassification: classification,
		PromptApplication: &ApplicationEvaluationPromptFixture{Variables: map[string]any{"topic": "radish", "count": 2}},
	}
}

func applicationEvaluationPendingCampaign(ctx ApplicationEvaluationContext, clientKey, now string) ApplicationEvaluationCampaign {
	campaignID := applicationEvaluationDeterministicCampaignID(ctx, clientKey)
	return ApplicationEvaluationCampaign{
		SchemaVersion: applicationEvaluationCampaignSchemaVersion, CampaignID: campaignID,
		ClientCampaignKey: clientKey, RecordVersion: 1,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Environment: ctx.Environment, ApplicationID: ctx.ApplicationID,
		PlanID: "aeplan_aaaaaaaaaaaaaaaa", PlanVersion: 1,
		PlanDigest: "sha256:" + strings.Repeat("a", 64), ExecutionProfile: applicationInteractionProfileWorkflow,
		QuotaAPIKeyID: "key_aaaaaaaaaaaaaaaa",
		State:         applicationEvaluationCampaignStatePending,
		Items: []ApplicationEvaluationCampaignItem{{
			ItemKey: "first", RunID: applicationEvaluationDeterministicRunID(campaignID, "first"), State: applicationEvaluationCampaignItemPending,
		}},
		CreatedAt: now, CreatedByActorRef: ctx.ActorRef, UpdatedByActorRef: ctx.ActorRef,
		RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
}

func applicationEvaluationWorkflowAuthority(t *testing.T, ctx ApplicationEvaluationContext) ApplicationEvaluationCampaignAuthority {
	t.Helper()
	snapshot := ApplicationInteractionAuthoritySnapshot{
		SchemaVersion: applicationInteractionAuthoritySchemaVersion, ExecutionProfile: applicationInteractionProfileWorkflow,
		ApplicationID: ctx.ApplicationID, ApplicationRecordVersion: 1, ApplicationLifecycle: applicationCatalogLifecycleActive,
		WorkflowDefinition: &ApplicationInteractionWorkflowAuthority{
			DefinitionID: "definition-one", DefinitionVersion: 1,
			DefinitionDigest: "sha256:" + strings.Repeat("b", 64), ActivationPointerVersion: 1,
			CandidateID: "candidate-one", CandidateReviewVersion: 1,
		},
	}
	var err error
	snapshot.AuthorityDigest, err = applicationInteractionAuthorityDigest(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	return ApplicationEvaluationCampaignAuthority{ExecutionProfile: snapshot.ExecutionProfile, AuthorityDigest: snapshot.AuthorityDigest, Snapshot: payload}
}
