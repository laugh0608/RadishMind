//go:build darwin || linux

package developmentsandbox

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	control "radishmind.local/services/platform/internal/developmentcontrol"
)

// This executable models only the inventory/inspection/removal arguments. It
// never connects to Docker and keeps all state beneath the test directory.
const taskDockerFixture = `#!/bin/sh
set -eu
dir=${0%/*}
echo "$*" >> "$dir/commands"
case "$1 $2" in
  "info --format") cat "$dir/daemon";;
  "container ls")
    case "$6" in label=*) if [ -f "$dir/hide-label" ]; then exit 0; fi;; esac
    if [ -f "$dir/present" ]; then cat "$dir/id"; fi;;
  "container inspect") cat "$dir/inspection";;
  "container rm") test "$3" = "--force" && test "$4" = "$(cat "$dir/id")"; mv "$dir/present" "$dir/removed";;
  *) exit 1;;
esac
`

func TestTaskWorkerChecksOwnerDaemonAndContainerBeforeRemoval(t *testing.T) {
	cfg := fixture(t)
	root := t.TempDir()
	if err := os.Chmod(root, 0700); err != nil {
		t.Fatal(err)
	}
	put := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0700); err != nil {
			t.Fatal(err)
		}
	}
	put("docker-fixture", taskDockerFixture)
	put("daemon", "daemon-original-1234")
	d := cfg.Delegation
	d.Scope.RunID = d.Scope.TaskID
	raw, _ := json.Marshal([]control.NativeMessage{{Role: "system", Content: "Use only the approved development tool."}, {Role: "user", Content: "Repair the frozen list filter fixture."}})
	cfg.Plan.InitialMessagesDigest = control.PayloadDigest(raw)
	d.Scope.PlanDigest, _ = cfg.Plan.Digest()
	scope := control.TaskScope{Identity: control.Identity{TenantRef: "tenant_fixture", WorkspaceID: d.Scope.WorkspaceID, AccountID: d.Scope.AccountID}, ApplicationID: d.Scope.ApplicationID}
	o := control.TaskResourceOwner{Scope: scope, TaskID: d.Scope.TaskID, OwnerInstanceID: "owner_fixture", Sandbox: control.SandboxBindingFor(d), SDKImageDigest: cfg.Image, GoImageDigest: cfg.Image, ExecutorDigest: cfg.RunnerDigest}
	p := control.TaskPlan{Revision: 1, TaskProfile: control.TaskProfile{AttemptPlan: cfg.Plan, SourceProfileID: cfg.Probe.ProbeID, SourceVersion: cfg.Probe.TaskID, ExecutionMode: "fixed_response_only", SDKImageDigest: cfg.Image, GoImageDigest: cfg.Image, ExecutorDigest: cfg.RunnerDigest}}
	h := TaskWorkers{Directory: root, DockerPath: filepath.Join(root, "docker-fixture"), GoImageDigest: cfg.Image, SDKImageDigest: cfg.Image, RunnerPath: cfg.RunnerPath, Source: cfg.Source, Probe: cfg.Probe, SDKPath: "/usr/bin/false", ChannelURL: "http://127.0.0.1:23456"}
	launch, cleanup, err := h.Prepare(t.Context(), o, p, d)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := cleanup(); err != nil {
			t.Error(err)
		}
	})
	s := launch.Sandbox.(*Session)
	id := strings.Repeat("d", 64)
	put("id", id)
	put("present", "yes")
	name := "radishmind-code-" + s.ledger.SessionID + "-request_run"
	s.ledger.Resources = []resource{{RequestID: "request_run", Name: name, ContainerID: id, CandidateDigest: o.Sandbox.BaselineDigest, State: "dispatched"}}
	if err = s.save(); err != nil {
		t.Fatal(err)
	}
	inspection := map[string]any{"Id": id, "Name": "/" + name, "Image": cfg.Image, "Config": map[string]any{"Labels": map[string]string{ownerLabel: s.ledger.SessionID, attemptLabel: o.Sandbox.AttemptID}}}
	setInspect := func() {
		raw, err := json.Marshal([]any{inspection})
		if err != nil {
			t.Fatal(err)
		}
		put("inspection", string(raw))
	}
	setInspect()
	refuse := func(owner control.TaskResourceOwner) {
		t.Helper()
		if ok, err := h.RecoverCode(t.Context(), owner); err == nil || ok {
			t.Fatal("accepted wrong or missing ownership")
		}
		if _, err := os.Stat(filepath.Join(root, "present")); err != nil {
			t.Fatal("touched foreign resource", err)
		}
	}
	foreign := o
	foreign.OwnerInstanceID = "owner_elsewhere"
	refuse(foreign)
	foreign = o
	foreign.Scope.TenantRef = "tenant_elsewhere"
	refuse(foreign)
	foreign = o
	foreign.SDKImageDigest = control.PayloadDigest([]byte("other sdk"))
	refuse(foreign)
	put("daemon", "daemon-different-1234")
	refuse(o)
	put("daemon", "daemon-original-1234")
	inspection["Id"] = strings.Repeat("e", 64)
	setInspect()
	refuse(o)
	inspection["Id"] = id
	inspection["Image"] = control.PayloadDigest([]byte("foreign image"))
	setInspect()
	refuse(o)
	inspection["Image"] = cfg.Image
	setInspect()
	put("hide-label", "yes")
	refuse(o)
	if err := os.Remove(filepath.Join(root, "hide-label")); err != nil {
		t.Fatal(err)
	}
	old := s.ledger.Binding
	s.ledger.Binding.ExecutorInstanceID = "executor_other"
	if err = s.save(); err != nil {
		t.Fatal(err)
	}
	refuse(o)
	s.ledger.Binding = old
	if err = s.save(); err != nil {
		t.Fatal(err)
	}
	if ok, err := h.RecoverCode(t.Context(), o); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if ok, err := h.RecoverCode(context.Background(), o); err != nil || !ok {
		t.Fatal("repeat inspection", err)
	}
	if _, err = os.Stat(filepath.Join(root, "present")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("owned resource remained")
	}
	commands, err := os.ReadFile(filepath.Join(root, "commands"))
	if err != nil || strings.Count(string(commands), "container rm --force "+id) != 1 || strings.Contains(string(commands), "container start") {
		t.Fatal(string(commands), err)
	}
}
