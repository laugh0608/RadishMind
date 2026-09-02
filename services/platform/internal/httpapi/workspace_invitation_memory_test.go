package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

type workspaceInvitationTestClock struct {
	mu  sync.RWMutex
	now time.Time
}

func (clock *workspaceInvitationTestClock) read() time.Time {
	clock.mu.RLock()
	defer clock.mu.RUnlock()
	return clock.now
}

func (clock *workspaceInvitationTestClock) set(now time.Time) {
	clock.mu.Lock()
	defer clock.mu.Unlock()
	clock.now = now.UTC()
}

type workspaceInvitationTestFixture struct {
	repository *memoryLocalIdentityRepository
	service    *workspaceInvitationService
	admin      LocalIdentityAdministrationActor
	claimant   WorkspaceInvitationClaimantActor
	clock      *workspaceInvitationTestClock
}

func newWorkspaceInvitationTestFixture(t *testing.T) workspaceInvitationTestFixture {
	t.Helper()
	administration := newLocalIdentityAdministrationTestFixture(t)
	claimantUserID := "usr_00000000000000c1"
	createLocalIdentityAdministrationTestAccount(t, administration.repository, claimantUserID, 193, administration.now)
	clock := &workspaceInvitationTestClock{now: administration.now}
	service := newWorkspaceInvitationService(administration.repository)
	service.now = clock.read
	var idMu sync.Mutex
	nextID := 1000
	service.newID = func(prefix string) (string, error) {
		idMu.Lock()
		defer idMu.Unlock()
		nextID++
		if prefix == "wsi_" {
			return fmt.Sprintf("%s%032x", prefix, nextID), nil
		}
		return fmt.Sprintf("%s%016x", prefix, nextID), nil
	}
	return workspaceInvitationTestFixture{
		repository: administration.repository,
		service:    service,
		admin:      administration.actor,
		claimant: WorkspaceInvitationClaimantActor{
			UserID: claimantUserID, TenantRef: administration.actor.TenantRef, AuthenticatedAt: administration.now,
		},
		clock: clock,
	}
}

func createWorkspaceInvitationForTest(
	t *testing.T,
	fixture workspaceInvitationTestFixture,
	roleKey string,
	ttlPolicy string,
) WorkspaceInvitationCreation {
	t.Helper()
	definition, exists := builtInLocalIdentityRole(roleKey)
	if !exists {
		t.Fatalf("missing role definition: %s", roleKey)
	}
	creation, err := fixture.service.Create(context.Background(), fixture.admin, WorkspaceInvitationCreateInput{
		TenantRef: fixture.admin.TenantRef, WorkspaceID: fixture.admin.WorkspaceID,
		RoleKey: roleKey, ExpectedCatalogVersion: definition.CatalogVersion,
		ExpectedRoleDefinitionDigest: definition.DefinitionDigest,
		TTLPolicy:                    ttlPolicy, Confirmed: true, RequestRef: "request:invitation-create", AuditRef: "audit:invitation-create",
	})
	if err != nil {
		t.Fatalf("create workspace invitation: %v", err)
	}
	return creation
}

func claimWorkspaceInvitationForTest(
	fixture workspaceInvitationTestFixture,
	creation WorkspaceInvitationCreation,
) (WorkspaceInvitationMutation, error) {
	return fixture.service.Claim(context.Background(), fixture.claimant, WorkspaceInvitationClaimInput{
		InvitationCode: creation.InvitationCode, ExpectedVersion: creation.Invitation.RecordVersion,
		Confirmed: true, RequestRef: "request:invitation-claim", AuditRef: "audit:invitation-claim",
	})
}

