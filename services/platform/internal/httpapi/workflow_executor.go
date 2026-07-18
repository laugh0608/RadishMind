package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/bridge"
)

const (
	workflowRunRecordSchemaVersion       = "workflow_run_record.v1"
	workflowRunRecordLegacySchemaVersion = "workflow_run_record.v0"
	workflowRunRecordToolSchemaVersion   = "workflow_run_record.v2"
	workflowRunRecordRAGSchemaVersion    = "workflow_run_record.v3"
	workflowExecutorProtocol             = "workflow-executor-v0"
	workflowExecutorRoute                = "/v1/user-workspace/workflow-drafts/{draft_id}/runs"

	workflowExecutorMaxNodes          = 16
	workflowExecutorMaxEdges          = 32
	workflowExecutorMaxLLMCalls       = 4
	workflowExecutorMaxInputBytes     = 8 * 1024
	workflowExecutorMaxPacketBytes    = 16 * 1024
	workflowExecutorMaxOutputBytes    = 16 * 1024
	workflowExecutorMaxModelChars     = 256
	workflowExecutorMaxConditions     = 16
	workflowExecutorNodePreviewRunes  = 512
	workflowExecutorDefaultMaxRuntime = 30 * time.Second
)

type WorkflowRunStatus string

const (
	WorkflowRunStatusRunning        WorkflowRunStatus = "running"
	WorkflowRunStatusSucceeded      WorkflowRunStatus = "succeeded"
	WorkflowRunStatusFailed         WorkflowRunStatus = "failed"
	WorkflowRunStatusCanceled       WorkflowRunStatus = "canceled"
	WorkflowRunStatusOutcomeUnknown WorkflowRunStatus = "outcome_unknown"
)

type WorkflowRunNodeStatus string

const (
	WorkflowRunNodeStatusPending   WorkflowRunNodeStatus = "pending"
	WorkflowRunNodeStatusRunning   WorkflowRunNodeStatus = "running"
	WorkflowRunNodeStatusSucceeded WorkflowRunNodeStatus = "succeeded"
	WorkflowRunNodeStatusSkipped   WorkflowRunNodeStatus = "skipped"
	WorkflowRunNodeStatusFailed    WorkflowRunNodeStatus = "failed"
)

type WorkflowRunFailureCode string

const (
	WorkflowRunFailureScopeDenied             WorkflowRunFailureCode = "workflow_run_scope_denied"
	WorkflowRunFailureDraftNotFound           WorkflowRunFailureCode = "workflow_run_draft_not_found"
	WorkflowRunFailureDraftVersionUnavailable WorkflowRunFailureCode = "workflow_run_draft_version_unavailable"
	WorkflowRunFailureDraftNotEligible        WorkflowRunFailureCode = "workflow_run_draft_not_eligible"
	WorkflowRunFailureInputInvalid            WorkflowRunFailureCode = "workflow_run_input_invalid"
	WorkflowRunFailureGraphInvalid            WorkflowRunFailureCode = "workflow_run_graph_invalid"
	WorkflowRunFailureBudgetExceeded          WorkflowRunFailureCode = "workflow_run_budget_exceeded"
	WorkflowRunFailureGatewayFailed           WorkflowRunFailureCode = "workflow_run_gateway_failed"
	WorkflowRunFailureOutputUnavailable       WorkflowRunFailureCode = "workflow_run_output_unavailable"
	WorkflowRunFailureCanceled                WorkflowRunFailureCode = "workflow_run_canceled"
	WorkflowRunFailureRecordNotFound          WorkflowRunFailureCode = "workflow_run_record_not_found"
	WorkflowRunFailureStoreUnavailable        WorkflowRunFailureCode = "workflow_run_store_unavailable"
	WorkflowRunFailureStoreContractMismatch   WorkflowRunFailureCode = "workflow_run_store_contract_mismatch"
	WorkflowRunFailureFilterInvalid           WorkflowRunFailureCode = "workflow_run_filter_invalid"
	WorkflowRunFailureCursorInvalid           WorkflowRunFailureCode = "workflow_run_cursor_invalid"
	WorkflowRunFailureStoreModeInvalid        WorkflowRunFailureCode = "workflow_run_store_mode_invalid"
	WorkflowRunFailureStoreModeDisabled       WorkflowRunFailureCode = "workflow_run_store_mode_disabled"
	WorkflowRunFailureComparisonInvalid       WorkflowRunFailureCode = "workflow_run_comparison_invalid"
	WorkflowRunFailureSideEffectUnsupported   WorkflowRunFailureCode = "workflow_run_side_effect_profile_unsupported"
	WorkflowRunFailureRetrievalUnsupported    WorkflowRunFailureCode = "workflow_run_retrieval_profile_unsupported"
	WorkflowRunFailureToolPolicy              WorkflowRunFailureCode = "workflow_tool_policy_denied"
	WorkflowRunFailureToolConfirmation        WorkflowRunFailureCode = "workflow_tool_confirmation_invalid"
	WorkflowRunFailureToolTransport           WorkflowRunFailureCode = "workflow_tool_transport_failed"
	WorkflowRunFailureToolTimeout             WorkflowRunFailureCode = "workflow_tool_timeout"
	WorkflowRunFailureToolResponseStatus      WorkflowRunFailureCode = "workflow_tool_response_status_invalid"
	WorkflowRunFailureToolResponseTooLarge    WorkflowRunFailureCode = "workflow_tool_response_too_large"
	WorkflowRunFailureToolResponseInvalid     WorkflowRunFailureCode = "workflow_tool_response_invalid"
	WorkflowRunFailureToolStore               WorkflowRunFailureCode = "workflow_tool_store_unavailable"
	WorkflowRunFailureToolOutcomeUnknown      WorkflowRunFailureCode = "workflow_tool_outcome_unknown"
)

type WorkflowRunContext struct {
	RequestContext context.Context
	RequestID      string
	TenantRef      string
	WorkspaceID    string
	ApplicationID  string
	ActorRef       string
	ScopeGrants    []string
	AuditRef       string
}

type WorkflowRunRequest struct {
	DraftID            string
	InputText          string
	ConditionValues    map[string]bool
	Model              string
	Temperature        *float64
	DevFailureScenario WorkflowRunDevFailureScenario
}

