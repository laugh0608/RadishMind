package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLocalIdentityBuiltInRoleCatalogContract(t *testing.T) {
	catalog := LocalIdentityBuiltInRoleCatalog()
	if catalog.DefinitionDigest != "sha256:d784ef5d5595f4fa3ed96f32c86f3fd12edbd4098da46668366f97ce42e2d4d0" {
		t.Fatalf("role catalog changed without an explicit catalog version decision: %s", catalog.DefinitionDigest)
	}
	if catalog.SchemaVersion != localIdentityRoleCatalogSchemaVersion ||
		catalog.CatalogVersion != localIdentityRoleCatalogVersion ||
		len(catalog.Roles) != 4 || !strings.HasPrefix(catalog.DefinitionDigest, "sha256:") {
		t.Fatalf("role catalog contract drifted: %#v", catalog)
	}
	byKey := make(map[string]LocalIdentityRoleDefinition, len(catalog.Roles))
	for _, definition := range catalog.Roles {
		if definition.CatalogVersion != catalog.CatalogVersion ||
			!localRoleKeyPattern.MatchString(definition.RoleKey) ||
			definition.DisplayName == "" || definition.Summary == "" ||
			!slices.IsSorted(definition.PermissionGrants) ||
			len(definition.PermissionGrants) == 0 || !strings.HasPrefix(definition.DefinitionDigest, "sha256:") {
			t.Fatalf("invalid role definition: %#v", definition)
		}
		for _, grant := range definition.PermissionGrants {
			if _, allowed := workspacePermissionAllowlist[grant]; !allowed {
				t.Fatalf("role %s contains undeclared grant %q", definition.RoleKey, grant)
			}
		}
		if _, duplicate := byKey[definition.RoleKey]; duplicate {
			t.Fatalf("duplicate role key: %s", definition.RoleKey)
		}
		byKey[definition.RoleKey] = definition
	}
	reader := byKey[localIdentityRoleWorkspaceReader]
	builder := byKey[localIdentityRoleWorkspaceBuilder]
	reviewer := byKey[localIdentityRoleWorkspaceReviewer]
	administrator := byKey[localIdentityRoleWorkspaceAdmin]
	if !localIdentityGrantSubset(reader.PermissionGrants, builder.PermissionGrants) ||
		!localIdentityGrantSubset(builder.PermissionGrants, reviewer.PermissionGrants) ||
		!localIdentityGrantSubset(reviewer.PermissionGrants, administrator.PermissionGrants) {
		t.Fatal("built-in role grants are not cumulative")
	}
	allowed := make([]string, 0, len(workspacePermissionAllowlist))
	for permission := range workspacePermissionAllowlist {
		allowed = append(allowed, permission)
	}
	slices.Sort(allowed)
	if !slices.Equal(administrator.PermissionGrants, allowed) {
		t.Fatalf("workspace_admin must deliberately cover the complete allowlist:\nwant=%v\n got=%v", allowed, administrator.PermissionGrants)
	}
	for _, definition := range catalog.Roles {
		wantManagement := definition.RoleKey == localIdentityRoleWorkspaceAdmin
		if definition.CanManageLocalIdentity != wantManagement {
			t.Fatalf("identity management capability drifted for %s", definition.RoleKey)
		}
		for _, permission := range localIdentityManagementPermissions {
			if slices.Contains(definition.PermissionGrants, permission) != wantManagement {
				t.Fatalf("management permission %s leaked into role %s", permission, definition.RoleKey)
			}
		}
	}
	catalog.Roles[0].PermissionGrants[0] = "applications:archive"
	if LocalIdentityBuiltInRoleCatalog().Roles[0].PermissionGrants[0] == "applications:archive" {
		t.Fatal("role catalog caller mutated canonical grants")
	}
}

