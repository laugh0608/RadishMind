package httpapi

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
)

type gatewayRouteBenchmarkBridge struct{}

func (gatewayRouteBenchmarkBridge) DescribeProviders(context.Context) ([]bridge.ProviderDescription, error) {
	return nil, nil
}

func (gatewayRouteBenchmarkBridge) DescribeInventory(context.Context) (bridge.ProviderInventory, error) {
	return bridge.ProviderInventory{}, nil
}

func (gatewayRouteBenchmarkBridge) HandleEnvelope(
	context.Context,
	[]byte,
	bridge.EnvelopeOptions,
) (bridge.GatewayEnvelope, error) {
	return bridge.GatewayEnvelope{
		SchemaVersion: 1,
		Status:        "ok",
		RequestID:     "gateway-route-benchmark",
		Project:       "radish",
		Task:          "answer_docs_question",
		Response: map[string]any{
			"summary": "benchmark response",
		},
		Metadata: map[string]any{
			"duration_ms":          0,
			"provider_duration_ms": 0,
		},
	}, nil
}

func (gatewayRouteBenchmarkBridge) StreamEnvelope(
	context.Context,
	[]byte,
	bridge.EnvelopeOptions,
	func(bridge.StreamEvent) error,
) error {
	return nil
}

func BenchmarkGatewayChatCompletionsRoute(b *testing.B) {
	previousLogWriter := log.Writer()
	log.SetOutput(io.Discard)
	b.Cleanup(func() {
		log.SetOutput(previousLogWriter)
	})

	server := &Server{
		bridge: gatewayRouteBenchmarkBridge{},
		config: config.Config{
			BridgeTimeout: time.Second,
			Provider:      "mock",
		},
		options: Options{BuildVersion: "benchmark"},
	}
	const requestBody = `{"messages":[{"role":"user","content":"benchmark gateway route"}]}`

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		request := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(requestBody))
		recorder := httptest.NewRecorder()
		server.handleChatCompletions(recorder, request)
		if recorder.Code != http.StatusOK {
			b.Fatalf("unexpected benchmark response: status=%d body=%s", recorder.Code, recorder.Body.String())
		}
	}
}
