package workbench

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"radishmind.local/services/platform/internal/config"
)

func initializedDatabase(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "data", "radishmind.db")
	if err := Initialize(context.Background(), path); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestInitializeAndDirectoryExclusion(t *testing.T) {
	path := initializedDatabase(t)
	if _, err := Inspect(context.Background(), path, false); err != nil {
		t.Fatal(err)
	}
	if err := Initialize(context.Background(), path); err == nil {
		t.Fatal("overwrote existing database")
	}
	lock, err := Acquire(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if other, err := Acquire(filepath.Dir(path)); err == nil {
		other.Close()
		t.Fatal("allowed simultaneous owner")
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
	other, err := Acquire(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	other.Close()
}

func TestInspectRejectsUnknownChangedAndPendingMigrations(t *testing.T) {
	for _, change := range []string{
		`UPDATE radishmind_schema_migrations SET migration_checksum='changed' WHERE component=(SELECT component FROM radishmind_schema_migrations LIMIT 1)`,
		`UPDATE radishmind_schema_migrations SET component='future' WHERE component=(SELECT component FROM radishmind_schema_migrations LIMIT 1)`,
		`DELETE FROM radishmind_schema_migrations WHERE component=(SELECT component FROM radishmind_schema_migrations LIMIT 1)`,
	} {
		path := initializedDatabase(t)
		db, err := sql.Open("sqlite", path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(change); err != nil {
			t.Fatal(err)
		}
		db.Close()
		if _, err := Inspect(context.Background(), path, false); err == nil {
			t.Fatal("accepted incompatible migration set")
		}
		_, err = Inspect(context.Background(), path, true)
		if strings.HasPrefix(change, "DELETE") != (err == nil) {
			t.Fatalf("unexpected upgrade eligibility: %v", err)
		}
	}
}

func TestColdBackupRestorePreservesDataAndRejectsTampering(t *testing.T) {
	path := initializedDatabase(t)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE recovery_evidence(value TEXT NOT NULL); INSERT INTO recovery_evidence VALUES ('retained result')`); err != nil {
		t.Fatal(err)
	}
	db.Close()
	cfg := config.Config{SQLiteDevDatabasePath: path, Provider: "mock", Model: "mock", APIKey: "must-not-be-backed-up", Workbench: config.WorkbenchConfig{TenantRef: "tenant_local", WorkspaceID: "workspace_local"}}
	backup := filepath.Join(t.TempDir(), "backup")
	lock, err := Acquire(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if err := Backup(context.Background(), cfg, backup, "test"); err == nil {
		t.Fatal("backed up active database")
	}
	lock.Close()
	if err := Backup(context.Background(), cfg, backup, "test"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"platform.json", "manifest.json"} {
		data, err := os.ReadFile(filepath.Join(backup, name))
		if err != nil || strings.Contains(string(data), cfg.APIKey) {
			t.Fatal("secret copied or file missing")
		}
	}
	restored := filepath.Join(t.TempDir(), "restored")
	if _, err := Restore(context.Background(), backup, restored); err != nil {
		t.Fatal(err)
	}
	db, err = openReadOnly(filepath.Join(restored, "data", "radishmind.db"))
	if err != nil {
		t.Fatal(err)
	}
	var value string
	if err := db.QueryRow(`SELECT value FROM recovery_evidence`).Scan(&value); err != nil || value != "retained result" {
		t.Fatal("lost backup data")
	}
	db.Close()
	if _, err := Restore(context.Background(), backup, restored); err == nil {
		t.Fatal("overwrote restore destination")
	}
	if err := os.WriteFile(filepath.Join(backup, "platform.json"), []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Restore(context.Background(), backup, filepath.Join(t.TempDir(), "tampered")); err == nil {
		t.Fatal("restored tampered backup")
	}
}

func TestInspectMissingDoesNotCreateDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.db")
	if _, err := Inspect(context.Background(), path, false); err == nil {
		t.Fatal("accepted missing database")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("inspection created database")
	}
}

func loadedWorkbenchConfig(t *testing.T) config.Config {
	t.Helper()
	for _, entry := range os.Environ() {
		name := strings.SplitN(entry, "=", 2)[0]
		if strings.HasPrefix(name, "RADISHMIND_") {
			t.Setenv(name, "")
			os.Unsetenv(name)
		}
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "platform.json")
	document := `{"runtime_profile":"local-workbench","temperature":0.35,"bridge_timeout":"42s","python_binary":"python3","bridge_worker_count":2}`
	if err := os.WriteFile(path, []byte(document), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("RADISHMIND_PLATFORM_CONFIG", path)
	cfg, err := config.LoadFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if err := Initialize(context.Background(), cfg.SQLiteDevDatabasePath); err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestRestoredConfigurationCanBeLoadedAndKeepsRuntimeSettings(t *testing.T) {
	cfg := loadedWorkbenchConfig(t)
	backup := filepath.Join(t.TempDir(), "backup")
	if err := Backup(context.Background(), cfg, backup, "test-build"); err != nil {
		t.Fatal(err)
	}
	restored := filepath.Join(t.TempDir(), "restored")
	if _, err := Restore(context.Background(), backup, restored); err != nil {
		t.Fatal(err)
	}
	t.Setenv("RADISHMIND_PLATFORM_CONFIG", filepath.Join(restored, "platform.json"))
	actual, err := config.LoadFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if err := config.ValidateServerStart(actual); err != nil {
		t.Fatal(err)
	}
	if actual.SQLiteDevDatabasePath != filepath.Join(restored, "data", "radishmind.db") || actual.Temperature != cfg.Temperature || actual.BridgeTimeout != cfg.BridgeTimeout || actual.BridgeWorkerCount != cfg.BridgeWorkerCount || actual.Workbench != cfg.Workbench {
		t.Fatal("restored configuration changed runtime settings or still points at source database")
	}
	if _, err := Inspect(context.Background(), actual.SQLiteDevDatabasePath, false); err != nil {
		t.Fatal(err)
	}
}

func TestUpgradeBacksUpBeforeMigrationAndPreservesFailedState(t *testing.T) {
	for _, failMigration := range []bool{false, true} {
		t.Run(map[bool]string{false: "known pending", true: "conflicting schema"}[failMigration], func(t *testing.T) {
			cfg := loadedWorkbenchConfig(t)
			db, err := sql.Open("sqlite", cfg.SQLiteDevDatabasePath)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(`DELETE FROM radishmind_schema_migrations WHERE component='application_catalog_records'`); err != nil {
				t.Fatal(err)
			}
			if !failMigration {
				if _, err := db.Exec(`DROP TABLE application_catalog_records`); err != nil {
					t.Fatal(err)
				}
			}
			if err := db.Close(); err != nil {
				t.Fatal(err)
			}
			before, err := Inspect(context.Background(), cfg.SQLiteDevDatabasePath, true)
			if err != nil {
				t.Fatal(err)
			}
			backup := filepath.Join(t.TempDir(), "before-upgrade")
			err = Upgrade(context.Background(), cfg, backup, "upgrade-test")
			if failMigration != (err != nil) {
				t.Fatalf("unexpected migration outcome: %v", err)
			}
			manifest, err := verifyBackup(context.Background(), backup)
			if err != nil {
				t.Fatal(err)
			}
			expectedJSON, _ := json.Marshal(before)
			actualJSON, _ := json.Marshal(manifest.Migrations)
			if string(expectedJSON) != string(actualJSON) {
				t.Fatal("backup was taken after migration")
			}
			if _, err := Inspect(context.Background(), cfg.SQLiteDevDatabasePath, false); (err != nil) != failMigration {
				t.Fatalf("unexpected post-upgrade migration state: %v", err)
			}
			if failMigration {
				restored := filepath.Join(t.TempDir(), "recover-failed-upgrade")
				if _, err := Restore(context.Background(), backup, restored); err != nil {
					t.Fatal(err)
				}
				lock, err := Acquire(filepath.Dir(cfg.SQLiteDevDatabasePath))
				if err != nil {
					t.Fatal("failed upgrade retained the process lock")
				}
				lock.Close()
			}
		})
	}
}

func TestExistingLegacyDatabaseCannotBeAdoptedWhileItsProcessMayBeRunning(t *testing.T) {
	source := initializedDatabase(t)
	directory := t.TempDir()
	path := filepath.Join(directory, "legacy.db")
	if err := copyNewFile(source, path); err != nil {
		t.Fatal(err)
	}
	if lock, err := Acquire(directory); err == nil {
		lock.Close()
		t.Fatal("adopted unmarked legacy directory")
	}
	if err := Initialize(context.Background(), path); err == nil {
		t.Fatal("initialized existing legacy database")
	}
	if _, err := os.Stat(filepath.Join(directory, ".workbench.lock")); !os.IsNotExist(err) {
		t.Fatal("failed init marked legacy data as workbench-owned")
	}
	if _, err := Inspect(context.Background(), path, false); err != nil {
		t.Fatal("modified legacy database")
	}
}
