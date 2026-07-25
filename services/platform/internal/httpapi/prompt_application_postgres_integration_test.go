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

func TestPostgresPromptApplicationInvocationAndSessionRestartPrivacy(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeURL := postgresIntegrationDatabaseURLForCredentials(
		t,
		runtimeUser,
		os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := workflowrunmigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, admin)
	resetPostgresWorkflowRunSchema(t, ctx, admin)
	preparePostgresIntegrationRuntimeRole(t, ctx, admin, runtimeUser)
	t.Cleanup(func() {
		cleanup, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		resetPostgresWorkflowRunSchema(t, cleanup, admin)
		admin.Close()
	})
	state, err := workflowrunmigrations.Apply(ctx, admin)
	if err != nil || state.MigrationState != workflowrunmigrations.MigrationStateApplied ||
		state.StoreSchemaVersion != workflowrunmigrations.StoreSchemaVersion {
		t.Fatalf("apply Prompt application runtime projections: %#v %v", state, err)
	}
	runtimePool, err := workflowrunmigrations.OpenPool(ctx, runtimeURL)
	if err != nil {
		t.Fatal(err)
	}

	fixture := newPromptApplicationInvocationFixture(t)
	fixture.ctx.RequestContext = ctx
	legacyRuns := newPostgresWorkflowRunStore(runtimePool)
	promptRuns, err := newPromptApplicationRunStoreForWorkflowRunStore(legacyRuns)
	if err != nil {
		t.Fatalf("create PostgreSQL Prompt run store: %v", err)
	}
	combinedRuns := newCombinedWorkflowRunStore(legacyRuns, promptRuns)
	fixture.service.runStore = combinedRuns

	legacySessions := newPostgresApplicationInteractionSessionRepository(runtimePool)
	promptSessions, err := newPromptApplicationSessionRepositoryForLegacy(legacySessions)
	if err != nil {
		t.Fatalf("create PostgreSQL Prompt session repository: %v", err)
	}
	sessionRepository := newCombinedApplicationInteractionSessionRepository(legacySessions, promptSessions)
	resolver := newExactApplicationInteractionAuthorityResolver(
		fixture.catalog,
		nil,
		nil,
		workflowRAGApplicationAuthorityResolver{},
	)
	resolver.resolvePrompt = func(runtimeContext PromptApplicationRuntimeContext) (PromptApplicationRuntimeAuthorityV2, string) {
		authority, failure := fixture.service.resolveAuthority(runtimeContext)
		return authority.Snapshot, failure
	}
	sessions := newApplicationInteractionSessionService(sessionRepository, resolver)
	sessionContext := ApplicationInteractionContext{
		RequestContext:  ctx,
		RequestID:       fixture.ctx.RequestID,
		TenantRef:       fixture.ctx.TenantRef,
		WorkspaceID:     fixture.ctx.WorkspaceID,
		ApplicationID:   fixture.ctx.ApplicationID,
		ActorRef:        fixture.ctx.ActorRef,
		OwnerSubjectRef: fixture.ctx.OwnerSubjectRef,
		AuditRef:        fixture.ctx.AuditRef,
		WriteEnabled:    true,
	}
	created := sessions.Create(sessionContext, ApplicationInteractionSessionCreateInput{
		ProfileBinding: ApplicationInteractionProfileBinding{
			ExecutionProfile: applicationInteractionProfilePrompt,
		},
	})
	if created.FailureCode != "" || created.Session == nil ||
		created.Session.SchemaVersion != promptApplicationSessionV2Schema {
		t.Fatalf("create PostgreSQL Prompt session: %#v", created)
	}
	sensitiveInput := "postgres-private-prompt-input-must-not-persist"
	coordinator := newApplicationInteractionTurnCoordinator(sessions, resolver, nil, nil, fixture.service.Invoke)
	completed := coordinator.Execute(
		sessionContext,
		created.Session.SessionID,
		ApplicationInteractionTurnExecutionInput{
			ExpectedSessionVersion: created.Session.RecordVersion,
			ClientTurnKey:          "postgres-prompt-turn",
			PromptVariables: map[string]any{
				"question": sensitiveInput,
				"tone":     "clear",
			},
		},
	)
	if completed.FailureCode != "" || completed.Turn == nil || completed.Turn.RunRef == nil ||
		completed.Turn.RunRef.SchemaVersion != workflowRunRecordPromptSchemaVersion ||
		completed.PromptOutput == "" || fixture.bridge.callCount() != 1 {
		t.Fatalf("execute PostgreSQL Prompt turn: %#v calls=%d", completed, fixture.bridge.callCount())
	}
	runID := completed.Turn.RunRef.RunID
	var persistedPayloads string
	if err = runtimePool.QueryRow(ctx, `SELECT
		r.sanitized_run_payload::text || s.sanitized_session_payload::text || t.sanitized_turn_payload::text
		FROM prompt_application_run_records r
		JOIN prompt_application_sessions s
		  ON s.tenant_ref=r.tenant_ref AND s.workspace_id=r.workspace_id AND s.application_id=r.application_id
		JOIN prompt_application_session_turns t
		  ON t.tenant_ref=s.tenant_ref AND t.workspace_id=s.workspace_id AND t.application_id=s.application_id
		 AND t.owner_subject_ref=s.owner_subject_ref AND t.session_id=s.session_id
		WHERE r.run_id=$1 AND s.session_id=$2`,
		runID,
		created.Session.SessionID,
	).Scan(&persistedPayloads); err != nil {
		t.Fatalf("read PostgreSQL Prompt sanitized payloads: %v", err)
	}
	if strings.Contains(persistedPayloads, sensitiveInput) ||
		strings.Contains(persistedPayloads, completed.PromptOutput) ||
		strings.Contains(persistedPayloads, "rendered_messages") ||
		strings.Contains(persistedPayloads, "provider_raw_response") {
		t.Fatalf("PostgreSQL Prompt persistence leaked transient content: %s", persistedPayloads)
	}
	if _, err = runtimePool.Exec(ctx, `DELETE FROM prompt_application_run_records WHERE run_id=$1`, runID); err == nil {
		t.Fatal("PostgreSQL Prompt run accepted DELETE")
	}
	if _, err = runtimePool.Exec(ctx, `UPDATE prompt_application_session_turns SET client_turn_key='changed' WHERE session_id=$1`, created.Session.SessionID); err == nil {
		t.Fatal("PostgreSQL Prompt turn accepted identity mutation")
	}

	runtimePool.Close()
	reopened, err := workflowrunmigrations.OpenPool(ctx, runtimeURL)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	restartedLegacyRuns := newPostgresWorkflowRunStore(reopened)
	restartedPromptRuns, err := newPromptApplicationRunStoreForWorkflowRunStore(restartedLegacyRuns)
	if err != nil {
		t.Fatalf("recreate PostgreSQL Prompt run store: %v", err)
	}
	restartedRuns := newCombinedWorkflowRunStore(restartedLegacyRuns, restartedPromptRuns)
	runContext := WorkflowRunContext{
		RequestContext: ctx,
		TenantRef:      fixture.ctx.TenantRef,
		WorkspaceID:    fixture.ctx.WorkspaceID,
		ApplicationID:  fixture.ctx.ApplicationID,
	}
	restoredRun, found, err := restartedRuns.ReadRun(runContext, runID)
	if err != nil || !found || restoredRun.SchemaVersion != workflowRunRecordPromptSchemaVersion ||
		restoredRun.Status != WorkflowRunStatusSucceeded || restoredRun.Output != "" {
		t.Fatalf("restart PostgreSQL Prompt run: found=%t run=%#v err=%v", found, restoredRun, err)
	}
	restartedLegacySessions := newPostgresApplicationInteractionSessionRepository(reopened)
	restartedPromptSessions, err := newPromptApplicationSessionRepositoryForLegacy(restartedLegacySessions)
	if err != nil {
		t.Fatalf("recreate PostgreSQL Prompt session repository: %v", err)
	}
	restartedSessions := newApplicationInteractionSessionService(
		newCombinedApplicationInteractionSessionRepository(restartedLegacySessions, restartedPromptSessions),
		resolver,
	)
	restoredSession := restartedSessions.Read(sessionContext, created.Session.SessionID)
	restoredTurns, failure := restartedSessions.ListTurns(sessionContext, created.Session.SessionID)
	if restoredSession.FailureCode != "" || restoredSession.Session == nil ||
		restoredSession.Session.SchemaVersion != promptApplicationSessionV2Schema ||
		failure != "" || len(restoredTurns) != 1 || restoredTurns[0].RunRef == nil ||
		restoredTurns[0].RunRef.RunID != runID {
		t.Fatalf("restart PostgreSQL Prompt session: session=%#v turns=%#v failure=%s", restoredSession, restoredTurns, failure)
	}
}

