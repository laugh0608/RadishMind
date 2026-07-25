package httpapi

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSQLiteAgentCopilotRuntimeRestartCASCorruptionAndSensitiveMaterial(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "agent-copilot-runtime.db")
	firstRuntime := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	ctx := agentCopilotBatchCRuntimeContext()
	assignment, event := validAgentCopilotRuntimeMutation(t, ctx)
	repository := newSQLiteAgentCopilotRuntimeRepository(firstRuntime.DB())
	if err := repository.Apply(ctx, 0, assignment, event); err != nil {
		t.Fatalf("apply SQLite Agent Copilot assignment: %v", err)
	}
	if err := repository.Apply(ctx, 0, assignment, event); !errorsIsAgentCopilotRuntimeVersionConflict(err) {
		t.Fatalf("stale SQLite Agent Copilot assignment must fail CAS: %v", err)
	}
	var storedAssignment, storedEvent string
	if err := firstRuntime.DB().QueryRow(`SELECT sanitized_assignment_payload FROM agent_copilot_runtime_assignments
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef,
	).Scan(&storedAssignment); err != nil {
		t.Fatalf("read stored Agent Copilot assignment: %v", err)
	}
	if err := firstRuntime.DB().QueryRow(`SELECT sanitized_event_payload FROM agent_copilot_runtime_assignment_events
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef,
	).Scan(&storedEvent); err != nil {
		t.Fatalf("read stored Agent Copilot event: %v", err)
	}
	for _, forbidden := range []string{"profile_source", "credential", "token", "header", "endpoint", "dsn", "system_prompt", "provider"} {
		if strings.Contains(strings.ToLower(storedAssignment), forbidden) || strings.Contains(strings.ToLower(storedEvent), forbidden) {
			t.Fatalf("Agent Copilot runtime storage leaked forbidden material %q", forbidden)
		}
	}
	if err := firstRuntime.Close(); err != nil {
		t.Fatalf("close first SQLite Agent Copilot runtime: %v", err)
	}
	secondRuntime := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	defer secondRuntime.Close()
	restarted := newSQLiteAgentCopilotRuntimeRepository(secondRuntime.DB())
	restored, events, err := restarted.Read(ctx)
	if err != nil || restored.AssignmentDigest != assignment.AssignmentDigest || len(events) != 1 ||
		events[0].EventID != event.EventID {
		t.Fatalf("restore SQLite Agent Copilot assignment: assignment=%#v events=%#v err=%v", restored, events, err)
	}
	if _, err = secondRuntime.DB().Exec(`UPDATE agent_copilot_runtime_assignments
SET sanitized_assignment_payload=json_set(sanitized_assignment_payload,'$.assignment_digest',?)
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=?`,
		"sha256:"+strings.Repeat("f", 64), ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef,
	); err == nil {
		t.Fatal("SQLite Agent Copilot assignment trigger accepted an uncontrolled update")
	}
}
