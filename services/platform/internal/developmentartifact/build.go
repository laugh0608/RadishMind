package developmentartifact

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	control "radishmind.local/services/platform/internal/developmentcontrol"
	probe "radishmind.local/services/platform/internal/developmentprobe"
)

func Build(in Input) (Bundle, error) {
	// Bound and validate raw text before any JSON encoding, which would replace
	// invalid UTF-8 with U+FFFD and could otherwise silently change source code.
	if len(in.Files) != 3 || len(in.Tests) > 8 {
		return Bundle{}, ErrInvalid
	}
	for _, f := range in.Files {
		if !textValid(f.Base) || !textValid(f.Candidate) {
			return Bundle{}, ErrInvalid
		}
	}
	for _, test := range in.Tests {
		if len(test.Log) > MaxTextBytes || !utf8.ValidString(test.Log) || strings.ContainsRune(test.Log, 0) {
			return Bundle{}, ErrInvalid
		}
	}
	copy := in
	copy.Files = slices.Clone(in.Files)
	copy.Tests = slices.Clone(in.Tests)
	for i := range copy.Tests {
		copy.Tests[i].Argv = slices.Clone(in.Tests[i].Argv)
	}
	copy.Plan.Tools.ReadPaths = slices.Clone(in.Plan.Tools.ReadPaths)
	copy.Plan.Tools.WritePaths = slices.Clone(in.Plan.Tools.WritePaths)
	copy.Probe.SourceFiles = slices.Clone(in.Probe.SourceFiles)
	copy.Probe.WritablePaths = slices.Clone(in.Probe.WritablePaths)
	copy.Probe.VerifyArgv = slices.Clone(in.Probe.VerifyArgv)
	b := Bundle{SchemaVersion: Schema, SourceCreatedAt: copy.SourceCreatedAt.UTC().Format(time.RFC3339Nano), Scope: copy.Scope, Origin: copy.Origin, Plan: copy.Plan, Probe: copy.Probe,
		SourceKind: "frozen_probe_tree", AttemptState: copy.AttemptState, SnapshotAt: copy.SnapshotAt.UTC().Format(time.RFC3339Nano),
		ExpiresAt: copy.SourceCreatedAt.Add(TemporaryLifetime).UTC().Format(time.RFC3339Nano), Files: copy.Files, Tests: copy.Tests,
		ReviewState: "not_reviewed", AcceptanceState: "not_accepted", SensitivePolicy: "text_allowlist_v1"}
	if b.Tests == nil {
		b.Tests = []Test{}
	}
	for i := range b.Files {
		f := &b.Files[i]
		f.BaseDigest = probe.Digest([]byte(f.Base))
		f.CandidateDigest = probe.Digest([]byte(f.Candidate))
	}
	b.CandidateDigest = tree(b.Files, false)
	b.Patch = patch(b.Files)
	b.PatchDigest = probe.Digest([]byte(b.Patch))
	for i := range b.Tests {
		t := &b.Tests[i]
		cleaned := redactLog(t.Log)
		t.LogRedacted = t.LogRedacted || cleaned != t.Log
		t.Log = cleaned
		t.LogDigest = probe.Digest([]byte(cleaned))
		t.Status = testStatus(*t, b.CandidateDigest)
	}
	b.VerificationState = verificationState(b.Tests)
	b.Digest = bundleDigest(b)
	if err := b.Validate(); err != nil {
		return Bundle{}, err
	}
	return b, nil
}
func tree(files []File, base bool) string {
	entries := make([]probe.SourceFile, len(files))
	for i, f := range files {
		d := f.CandidateDigest
		if base {
			d = f.BaseDigest
		}
		entries[i] = probe.SourceFile{Path: f.Path, SHA256: d}
	}
	return probe.TreeDigest(entries)
}

// A whole-file unified hunk is deterministic, bounded and handles missing LF.
// It deliberately avoids executing git or an external diff program.
func patch(files []File) string {
	var out strings.Builder
	for _, f := range files {
		if f.Base == f.Candidate {
			continue
		}
		old, new := lines(f.Base), lines(f.Candidate)
		fmt.Fprintf(&out, "--- a/%s\n+++ b/%s\n@@ -1,%d +1,%d @@\n", f.Path, f.Path, len(old), len(new))
		for _, group := range []struct {
			prefix string
			values []string
		}{{"-", old}, {"+", new}} {
			for _, line := range group.values {
				out.WriteString(group.prefix + line)
				if !strings.HasSuffix(line, "\n") {
					out.WriteString("\n\\ No newline at end of file\n")
				}
			}
		}
	}
	return out.String()
}
func lines(s string) []string {
	parts := strings.SplitAfter(s, "\n")
	if parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}
