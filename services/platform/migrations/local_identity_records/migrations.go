package localidentitymigrations

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
	Component                           = "local_identity_records"
	MigrationID                         = "0003_local_identity_administration"
	StoreSchemaVersion                  = "local_identity_records_store_v3"
	legacyMigrationID                   = "0001_local_identity_records"
	legacyStoreSchemaVersion            = "local_identity_records_store_v1"
	oidcMigrationID                     = "0002_local_identity_oidc_authorization_transactions"
	oidcStoreSchemaVersion              = "local_identity_records_store_v2"
	MigrationStateApplied               = "applied"
	MigrationStateNotApplied            = "not_applied"
	MigrationStateUpgradeRequired       = "upgrade_required"
	MigrationStateMismatch              = "mismatch"
	localIdentityMigrationLock    int64 = 0x524d4944454e5431
)

const schemaMarkerSQL = `
CREATE TABLE IF NOT EXISTS local_identity_schema_versions (
    component text PRIMARY KEY,
    migration_id text NOT NULL,
    store_schema_version text NOT NULL,
    migration_checksum text NOT NULL,
    applied_at timestamptz NOT NULL DEFAULT now()
);`

//go:embed 0001_local_identity_records.up.sql
var legacyUpSQL string

//go:embed 0001_local_identity_records.down.sql
var legacyDownSQL string

//go:embed 0002_local_identity_oidc_authorization_transactions.up.sql
var oidcAuthorizationUpSQL string

//go:embed 0002_local_identity_oidc_authorization_transactions.down.sql
var oidcAuthorizationDownSQL string

//go:embed 0003_local_identity_administration.up.sql
var administrationUpSQL string

//go:embed 0003_local_identity_administration.down.sql
var administrationDownSQL string

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
		return nil, errors.New("local identity PostgreSQL database URL is missing")
	}
	configuration, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, errors.New("parse local identity PostgreSQL configuration")
	}
	configuration.MaxConns = 8
	configuration.MaxConnLifetime = 30 * time.Minute
	configuration.MaxConnIdleTime = 5 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, configuration)
	if err != nil {
		return nil, errors.New("create local identity PostgreSQL pool")
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, errors.New("connect local identity PostgreSQL")
	}
	return pool, nil
}

func ExpectedChecksum() string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(legacyUpSQL+"\n"+oidcAuthorizationUpSQL+"\n"+administrationUpSQL)))
}

func legacyExpectedChecksum() string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(legacyUpSQL)))
}

func oidcExpectedChecksum() string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(legacyUpSQL+"\n"+oidcAuthorizationUpSQL)))
}

func Inspect(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("local identity PostgreSQL pool is missing")
	}
	return inspect(ctx, pool)
}

func Apply(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("local identity PostgreSQL pool is missing")
	}
	connection, err := pool.Acquire(ctx)
	if err != nil {
		return State{}, errors.New("acquire local identity migration connection")
	}
	defer connection.Release()
	if _, err := connection.Exec(ctx, "SELECT pg_advisory_lock($1)", localIdentityMigrationLock); err != nil {
		return State{}, errors.New("lock local identity migration")
	}
	defer func() {
		_, _ = connection.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", localIdentityMigrationLock)
	}()
	transaction, err := connection.Begin(ctx)
	if err != nil {
		return State{}, errors.New("begin local identity migration")
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	if _, err := transaction.Exec(ctx, schemaMarkerSQL); err != nil {
		return State{}, errors.New("create local identity migration marker")
	}
	state, err := inspect(ctx, transaction)
	if err != nil {
		return State{}, err
	}
	if state.MigrationState == MigrationStateApplied {
		if err := transaction.Commit(ctx); err != nil {
			return State{}, errors.New("commit local identity migration check")
		}
		return state, nil
	}
	if state.MigrationState == MigrationStateMismatch {
		return State{}, errors.New("local identity migration marker mismatch")
	}
	if state.MigrationState == MigrationStateUpgradeRequired {
		if state.MigrationID == legacyMigrationID {
			if _, err := transaction.Exec(ctx, oidcAuthorizationUpSQL); err != nil {
				return State{}, errors.New("upgrade local identity OIDC authorization migration")
			}
		}
		if _, err := transaction.Exec(ctx, administrationUpSQL); err != nil {
			return State{}, errors.New("upgrade local identity administration migration")
		}
		if _, err := transaction.Exec(ctx, `UPDATE local_identity_schema_versions SET
            migration_id=$2, store_schema_version=$3, migration_checksum=$4, applied_at=now() WHERE component=$1`,
			Component, MigrationID, StoreSchemaVersion, ExpectedChecksum()); err != nil {
			return State{}, errors.New("update local identity migration marker")
		}
	} else {
		if _, err := transaction.Exec(ctx, legacyUpSQL); err != nil {
			return State{}, errors.New("apply local identity base migration")
		}
		if _, err := transaction.Exec(ctx, oidcAuthorizationUpSQL); err != nil {
			return State{}, errors.New("apply local identity OIDC authorization migration")
		}
		if _, err := transaction.Exec(ctx, administrationUpSQL); err != nil {
			return State{}, errors.New("apply local identity administration migration")
		}
		if _, err := transaction.Exec(ctx, `INSERT INTO local_identity_schema_versions
            (component, migration_id, store_schema_version, migration_checksum) VALUES ($1,$2,$3,$4)`,
			Component, MigrationID, StoreSchemaVersion, ExpectedChecksum()); err != nil {
			return State{}, errors.New("write local identity migration marker")
		}
	}
	if err := transaction.Commit(ctx); err != nil {
		return State{}, errors.New("commit local identity migration")
	}
	return Inspect(ctx, pool)
}

