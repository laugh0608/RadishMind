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
	providers    []bridge.ProviderDescription
	inventory    bridge.ProviderInventory
	envelope     bridge.GatewayEnvelope
	lastRequest  []byte
	lastOptions  bridge.EnvelopeOptions
	streamCalled bool
	streamErr    error
}

func (f *fakeBridge) DescribeProviders(context.Context) ([]bridge.ProviderDescription, error) {
	return f.providers, nil
}

func (f *fakeBridge) DescribeInventory(context.Context) (bridge.ProviderInventory, error) {
	return f.inventory, nil
}

func (f *fakeBridge) HandleEnvelope(_ context.Context, canonicalRequest []byte, options bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
	f.lastRequest = append(f.lastRequest[:0], canonicalRequest...)
	f.lastOptions = options
	return f.envelope, nil
}

func (f *fakeBridge) StreamEnvelope(_ context.Context, canonicalRequest []byte, options bridge.EnvelopeOptions, handleEvent func(bridge.StreamEvent) error) error {
	f.lastRequest = append(f.lastRequest[:0], canonicalRequest...)
	f.lastOptions = options
	f.streamCalled = true
	if f.streamErr != nil {
		return f.streamErr
	}
	if handleEvent != nil {
		summary := ""
		if f.envelope.Response != nil {
			if rawSummary, ok := f.envelope.Response["summary"].(string); ok {
				summary = rawSummary
			}
		}
		if summary != "" {
			if err := handleEvent(bridge.StreamEvent{Type: "delta", Delta: summary}); err != nil {
				return err
			}
		}
		if err := handleEvent(bridge.StreamEvent{Type: "completed", Envelope: &f.envelope}); err != nil {
			return err
		}
	}
	return nil
}

func findNorthboundModel(data []openAIModelObject, modelID string) (openAIModelObject, bool) {
	for _, model := range data {
		if model.ID == modelID {
			return model, true
		}
	}
	return openAIModelObject{}, false
}

func decodeCanonicalRequest(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var canonicalRequest map[string]any
	if err := json.Unmarshal(raw, &canonicalRequest); err != nil {
		t.Fatalf("decode canonical request: %v", err)
	}
	return canonicalRequest
}

func canonicalNorthboundContext(t *testing.T, canonicalRequest map[string]any) map[string]any {
	t.Helper()
	contextBlock, ok := canonicalRequest["context"].(map[string]any)
	if !ok {
		t.Fatalf("missing context block: %#v", canonicalRequest["context"])
	}
	northbound, ok := contextBlock["northbound"].(map[string]any)
	if !ok {
		t.Fatalf("missing northbound context: %#v", contextBlock["northbound"])
	}
	return northbound
}

