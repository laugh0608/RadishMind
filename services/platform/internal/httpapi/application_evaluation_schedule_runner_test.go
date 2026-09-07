package httpapi

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"
)

func TestApplicationEvaluationScheduleDelegatedUserID(t *testing.T) {
	if userID, ok := applicationEvaluationScheduleDelegatedUserID("user:usr_aaaaaaaaaaaaaaaa"); !ok || userID != "usr_aaaaaaaaaaaaaaaa" {
		t.Fatalf("parse delegated local user: id=%q ok=%v", userID, ok)
	}
	for _, actorRef := range []string{"usr_aaaaaaaaaaaaaaaa", "user:workspace-admin", "system:application-evaluation-scheduler", "user:usr_bad"} {
		if userID, ok := applicationEvaluationScheduleDelegatedUserID(actorRef); ok {
			t.Fatalf("accepted invalid delegated actor %q as %q", actorRef, userID)
		}
	}
}

func TestApplicationEvaluationScheduleOccurrenceRevalidationRequiresCurrentMembership(t *testing.T) {
	ctx, schedule, version := applicationEvaluationScheduleTestRecords(t)
	const userID = "usr_aaaaaaaaaaaaaaaa"
	ctx.ActorRef = "user:" + userID
	version.Authorization.DelegatedByUserRef = ctx.ActorRef
	version.CreatedByActorRef = ctx.ActorRef
	version.ScheduleDigest = ""
	var err error
	version.ScheduleDigest, err = applicationEvaluationScheduleDigest(version)
	if err != nil {
		t.Fatal(err)
	}
	schedule.LatestScheduleDigest = version.ScheduleDigest
	schedule.DelegatedByUserRef = ctx.ActorRef
	schedule.CreatedByActorRef = ctx.ActorRef
	schedule.UpdatedByActorRef = ctx.ActorRef

	schedules := newMemoryApplicationEvaluationScheduleRepository(4, 4)
	if err = schedules.CreateSchedule(ctx, schedule, version); err != nil {
		t.Fatalf("create delegated schedule: %v", err)
	}
	schedule = applicationEvaluationActivateScheduleRepository(t, schedules, ctx, schedule, "2026-08-30T09:30:00Z")
	identities := newMemoryLocalIdentityRepository()
	account, credential := localIdentityTestAccount(userID, "cred_aaaaaaaaaaaaaaaa", "schedule-owner@example.com", time.Now().UTC())
	if err = identities.CreateAccount(context.Background(), account, credential); err != nil {
		t.Fatalf("create delegated local account: %v", err)
	}
	scheduledFor := *schedule.NextDueAt
	delegated := ctx
	delegated.ScheduleExecution = &ApplicationEvaluationScheduleExecutionRef{
		AuthorizationModel: applicationEvaluationScheduleAuthorizationModel,
		ScheduleID:         version.ScheduleID,
		ScheduleVersion:    version.ScheduleVersion,
		ScheduleDigest:     version.ScheduleDigest,
		ScheduledForUTC:    scheduledFor,
		ClientCampaignKey:  applicationEvaluationScheduleClientCampaignKey(version.ScheduleID, version.ScheduleVersion, scheduledFor),
		SystemActorRef:     version.Authorization.SystemActorRef,
		DelegatedByUserRef: version.Authorization.DelegatedByUserRef,
	}
	server := &Server{applicationEvaluationScheduleRepository: schedules, localIdentityRepository: identities}
	if failure := server.revalidateApplicationEvaluationScheduleOccurrence(context.Background(), delegated, version); failure != ApplicationEvaluationScheduleFailureMembershipDenied {
		t.Fatalf("missing current membership must fail closed: %s", failure)
	}
}

