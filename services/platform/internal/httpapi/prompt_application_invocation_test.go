package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
)

type promptApplicationInvocationFixture struct {
	service    promptApplicationInvocationService
	ctx        PromptApplicationRuntimeContext
	runStore   workflowRunStore
	runtime    *memoryPromptApplicationRuntimeRepository
	candidates *memoryApplicationPublishCandidateRepository
	bridge     *workflowExecutorTestBridge
	catalog    *memoryApplicationCatalogRepository
	catalogCtx ApplicationCatalogContext
}

func TestPromptApplicationInvocationUsesExactAuthorityAndDoesNotReplay(t *testing.T) {
	fixture := newPromptApplicationInvocationFixture(t)
	input := PromptApplicationInvocationInput{
		ClientInvocationKey: "client-invocation-1",
		Variables:           map[string]any{"question": "How is the release reviewed?", "tone": "clear"},
	}
	result := fixture.service.Invoke(fixture.ctx, input)
	if result.FailureCode != "" || result.Run == nil || result.Run.SchemaVersion != workflowRunRecordPromptSchemaVersion ||
		result.Run.Status != WorkflowRunStatusSucceeded || result.Output == "" || fixture.bridge.callCount() != 1 {
		document, _ := promptApplicationRunDocument(*result.Run)
		t.Fatalf("Prompt invocation did not complete exactly once: %#v run=%+v diagnostic=%+v strict=%v authority=%v names=%v selection=%v usage=%v side=%v calls=%d",
			result, result.Run, result.Run.PromptDiagnostic, validatePromptApplicationRunRecordV6(document),
			validatePromptApplicationRuntimeAuthorityV2(document.Authority),
			validPromptApplicationVariableNames(document.VariableNames, document.VariableNamesDigest),
			validPromptApplicationSafeSelection(document.RequestedModel, document.SelectedProvider, document.SelectedProfile, document.SelectedModel, document.UpstreamModel, document.SelectionSource),
			validPromptApplicationRunUsage(document.Usage), validPromptApplicationRunSideEffects(document.SideEffects), fixture.bridge.callCount())
	}
	if result.Run.Output != "" || result.Run.SideEffects.ProviderCalls != 1 || result.Run.SideEffects.RetrievalCalls != 0 ||
		result.Run.PromptApplication == nil || result.Run.VariableNamesDigest == "" ||
		strings.Contains(string(fixture.bridge.lastRequest()), "client-invocation-1") {
		t.Fatalf("Prompt invocation metadata boundary is invalid: %#v request=%s", result.Run, fixture.bridge.lastRequest())
	}
	replayed := fixture.service.Invoke(fixture.ctx, input)
	if !replayed.IdempotentReplay || replayed.Run == nil || replayed.Output != "" || replayed.FailureCode != "" || fixture.bridge.callCount() != 1 {
		t.Fatalf("terminal idempotent retry replayed provider or fabricated output: %#v calls=%d", replayed, fixture.bridge.callCount())
	}
	conflict := fixture.service.Invoke(fixture.ctx, PromptApplicationInvocationInput{
		ClientInvocationKey: input.ClientInvocationKey,
		Variables:           map[string]any{"question": "different input", "tone": "clear"},
	})
	if conflict.FailureCode != PromptApplicationInvocationFailureInputInvalid || fixture.bridge.callCount() != 1 {
		t.Fatalf("idempotency key accepted different input: %#v calls=%d", conflict, fixture.bridge.callCount())
	}
}

func TestPromptApplicationInvocationFailsBeforeProviderOnAuthorityDrift(t *testing.T) {
	fixture := newPromptApplicationInvocationFixture(t)
	fixture.service.runStore = authorityDriftPromptRunStore{
		workflowRunStore: fixture.runStore,
		drift: func() {
			record, err := fixture.catalog.Read(fixture.catalogCtx, fixture.ctx.ApplicationID)
			if err != nil {
				t.Fatalf("read catalog before drift: %v", err)
			}
			record.Description = "Changed after run reservation."
			record.UpdatedAt = "2026-07-22T08:05:30Z"
			if _, err = fixture.catalog.UpdateMetadata(fixture.catalogCtx, record.ApplicationID, record.RecordVersion, record); err != nil {
				t.Fatalf("drift catalog: %v", err)
			}
		},
	}
	result := fixture.service.Invoke(fixture.ctx, PromptApplicationInvocationInput{
		ClientInvocationKey: "client-invocation-drift",
		Variables:           map[string]any{"question": "authority drift", "tone": "clear"},
	})
	if result.FailureCode != PromptApplicationRuntimeFailureAuthorityChanged || result.Run == nil ||
		result.Run.Status != WorkflowRunStatusFailed || fixture.bridge.callCount() != 0 {
		t.Fatalf("authority drift crossed provider checkpoint: %#v calls=%d", result, fixture.bridge.callCount())
	}
}

