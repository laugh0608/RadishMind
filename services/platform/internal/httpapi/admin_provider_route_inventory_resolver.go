package httpapi

import (
	"context"
	"sort"
	"strings"

	"radishmind.local/services/platform/internal/bridge"
)

type bridgeAdminProviderInventoryResolver struct {
	bridge bridgeClient
}

func (resolver bridgeAdminProviderInventoryResolver) ResolveProviderProfile(
	ctx context.Context,
	environment string,
	providerID string,
	runtimeProfileRef string,
) (AdminProviderInventoryBinding, error) {
	if resolver.bridge == nil {
		return AdminProviderInventoryBinding{}, errAdminProviderRouteInventoryUnavailable
	}
	environment = strings.TrimSpace(environment)
	providerID = strings.TrimSpace(providerID)
	runtimeProfileRef = strings.TrimSpace(runtimeProfileRef)
	prefix := "ref:radishmind/" + environment + "/provider-profiles/"
	profileKey := strings.TrimPrefix(runtimeProfileRef, prefix)
	if profileKey == runtimeProfileRef || profileKey == "" {
		return AdminProviderInventoryBinding{}, errAdminProviderRouteInventoryNotFound
	}
	inventory, err := resolver.bridge.DescribeInventory(ctx)
	if err != nil {
		return AdminProviderInventoryBinding{}, errAdminProviderRouteInventoryUnavailable
	}
	var matched *bridge.ProviderProfileDescription
	for index := range inventory.Profiles {
		profile := inventory.Profiles[index]
		if strings.TrimSpace(profile.ProviderID) != providerID ||
			strings.TrimSpace(profile.Profile) != profileKey {
			continue
		}
		if matched != nil {
			return AdminProviderInventoryBinding{}, errAdminProviderRouteInventoryUnavailable
		}
		copied := profile
		matched = &copied
	}
	if matched == nil {
		return AdminProviderInventoryBinding{}, errAdminProviderRouteInventoryNotFound
	}
	capabilities := adminProviderRouteInventoryCapabilities(*matched)
	northboundProtocols := append([]string{}, matched.NorthboundProtocols...)
	northboundRoutes := append([]string{}, matched.NorthboundRoutes...)
	sort.Strings(northboundProtocols)
	sort.Strings(northboundRoutes)
	digest, err := adminProviderRouteCanonicalDigest(struct {
		Environment           string         `json:"environment"`
		Profile               string         `json:"profile"`
		NormalizedProfile     string         `json:"normalized_profile"`
		ProviderID            string         `json:"provider_id"`
		ResolvedModel         string         `json:"resolved_model"`
		APIStyle              string         `json:"api_style"`
		HasBaseURL            bool           `json:"has_base_url"`
		HasAPIKey             bool           `json:"has_api_key"`
		RequestTimeoutSeconds float64        `json:"request_timeout_seconds"`
		Active                bool           `json:"active"`
		Fallback              bool           `json:"fallback"`
		ChainIndex            int            `json:"chain_index"`
		Capabilities          map[string]any `json:"capabilities"`
		NorthboundProtocols   []string       `json:"northbound_protocols"`
		NorthboundRoutes      []string       `json:"northbound_routes"`
		CredentialState       string         `json:"credential_state"`
		DeploymentMode        string         `json:"deployment_mode"`
		AuthMode              string         `json:"auth_mode"`
		Streaming             bool           `json:"streaming"`
	}{
		Environment: environment, Profile: strings.TrimSpace(matched.Profile),
		NormalizedProfile:     strings.TrimSpace(matched.NormalizedProfile),
		ProviderID:            strings.TrimSpace(matched.ProviderID),
		ResolvedModel:         strings.TrimSpace(matched.ResolvedModel),
		APIStyle:              strings.TrimSpace(matched.APIStyle),
		HasBaseURL:            matched.HasBaseURL,
		HasAPIKey:             matched.HasAPIKey,
		RequestTimeoutSeconds: matched.RequestTimeoutSeconds,
		Active:                matched.Active,
		Fallback:              matched.Fallback,
		ChainIndex:            matched.ChainIndex,
		Capabilities:          matched.Capabilities,
		NorthboundProtocols:   northboundProtocols,
		NorthboundRoutes:      northboundRoutes,
		CredentialState:       strings.TrimSpace(matched.CredentialState),
		DeploymentMode:        strings.TrimSpace(matched.DeploymentMode),
		AuthMode:              strings.TrimSpace(matched.AuthMode),
		Streaming:             matched.Streaming,
	})
	if err != nil {
		return AdminProviderInventoryBinding{}, errAdminProviderRouteInventoryUnavailable
	}
	return AdminProviderInventoryBinding{
		ProfileID: profileKey, ProviderID: providerID, RuntimeProfileRef: runtimeProfileRef,
		Environment: environment, Capabilities: capabilities, InventoryDigest: digest,
		Enabled: matched.Active,
	}, nil
}

func adminProviderRouteInventoryCapabilities(profile bridge.ProviderProfileDescription) []string {
	values := make(map[string]bool, 3)
	for _, protocol := range profile.NorthboundProtocols {
		switch strings.TrimSpace(protocol) {
		case "chat.completions":
			values["chat_completions"] = true
		case "responses":
			values["responses"] = true
		case "messages":
			values["messages"] = true
		}
	}
	for source, target := range map[string]string{
		"chat": "chat_completions", "responses": "responses", "messages": "messages",
	} {
		if enabled, ok := profile.Capabilities[source].(bool); ok && enabled {
			values[target] = true
		}
	}
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

var _ AdminProviderInventoryResolver = bridgeAdminProviderInventoryResolver{}
