package httpapi

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/sqlitedev"
	sqliteworkflowrunmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_runs"
)

func TestWorkflowRAGEvaluationSQLiteLifecycleRestartAndNoFallback(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "workflow-rag-evaluation.db")
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{DatabasePath: databasePath, Migrations: sqliteworkflowrunmigrations.Migrations()})
	if err != nil {
		t.Fatalf("open workflow RAG evaluation SQLite runtime: %v", err)
	}
	ctx := workflowRAGTestContext()
	snapshots := newSQLiteWorkflowRAGSnapshotRepository(runtime.DB())
	baseline := createWorkflowRAGEvaluationTestSnapshot(t, snapshots, ctx, "rags_dddddddddddddddd", "sqlite_baseline", "public", []WorkflowRAGFragmentInput{{
		FragmentRef: "official_guide", SourceType: "manual", SourceRef: "manual.sqlite", PageSlug: "sqlite/guide",
		Title: "SQLite guide", IsOfficial: true, Content: "approved sqlite pressure guidance",
	}})
	candidate := createWorkflowRAGEvaluationTestSnapshot(t, snapshots, ctx, "rags_eeeeeeeeeeeeeeee", "sqlite_candidate", "public", []WorkflowRAGFragmentInput{{
		FragmentRef: "candidate_note", SourceType: "wiki", SourceRef: "wiki.sqlite", PageSlug: "sqlite/note",
		Title: "SQLite note", Content: "unrelated sqlite note",
	}})
	repository := newSQLiteWorkflowRAGEvaluationDatasetRepository(runtime.DB())
	service := newWorkflowRAGEvaluationDatasetService(repository, snapshots)
	now := time.Date(2026, 7, 18, 9, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	service.newID = workflowRAGEvaluationTestIDGenerator()
	created := service.Create(ctx, WorkflowRAGEvaluationDatasetCreateInput{
		DatasetKey: "sqlite_pressure_review", DisplayName: "SQLite pressure review", ContentClassification: "synthetic_public",
		BaselineSnapshot: workflowRAGEvaluationSnapshotBinding(baseline), Thresholds: workflowRAGEvaluationPerfectThresholds(),
		ReviewSummary: "SQLite baseline reviewed.", Samples: []WorkflowRAGEvaluationSample{{
			SampleID: "sqlite_pressure", QueryText: "approved sqlite pressure guidance", Expectation: "evidence_required",
			ExpectedCitationRefs: []string{"official_guide"}, RequiredOfficialRefs: []string{"official_guide"}, TopK: 1, ReviewNote: "Use official evidence.",
		}},
	})
	if created.FailureCode != "" || created.Resource == nil || created.Version == nil {
		t.Fatalf("create SQLite evaluation dataset: %#v", created)
	}

	now = now.Add(time.Minute)
	ctx.RequestID, ctx.AuditRef = "request_sqlite_review", "audit_sqlite_review"
	reviewed := service.CreateCandidateReview(ctx, created.Resource.DatasetID, WorkflowRAGCandidateReviewInput{
		DatasetVersion: 1, DatasetDigest: created.Version.Dataset.DatasetDigest, CandidateSnapshot: workflowRAGEvaluationSnapshotBinding(candidate),
	})
	if reviewed.FailureCode != "" || reviewed.Review == nil || reviewed.Review.Conclusion != "regressed" {
		t.Fatalf("create SQLite candidate review: %#v", reviewed)
	}

	now = now.Add(time.Minute)
	ctx.RequestID, ctx.AuditRef = "request_sqlite_version", "audit_sqlite_version"
	versioned := service.Version(ctx, created.Resource.DatasetID, WorkflowRAGEvaluationDatasetVersionInput{
		ExpectedLatestVersion: 1, DisplayName: "SQLite pressure review v2", ContentClassification: "synthetic_public",
		BaselineSnapshot: workflowRAGEvaluationSnapshotBinding(baseline), Thresholds: workflowRAGEvaluationPerfectThresholds(),
		ReviewSummary: "SQLite second review.", Samples: created.Version.Dataset.Samples,
	})
	if versioned.FailureCode != "" || versioned.Version == nil || versioned.Version.Dataset.DatasetVersion != 2 {
		t.Fatalf("version SQLite evaluation dataset: %#v", versioned)
	}
	if stale := service.Version(ctx, created.Resource.DatasetID, WorkflowRAGEvaluationDatasetVersionInput{ExpectedLatestVersion: 1}); stale.FailureCode != WorkflowRAGEvaluationFailureVersionConflict {
		t.Fatalf("SQLite stale CAS did not fail closed: %#v", stale)
	}

	now = now.Add(time.Minute)
	ctx.RequestID, ctx.AuditRef = "request_sqlite_archive", "audit_sqlite_archive"
	archived := service.Archive(ctx, created.Resource.DatasetID, 2)
	if archived.FailureCode != "" || archived.Resource == nil || archived.Resource.LifecycleState != workflowRAGEvaluationArchived {
		t.Fatalf("archive SQLite evaluation dataset: %#v", archived)
	}
	if err = runtime.Close(); err != nil {
		t.Fatalf("close workflow RAG evaluation SQLite runtime: %v", err)
	}

	restarted, err := sqlitedev.Open(context.Background(), sqlitedev.Options{DatabasePath: databasePath, Migrations: sqliteworkflowrunmigrations.Migrations()})
	if err != nil {
		t.Fatalf("restart workflow RAG evaluation SQLite runtime: %v", err)
	}
	t.Cleanup(func() { _ = restarted.Close() })
	restartedService := newWorkflowRAGEvaluationDatasetService(
		newSQLiteWorkflowRAGEvaluationDatasetRepository(restarted.DB()),
		newSQLiteWorkflowRAGSnapshotRepository(restarted.DB()),
	)
	restored := restartedService.Read(ctx, created.Resource.DatasetID, 1)
	if restored.FailureCode != "" || restored.Version == nil || restored.Version.Dataset.DatasetVersion != 1 ||
		restored.Resource == nil || restored.Resource.LifecycleState != workflowRAGEvaluationArchived ||
		restored.Version.Dataset.Samples[0].QueryText != "approved sqlite pressure guidance" {
		t.Fatalf("SQLite restart did not restore exact immutable dataset version: %#v", restored)
	}
	restoredReview := restartedService.ReadCandidateReview(ctx, created.Resource.DatasetID, reviewed.Review.ReviewID)
	if restoredReview.FailureCode != "" || restoredReview.Review == nil || restoredReview.Review.Conclusion != "regressed" {
		t.Fatalf("SQLite restart did not restore candidate review: %#v", restoredReview)
	}
	otherScope := ctx
	otherScope.ApplicationID = "app_other"
	if denied := restartedService.Read(otherScope, created.Resource.DatasetID, 1); denied.FailureCode != WorkflowRAGEvaluationFailureScopeDenied {
		t.Fatalf("SQLite evaluation store accepted cross-scope read: %#v", denied)
	}

	var versionCount, reviewCount, auditCount int
	if err = restarted.DB().QueryRow(`SELECT count(*) FROM workflow_rag_evaluation_dataset_versions`).Scan(&versionCount); err != nil || versionCount != 2 {
		t.Fatalf("unexpected SQLite evaluation version count: count=%d err=%v", versionCount, err)
	}
	if err = restarted.DB().QueryRow(`SELECT count(*) FROM workflow_rag_candidate_snapshot_reviews`).Scan(&reviewCount); err != nil || reviewCount != 1 {
		t.Fatalf("unexpected SQLite candidate review count: count=%d err=%v", reviewCount, err)
	}
	if err = restarted.DB().QueryRow(`SELECT count(*) FROM workflow_rag_evaluation_audits`).Scan(&auditCount); err != nil || auditCount != 4 {
		t.Fatalf("unexpected SQLite evaluation audit count: count=%d err=%v", auditCount, err)
	}
	if _, err = restarted.DB().Exec(`UPDATE workflow_rag_evaluation_dataset_versions SET dataset_version=3 WHERE dataset_version=2`); err == nil {
		t.Fatal("SQLite accepted mutation of immutable evaluation dataset version")
	}
	if _, err = restarted.DB().Exec(`DELETE FROM workflow_rag_candidate_snapshot_reviews`); err == nil {
		t.Fatal("SQLite accepted deletion of append-only candidate review")
	}
	if _, err = restarted.DB().Exec(`UPDATE workflow_rag_evaluation_audits SET event_kind='dataset_created'`); err == nil {
		t.Fatal("SQLite accepted mutation of append-only evaluation audit")
	}
	if _, err = restarted.DB().Exec(`UPDATE workflow_rag_evaluation_dataset_resources SET sanitized_resource_payload='{}' WHERE dataset_id=?`, created.Resource.DatasetID); err != nil {
		t.Fatalf("seed corrupt SQLite evaluation resource: %v", err)
	}
	_, _, err = newSQLiteWorkflowRAGEvaluationDatasetRepository(restarted.DB()).ReadLatest(ctx, created.Resource.DatasetID)
	if !errors.Is(err, errWorkflowRAGEvaluationContract) {
		t.Fatalf("SQLite corrupt record did not fail closed without fallback: %v", err)
	}
}
