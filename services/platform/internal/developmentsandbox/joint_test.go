package developmentsandbox

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	artifact "radishmind.local/services/platform/internal/developmentartifact"
	control "radishmind.local/services/platform/internal/developmentcontrol"
	probe "radishmind.local/services/platform/internal/developmentprobe"
)

func TestJointSDKCandidateDelivery(t *testing.T) {
	for _, mode := range []string{"normal", "stale", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			f := newJointFixture(t, mode)
			ctx := context.Background()
			type outcome struct {
				proof control.WorkerResult
				err   error
			}
			done := make(chan outcome, 1)
			go func() {
				proof, err := f.runtime.RunWorker(ctx, f.identity, f.issued, f.launch, nil)
				done <- outcome{proof, err}
			}()
			if mode == "cancel" {
				f.cancelRunning(t)
			}
			var result outcome
			select {
			case result = <-done:
			case <-time.After(150 * time.Second):
				t.Fatal("joint task exceeded deadline")
			}
			if mode == "cancel" {
				elapsed := time.Since(f.cancelStarted)
				if elapsed > 10*time.Second {
					t.Fatal("joint cancellation exceeded ten seconds", elapsed)
				}
				if err := atomicJSON(filepath.Join(f.root, "cancellation.json"), map[string]any{"elapsed_ms": elapsed.Milliseconds(), "worker": result.proof}); err != nil {
					t.Fatal(err)
				}
			}
			if !result.proof.ProcessExited || !result.proof.GroupEmpty {
				t.Fatalf("worker not reaped: %+v %v", result.proof, result.err)
			}
			if mode == "cancel" {
				if result.err == nil || result.proof.Finished {
					t.Fatal("cancel reported success")
				}
			} else if result.err != nil || !result.proof.Finished {
				if diagnostic, e := os.ReadFile(filepath.Join(f.launch.Command.Dir, "boundary-failure.json")); e == nil {
					t.Log(string(diagnostic))
				}
				t.Fatalf("joint failed: %+v %v", result.proof, result.err)
			}
			raw, err := os.ReadFile(filepath.Join(f.launch.Command.Dir, "boundaries.json"))
			if err != nil {
				t.Fatal("SDK isolation proof missing", err)
			}
			var boundaries struct {
				UID          int  `json:"uid"`
				DockerDenied bool `json:"docker_denied"`
				DataDenied   bool `json:"controller_data_denied"`
				Caps         bool `json:"capabilities_zero"`
			}
			if json.Unmarshal(raw, &boundaries) != nil || boundaries.UID != 65532 || !boundaries.DockerDenied || !boundaries.DataDenied || !boundaries.Caps {
				t.Fatal("SDK isolation invalid")
			}
			if err = os.WriteFile(filepath.Join(f.root, "sdk-boundaries.json"), raw, 0600); err != nil {
				t.Fatal(err)
			}
			record, err := f.runtime.Read(ctx, f.identity, f.issued.Delegation.Scope.AttemptID)
			if err != nil {
				t.Fatal(err)
			}
			events := record.Control.Events()
			last := events[len(events)-1]
			expectedCalls := int32(4)
			if mode == "cancel" {
				expectedCalls = 3
			}
			if f.backend.calls.Load() != expectedCalls || last.Budget.SpentTokens != int64(expectedCalls)*30 {
				t.Fatal("unexpected model calls or usage", f.backend.calls.Load(), last.Budget)
			}
			kinds := map[string]bool{}
			for i, e := range events {
				if e.Sequence != int64(i+1) {
					t.Fatal("event gap")
				}
				kinds[e.Type] = true
			}
			for _, kind := range []string{"worker_claimed", "worker_ready", "worker_waiting_tool", "worker_exited"} {
				if !kinds[kind] {
					t.Fatal("missing event", kind)
				}
			}
			if mode == "cancel" {
				if last.TerminationStatus != "confirmed" || last.Phase == "completed" {
					t.Fatal("cancel termination unconfirmed", last)
				}
			} else if last.Phase != "completed" || !kinds["sandbox_closed"] {
				t.Fatal("completion missing sandbox closure", last)
			}
			if err = atomicJSON(filepath.Join(f.root, "events.json"), events); err != nil {
				t.Fatal(err)
			}
			current, err := probe.Inspect(f.session.config.Probe, f.session.config.Source)
			if err != nil || len(current.ChangedPaths) != 0 {
				t.Fatal("baseline changed", err)
			}
			ids, err := f.session.owned(ctx)
			if err != nil || len(ids) != 0 {
				t.Fatal("owned containers remain", ids, err)
			}
			f.verifyDelivery(t, mode)
		})
	}
}
func (f *jointFixture) cancelRunning(t *testing.T) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		ids, err := f.session.owned(ctx)
		running := false
		if err == nil && len(ids) == 1 {
			item, e := f.session.inspectContainer(ctx, ids[0])
			running = e == nil && item.State.Running
		}
		cancel()
		if running {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("code container did not start")
		}
		time.Sleep(50 * time.Millisecond)
	}
	record, err := f.runtime.Read(context.Background(), f.identity, f.issued.Delegation.Scope.AttemptID)
	if err != nil {
		t.Fatal(err)
	}
	// Cancel uses the persisted CAS version, not a local goroutine cancellation.
	f.cancelStarted = time.Now()
	if _, err = f.runtime.Cancel(context.Background(), f.identity, f.issued.Delegation.Scope.AttemptID, record.Version); err != nil {
		t.Fatal(err)
	}
}
func (f *jointFixture) verifyDelivery(t *testing.T, mode string) {
	t.Helper()
	ctx := context.Background()
	scope := f.authority.scope
	attempt := f.issued.Delegation.Scope.AttemptID
	b, err := f.service.Preview(ctx, scope, attempt)
	if err != nil {
		t.Fatal("preview", err)
	}
	wanted := map[string]string{"normal": "command_passed", "stale": "stale", "cancel": "unknown"}[mode]
	if b.VerificationState != wanted || len(b.Tests) != 1 || b.AcceptanceState != "not_accepted" || b.ReviewState != "not_reviewed" {
		t.Fatalf("unexpected delivery evidence: %+v", b)
	}
	if b.Origin.AttemptID != attempt || b.Plan.BaselineDigest != f.session.config.Plan.BaselineDigest || !strings.Contains(b.Patch, "list.go") {
		t.Fatal("lineage/patch lost")
	}
	if mode == "normal" {
		count := 0
		for _, line := range strings.Split(b.Tests[0].Log, "\n") {
			var e struct{ Action, Test string }
			if json.Unmarshal([]byte(line), &e) == nil && e.Action == "pass" && strings.Contains(e.Test, "/") {
				count++
			}
		}
		if count != 8 || b.Tests[0].CandidateDigest != b.CandidateDigest {
			t.Fatal("frozen tests or candidate binding", count)
		}
	}
	if mode == "stale" && (!b.Tests[0].ExitKnown || b.Tests[0].ExitCode == 0 || b.Tests[0].CandidateDigest == b.CandidateDigest || !strings.Contains(b.Tests[0].Log, "filter_before_limit") || !strings.Contains(b.Tests[0].Log, "limit_two")) {
		t.Fatal("baseline failure/old evidence not retained")
	}
	var count int
	if err = f.db.DB().QueryRow(`SELECT COUNT(*) FROM development_artifacts`).Scan(&count); err != nil || count != 0 {
		t.Fatal("preview saved implicitly", count, err)
	}
	if _, err = f.service.Save(ctx, scope, attempt, "save_stale", probe.Digest([]byte("unconfirmed"))); !errors.Is(err, artifact.ErrConflict) {
		t.Fatal("stale confirmation accepted", err)
	}
	if _, err = f.db.DB().Exec(`CREATE TRIGGER joint_fail_save BEFORE INSERT ON development_artifacts BEGIN SELECT RAISE(ABORT,'joint injected failure'); END`); err != nil {
		t.Fatal(err)
	}
	if _, err = f.service.Save(ctx, scope, attempt, "save_joint", b.Digest); err == nil {
		t.Fatal("failed write succeeded")
	}
	if err = f.db.DB().QueryRow(`SELECT COUNT(*) FROM development_artifacts`).Scan(&count); err != nil || count != 0 {
		t.Fatal("partial save", err)
	}
	if _, err = f.db.DB().Exec(`DROP TRIGGER joint_fail_save`); err != nil {
		t.Fatal(err)
	}
	saved, err := f.service.Save(ctx, scope, attempt, "save_joint", b.Digest)
	if err != nil {
		t.Fatal("save retry", err)
	}
	retry, err := f.service.Save(ctx, scope, attempt, "save_joint", b.Digest)
	if err != nil || retry.ArtifactID != saved.ArtifactID {
		t.Fatal("save not idempotent", err)
	}
	if err = f.runtime.Close(ctx); err != nil {
		t.Fatal(err)
	}
	f.runtimeClosed = true
	if err = f.db.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := jointDatabase(filepath.Join(f.root, "joint.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	controlStore, err := control.NewSQLiteStore(reopened.DB())
	if err != nil {
		t.Fatal(err)
	}
	restarted, err := control.NewRuntime(controlStore, f.authority, f.backend)
	if err != nil {
		t.Fatal(err)
	}
	defer restarted.Close(ctx)
	persisted, err := restarted.Read(ctx, f.identity, attempt)
	if err != nil || persisted.Control.Events()[len(persisted.Control.Events())-1].Phase != b.AttemptState {
		t.Fatal("attempt restart read", err)
	}
	store, err := artifact.NewSQLiteStore(reopened.DB())
	if err != nil {
		t.Fatal(err)
	}
	readOnly, err := artifact.NewService(store, f.authority, nil)
	if err != nil {
		t.Fatal(err)
	}
	got, err := readOnly.Read(ctx, scope, saved.ArtifactID)
	if err != nil || got.Bundle.Digest != b.Digest {
		t.Fatal("restart read", err)
	}
	exported, err := readOnly.Export(ctx, scope, saved.ArtifactID, "export_joint")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.MarshalIndent(exported, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := artifact.ParseExport(raw)
	if err != nil || parsed.Record.Bundle.Digest != b.Digest {
		t.Fatal("export failed strict reread", err)
	}
	if err = os.WriteFile(filepath.Join(f.root, "export.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	f.authority.exportDenied = true
	if _, err = readOnly.Export(ctx, scope, saved.ArtifactID, "export_denied"); !errors.Is(err, artifact.ErrDenied) {
		t.Fatal("export revocation ignored", err)
	}
	if _, err = readOnly.Read(ctx, scope, saved.ArtifactID); err != nil {
		t.Fatal("independent read denied", err)
	}
	if f.backend.calls.Load() != map[string]int32{"normal": 4, "stale": 4, "cancel": 3}[mode] {
		t.Fatal("restart replayed model")
	}
	t.Log("saved/exported", saved.ArtifactID, b.Digest, b.VerificationState)
}
