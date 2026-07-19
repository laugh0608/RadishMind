package httpapi

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	ApplicationInteractionFailureExecutionUnavailable = "application_session_execution_profile_unavailable"
	ApplicationInteractionFailureDelegatedRunContract = "application_session_delegated_run_contract_mismatch"
	ApplicationInteractionFailureRunOutcomeUnknown    = "application_session_run_outcome_unknown"
	ApplicationInteractionFailureTurnInterrupted      = "application_session_turn_interrupted"
	ApplicationInteractionFailureTurnCanceled         = "application_session_turn_canceled"
	applicationInteractionReconcilerActorRef          = "system:application_session_reconciler"
)

type ApplicationInteractionTurnExecutionInput struct {
	ExpectedSessionVersion int
	ClientTurnKey          string
	InputText              string
	ConditionValues        map[string]bool
	Model                  string
	Temperature            *float64
}

type ApplicationInteractionTurnExecutionResult struct {
	Session          *ApplicationInteractionSession
	Turn             *ApplicationInteractionTurn
	AdvisoryOutput   string
	Answer           *WorkflowRAGApplicationAnswer
	FailureCode      string
	FailureSummary   string
	IdempotentReplay bool
}

type ApplicationInteractionReconciliationResult struct {
	Reconciled  int
	FailureCode string
}

type applicationInteractionWorkflowDelegate func(WorkflowRunContext, WorkflowDefinitionRunRequest) WorkflowRunResult
type applicationInteractionRAGDelegate func(WorkflowRAGApplicationRuntimeContext, WorkflowRAGApplicationInvocationInput) WorkflowRAGApplicationInvocationResult

type applicationInteractionTurnCoordinator struct {
	sessions        applicationInteractionSessionService
	resolver        applicationInteractionAuthorityResolver
	executeWorkflow applicationInteractionWorkflowDelegate
	invokeRAG       applicationInteractionRAGDelegate
	now             func() time.Time
	staleAfter      time.Duration
}

func newApplicationInteractionTurnCoordinator(
	sessions applicationInteractionSessionService,
	resolver applicationInteractionAuthorityResolver,
	executeWorkflow applicationInteractionWorkflowDelegate,
	invokeRAG applicationInteractionRAGDelegate,
) applicationInteractionTurnCoordinator {
	return applicationInteractionTurnCoordinator{
		sessions: sessions, resolver: resolver, executeWorkflow: executeWorkflow, invokeRAG: invokeRAG,
		now: func() time.Time { return time.Now().UTC() }, staleAfter: workflowExecutorDefaultMaxRuntime,
	}
}

