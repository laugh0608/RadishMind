package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"time"
)

var workflowHTTPToolExecutionBaseScopes = []string{
	"workflow_tool_actions:execute",
	"workflow_runs:execute",
}

var workflowHTTPToolExecutionRequiredScopes = []string{
	"workflow_tool_actions:execute",
	"workflow_runs:execute",
	"workflow_drafts:read",
}

var workflowDefinitionHTTPToolExecutionRequiredScopes = []string{
	"workflow_tool_actions:execute",
	"workflow_runs:execute",
	"workflow_definitions:read",
}

const workflowHTTPToolReconcilerActorRef = "system:workflow_http_tool_reconciler"

type WorkflowHTTPToolExecutionRequest struct {
	PlanID                string
	ExpectedRecordVersion int
	InputText             string
	Model                 string
	Temperature           *float64
}

type WorkflowHTTPToolExecutionResult struct {
	ActionPlan     *WorkflowHTTPToolActionPlan
	Record         *WorkflowRunRecord
	FailureCode    WorkflowRunFailureCode
	FailureSummary string
}

type workflowHTTPToolExecutionService struct {
	actions      workflowHTTPToolActionService
	store        workflowHTTPToolExecutionStore
	executor     workflowExecutorService
	applications applicationCatalogRepository
	transport    workflowHTTPToolTransport
	maxRuntime   time.Duration
	now          func() time.Time
	newID        func(string) (string, error)
	newRunID     func() (string, error)
}

func newWorkflowHTTPToolExecutionService(
	actions workflowHTTPToolActionService,
	store workflowHTTPToolExecutionStore,
	executor workflowExecutorService,
	applications applicationCatalogRepository,
) workflowHTTPToolExecutionService {
	return workflowHTTPToolExecutionService{
		actions: actions, store: store, executor: executor, applications: applications,
		transport: newWorkflowHTTPToolTransport(), maxRuntime: workflowExecutorDefaultMaxRuntime,
		now: func() time.Time { return time.Now().UTC() }, newID: newWorkflowHTTPToolActionID, newRunID: newWorkflowRunID,
	}
}

func (s *Server) workflowHTTPToolExecutionService() (workflowHTTPToolExecutionService, error) {
	if s == nil || s.workflowRunStore == nil || s.workflowHTTPToolActionStore == nil ||
		s.workflowHTTPToolExecutionStore == nil || s.bridge == nil {
		return workflowHTTPToolExecutionService{}, errWorkflowHTTPToolActionUnavailable
	}
	actions, err := newWorkflowHTTPToolActionService(
		s.savedWorkflowDraftService().ReadDraft,
		s.workflowDefinitionReleaseRepository,
		s.workflowHTTPToolActionStore,
	)
	if err != nil {
		return workflowHTTPToolExecutionService{}, err
	}
	executor := newWorkflowExecutorService(s.savedWorkflowDraftService().ReadDraft, s.bridge, s.workflowRunStore)
	executor.defaultTemperature = s.config.Temperature
	executor.resolveSelection = func(ctx context.Context, requestedModel string) northboundSelection {
		return s.resolveNorthboundSelection(ctx, requestedModel, nil)
	}
	executor.envelopeOptions = s.buildBridgeEnvelopeOptions
	service := newWorkflowHTTPToolExecutionService(actions, s.workflowHTTPToolExecutionStore, executor, s.applicationCatalogRepository)
	if s.workflowHTTPToolExecutionTransport != nil {
		service.transport = *s.workflowHTTPToolExecutionTransport
	}
	service.transport.allowTestOnlyLoopback = s.config.WorkflowHTTPToolTestLoopbackEnabled && s.options.TestOnly
	return service, nil
}

