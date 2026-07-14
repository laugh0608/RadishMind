package httpapi

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/sqlitedev"
	sqliteworkflowrunmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_runs"
)

func TestSQLiteWorkflowRunStoreContracts(t *testing.T) {
	t.Run("scope_filter_and_cursor", func(t *testing.T) {
		runWorkflowRunHistoryScopeFilterAndCursor(t, newSQLiteWorkflowRunStore(
			openWorkflowRunSQLiteRuntime(t, filepath.Join(t.TempDir(), "scope.db")).DB(),
		))
	})
	t.Run("concurrent_terminal_and_terminal_rewrite", func(t *testing.T) {
		runWorkflowRunStoreRejectsConcurrentOldVersionAndTerminalRewrite(t, newSQLiteWorkflowRunStore(
			openWorkflowRunSQLiteRuntime(t, filepath.Join(t.TempDir(), "terminal.db")).DB(),
		))
	})
	t.Run("forbidden_side_effects", func(t *testing.T) {
		runWorkflowRunStoreRejectsForbiddenSideEffects(t, newSQLiteWorkflowRunStore(
			openWorkflowRunSQLiteRuntime(t, filepath.Join(t.TempDir(), "side-effects.db")).DB(),
		))
	})
	t.Run("diagnostic_filters", func(t *testing.T) {
		runWorkflowRunDiagnosticFiltersAndCursorBinding(t, newSQLiteWorkflowRunStore(
			openWorkflowRunSQLiteRuntime(t, filepath.Join(t.TempDir(), "diagnostics.db")).DB(),
		))
	})
}

func TestSQLiteWorkflowRunStoreStableOrderingAndCompleteScopeIsolation(t *testing.T) {
	runtime := openWorkflowRunSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
	store := newSQLiteWorkflowRunStore(runtime.DB())
	runContext := workflowExecutorTestContext()
	startedAt := time.Date(2026, 7, 14, 9, 0, 0, 123, time.UTC)
	for _, runID := range []string{"run_equal_a", "run_equal_c", "run_equal_b"} {
		record := workflowRunHistoryTestRecord(runContext, runID, "draft_equal", startedAt)
		if err := store.UpsertRun(runContext, &record); err != nil {
			t.Fatalf("create equal-time workflow run %s: %v", runID, err)
		}
	}
	page, err := store.ListRuns(runContext, WorkflowRunListFilter{Limit: 2})
	if err != nil || len(page.Records) != 2 || !page.HasMore ||
		page.Records[0].RunID != "run_equal_c" || page.Records[1].RunID != "run_equal_b" {
		t.Fatalf("unexpected equal-time workflow run order: %#v err=%v", page, err)
	}

	for name, mutate := range map[string]func(*WorkflowRunContext){
		"tenant":      func(ctx *WorkflowRunContext) { ctx.TenantRef = "tenant_other" },
		"workspace":   func(ctx *WorkflowRunContext) { ctx.WorkspaceID = "workspace_other" },
		"application": func(ctx *WorkflowRunContext) { ctx.ApplicationID = "application_other" },
	} {
		t.Run(name, func(t *testing.T) {
			other := runContext
			mutate(&other)
			if _, found, readErr := store.ReadRun(other, "run_equal_a"); readErr != nil || found {
				t.Fatalf("cross-%s read leaked: found=%v err=%v", name, found, readErr)
			}
			scoped, listErr := store.ListRuns(other, WorkflowRunListFilter{Limit: 10})
			if listErr != nil || len(scoped.Records) != 0 {
				t.Fatalf("cross-%s list leaked: %#v err=%v", name, scoped, listErr)
			}
		})
	}
}

