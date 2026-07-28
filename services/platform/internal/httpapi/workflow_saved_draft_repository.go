package httpapi

import (
	"context"
	"time"
)

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

type SavedWorkflowDraftLibraryRepository interface {
	ListWorkflowDraftLibraryPage(
		ctx context.Context,
		actor SavedWorkflowDraftRepositoryActorContext,
		request ListWorkflowDraftLibraryPageRequest,
	) ListWorkflowDraftLibraryPageResult
	TransitionWorkflowDraftLifecycle(
		ctx context.Context,
		actor SavedWorkflowDraftRepositoryActorContext,
		request TransitionWorkflowDraftLifecycleRecordRequest,
	) TransitionWorkflowDraftLifecycleRecordResult
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

type ListWorkflowDraftLibraryPageRequest struct {
	LifecycleState        SavedWorkflowDraftLifecycleState
	Limit                 int
	NamePrefix            string
	ValidationState       SavedWorkflowDraftStatus
	ProvenanceKind        SavedWorkflowDraftProvenanceKind
	AfterLibraryUpdatedAt string
	AfterDraftID          string
}

type ListWorkflowDraftLibraryPageResult struct {
	Summaries   []SavedWorkflowDraftSummary
	HasMore     bool
	FailureCode SavedWorkflowDraftFailureCode
}

type TransitionWorkflowDraftLifecycleRecordRequest struct {
	DraftID                  string
	TargetState              SavedWorkflowDraftLifecycleState
	ExpectedDraftVersion     int
	ExpectedLifecycleVersion int
	OccurredAt               time.Time
}

type TransitionWorkflowDraftLifecycleRecordResult struct {
	Draft                   *SavedWorkflowDraft
	Event                   *SavedWorkflowDraftLifecycleEvent
	FailureCode             SavedWorkflowDraftFailureCode
	CurrentDraftVersion     int
	CurrentLifecycleVersion int
	CurrentLifecycleState   SavedWorkflowDraftLifecycleState
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

type savedWorkflowDraftRepositoryLibraryPageQuery struct {
	ActorContext SavedWorkflowDraftRepositoryActorContext
	Filter       savedWorkflowDraftLibraryListFilter
}

type savedWorkflowDraftRepositoryLifecycleTransitionQuery struct {
	ActorContext             SavedWorkflowDraftRepositoryActorContext
	DraftID                  string
	TargetState              SavedWorkflowDraftLifecycleState
	ExpectedDraftVersion     int
	ExpectedLifecycleVersion int
	OccurredAt               time.Time
}

type savedWorkflowDraftRepositoryQueryLibraryPageResult struct {
	Records     []SavedWorkflowDraftRepositoryStoredRecord
	HasMore     bool
	FailureCode SavedWorkflowDraftFailureCode
}

type savedWorkflowDraftRepositoryQueryLifecycleTransitionResult struct {
	Record                  SavedWorkflowDraftRepositoryStoredRecord
	Event                   SavedWorkflowDraftLifecycleEvent
	FailureCode             SavedWorkflowDraftFailureCode
	CurrentDraftVersion     int
	CurrentLifecycleVersion int
	CurrentLifecycleState   SavedWorkflowDraftLifecycleState
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

type SavedWorkflowDraftLibraryRepositoryQueryExecutor interface {
	ListWorkflowDraftLibraryPage(
		ctx context.Context,
		query savedWorkflowDraftRepositoryLibraryPageQuery,
	) savedWorkflowDraftRepositoryQueryLibraryPageResult
	TransitionWorkflowDraftLifecycle(
		ctx context.Context,
		query savedWorkflowDraftRepositoryLifecycleTransitionQuery,
	) savedWorkflowDraftRepositoryQueryLifecycleTransitionResult
}

type SavedWorkflowDraftRepositorySchemaPreflight struct {
	StoreSchemaVersion string
	MigrationState     string
}

type SavedWorkflowDraftRepositoryAdapterConfig struct {
	QueryExecutor   SavedWorkflowDraftRepositoryQueryExecutor
	SchemaPreflight SavedWorkflowDraftRepositorySchemaPreflight
}
