package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"
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

func handleModels(writer http.ResponseWriter, server *Server) {
	ctx, cancel := context.WithTimeout(context.Background(), server.config.BridgeTimeout)
	defer cancel()

	inventory, err := server.bridge.DescribeInventory(ctx)
	if err != nil {
		writeOpenAIError(writer, http.StatusBadGateway, "PROVIDER_REGISTRY_UNAVAILABLE", err.Error())
		return
	}

	now := time.Now().Unix()
	models := make([]openAIModelObject, 0, len(inventory.Providers)+len(inventory.Profiles)+1)
	seenModelIDs := map[string]struct{}{}
	if configuredModel := strings.TrimSpace(server.config.Model); configuredModel != "" {
		seenModelIDs[configuredModel] = struct{}{}
		models = append(models, openAIModelObject{
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
		})
	}
	for _, provider := range inventory.Providers {
		if _, seen := seenModelIDs[provider.ProviderID]; seen {
			continue
		}
		seenModelIDs[provider.ProviderID] = struct{}{}
		models = append(models, openAIModelObject{
			ID:      provider.ProviderID,
			Object:  "model",
			Created: now,
			OwnedBy: "radishmind",
			Metadata: map[string]any{
				"source":               "provider_registry",
				"provider_id":          provider.ProviderID,
				"display_name":         provider.DisplayName,
				"default_api_style":    provider.DefaultAPIStyle,
				"supported_api_styles": provider.SupportedAPIStyles,
				"profile_driven":       provider.ProfileDriven,
				"northbound_routes":    []string{"/v1/chat/completions", "/v1/responses", "/v1/messages", "/v1/models"},
				"capabilities":         provider.Capabilities,
			},
		})
	}
	for _, profile := range inventory.Profiles {
		modelID := buildNorthboundProfileModelID(profile.Profile)
		if _, seen := seenModelIDs[modelID]; seen {
			continue
		}
		seenModelIDs[modelID] = struct{}{}
		models = append(models, openAIModelObject{
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
		})
	}

	writeJSON(writer, http.StatusOK, openAIModelListResponse{
		Object: "list",
		Data:   models,
	})
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
