package httpapi

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"
)

// repositorySavedWorkflowDraftLibraryStore exposes the repository library contract
// to service callers that explicitly opt into the batch C HTTP surface. Keeping the
// wrapper distinct prevents the legacy HTTP list route from silently adopting the
// paginated default before its request and response contract is extended.
type repositorySavedWorkflowDraftLibraryStore struct {
	*repositorySavedWorkflowDraftStore
}

func newRepositorySavedWorkflowDraftLibraryStore(
	store *repositorySavedWorkflowDraftStore,
) *repositorySavedWorkflowDraftLibraryStore {
	return &repositorySavedWorkflowDraftLibraryStore{
		repositorySavedWorkflowDraftStore: store,
	}
}

func (adapter SavedWorkflowDraftRepositoryAdapter) ListWorkflowDraftLibraryPage(
	ctx context.Context,
	actor SavedWorkflowDraftRepositoryActorContext,
	request ListWorkflowDraftLibraryPageRequest,
) ListWorkflowDraftLibraryPageResult {
	if ctx == nil {
		return ListWorkflowDraftLibraryPageResult{
			FailureCode: SavedWorkflowDraftFailureAuthContextMismatch,
		}
	}
	actor = normalizeSavedWorkflowDraftRepositoryActorContext(actor)
	if failureCode := savedWorkflowDraftRepositoryActorFailure(
		actor,
		"workflow_drafts:read",
	); failureCode != "" {
		return ListWorkflowDraftLibraryPageResult{FailureCode: failureCode}
	}
	if failureCode := adapter.schemaPreflight.failureCodeFor(
		savedWorkflowDraftSchemaVersion,
	); failureCode != "" {
		return ListWorkflowDraftLibraryPageResult{FailureCode: failureCode}
	}
	filter, failureCode := savedWorkflowDraftRepositoryLibraryFilter(request)
	if failureCode != "" {
		return ListWorkflowDraftLibraryPageResult{FailureCode: failureCode}
	}
	executor, ok := adapter.queryExecutor.(SavedWorkflowDraftLibraryRepositoryQueryExecutor)
	if !ok {
		return ListWorkflowDraftLibraryPageResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	queryResult := executor.ListWorkflowDraftLibraryPage(
		ctx,
		savedWorkflowDraftRepositoryLibraryPageQuery{
			ActorContext: actor,
			Filter:       filter,
		},
	)
	if queryResult.FailureCode != "" {
		return ListWorkflowDraftLibraryPageResult{FailureCode: queryResult.FailureCode}
	}
	if len(queryResult.Records) > filter.Limit || queryResult.HasMore && len(queryResult.Records) == 0 {
		return ListWorkflowDraftLibraryPageResult{
			FailureCode: SavedWorkflowDraftFailureLifecycleStoreContract,
		}
	}
	summaries := make([]SavedWorkflowDraftSummary, 0, len(queryResult.Records))
	for _, record := range queryResult.Records {
		if failureCode := adapter.validateStoredRecord(actor, record); failureCode != "" {
			return ListWorkflowDraftLibraryPageResult{FailureCode: failureCode}
		}
		normalized, ok := normalizeAndValidateSavedWorkflowDraftLifecycle(record.Draft)
		if !ok ||
			!savedWorkflowDraftMatchesLibraryFilter(normalized, filter) ||
			!savedWorkflowDraftIsAfterLibraryCursor(normalized, filter) {
			return ListWorkflowDraftLibraryPageResult{
				FailureCode: SavedWorkflowDraftFailureLifecycleStoreContract,
			}
		}
		summaries = append(summaries, savedWorkflowDraftSummaryFromDraft(normalized))
	}
	return ListWorkflowDraftLibraryPageResult{
		Summaries: summaries,
		HasMore:   queryResult.HasMore,
	}
}

func (adapter SavedWorkflowDraftRepositoryAdapter) TransitionWorkflowDraftLifecycle(
	ctx context.Context,
	actor SavedWorkflowDraftRepositoryActorContext,
	request TransitionWorkflowDraftLifecycleRecordRequest,
) TransitionWorkflowDraftLifecycleRecordResult {
	if ctx == nil {
		return TransitionWorkflowDraftLifecycleRecordResult{
			FailureCode: SavedWorkflowDraftFailureAuthContextMismatch,
		}
	}
	actor = normalizeSavedWorkflowDraftRepositoryActorContext(actor)
	if failureCode := savedWorkflowDraftRepositoryActorFailure(
		actor,
		"workflow_drafts:write",
	); failureCode != "" {
		return TransitionWorkflowDraftLifecycleRecordResult{FailureCode: failureCode}
	}
	if failureCode := adapter.schemaPreflight.failureCodeFor(
		savedWorkflowDraftSchemaVersion,
	); failureCode != "" {
		return TransitionWorkflowDraftLifecycleRecordResult{FailureCode: failureCode}
	}
	request.DraftID = strings.TrimSpace(request.DraftID)
	request.OccurredAt = request.OccurredAt.UTC()
	if request.DraftID == "" ||
		request.ExpectedDraftVersion < 1 ||
		request.ExpectedLifecycleVersion < 1 ||
		request.OccurredAt.IsZero() ||
		!validSavedWorkflowDraftLifecycleState(request.TargetState) {
		return TransitionWorkflowDraftLifecycleRecordResult{
			FailureCode: SavedWorkflowDraftFailurePayloadInvalid,
		}
	}
	executor, ok := adapter.queryExecutor.(SavedWorkflowDraftLibraryRepositoryQueryExecutor)
	if !ok {
		return TransitionWorkflowDraftLifecycleRecordResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	queryResult := executor.TransitionWorkflowDraftLifecycle(
		ctx,
		savedWorkflowDraftRepositoryLifecycleTransitionQuery{
			ActorContext:             actor,
			DraftID:                  request.DraftID,
			TargetState:              request.TargetState,
			ExpectedDraftVersion:     request.ExpectedDraftVersion,
			ExpectedLifecycleVersion: request.ExpectedLifecycleVersion,
			OccurredAt:               request.OccurredAt,
		},
	)
	result := TransitionWorkflowDraftLifecycleRecordResult{
		FailureCode:             queryResult.FailureCode,
		CurrentDraftVersion:     queryResult.CurrentDraftVersion,
		CurrentLifecycleVersion: queryResult.CurrentLifecycleVersion,
		CurrentLifecycleState:   queryResult.CurrentLifecycleState,
	}
	if queryResult.FailureCode != "" {
		return result
	}
	if failureCode := adapter.validateStoredRecord(actor, queryResult.Record); failureCode != "" {
		result.FailureCode = failureCode
		return result
	}
	if !validSavedWorkflowDraftLifecycleTransition(
		SavedWorkflowDraftContext{
			RequestID:       actor.RequestID,
			TenantRef:       actor.TenantRef,
			WorkspaceID:     actor.WorkspaceID,
			ApplicationID:   actor.ApplicationID,
			ActorRef:        actor.ActorSubjectRef,
			OwnerSubjectRef: actor.OwnerSubjectRef,
			AuditRef:        actor.AuditRef,
		},
		queryResult.Record.Draft,
		queryResult.Event,
		request.TargetState,
		request.ExpectedDraftVersion,
		request.ExpectedLifecycleVersion,
	) {
		result.FailureCode = SavedWorkflowDraftFailureLifecycleStoreContract
		return result
	}
	result.Draft = cloneSavedWorkflowDraftPointer(queryResult.Record.Draft)
	result.Event = cloneSavedWorkflowDraftLifecycleEventPointer(queryResult.Event)
	return result
}

func savedWorkflowDraftRepositoryLibraryFilter(
	request ListWorkflowDraftLibraryPageRequest,
) (savedWorkflowDraftLibraryListFilter, SavedWorkflowDraftFailureCode) {
	filter := savedWorkflowDraftLibraryListFilter{
		LifecycleState:  request.LifecycleState,
		Limit:           request.Limit,
		NamePrefix:      strings.TrimSpace(request.NamePrefix),
		ValidationState: request.ValidationState,
		ProvenanceKind:  request.ProvenanceKind,
		AfterUpdatedAt:  strings.TrimSpace(request.AfterLibraryUpdatedAt),
		AfterDraftID:    strings.TrimSpace(request.AfterDraftID),
	}
	if !validSavedWorkflowDraftLifecycleState(filter.LifecycleState) ||
		filter.Limit < 1 ||
		filter.Limit > maxSavedWorkflowDraftListLimit ||
		!utf8.ValidString(filter.NamePrefix) ||
		utf8.RuneCountInString(filter.NamePrefix) > maxSavedWorkflowDraftNamePrefix ||
		!validSavedWorkflowDraftValidationFilter(filter.ValidationState) ||
		!validSavedWorkflowDraftProvenanceFilter(filter.ProvenanceKind) ||
		filter.NamePrefix != request.NamePrefix ||
		filter.AfterUpdatedAt != request.AfterLibraryUpdatedAt ||
		filter.AfterDraftID != request.AfterDraftID {
		return savedWorkflowDraftLibraryListFilter{}, SavedWorkflowDraftFailureListFilterInvalid
	}
	if (filter.AfterUpdatedAt == "") != (filter.AfterDraftID == "") {
		return savedWorkflowDraftLibraryListFilter{}, SavedWorkflowDraftFailureListCursorInvalid
	}
	if filter.AfterUpdatedAt != "" {
		parsed, err := time.Parse(time.RFC3339Nano, filter.AfterUpdatedAt)
		if err != nil || parsed.UTC().Format(time.RFC3339Nano) != filter.AfterUpdatedAt {
			return savedWorkflowDraftLibraryListFilter{}, SavedWorkflowDraftFailureListCursorInvalid
		}
	}
	return filter, ""
}

func (store *repositorySavedWorkflowDraftLibraryStore) ListDraftLibraryPage(
	requestContext SavedWorkflowDraftContext,
	filter savedWorkflowDraftLibraryListFilter,
) (savedWorkflowDraftLibraryPage, error) {
	actor, failureCode := repositoryActorFromSavedWorkflowDraftContext(requestContext)
	if failureCode != "" {
		return savedWorkflowDraftLibraryPage{}, savedWorkflowDraftStoreOperationFailure(failureCode)
	}
	if store == nil || store.repositorySavedWorkflowDraftStore == nil {
		return savedWorkflowDraftLibraryPage{}, savedWorkflowDraftStoreOperationFailure(
			SavedWorkflowDraftFailureStoreUnavailable,
		)
	}
	repository, ok := store.repository.(SavedWorkflowDraftLibraryRepository)
	if !ok {
		return savedWorkflowDraftLibraryPage{}, savedWorkflowDraftStoreOperationFailure(
			SavedWorkflowDraftFailureStoreUnavailable,
		)
	}
	result := repository.ListWorkflowDraftLibraryPage(
		requestContext.RequestContext,
		actor,
		ListWorkflowDraftLibraryPageRequest{
			LifecycleState:        filter.LifecycleState,
			Limit:                 filter.Limit,
			NamePrefix:            filter.NamePrefix,
			ValidationState:       filter.ValidationState,
			ProvenanceKind:        filter.ProvenanceKind,
			AfterLibraryUpdatedAt: filter.AfterUpdatedAt,
			AfterDraftID:          filter.AfterDraftID,
		},
	)
	if result.FailureCode != "" {
		return savedWorkflowDraftLibraryPage{}, savedWorkflowDraftStoreOperationFailure(result.FailureCode)
	}
	return savedWorkflowDraftLibraryPage{
		Summaries: append([]SavedWorkflowDraftSummary{}, result.Summaries...),
		HasMore:   result.HasMore,
	}, nil
}

func (store *repositorySavedWorkflowDraftLibraryStore) TransitionDraftLifecycle(
	requestContext SavedWorkflowDraftContext,
	draftID string,
	target SavedWorkflowDraftLifecycleState,
	expectedDraftVersion int,
	expectedLifecycleVersion int,
	now time.Time,
) (SavedWorkflowDraft, SavedWorkflowDraftLifecycleEvent, error) {
	actor, failureCode := repositoryActorFromSavedWorkflowDraftContext(requestContext)
	if failureCode != "" {
		return SavedWorkflowDraft{}, SavedWorkflowDraftLifecycleEvent{},
			savedWorkflowDraftStoreOperationFailure(failureCode)
	}
	if store == nil || store.repositorySavedWorkflowDraftStore == nil {
		return SavedWorkflowDraft{}, SavedWorkflowDraftLifecycleEvent{},
			savedWorkflowDraftStoreOperationFailure(SavedWorkflowDraftFailureStoreUnavailable)
	}
	repository, ok := store.repository.(SavedWorkflowDraftLibraryRepository)
	if !ok {
		return SavedWorkflowDraft{}, SavedWorkflowDraftLifecycleEvent{},
			savedWorkflowDraftStoreOperationFailure(SavedWorkflowDraftFailureStoreUnavailable)
	}
	result := repository.TransitionWorkflowDraftLifecycle(
		requestContext.RequestContext,
		actor,
		TransitionWorkflowDraftLifecycleRecordRequest{
			DraftID:                  draftID,
			TargetState:              target,
			ExpectedDraftVersion:     expectedDraftVersion,
			ExpectedLifecycleVersion: expectedLifecycleVersion,
			OccurredAt:               now,
		},
	)
	if result.FailureCode != "" {
		current := SavedWorkflowDraft{
			DraftVersion:     result.CurrentDraftVersion,
			LifecycleVersion: result.CurrentLifecycleVersion,
			LifecycleState:   result.CurrentLifecycleState,
		}
		return current, SavedWorkflowDraftLifecycleEvent{},
			savedWorkflowDraftStoreOperationFailure(result.FailureCode)
	}
	if result.Draft == nil || result.Event == nil {
		return SavedWorkflowDraft{}, SavedWorkflowDraftLifecycleEvent{},
			savedWorkflowDraftStoreOperationFailure(SavedWorkflowDraftFailureStoreContractMismatch)
	}
	store.mu.Lock()
	store.sideEffects.LifecycleTransitionCount++
	store.sideEffects.LifecycleEventWriteCount++
	store.sideEffects.ExternalRepositoryWrites++
	store.mu.Unlock()
	return cloneSavedWorkflowDraft(*result.Draft), *cloneSavedWorkflowDraftLifecycleEventPointer(*result.Event), nil
}

var _ SavedWorkflowDraftLibraryRepository = SavedWorkflowDraftRepositoryAdapter{}
var _ savedWorkflowDraftLibraryStore = (*repositorySavedWorkflowDraftLibraryStore)(nil)
