package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/bridge"
)

const (
	northboundProtocolChatCompletions = "openai-chat-completions"
	northboundProtocolResponses       = "openai-responses"
	northboundProtocolMessages        = "anthropic-messages"
)

type northboundSelection struct {
	provider        string
	providerProfile string
	model           string
	upstreamModel   string
	source          string
}

type northboundCanonicalRequestOptions struct {
	route            string
	protocol         string
	locale           string
	promptText       string
	northboundFields map[string]any
}

type northboundInventoryLookup struct {
	providers                map[string]bridge.ProviderDescription
	profilesByModelID        map[string]bridge.ProviderProfileDescription
	profilesByQualifiedAlias map[string]bridge.ProviderProfileDescription
	profilesByPlainName      map[string]bridge.ProviderProfileDescription
	activeProfileByProvider  map[string]string
}

func buildNorthboundInventoryLookup(inventory bridge.ProviderInventory) northboundInventoryLookup {
	lookup := northboundInventoryLookup{
		providers:                make(map[string]bridge.ProviderDescription, len(inventory.Providers)),
		profilesByModelID:        make(map[string]bridge.ProviderProfileDescription, len(inventory.Profiles)),
		profilesByQualifiedAlias: make(map[string]bridge.ProviderProfileDescription, len(inventory.Profiles)),
		profilesByPlainName:      make(map[string]bridge.ProviderProfileDescription, len(inventory.Profiles)),
		activeProfileByProvider:  make(map[string]string, len(inventory.Profiles)),
	}
	fallbackProfileByProvider := make(map[string]string, len(inventory.Profiles))
	for _, provider := range inventory.Providers {
		normalizedProviderID := strings.TrimSpace(provider.ProviderID)
		if normalizedProviderID == "" {
			continue
		}
		lookup.providers[normalizedProviderID] = provider
	}
	for _, profile := range inventory.Profiles {
		providerID := strings.TrimSpace(profile.ProviderID)
		profileName := strings.TrimSpace(profile.Profile)
		if providerID == "" || profileName == "" {
			continue
		}
		modelID := buildNorthboundProfileModelID(providerID, profileName)
		lookup.profilesByModelID[modelID] = profile
		lookup.profilesByQualifiedAlias["provider:"+providerID+":profile:"+profileName] = profile
		if existing, ok := lookup.profilesByPlainName[profileName]; !ok {
			lookup.profilesByPlainName[profileName] = profile
		} else if existing.ProviderID == "openai-compatible" {
			lookup.profilesByPlainName[profileName] = existing
		} else if providerID == "openai-compatible" {
			lookup.profilesByPlainName[profileName] = profile
		} else if existing.ProviderID != providerID {
			delete(lookup.profilesByPlainName, profileName)
		}
		if _, ok := fallbackProfileByProvider[providerID]; !ok {
			fallbackProfileByProvider[providerID] = profileName
		}
		if profile.Active {
			lookup.activeProfileByProvider[providerID] = profileName
		}
	}
	for providerID, profileName := range fallbackProfileByProvider {
		if _, ok := lookup.activeProfileByProvider[providerID]; !ok {
			lookup.activeProfileByProvider[providerID] = profileName
		}
	}
	return lookup
}

