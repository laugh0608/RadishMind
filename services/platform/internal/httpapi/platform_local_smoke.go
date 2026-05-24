package httpapi

import (
	"context"
	"net/http"
)

const platformLocalSmokeRoute = "/v1/platform/local-smoke"

func (s *Server) handlePlatformLocalSmoke(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, platformLocalSmokeRoute)
	ctx, cancel := context.WithTimeout(context.Background(), s.config.BridgeTimeout)
	defer cancel()

	models := s.buildPlatformOverviewModels(ctx)
	sessionMetadata := buildSessionMetadataDocument()
	toolsMetadata := buildToolsMetadataDocument()
	blockedAction := buildLocalSmokeBlockedAction(trace)
	blockedActionSummary := buildLocalSmokeBlockedActionSummary(blockedAction)
	stopLines := buildPlatformStopLinesDocument()

	modelInventoryReadable := models["status"] == "ok"
	sessionMetadataReadable := sessionMetadata["kind"] == "session_metadata"
	toolsMetadataReadable := toolsMetadata["kind"] == "tooling_metadata"
	blockedActionNoSideEffects := blockedActionSummary["no_side_effects"] == true
	stopLinesEnforced := localSmokeStopLinesEnforced(stopLines)
	localConsoleReady := modelInventoryReadable &&
		sessionMetadataReadable &&
		toolsMetadataReadable &&
		blockedActionSummary["status"] == "blocked" &&
		blockedActionNoSideEffects &&
		stopLinesEnforced
	status := "ok"
	if !localConsoleReady {
		status = "degraded"
	}

	document := map[string]any{
		"schema_version": 1,
		"kind":           "platform_local_smoke",
		"stage":          "P3 Local Product Shell / Ops Surface",
		"route":          platformLocalSmokeRoute,
		"summary": map[string]any{
			"status":              status,
			"local_console_ready": localConsoleReady,
			"read_only":           true,
			"default_ports": map[string]any{
				"platform": 7000,
				"console":  4000,
			},
		},
		"checks": map[string]any{
			"healthz": map[string]any{
				"route":    "/healthz",
				"status":   "ok",
				"readable": true,
			},
			"overview": map[string]any{
				"route":          platformOverviewRoute,
				"readable":       true,
				"contract_kind":  "platform_overview",
				"ui_consumable":  true,
				"product_routes": platformProductSurfaceRoutes(),
			},
			"model_inventory": map[string]any{
				"route":          "/v1/models",
				"detail_route":   "/v1/models/{id}",
				"status":         models["status"],
				"readable":       modelInventoryReadable,
				"inventory_kind": models["inventory_kind"],
				"model_count":    models["model_count"],
				"provider_count": models["provider_count"],
				"profile_count":  models["profile_count"],
				"failure_code":   models["failure_code"],
			},
			"session_tooling": map[string]any{
				"session_metadata_route":         sessionMetadataRoute,
				"tools_metadata_route":           toolsMetadataRoute,
				"blocked_action_route":           toolActionsRoute,
				"session_metadata_readable":      sessionMetadataReadable,
				"tools_metadata_readable":        toolsMetadataReadable,
				"tool_count":                     localSmokeToolCount(toolsMetadata),
				"metadata_only":                  true,
				"execution_enabled":              false,
				"blocked_action_status":          blockedActionSummary["status"],
				"blocked_action_primary_code":    blockedActionSummary["primary_code"],
				"requires_confirmation":          blockedActionSummary["requires_confirmation"],
				"blocked_action_no_side_effects": blockedActionNoSideEffects,
			},
			"local_console": map[string]any{
				"frontend_origin_default": "http://127.0.0.1:4000",
				"backend_url_default":     "http://127.0.0.1:7000",
				"allowed_cors_origins":    localConsoleAllowedOrigins(),
				"cors_preflight_methods":  []string{"GET", "POST", "OPTIONS"},
				"cors_scope":              "local_dev_only",
			},
		},
		"stop_lines": stopLines,
		"failure_hints": []map[string]any{
			{
				"code":    "PORT_IN_USE",
				"message": "default local ports are platform 7000 and console 4000; release the port or confirm the existing process is the expected RadishMind service",
			},
			{
				"code":    "CORS_ORIGIN_NOT_ALLOWED",
				"message": "local console CORS is only allowed for http://127.0.0.1:4000 and http://localhost:4000",
			},
			{
				"code":    "ERR_UNSAFE_PORT",
				"message": "some browser ports are blocked by unsafe-port policy; prefer the default local console and platform ports",
			},
		},
		"audit": map[string]any{
			"advisory_only":         true,
			"writes_business_truth": false,
			"notes": []string{
				"local smoke summarizes existing read-only platform routes for development-time console readiness",
				"it does not start processes or enable executor, durable storage, confirmation flow, business writeback, or replay",
			},
		},
	}
	writeObservedJSON(writer, http.StatusOK, trace, document)
}

func buildLocalSmokeBlockedAction(trace requestTrace) map[string]any {
	tool, found := lookupPlatformToolMetadata("radishflow.suggest_edits.candidate_builder.v1")
	return buildBlockedToolActionDocument(
		trace,
		toolActionRequest{
			ToolID:    "radishflow.suggest_edits.candidate_builder.v1",
			Action:    "execute",
			SessionID: "local-smoke-session",
			TurnID:    "local-smoke-turn",
		},
		tool,
		found,
	)
}

func buildLocalSmokeBlockedActionSummary(blockedAction map[string]any) map[string]any {
	policyDecision, _ := blockedAction["policy_decision"].(map[string]any)
	execution, _ := blockedAction["execution"].(map[string]any)
	result, _ := blockedAction["result"].(map[string]any)
	sideEffects, _ := blockedAction["side_effects"].(map[string]any)
	return map[string]any{
		"status":                blockedAction["status"],
		"primary_code":          policyDecision["primary_code"],
		"requires_confirmation": policyDecision["requires_confirmation"],
		"no_side_effects": execution["executed"] == false &&
			result["result_ref"] == nil &&
			sideEffects["network_request_sent"] == false &&
			sideEffects["durable_memory_written"] == false &&
			sideEffects["writes_business_truth"] == false &&
			sideEffects["automatic_replay_started"] == false,
	}
}

func localSmokeToolCount(toolsMetadata map[string]any) int {
	tools, ok := toolsMetadata["tools"].([]map[string]any)
	if !ok {
		return 0
	}
	return len(tools)
}

func localSmokeStopLinesEnforced(stopLines map[string]any) bool {
	for _, key := range []string{
		"real_executor_enabled",
		"durable_store_enabled",
		"confirmation_flow_connected",
		"materialized_result_reader",
		"long_term_memory_enabled",
		"business_truth_write_enabled",
		"automatic_replay_enabled",
	} {
		if stopLines[key] != false {
			return false
		}
	}
	return true
}
