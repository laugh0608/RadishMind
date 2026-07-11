package workflowrunmigrations

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
	Component                              = "workflow_runs"
	MigrationID                            = "0005_workflow_evaluation_suites"
	StoreSchemaVersion                     = "workflow_run_store_v5"
	legacyMigrationID                      = "0001_workflow_runs"
	legacyStoreSchemaVersion               = "workflow_run_store_v1"
	diagnosticsMigrationID                 = "0002_workflow_run_diagnostics"
	diagnosticsStoreSchemaVersion          = "workflow_run_store_v2"
	evaluationMigrationID                  = "0003_workflow_evaluation_cases"
	evaluationStoreSchemaVersion           = "workflow_run_store_v3"
	caseVersioningMigrationID              = "0004_workflow_evaluation_case_revisions"
	caseVersioningStoreSchemaVersion       = "workflow_run_store_v4"
	MigrationStateApplied                  = "applied"
	MigrationStatePending                  = "pending"
	MigrationStateNotApplied               = "not_applied"
	MigrationStateMismatch                 = "mismatch"
	migrationAdvisoryLockKey         int64 = 0x524d52554e533031
)

const markerSQL = `CREATE TABLE IF NOT EXISTS workflow_run_schema_versions (
component text PRIMARY KEY, migration_id text NOT NULL, store_schema_version text NOT NULL,
migration_checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now());`

//go:embed 0001_workflow_runs.up.sql
var upSQLV1 string

//go:embed 0001_workflow_runs.down.sql
var downSQLV1 string

//go:embed 0002_workflow_run_diagnostics.up.sql
var upSQLV2 string

//go:embed 0002_workflow_run_diagnostics.down.sql
var downSQLV2 string

//go:embed 0003_workflow_evaluation_cases.up.sql
var upSQLV3 string

//go:embed 0003_workflow_evaluation_cases.down.sql
var downSQLV3 string

//go:embed 0004_workflow_evaluation_case_revisions.up.sql
var upSQLV4 string

//go:embed 0004_workflow_evaluation_case_revisions.down.sql
var downSQLV4 string

//go:embed 0005_workflow_evaluation_suites.up.sql
var upSQLV5 string

//go:embed 0005_workflow_evaluation_suites.down.sql
var downSQLV5 string

var upSQL = upSQLV1 + "\n" + upSQLV2 + "\n" + upSQLV3 + "\n" + upSQLV4 + "\n" + upSQLV5
var downSQL = downSQLV5 + "\n" + downSQLV4 + "\n" + downSQLV3 + "\n" + downSQLV2 + "\n" + downSQLV1

type State struct {
	MigrationState, MigrationID, StoreSchemaVersion, MigrationChecksum string
	AppliedAt                                                          time.Time
}
type rowQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func OpenPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	if strings.TrimSpace(databaseURL) == "" {
		return nil, errors.New("workflow run PostgreSQL database URL is missing")
	}
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, safeDatabaseError("parse workflow run PostgreSQL config", err)
	}
	cfg.MaxConns, cfg.MinConns = 8, 0
	cfg.MaxConnLifetime, cfg.MaxConnIdleTime, cfg.HealthCheckPeriod = 30*time.Minute, 5*time.Minute, 30*time.Second
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, safeDatabaseError("create workflow run PostgreSQL pool", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, safeDatabaseError("connect workflow run PostgreSQL", err)
	}
	return pool, nil
}

func ExpectedChecksum() string { return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(upSQL))) }

func legacyChecksum() string { return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(upSQLV1))) }
func diagnosticsChecksum() string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(upSQLV1+"\n"+upSQLV2)))
}
func evaluationChecksum() string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(upSQLV1+"\n"+upSQLV2+"\n"+upSQLV3)))
}
func caseVersioningChecksum() string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(upSQLV1+"\n"+upSQLV2+"\n"+upSQLV3+"\n"+upSQLV4)))
}

func Inspect(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("workflow run PostgreSQL pool is missing")
	}
	return inspect(ctx, pool)
}

