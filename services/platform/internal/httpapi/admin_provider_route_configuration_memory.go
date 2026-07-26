package httpapi

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type memoryAdminProviderRouteRepository struct {
	mu          sync.RWMutex
	drafts      map[string]AdminProviderRouteConfigurationDraft
	candidates  map[string]AdminProviderRouteCandidate
	snapshots   map[string]AdminProviderRouteSnapshot
	activations map[string][]AdminProviderRouteActivationRecord
	unavailable bool
}

func newMemoryAdminProviderRouteRepository() *memoryAdminProviderRouteRepository {
	return &memoryAdminProviderRouteRepository{
		drafts:      make(map[string]AdminProviderRouteConfigurationDraft),
		candidates:  make(map[string]AdminProviderRouteCandidate),
		snapshots:   make(map[string]AdminProviderRouteSnapshot),
		activations: make(map[string][]AdminProviderRouteActivationRecord),
	}
}

func (repository *memoryAdminProviderRouteRepository) PutDraft(
	ctx AdminProviderRouteContext,
	expectedRevision int,
	draft AdminProviderRouteConfigurationDraft,
	now time.Time,
) (AdminProviderRouteConfigurationDraft, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return AdminProviderRouteConfigurationDraft{}, errAdminProviderRouteStoreUnavailable
	}
	key := adminProviderRouteConfigurationKey(ctx, draft.ConfigurationID)
	current, exists := repository.drafts[key]
	currentRevision := 0
	if exists {
		currentRevision = current.DraftRevision
	}
	if currentRevision != expectedRevision {
		return AdminProviderRouteConfigurationDraft{}, adminProviderRouteDraftConflictError{CurrentRevision: currentRevision}
	}
	timestamp := now.UTC().Format(time.RFC3339Nano)
	draft.DraftRevision = currentRevision + 1
	draft.CreatedAt = timestamp
	draft.CreatedByActorRef = ctx.ActorRef
	if exists {
		draft.CreatedAt = current.CreatedAt
		draft.CreatedByActorRef = current.CreatedByActorRef
	}
	draft.UpdatedAt = timestamp
	draft.UpdatedByActorRef = ctx.ActorRef
	draft.RequestID = ctx.RequestID
	draft.AuditRef = ctx.AuditRef
	repository.drafts[key] = cloneAdminProviderRouteDraft(draft)
	return cloneAdminProviderRouteDraft(draft), nil
}

func (repository *memoryAdminProviderRouteRepository) ReadDraft(
	ctx AdminProviderRouteContext,
	configurationID string,
) (AdminProviderRouteConfigurationDraft, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return AdminProviderRouteConfigurationDraft{}, errAdminProviderRouteStoreUnavailable
	}
	value, exists := repository.drafts[adminProviderRouteConfigurationKey(ctx, configurationID)]
	if !exists {
		return AdminProviderRouteConfigurationDraft{}, errAdminProviderRouteDraftNotFound
	}
	return cloneAdminProviderRouteDraft(value), nil
}

func (repository *memoryAdminProviderRouteRepository) CreateCandidate(
	ctx AdminProviderRouteContext,
	candidate AdminProviderRouteCandidate,
) (AdminProviderRouteCandidate, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return AdminProviderRouteCandidate{}, errAdminProviderRouteStoreUnavailable
	}
	key := adminProviderRouteCandidateKey(ctx, candidate.ConfigurationID, candidate.CandidateID)
	if existing, exists := repository.candidates[key]; exists {
		if existing.CandidateDigest == candidate.CandidateDigest {
			return AdminProviderRouteCandidate{}, errAdminProviderRouteCandidateConflict
		}
		return AdminProviderRouteCandidate{}, errAdminProviderRouteCandidateConflict
	}
	repository.candidates[key] = cloneAdminProviderRouteCandidate(candidate)
	return cloneAdminProviderRouteCandidate(candidate), nil
}

func (repository *memoryAdminProviderRouteRepository) ReadCandidate(
	ctx AdminProviderRouteContext,
	configurationID string,
	candidateID string,
) (AdminProviderRouteCandidate, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return AdminProviderRouteCandidate{}, errAdminProviderRouteStoreUnavailable
	}
	value, exists := repository.candidates[adminProviderRouteCandidateKey(ctx, configurationID, candidateID)]
	if !exists {
		return AdminProviderRouteCandidate{}, errAdminProviderRouteCandidateNotFound
	}
	return cloneAdminProviderRouteCandidate(value), nil
}

