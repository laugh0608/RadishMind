//go:build postgres_integration

package httpapi

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	workflowrunmigrations "radishmind.local/services/platform/migrations/workflow_runs"
)

func TestPostgresWorkflowRunStoreIntegration(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeDatabaseURL := postgresIntegrationDatabaseURLForCredentials(t, runtimeUser, os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := workflowrunmigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = workflowrunmigrations.RollbackForDevTest(ctx, pool)
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, pool)
	preparePostgresIntegrationRuntimeRole(t, ctx, pool, runtimeUser)
	t.Cleanup(func() {
		cleanup, cancelCleanup := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelCleanup()
		_, _ = workflowrunmigrations.RollbackForDevTest(cleanup, pool)
		pool.Close()
	})
	state, err := workflowrunmigrations.Apply(ctx, pool)
	if err != nil || state.MigrationState != workflowrunmigrations.MigrationStateApplied {
		t.Fatalf("apply migration: %#v %v", state, err)
	}
	if repeated, err := workflowrunmigrations.Apply(ctx, pool); err != nil || repeated.MigrationChecksum != state.MigrationChecksum {
		t.Fatalf("repeat migration: %#v %v", repeated, err)
	}
	runtimePool, err := workflowrunmigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	if _, ddlErr := runtimePool.Exec(ctx, "CREATE TABLE workflow_run_runtime_must_not_create (id integer)"); ddlErr == nil {
		t.Fatal("workflow run runtime role must not have schema DDL permission")
	}
	store := newPostgresWorkflowRunStore(runtimePool)
	runContext := workflowExecutorTestContext()
	runContext.RequestContext = ctx
	base := time.Now().UTC().Add(-time.Minute)
	for index := 0; index < 3; index++ {
		record := workflowRunHistoryTestRecord(runContext, "run_pg_"+string(rune('a'+index)), "draft_pg", base.Add(time.Duration(index)*time.Second))
		if err = store.UpsertRun(runContext, &record); err != nil {
			t.Fatal(err)
		}
		record.Status = WorkflowRunStatusSucceeded
		record.CompletedAt = workflowRunTimestamp(base.Add(time.Duration(index)*time.Second + time.Millisecond))
		record.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
		if err = store.UpsertRun(runContext, &record); err != nil {
			t.Fatal(err)
		}
	}
	page, err := store.ListRuns(runContext, WorkflowRunListFilter{Limit: 2})
	if err != nil || len(page.Records) != 2 || !page.HasMore || page.Records[0].RunID != "run_pg_c" {
		t.Fatalf("unexpected PostgreSQL page: %#v %v", page, err)
	}
	legacy := workflowRunHistoryTestRecord(runContext, "run_pg_legacy", "draft_pg", base.Add(-time.Second))
	legacy.SchemaVersion = workflowRunRecordLegacySchemaVersion
	legacy.Diagnostic = nil
	if err = store.UpsertRun(runContext, &legacy); err != nil {
		t.Fatalf("write legacy run: %v", err)
	}
	if decoded, found, readErr := store.ReadRun(runContext, legacy.RunID); readErr != nil || !found || decoded.SchemaVersion != workflowRunRecordLegacySchemaVersion || decoded.Diagnostic != nil {
		t.Fatalf("legacy v0 read compatibility failed: found=%v record=%#v err=%v", found, decoded, readErr)
	}
	diagnostic := workflowRunHistoryTestRecord(runContext, "run_pg_diagnostic", "draft_pg", base.Add(4*time.Second))
	diagnostic.SelectedProvider = "mock"
	diagnostic.SelectedModel = "model-diagnostic"
	if err = store.UpsertRun(runContext, &diagnostic); err != nil {
		t.Fatal(err)
	}
	diagnostic.Status = WorkflowRunStatusFailed
	diagnostic.FailureCode = WorkflowRunFailureGatewayFailed
	diagnostic.FailureSummary = "Gateway timed out while executing the workflow model node."
	diagnostic.CompletedAt = workflowRunTimestamp(base.Add(5 * time.Second))
	setWorkflowRunFailureDiagnostic(&diagnostic, diagnostic.FailureCode, "node_model", WorkflowRunGatewayFailureTimeout)
	diagnostic.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
	if err = store.UpsertRun(runContext, &diagnostic); err != nil {
		t.Fatal(err)
	}
	diagnosticPage, err := store.ListRuns(runContext, WorkflowRunListFilter{Limit: 10, FailureCode: WorkflowRunFailureGatewayFailed, FailureBoundary: WorkflowRunFailureBoundaryGateway, Provider: "mock", Model: "model-diagnostic"})
	if err != nil || len(diagnosticPage.Records) != 1 || diagnosticPage.Records[0].RunID != diagnostic.RunID {
		t.Fatalf("diagnostic PostgreSQL filter failed: %#v %v", diagnosticPage, err)
	}
	evaluationStore := newPostgresWorkflowEvaluationStore(runtimePool)
	evaluationService := newWorkflowEvaluationService(evaluationStore, store)
	evaluationService.newCaseID = func() (string, error) { return "eval_pg_restart", nil }
	evaluation := evaluationService.Create(runContext, WorkflowEvaluationCreateRequest{Name: "PostgreSQL restart review", BaselineRunID: diagnostic.RunID, Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: "run_pg_c", ExpectedClassification: WorkflowRunComparisonImprovement}}})
	if evaluation.FailureCode != "" || evaluation.Case == nil { t.Fatalf("create PostgreSQL evaluation case: %#v", evaluation) }
	other := runContext
	other.TenantRef = "tenant_other"
	if scoped, err := store.ListRuns(other, WorkflowRunListFilter{Limit: 10}); err != nil || len(scoped.Records) != 0 {
		t.Fatalf("scope leaked: %#v %v", scoped, err)
	}
	runtimePool.Close()
	reopened, err := workflowrunmigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	store = newPostgresWorkflowRunStore(reopened)
	evaluationStore = newPostgresWorkflowEvaluationStore(reopened)
	recovered, found, err := store.ReadRun(runContext, "run_pg_c")
	if err != nil || !found || recovered.Status != WorkflowRunStatusSucceeded {
		t.Fatalf("restart recovery failed: %#v %v", recovered, err)
	}
	comparisonService := newWorkflowExecutorService(nil, nil, store)
	comparison := comparisonService.CompareRuns(runContext, diagnostic.RunID, "run_pg_c")
	if comparison.FailureCode != "" || comparison.Comparison == nil || comparison.Comparison.Classification != WorkflowRunComparisonImprovement {
		t.Fatalf("restart comparison failed: %#v", comparison)
	}
	if scopedComparison := comparisonService.CompareRuns(other, diagnostic.RunID, "run_pg_c"); scopedComparison.FailureCode != WorkflowRunFailureRecordNotFound {
		t.Fatalf("restart comparison leaked cross scope: %#v", scopedComparison)
	}
	restartedEvaluationService := newWorkflowEvaluationService(evaluationStore, store)
	restartedReview := restartedEvaluationService.Review(runContext, evaluation.Case.CaseID)
	if restartedReview.FailureCode != "" || restartedReview.Review == nil || restartedReview.Review.Outcome != "passed" { t.Fatalf("restart batch evaluation failed: %#v", restartedReview) }
	if leaked := restartedEvaluationService.Read(other, evaluation.Case.CaseID); leaked.FailureCode != WorkflowEvaluationFailureNotFound { t.Fatalf("evaluation case leaked scope: %#v", leaked) }
	running := workflowRunHistoryTestRecord(runContext, "run_pg_concurrent", "draft_pg", time.Now().UTC())
	if err = store.UpsertRun(runContext, &running); err != nil {
		t.Fatal(err)
	}
	left, right := cloneWorkflowRunRecord(running), cloneWorkflowRunRecord(running)
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, candidate := range []*WorkflowRunRecord{&left, &right} {
		wg.Add(1)
		go func(value *WorkflowRunRecord) {
			defer wg.Done()
			value.Status = WorkflowRunStatusFailed
			value.FailureCode = WorkflowRunFailureGatewayFailed
			value.FailureSummary = "Gateway failed."
			value.CompletedAt = workflowRunTimestamp(time.Now())
			setWorkflowRunFailureDiagnostic(value, value.FailureCode, "node_model", WorkflowRunGatewayFailureUnavailable)
			value.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
			results <- store.UpsertRun(runContext, value)
		}(candidate)
	}
	wg.Wait()
	close(results)
	successes := 0
	for result := range results {
		if result == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("expected one PostgreSQL CAS winner, got %d", successes)
	}

	reopened.Close()
	if _, err = pool.Exec(ctx, "UPDATE workflow_run_schema_versions SET migration_checksum='sha256:mismatch' WHERE component=$1", workflowrunmigrations.Component); err != nil {
		t.Fatal(err)
	}
	if mismatch, inspectErr := workflowrunmigrations.Inspect(ctx, pool); inspectErr != nil || mismatch.MigrationState != workflowrunmigrations.MigrationStateMismatch {
		t.Fatalf("marker mismatch not detected: %#v %v", mismatch, inspectErr)
	}
	if _, err = pool.Exec(ctx, "UPDATE workflow_run_schema_versions SET migration_checksum=$1 WHERE component=$2", workflowrunmigrations.ExpectedChecksum(), workflowrunmigrations.Component); err != nil {
		t.Fatal(err)
	}
	if _, err = workflowrunmigrations.RollbackForDevTest(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if _, err = workflowrunmigrations.Apply(ctx, pool); err != nil {
		t.Fatalf("reapply after rollback: %v", err)
	}
	if _, err = workflowrunmigrations.RollbackForDevTest(ctx, pool); err != nil {
		t.Fatal(err)
	}
	legacySQL, err := os.ReadFile("../../migrations/workflow_runs/0001_workflow_runs.up.sql")
	if err != nil {
		t.Fatalf("read legacy workflow run migration: %v", err)
	}
	if _, err = pool.Exec(ctx, string(legacySQL)); err != nil {
		t.Fatalf("apply legacy workflow run migration: %v", err)
	}
	legacyChecksum := fmt.Sprintf("sha256:%x", sha256.Sum256(legacySQL))
	if _, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS workflow_run_schema_versions (component text PRIMARY KEY, migration_id text NOT NULL, store_schema_version text NOT NULL, migration_checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workflow_run_schema_versions(component,migration_id,store_schema_version,migration_checksum) VALUES($1,'0001_workflow_runs','workflow_run_store_v1',$2)`, workflowrunmigrations.Component, legacyChecksum); err != nil {
		t.Fatal(err)
	}
	if pending, inspectErr := workflowrunmigrations.Inspect(ctx, pool); inspectErr != nil || pending.MigrationState != workflowrunmigrations.MigrationStatePending {
		t.Fatalf("legacy marker was not recognized as pending: %#v %v", pending, inspectErr)
	}
	if upgraded, upgradeErr := workflowrunmigrations.Apply(ctx, pool); upgradeErr != nil || upgraded.MigrationState != workflowrunmigrations.MigrationStateApplied {
		t.Fatalf("legacy migration upgrade failed: %#v %v", upgraded, upgradeErr)
	}
	if _, err = workflowrunmigrations.RollbackForDevTest(ctx, pool); err != nil { t.Fatal(err) }
	diagnosticsSQL, err := os.ReadFile("../../migrations/workflow_runs/0002_workflow_run_diagnostics.up.sql"); if err != nil { t.Fatal(err) }
	if _, err = pool.Exec(ctx, string(legacySQL)+"\n"+string(diagnosticsSQL)); err != nil { t.Fatalf("apply diagnostics migration: %v", err) }
	diagnosticsChecksum := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(string(legacySQL)+"\n"+string(diagnosticsSQL))))
	if _, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS workflow_run_schema_versions (component text PRIMARY KEY, migration_id text NOT NULL, store_schema_version text NOT NULL, migration_checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil { t.Fatal(err) }
	if _, err = pool.Exec(ctx, `INSERT INTO workflow_run_schema_versions(component,migration_id,store_schema_version,migration_checksum) VALUES($1,'0002_workflow_run_diagnostics','workflow_run_store_v2',$2)`, workflowrunmigrations.Component, diagnosticsChecksum); err != nil { t.Fatal(err) }
	if pending, inspectErr := workflowrunmigrations.Inspect(ctx, pool); inspectErr != nil || pending.MigrationState != workflowrunmigrations.MigrationStatePending { t.Fatalf("diagnostics marker was not pending: %#v %v", pending, inspectErr) }
	if upgraded, upgradeErr := workflowrunmigrations.Apply(ctx, pool); upgradeErr != nil || upgraded.MigrationState != workflowrunmigrations.MigrationStateApplied { t.Fatalf("diagnostics migration upgrade failed: %#v %v", upgraded, upgradeErr) }
}
