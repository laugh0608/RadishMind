package httpapi

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/sqlitedev"
	sqliteworkflowrunmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_runs"
)

func TestWorkflowRAGExecutionSQLiteRestartReconciliationAndNoFallback(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "workflow-rag-execution.db")
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath, Migrations: sqliteworkflowrunmigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open workflow RAG execution SQLite runtime: %v", err)
	}
	runStore := newSQLiteWorkflowRunStore(runtime.DB())
	snapshotRepository := newSQLiteWorkflowRAGSnapshotRepository(runtime.DB())
	service, testBridge, runContext, snapshot, draft := workflowRAGExecutionStoreFixture(t, runStore, snapshotRepository, "run_sqliteexec000001")
	probePlan, probeFailure := buildWorkflowRAGExecutionPlan(draft, draft.DraftVersion)
	if probeFailure != "" {
		t.Fatalf("build SQLite v3 probe plan: %s", probeFailure)
	}
	probeDigest, _ := workflowRAGDraftDigest(draft)
	probe := newWorkflowRAGRunRecord(
		runContext, WorkflowRAGExecutionRequest{DraftID: draft.DraftID, DraftVersion: draft.DraftVersion, InputText: "probe", Model: "mock-rag"},
		draft, probeDigest, probePlan, snapshot, workflowRAGLexicalProfile(), workflowRAGTestSelection(),
		"run_sqliteprobe00001", time.Now().UTC(),
	)
	probeContract := cloneWorkflowRunRecord(probe)
	probeContract.RecordVersion = 1
	if payload, contractErr := marshalWorkflowRAGRunRecord(probeContract); contractErr != nil {
		t.Fatalf("marshal SQLite v3 probe: %v", contractErr)
	} else if contractErr = validateWorkflowRAGContractJSON(workflowRunRecordRAGSchemaVersion, payload); contractErr != nil {
		t.Fatalf("SQLite v3 probe contract is invalid: %v payload=%s", contractErr, payload)
	}
	if err = validateWorkflowRAGRunStoreRecord(runContext, &probe); err != nil {
		t.Fatalf("SQLite v3 probe store contract is invalid: %v", err)
	}
	if err = validateWorkflowRunStoreRecord(runContext, &probe); err != nil {
		t.Fatalf("generic SQLite v3 probe contract is invalid: %v record=%#v context=%#v", err, probe, runContext)
	}
	probePersisted := cloneWorkflowRunRecord(probe)
	probePersisted.RecordVersion = 1
	if payload, _, _, encodeErr := encodeWorkflowRunStorageRecord(probePersisted); encodeErr != nil {
		t.Fatalf("encode SQLite v3 probe: %v", encodeErr)
	} else if _, decodeErr := decodeWorkflowRunStorageRecord(runContext, payload); decodeErr != nil {
		t.Fatalf("decode SQLite v3 probe: %v payload=%s", decodeErr, payload)
	}
	if err = runStore.UpsertRun(runContext, &probe); err != nil {
		t.Fatalf("SQLite run store rejected a valid running v3 record: %v", err)
	}
	result := service.Execute(runContext, WorkflowRAGExecutionRequest{
		DraftID: draft.DraftID, DraftVersion: draft.DraftVersion, InputText: "official retrieval guidance", Model: "mock-rag",
	})
	if result.FailureCode != "" || result.Record == nil || result.Record.Status != WorkflowRunStatusSucceeded || testBridge.handleCalls != 1 {
		t.Fatalf("SQLite retrieval execution failed: %#v calls=%d", result, testBridge.handleCalls)
	}

	plan, failure := buildWorkflowRAGExecutionPlan(draft, draft.DraftVersion)
	if failure != "" {
		t.Fatalf("build SQLite stale plan: %s", failure)
	}
	draftDigest, _ := workflowRAGDraftDigest(draft)
	stale := newWorkflowRAGRunRecord(
		runContext, WorkflowRAGExecutionRequest{DraftID: draft.DraftID, DraftVersion: draft.DraftVersion, InputText: "official retrieval guidance", Model: "mock-rag"},
		draft, draftDigest, plan, snapshot, workflowRAGLexicalProfile(), workflowRAGTestSelection(), "run_sqlitestale00001",
		time.Now().UTC().Add(-workflowExecutorDefaultMaxRuntime-time.Second),
	)
	if err = runStore.UpsertRun(runContext, &stale); err != nil {
		t.Fatalf("seed SQLite stale v3 run: %v", err)
	}
	if err = runtime.Close(); err != nil {
		t.Fatalf("close workflow RAG execution SQLite runtime: %v", err)
	}

	restarted, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath, Migrations: sqliteworkflowrunmigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("restart workflow RAG execution SQLite runtime: %v", err)
	}
	restartedRunStore := newSQLiteWorkflowRunStore(restarted.DB())
	restartedRepository := newSQLiteWorkflowRAGSnapshotRepository(restarted.DB())
	recovered, found, err := restartedRunStore.ReadRun(runContext, result.Record.RunID)
	if err != nil || !found || recovered.Status != WorkflowRunStatusSucceeded || recovered.RAGAnswer != nil ||
		recovered.RetrievalAttempt == nil || len(recovered.RetrievalAttempt.CitationRefs) != 1 {
		t.Fatalf("SQLite restart did not restore metadata-only v3 success: %#v found=%t err=%v", recovered, found, err)
	}
	reconciliation, restartedBridge, _, _, _ := workflowRAGExecutionStoreFixtureFromExisting(
		t, restartedRunStore, restartedRepository, draft, "run_sqliteunused0001",
	)
	rankCalls := 0
	reconciliation.rank = func(string, []WorkflowRAGFragment, int) WorkflowRAGRankingResult {
		rankCalls++
		return WorkflowRAGRankingResult{}
	}
	reconciled := reconciliation.ReconcileStale(runContext)
	if reconciled.FailureCode != "" || reconciled.Reconciled != 1 || rankCalls != 0 || restartedBridge.handleCalls != 0 {
		t.Fatalf("SQLite restart reconciliation retried execution: %#v rank=%d gateway=%d", reconciled, rankCalls, restartedBridge.handleCalls)
	}
	recoveredStale, found, err := restartedRunStore.ReadRun(runContext, stale.RunID)
	if err != nil || !found || recoveredStale.Status != WorkflowRunStatusFailed ||
		recoveredStale.FailureCode != WorkflowRunFailureCode(WorkflowRAGFailureInterrupted) {
		t.Fatalf("SQLite stale v3 run was not closed: %#v found=%t err=%v", recoveredStale, found, err)
	}
	var auditCount int
	if err = restarted.DB().QueryRow(`SELECT count(*) FROM workflow_rag_execution_audits WHERE snapshot_id=?`, snapshot.SnapshotID).Scan(&auditCount); err != nil || auditCount != 4 {
		t.Fatalf("SQLite execution audits did not survive restart: count=%d err=%v", auditCount, err)
	}

	if err = restarted.Close(); err != nil {
		t.Fatalf("close restarted SQLite runtime: %v", err)
	}
	noFallbackService, noFallbackBridge, _, _, _ := workflowRAGExecutionStoreFixtureFromExisting(
		t, restartedRunStore, restartedRepository, draft, "run_sqlitenofallback1",
	)
	noFallback := noFallbackService.Execute(runContext, WorkflowRAGExecutionRequest{
		DraftID: draft.DraftID, DraftVersion: draft.DraftVersion, InputText: "official retrieval guidance", Model: "mock-rag",
	})
	if noFallback.FailureCode != WorkflowRunFailureCode(WorkflowRAGFailureStoreUnavailable) || noFallback.Record != nil || noFallbackBridge.handleCalls != 0 {
		t.Fatalf("closed SQLite backend fell back or called Gateway: %#v calls=%d", noFallback, noFallbackBridge.handleCalls)
	}
}

