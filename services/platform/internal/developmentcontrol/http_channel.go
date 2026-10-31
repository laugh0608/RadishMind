package developmentcontrol

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
)

func (r *Runtime) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /internal/development/v1/attempts/{attempt_id}/chat/completions", r.handleModel)
	mux.HandleFunc("POST /internal/development/v1/attempts/{attempt_id}/tool-calls", r.handleTool)
	mux.HandleFunc("GET /internal/development/v1/attempts/{attempt_id}/tool-offers/{tool_call_id}", r.handleOffer)
	mux.HandleFunc("GET /internal/development/v1/attempts/{attempt_id}/tool-catalog", r.handleCatalog)
	mux.HandleFunc("GET /internal/development/v1/attempts/{attempt_id}/events", r.handleEvents)
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		host, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil || net.ParseIP(host) == nil || !net.ParseIP(host).IsLoopback() || req.Header.Get("Origin") != "" {
			channelError(w, ErrCredential)
			return
		}
		authority := req.Host
		if parsed, _, err := net.SplitHostPort(authority); err == nil {
			authority = parsed
		}
		if authority != "localhost" && (net.ParseIP(authority) == nil || !net.ParseIP(authority).IsLoopback()) {
			channelError(w, ErrCredential)
			return
		}
		mux.ServeHTTP(w, req)
	})
}
func (r *Runtime) requestRecord(req *http.Request) (*Record, error) {
	values := req.Header.Values("Authorization")
	if len(values) != 1 || !strings.HasPrefix(values[0], "Bearer devtask_") {
		return nil, ErrCredential
	}
	token := strings.TrimPrefix(values[0], "Bearer ")
	if len(token) != 51 {
		return nil, ErrCredential
	}
	record, err := r.store.authenticate(req.Context(), r.instanceID, req.PathValue("attempt_id"), token)
	if err != nil {
		return nil, err
	}
	deadline, live := r.attemptDeadline(record.Control.delegation.Scope.AttemptID)
	if !live || !r.now().Before(deadline) || record.Revoked || r.now().UnixMilli() >= record.Control.delegation.ExpiresAtMS {
		return nil, ErrCredential
	}
	return record, nil
}
func channelBody(w http.ResponseWriter, req *http.Request) ([]byte, error) {
	if len(req.URL.Query()) != 0 || strings.Split(req.Header.Get("Content-Type"), ";")[0] != "application/json" {
		return nil, ErrContext
	}
	return io.ReadAll(http.MaxBytesReader(w, req.Body, 262144))
}
func (r *Runtime) handleModel(w http.ResponseWriter, req *http.Request) {
	record, err := r.requestRecord(req)
	if err != nil {
		channelError(w, err)
		return
	}
	ids := req.Header.Values("X-RadishMind-Development-Request-ID")
	if len(ids) != 1 || !idPattern.MatchString(ids[0]) {
		channelError(w, ErrContext)
		return
	}
	data, err := channelBody(w, req)
	if err != nil {
		channelError(w, ErrContext)
		return
	}
	input, err := ParseNativeRequest(data)
	if err != nil {
		channelError(w, ErrContext)
		return
	}
	result, err := r.model(req.Context(), record, ids[0], input)
	if err != nil {
		channelError(w, err)
		return
	}
	response := map[string]any{"id": "chatcmpl_" + ids[0], "object": "chat.completion", "created": r.now().Unix(), "model": record.Binding.Selection.WireModel, "choices": []map[string]any{{"index": 0, "message": result.Message, "finish_reason": result.FinishReason}}}
	if result.Usage != nil {
		response["usage"] = result.Usage
	}
	writeChannelJSON(w, http.StatusOK, response)
}
func (r *Runtime) handleTool(w http.ResponseWriter, req *http.Request) {
	record, err := r.requestRecord(req)
	if err != nil {
		channelError(w, err)
		return
	}
	data, err := channelBody(w, req)
	if err != nil {
		channelError(w, ErrContext)
		return
	}
	call, err := ParseToolCall(data)
	if err != nil {
		channelError(w, ErrContext)
		return
	}
	record, err = r.QueueTool(req.Context(), record, call)
	if err != nil {
		channelError(w, err)
		return
	}
	writeChannelJSON(w, http.StatusAccepted, map[string]any{"status": "queued", "request_id": call.RequestID, "record_version": record.Version})
}
func (r *Runtime) handleOffer(w http.ResponseWriter, req *http.Request) {
	record, err := r.requestRecord(req)
	if err != nil {
		channelError(w, err)
		return
	}
	if len(req.URL.Query()) != 0 {
		channelError(w, ErrContext)
		return
	}
	for _, offer := range record.Offers {
		if offer.ToolCallID == req.PathValue("tool_call_id") {
			writeChannelJSON(w, 200, offer)
			return
		}
	}
	channelError(w, ErrNotFound)
}
func (r *Runtime) handleCatalog(w http.ResponseWriter, req *http.Request) {
	if _, err := r.requestRecord(req); err != nil {
		channelError(w, err)
		return
	}
	if len(req.URL.Query()) != 0 {
		channelError(w, ErrContext)
		return
	}
	tools, err := NativeTools()
	if err != nil {
		channelError(w, ErrUnavailable)
		return
	}
	writeChannelJSON(w, 200, map[string]any{"tools": tools})
}
func (r *Runtime) handleEvents(w http.ResponseWriter, req *http.Request) {
	record, err := r.requestRecord(req)
	if err != nil {
		channelError(w, err)
		return
	}
	after, limit := 0, 50
	for key, values := range req.URL.Query() {
		if len(values) != 1 {
			channelError(w, ErrContext)
			return
		}
		n, err := strconv.Atoi(values[0])
		if err != nil {
			channelError(w, ErrContext)
			return
		}
		switch key {
		case "after_sequence":
			after = n
		case "limit":
			limit = n
		default:
			channelError(w, ErrContext)
			return
		}
	}
	events := record.Control.Events()
	if after < 0 || after > len(events) || limit < 1 || limit > 50 {
		channelError(w, ErrContext)
		return
	}
	end := min(after+limit, len(events))
	writeChannelJSON(w, 200, map[string]any{"events": events[after:end], "next_sequence": end, "has_more": end < len(events), "record_version": record.Version})
}
func writeChannelJSON(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		channelError(w, ErrUnavailable)
		return
	}
	w.WriteHeader(status)
	_, _ = w.Write(append(data, '\n'))
}
func channelError(w http.ResponseWriter, err error) {
	status, code := http.StatusServiceUnavailable, "development_unavailable"
	switch {
	case errors.Is(err, ErrCredential):
		status, code = 401, ErrCredential.Error()
	case errors.Is(err, ErrNotFound):
		status, code = 404, ErrNotFound.Error()
	case errors.Is(err, ErrIneligible), errors.Is(err, DelegationInvalid):
		status, code = 403, "development_qualification_denied"
	case errors.Is(err, ErrContext):
		status, code = 400, "development_request_invalid"
	case errors.Is(err, BudgetExhausted):
		status, code = 429, "development_budget_exhausted"
	case errors.Is(err, ErrConflict), errors.Is(err, OperationDuplicate), errors.Is(err, OperationSequenceInvalid), errors.Is(err, OperationStateInvalid), errors.Is(err, AttemptStopped):
		status, code = 409, "development_state_conflict"
	}
	data, _ := json.Marshal(map[string]any{"error": map[string]string{"code": code, "type": "invalid_request_error", "message": code}})
	w.WriteHeader(status)
	_, _ = w.Write(append(data, '\n'))
}
