//go:build postgres_integration

package httpapi

import (
	"bytes"
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

func TestLocalIdentityPostgresV3SelfServiceUpgrade(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	pool, err := localidentitymigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, pool)
	resetPostgresLocalIdentitySchema(t, ctx, pool)
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cleanupCancel()
		resetPostgresLocalIdentitySchema(t, cleanupContext, pool)
		pool.Close()
	})
	installPostgresLocalIdentityV3Schema(t, ctx, pool)
	legacyState, err := localidentitymigrations.Inspect(ctx, pool)
	if err != nil || legacyState.MigrationState != localidentitymigrations.MigrationStateUpgradeRequired {
		t.Fatalf("inspect local identity v3 migration: state=%#v err=%v", legacyState, err)
	}
	account, credential := localIdentityTestAccount(
		"usr_90000000000000f3", "cred_90000000000000f3", "postgres-v3-upgrade@example.com", localIdentityTestNow,
	)
	if err := newPostgresLocalIdentityRepository(pool).CreateAccount(ctx, account, credential); err != nil {
		t.Fatalf("create pre-self-service PostgreSQL account: %v", err)
	}
	state, err := localidentitymigrations.Apply(ctx, pool)
	if err != nil || state.MigrationState != localidentitymigrations.MigrationStateApplied ||
		state.MigrationID != localidentitymigrations.MigrationID {
		t.Fatalf("upgrade local identity v3 migration: state=%#v err=%v", state, err)
	}
	if _, err := newPostgresLocalIdentityRepository(pool).ReadAccount(ctx, account.UserID); err != nil {
		t.Fatalf("read account after PostgreSQL v3 self-service upgrade: %v", err)
	}
}

func TestLocalIdentityPostgresV4WorkspaceInvitationUpgrade(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	pool, err := localidentitymigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, pool)
	resetPostgresLocalIdentitySchema(t, ctx, pool)
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cleanupCancel()
		resetPostgresLocalIdentitySchema(t, cleanupContext, pool)
		pool.Close()
	})
	installPostgresLocalIdentityV4Schema(t, ctx, pool)
	legacyState, err := localidentitymigrations.Inspect(ctx, pool)
	if err != nil || legacyState.MigrationState != localidentitymigrations.MigrationStateUpgradeRequired {
		t.Fatalf("inspect local identity v4 migration: state=%#v err=%v", legacyState, err)
	}
	account, credential := localIdentityTestAccount(
		"usr_90000000000000f4", "cred_90000000000000f4", "postgres-v4-upgrade@example.com", localIdentityTestNow,
	)
	if err := newPostgresLocalIdentityRepository(pool).CreateAccount(ctx, account, credential); err != nil {
		t.Fatalf("create pre-invitation PostgreSQL account: %v", err)
	}
	state, err := localidentitymigrations.Apply(ctx, pool)
	if err != nil || state.MigrationState != localidentitymigrations.MigrationStateApplied ||
		state.MigrationID != localidentitymigrations.MigrationID {
		t.Fatalf("upgrade local identity v4 migration: state=%#v err=%v", state, err)
	}
	if _, err := newPostgresLocalIdentityRepository(pool).ReadAccount(ctx, account.UserID); err != nil {
		t.Fatalf("read account after PostgreSQL v4 invitation upgrade: %v", err)
	}
}

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
	selfServiceState := runDurableLocalIdentitySelfServiceContract(t, repository)
	invitationState := runDurableWorkspaceInvitationContract(t, repository)
	assertPostgresLocalIdentityDirectoryQueryPlan(t, ctx, adminPool, runtimePool)
	assertPostgresLocalIdentitySelfServiceQueryPlan(t, ctx, adminPool, runtimePool)
	assertPostgresWorkspaceInvitationQueryPlan(t, ctx, adminPool, runtimePool)
	assertPostgresWorkspaceInvitationCorruptPayloadFailsClosed(t, ctx, adminPool, repository, invitationState)
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
	parsedInvitation, parseErr := parseWorkspaceInvitationCode(invitationState.InvitationCode)
	if parseErr != nil {
		t.Fatalf("parse PostgreSQL invitation code for unavailable test: %v", parseErr)
	}
	if _, unavailableErr := repository.PreviewWorkspaceInvitation(
		ctx, invitationState.Claimant.UserID, invitationState.Claimant.TenantRef,
		parsedInvitation.InvitationID, digestWorkspaceInvitationSecret(parsedInvitation.secret), invitationState.Now,
	); !errors.Is(unavailableErr, errWorkspaceInvitationStoreUnavailable) {
		t.Fatalf("closed PostgreSQL invitation owner fell back instead of failing closed: %v", unavailableErr)
	}

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
	assertDurableLocalIdentitySelfServiceRestart(t, restarted, selfServiceState)
	assertDurableWorkspaceInvitationRestart(t, restarted, invitationState)
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

