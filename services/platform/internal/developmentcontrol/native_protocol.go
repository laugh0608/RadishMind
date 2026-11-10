package developmentcontrol

import (
	_ "embed"
	"encoding/json"
	"errors"
	"unicode/utf8"

	"radishmind.local/services/platform/internal/strictjson"
)

// This generated distribution copy is checked byte-for-byte against contracts/.
//
//go:embed assets/development-tool-action.schema.json
var actionSchema []byte

type NativeFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}
type NativeCall struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Function NativeFunction `json:"function"`
}
type NativeMessage struct {
	Role       string       `json:"role"`
	Content    string       `json:"content,omitempty"`
	ToolCalls  []NativeCall `json:"tool_calls,omitempty"`
	ToolCallID string       `json:"tool_call_id,omitempty"`
	Name       string       `json:"name,omitempty"`
}
type NativeRequest struct {
	Model               string            `json:"model"`
	Messages            []NativeMessage   `json:"messages"`
	MaxCompletionTokens int64             `json:"max_completion_tokens"`
	Tools               []json.RawMessage `json:"tools"`
	Stream              bool              `json:"stream,omitempty"`
	ParallelToolCalls   bool              `json:"parallel_tool_calls,omitempty"`
}
type NativeUsage struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
}
type NativeResult struct {
	Message      NativeMessage
	Usage        *NativeUsage
	FinishReason string
}

func NativeTools() ([]json.RawMessage, error) {
	var parameters map[string]json.RawMessage
	if err := json.Unmarshal(actionSchema, &parameters); err != nil {
		return nil, err
	}
	delete(parameters, "$schema")
	delete(parameters, "$id")
	var properties map[string]json.RawMessage
	if err := json.Unmarshal(parameters["properties"], &properties); err != nil {
		return nil, err
	}
	properties["summary"] = json.RawMessage(`{"type":"string","maxLength":512,"description":"Untrusted display metadata; not an authorization field."}`)
	value, err := json.Marshal(properties)
	if err != nil {
		return nil, err
	}
	parameters["properties"] = value
	raw, err := json.Marshal(map[string]any{"type": "function", "function": map[string]any{"name": "development_tool", "description": "Request a bounded development action through the task controller.", "parameters": parameters}})
	if err != nil {
		return nil, err
	}
	return []json.RawMessage{raw}, nil
}

func ParseNativeRequest(data []byte) (NativeRequest, error) {
	var request NativeRequest
	if err := strictjson.Decode(data, &request); err != nil {
		return request, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return request, err
	}
	names := []string{"model", "messages", "max_completion_tokens", "tools"}
	for _, name := range []string{"stream", "parallel_tool_calls"} {
		if _, ok := fields[name]; ok {
			names = append(names, name)
		}
	}
	if _, err := strictjson.ObjectFields(data, names...); err != nil {
		return request, err
	}
	if request.Stream || request.ParallelToolCalls || len(request.Messages) < 1 || len(request.Messages) > 64 || request.MaxCompletionTokens < 1 || request.MaxCompletionTokens > 262144 {
		return request, ErrContext
	}
	expected, err := NativeTools()
	if err != nil {
		return request, err
	}
	if len(request.Tools) != 1 {
		return request, ErrContext
	}
	var actualValue, expectedValue any
	if json.Unmarshal(request.Tools[0], &actualValue) != nil || json.Unmarshal(expected[0], &expectedValue) != nil {
		return request, ErrContext
	}
	actualBytes, _ := json.Marshal(actualValue)
	expectedBytes, _ := json.Marshal(expectedValue)
	if string(actualBytes) != string(expectedBytes) {
		return request, ErrContext
	}
	var rawMessages []json.RawMessage
	if err := json.Unmarshal(fields["messages"], &rawMessages); err != nil {
		return request, err
	}
	for index, message := range request.Messages {
		if err := validateMessage(message); err != nil {
			return request, err
		}
		// Nested JSON keys need the same exact casing/presence rules as the root.
		var messageFields map[string]json.RawMessage
		if err := json.Unmarshal(rawMessages[index], &messageFields); err != nil {
			return request, err
		}
		required := []string{"role"}
		for _, key := range []string{"content", "tool_calls", "tool_call_id", "name"} {
			if _, ok := messageFields[key]; ok {
				required = append(required, key)
			}
		}
		if _, err := strictjson.ObjectFields(rawMessages[index], required...); err != nil {
			return request, err
		}
		var rawCalls []json.RawMessage
		if len(message.ToolCalls) != 0 {
			if err := json.Unmarshal(messageFields["tool_calls"], &rawCalls); err != nil {
				return request, err
			}
		}
		for i := range message.ToolCalls {
			callFields, err := strictjson.ObjectFields(rawCalls[i], "id", "type", "function")
			if err != nil {
				return request, err
			}
			if _, err := strictjson.ObjectFields(callFields["function"], "name", "arguments"); err != nil {
				return request, err
			}
			canonical, err := normalizeNativeCall(message.ToolCalls[i])
			if err != nil {
				return request, err
			}
			request.Messages[index].ToolCalls[i] = canonical
		}
	}
	return request, nil
}