type WorkflowRunSideEffects struct {
	RetrievalCalls    int `json:"retrieval_calls,omitempty"`
	ProviderCalls     int `json:"provider_calls"`
	ToolCalls         int `json:"tool_calls"`
	ConfirmationCalls int `json:"confirmation_calls"`
	BusinessWrites    int `json:"business_writes"`
	ReplayWrites      int `json:"replay_writes"`
}

type WorkflowRunNodeRecord struct {
	NodeID             string                 `json:"node_id"`
	NodeType           string                 `json:"node_type"`
	Label              string                 `json:"label"`
	Status             WorkflowRunNodeStatus  `json:"status"`
	StartedAt          string                 `json:"started_at"`
	CompletedAt        string                 `json:"completed_at"`
	DurationMS         int64                  `json:"duration_ms"`
	PredecessorNodeIDs []string               `json:"predecessor_node_ids"`
	ProviderRef        string                 `json:"provider_ref"`
	OutputPreview      string                 `json:"output_preview"`
	FailureCode        WorkflowRunFailureCode `json:"failure_code"`
}

type WorkflowRunRecord struct {
	SchemaVersion    string                            `json:"schema_version"`
	RecordVersion    int                               `json:"record_version"`
	RunID            string                            `json:"run_id"`
	PlanID           string                            `json:"plan_id,omitempty"`
	ConfirmationID   string                            `json:"confirmation_id,omitempty"`
	TenantRef        string                            `json:"tenant_ref,omitempty"`
	DraftID          string                            `json:"draft_id"`
	DraftVersion     int                               `json:"draft_version"`
	DraftDigest      string                            `json:"draft_digest,omitempty"`
	WorkspaceID      string                            `json:"workspace_id"`
	ApplicationID    string                            `json:"application_id"`
	Status           WorkflowRunStatus                 `json:"status"`
	FailureCode      WorkflowRunFailureCode            `json:"failure_code"`
	FailureSummary   string                            `json:"failure_summary"`
	StartedAt        string                            `json:"started_at"`
	CompletedAt      string                            `json:"completed_at"`
	InputBytes       int                               `json:"input_bytes"`
	ConditionNodeIDs []string                          `json:"condition_node_ids"`
	RequestedModel   string                            `json:"requested_model"`
	SelectedProvider string                            `json:"selected_provider"`
	SelectedProfile  string                            `json:"selected_profile"`
	SelectedModel    string                            `json:"selected_model"`
	UpstreamModel    string                            `json:"upstream_model"`
	SelectionSource  string                            `json:"selection_source"`
	Nodes            []WorkflowRunNodeRecord           `json:"nodes"`
	ToolAttempt      *WorkflowHTTPToolExecutionAttempt `json:"tool_attempt,omitempty"`
	RAGSnapshot      *workflowRAGRunSnapshotBinding    `json:"snapshot,omitempty"`
	RetrievalAttempt *workflowRAGRunRetrievalAttempt   `json:"retrieval_attempt,omitempty"`
	RAGAnswer        *WorkflowRAGAnswer                `json:"answer,omitempty"`
	Output           string                            `json:"output"`
	RequestID        string                            `json:"request_id"`
	AuditRef         string                            `json:"audit_ref"`
	ActorRef         string                            `json:"actor_ref"`
	SideEffects      WorkflowRunSideEffects            `json:"side_effects"`
	Diagnostic       *WorkflowRunDiagnostic            `json:"diagnostic,omitempty"`
}

type WorkflowRunResult struct {
	Record          *WorkflowRunRecord
	RetrievalAnswer *WorkflowRAGAnswer
	FailureCode     WorkflowRunFailureCode
	FailureSummary  string
}

type workflowSavedDraftReader func(
	context SavedWorkflowDraftContext,
	request ReadWorkflowDraftRequest,
) SavedWorkflowDraftResult

type workflowExecutorService struct {
	draftReader           workflowSavedDraftReader
	bridge                bridgeClient
	store                 workflowRunStore
	maxRuntime            time.Duration
	defaultTemperature    float64
	resolveSelection      func(context.Context, string) northboundSelection
	envelopeOptions       func(northboundSelection, float64) bridge.EnvelopeOptions
	newRunID              func() (string, error)
	diagnosticsDevEnabled bool
}

type workflowExecutionPlan struct {
	order        []string
	nodes        map[string]SavedWorkflowDraftNode
	incoming     map[string][]SavedWorkflowDraftEdge
	outgoing     map[string][]SavedWorkflowDraftEdge
	rootNodeID   string
	outputNodeID string
}

func newWorkflowExecutorService(
	draftReader workflowSavedDraftReader,
	bridgeClient bridgeClient,
	store workflowRunStore,
) workflowExecutorService {
	return workflowExecutorService{
		draftReader: draftReader,
		bridge:      bridgeClient,
		store:       store,
		maxRuntime:  workflowExecutorDefaultMaxRuntime,
		newRunID:    newWorkflowRunID,
	}
}

