package controlplanereadmigrations

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
	Component                           = "control_plane_admin_read"
	MigrationID                         = "0001_admin_tenant_audit_read"
	StoreSchemaVersion                  = "control_plane_admin_read_store_v1"
	TenantProjectionSchemaVersion       = "control_plane_tenant_summary_projection.v1"
	AuditProjectionSchemaVersion        = "control_plane_audit_summary_projection.v1"
	MigrationStateApplied               = "applied"
	MigrationStateNotApplied            = "not_applied"
	MigrationStateMismatch              = "mismatch"
	controlPlaneReadMigrationLock int64 = 0x524d435052443031
)

//go:embed 0001_admin_tenant_audit_read.up.sql
var upSQL string

//go:embed 0001_admin_tenant_audit_read.down.sql
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
		return nil, errors.New("control plane read PostgreSQL database URL is missing")
	}
	configuration, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, errors.New("parse control plane read PostgreSQL configuration")
	}
	configuration.MaxConns = 4
	configuration.MaxConnLifetime = 30 * time.Minute
	configuration.MaxConnIdleTime = 5 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, configuration)
	if err != nil {
		return nil, errors.New("create control plane read PostgreSQL pool")
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, errors.New("connect control plane read PostgreSQL")
	}
	return pool, nil
}

func ExpectedChecksum() string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(upSQL)))
}

func Inspect(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("control plane read PostgreSQL pool is missing")
	}
	return inspect(ctx, pool)
}

func Apply(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("control plane read PostgreSQL pool is missing")
	}
	connection, err := pool.Acquire(ctx)
	if err != nil {
		return State{}, errors.New("acquire control plane read migration connection")
	}
	defer connection.Release()
	if _, err := connection.Exec(ctx, "SELECT pg_advisory_lock($1)", controlPlaneReadMigrationLock); err != nil {
		return State{}, errors.New("lock control plane read migration")
	}
	defer func() {
		_, _ = connection.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", controlPlaneReadMigrationLock)
	}()
	transaction, err := connection.Begin(ctx)
	if err != nil {
		return State{}, errors.New("begin control plane read migration")
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	state, err := inspect(ctx, transaction)
	if err != nil {
		return State{}, err
	}
	if state.MigrationState == MigrationStateApplied {
		if err := transaction.Commit(ctx); err != nil {
			return State{}, errors.New("commit control plane read migration check")
		}
		return state, nil
	}
	if state.MigrationState == MigrationStateMismatch {
		return State{}, errors.New("control plane read migration marker mismatch")
	}
	if _, err := transaction.Exec(ctx, upSQL); err != nil {
		return State{}, errors.New("apply control plane read migration")
	}
	if _, err := transaction.Exec(ctx, `INSERT INTO control_plane_read_schema_versions
        (component, migration_id, store_schema_version, migration_checksum) VALUES ($1, $2, $3, $4)`,
		Component, MigrationID, StoreSchemaVersion, ExpectedChecksum()); err != nil {
		return State{}, errors.New("write control plane read migration marker")
	}
	if err := transaction.Commit(ctx); err != nil {
		return State{}, errors.New("commit control plane read migration")
	}
	return Inspect(ctx, pool)
}

func RollbackForDevTest(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("control plane read PostgreSQL pool is missing")
	}
	connection, err := pool.Acquire(ctx)
	if err != nil {
		return State{}, errors.New("acquire control plane read rollback connection")
	}
	defer connection.Release()
	if _, err := connection.Exec(ctx, "SELECT pg_advisory_lock($1)", controlPlaneReadMigrationLock); err != nil {
		return State{}, errors.New("lock control plane read rollback")
	}
	defer func() {
		_, _ = connection.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", controlPlaneReadMigrationLock)
	}()
	transaction, err := connection.Begin(ctx)
	if err != nil {
		return State{}, errors.New("begin control plane read rollback")
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	state, err := inspect(ctx, transaction)
	if err != nil {
		return State{}, err
	}
	if state.MigrationState == MigrationStateNotApplied {
		if err := transaction.Commit(ctx); err != nil {
			return State{}, errors.New("commit control plane read rollback check")
		}
		return state, nil
	}
	if state.MigrationState != MigrationStateApplied {
		return State{}, errors.New("control plane read rollback requires matching migration marker")
	}
	if _, err := transaction.Exec(ctx, downSQL); err != nil {
		return State{}, errors.New("rollback control plane read migration")
	}
	if err := transaction.Commit(ctx); err != nil {
		return State{}, errors.New("commit control plane read rollback")
	}
	return State{MigrationState: MigrationStateNotApplied}, nil
}

func PreflightRuntime(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	state, err := Inspect(ctx, pool)
	if err != nil {
		return State{}, err
	}
	if state.MigrationState != MigrationStateApplied {
		return State{}, errors.New("control plane read schema migration is not ready")
	}
	for _, query := range []string{
		"SELECT tenant_ref FROM control_plane_tenant_summary_projections LIMIT 0",
		"SELECT tenant_ref FROM control_plane_audit_summary_projections LIMIT 0",
	} {
		rows, queryErr := pool.Query(ctx, query)
		if queryErr != nil {
			return State{}, errors.New("control plane read runtime SELECT preflight failed")
		}
		rows.Close()
	}
	return state, nil
}

func inspect(ctx context.Context, query rowQuerier) (State, error) {
	var exists bool
	if err := query.QueryRow(ctx, "SELECT to_regclass('public.control_plane_read_schema_versions') IS NOT NULL").Scan(&exists); err != nil {
		return State{}, errors.New("inspect control plane read migration marker")
	}
	if !exists {
		return State{MigrationState: MigrationStateNotApplied}, nil
	}
	state := State{}
	err := query.QueryRow(ctx, `SELECT migration_id, store_schema_version, migration_checksum, applied_at
        FROM control_plane_read_schema_versions WHERE component = $1`, Component).Scan(
		&state.MigrationID, &state.StoreSchemaVersion, &state.MigrationChecksum, &state.AppliedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return State{MigrationState: MigrationStateNotApplied}, nil
	}
	if err != nil {
		return State{}, errors.New("read control plane read migration marker")
	}
	state.MigrationState = MigrationStateApplied
	if state.MigrationID != MigrationID || state.StoreSchemaVersion != StoreSchemaVersion || state.MigrationChecksum != ExpectedChecksum() {
		state.MigrationState = MigrationStateMismatch
	}
	return state, nil
}
