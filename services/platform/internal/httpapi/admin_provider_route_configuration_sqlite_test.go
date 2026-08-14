package httpapi

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"radishmind.local/services/platform/internal/sqlitedev"
	sqliteadminproviderroutemigrations "radishmind.local/services/platform/migrations/sqlite/admin_provider_routes"
)

func TestAdminProviderRouteSQLitePersistsLifecycleAcrossRestart(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "admin-provider-routes.db")
	first := openAdminProviderRouteSQLiteRuntime(t, databasePath)
	resolver := newFakeAdminProviderInventoryResolver()
	service := newAdminProviderRouteService(newSQLiteAdminProviderRouteRepository(first.DB()), resolver)
	ctx := adminProviderRouteTestContext()

	adminProviderRoutePrepareCandidate(t, service, ctx, "candidate-one")
	approval := service.ReviewCandidate(ctx, "gateway-default", "candidate-one", AdminProviderRouteReviewInput{
		ExpectedReviewVersion: 0,
		Decision:              adminProviderRouteDecisionApprove,
		Reason:                "Approve the durable provider route candidate.",
	})
	if approval.FailureCode != "" {
		t.Fatalf("approve SQLite candidate: %#v", approval)
	}
	activation := service.Activate(ctx, AdminProviderRouteActivationInput{
		ConfigurationID:    "gateway-default",
		CandidateID:        "candidate-one",
		ExpectedGeneration: 0,
		Action:             adminProviderRouteActionActivate,
		Reason:             "Activate the durable provider route candidate.",
	})
	if activation.FailureCode != "" || activation.Snapshot == nil || activation.Snapshot.Generation != 1 {
		t.Fatalf("activate SQLite candidate: %#v", activation)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("close first SQLite runtime: %v", err)
	}

	restarted := openAdminProviderRouteSQLiteRuntime(t, databasePath)
	restartedService := newAdminProviderRouteService(
		newSQLiteAdminProviderRouteRepository(restarted.DB()),
		resolver,
	)
	draft := restartedService.ReadDraft(ctx, "gateway-default")
	candidate := restartedService.ReadCandidate(ctx, "gateway-default", "candidate-one")
	snapshot := restartedService.ReadActiveSnapshot(ctx, "gateway-default")
	history := restartedService.ListActivations(ctx, "gateway-default")
	if draft.FailureCode != "" || draft.Draft == nil || draft.Draft.DraftRevision != 1 {
		t.Fatalf("restore SQLite draft: %#v", draft)
	}
	if candidate.FailureCode != "" || candidate.Candidate == nil ||
		candidate.Candidate.CandidateState != adminProviderRouteCandidateApproved ||
		candidate.Candidate.Review == nil {
		t.Fatalf("restore SQLite candidate and review: %#v", candidate)
	}
	if snapshot.FailureCode != "" || snapshot.Snapshot == nil || snapshot.Snapshot.Generation != 1 {
		t.Fatalf("restore SQLite snapshot: %#v", snapshot)
	}
	if history.FailureCode != "" || len(history.ActivationHistory) != 1 ||
		history.ActivationHistory[0].RecordDigest != activation.Activation.RecordDigest {
		t.Fatalf("restore SQLite activation history: %#v", history)
	}
	secondDraftInput := adminProviderRouteTestDraftInput(1, "mock-secondary")
	secondDraftInput.DisplayName = "SQLite replacement routing"
	secondDraft := restartedService.PutDraft(ctx, secondDraftInput)
	if secondDraft.FailureCode != "" || secondDraft.Draft == nil || secondDraft.Draft.DraftRevision != 2 {
		t.Fatalf("save replacement SQLite draft: %#v", secondDraft)
	}
	adminProviderRoutePrepareCandidateWithRevision(
		t,
		restartedService,
		ctx,
		"candidate-two",
		secondDraft.Draft.DraftRevision,
	)
	if reviewed := restartedService.ReviewCandidate(
		ctx,
		"gateway-default",
		"candidate-two",
		AdminProviderRouteReviewInput{
			ExpectedReviewVersion: 0,
			Decision:              adminProviderRouteDecisionApprove,
			Reason:                "Approve the replacement SQLite provider route candidate.",
		},
	); reviewed.FailureCode != "" {
		t.Fatalf("approve replacement SQLite candidate: %#v", reviewed)
	}
	if activated := restartedService.Activate(ctx, AdminProviderRouteActivationInput{
		ConfigurationID:    "gateway-default",
		CandidateID:        "candidate-two",
		ExpectedGeneration: 1,
		Action:             adminProviderRouteActionActivate,
		Reason:             "Activate the replacement SQLite provider route candidate.",
	}); activated.FailureCode != "" || activated.Snapshot == nil || activated.Snapshot.Generation != 2 {
		t.Fatalf("activate replacement SQLite candidate: %#v", activated)
	}
	if rollback := restartedService.Activate(ctx, AdminProviderRouteActivationInput{
		ConfigurationID:    "gateway-default",
		CandidateID:        "candidate-one",
		ExpectedGeneration: 2,
		Action:             adminProviderRouteActionRollback,
		Reason:             "Restore the previous SQLite provider route candidate.",
	}); rollback.FailureCode != "" || rollback.Snapshot == nil || rollback.Snapshot.Generation != 3 {
		t.Fatalf("rollback SQLite candidate: %#v", rollback)
	}
	if err := restarted.Close(); err != nil {
		t.Fatalf("close second SQLite runtime: %v", err)
	}
	finalRuntime := openAdminProviderRouteSQLiteRuntime(t, databasePath)
	t.Cleanup(func() { _ = finalRuntime.Close() })
	finalService := newAdminProviderRouteService(
		newSQLiteAdminProviderRouteRepository(finalRuntime.DB()),
		resolver,
	)
	finalSnapshot := finalService.ReadActiveSnapshot(ctx, "gateway-default")
	finalHistory := finalService.ListActivations(ctx, "gateway-default")
	if finalSnapshot.FailureCode != "" || finalSnapshot.Snapshot == nil ||
		finalSnapshot.Snapshot.Generation != 3 || finalSnapshot.Snapshot.CandidateID != "candidate-one" {
		t.Fatalf("restore final SQLite snapshot: %#v", finalSnapshot)
	}
	if finalHistory.FailureCode != "" || len(finalHistory.ActivationHistory) != 3 ||
		finalHistory.ActivationHistory[2].Action != adminProviderRouteActionRollback {
		t.Fatalf("restore final SQLite activation history: %#v", finalHistory)
	}

	for _, storagePath := range []string{databasePath, databasePath + "-wal", databasePath + "-shm"} {
		payload, err := os.ReadFile(storagePath)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatalf("read SQLite storage for privacy scan: %v", err)
		}
		for _, forbidden := range [][]byte{
			[]byte("sk-admin-provider-route-secret"),
			[]byte("https://provider.example.invalid/v1"),
			[]byte("Authorization: Bearer"),
		} {
			if bytes.Contains(payload, forbidden) {
				t.Fatalf("SQLite storage %s contains forbidden material %q", storagePath, forbidden)
			}
		}
	}
}

