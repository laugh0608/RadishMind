package httpapi

import (
	"testing"
	"time"
)

func TestApplicationEvaluationScheduleServiceStrictLifecycleAndList(t *testing.T) {
	planService, ctx := newApplicationEvaluationPlanTestService(t, "prompt_application")
	planService.newPlanID = func() (string, error) { return "aeplan_aaaaaaaaaaaaaaaa", nil }
	plan := planService.Create(ctx, applicationEvaluationPromptPlanInput("Scheduled Prompt regression", WorkflowRunComparisonUnchanged))
	if plan.FailureCode != "" || plan.Plan == nil || plan.Version == nil {
		t.Fatalf("create Prompt plan: %+v", plan)
	}

	repository := newMemoryApplicationEvaluationScheduleRepository(10, 20)
	authorityReads, quotaReads := 0, 0
	service := newApplicationEvaluationScheduleService(
		repository,
		planService.repository,
		func(ApplicationEvaluationContext, ApplicationEvaluationPlanVersion) string {
			authorityReads++
			return ""
		},
		func(ApplicationEvaluationContext, string) string {
			quotaReads++
			return ""
		},
	)
	service.now = func() time.Time { return time.Date(2026, 8, 30, 8, 0, 0, 0, time.UTC) }
	service.newScheduleID = func() (string, error) { return "aesch_aaaaaaaaaaaaaaaa", nil }
	definition := ApplicationEvaluationScheduleDefinitionInput{
		PlanID: plan.Plan.PlanID, PlanVersion: plan.Version.PlanVersion, PlanDigest: plan.Version.PlanDigest,
		ExpectedPlanRecordVersion: plan.Plan.RecordVersion, QuotaAPIKeyID: "key_aaaaaaaaaaaaaaaa",
		Schedule:                       ApplicationEvaluationScheduleDailyUTC{Rule: applicationEvaluationScheduleRuleDailyUTC, Hour: 9, Minute: 30},
		AcknowledgeProviderConsumption: true,
	}
	created := service.Create(ctx, ApplicationEvaluationScheduleCreateInput{ApplicationEvaluationScheduleDefinitionInput: definition})
	if created.FailureCode != "" || created.Schedule == nil || created.Version == nil ||
		created.Schedule.LifecycleState != applicationEvaluationScheduleStateDraft || created.Schedule.NextDueAt != nil ||
		created.Version.ItemCount != 1 || created.Version.MaxProviderAttempts != 1 {
		t.Fatalf("create schedule: %+v", created)
	}
	if quotaReads != 1 || authorityReads != 0 {
		t.Fatalf("create must validate quota but defer authority activation check: quota=%d authority=%d", quotaReads, authorityReads)
	}

	active := service.Activate(ctx, created.Schedule.ScheduleID, ApplicationEvaluationScheduleLifecycleInput{
		ExpectedVersion: created.Schedule.RecordVersion, AcknowledgeProviderConsumption: true,
	})
	if active.FailureCode != "" || active.Schedule == nil || active.Schedule.RecordVersion != 2 ||
		active.Schedule.LifecycleState != applicationEvaluationScheduleStateActive || active.Schedule.NextDueAt == nil ||
		*active.Schedule.NextDueAt != "2026-08-30T09:30:00Z" {
		t.Fatalf("activate schedule: %+v", active)
	}
	if quotaReads != 2 || authorityReads != 1 {
		t.Fatalf("activation must reread quota and authority: quota=%d authority=%d", quotaReads, authorityReads)
	}

	stale := service.Pause(ctx, active.Schedule.ScheduleID, ApplicationEvaluationScheduleLifecycleInput{ExpectedVersion: 1})
	if stale.FailureCode != ApplicationEvaluationScheduleFailureVersionConflict || stale.CurrentRecordVersion != 2 ||
		stale.CurrentState != applicationEvaluationScheduleStateActive {
		t.Fatalf("stale lifecycle CAS was accepted: %+v", stale)
	}
	paused := service.Pause(ctx, active.Schedule.ScheduleID, ApplicationEvaluationScheduleLifecycleInput{ExpectedVersion: 2})
	if paused.FailureCode != "" || paused.Schedule == nil || paused.Schedule.RecordVersion != 3 ||
		paused.Schedule.LifecycleState != applicationEvaluationScheduleStatePaused || paused.Schedule.NextDueAt != nil {
		t.Fatalf("pause schedule: %+v", paused)
	}

	definition.Schedule.Minute = 45
	revised := service.Revise(ctx, paused.Schedule.ScheduleID, ApplicationEvaluationScheduleReviseInput{
		ExpectedVersion: 3, ApplicationEvaluationScheduleDefinitionInput: definition,
	})
	if revised.FailureCode != "" || revised.Schedule == nil || revised.Version == nil || revised.Schedule.RecordVersion != 4 ||
		revised.Version.ScheduleVersion != 2 || revised.Version.Schedule.Minute != 45 {
		t.Fatalf("revise paused schedule: %+v", revised)
	}
	oldVersion := service.ReadVersion(ctx, revised.Schedule.ScheduleID, 1)
	if oldVersion.FailureCode != "" || oldVersion.Version == nil || oldVersion.Version.Schedule.Minute != 30 {
		t.Fatalf("read exact immutable schedule version: %+v", oldVersion)
	}

	resumed := service.Resume(ctx, revised.Schedule.ScheduleID, ApplicationEvaluationScheduleLifecycleInput{
		ExpectedVersion: 4, AcknowledgeProviderConsumption: true,
	})
	if resumed.FailureCode != "" || resumed.Schedule == nil || resumed.Schedule.RecordVersion != 5 ||
		resumed.Schedule.NextDueAt == nil || *resumed.Schedule.NextDueAt != "2026-08-30T09:45:00Z" {
		t.Fatalf("resume schedule: %+v", resumed)
	}
	if quotaReads != 4 || authorityReads != 2 {
		t.Fatalf("resume must reread quota and authority after revision: quota=%d authority=%d", quotaReads, authorityReads)
	}

	listed := service.List(ctx, ApplicationEvaluationScheduleListInput{LifecycleState: applicationEvaluationScheduleStateActive, Limit: 1})
	if listed.FailureCode != "" || len(listed.Schedules) != 1 || listed.Schedules[0].ScheduleID != resumed.Schedule.ScheduleID || listed.HasMore {
		t.Fatalf("list active schedules: %+v", listed)
	}
	missingAcknowledgement := service.Archive(ctx, resumed.Schedule.ScheduleID, ApplicationEvaluationScheduleLifecycleInput{ExpectedVersion: 5})
	if missingAcknowledgement.FailureCode != ApplicationEvaluationFailurePayloadInvalid {
		t.Fatalf("archive without no-future-occurrences acknowledgement was accepted: %+v", missingAcknowledgement)
	}
	archived := service.Archive(ctx, resumed.Schedule.ScheduleID, ApplicationEvaluationScheduleLifecycleInput{
		ExpectedVersion: 5, AcknowledgeNoFutureOccurrences: true,
	})
	if archived.FailureCode != "" || archived.Schedule == nil || archived.Schedule.LifecycleState != applicationEvaluationScheduleStateArchived ||
		archived.Schedule.NextDueAt != nil {
		t.Fatalf("archive schedule: %+v", archived)
	}
}

