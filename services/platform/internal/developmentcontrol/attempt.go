package developmentcontrol

import (
	"fmt"
	"sync"
)

type Failure string

func (f Failure) Error() string { return string(f) }

const (
	DelegationInvalid        Failure = "delegation_invalid"
	BudgetExhausted          Failure = "budget_exhausted"
	OperationDuplicate       Failure = "operation_duplicate"
	OperationSequenceInvalid Failure = "operation_sequence_invalid"
	OperationStateInvalid    Failure = "operation_state_invalid"
	AttemptStopped           Failure = "attempt_stopped"
	ReceiptInvalid           Failure = "receipt_invalid"
)

type operation struct {
	request Reservation
	state   string
}

// AttemptControl is an in-memory domain rule implementation, not an execution
// service. A future repository must commit its state/event before any dispatch.
type AttemptControl struct {
	mu                sync.Mutex
	delegation        Delegation
	phase             string
	reason            string
	budget            Budget
	operations        map[string]*operation
	requestSequence   int64
	events            []Event
	lastTimeMS        int64
	terminationStatus string
}

func NewAttemptControl(d Delegation, a Authority, nowMS int64) (*AttemptControl, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}
	c := &AttemptControl{delegation: d, phase: "active", terminationStatus: "not_requested", operations: map[string]*operation{}}
	if err := c.authority(a, nowMS); err != nil {
		return nil, err
	}
	c.emit("attempt_opened", "", nowMS)
	return c, nil
}

func (c *AttemptControl) authority(a Authority, nowMS int64) error {
	d := c.delegation
	if nowMS < d.IssuedAtMS || nowMS >= d.ExpiresAtMS || a.Revoked || a.DelegationRevision != d.Revision || a.ExecutorInstanceID != d.ExecutorInstanceID || a.Scope != d.Scope {
		return DelegationInvalid
	}
	return nil
}

func (c *AttemptControl) emit(kind, requestID string, nowMS int64) Event {
	// Event sequence is authoritative; wall clock adjustments never reorder events.
	if nowMS < c.lastTimeMS {
		nowMS = c.lastTimeMS
	}
	c.lastTimeMS = nowMS
	sequence := int64(len(c.events) + 1)
	usage := "not_applicable"
	if c.budget.ModelReservations > 0 {
		usage = "settled"
	}
	if c.budget.HeldTokens > 0 || c.budget.HeldCostMicrousd > 0 {
		usage = "reserved"
	}
	for _, op := range c.operations {
		if op.request.Kind == "model" && (op.state == "uncertain" || op.state == "unmetered") {
			usage = "unknown"
			break
		}
	}
	e := Event{"development_attempt_event.v1", fmt.Sprintf("evt_%s_%d", c.delegation.Scope.AttemptID, sequence), c.delegation.Scope.AttemptID, c.delegation.DelegationID, sequence, nowMS, kind, requestID, c.phase, c.reason, usage, c.terminationStatus, c.budget}
	c.events = append(c.events, e)
	return e
}

func (c *AttemptControl) Events() []Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]Event(nil), c.events...)
}

// Reserve accepts bounds calculated by trusted Go owners, never SDK estimates.
// One reservation may be in flight at a time in this first single-agent profile.
func (c *AttemptControl) Reserve(r Reservation, a Authority, nowMS int64) (Event, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := r.Validate(); err != nil {
		return Event{}, err
	}
	d := c.delegation
	if r.DelegationID != d.DelegationID || r.DelegationRevision != d.Revision || r.AttemptID != d.Scope.AttemptID {
		return Event{}, DelegationInvalid
	}
	if err := c.authority(a, nowMS); err != nil {
		return Event{}, err
	}
	if c.phase != "active" {
		return Event{}, AttemptStopped
	}
	if _, exists := c.operations[r.RequestID]; exists {
		return Event{}, OperationDuplicate
	}
	if r.Sequence != c.requestSequence+1 {
		return Event{}, OperationSequenceInvalid
	}
	b, l := c.budget, d.Limits
	if b.InFlight != 0 || b.SpentTokens+b.HeldTokens+r.TokenBound > l.MaxTokens || b.SpentCostMicrousd+b.HeldCostMicrousd+r.CostBoundMicrousd > l.MaxCostMicrousd {
		return Event{}, BudgetExhausted
	}
	if r.Kind == "model" {
		if b.ModelReservations >= l.MaxModelCalls {
			return Event{}, BudgetExhausted
		}
		c.budget.ModelReservations++
	} else {
		if b.ToolReservations >= l.MaxToolCalls {
			return Event{}, BudgetExhausted
		}
		c.budget.ToolReservations++
	}
	c.budget.HeldTokens += r.TokenBound
	c.budget.HeldCostMicrousd += r.CostBoundMicrousd
	c.budget.InFlight++
	c.requestSequence = r.Sequence
	c.operations[r.RequestID] = &operation{r, "reserved"}
	return c.emit("reservation_created", r.RequestID, nowMS), nil
}