func TestPromptApplicationInvocationConcurrentDuplicateCallsProviderOnce(t *testing.T) {
	fixture := newPromptApplicationInvocationFixture(t)
	entered, release := make(chan struct{}), make(chan struct{})
	fixture.bridge.handle = func(context.Context, []byte, bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
		close(entered)
		<-release
		return successfulWorkflowExecutorEnvelope("one controlled output"), nil
	}
	input := PromptApplicationInvocationInput{
		ClientInvocationKey: "client-concurrent",
		Variables:           map[string]any{"question": "private concurrent input", "tone": "clear"},
	}
	firstResult := make(chan PromptApplicationInvocationResult, 1)
	go func() { firstResult <- fixture.service.Invoke(fixture.ctx, input) }()
	<-entered
	duplicate := fixture.service.Invoke(fixture.ctx, input)
	if duplicate.FailureCode != PromptApplicationInvocationFailureDuplicateRunning ||
		duplicate.Run == nil || duplicate.Run.Status != WorkflowRunStatusRunning ||
		fixture.bridge.callCount() != 1 {
		t.Fatalf("concurrent duplicate crossed provider boundary: %#v calls=%d", duplicate, fixture.bridge.callCount())
	}
	close(release)
	first := <-firstResult
	if first.FailureCode != "" || first.Run == nil || first.Run.Status != WorkflowRunStatusSucceeded ||
		fixture.bridge.callCount() != 1 {
		t.Fatalf("first concurrent invocation did not finish exactly once: %#v calls=%d", first, fixture.bridge.callCount())
	}
}

func TestPromptApplicationInvocationCancellationTimeoutOutputAndTerminalUncertaintyDoNotReplay(t *testing.T) {
	t.Run("canceled", func(t *testing.T) {
		fixture := newPromptApplicationInvocationFixture(t)
		requestContext, cancel := context.WithCancel(context.Background())
		cancel()
		fixture.ctx.RequestContext = requestContext
		fixture.bridge.handle = func(ctx context.Context, _ []byte, _ bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
			<-ctx.Done()
			return bridge.GatewayEnvelope{}, ctx.Err()
		}
		result := fixture.service.Invoke(fixture.ctx, PromptApplicationInvocationInput{
			ClientInvocationKey: "client-canceled",
			Variables:           map[string]any{"question": "private canceled input", "tone": "clear"},
		})
		if result.FailureCode != PromptApplicationInvocationFailureCanceled || result.Run == nil ||
			result.Run.Status != WorkflowRunStatusCanceled || fixture.bridge.callCount() != 1 {
			t.Fatalf("canceled Prompt invocation drifted: %#v calls=%d", result, fixture.bridge.callCount())
		}
		replayed := fixture.service.Invoke(fixture.ctx, PromptApplicationInvocationInput{
			ClientInvocationKey: "client-canceled",
			Variables:           map[string]any{"question": "private canceled input", "tone": "clear"},
		})
		if !replayed.IdempotentReplay || fixture.bridge.callCount() != 1 {
			t.Fatalf("canceled invocation replayed provider: %#v calls=%d", replayed, fixture.bridge.callCount())
		}
	})
	t.Run("timeout", func(t *testing.T) {
		fixture := newPromptApplicationInvocationFixture(t)
		fixture.service.maxRuntime = time.Millisecond
		fixture.bridge.handle = func(ctx context.Context, _ []byte, _ bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
			<-ctx.Done()
			return bridge.GatewayEnvelope{}, ctx.Err()
		}
		result := fixture.service.Invoke(fixture.ctx, PromptApplicationInvocationInput{
			ClientInvocationKey: "client-timeout",
			Variables:           map[string]any{"question": "private timeout input", "tone": "clear"},
		})
		if result.FailureCode != PromptApplicationInvocationFailureOutcomeUnknown || result.Run == nil ||
			result.Run.Status != WorkflowRunStatusOutcomeUnknown || fixture.bridge.callCount() != 1 {
			t.Fatalf("timed out Prompt invocation drifted: %#v calls=%d", result, fixture.bridge.callCount())
		}
	})
	t.Run("output_contract", func(t *testing.T) {
		fixture := newPromptApplicationInvocationFixture(t)
		fixture.bridge.handle = func(context.Context, []byte, bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
			return bridge.GatewayEnvelope{Status: "ok", Response: map[string]any{"summary": strings.Repeat("x", 4097)}}, nil
		}
		result := fixture.service.Invoke(fixture.ctx, PromptApplicationInvocationInput{
			ClientInvocationKey: "client-output-contract",
			Variables:           map[string]any{"question": "private invalid output", "tone": "clear"},
		})
		if result.FailureCode != PromptApplicationInvocationFailureOutputContract || result.Run == nil ||
			result.Run.Status != WorkflowRunStatusFailed || fixture.bridge.callCount() != 1 {
			t.Fatalf("Prompt output contract failure drifted: %#v calls=%d", result, fixture.bridge.callCount())
		}
	})
	t.Run("terminal_write_unknown", func(t *testing.T) {
		fixture := newPromptApplicationInvocationFixture(t)
		fixture.service.runStore = terminalFailWorkflowRunStore{delegate: fixture.runStore}
		input := PromptApplicationInvocationInput{
			ClientInvocationKey: "client-terminal-unknown",
			Variables:           map[string]any{"question": "private terminal input", "tone": "clear"},
		}
		result := fixture.service.Invoke(fixture.ctx, input)
		if result.FailureCode != PromptApplicationInvocationFailureOutcomeUnknown || result.Run == nil ||
			result.Run.Status != WorkflowRunStatusOutcomeUnknown || result.Run.PromptDiagnostic == nil ||
			result.Run.PromptDiagnostic.TerminalWriteState != "unknown" || fixture.bridge.callCount() != 1 {
			t.Fatalf("terminal uncertainty drifted: %#v calls=%d", result, fixture.bridge.callCount())
		}
		replayed := fixture.service.Invoke(fixture.ctx, input)
		if replayed.FailureCode != PromptApplicationInvocationFailureDuplicateRunning ||
			fixture.bridge.callCount() != 1 {
			t.Fatalf("terminal uncertainty replayed provider: %#v calls=%d", replayed, fixture.bridge.callCount())
		}
	})
}

