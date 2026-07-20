package sqlitedev

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestRuntimeCreatesMigratesReopensAndCloses(t *testing.T) {
	ctx := context.Background()
	databasePath := filepath.Join(t.TempDir(), "sqlite-dev", "radishmind.db")
	migrations := testMigrations()

	first, err := Open(ctx, Options{DatabasePath: databasePath, BusyTimeout: 1200 * time.Millisecond, Migrations: migrations})
	if err != nil {
		t.Fatalf("open first SQLite development runtime: %v", err)
	}
	assertRuntimeSettings(t, first.DB(), 1200)
	if _, err := first.DB().ExecContext(ctx, `INSERT INTO test_records (record_id, payload) VALUES (?, ?)`, "record-1", `{"safe":true}`); err != nil {
		t.Fatalf("insert test record: %v", err)
	}
	assertMigrationMarkerCount(t, first.DB(), len(migrations))
	assertSQLiteFilePermissions(t, databasePath)
	if err := first.Close(); err != nil {
		t.Fatalf("close first SQLite development runtime: %v", err)
	}
	assertSQLiteCheckpointed(t, databasePath)
	if err := first.Close(); err != nil {
		t.Fatalf("close first SQLite development runtime twice: %v", err)
	}

	second, err := Open(ctx, Options{DatabasePath: databasePath, BusyTimeout: 1200 * time.Millisecond, Migrations: migrations})
	if err != nil {
		t.Fatalf("reopen SQLite development runtime: %v", err)
	}
	t.Cleanup(func() { _ = second.Close() })
	var payload string
	if err := second.DB().QueryRowContext(ctx, `SELECT payload FROM test_records WHERE record_id = ?`, "record-1").Scan(&payload); err != nil {
		t.Fatalf("read recovered test record: %v", err)
	}
	if payload != `{"safe":true}` {
		t.Fatalf("unexpected recovered payload: %s", payload)
	}
	assertMigrationMarkerCount(t, second.DB(), len(migrations))

	assertSQLiteFilePermissions(t, databasePath)
}

func TestRuntimeRejectsChangedAndUnexpectedMigrationMarkers(t *testing.T) {
	ctx := context.Background()
	databasePath := filepath.Join(t.TempDir(), "radishmind.db")
	migrations := testMigrations()
	runtime, err := Open(ctx, Options{DatabasePath: databasePath, Migrations: migrations})
	if err != nil {
		t.Fatalf("open SQLite development runtime: %v", err)
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("close SQLite development runtime: %v", err)
	}

	changed := append([]Migration(nil), migrations...)
	changed[0].UpSQL += "\nCREATE INDEX changed_test_records_idx ON test_records(payload);"
	if _, err := Open(ctx, Options{DatabasePath: databasePath, Migrations: changed}); err == nil || err.Error() != "SQLite development migration marker mismatch" {
		t.Fatalf("changed migration must fail with marker mismatch, got %v", err)
	}

	database := openRawDatabase(t, databasePath)
	if _, err := database.ExecContext(ctx, `INSERT INTO radishmind_schema_migrations
        (component, migration_id, store_schema_version, migration_checksum, applied_at)
        VALUES (?, ?, ?, ?, ?)`, "unknown_component", "0001_unknown", "unknown_store_v1", "sha256:unknown", time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		t.Fatalf("insert unexpected migration marker: %v", err)
	}
	if err := database.Close(); err != nil {
		t.Fatalf("close raw SQLite database: %v", err)
	}
	if _, err := Open(ctx, Options{DatabasePath: databasePath, Migrations: migrations}); err == nil || err.Error() != "SQLite development database contains an unexpected migration marker" {
		t.Fatalf("unexpected migration marker must fail closed, got %v", err)
	}
}

