package developmentcontrol

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func nativeHTTPTestBody(result NativeResult) string {
	doc := map[string]any{"id": "fixed-response", "object": "chat.completion", "created": int64(1), "model": "fixture-model", "choices": []any{map[string]any{"index": 0, "message": result.Message, "finish_reason": result.FinishReason}}, "usage": result.Usage}
	data, _ := json.Marshal(doc)
	return string(data)
}

func nativeHTTPTestConnection(endpoint string) NativeHTTPConnection {
	return NativeHTTPConnection{ModelRef: "profile:fixture", Model: "fixture-model", BaseURL: endpoint + "/v1", APIKey: "synthetic-provider-secret", Policy: NativePolicy{100, 50, time.Second}}
}

func nativeHTTPTestBackend(t *testing.T, connection NativeHTTPConnection) NativeBackend {
	t.Helper()
	b, err := NewNativeHTTPBackend(context.Background(), func(context.Context) (NativeHTTPConnection, error) { return connection, nil }, true)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestNativeHTTPFailuresAndNoRetry(t *testing.T) {
	for _, tc := range []struct {
		name       string
		status     int
		body, code string
	}{
		{"unauthorized", 401, "synthetic-provider-secret", "native_authentication_failed"},
		{"rate", 429, "synthetic-provider-secret", "native_rate_limited"},
		{"server", 503, "synthetic-provider-secret", "native_upstream_failed"},
		{"redirect", 307, "synthetic-provider-secret", "native_redirect_denied"},
		{"json", 200, "{", "native_response_invalid"},
		{"oversize", 200, strings.Repeat("x", nativeHTTPBodyLimit+1), "native_response_too_large"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Location", "http://127.0.0.1:1/stolen")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer upstream.Close()
			b := nativeHTTPTestBackend(t, nativeHTTPTestConnection(upstream.URL))
			_, err := b.Complete(context.Background(), "profile:fixture", initialRequest(t))
			if err == nil || err.Error() != tc.code || !errors.Is(err, ErrTransport) || calls.Load() != 1 {
				t.Fatalf("code=%v calls=%d", err, calls.Load())
			}
			w := httptest.NewRecorder()
			channelError(w, err)
			if !strings.Contains(w.Body.String(), tc.code) || strings.Contains(w.Body.String(), "synthetic-provider-secret") {
				t.Fatal("unsafe/missing failure category")
			}
		})
	}
}

func TestNativeHTTPCancelAndTimeout(t *testing.T) {
	for _, mode := range []string{"cancel", "timeout"} {
		t.Run(mode, func(t *testing.T) {
			started, stopped := make(chan struct{}), make(chan struct{})
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.Copy(io.Discard, r.Body)
				close(started)
				<-r.Context().Done()
				close(stopped)
			}))
			defer upstream.Close()
			c := nativeHTTPTestConnection(upstream.URL)
			c.Policy.Timeout = 100 * time.Millisecond
			b := nativeHTTPTestBackend(t, c)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() { _, err := b.Complete(ctx, c.ModelRef, initialRequest(t)); done <- err }()
			select {
			case <-started:
			case <-time.After(time.Second):
				t.Fatal("not dispatched")
			}
			if mode == "cancel" {
				cancel()
			}
			select {
			case err := <-done:
				want := "native_timeout"
				if mode == "cancel" {
					want = "native_canceled"
				}
				if err == nil || err.Error() != want {
					t.Fatal(err)
				}
			case <-time.After(time.Second):
				t.Fatal("request did not stop")
			}
			select {
			case <-stopped:
			case <-time.After(time.Second):
				t.Fatal("upstream connection remained live")
			}
		})
	}
}

func TestNativeHTTPConnectionBoundaries(t *testing.T) {
	for _, raw := range []string{"http://example.com/v1", "https://user:password@example.com/v1", "https://example.com/v1?key=secret", "https://example.com/v1#fragment", "http://127.0.0.1:1234/v1", "https://example.com:0/v1", "https://example.com/v1?"} {
		c := nativeHTTPTestConnection("https://example.com")
		c.BaseURL = raw
		if _, err := NewNativeHTTPBackend(context.Background(), func(context.Context) (NativeHTTPConnection, error) { return c, nil }, false); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
	for _, raw := range []string{"127.0.0.1", "10.0.0.1", "169.254.169.254", "::1", "::ffff:127.0.0.1", "100.64.0.1", "192.0.2.1", "fc00::1", "64:ff9b::7f00:1"} {
		if nativePublicAddress(netip.MustParseAddr(raw)) {
			t.Fatalf("allowed reserved address %s", raw)
		}
	}
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, nativeHTTPTestBody(stopResult()))
	}))
	defer upstream.Close()
	c := nativeHTTPTestConnection(upstream.URL)
	b, err := NewNativeHTTPBackend(context.Background(), func(context.Context) (NativeHTTPConnection, error) { return c, nil }, true)
	if err != nil {
		t.Fatal(err)
	}
	c.APIKey = "rotated-synthetic-key"
	if _, err = b.Complete(context.Background(), c.ModelRef, initialRequest(t)); err == nil || err.Error() != "native_configuration_changed" || calls.Load() != 0 {
		t.Fatal("changed credentials dispatched", err)
	}
	if _, err = b.Policy(context.Background(), c.ModelRef); err == nil {
		t.Fatal("changed policy accepted")
	}
	_, err = nativeHTTPDialer("127.0.0.1", false)(context.Background(), "tcp", "127.0.0.1:1234")
	if err == nil || err.Error() != "native_endpoint_denied" {
		t.Fatal("private destination accepted", err)
	}
}

func TestNativeHTTPWireContractAndUnknownUsage(t *testing.T) {
	base := nativeHTTPTestBody(stopResult())
	for _, tc := range []struct {
		name, body     string
		valid, unknown bool
	}{
		{"valid", base, true, false},
		{"tool", nativeHTTPTestBody(toolResult()), true, false},
		{"null_usage", strings.Replace(base, `"usage":{"prompt_tokens":30,"completion_tokens":10,"total_tokens":40}`, `"usage":null`, 1), true, true},
		{"partial_usage", strings.Replace(base, `"prompt_tokens":30,`, "", 1), false, false},
		{"null_counter", strings.Replace(base, `"prompt_tokens":30`, `"prompt_tokens":null`, 1), false, false},
		{"duplicate", strings.Replace(base, `"object":"chat.completion"`, `"object":"chat.completion","object":"chat.completion"`, 1), false, false},
		{"case", strings.Replace(base, `"choices"`, `"Choices"`, 1), false, false},
		{"wrong_model", strings.Replace(base, "fixture-model", "other-model", 1), false, false},
		{"wrong_finish", strings.Replace(base, `"finish_reason":"stop"`, `"finish_reason":"length"`, 1), false, false},
		{"wrong_total", strings.Replace(base, `"total_tokens":40`, `"total_tokens":0`, 1), false, false},
		{"unknown_tool", strings.Replace(nativeHTTPTestBody(toolResult()), "development_tool", "unknown_tool", 1), false, false},
		{"trailing", base + " {}", false, false},
		{"depth", strings.Repeat("[", 26) + "0" + strings.Repeat("]", 26), false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := parseNativeHTTPResponse([]byte(tc.body), "fixture-model")
			if (err == nil) != tc.valid {
				t.Fatal("unexpected validity", err)
			}
			if tc.valid && (r.Usage == nil) != tc.unknown {
				t.Fatal("unknown usage changed")
			}
		})
	}
}