func testStatus(t Test, candidate string) string {
	if t.CandidateDigest != candidate {
		return "stale"
	}
	if !t.ExitKnown {
		return "unknown"
	}
	if t.OutputTruncated {
		return "output_truncated"
	}
	if t.OOMKilled {
		return "oom_killed"
	}
	if t.ExitCode == 124 {
		return "timeout"
	}
	if t.ExitCode != 0 {
		return "command_failed"
	}
	return "command_passed"
}
func verificationState(tests []Test) string {
	if len(tests) == 0 {
		return "missing"
	}
	// Latest invocation is authoritative; a later stale/unknown result cannot
	// silently expose an older passing run as current evidence.
	return tests[len(tests)-1].Status
}
func bundleDigest(b Bundle) string {
	b.Digest = ""
	raw, _ := json.Marshal(b)
	return probe.Digest(raw)
}
func textValid(s string) bool {
	return len(s) > 0 && len(s) <= MaxTextBytes && utf8.ValidString(s) && !strings.ContainsRune(s, 0)
}
func (b Bundle) Validate() error {
	if b.SchemaVersion != Schema || b.Scope.Validate() != nil || b.Plan.Validate() != nil || b.Probe.Validate() != nil || b.SourceKind != "frozen_probe_tree" || b.ReviewState != "not_reviewed" || b.AcceptanceState != "not_accepted" || b.SensitivePolicy != "text_allowlist_v1" {
		return ErrInvalid
	}
	o := b.Origin
	if o.AccountID != b.Scope.AccountID || o.WorkspaceID != b.Scope.WorkspaceID || o.ApplicationID != b.Scope.ApplicationID || b.Plan.ApplicationID != o.ApplicationID || o.ModelRef != b.Plan.ModelRef || o.BindingRevision < 1 || o.BindingRevision > 1000000000 || o.AuthorizationRevision < 1 || o.AuthorizationRevision > 1000000000 {
		return ErrInvalid
	}
	for _, id := range []string{o.BindingID, o.RunID, o.TaskID, o.AttemptID} {
		if !idPattern.MatchString(id) {
			return ErrInvalid
		}
	}
	for _, d := range []string{o.ApplicationAuthorityDigest, o.PricingDigest, o.ToolPolicyDigest} {
		if !digestPattern.MatchString(d) {
			return ErrInvalid
		}
	}
	pd, err := b.Plan.Digest()
	tools, _ := json.Marshal(b.Plan.Tools)
	if err != nil || pd != o.PlanDigest || control.PayloadDigest(tools) != o.ToolPolicyDigest || o.BaselineDigest != b.Plan.BaselineDigest || o.BaselineDigest != probe.TreeDigest(b.Probe.SourceFiles) || !reflect.DeepEqual(b.Plan.Tools.ReadPaths, []string{"go.mod", "list.go", "list_test.go"}) || !reflect.DeepEqual(b.Plan.Tools.WritePaths, []string{"list.go"}) {
		return ErrInvalid
	}
	switch b.AttemptState {
	case "completed", "failed", "cancelled", "outcome_unknown":
	default:
		return ErrInvalid
	}
	start, e1 := time.Parse(time.RFC3339Nano, b.SnapshotAt)
	end, e2 := time.Parse(time.RFC3339Nano, b.ExpiresAt)
	created, e3 := time.Parse(time.RFC3339Nano, b.SourceCreatedAt)
	if e1 != nil || e2 != nil || e3 != nil || created.IsZero() || start.Before(created) || !start.Before(end) || end.Sub(created) != TemporaryLifetime {
		return ErrInvalid
	}
	if len(b.Files) != 3 || b.Tests == nil || len(b.Tests) > int(b.Probe.Limits.MaxToolCalls) {
		return ErrInvalid
	}
	for i, f := range b.Files {
		if f.Path != b.Probe.SourceFiles[i].Path || !textValid(f.Base) || !textValid(f.Candidate) || f.BaseDigest != probe.Digest([]byte(f.Base)) || f.BaseDigest != b.Probe.SourceFiles[i].SHA256 || f.CandidateDigest != probe.Digest([]byte(f.Candidate)) || (f.Path != "list.go" && f.Base != f.Candidate) {
			return ErrInvalid
		}
		if sensitive(f.Base) || sensitive(f.Candidate) {
			return ErrSensitive
		}
	}
	if tree(b.Files, true) != o.BaselineDigest || tree(b.Files, false) != b.CandidateDigest || b.Patch != patch(b.Files) || b.PatchDigest != probe.Digest([]byte(b.Patch)) {
		return ErrInvalid
	}
	seen := map[string]bool{}
	var previous time.Time
	for _, t := range b.Tests {
		begin, e1 := time.Parse(time.RFC3339Nano, t.StartedAt)
		finish, e2 := time.Parse(time.RFC3339Nano, t.FinishedAt)
		if !idPattern.MatchString(t.RequestID) || seen[t.RequestID] || !digestPattern.MatchString(t.CandidateDigest) || t.Profile != "go_test_v1" || !reflect.DeepEqual(t.Argv, []string{"/usr/local/go/bin/go", "test", "-json", "-count=1", "./..."}) || !digestPattern.MatchString(t.ImageDigest) || !digestPattern.MatchString(t.RunnerDigest) || e1 != nil || e2 != nil || begin.IsZero() || begin.Before(created) || finish.Before(begin) || begin.Before(previous) || finish.After(start) || t.ExitCode < 0 || t.ExitCode > 255 || (!t.ExitKnown && t.ExitCode != 0) || len(t.Log) > MaxTextBytes || !utf8.ValidString(t.Log) || strings.ContainsRune(t.Log, 0) || t.LogDigest != probe.Digest([]byte(t.Log)) || t.Status != testStatus(t, b.CandidateDigest) {
			return ErrInvalid
		}
		if sensitive(t.Log) {
			return ErrSensitive
		}
		seen[t.RequestID] = true
		previous = finish
	}
	raw, _ := json.Marshal(b)
	if sensitive(o.ModelRef) {
		return ErrSensitive
	}
	if len(raw) > MaxPackageBytes || b.VerificationState != verificationState(b.Tests) || b.Digest != bundleDigest(b) {
		return ErrInvalid
	}
	return nil
}
