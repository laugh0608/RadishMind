package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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

type openAIChatCompletionStreamChunk struct {
	ID      string                             `json:"id"`
	Object  string                             `json:"object"`
	Created int64                              `json:"created"`
	Model   string                             `json:"model"`
	Choices []openAIChatCompletionStreamChoice `json:"choices"`
}

type openAIChatCompletionChoice struct {
	Index        int               `json:"index"`
	Message      openAIChatMessage `json:"message"`
	FinishReason string            `json:"finish_reason"`
}

type openAIChatCompletionStreamChoice struct {
	Index        int                       `json:"index"`
	Delta        openAIChatCompletionDelta `json:"delta"`
	FinishReason *string                   `json:"finish_reason"`
}

type openAIChatCompletionDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (s *Server) handleChatCompletions(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, "/v1/chat/completions")
	var chatRequest chatCompletionRequest
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&chatRequest); err != nil {
		s.writePlatformError(writer, trace, "INVALID_JSON", fmt.Sprintf("invalid chat completion request: %v", err))
		return
	}

	if len(chatRequest.Messages) == 0 {
		s.writePlatformError(writer, trace, "MISSING_MESSAGES", "messages must contain at least one item")
		return
	}

	ctx, cancel := context.WithTimeout(request.Context(), s.config.BridgeTimeout)
	defer cancel()

	locale := "zh-CN"
	if chatRequest.RadishMind != nil && strings.TrimSpace(chatRequest.RadishMind.Locale) != "" {
		locale = strings.TrimSpace(chatRequest.RadishMind.Locale)
	}

	selection := s.resolveNorthboundSelection(ctx, chatRequest.Model, chatRequest.RadishMind)
	trace.applySelection(selection)
	temperature := effectiveTemperature(chatRequest.Temperature, s.config.Temperature)

	canonicalRequest, err := buildChatCanonicalRequest(chatRequest, locale, selection, trace.requestID)
	if err != nil {
		s.writePlatformError(writer, trace, "INVALID_CHAT_MESSAGES", err.Error())
		return
	}

	if chatRequest.Stream {
		if err := s.streamOpenAIChatCompletionResponse(ctx, writer, canonicalRequest, selection, temperature, trace); err != nil {
			s.writePlatformError(writer, trace, "PLATFORM_BRIDGE_FAILED", err.Error())
		}
		return
	}

	envelope, err := s.bridge.HandleEnvelope(
		ctx,
		canonicalRequest,
		s.buildBridgeEnvelopeOptions(selection, temperature),
	)
	if err != nil {
		s.writePlatformError(writer, trace, "PLATFORM_BRIDGE_FAILED", err.Error())
		return
	}
	if strings.EqualFold(envelope.Status, "failed") {
		s.writePlatformError(writer, trace, gatewayErrorCode(envelope.Error), gatewayErrorMessage(envelope.Error))
		return
	}

	openAIResponse, err := buildOpenAIChatCompletionResponse(envelope, selection.model)
	if err != nil {
		s.writePlatformError(writer, trace, "PLATFORM_RESPONSE_INVALID", err.Error())
		return
	}
	writeObservedJSON(writer, http.StatusOK, trace, openAIResponse)
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

func buildChatCanonicalRequest(
	chatRequest chatCompletionRequest,
	locale string,
	selection northboundSelection,
	requestID string,
) ([]byte, error) {
	promptText, err := buildPromptFromMessages(chatRequest.Messages)
	if err != nil {
		return nil, err
	}
	northboundFields := map[string]any{
		"request_kind":  "chat_completion",
		"message_count": len(chatRequest.Messages),
		"stream":        chatRequest.Stream,
	}
	for key, value := range buildNorthboundSelectionFields(chatRequest.Model, selection, chatRequest.RadishMind) {
		northboundFields[key] = value
	}
	if len(chatRequest.Metadata) > 0 {
		northboundFields["metadata"] = chatRequest.Metadata
	}
	return buildNorthboundCanonicalRequest(northboundCanonicalRequestOptions{
		requestID:        requestID,
		route:            "/v1/chat/completions",
		protocol:         northboundProtocolChatCompletions,
		locale:           locale,
		promptText:       promptText,
		northboundFields: northboundFields,
	})
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

	content := buildNorthboundResponseContent(envelope)

	responseID := buildNorthboundResponseID("chatcmpl-", envelope.RequestID)

	return openAIChatCompletionResponse{
		ID:      responseID,
		Object:  "chat.completion",
		Created: timeNowUnix(),
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

func (s *Server) streamOpenAIChatCompletionResponse(
	ctx context.Context,
	writer http.ResponseWriter,
	canonicalRequest []byte,
	selection northboundSelection,
	temperature float64,
	trace requestTrace,
) error {
	responseID := buildNorthboundResponseID("chatcmpl-", trace.requestID)
	createdAt := timeNowUnix()
	streamStarted := false
	streamCompleted := false
	emittedContent := false

	startStream := func() error {
		if streamStarted {
			return nil
		}
		prepareSSEHeaders(writer)
		writeTraceHeaders(writer, trace)
		if err := writeSSEEvent(writer, "", openAIChatCompletionStreamChunk{
			ID:      responseID,
			Object:  "chat.completion.chunk",
			Created: createdAt,
			Model:   selection.model,
			Choices: []openAIChatCompletionStreamChoice{
				{
					Index: 0,
					Delta: openAIChatCompletionDelta{Role: "assistant"},
				},
			},
		}); err != nil {
			return err
		}
		streamStarted = true
		return nil
	}

	err := s.bridge.StreamEnvelope(
		ctx,
		canonicalRequest,
		s.buildBridgeEnvelopeOptions(selection, temperature),
		func(event bridge.StreamEvent) error {
			switch strings.TrimSpace(event.Type) {
			case "delta":
				if event.Delta == "" {
					return nil
				}
				if err := startStream(); err != nil {
					return err
				}
				emittedContent = true
				return writeSSEEvent(writer, "", openAIChatCompletionStreamChunk{
					ID:      responseID,
					Object:  "chat.completion.chunk",
					Created: createdAt,
					Model:   selection.model,
					Choices: []openAIChatCompletionStreamChoice{
						{
							Index: 0,
							Delta: openAIChatCompletionDelta{Content: event.Delta},
						},
					},
				})
			case "completed":
				streamCompleted = true
				if event.Envelope != nil && strings.EqualFold(event.Envelope.Status, "failed") && !emittedContent {
					return fmt.Errorf(gatewayErrorMessage(event.Envelope.Error))
				}
				return nil
			case "error":
				return fmt.Errorf(gatewayErrorMessage(event.Error))
			default:
				return nil
			}
		},
	)
	if err != nil {
		return err
	}
	if !streamCompleted {
		return fmt.Errorf("platform bridge stream ended without completion event")
	}
	if err := startStream(); err != nil {
		return err
	}

	finishReason := "stop"
	if err := writeSSEEvent(writer, "", openAIChatCompletionStreamChunk{
		ID:      responseID,
		Object:  "chat.completion.chunk",
		Created: createdAt,
		Model:   selection.model,
		Choices: []openAIChatCompletionStreamChoice{
			{
				Index:        0,
				Delta:        openAIChatCompletionDelta{},
				FinishReason: &finishReason,
			},
		},
	}); err != nil {
		return err
	}

	if err := writeSSEEvent(writer, "", "[DONE]"); err != nil {
		return err
	}
	logRequestTrace(trace, http.StatusOK, "", "")
	return nil
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
