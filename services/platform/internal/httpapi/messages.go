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
	trace := newRequestTrace(request, "/v1/messages")
	var messageRequest anthropicMessagesRequest
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&messageRequest); err != nil {
		s.writePlatformError(writer, trace, "INVALID_JSON", fmt.Sprintf("invalid messages request: %v", err))
		return
	}
	if len(messageRequest.Messages) == 0 {
		s.writePlatformError(writer, trace, "MISSING_MESSAGES", "messages must contain at least one item")
		return
	}

	ctx, cancel := context.WithTimeout(request.Context(), s.config.BridgeTimeout)
	defer cancel()

	selection := s.resolveNorthboundSelection(ctx, messageRequest.Model, messageRequest.RadishMind)
	trace.applySelection(selection)
	promptText, northboundFields, err := buildMessagesPromptText(messageRequest, selection)
	if err != nil {
		s.writePlatformError(writer, trace, "INVALID_MESSAGES_REQUEST", err.Error())
		return
	}

	canonicalRequest, err := buildNorthboundCanonicalRequest(northboundCanonicalRequestOptions{
		requestID:        trace.requestID,
		route:            "/v1/messages",
		protocol:         northboundProtocolMessages,
		locale:           resolveNorthboundLocale(messageRequest.RadishMind),
		promptText:       promptText,
		northboundFields: northboundFields,
	})
	if err != nil {
		s.writePlatformError(writer, trace, "INVALID_MESSAGES_REQUEST", err.Error())
		return
	}

	if messageRequest.Stream {
		if err := s.streamAnthropicMessagesResponse(ctx, writer, canonicalRequest, selection, effectiveTemperature(messageRequest.Temperature, s.config.Temperature), trace); err != nil {
			s.writePlatformError(writer, trace, "PLATFORM_BRIDGE_FAILED", err.Error())
			return
		}
		return
	}

	envelope, err := s.bridge.HandleEnvelope(
		ctx,
		canonicalRequest,
		s.buildBridgeEnvelopeOptions(selection, effectiveTemperature(messageRequest.Temperature, s.config.Temperature)),
	)
	if err != nil {
		s.writePlatformError(writer, trace, "PLATFORM_BRIDGE_FAILED", err.Error())
		return
	}
	if strings.EqualFold(envelope.Status, "failed") {
		s.writePlatformError(writer, trace, gatewayErrorCode(envelope.Error), gatewayErrorMessage(envelope.Error))
		return
	}

	responseDocument, err := buildAnthropicMessagesResponse(envelope, selection.model)
	if err != nil {
		s.writePlatformError(writer, trace, "PLATFORM_RESPONSE_INVALID", err.Error())
		return
	}
	writeObservedJSON(writer, http.StatusOK, trace, responseDocument)
}

func buildMessagesPromptText(request anthropicMessagesRequest, selection northboundSelection) (string, map[string]any, error) {
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
		"request_kind":  "anthropic_messages",
		"message_count": len(request.Messages),
		"max_tokens":    request.MaxTokens,
		"stream":        request.Stream,
	}
	for key, value := range buildNorthboundSelectionFields(request.Model, selection, request.RadishMind) {
		northboundFields[key] = value
	}
	if len(request.Metadata) > 0 {
		northboundFields["metadata"] = request.Metadata
	}
	return promptText, northboundFields, nil
}

func buildAnthropicMessagesResponse(envelope bridge.GatewayEnvelope, model string) (anthropicMessagesResponse, error) {
	return buildAnthropicMessagesResponseWithID(envelope, model, buildNorthboundResponseID("msg-", envelope.RequestID))
}

func buildAnthropicMessagesResponseWithID(envelope bridge.GatewayEnvelope, model string, responseID string) (anthropicMessagesResponse, error) {
	if envelope.Response == nil {
		return anthropicMessagesResponse{}, fmt.Errorf("gateway envelope is missing response payload")
	}

	content := buildNorthboundResponseContent(envelope)

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

func buildAnthropicMessagesResponseSkeleton(responseID string, model string) anthropicMessagesResponse {
	return anthropicMessagesResponse{
		ID:         responseID,
		Type:       "message",
		Role:       "assistant",
		Model:      model,
		StopReason: "",
		Content:    nil,
		Usage: anthropicMessagesUsage{
			InputTokens:  0,
			OutputTokens: 0,
		},
	}
}

func (s *Server) streamAnthropicMessagesResponse(
	ctx context.Context,
	writer http.ResponseWriter,
	canonicalRequest []byte,
	selection northboundSelection,
	temperature float64,
	trace requestTrace,
) error {
	responseID := buildNorthboundResponseID("msg-", trace.requestID)
	streamStarted := false
	streamCompleted := false
	emittedContent := false
	var completedEnvelope *bridge.GatewayEnvelope

	startStream := func() error {
		if streamStarted {
			return nil
		}
		prepareSSEHeaders(writer)
		writeTraceHeaders(writer, trace)
		if err := writeSSEEvent(writer, "message_start", anthropicMessageStartEvent{
			Type:    "message_start",
			Message: buildAnthropicMessagesResponseSkeleton(responseID, selection.model),
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
				return writeSSEEvent(writer, "content_block_delta", anthropicContentBlockEvent{
					Type:  "content_block_delta",
					Index: 0,
					Delta: anthropicContentDelta{
						Type: "text_delta",
						Text: event.Delta,
					},
				})
			case "completed":
				streamCompleted = true
				if event.Envelope != nil {
					completedEnvelope = event.Envelope
				}
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
	if completedEnvelope == nil {
		return fmt.Errorf("platform bridge stream completed without envelope")
	}

	responseDocument, err := buildAnthropicMessagesResponseWithID(*completedEnvelope, selection.model, responseID)
	if err != nil {
		return err
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
	if err := writeSSEEvent(writer, "", "[DONE]"); err != nil {
		return err
	}
	logRequestTrace(trace, http.StatusOK, "", "")
	return nil
}
