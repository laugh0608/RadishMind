package httpapi

import (
	"errors"
	"strings"
)

type WorkflowSavedDraftStoreMode string

const (
	WorkflowSavedDraftStoreModeMemoryDev          WorkflowSavedDraftStoreMode = "memory_dev"
	WorkflowSavedDraftStoreModeSQLiteDev          WorkflowSavedDraftStoreMode = "sqlite_dev"
	WorkflowSavedDraftStoreModePostgresDevTest    WorkflowSavedDraftStoreMode = "postgres_dev_test"
	WorkflowSavedDraftStoreModeRepositoryDisabled WorkflowSavedDraftStoreMode = "repository_disabled"
	WorkflowSavedDraftStoreModeRepository         WorkflowSavedDraftStoreMode = "repository"
)

type WorkflowSavedDraftStoreSelector struct {
	MemoryDevStore       savedWorkflowDraftStore
	SQLiteDevStore       savedWorkflowDraftStore
	PostgresDevTestStore savedWorkflowDraftStore
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
	case WorkflowSavedDraftStoreModeSQLiteDev:
		if selector.SQLiteDevStore == nil {
			return disabledWorkflowSavedDraftStoreSelection(
				WorkflowSavedDraftStoreModeSQLiteDev,
				SavedWorkflowDraftFailureStoreUnavailable,
			)
		}
		return WorkflowSavedDraftStoreSelection{
			Mode:  WorkflowSavedDraftStoreModeSQLiteDev,
			Store: selector.SQLiteDevStore,
		}
	case WorkflowSavedDraftStoreModePostgresDevTest:
		if selector.PostgresDevTestStore == nil {
			return disabledWorkflowSavedDraftStoreSelection(
				WorkflowSavedDraftStoreModePostgresDevTest,
				SavedWorkflowDraftFailureStoreUnavailable,
			)
		}
		return WorkflowSavedDraftStoreSelection{
			Mode:  WorkflowSavedDraftStoreModePostgresDevTest,
			Store: selector.PostgresDevTestStore,
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
	_ SavedWorkflowDraftContext,
	_ string,
) (SavedWorkflowDraft, bool, error) {
	return SavedWorkflowDraft{}, false, savedWorkflowDraftStoreSelectionFailure(store.failureCode)
}

func (store disabledSavedWorkflowDraftStore) ListDraftSummariesByScope(
	_ SavedWorkflowDraftContext,
) ([]SavedWorkflowDraftSummary, error) {
	return nil, savedWorkflowDraftStoreSelectionFailure(store.failureCode)
}

func (store disabledSavedWorkflowDraftStore) WriteDraft(
	_ SavedWorkflowDraftContext,
	_ SavedWorkflowDraft,
	_ int,
) (int, error) {
	return 0, savedWorkflowDraftStoreSelectionFailure(store.failureCode)
}

func (store disabledSavedWorkflowDraftStore) SideEffects() SavedWorkflowDraftSideEffects {
	return SavedWorkflowDraftSideEffects{}
}

type savedWorkflowDraftStoreSelectionError struct {
	failureCode SavedWorkflowDraftFailureCode
}

type savedWorkflowDraftStoreWriteError struct {
	failureCode SavedWorkflowDraftFailureCode
}

type savedWorkflowDraftStoreOperationError struct {
	failureCode SavedWorkflowDraftFailureCode
}

func (err *savedWorkflowDraftStoreSelectionError) Error() string {
	return string(err.failureCode)
}

func savedWorkflowDraftStoreSelectionFailure(failureCode SavedWorkflowDraftFailureCode) error {
	return &savedWorkflowDraftStoreSelectionError{failureCode: failureCode}
}

func (err *savedWorkflowDraftStoreWriteError) Error() string {
	return string(err.failureCode)
}

func savedWorkflowDraftStoreWriteFailure(failureCode SavedWorkflowDraftFailureCode) error {
	return &savedWorkflowDraftStoreWriteError{failureCode: failureCode}
}

func (err *savedWorkflowDraftStoreOperationError) Error() string {
	return string(err.failureCode)
}

func savedWorkflowDraftStoreOperationFailure(failureCode SavedWorkflowDraftFailureCode) error {
	return &savedWorkflowDraftStoreOperationError{failureCode: failureCode}
}

func savedWorkflowDraftStoreFailureCode(err error) SavedWorkflowDraftFailureCode {
	if err == nil {
		return ""
	}
	var selectionErr *savedWorkflowDraftStoreSelectionError
	if errors.As(err, &selectionErr) {
		return selectionErr.failureCode
	}
	var writeErr *savedWorkflowDraftStoreWriteError
	if errors.As(err, &writeErr) {
		return writeErr.failureCode
	}
	var operationErr *savedWorkflowDraftStoreOperationError
	if errors.As(err, &operationErr) {
		return operationErr.failureCode
	}
	return SavedWorkflowDraftFailureStoreUnavailable
}