func TestPromptApplicationSessionV2DelegatesUniqueInvocationService(t *testing.T) {
	fixture := newPromptApplicationInvocationFixture(t)
	resolver := newExactApplicationInteractionAuthorityResolver(
		fixture.catalog, nil, nil, workflowRAGApplicationAuthorityResolver{},
	)
	resolver.resolvePrompt = func(ctx PromptApplicationRuntimeContext) (PromptApplicationRuntimeAuthorityV2, string) {
		authority, failure := fixture.service.resolveAuthority(ctx)
		return authority.Snapshot, failure
	}
	repository := newCombinedApplicationInteractionSessionRepository(
		newMemoryApplicationInteractionSessionRepository(), newMemoryApplicationInteractionSessionRepository(),
	)
	sessions := newApplicationInteractionSessionService(repository, resolver)
	sessions.now = func() time.Time { return time.Date(2026, 7, 22, 8, 6, 30, 0, time.UTC) }
	sessionContext := ApplicationInteractionContext{
		RequestContext: context.Background(), RequestID: fixture.ctx.RequestID, TenantRef: fixture.ctx.TenantRef,
		WorkspaceID: fixture.ctx.WorkspaceID, ApplicationID: fixture.ctx.ApplicationID,
		ActorRef: fixture.ctx.ActorRef, OwnerSubjectRef: fixture.ctx.OwnerSubjectRef,
		AuditRef: fixture.ctx.AuditRef, WriteEnabled: true,
	}
	created := sessions.Create(sessionContext, ApplicationInteractionSessionCreateInput{
		ProfileBinding: ApplicationInteractionProfileBinding{ExecutionProfile: applicationInteractionProfilePrompt},
	})
	if created.Session == nil || created.Session.SchemaVersion != promptApplicationSessionV2Schema {
		t.Fatalf("create Prompt session v2: %#v", created)
	}
	assertPromptApplicationStrictJSON(t, *created.Session, promptApplicationSessionV2Schema)
	coordinator := newApplicationInteractionTurnCoordinator(sessions, resolver, nil, nil, fixture.service.Invoke)
	coordinator.now = func() time.Time { return time.Date(2026, 7, 22, 8, 7, 0, 0, time.UTC) }
	result := coordinator.Execute(sessionContext, created.Session.SessionID, ApplicationInteractionTurnExecutionInput{
		ExpectedSessionVersion: 1, ClientTurnKey: "prompt-turn-1",
		PromptVariables: map[string]any{"question": "Session invocation", "tone": "clear"},
	})
	if result.FailureCode != "" || result.Session == nil || result.Turn == nil ||
		result.Turn.SchemaVersion != promptApplicationSessionTurnV2Schema ||
		result.Turn.RunRef == nil || result.Turn.RunRef.SchemaVersion != workflowRunRecordPromptSchemaVersion ||
		result.PromptOutput == "" || fixture.bridge.callCount() != 1 {
		t.Fatalf("Prompt session did not delegate v6 invocation: %#v calls=%d", result, fixture.bridge.callCount())
	}
	assertPromptApplicationStrictJSON(t, *result.Turn, promptApplicationSessionTurnV2Schema)
	replayed := coordinator.Execute(sessionContext, created.Session.SessionID, ApplicationInteractionTurnExecutionInput{
		ExpectedSessionVersion: 1, ClientTurnKey: "prompt-turn-1",
		PromptVariables: map[string]any{"question": "Session invocation", "tone": "clear"},
	})
	if !replayed.IdempotentReplay || replayed.PromptOutput != "" || fixture.bridge.callCount() != 1 {
		t.Fatalf("Prompt session retry replayed provider or output: %#v calls=%d", replayed, fixture.bridge.callCount())
	}
}

