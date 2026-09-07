package httpapi

import "context"

func validApplicationEvaluationScheduleCreate(
	ctx ApplicationEvaluationContext,
	schedule ApplicationEvaluationSchedule,
	version ApplicationEvaluationScheduleVersion,
) bool {
	return ctx.WriteEnabled && validateApplicationEvaluationSchedule(ctx, schedule) == nil &&
		validateApplicationEvaluationScheduleVersion(ctx, version) == nil && applicationEvaluationScheduleMatchesVersion(schedule, version) &&
		schedule.RecordVersion == 1 && version.ScheduleVersion == 1 && schedule.LifecycleState == applicationEvaluationScheduleStateDraft &&
		schedule.NextDueAt == nil && ctx.ActorRef == schedule.DelegatedByUserRef && schedule.CreatedByActorRef == ctx.ActorRef &&
		schedule.UpdatedByActorRef == ctx.ActorRef && version.CreatedByActorRef == ctx.ActorRef && schedule.CreatedAt == version.CreatedAt &&
		schedule.UpdatedAt == version.CreatedAt && schedule.RequestID == ctx.RequestID && schedule.AuditRef == ctx.AuditRef &&
		version.RequestID == ctx.RequestID && version.AuditRef == ctx.AuditRef
}

func validApplicationEvaluationScheduleRevision(
	ctx ApplicationEvaluationContext,
	expected int,
	schedule ApplicationEvaluationSchedule,
	version ApplicationEvaluationScheduleVersion,
) bool {
	return ctx.WriteEnabled && expected >= 1 && validateApplicationEvaluationSchedule(ctx, schedule) == nil &&
		validateApplicationEvaluationScheduleVersion(ctx, version) == nil && applicationEvaluationScheduleMatchesVersion(schedule, version) &&
		schedule.RecordVersion == expected+1 && schedule.LatestScheduleVersion == version.ScheduleVersion &&
		version.CreatedByActorRef == ctx.ActorRef && ctx.ActorRef == schedule.DelegatedByUserRef && schedule.UpdatedByActorRef == ctx.ActorRef &&
		schedule.UpdatedAt == version.CreatedAt && schedule.RequestID == ctx.RequestID && schedule.AuditRef == ctx.AuditRef &&
		version.RequestID == ctx.RequestID && version.AuditRef == ctx.AuditRef
}

func validApplicationEvaluationScheduleRevisionAgainstCurrent(
	current ApplicationEvaluationSchedule,
	currentVersion ApplicationEvaluationScheduleVersion,
	next ApplicationEvaluationSchedule,
	nextVersion ApplicationEvaluationScheduleVersion,
) bool {
	return validateApplicationEvaluationScheduleVersion(applicationEvaluationContextForSchedule(current), currentVersion) == nil &&
		applicationEvaluationScheduleMatchesVersion(current, currentVersion) &&
		(current.LifecycleState == applicationEvaluationScheduleStateDraft || current.LifecycleState == applicationEvaluationScheduleStatePaused) &&
		next.LifecycleState == current.LifecycleState && next.NextDueAt == nil && next.LatestScheduleVersion == current.LatestScheduleVersion+1 &&
		nextVersion.PreviousScheduleVersion == current.LatestScheduleVersion && sameApplicationEvaluationScheduleIdentity(current, next) &&
		next.CreatedAt == current.CreatedAt && next.CreatedByActorRef == current.CreatedByActorRef &&
		applicationEvaluationScheduleTimestampAfter(next.UpdatedAt, current.UpdatedAt)
}

func validApplicationEvaluationScheduleUpdateInput(
	ctx ApplicationEvaluationContext,
	expected int,
	schedule ApplicationEvaluationSchedule,
) bool {
	return ctx.WriteEnabled && expected >= 1 && validateApplicationEvaluationSchedule(ctx, schedule) == nil &&
		schedule.RecordVersion == expected+1 && schedule.UpdatedByActorRef == ctx.ActorRef && schedule.RequestID == ctx.RequestID &&
		schedule.AuditRef == ctx.AuditRef
}

func validApplicationEvaluationScheduleUpdateAgainstCurrent(
	ctx ApplicationEvaluationContext,
	current ApplicationEvaluationSchedule,
	version ApplicationEvaluationScheduleVersion,
	next ApplicationEvaluationSchedule,
) bool {
	if validateApplicationEvaluationSchedule(ctx, current) != nil || validateApplicationEvaluationScheduleVersion(ctx, version) != nil ||
		!applicationEvaluationScheduleMatchesVersion(current, version) || !sameApplicationEvaluationScheduleDefinition(current, next) ||
		next.CreatedAt != current.CreatedAt || next.CreatedByActorRef != current.CreatedByActorRef ||
		!applicationEvaluationScheduleTimestampAfter(next.UpdatedAt, current.UpdatedAt) ||
		!validApplicationEvaluationScheduleLifecycleTransition(current.LifecycleState, next.LifecycleState) {
		return false
	}
	if current.LifecycleState == applicationEvaluationScheduleStateActive && next.LifecycleState == applicationEvaluationScheduleStateActive {
		return ctx.ActorRef == current.SystemActorRef && current.NextDueAt != nil && next.NextDueAt != nil &&
			applicationEvaluationScheduleTimestampAfter(*next.NextDueAt, *current.NextDueAt) &&
			applicationEvaluationScheduleProjectionIsExact(version.Schedule, current.NextDueAt, next.UpdatedAt, next.NextDueAt)
	}
	return ctx.ActorRef == current.DelegatedByUserRef &&
		(next.LifecycleState != applicationEvaluationScheduleStateActive || applicationEvaluationScheduleActivationDueIsExact(version.Schedule, next.UpdatedAt, next.NextDueAt))
}

