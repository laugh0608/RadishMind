package httpapi

import (
	"context"
	"database/sql"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestWorkflowTemplateCatalogSQLiteLifecycleRestartAndCorruption(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "workflow-template-catalog.db")
	runtime := openSavedWorkflowDraftSQLiteRuntime(t, databasePath)
	fixture := newWorkflowTemplateTestFixture(t)
	fixture.ctx.RequestContext = context.Background()
	repository := newSQLiteWorkflowTemplateCatalogRepository(runtime.DB())
	draftStore := newRepositorySavedWorkflowDraftLibraryStore(newSQLiteSavedWorkflowDraftStore(runtime.DB()))
	service := newWorkflowTemplateCatalogService(repository, fixture.definitions, fixture.applications, draftStore)
	service.now = func() time.Time { return time.Date(2026, 8, 27, 11, 0, 0, 123, time.UTC) }

	created := service.CreateCandidate(fixture.ctx, workflowTemplateCandidateTestInput())
	if created.FailureCode != "" || created.Candidate == nil {
		t.Fatalf("create SQLite workflow template candidate: %#v", created)
	}
	reviewed := service.ReviewCandidate(fixture.ctx, workflowTemplateTestCandidate, WorkflowTemplateReviewInput{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: "批准 SQLite 模板候选",
	})
	if reviewed.FailureCode != "" || reviewed.Version == nil || reviewed.Version.Version != 1 {
		t.Fatalf("review SQLite workflow template candidate: %#v", reviewed)
	}
	listed := service.DecideListing(fixture.ctx, workflowTemplateTestTemplate, WorkflowTemplateListingInput{
		ExpectedPointerVersion: 0, Decision: "list", Version: 1, Reason: "上架 SQLite 模板版本",
	})
	if listed.FailureCode != "" || listed.Lineage == nil || listed.Lineage.PointerVersion != 1 {
		t.Fatalf("list SQLite workflow template version: %#v", listed)
	}
	derived := service.Derive(fixture.ctx, workflowTemplateTestTemplate, WorkflowTemplateDerivationInput{
		ExpectedPointerVersion: 1, TemplateVersion: 1, TargetApplicationID: workflowTemplateTestTargetApplication,
		DraftID: "draft_from_sqlite_template", Name: "SQLite 模板派生草案", Confirmed: true,
	})
	if derived.FailureCode != "" || derived.Draft == nil || derived.Draft.ProvenanceKind != SavedWorkflowDraftProvenanceTemplate {
		t.Fatalf("derive Saved Draft through SQLite owner: %#v", derived)
	}

	for table, expected := range map[string]int{
		"workflow_template_candidates":     1,
		"workflow_template_decisions":      1,
		"workflow_template_versions":       1,
		"workflow_template_lineages":       1,
		"workflow_template_listing_events": 1,
		"workflow_template_audits":         4,
	} {
		var count int
		if err := runtime.DB().QueryRowContext(context.Background(), "SELECT count(*) FROM "+table).Scan(&count); err != nil || count != expected {
			t.Fatalf("unexpected SQLite %s row count: count=%d err=%v", table, count, err)
		}
	}
	if _, err := runtime.DB().ExecContext(context.Background(), `UPDATE workflow_template_versions SET template_digest=template_digest WHERE template_id=?`, workflowTemplateTestTemplate); err == nil {
		t.Fatal("SQLite immutable template version accepted an update")
	}

	if err := runtime.Close(); err != nil {
		t.Fatalf("close first SQLite workflow template runtime: %v", err)
	}
	if _, err := repository.ReadCandidate(fixture.ctx, workflowTemplateTestCandidate); !errorsIsWorkflowTemplateStoreUnavailable(err) {
		t.Fatalf("closed SQLite catalog fell back instead of failing: %v", err)
	}

	restarted := openSavedWorkflowDraftSQLiteRuntime(t, databasePath)
	restartedRepository := newSQLiteWorkflowTemplateCatalogRepository(restarted.DB())
	restored, err := restartedRepository.ReadLineage(fixture.ctx, workflowTemplateTestTemplate)
	if err != nil || restored.PointerVersion != 1 || restored.ListedDigest != reviewed.Version.TemplateDigest {
		t.Fatalf("restore SQLite workflow template lineage: lineage=%#v err=%v", restored, err)
	}
	restartedDraftService := newSavedWorkflowDraftService(newSQLiteSavedWorkflowDraftStore(restarted.DB()))
	draftContext := SavedWorkflowDraftContext{
		RequestContext: context.Background(), RequestID: fixture.ctx.RequestID, TenantRef: fixture.ctx.TenantRef,
		WorkspaceID: fixture.ctx.WorkspaceID, ApplicationID: workflowTemplateTestTargetApplication,
		ActorRef: fixture.ctx.ActorRef, OwnerSubjectRef: fixture.ctx.OwnerSubjectRef,
		ScopeGrants: []string{"workflow_drafts:read"}, AuditRef: fixture.ctx.AuditRef,
	}
	draft := restartedDraftService.ReadDraft(draftContext, ReadWorkflowDraftRequest{DraftID: "draft_from_sqlite_template"})
	if draft.FailureCode != "" || draft.Draft == nil || draft.Draft.ProvenanceKind != SavedWorkflowDraftProvenanceTemplate {
		t.Fatalf("restore derived SQLite Saved Draft: %#v", draft)
	}

	if _, err := restarted.DB().ExecContext(context.Background(), `UPDATE workflow_template_candidates SET updated_at_unix_nano=updated_at_unix_nano+1 WHERE candidate_id=?`, workflowTemplateTestCandidate); err != nil {
		t.Fatalf("prepare SQLite catalog corruption: %v", err)
	}
	if _, err := restartedRepository.ReadCandidate(fixture.ctx, workflowTemplateTestCandidate); !errorsIsWorkflowTemplateStoreUnavailable(err) {
		t.Fatalf("SQLite projection corruption did not fail closed: %v", err)
	}
}

