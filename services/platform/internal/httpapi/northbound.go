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

func (s *Server) resolveNorthboundSelection(ctx context.Context, requestedModel string, extension *chatCompletionExtension) northboundSelection {
	requestedModel = strings.TrimSpace(requestedModel)
	selection := northboundSelection{
		provider:        strings.TrimSpace(s.config.Provider),
		providerProfile: strings.TrimSpace(s.config.ProviderProfile),
		model:           requestedModel,
		upstreamModel:   strings.TrimSpace(s.config.Model),
		source:          "configured_default",
	}
	if selection.provider == "" {
		selection.provider = "mock"
	}
	if selection.model == "" {
		selection.model = selection.provider
	}
	if selection.upstreamModel == "" {
		selection.upstreamModel = selection.model
	}

	inventory, err := s.bridge.DescribeInventory(ctx)
	if err != nil {
		return selection
	}

	providers := make(map[string]bridge.ProviderDescription, len(inventory.Providers))
	for _, provider := range inventory.Providers {
		normalizedProviderID := strings.TrimSpace(provider.ProviderID)
		if normalizedProviderID == "" {
			continue
		}
		providers[normalizedProviderID] = provider
	}

	profiles := make(map[string]bridge.ProviderProfileDescription, len(inventory.Profiles))
	activeProfile := ""
	for _, profile := range inventory.Profiles {
		normalizedProfile := strings.TrimSpace(profile.Profile)
		if normalizedProfile != "" {
			profiles[normalizedProfile] = profile
		}
		if activeProfile == "" && profile.Active {
			activeProfile = normalizedProfile
		}
	}
	if activeProfile == "" && len(inventory.ActiveProfileChain) > 0 {
		activeProfile = strings.TrimSpace(inventory.ActiveProfileChain[0])
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

	if explicitProfile != "" {
		if profile, ok := profiles[explicitProfile]; ok {
			selection.provider = strings.TrimSpace(profile.ProviderID)
			selection.upstreamModel = strings.TrimSpace(profile.ResolvedModel)
			selection.source = "radishmind.provider_profile+inventory"
		}
	}
	if explicitProvider != "" && explicitProfile == "" && selection.provider != "openai-compatible" {
		selection.providerProfile = ""
	}

	if explicitProvider == "" && explicitProfile == "" {
		if requestedProfile := strings.TrimPrefix(requestedModel, "profile:"); requestedProfile != requestedModel && requestedProfile != "" {
			if profile, ok := profiles[requestedProfile]; ok {
				selection.provider = strings.TrimSpace(profile.ProviderID)
				selection.providerProfile = strings.TrimSpace(profile.Profile)
				selection.upstreamModel = strings.TrimSpace(profile.ResolvedModel)
				selection.model = requestedModel
				selection.source = "requested_profile_model"
			}
		}
	}

	if selection.providerProfile != "" && (selection.provider == "openai-compatible" || explicitProfile != "") {
		if profile, ok := profiles[selection.providerProfile]; ok {
			selection.provider = strings.TrimSpace(profile.ProviderID)
			if selection.upstreamModel == "" || selection.upstreamModel == selection.model {
				selection.upstreamModel = strings.TrimSpace(profile.ResolvedModel)
			}
		}
	}

	if explicitProvider == "" && explicitProfile == "" {
		if _, ok := providers[selection.model]; ok {
			selection.provider = selection.model
			selection.source = "requested_provider_model"
			if selection.providerProfile == "" && selection.provider == "openai-compatible" {
				if activeProfile != "" {
					selection.providerProfile = activeProfile
				}
			}
		}
	}
	if explicitProfile == "" && selection.provider != "openai-compatible" {
		selection.providerProfile = ""
	}

	if selection.provider == "openai-compatible" && selection.providerProfile == "" {
		if activeProfile != "" {
			selection.providerProfile = activeProfile
		}
	}
	if selection.providerProfile != "" {
		if profile, ok := profiles[selection.providerProfile]; ok {
			if selection.upstreamModel == "" || selection.upstreamModel == selection.model {
				selection.upstreamModel = strings.TrimSpace(profile.ResolvedModel)
			}
		}
	}
	if selection.upstreamModel == "" {
		selection.upstreamModel = selection.model
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