func (service workflowExecutorService) StartRun(
	runContext WorkflowRunContext,
	request WorkflowRunRequest,
) WorkflowRunResult {
	normalizedRequest, failureCode, failureSummary := normalizeWorkflowRunRequest(request)
	if failureCode != "" {
		return workflowRunFailure(failureCode, failureSummary)
	}
	if normalizedRequest.DevFailureScenario != "" && !service.diagnosticsDevEnabled {
		return workflowRunFailure(WorkflowRunFailureInputInvalid, "Workflow diagnostics dev failure scenario is disabled.")
	}
	draftResult := service.draftReader(
		SavedWorkflowDraftContext{
			RequestContext:  runContext.RequestContext,
			RequestID:       runContext.RequestID,
			TenantRef:       runContext.TenantRef,
			WorkspaceID:     runContext.WorkspaceID,
			ApplicationID:   runContext.ApplicationID,
			ActorRef:        runContext.ActorRef,
			OwnerSubjectRef: runContext.ActorRef,
			ScopeGrants:     cloneStringSlice(runContext.ScopeGrants),
			AuditRef:        runContext.AuditRef,
		},
		ReadWorkflowDraftRequest{DraftID: normalizedRequest.DraftID},
	)
	if draftResult.FailureCode != "" || draftResult.Draft == nil {
		return workflowRunFailureForDraftRead(draftResult.FailureCode)
	}
	draft := *draftResult.Draft
	plan, failureCode, failureSummary := buildWorkflowExecutionPlan(draft, normalizedRequest.ConditionValues)
	if failureCode != "" {
		return workflowRunFailure(failureCode, failureSummary)
	}
	runID, err := service.newRunID()
	if err != nil {
		return workflowRunFailure(WorkflowRunFailureStoreUnavailable, "Unable to allocate a workflow run identifier.")
	}

	requestContext := runContext.RequestContext
	if requestContext == nil {
		requestContext = context.Background()
	}
	maxRuntime := service.maxRuntime
	if maxRuntime <= 0 {
		maxRuntime = workflowExecutorDefaultMaxRuntime
	}
	executionContext, cancel := context.WithTimeout(requestContext, maxRuntime)
	defer cancel()

	selection := resolveWorkflowRunSelection(service, executionContext, normalizedRequest.Model)
	record := newWorkflowRunRecord(runContext, normalizedRequest, draft, plan, selection, runID)
	if normalizedRequest.DevFailureScenario == WorkflowRunDevFailureStoreUnavailable {
		return workflowRunFailure(WorkflowRunFailureStoreUnavailable, "Workflow run record storage is unavailable.")
	}
	if normalizedRequest.DevFailureScenario == WorkflowRunDevFailureStaleRunning || normalizedRequest.DevFailureScenario == WorkflowRunDevFailureTerminalConflict {
		record.StartedAt = workflowRunTimestamp(time.Now().Add(-workflowExecutorDefaultMaxRuntime - time.Second))
	}
	if err := service.store.UpsertRun(runContext, &record); err != nil {
		return workflowRunFailure(WorkflowRunFailureStoreUnavailable, "Workflow run record storage is unavailable.")
	}
	if normalizedRequest.DevFailureScenario == WorkflowRunDevFailureStaleRunning {
		return WorkflowRunResult{Record: workflowRunRecordPointer(record)}
	}
	if normalizedRequest.DevFailureScenario == WorkflowRunDevFailureTerminalConflict {
		setWorkflowRunFailureDiagnostic(&record, WorkflowRunFailureStoreUnavailable, "", WorkflowRunGatewayFailureNone)
		return WorkflowRunResult{
			Record: workflowRunRecordPointer(record), FailureCode: WorkflowRunFailureStoreUnavailable,
			FailureSummary: "Workflow run terminal state could not be stored.",
		}
	}
	return service.executePlan(executionContext, runContext, normalizedRequest, draft, plan, selection, record)
}

func (service workflowExecutorService) ReadRun(
	runContext WorkflowRunContext,
	runID string,
) WorkflowRunResult {
	normalizedRunID := strings.TrimSpace(runID)
	if normalizedRunID == "" {
		return workflowRunFailure(WorkflowRunFailureInputInvalid, "Workflow run id is required.")
	}
	record, found, err := service.store.ReadRun(runContext, normalizedRunID)
	if err != nil {
		return workflowRunFailure(WorkflowRunFailureStoreUnavailable, "Workflow run record storage is unavailable.")
	}
	if !found {
		return workflowRunFailure(WorkflowRunFailureRecordNotFound, "Workflow run record was not found in the current scope.")
	}
	return WorkflowRunResult{Record: workflowRunRecordPointer(record)}
}