func RollbackForDevTest(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("local identity PostgreSQL pool is missing")
	}
	connection, err := pool.Acquire(ctx)
	if err != nil {
		return State{}, errors.New("acquire local identity rollback connection")
	}
	defer connection.Release()
	if _, err := connection.Exec(ctx, "SELECT pg_advisory_lock($1)", localIdentityMigrationLock); err != nil {
		return State{}, errors.New("lock local identity rollback")
	}
	defer func() {
		_, _ = connection.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", localIdentityMigrationLock)
	}()
	transaction, err := connection.Begin(ctx)
	if err != nil {
		return State{}, errors.New("begin local identity rollback")
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	state, err := inspect(ctx, transaction)
	if err != nil {
		return State{}, err
	}
	if state.MigrationState == MigrationStateNotApplied {
		if err := transaction.Commit(ctx); err != nil {
			return State{}, errors.New("commit local identity rollback check")
		}
		return state, nil
	}
	if state.MigrationState != MigrationStateApplied && state.MigrationState != MigrationStateUpgradeRequired {
		return State{}, errors.New("local identity rollback requires matching migration marker")
	}
	if state.MigrationState == MigrationStateApplied {
		if _, err := transaction.Exec(ctx, administrationDownSQL); err != nil {
			return State{}, errors.New("rollback local identity administration migration")
		}
		if _, err := transaction.Exec(ctx, oidcAuthorizationDownSQL); err != nil {
			return State{}, errors.New("rollback local identity OIDC authorization migration")
		}
	} else if state.MigrationID == oidcMigrationID {
		if _, err := transaction.Exec(ctx, oidcAuthorizationDownSQL); err != nil {
			return State{}, errors.New("rollback local identity OIDC authorization migration")
		}
	}
	if _, err := transaction.Exec(ctx, legacyDownSQL); err != nil {
		return State{}, errors.New("rollback local identity migration")
	}
	if err := transaction.Commit(ctx); err != nil {
		return State{}, errors.New("commit local identity rollback")
	}
	return State{MigrationState: MigrationStateNotApplied}, nil
}

func inspect(ctx context.Context, query rowQuerier) (State, error) {
	var exists bool
	if err := query.QueryRow(ctx, "SELECT to_regclass('public.local_identity_schema_versions') IS NOT NULL").Scan(&exists); err != nil {
		return State{}, errors.New("inspect local identity migration marker")
	}
	if !exists {
		return State{MigrationState: MigrationStateNotApplied}, nil
	}
	state := State{}
	err := query.QueryRow(ctx, `SELECT migration_id, store_schema_version, migration_checksum, applied_at
        FROM local_identity_schema_versions WHERE component=$1`, Component).Scan(
		&state.MigrationID, &state.StoreSchemaVersion, &state.MigrationChecksum, &state.AppliedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return State{MigrationState: MigrationStateNotApplied}, nil
	}
	if err != nil {
		return State{}, errors.New("read local identity migration marker")
	}
	for _, table := range []string{
		"local_user_accounts", "local_credentials", "external_identity_bindings", "local_web_sessions",
		"local_role_assignments", "local_workspace_memberships",
	} {
		var tableExists bool
		if err := query.QueryRow(ctx, "SELECT to_regclass($1) IS NOT NULL", "public."+table).Scan(&tableExists); err != nil {
			return State{}, errors.New("inspect local identity table")
		}
		if !tableExists {
			state.MigrationState = MigrationStateMismatch
			return state, nil
		}
	}
	if state.MigrationID == legacyMigrationID && state.StoreSchemaVersion == legacyStoreSchemaVersion &&
		state.MigrationChecksum == legacyExpectedChecksum() {
		state.MigrationState = MigrationStateUpgradeRequired
		return state, nil
	}
	if state.MigrationID == oidcMigrationID && state.StoreSchemaVersion == oidcStoreSchemaVersion &&
		state.MigrationChecksum == oidcExpectedChecksum() {
		state.MigrationState = MigrationStateUpgradeRequired
		return state, nil
	}
	state.MigrationState = MigrationStateApplied
	if state.MigrationID != MigrationID || state.StoreSchemaVersion != StoreSchemaVersion || state.MigrationChecksum != ExpectedChecksum() {
		state.MigrationState = MigrationStateMismatch
		return state, nil
	}
	var oidcTableExists bool
	if err := query.QueryRow(ctx, "SELECT to_regclass('public.local_identity_oidc_authorization_transactions') IS NOT NULL").Scan(&oidcTableExists); err != nil {
		return State{}, errors.New("inspect local identity OIDC authorization table")
	}
	if !oidcTableExists {
		state.MigrationState = MigrationStateMismatch
		return state, nil
	}
	var catalogColumnCount int
	if err := query.QueryRow(ctx, `SELECT count(*) FROM information_schema.columns
        WHERE table_schema='public' AND table_name='local_role_assignments'
        AND column_name IN ('role_catalog_version','role_definition_digest')`).Scan(&catalogColumnCount); err != nil {
		return State{}, errors.New("inspect local identity administration columns")
	}
	var directoryIndexExists bool
	if err := query.QueryRow(ctx, "SELECT to_regclass('public.local_workspace_memberships_directory_idx') IS NOT NULL").Scan(&directoryIndexExists); err != nil {
		return State{}, errors.New("inspect local identity administration index")
	}
	if catalogColumnCount != 2 || !directoryIndexExists {
		state.MigrationState = MigrationStateMismatch
	}
	return state, nil
}