func TestPromptApplicationRunFeedsHistoryComparisonEvaluationAndOperationsMetadata(t *testing.T) {
	fixture := newPromptApplicationInvocationFixture(t)
	first := fixture.service.Invoke(fixture.ctx, PromptApplicationInvocationInput{
		ClientInvocationKey: "client-metadata-a",
		Variables:           map[string]any{"question": "first private input", "tone": "clear"},
	})
	second := fixture.service.Invoke(fixture.ctx, PromptApplicationInvocationInput{
		ClientInvocationKey: "client-metadata-b",
		Variables:           map[string]any{"question": "second private input", "tone": "clear"},
	})
	if first.Run == nil || second.Run == nil || first.FailureCode != "" || second.FailureCode != "" {
		t.Fatalf("seed Prompt metadata runs: first=%#v second=%#v", first, second)
	}
	runContext := WorkflowRunContext{
		RequestContext: fixture.ctx.RequestContext, RequestID: fixture.ctx.RequestID,
		TenantRef: fixture.ctx.TenantRef, WorkspaceID: fixture.ctx.WorkspaceID,
		ApplicationID: fixture.ctx.ApplicationID, ActorRef: fixture.ctx.ActorRef, AuditRef: fixture.ctx.AuditRef,
	}
	runService := newWorkflowExecutorService(nil, nil, fixture.runStore)
	history := runService.ListRuns(runContext, WorkflowRunListRequest{
		ExecutionSourceKind: promptApplicationExecutionSourceKind,
		ExecutionSourceID:   first.Run.ExecutionSource.ID,
		Limit:               10,
	})
	if history.FailureCode != "" || len(history.Runs) != 2 ||
		history.Runs[0].ExecutionProfile != promptApplicationInvocationProfile ||
		history.Runs[0].AuthorityDigest == "" || history.Runs[0].VariableNamesDigest == "" ||
		history.Runs[0].UsageState == "" || history.Runs[0].SideEffects.ProviderCalls != 1 {
		t.Fatalf("Prompt History/Operations metadata projection drifted: %#v", history)
	}
	comparison := runService.CompareRuns(runContext, first.Run.RunID, second.Run.RunID)
	if comparison.FailureCode != "" || comparison.Comparison == nil ||
		comparison.Comparison.SchemaVersion != promptApplicationRunComparisonSchemaVersion ||
		comparison.Comparison.RunProfile != promptApplicationInvocationProfile ||
		comparison.Comparison.Classification != WorkflowRunComparisonUnchanged {
		t.Fatalf("Prompt comparison did not recognize v6 lineage: %#v", comparison)
	}
	evaluations := newWorkflowEvaluationService(newMemoryWorkflowEvaluationStore(10), fixture.runStore)
	created := evaluations.Create(runContext, WorkflowEvaluationCreateRequest{
		Name:          "Prompt metadata regression",
		BaselineRunID: first.Run.RunID,
		Expectations: []WorkflowEvaluationExpectation{{
			CandidateRunID: second.Run.RunID, ExpectedClassification: WorkflowRunComparisonUnchanged,
		}},
	})
	if created.FailureCode != "" || created.Case == nil {
		t.Fatalf("Prompt evaluation did not accept compatible v6 runs: %#v", created)
	}
	review := evaluations.Review(runContext, created.Case.CaseID)
	if review.FailureCode != "" || review.Review == nil || review.Review.Outcome != "passed" ||
		review.Review.RunProfile != promptApplicationInvocationProfile {
		t.Fatalf("Prompt evaluation review drifted: %#v", review)
	}
	payload, err := json.Marshal(first.Run)
	if err != nil {
		t.Fatalf("marshal strict Prompt run: %v", err)
	}
	if _, err = decodePromptApplicationVNextContract(promptApplicationRunV6Schema, payload); err != nil {
		t.Fatalf("History detail is not strict workflow_run_record.v6: %v payload=%s", err, payload)
	}
	for _, forbidden := range []string{"first private input", "second private input", first.Output, "rendered_messages", "provider_raw_response"} {
		if forbidden != "" && strings.Contains(string(payload), forbidden) {
			t.Fatalf("Prompt metadata projection leaked %q: %s", forbidden, payload)
		}
	}
}

