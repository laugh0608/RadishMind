package developmentcontrol

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/sqlitedev"
	migrations "radishmind.local/services/platform/migrations/sqlite/development_attempts"
)

func TestDurableDatabaseReopenAndExplicitRecovery(t *testing.T) {
	f := newRuntimeFixture(t, func(context.Context, NativeRequest) (NativeResult, error) { return toolResult(), nil })
	if w := f.request(t, "POST", "/chat/completions", initialRequest(t), "model_first"); w.Code != 200 {
		t.Fatal(w.Body)
	}
	before := f.read(t)
	var path string
	rows, err := f.database.DB().Query(`PRAGMA database_list`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var seq int
		var name, file string
		if err = rows.Scan(&seq, &name, &file); err != nil {
			t.Fatal(err)
		}
		if name == "main" {
			path = file
		}
	}
	if err = rows.Err(); err != nil {
		t.Fatal(err)
	}
	rows.Close()
	// Simulate abrupt ownership loss: close the database without a Runtime close.
	if err = f.database.Close(); err != nil {
		t.Fatal(err)
	}
	db, err := sqlitedev.Open(context.Background(), sqlitedev.Options{DatabasePath: path, Migrations: migrations.Migrations()})
	if err != nil {
		t.Fatal(err)
	}
	f.database = db
	f.store, err = NewSQLiteStore(db.DB())
	if err != nil {
		t.Fatal(err)
	}
	restarted, err := NewRuntime(f.store, f.reader, f.backend)
	if err != nil {
		t.Fatal(err)
	}
	// Replace fixture cleanup owner; new owner must not adopt the old credential.
	f.runtime = restarted
	recovered, err := restarted.Read(context.Background(), f.identity, f.issued.Delegation.Scope.AttemptID)
	if err != nil {
		t.Fatal(err)
	}
	if recovered.Version != before.Version || len(recovered.Control.Events()) != len(before.Control.Events()) || f.backend.calls.Load() != 1 {
		t.Fatal("read resumed/rebuilt the attempt")
	}
	if _, err = f.store.authenticate(context.Background(), restarted.instanceID, f.issued.Delegation.Scope.AttemptID, f.issued.Credential); !errors.Is(err, ErrCredential) {
		t.Fatal("restarted credential accepted", err)
	}
	stopped, err := restarted.Interrupt(context.Background(), f.identity, f.issued.Delegation.Scope.AttemptID, recovered.Version)
	if err != nil {
		t.Fatal(err)
	}
	if !stopped.Revoked || stopped.Control.phase != "outcome_unknown" || stopped.Control.reason != "runtime_interrupted" {
		t.Fatal("recovery asserted successful cancellation")
	}
}