func assertPostgresLocalIdentitySelfServiceQueryPlan(
	t *testing.T,
	ctx context.Context,
	adminPool *pgxpool.Pool,
	pool *pgxpool.Pool,
) {
	t.Helper()
	if _, err := adminPool.Exec(ctx, "ANALYZE local_web_sessions"); err != nil {
		t.Fatalf("analyze PostgreSQL self-service session statistics: %v", err)
	}
	transaction, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin PostgreSQL self-service query plan transaction: %v", err)
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	if _, err := transaction.Exec(ctx, "SET LOCAL enable_seqscan=off"); err != nil {
		t.Fatalf("configure PostgreSQL self-service query plan: %v", err)
	}
	rows, err := transaction.Query(ctx, `EXPLAIN (FORMAT TEXT)
        SELECT session_id FROM local_web_sessions
        WHERE user_id=$1 AND created_at<=$2
        ORDER BY created_at DESC, session_id DESC LIMIT 51`,
		"usr_9000000000000001", localIdentitySelfServiceTestNow,
	)
	if err != nil {
		t.Fatalf("explain PostgreSQL self-service session query: %v", err)
	}
	defer rows.Close()
	plan := strings.Builder{}
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatalf("scan PostgreSQL self-service query plan: %v", err)
		}
		plan.WriteString(line)
	}
	if rows.Err() != nil || !strings.Contains(plan.String(), "local_web_sessions_self_service_list_idx") {
		t.Fatalf("PostgreSQL self-service query did not use its ordered index: %s", plan.String())
	}
}

func assertPostgresWorkspaceInvitationQueryPlan(
	t *testing.T,
	ctx context.Context,
	adminPool *pgxpool.Pool,
	pool *pgxpool.Pool,
) {
	t.Helper()
	if _, err := adminPool.Exec(ctx, "ANALYZE local_workspace_invitations"); err != nil {
		t.Fatalf("analyze PostgreSQL invitation statistics: %v", err)
	}
	transaction, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin PostgreSQL invitation query plan transaction: %v", err)
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	if _, err := transaction.Exec(ctx, "SET LOCAL enable_seqscan=off"); err != nil {
		t.Fatalf("configure PostgreSQL invitation query plan: %v", err)
	}
	rows, err := transaction.Query(ctx, `EXPLAIN (FORMAT TEXT)
        SELECT invitation_id FROM local_workspace_invitations
        WHERE tenant_ref=$1 AND workspace_id=$2 AND lifecycle_state=$3
        ORDER BY updated_at DESC, invitation_id DESC LIMIT 51`,
		"tenant_invitation", "workspace_invitation", workspaceInvitationLifecyclePending,
	)
	if err != nil {
		t.Fatalf("explain PostgreSQL invitation directory query: %v", err)
	}
	defer rows.Close()
	plan := strings.Builder{}
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatalf("scan PostgreSQL invitation query plan: %v", err)
		}
		plan.WriteString(line)
	}
	if rows.Err() != nil || !strings.Contains(plan.String(), "local_workspace_invitations_directory_idx") {
		t.Fatalf("PostgreSQL invitation query did not use ordered index: %s", plan.String())
	}
}

