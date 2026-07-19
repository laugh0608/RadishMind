package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	applicationcatalogmigrations "radishmind.local/services/platform/migrations/sqlite/application_catalog_records"
	sqliteworkflowrunmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_runs"
	saveddraftmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_saved_drafts"
)

func TestSQLiteWorkflowDefinitionContinuousProductChainRestartAndDeactivation(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "workflow-definition-product-chain.db")
	migrations := append(applicationcatalogmigrations.Migrations(), saveddraftmigrations.Migrations()...)
	migrations = append(migrations, sqliteworkflowrunmigrations.Migrations()...)
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{DatabasePath: databasePath, Migrations: migrations})
	if err != nil {
		t.Fatal(err)
	}

	applicationID := "app_aaaaaaaaaaaaaaaa"
	owner := "subject_definition_sqlite"
	applicationRepository := newSQLiteApplicationCatalogRepository(runtime.DB())
	applicationService := newApplicationCatalogService(applicationRepository)
	applicationService.newID = func() (string, error) { return applicationID, nil }
	applicationContext := ApplicationCatalogContext{RequestContext: context.Background(), RequestID: "request_definition_sqlite_app", TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", ActorRef: owner, OwnerSubjectRef: owner, AuditRef: "audit_definition_sqlite_app", WriteEnabled: true}
	if created := applicationService.Create(applicationContext, ApplicationCatalogCreateInput{DisplayName: "SQLite definition product", ApplicationKind: "workflow_copilot"}); created.FailureCode != "" {
		t.Fatalf("create application: %#v", created)
	}

	releaseContext := WorkflowDefinitionReleaseContext{RequestContext: context.Background(), TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", ApplicationID: applicationID, OwnerSubjectRef: owner, ActorRef: owner, RequestID: "request_definition_sqlite_release", AuditRef: "audit_definition_sqlite_release"}
	draft := executableWorkflowDraftForTest()
	draft.ApplicationID, draft.ToolRefs, draft.RAGRefs, draft.RequestedCapabilities = applicationID, []string{}, []string{}, []string{}
	repository := newSQLiteWorkflowDefinitionReleaseRepository(runtime.DB())
	savedDraftStore := newSQLiteSavedWorkflowDraftStore(runtime.DB())
	savedDraftService := newSavedWorkflowDraftService(savedDraftStore)
	savedDraftContext := SavedWorkflowDraftContext{RequestContext: context.Background(), RequestID: "request_definition_sqlite_draft", TenantRef: releaseContext.TenantRef, WorkspaceID: releaseContext.WorkspaceID, ApplicationID: applicationID, ActorRef: owner, OwnerSubjectRef: owner, ScopeGrants: []string{"workflow_drafts:read", "workflow_drafts:write"}, AuditRef: "audit_definition_sqlite_draft", WriteEnabled: true}
	saved := savedDraftService.SaveDraft(savedDraftContext, SaveWorkflowDraftRequest{Payload: savedWorkflowDraftPayloadFromDraft(draft)})
	if saved.FailureCode != "" || saved.Draft == nil || saved.Draft.DraftVersion != 1 {
		t.Fatalf("save exact draft: %#v", saved)
	}
	releaseService := newWorkflowDefinitionReleaseService(savedDraftStore, repository)
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	releaseService.now = func() time.Time { return now }
	created := releaseService.Create(releaseContext, WorkflowDefinitionCandidateCreateInput{CandidateID: "candidate_sqlite_product", DefinitionID: "definition_sqlite_product", DraftID: saved.Draft.DraftID, ExpectedDraftVersion: saved.Draft.DraftVersion})
	if created.FailureCode != "" || created.Candidate == nil {
		t.Fatalf("create candidate from exact saved draft: %#v", created)
	}
	createdJSON, err := json.Marshal(created.Candidate)
	if err != nil || !strings.Contains(string(createdJSON), `"reviews":[]`) {
		t.Fatalf("pending candidate must expose a strict empty review array: %s err=%v", createdJSON, err)
	}
	candidate := *created.Candidate
	releaseService.now = func() time.Time { return now.Add(time.Minute) }
	reviewed := releaseService.Review(releaseContext, candidate.CandidateID, WorkflowDefinitionReviewInput{ExpectedReviewVersion: 0, Decision: "approve", Reason: "approve continuous SQLite product chain"})
	if reviewed.FailureCode != "" || reviewed.Version == nil {
		t.Fatalf("review candidate: %#v", reviewed)
	}
	version := reviewed.Version
	releaseService.now = func() time.Time { return now.Add(2 * time.Minute) }
	activated := releaseService.DecideActivation(releaseContext, version.DefinitionID, WorkflowDefinitionActivationInput{ExpectedPointerVersion: 0, Decision: "activate", Version: version.Version, Reason: "activate exact SQLite version"})
	if activated.FailureCode != "" || activated.Activation == nil {
		t.Fatalf("activate definition: %#v", activated)
	}
	activation := *activated.Activation

	runStore := newSQLiteWorkflowRunStore(runtime.DB())
	bridgeClient := &workflowExecutorTestBridge{}
	executor := newWorkflowExecutorService(nil, bridgeClient, runStore)
	execution := newWorkflowDefinitionExecutionService(repository, applicationRepository, executor)
	runContext := WorkflowRunContext{RequestContext: context.Background(), RequestID: "request_definition_sqlite_run", TenantRef: releaseContext.TenantRef, WorkspaceID: releaseContext.WorkspaceID, ApplicationID: applicationID, ActorRef: owner, AuditRef: "audit_definition_sqlite_run"}
	runRequest := WorkflowDefinitionRunRequest{DefinitionID: version.DefinitionID, ExpectedPointerVersion: activation.PointerVersion, ExpectedDefinitionVersion: version.Version, ExpectedDefinitionDigest: version.DefinitionDigest, InputText: "private SQLite continuous-chain input", ConditionValues: map[string]bool{}}
	baseline := execution.StartRun(runContext, runRequest)
	candidateRun := execution.StartRun(runContext, runRequest)
	if baseline.Record == nil || candidateRun.Record == nil || baseline.FailureCode != "" || candidateRun.FailureCode != "" {
		t.Fatalf("execute definition runs: baseline=%#v candidate=%#v", baseline, candidateRun)
	}
	comparison := executor.CompareRuns(runContext, baseline.Record.RunID, candidateRun.Record.RunID)
	if comparison.FailureCode != "" || comparison.Comparison == nil || comparison.Comparison.RunProfile != workflowDefinitionEvaluationProfile || bridgeClient.callCount() != 2 {
		t.Fatalf("read-only v5 comparison: %#v bridge=%d", comparison, bridgeClient.callCount())
	}
	releaseService.now = func() time.Time { return now.Add(3 * time.Minute) }
	deactivationResult := releaseService.DecideActivation(releaseContext, version.DefinitionID, WorkflowDefinitionActivationInput{ExpectedPointerVersion: activation.PointerVersion, Decision: "deactivate", Reason: "stop exact SQLite authority"})
	if deactivationResult.FailureCode != "" || deactivationResult.Activation == nil {
		t.Fatalf("deactivate definition: %#v", deactivationResult)
	}
	deactivated := *deactivationResult.Activation
	blocked := execution.StartRun(runContext, runRequest)
	if blocked.FailureCode != WorkflowRunFailureDefinitionAuthority || bridgeClient.callCount() != 2 {
		t.Fatalf("deactivated authority reached provider: %#v bridge=%d", blocked, bridgeClient.callCount())
	}
	if err = runtime.Close(); err != nil {
		t.Fatal(err)
	}

	restarted, err := sqlitedev.Open(context.Background(), sqlitedev.Options{DatabasePath: databasePath, Migrations: migrations})
	if err != nil {
		t.Fatal(err)
	}
	defer restarted.Close()
	restartedRepository := newSQLiteWorkflowDefinitionReleaseRepository(restarted.DB())
	restoredActivation, err := restartedRepository.ReadActivation(releaseContext, version.DefinitionID)
	if err != nil || restoredActivation.State != workflowDefinitionActivationInactive || restoredActivation.PointerVersion != deactivated.PointerVersion || len(restoredActivation.Events) != 2 {
		t.Fatalf("restart activation evidence: %#v %v", restoredActivation, err)
	}
	restoredRun, found, err := newSQLiteWorkflowRunStore(restarted.DB()).ReadRun(runContext, baseline.Record.RunID)
	if err != nil || !found || restoredRun.SchemaVersion != workflowRunRecordDefinitionSchemaVersion || restoredRun.DefinitionAuthority == nil || restoredRun.Output != "" {
		t.Fatalf("restart v5 run evidence: %#v found=%t err=%v", restoredRun, found, err)
	}
	var storedPayload string
	if err = restarted.DB().QueryRowContext(context.Background(), `SELECT sanitized_run_record FROM workflow_run_records WHERE run_id=?`, baseline.Record.RunID).Scan(&storedPayload); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(storedPayload, runRequest.InputText) || strings.Contains(storedPayload, baseline.AdvisoryOutput) {
		t.Fatal("SQLite v5 payload persisted raw input or advisory output")
	}
}

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
