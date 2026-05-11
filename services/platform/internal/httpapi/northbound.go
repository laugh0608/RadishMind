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
}

type northboundCanonicalRequestOptions struct {
	route            string
	protocol         string
	locale           string
	promptText       string
	northboundFields map[string]any
}

func (s *Server) resolveNorthboundSelection(ctx context.Context, requestedModel string, extension *chatCompletionExtension) northboundSelection {
	provider := strings.TrimSpace(s.config.Provider)
	providerProfile := strings.TrimSpace(s.config.ProviderProfile)
	if extension != nil {
		if explicitProvider := strings.TrimSpace(extension.Provider); explicitProvider != "" {
			provider = explicitProvider
		}
		if explicitProviderProfile := strings.TrimSpace(extension.ProviderProfile); explicitProviderProfile != "" {
			providerProfile = explicitProviderProfile
		}
	}
	if provider == "" {
		provider = "mock"
	}
	if extension == nil || strings.TrimSpace(extension.Provider) == "" {
		if resolvedProvider := s.resolveKnownProvider(ctx, requestedModel); resolvedProvider != "" {
			provider = resolvedProvider
		}
	}

	model := strings.TrimSpace(requestedModel)
	if model == "" {
		model = strings.TrimSpace(s.config.Model)
	}
	if model == "" {
		model = provider
	}
	if model == "" {
		model = serviceName
	}

	return northboundSelection{
		provider:        provider,
		providerProfile: providerProfile,
		model:           model,
	}
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