func TestApplicationEvaluationScheduleRunnerCreatesOneCampaignAndProjectsNextDue(t *testing.T) {
	fixture := newApplicationEvaluationScheduleRunnerFixture(t, "aesch_cccccccccccccccc")
	var invocationCalls atomic.Int32
	var observedExecution *ApplicationEvaluationScheduleExecutionRef
	fixture.runner.campaigns.invoke = func(
		ctx ApplicationEvaluationContext,
		_ ApplicationEvaluationPlanVersion,
		_ ApplicationEvaluationPlanItem,
		runID string,
	) (*WorkflowRunRecord, string, string) {
		invocationCalls.Add(1)
		observedExecution = cloneApplicationEvaluationScheduleExecutionRef(ctx.ScheduleExecution)
		return applicationEvaluationScheduleRunnerSucceededRun(ctx, fixture.authority, runID), "", ""
	}

	if err := fixture.runner.PollOnce(context.Background()); err != nil {
		t.Fatalf("poll due schedule: %v", err)
	}
	if invocationCalls.Load() != 1 {
		t.Fatalf("runner must invoke one existing Campaign item, calls=%d", invocationCalls.Load())
	}
	if observedExecution == nil || observedExecution.SystemActorRef != applicationEvaluationScheduleSystemActorRef ||
		observedExecution.DelegatedByUserRef != fixture.ctx.ActorRef || observedExecution.ScheduleID != fixture.schedule.ScheduleID {
		t.Fatalf("run metadata did not preserve system/delegated execution: %+v", observedExecution)
	}
	current, found, err := fixture.repository.ReadSchedule(fixture.ctx, fixture.schedule.ScheduleID)
	if err != nil || !found || current.RecordVersion != fixture.schedule.RecordVersion+1 || current.NextDueAt == nil ||
		*current.NextDueAt != "2026-08-31T09:30:00Z" {
		t.Fatalf("next due projection is not exact: found=%v err=%v schedule=%+v", found, err, current)
	}
	occurrence, found, err := fixture.repository.ReadOccurrence(
		fixture.ctx, fixture.schedule.ScheduleID, fixture.version.ScheduleVersion, *fixture.schedule.NextDueAt,
	)
	if err != nil || !found || occurrence.State != applicationEvaluationScheduleOccurrenceStateSucceeded || occurrence.CampaignID == nil {
		t.Fatalf("terminal occurrence evidence missing: found=%v err=%v occurrence=%+v", found, err, occurrence)
	}
	if err = fixture.runner.PollOnce(context.Background()); err != nil || invocationCalls.Load() != 1 {
		t.Fatalf("repeated poll replayed provider execution: err=%v calls=%d", err, invocationCalls.Load())
	}
}

func TestApplicationEvaluationScheduleRunnerConcurrentClaimHasOneProviderPath(t *testing.T) {
	fixture := newApplicationEvaluationScheduleRunnerFixture(t, "aesch_dddddddddddddddd")
	var invocationCalls atomic.Int32
	fixture.runner.campaigns.invoke = func(
		ctx ApplicationEvaluationContext,
		_ ApplicationEvaluationPlanVersion,
		_ ApplicationEvaluationPlanItem,
		runID string,
	) (*WorkflowRunRecord, string, string) {
		invocationCalls.Add(1)
		return applicationEvaluationScheduleRunnerSucceededRun(ctx, fixture.authority, runID), "", ""
	}
	runners := make([]*applicationEvaluationScheduleRunner, 16)
	for index := range runners {
		runners[index] = newApplicationEvaluationScheduleRunner(fixture.repository, fixture.runner.campaigns, fixture.runner.revalidate)
		runners[index].now = fixture.runner.now
	}
	errorsByRunner := make(chan error, len(runners))
	for _, runner := range runners {
		go func(candidate *applicationEvaluationScheduleRunner) {
			errorsByRunner <- candidate.PollOnce(context.Background())
		}(runner)
	}
	for range runners {
		if err := <-errorsByRunner; err != nil {
			t.Fatalf("concurrent runner failed: %v", err)
		}
	}
	if invocationCalls.Load() != 1 {
		t.Fatalf("concurrent runners created duplicate provider paths: %d", invocationCalls.Load())
	}
}

