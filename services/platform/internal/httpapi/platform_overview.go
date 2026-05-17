package httpapi

import (
	"context"
	"net/http"
	"strings"
)

const platformOverviewRoute = "/v1/platform/overview"

func (s *Server) handlePlatformOverview(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, platformOverviewRoute)
	ctx, cancel := context.WithTimeout(context.Background(), s.config.BridgeTimeout)
	defer cancel()

	models := s.buildPlatformOverviewModels(ctx)
	document := map[string]any{
		"schema_version": 1,
		"kind":           "platform_overview",
		"stage":          "P3 Local Product Shell",
		"route":          platformOverviewRoute,
		"service": map[string]any{
			"name":    serviceName,
			"version": s.options.BuildVersion,
			"status":  "ok",
		},
		"product_surface": map[string]any{
			"mode":          "local_read_only_product_shell",
			"implemented":   true,
			"ui_consumable": true,
			"routes": []string{
				"/healthz",
				platformOverviewRoute,
				"/v1/models",
				"/v1/models/{id}",
				sessionMetadataRoute,
				toolsMetadataRoute,
				toolActionsRoute,
			},
		},
		"models": models,
		"session_tooling": map[string]any{
			"stage":                      "P2 close candidate shell",
			"session_metadata_route":     sessionMetadataRoute,
			"tools_metadata_route":       toolsMetadataRoute,
			"blocked_action_route":       toolActionsRoute,
			"metadata_only":              true,
			"tool_count":                 len(platformToolsMetadata()),
			"execution_enabled":          false,
			"tool_action_status":         "blocked",
			"requires_confirmation_path": "future_upper_layer_confirmation_flow",
		},
		"stop_lines": map[string]any{
			"real_executor_enabled":           false,
			"durable_store_enabled":           false,
			"confirmation_flow_connected":     false,
			"materialized_result_reader":      false,
			"long_term_memory_enabled":        false,
			"business_truth_write_enabled":    false,
			"automatic_replay_enabled":        false,
			"production_secret_backend_ready": false,
		},
		"audit": map[string]any{
			"advisory_only":         true,
			"writes_business_truth": false,
			"notes": []string{
				"overview aggregates existing local product surface metadata for UI or upper-layer discovery",
				"it does not enable executor, durable storage, confirmation flow, long-term memory, business writeback, or replay",
			},
		},
	}
	writeObservedJSON(writer, http.StatusOK, trace, document)
}

func (s *Server) buildPlatformOverviewModels(ctx context.Context) map[string]any {
	models := map[string]any{
		"route":            "/v1/models",
		"detail_route":     "/v1/models/{id}",
		"inventory_kind":   "bridge_backed_provider_profile_inventory",
		"default_provider": strings.TrimSpace(s.config.Provider),
		"default_profile":  strings.TrimSpace(s.config.ProviderProfile),
		"default_model":    strings.TrimSpace(s.config.Model),
	}

	inventory, err := s.bridge.DescribeInventory(ctx)
	if err != nil {
		models["status"] = "unavailable"
		models["failure_code"] = "PROVIDER_INVENTORY_UNAVAILABLE"
		models["failure_boundary"] = errorBoundaryPythonBridge
		models["message"] = sanitizePlatformErrorDetail(s.config, err.Error())
		return models
	}

	catalog := buildNorthboundModelCatalog(s, inventory)
	selectableModelIDs := make([]string, 0, len(catalog.list.Data))
	for _, model := range catalog.list.Data {
		modelID := strings.TrimSpace(model.ID)
		if modelID != "" {
			selectableModelIDs = append(selectableModelIDs, modelID)
		}
	}

	models["status"] = "ok"
	models["model_count"] = len(catalog.list.Data)
	models["provider_count"] = len(inventory.Providers)
	models["profile_count"] = len(inventory.Profiles)
	models["active_profile_chain"] = inventory.ActiveProfileChain
	models["selectable_model_ids"] = selectableModelIDs
	return models
}