func TestDurableQueueSingleConsumerAndEventPagination(t *testing.T) {
	f := newRuntimeFixture(t, func(context.Context, NativeRequest) (NativeResult, error) { return toolResult(), nil })
	if w := f.request(t, "POST", "/chat/completions", initialRequest(t), "model_first"); w.Code != 200 {
		t.Fatal(w.Body)
	}
	record := f.read(t)
	var offer Offer
	for _, o := range record.Offers {
		offer = o
	}
	call := ToolCall{"development_tool_call.v1", f.issued.Delegation.Scope.AttemptID, offer.RequestID, offer.ParentRequestID, offer.ToolCallID, readAction()}
	var winners atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := f.runtime.QueueTool(context.Background(), record, call); err == nil {
				winners.Add(1)
			} else if !errors.Is(err, ErrContext) {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if winners.Load() != 1 {
		t.Fatal(winners.Load())
	}
	queued := f.read(t)
	if queued.Control.budget.ToolReservations != 1 {
		t.Fatal("double reservation")
	}
	if w := f.request(t, "GET", "/tool-offers/"+offer.ToolCallID, nil, ""); w.Code != 200 {
		t.Fatal(w.Code, w.Body)
	}
	if w := f.request(t, "GET", "/events?after_sequence=0&limit=2", nil, ""); w.Code != 200 {
		t.Fatal(w.Code, w.Body)
	} else {
		var page struct {
			Events []Event `json:"events"`
			Next   int     `json:"next_sequence"`
			More   bool    `json:"has_more"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil {
			t.Fatal(err)
		}
		if len(page.Events) != 2 || page.Next != 2 || !page.More {
			t.Fatal(page)
		}
	}
	for _, query := range []string{"?after_sequence=999", "?limit=51", "?limit=1&limit=2", "?extra=1"} {
		if w := f.request(t, "GET", "/events"+query, nil, ""); w.Code != 400 {
			t.Fatal(query, w.Code)
		}
	}
	// Request identity and offer identity both remain immutable on disk.
	if _, err := f.database.DB().Exec(`UPDATE development_requests SET payload_digest='tampered'`); err == nil {
		t.Fatal("request identity mutable")
	}
	if _, err := f.database.DB().Exec(`UPDATE development_tool_offers SET parent_request_id='other'`); err == nil {
		t.Fatal("offer identity mutable")
	}
}

func TestDurableFailureBeforeDispatchAndCorruptState(t *testing.T) {
	f := newRuntimeFixture(t, func(context.Context, NativeRequest) (NativeResult, error) {
		t.Error("dispatch after failed commit")
		return stopResult(), nil
	})
	if _, err := f.database.DB().Exec(`CREATE TRIGGER fail_dispatch BEFORE INSERT ON development_events WHEN json_extract(NEW.event_record,'$.type')='dispatch_recorded' BEGIN SELECT RAISE(ABORT,'injected failure'); END`); err != nil {
		t.Fatal(err)
	}
	before := f.read(t)
	if w := f.request(t, "POST", "/chat/completions", initialRequest(t), "model_first"); w.Code != 503 {
		t.Fatal(w.Code, w.Body)
	}
	after := f.read(t)
	if after.Version != before.Version || after.Control.budget.ModelReservations != 0 || f.backend.calls.Load() != 0 {
		t.Fatal("failed commit dispatched or partially charged")
	}
	if _, err := f.database.DB().Exec(`UPDATE development_attempts SET state_digest='corrupted'`); err != nil {
		t.Fatal(err)
	}
	if _, err := f.runtime.Read(context.Background(), f.identity, f.issued.Delegation.Scope.AttemptID); !errors.Is(err, ErrUnavailable) {
		t.Fatal("corruption repaired silently", err)
	}
	// Restore only the checksum of the untouched canonical record for cleanup.
	var state string
	if err := f.database.DB().QueryRow(`SELECT state_record FROM development_attempts`).Scan(&state); err != nil {
		t.Fatal(err)
	}
	if _, err := f.database.DB().Exec(`UPDATE development_attempts SET state_digest=?`, PayloadDigest([]byte(state))); err != nil {
		t.Fatal(err)
	}
}

func TestInitialTaskAndConfirmationAreFrozen(t *testing.T) {
	f := newRuntimeFixture(t, func(context.Context, NativeRequest) (NativeResult, error) {
		t.Fatal("unconfirmed task dispatched")
		return NativeResult{}, nil
	})
	changed := initialRequest(t)
	changed.Messages[1].Content = "replacement task outside the approved input"
	if w := f.request(t, "POST", "/chat/completions", changed, "model_replaced"); w.Code != 400 {
		t.Fatal(w.Code, w.Body)
	}
	if f.read(t).Control.budget.ModelReservations != 0 {
		t.Fatal("rejected task charged")
	}
	prepared, err := f.runtime.Prepare(context.Background(), f.identity, f.plan)
	if err != nil {
		t.Fatal(err)
	}
	f.reader.mu.Lock()
	f.reader.selection.PricingDigest = PayloadDigest([]byte("changed-price"))
	f.reader.mu.Unlock()
	if _, err := f.runtime.Issue(context.Background(), f.identity, f.plan, "confirmation_changed", prepared.PlanDigest, prepared.Selection); !errors.Is(err, ErrConflict) {
		t.Fatal("stale confirmation accepted", err)
	}
}

func TestPlanFixtureStrictnessAndIssuedSecretHandoff(t *testing.T) {
	data, err := os.ReadFile("../../../../scripts/eval/fixtures/development-control/plan.json")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := ParsePlan(data)
	if err != nil {
		t.Fatal(err)
	}
	plan.Tools.WritePaths = []string{"outside.go"}
	if plan.Validate() == nil {
		t.Fatal("write policy expanded")
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	for name := range fields {
		copyFields := map[string]json.RawMessage{}
		for k, v := range fields {
			copyFields[k] = v
		}
		delete(copyFields, name)
		invalid, _ := json.Marshal(copyFields)
		if _, err := ParsePlan(invalid); err == nil {
			t.Fatal("missing plan field accepted", name)
		}
	}
	for _, entry := range []string{`{"read_paths":["main.go"],"write_paths":[],"Verification_profile":"go_test_v1"}`, `{"read_paths":["main.go"],"write_paths":[],"verification_profile":"shell"}`} {
		fields["tools"] = json.RawMessage(entry)
		invalid, _ := json.Marshal(fields)
		if _, err := ParsePlan(invalid); err == nil {
			t.Fatal("invalid tool policy accepted")
		}
	}
	f := newRuntimeFixture(t, func(context.Context, NativeRequest) (NativeResult, error) { return stopResult(), nil })
	encoded, err := json.Marshal(f.issued)
	if err != nil {
		t.Fatal(err)
	}
	var safe map[string]json.RawMessage
	if err = json.Unmarshal(encoded, &safe); err != nil {
		t.Fatal(err)
	}
	if _, leaked := safe["credential"]; leaked {
		t.Fatal("bearer automatically serialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err = f.runtime.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err = f.runtime.Issue(ctx, f.identity, f.plan, "confirmation_afterclose", "", Selection{}); !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
}

func TestEventOnlyTransactionPersistsAtTheSameMillisecond(t *testing.T) {
	f := newRuntimeFixture(t, func(context.Context, NativeRequest) (NativeResult, error) { return stopResult(), nil })
	f.runtime.now = func() time.Time { return time.UnixMilli(f.issued.Delegation.IssuedAtMS) }
	before := f.read(t)
	for i := 0; i < 2; i++ {
		if _, err := f.runtime.workerEvent(context.Background(), f.identity, f.issued.Delegation.Scope.AttemptID, "sdk_system_prompt_observed", ""); err != nil {
			t.Fatal(err)
		}
	}
	after := f.read(t)
	if len(after.Control.Events()) != len(before.Control.Events())+2 || after.Version != before.Version+2 {
		t.Fatal("same-time events were discarded as a no-op")
	}
}

func TestOversizedEventStreamCannotCommitUnreadableState(t *testing.T) {
	f := newRuntimeFixture(t, func(context.Context, NativeRequest) (NativeResult, error) { return stopResult(), nil })
	before := f.read(t)
	if _, err := f.store.change(context.Background(), f.identity, f.issued.Delegation.Scope.AttemptID, 0, f.runtime.instanceID, func(record *Record) error {
		for i := 0; i < 200; i++ {
			record.Control.emit("sdk_user_message_observed", "", f.runtime.now().UnixMilli())
		}
		return nil
	}); !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	after := f.read(t)
	if before.Version != after.Version || len(after.Control.Events()) != len(before.Control.Events()) {
		t.Fatal("oversized transaction leaked events")
	}
}