func TestMemoryLocalIdentityAdministrationPaginationCursorAndSanitizedDetail(t *testing.T) {
	fixture := newLocalIdentityAdministrationTestFixture(t)
	ctx := context.Background()
	for index := 0; index < 120; index++ {
		userID := fmt.Sprintf("usr_%016x", index+1000)
		createLocalIdentityAdministrationTestAccount(t, fixture.repository, userID, index+1000, fixture.now)
		membership := WorkspaceMembership{
			SchemaVersion:  localIdentitySchemaVersion,
			MembershipID:   fmt.Sprintf("mbr_%016x", index+1000),
			UserID:         userID,
			TenantRef:      fixture.actor.TenantRef,
			WorkspaceID:    fixture.actor.WorkspaceID,
			LifecycleState: localIdentityStateActive,
			RecordVersion:  1,
			CreatedAt:      fixture.now,
			UpdatedAt:      fixture.now,
			AuditRef:       "audit:member-seed",
		}
		if err := fixture.repository.CreateWorkspaceMembership(ctx, membership); err != nil {
			t.Fatalf("seed workspace member %d: %v", index, err)
		}
	}

	memberIDs := make([]string, 0, 121)
	cursor := ""
	firstCursor := ""
	for pageIndex := 0; ; pageIndex++ {
		page, err := fixture.service.ListWorkspaceMembers(ctx, fixture.actor, LocalIdentityWorkspaceMemberListQuery{
			TenantRef:   fixture.actor.TenantRef,
			WorkspaceID: fixture.actor.WorkspaceID,
			Limit:       50,
			Cursor:      cursor,
		})
		if err != nil {
			t.Fatalf("list member page %d: %v", pageIndex, err)
		}
		if len(page.Members) == 0 || len(page.Members) > 50 {
			t.Fatalf("unexpected member page size: %d", len(page.Members))
		}
		for _, member := range page.Members {
			if !member.MembershipEffective {
				t.Fatalf("active unexpired membership was not projected as effective: %#v", member)
			}
			memberIDs = append(memberIDs, member.MembershipID)
		}
		if pageIndex == 0 {
			firstCursor = page.NextCursor
		}
		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
		if pageIndex > 3 {
			t.Fatal("member pagination did not terminate")
		}
	}
	if len(memberIDs) != 121 {
		t.Fatalf("member pagination count mismatch: %d", len(memberIDs))
	}
	if !slices.IsSortedFunc(memberIDs, func(left, right string) int { return strings.Compare(right, left) }) {
		t.Fatalf("same-timestamp members are not ordered by membership_id DESC: %v", memberIDs)
	}
	seen := make(map[string]struct{}, len(memberIDs))
	for _, membershipID := range memberIDs {
		if _, duplicate := seen[membershipID]; duplicate {
			t.Fatalf("member pagination repeated %s", membershipID)
		}
		seen[membershipID] = struct{}{}
	}
	if firstCursor == "" {
		t.Fatal("first member page did not return a cursor")
	}
	for name, query := range map[string]LocalIdentityWorkspaceMemberListQuery{
		"limit drift": {
			TenantRef:   fixture.actor.TenantRef,
			WorkspaceID: fixture.actor.WorkspaceID,
			Limit:       25,
			Cursor:      firstCursor,
		},
		"state drift": {
			TenantRef:       fixture.actor.TenantRef,
			WorkspaceID:     fixture.actor.WorkspaceID,
			MembershipState: localIdentityStateRevoked,
			Limit:           50,
			Cursor:          firstCursor,
		},
		"tampered": {
			TenantRef:   fixture.actor.TenantRef,
			WorkspaceID: fixture.actor.WorkspaceID,
			Limit:       50,
			Cursor:      firstCursor[:len(firstCursor)-1] + "A",
		},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := fixture.service.ListWorkspaceMembers(ctx, fixture.actor, query); !errors.Is(err, errLocalIdentityMemberCursorInvalid) {
				t.Fatalf("cursor binding drift was accepted: %v", err)
			}
		})
	}

	detail, err := fixture.service.ReadWorkspaceMember(
		ctx,
		fixture.actor,
		fixture.actor.TenantRef,
		fixture.actor.WorkspaceID,
		fixture.actor.UserID,
	)
	if err != nil || len(detail.Memberships) != 1 || len(detail.RoleAssignments) != 1 ||
		!detail.CanManageLocalIdentity || detail.RoleAssignments[0].CatalogDrift {
		t.Fatalf("administrator detail mismatch: detail=%#v err=%v", detail, err)
	}
	payload, err := json.Marshal(detail)
	if err != nil {
		t.Fatalf("marshal member detail: %v", err)
	}
	for _, forbidden := range []string{
		`"login_identifier":`, `"credential":`, `"session":`, `"issuer":`, `"subject":`, `"audit_ref":`, "audit:bootstrap",
	} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("member detail leaked %q: %s", forbidden, payload)
		}
	}
}