func TestPromptApplicationRunAndSessionPersistAcrossSQLiteRestart(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "prompt-invocation.db")
	runtime := openApplicationInteractionSQLiteRuntime(t, databasePath)
	fixture := newPromptApplicationInvocationFixture(t)
	legacyRunStore := newSQLiteWorkflowRunStore(runtime.DB())
	promptRunStore, err := newPromptApplicationRunStoreForWorkflowRunStore(legacyRunStore)
	if err != nil {
		t.Fatalf("create SQLite Prompt run store: %v", err)
	}
	fixture.service.runStore = newCombinedWorkflowRunStore(legacyRunStore, promptRunStore)
	legacySessions := newSQLiteApplicationInteractionSessionRepository(runtime.DB())
	promptSessions, err := newPromptApplicationSessionRepositoryForLegacy(legacySessions)
	if err != nil {
		t.Fatalf("create SQLite Prompt session repository: %v", err)
	}
	sessionRepository := newCombinedApplicationInteractionSessionRepository(legacySessions, promptSessions)
	resolver := newExactApplicationInteractionAuthorityResolver(fixture.catalog, nil, nil, workflowRAGApplicationAuthorityResolver{})
	resolver.resolvePrompt = func(ctx PromptApplicationRuntimeContext) (PromptApplicationRuntimeAuthorityV2, string) {
		authority, failure := fixture.service.resolveAuthority(ctx)
		return authority.Snapshot, failure
	}
	sessions := newApplicationInteractionSessionService(sessionRepository, resolver)
	sessionContext := ApplicationInteractionContext{
		RequestContext: context.Background(), RequestID: fixture.ctx.RequestID, TenantRef: fixture.ctx.TenantRef,
		WorkspaceID: fixture.ctx.WorkspaceID, ApplicationID: fixture.ctx.ApplicationID,
		ActorRef: fixture.ctx.ActorRef, OwnerSubjectRef: fixture.ctx.OwnerSubjectRef,
		AuditRef: fixture.ctx.AuditRef, WriteEnabled: true,
	}
	created := sessions.Create(sessionContext, ApplicationInteractionSessionCreateInput{
		ProfileBinding: ApplicationInteractionProfileBinding{ExecutionProfile: applicationInteractionProfilePrompt},
	})
	if created.Session == nil {
		t.Fatalf("create SQLite Prompt session: %#v", created)
	}
	coordinator := newApplicationInteractionTurnCoordinator(sessions, resolver, nil, nil, fixture.service.Invoke)
	completed := coordinator.Execute(sessionContext, created.Session.SessionID, ApplicationInteractionTurnExecutionInput{
		ExpectedSessionVersion: created.Session.RecordVersion, ClientTurnKey: "sqlite-prompt-turn",
		PromptVariables: map[string]any{"question": "private sqlite input", "tone": "clear"},
	})
	if completed.FailureCode != "" || completed.Turn == nil || completed.Turn.RunRef == nil {
		t.Fatalf("complete SQLite Prompt turn: %#v", completed)
	}
	runID := completed.Turn.RunRef.RunID
	if err = runtime.Close(); err != nil {
		t.Fatalf("close SQLite Prompt runtime: %v", err)
	}

	restarted := openApplicationInteractionSQLiteRuntime(t, databasePath)
	defer restarted.Close()
	restartedLegacyRuns := newSQLiteWorkflowRunStore(restarted.DB())
	restartedPromptRuns, err := newPromptApplicationRunStoreForWorkflowRunStore(restartedLegacyRuns)
	if err != nil {
		t.Fatalf("recreate SQLite Prompt run store: %v", err)
	}
	restartedRuns := newCombinedWorkflowRunStore(restartedLegacyRuns, restartedPromptRuns)
	runContext := WorkflowRunContext{
		RequestContext: context.Background(), TenantRef: fixture.ctx.TenantRef,
		WorkspaceID: fixture.ctx.WorkspaceID, ApplicationID: fixture.ctx.ApplicationID,
	}
	restoredRun, found, err := restartedRuns.ReadRun(runContext, runID)
	if err != nil || !found || restoredRun.SchemaVersion != workflowRunRecordPromptSchemaVersion ||
		restoredRun.Status != WorkflowRunStatusSucceeded {
		t.Fatalf("restore SQLite Prompt run: found=%v err=%v run=%#v", found, err, restoredRun)
	}
	restartedLegacySessions := newSQLiteApplicationInteractionSessionRepository(restarted.DB())
	restartedPromptSessions, err := newPromptApplicationSessionRepositoryForLegacy(restartedLegacySessions)
	if err != nil {
		t.Fatalf("recreate SQLite Prompt session repository: %v", err)
	}
	restartedSessionService := newApplicationInteractionSessionService(
		newCombinedApplicationInteractionSessionRepository(restartedLegacySessions, restartedPromptSessions), resolver,
	)
	restoredSession := restartedSessionService.Read(sessionContext, created.Session.SessionID)
	restoredTurns, failure := restartedSessionService.ListTurns(sessionContext, created.Session.SessionID)
	if restoredSession.FailureCode != "" || restoredSession.Session == nil ||
		restoredSession.Session.SchemaVersion != promptApplicationSessionV2Schema ||
		failure != "" || len(restoredTurns) != 1 || restoredTurns[0].RunRef == nil ||
		restoredTurns[0].RunRef.RunID != runID {
		t.Fatalf("restore SQLite Prompt session: session=%#v turns=%#v failure=%s", restoredSession, restoredTurns, failure)
	}
	var payloads string
	if err = restarted.DB().QueryRowContext(context.Background(), `SELECT sanitized_session_payload || sanitized_turn_payload
FROM prompt_application_sessions JOIN prompt_application_session_turns USING (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id)
WHERE session_id=?`, created.Session.SessionID).Scan(&payloads); err != nil {
		t.Fatalf("read SQLite Prompt sanitized payloads: %v", err)
	}
	if strings.Contains(payloads, "private sqlite input") || strings.Contains(payloads, "rendered_messages") {
		t.Fatalf("SQLite Prompt persistence leaked transient content: %s", payloads)
	}
}

