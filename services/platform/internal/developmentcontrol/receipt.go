package developmentcontrol

// RecordReceipt may reconcile a late trusted receipt after cancellation. It
// changes accounting only and never reactivates a terminal attempt.
func (c *AttemptControl) RecordReceipt(r Receipt, nowMS int64) (Event, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := r.Validate(); err != nil {
		return Event{}, err
	}
	if r.AttemptID != c.delegation.Scope.AttemptID {
		return Event{}, ReceiptInvalid
	}
	op, ok := c.operations[r.RequestID]
	if !ok {
		return Event{}, OperationStateInvalid
	}
	if r.Status == "not_dispatched" {
		if op.state != "reserved" {
			return Event{}, OperationStateInvalid
		}
	} else {
		if op.state != "dispatched" && op.state != "uncertain" && op.state != "unmetered" {
			return Event{}, OperationStateInvalid
		}
		if (r.Status == "missing" || r.Status == "unknown") && op.state != "dispatched" {
			return Event{}, OperationStateInvalid
		}
	}
	if r.Status == "reported" && op.request.Kind != "model" || r.Status == "missing" && op.request.Kind != "model" || r.Status == "not_applicable" && op.request.Kind != "tool" {
		return Event{}, ReceiptInvalid
	}
	if r.Status == "missing" || r.Status == "unknown" {
		c.budget.InFlight--
		c.budget.Unresolved++
		op.state = "unmetered"
		if r.Status == "unknown" {
			op.state = "uncertain"
			if c.phase == "active" || c.phase == "cancel_requested" {
				c.phase = "outcome_unknown"
				c.reason = "execution_unknown"
				c.terminationStatus = "pending"
			}
		}
	} else {
		if op.state == "uncertain" || op.state == "unmetered" {
			c.budget.Unresolved--
		} else {
			c.budget.InFlight--
		}
		c.budget.HeldTokens -= op.request.TokenBound
		c.budget.HeldCostMicrousd -= op.request.CostBoundMicrousd
		c.budget.SpentTokens += r.Tokens
		c.budget.SpentCostMicrousd += r.CostMicrousd
		op.state = "settled"
		if r.Status == "not_dispatched" {
			op.state = "not_dispatched"
		}
		if r.Tokens > op.request.TokenBound || r.CostMicrousd > op.request.CostBoundMicrousd {
			// Preserve the actual report in full; never clamp it to the reservation.
			c.reason = "reservation_exceeded"
			if c.phase == "active" || c.phase == "cancel_requested" {
				c.phase = "failed"
				c.terminationStatus = "pending"
			}
		}
	}
	return c.emit("receipt_recorded", r.RequestID, nowMS), nil
}