func (service workflowExecutorService) executePlan(
	executionContext context.Context,
	runContext WorkflowRunContext,
	request WorkflowRunRequest,
	draft SavedWorkflowDraft,
	plan workflowExecutionPlan,
	selection northboundSelection,
	record WorkflowRunRecord,
) WorkflowRunResult {
	edgeActive := make(map[string]bool, len(draft.Edges))
	nodeOutputs := make(map[string]string, len(draft.Nodes))
	nodeRecordIndex := make(map[string]int, len(record.Nodes))
	for index, nodeRecord := range record.Nodes {
		nodeRecordIndex[nodeRecord.NodeID] = index
	}

	for _, nodeID := range plan.order {
		if err := executionContext.Err(); err != nil {
			return service.finishFailedRun(runContext, record, WorkflowRunFailureCanceled, "Workflow run was canceled or exceeded its execution deadline.", true)
		}
		node := plan.nodes[nodeID]
		recordIndex := nodeRecordIndex[nodeID]
		active := nodeID == plan.rootNodeID || workflowNodeHasActiveInput(plan.incoming[nodeID], edgeActive)
		if !active {
			record.Nodes[recordIndex].Status = WorkflowRunNodeStatusSkipped
			record.Nodes[recordIndex].CompletedAt = workflowRunTimestamp(time.Now())
			if record.Diagnostic != nil {
				record.Diagnostic.LastCompletedNodeID = nodeID
				record.Diagnostic.ObservedAt = workflowRunTimestamp(time.Now())
			}
			for _, edge := range plan.outgoing[nodeID] {
				edgeActive[edge.EdgeID] = false
			}
			if err := service.store.UpsertRun(runContext, &record); err != nil {
				return service.finishFailedRun(runContext, record, WorkflowRunFailureStoreUnavailable, "Workflow run record storage is unavailable.", false)
			}
			continue
		}

		nodeStartedAt := time.Now().UTC()
		record.Nodes[recordIndex].Status = WorkflowRunNodeStatusRunning
		record.Nodes[recordIndex].StartedAt = workflowRunTimestamp(nodeStartedAt)
		if err := service.store.UpsertRun(runContext, &record); err != nil {
			return service.finishFailedRun(runContext, record, WorkflowRunFailureStoreUnavailable, "Workflow run record storage is unavailable.", false)
		}

		output, failureCode, failureSummary := service.executeNode(
			executionContext,
			request,
			draft,
			plan,
			node,
			edgeActive,
			nodeOutputs,
			selection,
			&record,
		)
		nodeCompletedAt := time.Now().UTC()
		record.Nodes[recordIndex].CompletedAt = workflowRunTimestamp(nodeCompletedAt)
		record.Nodes[recordIndex].DurationMS = nodeCompletedAt.Sub(nodeStartedAt).Milliseconds()
		if failureCode != "" {
			record.Nodes[recordIndex].Status = WorkflowRunNodeStatusFailed
			record.Nodes[recordIndex].FailureCode = failureCode
			if record.Diagnostic != nil && record.Diagnostic.FailedNodeID == "" {
				setWorkflowRunFailureDiagnostic(&record, failureCode, nodeID, WorkflowRunGatewayFailureNone)
			}
			canceled := failureCode == WorkflowRunFailureCanceled
			return service.finishFailedRun(runContext, record, failureCode, failureSummary, canceled)
		}

		nodeOutputs[nodeID] = output
		record.Nodes[recordIndex].Status = WorkflowRunNodeStatusSucceeded
		record.Nodes[recordIndex].OutputPreview = workflowRunNodeOutputPreview(node.NodeType, output)
		if record.Diagnostic != nil {
			record.Diagnostic.LastCompletedNodeID = nodeID
			record.Diagnostic.ObservedAt = workflowRunTimestamp(time.Now())
		}
		if node.NodeType == "condition" {
			activateWorkflowConditionEdges(request.ConditionValues[nodeID], plan.outgoing[nodeID], edgeActive)
		} else {
			for _, edge := range plan.outgoing[nodeID] {
				edgeActive[edge.EdgeID] = true
			}
		}
		if err := service.store.UpsertRun(runContext, &record); err != nil {
			return service.finishFailedRun(runContext, record, WorkflowRunFailureStoreUnavailable, "Workflow run record storage is unavailable.", false)
		}
	}

	finalOutput, found := nodeOutputs[plan.outputNodeID]
	if !found || strings.TrimSpace(finalOutput) == "" {
		return service.finishFailedRun(runContext, record, WorkflowRunFailureOutputUnavailable, "No active workflow path produced an output node result.", false)
	}
	record.Status = WorkflowRunStatusSucceeded
	record.Output = finalOutput
	record.CompletedAt = workflowRunTimestamp(time.Now())
	record.FailureCode = ""
	record.FailureSummary = ""
	if record.Diagnostic != nil {
		record.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
		record.Diagnostic.ObservedAt = workflowRunTimestamp(time.Now())
	}
	if err := service.store.UpsertRun(runContext, &record); err != nil {
		if record.Diagnostic != nil {
			record.Diagnostic.TerminalWriteState = WorkflowRunTerminalWritePending
			setWorkflowRunFailureDiagnostic(&record, WorkflowRunFailureStoreUnavailable, "", WorkflowRunGatewayFailureNone)
		}
		return WorkflowRunResult{Record: workflowRunRecordPointer(record), FailureCode: WorkflowRunFailureStoreUnavailable, FailureSummary: "Workflow run completed but its terminal record could not be stored."}
	}
	return WorkflowRunResult{Record: workflowRunRecordPointer(record)}
}

func (service workflowExecutorService) executeNode(
	executionContext context.Context,
	request WorkflowRunRequest,
	draft SavedWorkflowDraft,
	plan workflowExecutionPlan,
	node SavedWorkflowDraftNode,
	edgeActive map[string]bool,
	nodeOutputs map[string]string,
	selection northboundSelection,
	record *WorkflowRunRecord,
) (string, WorkflowRunFailureCode, string) {
	if failureCode, summary, category, matched := workflowRunInjectedNodeFailure(request.DevFailureScenario, node.NodeType); matched {
		if node.NodeType == "llm" {
			record.SideEffects.ProviderCalls++
		}
		setWorkflowRunFailureDiagnostic(record, failureCode, node.NodeID, category)
		return "", failureCode, summary
	}
	switch node.NodeType {
	case "prompt":
		packet := buildWorkflowPromptPacket(node, request.InputText)
		if len([]byte(packet)) > workflowExecutorMaxPacketBytes {
			return "", WorkflowRunFailureBudgetExceeded, "Prompt node packet exceeded the workflow execution budget."
		}
		return packet, "", ""
	case "condition":
		if request.ConditionValues[node.NodeID] {
			return "true", "", ""
		}
		return "false", "", ""
	case "llm":
		packet, failureCode, failureSummary := buildWorkflowNodeInputPacket(
			node,
			plan.incoming[node.NodeID],
			edgeActive,
			nodeOutputs,
			true,
		)
		if failureCode != "" {
			return "", failureCode, failureSummary
		}
		canonicalRequest, err := buildNorthboundCanonicalRequest(northboundCanonicalRequestOptions{
			requestID:  record.RunID + "-" + node.NodeID,
			route:      workflowExecutorRoute,
			protocol:   workflowExecutorProtocol,
			locale:     "zh-CN",
			promptText: packet,
			northboundFields: map[string]any{
				"request_kind":           workflowExecutorProtocol,
				"workflow_run_id":        record.RunID,
				"workflow_draft_id":      draft.DraftID,
				"workflow_draft_version": draft.DraftVersion,
				"workflow_node_id":       node.NodeID,
				"allow_tool_calls":       false,
				"allow_retrieval":        false,
				"writes_business_truth":  false,
			},
		})
		if err != nil {
			return "", WorkflowRunFailureGatewayFailed, "Workflow model request could not be assembled."
		}
		temperature := service.defaultTemperature
		if request.Temperature != nil {
			temperature = *request.Temperature
		}
		record.SideEffects.ProviderCalls++
		envelope, err := service.bridge.HandleEnvelope(
			executionContext,
			canonicalRequest,
			workflowRunBridgeEnvelopeOptions(service, selection, temperature),
		)
		if err != nil {
			if errors.Is(executionContext.Err(), context.Canceled) || errors.Is(executionContext.Err(), context.DeadlineExceeded) {
				return "", WorkflowRunFailureCanceled, "Workflow run was canceled or exceeded its execution deadline."
			}
			setWorkflowRunFailureDiagnostic(record, WorkflowRunFailureGatewayFailed, node.NodeID, workflowRunGatewayCategory(err))
			return "", WorkflowRunFailureGatewayFailed, "Gateway could not complete the workflow model node."
		}
		if !strings.EqualFold(strings.TrimSpace(envelope.Status), "ok") || envelope.Error != nil {
			setWorkflowRunFailureDiagnostic(record, WorkflowRunFailureGatewayFailed, node.NodeID, WorkflowRunGatewayFailureProviderFailed)
			return "", WorkflowRunFailureGatewayFailed, "Gateway returned a failed workflow model node envelope."
		}
		output := strings.TrimSpace(buildNorthboundResponseContent(envelope))
		if output == "" {
			setWorkflowRunFailureDiagnostic(record, WorkflowRunFailureOutputUnavailable, node.NodeID, WorkflowRunGatewayFailureOutputUnavailable)
			return "", WorkflowRunFailureOutputUnavailable, "Gateway returned no reviewable workflow model output."
		}
		if len([]byte(output)) > workflowExecutorMaxOutputBytes {
			return "", WorkflowRunFailureBudgetExceeded, "Workflow model output exceeded the execution budget."
		}
		return output, "", ""
	case "output":
		packet, failureCode, failureSummary := buildWorkflowNodeInputPacket(
			node,
			plan.incoming[node.NodeID],
			edgeActive,
			nodeOutputs,
			false,
		)
		if failureCode != "" {
			return "", failureCode, failureSummary
		}
		if len([]byte(packet)) > workflowExecutorMaxOutputBytes {
			return "", WorkflowRunFailureBudgetExceeded, "Workflow output exceeded the execution budget."
		}
		return packet, "", ""
	default:
		return "", WorkflowRunFailureDraftNotEligible, "Workflow node type is not allowed by executor v0."
	}
}