func TestPromptApplicationInvocationHTTPUsesDedicatedAPIKeyScopeAndStrictBody(t *testing.T) {
	fixture := newPromptApplicationInvocationFixture(t)
	apiKeys := newMemoryAPIKeyRepository()
	invokeToken := seedPromptApplicationInvocationAPIKey(t, apiKeys, fixture, "key_promptinvokeaaaa", []string{"prompt_application:invoke"})
	chatToken := seedPromptApplicationInvocationAPIKey(t, apiKeys, fixture, "key_promptchatonlyaa", []string{"chat:invoke"})
	server := &Server{
		config: config.Config{
			PromptApplicationRuntimeDevHTTPEnabled: true,
			ApplicationCatalogDevHTTPEnabled:       true,
			GatewayAuthMode:                        gatewayAPIKeyAuthenticationSource,
			Provider:                               "mock",
		},
		bridge:                                fixture.bridge,
		applicationCatalogRepository:          fixture.catalog,
		applicationDraftRepository:            fixture.service.authorityResolver.draftRepository,
		applicationPublishCandidateRepository: fixture.candidates,
		promptApplicationTemplateRepository:   fixture.service.authorityResolver.templateRepository,
		promptApplicationRuntimeRepository:    fixture.runtime,
		apiKeyRepository:                      apiKeys,
		workflowRunStore:                      fixture.runStore,
		applicationRunStore:                   fixture.runStore,
	}
	denied := httptest.NewRequest(http.MethodPost, promptApplicationInvocationRoute, strings.NewReader(
		`{"variables":{"question":"denied","tone":"clear"},"client_invocation_key":"http-denied"}`,
	))
	denied.Header.Set("Authorization", "Bearer "+chatToken)
	deniedRecorder := httptest.NewRecorder()
	server.handlePromptApplicationInvocation(deniedRecorder, denied)
	if deniedRecorder.Code != http.StatusForbidden || !strings.Contains(deniedRecorder.Body.String(), APIKeyFailureScopeDenied) ||
		fixture.bridge.callCount() != 0 {
		t.Fatalf("non-Prompt API key crossed invocation scope: status=%d body=%s calls=%d", deniedRecorder.Code, deniedRecorder.Body.String(), fixture.bridge.callCount())
	}
	unknown := httptest.NewRequest(http.MethodPost, promptApplicationInvocationRoute, strings.NewReader(
		`{"variables":{"question":"strict","tone":"clear"},"client_invocation_key":"http-unknown","model":"client-authority-forbidden"}`,
	))
	unknown.Header.Set("Authorization", "Bearer "+invokeToken)
	unknownRecorder := httptest.NewRecorder()
	server.handlePromptApplicationInvocation(unknownRecorder, unknown)
	if unknownRecorder.Code != http.StatusBadRequest || !strings.Contains(unknownRecorder.Body.String(), "INVALID_JSON") ||
		fixture.bridge.callCount() != 0 {
		t.Fatalf("Prompt invocation accepted client authority: status=%d body=%s calls=%d", unknownRecorder.Code, unknownRecorder.Body.String(), fixture.bridge.callCount())
	}
	success := httptest.NewRequest(http.MethodPost, promptApplicationInvocationRoute, strings.NewReader(
		`{"variables":{"question":"private HTTP input","tone":"clear"},"client_invocation_key":"http-success"}`,
	))
	success.Header.Set("Authorization", "Bearer "+invokeToken)
	successRecorder := httptest.NewRecorder()
	server.handlePromptApplicationInvocation(successRecorder, success)
	var envelope promptApplicationInvocationEnvelope
	if successRecorder.Code != http.StatusOK || json.Unmarshal(successRecorder.Body.Bytes(), &envelope) != nil ||
		envelope.FailureCode != nil || envelope.Run == nil || envelope.Run.SchemaVersion != workflowRunRecordPromptSchemaVersion ||
		envelope.Output == "" || envelope.Run.SideEffects.ProviderCalls != 1 || fixture.bridge.callCount() != 1 {
		t.Fatalf("Prompt API key invocation failed: status=%d body=%s calls=%d", successRecorder.Code, successRecorder.Body.String(), fixture.bridge.callCount())
	}
	if strings.Contains(successRecorder.Body.String(), "private HTTP input") ||
		strings.Contains(successRecorder.Body.String(), invokeToken) ||
		strings.Contains(successRecorder.Body.String(), "rendered_messages") {
		t.Fatalf("Prompt invocation response leaked protected material: %s", successRecorder.Body.String())
	}
}

