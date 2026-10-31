// Package developmentsandbox owns bounded code containers for the frozen Go
// probe. It has no model credentials, public API or default activation.
package developmentsandbox

import (
	"context"
	"errors"
	"regexp"
	"time"

	control "radishmind.local/services/platform/internal/developmentcontrol"
	probe "radishmind.local/services/platform/internal/developmentprobe"
)

var (
	ErrSandbox       = errors.New("development_sandbox_denied")
	ErrUnknown       = errors.New("development_sandbox_termination_unconfirmed")
	idPattern        = regexp.MustCompile(`^[a-z][a-z0-9_-]{2,63}$`)
	imagePattern     = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	sessionPattern   = regexp.MustCompile(`^sandbox_[0-9a-f]{32}$`)
	containerPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

type Config struct {
	DockerPath   string
	Image        string
	RunnerPath   string
	RunnerDigest string
	Source       string
	Directory    string // Must not exist; contains private copies and the resource ledger.
	Probe        probe.Plan
	Plan         control.Plan
	Delegation   control.Delegation
}

type resource struct {
	RequestID       string `json:"request_id"`
	Name            string `json:"name"`
	ContainerID     string `json:"container_id"`
	CandidateDigest string `json:"candidate_digest"`
	State           string `json:"state"`
}

type ledger struct {
	SchemaVersion string                 `json:"schema_version"`
	SessionID     string                 `json:"session_id"`
	Binding       control.SandboxBinding `json:"binding"`
	Image         string                 `json:"image"`
	RunnerDigest  string                 `json:"runner_digest"`
	Resources     []resource             `json:"resources"`
	Closed        bool                   `json:"closed"`
}

type Verification struct {
	Argv            []string `json:"argv"`
	StartedAt       string   `json:"started_at"`
	FinishedAt      string   `json:"finished_at"`
	DurationMS      int64    `json:"duration_ms"`
	Profile         string   `json:"profile"`
	Image           string   `json:"image"`
	RunnerDigest    string   `json:"runner_digest"`
	ExitCode        int      `json:"exit_code"`
	OOMKilled       bool     `json:"oom_killed"`
	OutputTruncated bool     `json:"output_truncated"`
	Output          string   `json:"output"`
	Status          string   `json:"status"` // command_passed is not a code quality or acceptance claim.
}

type Observation struct {
	SchemaVersion   string        `json:"schema_version"`
	Action          string        `json:"action"`
	Path            string        `json:"path"`
	FileDigest      string        `json:"file_digest"`
	CandidateDigest string        `json:"candidate_digest"`
	Content         string        `json:"content"`
	Verification    *Verification `json:"verification,omitempty"`
}

const commandTimeout = 125 * time.Second

// Container commands always use a fresh bounded context; a cancelled SDK/model
// context must never prevent the cleanup command from reaching Docker.
func cleanupContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
