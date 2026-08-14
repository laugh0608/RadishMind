package adminproviderroutesmigrations

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
	Component                       = "admin_provider_routes"
	MigrationID                     = "0002_admin_provider_route_v2"
	StoreSchemaVersion              = "admin_provider_routes_store_v2"
	initialMigrationID              = "0001_admin_provider_routes"
	initialStoreSchemaVersion       = "admin_provider_routes_store_v1"
	MigrationStateApplied           = "applied"
	MigrationStateNotApplied        = "not_applied"
	MigrationStateUpgradeRequired   = "upgrade_required"
	MigrationStateMismatch          = "mismatch"
	adminProviderRouteMigrationLock = int64(0x524d415052543031)
)

const schemaMarkerSQL = `
CREATE TABLE IF NOT EXISTS admin_provider_route_schema_versions (
    component text PRIMARY KEY,
    migration_id text NOT NULL,
    store_schema_version text NOT NULL,
    migration_checksum text NOT NULL,
    applied_at timestamptz NOT NULL DEFAULT now()
);`

//go:embed 0001_admin_provider_routes.up.sql
var initialUpSQL string

//go:embed 0001_admin_provider_routes.down.sql
var initialDownSQL string

//go:embed 0002_admin_provider_route_v2.up.sql
var routeV2UpSQL string

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
		return nil, errors.New("admin provider route PostgreSQL database URL is missing")
	}
	configuration, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, errors.New("parse admin provider route PostgreSQL configuration")
	}
	configuration.MaxConns = 8
	configuration.MaxConnLifetime = 30 * time.Minute
	configuration.MaxConnIdleTime = 5 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, configuration)
	if err != nil {
		return nil, errors.New("create admin provider route PostgreSQL pool")
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, errors.New("connect admin provider route PostgreSQL")
	}
	return pool, nil
}

func ExpectedChecksum() string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(initialUpSQL+"\n"+routeV2UpSQL)))
}

func initialExpectedChecksum() string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(initialUpSQL)))
}

func Inspect(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("admin provider route PostgreSQL pool is missing")
	}
	return inspect(ctx, pool)
}

func Apply(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("admin provider route PostgreSQL pool is missing")
	}
	connection, err := pool.Acquire(ctx)
	if err != nil {
		return State{}, errors.New("acquire admin provider route migration connection")
	}
	defer connection.Release()
	if _, err := connection.Exec(ctx, "SELECT pg_advisory_lock($1)", adminProviderRouteMigrationLock); err != nil {
		return State{}, errors.New("lock admin provider route migration")
	}
	defer func() {
		_, _ = connection.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", adminProviderRouteMigrationLock)
	}()
	transaction, err := connection.Begin(ctx)
	if err != nil {
		return State{}, errors.New("begin admin provider route migration")
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	if _, err := transaction.Exec(ctx, schemaMarkerSQL); err != nil {
		return State{}, errors.New("create admin provider route migration marker")
	}
	state, err := inspect(ctx, transaction)
	if err != nil {
		return State{}, err
	}
	if state.MigrationState == MigrationStateApplied {
		if err := transaction.Commit(ctx); err != nil {
			return State{}, errors.New("commit admin provider route migration check")
		}
		return state, nil
	}
	if state.MigrationState == MigrationStateMismatch {
		return State{}, errors.New("admin provider route migration marker mismatch")
	}
	if state.MigrationState == MigrationStateNotApplied {
		if _, err := transaction.Exec(ctx, initialUpSQL); err != nil {
			return State{}, errors.New("apply initial admin provider route migration")
		}
		if _, err := transaction.Exec(ctx, routeV2UpSQL); err != nil {
			return State{}, errors.New("apply admin provider route v2 migration")
		}
		if _, err := transaction.Exec(ctx, `INSERT INTO admin_provider_route_schema_versions
            (component, migration_id, store_schema_version, migration_checksum) VALUES ($1,$2,$3,$4)`,
			Component, MigrationID, StoreSchemaVersion, ExpectedChecksum()); err != nil {
			return State{}, errors.New("write admin provider route migration marker")
		}
	} else {
		if _, err := transaction.Exec(ctx, routeV2UpSQL); err != nil {
			return State{}, errors.New("upgrade admin provider route v2 migration")
		}
		if _, err := transaction.Exec(ctx, `UPDATE admin_provider_route_schema_versions
            SET migration_id=$1, store_schema_version=$2, migration_checksum=$3, applied_at=now()
            WHERE component=$4`, MigrationID, StoreSchemaVersion, ExpectedChecksum(), Component); err != nil {
			return State{}, errors.New("update admin provider route migration marker")
		}
	}
	if err := transaction.Commit(ctx); err != nil {
		return State{}, errors.New("commit admin provider route migration")
	}
	return Inspect(ctx, pool)
}

