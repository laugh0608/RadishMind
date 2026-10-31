// Package developmentprobe defines offline preparation for the bounded SDK probe.
// It does not grant platform identity permissions or execute tools and models.
package developmentprobe

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
)

const PlanSchema = "development_task_probe_plan.v1"

type SourceFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type SDKCandidate struct {
	Package string `json:"package"`
	Version string `json:"version"`
}

type Limits struct {
	MaxModelRequests      int `json:"max_model_requests"`
	MaxToolCalls          int `json:"max_tool_calls"`
	MaxRequestBytes       int `json:"max_request_bytes"`
	MaxOutputBytes        int `json:"max_output_bytes"`
	CommandTimeoutSeconds int `json:"command_timeout_seconds"`
	WallTimeoutSeconds    int `json:"wall_timeout_seconds"`
}

type Plan struct {
	SchemaVersion string       `json:"schema_version"`
	ProbeID       string       `json:"probe_id"`
	Mode          string       `json:"mode"`
	TaskID        string       `json:"task_id"`
	SDKCandidate  SDKCandidate `json:"sdk_candidate"`
	SourceFiles   []SourceFile `json:"source_files"`
	WritablePaths []string     `json:"writable_paths"`
	VerifyArgv    []string     `json:"verify_argv"`
	Limits        Limits       `json:"limits"`
}

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var probeIDPattern = regexp.MustCompile(`^probe_[a-z0-9_]{1,48}$`)
var sdkVersionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

func ParsePlan(data []byte) (Plan, error) {
	var plan Plan
	if err := DecodeStrict(data, &plan); err != nil {
		return Plan{}, fmt.Errorf("invalid probe plan: %w", err)
	}
	fields, err := exactKeys(data, "schema_version", "probe_id", "mode", "task_id", "sdk_candidate", "source_files", "writable_paths", "verify_argv", "limits")
	if err != nil {
		return Plan{}, err
	}
	if _, err := exactKeys(fields["sdk_candidate"], "package", "version"); err != nil {
		return Plan{}, err
	}
	if _, err := exactKeys(fields["limits"], "max_model_requests", "max_tool_calls", "max_request_bytes", "max_output_bytes", "command_timeout_seconds", "wall_timeout_seconds"); err != nil {
		return Plan{}, err
	}
	var files []json.RawMessage
	if err := json.Unmarshal(fields["source_files"], &files); err != nil {
		return Plan{}, err
	}
	for _, file := range files {
		if _, err := exactKeys(file, "path", "sha256"); err != nil {
			return Plan{}, err
		}
	}
	if err := plan.Validate(); err != nil {
		return Plan{}, err
	}
	return plan, nil
}

func (p Plan) Validate() error {
	if p.SchemaVersion != PlanSchema || !probeIDPattern.MatchString(p.ProbeID) || p.Mode != "fixed_response_only" || p.TaskID != "task_list_filter_v1" {
		return errors.New("unsupported probe identity or mode")
	}
	if p.SDKCandidate.Package != "openhands-sdk" || !sdkVersionPattern.MatchString(p.SDKCandidate.Version) {
		return errors.New("probe requires an exact SDK candidate version")
	}
	paths := []string{"go.mod", "list.go", "list_test.go"}
	if len(p.SourceFiles) != len(paths) {
		return errors.New("probe requires exactly the three task source files")
	}
	for index, file := range p.SourceFiles {
		if file.Path != paths[index] || !digestPattern.MatchString(file.SHA256) {
			return errors.New("probe source paths, order or digest invalid")
		}
	}
	if len(p.WritablePaths) != 1 || p.WritablePaths[0] != "list.go" {
		return errors.New("only list.go may be changed by the probe")
	}
	command := []string{"go", "test", "-json", "-count=1", "./..."}
	if len(p.VerifyArgv) != len(command) {
		return errors.New("probe verification command is fixed")
	}
	for index, argument := range command {
		if p.VerifyArgv[index] != argument {
			return errors.New("probe verification command is fixed")
		}
	}
	l := p.Limits
	if l.MaxModelRequests < 1 || l.MaxModelRequests > 8 || l.MaxToolCalls < 1 || l.MaxToolCalls > 8 ||
		l.MaxRequestBytes < 1024 || l.MaxRequestBytes > 262144 || l.MaxOutputBytes < 1024 || l.MaxOutputBytes > 65536 ||
		l.CommandTimeoutSeconds < 1 || l.CommandTimeoutSeconds > 120 || l.WallTimeoutSeconds < l.CommandTimeoutSeconds || l.WallTimeoutSeconds > 600 {
		return errors.New("probe limits exceed the bounded validation profile")
	}
	return nil
}

func Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func PlanDigest(plan Plan) string {
	// Plan contains concrete JSON fields only; validation precedes this call.
	data, _ := json.Marshal(plan)
	return Digest(data)
}

func exactKeys(data []byte, names ...string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || len(fields) != len(names) {
		return nil, errors.New("probe fields must match the canonical contract exactly")
	}
	for _, name := range names {
		if _, ok := fields[name]; !ok {
			return nil, errors.New("probe field is missing or incorrectly cased")
		}
	}
	return fields, nil
}

// DecodeStrict rejects duplicate keys, nulls, unknown fields and trailing values.
// All v1 probe inputs use concrete values, so null is never an implicit default.
func DecodeStrict(data []byte, result any) error {
	if len(data) > 262144 {
		return errors.New("probe JSON exceeds size limit")
	}
	tokens := json.NewDecoder(bytes.NewReader(data))
	if err := inspectJSONValue(tokens); err != nil {
		return err
	}
	if _, err := tokens.Token(); err != io.EOF {
		return errors.New("probe JSON has trailing content")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(result)
}

func inspectJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if token == nil {
		return errors.New("null is not allowed in probe JSON")
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	if delimiter != '{' && delimiter != '[' {
		return errors.New("unexpected JSON delimiter")
	}
	seen := map[string]bool{}
	for decoder.More() {
		if delimiter == '{' {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || seen[name] {
				return errors.New("duplicate or invalid probe JSON key")
			}
			seen[name] = true
		}
		if err := inspectJSONValue(decoder); err != nil {
			return err
		}
	}
	_, err = decoder.Token()
	return err
}
