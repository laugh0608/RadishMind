package httpapi

import (
	"context"
	"encoding/json"
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
	SaveResult             bool
	InputText              string
	Inputs                 map[string]any
	ConditionValues        map[string]bool
	Model                  string
	Temperature            *float64
	PromptVariables        map[string]any
	AgentTask              string
	AgentLocale            string
	AgentConversationID    string
	AgentArtifacts         []AgentCopilotArtifact
	AgentContext           map[string]any
	inputTextProvided      bool
	inputsProvided         bool
	structuredInput        *workflowStructuredInputNormalization
}

type applicationInteractionTurnInputMetadata struct {
	InputDigest         string
	InputBytes          int
	InputContractID     string
	InputContractDigest string
	InputFields         []WorkflowStructuredInputMetadataField
}

type ApplicationInteractionTurnExecutionResult struct {
	Session                   *ApplicationInteractionSession
	Turn                      *ApplicationInteractionTurn
	ResultArtifact            *ApplicationResultArtifactSummary
	ResultArtifactFailureCode string
	AdvisoryOutput            string
	Answer                    *WorkflowRAGApplicationAnswer
	PromptOutput              string
	AgentResponse             *AgentCopilotResponse
	FailureCode               string
	FailureSummary            string
	IdempotentReplay          bool
}

type ApplicationInteractionReconciliationResult struct {
	Reconciled  int
	FailureCode string
}

type applicationInteractionWorkflowDelegate func(WorkflowRunContext, WorkflowDefinitionRunRequest) WorkflowRunResult
type applicationInteractionRAGDelegate func(WorkflowRAGApplicationRuntimeContext, WorkflowRAGApplicationInvocationInput) WorkflowRAGApplicationInvocationResult
type applicationInteractionPromptDelegate func(PromptApplicationRuntimeContext, PromptApplicationInvocationInput) PromptApplicationInvocationResult
type applicationInteractionAgentCopilotDelegate func(AgentCopilotRuntimeContext, AgentCopilotInvocationInput) AgentCopilotInvocationResult

type applicationInteractionTurnCoordinator struct {
	sessions        applicationInteractionSessionService
	resultArtifacts *applicationResultArtifactService
	resolver        applicationInteractionAuthorityResolver
	executeWorkflow applicationInteractionWorkflowDelegate
	invokeRAG       applicationInteractionRAGDelegate
	invokePrompt    applicationInteractionPromptDelegate
	invokeAgent     applicationInteractionAgentCopilotDelegate
	now             func() time.Time
	staleAfter      time.Duration
}

func newApplicationInteractionTurnCoordinator(
	sessions applicationInteractionSessionService,
	resolver applicationInteractionAuthorityResolver,
	executeWorkflow applicationInteractionWorkflowDelegate,
	invokeRAG applicationInteractionRAGDelegate,
	invokePrompt ...applicationInteractionPromptDelegate,
) applicationInteractionTurnCoordinator {
	var promptDelegate applicationInteractionPromptDelegate
	if len(invokePrompt) != 0 {
		promptDelegate = invokePrompt[0]
	}
	return applicationInteractionTurnCoordinator{
		sessions: sessions, resolver: resolver, executeWorkflow: executeWorkflow, invokeRAG: invokeRAG, invokePrompt: promptDelegate,
		now: func() time.Time { return time.Now().UTC() }, staleAfter: workflowExecutorDefaultMaxRuntime,
	}
}

func (coordinator applicationInteractionTurnCoordinator) withAgentCopilot(delegate applicationInteractionAgentCopilotDelegate) applicationInteractionTurnCoordinator {
	coordinator.invokeAgent = delegate
	return coordinator
}