func seedPromptApplicationInvocationAPIKey(
	t *testing.T,
	repository apiKeyRepository,
	fixture promptApplicationInvocationFixture,
	apiKeyID string,
	scopes []string,
) string {
	t.Helper()
	token, digest, err := newAPIKeyCredential(apiKeyID)
	if err != nil {
		t.Fatalf("create Prompt API key credential: %v", err)
	}
	now := time.Now().UTC()
	ctx := APIKeyContext{
		RequestContext: context.Background(), RequestID: "request_prompt_api_key",
		TenantRef: fixture.ctx.TenantRef, WorkspaceID: fixture.ctx.WorkspaceID,
		ActorRef: fixture.ctx.ActorRef, OwnerSubjectRef: fixture.ctx.OwnerSubjectRef,
		AuditRef: "audit_prompt_api_key", WriteEnabled: true,
	}
	record := APIKeyRecord{
		SchemaVersion: apiKeyRecordSchemaVersion, APIKeyID: apiKeyID,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: fixture.ctx.ApplicationID,
		OwnerSubjectRef: ctx.OwnerSubjectRef, DisplayName: "Prompt invocation key", Scopes: scopes,
		LifecycleState: apiKeyLifecycleActive, EffectiveState: apiKeyLifecycleActive, RecordVersion: 1,
		CreatedAt: now.Format(time.RFC3339Nano), ExpiresAt: now.Add(24 * time.Hour).Format(time.RFC3339Nano),
		CreatedByActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
		credentialDigest: digest,
	}
	if _, err = repository.Create(ctx, record); err != nil {
		t.Fatalf("seed Prompt API key: %v", err)
	}
	return token
}

func assertPromptApplicationStrictJSON(t *testing.T, value any, schemaVersion string) {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal %s: %v", schemaVersion, err)
	}
	if _, err = decodePromptApplicationVNextContract(schemaVersion, payload); err != nil {
		t.Fatalf("non-strict %s response: %v payload=%s", schemaVersion, err, payload)
	}
}

type authorityDriftPromptRunStore struct {
	workflowRunStore
	drift func()
}

func (store authorityDriftPromptRunStore) UpsertRun(ctx WorkflowRunContext, record *WorkflowRunRecord) error {
	err := store.workflowRunStore.UpsertRun(ctx, record)
	if err == nil && record.RecordVersion == 1 && store.drift != nil {
		store.drift()
	}
	return err
}

