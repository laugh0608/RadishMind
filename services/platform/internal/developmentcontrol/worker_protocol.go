package developmentcontrol

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"path/filepath"
	"slices"
	"strconv"

	"radishmind.local/services/platform/internal/strictjson"
)

var ErrWorker = errors.New("development_sdk_worker_failed")

// WorkerCommand is configured by the trusted host, never read from an SDK or
// model request. Dir must be a fresh private directory owned by this launch.
type WorkerCommand struct {
	Path string
	Args []string
	Dir  string
}
type WorkerInput struct {
	SchemaVersion      string            `json:"schema_version"`
	AttemptID          string            `json:"attempt_id"`
	ExecutorInstanceID string            `json:"executor_instance_id"`
	ChannelURL         string            `json:"channel_url"`
	Credential         string            `json:"credential"`
	Model              string            `json:"model"`
	InitialMessages    []NativeMessage   `json:"initial_messages"`
	Tools              []json.RawMessage `json:"tools"`
	MaxOutputTokens    int64             `json:"max_output_tokens"`
	MaxModelCalls      int64             `json:"max_model_calls"`
}
type workerFrame struct {
	SchemaVersion string    `json:"schema_version"`
	Sequence      int64     `json:"sequence"`
	Kind          string    `json:"kind"`
	EventType     string    `json:"event_type"`
	RequestID     string    `json:"request_id"`
	Call          *ToolCall `json:"call,omitempty"`
	ErrorCode     string    `json:"error_code,omitempty"`
}
type workerReply struct {
	SchemaVersion string `json:"schema_version"`
	Sequence      int64  `json:"sequence"`
	RequestID     string `json:"request_id"`
	Observation   string `json:"observation"`
}
type WorkerResult struct {
	ProcessExited bool
	GroupEmpty    bool
	Finished      bool
	Frames        int64
}

// ToolHandler runs only in the trusted Go host. No filesystem / command
// implementation is provided here. A nil handler cannot manufacture a result.
type ToolHandler func(context.Context, ToolCall) (string, error)

func (command WorkerCommand) validate() error {
	if !filepath.IsAbs(command.Path) || !filepath.IsAbs(command.Dir) || len(command.Args) > 16 {
		return ErrWorker
	}
	for _, arg := range command.Args {
		if len(arg) > 4096 {
			return ErrWorker
		}
	}
	return nil
}
func validWorkerURL(value string) bool {
	u, err := url.Parse(value)
	if err != nil || u.Scheme != "http" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "" {
		return false
	}
	host, port, err := net.SplitHostPort(u.Host)
	number, portErr := strconv.Atoi(port)
	return err == nil && portErr == nil && number > 0 && number <= 65535 && net.ParseIP(host) != nil && net.ParseIP(host).IsLoopback()
}
func parseWorkerFrame(data []byte, sequence int64) (workerFrame, error) {
	var f workerFrame
	if err := strictjson.Decode(data, &f); err != nil {
		return f, ErrWorker
	}
	names := []string{"schema_version", "sequence", "kind", "event_type", "request_id"}
	if f.Kind == "event" && f.EventType == "error" {
		names = append(names, "error_code")
	}
	if f.Kind == "tool" {
		names = append(names, "call")
	}
	fields, err := strictjson.ObjectFields(data, names...)
	if err != nil || f.SchemaVersion != "development_worker_message.v1" || f.Sequence != sequence {
		return f, ErrWorker
	}
	switch f.Kind {
	case "failed":
		if !slices.Contains([]string{"input", "agent_setup", "conversation_setup", "run", "close"}, f.EventType) || f.RequestID != "" {
			return f, ErrWorker
		}
	case "ready", "finished":
		if f.EventType != "" || f.RequestID != "" {
			return f, ErrWorker
		}
	case "event":
		if f.EventType == "error" && !slices.Contains([]string{"sdk_runtime", "generate_options", "message_serialization", "model_channel", "response_conversion", "tool_exchange", "sdk_type_error", "sdk_attribute_error", "sdk_validation_error"}, f.ErrorCode) {
			return f, ErrWorker
		}
		if _, ok := sdkEventTypes[f.EventType]; !ok || f.RequestID != "" && !idPattern.MatchString(f.RequestID) {
			return f, ErrWorker
		}
	case "tool":
		call, err := ParseToolCall(fields["call"])
		if err != nil || f.EventType != "" || f.RequestID != call.RequestID {
			return f, ErrWorker
		}
		f.Call = &call
	default:
		return f, ErrWorker
	}
	return f, nil
}

var sdkEventTypes = map[string]string{
	"system_prompt": "sdk_system_prompt_observed", "user_message": "sdk_user_message_observed",
	"assistant_message": "sdk_assistant_message_observed", "action": "sdk_action_observed",
	"observation": "sdk_observation_observed", "error": "sdk_error_observed", "security": "sdk_security_observed",
}