func TestMemoryWorkspaceInvitationCreatePreviewListAndRevoke(t *testing.T) {
	fixture := newWorkspaceInvitationTestFixture(t)
	for _, ttlPolicy := range []string{
		workspaceInvitationTTL1Hour,
		workspaceInvitationTTL24Hours,
		workspaceInvitationTTL72Hours,
		workspaceInvitationTTL7Days,
	} {
		creation := createWorkspaceInvitationForTest(t, fixture, localIdentityRoleWorkspaceReader, ttlPolicy)
		duration, _ := workspaceInvitationTTLDuration(ttlPolicy)
		if !creation.Invitation.ExpiresAt.Equal(fixture.clock.read().Add(duration)) ||
			creation.Invitation.EffectiveState != workspaceInvitationEffectivePending {
			t.Fatalf("TTL projection drifted for %s", ttlPolicy)
		}
	}

	definition, _ := builtInLocalIdentityRole(localIdentityRoleWorkspaceAdmin)
	_, err := fixture.service.Create(context.Background(), fixture.admin, WorkspaceInvitationCreateInput{
		TenantRef: fixture.admin.TenantRef, WorkspaceID: fixture.admin.WorkspaceID,
		RoleKey: definition.RoleKey, ExpectedCatalogVersion: definition.CatalogVersion,
		ExpectedRoleDefinitionDigest: definition.DefinitionDigest,
		TTLPolicy:                    workspaceInvitationTTL1Hour, Confirmed: true,
		RequestRef: "request:admin-role", AuditRef: "audit:admin-role",
	})
	if !errors.Is(err, errWorkspaceInvitationRoleIneligible) {
		t.Fatalf("workspace_admin invitation did not fail closed: %v", err)
	}
	_, err = fixture.service.Create(context.Background(), fixture.admin, WorkspaceInvitationCreateInput{
		TenantRef: fixture.admin.TenantRef, WorkspaceID: fixture.admin.WorkspaceID,
		RoleKey: "workspace_owner", ExpectedCatalogVersion: "local_identity_builtin_roles_v2",
		ExpectedRoleDefinitionDigest: "sha256:" + strings.Repeat("0", 64),
		TTLPolicy:                    workspaceInvitationTTL1Hour, Confirmed: true,
		RequestRef: "request:unknown-role", AuditRef: "audit:unknown-role",
	})
	if !errors.Is(err, errWorkspaceInvitationRoleIneligible) {
		t.Fatalf("unknown invitation role did not fail closed: %v", err)
	}
	reader, _ := builtInLocalIdentityRole(localIdentityRoleWorkspaceReader)
	_, err = fixture.service.Create(context.Background(), fixture.admin, WorkspaceInvitationCreateInput{
		TenantRef: fixture.admin.TenantRef, WorkspaceID: fixture.admin.WorkspaceID,
		RoleKey: reader.RoleKey, ExpectedCatalogVersion: reader.CatalogVersion,
		ExpectedRoleDefinitionDigest: "sha256:" + strings.Repeat("0", 64),
		TTLPolicy:                    workspaceInvitationTTL1Hour, Confirmed: true,
		RequestRef: "request:catalog-drift", AuditRef: "audit:catalog-drift",
	})
	if !errors.Is(err, errWorkspaceInvitationRoleCatalogMismatch) {
		t.Fatalf("catalog mismatch did not fail closed: %v", err)
	}

	page, err := fixture.service.List(context.Background(), fixture.admin, WorkspaceInvitationListQuery{
		TenantRef: fixture.admin.TenantRef, WorkspaceID: fixture.admin.WorkspaceID, EffectiveState: workspaceInvitationEffectivePending,
	})
	if err != nil || len(page.Invitations) != 4 || page.AsOf != fixture.clock.read() {
		t.Fatalf("list pending invitations: count=%d as_of=%s err=%v", len(page.Invitations), page.AsOf, err)
	}
	creation := createWorkspaceInvitationForTest(t, fixture, localIdentityRoleWorkspaceReviewer, workspaceInvitationTTL1Hour)
	preview, err := fixture.service.Preview(context.Background(), fixture.claimant, creation.InvitationCode)
	if err != nil {
		t.Fatalf("preview invitation: %v", err)
	}
	if preview.InvitationID != creation.Invitation.InvitationID || preview.RecordVersion != 1 ||
		preview.TenantRef != fixture.claimant.TenantRef || preview.WorkspaceID != fixture.admin.WorkspaceID ||
		preview.Role.RoleKey != localIdentityRoleWorkspaceReviewer || preview.EffectiveState != workspaceInvitationEffectivePending {
		t.Fatalf("preview projection drifted: %#v", preview)
	}

	mutation, err := fixture.service.Revoke(context.Background(), fixture.admin, WorkspaceInvitationRevokeInput{
		TenantRef: fixture.admin.TenantRef, WorkspaceID: fixture.admin.WorkspaceID,
		InvitationID: creation.Invitation.InvitationID, ExpectedVersion: creation.Invitation.RecordVersion,
		Confirmed: true, RequestRef: "request:invitation-revoke", AuditRef: "audit:invitation-revoke",
	})
	if err != nil {
		t.Fatalf("revoke invitation: %v", err)
	}
	if mutation.Invitation.EffectiveState != workspaceInvitationEffectiveRevoked ||
		mutation.Invitation.RecordVersion != 2 || mutation.Invitation.RevokedAt == nil || mutation.Membership != nil ||
		mutation.RoleAssignment != nil {
		t.Fatalf("revoke mutation drifted: %#v", mutation)
	}
	if _, err := fixture.service.Preview(context.Background(), fixture.claimant, creation.InvitationCode); !errors.Is(err, errWorkspaceInvitationNotClaimable) {
		t.Fatalf("revoked invitation remained previewable: %v", err)
	}
	revoked, err := fixture.service.List(context.Background(), fixture.admin, WorkspaceInvitationListQuery{
		TenantRef: fixture.admin.TenantRef, WorkspaceID: fixture.admin.WorkspaceID,
		EffectiveState: workspaceInvitationEffectiveRevoked,
	})
	if err != nil || len(revoked.Invitations) != 1 || revoked.Invitations[0].InvitationID != creation.Invitation.InvitationID {
		t.Fatalf("revoked directory projection drifted: %#v err=%v", revoked, err)
	}
}

