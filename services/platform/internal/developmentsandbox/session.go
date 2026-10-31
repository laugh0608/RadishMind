package developmentsandbox

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sync"

	control "radishmind.local/services/platform/internal/developmentcontrol"
	probe "radishmind.local/services/platform/internal/developmentprobe"
)

type Session struct {
	config    Config
	ledger    ledger
	mu        sync.Mutex // state/cancel; independent from tool IO and container cleanup
	operation chan struct{}
	cancel    context.CancelFunc
	closing   bool
	used      map[string]bool
}

func New(config Config) (*Session, error) {
	config.Plan.Tools.ReadPaths = slices.Clone(config.Plan.Tools.ReadPaths)
	config.Plan.Tools.WritePaths = slices.Clone(config.Plan.Tools.WritePaths)
	config.Probe.SourceFiles = slices.Clone(config.Probe.SourceFiles)
	config.Probe.WritablePaths = slices.Clone(config.Probe.WritablePaths)
	config.Probe.VerifyArgv = slices.Clone(config.Probe.VerifyArgv)

	if config.Probe.Validate() != nil || config.Plan.Validate() != nil || config.Delegation.Validate() != nil || !filepath.IsAbs(config.DockerPath) || !filepath.IsAbs(config.RunnerPath) || !filepath.IsAbs(config.Source) || !filepath.IsAbs(config.Directory) || !imagePattern.MatchString(config.Image) || !imagePattern.MatchString(config.RunnerDigest) {
		return nil, ErrSandbox
	}
	digest, err := config.Plan.Digest()
	if err != nil || digest != config.Delegation.Scope.PlanDigest || config.Plan.BaselineDigest != probe.TreeDigest(config.Probe.SourceFiles) || config.Plan.BaselineDigest != config.Delegation.Scope.BaselineDigest || !reflect.DeepEqual(config.Plan.Tools.ReadPaths, []string{"go.mod", "list.go", "list_test.go"}) || !reflect.DeepEqual(config.Plan.Tools.WritePaths, []string{"list.go"}) {
		return nil, ErrSandbox
	}
	info, err := os.Lstat(config.Source)
	if err != nil || !info.IsDir() {
		return nil, ErrSandbox
	}
	baseline, err := probe.Inspect(config.Probe, config.Source)
	if err != nil || len(baseline.ChangedPaths) != 0 {
		return nil, ErrSandbox
	}
	runnerInfo, err := os.Lstat(config.RunnerPath)
	if err != nil || !runnerInfo.Mode().IsRegular() || !singleLink(runnerInfo) || runnerInfo.Size() < 1 || runnerInfo.Size() > 16<<20 {
		return nil, ErrSandbox
	}
	runnerFile, err := os.Open(config.RunnerPath)
	if err != nil {
		return nil, ErrSandbox
	}
	opened, statErr := runnerFile.Stat()
	runner, readErr := io.ReadAll(io.LimitReader(runnerFile, 16<<20+1))
	closeErr := runnerFile.Close()
	if statErr != nil || !os.SameFile(runnerInfo, opened) || readErr != nil || closeErr != nil || len(runner) != int(runnerInfo.Size()) || control.PayloadDigest(runner) != config.RunnerDigest {
		return nil, ErrSandbox
	}
	if err = os.Mkdir(config.Directory, 0700); err != nil {
		return nil, ErrSandbox
	}
	if err = os.Mkdir(filepath.Join(config.Directory, "candidate"), 0700); err != nil {
		return nil, ErrSandbox
	}
	for _, file := range config.Probe.SourceFiles {
		data, err := readRegular(config.Source, file.Path)
		if err != nil || probe.Digest(data) != file.SHA256 {
			return nil, ErrSandbox
		}
		if err = os.WriteFile(filepath.Join(config.Directory, "candidate", file.Path), data, 0600); err != nil {
			return nil, ErrSandbox
		}
	}
	var nonce [16]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		return nil, ErrSandbox
	}
	s := &Session{config: config, operation: make(chan struct{}, 1), used: map[string]bool{}}
	s.ledger = ledger{"development_sandbox_resources.v1", "sandbox_" + hex.EncodeToString(nonce[:]), control.SandboxBindingFor(config.Delegation), config.Image, config.RunnerDigest, []resource{}, false}
	// Freeze the runner in the private ledger directory; callers cannot swap it
	// between initial digest validation and Docker's later bind mount.
	if err = os.WriteFile(filepath.Join(config.Directory, "runner"), runner, 0555); err != nil {
		return nil, ErrSandbox
	}
	s.config.RunnerPath = filepath.Join(config.Directory, "runner")
	if err = s.save(); err != nil {
		return nil, err
	}
	return s, nil
}
func (s *Session) Binding() control.SandboxBinding { return s.ledger.Binding }
func (s *Session) save() error {
	return atomicJSON(filepath.Join(s.config.Directory, "resources.json"), s.ledger)
}
func (s *Session) HandleTool(ctx context.Context, call control.ToolCall) (string, error) {
	if call.Validate() != nil || call.AttemptID != s.Binding().AttemptID || !s.config.Plan.Tools.Allows(call.Arguments) {
		return "", ErrSandbox
	}
	select {
	case s.operation <- struct{}{}:
		defer func() { <-s.operation }()
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		return "", ErrSandbox
	}
	s.mu.Lock()
	if s.closing || s.used[call.RequestID] || len(s.used) >= int(s.config.Plan.Limits.MaxToolCalls) {
		s.mu.Unlock()
		return "", ErrSandbox
	}
	s.used[call.RequestID] = true
	opCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.mu.Unlock()
	defer func() { cancel(); s.mu.Lock(); s.cancel = nil; s.mu.Unlock() }()
	if err := opCtx.Err(); err != nil {
		return "", err
	}
	candidate, err := s.inspect()
	if err != nil {
		return "", err
	}
	action := call.Arguments
	result := observation(action, candidate.CandidateDigest)
	switch action.Action {
	case "read_file":
		data, err := readRegular(filepath.Join(s.config.Directory, "candidate"), action.Path)
		if err != nil {
			return "", err
		}
		result.Content = string(data)
		result.FileDigest = control.PayloadDigest(data)
	case "replace_file":
		if len(action.Content) == 0 {
			return "", ErrSandbox
		}
		directory := filepath.Join(s.config.Directory, "candidate")
		data, err := readRegular(directory, action.Path)
		if err != nil || control.PayloadDigest(data) != action.ExpectedDigest {
			return "", ErrSandbox
		}
		if err = atomicFile(filepath.Join(directory, action.Path), []byte(action.Content), 0600); err != nil {
			return "", err
		}
		current, err := s.inspect()
		if err != nil {
			return "", err
		}
		result.FileDigest = control.PayloadDigest([]byte(action.Content))
		result.CandidateDigest = current.CandidateDigest
	case "run_tests":
		if action.CandidateDigest != candidate.CandidateDigest {
			return "", ErrSandbox
		}
		verification, err := s.runTests(opCtx, call.RequestID, candidate.CandidateDigest)
		if err != nil {
			return "", err
		}
		current, err := s.inspect()
		if err != nil || current.CandidateDigest != candidate.CandidateDigest {
			return "", ErrSandbox
		}
		result.Verification = &verification
	default:
		return "", ErrSandbox
	}
	if opCtx.Err() != nil {
		return "", opCtx.Err()
	}
	return marshalObservation(result)
}
func (s *Session) Close(ctx context.Context) (bool, error) {
	s.mu.Lock()
	s.closing = true
	if s.cancel != nil {
		s.cancel()
	}
	s.mu.Unlock()
	select {
	case s.operation <- struct{}{}:
		defer func() { <-s.operation }()
	case <-ctx.Done():
		return false, ErrUnknown
	}
	empty, err := s.cleanup(ctx)
	if err != nil || !empty {
		return false, errors.Join(ErrUnknown, err)
	}
	s.ledger.Closed = true
	if err = s.save(); err != nil {
		return false, err
	}
	return true, nil
}
