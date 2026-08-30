package httpapi

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestApplicationEvaluationScheduleContractAndUTCProjection(t *testing.T) {
	ctx, schedule, version := applicationEvaluationScheduleTestRecords(t)
	if err := validateApplicationEvaluationSchedule(ctx, schedule); err != nil {
		t.Fatalf("validate schedule: %v", err)
	}
	if err := validateApplicationEvaluationScheduleVersion(ctx, version); err != nil {
		t.Fatalf("validate version: %v", err)
	}
	next, err := applicationEvaluationScheduleNextDue(time.Date(2026, 8, 30, 9, 30, 0, 0, time.UTC), version.Schedule)
	if err != nil || next.Format(time.RFC3339Nano) != "2026-08-31T09:30:00Z" {
		t.Fatalf("equal instant must roll to next UTC day: next=%s err=%v", next, err)
	}
	next, err = applicationEvaluationScheduleNextDue(time.Date(2026, 8, 30, 8, 0, 0, 0, time.FixedZone("CST", 8*60*60)), version.Schedule)
	if err != nil || next.Format(time.RFC3339Nano) != "2026-08-30T09:30:00Z" {
		t.Fatalf("schedule must be calculated in UTC: next=%s err=%v", next, err)
	}

	invalid := version
	invalid.Authorization.RequiredPermissions = []string{
		applicationEvaluationSchedulePermissionWorkflowRunExecute,
		applicationEvaluationSchedulePermissionEvaluationExecute,
	}
	invalid.ScheduleDigest, _ = applicationEvaluationScheduleDigest(invalid)
	if err := validateApplicationEvaluationScheduleVersion(ctx, invalid); !errors.Is(err, errApplicationEvaluationScheduleStoreContract) {
		t.Fatalf("permission order drift must fail closed: %v", err)
	}
	invalid = version
	invalid.Authorization.Model = "bearer delegated-token"
	invalid.ScheduleDigest, _ = applicationEvaluationScheduleDigest(invalid)
	if err := validateApplicationEvaluationScheduleVersion(ctx, invalid); !errors.Is(err, errApplicationEvaluationScheduleStoreContract) {
		t.Fatalf("bearer delegation must fail closed: %v", err)
	}
	invalid = version
	invalid.Authorization.SystemActorRef = "system:token=delegated-secret"
	invalid.ScheduleDigest, _ = applicationEvaluationScheduleDigest(invalid)
	if err := validateApplicationEvaluationScheduleVersion(ctx, invalid); !errors.Is(err, errApplicationEvaluationScheduleStoreContract) {
		t.Fatalf("secret-shaped actor reference must fail closed: %v", err)
	}
	invalid = version
	invalid.ExecutionProfile = applicationInteractionProfileAgentCopilot
	invalid.ScheduleDigest, _ = applicationEvaluationScheduleDigest(invalid)
	if err := validateApplicationEvaluationScheduleVersion(ctx, invalid); !errors.Is(err, errApplicationEvaluationScheduleStoreContract) {
		t.Fatalf("non-prompt profile must fail closed: %v", err)
	}
	invalid = version
	invalid.MaxProviderAttempts++
	invalid.ScheduleDigest, _ = applicationEvaluationScheduleDigest(invalid)
	if err := validateApplicationEvaluationScheduleVersion(ctx, invalid); !errors.Is(err, errApplicationEvaluationScheduleStoreContract) {
		t.Fatalf("inexact provider budget must fail closed: %v", err)
	}
}