func TestApplicationEvaluationScheduleRunnerRecoveryNeverReplaysMissingOrExistingCampaign(t *testing.T) {
	for _, testCase := range []struct {
		name           string
		createCampaign bool
		wantState      string
		wantInvokes    int32
	}{
		{name: "crash before Campaign create", wantState: applicationEvaluationScheduleOccurrenceStateInterrupted},
		{name: "crash after Campaign create", createCampaign: true, wantState: applicationEvaluationScheduleOccurrenceStateSucceeded, wantInvokes: 1},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newApplicationEvaluationScheduleRunnerFixture(t, "aesch_eeeeeeeeeeeeeeee")
			var invocationCalls atomic.Int32
			fixture.runner.campaigns.invoke = func(
				ctx ApplicationEvaluationContext,
				_ ApplicationEvaluationPlanVersion,
				_ ApplicationEvaluationPlanItem,
				runID string,
			) (*WorkflowRunRecord, string, string) {
				invocationCalls.Add(1)
				return applicationEvaluationScheduleRunnerSucceededRun(ctx, fixture.authority, runID), "", ""
			}
			systemCtx := applicationEvaluationScheduleRunnerContext(context.Background(), fixture.schedule, *fixture.schedule.NextDueAt)
			due, claimed := applicationEvaluationScheduleRunnerClaim(systemCtx, fixture.schedule, fixture.version, fixture.runner.currentTime())
			stored, won, err := fixture.repository.ClaimOccurrence(systemCtx, due, claimed)
			if err != nil || !won {
				t.Fatalf("seed claimed occurrence: won=%v err=%v", won, err)
			}
			if testCase.createCampaign {
				if err = fixture.runner.advanceProjection(systemCtx, fixture.schedule, fixture.version, fixture.runner.currentTime()); err != nil {
					t.Fatalf("advance projection before simulated crash: %v", err)
				}
				ctx := applicationEvaluationScheduleRunnerDelegatedContext(context.Background(), stored)
				result := fixture.runner.campaigns.Execute(ctx, ApplicationEvaluationCampaignExecuteInput{
					PlanID: fixture.version.PlanID, PlanVersion: fixture.version.PlanVersion, PlanDigest: fixture.version.PlanDigest,
					ExpectedPlanRecordVersion: fixture.plan.RecordVersion, ClientCampaignKey: stored.ClientCampaignKey,
					QuotaAPIKeyID: fixture.version.QuotaAPIKeyID, AcknowledgeSequentialExecution: true, AcknowledgeQuotaConsumption: true,
				})
				if result.Campaign == nil || result.Campaign.State != applicationEvaluationCampaignStateSucceeded {
					t.Fatalf("seed exact Campaign before crash: %+v", result)
				}
			}
			fixture.runner.recoveryAfter = 0
			fixture.runner.now = func() time.Time { return time.Date(2026, 8, 30, 9, 32, 0, 0, time.UTC) }
			if err = fixture.runner.PollOnce(context.Background()); err != nil {
				t.Fatalf("recover claimed occurrence: %v", err)
			}
			recovered, found, readErr := fixture.repository.ReadOccurrence(
				fixture.ctx, stored.ScheduleID, stored.ScheduleVersion, stored.ScheduledForUTC,
			)
			if readErr != nil || !found || recovered.State != testCase.wantState || invocationCalls.Load() != testCase.wantInvokes {
				t.Fatalf("recovery drifted: found=%v err=%v occurrence=%+v invokes=%d", found, readErr, recovered, invocationCalls.Load())
			}
			current, found, readErr := fixture.repository.ReadSchedule(fixture.ctx, fixture.schedule.ScheduleID)
			if readErr != nil || !found || current.NextDueAt == nil || *current.NextDueAt != "2026-08-31T09:30:00Z" {
				t.Fatalf("recovery did not preserve exact next projection: found=%v err=%v schedule=%+v", found, readErr, current)
			}
		})
	}
}

