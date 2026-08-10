package httpapi

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/sqlitedev"
	applicationcatalogmigrations "radishmind.local/services/platform/migrations/sqlite/application_catalog_records"
	sqliteworkflowrunmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_runs"
	saveddraftmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_saved_drafts"
)

func TestSQLiteWorkflowDefinitionStructuredProductChainRestartPrivacyAndCorruption(t *testing.T) {
	ctx := context.Background()
	databasePath := filepath.Join(t.TempDir(), "workflow-definition-structured-product-chain.db")
	migrations := append(applicationcatalogmigrations.Migrations(), saveddraftmigrations.Migrations()...)
	migrations = append(migrations, sqliteworkflowrunmigrations.Migrations()...)
	runtime, err := sqlitedev.Open(ctx, sqlitedev.Options{DatabasePath: databasePath, Migrations: migrations})
	if err != nil {
		t.Fatal(err)
	}

	applicationID := "app_dddddddddddddddd"
	owner := "subject_structured_sqlite"
	applicationRepository := newSQLiteApplicationCatalogRepository(runtime.DB())
	applicationService := newApplicationCatalogService(applicationRepository)
	applicationService.newID = func() (string, error) { return applicationID, nil }
	applicationContext := ApplicationCatalogContext{
		RequestContext: ctx, RequestID: "request_structured_sqlite_app", TenantRef: "tenant_demo",
		WorkspaceID: "workspace_demo", ActorRef: owner, OwnerSubjectRef: owner,
		AuditRef: "audit_structured_sqlite_app", WriteEnabled: true,
	}
	if result := applicationService.Create(applicationContext, ApplicationCatalogCreateInput{DisplayName: "Structured SQLite definition", ApplicationKind: "workflow_copilot"}); result.FailureCode != "" {
		t.Fatalf("create application: %#v", result)
	}

	releaseContext := WorkflowDefinitionReleaseContext{
		RequestContext: ctx, TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", ApplicationID: applicationID,
		OwnerSubjectRef: owner, ActorRef: owner, RequestID: "request_structured_sqlite_release", AuditRef: "audit_structured_sqlite_release",
	}
	savedDraftContext := SavedWorkflowDraftContext{
		RequestContext: ctx, RequestID: "request_structured_sqlite_draft", TenantRef: releaseContext.TenantRef,
		WorkspaceID: releaseContext.WorkspaceID, ApplicationID: applicationID, ActorRef: owner, OwnerSubjectRef: owner,
		ScopeGrants: []string{"workflow_drafts:read", "workflow_drafts:write"}, AuditRef: "audit_structured_sqlite_draft", WriteEnabled: true,
	}
	draft := executableWorkflowStructuredDraftForTest(applicationID)
	draftPayload := savedWorkflowDraftPayloadFromDraft(draft)
	if draftPayload.SchemaVersion != savedWorkflowDraftStructuredSchemaVersion {
		t.Fatalf("structured draft fixture lost schema identity: %#v", draftPayload)
	}
	savedDraftStore := newSQLiteSavedWorkflowDraftStore(runtime.DB())
	saved := newSavedWorkflowDraftService(savedDraftStore).SaveDraft(savedDraftContext, SaveWorkflowDraftRequest{Payload: draftPayload})
	if saved.FailureCode != "" || saved.Draft == nil || saved.Draft.SchemaVersion != savedWorkflowDraftStructuredSchemaVersion ||
		saved.Draft.InputContract.ContractDigest == "" || len(saved.Draft.InputContract.Fields) != 4 {
		t.Fatalf("save exact structured draft: %#v", saved)
	}

	repository := newSQLiteWorkflowDefinitionReleaseRepository(runtime.DB())
	releaseService := newWorkflowDefinitionReleaseService(savedDraftStore, repository)
	now := time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC)
	releaseService.now = func() time.Time { return now }
	created := releaseService.Create(releaseContext, WorkflowDefinitionCandidateCreateInput{
		CandidateID: "candidate_structured_sqlite", DefinitionID: "definition_structured_sqlite",
		DraftID: saved.Draft.DraftID, ExpectedDraftVersion: saved.Draft.DraftVersion, ExpectedLifecycleVersion: saved.Draft.LifecycleVersion,
	})
	if created.FailureCode != "" || created.Candidate == nil || created.Candidate.SchemaVersion != workflowDefinitionCandidateStructuredSchemaVersion {
		t.Fatalf("create structured candidate: %#v", created)
	}
	releaseService.now = func() time.Time { return now.Add(time.Minute) }
	reviewed := releaseService.Review(releaseContext, created.Candidate.CandidateID, WorkflowDefinitionReviewInput{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: "approve structured SQLite product chain",
	})
	if reviewed.FailureCode != "" || reviewed.Version == nil || reviewed.Version.SchemaVersion != workflowDefinitionVersionStructuredSchemaVersion {
		t.Fatalf("approve structured definition: %#v", reviewed)
	}
	version := reviewed.Version
	releaseService.now = func() time.Time { return now.Add(2 * time.Minute) }
	activated := releaseService.DecideActivation(releaseContext, version.DefinitionID, WorkflowDefinitionActivationInput{
		ExpectedPointerVersion: 0, Decision: "activate", Version: version.Version, Reason: "activate structured SQLite definition",
	})
	if activated.FailureCode != "" || activated.Activation == nil {
		t.Fatalf("activate structured definition: %#v", activated)
	}

	runStore := newSQLiteWorkflowRunStore(runtime.DB())
	bridgeClient := &workflowExecutorTestBridge{}
	executor := newWorkflowExecutorService(nil, bridgeClient, runStore)
	execution := newWorkflowDefinitionExecutionService(repository, applicationRepository, executor)
	runContext := WorkflowRunContext{
		RequestContext: ctx, RequestID: "request_structured_sqlite_run", TenantRef: releaseContext.TenantRef,
		WorkspaceID: releaseContext.WorkspaceID, ApplicationID: applicationID, ActorRef: owner, AuditRef: "audit_structured_sqlite_run",
	}
	privateValue := "private-structured-sqlite-customer"
	runRequest := WorkflowDefinitionRunRequest{
		DefinitionID: version.DefinitionID, ExpectedPointerVersion: activated.Activation.PointerVersion,
		ExpectedDefinitionVersion: version.Version, ExpectedDefinitionDigest: version.DefinitionDigest,
		Inputs: map[string]any{"customer_name": privateValue, "retry_count": 3}, ConditionValues: map[string]bool{},
	}
	baseline := execution.StartRun(runContext, runRequest)
	candidate := execution.StartRun(runContext, runRequest)
	if baseline.FailureCode != "" || candidate.FailureCode != "" || baseline.Record == nil || candidate.Record == nil || bridgeClient.callCount() != 2 {
		t.Fatalf("execute structured SQLite runs: baseline=%#v candidate=%#v bridge=%d", baseline, candidate, bridgeClient.callCount())
	}
	comparison := executor.CompareRuns(runContext, baseline.Record.RunID, candidate.Record.RunID)
	if comparison.FailureCode != "" || comparison.Comparison == nil ||
		comparison.Comparison.SchemaVersion != workflowDefinitionStructuredRunComparisonSchemaVersion ||
		comparison.Comparison.RunProfile != workflowDefinitionStructuredEvaluationProfile {
		t.Fatalf("compare structured SQLite runs: %#v", comparison)
	}

	var storedPayload, projectedContractID, projectedContractDigest string
	if err = runtime.DB().QueryRowContext(ctx, `SELECT sanitized_run_record,input_contract_id,input_contract_digest FROM workflow_run_records WHERE run_id=?`, baseline.Record.RunID).
		Scan(&storedPayload, &projectedContractID, &projectedContractDigest); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(storedPayload, privateValue) || strings.Contains(storedPayload, baseline.AdvisoryOutput) ||
		projectedContractID != baseline.Record.InputContractID || projectedContractDigest != baseline.Record.InputContractDigest {
		t.Fatalf("SQLite Run v8 privacy/projection drifted: id=%s digest=%s payload=%s", projectedContractID, projectedContractDigest, storedPayload)
	}
	if _, err = runtime.DB().ExecContext(ctx, `UPDATE workflow_run_records SET input_contract_digest=? WHERE run_id=?`, "sha256:"+strings.Repeat("f", 64), baseline.Record.RunID); err == nil {
		t.Fatal("SQLite accepted a structured input projection that disagrees with the sanitized record")
	}
	if err = runtime.Close(); err != nil {
		t.Fatal(err)
	}

	restarted, err := sqlitedev.Open(ctx, sqlitedev.Options{DatabasePath: databasePath, Migrations: migrations})
	if err != nil {
		t.Fatal(err)
	}
	restartedDraft := newSavedWorkflowDraftService(newSQLiteSavedWorkflowDraftStore(restarted.DB())).ReadDraft(savedDraftContext, ReadWorkflowDraftRequest{DraftID: saved.Draft.DraftID})
	if restartedDraft.FailureCode != "" || restartedDraft.Draft == nil || restartedDraft.Draft.InputContract.ContractDigest != saved.Draft.InputContract.ContractDigest {
		_ = restarted.Close()
		t.Fatalf("restart structured draft evidence: %#v", restartedDraft)
	}
	restored, found, err := newSQLiteWorkflowRunStore(restarted.DB()).ReadRun(runContext, baseline.Record.RunID)
	if err != nil || !found || restored.SchemaVersion != workflowRunRecordDefinitionStructuredSchemaVersion ||
		restored.InputContractDigest != baseline.Record.InputContractDigest || restored.Output != "" {
		_ = restarted.Close()
		t.Fatalf("restart Run v8 evidence: %#v found=%t err=%v", restored, found, err)
	}

	restartedBridge := &workflowExecutorTestBridge{}
	restartedExecution := newWorkflowDefinitionExecutionService(
		newSQLiteWorkflowDefinitionReleaseRepository(restarted.DB()), newSQLiteApplicationCatalogRepository(restarted.DB()),
		newWorkflowExecutorService(nil, restartedBridge, newSQLiteWorkflowRunStore(restarted.DB())),
	)
	wrongShape := runRequest
	wrongShape.Inputs = nil
	wrongShape.InputText = "legacy input"
	if rejected := restartedExecution.StartRun(runContext, wrongShape); rejected.FailureCode != WorkflowRunFailureInputSchemaUnsupported || rejected.Record != nil || restartedBridge.callCount() != 0 {
		_ = restarted.Close()
		t.Fatalf("restarted structured definition fell back to v1 input: %#v bridge=%d", rejected, restartedBridge.callCount())
	}

	connection, err := restarted.DB().Conn(ctx)
	if err != nil {
		_ = restarted.Close()
		t.Fatal(err)
	}
	if _, err = connection.ExecContext(ctx, `PRAGMA ignore_check_constraints = ON`); err == nil {
		_, err = connection.ExecContext(ctx, `UPDATE workflow_run_records SET input_contract_digest=? WHERE run_id=?`, "sha256:"+strings.Repeat("e", 64), baseline.Record.RunID)
	}
	_, restoreConstraintErr := connection.ExecContext(ctx, `PRAGMA ignore_check_constraints = OFF`)
	closeErr := connection.Close()
	if err != nil || restoreConstraintErr != nil || closeErr != nil {
		_ = restarted.Close()
		t.Fatalf("prepare SQLite corruption fixture: update=%v restore=%v close=%v", err, restoreConstraintErr, closeErr)
	}
	if _, found, err = newSQLiteWorkflowRunStore(restarted.DB()).ReadRun(runContext, baseline.Record.RunID); err == nil || found {
		_ = restarted.Close()
		t.Fatalf("SQLite Run v8 projection corruption did not fail closed: found=%t err=%v", found, err)
	}
	if err = restarted.Close(); err != nil {
		t.Fatal(err)
	}
}