func TestMemoryApplicationEvaluationScheduleVersionsCASAndLifecycle(t *testing.T) {
	ctx, schedule, version := applicationEvaluationScheduleTestRecords(t)
	repository := newMemoryApplicationEvaluationScheduleRepository(10, 20)
	if err := repository.CreateSchedule(ctx, schedule, version); err != nil {
		t.Fatalf("create schedule: %v", err)
	}

	revisedVersion := version
	revisedVersion.ScheduleVersion = 2
	revisedVersion.PreviousScheduleVersion = 1
	revisedVersion.Schedule.Minute = 45
	revisedVersion.CreatedAt = "2026-08-30T08:01:00Z"
	revisedVersion.RequestID = "request-schedule-revise"
	revisedVersion.AuditRef = "audit:schedule-revise"
	revisedVersion.ScheduleDigest, _ = applicationEvaluationScheduleDigest(revisedVersion)
	revised := schedule
	revised.RecordVersion = 2
	revised.LatestScheduleVersion = 2
	revised.LatestScheduleDigest = revisedVersion.ScheduleDigest
	revised.UpdatedAt = revisedVersion.CreatedAt
	revised.RequestID = revisedVersion.RequestID
	revised.AuditRef = revisedVersion.AuditRef
	revisionCtx := ctx
	revisionCtx.RequestID = revised.RequestID
	revisionCtx.AuditRef = revised.AuditRef
	stored, updated, err := repository.ReviseSchedule(revisionCtx, 1, revised, revisedVersion)
	if err != nil || !updated || stored.LatestScheduleVersion != 2 {
		t.Fatalf("revise schedule: updated=%v value=%#v err=%v", updated, stored, err)
	}
	if _, updated, err = repository.ReviseSchedule(revisionCtx, 1, revised, revisedVersion); !errors.Is(err, errApplicationEvaluationScheduleVersionConflict) || updated {
		t.Fatalf("stale revision must conflict: updated=%v err=%v", updated, err)
	}
	readV1, found, err := repository.ReadScheduleVersion(ctx, schedule.ScheduleID, 1)
	if err != nil || !found || readV1.Schedule.Minute != 30 || readV1.ScheduleDigest != version.ScheduleDigest {
		t.Fatalf("immutable version changed: found=%v value=%#v err=%v", found, readV1, err)
	}

	active := revised
	active.RecordVersion = 3
	active.LifecycleState = applicationEvaluationScheduleStateActive
	active.UpdatedAt = "2026-08-30T08:02:00Z"
	active.RequestID = "request-schedule-activate"
	active.AuditRef = "audit:schedule-activate"
	nextDue := "2026-08-30T09:45:00Z"
	active.NextDueAt = &nextDue
	wrongDue := active
	wrongDueAt := "2026-08-30T10:45:00Z"
	wrongDue.NextDueAt = &wrongDueAt
	wrongActivationCtx := ctx
	wrongActivationCtx.RequestID = wrongDue.RequestID
	wrongActivationCtx.AuditRef = wrongDue.AuditRef
	if _, _, err = repository.UpdateSchedule(wrongActivationCtx, 2, wrongDue); !errors.Is(err, errApplicationEvaluationScheduleStoreContract) {
		t.Fatalf("activation must use exact daily UTC projection: %v", err)
	}
	activationCtx := ctx
	activationCtx.RequestID = active.RequestID
	activationCtx.AuditRef = active.AuditRef
	stored, updated, err = repository.UpdateSchedule(activationCtx, 2, active)
	if err != nil || !updated || stored.LifecycleState != applicationEvaluationScheduleStateActive {
		t.Fatalf("activate schedule: updated=%v value=%#v err=%v", updated, stored, err)
	}
	activeRevision := revisedVersion
	activeRevision.ScheduleVersion = 3
	activeRevision.PreviousScheduleVersion = 2
	activeRevision.CreatedAt = "2026-08-30T08:03:00Z"
	activeRevision.ScheduleDigest, _ = applicationEvaluationScheduleDigest(activeRevision)
	activeRecord := active
	activeRecord.RecordVersion = 4
	activeRecord.LatestScheduleVersion = 3
	activeRecord.LatestScheduleDigest = activeRevision.ScheduleDigest
	activeRecord.NextDueAt = nil
	activeRecord.UpdatedAt = activeRevision.CreatedAt
	if _, _, err = repository.ReviseSchedule(ctx, 3, activeRecord, activeRevision); !errors.Is(err, errApplicationEvaluationScheduleStoreContract) {
		t.Fatalf("active schedule revision must fail closed: %v", err)
	}

	paused := active
	paused.RecordVersion = 4
	paused.LifecycleState = applicationEvaluationScheduleStatePaused
	paused.NextDueAt = nil
	paused.UpdatedAt = "2026-08-30T08:04:00Z"
	paused.RequestID = "request-schedule-pause"
	paused.AuditRef = "audit:schedule-pause"
	pauseCtx := ctx
	pauseCtx.RequestID = paused.RequestID
	pauseCtx.AuditRef = paused.AuditRef
	stored, updated, err = repository.UpdateSchedule(pauseCtx, 3, paused)
	if err != nil || !updated || stored.LifecycleState != applicationEvaluationScheduleStatePaused {
		t.Fatalf("pause schedule: updated=%v value=%#v err=%v", updated, stored, err)
	}
}