func (service workflowExecutorService) finishFailedRun(
	runContext WorkflowRunContext,
	record WorkflowRunRecord,
	failureCode WorkflowRunFailureCode,
	failureSummary string,
	canceled bool,
) WorkflowRunResult {
	if canceled {
		record.Status = WorkflowRunStatusCanceled
	} else {
		record.Status = WorkflowRunStatusFailed
	}
	record.FailureCode = failureCode
	record.FailureSummary = failureSummary
	record.CompletedAt = workflowRunTimestamp(time.Now())
	if record.Diagnostic != nil {
		if record.Diagnostic.FailureBoundary == "" {
			setWorkflowRunFailureDiagnostic(&record, failureCode, record.Diagnostic.FailedNodeID, record.Diagnostic.GatewayFailureCategory)
		}
		record.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
		record.Diagnostic.ObservedAt = workflowRunTimestamp(time.Now())
	}
	if err := service.store.UpsertRun(runContext, &record); err != nil {
		if record.Diagnostic != nil {
			record.Diagnostic.TerminalWriteState = WorkflowRunTerminalWritePending
			setWorkflowRunFailureDiagnostic(&record, WorkflowRunFailureStoreUnavailable, record.Diagnostic.FailedNodeID, WorkflowRunGatewayFailureNone)
		}
		return WorkflowRunResult{
			Record:         workflowRunRecordPointer(record),
			FailureCode:    WorkflowRunFailureStoreUnavailable,
			FailureSummary: "Workflow run failed and its terminal record could not be stored.",
		}
	}
	return WorkflowRunResult{
		Record:         workflowRunRecordPointer(record),
		FailureCode:    failureCode,
		FailureSummary: failureSummary,
	}
}

func normalizeWorkflowRunRequest(request WorkflowRunRequest) (WorkflowRunRequest, WorkflowRunFailureCode, string) {
	normalized := WorkflowRunRequest{
		DraftID:            strings.TrimSpace(request.DraftID),
		InputText:          strings.TrimSpace(request.InputText),
		ConditionValues:    make(map[string]bool, len(request.ConditionValues)),
		Model:              strings.TrimSpace(request.Model),
		Temperature:        request.Temperature,
		DevFailureScenario: WorkflowRunDevFailureScenario(strings.TrimSpace(string(request.DevFailureScenario))),
	}
	if normalized.DraftID == "" || normalized.InputText == "" {
		return WorkflowRunRequest{}, WorkflowRunFailureInputInvalid, "Workflow draft id and input text are required."
	}
	if len([]byte(normalized.InputText)) > workflowExecutorMaxInputBytes {
		return WorkflowRunRequest{}, WorkflowRunFailureBudgetExceeded, "Workflow run input exceeded the execution budget."
	}
	if len([]rune(normalized.Model)) > workflowExecutorMaxModelChars {
		return WorkflowRunRequest{}, WorkflowRunFailureInputInvalid, "Workflow run model selector is too long."
	}
	if len(request.ConditionValues) > workflowExecutorMaxConditions {
		return WorkflowRunRequest{}, WorkflowRunFailureBudgetExceeded, "Workflow condition input count exceeded the execution budget."
	}
	if normalized.DevFailureScenario != "" && !validWorkflowRunDevFailureScenario(normalized.DevFailureScenario) {
		return WorkflowRunRequest{}, WorkflowRunFailureInputInvalid, "Workflow diagnostics dev failure scenario is invalid."
	}
	for nodeID, value := range request.ConditionValues {
		normalizedNodeID := strings.TrimSpace(nodeID)
		if normalizedNodeID == "" || len([]rune(normalizedNodeID)) > 160 {
			return WorkflowRunRequest{}, WorkflowRunFailureInputInvalid, "Workflow condition node id is invalid."
		}
		normalized.ConditionValues[normalizedNodeID] = value
	}
	if normalized.Temperature != nil {
		value := *normalized.Temperature
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 2 {
			return WorkflowRunRequest{}, WorkflowRunFailureInputInvalid, "Workflow run temperature must be between 0 and 2."
		}
	}
	return normalized, "", ""
}

