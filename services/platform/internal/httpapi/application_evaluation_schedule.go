package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	applicationEvaluationScheduleSchemaVersion           = "application_evaluation_schedule.v1"
	applicationEvaluationScheduleVersionSchemaVersion    = "application_evaluation_schedule_version.v1"
	applicationEvaluationScheduleOccurrenceSchemaVersion = "application_evaluation_schedule_occurrence.v1"

	applicationEvaluationScheduleStateDraft    = "draft"
	applicationEvaluationScheduleStateActive   = "active"
	applicationEvaluationScheduleStatePaused   = "paused"
	applicationEvaluationScheduleStateArchived = "archived"

	applicationEvaluationScheduleOccurrenceStateDue             = "due"
	applicationEvaluationScheduleOccurrenceStateClaimed         = "claimed"
	applicationEvaluationScheduleOccurrenceStateCampaignCreated = "campaign_created"
	applicationEvaluationScheduleOccurrenceStateObserving       = "observing"
	applicationEvaluationScheduleOccurrenceStateSucceeded       = "succeeded"
	applicationEvaluationScheduleOccurrenceStateFailed          = "failed"
	applicationEvaluationScheduleOccurrenceStateInterrupted     = "interrupted"
	applicationEvaluationScheduleOccurrenceStateSkipped         = "skipped"

	applicationEvaluationScheduleRuleDailyUTC                 = "daily_utc"
	applicationEvaluationScheduleMissedWindowPolicy           = "record_only_no_catch_up"
	applicationEvaluationScheduleOverlapPolicy                = "skip_while_campaign_non_terminal"
	applicationEvaluationScheduleAuthorizationModel           = "system_actor_schedule_scoped_delegation_v1"
	applicationEvaluationScheduleRevalidationPolicy           = "every_occurrence"
	applicationEvaluationScheduleAPIKeyOwnershipPolicy        = "delegated_user_current_owner"
	applicationEvaluationScheduleRevocationPolicy             = "fail_closed_immediate"
	applicationEvaluationSchedulePermissionEvaluationExecute  = "application_evaluations:execute"
	applicationEvaluationSchedulePermissionWorkflowRunExecute = "workflow_runs:execute"

	applicationEvaluationScheduleMemoryCapacity   = 200
	applicationEvaluationOccurrenceMemoryCapacity = 1000
)

const (
	ApplicationEvaluationScheduleFailureAuthorizationUnavailable = "application_evaluation_schedule_authorization_unavailable"
	ApplicationEvaluationScheduleFailureMembershipDenied         = "application_evaluation_schedule_membership_denied"
	ApplicationEvaluationScheduleFailurePlanChanged              = "application_evaluation_schedule_plan_changed"
	ApplicationEvaluationScheduleFailureAuthorityChanged         = "application_evaluation_schedule_authority_changed"
	ApplicationEvaluationScheduleFailureQuotaConsumerInvalid     = "application_evaluation_schedule_quota_consumer_invalid"
	ApplicationEvaluationScheduleFailureQuotaDenied              = "application_evaluation_schedule_quota_denied"
	ApplicationEvaluationScheduleFailureOverlapBlocked           = "application_evaluation_schedule_overlap_blocked"
	ApplicationEvaluationScheduleFailureMissedWindow             = "application_evaluation_schedule_missed_window"
	ApplicationEvaluationScheduleFailureClaimConflict            = "application_evaluation_schedule_claim_conflict"
	ApplicationEvaluationScheduleFailureCampaignFailed           = "application_evaluation_schedule_campaign_failed"
	ApplicationEvaluationScheduleFailureCampaignInterrupted      = "application_evaluation_schedule_campaign_interrupted"
	ApplicationEvaluationScheduleFailureStoreUnavailable         = "application_evaluation_schedule_store_unavailable"
	ApplicationEvaluationScheduleFailureStoreContract            = "application_evaluation_schedule_store_contract_mismatch"
	ApplicationEvaluationScheduleFailureNotFound                 = "application_evaluation_schedule_not_found"
	ApplicationEvaluationScheduleFailureVersionConflict          = "application_evaluation_schedule_version_conflict"
)

