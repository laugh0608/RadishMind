package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type sqliteSavedWorkflowDraftQueryExecutor struct {
	database *sql.DB
}

type sqliteSavedWorkflowDraftRow interface {
	Scan(...any) error
}

func newSQLiteSavedWorkflowDraftQueryExecutor(database *sql.DB) *sqliteSavedWorkflowDraftQueryExecutor {
	return &sqliteSavedWorkflowDraftQueryExecutor{database: database}
}

func newSQLiteSavedWorkflowDraftStore(database *sql.DB) *repositorySavedWorkflowDraftStore {
	executor := newSQLiteSavedWorkflowDraftQueryExecutor(database)
	repository := NewSavedWorkflowDraftRepositoryAdapter(SavedWorkflowDraftRepositoryAdapterConfig{
		QueryExecutor: executor,
		SchemaPreflight: SavedWorkflowDraftRepositorySchemaPreflight{
			StoreSchemaVersion: savedWorkflowDraftRepositoryStoreSchemaVersion,
			MigrationState:     savedWorkflowDraftRepositoryMigrationApplied,
		},
	})
	return newRepositorySavedWorkflowDraftStore(repository)
}

func (executor *sqliteSavedWorkflowDraftQueryExecutor) SaveWorkflowDraftRecord(
	ctx context.Context,
	query savedWorkflowDraftRepositorySaveQuery,
) savedWorkflowDraftRepositoryQuerySaveResult {
	if executor == nil || executor.database == nil || ctx == nil {
		return savedWorkflowDraftRepositoryQuerySaveResult{FailureCode: SavedWorkflowDraftFailureStoreUnavailable}
	}
	if query.ExpectedDraftVersion < 0 || query.Record.Draft.DraftVersion != query.ExpectedDraftVersion+1 {
		return savedWorkflowDraftRepositoryQuerySaveResult{FailureCode: SavedWorkflowDraftFailureStoreContractMismatch}
	}
	if !validSavedWorkflowDraftRevisionWriteMetadata(query.RevisionKind, query.RestoredFromVersion) {
		return savedWorkflowDraftRepositoryQuerySaveResult{FailureCode: SavedWorkflowDraftFailureStoreContractMismatch}
	}
	payload, validation, blocked, createdAt, updatedAt, failureCode := savedWorkflowDraftRecordValues(query.Record)
	if failureCode != "" {
		return savedWorkflowDraftRepositoryQuerySaveResult{FailureCode: failureCode}
	}
	createdAtUnixNano, err := savedWorkflowDraftUnixNano(createdAt)
	if err != nil {
		return savedWorkflowDraftRepositoryQuerySaveResult{FailureCode: SavedWorkflowDraftFailureStoreContractMismatch}
	}
	updatedAtUnixNano, err := savedWorkflowDraftUnixNano(updatedAt)
	if err != nil {
		return savedWorkflowDraftRepositoryQuerySaveResult{FailureCode: SavedWorkflowDraftFailureStoreContractMismatch}
	}
	libraryUpdatedAt, err := time.Parse(time.RFC3339Nano, query.Record.Draft.LibraryUpdatedAt)
	if err != nil || !libraryUpdatedAt.Equal(updatedAt) {
		return savedWorkflowDraftRepositoryQuerySaveResult{
			FailureCode: SavedWorkflowDraftFailureLifecycleStoreContract,
		}
	}
	libraryUpdatedAtUnixNano, err := savedWorkflowDraftUnixNano(libraryUpdatedAt)
	if err != nil {
		return savedWorkflowDraftRepositoryQuerySaveResult{
			FailureCode: SavedWorkflowDraftFailureLifecycleStoreContract,
		}
	}

	transaction, err := executor.database.BeginTx(ctx, nil)
	if err != nil {
		return savedWorkflowDraftRepositoryQuerySaveResult{FailureCode: SavedWorkflowDraftFailureStoreUnavailable}
	}
	defer func() { _ = transaction.Rollback() }()

	var row sqliteSavedWorkflowDraftRow
	if query.ExpectedDraftVersion == 0 {
		row = transaction.QueryRowContext(
			ctx,
			sqliteSavedWorkflowDraftInsertSQL,
			query.Record.TenantRef,
			query.Record.WorkspaceID,
			query.Record.ApplicationID,
			query.Record.DraftID,
			query.Record.OwnerSubjectRef,
			query.Record.StoreSchemaVersion,
			query.Record.Draft.SchemaVersion,
			query.Record.Draft.DraftVersion,
			query.Record.Draft.DraftStatus,
			query.Record.Draft.LifecycleState,
			query.Record.Draft.LifecycleVersion,
			nil,
			libraryUpdatedAtUnixNano,
			query.Record.Draft.LifecycleUpdatedByActorRef,
			query.Record.Draft.Name,
			query.Record.Draft.ValidationSummary.ValidationState,
			query.Record.Draft.ProvenanceKind,
			string(payload),
			string(validation),
			string(blocked),
			createdAtUnixNano,
			updatedAtUnixNano,
			query.Record.Draft.CreatedByActorRef,
			query.Record.Draft.UpdatedByActorRef,
			query.Record.Draft.RequestAuditMetadata.RequestID,
			query.Record.Draft.RequestAuditMetadata.AuditRef,
		)
	} else {
		row = transaction.QueryRowContext(
			ctx,
			sqliteSavedWorkflowDraftUpdateSQL,
			query.Record.StoreSchemaVersion,
			query.Record.Draft.SchemaVersion,
			query.Record.Draft.DraftVersion,
			query.Record.Draft.DraftStatus,
			query.Record.Draft.Name,
			query.Record.Draft.ValidationSummary.ValidationState,
			query.Record.Draft.ProvenanceKind,
			string(payload),
			string(validation),
			string(blocked),
			updatedAtUnixNano,
			libraryUpdatedAtUnixNano,
			query.Record.Draft.UpdatedByActorRef,
			query.Record.Draft.RequestAuditMetadata.RequestID,
			query.Record.Draft.RequestAuditMetadata.AuditRef,
			query.Record.TenantRef,
			query.Record.WorkspaceID,
			query.Record.ApplicationID,
			query.Record.DraftID,
			query.Record.OwnerSubjectRef,
			query.ExpectedDraftVersion,
			query.Record.Draft.LifecycleVersion,
			createdAtUnixNano,
		)
	}

	record, err := scanSQLiteSavedWorkflowDraftRecord(row)
	if err == nil {
		if _, insertErr := transaction.ExecContext(
			ctx,
			sqliteSavedWorkflowDraftRevisionInsertSQL,
			query.Record.TenantRef,
			query.Record.WorkspaceID,
			query.Record.ApplicationID,
			query.Record.DraftID,
			query.Record.OwnerSubjectRef,
			query.Record.Draft.DraftVersion,
			query.RevisionKind,
			query.RestoredFromVersion,
			string(payload),
		); insertErr != nil {
			return savedWorkflowDraftRepositoryQuerySaveResult{
				FailureCode: SavedWorkflowDraftFailureStoreContractMismatch,
			}
		}
		if commitErr := transaction.Commit(); commitErr != nil {
			return savedWorkflowDraftRepositoryQuerySaveResult{
				FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
			}
		}
		return savedWorkflowDraftRepositoryQuerySaveResult{
			Record:              record,
			CurrentDraftVersion: record.Draft.DraftVersion,
		}
	}
	if errors.Is(err, errSavedWorkflowDraftStoredLibraryProjection) {
		return savedWorkflowDraftRepositoryQuerySaveResult{
			FailureCode: SavedWorkflowDraftFailureLifecycleStoreContract,
		}
	}
	if errors.Is(err, errSavedWorkflowDraftStoredRecordContract) {
		return savedWorkflowDraftRepositoryQuerySaveResult{FailureCode: SavedWorkflowDraftFailureStoreContractMismatch}
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return savedWorkflowDraftRepositoryQuerySaveResult{FailureCode: SavedWorkflowDraftFailureStoreUnavailable}
	}
	_ = transaction.Rollback()
	return executor.failedCASResult(ctx, query)
}

