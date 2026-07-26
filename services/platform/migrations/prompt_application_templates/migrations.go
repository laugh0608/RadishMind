package promptapplicationtemplatemigrations

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
	Component                                    = "prompt_application_templates"
	MigrationID                                  = "0001_prompt_application_templates"
	StoreSchemaVersion                           = "prompt_application_template_store_v1"
	MigrationStateApplied                        = "applied"
	MigrationStateNotApplied                     = "not_applied"
	MigrationStateMismatch                       = "mismatch"
	promptApplicationTemplateMigrationLock int64 = 0x524d50544d503031
)

const markerSQL = `
CREATE TABLE IF NOT EXISTS prompt_application_template_schema_versions (
    component text PRIMARY KEY,
    migration_id text NOT NULL,
    store_schema_version text NOT NULL,
    migration_checksum text NOT NULL,
    applied_at timestamptz NOT NULL DEFAULT now()
);`

//go:embed 0001_prompt_application_templates.up.sql
var upSQL string

//go:embed 0001_prompt_application_templates.down.sql
var downSQL string

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
		return nil, errors.New("prompt application template PostgreSQL database URL is missing")
	}
	configuration, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, errors.New("parse prompt application template PostgreSQL configuration")
	}
	configuration.MaxConns = 8
	configuration.MaxConnLifetime = 30 * time.Minute
	configuration.MaxConnIdleTime = 5 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, configuration)
	if err != nil {
		return nil, errors.New("create prompt application template PostgreSQL pool")
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, errors.New("connect prompt application template PostgreSQL")
	}
	return pool, nil
}

func ExpectedChecksum() string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(upSQL)))
}

func Inspect(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("prompt application template PostgreSQL pool is missing")
	}
	return inspect(ctx, pool)
}

func Apply(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("prompt application template PostgreSQL pool is missing")
	}
	connection, err := pool.Acquire(ctx)
	if err != nil {
		return State{}, safeDatabaseError("acquire prompt application template migration connection", err)
	}
	defer connection.Release()
	if _, err := connection.Exec(ctx, "SELECT pg_advisory_lock($1)", promptApplicationTemplateMigrationLock); err != nil {
		return State{}, safeDatabaseError("lock prompt application template migration", err)
	}
	defer func() {
		unlock, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, _ = connection.Exec(unlock, "SELECT pg_advisory_unlock($1)", promptApplicationTemplateMigrationLock)
	}()
	transaction, err := connection.Begin(ctx)
	if err != nil {
		return State{}, safeDatabaseError("begin prompt application template migration", err)
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	if _, err := transaction.Exec(ctx, markerSQL); err != nil {
		return State{}, safeDatabaseError("create prompt application template migration marker", err)
	}
	state, err := inspect(ctx, transaction)
	if err != nil {
		return State{}, err
	}
	if state.MigrationState == MigrationStateApplied {
		if err := transaction.Commit(ctx); err != nil {
			return State{}, safeDatabaseError("commit prompt application template migration check", err)
		}
		return state, nil
	}
	if state.MigrationState == MigrationStateMismatch {
		return State{}, errors.New("prompt application template migration marker mismatch")
	}
	if _, err := transaction.Exec(ctx, upSQL); err != nil {
		return State{}, safeDatabaseError("apply prompt application template migration", err)
	}
	if _, err := transaction.Exec(ctx, `INSERT INTO prompt_application_template_schema_versions
        (component, migration_id, store_schema_version, migration_checksum) VALUES ($1,$2,$3,$4)`, Component, MigrationID, StoreSchemaVersion, ExpectedChecksum()); err != nil {
		return State{}, safeDatabaseError("write prompt application template migration marker", err)
	}
	state, err = inspect(ctx, transaction)
	if err != nil || state.MigrationState != MigrationStateApplied {
		return State{}, errors.New("prompt application template migration preflight failed")
	}
	if err := transaction.Commit(ctx); err != nil {
		return State{}, safeDatabaseError("commit prompt application template migration", err)
	}
	return state, nil
}