func (coordinator applicationInteractionTurnCoordinator) Execute(
	ctx ApplicationInteractionContext,
	sessionID string,
	input ApplicationInteractionTurnExecutionInput,
) ApplicationInteractionTurnExecutionResult {
	if validateApplicationInteractionContext(ctx) != nil || !applicationSessionIDPattern.MatchString(strings.TrimSpace(sessionID)) ||
		input.ExpectedSessionVersion < 1 || !applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(input.ClientTurnKey)) {
		return applicationInteractionTurnExecutionFailure(ApplicationInteractionFailurePayloadInvalid, "Application session turn input is invalid.")
	}
	if reconciled := coordinator.ReconcileStale(ctx, sessionID); reconciled.FailureCode != "" {
		return applicationInteractionTurnExecutionFailure(reconciled.FailureCode, "Application session stale turn reconciliation is unavailable.")
	}
	current := coordinator.sessions.Read(ctx, sessionID)
	if current.FailureCode != "" || current.Session == nil {
		return applicationInteractionTurnExecutionFromSessionResult(current)
	}
	normalized, failure, summary := normalizeApplicationInteractionTurnExecutionInput(*current.Session, input)
	if failure != "" {
		return applicationInteractionTurnExecutionFailure(failure, summary)
	}
	if (current.Session.ProfileBinding.ExecutionProfile == applicationInteractionProfileWorkflow && coordinator.executeWorkflow == nil) ||
		(current.Session.ProfileBinding.ExecutionProfile == applicationInteractionProfileRAG && coordinator.invokeRAG == nil) {
		return applicationInteractionTurnExecutionFailure(ApplicationInteractionFailureExecutionUnavailable, "The selected application session execution profile is unavailable.")
	}
	startedAt := coordinator.currentTime()
	reserved := coordinator.sessions.ReserveTurn(ctx, sessionID, ApplicationInteractionTurnReservationInput{
		ExpectedSessionVersion: normalized.ExpectedSessionVersion,
		ClientTurnKey:          normalized.ClientTurnKey,
		InputDigest:            workflowDefinitionInputDigest(normalized.InputText),
		InputBytes:             len([]byte(normalized.InputText)),
		StartedAt:              startedAt,
	})
	if reserved.FailureCode != "" || reserved.Turn == nil || reserved.Session == nil {
		return applicationInteractionTurnExecutionFromSessionResult(reserved)
	}
	if reserved.IdempotentReplay {
		result := applicationInteractionTurnExecutionFromSessionResult(reserved)
		if reserved.Turn.Status != string(WorkflowRunStatusSucceeded) && reserved.Turn.Status != string(WorkflowRunStatusRunning) {
			result.FailureCode, result.FailureSummary = reserved.Turn.FailureCode, reserved.Turn.FailureSummary
		}
		return result
	}
	currentAuthority, authorityFailure := coordinator.resolver.Resolve(ctx, reserved.Session.ProfileBinding)
	if authorityFailure != "" || currentAuthority.AuthorityDigest != reserved.Turn.Authority.AuthorityDigest {
		if authorityFailure == "" {
			authorityFailure = ApplicationInteractionFailureAuthorityChanged
		}
		return coordinator.completeFailure(ctx, *reserved.Turn, authorityFailure, "Application runtime authority changed before delegated execution.", nil)
	}
	if reserved.Turn.ExecutionProfile == applicationInteractionProfileWorkflow {
		return coordinator.executeWorkflowTurn(ctx, *reserved.Turn, normalized)
	}
	return coordinator.executeRAGTurn(ctx, *reserved.Turn, normalized)
}

func (coordinator applicationInteractionTurnCoordinator) executeWorkflowTurn(
	ctx ApplicationInteractionContext,
	turn ApplicationInteractionTurn,
	input ApplicationInteractionTurnExecutionInput,
) ApplicationInteractionTurnExecutionResult {
	authority := turn.Authority.WorkflowDefinition
	if authority == nil {
		return coordinator.completeFailure(ctx, turn, ApplicationInteractionFailureDelegatedRunContract, "Workflow definition authority is invalid.", nil)
	}
	runContext := WorkflowRunContext{RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef, AuditRef: ctx.AuditRef}
	result := coordinator.executeWorkflow(runContext, WorkflowDefinitionRunRequest{
		DefinitionID: authority.DefinitionID, ExpectedPointerVersion: authority.ActivationPointerVersion,
		ExpectedDefinitionVersion: authority.DefinitionVersion, ExpectedDefinitionDigest: authority.DefinitionDigest,
		InputText: input.InputText, ConditionValues: input.ConditionValues, Model: input.Model, Temperature: input.Temperature,
	})
	status, failureCode, failureSummary, runRef := applicationInteractionWorkflowTerminal(result)
	completed := coordinator.sessions.CompleteTurn(ctx, turn.SessionID, turn.TurnID, ApplicationInteractionTurnCompletionInput{
		Status: status, RunRef: runRef, FailureCode: failureCode, FailureSummary: failureSummary, CompletedAt: coordinator.currentTime(),
	})
	response := applicationInteractionTurnExecutionFromSessionResult(completed)
	if completed.FailureCode != "" {
		response.FailureSummary = "Application session turn terminal evidence could not be stored."
		return response
	}
	response.FailureCode, response.FailureSummary = failureCode, failureSummary
	if completed.Turn != nil && completed.Turn.Status == string(WorkflowRunStatusSucceeded) {
		response.AdvisoryOutput = result.AdvisoryOutput
		response.FailureCode, response.FailureSummary = "", ""
	}
	return response
}

