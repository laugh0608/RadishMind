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
	list    openAIModelListResponse
	byID    map[string]openAIModelObject
	aliases map[string]string
}

func handleModels(writer http.ResponseWriter, request *http.Request, server *Server) {
	trace := newRequestTrace(request, "/v1/models")
	ctx, cancel := context.WithTimeout(context.Background(), server.config.BridgeTimeout)
	defer cancel()

	inventory, err := server.bridge.DescribeInventory(ctx)
	if err != nil {
		server.writePlatformError(writer, trace, "PROVIDER_INVENTORY_UNAVAILABLE", err.Error())
		return
	}

	catalog := buildNorthboundModelCatalog(server, inventory)
	writeObservedJSON(writer, http.StatusOK, trace, catalog.list)
}

func handleModel(writer http.ResponseWriter, request *http.Request, server *Server) {
	trace := newRequestTrace(request, "/v1/models/{id}")
	modelID := strings.TrimSpace(request.PathValue("id"))
	if modelID == "" {
		server.writePlatformError(writer, trace, "MISSING_MODEL_ID", "model id is required")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), server.config.BridgeTimeout)
	defer cancel()

	inventory, err := server.bridge.DescribeInventory(ctx)
	if err != nil {
		server.writePlatformError(writer, trace, "PROVIDER_INVENTORY_UNAVAILABLE", err.Error())
		return
	}

	catalog := buildNorthboundModelCatalog(server, inventory)
	model, ok := catalog.lookup(modelID)
	if !ok {
		server.writePlatformError(writer, trace, "MODEL_NOT_FOUND", fmt.Sprintf("model not found: %s", modelID))
		return
	}

	writeObservedJSON(writer, http.StatusOK, trace, model)
}

func buildNorthboundModelCatalog(server *Server, inventory bridge.ProviderInventory) northboundModelCatalog {
	now := time.Now().Unix()
	models := make([]openAIModelObject, 0, len(inventory.Providers)+len(inventory.Profiles)+1)
	byID := make(map[string]openAIModelObject, cap(models))
	aliases := make(map[string]string, cap(models))

	registerAlias := func(alias string, modelID string) {
		normalizedAlias := strings.TrimSpace(alias)
		normalizedModelID := strings.TrimSpace(modelID)
		if normalizedAlias == "" || normalizedModelID == "" || normalizedAlias == normalizedModelID {
			return
		}
		if _, seen := byID[normalizedAlias]; seen {
			return
		}
		if _, seen := aliases[normalizedAlias]; seen {
			return
		}
		aliases[normalizedAlias] = normalizedModelID
	}

	appendModel := func(model openAIModelObject, modelAliases ...string) {
		modelID := strings.TrimSpace(model.ID)
		if modelID == "" {
			return
		}
		if _, seen := byID[modelID]; seen {
			return
		}
		byID[modelID] = model
		models = append(models, model)
		for _, alias := range modelAliases {
			registerAlias(alias, modelID)
		}
	}

	if configuredModel := strings.TrimSpace(server.config.Model); configuredModel != "" {
		appendModel(buildNorthboundConfiguredModel(now, server, configuredModel, inventory), buildNorthboundConfiguredModelAliases(server, configuredModel)...)
	}
	for _, provider := range inventory.Providers {
		appendModel(buildNorthboundProviderModel(now, provider), buildNorthboundProviderModelAliases(provider)...)
	}
	for _, profile := range inventory.Profiles {
		appendModel(buildNorthboundProfileModel(now, profile), buildNorthboundProfileModelAliases(profile)...)
	}

	return northboundModelCatalog{
		list: openAIModelListResponse{
			Object: "list",
			Data:   models,
		},
		byID:    byID,
		aliases: aliases,
	}
}