func (repository *memoryAdminProviderRouteRepository) ReviewCandidate(
	ctx AdminProviderRouteContext,
	configurationID string,
	candidateID string,
	expectedReviewVersion int,
	review AdminProviderRouteReview,
) (AdminProviderRouteCandidate, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return AdminProviderRouteCandidate{}, errAdminProviderRouteStoreUnavailable
	}
	key := adminProviderRouteCandidateKey(ctx, configurationID, candidateID)
	current, exists := repository.candidates[key]
	if !exists {
		return AdminProviderRouteCandidate{}, errAdminProviderRouteCandidateNotFound
	}
	if current.ReviewVersion != expectedReviewVersion {
		return AdminProviderRouteCandidate{}, adminProviderRouteReviewConflictError{
			CurrentReviewVersion: current.ReviewVersion, CurrentState: current.CandidateState,
		}
	}
	if current.CandidateState != adminProviderRouteCandidatePending || current.Review != nil {
		return AdminProviderRouteCandidate{}, errAdminProviderRouteReviewTransition
	}
	if review.ReviewVersion != expectedReviewVersion+1 ||
		(review.ResultingState != adminProviderRouteCandidateApproved && review.ResultingState != adminProviderRouteCandidateRejected) ||
		(review.Decision == adminProviderRouteDecisionApprove && review.ResultingState != adminProviderRouteCandidateApproved) ||
		(review.Decision == adminProviderRouteDecisionReject && review.ResultingState != adminProviderRouteCandidateRejected) {
		return AdminProviderRouteCandidate{}, errAdminProviderRouteReviewTransition
	}
	current.ReviewVersion = review.ReviewVersion
	current.CandidateState = review.ResultingState
	reviewCopy := review
	current.Review = &reviewCopy
	repository.candidates[key] = cloneAdminProviderRouteCandidate(current)
	return cloneAdminProviderRouteCandidate(current), nil
}

func (repository *memoryAdminProviderRouteRepository) CommitActivation(
	ctx AdminProviderRouteContext,
	configurationID string,
	candidateID string,
	expectedGeneration int,
	action string,
	reason string,
	snapshot AdminProviderRouteSnapshot,
	now time.Time,
) (AdminProviderRouteSnapshot, AdminProviderRouteActivationRecord, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return AdminProviderRouteSnapshot{}, AdminProviderRouteActivationRecord{}, errAdminProviderRouteStoreUnavailable
	}
	candidate, exists := repository.candidates[adminProviderRouteCandidateKey(ctx, configurationID, candidateID)]
	if !exists {
		return AdminProviderRouteSnapshot{}, AdminProviderRouteActivationRecord{}, errAdminProviderRouteCandidateNotFound
	}
	if candidate.CandidateState != adminProviderRouteCandidateApproved || candidate.Review == nil {
		return AdminProviderRouteSnapshot{}, AdminProviderRouteActivationRecord{}, errAdminProviderRouteReviewTransition
	}
	if snapshot.ConfigurationID != candidate.ConfigurationID ||
		snapshot.CandidateID != candidate.CandidateID ||
		snapshot.CandidateDigest != candidate.CandidateDigest ||
		!adminProviderRouteValuesEqual(snapshot.Configuration, candidate.Configuration) ||
		!adminProviderRouteValuesEqual(snapshot.InventoryBindings, candidate.InventoryBindings) {
		return AdminProviderRouteSnapshot{}, AdminProviderRouteActivationRecord{}, errAdminProviderRouteStoreUnavailable
	}
	key := adminProviderRouteConfigurationKey(ctx, configurationID)
	current, hasCurrent := repository.snapshots[key]
	currentGeneration := 0
	if hasCurrent {
		currentGeneration = current.Generation
	}
	if currentGeneration != expectedGeneration {
		return AdminProviderRouteSnapshot{}, AdminProviderRouteActivationRecord{},
			adminProviderRouteGenerationConflictError{CurrentGeneration: currentGeneration}
	}
	if hasCurrent && current.CandidateID == candidateID {
		return AdminProviderRouteSnapshot{}, AdminProviderRouteActivationRecord{}, errAdminProviderRouteAlreadyActive
	}
	if action == adminProviderRouteActionRollback && !repository.wasCandidateActivatedLocked(key, candidateID) {
		return AdminProviderRouteSnapshot{}, AdminProviderRouteActivationRecord{}, errAdminProviderRouteRollbackTargetInvalid
	}
	if action != adminProviderRouteActionActivate && action != adminProviderRouteActionRollback {
		return AdminProviderRouteSnapshot{}, AdminProviderRouteActivationRecord{}, errAdminProviderRouteReviewTransition
	}
	timestamp := now.UTC().Format(time.RFC3339Nano)
	snapshot.Generation = currentGeneration + 1
	snapshot.ActivatedAt = timestamp
	beforeCandidateID := ""
	beforeSnapshotDigest := ""
	if hasCurrent {
		beforeCandidateID = current.CandidateID
		beforeSnapshotDigest = current.SnapshotDigest
	}
	previousRecordDigest := ""
	if history := repository.activations[key]; len(history) > 0 {
		previousRecordDigest = history[len(history)-1].RecordDigest
	}
	activation := AdminProviderRouteActivationRecord{
		SchemaVersion:   adminProviderRouteActivationSchemaVersion,
		ActivationID:    fmt.Sprintf("provider-route-activation-%d", snapshot.Generation),
		ConfigurationID: configurationID, Action: action, Reason: reason,
		BeforeGeneration: currentGeneration, AfterGeneration: snapshot.Generation,
		BeforeCandidateID: beforeCandidateID, BeforeSnapshotDigest: beforeSnapshotDigest,
		AfterCandidateID: snapshot.CandidateID, AfterSnapshotDigest: snapshot.SnapshotDigest,
		PreviousRecordDigest: previousRecordDigest,
		CreatedAt:            timestamp, ActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	recordDigest, err := adminProviderRouteActivationRecordDigest(activation)
	if err != nil {
		return AdminProviderRouteSnapshot{}, AdminProviderRouteActivationRecord{}, errAdminProviderRouteStoreUnavailable
	}
	activation.RecordDigest = recordDigest
	repository.snapshots[key] = cloneAdminProviderRouteSnapshot(snapshot)
	repository.activations[key] = append(repository.activations[key], activation)
	return cloneAdminProviderRouteSnapshot(snapshot), activation, nil
}

