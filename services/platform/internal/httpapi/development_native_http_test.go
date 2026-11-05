package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
	control "radishmind.local/services/platform/internal/developmentcontrol"
)

func newNativeHTTPServerFixture(t *testing.T, endpoint string, timeout time.Duration) *developmentServerFixture {
	t.Helper()
	return newDevelopmentServerFixtureWithConnector(t, func(s *Server) (*control.Runtime, error) {
		s.options.TestOnly = true
		s.config.Provider = "openai-compatible"
		s.config.ProviderProfile = "development"
		s.config.Model = "synthetic-model"
		s.config.BaseURL = endpoint + "/v1"
		// Exercise the existing credential-file owner; no credential is supplied
		// by the SDK request or synthetic inventory.
		s.config.APIKey = ""
		s.config.APIKeyFile = filepath.Join(t.TempDir(), "provider-key")
		if err := os.WriteFile(s.config.APIKeyFile, []byte("synthetic-provider-secret\n"), 0600); err != nil {
			return nil, err
		}
		if err := config.ResolveWorkbenchCredential(&s.config); err != nil {
			return nil, err
		}
		return s.ConnectDevelopmentNativeHTTP(context.Background(), control.NativePolicy{InputTokenCeiling: 100, MaxOutputTokens: 50, Timeout: timeout})
	})
}

func issueNativeHTTPAttempt(t *testing.T, f *developmentServerFixture) control.Issued {
	t.Helper()
	p, err := f.server.PrepareDevelopmentAttempt(f.request, f.plan)
	if err != nil {
		t.Fatal(err)
	}
	i, err := f.server.IssueDevelopmentAttempt(f.request, f.plan, "confirmation_http", p)
	if err != nil {
		t.Fatal(err)
	}
	return i
}

func nativeHTTPChannelRequest(t *testing.T, f *developmentServerFixture, i control.Issued, path, id string, body any) *httptest.ResponseRecorder {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("POST", "http://127.0.0.1/internal/development/v1/attempts/"+i.Delegation.Scope.AttemptID+path, bytes.NewReader(data))
	r.RemoteAddr = "127.0.0.1:12345"
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+i.Credential)
	r.Header.Set("X-RadishMind-Development-Request-ID", id)
	w := httptest.NewRecorder()
	f.runtime.Handler().ServeHTTP(w, r)
	return w
}

func nativeHTTPInitialRequest(t *testing.T) control.NativeRequest {
	t.Helper()
	tools, err := control.NativeTools()
	if err != nil {
		t.Fatal(err)
	}
	return control.NativeRequest{Model: "synthetic-model", Messages: []control.NativeMessage{{Role: "user", Content: "synthetic task"}}, MaxCompletionTokens: 50, Tools: tools}
}

