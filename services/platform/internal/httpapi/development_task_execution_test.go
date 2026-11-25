//go:build darwin || linux

package httpapi

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

	control "radishmind.local/services/platform/internal/developmentcontrol"
	"radishmind.local/services/platform/internal/developmentsandbox"
)

// Use only the worker's stdlib wire protocol, without importing the SDK or
// connecting to Docker. The same private channel carries all three tool calls.
const fixedTaskWireDriver = `
import json, runpy, sys
worker = runpy.run_path(sys.argv[1])
config = worker["read_input"]()
bridge = worker["Bridge"](config)
bridge.emit("ready")
messages = config["initial_messages"].copy()
for n in range(1,5):
    request_id = "model_" + str(n)
    result = bridge.request("POST", "/chat/completions", dict(model=config["model"],
        messages=messages, max_completion_tokens=config["max_output_tokens"], tools=config["tools"]), request_id)
    answer = result["choices"][0]["message"]
    if "tool_calls" not in answer:
        bridge.emit("finished")
        break
    call = answer["tool_calls"][0]
    offer = bridge.request("GET", "/tool-offers/" + call["id"])
    action = json.loads(call["function"]["arguments"])
    bridge.pending = dict(schema_version="development_tool_call.v1", attempt_id=config["attempt_id"],
        request_id=offer["request_id"], parent_request_id=request_id, tool_call_id=call["id"], arguments=action)
    observation = bridge.observation(action)
    messages.extend([answer, dict(role="tool", name="development_tool", tool_call_id=call["id"], content=observation)])
`

type taskHTTPEnvironment struct {
	available        atomic.Bool
	launches         atomic.Int32
	command          control.WorkerCommand
	address, source  string
	started, release chan struct{}
}

func (e *taskHTTPEnvironment) Check(context.Context) error {
	if !e.available.Load() {
		return control.ErrUnavailable
	}
	return nil
}
func (e *taskHTTPEnvironment) Prepare(_ context.Context, o control.TaskResourceOwner, _ control.TaskPlan, _ control.Delegation) (control.WorkerLaunch, func() error, error) {
	e.launches.Add(1)
	return control.WorkerLaunch{Command: e.command, ChannelURL: e.address, InitialMessages: []control.NativeMessage{{Role: "system", Content: "Use only the approved development tool."}, {Role: "user", Content: "Repair the frozen list filter fixture."}}, Sandbox: &taskHTTPTool{env: e, binding: o.Sandbox, candidate: o.Sandbox.BaselineDigest}}, nil, nil
}

type taskHTTPTool struct {
	env       *taskHTTPEnvironment
	binding   control.SandboxBinding
	candidate string
}

func (s *taskHTTPTool) Binding() control.SandboxBinding     { return s.binding }
func (s *taskHTTPTool) Close(context.Context) (bool, error) { return true, nil }
func (s *taskHTTPTool) HandleTool(ctx context.Context, c control.ToolCall) (string, error) {
	o := developmentsandbox.Observation{SchemaVersion: "development_sandbox_observation.v1", Action: c.Arguments.Action, Path: c.Arguments.Path, CandidateDigest: s.candidate}
	switch o.Action {
	case "read_file":
		o.Content = s.env.source
		o.FileDigest = control.PayloadDigest([]byte(o.Content))
	case "replace_file":
		if c.Arguments.ExpectedDigest != control.PayloadDigest([]byte(s.env.source)) || !strings.Contains(c.Arguments.Content, "result[:limit]") {
			return "", control.ErrContext
		}
		s.candidate = control.PayloadDigest([]byte(c.Arguments.Content))
		o.CandidateDigest = s.candidate
	case "run_tests":
		if c.Arguments.CandidateDigest != s.candidate {
			return "", control.ErrContext
		}
		close(s.env.started)
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-s.env.release:
		}
		o.Verification = &developmentsandbox.Verification{Status: "command_passed"}
	default:
		return "", control.ErrContext
	}
	raw, err := json.Marshal(o)
	return string(raw), err
}

