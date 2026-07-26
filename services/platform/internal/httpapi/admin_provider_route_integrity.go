package httpapi

import (
	"encoding/json"
	"strconv"
)

func validAdminProviderRouteDraftIntegrity(
	ctx AdminProviderRouteContext,
	draft AdminProviderRouteConfigurationDraft,
) bool {
	if draft.SchemaVersion != adminProviderRouteDraftSchemaVersion ||
		draft.TenantRef != ctx.TenantRef || draft.WorkspaceID != ctx.WorkspaceID ||
		draft.Environment != ctx.Environment || draft.DraftRevision < 1 ||
		!adminProviderRouteIdentifierPattern.MatchString(draft.ConfigurationID) ||
		!adminProviderRouteDigestPattern.MatchString(draft.DraftDigest) {
		return false
	}
	stored := adminProviderRouteConfigurationFromDraft(draft)
	normalized, failureCode := normalizeAdminProviderRouteConfiguration(
		stored.DisplayName, stored.ProviderProfiles, stored.ModelRoutes, draft.Environment,
	)
	if failureCode != "" || !adminProviderRouteValuesEqual(stored, normalized) {
		return false
	}
	digest, err := adminProviderRouteCanonicalDigest(normalized)
	return err == nil && digest == draft.DraftDigest
}

func validAdminProviderRouteCandidateIntegrity(candidate AdminProviderRouteCandidate) bool {
	if candidate.SchemaVersion != adminProviderRouteCandidateSchemaVersion ||
		candidate.SourceDraftRevision < 1 ||
		!adminProviderRouteIdentifierPattern.MatchString(candidate.ConfigurationID) ||
		!adminProviderRouteIdentifierPattern.MatchString(candidate.CandidateID) ||
		!adminProviderRouteDigestPattern.MatchString(candidate.SourceDraftDigest) ||
		!adminProviderRouteDigestPattern.MatchString(candidate.CandidateDigest) ||
		!validAdminProviderRouteCandidateReview(candidate) {
		return false
	}
	normalized, failureCode := normalizeAdminProviderRouteConfiguration(
		candidate.Configuration.DisplayName,
		candidate.Configuration.ProviderProfiles,
		candidate.Configuration.ModelRoutes,
		candidate.Environment,
	)
	if failureCode != "" || !adminProviderRouteValuesEqual(candidate.Configuration, normalized) ||
		!validAdminProviderRouteBindings(candidate.Configuration.ProviderProfiles, candidate.InventoryBindings, candidate.Environment) {
		return false
	}
	digest, err := adminProviderRouteCandidateDigest(
		candidate.ConfigurationID,
		candidate.SourceDraftRevision,
		candidate.SourceDraftDigest,
		candidate.Configuration,
		candidate.InventoryBindings,
	)
	return err == nil && digest == candidate.CandidateDigest
}

func validAdminProviderRouteSnapshotIntegrity(
	ctx AdminProviderRouteContext,
	snapshot AdminProviderRouteSnapshot,
) bool {
	if snapshot.SchemaVersion != adminProviderRouteSnapshotSchemaVersion ||
		snapshot.TenantRef != ctx.TenantRef || snapshot.WorkspaceID != ctx.WorkspaceID ||
		snapshot.Environment != ctx.Environment || snapshot.Generation < 1 ||
		!adminProviderRouteIdentifierPattern.MatchString(snapshot.ConfigurationID) ||
		!adminProviderRouteIdentifierPattern.MatchString(snapshot.CandidateID) ||
		!adminProviderRouteDigestPattern.MatchString(snapshot.CandidateDigest) ||
		!adminProviderRouteDigestPattern.MatchString(snapshot.SnapshotDigest) {
		return false
	}
	normalized, failureCode := normalizeAdminProviderRouteConfiguration(
		snapshot.Configuration.DisplayName,
		snapshot.Configuration.ProviderProfiles,
		snapshot.Configuration.ModelRoutes,
		snapshot.Environment,
	)
	if failureCode != "" || !adminProviderRouteValuesEqual(snapshot.Configuration, normalized) ||
		!validAdminProviderRouteBindings(snapshot.Configuration.ProviderProfiles, snapshot.InventoryBindings, snapshot.Environment) {
		return false
	}
	digest, err := adminProviderRouteCanonicalDigest(struct {
		ConfigurationID string                                  `json:"configuration_id"`
		CandidateID     string                                  `json:"candidate_id"`
		CandidateDigest string                                  `json:"candidate_digest"`
		Configuration   AdminProviderRouteConfigurationSnapshot `json:"configuration"`
		Bindings        []AdminProviderInventoryBinding         `json:"inventory_bindings"`
	}{
		ConfigurationID: snapshot.ConfigurationID,
		CandidateID:     snapshot.CandidateID,
		CandidateDigest: snapshot.CandidateDigest,
		Configuration:   snapshot.Configuration,
		Bindings:        snapshot.InventoryBindings,
	})
	return err == nil && digest == snapshot.SnapshotDigest
}

