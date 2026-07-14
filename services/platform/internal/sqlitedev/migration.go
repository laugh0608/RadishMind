package sqlitedev

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const schemaMigrationTableSQL = `
CREATE TABLE IF NOT EXISTS radishmind_schema_migrations (
    component TEXT NOT NULL,
    migration_id TEXT NOT NULL,
    store_schema_version TEXT NOT NULL,
    migration_checksum TEXT NOT NULL,
    applied_at TEXT NOT NULL,
    PRIMARY KEY (component, migration_id)
) WITHOUT ROWID;`

var (
	migrationLabelPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,127}$`)
	migrationIDPattern    = regexp.MustCompile(`^[0-9]{4}_[a-z][a-z0-9_]{0,122}$`)
)

type Migration struct {
	Component          string
	ID                 string
	StoreSchemaVersion string
	UpSQL              string
}

type migrationState struct {
	Component          string
	ID                 string
	StoreSchemaVersion string
	Checksum           string
	AppliedAt          time.Time
}

func (runtime *Runtime) VerifyMigrations(ctx context.Context, migrations []Migration) error {
	if runtime == nil || runtime.database == nil {
		return errors.New("SQLite development runtime is unavailable")
	}
	if err := validateMigrations(migrations); err != nil {
		return err
	}
	expectedCountByComponent := make(map[string]int)
	for _, migration := range migrations {
		var storeSchemaVersion string
		var checksum string
		err := runtime.database.QueryRowContext(ctx, `SELECT store_schema_version, migration_checksum
            FROM radishmind_schema_migrations WHERE component=? AND migration_id=?`, migration.Component, migration.ID).
			Scan(&storeSchemaVersion, &checksum)
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("SQLite development component migration is not applied")
		}
		if err != nil {
			return errors.New("verify SQLite development component migration")
		}
		if storeSchemaVersion != migration.StoreSchemaVersion || checksum != migration.Checksum() {
			return errors.New("SQLite development component migration marker mismatch")
		}
		expectedCountByComponent[migration.Component]++
	}
	for component, expectedCount := range expectedCountByComponent {
		var actualCount int
		if err := runtime.database.QueryRowContext(ctx, `SELECT count(*) FROM radishmind_schema_migrations WHERE component=?`, component).Scan(&actualCount); err != nil {
			return errors.New("count SQLite development component migrations")
		}
		if actualCount != expectedCount {
			return errors.New("SQLite development component contains an unexpected migration marker")
		}
	}
	return nil
}

func (migration Migration) Checksum() string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(migration.UpSQL)))
}

func applyMigrations(ctx context.Context, database *sql.DB, migrations []Migration) error {
	connection, err := database.Conn(ctx)
	if err != nil {
		return errors.New("acquire SQLite development migration connection")
	}
	defer connection.Close()
	if _, err := connection.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return errors.New("begin SQLite development migration")
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = connection.ExecContext(context.Background(), "ROLLBACK")
		}
	}()

	if _, err := connection.ExecContext(ctx, schemaMigrationTableSQL); err != nil {
		return errors.New("create SQLite development migration marker")
	}
	existingStates, err := readMigrationStates(ctx, connection)
	if err != nil {
		return err
	}
	expectedMigrations := make(map[string]Migration, len(migrations))
	for _, migration := range migrations {
		expectedMigrations[migrationKey(migration.Component, migration.ID)] = migration
	}
	for _, state := range existingStates {
		expected, ok := expectedMigrations[migrationKey(state.Component, state.ID)]
		if !ok {
			return errors.New("SQLite development database contains an unexpected migration marker")
		}
		if state.StoreSchemaVersion != expected.StoreSchemaVersion || state.Checksum != expected.Checksum() {
			return errors.New("SQLite development migration marker mismatch")
		}
		delete(expectedMigrations, migrationKey(state.Component, state.ID))
	}

	for _, migration := range migrations {
		if _, pending := expectedMigrations[migrationKey(migration.Component, migration.ID)]; !pending {
			continue
		}
		if _, err := connection.ExecContext(ctx, migration.UpSQL); err != nil {
			return sanitizedSQLiteError("apply SQLite development component migration")
		}
		if _, err := connection.ExecContext(ctx, `INSERT INTO radishmind_schema_migrations
            (component, migration_id, store_schema_version, migration_checksum, applied_at)
            VALUES (?, ?, ?, ?, ?)`,
			migration.Component,
			migration.ID,
			migration.StoreSchemaVersion,
			migration.Checksum(),
			time.Now().UTC().Format(time.RFC3339Nano),
		); err != nil {
			return errors.New("write SQLite development migration marker")
		}
	}
	if _, err := connection.ExecContext(ctx, "COMMIT"); err != nil {
		return errors.New("commit SQLite development migration")
	}
	committed = true
	return nil
}

func readMigrationStates(ctx context.Context, connection *sql.Conn) ([]migrationState, error) {
	rows, err := connection.QueryContext(ctx, `SELECT component, migration_id, store_schema_version,
        migration_checksum, applied_at FROM radishmind_schema_migrations ORDER BY component, migration_id`)
	if err != nil {
		return nil, errors.New("read SQLite development migration markers")
	}
	defer rows.Close()
	states := make([]migrationState, 0)
	for rows.Next() {
		var state migrationState
		var appliedAt string
		if err := rows.Scan(&state.Component, &state.ID, &state.StoreSchemaVersion, &state.Checksum, &appliedAt); err != nil {
			return nil, errors.New("scan SQLite development migration marker")
		}
		parsedAppliedAt, err := time.Parse(time.RFC3339Nano, appliedAt)
		if err != nil {
			return nil, errors.New("parse SQLite development migration applied time")
		}
		state.AppliedAt = parsedAppliedAt
		states = append(states, state)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.New("complete SQLite development migration marker read")
	}
	return states, nil
}

func validateMigrations(migrations []Migration) error {
	seen := make(map[string]struct{}, len(migrations))
	lastIDByComponent := make(map[string]string)
	for _, migration := range migrations {
		if !migrationLabelPattern.MatchString(migration.Component) {
			return errors.New("SQLite development migration component is invalid")
		}
		if !migrationIDPattern.MatchString(migration.ID) {
			return errors.New("SQLite development migration id is invalid")
		}
		if !migrationLabelPattern.MatchString(migration.StoreSchemaVersion) {
			return errors.New("SQLite development store schema version is invalid")
		}
		if strings.TrimSpace(migration.UpSQL) == "" {
			return errors.New("SQLite development migration SQL is missing")
		}
		key := migrationKey(migration.Component, migration.ID)
		if _, exists := seen[key]; exists {
			return errors.New("SQLite development migration is duplicated")
		}
		if previousID, exists := lastIDByComponent[migration.Component]; exists && migration.ID <= previousID {
			return errors.New("SQLite development component migrations are not ordered")
		}
		lastIDByComponent[migration.Component] = migration.ID
		seen[key] = struct{}{}
	}
	return nil
}

func migrationKey(component string, migrationID string) string {
	return component + "\x00" + migrationID
}
