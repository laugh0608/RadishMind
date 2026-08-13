package httpapi

import (
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/bridge"
)

const (
	gatewayProviderAttemptRecordSchemaVersion      = "gateway_provider_attempt_record.v1"
	gatewayProviderAttemptCostSummarySchemaVersion = "gateway_request_attempt_cost_summary.v1"
)

type GatewayProviderAttemptPhase string
type GatewayProviderAttemptStatus string

const (
	GatewayProviderAttemptPhaseNotStarted      GatewayProviderAttemptPhase = "not_started"
	GatewayProviderAttemptPhasePrimaryRunning  GatewayProviderAttemptPhase = "primary_running"
	GatewayProviderAttemptPhaseFallbackPending GatewayProviderAttemptPhase = "fallback_pending"
	GatewayProviderAttemptPhaseFallbackRunning GatewayProviderAttemptPhase = "fallback_running"
	GatewayProviderAttemptPhaseTerminalPending GatewayProviderAttemptPhase = "terminal_pending"
	GatewayProviderAttemptPhaseTerminal        GatewayProviderAttemptPhase = "terminal"
)

const (
	GatewayProviderAttemptStatusRunning        GatewayProviderAttemptStatus = "running"
	GatewayProviderAttemptStatusSucceeded      GatewayProviderAttemptStatus = "succeeded"
	GatewayProviderAttemptStatusFailed         GatewayProviderAttemptStatus = "failed"
	GatewayProviderAttemptStatusQuotaRejected  GatewayProviderAttemptStatus = "quota_rejected"
	GatewayProviderAttemptStatusOutcomeUnknown GatewayProviderAttemptStatus = "outcome_unknown"
)

const (
	GatewayProviderAttemptCostCoverageNone     = "none"
	GatewayProviderAttemptCostCoveragePartial  = "partial"
	GatewayProviderAttemptCostCoverageComplete = "complete"
)

type GatewayProviderAttemptRecord struct {
	SchemaVersion       string                         `json:"schema_version"`
	AttemptID           string                         `json:"attempt_id"`
	Ordinal             int                            `json:"ordinal"`
	Status              GatewayProviderAttemptStatus   `json:"status"`
	ConfiguredProfileID string                         `json:"configured_profile_id"`
	ProviderID          string                         `json:"provider_id"`
	RuntimeProfile      string                         `json:"runtime_profile"`
	SelectedModel       string                         `json:"selected_model"`
	UpstreamModel       string                         `json:"upstream_model"`
	RouteGeneration     int                            `json:"route_generation"`
	RouteSnapshotDigest string                         `json:"route_snapshot_digest"`
	InventoryDigest     string                         `json:"inventory_digest"`
	QuotaAdmissionID    string                         `json:"quota_admission_id,omitempty"`
	QuotaRejectionCode  string                         `json:"quota_rejection_code,omitempty"`
	StartedAt           string                         `json:"started_at"`
	CompletedAt         string                         `json:"completed_at,omitempty"`
	DurationMS          int64                          `json:"duration_ms"`
	Failure             *bridge.ProviderAttemptFailure `json:"failure,omitempty"`
	FailureBoundary     string                         `json:"failure_boundary,omitempty"`
	Usage               GatewayRequestUsage            `json:"usage"`
	CostEstimate        GatewayRequestCostEstimate     `json:"cost_estimate"`
}

type GatewayProviderAttemptCostSummary struct {
	SchemaVersion         string `json:"schema_version"`
	KnownCostMicros       int64  `json:"known_cost_micros"`
	Coverage              string `json:"coverage"`
	EstimatedAttemptCount int    `json:"estimated_attempt_count"`
	UnknownAttemptCount   int    `json:"unknown_attempt_count"`
}

type gatewayProviderAttemptHistoryService struct {
	store gatewayRequestStore
}

func newGatewayProviderAttemptHistoryService(store gatewayRequestStore) gatewayProviderAttemptHistoryService {
	return gatewayProviderAttemptHistoryService{store: store}
}

