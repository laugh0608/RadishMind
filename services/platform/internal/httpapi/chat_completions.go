package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/bridge"
)

type chatCompletionRequest struct {
	Model       string                   `json:"model"`
	Messages    []chatCompletionMessage  `json:"messages"`
	Stream      bool                     `json:"stream,omitempty"`
	Temperature *float64                 `json:"temperature,omitempty"`
	RadishMind  *chatCompletionExtension `json:"radishmind,omitempty"`
	Metadata    map[string]any           `json:"metadata,omitempty"`
}

type chatCompletionMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
	Name    string          `json:"name,omitempty"`
}

type chatCompletionExtension struct {
	Locale          string `json:"locale,omitempty"`
	ConversationID  string `json:"conversation_id,omitempty"`
	Provider        string `json:"provider,omitempty"`
	ProviderProfile string `json:"provider_profile,omitempty"`
}

type openAIChatCompletionResponse struct {
	ID      string                       `json:"id"`
	Object  string                       `json:"object"`
	Created int64                        `json:"created"`
	Model   string                       `json:"model"`
	Choices []openAIChatCompletionChoice `json:"choices"`
}

type openAIChatCompletionChoice struct {
	Index        int               `json:"index"`
	Message      openAIChatMessage `json:"message"`
	FinishReason string            `json:"finish_reason"`
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (s *Server) handleChatCompletions(writer http.ResponseWriter, request *http.Request) {
	var chatRequest chatCompletionRequest
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&chatRequest); err != nil {
		writeOpenAIError(writer, http.StatusBadRequest, "INVALID_JSON", fmt.Sprintf("invalid chat completion request: %v", err))
		return
	}

	if chatRequest.Stream {
		writeOpenAIError(writer, http.StatusNotImplemented, "STREAMING_NOT_IMPLEMENTED", "streaming chat completions are not implemented yet")
		return
	}
	if len(chatRequest.Messages) == 0 {
		writeOpenAIError(writer, http.StatusBadRequest, "MISSING_MESSAGES", "messages must contain at least one item")
		return
	}

	ctx, cancel := context.WithTimeout(request.Context(), s.config.BridgeTimeout)
	defer cancel()

	locale := "zh-CN"
	if chatRequest.RadishMind != nil && strings.TrimSpace(chatRequest.RadishMind.Locale) != "" {
		locale = strings.TrimSpace(chatRequest.RadishMind.Locale)
	}

	provider := s.config.Provider
	if chatRequest.RadishMind != nil && strings.TrimSpace(chatRequest.RadishMind.Provider) != "" {
		provider = strings.TrimSpace(chatRequest.RadishMind.Provider)
	} else if modelProvider := s.resolveKnownProvider(ctx, chatRequest.Model); modelProvider != "" {
		provider = modelProvider
	}

	providerProfile := s.config.ProviderProfile
	if chatRequest.RadishMind != nil && strings.TrimSpace(chatRequest.RadishMind.ProviderProfile) != "" {
		providerProfile = strings.TrimSpace(chatRequest.RadishMind.ProviderProfile)
	}

	canonicalRequest, err := buildChatCanonicalRequest(chatRequest, locale)
	if err != nil {
		writeOpenAIError(writer, http.StatusBadRequest, "INVALID_CHAT_MESSAGES", err.Error())
		return
	}

	envelope, err := s.bridge.HandleEnvelope(
		ctx,
		canonicalRequest,
		bridge.EnvelopeOptions{
			Provider:        provider,
			ProviderProfile: providerProfile,
			Model:           s.config.Model,
			BaseURL:         s.config.BaseURL,
			APIKey:          s.config.APIKey,
			Temperature:     effectiveTemperature(chatRequest.Temperature, s.config.Temperature),
			RequestTimeout:  s.config.BridgeTimeout,
		},
	)
	if err != nil {
		writeOpenAIError(writer, http.StatusBadGateway, "PLATFORM_BRIDGE_FAILED", err.Error())
		return
	}
	if strings.EqualFold(envelope.Status, "failed") {
		writeOpenAIError(writer, gatewayStatusToHTTPStatus(envelope.Error), gatewayErrorCode(envelope.Error), gatewayErrorMessage(envelope.Error))
		return
	}

	responseModel := provider
	if responseModel == "" {
		responseModel = s.config.Provider
	}
	openAIResponse, err := buildOpenAIChatCompletionResponse(envelope, responseModel)
	if err != nil {
		writeOpenAIError(writer, http.StatusBadGateway, "PLATFORM_RESPONSE_INVALID", err.Error())
		return
	}
	writeJSON(writer, http.StatusOK, openAIResponse)
}

func (s *Server) resolveKnownProvider(ctx context.Context, requestedModel string) string {
	normalizedRequestedModel := strings.TrimSpace(requestedModel)
	if normalizedRequestedModel == "" {
		return ""
	}

	descriptions, err := s.bridge.DescribeProviders(ctx)
	if err != nil {
		return ""
	}
	for _, description := range descriptions {
		if strings.TrimSpace(description.ProviderID) == normalizedRequestedModel {
			return normalizedRequestedModel
		}
	}
	return ""
}