func TestApplicationEvaluationScheduleRunnerRevalidationAndQuotaFailClosed(t *testing.T) {
	for _, testCase := range []struct {
		name            string
		revalidation    string
		campaignFailure string
		wantFailure     string
		wantInvokes     int32
	}{
		{name: "membership revoked", revalidation: ApplicationEvaluationScheduleFailureMembershipDenied, wantFailure: ApplicationEvaluationScheduleFailureMembershipDenied},
		{name: "authority drift", revalidation: ApplicationEvaluationScheduleFailureAuthorityChanged, wantFailure: ApplicationEvaluationScheduleFailureAuthorityChanged},
		{name: "API key rotated", revalidation: ApplicationEvaluationScheduleFailureQuotaConsumerInvalid, wantFailure: ApplicationEvaluationScheduleFailureQuotaConsumerInvalid},
		{name: "quota denied", campaignFailure: GatewayRequestQuotaFailureExceeded, wantFailure: ApplicationEvaluationScheduleFailureQuotaDenied, wantInvokes: 1},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newApplicationEvaluationScheduleRunnerFixture(t, "aesch_ffffffffffffffff")
			fixture.runner.revalidate = func(context.Context, ApplicationEvaluationContext, ApplicationEvaluationScheduleVersion) string {
				return testCase.revalidation
			}
			var invocationCalls atomic.Int32
			fixture.runner.campaigns.invoke = func(
				ctx ApplicationEvaluationContext,
				_ ApplicationEvaluationPlanVersion,
				_ ApplicationEvaluationPlanItem,
				runID string,
			) (*WorkflowRunRecord, string, string) {
				invocationCalls.Add(1)
				run := applicationEvaluationScheduleRunnerSucceededRun(ctx, fixture.authority, runID)
				if testCase.campaignFailure != "" {
					run.Status = WorkflowRunStatusFailed
					run.FailureCode = WorkflowRunFailureCode(testCase.campaignFailure)
					run.FailureSummary = "quota denied before provider"
					return run, testCase.campaignFailure, run.FailureSummary
				}
				return run, "", ""
			}
			if err := fixture.runner.PollOnce(context.Background()); err != nil {
				t.Fatalf("poll drifted occurrence: %v", err)
			}
			occurrence, found, err := fixture.repository.ReadOccurrence(
				fixture.ctx, fixture.schedule.ScheduleID, fixture.version.ScheduleVersion, *fixture.schedule.NextDueAt,
			)
			if err != nil || !found || occurrence.State != applicationEvaluationScheduleOccurrenceStateFailed ||
				occurrence.FailureCode == nil || *occurrence.FailureCode != testCase.wantFailure || invocationCalls.Load() != testCase.wantInvokes {
				t.Fatalf("drift did not fail closed: found=%v err=%v occurrence=%+v invokes=%d", found, err, occurrence, invocationCalls.Load())
			}
		})
	}
}

func TestApplicationEvaluationScheduleRunnerMissedAndOverlapNeverInvokeCampaign(t *testing.T) {
	t.Run("missed window", func(t *testing.T) {
		fixture := newApplicationEvaluationScheduleRunnerFixture(t, "aesch_gggggggggggggggg")
		var invocationCalls atomic.Int32
		fixture.runner.campaigns.invoke = func(ApplicationEvaluationContext, ApplicationEvaluationPlanVersion, ApplicationEvaluationPlanItem, string) (*WorkflowRunRecord, string, string) {
			invocationCalls.Add(1)
			return nil, "", ""
		}
		fixture.runner.now = func() time.Time { return time.Date(2026, 8, 31, 9, 31, 0, 0, time.UTC) }
		if err := fixture.runner.PollOnce(context.Background()); err != nil {
			t.Fatalf("poll missed occurrence: %v", err)
		}
		occurrence, found, err := fixture.repository.ReadOccurrence(
			fixture.ctx, fixture.schedule.ScheduleID, fixture.version.ScheduleVersion, *fixture.schedule.NextDueAt,
		)
		if err != nil || !found || occurrence.State != applicationEvaluationScheduleOccurrenceStateSkipped ||
			occurrence.FailureCode == nil || *occurrence.FailureCode != ApplicationEvaluationScheduleFailureMissedWindow || invocationCalls.Load() != 0 {
			t.Fatalf("missed occurrence did not skip before Campaign: found=%v err=%v occurrence=%+v calls=%d", found, err, occurrence, invocationCalls.Load())
		}
	})

	t.Run("overlap", func(t *testing.T) {
		fixture := newApplicationEvaluationScheduleRunnerFixture(t, "aesch_hhhhhhhhhhhhhhhh")
		var invocationCalls atomic.Int32
		fixture.runner.campaigns.invoke = func(ApplicationEvaluationContext, ApplicationEvaluationPlanVersion, ApplicationEvaluationPlanItem, string) (*WorkflowRunRecord, string, string) {
			invocationCalls.Add(1)
			return nil, "", ""
		}
		systemCtx := applicationEvaluationScheduleRunnerContext(context.Background(), fixture.schedule, *fixture.schedule.NextDueAt)
		due, claimed := applicationEvaluationScheduleRunnerClaim(systemCtx, fixture.schedule, fixture.version, fixture.runner.currentTime())
		if _, won, err := fixture.repository.ClaimOccurrence(systemCtx, due, claimed); err != nil || !won {
			t.Fatalf("seed prior open occurrence: won=%v err=%v", won, err)
		}
		if err := fixture.runner.advanceProjection(systemCtx, fixture.schedule, fixture.version, fixture.runner.currentTime()); err != nil {
			t.Fatalf("advance to overlapping window: %v", err)
		}
		fixture.runner.recoveryAfter = 48 * time.Hour
		fixture.runner.now = func() time.Time { return time.Date(2026, 8, 31, 9, 31, 0, 0, time.UTC) }
		if err := fixture.runner.PollOnce(context.Background()); err != nil {
			t.Fatalf("poll overlapping occurrence: %v", err)
		}
		occurrence, found, err := fixture.repository.ReadOccurrence(
			fixture.ctx, fixture.schedule.ScheduleID, fixture.version.ScheduleVersion, "2026-08-31T09:30:00Z",
		)
		if err != nil || !found || occurrence.State != applicationEvaluationScheduleOccurrenceStateSkipped ||
			occurrence.FailureCode == nil || *occurrence.FailureCode != ApplicationEvaluationScheduleFailureOverlapBlocked || invocationCalls.Load() != 0 {
			t.Fatalf("overlap did not skip before Campaign: found=%v err=%v occurrence=%+v calls=%d", found, err, occurrence, invocationCalls.Load())
		}
	})
}

