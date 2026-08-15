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
	profileKey, ok := adminProviderRouteRuntimeProfileKey(environment, runtimeProfileRef)
	if !ok {
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
	return adminProviderRouteInventoryBindingFromProfile(environment, providerID, runtimeProfileRef, *matched)
}

func adminProviderRouteRuntimeProfileKey(environment string, runtimeProfileRef string) (string, bool) {
	prefix := "ref:radishmind/" + strings.TrimSpace(environment) + "/provider-profiles/"
	normalizedRef := strings.TrimSpace(runtimeProfileRef)
	profileKey := strings.TrimPrefix(normalizedRef, prefix)
	return profileKey, profileKey != normalizedRef && profileKey != ""
}

func adminProviderRouteInventoryBindingFromProfile(
	environment string,
	providerID string,
	runtimeProfileRef string,
	profile bridge.ProviderProfileDescription,
) (AdminProviderInventoryBinding, error) {
	profileKey, ok := adminProviderRouteRuntimeProfileKey(environment, runtimeProfileRef)
	if !ok || strings.TrimSpace(profile.ProviderID) != strings.TrimSpace(providerID) ||
		strings.TrimSpace(profile.Profile) != profileKey {
		return AdminProviderInventoryBinding{}, errAdminProviderRouteInventoryNotFound
	}
	capabilities := adminProviderRouteInventoryCapabilities(profile)
	northboundProtocols := append([]string{}, profile.NorthboundProtocols...)
	northboundRoutes := append([]string{}, profile.NorthboundRoutes...)
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
		Enabled               bool           `json:"enabled"`
		Capabilities          map[string]any `json:"capabilities"`
		NorthboundProtocols   []string       `json:"northbound_protocols"`
		NorthboundRoutes      []string       `json:"northbound_routes"`
		CredentialState       string         `json:"credential_state"`
		DeploymentMode        string         `json:"deployment_mode"`
		AuthMode              string         `json:"auth_mode"`
		Streaming             bool           `json:"streaming"`
	}{
		Environment: environment, Profile: strings.TrimSpace(profile.Profile),
		NormalizedProfile:     strings.TrimSpace(profile.NormalizedProfile),
		ProviderID:            strings.TrimSpace(profile.ProviderID),
		ResolvedModel:         strings.TrimSpace(profile.ResolvedModel),
		APIStyle:              strings.TrimSpace(profile.APIStyle),
		HasBaseURL:            profile.HasBaseURL,
		HasAPIKey:             profile.HasAPIKey,
		RequestTimeoutSeconds: profile.RequestTimeoutSeconds,
		Active:                profile.Active,
		Fallback:              profile.Fallback,
		ChainIndex:            profile.ChainIndex,
		Enabled:               profile.Enabled,
		Capabilities:          profile.Capabilities,
		NorthboundProtocols:   northboundProtocols,
		NorthboundRoutes:      northboundRoutes,
		CredentialState:       strings.TrimSpace(profile.CredentialState),
		DeploymentMode:        strings.TrimSpace(profile.DeploymentMode),
		AuthMode:              strings.TrimSpace(profile.AuthMode),
		Streaming:             profile.Streaming,
	})
	if err != nil {
		return AdminProviderInventoryBinding{}, errAdminProviderRouteInventoryUnavailable
	}
	return AdminProviderInventoryBinding{
		ProfileID: profileKey, ProviderID: providerID, RuntimeProfileRef: runtimeProfileRef,
		Environment: environment, Capabilities: capabilities, InventoryDigest: digest,
		Enabled: profile.Enabled,
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
