package httpapi

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/bridge"
)

type applicationInteractionAuthorityResolverFunc func(ApplicationInteractionContext, ApplicationInteractionProfileBinding) (ApplicationInteractionAuthoritySnapshot, string)

type failingApplicationInteractionTurnCompletionRepository struct {
	applicationInteractionSessionRepository
}

func (repository failingApplicationInteractionTurnCompletionRepository) CompleteTurn(ApplicationInteractionContext, ApplicationInteractionTurn) (ApplicationInteractionSession, ApplicationInteractionTurn, bool, error) {
	return ApplicationInteractionSession{}, ApplicationInteractionTurn{}, false, errApplicationSessionStore
}

func (resolve applicationInteractionAuthorityResolverFunc) Resolve(ctx ApplicationInteractionContext, binding ApplicationInteractionProfileBinding) (ApplicationInteractionAuthoritySnapshot, string) {
	return resolve(ctx, binding)
}

func TestApplicationInteractionTurnCoordinatorDelegatesWorkflowV5OnceAndKeepsContentTransient(t *testing.T) {
	coordinator, ctx, session, bridgeClient, repository := workflowApplicationInteractionCoordinatorFixture(t)
	input := "private workflow session input needle"
	result := coordinator.Execute(ctx, session.SessionID, ApplicationInteractionTurnExecutionInput{ExpectedSessionVersion: session.RecordVersion, ClientTurnKey: "turn_workflow_001", InputText: input, ConditionValues: map[string]bool{}})
	if result.FailureCode != "" || result.Turn == nil || result.Turn.Status != string(WorkflowRunStatusSucceeded) || result.Turn.RunRef == nil || result.Turn.RunRef.SchemaVersion != workflowRunRecordDefinitionSchemaVersion || result.AdvisoryOutput == "" || result.Answer != nil || bridgeClient.callCount() != 1 {
		t.Fatalf("workflow turn execution failed: %#v bridge=%d", result, bridgeClient.callCount())
	}
	replay := coordinator.Execute(ctx, session.SessionID, ApplicationInteractionTurnExecutionInput{ExpectedSessionVersion: session.RecordVersion, ClientTurnKey: "turn_workflow_001", InputText: input, ConditionValues: map[string]bool{}})
	if replay.FailureCode != "" || !replay.IdempotentReplay || replay.Turn == nil || replay.Turn.TurnID != result.Turn.TurnID || replay.AdvisoryOutput != "" || bridgeClient.callCount() != 1 {
		t.Fatalf("workflow turn idempotency repeated execution: %#v bridge=%d", replay, bridgeClient.callCount())
	}
	turns, err := repository.ListTurns(ctx, session.SessionID)
	if err != nil {
		t.Fatalf("list stored turns: %v", err)
	}
	payload, err := json.Marshal(turns)
	if err != nil || strings.Contains(string(payload), input) || strings.Contains(string(payload), result.AdvisoryOutput) {
		t.Fatalf("stored workflow turn leaked transient content: payload=%s err=%v", payload, err)
	}
}

func TestApplicationInteractionTurnCoordinatorConcurrentClientKeyCallsDelegateOnce(t *testing.T) {
	coordinator, ctx, session, _, _ := workflowApplicationInteractionCoordinatorFixture(t)
	entered := make(chan struct{})
	release := make(chan struct{})
	delegateCalls := make(chan struct{}, 2)
	coordinator.executeWorkflow = func(WorkflowRunContext, WorkflowDefinitionRunRequest) WorkflowRunResult {
		delegateCalls <- struct{}{}
		close(entered)
		<-release
		return WorkflowRunResult{Record: &WorkflowRunRecord{SchemaVersion: workflowRunRecordDefinitionSchemaVersion, RunID: "run_appsessionconcurrent01", Status: WorkflowRunStatusSucceeded}, AdvisoryOutput: "transient concurrent answer"}
	}
	results := make(chan ApplicationInteractionTurnExecutionResult, 2)
	input := ApplicationInteractionTurnExecutionInput{ExpectedSessionVersion: 1, ClientTurnKey: "turn_concurrent_001", InputText: "same concurrent input", ConditionValues: map[string]bool{}}
	go func() { results <- coordinator.Execute(ctx, session.SessionID, input) }()
	<-entered
	go func() { results <- coordinator.Execute(ctx, session.SessionID, input) }()
	second := <-results
	if !second.IdempotentReplay || second.Turn == nil || second.Turn.Status != string(WorkflowRunStatusRunning) {
		t.Fatalf("concurrent client key did not return the running reservation: %#v", second)
	}
	close(release)
	first := <-results
	if first.FailureCode != "" || first.Turn == nil || first.Turn.Status != string(WorkflowRunStatusSucceeded) {
		t.Fatalf("concurrent client key owner did not complete: %#v", first)
	}
	if len(delegateCalls) != 1 {
		t.Fatalf("concurrent client key called delegate %d times", len(delegateCalls))
	}
}