func assertPostgresWorkspaceInvitationCorruptPayloadFailsClosed(
	t *testing.T,
	ctx context.Context,
	adminPool *pgxpool.Pool,
	repository durableWorkspaceInvitationRepository,
	state durableWorkspaceInvitationState,
) {
	t.Helper()
	service := newWorkspaceInvitationService(repository)
	service.now = func() time.Time { return state.Now }
	service.newID = func(prefix string) (string, error) {
		switch prefix {
		case "wsi_":
			return "wsi_00000000000000000000000000005bad", nil
		case "mbr_":
			return "mbr_0000000000005bad", nil
		default:
			return "rla_0000000000005bad", nil
		}
	}
	reader, _ := builtInLocalIdentityRole(localIdentityRoleWorkspaceReader)
	creation, err := service.Create(ctx, state.Administrator, WorkspaceInvitationCreateInput{
		TenantRef: state.TenantRef, WorkspaceID: state.WorkspaceID, RoleKey: reader.RoleKey,
		ExpectedCatalogVersion: reader.CatalogVersion, ExpectedRoleDefinitionDigest: reader.DefinitionDigest,
		TTLPolicy: workspaceInvitationTTL24Hours, Confirmed: true,
		RequestRef: "request:postgres-corrupt-invitation", AuditRef: "audit:postgres-corrupt-invitation",
	})
	if err != nil {
		t.Fatalf("create PostgreSQL invitation corruption probe: %v", err)
	}
	if _, err := adminPool.Exec(ctx, `UPDATE local_workspace_invitations
        SET schema_version='corrupt.workspace_invitation' WHERE invitation_id=$1`, creation.Invitation.InvitationID); err != nil {
		t.Fatalf("corrupt PostgreSQL invitation payload: %v", err)
	}
	if _, err := service.Preview(
		ctx, state.Claimant, creation.InvitationCode,
	); !errors.Is(err, errWorkspaceInvitationStoreUnavailable) {
		t.Fatalf("corrupt PostgreSQL invitation did not fail closed: %v", err)
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

func installPostgresLocalIdentityV3Schema(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	paths := []string{
		"../../migrations/local_identity_records/0001_local_identity_records.up.sql",
		"../../migrations/local_identity_records/0002_local_identity_oidc_authorization_transactions.up.sql",
		"../../migrations/local_identity_records/0003_local_identity_administration.up.sql",
	}
	parts := make([][]byte, 0, len(paths))
	for _, path := range paths {
		migrationSQL, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read local identity v3 migration part %s: %v", path, err)
		}
		if _, err := pool.Exec(ctx, string(migrationSQL)); err != nil {
			t.Fatalf("apply local identity v3 migration part %s: %v", path, err)
		}
		parts = append(parts, migrationSQL)
	}
	if _, err := pool.Exec(ctx, `CREATE TABLE local_identity_schema_versions (
		component text PRIMARY KEY, migration_id text NOT NULL, store_schema_version text NOT NULL,
		migration_checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		t.Fatalf("create local identity v3 migration marker: %v", err)
	}
	checksum := fmt.Sprintf("sha256:%x", sha256.Sum256(bytes.Join(parts, []byte("\n"))))
	if _, err := pool.Exec(ctx, `INSERT INTO local_identity_schema_versions
		(component, migration_id, store_schema_version, migration_checksum)
		VALUES ('local_identity_records','0003_local_identity_administration','local_identity_records_store_v3',$1)`, checksum); err != nil {
		t.Fatalf("write local identity v3 migration marker: %v", err)
	}
}

func installPostgresLocalIdentityV4Schema(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	paths := []string{
		"../../migrations/local_identity_records/0001_local_identity_records.up.sql",
		"../../migrations/local_identity_records/0002_local_identity_oidc_authorization_transactions.up.sql",
		"../../migrations/local_identity_records/0003_local_identity_administration.up.sql",
		"../../migrations/local_identity_records/0004_local_identity_self_service_sessions.up.sql",
	}
	parts := make([][]byte, 0, len(paths))
	for _, path := range paths {
		migrationSQL, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read local identity v4 migration part %s: %v", path, err)
		}
		if _, err := pool.Exec(ctx, string(migrationSQL)); err != nil {
			t.Fatalf("apply local identity v4 migration part %s: %v", path, err)
		}
		parts = append(parts, migrationSQL)
	}
	if _, err := pool.Exec(ctx, `CREATE TABLE local_identity_schema_versions (
		component text PRIMARY KEY, migration_id text NOT NULL, store_schema_version text NOT NULL,
		migration_checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		t.Fatalf("create local identity v4 migration marker: %v", err)
	}
	checksum := fmt.Sprintf("sha256:%x", sha256.Sum256(bytes.Join(parts, []byte("\n"))))
	if _, err := pool.Exec(ctx, `INSERT INTO local_identity_schema_versions
		(component, migration_id, store_schema_version, migration_checksum)
		VALUES ('local_identity_records','0004_local_identity_self_service_sessions','local_identity_records_store_v4',$1)`, checksum); err != nil {
		t.Fatalf("write local identity v4 migration marker: %v", err)
	}
}

func resetPostgresLocalIdentitySchema(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	for _, table := range []string{
		"local_workspace_invitations",
		"local_identity_oidc_authorization_transactions",
		"local_workspace_memberships", "local_role_assignments", "local_web_sessions",
		"external_identity_bindings", "local_credentials", "local_user_accounts", "local_identity_schema_versions",
	} {
		if _, err := pool.Exec(ctx, "DROP TABLE IF EXISTS "+table); err != nil {
			t.Fatalf("drop local identity integration table %s: %v", table, err)
		}
	}
}
