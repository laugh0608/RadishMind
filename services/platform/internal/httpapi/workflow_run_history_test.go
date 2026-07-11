package httpapi

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestMemoryWorkflowRunHistoryScopeFilterAndCursor(t *testing.T) {
	store := newMemoryWorkflowRunStore(20)
	runContext := workflowExecutorTestContext()
	base := time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC)
	for index, status := range []WorkflowRunStatus{WorkflowRunStatusSucceeded, WorkflowRunStatusFailed, WorkflowRunStatusSucceeded} {
		record := workflowRunHistoryTestRecord(runContext, "run_"+string(rune('a'+index)), "draft_a", base.Add(time.Duration(index)*time.Minute))
		if err := store.UpsertRun(runContext, &record); err != nil {
			t.Fatal(err)
		}
		record.Status = status
		record.CompletedAt = workflowRunTimestamp(base.Add(time.Duration(index)*time.Minute + time.Second))
		record.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
		if err := store.UpsertRun(runContext, &record); err != nil {
			t.Fatal(err)
		}
	}
	service := newWorkflowExecutorService(nil, nil, store)
	first := service.ListRuns(runContext, WorkflowRunListRequest{Limit: 2})
	if first.FailureCode != "" || len(first.Runs) != 2 || !first.HasMore || first.NextCursor == "" || first.Runs[0].RunID != "run_c" {
		t.Fatalf("unexpected first page: %#v", first)
	}
	second := service.ListRuns(runContext, WorkflowRunListRequest{Limit: 2, Cursor: first.NextCursor})
	if second.FailureCode != "" || len(second.Runs) != 1 || second.Runs[0].RunID != "run_a" {
		t.Fatalf("unexpected second page: %#v", second)
	}
	filtered := service.ListRuns(runContext, WorkflowRunListRequest{Status: WorkflowRunStatusFailed})
	if len(filtered.Runs) != 1 || filtered.Runs[0].RunID != "run_b" {
		t.Fatalf("unexpected filter: %#v", filtered)
	}
	if changed := service.ListRuns(runContext, WorkflowRunListRequest{Limit: 3, Cursor: first.NextCursor}); changed.FailureCode != WorkflowRunFailureCursorInvalid {
		t.Fatalf("cursor must bind filters: %#v", changed)
	}
	other := runContext
	other.TenantRef = "tenant_other"
	if scoped := service.ListRuns(other, WorkflowRunListRequest{}); len(scoped.Runs) != 0 {
		t.Fatalf("cross-tenant history leaked: %#v", scoped)
	}
}

func TestMemoryWorkflowRunStoreRejectsConcurrentOldVersionAndTerminalRewrite(t *testing.T) {
	store := newMemoryWorkflowRunStore(10)
	runContext := workflowExecutorTestContext()
	record := workflowRunHistoryTestRecord(runContext, "run_concurrent", "draft_a", time.Now().UTC())
	if err := store.UpsertRun(runContext, &record); err != nil {
		t.Fatal(err)
	}
	left, right := cloneWorkflowRunRecord(record), cloneWorkflowRunRecord(record)
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, candidate := range []*WorkflowRunRecord{&left, &right} {
		wg.Add(1)
		go func(value *WorkflowRunRecord) {
			defer wg.Done()
			value.Status = WorkflowRunStatusSucceeded
			value.CompletedAt = workflowRunTimestamp(time.Now())
			value.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
			results <- store.UpsertRun(runContext, value)
		}(candidate)
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("exactly one concurrent terminal update must succeed: %d", successes)
	}
	stored, _, _ := store.ReadRun(runContext, record.RunID)
	stored.Status = WorkflowRunStatusRunning
	stored.CompletedAt = ""
	stored.Diagnostic.TerminalWriteState = WorkflowRunTerminalWritePending
	if err := store.UpsertRun(runContext, &stored); err != errWorkflowRunStoreConflict {
		t.Fatalf("terminal rewrite must fail: %v", err)
	}
}

