package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

func adminProviderRouteDatabaseContext(ctx AdminProviderRouteContext) context.Context {
	if ctx.RequestContext != nil {
		return ctx.RequestContext
	}
	return context.Background()
}

func validAdminProviderRouteStorageContext(ctx AdminProviderRouteContext) bool {
	return strings.TrimSpace(ctx.TenantRef) != "" &&
		strings.TrimSpace(ctx.WorkspaceID) != "" &&
		(ctx.Environment == adminProviderRouteEnvironmentDevelopment ||
			ctx.Environment == adminProviderRouteEnvironmentTest)
}

func encodeAdminProviderRouteStored(value any) ([]byte, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, errAdminProviderRouteStoreUnavailable
	}
	return payload, nil
}

func decodeAdminProviderRouteDraft(
	payload []byte,
	ctx AdminProviderRouteContext,
	configurationID string,
	revision int,
	digest string,
) (AdminProviderRouteConfigurationDraft, error) {
	var value AdminProviderRouteConfigurationDraft
	if decodeStrictStoredJSON(payload, &value) != nil ||
		value.ConfigurationID != configurationID ||
		value.DraftRevision != revision ||
		value.DraftDigest != digest ||
		!validAdminProviderRouteDraftIntegrity(ctx, value) {
		return AdminProviderRouteConfigurationDraft{}, errAdminProviderRouteStoreUnavailable
	}
	return value, nil
}

func decodeAdminProviderRouteCandidate(
	payload []byte,
	ctx AdminProviderRouteContext,
	configurationID string,
	candidateID string,
	sourceDraftRevision int,
	sourceDraftDigest string,
	candidateDigest string,
	candidateState string,
	reviewVersion int,
) (AdminProviderRouteCandidate, error) {
	var value AdminProviderRouteCandidate
	if decodeStrictStoredJSON(payload, &value) != nil ||
		value.TenantRef != ctx.TenantRef ||
		value.WorkspaceID != ctx.WorkspaceID ||
		value.Environment != ctx.Environment ||
		value.ConfigurationID != configurationID ||
		value.CandidateID != candidateID ||
		value.SourceDraftRevision != sourceDraftRevision ||
		value.SourceDraftDigest != sourceDraftDigest ||
		value.CandidateDigest != candidateDigest ||
		value.CandidateState != candidateState ||
		value.ReviewVersion != reviewVersion ||
		!validAdminProviderRouteCandidateIntegrity(value) {
		return AdminProviderRouteCandidate{}, errAdminProviderRouteStoreUnavailable
	}
	return value, nil
}

func decodeAdminProviderRouteReview(
	payload []byte,
	reviewVersion int,
	decision string,
	resultingState string,
) (AdminProviderRouteReview, error) {
	var value AdminProviderRouteReview
	if decodeStrictStoredJSON(payload, &value) != nil ||
		value.SchemaVersion != adminProviderRouteReviewSchemaVersion ||
		value.ReviewVersion != reviewVersion ||
		value.Decision != decision ||
		value.ResultingState != resultingState ||
		reviewVersion != 1 ||
		!validAdminProviderRouteReason(value.Reason) ||
		(decision == adminProviderRouteDecisionApprove && resultingState != adminProviderRouteCandidateApproved) ||
		(decision == adminProviderRouteDecisionReject && resultingState != adminProviderRouteCandidateRejected) {
		return AdminProviderRouteReview{}, errAdminProviderRouteStoreUnavailable
	}
	return value, nil
}

func decodeAdminProviderRouteSnapshot(
	payload []byte,
	ctx AdminProviderRouteContext,
	configurationID string,
	generation int,
	candidateID string,
	candidateDigest string,
	snapshotDigest string,
) (AdminProviderRouteSnapshot, error) {
	var value AdminProviderRouteSnapshot
	if decodeStrictStoredJSON(payload, &value) != nil ||
		value.ConfigurationID != configurationID ||
		value.Generation != generation ||
		value.CandidateID != candidateID ||
		value.CandidateDigest != candidateDigest ||
		value.SnapshotDigest != snapshotDigest ||
		!validAdminProviderRouteSnapshotIntegrity(ctx, value) {
		return AdminProviderRouteSnapshot{}, errAdminProviderRouteStoreUnavailable
	}
	return value, nil
}

func decodeAdminProviderRouteActivation(
	payload []byte,
	configurationID string,
	afterGeneration int,
	activationID string,
	action string,
	afterCandidateID string,
	afterSnapshotDigest string,
	previousRecordDigest string,
	recordDigest string,
) (AdminProviderRouteActivationRecord, error) {
	var value AdminProviderRouteActivationRecord
	if decodeStrictStoredJSON(payload, &value) != nil ||
		value.SchemaVersion != adminProviderRouteActivationSchemaVersion ||
		value.ConfigurationID != configurationID ||
		value.AfterGeneration != afterGeneration ||
		value.ActivationID != activationID ||
		value.Action != action ||
		value.AfterCandidateID != afterCandidateID ||
		value.AfterSnapshotDigest != afterSnapshotDigest ||
		value.PreviousRecordDigest != previousRecordDigest ||
		value.RecordDigest != recordDigest ||
		value.BeforeGeneration != afterGeneration-1 ||
		!validAdminProviderRouteReason(value.Reason) {
		return AdminProviderRouteActivationRecord{}, errAdminProviderRouteStoreUnavailable
	}
	computed, err := adminProviderRouteActivationRecordDigest(value)
	if err != nil || computed != value.RecordDigest {
		return AdminProviderRouteActivationRecord{}, errAdminProviderRouteStoreUnavailable
	}
	return value, nil
}

func verifyAdminProviderRouteCandidateReview(
	candidate AdminProviderRouteCandidate,
	review *AdminProviderRouteReview,
) error {
	if candidate.ReviewVersion == 0 {
		if candidate.Review != nil || review != nil {
			return errAdminProviderRouteStoreUnavailable
		}
		return nil
	}
	if candidate.Review == nil || review == nil ||
		!adminProviderRouteValuesEqual(*candidate.Review, *review) {
		return errAdminProviderRouteStoreUnavailable
	}
	return nil
}

func adminProviderRouteStorageError(err error) error {
	if err == nil ||
		errors.Is(err, errAdminProviderRouteDraftNotFound) ||
		errors.Is(err, errAdminProviderRouteDraftConflict) ||
		errors.Is(err, errAdminProviderRouteCandidateNotFound) ||
		errors.Is(err, errAdminProviderRouteCandidateConflict) ||
		errors.Is(err, errAdminProviderRouteReviewConflict) ||
		errors.Is(err, errAdminProviderRouteReviewTransition) ||
		errors.Is(err, errAdminProviderRouteGenerationConflict) ||
		errors.Is(err, errAdminProviderRouteAlreadyActive) ||
		errors.Is(err, errAdminProviderRouteRollbackTargetInvalid) {
		return err
	}
	return errAdminProviderRouteStoreUnavailable
}
