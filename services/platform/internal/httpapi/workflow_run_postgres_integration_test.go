//go:build postgres_integration

package httpapi

import (
	"context"
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
		if err = store.UpsertRun(runContext, &record); err != nil {
			t.Fatal(err)
		}
	}
	page, err := store.ListRuns(runContext, WorkflowRunListFilter{Limit: 2})
	if err != nil || len(page.Records) != 2 || !page.HasMore || page.Records[0].RunID != "run_pg_c" {
		t.Fatalf("unexpected PostgreSQL page: %#v %v", page, err)
	}
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
	recovered, found, err := store.ReadRun(runContext, "run_pg_c")
	if err != nil || !found || recovered.Status != WorkflowRunStatusSucceeded {
		t.Fatalf("restart recovery failed: %#v %v", recovered, err)
	}
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
}