func (catalog northboundModelCatalog) lookup(modelID string) (openAIModelObject, bool) {
	normalizedModelID := strings.TrimSpace(modelID)
	if normalizedModelID == "" {
		return openAIModelObject{}, false
	}
	if model, ok := catalog.byID[normalizedModelID]; ok {
		return model, true
	}
	if aliasTargetID, ok := catalog.aliases[normalizedModelID]; ok {
		model, ok := catalog.byID[aliasTargetID]
		return model, ok
	}
	return openAIModelObject{}, false
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
			"source":                    "configured_default",
			"provider_id":               strings.TrimSpace(server.config.Provider),
			"provider_profile":          strings.TrimSpace(server.config.ProviderProfile),
			"northbound_routes":         []string{"/v1/chat/completions", "/v1/responses", "/v1/messages", "/v1/models"},
			"northbound_protocols":      []string{northboundProtocolChatCompletions, northboundProtocolResponses, northboundProtocolMessages},
			"inventory_kind":            "configured_default",
			"active_profile_chain":      inventory.ActiveProfileChain,
			"profile_inventory":         inventory.Profiles,
			"profile_inventory_count":   len(inventory.Profiles),
			"provider_inventory_count":  len(inventory.Providers),
			"provider_registry_version": "bridge-backed",
			"selection": buildNorthboundSelectionMetadata(northboundSelection{
				provider:        server.config.Provider,
				providerProfile: server.config.ProviderProfile,
				model:           configuredModel,
				upstreamModel:   configuredModel,
				source:          "configured_default",
				inventoryKind:   "configured_default",
			}),
		},
	}
}

func buildNorthboundConfiguredModelAliases(server *Server, configuredModel string) []string {
	aliases := make([]string, 0, 1)
	if providerProfile := strings.TrimSpace(server.config.ProviderProfile); providerProfile != "" {
		aliases = append(aliases, "configured_provider_profile:"+providerProfile)
	}
	return aliases
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
			"lookup_aliases":       buildNorthboundProviderModelAliases(provider),
		},
	}
}

func buildNorthboundProviderModelAliases(provider bridge.ProviderDescription) []string {
	providerID := strings.TrimSpace(provider.ProviderID)
	if providerID == "" {
		return nil
	}
	return []string{"provider:" + providerID}
}

func buildNorthboundProfileModel(now int64, profile bridge.ProviderProfileDescription) openAIModelObject {
	modelID := buildNorthboundProfileModelID(profile.ProviderID, profile.Profile)
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
			"northbound_routes":       profile.NorthboundRoutes,
			"northbound_protocols":    profile.NorthboundProtocols,
			"capabilities":            profile.Capabilities,
			"credential_state":        profile.CredentialState,
			"deployment_mode":         profile.DeploymentMode,
			"auth_mode":               profile.AuthMode,
			"streaming":               profile.Streaming,
			"selection":               buildNorthboundSelectionMetadata(selectionFromProfileModel(profile, modelID)),
			"lookup_aliases":          buildNorthboundProfileModelAliases(profile),
		},
	}
}

func buildNorthboundProfileModelAliases(profile bridge.ProviderProfileDescription) []string {
	providerID := strings.TrimSpace(profile.ProviderID)
	profileName := strings.TrimSpace(profile.Profile)
	if providerID == "" || profileName == "" {
		return nil
	}
	aliases := []string{"provider:" + providerID + ":profile:" + profileName}
	if providerID == "openai-compatible" {
		aliases = append(aliases, "profile:"+profileName)
	}
	return aliases
}

func buildNorthboundProfileModelID(providerID string, profile string) string {
	trimmedProfile := strings.TrimSpace(profile)
	trimmedProviderID := strings.TrimSpace(providerID)
	if trimmedProviderID == "openai-compatible" {
		if trimmedProfile == "" {
			return "profile:default"
		}
		return "profile:" + trimmedProfile
	}
	if trimmedProviderID == "" {
		if trimmedProfile == "" {
			return "profile:default"
		}
		return "profile:" + trimmedProfile
	}
	if trimmedProfile == "" {
		trimmedProfile = "default"
	}
	return "provider:" + trimmedProviderID + ":profile:" + trimmedProfile
}

func selectionFromProfileModel(profile bridge.ProviderProfileDescription, modelID string) northboundSelection {
	selection := northboundSelection{
		model:         strings.TrimSpace(modelID),
		upstreamModel: strings.TrimSpace(profile.ResolvedModel),
		source:        "provider_profile_inventory",
	}
	selection.applyProfile(profile)
	return selection
}
