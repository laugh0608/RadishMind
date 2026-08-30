package httpapi

import (
	"errors"
	"strings"
	"time"
)

const applicationEvaluationScheduleSystemActorRef = "system:application-evaluation-scheduler"

type ApplicationEvaluationScheduleDefinitionInput struct {
	PlanID                         string
	PlanVersion                    int
	PlanDigest                     string
	ExpectedPlanRecordVersion      int
	QuotaAPIKeyID                  string
	Schedule                       ApplicationEvaluationScheduleDailyUTC
	AcknowledgeProviderConsumption bool
}

type ApplicationEvaluationScheduleCreateInput struct {
	ApplicationEvaluationScheduleDefinitionInput
}

type ApplicationEvaluationScheduleReviseInput struct {
	ExpectedVersion int
	ApplicationEvaluationScheduleDefinitionInput
}

type ApplicationEvaluationScheduleLifecycleInput struct {
	ExpectedVersion                int
	AcknowledgeProviderConsumption bool
	AcknowledgeNoFutureOccurrences bool
}

type ApplicationEvaluationScheduleListInput struct {
	LifecycleState string
	Limit          int
	Cursor         string
}

type ApplicationEvaluationScheduleResult struct {
	Schedule             *ApplicationEvaluationSchedule
	Version              *ApplicationEvaluationScheduleVersion
	Occurrence           *ApplicationEvaluationScheduleOccurrence
	FailureCode          string
	FailureSummary       string
	CurrentRecordVersion int
	CurrentState         string
}

type ApplicationEvaluationScheduleListResult struct {
	Schedules      []ApplicationEvaluationSchedule
	NextCursor     string
	HasMore        bool
	FailureCode    string
	FailureSummary string
}

type applicationEvaluationScheduleAuthorityValidator func(ApplicationEvaluationContext, ApplicationEvaluationPlanVersion) string
type applicationEvaluationScheduleQuotaValidator func(ApplicationEvaluationContext, string) string

type applicationEvaluationScheduleService struct {
	repository        applicationEvaluationScheduleRepository
	plans             applicationEvaluationRepository
	validateAuthority applicationEvaluationScheduleAuthorityValidator
	validateQuota     applicationEvaluationScheduleQuotaValidator
	now               func() time.Time
	newScheduleID     func() (string, error)
}

func newApplicationEvaluationScheduleService(
	repository applicationEvaluationScheduleRepository,
	plans applicationEvaluationRepository,
	validateAuthority applicationEvaluationScheduleAuthorityValidator,
	validateQuota applicationEvaluationScheduleQuotaValidator,
) applicationEvaluationScheduleService {
	return applicationEvaluationScheduleService{
		repository: repository, plans: plans, validateAuthority: validateAuthority, validateQuota: validateQuota,
		now:           func() time.Time { return time.Now().UTC() },
		newScheduleID: func() (string, error) { return newApplicationEvaluationID("aesch_") },
	}
}

func (service applicationEvaluationScheduleService) Create(
	ctx ApplicationEvaluationContext,
	input ApplicationEvaluationScheduleCreateInput,
) ApplicationEvaluationScheduleResult {
	if failure := validateApplicationEvaluationMutationContext(ctx); failure != "" {
		return applicationEvaluationScheduleFailure(failure)
	}
	definition, failure := service.resolveDefinition(ctx, input.ApplicationEvaluationScheduleDefinitionInput)
	if failure != "" {
		return applicationEvaluationScheduleFailure(failure)
	}
	for attempt := 0; attempt < 3; attempt++ {
		scheduleID, err := service.newScheduleID()
		if err != nil || !applicationEvaluationScheduleIDPattern.MatchString(scheduleID) {
			return applicationEvaluationScheduleFailure(ApplicationEvaluationScheduleFailureStoreUnavailable)
		}
		at := service.currentTime().Format(time.RFC3339Nano)
		version, err := applicationEvaluationScheduleVersionFromDefinition(ctx, scheduleID, 1, at, definition)
		if err != nil {
			return applicationEvaluationScheduleFailure(ApplicationEvaluationScheduleFailureStoreContract)
		}
		schedule := applicationEvaluationScheduleFromVersion(ctx, version, 1, applicationEvaluationScheduleStateDraft, nil, at, at)
		err = service.repository.CreateSchedule(ctx, schedule, version)
		if errors.Is(err, errApplicationEvaluationScheduleVersionConflict) {
			continue
		}
		if err != nil {
			return applicationEvaluationScheduleFailure(applicationEvaluationScheduleRepositoryFailure(err))
		}
		return applicationEvaluationScheduleSuccess(schedule, &version)
	}
	return applicationEvaluationScheduleFailure(ApplicationEvaluationScheduleFailureStoreUnavailable)
}