func (executor *sqliteSavedWorkflowDraftQueryExecutor) ReadWorkflowDraftRecord(
	ctx context.Context,
	query savedWorkflowDraftRepositoryReadQuery,
) savedWorkflowDraftRepositoryQueryReadResult {
	if executor == nil || executor.database == nil || ctx == nil {
		return savedWorkflowDraftRepositoryQueryReadResult{FailureCode: SavedWorkflowDraftFailureStoreUnavailable}
	}
	record, err := scanSQLiteSavedWorkflowDraftRecord(executor.database.QueryRowContext(
		ctx,
		sqliteSavedWorkflowDraftReadSQL,
		query.ActorContext.TenantRef,
		query.ActorContext.WorkspaceID,
		query.ActorContext.ApplicationID,
		query.DraftID,
		query.ActorContext.OwnerSubjectRef,
	))
	if err == nil {
		return savedWorkflowDraftRepositoryQueryReadResult{
			Record:              record,
			CurrentDraftVersion: record.Draft.DraftVersion,
		}
	}
	if errors.Is(err, errSavedWorkflowDraftStoredLibraryProjection) {
		return savedWorkflowDraftRepositoryQueryReadResult{
			FailureCode: SavedWorkflowDraftFailureLifecycleStoreContract,
		}
	}
	if errors.Is(err, errSavedWorkflowDraftStoredRecordContract) {
		return savedWorkflowDraftRepositoryQueryReadResult{FailureCode: SavedWorkflowDraftFailureStoreContractMismatch}
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return savedWorkflowDraftRepositoryQueryReadResult{FailureCode: SavedWorkflowDraftFailureStoreUnavailable}
	}
	currentVersion, owner, found, lookupFailed := executor.currentVersionAndOwner(ctx, query.ActorContext, query.DraftID)
	if lookupFailed {
		return savedWorkflowDraftRepositoryQueryReadResult{FailureCode: SavedWorkflowDraftFailureStoreUnavailable}
	}
	if !found {
		return savedWorkflowDraftRepositoryQueryReadResult{FailureCode: SavedWorkflowDraftFailureNotFound}
	}
	if owner != query.ActorContext.OwnerSubjectRef {
		return savedWorkflowDraftRepositoryQueryReadResult{
			FailureCode:         SavedWorkflowDraftFailureScopeDenied,
			CurrentDraftVersion: currentVersion,
		}
	}
	return savedWorkflowDraftRepositoryQueryReadResult{
		FailureCode:         SavedWorkflowDraftFailureStoreContractMismatch,
		CurrentDraftVersion: currentVersion,
	}
}

