package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseWorkflowRAGEvaluationConfigRequiresExplicitSafePaths(t *testing.T) {
	if _, err := parseWorkflowRAGEvaluationConfig(nil, &strings.Builder{}); err == nil {
		t.Fatal("CLI accepted missing paths")
	}
	if _, err := parseWorkflowRAGEvaluationConfig([]string{
		"--snapshot", "snapshot.json", "--dataset", "dataset.json", "--output", "snapshot.json",
	}, &strings.Builder{}); err == nil {
		t.Fatal("CLI accepted an output path that overwrites an input")
	}
	config, err := parseWorkflowRAGEvaluationConfig([]string{
		"--snapshot", "snapshot.json", "--dataset", "dataset.json", "--output", "review.json", "--check",
	}, &strings.Builder{})
	if err != nil || !config.check {
		t.Fatalf("parse valid check config: config=%#v err=%v", config, err)
	}
}

func TestWorkflowRAGEvaluationAtomicFileIO(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "review.json")
	payload := []byte("{\"status\":\"passed\"}\n")
	if err := writeWorkflowRAGEvaluationFile(path, payload); err != nil {
		t.Fatalf("write atomic review: %v", err)
	}
	read, err := readWorkflowRAGEvaluationFile(path)
	if err != nil || string(read) != string(payload) {
		t.Fatalf("read atomic review: payload=%q err=%v", read, err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o644 {
		t.Fatalf("unexpected review permissions: info=%v err=%v", info, err)
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil || len(entries) != 1 {
		t.Fatalf("temporary file was not cleaned up: entries=%v err=%v", entries, err)
	}
}
