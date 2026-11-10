package developmentcontrol

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSDKMetadataIsSeparatedFromCanonicalToolAction(t *testing.T) {
	original := fixture(t, "tool-action")
	action, err := ParseToolAction(original)
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	if err := json.Unmarshal(original, &wire); err != nil {
		t.Fatal(err)
	}
	wire["summary"] = "Replace the fixed task source"
	data, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseToolAction(data); err == nil {
		t.Fatal("SDK metadata entered canonical authorization")
	}
	translated, summary, err := ParseSDKToolArguments(data)
	if err != nil || translated != action || summary != wire["summary"] {
		t.Fatalf("explicit metadata translation failed: %v", err)
	}
	delete(wire, "summary")
	wire["api_key"] = "forbidden"
	data, _ = json.Marshal(wire)
	if _, _, err := ParseSDKToolArguments(data); err == nil {
		t.Fatal("unknown transport metadata accepted")
	}
	for _, bad := range []string{`null`, `{"summary":true}`, strings.Replace(strings.TrimSpace(string(original)), `"content":`, `"content":null,"content":`, 1)} {
		if _, _, err := ParseSDKToolArguments([]byte(bad)); err == nil {
			t.Fatal("invalid SDK arguments accepted")
		}
	}
	for _, path := range []string{"../list.go", "/list.go", "a/../list.go", "a//b", "a/", "a\\b", "C:target", ".", "..", "secret\x00file", strings.Repeat("中", 100)} {
		changed := action
		changed.Path = path
		if err := changed.Validate(); err == nil {
			t.Fatalf("unsafe path accepted: %q", path)
		}
	}
	action.Action = "read_file"
	action.Content = ""
	action.ExpectedDigest = ""
	if err := action.Validate(); err != nil {
		t.Fatal(err)
	}
	action.Action = "run_tests"
	action.Path = ""
	action.CandidateDigest = "sha256:" + strings.Repeat("a", 64)
	if err := action.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestToolReservationBindsTrustedOfferAndPayload(t *testing.T) {
	_, _, _, r := setup(t)
	call, err := ParseToolCall(fixture(t, "tool-call"))
	if err != nil {
		t.Fatal(err)
	}
	r.Kind = "tool"
	r.RequestID = call.RequestID
	r.TokenBound = 0
	r.CostBoundMicrousd = 0
	data, err := json.Marshal(call)
	if err != nil {
		t.Fatal(err)
	}
	r.PayloadDigest = PayloadDigest(data)
	if err := ValidateToolReservation(r, call, call); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*ToolCall){func(c *ToolCall) { c.ParentRequestID = "req_other" }, func(c *ToolCall) { c.ToolCallID = "call_other" }, func(c *ToolCall) { c.Arguments.Content = "changed" }, func(c *ToolCall) { c.AttemptID = "attempt_other" }} {
		changed := call
		mutate(&changed)
		if err := ValidateToolReservation(r, changed, call); err == nil {
			t.Fatal("tool identity substitution accepted")
		}
	}
	r.PayloadDigest = "sha256:" + strings.Repeat("b", 64)
	if err := ValidateToolReservation(r, call, call); err == nil {
		t.Fatal("stale payload reservation accepted")
	}
}

func TestDelegationUsesExistingCatalogueIdentityWithoutURL(t *testing.T) {
	_, d, _, _ := setup(t)
	for _, ref := range []string{"profile:probe", "provider:huggingface:profile:dev", "mock"} {
		d.Scope.ModelRef = ref
		if err := d.Validate(); err != nil {
			t.Fatalf("catalogue identity rejected: %s %v", ref, err)
		}
	}
	for _, ref := range []string{"", "https://outside/model", "profile:bad name", "profile:probe\n"} {
		d.Scope.ModelRef = ref
		if err := d.Validate(); err == nil {
			t.Fatalf("invalid reference accepted: %q", ref)
		}
	}
}
