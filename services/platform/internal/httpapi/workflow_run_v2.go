package httpapi

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"
)

const workflowHTTPToolAttemptClaimedBudget = workflowExecutorDefaultMaxRuntime

var workflowHTTPToolAttemptIDPattern = regexp.MustCompile(`^wtea_[a-z0-9]{16,64}$`)
var workflowHTTPToolRunIDPattern = regexp.MustCompile(`^run_[a-z0-9]{16,64}$`)
var workflowHTTPToolProjectionKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

type WorkflowHTTPToolAttemptStatus string

const (
	WorkflowHTTPToolAttemptClaimed        WorkflowHTTPToolAttemptStatus = "claimed"
	WorkflowHTTPToolAttemptSucceeded      WorkflowHTTPToolAttemptStatus = "succeeded"
	WorkflowHTTPToolAttemptFailed         WorkflowHTTPToolAttemptStatus = "failed"
	WorkflowHTTPToolAttemptOutcomeUnknown WorkflowHTTPToolAttemptStatus = "outcome_unknown"
)

type WorkflowHTTPToolExecutionAttempt struct {
	AttemptID        string                        `json:"attempt_id"`
	NodeID           string                        `json:"node_id"`
	ToolID           string                        `json:"tool_id"`
	DefinitionDigest string                        `json:"definition_digest"`
	ProfileID        string                        `json:"profile_id"`
	ProfileDigest    string                        `json:"profile_digest"`
	ToolPlanDigest   string                        `json:"tool_plan_digest"`
	ConfirmationID   string                        `json:"confirmation_id"`
	Status           WorkflowHTTPToolAttemptStatus `json:"status"`
	ClaimedAt        string                        `json:"claimed_at"`
	CompletedAt      string                        `json:"completed_at"`
	HTTPStatusClass  string                        `json:"http_status_class"`
	ResponseBytes    int                           `json:"response_bytes"`
	DurationMS       int                           `json:"duration_ms"`
	OutputProjection map[string]any                `json:"output_projection"`
	FailureCode      WorkflowRunFailureCode        `json:"failure_code"`
}

type workflowRunRecordAlias WorkflowRunRecord

func (record WorkflowRunRecord) MarshalJSON() ([]byte, error) {
	if record.SchemaVersion != workflowRunRecordToolSchemaVersion {
		return json.Marshal(workflowRunRecordAlias(record))
	}
	nodes := make([]workflowRunNodeV2Document, 0, len(record.Nodes))
	for _, node := range record.Nodes {
		nodes = append(nodes, workflowRunNodeV2Document{
			NodeID: node.NodeID, NodeType: node.NodeType, Label: node.Label, Status: node.Status,
			StartedAt: nullableWorkflowRunString(node.StartedAt), CompletedAt: nullableWorkflowRunString(node.CompletedAt),
			DurationMS: node.DurationMS, PredecessorNodeIDs: node.PredecessorNodeIDs,
			ProviderRef: node.ProviderRef, OutputPreview: node.OutputPreview,
			FailureCode: nullableWorkflowRunFailureCode(node.FailureCode),
		})
	}
	return json.Marshal(workflowRunRecordV2Document{
		SchemaVersion: record.SchemaVersion, RecordVersion: record.RecordVersion, RunID: record.RunID,
		PlanID: record.PlanID, ConfirmationID: record.ConfirmationID, TenantRef: record.TenantRef,
		WorkspaceID: record.WorkspaceID, ApplicationID: record.ApplicationID,
		DraftID: record.DraftID, DraftVersion: record.DraftVersion, Status: record.Status,
		FailureCode: nullableWorkflowRunFailureCode(record.FailureCode), FailureSummary: record.FailureSummary,
		StartedAt: record.StartedAt, CompletedAt: nullableWorkflowRunString(record.CompletedAt), InputBytes: record.InputBytes,
		RequestedModel: record.RequestedModel, SelectedProvider: record.SelectedProvider,
		SelectedProfile: record.SelectedProfile, SelectedModel: record.SelectedModel,
		UpstreamModel: record.UpstreamModel, SelectionSource: record.SelectionSource,
		Nodes: nodes, ToolAttempt: workflowHTTPToolAttemptV2DocumentFrom(record.ToolAttempt), Output: record.Output,
		RequestID: record.RequestID, AuditRef: record.AuditRef, ActorRef: record.ActorRef,
		SideEffects: record.SideEffects, Diagnostic: workflowRunDiagnosticV2DocumentFrom(record.Diagnostic),
	})
}

