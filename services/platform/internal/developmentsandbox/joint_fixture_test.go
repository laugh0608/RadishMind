package developmentsandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	artifact "radishmind.local/services/platform/internal/developmentartifact"
	control "radishmind.local/services/platform/internal/developmentcontrol"
	probe "radishmind.local/services/platform/internal/developmentprobe"
	"radishmind.local/services/platform/internal/sqlitedev"
	artifacts "radishmind.local/services/platform/migrations/sqlite/development_artifacts"
	attempts "radishmind.local/services/platform/migrations/sqlite/development_attempts"
)

// Qualification and provider are synthetic; the SDK, control store, sandbox,
// source reader and artifact store are real. No provider or Web claim is made.
type jointAuthority struct {
	scope        artifact.Scope
	exportDenied bool
}

func (a *jointAuthority) Authorize(_ context.Context, s artifact.Scope, p string) error {
	if s != a.scope || (p == artifact.PermissionExport && a.exportDenied) {
		return artifact.ErrDenied
	}
	return nil
}
func (a *jointAuthority) Resolve(_ context.Context, id control.Identity, p control.Plan) (control.Selection, error) {
	if id.AccountID != a.scope.AccountID || id.WorkspaceID != a.scope.WorkspaceID || id.TenantRef != a.scope.TenantRef || p.ApplicationID != a.scope.ApplicationID {
		return control.Selection{}, control.ErrIneligible
	}
	sha := control.PayloadDigest([]byte("joint-synthetic-authority"))
	return control.Selection{ApplicationAuthorityDigest: sha, IdentityDigest: sha, ModelDigest: sha, WireModel: "fixture-model", PricingDigest: sha, TokenUnit: 1000000, InputPriceMicros: 1000000, OutputPriceMicros: 1000000}, nil
}

type jointBackend struct {
	mode    string
	calls   atomic.Int32
	initial Observation
}

func (*jointBackend) Policy(context.Context, string) (control.NativePolicy, error) {
	return control.NativePolicy{InputTokenCeiling: 100, MaxOutputTokens: 100, Timeout: 5 * time.Second}, nil
}
func (b *jointBackend) Complete(_ context.Context, _ string, in control.NativeRequest) (control.NativeResult, error) {
	n := b.calls.Add(1)
	a := action("read_file", "list.go")
	if n > 1 {
		var previous Observation
		if json.Unmarshal([]byte(in.Messages[len(in.Messages)-1].Content), &previous) != nil {
			return control.NativeResult{}, control.ErrContext
		}
		if n == 2 {
			b.initial = previous
		}
		switch n {
		case 2:
			if b.mode == "stale" {
				a = action("run_tests", "")
				a.CandidateDigest = previous.CandidateDigest
			} else {
				a = b.replacement()
			}
		case 3:
			if b.mode == "stale" {
				a = b.replacement()
			} else {
				a = action("run_tests", "")
				a.CandidateDigest = previous.CandidateDigest
			}
		case 4:
			if b.mode == "normal" && (previous.Verification == nil || previous.Verification.Status != "command_passed") {
				return control.NativeResult{}, control.ErrContext
			}
			return control.NativeResult{Message: control.NativeMessage{Role: "assistant", Content: "fixed response complete"}, Usage: &control.NativeUsage{PromptTokens: 20, CompletionTokens: 10, TotalTokens: 30}, FinishReason: "stop"}, nil
		default:
			return control.NativeResult{}, control.ErrContext
		}
	}
	raw, err := json.Marshal(a)
	if err != nil {
		return control.NativeResult{}, err
	}
	return control.NativeResult{Message: control.NativeMessage{Role: "assistant", ToolCalls: []control.NativeCall{{ID: fmt.Sprintf("call_joint_%d", n), Type: "function", Function: control.NativeFunction{Name: "development_tool", Arguments: string(raw)}}}}, Usage: &control.NativeUsage{PromptTokens: 20, CompletionTokens: 10, TotalTokens: 30}, FinishReason: "tool_calls"}, nil
}
func (b *jointBackend) replacement() control.ToolAction {
	a := action("replace_file", "list.go")
	a.ExpectedDigest = b.initial.FileDigest
	a.Content = strings.ReplaceAll(b.initial.Content, "result[:limit-1]", "result[:limit]")
	if b.mode == "cancel" {
		a.Content = "package tasklist\nimport(\"time\";\"net/http\")\ntype Task struct{ID string;State string}\nfunc Handler([]Task)http.Handler{return nil}\nfunc init(){for {time.Sleep(time.Second)}}\n"
	}
	return a
}

type jointFixture struct {
	runtimeClosed bool
	cancelStarted time.Time
	root          string
	db            *sqlitedev.Runtime
	runtime       *control.Runtime
	session       *Session
	identity      control.Identity
	issued        control.Issued
	authority     *jointAuthority
	backend       *jointBackend
	launch        control.WorkerLaunch
	service       *artifact.Service
	server        *httptest.Server
}