func TestMemoryWorkspaceInvitationActorEligibilityFailsClosed(t *testing.T) {
	fixture := newWorkspaceInvitationTestFixture(t)
	reader, _ := builtInLocalIdentityRole(localIdentityRoleWorkspaceReader)
	createInput := WorkspaceInvitationCreateInput{
		TenantRef: fixture.admin.TenantRef, WorkspaceID: fixture.admin.WorkspaceID,
		RoleKey: reader.RoleKey, ExpectedCatalogVersion: reader.CatalogVersion,
		ExpectedRoleDefinitionDigest: reader.DefinitionDigest,
		TTLPolicy:                    workspaceInvitationTTL1Hour, Confirmed: true,
		RequestRef: "request:eligibility", AuditRef: "audit:eligibility",
	}
	nonAdministrator := LocalIdentityAdministrationActor{
		UserID: fixture.claimant.UserID, TenantRef: fixture.admin.TenantRef,
		WorkspaceID: fixture.admin.WorkspaceID, AuthenticatedAt: fixture.clock.read(),
	}
	if _, err := fixture.service.Create(context.Background(), nonAdministrator, createInput); !errors.Is(err, errLocalIdentityMembershipDenied) {
		t.Fatalf("non-member administrator actor did not fail closed: %v", err)
	}
	staleAdministrator := fixture.admin
	staleAdministrator.AuthenticatedAt = fixture.clock.read().Add(-workspaceInvitationRecentAuthAge - time.Second)
	if _, err := fixture.service.Create(context.Background(), staleAdministrator, createInput); !errors.Is(err, errLocalIdentityRecentAuthentication) {
		t.Fatalf("stale administrator authentication was accepted: %v", err)
	}

	creation := createWorkspaceInvitationForTest(t, fixture, localIdentityRoleWorkspaceReader, workspaceInvitationTTL1Hour)
	wrongTenant := fixture.claimant
	wrongTenant.TenantRef = "tenant_other"
	if _, err := fixture.service.Preview(context.Background(), wrongTenant, creation.InvitationCode); !errors.Is(err, errWorkspaceInvitationAccountIneligible) {
		t.Fatalf("cross-tenant claimant was accepted: %v", err)
	}
	staleClaimant := fixture.claimant
	staleClaimant.AuthenticatedAt = fixture.clock.read().Add(-workspaceInvitationRecentAuthAge - time.Second)
	if _, err := fixture.service.Preview(context.Background(), staleClaimant, creation.InvitationCode); !errors.Is(err, errLocalIdentityRecentAuthentication) {
		t.Fatalf("stale claimant authentication was accepted: %v", err)
	}
	account, err := fixture.repository.ReadAccount(context.Background(), fixture.claimant.UserID)
	if err != nil {
		t.Fatalf("read claimant account: %v", err)
	}
	if _, err := fixture.repository.DisableAccount(
		context.Background(), account.UserID, account.RecordVersion, fixture.clock.read(), "audit:disable-claimant",
	); err != nil {
		t.Fatalf("disable claimant account: %v", err)
	}
	if _, err := fixture.service.Preview(context.Background(), fixture.claimant, creation.InvitationCode); !errors.Is(err, errWorkspaceInvitationAccountIneligible) {
		t.Fatalf("disabled claimant account was accepted: %v", err)
	}
}

