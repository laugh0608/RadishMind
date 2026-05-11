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

	providers, err := server.bridge.DescribeProviders(ctx)
	if err != nil {
		writeOpenAIError(writer, http.StatusBadGateway, "PROVIDER_REGISTRY_UNAVAILABLE", err.Error())
		return
	}

	now := time.Now().Unix()
	models := make([]openAIModelObject, 0, len(providers)+1)
	seenModelIDs := map[string]struct{}{}
	if configuredModel := strings.TrimSpace(server.config.Model); configuredModel != "" {
		seenModelIDs[configuredModel] = struct{}{}
		models = append(models, openAIModelObject{
			ID:      configuredModel,
			Object:  "model",
			Created: now,
			OwnedBy: strings.TrimSpace(server.config.Provider),
			Metadata: map[string]any{
				"source":               "configured_default",
				"provider_id":          strings.TrimSpace(server.config.Provider),
				"provider_profile":     strings.TrimSpace(server.config.ProviderProfile),
				"northbound_routes":    []string{"/v1/chat/completions", "/v1/responses", "/v1/messages", "/v1/models"},
				"northbound_protocols": []string{northboundProtocolChatCompletions, northboundProtocolResponses, northboundProtocolMessages},
			},
		})
	}
	for _, provider := range providers {
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

	writeJSON(writer, http.StatusOK, openAIModelListResponse{
		Object: "list",
		Data:   models,
	})
}