func (service workflowHTTPToolExecutionService) Execute(
	ctx WorkflowHTTPToolActionContext,
	request WorkflowHTTPToolExecutionRequest,
) WorkflowHTTPToolExecutionResult {
	normalized, failureCode, failureSummary := normalizeWorkflowHTTPToolExecutionRequest(request)
	if failureCode != "" {
		return workflowHTTPToolExecutionFailure(failureCode, failureSummary)
	}
	if validateWorkflowHTTPToolActionContext(ctx) != "" || !workflowHTTPToolExecutionBaseScopesAllowed(ctx.ScopeGrants) {
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureScopeDenied, "Workflow HTTP tool execution scope is denied.")
	}
	if service.store == nil || service.actions.store == nil || service.actions.readDraft == nil ||
		service.executor.bridge == nil || service.executor.store == nil || service.now == nil || service.newID == nil || service.newRunID == nil {
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureToolStore, "Workflow HTTP tool execution service is unavailable.")
	}

	rawPlan, found, err := service.actions.store.ReadPlan(ctx, normalized.PlanID)
	if err != nil {
		return workflowHTTPToolExecutionStoreFailure(err)
	}
	if !found {
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureToolConfirmation, "Workflow HTTP tool approval is not eligible for execution.")
	}
	if !workflowHTTPToolExecutionSourceScopeAllowed(ctx.ScopeGrants, rawPlan) {
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureScopeDenied, "Workflow HTTP tool execution source scope is denied.")
	}
	planResult := service.actions.ReadPlan(ctx, normalized.PlanID)
	if planResult.FailureCode != "" {
		return workflowHTTPToolExecutionFailureForAction(planResult)
	}
	if planResult.ActionPlan == nil || planResult.ActionPlan.RecordVersion != normalized.ExpectedRecordVersion ||
		planResult.ActionPlan.Status != WorkflowHTTPToolActionStatusApproved {
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureToolConfirmation, "Workflow HTTP tool approval is not eligible for execution.")
	}
	approvedPlan := cloneWorkflowHTTPToolActionPlan(*planResult.ActionPlan)
	confirmation, found, err := service.store.ReadApprovedConfirmation(ctx, approvedPlan.PlanID, approvedPlan.RecordVersion)
	if err != nil {
		return workflowHTTPToolExecutionStoreFailure(err)
	}
	if !found || !workflowHTTPToolApprovalMatchesPlan(confirmation, approvedPlan) {
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureToolConfirmation, "Workflow HTTP tool approval could not be matched to the approved plan.")
	}

	source, sourceResult := service.resolveExecutionSource(ctx, approvedPlan)
	if sourceResult.FailureCode != "" {
		return sourceResult
	}
	executionPlan, failureCode, failureSummary := buildWorkflowHTTPToolExecutionPlan(
		source,
		approvedPlan.NodeID,
		service.actions.registry.definition,
	)
	if failureCode != "" {
		return workflowHTTPToolExecutionFailure(failureCode, failureSummary)
	}
	if err := validateWorkflowHTTPToolExecutionBinding(
		approvedPlan,
		service.actions.registry.profile,
		ctx.RequestID,
		service.transport.allowTestOnlyLoopback,
	); err != nil {
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureToolPolicy, "Workflow HTTP tool policy no longer matches the approved plan.")
	}
	promptPacket := buildWorkflowPromptPacket(executionPlan.nodes[executionPlan.rootNodeID], normalized.InputText)
	if len([]byte(promptPacket)) > workflowExecutorMaxPacketBytes {
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureBudgetExceeded, "Workflow prompt packet exceeded the execution budget.")
	}

	requestContext := ctx.RequestContext
	maxRuntime := service.maxRuntime
	if maxRuntime <= 0 || maxRuntime > workflowExecutorDefaultMaxRuntime {
		maxRuntime = workflowExecutorDefaultMaxRuntime
	}
	executionContext, cancel := context.WithTimeout(requestContext, maxRuntime)
	defer cancel()
	claimedAt := service.now().UTC()
	attemptID, runID, startedAuditID, idErr := service.allocateClaimIDs()
	if idErr != nil {
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureToolStore, "Workflow HTTP tool execution identifiers could not be allocated.")
	}
	attempt := newWorkflowHTTPToolExecutionAttempt(approvedPlan, confirmation, attemptID, claimedAt)
	run := newWorkflowHTTPToolRunRecord(
		ctx,
		normalized,
		source,
		executionPlan,
		northboundSelection{},
		runID,
		approvedPlan,
		confirmation,
		attempt,
		claimedAt,
	)
	consumedPlan := cloneWorkflowHTTPToolActionPlan(approvedPlan)
	consumedPlan.Status = WorkflowHTTPToolActionStatusConsumed
	consumedPlan.RecordVersion++
	startedAudit := newWorkflowHTTPToolExecutionStartedAudit(ctx, consumedPlan, confirmation, attempt, run, startedAuditID)
	if err := service.store.ClaimExecution(ctx, &consumedPlan, confirmation, &attempt, &run, startedAudit); err != nil {
		return workflowHTTPToolExecutionClaimFailure(err)
	}

	result := service.executeClaimedPlan(
		executionContext,
		ctx,
		normalized,
		source,
		executionPlan,
		approvedPlan,
		attempt,
		run,
		promptPacket,
	)
	result.ActionPlan = workflowHTTPToolActionPlanPointer(consumedPlan)
	return result
}

func (service workflowHTTPToolExecutionService) ReconcileStale(
	ctx WorkflowHTTPToolActionContext,
	planID string,
) WorkflowHTTPToolExecutionResult {
	planID = strings.TrimSpace(planID)
	if !workflowHTTPToolPlanIDPattern.MatchString(planID) {
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureInputInvalid, "Workflow HTTP tool plan id is invalid.")
	}
	if validateWorkflowHTTPToolActionContext(ctx) != "" || !workflowHTTPToolExecutionBaseScopesAllowed(ctx.ScopeGrants) {
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureScopeDenied, "Workflow HTTP tool reconciliation scope is denied.")
	}
	if service.store == nil || service.now == nil || service.newID == nil {
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureToolStore, "Workflow HTTP tool reconciliation store is unavailable.")
	}
	attempt, run, found, err := service.store.ReadClaimedExecution(ctx, planID)
	if err != nil {
		return workflowHTTPToolExecutionStoreFailure(err)
	}
	if !found {
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureToolConfirmation, "Workflow HTTP tool plan has no claimed execution requiring reconciliation.")
	}
	if !workflowHTTPToolExecutionRunSourceScopeAllowed(ctx.ScopeGrants, run) {
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureScopeDenied, "Workflow HTTP tool reconciliation source scope is denied.")
	}
	claimedAt, err := time.Parse(time.RFC3339Nano, attempt.ClaimedAt)
	if err != nil {
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureToolStore, "Workflow HTTP tool claimed timestamp is incompatible.")
	}
	observedAt := service.now().UTC()
	if observedAt.Before(claimedAt.Add(workflowHTTPToolAttemptClaimedBudget)) {
		return WorkflowHTTPToolExecutionResult{Record: workflowRunRecordPointer(run)}
	}
	attempt.Status = WorkflowHTTPToolAttemptOutcomeUnknown
	attempt.CompletedAt = workflowRunTimestamp(observedAt)
	attempt.OutputProjection = map[string]any{}
	attempt.FailureCode = WorkflowRunFailureToolOutcomeUnknown
	run.Status = WorkflowRunStatusOutcomeUnknown
	run.FailureCode = WorkflowRunFailureToolOutcomeUnknown
	run.FailureSummary = "Workflow HTTP tool execution exceeded the claimed runtime budget and requires manual review."
	run.CompletedAt = workflowRunTimestamp(observedAt)
	run.ToolAttempt = &attempt
	setWorkflowRunFailureDiagnostic(&run, WorkflowRunFailureToolOutcomeUnknown, attempt.NodeID, WorkflowRunGatewayFailureNone)
	if run.Diagnostic != nil {
		run.Diagnostic.ToolFailureCategory = WorkflowHTTPToolFailureOutcomeUnknown
		run.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
		run.Diagnostic.ObservedAt = run.CompletedAt
	}
	auditID, err := service.newID("wtae_")
	if err != nil {
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureToolStore, "Workflow HTTP tool reconciliation audit id could not be allocated.")
	}
	audit := newWorkflowHTTPToolExecutionCompletionAudit(ctx, attempt, run, auditID, 1)
	audit.ActorRef = workflowHTTPToolReconcilerActorRef
	audit.ActorSource = "system"
	if err := service.store.CompleteExecution(ctx, &attempt, &run, audit); err != nil {
		if errors.Is(err, errWorkflowHTTPToolExecutionConflict) {
			return workflowHTTPToolExecutionFailure(WorkflowRunFailureToolConfirmation, "Workflow HTTP tool execution changed before reconciliation was committed.")
		}
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureToolStore, "Workflow HTTP tool reconciliation could not commit the terminal state.")
	}
	return WorkflowHTTPToolExecutionResult{
		Record: workflowRunRecordPointer(run), FailureCode: run.FailureCode, FailureSummary: run.FailureSummary,
	}
}

