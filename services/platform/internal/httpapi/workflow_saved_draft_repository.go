package httpapi

import "context"

const savedWorkflowDraftRepositoryStoreSchemaVersion = "saved_workflow_drafts_store_v1"

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

type SaveWorkflowDraftRecordRequest struct {
	ExpectedDraftVersion int
	Draft                SavedWorkflowDraft
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
}

type savedWorkflowDraftRepositoryReadQuery struct {
	ActorContext SavedWorkflowDraftRepositoryActorContext
	DraftID      string
}

type savedWorkflowDraftRepositoryListQuery struct {
	ActorContext SavedWorkflowDraftRepositoryActorContext
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

type SavedWorkflowDraftRepositorySchemaPreflight struct {
	StoreSchemaVersion string
	MigrationState     string
}

type SavedWorkflowDraftRepositoryAdapterConfig struct {
	QueryExecutor   SavedWorkflowDraftRepositoryQueryExecutor
	SchemaPreflight SavedWorkflowDraftRepositorySchemaPreflight
}
