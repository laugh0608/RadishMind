package developmentcontrol

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrTaskInvalid  = errors.New("development_task_invalid")
	ErrTaskCapacity = errors.New("development_task_capacity_exceeded")
	ErrTaskExpired  = errors.New("development_task_preparation_expired")
	ErrTaskCorrupt  = errors.New("development_task_corrupt")
	ErrSlotBusy     = errors.New("development_task_execution_slot_busy")
)

const taskSource = "probe_list_filter"

type TaskScope struct {
	Identity
	ApplicationID string `json:"application_id"`
}

func (s TaskScope) valid() bool {
	return s.TenantRef != "" && len(s.TenantRef) <= 128 && utf8.ValidString(s.TenantRef) && !strings.ContainsAny(s.TenantRef, "\x00\r\n\t") && idPattern.MatchString(s.AccountID) && idPattern.MatchString(s.WorkspaceID) && idPattern.MatchString(s.ApplicationID)
}

type TaskChoice struct {
	ModelRef string `json:"model_ref"`
	Limits   Limits `json:"limits"`
}

func (c TaskChoice) valid() bool {
	l := c.Limits
	return modelRefPattern.MatchString(c.ModelRef) && l.MaxModelCalls >= 1 && l.MaxModelCalls <= 4 && l.MaxToolCalls >= 1 && l.MaxToolCalls <= 3 && l.MaxTokens >= 1 && l.MaxTokens <= 262144 && l.MaxCostMicrousd >= 1 && l.MaxCostMicrousd <= 1000000000
}

type TaskProfile struct {
	AttemptPlan            Plan   `json:"attempt_plan"`
	SourceProfileID        string `json:"source_profile_id"`
	SourceVersion          string `json:"source_version"`
	ExecutionMode          string `json:"execution_mode"`
	SDKImageDigest         string `json:"sdk_image_digest"`
	GoImageDigest          string `json:"go_image_digest"`
	ExecutorDigest         string `json:"executor_digest"`
	RetentionPolicyVersion string `json:"retention_policy_version"`
}

func (p TaskProfile) valid(scope TaskScope, choice TaskChoice) bool {
	a := p.AttemptPlan
	return choice.valid() && a.Validate() == nil && a.ApplicationID == scope.ApplicationID && a.ModelRef == choice.ModelRef && a.Limits == choice.Limits &&
		p.SourceProfileID == taskSource && p.ExecutionMode == "fixed_response_only" && idPattern.MatchString(p.SourceVersion) && idPattern.MatchString(p.RetentionPolicyVersion) &&
		digestPattern.MatchString(p.SDKImageDigest) && digestPattern.MatchString(p.GoImageDigest) && digestPattern.MatchString(p.ExecutorDigest) &&
		slices.Equal(a.Tools.ReadPaths, []string{"go.mod", "list.go", "list_test.go"}) && slices.Equal(a.Tools.WritePaths, []string{"list.go"})
}

type TaskOperation string

const (
	TaskRead         TaskOperation = "read"
	TaskCreate       TaskOperation = "create_task"
	TaskRevise       TaskOperation = "revise_plan"
	TaskPrepare      TaskOperation = "prepare"
	TaskStart        TaskOperation = "start"
	TaskCancel       TaskOperation = "cancel"
	TaskSaveArtifact TaskOperation = "save_artifact"
	TaskReadArtifact TaskOperation = "read_artifact"
)

// TaskHost reads current session, membership, exact permissions and application
// state. FixedProfile resolves only a configured fixed-response model and returns
// the earliest expiry of the current qualifications. Both are read-only and may
// run while the coordinator holds the SQLite write transaction.
type TaskHost interface {
	AuthorizeTask(context.Context, TaskScope, TaskOperation) error
	FixedProfile(context.Context, TaskScope, TaskChoice) (TaskProfile, time.Time, error)
}

// TaskMetadata owns workbench planning and history. It has no runtime,
// credential issuer or worker to activate from a metadata request.
type TaskMetadata struct {
	store *SQLiteStore
	host  TaskHost
	now   func() time.Time
}

type TaskCoordinator struct {
	*TaskMetadata
	runtime *Runtime
}

func NewTaskMetadata(store *SQLiteStore, h TaskHost) (*TaskMetadata, error) {
	if store == nil || h == nil {
		return nil, ErrUnavailable
	}
	var name string
	if err := store.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='development_tasks'").Scan(&name); err != nil {
		return nil, ErrUnavailable
	}
	return &TaskMetadata{store: store, host: h, now: time.Now}, nil
}

func NewTaskCoordinator(r *Runtime, h TaskHost) (*TaskCoordinator, error) {
	if r == nil {
		return nil, ErrUnavailable
	}
	metadata, err := NewTaskMetadata(r.store, h)
	if err != nil {
		return nil, err
	}
	metadata.now = func() time.Time { return r.now() }
	return &TaskCoordinator{metadata, r}, nil
}

type Task struct {
	TaskID          string `json:"task_id"`
	SourceProfileID string `json:"source_profile_id"`
	Version         int64  `json:"record_version"`
	CreatedAtMS     int64  `json:"created_at_ms"`
	UpdatedAtMS     int64  `json:"updated_at_ms"`
	Revision        int    `json:"plan_revision"`
	AttemptID       string `json:"attempt_id"`
}

type TaskPlan struct {
	Revision int    `json:"plan_revision"`
	Digest   string `json:"plan_digest"`
	TaskProfile
}

type TaskPreparation struct {
	PreparationID      string   `json:"preparation_id"`
	TaskID             string   `json:"task_id"`
	TaskVersion        int64    `json:"task_version"`
	Revision           int      `json:"plan_revision"`
	ConfirmationDigest string   `json:"confirmation_digest"`
	PreparedAtMS       int64    `json:"prepared_at_ms"`
	ExpiresAtMS        int64    `json:"expires_at_ms"`
	Plan               TaskPlan `json:"plan"`
	EligibilityDigest  string   `json:"eligibility_digest"`
}

type TaskReceipt struct {
	RequestID     string        `json:"request_id"`
	Operation     TaskOperation `json:"operation"`
	RequestDigest string        `json:"request_digest"`
	TaskID        string        `json:"task_id"`
	AttemptID     string        `json:"attempt_id,omitempty"`
	ArtifactID    string        `json:"artifact_id,omitempty"`
	// Private response locators reproduce the original revision or preparation.
	Revision      int    `json:"-"`
	PreparationID string `json:"-"`
}

func taskDigest(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", ErrTaskInvalid
	}
	return PayloadDigest(data), nil
}
