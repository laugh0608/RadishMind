package workflowdraftmigrations

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	Component          = "workflow_saved_drafts"
	MigrationID        = "0001_saved_workflow_drafts"
	StoreSchemaVersion = "saved_workflow_drafts_store_v1"

	MigrationStateApplied    = "applied"
	MigrationStateNotApplied = "not_applied"
	MigrationStateMismatch   = "mismatch"

	migrationAdvisoryLockKey int64 = 0x524d445241465431
)

const schemaMarkerBootstrapSQL = `
CREATE TABLE IF NOT EXISTS workflow_saved_draft_schema_versions (
    component text PRIMARY KEY,
    migration_id text NOT NULL,
    store_schema_version text NOT NULL,
    migration_checksum text NOT NULL,
    applied_at timestamptz NOT NULL DEFAULT now()
);`

//go:embed 0001_saved_workflow_drafts.up.sql
var upMigrationSQL string

//go:embed 0001_saved_workflow_drafts.down.sql
var downMigrationSQL string

type State struct {
	MigrationState     string
	MigrationID        string
	StoreSchemaVersion string
	MigrationChecksum  string
	AppliedAt          time.Time
}

type rowQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func OpenPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	if strings.TrimSpace(databaseURL) == "" {
		return nil, errors.New("saved workflow draft PostgreSQL database URL is missing")
	}
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, safeDatabaseError("parse saved workflow draft PostgreSQL config", err)
	}
	poolConfig.MaxConns = 8
	poolConfig.MinConns = 0
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute
	poolConfig.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, safeDatabaseError("create saved workflow draft PostgreSQL pool", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, safeDatabaseError("connect saved workflow draft PostgreSQL", err)
	}
	return pool, nil
}

func ExpectedChecksum() string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(upMigrationSQL)))
}

func Inspect(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("saved workflow draft PostgreSQL pool is missing")
	}
	return inspectWithQuery(ctx, pool)
}

func Apply(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("saved workflow draft PostgreSQL pool is missing")
	}
	connection, err := pool.Acquire(ctx)
	if err != nil {
		return State{}, safeDatabaseError("acquire saved workflow draft migration connection", err)
	}
	defer connection.Release()

	if _, err := connection.Exec(ctx, "SELECT pg_advisory_lock($1)", migrationAdvisoryLockKey); err != nil {
		return State{}, safeDatabaseError("lock saved workflow draft migration", err)
	}
	defer func() {
		unlockContext, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, _ = connection.Exec(unlockContext, "SELECT pg_advisory_unlock($1)", migrationAdvisoryLockKey)
	}()

	transaction, err := connection.Begin(ctx)
	if err != nil {
		return State{}, safeDatabaseError("begin saved workflow draft migration", err)
	}
	defer func() {
		_ = transaction.Rollback(context.Background())
	}()

	if _, err := transaction.Exec(ctx, schemaMarkerBootstrapSQL); err != nil {
		return State{}, safeDatabaseError("create saved workflow draft schema marker", err)
	}
	state, err := inspectWithQuery(ctx, transaction)
	if err != nil {
		return State{}, err
	}
	switch state.MigrationState {
	case MigrationStateApplied:
		if err := transaction.Commit(ctx); err != nil {
			return State{}, safeDatabaseError("commit saved workflow draft migration check", err)
		}
		return state, nil
	case MigrationStateMismatch:
		return State{}, errors.New("saved workflow draft migration marker does not match the embedded migration")
	}

	if _, err := transaction.Exec(ctx, upMigrationSQL); err != nil {
		return State{}, safeDatabaseError("apply saved workflow draft migration", err)
	}
	if _, err := transaction.Exec(
		ctx,
		`INSERT INTO workflow_saved_draft_schema_versions
            (component, migration_id, store_schema_version, migration_checksum)
         VALUES ($1, $2, $3, $4)`,
		Component,
		MigrationID,
		StoreSchemaVersion,
		ExpectedChecksum(),
	); err != nil {
		return State{}, safeDatabaseError("write saved workflow draft migration marker", err)
	}

	state, err = inspectWithQuery(ctx, transaction)
	if err != nil {
		return State{}, err
	}
	if state.MigrationState != MigrationStateApplied {
		return State{}, errors.New("saved workflow draft migration preflight did not become applied")
	}
	if err := transaction.Commit(ctx); err != nil {
		return State{}, safeDatabaseError("commit saved workflow draft migration", err)
	}
	return state, nil
}

