package developmentprobe

import (
	"errors"
	"strings"
	"unicode/utf8"

	"radishmind.local/services/platform/internal/strictjson"
)

// ToolRequest is a validation-probe input, not a production tool authorization.
// Parsing bounds the sequence; the future executor must enforce actual ordering.
type ToolRequest struct {
	SchemaVersion   string `json:"schema_version"`
	ProbeID         string `json:"probe_id"`
	Sequence        int    `json:"sequence"`
	Action          string `json:"action"`
	Path            string `json:"path"`
	ExpectedDigest  string `json:"expected_digest"`
	Content         string `json:"content"`
	CandidateDigest string `json:"candidate_digest"`
}

func ParseToolRequest(plan Plan, data []byte) (ToolRequest, error) {
	if err := plan.Validate(); err != nil {
		return ToolRequest{}, err
	}
	if len(data) > plan.Limits.MaxRequestBytes || !utf8.Valid(data) {
		return ToolRequest{}, errors.New("probe tool request is too large or not UTF-8")
	}
	var request ToolRequest
	if err := strictjson.Decode(data, &request); err != nil {
		return ToolRequest{}, err
	}
	if _, err := strictjson.ObjectFields(data, "schema_version", "probe_id", "sequence", "action", "path", "expected_digest", "content", "candidate_digest"); err != nil {
		return ToolRequest{}, err
	}
	if request.SchemaVersion != "development_task_probe_tool_request.v1" || request.ProbeID != plan.ProbeID || request.Sequence < 1 || request.Sequence > plan.Limits.MaxToolCalls {
		return ToolRequest{}, errors.New("probe tool request scope or sequence invalid")
	}
	switch request.Action {
	case "read_file":
		if request.Path != "list.go" || request.ExpectedDigest != "" || request.Content != "" || request.CandidateDigest != "" {
			return ToolRequest{}, errors.New("probe read only accepts list.go")
		}
	case "replace_file":
		if request.Path != "list.go" || !digestPattern.MatchString(request.ExpectedDigest) || request.Content == "" || len(request.Content) > maxSourceBytes || strings.ContainsRune(request.Content, 0) || request.CandidateDigest != "" {
			return ToolRequest{}, errors.New("probe replacement requires exact file identity and bounded text")
		}
	case "run_tests":
		if !digestPattern.MatchString(request.CandidateDigest) || request.Path != "" || request.ExpectedDigest != "" || request.Content != "" {
			return ToolRequest{}, errors.New("probe verification requires exact candidate identity")
		}
	default:
		return ToolRequest{}, errors.New("probe tool action is not allowed")
	}
	return request, nil
}
