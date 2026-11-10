package developmentsandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	control "radishmind.local/services/platform/internal/developmentcontrol"
	probe "radishmind.local/services/platform/internal/developmentprobe"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var errOutputLimit = errors.New("development_sandbox_output_limit")

const ownerLabel = "local.radishmind.development.sandbox"
const attemptLabel = "local.radishmind.development.attempt"

type cappedBuffer struct {
	bytes.Buffer
	exceeded bool
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	n := len(p)
	if b.Len()+n > 65536 {
		b.exceeded = true
		p = p[:65536-b.Len()]
	}
	_, err := b.Buffer.Write(p)
	return n, err
}
func (s *Session) docker(ctx context.Context, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, s.config.DockerPath, args...)
	command.Stdin = nil
	command.Stderr = io.Discard
	command.WaitDelay = time.Second
	var output cappedBuffer
	command.Stdout = &output
	err := command.Run()
	if output.exceeded {
		return output.Bytes(), errOutputLimit
	}
	return output.Bytes(), err
}
func (s *Session) createArgs(name, snapshot string) []string {
	return []string{"container", "create", "--pull=never", "--name", name, "--label", ownerLabel + "=" + s.ledger.SessionID, "--label", attemptLabel + "=" + s.Binding().AttemptID,
		"--network", "none", "--read-only", "--cap-drop", "ALL", "--security-opt", "no-new-privileges", "--user", "65532:65532", "--memory", "1g", "--memory-swap", "1g", "--cpus", "2", "--pids-limit", "128", "--ulimit", "nofile=256:256", "--restart", "no", "--stop-timeout", "1", "--log-driver", "none",
		"--tmpfs", "/tmp:rw,exec,size=512m,mode=1777,nosuid,nodev", "--workdir", "/workspace",
		"--mount", "type=bind,src=" + snapshot + ",dst=/workspace,readonly",
		"--mount", "type=bind,src=" + s.config.RunnerPath + ",dst=/opt/code-runner,readonly",
		"--entrypoint", "/opt/code-runner", s.config.Image, strconv.Itoa(s.config.Probe.Limits.CommandTimeoutSeconds)}
}

type inspection struct {
	ID    string `json:"Id"`
	Name  string
	Image string
	State struct {
		Running    bool
		ExitCode   int
		OOMKilled  bool
		StartedAt  string
		FinishedAt string
	}
	Config struct {
		User       string
		Labels     map[string]string
		Entrypoint []string
		Cmd        []string
		WorkingDir string
	}
	HostConfig struct {
		NetworkMode    string
		ReadonlyRootfs bool
		Memory         int64
		MemorySwap     int64
		NanoCpus       int64
		PidsLimit      int64
		CapDrop        []string
		SecurityOpt    []string
		Privileged     bool
		PidMode        string
		IpcMode        string
		AutoRemove     bool
		Tmpfs          map[string]string
		PortBindings   map[string]any
		RestartPolicy  struct{ Name string }
	}
	Mounts []struct {
		Type        string
		Source      string
		Destination string
		RW          bool
	}
}

