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
	var responseRequest openAIResponsesRequest
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&responseRequest); err != nil {
		writeOpenAIError(writer, http.StatusBadRequest, "INVALID_JSON", fmt.Sprintf("invalid responses request: %v", err))
		return
	}

	ctx, cancel := context.WithTimeout(request.Context(), s.config.BridgeTimeout)
	defer cancel()

	selection := s.resolveNorthboundSelection(ctx, responseRequest.Model, responseRequest.RadishMind)
	promptText, northboundFields, err := buildResponsesPromptText(responseRequest)
	if err != nil {
		writeOpenAIError(writer, http.StatusBadRequest, "INVALID_RESPONSES_REQUEST", err.Error())
		return
	}

	canonicalRequest, err := buildNorthboundCanonicalRequest(northboundCanonicalRequestOptions{
		route:            "/v1/responses",
		protocol:         northboundProtocolResponses,
		locale:           resolveNorthboundLocale(responseRequest.RadishMind),
		promptText:       promptText,
		northboundFields: northboundFields,
	})
	if err != nil {
		writeOpenAIError(writer, http.StatusBadRequest, "INVALID_RESPONSES_REQUEST", err.Error())
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
			Temperature:     effectiveTemperature(responseRequest.Temperature, s.config.Temperature),
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

	if responseRequest.Stream {
		if err := streamOpenAIResponsesResponse(writer, envelope, selection.model); err != nil {
			return
		}
		return
	}

	responseDocument, err := buildOpenAIResponsesResponse(envelope, selection.model)
	if err != nil {
		writeOpenAIError(writer, http.StatusBadGateway, "PLATFORM_RESPONSE_INVALID", err.Error())
		return
	}
	writeJSON(writer, http.StatusOK, responseDocument)
}

func buildResponsesPromptText(request openAIResponsesRequest) (string, map[string]any, error) {
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
		"request_kind":    "responses",
		"requested_model": request.Model,
		"input_kind":      describeNorthboundInputKind(request.Input),
		"message_count":   len(request.Messages),
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

func buildOpenAIResponsesResponse(envelope bridge.GatewayEnvelope, model string) (openAIResponsesResponse, error) {
	if envelope.Response == nil {
		return openAIResponsesResponse{}, fmt.Errorf("gateway envelope is missing response payload")
	}

	content := buildNorthboundResponseContent(envelope)
	responseID := buildNorthboundResponseID("resp-", envelope.RequestID)
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

func streamOpenAIResponsesResponse(
	writer http.ResponseWriter,
	envelope bridge.GatewayEnvelope,
	model string,
) error {
	responseDocument, err := buildOpenAIResponsesResponse(envelope, model)
	if err != nil {
		return err
	}

	prepareSSEHeaders(writer)
	if err := writeSSEEvent(writer, "response.created", openAIResponsesStreamEvent{
		Type:       "response.created",
		ResponseID: responseDocument.ID,
		Response: &openAIResponsesResponse{
			ID:        responseDocument.ID,
			Object:    responseDocument.Object,
			CreatedAt: responseDocument.CreatedAt,
			Model:     responseDocument.Model,
			Status:    "in_progress",
		},
	}); err != nil {
		return err
	}

	for _, chunkText := range splitTextForStreaming(responseDocument.OutputText, 96) {
		if chunkText == "" {
			continue
		}
		if err := writeSSEEvent(writer, "response.output_text.delta", openAIResponsesStreamEvent{
			Type:         "response.output_text.delta",
			ResponseID:   responseDocument.ID,
			OutputIndex:  0,
			ContentIndex: 0,
			Delta:        chunkText,
		}); err != nil {
			return err
		}
	}

	if err := writeSSEEvent(writer, "response.output_text.done", openAIResponsesStreamEvent{
		Type:         "response.output_text.done",
		ResponseID:   responseDocument.ID,
		OutputIndex:  0,
		ContentIndex: 0,
	}); err != nil {
		return err
	}
	if err := writeSSEEvent(writer, "response.completed", openAIResponsesStreamEvent{
		Type:       "response.completed",
		ResponseID: responseDocument.ID,
		Response:   &responseDocument,
	}); err != nil {
		return err
	}
	return writeSSEEvent(writer, "", "[DONE]")
}