func TestMemoryWorkspaceInvitationInvalidCodeUsesUniformFailure(t *testing.T) {
	fixture := newWorkspaceInvitationTestFixture(t)
	creation := createWorkspaceInvitationForTest(t, fixture, localIdentityRoleWorkspaceReader, workspaceInvitationTTL1Hour)
	wrongCode, _, err := newWorkspaceInvitationCode(creation.Invitation.InvitationID)
	if err != nil {
		t.Fatalf("generate wrong secret: %v", err)
	}
	unknownID := "wsi_ffffffffffffffffffffffffffffffff"
	unknownCode, _, err := newWorkspaceInvitationCode(unknownID)
	if err != nil {
		t.Fatalf("generate unknown locator: %v", err)
	}
	for _, invalidCode := range []string{"malformed", wrongCode, unknownCode} {
		_, previewErr := fixture.service.Preview(context.Background(), fixture.claimant, invalidCode)
		if !errors.Is(previewErr, errWorkspaceInvitationInvalid) || workspaceInvitationFailureCode(previewErr) != WorkspaceInvitationFailureInvalid {
			t.Fatalf("invalid invitation failure diverged: %v", previewErr)
		}
	}
}

func TestMemoryWorkspaceInvitationCursorBindsFilterAndAsOf(t *testing.T) {
	fixture := newWorkspaceInvitationTestFixture(t)
	for index := 0; index < 3; index++ {
		createWorkspaceInvitationForTest(t, fixture, localIdentityRoleWorkspaceReader, workspaceInvitationTTL1Hour)
	}
	first, err := fixture.service.List(context.Background(), fixture.admin, WorkspaceInvitationListQuery{
		TenantRef: fixture.admin.TenantRef, WorkspaceID: fixture.admin.WorkspaceID,
		EffectiveState: workspaceInvitationEffectivePending, Limit: 1,
	})
	if err != nil || len(first.Invitations) != 1 || first.NextCursor == "" {
		t.Fatalf("first invitation page: %#v err=%v", first, err)
	}
	firstAsOf := first.AsOf
	fixture.clock.set(firstAsOf.Add(2 * time.Hour))
	second, err := fixture.service.List(context.Background(), fixture.admin, WorkspaceInvitationListQuery{
		TenantRef: fixture.admin.TenantRef, WorkspaceID: fixture.admin.WorkspaceID,
		EffectiveState: workspaceInvitationEffectivePending, Limit: 1, Cursor: first.NextCursor,
	})
	if err != nil || len(second.Invitations) != 1 || second.AsOf != firstAsOf {
		t.Fatalf("cursor did not preserve as_of: %#v err=%v", second, err)
	}
	freshPending, err := fixture.service.List(context.Background(), fixture.admin, WorkspaceInvitationListQuery{
		TenantRef: fixture.admin.TenantRef, WorkspaceID: fixture.admin.WorkspaceID,
		EffectiveState: workspaceInvitationEffectivePending,
	})
	if err != nil || len(freshPending.Invitations) != 0 {
		t.Fatalf("fresh pending view retained expired invitations: %#v err=%v", freshPending, err)
	}
	expired, err := fixture.service.List(context.Background(), fixture.admin, WorkspaceInvitationListQuery{
		TenantRef: fixture.admin.TenantRef, WorkspaceID: fixture.admin.WorkspaceID,
		EffectiveState: workspaceInvitationEffectiveExpired,
	})
	if err != nil || len(expired.Invitations) != 3 {
		t.Fatalf("expired view mismatch: count=%d err=%v", len(expired.Invitations), err)
	}

	tampered := first.NextCursor[:len(first.NextCursor)-1] + "A"
	if tampered == first.NextCursor {
		tampered = first.NextCursor[:len(first.NextCursor)-1] + "B"
	}
	for _, query := range []WorkspaceInvitationListQuery{
		{TenantRef: fixture.admin.TenantRef, WorkspaceID: fixture.admin.WorkspaceID, EffectiveState: workspaceInvitationEffectivePending, Limit: 1, Cursor: tampered},
		{TenantRef: fixture.admin.TenantRef, WorkspaceID: fixture.admin.WorkspaceID, EffectiveState: workspaceInvitationEffectiveExpired, Limit: 1, Cursor: first.NextCursor},
		{TenantRef: fixture.admin.TenantRef, WorkspaceID: fixture.admin.WorkspaceID, EffectiveState: workspaceInvitationEffectivePending, Limit: 2, Cursor: first.NextCursor},
	} {
		if _, err := fixture.service.List(context.Background(), fixture.admin, query); !errors.Is(err, errWorkspaceInvitationCursorInvalid) {
			t.Fatalf("cursor tamper or filter mismatch was accepted: %v", err)
		}
	}
}

