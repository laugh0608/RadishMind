package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

func (executor *sqliteSavedWorkflowDraftQueryExecutor) ListWorkflowDraftLibraryPage(
	ctx context.Context,
	query savedWorkflowDraftRepositoryLibraryPageQuery,
) savedWorkflowDraftRepositoryQueryLibraryPageResult {
	if executor == nil || executor.database == nil || ctx == nil {
		return savedWorkflowDraftRepositoryQueryLibraryPageResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	afterUnixNano := int64(0)
	hasAnchor := 0
	if query.Filter.AfterUpdatedAt != "" {
		after, err := time.Parse(time.RFC3339Nano, query.Filter.AfterUpdatedAt)
		if err != nil {
			return savedWorkflowDraftRepositoryQueryLibraryPageResult{
				FailureCode: SavedWorkflowDraftFailureListCursorInvalid,
			}
		}
		afterUnixNano, err = savedWorkflowDraftUnixNano(after)
		if err != nil {
			return savedWorkflowDraftRepositoryQueryLibraryPageResult{
				FailureCode: SavedWorkflowDraftFailureListCursorInvalid,
			}
		}
		hasAnchor = 1
	}
	rows, err := executor.database.QueryContext(
		ctx,
		sqliteSavedWorkflowDraftLibraryPageSQL,
		query.ActorContext.TenantRef,
		query.ActorContext.WorkspaceID,
		query.ActorContext.ApplicationID,
		query.ActorContext.OwnerSubjectRef,
		query.Filter.LifecycleState,
		query.Filter.NamePrefix,
		query.Filter.NamePrefix,
		query.Filter.ValidationState,
		query.Filter.ValidationState,
		query.Filter.ProvenanceKind,
		query.Filter.ProvenanceKind,
		hasAnchor,
		afterUnixNano,
		afterUnixNano,
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
		record, scanErr := scanSQLiteSavedWorkflowDraftRecord(rows)
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

func (executor *sqliteSavedWorkflowDraftQueryExecutor) TransitionWorkflowDraftLifecycle(
	ctx context.Context,
	query savedWorkflowDraftRepositoryLifecycleTransitionQuery,
) savedWorkflowDraftRepositoryQueryLifecycleTransitionResult {
	if executor == nil || executor.database == nil || ctx == nil {
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
	occurredAtUnixNano, err := savedWorkflowDraftUnixNano(query.OccurredAt)
	if err != nil {
		return savedWorkflowDraftRepositoryQueryLifecycleTransitionResult{
			FailureCode: SavedWorkflowDraftFailurePayloadInvalid,
		}
	}
	archivedAtUnixNano := any(nil)
	if archivedAt != nil {
		archivedAtUnixNano = occurredAtUnixNano
	}
	transaction, err := executor.database.BeginTx(ctx, nil)
	if err != nil {
		return savedWorkflowDraftRepositoryQueryLifecycleTransitionResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	defer func() { _ = transaction.Rollback() }()

	record, err := scanSQLiteSavedWorkflowDraftRecord(transaction.QueryRowContext(
		ctx,
		sqliteSavedWorkflowDraftLifecycleTransitionSQL,
		query.TargetState,
		archivedAtUnixNano,
		occurredAtUnixNano,
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
	if errors.Is(err, sql.ErrNoRows) {
		_ = transaction.Rollback()
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
	if _, err := transaction.ExecContext(
		ctx,
		sqliteSavedWorkflowDraftLifecycleEventInsertSQL,
		event.TenantRef,
		event.WorkspaceID,
		event.ApplicationID,
		event.DraftID,
		event.OwnerSubjectRef,
		event.LifecycleVersion,
		event.FromState,
		event.ToState,
		event.TransitionKind,
		occurredAtUnixNano,
		event.ActorRef,
		event.RequestID,
		event.AuditRef,
	); err != nil {
		return savedWorkflowDraftRepositoryQueryLifecycleTransitionResult{
			FailureCode: SavedWorkflowDraftFailureLifecycleEventWrite,
		}
	}
	if err := transaction.Commit(); err != nil {
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

func (executor *sqliteSavedWorkflowDraftQueryExecutor) failedLifecycleTransitionResult(
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

type savedWorkflowDraftCurrentState struct {
	DraftVersion     int
	LifecycleVersion int
	LifecycleState   SavedWorkflowDraftLifecycleState
	OwnerSubjectRef  string
}

func (executor *sqliteSavedWorkflowDraftQueryExecutor) currentDraftState(
	ctx context.Context,
	actor SavedWorkflowDraftRepositoryActorContext,
	draftID string,
) (savedWorkflowDraftCurrentState, bool, bool) {
	state := savedWorkflowDraftCurrentState{}
	err := executor.database.QueryRowContext(
		ctx,
		sqliteSavedWorkflowDraftCurrentStateSQL,
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
	if errors.Is(err, sql.ErrNoRows) {
		return savedWorkflowDraftCurrentState{}, false, false
	}
	if err != nil {
		return savedWorkflowDraftCurrentState{}, false, true
	}
	return state, true, false
}

func savedWorkflowDraftTransitionValues(
	target SavedWorkflowDraftLifecycleState,
	occurredAt time.Time,
) (
	SavedWorkflowDraftLifecycleState,
	SavedWorkflowDraftLifecycleTransitionKind,
	any,
	bool,
) {
	switch target {
	case SavedWorkflowDraftLifecycleArchived:
		return SavedWorkflowDraftLifecycleActive,
			SavedWorkflowDraftLifecycleTransitionArchived,
			occurredAt.UTC(),
			true
	case SavedWorkflowDraftLifecycleActive:
		return SavedWorkflowDraftLifecycleArchived,
			SavedWorkflowDraftLifecycleTransitionUnarchived,
			nil,
			true
	default:
		return "", "", nil, false
	}
}

func savedWorkflowDraftLifecycleEventFromTransition(
	query savedWorkflowDraftRepositoryLifecycleTransitionQuery,
	lifecycleVersion int,
	fromState SavedWorkflowDraftLifecycleState,
	transitionKind SavedWorkflowDraftLifecycleTransitionKind,
) SavedWorkflowDraftLifecycleEvent {
	return SavedWorkflowDraftLifecycleEvent{
		SchemaVersion:    savedWorkflowDraftLifecycleEventSchemaVersion,
		TenantRef:        query.ActorContext.TenantRef,
		WorkspaceID:      query.ActorContext.WorkspaceID,
		ApplicationID:    query.ActorContext.ApplicationID,
		OwnerSubjectRef:  query.ActorContext.OwnerSubjectRef,
		DraftID:          query.DraftID,
		LifecycleVersion: lifecycleVersion,
		FromState:        fromState,
		ToState:          query.TargetState,
		TransitionKind:   transitionKind,
		OccurredAt:       query.OccurredAt.UTC().Format(time.RFC3339Nano),
		ActorRef:         query.ActorContext.ActorSubjectRef,
		RequestID:        query.ActorContext.RequestID,
		AuditRef:         query.ActorContext.AuditRef,
	}
}

const sqliteSavedWorkflowDraftLibraryPageSQL = `
SELECT ` + sqliteSavedWorkflowDraftReturningColumns + `
  FROM saved_workflow_drafts
 WHERE tenant_ref=?
   AND workspace_id=?
   AND application_id=?
   AND owner_subject_ref=?
   AND lifecycle_state=?
   AND substr(draft_name, 1, length(?))=?
   AND (?='' OR validation_state=?)
   AND (?='' OR provenance_kind=?)
   AND (
       ?=0 OR
       library_updated_at_unix_nano < ? OR
       (library_updated_at_unix_nano = ? AND draft_id > ?)
   )
 ORDER BY library_updated_at_unix_nano DESC, draft_id ASC
 LIMIT ?`

const sqliteSavedWorkflowDraftLifecycleTransitionSQL = `
UPDATE saved_workflow_drafts
   SET lifecycle_state=?,
       lifecycle_version=lifecycle_version+1,
       archived_at_unix_nano=?,
       library_updated_at_unix_nano=?,
       lifecycle_updated_by_actor_ref=?,
       request_id=?,
       audit_ref=?
 WHERE tenant_ref=?
   AND workspace_id=?
   AND application_id=?
   AND draft_id=?
   AND owner_subject_ref=?
   AND draft_version=?
   AND lifecycle_version=?
   AND lifecycle_state=?
RETURNING ` + sqliteSavedWorkflowDraftReturningColumns

const sqliteSavedWorkflowDraftLifecycleEventInsertSQL = `
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
    occurred_at_unix_nano,
    actor_ref,
    request_id,
    audit_ref
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`

const sqliteSavedWorkflowDraftCurrentStateSQL = `
SELECT draft_version, lifecycle_version, lifecycle_state, owner_subject_ref
  FROM saved_workflow_drafts
 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND draft_id=?`

var _ SavedWorkflowDraftLibraryRepositoryQueryExecutor = (*sqliteSavedWorkflowDraftQueryExecutor)(nil)