func TestApplicationEvaluationScheduleServiceActivationRevalidationFailsClosed(t *testing.T) {
	planService, ctx := newApplicationEvaluationPlanTestService(t, "prompt_application")
	planService.newPlanID = func() (string, error) { return "aeplan_bbbbbbbbbbbbbbbb", nil }
	plan := planService.Create(ctx, applicationEvaluationPromptPlanInput("Revalidation", WorkflowRunComparisonUnchanged))
	if plan.FailureCode != "" || plan.Plan == nil || plan.Version == nil {
		t.Fatalf("create Prompt plan: %+v", plan)
	}
	quotaFailure, authorityFailure := "", ""
	service := newApplicationEvaluationScheduleService(
		newMemoryApplicationEvaluationScheduleRepository(10, 20), planService.repository,
		func(ApplicationEvaluationContext, ApplicationEvaluationPlanVersion) string { return authorityFailure },
		func(ApplicationEvaluationContext, string) string { return quotaFailure },
	)
	service.now = func() time.Time { return time.Date(2026, 8, 30, 8, 0, 0, 0, time.UTC) }
	service.newScheduleID = func() (string, error) { return "aesch_bbbbbbbbbbbbbbbb", nil }
	definition := ApplicationEvaluationScheduleDefinitionInput{
		PlanID: plan.Plan.PlanID, PlanVersion: plan.Version.PlanVersion, PlanDigest: plan.Version.PlanDigest,
		ExpectedPlanRecordVersion: plan.Plan.RecordVersion, QuotaAPIKeyID: "key_aaaaaaaaaaaaaaaa",
		Schedule:                       ApplicationEvaluationScheduleDailyUTC{Rule: applicationEvaluationScheduleRuleDailyUTC, Hour: 10, Minute: 0},
		AcknowledgeProviderConsumption: true,
	}
	created := service.Create(ctx, ApplicationEvaluationScheduleCreateInput{ApplicationEvaluationScheduleDefinitionInput: definition})
	if created.FailureCode != "" || created.Schedule == nil {
		t.Fatalf("create schedule: %+v", created)
	}

	quotaFailure = ApplicationEvaluationScheduleFailureQuotaConsumerInvalid
	deniedQuota := service.Activate(ctx, created.Schedule.ScheduleID, ApplicationEvaluationScheduleLifecycleInput{
		ExpectedVersion: 1, AcknowledgeProviderConsumption: true,
	})
	if deniedQuota.FailureCode != ApplicationEvaluationScheduleFailureQuotaConsumerInvalid {
		t.Fatalf("revoked quota consumer did not fail closed: %+v", deniedQuota)
	}
	quotaFailure = ""
	authorityFailure = ApplicationEvaluationScheduleFailureAuthorityChanged
	deniedAuthority := service.Activate(ctx, created.Schedule.ScheduleID, ApplicationEvaluationScheduleLifecycleInput{
		ExpectedVersion: 1, AcknowledgeProviderConsumption: true,
	})
	if deniedAuthority.FailureCode != ApplicationEvaluationScheduleFailureAuthorityChanged {
		t.Fatalf("authority drift did not fail closed: %+v", deniedAuthority)
	}

	authorityFailure = ""
	archivedPlan := planService.Archive(ctx, plan.Plan.PlanID, ApplicationEvaluationPlanArchiveInput{
		ExpectedVersion: plan.Plan.RecordVersion, AcknowledgeNoNewCampaigns: true,
	})
	if archivedPlan.FailureCode != "" {
		t.Fatalf("archive bound plan: %+v", archivedPlan)
	}
	changedPlan := service.Activate(ctx, created.Schedule.ScheduleID, ApplicationEvaluationScheduleLifecycleInput{
		ExpectedVersion: 1, AcknowledgeProviderConsumption: true,
	})
	if changedPlan.FailureCode != ApplicationEvaluationScheduleFailurePlanChanged {
		t.Fatalf("inactive exact plan did not fail closed: %+v", changedPlan)
	}
	read := service.Read(ctx, created.Schedule.ScheduleID)
	if read.FailureCode != "" || read.Schedule == nil || read.Schedule.RecordVersion != 1 || read.Schedule.LifecycleState != applicationEvaluationScheduleStateDraft {
		t.Fatalf("failed revalidation mutated schedule: %+v", read)
	}
}
