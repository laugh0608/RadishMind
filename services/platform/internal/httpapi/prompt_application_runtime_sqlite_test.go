package httpapi

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSQLitePromptApplicationRuntimeRestartAndCAS(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "prompt-application-runtime.db")
	firstRuntime := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	ctx := validPromptApplicationRuntimeContext()
	assignment, event := validPromptApplicationRuntimeMutation(t, ctx)
	repository := newSQLitePromptApplicationRuntimeRepository(firstRuntime.DB())
	if err := repository.Apply(ctx, 0, assignment, event); err != nil {
		t.Fatalf("apply SQLite Prompt assignment: %v", err)
	}
	if err := repository.Apply(ctx, 0, assignment, event); !errorsIsPromptRuntimeVersionConflict(err) {
		t.Fatalf("stale SQLite Prompt assignment must fail CAS: %v", err)
	}
	if err := firstRuntime.Close(); err != nil {
		t.Fatalf("close first SQLite runtime: %v", err)
	}
	secondRuntime := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	defer secondRuntime.Close()
	restored, events, err := newSQLitePromptApplicationRuntimeRepository(secondRuntime.DB()).Read(ctx)
	if err != nil || restored.AssignmentDigest != assignment.AssignmentDigest || len(events) != 1 || events[0].EventID != event.EventID {
		t.Fatalf("restore SQLite Prompt assignment: assignment=%#v events=%#v err=%v", restored, events, err)
	}
}

func validPromptApplicationRuntimeContext() PromptApplicationRuntimeContext {
	return PromptApplicationRuntimeContext{
		RequestContext: validApplicationDraftContext().RequestContext, RequestID: "request_prompt_runtime_store", TenantRef: "tenant_demo",
		WorkspaceID: "workspace_demo", ApplicationID: "app_aaaaaaaaaaaaaaaa", ActorRef: "subject_owner",
		OwnerSubjectRef: "subject_owner", AuditRef: "audit_prompt_runtime_store", WriteEnabled: true,
	}
}

func validPromptApplicationRuntimeMutation(t *testing.T, ctx PromptApplicationRuntimeContext) (PromptApplicationRuntimeAssignment, PromptApplicationRuntimeAssignmentEvent) {
	t.Helper()
	at := "2026-07-21T13:00:00Z"
	ref := PromptApplicationTemplateRef{TemplateID: "ptpl_aaaaaaaaaaaaaaaa", TemplateVersion: 1, TemplateDigest: "sha256:" + strings.Repeat("a", 64)}
	assignment := PromptApplicationRuntimeAssignment{
		SchemaVersion: promptApplicationRuntimeAssignmentSchema, AssignmentID: "ptra_aaaaaaaaaaaaaaaa", TenantRef: ctx.TenantRef,
		WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, OwnerSubjectRef: ctx.OwnerSubjectRef, AssignmentVersion: 1,
		State: "active", CandidateID: "candidate-prompt-runtime", CandidateReviewVersion: 1, DraftID: "draft-prompt-runtime", DraftVersion: 2,
		DraftDigest: "sha256:" + strings.Repeat("b", 64), PromptTemplateRef: ref, ActivatedAt: at, UpdatedAt: at,
		ActivatedByActorRef: ctx.ActorRef, UpdatedByActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	digest, err := promptApplicationRuntimeAssignmentDigest(assignment)
	if err != nil {
		t.Fatalf("digest Prompt assignment: %v", err)
	}
	assignment.AssignmentDigest = digest
	event := PromptApplicationRuntimeAssignmentEvent{
		SchemaVersion: promptApplicationRuntimeAssignmentEventSchema, EventID: "ptrae_aaaaaaaaaaaaaaaa", AssignmentID: assignment.AssignmentID,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, OwnerSubjectRef: ctx.OwnerSubjectRef,
		EventSequence: 1, Action: "activate", ExpectedAssignmentVersion: 0, ResultingAssignmentVersion: 1,
		CandidateID: assignment.CandidateID, CandidateReviewVersion: assignment.CandidateReviewVersion, DraftID: assignment.DraftID,
		DraftVersion: assignment.DraftVersion, DraftDigest: assignment.DraftDigest, PromptTemplateRef: ref, AssignmentDigest: digest,
		OccurredAt: at, ActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	return assignment, event
}

func errorsIsPromptRuntimeVersionConflict(err error) bool {
	return err == errPromptApplicationRuntimeVersionConflict
}
