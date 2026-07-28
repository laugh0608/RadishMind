package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/sqlitedev"
)

func TestSavedWorkflowDraftSQLiteLibraryMatchesMemoryGoldenMatrix(t *testing.T) {
	requestContext := savedWorkflowDraftSQLiteContext()
	memoryService := newSavedWorkflowDraftService(newMemorySavedWorkflowDraftStore())
	sqliteRuntime := openSavedWorkflowDraftSQLiteRuntime(
		t,
		filepath.Join(t.TempDir(), "library-matrix.db"),
	)
	sqliteService := newSavedWorkflowDraftService(
		newRepositorySavedWorkflowDraftLibraryStore(
			newSQLiteSavedWorkflowDraftStore(sqliteRuntime.DB()),
		),
	)
	populateSavedWorkflowDraftLibraryFixture(t, &memoryService, requestContext)
	populateSavedWorkflowDraftLibraryFixture(t, &sqliteService, requestContext)

	cases := []ListWorkflowDraftsRequest{
		{LifecycleState: SavedWorkflowDraftLifecycleActive, Limit: 5},
		{LifecycleState: SavedWorkflowDraftLifecycleArchived, Limit: 5},
		{
			LifecycleState: SavedWorkflowDraftLifecycleActive,
			Limit:          5,
			NamePrefix:     "Alpha",
		},
		{
			LifecycleState:  SavedWorkflowDraftLifecycleActive,
			Limit:           5,
			ValidationState: SavedWorkflowDraftStatusBlockedCapability,
		},
		{
			LifecycleState: SavedWorkflowDraftLifecycleActive,
			Limit:          5,
			ProvenanceKind: SavedWorkflowDraftProvenanceDraftDerived,
		},
		{
			LifecycleState: SavedWorkflowDraftLifecycleActive,
			Limit:          5,
			ProvenanceKind: SavedWorkflowDraftProvenanceUnversioned,
		},
	}
	for _, request := range cases {
		memoryIDs := collectSavedWorkflowDraftLibraryIDs(
			t,
			memoryService,
			requestContext,
			request,
		)
		sqliteIDs := collectSavedWorkflowDraftLibraryIDs(
			t,
			sqliteService,
			requestContext,
			request,
		)
		if !reflect.DeepEqual(sqliteIDs, memoryIDs) {
			t.Fatalf("SQLite library matrix drifted for %#v: sqlite=%v memory=%v",
				request, sqliteIDs, memoryIDs)
		}
	}
}

func TestSavedWorkflowDraftSQLiteLibraryContractRequiresExplicitOptIn(t *testing.T) {
	runtime := openSavedWorkflowDraftSQLiteRuntime(
		t,
		filepath.Join(t.TempDir(), "library-opt-in.db"),
	)
	store := newSQLiteSavedWorkflowDraftStore(runtime.DB())
	if _, ok := any(store).(savedWorkflowDraftLibraryStore); ok {
		t.Fatal("legacy SQLite store unexpectedly exposes the library service contract")
	}

	service := newSavedWorkflowDraftService(store)
	requestContext := savedWorkflowDraftSQLiteContext()
	for index := 0; index < defaultSavedWorkflowDraftListLimit+5; index++ {
		payload := validSavedWorkflowDraftPayload()
		payload.DraftID = fmt.Sprintf("draft_library_legacy_%02d", index)
		payload.Name = fmt.Sprintf("Legacy library draft %02d", index)
		result := service.SaveDraft(
			requestContext,
			SaveWorkflowDraftRequest{Payload: payload},
		)
		if result.FailureCode != "" || result.Draft == nil {
			t.Fatalf("save legacy list fixture %d: %#v", index, result)
		}
	}
	result := service.ListDrafts(requestContext, ListWorkflowDraftsRequest{})
	if result.FailureCode != "" ||
		len(result.Summaries) != defaultSavedWorkflowDraftListLimit+5 ||
		result.HasMore ||
		result.NextCursor != "" {
		t.Fatalf("legacy SQLite list behavior changed before HTTP batch C: %#v", result)
	}

	libraryStore := newRepositorySavedWorkflowDraftLibraryStore(store)
	if _, ok := any(libraryStore).(savedWorkflowDraftLibraryStore); !ok {
		t.Fatal("repository library wrapper does not expose the library service contract")
	}
}

