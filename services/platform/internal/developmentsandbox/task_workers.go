package developmentsandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	control "radishmind.local/services/platform/internal/developmentcontrol"
	probe "radishmind.local/services/platform/internal/developmentprobe"
)

// TaskWorkers is configured only by a trusted controller inside the approved
// SDK environment. Its launcher must isolate the SDK from the owner directory
// and Docker endpoint. No program, image, argument or path comes from HTTP.
type TaskWorkers struct {
	Directory      string
	DockerPath     string
	GoImageDigest  string
	SDKImageDigest string
	RunnerPath     string
	Source         string
	Probe          probe.Plan
	SDKPath        string
	SDKArgs        []string
	ChannelURL     string
}

type taskOwnerFile struct {
	Owner    control.TaskResourceOwner `json:"owner"`
	DaemonID string                    `json:"daemon_id"`
}

var daemonIDPattern = regexp.MustCompile(`^[a-zA-Z0-9:-]{12,128}$`)

// A checked executable cannot be replaced between hashing and execution by the
// unprivileged SDK. This formal profile runs under a root-owned Linux controller.
func CheckControllerExecutable(path string) error {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return ErrSandbox
	}
	for current := path; ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err != nil || !ownedFile(info) || info.Mode().Perm()&0022 != 0 {
			return ErrSandbox
		}
		if current == path {
			if !info.Mode().IsRegular() || info.Mode().Perm()&0100 == 0 || !singleLink(info) {
				return ErrSandbox
			}
		} else if !info.IsDir() {
			return ErrSandbox
		}
		if filepath.Dir(current) == current {
			return nil
		}
	}
}

func (h TaskWorkers) taskDirectory(o control.TaskResourceOwner) (string, error) {
	if !filepath.IsAbs(h.Directory) || !filepath.IsAbs(h.DockerPath) || !idPattern.MatchString(o.OwnerInstanceID) || !idPattern.MatchString(o.Sandbox.AttemptID) {
		return "", ErrSandbox
	}
	info, err := os.Lstat(h.Directory)
	if err != nil || !info.IsDir() || !ownedFile(info) || info.Mode().Perm() != 0700 {
		return "", ErrSandbox
	}
	return filepath.Join(h.Directory, o.Sandbox.AttemptID), nil
}

func (h TaskWorkers) daemon(ctx context.Context) (string, error) {
	s := Session{config: Config{DockerPath: h.DockerPath}}
	raw, err := s.docker(ctx, "info", "--format", "{{.ID}}")
	id := strings.TrimSpace(string(raw))
	if err != nil || !daemonIDPattern.MatchString(id) {
		return "", ErrUnknown
	}
	return id, nil
}

// CheckDaemon binds the launcher's read-only preflight to the code executor's daemon.
func (h TaskWorkers) CheckDaemon(ctx context.Context, expected []byte) error {
	id, err := h.daemon(ctx)
	if err != nil || id != string(expected) {
		return ErrUnknown
	}
	return nil
}

func (h TaskWorkers) Prepare(ctx context.Context, o control.TaskResourceOwner, p control.TaskPlan, d control.Delegation) (control.WorkerLaunch, func() error, error) {
	var none control.WorkerLaunch
	dir, err := h.taskDirectory(o)
	if err != nil {
		return none, nil, err
	}
	if o.GoImageDigest != p.GoImageDigest || o.SDKImageDigest != p.SDKImageDigest || h.SDKImageDigest != o.SDKImageDigest || h.GoImageDigest != o.GoImageDigest || o.ExecutorDigest != p.ExecutorDigest ||
		o.Sandbox != control.SandboxBindingFor(d) || o.TaskID != d.Scope.TaskID || o.TaskID != d.Scope.RunID || o.Scope.ApplicationID != d.Scope.ApplicationID || o.Scope.AccountID != d.Scope.AccountID || o.Scope.WorkspaceID != d.Scope.WorkspaceID ||
		h.Probe.ProbeID != p.SourceProfileID || h.Probe.TaskID != p.SourceVersion || p.ExecutionMode != "fixed_response_only" || !filepath.IsAbs(h.SDKPath) {
		return none, nil, ErrSandbox
	}
	messages := []control.NativeMessage{{Role: "system", Content: "Use only the approved development tool."}, {Role: "user", Content: "Repair the frozen list filter fixture."}}
	raw, err := json.Marshal(messages)
	if err != nil || control.PayloadDigest(raw) != p.AttemptPlan.InitialMessagesDigest {
		return none, nil, ErrSandbox
	}
	daemon, err := h.daemon(ctx)
	if err != nil {
		return none, nil, err
	}
	session, err := New(Config{DockerPath: h.DockerPath, Image: o.GoImageDigest, RunnerPath: h.RunnerPath, RunnerDigest: o.ExecutorDigest, Source: h.Source, Directory: dir, Probe: h.Probe, Plan: p.AttemptPlan, Delegation: d})
	if err != nil {
		return none, nil, err
	}
	if err = atomicJSON(filepath.Join(dir, "task-owner.json"), taskOwnerFile{o, daemon}); err != nil {
		return none, nil, err
	}
	workerDir, err := os.MkdirTemp("", "radishmind-task-sdk-")
	if err != nil {
		return none, nil, ErrSandbox
	}
	launch := control.WorkerLaunch{Command: control.WorkerCommand{Path: h.SDKPath, Args: slices.Clone(h.SDKArgs), Dir: workerDir}, InitialMessages: messages, ChannelURL: h.ChannelURL, Sandbox: session}
	return launch, func() error { return os.RemoveAll(workerDir) }, nil
}

// RecoverCode verifies the workbench's original owner file, daemon and sandbox
// ledger before cleanup may inspect or remove any recorded container.
func (h TaskWorkers) RecoverCode(ctx context.Context, o control.TaskResourceOwner) (bool, error) {
	dir, err := h.taskDirectory(o)
	if err != nil {
		return false, err
	}
	info, err := os.Lstat(dir)
	if err != nil || !info.IsDir() || !ownedFile(info) || info.Mode().Perm() != 0700 {
		return false, ErrUnknown
	}
	path := filepath.Join(dir, "task-owner.json")
	info, err = os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || !singleLink(info) || !ownedFile(info) || info.Mode().Perm() != 0600 || info.Size() > 4096 {
		return false, ErrUnknown
	}
	f, err := os.Open(path)
	if err != nil {
		return false, ErrUnknown
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil || !os.SameFile(info, opened) {
		return false, ErrUnknown
	}
	var owner taskOwnerFile
	raw, err := io.ReadAll(io.LimitReader(f, 4097))
	if err != nil || json.Unmarshal(raw, &owner) != nil || owner.Owner != o || !daemonIDPattern.MatchString(owner.DaemonID) {
		return false, ErrUnknown
	}
	expected, err := json.MarshalIndent(owner, "", "  ")
	if err != nil || !bytes.Equal(raw, append(expected, '\n')) {
		return false, ErrUnknown
	}
	daemon, err := h.daemon(ctx)
	if err != nil || daemon != owner.DaemonID {
		return false, ErrUnknown
	}
	return cleanupRecovered(ctx, h.DockerPath, dir, &o)
}