func TestMemoryLocalIdentityAdministrationMutationSafety(t *testing.T) {
	fixture := newLocalIdentityAdministrationTestFixture(t)
	ctx := context.Background()
	targetUserID := "usr_00000000000000aa"
	createLocalIdentityAdministrationTestAccount(t, fixture.repository, targetUserID, 0xaa, fixture.now)

	staleActor := fixture.actor
	staleActor.AuthenticatedAt = fixture.now.Add(-localIdentityAdministrationRecentAuthenticationAge - time.Second)
	if _, err := fixture.service.CreateWorkspaceMembership(ctx, staleActor, LocalIdentityCreateWorkspaceMembershipInput{
		TenantRef:   fixture.actor.TenantRef,
		WorkspaceID: fixture.actor.WorkspaceID,
		UserID:      targetUserID,
		AuditRef:    "audit:stale-create",
	}); !errors.Is(err, errLocalIdentityRecentAuthentication) {
		t.Fatalf("stale actor created membership: %v", err)
	}
	scopeActor := fixture.actor
	if _, err := fixture.service.CreateWorkspaceMembership(ctx, scopeActor, LocalIdentityCreateWorkspaceMembershipInput{
		TenantRef:   fixture.actor.TenantRef,
		WorkspaceID: "workspace_other",
		UserID:      targetUserID,
		AuditRef:    "audit:scope-create",
	}); !errors.Is(err, errLocalIdentityAdminScopeMismatch) {
		t.Fatalf("scope mismatch created membership: %v", err)
	}

	membership, err := fixture.service.CreateWorkspaceMembership(ctx, fixture.actor, LocalIdentityCreateWorkspaceMembershipInput{
		TenantRef:   fixture.actor.TenantRef,
		WorkspaceID: fixture.actor.WorkspaceID,
		UserID:      targetUserID,
		AuditRef:    "audit:member-create",
	})
	if err != nil {
		t.Fatalf("create member: %v", err)
	}
	reader, _ := builtInLocalIdentityRole(localIdentityRoleWorkspaceReader)
	assignment, err := fixture.service.AssignWorkspaceRole(ctx, fixture.actor, LocalIdentityAssignWorkspaceRoleInput{
		TenantRef:                    fixture.actor.TenantRef,
		WorkspaceID:                  fixture.actor.WorkspaceID,
		UserID:                       targetUserID,
		RoleKey:                      reader.RoleKey,
		ExpectedCatalogVersion:       reader.CatalogVersion,
		ExpectedRoleDefinitionDigest: reader.DefinitionDigest,
		AuditRef:                     "audit:reader-assign",
	})
	if err != nil || !slices.Equal(assignment.PermissionGrants, reader.PermissionGrants) {
		t.Fatalf("assign catalog reader: assignment=%#v err=%v", assignment, err)
	}
	if _, err := fixture.service.AssignWorkspaceRole(ctx, fixture.actor, LocalIdentityAssignWorkspaceRoleInput{
		TenantRef:                    fixture.actor.TenantRef,
		WorkspaceID:                  fixture.actor.WorkspaceID,
		UserID:                       targetUserID,
		RoleKey:                      reader.RoleKey,
		ExpectedCatalogVersion:       reader.CatalogVersion,
		ExpectedRoleDefinitionDigest: "sha256:" + strings.Repeat("0", 64),
		AuditRef:                     "audit:stale-catalog",
	}); !errors.Is(err, errLocalIdentityRoleCatalogMismatch) {
		t.Fatalf("stale role catalog assignment was accepted: %v", err)
	}
	if _, err := fixture.service.AssignWorkspaceRole(ctx, fixture.actor, LocalIdentityAssignWorkspaceRoleInput{
		TenantRef:                    fixture.actor.TenantRef,
		WorkspaceID:                  fixture.actor.WorkspaceID,
		UserID:                       targetUserID,
		RoleKey:                      reader.RoleKey,
		ExpectedCatalogVersion:       reader.CatalogVersion,
		ExpectedRoleDefinitionDigest: reader.DefinitionDigest,
		AuditRef:                     "audit:duplicate-role",
	}); !errors.Is(err, errLocalIdentityRoleAssignmentConflict) {
		t.Fatalf("duplicate role assignment was accepted: %v", err)
	}
	if _, err := fixture.repository.AuthorizeWorkspace(
		ctx,
		targetUserID,
		fixture.actor.TenantRef,
		fixture.actor.WorkspaceID,
		[]string{"applications:read"},
		fixture.now,
	); err != nil {
		t.Fatalf("catalog role did not authorize business permission: %v", err)
	}

	revocation, err := fixture.service.RevokeWorkspaceMembership(ctx, fixture.actor, LocalIdentityRevokeWorkspaceMembershipInput{
		TenantRef:       fixture.actor.TenantRef,
		WorkspaceID:     fixture.actor.WorkspaceID,
		MembershipID:    membership.MembershipID,
		ExpectedVersion: membership.RecordVersion,
		Confirmed:       true,
		AuditRef:        "audit:member-revoke",
	})
	if err != nil || revocation.Membership.LifecycleState != localIdentityStateRevoked ||
		len(revocation.RevokedRoleAssignments) != 1 || revocation.RevokedRoleAssignments[0].AssignmentID != assignment.AssignmentID {
		t.Fatalf("aggregate membership revoke mismatch: revocation=%#v err=%v", revocation, err)
	}
	if _, err := fixture.repository.AuthorizeWorkspace(
		ctx,
		targetUserID,
		fixture.actor.TenantRef,
		fixture.actor.WorkspaceID,
		[]string{"applications:read"},
		fixture.now,
	); !errors.Is(err, errLocalIdentityMembershipDenied) {
		t.Fatalf("membership revoke did not fail closed immediately: %v", err)
	}
	newMembership, err := fixture.service.CreateWorkspaceMembership(ctx, fixture.actor, LocalIdentityCreateWorkspaceMembershipInput{
		TenantRef:   fixture.actor.TenantRef,
		WorkspaceID: fixture.actor.WorkspaceID,
		UserID:      targetUserID,
		AuditRef:    "audit:member-recreate",
	})
	if err != nil || newMembership.MembershipID == membership.MembershipID {
		t.Fatalf("recreate member after aggregate revoke: membership=%#v err=%v", newMembership, err)
	}
	if _, err := fixture.repository.AuthorizeWorkspace(
		ctx,
		targetUserID,
		fixture.actor.TenantRef,
		fixture.actor.WorkspaceID,
		[]string{"applications:read"},
		fixture.now,
	); !errors.Is(err, errLocalIdentityPermissionDenied) {
		t.Fatalf("revoked grants silently returned after membership recreation: %v", err)
	}
	legacyAssignment := LocalRoleAssignment{
		SchemaVersion:    localIdentitySchemaVersion,
		AssignmentID:     "rla_00000000000000ee",
		UserID:           targetUserID,
		TenantRef:        fixture.actor.TenantRef,
		WorkspaceID:      fixture.actor.WorkspaceID,
		RoleKey:          localIdentityRoleWorkspaceReader,
		PermissionGrants: []string{"applications:read"},
		LifecycleState:   localIdentityStateActive,
		RecordVersion:    1,
		CreatedAt:        fixture.now,
		UpdatedAt:        fixture.now,
		AuditRef:         "audit:legacy-role",
	}
	if err := fixture.repository.CreateRoleAssignment(ctx, legacyAssignment); err != nil {
		t.Fatalf("create legacy role assignment: %v", err)
	}
	detail, err := fixture.service.ReadWorkspaceMember(
		ctx,
		fixture.actor,
		fixture.actor.TenantRef,
		fixture.actor.WorkspaceID,
		targetUserID,
	)
	if err != nil || len(detail.RoleAssignments) != 2 || !detail.RoleAssignments[1].CatalogDrift {
		t.Fatalf("legacy role catalog drift was not exposed: detail=%#v err=%v", detail, err)
	}
	if _, err := fixture.service.RevokeWorkspaceMembership(ctx, fixture.actor, LocalIdentityRevokeWorkspaceMembershipInput{
		TenantRef:       fixture.actor.TenantRef,
		WorkspaceID:     fixture.actor.WorkspaceID,
		MembershipID:    fixture.bootstrap.Membership.MembershipID,
		ExpectedVersion: fixture.bootstrap.Membership.RecordVersion,
		Confirmed:       true,
		AuditRef:        "audit:self-revoke",
	}); !errors.Is(err, errLocalIdentitySelfMembershipRevoke) {
		t.Fatalf("administrator revoked own current membership: %v", err)
	}
}