func RollbackForDevTest(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("prompt application template PostgreSQL pool is missing")
	}
	connection, err := pool.Acquire(ctx)
	if err != nil {
		return State{}, safeDatabaseError("acquire prompt application template rollback connection", err)
	}
	defer connection.Release()
	if _, err := connection.Exec(ctx, "SELECT pg_advisory_lock($1)", promptApplicationTemplateMigrationLock); err != nil {
		return State{}, safeDatabaseError("lock prompt application template rollback", err)
	}
	defer func() {
		unlock, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, _ = connection.Exec(unlock, "SELECT pg_advisory_unlock($1)", promptApplicationTemplateMigrationLock)
	}()
	transaction, err := connection.Begin(ctx)
	if err != nil {
		return State{}, safeDatabaseError("begin prompt application template rollback", err)
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	state, err := inspect(ctx, transaction)
	if err != nil {
		return State{}, err
	}
	if state.MigrationState == MigrationStateNotApplied {
		_ = transaction.Commit(ctx)
		return state, nil
	}
	if state.MigrationState != MigrationStateApplied {
		return State{}, errors.New("prompt application template rollback requires matching migration marker")
	}
	if _, err := transaction.Exec(ctx, downSQL); err != nil {
		return State{}, safeDatabaseError("rollback prompt application template migration", err)
	}
	if _, err := transaction.Exec(ctx, `DELETE FROM prompt_application_template_schema_versions WHERE component=$1`, Component); err != nil {
		return State{}, safeDatabaseError("clear prompt application template migration marker", err)
	}
	if err := transaction.Commit(ctx); err != nil {
		return State{}, safeDatabaseError("commit prompt application template rollback", err)
	}
	return State{MigrationState: MigrationStateNotApplied}, nil
}

func inspect(ctx context.Context, query rowQuerier) (State, error) {
	var markerExists bool
	if err := query.QueryRow(ctx, "SELECT to_regclass('public.prompt_application_template_schema_versions') IS NOT NULL").Scan(&markerExists); err != nil {
		return State{}, safeDatabaseError("inspect prompt application template migration marker", err)
	}
	if !markerExists {
		return State{MigrationState: MigrationStateNotApplied}, nil
	}
	state := State{}
	err := query.QueryRow(ctx, `SELECT migration_id,store_schema_version,migration_checksum,applied_at
        FROM prompt_application_template_schema_versions WHERE component=$1`, Component).Scan(&state.MigrationID, &state.StoreSchemaVersion, &state.MigrationChecksum, &state.AppliedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return State{MigrationState: MigrationStateNotApplied}, nil
	}
	if err != nil {
		return State{}, safeDatabaseError("read prompt application template migration marker", err)
	}
	var draftsExist, versionsExist bool
	if err := query.QueryRow(ctx, "SELECT to_regclass('public.prompt_application_template_drafts') IS NOT NULL, to_regclass('public.prompt_application_template_versions') IS NOT NULL").Scan(&draftsExist, &versionsExist); err != nil {
		return State{}, safeDatabaseError("inspect prompt application template tables", err)
	}
	var triggerCount int
	if err := query.QueryRow(ctx, `SELECT count(*) FROM pg_trigger trigger
		JOIN pg_class relation ON relation.oid=trigger.tgrelid
		JOIN pg_namespace namespace ON namespace.oid=relation.relnamespace
		WHERE NOT trigger.tgisinternal AND namespace.nspname='public' AND trigger.tgname IN (
			'prompt_application_template_drafts_controlled_update','prompt_application_template_drafts_no_delete',
			'prompt_application_template_versions_no_update','prompt_application_template_versions_no_delete'
		)`).Scan(&triggerCount); err != nil {
		return State{}, safeDatabaseError("inspect prompt application template triggers", err)
	}
	state.MigrationState = MigrationStateApplied
	if state.MigrationID != MigrationID || state.StoreSchemaVersion != StoreSchemaVersion || state.MigrationChecksum != ExpectedChecksum() || !draftsExist || !versionsExist || triggerCount != 4 {
		state.MigrationState = MigrationStateMismatch
	}
	return state, nil
}

func safeDatabaseError(operation string, err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code != "" {
		return fmt.Errorf("%s failed (SQLSTATE %s)", operation, pgError.Code)
	}
	return fmt.Errorf("%s failed", operation)
}
