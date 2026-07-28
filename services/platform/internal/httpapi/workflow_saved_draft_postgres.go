package httpapi

import (
	"context"
	"errors"
	"time"

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
	if !validSavedWorkflowDraftRevisionWriteMetadata(query.RevisionKind, query.RestoredFromVersion) {
		return savedWorkflowDraftRepositoryQuerySaveResult{
			FailureCode: SavedWorkflowDraftFailureStoreContractMismatch,
		}
	}
	payload, validation, blocked, createdAt, updatedAt, failureCode :=
		savedWorkflowDraftRecordValues(query.Record)
	if failureCode != "" {
		return savedWorkflowDraftRepositoryQuerySaveResult{FailureCode: failureCode}
	}
	libraryUpdatedAt, err := time.Parse(time.RFC3339Nano, query.Record.Draft.LibraryUpdatedAt)
	if err != nil || !libraryUpdatedAt.Equal(updatedAt) {
		return savedWorkflowDraftRepositoryQuerySaveResult{
			FailureCode: SavedWorkflowDraftFailureLifecycleStoreContract,
		}
	}

	transaction, err := executor.pool.Begin(ctx)
	if err != nil {
		return savedWorkflowDraftRepositoryQuerySaveResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()

	var row savedWorkflowDraftRowScanner
	if query.ExpectedDraftVersion == 0 {
		row = transaction.QueryRow(
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
			query.Record.Draft.LifecycleState,
			query.Record.Draft.LifecycleVersion,
			nil,
			libraryUpdatedAt,
			query.Record.Draft.LifecycleUpdatedByActorRef,
			query.Record.Draft.Name,
			query.Record.Draft.ValidationSummary.ValidationState,
			query.Record.Draft.ProvenanceKind,
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
		row = transaction.QueryRow(
			ctx,
			postgresSavedWorkflowDraftUpdateSQL,
			query.Record.StoreSchemaVersion,
			query.Record.Draft.SchemaVersion,
			query.Record.Draft.DraftVersion,
			query.Record.Draft.DraftStatus,
			query.Record.Draft.Name,
			query.Record.Draft.ValidationSummary.ValidationState,
			query.Record.Draft.ProvenanceKind,
			payload,
			validation,
			blocked,
			updatedAt,
			libraryUpdatedAt,
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
		)
	}

	record, err := scanPostgresSavedWorkflowDraftRecord(row)
	if err == nil {
		if _, insertErr := transaction.Exec(
			ctx,
			postgresSavedWorkflowDraftRevisionInsertSQL,
			query.Record.TenantRef,
			query.Record.WorkspaceID,
			query.Record.ApplicationID,
			query.Record.DraftID,
			query.Record.OwnerSubjectRef,
			query.Record.Draft.DraftVersion,
			query.RevisionKind,
			query.RestoredFromVersion,
			payload,
		); insertErr != nil {
			return savedWorkflowDraftRepositoryQuerySaveResult{
				FailureCode: SavedWorkflowDraftFailureStoreContractMismatch,
			}
		}
		if commitErr := transaction.Commit(ctx); commitErr != nil {
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
		return savedWorkflowDraftRepositoryQuerySaveResult{
			FailureCode: SavedWorkflowDraftFailureStoreContractMismatch,
		}
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return savedWorkflowDraftRepositoryQuerySaveResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	_ = transaction.Rollback(ctx)
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
	if errors.Is(err, errSavedWorkflowDraftStoredLibraryProjection) {
		return savedWorkflowDraftRepositoryQueryReadResult{
			FailureCode: SavedWorkflowDraftFailureLifecycleStoreContract,
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
			failureCode := SavedWorkflowDraftFailureStoreContractMismatch
			if errors.Is(scanErr, errSavedWorkflowDraftStoredLibraryProjection) {
				failureCode = SavedWorkflowDraftFailureLifecycleStoreContract
			}
			return savedWorkflowDraftRepositoryQueryListResult{
				FailureCode: failureCode,
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

func (executor *postgresSavedWorkflowDraftQueryExecutor) ReadWorkflowDraftRevision(
	ctx context.Context,
	query savedWorkflowDraftRepositoryRevisionReadQuery,
) savedWorkflowDraftRepositoryQueryRevisionReadResult {
	if executor == nil || executor.pool == nil || ctx == nil {
		return savedWorkflowDraftRepositoryQueryRevisionReadResult{
			FailureCode: SavedWorkflowDraftFailureStoreUnavailable,
		}
	}
	revision, err := scanPostgresSavedWorkflowDraftRevision(executor.pool.QueryRow(
		ctx,
		postgresSavedWorkflowDraftRevisionReadSQL,
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
	if !errors.Is(err, pgx.ErrNoRows) {
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

func (executor *postgresSavedWorkflowDraftQueryExecutor) ListWorkflowDraftRevisions(
	ctx context.Context,
	query savedWorkflowDraftRepositoryRevisionListQuery,
) savedWorkflowDraftRepositoryQueryRevisionListResult {
	if executor == nil || executor.pool == nil || ctx == nil {
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
	rows, err := executor.pool.Query(
		ctx,
		postgresSavedWorkflowDraftRevisionListSQL,
		query.ActorContext.TenantRef,
		query.ActorContext.WorkspaceID,
		query.ActorContext.ApplicationID,
		query.DraftID,
		query.ActorContext.OwnerSubjectRef,
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
		revision, scanErr := scanPostgresSavedWorkflowDraftRevision(rows)
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

func (executor *postgresSavedWorkflowDraftQueryExecutor) failedCASResult(
	ctx context.Context,
	query savedWorkflowDraftRepositorySaveQuery,
) savedWorkflowDraftRepositoryQuerySaveResult {
	state, found, lookupFailed := executor.currentDraftState(
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

func (executor *postgresSavedWorkflowDraftQueryExecutor) currentVersionAndOwner(
	ctx context.Context,
	actor SavedWorkflowDraftRepositoryActorContext,
	draftID string,
) (int, string, bool, bool) {
	state, found, failed := executor.currentDraftState(ctx, actor, draftID)
	return state.DraftVersion, state.OwnerSubjectRef, found, failed
}

func scanPostgresSavedWorkflowDraftRecord(
	row savedWorkflowDraftRowScanner,
) (SavedWorkflowDraftRepositoryStoredRecord, error) {
	record := SavedWorkflowDraftRepositoryStoredRecord{}
	var payload []byte
	var draftVersion int
	var schemaVersion string
	var draftStatus string
	var lifecycleState string
	var lifecycleVersion int
	var archivedAt *time.Time
	var libraryUpdatedAt time.Time
	var lifecycleUpdatedByActorRef string
	var draftName string
	var validationState string
	var provenanceKind string
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
		&archivedAt,
		&libraryUpdatedAt,
		&lifecycleUpdatedByActorRef,
		&draftName,
		&validationState,
		&provenanceKind,
		&payload,
	); err != nil {
		return SavedWorkflowDraftRepositoryStoredRecord{}, err
	}
	decoded, err := decodeSavedWorkflowDraftStoredRecord(record, payload)
	if err != nil || decoded.Draft.SchemaVersion != schemaVersion ||
		decoded.Draft.DraftVersion != draftVersion || string(decoded.Draft.DraftStatus) != draftStatus {
		return SavedWorkflowDraftRepositoryStoredRecord{}, errSavedWorkflowDraftStoredRecordContract
	}
	archivedAtText := ""
	if archivedAt != nil {
		archivedAtText = archivedAt.UTC().Format(time.RFC3339Nano)
	}
	return applySavedWorkflowDraftStoredLibraryProjection(
		decoded,
		lifecycleState,
		lifecycleVersion,
		archivedAtText,
		libraryUpdatedAt.UTC().Format(time.RFC3339Nano),
		lifecycleUpdatedByActorRef,
		draftName,
		validationState,
		provenanceKind,
	)
}

func scanPostgresSavedWorkflowDraftRevision(
	row savedWorkflowDraftRowScanner,
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
    lifecycle_state,
    lifecycle_version,
    archived_at,
    library_updated_at,
    lifecycle_updated_by_actor_ref,
    draft_name,
    validation_state,
    provenance_kind,
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
    lifecycle_state,
    lifecycle_version,
    archived_at,
    library_updated_at,
    lifecycle_updated_by_actor_ref,
    draft_name,
    validation_state,
    provenance_kind,
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
    $10, $11, $12, $13, $14, $15, $16, $17,
    $18, $19, $20, $21, $22, $23, $24, $25, $26
)
ON CONFLICT (tenant_ref, workspace_id, application_id, draft_id) DO NOTHING
RETURNING ` + postgresSavedWorkflowDraftReturningColumns

const postgresSavedWorkflowDraftUpdateSQL = `
UPDATE saved_workflow_drafts
   SET store_schema_version = $1,
       schema_version = $2,
       draft_version = $3,
       draft_status = $4,
       draft_name = $5,
       validation_state = $6,
       provenance_kind = $7,
       sanitized_draft_payload = $8,
       validation_summary = $9,
       blocked_capability_summary = $10,
       updated_at = $11,
       library_updated_at = $12,
       updated_by_actor_ref = $13,
       request_id = $14,
       audit_ref = $15
 WHERE tenant_ref = $16
   AND workspace_id = $17
   AND application_id = $18
   AND draft_id = $19
   AND owner_subject_ref = $20
   AND draft_version = $21
   AND lifecycle_version = $22
   AND lifecycle_state = 'active'
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

const postgresSavedWorkflowDraftRevisionColumns = `
    tenant_ref,
    workspace_id,
    application_id,
    draft_id,
    owner_subject_ref,
    draft_version,
    revision_kind,
    restored_from_version,
    sanitized_revision_record`

const postgresSavedWorkflowDraftRevisionInsertSQL = `
INSERT INTO saved_workflow_draft_revisions (
    tenant_ref,
    workspace_id,
    application_id,
    draft_id,
    owner_subject_ref,
    draft_version,
    revision_kind,
    restored_from_version,
    sanitized_revision_record
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

const postgresSavedWorkflowDraftRevisionReadSQL = `
SELECT ` + postgresSavedWorkflowDraftRevisionColumns + `
  FROM saved_workflow_draft_revisions
 WHERE tenant_ref = $1
   AND workspace_id = $2
   AND application_id = $3
   AND draft_id = $4
   AND owner_subject_ref = $5
   AND draft_version = $6`

const postgresSavedWorkflowDraftRevisionListSQL = `
SELECT ` + postgresSavedWorkflowDraftRevisionColumns + `
  FROM saved_workflow_draft_revisions
 WHERE tenant_ref = $1
   AND workspace_id = $2
   AND application_id = $3
   AND draft_id = $4
   AND owner_subject_ref = $5
   AND ($6 = 0 OR draft_version < $6)
 ORDER BY draft_version DESC
 LIMIT $7`

var _ SavedWorkflowDraftRepositoryQueryExecutor = (*postgresSavedWorkflowDraftQueryExecutor)(nil)
var _ SavedWorkflowDraftRevisionRepositoryQueryExecutor = (*postgresSavedWorkflowDraftQueryExecutor)(nil)