func TestRuntimeVerifiesRequestedComponentMigrations(t *testing.T) {
	ctx := context.Background()
	migrations := testMigrations()
	runtime, err := Open(ctx, Options{DatabasePath: filepath.Join(t.TempDir(), "radishmind.db"), Migrations: migrations})
	if err != nil {
		t.Fatalf("open SQLite development runtime: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })
	if err := runtime.VerifyMigrations(ctx, migrations); err != nil {
		t.Fatalf("verify applied component migrations: %v", err)
	}
	missing := []Migration{{
		Component:          "missing_records",
		ID:                 "0001_missing_records",
		StoreSchemaVersion: "missing_records_store_v1",
		UpSQL:              "CREATE TABLE missing_records (record_id TEXT PRIMARY KEY);",
	}}
	if err := runtime.VerifyMigrations(ctx, missing); err == nil || err.Error() != "SQLite development component migration is not applied" {
		t.Fatalf("missing component migration must fail verification, got %v", err)
	}
	if _, err := runtime.DB().ExecContext(ctx, `UPDATE radishmind_schema_migrations SET migration_checksum=?
        WHERE component=? AND migration_id=?`, "sha256:changed", migrations[0].Component, migrations[0].ID); err != nil {
		t.Fatalf("change component migration marker: %v", err)
	}
	if err := runtime.VerifyMigrations(ctx, migrations); err == nil || err.Error() != "SQLite development component migration marker mismatch" {
		t.Fatalf("changed component migration must fail verification, got %v", err)
	}
	if _, err := runtime.DB().ExecContext(ctx, `UPDATE radishmind_schema_migrations SET migration_checksum=?
        WHERE component=? AND migration_id=?`, migrations[0].Checksum(), migrations[0].Component, migrations[0].ID); err != nil {
		t.Fatalf("restore component migration marker: %v", err)
	}
	if _, err := runtime.DB().ExecContext(ctx, `INSERT INTO radishmind_schema_migrations
        (component, migration_id, store_schema_version, migration_checksum, applied_at)
        VALUES (?, ?, ?, ?, ?)`, migrations[0].Component, "0003_unexpected", "test_records_store_v3", "sha256:unexpected", time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		t.Fatalf("insert unexpected component migration marker: %v", err)
	}
	if err := runtime.VerifyMigrations(ctx, migrations); err == nil || err.Error() != "SQLite development component contains an unexpected migration marker" {
		t.Fatalf("unexpected component migration must fail verification, got %v", err)
	}
}

func TestRuntimeRollsBackFailedMigration(t *testing.T) {
	ctx := context.Background()
	databasePath := filepath.Join(t.TempDir(), "radishmind.db")
	broken := []Migration{{
		Component:          "test_records",
		ID:                 "0001_test_records",
		StoreSchemaVersion: "test_records_store_v1",
		UpSQL:              "CREATE TABLE test_records (record_id TEXT PRIMARY KEY); THIS IS NOT SQL;",
	}}
	if _, err := Open(ctx, Options{DatabasePath: databasePath, Migrations: broken}); err == nil || !strings.Contains(err.Error(), "SQLite operation failed") {
		t.Fatalf("broken migration must fail without exposing driver detail, got %v", err)
	}

	recovered, err := Open(ctx, Options{DatabasePath: databasePath, Migrations: testMigrations()[:1]})
	if err != nil {
		t.Fatalf("open after failed migration rollback: %v", err)
	}
	t.Cleanup(func() { _ = recovered.Close() })
	assertMigrationMarkerCount(t, recovered.DB(), 1)
}

func TestRuntimeRejectsCorruptDatabaseAndInvalidPaths(t *testing.T) {
	ctx := context.Background()
	corruptPath := filepath.Join(t.TempDir(), "corrupt.db")
	if err := os.WriteFile(corruptPath, []byte("not a SQLite database"), 0o600); err != nil {
		t.Fatalf("write corrupt database: %v", err)
	}
	if _, err := Open(ctx, Options{DatabasePath: corruptPath}); err == nil {
		t.Fatal("corrupt SQLite database must fail closed")
	}

	parentFile := filepath.Join(t.TempDir(), "parent-file")
	if err := os.WriteFile(parentFile, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("write parent file: %v", err)
	}
	if _, err := Open(ctx, Options{DatabasePath: filepath.Join(parentFile, "radishmind.db")}); err == nil || err.Error() != "SQLite development database parent is not a directory" {
		t.Fatalf("non-directory parent must fail deterministically, got %v", err)
	}
	if runtime.GOOS != "windows" {
		symlinkTarget := filepath.Join(t.TempDir(), "target.db")
		if err := os.WriteFile(symlinkTarget, nil, 0o600); err != nil {
			t.Fatalf("write SQLite symlink target: %v", err)
		}
		symlinkPath := filepath.Join(t.TempDir(), "symlink.db")
		if err := os.Symlink(symlinkTarget, symlinkPath); err != nil {
			t.Fatalf("create SQLite database symlink: %v", err)
		}
		if _, err := Open(ctx, Options{DatabasePath: symlinkPath}); err == nil || err.Error() != "SQLite development database path is not a regular file" {
			t.Fatalf("database symlink must fail closed, got %v", err)
		}
	}
	if _, err := Open(ctx, Options{}); err == nil || err.Error() != "SQLite development database path is missing" {
		t.Fatalf("missing path must fail deterministically, got %v", err)
	}
	if _, err := Open(ctx, Options{DatabasePath: filepath.Join(t.TempDir(), "timeout.db"), BusyTimeout: -time.Second}); err == nil || err.Error() != "SQLite development busy timeout must be at least one millisecond" {
		t.Fatalf("invalid busy timeout must fail deterministically, got %v", err)
	}
}

func TestRuntimeValidatesMigrationDefinitions(t *testing.T) {
	testCases := []struct {
		name       string
		migrations []Migration
	}{
		{name: "invalid component", migrations: []Migration{{Component: "Bad Component", ID: "0001_valid", StoreSchemaVersion: "valid_store_v1", UpSQL: "SELECT 1;"}}},
		{name: "missing SQL", migrations: []Migration{{Component: "valid", ID: "0001_valid", StoreSchemaVersion: "valid_store_v1"}}},
		{name: "duplicate", migrations: []Migration{
			{Component: "valid", ID: "0001_valid", StoreSchemaVersion: "valid_store_v1", UpSQL: "SELECT 1;"},
			{Component: "valid", ID: "0001_valid", StoreSchemaVersion: "valid_store_v1", UpSQL: "SELECT 1;"},
		}},
		{name: "out of order", migrations: []Migration{
			{Component: "valid", ID: "0002_valid", StoreSchemaVersion: "valid_store_v2", UpSQL: "SELECT 2;"},
			{Component: "valid", ID: "0001_valid", StoreSchemaVersion: "valid_store_v1", UpSQL: "SELECT 1;"},
		}},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := Open(context.Background(), Options{DatabasePath: filepath.Join(t.TempDir(), "test.db"), Migrations: testCase.migrations}); err == nil {
				t.Fatal("invalid migration definition must fail before opening the database")
			}
		})
	}
}