func TestWorkflowTemplateCatalogSQLiteCASCursorAndScopeIsolation(t *testing.T) {
	runtime := openSavedWorkflowDraftSQLiteRuntime(t, filepath.Join(t.TempDir(), "workflow-template-cas.db"))
	fixture := newWorkflowTemplateTestFixture(t)
	fixture.ctx.RequestContext = context.Background()
	repository := newSQLiteWorkflowTemplateCatalogRepository(runtime.DB())
	service := newWorkflowTemplateCatalogService(repository, fixture.definitions, fixture.applications, fixture.drafts)

	for index := 1; index <= 3; index++ {
		input := workflowTemplateCandidateTestInput()
		input.CandidateID = "candidate_sqlite_cursor_" + integerString(index)
		input.TemplateID = "template_sqlite_cursor_" + integerString(index)
		service.now = func() time.Time { return time.Date(2026, 8, 27, 12, index, 0, 0, time.UTC) }
		if result := service.CreateCandidate(fixture.ctx, input); result.FailureCode != "" {
			t.Fatalf("create SQLite cursor candidate %d: %#v", index, result)
		}
	}
	first := service.ListCandidates(fixture.ctx, WorkflowTemplateListInput{Limit: 2, State: workflowTemplateCandidatePending})
	if first.FailureCode != "" || len(first.Candidates) != 2 || first.NextCursor == "" {
		t.Fatalf("first SQLite keyset page drifted: %#v", first)
	}
	second := service.ListCandidates(fixture.ctx, WorkflowTemplateListInput{Cursor: first.NextCursor, Limit: 2, State: workflowTemplateCandidatePending})
	if second.FailureCode != "" || len(second.Candidates) != 1 || second.NextCursor != "" || second.Candidates[0].CandidateID == first.Candidates[0].CandidateID {
		t.Fatalf("second SQLite keyset page drifted: %#v", second)
	}

	input := workflowTemplateCandidateTestInput()
	if result := service.CreateCandidate(fixture.ctx, input); result.FailureCode != "" {
		t.Fatalf("create SQLite CAS candidate: %#v", result)
	}
	const writers = 8
	results := make(chan WorkflowTemplateCatalogResult, writers)
	var group sync.WaitGroup
	for range writers {
		group.Add(1)
		go func() {
			defer group.Done()
			results <- service.ReviewCandidate(fixture.ctx, workflowTemplateTestCandidate, WorkflowTemplateReviewInput{
				ExpectedReviewVersion: 0, Decision: "approve", Reason: "并发批准 SQLite 模板候选",
			})
		}()
	}
	group.Wait()
	close(results)
	successes := 0
	for result := range results {
		if result.FailureCode == "" {
			successes++
		} else if result.FailureCode != WorkflowTemplateFailureCandidateVersionConflict && result.FailureCode != WorkflowTemplateFailureReviewTransitionInvalid {
			t.Fatalf("unexpected SQLite review CAS result: %#v", result)
		}
	}
	if successes != 1 {
		t.Fatalf("SQLite review CAS winners drifted: %d", successes)
	}
	listingResults := make(chan WorkflowTemplateCatalogResult, writers)
	for range writers {
		group.Add(1)
		go func() {
			defer group.Done()
			listingResults <- service.DecideListing(fixture.ctx, workflowTemplateTestTemplate, WorkflowTemplateListingInput{
				ExpectedPointerVersion: 0, Decision: "list", Version: 1, Reason: "并发上架 SQLite 模板版本",
			})
		}()
	}
	group.Wait()
	close(listingResults)
	listingSuccesses := 0
	for result := range listingResults {
		if result.FailureCode == "" {
			listingSuccesses++
		} else if result.FailureCode != WorkflowTemplateFailurePointerVersionConflict {
			t.Fatalf("unexpected SQLite listing CAS result: %#v", result)
		}
	}
	if listingSuccesses != 1 {
		t.Fatalf("SQLite listing CAS winners drifted: %d", listingSuccesses)
	}

	otherWorkspace := fixture.ctx
	otherWorkspace.WorkspaceID = "workspace_other"
	listed, err := repository.ListCandidates(otherWorkspace)
	if err != nil || len(listed) != 0 {
		t.Fatalf("SQLite workspace scope leaked candidates: candidates=%#v err=%v", listed, err)
	}
}