func TestMemoryApplicationEvaluationScheduleRestrictedActors(t *testing.T) {
	userCtx, schedule, version := applicationEvaluationScheduleTestRecords(t)
	repository := newMemoryApplicationEvaluationScheduleRepository(10, 20)
	if err := repository.CreateSchedule(userCtx, schedule, version); err != nil {
		t.Fatalf("create schedule: %v", err)
	}
	active := applicationEvaluationActivateSchedule(t, repository, userCtx, schedule, "2026-08-30T09:30:00Z")

	systemCtx := userCtx
	systemCtx.ActorRef = version.Authorization.SystemActorRef
	systemCtx.RequestID = "request-system-projection"
	systemCtx.AuditRef = "audit:system-projection"
	driftedProjection := active
	driftedProjection.RecordVersion++
	driftedProjection.UpdatedByActorRef = systemCtx.ActorRef
	driftedProjection.UpdatedAt = "2026-08-30T09:31:00Z"
	driftedProjection.RequestID = systemCtx.RequestID
	driftedProjection.AuditRef = systemCtx.AuditRef
	driftedDue := "2026-09-01T09:30:00Z"
	driftedProjection.NextDueAt = &driftedDue
	if _, _, err := repository.UpdateSchedule(systemCtx, active.RecordVersion, driftedProjection); !errors.Is(err, errApplicationEvaluationScheduleStoreContract) {
		t.Fatalf("system projection must not skip a UTC day: %v", err)
	}
	advanced := active
	advanced.RecordVersion++
	advanced.UpdatedByActorRef = systemCtx.ActorRef
	advanced.UpdatedAt = "2026-08-30T09:31:00Z"
	advanced.RequestID = systemCtx.RequestID
	advanced.AuditRef = systemCtx.AuditRef
	nextDue := "2026-08-31T09:30:00Z"
	advanced.NextDueAt = &nextDue
	stored, updated, err := repository.UpdateSchedule(systemCtx, active.RecordVersion, advanced)
	if err != nil || !updated || stored.UpdatedByActorRef != version.Authorization.SystemActorRef {
		t.Fatalf("system projection advance: updated=%v value=%#v err=%v", updated, stored, err)
	}

	paused := advanced
	paused.RecordVersion++
	paused.LifecycleState = applicationEvaluationScheduleStatePaused
	paused.NextDueAt = nil
	paused.UpdatedAt = "2026-08-30T09:32:00Z"
	if _, _, err = repository.UpdateSchedule(systemCtx, advanced.RecordVersion, paused); !errors.Is(err, errApplicationEvaluationScheduleStoreContract) {
		t.Fatalf("system actor must not control user lifecycle: %v", err)
	}
	userAdvance := advanced
	userAdvance.RecordVersion++
	userAdvance.UpdatedByActorRef = userCtx.ActorRef
	userAdvance.UpdatedAt = "2026-08-30T09:33:00Z"
	userNextDue := "2026-09-01T09:30:00Z"
	userAdvance.NextDueAt = &userNextDue
	if _, _, err = repository.UpdateSchedule(userCtx, advanced.RecordVersion, userAdvance); !errors.Is(err, errApplicationEvaluationScheduleStoreContract) {
		t.Fatalf("delegated user must not impersonate system projection advance: %v", err)
	}
}

