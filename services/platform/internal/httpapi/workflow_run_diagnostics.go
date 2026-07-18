package httpapi

import (
	"strings"
	"time"

	"radishmind.local/services/platform/internal/bridge"
)

type WorkflowRunFailureBoundary string

const (
	WorkflowRunFailureBoundaryDraftRead         WorkflowRunFailureBoundary = "draft_read"
	WorkflowRunFailureBoundaryExecutor          WorkflowRunFailureBoundary = "executor"
	WorkflowRunFailureBoundaryGateway           WorkflowRunFailureBoundary = "gateway"
	WorkflowRunFailureBoundaryProvider          WorkflowRunFailureBoundary = "provider"
	WorkflowRunFailureBoundaryRunStore          WorkflowRunFailureBoundary = "run_store"
	WorkflowRunFailureBoundaryRequest           WorkflowRunFailureBoundary = "request"
	WorkflowRunFailureBoundaryToolPolicy        WorkflowRunFailureBoundary = "tool_policy"
	WorkflowRunFailureBoundaryToolConfirmation  WorkflowRunFailureBoundary = "tool_confirmation"
	WorkflowRunFailureBoundaryToolTransport     WorkflowRunFailureBoundary = "tool_transport"
	WorkflowRunFailureBoundaryToolResponse      WorkflowRunFailureBoundary = "tool_response"
	WorkflowRunFailureBoundaryToolStore         WorkflowRunFailureBoundary = "tool_store"
	WorkflowRunFailureBoundaryRetrievalPolicy   WorkflowRunFailureBoundary = "retrieval_policy"
	WorkflowRunFailureBoundaryRetrievalStore    WorkflowRunFailureBoundary = "retrieval_store"
	WorkflowRunFailureBoundaryRetrievalRank     WorkflowRunFailureBoundary = "retrieval_rank"
	WorkflowRunFailureBoundaryRetrievalContext  WorkflowRunFailureBoundary = "retrieval_context"
	WorkflowRunFailureBoundaryRetrievalCitation WorkflowRunFailureBoundary = "retrieval_citation"
	WorkflowRunFailureBoundaryProviderSelection WorkflowRunFailureBoundary = "provider_selection"
	WorkflowRunFailureBoundaryProviderCall      WorkflowRunFailureBoundary = "provider_call"
)

type WorkflowRunGatewayFailureCategory string

const (
	WorkflowRunGatewayFailureNone              WorkflowRunGatewayFailureCategory = "none"
	WorkflowRunGatewayFailureQueueFull         WorkflowRunGatewayFailureCategory = "queue_full"
	WorkflowRunGatewayFailureTimeout           WorkflowRunGatewayFailureCategory = "timeout"
	WorkflowRunGatewayFailureCanceled          WorkflowRunGatewayFailureCategory = "canceled"
	WorkflowRunGatewayFailureWorkerCrash       WorkflowRunGatewayFailureCategory = "worker_crash"
	WorkflowRunGatewayFailureProtocol          WorkflowRunGatewayFailureCategory = "protocol"
	WorkflowRunGatewayFailureProviderFailed    WorkflowRunGatewayFailureCategory = "provider_failed"
	WorkflowRunGatewayFailureOutputUnavailable WorkflowRunGatewayFailureCategory = "output_unavailable"
	WorkflowRunGatewayFailureUnavailable       WorkflowRunGatewayFailureCategory = "unavailable"
)

type WorkflowRunReviewAction string

const (
	WorkflowRunReviewDraft                 WorkflowRunReviewAction = "review_draft"
	WorkflowRunReviewGatewayCapacity       WorkflowRunReviewAction = "check_gateway_capacity"
	WorkflowRunReviewProviderConfiguration WorkflowRunReviewAction = "check_provider_configuration"
	WorkflowRunReviewRunStore              WorkflowRunReviewAction = "check_run_store"
	WorkflowRunReviewStartNewRun           WorkflowRunReviewAction = "start_new_run"
	WorkflowRunReviewToolPolicy            WorkflowRunReviewAction = "check_tool_policy"
	WorkflowRunReviewToolOutcome           WorkflowRunReviewAction = "review_tool_outcome"
	WorkflowRunReviewRetrievalEvidence     WorkflowRunReviewAction = "review_retrieval_evidence"
)

type WorkflowRunTerminalWriteState string

const (
	WorkflowRunTerminalWritePending WorkflowRunTerminalWriteState = "pending"
	WorkflowRunTerminalWriteStored  WorkflowRunTerminalWriteState = "stored"
)

