package httpapi

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/sqlitedev"
	sqlitelocalidentitymigrations "radishmind.local/services/platform/migrations/sqlite/local_identity_records"
)

type durableWorkspaceInvitationRepository interface {
	localIdentityRepository
	localIdentityAdministrationRepository
	workspaceInvitationRepository
}

type durableWorkspaceInvitationState struct {
	Now            time.Time
	TenantRef      string
	WorkspaceID    string
	Administrator  LocalIdentityAdministrationActor
	Claimant       WorkspaceInvitationClaimantActor
	Claimed        WorkspaceInvitationMutation
	InvitationCode string
}

func TestSQLiteWorkspaceInvitationDurableContractRestartAndQueryPlan(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "workspace-invitations.db")
	runtime := openSQLiteLocalIdentityAdministrationRuntime(
		t, databasePath, sqlitelocalidentitymigrations.Migrations(),
	)
	repository := newSQLiteLocalIdentityRepository(runtime.DB())
	state := runDurableWorkspaceInvitationContract(t, repository)
	assertSQLiteWorkspaceInvitationQueryPlan(t, runtime)
	if err := runtime.Close(); err != nil {
		t.Fatalf("close SQLite workspace invitation runtime: %v", err)
	}

	restarted := openSQLiteLocalIdentityAdministrationRuntime(
		t, databasePath, sqlitelocalidentitymigrations.Migrations(),
	)
	t.Cleanup(func() { _ = restarted.Close() })
	restored := newSQLiteLocalIdentityRepository(restarted.DB())
	assertDurableWorkspaceInvitationRestart(t, restored, state)
}

func TestSQLiteWorkspaceInvitationMigrationV4UpgradeAndReapply(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "workspace-invitations-upgrade.db")
	migrations := sqlitelocalidentitymigrations.Migrations()
	v4 := openSQLiteLocalIdentityAdministrationRuntime(t, databasePath, migrations[:4])
	account, credential := localIdentityTestAccount(
		"usr_0000000000005a01", "cred_0000000000005a01", "invitation-v4@example.com", localIdentityTestNow,
	)
	if err := newSQLiteLocalIdentityRepository(v4.DB()).CreateAccount(context.Background(), account, credential); err != nil {
		t.Fatalf("create pre-invitation SQLite account: %v", err)
	}
	if err := v4.Close(); err != nil {
		t.Fatalf("close pre-invitation SQLite runtime: %v", err)
	}

	upgraded := openSQLiteLocalIdentityAdministrationRuntime(t, databasePath, migrations)
	if _, err := newSQLiteLocalIdentityRepository(upgraded.DB()).ReadAccount(context.Background(), account.UserID); err != nil {
		t.Fatalf("read account after SQLite invitation upgrade: %v", err)
	}
	if err := upgraded.VerifyMigrations(context.Background(), migrations); err != nil {
		t.Fatalf("verify upgraded SQLite invitation migration: %v", err)
	}
	if err := upgraded.Close(); err != nil {
		t.Fatalf("close upgraded SQLite invitation runtime: %v", err)
	}
	reapplied := openSQLiteLocalIdentityAdministrationRuntime(t, databasePath, migrations)
	if err := reapplied.VerifyMigrations(context.Background(), migrations); err != nil {
		t.Fatalf("verify reapplied SQLite invitation migration: %v", err)
	}
	if err := reapplied.Close(); err != nil {
		t.Fatalf("close reapplied SQLite invitation runtime: %v", err)
	}
}

