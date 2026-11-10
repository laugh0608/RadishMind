package developmentsandbox

import (
	"context"
	"encoding/json"
	"path/filepath"
	"time"

	artifact "radishmind.local/services/platform/internal/developmentartifact"
	control "radishmind.local/services/platform/internal/developmentcontrol"
)

// DeliverySource joins persisted attempt state with a closed trusted sandbox.
// Client paths, test reports and claimed terminal states are not accepted.
type DeliverySource struct {
	runtime  *control.Runtime
	session  *Session
	identity control.Identity
}

func NewDeliverySource(runtime *control.Runtime, session *Session, identity control.Identity) (*DeliverySource, error) {
	if runtime == nil || session == nil || identity.TenantRef == "" || identity.AccountID != session.config.Delegation.Scope.AccountID || identity.WorkspaceID != session.config.Delegation.Scope.WorkspaceID {
		return nil, artifact.ErrDenied
	}
	return &DeliverySource{runtime: runtime, session: session, identity: identity}, nil
}

func (source DeliverySource) Snapshot(ctx context.Context, scope artifact.Scope, attempt string) (artifact.Input, error) {
	if source.runtime == nil || source.session == nil || scope.TenantRef != source.identity.TenantRef || scope.AccountID != source.identity.AccountID || scope.WorkspaceID != source.identity.WorkspaceID || attempt != source.session.Binding().AttemptID {
		return artifact.Input{}, artifact.ErrDenied
	}
	record, err := source.runtime.Read(ctx, control.Identity{TenantRef: scope.TenantRef, WorkspaceID: scope.WorkspaceID, AccountID: scope.AccountID}, attempt)
	if err != nil || record.Binding.Plan.ApplicationID != scope.ApplicationID || record.Control.Delegation() != source.session.config.Delegation {
		return artifact.Input{}, artifact.ErrDenied
	}
	events := record.Control.Events()
	if len(events) == 0 {
		return artifact.Input{}, artifact.ErrInvalid
	}
	last := events[len(events)-1]
	switch last.Phase {
	case "completed", "failed", "cancelled", "outcome_unknown":
	default:
		return artifact.Input{}, artifact.ErrInvalid
	}
	if last.Phase != "completed" && last.TerminationStatus != "confirmed" {
		return artifact.Input{}, artifact.ErrConflict
	}
	return source.session.deliveryInput(ctx, scope, last.Phase, record.Binding.Plan)
}
func (s *Session) deliveryInput(ctx context.Context, scope artifact.Scope, phase string, plan control.Plan) (artifact.Input, error) {
	select {
	case s.operation <- struct{}{}:
		defer func() { <-s.operation }()
	case <-ctx.Done():
		return artifact.Input{}, ctx.Err()
	}
	if !s.ledger.Closed {
		return artifact.Input{}, artifact.ErrConflict
	}
	planDigest, err := plan.Digest()
	if err != nil || planDigest != s.config.Delegation.Scope.PlanDigest || scope.AccountID != s.config.Delegation.Scope.AccountID || scope.WorkspaceID != s.config.Delegation.Scope.WorkspaceID || scope.ApplicationID != s.config.Delegation.Scope.ApplicationID {
		return artifact.Input{}, artifact.ErrDenied
	}
	if _, err = s.inspect(); err != nil {
		return artifact.Input{}, artifact.ErrInvalid
	}
	in := artifact.Input{SourceCreatedAt: s.artifactCreatedAt, Scope: scope, Origin: s.config.Delegation.Scope, Plan: plan, Probe: s.config.Probe, AttemptState: phase, SnapshotAt: s.artifactSnapshotAt, Files: []artifact.File{}, Tests: s.artifactTests}
	if in.SnapshotAt.IsZero() {
		return artifact.Input{}, artifact.ErrInvalid
	}
	for _, f := range s.config.Probe.SourceFiles {
		base, err := readRegular(s.config.Source, f.Path)
		if err != nil {
			return artifact.Input{}, artifact.ErrInvalid
		}
		candidate, err := readRegular(filepath.Join(s.config.Directory, "candidate"), f.Path)
		if err != nil {
			return artifact.Input{}, artifact.ErrInvalid
		}
		in.Files = append(in.Files, artifact.File{Path: f.Path, Base: string(base), Candidate: string(candidate)})
	}
	raw, err := json.Marshal(in)
	if err != nil {
		return artifact.Input{}, artifact.ErrInvalid
	}
	var clone artifact.Input
	if json.Unmarshal(raw, &clone) != nil {
		return artifact.Input{}, artifact.ErrInvalid
	}
	return clone, nil
}
func (s *Session) retainVerification(request, digest string, v Verification, err error) {
	if err != nil {
		now := time.Now().UTC().Format(time.RFC3339Nano)
		s.artifactTests = append(s.artifactTests, artifact.Test{RequestID: request, CandidateDigest: digest, Profile: "go_test_v1", Argv: []string{"/usr/local/go/bin/go", "test", "-json", "-count=1", "./..."}, ImageDigest: s.config.Image, RunnerDigest: s.config.RunnerDigest, StartedAt: now, FinishedAt: now})
		return
	}
	s.artifactTests = append(s.artifactTests, artifact.Test{RequestID: request, CandidateDigest: digest, Profile: v.Profile, Argv: v.Argv, ImageDigest: v.Image, RunnerDigest: v.RunnerDigest, StartedAt: v.StartedAt, FinishedAt: v.FinishedAt, ExitKnown: true, ExitCode: v.ExitCode, OOMKilled: v.OOMKilled, OutputTruncated: v.OutputTruncated, Log: v.Output})
}