func TestWorkflowTemplateCatalogSQLiteTransactionRollback(t *testing.T) {
	runtime := openSavedWorkflowDraftSQLiteRuntime(t, filepath.Join(t.TempDir(), "workflow-template-rollback.db"))
	fixture := newWorkflowTemplateTestFixture(t)
	fixture.ctx.RequestContext = context.Background()
	repository := newSQLiteWorkflowTemplateCatalogRepository(runtime.DB())
	service := newWorkflowTemplateCatalogService(repository, fixture.definitions, fixture.applications, fixture.drafts)
	if result := service.CreateCandidate(fixture.ctx, workflowTemplateCandidateTestInput()); result.FailureCode != "" {
		t.Fatalf("create SQLite rollback candidate: %#v", result)
	}
	if _, err := runtime.DB().ExecContext(context.Background(), `CREATE TRIGGER workflow_template_test_reject_version
BEFORE INSERT ON workflow_template_versions BEGIN SELECT RAISE(ABORT, 'reject version insert'); END`); err != nil {
		t.Fatalf("create SQLite version failure trigger: %v", err)
	}
	failedReview := service.ReviewCandidate(fixture.ctx, workflowTemplateTestCandidate, WorkflowTemplateReviewInput{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: "批准后验证事务回滚",
	})
	if failedReview.FailureCode != WorkflowTemplateFailureStoreUnavailable {
		t.Fatalf("SQLite version insert failure did not fail closed: %#v", failedReview)
	}
	candidate, err := repository.ReadCandidate(fixture.ctx, workflowTemplateTestCandidate)
	if err != nil || candidate.State != workflowTemplateCandidatePending || candidate.ReviewVersion != 0 || len(candidate.Decisions) != 0 {
		t.Fatalf("SQLite failed review committed partial candidate: candidate=%#v err=%v", candidate, err)
	}
	assertSQLiteWorkflowTemplateCounts(t, runtime.DB(), map[string]int{
		"workflow_template_decisions": 0,
		"workflow_template_versions":  0,
		"workflow_template_lineages":  0,
		"workflow_template_audits":    1,
	})
	if _, err := runtime.DB().ExecContext(context.Background(), "DROP TRIGGER workflow_template_test_reject_version"); err != nil {
		t.Fatalf("drop SQLite version failure trigger: %v", err)
	}
	approved := service.ReviewCandidate(fixture.ctx, workflowTemplateTestCandidate, WorkflowTemplateReviewInput{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: "批准 SQLite 回滚候选",
	})
	if approved.FailureCode != "" || approved.Version == nil {
		t.Fatalf("approve SQLite candidate after rollback: %#v", approved)
	}
	if _, err := runtime.DB().ExecContext(context.Background(), `CREATE TRIGGER workflow_template_test_reject_event
BEFORE INSERT ON workflow_template_listing_events BEGIN SELECT RAISE(ABORT, 'reject event insert'); END`); err != nil {
		t.Fatalf("create SQLite event failure trigger: %v", err)
	}
	failedListing := service.DecideListing(fixture.ctx, workflowTemplateTestTemplate, WorkflowTemplateListingInput{
		ExpectedPointerVersion: 0, Decision: "list", Version: 1, Reason: "验证上架事务回滚",
	})
	if failedListing.FailureCode != WorkflowTemplateFailureStoreUnavailable {
		t.Fatalf("SQLite event insert failure did not fail closed: %#v", failedListing)
	}
	lineage, err := repository.ReadLineage(fixture.ctx, workflowTemplateTestTemplate)
	if err != nil || lineage.PointerVersion != 0 || lineage.Lifecycle != workflowTemplateLineageUnlisted || len(lineage.Events) != 0 {
		t.Fatalf("SQLite failed listing committed partial pointer: lineage=%#v err=%v", lineage, err)
	}
	assertSQLiteWorkflowTemplateCounts(t, runtime.DB(), map[string]int{
		"workflow_template_listing_events": 0,
		"workflow_template_audits":         3,
	})
}