func TestMemoryApplicationEvaluationScheduleOccurrenceSingleWinnerAndTerminalState(t *testing.T) {
	userCtx, schedule, version := applicationEvaluationScheduleTestRecords(t)
	repository := newMemoryApplicationEvaluationScheduleRepository(10, 20)
	if err := repository.CreateSchedule(userCtx, schedule, version); err != nil {
		t.Fatalf("create schedule: %v", err)
	}
	applicationEvaluationActivateSchedule(t, repository, userCtx, schedule, "2026-08-30T09:30:00Z")
	systemCtx, due, claimed := applicationEvaluationScheduleOccurrenceTestRecords(userCtx, version, "2026-08-30T09:30:00Z")

	var winners atomic.Int32
	var unexpected atomic.Int32
	var wait sync.WaitGroup
	for range 12 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, won, err := repository.ClaimOccurrence(systemCtx, due, claimed)
			if err == nil && won {
				winners.Add(1)
				return
			}
			if !errors.Is(err, errApplicationEvaluationScheduleClaimConflict) || won {
				unexpected.Add(1)
			}
		}()
	}
	wait.Wait()
	if winners.Load() != 1 || unexpected.Load() != 0 {
		t.Fatalf("claim must have one winner: winners=%d unexpected=%d", winners.Load(), unexpected.Load())
	}

	userClaim := due
	userClaim.RequestID = "request-user-claim"
	if _, _, err := repository.ClaimOccurrence(userCtx, userClaim, claimed); !errors.Is(err, errApplicationEvaluationScheduleStoreContract) {
		t.Fatalf("delegated user must not claim as system actor: %v", err)
	}

	campaignID := "aecamp_bbbbbbbbbbbbbbbb"
	campaignCreated := claimed
	campaignCreated.RecordVersion = 3
	campaignCreated.State = applicationEvaluationScheduleOccurrenceStateCampaignCreated
	campaignCreated.CampaignID = &campaignID
	campaignCreated.UpdatedAt = "2026-08-30T09:30:02Z"
	campaignCreated.RequestID = "request-campaign-created"
	campaignCreated.AuditRef = "audit:campaign-created"
	campaignContext := systemCtx
	campaignContext.RequestID = campaignCreated.RequestID
	campaignContext.AuditRef = campaignCreated.AuditRef
	stored, updated, err := repository.UpdateOccurrence(campaignContext, 2, campaignCreated)
	if err != nil || !updated || stored.State != applicationEvaluationScheduleOccurrenceStateCampaignCreated {
		t.Fatalf("record campaign: updated=%v value=%#v err=%v", updated, stored, err)
	}
	observing := campaignCreated
	observing.RecordVersion = 4
	observing.State = applicationEvaluationScheduleOccurrenceStateObserving
	observing.UpdatedAt = "2026-08-30T09:30:03Z"
	observing.RequestID = "request-observing"
	observing.AuditRef = "audit:observing"
	observingContext := systemCtx
	observingContext.RequestID = observing.RequestID
	observingContext.AuditRef = observing.AuditRef
	if _, updated, err = repository.UpdateOccurrence(observingContext, 3, observing); err != nil || !updated {
		t.Fatalf("observe campaign: updated=%v err=%v", updated, err)
	}
	completedAt := "2026-08-30T09:30:04Z"
	succeeded := observing
	succeeded.RecordVersion = 5
	succeeded.State = applicationEvaluationScheduleOccurrenceStateSucceeded
	succeeded.UpdatedAt = completedAt
	succeeded.CompletedAt = &completedAt
	succeeded.RequestID = "request-succeeded"
	succeeded.AuditRef = "audit:succeeded"
	succeededContext := systemCtx
	succeededContext.RequestID = succeeded.RequestID
	succeededContext.AuditRef = succeeded.AuditRef
	stored, updated, err = repository.UpdateOccurrence(succeededContext, 4, succeeded)
	if err != nil || !updated || stored.State != applicationEvaluationScheduleOccurrenceStateSucceeded {
		t.Fatalf("complete occurrence: updated=%v value=%#v err=%v", updated, stored, err)
	}
	retry := observing
	retry.RecordVersion = 6
	retry.UpdatedAt = "2026-08-30T09:30:05Z"
	retryContext := systemCtx
	retryContext.RequestID = retry.RequestID
	retryContext.AuditRef = retry.AuditRef
	if _, _, err = repository.UpdateOccurrence(retryContext, 5, retry); !errors.Is(err, errApplicationEvaluationScheduleStoreContract) {
		t.Fatalf("terminal occurrence must not replay: %v", err)
	}
}