func TestMemoryWorkspaceInvitationClaimCreatesExistingAuthorizationOwners(t *testing.T) {
	fixture := newWorkspaceInvitationTestFixture(t)
	creation := createWorkspaceInvitationForTest(t, fixture, localIdentityRoleWorkspaceBuilder, workspaceInvitationTTL24Hours)
	mutation, err := claimWorkspaceInvitationForTest(fixture, creation)
	if err != nil {
		t.Fatalf("claim invitation: %v", err)
	}
	if mutation.Invitation.EffectiveState != workspaceInvitationEffectiveClaimed ||
		mutation.Invitation.ClaimedByUserID != fixture.claimant.UserID || mutation.Membership == nil ||
		mutation.RoleAssignment == nil || mutation.RoleAssignment.RoleKey != localIdentityRoleWorkspaceBuilder {
		t.Fatalf("claim mutation drifted: %#v", mutation)
	}
	authorization, err := fixture.repository.AuthorizeWorkspace(
		context.Background(), fixture.claimant.UserID, fixture.claimant.TenantRef, fixture.admin.WorkspaceID,
		[]string{"applications:write", "workflow_runs:execute"}, fixture.clock.read(),
	)
	if err != nil || authorization.Membership.MembershipID != mutation.Membership.MembershipID ||
		len(authorization.RoleAssignments) != 1 || authorization.RoleAssignments[0].AssignmentID != mutation.RoleAssignment.AssignmentID {
		t.Fatalf("claimed authorization owner mismatch: %#v err=%v", authorization, err)
	}
	if _, err := claimWorkspaceInvitationForTest(fixture, creation); !errors.Is(err, errWorkspaceInvitationNotClaimable) {
		t.Fatalf("claimed invitation replay did not fail closed: %v", err)
	}
	claimed, err := fixture.service.List(context.Background(), fixture.admin, WorkspaceInvitationListQuery{
		TenantRef: fixture.admin.TenantRef, WorkspaceID: fixture.admin.WorkspaceID,
		EffectiveState: workspaceInvitationEffectiveClaimed,
	})
	if err != nil || len(claimed.Invitations) != 1 ||
		claimed.Invitations[0].ClaimedByUserID != fixture.claimant.UserID ||
		claimed.Invitations[0].MembershipID != mutation.Membership.MembershipID ||
		claimed.Invitations[0].AssignmentID != mutation.RoleAssignment.AssignmentID {
		t.Fatalf("claimed directory projection mismatch: %#v err=%v", claimed, err)
	}

	payload, err := json.Marshal(mutation)
	if err != nil {
		t.Fatalf("marshal claim mutation: %v", err)
	}
	fixture.repository.mu.RLock()
	stored := fixture.repository.workspaceInvitations[creation.Invitation.InvitationID]
	fixture.repository.mu.RUnlock()
	digestText := fmt.Sprintf("%x", stored.secretDigest)
	if strings.Contains(string(payload), creation.InvitationCode) || strings.Contains(string(payload), digestText) ||
		strings.Contains(string(payload), "secret_digest") || strings.Contains(string(payload), "permission_grants") {
		t.Fatal("claim mutation exposed secret digest, invitation code, or complete grants")
	}
}

