package developmentartifact

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	control "radishmind.local/services/platform/internal/developmentcontrol"
	probe "radishmind.local/services/platform/internal/developmentprobe"
	"strings"
	"testing"
	"time"
)

func fixtureInput(t *testing.T) Input {
	t.Helper()
	root := filepath.Join("..", "..", "..", "..", "scripts", "eval", "fixtures", "development-task")
	raw, err := os.ReadFile(filepath.Join(root, "list-filter-plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := probe.ParsePlan(raw)
	if err != nil {
		t.Fatal(err)
	}
	hash := probe.Digest([]byte("synthetic artifact fixture"))
	plan := control.Plan{SchemaVersion: "development_attempt_plan.v1", ApplicationID: "application_fixture", ExecutionProfile: "prompt_invoke_v2", ModelRef: "profile:fixture", BaselineDigest: probe.TreeDigest(p.SourceFiles), InitialMessagesDigest: hash, Tools: control.ToolPolicy{ReadPaths: []string{"go.mod", "list.go", "list_test.go"}, WritePaths: []string{"list.go"}, VerificationProfile: "go_test_v1"}, Limits: control.Limits{MaxModelCalls: 8, MaxToolCalls: 8, MaxTokens: 1000, MaxCostMicrousd: 1000}, LifetimeMS: 600000, MaxOutputTokens: 50}
	pd, _ := plan.Digest()
	toolRaw, _ := json.Marshal(plan.Tools)
	origin := control.Scope{AccountID: "account_fixture", WorkspaceID: "workspace_fixture", ApplicationID: plan.ApplicationID, ApplicationAuthorityDigest: hash, BindingID: "binding_fixture", BindingRevision: 1, AuthorizationRevision: 1, RunID: "run_fixture", TaskID: "task_fixture", AttemptID: "attempt_fixture", PlanDigest: pd, BaselineDigest: plan.BaselineDigest, ModelRef: plan.ModelRef, PricingDigest: hash, ToolPolicyDigest: probe.Digest(toolRaw)}
	now := time.Date(2026, 10, 31, 8, 0, 0, 0, time.UTC)
	in := Input{Scope: Scope{"tenant_fixture", "workspace_fixture", "account_fixture", plan.ApplicationID}, Origin: origin, Plan: plan, Probe: p, AttemptState: "completed", SourceCreatedAt: now.Add(-time.Minute), SnapshotAt: now, Files: []File{}, Tests: []Test{}}
	for _, f := range p.SourceFiles {
		raw, err := os.ReadFile(filepath.Join(root, "list-filter", f.Path))
		if err != nil {
			t.Fatal(err)
		}
		content := string(raw)
		candidate := content
		if f.Path == "list.go" {
			candidate = strings.ReplaceAll(content, "result[:limit-1]", "result[:limit]")
		}
		in.Files = append(in.Files, File{Path: f.Path, Base: content, Candidate: candidate})
	}
	b, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	in.Tests = []Test{{RequestID: "tool_test", CandidateDigest: b.CandidateDigest, Profile: "go_test_v1", Argv: []string{"/usr/local/go/bin/go", "test", "-json", "-count=1", "./..."}, ImageDigest: hash, RunnerDigest: hash, StartedAt: now.Add(-2 * time.Second).Format(time.RFC3339Nano), FinishedAt: now.Add(-time.Second).Format(time.RFC3339Nano), ExitKnown: true, Log: "synthetic command result\n"}}
	return in
}
func mustBuild(t *testing.T, in Input) Bundle {
	t.Helper()
	b, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func TestBoundPackagePatchAndEvidence(t *testing.T) {
	in := fixtureInput(t)
	b := mustBuild(t, in)
	if b.VerificationState != "command_passed" || b.AcceptanceState != "not_accepted" || b.ReviewState != "not_reviewed" || b.Digest == "" {
		t.Fatal("incorrect evidence projection")
	}
	raw, _ := json.Marshal(b)
	parsed, err := Parse(raw)
	if err != nil || parsed.Digest != b.Digest {
		t.Fatal(err)
	}
	if git, err := exec.LookPath("git"); err == nil {
		directory := t.TempDir()
		for _, f := range b.Files {
			if err = os.WriteFile(filepath.Join(directory, f.Path), []byte(f.Base), 0600); err != nil {
				t.Fatal(err)
			}
		}
		cmd := exec.Command(git, "apply", "--whitespace=nowarn", "-")
		cmd.Dir = directory
		cmd.Stdin = strings.NewReader(b.Patch)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("patch not applicable: %v %s", err, output)
		}
		for _, f := range b.Files {
			raw, err := os.ReadFile(filepath.Join(directory, f.Path))
			if err != nil || string(raw) != f.Candidate {
				t.Fatal("patch did not reconstruct candidate")
			}
		}
	}
	cases := map[string]func(*Input){
		"wrong scope":     func(x *Input) { x.Scope.AccountID = "account_other" },
		"wrong base":      func(x *Input) { x.Files[0].Base += "x" },
		"protected file":  func(x *Input) { x.Files[0].Candidate += "x" },
		"escape":          func(x *Input) { x.Files[1].Path = "../list.go" },
		"extra file":      func(x *Input) { x.Files = append(x.Files, x.Files[0]) },
		"active":          func(x *Input) { x.AttemptState = "active" },
		"policy":          func(x *Input) { x.Origin.ToolPolicyDigest = probe.Digest([]byte("other")) },
		"oversize source": func(x *Input) { x.Files[1].Candidate = strings.Repeat("a", MaxTextBytes+1) },
		"invalid UTF8":    func(x *Input) { x.Files[1].Candidate = string([]byte{0xff}) },
		"duplicate test":  func(x *Input) { x.Tests = append(x.Tests, x.Tests[0]) },
		"shell":           func(x *Input) { x.Tests[0].Argv = []string{"sh", "-c", "go test"} },
		"future evidence": func(x *Input) { x.Tests[0].FinishedAt = x.SnapshotAt.Add(time.Second).Format(time.RFC3339Nano) },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			in := fixtureInput(t)
			change(&in)
			if _, err := Build(in); err == nil {
				t.Fatal("accepted invalid input")
			}
		})
	}
	b.Files[1].Candidate += "modified"
	if b.Validate() == nil {
		t.Fatal("accepted modified bytes")
	}
}
func TestPartialEvidenceAndSensitiveHandling(t *testing.T) {
	for _, phase := range []string{"failed", "cancelled", "outcome_unknown"} {
		in := fixtureInput(t)
		in.AttemptState = phase
		in.Tests = nil
		if b := mustBuild(t, in); b.VerificationState != "missing" {
			t.Fatal(b.VerificationState)
		}
	}
	for _, tc := range []struct {
		name   string
		change func(*Test)
		want   string
	}{
		{"stale", func(x *Test) { x.CandidateDigest = probe.Digest([]byte("old")) }, "stale"},
		{"unknown", func(x *Test) { x.ExitKnown = false }, "unknown"},
		{"truncated", func(x *Test) { x.OutputTruncated = true }, "output_truncated"},
		{"oom", func(x *Test) { x.OOMKilled = true }, "oom_killed"},
		{"timeout", func(x *Test) { x.ExitCode = 124 }, "timeout"},
		{"failed", func(x *Test) { x.ExitCode = 1 }, "command_failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := fixtureInput(t)
			tc.change(&in.Tests[0])
			if b := mustBuild(t, in); b.VerificationState != tc.want {
				t.Fatal(b.VerificationState)
			}
		})
	}
	in := fixtureInput(t)
	in.Tests[0].Log = "safe\nAuthorization: Bearer synthetic-private-value\n/Users/example/private.go\n"
	b := mustBuild(t, in)
	if !b.Tests[0].LogRedacted || strings.Contains(b.Tests[0].Log, "synthetic-private-value") || strings.Contains(b.Tests[0].Log, "/Users/") {
		t.Fatal("raw secrets leaked")
	}
	for _, secret := range []string{"api_key = synthetic-secret", "-----BEGIN PRIVATE KEY-----", "/home/test/private", "postgres://user:pass@host/db"} {
		in := fixtureInput(t)
		in.Files[1].Candidate += "\n// " + secret
		if _, err := Build(in); err != ErrSensitive {
			t.Fatalf("sensitive source accepted: %v", err)
		}
	}
}
func TestStrictJSONAndCanonicalFixture(t *testing.T) {
	b := mustBuild(t, fixtureInput(t))
	raw, _ := json.Marshal(b)
	for _, bad := range [][]byte{append(append([]byte{}, raw...), []byte(" {}")...), bytes.Replace(raw, []byte(`"scope":`), []byte(`"Scope":`), 1), bytes.Replace(raw, []byte(`"exit_code":0,`), nil, 1), bytes.Replace(raw, []byte(`"exit_code":0`), []byte(`"exit_code":null`), 1), bytes.Replace(raw, []byte(`"exit_code":0`), []byte(`"exit_code":0,"exit_code":0`), 1), append([]byte(`{"extra":1,`), raw[1:]...)} {
		if _, err := Parse(bad); err == nil {
			t.Fatal("accepted noncanonical fields")
		}
	}
	path := filepath.Join("..", "..", "..", "..", "scripts", "eval", "fixtures", "development-control", "artifact.json")
	if os.Getenv("UPDATE_DEVELOPMENT_ARTIFACT_FIXTURE") == "1" {
		pretty, _ := json.MarshalIndent(b, "", "  ")
		if err := os.WriteFile(path, append(pretty, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
	fixture, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := Parse(fixture)
	if err != nil || parsed.Digest != b.Digest {
		t.Fatal("canonical fixture differs", err)
	}
}
func TestPatchMissingFinalNewline(t *testing.T) {
	got := patch([]File{{Path: "list.go", Base: "old", Candidate: "new\n"}})
	want := "--- a/list.go\n+++ b/list.go\n@@ -1,1 +1,1 @@\n-old\n\\ No newline at end of file\n+new\n"
	if got != want {
		t.Fatalf("incorrect patch %q", got)
	}
}

func TestLargeBoundedPackageAndExportTampering(t *testing.T) {
	in := fixtureInput(t)
 in.Probe.Limits.MaxToolCalls=8
	for i := 0; i < 7; i++ {
		test := in.Tests[0]
		test.RequestID = fmt.Sprintf("tool_%d", i)
		test.StartedAt = in.SnapshotAt.Add(time.Duration(-8+i) * time.Second).Format(time.RFC3339Nano)
		test.FinishedAt = test.StartedAt
		test.Log = strings.Repeat("ordinary output ", 3000)
		if i == 0 {
			in.Tests = []Test{}
		}
		in.Tests = append(in.Tests, test)
	}
	b := mustBuild(t, in)
	raw, _ := json.Marshal(b)
	if len(raw) <= 262144 {
		t.Fatal("fixture does not exercise larger bound")
	}
	if _, err := Parse(raw); err != nil {
		t.Fatal(err)
	}
	if _, err := Parse(bytes.Repeat([]byte(" "), MaxPackageBytes+1)); err == nil {
		t.Fatal("oversize JSON accepted")
	}
	in.Tests[0].Log = strings.Repeat("x", MaxTextBytes+1)
	if _, err := Build(in); err == nil {
		t.Fatal("oversize log accepted")
	}
	s, _, src, _, _ := newServiceFixture(t)
	preview := mustBuild(t, src.in)
	r, err := s.Save(context.Background(), src.in.Scope, "attempt_fixture", "save_export", preview.Digest)
	if err != nil {
		t.Fatal(err)
	}
	e, err := s.Export(context.Background(), src.in.Scope, r.ArtifactID, "export_check")
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(e)
	if _, err := ParseExport(encoded); err != nil {
		t.Fatal(err)
	}
	e.Record.Bundle.Files[1].Candidate += "tampered"
	encoded, _ = json.Marshal(e)
	if _, err := ParseExport(encoded); err == nil {
		t.Fatal("tampered export accepted")
	}
}
