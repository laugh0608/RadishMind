package developmentprobe

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func taskFixture(t *testing.T) (Plan, string) {
	t.Helper()
	root, err := filepath.Abs("../../../../scripts/eval/fixtures/development-task/list-filter")
	if err != nil {
		t.Fatal(err)
	}
	plan := Plan{SchemaVersion: PlanSchema, ProbeID: "probe_list_filter", Mode: "fixed_response_only", TaskID: "task_list_filter_v1",
		SDKCandidate: SDKCandidate{Package: "openhands-sdk", Version: "1.59.0"}, WritablePaths: []string{"list.go"},
		VerifyArgv: []string{"go", "test", "-json", "-count=1", "./..."}, Limits: Limits{4, 3, 262144, 65536, 120, 600}}
	for _, name := range []string{"go.mod", "list.go", "list_test.go"} {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		plan.SourceFiles = append(plan.SourceFiles, SourceFile{Path: name, SHA256: Digest(data)})
	}
	return plan, root
}

func TestCanonicalPlanAndStrictJSON(t *testing.T) {
	plan, _ := taskFixture(t)
	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParsePlan(data)
	if err != nil || PlanDigest(parsed) != PlanDigest(plan) {
		t.Fatalf("valid plan: %v", err)
	}
	for name, invalid := range map[string]string{
		"duplicate":         strings.Replace(string(data), `"mode":"fixed_response_only"`, `"mode":"fixed_response_only","mode":"fixed_response_only"`, 1),
		"duplicate_nested":  strings.Replace(string(data), `"package":"openhands-sdk"`, `"package":"openhands-sdk","package":"openhands-sdk"`, 1),
		"unknown":           strings.Replace(string(data), `"mode":`, `"api_key":"secret","mode":`, 1),
		"case_alias":        strings.Replace(string(data), `"mode":`, `"MODE":`, 1),
		"nested_case_alias": strings.Replace(string(data), `"max_tool_calls":`, `"MAX_TOOL_CALLS":`, 1),
		"missing":           strings.Replace(string(data), `"mode":"fixed_response_only",`, ``, 1),
		"null":              strings.Replace(string(data), `"writable_paths":["list.go"]`, `"writable_paths":null`, 1),
		"trailing":          string(data) + `{}`,
		"oversized":         strings.Repeat(" ", 262145),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParsePlan([]byte(invalid)); err == nil {
				t.Fatal("accepted noncanonical JSON")
			}
		})
	}
}

func TestPlanRejectsScopeAndBudgetExpansion(t *testing.T) {
	for name, mutate := range map[string]func(*Plan){
		"remote_mode":       func(p *Plan) { p.Mode = "real_provider" },
		"unknown_sdk":       func(p *Plan) { p.SDKCandidate.Package = "another-framework" },
		"floating_sdk":      func(p *Plan) { p.SDKCandidate.Version = "latest" },
		"unapproved_patch":  func(p *Plan) { p.SDKCandidate.Version = "1.59.0+radishmind.2" },
		"foreign_patch":     func(p *Plan) { p.SDKCandidate.Version = "1.59.0+other.1" },
		"path_escape":       func(p *Plan) { p.SourceFiles[0].Path = "../go.mod" },
		"unsorted_files":    func(p *Plan) { p.SourceFiles[0], p.SourceFiles[1] = p.SourceFiles[1], p.SourceFiles[0] },
		"extra_write":       func(p *Plan) { p.WritablePaths = append(p.WritablePaths, "list_test.go") },
		"arbitrary_command": func(p *Plan) { p.VerifyArgv = []string{"sh", "-c", "go test ./..."} },
		"no_limits":         func(p *Plan) { p.Limits = Limits{} },
		"too_many_calls":    func(p *Plan) { p.Limits.MaxModelRequests = 9 },
		"too_many_tools":    func(p *Plan) { p.Limits.MaxToolCalls = 9 },
		"too_long":          func(p *Plan) { p.Limits.WallTimeoutSeconds = 601 },
		"invalid_digest":    func(p *Plan) { p.SourceFiles[0].SHA256 = "unknown" },
	} {
		t.Run(name, func(t *testing.T) {
			plan, _ := taskFixture(t)
			mutate(&plan)
			if err := plan.Validate(); err == nil {
				t.Fatal("accepted expanded scope")
			}
		})
	}
}

func TestVariantChangesPlanIdentity(t *testing.T) {
	plan, _ := taskFixture(t)
	stock := PlanDigest(plan)
	plan.SDKCandidate.Version = "1.59.0+radishmind.1"
	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParsePlan(data)
	if err != nil || PlanDigest(parsed) == stock {
		t.Fatalf("variant must be valid and separately bound: %v", err)
	}
}