func TestWorkflowTemplateCatalogFactoryReusesSelectedWorkflowBackend(t *testing.T) {
	memoryStore := newMemorySavedWorkflowDraftStore()
	memoryRepository, err := newWorkflowTemplateCatalogRepositoryForSavedDraftStore(memoryStore)
	if err != nil {
		t.Fatalf("select memory workflow template catalog: %v", err)
	}
	if _, ok := memoryRepository.(*memoryWorkflowTemplateCatalogRepository); !ok {
		t.Fatalf("memory workflow backend selected %T", memoryRepository)
	}

	runtime := openSavedWorkflowDraftSQLiteRuntime(t, filepath.Join(t.TempDir(), "workflow-template-factory.db"))
	sqliteStore := newRepositorySavedWorkflowDraftLibraryStore(newSQLiteSavedWorkflowDraftStore(runtime.DB()))
	sqliteRepository, err := newWorkflowTemplateCatalogRepositoryForSavedDraftStore(sqliteStore)
	if err != nil {
		t.Fatalf("select SQLite workflow template catalog: %v", err)
	}
	selectedSQLite, ok := sqliteRepository.(*sqliteWorkflowTemplateCatalogRepository)
	if !ok || selectedSQLite.database != runtime.DB() {
		t.Fatalf("workflow template catalog did not reuse shared SQLite database: %T", sqliteRepository)
	}

	pool := &pgxpool.Pool{}
	postgresAdapter := NewSavedWorkflowDraftRepositoryAdapter(SavedWorkflowDraftRepositoryAdapterConfig{
		QueryExecutor: newPostgresSavedWorkflowDraftQueryExecutor(pool),
		SchemaPreflight: SavedWorkflowDraftRepositorySchemaPreflight{
			StoreSchemaVersion: savedWorkflowDraftRepositoryStoreSchemaVersion,
			MigrationState:     savedWorkflowDraftRepositoryMigrationApplied,
		},
	})
	postgresStore := newRepositorySavedWorkflowDraftLibraryStore(newRepositorySavedWorkflowDraftStore(postgresAdapter))
	postgresRepository, err := newWorkflowTemplateCatalogRepositoryForSavedDraftStore(postgresStore)
	if err != nil {
		t.Fatalf("select PostgreSQL workflow template catalog: %v", err)
	}
	selectedPostgres, ok := postgresRepository.(*postgresWorkflowTemplateCatalogRepository)
	if !ok || selectedPostgres.pool != pool {
		t.Fatalf("workflow template catalog did not reuse shared PostgreSQL pool: %T", postgresRepository)
	}

	if repository, err := newWorkflowTemplateCatalogRepositoryForSavedDraftStore(disabledSavedWorkflowDraftStore{}); err == nil || repository != nil {
		t.Fatalf("disabled workflow backend unexpectedly fell back: repository=%T err=%v", repository, err)
	}
}

func errorsIsWorkflowTemplateStoreUnavailable(err error) bool {
	return err == errWorkflowTemplateStoreUnavailable
}

func assertSQLiteWorkflowTemplateCounts(t *testing.T, database interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, expected map[string]int) {
	t.Helper()
	for table, want := range expected {
		var count int
		if err := database.QueryRowContext(context.Background(), "SELECT count(*) FROM "+table).Scan(&count); err != nil || count != want {
			t.Fatalf("unexpected SQLite %s row count: count=%d want=%d err=%v", table, count, want, err)
		}
	}
}
