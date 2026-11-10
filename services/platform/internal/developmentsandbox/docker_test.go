package developmentsandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	control "radishmind.local/services/platform/internal/developmentcontrol"
	probe "radishmind.local/services/platform/internal/developmentprobe"
)

func dockerFixture(t *testing.T) *Session {
	t.Helper()
	docker, image, runner, root := os.Getenv("RADISHMIND_SANDBOX_DOCKER"), os.Getenv("RADISHMIND_SANDBOX_IMAGE"), os.Getenv("RADISHMIND_SANDBOX_RUNNER"), os.Getenv("RADISHMIND_SANDBOX_CASES")
	if docker == "" || image == "" || runner == "" || root == "" {
		t.Skip("real Docker sandbox requires explicit isolated runtime window")
	}
	config := fixture(t)
	body, err := os.ReadFile(runner)
	if err != nil {
		t.Fatal(err)
	}
	directory, err := os.MkdirTemp(root, "case-")
	if err != nil {
		t.Fatal(err)
	}
	config.Directory = filepath.Join(directory, "sandbox")
	config.DockerPath = docker
	config.Image = image
	config.RunnerPath = runner
	config.RunnerDigest = control.PayloadDigest(body)
	s, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if empty, err := s.Close(ctx); err != nil || !empty {
			t.Errorf("sandbox cleanup: %v %v", empty, err)
		}
	})
	t.Log("resource ledger:", filepath.Join(config.Directory, "resources.json"))
	return s
}
func TestDockerSandboxFrozenTask(t *testing.T) {
	s := dockerFixture(t)
	initial := invoke(t, s, call(action("read_file", "list.go"), "tool_read"))
	verify := action("run_tests", "")
	verify.CandidateDigest = initial.CandidateDigest
	baseline := invoke(t, s, call(verify, "tool_baseline"))
	if baseline.Verification == nil || baseline.Verification.Status != "command_failed" || !strings.Contains(baseline.Verification.Output, "filter_before_limit") || !strings.Contains(baseline.Verification.Output, "limit_two") {
		t.Fatalf("expected frozen baseline failures: %+v", baseline)
	}
	replace := action("replace_file", "list.go")
	replace.ExpectedDigest = initial.FileDigest
	replace.Content = strings.ReplaceAll(initial.Content, "result[:limit-1]", "result[:limit]")
	candidate := invoke(t, s, call(replace, "tool_replace"))
	if _, err := s.HandleTool(context.Background(), call(verify, "tool_stale")); err == nil {
		t.Fatal("stale candidate ran")
	}
	verify.CandidateDigest = candidate.CandidateDigest
	result := invoke(t, s, call(verify, "tool_candidate"))
	if result.Verification.Status != "command_passed" || result.Verification.ExitCode != 0 || result.CandidateDigest != candidate.CandidateDigest {
		t.Fatalf("candidate: %+v", result)
	}
	count := 0
	for _, line := range strings.Split(result.Verification.Output, "\n") {
		var event struct{ Action, Test string }
		if json.Unmarshal([]byte(line), &event) == nil && event.Action == "pass" && strings.Contains(event.Test, "/") {
			count++
		}
	}
	if count != 8 {
		t.Fatalf("expected 8 frozen subtests, got %d: %s", count, result.Verification.Output)
	}
	source, err := probe.Inspect(s.config.Probe, s.config.Source)
	if err != nil || len(source.ChangedPaths) != 0 {
		t.Fatal("source changed", err)
	}
	ctx, cancel := cleanupContext()
	defer cancel()
	if empty, err := s.Close(ctx); err != nil || !empty {
		t.Fatal(empty, err)
	}
}
func installSource(t *testing.T, s *Session, source string) {
	t.Helper()
	old := invoke(t, s, call(action("read_file", "list.go"), "tool_initial"))
	replace := action("replace_file", "list.go")
	replace.ExpectedDigest = old.FileDigest
	replace.Content = source
	invoke(t, s, call(replace, "tool_install"))
}
func TestDockerSandboxIsolationAndResourceLimits(t *testing.T) {
	t.Setenv("RADISHMIND_SANDBOX_SECRET", "synthetic-host-only")
	s := dockerFixture(t)
	// The protected test source stays frozen. A candidate init checks real OS
	// boundaries, then exits nonzero deliberately: this is never a passing task.
	installSource(t, s, `package tasklist
import("os";"net";"net/http";"fmt";"time";"os/exec";"bytes")
type Task struct {ID string;State string}
func Handler([]Task)http.Handler{return nil}
func init(){
 for _,p:=range []string{"/workspace/go.mod","/workspace/list.go","/workspace/list_test.go","/opt/code-runner","/root/forbidden"}{if os.WriteFile(p,[]byte("bad"),0600)==nil{panic("write allowed: "+p)}}
 if _,err:=os.Stat("/var/run/docker.sock");err==nil{panic("docker socket")}
 if os.Getenv("RADISHMIND_SANDBOX_SECRET")!=""{panic("secret inherited")}
 conn,err:=net.DialTimeout("tcp","1.1.1.1:443",time.Second);if err==nil{conn.Close();panic("network allowed")}
 file,err:=os.Create("/tmp/disk-limit");if err!=nil{panic(err)}
 chunk:=bytes.Repeat([]byte("x"),1<<20);blocked:=false
 for i:=0;i<600;i++{if _,err=file.Write(chunk);err!=nil{blocked=true;break}}
 file.Close();os.Remove("/tmp/disk-limit");if !blocked{panic("disk limit missing")}
 children:=[]*exec.Cmd{};blocked=false
 for i:=0;i<160;i++{child:=exec.Command("/bin/sleep","30");if err=child.Start();err!=nil{blocked=true;break};children=append(children,child)}
 for _,child:=range children{child.Process.Kill();child.Wait()};if !blocked{panic("pids limit missing")}
 fmt.Println("BOUNDARIES_CONFIRMED");os.Exit(1)
}`)
	current, err := s.inspect()
	if err != nil {
		t.Fatal(err)
	}
	verify := action("run_tests", "")
	verify.CandidateDigest = current.CandidateDigest
	result := invoke(t, s, call(verify, "tool_boundaries"))
	if result.Verification.Status != "command_failed" || !strings.Contains(result.Verification.Output, "BOUNDARIES_CONFIRMED") {
		t.Fatalf("boundaries not observed: %+v", result)
	}
}
func TestDockerSandboxCancellationAndRecovery(t *testing.T) {
	s := dockerFixture(t)
	installSource(t, s, "package tasklist\nimport(\"time\";\"net/http\")\ntype Task struct{ID string;State string}\nfunc Handler([]Task)http.Handler{return nil}\nfunc init(){for {time.Sleep(time.Second)}}\n")
	current, err := s.inspect()
	if err != nil {
		t.Fatal(err)
	}
	verify := action("run_tests", "")
	verify.CandidateDigest = current.CandidateDigest
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := s.HandleTool(ctx, call(verify, "tool_wait")); done <- err }()
	until := time.Now().Add(15 * time.Second)
	for {
		check, stop := context.WithTimeout(context.Background(), time.Second)
		ids, err := s.owned(check)
		stop()
		running := false
		if err == nil && len(ids) == 1 {
			check, stop = context.WithTimeout(context.Background(), time.Second)
			item, e := s.inspectContainer(check, ids[0])
			stop()
			running = e == nil && item.State.Running
		}
		if running {
			break
		}
		if time.Now().After(until) {
			t.Fatal("container not running")
		}
		time.Sleep(50 * time.Millisecond)
	}
	stopped := time.Now()
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("cancelled command succeeded")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("cancel exceeded cleanup deadline")
	}
	check, stop := cleanupContext()
	defer stop()
	empty, err := CleanupRecovered(check, s.config.DockerPath, s.config.Directory)
	if err != nil || !empty {
		t.Fatal("explicit recovery", empty, err)
	}
	t.Log("cancel and empty resource reconciliation:", time.Since(stopped))
}
func TestDockerSandboxWatchdog(t *testing.T) {
	s := dockerFixture(t)
	// Shorten only this admitted test profile. The helper still has no arbitrary
	// command entry point. No host cancellation or cleanup causes this exit.
	s.config.Probe.Limits.CommandTimeoutSeconds = 1
	current, err := s.inspect()
	if err != nil {
		t.Fatal(err)
	}
	verify := action("run_tests", "")
	verify.CandidateDigest = current.CandidateDigest
	result := invoke(t, s, call(verify, "tool_watchdog"))
	if result.Verification.Status != "timeout" || result.Verification.ExitCode != 124 {
		t.Fatalf("watchdog: %+v", result)
	}
}