type WorkflowRunDiagnostic struct {
	FailureBoundary          WorkflowRunFailureBoundary        `json:"failure_boundary"`
	FailureStage             string                            `json:"failure_stage"`
	FailedNodeID             string                            `json:"failed_node_id"`
	LastCompletedNodeID      string                            `json:"last_completed_node_id"`
	TerminalWriteState       WorkflowRunTerminalWriteState     `json:"terminal_write_state"`
	GatewayFailureCategory   WorkflowRunGatewayFailureCategory `json:"gateway_failure_category"`
	ToolFailureCategory      WorkflowHTTPToolFailureCategory   `json:"tool_failure_category,omitempty"`
	RetrievalFailureCategory string                            `json:"retrieval_failure_category,omitempty"`
	Summary                  string                            `json:"summary"`
	RecommendedReviewAction  WorkflowRunReviewAction           `json:"recommended_review_action"`
	ObservedAt               string                            `json:"observed_at"`
}

type WorkflowRunDevFailureScenario string

const (
	WorkflowRunDevFailureGatewayTimeout    WorkflowRunDevFailureScenario = "gateway_timeout"
	WorkflowRunDevFailureGatewayQueueFull  WorkflowRunDevFailureScenario = "gateway_queue_full"
	WorkflowRunDevFailureGatewayCrash      WorkflowRunDevFailureScenario = "gateway_worker_crash"
	WorkflowRunDevFailureGatewayProtocol   WorkflowRunDevFailureScenario = "gateway_protocol_failure"
	WorkflowRunDevFailureProviderFailed    WorkflowRunDevFailureScenario = "provider_failed"
	WorkflowRunDevFailureOutputUnavailable WorkflowRunDevFailureScenario = "output_unavailable"
	WorkflowRunDevFailureRequestCanceled   WorkflowRunDevFailureScenario = "request_canceled"
	WorkflowRunDevFailureStoreUnavailable  WorkflowRunDevFailureScenario = "run_store_unavailable"
	WorkflowRunDevFailureTerminalConflict  WorkflowRunDevFailureScenario = "terminal_write_conflict"
	WorkflowRunDevFailureBudgetExceeded    WorkflowRunDevFailureScenario = "budget_exceeded"
	WorkflowRunDevFailureStaleRunning      WorkflowRunDevFailureScenario = "stale_running"
)

func validWorkflowRunDevFailureScenario(value WorkflowRunDevFailureScenario) bool {
	switch value {
	case WorkflowRunDevFailureGatewayTimeout, WorkflowRunDevFailureGatewayQueueFull,
		WorkflowRunDevFailureGatewayCrash, WorkflowRunDevFailureGatewayProtocol,
		WorkflowRunDevFailureProviderFailed, WorkflowRunDevFailureOutputUnavailable,
		WorkflowRunDevFailureRequestCanceled, WorkflowRunDevFailureStoreUnavailable,
		WorkflowRunDevFailureTerminalConflict, WorkflowRunDevFailureBudgetExceeded,
		WorkflowRunDevFailureStaleRunning:
		return true
	default:
		return false
	}
}

func newWorkflowRunDiagnostic() *WorkflowRunDiagnostic {
	return &WorkflowRunDiagnostic{
		TerminalWriteState:     WorkflowRunTerminalWritePending,
		GatewayFailureCategory: WorkflowRunGatewayFailureNone,
		ObservedAt:             workflowRunTimestamp(time.Now()),
	}
}

func setWorkflowRunFailureDiagnostic(record *WorkflowRunRecord, code WorkflowRunFailureCode, nodeID string, category WorkflowRunGatewayFailureCategory) {
	if record == nil || record.Diagnostic == nil {
		return
	}
	boundary, stage, action := workflowRunDiagnosticClassification(code, category)
	record.Diagnostic.FailureBoundary = boundary
	record.Diagnostic.FailureStage = stage
	record.Diagnostic.FailedNodeID = strings.TrimSpace(nodeID)
	record.Diagnostic.GatewayFailureCategory = category
	record.Diagnostic.Summary = workflowRunDiagnosticSummary(code, category)
	record.Diagnostic.RecommendedReviewAction = action
	record.Diagnostic.ObservedAt = workflowRunTimestamp(time.Now())
}