func (service workflowHTTPToolExecutionService) executeClaimedPlan(
	executionContext context.Context,
	ctx WorkflowHTTPToolActionContext,
	request WorkflowHTTPToolExecutionRequest,
	source workflowHTTPToolResolvedExecutionSource,
	plan workflowExecutionPlan,
	approvedPlan WorkflowHTTPToolActionPlan,
	attempt WorkflowHTTPToolExecutionAttempt,
	run WorkflowRunRecord,
	promptPacket string,
) WorkflowHTTPToolExecutionResult {
	draft := source.draft
	edgeActive := make(map[string]bool, len(draft.Edges))
	nodeOutputs := make(map[string]string, len(draft.Nodes))
	nodeRecordIndex := make(map[string]int, len(run.Nodes))
	for index, nodeRecord := range run.Nodes {
		nodeRecordIndex[nodeRecord.NodeID] = index
	}
	networkAttempts := 0
	toolCompleted := false
	executorRequest := WorkflowRunRequest{
		DraftID: draft.DraftID, InputText: request.InputText, Model: request.Model, Temperature: request.Temperature,
	}
	selection := resolveWorkflowRunSelection(service.executor, executionContext, request.Model)
	applyWorkflowHTTPToolRunSelection(&run, selection)

	for _, nodeID := range plan.order {
		if err := executionContext.Err(); err != nil {
			return service.completeClaimedFailure(
				ctx, &attempt, &run, WorkflowRunFailureCanceled,
				"Workflow run was canceled or exceeded its execution deadline.",
				WorkflowHTTPToolFailureNone, nodeID, networkAttempts, toolCompleted,
			)
		}
		node := plan.nodes[nodeID]
		recordIndex := nodeRecordIndex[nodeID]
		nodeStartedAt := service.now().UTC()
		run.Nodes[recordIndex].Status = WorkflowRunNodeStatusRunning
		run.Nodes[recordIndex].StartedAt = workflowRunTimestamp(nodeStartedAt)

		var output string
		var failureCode WorkflowRunFailureCode
		var failureSummary string
		toolFailureCategory := WorkflowHTTPToolFailureNone
		if node.NodeType == "http_tool" {
			if source.definitionBound() {
				currentSource, sourceResult := service.resolveDefinitionExecutionSource(ctx, approvedPlan)
				if sourceResult.FailureCode != "" || !workflowHTTPToolDefinitionSourceMatches(source, currentSource) {
					return service.completeClaimedFailure(
						ctx, &attempt, &run, WorkflowRunFailureDefinitionAuthority,
						"Workflow definition execution authority changed after the tool execution claim.",
						WorkflowHTTPToolFailureNone, nodeID, networkAttempts, false,
					)
				}
			}
			transportResult := service.transport.Execute(
				executionContext,
				approvedPlan,
				service.actions.registry.profile,
				ctx.RequestID,
			)
			if transportResult.NetworkAttempted {
				networkAttempts = 1
			}
			attempt.HTTPStatusClass = transportResult.HTTPStatusClass
			attempt.ResponseBytes = transportResult.ResponseBytes
			attempt.DurationMS = transportResult.DurationMS
			attempt.CompletedAt = workflowRunTimestamp(service.now().UTC())
			if transportResult.FailureCode != "" {
				if transportResult.NetworkAttempted &&
					(transportResult.FailureCode == WorkflowRunFailureToolTimeout || transportResult.FailureCode == WorkflowRunFailureToolTransport) {
					return service.completeClaimedOutcomeUnknown(
						ctx,
						&attempt,
						&run,
						nodeID,
						networkAttempts,
						"Workflow HTTP tool transport outcome could not be determined safely.",
					)
				}
				failureCode = transportResult.FailureCode
				failureSummary = transportResult.FailureSummary
				toolFailureCategory = transportResult.FailureCategory
			} else {
				attempt.Status = WorkflowHTTPToolAttemptSucceeded
				attempt.OutputProjection = cloneWorkflowHTTPToolProjection(transportResult.OutputProjection)
				encodedProjection, err := json.Marshal(attempt.OutputProjection)
				if err != nil || len(encodedProjection) > approvedPlan.MaxOutputBytes {
					return service.completeClaimedFailure(
						ctx, &attempt, &run, WorkflowRunFailureToolResponseInvalid,
						"Workflow HTTP tool projection could not be encoded within the reviewed budget.",
						WorkflowHTTPToolFailureResponseInvalid, nodeID, networkAttempts, false,
					)
				}
				output = string(encodedProjection)
				toolCompleted = true
			}
		} else if node.NodeType == "prompt" {
			output = promptPacket
		} else {
			output, failureCode, failureSummary = service.executor.executeNode(
				executionContext,
				executorRequest,
				draft,
				plan,
				node,
				edgeActive,
				nodeOutputs,
				selection,
				&run,
			)
		}

		nodeCompletedAt := service.now().UTC()
		run.Nodes[recordIndex].CompletedAt = workflowRunTimestamp(nodeCompletedAt)
		run.Nodes[recordIndex].DurationMS = nodeCompletedAt.Sub(nodeStartedAt).Milliseconds()
		if failureCode != "" {
			run.Nodes[recordIndex].Status = WorkflowRunNodeStatusFailed
			run.Nodes[recordIndex].FailureCode = failureCode
			return service.completeClaimedFailure(
				ctx, &attempt, &run, failureCode, failureSummary,
				toolFailureCategory, nodeID, networkAttempts, toolCompleted,
			)
		}
		nodeOutputs[nodeID] = output
		run.Nodes[recordIndex].Status = WorkflowRunNodeStatusSucceeded
		if !source.definitionBound() {
			run.Nodes[recordIndex].OutputPreview = workflowRunNodeOutputPreview(node.NodeType, output)
		}
		if run.Diagnostic != nil {
			run.Diagnostic.LastCompletedNodeID = nodeID
			run.Diagnostic.ObservedAt = workflowRunTimestamp(nodeCompletedAt)
		}
		for _, edge := range plan.outgoing[nodeID] {
			edgeActive[edge.EdgeID] = true
		}
	}

	output := strings.TrimSpace(nodeOutputs[plan.outputNodeID])
	if output == "" {
		return service.completeClaimedFailure(
			ctx, &attempt, &run, WorkflowRunFailureOutputUnavailable,
			"No workflow output was produced after the reviewed HTTP tool execution.",
			WorkflowHTTPToolFailureNone, plan.outputNodeID, networkAttempts, toolCompleted,
		)
	}
	run.Status = WorkflowRunStatusSucceeded
	if !source.definitionBound() {
		run.Output = output
	}
	run.CompletedAt = workflowRunTimestamp(service.now().UTC())
	run.FailureCode = ""
	run.FailureSummary = ""
	if run.Diagnostic != nil {
		run.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
		run.Diagnostic.ToolFailureCategory = WorkflowHTTPToolFailureNone
		run.Diagnostic.ObservedAt = run.CompletedAt
	}
	run.ToolAttempt = &attempt
	return service.commitClaimedTerminal(ctx, attempt, run, networkAttempts)
}

