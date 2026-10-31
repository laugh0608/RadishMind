package developmentsandbox

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	control "radishmind.local/services/platform/internal/developmentcontrol"
	probe "radishmind.local/services/platform/internal/developmentprobe"
)

func fixture(t *testing.T) Config {
	t.Helper()
	root := filepath.Join("..", "..", "..", "..", "scripts", "eval", "fixtures", "development-task")
	raw, err := os.ReadFile(filepath.Join(root, "list-filter-plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	plan, err := probe.ParsePlan(raw)
	if err != nil {
		t.Fatal(err)
	}
	source, err := filepath.Abs(filepath.Join(root, "list-filter"))
	if err != nil {
		t.Fatal(err)
	}
	sha := control.PayloadDigest([]byte("synthetic"))
	attempt := control.Plan{SchemaVersion: "development_attempt_plan.v1", ApplicationID: "application_fixture", ExecutionProfile: "prompt_invoke_v2", ModelRef: "profile:fixture", BaselineDigest: probe.TreeDigest(plan.SourceFiles), InitialMessagesDigest: sha, Tools: control.ToolPolicy{ReadPaths: []string{"go.mod", "list.go", "list_test.go"}, WritePaths: []string{"list.go"}, VerificationProfile: "go_test_v1"}, Limits: control.Limits{MaxModelCalls: 8, MaxToolCalls: 8, MaxTokens: 1000, MaxCostMicrousd: 1000}, LifetimeMS: 600000, MaxOutputTokens: 50}
	digest, err := attempt.Digest()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UnixMilli()
	delegation := control.Delegation{SchemaVersion: "development_attempt_delegation.v1", DelegationID: "delegation_fixture", Revision: 1, Audience: control.Audience, ExecutorInstanceID: "executor_fixture", IssuedAtMS: now, ExpiresAtMS: now + 600000, Limits: attempt.Limits, Scope: control.Scope{AccountID: "account_fixture", WorkspaceID: "workspace_fixture", ApplicationID: attempt.ApplicationID, ApplicationAuthorityDigest: sha, BindingID: "binding_fixture", BindingRevision: 1, AuthorizationRevision: 1, RunID: "run_fixture", TaskID: "task_fixture", AttemptID: "attempt_fixture", PlanDigest: digest, BaselineDigest: attempt.BaselineDigest, ModelRef: attempt.ModelRef, PricingDigest: sha, ToolPolicyDigest: sha}}
	directory := t.TempDir()
	runner := filepath.Join(directory, "runner")
	if err = os.WriteFile(runner, []byte("synthetic runner"), 0700); err != nil {
		t.Fatal(err)
	}
	return Config{DockerPath: "/usr/bin/false", Image: "sha256:" + strings.Repeat("a", 64), RunnerPath: runner, RunnerDigest: control.PayloadDigest([]byte("synthetic runner")), Source: source, Directory: filepath.Join(directory, "case"), Probe: plan, Plan: attempt, Delegation: delegation}
}
func call(action control.ToolAction, request string) control.ToolCall {
	return control.ToolCall{SchemaVersion: "development_tool_call.v1", AttemptID: "attempt_fixture", RequestID: request, ParentRequestID: "model_fixture", ToolCallID: "call_fixture", Arguments: action}
}
func action(kind, path string) control.ToolAction {
	return control.ToolAction{SchemaVersion: "development_tool_action.v1", Action: kind, Path: path}
}
func readObservation(t *testing.T, body string, err error) Observation {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	var v Observation
	if json.Unmarshal([]byte(body), &v) != nil {
		t.Fatal(body)
	}
	return v
}
func invoke(t *testing.T, s *Session, c control.ToolCall) Observation {
	t.Helper()
	body, err := s.HandleTool(context.Background(), c)
	return readObservation(t, body, err)
}
func TestFileToolsBindCandidateAndRejectExpansion(t *testing.T) {
	config := fixture(t)
	s, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	initial := invoke(t, s, call(action("read_file", "list.go"), "tool_read"))
	replacement := action("replace_file", "list.go")
	replacement.ExpectedDigest = initial.FileDigest
	replacement.Content = strings.ReplaceAll(initial.Content, "result[:limit-1]", "result[:limit]")
	changed := invoke(t, s, call(replacement, "tool_replace"))
	if changed.CandidateDigest == initial.CandidateDigest {
		t.Fatal("candidate did not change")
	}
	for _, c := range []control.ToolCall{call(replacement, "tool_cas"), call(action("read_file", "../go.mod"), "tool_escape"), call(action("replace_file", "go.mod"), "tool_protected"), call(action("read_file", "list.go"), "tool_read")} {
		if _, err = s.HandleTool(context.Background(), c); err == nil {
			t.Fatal("expansion/replay accepted", c)
		}
	}
	protected, err := readRegular(config.Source, "list.go")
	if err != nil || control.PayloadDigest(protected) != initial.FileDigest {
		t.Fatal("source changed")
	}
	if _, err = New(config); err == nil {
		t.Fatal("overwrote session")
	}
	if _, err = s.Close(context.Background()); err == nil {
		t.Fatal("unavailable Docker claimed clean")
	}
}
func TestSourceAndCandidateRejectLinksAndMutation(t *testing.T) {
	for _, kind := range []string{"symlink", "hardlink", "extra", "protected"} {
		t.Run(kind, func(t *testing.T) {
			config := fixture(t)
			s, err := New(config)
			if err != nil {
				t.Fatal(err)
			}
			directory := filepath.Join(config.Directory, "candidate")
			switch kind {
			case "symlink", "hardlink":
				target := filepath.Join(directory, "list.go")
				outside := filepath.Join(t.TempDir(), "outside")
				if err = os.WriteFile(outside, []byte("outside"), 0600); err != nil {
					t.Fatal(err)
				}
				if err = os.Remove(target); err != nil {
					t.Fatal(err)
				}
				if kind == "symlink" {
					err = os.Symlink(outside, target)
				} else {
					err = os.Link(outside, target)
				}
			case "extra":
				err = os.WriteFile(filepath.Join(directory, "secret"), []byte("secret"), 0600)
			case "protected":
				err = os.WriteFile(filepath.Join(directory, "go.mod"), []byte("changed"), 0600)
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err = s.HandleTool(context.Background(), call(action("read_file", "list.go"), "tool_bad")); err == nil {
				t.Fatal("unsafe candidate read")
			}
		})
	}
}
func TestRecoveryRejectsBadLedgerBeforeDocker(t *testing.T) {
	config := fixture(t)
	s, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	ledgerFile := filepath.Join(config.Directory, "resources.json")
	data, err := os.ReadFile(ledgerFile)
	if err != nil {
		t.Fatal(err)
	}
	data = []byte(strings.Replace(string(data), `"resources": []`, `"resources": null`, 1))
	if err = os.WriteFile(ledgerFile, data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = CleanupRecovered(context.Background(), config.DockerPath, config.Directory); err == nil {
		t.Fatal("bad ledger accepted")
	}
	if s.Binding() != control.SandboxBindingFor(config.Delegation) {
		t.Fatal("binding changed")
	}
}

func TestConfirmedConfigIsFrozenAgainstCallerReuse(t *testing.T) {
	config := fixture(t)
	s, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	config.Plan.Tools.WritePaths[0] = "go.mod"
	config.Probe.SourceFiles[0].SHA256 = control.PayloadDigest([]byte("changed"))
	initial := invoke(t, s, call(action("read_file", "list.go"), "tool_read"))
	replace := action("replace_file", "list.go")
	replace.ExpectedDigest = initial.FileDigest
	replace.Content = initial.Content + "\n// synthetic candidate\n"
	invoke(t, s, call(replace, "tool_replace"))
}