func TestSavedWorkflowDraftSQLiteLifecycleTransitionIsAtomicAndRestartSafe(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "lifecycle.db")
	runtime := openSavedWorkflowDraftSQLiteRuntime(t, databasePath)
	store := newRepositorySavedWorkflowDraftLibraryStore(
		newSQLiteSavedWorkflowDraftStore(runtime.DB()),
	)
	service := newSavedWorkflowDraftService(store)
	service.now = func() time.Time {
		return time.Date(2026, 7, 28, 13, 14, 15, 123456000, time.UTC)
	}
	requestContext := savedWorkflowDraftSQLiteContext()
	saved := service.SaveDraft(
		requestContext,
		SaveWorkflowDraftRequest{Payload: validSavedWorkflowDraftPayload()},
	)
	if saved.FailureCode != "" || saved.Draft == nil {
		t.Fatalf("save SQLite lifecycle fixture: %#v", saved)
	}
	archived := service.ArchiveDraft(
		requestContext,
		TransitionSavedWorkflowDraftLifecycleRequest{
			DraftID:                  saved.Draft.DraftID,
			ExpectedDraftVersion:     saved.Draft.DraftVersion,
			ExpectedLifecycleVersion: saved.Draft.LifecycleVersion,
		},
	)
	if archived.FailureCode != "" || archived.Draft == nil || archived.Event == nil ||
		archived.Draft.LifecycleState != SavedWorkflowDraftLifecycleArchived ||
		archived.Draft.LifecycleVersion != 2 ||
		archived.Event.LifecycleVersion != 2 {
		t.Fatalf("archive SQLite lifecycle fixture: %#v", archived)
	}
	var eventCount int
	if err := runtime.DB().QueryRowContext(
		context.Background(),
		`SELECT count(*) FROM saved_workflow_draft_lifecycle_events WHERE draft_id=?`,
		saved.Draft.DraftID,
	).Scan(&eventCount); err != nil || eventCount != 1 {
		t.Fatalf("inspect SQLite lifecycle event: count=%d err=%v", eventCount, err)
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("close SQLite lifecycle runtime: %v", err)
	}
	restarted, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   savedWorkflowDraftSQLiteMigrations(),
	})
	if err != nil {
		t.Fatalf("restart SQLite lifecycle runtime: %v", err)
	}
	t.Cleanup(func() { _ = restarted.Close() })
	restartedService := newSavedWorkflowDraftService(
		newRepositorySavedWorkflowDraftLibraryStore(
			newSQLiteSavedWorkflowDraftStore(restarted.DB()),
		),
	)
	read := restartedService.ReadDraft(
		requestContext,
		ReadWorkflowDraftRequest{DraftID: saved.Draft.DraftID},
	)
	if read.FailureCode != "" || read.Draft == nil ||
		read.Draft.LifecycleState != SavedWorkflowDraftLifecycleArchived ||
		read.Draft.LifecycleVersion != 2 ||
		read.Draft.ArchivedAt != archived.Draft.ArchivedAt {
		t.Fatalf("SQLite lifecycle did not survive restart: %#v", read)
	}
}