func (repository *memoryAdminProviderRouteRepository) ReadActiveSnapshot(
	ctx AdminProviderRouteContext,
	configurationID string,
) (AdminProviderRouteSnapshot, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return AdminProviderRouteSnapshot{}, errAdminProviderRouteStoreUnavailable
	}
	value, exists := repository.snapshots[adminProviderRouteConfigurationKey(ctx, configurationID)]
	if !exists {
		return AdminProviderRouteSnapshot{}, errAdminProviderRouteCandidateNotFound
	}
	return cloneAdminProviderRouteSnapshot(value), nil
}

func (repository *memoryAdminProviderRouteRepository) ListActivations(
	ctx AdminProviderRouteContext,
	configurationID string,
) ([]AdminProviderRouteActivationRecord, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return nil, errAdminProviderRouteStoreUnavailable
	}
	stored := repository.activations[adminProviderRouteConfigurationKey(ctx, configurationID)]
	values := append([]AdminProviderRouteActivationRecord(nil), stored...)
	sort.Slice(values, func(i, j int) bool {
		return values[i].AfterGeneration < values[j].AfterGeneration
	})
	return values, nil
}

func (repository *memoryAdminProviderRouteRepository) wasCandidateActivatedLocked(key, candidateID string) bool {
	for _, activation := range repository.activations[key] {
		if activation.AfterCandidateID == candidateID {
			return true
		}
	}
	return false
}

func (repository *memoryAdminProviderRouteRepository) setUnavailableForTest(unavailable bool) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	repository.unavailable = unavailable
}

func adminProviderRouteScopePrefix(ctx AdminProviderRouteContext) string {
	return strings.Join([]string{
		strings.TrimSpace(ctx.TenantRef),
		strings.TrimSpace(ctx.WorkspaceID),
		strings.TrimSpace(ctx.Environment),
	}, "\x00")
}

func adminProviderRouteConfigurationKey(ctx AdminProviderRouteContext, configurationID string) string {
	return adminProviderRouteScopePrefix(ctx) + "\x00" + strings.TrimSpace(configurationID)
}

func adminProviderRouteCandidateKey(ctx AdminProviderRouteContext, configurationID, candidateID string) string {
	return adminProviderRouteConfigurationKey(ctx, configurationID) + "\x00" + strings.TrimSpace(candidateID)
}
