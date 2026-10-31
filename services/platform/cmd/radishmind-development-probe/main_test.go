package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestPreparationCLI(t *testing.T) {
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	fixture := filepath.Join(root, "scripts/eval/fixtures/development-task")
	plan := filepath.Join(fixture, "list-filter-plan.json")
	output := filepath.Join(t.TempDir(), "probe")
	var stdout, stderr bytes.Buffer
	for _, args := range [][]string{
		{"check", "--plan", plan},
		{"prepare", "--plan", plan, "--source", filepath.Join(fixture, "list-filter"), "--output", output},
		{"inspect", "--plan", plan, "--candidate", filepath.Join(output, "candidate")},
	} {
		stdout.Reset()
		if err := run(args, &stdout, &stderr); err != nil || stdout.Len() == 0 {
			t.Fatalf("%v: %v, stderr=%s", args, err, &stderr)
		}
	}
	before, err := os.ReadFile(filepath.Join(output, "workspace.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		nil,
		{"serve", "--plan", plan},
		{"check", "--plan", plan, "--source", fixture},
		{"check", "--plan", plan, "unexpected"},
		{"prepare", "--plan", plan},
		{"prepare", "--plan", plan, "--source", filepath.Join(fixture, "list-filter"), "--output", output},
		{"inspect", "--plan", plan},
		{"check", "--plan", filepath.Join(output, "missing.json")},
	} {
		stdout.Reset()
		if err := run(args, &stdout, &stderr); err == nil || stdout.Len() != 0 {
			t.Fatalf("accepted invalid action %v: %v", args, err)
		}
	}
	after, err := os.ReadFile(filepath.Join(output, "workspace.json"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("failed command changed prepared metadata")
	}
}

func TestPlanFileMustBeBounded(t *testing.T) {
	root := t.TempDir()
	oversized := filepath.Join(root, "large.json")
	if err := os.WriteFile(oversized, bytes.Repeat([]byte(" "), 262145), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{root, oversized} {
		var stdout, stderr bytes.Buffer
		if err := run([]string{"check", "--plan", path}, &stdout, &stderr); err == nil || stdout.Len() != 0 {
			t.Fatalf("accepted nonregular or oversized input %s: %v", path, err)
		}
	}
}