func TestMemoryApplicationEvaluationScheduleOccurrenceSkipAndCorruptionClose(t *testing.T) {
	userCtx, schedule, version := applicationEvaluationScheduleTestRecords(t)
	repository := newMemoryApplicationEvaluationScheduleRepository(10, 20)
	if err := repository.CreateSchedule(userCtx, schedule, version); err != nil {
		t.Fatalf("create schedule: %v", err)
	}
	applicationEvaluationActivateSchedule(t, repository, userCtx, schedule, "2026-08-30T09:30:00Z")
	systemCtx, due, claimed := applicationEvaluationScheduleOccurrenceTestRecords(userCtx, version, "2026-08-30T09:30:00Z")
	if _, won, err := repository.ClaimOccurrence(systemCtx, due, claimed); err != nil || !won {
		t.Fatalf("claim occurrence: won=%v err=%v", won, err)
	}
	completedAt := "2026-08-30T09:30:02Z"
	failure := ApplicationEvaluationScheduleFailureMissedWindow
	skipped := claimed
	skipped.RecordVersion = 3
	skipped.State = applicationEvaluationScheduleOccurrenceStateSkipped
	skipped.UpdatedAt = completedAt
	skipped.CompletedAt = &completedAt
	skipped.FailureCode = &failure
	skipped.RequestID = "request-missed-window"
	skipped.AuditRef = "audit:missed-window"
	skippedContext := systemCtx
	skippedContext.RequestID = skipped.RequestID
	skippedContext.AuditRef = skipped.AuditRef
	if _, updated, err := repository.UpdateOccurrence(skippedContext, 2, skipped); err != nil || !updated {
		t.Fatalf("skip missed occurrence: updated=%v err=%v", updated, err)
	}
	badFailure := ApplicationEvaluationScheduleFailureQuotaDenied
	invalidSkip := claimed
	invalidSkip.RecordVersion = 3
	invalidSkip.State = applicationEvaluationScheduleOccurrenceStateSkipped
	invalidSkip.UpdatedAt = completedAt
	invalidSkip.CompletedAt = &completedAt
	invalidSkip.FailureCode = &badFailure
	if err := validateApplicationEvaluationScheduleOccurrence(systemCtx, invalidSkip); !errors.Is(err, errApplicationEvaluationScheduleStoreContract) {
		t.Fatalf("quota denial cannot masquerade as skipped: %v", err)
	}

	repository.mu.Lock()
	key := applicationEvaluationScheduleOccurrenceKey(systemCtx, due.ScheduleID, due.ScheduleVersion, due.ScheduledForUTC)
	corrupted := repository.occurrences[key]
	corrupted.ClientCampaignKey = "scheduled_campaign_corrupted"
	repository.occurrences[key] = corrupted
	repository.mu.Unlock()
	if _, _, err := repository.ReadOccurrence(systemCtx, due.ScheduleID, due.ScheduleVersion, due.ScheduledForUTC); !errors.Is(err, errApplicationEvaluationScheduleStoreContract) {
		t.Fatalf("corrupted occurrence must fail closed: %v", err)
	}

	repository.unavailable = true
	if _, _, err := repository.ReadSchedule(userCtx, schedule.ScheduleID); !errors.Is(err, errApplicationEvaluationScheduleStoreUnavailable) {
		t.Fatalf("store outage must not fall back: %v", err)
	}
}

func TestMemoryApplicationEvaluationScheduleWriteGateAndCapacity(t *testing.T) {
	ctx, schedule, version := applicationEvaluationScheduleTestRecords(t)
	readOnlyContext := ctx
	readOnlyContext.WriteEnabled = false
	if err := newMemoryApplicationEvaluationScheduleRepository(1, 1).CreateSchedule(readOnlyContext, schedule, version); !errors.Is(err, errApplicationEvaluationScheduleStoreContract) {
		t.Fatalf("read-only context must not mutate schedule owner: %v", err)
	}

	repository := newMemoryApplicationEvaluationScheduleRepository(1, 1)
	if err := repository.CreateSchedule(ctx, schedule, version); err != nil {
		t.Fatalf("fill schedule capacity: %v", err)
	}
	secondVersion := version
	secondVersion.ScheduleID = "aesch_bbbbbbbbbbbbbbbb"
	secondVersion.ScheduleDigest, _ = applicationEvaluationScheduleDigest(secondVersion)
	secondSchedule := schedule
	secondSchedule.ScheduleID = secondVersion.ScheduleID
	secondSchedule.LatestScheduleDigest = secondVersion.ScheduleDigest
	if err := repository.CreateSchedule(ctx, secondSchedule, secondVersion); !errors.Is(err, errApplicationEvaluationScheduleStoreUnavailable) {
		t.Fatalf("memory owner must fail closed at capacity: %v", err)
	}
}

