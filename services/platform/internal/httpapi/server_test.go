package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
)

type fakeBridge struct {
	providers   []bridge.ProviderDescription
	envelope    bridge.GatewayEnvelope
	lastRequest []byte
	lastOptions bridge.EnvelopeOptions
}

func (f *fakeBridge) DescribeProviders(context.Context) ([]bridge.ProviderDescription, error) {
	return f.providers, nil
}

func (f *fakeBridge) HandleEnvelope(_ context.Context, canonicalRequest []byte, options bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
	f.lastRequest = append(f.lastRequest[:0], canonicalRequest...)
	f.lastOptions = options
	return f.envelope, nil
}

func TestPlatformNorthboundRoutes(t *testing.T) {
	fb := &fakeBridge{
		providers: []bridge.ProviderDescription{
			{
				ProviderID:         "mock",
				DisplayName:        "Mock provider",
				DefaultAPIStyle:    "mock",
				SupportedAPIStyles: []string{"mock"},
				ProfileDriven:      false,
				Notes:              "test provider",
				Capabilities:       map[string]any{"chat": false},
			},
		},
		envelope: bridge.GatewayEnvelope{
			SchemaVersion: 1,
			Status:        "ok",
			RequestID:     "req-123",
			Project:       "radish",
			Task:          "answer_docs_question",
			Response: map[string]any{
				"summary": "bridge summary",
			},
			Metadata: map[string]any{},
		},
	}

	server := &Server{
		bridge: fb,
		config: config.Config{
			BridgeTimeout:   time.Second,
			Provider:        "mock",
			ProviderProfile: "default",
			Model:           "platform-model",
		},
		options: Options{BuildVersion: "test"},
	}

	t.Run("responses", func(t *testing.T) {
		body := `{"model":"platform-model","input":"Please answer this","metadata":{"source":"test"}}`
		req := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(body))
		rec := httptest.NewRecorder()

		server.handleResponses(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
		}
		var response openAIResponsesResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if response.Object != "response" {
			t.Fatalf("unexpected object: %s", response.Object)
		}
		if response.OutputText != "bridge summary" {
			t.Fatalf("unexpected output text: %s", response.OutputText)
		}
		if response.Model != "platform-model" {
			t.Fatalf("unexpected model: %s", response.Model)
		}

		var canonicalRequest map[string]any
		if err := json.Unmarshal(fb.lastRequest, &canonicalRequest); err != nil {
			t.Fatalf("decode canonical request: %v", err)
		}
		contextBlock, ok := canonicalRequest["context"].(map[string]any)
		if !ok {
			t.Fatalf("missing context block: %#v", canonicalRequest["context"])
		}
		northbound, ok := contextBlock["northbound"].(map[string]any)
		if !ok {
			t.Fatalf("missing northbound context: %#v", contextBlock["northbound"])
		}
		if northbound["protocol"] != northboundProtocolResponses {
			t.Fatalf("unexpected protocol: %#v", northbound["protocol"])
		}
		if northbound["request_kind"] != "responses" {
			t.Fatalf("unexpected request kind: %#v", northbound["request_kind"])
		}
		if fb.lastOptions.Model != "platform-model" {
			t.Fatalf("unexpected bridge model: %s", fb.lastOptions.Model)
		}
	})

	t.Run("responses stream", func(t *testing.T) {
		body := `{"model":"platform-model","input":"Please answer this","stream":true}`
		req := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(body))
		rec := httptest.NewRecorder()

		server.handleResponses(rec, req)

		if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/event-stream") {
			t.Fatalf("unexpected content type: %s", got)
		}
		streamBody := rec.Body.String()
		if !strings.Contains(streamBody, "event: response.output_text.delta") {
			t.Fatalf("missing response delta event: %s", streamBody)
		}
		if !strings.Contains(streamBody, "event: response.completed") || !strings.Contains(streamBody, "data: [DONE]") {
			t.Fatalf("missing response completion markers: %s", streamBody)
		}
	})

	t.Run("messages", func(t *testing.T) {
		body := `{"model":"platform-model","system":"You are helpful","messages":[{"role":"user","content":"你好"}]}`
		req := httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(body))
		rec := httptest.NewRecorder()

		server.handleMessages(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
		}
		var response anthropicMessagesResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if response.Type != "message" {
			t.Fatalf("unexpected type: %s", response.Type)
		}
		if len(response.Content) != 1 || response.Content[0].Text != "bridge summary" {
			t.Fatalf("unexpected content: %#v", response.Content)
		}
		if response.Model != "platform-model" {
			t.Fatalf("unexpected model: %s", response.Model)
		}

		var canonicalRequest map[string]any
		if err := json.Unmarshal(fb.lastRequest, &canonicalRequest); err != nil {
			t.Fatalf("decode canonical request: %v", err)
		}
		contextBlock, ok := canonicalRequest["context"].(map[string]any)
		if !ok {
			t.Fatalf("missing context block: %#v", canonicalRequest["context"])
		}
		northbound, ok := contextBlock["northbound"].(map[string]any)
		if !ok {
			t.Fatalf("missing northbound context: %#v", contextBlock["northbound"])
		}
		if northbound["protocol"] != northboundProtocolMessages {
			t.Fatalf("unexpected protocol: %#v", northbound["protocol"])
		}
		if northbound["request_kind"] != "anthropic_messages" {
			t.Fatalf("unexpected request kind: %#v", northbound["request_kind"])
		}
	})

	t.Run("messages stream", func(t *testing.T) {
		body := `{"model":"platform-model","system":"You are helpful","messages":[{"role":"user","content":"你好"}],"stream":true}`
		req := httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(body))
		rec := httptest.NewRecorder()

		server.handleMessages(rec, req)

		if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/event-stream") {
			t.Fatalf("unexpected content type: %s", got)
		}
		streamBody := rec.Body.String()
		if !strings.Contains(streamBody, "event: message_start") || !strings.Contains(streamBody, "event: message_stop") {
			t.Fatalf("missing anthropic lifecycle events: %s", streamBody)
		}
		if !strings.Contains(streamBody, "event: content_block_delta") || !strings.Contains(streamBody, "data: [DONE]") {
			t.Fatalf("missing anthropic delta markers: %s", streamBody)
		}
	})

	t.Run("chat stream", func(t *testing.T) {
		body := `{"model":"platform-model","messages":[{"role":"user","content":"hello"}],"stream":true}`
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
		rec := httptest.NewRecorder()

		server.handleChatCompletions(rec, req)

		if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/event-stream") {
			t.Fatalf("unexpected content type: %s", got)
		}
		streamBody := rec.Body.String()
		if !strings.Contains(streamBody, "chat.completion.chunk") || !strings.Contains(streamBody, "data: [DONE]") {
			t.Fatalf("missing chat completion stream markers: %s", streamBody)
		}
	})

	t.Run("models", func(t *testing.T) {
		rec := httptest.NewRecorder()

		handleModels(rec, server)

		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
		}
		var response openAIModelListResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if response.Object != "list" {
			t.Fatalf("unexpected object: %s", response.Object)
		}
		if len(response.Data) == 0 {
			t.Fatalf("expected model inventory")
		}
		if response.Data[0].ID != "platform-model" {
			t.Fatalf("unexpected first model id: %s", response.Data[0].ID)
		}
		if response.Data[0].Metadata["source"] != "configured_default" {
			t.Fatalf("unexpected metadata source: %#v", response.Data[0].Metadata["source"])
		}
	})
}