func workflowRunDiagnosticClassification(code WorkflowRunFailureCode, category WorkflowRunGatewayFailureCategory) (WorkflowRunFailureBoundary, string, WorkflowRunReviewAction) {
	switch code {
	case WorkflowRunFailureStoreUnavailable, WorkflowRunFailureStoreContractMismatch:
		return WorkflowRunFailureBoundaryRunStore, "record_write", WorkflowRunReviewRunStore
	case WorkflowRunFailureToolStore:
		return WorkflowRunFailureBoundaryToolStore, "tool_state_write", WorkflowRunReviewRunStore
	case WorkflowRunFailureToolPolicy:
		return WorkflowRunFailureBoundaryToolPolicy, "target_policy", WorkflowRunReviewToolPolicy
	case WorkflowRunFailureToolConfirmation:
		return WorkflowRunFailureBoundaryToolConfirmation, "execution_claim", WorkflowRunReviewToolPolicy
	case WorkflowRunFailureToolTransport, WorkflowRunFailureToolTimeout:
		return WorkflowRunFailureBoundaryToolTransport, "http_request", WorkflowRunReviewStartNewRun
	case WorkflowRunFailureToolResponseStatus, WorkflowRunFailureToolResponseTooLarge, WorkflowRunFailureToolResponseInvalid:
		return WorkflowRunFailureBoundaryToolResponse, "response_validation", WorkflowRunReviewToolPolicy
	case WorkflowRunFailureToolOutcomeUnknown:
		return WorkflowRunFailureBoundaryToolStore, "outcome_reconciliation", WorkflowRunReviewToolOutcome
	case WorkflowRunFailureCanceled:
		return WorkflowRunFailureBoundaryRequest, "execution", WorkflowRunReviewStartNewRun
	case WorkflowRunFailureGatewayFailed:
		if category == WorkflowRunGatewayFailureProviderFailed || category == WorkflowRunGatewayFailureOutputUnavailable {
			return WorkflowRunFailureBoundaryProvider, "model_node", WorkflowRunReviewProviderConfiguration
		}
		return WorkflowRunFailureBoundaryGateway, "model_node", WorkflowRunReviewGatewayCapacity
	case WorkflowRunFailureOutputUnavailable:
		return WorkflowRunFailureBoundaryProvider, "output", WorkflowRunReviewProviderConfiguration
	case WorkflowRunFailureDraftNotEligible, WorkflowRunFailureDraftNotFound, WorkflowRunFailureDraftVersionUnavailable:
		return WorkflowRunFailureBoundaryDraftRead, "draft_eligibility", WorkflowRunReviewDraft
	default:
		return WorkflowRunFailureBoundaryExecutor, "node_execution", WorkflowRunReviewDraft
	}
}

func workflowRunDiagnosticSummary(code WorkflowRunFailureCode, category WorkflowRunGatewayFailureCategory) string {
	if code == WorkflowRunFailureGatewayFailed {
		switch category {
		case WorkflowRunGatewayFailureQueueFull:
			return "Gateway capacity was unavailable for the workflow model node."
		case WorkflowRunGatewayFailureTimeout:
			return "Gateway timed out while executing the workflow model node."
		case WorkflowRunGatewayFailureWorkerCrash:
			return "Gateway worker exited while executing the workflow model node."
		case WorkflowRunGatewayFailureProtocol:
			return "Gateway protocol validation failed for the workflow model node."
		case WorkflowRunGatewayFailureProviderFailed:
			return "Provider returned a failed workflow model response."
		default:
			return "Gateway could not complete the workflow model node."
		}
	}
	switch code {
	case WorkflowRunFailureCanceled:
		return "Workflow execution was canceled or exceeded its deadline."
	case WorkflowRunFailureBudgetExceeded:
		return "Workflow execution exceeded a configured budget."
	case WorkflowRunFailureOutputUnavailable:
		return "Workflow execution produced no reviewable output."
	case WorkflowRunFailureStoreUnavailable, WorkflowRunFailureStoreContractMismatch:
		return "Workflow run record storage could not commit the required state."
	case WorkflowRunFailureToolPolicy:
		return "Workflow tool execution was rejected by the reviewed target policy."
	case WorkflowRunFailureToolConfirmation:
		return "Workflow tool confirmation could not be consumed exactly once."
	case WorkflowRunFailureToolTransport:
		return "Workflow tool transport failed before a valid response was available."
	case WorkflowRunFailureToolTimeout:
		return "Workflow tool execution exceeded its request deadline."
	case WorkflowRunFailureToolResponseStatus, WorkflowRunFailureToolResponseTooLarge, WorkflowRunFailureToolResponseInvalid:
		return "Workflow tool response did not satisfy the reviewed response policy."
	case WorkflowRunFailureToolStore:
		return "Workflow tool state could not commit the required transition."
	case WorkflowRunFailureToolOutcomeUnknown:
		return "Workflow tool execution outcome requires manual review."
	default:
		return "Workflow execution requires draft review."
	}
}