func applicationEvaluationScheduleTestRecords(t *testing.T) (ApplicationEvaluationContext, ApplicationEvaluationSchedule, ApplicationEvaluationScheduleVersion) {
	t.Helper()
	ctx := ApplicationEvaluationContext{
		RequestContext: context.Background(), RequestID: "request-schedule-create", TenantRef: "tenant-one",
		WorkspaceID: "workspace-one", Environment: applicationEvaluationEnvironmentTest,
		ApplicationID: "app_aaaaaaaaaaaaaaaa", ActorRef: "subject-owner", AuditRef: "audit:schedule-create", WriteEnabled: true,
	}
	version := ApplicationEvaluationScheduleVersion{
		SchemaVersion: applicationEvaluationScheduleVersionSchemaVersion, ScheduleID: "aesch_aaaaaaaaaaaaaaaa",
		ScheduleVersion: 1, PreviousScheduleVersion: 0, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
		Environment: ctx.Environment, ApplicationID: ctx.ApplicationID, PlanID: "aeplan_aaaaaaaaaaaaaaaa", PlanVersion: 1,
		PlanDigest: "sha256:" + strings.Repeat("a", 64), ExecutionProfile: applicationInteractionProfilePrompt,
		QuotaAPIKeyID: "key_aaaaaaaaaaaaaaaa", Schedule: ApplicationEvaluationScheduleDailyUTC{Rule: applicationEvaluationScheduleRuleDailyUTC, Hour: 9, Minute: 30},
		ItemCount: 2, MaxProviderAttempts: 2, MissedWindowPolicy: applicationEvaluationScheduleMissedWindowPolicy,
		OverlapPolicy: applicationEvaluationScheduleOverlapPolicy,
		Authorization: ApplicationEvaluationScheduleAuthorization{
			Model: applicationEvaluationScheduleAuthorizationModel, SystemActorRef: "system:application-evaluation-scheduler",
			DelegatedByUserRef: ctx.ActorRef, RequiredPermissions: append([]string(nil), applicationEvaluationScheduleRequiredPermissions...),
			RevalidationPolicy: applicationEvaluationScheduleRevalidationPolicy, APIKeyOwnershipPolicy: applicationEvaluationScheduleAPIKeyOwnershipPolicy,
			RevocationPolicy: applicationEvaluationScheduleRevocationPolicy,
		},
		CreatedAt: "2026-08-30T08:00:00Z", CreatedByActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	var err error
	version.ScheduleDigest, err = applicationEvaluationScheduleDigest(version)
	if err != nil {
		t.Fatalf("schedule digest: %v", err)
	}
	schedule := ApplicationEvaluationSchedule{
		SchemaVersion: applicationEvaluationScheduleSchemaVersion, ScheduleID: version.ScheduleID, RecordVersion: 1,
		LatestScheduleVersion: version.ScheduleVersion, LatestScheduleDigest: version.ScheduleDigest,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Environment: ctx.Environment, ApplicationID: ctx.ApplicationID,
		PlanID: version.PlanID, PlanVersion: version.PlanVersion, PlanDigest: version.PlanDigest, ExecutionProfile: version.ExecutionProfile,
		QuotaAPIKeyID: version.QuotaAPIKeyID, AuthorizationModel: version.Authorization.Model,
		SystemActorRef: version.Authorization.SystemActorRef, DelegatedByUserRef: version.Authorization.DelegatedByUserRef,
		LifecycleState: applicationEvaluationScheduleStateDraft, CreatedAt: version.CreatedAt, UpdatedAt: version.CreatedAt,
		CreatedByActorRef: ctx.ActorRef, UpdatedByActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	return ctx, schedule, version
}

func applicationEvaluationActivateSchedule(
	t *testing.T,
	repository *memoryApplicationEvaluationScheduleRepository,
	ctx ApplicationEvaluationContext,
	schedule ApplicationEvaluationSchedule,
	nextDueAt string,
) ApplicationEvaluationSchedule {
	t.Helper()
	active := schedule
	active.RecordVersion++
	active.LifecycleState = applicationEvaluationScheduleStateActive
	active.NextDueAt = &nextDueAt
	active.UpdatedAt = "2026-08-30T08:01:00Z"
	active.RequestID = "request-schedule-activate"
	active.AuditRef = "audit:schedule-activate"
	activationContext := ctx
	activationContext.RequestID = active.RequestID
	activationContext.AuditRef = active.AuditRef
	stored, updated, err := repository.UpdateSchedule(activationContext, schedule.RecordVersion, active)
	if err != nil || !updated {
		t.Fatalf("activate schedule: updated=%v value=%#v err=%v", updated, stored, err)
	}
	return stored
}

func applicationEvaluationScheduleOccurrenceTestRecords(
	userCtx ApplicationEvaluationContext,
	version ApplicationEvaluationScheduleVersion,
	scheduledForUTC string,
) (ApplicationEvaluationContext, ApplicationEvaluationScheduleOccurrence, ApplicationEvaluationScheduleOccurrence) {
	systemCtx := userCtx
	systemCtx.ActorRef = version.Authorization.SystemActorRef
	systemCtx.RequestID = "request-occurrence-claim"
	systemCtx.AuditRef = "audit:occurrence-claim"
	due := ApplicationEvaluationScheduleOccurrence{
		SchemaVersion: applicationEvaluationScheduleOccurrenceSchemaVersion, RecordVersion: 1,
		TenantRef: userCtx.TenantRef, WorkspaceID: userCtx.WorkspaceID, Environment: userCtx.Environment, ApplicationID: userCtx.ApplicationID,
		ScheduleID: version.ScheduleID, ScheduleVersion: version.ScheduleVersion, ScheduleDigest: version.ScheduleDigest,
		ScheduledForUTC: scheduledForUTC, State: applicationEvaluationScheduleOccurrenceStateDue,
		ClientCampaignKey: applicationEvaluationScheduleClientCampaignKey(version.ScheduleID, version.ScheduleVersion, scheduledForUTC),
		SystemActorRef:    version.Authorization.SystemActorRef, DelegatedByUserRef: version.Authorization.DelegatedByUserRef,
		CreatedAt: scheduledForUTC, UpdatedAt: scheduledForUTC, RequestID: systemCtx.RequestID, AuditRef: systemCtx.AuditRef,
	}
	claimedAt := "2026-08-30T09:30:01Z"
	claimed := due
	claimed.RecordVersion = 2
	claimed.State = applicationEvaluationScheduleOccurrenceStateClaimed
	claimed.ClaimedAt = &claimedAt
	claimed.UpdatedAt = claimedAt
	return systemCtx, due, claimed
}

func TestApplicationEvaluationScheduleCloneIsolation(t *testing.T) {
	_, schedule, version := applicationEvaluationScheduleTestRecords(t)
	nextDue := "2026-08-30T09:30:00Z"
	schedule.NextDueAt = &nextDue
	clonedSchedule := cloneApplicationEvaluationSchedule(schedule)
	*clonedSchedule.NextDueAt = "2026-08-31T09:30:00Z"
	if *schedule.NextDueAt == *clonedSchedule.NextDueAt {
		t.Fatal("schedule pointer was not cloned")
	}
	clonedVersion := cloneApplicationEvaluationScheduleVersion(version)
	clonedVersion.Authorization.RequiredPermissions[0] = "changed"
	if reflect.DeepEqual(version.Authorization.RequiredPermissions, clonedVersion.Authorization.RequiredPermissions) {
		t.Fatal("authorization permissions were not cloned")
	}
}
