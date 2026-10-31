package developmentcontrol

import (
	"context"
	"encoding/json"
	"time"
)

func (r *Runtime) model(ctx context.Context, record *Record, requestID string, request NativeRequest) (NativeResult, error) {
	if r.backend == nil {
		return NativeResult{}, ErrTransport
	}
	if !idPattern.MatchString(requestID) || request.Model != record.Binding.Selection.WireModel || request.MaxCompletionTokens > record.Binding.Plan.MaxOutputTokens {
		return NativeResult{}, ErrContext
	}
	policy, err := r.backend.Policy(ctx, record.Binding.Plan.ModelRef)
	if err != nil || policy.InputTokenCeiling != record.Binding.InputTokenCeiling || policy.MaxOutputTokens < request.MaxCompletionTokens || policy.Timeout <= 0 || policy.Timeout > 2*time.Minute {
		return NativeResult{}, ErrTransport
	}
	deadline := r.now().Add(policy.Timeout)
	expiry, live := r.attemptDeadline(record.Control.delegation.Scope.AttemptID)
	if !live {
		return NativeResult{}, ErrCredential
	}
	if expiry.Before(deadline) {
		deadline = expiry
	}
	callCtx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()
	attemptID := record.Control.delegation.Scope.AttemptID
	if err := r.beginCall(attemptID, cancel); err != nil {
		return NativeResult{}, err
	}
	defer r.endCall(attemptID)
	data, err := json.Marshal(request)
	if err != nil {
		return NativeResult{}, ErrContext
	}
	cost, err := record.Binding.Selection.ReserveCost(policy.InputTokenCeiling, request.MaxCompletionTokens)
	if err != nil || cost < 1 {
		return NativeResult{}, ErrIneligible
	}
	var refused error
	prepared, err := r.store.change(callCtx, record.Binding.Identity, attemptID, 0, r.instanceID, func(current *Record) error {
		refused = r.revalidate(callCtx, current)
		if refused != nil {
			return nil
		}
		if err := checkContext(current, request.Messages); err != nil {
			return err
		}
		d := current.Control.delegation
		reservation := Reservation{"development_attempt_reservation.v1", d.DelegationID, d.Revision, attemptID, requestID, current.Control.requestSequence + 1, "model", PayloadDigest(data), policy.InputTokenCeiling + request.MaxCompletionTokens, cost}
		if _, err := current.Control.Reserve(reservation, authorityFor(current), r.now().UnixMilli()); err != nil {
			return err
		}
		_, err := current.Control.RecordDispatch(requestID, authorityFor(current), r.now().UnixMilli())
		return err
	})
	if err != nil {
		return NativeResult{}, err
	}
	if refused != nil {
		return NativeResult{}, refused
	}
	watched, stopWatch := context.WithCancel(callCtx)
	defer stopWatch()
	watchDone := make(chan struct{})
	go func() {
		defer close(watchDone)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-watched.Done():
				return
			case <-ticker.C:
				var refusal error
				_, err := r.store.change(watched, record.Binding.Identity, attemptID, 0, r.instanceID, func(current *Record) error { refusal = r.revalidate(watched, current); return nil })
				if err != nil || refusal != nil {
					cancel()
					return
				}
			}
		}
	}()
	result, backendErr := r.backend.Complete(callCtx, record.Binding.Plan.ModelRef, request)
	stopWatch()
	<-watchDone
	// Reconciliation is local accounting and may outlive client cancellation. It
	// never retries a model or command and never publishes a new offer after a stop.
	finishCtx, finishCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer finishCancel()
	if backendErr != nil {
		_, err := r.store.change(finishCtx, record.Binding.Identity, attemptID, 0, "", func(current *Record) error {
			if op := current.Control.operations[requestID]; op != nil && op.state == "uncertain" {
				current.Revoked = true
				return nil // A committed interruption already retained the unknown hold.
			}
			_, err := current.Control.RecordReceipt(Receipt{"development_attempt_receipt.v1", attemptID, requestID, "unknown", 0, 0}, r.now().UnixMilli())
			current.Revoked = true
			return err
		})
		if err != nil {
			return NativeResult{}, err
		}
		return NativeResult{}, ErrTransport
	}
	normalized, action, normalizeErr := normalizedResult(result, record.Binding.Plan.Tools)
	if action != nil && len(request.Messages) > 62 {
		normalizeErr = ErrContext
	}
	responseReceipt := Receipt{SchemaVersion: "development_attempt_receipt.v1", AttemptID: attemptID, RequestID: requestID, Status: "missing"}
	if result.Usage != nil && result.Usage.PromptTokens >= 0 && result.Usage.CompletionTokens >= 0 && result.Usage.TotalTokens == result.Usage.PromptTokens+result.Usage.CompletionTokens && result.Usage.TotalTokens <= 1000000000000 {
		responseReceipt.Status = "reported"
		responseReceipt.Tokens = result.Usage.TotalTokens
		responseReceipt.CostMicrousd, err = prepared.Binding.Selection.Cost(result.Usage.PromptTokens, result.Usage.CompletionTokens)
		if err != nil || responseReceipt.Validate() != nil {
			responseReceipt = Receipt{"development_attempt_receipt.v1", attemptID, requestID, "unknown", 0, 0}
			normalizeErr = ErrContext
		}
	}
	var stopped bool
	_, err = r.store.change(finishCtx, record.Binding.Identity, attemptID, 0, "", func(current *Record) error {
		if _, err := current.Control.RecordReceipt(responseReceipt, r.now().UnixMilli()); err != nil {
			return err
		}
		if current.OwnerInstanceID != r.instanceID || current.Revoked || current.Control.phase != "active" {
			current.Revoked = true
			stopped = true
			return nil
		}
		if refusal := r.revalidate(finishCtx, current); refusal != nil {
			stopped = true
			return nil
		}
		if normalizeErr != nil {
			if _, err := current.Control.Fail("native_response_invalid", r.now().UnixMilli()); err != nil {
				return err
			}
			current.Revoked = true
			return nil
		}
		if action == nil {
			_, err := current.Control.Complete(r.now().UnixMilli())
			current.Revoked = true
			return err
		}
		call := normalized.Message.ToolCalls[0]
		toolRequestID, err := randomID("tool_")
		if err != nil {
			return err
		}
		proposal := ToolCall{"development_tool_call.v1", attemptID, toolRequestID, requestID, call.ID, *action}
		body, err := json.Marshal(proposal)
		if err != nil {
			return err
		}
		arguments, err := json.Marshal(*action)
		if err != nil {
			return err
		}
		for _, prior := range current.Offers {
			if prior.ToolCallID == call.ID || prior.ParentRequestID == requestID {
				return OperationDuplicate
			}
		}
		current.Offers[toolRequestID] = Offer{toolRequestID, requestID, call.ID, PayloadDigest(body), PayloadDigest(arguments), "pending", ""}
		current.Control.emit("tool_offer_registered", toolRequestID, r.now().UnixMilli())
		current.ContextDigest, err = messagesDigest(append(append([]NativeMessage(nil), request.Messages...), normalized.Message))
		if err != nil {
			return err
		}
		current.ContextCount = int64(len(request.Messages) + 1)
		current.PendingToolRequestID = toolRequestID
		return nil
	})
	if err != nil {
		return NativeResult{}, err
	}
	if normalizeErr != nil {
		return NativeResult{}, ErrContext
	}
	if stopped {
		return NativeResult{}, AttemptStopped
	}
	return normalized, nil
}