func TestApplicationEvaluationScheduleRunnerIgnoresPausedAndArchivedSchedules(t *testing.T) {
	for _, state := range []string{applicationEvaluationScheduleStatePaused, applicationEvaluationScheduleStateArchived} {
		t.Run(state, func(t *testing.T) {
			fixture := newApplicationEvaluationScheduleRunnerFixture(t, "aesch_jjjjjjjjjjjjjjjj")
			current := cloneApplicationEvaluationSchedule(fixture.schedule)
			current.RecordVersion++
			current.LifecycleState = state
			current.NextDueAt = nil
			current.UpdatedAt = "2026-08-30T09:00:00Z"
			current.RequestID = "request-schedule-stop"
			current.AuditRef = "audit:schedule-stop"
			updateContext := fixture.ctx
			updateContext.RequestID = current.RequestID
			updateContext.AuditRef = current.AuditRef
			if _, updated, err := fixture.repository.UpdateSchedule(updateContext, fixture.schedule.RecordVersion, current); err != nil || !updated {
				t.Fatalf("move Schedule to %s: updated=%v err=%v", state, updated, err)
			}
			var invocationCalls atomic.Int32
			fixture.runner.campaigns.invoke = func(ApplicationEvaluationContext, ApplicationEvaluationPlanVersion, ApplicationEvaluationPlanItem, string) (*WorkflowRunRecord, string, string) {
				invocationCalls.Add(1)
				return nil, "", ""
			}
			if err := fixture.runner.PollOnce(context.Background()); err != nil || invocationCalls.Load() != 0 {
				t.Fatalf("%s Schedule reached Campaign: err=%v calls=%d", state, err, invocationCalls.Load())
			}
			if _, found, err := fixture.repository.ReadOccurrence(
				fixture.ctx, fixture.schedule.ScheduleID, fixture.version.ScheduleVersion, *fixture.schedule.NextDueAt,
			); err != nil || found {
				t.Fatalf("%s Schedule created occurrence: found=%v err=%v", state, found, err)
			}
		})
	}
}

