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

	var row sqliteSavedWorkflowDraftRow
	if query.ExpectedDraftVersion == 0 {
		row = executor.database.QueryRowContext(
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
		row = executor.database.QueryRowContext(
			ctx,
			sqliteSavedWorkflowDraftUpdateSQL,
			query.Record.StoreSchemaVersion,
			query.Record.Draft.SchemaVersion,
			query.Record.Draft.DraftVersion,
			query.Record.Draft.DraftStatus,
			string(payload),
			string(validation),
			string(blocked),
			updatedAtUnixNano,
			query.Record.Draft.UpdatedByActorRef,
			query.Record.Draft.RequestAuditMetadata.RequestID,
			query.Record.Draft.RequestAuditMetadata.AuditRef,
			query.Record.TenantRef,
			query.Record.WorkspaceID,
			query.Record.ApplicationID,
			query.Record.DraftID,
			query.Record.OwnerSubjectRef,
			query.ExpectedDraftVersion,
			createdAtUnixNano,
		)
	}

	record, err := scanSQLiteSavedWorkflowDraftRecord(row)
	if err == nil {
		return savedWorkflowDraftRepositoryQuerySaveResult{
			Record:              record,
			CurrentDraftVersion: record.Draft.DraftVersion,
		}
	}
	if errors.Is(err, errSavedWorkflowDraftStoredRecordContract) {
		return savedWorkflowDraftRepositoryQuerySaveResult{FailureCode: SavedWorkflowDraftFailureStoreContractMismatch}
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return savedWorkflowDraftRepositoryQuerySaveResult{FailureCode: SavedWorkflowDraftFailureStoreUnavailable}
	}
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
			if errors.Is(scanErr, errSavedWorkflowDraftStoredRecordContract) {
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

func (executor *sqliteSavedWorkflowDraftQueryExecutor) failedCASResult(
	ctx context.Context,
	query savedWorkflowDraftRepositorySaveQuery,
) savedWorkflowDraftRepositoryQuerySaveResult {
	currentVersion, owner, found, lookupFailed := executor.currentVersionAndOwner(
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
	if owner != query.Record.OwnerSubjectRef {
		return savedWorkflowDraftRepositoryQuerySaveResult{
			FailureCode:         SavedWorkflowDraftFailureScopeDenied,
			CurrentDraftVersion: currentVersion,
		}
	}
	return savedWorkflowDraftRepositoryQuerySaveResult{
		FailureCode:         SavedWorkflowDraftFailureVersionConflict,
		CurrentDraftVersion: currentVersion,
	}
}

func (executor *sqliteSavedWorkflowDraftQueryExecutor) currentVersionAndOwner(
	ctx context.Context,
	actor SavedWorkflowDraftRepositoryActorContext,
	draftID string,
) (int, string, bool, bool) {
	var currentVersion int
	var owner string
	err := executor.database.QueryRowContext(
		ctx,
		sqliteSavedWorkflowDraftCurrentVersionSQL,
		actor.TenantRef,
		actor.WorkspaceID,
		actor.ApplicationID,
		draftID,
	).Scan(&currentVersion, &owner)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", false, false
	}
	if err != nil {
		return 0, "", false, true
	}
	return currentVersion, owner, true, false
}

func scanSQLiteSavedWorkflowDraftRecord(
	row sqliteSavedWorkflowDraftRow,
) (SavedWorkflowDraftRepositoryStoredRecord, error) {
	record := SavedWorkflowDraftRepositoryStoredRecord{}
	var payload []byte
	var draftVersion int
	var schemaVersion string
	var draftStatus string
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
	return decoded, nil
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
    sanitized_draft_payload,
    created_at_unix_nano,
    updated_at_unix_nano`

const sqliteSavedWorkflowDraftInsertSQL = `
INSERT INTO saved_workflow_drafts (
    tenant_ref, workspace_id, application_id, draft_id, owner_subject_ref,
    store_schema_version, schema_version, draft_version, draft_status,
    sanitized_draft_payload, validation_summary, blocked_capability_summary,
    created_at_unix_nano, updated_at_unix_nano, created_by_actor_ref,
    updated_by_actor_ref, request_id, audit_ref
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
ON CONFLICT (tenant_ref, workspace_id, application_id, draft_id) DO NOTHING
RETURNING ` + sqliteSavedWorkflowDraftReturningColumns

const sqliteSavedWorkflowDraftUpdateSQL = `
UPDATE saved_workflow_drafts
   SET store_schema_version=?, schema_version=?, draft_version=?, draft_status=?,
       sanitized_draft_payload=?, validation_summary=?, blocked_capability_summary=?,
       updated_at_unix_nano=?, updated_by_actor_ref=?, request_id=?, audit_ref=?
 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND draft_id=?
   AND owner_subject_ref=? AND draft_version=? AND created_at_unix_nano=?
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

const sqliteSavedWorkflowDraftCurrentVersionSQL = `
SELECT draft_version, owner_subject_ref
  FROM saved_workflow_drafts
 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND draft_id=?`

var _ SavedWorkflowDraftRepositoryQueryExecutor = (*sqliteSavedWorkflowDraftQueryExecutor)(nil)
