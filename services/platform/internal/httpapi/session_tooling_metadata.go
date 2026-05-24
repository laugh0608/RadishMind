package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	sessionMetadataRoute = "/v1/session/metadata"
	toolsMetadataRoute   = "/v1/tools/metadata"
	toolActionsRoute     = "/v1/tools/actions"
)

type platformToolMetadata struct {
	ToolID                         string
	DisplayName                    string
	ToolType                       string
	ProjectScope                   string
	RiskLevel                      string
	RequiresConfirmationForActions bool
	InputSchemaRef                 string
	OutputSchemaRef                string
}

type toolActionRequest struct {
	ToolID    string         `json:"tool_id"`
	Action    string         `json:"action"`
	RequestID string         `json:"request_id,omitempty"`
	SessionID string         `json:"session_id,omitempty"`
	TurnID    string         `json:"turn_id,omitempty"`
	Payload   map[string]any `json:"payload,omitempty"`
}

func (s *Server) handleSessionMetadata(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, sessionMetadataRoute)
	writeObservedJSON(writer, http.StatusOK, trace, buildSessionMetadataDocument())
}

func (s *Server) handleToolsMetadata(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, toolsMetadataRoute)
	writeObservedJSON(writer, http.StatusOK, trace, buildToolsMetadataDocument())
}

func (s *Server) handleToolAction(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, toolActionsRoute)
	var actionRequest toolActionRequest
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&actionRequest); err != nil {
		s.writePlatformError(writer, trace, "INVALID_JSON", fmt.Sprintf("invalid tool action request: %v", err))
		return
	}

	toolID := strings.TrimSpace(actionRequest.ToolID)
	if toolID == "" {
		s.writePlatformError(writer, trace, "MISSING_TOOL_ID", "tool_id is required")
		return
	}

	tool, found := lookupPlatformToolMetadata(toolID)
	if !found {
		writeObservedJSON(writer, http.StatusOK, trace, buildBlockedToolActionDocument(trace, actionRequest, platformToolMetadata{}, false))
		return
	}
	writeObservedJSON(writer, http.StatusOK, trace, buildBlockedToolActionDocument(trace, actionRequest, tool, true))
}

func buildSessionMetadataDocument() map[string]any {
	return map[string]any{
		"schema_version": 1,
		"kind":           "session_metadata",
		"stage":          "P2 Session & Tooling Foundation",
		"route":          sessionMetadataRoute,
		"api_boundary": map[string]any{
			"surface":        "platform_metadata",
			"implemented":    true,
			"response_shape": "metadata_only",
		},
		"supported_extension_fields": []string{
			"conversation_id",
			"turn_id",
			"parent_turn_id",
			"history_policy",
			"history_window",
		},
		"history_policy": map[string]any{
			"supported_modes":      []string{"windowed", "stateless", "summary_only", "disabled"},
			"default_mode":         "windowed",
			"default_max_turns":    8,
			"include_system":       true,
			"include_tool_results": false,
			"compression": map[string]any{
				"enabled":  false,
				"strategy": "none",
			},
		},
		"state_policy": map[string]any{
			"session_state_scope":       "northbound_metadata",
			"tool_result_cache_scope":   "metadata_only",
			"recovery_checkpoint_scope": "audit_refs_only",
			"durable_memory_enabled":    false,
		},
		"capabilities": map[string]any{
			"session_metadata":         true,
			"durable_session_store":    false,
			"durable_checkpoint_store": false,
			"long_term_memory":         false,
			"automatic_replay":         false,
			"business_truth_write":     false,
		},
		"recovery_boundary": map[string]any{
			"checkpoint_read_route":             sessionRecoveryCheckpointRoute,
			"metadata_only":                     true,
			"materialized_results_included":     false,
			"result_ref_reader_enabled":         false,
			"replay_executor_enabled":           false,
			"requires_confirmation_for_actions": true,
		},
		"audit": map[string]any{
			"advisory_only":         true,
			"writes_business_truth": false,
			"notes": []string{
				"session metadata describes northbound compatibility and recovery boundaries only",
				"platform service does not own a durable session store in the current P2 shell",
			},
		},
	}
}

