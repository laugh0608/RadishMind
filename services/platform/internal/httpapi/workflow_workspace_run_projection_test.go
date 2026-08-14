package httpapi

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestWorkspaceRunProjectionMemoryScopeCursorAndApplicationFilter(t *testing.T) {
	runWorkspaceRunProjectionScopeCursorAndApplicationFilter(t, newMemoryWorkflowRunStore(20))
}

func TestWorkspaceRunProjectionSQLiteScopeCursorAndApplicationFilter(t *testing.T) {
	runtime := openWorkflowRunSQLiteRuntime(t, filepath.Join(t.TempDir(), "workspace-runs.db"))
	runWorkspaceRunProjectionScopeCursorAndApplicationFilter(t, newSQLiteWorkflowRunStore(runtime.DB()))
}

func TestCombinedWorkspaceRunProjectionMergesAllRunOwners(t *testing.T) {
	workflowRuns := newMemoryWorkflowRunStore(10)
	promptRuns := newMemoryWorkflowRunStore(10)
	agentRuns := newMemoryWorkflowRunStore(10)
	base := workflowExecutorTestContext()
	base.ActorRef = "subject_owner"
	startedAt := time.Date(2026, 7, 26, 13, 0, 0, 0, time.UTC)
	for index, fixture := range []struct {
		store         workflowRunStore
		applicationID string
		runID         string
	}{
		{store: workflowRuns, applicationID: "app_aaaaaaaaaaaaaaaa", runID: "run_workflow_owner"},
		{store: promptRuns, applicationID: "app_bbbbbbbbbbbbbbbb", runID: "run_prompt_owner"},
		{store: agentRuns, applicationID: "app_cccccccccccccccc", runID: "run_agent_owner"},
	} {
		runContext := base
		runContext.ApplicationID = fixture.applicationID
		seedWorkspaceProjectionRun(t, fixture.store, runContext, fixture.runID, startedAt.Add(time.Duration(index)*time.Minute))
	}
	combined := newCombinedWorkflowRunStoreWithAgent(workflowRuns, promptRuns, agentRuns)
	repository := newWorkspaceScopedControlPlaneReadRepository(
		newControlPlaneReadRepository(newControlPlaneReadFakeStore()),
		newMemoryApplicationCatalogRepository(), newMemoryAPIKeyRepository(), nil, combined,
	)
	readContext := ReadRepositoryContext{
		RequestContext: context.Background(), TenantRef: base.TenantRef,
		WorkspaceID: base.WorkspaceID, SubjectRef: base.ActorRef, AuditRef: "audit_combined_runs",
	}
	first := repository.ListRunRecordSummaries(readContext, ListRunRecordSummariesRequest{
		ReadRepositoryRequest: ReadRepositoryRequest{Limit: 2},
	})
	if first.FailureCode != "" || len(first.Items) != 2 || first.NextCursor == nil ||
		first.Items[0].RunID != "run_agent_owner" || first.Items[1].RunID != "run_prompt_owner" {
		t.Fatalf("combined workspace run first page mismatch: %#v", first)
	}
	second := repository.ListRunRecordSummaries(readContext, ListRunRecordSummariesRequest{
		ReadRepositoryRequest: ReadRepositoryRequest{Limit: 2, Cursor: *first.NextCursor},
	})
	if second.FailureCode != "" || len(second.Items) != 1 || second.NextCursor != nil ||
		second.Items[0].RunID != "run_workflow_owner" {
		t.Fatalf("combined workspace run second page mismatch: %#v", second)
	}
}

