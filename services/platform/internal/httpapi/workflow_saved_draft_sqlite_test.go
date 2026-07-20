package httpapi

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/sqlitedev"
)

func TestSavedWorkflowDraftSQLiteDomainContract(t *testing.T) {
	runtime := openSavedWorkflowDraftSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
	store := newSQLiteSavedWorkflowDraftStore(runtime.DB())
	service := newSavedWorkflowDraftService(store)
	fixedTime := time.Date(2026, 7, 14, 10, 30, 0, 0, time.UTC)
	service.now = func() time.Time { return fixedTime }
	requestContext := savedWorkflowDraftSQLiteContext()

	payloadB := validSavedWorkflowDraftPayload()
	payloadB.DraftID = "draft_b"
	payloadB.Name = "Draft B"
	if result := service.SaveDraft(requestContext, SaveWorkflowDraftRequest{Payload: payloadB}); result.FailureCode != "" || result.Draft == nil {
		t.Fatalf("create SQLite draft B: %#v", result)
	}
	payloadA := validSavedWorkflowDraftPayload()
	payloadA.DraftID = "draft_a"
	payloadA.Name = "Draft A"
	if result := service.SaveDraft(requestContext, SaveWorkflowDraftRequest{Payload: payloadA}); result.FailureCode != "" || result.Draft == nil {
		t.Fatalf("create SQLite draft A: %#v", result)
	}
	payloadA.Name = "Draft A reviewed"
	updated := service.SaveDraft(requestContext, SaveWorkflowDraftRequest{
		ExpectedDraftVersion: 1,
		Payload:              payloadA,
	})
	if updated.FailureCode != "" || updated.Draft == nil || updated.CurrentDraftVersion != 2 {
		t.Fatalf("update SQLite draft A: %#v", updated)
	}

	listed := service.ListDrafts(requestContext, ListWorkflowDraftsRequest{})
	if listed.FailureCode != "" || len(listed.Summaries) != 2 ||
		listed.Summaries[0].DraftID != "draft_a" || listed.Summaries[1].DraftID != "draft_b" {
		t.Fatalf("SQLite saved draft stable list order drifted: %#v", listed)
	}
	read := service.ReadDraft(requestContext, ReadWorkflowDraftRequest{DraftID: "draft_a"})
	if read.FailureCode != "" || read.Draft == nil || read.Draft.DraftVersion != 2 || read.Draft.Name != payloadA.Name {
		t.Fatalf("read updated SQLite saved draft: %#v", read)
	}

	otherOwner := requestContext
	otherOwner.OwnerSubjectRef = "subject_other"
	if result := service.ReadDraft(otherOwner, ReadWorkflowDraftRequest{DraftID: "draft_a"}); result.FailureCode != SavedWorkflowDraftFailureScopeDenied || result.Draft != nil {
		t.Fatalf("SQLite saved draft owner scope leaked: %#v", result)
	}
	if result := service.ListDrafts(otherOwner, ListWorkflowDraftsRequest{}); result.FailureCode != "" || len(result.Summaries) != 0 {
		t.Fatalf("SQLite saved draft owner list scope leaked: %#v", result)
	}
	for _, mutate := range []func(*SavedWorkflowDraftContext){
		func(ctx *SavedWorkflowDraftContext) { ctx.TenantRef = "tenant:other" },
		func(ctx *SavedWorkflowDraftContext) { ctx.WorkspaceID = "workspace_other" },
		func(ctx *SavedWorkflowDraftContext) { ctx.ApplicationID = "app_other" },
	} {
		candidate := requestContext
		mutate(&candidate)
		result := service.ReadDraft(candidate, ReadWorkflowDraftRequest{DraftID: "draft_a"})
		if result.FailureCode == "" || result.Draft != nil {
			t.Fatalf("SQLite saved draft caller scope leaked: %#v", result)
		}
	}

	const writers = 16
	results := make(chan SavedWorkflowDraftResult, writers)
	var waitGroup sync.WaitGroup
	for index := 0; index < writers; index++ {
		index := index
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			candidate := payloadA
			candidate.Name = fmt.Sprintf("Concurrent candidate %02d", index)
			results <- service.SaveDraft(requestContext, SaveWorkflowDraftRequest{
				ExpectedDraftVersion: 2,
				Payload:              candidate,
			})
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
			if result.CurrentDraftVersion != 3 {
				t.Fatalf("SQLite CAS winner returned wrong version: %#v", result)
			}
		case SavedWorkflowDraftFailureVersionConflict:
			conflicts++
			if result.CurrentDraftVersion != 3 {
				t.Fatalf("SQLite CAS loser did not receive current version: %#v", result)
			}
		default:
			t.Fatalf("SQLite CAS returned unexpected failure: %#v", result)
		}
	}
	if successes != 1 || conflicts != writers-1 {
		t.Fatalf("SQLite saved draft CAS drifted: successes=%d conflicts=%d", successes, conflicts)
	}
	if sideEffects := store.SideEffects(); sideEffects.DraftWriteCount != 4 ||
		sideEffects.ExternalRepositoryWrites != 4 || sideEffects.ExecutorCallCount != 0 ||
		sideEffects.ConfirmationCallCount != 0 || sideEffects.BusinessWritebackCount != 0 ||
		sideEffects.ReplayCallCount != 0 || sideEffects.MaterializedResultReads != 0 {
		t.Fatalf("SQLite saved draft side effects drifted: %#v", sideEffects)
	}
}