func TestPlatformNorthboundRoutes(t *testing.T) {
	providerDescriptions := []bridge.ProviderDescription{
		{
			ProviderID:         "mock",
			DisplayName:        "Mock provider",
			DefaultAPIStyle:    "mock",
			SupportedAPIStyles: []string{"mock"},
			ProfileDriven:      false,
			Notes:              "test provider",
			Capabilities:       map[string]any{"chat": false, "streaming": false},
		},
		{
			ProviderID:         "openai-compatible",
			DisplayName:        "OpenAI-compatible provider family",
			DefaultAPIStyle:    "openai-compatible",
			SupportedAPIStyles: []string{"openai-compatible", "gemini-native", "anthropic-messages"},
			ProfileDriven:      true,
			Notes:              "test provider",
			Capabilities:       map[string]any{"chat": true, "streaming": true, "auth_mode": "profile", "deployment_mode": "remote_api"},
		},
		{
			ProviderID:         "huggingface",
			DisplayName:        "Hugging Face chat-completions provider",
			DefaultAPIStyle:    "huggingface-chat-completions",
			SupportedAPIStyles: []string{"huggingface-chat-completions"},
			ProfileDriven:      true,
			Notes:              "test provider",
			Capabilities:       map[string]any{"chat": true, "streaming": true, "auth_mode": "profile", "deployment_mode": "remote_api"},
		},
		{
			ProviderID:         "ollama",
			DisplayName:        "Ollama chat-completions provider",
			DefaultAPIStyle:    "ollama-chat-completions",
			SupportedAPIStyles: []string{"ollama-chat-completions"},
			ProfileDriven:      true,
			Notes:              "test provider",
			Capabilities:       map[string]any{"chat": true, "streaming": true, "auth_mode": "optional", "deployment_mode": "local_daemon"},
		},
	}

	fb := &fakeBridge{
		providers: providerDescriptions,
		inventory: bridge.ProviderInventory{
			Providers: providerDescriptions,
			Profiles: []bridge.ProviderProfileDescription{
				{
					Profile:               "anyrouter",
					NormalizedProfile:     "ANYROUTER",
					ProviderID:            "openai-compatible",
					ResolvedModel:         "deepseek-chat",
					APIStyle:              "openai-compatible",
					HasBaseURL:            true,
					HasAPIKey:             true,
					RequestTimeoutSeconds: 120,
					Active:                true,
					Fallback:              false,
					ChainIndex:            0,
					Capabilities:          map[string]any{"chat": true, "streaming": true, "auth_mode": "profile", "deployment_mode": "remote_api"},
					NorthboundProtocols:   []string{"chat.completions"},
					NorthboundRoutes:      []string{"/v1/models", "/v1/chat/completions"},
					CredentialState:       "configured",
					DeploymentMode:        "remote_api",
					AuthMode:              "profile",
					Streaming:             true,
				},
				{
					Profile:               "hf-chat",
					NormalizedProfile:     "HF_CHAT",
					ProviderID:            "huggingface",
					ResolvedModel:         "meta-llama/Meta-Llama-3.1-8B-Instruct",
					APIStyle:              "huggingface-chat-completions",
					HasBaseURL:            true,
					HasAPIKey:             true,
					RequestTimeoutSeconds: 90,
					Active:                true,
					Fallback:              false,
					ChainIndex:            0,
					Capabilities:          map[string]any{"chat": true, "streaming": true, "auth_mode": "profile", "deployment_mode": "remote_api"},
					NorthboundProtocols:   []string{"chat.completions"},
					NorthboundRoutes:      []string{"/v1/models", "/v1/chat/completions"},
					CredentialState:       "configured",
					DeploymentMode:        "remote_api",
					AuthMode:              "profile",
					Streaming:             true,
				},
				{
					Profile:               "local",
					NormalizedProfile:     "LOCAL",
					ProviderID:            "ollama",
					ResolvedModel:         "qwen2.5:7b-instruct",
					APIStyle:              "ollama-chat-completions",
					HasBaseURL:            true,
					HasAPIKey:             false,
					RequestTimeoutSeconds: 90,
					Active:                true,
					Fallback:              false,
					ChainIndex:            0,
					Capabilities:          map[string]any{"chat": true, "streaming": true, "auth_mode": "optional", "deployment_mode": "local_daemon"},
					NorthboundProtocols:   []string{"chat.completions"},
					NorthboundRoutes:      []string{"/v1/models", "/v1/chat/completions"},
					CredentialState:       "optional_missing",
					DeploymentMode:        "local_daemon",
					AuthMode:              "optional",
					Streaming:             true,
				},
			},
			ActiveProfileChain: []string{
				"profile:anyrouter",
				"provider:huggingface:profile:hf-chat",
				"provider:ollama:profile:local",
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
			BaseURL:         "https://configured.example/v1",
			APIKey:          "configured-key",
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

		northbound := canonicalNorthboundContext(t, decodeCanonicalRequest(t, fb.lastRequest))
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
		fb.streamCalled = false
		body := `{"model":"platform-model","input":"Please answer this","stream":true}`
		req := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(body))
		rec := httptest.NewRecorder()

		server.handleResponses(rec, req)

		if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/event-stream") {
			t.Fatalf("unexpected content type: %s", got)
		}
		if !fb.streamCalled {
			t.Fatalf("expected bridge stream path")
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

		northbound := canonicalNorthboundContext(t, decodeCanonicalRequest(t, fb.lastRequest))
		if northbound["protocol"] != northboundProtocolMessages {
			t.Fatalf("unexpected protocol: %#v", northbound["protocol"])
		}
		if northbound["request_kind"] != "anthropic_messages" {
			t.Fatalf("unexpected request kind: %#v", northbound["request_kind"])
		}
	})

	t.Run("messages stream", func(t *testing.T) {
		fb.streamCalled = false
		body := `{"model":"platform-model","system":"You are helpful","messages":[{"role":"user","content":"你好"}],"stream":true}`
		req := httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(body))
		rec := httptest.NewRecorder()

		server.handleMessages(rec, req)

		if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/event-stream") {
			t.Fatalf("unexpected content type: %s", got)
		}
		if !fb.streamCalled {
			t.Fatalf("expected bridge stream path")
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

	t.Run("chat profile selection", func(t *testing.T) {
		body := `{"model":"profile:anyrouter","messages":[{"role":"user","content":"hello"}]}`
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
		rec := httptest.NewRecorder()

		server.handleChatCompletions(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
		}
		if fb.lastOptions.Provider != "openai-compatible" {
			t.Fatalf("unexpected bridge provider: %s", fb.lastOptions.Provider)
		}
		if fb.lastOptions.ProviderProfile != "anyrouter" {
			t.Fatalf("unexpected bridge provider profile: %s", fb.lastOptions.ProviderProfile)
		}
		if fb.lastOptions.Model != "deepseek-chat" {
			t.Fatalf("unexpected bridge model: %s", fb.lastOptions.Model)
		}

		northbound := canonicalNorthboundContext(t, decodeCanonicalRequest(t, fb.lastRequest))
		if northbound["selected_provider"] != "openai-compatible" {
			t.Fatalf("unexpected selected provider: %#v", northbound["selected_provider"])
		}
		if northbound["selected_provider_profile"] != "anyrouter" {
			t.Fatalf("unexpected selected provider profile: %#v", northbound["selected_provider_profile"])
		}
		if northbound["upstream_model"] != "deepseek-chat" {
			t.Fatalf("unexpected upstream model: %#v", northbound["upstream_model"])
		}
		if northbound["credential_state"] != "configured" {
			t.Fatalf("unexpected credential state: %#v", northbound["credential_state"])
		}
		if northbound["deployment_mode"] != "remote_api" {
			t.Fatalf("unexpected deployment mode: %#v", northbound["deployment_mode"])
		}
		if northbound["selection_inventory_kind"] != "provider_profile" {
			t.Fatalf("unexpected selection inventory kind: %#v", northbound["selection_inventory_kind"])
		}
		if routes, ok := northbound["northbound_routes"].([]any); !ok || len(routes) == 0 {
			t.Fatalf("missing northbound routes: %#v", northbound["northbound_routes"])
		}
	})

	t.Run("northbound request observability", func(t *testing.T) {
		cases := []struct {
			name     string
			route    string
			body     string
			protocol string
			handle   func(http.ResponseWriter, *http.Request)
		}{
			{
				name:     "chat",
				route:    "/v1/chat/completions",
				body:     `{"model":"profile:anyrouter","messages":[{"role":"user","content":"hello"}]}`,
				protocol: northboundProtocolChatCompletions,
				handle:   server.handleChatCompletions,
			},
			{
				name:     "responses",
				route:    "/v1/responses",
				body:     `{"model":"profile:anyrouter","input":"Please answer this"}`,
				protocol: northboundProtocolResponses,
				handle:   server.handleResponses,
			},
			{
				name:     "messages",
				route:    "/v1/messages",
				body:     `{"model":"profile:anyrouter","messages":[{"role":"user","content":"hello"}]}`,
				protocol: northboundProtocolMessages,
				handle:   server.handleMessages,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				requestID := "req-observe-" + tc.name
				req := httptest.NewRequest(http.MethodPost, tc.route, strings.NewReader(tc.body))
				req.Header.Set("X-Request-Id", requestID)
				rec := httptest.NewRecorder()

				tc.handle(rec, req)

				if rec.Code != http.StatusOK {
					t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
				}
				if got := rec.Header().Get("X-Request-Id"); got != requestID {
					t.Fatalf("unexpected response request id header: %s", got)
				}
				canonicalRequest := decodeCanonicalRequest(t, fb.lastRequest)
				if canonicalRequest["request_id"] != requestID {
					t.Fatalf("unexpected canonical request id: %#v", canonicalRequest["request_id"])
				}
				contextBlock := canonicalRequest["context"].(map[string]any)
				if contextBlock["route"] != tc.route {
					t.Fatalf("unexpected canonical route: %#v", contextBlock["route"])
				}
				northbound := canonicalNorthboundContext(t, canonicalRequest)
				if northbound["request_id"] != requestID {
					t.Fatalf("unexpected northbound request id: %#v", northbound["request_id"])
				}
				if northbound["protocol"] != tc.protocol {
					t.Fatalf("unexpected protocol: %#v", northbound["protocol"])
				}
				if northbound["selected_provider"] != "openai-compatible" {
					t.Fatalf("unexpected selected provider: %#v", northbound["selected_provider"])
				}
				if northbound["selected_provider_profile"] != "anyrouter" {
					t.Fatalf("unexpected selected provider profile: %#v", northbound["selected_provider_profile"])
				}
				if northbound["selected_model"] != "profile:anyrouter" {
					t.Fatalf("unexpected selected model: %#v", northbound["selected_model"])
				}
				if northbound["upstream_model"] != "deepseek-chat" {
					t.Fatalf("unexpected upstream model: %#v", northbound["upstream_model"])
				}
				if northbound["selection_source"] != "requested_profile_model" {
					t.Fatalf("unexpected selection source: %#v", northbound["selection_source"])
				}
			})
		}
	})

	t.Run("error envelope observability", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(`{`))
		req.Header.Set("X-Request-Id", "req-error-123")
		rec := httptest.NewRecorder()

		server.handleResponses(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
		}
		if got := rec.Header().Get("X-Request-Id"); got != "req-error-123" {
			t.Fatalf("unexpected response request id header: %s", got)
		}
		var response errorDocument
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode error response: %v", err)
		}
		if response.Error.Code != "INVALID_JSON" {
			t.Fatalf("unexpected error code: %#v", response.Error.Code)
		}
		if response.Error.RequestID != "req-error-123" {
			t.Fatalf("unexpected error request id: %#v", response.Error.RequestID)
		}
		if response.Error.Route != "/v1/responses" {
			t.Fatalf("unexpected error route: %#v", response.Error.Route)
		}
		if response.Error.FailureBoundary != errorBoundaryNorthboundRequest {
			t.Fatalf("unexpected failure boundary: %#v", response.Error.FailureBoundary)
		}
		if response.Error.Metadata["latency_ms"] == nil {
			t.Fatalf("expected error latency metadata")
		}
	})

	t.Run("chat provider profile selection", func(t *testing.T) {
		body := `{"model":"provider:huggingface:profile:hf-chat","messages":[{"role":"user","content":"hello"}]}`
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
		rec := httptest.NewRecorder()

		server.handleChatCompletions(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
		}
		if fb.lastOptions.Provider != "huggingface" {
			t.Fatalf("unexpected bridge provider: %s", fb.lastOptions.Provider)
		}
		if fb.lastOptions.ProviderProfile != "hf-chat" {
			t.Fatalf("unexpected bridge provider profile: %s", fb.lastOptions.ProviderProfile)
		}
		if fb.lastOptions.Model != "meta-llama/Meta-Llama-3.1-8B-Instruct" {
			t.Fatalf("unexpected bridge model: %s", fb.lastOptions.Model)
		}
	})

	t.Run("chat concrete model override", func(t *testing.T) {
		body := `{"model":"custom-runtime-model","messages":[{"role":"user","content":"hello"}]}`
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
		rec := httptest.NewRecorder()

		server.handleChatCompletions(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
		}
		if fb.lastOptions.Provider != "mock" {
			t.Fatalf("unexpected bridge provider: %s", fb.lastOptions.Provider)
		}
		if fb.lastOptions.Model != "custom-runtime-model" {
			t.Fatalf("unexpected bridge model: %s", fb.lastOptions.Model)
		}
		if fb.lastOptions.BaseURL != "https://configured.example/v1" {
			t.Fatalf("unexpected forwarded base url: %s", fb.lastOptions.BaseURL)
		}
		if fb.lastOptions.APIKey != "configured-key" {
			t.Fatalf("unexpected forwarded api key: %s", fb.lastOptions.APIKey)
		}
	})

	t.Run("chat explicit provider isolation", func(t *testing.T) {
		body := `{"model":"qwen3:8b","messages":[{"role":"user","content":"hello"}],"radishmind":{"provider":"ollama"}}`
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
		rec := httptest.NewRecorder()

		server.handleChatCompletions(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
		}
		if fb.lastOptions.Provider != "ollama" {
			t.Fatalf("unexpected bridge provider: %s", fb.lastOptions.Provider)
		}
		if fb.lastOptions.ProviderProfile != "" {
			t.Fatalf("unexpected bridge provider profile: %s", fb.lastOptions.ProviderProfile)
		}
		if fb.lastOptions.Model != "qwen3:8b" {
			t.Fatalf("unexpected bridge model: %s", fb.lastOptions.Model)
		}
		if fb.lastOptions.BaseURL != "" {
			t.Fatalf("expected provider-specific base url lookup, got %s", fb.lastOptions.BaseURL)
		}
		if fb.lastOptions.APIKey != "" {
			t.Fatalf("expected provider-specific api key lookup, got %s", fb.lastOptions.APIKey)
		}
	})

	t.Run("chat explicit provider profile selection", func(t *testing.T) {
		body := `{"messages":[{"role":"user","content":"hello"}],"radishmind":{"provider":"huggingface","provider_profile":"hf-chat"}}`
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
		rec := httptest.NewRecorder()

		server.handleChatCompletions(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
		}
		if fb.lastOptions.Provider != "huggingface" {
			t.Fatalf("unexpected bridge provider: %s", fb.lastOptions.Provider)
		}
		if fb.lastOptions.ProviderProfile != "hf-chat" {
			t.Fatalf("unexpected bridge provider profile: %s", fb.lastOptions.ProviderProfile)
		}
		if fb.lastOptions.Model != "meta-llama/Meta-Llama-3.1-8B-Instruct" {
			t.Fatalf("unexpected bridge model: %s", fb.lastOptions.Model)
		}
	})

	t.Run("models", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
		rec := httptest.NewRecorder()

		handleModels(rec, req, server)

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
		defaultModel, ok := findNorthboundModel(response.Data, "platform-model")
		if !ok {
			t.Fatalf("missing configured default model: %#v", response.Data)
		}
		if defaultModel.Metadata["source"] != "configured_default" {
			t.Fatalf("unexpected metadata source: %#v", defaultModel.Metadata["source"])
		}
		if got, ok := defaultModel.Metadata["profile_inventory_count"].(float64); !ok || got != 3 {
			t.Fatalf("unexpected profile inventory count: %#v", defaultModel.Metadata["profile_inventory_count"])
		}
		if got := defaultModel.Metadata["active_profile_chain"]; got == nil {
			t.Fatalf("expected active profile chain metadata")
		}

		if _, ok := findNorthboundModel(response.Data, "huggingface"); !ok {
			t.Fatalf("missing huggingface provider model: %#v", response.Data)
		}
		if _, ok := findNorthboundModel(response.Data, "ollama"); !ok {
			t.Fatalf("missing ollama provider model: %#v", response.Data)
		}
		if _, ok := findNorthboundModel(response.Data, "provider:huggingface:profile:hf-chat"); !ok {
			t.Fatalf("missing huggingface profile model: %#v", response.Data)
		}
		if _, ok := findNorthboundModel(response.Data, "provider:ollama:profile:local"); !ok {
			t.Fatalf("missing ollama profile model: %#v", response.Data)
		}

		profileModel, ok := findNorthboundModel(response.Data, "profile:anyrouter")
		if !ok {
			t.Fatalf("missing profile inventory model: %#v", response.Data)
		}
		if profileModel.Metadata["source"] != "provider_profile_inventory" {
			t.Fatalf("unexpected profile inventory source: %#v", profileModel.Metadata["source"])
		}
		if profileModel.Metadata["credential_state"] != "configured" {
			t.Fatalf("unexpected profile credential state: %#v", profileModel.Metadata["credential_state"])
		}
		if profileModel.Metadata["deployment_mode"] != "remote_api" {
			t.Fatalf("unexpected profile deployment mode: %#v", profileModel.Metadata["deployment_mode"])
		}
		if profileModel.Metadata["streaming"] != true {
			t.Fatalf("unexpected profile streaming flag: %#v", profileModel.Metadata["streaming"])
		}
		selectionMetadata, ok := profileModel.Metadata["selection"].(map[string]any)
		if !ok {
			t.Fatalf("missing profile selection metadata: %#v", profileModel.Metadata["selection"])
		}
		if selectionMetadata["selected_provider"] != "openai-compatible" {
			t.Fatalf("unexpected selection provider: %#v", selectionMetadata["selected_provider"])
		}
		if selectionMetadata["credential_state"] != "configured" {
			t.Fatalf("unexpected selection credential state: %#v", selectionMetadata["credential_state"])
		}
		if selectionMetadata["selection_inventory_kind"] != "provider_profile" {
			t.Fatalf("unexpected selection inventory kind: %#v", selectionMetadata["selection_inventory_kind"])
		}
	})

	t.Run("model detail default", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/models/platform-model", nil)
		req.SetPathValue("id", "platform-model")
		rec := httptest.NewRecorder()

		server.handleModel(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
		}
		var response openAIModelObject
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if response.ID != "platform-model" {
			t.Fatalf("unexpected model id: %s", response.ID)
		}
		if response.Metadata["source"] != "configured_default" {
			t.Fatalf("unexpected metadata source: %#v", response.Metadata["source"])
		}
	})

	t.Run("model detail provider", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/models/mock", nil)
		req.SetPathValue("id", "mock")
		rec := httptest.NewRecorder()

		server.handleModel(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
		}
		var response openAIModelObject
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if response.ID != "mock" {
			t.Fatalf("unexpected model id: %s", response.ID)
		}
		if response.Metadata["source"] != "provider_registry" {
			t.Fatalf("unexpected metadata source: %#v", response.Metadata["source"])
		}
	})

	t.Run("model detail provider alias", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/models/provider:mock", nil)
		req.SetPathValue("id", "provider:mock")
		rec := httptest.NewRecorder()

		server.handleModel(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
		}
		var response openAIModelObject
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if response.ID != "mock" {
			t.Fatalf("unexpected model id: %s", response.ID)
		}
	})

	t.Run("model detail profile", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/models/profile:anyrouter", nil)
		req.SetPathValue("id", "profile:anyrouter")
		rec := httptest.NewRecorder()

		server.handleModel(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
		}
		var response openAIModelObject
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if response.ID != "profile:anyrouter" {
			t.Fatalf("unexpected model id: %s", response.ID)
		}
		if response.Metadata["source"] != "provider_profile_inventory" {
			t.Fatalf("unexpected metadata source: %#v", response.Metadata["source"])
		}
		if response.Metadata["resolved_model"] != "deepseek-chat" {
			t.Fatalf("unexpected resolved model: %#v", response.Metadata["resolved_model"])
		}
	})

	t.Run("model detail profile alias", func(t *testing.T) {
		alias := "provider:openai-compatible:profile:anyrouter"
		req := httptest.NewRequest(http.MethodGet, "/v1/models/"+alias, nil)
		req.SetPathValue("id", alias)
		rec := httptest.NewRecorder()

		server.handleModel(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
		}
		var response openAIModelObject
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if response.ID != "profile:anyrouter" {
			t.Fatalf("unexpected model id: %s", response.ID)
		}
	})

	t.Run("model detail provider profile", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/models/provider:huggingface:profile:hf-chat", nil)
		req.SetPathValue("id", "provider:huggingface:profile:hf-chat")
		rec := httptest.NewRecorder()

		server.handleModel(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
		}
		var response openAIModelObject
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if response.ID != "provider:huggingface:profile:hf-chat" {
			t.Fatalf("unexpected model id: %s", response.ID)
		}
		if response.Metadata["provider_id"] != "huggingface" {
			t.Fatalf("unexpected provider id: %#v", response.Metadata["provider_id"])
		}
		if response.Metadata["provider_profile"] != "hf-chat" {
			t.Fatalf("unexpected provider profile: %#v", response.Metadata["provider_profile"])
		}
		if response.Metadata["credential_state"] != "configured" {
			t.Fatalf("unexpected credential state: %#v", response.Metadata["credential_state"])
		}
		if response.Metadata["deployment_mode"] != "remote_api" {
			t.Fatalf("unexpected deployment mode: %#v", response.Metadata["deployment_mode"])
		}
	})

	t.Run("model detail missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/models/does-not-exist", nil)
		req.SetPathValue("id", "does-not-exist")
		rec := httptest.NewRecorder()

		server.handleModel(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
		}
		var response errorDocument
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode error response: %v", err)
		}
		if response.Error.Code != "MODEL_NOT_FOUND" {
			t.Fatalf("unexpected error code: %#v", response.Error.Code)
		}
	})
}