func validateMessage(m NativeMessage) error {
	if !utf8.ValidString(m.Content) || len(m.Content) > 65536 {
		return ErrContext
	}
	switch m.Role {
	case "system", "user":
		if m.Content == "" || len(m.ToolCalls) != 0 || m.ToolCallID != "" || m.Name != "" {
			return ErrContext
		}
	case "assistant":
		if m.ToolCallID != "" || m.Name != "" || len(m.ToolCalls) > 1 || m.Content == "" && len(m.ToolCalls) == 0 {
			return ErrContext
		}
	case "tool":
		if !toolCallIDPattern.MatchString(m.ToolCallID) || m.Name != "development_tool" || len(m.ToolCalls) != 0 {
			return ErrContext
		}
	default:
		return ErrContext
	}
	return nil
}

func normalizeNativeCall(call NativeCall) (NativeCall, error) {
	if !toolCallIDPattern.MatchString(call.ID) || call.Type != "function" || call.Function.Name != "development_tool" {
		return call, ErrContext
	}
	action, _, err := ParseSDKToolArguments([]byte(call.Function.Arguments))
	if err != nil {
		return call, err
	}
	data, err := json.Marshal(action)
	if err != nil {
		return call, err
	}
	call.Function.Arguments = string(data)
	return call, nil
}

func normalizedResult(result NativeResult, policy ToolPolicy) (NativeResult, *ToolAction, error) {
	if result.Message.Role != "assistant" || validateMessage(result.Message) != nil {
		return result, nil, ErrContext
	}
	if result.Usage != nil && (result.Usage.PromptTokens < 0 || result.Usage.CompletionTokens < 0 || result.Usage.TotalTokens != result.Usage.PromptTokens+result.Usage.CompletionTokens || result.Usage.TotalTokens > 1000000000000) {
		return result, nil, ErrContext
	}
	if len(result.Message.ToolCalls) == 0 {
		if result.FinishReason != "stop" {
			return result, nil, errors.New("unexpected model finish reason")
		}
		return result, nil, nil
	}
	if result.FinishReason != "tool_calls" {
		return result, nil, ErrContext
	}
	call, err := normalizeNativeCall(result.Message.ToolCalls[0])
	if err != nil {
		return result, nil, err
	}
	action, err := ParseToolAction([]byte(call.Function.Arguments))
	if err != nil || !policy.Allows(action) {
		return result, nil, ErrIneligible
	}
	result.Message.ToolCalls[0] = call
	return result, &action, nil
}

func messagesDigest(messages []NativeMessage) (string, error) {
	data, err := json.Marshal(messages)
	if err != nil {
		return "", err
	}
	return PayloadDigest(data), nil
}
