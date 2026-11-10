package workbench

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"radishmind.local/services/platform/internal/sqlitedev"
	localmigrations "radishmind.local/services/platform/migrations/sqlite"
)

type MigrationRecord struct {
	Component string `json:"component"`
	ID        string `json:"id"`
	Version   string `json:"version"`
	Checksum  string `json:"checksum"`
}

// Inspect never applies migrations. It rejects unknown or altered markers even
// when upgrade is allowed; only a known subset may be upgraded by this binary.
func Inspect(ctx context.Context, path string, allowPending bool) ([]MigrationRecord, error) {
	db, err := openReadOnly(path)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	if err := verifyIntegrity(ctx, db); err != nil {
		return nil, err
	}
	expected := map[string]sqlitedev.Migration{}
	for _, migration := range localmigrations.Migrations() {
		expected[migration.Component+"/"+migration.ID] = migration
	}
	rows, err := db.QueryContext(ctx, `SELECT component, migration_id, store_schema_version, migration_checksum FROM radishmind_schema_migrations ORDER BY component, migration_id`)
	if err != nil {
		return nil, errors.New("workbench database migration registry is missing or unreadable")
	}
	defer rows.Close()
	records := []MigrationRecord{}
	for rows.Next() {
		var record MigrationRecord
		if err := rows.Scan(&record.Component, &record.ID, &record.Version, &record.Checksum); err != nil {
			return nil, errors.New("read workbench migration marker")
		}
		key := record.Component + "/" + record.ID
		migration, found := expected[key]
		if !found || record.Version != migration.StoreSchemaVersion || record.Checksum != migration.Checksum() {
			return nil, errors.New("workbench database has unknown or altered migrations; use the matching binary and preserve this database")
		}
		delete(expected, key)
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.New("complete workbench migration inspection")
	}
	if len(expected) > 0 && !allowPending {
		return nil, errors.New("workbench database requires explicit upgrade before serve")
	}
	return records, nil
}

func openReadOnly(path string) (*sql.DB, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, errors.New("workbench database is missing; run init explicitly")
	}
	if err != nil || !info.Mode().IsRegular() {
		return nil, errors.New("workbench database must be an accessible regular file")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	location := url.URL{Scheme: "file", Path: filepath.ToSlash(absolute)}
	query := location.Query()
	query.Set("mode", "ro")
	query.Add("_pragma", "busy_timeout(5000)")
	location.RawQuery = query.Encode()
	db, err := sql.Open("sqlite", location.String())
	if err != nil {
		return nil, errors.New("open workbench database for inspection")
	}
	db.SetMaxOpenConns(1)
	return db, nil
}

func verifyIntegrity(ctx context.Context, db *sql.DB) error {
	var result string
	if err := db.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&result); err != nil || result != "ok" {
		return errors.New("workbench database integrity check failed")
	}
	rows, err := db.QueryContext(ctx, "PRAGMA foreign_key_check")
	if err != nil {
		return errors.New("workbench database foreign key check failed")
	}
	defer rows.Close()
	if rows.Next() {
		return errors.New("workbench database contains foreign key violations")
	}
	return rows.Err()
}

// Initialize publishes a fully migrated, closed database without replacing an
// existing target. A failed temporary database remains available for inspection.
func Initialize(ctx context.Context, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
		return errors.New("init refuses an existing or inaccessible database")
	}
	lock, err := acquire(filepath.Dir(path), true)
	if err != nil {
		return err
	}
	defer lock.Close()
	if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
		return errors.New("init refuses an existing or inaccessible database")
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".initialize-*.db")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	if err := temporary.Close(); err != nil {
		return err
	}
	runtime, err := sqlitedev.Open(ctx, sqlitedev.Options{DatabasePath: temporaryPath, Migrations: localmigrations.Migrations()})
	if err != nil {
		return fmt.Errorf("initialize database (temporary file preserved): %w", err)
	}
	if err := runtime.Close(); err != nil {
		return err
	}
	if _, err := Inspect(ctx, temporaryPath, false); err != nil {
		return err
	}
	if err := os.Link(temporaryPath, path); err != nil {
		return fmt.Errorf("publish initialized database without overwrite: %w", err)
	}
	return os.Remove(temporaryPath)
}