func buildWorkflowExecutionPlan(
	draft SavedWorkflowDraft,
	conditionValues map[string]bool,
) (workflowExecutionPlan, WorkflowRunFailureCode, string) {
	if draft.DraftVersion <= 0 {
		return workflowExecutionPlan{}, WorkflowRunFailureDraftVersionUnavailable, "Workflow draft does not have a persisted executable version."
	}
	if draft.SchemaVersion != savedWorkflowDraftSchemaVersion || !draft.ValidationSummary.ValidForReview ||
		draft.ValidationSummary.ValidationState != SavedWorkflowDraftStatusValidForReview {
		return workflowExecutionPlan{}, WorkflowRunFailureDraftNotEligible, "Workflow draft is not valid for executor v0 review."
	}
	if len(draft.Nodes) < 3 || len(draft.Nodes) > workflowExecutorMaxNodes ||
		len(draft.Edges) < 2 || len(draft.Edges) > workflowExecutorMaxEdges {
		return workflowExecutionPlan{}, WorkflowRunFailureBudgetExceeded, "Workflow graph size is outside the executor v0 budget."
	}
	for _, capability := range draft.RequestedCapabilities {
		if strings.TrimSpace(capability) != "" {
			return workflowExecutionPlan{}, WorkflowRunFailureDraftNotEligible, "Workflow draft requests a capability that executor v0 does not allow."
		}
	}

	plan := workflowExecutionPlan{
		nodes:    make(map[string]SavedWorkflowDraftNode, len(draft.Nodes)),
		incoming: make(map[string][]SavedWorkflowDraftEdge, len(draft.Nodes)),
		outgoing: make(map[string][]SavedWorkflowDraftEdge, len(draft.Nodes)),
	}
	nodePosition := make(map[string]int, len(draft.Nodes))
	promptCount := 0
	outputCount := 0
	llmCount := 0
	conditionNodeIDs := make(map[string]struct{})
	for index, node := range draft.Nodes {
		nodeID := strings.TrimSpace(node.NodeID)
		if nodeID == "" {
			return workflowExecutionPlan{}, WorkflowRunFailureGraphInvalid, "Workflow graph contains a node without an id."
		}
		if _, found := plan.nodes[nodeID]; found {
			return workflowExecutionPlan{}, WorkflowRunFailureGraphInvalid, "Workflow graph contains duplicate node ids."
		}
		node.NodeID = nodeID
		node.NodeType = strings.ToLower(strings.TrimSpace(node.NodeType))
		if !workflowExecutorAllowsNode(node) {
			return workflowExecutionPlan{}, WorkflowRunFailureDraftNotEligible, fmt.Sprintf("Workflow node %s is outside the executor v0 capability boundary.", nodeID)
		}
		switch node.NodeType {
		case "prompt":
			promptCount++
		case "llm":
			llmCount++
		case "condition":
			conditionNodeIDs[nodeID] = struct{}{}
		case "output":
			outputCount++
		}
		plan.nodes[nodeID] = node
		nodePosition[nodeID] = index
	}
	if promptCount != 1 || outputCount != 1 || llmCount < 1 {
		return workflowExecutionPlan{}, WorkflowRunFailureGraphInvalid, "Executor v0 requires exactly one prompt, one output, and at least one LLM node."
	}
	if llmCount > workflowExecutorMaxLLMCalls {
		return workflowExecutionPlan{}, WorkflowRunFailureBudgetExceeded, "Workflow LLM node count exceeded the execution budget."
	}
	if len(conditionValues) != len(conditionNodeIDs) {
		return workflowExecutionPlan{}, WorkflowRunFailureInputInvalid, "Every condition node requires exactly one explicit boolean input."
	}
	for nodeID := range conditionValues {
		if _, found := conditionNodeIDs[nodeID]; !found {
			return workflowExecutionPlan{}, WorkflowRunFailureInputInvalid, "Workflow run includes a condition value for an unknown node."
		}
	}

	edgeIDs := make(map[string]struct{}, len(draft.Edges))
	edgePairs := make(map[string]struct{}, len(draft.Edges))
	indegree := make(map[string]int, len(draft.Nodes))
	for nodeID := range plan.nodes {
		indegree[nodeID] = 0
	}
	for _, edge := range draft.Edges {
		edge.EdgeID = strings.TrimSpace(edge.EdgeID)
		edge.FromNodeID = strings.TrimSpace(edge.FromNodeID)
		edge.ToNodeID = strings.TrimSpace(edge.ToNodeID)
		if edge.EdgeID == "" || edge.FromNodeID == "" || edge.ToNodeID == "" || edge.FromNodeID == edge.ToNodeID {
			return workflowExecutionPlan{}, WorkflowRunFailureGraphInvalid, "Workflow graph contains an invalid edge."
		}
		if _, found := edgeIDs[edge.EdgeID]; found {
			return workflowExecutionPlan{}, WorkflowRunFailureGraphInvalid, "Workflow graph contains duplicate edge ids."
		}
		edgeIDs[edge.EdgeID] = struct{}{}
		if _, found := plan.nodes[edge.FromNodeID]; !found {
			return workflowExecutionPlan{}, WorkflowRunFailureGraphInvalid, "Workflow edge source does not exist."
		}
		if _, found := plan.nodes[edge.ToNodeID]; !found {
			return workflowExecutionPlan{}, WorkflowRunFailureGraphInvalid, "Workflow edge target does not exist."
		}
		pairKey := edge.FromNodeID + "\x00" + edge.ToNodeID
		if _, found := edgePairs[pairKey]; found {
			return workflowExecutionPlan{}, WorkflowRunFailureGraphInvalid, "Workflow graph contains duplicate node-to-node edges."
		}
		edgePairs[pairKey] = struct{}{}
		fromNode := plan.nodes[edge.FromNodeID]
		_, conditionRoute := workflowConditionEdgeRoute(edge.ConditionSummary)
		if fromNode.NodeType == "condition" && !conditionRoute {
			return workflowExecutionPlan{}, WorkflowRunFailureGraphInvalid, "Condition node edges must use when:true, when:false, or always."
		}
		if fromNode.NodeType != "condition" && conditionRoute {
			return workflowExecutionPlan{}, WorkflowRunFailureGraphInvalid, "Conditional routing can only originate from a condition node."
		}
		plan.outgoing[edge.FromNodeID] = append(plan.outgoing[edge.FromNodeID], edge)
		plan.incoming[edge.ToNodeID] = append(plan.incoming[edge.ToNodeID], edge)
		indegree[edge.ToNodeID]++
	}

	roots := make([]string, 0, 1)
	terminals := make([]string, 0, 1)
	for nodeID, node := range plan.nodes {
		if indegree[nodeID] == 0 {
			roots = append(roots, nodeID)
		}
		if len(plan.outgoing[nodeID]) == 0 {
			terminals = append(terminals, nodeID)
		}
		if node.NodeType == "condition" && len(plan.outgoing[nodeID]) == 0 {
			return workflowExecutionPlan{}, WorkflowRunFailureGraphInvalid, "Condition node must have an outgoing branch."
		}
	}
	if len(roots) != 1 || plan.nodes[roots[0]].NodeType != "prompt" ||
		len(terminals) != 1 || plan.nodes[terminals[0]].NodeType != "output" {
		return workflowExecutionPlan{}, WorkflowRunFailureGraphInvalid, "Workflow graph must have one prompt root and one output terminal."
	}
	plan.rootNodeID = roots[0]
	plan.outputNodeID = terminals[0]

	order := workflowStableTopologicalOrder(indegree, plan.outgoing, nodePosition)
	if len(order) != len(plan.nodes) {
		return workflowExecutionPlan{}, WorkflowRunFailureGraphInvalid, "Workflow graph must be acyclic."
	}
	plan.order = order
	if !workflowAllNodesReachable(plan.rootNodeID, plan.outgoing, len(plan.nodes)) ||
		!workflowAllNodesCanReachOutput(plan.outputNodeID, plan.incoming, len(plan.nodes)) {
		return workflowExecutionPlan{}, WorkflowRunFailureGraphInvalid, "Every workflow node must be on a path from prompt to output."
	}
	return plan, "", ""
}