func TestApplicationEvaluationScheduleRunnerStoreOutageStopsClaimAndRecoversWithoutFallback(t *testing.T) {
	fixture := newApplicationEvaluationScheduleRunnerFixture(t, "aesch_kkkkkkkkkkkkkkkk")
	var invocationCalls atomic.Int32
	fixture.runner.campaigns.invoke = func(
		ctx ApplicationEvaluationContext,
		_ ApplicationEvaluationPlanVersion,
		_ ApplicationEvaluationPlanItem,
		runID string,
	) (*WorkflowRunRecord, string, string) {
		invocationCalls.Add(1)
		return applicationEvaluationScheduleRunnerSucceededRun(ctx, fixture.authority, runID), "", ""
	}
	fixture.repository.mu.Lock()
	fixture.repository.unavailable = true
	fixture.repository.mu.Unlock()
	if err := fixture.runner.PollOnce(context.Background()); err == nil || invocationCalls.Load() != 0 {
		t.Fatalf("store outage must stop claim: err=%v calls=%d", err, invocationCalls.Load())
	}
	fixture.repository.mu.Lock()
	fixture.repository.unavailable = false
	fixture.repository.mu.Unlock()
	if err := fixture.runner.PollOnce(context.Background()); err != nil || invocationCalls.Load() != 1 {
		t.Fatalf("runner did not resume exact store after recovery: err=%v calls=%d", err, invocationCalls.Load())
	}
}

func TestApplicationEvaluationScheduleRunnerCancelJoin(t *testing.T) {
	fixture := newApplicationEvaluationScheduleRunnerFixture(t, "aesch_bbbbbbbbbbbbbbbb")
	fixture.runner.pollInterval = time.Hour
	fixture.runner.now = func() time.Time { return time.Date(2026, 8, 30, 8, 30, 0, 0, time.UTC) }
	if err := fixture.runner.Start(context.Background()); err != nil {
		t.Fatalf("start runner: %v", err)
	}
	done := make(chan struct{})
	go func() {
		fixture.runner.Stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("runner cancel did not join worker")
	}
	if err := fixture.runner.Start(context.Background()); err == nil {
		t.Fatal("runner started twice")
	}
}

type applicationEvaluationScheduleRunnerFixture struct {
	ctx        ApplicationEvaluationContext
	plan       ApplicationEvaluationPlan
	version    ApplicationEvaluationScheduleVersion
	schedule   ApplicationEvaluationSchedule
	repository *memoryApplicationEvaluationScheduleRepository
	authority  ApplicationEvaluationCampaignAuthority
	runner     *applicationEvaluationScheduleRunner
}