func TestApplicationInteractionTurnCoordinatorDelegatesApplicationRAGV4Once(t *testing.T) {
	fixture := newWorkflowRAGApplicationRuntimeTestFixture(t)
	activated := fixture.runtimeService.Decide(fixture.runtimeContext, WorkflowRAGApplicationRuntimeDecisionInput{ExpectedRecordVersion: 0, Decision: workflowRAGApplicationRuntimeDecisionActivate, PublishCandidateID: fixture.publishCandidate.CandidateID, Reason: "人工激活用于应用会话回合"})
	if activated.FailureCode != "" {
		t.Fatalf("activate application RAG runtime: %#v", activated)
	}
	applications := newMemoryApplicationCatalogRepository()
	seedWorkflowRAGApplicationRuntimeCatalogRecord(t, applications, fixture)
	resolver := newExactApplicationInteractionAuthorityResolver(applications, nil, fixture.runtimeRepository, fixture.resolver)
	repository := newMemoryApplicationInteractionSessionRepository()
	sessions := newApplicationInteractionSessionService(repository, resolver)
	sessions.newID = applicationInteractionStableIDGenerator()
	sessions.now = func() time.Time { return fixture.now }
	ctx := ApplicationInteractionContext{RequestContext: context.Background(), RequestID: "request_application_session_rag", TenantRef: fixture.runtimeContext.TenantRef, WorkspaceID: fixture.runtimeContext.WorkspaceID, ApplicationID: fixture.runtimeContext.ApplicationID, ActorRef: fixture.runtimeContext.ActorRef, OwnerSubjectRef: fixture.runtimeContext.OwnerSubjectRef, AuditRef: "audit_application_session_rag", WriteEnabled: true}
	created := sessions.Create(ctx, ApplicationInteractionSessionCreateInput{ProfileBinding: ApplicationInteractionProfileBinding{ExecutionProfile: applicationInteractionProfileRAG}})
	if created.FailureCode != "" || created.Session == nil {
		t.Fatalf("create application RAG session: %#v", created)
	}
	invocation := newWorkflowRAGApplicationInvocationService(fixture.runtimeRepository, fixture.resolver, fixture.runStore, fixture.bridge)
	invocation.resolveSelection = func(context.Context, string) northboundSelection { return workflowRAGTestSelection() }
	invocation.envelopeOptions = func(northboundSelection, float64) bridge.EnvelopeOptions { return bridge.EnvelopeOptions{} }
	invocation.newRunID = func() (string, error) { return "run_appsessionrag0001", nil }
	invocation.now = func() time.Time { return fixture.now.Add(time.Second) }
	coordinator := newApplicationInteractionTurnCoordinator(sessions, resolver, nil, invocation.Invoke)
	coordinator.now = func() time.Time { return fixture.now.Add(time.Second) }
	input := "approved promotion authority guidance"
	result := coordinator.Execute(ctx, created.Session.SessionID, ApplicationInteractionTurnExecutionInput{ExpectedSessionVersion: 1, ClientTurnKey: "turn_rag_001", InputText: input})
	if result.FailureCode != "" || result.Turn == nil || result.Turn.Status != string(WorkflowRunStatusSucceeded) || result.Turn.RunRef == nil || result.Turn.RunRef.SchemaVersion != workflowRunRecordAppRAGSchemaVersion || result.Answer == nil || result.AdvisoryOutput != "" || fixture.bridge.callCount() != 1 {
		t.Fatalf("application RAG session turn failed: %#v bridge=%d", result, fixture.bridge.callCount())
	}
	payload, err := json.Marshal(result.Turn)
	if err != nil || strings.Contains(string(payload), input) || strings.Contains(string(payload), result.Answer.Answer) {
		t.Fatalf("stored RAG turn leaked transient content: payload=%s err=%v", payload, err)
	}
}