func workflowExecutorAllowsNode(node SavedWorkflowDraftNode) bool {
	switch node.NodeType {
	case "prompt", "llm", "condition", "output":
	default:
		return false
	}
	riskLevel := strings.ToLower(strings.TrimSpace(node.RiskLevel))
	return (riskLevel == "" || riskLevel == "low") &&
		!node.RequiresConfirmation &&
		strings.TrimSpace(node.ToolRef) == "" &&
		strings.TrimSpace(node.RAGRef) == ""
}

func workflowStableTopologicalOrder(
	indegree map[string]int,
	outgoing map[string][]SavedWorkflowDraftEdge,
	nodePosition map[string]int,
) []string {
	remaining := make(map[string]int, len(indegree))
	ready := make([]string, 0)
	for nodeID, degree := range indegree {
		remaining[nodeID] = degree
		if degree == 0 {
			ready = append(ready, nodeID)
		}
	}
	sort.Slice(ready, func(left int, right int) bool { return nodePosition[ready[left]] < nodePosition[ready[right]] })
	order := make([]string, 0, len(indegree))
	for len(ready) > 0 {
		nodeID := ready[0]
		ready = ready[1:]
		order = append(order, nodeID)
		for _, edge := range outgoing[nodeID] {
			remaining[edge.ToNodeID]--
			if remaining[edge.ToNodeID] == 0 {
				ready = append(ready, edge.ToNodeID)
			}
		}
		sort.Slice(ready, func(left int, right int) bool { return nodePosition[ready[left]] < nodePosition[ready[right]] })
	}
	return order
}

func workflowAllNodesReachable(rootNodeID string, outgoing map[string][]SavedWorkflowDraftEdge, expected int) bool {
	visited := map[string]bool{rootNodeID: true}
	queue := []string{rootNodeID}
	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]
		for _, edge := range outgoing[nodeID] {
			if !visited[edge.ToNodeID] {
				visited[edge.ToNodeID] = true
				queue = append(queue, edge.ToNodeID)
			}
		}
	}
	return len(visited) == expected
}

func workflowAllNodesCanReachOutput(outputNodeID string, incoming map[string][]SavedWorkflowDraftEdge, expected int) bool {
	visited := map[string]bool{outputNodeID: true}
	queue := []string{outputNodeID}
	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]
		for _, edge := range incoming[nodeID] {
			if !visited[edge.FromNodeID] {
				visited[edge.FromNodeID] = true
				queue = append(queue, edge.FromNodeID)
			}
		}
	}
	return len(visited) == expected
}

func buildWorkflowPromptPacket(node SavedWorkflowDraftNode, inputText string) string {
	parts := make([]string, 0, 2)
	if instruction := strings.TrimSpace(node.InputSummary); instruction != "" {
		parts = append(parts, "Workflow instruction:\n"+instruction)
	}
	parts = append(parts, "User input:\n"+strings.TrimSpace(inputText))
	return strings.Join(parts, "\n\n")
}

func buildWorkflowNodeInputPacket(
	node SavedWorkflowDraftNode,
	incoming []SavedWorkflowDraftEdge,
	edgeActive map[string]bool,
	nodeOutputs map[string]string,
	includeInstruction bool,
) (string, WorkflowRunFailureCode, string) {
	parts := make([]string, 0, len(incoming)+1)
	activeOutputs := make([]string, 0, len(incoming))
	if instruction := strings.TrimSpace(node.InputSummary); includeInstruction && instruction != "" {
		parts = append(parts, "Workflow instruction:\n"+instruction)
	}
	for _, edge := range incoming {
		if !edgeActive[edge.EdgeID] {
			continue
		}
		output, found := nodeOutputs[edge.FromNodeID]
		if !found {
			return "", WorkflowRunFailureOutputUnavailable, "An active predecessor did not produce workflow node output."
		}
		activeOutputs = append(activeOutputs, output)
		parts = append(parts, "Predecessor "+edge.FromNodeID+":\n"+output)
	}
	if !includeInstruction && len(activeOutputs) == 1 {
		return activeOutputs[0], "", ""
	}
	if len(parts) == 0 || (len(parts) == 1 && strings.HasPrefix(parts[0], "Workflow instruction:")) {
		return "", WorkflowRunFailureOutputUnavailable, "Workflow node has no active predecessor output."
	}
	packet := strings.Join(parts, "\n\n")
	if len([]byte(packet)) > workflowExecutorMaxPacketBytes {
		return "", WorkflowRunFailureBudgetExceeded, "Workflow node input packet exceeded the execution budget."
	}
	return packet, "", ""
}

func workflowNodeHasActiveInput(incoming []SavedWorkflowDraftEdge, edgeActive map[string]bool) bool {
	for _, edge := range incoming {
		if edgeActive[edge.EdgeID] {
			return true
		}
	}
	return false
}