func Apply(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("workflow run PostgreSQL pool is missing")
	}
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return State{}, safeDatabaseError("acquire workflow run migration connection", err)
	}
	defer conn.Release()
	if _, err = conn.Exec(ctx, "SELECT pg_advisory_lock($1)", migrationAdvisoryLockKey); err != nil {
		return State{}, safeDatabaseError("lock workflow run migration", err)
	}
	defer func() {
		unlock, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, _ = conn.Exec(unlock, "SELECT pg_advisory_unlock($1)", migrationAdvisoryLockKey)
	}()
	tx, err := conn.Begin(ctx)
	if err != nil {
		return State{}, safeDatabaseError("begin workflow run migration", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err = tx.Exec(ctx, markerSQL); err != nil {
		return State{}, safeDatabaseError("create workflow run schema marker", err)
	}
	state, err := inspect(ctx, tx)
	if err != nil {
		return State{}, err
	}
	if state.MigrationState == MigrationStateApplied {
		if err = tx.Commit(ctx); err != nil {
			return State{}, safeDatabaseError("commit workflow run migration check", err)
		}
		return state, nil
	}
	if state.MigrationState == MigrationStateMismatch {
		return State{}, errors.New("workflow run migration marker mismatch")
	}
	if state.MigrationState == MigrationStatePending {
		pendingSQL := upSQLV5
		if state.MigrationID == legacyMigrationID {
			pendingSQL = upSQLV2 + "\n" + upSQLV3 + "\n" + upSQLV4 + "\n" + upSQLV5
		} else if state.MigrationID == diagnosticsMigrationID {
			pendingSQL = upSQLV3 + "\n" + upSQLV4 + "\n" + upSQLV5
		} else if state.MigrationID == evaluationMigrationID {
			pendingSQL = upSQLV4 + "\n" + upSQLV5
		}
		if _, err = tx.Exec(ctx, pendingSQL); err != nil {
			return State{}, safeDatabaseError("apply workflow run pending migrations", err)
		}
		if _, err = tx.Exec(ctx, `UPDATE workflow_run_schema_versions SET migration_id=$2,store_schema_version=$3,migration_checksum=$4,applied_at=now() WHERE component=$1`, Component, MigrationID, StoreSchemaVersion, ExpectedChecksum()); err != nil {
			return State{}, safeDatabaseError("update workflow run migration marker", err)
		}
	} else if _, err = tx.Exec(ctx, upSQL); err != nil {
		return State{}, safeDatabaseError("apply workflow run migration", err)
	} else if _, err = tx.Exec(ctx, `INSERT INTO workflow_run_schema_versions (component,migration_id,store_schema_version,migration_checksum) VALUES ($1,$2,$3,$4)`, Component, MigrationID, StoreSchemaVersion, ExpectedChecksum()); err != nil {
		return State{}, safeDatabaseError("write workflow run migration marker", err)
	}
	state, err = inspect(ctx, tx)
	if err != nil || state.MigrationState != MigrationStateApplied {
		return State{}, errors.New("workflow run migration preflight failed")
	}
	if err = tx.Commit(ctx); err != nil {
		return State{}, safeDatabaseError("commit workflow run migration", err)
	}
	return state, nil
}

func RollbackForDevTest(ctx context.Context, pool *pgxpool.Pool) (State, error) {
	if pool == nil {
		return State{}, errors.New("workflow run PostgreSQL pool is missing")
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return State{}, safeDatabaseError("begin workflow run rollback", err)
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
	if state.MigrationState != MigrationStateApplied && state.MigrationState != MigrationStatePending {
		return State{}, errors.New("workflow run rollback requires matching migration")
	}
	rollbackSQL := downSQL
	if state.MigrationState == MigrationStatePending {
		rollbackSQL = downSQLV4 + "\n" + downSQLV3 + "\n" + downSQLV2 + "\n" + downSQLV1
		if state.MigrationID == legacyMigrationID {
			rollbackSQL = downSQLV1
		} else if state.MigrationID == diagnosticsMigrationID {
			rollbackSQL = downSQLV2 + "\n" + downSQLV1
		} else if state.MigrationID == evaluationMigrationID {
			rollbackSQL = downSQLV3 + "\n" + downSQLV2 + "\n" + downSQLV1
		}
	}
	if _, err = tx.Exec(ctx, rollbackSQL); err != nil {
		return State{}, safeDatabaseError("rollback workflow run migration", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return State{}, safeDatabaseError("commit workflow run rollback", err)
	}
	return State{MigrationState: MigrationStateNotApplied}, nil
}

func inspect(ctx context.Context, query rowQuerier) (State, error) {
	var exists bool
	if err := query.QueryRow(ctx, "SELECT to_regclass('public.workflow_run_schema_versions') IS NOT NULL").Scan(&exists); err != nil {
		return State{}, safeDatabaseError("inspect workflow run marker", err)
	}
	if !exists {
		return State{MigrationState: MigrationStateNotApplied}, nil
	}
	state := State{}
	err := query.QueryRow(ctx, `SELECT migration_id,store_schema_version,migration_checksum,applied_at FROM workflow_run_schema_versions WHERE component=$1`, Component).Scan(&state.MigrationID, &state.StoreSchemaVersion, &state.MigrationChecksum, &state.AppliedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return State{MigrationState: MigrationStateNotApplied}, nil
	}
	if err != nil {
		return State{}, safeDatabaseError("read workflow run marker", err)
	}
	var tableExists bool
	if err = query.QueryRow(ctx, "SELECT to_regclass('public.workflow_run_records') IS NOT NULL").Scan(&tableExists); err != nil {
		return State{}, safeDatabaseError("inspect workflow run table", err)
	}
	if state.MigrationID == legacyMigrationID && state.StoreSchemaVersion == legacyStoreSchemaVersion && state.MigrationChecksum == legacyChecksum() && tableExists {
		state.MigrationState = MigrationStatePending
	} else if state.MigrationID == diagnosticsMigrationID && state.StoreSchemaVersion == diagnosticsStoreSchemaVersion && state.MigrationChecksum == diagnosticsChecksum() && tableExists {
		state.MigrationState = MigrationStatePending
	} else if state.MigrationID == evaluationMigrationID && state.StoreSchemaVersion == evaluationStoreSchemaVersion && state.MigrationChecksum == evaluationChecksum() && tableExists {
		state.MigrationState = MigrationStatePending
	} else if state.MigrationID == caseVersioningMigrationID && state.StoreSchemaVersion == caseVersioningStoreSchemaVersion && state.MigrationChecksum == caseVersioningChecksum() && tableExists {
		state.MigrationState = MigrationStatePending
	} else {
		var diagnosticColumnCount int
		if err = query.QueryRow(ctx, `SELECT count(*) FROM information_schema.columns WHERE table_schema='public' AND table_name='workflow_run_records' AND column_name IN ('failure_code','failure_boundary','selected_provider','selected_model')`).Scan(&diagnosticColumnCount); err != nil {
			return State{}, safeDatabaseError("inspect workflow run diagnostics columns", err)
		}
		var evaluationTableExists bool
		if err = query.QueryRow(ctx, "SELECT to_regclass('public.workflow_evaluation_cases') IS NOT NULL").Scan(&evaluationTableExists); err != nil {
			return State{}, safeDatabaseError("inspect workflow evaluation table", err)
		}
		var revisionTableExists bool
		if err = query.QueryRow(ctx, "SELECT to_regclass('public.workflow_evaluation_case_revisions') IS NOT NULL").Scan(&revisionTableExists); err != nil {
			return State{}, safeDatabaseError("inspect workflow evaluation revision table", err)
		}
		var currentVersionColumnCount int
		if err = query.QueryRow(ctx, `SELECT count(*) FROM information_schema.columns WHERE table_schema='public' AND table_name='workflow_evaluation_cases' AND column_name='current_version'`).Scan(&currentVersionColumnCount); err != nil {
			return State{}, safeDatabaseError("inspect workflow evaluation version column", err)
		}
		var suiteTableExists, decisionTableExists bool
		if err = query.QueryRow(ctx, "SELECT to_regclass('public.workflow_evaluation_suites') IS NOT NULL").Scan(&suiteTableExists); err != nil {
			return State{}, safeDatabaseError("inspect workflow evaluation suite table", err)
		}
		if err = query.QueryRow(ctx, "SELECT to_regclass('public.workflow_evaluation_suite_decisions') IS NOT NULL").Scan(&decisionTableExists); err != nil {
			return State{}, safeDatabaseError("inspect workflow evaluation suite decision table", err)
		}
		if state.MigrationID != MigrationID || state.StoreSchemaVersion != StoreSchemaVersion || state.MigrationChecksum != ExpectedChecksum() || !tableExists || diagnosticColumnCount != 4 || !evaluationTableExists || !revisionTableExists || currentVersionColumnCount != 1 || !suiteTableExists || !decisionTableExists {
			state.MigrationState = MigrationStateMismatch
		} else {
			state.MigrationState = MigrationStateApplied
		}
	}
	return state, nil
}

func safeDatabaseError(operation string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return fmt.Errorf("%s failed (SQLSTATE %s)", operation, pgErr.Code)
	}
	return fmt.Errorf("%s failed", operation)
}