func (service workflowHTTPToolExecutionService) completeClaimedFailure(
	ctx WorkflowHTTPToolActionContext,
	attempt *WorkflowHTTPToolExecutionAttempt,
	run *WorkflowRunRecord,
	failureCode WorkflowRunFailureCode,
	failureSummary string,
	toolCategory WorkflowHTTPToolFailureCategory,
	nodeID string,
	networkAttempts int,
	toolCompleted bool,
) WorkflowHTTPToolExecutionResult {
	completedAt := service.now().UTC()
	if !toolCompleted {
		attempt.Status = WorkflowHTTPToolAttemptFailed
		attempt.CompletedAt = workflowRunTimestamp(completedAt)
		attempt.OutputProjection = map[string]any{}
		attempt.FailureCode = failureCode
	}
	if failureCode == WorkflowRunFailureCanceled {
		run.Status = WorkflowRunStatusCanceled
	} else {
		run.Status = WorkflowRunStatusFailed
	}
	run.FailureCode = failureCode
	run.FailureSummary = failureSummary
	run.CompletedAt = workflowRunTimestamp(completedAt)
	if run.Diagnostic != nil {
		if run.Diagnostic.FailureBoundary == "" || toolCategory != WorkflowHTTPToolFailureNone {
			setWorkflowRunFailureDiagnostic(run, failureCode, nodeID, run.Diagnostic.GatewayFailureCategory)
		}
		run.Diagnostic.ToolFailureCategory = toolCategory
		run.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
		run.Diagnostic.ObservedAt = run.CompletedAt
	}
	run.ToolAttempt = attempt
	return service.commitClaimedTerminal(ctx, *attempt, *run, networkAttempts)
}

func (service workflowHTTPToolExecutionService) completeClaimedOutcomeUnknown(
	ctx WorkflowHTTPToolActionContext,
	attempt *WorkflowHTTPToolExecutionAttempt,
	run *WorkflowRunRecord,
	nodeID string,
	networkAttempts int,
	summary string,
) WorkflowHTTPToolExecutionResult {
	completedAt := service.now().UTC()
	attempt.Status = WorkflowHTTPToolAttemptOutcomeUnknown
	attempt.CompletedAt = workflowRunTimestamp(completedAt)
	attempt.OutputProjection = map[string]any{}
	attempt.FailureCode = WorkflowRunFailureToolOutcomeUnknown
	run.Status = WorkflowRunStatusOutcomeUnknown
	run.FailureCode = WorkflowRunFailureToolOutcomeUnknown
	run.FailureSummary = summary
	run.CompletedAt = workflowRunTimestamp(completedAt)
	setWorkflowRunFailureDiagnostic(run, WorkflowRunFailureToolOutcomeUnknown, nodeID, WorkflowRunGatewayFailureNone)
	if run.Diagnostic != nil {
		run.Diagnostic.ToolFailureCategory = WorkflowHTTPToolFailureOutcomeUnknown
		run.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
		run.Diagnostic.ObservedAt = run.CompletedAt
	}
	run.ToolAttempt = attempt
	return service.commitClaimedTerminal(ctx, *attempt, *run, networkAttempts)
}

