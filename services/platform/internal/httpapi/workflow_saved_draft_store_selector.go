package httpapi

import (
	"errors"
	"strings"
)

type WorkflowSavedDraftStoreMode string

const (
	WorkflowSavedDraftStoreModeMemoryDev          WorkflowSavedDraftStoreMode = "memory_dev"
	WorkflowSavedDraftStoreModeRepositoryDisabled WorkflowSavedDraftStoreMode = "repository_disabled"
	WorkflowSavedDraftStoreModeRepository         WorkflowSavedDraftStoreMode = "repository"
)

type WorkflowSavedDraftStoreSelector struct {
	MemoryDevStore savedWorkflowDraftStore
}

type WorkflowSavedDraftStoreSelection struct {
	Mode        WorkflowSavedDraftStoreMode
	Store       savedWorkflowDraftStore
	FailureCode SavedWorkflowDraftFailureCode
}

func SelectWorkflowSavedDraftStore(
	mode string,
	selector WorkflowSavedDraftStoreSelector,
) WorkflowSavedDraftStoreSelection {
	normalizedMode := strings.TrimSpace(mode)
	if normalizedMode == "" {
		normalizedMode = string(WorkflowSavedDraftStoreModeMemoryDev)
	}

	switch WorkflowSavedDraftStoreMode(normalizedMode) {
	case WorkflowSavedDraftStoreModeMemoryDev:
		store := selector.MemoryDevStore
		if store == nil {
			store = newMemorySavedWorkflowDraftStore()
		}
		return WorkflowSavedDraftStoreSelection{
			Mode:  WorkflowSavedDraftStoreModeMemoryDev,
			Store: store,
		}
	case WorkflowSavedDraftStoreModeRepositoryDisabled:
		return disabledWorkflowSavedDraftStoreSelection(
			WorkflowSavedDraftStoreModeRepositoryDisabled,
			SavedWorkflowDraftFailureRepositoryStoreDisabled,
		)
	case WorkflowSavedDraftStoreModeRepository:
		return disabledWorkflowSavedDraftStoreSelection(
			WorkflowSavedDraftStoreModeRepository,
			SavedWorkflowDraftFailureRepositoryStoreDisabled,
		)
	default:
		return disabledWorkflowSavedDraftStoreSelection(
			WorkflowSavedDraftStoreMode(normalizedMode),
			SavedWorkflowDraftFailureInvalidStoreMode,
		)
	}
}

type disabledSavedWorkflowDraftStore struct {
	failureCode SavedWorkflowDraftFailureCode
}

func disabledWorkflowSavedDraftStoreSelection(
	mode WorkflowSavedDraftStoreMode,
	failureCode SavedWorkflowDraftFailureCode,
) WorkflowSavedDraftStoreSelection {
	return WorkflowSavedDraftStoreSelection{
		Mode:        mode,
		Store:       disabledSavedWorkflowDraftStore{failureCode: failureCode},
		FailureCode: failureCode,
	}
}

func (store disabledSavedWorkflowDraftStore) ReadDraftByID(
	_ string,
) (SavedWorkflowDraft, bool, error) {
	return SavedWorkflowDraft{}, false, savedWorkflowDraftStoreSelectionFailure(store.failureCode)
}

func (store disabledSavedWorkflowDraftStore) ListDraftsByScope(
	_ string,
	_ string,
) ([]SavedWorkflowDraft, error) {
	return nil, savedWorkflowDraftStoreSelectionFailure(store.failureCode)
}

func (store disabledSavedWorkflowDraftStore) WriteDraft(_ SavedWorkflowDraft) error {
	return savedWorkflowDraftStoreSelectionFailure(store.failureCode)
}

func (store disabledSavedWorkflowDraftStore) SideEffects() SavedWorkflowDraftSideEffects {
	return SavedWorkflowDraftSideEffects{}
}

type savedWorkflowDraftStoreSelectionError struct {
	failureCode SavedWorkflowDraftFailureCode
}

func (err *savedWorkflowDraftStoreSelectionError) Error() string {
	return string(err.failureCode)
}

func savedWorkflowDraftStoreSelectionFailure(failureCode SavedWorkflowDraftFailureCode) error {
	return &savedWorkflowDraftStoreSelectionError{failureCode: failureCode}
}

func savedWorkflowDraftStoreFailureCode(err error) SavedWorkflowDraftFailureCode {
	if err == nil {
		return ""
	}
	var selectionErr *savedWorkflowDraftStoreSelectionError
	if errors.As(err, &selectionErr) {
		return selectionErr.failureCode
	}
	return SavedWorkflowDraftFailureStoreUnavailable
}