func TestSQLiteWorkspaceInvitationCorruptPayloadAndUnavailableStoreFailClosed(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "workspace-invitations-failure.db")
	runtime := openSQLiteLocalIdentityAdministrationRuntime(
		t, databasePath, sqlitelocalidentitymigrations.Migrations(),
	)
	repository := newSQLiteLocalIdentityRepository(runtime.DB())
	fixture := newDurableWorkspaceInvitationFixture(t, repository)
	creation := createDurableWorkspaceInvitation(t, fixture, localIdentityRoleWorkspaceReader, workspaceInvitationTTL24Hours)
	if _, err := runtime.DB().ExecContext(context.Background(), `UPDATE local_workspace_invitations
        SET schema_version='corrupt.workspace_invitation' WHERE invitation_id=?`, creation.Invitation.InvitationID); err != nil {
		t.Fatalf("corrupt SQLite invitation payload: %v", err)
	}
	if _, err := fixture.service.Preview(
		context.Background(), fixture.claimant, creation.InvitationCode,
	); !errors.Is(err, errWorkspaceInvitationStoreUnavailable) {
		t.Fatalf("corrupt SQLite invitation did not fail closed: %v", err)
	}
	var membershipCount, assignmentCount int
	if err := runtime.DB().QueryRowContext(context.Background(), `SELECT
        (SELECT count(*) FROM local_workspace_memberships WHERE user_id=? AND tenant_ref=? AND workspace_id=?),
        (SELECT count(*) FROM local_role_assignments WHERE user_id=? AND tenant_ref=? AND workspace_id=?)`,
		fixture.claimant.UserID, fixture.claimant.TenantRef, fixture.admin.WorkspaceID,
		fixture.claimant.UserID, fixture.claimant.TenantRef, fixture.admin.WorkspaceID,
	).Scan(&membershipCount, &assignmentCount); err != nil {
		t.Fatalf("count corrupt SQLite invitation owners: %v", err)
	}
	if membershipCount != 0 || assignmentCount != 0 {
		t.Fatalf("corrupt invitation left partial owners: memberships=%d assignments=%d", membershipCount, assignmentCount)
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("close SQLite invitation failure runtime: %v", err)
	}
	parsed, err := parseWorkspaceInvitationCode(creation.InvitationCode)
	if err != nil {
		t.Fatalf("parse invitation code for unavailable store test: %v", err)
	}
	if _, err := repository.PreviewWorkspaceInvitation(
		context.Background(), fixture.claimant.UserID, fixture.claimant.TenantRef,
		parsed.InvitationID, digestWorkspaceInvitationSecret(parsed.secret), fixture.clock.read(),
	); !errors.Is(err, errWorkspaceInvitationStoreUnavailable) {
		t.Fatalf("closed SQLite owner fell back instead of failing closed: %v", err)
	}
}

type durableWorkspaceInvitationFixture struct {
	repository durableWorkspaceInvitationRepository
	service    *workspaceInvitationService
	admin      LocalIdentityAdministrationActor
	claimant   WorkspaceInvitationClaimantActor
	clock      *workspaceInvitationTestClock
}

func newDurableWorkspaceInvitationFixture(
	t *testing.T,
	repository durableWorkspaceInvitationRepository,
) durableWorkspaceInvitationFixture {
	t.Helper()
	now := localIdentityTestNow.Add(120 * time.Hour)
	tenantRef := "tenant_invitation"
	workspaceID := "workspace_invitation"
	adminUserID := "usr_0000000000005b01"
	claimantUserID := "usr_0000000000005b02"
	createDurableLocalIdentityAdministrationAccount(t, repository, adminUserID, 0x5b01, now)
	createDurableLocalIdentityAdministrationAccount(t, repository, claimantUserID, 0x5b02, now)
	administration := newLocalIdentityAdministrationService(repository)
	administration.now = func() time.Time { return now }
	administration.newID = func(prefix string) (string, error) {
		if prefix == "mbr_" {
			return "mbr_0000000000005b01", nil
		}
		return "rla_0000000000005b01", nil
	}
	if _, err := administration.BootstrapWorkspaceAdministrator(
		context.Background(),
		LocalIdentityBootstrapWorkspaceAdministratorInput{
			TenantRef: tenantRef, WorkspaceID: workspaceID, UserID: adminUserID,
			AuditRef: "audit:invitation-admin-bootstrap",
		},
	); err != nil {
		t.Fatalf("bootstrap durable invitation administrator: %v", err)
	}
	clock := &workspaceInvitationTestClock{now: now}
	service := newWorkspaceInvitationService(repository)
	service.now = clock.read
	var idMu sync.Mutex
	nextID := 0x5c00
	service.newID = func(prefix string) (string, error) {
		idMu.Lock()
		defer idMu.Unlock()
		nextID++
		if prefix == "wsi_" {
			return fmt.Sprintf("%s%032x", prefix, nextID), nil
		}
		return fmt.Sprintf("%s%016x", prefix, nextID), nil
	}
	return durableWorkspaceInvitationFixture{
		repository: repository,
		service:    service,
		admin: LocalIdentityAdministrationActor{
			UserID: adminUserID, TenantRef: tenantRef, WorkspaceID: workspaceID, AuthenticatedAt: now,
		},
		claimant: WorkspaceInvitationClaimantActor{
			UserID: claimantUserID, TenantRef: tenantRef, AuthenticatedAt: now,
		},
		clock: clock,
	}
}

func createDurableWorkspaceInvitation(
	t *testing.T,
	fixture durableWorkspaceInvitationFixture,
	roleKey string,
	ttlPolicy string,
) WorkspaceInvitationCreation {
	t.Helper()
	definition, exists := builtInLocalIdentityRole(roleKey)
	if !exists {
		t.Fatalf("missing durable invitation role %s", roleKey)
	}
	creation, err := fixture.service.Create(context.Background(), fixture.admin, WorkspaceInvitationCreateInput{
		TenantRef: fixture.admin.TenantRef, WorkspaceID: fixture.admin.WorkspaceID,
		RoleKey: roleKey, ExpectedCatalogVersion: definition.CatalogVersion,
		ExpectedRoleDefinitionDigest: definition.DefinitionDigest,
		TTLPolicy:                    ttlPolicy, Confirmed: true,
		RequestRef: "request:durable-invitation-create", AuditRef: "audit:durable-invitation-create",
	})
	if err != nil {
		t.Fatalf("create durable workspace invitation: %v", err)
	}
	return creation
}

