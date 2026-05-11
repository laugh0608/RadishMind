package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/bridge"
)

type openAIModelListResponse struct {
	Object string              `json:"object"`
	Data   []openAIModelObject `json:"data"`
}

type openAIModelObject struct {
	ID       string         `json:"id"`
	Object   string         `json:"object"`
	Created  int64          `json:"created"`
	OwnedBy  string         `json:"owned_by"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type northboundModelCatalog struct {
	list openAIModelListResponse
	byID map[string]openAIModelObject
}

func handleModels(writer http.ResponseWriter, server *Server) {
	ctx, cancel := context.WithTimeout(context.Background(), server.config.BridgeTimeout)
	defer cancel()

	inventory, err := server.bridge.DescribeInventory(ctx)
	if err != nil {
		writeOpenAIError(writer, http.StatusBadGateway, "PROVIDER_REGISTRY_UNAVAILABLE", err.Error())
		return
	}

	catalog := buildNorthboundModelCatalog(server, inventory)
	writeJSON(writer, http.StatusOK, catalog.list)
}

func handleModel(writer http.ResponseWriter, request *http.Request, server *Server) {
	modelID := strings.TrimSpace(request.PathValue("id"))
	if modelID == "" {
		writeOpenAIError(writer, http.StatusBadRequest, "MISSING_MODEL_ID", "model id is required")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), server.config.BridgeTimeout)
	defer cancel()

	inventory, err := server.bridge.DescribeInventory(ctx)
	if err != nil {
		writeOpenAIError(writer, http.StatusBadGateway, "PROVIDER_REGISTRY_UNAVAILABLE", err.Error())
		return
	}

	catalog := buildNorthboundModelCatalog(server, inventory)
	model, ok := catalog.byID[modelID]
	if !ok {
		writeOpenAIError(writer, http.StatusNotFound, "MODEL_NOT_FOUND", fmt.Sprintf("model not found: %s", modelID))
		return
	}

	writeJSON(writer, http.StatusOK, model)
}

func buildNorthboundModelCatalog(server *Server, inventory bridge.ProviderInventory) northboundModelCatalog {
	now := time.Now().Unix()
	models := make([]openAIModelObject, 0, len(inventory.Providers)+len(inventory.Profiles)+1)
	byID := make(map[string]openAIModelObject, cap(models))

	appendModel := func(model openAIModelObject) {
		modelID := strings.TrimSpace(model.ID)
		if modelID == "" {
			return
		}
		if _, seen := byID[modelID]; seen {
			return
		}
		byID[modelID] = model
		models = append(models, model)
	}

	if configuredModel := strings.TrimSpace(server.config.Model); configuredModel != "" {
		appendModel(buildNorthboundConfiguredModel(now, server, configuredModel, inventory))
	}
	for _, provider := range inventory.Providers {
		appendModel(buildNorthboundProviderModel(now, provider))
	}
	for _, profile := range inventory.Profiles {
		appendModel(buildNorthboundProfileModel(now, profile))
	}

	return northboundModelCatalog{
		list: openAIModelListResponse{
			Object: "list",
			Data:   models,
		},
		byID: byID,
	}
}

func buildNorthboundConfiguredModel(
	now int64,
	server *Server,
	configuredModel string,
	inventory bridge.ProviderInventory,
) openAIModelObject {
	return openAIModelObject{
		ID:      configuredModel,
		Object:  "model",
		Created: now,
		OwnedBy: strings.TrimSpace(server.config.Provider),
		Metadata: map[string]any{
			"source":                     "configured_default",
			"provider_id":                strings.TrimSpace(server.config.Provider),
			"provider_profile":           strings.TrimSpace(server.config.ProviderProfile),
			"northbound_routes":          []string{"/v1/chat/completions", "/v1/responses", "/v1/messages", "/v1/models"},
			"northbound_protocols":       []string{northboundProtocolChatCompletions, northboundProtocolResponses, northboundProtocolMessages},
			"inventory_kind":             "configured_default",
			"active_profile_chain":       inventory.ActiveProfileChain,
			"profile_inventory":          inventory.Profiles,
			"profile_inventory_count":    len(inventory.Profiles),
			"provider_inventory_count":   len(inventory.Providers),
			"provider_registry_version":  "bridge-backed",
			"resolved_selection_summary": buildNorthboundSelectionSummary(server.config.Provider, server.config.ProviderProfile, configuredModel),
		},
	}
}

func buildNorthboundProviderModel(now int64, provider bridge.ProviderDescription) openAIModelObject {
	return openAIModelObject{
		ID:      strings.TrimSpace(provider.ProviderID),
		Object:  "model",
		Created: now,
		OwnedBy: "radishmind",
		Metadata: map[string]any{
			"source":               "provider_registry",
			"provider_id":          strings.TrimSpace(provider.ProviderID),
			"display_name":         provider.DisplayName,
			"default_api_style":    provider.DefaultAPIStyle,
			"supported_api_styles": provider.SupportedAPIStyles,
			"profile_driven":       provider.ProfileDriven,
			"northbound_routes":    []string{"/v1/chat/completions", "/v1/responses", "/v1/messages", "/v1/models"},
			"capabilities":         provider.Capabilities,
		},
	}
}

func buildNorthboundProfileModel(now int64, profile bridge.ProviderProfileDescription) openAIModelObject {
	modelID := buildNorthboundProfileModelID(profile.Profile)
	return openAIModelObject{
		ID:      modelID,
		Object:  "model",
		Created: now,
		OwnedBy: "radishmind",
		Metadata: map[string]any{
			"source":                  "provider_profile_inventory",
			"inventory_kind":          "provider_profile",
			"provider_id":             profile.ProviderID,
			"provider_profile":        profile.Profile,
			"normalized_profile":      profile.NormalizedProfile,
			"resolved_model":          profile.ResolvedModel,
			"api_style":               profile.APIStyle,
			"active":                  profile.Active,
			"fallback":                profile.Fallback,
			"chain_index":             profile.ChainIndex,
			"has_base_url":            profile.HasBaseURL,
			"has_api_key":             profile.HasAPIKey,
			"request_timeout_seconds": profile.RequestTimeoutSeconds,
			"northbound_routes":       []string{"/v1/chat/completions", "/v1/responses", "/v1/messages", "/v1/models"},
			"selection_summary":       buildNorthboundSelectionSummary(profile.ProviderID, profile.Profile, profile.ResolvedModel),
		},
	}
}

func buildNorthboundSelectionSummary(provider string, providerProfile string, model string) map[string]any {
	return map[string]any{
		"provider":         strings.TrimSpace(provider),
		"provider_profile": strings.TrimSpace(providerProfile),
		"model":            strings.TrimSpace(model),
	}
}

func buildNorthboundProfileModelID(profile string) string {
	trimmedProfile := strings.TrimSpace(profile)
	if trimmedProfile == "" {
		return "profile:default"
	}
	return "profile:" + trimmedProfile
}