type workflowRunRecordV2Document struct {
	SchemaVersion    string                             `json:"schema_version"`
	RecordVersion    int                                `json:"record_version"`
	RunID            string                             `json:"run_id"`
	PlanID           string                             `json:"plan_id"`
	ConfirmationID   string                             `json:"confirmation_id"`
	TenantRef        string                             `json:"tenant_ref"`
	WorkspaceID      string                             `json:"workspace_id"`
	ApplicationID    string                             `json:"application_id"`
	DraftID          string                             `json:"draft_id"`
	DraftVersion     int                                `json:"draft_version"`
	Status           WorkflowRunStatus                  `json:"status"`
	FailureCode      any                                `json:"failure_code"`
	FailureSummary   string                             `json:"failure_summary"`
	StartedAt        string                             `json:"started_at"`
	CompletedAt      any                                `json:"completed_at"`
	InputBytes       int                                `json:"input_bytes"`
	RequestedModel   string                             `json:"requested_model"`
	SelectedProvider string                             `json:"selected_provider"`
	SelectedProfile  string                             `json:"selected_profile"`
	SelectedModel    string                             `json:"selected_model"`
	UpstreamModel    string                             `json:"upstream_model"`
	SelectionSource  string                             `json:"selection_source"`
	Nodes            []workflowRunNodeV2Document        `json:"nodes"`
	ToolAttempt      *workflowHTTPToolAttemptV2Document `json:"tool_attempt"`
	Output           string                             `json:"output"`
	RequestID        string                             `json:"request_id"`
	AuditRef         string                             `json:"audit_ref"`
	ActorRef         string                             `json:"actor_ref"`
	SideEffects      WorkflowRunSideEffects             `json:"side_effects"`
	Diagnostic       *workflowRunDiagnosticV2Document   `json:"diagnostic"`
}

type workflowRunNodeV2Document struct {
	NodeID             string                `json:"node_id"`
	NodeType           string                `json:"node_type"`
	Label              string                `json:"label"`
	Status             WorkflowRunNodeStatus `json:"status"`
	StartedAt          any                   `json:"started_at"`
	CompletedAt        any                   `json:"completed_at"`
	DurationMS         int64                 `json:"duration_ms"`
	PredecessorNodeIDs []string              `json:"predecessor_node_ids"`
	ProviderRef        string                `json:"provider_ref"`
	OutputPreview      string                `json:"output_preview"`
	FailureCode        any                   `json:"failure_code"`
}

type workflowHTTPToolAttemptV2Document struct {
	AttemptID        string                        `json:"attempt_id"`
	NodeID           string                        `json:"node_id"`
	ToolID           string                        `json:"tool_id"`
	DefinitionDigest string                        `json:"definition_digest"`
	ProfileID        string                        `json:"profile_id"`
	ProfileDigest    string                        `json:"profile_digest"`
	ToolPlanDigest   string                        `json:"tool_plan_digest"`
	ConfirmationID   string                        `json:"confirmation_id"`
	Status           WorkflowHTTPToolAttemptStatus `json:"status"`
	ClaimedAt        string                        `json:"claimed_at"`
	CompletedAt      any                           `json:"completed_at"`
	HTTPStatusClass  any                           `json:"http_status_class"`
	ResponseBytes    int                           `json:"response_bytes"`
	DurationMS       int                           `json:"duration_ms"`
	OutputProjection map[string]any                `json:"output_projection"`
	FailureCode      any                           `json:"failure_code"`
}

type workflowRunDiagnosticV2Document struct {
	FailureBoundary         any                               `json:"failure_boundary"`
	FailureStage            string                            `json:"failure_stage"`
	FailedNodeID            any                               `json:"failed_node_id"`
	LastCompletedNodeID     any                               `json:"last_completed_node_id"`
	TerminalWriteState      WorkflowRunTerminalWriteState     `json:"terminal_write_state"`
	GatewayFailureCategory  WorkflowRunGatewayFailureCategory `json:"gateway_failure_category"`
	ToolFailureCategory     WorkflowHTTPToolFailureCategory   `json:"tool_failure_category"`
	Summary                 string                            `json:"summary"`
	RecommendedReviewAction WorkflowRunReviewAction           `json:"recommended_review_action"`
	ObservedAt              string                            `json:"observed_at"`
}