func TestPostgresAgentCopilotInvocationAndSessionRestartPrivacy(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeURL := postgresIntegrationDatabaseURLForCredentials(
		t,
		runtimeUser,
		os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := workflowrunmigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, admin)
	resetPostgresWorkflowRunSchema(t, ctx, admin)
	preparePostgresIntegrationRuntimeRole(t, ctx, admin, runtimeUser)
	t.Cleanup(func() {
		cleanup, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		resetPostgresWorkflowRunSchema(t, cleanup, admin)
		admin.Close()
	})
	state, err := workflowrunmigrations.Apply(ctx, admin)
	if err != nil || state.MigrationState != workflowrunmigrations.MigrationStateApplied ||
		state.StoreSchemaVersion != workflowrunmigrations.StoreSchemaVersion {
		t.Fatalf("apply Agent Copilot invocation projections: %#v %v", state, err)
	}
	runtimePool, err := workflowrunmigrations.OpenPool(ctx, runtimeURL)
	if err != nil {
		t.Fatal(err)
	}

	fixture := newAgentCopilotInvocationFixture(t)
	fixture.ctx.RequestContext = ctx
	legacyRuns := newPostgresWorkflowRunStore(runtimePool)
	promptRuns, err := newPromptApplicationRunStoreForWorkflowRunStore(legacyRuns)
	if err != nil {
		t.Fatalf("create PostgreSQL Prompt run store: %v", err)
	}
	agentRuns, err := newAgentCopilotRunStoreForWorkflowRunStore(legacyRuns)
	if err != nil {
		t.Fatalf("create PostgreSQL Agent Copilot run store: %v", err)
	}
	combinedRuns := newCombinedWorkflowRunStoreWithAgent(legacyRuns, promptRuns, agentRuns)
	fixture.service.runStore = combinedRuns

	legacySessions := newPostgresApplicationInteractionSessionRepository(runtimePool)
	promptSessions, err := newPromptApplicationSessionRepositoryForLegacy(legacySessions)
	if err != nil {
		t.Fatalf("create PostgreSQL Prompt session repository: %v", err)
	}
	agentSessions, err := newAgentCopilotSessionRepositoryForLegacy(legacySessions)
	if err != nil {
		t.Fatalf("create PostgreSQL Agent Copilot session repository: %v", err)
	}
	sessionRepository := newCombinedApplicationInteractionSessionRepositoryWithAgent(legacySessions, promptSessions, agentSessions)
	resolver := newExactApplicationInteractionAuthorityResolver(fixture.catalog, nil, nil, workflowRAGApplicationAuthorityResolver{})
	resolver.resolveAgentCopilot = func(runtimeContext AgentCopilotRuntimeContext) (AgentCopilotRuntimeAuthorityV3, string) {
		authority, failure := fixture.service.resolveAuthority(runtimeContext)
		return authority.Snapshot, failure
	}
	sessions := newApplicationInteractionSessionService(sessionRepository, resolver)
	sessionContext := ApplicationInteractionContext{
		RequestContext: ctx, RequestID: fixture.ctx.RequestID, TenantRef: fixture.ctx.TenantRef,
		WorkspaceID: fixture.ctx.WorkspaceID, ApplicationID: fixture.ctx.ApplicationID,
		ActorRef: fixture.ctx.ActorRef, OwnerSubjectRef: fixture.ctx.OwnerSubjectRef,
		AuditRef: fixture.ctx.AuditRef, WriteEnabled: true,
	}
	created := sessions.Create(sessionContext, ApplicationInteractionSessionCreateInput{
		ProfileBinding: ApplicationInteractionProfileBinding{ExecutionProfile: applicationInteractionProfileAgentCopilot},
	})
	if created.FailureCode != "" || created.Session == nil ||
		created.Session.SchemaVersion != agentCopilotSessionV3Schema {
		t.Fatalf("create PostgreSQL Agent Copilot session: %#v", created)
	}
	input := validAgentCopilotInvocationInput()
	coordinator := newApplicationInteractionTurnCoordinator(sessions, resolver, nil, nil).withAgentCopilot(fixture.service.Invoke)
	completed := coordinator.Execute(sessionContext, created.Session.SessionID, ApplicationInteractionTurnExecutionInput{
		ExpectedSessionVersion: created.Session.RecordVersion, ClientTurnKey: "postgres-agent-turn",
		AgentTask: input.Task, AgentLocale: input.Locale, AgentConversationID: input.ConversationID,
		AgentArtifacts: input.Artifacts, AgentContext: input.Context,
	})
	if completed.FailureCode != "" || completed.Turn == nil || completed.Turn.RunRef == nil ||
		completed.Turn.RunRef.SchemaVersion != agentCopilotRunV7Schema ||
		completed.AgentResponse == nil || fixture.bridge.callCount() != 1 {
		t.Fatalf("execute PostgreSQL Agent Copilot turn: %#v calls=%d", completed, fixture.bridge.callCount())
	}
	runID := completed.Turn.RunRef.RunID
	var persistedPayloads string
	if err = runtimePool.QueryRow(ctx, `SELECT
		r.sanitized_run_payload::text || s.sanitized_session_payload::text || t.sanitized_turn_payload::text
		FROM agent_copilot_run_records r
		JOIN agent_copilot_sessions s
		  ON s.tenant_ref=r.tenant_ref AND s.workspace_id=r.workspace_id AND s.application_id=r.application_id
		JOIN agent_copilot_session_turns t
		  ON t.tenant_ref=s.tenant_ref AND t.workspace_id=s.workspace_id AND t.application_id=s.application_id
		 AND t.owner_subject_ref=s.owner_subject_ref AND t.session_id=s.session_id
		WHERE r.run_id=$1 AND s.session_id=$2`,
		runID,
		created.Session.SessionID,
	).Scan(&persistedPayloads); err != nil {
		t.Fatalf("read PostgreSQL Agent Copilot sanitized payloads: %v", err)
	}
	for _, forbidden := range []string{"private selection", "selected_unit_ids", "structured_answer", "provider_raw_response"} {
		if strings.Contains(persistedPayloads, forbidden) {
			t.Fatalf("PostgreSQL Agent Copilot persistence leaked %q: %s", forbidden, persistedPayloads)
		}
	}
	if _, err = runtimePool.Exec(ctx, `DELETE FROM agent_copilot_run_records WHERE run_id=$1`, runID); err == nil {
		t.Fatal("PostgreSQL Agent Copilot run accepted DELETE")
	}
	if _, err = runtimePool.Exec(ctx, `UPDATE agent_copilot_session_turns SET client_turn_key='changed' WHERE session_id=$1`, created.Session.SessionID); err == nil {
		t.Fatal("PostgreSQL Agent Copilot turn accepted identity mutation")
	}

	runtimePool.Close()
	reopened, err := workflowrunmigrations.OpenPool(ctx, runtimeURL)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	restartedLegacyRuns := newPostgresWorkflowRunStore(reopened)
	restartedPromptRuns, err := newPromptApplicationRunStoreForWorkflowRunStore(restartedLegacyRuns)
	if err != nil {
		t.Fatalf("recreate PostgreSQL Prompt run store: %v", err)
	}
	restartedAgentRuns, err := newAgentCopilotRunStoreForWorkflowRunStore(restartedLegacyRuns)
	if err != nil {
		t.Fatalf("recreate PostgreSQL Agent Copilot run store: %v", err)
	}
	restartedRuns := newCombinedWorkflowRunStoreWithAgent(restartedLegacyRuns, restartedPromptRuns, restartedAgentRuns)
	restoredRun, found, err := restartedRuns.ReadRun(agentCopilotWorkflowRunContext(fixture.ctx), runID)
	if err != nil || !found || restoredRun.SchemaVersion != agentCopilotRunV7Schema ||
		restoredRun.Status != WorkflowRunStatusSucceeded || restoredRun.Output != "" {
		t.Fatalf("restart PostgreSQL Agent Copilot run: found=%t run=%#v err=%v", found, restoredRun, err)
	}
	restartedLegacySessions := newPostgresApplicationInteractionSessionRepository(reopened)
	restartedPromptSessions, err := newPromptApplicationSessionRepositoryForLegacy(restartedLegacySessions)
	if err != nil {
		t.Fatalf("recreate PostgreSQL Prompt session repository: %v", err)
	}
	restartedAgentSessions, err := newAgentCopilotSessionRepositoryForLegacy(restartedLegacySessions)
	if err != nil {
		t.Fatalf("recreate PostgreSQL Agent Copilot session repository: %v", err)
	}
	restartedSessions := newApplicationInteractionSessionService(
		newCombinedApplicationInteractionSessionRepositoryWithAgent(restartedLegacySessions, restartedPromptSessions, restartedAgentSessions),
		resolver,
	)
	restoredSession := restartedSessions.Read(sessionContext, created.Session.SessionID)
	restoredTurns, failure := restartedSessions.ListTurns(sessionContext, created.Session.SessionID)
	if restoredSession.FailureCode != "" || restoredSession.Session == nil ||
		restoredSession.Session.SchemaVersion != agentCopilotSessionV3Schema ||
		failure != "" || len(restoredTurns) != 1 || restoredTurns[0].RunRef == nil ||
		restoredTurns[0].RunRef.RunID != runID {
		t.Fatalf("restart PostgreSQL Agent Copilot session: session=%#v turns=%#v failure=%s", restoredSession, restoredTurns, failure)
	}
}
