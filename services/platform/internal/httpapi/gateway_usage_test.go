package httpapi

import (
	"testing"

	"radishmind.local/services/platform/internal/bridge"
)

func TestGatewayUsageFromEnvelopeAcceptsOnlyValidReportedUsage(t *testing.T) {
	validEnvelope := bridge.GatewayEnvelope{Metadata: map[string]any{
		"usage": map[string]any{
			"availability":  "reported",
			"source":        "gemini_usage_metadata",
			"input_tokens":  float64(21),
			"output_tokens": float64(8),
			"total_tokens":  float64(29),
		},
	}}
	usage := gatewayUsageFromEnvelope(validEnvelope)
	if usage.Availability != GatewayRequestUsageReported ||
		usage.Source != "gemini_usage_metadata" ||
		usage.InputTokens != 21 || usage.OutputTokens != 8 || usage.TotalTokens != 29 {
		t.Fatalf("unexpected normalized usage: %#v", usage)
	}

	invalidEnvelopes := []bridge.GatewayEnvelope{
		{Metadata: map[string]any{}},
		{Metadata: map[string]any{"usage": map[string]any{
			"availability": "reported", "source": "unknown_usage",
			"input_tokens": 2, "output_tokens": 3, "total_tokens": 5,
		}}},
		{Metadata: map[string]any{"usage": map[string]any{
			"availability": "reported", "source": "openai_compatible_usage",
			"input_tokens": 2, "output_tokens": 3, "total_tokens": 9,
		}}},
		{Metadata: map[string]any{"usage": map[string]any{
			"availability": "reported", "source": "openai_compatible_usage",
			"input_tokens": true, "output_tokens": 3, "total_tokens": 3,
		}}},
	}
	for index, envelope := range invalidEnvelopes {
		usage = gatewayUsageFromEnvelope(envelope)
		if usage != (GatewayRequestUsage{Availability: GatewayRequestUsageNotReported}) {
			t.Fatalf("case %d must remain not_reported, got %#v", index, usage)
		}
	}
}

func TestReportedGatewayUsageProjectsToNorthboundProtocolsAndRecorder(t *testing.T) {
	envelope := bridge.GatewayEnvelope{
		RequestID: "request_usage_projection",
		Response:  map[string]any{"summary": "ok"},
		Metadata: map[string]any{
			"usage": map[string]any{
				"availability":  "reported",
				"source":        "anthropic_usage",
				"input_tokens":  13,
				"output_tokens": 5,
				"total_tokens":  18,
			},
		},
	}

	chatResponse, err := buildOpenAIChatCompletionResponse(envelope, "model")
	if err != nil || chatResponse.Usage == nil ||
		chatResponse.Usage.PromptTokens != 13 ||
		chatResponse.Usage.CompletionTokens != 5 ||
		chatResponse.Usage.TotalTokens != 18 {
		t.Fatalf("unexpected chat usage projection: response=%#v err=%v", chatResponse, err)
	}
	responsesResponse, err := buildOpenAIResponsesResponse(envelope, "model")
	if err != nil || responsesResponse.Usage == nil ||
		responsesResponse.Usage.InputTokens != 13 ||
		responsesResponse.Usage.OutputTokens != 5 ||
		responsesResponse.Usage.TotalTokens != 18 {
		t.Fatalf("unexpected responses usage projection: response=%#v err=%v", responsesResponse, err)
	}
	messagesResponse, err := buildAnthropicMessagesResponse(envelope, "model")
	if err != nil || messagesResponse.Usage == nil ||
		messagesResponse.Usage.InputTokens != 13 ||
		messagesResponse.Usage.OutputTokens != 5 {
		t.Fatalf("unexpected messages usage projection: response=%#v err=%v", messagesResponse, err)
	}

	server := &Server{}
	record := GatewayRequestRecord{Usage: GatewayRequestUsage{Availability: GatewayRequestUsageNotReported}}
	trace := requestTrace{gatewayRequest: &record}
	server.applyGatewayEnvelopeToTrace(&trace, envelope)
	if record.Usage.Availability != GatewayRequestUsageReported ||
		record.Usage.Source != "anthropic_usage" ||
		record.Usage.TotalTokens != 18 {
		t.Fatalf("recorder did not retain reported usage: %#v", record.Usage)
	}
}

func TestNorthboundProtocolsOmitUnreportedUsage(t *testing.T) {
	envelope := bridge.GatewayEnvelope{
		RequestID: "request_usage_omitted",
		Response:  map[string]any{"summary": "ok"},
		Metadata:  map[string]any{},
	}
	chatResponse, _ := buildOpenAIChatCompletionResponse(envelope, "model")
	responsesResponse, _ := buildOpenAIResponsesResponse(envelope, "model")
	messagesResponse, _ := buildAnthropicMessagesResponse(envelope, "model")
	if chatResponse.Usage != nil || responsesResponse.Usage != nil || messagesResponse.Usage != nil {
		t.Fatalf(
			"unreported usage must be omitted: chat=%#v responses=%#v messages=%#v",
			chatResponse.Usage,
			responsesResponse.Usage,
			messagesResponse.Usage,
		)
	}
}