func newPromptApplicationInvocationFixture(t *testing.T) promptApplicationInvocationFixture {
	t.Helper()
	draftRepository := newMemoryApplicationConfigurationDraftRepository()
	candidateRepository := newMemoryApplicationPublishCandidateRepository()
	templateRepository := newMemoryPromptApplicationTemplateRepository()
	catalogRepository := newMemoryApplicationCatalogRepository()
	runtimeRepository := newMemoryPromptApplicationRuntimeRepository()
	runStore := newMemoryWorkflowRunStore(20)
	testBridge := &workflowExecutorTestBridge{}

	templateContext := validPromptApplicationTemplateContext()
	templateInput := validPromptApplicationTemplateDraftInput()
	templateService := newPromptApplicationTemplateService(templateRepository)
	if saved := templateService.SaveDraft(templateContext, templateInput, 0); saved.Draft == nil {
		t.Fatalf("seed Prompt template: %#v", saved)
	}
	if version := templateService.CreateVersion(templateContext, templateInput.TemplateID, 1); version.Version == nil {
		t.Fatalf("create Prompt template version: %#v", version)
	}

	draftContext := validApplicationDraftContext()
	draftContext.ApplicationID = templateContext.ApplicationID
	draftContext.ActorRef, draftContext.OwnerSubjectRef = templateContext.ActorRef, templateContext.OwnerSubjectRef
	draftContext.TemplateBindingEnabled = true
	payload := validApplicationDraftPayload()
	payload.DraftID, payload.ApplicationID, payload.ApplicationKind = "app-config-prompt-invocation", draftContext.ApplicationID, "prompt_application"
	payload.BaseApplicationUpdatedAt = "2026-07-22T08:00:00Z"
	draftService := newApplicationConfigurationDraftService(draftRepository)
	draftService.readPromptTemplateVersion = func(ApplicationConfigurationDraftContext, string, int) (PromptApplicationTemplateVersion, string) {
		value, err := templateRepository.ReadVersion(templateContext, templateInput.TemplateID, 1)
		if err != nil {
			return PromptApplicationTemplateVersion{}, PromptApplicationTemplateFailureVersionNotFound
		}
		return value, ""
	}
	if saved := draftService.Save(draftContext, payload, 0); saved.Draft == nil {
		t.Fatalf("seed Prompt configuration: %#v", saved)
	}
	if bound := draftService.BindPromptTemplate(draftContext, payload.DraftID, PromptApplicationTemplateBindingInput{
		ExpectedDraftVersion: 1, TemplateID: templateInput.TemplateID, TemplateVersion: 1,
	}); bound.Draft == nil {
		t.Fatalf("bind Prompt template: %#v", bound)
	}

	catalogContext := ApplicationCatalogContext{
		RequestContext: context.Background(), RequestID: "request_prompt_invocation_catalog",
		TenantRef: templateContext.TenantRef, WorkspaceID: templateContext.WorkspaceID,
		ActorRef: templateContext.ActorRef, OwnerSubjectRef: templateContext.OwnerSubjectRef,
		AuditRef: "audit_prompt_invocation_catalog", WriteEnabled: true,
	}
	catalog := ApplicationCatalogRecord{
		SchemaVersion: applicationCatalogSchemaVersion, ApplicationID: templateContext.ApplicationID,
		TenantRef: catalogContext.TenantRef, WorkspaceID: catalogContext.WorkspaceID,
		OwnerSubjectRef: catalogContext.OwnerSubjectRef, DisplayName: "Prompt invocation",
		Description: "Controlled Prompt invocation.", ApplicationKind: "prompt_application",
		LifecycleState: applicationCatalogLifecycleActive, RecordVersion: 1,
		CreatedAt: payload.BaseApplicationUpdatedAt, UpdatedAt: payload.BaseApplicationUpdatedAt,
		CreatedByActorRef: catalogContext.ActorRef, UpdatedByActorRef: catalogContext.ActorRef,
		RequestID: catalogContext.RequestID, AuditRef: catalogContext.AuditRef,
	}
	if _, err := catalogRepository.Create(catalogContext, catalog); err != nil {
		t.Fatalf("seed Prompt application catalog: %v", err)
	}

	publishContext := validApplicationPublishContext()
	publishContext.TenantRef, publishContext.WorkspaceID, publishContext.ApplicationID = templateContext.TenantRef, templateContext.WorkspaceID, templateContext.ApplicationID
	publishContext.ActorRef, publishContext.OwnerSubjectRef = templateContext.ActorRef, templateContext.OwnerSubjectRef
	publishContext.PromptTemplateSourceReadEnabled = true
	readBaseline := func(ctx ApplicationPublishContext) (ApplicationSummary, error) {
		return ApplicationSummary{
			ApplicationRef: ctx.ApplicationID, ApplicationKind: "prompt_application",
			UpdatedAt: payload.BaseApplicationUpdatedAt,
		}, nil
	}
	publishService := newApplicationPublishCandidateService(draftRepository, candidateRepository, readBaseline)
	publishService.readPromptTemplateVersion = func(ApplicationPublishContext, PromptApplicationTemplateRef) (PromptApplicationTemplateVersion, string) {
		value, err := templateRepository.ReadVersion(templateContext, templateInput.TemplateID, 1)
		if err != nil {
			return PromptApplicationTemplateVersion{}, PromptApplicationTemplateFailureVersionNotFound
		}
		return value, ""
	}
	created := publishService.Create(publishContext, ApplicationPublishCreateInput{
		CandidateID: "candidate-prompt-invocation", DraftID: payload.DraftID, ExpectedDraftVersion: 2,
	})
	if created.Candidate == nil {
		t.Fatalf("create Prompt candidate: %#v", created)
	}
	approved := publishService.Review(publishContext, created.Candidate.CandidateID, ApplicationPublishReviewInput{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: "Reviewed controlled invocation authority.",
	})
	if approved.Candidate == nil {
		t.Fatalf("approve Prompt candidate: %#v", approved)
	}

	runtimeContext := PromptApplicationRuntimeContext{
		RequestContext: context.Background(), RequestID: "request_prompt_invocation",
		TenantRef: templateContext.TenantRef, WorkspaceID: templateContext.WorkspaceID,
		ApplicationID: templateContext.ApplicationID, ActorRef: templateContext.ActorRef,
		OwnerSubjectRef: templateContext.OwnerSubjectRef, AuditRef: "audit_prompt_invocation", WriteEnabled: true,
	}
	resolver := promptApplicationRuntimeAuthorityResolver{
		publishRepository: candidateRepository, draftRepository: draftRepository,
		templateRepository: templateRepository, readApplication: readBaseline,
	}
	runtimeService := newPromptApplicationRuntimeService(runtimeRepository, resolver)
	runtimeService.now = func() time.Time { return time.Date(2026, 7, 22, 8, 5, 0, 0, time.UTC) }
	runtimeService.newID = func(prefix string) (string, error) {
		if prefix == "ptra_" {
			return "ptra_aaaaaaaaaaaaaaaa", nil
		}
		return "ptrae_aaaaaaaaaaaaaaaa", nil
	}
	activated := runtimeService.Decide(runtimeContext, PromptApplicationRuntimeDecisionInput{
		ExpectedAssignmentVersion: 0, Action: "activate", CandidateID: approved.Candidate.CandidateID,
	})
	if activated.Assignment == nil {
		t.Fatalf("activate Prompt assignment: %#v", activated)
	}

	service := newPromptApplicationInvocationService(
		runtimeRepository, resolver, catalogRepository, runStore, testBridge,
	)
	service.resolveSelection = func(context.Context, string) northboundSelection {
		return northboundSelection{
			provider: "mock", providerProfile: "default", model: "profile:local-dev",
			upstreamModel: "profile:local-dev", source: "test_selection",
		}
	}
	service.now = func() time.Time { return time.Date(2026, 7, 22, 8, 6, 0, 0, time.UTC) }
	return promptApplicationInvocationFixture{
		service: service, ctx: runtimeContext, runStore: runStore, runtime: runtimeRepository,
		candidates: candidateRepository, bridge: testBridge, catalog: catalogRepository, catalogCtx: catalogContext,
	}
}