var (
	applicationEvaluationScheduleIDPattern = regexp.MustCompile(`^aesch_[a-z2-7]{16}$`)

	errApplicationEvaluationScheduleNotFound         = errors.New(ApplicationEvaluationScheduleFailureNotFound)
	errApplicationEvaluationScheduleVersionConflict  = errors.New(ApplicationEvaluationScheduleFailureVersionConflict)
	errApplicationEvaluationScheduleClaimConflict    = errors.New(ApplicationEvaluationScheduleFailureClaimConflict)
	errApplicationEvaluationScheduleStoreUnavailable = errors.New(ApplicationEvaluationScheduleFailureStoreUnavailable)
	errApplicationEvaluationScheduleStoreContract    = errors.New(ApplicationEvaluationScheduleFailureStoreContract)
)

var applicationEvaluationScheduleRequiredPermissions = []string{
	applicationEvaluationSchedulePermissionEvaluationExecute,
	applicationEvaluationSchedulePermissionWorkflowRunExecute,
}

type ApplicationEvaluationSchedule struct {
	SchemaVersion         string  `json:"schema_version"`
	ScheduleID            string  `json:"schedule_id"`
	RecordVersion         int     `json:"record_version"`
	LatestScheduleVersion int     `json:"latest_schedule_version"`
	LatestScheduleDigest  string  `json:"latest_schedule_digest"`
	TenantRef             string  `json:"tenant_ref"`
	WorkspaceID           string  `json:"workspace_id"`
	Environment           string  `json:"environment"`
	ApplicationID         string  `json:"application_id"`
	PlanID                string  `json:"plan_id"`
	PlanVersion           int     `json:"plan_version"`
	PlanDigest            string  `json:"plan_digest"`
	ExecutionProfile      string  `json:"execution_profile"`
	QuotaAPIKeyID         string  `json:"quota_api_key_id"`
	AuthorizationModel    string  `json:"authorization_model"`
	SystemActorRef        string  `json:"system_actor_ref"`
	DelegatedByUserRef    string  `json:"delegated_by_user_ref"`
	LifecycleState        string  `json:"lifecycle_state"`
	NextDueAt             *string `json:"next_due_at"`
	CreatedAt             string  `json:"created_at"`
	UpdatedAt             string  `json:"updated_at"`
	CreatedByActorRef     string  `json:"created_by_actor_ref"`
	UpdatedByActorRef     string  `json:"updated_by_actor_ref"`
	RequestID             string  `json:"request_id"`
	AuditRef              string  `json:"audit_ref"`
}

type ApplicationEvaluationScheduleDailyUTC struct {
	Rule   string `json:"rule"`
	Hour   int    `json:"hour"`
	Minute int    `json:"minute"`
}

type ApplicationEvaluationScheduleAuthorization struct {
	Model                 string   `json:"model"`
	SystemActorRef        string   `json:"system_actor_ref"`
	DelegatedByUserRef    string   `json:"delegated_by_user_ref"`
	RequiredPermissions   []string `json:"required_permissions"`
	RevalidationPolicy    string   `json:"revalidation_policy"`
	APIKeyOwnershipPolicy string   `json:"api_key_ownership_policy"`
	RevocationPolicy      string   `json:"revocation_policy"`
}

type ApplicationEvaluationScheduleVersion struct {
	SchemaVersion           string                                     `json:"schema_version"`
	ScheduleID              string                                     `json:"schedule_id"`
	ScheduleVersion         int                                        `json:"schedule_version"`
	PreviousScheduleVersion int                                        `json:"previous_schedule_version"`
	ScheduleDigest          string                                     `json:"schedule_digest"`
	TenantRef               string                                     `json:"tenant_ref"`
	WorkspaceID             string                                     `json:"workspace_id"`
	Environment             string                                     `json:"environment"`
	ApplicationID           string                                     `json:"application_id"`
	PlanID                  string                                     `json:"plan_id"`
	PlanVersion             int                                        `json:"plan_version"`
	PlanDigest              string                                     `json:"plan_digest"`
	ExecutionProfile        string                                     `json:"execution_profile"`
	QuotaAPIKeyID           string                                     `json:"quota_api_key_id"`
	Schedule                ApplicationEvaluationScheduleDailyUTC      `json:"schedule"`
	ItemCount               int                                        `json:"item_count"`
	MaxProviderAttempts     int                                        `json:"max_provider_attempts"`
	MissedWindowPolicy      string                                     `json:"missed_window_policy"`
	OverlapPolicy           string                                     `json:"overlap_policy"`
	Authorization           ApplicationEvaluationScheduleAuthorization `json:"authorization"`
	CreatedAt               string                                     `json:"created_at"`
	CreatedByActorRef       string                                     `json:"created_by_actor_ref"`
	RequestID               string                                     `json:"request_id"`
	AuditRef                string                                     `json:"audit_ref"`
}