func (coordinator applicationInteractionTurnCoordinator) withResultArtifacts(service applicationResultArtifactService) applicationInteractionTurnCoordinator {
	coordinator.resultArtifacts = &service
	return coordinator
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
	if ((current.Session.ProfileBinding.ExecutionProfile == applicationInteractionProfileWorkflow || current.Session.ProfileBinding.ExecutionProfile == applicationInteractionProfileWorkflowStructured) && coordinator.executeWorkflow == nil) ||
		(current.Session.ProfileBinding.ExecutionProfile == applicationInteractionProfileRAG && coordinator.invokeRAG == nil) ||
		(current.Session.ProfileBinding.ExecutionProfile == applicationInteractionProfilePrompt && coordinator.invokePrompt == nil) ||
		(current.Session.ProfileBinding.ExecutionProfile == applicationInteractionProfileAgentCopilot && coordinator.invokeAgent == nil) {
		return applicationInteractionTurnExecutionFailure(ApplicationInteractionFailureExecutionUnavailable, "The selected application session execution profile is unavailable.")
	}
	startedAt := coordinator.currentTime()
	inputMetadata, digestErr := applicationInteractionInputMetadata(*current.Session, normalized)
	if digestErr != nil {
		return applicationInteractionTurnExecutionFailure(ApplicationInteractionFailurePayloadInvalid, "Application session turn input is invalid.")
	}
	reserved := coordinator.sessions.ReserveTurn(ctx, sessionID, ApplicationInteractionTurnReservationInput{
		ExpectedSessionVersion: normalized.ExpectedSessionVersion,
		ClientTurnKey:          normalized.ClientTurnKey,
		InputDigest:            inputMetadata.InputDigest,
		InputBytes:             inputMetadata.InputBytes,
		InputContractID:        inputMetadata.InputContractID,
		InputContractDigest:    inputMetadata.InputContractDigest,
		InputFields:            inputMetadata.InputFields,
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
		if normalized.SaveResult && reserved.Turn.Status == string(WorkflowRunStatusSucceeded) {
			result = coordinator.attachExistingResultArtifact(ctx, result)
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
	if reserved.Turn.ExecutionProfile == applicationInteractionProfileWorkflow || reserved.Turn.ExecutionProfile == applicationInteractionProfileWorkflowStructured {
		return coordinator.executeWorkflowTurn(ctx, *reserved.Turn, normalized)
	}
	if reserved.Turn.ExecutionProfile == applicationInteractionProfilePrompt {
		return coordinator.executePromptTurn(ctx, *reserved.Turn, normalized)
	}
	if reserved.Turn.ExecutionProfile == applicationInteractionProfileAgentCopilot {
		return coordinator.executeAgentCopilotTurn(ctx, *reserved.Turn, normalized)
	}
	return coordinator.executeRAGTurn(ctx, *reserved.Turn, normalized)
}

func (coordinator applicationInteractionTurnCoordinator) executeAgentCopilotTurn(
	ctx ApplicationInteractionContext,
	turn ApplicationInteractionTurn,
	input ApplicationInteractionTurnExecutionInput,
) ApplicationInteractionTurnExecutionResult {
	result := coordinator.invokeAgent(AgentCopilotRuntimeContext{
		RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef,
		WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef,
		OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef,
	}, AgentCopilotInvocationInput{
		Task: input.AgentTask, Locale: input.AgentLocale, ConversationID: input.AgentConversationID,
		Artifacts: input.AgentArtifacts, Context: input.AgentContext,
		ClientInvocationKey: turn.SessionID + ":" + turn.ClientTurnKey,
	})
	status, failureCode, failureSummary, runRef := applicationInteractionAgentCopilotTerminal(ctx.RequestContext, result)
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
		response.AgentResponse, response.FailureCode, response.FailureSummary = result.Response, "", ""
	}
	if input.SaveResult && response.AgentResponse != nil {
		content, err := json.Marshal(response.AgentResponse)
		if err != nil {
			response.ResultArtifactFailureCode = ApplicationResultArtifactFailurePayloadInvalid
			return response
		}
		return coordinator.captureResultArtifact(ctx, response, "application/json", string(content))
	}
	return response
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
		InputText: input.InputText, Inputs: cloneWorkflowStructuredInputValues(input.Inputs), ConditionValues: input.ConditionValues, Model: input.Model, Temperature: input.Temperature,
		inputTextProvided: input.inputTextProvided, inputsProvided: input.inputsProvided,
	})
	status, failureCode, failureSummary, runRef := applicationInteractionWorkflowTerminal(turn.ExecutionProfile, result)
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
	if input.SaveResult {
		return coordinator.captureResultArtifact(ctx, response, "text/markdown", response.AdvisoryOutput)
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
	if input.SaveResult && response.Answer != nil {
		content, err := json.Marshal(response.Answer)
		if err != nil {
			response.ResultArtifactFailureCode = ApplicationResultArtifactFailurePayloadInvalid
			return response
		}
		return coordinator.captureResultArtifact(ctx, response, "application/json", string(content))
	}
	return response
}

func (coordinator applicationInteractionTurnCoordinator) executePromptTurn(
	ctx ApplicationInteractionContext,
	turn ApplicationInteractionTurn,
	input ApplicationInteractionTurnExecutionInput,
) ApplicationInteractionTurnExecutionResult {
	runtimeContext := PromptApplicationRuntimeContext{
		RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef,
		WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef,
		OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef,
	}
	result := coordinator.invokePrompt(runtimeContext, PromptApplicationInvocationInput{
		Variables: input.PromptVariables, ClientInvocationKey: turn.SessionID + ":" + turn.ClientTurnKey,
	})
	status, failureCode, failureSummary, runRef := applicationInteractionPromptTerminal(ctx.RequestContext, result)
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
		response.PromptOutput = result.Output
		response.FailureCode, response.FailureSummary = "", ""
	}
	if input.SaveResult {
		return coordinator.captureResultArtifact(ctx, response, "text/markdown", response.PromptOutput)
	}
	return response
}

