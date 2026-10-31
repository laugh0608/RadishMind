// Package developmentartifact owns explicit, immutable delivery packages for
// the frozen Go probe. It neither executes code nor accepts SDK-authored evidence.
package developmentartifact

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	control "radishmind.local/services/platform/internal/developmentcontrol"
	probe "radishmind.local/services/platform/internal/developmentprobe"
)

const (
	Schema            = "development_artifact.v1"
	ExportSchema      = "development_artifact_export.v1"
	MaxPackageBytes   = 1 << 20
	MaxTextBytes      = 64 << 10
	MaxSavedCount     = 64
	MaxSavedBytes     = 64 << 20
	TemporaryLifetime = 24 * time.Hour
	PermissionRead    = "development_artifacts:read"
	PermissionSave    = "development_artifacts:save"
	PermissionExport  = "development_artifacts:export"
)

var (
	ErrInvalid     = errors.New("development_artifact_invalid")
	ErrDenied      = errors.New("development_artifact_denied")
	ErrSensitive   = errors.New("development_artifact_sensitive")
	ErrExpired     = errors.New("development_artifact_source_expired")
	ErrConflict    = errors.New("development_artifact_conflict")
	ErrQuota       = errors.New("development_artifact_quota_exceeded")
	ErrUnavailable = errors.New("development_artifact_store_unavailable")
	ErrCorrupt     = errors.New("development_artifact_corrupt")
	ErrNotFound    = errors.New("development_artifact_not_found")
	idPattern      = regexp.MustCompile(`^[a-z][a-z0-9_-]{2,63}$`)
	digestPattern  = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
)

type Scope struct {
	TenantRef     string `json:"tenant_ref"`
	WorkspaceID   string `json:"workspace_id"`
	AccountID     string `json:"account_id"`
	ApplicationID string `json:"application_id"`
}
type File struct {
	Path            string `json:"path"`
	Base            string `json:"base"`
	BaseDigest      string `json:"base_digest"`
	Candidate       string `json:"candidate"`
	CandidateDigest string `json:"candidate_digest"`
}
type Test struct {
	RequestID       string   `json:"request_id"`
	CandidateDigest string   `json:"candidate_digest"`
	Profile         string   `json:"profile"`
	Argv            []string `json:"argv"`
	ImageDigest     string   `json:"image_digest"`
	RunnerDigest    string   `json:"runner_digest"`
	StartedAt       string   `json:"started_at"`
	FinishedAt      string   `json:"finished_at"`
	ExitKnown       bool     `json:"exit_known"`
	ExitCode        int      `json:"exit_code"`
	OOMKilled       bool     `json:"oom_killed"`
	OutputTruncated bool     `json:"output_truncated"`
	Log             string   `json:"log"`
	LogDigest       string   `json:"log_digest"`
	LogRedacted     bool     `json:"log_redacted"`
	Status          string   `json:"status"`
}
type Bundle struct {
	SourceCreatedAt   string        `json:"source_created_at"`
	SchemaVersion     string        `json:"schema_version"`
	Scope             Scope         `json:"scope"`
	Origin            control.Scope `json:"origin"`
	Plan              control.Plan  `json:"plan"`
	Probe             probe.Plan    `json:"probe"`
	SourceKind        string        `json:"source_kind"`
	AttemptState      string        `json:"attempt_state"`
	SnapshotAt        string        `json:"snapshot_at"`
	ExpiresAt         string        `json:"expires_at"`
	Files             []File        `json:"files"`
	CandidateDigest   string        `json:"candidate_digest"`
	Patch             string        `json:"patch"`
	PatchDigest       string        `json:"patch_digest"`
	Tests             []Test        `json:"tests"`
	VerificationState string        `json:"verification_state"`
	ReviewState       string        `json:"review_state"`
	AcceptanceState   string        `json:"acceptance_state"`
	SensitivePolicy   string        `json:"sensitive_policy"`
	Digest            string        `json:"digest"`
}

// Input is supplied only by a trusted host snapshot reader after writes stop.
// SourceCreatedAt fixes retention before execution; SnapshotAt is fixed when
// writes stop. Neither timestamp is refreshed on retry.
type Input struct {
	SourceCreatedAt time.Time
	Scope           Scope
	Origin          control.Scope
	Plan            control.Plan
	Probe           probe.Plan
	AttemptState    string
	SnapshotAt      time.Time
	Files           []File
	Tests           []Test
}
type Record struct {
	SchemaVersion string `json:"schema_version"`
	ArtifactID    string `json:"artifact_id"`
	RequestID     string `json:"request_id"`
	SavedAt       string `json:"saved_at"`
	Bundle        Bundle `json:"bundle"`
}
type Export struct {
	SchemaVersion string `json:"schema_version"`
	Record        Record `json:"record"`
	ExportedAt    string `json:"exported_at"`
	RequestID     string `json:"request_id"`
	Digest        string `json:"digest"`
}

// Authorizer rereads current account, membership, permission and application.
// Scope.AccountID is the verified actor; task credentials are never identities.
type Authorizer interface {
	Authorize(context.Context, Scope, string) error
}
type SourceReader interface {
	Snapshot(context.Context, Scope, string) (Input, error)
}

func (s Scope) Validate() error {
	if len(s.TenantRef) < 1 || len(s.TenantRef) > 128 || !utf8.ValidString(s.TenantRef) || strings.ContainsAny(s.TenantRef, "\x00\r\n\t") || sensitive(s.TenantRef) {
		return ErrInvalid
	}
	for _, id := range []string{s.AccountID, s.WorkspaceID, s.ApplicationID} {
		if !idPattern.MatchString(id) {
			return ErrInvalid
		}
	}
	return nil
}