func TestMemoryLocalIdentityAdministrationLastAdminAndConcurrentCAS(t *testing.T) {
	fixture := newLocalIdentityAdministrationTestFixture(t)
	ctx := context.Background()
	if _, err := fixture.service.RevokeWorkspaceRole(ctx, fixture.actor, LocalIdentityRevokeWorkspaceRoleInput{
		TenantRef:       fixture.actor.TenantRef,
		WorkspaceID:     fixture.actor.WorkspaceID,
		AssignmentID:    fixture.bootstrap.RoleAssignment.AssignmentID,
		ExpectedVersion: fixture.bootstrap.RoleAssignment.RecordVersion,
		Confirmed:       true,
		AuditRef:        "audit:last-admin-role-revoke",
	}); !errors.Is(err, errLocalIdentityLastAdminRemoval) {
		t.Fatalf("last administrator role was revoked: %v", err)
	}
	fixture.repository.mu.RLock()
	adminAssignment := fixture.repository.roleAssignments[fixture.bootstrap.RoleAssignment.AssignmentID]
	fixture.repository.mu.RUnlock()
	if adminAssignment.LifecycleState != localIdentityStateActive || adminAssignment.RecordVersion != 1 {
		t.Fatalf("last-administrator denial changed the role assignment: %#v", adminAssignment)
	}

	targetUserID := "usr_00000000000000bb"
	createLocalIdentityAdministrationTestAccount(t, fixture.repository, targetUserID, 0xbb, fixture.now)
	membership, err := fixture.service.CreateWorkspaceMembership(ctx, fixture.actor, LocalIdentityCreateWorkspaceMembershipInput{
		TenantRef:   fixture.actor.TenantRef,
		WorkspaceID: fixture.actor.WorkspaceID,
		UserID:      targetUserID,
		AuditRef:    "audit:concurrent-member",
	})
	if err != nil {
		t.Fatalf("create concurrent target member: %v", err)
	}
	reader, _ := builtInLocalIdentityRole(localIdentityRoleWorkspaceReader)
	assignment, err := fixture.service.AssignWorkspaceRole(ctx, fixture.actor, LocalIdentityAssignWorkspaceRoleInput{
		TenantRef:                    fixture.actor.TenantRef,
		WorkspaceID:                  fixture.actor.WorkspaceID,
		UserID:                       targetUserID,
		RoleKey:                      reader.RoleKey,
		ExpectedCatalogVersion:       reader.CatalogVersion,
		ExpectedRoleDefinitionDigest: reader.DefinitionDigest,
		AuditRef:                     "audit:concurrent-reader",
	})
	if err != nil {
		t.Fatalf("assign concurrent target role: %v", err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	var wait sync.WaitGroup
	for index := 0; index < 2; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			<-start
			_, revokeErr := fixture.repository.RevokeWorkspaceMembershipAndAssignments(
				ctx,
				fixture.actor.TenantRef,
				fixture.actor.WorkspaceID,
				membership.MembershipID,
				membership.RecordVersion,
				fixture.actor.UserID,
				fixture.now.Add(time.Duration(index+1)*time.Second),
				fmt.Sprintf("audit:concurrent-revoke-%d", index),
			)
			results <- revokeErr
		}(index)
	}
	close(start)
	wait.Wait()
	close(results)
	winners := 0
	conflicts := 0
	for result := range results {
		switch {
		case result == nil:
			winners++
		case errors.Is(result, errLocalIdentityMembershipConflict):
			conflicts++
		default:
			t.Fatalf("unexpected concurrent membership result: %v", result)
		}
	}
	if winners != 1 || conflicts != 1 {
		t.Fatalf("membership CAS single-winner mismatch: winners=%d conflicts=%d", winners, conflicts)
	}
	fixture.repository.mu.RLock()
	storedAssignment := fixture.repository.roleAssignments[assignment.AssignmentID]
	fixture.repository.mu.RUnlock()
	if storedAssignment.LifecycleState != localIdentityStateRevoked || storedAssignment.RecordVersion != 2 {
		t.Fatalf("aggregate role revocation was not exactly once: %#v", storedAssignment)
	}
}

