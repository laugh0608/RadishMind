package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
)

// The legacy advisory endpoint is not an SDK tool transport. A successful HTTP
// response must never be interpreted as development-tool protocol acceptance.
func TestSDKToolWireDoesNotGrantLegacyToolExecution(t *testing.T) {
	for _, name := range []string{"sdk-first-request.json", "sdk-observation-request.json"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("../../../../scripts/eval/development_probe/fixtures", name)
			payload, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var request map[string]any
			if err := json.Unmarshal(payload, &request); err != nil {
				t.Fatal(err)
			}
			// Only selection changes: the model metadata name in the SDK fixture is
			// replaced by the configured local mock; native tool fields stay intact.
			delete(request, "model")
			payload, err = json.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}
			fb := &fakeBridge{envelope: bridge.GatewayEnvelope{
				SchemaVersion: 1, Status: "ok", Project: "radish", Task: "answer_docs_question",
				Response: map[string]any{"summary": "fixed legacy advisory response"},
			}}
			server := &Server{bridge: fb, config: config.Config{Provider: "mock", BridgeTimeout: time.Second}}
			recorder := httptest.NewRecorder()
			server.handleChatCompletions(recorder, httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(payload)))
			if recorder.Code != http.StatusOK || fb.handleCalls != 1 {
				t.Fatalf("legacy fixture behavior changed: status=%d calls=%d body=%s", recorder.Code, fb.handleCalls, recorder.Body.String())
			}
			var canonical struct {
				ToolHints map[string]bool `json:"tool_hints"`
			}
			if err := json.Unmarshal(fb.lastRequest, &canonical); err != nil {
				t.Fatal(err)
			}
			allowed, present := canonical.ToolHints["allow_tool_calls"]
			if !present || allowed || strings.Contains(string(fb.lastRequest), "call-probe-001") {
				t.Fatal("legacy bridge acquired tool permissions or forwarded a tool-call identity")
			}
			var response map[string]any
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			choices := response["choices"].([]any)
			choice := choices[0].(map[string]any)
			message := choice["message"].(map[string]any)
			if choice["finish_reason"] != "stop" || message["tool_calls"] != nil {
				t.Fatal("legacy response became a development-tool completion")
			}
			t.Log("observed: legacy HTTP 200 is advisory-only; native SDK tool transport is unsupported")
		})
	}
}