func TestAdminProviderRouteSQLiteV2PersistsAcrossRestartAndKeepsV1Readable(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "admin-provider-routes-v2.db")
	first := openAdminProviderRouteSQLiteRuntime(t, databasePath)
	service := newAdminProviderRouteService(
		newSQLiteAdminProviderRouteRepository(first.DB()),
		newFakeAdminProviderInventoryResolver(),
	)
	ctx := adminProviderRouteTestContext()
	v2Input := adminProviderRouteV2TestDraftInput(0)
	draft := service.PutDraft(ctx, v2Input)
	if draft.FailureCode != "" || draft.Draft == nil ||
		draft.Draft.SchemaVersion != adminProviderRouteDraftSchemaVersionV2 {
		t.Fatalf("persist SQLite v2 draft: %#v", draft)
	}
	candidate := service.CreateCandidate(ctx, AdminProviderRouteCandidateInput{
		ConfigurationID:       v2Input.ConfigurationID,
		CandidateID:           "candidate-sqlite-v2",
		ExpectedDraftRevision: 1,
	})
	if candidate.FailureCode != "" || candidate.Candidate == nil ||
		candidate.Candidate.SchemaVersion != adminProviderRouteCandidateSchemaVersionV2 {
		t.Fatalf("persist SQLite v2 candidate: %#v", candidate)
	}
	if reviewed := service.ReviewCandidate(ctx, v2Input.ConfigurationID, "candidate-sqlite-v2", AdminProviderRouteReviewInput{
		ExpectedReviewVersion: 0,
		Decision:              adminProviderRouteDecisionApprove,
		Reason:                "Approve the durable SQLite fallback route.",
	}); reviewed.FailureCode != "" {
		t.Fatalf("approve SQLite v2 candidate: %#v", reviewed)
	}
	activated := service.Activate(ctx, AdminProviderRouteActivationInput{
		ConfigurationID:    v2Input.ConfigurationID,
		CandidateID:        "candidate-sqlite-v2",
		ExpectedGeneration: 0,
		Action:             adminProviderRouteActionActivate,
		Reason:             "Activate the durable SQLite fallback route.",
	})
	if activated.FailureCode != "" || activated.Snapshot == nil ||
		activated.Snapshot.SchemaVersion != adminProviderRouteSnapshotSchemaVersionV2 {
		t.Fatalf("activate SQLite v2 candidate: %#v", activated)
	}
	v1 := service.PutDraft(ctx, adminProviderRouteTestDraftInput(0, "mock-primary"))
	if v1.FailureCode != "" || v1.Draft == nil || v1.Draft.SchemaVersion != adminProviderRouteDraftSchemaVersion {
		t.Fatalf("persist SQLite v1 compatibility record: %#v", v1)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("close first SQLite v2 runtime: %v", err)
	}

	restarted := openAdminProviderRouteSQLiteRuntime(t, databasePath)
	t.Cleanup(func() { _ = restarted.Close() })
	restartedService := newAdminProviderRouteService(
		newSQLiteAdminProviderRouteRepository(restarted.DB()),
		newFakeAdminProviderInventoryResolver(),
	)
	restored := restartedService.ReadActiveSnapshot(ctx, v2Input.ConfigurationID)
	if restored.FailureCode != "" || restored.Snapshot == nil ||
		restored.Snapshot.SchemaVersion != adminProviderRouteSnapshotSchemaVersionV2 ||
		restored.Snapshot.Configuration.ModelRoutes[0].ExecutionMode != AdminProviderRouteExecutionSequentialFallback ||
		len(restored.Snapshot.Configuration.ModelRoutes[0].AttemptTargets) != 2 {
		t.Fatalf("restore SQLite v2 snapshot: %#v", restored)
	}
	restoredV1 := restartedService.ReadDraft(ctx, "gateway-default")
	if restoredV1.FailureCode != "" || restoredV1.Draft == nil ||
		restoredV1.Draft.SchemaVersion != adminProviderRouteDraftSchemaVersion {
		t.Fatalf("restore SQLite v1 compatibility record: %#v", restoredV1)
	}
}