func TestLocalIdentityAdministrationFailureCodes(t *testing.T) {
	tests := []struct {
		err  error
		code string
	}{
		{errLocalIdentityAdminUnavailable, LocalIdentityFailureAdminUnavailable},
		{errLocalIdentityAdminScopeMismatch, LocalIdentityFailureAdminScopeMismatch},
		{errLocalIdentityMemberUnavailable, LocalIdentityFailureMemberUnavailable},
		{errLocalIdentityMemberCursorInvalid, LocalIdentityFailureMemberCursorInvalid},
		{errLocalIdentityRoleCatalogMismatch, LocalIdentityFailureRoleCatalogMismatch},
		{errLocalIdentityMembershipConflict, LocalIdentityFailureMembershipConflict},
		{errLocalIdentityRoleAssignmentConflict, LocalIdentityFailureRoleAssignmentConflict},
		{errLocalIdentitySelfMembershipRevoke, LocalIdentityFailureSelfMembershipRevoke},
		{errLocalIdentityLastAdminRemoval, LocalIdentityFailureLastAdminRemoval},
		{errLocalIdentityRecentAuthentication, LocalIdentityFailureRecentAuthentication},
		{errLocalIdentityAdminBootstrapDenied, LocalIdentityFailureAdminBootstrapDenied},
	}
	for _, test := range tests {
		if got := localIdentityRepositoryError(test.err); got != test.code {
			t.Errorf("failure code mismatch for %v: got %q, want %q", test.err, got, test.code)
		}
	}
}

