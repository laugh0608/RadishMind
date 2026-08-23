package httpapi

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/sqlitedev"
	sqlitelocalidentitymigrations "radishmind.local/services/platform/migrations/sqlite/local_identity_records"
)

type durableLocalIdentityAdministrationRepository interface {
	localIdentityRepository
	localIdentityAdministrationRepository
}

type durableLocalIdentityAdministrationState struct {
	Now            time.Time
	TenantRef      string
	WorkspaceID    string
	Administrator  LocalIdentityAdministrationActor
	TargetUserID   string
	MembershipID   string
	CatalogVersion string
}

func TestSQLiteLocalIdentityAdministrationContractRestartAndQueryPlan(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "identity-administration.db")
	runtime := openSQLiteLocalIdentityAdministrationRuntime(t, databasePath, sqlitelocalidentitymigrations.Migrations())
	repository := newSQLiteLocalIdentityRepository(runtime.DB())
	state := runDurableLocalIdentityAdministrationContract(t, repository)
	assertSQLiteLocalIdentityDirectoryQueryPlan(t, runtime)
	if err := runtime.Close(); err != nil {
		t.Fatalf("close SQLite administration runtime: %v", err)
	}

	restarted := openSQLiteLocalIdentityAdministrationRuntime(t, databasePath, sqlitelocalidentitymigrations.Migrations())
	t.Cleanup(func() { _ = restarted.Close() })
	restored := newSQLiteLocalIdentityRepository(restarted.DB())
	service := newLocalIdentityAdministrationService(restored)
	service.now = func() time.Time { return state.Now.Add(20 * time.Minute) }
	detail, err := service.ReadWorkspaceMember(
		context.Background(), state.Administrator, state.TenantRef, state.WorkspaceID, state.TargetUserID,
	)
	if err != nil || len(detail.Memberships) != 2 || len(detail.RoleAssignments) != 2 ||
		detail.Memberships[0].LifecycleState != localIdentityStateActive ||
		detail.RoleAssignments[0].RoleCatalogVersion != state.CatalogVersion {
		t.Fatalf("SQLite administration restart projection mismatch: detail=%#v err=%v", detail, err)
	}
}

func TestSQLiteLocalIdentityDevTestBootstrapCoordinatorIsExplicitAndOneShot(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "identity-bootstrap.db")
	runtime := openSQLiteLocalIdentityAdministrationRuntime(t, databasePath, localPersistenceSQLiteMigrations())
	repository := newSQLiteLocalIdentityRepository(runtime.DB())
	account, credential := localIdentityTestAccount(
		"usr_0000000000000b01", "cred_0000000000000b01", "bootstrap-coordinator@example.com", localIdentityTestNow,
	)
	if err := repository.CreateAccount(context.Background(), account, credential); err != nil {
		t.Fatalf("create SQLite bootstrap target account: %v", err)
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("close SQLite bootstrap preparation runtime: %v", err)
	}

	options := LocalIdentityDevTestBootstrapOptions{
		StoreMode: localIdentityStoreModeSQLiteDev, SQLiteDatabasePath: databasePath, DatabaseTimeout: time.Second,
		TenantRef: "tenant_bootstrap", WorkspaceID: "workspace_bootstrap", UserID: account.UserID,
		AuditRef: "audit:sqlite-bootstrap-coordinator",
	}
	created, err := BootstrapLocalIdentityWorkspaceAdministratorDevTest(context.Background(), options)
	if err != nil || created.UserID != account.UserID || created.RoleKey != localIdentityRoleWorkspaceAdmin ||
		created.RoleCatalogVersion != LocalIdentityBuiltInRoleCatalog().CatalogVersion ||
		created.AuditRef != options.AuditRef {
		t.Fatalf("explicit SQLite bootstrap mismatch: result=%#v err=%v", created, err)
	}
	if _, err := BootstrapLocalIdentityWorkspaceAdministratorDevTest(context.Background(), options); !errors.Is(err, errLocalIdentityAdminBootstrapDenied) {
		t.Fatalf("repeated SQLite bootstrap was not denied: %v", err)
	}

	restarted := openSQLiteLocalIdentityAdministrationRuntime(t, databasePath, localPersistenceSQLiteMigrations())
	t.Cleanup(func() { _ = restarted.Close() })
	detail, err := newSQLiteLocalIdentityRepository(restarted.DB()).ReadWorkspaceMember(
		context.Background(), options.TenantRef, options.WorkspaceID, options.UserID, localIdentityTestNow.Add(time.Hour),
	)
	if err != nil || len(detail.Memberships) != 1 || len(detail.RoleAssignments) != 1 ||
		detail.RoleAssignments[0].RoleDefinitionDigest != created.RoleDefinitionDigest {
		t.Fatalf("SQLite bootstrap restart mismatch: detail=%#v err=%v", detail, err)
	}
}

