package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"strconv"
	"sync"
	"time"
)

const (
	applicationEvaluationScheduleRunnerPollInterval = 30 * time.Second
	applicationEvaluationScheduleRunnerBatchSize    = 50
)

type applicationEvaluationScheduleRunnerRevalidator func(
	context.Context,
	ApplicationEvaluationContext,
	ApplicationEvaluationScheduleVersion,
) string

type applicationEvaluationScheduleRunner struct {
	repository applicationEvaluationScheduleRepository
	campaigns  applicationEvaluationCampaignService
	revalidate applicationEvaluationScheduleRunnerRevalidator

	pollInterval  time.Duration
	recoveryAfter time.Duration
	batchSize     int
	now           func() time.Time

	startOnce sync.Once
	stopOnce  sync.Once
	cancel    context.CancelFunc
	done      chan struct{}

	stateMu     sync.RWMutex
	lastFailure string
}

func newApplicationEvaluationScheduleRunner(
	repository applicationEvaluationScheduleRepository,
	campaigns applicationEvaluationCampaignService,
	revalidate applicationEvaluationScheduleRunnerRevalidator,
) *applicationEvaluationScheduleRunner {
	return &applicationEvaluationScheduleRunner{
		repository:    repository,
		campaigns:     campaigns,
		revalidate:    revalidate,
		pollInterval:  applicationEvaluationScheduleRunnerPollInterval,
		recoveryAfter: 2 * applicationEvaluationScheduleRunnerPollInterval,
		batchSize:     applicationEvaluationScheduleRunnerBatchSize,
		now:           func() time.Time { return time.Now().UTC() },
		done:          make(chan struct{}),
	}
}

func (runner *applicationEvaluationScheduleRunner) Start(parent context.Context) error {
	if runner == nil || runner.repository == nil || runner.campaigns.repository == nil || runner.revalidate == nil || parent == nil {
		return errApplicationEvaluationScheduleStoreContract
	}
	started := false
	runner.startOnce.Do(func() {
		started = true
		ctx, cancel := context.WithCancel(parent)
		runner.cancel = cancel
		go runner.loop(ctx)
	})
	if !started {
		return errApplicationEvaluationScheduleStoreContract
	}
	return nil
}

func (runner *applicationEvaluationScheduleRunner) Stop() {
	if runner == nil {
		return
	}
	runner.stopOnce.Do(func() {
		if runner.cancel == nil {
			return
		}
		runner.cancel()
		<-runner.done
	})
}