func workflowRunGatewayCategory(err error) WorkflowRunGatewayFailureCategory {
	switch bridge.ErrorCode(err) {
	case bridge.ErrorCodeWorkerQueueFull:
		return WorkflowRunGatewayFailureQueueFull
	case bridge.ErrorCodeWorkerTimeout:
		return WorkflowRunGatewayFailureTimeout
	case bridge.ErrorCodeWorkerCanceled:
		return WorkflowRunGatewayFailureCanceled
	case bridge.ErrorCodeWorkerExited:
		return WorkflowRunGatewayFailureWorkerCrash
	case bridge.ErrorCodeWorkerProtocol:
		return WorkflowRunGatewayFailureProtocol
	default:
		return WorkflowRunGatewayFailureUnavailable
	}
}

func validWorkflowRunFailureBoundary(value WorkflowRunFailureBoundary) bool {
	switch value {
	case WorkflowRunFailureBoundaryDraftRead, WorkflowRunFailureBoundaryExecutor,
		WorkflowRunFailureBoundaryGateway, WorkflowRunFailureBoundaryProvider,
		WorkflowRunFailureBoundaryRunStore, WorkflowRunFailureBoundaryRequest,
		WorkflowRunFailureBoundaryToolPolicy, WorkflowRunFailureBoundaryToolConfirmation,
		WorkflowRunFailureBoundaryToolTransport, WorkflowRunFailureBoundaryToolResponse,
		WorkflowRunFailureBoundaryToolStore, WorkflowRunFailureBoundaryRetrievalPolicy,
		WorkflowRunFailureBoundaryRetrievalStore, WorkflowRunFailureBoundaryRetrievalRank,
		WorkflowRunFailureBoundaryRetrievalContext, WorkflowRunFailureBoundaryRetrievalCitation,
		WorkflowRunFailureBoundaryProviderSelection, WorkflowRunFailureBoundaryProviderCall:
		return true
	default:
		return false
	}
}

func validWorkflowRunFailureCode(value WorkflowRunFailureCode) bool {
	switch value {
	case WorkflowRunFailureScopeDenied, WorkflowRunFailureDraftNotFound,
		WorkflowRunFailureDraftVersionUnavailable, WorkflowRunFailureDraftNotEligible,
		WorkflowRunFailureInputInvalid, WorkflowRunFailureGraphInvalid,
		WorkflowRunFailureBudgetExceeded, WorkflowRunFailureGatewayFailed,
		WorkflowRunFailureOutputUnavailable, WorkflowRunFailureCanceled,
		WorkflowRunFailureRecordNotFound, WorkflowRunFailureStoreUnavailable,
		WorkflowRunFailureStoreContractMismatch, WorkflowRunFailureFilterInvalid,
		WorkflowRunFailureCursorInvalid, WorkflowRunFailureStoreModeInvalid,
		WorkflowRunFailureStoreModeDisabled, WorkflowRunFailureComparisonInvalid,
		WorkflowRunFailureSideEffectUnsupported, WorkflowRunFailureToolPolicy,
		WorkflowRunFailureToolConfirmation, WorkflowRunFailureToolTransport,
		WorkflowRunFailureToolTimeout, WorkflowRunFailureToolResponseStatus,
		WorkflowRunFailureToolResponseTooLarge, WorkflowRunFailureToolResponseInvalid,
		WorkflowRunFailureToolStore, WorkflowRunFailureToolOutcomeUnknown,
		WorkflowRunFailureRetrievalUnsupported:
		return true
	default:
		return false
	}
}

