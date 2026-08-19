//go:build postgres_integration

package httpapi

import (
	"context"
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

func resetPostgresLocalIdentitySchema(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	for _, table := range []string{
		"local_workspace_memberships", "local_role_assignments", "local_web_sessions",
		"external_identity_bindings", "local_credentials", "local_user_accounts", "local_identity_schema_versions",
	} {
		if _, err := pool.Exec(ctx, "DROP TABLE IF EXISTS "+table); err != nil {
			t.Fatalf("drop local identity integration table %s: %v", table, err)
		}
	}
}