func newGatewayProviderAttemptHistoryRecord(
	base GatewayRequestRecord,
	plan GatewayProviderAttemptPlan,
) (GatewayRequestRecord, error) {
	if base.Status != GatewayRequestStatusStarted || base.RecordVersion != 0 ||
		!validGatewayProviderAttemptPlan(plan) || base.RequestID != plan.RootRequestID ||
		base.Route != plan.Route || base.Protocol != plan.Protocol || len(plan.Targets) < 1 {
		return GatewayRequestRecord{}, errGatewayRequestStoreContract
	}
	base.SchemaVersion = gatewayRequestRecordSchemaVersionV3
	base.SelectionSource = gatewayProviderRouteSelectionSource
	base.SelectedProvider = plan.Targets[0].ProviderID
	base.SelectedProfile = plan.Targets[0].RuntimeProfile
	base.SelectedModel = plan.Targets[0].SelectedModel
	base.ProviderRouteConfigurationID = plan.ConfigurationID
	base.ProviderRouteGeneration = plan.RouteGeneration
	base.ProviderRouteSnapshotDigest = plan.RouteSnapshotDigest
	planCopy := cloneGatewayProviderAttemptPlan(plan)
	base.ProviderAttemptPlan = &planCopy
	base.CostEstimate = gatewayRequestCostUnavailable(GatewayRequestCostNotApplicable, "provider_not_attempted")
	base.ProviderAttemptPhase = GatewayProviderAttemptPhaseNotStarted
	base.ProviderAttemptCount = 0
	base.FallbackAllowed = plan.FallbackAllowed
	base.FallbackUsed = false
	base.TerminalAttemptID = ""
	base.ProviderAttempts = nil
	base.ProviderAttemptCostSummary = gatewayProviderAttemptCostSummary(nil)
	if !validGatewayProviderAttemptHistoryRecord(base) {
		return GatewayRequestRecord{}, errGatewayRequestStoreContract
	}
	return base, nil
}

func (service gatewayProviderAttemptHistoryService) StartAttempt(
	requestContext GatewayRequestContext,
	requestID string,
	target GatewayProviderAttemptPlanTarget,
	quotaAdmissionID string,
	now time.Time,
) (GatewayRequestRecord, error) {
	record, found, err := service.store.ReadRequest(requestContext, strings.TrimSpace(requestID))
	if err != nil || !found {
		return GatewayRequestRecord{}, errGatewayRequestStoreUnavailable
	}
	if record.SchemaVersion != gatewayRequestRecordSchemaVersionV3 || record.ProviderAttemptPlan == nil ||
		(target.Ordinal == 1 && record.ProviderAttemptPhase != GatewayProviderAttemptPhaseNotStarted) ||
		(target.Ordinal == 2 && record.ProviderAttemptPhase != GatewayProviderAttemptPhaseFallbackPending) ||
		target.Ordinal < 1 || target.Ordinal > len(record.ProviderAttemptPlan.Targets) ||
		!reflect.DeepEqual(target, record.ProviderAttemptPlan.Targets[target.Ordinal-1]) ||
		!validGatewayRequestReference(strings.TrimSpace(quotaAdmissionID), 160) {
		return GatewayRequestRecord{}, errGatewayRequestStoreContract
	}
	startedAt := now.UTC().Format(time.RFC3339Nano)
	attempt := GatewayProviderAttemptRecord{
		SchemaVersion: gatewayProviderAttemptRecordSchemaVersion,
		AttemptID:     target.AttemptID, Ordinal: target.Ordinal, Status: GatewayProviderAttemptStatusRunning,
		ConfiguredProfileID: target.ProviderProfileID, ProviderID: target.ProviderID,
		RuntimeProfile: target.RuntimeProfile, SelectedModel: target.SelectedModel, UpstreamModel: target.UpstreamModel,
		RouteGeneration: record.ProviderRouteGeneration, RouteSnapshotDigest: record.ProviderRouteSnapshotDigest,
		InventoryDigest: target.InventoryDigest, QuotaAdmissionID: strings.TrimSpace(quotaAdmissionID),
		StartedAt: startedAt, Usage: GatewayRequestUsage{Availability: GatewayRequestUsageNotReported},
		CostEstimate: gatewayRequestCostUnavailable(GatewayRequestCostUsageNotReported, "provider_usage_not_reported"),
	}
	record.ProviderAttempts = append(record.ProviderAttempts, attempt)
	record.ProviderAttemptCount = len(record.ProviderAttempts)
	if target.Ordinal == 1 {
		record.ProviderAttemptPhase = GatewayProviderAttemptPhasePrimaryRunning
	} else {
		record.ProviderAttemptPhase = GatewayProviderAttemptPhaseFallbackRunning
		record.FallbackUsed = true
	}
	record.ProviderAttemptCostSummary = gatewayProviderAttemptCostSummary(record.ProviderAttempts)
	if err := service.store.UpdateRequest(requestContext, &record); err != nil {
		return GatewayRequestRecord{}, err
	}
	return cloneGatewayRequestRecord(record), nil
}