func (coordinator applicationInteractionTurnCoordinator) executeRAGTurn(
	ctx ApplicationInteractionContext,
	turn ApplicationInteractionTurn,
	input ApplicationInteractionTurnExecutionInput,
) ApplicationInteractionTurnExecutionResult {
	runtimeContext := WorkflowRAGApplicationRuntimeContext{RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef}
	result := coordinator.invokeRAG(runtimeContext, WorkflowRAGApplicationInvocationInput{Input: input.InputText})
	status, failureCode, failureSummary, runRef := applicationInteractionRAGTerminal(ctx.RequestContext, result)
	completed := coordinator.sessions.CompleteTurn(ctx, turn.SessionID, turn.TurnID, ApplicationInteractionTurnCompletionInput{
		Status: status, RunRef: runRef, FailureCode: failureCode, FailureSummary: failureSummary, CompletedAt: coordinator.currentTime(),
	})
	response := applicationInteractionTurnExecutionFromSessionResult(completed)
	if completed.FailureCode != "" {
		response.FailureSummary = "Application session turn terminal evidence could not be stored."
		return response
	}
	response.FailureCode, response.FailureSummary = failureCode, failureSummary
	if completed.Turn != nil && completed.Turn.Status == string(WorkflowRunStatusSucceeded) {
		response.Answer = result.Answer
		response.FailureCode, response.FailureSummary = "", ""
	}
	return response
}

func (coordinator applicationInteractionTurnCoordinator) completeFailure(
	ctx ApplicationInteractionContext,
	turn ApplicationInteractionTurn,
	failureCode string,
	failureSummary string,
	runRef *ApplicationInteractionRunRef,
) ApplicationInteractionTurnExecutionResult {
	completed := coordinator.sessions.CompleteTurn(ctx, turn.SessionID, turn.TurnID, ApplicationInteractionTurnCompletionInput{
		Status: string(WorkflowRunStatusFailed), RunRef: runRef, FailureCode: failureCode,
		FailureSummary: failureSummary, CompletedAt: coordinator.currentTime(),
	})
	result := applicationInteractionTurnExecutionFromSessionResult(completed)
	if completed.FailureCode == "" {
		result.FailureCode, result.FailureSummary = failureCode, failureSummary
	}
	return result
}

func (coordinator applicationInteractionTurnCoordinator) ReconcileStale(ctx ApplicationInteractionContext, sessionID string) ApplicationInteractionReconciliationResult {
	if validateApplicationInteractionContext(ctx) != nil || !applicationSessionIDPattern.MatchString(strings.TrimSpace(sessionID)) {
		return ApplicationInteractionReconciliationResult{FailureCode: ApplicationInteractionFailurePayloadInvalid}
	}
	turns, err := coordinator.sessions.repository.ListTurns(ctx, strings.TrimSpace(sessionID))
	if err != nil {
		return ApplicationInteractionReconciliationResult{FailureCode: applicationInteractionRepositoryFailure(err).FailureCode}
	}
	result := ApplicationInteractionReconciliationResult{}
	now := coordinator.currentTime()
	staleAfter := coordinator.staleAfter
	if staleAfter <= 0 {
		staleAfter = workflowExecutorDefaultMaxRuntime
	}
	for _, current := range turns {
		startedAt := parseApplicationInteractionTimestamp(current.StartedAt)
		if current.Status != string(WorkflowRunStatusRunning) || startedAt == nil || now.Sub(*startedAt) <= staleAfter {
			continue
		}
		completedAt := now.Format(time.RFC3339Nano)
		terminal := current
		terminal.Status = string(WorkflowRunStatusOutcomeUnknown)
		terminal.FailureCode = ApplicationInteractionFailureTurnInterrupted
		terminal.FailureSummary = "Application session turn was interrupted and was not replayed."
		terminal.CompletedAt = &completedAt
		terminal.ActorRef = applicationInteractionReconcilerActorRef
		terminal.RequestID, terminal.AuditRef = ctx.RequestID, ctx.AuditRef
		_, _, _, err = coordinator.sessions.repository.CompleteTurn(ctx, terminal)
		if errors.Is(err, errApplicationSessionIdempotency) {
			latest, readErr := coordinator.sessions.repository.ReadTurn(ctx, current.SessionID, current.TurnID)
			if readErr == nil && latest.Status != string(WorkflowRunStatusRunning) {
				continue
			}
		}
		if err != nil {
			return ApplicationInteractionReconciliationResult{Reconciled: result.Reconciled, FailureCode: applicationInteractionRepositoryFailure(err).FailureCode}
		}
		result.Reconciled++
	}
	return result
}