func (executor *sqliteSavedWorkflowDraftQueryExecutor) ListWorkflowDraftRecords(
	ctx context.Context,
	query savedWorkflowDraftRepositoryListQuery,
) savedWorkflowDraftRepositoryQueryListResult {
	if executor == nil || executor.database == nil || ctx == nil {
		return savedWorkflowDraftRepositoryQueryListResult{FailureCode: SavedWorkflowDraftFailureStoreUnavailable}
	}
	rows, err := executor.database.QueryContext(
		ctx,
		sqliteSavedWorkflowDraftListSQL,
		query.ActorContext.TenantRef,
		query.ActorContext.WorkspaceID,
		query.ActorContext.ApplicationID,
		query.ActorContext.OwnerSubjectRef,
		savedWorkflowDraftRepositoryListLimit,
	)
	if err != nil {
		return savedWorkflowDraftRepositoryQueryListResult{FailureCode: SavedWorkflowDraftFailureStoreUnavailable}
	}
	defer rows.Close()

	records := make([]SavedWorkflowDraftRepositoryStoredRecord, 0)
	for rows.Next() {
		record, scanErr := scanSQLiteSavedWorkflowDraftRecord(rows)
		if scanErr != nil {
			failureCode := SavedWorkflowDraftFailureStoreUnavailable
			if errors.Is(scanErr, errSavedWorkflowDraftStoredLibraryProjection) {
				failureCode = SavedWorkflowDraftFailureLifecycleStoreContract
			} else if errors.Is(scanErr, errSavedWorkflowDraftStoredRecordContract) {
				failureCode = SavedWorkflowDraftFailureStoreContractMismatch
			}
			return savedWorkflowDraftRepositoryQueryListResult{FailureCode: failureCode}
		}
		records = append(records, record)
	}
	if rows.Err() != nil {
		return savedWorkflowDraftRepositoryQueryListResult{FailureCode: SavedWorkflowDraftFailureStoreUnavailable}
	}
	return savedWorkflowDraftRepositoryQueryListResult{Records: records}
}

