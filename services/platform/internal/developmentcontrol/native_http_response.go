package developmentcontrol

import (
	"bytes"
	"encoding/json"
	"io"
	"slices"
	"unicode/utf8"

	"radishmind.local/services/platform/internal/strictjson"
)

// Provider JSON permits explicit null for content/usage and metadata. Validate
// duplicate keys and depth before selecting fields, without relaxing the strict
// internal contracts or silently supplying missing counters with zero values.
func inspectNativeJSON(d *json.Decoder, depth int) error {
	if depth > 24 {
		return ErrContext
	}
	token, err := d.Token()
	if err != nil {
		return ErrContext
	}
	delim, nested := token.(json.Delim)
	if !nested {
		return nil
	}
	if delim != '{' && delim != '[' {
		return ErrContext
	}
	seen := map[string]bool{}
	for d.More() {
		if delim == '{' {
			key, err := d.Token()
			name, ok := key.(string)
			if err != nil || !ok || seen[name] {
				return ErrContext
			}
			seen[name] = true
		}
		if err := inspectNativeJSON(d, depth+1); err != nil {
			return err
		}
	}
	_, err = d.Token()
	return err
}

func nativeResponseFields(data []byte, required, optional []string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(data, &fields) != nil || fields == nil {
		return nil, ErrContext
	}
	for _, name := range required {
		if len(fields[name]) == 0 || bytes.Equal(bytes.TrimSpace(fields[name]), []byte("null")) {
			return nil, ErrContext
		}
	}
	for name := range fields {
		if !slices.Contains(required, name) && !slices.Contains(optional, name) {
			return nil, ErrContext
		}
	}
	return fields, nil
}

func nativeNullOrMissing(data []byte) bool {
	return len(data) == 0 || bytes.Equal(bytes.TrimSpace(data), []byte("null"))
}

func parseNativeHTTPResponse(data []byte, model string) (NativeResult, error) {
	var result NativeResult
	if len(data) > nativeHTTPBodyLimit || !utf8.Valid(data) {
		return result, ErrContext
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	if err := inspectNativeJSON(d, 0); err != nil {
		return result, ErrContext
	}
	if _, err := d.Token(); err != io.EOF {
		return result, ErrContext
	}
	fields, err := nativeResponseFields(data, []string{"id", "object", "created", "model", "choices"}, []string{"usage", "system_fingerprint", "service_tier"})
	if err != nil {
		return result, err
	}
	var id, object, wireModel string
	var created int64
	if json.Unmarshal(fields["id"], &id) != nil || id == "" || len(id) > 512 || json.Unmarshal(fields["object"], &object) != nil || object != "chat.completion" || json.Unmarshal(fields["model"], &wireModel) != nil || wireModel != model || json.Unmarshal(fields["created"], &created) != nil || created < 0 {
		return result, ErrContext
	}
	for _, key := range []string{"system_fingerprint", "service_tier"} {
		var value string
		if !nativeNullOrMissing(fields[key]) && (json.Unmarshal(fields[key], &value) != nil || len(value) > 512) {
			return result, ErrContext
		}
	}
	var choices []json.RawMessage
	if json.Unmarshal(fields["choices"], &choices) != nil || len(choices) != 1 {
		return result, ErrContext
	}
	choice, err := nativeResponseFields(choices[0], []string{"index", "message", "finish_reason"}, []string{"logprobs"})
	if err != nil {
		return result, err
	}
	var index int
	if json.Unmarshal(choice["index"], &index) != nil || index != 0 || !nativeNullOrMissing(choice["logprobs"]) || json.Unmarshal(choice["finish_reason"], &result.FinishReason) != nil {
		return result, ErrContext
	}
	message, err := nativeResponseFields(choice["message"], []string{"role"}, []string{"content", "tool_calls", "refusal"})
	if err != nil || !nativeNullOrMissing(message["refusal"]) {
		return result, ErrContext
	}
	if json.Unmarshal(message["role"], &result.Message.Role) != nil || result.Message.Role != "assistant" {
		return result, ErrContext
	}
	if !nativeNullOrMissing(message["content"]) && json.Unmarshal(message["content"], &result.Message.Content) != nil {
		return result, ErrContext
	}
	if calls, exists := message["tool_calls"]; exists {
		if nativeNullOrMissing(calls) || strictjson.Decode(calls, &result.Message.ToolCalls) != nil {
			return result, ErrContext
		}
		var rawCalls []json.RawMessage
		if json.Unmarshal(calls, &rawCalls) != nil {
			return result, ErrContext
		}
		for _, call := range rawCalls {
			f, err := strictjson.ObjectFields(call, "id", "type", "function")
			if err != nil {
				return result, ErrContext
			}
			if _, err := strictjson.ObjectFields(f["function"], "name", "arguments"); err != nil {
				return result, ErrContext
			}
		}
	}
	if validateMessage(result.Message) != nil {
		return result, ErrContext
	}
	if len(result.Message.ToolCalls) == 0 {
		if result.FinishReason != "stop" {
			return result, ErrContext
		}
	} else {
		if result.FinishReason != "tool_calls" {
			return result, ErrContext
		}
		call, err := normalizeNativeCall(result.Message.ToolCalls[0])
		if err != nil {
			return result, ErrContext
		}
		result.Message.ToolCalls[0] = call
	}
	if !nativeNullOrMissing(fields["usage"]) {
		usage, err := nativeResponseFields(fields["usage"], []string{"prompt_tokens", "completion_tokens", "total_tokens"}, []string{"prompt_tokens_details", "completion_tokens_details"})
		if err != nil {
			return result, ErrContext
		}
		u := &NativeUsage{}
		for key, target := range map[string]*int64{"prompt_tokens": &u.PromptTokens, "completion_tokens": &u.CompletionTokens, "total_tokens": &u.TotalTokens} {
			if json.Unmarshal(usage[key], target) != nil || *target < 0 || *target > 1000000000000 {
				return result, ErrContext
			}
		}
		if u.TotalTokens != u.PromptTokens+u.CompletionTokens {
			return result, ErrContext
		}
		result.Usage = u
	}
	return result, nil
}