func (service gatewayProviderAttemptHistoryService) CompleteAttempt(
	requestContext GatewayRequestContext,
	requestID string,
	attemptID string,
	usage GatewayRequestUsage,
	cost GatewayRequestCostEstimate,
	failure *bridge.ProviderAttemptFailure,
	prepareFallback bool,
	now time.Time,
) (GatewayRequestRecord, error) {
	record, found, err := service.store.ReadRequest(requestContext, strings.TrimSpace(requestID))
	if err != nil || !found {
		return GatewayRequestRecord{}, errGatewayRequestStoreUnavailable
	}
	if record.SchemaVersion != gatewayRequestRecordSchemaVersionV3 || len(record.ProviderAttempts) < 1 {
		return GatewayRequestRecord{}, errGatewayRequestStoreContract
	}
	index := len(record.ProviderAttempts) - 1
	attempt := record.ProviderAttempts[index]
	if attempt.AttemptID != strings.TrimSpace(attemptID) || attempt.Status != GatewayProviderAttemptStatusRunning ||
		!validGatewayRequestUsage(usage) || !validGatewayRequestCostEstimate(cost) {
		return GatewayRequestRecord{}, errGatewayRequestStoreContract
	}
	completedAt := now.UTC()
	startedAt, parseErr := time.Parse(time.RFC3339Nano, attempt.StartedAt)
	if parseErr != nil || completedAt.Before(startedAt) {
		return GatewayRequestRecord{}, errGatewayRequestStoreContract
	}
	attempt.CompletedAt = completedAt.Format(time.RFC3339Nano)
	attempt.DurationMS = completedAt.Sub(startedAt).Milliseconds()
	attempt.Usage = usage
	attempt.CostEstimate = cloneGatewayRequestCostEstimate(cost)
	if failure == nil {
		if prepareFallback {
			return GatewayRequestRecord{}, errGatewayRequestStoreContract
		}
		attempt.Status = GatewayProviderAttemptStatusSucceeded
	} else {
		failureCopy := *failure
		if !bridge.ValidProviderAttemptFailure(failureCopy) {
			return GatewayRequestRecord{}, errGatewayRequestStoreContract
		}
		attempt.Failure = &failureCopy
		attempt.FailureBoundary = errorBoundarySouthboundProvider
		if failureCopy.Outcome == bridge.ProviderAttemptUnknown {
			attempt.Status = GatewayProviderAttemptStatusOutcomeUnknown
		} else {
			attempt.Status = GatewayProviderAttemptStatusFailed
		}
		if prepareFallback && (!record.FallbackAllowed || attempt.Ordinal != 1 || !bridge.ProviderAttemptFailureEligible(failureCopy)) {
			return GatewayRequestRecord{}, errGatewayRequestStoreContract
		}
	}
	record.ProviderAttempts[index] = attempt
	if prepareFallback {
		record.ProviderAttemptPhase = GatewayProviderAttemptPhaseFallbackPending
	} else {
		record.ProviderAttemptPhase = GatewayProviderAttemptPhaseTerminalPending
	}
	record.ProviderAttemptCostSummary = gatewayProviderAttemptCostSummary(record.ProviderAttempts)
	if err := service.store.UpdateRequest(requestContext, &record); err != nil {
		return GatewayRequestRecord{}, err
	}
	return cloneGatewayRequestRecord(record), nil
}