func TestApplicationInteractionTurnCoordinatorFailsAuthorityDriftBeforeDelegate(t *testing.T) {
	coordinator, ctx, session, bridgeClient, _ := workflowApplicationInteractionCoordinatorFixture(t)
	current, failure := coordinator.resolver.Resolve(ctx, session.ProfileBinding)
	if failure != "" {
		t.Fatalf("resolve baseline authority: %s", failure)
	}
	drifted := cloneApplicationInteractionAuthority(current)
	drifted.ApplicationRecordVersion++
	drifted.AuthorityDigest, _ = applicationInteractionAuthorityDigest(drifted)
	coordinator.resolver = applicationInteractionAuthorityResolverFunc(func(ApplicationInteractionContext, ApplicationInteractionProfileBinding) (ApplicationInteractionAuthoritySnapshot, string) {
		return drifted, ""
	})
	calls := 0
	coordinator.executeWorkflow = func(WorkflowRunContext, WorkflowDefinitionRunRequest) WorkflowRunResult {
		calls++
		return WorkflowRunResult{}
	}
	result := coordinator.Execute(ctx, session.SessionID, ApplicationInteractionTurnExecutionInput{ExpectedSessionVersion: 1, ClientTurnKey: "turn_drift_001", InputText: "bounded authority drift input", ConditionValues: map[string]bool{}})
	if result.FailureCode != ApplicationInteractionFailureAuthorityChanged || result.Turn == nil || result.Turn.Status != string(WorkflowRunStatusFailed) || calls != 0 || bridgeClient.callCount() != 0 {
		t.Fatalf("authority drift reached delegated provider: %#v delegates=%d bridge=%d", result, calls, bridgeClient.callCount())
	}
}

func TestApplicationInteractionTurnCoordinatorMapsCancellationAndPendingTerminalEvidence(t *testing.T) {
	coordinator, ctx, session, bridgeClient, _ := workflowApplicationInteractionCoordinatorFixture(t)
	canceledContext, cancel := context.WithCancel(context.Background())
	cancel()
	ctx.RequestContext = canceledContext
	canceled := coordinator.Execute(ctx, session.SessionID, ApplicationInteractionTurnExecutionInput{ExpectedSessionVersion: 1, ClientTurnKey: "turn_cancel_001", InputText: "cancel before provider", ConditionValues: map[string]bool{}})
	if canceled.FailureCode != string(WorkflowRunFailureCanceled) || canceled.Turn == nil || canceled.Turn.Status != string(WorkflowRunStatusCanceled) || bridgeClient.callCount() != 0 {
		t.Fatalf("canceled turn did not fail closed: %#v bridge=%d", canceled, bridgeClient.callCount())
	}

	coordinator, ctx, session, _, _ = workflowApplicationInteractionCoordinatorFixture(t)
	coordinator.executeWorkflow = func(WorkflowRunContext, WorkflowDefinitionRunRequest) WorkflowRunResult {
		return WorkflowRunResult{Record: &WorkflowRunRecord{SchemaVersion: workflowRunRecordDefinitionSchemaVersion, RunID: "run_appsessionpending01", Status: WorkflowRunStatusRunning, Diagnostic: &WorkflowRunDiagnostic{TerminalWriteState: WorkflowRunTerminalWritePending}}, FailureCode: WorkflowRunFailureStoreUnavailable, FailureSummary: "terminal write unavailable"}
	}
	pending := coordinator.Execute(ctx, session.SessionID, ApplicationInteractionTurnExecutionInput{ExpectedSessionVersion: 1, ClientTurnKey: "turn_pending_001", InputText: "pending terminal metadata", ConditionValues: map[string]bool{}})
	if pending.FailureCode != ApplicationInteractionFailureRunOutcomeUnknown || pending.Turn == nil || pending.Turn.Status != string(WorkflowRunStatusOutcomeUnknown) || pending.Turn.RunRef == nil {
		t.Fatalf("pending delegated terminal evidence was not marked outcome_unknown: %#v", pending)
	}
}

func TestApplicationInteractionTurnCoordinatorDoesNotReturnAnswerWhenSessionTerminalWriteFails(t *testing.T) {
	coordinator, ctx, session, bridgeClient, repository := workflowApplicationInteractionCoordinatorFixture(t)
	coordinator.sessions.repository = failingApplicationInteractionTurnCompletionRepository{applicationInteractionSessionRepository: repository}
	result := coordinator.Execute(ctx, session.SessionID, ApplicationInteractionTurnExecutionInput{ExpectedSessionVersion: 1, ClientTurnKey: "turn_store_failure_001", InputText: "transient answer must be withheld", ConditionValues: map[string]bool{}})
	if result.FailureCode != ApplicationInteractionFailureStoreUnavailable || result.AdvisoryOutput != "" || result.Answer != nil || bridgeClient.callCount() != 1 {
		t.Fatalf("terminal session write failure returned transient answer or repeated provider: %#v bridge=%d", result, bridgeClient.callCount())
	}
	turns, err := repository.ListTurns(ctx, session.SessionID)
	if err != nil || len(turns) != 1 || turns[0].Status != string(WorkflowRunStatusRunning) {
		t.Fatalf("failed terminal write did not leave honest running reservation: turns=%#v err=%v", turns, err)
	}
}