func TestDevelopmentTaskHTTPExecution(t *testing.T) {
	for _, cancel := range []bool{false, true} {
		t.Run(fmt.Sprintf("cancel=%v", cancel), func(t *testing.T) {
			f := newTaskWeb(t)
			// Keep the administrator as the granting actor. Requests and app
			// ownership belong to a separate member with only workspace_developer.
			_, cookies, member := registerLocalIdentityHTTPTestAccount(t, &localIdentityHTTPTestFixture{handler: f.s.httpServer.Handler}, "task-owner@example.test", "another sufficiently long password")
			_, err := f.s.localIdentityAdministrationService.CreateWorkspaceMembership(t.Context(), f.actor, LocalIdentityCreateWorkspaceMembershipInput{TenantRef: f.actor.TenantRef, WorkspaceID: f.actor.WorkspaceID, UserID: member.Account.UserID, AuditRef: "audit:task-owner-member"})
			if err != nil {
				t.Fatal(err)
			}
			f.cookies, f.user, f.sessionID = cookies, member.Account.UserID, member.Session.SessionID
			f.csrf = localIdentityCookieValue(t, cookies, f.s.localIdentityHTTPService.csrfCookieName())
			grant := f.grant()
			f.app = f.createApp()
			created := taskWebJSON[developmentTaskDetail](t, f.call("POST", "development-tasks", taskCreateJSON, 201))
			path := "development-tasks/" + created.Task.TaskID
			prepBody := fmt.Sprintf(`{"request_id":"prepare_http","expected_version":1,"plan_digest":%q}`, created.Plan.Digest)
			f.call("POST", path+"/preparations", prepBody, 503)
			python, err := exec.LookPath("python3")
			if err != nil {
				t.Fatal(err)
			}
			worker, err := filepath.Abs("../../../runtime/development_sdk/worker.py")
			if err != nil {
				t.Fatal(err)
			}
			source, err := os.ReadFile("../../../../scripts/eval/fixtures/development-task/list-filter/list.go")
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			if err := os.Chmod(dir, 0700); err != nil {
				t.Fatal(err)
			}
			server := httptest.NewServer(f.s.developmentTaskRuntime.Handler())
			defer server.Close()
			env := &taskHTTPEnvironment{command: control.WorkerCommand{Path: python, Args: []string{"-I", "-c", fixedTaskWireDriver, worker}, Dir: dir}, address: server.URL, source: string(source), started: make(chan struct{}), release: make(chan struct{})}
			env.available.Store(true)
			f.s.taskEnvironment = env
			before := taskWebJSON[developmentTaskDetail](t, f.call("GET", path, "", 200))
			if !before.Eligibility["prepare"].Allowed || before.Eligibility["start"].Reason != "preparation_required" {
				t.Fatal(before.Eligibility)
			}
			f.call("POST", path+"/preparations", strings.Replace(prepBody, "expected_version", "Expected_Version", 1), 400)
			req := f.request("POST", path+"/preparations", prepBody)
			req.Header.Del(localIdentityCSRFHeader)
			f.serve(req, 403)
			p := taskWebJSON[control.TaskPreparation](t, f.call("POST", path+"/preparations", prepBody, 200))
			if p.Plan.Digest != created.Plan.Digest || p.ExpiresAtMS <= p.PreparedAtMS {
				t.Fatal(p)
			}
			startBody := fmt.Sprintf(`{"request_id":"start_http","expected_version":1,"preparation_id":%q,"confirmation_digest":%q}`, p.PreparationID, p.ConfirmationDigest)
			env.available.Store(false)
			f.call("POST", path+"/attempts", startBody, 503)
			env.available.Store(true)
			f.call("POST", path+"/attempts", strings.Replace(startBody, p.ConfirmationDigest, control.PayloadDigest(nil), 1), 409)
			ref := taskWebJSON[map[string]string](t, f.call("POST", path+"/attempts", startBody, 202))
			if len(ref) != 2 || ref["task_id"] != p.TaskID || ref["attempt_id"] == "" {
				t.Fatal(ref)
			}
			attempt := path + "/attempts/" + ref["attempt_id"]
			select {
			case <-env.started:
			case <-time.After(8 * time.Second):
				t.Fatalf("worker did not reach verification: %s", f.call("GET", attempt, "", 200))
			}
			v := taskWebJSON[control.TaskAttemptView](t, f.call("GET", attempt, "", 200))
			if v.Tool.State != "active" || v.Tool.Sequence != 3 || v.Tool.Name != "run_tests" || v.Tool.Candidate == fixedTaskBaseline || v.Tool.Candidate == "" || v.Tool.Started <= 0 {
				t.Fatal(v)
			}
			env.available.Store(false)
			f.call("POST", path+"/attempts", startBody, 202)
			if cancel {
				body := fmt.Sprintf(`{"request_id":"cancel_http","expected_version":%d}`, v.Version)
				f.call("POST", attempt+"/cancel", strings.Replace(body, fmt.Sprint(v.Version), "999999", 1), 409)
				result := taskWebJSON[struct {
					Requested bool `json:"cancellation_requested"`
				}](t, f.call("POST", attempt+"/cancel", body, 202))
				if !result.Requested {
					t.Fatal("missing accepted cancellation")
				}
				f.call("POST", attempt+"/cancel", body, 202)
			} else {
				close(env.release)
			}
			deadline := time.Now().Add(8 * time.Second)
			for {
				v = taskWebJSON[control.TaskAttemptView](t, f.call("GET", attempt, "", 200))
				if v.ResourceCheck == "confirmed" {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal(v)
				}
				time.Sleep(20 * time.Millisecond)
			}
			if !cancel && (v.Phase != "completed" || v.Budget.ModelReservations != 4 || v.Budget.ToolReservations != 3) || cancel && (v.Phase != "outcome_unknown" || v.Budget.Unresolved != 1) || env.launches.Load() != 1 || v.Tool.State != "idle" {
				t.Fatal(v, env.launches.Load())
			}
			f.s.Close()
			f.open()
			f.call("POST", path+"/attempts", startBody, 202)
			f.call("POST", path+"/preparations", prepBody, 200)
			f.call("GET", attempt, "", 200)
			old := f.app
			f.app = f.createApp()
			f.call("GET", attempt, "", 404)
			f.app = old
			_, err = f.s.localIdentityAdministrationService.RevokeWorkspaceRole(t.Context(), f.actor, LocalIdentityRevokeWorkspaceRoleInput{TenantRef: f.actor.TenantRef, WorkspaceID: f.actor.WorkspaceID, AssignmentID: grant.AssignmentID, ExpectedVersion: grant.RecordVersion, Confirmed: true, AuditRef: "audit:execution-revoked"})
			if err != nil {
				t.Fatal(err)
			}
			f.call("POST", path+"/attempts", startBody, 403)
		})
	}
}

func TestFixedTaskLauncherMatchesProvisionedSource(t *testing.T) {
	raw, err := taskInputFile("../../../runtime/development_sdk/workbench_launcher.py", 64<<10)
	if err != nil || control.PayloadDigest(raw) != fixedTaskLauncher {
		t.Fatal("frozen launcher digest differs from the shipped source", err)
	}
}