func (service applicationEvaluationScheduleService) Revise(
	ctx ApplicationEvaluationContext,
	scheduleID string,
	input ApplicationEvaluationScheduleReviseInput,
) ApplicationEvaluationScheduleResult {
	if failure := validateApplicationEvaluationMutationContext(ctx); failure != "" {
		return applicationEvaluationScheduleFailure(failure)
	}
	scheduleID = strings.TrimSpace(scheduleID)
	if !applicationEvaluationScheduleIDPattern.MatchString(scheduleID) || input.ExpectedVersion < 1 {
		return applicationEvaluationScheduleFailure(ApplicationEvaluationFailurePayloadInvalid)
	}
	current, found, err := service.repository.ReadSchedule(ctx, scheduleID)
	if err != nil {
		return applicationEvaluationScheduleFailure(applicationEvaluationScheduleRepositoryFailure(err))
	}
	if !found {
		return applicationEvaluationScheduleFailure(ApplicationEvaluationScheduleFailureNotFound)
	}
	if current.RecordVersion != input.ExpectedVersion {
		return applicationEvaluationScheduleConflict(current)
	}
	if current.LifecycleState != applicationEvaluationScheduleStateDraft && current.LifecycleState != applicationEvaluationScheduleStatePaused {
		return applicationEvaluationScheduleConflict(current)
	}
	definition, failure := service.resolveDefinition(ctx, input.ApplicationEvaluationScheduleDefinitionInput)
	if failure != "" {
		return applicationEvaluationScheduleFailure(failure)
	}
	at := service.timeAfter(current.UpdatedAt)
	version, err := applicationEvaluationScheduleVersionFromDefinition(ctx, scheduleID, current.LatestScheduleVersion+1, at, definition)
	if err != nil {
		return applicationEvaluationScheduleFailure(ApplicationEvaluationScheduleFailureStoreContract)
	}
	next := applicationEvaluationScheduleFromVersion(ctx, version, current.RecordVersion+1, current.LifecycleState, nil, current.CreatedAt, at)
	next.CreatedByActorRef = current.CreatedByActorRef
	stored, updated, err := service.repository.ReviseSchedule(ctx, input.ExpectedVersion, next, version)
	if err != nil {
		return applicationEvaluationScheduleRepositoryResult(err, stored)
	}
	if !updated {
		return applicationEvaluationScheduleFailure(ApplicationEvaluationScheduleFailureNotFound)
	}
	return applicationEvaluationScheduleSuccess(stored, &version)
}

func (service applicationEvaluationScheduleService) Activate(
	ctx ApplicationEvaluationContext,
	scheduleID string,
	input ApplicationEvaluationScheduleLifecycleInput,
) ApplicationEvaluationScheduleResult {
	return service.transition(ctx, scheduleID, input, applicationEvaluationScheduleStateActive, applicationEvaluationScheduleStateDraft)
}

func (service applicationEvaluationScheduleService) Pause(
	ctx ApplicationEvaluationContext,
	scheduleID string,
	input ApplicationEvaluationScheduleLifecycleInput,
) ApplicationEvaluationScheduleResult {
	return service.transition(ctx, scheduleID, input, applicationEvaluationScheduleStatePaused, applicationEvaluationScheduleStateActive)
}

func (service applicationEvaluationScheduleService) Resume(
	ctx ApplicationEvaluationContext,
	scheduleID string,
	input ApplicationEvaluationScheduleLifecycleInput,
) ApplicationEvaluationScheduleResult {
	return service.transition(ctx, scheduleID, input, applicationEvaluationScheduleStateActive, applicationEvaluationScheduleStatePaused)
}

func (service applicationEvaluationScheduleService) Archive(
	ctx ApplicationEvaluationContext,
	scheduleID string,
	input ApplicationEvaluationScheduleLifecycleInput,
) ApplicationEvaluationScheduleResult {
	return service.transition(ctx, scheduleID, input, applicationEvaluationScheduleStateArchived, "")
}

