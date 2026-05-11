package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"radishmind.local/services/platform/internal/bridge"
)

type anthropicMessagesRequest struct {
	Model       string                   `json:"model"`
	System      any                      `json:"system,omitempty"`
	Messages    []chatCompletionMessage  `json:"messages"`
	Stream      bool                     `json:"stream,omitempty"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
	Temperature *float64                 `json:"temperature,omitempty"`
	RadishMind  *chatCompletionExtension `json:"radishmind,omitempty"`
	Metadata    map[string]any           `json:"metadata,omitempty"`
}

type anthropicMessagesResponse struct {
	ID           string                    `json:"id"`
	Type         string                    `json:"type"`
	Role         string                    `json:"role"`
	Content      []anthropicMessageContent `json:"content"`
	Model        string                    `json:"model"`
	StopReason   string                    `json:"stop_reason"`
	StopSequence any                       `json:"stop_sequence"`
	Usage        anthropicMessagesUsage    `json:"usage"`
}

type anthropicMessageContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicMessagesUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicMessageStartEvent struct {
	Type    string                    `json:"type"`
	Message anthropicMessagesResponse `json:"message"`
}

type anthropicContentBlockEvent struct {
	Type         string                  `json:"type"`
	Index        int                     `json:"index"`
	ContentBlock anthropicMessageContent `json:"content_block,omitempty"`
	Delta        anthropicContentDelta   `json:"delta,omitempty"`
}

type anthropicContentDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicMessageDeltaEvent struct {
	Type  string                 `json:"type"`
	Delta anthropicStopDelta     `json:"delta"`
	Usage anthropicMessagesUsage `json:"usage"`
}

type anthropicStopDelta struct {
	StopReason   string `json:"stop_reason"`
	StopSequence any    `json:"stop_sequence"`
}

func (s *Server) handleMessages(writer http.ResponseWriter, request *http.Request) {
	var messageRequest anthropicMessagesRequest
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&messageRequest); err != nil {
		writeOpenAIError(writer, http.StatusBadRequest, "INVALID_JSON", fmt.Sprintf("invalid messages request: %v", err))
		return
	}
	if len(messageRequest.Messages) == 0 {
		writeOpenAIError(writer, http.StatusBadRequest, "MISSING_MESSAGES", "messages must contain at least one item")
		return
	}

	ctx, cancel := context.WithTimeout(request.Context(), s.config.BridgeTimeout)
	defer cancel()

	selection := s.resolveNorthboundSelection(ctx, messageRequest.Model, messageRequest.RadishMind)
	promptText, northboundFields, err := buildMessagesPromptText(messageRequest)
	if err != nil {
		writeOpenAIError(writer, http.StatusBadRequest, "INVALID_MESSAGES_REQUEST", err.Error())
		return
	}

	canonicalRequest, err := buildNorthboundCanonicalRequest(northboundCanonicalRequestOptions{
		route:            "/v1/messages",
		protocol:         northboundProtocolMessages,
		locale:           resolveNorthboundLocale(messageRequest.RadishMind),
		promptText:       promptText,
		northboundFields: northboundFields,
	})
	if err != nil {
		writeOpenAIError(writer, http.StatusBadRequest, "INVALID_MESSAGES_REQUEST", err.Error())
		return
	}

	envelope, err := s.bridge.HandleEnvelope(
		ctx,
		canonicalRequest,
		bridge.EnvelopeOptions{
			Provider:        selection.provider,
			ProviderProfile: selection.providerProfile,
			Model:           selection.model,
			BaseURL:         s.config.BaseURL,
			APIKey:          s.config.APIKey,
			Temperature:     effectiveTemperature(messageRequest.Temperature, s.config.Temperature),
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

	if messageRequest.Stream {
		if err := streamAnthropicMessagesResponse(writer, envelope, selection.model); err != nil {
			return
		}
		return
	}

	responseDocument, err := buildAnthropicMessagesResponse(envelope, selection.model)
	if err != nil {
		writeOpenAIError(writer, http.StatusBadGateway, "PLATFORM_RESPONSE_INVALID", err.Error())
		return
	}
	writeJSON(writer, http.StatusOK, responseDocument)
}

func buildMessagesPromptText(request anthropicMessagesRequest) (string, map[string]any, error) {
	sections := make([]string, 0, 4)

	if systemText := strings.TrimSpace(buildNorthboundTextFromValue(request.System)); systemText != "" {
		sections = append(sections, "System:\n"+systemText)
	}
	if transcript, err := buildConversationTranscript(request.Messages); err != nil {
		return "", nil, err
	} else if transcript != "" {
		sections = append(sections, "Messages:\n"+transcript)
	}

	promptText := buildNorthboundPromptText(sections...)
	if promptText == "" {
		return "", nil, fmt.Errorf("messages request must include system text or messages")
	}

	northboundFields := map[string]any{
		"request_kind":    "anthropic_messages",
		"requested_model": request.Model,
		"message_count":   len(request.Messages),
		"max_tokens":      request.MaxTokens,
		"stream":          request.Stream,
	}
	if request.RadishMind != nil {
		northboundFields["provider"] = strings.TrimSpace(request.RadishMind.Provider)
		northboundFields["provider_profile"] = strings.TrimSpace(request.RadishMind.ProviderProfile)
		northboundFields["locale"] = strings.TrimSpace(request.RadishMind.Locale)
	}
	if len(request.Metadata) > 0 {
		northboundFields["metadata"] = request.Metadata
	}
	return promptText, northboundFields, nil
}

func buildAnthropicMessagesResponse(envelope bridge.GatewayEnvelope, model string) (anthropicMessagesResponse, error) {
	if envelope.Response == nil {
		return anthropicMessagesResponse{}, fmt.Errorf("gateway envelope is missing response payload")
	}

	content := buildNorthboundResponseContent(envelope)
	responseID := buildNorthboundResponseID("msg-", envelope.RequestID)

	return anthropicMessagesResponse{
		ID:           responseID,
		Type:         "message",
		Role:         "assistant",
		Model:        model,
		StopReason:   "end_turn",
		StopSequence: nil,
		Content: []anthropicMessageContent{
			{
				Type: "text",
				Text: content,
			},
		},
		Usage: anthropicMessagesUsage{
			InputTokens:  0,
			OutputTokens: 0,
		},
	}, nil
}

func streamAnthropicMessagesResponse(
	writer http.ResponseWriter,
	envelope bridge.GatewayEnvelope,
	model string,
) error {
	responseDocument, err := buildAnthropicMessagesResponse(envelope, model)
	if err != nil {
		return err
	}
	contentText := ""
	if len(responseDocument.Content) > 0 {
		contentText = responseDocument.Content[0].Text
	}

	prepareSSEHeaders(writer)
	if err := writeSSEEvent(writer, "message_start", anthropicMessageStartEvent{
		Type:    "message_start",
		Message: responseDocument,
	}); err != nil {
		return err
	}
	if err := writeSSEEvent(writer, "content_block_start", anthropicContentBlockEvent{
		Type:  "content_block_start",
		Index: 0,
		ContentBlock: anthropicMessageContent{
			Type: "text",
			Text: "",
		},
	}); err != nil {
		return err
	}
	for _, chunkText := range splitTextForStreaming(contentText, 96) {
		if chunkText == "" {
			continue
		}
		if err := writeSSEEvent(writer, "content_block_delta", anthropicContentBlockEvent{
			Type:  "content_block_delta",
			Index: 0,
			Delta: anthropicContentDelta{
				Type: "text_delta",
				Text: chunkText,
			},
		}); err != nil {
			return err
		}
	}
	if err := writeSSEEvent(writer, "content_block_stop", anthropicContentBlockEvent{
		Type:  "content_block_stop",
		Index: 0,
	}); err != nil {
		return err
	}
	if err := writeSSEEvent(writer, "message_delta", anthropicMessageDeltaEvent{
		Type:  "message_delta",
		Delta: anthropicStopDelta{StopReason: responseDocument.StopReason, StopSequence: responseDocument.StopSequence},
		Usage: responseDocument.Usage,
	}); err != nil {
		return err
	}
	if err := writeSSEEvent(writer, "message_stop", map[string]string{
		"type": "message_stop",
	}); err != nil {
		return err
	}
	return writeSSEEvent(writer, "", "[DONE]")
}
