package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"radishmind.local/services/platform/internal/bridge"
)

type openAIResponsesRequest struct {
	Model        string                   `json:"model"`
	Input        any                      `json:"input,omitempty"`
	Instructions any                      `json:"instructions,omitempty"`
	Messages     []chatCompletionMessage  `json:"messages,omitempty"`
	Stream       bool                     `json:"stream,omitempty"`
	Temperature  *float64                 `json:"temperature,omitempty"`
	RadishMind   *chatCompletionExtension `json:"radishmind,omitempty"`
	Metadata     map[string]any           `json:"metadata,omitempty"`
}

type openAIResponsesResponse struct {
	ID         string                      `json:"id"`
	Object     string                      `json:"object"`
	CreatedAt  int64                       `json:"created_at"`
	Model      string                      `json:"model"`
	Status     string                      `json:"status"`
	Output     []openAIResponsesOutputItem `json:"output"`
	OutputText string                      `json:"output_text"`
	Usage      openAIResponsesUsage        `json:"usage"`
	Metadata   map[string]any              `json:"metadata,omitempty"`
}

type openAIResponsesOutputItem struct {
	ID      string                       `json:"id"`
	Type    string                       `json:"type"`
	Role    string                       `json:"role"`
	Status  string                       `json:"status,omitempty"`
	Content []openAIResponsesContentPart `json:"content"`
}

type openAIResponsesContentPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type openAIResponsesUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type openAIResponsesStreamEvent struct {
	Type         string                   `json:"type"`
	ResponseID   string                   `json:"response_id,omitempty"`
	OutputIndex  int                      `json:"output_index,omitempty"`
	ContentIndex int                      `json:"content_index,omitempty"`
	Delta        string                   `json:"delta,omitempty"`
	Response     *openAIResponsesResponse `json:"response,omitempty"`
}

func (s *Server) handleResponses(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, "/v1/responses")
	var responseRequest openAIResponsesRequest
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&responseRequest); err != nil {
		s.writePlatformError(writer, trace, "INVALID_JSON", fmt.Sprintf("invalid responses request: %v", err))
		return
	}

	ctx, cancel := context.WithTimeout(request.Context(), s.config.BridgeTimeout)
	defer cancel()

	selection := s.resolveNorthboundSelection(ctx, responseRequest.Model, responseRequest.RadishMind)
	trace.applySelection(selection)
	promptText, northboundFields, err := buildResponsesPromptText(responseRequest, selection)
	if err != nil {
		s.writePlatformError(writer, trace, "INVALID_RESPONSES_REQUEST", err.Error())
		return
	}

	canonicalRequest, err := buildNorthboundCanonicalRequest(northboundCanonicalRequestOptions{
		requestID:        trace.requestID,
		route:            "/v1/responses",
		protocol:         northboundProtocolResponses,
		locale:           resolveNorthboundLocale(responseRequest.RadishMind),
		promptText:       promptText,
		northboundFields: northboundFields,
	})
	if err != nil {
		s.writePlatformError(writer, trace, "INVALID_RESPONSES_REQUEST", err.Error())
		return
	}

	if responseRequest.Stream {
		if err := s.streamOpenAIResponsesResponse(ctx, writer, canonicalRequest, selection, effectiveTemperature(responseRequest.Temperature, s.config.Temperature), trace); err != nil {
			s.writePlatformError(writer, trace, "PLATFORM_BRIDGE_FAILED", err.Error())
			return
		}
		return
	}

	envelope, err := s.bridge.HandleEnvelope(
		ctx,
		canonicalRequest,
		s.buildBridgeEnvelopeOptions(selection, effectiveTemperature(responseRequest.Temperature, s.config.Temperature)),
	)
	if err != nil {
		s.writePlatformError(writer, trace, "PLATFORM_BRIDGE_FAILED", err.Error())
		return
	}
	if strings.EqualFold(envelope.Status, "failed") {
		s.writePlatformError(writer, trace, gatewayErrorCode(envelope.Error), gatewayErrorMessage(envelope.Error))
		return
	}

	responseDocument, err := buildOpenAIResponsesResponse(envelope, selection.model)
	if err != nil {
		s.writePlatformError(writer, trace, "PLATFORM_RESPONSE_INVALID", err.Error())
		return
	}
	writeObservedJSON(writer, http.StatusOK, trace, responseDocument)
}