func (service gatewayProviderAttemptHistoryService) RejectFallbackQuota(
	requestContext GatewayRequestContext,
	requestID string,
	target GatewayProviderAttemptPlanTarget,
	quotaRejectionCode string,
	now time.Time,
) (GatewayRequestRecord, error) {
	record, found, err := service.store.ReadRequest(requestContext, strings.TrimSpace(requestID))
	if err != nil || !found {
		return GatewayRequestRecord{}, errGatewayRequestStoreUnavailable
	}
	quotaRejectionCode = strings.TrimSpace(quotaRejectionCode)
	if record.SchemaVersion != gatewayRequestRecordSchemaVersionV3 || record.ProviderAttemptPlan == nil ||
		record.ProviderAttemptPhase != GatewayProviderAttemptPhaseFallbackPending || target.Ordinal != 2 ||
		len(record.ProviderAttemptPlan.Targets) != 2 ||
		!reflect.DeepEqual(target, record.ProviderAttemptPlan.Targets[1]) ||
		!validGatewayRequestReference(quotaRejectionCode, 160) {
		return GatewayRequestRecord{}, errGatewayRequestStoreContract
	}
	timestamp := now.UTC().Format(time.RFC3339Nano)
	record.ProviderAttempts = append(record.ProviderAttempts, GatewayProviderAttemptRecord{
		SchemaVersion: gatewayProviderAttemptRecordSchemaVersion,
		AttemptID:     target.AttemptID, Ordinal: 2, Status: GatewayProviderAttemptStatusQuotaRejected,
		ConfiguredProfileID: target.ProviderProfileID, ProviderID: target.ProviderID,
		RuntimeProfile: target.RuntimeProfile, SelectedModel: target.SelectedModel, UpstreamModel: target.UpstreamModel,
		RouteGeneration: record.ProviderRouteGeneration, RouteSnapshotDigest: record.ProviderRouteSnapshotDigest,
		InventoryDigest: target.InventoryDigest, QuotaRejectionCode: quotaRejectionCode,
		StartedAt: timestamp, CompletedAt: timestamp,
		FailureBoundary: errorBoundaryQuotaAdmission,
		Usage:           GatewayRequestUsage{Availability: GatewayRequestUsageNotApplicable},
		CostEstimate:    gatewayRequestCostUnavailable(GatewayRequestCostNotApplicable, "provider_not_attempted"),
	})
	record.ProviderAttemptCount = len(record.ProviderAttempts)
	record.ProviderAttemptPhase = GatewayProviderAttemptPhaseTerminalPending
	record.ProviderAttemptCostSummary = gatewayProviderAttemptCostSummary(record.ProviderAttempts)
	if err := service.store.UpdateRequest(requestContext, &record); err != nil {
		return GatewayRequestRecord{}, err
	}
	return cloneGatewayRequestRecord(record), nil
}

func (service gatewayProviderAttemptHistoryService) Finalize(
	requestContext GatewayRequestContext,
	requestID string,
	status GatewayRequestStatus,
	httpStatusCode int,
	failureCode string,
	failureBoundary string,
	now time.Time,
) (GatewayRequestRecord, error) {
	record, found, err := service.store.ReadRequest(requestContext, strings.TrimSpace(requestID))
	if err != nil || !found {
		return GatewayRequestRecord{}, errGatewayRequestStoreUnavailable
	}
	if record.SchemaVersion != gatewayRequestRecordSchemaVersionV3 ||
		record.ProviderAttemptPhase != GatewayProviderAttemptPhaseTerminalPending || len(record.ProviderAttempts) < 1 ||
		!isTerminalGatewayRequestStatus(status) {
		return GatewayRequestRecord{}, errGatewayRequestStoreContract
	}
	terminalSucceeded := record.ProviderAttempts[len(record.ProviderAttempts)-1].Status == GatewayProviderAttemptStatusSucceeded
	if (status == GatewayRequestStatusSucceeded) != terminalSucceeded {
		return GatewayRequestRecord{}, errGatewayRequestStoreContract
	}
	completedAt := now.UTC()
	startedAt, parseErr := time.Parse(time.RFC3339Nano, record.StartedAt)
	if parseErr != nil || completedAt.Before(startedAt) {
		return GatewayRequestRecord{}, errGatewayRequestStoreContract
	}
	record.Status = status
	record.ProviderAttemptPhase = GatewayProviderAttemptPhaseTerminal
	record.TerminalAttemptID = record.ProviderAttempts[len(record.ProviderAttempts)-1].AttemptID
	record.CompletedAt = completedAt.Format(time.RFC3339Nano)
	record.DurationMS = completedAt.Sub(startedAt).Milliseconds()
	record.HTTPStatusCode = httpStatusCode
	record.FailureCode = strings.TrimSpace(failureCode)
	record.FailureBoundary = strings.TrimSpace(failureBoundary)
	record.CostEstimate = cloneGatewayRequestCostEstimate(record.ProviderAttempts[0].CostEstimate)
	if err := service.store.UpdateRequest(requestContext, &record); err != nil {
		return GatewayRequestRecord{}, err
	}
	return cloneGatewayRequestRecord(record), nil
}

