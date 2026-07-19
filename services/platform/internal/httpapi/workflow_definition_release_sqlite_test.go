package httpapi

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	sqliteworkflowrunmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_runs"
)

func TestSQLiteWorkflowDefinitionReleaseLifecycleRestartAndAppendOnly(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "workflow-definition-release.db")
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{DatabasePath: databasePath, Migrations: sqliteworkflowrunmigrations.Migrations()})
	if err != nil {
		t.Fatal(err)
	}
	repository := newSQLiteWorkflowDefinitionReleaseRepository(runtime.DB())
	ctx := workflowDefinitionTestContext()
	ctx.RequestContext = context.Background()
	now := time.Date(2026, 7, 19, 15, 0, 0, 0, time.UTC)
	candidate, err := repository.CreateCandidate(ctx, "candidate-sqlite", "definition-sqlite", workflowDefinitionTestDraft(), now)
	if err != nil {
		t.Fatalf("create candidate: %v", err)
	}
	approved, version, err := repository.Review(ctx, candidate.CandidateID, 0, "approve", "reviewed durable candidate", candidate.SourceDraftDigest, now.Add(time.Minute))
	if err != nil || version == nil || approved.State != workflowDefinitionStateApproved {
		t.Fatalf("review: %#v %#v %v", approved, version, err)
	}
	activation, err := repository.DecideActivation(ctx, candidate.DefinitionID, 0, "activate", 1, "activate durable version", now.Add(2*time.Minute))
	if err != nil || activation.PointerVersion != 1 {
		t.Fatalf("activate: %#v %v", activation, err)
	}
	summaries := repository.ListSummaries(ReadRepositoryContext{RequestContext: context.Background(), TenantRef: ctx.TenantRef, SubjectRef: ctx.OwnerSubjectRef, AuditRef: "audit_summary"}, ListWorkflowDefinitionSummariesRequest{ReadRepositoryRequest: ReadRepositoryRequest{Filters: ReadRepositoryFilters{"application_ref": ctx.ApplicationID}, Sort: "updated_at_desc"}})
	if summaries.FailureCode != "" || len(summaries.Items) != 1 || summaries.Items[0].WorkflowDefinitionID != candidate.DefinitionID || summaries.Items[0].DefinitionStatus != workflowDefinitionActivationActive || summaries.Items[0].Version != 1 {
		t.Fatalf("live definition summary: %#v", summaries)
	}
	if err = runtime.Close(); err != nil {
		t.Fatal(err)
	}

	restarted, err := sqlitedev.Open(context.Background(), sqlitedev.Options{DatabasePath: databasePath, Migrations: sqliteworkflowrunmigrations.Migrations()})
	if err != nil {
		t.Fatal(err)
	}
	defer restarted.Close()
	restartedRepository := newSQLiteWorkflowDefinitionReleaseRepository(restarted.DB())
	storedCandidate, err := restartedRepository.ReadCandidate(ctx, candidate.CandidateID)
	if err != nil || storedCandidate.DefinitionDigest != candidate.DefinitionDigest {
		t.Fatalf("restart candidate: %#v %v", storedCandidate, err)
	}
	storedVersions, err := restartedRepository.ListVersions(ctx, candidate.DefinitionID)
	if err != nil || len(storedVersions) != 1 || storedVersions[0].DefinitionDigest != candidate.DefinitionDigest {
		t.Fatalf("restart versions: %#v %v", storedVersions, err)
	}
	storedActivation, err := restartedRepository.ReadActivation(ctx, candidate.DefinitionID)
	if err != nil || storedActivation.ActiveVersion != 1 || len(storedActivation.Events) != 1 {
		t.Fatalf("restart activation: %#v %v", storedActivation, err)
	}
	if _, err = restarted.DB().ExecContext(context.Background(), `UPDATE workflow_definition_versions SET definition_version=2 WHERE definition_id='definition-sqlite'`); err == nil {
		t.Fatal("immutable definition version accepted UPDATE")
	}
}

