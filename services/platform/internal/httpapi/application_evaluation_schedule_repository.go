package httpapi

import (
	"reflect"
	"strconv"
	"sync"
	"time"
)

type applicationEvaluationScheduleRepository interface {
	CreateSchedule(ApplicationEvaluationContext, ApplicationEvaluationSchedule, ApplicationEvaluationScheduleVersion) error
	ReviseSchedule(ApplicationEvaluationContext, int, ApplicationEvaluationSchedule, ApplicationEvaluationScheduleVersion) (ApplicationEvaluationSchedule, bool, error)
	UpdateSchedule(ApplicationEvaluationContext, int, ApplicationEvaluationSchedule) (ApplicationEvaluationSchedule, bool, error)
	ReadSchedule(ApplicationEvaluationContext, string) (ApplicationEvaluationSchedule, bool, error)
	ReadScheduleVersion(ApplicationEvaluationContext, string, int) (ApplicationEvaluationScheduleVersion, bool, error)
	ClaimOccurrence(ApplicationEvaluationContext, ApplicationEvaluationScheduleOccurrence, ApplicationEvaluationScheduleOccurrence) (ApplicationEvaluationScheduleOccurrence, bool, error)
	UpdateOccurrence(ApplicationEvaluationContext, int, ApplicationEvaluationScheduleOccurrence) (ApplicationEvaluationScheduleOccurrence, bool, error)
	ReadOccurrence(ApplicationEvaluationContext, string, int, string) (ApplicationEvaluationScheduleOccurrence, bool, error)
}

type memoryApplicationEvaluationScheduleRepository struct {
	mu                 sync.RWMutex
	scheduleCapacity   int
	occurrenceCapacity int
	schedules          map[string]ApplicationEvaluationSchedule
	versions           map[string]map[int]ApplicationEvaluationScheduleVersion
	occurrences        map[string]ApplicationEvaluationScheduleOccurrence
	unavailable        bool
}

func newMemoryApplicationEvaluationScheduleRepository(scheduleCapacity, occurrenceCapacity int) *memoryApplicationEvaluationScheduleRepository {
	if scheduleCapacity <= 0 {
		scheduleCapacity = applicationEvaluationScheduleMemoryCapacity
	}
	if occurrenceCapacity <= 0 {
		occurrenceCapacity = applicationEvaluationOccurrenceMemoryCapacity
	}
	return &memoryApplicationEvaluationScheduleRepository{
		scheduleCapacity: scheduleCapacity, occurrenceCapacity: occurrenceCapacity,
		schedules:   make(map[string]ApplicationEvaluationSchedule, scheduleCapacity),
		versions:    make(map[string]map[int]ApplicationEvaluationScheduleVersion, scheduleCapacity),
		occurrences: make(map[string]ApplicationEvaluationScheduleOccurrence, occurrenceCapacity),
	}
}

func (repository *memoryApplicationEvaluationScheduleRepository) CreateSchedule(
	ctx ApplicationEvaluationContext,
	schedule ApplicationEvaluationSchedule,
	version ApplicationEvaluationScheduleVersion,
) error {
	if !ctx.WriteEnabled || validateApplicationEvaluationSchedule(ctx, schedule) != nil || validateApplicationEvaluationScheduleVersion(ctx, version) != nil ||
		!applicationEvaluationScheduleMatchesVersion(schedule, version) || schedule.RecordVersion != 1 || version.ScheduleVersion != 1 ||
		schedule.LifecycleState != applicationEvaluationScheduleStateDraft || schedule.NextDueAt != nil || ctx.ActorRef != schedule.DelegatedByUserRef ||
		schedule.CreatedByActorRef != ctx.ActorRef || schedule.UpdatedByActorRef != ctx.ActorRef || version.CreatedByActorRef != ctx.ActorRef ||
		schedule.CreatedAt != version.CreatedAt || schedule.UpdatedAt != version.CreatedAt || schedule.RequestID != ctx.RequestID ||
		schedule.AuditRef != ctx.AuditRef || version.RequestID != ctx.RequestID || version.AuditRef != ctx.AuditRef {
		return errApplicationEvaluationScheduleStoreContract
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return errApplicationEvaluationScheduleStoreUnavailable
	}
	key := applicationEvaluationScheduleKey(ctx, schedule.ScheduleID)
	if _, exists := repository.schedules[key]; exists {
		return errApplicationEvaluationScheduleVersionConflict
	}
	if len(repository.schedules) >= repository.scheduleCapacity {
		return errApplicationEvaluationScheduleStoreUnavailable
	}
	repository.schedules[key] = cloneApplicationEvaluationSchedule(schedule)
	repository.versions[key] = map[int]ApplicationEvaluationScheduleVersion{version.ScheduleVersion: cloneApplicationEvaluationScheduleVersion(version)}
	return nil
}