func TestSavedWorkflowDraftSQLiteLifecycleEventFailureRollsBackCurrentRecord(t *testing.T) {
	runtime := openSavedWorkflowDraftSQLiteRuntime(
		t,
		filepath.Join(t.TempDir(), "event-rollback.db"),
	)
	service := newSavedWorkflowDraftService(newRepositorySavedWorkflowDraftLibraryStore(
		newSQLiteSavedWorkflowDraftStore(runtime.DB()),
	))
	service.now = func() time.Time {
		return time.Date(2026, 7, 28, 14, 0, 0, 0, time.UTC)
	}
	requestContext := savedWorkflowDraftSQLiteContext()
	saved := service.SaveDraft(
		requestContext,
		SaveWorkflowDraftRequest{Payload: validSavedWorkflowDraftPayload()},
	)
	if saved.FailureCode != "" || saved.Draft == nil {
		t.Fatalf("save SQLite rollback fixture: %#v", saved)
	}
	if _, err := runtime.DB().ExecContext(context.Background(), `
CREATE TRIGGER fail_saved_workflow_draft_lifecycle_event
BEFORE INSERT ON saved_workflow_draft_lifecycle_events
BEGIN
    SELECT RAISE(ABORT, 'forced lifecycle event failure');
END`); err != nil {
		t.Fatalf("install SQLite lifecycle event failure trigger: %v", err)
	}
	result := service.ArchiveDraft(
		requestContext,
		TransitionSavedWorkflowDraftLifecycleRequest{
			DraftID:                  saved.Draft.DraftID,
			ExpectedDraftVersion:     1,
			ExpectedLifecycleVersion: 1,
		},
	)
	if result.FailureCode != SavedWorkflowDraftFailureLifecycleEventWrite ||
		result.Draft != nil || result.Event != nil {
		t.Fatalf("SQLite event failure did not fail atomically: %#v", result)
	}
	read := service.ReadDraft(
		requestContext,
		ReadWorkflowDraftRequest{DraftID: saved.Draft.DraftID},
	)
	if read.FailureCode != "" || read.Draft == nil ||
		read.Draft.LifecycleState != SavedWorkflowDraftLifecycleActive ||
		read.Draft.LifecycleVersion != 1 {
		t.Fatalf("SQLite event failure committed current lifecycle: %#v", read)
	}
	var eventCount int
	if err := runtime.DB().QueryRowContext(
		context.Background(),
		"SELECT count(*) FROM saved_workflow_draft_lifecycle_events",
	).Scan(&eventCount); err != nil || eventCount != 0 {
		t.Fatalf("SQLite event failure left an event: count=%d err=%v", eventCount, err)
	}
}

func TestSavedWorkflowDraftSQLiteConcurrentLifecycleTransitionHasOneWinner(t *testing.T) {
	runtime := openSavedWorkflowDraftSQLiteRuntime(
		t,
		filepath.Join(t.TempDir(), "lifecycle-concurrency.db"),
	)
	service := newSavedWorkflowDraftService(newRepositorySavedWorkflowDraftLibraryStore(
		newSQLiteSavedWorkflowDraftStore(runtime.DB()),
	))
	service.now = func() time.Time {
		return time.Date(2026, 7, 28, 15, 0, 0, 0, time.UTC)
	}
	requestContext := savedWorkflowDraftSQLiteContext()
	saved := service.SaveDraft(
		requestContext,
		SaveWorkflowDraftRequest{Payload: validSavedWorkflowDraftPayload()},
	)
	if saved.FailureCode != "" || saved.Draft == nil {
		t.Fatalf("save SQLite concurrent lifecycle fixture: %#v", saved)
	}

	const writers = 16
	results := make(chan SavedWorkflowDraftLifecycleTransitionResult, writers)
	var waitGroup sync.WaitGroup
	for index := 0; index < writers; index++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			results <- service.ArchiveDraft(
				requestContext,
				TransitionSavedWorkflowDraftLifecycleRequest{
					DraftID:                  saved.Draft.DraftID,
					ExpectedDraftVersion:     1,
					ExpectedLifecycleVersion: 1,
				},
			)
		}()
	}
	waitGroup.Wait()
	close(results)
	successes := 0
	conflicts := 0
	for result := range results {
		switch result.FailureCode {
		case "":
			successes++
		case SavedWorkflowDraftFailureLifecycleVersionConflict,
			SavedWorkflowDraftFailureLifecycleStateConflict:
			conflicts++
		default:
			t.Fatalf("SQLite concurrent transition returned unexpected result: %#v", result)
		}
	}
	if successes != 1 || conflicts != writers-1 {
		t.Fatalf("SQLite lifecycle CAS drifted: successes=%d conflicts=%d", successes, conflicts)
	}
}

