package httpapi

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

func (executor *postgresSavedWorkflowDraftQueryExecutor) ListWorkflowDraftLibraryPage(
	ctx context.Context,
	query savedWorkflowDraftRepositoryLibraryPageQuery,
) savedWorkflowDraftRepositoryQueryLibraryPageResult {
	if executor == nil || executor.pool == nil || ctx == nil {
		return savedWorkflowDraftRepositoryQueryLibraryPageResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	var after time.Time
	hasAnchor := false
	if query.Filter.AfterUpdatedAt != "" {
		parsed, err := time.Parse(time.RFC3339Nano, query.Filter.AfterUpdatedAt)
		if err != nil {
			return savedWorkflowDraftRepositoryQueryLibraryPageResult{
				FailureCode: SavedWorkflowDraftFailureListCursorInvalid,
			}
		}
		after = parsed.UTC()
		hasAnchor = true
	}
	rows, err := executor.pool.Query(
		ctx,
		postgresSavedWorkflowDraftLibraryPageSQL,
		query.ActorContext.TenantRef,
		query.ActorContext.WorkspaceID,
		query.ActorContext.ApplicationID,
		query.ActorContext.OwnerSubjectRef,
		query.Filter.LifecycleState,
		query.Filter.NamePrefix,
		query.Filter.ValidationState,
		query.Filter.ProvenanceKind,
		hasAnchor,
		after,
		query.Filter.AfterDraftID,
		query.Filter.Limit+1,
	)
	if err != nil {
		return savedWorkflowDraftRepositoryQueryLibraryPageResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	defer rows.Close()

	records := make([]SavedWorkflowDraftRepositoryStoredRecord, 0, query.Filter.Limit+1)
	for rows.Next() {
		record, scanErr := scanPostgresSavedWorkflowDraftRecord(rows)
		if scanErr != nil {
			failureCode := SavedWorkflowDraftFailureStoreUnavailable
			if errors.Is(scanErr, errSavedWorkflowDraftStoredLibraryProjection) {
				failureCode = SavedWorkflowDraftFailureLifecycleStoreContract
			} else if errors.Is(scanErr, errSavedWorkflowDraftStoredRecordContract) {
				failureCode = SavedWorkflowDraftFailureStoreContractMismatch
			}
			return savedWorkflowDraftRepositoryQueryLibraryPageResult{FailureCode: failureCode}
		}
		records = append(records, record)
	}
	if rows.Err() != nil {
		return savedWorkflowDraftRepositoryQueryLibraryPageResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	hasMore := len(records) > query.Filter.Limit
	if hasMore {
		records = records[:query.Filter.Limit]
	}
	return savedWorkflowDraftRepositoryQueryLibraryPageResult{
		Records: records,
		HasMore: hasMore,
	}
}

func (executor *postgresSavedWorkflowDraftQueryExecutor) TransitionWorkflowDraftLifecycle(
	ctx context.Context,
	query savedWorkflowDraftRepositoryLifecycleTransitionQuery,
) savedWorkflowDraftRepositoryQueryLifecycleTransitionResult {
	if executor == nil || executor.pool == nil || ctx == nil {
		return savedWorkflowDraftRepositoryQueryLifecycleTransitionResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	fromState, transitionKind, archivedAt, valid := savedWorkflowDraftTransitionValues(
		query.TargetState,
		query.OccurredAt,
	)
	if !valid {
		return savedWorkflowDraftRepositoryQueryLifecycleTransitionResult{
			FailureCode: SavedWorkflowDraftFailurePayloadInvalid,
		}
	}
	transaction, err := executor.pool.Begin(ctx)
	if err != nil {
		return savedWorkflowDraftRepositoryQueryLifecycleTransitionResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()

	record, err := scanPostgresSavedWorkflowDraftRecord(transaction.QueryRow(
		ctx,
		postgresSavedWorkflowDraftLifecycleTransitionSQL,
		query.TargetState,
		archivedAt,
		query.OccurredAt.UTC(),
		query.ActorContext.ActorSubjectRef,
		query.ActorContext.RequestID,
		query.ActorContext.AuditRef,
		query.ActorContext.TenantRef,
		query.ActorContext.WorkspaceID,
		query.ActorContext.ApplicationID,
		query.DraftID,
		query.ActorContext.OwnerSubjectRef,
		query.ExpectedDraftVersion,
		query.ExpectedLifecycleVersion,
		fromState,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		_ = transaction.Rollback(ctx)
		return executor.failedLifecycleTransitionResult(ctx, query)
	}
	if err != nil {
		failureCode := SavedWorkflowDraftFailureStoreUnavailable
		if errors.Is(err, errSavedWorkflowDraftStoredLibraryProjection) {
			failureCode = SavedWorkflowDraftFailureLifecycleStoreContract
		} else if errors.Is(err, errSavedWorkflowDraftStoredRecordContract) {
			failureCode = SavedWorkflowDraftFailureStoreContractMismatch
		}
		return savedWorkflowDraftRepositoryQueryLifecycleTransitionResult{FailureCode: failureCode}
	}
	event := savedWorkflowDraftLifecycleEventFromTransition(
		query,
		record.Draft.LifecycleVersion,
		fromState,
		transitionKind,
	)
	if _, err := transaction.Exec(
		ctx,
		postgresSavedWorkflowDraftLifecycleEventInsertSQL,
		event.TenantRef,
		event.WorkspaceID,
		event.ApplicationID,
		event.DraftID,
		event.OwnerSubjectRef,
		event.LifecycleVersion,
		event.FromState,
		event.ToState,
		event.TransitionKind,
		query.OccurredAt.UTC(),
		event.ActorRef,
		event.RequestID,
		event.AuditRef,
	); err != nil {
		return savedWorkflowDraftRepositoryQueryLifecycleTransitionResult{
			FailureCode: SavedWorkflowDraftFailureLifecycleEventWrite,
		}
	}
	if err := transaction.Commit(ctx); err != nil {
		return savedWorkflowDraftRepositoryQueryLifecycleTransitionResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	return savedWorkflowDraftRepositoryQueryLifecycleTransitionResult{
		Record:                  record,
		Event:                   event,
		CurrentDraftVersion:     record.Draft.DraftVersion,
		CurrentLifecycleVersion: record.Draft.LifecycleVersion,
		CurrentLifecycleState:   record.Draft.LifecycleState,
	}
}

func (executor *postgresSavedWorkflowDraftQueryExecutor) failedLifecycleTransitionResult(
	ctx context.Context,
	query savedWorkflowDraftRepositoryLifecycleTransitionQuery,
) savedWorkflowDraftRepositoryQueryLifecycleTransitionResult {
	state, found, failed := executor.currentDraftState(ctx, query.ActorContext, query.DraftID)
	if failed {
		return savedWorkflowDraftRepositoryQueryLifecycleTransitionResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	result := savedWorkflowDraftRepositoryQueryLifecycleTransitionResult{
		CurrentDraftVersion:     state.DraftVersion,
		CurrentLifecycleVersion: state.LifecycleVersion,
		CurrentLifecycleState:   state.LifecycleState,
	}
	if !found {
		result.FailureCode = SavedWorkflowDraftFailureNotFound
	} else if state.OwnerSubjectRef != query.ActorContext.OwnerSubjectRef {
		result.FailureCode = SavedWorkflowDraftFailureScopeDenied
	} else if state.DraftVersion != query.ExpectedDraftVersion {
		result.FailureCode = SavedWorkflowDraftFailureVersionConflict
	} else if state.LifecycleVersion != query.ExpectedLifecycleVersion {
		result.FailureCode = SavedWorkflowDraftFailureLifecycleVersionConflict
	} else {
		result.FailureCode = SavedWorkflowDraftFailureLifecycleStateConflict
	}
	return result
}

func (executor *postgresSavedWorkflowDraftQueryExecutor) currentDraftState(
	ctx context.Context,
	actor SavedWorkflowDraftRepositoryActorContext,
	draftID string,
) (savedWorkflowDraftCurrentState, bool, bool) {
	state := savedWorkflowDraftCurrentState{}
	err := executor.pool.QueryRow(
		ctx,
		postgresSavedWorkflowDraftCurrentStateSQL,
		actor.TenantRef,
		actor.WorkspaceID,
		actor.ApplicationID,
		draftID,
	).Scan(
		&state.DraftVersion,
		&state.LifecycleVersion,
		&state.LifecycleState,
		&state.OwnerSubjectRef,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return savedWorkflowDraftCurrentState{}, false, false
	}
	if err != nil {
		return savedWorkflowDraftCurrentState{}, false, true
	}
	return state, true, false
}

const postgresSavedWorkflowDraftLibraryPageSQL = `
SELECT ` + postgresSavedWorkflowDraftReturningColumns + `
  FROM saved_workflow_drafts
 WHERE tenant_ref = $1
   AND workspace_id = $2
   AND application_id = $3
   AND owner_subject_ref = $4
   AND lifecycle_state = $5
   AND left(draft_name, char_length($6)) = $6
   AND ($7 = '' OR validation_state = $7)
   AND ($8 = '' OR provenance_kind = $8)
   AND (
       NOT $9 OR
       library_updated_at < $10 OR
       (library_updated_at = $10 AND draft_id > $11)
   )
 ORDER BY library_updated_at DESC, draft_id ASC
 LIMIT $12`

const postgresSavedWorkflowDraftLifecycleTransitionSQL = `
UPDATE saved_workflow_drafts
   SET lifecycle_state = $1,
       lifecycle_version = lifecycle_version + 1,
       archived_at = $2,
       library_updated_at = $3,
       lifecycle_updated_by_actor_ref = $4,
       request_id = $5,
       audit_ref = $6
 WHERE tenant_ref = $7
   AND workspace_id = $8
   AND application_id = $9
   AND draft_id = $10
   AND owner_subject_ref = $11
   AND draft_version = $12
   AND lifecycle_version = $13
   AND lifecycle_state = $14
RETURNING ` + postgresSavedWorkflowDraftReturningColumns

const postgresSavedWorkflowDraftLifecycleEventInsertSQL = `
INSERT INTO saved_workflow_draft_lifecycle_events (
    tenant_ref,
    workspace_id,
    application_id,
    draft_id,
    owner_subject_ref,
    lifecycle_version,
    from_state,
    to_state,
    transition_kind,
    occurred_at,
    actor_ref,
    request_id,
    audit_ref
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)`

const postgresSavedWorkflowDraftCurrentStateSQL = `
SELECT draft_version, lifecycle_version, lifecycle_state, owner_subject_ref
  FROM saved_workflow_drafts
 WHERE tenant_ref = $1
   AND workspace_id = $2
   AND application_id = $3
   AND draft_id = $4`

var _ SavedWorkflowDraftLibraryRepositoryQueryExecutor = (*postgresSavedWorkflowDraftQueryExecutor)(nil)