func TestSQLiteWorkflowRunStoreAllowsOneOfSixteenConcurrentTerminalWriters(t *testing.T) {
	runtime := openWorkflowRunSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
	store := newSQLiteWorkflowRunStore(runtime.DB())
	runContext := workflowExecutorTestContext()
	record := workflowRunHistoryTestRecord(runContext, "run_sixteen_writers", "draft_concurrent", time.Now().UTC().Add(-time.Second))
	if err := store.UpsertRun(runContext, &record); err != nil {
		t.Fatalf("create concurrent workflow run: %v", err)
	}

	var waitGroup sync.WaitGroup
	results := make(chan error, 16)
	for index := 0; index < 16; index++ {
		candidate := cloneWorkflowRunRecord(record)
		candidate.Status = WorkflowRunStatusSucceeded
		candidate.CompletedAt = workflowRunTimestamp(time.Now().UTC())
		candidate.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
		waitGroup.Add(1)
		go func(value WorkflowRunRecord) {
			defer waitGroup.Done()
			results <- store.UpsertRun(runContext, &value)
		}(candidate)
	}
	waitGroup.Wait()
	close(results)
	winners := 0
	conflicts := 0
	for err := range results {
		switch {
		case err == nil:
			winners++
		case errors.Is(err, errWorkflowRunStoreConflict):
			conflicts++
		default:
			t.Fatalf("unexpected concurrent workflow run result: %v", err)
		}
	}
	if winners != 1 || conflicts != 15 {
		t.Fatalf("expected one winner and fifteen conflicts, got winners=%d conflicts=%d", winners, conflicts)
	}
	stored, found, err := store.ReadRun(runContext, record.RunID)
	if err != nil || !found || stored.RecordVersion != 2 || stored.Status != WorkflowRunStatusSucceeded {
		t.Fatalf("unexpected concurrent workflow run winner: found=%v record=%#v err=%v", found, stored, err)
	}
}

func TestSQLiteWorkflowRunStoreExecutorRestartNoFallbackAndSensitiveInputBoundary(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "radishmind.db")
	firstRuntime := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	store := newSQLiteWorkflowRunStore(firstRuntime.DB())
	runContext := workflowExecutorTestContext()
	sensitiveMarker := "private-workflow-input-must-not-persist-7f78e3"
	draft := executableWorkflowDraftForTest()
	service := workflowExecutorTestService(draft, &workflowExecutorTestBridge{}, store)
	result := service.StartRun(runContext, WorkflowRunRequest{DraftID: draft.DraftID, InputText: sensitiveMarker})
	if result.FailureCode != "" || result.Record == nil || result.Record.Status != WorkflowRunStatusSucceeded ||
		result.Record.SideEffects.ProviderCalls != 1 || result.Record.SideEffects.ToolCalls != 0 ||
		result.Record.SideEffects.ConfirmationCalls != 0 || result.Record.SideEffects.BusinessWrites != 0 ||
		result.Record.SideEffects.ReplayWrites != 0 {
		t.Fatalf("execute SQLite-backed workflow run: %#v", result)
	}
	runID := result.Record.RunID
	if err := firstRuntime.Close(); err != nil {
		t.Fatalf("close first workflow run SQLite runtime: %v", err)
	}
	if closed := service.ListRuns(runContext, WorkflowRunListRequest{}); closed.FailureCode != WorkflowRunFailureStoreUnavailable {
		t.Fatalf("closed workflow run store must not fall back to memory: %#v", closed)
	}
	assertSQLiteFilesExcludeMarker(t, databasePath, sensitiveMarker)

	secondRuntime := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	secondStore := newSQLiteWorkflowRunStore(secondRuntime.DB())
	recovered, found, err := secondStore.ReadRun(runContext, runID)
	if err != nil || !found || recovered.Status != WorkflowRunStatusSucceeded || recovered.RecordVersion < 2 {
		t.Fatalf("recover SQLite workflow run after restart: found=%v record=%#v err=%v", found, recovered, err)
	}
	restartedService := newWorkflowExecutorService(nil, nil, secondStore)
	page := restartedService.ListRuns(runContext, WorkflowRunListRequest{Limit: 10, Status: WorkflowRunStatusSucceeded})
	if page.FailureCode != "" || len(page.Runs) != 1 || page.Runs[0].RunID != runID {
		t.Fatalf("list recovered SQLite workflow run: %#v", page)
	}

	healthy := workflowRunHistoryTestRecord(runContext, "run_healthy_after_restart", "draft_healthy", time.Now().UTC().Add(time.Second))
	if err := secondStore.UpsertRun(runContext, &healthy); err != nil {
		t.Fatalf("create healthy workflow run after restart: %v", err)
	}
	if _, err := secondRuntime.DB().ExecContext(context.Background(), `UPDATE workflow_run_records
        SET selected_model='physical-column-drift' WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND run_id=?`,
		runContext.TenantRef, runContext.WorkspaceID, runContext.ApplicationID, runID,
	); err != nil {
		t.Fatalf("inject workflow run physical-column drift: %v", err)
	}
	if _, _, err := secondStore.ReadRun(runContext, runID); !errors.Is(err, errWorkflowRunStoreContract) {
		t.Fatalf("workflow run physical-column drift was not rejected: %v", err)
	}
	if partial, err := secondStore.ListRuns(runContext, WorkflowRunListFilter{Limit: 10}); !errors.Is(err, errWorkflowRunStoreContract) || len(partial.Records) != 0 {
		t.Fatalf("corrupted workflow run list returned partial data: page=%#v err=%v", partial, err)
	}
	if err := secondRuntime.Close(); err != nil {
		t.Fatalf("close second workflow run SQLite runtime: %v", err)
	}
	assertSQLiteFilesExcludeMarker(t, databasePath, sensitiveMarker)
}