func TestMemoryWorkspaceInvitationClaimSingleWinnerAndAtomicCommit(t *testing.T) {
	fixture := newWorkspaceInvitationTestFixture(t)
	creation := createWorkspaceInvitationForTest(t, fixture, localIdentityRoleWorkspaceReviewer, workspaceInvitationTTL24Hours)
	const contenders = 24
	start := make(chan struct{})
	results := make(chan error, contenders)
	var wait sync.WaitGroup
	for index := 0; index < contenders; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			_, err := claimWorkspaceInvitationForTest(fixture, creation)
			results <- err
		}()
	}
	close(start)
	wait.Wait()
	close(results)
	winners := 0
	losers := 0
	for err := range results {
		switch {
		case err == nil:
			winners++
		case errors.Is(err, errWorkspaceInvitationNotClaimable):
			losers++
		default:
			t.Fatalf("unexpected concurrent claim result: %v", err)
		}
	}
	fixture.repository.mu.RLock()
	defer fixture.repository.mu.RUnlock()
	claimantMemberships := 0
	claimantAssignments := 0
	for _, membership := range fixture.repository.memberships {
		if membership.UserID == fixture.claimant.UserID && membership.TenantRef == fixture.claimant.TenantRef &&
			membership.WorkspaceID == fixture.admin.WorkspaceID {
			claimantMemberships++
		}
	}
	for _, assignment := range fixture.repository.roleAssignments {
		if assignment.UserID == fixture.claimant.UserID && assignment.TenantRef == fixture.claimant.TenantRef &&
			assignment.WorkspaceID == fixture.admin.WorkspaceID {
			claimantAssignments++
		}
	}
	stored := fixture.repository.workspaceInvitations[creation.Invitation.InvitationID].invitation
	if winners != 1 || losers != contenders-1 || claimantMemberships != 1 || claimantAssignments != 1 ||
		stored.LifecycleState != workspaceInvitationLifecycleClaimed ||
		fixture.repository.memberships[stored.MembershipID].MembershipID == "" ||
		fixture.repository.roleAssignments[stored.AssignmentID].AssignmentID == "" {
		t.Fatalf(
			"single-winner atomicity mismatch: winners=%d losers=%d memberships=%d assignments=%d invitation=%#v",
			winners, losers, claimantMemberships, claimantAssignments, stored,
		)
	}
}