func (service applicationEvaluationScheduleService) transition(
	ctx ApplicationEvaluationContext,
	scheduleID string,
	input ApplicationEvaluationScheduleLifecycleInput,
	targetState string,
	requiredState string,
) ApplicationEvaluationScheduleResult {
	if failure := validateApplicationEvaluationMutationContext(ctx); failure != "" {
		return applicationEvaluationScheduleFailure(failure)
	}
	scheduleID = strings.TrimSpace(scheduleID)
	if !applicationEvaluationScheduleIDPattern.MatchString(scheduleID) || input.ExpectedVersion < 1 {
		return applicationEvaluationScheduleFailure(ApplicationEvaluationFailurePayloadInvalid)
	}
	current, found, err := service.repository.ReadSchedule(ctx, scheduleID)
	if err != nil {
		return applicationEvaluationScheduleFailure(applicationEvaluationScheduleRepositoryFailure(err))
	}
	if !found {
		return applicationEvaluationScheduleFailure(ApplicationEvaluationScheduleFailureNotFound)
	}
	if current.RecordVersion != input.ExpectedVersion || (requiredState != "" && current.LifecycleState != requiredState) ||
		current.LifecycleState == applicationEvaluationScheduleStateArchived {
		return applicationEvaluationScheduleConflict(current)
	}
	if targetState == applicationEvaluationScheduleStateArchived && !input.AcknowledgeNoFutureOccurrences {
		return applicationEvaluationScheduleFailure(ApplicationEvaluationFailurePayloadInvalid)
	}
	version, found, err := service.repository.ReadScheduleVersion(ctx, scheduleID, current.LatestScheduleVersion)
	if err != nil {
		return applicationEvaluationScheduleFailure(applicationEvaluationScheduleRepositoryFailure(err))
	}
	if !found || !applicationEvaluationScheduleMatchesVersion(current, version) {
		return applicationEvaluationScheduleFailure(ApplicationEvaluationScheduleFailureStoreContract)
	}
	if targetState == applicationEvaluationScheduleStateActive {
		if !input.AcknowledgeProviderConsumption {
			return applicationEvaluationScheduleFailure(ApplicationEvaluationFailurePayloadInvalid)
		}
		if failure := service.revalidateActivation(ctx, version); failure != "" {
			return applicationEvaluationScheduleFailure(failure)
		}
	}
	at := service.timeAfter(current.UpdatedAt)
	next := cloneApplicationEvaluationSchedule(current)
	next.RecordVersion++
	next.LifecycleState = targetState
	next.NextDueAt = nil
	if targetState == applicationEvaluationScheduleStateActive {
		due, dueErr := applicationEvaluationScheduleNextDue(service.timeValue(at), version.Schedule)
		if dueErr != nil {
			return applicationEvaluationScheduleFailure(ApplicationEvaluationScheduleFailureStoreContract)
		}
		dueText := due.Format(time.RFC3339Nano)
		next.NextDueAt = &dueText
	}
	next.UpdatedAt, next.UpdatedByActorRef = at, ctx.ActorRef
	next.RequestID, next.AuditRef = ctx.RequestID, ctx.AuditRef
	stored, updated, err := service.repository.UpdateSchedule(ctx, input.ExpectedVersion, next)
	if err != nil {
		return applicationEvaluationScheduleRepositoryResult(err, stored)
	}
	if !updated {
		return applicationEvaluationScheduleFailure(ApplicationEvaluationScheduleFailureNotFound)
	}
	return applicationEvaluationScheduleSuccess(stored, nil)
}

func (service applicationEvaluationScheduleService) Read(ctx ApplicationEvaluationContext, scheduleID string) ApplicationEvaluationScheduleResult {
	if !validApplicationEvaluationContext(ctx) || !applicationEvaluationScheduleIDPattern.MatchString(strings.TrimSpace(scheduleID)) {
		return applicationEvaluationScheduleFailure(ApplicationEvaluationFailurePayloadInvalid)
	}
	value, found, err := service.repository.ReadSchedule(ctx, strings.TrimSpace(scheduleID))
	if err != nil {
		return applicationEvaluationScheduleFailure(applicationEvaluationScheduleRepositoryFailure(err))
	}
	if !found {
		return applicationEvaluationScheduleFailure(ApplicationEvaluationScheduleFailureNotFound)
	}
	return applicationEvaluationScheduleSuccess(value, nil)
}