func TestDockerSandboxOrphanRecovery(t *testing.T) {
	s := dockerFixture(t)
	snapshot, err := s.snapshot("tool_orphan")
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := s.inspect()
	if err != nil {
		t.Fatal(err)
	}
	name := "radishmind-code-" + s.ledger.SessionID + "-tool_orphan"
	s.ledger.Resources = append(s.ledger.Resources, resource{RequestID: "tool_orphan", Name: name, CandidateDigest: candidate.CandidateDigest, State: "intent"})
	if err = s.save(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := cleanupContext()
	defer cancel()
	raw, err := s.docker(ctx, s.createArgs(name, snapshot)...)
	if err != nil {
		t.Fatal(err)
	}
	id := strings.TrimSpace(string(raw))
	if !containerPattern.MatchString(id) {
		t.Fatal("container identity")
	}
	s.ledger.Resources[0].ContainerID = id
	s.ledger.Resources[0].State = "created"
	if err = s.save(); err != nil {
		t.Fatal(err)
	}
	item, err := s.inspectContainer(ctx, id)
	if err != nil || s.verifyContainer(item, snapshot, name) != nil {
		t.Fatal("container scope", err)
	}
	if _, err = s.docker(ctx, "container", "start", id); err != nil {
		t.Fatal(err)
	}
	item, err = s.inspectContainer(ctx, id)
	if err != nil || !item.State.Running {
		t.Fatal("orphan not running", err)
	}
	// A separate object reads the on-disk ledger. No SDK session, original
	// in-memory resource list or command replay participates in recovery.
	stopped := time.Now()
	if empty, err := CleanupRecovered(ctx, s.config.DockerPath, s.config.Directory); err != nil || !empty {
		t.Fatal("orphan cleanup", empty, err)
	}
	ids, err := s.owned(ctx)
	if err != nil || len(ids) != 0 {
		t.Fatal("orphan remains", err)
	}
	t.Log("explicit orphan reconciliation:", time.Since(stopped))
}
func TestDockerSandboxOutputLimitIsNotPass(t *testing.T) {
	for _, direct := range []bool{false, true} {
		t.Run(fmt.Sprint(direct), func(t *testing.T) {
			s := dockerFixture(t)
			emit := `fmt.Println(strings.Repeat("synthetic-output-",6000));os.Exit(0)`
			if direct {
				emit = `file,err:=os.OpenFile("/proc/1/fd/1",os.O_WRONLY,0);if err!=nil{panic(err)};fmt.Fprintln(file,strings.Repeat("synthetic-direct-",20000));file.Close();os.Exit(0)`
			}
			installSource(t, s, `package tasklist
import("os";"fmt";"strings";"net/http")
type Task struct{ID string;State string}
func Handler([]Task)http.Handler{return nil}
func init(){`+emit+`}
`)
			candidate, err := s.inspect()
			if err != nil {
				t.Fatal(err)
			}
			verify := action("run_tests", "")
			verify.CandidateDigest = candidate.CandidateDigest
			result := invoke(t, s, call(verify, "tool_output"))
			if result.Verification.Status != "output_truncated" || !result.Verification.OutputTruncated || len(result.Verification.Output) > 8192 {
				t.Fatalf("truncated output passed: %+v", result.Verification)
			}
		})
	}
}