func workflowConditionEdgeRoute(summary string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(summary)) {
	case "when:true":
		return "true", true
	case "when:false":
		return "false", true
	case "always":
		return "always", true
	default:
		return "", false
	}
}

func activateWorkflowConditionEdges(
	value bool,
	edges []SavedWorkflowDraftEdge,
	edgeActive map[string]bool,
) {
	expected := "false"
	if value {
		expected = "true"
	}
	for _, edge := range edges {
		route, _ := workflowConditionEdgeRoute(edge.ConditionSummary)
		edgeActive[edge.EdgeID] = route == "always" || route == expected
	}
}

func workflowRunNodeOutputPreview(nodeType string, output string) string {
	switch nodeType {
	case "prompt":
		return "input packet accepted; raw input not retained"
	case "condition":
		return "condition evaluated; value not retained"
	default:
		return truncateWorkflowRunes(strings.TrimSpace(output), workflowExecutorNodePreviewRunes)
	}
}

func truncateWorkflowRunes(value string, maxRunes int) string {
	runes := []rune(value)
	if len(runes) <= maxRunes {
		return value
	}
	return string(runes[:maxRunes]) + "…"
}

func newWorkflowRunRecord(
	runContext WorkflowRunContext,
	request WorkflowRunRequest,
	draft SavedWorkflowDraft,
	plan workflowExecutionPlan,
	selection northboundSelection,
	runID string,
) WorkflowRunRecord {
	nodeRecords := make([]WorkflowRunNodeRecord, 0, len(plan.order))
	conditionNodeIDs := make([]string, 0, len(request.ConditionValues))
	for _, nodeID := range plan.order {
		node := plan.nodes[nodeID]
		predecessors := make([]string, 0, len(plan.incoming[nodeID]))
		for _, edge := range plan.incoming[nodeID] {
			predecessors = append(predecessors, edge.FromNodeID)
		}
		if node.NodeType == "condition" {
			conditionNodeIDs = append(conditionNodeIDs, nodeID)
		}
		nodeRecords = append(nodeRecords, WorkflowRunNodeRecord{
			NodeID:             node.NodeID,
			NodeType:           node.NodeType,
			Label:              strings.TrimSpace(node.Label),
			Status:             WorkflowRunNodeStatusPending,
			PredecessorNodeIDs: predecessors,
			ProviderRef:        strings.TrimSpace(node.ProviderRef),
		})
	}
	return WorkflowRunRecord{
		SchemaVersion:    workflowRunRecordSchemaVersion,
		RunID:            runID,
		DraftID:          draft.DraftID,
		DraftVersion:     draft.DraftVersion,
		WorkspaceID:      runContext.WorkspaceID,
		ApplicationID:    runContext.ApplicationID,
		Status:           WorkflowRunStatusRunning,
		StartedAt:        workflowRunTimestamp(time.Now()),
		InputBytes:       len([]byte(request.InputText)),
		ConditionNodeIDs: conditionNodeIDs,
		RequestedModel:   request.Model,
		SelectedProvider: strings.TrimSpace(selection.provider),
		SelectedProfile:  strings.TrimSpace(selection.providerProfile),
		SelectedModel:    strings.TrimSpace(selection.model),
		UpstreamModel:    strings.TrimSpace(selection.upstreamModel),
		SelectionSource:  strings.TrimSpace(selection.source),
		Nodes:            nodeRecords,
		RequestID:        runContext.RequestID,
		AuditRef:         runContext.AuditRef,
		ActorRef:         runContext.ActorRef,
		Diagnostic:       newWorkflowRunDiagnostic(),
	}
}

func resolveWorkflowRunSelection(
	service workflowExecutorService,
	ctx context.Context,
	requestedModel string,
) northboundSelection {
	if service.resolveSelection != nil {
		return service.resolveSelection(ctx, requestedModel)
	}
	server := &Server{bridge: service.bridge}
	return server.resolveNorthboundSelection(ctx, requestedModel, nil)
}

func workflowRunBridgeEnvelopeOptions(
	service workflowExecutorService,
	selection northboundSelection,
	temperature float64,
) bridge.EnvelopeOptions {
	if service.envelopeOptions != nil {
		return service.envelopeOptions(selection, temperature)
	}
	server := &Server{bridge: service.bridge}
	return server.buildBridgeEnvelopeOptions(selection, temperature)
}

func workflowRunTimestamp(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func newWorkflowRunID() (string, error) {
	randomBytes := make([]byte, 12)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return "run_" + hex.EncodeToString(randomBytes), nil
}

func workflowRunFailureForDraftRead(code SavedWorkflowDraftFailureCode) WorkflowRunResult {
	switch code {
	case SavedWorkflowDraftFailureNotFound:
		return workflowRunFailure(WorkflowRunFailureDraftNotFound, "Saved workflow draft was not found in the current scope.")
	case SavedWorkflowDraftFailureScopeDenied,
		SavedWorkflowDraftFailureAuthContextMismatch,
		SavedWorkflowDraftFailureScopeGrantMissing,
		SavedWorkflowDraftFailureWorkspaceMembershipDenied,
		SavedWorkflowDraftFailureApplicationScopeDenied,
		SavedWorkflowDraftFailureOwnerScopeDenied:
		return workflowRunFailure(WorkflowRunFailureScopeDenied, "Workflow run scope does not allow reading the saved draft.")
	case SavedWorkflowDraftFailureStoreUnavailable,
		SavedWorkflowDraftFailureRepositoryStoreDisabled,
		SavedWorkflowDraftFailureInvalidStoreMode:
		return workflowRunFailure(WorkflowRunFailureStoreUnavailable, "Saved workflow draft storage is unavailable for execution.")
	default:
		return workflowRunFailure(WorkflowRunFailureDraftNotEligible, "Saved workflow draft cannot enter executor v0.")
	}
}

func workflowRunFailure(code WorkflowRunFailureCode, summary string) WorkflowRunResult {
	return WorkflowRunResult{FailureCode: code, FailureSummary: summary}
}

func workflowRunRecordPointer(record WorkflowRunRecord) *WorkflowRunRecord {
	cloned := cloneWorkflowRunRecord(record)
	return &cloned
}