func TestWorkflowRunStoreRejectsForbiddenSideEffects(t *testing.T) {
	store := newMemoryWorkflowRunStore(10)
	runContext := workflowExecutorTestContext()
	record := workflowRunHistoryTestRecord(runContext, "run_forbidden", "draft_a", time.Now().UTC())
	record.SideEffects.ToolCalls = 1
	if err := store.UpsertRun(runContext, &record); err != errWorkflowRunStoreContract {
		t.Fatalf("forbidden side effect accepted: %v", err)
	}
	record = workflowRunHistoryTestRecord(runContext, "run_endpoint", "draft_a", time.Now().UTC())
	record.SelectedProvider = "https://provider.invalid/raw"
	if err := store.UpsertRun(runContext, &record); err != errWorkflowRunStoreContract {
		t.Fatalf("provider endpoint accepted: %v", err)
	}
}

func workflowRunHistoryTestRecord(runContext WorkflowRunContext, runID, draftID string, startedAt time.Time) WorkflowRunRecord {
	return WorkflowRunRecord{SchemaVersion: workflowRunRecordSchemaVersion, RunID: runID, DraftID: draftID, DraftVersion: 1, WorkspaceID: runContext.WorkspaceID, ApplicationID: runContext.ApplicationID, Status: WorkflowRunStatusRunning, StartedAt: workflowRunTimestamp(startedAt), ActorRef: runContext.ActorRef, RequestID: "request_test", AuditRef: "audit_test", Diagnostic: newWorkflowRunDiagnostic()}
}

func TestWorkflowRunListFilterValidation(t *testing.T) {
	future := time.Now().UTC().Add(time.Hour)
	service := newWorkflowExecutorService(nil, nil, newMemoryWorkflowRunStore(10))
	result := service.ListRuns(WorkflowRunContext{RequestContext: context.Background()}, WorkflowRunListRequest{Limit: 101, StartedTo: &future})
	if result.FailureCode != WorkflowRunFailureFilterInvalid {
		t.Fatalf("invalid filter accepted: %#v", result)
	}
}

func TestWorkflowRunDiagnosticFiltersAndCursorBinding(t *testing.T) {
	store := newMemoryWorkflowRunStore(10)
	runContext := workflowExecutorTestContext()
	stale := workflowRunHistoryTestRecord(runContext, "run_stale", "draft_a", time.Now().UTC().Add(-time.Minute))
	stale.SelectedProvider = "mock"
	stale.SelectedModel = "model-a"
	stale.FailureCode = WorkflowRunFailureGatewayFailed
	setWorkflowRunFailureDiagnostic(&stale, stale.FailureCode, "node_model", WorkflowRunGatewayFailureTimeout)
	if err := store.UpsertRun(runContext, &stale); err != nil {
		t.Fatal(err)
	}
	fresh := workflowRunHistoryTestRecord(runContext, "run_fresh", "draft_b", time.Now().UTC())
	fresh.SelectedProvider = "mock"
	fresh.SelectedModel = "model-b"
	if err := store.UpsertRun(runContext, &fresh); err != nil {
		t.Fatal(err)
	}
	service := newWorkflowExecutorService(nil, nil, store)
	staleOnly := true
	result := service.ListRuns(runContext, WorkflowRunListRequest{
		FailureCode: WorkflowRunFailureGatewayFailed, FailureBoundary: WorkflowRunFailureBoundaryGateway,
		Provider: "mock", Model: "model-a", StaleRunning: &staleOnly,
	})
	if result.FailureCode != "" || len(result.Runs) != 1 || result.Runs[0].RunID != stale.RunID || result.Runs[0].GatewayFailureCategory != WorkflowRunGatewayFailureTimeout {
		t.Fatalf("unexpected diagnostic filter result: %#v", result)
	}
	if invalid := service.ListRuns(runContext, WorkflowRunListRequest{Provider: "https://provider.invalid"}); invalid.FailureCode != WorkflowRunFailureFilterInvalid {
		t.Fatalf("endpoint-like provider filter accepted: %#v", invalid)
	}
	if _, code := parseWorkflowRunListRequest(map[string][]string{"stale_running": {"maybe"}}); code != WorkflowRunFailureFilterInvalid {
		t.Fatalf("invalid stale_running filter accepted: %s", code)
	}
}