func TestSQLiteWorkflowDefinitionReleaseCASAndCorruptionFailClosed(t *testing.T) {
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{DatabasePath: filepath.Join(t.TempDir(), "workflow-definition-cas.db"), Migrations: sqliteworkflowrunmigrations.Migrations()})
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Close()
	repository := newSQLiteWorkflowDefinitionReleaseRepository(runtime.DB())
	ctx := workflowDefinitionTestContext()
	ctx.RequestContext = context.Background()
	now := time.Now().UTC()
	candidate, err := repository.CreateCandidate(ctx, "candidate-sqlite-cas", "definition-sqlite-cas", workflowDefinitionTestDraft(), now)
	if err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	var lock sync.Mutex
	successes, conflicts := 0, 0
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, _, reviewErr := repository.Review(ctx, candidate.CandidateID, 0, "approve", "concurrent durable review", candidate.SourceDraftDigest, now.Add(time.Minute))
			lock.Lock()
			defer lock.Unlock()
			if reviewErr == nil {
				successes++
			} else if errors.Is(reviewErr, errWorkflowDefinitionConflict) || errors.Is(reviewErr, errWorkflowDefinitionInvalidState) {
				conflicts++
			}
		}()
	}
	wait.Wait()
	if successes != 1 || conflicts != 1 {
		t.Fatalf("review successes=%d conflicts=%d", successes, conflicts)
	}
	if _, err = runtime.DB().ExecContext(context.Background(), `UPDATE workflow_definition_release_candidates SET sanitized_candidate_payload=json_set(sanitized_candidate_payload,'$.definition_digest','sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa') WHERE candidate_id=?`, candidate.CandidateID); err != nil {
		t.Fatal(err)
	}
	if _, err = repository.ReadCandidate(ctx, candidate.CandidateID); !errors.Is(err, errWorkflowDefinitionStore) {
		t.Fatalf("corrupt projection must fail closed: %v", err)
	}
}

func TestWorkflowDefinitionReleaseRepositoryFactoryUsesSharedBackendAndRejectsUnknown(t *testing.T) {
	memory, err := newWorkflowDefinitionReleaseRepositoryForRunStore(newMemoryWorkflowRunStore(defaultWorkflowRunStoreCapacity))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := memory.(*workflowDefinitionReleaseStore); !ok {
		t.Fatalf("unexpected memory repository %T", memory)
	}
	if _, err = newWorkflowDefinitionReleaseRepositoryForRunStore(nil); err == nil {
		t.Fatal("unknown workflow backend must fail closed")
	}
}

func TestWorkflowDefinitionReleaseLiveProjectionReplacesOfflineSample(t *testing.T) {
	server := NewServer(config.Config{ControlPlaneReadDevAuthEnabled: true, WorkflowSavedDraftDevHTTPEnabled: true, WorkflowSavedDraftDevWriteEnabled: true, WorkflowDefinitionReleaseDevEnabled: true}, Options{BuildVersion: "test"})
	defer server.Close()
	ctx := workflowDefinitionTestContext()
	candidate, err := server.workflowDefinitionReleaseRepository.CreateCandidate(ctx, "candidate-live-summary", "definition-live-summary", workflowDefinitionTestDraft(), time.Date(2026, 7, 19, 16, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = server.workflowDefinitionReleaseRepository.Review(ctx, candidate.CandidateID, 0, "approve", "reviewed live projection", candidate.SourceDraftDigest, time.Date(2026, 7, 19, 16, 1, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	result := server.controlPlaneReadRepository().ListWorkflowDefinitionSummaries(ReadRepositoryContext{TenantRef: ctx.TenantRef, SubjectRef: ctx.OwnerSubjectRef, AuditRef: "audit_live"}, ListWorkflowDefinitionSummariesRequest{})
	if result.FailureCode != "" || len(result.Items) != 1 || result.Items[0].WorkflowDefinitionID != candidate.DefinitionID || result.Items[0].DefinitionStatus != workflowDefinitionActivationInactive {
		t.Fatalf("live projection mixed or omitted repository state: %#v", result)
	}
}