func TestSQLiteLocalIdentityAdministrationMigrationUpgradeAndReapply(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "identity-administration-upgrade.db")
	allMigrations := sqlitelocalidentitymigrations.Migrations()
	legacy := openSQLiteLocalIdentityAdministrationRuntime(t, databasePath, allMigrations[:2])
	repository := newSQLiteLocalIdentityRepository(legacy.DB())
	account, credential := localIdentityTestAccount(
		"usr_0000000000000f01", "cred_0000000000000f01", "legacy-catalog@example.com", localIdentityTestNow,
	)
	if err := repository.CreateAccount(context.Background(), account, credential); err != nil {
		t.Fatalf("create pre-administration SQLite account: %v", err)
	}
	if _, err := legacy.DB().ExecContext(context.Background(), `INSERT INTO local_role_assignments
        (assignment_id,user_id,schema_version,tenant_ref,workspace_id,role_key,permission_grants_json,
         lifecycle_state,record_version,created_at_unix_nano,updated_at_unix_nano,expires_at_unix_nano,
         revoked_at_unix_nano,audit_ref) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		"rla_0000000000000f01", account.UserID, localIdentitySchemaVersion, "tenant_demo", "workspace_demo",
		"workspace_reader", `["applications:read"]`, localIdentityStateActive, 1,
		localIdentityTestNow.UnixNano(), localIdentityTestNow.UnixNano(), nil, nil, "audit:legacy-role",
	); err != nil {
		t.Fatalf("create pre-administration SQLite role: %v", err)
	}
	if err := legacy.Close(); err != nil {
		t.Fatalf("close pre-administration SQLite runtime: %v", err)
	}

	upgraded := openSQLiteLocalIdentityAdministrationRuntime(t, databasePath, allMigrations)
	profile, err := newSQLiteLocalIdentityRepository(upgraded.DB()).ReadAccountAccessProfile(
		context.Background(), account.UserID,
	)
	if err != nil || len(profile.RoleAssignments) != 1 || profile.RoleAssignments[0].RoleCatalogVersion != "" ||
		profile.RoleAssignments[0].RoleDefinitionDigest != "" {
		t.Fatalf("legacy SQLite role metadata changed during upgrade: profile=%#v err=%v", profile, err)
	}
	assertSQLiteLocalIdentityDirectoryQueryPlan(t, upgraded)
	if err := upgraded.Close(); err != nil {
		t.Fatalf("close upgraded SQLite runtime: %v", err)
	}
	reapplied := openSQLiteLocalIdentityAdministrationRuntime(t, databasePath, allMigrations)
	if err := reapplied.VerifyMigrations(context.Background(), allMigrations); err != nil {
		t.Fatalf("verify reapplied SQLite administration migration: %v", err)
	}
	if err := reapplied.Close(); err != nil {
		t.Fatalf("close reapplied SQLite runtime: %v", err)
	}
}

func runDurableLocalIdentityAdministrationContract(
	t *testing.T,
	repository durableLocalIdentityAdministrationRepository,
) durableLocalIdentityAdministrationState {
	t.Helper()
	ctx := context.Background()
	now := localIdentityTestNow.Add(48 * time.Hour)
	tenantRef := "tenant_demo"
	workspaceID := "workspace_demo"
	administratorUserID := "usr_0000000000000101"
	createDurableLocalIdentityAdministrationAccount(t, repository, administratorUserID, 0x101, now)

	bootstrapServices := make([]*localIdentityAdministrationService, 2)
	for index := range bootstrapServices {
		service := newLocalIdentityAdministrationService(repository)
		service.now = func() time.Time { return now }
		sequence := index + 1
		service.newID = func(prefix string) (string, error) {
			if prefix == "mbr_" {
				return fmt.Sprintf("mbr_%016x", 0x100+sequence), nil
			}
			return fmt.Sprintf("rla_%016x", 0x100+sequence), nil
		}
		bootstrapServices[index] = service
	}
	start := make(chan struct{})
	results := make(chan error, len(bootstrapServices))
	bootstraps := make(chan LocalIdentityWorkspaceAdministratorBootstrap, len(bootstrapServices))
	var wait sync.WaitGroup
	for index, service := range bootstrapServices {
		wait.Add(1)
		go func(index int, service *localIdentityAdministrationService) {
			defer wait.Done()
			<-start
			created, err := service.BootstrapWorkspaceAdministrator(ctx, LocalIdentityBootstrapWorkspaceAdministratorInput{
				TenantRef: tenantRef, WorkspaceID: workspaceID, UserID: administratorUserID,
				AuditRef: fmt.Sprintf("audit:durable-bootstrap-%d", index),
			})
			results <- err
			if err == nil {
				bootstraps <- created
			}
		}(index, service)
	}
	close(start)
	wait.Wait()
	close(results)
	close(bootstraps)
	winners := 0
	denied := 0
	for result := range results {
		switch {
		case result == nil:
			winners++
		case errors.Is(result, errLocalIdentityAdminBootstrapDenied):
			denied++
		default:
			t.Fatalf("unexpected durable bootstrap result: %v", result)
		}
	}
	if winners != 1 || denied != 1 {
		t.Fatalf("durable bootstrap single-winner mismatch: winners=%d denied=%d", winners, denied)
	}
	bootstrap := <-bootstraps
	actor := LocalIdentityAdministrationActor{
		UserID: administratorUserID, TenantRef: tenantRef, WorkspaceID: workspaceID, AuthenticatedAt: now,
	}

	for index := 0; index < 120; index++ {
		sequence := 0x200 + index
		userID := fmt.Sprintf("usr_%016x", sequence)
		createDurableLocalIdentityAdministrationAccount(t, repository, userID, sequence, now)
		if err := repository.CreateWorkspaceMembership(ctx, WorkspaceMembership{
			SchemaVersion: localIdentitySchemaVersion, MembershipID: fmt.Sprintf("mbr_%016x", sequence), UserID: userID,
			TenantRef: tenantRef, WorkspaceID: workspaceID, LifecycleState: localIdentityStateActive,
			RecordVersion: 1, CreatedAt: now, UpdatedAt: now, AuditRef: "audit:durable-member-seed",
		}); err != nil {
			t.Fatalf("seed durable member %d: %v", index, err)
		}
	}
	service := newLocalIdentityAdministrationService(repository)
	service.now = func() time.Time { return now }
	var idMu sync.Mutex
	nextID := 0x1000
	service.newID = func(prefix string) (string, error) {
		idMu.Lock()
		defer idMu.Unlock()
		nextID++
		return fmt.Sprintf("%s%016x", prefix, nextID), nil
	}
	memberIDs := make([]string, 0, 121)
	cursor := ""
	for pageIndex := 0; ; pageIndex++ {
		page, err := service.ListWorkspaceMembers(ctx, actor, LocalIdentityWorkspaceMemberListQuery{
			TenantRef: tenantRef, WorkspaceID: workspaceID, Limit: 50, Cursor: cursor,
		})
		if err != nil {
			t.Fatalf("list durable member page %d: %v", pageIndex, err)
		}
		for _, member := range page.Members {
			memberIDs = append(memberIDs, member.MembershipID)
		}
		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}
	if len(memberIDs) != 121 || !slices.IsSortedFunc(memberIDs, func(left, right string) int {
		return strings.Compare(right, left)
	}) {
		t.Fatalf("durable member pagination mismatch: count=%d", len(memberIDs))
	}

	targetUserID := "usr_0000000000000200"
	reader, _ := builtInLocalIdentityRole(localIdentityRoleWorkspaceReader)
	assignment, err := service.AssignWorkspaceRole(ctx, actor, LocalIdentityAssignWorkspaceRoleInput{
		TenantRef: tenantRef, WorkspaceID: workspaceID, UserID: targetUserID, RoleKey: reader.RoleKey,
		ExpectedCatalogVersion: reader.CatalogVersion, ExpectedRoleDefinitionDigest: reader.DefinitionDigest,
		AuditRef: "audit:durable-reader",
	})
	if err != nil {
		t.Fatalf("assign durable catalog role: %v", err)
	}
	detail, err := service.ReadWorkspaceMember(ctx, actor, tenantRef, workspaceID, targetUserID)
	if err != nil || len(detail.RoleAssignments) != 1 || detail.RoleAssignments[0].CatalogDrift ||
		detail.RoleAssignments[0].RoleCatalogVersion != reader.CatalogVersion {
		t.Fatalf("durable catalog metadata projection mismatch: detail=%#v err=%v", detail, err)
	}
	if _, err := repository.RevokeRoleAssignment(
		ctx, assignment.AssignmentID, assignment.RecordVersion, now.Add(time.Minute), "audit:primitive-canonical-revoke",
	); !errors.Is(err, errLocalIdentityVersionConflict) {
		t.Fatalf("primitive role revocation accepted catalog-owned assignment: %v", err)
	}
	if _, err := repository.RevokeWorkspaceMembership(
		ctx, "mbr_0000000000000200", 1, now.Add(time.Minute), "audit:primitive-membership-revoke",
	); !errors.Is(err, errLocalIdentityVersionConflict) {
		t.Fatalf("primitive membership revocation accepted active workspace assignments: %v", err)
	}

	revokeResults := make(chan error, 2)
	revokeStart := make(chan struct{})
	for range 2 {
		go func() {
			<-revokeStart
			_, revokeErr := service.RevokeWorkspaceRole(ctx, actor, LocalIdentityRevokeWorkspaceRoleInput{
				TenantRef: tenantRef, WorkspaceID: workspaceID, AssignmentID: assignment.AssignmentID,
				ExpectedVersion: assignment.RecordVersion, Confirmed: true, AuditRef: "audit:durable-reader-cas",
			})
			revokeResults <- revokeErr
		}()
	}
	close(revokeStart)
	revokeWinners := 0
	revokeConflicts := 0
	for range 2 {
		switch revokeErr := <-revokeResults; {
		case revokeErr == nil:
			revokeWinners++
		case errors.Is(revokeErr, errLocalIdentityRoleAssignmentConflict):
			revokeConflicts++
		default:
			t.Fatalf("unexpected durable role CAS result: %v", revokeErr)
		}
	}
	if revokeWinners != 1 || revokeConflicts != 1 {
		t.Fatalf("durable role CAS single-winner mismatch: winners=%d conflicts=%d", revokeWinners, revokeConflicts)
	}
	assignment, err = service.AssignWorkspaceRole(ctx, actor, LocalIdentityAssignWorkspaceRoleInput{
		TenantRef: tenantRef, WorkspaceID: workspaceID, UserID: targetUserID, RoleKey: reader.RoleKey,
		ExpectedCatalogVersion: reader.CatalogVersion, ExpectedRoleDefinitionDigest: reader.DefinitionDigest,
		AuditRef: "audit:durable-reader-reassign",
	})
	if err != nil {
		t.Fatalf("reassign durable catalog role after CAS: %v", err)
	}

	revocation, err := service.RevokeWorkspaceMembership(ctx, actor, LocalIdentityRevokeWorkspaceMembershipInput{
		TenantRef: tenantRef, WorkspaceID: workspaceID, MembershipID: "mbr_0000000000000200",
		ExpectedVersion: 1, Confirmed: true, AuditRef: "audit:durable-member-revoke",
	})
	if err != nil || len(revocation.RevokedRoleAssignments) != 1 ||
		revocation.RevokedRoleAssignments[0].AssignmentID != assignment.AssignmentID {
		t.Fatalf("durable aggregate revoke mismatch: revocation=%#v err=%v", revocation, err)
	}
	if _, err := repository.AuthorizeWorkspace(
		ctx, targetUserID, tenantRef, workspaceID, []string{"applications:read"}, now.Add(time.Minute),
	); !errors.Is(err, errLocalIdentityMembershipDenied) {
		t.Fatalf("durable membership revoke did not fail closed immediately: %v", err)
	}
	recreated, err := service.CreateWorkspaceMembership(ctx, actor, LocalIdentityCreateWorkspaceMembershipInput{
		TenantRef: tenantRef, WorkspaceID: workspaceID, UserID: targetUserID, AuditRef: "audit:durable-member-recreate",
	})
	if err != nil {
		t.Fatalf("recreate durable membership: %v", err)
	}
	if _, err := repository.AuthorizeWorkspace(
		ctx, targetUserID, tenantRef, workspaceID, []string{"applications:read"}, now.Add(time.Minute),
	); !errors.Is(err, errLocalIdentityPermissionDenied) {
		t.Fatalf("durable membership recreation restored old grants: %v", err)
	}
	if _, err := service.RevokeWorkspaceRole(ctx, actor, LocalIdentityRevokeWorkspaceRoleInput{
		TenantRef: tenantRef, WorkspaceID: workspaceID, AssignmentID: bootstrap.RoleAssignment.AssignmentID,
		ExpectedVersion: bootstrap.RoleAssignment.RecordVersion, Confirmed: true, AuditRef: "audit:durable-last-admin",
	}); !errors.Is(err, errLocalIdentityLastAdminRemoval) {
		t.Fatalf("durable last administrator role was revoked: %v", err)
	}
	if _, err := service.RevokeWorkspaceMembership(ctx, actor, LocalIdentityRevokeWorkspaceMembershipInput{
		TenantRef: tenantRef, WorkspaceID: workspaceID, MembershipID: bootstrap.Membership.MembershipID,
		ExpectedVersion: bootstrap.Membership.RecordVersion, Confirmed: true, AuditRef: "audit:durable-self-membership",
	}); !errors.Is(err, errLocalIdentitySelfMembershipRevoke) {
		t.Fatalf("durable self membership was revoked: %v", err)
	}
	return durableLocalIdentityAdministrationState{
		Now: now, TenantRef: tenantRef, WorkspaceID: workspaceID, Administrator: actor,
		TargetUserID: targetUserID, MembershipID: recreated.MembershipID, CatalogVersion: reader.CatalogVersion,
	}
}

func createDurableLocalIdentityAdministrationAccount(
	t *testing.T,
	repository localIdentityRepository,
	userID string,
	sequence int,
	now time.Time,
) {
	t.Helper()
	account, credential := localIdentityTestAccount(
		userID, fmt.Sprintf("cred_%016x", sequence), fmt.Sprintf("durable-%x@example.com", sequence), now,
	)
	if err := repository.CreateAccount(context.Background(), account, credential); err != nil {
		t.Fatalf("create durable administration account %s: %v", userID, err)
	}
}

func openSQLiteLocalIdentityAdministrationRuntime(
	t *testing.T,
	databasePath string,
	migrations []sqlitedev.Migration,
) *sqlitedev.Runtime {
	t.Helper()
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   migrations,
	})
	if err != nil {
		t.Fatalf("open SQLite administration runtime: %v", err)
	}
	return runtime
}

func assertSQLiteLocalIdentityDirectoryQueryPlan(t *testing.T, runtime *sqlitedev.Runtime) {
	t.Helper()
	rows, err := runtime.DB().QueryContext(context.Background(), `EXPLAIN QUERY PLAN
        SELECT membership_id FROM local_workspace_memberships
        WHERE tenant_ref=? AND workspace_id=? AND lifecycle_state=?
        ORDER BY updated_at_unix_nano DESC, membership_id DESC LIMIT 51`,
		"tenant_demo", "workspace_demo", localIdentityStateActive,
	)
	if err != nil {
		t.Fatalf("explain SQLite member directory query: %v", err)
	}
	defer rows.Close()
	plan := strings.Builder{}
	for rows.Next() {
		var id, parent, unused int
		var detail string
		if err := rows.Scan(&id, &parent, &unused, &detail); err != nil {
			t.Fatalf("scan SQLite member directory query plan: %v", err)
		}
		plan.WriteString(detail)
	}
	if rows.Err() != nil || !strings.Contains(plan.String(), "local_workspace_memberships_directory_idx") {
		t.Fatalf("SQLite member directory query did not use its ordered index: %s", plan.String())
	}
}