func (service workflowHTTPToolExecutionService) commitClaimedTerminal(
	ctx WorkflowHTTPToolActionContext,
	attempt WorkflowHTTPToolExecutionAttempt,
	run WorkflowRunRecord,
	networkAttempts int,
) WorkflowHTTPToolExecutionResult {
	auditID, err := service.newID("wtae_")
	if err != nil {
		return service.workflowHTTPToolOutcomeUnknownAfterTerminalWrite(run)
	}
	audit := newWorkflowHTTPToolExecutionCompletionAudit(ctx, attempt, run, auditID, networkAttempts)
	if err := service.store.CompleteExecution(ctx, &attempt, &run, audit); err != nil {
		return service.workflowHTTPToolOutcomeUnknownAfterTerminalWrite(run)
	}
	return WorkflowHTTPToolExecutionResult{
		Record: workflowRunRecordPointer(run), FailureCode: run.FailureCode, FailureSummary: run.FailureSummary,
	}
}

func (service workflowHTTPToolExecutionService) workflowHTTPToolOutcomeUnknownAfterTerminalWrite(run WorkflowRunRecord) WorkflowHTTPToolExecutionResult {
	completedAt := workflowRunTimestamp(service.now().UTC())
	if run.ToolAttempt == nil {
		run.ToolAttempt = &WorkflowHTTPToolExecutionAttempt{OutputProjection: map[string]any{}}
	}
	run.Status = WorkflowRunStatusOutcomeUnknown
	run.FailureCode = WorkflowRunFailureToolOutcomeUnknown
	run.FailureSummary = "Workflow HTTP tool terminal state could not be stored; reconciliation is required."
	run.CompletedAt = completedAt
	run.ToolAttempt.Status = WorkflowHTTPToolAttemptOutcomeUnknown
	run.ToolAttempt.CompletedAt = completedAt
	run.ToolAttempt.OutputProjection = map[string]any{}
	run.ToolAttempt.FailureCode = WorkflowRunFailureToolOutcomeUnknown
	setWorkflowRunFailureDiagnostic(&run, WorkflowRunFailureToolOutcomeUnknown, run.ToolAttempt.NodeID, WorkflowRunGatewayFailureNone)
	if run.Diagnostic != nil {
		run.Diagnostic.ToolFailureCategory = WorkflowHTTPToolFailureOutcomeUnknown
		run.Diagnostic.TerminalWriteState = WorkflowRunTerminalWritePending
		run.Diagnostic.ObservedAt = completedAt
	}
	return WorkflowHTTPToolExecutionResult{
		Record: workflowRunRecordPointer(run), FailureCode: WorkflowRunFailureToolOutcomeUnknown,
		FailureSummary: run.FailureSummary,
	}
}

func normalizeWorkflowHTTPToolExecutionRequest(
	request WorkflowHTTPToolExecutionRequest,
) (WorkflowHTTPToolExecutionRequest, WorkflowRunFailureCode, string) {
	normalized := WorkflowHTTPToolExecutionRequest{
		PlanID: strings.TrimSpace(request.PlanID), ExpectedRecordVersion: request.ExpectedRecordVersion,
		InputText: strings.TrimSpace(request.InputText), Model: strings.TrimSpace(request.Model), Temperature: request.Temperature,
	}
	if !workflowHTTPToolPlanIDPattern.MatchString(normalized.PlanID) || normalized.ExpectedRecordVersion < 1 || normalized.InputText == "" {
		return WorkflowHTTPToolExecutionRequest{}, WorkflowRunFailureInputInvalid, "Workflow HTTP tool execution input is invalid."
	}
	if len([]byte(normalized.InputText)) > workflowExecutorMaxInputBytes || len([]rune(normalized.Model)) > workflowExecutorMaxModelChars {
		return WorkflowHTTPToolExecutionRequest{}, WorkflowRunFailureBudgetExceeded, "Workflow HTTP tool execution input exceeded its budget."
	}
	if normalized.Temperature != nil {
		value := *normalized.Temperature
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 2 {
			return WorkflowHTTPToolExecutionRequest{}, WorkflowRunFailureInputInvalid, "Workflow run temperature must be between 0 and 2."
		}
	}
	return normalized, "", ""
}