func validWorkflowRunDiagnostic(diagnostic *WorkflowRunDiagnostic, terminal bool) bool {
	if diagnostic == nil || len([]rune(diagnostic.Summary)) > 256 || strings.Contains(diagnostic.Summary, "://") {
		return false
	}
	if !validWorkflowRunGatewayFailureCategory(diagnostic.GatewayFailureCategory) ||
		(diagnostic.ToolFailureCategory != "" && !validWorkflowHTTPToolFailureCategory(diagnostic.ToolFailureCategory)) ||
		(diagnostic.RecommendedReviewAction != "" && !validWorkflowRunReviewAction(diagnostic.RecommendedReviewAction)) ||
		len([]rune(diagnostic.FailureStage)) > 64 || len([]rune(diagnostic.FailedNodeID)) > 160 ||
		len([]rune(diagnostic.LastCompletedNodeID)) > 160 || strings.Contains(diagnostic.FailureStage, "://") ||
		strings.Contains(diagnostic.FailedNodeID, "://") || strings.Contains(diagnostic.LastCompletedNodeID, "://") {
		return false
	}
	if terminal && diagnostic.TerminalWriteState != WorkflowRunTerminalWriteStored {
		return false
	}
	if !terminal && diagnostic.TerminalWriteState != WorkflowRunTerminalWritePending {
		return false
	}
	if diagnostic.FailureBoundary != "" && !validWorkflowRunFailureBoundary(diagnostic.FailureBoundary) {
		return false
	}
	if diagnostic.ObservedAt == "" {
		return false
	}
	_, err := time.Parse(time.RFC3339Nano, diagnostic.ObservedAt)
	return err == nil
}

func validWorkflowRunGatewayFailureCategory(value WorkflowRunGatewayFailureCategory) bool {
	switch value {
	case WorkflowRunGatewayFailureNone, WorkflowRunGatewayFailureQueueFull,
		WorkflowRunGatewayFailureTimeout, WorkflowRunGatewayFailureCanceled,
		WorkflowRunGatewayFailureWorkerCrash, WorkflowRunGatewayFailureProtocol,
		WorkflowRunGatewayFailureProviderFailed, WorkflowRunGatewayFailureOutputUnavailable,
		WorkflowRunGatewayFailureUnavailable:
		return true
	default:
		return false
	}
}

func validWorkflowRunReviewAction(value WorkflowRunReviewAction) bool {
	switch value {
	case WorkflowRunReviewDraft, WorkflowRunReviewGatewayCapacity,
		WorkflowRunReviewProviderConfiguration, WorkflowRunReviewRunStore,
		WorkflowRunReviewStartNewRun, WorkflowRunReviewToolPolicy, WorkflowRunReviewToolOutcome,
		WorkflowRunReviewRetrievalEvidence:
		return true
	default:
		return false
	}
}

func workflowRunInjectedNodeFailure(scenario WorkflowRunDevFailureScenario, nodeType string) (WorkflowRunFailureCode, string, WorkflowRunGatewayFailureCategory, bool) {
	if scenario == WorkflowRunDevFailureRequestCanceled {
		return WorkflowRunFailureCanceled, "Workflow run was canceled by the diagnostics scenario.", WorkflowRunGatewayFailureCanceled, true
	}
	if scenario == WorkflowRunDevFailureBudgetExceeded && nodeType == "prompt" {
		return WorkflowRunFailureBudgetExceeded, "Workflow prompt exceeded the diagnostics execution budget.", WorkflowRunGatewayFailureNone, true
	}
	if nodeType != "llm" {
		return "", "", "", false
	}
	switch scenario {
	case WorkflowRunDevFailureGatewayTimeout:
		return WorkflowRunFailureGatewayFailed, "Gateway timed out while executing the workflow model node.", WorkflowRunGatewayFailureTimeout, true
	case WorkflowRunDevFailureGatewayQueueFull:
		return WorkflowRunFailureGatewayFailed, "Gateway capacity was unavailable for the workflow model node.", WorkflowRunGatewayFailureQueueFull, true
	case WorkflowRunDevFailureGatewayCrash:
		return WorkflowRunFailureGatewayFailed, "Gateway worker exited while executing the workflow model node.", WorkflowRunGatewayFailureWorkerCrash, true
	case WorkflowRunDevFailureGatewayProtocol:
		return WorkflowRunFailureGatewayFailed, "Gateway protocol validation failed for the workflow model node.", WorkflowRunGatewayFailureProtocol, true
	case WorkflowRunDevFailureProviderFailed:
		return WorkflowRunFailureGatewayFailed, "Provider returned a failed workflow model response.", WorkflowRunGatewayFailureProviderFailed, true
	case WorkflowRunDevFailureOutputUnavailable:
		return WorkflowRunFailureOutputUnavailable, "Gateway returned no reviewable workflow model output.", WorkflowRunGatewayFailureOutputUnavailable, true
	default:
		return "", "", "", false
	}
}