type ApplicationEvaluationScheduleOccurrence struct {
	SchemaVersion      string  `json:"schema_version"`
	RecordVersion      int     `json:"record_version"`
	TenantRef          string  `json:"tenant_ref"`
	WorkspaceID        string  `json:"workspace_id"`
	Environment        string  `json:"environment"`
	ApplicationID      string  `json:"application_id"`
	ScheduleID         string  `json:"schedule_id"`
	ScheduleVersion    int     `json:"schedule_version"`
	ScheduleDigest     string  `json:"schedule_digest"`
	ScheduledForUTC    string  `json:"scheduled_for_utc"`
	State              string  `json:"state"`
	ClientCampaignKey  string  `json:"client_campaign_key"`
	CampaignID         *string `json:"campaign_id"`
	SystemActorRef     string  `json:"system_actor_ref"`
	DelegatedByUserRef string  `json:"delegated_by_user_ref"`
	ClaimedAt          *string `json:"claimed_at"`
	FailureCode        *string `json:"failure_code"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
	CompletedAt        *string `json:"completed_at"`
	RequestID          string  `json:"request_id"`
	AuditRef           string  `json:"audit_ref"`
}

type ApplicationEvaluationScheduleExecutionRef struct {
	AuthorizationModel string `json:"authorization_model"`
	ScheduleID         string `json:"schedule_id"`
	ScheduleVersion    int    `json:"schedule_version"`
	ScheduleDigest     string `json:"schedule_digest"`
	ScheduledForUTC    string `json:"scheduled_for_utc"`
	ClientCampaignKey  string `json:"client_campaign_key"`
	SystemActorRef     string `json:"system_actor_ref"`
	DelegatedByUserRef string `json:"delegated_by_user_ref"`
}

func validateApplicationEvaluationScheduleExecutionRef(ref ApplicationEvaluationScheduleExecutionRef) error {
	if ref.AuthorizationModel != applicationEvaluationScheduleAuthorizationModel ||
		!applicationEvaluationScheduleIDPattern.MatchString(ref.ScheduleID) || ref.ScheduleVersion < 1 ||
		!workflowRAGDigestPattern.MatchString(ref.ScheduleDigest) ||
		ref.ClientCampaignKey != applicationEvaluationScheduleClientCampaignKey(ref.ScheduleID, ref.ScheduleVersion, ref.ScheduledForUTC) ||
		!validApplicationEvaluationScheduleReference(ref.SystemActorRef) || !validApplicationEvaluationScheduleReference(ref.DelegatedByUserRef) {
		return errApplicationEvaluationScheduleStoreContract
	}
	if _, ok := parseApplicationEvaluationScheduleUTCTimestamp(ref.ScheduledForUTC); !ok {
		return errApplicationEvaluationScheduleStoreContract
	}
	return nil
}

func applicationEvaluationScheduleDigest(version ApplicationEvaluationScheduleVersion) (string, error) {
	payload, err := json.Marshal(struct {
		ScheduleID          string                                     `json:"schedule_id"`
		ScheduleVersion     int                                        `json:"schedule_version"`
		TenantRef           string                                     `json:"tenant_ref"`
		WorkspaceID         string                                     `json:"workspace_id"`
		Environment         string                                     `json:"environment"`
		ApplicationID       string                                     `json:"application_id"`
		PlanID              string                                     `json:"plan_id"`
		PlanVersion         int                                        `json:"plan_version"`
		PlanDigest          string                                     `json:"plan_digest"`
		ExecutionProfile    string                                     `json:"execution_profile"`
		QuotaAPIKeyID       string                                     `json:"quota_api_key_id"`
		Schedule            ApplicationEvaluationScheduleDailyUTC      `json:"schedule"`
		ItemCount           int                                        `json:"item_count"`
		MaxProviderAttempts int                                        `json:"max_provider_attempts"`
		MissedWindowPolicy  string                                     `json:"missed_window_policy"`
		OverlapPolicy       string                                     `json:"overlap_policy"`
		Authorization       ApplicationEvaluationScheduleAuthorization `json:"authorization"`
	}{
		ScheduleID: version.ScheduleID, ScheduleVersion: version.ScheduleVersion,
		TenantRef: version.TenantRef, WorkspaceID: version.WorkspaceID, Environment: version.Environment,
		ApplicationID: version.ApplicationID, PlanID: version.PlanID, PlanVersion: version.PlanVersion,
		PlanDigest: version.PlanDigest, ExecutionProfile: version.ExecutionProfile, QuotaAPIKeyID: version.QuotaAPIKeyID,
		Schedule: version.Schedule, ItemCount: version.ItemCount, MaxProviderAttempts: version.MaxProviderAttempts,
		MissedWindowPolicy: version.MissedWindowPolicy, OverlapPolicy: version.OverlapPolicy,
		Authorization: version.Authorization,
	})
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func applicationEvaluationScheduleNextDue(after time.Time, schedule ApplicationEvaluationScheduleDailyUTC) (time.Time, error) {
	if after.IsZero() || !validApplicationEvaluationScheduleDailyUTC(schedule) {
		return time.Time{}, errApplicationEvaluationScheduleStoreContract
	}
	after = after.UTC()
	next := time.Date(after.Year(), after.Month(), after.Day(), schedule.Hour, schedule.Minute, 0, 0, time.UTC)
	if !next.After(after) {
		next = next.AddDate(0, 0, 1)
	}
	return next, nil
}

func applicationEvaluationScheduleClientCampaignKey(scheduleID string, scheduleVersion int, scheduledForUTC string) string {
	payload := scheduleID + "\x00v" + strconv.Itoa(scheduleVersion) + "\x00" + scheduledForUTC
	digest := sha256.Sum256([]byte(payload))
	return "scheduled_campaign_" + hex.EncodeToString(digest[:12])
}

func validateApplicationEvaluationSchedule(ctx ApplicationEvaluationContext, schedule ApplicationEvaluationSchedule) error {
	createdAt, createdOK := parseApplicationEvaluationScheduleUTCTimestamp(schedule.CreatedAt)
	updatedAt, updatedOK := parseApplicationEvaluationScheduleUTCTimestamp(schedule.UpdatedAt)
	if !validApplicationEvaluationContext(ctx) || schedule.SchemaVersion != applicationEvaluationScheduleSchemaVersion ||
		!applicationEvaluationScheduleIDPattern.MatchString(schedule.ScheduleID) || schedule.RecordVersion < 1 || schedule.LatestScheduleVersion < 1 ||
		!workflowRAGDigestPattern.MatchString(schedule.LatestScheduleDigest) || schedule.TenantRef != ctx.TenantRef ||
		schedule.WorkspaceID != ctx.WorkspaceID || schedule.Environment != ctx.Environment || schedule.ApplicationID != ctx.ApplicationID ||
		!applicationEvaluationPlanIDPattern.MatchString(schedule.PlanID) || schedule.PlanVersion < 1 || !workflowRAGDigestPattern.MatchString(schedule.PlanDigest) ||
		schedule.ExecutionProfile != applicationInteractionProfilePrompt || !apiKeyIDPattern.MatchString(schedule.QuotaAPIKeyID) ||
		schedule.AuthorizationModel != applicationEvaluationScheduleAuthorizationModel || !validApplicationEvaluationScheduleReference(schedule.SystemActorRef) ||
		!validApplicationEvaluationScheduleReference(schedule.DelegatedByUserRef) || !validApplicationEvaluationScheduleState(schedule.LifecycleState) ||
		!createdOK || !updatedOK || updatedAt.Before(createdAt) || !validApplicationEvaluationScheduleReference(schedule.CreatedByActorRef) ||
		!validApplicationEvaluationScheduleReference(schedule.UpdatedByActorRef) || schedule.CreatedByActorRef != schedule.DelegatedByUserRef ||
		(schedule.UpdatedByActorRef != schedule.DelegatedByUserRef && schedule.UpdatedByActorRef != schedule.SystemActorRef) ||
		!validApplicationEvaluationScheduleReference(schedule.RequestID) || !validApplicationEvaluationScheduleReference(schedule.AuditRef) {
		return errApplicationEvaluationScheduleStoreContract
	}
	if schedule.LifecycleState == applicationEvaluationScheduleStateActive {
		nextDueAt, ok := parseApplicationEvaluationScheduleUTCTimestampPointer(schedule.NextDueAt)
		if !ok || !nextDueAt.After(updatedAt) {
			return errApplicationEvaluationScheduleStoreContract
		}
	} else if schedule.NextDueAt != nil {
		return errApplicationEvaluationScheduleStoreContract
	}
	return nil
}

func validateApplicationEvaluationScheduleVersion(ctx ApplicationEvaluationContext, version ApplicationEvaluationScheduleVersion) error {
	if !validApplicationEvaluationContext(ctx) || version.SchemaVersion != applicationEvaluationScheduleVersionSchemaVersion ||
		!applicationEvaluationScheduleIDPattern.MatchString(version.ScheduleID) || version.ScheduleVersion < 1 ||
		version.PreviousScheduleVersion != version.ScheduleVersion-1 || !workflowRAGDigestPattern.MatchString(version.ScheduleDigest) ||
		version.TenantRef != ctx.TenantRef || version.WorkspaceID != ctx.WorkspaceID || version.Environment != ctx.Environment ||
		version.ApplicationID != ctx.ApplicationID || !applicationEvaluationPlanIDPattern.MatchString(version.PlanID) || version.PlanVersion < 1 ||
		!workflowRAGDigestPattern.MatchString(version.PlanDigest) || version.ExecutionProfile != applicationInteractionProfilePrompt ||
		!apiKeyIDPattern.MatchString(version.QuotaAPIKeyID) || !validApplicationEvaluationScheduleDailyUTC(version.Schedule) ||
		version.ItemCount < 1 || version.ItemCount > applicationEvaluationMaximumItems || version.MaxProviderAttempts != version.ItemCount ||
		version.MissedWindowPolicy != applicationEvaluationScheduleMissedWindowPolicy || version.OverlapPolicy != applicationEvaluationScheduleOverlapPolicy ||
		validateApplicationEvaluationScheduleAuthorization(version.Authorization) != nil || version.CreatedByActorRef != version.Authorization.DelegatedByUserRef ||
		!validApplicationEvaluationScheduleReference(version.CreatedByActorRef) || !validApplicationEvaluationScheduleReference(version.RequestID) ||
		!validApplicationEvaluationScheduleReference(version.AuditRef) {
		return errApplicationEvaluationScheduleStoreContract
	}
	if _, ok := parseApplicationEvaluationScheduleUTCTimestamp(version.CreatedAt); !ok {
		return errApplicationEvaluationScheduleStoreContract
	}
	digest, err := applicationEvaluationScheduleDigest(version)
	if err != nil || digest != version.ScheduleDigest {
		return errApplicationEvaluationScheduleStoreContract
	}
	return nil
}

func validateApplicationEvaluationScheduleAuthorization(authorization ApplicationEvaluationScheduleAuthorization) error {
	if authorization.Model != applicationEvaluationScheduleAuthorizationModel ||
		!validApplicationEvaluationScheduleReference(authorization.SystemActorRef) ||
		!validApplicationEvaluationScheduleReference(authorization.DelegatedByUserRef) ||
		!reflect.DeepEqual(authorization.RequiredPermissions, applicationEvaluationScheduleRequiredPermissions) ||
		authorization.RevalidationPolicy != applicationEvaluationScheduleRevalidationPolicy ||
		authorization.APIKeyOwnershipPolicy != applicationEvaluationScheduleAPIKeyOwnershipPolicy ||
		authorization.RevocationPolicy != applicationEvaluationScheduleRevocationPolicy {
		return errApplicationEvaluationScheduleStoreContract
	}
	return nil
}

func validateApplicationEvaluationScheduleOccurrence(ctx ApplicationEvaluationContext, occurrence ApplicationEvaluationScheduleOccurrence) error {
	scheduledFor, scheduledOK := parseApplicationEvaluationScheduleUTCTimestamp(occurrence.ScheduledForUTC)
	createdAt, createdOK := parseApplicationEvaluationScheduleUTCTimestamp(occurrence.CreatedAt)
	updatedAt, updatedOK := parseApplicationEvaluationScheduleUTCTimestamp(occurrence.UpdatedAt)
	if !validApplicationEvaluationContext(ctx) || occurrence.SchemaVersion != applicationEvaluationScheduleOccurrenceSchemaVersion || occurrence.RecordVersion < 1 ||
		occurrence.TenantRef != ctx.TenantRef || occurrence.WorkspaceID != ctx.WorkspaceID || occurrence.Environment != ctx.Environment || occurrence.ApplicationID != ctx.ApplicationID ||
		!applicationEvaluationScheduleIDPattern.MatchString(occurrence.ScheduleID) || occurrence.ScheduleVersion < 1 ||
		!workflowRAGDigestPattern.MatchString(occurrence.ScheduleDigest) || !scheduledOK || !validApplicationEvaluationScheduleOccurrenceState(occurrence.State) ||
		occurrence.ClientCampaignKey != applicationEvaluationScheduleClientCampaignKey(occurrence.ScheduleID, occurrence.ScheduleVersion, occurrence.ScheduledForUTC) ||
		!validApplicationEvaluationScheduleReference(occurrence.SystemActorRef) || !validApplicationEvaluationScheduleReference(occurrence.DelegatedByUserRef) ||
		!createdOK || !updatedOK || createdAt.Before(scheduledFor) || updatedAt.Before(createdAt) ||
		!validApplicationEvaluationScheduleReference(occurrence.RequestID) || !validApplicationEvaluationScheduleReference(occurrence.AuditRef) {
		return errApplicationEvaluationScheduleStoreContract
	}
	if occurrence.CampaignID != nil && !applicationEvaluationCampaignIDPattern.MatchString(*occurrence.CampaignID) {
		return errApplicationEvaluationScheduleStoreContract
	}
	claimedAt, claimedOK := parseApplicationEvaluationScheduleUTCTimestampPointer(occurrence.ClaimedAt)
	completedAt, completedOK := parseApplicationEvaluationScheduleUTCTimestampPointer(occurrence.CompletedAt)
	if claimedOK && (claimedAt.Before(createdAt) || updatedAt.Before(claimedAt)) {
		return errApplicationEvaluationScheduleStoreContract
	}
	if completedOK && (!claimedOK || completedAt.Before(claimedAt) || completedAt != updatedAt) {
		return errApplicationEvaluationScheduleStoreContract
	}
	if occurrence.FailureCode != nil && !validApplicationEvaluationScheduleFailure(*occurrence.FailureCode) {
		return errApplicationEvaluationScheduleStoreContract
	}
	switch occurrence.State {
	case applicationEvaluationScheduleOccurrenceStateDue:
		if occurrence.RecordVersion != 1 || occurrence.ClaimedAt != nil || occurrence.CampaignID != nil || occurrence.CompletedAt != nil || occurrence.FailureCode != nil || occurrence.UpdatedAt != occurrence.CreatedAt {
			return errApplicationEvaluationScheduleStoreContract
		}
	case applicationEvaluationScheduleOccurrenceStateClaimed:
		if !claimedOK || occurrence.CampaignID != nil || occurrence.CompletedAt != nil || occurrence.FailureCode != nil {
			return errApplicationEvaluationScheduleStoreContract
		}
	case applicationEvaluationScheduleOccurrenceStateCampaignCreated, applicationEvaluationScheduleOccurrenceStateObserving:
		if !claimedOK || occurrence.CampaignID == nil || occurrence.CompletedAt != nil || occurrence.FailureCode != nil {
			return errApplicationEvaluationScheduleStoreContract
		}
	case applicationEvaluationScheduleOccurrenceStateSucceeded:
		if !claimedOK || occurrence.CampaignID == nil || !completedOK || occurrence.FailureCode != nil {
			return errApplicationEvaluationScheduleStoreContract
		}
	case applicationEvaluationScheduleOccurrenceStateFailed, applicationEvaluationScheduleOccurrenceStateInterrupted:
		if !claimedOK || !completedOK || occurrence.FailureCode == nil {
			return errApplicationEvaluationScheduleStoreContract
		}
	case applicationEvaluationScheduleOccurrenceStateSkipped:
		if !claimedOK || occurrence.CampaignID != nil || !completedOK || occurrence.FailureCode == nil ||
			(*occurrence.FailureCode != ApplicationEvaluationScheduleFailureMissedWindow && *occurrence.FailureCode != ApplicationEvaluationScheduleFailureOverlapBlocked) {
			return errApplicationEvaluationScheduleStoreContract
		}
	}
	return nil
}

func validApplicationEvaluationScheduleDailyUTC(schedule ApplicationEvaluationScheduleDailyUTC) bool {
	return schedule.Rule == applicationEvaluationScheduleRuleDailyUTC && schedule.Hour >= 0 && schedule.Hour <= 23 && schedule.Minute >= 0 && schedule.Minute <= 59
}

func validApplicationEvaluationScheduleState(value string) bool {
	switch value {
	case applicationEvaluationScheduleStateDraft, applicationEvaluationScheduleStateActive, applicationEvaluationScheduleStatePaused, applicationEvaluationScheduleStateArchived:
		return true
	default:
		return false
	}
}

func validApplicationEvaluationScheduleOccurrenceState(value string) bool {
	switch value {
	case applicationEvaluationScheduleOccurrenceStateDue, applicationEvaluationScheduleOccurrenceStateClaimed,
		applicationEvaluationScheduleOccurrenceStateCampaignCreated, applicationEvaluationScheduleOccurrenceStateObserving,
		applicationEvaluationScheduleOccurrenceStateSucceeded, applicationEvaluationScheduleOccurrenceStateFailed,
		applicationEvaluationScheduleOccurrenceStateInterrupted, applicationEvaluationScheduleOccurrenceStateSkipped:
		return true
	default:
		return false
	}
}

func validApplicationEvaluationScheduleFailure(value string) bool {
	switch value {
	case ApplicationEvaluationScheduleFailureAuthorizationUnavailable, ApplicationEvaluationScheduleFailureMembershipDenied,
		ApplicationEvaluationScheduleFailurePlanChanged, ApplicationEvaluationScheduleFailureAuthorityChanged,
		ApplicationEvaluationScheduleFailureQuotaConsumerInvalid, ApplicationEvaluationScheduleFailureQuotaDenied,
		ApplicationEvaluationScheduleFailureOverlapBlocked, ApplicationEvaluationScheduleFailureMissedWindow,
		ApplicationEvaluationScheduleFailureClaimConflict, ApplicationEvaluationScheduleFailureStoreUnavailable,
		ApplicationEvaluationScheduleFailureStoreContract, ApplicationEvaluationScheduleFailureCampaignFailed,
		ApplicationEvaluationScheduleFailureCampaignInterrupted:
		return true
	default:
		return false
	}
}

func validApplicationEvaluationScheduleReference(value string) bool {
	value = strings.TrimSpace(value)
	return workflowHTTPToolReferencePattern.MatchString(value) && !workflowRAGContainsForbiddenMaterial(value) && !applicationDraftStringContainsSecret(value)
}

func parseApplicationEvaluationScheduleUTCTimestamp(value string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil || value != parsed.UTC().Format(time.RFC3339Nano) {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func parseApplicationEvaluationScheduleUTCTimestampPointer(value *string) (time.Time, bool) {
	if value == nil {
		return time.Time{}, false
	}
	return parseApplicationEvaluationScheduleUTCTimestamp(*value)
}

func applicationEvaluationScheduleMatchesVersion(schedule ApplicationEvaluationSchedule, version ApplicationEvaluationScheduleVersion) bool {
	return schedule.ScheduleID == version.ScheduleID && schedule.LatestScheduleVersion == version.ScheduleVersion &&
		schedule.LatestScheduleDigest == version.ScheduleDigest && schedule.TenantRef == version.TenantRef &&
		schedule.WorkspaceID == version.WorkspaceID && schedule.Environment == version.Environment && schedule.ApplicationID == version.ApplicationID &&
		schedule.PlanID == version.PlanID && schedule.PlanVersion == version.PlanVersion && schedule.PlanDigest == version.PlanDigest &&
		schedule.ExecutionProfile == version.ExecutionProfile && schedule.QuotaAPIKeyID == version.QuotaAPIKeyID &&
		schedule.AuthorizationModel == version.Authorization.Model && schedule.SystemActorRef == version.Authorization.SystemActorRef &&
		schedule.DelegatedByUserRef == version.Authorization.DelegatedByUserRef
}

func validApplicationEvaluationScheduleLifecycleTransition(before, after string) bool {
	switch before {
	case applicationEvaluationScheduleStateDraft:
		return after == applicationEvaluationScheduleStateActive || after == applicationEvaluationScheduleStateArchived
	case applicationEvaluationScheduleStateActive:
		return after == applicationEvaluationScheduleStateActive || after == applicationEvaluationScheduleStatePaused || after == applicationEvaluationScheduleStateArchived
	case applicationEvaluationScheduleStatePaused:
		return after == applicationEvaluationScheduleStateActive || after == applicationEvaluationScheduleStateArchived
	default:
		return false
	}
}

func validApplicationEvaluationScheduleOccurrenceTransition(before, after string) bool {
	switch before {
	case applicationEvaluationScheduleOccurrenceStateDue:
		return after == applicationEvaluationScheduleOccurrenceStateClaimed
	case applicationEvaluationScheduleOccurrenceStateClaimed:
		return after == applicationEvaluationScheduleOccurrenceStateCampaignCreated || after == applicationEvaluationScheduleOccurrenceStateFailed ||
			after == applicationEvaluationScheduleOccurrenceStateInterrupted || after == applicationEvaluationScheduleOccurrenceStateSkipped
	case applicationEvaluationScheduleOccurrenceStateCampaignCreated:
		return after == applicationEvaluationScheduleOccurrenceStateObserving || after == applicationEvaluationScheduleOccurrenceStateFailed || after == applicationEvaluationScheduleOccurrenceStateInterrupted
	case applicationEvaluationScheduleOccurrenceStateObserving:
		return after == applicationEvaluationScheduleOccurrenceStateSucceeded || after == applicationEvaluationScheduleOccurrenceStateFailed || after == applicationEvaluationScheduleOccurrenceStateInterrupted
	default:
		return false
	}
}

func cloneApplicationEvaluationSchedule(value ApplicationEvaluationSchedule) ApplicationEvaluationSchedule {
	value.NextDueAt = cloneApplicationEvaluationScheduleString(value.NextDueAt)
	return value
}

func cloneApplicationEvaluationScheduleVersion(value ApplicationEvaluationScheduleVersion) ApplicationEvaluationScheduleVersion {
	value.Authorization.RequiredPermissions = append([]string(nil), value.Authorization.RequiredPermissions...)
	return value
}

func cloneApplicationEvaluationScheduleOccurrence(value ApplicationEvaluationScheduleOccurrence) ApplicationEvaluationScheduleOccurrence {
	value.CampaignID = cloneApplicationEvaluationScheduleString(value.CampaignID)
	value.ClaimedAt = cloneApplicationEvaluationScheduleString(value.ClaimedAt)
	value.FailureCode = cloneApplicationEvaluationScheduleString(value.FailureCode)
	value.CompletedAt = cloneApplicationEvaluationScheduleString(value.CompletedAt)
	return value
}

func cloneApplicationEvaluationScheduleExecutionRef(value *ApplicationEvaluationScheduleExecutionRef) *ApplicationEvaluationScheduleExecutionRef {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func sameApplicationEvaluationScheduleExecutionRef(left, right *ApplicationEvaluationScheduleExecutionRef) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func cloneApplicationEvaluationScheduleString(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
