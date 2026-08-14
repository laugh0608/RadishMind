package httpapi

import (
	"errors"
	"sort"
	"sync"
)

type memorySavedWorkflowDraftStore struct {
	mu                         sync.RWMutex
	drafts                     map[string]SavedWorkflowDraft
	revisions                  map[string]map[int]SavedWorkflowDraftRevision
	lifecycleEvents            map[string]map[int]SavedWorkflowDraftLifecycleEvent
	sideEffects                SavedWorkflowDraftSideEffects
	unavailable                bool
	lifecycleEventWriteFailure bool
}

func newMemorySavedWorkflowDraftStore() *memorySavedWorkflowDraftStore {
	return &memorySavedWorkflowDraftStore{
		drafts:          make(map[string]SavedWorkflowDraft),
		revisions:       make(map[string]map[int]SavedWorkflowDraftRevision),
		lifecycleEvents: make(map[string]map[int]SavedWorkflowDraftLifecycleEvent),
	}
}

func (store *memorySavedWorkflowDraftStore) ReadDraftByID(
	_ SavedWorkflowDraftContext,
	draftID string,
) (SavedWorkflowDraft, bool, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	if store.unavailable {
		return SavedWorkflowDraft{}, false, errors.New("saved workflow draft store unavailable")
	}
	draft, found := store.drafts[draftID]
	if !found {
		return SavedWorkflowDraft{}, false, nil
	}
	normalized, ok := normalizeAndValidateSavedWorkflowDraftLifecycle(draft)
	if !ok {
		return SavedWorkflowDraft{}, false, savedWorkflowDraftStoreOperationFailure(
			SavedWorkflowDraftFailureLifecycleStoreContract,
		)
	}
	return normalized, true, nil
}

func (store *memorySavedWorkflowDraftStore) ListDraftSummariesByScope(
	requestContext SavedWorkflowDraftContext,
) ([]SavedWorkflowDraftSummary, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	if store.unavailable {
		return nil, errors.New("saved workflow draft store unavailable")
	}
	drafts := make([]SavedWorkflowDraft, 0)
	for _, draft := range store.drafts {
		if savedWorkflowDraftMatchesScope(draft, requestContext.WorkspaceID, requestContext.ApplicationID) {
			drafts = append(drafts, cloneSavedWorkflowDraft(draft))
		}
	}
	sort.Slice(drafts, func(i, j int) bool {
		if drafts[i].UpdatedAt == drafts[j].UpdatedAt {
			return drafts[i].DraftID < drafts[j].DraftID
		}
		return drafts[i].UpdatedAt > drafts[j].UpdatedAt
	})
	summaries := make([]SavedWorkflowDraftSummary, 0, len(drafts))
	for _, draft := range drafts {
		summaries = append(summaries, savedWorkflowDraftSummaryFromDraft(draft))
	}
	return summaries, nil
}

func (store *memorySavedWorkflowDraftStore) WriteDraft(
	_ SavedWorkflowDraftContext,
	draft SavedWorkflowDraft,
	expectedDraftVersion int,
) (int, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	return store.writeDraftLocked(draft, expectedDraftVersion, SavedWorkflowDraftRevisionKindSaved, 0)
}

func (store *memorySavedWorkflowDraftStore) writeDraftLocked(
	draft SavedWorkflowDraft,
	expectedDraftVersion int,
	revisionKind SavedWorkflowDraftRevisionKind,
	restoredFromVersion int,
) (int, error) {
	if store.unavailable {
		return 0, errors.New("saved workflow draft store unavailable")
	}
	existing, found := store.drafts[draft.DraftID]
	if found && !savedWorkflowDraftMatchesScope(existing, draft.WorkspaceID, draft.ApplicationID) {
		return 0, savedWorkflowDraftStoreWriteFailure(SavedWorkflowDraftFailureScopeDenied)
	}
	if found {
		var lifecycleOK bool
		existing, lifecycleOK = normalizeAndValidateSavedWorkflowDraftLifecycle(existing)
		if !lifecycleOK {
			return 0, savedWorkflowDraftStoreWriteFailure(
				SavedWorkflowDraftFailureLifecycleStoreContract,
			)
		}
		if existing.LifecycleState != SavedWorkflowDraftLifecycleActive {
			return existing.DraftVersion, savedWorkflowDraftStoreWriteFailure(
				SavedWorkflowDraftFailureArchived,
			)
		}
	}
	if found && existing.DraftVersion != expectedDraftVersion {
		return existing.DraftVersion, savedWorkflowDraftStoreWriteFailure(SavedWorkflowDraftFailureVersionConflict)
	}
	if !found && expectedDraftVersion != 0 {
		return 0, savedWorkflowDraftStoreWriteFailure(SavedWorkflowDraftFailureNotFound)
	}
	normalizedDraft, lifecycleOK := normalizeAndValidateSavedWorkflowDraftLifecycle(draft)
	if !lifecycleOK {
		return expectedDraftVersion, savedWorkflowDraftStoreWriteFailure(
			SavedWorkflowDraftFailureLifecycleStoreContract,
		)
	}
	if normalizedDraft.LifecycleState != SavedWorkflowDraftLifecycleActive ||
		normalizedDraft.ArchivedAt != "" ||
		normalizedDraft.LibraryUpdatedAt != normalizedDraft.UpdatedAt ||
		(!found &&
			(normalizedDraft.LifecycleVersion != 1 ||
				normalizedDraft.LifecycleUpdatedByActorRef != "")) {
		return expectedDraftVersion, savedWorkflowDraftStoreWriteFailure(
			SavedWorkflowDraftFailureLifecycleStoreContract,
		)
	}
	if found &&
		(normalizedDraft.LifecycleState != existing.LifecycleState ||
			normalizedDraft.LifecycleVersion != existing.LifecycleVersion ||
			normalizedDraft.ArchivedAt != existing.ArchivedAt ||
			normalizedDraft.LifecycleUpdatedByActorRef != existing.LifecycleUpdatedByActorRef) {
		return existing.DraftVersion, savedWorkflowDraftStoreWriteFailure(
			SavedWorkflowDraftFailureLifecycleVersionConflict,
		)
	}
	revisions := store.revisions[draft.DraftID]
	if revisions == nil {
		revisions = make(map[int]SavedWorkflowDraftRevision)
	}
	if _, duplicate := revisions[draft.DraftVersion]; duplicate {
		return expectedDraftVersion, savedWorkflowDraftStoreWriteFailure(
			SavedWorkflowDraftFailureStoreContractMismatch,
		)
	}
	revision := newSavedWorkflowDraftRevision(normalizedDraft, revisionKind, restoredFromVersion)
	store.drafts[normalizedDraft.DraftID] = cloneSavedWorkflowDraft(normalizedDraft)
	revisions[normalizedDraft.DraftVersion] = cloneSavedWorkflowDraftRevision(revision)
	store.revisions[normalizedDraft.DraftID] = revisions
	store.sideEffects.DraftWriteCount++
	return normalizedDraft.DraftVersion, nil
}

func (store *memorySavedWorkflowDraftStore) SideEffects() SavedWorkflowDraftSideEffects {
	store.mu.RLock()
	defer store.mu.RUnlock()

	return store.sideEffects
}
