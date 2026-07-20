package apikeymigrations

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	Component                      = "api_key_records"
	MigrationID                    = "0001_api_key_records"
	StoreSchemaVersion             = "api_key_records_store_v1"
	MigrationStateApplied          = "applied"
	MigrationStateNotApplied       = "not_applied"
	MigrationStateMismatch         = "mismatch"
	apiKeyMigrationLock      int64 = 0x524d41504b455931
)

const schemaMarkerSQL = `
CREATE TABLE IF NOT EXISTS api_key_schema_versions (
    component text PRIMARY KEY,
    migration_id text NOT NULL,
    store_schema_version text NOT NULL,
    migration_checksum text NOT NULL,
    applied_at timestamptz NOT NULL DEFAULT now()
);`

//go:embed 0001_api_key_records.up.sql
var upSQL string

//go:embed 0001_api_key_records.down.sql
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
		return nil, errors.New("API key PostgreSQL database URL is missing")
	}
	configuration, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, errors.New("parse API key PostgreSQL configuration")
	}
	configuration.MaxConns = 8
	configuration.MaxConnLifetime = 30 * time.Minute
	configuration.MaxConnIdleTime = 5 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, configuration)
	if err != nil {
		return nil, errors.New("create API key PostgreSQL pool")
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, errors.New("connect API key PostgreSQL")
	}
	return pool, nil
}

func ExpectedChecksum() string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(upSQL)))
}

func Inspect(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("API key PostgreSQL pool is missing")
	}
	return inspect(ctx, pool)
}

func Apply(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("API key PostgreSQL pool is missing")
	}
	connection, err := pool.Acquire(ctx)
	if err != nil {
		return State{}, errors.New("acquire API key migration connection")
	}
	defer connection.Release()
	if _, err := connection.Exec(ctx, "SELECT pg_advisory_lock($1)", apiKeyMigrationLock); err != nil {
		return State{}, errors.New("lock API key migration")
	}
	defer func() {
		_, _ = connection.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", apiKeyMigrationLock)
	}()
	transaction, err := connection.Begin(ctx)
	if err != nil {
		return State{}, errors.New("begin API key migration")
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	if _, err := transaction.Exec(ctx, schemaMarkerSQL); err != nil {
		return State{}, errors.New("create API key migration marker")
	}
	state, err := inspect(ctx, transaction)
	if err != nil {
		return State{}, err
	}
	if state.MigrationState == MigrationStateApplied {
		if err := transaction.Commit(ctx); err != nil {
			return State{}, errors.New("commit API key migration check")
		}
		return state, nil
	}
	if state.MigrationState == MigrationStateMismatch {
		return State{}, errors.New("API key migration marker mismatch")
	}
	if _, err := transaction.Exec(ctx, upSQL); err != nil {
		return State{}, errors.New("apply API key migration")
	}
	if _, err := transaction.Exec(ctx, `INSERT INTO api_key_schema_versions
        (component, migration_id, store_schema_version, migration_checksum) VALUES ($1,$2,$3,$4)`,
		Component, MigrationID, StoreSchemaVersion, ExpectedChecksum()); err != nil {
		return State{}, errors.New("write API key migration marker")
	}
	if err := transaction.Commit(ctx); err != nil {
		return State{}, errors.New("commit API key migration")
	}
	return Inspect(ctx, pool)
}

func RollbackForDevTest(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("API key PostgreSQL pool is missing")
	}
	connection, err := pool.Acquire(ctx)
	if err != nil {
		return State{}, errors.New("acquire API key rollback connection")
	}
	defer connection.Release()
	if _, err := connection.Exec(ctx, "SELECT pg_advisory_lock($1)", apiKeyMigrationLock); err != nil {
		return State{}, errors.New("lock API key rollback")
	}
	defer func() {
		_, _ = connection.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", apiKeyMigrationLock)
	}()
	transaction, err := connection.Begin(ctx)
	if err != nil {
		return State{}, errors.New("begin API key rollback")
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	state, err := inspect(ctx, transaction)
	if err != nil {
		return State{}, err
	}
	if state.MigrationState == MigrationStateNotApplied {
		if err := transaction.Commit(ctx); err != nil {
			return State{}, errors.New("commit API key rollback check")
		}
		return state, nil
	}
	if state.MigrationState != MigrationStateApplied {
		return State{}, errors.New("API key rollback requires matching migration marker")
	}
	if _, err := transaction.Exec(ctx, downSQL); err != nil {
		return State{}, errors.New("rollback API key migration")
	}
	if err := transaction.Commit(ctx); err != nil {
		return State{}, errors.New("commit API key rollback")
	}
	return State{MigrationState: MigrationStateNotApplied}, nil
}

func inspect(ctx context.Context, query rowQuerier) (State, error) {
	var exists bool
	if err := query.QueryRow(ctx, "SELECT to_regclass('public.api_key_schema_versions') IS NOT NULL").Scan(&exists); err != nil {
		return State{}, errors.New("inspect API key migration marker")
	}
	if !exists {
		return State{MigrationState: MigrationStateNotApplied}, nil
	}
	state := State{}
	err := query.QueryRow(ctx, `SELECT migration_id, store_schema_version, migration_checksum, applied_at
        FROM api_key_schema_versions WHERE component=$1`, Component).Scan(&state.MigrationID, &state.StoreSchemaVersion, &state.MigrationChecksum, &state.AppliedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return State{MigrationState: MigrationStateNotApplied}, nil
	}
	if err != nil {
		return State{}, errors.New("read API key migration marker")
	}
	var tableExists bool
	if err := query.QueryRow(ctx, "SELECT to_regclass('public.api_key_records') IS NOT NULL").Scan(&tableExists); err != nil {
		return State{}, errors.New("inspect API key records table")
	}
	state.MigrationState = MigrationStateApplied
	if state.MigrationID != MigrationID || state.StoreSchemaVersion != StoreSchemaVersion ||
		state.MigrationChecksum != ExpectedChecksum() || !tableExists {
		state.MigrationState = MigrationStateMismatch
	}
	return state, nil
}