func jointDatabase(path string) (*sqlitedev.Runtime, error) {
	return sqlitedev.Open(context.Background(), sqlitedev.Options{DatabasePath: path, Migrations: append(attempts.Migrations(), artifacts.Migrations()...)})
}
func newJointFixture(t *testing.T, mode string) *jointFixture {
	t.Helper()
	root := os.Getenv("RADISHMIND_JOINT_ROOT")
	if root == "" {
		t.Skip("joint SDK/Docker acceptance requires an explicitly authorized Linux controller")
	}
	if os.Geteuid() != 0 {
		t.Fatal("joint controller must own the protected management directory")
	}
	version, err := exec.Command("/opt/docker", "--version").Output()
	if err != nil || !strings.HasPrefix(string(version), "Docker version 29.8.2,") {
		t.Fatal("unexpected Linux Docker CLI", string(version), err)
	}
	dir, err := os.MkdirTemp(root, "case-"+mode+"-")
	if err != nil {
		t.Fatal(err)
	}
	f := &jointFixture{root: dir, identity: control.Identity{TenantRef: "tenant_fixture", AccountID: "usr_fixture", WorkspaceID: "workspace_fixture"}, backend: &jointBackend{mode: mode}}
	f.authority = &jointAuthority{scope: artifact.Scope{TenantRef: f.identity.TenantRef, AccountID: f.identity.AccountID, WorkspaceID: f.identity.WorkspaceID, ApplicationID: "application_fixture"}}
	f.db, err = jointDatabase(filepath.Join(dir, "joint.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		// Stop execution before closing HTTP or SQLite, including failed startup.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if f.runtime != nil && !f.runtimeClosed {
			if err := f.runtime.Close(ctx); err != nil {
				t.Error(err)
			}
		}
		if f.session != nil {
			cleanup, stop := cleanupContext()
			if empty, err := f.session.Close(cleanup); err != nil || !empty {
				t.Errorf("joint sandbox cleanup: %v %v", empty, err)
			}
			stop()
		}
		if f.server != nil {
			f.server.CloseClientConnections()
			f.server.Close()
		}
		if f.launch.Command.Dir != "" {
			if err := os.RemoveAll(f.launch.Command.Dir); err != nil {
				t.Error(err)
			}
		}
		if err := f.db.Close(); err != nil {
			t.Error(err)
		}
	})
	store, err := control.NewSQLiteStore(f.db.DB())
	if err != nil {
		t.Fatal(err)
	}
	f.runtime, err = control.NewRuntime(store, f.authority, f.backend)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(root, "input", "list-filter-plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	pp, err := probe.ParsePlan(raw)
	if err != nil {
		t.Fatal(err)
	}
	messages := []control.NativeMessage{{Role: "system", Content: "Use only the approved development tool."}, {Role: "user", Content: "Repair the frozen list filter fixture."}}
	raw, err = json.Marshal(messages)
	if err != nil {
		t.Fatal(err)
	}
	plan := control.Plan{SchemaVersion: "development_attempt_plan.v1", ApplicationID: f.authority.scope.ApplicationID, ExecutionProfile: "prompt_invoke_v2", ModelRef: "profile:fixture", BaselineDigest: probe.TreeDigest(pp.SourceFiles), InitialMessagesDigest: control.PayloadDigest(raw), Tools: control.ToolPolicy{ReadPaths: []string{"go.mod", "list.go", "list_test.go"}, WritePaths: []string{"list.go"}, VerificationProfile: "go_test_v1"}, Limits: control.Limits{MaxModelCalls: 4, MaxToolCalls: 3, MaxTokens: 1000, MaxCostMicrousd: 1000}, LifetimeMS: 600000, MaxOutputTokens: 50}
	prepared, err := f.runtime.Prepare(context.Background(), f.identity, plan)
	if err != nil {
		t.Fatal(err)
	}
	f.issued, err = f.runtime.Issue(context.Background(), f.identity, plan, "confirmation_joint", prepared.PlanDigest, prepared.Selection)
	if err != nil {
		t.Fatal(err)
	}
	runner := filepath.Join(root, "code-runner")
	body, err := os.ReadFile(runner)
	if err != nil {
		t.Fatal(err)
	}
	f.session, err = New(Config{DockerPath: "/opt/docker", Image: os.Getenv("RADISHMIND_SANDBOX_IMAGE"), RunnerPath: runner, RunnerDigest: control.PayloadDigest(body), Source: filepath.Join(root, "input", "list-filter"), Directory: filepath.Join(dir, "sandbox"), Probe: pp, Plan: plan, Delegation: f.issued.Delegation})
	if err != nil {
		t.Fatal(err)
	}
	f.server = httptest.NewServer(f.runtime.Handler())
	worker, err := os.MkdirTemp("/tmp", "joint-worker-")
	if err != nil {
		t.Fatal(err)
	}
	f.launch = control.WorkerLaunch{Command: control.WorkerCommand{Path: "/usr/local/bin/python", Args: []string{"/opt/joint-worker.py", "/opt/sdk-worker/worker.py", dir}, Dir: worker}, ChannelURL: f.server.URL, InitialMessages: messages, Sandbox: f.session}
	source, err := NewDeliverySource(f.runtime, f.session, f.identity)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := artifact.NewSQLiteStore(f.db.DB())
	if err != nil {
		t.Fatal(err)
	}
	f.service, err = artifact.NewService(saved, f.authority, source)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("joint evidence:", dir)
	return f
}