func runWorkspaceRunProjectionScopeCursorAndApplicationFilter(t *testing.T, store workflowRunStore) {
	t.Helper()
	startedAt := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	base := workflowExecutorTestContext()
	base.ActorRef = "subject_owner"
	base.ApplicationID = "app_aaaaaaaaaaaaaaaa"
	seedWorkspaceProjectionRun(t, store, base, "run_shared_aaaaaaaa", startedAt)

	secondApplication := base
	secondApplication.ApplicationID = "app_bbbbbbbbbbbbbbbb"
	seedWorkspaceProjectionRun(t, store, secondApplication, "run_shared_aaaaaaaa", startedAt)

	otherSubject := base
	otherSubject.ActorRef = "subject_other"
	otherSubject.ApplicationID = "app_cccccccccccccccc"
	seedWorkspaceProjectionRun(t, store, otherSubject, "run_subject_other", startedAt.Add(time.Hour))

	otherWorkspace := base
	otherWorkspace.WorkspaceID = "workspace_other"
	otherWorkspace.ApplicationID = "app_dddddddddddddddd"
	seedWorkspaceProjectionRun(t, store, otherWorkspace, "run_workspace_other", startedAt.Add(time.Hour))

	repository := newWorkspaceScopedControlPlaneReadRepository(
		newControlPlaneReadRepository(newControlPlaneReadFakeStore()),
		newMemoryApplicationCatalogRepository(), newMemoryAPIKeyRepository(), nil, store,
	)
	readContext := ReadRepositoryContext{
		RequestContext: context.Background(), RequestID: "request_workspace_runs",
		TenantRef: base.TenantRef, WorkspaceID: base.WorkspaceID, SubjectRef: base.ActorRef,
		AuditRef: "audit_workspace_runs",
	}
	first := repository.ListRunRecordSummaries(readContext, ListRunRecordSummariesRequest{
		ReadRepositoryRequest: ReadRepositoryRequest{Limit: 1, Sort: "started_at_desc"},
	})
	if first.FailureCode != "" || len(first.Items) != 1 || first.NextCursor == nil ||
		first.Items[0].ApplicationRef != secondApplication.ApplicationID {
		t.Fatalf("unexpected workspace run first page: %#v", first)
	}
	second := repository.ListRunRecordSummaries(readContext, ListRunRecordSummariesRequest{
		ReadRepositoryRequest: ReadRepositoryRequest{
			Limit: 1, Sort: "started_at_desc", Cursor: *first.NextCursor,
		},
	})
	if second.FailureCode != "" || len(second.Items) != 1 || second.NextCursor != nil ||
		second.Items[0].ApplicationRef != base.ApplicationID {
		t.Fatalf("unexpected workspace run second page: %#v", second)
	}

	filtered := repository.ListRunRecordSummaries(readContext, ListRunRecordSummariesRequest{
		ReadRepositoryRequest: ReadRepositoryRequest{
			Filters: ReadRepositoryFilters{"application_ref": base.ApplicationID},
		},
	})
	if filtered.FailureCode != "" || len(filtered.Items) != 1 ||
		filtered.Items[0].ApplicationRef != base.ApplicationID {
		t.Fatalf("application-filtered workspace run projection mismatch: %#v", filtered)
	}

	changedWorkspace := readContext
	changedWorkspace.WorkspaceID = otherWorkspace.WorkspaceID
	reused := repository.ListRunRecordSummaries(changedWorkspace, ListRunRecordSummariesRequest{
		ReadRepositoryRequest: ReadRepositoryRequest{Limit: 1, Sort: "started_at_desc", Cursor: *first.NextCursor},
	})
	if reused.FailureCode != ReadRepositoryFailureInvalidFilter || len(reused.Items) != 0 {
		t.Fatalf("cross-workspace cursor was accepted: %#v", reused)
	}

	changedSubject := readContext
	changedSubject.SubjectRef = otherSubject.ActorRef
	reused = repository.ListRunRecordSummaries(changedSubject, ListRunRecordSummariesRequest{
		ReadRepositoryRequest: ReadRepositoryRequest{Limit: 1, Sort: "started_at_desc", Cursor: *first.NextCursor},
	})
	if reused.FailureCode != ReadRepositoryFailureInvalidFilter || len(reused.Items) != 0 {
		t.Fatalf("cross-subject cursor was accepted: %#v", reused)
	}
}

func seedWorkspaceProjectionRun(
	t *testing.T,
	store workflowRunStore,
	runContext WorkflowRunContext,
	runID string,
	startedAt time.Time,
) {
	t.Helper()
	record := workflowRunHistoryTestRecord(runContext, runID, "draft_workspace_projection", startedAt)
	if err := store.UpsertRun(runContext, &record); err != nil {
		t.Fatalf("seed workspace run %s/%s: %v", runContext.ApplicationID, runID, err)
	}
}