func RollbackForDevTest(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("admin provider route PostgreSQL pool is missing")
	}
	connection, err := pool.Acquire(ctx)
	if err != nil {
		return State{}, errors.New("acquire admin provider route rollback connection")
	}
	defer connection.Release()
	if _, err := connection.Exec(ctx, "SELECT pg_advisory_lock($1)", adminProviderRouteMigrationLock); err != nil {
		return State{}, errors.New("lock admin provider route rollback")
	}
	defer func() {
		_, _ = connection.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", adminProviderRouteMigrationLock)
	}()
	transaction, err := connection.Begin(ctx)
	if err != nil {
		return State{}, errors.New("begin admin provider route rollback")
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	state, err := inspect(ctx, transaction)
	if err != nil {
		return State{}, err
	}
	if state.MigrationState == MigrationStateNotApplied {
		if err := transaction.Commit(ctx); err != nil {
			return State{}, errors.New("commit admin provider route rollback check")
		}
		return state, nil
	}
	if state.MigrationState != MigrationStateApplied && state.MigrationState != MigrationStateUpgradeRequired {
		return State{}, errors.New("admin provider route rollback requires matching migration marker")
	}
	if _, err := transaction.Exec(ctx, initialDownSQL); err != nil {
		return State{}, errors.New("rollback admin provider route migration")
	}
	if err := transaction.Commit(ctx); err != nil {
		return State{}, errors.New("commit admin provider route rollback")
	}
	return State{MigrationState: MigrationStateNotApplied}, nil
}

func inspect(ctx context.Context, query rowQuerier) (State, error) {
	var markerExists bool
	if err := query.QueryRow(ctx, "SELECT to_regclass('public.admin_provider_route_schema_versions') IS NOT NULL").Scan(&markerExists); err != nil {
		return State{}, errors.New("inspect admin provider route migration marker")
	}
	if !markerExists {
		return State{MigrationState: MigrationStateNotApplied}, nil
	}
	state := State{}
	err := query.QueryRow(ctx, `SELECT migration_id, store_schema_version, migration_checksum, applied_at
        FROM admin_provider_route_schema_versions WHERE component=$1`, Component).
		Scan(&state.MigrationID, &state.StoreSchemaVersion, &state.MigrationChecksum, &state.AppliedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return State{MigrationState: MigrationStateNotApplied}, nil
	}
	if err != nil {
		return State{}, errors.New("read admin provider route migration marker")
	}
	expectedTables := []string{
		"admin_provider_route_drafts",
		"admin_provider_route_candidates",
		"admin_provider_route_reviews",
		"admin_provider_route_active_snapshots",
		"admin_provider_route_activation_records",
	}
	tablesPresent := true
	for _, table := range expectedTables {
		var exists bool
		if err := query.QueryRow(ctx, "SELECT to_regclass($1) IS NOT NULL", "public."+table).Scan(&exists); err != nil {
			return State{}, errors.New("inspect admin provider route table")
		}
		tablesPresent = tablesPresent && exists
	}
	state.MigrationState = MigrationStateApplied
	if state.MigrationID == initialMigrationID && state.StoreSchemaVersion == initialStoreSchemaVersion &&
		state.MigrationChecksum == initialExpectedChecksum() && tablesPresent {
		state.MigrationState = MigrationStateUpgradeRequired
	} else if state.MigrationID != MigrationID || state.StoreSchemaVersion != StoreSchemaVersion ||
		state.MigrationChecksum != ExpectedChecksum() || !tablesPresent {
		state.MigrationState = MigrationStateMismatch
	} else {
		var v2ConstraintsPresent bool
		if err := query.QueryRow(ctx, `SELECT count(*) = 3
            FROM pg_constraint
            WHERE conname IN (
                'admin_provider_route_draft_payload_v2_check',
                'admin_provider_route_candidate_payload_v2_check',
                'admin_provider_route_snapshot_payload_v2_check'
            )`).Scan(&v2ConstraintsPresent); err != nil {
			return State{}, errors.New("inspect admin provider route v2 constraints")
		}
		if !v2ConstraintsPresent {
			state.MigrationState = MigrationStateMismatch
		}
	}
	return state, nil
}