func buildChatCanonicalRequest(chatRequest chatCompletionRequest, locale string) ([]byte, error) {
	promptText, err := buildPromptFromMessages(chatRequest.Messages)
	if err != nil {
		return nil, err
	}

	canonicalRequest := map[string]any{
		"schema_version": 1,
		"project":        "radish",
		"task":           "answer_docs_question",
		"locale":         locale,
		"artifacts": []map[string]any{
			{
				"kind":      "text",
				"role":      "primary",
				"name":      "chat_prompt",
				"mime_type": "text/plain",
				"content":   promptText,
			},
		},
		"context": map[string]any{
			"current_app": "radishmind-platform",
			"route":       "/v1/chat/completions",
			"resource": map[string]any{
				"kind":  "chat_completion",
				"slug":  "openai-chat-completion",
				"title": truncateForTitle(promptText),
			},
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

func buildPromptFromMessages(messages []chatCompletionMessage) (string, error) {
	var userText string
	for index := len(messages) - 1; index >= 0; index-- {
		message := messages[index]
		if strings.TrimSpace(message.Role) != "user" {
			continue
		}
		text, err := extractMessageText(message.Content)
		if err != nil {
			return "", err
		}
		text = strings.TrimSpace(text)
		if text != "" {
			userText = text
			break
		}
	}
	if userText == "" {
		for index := len(messages) - 1; index >= 0; index-- {
			text, err := extractMessageText(messages[index].Content)
			if err != nil {
				return "", err
			}
			text = strings.TrimSpace(text)
			if text != "" {
				userText = text
				break
			}
		}
	}
	if userText == "" {
		return "", fmt.Errorf("no textual user content was found")
	}
	return userText, nil
}

func extractMessageText(raw json.RawMessage) (string, error) {
	if len(raw) == 0 {
		return "", nil
	}

	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text, nil
	}

	var parts []map[string]any
	if err := json.Unmarshal(raw, &parts); err == nil {
		var builder strings.Builder
		for _, part := range parts {
			if part == nil {
				continue
			}
			contentType, _ := part["type"].(string)
			if contentType != "" && contentType != "text" {
				continue
			}
			if partText := strings.TrimSpace(normalizeAnyText(part["text"])); partText != "" {
				if builder.Len() > 0 {
					builder.WriteString("\n")
				}
				builder.WriteString(partText)
			}
		}
		return builder.String(), nil
	}

	var structured any
	if err := json.Unmarshal(raw, &structured); err == nil {
		return normalizeAnyText(structured), nil
	}

	return "", fmt.Errorf("unsupported message content format")
}

func normalizeAnyText(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case map[string]any:
		if text, ok := typed["text"].(string); ok {
			return strings.TrimSpace(text)
		}
		if content, ok := typed["content"].(string); ok {
			return strings.TrimSpace(content)
		}
	case []any:
		var parts []string
		for _, item := range typed {
			if partText := normalizeAnyText(item); partText != "" {
				parts = append(parts, partText)
			}
		}
		return strings.Join(parts, "\n")
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func buildOpenAIChatCompletionResponse(envelope bridge.GatewayEnvelope, model string) (openAIChatCompletionResponse, error) {
	if envelope.Response == nil {
		return openAIChatCompletionResponse{}, fmt.Errorf("gateway envelope is missing response payload")
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

	responseID := envelope.RequestID
	if responseID == "" {
		responseID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if !strings.HasPrefix(responseID, "chatcmpl-") {
		responseID = "chatcmpl-" + responseID
	}

	return openAIChatCompletionResponse{
		ID:      responseID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []openAIChatCompletionChoice{
			{
				Index: 0,
				Message: openAIChatMessage{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: "stop",
			},
		},
	}, nil
}

func effectiveTemperature(requestTemperature *float64, fallback float64) float64 {
	if requestTemperature != nil {
		return *requestTemperature
	}
	return fallback
}

func truncateForTitle(text string) string {
	normalized := strings.TrimSpace(text)
	if normalized == "" {
		return "chat completion"
	}
	if len(normalized) <= 80 {
		return normalized
	}
	return normalized[:79]
}

func gatewayStatusToHTTPStatus(err *bridge.GatewayError) int {
	if err == nil {
		return http.StatusBadGateway
	}
	switch err.Code {
	case "REQUEST_SCHEMA_INVALID":
		return http.StatusBadRequest
	case "UNSUPPORTED_PROVIDER":
		return http.StatusBadGateway
	case "UNSUPPORTED_TASK":
		return http.StatusBadGateway
	default:
		return http.StatusBadGateway
	}
}

func gatewayErrorCode(err *bridge.GatewayError) string {
	if err == nil || strings.TrimSpace(err.Code) == "" {
		return "PLATFORM_GATEWAY_FAILED"
	}
	return strings.TrimSpace(err.Code)
}

func gatewayErrorMessage(err *bridge.GatewayError) string {
	if err == nil || strings.TrimSpace(err.Message) == "" {
		return "platform gateway failed"
	}
	return strings.TrimSpace(err.Message)
}