func (service applicationEvaluationScheduleService) ReadVersion(
	ctx ApplicationEvaluationContext,
	scheduleID string,
	versionNumber int,
) ApplicationEvaluationScheduleResult {
	if !validApplicationEvaluationContext(ctx) || !applicationEvaluationScheduleIDPattern.MatchString(strings.TrimSpace(scheduleID)) || versionNumber < 1 {
		return applicationEvaluationScheduleFailure(ApplicationEvaluationFailurePayloadInvalid)
	}
	value, found, err := service.repository.ReadScheduleVersion(ctx, strings.TrimSpace(scheduleID), versionNumber)
	if err != nil {
		return applicationEvaluationScheduleFailure(applicationEvaluationScheduleRepositoryFailure(err))
	}
	if !found {
		return applicationEvaluationScheduleFailure(ApplicationEvaluationScheduleFailureNotFound)
	}
	return ApplicationEvaluationScheduleResult{Version: &value}
}

func (service applicationEvaluationScheduleService) ReadOccurrence(
	ctx ApplicationEvaluationContext,
	scheduleID string,
	scheduleVersion int,
	scheduledForUTC string,
) ApplicationEvaluationScheduleResult {
	scheduleID, scheduledForUTC = strings.TrimSpace(scheduleID), strings.TrimSpace(scheduledForUTC)
	if !validApplicationEvaluationContext(ctx) || !applicationEvaluationScheduleIDPattern.MatchString(scheduleID) || scheduleVersion < 1 {
		return applicationEvaluationScheduleFailure(ApplicationEvaluationFailurePayloadInvalid)
	}
	if _, ok := parseApplicationEvaluationScheduleUTCTimestamp(scheduledForUTC); !ok {
		return applicationEvaluationScheduleFailure(ApplicationEvaluationFailurePayloadInvalid)
	}
	value, found, err := service.repository.ReadOccurrence(ctx, scheduleID, scheduleVersion, scheduledForUTC)
	if err != nil {
		return applicationEvaluationScheduleFailure(applicationEvaluationScheduleRepositoryFailure(err))
	}
	if !found {
		return applicationEvaluationScheduleFailure(ApplicationEvaluationScheduleFailureNotFound)
	}
	return ApplicationEvaluationScheduleResult{Occurrence: &value, CurrentRecordVersion: value.RecordVersion, CurrentState: value.State}
}

func (service applicationEvaluationScheduleService) List(
	ctx ApplicationEvaluationContext,
	input ApplicationEvaluationScheduleListInput,
) ApplicationEvaluationScheduleListResult {
	if !validApplicationEvaluationContext(ctx) {
		return applicationEvaluationScheduleListFailure(ApplicationEvaluationFailureScopeDenied)
	}
	state := strings.TrimSpace(input.LifecycleState)
	if state == "" {
		state = applicationEvaluationScheduleStateActive
	}
	limit := input.Limit
	if limit == 0 {
		limit = applicationEvaluationDefaultListLimit
	}
	if !validApplicationEvaluationScheduleState(state) || limit < 1 || limit > applicationEvaluationMaximumListLimit {
		return applicationEvaluationScheduleListFailure(ApplicationEvaluationFailurePayloadInvalid)
	}
	filter := ApplicationEvaluationScheduleListFilter{LifecycleState: state, Limit: limit}
	if strings.TrimSpace(input.Cursor) != "" {
		cursor, err := decodeApplicationEvaluationCursor(input.Cursor)
		if err != nil || !applicationEvaluationCursorMatches(ctx, cursor, "schedules", state, limit) ||
			!applicationEvaluationScheduleIDPattern.MatchString(cursor.BeforeID) {
			return applicationEvaluationScheduleListFailure(ApplicationEvaluationFailureCursorInvalid)
		}
		if _, ok := parseApplicationEvaluationScheduleUTCTimestamp(cursor.BeforeTime); !ok {
			return applicationEvaluationScheduleListFailure(ApplicationEvaluationFailureCursorInvalid)
		}
		filter.BeforeUpdatedAt, filter.BeforeScheduleID = cursor.BeforeTime, cursor.BeforeID
	}
	page, err := service.repository.ListSchedules(ctx, filter)
	if err != nil {
		return applicationEvaluationScheduleListFailure(applicationEvaluationScheduleRepositoryFailure(err))
	}
	result := ApplicationEvaluationScheduleListResult{Schedules: page.Schedules, HasMore: page.HasMore}
	if page.HasMore && len(page.Schedules) > 0 {
		last := page.Schedules[len(page.Schedules)-1]
		result.NextCursor, _ = encodeApplicationEvaluationCursor(applicationEvaluationCursor{
			Version: 1, Kind: "schedules", TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
			Environment: ctx.Environment, ApplicationID: ctx.ApplicationID, Filter: state,
			BeforeTime: last.UpdatedAt, BeforeID: last.ScheduleID, Limit: limit,
		})
	}
	return result
}