func normalizeApplicationInteractionTurnExecutionInput(
	session ApplicationInteractionSession,
	input ApplicationInteractionTurnExecutionInput,
) (ApplicationInteractionTurnExecutionInput, string, string) {
	input.ClientTurnKey = strings.TrimSpace(input.ClientTurnKey)
	input.InputText = strings.TrimSpace(input.InputText)
	input.Model = strings.TrimSpace(input.Model)
	if input.InputText == "" || !utf8Safe(input.InputText) {
		return ApplicationInteractionTurnExecutionInput{}, ApplicationInteractionFailurePayloadInvalid, "Application session turn input is invalid."
	}
	switch session.ProfileBinding.ExecutionProfile {
	case applicationInteractionProfileWorkflow:
		authority := session.Authority.WorkflowDefinition
		if authority == nil {
			return ApplicationInteractionTurnExecutionInput{}, ApplicationInteractionFailureDelegatedRunContract, "Workflow definition authority is invalid."
		}
		normalized, code, summary := normalizeWorkflowDefinitionRunRequest(WorkflowDefinitionRunRequest{
			DefinitionID: authority.DefinitionID, ExpectedPointerVersion: authority.ActivationPointerVersion,
			ExpectedDefinitionVersion: authority.DefinitionVersion, ExpectedDefinitionDigest: authority.DefinitionDigest,
			InputText: input.InputText, ConditionValues: input.ConditionValues, Model: input.Model, Temperature: input.Temperature,
		})
		if code != "" {
			return ApplicationInteractionTurnExecutionInput{}, string(code), summary
		}
		input.InputText, input.ConditionValues, input.Model, input.Temperature = normalized.InputText, normalized.ConditionValues, normalized.Model, normalized.Temperature
	case applicationInteractionProfileRAG:
		if len([]byte(input.InputText)) > workflowRAGApplicationInvocationMaxBytes || len(input.ConditionValues) != 0 || input.Model != "" || input.Temperature != nil {
			return ApplicationInteractionTurnExecutionInput{}, ApplicationInteractionFailurePayloadInvalid, "Application RAG session turns do not accept workflow execution options."
		}
		if workflowRAGContainsForbiddenMaterial(input.InputText) || applicationDraftStringContainsSecret(input.InputText) {
			return ApplicationInteractionTurnExecutionInput{}, WorkflowRAGApplicationFailureSecretForbidden, workflowRAGApplicationFailureSummary(WorkflowRAGApplicationFailureSecretForbidden)
		}
	default:
		return ApplicationInteractionTurnExecutionInput{}, ApplicationInteractionFailureExecutionUnavailable, "The selected application session execution profile is unavailable."
	}
	return input, "", ""
}