func checkContext(record *Record, messages []NativeMessage) error {
	if len(messages) == 0 {
		return ErrContext
	}
	if record.ContextCount == 0 {
		digest, err := messagesDigest(messages)
		if err != nil || digest != record.Binding.Plan.InitialMessagesDigest {
			return ErrContext
		}
		for _, message := range messages {
			if message.Role != "system" && message.Role != "user" {
				return ErrContext
			}
		}
		if messages[len(messages)-1].Role != "user" {
			return ErrContext
		}
		return nil
	}
	if int64(len(messages)) != record.ContextCount+1 {
		return ErrContext
	}
	digest, err := messagesDigest(messages[:record.ContextCount])
	if err != nil || digest != record.ContextDigest {
		return ErrContext
	}
	offer, ok := record.Offers[record.PendingToolRequestID]
	if !ok || offer.State != "completed" {
		return ErrContext
	}
	last := messages[len(messages)-1]
	body, err := json.Marshal(last)
	if err != nil || last.Role != "tool" || last.ToolCallID != offer.ToolCallID || PayloadDigest(body) != offer.ObservationDigest {
		return ErrContext
	}
	return nil
}

func (r *Runtime) QueueTool(ctx context.Context, record *Record, call ToolCall) (*Record, error) {
	if call.Validate() != nil || call.AttemptID != record.Control.delegation.Scope.AttemptID {
		return nil, ErrContext
	}
	data, err := json.Marshal(call)
	if err != nil {
		return nil, ErrContext
	}
	arguments, err := json.Marshal(call.Arguments)
	if err != nil {
		return nil, ErrContext
	}
	var refused error
	updated, err := r.store.change(ctx, record.Binding.Identity, call.AttemptID, 0, r.instanceID, func(current *Record) error {
		refused = r.revalidate(ctx, current)
		if refused != nil {
			return nil
		}
		offer, ok := current.Offers[call.RequestID]
		if !ok || offer.State != "pending" || offer.ParentRequestID != call.ParentRequestID || offer.ToolCallID != call.ToolCallID || offer.PayloadDigest != PayloadDigest(data) || offer.ArgumentsDigest != PayloadDigest(arguments) || !current.Binding.Plan.Tools.Allows(call.Arguments) {
			return ErrContext
		}
		d := current.Control.delegation
		reservation := Reservation{"development_attempt_reservation.v1", d.DelegationID, d.Revision, d.Scope.AttemptID, call.RequestID, current.Control.requestSequence + 1, "tool", offer.PayloadDigest, 0, 0}
		if _, err := current.Control.Reserve(reservation, authorityFor(current), r.now().UnixMilli()); err != nil {
			return err
		}
		offer.State = "queued"
		current.Offers[call.RequestID] = offer
		current.Control.emit("tool_offer_queued", call.RequestID, r.now().UnixMilli())
		return nil
	})
	if err != nil {
		return nil, err
	}
	if refused != nil {
		return nil, refused
	}
	return updated, nil
}