func (runner *applicationEvaluationScheduleRunner) loop(ctx context.Context) {
	defer close(runner.done)
	interval := runner.pollInterval
	if interval <= 0 {
		interval = applicationEvaluationScheduleRunnerPollInterval
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if err := runner.PollOnce(ctx); err != nil {
			runner.recordFailure(applicationEvaluationScheduleRepositoryFailure(err))
		} else {
			runner.recordFailure("")
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (runner *applicationEvaluationScheduleRunner) PollOnce(ctx context.Context) error {
	if runner == nil || runner.repository == nil || runner.campaigns.repository == nil || runner.revalidate == nil || ctx == nil {
		return errApplicationEvaluationScheduleStoreContract
	}
	now := runner.currentTime()
	limit := runner.batchSize
	if limit <= 0 {
		limit = applicationEvaluationScheduleRunnerBatchSize
	}
	openPage, err := runner.repository.ListOpenOccurrences(ctx, limit)
	if err != nil {
		return err
	}
	for _, occurrence := range openPage.Occurrences {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if runner.occurrenceReadyForRecovery(occurrence, now) {
			if err = runner.recoverOccurrence(ctx, occurrence, now); err != nil {
				return err
			}
		}
	}
	openPage, err = runner.repository.ListOpenOccurrences(ctx, limit)
	if err != nil {
		return err
	}
	if openPage.HasMore {
		return nil
	}
	openBySchedule := make(map[string]ApplicationEvaluationScheduleOccurrence, len(openPage.Occurrences))
	for _, occurrence := range openPage.Occurrences {
		openBySchedule[applicationEvaluationRunnerScheduleScopeKey(occurrence)] = occurrence
	}
	duePage, err := runner.repository.ListDueSchedules(ctx, now.Format(time.RFC3339Nano), limit)
	if err != nil {
		return err
	}
	for _, schedule := range duePage.Schedules {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		open, overlaps := openBySchedule[applicationEvaluationRunnerScheduleScopeKeyFromSchedule(schedule)]
		if err = runner.claimDueSchedule(ctx, schedule, open, overlaps, now); err != nil {
			return err
		}
	}
	return nil
}

func (runner *applicationEvaluationScheduleRunner) claimDueSchedule(
	requestContext context.Context,
	schedule ApplicationEvaluationSchedule,
	open ApplicationEvaluationScheduleOccurrence,
	hasOpen bool,
	now time.Time,
) error {
	if schedule.NextDueAt == nil {
		return errApplicationEvaluationScheduleStoreContract
	}
	ctx := applicationEvaluationScheduleRunnerContext(requestContext, schedule, *schedule.NextDueAt)
	version, found, err := runner.repository.ReadScheduleVersion(ctx, schedule.ScheduleID, schedule.LatestScheduleVersion)
	if err != nil {
		return err
	}
	if !found || !applicationEvaluationScheduleMatchesVersion(schedule, version) {
		return errApplicationEvaluationScheduleStoreContract
	}
	due, claimed := applicationEvaluationScheduleRunnerClaim(ctx, schedule, version, now)
	stored, won, err := runner.repository.ClaimOccurrence(ctx, due, claimed)
	if err != nil && !errors.Is(err, errApplicationEvaluationScheduleClaimConflict) {
		if errors.Is(err, errApplicationEvaluationScheduleStoreContract) {
			current, found, readErr := runner.repository.ReadSchedule(ctx, schedule.ScheduleID)
			if readErr == nil && found && (current.RecordVersion != schedule.RecordVersion || current.NextDueAt == nil ||
				*current.NextDueAt != due.ScheduledForUTC) {
				return nil
			}
		}
		return err
	}
	if !won {
		if stored.ScheduleID != schedule.ScheduleID || stored.ScheduleVersion != schedule.LatestScheduleVersion ||
			stored.ScheduledForUTC != *schedule.NextDueAt {
			return errApplicationEvaluationScheduleStoreContract
		}
		if applicationEvaluationScheduleOccurrenceTerminal(stored.State) {
			if projectionErr := runner.advanceProjection(ctx, schedule, version, now); projectionErr != nil {
				if !errors.Is(projectionErr, errApplicationEvaluationScheduleVersionConflict) {
					return projectionErr
				}
				current, found, readErr := runner.repository.ReadSchedule(ctx, schedule.ScheduleID)
				if readErr != nil {
					return readErr
				}
				if !found || current.RecordVersion == schedule.RecordVersion && current.NextDueAt != nil && *current.NextDueAt == stored.ScheduledForUTC {
					return errApplicationEvaluationScheduleStoreContract
				}
			}
		}
		return nil
	}
	if err = runner.advanceProjection(ctx, schedule, version, now); err != nil {
		if errors.Is(err, errApplicationEvaluationScheduleVersionConflict) {
			return runner.finishOccurrence(ctx, stored, applicationEvaluationScheduleOccurrenceStateInterrupted,
				ApplicationEvaluationScheduleFailureAuthorizationUnavailable, now)
		}
		return err
	}
	if runner.windowWasMissed(stored, version, now) {
		return runner.finishOccurrence(ctx, stored, applicationEvaluationScheduleOccurrenceStateSkipped,
			ApplicationEvaluationScheduleFailureMissedWindow, now)
	}
	if hasOpen && (open.ScheduleVersion != stored.ScheduleVersion || open.ScheduledForUTC != stored.ScheduledForUTC) {
		return runner.finishOccurrence(ctx, stored, applicationEvaluationScheduleOccurrenceStateSkipped,
			ApplicationEvaluationScheduleFailureOverlapBlocked, now)
	}
	return runner.executeClaimedOccurrence(requestContext, stored, version, now)
}

func (runner *applicationEvaluationScheduleRunner) recoverOccurrence(
	requestContext context.Context,
	occurrence ApplicationEvaluationScheduleOccurrence,
	now time.Time,
) error {
	ctx := applicationEvaluationScheduleRunnerContextForOccurrence(requestContext, occurrence)
	version, found, err := runner.repository.ReadScheduleVersion(ctx, occurrence.ScheduleID, occurrence.ScheduleVersion)
	if err != nil {
		return err
	}
	if !found || version.ScheduleDigest != occurrence.ScheduleDigest ||
		version.Authorization.SystemActorRef != occurrence.SystemActorRef ||
		version.Authorization.DelegatedByUserRef != occurrence.DelegatedByUserRef {
		return runner.finishOccurrence(ctx, occurrence, applicationEvaluationScheduleOccurrenceStateInterrupted,
			ApplicationEvaluationScheduleFailureStoreContract, now)
	}
	return runner.observeExistingCampaign(requestContext, occurrence, version, now)
}

func (runner *applicationEvaluationScheduleRunner) executeClaimedOccurrence(
	requestContext context.Context,
	occurrence ApplicationEvaluationScheduleOccurrence,
	version ApplicationEvaluationScheduleVersion,
	now time.Time,
) error {
	ctx := applicationEvaluationScheduleRunnerDelegatedContext(requestContext, occurrence)
	if failure := runner.revalidate(requestContext, ctx, version); failure != "" {
		return runner.finishOccurrence(applicationEvaluationScheduleRunnerContextForOccurrence(requestContext, occurrence), occurrence,
			applicationEvaluationScheduleOccurrenceStateFailed, failure, now)
	}
	plan, found, err := runner.campaigns.repository.ReadPlan(ctx, version.PlanID)
	if err != nil {
		return err
	}
	if !found || plan.LifecycleState != applicationEvaluationPlanStateActive || plan.LatestPlanVersion != version.PlanVersion ||
		plan.LatestPlanDigest != version.PlanDigest || plan.ExecutionProfile != version.ExecutionProfile {
		return runner.finishOccurrence(applicationEvaluationScheduleRunnerContextForOccurrence(requestContext, occurrence), occurrence,
			applicationEvaluationScheduleOccurrenceStateFailed, ApplicationEvaluationScheduleFailurePlanChanged, now)
	}
	result := runner.campaigns.Execute(ctx, ApplicationEvaluationCampaignExecuteInput{
		PlanID: version.PlanID, PlanVersion: version.PlanVersion, PlanDigest: version.PlanDigest,
		ExpectedPlanRecordVersion: plan.RecordVersion, ClientCampaignKey: occurrence.ClientCampaignKey,
		QuotaAPIKeyID: version.QuotaAPIKeyID, AcknowledgeSequentialExecution: true, AcknowledgeQuotaConsumption: true,
	})
	if result.Campaign == nil {
		campaign, exists, readErr := runner.campaigns.repository.ReadCampaign(ctx,
			applicationEvaluationDeterministicCampaignID(ctx, occurrence.ClientCampaignKey))
		if readErr != nil {
			return readErr
		}
		if exists {
			return runner.attachAndFinishCampaign(requestContext, occurrence, version, campaign, now)
		}
		return runner.finishOccurrence(applicationEvaluationScheduleRunnerContextForOccurrence(requestContext, occurrence), occurrence,
			applicationEvaluationScheduleOccurrenceStateFailed, applicationEvaluationScheduleCampaignFailure(result.FailureCode), now)
	}
	return runner.attachAndFinishCampaign(requestContext, occurrence, version, *result.Campaign, now)
}

func (runner *applicationEvaluationScheduleRunner) observeExistingCampaign(
	requestContext context.Context,
	occurrence ApplicationEvaluationScheduleOccurrence,
	version ApplicationEvaluationScheduleVersion,
	now time.Time,
) error {
	ctx := applicationEvaluationScheduleRunnerDelegatedContext(requestContext, occurrence)
	campaignID := applicationEvaluationDeterministicCampaignID(ctx, occurrence.ClientCampaignKey)
	campaign, found, err := runner.campaigns.repository.ReadCampaign(ctx, campaignID)
	if err != nil {
		return err
	}
	if !found {
		return runner.finishOccurrence(applicationEvaluationScheduleRunnerContextForOccurrence(requestContext, occurrence), occurrence,
			applicationEvaluationScheduleOccurrenceStateInterrupted, ApplicationEvaluationScheduleFailureCampaignInterrupted, now)
	}
	if campaign.State == applicationEvaluationCampaignStateRunning {
		result := runner.campaigns.Reconcile(ctx, campaign.CampaignID, campaign.RecordVersion)
		if result.Campaign == nil {
			return errApplicationEvaluationScheduleStoreUnavailable
		}
		campaign = *result.Campaign
	}
	return runner.attachAndFinishCampaign(requestContext, occurrence, version, campaign, now)
}

func (runner *applicationEvaluationScheduleRunner) attachAndFinishCampaign(
	requestContext context.Context,
	occurrence ApplicationEvaluationScheduleOccurrence,
	version ApplicationEvaluationScheduleVersion,
	campaign ApplicationEvaluationCampaign,
	now time.Time,
) error {
	ctx := applicationEvaluationScheduleRunnerContextForOccurrence(requestContext, occurrence)
	wantID := applicationEvaluationDeterministicCampaignID(applicationEvaluationScheduleRunnerDelegatedContext(requestContext, occurrence), occurrence.ClientCampaignKey)
	if campaign.CampaignID != wantID || campaign.ClientCampaignKey != occurrence.ClientCampaignKey ||
		campaign.PlanID != version.PlanID || campaign.PlanVersion != version.PlanVersion || campaign.PlanDigest != version.PlanDigest ||
		campaign.QuotaAPIKeyID != version.QuotaAPIKeyID || campaign.CreatedByActorRef != occurrence.DelegatedByUserRef {
		return runner.finishOccurrence(ctx, occurrence, applicationEvaluationScheduleOccurrenceStateInterrupted,
			ApplicationEvaluationScheduleFailureStoreContract, now)
	}
	var err error
	if occurrence.State == applicationEvaluationScheduleOccurrenceStateClaimed {
		occurrence, err = runner.transitionOccurrence(ctx, occurrence, applicationEvaluationScheduleOccurrenceStateCampaignCreated, &campaign.CampaignID, "", now)
		if err != nil {
			return err
		}
	}
	if occurrence.State == applicationEvaluationScheduleOccurrenceStateCampaignCreated {
		occurrence, err = runner.transitionOccurrence(ctx, occurrence, applicationEvaluationScheduleOccurrenceStateObserving, &campaign.CampaignID, "", now)
		if err != nil {
			return err
		}
	}
	switch campaign.State {
	case applicationEvaluationCampaignStateSucceeded:
		return runner.finishOccurrence(ctx, occurrence, applicationEvaluationScheduleOccurrenceStateSucceeded, "", now)
	case applicationEvaluationCampaignStateFailed:
		return runner.finishOccurrence(ctx, occurrence, applicationEvaluationScheduleOccurrenceStateFailed,
			applicationEvaluationScheduleCampaignFailure(campaign.FailureCode), now)
	case applicationEvaluationCampaignStateInterrupted, applicationEvaluationCampaignStatePending, applicationEvaluationCampaignStateRunning:
		return runner.finishOccurrence(ctx, occurrence, applicationEvaluationScheduleOccurrenceStateInterrupted,
			ApplicationEvaluationScheduleFailureCampaignInterrupted, now)
	default:
		return runner.finishOccurrence(ctx, occurrence, applicationEvaluationScheduleOccurrenceStateInterrupted,
			ApplicationEvaluationScheduleFailureStoreContract, now)
	}
}

func (runner *applicationEvaluationScheduleRunner) advanceProjection(
	ctx ApplicationEvaluationContext,
	schedule ApplicationEvaluationSchedule,
	version ApplicationEvaluationScheduleVersion,
	now time.Time,
) error {
	updatedAt := applicationEvaluationScheduleRunnerTimeAfter(now, schedule.UpdatedAt)
	nextDue, err := applicationEvaluationScheduleNextDue(updatedAt, version.Schedule)
	if err != nil {
		return errApplicationEvaluationScheduleStoreContract
	}
	next := cloneApplicationEvaluationSchedule(schedule)
	next.RecordVersion++
	next.UpdatedAt = updatedAt.Format(time.RFC3339Nano)
	nextDueValue := nextDue.Format(time.RFC3339Nano)
	next.NextDueAt = &nextDueValue
	next.UpdatedByActorRef = ctx.ActorRef
	next.RequestID, next.AuditRef = ctx.RequestID, ctx.AuditRef
	_, updated, updateErr := runner.repository.UpdateSchedule(ctx, schedule.RecordVersion, next)
	if updateErr != nil {
		return updateErr
	}
	if !updated {
		return errApplicationEvaluationScheduleVersionConflict
	}
	return nil
}

func (runner *applicationEvaluationScheduleRunner) transitionOccurrence(
	ctx ApplicationEvaluationContext,
	occurrence ApplicationEvaluationScheduleOccurrence,
	state string,
	campaignID *string,
	failure string,
	now time.Time,
) (ApplicationEvaluationScheduleOccurrence, error) {
	next := cloneApplicationEvaluationScheduleOccurrence(occurrence)
	next.RecordVersion++
	next.State = state
	next.CampaignID = cloneApplicationEvaluationScheduleString(campaignID)
	next.UpdatedAt = applicationEvaluationScheduleRunnerTimeAfter(now, occurrence.UpdatedAt).Format(time.RFC3339Nano)
	next.RequestID, next.AuditRef = ctx.RequestID, ctx.AuditRef
	if failure != "" {
		next.FailureCode = cloneApplicationEvaluationScheduleString(&failure)
	} else {
		next.FailureCode = nil
	}
	if applicationEvaluationScheduleOccurrenceTerminal(state) {
		completedAt := next.UpdatedAt
		next.CompletedAt = &completedAt
	} else {
		next.CompletedAt = nil
	}
	stored, updated, err := runner.repository.UpdateOccurrence(ctx, occurrence.RecordVersion, next)
	if err != nil {
		return ApplicationEvaluationScheduleOccurrence{}, err
	}
	if !updated {
		return ApplicationEvaluationScheduleOccurrence{}, errApplicationEvaluationScheduleVersionConflict
	}
	return stored, nil
}

func (runner *applicationEvaluationScheduleRunner) finishOccurrence(
	ctx ApplicationEvaluationContext,
	occurrence ApplicationEvaluationScheduleOccurrence,
	state string,
	failure string,
	now time.Time,
) error {
	if applicationEvaluationScheduleOccurrenceTerminal(occurrence.State) {
		return nil
	}
	_, err := runner.transitionOccurrence(ctx, occurrence, state, occurrence.CampaignID, failure, now)
	return err
}

func (runner *applicationEvaluationScheduleRunner) windowWasMissed(
	occurrence ApplicationEvaluationScheduleOccurrence,
	version ApplicationEvaluationScheduleVersion,
	now time.Time,
) bool {
	scheduledFor, ok := parseApplicationEvaluationScheduleUTCTimestamp(occurrence.ScheduledForUTC)
	if !ok {
		return true
	}
	nextWindow, err := applicationEvaluationScheduleNextDue(scheduledFor, version.Schedule)
	return err != nil || !nextWindow.After(now)
}

func (runner *applicationEvaluationScheduleRunner) occurrenceReadyForRecovery(
	occurrence ApplicationEvaluationScheduleOccurrence,
	now time.Time,
) bool {
	updatedAt, ok := parseApplicationEvaluationScheduleUTCTimestamp(occurrence.UpdatedAt)
	if !ok {
		return true
	}
	recoveryAfter := runner.recoveryAfter
	if recoveryAfter < 0 {
		recoveryAfter = 0
	}
	return !updatedAt.Add(recoveryAfter).After(now)
}

func (runner *applicationEvaluationScheduleRunner) currentTime() time.Time {
	if runner.now == nil {
		return time.Now().UTC()
	}
	return runner.now().UTC()
}

func (runner *applicationEvaluationScheduleRunner) recordFailure(failure string) {
	runner.stateMu.Lock()
	previous := runner.lastFailure
	runner.lastFailure = failure
	runner.stateMu.Unlock()
	if failure != "" && failure != previous {
		log.Printf("radishmind_application_evaluation_schedule_runner failure_code=%s", failure)
	}
}

func applicationEvaluationScheduleRunnerClaim(
	ctx ApplicationEvaluationContext,
	schedule ApplicationEvaluationSchedule,
	version ApplicationEvaluationScheduleVersion,
	now time.Time,
) (ApplicationEvaluationScheduleOccurrence, ApplicationEvaluationScheduleOccurrence) {
	createdAt := applicationEvaluationScheduleRunnerTimeAfter(now, *schedule.NextDueAt).Format(time.RFC3339Nano)
	due := ApplicationEvaluationScheduleOccurrence{
		SchemaVersion: applicationEvaluationScheduleOccurrenceSchemaVersion, RecordVersion: 1,
		TenantRef: schedule.TenantRef, WorkspaceID: schedule.WorkspaceID, Environment: schedule.Environment, ApplicationID: schedule.ApplicationID,
		ScheduleID: schedule.ScheduleID, ScheduleVersion: version.ScheduleVersion, ScheduleDigest: version.ScheduleDigest,
		ScheduledForUTC: *schedule.NextDueAt, State: applicationEvaluationScheduleOccurrenceStateDue,
		ClientCampaignKey: applicationEvaluationScheduleClientCampaignKey(schedule.ScheduleID, version.ScheduleVersion, *schedule.NextDueAt),
		SystemActorRef:    schedule.SystemActorRef, DelegatedByUserRef: schedule.DelegatedByUserRef,
		CreatedAt: createdAt, UpdatedAt: createdAt, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	claimed := cloneApplicationEvaluationScheduleOccurrence(due)
	claimed.RecordVersion = 2
	claimed.State = applicationEvaluationScheduleOccurrenceStateClaimed
	claimedAt := applicationEvaluationScheduleRunnerTimeAfter(now, due.UpdatedAt).Format(time.RFC3339Nano)
	claimed.ClaimedAt = &claimedAt
	claimed.UpdatedAt = claimedAt
	return due, claimed
}

func applicationEvaluationScheduleRunnerContext(
	requestContext context.Context,
	schedule ApplicationEvaluationSchedule,
	scheduledFor string,
) ApplicationEvaluationContext {
	requestID, auditRef := applicationEvaluationScheduleRunnerTrace(schedule.ScheduleID, schedule.LatestScheduleVersion, scheduledFor)
	return ApplicationEvaluationContext{
		RequestContext: requestContext, RequestID: requestID, TenantRef: schedule.TenantRef, WorkspaceID: schedule.WorkspaceID,
		Environment: schedule.Environment, ApplicationID: schedule.ApplicationID, ActorRef: schedule.SystemActorRef,
		AuditRef: auditRef, WriteEnabled: true,
	}
}

func applicationEvaluationScheduleRunnerContextForOccurrence(
	requestContext context.Context,
	occurrence ApplicationEvaluationScheduleOccurrence,
) ApplicationEvaluationContext {
	return ApplicationEvaluationContext{
		RequestContext: requestContext, RequestID: occurrence.RequestID, TenantRef: occurrence.TenantRef, WorkspaceID: occurrence.WorkspaceID,
		Environment: occurrence.Environment, ApplicationID: occurrence.ApplicationID, ActorRef: occurrence.SystemActorRef,
		AuditRef: occurrence.AuditRef, WriteEnabled: true,
	}
}

func applicationEvaluationScheduleRunnerDelegatedContext(
	requestContext context.Context,
	occurrence ApplicationEvaluationScheduleOccurrence,
) ApplicationEvaluationContext {
	return ApplicationEvaluationContext{
		RequestContext: requestContext, RequestID: occurrence.RequestID, TenantRef: occurrence.TenantRef, WorkspaceID: occurrence.WorkspaceID,
		Environment: occurrence.Environment, ApplicationID: occurrence.ApplicationID, ActorRef: occurrence.DelegatedByUserRef,
		AuditRef: occurrence.AuditRef, WriteEnabled: true,
		ScheduleExecution: &ApplicationEvaluationScheduleExecutionRef{
			AuthorizationModel: applicationEvaluationScheduleAuthorizationModel, ScheduleID: occurrence.ScheduleID,
			ScheduleVersion: occurrence.ScheduleVersion, ScheduleDigest: occurrence.ScheduleDigest,
			ScheduledForUTC: occurrence.ScheduledForUTC, ClientCampaignKey: occurrence.ClientCampaignKey,
			SystemActorRef: occurrence.SystemActorRef, DelegatedByUserRef: occurrence.DelegatedByUserRef,
		},
	}
}

func applicationEvaluationScheduleRunnerTrace(scheduleID string, version int, scheduledFor string) (string, string) {
	digest := sha256.Sum256([]byte(scheduleID + "\x00" + scheduledFor))
	suffix := hex.EncodeToString(digest[:10])
	return "request_schedule_" + suffix, "audit_schedule_v" + strconv.Itoa(version) + "_" + suffix
}

func applicationEvaluationScheduleRunnerTimeAfter(now time.Time, previous string) time.Time {
	now = now.UTC()
	if parsed, ok := parseApplicationEvaluationScheduleUTCTimestamp(previous); ok && !now.After(parsed) {
		return parsed.Add(time.Nanosecond)
	}
	return now
}

func applicationEvaluationScheduleOccurrenceTerminal(state string) bool {
	switch state {
	case applicationEvaluationScheduleOccurrenceStateSucceeded, applicationEvaluationScheduleOccurrenceStateFailed,
		applicationEvaluationScheduleOccurrenceStateInterrupted, applicationEvaluationScheduleOccurrenceStateSkipped:
		return true
	default:
		return false
	}
}

func applicationEvaluationScheduleCampaignFailure(code string) string {
	if gatewayRequestQuotaFailureCodeFromValue(code) != "" {
		return ApplicationEvaluationScheduleFailureQuotaDenied
	}
	switch code {
	case ApplicationEvaluationFailureAuthorityChanged, PromptApplicationRuntimeFailureAuthorityChanged:
		return ApplicationEvaluationScheduleFailureAuthorityChanged
	case ApplicationEvaluationFailureStoreUnavailable, PromptApplicationRuntimeFailureStoreUnavailable:
		return ApplicationEvaluationScheduleFailureStoreUnavailable
	case ApplicationEvaluationFailureStoreContract, PromptApplicationRuntimeFailureStoreContract:
		return ApplicationEvaluationScheduleFailureStoreContract
	default:
		return ApplicationEvaluationScheduleFailureCampaignFailed
	}
}

func applicationEvaluationRunnerScheduleScopeKey(occurrence ApplicationEvaluationScheduleOccurrence) string {
	return occurrence.TenantRef + "\x00" + occurrence.WorkspaceID + "\x00" + occurrence.Environment + "\x00" +
		occurrence.ApplicationID + "\x00" + occurrence.ScheduleID
}

func applicationEvaluationRunnerScheduleScopeKeyFromSchedule(schedule ApplicationEvaluationSchedule) string {
	return schedule.TenantRef + "\x00" + schedule.WorkspaceID + "\x00" + schedule.Environment + "\x00" +
		schedule.ApplicationID + "\x00" + schedule.ScheduleID
}
