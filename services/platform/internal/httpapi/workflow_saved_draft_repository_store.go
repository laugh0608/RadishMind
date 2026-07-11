package httpapi

import "sync"

type repositorySavedWorkflowDraftStore struct {
	repository  SavedWorkflowDraftRepository
	mu          sync.RWMutex
	sideEffects SavedWorkflowDraftSideEffects
}

func newRepositorySavedWorkflowDraftStore(
	repository SavedWorkflowDraftRepository,
) *repositorySavedWorkflowDraftStore {
	return &repositorySavedWorkflowDraftStore{repository: repository}
}

func (store *repositorySavedWorkflowDraftStore) ReadDraftByID(
	requestContext SavedWorkflowDraftContext,
	draftID string,
) (SavedWorkflowDraft, bool, error) {
	actor, failureCode := repositoryActorFromSavedWorkflowDraftContext(requestContext)
	if failureCode != "" {
		return SavedWorkflowDraft{}, false, savedWorkflowDraftStoreOperationFailure(failureCode)
	}
	if store == nil || store.repository == nil {
		return SavedWorkflowDraft{}, false, savedWorkflowDraftStoreOperationFailure(
			SavedWorkflowDraftFailureStoreUnavailable,
		)
	}
	result := store.repository.ReadWorkflowDraftRecord(
		requestContext.RequestContext,
		actor,
		ReadWorkflowDraftRecordRequest{DraftID: draftID},
	)
	if result.FailureCode == SavedWorkflowDraftFailureNotFound {
		return SavedWorkflowDraft{}, false, nil
	}
	if result.FailureCode != "" {
		return SavedWorkflowDraft{}, false, savedWorkflowDraftStoreOperationFailure(result.FailureCode)
	}
	if result.Draft == nil {
		return SavedWorkflowDraft{}, false, savedWorkflowDraftStoreOperationFailure(
			SavedWorkflowDraftFailureStoreContractMismatch,
		)
	}
	return cloneSavedWorkflowDraft(*result.Draft), true, nil
}

func (store *repositorySavedWorkflowDraftStore) ListDraftSummariesByScope(
	requestContext SavedWorkflowDraftContext,
) ([]SavedWorkflowDraftSummary, error) {
	actor, failureCode := repositoryActorFromSavedWorkflowDraftContext(requestContext)
	if failureCode != "" {
		return nil, savedWorkflowDraftStoreOperationFailure(failureCode)
	}
	if store == nil || store.repository == nil {
		return nil, savedWorkflowDraftStoreOperationFailure(SavedWorkflowDraftFailureStoreUnavailable)
	}
	result := store.repository.ListWorkflowDraftRecords(
		requestContext.RequestContext,
		actor,
		ListWorkflowDraftRecordsRequest{},
	)
	if result.FailureCode != "" {
		return nil, savedWorkflowDraftStoreOperationFailure(result.FailureCode)
	}
	return append([]SavedWorkflowDraftSummary{}, result.Summaries...), nil
}

func (store *repositorySavedWorkflowDraftStore) WriteDraft(
	requestContext SavedWorkflowDraftContext,
	draft SavedWorkflowDraft,
	expectedDraftVersion int,
) (int, error) {
	actor, failureCode := repositoryActorFromSavedWorkflowDraftContext(requestContext)
	if failureCode != "" {
		return 0, savedWorkflowDraftStoreOperationFailure(failureCode)
	}
	if store == nil || store.repository == nil {
		return 0, savedWorkflowDraftStoreOperationFailure(SavedWorkflowDraftFailureStoreUnavailable)
	}
	result := store.repository.SaveWorkflowDraftRecord(
		requestContext.RequestContext,
		actor,
		SaveWorkflowDraftRecordRequest{
			ExpectedDraftVersion: expectedDraftVersion,
			Draft:                draft,
		},
	)
	if result.FailureCode != "" {
		return result.CurrentDraftVersion, savedWorkflowDraftStoreOperationFailure(result.FailureCode)
	}
	if result.Draft == nil {
		return 0, savedWorkflowDraftStoreOperationFailure(SavedWorkflowDraftFailureStoreContractMismatch)
	}
	store.mu.Lock()
	store.sideEffects.DraftWriteCount++
	store.sideEffects.ExternalRepositoryWrites++
	store.mu.Unlock()
	return result.CurrentDraftVersion, nil
}

func (store *repositorySavedWorkflowDraftStore) SideEffects() SavedWorkflowDraftSideEffects {
	if store == nil {
		return SavedWorkflowDraftSideEffects{}
	}
	store.mu.RLock()
	defer store.mu.RUnlock()
	return store.sideEffects
}

func repositoryActorFromSavedWorkflowDraftContext(
	requestContext SavedWorkflowDraftContext,
) (SavedWorkflowDraftRepositoryActorContext, SavedWorkflowDraftFailureCode) {
	actor := SavedWorkflowDraftRepositoryActorContext{
		RequestID:       requestContext.RequestID,
		TenantRef:       requestContext.TenantRef,
		WorkspaceID:     requestContext.WorkspaceID,
		ApplicationID:   requestContext.ApplicationID,
		ActorSubjectRef: requestContext.ActorRef,
		OwnerSubjectRef: requestContext.OwnerSubjectRef,
		ScopeGrants:     append([]string{}, requestContext.ScopeGrants...),
		AuditRef:        requestContext.AuditRef,
	}
	if requestContext.RequestContext == nil {
		return actor, SavedWorkflowDraftFailureAuthContextMismatch
	}
	return actor, ""
}