func validAdminProviderRouteActivationHistory(
	configurationID string,
	history []AdminProviderRouteActivationRecord,
) bool {
	for index, record := range history {
		expectedGeneration := index + 1
		if record.SchemaVersion != adminProviderRouteActivationSchemaVersion ||
			record.ConfigurationID != configurationID ||
			record.BeforeGeneration != index || record.AfterGeneration != expectedGeneration ||
			record.ActivationID != "provider-route-activation-"+strconv.Itoa(expectedGeneration) ||
			(record.Action != adminProviderRouteActionActivate && record.Action != adminProviderRouteActionRollback) ||
			!validAdminProviderRouteReason(record.Reason) ||
			!adminProviderRouteIdentifierPattern.MatchString(record.AfterCandidateID) ||
			!adminProviderRouteDigestPattern.MatchString(record.AfterSnapshotDigest) ||
			!adminProviderRouteDigestPattern.MatchString(record.RecordDigest) {
			return false
		}
		digest, err := adminProviderRouteActivationRecordDigest(record)
		if err != nil || digest != record.RecordDigest {
			return false
		}
		if index == 0 {
			if record.BeforeCandidateID != "" || record.BeforeSnapshotDigest != "" ||
				record.PreviousRecordDigest != "" || record.Action == adminProviderRouteActionRollback {
				return false
			}
			continue
		}
		previous := history[index-1]
		if record.BeforeCandidateID != previous.AfterCandidateID ||
			record.BeforeSnapshotDigest != previous.AfterSnapshotDigest ||
			record.PreviousRecordDigest != previous.RecordDigest {
			return false
		}
	}
	return true
}

func adminProviderRouteActivationRecordDigest(record AdminProviderRouteActivationRecord) (string, error) {
	record.RecordDigest = ""
	return adminProviderRouteCanonicalDigest(record)
}

func validAdminProviderRouteCandidateReview(candidate AdminProviderRouteCandidate) bool {
	switch candidate.CandidateState {
	case adminProviderRouteCandidatePending:
		return candidate.ReviewVersion == 0 && candidate.Review == nil
	case adminProviderRouteCandidateApproved, adminProviderRouteCandidateRejected:
		if candidate.ReviewVersion != 1 || candidate.Review == nil ||
			candidate.Review.SchemaVersion != adminProviderRouteReviewSchemaVersion ||
			candidate.Review.ReviewVersion != candidate.ReviewVersion ||
			candidate.Review.ResultingState != candidate.CandidateState ||
			!validAdminProviderRouteReason(candidate.Review.Reason) {
			return false
		}
		if candidate.CandidateState == adminProviderRouteCandidateApproved {
			return candidate.Review.Decision == adminProviderRouteDecisionApprove
		}
		return candidate.Review.Decision == adminProviderRouteDecisionReject
	default:
		return false
	}
}

func validAdminProviderRouteBindings(
	profiles []AdminProviderProfileAssignment,
	bindings []AdminProviderInventoryBinding,
	environment string,
) bool {
	if len(profiles) != len(bindings) {
		return false
	}
	profilesByID := make(map[string]AdminProviderProfileAssignment, len(profiles))
	for _, profile := range profiles {
		profilesByID[profile.ProfileID] = profile
	}
	previousProfileID := ""
	for _, binding := range bindings {
		normalized, ok := normalizeAdminProviderInventoryBinding(binding)
		profile, exists := profilesByID[binding.ProfileID]
		if !ok || !exists || !adminProviderRouteValuesEqual(binding, normalized) ||
			binding.ProfileID <= previousProfileID || !binding.Enabled ||
			binding.ProviderID != profile.ProviderID ||
			binding.RuntimeProfileRef != profile.RuntimeProfileRef ||
			binding.Environment != environment ||
			!adminProviderRouteCapabilitiesContain(binding.Capabilities, profile.Capabilities) {
			return false
		}
		previousProfileID = binding.ProfileID
	}
	return true
}

func adminProviderRouteValuesEqual(left, right any) bool {
	leftPayload, leftErr := json.Marshal(left)
	rightPayload, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && string(leftPayload) == string(rightPayload)
}