func applicationInteractionWorkflowTerminal(result WorkflowRunResult) (string, string, string, *ApplicationInteractionRunRef) {
	ref, valid := applicationInteractionDelegatedRunRef(applicationInteractionProfileWorkflow, result.Record)
	if result.Record != nil && !valid {
		return string(WorkflowRunStatusFailed), ApplicationInteractionFailureDelegatedRunContract, "Delegated workflow run metadata is invalid.", nil
	}
	if result.Record != nil && (result.Record.Status == WorkflowRunStatusRunning || result.Record.Diagnostic != nil && result.Record.Diagnostic.TerminalWriteState == WorkflowRunTerminalWritePending) {
		return string(WorkflowRunStatusOutcomeUnknown), ApplicationInteractionFailureRunOutcomeUnknown, "Delegated workflow run terminal evidence is unavailable.", ref
	}
	if result.Record != nil && result.Record.Status == WorkflowRunStatusSucceeded && result.FailureCode == "" && strings.TrimSpace(result.AdvisoryOutput) != "" {
		return string(WorkflowRunStatusSucceeded), "", "", ref
	}
	if result.Record != nil && result.Record.Status == WorkflowRunStatusCanceled {
		return string(WorkflowRunStatusCanceled), string(result.FailureCode), result.FailureSummary, ref
	}
	failureCode := string(result.FailureCode)
	if failureCode == "" {
		failureCode = ApplicationInteractionFailureDelegatedRunContract
	}
	failureSummary := strings.TrimSpace(result.FailureSummary)
	if failureSummary == "" {
		failureSummary = "Delegated workflow execution failed."
	}
	return string(WorkflowRunStatusFailed), failureCode, failureSummary, ref
}

func applicationInteractionRAGTerminal(ctx context.Context, result WorkflowRAGApplicationInvocationResult) (string, string, string, *ApplicationInteractionRunRef) {
	ref, valid := applicationInteractionDelegatedRunRef(applicationInteractionProfileRAG, result.Run)
	if result.Run != nil && !valid {
		return string(WorkflowRunStatusFailed), ApplicationInteractionFailureDelegatedRunContract, "Delegated application RAG run metadata is invalid.", nil
	}
	if result.Run != nil && (result.Run.Status == WorkflowRunStatusRunning || result.Run.Diagnostic != nil && result.Run.Diagnostic.TerminalWriteState == WorkflowRunTerminalWritePending) {
		return string(WorkflowRunStatusOutcomeUnknown), ApplicationInteractionFailureRunOutcomeUnknown, "Delegated application RAG run terminal evidence is unavailable.", ref
	}
	if result.Run != nil && result.Run.Status == WorkflowRunStatusSucceeded && result.FailureCode == "" && result.Answer != nil {
		return string(WorkflowRunStatusSucceeded), "", "", ref
	}
	if (ctx != nil && ctx.Err() != nil) || (result.Run != nil && result.Run.Diagnostic != nil && result.Run.Diagnostic.GatewayFailureCategory == WorkflowRunGatewayFailureCanceled) {
		return string(WorkflowRunStatusCanceled), ApplicationInteractionFailureTurnCanceled, "Application session turn was canceled.", ref
	}
	failureCode := strings.TrimSpace(result.FailureCode)
	if failureCode == "" {
		failureCode = ApplicationInteractionFailureDelegatedRunContract
	}
	failureSummary := strings.TrimSpace(result.FailureSummary)
	if failureSummary == "" {
		failureSummary = "Delegated application RAG execution failed."
	}
	return string(WorkflowRunStatusFailed), failureCode, failureSummary, ref
}

func applicationInteractionDelegatedRunRef(profile string, record *WorkflowRunRecord) (*ApplicationInteractionRunRef, bool) {
	if record == nil {
		return nil, true
	}
	ref := &ApplicationInteractionRunRef{RunID: strings.TrimSpace(record.RunID), SchemaVersion: strings.TrimSpace(record.SchemaVersion)}
	return ref, validateApplicationInteractionRunRef(profile, ref) == nil
}

func applicationInteractionTurnExecutionFromSessionResult(result ApplicationInteractionSessionResult) ApplicationInteractionTurnExecutionResult {
	return ApplicationInteractionTurnExecutionResult{Session: result.Session, Turn: result.Turn, FailureCode: result.FailureCode, IdempotentReplay: result.IdempotentReplay}
}

func applicationInteractionTurnExecutionFailure(code, summary string) ApplicationInteractionTurnExecutionResult {
	return ApplicationInteractionTurnExecutionResult{FailureCode: strings.TrimSpace(code), FailureSummary: strings.TrimSpace(summary)}
}

func (coordinator applicationInteractionTurnCoordinator) currentTime() time.Time {
	if coordinator.now == nil {
		return time.Now().UTC()
	}
	return coordinator.now().UTC()
}