func validGatewayProviderAttemptHistoryRecord(record GatewayRequestRecord) bool {
	if record.SchemaVersion != gatewayRequestRecordSchemaVersionV3 {
		return record.ProviderAttemptPlan == nil && record.ProviderAttemptPhase == "" && record.ProviderAttemptCount == 0 && !record.FallbackAllowed &&
			!record.FallbackUsed && record.TerminalAttemptID == "" && len(record.ProviderAttempts) == 0 &&
			reflect.DeepEqual(record.ProviderAttemptCostSummary, GatewayProviderAttemptCostSummary{})
	}
	if record.ProviderAttemptPlan == nil || !validGatewayProviderAttemptPlan(*record.ProviderAttemptPlan) ||
		record.ProviderAttemptPlan.RootRequestID != record.RequestID ||
		record.ProviderAttemptPlan.Route != record.Route || record.ProviderAttemptPlan.Protocol != record.Protocol ||
		record.ProviderAttemptPlan.ConfigurationID != record.ProviderRouteConfigurationID ||
		record.ProviderAttemptPlan.RouteGeneration != record.ProviderRouteGeneration ||
		record.ProviderAttemptPlan.RouteSnapshotDigest != record.ProviderRouteSnapshotDigest ||
		record.ProviderAttemptPlan.FallbackAllowed != record.FallbackAllowed ||
		record.SelectionSource != gatewayProviderRouteSelectionSource ||
		record.SelectedProvider != record.ProviderAttemptPlan.Targets[0].ProviderID ||
		record.SelectedProfile != record.ProviderAttemptPlan.Targets[0].RuntimeProfile ||
		record.SelectedModel != record.ProviderAttemptPlan.Targets[0].SelectedModel ||
		record.ProviderAttemptCount != len(record.ProviderAttempts) || len(record.ProviderAttempts) > 2 ||
		!validGatewayProviderAttemptCostSummary(record.ProviderAttemptCostSummary, len(record.ProviderAttempts)) ||
		!reflect.DeepEqual(record.ProviderAttemptCostSummary, gatewayProviderAttemptCostSummary(record.ProviderAttempts)) ||
		record.FallbackUsed && (len(record.ProviderAttempts) != 2 || !record.FallbackAllowed) {
		return false
	}
	for index, attempt := range record.ProviderAttempts {
		if !validGatewayProviderAttemptRecord(record, attempt, index+1) ||
			index >= len(record.ProviderAttemptPlan.Targets) ||
			!gatewayProviderAttemptRecordMatchesTarget(attempt, record.ProviderAttemptPlan.Targets[index]) {
			return false
		}
	}
	if record.Status == GatewayRequestStatusStarted {
		if record.TerminalAttemptID != "" {
			return false
		}
		switch record.ProviderAttemptPhase {
		case GatewayProviderAttemptPhaseNotStarted:
			return len(record.ProviderAttempts) == 0
		case GatewayProviderAttemptPhasePrimaryRunning:
			return len(record.ProviderAttempts) == 1 && record.ProviderAttempts[0].Status == GatewayProviderAttemptStatusRunning
		case GatewayProviderAttemptPhaseFallbackPending:
			return record.FallbackAllowed && len(record.ProviderAttempts) == 1 &&
				record.ProviderAttempts[0].Status == GatewayProviderAttemptStatusFailed &&
				record.ProviderAttempts[0].Failure != nil && bridge.ProviderAttemptFailureEligible(*record.ProviderAttempts[0].Failure)
		case GatewayProviderAttemptPhaseFallbackRunning:
			return record.FallbackAllowed && record.FallbackUsed && len(record.ProviderAttempts) == 2 &&
				record.ProviderAttempts[1].Status == GatewayProviderAttemptStatusRunning
		case GatewayProviderAttemptPhaseTerminalPending:
			return len(record.ProviderAttempts) >= 1 && isTerminalGatewayProviderAttemptStatus(record.ProviderAttempts[len(record.ProviderAttempts)-1].Status)
		default:
			return false
		}
	}
	if len(record.ProviderAttempts) == 0 || record.ProviderAttemptPhase != GatewayProviderAttemptPhaseTerminal ||
		record.TerminalAttemptID != record.ProviderAttempts[len(record.ProviderAttempts)-1].AttemptID {
		return false
	}
	terminalSucceeded := record.ProviderAttempts[len(record.ProviderAttempts)-1].Status == GatewayProviderAttemptStatusSucceeded
	return (record.Status == GatewayRequestStatusSucceeded) == terminalSucceeded
}