func TestSavedWorkflowDraftSQLiteHTTPAndRestartNoFallback(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "radishmind.db")
	firstRuntime := openSavedWorkflowDraftSQLiteRuntime(t, databasePath)
	store := newSQLiteSavedWorkflowDraftStore(firstRuntime.DB())
	server := newSavedWorkflowDraftHTTPTestServer(true)
	t.Cleanup(server.Close)
	server.savedWorkflowDraftStore = store

	payload := validSavedWorkflowDraftPayload()
	payload.AdditionalFields = validSavedWorkflowDraftDesignerLayoutAdditionalFields()
	body := mustSavedWorkflowDraftJSON(t, savedWorkflowDraftSaveHTTPBody{
		ExpectedDraftVersion: 0,
		Draft:                savedWorkflowDraftPayloadDocumentFromDraftPayload(payload),
	})
	request := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-drafts", bytes.NewReader(body))
	setSavedWorkflowDraftDevHeaders(request, "workflow_drafts:read,workflow_drafts:write")
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	saved := decodeSavedWorkflowDraftEnvelope(t, recorder, http.StatusOK)
	if saved.FailureCode != nil || saved.Draft == nil || saved.CurrentDraftVersion != 1 {
		t.Fatalf("save SQLite draft through HTTP: %#v", saved)
	}

	var storedVersion int
	var storedCreatedAt int64
	var storedUpdatedAt int64
	var storedPayload string
	if err := firstRuntime.DB().QueryRowContext(context.Background(), `SELECT draft_version,
        created_at_unix_nano, updated_at_unix_nano, sanitized_draft_payload
        FROM saved_workflow_drafts WHERE draft_id=?`, payload.DraftID).Scan(
		&storedVersion,
		&storedCreatedAt,
		&storedUpdatedAt,
		&storedPayload,
	); err != nil || storedVersion != 1 || storedCreatedAt != storedUpdatedAt ||
		!bytes.Contains([]byte(storedPayload), []byte("designer_layout_v1")) {
		t.Fatalf("inspect SQLite saved draft physical record: version=%d created=%d updated=%d payload=%s err=%v",
			storedVersion, storedCreatedAt, storedUpdatedAt, storedPayload, err)
	}

	if err := firstRuntime.Close(); err != nil {
		t.Fatalf("close first saved draft SQLite runtime: %v", err)
	}
	requestContext := savedWorkflowDraftSQLiteContext()
	requestContext.TenantRef = "tenant_demo"
	requestContext.ActorRef = "subject_demo_user"
	requestContext.OwnerSubjectRef = "subject_demo_user"
	closedService := newSavedWorkflowDraftService(store)
	if result := closedService.ReadDraft(requestContext, ReadWorkflowDraftRequest{DraftID: payload.DraftID}); result.FailureCode != SavedWorkflowDraftFailureStoreUnavailable || result.Draft != nil {
		t.Fatalf("closed SQLite saved draft read fell back: %#v", result)
	}
	if result := closedService.ListDrafts(requestContext, ListWorkflowDraftsRequest{}); result.FailureCode != SavedWorkflowDraftFailureStoreUnavailable ||
		len(result.Summaries) != 0 {
		t.Fatalf("closed SQLite saved draft list fell back: %#v", result)
	}

	secondRuntime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   savedWorkflowDraftSQLiteMigrations(),
	})
	if err != nil {
		t.Fatalf("reopen saved draft SQLite runtime: %v", err)
	}
	t.Cleanup(func() { _ = secondRuntime.Close() })
	restartedService := newSavedWorkflowDraftService(newSQLiteSavedWorkflowDraftStore(secondRuntime.DB()))
	restored := restartedService.ReadDraft(requestContext, ReadWorkflowDraftRequest{DraftID: payload.DraftID})
	if restored.FailureCode != "" || restored.Draft == nil || restored.Draft.DraftVersion != 1 ||
		restored.Draft.Name != payload.Name {
		t.Fatalf("restore SQLite saved draft after restart: %#v", restored)
	}
	goodPayload := validSavedWorkflowDraftPayload()
	goodPayload.DraftID = "draft_good_before_corrupt_record"
	goodPayload.Name = "Good draft before corrupt record"
	if result := restartedService.SaveDraft(requestContext, SaveWorkflowDraftRequest{Payload: goodPayload}); result.FailureCode != "" || result.Draft == nil {
		t.Fatalf("create good SQLite saved draft before corruption test: %#v", result)
	}

	if _, err := secondRuntime.DB().ExecContext(context.Background(), `UPDATE saved_workflow_drafts
        SET updated_at_unix_nano=updated_at_unix_nano+1 WHERE draft_id=?`, payload.DraftID); err != nil {
		t.Fatalf("prepare SQLite saved draft timestamp drift: %v", err)
	}
	if result := restartedService.ReadDraft(requestContext, ReadWorkflowDraftRequest{DraftID: payload.DraftID}); result.FailureCode != SavedWorkflowDraftFailureStoreContractMismatch || result.Draft != nil {
		t.Fatalf("SQLite timestamp drift did not fail closed: %#v", result)
	}
	if _, err := secondRuntime.DB().ExecContext(context.Background(), `UPDATE saved_workflow_drafts
        SET updated_at_unix_nano=? WHERE draft_id=?`, storedUpdatedAt, payload.DraftID); err != nil {
		t.Fatalf("restore SQLite saved draft timestamp: %v", err)
	}
	if _, err := secondRuntime.DB().ExecContext(context.Background(), `UPDATE saved_workflow_drafts
        SET sanitized_draft_payload='{"unexpected":true}' WHERE draft_id=?`, payload.DraftID); err != nil {
		t.Fatalf("prepare SQLite saved draft contract mismatch: %v", err)
	}
	if result := restartedService.ListDrafts(requestContext, ListWorkflowDraftsRequest{}); result.FailureCode != SavedWorkflowDraftFailureStoreContractMismatch || len(result.Summaries) != 0 {
		t.Fatalf("SQLite corrupted list returned partial records: %#v", result)
	}
}