func newApplicationEvaluationScheduleRunnerFixture(t *testing.T, scheduleID string) applicationEvaluationScheduleRunnerFixture {
	t.Helper()
	planService, ctx := newApplicationEvaluationPlanTestService(t, "prompt_application")
	planService.newPlanID = func() (string, error) { return "aeplan_cccccccccccccccc", nil }
	createdPlan := planService.Create(ctx, applicationEvaluationPromptPlanInput("Scheduled runner", WorkflowRunComparisonUnchanged))
	if createdPlan.FailureCode != "" || createdPlan.Plan == nil || createdPlan.Version == nil {
		t.Fatalf("create runner Plan: %+v", createdPlan)
	}
	repository := newMemoryApplicationEvaluationScheduleRepository(20, 100)
	scheduleService := newApplicationEvaluationScheduleService(
		repository, planService.repository,
		func(ApplicationEvaluationContext, ApplicationEvaluationPlanVersion) string { return "" },
		func(ApplicationEvaluationContext, string) string { return "" },
	)
	scheduleService.now = func() time.Time { return time.Date(2026, 8, 30, 8, 0, 0, 0, time.UTC) }
	scheduleService.newScheduleID = func() (string, error) { return scheduleID, nil }
	created := scheduleService.Create(ctx, ApplicationEvaluationScheduleCreateInput{ApplicationEvaluationScheduleDefinitionInput: ApplicationEvaluationScheduleDefinitionInput{
		PlanID: createdPlan.Plan.PlanID, PlanVersion: createdPlan.Version.PlanVersion, PlanDigest: createdPlan.Version.PlanDigest,
		ExpectedPlanRecordVersion: createdPlan.Plan.RecordVersion, QuotaAPIKeyID: "key_cccccccccccccccc",
		Schedule:                       ApplicationEvaluationScheduleDailyUTC{Rule: applicationEvaluationScheduleRuleDailyUTC, Hour: 9, Minute: 30},
		AcknowledgeProviderConsumption: true,
	}})
	if created.FailureCode != "" || created.Schedule == nil || created.Version == nil {
		t.Fatalf("create runner Schedule: %+v", created)
	}
	active := scheduleService.Activate(ctx, created.Schedule.ScheduleID, ApplicationEvaluationScheduleLifecycleInput{
		ExpectedVersion: created.Schedule.RecordVersion, AcknowledgeProviderConsumption: true,
	})
	if active.FailureCode != "" || active.Schedule == nil || active.Schedule.NextDueAt == nil {
		t.Fatalf("activate runner Schedule: %+v", active)
	}
	contracts := promptApplicationVNextContractFixtures(t)
	authoritySnapshot := contracts[promptApplicationRuntimeAuthorityV2Schema].(PromptApplicationRuntimeAuthorityV2)
	authoritySnapshot.ApplicationID = ctx.ApplicationID
	authoritySnapshot.AuthorityDigest = ""
	digest, err := promptApplicationRuntimeAuthorityV2Digest(authoritySnapshot)
	if err != nil {
		t.Fatalf("digest runner authority: %v", err)
	}
	authoritySnapshot.AuthorityDigest = digest
	authorityPayload, err := json.Marshal(authoritySnapshot)
	if err != nil {
		t.Fatalf("marshal runner authority: %v", err)
	}
	authority := ApplicationEvaluationCampaignAuthority{
		ExecutionProfile: applicationInteractionProfilePrompt, AuthorityDigest: authoritySnapshot.AuthorityDigest, Snapshot: authorityPayload,
	}
	campaignService := newApplicationEvaluationCampaignService(
		planService.repository,
		func(ApplicationEvaluationContext, ApplicationEvaluationPlanVersion) (ApplicationEvaluationCampaignAuthority, string) {
			return authority, ""
		},
		func(ApplicationEvaluationContext, ApplicationEvaluationPlanVersion, ApplicationEvaluationPlanItem, string) (*WorkflowRunRecord, string, string) {
			return nil, ApplicationEvaluationFailureRunUnavailable, "runner invocation was not configured"
		},
		func(ApplicationEvaluationContext, string) (WorkflowRunRecord, bool, error) {
			return WorkflowRunRecord{}, false, nil
		},
	)
	campaignService.now = func() time.Time { return time.Date(2026, 8, 30, 9, 31, 0, 0, time.UTC) }
	runner := newApplicationEvaluationScheduleRunner(
		repository, campaignService,
		func(_ context.Context, delegated ApplicationEvaluationContext, version ApplicationEvaluationScheduleVersion) string {
			if delegated.ActorRef != version.Authorization.DelegatedByUserRef || delegated.ScheduleExecution == nil ||
				delegated.ScheduleExecution.SystemActorRef != version.Authorization.SystemActorRef {
				return ApplicationEvaluationScheduleFailureAuthorizationUnavailable
			}
			return ""
		},
	)
	runner.now = func() time.Time { return time.Date(2026, 8, 30, 9, 31, 0, 0, time.UTC) }
	return applicationEvaluationScheduleRunnerFixture{
		ctx: ctx, plan: *createdPlan.Plan, version: *created.Version, schedule: *active.Schedule,
		repository: repository, authority: authority, runner: runner,
	}
}

func applicationEvaluationScheduleRunnerSucceededRun(
	ctx ApplicationEvaluationContext,
	authority ApplicationEvaluationCampaignAuthority,
	runID string,
) *WorkflowRunRecord {
	var snapshot PromptApplicationRuntimeAuthorityV2
	_ = json.Unmarshal(authority.Snapshot, &snapshot)
	return &WorkflowRunRecord{
		SchemaVersion: workflowRunRecordPromptSchemaVersion, RunID: runID,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		ExecutionProfile: applicationInteractionProfilePrompt, PromptApplication: &snapshot,
		Status: WorkflowRunStatusSucceeded, CompletedAt: "2026-08-30T09:31:01Z", ActorRef: ctx.ActorRef,
		ScheduleExecution: cloneApplicationEvaluationScheduleExecutionRef(ctx.ScheduleExecution),
	}
}