func gatewayProviderAttemptRecordMatchesTarget(
	attempt GatewayProviderAttemptRecord,
	target GatewayProviderAttemptPlanTarget,
) bool {
	return attempt.AttemptID == target.AttemptID && attempt.Ordinal == target.Ordinal &&
		attempt.ConfiguredProfileID == target.ProviderProfileID && attempt.ProviderID == target.ProviderID &&
		attempt.RuntimeProfile == target.RuntimeProfile && attempt.SelectedModel == target.SelectedModel &&
		attempt.UpstreamModel == target.UpstreamModel && attempt.InventoryDigest == target.InventoryDigest
}

func validGatewayProviderAttemptRecord(root GatewayRequestRecord, attempt GatewayProviderAttemptRecord, ordinal int) bool {
	if attempt.SchemaVersion != gatewayProviderAttemptRecordSchemaVersion || attempt.Ordinal != ordinal ||
		attempt.AttemptID != root.RequestID+".pa"+strconv.Itoa(ordinal) ||
		!adminProviderRouteIdentifierPattern.MatchString(attempt.ConfiguredProfileID) ||
		!adminProviderRouteIdentifierPattern.MatchString(attempt.ProviderID) ||
		!validGatewayRequestReference(attempt.RuntimeProfile, 256) ||
		!adminProviderRouteModelPattern.MatchString(attempt.SelectedModel) ||
		!adminProviderRouteModelPattern.MatchString(attempt.UpstreamModel) ||
		attempt.RouteGeneration != root.ProviderRouteGeneration ||
		attempt.RouteSnapshotDigest != root.ProviderRouteSnapshotDigest ||
		!adminProviderRouteDigestPattern.MatchString(attempt.InventoryDigest) ||
		attempt.DurationMS < 0 || !validGatewayRequestUsage(attempt.Usage) || !validGatewayRequestCostEstimate(attempt.CostEstimate) {
		return false
	}
	startedAt, err := time.Parse(time.RFC3339Nano, attempt.StartedAt)
	if err != nil || startedAt.IsZero() {
		return false
	}
	if attempt.Status == GatewayProviderAttemptStatusRunning {
		return validGatewayRequestReference(attempt.QuotaAdmissionID, 160) && attempt.QuotaRejectionCode == "" &&
			attempt.CompletedAt == "" && attempt.Failure == nil && attempt.FailureBoundary == ""
	}
	completedAt, err := time.Parse(time.RFC3339Nano, attempt.CompletedAt)
	if err != nil || completedAt.Before(startedAt) {
		return false
	}
	switch attempt.Status {
	case GatewayProviderAttemptStatusSucceeded:
		return validGatewayRequestReference(attempt.QuotaAdmissionID, 160) && attempt.QuotaRejectionCode == "" &&
			attempt.Failure == nil && attempt.FailureBoundary == ""
	case GatewayProviderAttemptStatusFailed, GatewayProviderAttemptStatusOutcomeUnknown:
		if !validGatewayRequestReference(attempt.QuotaAdmissionID, 160) || attempt.QuotaRejectionCode != "" ||
			attempt.Failure == nil || !bridge.ValidProviderAttemptFailure(*attempt.Failure) ||
			attempt.FailureBoundary != errorBoundarySouthboundProvider {
			return false
		}
		return (attempt.Status == GatewayProviderAttemptStatusOutcomeUnknown) ==
			(attempt.Failure.Outcome == bridge.ProviderAttemptUnknown)
	case GatewayProviderAttemptStatusQuotaRejected:
		return attempt.QuotaAdmissionID == "" && validGatewayRequestReference(attempt.QuotaRejectionCode, 160) &&
			attempt.Failure == nil && attempt.FailureBoundary == errorBoundaryQuotaAdmission && attempt.Ordinal == 2
	default:
		return false
	}
}

