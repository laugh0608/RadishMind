package httpapi

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	control "radishmind.local/services/platform/internal/developmentcontrol"
	"radishmind.local/services/platform/internal/developmentsandbox"
)

// Each answer is selected from the trusted conversation's last observation, so
// separate attempts never share a counter, file body or candidate digest.
type fixedTaskBackend struct{}

func (fixedTaskBackend) Policy(_ context.Context, ref string) (control.NativePolicy, error) {
	if ref != fixedTaskModel {
		return control.NativePolicy{}, control.ErrIneligible
	}
	return control.NativePolicy{InputTokenCeiling: 100, MaxOutputTokens: 50, Timeout: 5 * time.Second}, nil
}
func (fixedTaskBackend) Complete(ctx context.Context, ref string, in control.NativeRequest) (control.NativeResult, error) {
	if ctx.Err() != nil {
		return control.NativeResult{}, ctx.Err()
	}
	if ref != fixedTaskModel || in.Model != fixedTaskModel || len(in.Messages) < 2 {
		return control.NativeResult{}, control.ErrContext
	}
	usage := &control.NativeUsage{PromptTokens: 20, CompletionTokens: 10, TotalTokens: 30}
	last := in.Messages[len(in.Messages)-1]
	a := control.ToolAction{SchemaVersion: "development_tool_action.v1", Action: "read_file", Path: "list.go"}
	if last.Role != "user" || len(in.Messages) != 2 {
		var o developmentsandbox.Observation
		if last.Role != "tool" || json.Unmarshal([]byte(last.Content), &o) != nil || o.SchemaVersion != "development_sandbox_observation.v1" {
			return control.NativeResult{}, control.ErrContext
		}
		switch o.Action {
		case "read_file":
			if o.Path != "list.go" || o.CandidateDigest != fixedTaskBaseline || o.FileDigest != control.PayloadDigest([]byte(o.Content)) || strings.Count(o.Content, "result[:limit-1]") != 1 {
				return control.NativeResult{}, control.ErrContext
			}
			a.Action = "replace_file"
			a.ExpectedDigest = o.FileDigest
			a.Content = strings.Replace(o.Content, "result[:limit-1]", "result[:limit]", 1)
		case "replace_file":
			if o.Path != "list.go" {
				return control.NativeResult{}, control.ErrContext
			}
			a.Action = "run_tests"
			a.Path = ""
			a.CandidateDigest = o.CandidateDigest
		case "run_tests":
			if o.Verification == nil || o.Verification.Status != "command_passed" {
				return control.NativeResult{}, control.ErrContext
			}
			return control.NativeResult{Message: control.NativeMessage{Role: "assistant", Content: "Fixed-response verification finished."}, Usage: usage, FinishReason: "stop"}, nil
		default:
			return control.NativeResult{}, control.ErrContext
		}
	}
	if a.Validate() != nil {
		return control.NativeResult{}, control.ErrContext
	}
	raw, err := json.Marshal(a)
	if err != nil {
		return control.NativeResult{}, err
	}
	return control.NativeResult{Message: control.NativeMessage{Role: "assistant", ToolCalls: []control.NativeCall{{ID: "call_fixed_" + a.Action, Type: "function", Function: control.NativeFunction{Name: "development_tool", Arguments: string(raw)}}}}, Usage: usage, FinishReason: "tool_calls"}, nil
}