func workflowHTTPToolAttemptV2DocumentFrom(attempt *WorkflowHTTPToolExecutionAttempt) *workflowHTTPToolAttemptV2Document {
	if attempt == nil {
		return nil
	}
	return &workflowHTTPToolAttemptV2Document{
		AttemptID: attempt.AttemptID, NodeID: attempt.NodeID, ToolID: attempt.ToolID,
		DefinitionDigest: attempt.DefinitionDigest, ProfileID: attempt.ProfileID,
		ProfileDigest: attempt.ProfileDigest, ToolPlanDigest: attempt.ToolPlanDigest,
		ConfirmationID: attempt.ConfirmationID, Status: attempt.Status, ClaimedAt: attempt.ClaimedAt,
		CompletedAt: nullableWorkflowRunString(attempt.CompletedAt), HTTPStatusClass: nullableWorkflowRunString(attempt.HTTPStatusClass),
		ResponseBytes: attempt.ResponseBytes, DurationMS: attempt.DurationMS,
		OutputProjection: attempt.OutputProjection, FailureCode: nullableWorkflowRunFailureCode(attempt.FailureCode),
	}
}

func workflowRunDiagnosticV2DocumentFrom(diagnostic *WorkflowRunDiagnostic) *workflowRunDiagnosticV2Document {
	if diagnostic == nil {
		return nil
	}
	return &workflowRunDiagnosticV2Document{
		FailureBoundary: nullableWorkflowRunString(string(diagnostic.FailureBoundary)),
		FailureStage:    diagnostic.FailureStage, FailedNodeID: nullableWorkflowRunString(diagnostic.FailedNodeID),
		LastCompletedNodeID: nullableWorkflowRunString(diagnostic.LastCompletedNodeID),
		TerminalWriteState:  diagnostic.TerminalWriteState, GatewayFailureCategory: diagnostic.GatewayFailureCategory,
		ToolFailureCategory: diagnostic.ToolFailureCategory, Summary: diagnostic.Summary,
		RecommendedReviewAction: diagnostic.RecommendedReviewAction, ObservedAt: diagnostic.ObservedAt,
	}
}

func nullableWorkflowRunString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableWorkflowRunFailureCode(value WorkflowRunFailureCode) any {
	if value == "" {
		return nil
	}
	return value
}

func validateWorkflowRunToolRecord(runContext WorkflowRunContext, record *WorkflowRunRecord) error {
	if record == nil || record.SchemaVersion != workflowRunRecordToolSchemaVersion ||
		record.TenantRef != strings.TrimSpace(runContext.TenantRef) || !workflowHTTPToolReferencePattern.MatchString(record.TenantRef) ||
		!workflowHTTPToolRunIDPattern.MatchString(record.RunID) ||
		!workflowHTTPToolPlanIDPattern.MatchString(record.PlanID) ||
		!workflowHTTPToolConfirmationIDPattern.MatchString(record.ConfirmationID) ||
		record.SideEffects.ToolCalls != 1 || record.SideEffects.ConfirmationCalls != 1 ||
		record.SideEffects.ProviderCalls < 0 || record.SideEffects.ProviderCalls > workflowExecutorMaxLLMCalls ||
		record.SideEffects.BusinessWrites != 0 || record.SideEffects.ReplayWrites != 0 ||
		len(record.ConditionNodeIDs) != 0 || record.ToolAttempt == nil ||
		!validWorkflowHTTPToolExecutionAttempt(*record.ToolAttempt, record) ||
		record.Diagnostic == nil || !validWorkflowHTTPToolFailureCategory(record.Diagnostic.ToolFailureCategory) {
		return errWorkflowRunStoreContract
	}
	toolNodes := 0
	for _, node := range record.Nodes {
		switch node.NodeType {
		case "prompt", "llm", "output":
		case "http_tool":
			toolNodes++
			if node.NodeID != record.ToolAttempt.NodeID {
				return errWorkflowRunStoreContract
			}
		default:
			return errWorkflowRunStoreContract
		}
	}
	if toolNodes != 1 {
		return errWorkflowRunStoreContract
	}
	switch record.Status {
	case WorkflowRunStatusRunning:
		if record.ToolAttempt.Status != WorkflowHTTPToolAttemptClaimed || record.FailureCode != "" {
			return errWorkflowRunStoreContract
		}
	case WorkflowRunStatusSucceeded:
		if record.ToolAttempt.Status != WorkflowHTTPToolAttemptSucceeded || record.FailureCode != "" {
			return errWorkflowRunStoreContract
		}
	case WorkflowRunStatusFailed, WorkflowRunStatusCanceled:
		if (record.ToolAttempt.Status != WorkflowHTTPToolAttemptFailed &&
			record.ToolAttempt.Status != WorkflowHTTPToolAttemptSucceeded) || record.FailureCode == "" ||
			(record.ToolAttempt.Status == WorkflowHTTPToolAttemptFailed && record.ToolAttempt.FailureCode != record.FailureCode) {
			return errWorkflowRunStoreContract
		}
	case WorkflowRunStatusOutcomeUnknown:
		if record.ToolAttempt.Status != WorkflowHTTPToolAttemptOutcomeUnknown ||
			record.FailureCode != WorkflowRunFailureToolOutcomeUnknown {
			return errWorkflowRunStoreContract
		}
	default:
		return errWorkflowRunStoreContract
	}
	return nil
}