func (coordinator applicationInteractionTurnCoordinator) captureResultArtifact(
	ctx ApplicationInteractionContext,
	response ApplicationInteractionTurnExecutionResult,
	contentType string,
	content string,
) ApplicationInteractionTurnExecutionResult {
	if response.Turn == nil || response.Turn.Status != string(WorkflowRunStatusSucceeded) || response.Turn.RunRef == nil || strings.TrimSpace(content) == "" {
		response.ResultArtifactFailureCode = ApplicationResultArtifactFailureSourceUnavailable
		return response
	}
	if coordinator.resultArtifacts == nil {
		response.ResultArtifactFailureCode = ApplicationResultArtifactFailureStoreUnavailable
		return response
	}
	result := coordinator.resultArtifacts.Capture(ctx, ApplicationResultArtifactCaptureInput{
		Turn: *response.Turn, ContentType: contentType, Content: content,
	})
	response.ResultArtifact = result.Summary
	response.ResultArtifactFailureCode = result.FailureCode
	return response
}

func (coordinator applicationInteractionTurnCoordinator) attachExistingResultArtifact(
	ctx ApplicationInteractionContext,
	response ApplicationInteractionTurnExecutionResult,
) ApplicationInteractionTurnExecutionResult {
	if response.Turn == nil || coordinator.resultArtifacts == nil {
		response.ResultArtifactFailureCode = ApplicationResultArtifactFailureSourceUnavailable
		return response
	}
	result := coordinator.resultArtifacts.ReadByTurn(ctx, response.Turn.SessionID, response.Turn.TurnID)
	if result.FailureCode == ApplicationResultArtifactFailureNotFound {
		response.ResultArtifactFailureCode = ApplicationResultArtifactFailureSourceUnavailable
		return response
	}
	response.ResultArtifact = result.Summary
	response.ResultArtifactFailureCode = result.FailureCode
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
	switch session.ProfileBinding.ExecutionProfile {
	case applicationInteractionProfileWorkflow:
		if !input.inputTextProvided {
			input.inputTextProvided = input.InputText != ""
		}
		if input.InputText == "" || !utf8Safe(input.InputText) || input.PromptVariables != nil || input.inputsProvided || input.Inputs != nil {
			return ApplicationInteractionTurnExecutionInput{}, ApplicationInteractionFailurePayloadInvalid, "Application session turn input is invalid."
		}
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
		input.inputTextProvided = true
	case applicationInteractionProfileWorkflowStructured:
		if !input.inputsProvided {
			input.inputsProvided = input.Inputs != nil
		}
		authority := session.Authority.WorkflowDefinition
		if authority == nil || authority.InputContract == nil || input.inputTextProvided || input.InputText != "" || !input.inputsProvided || input.Inputs == nil || input.PromptVariables != nil {
			return ApplicationInteractionTurnExecutionInput{}, ApplicationInteractionFailurePayloadInvalid, "Structured workflow session turn input is invalid."
		}
		normalized, code, summary := normalizeWorkflowStructuredInputValues(savedWorkflowDraftContractFromDefinition(*authority.InputContract), input.Inputs)
		if code != "" {
			return ApplicationInteractionTurnExecutionInput{}, string(code), summary
		}
		request, code, summary := normalizeWorkflowDefinitionRunRequest(WorkflowDefinitionRunRequest{
			DefinitionID: authority.DefinitionID, ExpectedPointerVersion: authority.ActivationPointerVersion,
			ExpectedDefinitionVersion: authority.DefinitionVersion, ExpectedDefinitionDigest: authority.DefinitionDigest,
			Inputs: cloneWorkflowStructuredInputValues(input.Inputs), ConditionValues: input.ConditionValues, Model: input.Model, Temperature: input.Temperature,
			inputsProvided: true,
		})
		if code != "" {
			return ApplicationInteractionTurnExecutionInput{}, string(code), summary
		}
		input.Inputs, input.ConditionValues, input.Model, input.Temperature = request.Inputs, request.ConditionValues, request.Model, request.Temperature
		input.inputsProvided, input.structuredInput = true, &normalized
	case applicationInteractionProfileRAG:
		if input.InputText == "" || !utf8Safe(input.InputText) || input.PromptVariables != nil || input.Inputs != nil || input.inputsProvided ||
			len([]byte(input.InputText)) > workflowRAGApplicationInvocationMaxBytes || len(input.ConditionValues) != 0 || input.Model != "" || input.Temperature != nil {
			return ApplicationInteractionTurnExecutionInput{}, ApplicationInteractionFailurePayloadInvalid, "Application RAG session turns do not accept workflow execution options."
		}
		if workflowRAGContainsForbiddenMaterial(input.InputText) || applicationDraftStringContainsSecret(input.InputText) {
			return ApplicationInteractionTurnExecutionInput{}, WorkflowRAGApplicationFailureSecretForbidden, workflowRAGApplicationFailureSummary(WorkflowRAGApplicationFailureSecretForbidden)
		}
	case applicationInteractionProfilePrompt:
		if input.InputText != "" || input.Inputs != nil || input.inputsProvided || input.PromptVariables == nil || len(input.ConditionValues) != 0 || input.Model != "" || input.Temperature != nil {
			return ApplicationInteractionTurnExecutionInput{}, ApplicationInteractionFailurePayloadInvalid, "Prompt application session turns only accept variables."
		}
		if _, _, err := canonicalPromptApplicationInvocationInput(input.PromptVariables); err != nil {
			return ApplicationInteractionTurnExecutionInput{}, ApplicationInteractionFailurePayloadInvalid, "Prompt application session variables are invalid."
		}
	case applicationInteractionProfileAgentCopilot:
		if input.InputText != "" || input.Inputs != nil || input.inputsProvided || input.PromptVariables != nil || len(input.ConditionValues) != 0 ||
			input.Model != "" || input.Temperature != nil || strings.TrimSpace(input.AgentTask) == "" ||
			strings.TrimSpace(input.AgentLocale) == "" || input.AgentContext == nil {
			return ApplicationInteractionTurnExecutionInput{}, ApplicationInteractionFailurePayloadInvalid, "Agent Copilot session turns only accept canonical request data."
		}
		input.AgentTask, input.AgentLocale = strings.TrimSpace(input.AgentTask), strings.TrimSpace(input.AgentLocale)
		input.AgentConversationID = strings.TrimSpace(input.AgentConversationID)
	default:
		return ApplicationInteractionTurnExecutionInput{}, ApplicationInteractionFailureExecutionUnavailable, "The selected application session execution profile is unavailable."
	}
	return input, "", ""
}