func runDurableWorkspaceInvitationContract(
	t *testing.T,
	repository durableWorkspaceInvitationRepository,
) durableWorkspaceInvitationState {
	t.Helper()
	fixture := newDurableWorkspaceInvitationFixture(t, repository)
	for index := 0; index < 3; index++ {
		createDurableWorkspaceInvitation(t, fixture, localIdentityRoleWorkspaceReader, workspaceInvitationTTL1Hour)
	}
	first, err := fixture.service.List(context.Background(), fixture.admin, WorkspaceInvitationListQuery{
		TenantRef: fixture.admin.TenantRef, WorkspaceID: fixture.admin.WorkspaceID,
		EffectiveState: workspaceInvitationEffectivePending, Limit: 1,
	})
	if err != nil || len(first.Invitations) != 1 || first.NextCursor == "" {
		t.Fatalf("list first durable invitation page: page=%#v err=%v", first, err)
	}
	fixture.clock.set(fixture.clock.read().Add(2 * time.Hour))
	second, err := fixture.service.List(context.Background(), fixture.admin, WorkspaceInvitationListQuery{
		TenantRef: fixture.admin.TenantRef, WorkspaceID: fixture.admin.WorkspaceID,
		EffectiveState: workspaceInvitationEffectivePending, Limit: 1, Cursor: first.NextCursor,
	})
	if err != nil || len(second.Invitations) != 1 || !second.AsOf.Equal(first.AsOf) {
		t.Fatalf("durable invitation cursor lost as_of: page=%#v err=%v", second, err)
	}
	expired, err := fixture.service.List(context.Background(), fixture.admin, WorkspaceInvitationListQuery{
		TenantRef: fixture.admin.TenantRef, WorkspaceID: fixture.admin.WorkspaceID,
		EffectiveState: workspaceInvitationEffectiveExpired,
	})
	if err != nil || len(expired.Invitations) != 3 {
		t.Fatalf("list durable expired invitations: count=%d err=%v", len(expired.Invitations), err)
	}
	fixture.admin.AuthenticatedAt = fixture.clock.read()
	fixture.claimant.AuthenticatedAt = fixture.clock.read()

	revocable := createDurableWorkspaceInvitation(
		t, fixture, localIdentityRoleWorkspaceReviewer, workspaceInvitationTTL24Hours,
	)
	if _, err := fixture.service.Revoke(context.Background(), fixture.admin, WorkspaceInvitationRevokeInput{
		TenantRef: fixture.admin.TenantRef, WorkspaceID: fixture.admin.WorkspaceID,
		InvitationID: revocable.Invitation.InvitationID, ExpectedVersion: revocable.Invitation.RecordVersion,
		Confirmed: true, RequestRef: "request:durable-invitation-revoke", AuditRef: "audit:durable-invitation-revoke",
	}); err != nil {
		t.Fatalf("revoke durable invitation: %v", err)
	}
	claimable := createDurableWorkspaceInvitation(
		t, fixture, localIdentityRoleWorkspaceBuilder, workspaceInvitationTTL24Hours,
	)
	preview, err := fixture.service.Preview(context.Background(), fixture.claimant, claimable.InvitationCode)
	if err != nil || preview.InvitationID != claimable.Invitation.InvitationID ||
		preview.Role.RoleKey != localIdentityRoleWorkspaceBuilder {
		t.Fatalf("preview durable invitation: preview=%#v err=%v", preview, err)
	}

	const contenders = 16
	start := make(chan struct{})
	results := make(chan error, contenders)
	winners := make(chan WorkspaceInvitationMutation, contenders)
	var wait sync.WaitGroup
	for index := 0; index < contenders; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			mutation, claimErr := fixture.service.Claim(context.Background(), fixture.claimant, WorkspaceInvitationClaimInput{
				InvitationCode: claimable.InvitationCode, ExpectedVersion: claimable.Invitation.RecordVersion,
				Confirmed: true, RequestRef: "request:durable-invitation-claim", AuditRef: "audit:durable-invitation-claim",
			})
			results <- claimErr
			if claimErr == nil {
				winners <- mutation
			}
		}()
	}
	close(start)
	wait.Wait()
	close(results)
	close(winners)
	winnerCount := 0
	loserCount := 0
	for claimErr := range results {
		switch {
		case claimErr == nil:
			winnerCount++
		case errors.Is(claimErr, errWorkspaceInvitationNotClaimable):
			loserCount++
		default:
			t.Fatalf("unexpected durable concurrent claim result: %v", claimErr)
		}
	}
	if winnerCount != 1 || loserCount != contenders-1 {
		t.Fatalf("durable claim single-winner mismatch: winners=%d losers=%d", winnerCount, loserCount)
	}
	claimed := <-winners
	if claimed.Membership == nil || claimed.RoleAssignment == nil ||
		claimed.Invitation.EffectiveState != workspaceInvitationEffectiveClaimed {
		t.Fatalf("durable claim mutation incomplete: %#v", claimed)
	}
	authorization, err := repository.AuthorizeWorkspace(
		context.Background(), fixture.claimant.UserID, fixture.claimant.TenantRef, fixture.admin.WorkspaceID,
		[]string{"applications:write", "workflow_runs:execute"}, fixture.clock.read(),
	)
	if err != nil || authorization.Membership.MembershipID != claimed.Membership.MembershipID ||
		len(authorization.RoleAssignments) != 1 ||
		authorization.RoleAssignments[0].AssignmentID != claimed.RoleAssignment.AssignmentID {
		t.Fatalf("durable claimed authorization owners mismatch: authorization=%#v err=%v", authorization, err)
	}
	return durableWorkspaceInvitationState{
		Now: fixture.clock.read(), TenantRef: fixture.admin.TenantRef, WorkspaceID: fixture.admin.WorkspaceID,
		Administrator: fixture.admin, Claimant: fixture.claimant, Claimed: claimed,
		InvitationCode: claimable.InvitationCode,
	}
}