func TestSQLiteWorkflowRunStoreRejectsTimeOutsideNanosecondRange(t *testing.T) {
	outsideRange := time.Date(2500, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, err := workflowRunUnixNano(outsideRange); !errors.Is(err, errWorkflowRunStoreContract) {
		t.Fatalf("workflow run time outside SQLite nanosecond range was accepted: %v", err)
	}

	runtime := openWorkflowRunSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
	store := newSQLiteWorkflowRunStore(runtime.DB())
	runContext := workflowExecutorTestContext()
	record := workflowRunHistoryTestRecord(runContext, "run_time_outside_range", "draft_time", outsideRange)
	if err := store.UpsertRun(runContext, &record); !errors.Is(err, errWorkflowRunStoreContract) {
		t.Fatalf("workflow run outside SQLite time range was persisted: %v", err)
	}
}

func TestSQLiteWorkflowRunStoreRejectsUnknownStoredDocumentField(t *testing.T) {
	runtime := openWorkflowRunSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
	store := newSQLiteWorkflowRunStore(runtime.DB())
	runContext := workflowExecutorTestContext()
	record := workflowRunHistoryTestRecord(runContext, "run_unknown_field", "draft_unknown_field", time.Now().UTC())
	if err := store.UpsertRun(runContext, &record); err != nil {
		t.Fatalf("create workflow run for strict document check: %v", err)
	}
	if _, err := runtime.DB().ExecContext(context.Background(), `UPDATE workflow_run_records
        SET sanitized_run_record=json_set(sanitized_run_record, '$.unexpected_field', 'rejected')
        WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND run_id=?`,
		runContext.TenantRef, runContext.WorkspaceID, runContext.ApplicationID, record.RunID,
	); err != nil {
		t.Fatalf("inject unknown stored workflow run field: %v", err)
	}
	if _, _, err := store.ReadRun(runContext, record.RunID); !errors.Is(err, errWorkflowRunStoreContract) {
		t.Fatalf("unknown stored workflow run field was accepted: %v", err)
	}
}

func openWorkflowRunSQLiteRuntime(t *testing.T, databasePath string) *sqlitedev.Runtime {
	t.Helper()
	runtime := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	t.Cleanup(func() { _ = runtime.Close() })
	return runtime
}

func openWorkflowRunSQLiteRuntimeWithoutCleanup(t *testing.T, databasePath string) *sqlitedev.Runtime {
	t.Helper()
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   sqliteworkflowrunmigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open workflow run SQLite runtime: %v", err)
	}
	return runtime
}
