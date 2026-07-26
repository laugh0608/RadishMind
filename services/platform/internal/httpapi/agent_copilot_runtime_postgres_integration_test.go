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

func TestPostgresAgentCopilotRuntimeRestartConcurrentCASCorruptionAndSensitiveMaterial(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeDatabaseURL := postgresIntegrationDatabaseURLForCredentials(
		t, runtimeUser, os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"),
	)
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
	repository := newPostgresAgentCopilotRuntimeRepository(runtimePool)
	runtimeContext := agentCopilotBatchCRuntimeContext()
	assignment, event := validAgentCopilotRuntimeMutation(t, runtimeContext)

	const writers = 8
	var wait sync.WaitGroup
	results := make(chan error, writers)
	for range writers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			results <- repository.Apply(runtimeContext, 0, assignment, event)
		}()
	}
	wait.Wait()
	close(results)
	successes, conflicts := 0, 0
	for result := range results {
		switch {
		case result == nil:
			successes++
		case errorsIsAgentCopilotRuntimeVersionConflict(result):
			conflicts++
		default:
			t.Fatalf("unexpected PostgreSQL Agent Copilot CAS result: %v", result)
		}
	}
	if successes != 1 || conflicts != writers-1 {
		t.Fatalf("PostgreSQL Agent Copilot CAS accepted %d successes and %d conflicts", successes, conflicts)
	}
	restored, events, err := newPostgresAgentCopilotRuntimeRepository(runtimePool).Read(runtimeContext)
	if err != nil || restored.AssignmentDigest != assignment.AssignmentDigest ||
		len(events) != 1 || events[0].EventID != event.EventID {
		t.Fatalf("restore PostgreSQL Agent Copilot assignment: assignment=%#v events=%#v err=%v", restored, events, err)
	}
	var assignmentPayload, eventPayload string
	if err = runtimePool.QueryRow(ctx, `SELECT sanitized_assignment_payload::text FROM agent_copilot_runtime_assignments
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4`,
		runtimeContext.TenantRef, runtimeContext.WorkspaceID, runtimeContext.ApplicationID, runtimeContext.OwnerSubjectRef,
	).Scan(&assignmentPayload); err != nil {
		t.Fatalf("read PostgreSQL Agent Copilot assignment payload: %v", err)
	}
	if err = runtimePool.QueryRow(ctx, `SELECT sanitized_event_payload::text FROM agent_copilot_runtime_assignment_events
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4`,
		runtimeContext.TenantRef, runtimeContext.WorkspaceID, runtimeContext.ApplicationID, runtimeContext.OwnerSubjectRef,
	).Scan(&eventPayload); err != nil {
		t.Fatalf("read PostgreSQL Agent Copilot event payload: %v", err)
	}
	for _, forbidden := range []string{"profile_source", "credential", "token", "header", "endpoint", "dsn", "system_prompt", "provider"} {
		if strings.Contains(strings.ToLower(assignmentPayload), forbidden) ||
			strings.Contains(strings.ToLower(eventPayload), forbidden) {
			t.Fatalf("PostgreSQL Agent Copilot runtime leaked forbidden material %q", forbidden)
		}
	}
	if _, err = runtimePool.Exec(ctx, `UPDATE agent_copilot_runtime_assignments
SET sanitized_assignment_payload=jsonb_set(sanitized_assignment_payload,'{assignment_digest}',to_jsonb($1::text))
WHERE tenant_ref=$2 AND workspace_id=$3 AND application_id=$4 AND owner_subject_ref=$5`,
		"sha256:"+strings.Repeat("f", 64), runtimeContext.TenantRef, runtimeContext.WorkspaceID,
		runtimeContext.ApplicationID, runtimeContext.OwnerSubjectRef,
	); err == nil {
		t.Fatal("PostgreSQL Agent Copilot assignment trigger accepted uncontrolled corruption")
	}
}