func buildWorkflowHTTPToolExecutionPlan(
	source workflowHTTPToolResolvedExecutionSource,
	targetNodeID string,
	definition WorkflowHTTPToolDefinition,
) (workflowExecutionPlan, WorkflowRunFailureCode, string) {
	draft := source.draft
	if draft.DraftVersion < 1 {
		return workflowExecutionPlan{}, WorkflowRunFailureDraftVersionUnavailable, "Workflow draft does not have a persisted executable version."
	}
	var eligibilityError error
	if source.definitionBound() {
		eligibilityError = validateWorkflowHTTPToolGraph(draft, targetNodeID, definition)
	} else {
		eligibilityError = validateWorkflowHTTPToolDraft(draft, targetNodeID, definition)
	}
	if eligibilityError != nil {
		return workflowExecutionPlan{}, WorkflowRunFailureDraftNotEligible, eligibilityError.Error()
	}
	plan := workflowExecutionPlan{
		nodes:    make(map[string]SavedWorkflowDraftNode, len(draft.Nodes)),
		incoming: make(map[string][]SavedWorkflowDraftEdge, len(draft.Nodes)),
		outgoing: make(map[string][]SavedWorkflowDraftEdge, len(draft.Nodes)),
	}
	positions := make(map[string]int, len(draft.Nodes))
	indegree := make(map[string]int, len(draft.Nodes))
	for index, node := range draft.Nodes {
		node.NodeID = strings.TrimSpace(node.NodeID)
		node.NodeType = strings.ToLower(strings.TrimSpace(node.NodeType))
		plan.nodes[node.NodeID] = node
		positions[node.NodeID] = index
		indegree[node.NodeID] = 0
	}
	for _, edge := range draft.Edges {
		edge.EdgeID = strings.TrimSpace(edge.EdgeID)
		edge.FromNodeID = strings.TrimSpace(edge.FromNodeID)
		edge.ToNodeID = strings.TrimSpace(edge.ToNodeID)
		plan.outgoing[edge.FromNodeID] = append(plan.outgoing[edge.FromNodeID], edge)
		plan.incoming[edge.ToNodeID] = append(plan.incoming[edge.ToNodeID], edge)
		indegree[edge.ToNodeID]++
	}
	for nodeID, node := range plan.nodes {
		if indegree[nodeID] == 0 {
			plan.rootNodeID = nodeID
		}
		if len(plan.outgoing[nodeID]) == 0 && node.NodeType == "output" {
			plan.outputNodeID = nodeID
		}
	}
	plan.order = workflowStableTopologicalOrder(indegree, plan.outgoing, positions)
	if len(plan.order) != len(plan.nodes) || plan.rootNodeID == "" || plan.outputNodeID == "" {
		return workflowExecutionPlan{}, WorkflowRunFailureGraphInvalid, "Workflow HTTP tool graph could not be ordered safely."
	}
	return plan, "", ""
}

func newWorkflowHTTPToolExecutionAttempt(
	plan WorkflowHTTPToolActionPlan,
	confirmation WorkflowHTTPToolConfirmationDecision,
	attemptID string,
	claimedAt time.Time,
) WorkflowHTTPToolExecutionAttempt {
	return WorkflowHTTPToolExecutionAttempt{
		AttemptID: attemptID, NodeID: plan.NodeID, ToolID: plan.ToolID,
		DefinitionDigest: plan.DefinitionDigest, ProfileID: plan.ProfileID, ProfileDigest: plan.ProfileDigest,
		ToolPlanDigest: plan.ToolPlanDigest, ConfirmationID: confirmation.ConfirmationID,
		Status: WorkflowHTTPToolAttemptClaimed, ClaimedAt: workflowRunTimestamp(claimedAt), OutputProjection: map[string]any{},
	}
}

func newWorkflowHTTPToolRunRecord(
	ctx WorkflowHTTPToolActionContext,
	request WorkflowHTTPToolExecutionRequest,
	source workflowHTTPToolResolvedExecutionSource,
	executionPlan workflowExecutionPlan,
	selection northboundSelection,
	runID string,
	plan WorkflowHTTPToolActionPlan,
	confirmation WorkflowHTTPToolConfirmationDecision,
	attempt WorkflowHTTPToolExecutionAttempt,
	claimedAt time.Time,
) WorkflowRunRecord {
	draft := source.draft
	runContext := workflowRunContextFromToolAction(ctx)
	runRequest := WorkflowRunRequest{DraftID: draft.DraftID, InputText: request.InputText, Model: request.Model, Temperature: request.Temperature}
	record := newWorkflowRunRecord(runContext, runRequest, draft, executionPlan, selection, runID)
	record.SchemaVersion = workflowRunRecordToolSchemaVersion
	record.PlanID = plan.PlanID
	record.ConfirmationID = confirmation.ConfirmationID
	record.TenantRef = ctx.TenantRef
	record.StartedAt = workflowRunTimestamp(claimedAt)
	record.ToolAttempt = &attempt
	record.SideEffects.ToolCalls = 1
	record.SideEffects.ConfirmationCalls = 1
	record.ConditionNodeIDs = nil
	if source.definitionBound() {
		record.SchemaVersion = workflowRunRecordDefinitionToolSchemaVersion
		record.DraftID = ""
		record.DraftVersion = 0
		record.DraftDigest = ""
		record.ExecutionKind = workflowDefinitionHTTPToolExecutionKind
		record.ExecutionSourceKind = workflowDefinitionExecutionSourceKind
		record.ExecutionSourceID = source.version.DefinitionID
		record.ExecutionSourceVersion = source.version.Version
		record.ExecutionProfile = workflowDefinitionHTTPToolProfile
		record.ExecutionSource = &workflowRunExecutionSource{
			Kind: workflowDefinitionHTTPToolExecutionKind, SourceKind: workflowDefinitionExecutionSourceKind,
			ID: source.version.DefinitionID, Version: source.version.Version,
		}
		record.DefinitionAuthority = workflowDefinitionHTTPToolRunAuthority(source)
		record.InputDigest = workflowDefinitionInputDigest(request.InputText)
		record.Output = ""
		record.ConditionNodeIDs = []string{}
		for index := range record.Nodes {
			record.Nodes[index].OutputPreview = ""
		}
	}
	if record.Diagnostic != nil {
		record.Diagnostic.ToolFailureCategory = WorkflowHTTPToolFailureNone
		record.Diagnostic.ObservedAt = record.StartedAt
	}
	return record
}