func TestMemoryLocalIdentityAdministrationMutationRechecksActorAuthorization(t *testing.T) {
	fixture := newLocalIdentityAdministrationTestFixture(t)
	ctx := context.Background()
	secondAdminUserID := "usr_00000000000000d1"
	createLocalIdentityAdministrationTestAccount(t, fixture.repository, secondAdminUserID, 0xd1, fixture.now)
	if _, err := fixture.service.CreateWorkspaceMembership(ctx, fixture.actor, LocalIdentityCreateWorkspaceMembershipInput{
		TenantRef:   fixture.actor.TenantRef,
		WorkspaceID: fixture.actor.WorkspaceID,
		UserID:      secondAdminUserID,
		AuditRef:    "audit:second-admin-member",
	}); err != nil {
		t.Fatalf("create second administrator membership: %v", err)
	}
	administrator, _ := builtInLocalIdentityRole(localIdentityRoleWorkspaceAdmin)
	if _, err := fixture.service.AssignWorkspaceRole(ctx, fixture.actor, LocalIdentityAssignWorkspaceRoleInput{
		TenantRef:                    fixture.actor.TenantRef,
		WorkspaceID:                  fixture.actor.WorkspaceID,
		UserID:                       secondAdminUserID,
		RoleKey:                      administrator.RoleKey,
		ExpectedCatalogVersion:       administrator.CatalogVersion,
		ExpectedRoleDefinitionDigest: administrator.DefinitionDigest,
		AuditRef:                     "audit:second-admin-role",
	}); err != nil {
		t.Fatalf("assign second administrator: %v", err)
	}
	if _, err := fixture.service.RevokeWorkspaceRole(ctx, fixture.actor, LocalIdentityRevokeWorkspaceRoleInput{
		TenantRef:       fixture.actor.TenantRef,
		WorkspaceID:     fixture.actor.WorkspaceID,
		AssignmentID:    fixture.bootstrap.RoleAssignment.AssignmentID,
		ExpectedVersion: fixture.bootstrap.RoleAssignment.RecordVersion,
		Confirmed:       true,
		AuditRef:        "audit:first-admin-role-revoke",
	}); err != nil {
		t.Fatalf("revoke first of two administrator roles: %v", err)
	}

	targetUserID := "usr_00000000000000d2"
	createLocalIdentityAdministrationTestAccount(t, fixture.repository, targetUserID, 0xd2, fixture.now)
	targetMembership := WorkspaceMembership{
		SchemaVersion:  localIdentitySchemaVersion,
		MembershipID:   "mbr_00000000000000d2",
		UserID:         targetUserID,
		TenantRef:      fixture.actor.TenantRef,
		WorkspaceID:    fixture.actor.WorkspaceID,
		LifecycleState: localIdentityStateActive,
		RecordVersion:  1,
		CreatedAt:      fixture.now,
		UpdatedAt:      fixture.now,
		AuditRef:       "audit:trailing-write",
	}
	if err := fixture.repository.CreateWorkspaceMembershipForAdministration(
		ctx,
		fixture.actor.UserID,
		targetMembership,
		fixture.now,
	); !errors.Is(err, errLocalIdentityPermissionDenied) {
		t.Fatalf("repository accepted a trailing mutation after actor role revoke: %v", err)
	}
	fixture.repository.mu.RLock()
	_, created := fixture.repository.memberships[targetMembership.MembershipID]
	fixture.repository.mu.RUnlock()
	if created {
		t.Fatal("denied trailing mutation produced a membership side effect")
	}

	forged := fixture.bootstrap.RoleAssignment
	forged.AssignmentID = "rla_00000000000000ff"
	forged.UserID = targetUserID
	forged.RoleCatalogVersion = ""
	forged.RoleDefinitionDigest = ""
	if err := fixture.repository.CreateRoleAssignment(ctx, forged); !errors.Is(err, errLocalIdentityContractMismatch) {
		t.Fatalf("legacy arbitrary-grant path accepted identity management permissions: %v", err)
	}
}

