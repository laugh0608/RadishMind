package httpapi

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresSavedWorkflowDraftQueryExecutor struct {
	pool *pgxpool.Pool
}

type savedWorkflowDraftRowScanner interface {
	Scan(...any) error
}

func newPostgresSavedWorkflowDraftQueryExecutor(
	pool *pgxpool.Pool,
) *postgresSavedWorkflowDraftQueryExecutor {
	return &postgresSavedWorkflowDraftQueryExecutor{pool: pool}
}

func (executor *postgresSavedWorkflowDraftQueryExecutor) SaveWorkflowDraftRecord(
	ctx context.Context,
	query savedWorkflowDraftRepositorySaveQuery,
) savedWorkflowDraftRepositoryQuerySaveResult {
	if executor == nil || executor.pool == nil || ctx == nil {
		return savedWorkflowDraftRepositoryQuerySaveResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	if query.ExpectedDraftVersion < 0 || query.Record.Draft.DraftVersion != query.ExpectedDraftVersion+1 {
		return savedWorkflowDraftRepositoryQuerySaveResult{
			FailureCode: SavedWorkflowDraftFailureStoreContractMismatch,
		}
	}
	payload, validation, blocked, createdAt, updatedAt, failureCode :=
		savedWorkflowDraftRecordValues(query.Record)
	if failureCode != "" {
		return savedWorkflowDraftRepositoryQuerySaveResult{FailureCode: failureCode}
	}

	var row savedWorkflowDraftRowScanner
	if query.ExpectedDraftVersion == 0 {
		row = executor.pool.QueryRow(
			ctx,
			postgresSavedWorkflowDraftInsertSQL,
			query.Record.TenantRef,
			query.Record.WorkspaceID,
			query.Record.ApplicationID,
			query.Record.DraftID,
			query.Record.OwnerSubjectRef,
			query.Record.StoreSchemaVersion,
			query.Record.Draft.SchemaVersion,
			query.Record.Draft.DraftVersion,
			query.Record.Draft.DraftStatus,
			payload,
			validation,
			blocked,
			createdAt,
			updatedAt,
			query.Record.Draft.CreatedByActorRef,
			query.Record.Draft.UpdatedByActorRef,
			query.Record.Draft.RequestAuditMetadata.RequestID,
			query.Record.Draft.RequestAuditMetadata.AuditRef,
		)
	} else {
		row = executor.pool.QueryRow(
			ctx,
			postgresSavedWorkflowDraftUpdateSQL,
			query.Record.StoreSchemaVersion,
			query.Record.Draft.SchemaVersion,
			query.Record.Draft.DraftVersion,
			query.Record.Draft.DraftStatus,
			payload,
			validation,
			blocked,
			updatedAt,
			query.Record.Draft.UpdatedByActorRef,
			query.Record.Draft.RequestAuditMetadata.RequestID,
			query.Record.Draft.RequestAuditMetadata.AuditRef,
			query.Record.TenantRef,
			query.Record.WorkspaceID,
			query.Record.ApplicationID,
			query.Record.DraftID,
			query.Record.OwnerSubjectRef,
			query.ExpectedDraftVersion,
		)
	}

	record, err := scanPostgresSavedWorkflowDraftRecord(row)
	if err == nil {
		return savedWorkflowDraftRepositoryQuerySaveResult{
			Record:              record,
			CurrentDraftVersion: record.Draft.DraftVersion,
		}
	}
	if errors.Is(err, errSavedWorkflowDraftStoredRecordContract) {
		return savedWorkflowDraftRepositoryQuerySaveResult{
			FailureCode: SavedWorkflowDraftFailureStoreContractMismatch,
		}
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return savedWorkflowDraftRepositoryQuerySaveResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	return executor.failedCASResult(ctx, query)
}

func (executor *postgresSavedWorkflowDraftQueryExecutor) ReadWorkflowDraftRecord(
	ctx context.Context,
	query savedWorkflowDraftRepositoryReadQuery,
) savedWorkflowDraftRepositoryQueryReadResult {
	if executor == nil || executor.pool == nil || ctx == nil {
		return savedWorkflowDraftRepositoryQueryReadResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	record, err := scanPostgresSavedWorkflowDraftRecord(executor.pool.QueryRow(
		ctx,
		postgresSavedWorkflowDraftReadSQL,
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
		return savedWorkflowDraftRepositoryQueryReadResult{
			FailureCode: SavedWorkflowDraftFailureStoreContractMismatch,
		}
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return savedWorkflowDraftRepositoryQueryReadResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	currentVersion, owner, found, lookupFailed := executor.currentVersionAndOwner(
		ctx,
		query.ActorContext,
		query.DraftID,
	)
	if lookupFailed {
		return savedWorkflowDraftRepositoryQueryReadResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	if !found {
		return savedWorkflowDraftRepositoryQueryReadResult{
			FailureCode: SavedWorkflowDraftFailureNotFound,
		}
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

func (executor *postgresSavedWorkflowDraftQueryExecutor) ListWorkflowDraftRecords(
	ctx context.Context,
	query savedWorkflowDraftRepositoryListQuery,
) savedWorkflowDraftRepositoryQueryListResult {
	if executor == nil || executor.pool == nil || ctx == nil {
		return savedWorkflowDraftRepositoryQueryListResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	rows, err := executor.pool.Query(
		ctx,
		postgresSavedWorkflowDraftListSQL,
		query.ActorContext.TenantRef,
		query.ActorContext.WorkspaceID,
		query.ActorContext.ApplicationID,
		query.ActorContext.OwnerSubjectRef,
		savedWorkflowDraftRepositoryListLimit,
	)
	if err != nil {
		return savedWorkflowDraftRepositoryQueryListResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	defer rows.Close()

	records := make([]SavedWorkflowDraftRepositoryStoredRecord, 0)
	for rows.Next() {
		record, scanErr := scanPostgresSavedWorkflowDraftRecord(rows)
		if scanErr != nil {
			return savedWorkflowDraftRepositoryQueryListResult{
				FailureCode: SavedWorkflowDraftFailureStoreContractMismatch,
			}
		}
		records = append(records, record)
	}
	if rows.Err() != nil {
		return savedWorkflowDraftRepositoryQueryListResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	return savedWorkflowDraftRepositoryQueryListResult{Records: records}
}

func (executor *postgresSavedWorkflowDraftQueryExecutor) failedCASResult(
	ctx context.Context,
	query savedWorkflowDraftRepositorySaveQuery,
) savedWorkflowDraftRepositoryQuerySaveResult {
	currentVersion, owner, found, lookupFailed := executor.currentVersionAndOwner(
		ctx,
		query.ActorContext,
		query.Record.DraftID,
	)
	if lookupFailed {
		return savedWorkflowDraftRepositoryQuerySaveResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	if !found {
		return savedWorkflowDraftRepositoryQuerySaveResult{
			FailureCode: SavedWorkflowDraftFailureNotFound,
		}
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

func (executor *postgresSavedWorkflowDraftQueryExecutor) currentVersionAndOwner(
	ctx context.Context,
	actor SavedWorkflowDraftRepositoryActorContext,
	draftID string,
) (int, string, bool, bool) {
	var currentVersion int
	var owner string
	err := executor.pool.QueryRow(
		ctx,
		postgresSavedWorkflowDraftCurrentVersionSQL,
		actor.TenantRef,
		actor.WorkspaceID,
		actor.ApplicationID,
		draftID,
	).Scan(&currentVersion, &owner)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, "", false, false
	}
	if err != nil {
		return 0, "", false, true
	}
	return currentVersion, owner, true, false
}

func scanPostgresSavedWorkflowDraftRecord(
	row savedWorkflowDraftRowScanner,
) (SavedWorkflowDraftRepositoryStoredRecord, error) {
	record := SavedWorkflowDraftRepositoryStoredRecord{}
	var payload []byte
	var draftVersion int
	var schemaVersion string
	var draftStatus string
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
	); err != nil {
		return SavedWorkflowDraftRepositoryStoredRecord{}, err
	}
	decoded, err := decodeSavedWorkflowDraftStoredRecord(record, payload)
	if err != nil || decoded.Draft.SchemaVersion != schemaVersion ||
		decoded.Draft.DraftVersion != draftVersion || string(decoded.Draft.DraftStatus) != draftStatus {
		return SavedWorkflowDraftRepositoryStoredRecord{}, errSavedWorkflowDraftStoredRecordContract
	}
	return decoded, nil
}

const postgresSavedWorkflowDraftReturningColumns = `
    tenant_ref,
    workspace_id,
    application_id,
    draft_id,
    owner_subject_ref,
    store_schema_version,
    schema_version,
    draft_version,
    draft_status,
    sanitized_draft_payload`

const postgresSavedWorkflowDraftInsertSQL = `
INSERT INTO saved_workflow_drafts (
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
    validation_summary,
    blocked_capability_summary,
    created_at,
    updated_at,
    created_by_actor_ref,
    updated_by_actor_ref,
    request_id,
    audit_ref
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9,
    $10, $11, $12, $13, $14, $15, $16, $17, $18
)
ON CONFLICT (tenant_ref, workspace_id, application_id, draft_id) DO NOTHING
RETURNING ` + postgresSavedWorkflowDraftReturningColumns

const postgresSavedWorkflowDraftUpdateSQL = `
UPDATE saved_workflow_drafts
   SET store_schema_version = $1,
       schema_version = $2,
       draft_version = $3,
       draft_status = $4,
       sanitized_draft_payload = $5,
       validation_summary = $6,
       blocked_capability_summary = $7,
       updated_at = $8,
       updated_by_actor_ref = $9,
       request_id = $10,
       audit_ref = $11
 WHERE tenant_ref = $12
   AND workspace_id = $13
   AND application_id = $14
   AND draft_id = $15
   AND owner_subject_ref = $16
   AND draft_version = $17
RETURNING ` + postgresSavedWorkflowDraftReturningColumns

const postgresSavedWorkflowDraftReadSQL = `
SELECT ` + postgresSavedWorkflowDraftReturningColumns + `
  FROM saved_workflow_drafts
 WHERE tenant_ref = $1
   AND workspace_id = $2
   AND application_id = $3
   AND draft_id = $4
   AND owner_subject_ref = $5`

const postgresSavedWorkflowDraftListSQL = `
SELECT ` + postgresSavedWorkflowDraftReturningColumns + `
  FROM saved_workflow_drafts
 WHERE tenant_ref = $1
   AND workspace_id = $2
   AND application_id = $3
   AND owner_subject_ref = $4
 ORDER BY updated_at DESC, draft_id ASC
 LIMIT $5`

const postgresSavedWorkflowDraftCurrentVersionSQL = `
SELECT draft_version, owner_subject_ref
  FROM saved_workflow_drafts
 WHERE tenant_ref = $1
   AND workspace_id = $2
   AND application_id = $3
   AND draft_id = $4`
