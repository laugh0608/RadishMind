package gatewayrequestmigrations

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
	Component                      = "gateway_requests"
	MigrationID                    = "0001_gateway_requests"
	StoreSchemaVersion             = "gateway_request_store_v1"
	MigrationStateApplied          = "applied"
	MigrationStateNotApplied       = "not_applied"
	MigrationStateMismatch         = "mismatch"
	migrationAdvisoryLockKey int64 = 0x524d475752513031
)

const markerSQL = `CREATE TABLE IF NOT EXISTS gateway_request_schema_versions (
component text PRIMARY KEY, migration_id text NOT NULL, store_schema_version text NOT NULL,
migration_checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now());`

//go:embed 0001_gateway_requests.up.sql
var upSQL string

//go:embed 0001_gateway_requests.down.sql
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
		return nil, errors.New("Gateway request PostgreSQL database URL is missing")
	}
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, safeDatabaseError("parse Gateway request PostgreSQL config", err)
	}
	config.MaxConns, config.MinConns = 8, 0
	config.MaxConnLifetime, config.MaxConnIdleTime, config.HealthCheckPeriod = 30*time.Minute, 5*time.Minute, 30*time.Second
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, safeDatabaseError("create Gateway request PostgreSQL pool", err)
	}
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, safeDatabaseError("connect Gateway request PostgreSQL", err)
	}
	return pool, nil
}

func ExpectedChecksum() string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(upSQL)))
}

func Inspect(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("Gateway request PostgreSQL pool is missing")
	}
	return inspect(ctx, pool)
}

func Apply(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("Gateway request PostgreSQL pool is missing")
	}
	connection, err := pool.Acquire(ctx)
	if err != nil {
		return State{}, safeDatabaseError("acquire Gateway request migration connection", err)
	}
	defer connection.Release()
	if _, err = connection.Exec(ctx, "SELECT pg_advisory_lock($1)", migrationAdvisoryLockKey); err != nil {
		return State{}, safeDatabaseError("lock Gateway request migration", err)
	}
	defer func() {
		unlock, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, _ = connection.Exec(unlock, "SELECT pg_advisory_unlock($1)", migrationAdvisoryLockKey)
	}()
	tx, err := connection.Begin(ctx)
	if err != nil {
		return State{}, safeDatabaseError("begin Gateway request migration", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err = tx.Exec(ctx, markerSQL); err != nil {
		return State{}, safeDatabaseError("create Gateway request schema marker", err)
	}
	state, err := inspect(ctx, tx)
	if err != nil {
		return State{}, err
	}
	if state.MigrationState == MigrationStateApplied {
		if err = tx.Commit(ctx); err != nil {
			return State{}, safeDatabaseError("commit Gateway request migration check", err)
		}
		return state, nil
	}
	if state.MigrationState == MigrationStateMismatch {
		return State{}, errors.New("Gateway request migration marker mismatch")
	}
	if _, err = tx.Exec(ctx, upSQL); err != nil {
		return State{}, safeDatabaseError("apply Gateway request migration", err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO gateway_request_schema_versions(component,migration_id,store_schema_version,migration_checksum) VALUES($1,$2,$3,$4)`, Component, MigrationID, StoreSchemaVersion, ExpectedChecksum()); err != nil {
		return State{}, safeDatabaseError("write Gateway request migration marker", err)
	}
	state, err = inspect(ctx, tx)
	if err != nil || state.MigrationState != MigrationStateApplied {
		return State{}, errors.New("Gateway request migration preflight failed")
	}
	if err = tx.Commit(ctx); err != nil {
		return State{}, safeDatabaseError("commit Gateway request migration", err)
	}
	return state, nil
}

func RollbackForDevTest(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("Gateway request PostgreSQL pool is missing")
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return State{}, safeDatabaseError("begin Gateway request rollback", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	state, err := inspect(ctx, tx)
	if err != nil {
		return State{}, err
	}
	if state.MigrationState == MigrationStateNotApplied {
		_ = tx.Commit(ctx)
		return state, nil
	}
	if state.MigrationState != MigrationStateApplied {
		return State{}, errors.New("Gateway request rollback requires matching migration")
	}
	if _, err = tx.Exec(ctx, downSQL); err != nil {
		return State{}, safeDatabaseError("rollback Gateway request migration", err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM gateway_request_schema_versions WHERE component=$1`, Component); err != nil {
		return State{}, safeDatabaseError("clear Gateway request migration marker", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return State{}, safeDatabaseError("commit Gateway request rollback", err)
	}
	return State{MigrationState: MigrationStateNotApplied}, nil
}

func inspect(ctx context.Context, query rowQuerier) (State, error) {
	var markerExists bool
	if err := query.QueryRow(ctx, "SELECT to_regclass('public.gateway_request_schema_versions') IS NOT NULL").Scan(&markerExists); err != nil {
		return State{}, safeDatabaseError("inspect Gateway request marker", err)
	}
	if !markerExists {
		return State{MigrationState: MigrationStateNotApplied}, nil
	}
	state := State{}
	err := query.QueryRow(ctx, `SELECT migration_id,store_schema_version,migration_checksum,applied_at FROM gateway_request_schema_versions WHERE component=$1`, Component).Scan(&state.MigrationID, &state.StoreSchemaVersion, &state.MigrationChecksum, &state.AppliedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return State{MigrationState: MigrationStateNotApplied}, nil
	}
	if err != nil {
		return State{}, safeDatabaseError("read Gateway request marker", err)
	}
	var tableExists bool
	if err = query.QueryRow(ctx, "SELECT to_regclass('public.gateway_request_records') IS NOT NULL").Scan(&tableExists); err != nil {
		return State{}, safeDatabaseError("inspect Gateway request table", err)
	}
	if state.MigrationID != MigrationID || state.StoreSchemaVersion != StoreSchemaVersion || state.MigrationChecksum != ExpectedChecksum() || !tableExists {
		state.MigrationState = MigrationStateMismatch
	} else {
		state.MigrationState = MigrationStateApplied
	}
	return state, nil
}

func safeDatabaseError(operation string, err error) error {
	var databaseError *pgconn.PgError
	if errors.As(err, &databaseError) {
		return fmt.Errorf("%s failed (SQLSTATE %s)", operation, databaseError.Code)
	}
	return fmt.Errorf("%s failed", operation)
}