// DispatchTool and CompleteTool are trusted Go executor hooks, not SDK routes.
// The actual filesystem / command executor is intentionally not implemented here.
func (r *Runtime) DispatchTool(ctx context.Context, identity Identity, attemptID, requestID string) (*Record, error) {
	var refused error
	record, err := r.store.change(ctx, identity, attemptID, 0, r.instanceID, func(current *Record) error {
		refused = r.revalidate(ctx, current)
		if refused != nil {
			return nil
		}
		offer, ok := current.Offers[requestID]
		if !ok || offer.State != "queued" {
			return ErrContext
		}
		_, err := current.Control.RecordDispatch(requestID, authorityFor(current), r.now().UnixMilli())
		return err
	})
	if err != nil {
		return nil, err
	}
	if refused != nil {
		return nil, refused
	}
	return record, nil
}
func (r *Runtime) CompleteTool(ctx context.Context, identity Identity, attemptID, requestID, observation string) (*Record, error) {
	if len(observation) > 65536 {
		return nil, ErrContext
	}
	return r.store.change(ctx, identity, attemptID, 0, "", func(current *Record) error {
		offer, ok := current.Offers[requestID]
		if !ok || offer.State != "queued" {
			return ErrContext
		}
		if _, err := current.Control.RecordReceipt(Receipt{"development_attempt_receipt.v1", attemptID, requestID, "not_applicable", 0, 0}, r.now().UnixMilli()); err != nil {
			return err
		}
		body, err := json.Marshal(NativeMessage{Role: "tool", Content: observation, ToolCallID: offer.ToolCallID, Name: "development_tool"})
		if err != nil {
			return err
		}
		offer.State = "completed"
		offer.ObservationDigest = PayloadDigest(body)
		current.Offers[requestID] = offer
		return nil
	})
}