func applicationInteractionInputMetadata(session ApplicationInteractionSession, input ApplicationInteractionTurnExecutionInput) (applicationInteractionTurnInputMetadata, error) {
	if session.ProfileBinding.ExecutionProfile == applicationInteractionProfileWorkflowStructured {
		if input.structuredInput == nil {
			return applicationInteractionTurnInputMetadata{}, errApplicationSessionContract
		}
		return applicationInteractionTurnInputMetadata{
			InputDigest: input.structuredInput.InputDigest, InputBytes: input.structuredInput.InputBytes,
			InputContractID: input.structuredInput.Contract.ContractID, InputContractDigest: input.structuredInput.Contract.ContractDigest,
			InputFields: cloneWorkflowStructuredInputMetadataFields(input.structuredInput.Fields),
		}, nil
	}
	if session.ProfileBinding.ExecutionProfile != applicationInteractionProfilePrompt {
		if session.ProfileBinding.ExecutionProfile == applicationInteractionProfileAgentCopilot {
			payload, err := json.Marshal(struct {
				Task           string                 `json:"task"`
				Locale         string                 `json:"locale"`
				ConversationID string                 `json:"conversation_id,omitempty"`
				Artifacts      []AgentCopilotArtifact `json:"artifacts"`
				Context        map[string]any         `json:"context"`
			}{
				Task: input.AgentTask, Locale: input.AgentLocale, ConversationID: input.AgentConversationID,
				Artifacts: input.AgentArtifacts, Context: input.AgentContext,
			})
			if err != nil || len(payload) > agentCopilotMaximumInvocationBytes {
				return applicationInteractionTurnInputMetadata{}, errApplicationSessionContract
			}
			return applicationInteractionTurnInputMetadata{InputDigest: workflowRAGSHA256(string(payload)), InputBytes: len(payload)}, nil
		}
		return applicationInteractionTurnInputMetadata{InputDigest: workflowDefinitionInputDigest(input.InputText), InputBytes: len([]byte(input.InputText))}, nil
	}
	payload, _, err := canonicalPromptApplicationInvocationInput(input.PromptVariables)
	if err != nil {
		return applicationInteractionTurnInputMetadata{}, err
	}
	return applicationInteractionTurnInputMetadata{InputDigest: workflowRAGSHA256(string(payload)), InputBytes: len(payload)}, nil
}