func applyWorkflowHTTPToolRunSelection(record *WorkflowRunRecord, selection northboundSelection) {
	if record == nil {
		return
	}
	record.SelectedProvider = strings.TrimSpace(selection.provider)
	record.SelectedProfile = strings.TrimSpace(selection.providerProfile)
	record.SelectedModel = strings.TrimSpace(selection.model)
	record.UpstreamModel = strings.TrimSpace(selection.upstreamModel)
	record.SelectionSource = strings.TrimSpace(selection.source)
}

func newWorkflowHTTPToolExecutionStartedAudit(
	ctx WorkflowHTTPToolActionContext,
	plan WorkflowHTTPToolActionPlan,
	confirmation WorkflowHTTPToolConfirmationDecision,
	attempt WorkflowHTTPToolExecutionAttempt,
	run WorkflowRunRecord,
	eventID string,
) WorkflowHTTPToolExecutionAudit {
	attemptID, runID, confirmationID := attempt.AttemptID, run.RunID, confirmation.ConfirmationID
	return WorkflowHTTPToolExecutionAudit{
		SchemaVersion: workflowHTTPToolAuditSchemaForPlan(plan), EventID: eventID, EventKind: "tool_execution_started",
		OccurredAt: attempt.ClaimedAt, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
		ApplicationID: ctx.ApplicationID, DraftID: plan.DraftID, DraftVersion: plan.DraftVersion,
		SourceKind: plan.SourceKind, WorkflowDefinitionID: plan.WorkflowDefinitionID,
		WorkflowDefinitionVersion: plan.WorkflowDefinitionVersion, WorkflowDefinitionDigest: plan.WorkflowDefinitionDigest,
		ActivationPointerVersion: plan.ActivationPointerVersion,
		NodeID:                   plan.NodeID, PlanID: plan.PlanID, ConfirmationID: &confirmationID,
		ExecutionAttemptID: &attemptID, RunID: &runID, ToolID: plan.ToolID, ToolVersion: plan.ToolVersion,
		DefinitionDigest: plan.DefinitionDigest, ProfileID: plan.ProfileID, ProfileDigest: plan.ProfileDigest,
		ToolPlanDigest: plan.ToolPlanDigest, ActorRef: ctx.ActorRef, ActorSource: "human",
		RequestID: ctx.RequestID, AuditRef: "audit_" + eventID, AttemptStatus: "claimed",
		SideEffects: WorkflowHTTPToolAuditSideEffects{},
	}
}

func newWorkflowHTTPToolExecutionCompletionAudit(
	ctx WorkflowHTTPToolActionContext,
	attempt WorkflowHTTPToolExecutionAttempt,
	run WorkflowRunRecord,
	eventID string,
	networkAttempts int,
) WorkflowHTTPToolExecutionAudit {
	attemptID, runID, confirmationID := attempt.AttemptID, run.RunID, attempt.ConfirmationID
	audit := WorkflowHTTPToolExecutionAudit{
		SchemaVersion: workflowHTTPToolAuditSchema, EventID: eventID, OccurredAt: attempt.CompletedAt,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		DraftID: run.DraftID, DraftVersion: run.DraftVersion, NodeID: attempt.NodeID, PlanID: run.PlanID,
		ConfirmationID: &confirmationID, ExecutionAttemptID: &attemptID, RunID: &runID,
		ToolID: attempt.ToolID, ToolVersion: workflowHTTPToolVersion, DefinitionDigest: attempt.DefinitionDigest,
		ProfileID: attempt.ProfileID, ProfileDigest: attempt.ProfileDigest, ToolPlanDigest: attempt.ToolPlanDigest,
		ActorRef: ctx.ActorRef, ActorSource: "human", RequestID: ctx.RequestID, AuditRef: "audit_" + eventID,
		HTTPStatusClass: workflowHTTPToolOptionalString(attempt.HTTPStatusClass), ResponseBytes: attempt.ResponseBytes,
		DurationMS:  attempt.DurationMS,
		SideEffects: WorkflowHTTPToolAuditSideEffects{NetworkAttempts: networkAttempts, ToolCalls: 1},
	}
	if run.SchemaVersion == workflowRunRecordDefinitionToolSchemaVersion && run.DefinitionAuthority != nil {
		audit.SchemaVersion = workflowHTTPToolAuditSchemaV2
		audit.SourceKind = workflowHTTPToolSourceDefinition
		audit.DraftID = ""
		audit.DraftVersion = 0
		audit.WorkflowDefinitionID = run.DefinitionAuthority.DefinitionID
		audit.WorkflowDefinitionVersion = run.DefinitionAuthority.DefinitionVersion
		audit.WorkflowDefinitionDigest = run.DefinitionAuthority.DefinitionDigest
		audit.ActivationPointerVersion = run.DefinitionAuthority.ActivationPointerVersion
	}
	switch attempt.Status {
	case WorkflowHTTPToolAttemptSucceeded:
		audit.EventKind, audit.AttemptStatus = "tool_execution_succeeded", "succeeded"
	case WorkflowHTTPToolAttemptFailed:
		code := string(attempt.FailureCode)
		audit.EventKind, audit.AttemptStatus, audit.FailureCode = "tool_execution_failed", "failed", &code
	case WorkflowHTTPToolAttemptOutcomeUnknown:
		code := string(WorkflowRunFailureToolOutcomeUnknown)
		audit.EventKind, audit.AttemptStatus, audit.FailureCode = "tool_execution_outcome_unknown", "outcome_unknown", &code
	}
	if attempt.Status != WorkflowHTTPToolAttemptSucceeded && run.Diagnostic != nil && run.Diagnostic.FailureBoundary != "" {
		boundary := string(run.Diagnostic.FailureBoundary)
		audit.FailureBoundary = &boundary
	}
	return audit
}