func TestSavedWorkflowDraftSQLiteProjectionMismatchFailsClosed(t *testing.T) {
	runtime := openSavedWorkflowDraftSQLiteRuntime(
		t,
		filepath.Join(t.TempDir(), "projection.db"),
	)
	service := newSavedWorkflowDraftService(newRepositorySavedWorkflowDraftLibraryStore(
		newSQLiteSavedWorkflowDraftStore(runtime.DB()),
	))
	requestContext := savedWorkflowDraftSQLiteContext()
	saved := service.SaveDraft(
		requestContext,
		SaveWorkflowDraftRequest{Payload: validSavedWorkflowDraftPayload()},
	)
	if saved.FailureCode != "" || saved.Draft == nil {
		t.Fatalf("save SQLite projection fixture: %#v", saved)
	}
	if _, err := runtime.DB().ExecContext(
		context.Background(),
		"UPDATE saved_workflow_drafts SET draft_name='projection drift' WHERE draft_id=?",
		saved.Draft.DraftID,
	); err != nil {
		t.Fatalf("corrupt SQLite draft projection: %v", err)
	}
	result := service.ListDrafts(requestContext, ListWorkflowDraftsRequest{})
	if result.FailureCode != SavedWorkflowDraftFailureLifecycleStoreContract ||
		len(result.Summaries) != 0 {
		t.Fatalf("SQLite projection mismatch did not fail closed: %#v", result)
	}
}

