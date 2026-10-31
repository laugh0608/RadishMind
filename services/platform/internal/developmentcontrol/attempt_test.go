package developmentcontrol

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("../../../../scripts/eval/fixtures/development-control", name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func setup(t *testing.T) (*AttemptControl, Delegation, Authority, Reservation) {
	t.Helper()
	d, err := ParseDelegation(fixture(t, "delegation"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := ParseReservation(fixture(t, "reservation"))
	if err != nil {
		t.Fatal(err)
	}
	a := Authority{Scope: d.Scope, DelegationRevision: d.Revision, ExecutorInstanceID: d.ExecutorInstanceID}
	c, err := NewAttemptControl(d, a, 1000)
	if err != nil {
		t.Fatal(err)
	}
	return c, d, a, r
}
func requireEvent(t *testing.T, event Event, err error) Event {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	return event
}
func reserve(t *testing.T, c *AttemptControl, r Reservation, a Authority) {
	t.Helper()
	_, err := c.Reserve(r, a, 1001)
	if err != nil {
		t.Fatal(err)
	}
}
func dispatch(t *testing.T, c *AttemptControl, r Reservation, a Authority) {
	t.Helper()
	_, err := c.RecordDispatch(r.RequestID, a, 1002)
	if err != nil {
		t.Fatal(err)
	}
}
func receipt(r Reservation, status string, tokens, cost int64) Receipt {
	return Receipt{"development_attempt_receipt.v1", r.AttemptID, r.RequestID, status, tokens, cost}
}
func proof(d Delegation) TerminationEvidence {
	return TerminationEvidence{d.Scope.AttemptID, d.ExecutorInstanceID, true, true, true}
}
func latest(c *AttemptControl) Event { events := c.Events(); return events[len(events)-1] }

func TestCanonicalCancellationTrace(t *testing.T) {
	c, d, a, r := setup(t)
	reserve(t, c, r, a)
	dispatch(t, c, r, a)
	_, err := c.RequestCancel("user_cancelled", 1003)
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.ConfirmTermination(proof(d), 1004)
	if err != nil {
		t.Fatal(err)
	}
	p, err := ParseReceipt(fixture(t, "receipt"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.RecordReceipt(p, 1005)
	if err != nil {
		t.Fatal(err)
	}
	var expected []Event
	if err := json.Unmarshal(fixture(t, "cancel-trace"), &expected); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c.Events(), expected) {
		data, _ := json.MarshalIndent(c.Events(), "", "  ")
		t.Fatalf("trace changed:\n%s", data)
	}
	if _, err := c.RecordReceipt(p, 1006); err != OperationStateInvalid {
		t.Fatalf("duplicate receipt: %v", err)
	}
	if _, err := c.Reserve(r, a, 1006); err != AttemptStopped {
		t.Fatalf("terminal dispatch: %v", err)
	}
	copy := c.Events()
	copy[0].Budget.SpentTokens = 900
	if c.Events()[0].Budget.SpentTokens != 0 {
		t.Fatal("event snapshot is mutable")
	}
}

func TestConcurrentReservationCannotOverspend(t *testing.T) {
	c, _, a, r := setup(t)
	var accepted atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			request := r
			request.RequestID = fmt.Sprintf("req_concurrent_%d", i)
			if _, err := c.Reserve(request, a, 1001); err == nil {
				accepted.Add(1)
			} else if err != BudgetExhausted && err != OperationSequenceInvalid {
				t.Errorf("unexpected failure: %v", err)
			}
		}(i)
	}
	wg.Wait()
	e := latest(c)
	if accepted.Load() != 1 || e.Budget.InFlight != 1 || e.Budget.HeldTokens != 100 || e.Sequence != 2 {
		t.Fatalf("concurrent overspend: accepted=%d event=%+v", accepted.Load(), e)
	}
}

func TestMissingUsageRetainsReservationAndReconcilesOnce(t *testing.T) {
	c, _, a, r := setup(t)
	reserve(t, c, r, a)
	dispatch(t, c, r, a)
	e, err := c.RecordReceipt(receipt(r, "missing", 0, 0), 1003)
	requireEvent(t, e, err)
	if e.UsageStatus != "unknown" || e.Budget.HeldTokens != 100 || e.Budget.SpentTokens != 0 || e.Budget.Unresolved != 1 {
		t.Fatalf("missing usage became zero cost: %+v", e)
	}
	next := r
	next.RequestID = "req_model_002"
	next.Sequence = 2
	next.CostBoundMicrousd = 1001
	before := c.Events()
	if _, err := c.Reserve(next, a, 1004); err != BudgetExhausted {
		t.Fatalf("held cost was released: %v", err)
	}
	if !reflect.DeepEqual(before, c.Events()) {
		t.Fatal("rejected reservation changed state")
	}
	next.CostBoundMicrousd = 1000
	reserve(t, c, next, a)
	dispatch(t, c, next, a)
	e, err = c.RecordReceipt(receipt(next, "reported", 100, 1000), 1005)
	requireEvent(t, e, err)
	e, err = c.RecordReceipt(receipt(r, "reported", 50, 500), 1006)
	requireEvent(t, e, err)
	if e.Budget.HeldTokens != 0 || e.Budget.Unresolved != 0 || e.Budget.SpentTokens != 150 || e.Budget.SpentCostMicrousd != 1500 {
		t.Fatalf("incorrect reconciliation: %+v", e)
	}
	if _, err := c.Reserve(r, a, 1007); err != OperationDuplicate {
		t.Fatalf("request replay accepted: %v", err)
	}
}

func TestReleaseOnlyBeforeDispatchAndPreserveCallQuota(t *testing.T) {
	c, _, a, r := setup(t)
	reserve(t, c, r, a)
	e, err := c.RecordReceipt(receipt(r, "not_dispatched", 0, 0), 1002)
	requireEvent(t, e, err)
	if e.Budget.HeldTokens != 0 || e.Budget.ModelReservations != 1 || e.Budget.InFlight != 0 {
		t.Fatal("incorrect release")
	}
	if _, err := c.RecordDispatch(r.RequestID, a, 1003); err != OperationStateInvalid {
		t.Fatal("released operation dispatched")
	}
	next := r
	next.RequestID = "req_model_002"
	next.Sequence = 2
	reserve(t, c, next, a)
	dispatch(t, c, next, a)
	if _, err := c.RecordReceipt(receipt(next, "not_dispatched", 0, 0), 1004); err != OperationStateInvalid {
		t.Fatal("dispatched reservation refunded without outcome")
	}
}

func TestAuthorityAndExpiryAreCheckedAgain(t *testing.T) {
	for _, name := range []string{"workspace", "binding", "application_authority", "model", "revision", "executor", "revoked", "expired", "future"} {
		t.Run(name, func(t *testing.T) {
			c, d, a, r := setup(t)
			reserve(t, c, r, a)
			now := int64(1002)
			switch name {
			case "workspace":
				a.Scope.WorkspaceID = "workspace_other"
			case "binding":
				a.Scope.BindingRevision++
			case "application_authority":
				a.Scope.ApplicationAuthorityDigest = "sha256:" + strings.Repeat("b", 64)
			case "model":
				a.Scope.ModelRef = "model_other"
			case "revision":
				a.DelegationRevision++
			case "executor":
				a.ExecutorInstanceID = "executor_other"
			case "revoked":
				a.Revoked = true
			case "expired":
				now = d.ExpiresAtMS
			case "future":
				now = 999
			}
			if _, err := c.RecordDispatch(r.RequestID, a, now); err != DelegationInvalid {
				t.Fatalf("stale authority accepted: %v", err)
			}
			e, err := c.RefreshAuthority(a, now)
			if err != nil || e == nil || e.Phase != "cancel_requested" || e.Budget.HeldTokens != 0 {
				t.Fatalf("authority refresh did not fence work: %+v %v", e, err)
			}
		})
	}
	c, _, a, _ := setup(t)
	event, err := c.RefreshAuthority(a, 1001)
	if err != nil || event != nil {
		t.Fatal("unchanged authority emitted a fake event")
	}
}

func TestCancelDispatchRaceHasOnlyTwoConservativeOutcomes(t *testing.T) {
	for i := 0; i < 50; i++ {
		c, d, a, r := setup(t)
		reserve(t, c, r, a)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, err := c.RecordDispatch(r.RequestID, a, 1002)
			if err != nil && err != AttemptStopped {
				t.Error(err)
			}
		}()
		go func() {
			defer wg.Done()
			_, err := c.RequestCancel("user_cancelled", 1002)
			if err != nil {
				t.Error(err)
			}
		}()
		wg.Wait()
		e := latest(c)
		if e.Phase != "cancel_requested" || (e.Budget.HeldTokens != 0 && e.Budget.HeldTokens != 100) {
			t.Fatalf("invalid race state: %+v", e)
		}
		e, err := c.ConfirmTermination(proof(d), 1003)
		requireEvent(t, e, err)
		if e.Phase != "cancelled" || e.Budget.InFlight != 0 {
			t.Fatal("local termination not accounted")
		}
	}
}

func TestUnknownTerminationAndLateUsageNeverReactivate(t *testing.T) {
	c, d, a, r := setup(t)
	reserve(t, c, r, a)
	dispatch(t, c, r, a)
	if _, err := c.RequestCancel("user_cancelled", 1003); err != nil {
		t.Fatal(err)
	}
	bad := proof(d)
	bad.ExecutorInstanceID = "exec_other"
	if _, err := c.ConfirmTermination(bad, 1004); err != DelegationInvalid {
		t.Fatal("cross executor termination accepted")
	}
	unknown := proof(d)
	unknown.SandboxExited = false
	e, err := c.ConfirmTermination(unknown, 1004)
	requireEvent(t, e, err)
	if e.Phase != "outcome_unknown" || e.Budget.HeldTokens != 100 {
		t.Fatal("unconfirmed termination called cancelled")
	}
	e, err = c.RecordReceipt(receipt(r, "reported", 250, 3000), 1005)
	requireEvent(t, e, err)
	if e.Phase != "outcome_unknown" || e.Budget.SpentTokens != 250 || e.Budget.SpentCostMicrousd != 3000 || e.Budget.HeldTokens != 0 {
		t.Fatal("late actual overrun was hidden or reactivated work")
	}
	if _, err := c.Complete(1006); err != AttemptStopped {
		t.Fatal("unknown attempt became complete")
	}
	e, err = c.ConfirmTermination(proof(d), 1006)
	requireEvent(t, e, err)
	if e.Phase != "outcome_unknown" || e.TerminationStatus != "confirmed" {
		t.Fatal("late cleanup confirmation changed execution outcome")
	}
	if _, err := c.ConfirmTermination(proof(d), 1007); err != AttemptStopped {
		t.Fatal("termination proof replay accepted")
	}
}

func TestReceiptKindsAndOverrun(t *testing.T) {
	c, _, a, r := setup(t)
	reserve(t, c, r, a)
	dispatch(t, c, r, a)
	if _, err := c.RecordReceipt(receipt(r, "not_applicable", 0, 0), 1003); err != ReceiptInvalid {
		t.Fatal("model receipt treated as tool")
	}
	e, err := c.RecordReceipt(receipt(r, "reported", 201, 2001), 1004)
	requireEvent(t, e, err)
	if e.Phase != "failed" || e.Reason != "reservation_exceeded" || e.Budget.SpentTokens != 201 {
		t.Fatal("actual usage clipped")
	}
	c, _, a, r = setup(t)
	r.Kind = "tool"
	r.TokenBound = 0
	r.CostBoundMicrousd = 0
	reserve(t, c, r, a)
	dispatch(t, c, r, a)
	if _, err := c.RecordReceipt(receipt(r, "reported", 0, 0), 1003); err != ReceiptInvalid {
		t.Fatal("tool claimed model accounting")
	}
	e, err = c.RecordReceipt(receipt(r, "not_applicable", 0, 0), 1004)
	requireEvent(t, e, err)
	if e.Budget.ToolReservations != 1 || e.Budget.InFlight != 0 {
		t.Fatal("tool capacity not settled")
	}
	e, err = c.Complete(1005)
	requireEvent(t, e, err)
	if e.Phase != "completed" {
		t.Fatal("control closure failed")
	}
}

func TestStrictWireFieldsAndBinding(t *testing.T) {
	for _, name := range []string{"delegation", "reservation", "receipt"} {
		t.Run(name, func(t *testing.T) {
			original := fixture(t, name)
			parse := func(data []byte) error {
				switch name {
				case "delegation":
					_, err := ParseDelegation(data)
					return err
				case "reservation":
					_, err := ParseReservation(data)
					return err
				default:
					_, err := ParseReceipt(data)
					return err
				}
			}
			text := strings.TrimSpace(string(original))
			for _, bad := range []string{text + `{}`, text[:len(text)-1] + `,"extra":true}`, strings.Replace(text, `"schema_version":`, `"Schema_Version":`, 1), strings.Replace(text, `"schema_version":`, `"schema_version":null,"schema_version":`, 1)} {
				if err := parse([]byte(bad)); err == nil {
					t.Fatal("noncanonical wire accepted")
				}
			}
		})
	}
	c, _, a, r := setup(t)
	r.AttemptID = "attempt_other"
	if _, err := c.Reserve(r, a, 1001); err != DelegationInvalid {
		t.Fatal("cross attempt reservation accepted")
	}
	r.AttemptID = a.Scope.AttemptID
	r.Sequence = 2
	if _, err := c.Reserve(r, a, 1001); err != OperationSequenceInvalid {
		t.Fatal("sequence gap accepted")
	}
}

func TestUnknownExecutionCanRecordCleanupWithoutBecomingCancelled(t *testing.T) {
	c, d, a, r := setup(t)
	reserve(t, c, r, a)
	dispatch(t, c, r, a)
	e, err := c.RecordReceipt(receipt(r, "unknown", 0, 0), 1003)
	requireEvent(t, e, err)
	if e.Phase != "outcome_unknown" || e.Budget.HeldCostMicrousd != 1000 || e.Budget.InFlight != 0 || e.Budget.Unresolved != 1 {
		t.Fatal("uncertain effect was refunded")
	}
	e, err = c.ConfirmTermination(proof(d), 1004)
	requireEvent(t, e, err)
	if e.Phase != "outcome_unknown" {
		t.Fatal("cleanup rewrote an unknown outcome")
	}
	e, err = c.RecordReceipt(receipt(r, "reported", 0, 0), 999)
	requireEvent(t, e, err)
	if e.Phase != "outcome_unknown" || e.Budget.HeldCostMicrousd != 0 || e.Budget.Unresolved != 0 || e.OccurredAtMS != 1004 {
		t.Fatal("late explicit zero or clock rollback corrupted accounting")
	}
}

func TestLateOverrunStillAllowsOwnedCleanupOfNewWork(t *testing.T) {
	for _, send := range []bool{false, true} {
		t.Run(fmt.Sprint(send), func(t *testing.T) {
			c, d, a, r := setup(t)
			reserve(t, c, r, a)
			dispatch(t, c, r, a)
			if _, err := c.RecordReceipt(receipt(r, "missing", 0, 0), 1003); err != nil {
				t.Fatal(err)
			}
			next := r
			next.RequestID = "req_model_002"
			next.Sequence = 2
			reserve(t, c, next, a)
			if send {
				dispatch(t, c, next, a)
			}
			e, err := c.RecordReceipt(receipt(r, "reported", 201, 2001), 1004)
			requireEvent(t, e, err)
			if e.Phase != "failed" {
				t.Fatal("late overrun did not fence newer work")
			}
			e, err = c.ConfirmTermination(proof(d), 1005)
			requireEvent(t, e, err)
			if e.Phase != "failed" || e.Budget.InFlight != 0 {
				t.Fatal("failed attempt could not finish owned cleanup")
			}
			if send && (e.Budget.Unresolved != 1 || e.Budget.HeldTokens != 100) {
				t.Fatal("dispatched work refunded")
			}
			if !send && (e.Budget.Unresolved != 0 || e.Budget.HeldTokens != 0) {
				t.Fatal("never-dispatched work not released")
			}
		})
	}
}

func TestUnknownToolEffectDoesNotInventUnknownModelUsage(t *testing.T) {
	c, d, a, r := setup(t)
	r.Kind = "tool"
	r.TokenBound = 0
	r.CostBoundMicrousd = 0
	reserve(t, c, r, a)
	dispatch(t, c, r, a)
	e, err := c.RecordReceipt(receipt(r, "unknown", 0, 0), 1003)
	requireEvent(t, e, err)
	if e.Phase != "outcome_unknown" || e.UsageStatus != "not_applicable" || e.Budget.Unresolved != 1 {
		t.Fatalf("tool effect and model usage conflated: %+v", e)
	}
	e, err = c.ConfirmTermination(proof(d), 1004)
	requireEvent(t, e, err)
	if e.UsageStatus != "not_applicable" {
		t.Fatal("cleanup invented model usage")
	}
}