func TestWorkspacePreparationAndCandidateBinding(t *testing.T) {
	plan, source := taskFixture(t)
	output := filepath.Join(t.TempDir(), "new-probe")
	initial, err := Prepare(plan, source, output)
	if err != nil || initial.BaseDigest != initial.CandidateDigest || len(initial.ChangedPaths) != 0 {
		t.Fatalf("prepare: %#v %v", initial, err)
	}
	if _, err := Prepare(plan, source, output); err == nil {
		t.Fatal("overwrote existing preparation")
	}
	path := filepath.Join(output, "candidate", "list.go")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data = []byte(strings.Replace(string(data), "result[:limit-1]", "result[:limit]", 1))
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	updated, err := Inspect(plan, filepath.Join(output, "candidate"))
	if err != nil || updated.CandidateDigest == initial.CandidateDigest || len(updated.ChangedPaths) != 1 || updated.ChangedPaths[0] != "list.go" {
		t.Fatalf("candidate binding: %#v %v", updated, err)
	}
	base, err := Inspect(plan, filepath.Join(output, "base"))
	if err != nil || base.CandidateDigest != initial.CandidateDigest {
		t.Fatalf("original source was changed: %#v %v", base, err)
	}
}

func TestCandidateRejectsProtectedChangesAndNonregularFiles(t *testing.T) {
	for _, mode := range []string{"protected", "extra", "symlink", "directory", "missing", "too_large"} {
		t.Run(mode, func(t *testing.T) {
			plan, source := taskFixture(t)
			output := filepath.Join(t.TempDir(), "probe")
			if _, err := Prepare(plan, source, output); err != nil {
				t.Fatal(err)
			}
			candidate := filepath.Join(output, "candidate")
			path := filepath.Join(candidate, "list.go")
			var err error
			switch mode {
			case "protected":
				path = filepath.Join(candidate, "list_test.go")
				if err = os.Chmod(path, 0o600); err == nil {
					err = os.WriteFile(path, []byte("changed tests"), 0o600)
				}
			case "extra":
				err = os.WriteFile(filepath.Join(candidate, "extra.go"), []byte("extra"), 0o600)
			case "too_large":
				err = os.WriteFile(path, []byte(strings.Repeat("x", maxSourceBytes+1)), 0o600)
			default:
				if err = os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if mode == "symlink" {
					err = os.Symlink(filepath.Join(output, "base", "list.go"), path)
				} else if mode == "directory" {
					err = os.Mkdir(path, 0o700)
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err := Inspect(plan, candidate); err == nil {
				t.Fatal("accepted invalid candidate")
			}
		})
	}
}

func TestChangedSourceCannotCreateOutput(t *testing.T) {
	plan, source := taskFixture(t)
	plan.SourceFiles[0].SHA256 = "sha256:" + strings.Repeat("0", 64)
	output := filepath.Join(t.TempDir(), "should-not-exist")
	if _, err := Prepare(plan, source, output); err == nil {
		t.Fatal("accepted baseline drift")
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatal("created output for an invalid baseline")
	}
}

func TestToolRequestHasNoImplicitAuthorityOrCommands(t *testing.T) {
	plan, _ := taskFixture(t)
	request := ToolRequest{SchemaVersion: "development_task_probe_tool_request.v1", ProbeID: plan.ProbeID, Sequence: 1, Action: "read_file", Path: "list.go"}
	for _, action := range []string{"read_file", "replace_file", "run_tests"} {
		value := request
		value.Action = action
		if action == "replace_file" {
			value.ExpectedDigest, value.Content = plan.SourceFiles[1].SHA256, "package tasklist\n"
		}
		if action == "run_tests" {
			value.Path, value.CandidateDigest = "", TreeDigest(plan.SourceFiles)
		}
		data, _ := json.Marshal(value)
		if _, err := ParseToolRequest(plan, data); err != nil {
			t.Fatalf("valid %s: %v", action, err)
		}
	}
	for name, mutate := range map[string]func(*ToolRequest){
		"cross_probe":           func(r *ToolRequest) { r.ProbeID = "probe_other" },
		"path_escape":           func(r *ToolRequest) { r.Path = "../list.go" },
		"test_write":            func(r *ToolRequest) { r.Action, r.Path = "replace_file", "list_test.go" },
		"no_identity":           func(r *ToolRequest) { r.Action, r.Content = "replace_file", "content" },
		"command":               func(r *ToolRequest) { r.Action = "execute_shell" },
		"extra_content":         func(r *ToolRequest) { r.Content = "not allowed" },
		"over_budget":           func(r *ToolRequest) { r.Sequence = 9 },
		"verify_without_digest": func(r *ToolRequest) { r.Action, r.Path = "run_tests", "" },
	} {
		t.Run(name, func(t *testing.T) {
			value := request
			mutate(&value)
			data, _ := json.Marshal(value)
			if _, err := ParseToolRequest(plan, data); err == nil {
				t.Fatal("accepted unsafe tool request")
			}
		})
	}
	data, _ := json.Marshal(request)
	for _, invalid := range []string{
		strings.Replace(string(data), `"path":"list.go"`, `"path":"list.go","path":"list.go"`, 1),
		strings.Replace(string(data), `"content":"",`, ``, 1),
		strings.Replace(string(data), `"content":""`, `"CONTENT":""`, 1),
	} {
		if _, err := ParseToolRequest(plan, []byte(invalid)); err == nil {
			t.Fatal("accepted noncanonical tool JSON")
		}
	}
}
