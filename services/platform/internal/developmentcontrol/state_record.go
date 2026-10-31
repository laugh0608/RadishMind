package developmentcontrol

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
	"sort"

	"radishmind.local/services/platform/internal/strictjson"
)

type operationRecord struct {
	Request Reservation `json:"request"`
	State   string      `json:"state"`
	Receipt Receipt     `json:"receipt"`
}

type stateRecord struct {
	SchemaVersion     string            `json:"schema_version"`
	Delegation        Delegation        `json:"delegation"`
	Phase             string            `json:"phase"`
	Reason            string            `json:"reason"`
	TerminationStatus string            `json:"termination_status"`
	Budget            Budget            `json:"budget"`
	RequestSequence   int64             `json:"request_sequence"`
	LastTimeMS        int64             `json:"last_time_ms"`
	Operations        []operationRecord `json:"operations"`
}

func encodeState(c *AttemptControl) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	operations := make([]operationRecord, 0, len(c.operations))
	for _, op := range c.operations {
		operations = append(operations, operationRecord{op.request, op.state, op.receipt})
	}
	sort.Slice(operations, func(i, j int) bool { return operations[i].Request.Sequence < operations[j].Request.Sequence })
	return json.Marshal(stateRecord{"development_attempt_state.v1", c.delegation, c.phase, c.reason, c.terminationStatus, c.budget, c.requestSequence, c.lastTimeMS, operations})
}

// decodeState restores data only. It never calls NewAttemptControl, a backend,
// tool dispatcher or any of the state-changing domain methods.
func decodeState(data []byte, events []Event) (*AttemptControl, error) {
	var state stateRecord
	if err := strictjson.Decode(data, &state); err != nil {
		return nil, err
	}
	if state.SchemaVersion != "development_attempt_state.v1" || state.Delegation.Validate() != nil || len(events) < 1 || len(events) > 200 || len(state.Operations) > 40 || state.RequestSequence != int64(len(state.Operations)) {
		return nil, errors.New("invalid persisted development state")
	}
	c := &AttemptControl{delegation: state.Delegation, phase: state.Phase, reason: state.Reason, terminationStatus: state.TerminationStatus, budget: state.Budget, requestSequence: state.RequestSequence, lastTimeMS: state.LastTimeMS, operations: map[string]*operation{}, events: append([]Event(nil), events...)}
	var budget Budget
	for i, op := range state.Operations {
		r := op.Request
		if r.Validate() != nil || r.Sequence != int64(i+1) || r.AttemptID != c.delegation.Scope.AttemptID || r.DelegationID != c.delegation.DelegationID || r.DelegationRevision != c.delegation.Revision || c.operations[r.RequestID] != nil {
			return nil, errors.New("invalid persisted operation")
		}
		if r.Kind == "model" {
			budget.ModelReservations++
		} else {
			budget.ToolReservations++
		}
		empty := op.Receipt == (Receipt{})
		if !empty && (op.Receipt.Validate() != nil || op.Receipt.RequestID != r.RequestID || op.Receipt.AttemptID != r.AttemptID) {
			return nil, errors.New("invalid persisted receipt")
		}
		switch op.State {
		case "reserved", "dispatched":
			if !empty {
				return nil, errors.New("unfinished operation has a receipt")
			}
			budget.InFlight++
			budget.HeldTokens += r.TokenBound
			budget.HeldCostMicrousd += r.CostBoundMicrousd
		case "unmetered", "uncertain":
			if op.State == "unmetered" && (r.Kind != "model" || empty || op.Receipt.Status != "missing") {
				return nil, errors.New("invalid missing usage state")
			}
			if op.State == "uncertain" && !empty && op.Receipt.Status != "unknown" {
				return nil, errors.New("invalid uncertain state")
			}
			budget.Unresolved++
			budget.HeldTokens += r.TokenBound
			budget.HeldCostMicrousd += r.CostBoundMicrousd
		case "settled":
			if empty || (r.Kind == "model" && op.Receipt.Status != "reported") || (r.Kind == "tool" && op.Receipt.Status != "not_applicable") {
				return nil, errors.New("invalid settled operation")
			}
			budget.SpentTokens += op.Receipt.Tokens
			budget.SpentCostMicrousd += op.Receipt.CostMicrousd
		case "not_dispatched":
			if !empty && op.Receipt.Status != "not_dispatched" {
				return nil, errors.New("invalid released operation")
			}
		default:
			return nil, errors.New("unknown persisted operation state")
		}
		c.operations[r.RequestID] = &operation{request: r, state: op.State, receipt: op.Receipt}
	}
	if budget != state.Budget || budget.InFlight > 1 || budget.ModelReservations > c.delegation.Limits.MaxModelCalls || budget.ToolReservations > c.delegation.Limits.MaxToolCalls {
		return nil, errors.New("persisted budget mismatch")
	}
	for i, event := range events {
		if event.SchemaVersion != "development_attempt_event.v1" || event.Sequence != int64(i+1) || event.AttemptID != c.delegation.Scope.AttemptID || event.DelegationID != c.delegation.DelegationID || i > 0 && event.OccurredAtMS < events[i-1].OccurredAtMS {
			return nil, errors.New("invalid event stream")
		}
	}
	last := events[len(events)-1]
	if !slices.Contains([]string{"active", "cancel_requested", "cancelled", "completed", "failed", "outcome_unknown"}, c.phase) || !slices.Contains([]string{"not_requested", "pending", "unconfirmed", "confirmed"}, c.terminationStatus) || last.Phase != c.phase || last.TerminationStatus != c.terminationStatus || last.Reason != c.reason || last.Budget != c.budget || last.OccurredAtMS != c.lastTimeMS {
		return nil, errors.New("state does not match committed event")
	}
	if (c.phase == "active" || c.phase == "completed") && c.terminationStatus != "not_requested" || c.phase == "cancelled" && c.terminationStatus != "confirmed" {
		return nil, errors.New("invalid termination state")
	}
	// Re-encoding must preserve all private fields; no implicit defaults or repairs.
	normalized, err := encodeState(c)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(normalized, data) {
		return nil, errors.New("persisted state is not canonical")
	}
	return c, nil
}

// Interrupt fences ownership after an explicit recovery decision. It does not
// claim that an old executor, command or remote model has stopped.
func (c *AttemptControl) Interrupt(nowMS int64) (Event, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.phase != "active" && c.phase != "cancel_requested" {
		return Event{}, AttemptStopped
	}
	c.releaseUndispatched()
	c.phase = "outcome_unknown"
	c.reason = "runtime_interrupted"
	c.terminationStatus = "pending"
	for _, op := range c.operations {
		if op.state == "dispatched" {
			op.state = "uncertain"
			c.budget.InFlight--
			c.budget.Unresolved++
		}
	}
	return c.emit("attempt_interrupted", "", nowMS), nil
}

func (c *AttemptControl) Fail(reason string, nowMS int64) (Event, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.phase != "active" && c.phase != "cancel_requested" {
		return Event{}, AttemptStopped
	}
	if reason != "native_response_invalid" && reason != "sdk_worker_failed" {
		return Event{}, OperationStateInvalid
	}
	c.releaseUndispatched()
	c.phase = "failed"
	c.reason = reason
	c.terminationStatus = "pending"
	return c.emit("attempt_failed", "", nowMS), nil
}