func (s *Session) inspectContainer(ctx context.Context, id string) (inspection, error) {
	var result []inspection
	raw, err := s.docker(ctx, "container", "inspect", id)
	if err != nil || json.Unmarshal(raw, &result) != nil || len(result) != 1 {
		return inspection{}, ErrUnknown
	}
	item := result[0]
	if !containerPattern.MatchString(item.ID) || item.Image != s.config.Image || item.Config.Labels[ownerLabel] != s.ledger.SessionID || item.Config.Labels[attemptLabel] != s.Binding().AttemptID {
		return inspection{}, ErrUnknown
	}
	return item, nil
}
func (s *Session) verifyContainer(item inspection, snapshot, name string) error {
	h := item.HostConfig
	if item.Name != "/"+name || item.Config.User != "65532:65532" || !slices.Equal(item.Config.Entrypoint, []string{"/opt/code-runner"}) || !slices.Equal(item.Config.Cmd, []string{strconv.Itoa(s.config.Probe.Limits.CommandTimeoutSeconds)}) || item.Config.WorkingDir != "/workspace" || h.NetworkMode != "none" || !h.ReadonlyRootfs || h.Memory != 1<<30 || h.MemorySwap != 1<<30 || h.NanoCpus != 2_000_000_000 || h.PidsLimit != 128 || !slices.Equal(h.CapDrop, []string{"ALL"}) || !slices.Contains(h.SecurityOpt, "no-new-privileges") || h.Privileged || h.PidMode != "" || h.IpcMode == "host" || h.AutoRemove || h.RestartPolicy.Name != "no" || len(h.PortBindings) != 0 || len(h.Tmpfs) != 1 || h.Tmpfs["/tmp"] != "rw,exec,size=512m,mode=1777,nosuid,nodev" || len(item.Mounts) != 2 {
		return ErrSandbox
	}
	mounts := map[string]string{"/workspace": snapshot, "/opt/code-runner": s.config.RunnerPath}
	for _, mount := range item.Mounts {
		if mount.Type != "bind" || mount.RW || mounts[mount.Destination] != mount.Source {
			return ErrSandbox
		}
		delete(mounts, mount.Destination)
	}
	if len(mounts) != 0 {
		return ErrSandbox
	}
	return nil
}
func (s *Session) owned(ctx context.Context) ([]string, error) {
	raw, err := s.docker(ctx, "container", "ls", "--all", "--no-trunc", "--filter", "label="+ownerLabel+"="+s.ledger.SessionID, "--format", "{{.ID}}")
	if err != nil {
		return nil, ErrUnknown
	}
	ids := strings.Fields(string(raw))
	if len(ids) > 32 {
		return nil, ErrUnknown
	}
	for _, id := range ids {
		if !containerPattern.MatchString(id) {
			return nil, ErrUnknown
		}
	}
	return ids, nil
}
func (s *Session) cleanup(ctx context.Context) (bool, error) {
	ids, err := s.owned(ctx)
	if err != nil {
		return false, err
	}
	for _, id := range ids {
		item, err := s.inspectContainer(ctx, id)
		if err != nil {
			return false, err
		}
		matched := -1
		for index, r := range s.ledger.Resources {
			if item.Name == "/"+r.Name && (r.ContainerID == "" || r.ContainerID == id) {
				matched = index
				break
			}
		}
		if matched < 0 {
			return false, ErrUnknown
		}
		// rm -f kills the full container PID namespace immediately, independently
		// of attached stdout, process groups or cooperation from task code.
		if _, err = s.docker(ctx, "container", "rm", "--force", id); err != nil {
			return false, ErrUnknown
		}
		s.ledger.Resources[matched].ContainerID = id
		s.ledger.Resources[matched].State = "removed"
		if err = s.save(); err != nil {
			return false, err
		}
	}
	remaining, err := s.owned(ctx)
	if err != nil || len(remaining) != 0 {
		return false, ErrUnknown
	}
	for _, r := range s.ledger.Resources {
		if r.State == "create_unknown" || r.State == "intent" {
			return false, ErrUnknown
		}
	}
	return true, nil
}
func (s *Session) runTests(ctx context.Context, request, digest string) (verification Verification, returned error) {
	snapshot, err := s.snapshot(request)
	if err != nil {
		return verification, err
	}
	name := "radishmind-code-" + s.ledger.SessionID + "-" + request
	s.ledger.Resources = append(s.ledger.Resources, resource{RequestID: request, Name: name, CandidateDigest: digest, State: "intent"})
	index := len(s.ledger.Resources) - 1
	if err = s.save(); err != nil {
		return verification, err
	}
	defer func() {
		cleanup, cancel := cleanupContext()
		defer cancel()
		empty, err := s.cleanup(cleanup)
		if err != nil || !empty {
			returned = errors.Join(returned, ErrUnknown, err)
		}
	}()
	// Never cancel an in-progress create just because SDK context was cancelled:
	// await its bounded acknowledgement, then clean it without ever starting it.
	createCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	raw, err := s.docker(createCtx, s.createArgs(name, snapshot)...)
	uncertain := createCtx.Err() != nil
	cancel()
	if err != nil {
		if uncertain {
			s.ledger.Resources[index].State = "create_unknown"
			if saveErr := s.save(); saveErr != nil {
				return verification, errors.Join(ErrUnknown, saveErr)
			}
		}
		return verification, ErrUnknown
	}
	id := strings.TrimSpace(string(raw))
	if !containerPattern.MatchString(id) {
		return verification, ErrUnknown
	}
	s.ledger.Resources[index].ContainerID = id
	s.ledger.Resources[index].State = "created"
	if err = s.save(); err != nil {
		return verification, err
	}
	inspectCtx, inspectCancel := context.WithTimeout(context.Background(), 3*time.Second)
	item, err := s.inspectContainer(inspectCtx, id)
	inspectCancel()
	if err != nil || s.verifyContainer(item, snapshot, name) != nil {
		return verification, ErrSandbox
	}
	if err = atomicJSON(filepath.Join(s.config.Directory, request+".container.json"), item); err != nil {
		return verification, err
	}
	if ctx.Err() != nil {
		return verification, ctx.Err()
	}
	s.ledger.Resources[index].State = "dispatched"
	if err = s.save(); err != nil {
		return verification, err
	}
	runCtx, runCancel := context.WithTimeout(ctx, commandTimeout)
	defer runCancel()
	started := time.Now()
	output, runErr := s.docker(runCtx, "container", "start", "--attach", id)
	// Context cancellation is not evidence that the Docker container stopped.
	if runCtx.Err() != nil {
		s.ledger.Resources[index].State = "unknown"
		return verification, errors.Join(runCtx.Err(), s.save())
	}
	stateCtx, stateCancel := context.WithTimeout(context.Background(), 3*time.Second)
	item, err = s.inspectContainer(stateCtx, id)
	stateCancel()
	if err != nil || item.State.Running {
		return verification, ErrUnknown
	}
	if err = atomicJSON(filepath.Join(s.config.Directory, request+".exit.json"), item); err != nil {
		return verification, err
	}
	captureExceeded := errors.Is(runErr, errOutputLimit)
	if runErr != nil && !captureExceeded && item.State.ExitCode == 0 {
		return verification, ErrUnknown
	}
	checked, checkErr := probe.Inspect(s.config.Probe, snapshot)
	if checkErr != nil || checked.CandidateDigest != digest {
		return verification, ErrSandbox
	}
	for _, file := range checked.Files {
		if _, err := readRegular(snapshot, file.Path); err != nil {
			return verification, err
		}
	}
	if !utf8.Valid(output) {
		return verification, ErrSandbox
	}
	truncated := captureExceeded || len(output) > 8192 || item.State.ExitCode == 122
	full := output
	if len(output) > 8192 {
		output = output[:8192]
		for !utf8.Valid(output) {
			output = output[:len(output)-1]
		}
	}
	status := "command_failed"
	if item.State.ExitCode == 0 && !truncated && !item.State.OOMKilled {
		status = "command_passed"
	}
	if truncated {
		status = "output_truncated"
	}
	if item.State.ExitCode == 124 {
		status = "timeout"
	}
	if item.State.OOMKilled {
		status = "oom_killed"
	}
	verification = Verification{Argv: []string{"/usr/local/go/bin/go", "test", "-json", "-count=1", "./..."}, StartedAt: item.State.StartedAt, FinishedAt: item.State.FinishedAt, DurationMS: time.Since(started).Milliseconds(), Profile: "go_test_v1", Image: s.config.Image, RunnerDigest: s.config.RunnerDigest, ExitCode: item.State.ExitCode, OOMKilled: item.State.OOMKilled, OutputTruncated: truncated, Output: string(output), Status: status}
	if err = os.WriteFile(filepath.Join(s.config.Directory, request+".log"), full, 0600); err != nil {
		return verification, ErrSandbox
	}
	if err = atomicJSON(filepath.Join(s.config.Directory, request+".verification.json"), struct {
		CandidateDigest string       `json:"candidate_digest"`
		OutputDigest    string       `json:"output_digest"`
		Verification    Verification `json:"verification"`
	}{digest, control.PayloadDigest(full), verification}); err != nil {
		return verification, err
	}
	s.ledger.Resources[index].State = "finished"
	if err = s.save(); err != nil {
		return verification, err
	}
	return verification, nil
}