func buildResponsesPromptText(request openAIResponsesRequest, selection northboundSelection) (string, map[string]any, error) {
	sections := make([]string, 0, 4)

	if instructions := strings.TrimSpace(buildNorthboundTextFromValue(request.Instructions)); instructions != "" {
		sections = append(sections, "Instructions:\n"+instructions)
	}
	if inputText := strings.TrimSpace(buildNorthboundTextFromValue(request.Input)); inputText != "" {
		sections = append(sections, "Input:\n"+inputText)
	}
	if len(request.Messages) > 0 {
		transcript, err := buildConversationTranscript(request.Messages)
		if err != nil {
			return "", nil, err
		}
		if transcript != "" {
			sections = append(sections, "Messages:\n"+transcript)
		}
	}

	promptText := buildNorthboundPromptText(sections...)
	if promptText == "" {
		return "", nil, fmt.Errorf("responses request must include input, instructions, or messages")
	}

	northboundFields := map[string]any{
		"request_kind":  "responses",
		"input_kind":    describeNorthboundInputKind(request.Input),
		"message_count": len(request.Messages),
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

func buildOpenAIResponsesResponse(envelope bridge.GatewayEnvelope, model string) (openAIResponsesResponse, error) {
	return buildOpenAIResponsesResponseWithID(envelope, model, buildNorthboundResponseID("resp-", envelope.RequestID))
}

func buildOpenAIResponsesResponseWithID(envelope bridge.GatewayEnvelope, model string, responseID string) (openAIResponsesResponse, error) {
	if envelope.Response == nil {
		return openAIResponsesResponse{}, fmt.Errorf("gateway envelope is missing response payload")
	}

	content := buildNorthboundResponseContent(envelope)
	now := timeNowUnix()

	return openAIResponsesResponse{
		ID:        responseID,
		Object:    "response",
		CreatedAt: now,
		Model:     model,
		Status:    "completed",
		Output: []openAIResponsesOutputItem{
			{
				ID:     responseID + "-message",
				Type:   "message",
				Role:   "assistant",
				Status: "completed",
				Content: []openAIResponsesContentPart{
					{
						Type: "output_text",
						Text: content,
					},
				},
			},
		},
		OutputText: content,
		Usage: openAIResponsesUsage{
			InputTokens:  0,
			OutputTokens: 0,
			TotalTokens:  0,
		},
		Metadata: map[string]any{
			"route": "/v1/responses",
		},
	}, nil
}

func describeNorthboundInputKind(value any) string {
	switch value.(type) {
	case nil:
		return ""
	case string:
		return "text"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	default:
		return "structured"
	}
}

func (s *Server) streamOpenAIResponsesResponse(
	ctx context.Context,
	writer http.ResponseWriter,
	canonicalRequest []byte,
	selection northboundSelection,
	temperature float64,
	trace requestTrace,
) error {
	responseID := buildNorthboundResponseID("resp-", trace.requestID)
	createdAt := timeNowUnix()
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
		if err := writeSSEEvent(writer, "response.created", openAIResponsesStreamEvent{
			Type:       "response.created",
			ResponseID: responseID,
			Response: &openAIResponsesResponse{
				ID:        responseID,
				Object:    "response",
				CreatedAt: createdAt,
				Model:     selection.model,
				Status:    "in_progress",
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
				return writeSSEEvent(writer, "response.output_text.delta", openAIResponsesStreamEvent{
					Type:         "response.output_text.delta",
					ResponseID:   responseID,
					OutputIndex:  0,
					ContentIndex: 0,
					Delta:        event.Delta,
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

	responseDocument, err := buildOpenAIResponsesResponseWithID(*completedEnvelope, selection.model, responseID)
	if err != nil {
		return err
	}
	if err := writeSSEEvent(writer, "response.output_text.done", openAIResponsesStreamEvent{
		Type:         "response.output_text.done",
		ResponseID:   responseID,
		OutputIndex:  0,
		ContentIndex: 0,
	}); err != nil {
		return err
	}
	if err := writeSSEEvent(writer, "response.completed", openAIResponsesStreamEvent{
		Type:       "response.completed",
		ResponseID: responseID,
		Response:   &responseDocument,
	}); err != nil {
		return err
	}
	if err := writeSSEEvent(writer, "", "[DONE]"); err != nil {
		return err
	}
	logRequestTrace(trace, http.StatusOK, "", "")
	return nil
}
