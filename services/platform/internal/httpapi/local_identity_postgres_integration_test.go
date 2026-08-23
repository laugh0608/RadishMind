//go:build postgres_integration

package httpapi

import (
	"context"
	"crypto/sha256"
	"errors"
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
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
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
	administrationState := runDurableLocalIdentityAdministrationContract(t, repository)
	assertPostgresLocalIdentityDirectoryQueryPlan(t, ctx, adminPool, runtimePool)
	bootstrapOptions := LocalIdentityDevTestBootstrapOptions{
		StoreMode: localIdentityStoreModePostgresDevTest, PostgresDatabaseURL: runtimeDatabaseURL,
		DatabaseTimeout: 5 * time.Second, TenantRef: "tenant_cli", WorkspaceID: "workspace_cli",
		UserID: administrationState.Administrator.UserID, AuditRef: "audit:postgres-bootstrap-cli",
	}
	cliBootstrap, err := BootstrapLocalIdentityWorkspaceAdministratorDevTest(ctx, bootstrapOptions)
	if err != nil || cliBootstrap.RoleKey != localIdentityRoleWorkspaceAdmin ||
		cliBootstrap.UserID != administrationState.Administrator.UserID {
		t.Fatalf("PostgreSQL explicit bootstrap coordinator mismatch: result=%#v err=%v", cliBootstrap, err)
	}
	if _, err := BootstrapLocalIdentityWorkspaceAdministratorDevTest(ctx, bootstrapOptions); !errors.Is(err, errLocalIdentityAdminBootstrapDenied) {
		t.Fatalf("repeated PostgreSQL bootstrap coordinator was not denied: %v", err)
	}
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
	restartedAdministrationService := newLocalIdentityAdministrationService(restarted)
	restartedAdministrationService.now = func() time.Time { return administrationState.Now.Add(20 * time.Minute) }
	if detail, detailErr := restartedAdministrationService.ReadWorkspaceMember(
		ctx,
		administrationState.Administrator,
		administrationState.TenantRef,
		administrationState.WorkspaceID,
		administrationState.TargetUserID,
	); detailErr != nil || len(detail.Memberships) != 2 || len(detail.RoleAssignments) != 2 ||
		detail.RoleAssignments[0].RoleCatalogVersion != administrationState.CatalogVersion {
		t.Fatalf("restart PostgreSQL administration projection mismatch: detail=%#v err=%v", detail, detailErr)
	}
	restartedPool.Close()

	rolledBack, err := localidentitymigrations.RollbackForDevTest(ctx, adminPool)
	if err != nil || rolledBack.MigrationState != localidentitymigrations.MigrationStateNotApplied {
		t.Fatalf("rollback local identity migration: state=%#v err=%v", rolledBack, err)
	}
	if unavailable, closeUnavailable, unavailableErr := newLocalIdentityRepositoryFromOptions(localIdentityStoreOptions{
		Mode: localIdentityStoreModePostgresDevTest, PostgresDatabaseURL: runtimeDatabaseURL,
		DatabaseTimeout: 5 * time.Second,
	}); unavailableErr == nil || unavailable != nil || closeUnavailable != nil {
		if closeUnavailable != nil {
			closeUnavailable()
		}
		t.Fatalf("rolled-back PostgreSQL administration owner did not fail closed: %v", unavailableErr)
	}
	reapplied, err := localidentitymigrations.Apply(ctx, adminPool)
	if err != nil || reapplied.MigrationState != localidentitymigrations.MigrationStateApplied {
		t.Fatalf("reapply local identity migration: state=%#v err=%v", reapplied, err)
	}
}

func assertPostgresLocalIdentityDirectoryQueryPlan(
	t *testing.T,
	ctx context.Context,
	adminPool *pgxpool.Pool,
	pool *pgxpool.Pool,
) {
	t.Helper()
	if _, err := adminPool.Exec(ctx, "ANALYZE local_workspace_memberships"); err != nil {
		t.Fatalf("analyze PostgreSQL member directory statistics: %v", err)
	}
	transaction, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin PostgreSQL directory query plan transaction: %v", err)
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	if _, err := transaction.Exec(ctx, "SET LOCAL enable_seqscan=off"); err != nil {
		t.Fatalf("configure PostgreSQL directory query plan: %v", err)
	}
	rows, err := transaction.Query(ctx, `EXPLAIN (FORMAT TEXT)
        SELECT membership_id FROM local_workspace_memberships
        WHERE tenant_ref=$1 AND workspace_id=$2 AND lifecycle_state=$3
        ORDER BY updated_at DESC, membership_id DESC LIMIT 51`,
		"tenant_demo", "workspace_demo", localIdentityStateActive,
	)
	if err != nil {
		t.Fatalf("explain PostgreSQL member directory query: %v", err)
	}
	defer rows.Close()
	plan := strings.Builder{}
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatalf("scan PostgreSQL member directory query plan: %v", err)
		}
		plan.WriteString(line)
	}
	if rows.Err() != nil || !strings.Contains(plan.String(), "local_workspace_memberships_directory_idx") {
		t.Fatalf("PostgreSQL member directory query did not use its ordered index: %s", plan.String())
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
