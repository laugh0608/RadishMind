package httpapi

import "context"

const savedWorkflowDraftRepositoryStoreSchemaVersion = "saved_workflow_drafts_store_v1"

const savedWorkflowDraftRepositoryListLimit = 200

type SavedWorkflowDraftRepository interface {
	SaveWorkflowDraftRecord(
		ctx context.Context,
		actor SavedWorkflowDraftRepositoryActorContext,
		request SaveWorkflowDraftRecordRequest,
	) SaveWorkflowDraftRecordResult
	ReadWorkflowDraftRecord(
		ctx context.Context,
		actor SavedWorkflowDraftRepositoryActorContext,
		request ReadWorkflowDraftRecordRequest,
	) ReadWorkflowDraftRecordResult
	ListWorkflowDraftRecords(
		ctx context.Context,
		actor SavedWorkflowDraftRepositoryActorContext,
		request ListWorkflowDraftRecordsRequest,
	) ListWorkflowDraftRecordsResult
}

type SavedWorkflowDraftRevisionRepository interface {
	ReadWorkflowDraftRevision(
		ctx context.Context,
		actor SavedWorkflowDraftRepositoryActorContext,
		request ReadSavedWorkflowDraftRevisionRecordRequest,
	) ReadSavedWorkflowDraftRevisionRecordResult
	ListWorkflowDraftRevisions(
		ctx context.Context,
		actor SavedWorkflowDraftRepositoryActorContext,
		request ListSavedWorkflowDraftRevisionRecordsRequest,
	) ListSavedWorkflowDraftRevisionRecordsResult
}

type SaveWorkflowDraftRecordRequest struct {
	ExpectedDraftVersion int
	Draft                SavedWorkflowDraft
	RevisionKind         SavedWorkflowDraftRevisionKind
	RestoredFromVersion  int
}

type SaveWorkflowDraftRecordResult struct {
	Draft               *SavedWorkflowDraft
	FailureCode         SavedWorkflowDraftFailureCode
	CurrentDraftVersion int
}

type ReadWorkflowDraftRecordRequest struct {
	DraftID string
}

type ReadWorkflowDraftRecordResult struct {
	Draft               *SavedWorkflowDraft
	FailureCode         SavedWorkflowDraftFailureCode
	CurrentDraftVersion int
}

type ListWorkflowDraftRecordsRequest struct{}

type ListWorkflowDraftRecordsResult struct {
	Summaries   []SavedWorkflowDraftSummary
	FailureCode SavedWorkflowDraftFailureCode
}

type ReadSavedWorkflowDraftRevisionRecordRequest struct {
	DraftID      string
	DraftVersion int
}

type ReadSavedWorkflowDraftRevisionRecordResult struct {
	Revision    *SavedWorkflowDraftRevision
	FailureCode SavedWorkflowDraftFailureCode
}

type ListSavedWorkflowDraftRevisionRecordsRequest struct {
	DraftID       string
	BeforeVersion int
	Limit         int
}

type ListSavedWorkflowDraftRevisionRecordsResult struct {
	Revisions   []SavedWorkflowDraftRevisionSummary
	HasMore     bool
	FailureCode SavedWorkflowDraftFailureCode
}

type SavedWorkflowDraftRepositoryStoredRecord struct {
	TenantRef          string
	WorkspaceID        string
	ApplicationID      string
	DraftID            string
	OwnerSubjectRef    string
	StoreSchemaVersion string
	Draft              SavedWorkflowDraft
}

type savedWorkflowDraftRepositorySaveQuery struct {
	ActorContext         SavedWorkflowDraftRepositoryActorContext
	ExpectedDraftVersion int
	Record               SavedWorkflowDraftRepositoryStoredRecord
	RevisionKind         SavedWorkflowDraftRevisionKind
	RestoredFromVersion  int
}

type savedWorkflowDraftRepositoryReadQuery struct {
	ActorContext SavedWorkflowDraftRepositoryActorContext
	DraftID      string
}

type savedWorkflowDraftRepositoryListQuery struct {
	ActorContext SavedWorkflowDraftRepositoryActorContext
}

type savedWorkflowDraftRepositoryRevisionReadQuery struct {
	ActorContext SavedWorkflowDraftRepositoryActorContext
	DraftID      string
	DraftVersion int
}

type savedWorkflowDraftRepositoryRevisionListQuery struct {
	ActorContext  SavedWorkflowDraftRepositoryActorContext
	DraftID       string
	BeforeVersion int
	Limit         int
}

type savedWorkflowDraftRepositoryQuerySaveResult struct {
	Record              SavedWorkflowDraftRepositoryStoredRecord
	FailureCode         SavedWorkflowDraftFailureCode
	CurrentDraftVersion int
}

type savedWorkflowDraftRepositoryQueryReadResult struct {
	Record              SavedWorkflowDraftRepositoryStoredRecord
	FailureCode         SavedWorkflowDraftFailureCode
	CurrentDraftVersion int
}

type savedWorkflowDraftRepositoryQueryListResult struct {
	Records     []SavedWorkflowDraftRepositoryStoredRecord
	FailureCode SavedWorkflowDraftFailureCode
}

type savedWorkflowDraftRepositoryQueryRevisionReadResult struct {
	Revision    SavedWorkflowDraftRevision
	FailureCode SavedWorkflowDraftFailureCode
}

type savedWorkflowDraftRepositoryQueryRevisionListResult struct {
	Revisions   []SavedWorkflowDraftRevisionSummary
	HasMore     bool
	FailureCode SavedWorkflowDraftFailureCode
}

type SavedWorkflowDraftRepositoryQueryExecutor interface {
	SaveWorkflowDraftRecord(
		ctx context.Context,
		query savedWorkflowDraftRepositorySaveQuery,
	) savedWorkflowDraftRepositoryQuerySaveResult
	ReadWorkflowDraftRecord(
		ctx context.Context,
		query savedWorkflowDraftRepositoryReadQuery,
	) savedWorkflowDraftRepositoryQueryReadResult
	ListWorkflowDraftRecords(
		ctx context.Context,
		query savedWorkflowDraftRepositoryListQuery,
	) savedWorkflowDraftRepositoryQueryListResult
}

type SavedWorkflowDraftRevisionRepositoryQueryExecutor interface {
	ReadWorkflowDraftRevision(
		ctx context.Context,
		query savedWorkflowDraftRepositoryRevisionReadQuery,
	) savedWorkflowDraftRepositoryQueryRevisionReadResult
	ListWorkflowDraftRevisions(
		ctx context.Context,
		query savedWorkflowDraftRepositoryRevisionListQuery,
	) savedWorkflowDraftRepositoryQueryRevisionListResult
}

type SavedWorkflowDraftRepositorySchemaPreflight struct {
	StoreSchemaVersion string
	MigrationState     string
}

type SavedWorkflowDraftRepositoryAdapterConfig struct {
	QueryExecutor   SavedWorkflowDraftRepositoryQueryExecutor
	SchemaPreflight SavedWorkflowDraftRepositorySchemaPreflight
}