func TestDevelopmentNativeHTTPTwoToolRoundTrips(t *testing.T) {
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" || r.Method != "POST" || r.Header.Get("Authorization") != "Bearer synthetic-provider-secret" {
			t.Error("incorrect trusted connection")
		}
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		input, err := control.ParseNativeRequest(data)
		if err != nil {
			t.Error(err)
			return
		}
		var wire map[string]json.RawMessage
		if json.Unmarshal(data, &wire) != nil || string(wire["stream"]) != "false" || string(wire["parallel_tool_calls"]) != "false" {
			t.Error("provider defaults were not disabled")
		}
		n := calls.Add(1)
		if len(input.Messages) != int(2*n-1) || input.Stream || input.ParallelToolCalls {
			t.Error("tool conversation not preserved")
		}
		message := control.NativeMessage{Role: "assistant", Content: "fixed completion"}
		finish := "stop"
		if n < 3 {
			action, _ := json.Marshal(control.ToolAction{SchemaVersion: "development_tool_action.v1", Action: "read_file", Path: "main.go"})
			message = control.NativeMessage{Role: "assistant", ToolCalls: []control.NativeCall{{ID: fmt.Sprintf("call_http_%d", n), Type: "function", Function: control.NativeFunction{Name: "development_tool", Arguments: string(action)}}}}
			finish = "tool_calls"
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "fixed", "object": "chat.completion", "created": 1, "model": "synthetic-model", "choices": []any{map[string]any{"index": 0, "message": message, "finish_reason": finish}}, "usage": map[string]int{"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15}})
	}))
	defer upstream.Close()
	// Ambient proxy settings must never route a credential-bearing request.
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:1")
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:1")
	t.Setenv("NO_PROXY", "")
	f := newNativeHTTPServerFixture(t, upstream.URL, time.Second)
	f.plan.Limits.MaxModelCalls = 3
	i := issueNativeHTTPAttempt(t, f)
	input := nativeHTTPInitialRequest(t)
	for n := 1; n <= 3; n++ {
		response := nativeHTTPChannelRequest(t, f, i, "/chat/completions", fmt.Sprintf("model_http_%d", n), input)
		if response.Code != 200 {
			t.Fatalf("model: %d %s", response.Code, response.Body)
		}
		if strings.Contains(response.Body.String(), "synthetic-provider-secret") {
			t.Fatal("credential returned")
		}
		var doc struct {
			Choices []struct {
				Message control.NativeMessage `json:"message"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &doc); err != nil || len(doc.Choices) != 1 {
			t.Fatal("invalid response", err)
		}
		if n == 3 {
			break
		}
		record, err := f.runtime.Read(context.Background(), f.identity, i.Delegation.Scope.AttemptID)
		if err != nil {
			t.Fatal(err)
		}
		var offer control.Offer
		for _, o := range record.Offers {
			if o.ToolCallID == fmt.Sprintf("call_http_%d", n) {
				offer = o
			}
		}
		call := control.ToolCall{SchemaVersion: "development_tool_call.v1", AttemptID: i.Delegation.Scope.AttemptID, RequestID: offer.RequestID, ParentRequestID: offer.ParentRequestID, ToolCallID: offer.ToolCallID, Arguments: control.ToolAction{SchemaVersion: "development_tool_action.v1", Action: "read_file", Path: "main.go"}}
		forged := call
		forged.ToolCallID = "call_wrong"
		if w := nativeHTTPChannelRequest(t, f, i, "/tool-calls", "", forged); w.Code == 202 {
			t.Fatal("wrong tool identity accepted")
		}
		if w := nativeHTTPChannelRequest(t, f, i, "/tool-calls", "", call); w.Code != 202 {
			t.Fatal(w.Body.String())
		}
		if _, err = f.runtime.DispatchTool(context.Background(), f.identity, call.AttemptID, call.RequestID); err != nil {
			t.Fatal(err)
		}
		if _, err = f.runtime.CompleteTool(context.Background(), f.identity, call.AttemptID, call.RequestID, "synthetic observation"); err != nil {
			t.Fatal(err)
		}
		input.Messages = append(input.Messages, doc.Choices[0].Message, control.NativeMessage{Role: "tool", Content: "synthetic observation", ToolCallID: call.ToolCallID, Name: "development_tool"})
	}
	record, err := f.runtime.Read(context.Background(), f.identity, i.Delegation.Scope.AttemptID)
	if err != nil {
		t.Fatal(err)
	}
	events := record.Control.Events()
	last := events[len(events)-1]
	if calls.Load() != 3 || last.Phase != "completed" || last.Budget.SpentTokens != 45 || f.bridge.lastRequest != nil {
		t.Fatalf("bad completion: %+v", last)
	}
	rows, err := f.server.localPersistenceRuntime.DB().Query(`SELECT state_record||binding_record FROM development_attempts UNION ALL SELECT event_record FROM development_events`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			t.Fatal(err)
		}
		for _, secret := range []string{"synthetic-provider-secret", i.Credential, "synthetic observation", upstream.URL} {
			if strings.Contains(text, secret) {
				t.Fatal("private transport material persisted")
			}
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
}

func TestDevelopmentNativeHTTPUnknownUsageAndCurrentConfiguration(t *testing.T) {
	for _, mode := range []string{"missing_usage", "partial_usage", "credential_changed", "endpoint_changed", "inventory_changed", "member_revoked"} {
		t.Run(mode, func(t *testing.T) {
			var calls atomic.Int32
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.Header().Set("Content-Type", "application/json")
				usage := ""
				if mode == "partial_usage" {
					usage = `,"usage":{"total_tokens":0}`
				}
				fmt.Fprintf(w, `{"id":"fixed","object":"chat.completion","created":1,"model":"synthetic-model","choices":[{"index":0,"message":{"role":"assistant","content":"done"},"finish_reason":"stop"}]%s}`, usage)
			}))
			defer upstream.Close()
			f := newNativeHTTPServerFixture(t, upstream.URL, time.Second)
			i := issueNativeHTTPAttempt(t, f)
			switch mode {
			case "credential_changed":
				f.server.config.APIKey = "changed-key"
			case "endpoint_changed":
				f.server.config.BaseURL = upstream.URL + "/other"
			case "inventory_changed":
				f.bridge.inventory.Profiles[0].Enabled = false
			case "member_revoked":
				if _, err := revokeDevelopmentFixtureMembership(t, f); err != nil {
					t.Fatal(err)
				}
			}
			w := nativeHTTPChannelRequest(t, f, i, "/chat/completions", "model_http", nativeHTTPInitialRequest(t))
			if mode == "missing_usage" || mode == "partial_usage" {
				if mode == "missing_usage" && (w.Code != 200 || strings.Contains(w.Body.String(), `"usage"`)) {
					t.Fatal("missing usage manufactured", w.Body.String())
				}
				if mode == "partial_usage" && (w.Code != 502 || !strings.Contains(w.Body.String(), "native_response_invalid")) {
					t.Fatal(w.Body.String())
				}
				r, err := f.runtime.Read(context.Background(), f.identity, i.Delegation.Scope.AttemptID)
				if err != nil {
					t.Fatal(err)
				}
				events := r.Control.Events()
				budget := events[len(events)-1].Budget
				if budget.Unresolved != 1 || budget.HeldTokens != 150 || budget.SpentTokens != 0 || calls.Load() != 1 {
					t.Fatalf("unknown accounting lost: %+v", budget)
				}
			} else if w.Code == 200 || calls.Load() != 0 {
				t.Fatal("changed authority dispatched", w.Code, calls.Load())
			}
		})
	}
}

func TestDevelopmentNativeHTTPCancelPreservesUnknownReceipt(t *testing.T) {
	started, stopped := make(chan struct{}), make(chan struct{})
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		close(started)
		<-r.Context().Done()
		close(stopped)
	}))
	defer upstream.Close()
	f := newNativeHTTPServerFixture(t, upstream.URL, 5*time.Second)
	i := issueNativeHTTPAttempt(t, f)
	input := nativeHTTPInitialRequest(t)
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() { done <- nativeHTTPChannelRequest(t, f, i, "/chat/completions", "model_cancel", input) }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("HTTP was not dispatched")
	}
	r, err := f.runtime.Read(context.Background(), f.identity, i.Delegation.Scope.AttemptID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.runtime.Cancel(context.Background(), f.identity, i.Delegation.Scope.AttemptID, r.Version); err != nil {
		t.Fatal(err)
	}
	select {
	case w := <-done:
		if w.Code != 502 || !strings.Contains(w.Body.String(), "native_canceled") {
			t.Fatal(w.Body.String())
		}
	case <-time.After(time.Second):
		t.Fatal("model request remained active")
	}
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("upstream connection remained active")
	}
	r, err = f.runtime.Read(context.Background(), f.identity, i.Delegation.Scope.AttemptID)
	if err != nil {
		t.Fatal(err)
	}
	events := r.Control.Events()
	last := events[len(events)-1]
	if !r.Revoked || last.Phase != "outcome_unknown" || last.Budget.Unresolved != 1 || last.Budget.HeldTokens != 150 || last.Budget.InFlight != 0 {
		t.Fatalf("cancellation lost accounting: %+v", last)
	}
	if w := nativeHTTPChannelRequest(t, f, i, "/chat/completions", "model_after_cancel", input); w.Code != 401 {
		t.Fatal("revoked bearer accepted", w.Code)
	}
}
