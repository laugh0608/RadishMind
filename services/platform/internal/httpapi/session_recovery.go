package httpapi

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	sessionRecoveryCheckpointRoute     = "/v1/session/recovery/checkpoints/{checkpoint_id}"
	fixtureRecoveryCheckpointID        = "session-checkpoint-0001"
	fixtureRecoveryCheckpointSessionID = "radishflow-session-001"
	fixtureRecoveryCheckpointTurnID    = "turn-0003"
)

func (s *Server) handleSessionRecoveryCheckpoint(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, sessionRecoveryCheckpointRoute)
	checkpointID := strings.TrimSpace(request.PathValue("checkpoint_id"))
	if checkpointID == "" {
		s.writePlatformError(writer, trace, "MISSING_CHECKPOINT_ID", "checkpoint id is required")
		return
	}

	sessionID := strings.TrimSpace(request.URL.Query().Get("session_id"))
	if sessionID == "" {
		sessionID = fixtureRecoveryCheckpointSessionID
	}
	turnID := strings.TrimSpace(request.URL.Query().Get("turn_id"))
	if turnID == "" {
		turnID = fixtureRecoveryCheckpointTurnID
	}

	if checkpointID != fixtureRecoveryCheckpointID || sessionID != fixtureRecoveryCheckpointSessionID || turnID != fixtureRecoveryCheckpointTurnID {
		s.writePlatformError(writer, trace, "CHECKPOINT_NOT_FOUND", fmt.Sprintf("checkpoint not found: %s", checkpointID))
		return
	}

	writeObservedJSON(writer, http.StatusOK, trace, buildSessionRecoveryCheckpointReadResult(checkpointID, sessionID, turnID))
}

func buildSessionRecoveryCheckpointReadResult(checkpointID string, sessionID string, turnID string) map[string]any {
	return map[string]any{
		"schema_version": 1,
		"kind":           "session_recovery_checkpoint_read_result",
		"read_id":        "session-checkpoint-read-0001",
		"api_boundary": map[string]any{
			"surface":        "platform_metadata",
			"route":          sessionRecoveryCheckpointRoute,
			"implemented":    false,
			"response_shape": "metadata_refs_only",
		},
		"request": map[string]any{
			"checkpoint_id":                checkpointID,
			"session_id":                   sessionID,
			"turn_id":                      turnID,
			"include_refs":                 true,
			"include_materialized_results": false,
		},
		"result": map[string]any{
			"status":         "found",
			"checkpoint_id":  checkpointID,
			"session_id":     sessionID,
			"turn_id":        turnID,
			"checkpoint_ref": "scripts/checks/fixtures/session-recovery-checkpoint-basic.json",
			"refs": []map[string]any{
				{
					"kind":                  "request",
					"ref":                   "request:req-tooling-001",
					"required_for_recovery": true,
				},
				{
					"kind":                  "session_record",
					"ref":                   "session:radishflow-session-001:turn-0003",
					"required_for_recovery": true,
				},
				{
					"kind":                  "tool_audit",
					"ref":                   "tool-audit:tool-audit-0001",
					"required_for_recovery": true,
				},
				{
					"kind":                  "tool_state",
					"ref":                   "tool-state:tool-audit-0001",
					"required_for_recovery": false,
				},
				{
					"kind":                  "tool_result_metadata",
					"ref":                   "tool-result-cache:req-tooling-001:turn-0003:radishflow.suggest_edits.candidate_builder.v1",
					"required_for_recovery": false,
				},
			},
			"replay_policy": map[string]any{
				"replayable":                        true,
				"auto_replay_enabled":               false,
				"requires_confirmation_for_actions": true,
			},
			"state_summary": map[string]any{
				"contains_tool_result_metadata":      true,
				"contains_materialized_tool_results": false,
				"contains_business_truth":            false,
			},
		},
		"access_policy": map[string]any{
			"metadata_only":                 true,
			"materialized_results_included": false,
			"durable_memory_enabled":        false,
			"writes_business_truth":         false,
			"auto_replay_enabled":           false,
		},
		"audit": map[string]any{
			"advisory_only": true,
			"notes": []string{
				"read boundary exposes checkpoint refs and policy metadata only",
				"route is a contract boundary and is not implemented as a replay executor",
			},
		},
	}
}