func TestApplicationInteractionTurnCoordinatorReconcilesStaleWithoutDelegate(t *testing.T) {
	coordinator, ctx, session, bridgeClient, repository := workflowApplicationInteractionCoordinatorFixture(t)
	startedAt := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	reserved := coordinator.sessions.ReserveTurn(ctx, session.SessionID, ApplicationInteractionTurnReservationInput{ExpectedSessionVersion: 1, ClientTurnKey: "turn_stale_001", InputDigest: workflowDefinitionInputDigest("stale input"), InputBytes: len("stale input"), StartedAt: startedAt})
	if reserved.FailureCode != "" || reserved.Turn == nil {
		t.Fatalf("reserve stale turn fixture: %#v", reserved)
	}
	coordinator.now = func() time.Time { return startedAt.Add(workflowExecutorDefaultMaxRuntime + time.Second) }
	reconciled := coordinator.ReconcileStale(ctx, session.SessionID)
	stored, err := repository.ReadTurn(ctx, session.SessionID, reserved.Turn.TurnID)
	if reconciled.FailureCode != "" || reconciled.Reconciled != 1 || err != nil || stored.Status != string(WorkflowRunStatusOutcomeUnknown) || stored.FailureCode != ApplicationInteractionFailureTurnInterrupted || stored.RunRef != nil || bridgeClient.callCount() != 0 {
		t.Fatalf("stale turn reconciliation drifted: result=%#v turn=%#v err=%v bridge=%d", reconciled, stored, err, bridgeClient.callCount())
	}
	repeated := coordinator.ReconcileStale(ctx, session.SessionID)
	if repeated.FailureCode != "" || repeated.Reconciled != 0 || bridgeClient.callCount() != 0 {
		t.Fatalf("stale reconciliation was not idempotent: %#v bridge=%d", repeated, bridgeClient.callCount())
	}
}

func workflowApplicationInteractionCoordinatorFixture(t *testing.T) (applicationInteractionTurnCoordinator, ApplicationInteractionContext, ApplicationInteractionSession, *workflowExecutorTestBridge, *memoryApplicationInteractionSessionRepository) {
	t.Helper()
	definitionService, runContext, runRequest, bridgeClient, _ := workflowDefinitionExecutionFixture(t)
	resolver := newExactApplicationInteractionAuthorityResolver(definitionService.applications, definitionService.repository, nil, workflowRAGApplicationAuthorityResolver{})
	repository := newMemoryApplicationInteractionSessionRepository()
	sessions := newApplicationInteractionSessionService(repository, resolver)
	sessions.newID = applicationInteractionStableIDGenerator()
	baseTime := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	sessions.now = func() time.Time { return baseTime }
	ctx := ApplicationInteractionContext{RequestContext: context.Background(), RequestID: "request_application_session_workflow", TenantRef: runContext.TenantRef, WorkspaceID: runContext.WorkspaceID, ApplicationID: runContext.ApplicationID, ActorRef: runContext.ActorRef, OwnerSubjectRef: runContext.ActorRef, AuditRef: "audit_application_session_workflow", WriteEnabled: true}
	created := sessions.Create(ctx, ApplicationInteractionSessionCreateInput{ProfileBinding: ApplicationInteractionProfileBinding{ExecutionProfile: applicationInteractionProfileWorkflow, DefinitionID: runRequest.DefinitionID}})
	if created.FailureCode != "" || created.Session == nil {
		t.Fatalf("create workflow application session: %#v", created)
	}
	coordinator := newApplicationInteractionTurnCoordinator(sessions, resolver, definitionService.StartRun, nil)
	coordinator.now = func() time.Time { return baseTime.Add(time.Second) }
	return coordinator, ctx, *created.Session, bridgeClient, repository
}

func applicationInteractionStableIDGenerator() func(string) (string, error) {
	return func(prefix string) (string, error) {
		if prefix == "appsess" {
			return "appsess_aaaaaaaaaaaaaaaa", nil
		}
		return "appturn_aaaaaaaaaaaaaaaa", nil
	}
}