// RollbackForDevTest exercises the reviewed down migration in disposable development and test databases.
// It is intentionally not exposed by the operator CLI.
func RollbackForDevTest(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("saved workflow draft PostgreSQL pool is missing")
	}
	connection, err := pool.Acquire(ctx)
	if err != nil {
		return State{}, safeDatabaseError("acquire saved workflow draft rollback connection", err)
	}
	defer connection.Release()

	if _, err := connection.Exec(ctx, "SELECT pg_advisory_lock($1)", migrationAdvisoryLockKey); err != nil {
		return State{}, safeDatabaseError("lock saved workflow draft rollback", err)
	}
	defer func() {
		unlockContext, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, _ = connection.Exec(unlockContext, "SELECT pg_advisory_unlock($1)", migrationAdvisoryLockKey)
	}()

	transaction, err := connection.Begin(ctx)
	if err != nil {
		return State{}, safeDatabaseError("begin saved workflow draft rollback", err)
	}
	defer func() {
		_ = transaction.Rollback(context.Background())
	}()

	state, err := inspectWithQuery(ctx, transaction)
	if err != nil {
		return State{}, err
	}
	if state.MigrationState == MigrationStateNotApplied {
		if err := transaction.Commit(ctx); err != nil {
			return State{}, safeDatabaseError("commit saved workflow draft rollback check", err)
		}
		return state, nil
	}
	if state.MigrationState != MigrationStateApplied {
		return State{}, errors.New("saved workflow draft rollback requires the matching applied migration")
	}
	if _, err := transaction.Exec(ctx, downMigrationSQL); err != nil {
		return State{}, safeDatabaseError("rollback saved workflow draft migration", err)
	}
	if err := transaction.Commit(ctx); err != nil {
		return State{}, safeDatabaseError("commit saved workflow draft rollback", err)
	}
	return State{MigrationState: MigrationStateNotApplied}, nil
}

func inspectWithQuery(ctx context.Context, query rowQuerier) (State, error) {
	var markerTableExists bool
	if err := query.QueryRow(
		ctx,
		"SELECT to_regclass('public.workflow_saved_draft_schema_versions') IS NOT NULL",
	).Scan(&markerTableExists); err != nil {
		return State{}, safeDatabaseError("inspect saved workflow draft schema marker table", err)
	}
	if !markerTableExists {
		return State{MigrationState: MigrationStateNotApplied}, nil
	}

	state := State{}
	err := query.QueryRow(
		ctx,
		`SELECT migration_id, store_schema_version, migration_checksum, applied_at
           FROM workflow_saved_draft_schema_versions
          WHERE component = $1`,
		Component,
	).Scan(
		&state.MigrationID,
		&state.StoreSchemaVersion,
		&state.MigrationChecksum,
		&state.AppliedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return State{MigrationState: MigrationStateNotApplied}, nil
	}
	if err != nil {
		return State{}, safeDatabaseError("read saved workflow draft schema marker", err)
	}

	var draftTableExists bool
	if err := query.QueryRow(
		ctx,
		"SELECT to_regclass('public.saved_workflow_drafts') IS NOT NULL",
	).Scan(&draftTableExists); err != nil {
		return State{}, safeDatabaseError("inspect saved workflow draft table", err)
	}
	if state.MigrationID != MigrationID ||
		state.StoreSchemaVersion != StoreSchemaVersion ||
		state.MigrationChecksum != ExpectedChecksum() ||
		!draftTableExists {
		state.MigrationState = MigrationStateMismatch
		return state, nil
	}
	state.MigrationState = MigrationStateApplied
	return state, nil
}

func safeDatabaseError(operation string, err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && strings.TrimSpace(pgError.Code) != "" {
		return fmt.Errorf("%s failed (SQLSTATE %s)", operation, pgError.Code)
	}
	return fmt.Errorf("%s failed", operation)
}