func (executor *sqliteSavedWorkflowDraftQueryExecutor) ReadWorkflowDraftRevision(
	ctx context.Context,
	query savedWorkflowDraftRepositoryRevisionReadQuery,
) savedWorkflowDraftRepositoryQueryRevisionReadResult {
	if executor == nil || executor.database == nil || ctx == nil {
		return savedWorkflowDraftRepositoryQueryRevisionReadResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	revision, err := scanSQLiteSavedWorkflowDraftRevision(executor.database.QueryRowContext(
		ctx,
		sqliteSavedWorkflowDraftRevisionReadSQL,
		query.ActorContext.TenantRef,
		query.ActorContext.WorkspaceID,
		query.ActorContext.ApplicationID,
		query.DraftID,
		query.ActorContext.OwnerSubjectRef,
		query.DraftVersion,
	))
	if err == nil {
		return savedWorkflowDraftRepositoryQueryRevisionReadResult{Revision: revision}
	}
	if errors.Is(err, errSavedWorkflowDraftStoredRecordContract) {
		return savedWorkflowDraftRepositoryQueryRevisionReadResult{
			FailureCode: SavedWorkflowDraftFailureStoreContractMismatch,
		}
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return savedWorkflowDraftRepositoryQueryRevisionReadResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	_, owner, found, lookupFailed := executor.currentVersionAndOwner(
		ctx,
		query.ActorContext,
		query.DraftID,
	)
	if lookupFailed {
		return savedWorkflowDraftRepositoryQueryRevisionReadResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	if found && owner != query.ActorContext.OwnerSubjectRef {
		return savedWorkflowDraftRepositoryQueryRevisionReadResult{
			FailureCode: SavedWorkflowDraftFailureScopeDenied,
		}
	}
	return savedWorkflowDraftRepositoryQueryRevisionReadResult{
		FailureCode: SavedWorkflowDraftFailureRevisionNotFound,
	}
}

func (executor *sqliteSavedWorkflowDraftQueryExecutor) ListWorkflowDraftRevisions(
	ctx context.Context,
	query savedWorkflowDraftRepositoryRevisionListQuery,
) savedWorkflowDraftRepositoryQueryRevisionListResult {
	if executor == nil || executor.database == nil || ctx == nil {
		return savedWorkflowDraftRepositoryQueryRevisionListResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	_, owner, found, lookupFailed := executor.currentVersionAndOwner(
		ctx,
		query.ActorContext,
		query.DraftID,
	)
	if lookupFailed {
		return savedWorkflowDraftRepositoryQueryRevisionListResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	if !found {
		return savedWorkflowDraftRepositoryQueryRevisionListResult{
			FailureCode: SavedWorkflowDraftFailureNotFound,
		}
	}
	if owner != query.ActorContext.OwnerSubjectRef {
		return savedWorkflowDraftRepositoryQueryRevisionListResult{
			FailureCode: SavedWorkflowDraftFailureScopeDenied,
		}
	}
	rows, err := executor.database.QueryContext(
		ctx,
		sqliteSavedWorkflowDraftRevisionListSQL,
		query.ActorContext.TenantRef,
		query.ActorContext.WorkspaceID,
		query.ActorContext.ApplicationID,
		query.DraftID,
		query.ActorContext.OwnerSubjectRef,
		query.BeforeVersion,
		query.BeforeVersion,
		query.Limit+1,
	)
	if err != nil {
		return savedWorkflowDraftRepositoryQueryRevisionListResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	defer rows.Close()
	summaries := make([]SavedWorkflowDraftRevisionSummary, 0, query.Limit+1)
	for rows.Next() {
		revision, scanErr := scanSQLiteSavedWorkflowDraftRevision(rows)
		if scanErr != nil {
			return savedWorkflowDraftRepositoryQueryRevisionListResult{
				FailureCode: SavedWorkflowDraftFailureStoreContractMismatch,
			}
		}
		summaries = append(summaries, savedWorkflowDraftRevisionSummary(revision))
	}
	if rows.Err() != nil {
		return savedWorkflowDraftRepositoryQueryRevisionListResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	hasMore := len(summaries) > query.Limit
	if hasMore {
		summaries = summaries[:query.Limit]
	}
	return savedWorkflowDraftRepositoryQueryRevisionListResult{
		Revisions: summaries,
		HasMore:   hasMore,
	}
}

func (executor *sqliteSavedWorkflowDraftQueryExecutor) failedCASResult(
	ctx context.Context,
	query savedWorkflowDraftRepositorySaveQuery,
) savedWorkflowDraftRepositoryQuerySaveResult {
	state, found, lookupFailed := executor.currentDraftState(
		ctx,
		query.ActorContext,
		query.Record.DraftID,
	)
	if lookupFailed {
		return savedWorkflowDraftRepositoryQuerySaveResult{FailureCode: SavedWorkflowDraftFailureStoreUnavailable}
	}
	if !found {
		return savedWorkflowDraftRepositoryQuerySaveResult{FailureCode: SavedWorkflowDraftFailureNotFound}
	}
	if state.OwnerSubjectRef != query.Record.OwnerSubjectRef {
		return savedWorkflowDraftRepositoryQuerySaveResult{
			FailureCode:         SavedWorkflowDraftFailureScopeDenied,
			CurrentDraftVersion: state.DraftVersion,
		}
	}
	if state.LifecycleState != SavedWorkflowDraftLifecycleActive {
		return savedWorkflowDraftRepositoryQuerySaveResult{
			FailureCode:         SavedWorkflowDraftFailureArchived,
			CurrentDraftVersion: state.DraftVersion,
		}
	}
	if state.DraftVersion == query.ExpectedDraftVersion &&
		state.LifecycleVersion != query.Record.Draft.LifecycleVersion {
		return savedWorkflowDraftRepositoryQuerySaveResult{
			FailureCode:         SavedWorkflowDraftFailureLifecycleVersionConflict,
			CurrentDraftVersion: state.DraftVersion,
		}
	}
	return savedWorkflowDraftRepositoryQuerySaveResult{
		FailureCode:         SavedWorkflowDraftFailureVersionConflict,
		CurrentDraftVersion: state.DraftVersion,
	}
}

func (executor *sqliteSavedWorkflowDraftQueryExecutor) currentVersionAndOwner(
	ctx context.Context,
	actor SavedWorkflowDraftRepositoryActorContext,
	draftID string,
) (int, string, bool, bool) {
	state, found, failed := executor.currentDraftState(ctx, actor, draftID)
	return state.DraftVersion, state.OwnerSubjectRef, found, failed
}

func scanSQLiteSavedWorkflowDraftRecord(
	row sqliteSavedWorkflowDraftRow,
) (SavedWorkflowDraftRepositoryStoredRecord, error) {
	record := SavedWorkflowDraftRepositoryStoredRecord{}
	var payload []byte
	var draftVersion int
	var schemaVersion string
	var draftStatus string
	var lifecycleState string
	var lifecycleVersion int
	var archivedAtUnixNano sql.NullInt64
	var libraryUpdatedAtUnixNano int64
	var lifecycleUpdatedByActorRef string
	var draftName string
	var validationState string
	var provenanceKind string
	var createdAtUnixNano int64
	var updatedAtUnixNano int64
	if err := row.Scan(
		&record.TenantRef,
		&record.WorkspaceID,
		&record.ApplicationID,
		&record.DraftID,
		&record.OwnerSubjectRef,
		&record.StoreSchemaVersion,
		&schemaVersion,
		&draftVersion,
		&draftStatus,
		&lifecycleState,
		&lifecycleVersion,
		&archivedAtUnixNano,
		&libraryUpdatedAtUnixNano,
		&lifecycleUpdatedByActorRef,
		&draftName,
		&validationState,
		&provenanceKind,
		&payload,
		&createdAtUnixNano,
		&updatedAtUnixNano,
	); err != nil {
		return SavedWorkflowDraftRepositoryStoredRecord{}, err
	}
	decoded, err := decodeSavedWorkflowDraftStoredRecord(record, payload)
	if err != nil || decoded.Draft.SchemaVersion != schemaVersion ||
		decoded.Draft.DraftVersion != draftVersion || string(decoded.Draft.DraftStatus) != draftStatus {
		return SavedWorkflowDraftRepositoryStoredRecord{}, errSavedWorkflowDraftStoredRecordContract
	}
	createdAt, err := time.Parse(time.RFC3339, decoded.Draft.CreatedAt)
	if err != nil {
		return SavedWorkflowDraftRepositoryStoredRecord{}, errSavedWorkflowDraftStoredRecordContract
	}
	updatedAt, err := time.Parse(time.RFC3339, decoded.Draft.UpdatedAt)
	if err != nil || createdAt.UnixNano() != createdAtUnixNano || updatedAt.UnixNano() != updatedAtUnixNano {
		return SavedWorkflowDraftRepositoryStoredRecord{}, errSavedWorkflowDraftStoredRecordContract
	}
	archivedAt := ""
	if archivedAtUnixNano.Valid {
		archivedAt = time.Unix(0, archivedAtUnixNano.Int64).UTC().Format(time.RFC3339Nano)
	}
	libraryUpdatedAt := time.Unix(0, libraryUpdatedAtUnixNano).UTC().Format(time.RFC3339Nano)
	return applySavedWorkflowDraftStoredLibraryProjection(
		decoded,
		lifecycleState,
		lifecycleVersion,
		archivedAt,
		libraryUpdatedAt,
		lifecycleUpdatedByActorRef,
		draftName,
		validationState,
		provenanceKind,
	)
}

func scanSQLiteSavedWorkflowDraftRevision(
	row sqliteSavedWorkflowDraftRow,
) (SavedWorkflowDraftRevision, error) {
	record := SavedWorkflowDraftRepositoryStoredRecord{
		StoreSchemaVersion: savedWorkflowDraftRepositoryStoreSchemaVersion,
	}
	var draftVersion int
	var revisionKind string
	var restoredFromVersion int
	var payload []byte
	if err := row.Scan(
		&record.TenantRef,
		&record.WorkspaceID,
		&record.ApplicationID,
		&record.DraftID,
		&record.OwnerSubjectRef,
		&draftVersion,
		&revisionKind,
		&restoredFromVersion,
		&payload,
	); err != nil {
		return SavedWorkflowDraftRevision{}, err
	}
	decoded, err := decodeSavedWorkflowDraftStoredRecord(record, payload)
	if err != nil || decoded.Draft.DraftVersion != draftVersion {
		return SavedWorkflowDraftRevision{}, errSavedWorkflowDraftStoredRecordContract
	}
	revision := SavedWorkflowDraftRevision{
		SchemaVersion:       savedWorkflowDraftRevisionSchemaVersion,
		Draft:               decoded.Draft,
		RevisionKind:        SavedWorkflowDraftRevisionKind(revisionKind),
		RestoredFromVersion: restoredFromVersion,
	}
	if failure := validateSavedWorkflowDraftRevisionScope(
		SavedWorkflowDraftContext{
			WorkspaceID:   record.WorkspaceID,
			ApplicationID: record.ApplicationID,
		},
		revision,
		record.DraftID,
	); failure != "" {
		return SavedWorkflowDraftRevision{}, errSavedWorkflowDraftStoredRecordContract
	}
	return revision, nil
}

const sqliteSavedWorkflowDraftReturningColumns = `
    tenant_ref,
    workspace_id,
    application_id,
    draft_id,
    owner_subject_ref,
    store_schema_version,
    schema_version,
    draft_version,
    draft_status,
    lifecycle_state,
    lifecycle_version,
    archived_at_unix_nano,
    library_updated_at_unix_nano,
    lifecycle_updated_by_actor_ref,
    draft_name,
    validation_state,
    provenance_kind,
    sanitized_draft_payload,
    created_at_unix_nano,
    updated_at_unix_nano`

const sqliteSavedWorkflowDraftInsertSQL = `
INSERT INTO saved_workflow_drafts (
    tenant_ref, workspace_id, application_id, draft_id, owner_subject_ref,
    store_schema_version, schema_version, draft_version, draft_status,
    lifecycle_state, lifecycle_version, archived_at_unix_nano,
    library_updated_at_unix_nano, lifecycle_updated_by_actor_ref,
    draft_name, validation_state, provenance_kind,
    sanitized_draft_payload, validation_summary, blocked_capability_summary,
    created_at_unix_nano, updated_at_unix_nano, created_by_actor_ref,
    updated_by_actor_ref, request_id, audit_ref
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
ON CONFLICT (tenant_ref, workspace_id, application_id, draft_id) DO NOTHING
RETURNING ` + sqliteSavedWorkflowDraftReturningColumns

const sqliteSavedWorkflowDraftUpdateSQL = `
UPDATE saved_workflow_drafts
   SET store_schema_version=?, schema_version=?, draft_version=?, draft_status=?,
       draft_name=?, validation_state=?, provenance_kind=?,
       sanitized_draft_payload=?, validation_summary=?, blocked_capability_summary=?,
       updated_at_unix_nano=?, library_updated_at_unix_nano=?,
       updated_by_actor_ref=?, request_id=?, audit_ref=?
 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND draft_id=?
   AND owner_subject_ref=? AND draft_version=? AND lifecycle_version=?
   AND lifecycle_state='active' AND created_at_unix_nano=?
RETURNING ` + sqliteSavedWorkflowDraftReturningColumns

const sqliteSavedWorkflowDraftReadSQL = `
SELECT ` + sqliteSavedWorkflowDraftReturningColumns + `
  FROM saved_workflow_drafts
 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND draft_id=? AND owner_subject_ref=?`

const sqliteSavedWorkflowDraftListSQL = `
SELECT ` + sqliteSavedWorkflowDraftReturningColumns + `
  FROM saved_workflow_drafts
 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=?
 ORDER BY updated_at_unix_nano DESC, draft_id ASC
 LIMIT ?`

const sqliteSavedWorkflowDraftRevisionColumns = `
    tenant_ref,
    workspace_id,
    application_id,
    draft_id,
    owner_subject_ref,
    draft_version,
    revision_kind,
    restored_from_version,
    sanitized_revision_record`

const sqliteSavedWorkflowDraftRevisionInsertSQL = `
INSERT INTO saved_workflow_draft_revisions (
    tenant_ref, workspace_id, application_id, draft_id, owner_subject_ref,
    draft_version, revision_kind, restored_from_version, sanitized_revision_record
) VALUES (?,?,?,?,?,?,?,?,?)`

const sqliteSavedWorkflowDraftRevisionReadSQL = `
SELECT ` + sqliteSavedWorkflowDraftRevisionColumns + `
  FROM saved_workflow_draft_revisions
 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND draft_id=?
   AND owner_subject_ref=? AND draft_version=?`

const sqliteSavedWorkflowDraftRevisionListSQL = `
SELECT ` + sqliteSavedWorkflowDraftRevisionColumns + `
  FROM saved_workflow_draft_revisions
 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND draft_id=?
   AND owner_subject_ref=? AND (?=0 OR draft_version < ?)
 ORDER BY draft_version DESC
 LIMIT ?`

var _ SavedWorkflowDraftRepositoryQueryExecutor = (*sqliteSavedWorkflowDraftQueryExecutor)(nil)
var _ SavedWorkflowDraftRevisionRepositoryQueryExecutor = (*sqliteSavedWorkflowDraftQueryExecutor)(nil)