func parseNorthboundProviderProfileModel(model string) (string, string, bool) {
	normalizedModel := strings.TrimSpace(model)
	if !strings.HasPrefix(normalizedModel, "provider:") {
		return "", "", false
	}
	rest := strings.TrimPrefix(normalizedModel, "provider:")
	parts := strings.SplitN(rest, ":profile:", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	providerID := strings.TrimSpace(parts[0])
	profileName := strings.TrimSpace(parts[1])
	if providerID == "" || profileName == "" {
		return "", "", false
	}
	return providerID, profileName, true
}

func parseNorthboundProviderAlias(model string) (string, bool) {
	normalizedModel := strings.TrimSpace(model)
	if !strings.HasPrefix(normalizedModel, "provider:") {
		return "", false
	}
	if _, _, ok := parseNorthboundProviderProfileModel(normalizedModel); ok {
		return "", false
	}
	providerID := strings.TrimSpace(strings.TrimPrefix(normalizedModel, "provider:"))
	if providerID == "" {
		return "", false
	}
	return providerID, true
}

func lookupNorthboundProfile(lookup northboundInventoryLookup, model string) (bridge.ProviderProfileDescription, bool) {
	normalizedModel := strings.TrimSpace(model)
	if normalizedModel == "" {
		return bridge.ProviderProfileDescription{}, false
	}
	if profile, ok := lookup.profilesByModelID[normalizedModel]; ok {
		return profile, true
	}
	if profile, ok := lookup.profilesByQualifiedAlias[normalizedModel]; ok {
		return profile, true
	}
	if strings.HasPrefix(normalizedModel, "profile:") {
		profileName := strings.TrimPrefix(normalizedModel, "profile:")
		if profile, ok := lookup.profilesByPlainName[profileName]; ok {
			return profile, true
		}
	}
	return bridge.ProviderProfileDescription{}, false
}

func (s *Server) resolveNorthboundSelection(ctx context.Context, requestedModel string, extension *chatCompletionExtension) northboundSelection {
	requestedModel = strings.TrimSpace(requestedModel)
	configuredProvider := strings.TrimSpace(s.config.Provider)
	if configuredProvider == "" {
		configuredProvider = "mock"
	}
	configuredProfile := strings.TrimSpace(s.config.ProviderProfile)
	configuredModel := strings.TrimSpace(s.config.Model)
	selection := northboundSelection{
		provider:        configuredProvider,
		providerProfile: configuredProfile,
		model:           configuredModel,
		upstreamModel:   configuredModel,
		source:          "configured_default",
	}
	if selection.model == "" {
		if selection.providerProfile != "" {
			selection.model = buildNorthboundProfileModelID(selection.provider, selection.providerProfile)
		} else {
			selection.model = selection.provider
		}
	}

	inventory, err := s.bridge.DescribeInventory(ctx)
	inventoryLookup := northboundInventoryLookup{}
	if err == nil {
		inventoryLookup = buildNorthboundInventoryLookup(inventory)
	}

	requestedConcreteModel := requestedModel != ""
	requestedProfileMatched := false
	requestedProviderAlias := ""
	if requestedModel != "" {
		if providerID, profileName, ok := parseNorthboundProviderProfileModel(requestedModel); ok {
			requestedConcreteModel = false
			requestedProfileMatched = true
			selection.provider = providerID
			selection.providerProfile = profileName
			selection.model = requestedModel
			selection.source = "requested_provider_profile_model"
			if profile, ok := lookupNorthboundProfile(inventoryLookup, requestedModel); ok {
				selection.provider = strings.TrimSpace(profile.ProviderID)
				selection.providerProfile = strings.TrimSpace(profile.Profile)
				selection.upstreamModel = strings.TrimSpace(profile.ResolvedModel)
				selection.source = "requested_provider_profile_model+inventory"
			}
		} else if profile, ok := lookupNorthboundProfile(inventoryLookup, requestedModel); ok {
			requestedConcreteModel = false
			requestedProfileMatched = true
			selection.provider = strings.TrimSpace(profile.ProviderID)
			selection.providerProfile = strings.TrimSpace(profile.Profile)
			selection.model = requestedModel
			selection.upstreamModel = strings.TrimSpace(profile.ResolvedModel)
			selection.source = "requested_profile_model"
		} else if providerID, ok := parseNorthboundProviderAlias(requestedModel); ok {
			requestedConcreteModel = false
			requestedProviderAlias = providerID
			selection.provider = providerID
			selection.providerProfile = ""
			selection.model = requestedModel
			selection.upstreamModel = ""
			selection.source = "requested_provider_model"
			if activeProfile, ok := inventoryLookup.activeProfileByProvider[providerID]; ok && activeProfile != "" {
				selection.providerProfile = activeProfile
				selection.model = buildNorthboundProfileModelID(providerID, activeProfile)
				if profile, ok := lookupNorthboundProfile(inventoryLookup, selection.model); ok {
					selection.upstreamModel = strings.TrimSpace(profile.ResolvedModel)
				}
				selection.source = "requested_provider_model+inventory"
			}
		} else if _, ok := inventoryLookup.providers[requestedModel]; ok {
			requestedConcreteModel = false
			requestedProviderAlias = requestedModel
			selection.provider = requestedModel
			selection.providerProfile = ""
			selection.model = requestedModel
			selection.upstreamModel = ""
			selection.source = "requested_provider_model"
			if activeProfile, ok := inventoryLookup.activeProfileByProvider[requestedModel]; ok && activeProfile != "" {
				selection.providerProfile = activeProfile
				selection.model = buildNorthboundProfileModelID(requestedModel, activeProfile)
				if profile, ok := lookupNorthboundProfile(inventoryLookup, selection.model); ok {
					selection.upstreamModel = strings.TrimSpace(profile.ResolvedModel)
				}
				selection.source = "requested_provider_model+inventory"
			}
		} else {
			selection.model = requestedModel
			selection.upstreamModel = requestedModel
			selection.source = "requested_concrete_model"
		}
	}

	explicitProvider := ""
	explicitProfile := ""
	if extension != nil {
		explicitProvider = strings.TrimSpace(extension.Provider)
		explicitProfile = strings.TrimSpace(extension.ProviderProfile)
		if explicitProvider != "" {
			selection.provider = explicitProvider
			selection.source = "radishmind.provider"
		}
		if explicitProfile != "" {
			selection.providerProfile = explicitProfile
			selection.source = "radishmind.provider_profile"
		}
	}

	if explicitProvider != "" {
		selection.provider = explicitProvider
		selection.source = "radishmind.provider"
		if explicitProfile == "" {
			selection.providerProfile = ""
		}
		if !requestedConcreteModel {
			if activeProfile, ok := inventoryLookup.activeProfileByProvider[explicitProvider]; ok && activeProfile != "" {
				selection.providerProfile = activeProfile
				selection.model = buildNorthboundProfileModelID(explicitProvider, activeProfile)
				if profile, ok := lookupNorthboundProfile(inventoryLookup, selection.model); ok {
					selection.upstreamModel = strings.TrimSpace(profile.ResolvedModel)
				}
			} else if selection.providerProfile == "" {
				selection.model = explicitProvider
			}
		}
	}

	if explicitProfile != "" {
		selection.providerProfile = explicitProfile
		selection.source = "radishmind.provider_profile"
		if profile, ok := lookupNorthboundProfile(inventoryLookup, buildNorthboundProfileModelID(selection.provider, explicitProfile)); ok {
			selection.provider = strings.TrimSpace(profile.ProviderID)
			selection.providerProfile = strings.TrimSpace(profile.Profile)
			selection.model = buildNorthboundProfileModelID(selection.provider, selection.providerProfile)
			if !requestedConcreteModel {
				selection.upstreamModel = strings.TrimSpace(profile.ResolvedModel)
			}
			selection.source = "radishmind.provider_profile+inventory"
		} else if profile, ok := lookupNorthboundProfile(inventoryLookup, "profile:"+explicitProfile); ok {
			selection.provider = strings.TrimSpace(profile.ProviderID)
			selection.providerProfile = strings.TrimSpace(profile.Profile)
			selection.model = buildNorthboundProfileModelID(selection.provider, selection.providerProfile)
			if !requestedConcreteModel {
				selection.upstreamModel = strings.TrimSpace(profile.ResolvedModel)
			}
			selection.source = "radishmind.provider_profile+inventory"
		} else if selection.model == "" || selection.model == selection.provider {
			selection.model = buildNorthboundProfileModelID(selection.provider, selection.providerProfile)
		}
	}

	if explicitProvider == "" && explicitProfile == "" && !requestedProfileMatched && requestedProviderAlias == "" && requestedModel == "" {
		if selection.providerProfile == "" {
			if activeProfile, ok := inventoryLookup.activeProfileByProvider[selection.provider]; ok && activeProfile != "" {
				selection.providerProfile = activeProfile
				selection.model = buildNorthboundProfileModelID(selection.provider, activeProfile)
				if profile, ok := lookupNorthboundProfile(inventoryLookup, selection.model); ok {
					selection.upstreamModel = strings.TrimSpace(profile.ResolvedModel)
				}
			} else if selection.model == "" {
				selection.model = selection.provider
			}
		} else if selection.provider == "openai-compatible" {
			selection.model = buildNorthboundProfileModelID(selection.provider, selection.providerProfile)
		}
	}

	if selection.providerProfile != "" {
		if profile, ok := lookupNorthboundProfile(inventoryLookup, buildNorthboundProfileModelID(selection.provider, selection.providerProfile)); ok {
			selection.provider = strings.TrimSpace(profile.ProviderID)
			if !requestedConcreteModel {
				selection.upstreamModel = strings.TrimSpace(profile.ResolvedModel)
			}
			if selection.model == "" || selection.model == selection.provider {
				selection.model = buildNorthboundProfileModelID(selection.provider, selection.providerProfile)
			}
		}
	}

	if selection.upstreamModel == "" {
		if requestedConcreteModel {
			selection.upstreamModel = requestedModel
		} else if selection.provider == configuredProvider && selection.providerProfile == configuredProfile && configuredModel != "" {
			selection.upstreamModel = configuredModel
		} else if profile, ok := lookupNorthboundProfile(inventoryLookup, buildNorthboundProfileModelID(selection.provider, selection.providerProfile)); ok {
			selection.upstreamModel = strings.TrimSpace(profile.ResolvedModel)
		}
	}

	if selection.model == "" {
		if selection.providerProfile != "" {
			selection.model = buildNorthboundProfileModelID(selection.provider, selection.providerProfile)
		} else {
			selection.model = selection.provider
		}
	}

	return selection
}

func buildNorthboundCanonicalRequest(options northboundCanonicalRequestOptions) ([]byte, error) {
	northboundContext := map[string]any{
		"protocol": options.protocol,
	}
	for key, value := range options.northboundFields {
		if key == "" {
			continue
		}
		northboundContext[key] = value
	}

	canonicalRequest := map[string]any{
		"schema_version": 1,
		"project":        "radish",
		"task":           "answer_docs_question",
		"locale":         options.locale,
		"artifacts": []map[string]any{
			{
				"kind":      "text",
				"role":      "primary",
				"name":      "northbound_prompt",
				"mime_type": "text/plain",
				"content":   options.promptText,
			},
		},
		"context": map[string]any{
			"current_app": "radishmind-platform",
			"route":       options.route,
			"resource": map[string]any{
				"kind":  "northbound_request",
				"slug":  routeSlug(options.route),
				"title": truncateForTitle(options.promptText),
			},
			"northbound": northboundContext,
		},
		"tool_hints": map[string]any{
			"allow_retrieval":       false,
			"allow_tool_calls":      false,
			"allow_image_reasoning": false,
		},
		"safety": map[string]any{
			"mode":                              "advisory",
			"requires_confirmation_for_actions": false,
		},
	}
	return json.Marshal(canonicalRequest)
}

func buildNorthboundSelectionFields(
	requestedModel string,
	selection northboundSelection,
	extension *chatCompletionExtension,
) map[string]any {
	fields := map[string]any{
		"requested_model":           strings.TrimSpace(requestedModel),
		"selected_provider":         strings.TrimSpace(selection.provider),
		"selected_provider_profile": strings.TrimSpace(selection.providerProfile),
		"selected_model":            strings.TrimSpace(selection.model),
		"upstream_model":            strings.TrimSpace(selection.upstreamModel),
		"selection_source":          strings.TrimSpace(selection.source),
	}
	if extension != nil {
		if provider := strings.TrimSpace(extension.Provider); provider != "" {
			fields["requested_provider"] = provider
		}
		if providerProfile := strings.TrimSpace(extension.ProviderProfile); providerProfile != "" {
			fields["requested_provider_profile"] = providerProfile
		}
		if locale := strings.TrimSpace(extension.Locale); locale != "" {
			fields["locale"] = locale
		}
		if conversationID := strings.TrimSpace(extension.ConversationID); conversationID != "" {
			fields["conversation_id"] = conversationID
		}
	}
	return fields
}

func buildNorthboundPromptText(sections ...string) string {
	filtered := make([]string, 0, len(sections))
	for _, section := range sections {
		if trimmed := strings.TrimSpace(section); trimmed != "" {
			filtered = append(filtered, trimmed)
		}
	}
	return strings.Join(filtered, "\n\n")
}

func buildConversationTranscript(messages []chatCompletionMessage) (string, error) {
	transcript := make([]string, 0, len(messages))
	for _, message := range messages {
		text, err := extractMessageText(message.Content)
		if err != nil {
			return "", err
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		role := strings.TrimSpace(message.Role)
		if role == "" {
			role = "message"
		}
		if name := strings.TrimSpace(message.Name); name != "" {
			role = fmt.Sprintf("%s (%s)", role, name)
		}
		transcript = append(transcript, fmt.Sprintf("%s: %s", role, text))
	}
	return strings.Join(transcript, "\n"), nil
}

func buildNorthboundTextFromValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(typed)
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			if itemMap, ok := item.(map[string]any); ok {
				role := firstNonEmptyString(itemMap, "role", "type")
				if role != "" {
					if content := strings.TrimSpace(normalizeAnyText(itemMap["content"])); content != "" {
						parts = append(parts, fmt.Sprintf("%s: %s", role, content))
						continue
					}
				}
				if text := strings.TrimSpace(normalizeAnyText(itemMap["text"])); text != "" {
					parts = append(parts, text)
					continue
				}
			}
			if text := strings.TrimSpace(normalizeAnyText(item)); text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "\n")
	case map[string]any:
		if messages, ok := typed["messages"].([]any); ok {
			return buildNorthboundTextFromValue(messages)
		}
		if parts, ok := typed["parts"].([]any); ok {
			return buildNorthboundTextFromValue(parts)
		}
		if content := strings.TrimSpace(normalizeAnyText(typed["content"])); content != "" {
			return content
		}
		if text := strings.TrimSpace(normalizeAnyText(typed["text"])); text != "" {
			return text
		}
		if input, ok := typed["input"]; ok {
			return buildNorthboundTextFromValue(input)
		}
		return strings.TrimSpace(normalizeAnyText(typed))
	default:
		return strings.TrimSpace(normalizeAnyText(typed))
	}
}