func assertDurableWorkspaceInvitationRestart(
	t *testing.T,
	repository durableWorkspaceInvitationRepository,
	state durableWorkspaceInvitationState,
) {
	t.Helper()
	service := newWorkspaceInvitationService(repository)
	service.now = func() time.Time { return state.Now.Add(time.Minute) }
	state.Administrator.AuthenticatedAt = state.Now.Add(time.Minute)
	state.Claimant.AuthenticatedAt = state.Now.Add(time.Minute)
	claimed, err := service.List(context.Background(), state.Administrator, WorkspaceInvitationListQuery{
		TenantRef: state.TenantRef, WorkspaceID: state.WorkspaceID,
		EffectiveState: workspaceInvitationEffectiveClaimed,
	})
	if err != nil || len(claimed.Invitations) != 1 ||
		claimed.Invitations[0].InvitationID != state.Claimed.Invitation.InvitationID ||
		claimed.Invitations[0].MembershipID != state.Claimed.Membership.MembershipID ||
		claimed.Invitations[0].AssignmentID != state.Claimed.RoleAssignment.AssignmentID {
		t.Fatalf("durable invitation restart projection mismatch: page=%#v err=%v", claimed, err)
	}
	if _, err := service.Preview(
		context.Background(), state.Claimant, state.InvitationCode,
	); !errors.Is(err, errWorkspaceInvitationNotClaimable) {
		t.Fatalf("claimed invitation became previewable after restart: %v", err)
	}
	authorization, err := repository.AuthorizeWorkspace(
		context.Background(), state.Claimant.UserID, state.Claimant.TenantRef, state.WorkspaceID,
		[]string{"applications:write"}, state.Now.Add(time.Minute),
	)
	if err != nil || authorization.Membership.MembershipID != state.Claimed.Membership.MembershipID {
		t.Fatalf("claimed authorization owner did not survive restart: authorization=%#v err=%v", authorization, err)
	}
}

func assertSQLiteWorkspaceInvitationQueryPlan(t *testing.T, runtime *sqlitedev.Runtime) {
	t.Helper()
	rows, err := runtime.DB().QueryContext(context.Background(), `EXPLAIN QUERY PLAN
        SELECT invitation_id FROM local_workspace_invitations
        WHERE tenant_ref=? AND workspace_id=? AND lifecycle_state=?
        ORDER BY updated_at_unix_nano DESC, invitation_id DESC LIMIT 51`,
		"tenant_invitation", "workspace_invitation", workspaceInvitationLifecyclePending,
	)
	if err != nil {
		t.Fatalf("explain SQLite invitation directory query: %v", err)
	}
	defer rows.Close()
	plan := strings.Builder{}
	for rows.Next() {
		var id, parent, unused int
		var detail string
		if err := rows.Scan(&id, &parent, &unused, &detail); err != nil {
			t.Fatalf("scan SQLite invitation query plan: %v", err)
		}
		plan.WriteString(detail)
	}
	if rows.Err() != nil || !strings.Contains(plan.String(), "local_workspace_invitations_directory_idx") {
		t.Fatalf("SQLite invitation query did not use ordered index: %s", plan.String())
	}
}