func buildToolsMetadataDocument() map[string]any {
	tools := make([]map[string]any, 0, len(platformToolsMetadata()))
	for _, tool := range platformToolsMetadata() {
		tools = append(tools, map[string]any{
			"tool_id":                           tool.ToolID,
			"display_name":                      tool.DisplayName,
			"tool_type":                         tool.ToolType,
			"project_scope":                     tool.ProjectScope,
			"risk_level":                        tool.RiskLevel,
			"requires_confirmation_for_actions": tool.RequiresConfirmationForActions,
			"interface": map[string]any{
				"input_schema_ref":  tool.InputSchemaRef,
				"output_schema_ref": tool.OutputSchemaRef,
			},
			"execution": map[string]any{
				"mode":              "contract_only",
				"execution_enabled": false,
				"status":            "disabled",
			},
			"state_policy": map[string]any{
				"result_cache_mode":        "metadata_only",
				"durable_memory_enabled":   false,
				"writes_business_truth":    false,
				"materialized_result_read": false,
			},
		})
	}

	return map[string]any{
		"schema_version": 1,
		"kind":           "tooling_metadata",
		"registry_id":    "radishmind-tool-registry-v1",
		"route":          toolsMetadataRoute,
		"api_boundary": map[string]any{
			"surface":        "platform_metadata",
			"implemented":    true,
			"response_shape": "metadata_only",
		},
		"registry_policy": map[string]any{
			"execution_enabled":       false,
			"durable_memory_enabled":  false,
			"network_default":         "disabled",
			"default_timeout_seconds": 30,
			"max_retry_attempts":      1,
		},
		"tools":                tools,
		"blocked_action_route": toolActionsRoute,
		"audit": map[string]any{
			"advisory_only":         true,
			"writes_business_truth": false,
			"notes": []string{
				"tool metadata is a product-facing view over the current contract-only registry",
				"real execution, durable result storage, and business writeback remain disabled",
			},
		},
	}
}

func buildBlockedToolActionDocument(trace requestTrace, actionRequest toolActionRequest, tool platformToolMetadata, knownTool bool) map[string]any {
	toolID := strings.TrimSpace(actionRequest.ToolID)
	action := strings.TrimSpace(actionRequest.Action)
	if action == "" {
		action = "execute"
	}

	denialCodes := []string{"TOOL_EXECUTOR_DISABLED"}
	reason := "tool execution is disabled in the current P2 session/tooling shell"
	requiresConfirmation := false
	if knownTool {
		requiresConfirmation = tool.RequiresConfirmationForActions
		if requiresConfirmation {
			denialCodes = append(denialCodes, "CONFIRMATION_REQUIRED")
			reason = "tool execution is disabled and this tool requires upper-layer confirmation before any future execution"
		}
	} else {
		denialCodes = []string{"TOOL_NOT_REGISTERED", "TOOL_EXECUTOR_DISABLED"}
		reason = "tool is not registered in the current metadata registry and execution is disabled"
	}

	return map[string]any{
		"schema_version": 1,
		"kind":           "tool_action_blocked_response",
		"status":         "blocked",
		"route":          toolActionsRoute,
		"request_id":     trace.requestID,
		"action": map[string]any{
			"tool_id":    toolID,
			"action":     action,
			"known_tool": knownTool,
			"session_id": strings.TrimSpace(actionRequest.SessionID),
			"turn_id":    strings.TrimSpace(actionRequest.TurnID),
		},
		"policy_decision": map[string]any{
			"decision":              "blocked",
			"primary_code":          denialCodes[0],
			"denial_codes":          denialCodes,
			"requires_confirmation": requiresConfirmation,
			"reason":                reason,
		},
		"execution": map[string]any{
			"execution_enabled": false,
			"executed":          false,
			"status":            "not_executed",
			"duration_ms":       nil,
		},
		"result": map[string]any{
			"result_ref":                   nil,
			"materialized_result_included": false,
			"materialized_result_read":     false,
			"result_cache_mode":            "metadata_only",
		},
		"side_effects": map[string]any{
			"network_request_sent":     false,
			"durable_memory_written":   false,
			"writes_business_truth":    false,
			"automatic_replay_started": false,
		},
		"audit": map[string]any{
			"advisory_only":         true,
			"writes_business_truth": false,
			"notes": []string{
				"blocked action response is returned without running a tool executor",
				"future enablement still requires confirmation, storage, audit, and negative regression gates",
			},
		},
	}
}

func lookupPlatformToolMetadata(toolID string) (platformToolMetadata, bool) {
	for _, tool := range platformToolsMetadata() {
		if tool.ToolID == strings.TrimSpace(toolID) {
			return tool, true
		}
	}
	return platformToolMetadata{}, false
}

func platformToolsMetadata() []platformToolMetadata {
	return []platformToolMetadata{
		{
			ToolID:                         "radish.docs.retrieval_context.v1",
			DisplayName:                    "Radish Docs Retrieval Context",
			ToolType:                       "retrieval",
			ProjectScope:                   "radish",
			RiskLevel:                      "low",
			RequiresConfirmationForActions: false,
			InputSchemaRef:                 "contracts/copilot-request.schema.json#/properties/context",
			OutputSchemaRef:                "contracts/copilot-response.schema.json#/properties/citations",
		},
		{
			ToolID:                         "radishflow.suggest_edits.candidate_builder.v1",
			DisplayName:                    "RadishFlow Suggest Edits Candidate Builder",
			ToolType:                       "candidate_builder",
			ProjectScope:                   "radishflow",
			RiskLevel:                      "medium",
			RequiresConfirmationForActions: true,
			InputSchemaRef:                 "contracts/copilot-request.schema.json",
			OutputSchemaRef:                "contracts/copilot-response.schema.json#/properties/candidate_actions",
		},
	}
}
