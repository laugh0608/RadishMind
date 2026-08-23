//go:build postgres_integration

package httpapi

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	localidentitymigrations "radishmind.local/services/platform/migrations/local_identity_records"
)

func TestLocalIdentityPostgresMigrationRepositoryRuntimeRoleRestartAndRollback(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeDatabaseURL := postgresIntegrationDatabaseURLForCredentials(
		t, runtimeUser, os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	adminPool, err := localidentitymigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, adminPool)
	resetPostgresLocalIdentitySchema(t, ctx, adminPool)
	preparePostgresIntegrationRuntimeRole(t, ctx, adminPool, runtimeUser)
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cleanupCancel()
		resetPostgresLocalIdentitySchema(t, cleanupContext, adminPool)
		adminPool.Close()
	})

	installPostgresLocalIdentityV1Schema(t, ctx, adminPool)
	legacyState, err := localidentitymigrations.Inspect(ctx, adminPool)
	if err != nil || legacyState.MigrationState != localidentitymigrations.MigrationStateUpgradeRequired {
		t.Fatalf("inspect local identity v1 migration: state=%#v err=%v", legacyState, err)
	}
	state, err := localidentitymigrations.Apply(ctx, adminPool)
	if err != nil || state.MigrationState != localidentitymigrations.MigrationStateApplied {
		t.Fatalf("apply local identity migration: state=%#v err=%v", state, err)
	}
	if repeated, repeatErr := localidentitymigrations.Apply(ctx, adminPool); repeatErr != nil ||
		repeated.MigrationChecksum != state.MigrationChecksum {
		t.Fatalf("repeat local identity migration: state=%#v err=%v", repeated, repeatErr)
	}

	runtimePool, err := localidentitymigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	if _, ddlErr := runtimePool.Exec(ctx, "CREATE TABLE local_identity_runtime_must_not_create (id integer)"); ddlErr == nil {
		t.Fatal("local identity runtime role must not have schema DDL permission")
	}
	repository := newPostgresLocalIdentityRepository(runtimePool)
	runLocalIdentityRepositoryContract(t, repository)
	runConcurrentLocalIdentityBindingCreate(t, repository)
	pending := localIdentityTestOIDCTransaction(
		"oat_postgresrestartaa", "postgres-restart-state-with-more-than-thirty-two-characters", localIdentityTestNow.Add(3*time.Hour),
	)
	if err := repository.CreateOIDCAuthorizationTransaction(ctx, pending); err != nil {
		t.Fatalf("create PostgreSQL restart OIDC authorization transaction: %v", err)
	}
	runtimePool.Close()

	restartedPool, err := localidentitymigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	restarted := newPostgresLocalIdentityRepository(restartedPool)
	account, err := restarted.ReadAccount(ctx, "usr_aaaaaaaaaaaaaaaa")
	if err != nil || account.LifecycleState != localIdentityStateDisabled || account.RecordVersion != 2 {
		t.Fatalf("restart local identity account mismatch: account=%#v err=%v", account, err)
	}
	restartStateDigest, _ := localIdentityOIDCStateDigest("postgres-restart-state-with-more-than-thirty-two-characters")
	if restoredTransaction, consumeErr := restarted.ConsumeOIDCAuthorizationTransaction(ctx, restartStateDigest, localIdentityTestNow.Add(3*time.Hour+time.Minute)); consumeErr != nil ||
		restoredTransaction.TransactionID != pending.TransactionID || restoredTransaction.codeVerifier == "" {
		t.Fatalf("restart PostgreSQL OIDC authorization transaction mismatch: transaction=%#v err=%v", restoredTransaction, consumeErr)
	}
	restartedPool.Close()

	rolledBack, err := localidentitymigrations.RollbackForDevTest(ctx, adminPool)
	if err != nil || rolledBack.MigrationState != localidentitymigrations.MigrationStateNotApplied {
		t.Fatalf("rollback local identity migration: state=%#v err=%v", rolledBack, err)
	}
	reapplied, err := localidentitymigrations.Apply(ctx, adminPool)
	if err != nil || reapplied.MigrationState != localidentitymigrations.MigrationStateApplied {
		t.Fatalf("reapply local identity migration: state=%#v err=%v", reapplied, err)
	}
}

func installPostgresLocalIdentityV1Schema(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	legacySQL, err := os.ReadFile("../../migrations/local_identity_records/0001_local_identity_records.up.sql")
	if err != nil {
		t.Fatalf("read local identity v1 migration: %v", err)
	}
	if _, err := pool.Exec(ctx, string(legacySQL)); err != nil {
		t.Fatalf("apply local identity v1 migration: %v", err)
	}
	if _, err := pool.Exec(ctx, `CREATE TABLE local_identity_schema_versions (
		component text PRIMARY KEY, migration_id text NOT NULL, store_schema_version text NOT NULL,
		migration_checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		t.Fatalf("create local identity v1 migration marker: %v", err)
	}
	checksum := fmt.Sprintf("sha256:%x", sha256.Sum256(legacySQL))
	if _, err := pool.Exec(ctx, `INSERT INTO local_identity_schema_versions
		(component, migration_id, store_schema_version, migration_checksum)
		VALUES ('local_identity_records','0001_local_identity_records','local_identity_records_store_v1',$1)`, checksum); err != nil {
		t.Fatalf("write local identity v1 migration marker: %v", err)
	}
}

func resetPostgresLocalIdentitySchema(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	for _, table := range []string{
		"local_identity_oidc_authorization_transactions",
		"local_workspace_memberships", "local_role_assignments", "local_web_sessions",
		"external_identity_bindings", "local_credentials", "local_user_accounts", "local_identity_schema_versions",
	} {
		if _, err := pool.Exec(ctx, "DROP TABLE IF EXISTS "+table); err != nil {
			t.Fatalf("drop local identity integration table %s: %v", table, err)
		}
	}
}
