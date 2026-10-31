package developmentcontrol

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/sqlitedev"
	migrations "radishmind.local/services/platform/migrations/sqlite/development_attempts"
)

type currentQualification struct {
	mu        sync.Mutex
	selection Selection
	err       error
}

func (q *currentQualification) Resolve(context.Context, Identity, Plan) (Selection, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.selection, q.err
}
func (q *currentQualification) deny(err error) { q.mu.Lock(); defer q.mu.Unlock(); q.err = err }

type fixedNative struct {
	calls    atomic.Int32
	complete func(context.Context, NativeRequest) (NativeResult, error)
}

func (*fixedNative) Policy(context.Context, string) (NativePolicy, error) {
	return NativePolicy{100, 100, 5 * time.Second}, nil
}
func (b *fixedNative) Complete(ctx context.Context, _ string, input NativeRequest) (NativeResult, error) {
	b.calls.Add(1)
	return b.complete(ctx, input)
}

type runtimeFixture struct {
	runtime  *Runtime
	store    *SQLiteStore
	database *sqlitedev.Runtime
	reader   *currentQualification
	backend  *fixedNative
	identity Identity
	plan     Plan
	issued   Issued
	handler  http.Handler
}

func newRuntimeFixture(t *testing.T, complete func(context.Context, NativeRequest) (NativeResult, error)) *runtimeFixture {
	t.Helper()
	db, err := sqlitedev.Open(context.Background(), sqlitedev.Options{DatabasePath: filepath.Join(t.TempDir(), "control.db"), Migrations: migrations.Migrations()})
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewSQLiteStore(db.DB())
	if err != nil {
		t.Fatal(err)
	}
	sha := PayloadDigest([]byte("synthetic-current-authority"))
	reader := &currentQualification{selection: Selection{sha, sha, sha, "fixture-model", sha, 1000000, 1000000, 2000000}}
	backend := &fixedNative{complete: complete}
	runtime, err := NewRuntime(store, reader, backend)
	if err != nil {
		t.Fatal(err)
	}
	identity := Identity{"tenant_fixture", "usr_fixture", "workspace_fixture"}
	initialDigest, err := messagesDigest(initialRequest(t).Messages)
	if err != nil {
		t.Fatal(err)
	}
	plan := Plan{"development_attempt_plan.v1", "application_fixture", "prompt_invoke_v2", "", "profile:fixture", sha, initialDigest, ToolPolicy{[]string{"main.go"}, []string{"main.go"}, "go_test_v1"}, Limits{4, 4, 1000, 10000}, 60000, 50}
	prepared, err := runtime.Prepare(context.Background(), identity, plan)
	if err != nil {
		t.Fatal(err)
	}
	issued, err := runtime.Issue(context.Background(), identity, plan, "confirmation_fixture", prepared.PlanDigest, prepared.Selection)
	if err != nil {
		t.Fatal(err)
	}
	f := &runtimeFixture{runtime, store, db, reader, backend, identity, plan, issued, runtime.Handler()}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := f.runtime.Close(ctx); err != nil {
			t.Error(err)
		}
		if err := f.database.Close(); err != nil {
			t.Error(err)
		}
	})
	return f
}
func (f *runtimeFixture) read(t *testing.T) *Record {
	t.Helper()
	r, e := f.runtime.Read(context.Background(), f.identity, f.issued.Delegation.Scope.AttemptID)
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func (f *runtimeFixture) request(t *testing.T, method, path string, body any, requestID string) *httptest.ResponseRecorder {
	t.Helper()
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(method, "http://127.0.0.1/internal/development/v1/attempts/"+f.issued.Delegation.Scope.AttemptID+path, bytes.NewReader(data))
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("Authorization", "Bearer "+f.issued.Credential)
	req.Header.Set("Content-Type", "application/json")
	if requestID != "" {
		req.Header.Set("X-RadishMind-Development-Request-ID", requestID)
	}
	w := httptest.NewRecorder()
	f.handler.ServeHTTP(w, req)
	return w
}
func initialRequest(t *testing.T) NativeRequest {
	t.Helper()
	tools, e := NativeTools()
	if e != nil {
		t.Fatal(e)
	}
	return NativeRequest{Model: "fixture-model", Messages: []NativeMessage{{Role: "system", Content: "synthetic instructions"}, {Role: "user", Content: "synthetic private task body"}}, MaxCompletionTokens: 50, Tools: tools}
}
func readAction() ToolAction {
	return ToolAction{SchemaVersion: "development_tool_action.v1", Action: "read_file", Path: "main.go"}
}
func toolResult() NativeResult {
	args, _ := json.Marshal(readAction())
	return NativeResult{Message: NativeMessage{Role: "assistant", ToolCalls: []NativeCall{{"call_fixture", "function", NativeFunction{"development_tool", string(args)}}}}, Usage: &NativeUsage{20, 10, 30}, FinishReason: "tool_calls"}
}
func stopResult() NativeResult {
	return NativeResult{Message: NativeMessage{Role: "assistant", Content: "synthetic final answer"}, Usage: &NativeUsage{30, 10, 40}, FinishReason: "stop"}
}

func TestDurableNativeToolRoundTrip(t *testing.T) {
	f := newRuntimeFixture(t, func(_ context.Context, input NativeRequest) (NativeResult, error) {
		if len(input.Messages) == 2 {
			return toolResult(), nil
		}
		return stopResult(), nil
	})
	first := initialRequest(t)
	w := f.request(t, "POST", "/chat/completions", first, "model_first")
	if w.Code != 200 {
		t.Fatalf("model: %d %s", w.Code, w.Body)
	}
	var response struct {
		Choices []struct {
			Message NativeMessage `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	record := f.read(t)
	if len(record.Offers) != 1 {
		t.Fatal("offer not persisted")
	}
	var offer Offer
	for _, o := range record.Offers {
		offer = o
	}
	if offer.ParentRequestID != "model_first" || record.Control.budget.SpentTokens != 30 {
		t.Fatal(record)
	}
	call := ToolCall{"development_tool_call.v1", f.issued.Delegation.Scope.AttemptID, offer.RequestID, offer.ParentRequestID, offer.ToolCallID, readAction()}
	changed := call
	changed.Arguments.Path = "other.go"
	if w := f.request(t, "POST", "/tool-calls", changed, ""); w.Code == 202 {
		t.Fatal("changed proposal accepted")
	}
	// Inject failure after state/offer writes: every part must roll back.
	_, err := f.database.DB().Exec(`CREATE TRIGGER fail_queue BEFORE INSERT ON development_events WHEN json_extract(NEW.event_record,'$.type')='tool_offer_queued' BEGIN SELECT RAISE(ABORT,'injected failure'); END`)
	if err != nil {
		t.Fatal(err)
	}
	if w := f.request(t, "POST", "/tool-calls", call, ""); w.Code != 503 {
		t.Fatalf("rollback status: %d %s", w.Code, w.Body)
	}
	failed := f.read(t)
	if failed.Version != record.Version || failed.Offers[offer.RequestID].State != "pending" || failed.Control.budget.ToolReservations != 0 {
		t.Fatal("partial queue commit")
	}
	if _, err = f.database.DB().Exec(`DROP TRIGGER fail_queue`); err != nil {
		t.Fatal(err)
	}
	w = f.request(t, "POST", "/tool-calls", call, "")
	if w.Code != 202 || !strings.Contains(w.Body.String(), "queued") {
		t.Fatalf("queue: %d %s", w.Code, w.Body)
	}
	if w := f.request(t, "POST", "/tool-calls", call, ""); w.Code == 202 {
		t.Fatal("duplicate consumed")
	}
	if _, err = f.runtime.DispatchTool(context.Background(), f.identity, call.AttemptID, call.RequestID); err != nil {
		t.Fatal(err)
	}
	observation := "synthetic file body never persisted"
	if _, err = f.runtime.CompleteTool(context.Background(), f.identity, call.AttemptID, call.RequestID, observation); err != nil {
		t.Fatal(err)
	}
	next := first
	next.Messages = append(append([]NativeMessage(nil), first.Messages...), response.Choices[0].Message, NativeMessage{Role: "tool", Content: observation, ToolCallID: call.ToolCallID, Name: "development_tool"})
	forged := next
	forged.Messages = append([]NativeMessage(nil), next.Messages...)
	forged.Messages[len(forged.Messages)-1].Content = "forged observation"
	if w := f.request(t, "POST", "/chat/completions", forged, "model_forged"); w.Code == 200 {
		t.Fatal("forged observation accepted")
	}
	w = f.request(t, "POST", "/chat/completions", next, "model_second")
	if w.Code != 200 {
		t.Fatalf("second model: %d %s", w.Code, w.Body)
	}
	final := f.read(t)
	if final.Control.phase != "completed" || final.Control.budget.SpentTokens != 70 || !final.Revoked || f.backend.calls.Load() != 2 {
		t.Fatalf("bad final state: %+v", final)
	}
	// The on-disk records contain hashes and structured accounting, not task text.
	rows, err := f.database.DB().Query(`SELECT state_record||binding_record FROM development_attempts UNION ALL SELECT event_record FROM development_events`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var text string
		if err = rows.Scan(&text); err != nil {
			t.Fatal(err)
		}
		for _, secret := range []string{f.issued.Credential, observation, "synthetic private task body", "synthetic final answer"} {
			if strings.Contains(text, secret) {
				t.Fatalf("sensitive text persisted: %q", secret)
			}
		}
	}
	if err = rows.Err(); err != nil {
		t.Fatal(err)
	}
}

func TestDurableStoreConcurrencyRestartAndImmutability(t *testing.T) {
	f := newRuntimeFixture(t, func(context.Context, NativeRequest) (NativeResult, error) { return toolResult(), nil })
	record := f.read(t)
	// Revalidation without a change must not consume a CAS version.
	unchanged, err := f.store.change(context.Background(), f.identity, f.issued.Delegation.Scope.AttemptID, record.Version, f.runtime.instanceID, func(*Record) error { return nil })
	if err != nil || unchanged.Version != record.Version {
		t.Fatal(unchanged, err)
	}
	var winners atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := f.runtime.Cancel(context.Background(), f.identity, f.issued.Delegation.Scope.AttemptID, record.Version)
			if err == nil {
				winners.Add(1)
			} else if !errors.Is(err, ErrConflict) {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if winners.Load() != 1 {
		t.Fatal(winners.Load())
	}
	if _, err := f.database.DB().Exec(`UPDATE development_events SET event_digest='tamper'`); err == nil {
		t.Fatal("event mutation allowed")
	}
	reopened, err := NewRuntime(f.store, f.reader, f.backend)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.store.authenticate(context.Background(), reopened.instanceID, f.issued.Delegation.Scope.AttemptID, f.issued.Credential); !errors.Is(err, ErrCredential) {
		t.Fatalf("old bearer resumed: %v", err)
	}
	before := f.backend.calls.Load()
	if _, err := reopened.Read(context.Background(), f.identity, f.issued.Delegation.Scope.AttemptID); err != nil {
		t.Fatal(err)
	}
	if f.backend.calls.Load() != before {
		t.Fatal("read executed backend")
	}
	wrong := f.identity
	wrong.WorkspaceID = "workspace_other"
	if _, err := reopened.Read(context.Background(), wrong, f.issued.Delegation.Scope.AttemptID); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
}

func TestNativeChannelCancellationAndChangedQualification(t *testing.T) {
	for _, mode := range []string{"cancel", "qualification", "qualification_unavailable"} {
		t.Run(mode, func(t *testing.T) {
			started := make(chan struct{})
			f := newRuntimeFixture(t, func(ctx context.Context, _ NativeRequest) (NativeResult, error) {
				close(started)
				<-ctx.Done()
				return NativeResult{}, ctx.Err()
			})
			done := make(chan *httptest.ResponseRecorder, 1)
			go func() { done <- f.request(t, "POST", "/chat/completions", initialRequest(t), "model_wait") }()
			select {
			case <-started:
			case <-time.After(3 * time.Second):
				t.Fatal("backend not dispatched")
			}
			if mode == "cancel" {
				record := f.read(t)
				if _, err := f.runtime.Cancel(context.Background(), f.identity, f.issued.Delegation.Scope.AttemptID, record.Version); err != nil {
					t.Fatal(err)
				}
			} else {
				if mode == "qualification_unavailable" {
					f.reader.deny(ErrUnavailable)
				} else {
					f.reader.deny(ErrIneligible)
				}
			}
			select {
			case w := <-done:
				if w.Code == 200 {
					t.Fatal("stopped request succeeded")
				}
			case <-time.After(3 * time.Second):
				t.Fatal("independent cancellation not delivered")
			}
			record := f.read(t)
			if record.Control.phase != "outcome_unknown" || !record.Revoked || record.Control.budget.HeldTokens != 150 || record.Control.budget.Unresolved != 1 {
				t.Fatalf("unknown usage lost: %+v", record.Control)
			}
		})
	}
}

func TestNativeChannelMissingUsageAndMalformedResponse(t *testing.T) {
	for _, mode := range []string{"missing", "invalid", "overrun"} {
		t.Run(mode, func(t *testing.T) {
			f := newRuntimeFixture(t, func(context.Context, NativeRequest) (NativeResult, error) {
				result := toolResult()
				switch mode {
				case "missing":
					result.Usage = nil
				case "invalid":
					result.Message.ToolCalls[0].Function.Name = "shell"
				case "overrun":
					result.Usage = &NativeUsage{200, 10, 210}
				}
				return result, nil
			})
			w := f.request(t, "POST", "/chat/completions", initialRequest(t), "model_result")
			record := f.read(t)
			switch mode {
			case "missing":
				if w.Code != 200 || record.Control.budget.HeldTokens != 150 || record.Control.budget.Unresolved != 1 {
					t.Fatalf("missing became zero: %d %+v", w.Code, record.Control)
				}
			case "invalid", "overrun":
				if w.Code == 200 || record.Control.phase != "failed" || len(record.Offers) != 0 {
					t.Fatalf("invalid response accepted: %d %+v", w.Code, record.Control)
				}
			}
		})
	}
}

func TestNativeChannelInputAndCredentialFences(t *testing.T) {
	f := newRuntimeFixture(t, func(context.Context, NativeRequest) (NativeResult, error) {
		t.Fatal("unexpected dispatch")
		return NativeResult{}, nil
	})
	for _, mode := range []string{"origin", "host", "remote", "key", "cookie", "attempt", "expiry"} {
		t.Run(mode, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://127.0.0.1/internal/development/v1/attempts/"+f.issued.Delegation.Scope.AttemptID+"/tool-catalog", nil)
			req.RemoteAddr = "127.0.0.1:123"
			req.Header.Set("Authorization", "Bearer "+f.issued.Credential)
			switch mode {
			case "origin":
				req.Header.Set("Origin", "http://127.0.0.1")
			case "host":
				req.Host = "attacker.example"
			case "remote":
				req.RemoteAddr = "192.0.2.1:123"
			case "key":
				req.Header.Set("Authorization", "Bearer long-lived-key")
			case "cookie":
				req.Header.Del("Authorization")
				req.Header.Set("Cookie", "session="+f.issued.Credential)
			case "attempt":
				req.URL.Path = strings.Replace(req.URL.Path, f.issued.Delegation.Scope.AttemptID, "attempt_other", 1)
			case "expiry":
				f.runtime.now = func() time.Time { return time.UnixMilli(f.issued.Delegation.ExpiresAtMS) }
			}
			w := httptest.NewRecorder()
			f.handler.ServeHTTP(w, req)
			if w.Code != 401 {
				t.Fatalf("%s: %d %s", mode, w.Code, w.Body)
			}
		})
	}
	f.runtime.now = time.Now
	tools, _ := NativeTools()
	base, _ := json.Marshal(initialRequest(t))
	for _, invalid := range [][]byte{bytes.Replace(base, []byte(`"model"`), []byte(`"Model"`), 1), bytes.Replace(base, []byte(`"role"`), []byte(`"Role"`), 1), bytes.Replace(base, []byte(`"messages"`), []byte(`"messages":null,"unused"`), 1), []byte(`{"model":"x","model":"x"}`)} {
		if _, err := ParseNativeRequest(invalid); err == nil {
			t.Fatal("invalid JSON accepted")
		}
	}
	if len(tools) != 1 {
		t.Fatal("tool catalog missing")
	}
	// A real bounded HTTP server verifies listener interoperability and cleanup.
	server := httptest.NewServer(f.handler)
	defer server.Close()
	req, _ := http.NewRequest("GET", server.URL+"/internal/development/v1/attempts/"+f.issued.Delegation.Scope.AttemptID+"/tool-catalog", nil)
	req.Header.Set("Authorization", "Bearer "+f.issued.Credential)
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if _, err = io.Copy(io.Discard, resp.Body); err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatal(resp.Status)
	}
}

func TestCanonicalToolDistributionAndCostRounding(t *testing.T) {
	canonical, err := os.ReadFile("../../../../contracts/development-tool-action.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(canonical, actionSchema) {
		t.Fatal("embedded tool schema diverged from canonical")
	}
	sha := PayloadDigest([]byte("fixture"))
	s := Selection{sha, sha, sha, "fixture", sha, 1000000, 1, 1}
	if cost, e := s.Cost(1, 1); e != nil || cost != 0 {
		t.Fatal(cost, e)
	}
	if bound, e := s.ReserveCost(1, 1); e != nil || bound != 1 {
		t.Fatal(bound, e)
	}
}