type applicationEvaluationResolvedScheduleDefinition struct {
	planVersion   ApplicationEvaluationPlanVersion
	quotaAPIKeyID string
	schedule      ApplicationEvaluationScheduleDailyUTC
}

func (service applicationEvaluationScheduleService) resolveDefinition(
	ctx ApplicationEvaluationContext,
	input ApplicationEvaluationScheduleDefinitionInput,
) (applicationEvaluationResolvedScheduleDefinition, string) {
	input.PlanID, input.PlanDigest, input.QuotaAPIKeyID = strings.TrimSpace(input.PlanID), strings.TrimSpace(input.PlanDigest), strings.TrimSpace(input.QuotaAPIKeyID)
	if !applicationEvaluationPlanIDPattern.MatchString(input.PlanID) || input.PlanVersion < 1 || input.ExpectedPlanRecordVersion < 1 ||
		!workflowRAGDigestPattern.MatchString(input.PlanDigest) || !apiKeyIDPattern.MatchString(input.QuotaAPIKeyID) ||
		!validApplicationEvaluationScheduleDailyUTC(input.Schedule) || !input.AcknowledgeProviderConsumption {
		return applicationEvaluationResolvedScheduleDefinition{}, ApplicationEvaluationFailurePayloadInvalid
	}
	if service.plans == nil || service.repository == nil {
		return applicationEvaluationResolvedScheduleDefinition{}, ApplicationEvaluationScheduleFailureStoreUnavailable
	}
	plan, found, err := service.plans.ReadPlan(ctx, input.PlanID)
	if err != nil {
		return applicationEvaluationResolvedScheduleDefinition{}, ApplicationEvaluationScheduleFailureStoreUnavailable
	}
	if !found || plan.RecordVersion != input.ExpectedPlanRecordVersion || plan.LifecycleState != applicationEvaluationPlanStateActive ||
		plan.LatestPlanVersion != input.PlanVersion || plan.LatestPlanDigest != input.PlanDigest || plan.ExecutionProfile != applicationInteractionProfilePrompt {
		return applicationEvaluationResolvedScheduleDefinition{}, ApplicationEvaluationScheduleFailurePlanChanged
	}
	version, found, err := service.plans.ReadPlanVersion(ctx, input.PlanID, input.PlanVersion)
	if err != nil {
		return applicationEvaluationResolvedScheduleDefinition{}, ApplicationEvaluationScheduleFailureStoreUnavailable
	}
	if !found || version.PlanDigest != input.PlanDigest || version.ExecutionProfile != applicationInteractionProfilePrompt ||
		len(version.Items) < 1 || len(version.Items) > applicationEvaluationMaximumItems {
		return applicationEvaluationResolvedScheduleDefinition{}, ApplicationEvaluationScheduleFailurePlanChanged
	}
	if service.validateQuota == nil {
		return applicationEvaluationResolvedScheduleDefinition{}, ApplicationEvaluationScheduleFailureAuthorizationUnavailable
	}
	if failure := service.validateQuota(ctx, input.QuotaAPIKeyID); failure != "" {
		return applicationEvaluationResolvedScheduleDefinition{}, failure
	}
	return applicationEvaluationResolvedScheduleDefinition{planVersion: version, quotaAPIKeyID: input.QuotaAPIKeyID, schedule: input.Schedule}, ""
}

func (service applicationEvaluationScheduleService) revalidateActivation(
	ctx ApplicationEvaluationContext,
	version ApplicationEvaluationScheduleVersion,
) string {
	definition, failure := service.resolveDefinition(ctx, ApplicationEvaluationScheduleDefinitionInput{
		PlanID: version.PlanID, PlanVersion: version.PlanVersion, PlanDigest: version.PlanDigest,
		ExpectedPlanRecordVersion: version.PlanVersion, QuotaAPIKeyID: version.QuotaAPIKeyID,
		Schedule: version.Schedule, AcknowledgeProviderConsumption: true,
	})
	if failure != "" {
		return failure
	}
	if len(definition.planVersion.Items) != version.ItemCount || version.MaxProviderAttempts != len(definition.planVersion.Items) {
		return ApplicationEvaluationScheduleFailurePlanChanged
	}
	if service.validateAuthority == nil {
		return ApplicationEvaluationScheduleFailureAuthorizationUnavailable
	}
	return service.validateAuthority(ctx, definition.planVersion)
}

