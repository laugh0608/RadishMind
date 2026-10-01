package httpapi

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"radishmind.local/services/platform/internal/bridge"
)

func TestPromptApplicationCanonicalSummaryPreservesOutputContractOwnership(t *testing.T) {
	source := PromptApplicationTemplateSource{
		Messages: []PromptApplicationTemplateMessage{
			{Role: "system", Content: "只输出诊断 JSON。"},
			{Role: "user", Content: "{{ question }}"},
		},
		Variables: []PromptApplicationTemplateVariable{{Name: "question", Type: "string", Required: true}},
		OutputContract: PromptApplicationOutputContract{
			Kind: "json_object", MaxBytes: 4096,
			JSONSchema: &PromptApplicationJSONSchema{
				Type: "object", Required: []string{"diagnosis"},
				Properties: map[string]PromptApplicationJSONSchema{"diagnosis": {Type: "string"}},
			},
		},
	}
	for _, test := range []struct {
		name, output, failure string
	}{
		{"json", `{"diagnosis":"检查端口监听"}`, ""},
		{"missing_field", `{}`, PromptApplicationInvocationFailureOutputContract},
		{"not_json", "检查端口监听", PromptApplicationInvocationFailureOutputContract},
		{"fenced_json", "```json\n{\"diagnosis\":\"检查端口监听\"}\n```", PromptApplicationInvocationFailureOutputContract},
		{"extra_field", `{"diagnosis":"检查端口监听","requires_confirmation":false}`, PromptApplicationInvocationFailureOutputContract},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := newPromptApplicationInvocationFixture(t, source)
			fixture.bridge.handle = func(_ context.Context, packet []byte, _ bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
				var request struct {
					Context struct {
						Route      string         `json:"route"`
						Northbound map[string]any `json:"northbound"`
					} `json:"context"`
					Artifacts []struct {
						Content string `json:"content"`
					} `json:"artifacts"`
				}
				if err := json.Unmarshal(packet, &request); err != nil {
					t.Fatal(err)
				}
				if request.Context.Route != promptApplicationInvocationRoute ||
					request.Context.Northbound["protocol"] != promptApplicationInvocationProtocol ||
					request.Context.Northbound["request_kind"] != promptApplicationInvocationProtocol ||
					request.Context.Northbound["allow_tool_calls"] != false || len(request.Artifacts) != 1 {
					t.Fatal("canonical Prompt request boundary drifted")
				}
				var messages []PromptApplicationTemplateMessage
				if err := json.Unmarshal([]byte(request.Artifacts[0].Content), &messages); err != nil {
					t.Fatal(err)
				}
				if len(messages) != 2 || messages[0] != source.Messages[0] || messages[1].Content != "日志中的 {{ literal }}" {
					t.Fatal("rendered message roles or literal input changed")
				}
				// Python carries opaque application output in the canonical summary.
				return bridge.GatewayEnvelope{Status: "ok", Response: map[string]any{"summary": test.output}}, nil
			}
			input := PromptApplicationInvocationInput{ClientInvocationKey: "summary-output", Variables: map[string]any{"question": "日志中的 {{ literal }}"}}
			result := fixture.service.Invoke(fixture.ctx, input)
			if result.FailureCode != test.failure || result.Run == nil || fixture.bridge.callCount() != 1 {
				t.Fatalf("output contract result: failure=%q calls=%d", result.FailureCode, fixture.bridge.callCount())
			}
			if test.failure == "" && result.Output != test.output || test.failure != "" && result.Output != "" {
				t.Fatal("output was rewritten or invalid output was released")
			}
			stored, err := json.Marshal(result.Run)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(stored), "检查端口监听") || strings.Contains(string(stored), "日志中的") {
				t.Fatal("application input or output leaked into Run")
			}
			replay := fixture.service.Invoke(fixture.ctx, input)
			if !replay.IdempotentReplay || replay.Output != "" || fixture.bridge.callCount() != 1 {
				t.Fatal("terminal replay called provider or restored transient output")
			}
		})
	}
}