func TestMemoryWorkspaceInvitationClaimFailureLeavesNoPartialWrites(t *testing.T) {
	t.Run("expired invitation", func(t *testing.T) {
		fixture := newWorkspaceInvitationTestFixture(t)
		creation := createWorkspaceInvitationForTest(t, fixture, localIdentityRoleWorkspaceReader, workspaceInvitationTTL1Hour)
		fixture.clock.set(fixture.clock.read().Add(time.Hour))
		fixture.claimant.AuthenticatedAt = fixture.clock.read()
		beforeMemberships, beforeAssignments := workspaceInvitationOwnerCounts(fixture.repository)
		if _, err := fixture.service.Preview(context.Background(), fixture.claimant, creation.InvitationCode); !errors.Is(err, errWorkspaceInvitationNotClaimable) {
			t.Fatalf("expired invitation remained previewable: %v", err)
		}
		if _, err := claimWorkspaceInvitationForTest(fixture, creation); !errors.Is(err, errWorkspaceInvitationNotClaimable) {
			t.Fatalf("expired invitation remained claimable: %v", err)
		}
		assertWorkspaceInvitationNoPartialClaim(t, fixture, creation.Invitation.InvitationID, beforeMemberships, beforeAssignments)
	})

	t.Run("catalog drift", func(t *testing.T) {
		fixture := newWorkspaceInvitationTestFixture(t)
		creation := createWorkspaceInvitationForTest(t, fixture, localIdentityRoleWorkspaceReader, workspaceInvitationTTL24Hours)
		fixture.repository.mu.Lock()
		stored := fixture.repository.workspaceInvitations[creation.Invitation.InvitationID]
		stored.invitation.RoleCatalogVersion = "local_identity_builtin_roles_v1"
		stored.invitation.RoleDefinitionDigest = "sha256:" + strings.Repeat("0", 64)
		fixture.repository.workspaceInvitations[creation.Invitation.InvitationID] = stored
		fixture.repository.mu.Unlock()
		beforeMemberships, beforeAssignments := workspaceInvitationOwnerCounts(fixture.repository)
		if _, err := claimWorkspaceInvitationForTest(fixture, creation); !errors.Is(err, errWorkspaceInvitationNotClaimable) {
			t.Fatalf("catalog drift did not fail closed: %v", err)
		}
		assertWorkspaceInvitationNoPartialClaim(t, fixture, creation.Invitation.InvitationID, beforeMemberships, beforeAssignments)
	})

	t.Run("repository corruption", func(t *testing.T) {
		fixture := newWorkspaceInvitationTestFixture(t)
		creation := createWorkspaceInvitationForTest(t, fixture, localIdentityRoleWorkspaceReader, workspaceInvitationTTL24Hours)
		fixture.repository.mu.Lock()
		stored := fixture.repository.workspaceInvitations[creation.Invitation.InvitationID]
		stored.invitation.UpdatedAt = time.Time{}
		fixture.repository.workspaceInvitations[creation.Invitation.InvitationID] = stored
		fixture.repository.mu.Unlock()
		beforeMemberships, beforeAssignments := workspaceInvitationOwnerCounts(fixture.repository)
		if _, err := claimWorkspaceInvitationForTest(fixture, creation); !errors.Is(err, errWorkspaceInvitationStoreUnavailable) {
			t.Fatalf("repository corruption did not fail closed: %v", err)
		}
		assertWorkspaceInvitationNoPartialClaim(t, fixture, creation.Invitation.InvitationID, beforeMemberships, beforeAssignments)
	})

	t.Run("unrevoked expired membership", func(t *testing.T) {
		fixture := newWorkspaceInvitationTestFixture(t)
		creation := createWorkspaceInvitationForTest(t, fixture, localIdentityRoleWorkspaceReader, workspaceInvitationTTL24Hours)
		createdAt := fixture.clock.read().Add(-2 * time.Hour)
		expiresAt := fixture.clock.read().Add(-time.Hour)
		membership := WorkspaceMembership{
			SchemaVersion: localIdentitySchemaVersion, MembershipID: "mbr_0000000000000e01",
			UserID: fixture.claimant.UserID, TenantRef: fixture.claimant.TenantRef, WorkspaceID: fixture.admin.WorkspaceID,
			LifecycleState: localIdentityStateActive, RecordVersion: 1, CreatedAt: createdAt, UpdatedAt: createdAt,
			ExpiresAt: timePointer(expiresAt), AuditRef: "audit:expired-membership",
		}
		if err := fixture.repository.CreateWorkspaceMembership(context.Background(), membership); err != nil {
			t.Fatalf("seed expired active membership: %v", err)
		}
		beforeMemberships, beforeAssignments := workspaceInvitationOwnerCounts(fixture.repository)
		if _, err := claimWorkspaceInvitationForTest(fixture, creation); !errors.Is(err, errWorkspaceInvitationMembershipConflict) {
			t.Fatalf("unrevoked expired membership did not conflict: %v", err)
		}
		assertWorkspaceInvitationNoPartialClaim(t, fixture, creation.Invitation.InvitationID, beforeMemberships, beforeAssignments)
	})

	t.Run("active role assignment invariant", func(t *testing.T) {
		fixture := newWorkspaceInvitationTestFixture(t)
		creation := createWorkspaceInvitationForTest(t, fixture, localIdentityRoleWorkspaceReader, workspaceInvitationTTL24Hours)
		definition, _ := builtInLocalIdentityRole(localIdentityRoleWorkspaceBuilder)
		assignment := LocalRoleAssignment{
			SchemaVersion: localIdentitySchemaVersion, AssignmentID: "rla_0000000000000e02",
			UserID: fixture.claimant.UserID, TenantRef: fixture.claimant.TenantRef, WorkspaceID: fixture.admin.WorkspaceID,
			RoleKey: definition.RoleKey, RoleCatalogVersion: definition.CatalogVersion,
			RoleDefinitionDigest: definition.DefinitionDigest, PermissionGrants: append([]string(nil), definition.PermissionGrants...),
			LifecycleState: localIdentityStateActive, RecordVersion: 1, CreatedAt: fixture.clock.read(), UpdatedAt: fixture.clock.read(),
			AuditRef: "audit:orphan-assignment",
		}
		fixture.repository.mu.Lock()
		fixture.repository.roleAssignments[assignment.AssignmentID] = assignment
		fixture.repository.activeRoleByScope[localRoleScopeKey(
			assignment.UserID, assignment.TenantRef, assignment.WorkspaceID, assignment.RoleKey,
		)] = assignment.AssignmentID
		fixture.repository.mu.Unlock()
		beforeMemberships, beforeAssignments := workspaceInvitationOwnerCounts(fixture.repository)
		if _, err := claimWorkspaceInvitationForTest(fixture, creation); !errors.Is(err, errWorkspaceInvitationMembershipConflict) {
			t.Fatalf("active role assignment invariant did not conflict: %v", err)
		}
		assertWorkspaceInvitationNoPartialClaim(t, fixture, creation.Invitation.InvitationID, beforeMemberships, beforeAssignments)
	})

	t.Run("version conflict", func(t *testing.T) {
		fixture := newWorkspaceInvitationTestFixture(t)
		creation := createWorkspaceInvitationForTest(t, fixture, localIdentityRoleWorkspaceReader, workspaceInvitationTTL24Hours)
		beforeMemberships, beforeAssignments := workspaceInvitationOwnerCounts(fixture.repository)
		_, err := fixture.service.Claim(context.Background(), fixture.claimant, WorkspaceInvitationClaimInput{
			InvitationCode: creation.InvitationCode, ExpectedVersion: creation.Invitation.RecordVersion + 1,
			Confirmed: true, RequestRef: "request:stale-preview", AuditRef: "audit:stale-preview",
		})
		if !errors.Is(err, errWorkspaceInvitationVersionConflict) {
			t.Fatalf("stale preview version did not conflict: %v", err)
		}
		assertWorkspaceInvitationNoPartialClaim(t, fixture, creation.Invitation.InvitationID, beforeMemberships, beforeAssignments)
	})
}

