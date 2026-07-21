//go:build postgres_integration

package httpapi

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	workflowrunmigrations "radishmind.local/services/platform/migrations/workflow_runs"
)

func TestPostgresPromptApplicationRuntimeRestartAndCAS(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeDatabaseURL := postgresIntegrationDatabaseURLForCredentials(t, runtimeUser, os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	adminPool, err := workflowrunmigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, adminPool)
	resetPostgresWorkflowRunSchema(t, ctx, adminPool)
	preparePostgresIntegrationRuntimeRole(t, ctx, adminPool, runtimeUser)
	t.Cleanup(func() {
		cleanup, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		resetPostgresWorkflowRunSchema(t, cleanup, adminPool)
		adminPool.Close()
	})
	if _, err = workflowrunmigrations.Apply(ctx, adminPool); err != nil {
		t.Fatalf("apply Workflow migration family: %v", err)
	}
	runtimePool, err := workflowrunmigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer runtimePool.Close()
	repository := newPostgresPromptApplicationRuntimeRepository(runtimePool)
	runtimeContext := validPromptApplicationRuntimeContext()
	assignment, event := validPromptApplicationRuntimeMutation(t, runtimeContext)
	if err = repository.Apply(runtimeContext, 0, assignment, event); err != nil {
		t.Fatalf("apply PostgreSQL Prompt assignment: %v", err)
	}
	if err = repository.Apply(runtimeContext, 0, assignment, event); !errorsIsPromptRuntimeVersionConflict(err) {
		t.Fatalf("stale PostgreSQL Prompt assignment must fail CAS: %v", err)
	}
	restored, events, err := newPostgresPromptApplicationRuntimeRepository(runtimePool).Read(runtimeContext)
	if err != nil || restored.AssignmentDigest != assignment.AssignmentDigest || len(events) != 1 || events[0].EventID != event.EventID {
		t.Fatalf("restore PostgreSQL Prompt assignment: assignment=%#v events=%#v err=%v", restored, events, err)
	}
}