func (repository *memoryApplicationEvaluationScheduleRepository) ReviseSchedule(
	ctx ApplicationEvaluationContext,
	expected int,
	schedule ApplicationEvaluationSchedule,
	version ApplicationEvaluationScheduleVersion,
) (ApplicationEvaluationSchedule, bool, error) {
	if !ctx.WriteEnabled || expected < 1 || validateApplicationEvaluationSchedule(ctx, schedule) != nil || validateApplicationEvaluationScheduleVersion(ctx, version) != nil ||
		!applicationEvaluationScheduleMatchesVersion(schedule, version) || schedule.RecordVersion != expected+1 ||
		schedule.LatestScheduleVersion != version.ScheduleVersion || version.CreatedByActorRef != ctx.ActorRef ||
		ctx.ActorRef != schedule.DelegatedByUserRef || schedule.UpdatedByActorRef != ctx.ActorRef || schedule.UpdatedAt != version.CreatedAt ||
		schedule.RequestID != ctx.RequestID || schedule.AuditRef != ctx.AuditRef || version.RequestID != ctx.RequestID || version.AuditRef != ctx.AuditRef {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	key := applicationEvaluationScheduleKey(ctx, schedule.ScheduleID)
	current, found := repository.schedules[key]
	if !found {
		return ApplicationEvaluationSchedule{}, false, nil
	}
	currentVersion, currentVersionFound := repository.versions[key][current.LatestScheduleVersion]
	if !currentVersionFound || validateApplicationEvaluationSchedule(ctx, current) != nil ||
		validateApplicationEvaluationScheduleVersion(ctx, currentVersion) != nil || !applicationEvaluationScheduleMatchesVersion(current, currentVersion) {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	if current.RecordVersion != expected {
		return cloneApplicationEvaluationSchedule(current), false, errApplicationEvaluationScheduleVersionConflict
	}
	if (current.LifecycleState != applicationEvaluationScheduleStateDraft && current.LifecycleState != applicationEvaluationScheduleStatePaused) ||
		schedule.LifecycleState != current.LifecycleState || schedule.NextDueAt != nil ||
		schedule.LatestScheduleVersion != current.LatestScheduleVersion+1 || version.PreviousScheduleVersion != current.LatestScheduleVersion ||
		!sameApplicationEvaluationScheduleIdentity(current, schedule) || schedule.CreatedAt != current.CreatedAt ||
		schedule.CreatedByActorRef != current.CreatedByActorRef || !applicationEvaluationScheduleTimestampAfter(schedule.UpdatedAt, current.UpdatedAt) {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	repository.schedules[key] = cloneApplicationEvaluationSchedule(schedule)
	repository.versions[key][version.ScheduleVersion] = cloneApplicationEvaluationScheduleVersion(version)
	return cloneApplicationEvaluationSchedule(schedule), true, nil
}

func (repository *memoryApplicationEvaluationScheduleRepository) UpdateSchedule(
	ctx ApplicationEvaluationContext,
	expected int,
	schedule ApplicationEvaluationSchedule,
) (ApplicationEvaluationSchedule, bool, error) {
	if !ctx.WriteEnabled || expected < 1 || validateApplicationEvaluationSchedule(ctx, schedule) != nil || schedule.RecordVersion != expected+1 ||
		schedule.UpdatedByActorRef != ctx.ActorRef || schedule.RequestID != ctx.RequestID || schedule.AuditRef != ctx.AuditRef {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	key := applicationEvaluationScheduleKey(ctx, schedule.ScheduleID)
	current, found := repository.schedules[key]
	if !found {
		return ApplicationEvaluationSchedule{}, false, nil
	}
	currentVersion, currentVersionFound := repository.versions[key][current.LatestScheduleVersion]
	if !currentVersionFound || validateApplicationEvaluationSchedule(ctx, current) != nil ||
		validateApplicationEvaluationScheduleVersion(ctx, currentVersion) != nil || !applicationEvaluationScheduleMatchesVersion(current, currentVersion) {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	if current.RecordVersion != expected {
		return cloneApplicationEvaluationSchedule(current), false, errApplicationEvaluationScheduleVersionConflict
	}
	if !sameApplicationEvaluationScheduleDefinition(current, schedule) || schedule.CreatedAt != current.CreatedAt ||
		schedule.CreatedByActorRef != current.CreatedByActorRef || !applicationEvaluationScheduleTimestampAfter(schedule.UpdatedAt, current.UpdatedAt) ||
		!validApplicationEvaluationScheduleLifecycleTransition(current.LifecycleState, schedule.LifecycleState) {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	if current.LifecycleState == applicationEvaluationScheduleStateActive && schedule.LifecycleState == applicationEvaluationScheduleStateActive {
		if ctx.ActorRef != current.SystemActorRef || current.NextDueAt == nil || schedule.NextDueAt == nil ||
			!applicationEvaluationScheduleTimestampAfter(*schedule.NextDueAt, *current.NextDueAt) ||
			!applicationEvaluationScheduleProjectionIsExact(currentVersion.Schedule, current.NextDueAt, schedule.UpdatedAt, schedule.NextDueAt) {
			return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
		}
	} else {
		if ctx.ActorRef != current.DelegatedByUserRef {
			return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
		}
		if schedule.LifecycleState == applicationEvaluationScheduleStateActive &&
			!applicationEvaluationScheduleActivationDueIsExact(currentVersion.Schedule, schedule.UpdatedAt, schedule.NextDueAt) {
			return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
		}
	}
	repository.schedules[key] = cloneApplicationEvaluationSchedule(schedule)
	return cloneApplicationEvaluationSchedule(schedule), true, nil
}

func (repository *memoryApplicationEvaluationScheduleRepository) ReadSchedule(
	ctx ApplicationEvaluationContext,
	scheduleID string,
) (ApplicationEvaluationSchedule, bool, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	key := applicationEvaluationScheduleKey(ctx, scheduleID)
	schedule, found := repository.schedules[key]
	if !found {
		return ApplicationEvaluationSchedule{}, false, nil
	}
	version, versionFound := repository.versions[key][schedule.LatestScheduleVersion]
	if !versionFound || validateApplicationEvaluationSchedule(ctx, schedule) != nil || validateApplicationEvaluationScheduleVersion(ctx, version) != nil ||
		!applicationEvaluationScheduleMatchesVersion(schedule, version) {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	return cloneApplicationEvaluationSchedule(schedule), true, nil
}

func (repository *memoryApplicationEvaluationScheduleRepository) ReadScheduleVersion(
	ctx ApplicationEvaluationContext,
	scheduleID string,
	versionNumber int,
) (ApplicationEvaluationScheduleVersion, bool, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return ApplicationEvaluationScheduleVersion{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	value, found := repository.versions[applicationEvaluationScheduleKey(ctx, scheduleID)][versionNumber]
	if !found {
		return ApplicationEvaluationScheduleVersion{}, false, nil
	}
	if validateApplicationEvaluationScheduleVersion(ctx, value) != nil {
		return ApplicationEvaluationScheduleVersion{}, false, errApplicationEvaluationScheduleStoreContract
	}
	return cloneApplicationEvaluationScheduleVersion(value), true, nil
}

func (repository *memoryApplicationEvaluationScheduleRepository) ClaimOccurrence(
	ctx ApplicationEvaluationContext,
	due ApplicationEvaluationScheduleOccurrence,
	claimed ApplicationEvaluationScheduleOccurrence,
) (ApplicationEvaluationScheduleOccurrence, bool, error) {
	if !ctx.WriteEnabled || validateApplicationEvaluationScheduleOccurrence(ctx, due) != nil || validateApplicationEvaluationScheduleOccurrence(ctx, claimed) != nil ||
		!validApplicationEvaluationScheduleOccurrenceUpdate(due, claimed) || due.State != applicationEvaluationScheduleOccurrenceStateDue ||
		claimed.State != applicationEvaluationScheduleOccurrenceStateClaimed || ctx.ActorRef != due.SystemActorRef ||
		due.RequestID != ctx.RequestID || due.AuditRef != ctx.AuditRef || claimed.RequestID != ctx.RequestID || claimed.AuditRef != ctx.AuditRef {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreContract
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	scheduleKey := applicationEvaluationScheduleKey(ctx, due.ScheduleID)
	schedule, scheduleFound := repository.schedules[scheduleKey]
	version, versionFound := repository.versions[scheduleKey][due.ScheduleVersion]
	if !scheduleFound || !versionFound {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleNotFound
	}
	if validateApplicationEvaluationSchedule(ctx, schedule) != nil || validateApplicationEvaluationScheduleVersion(ctx, version) != nil ||
		!applicationEvaluationScheduleMatchesVersion(schedule, version) || schedule.LifecycleState != applicationEvaluationScheduleStateActive || schedule.LatestScheduleVersion != due.ScheduleVersion ||
		schedule.LatestScheduleDigest != due.ScheduleDigest || schedule.NextDueAt == nil || *schedule.NextDueAt != due.ScheduledForUTC ||
		version.ScheduleDigest != due.ScheduleDigest || version.Authorization.SystemActorRef != due.SystemActorRef ||
		version.Authorization.DelegatedByUserRef != due.DelegatedByUserRef {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreContract
	}
	occurrenceKey := applicationEvaluationScheduleOccurrenceKey(ctx, due.ScheduleID, due.ScheduleVersion, due.ScheduledForUTC)
	if current, exists := repository.occurrences[occurrenceKey]; exists {
		return cloneApplicationEvaluationScheduleOccurrence(current), false, errApplicationEvaluationScheduleClaimConflict
	}
	if len(repository.occurrences) >= repository.occurrenceCapacity {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	repository.occurrences[occurrenceKey] = cloneApplicationEvaluationScheduleOccurrence(claimed)
	return cloneApplicationEvaluationScheduleOccurrence(claimed), true, nil
}

func (repository *memoryApplicationEvaluationScheduleRepository) UpdateOccurrence(
	ctx ApplicationEvaluationContext,
	expected int,
	occurrence ApplicationEvaluationScheduleOccurrence,
) (ApplicationEvaluationScheduleOccurrence, bool, error) {
	if !ctx.WriteEnabled || expected < 1 || validateApplicationEvaluationScheduleOccurrence(ctx, occurrence) != nil || occurrence.RecordVersion != expected+1 ||
		ctx.ActorRef != occurrence.SystemActorRef || occurrence.RequestID != ctx.RequestID || occurrence.AuditRef != ctx.AuditRef {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreContract
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	key := applicationEvaluationScheduleOccurrenceKey(ctx, occurrence.ScheduleID, occurrence.ScheduleVersion, occurrence.ScheduledForUTC)
	current, found := repository.occurrences[key]
	if !found {
		return ApplicationEvaluationScheduleOccurrence{}, false, nil
	}
	if current.RecordVersion != expected {
		return cloneApplicationEvaluationScheduleOccurrence(current), false, errApplicationEvaluationScheduleVersionConflict
	}
	version, versionFound := repository.versions[applicationEvaluationScheduleKey(ctx, occurrence.ScheduleID)][occurrence.ScheduleVersion]
	if !versionFound || validateApplicationEvaluationScheduleVersion(ctx, version) != nil || version.ScheduleDigest != occurrence.ScheduleDigest ||
		version.Authorization.SystemActorRef != occurrence.SystemActorRef ||
		version.Authorization.DelegatedByUserRef != occurrence.DelegatedByUserRef || !validApplicationEvaluationScheduleOccurrenceUpdate(current, occurrence) {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreContract
	}
	repository.occurrences[key] = cloneApplicationEvaluationScheduleOccurrence(occurrence)
	return cloneApplicationEvaluationScheduleOccurrence(occurrence), true, nil
}

func (repository *memoryApplicationEvaluationScheduleRepository) ReadOccurrence(
	ctx ApplicationEvaluationContext,
	scheduleID string,
	scheduleVersion int,
	scheduledForUTC string,
) (ApplicationEvaluationScheduleOccurrence, bool, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	key := applicationEvaluationScheduleOccurrenceKey(ctx, scheduleID, scheduleVersion, scheduledForUTC)
	occurrence, found := repository.occurrences[key]
	if !found {
		return ApplicationEvaluationScheduleOccurrence{}, false, nil
	}
	version, versionFound := repository.versions[applicationEvaluationScheduleKey(ctx, scheduleID)][scheduleVersion]
	if !versionFound || validateApplicationEvaluationScheduleOccurrence(ctx, occurrence) != nil || version.ScheduleDigest != occurrence.ScheduleDigest ||
		version.Authorization.SystemActorRef != occurrence.SystemActorRef || version.Authorization.DelegatedByUserRef != occurrence.DelegatedByUserRef {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreContract
	}
	return cloneApplicationEvaluationScheduleOccurrence(occurrence), true, nil
}

func sameApplicationEvaluationScheduleIdentity(current, next ApplicationEvaluationSchedule) bool {
	return current.ScheduleID == next.ScheduleID && current.TenantRef == next.TenantRef && current.WorkspaceID == next.WorkspaceID &&
		current.Environment == next.Environment && current.ApplicationID == next.ApplicationID && current.AuthorizationModel == next.AuthorizationModel &&
		current.SystemActorRef == next.SystemActorRef && current.DelegatedByUserRef == next.DelegatedByUserRef
}

func sameApplicationEvaluationScheduleDefinition(current, next ApplicationEvaluationSchedule) bool {
	return sameApplicationEvaluationScheduleIdentity(current, next) && current.LatestScheduleVersion == next.LatestScheduleVersion &&
		current.LatestScheduleDigest == next.LatestScheduleDigest && current.PlanID == next.PlanID && current.PlanVersion == next.PlanVersion &&
		current.PlanDigest == next.PlanDigest && current.ExecutionProfile == next.ExecutionProfile && current.QuotaAPIKeyID == next.QuotaAPIKeyID
}

func validApplicationEvaluationScheduleOccurrenceUpdate(current, next ApplicationEvaluationScheduleOccurrence) bool {
	if current.RecordVersion+1 != next.RecordVersion || !validApplicationEvaluationScheduleOccurrenceTransition(current.State, next.State) ||
		current.SchemaVersion != next.SchemaVersion || current.TenantRef != next.TenantRef || current.WorkspaceID != next.WorkspaceID ||
		current.Environment != next.Environment || current.ApplicationID != next.ApplicationID || current.ScheduleID != next.ScheduleID ||
		current.ScheduleVersion != next.ScheduleVersion || current.ScheduleDigest != next.ScheduleDigest || current.ScheduledForUTC != next.ScheduledForUTC ||
		current.ClientCampaignKey != next.ClientCampaignKey || current.SystemActorRef != next.SystemActorRef ||
		current.DelegatedByUserRef != next.DelegatedByUserRef || current.CreatedAt != next.CreatedAt ||
		!applicationEvaluationScheduleTimestampAfter(next.UpdatedAt, current.UpdatedAt) {
		return false
	}
	if current.CampaignID != nil && !reflect.DeepEqual(current.CampaignID, next.CampaignID) {
		return false
	}
	return true
}

func applicationEvaluationScheduleTimestampAfter(after, before string) bool {
	afterTime, afterOK := parseApplicationEvaluationScheduleUTCTimestamp(after)
	beforeTime, beforeOK := parseApplicationEvaluationScheduleUTCTimestamp(before)
	return afterOK && beforeOK && afterTime.After(beforeTime)
}

func applicationEvaluationScheduleActivationDueIsExact(
	rule ApplicationEvaluationScheduleDailyUTC,
	updatedAt string,
	nextDueAt *string,
) bool {
	updated, updatedOK := parseApplicationEvaluationScheduleUTCTimestamp(updatedAt)
	if !updatedOK || nextDueAt == nil {
		return false
	}
	expected, err := applicationEvaluationScheduleNextDue(updated, rule)
	return err == nil && *nextDueAt == expected.Format(time.RFC3339Nano)
}

func applicationEvaluationScheduleProjectionIsExact(
	rule ApplicationEvaluationScheduleDailyUTC,
	currentNextDueAt *string,
	nextUpdatedAt string,
	nextDueAt *string,
) bool {
	currentDue, currentDueOK := parseApplicationEvaluationScheduleUTCTimestampPointer(currentNextDueAt)
	nextUpdated, nextUpdatedOK := parseApplicationEvaluationScheduleUTCTimestamp(nextUpdatedAt)
	if !currentDueOK || !nextUpdatedOK || nextUpdated.Before(currentDue) || nextDueAt == nil {
		return false
	}
	expected, err := applicationEvaluationScheduleNextDue(currentDue, rule)
	return err == nil && *nextDueAt == expected.Format(time.RFC3339Nano)
}

func applicationEvaluationScheduleKey(ctx ApplicationEvaluationContext, scheduleID string) string {
	return applicationEvaluationScopeKey(ctx) + "\x1f" + scheduleID
}

func applicationEvaluationScheduleOccurrenceKey(ctx ApplicationEvaluationContext, scheduleID string, scheduleVersion int, scheduledForUTC string) string {
	return applicationEvaluationScheduleKey(ctx, scheduleID) + "\x1f" + strconv.Itoa(scheduleVersion) + "\x1f" + scheduledForUTC
}

var _ applicationEvaluationScheduleRepository = (*memoryApplicationEvaluationScheduleRepository)(nil)