func workflowRAGExecutionStoreFixture(
	t *testing.T,
	runStore workflowRunStore,
	snapshotRepository workflowRAGSnapshotRepository,
	runID string,
) (workflowRAGExecutionService, *fakeBridge, WorkflowRunContext, WorkflowRAGSnapshotRecord, SavedWorkflowDraft) {
	t.Helper()
	snapshotContext := workflowRAGTestContext()
	created := newWorkflowRAGSnapshotService(snapshotRepository).Create(snapshotContext, WorkflowRAGSnapshotCreateInput{
		SnapshotKey: "sqlite_execution_manual", DisplayName: "SQLite execution manual",
		ContentClassification: "workspace_internal", Fragments: workflowRAGTestFragments(),
	})
	if created.FailureCode != "" || created.Record == nil {
		t.Fatalf("create store execution snapshot: %#v", created)
	}
	draft := workflowRAGEligibleDraft(created.Record.RAGRef)
	service, testBridge, runContext, _, _ := workflowRAGExecutionStoreFixtureFromExisting(t, runStore, snapshotRepository, draft, runID)
	return service, testBridge, runContext, *created.Record, draft
}

func workflowRAGExecutionStoreFixtureFromExisting(
	t *testing.T,
	runStore workflowRunStore,
	snapshotRepository workflowRAGSnapshotRepository,
	draft SavedWorkflowDraft,
	runID string,
) (workflowRAGExecutionService, *fakeBridge, WorkflowRunContext, WorkflowRAGSnapshotRecord, SavedWorkflowDraft) {
	t.Helper()
	runContext := WorkflowRunContext{
		RequestContext: context.Background(), RequestID: "request_rag_store", TenantRef: "tenant_demo",
		WorkspaceID: draft.WorkspaceID, ApplicationID: draft.ApplicationID, ActorRef: "subject_owner",
		ScopeGrants: append([]string{}, workflowRAGExecutionRequiredScopes...), AuditRef: "audit_rag_store",
	}
	reader := func(ctx SavedWorkflowDraftContext, request ReadWorkflowDraftRequest) SavedWorkflowDraftResult {
		if request.DraftID != draft.DraftID || ctx.TenantRef != runContext.TenantRef ||
			ctx.WorkspaceID != draft.WorkspaceID || ctx.ApplicationID != draft.ApplicationID {
			return SavedWorkflowDraftResult{FailureCode: SavedWorkflowDraftFailureScopeDenied}
		}
		return SavedWorkflowDraftResult{Draft: cloneSavedWorkflowDraftPointer(draft)}
	}
	testBridge := &fakeBridge{envelope: bridge.GatewayEnvelope{
		Status: "ok", Response: map[string]any{"structured_answer": workflowRAGTestAnswer("official_guide", "Official guidance supports the answer.")},
	}}
	service := newWorkflowRAGExecutionService(reader, snapshotRepository, testBridge, runStore)
	service.resolveSelection = func(context.Context, string) northboundSelection { return workflowRAGTestSelection() }
	service.envelopeOptions = func(northboundSelection, float64) bridge.EnvelopeOptions { return bridge.EnvelopeOptions{} }
	service.newRunID = func() (string, error) { return runID, nil }
	return service, testBridge, runContext, WorkflowRAGSnapshotRecord{}, draft
}