func firstNonEmptyString(document map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := document[key].(string); ok {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}

func (s *Server) buildBridgeEnvelopeOptions(selection northboundSelection, temperature float64) bridge.EnvelopeOptions {
	options := bridge.EnvelopeOptions{
		Provider:        strings.TrimSpace(selection.provider),
		ProviderProfile: strings.TrimSpace(selection.providerProfile),
		Model:           strings.TrimSpace(selection.upstreamModel),
		Temperature:     temperature,
		RequestTimeout:  s.config.BridgeTimeout,
	}

	configuredProvider := strings.TrimSpace(s.config.Provider)
	if configuredProvider == "" {
		configuredProvider = "mock"
	}
	configuredProfile := strings.TrimSpace(s.config.ProviderProfile)
	if options.Provider == configuredProvider && (configuredProvider != "openai-compatible" || options.ProviderProfile == configuredProfile) {
		options.BaseURL = strings.TrimSpace(s.config.BaseURL)
		options.APIKey = strings.TrimSpace(s.config.APIKey)
		if options.Model == "" {
			options.Model = strings.TrimSpace(s.config.Model)
		}
	}
	return options
}

func buildNorthboundResponseContent(envelope bridge.GatewayEnvelope) string {
	if envelope.Response == nil {
		return "RadishMind platform bridge returned an empty response."
	}
	content := strings.TrimSpace(normalizeAnyText(envelope.Response["summary"]))
	if content == "" {
		if answers, ok := envelope.Response["answers"].([]any); ok && len(answers) > 0 {
			if firstAnswer, ok := answers[0].(map[string]any); ok {
				content = strings.TrimSpace(normalizeAnyText(firstAnswer["text"]))
			}
		}
	}
	if content == "" {
		content = "RadishMind platform bridge returned an empty response."
	}
	return content
}

func buildNorthboundResponseID(prefix string, requestID string) string {
	normalizedRequestID := strings.TrimSpace(requestID)
	if normalizedRequestID == "" {
		normalizedRequestID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if strings.HasPrefix(normalizedRequestID, prefix) {
		return normalizedRequestID
	}
	return prefix + normalizedRequestID
}

func resolveNorthboundLocale(extension *chatCompletionExtension) string {
	if extension != nil {
		if locale := strings.TrimSpace(extension.Locale); locale != "" {
			return locale
		}
	}
	return "zh-CN"
}

func timeNowUnix() int64 {
	return time.Now().Unix()
}

func routeSlug(route string) string {
	return strings.NewReplacer("/", "-", " ", "-", "_", "-").Replace(strings.TrimSpace(route))
}