func TestSavedWorkflowDraftSQLiteRejectsSensitiveMaterialBeforePersistence(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "radishmind.db")
	runtime := openSavedWorkflowDraftSQLiteRuntime(t, databasePath)
	service := newSavedWorkflowDraftService(newSQLiteSavedWorkflowDraftStore(runtime.DB()))
	payload := validSavedWorkflowDraftPayload()
	sensitiveMarker := "sqlite-saved-draft-secret-marker"
	payload.AdditionalFields = map[string]any{"secret_value": sensitiveMarker}
	result := service.SaveDraft(savedWorkflowDraftSQLiteContext(), SaveWorkflowDraftRequest{Payload: payload})
	if result.FailureCode != SavedWorkflowDraftFailurePayloadInvalid || result.Draft != nil {
		t.Fatalf("SQLite saved draft accepted sensitive material: %#v", result)
	}
	var count int
	if err := runtime.DB().QueryRowContext(context.Background(), "SELECT count(*) FROM saved_workflow_drafts").Scan(&count); err != nil || count != 0 {
		t.Fatalf("rejected SQLite saved draft left a row: count=%d err=%v", count, err)
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("close sensitive material SQLite runtime: %v", err)
	}
	assertSQLiteFilesExcludeMarker(t, databasePath, sensitiveMarker)
}

func TestSavedWorkflowDraftSQLiteRejectsTimeOutsideNanosecondRange(t *testing.T) {
	if _, err := savedWorkflowDraftUnixNano(time.Date(2500, 1, 1, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("saved workflow draft SQLite accepted a time outside the nanosecond range")
	}
}

func savedWorkflowDraftSQLiteContext() SavedWorkflowDraftContext {
	requestContext := savedWorkflowDraftTestContext()
	requestContext.RequestContext = context.Background()
	requestContext.TenantRef = "tenant:demo"
	requestContext.OwnerSubjectRef = requestContext.ActorRef
	requestContext.ScopeGrants = []string{"workflow_drafts:read", "workflow_drafts:write"}
	return requestContext
}

func openSavedWorkflowDraftSQLiteRuntime(t *testing.T, databasePath string) *sqlitedev.Runtime {
	t.Helper()
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   savedWorkflowDraftSQLiteMigrations(),
	})
	if err != nil {
		t.Fatalf("open saved workflow draft SQLite runtime: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })
	return runtime
}
