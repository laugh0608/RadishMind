package developmentcontrol

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"path"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"radishmind.local/services/platform/internal/strictjson"
)

type ToolAction struct {
	SchemaVersion   string `json:"schema_version"`
	Action          string `json:"action"`
	Path            string `json:"path"`
	ExpectedDigest  string `json:"expected_digest"`
	Content         string `json:"content"`
	CandidateDigest string `json:"candidate_digest"`
}

var actionFields = []string{"schema_version", "action", "path", "expected_digest", "content", "candidate_digest"}
var toolCallIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

func PayloadDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func (a ToolAction) Validate() error {
	if a.SchemaVersion != "development_tool_action.v1" {
		return errors.New("unsupported development tool action")
	}
	if !utf8.ValidString(a.Content) || len(a.Content) > 32768 || strings.ContainsRune(a.Content, 0) {
		return errors.New("invalid tool content")
	}
	if a.Action == "run_tests" {
		if a.Path != "" || a.ExpectedDigest != "" || a.Content != "" || !digestPattern.MatchString(a.CandidateDigest) {
			return errors.New("verification requires an exact candidate")
		}
		return nil
	}
	if a.Path == "" || len(a.Path) > 256 || !utf8.ValidString(a.Path) || path.IsAbs(a.Path) || path.Clean(a.Path) != a.Path || a.Path == "." || a.Path == ".." || strings.HasPrefix(a.Path, "../") || strings.ContainsAny(a.Path, "\\:") || strings.IndexFunc(a.Path, unicode.IsControl) >= 0 {
		return errors.New("tool path must be canonical and relative")
	}
	switch a.Action {
	case "read_file":
		if a.ExpectedDigest != "" || a.Content != "" || a.CandidateDigest != "" {
			return errors.New("read includes unexpected fields")
		}
	case "replace_file":
		if !digestPattern.MatchString(a.ExpectedDigest) || a.CandidateDigest != "" {
			return errors.New("replacement requires current file identity")
		}
	default:
		return errors.New("unknown development tool action")
	}
	return nil
}

func ParseToolAction(data []byte) (ToolAction, error) {
	var a ToolAction
	if err := strictjson.Decode(data, &a); err != nil {
		return a, err
	}
	if _, err := strictjson.ObjectFields(data, actionFields...); err != nil {
		return a, err
	}
	return a, a.Validate()
}

// ParseSDKToolArguments translates only the declared optional SDK summary field.
// The returned summary is untrusted ephemeral display data, never an event or
// authorization field. Path policy, call identity and execution remain separate.
func ParseSDKToolArguments(data []byte) (ToolAction, string, error) {
	var fields map[string]json.RawMessage
	if err := strictjson.Decode(data, &fields); err != nil {
		return ToolAction{}, "", err
	}
	names := append([]string(nil), actionFields...)
	summary := ""
	if raw, exists := fields["summary"]; exists {
		names = append(names, "summary")
		if err := json.Unmarshal(raw, &summary); err != nil || len(summary) > 512 || strings.IndexFunc(summary, unicode.IsControl) >= 0 {
			return ToolAction{}, "", errors.New("invalid SDK display summary")
		}
	}
	if _, err := strictjson.ObjectFields(data, names...); err != nil {
		return ToolAction{}, "", err
	}
	delete(fields, "summary")
	canonical, err := json.Marshal(fields)
	if err != nil {
		return ToolAction{}, "", err
	}
	action, err := ParseToolAction(canonical)
	if err != nil {
		return ToolAction{}, "", err
	}
	return action, summary, nil
}

// ToolCall is the canonical envelope echoed by the SDK tool adapter. The trusted
// Gateway assigns request_id and records the offered action before SDK delivery.
type ToolCall struct {
	SchemaVersion   string     `json:"schema_version"`
	AttemptID       string     `json:"attempt_id"`
	RequestID       string     `json:"request_id"`
	ParentRequestID string     `json:"parent_request_id"`
	ToolCallID      string     `json:"tool_call_id"`
	Arguments       ToolAction `json:"arguments"`
}

func ParseToolCall(data []byte) (ToolCall, error) {
	var call ToolCall
	if err := strictjson.Decode(data, &call); err != nil {
		return call, err
	}
	fields, err := strictjson.ObjectFields(data, "schema_version", "attempt_id", "request_id", "parent_request_id", "tool_call_id", "arguments")
	if err != nil {
		return call, err
	}
	if _, err := ParseToolAction(fields["arguments"]); err != nil {
		return call, err
	}
	return call, call.Validate()
}

func (c ToolCall) Validate() error {
	if c.SchemaVersion != "development_tool_call.v1" || !idPattern.MatchString(c.AttemptID) || !idPattern.MatchString(c.RequestID) || !idPattern.MatchString(c.ParentRequestID) || c.ParentRequestID == c.RequestID || !toolCallIDPattern.MatchString(c.ToolCallID) {
		return errors.New("invalid tool call binding")
	}
	return c.Arguments.Validate()
}

// ValidateToolReservation compares with an immutable offer read by the trusted
// owner, not a second SDK-supplied object. Persistence/atomic consume are not
// provided here and must occur before the future executor receives any action.
func ValidateToolReservation(r Reservation, requested, offered ToolCall) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if err := requested.Validate(); err != nil {
		return err
	}
	if err := offered.Validate(); err != nil {
		return err
	}
	if requested != offered || r.Kind != "tool" || r.AttemptID != requested.AttemptID || r.RequestID != requested.RequestID {
		return errors.New("tool call differs from trusted offer")
	}
	data, err := json.Marshal(requested)
	if err != nil {
		return err
	}
	if r.PayloadDigest != PayloadDigest(data) {
		return errors.New("tool reservation payload mismatch")
	}
	return nil
}