func testMigrations() []Migration {
	return []Migration{
		{
			Component:          "test_records",
			ID:                 "0001_test_records",
			StoreSchemaVersion: "test_records_store_v1",
			UpSQL: `CREATE TABLE test_records (
                record_id TEXT PRIMARY KEY,
                payload TEXT NOT NULL CHECK (json_valid(payload))
            );`,
		},
		{
			Component:          "test_records",
			ID:                 "0002_test_records_index",
			StoreSchemaVersion: "test_records_store_v2",
			UpSQL:              "CREATE INDEX test_records_payload_idx ON test_records(payload);",
		},
	}
}

func assertRuntimeSettings(t *testing.T, database *sql.DB, expectedBusyTimeoutMilliseconds int64) {
	t.Helper()
	ctx := context.Background()
	connections := make([]*sql.Conn, 0, 2)
	for range 2 {
		connection, err := database.Conn(ctx)
		if err != nil {
			t.Fatalf("acquire SQLite connection: %v", err)
		}
		connections = append(connections, connection)
	}
	defer func() {
		for _, connection := range connections {
			_ = connection.Close()
		}
	}()
	for index, connection := range connections {
		var foreignKeys int
		var journalMode string
		var busyTimeout int64
		if err := connection.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
			t.Fatalf("read foreign_keys on connection %d: %v", index, err)
		}
		if err := connection.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&journalMode); err != nil {
			t.Fatalf("read journal_mode on connection %d: %v", index, err)
		}
		if err := connection.QueryRowContext(ctx, "PRAGMA busy_timeout").Scan(&busyTimeout); err != nil {
			t.Fatalf("read busy_timeout on connection %d: %v", index, err)
		}
		if foreignKeys != 1 || !strings.EqualFold(journalMode, "wal") || busyTimeout != expectedBusyTimeoutMilliseconds {
			t.Fatalf("unexpected settings on connection %d: foreign_keys=%d journal_mode=%s busy_timeout=%d", index, foreignKeys, journalMode, busyTimeout)
		}
	}
}

func assertMigrationMarkerCount(t *testing.T, database *sql.DB, expected int) {
	t.Helper()
	var count int
	if err := database.QueryRow(`SELECT count(*) FROM radishmind_schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count migration markers: %v", err)
	}
	if count != expected {
		t.Fatalf("unexpected migration marker count: got %d want %d", count, expected)
	}
}

func openRawDatabase(t *testing.T, databasePath string) *sql.DB {
	t.Helper()
	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatalf("open raw SQLite database: %v", err)
	}
	if err := database.Ping(); err != nil {
		database.Close()
		t.Fatalf("connect raw SQLite database: %v", err)
	}
	return database
}

func assertSQLiteFilePermissions(t *testing.T, databasePath string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		return
	}
	if info, err := os.Stat(filepath.Dir(databasePath)); err != nil || info.Mode().Perm() != 0o700 {
		t.Fatalf("unexpected SQLite directory permissions: info=%v err=%v", info, err)
	}
	for _, filePath := range []string{databasePath, databasePath + "-wal", databasePath + "-shm"} {
		info, err := os.Stat(filePath)
		if errors.Is(err, os.ErrNotExist) && filePath != databasePath {
			continue
		}
		if err != nil || info.Mode().Perm() != 0o600 {
			t.Fatalf("unexpected SQLite file permissions for %s: info=%v err=%v", filepath.Base(filePath), info, err)
		}
	}
}

func assertSQLiteCheckpointed(t *testing.T, databasePath string) {
	t.Helper()
	for _, sidecarPath := range []string{databasePath + "-wal", databasePath + "-shm"} {
		info, err := os.Stat(sidecarPath)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			t.Fatalf("inspect SQLite sidecar after close: %v", err)
		}
		if info.Size() != 0 {
			t.Fatalf("SQLite sidecar was not checkpointed: %s size=%d", filepath.Base(sidecarPath), info.Size())
		}
	}
}

func TestRuntimeCloseNilSafe(t *testing.T) {
	var runtime *Runtime
	if err := runtime.Close(); err != nil {
		t.Fatalf("nil runtime close should succeed: %v", err)
	}
}
