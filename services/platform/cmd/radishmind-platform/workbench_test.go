package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkbenchRejectsWrongActionFlagsBeforeConfigOrDataAccess(t *testing.T) {
	for _, arguments := range [][]string{
		{}, {"unknown"}, {"check", "--output", "unexpected"}, {"restore", "--config", "unexpected"}, {"serve", "extra"},
	} {
		var output bytes.Buffer
		if runWorkbench(arguments, &output, &output) == 0 {
			t.Fatalf("accepted %v", arguments)
		}
	}
}

func TestWorkbenchCLIInitializesOnlyExplicitlyAndRefusesOverwrite(t *testing.T) {
	for _, entry := range os.Environ() {
		name := strings.SplitN(entry, "=", 2)[0]
		if strings.HasPrefix(name, "RADISHMIND_") {
			t.Setenv(name, "")
			os.Unsetenv(name)
		}
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "platform.json")
	if err := os.WriteFile(path, []byte(`{"runtime_profile":"local-workbench"}`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("RADISHMIND_PLATFORM_CONFIG", path)
	var output bytes.Buffer
	// Missing Web assets must fail check without creating a database or data directory.
	if runWorkbench([]string{"check", "--config", path}, &output, &output) == 0 {
		t.Fatal("check accepted missing build")
	}
	if _, err := os.Stat(filepath.Join(directory, "data")); !os.IsNotExist(err) {
		t.Fatal("check created data")
	}
	output.Reset()
	if runWorkbench([]string{"init", "--config", path}, &output, &output) != 0 {
		t.Fatal(output.String())
	}
	if runWorkbench([]string{"init", "--config", path}, &output, &output) == 0 {
		t.Fatal("init overwrote existing database")
	}
	if runWorkbench([]string{"bootstrap-admin", "--config", path}, &output, &output) == 0 {
		t.Fatal("bootstrap accepted missing identity and audit reference")
	}
}