func applicationEvaluationScheduleVersionFromDefinition(
	ctx ApplicationEvaluationContext,
	scheduleID string,
	scheduleVersion int,
	at string,
	definition applicationEvaluationResolvedScheduleDefinition,
) (ApplicationEvaluationScheduleVersion, error) {
	version := ApplicationEvaluationScheduleVersion{
		SchemaVersion: applicationEvaluationScheduleVersionSchemaVersion, ScheduleID: scheduleID,
		ScheduleVersion: scheduleVersion, PreviousScheduleVersion: scheduleVersion - 1,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Environment: ctx.Environment, ApplicationID: ctx.ApplicationID,
		PlanID: definition.planVersion.PlanID, PlanVersion: definition.planVersion.PlanVersion, PlanDigest: definition.planVersion.PlanDigest,
		ExecutionProfile: applicationInteractionProfilePrompt, QuotaAPIKeyID: definition.quotaAPIKeyID, Schedule: definition.schedule,
		ItemCount: len(definition.planVersion.Items), MaxProviderAttempts: len(definition.planVersion.Items),
		MissedWindowPolicy: applicationEvaluationScheduleMissedWindowPolicy, OverlapPolicy: applicationEvaluationScheduleOverlapPolicy,
		Authorization: ApplicationEvaluationScheduleAuthorization{
			Model: applicationEvaluationScheduleAuthorizationModel, SystemActorRef: applicationEvaluationScheduleSystemActorRef,
			DelegatedByUserRef: ctx.ActorRef, RequiredPermissions: append([]string(nil), applicationEvaluationScheduleRequiredPermissions...),
			RevalidationPolicy:    applicationEvaluationScheduleRevalidationPolicy,
			APIKeyOwnershipPolicy: applicationEvaluationScheduleAPIKeyOwnershipPolicy,
			RevocationPolicy:      applicationEvaluationScheduleRevocationPolicy,
		},
		CreatedAt: at, CreatedByActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	var err error
	version.ScheduleDigest, err = applicationEvaluationScheduleDigest(version)
	return version, err
}

func applicationEvaluationScheduleFromVersion(
	ctx ApplicationEvaluationContext,
	version ApplicationEvaluationScheduleVersion,
	recordVersion int,
	state string,
	nextDueAt *string,
	createdAt string,
	updatedAt string,
) ApplicationEvaluationSchedule {
	return ApplicationEvaluationSchedule{
		SchemaVersion: applicationEvaluationScheduleSchemaVersion, ScheduleID: version.ScheduleID, RecordVersion: recordVersion,
		LatestScheduleVersion: version.ScheduleVersion, LatestScheduleDigest: version.ScheduleDigest,
		TenantRef: version.TenantRef, WorkspaceID: version.WorkspaceID, Environment: version.Environment, ApplicationID: version.ApplicationID,
		PlanID: version.PlanID, PlanVersion: version.PlanVersion, PlanDigest: version.PlanDigest, ExecutionProfile: version.ExecutionProfile,
		QuotaAPIKeyID: version.QuotaAPIKeyID, AuthorizationModel: version.Authorization.Model,
		SystemActorRef: version.Authorization.SystemActorRef, DelegatedByUserRef: version.Authorization.DelegatedByUserRef,
		LifecycleState: state, NextDueAt: nextDueAt, CreatedAt: createdAt, UpdatedAt: updatedAt,
		CreatedByActorRef: version.Authorization.DelegatedByUserRef, UpdatedByActorRef: ctx.ActorRef,
		RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
}

func (service applicationEvaluationScheduleService) currentTime() time.Time {
	if service.now == nil {
		return time.Now().UTC()
	}
	return service.now().UTC()
}

func (service applicationEvaluationScheduleService) timeAfter(previous string) string {
	now := service.currentTime()
	if parsed, ok := parseApplicationEvaluationScheduleUTCTimestamp(previous); ok && !now.After(parsed) {
		now = parsed.Add(time.Nanosecond)
	}
	return now.Format(time.RFC3339Nano)
}

func (service applicationEvaluationScheduleService) timeValue(value string) time.Time {
	parsed, _ := parseApplicationEvaluationScheduleUTCTimestamp(value)
	return parsed
}

func applicationEvaluationScheduleSuccess(
	schedule ApplicationEvaluationSchedule,
	version *ApplicationEvaluationScheduleVersion,
) ApplicationEvaluationScheduleResult {
	copy := cloneApplicationEvaluationSchedule(schedule)
	result := ApplicationEvaluationScheduleResult{Schedule: &copy, CurrentRecordVersion: copy.RecordVersion, CurrentState: copy.LifecycleState}
	if version != nil {
		versionCopy := cloneApplicationEvaluationScheduleVersion(*version)
		result.Version = &versionCopy
	}
	return result
}

func applicationEvaluationScheduleConflict(schedule ApplicationEvaluationSchedule) ApplicationEvaluationScheduleResult {
	result := applicationEvaluationScheduleFailure(ApplicationEvaluationScheduleFailureVersionConflict)
	result.CurrentRecordVersion, result.CurrentState = schedule.RecordVersion, schedule.LifecycleState
	return result
}

func applicationEvaluationScheduleRepositoryResult(err error, current ApplicationEvaluationSchedule) ApplicationEvaluationScheduleResult {
	result := applicationEvaluationScheduleFailure(applicationEvaluationScheduleRepositoryFailure(err))
	if current.RecordVersion > 0 {
		result.CurrentRecordVersion, result.CurrentState = current.RecordVersion, current.LifecycleState
	}
	return result
}

func applicationEvaluationScheduleRepositoryFailure(err error) string {
	switch {
	case errors.Is(err, errApplicationEvaluationScheduleNotFound):
		return ApplicationEvaluationScheduleFailureNotFound
	case errors.Is(err, errApplicationEvaluationScheduleVersionConflict):
		return ApplicationEvaluationScheduleFailureVersionConflict
	case errors.Is(err, errApplicationEvaluationScheduleClaimConflict):
		return ApplicationEvaluationScheduleFailureClaimConflict
	case errors.Is(err, errApplicationEvaluationScheduleStoreContract):
		return ApplicationEvaluationScheduleFailureStoreContract
	default:
		return ApplicationEvaluationScheduleFailureStoreUnavailable
	}
}

func applicationEvaluationScheduleFailure(code string) ApplicationEvaluationScheduleResult {
	return ApplicationEvaluationScheduleResult{FailureCode: code, FailureSummary: applicationEvaluationScheduleFailureSummary(code)}
}

func applicationEvaluationScheduleListFailure(code string) ApplicationEvaluationScheduleListResult {
	return ApplicationEvaluationScheduleListResult{Schedules: []ApplicationEvaluationSchedule{}, FailureCode: code, FailureSummary: applicationEvaluationScheduleFailureSummary(code)}
}

func applicationEvaluationScheduleFailureSummary(code string) string {
	switch code {
	case ApplicationEvaluationScheduleFailureAuthorizationUnavailable:
		return "Schedule authorization dependencies are unavailable."
	case ApplicationEvaluationScheduleFailureMembershipDenied:
		return "Delegated workspace membership is not currently authorized."
	case ApplicationEvaluationScheduleFailurePlanChanged:
		return "The exact active Prompt evaluation plan no longer matches the schedule."
	case ApplicationEvaluationScheduleFailureAuthorityChanged:
		return "The active Prompt application authority no longer matches."
	case ApplicationEvaluationScheduleFailureQuotaConsumerInvalid:
		return "The delegated user's API key is not an active quota consumer for this application."
	case ApplicationEvaluationScheduleFailureQuotaDenied:
		return "The scheduled evaluation quota admission was denied."
	case ApplicationEvaluationScheduleFailureClaimConflict:
		return "The occurrence was already claimed."
	case ApplicationEvaluationScheduleFailureStoreUnavailable:
		return "The schedule store is unavailable."
	case ApplicationEvaluationScheduleFailureStoreContract:
		return "The schedule store contract does not match the canonical record."
	case ApplicationEvaluationScheduleFailureNotFound:
		return "The schedule record was not found."
	case ApplicationEvaluationScheduleFailureVersionConflict:
		return "The schedule record version changed."
	default:
		return applicationEvaluationFailureSummary(code)
	}
}