func TestMemoryLocalIdentityAdministrationRoleRevocationCASSingleWinner(t *testing.T) {
	fixture := newLocalIdentityAdministrationTestFixture(t)
	ctx := context.Background()
	targetUserID := "usr_00000000000000e1"
	createLocalIdentityAdministrationTestAccount(t, fixture.repository, targetUserID, 0xe1, fixture.now)
	if _, err := fixture.service.CreateWorkspaceMembership(ctx, fixture.actor, LocalIdentityCreateWorkspaceMembershipInput{
		TenantRef:   fixture.actor.TenantRef,
		WorkspaceID: fixture.actor.WorkspaceID,
		UserID:      targetUserID,
		AuditRef:    "audit:role-cas-member",
	}); err != nil {
		t.Fatalf("create role CAS member: %v", err)
	}
	reader, _ := builtInLocalIdentityRole(localIdentityRoleWorkspaceReader)
	assignment, err := fixture.service.AssignWorkspaceRole(ctx, fixture.actor, LocalIdentityAssignWorkspaceRoleInput{
		TenantRef:                    fixture.actor.TenantRef,
		WorkspaceID:                  fixture.actor.WorkspaceID,
		UserID:                       targetUserID,
		RoleKey:                      reader.RoleKey,
		ExpectedCatalogVersion:       reader.CatalogVersion,
		ExpectedRoleDefinitionDigest: reader.DefinitionDigest,
		AuditRef:                     "audit:role-cas-assign",
	})
	if err != nil {
		t.Fatalf("assign role for CAS: %v", err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	var wait sync.WaitGroup
	for index := 0; index < 2; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			<-start
			_, revokeErr := fixture.repository.RevokeCatalogRoleAssignment(
				ctx,
				fixture.actor.UserID,
				fixture.actor.TenantRef,
				fixture.actor.WorkspaceID,
				assignment.AssignmentID,
				assignment.RecordVersion,
				fixture.now.Add(time.Duration(index+1)*time.Second),
				fmt.Sprintf("audit:role-cas-%d", index),
			)
			results <- revokeErr
		}(index)
	}
	close(start)
	wait.Wait()
	close(results)
	winners := 0
	conflicts := 0
	for result := range results {
		switch {
		case result == nil:
			winners++
		case errors.Is(result, errLocalIdentityRoleAssignmentConflict):
			conflicts++
		default:
			t.Fatalf("unexpected role CAS result: %v", result)
		}
	}
	if winners != 1 || conflicts != 1 {
		t.Fatalf("role CAS single-winner mismatch: winners=%d conflicts=%d", winners, conflicts)
	}
}