func applicationInteractionAgentCopilotTerminal(ctx context.Context, result AgentCopilotInvocationResult) (string, string, string, *ApplicationInteractionRunRef) {
	ref, valid := applicationInteractionDelegatedRunRef(applicationInteractionProfileAgentCopilot, result.Run)
	if result.Run != nil && !valid {
		return string(WorkflowRunStatusFailed), ApplicationInteractionFailureDelegatedRunContract, "Delegated Agent Copilot run metadata is invalid.", nil
	}
	if result.Run != nil && result.Run.Status == WorkflowRunStatusRunning {
		return string(WorkflowRunStatusOutcomeUnknown), ApplicationInteractionFailureRunOutcomeUnknown, "Delegated Agent Copilot run terminal evidence is unavailable.", ref
	}
	if result.Run != nil && result.Run.Status == WorkflowRunStatusSucceeded && result.FailureCode == "" && result.Response != nil {
		return string(WorkflowRunStatusSucceeded), "", "", ref
	}
	if (ctx != nil && ctx.Err() != nil) || result.Run != nil && result.Run.Status == WorkflowRunStatusCanceled {
		return string(WorkflowRunStatusCanceled), ApplicationInteractionFailureTurnCanceled, "Application session turn was canceled.", ref
	}
	failureCode := strings.TrimSpace(result.FailureCode)
	if failureCode == "" {
		failureCode = ApplicationInteractionFailureDelegatedRunContract
	}
	failureSummary := strings.TrimSpace(result.FailureSummary)
	if failureSummary == "" {
		failureSummary = "Delegated Agent Copilot execution failed."
	}
	status := string(WorkflowRunStatusFailed)
	if result.Run != nil && result.Run.Status == WorkflowRunStatusOutcomeUnknown {
		status = string(WorkflowRunStatusOutcomeUnknown)
	}
	return status, failureCode, failureSummary, ref
}

func applicationInteractionWorkflowTerminal(executionProfile string, result WorkflowRunResult) (string, string, string, *ApplicationInteractionRunRef) {
	ref, valid := applicationInteractionDelegatedRunRef(executionProfile, result.Record)
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

func applicationInteractionPromptTerminal(ctx context.Context, result PromptApplicationInvocationResult) (string, string, string, *ApplicationInteractionRunRef) {
	ref, valid := applicationInteractionDelegatedRunRef(applicationInteractionProfilePrompt, result.Run)
	if result.Run != nil && !valid {
		return string(WorkflowRunStatusFailed), ApplicationInteractionFailureDelegatedRunContract, "Delegated Prompt application run metadata is invalid.", nil
	}
	if result.Run != nil && result.Run.Status == WorkflowRunStatusRunning {
		return string(WorkflowRunStatusOutcomeUnknown), ApplicationInteractionFailureRunOutcomeUnknown, "Delegated Prompt application run terminal evidence is unavailable.", ref
	}
	if result.Run != nil && result.Run.Status == WorkflowRunStatusSucceeded && result.FailureCode == "" && result.Output != "" {
		return string(WorkflowRunStatusSucceeded), "", "", ref
	}
	if (ctx != nil && ctx.Err() != nil) || result.Run != nil && result.Run.Status == WorkflowRunStatusCanceled {
		return string(WorkflowRunStatusCanceled), ApplicationInteractionFailureTurnCanceled, "Application session turn was canceled.", ref
	}
	failureCode := strings.TrimSpace(result.FailureCode)
	if failureCode == "" {
		failureCode = ApplicationInteractionFailureDelegatedRunContract
	}
	failureSummary := strings.TrimSpace(result.FailureSummary)
	if failureSummary == "" {
		failureSummary = "Delegated Prompt application execution failed."
	}
	status := string(WorkflowRunStatusFailed)
	if result.Run != nil && result.Run.Status == WorkflowRunStatusOutcomeUnknown {
		status = string(WorkflowRunStatusOutcomeUnknown)
	}
	return status, failureCode, failureSummary, ref
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