func TestSavedWorkflowDraftSQLiteLegacyUpgradeBackfillsLifecycleWithoutEvent(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "legacy-upgrade.db")
	migrations := savedWorkflowDraftSQLiteMigrations()
	legacyRuntime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   migrations[:2],
	})
	if err != nil {
		t.Fatalf("open legacy SQLite draft runtime: %v", err)
	}
	legacyDraft := savedWorkflowDraftLegacyFixture(t)
	payload, err := json.Marshal(savedWorkflowDraftDocumentPointer(&legacyDraft))
	if err != nil {
		t.Fatalf("marshal legacy SQLite draft payload: %v", err)
	}
	validation, err := json.Marshal(
		savedWorkflowDraftValidationToDocument(legacyDraft.ValidationSummary),
	)
	if err != nil {
		t.Fatalf("marshal legacy SQLite validation: %v", err)
	}
	blocked, err := json.Marshal(
		savedWorkflowDraftBlockedToDocuments(legacyDraft.BlockedCapabilitySummary),
	)
	if err != nil {
		t.Fatalf("marshal legacy SQLite blocked summary: %v", err)
	}
	createdAt, _ := time.Parse(time.RFC3339Nano, legacyDraft.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339Nano, legacyDraft.UpdatedAt)
	createdAtUnixNano, _ := savedWorkflowDraftUnixNano(createdAt)
	updatedAtUnixNano, _ := savedWorkflowDraftUnixNano(updatedAt)
	requestContext := savedWorkflowDraftSQLiteContext()
	if _, err := legacyRuntime.DB().ExecContext(context.Background(), `
INSERT INTO saved_workflow_drafts (
    tenant_ref, workspace_id, application_id, draft_id, owner_subject_ref,
    store_schema_version, schema_version, draft_version, draft_status,
    sanitized_draft_payload, validation_summary, blocked_capability_summary,
    created_at_unix_nano, updated_at_unix_nano, created_by_actor_ref,
    updated_by_actor_ref, request_id, audit_ref
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		requestContext.TenantRef,
		legacyDraft.WorkspaceID,
		legacyDraft.ApplicationID,
		legacyDraft.DraftID,
		requestContext.OwnerSubjectRef,
		savedWorkflowDraftRepositoryStoreSchemaVersion,
		legacyDraft.SchemaVersion,
		legacyDraft.DraftVersion,
		legacyDraft.DraftStatus,
		string(payload),
		string(validation),
		string(blocked),
		createdAtUnixNano,
		updatedAtUnixNano,
		legacyDraft.CreatedByActorRef,
		legacyDraft.UpdatedByActorRef,
		legacyDraft.RequestAuditMetadata.RequestID,
		legacyDraft.RequestAuditMetadata.AuditRef,
	); err != nil {
		t.Fatalf("insert legacy SQLite draft: %v", err)
	}
	if _, err := legacyRuntime.DB().ExecContext(context.Background(), `
INSERT INTO saved_workflow_draft_revisions (
    tenant_ref, workspace_id, application_id, draft_id, owner_subject_ref,
    draft_version, revision_kind, restored_from_version, sanitized_revision_record
) VALUES (?,?,?,?,?,?,?,?,?)`,
		requestContext.TenantRef,
		legacyDraft.WorkspaceID,
		legacyDraft.ApplicationID,
		legacyDraft.DraftID,
		requestContext.OwnerSubjectRef,
		legacyDraft.DraftVersion,
		SavedWorkflowDraftRevisionKindSaved,
		0,
		string(payload),
	); err != nil {
		t.Fatalf("insert legacy SQLite revision: %v", err)
	}
	if err := legacyRuntime.Close(); err != nil {
		t.Fatalf("close legacy SQLite draft runtime: %v", err)
	}

	upgraded, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   migrations,
	})
	if err != nil {
		t.Fatalf("upgrade legacy SQLite draft runtime: %v", err)
	}
	t.Cleanup(func() { _ = upgraded.Close() })
	service := newSavedWorkflowDraftService(newRepositorySavedWorkflowDraftLibraryStore(
		newSQLiteSavedWorkflowDraftStore(upgraded.DB()),
	))
	read := service.ReadDraft(
		requestContext,
		ReadWorkflowDraftRequest{DraftID: legacyDraft.DraftID},
	)
	if read.FailureCode != "" || read.Draft == nil ||
		read.Draft.DraftVersion != legacyDraft.DraftVersion ||
		read.Draft.Name != legacyDraft.Name ||
		read.Draft.LifecycleState != SavedWorkflowDraftLifecycleActive ||
		read.Draft.LifecycleVersion != 1 ||
		read.Draft.LibraryUpdatedAt != legacyDraft.UpdatedAt {
		t.Fatalf("legacy SQLite lifecycle backfill drifted: %#v", read)
	}
	var eventCount int
	var revisionCount int
	if err := upgraded.DB().QueryRowContext(
		context.Background(),
		"SELECT count(*) FROM saved_workflow_draft_lifecycle_events",
	).Scan(&eventCount); err != nil {
		t.Fatalf("count upgraded SQLite lifecycle events: %v", err)
	}
	if err := upgraded.DB().QueryRowContext(
		context.Background(),
		"SELECT count(*) FROM saved_workflow_draft_revisions WHERE draft_id=?",
		legacyDraft.DraftID,
	).Scan(&revisionCount); err != nil {
		t.Fatalf("count upgraded SQLite revisions: %v", err)
	}
	if eventCount != 0 || revisionCount != 1 {
		t.Fatalf("legacy SQLite upgrade invented history: events=%d revisions=%d",
			eventCount, revisionCount)
	}
}

func populateSavedWorkflowDraftLibraryFixture(
	t *testing.T,
	service *savedWorkflowDraftService,
	requestContext SavedWorkflowDraftContext,
) {
	t.Helper()
	baseTime := time.Date(2026, 7, 28, 10, 0, 0, 0, time.UTC)
	for index := 0; index < 14; index++ {
		service.now = func() time.Time { return baseTime.Add(time.Duration(index) * time.Second) }
		payload := validSavedWorkflowDraftPayload()
		payload.DraftID = fmt.Sprintf("draft_library_%02d", index)
		payload.Name = fmt.Sprintf("Beta draft %02d", index)
		payload.BaseDefinitionVersion = 0
		payload.SourceDefinitionID = ""
		switch index % 3 {
		case 0:
			payload.Name = fmt.Sprintf("Alpha draft %02d", index)
		case 1:
			payload.BaseDefinitionVersion = 3
			payload.SourceDefinitionID = "workflow_definition_demo"
		case 2:
			payload.AdditionalFields = map[string]any{
				savedWorkflowDraftDerivationAdditionalField: map[string]any{
					"version":              1,
					"source_kind":          savedWorkflowDraftDerivationSourceKind,
					"source_draft_id":      "draft_library_01",
					"source_draft_version": 1,
				},
			}
		}
		if index%4 == 0 {
			payload.RequestedCapabilities = []string{"executor"}
		}
		expectedLifecycleVersion := 0
		if index%3 == 2 {
			expectedLifecycleVersion = 1
		}
		result := service.SaveDraft(
			requestContext,
			SaveWorkflowDraftRequest{
				Payload:                  payload,
				ExpectedLifecycleVersion: expectedLifecycleVersion,
			},
		)
		if result.FailureCode != "" || result.Draft == nil {
			t.Fatalf("save library fixture %d: %#v", index, result)
		}
		if index%5 == 0 {
			service.now = func() time.Time {
				return baseTime.Add(time.Hour + time.Duration(index)*time.Second)
			}
			archived := service.ArchiveDraft(
				requestContext,
				TransitionSavedWorkflowDraftLifecycleRequest{
					DraftID:                  result.Draft.DraftID,
					ExpectedDraftVersion:     result.Draft.DraftVersion,
					ExpectedLifecycleVersion: result.Draft.LifecycleVersion,
				},
			)
			if archived.FailureCode != "" {
				t.Fatalf("archive library fixture %d: %#v", index, archived)
			}
		}
	}
}

func collectSavedWorkflowDraftLibraryIDs(
	t *testing.T,
	service savedWorkflowDraftService,
	requestContext SavedWorkflowDraftContext,
	request ListWorkflowDraftsRequest,
) []string {
	t.Helper()
	ids := make([]string, 0)
	seen := make(map[string]bool)
	cursor := ""
	for {
		request.Cursor = cursor
		result := service.ListDrafts(requestContext, request)
		if result.FailureCode != "" {
			t.Fatalf("list draft library page for %#v: %#v", request, result)
		}
		for _, summary := range result.Summaries {
			if seen[summary.DraftID] {
				t.Fatalf("draft library traversal repeated %q for %#v", summary.DraftID, request)
			}
			seen[summary.DraftID] = true
			ids = append(ids, summary.DraftID)
		}
		if !result.HasMore {
			break
		}
		if result.NextCursor == "" {
			t.Fatalf("draft library page omitted cursor: %#v", result)
		}
		cursor = result.NextCursor
	}
	return ids
}

func savedWorkflowDraftLegacyFixture(t *testing.T) SavedWorkflowDraft {
	t.Helper()
	service := newSavedWorkflowDraftService(newMemorySavedWorkflowDraftStore())
	service.now = func() time.Time {
		return time.Date(2026, 7, 20, 8, 30, 0, 0, time.UTC)
	}
	result := service.SaveDraft(
		savedWorkflowDraftSQLiteContext(),
		SaveWorkflowDraftRequest{Payload: validSavedWorkflowDraftPayload()},
	)
	if result.FailureCode != "" || result.Draft == nil {
		t.Fatalf("build legacy SQLite draft fixture: %#v", result)
	}
	legacy := cloneSavedWorkflowDraft(*result.Draft)
	legacy.LifecycleState = ""
	legacy.LifecycleVersion = 0
	legacy.ArchivedAt = ""
	legacy.LibraryUpdatedAt = ""
	legacy.LifecycleUpdatedByActorRef = ""
	legacy.ProvenanceKind = ""
	return legacy
}