func (service workflowHTTPToolExecutionService) allocateClaimIDs() (string, string, string, error) {
	attemptID, err := service.newID("wtea_")
	if err != nil {
		return "", "", "", err
	}
	runID, err := service.newRunID()
	if err != nil {
		return "", "", "", err
	}
	auditID, err := service.newID("wtae_")
	if err != nil {
		return "", "", "", err
	}
	return attemptID, runID, auditID, nil
}

func workflowHTTPToolExecutionScopesAllowed(scopes []string) bool {
	return workflowHTTPToolAllScopesAllowed(scopes, workflowHTTPToolExecutionRequiredScopes)
}

func workflowHTTPToolExecutionBaseScopesAllowed(scopes []string) bool {
	return workflowHTTPToolAllScopesAllowed(scopes, workflowHTTPToolExecutionBaseScopes)
}

func workflowHTTPToolExecutionSourceScopeAllowed(scopes []string, plan WorkflowHTTPToolActionPlan) bool {
	if plan.SchemaVersion == workflowHTTPToolPlanSchemaV2 && plan.SourceKind == workflowHTTPToolSourceDefinition {
		return workflowHTTPToolAllScopesAllowed(scopes, workflowDefinitionHTTPToolExecutionRequiredScopes)
	}
	return workflowHTTPToolExecutionScopesAllowed(scopes)
}

func workflowHTTPToolExecutionRunSourceScopeAllowed(scopes []string, run WorkflowRunRecord) bool {
	if run.SchemaVersion == workflowRunRecordDefinitionToolSchemaVersion {
		return workflowHTTPToolAllScopesAllowed(scopes, workflowDefinitionHTTPToolExecutionRequiredScopes)
	}
	return workflowHTTPToolExecutionScopesAllowed(scopes)
}

func workflowHTTPToolAllScopesAllowed(scopes []string, requiredScopes []string) bool {
	for _, required := range requiredScopes {
		if !controlPlaneReadHasScope(scopes, required) {
			return false
		}
	}
	return true
}

func workflowHTTPToolApprovalMatchesPlan(
	confirmation WorkflowHTTPToolConfirmationDecision,
	plan WorkflowHTTPToolActionPlan,
) bool {
	return confirmation.SchemaVersion == workflowHTTPToolDecisionSchemaForPlan(plan) &&
		confirmation.Outcome == WorkflowHTTPToolConfirmationApprove && confirmation.ActorSource == "human" &&
		confirmation.PlanID == plan.PlanID && confirmation.TenantRef == plan.TenantRef &&
		confirmation.WorkspaceID == plan.WorkspaceID && confirmation.ApplicationID == plan.ApplicationID &&
		workflowHTTPToolDecisionSourceMatchesPlan(confirmation, plan) &&
		confirmation.NodeID == plan.NodeID && confirmation.ToolID == plan.ToolID &&
		confirmation.ToolVersion == plan.ToolVersion && confirmation.ToolPlanDigest == plan.ToolPlanDigest &&
		confirmation.ResultingRecordVersion == plan.RecordVersion
}

func cloneWorkflowHTTPToolProjection(source map[string]any) map[string]any {
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func workflowHTTPToolOptionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func workflowHTTPToolExecutionFailure(
	code WorkflowRunFailureCode,
	summary string,
) WorkflowHTTPToolExecutionResult {
	return WorkflowHTTPToolExecutionResult{FailureCode: code, FailureSummary: summary}
}

func workflowHTTPToolExecutionFailureForAction(result WorkflowHTTPToolActionResult) WorkflowHTTPToolExecutionResult {
	switch result.FailureCode {
	case WorkflowHTTPToolActionFailureScopeDenied:
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureScopeDenied, "Workflow HTTP tool execution scope is denied.")
	case WorkflowHTTPToolActionFailureStoreUnavailable, WorkflowHTTPToolActionFailureStoreContract:
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureToolStore, "Workflow HTTP tool execution state could not be read safely.")
	case WorkflowHTTPToolActionFailureDefinitionNotFound, WorkflowHTTPToolActionFailureDefinitionInactive,
		WorkflowHTTPToolActionFailureDefinitionDrift, WorkflowHTTPToolActionFailureDefinitionIneligible:
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureDefinitionAuthority, "Workflow definition execution authority changed before the tool execution claim.")
	case WorkflowHTTPToolActionFailureDefinitionMissing, WorkflowHTTPToolActionFailureProfileUnavailable,
		WorkflowHTTPToolActionFailureDraftIneligible, WorkflowHTTPToolActionFailureDraftVersion:
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureToolPolicy, "Workflow HTTP tool policy no longer matches the approved plan.")
	default:
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureToolConfirmation, "Workflow HTTP tool approval is not eligible for execution.")
	}
}

func workflowHTTPToolExecutionStoreFailure(err error) WorkflowHTTPToolExecutionResult {
	if errors.Is(err, errWorkflowHTTPToolExecutionContract) || errors.Is(err, errWorkflowHTTPToolExecutionUnavailable) {
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureToolStore, "Workflow HTTP tool execution store is unavailable or incompatible.")
	}
	return workflowHTTPToolExecutionFailure(WorkflowRunFailureToolConfirmation, "Workflow HTTP tool approval could not be read safely.")
}

func workflowHTTPToolExecutionClaimFailure(err error) WorkflowHTTPToolExecutionResult {
	if errors.Is(err, errWorkflowHTTPToolExecutionConflict) {
		return workflowHTTPToolExecutionFailure(WorkflowRunFailureToolConfirmation, "Workflow HTTP tool approval was already consumed or changed concurrently.")
	}
	return workflowHTTPToolExecutionFailure(WorkflowRunFailureToolStore, "Workflow HTTP tool execution claim could not be stored atomically.")
}