func validWorkflowHTTPToolExecutionAttempt(attempt WorkflowHTTPToolExecutionAttempt, record *WorkflowRunRecord) bool {
	claimedAt, claimedErr := time.Parse(time.RFC3339Nano, attempt.ClaimedAt)
	completedAt, completedErr := time.Time{}, error(nil)
	if attempt.CompletedAt != "" {
		completedAt, completedErr = time.Parse(time.RFC3339Nano, attempt.CompletedAt)
	}
	if claimedErr != nil || completedErr != nil ||
		!workflowHTTPToolAttemptIDPattern.MatchString(attempt.AttemptID) ||
		!workflowHTTPToolScopedIDPattern.MatchString(attempt.NodeID) ||
		!workflowHTTPToolIDPattern.MatchString(attempt.ToolID) ||
		!workflowHTTPToolDigestPattern.MatchString(attempt.DefinitionDigest) ||
		!workflowHTTPToolProfileIDPattern.MatchString(attempt.ProfileID) ||
		!workflowHTTPToolDigestPattern.MatchString(attempt.ProfileDigest) ||
		!workflowHTTPToolDigestPattern.MatchString(attempt.ToolPlanDigest) ||
		!workflowHTTPToolConfirmationIDPattern.MatchString(attempt.ConfirmationID) ||
		attempt.ConfirmationID != record.ConfirmationID || attempt.ResponseBytes < 0 ||
		attempt.ResponseBytes > workflowHTTPToolMaxResponseBytes || attempt.DurationMS < 0 ||
		attempt.DurationMS > int(workflowExecutorDefaultMaxRuntime.Milliseconds()) ||
		(attempt.HTTPStatusClass != "" && attempt.HTTPStatusClass != "2xx" && attempt.HTTPStatusClass != "3xx" &&
			attempt.HTTPStatusClass != "4xx" && attempt.HTTPStatusClass != "5xx") ||
		!workflowHTTPToolProjectionIsSafe(attempt.OutputProjection) {
		return false
	}
	if attempt.Status == WorkflowHTTPToolAttemptClaimed {
		return attempt.CompletedAt == "" && attempt.HTTPStatusClass == "" && attempt.ResponseBytes == 0 &&
			attempt.DurationMS == 0 && len(attempt.OutputProjection) == 0 && attempt.FailureCode == ""
	}
	if attempt.CompletedAt == "" || completedAt.Before(claimedAt) || attempt.Status == "" {
		return false
	}
	switch attempt.Status {
	case WorkflowHTTPToolAttemptSucceeded:
		return attempt.FailureCode == "" && attempt.HTTPStatusClass == "2xx" && len(attempt.OutputProjection) > 0
	case WorkflowHTTPToolAttemptFailed:
		return attempt.FailureCode != "" && validWorkflowRunFailureCode(attempt.FailureCode) && len(attempt.OutputProjection) == 0
	case WorkflowHTTPToolAttemptOutcomeUnknown:
		return attempt.FailureCode == WorkflowRunFailureToolOutcomeUnknown && len(attempt.OutputProjection) == 0
	default:
		return false
	}
}

func workflowHTTPToolProjectionIsSafe(projection map[string]any) bool {
	if projection == nil || len(projection) > 16 {
		return false
	}
	for key, value := range projection {
		if !workflowHTTPToolProjectionKeyPattern.MatchString(key) || workflowHTTPToolProjectionKeyForbidden(key) {
			return false
		}
		switch typed := value.(type) {
		case string:
			if len([]rune(typed)) > 4096 || strings.Contains(typed, "://") {
				return false
			}
		case float64, float32, int, int64, bool, nil:
		default:
			return false
		}
	}
	encoded, err := json.Marshal(projection)
	return err == nil && len(encoded) <= workflowHTTPToolMaxOutputBytes
}

func workflowHTTPToolProjectionKeyForbidden(key string) bool {
	switch key {
	case "endpoint", "url", "uri", "header", "headers", "authorization", "cookie", "credential",
		"raw_query", "raw_request", "raw_response", "dns", "ip_address", "internal_error":
		return true
	default:
		return false
	}
}