func TestMemoryLocalIdentityAdministrationBootstrapSingleWinner(t *testing.T) {
	repository := newMemoryLocalIdentityRepository()
	now := localIdentityTestNow.Add(24 * time.Hour)
	userIDs := []string{"usr_00000000000000c1", "usr_00000000000000c2"}
	for index, userID := range userIDs {
		createLocalIdentityAdministrationTestAccount(t, repository, userID, 0xc1+index, now)
	}
	services := []*localIdentityAdministrationService{
		newLocalIdentityAdministrationService(repository),
		newLocalIdentityAdministrationService(repository),
	}
	for index, service := range services {
		service.now = func() time.Time { return now }
		counter := index + 1
		service.newID = func(prefix string) (string, error) {
			return fmt.Sprintf("%s%016x", prefix, counter), nil
		}
	}
	start := make(chan struct{})
	results := make(chan error, len(services))
	var wait sync.WaitGroup
	for index, service := range services {
		wait.Add(1)
		go func(index int, candidate *localIdentityAdministrationService) {
			defer wait.Done()
			<-start
			_, err := candidate.BootstrapWorkspaceAdministrator(context.Background(), LocalIdentityBootstrapWorkspaceAdministratorInput{
				TenantRef:   "tenant_demo",
				WorkspaceID: "workspace_demo",
				UserID:      userIDs[index],
				AuditRef:    fmt.Sprintf("audit:bootstrap-%d", index),
			})
			results <- err
		}(index, service)
	}
	close(start)
	wait.Wait()
	close(results)
	winners := 0
	denied := 0
	for result := range results {
		switch {
		case result == nil:
			winners++
		case errors.Is(result, errLocalIdentityAdminBootstrapDenied):
			denied++
		default:
			t.Fatalf("unexpected bootstrap result: %v", result)
		}
	}
	repository.mu.RLock()
	membershipCount := len(repository.memberships)
	assignmentCount := len(repository.roleAssignments)
	repository.mu.RUnlock()
	if winners != 1 || denied != 1 || membershipCount != 1 || assignmentCount != 1 {
		t.Fatalf(
			"bootstrap single-winner mismatch: winners=%d denied=%d memberships=%d assignments=%d",
			winners,
			denied,
			membershipCount,
			assignmentCount,
		)
	}
	missingService := newLocalIdentityAdministrationService(newMemoryLocalIdentityRepository())
	missingService.now = func() time.Time { return now }
	missingService.newID = services[0].newID
	if _, err := missingService.BootstrapWorkspaceAdministrator(context.Background(), LocalIdentityBootstrapWorkspaceAdministratorInput{
		TenantRef:   "tenant_demo",
		WorkspaceID: "workspace_demo",
		UserID:      "usr_00000000000000ff",
		AuditRef:    "audit:missing-bootstrap",
	}); !errors.Is(err, errLocalIdentityMemberUnavailable) {
		t.Fatalf("missing bootstrap account did not fail closed: %v", err)
	}
}

type localIdentityAdministrationTestFixture struct {
	repository *memoryLocalIdentityRepository
	service    *localIdentityAdministrationService
	now        time.Time
	actor      LocalIdentityAdministrationActor
	bootstrap  LocalIdentityWorkspaceAdministratorBootstrap
}

func newLocalIdentityAdministrationTestFixture(t *testing.T) localIdentityAdministrationTestFixture {
	t.Helper()
	repository := newMemoryLocalIdentityRepository()
	now := localIdentityTestNow.Add(24 * time.Hour)
	userID := "usr_0000000000000001"
	createLocalIdentityAdministrationTestAccount(t, repository, userID, 1, now)
	service := newLocalIdentityAdministrationService(repository)
	service.now = func() time.Time { return now }
	var idMu sync.Mutex
	nextID := 1
	service.newID = func(prefix string) (string, error) {
		idMu.Lock()
		defer idMu.Unlock()
		identifier := fmt.Sprintf("%s%016x", prefix, nextID)
		nextID++
		return identifier, nil
	}
	bootstrap, err := service.BootstrapWorkspaceAdministrator(context.Background(), LocalIdentityBootstrapWorkspaceAdministratorInput{
		TenantRef:   "tenant_demo",
		WorkspaceID: "workspace_demo",
		UserID:      userID,
		AuditRef:    "audit:bootstrap",
	})
	if err != nil {
		t.Fatalf("bootstrap administration fixture: %v", err)
	}
	return localIdentityAdministrationTestFixture{
		repository: repository,
		service:    service,
		now:        now,
		actor: LocalIdentityAdministrationActor{
			UserID:          userID,
			TenantRef:       "tenant_demo",
			WorkspaceID:     "workspace_demo",
			AuthenticatedAt: now,
		},
		bootstrap: bootstrap,
	}
}

func createLocalIdentityAdministrationTestAccount(
	t *testing.T,
	repository *memoryLocalIdentityRepository,
	userID string,
	sequence int,
	now time.Time,
) {
	t.Helper()
	credentialID := fmt.Sprintf("cred_%016x", sequence)
	account, credential := localIdentityTestAccount(
		userID,
		credentialID,
		fmt.Sprintf("member-%x@example.com", sequence),
		now,
	)
	if err := repository.CreateAccount(context.Background(), account, credential); err != nil {
		t.Fatalf("create administration test account %s: %v", userID, err)
	}
}

func localIdentityGrantSubset(subset []string, superset []string) bool {
	for _, grant := range subset {
		if !slices.Contains(superset, grant) {
			return false
		}
	}
	return true
}