func TestAdminProviderRouteSQLiteConcurrentDraftCAS(t *testing.T) {
	runtime := openAdminProviderRouteSQLiteRuntime(t, filepath.Join(t.TempDir(), "admin-provider-routes.db"))
	t.Cleanup(func() { _ = runtime.Close() })
	service := newAdminProviderRouteService(
		newSQLiteAdminProviderRouteRepository(runtime.DB()),
		newFakeAdminProviderInventoryResolver(),
	)
	ctx := adminProviderRouteTestContext()
	if created := service.PutDraft(ctx, adminProviderRouteTestDraftInput(0, "mock-primary")); created.FailureCode != "" {
		t.Fatalf("create initial SQLite draft: %#v", created)
	}

	const writers = 12
	start := make(chan struct{})
	results := make(chan AdminProviderRouteResult, writers)
	var waitGroup sync.WaitGroup
	for index := 0; index < writers; index++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			results <- service.PutDraft(ctx, adminProviderRouteTestDraftInput(1, "mock-secondary"))
		}()
	}
	close(start)
	waitGroup.Wait()
	close(results)

	successes := 0
	conflicts := 0
	for result := range results {
		switch result.FailureCode {
		case "":
			successes++
		case AdminProviderRouteFailureDraftRevisionConflict:
			conflicts++
		default:
			t.Fatalf("unexpected concurrent SQLite result: %#v", result)
		}
	}
	if successes != 1 || conflicts != writers-1 {
		t.Fatalf("SQLite draft CAS results: successes=%d conflicts=%d", successes, conflicts)
	}
	if current := service.ReadDraft(ctx, "gateway-default"); current.FailureCode != "" ||
		current.Draft == nil || current.Draft.DraftRevision != 2 {
		t.Fatalf("read SQLite draft after CAS: %#v", current)
	}
}

func TestAdminProviderRouteSQLiteProtectsAppendOnlyReview(t *testing.T) {
	runtime := openAdminProviderRouteSQLiteRuntime(t, filepath.Join(t.TempDir(), "admin-provider-routes.db"))
	t.Cleanup(func() { _ = runtime.Close() })
	service := newAdminProviderRouteService(
		newSQLiteAdminProviderRouteRepository(runtime.DB()),
		newFakeAdminProviderInventoryResolver(),
	)
	ctx := adminProviderRouteTestContext()
	adminProviderRoutePrepareCandidate(t, service, ctx, "candidate-review-drift")
	if reviewed := service.ReviewCandidate(
		ctx,
		"gateway-default",
		"candidate-review-drift",
		AdminProviderRouteReviewInput{
			ExpectedReviewVersion: 0,
			Decision:              adminProviderRouteDecisionApprove,
			Reason:                "Approve before deliberate storage drift.",
		},
	); reviewed.FailureCode != "" {
		t.Fatalf("approve candidate before append-only check: %#v", reviewed)
	}
	if _, err := runtime.DB().ExecContext(context.Background(), `DELETE FROM admin_provider_route_reviews
        WHERE tenant_ref=? AND workspace_id=? AND environment=? AND configuration_id=? AND candidate_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, "gateway-default", "candidate-review-drift"); err == nil {
		t.Fatal("SQLite append-only review accepted deletion")
	}
	if result := service.ReadCandidate(ctx, "gateway-default", "candidate-review-drift"); result.FailureCode != "" || result.Candidate == nil || result.Candidate.Review == nil {
		t.Fatalf("protected SQLite review was not readable: %#v", result)
	}
}

func openAdminProviderRouteSQLiteRuntime(t *testing.T, databasePath string) *sqlitedev.Runtime {
	t.Helper()
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   sqliteadminproviderroutemigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open admin provider route SQLite runtime: %v", err)
	}
	return runtime
}