func TestMemoryWorkspaceInvitationRevokedMembershipCanRejoin(t *testing.T) {
	fixture := newWorkspaceInvitationTestFixture(t)
	creation := createWorkspaceInvitationForTest(t, fixture, localIdentityRoleWorkspaceReader, workspaceInvitationTTL24Hours)
	createdAt := fixture.clock.read().Add(-time.Hour)
	membership := WorkspaceMembership{
		SchemaVersion: localIdentitySchemaVersion, MembershipID: "mbr_0000000000000a01",
		UserID: fixture.claimant.UserID, TenantRef: fixture.claimant.TenantRef, WorkspaceID: fixture.admin.WorkspaceID,
		LifecycleState: localIdentityStateActive, RecordVersion: 1, CreatedAt: createdAt, UpdatedAt: createdAt,
		AuditRef: "audit:prior-membership",
	}
	if err := fixture.repository.CreateWorkspaceMembership(context.Background(), membership); err != nil {
		t.Fatalf("seed prior membership: %v", err)
	}
	if _, err := fixture.repository.RevokeWorkspaceMembership(
		context.Background(), membership.MembershipID, membership.RecordVersion, fixture.clock.read().Add(-time.Minute), "audit:prior-revoke",
	); err != nil {
		t.Fatalf("revoke prior membership: %v", err)
	}
	mutation, err := claimWorkspaceInvitationForTest(fixture, creation)
	if err != nil {
		t.Fatalf("claim after explicit membership revocation: %v", err)
	}
	if mutation.Membership == nil || mutation.Membership.MembershipID == membership.MembershipID {
		t.Fatalf("claim did not create a new membership stable ref: %#v", mutation.Membership)
	}
}

func workspaceInvitationOwnerCounts(repository *memoryLocalIdentityRepository) (int, int) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	return len(repository.memberships), len(repository.roleAssignments)
}

func assertWorkspaceInvitationNoPartialClaim(
	t *testing.T,
	fixture workspaceInvitationTestFixture,
	invitationID string,
	wantMemberships int,
	wantAssignments int,
) {
	t.Helper()
	fixture.repository.mu.RLock()
	defer fixture.repository.mu.RUnlock()
	stored := fixture.repository.workspaceInvitations[invitationID].invitation
	if len(fixture.repository.memberships) != wantMemberships || len(fixture.repository.roleAssignments) != wantAssignments ||
		stored.LifecycleState != workspaceInvitationLifecyclePending || stored.RecordVersion != 1 ||
		stored.MembershipID != "" || stored.AssignmentID != "" || stored.ClaimedByUserID != "" {
		t.Fatalf(
			"claim failure left partial writes: memberships=%d/%d assignments=%d/%d invitation=%#v",
			len(fixture.repository.memberships), wantMemberships,
			len(fixture.repository.roleAssignments), wantAssignments,
			stored,
		)
	}
}