func validApplicationEvaluationScheduleClaim(
	ctx ApplicationEvaluationContext,
	due ApplicationEvaluationScheduleOccurrence,
	claimed ApplicationEvaluationScheduleOccurrence,
) bool {
	return ctx.WriteEnabled && validateApplicationEvaluationScheduleOccurrence(ctx, due) == nil &&
		validateApplicationEvaluationScheduleOccurrence(ctx, claimed) == nil && validApplicationEvaluationScheduleOccurrenceUpdate(due, claimed) &&
		due.State == applicationEvaluationScheduleOccurrenceStateDue && claimed.State == applicationEvaluationScheduleOccurrenceStateClaimed &&
		ctx.ActorRef == due.SystemActorRef && due.RequestID == ctx.RequestID && due.AuditRef == ctx.AuditRef &&
		claimed.RequestID == ctx.RequestID && claimed.AuditRef == ctx.AuditRef
}

func validApplicationEvaluationScheduleClaimBinding(
	schedule ApplicationEvaluationSchedule,
	version ApplicationEvaluationScheduleVersion,
	due ApplicationEvaluationScheduleOccurrence,
) bool {
	ctx := applicationEvaluationContextForSchedule(schedule)
	return validateApplicationEvaluationSchedule(ctx, schedule) == nil && validateApplicationEvaluationScheduleVersion(ctx, version) == nil &&
		applicationEvaluationScheduleMatchesVersion(schedule, version) && schedule.LifecycleState == applicationEvaluationScheduleStateActive &&
		schedule.LatestScheduleVersion == due.ScheduleVersion && schedule.LatestScheduleDigest == due.ScheduleDigest &&
		schedule.NextDueAt != nil && *schedule.NextDueAt == due.ScheduledForUTC && version.ScheduleDigest == due.ScheduleDigest &&
		version.Authorization.SystemActorRef == due.SystemActorRef && version.Authorization.DelegatedByUserRef == due.DelegatedByUserRef
}

func validApplicationEvaluationScheduleOccurrenceUpdateInput(
	ctx ApplicationEvaluationContext,
	expected int,
	occurrence ApplicationEvaluationScheduleOccurrence,
) bool {
	return ctx.WriteEnabled && expected >= 1 && validateApplicationEvaluationScheduleOccurrence(ctx, occurrence) == nil &&
		occurrence.RecordVersion == expected+1 && ctx.ActorRef == occurrence.SystemActorRef && occurrence.RequestID == ctx.RequestID &&
		occurrence.AuditRef == ctx.AuditRef
}

func validApplicationEvaluationScheduleOccurrenceAgainstVersion(
	current ApplicationEvaluationScheduleOccurrence,
	next ApplicationEvaluationScheduleOccurrence,
	version ApplicationEvaluationScheduleVersion,
) bool {
	return validateApplicationEvaluationScheduleVersion(applicationEvaluationContextForOccurrence(next), version) == nil &&
		version.ScheduleDigest == next.ScheduleDigest && version.Authorization.SystemActorRef == next.SystemActorRef &&
		version.Authorization.DelegatedByUserRef == next.DelegatedByUserRef && validApplicationEvaluationScheduleOccurrenceUpdate(current, next)
}

func applicationEvaluationContextForSchedule(schedule ApplicationEvaluationSchedule) ApplicationEvaluationContext {
	return ApplicationEvaluationContext{
		RequestContext: context.Background(), TenantRef: schedule.TenantRef, WorkspaceID: schedule.WorkspaceID, Environment: schedule.Environment,
		ApplicationID: schedule.ApplicationID, ActorRef: schedule.UpdatedByActorRef, RequestID: schedule.RequestID,
		AuditRef: schedule.AuditRef, WriteEnabled: true,
	}
}

func applicationEvaluationContextForOccurrence(occurrence ApplicationEvaluationScheduleOccurrence) ApplicationEvaluationContext {
	return ApplicationEvaluationContext{
		RequestContext: context.Background(), TenantRef: occurrence.TenantRef, WorkspaceID: occurrence.WorkspaceID, Environment: occurrence.Environment,
		ApplicationID: occurrence.ApplicationID, ActorRef: occurrence.SystemActorRef, RequestID: occurrence.RequestID,
		AuditRef: occurrence.AuditRef, WriteEnabled: true,
	}
}

func firstApplicationEvaluationScheduleStoreError(err error) error {
	if err != nil {
		return err
	}
	return errApplicationEvaluationScheduleStoreContract
}