func validGatewayProviderAttemptRecordTransition(current, next GatewayRequestRecord) bool {
	if current.SchemaVersion != gatewayRequestRecordSchemaVersionV3 || next.SchemaVersion != gatewayRequestRecordSchemaVersionV3 {
		return current.SchemaVersion == next.SchemaVersion
	}
	if !reflect.DeepEqual(current.ProviderAttemptPlan, next.ProviderAttemptPlan) {
		return false
	}
	if len(next.ProviderAttempts) < len(current.ProviderAttempts) || len(next.ProviderAttempts) > len(current.ProviderAttempts)+1 {
		return false
	}
	for index := range current.ProviderAttempts {
		if current.ProviderAttempts[index].Status != GatewayProviderAttemptStatusRunning &&
			!reflect.DeepEqual(current.ProviderAttempts[index], next.ProviderAttempts[index]) {
			return false
		}
	}
	allowed := map[GatewayProviderAttemptPhase][]GatewayProviderAttemptPhase{
		GatewayProviderAttemptPhaseNotStarted:      {GatewayProviderAttemptPhasePrimaryRunning},
		GatewayProviderAttemptPhasePrimaryRunning:  {GatewayProviderAttemptPhaseFallbackPending, GatewayProviderAttemptPhaseTerminalPending},
		GatewayProviderAttemptPhaseFallbackPending: {GatewayProviderAttemptPhaseFallbackRunning, GatewayProviderAttemptPhaseTerminalPending},
		GatewayProviderAttemptPhaseFallbackRunning: {GatewayProviderAttemptPhaseTerminalPending},
		GatewayProviderAttemptPhaseTerminalPending: {GatewayProviderAttemptPhaseTerminal},
	}
	for _, phase := range allowed[current.ProviderAttemptPhase] {
		if next.ProviderAttemptPhase == phase {
			return true
		}
	}
	return false
}

func isTerminalGatewayProviderAttemptStatus(status GatewayProviderAttemptStatus) bool {
	return status == GatewayProviderAttemptStatusSucceeded || status == GatewayProviderAttemptStatusFailed ||
		status == GatewayProviderAttemptStatusQuotaRejected || status == GatewayProviderAttemptStatusOutcomeUnknown
}

func gatewayProviderAttemptCostSummary(attempts []GatewayProviderAttemptRecord) GatewayProviderAttemptCostSummary {
	summary := GatewayProviderAttemptCostSummary{
		SchemaVersion: gatewayProviderAttemptCostSummarySchemaVersion,
		Coverage:      GatewayProviderAttemptCostCoverageNone,
	}
	for _, attempt := range attempts {
		if attempt.CostEstimate.Availability == GatewayRequestCostEstimated && attempt.CostEstimate.EstimatedCostMicros != nil {
			if *attempt.CostEstimate.EstimatedCostMicros > math.MaxInt64-summary.KnownCostMicros {
				return GatewayProviderAttemptCostSummary{}
			}
			summary.KnownCostMicros += *attempt.CostEstimate.EstimatedCostMicros
			summary.EstimatedAttemptCount++
		} else {
			summary.UnknownAttemptCount++
		}
	}
	if len(attempts) > 0 && summary.EstimatedAttemptCount == len(attempts) {
		summary.Coverage = GatewayProviderAttemptCostCoverageComplete
	} else if summary.EstimatedAttemptCount > 0 {
		summary.Coverage = GatewayProviderAttemptCostCoveragePartial
	}
	return summary
}

func validGatewayProviderAttemptCostSummary(summary GatewayProviderAttemptCostSummary, attemptCount int) bool {
	if summary.SchemaVersion != gatewayProviderAttemptCostSummarySchemaVersion || summary.KnownCostMicros < 0 ||
		summary.EstimatedAttemptCount < 0 || summary.UnknownAttemptCount < 0 ||
		summary.EstimatedAttemptCount+summary.UnknownAttemptCount != attemptCount {
		return false
	}
	switch summary.Coverage {
	case GatewayProviderAttemptCostCoverageNone:
		return summary.EstimatedAttemptCount == 0 && summary.KnownCostMicros == 0
	case GatewayProviderAttemptCostCoveragePartial:
		return summary.EstimatedAttemptCount > 0 && summary.UnknownAttemptCount > 0
	case GatewayProviderAttemptCostCoverageComplete:
		return attemptCount > 0 && summary.EstimatedAttemptCount == attemptCount && summary.UnknownAttemptCount == 0
	default:
		return false
	}
}

func cloneGatewayProviderAttemptRecords(values []GatewayProviderAttemptRecord) []GatewayProviderAttemptRecord {
	cloned := append([]GatewayProviderAttemptRecord(nil), values...)
	for index := range cloned {
		cloned[index].CostEstimate = cloneGatewayRequestCostEstimate(cloned[index].CostEstimate)
		if cloned[index].Failure != nil {
			failure := *cloned[index].Failure
			cloned[index].Failure = &failure
		}
	}
	return cloned
}
