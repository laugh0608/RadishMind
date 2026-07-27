package httpapi

import (
	"context"
	"strings"
)

func (adapter SavedWorkflowDraftRepositoryAdapter) ReadWorkflowDraftRevision(
	ctx context.Context,
	actor SavedWorkflowDraftRepositoryActorContext,
	request ReadSavedWorkflowDraftRevisionRecordRequest,
) ReadSavedWorkflowDraftRevisionRecordResult {
	if ctx == nil {
		return ReadSavedWorkflowDraftRevisionRecordResult{
			FailureCode: SavedWorkflowDraftFailureAuthContextMismatch,
		}
	}
	actor = normalizeSavedWorkflowDraftRepositoryActorContext(actor)
	if failure := savedWorkflowDraftRepositoryActorFailure(actor, "workflow_drafts:read"); failure != "" {
		return ReadSavedWorkflowDraftRevisionRecordResult{FailureCode: failure}
	}
	draftID := strings.TrimSpace(request.DraftID)
	if draftID == "" || request.DraftVersion < 1 {
		return ReadSavedWorkflowDraftRevisionRecordResult{
			FailureCode: SavedWorkflowDraftFailurePayloadInvalid,
		}
	}
	if failure := adapter.schemaPreflight.failureCodeFor(savedWorkflowDraftSchemaVersion); failure != "" {
		return ReadSavedWorkflowDraftRevisionRecordResult{FailureCode: failure}
	}
	executor, ok := adapter.queryExecutor.(SavedWorkflowDraftRevisionRepositoryQueryExecutor)
	if !ok {
		return ReadSavedWorkflowDraftRevisionRecordResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	result := executor.ReadWorkflowDraftRevision(ctx, savedWorkflowDraftRepositoryRevisionReadQuery{
		ActorContext: actor,
		DraftID:      draftID,
		DraftVersion: request.DraftVersion,
	})
	if result.FailureCode != "" {
		return ReadSavedWorkflowDraftRevisionRecordResult{FailureCode: result.FailureCode}
	}
	revision := result.Revision
	if failure := validateSavedWorkflowDraftRevisionScope(
		SavedWorkflowDraftContext{
			WorkspaceID:   actor.WorkspaceID,
			ApplicationID: actor.ApplicationID,
		},
		revision,
		draftID,
	); failure != "" {
		return ReadSavedWorkflowDraftRevisionRecordResult{FailureCode: failure}
	}
	return ReadSavedWorkflowDraftRevisionRecordResult{
		Revision: cloneSavedWorkflowDraftRevisionPointer(revision),
	}
}

func (adapter SavedWorkflowDraftRepositoryAdapter) ListWorkflowDraftRevisions(
	ctx context.Context,
	actor SavedWorkflowDraftRepositoryActorContext,
	request ListSavedWorkflowDraftRevisionRecordsRequest,
) ListSavedWorkflowDraftRevisionRecordsResult {
	if ctx == nil {
		return ListSavedWorkflowDraftRevisionRecordsResult{
			FailureCode: SavedWorkflowDraftFailureAuthContextMismatch,
		}
	}
	actor = normalizeSavedWorkflowDraftRepositoryActorContext(actor)
	if failure := savedWorkflowDraftRepositoryActorFailure(actor, "workflow_drafts:read"); failure != "" {
		return ListSavedWorkflowDraftRevisionRecordsResult{FailureCode: failure}
	}
	draftID := strings.TrimSpace(request.DraftID)
	if draftID == "" || request.BeforeVersion < 0 || request.Limit < 1 ||
		request.Limit > maxSavedWorkflowDraftRevisionLimit {
		return ListSavedWorkflowDraftRevisionRecordsResult{
			FailureCode: SavedWorkflowDraftFailurePayloadInvalid,
		}
	}
	if failure := adapter.schemaPreflight.failureCodeFor(savedWorkflowDraftSchemaVersion); failure != "" {
		return ListSavedWorkflowDraftRevisionRecordsResult{FailureCode: failure}
	}
	executor, ok := adapter.queryExecutor.(SavedWorkflowDraftRevisionRepositoryQueryExecutor)
	if !ok {
		return ListSavedWorkflowDraftRevisionRecordsResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	result := executor.ListWorkflowDraftRevisions(ctx, savedWorkflowDraftRepositoryRevisionListQuery{
		ActorContext:  actor,
		DraftID:       draftID,
		BeforeVersion: request.BeforeVersion,
		Limit:         request.Limit,
	})
	if result.FailureCode != "" {
		return ListSavedWorkflowDraftRevisionRecordsResult{FailureCode: result.FailureCode}
	}
	for _, revision := range result.Revisions {
		if revision.DraftID != draftID || revision.DraftVersion < 1 {
			return ListSavedWorkflowDraftRevisionRecordsResult{
				FailureCode: SavedWorkflowDraftFailureStoreContractMismatch,
			}
		}
	}
	return ListSavedWorkflowDraftRevisionRecordsResult{
		Revisions: append([]SavedWorkflowDraftRevisionSummary{}, result.Revisions...),
		HasMore:   result.HasMore,
	}
}

func normalizedSavedWorkflowDraftRevisionKind(
	kind SavedWorkflowDraftRevisionKind,
) SavedWorkflowDraftRevisionKind {
	if kind == "" {
		return SavedWorkflowDraftRevisionKindSaved
	}
	return kind
}

var _ SavedWorkflowDraftRevisionRepository = SavedWorkflowDraftRepositoryAdapter{}
