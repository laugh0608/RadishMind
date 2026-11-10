package httpapi

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"radishmind.local/services/platform/internal/sqlitedev"
	sqliteworkflowrunmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_runs"
)

// Repository tests need independent databases with the current schema, not a
// repeated empty-to-current upgrade. Migration tests still open empty databases
// directly. The template is checkpointed and closed before its bytes are shared.
var workflowRunSQLiteFixture = sync.OnceValues(func() (contents []byte, err error) {
	directory, err := os.MkdirTemp("", "radishmind-workflow-schema-")
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(directory)) }()
	databasePath := filepath.Join(directory, "schema.db")
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath, Migrations: sqliteworkflowrunmigrations.Migrations(),
	})
	if err != nil {
		return nil, err
	}
	if err := runtime.Close(); err != nil {
		return nil, err
	}
	return os.ReadFile(databasePath)
})

func prepareWorkflowRunSQLiteFixture(t *testing.T, databasePath string) {
	t.Helper()
	if _, err := os.Lstat(databasePath); err == nil {
		// Reopening must validate the existing database, including corruption.
		return
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("inspect workflow SQLite fixture: %v", err)
	}
	contents, err := workflowRunSQLiteFixture()
	if err != nil {
		t.Fatalf("build workflow SQLite schema fixture: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o700); err != nil {
		t.Fatalf("create workflow SQLite fixture directory: %v", err)
	}
	file, err := os.OpenFile(databasePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatalf("create workflow SQLite fixture: %v", err)
	}
	_, writeErr := file.Write(contents)
	if err := errors.Join(writeErr, file.Close()); err != nil {
		t.Fatalf("write workflow SQLite fixture: %v", err)
	}
}

func TestWorkflowRunSQLiteFixtureIsIsolatedAndNeverResetsExistingData(t *testing.T) {
	for _, name := range []string{"first", "second"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "isolated.db")
			runtime := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, path)
			if err := runtime.VerifyMigrations(t.Context(), sqliteworkflowrunmigrations.Migrations()); err != nil {
				t.Fatal(err)
			}
			var version int
			if err := runtime.DB().QueryRow("PRAGMA user_version").Scan(&version); err != nil || version != 0 {
				t.Fatalf("fixture inherited another database's changes: version=%d err=%v", version, err)
			}
			if _, err := runtime.DB().Exec("PRAGMA user_version=37"); err != nil {
				t.Fatal(err)
			}
			if err := runtime.Close(); err != nil {
				t.Fatal(err)
			}
			reopened := openWorkflowRunSQLiteRuntime(t, path)
			if err := reopened.DB().QueryRow("PRAGMA user_version").Scan(&version); err != nil || version != 37 {
				t.Fatalf("fixture reset existing data: version=%d err=%v", version, err)
			}
		})
	}
}

func TestWorkflowRunSQLiteFixtureDoesNotReplaceCorruption(t *testing.T) {
	path := filepath.Join(t.TempDir(), "corrupt.db")
	const corrupted = "not a SQLite database"
	if err := os.WriteFile(path, []byte(corrupted), 0o600); err != nil {
		t.Fatal(err)
	}
	prepareWorkflowRunSQLiteFixture(t, path)
	runtime, err := sqlitedev.Open(t.Context(), sqlitedev.Options{
		DatabasePath: path, Migrations: sqliteworkflowrunmigrations.Migrations(),
	})
	if runtime != nil {
		if closeErr := runtime.Close(); closeErr != nil {
			t.Fatal(closeErr)
		}
	}
	if err == nil {
		t.Fatal("corrupt database was replaced by a valid fixture")
	}
	contents, err := os.ReadFile(path)
	if err != nil || string(contents) != corrupted {
		t.Fatalf("fixture overwrote corrupt database: err=%v", err)
	}
}