func (c *AttemptControl) RecordDispatch(requestID string, a Authority, nowMS int64) (Event, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.authority(a, nowMS); err != nil {
		return Event{}, err
	}
	if c.phase != "active" {
		return Event{}, AttemptStopped
	}
	op, ok := c.operations[requestID]
	if !ok || op.state != "reserved" {
		return Event{}, OperationStateInvalid
	}
	op.state = "dispatched"
	return c.emit("dispatch_recorded", requestID, nowMS), nil
}

// RefreshAuthority is the owner control loop's expiry/revocation check. It must
// also run while an SDK is blocked, independently of new model/tool requests.
func (c *AttemptControl) RefreshAuthority(a Authority, nowMS int64) (*Event, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.phase != "active" {
		return nil, AttemptStopped
	}
	if c.authority(a, nowMS) == nil {
		return nil, nil
	}
	reason := "authority_changed"
	if nowMS >= c.delegation.ExpiresAtMS {
		reason = "deadline_exceeded"
	}
	event := c.cancel(reason, nowMS)
	return &event, nil
}

func (c *AttemptControl) RequestCancel(reason string, nowMS int64) (Event, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if reason != "user_cancelled" && reason != "authority_changed" && reason != "deadline_exceeded" {
		return Event{}, OperationStateInvalid
	}
	if c.phase != "active" {
		return Event{}, AttemptStopped
	}
	return c.cancel(reason, nowMS), nil
}

func (c *AttemptControl) cancel(reason string, nowMS int64) Event {
	c.phase = "cancel_requested"
	c.reason = reason
	c.terminationStatus = "pending"
	c.releaseUndispatched()
	return c.emit("cancel_requested", "", nowMS)
}

func (c *AttemptControl) releaseUndispatched() {
	for _, op := range c.operations {
		if op.state == "reserved" {
			c.budget.HeldTokens -= op.request.TokenBound
			c.budget.HeldCostMicrousd -= op.request.CostBoundMicrousd
			c.budget.InFlight--
			op.state = "not_dispatched"
		}
	}
}

// TerminationEvidence is supplied only by the trusted process/container owner.
// SDK pause events, client assertions and HTTP disconnects are insufficient.
type TerminationEvidence struct {
	AttemptID             string
	ExecutorInstanceID    string
	ExecutorExited        bool
	SandboxExited         bool
	ModelConnectionClosed bool
}

func (c *AttemptControl) ConfirmTermination(proof TerminationEvidence, nowMS int64) (Event, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if (c.phase != "cancel_requested" && c.phase != "outcome_unknown" && c.phase != "failed") || c.terminationStatus == "confirmed" {
		return Event{}, AttemptStopped
	}
	if proof.AttemptID != c.delegation.Scope.AttemptID || proof.ExecutorInstanceID != c.delegation.ExecutorInstanceID {
		return Event{}, DelegationInvalid
	}
	complete := proof.ExecutorExited && proof.SandboxExited && proof.ModelConnectionClosed
	if c.terminationStatus == "unconfirmed" && !complete {
		return Event{}, OperationStateInvalid
	}
	c.terminationStatus = "confirmed"
	if !complete {
		c.terminationStatus = "unconfirmed"
	}
	c.releaseUndispatched()
	if c.phase == "cancel_requested" {
		c.phase = "cancelled"
	}
	if !proof.ExecutorExited || !proof.SandboxExited || !proof.ModelConnectionClosed {
		c.phase = "outcome_unknown"
		c.reason = "termination_unconfirmed"
	}
	for _, op := range c.operations {
		if op.state == "dispatched" {
			op.state = "uncertain"
			c.budget.Unresolved++
			c.budget.InFlight--
		}
	}
	return c.emit("termination_recorded", "", nowMS), nil
}

// Complete closes control work, not task verification or user acceptance.
func (c *AttemptControl) Complete(nowMS int64) (Event, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.phase != "active" {
		return Event{}, AttemptStopped
	}
	if c.budget.InFlight != 0 {
		return Event{}, OperationStateInvalid
	}
	c.phase = "completed"
	return c.emit("attempt_completed", "", nowMS), nil
}
