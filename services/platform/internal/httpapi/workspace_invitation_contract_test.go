package httpapi

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestWorkspaceInvitationCodePolicyAndCanonicalProjection(t *testing.T) {
	invitationID := "wsi_0123456789abcdef0123456789abcdef"
	code, digest, err := newWorkspaceInvitationCode(invitationID)
	if err != nil {
		t.Fatalf("generate invitation code: %v", err)
	}
	parsed, err := parseWorkspaceInvitationCode(code)
	if err != nil {
		t.Fatalf("parse generated invitation code: %v", err)
	}
	if parsed.InvitationID != invitationID ||
		!workspaceInvitationSecretMatches(digestWorkspaceInvitationSecret(parsed.secret), digest) ||
		!validWorkspaceInvitationSecretDigest(digest) {
		t.Fatal("generated invitation code did not round-trip through the secret policy")
	}
	otherCode, otherDigest, err := newWorkspaceInvitationCode(invitationID)
	if err != nil {
		t.Fatalf("generate second invitation code: %v", err)
	}
	if code == otherCode || workspaceInvitationSecretMatches(digest, otherDigest) {
		t.Fatal("independent invitation secrets were reused")
	}
	for _, invalid := range []string{
		"", "rmi_short.value", strings.ToUpper(code), " " + code, code + "=", strings.Replace(code, ".", "..", 1),
	} {
		if _, err := parseWorkspaceInvitationCode(invalid); !errors.Is(err, errWorkspaceInvitationInvalid) {
			t.Fatalf("malformed invitation code was accepted: %v", err)
		}
	}

	now := time.Date(2026, 9, 2, 8, 0, 0, 0, time.UTC)
	definition, _ := builtInLocalIdentityRole(localIdentityRoleWorkspaceReader)
	invitation := WorkspaceInvitation{
		SchemaVersion: workspaceInvitationSchemaVersion, InvitationID: invitationID, RecordVersion: 1,
		TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", RoleKey: definition.RoleKey,
		RoleCatalogVersion: definition.CatalogVersion, RoleDefinitionDigest: definition.DefinitionDigest,
		TTLPolicy: workspaceInvitationTTL1Hour, LifecycleState: workspaceInvitationLifecyclePending,
		ExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now,
		CreatedByActorRef: "user:usr_0000000000000001",
		CreatedRequestRef: "request:create", CreatedAuditRef: "audit:create",
		UpdatedRequestRef: "request:create", UpdatedAuditRef: "audit:create",
	}
	if !validWorkspaceInvitation(invitation) {
		t.Fatal("valid canonical invitation was rejected")
	}
	projected := projectWorkspaceInvitation(invitation, now)
	if projected.EffectiveState != workspaceInvitationEffectivePending {
		t.Fatalf("pending invitation projected as %q", projected.EffectiveState)
	}
	payload, err := json.Marshal(projected)
	if err != nil {
		t.Fatalf("marshal canonical invitation: %v", err)
	}
	if strings.Contains(string(payload), code) || strings.Contains(string(payload), "invitation_code") ||
		strings.Contains(string(payload), "secret_digest") || strings.Contains(string(payload), "permission_grants") {
		t.Fatal("canonical invitation projection exposed secret or authorization material")
	}
}

func TestWorkspaceInvitationTTLRoleAndFailureContracts(t *testing.T) {
	wantDurations := map[string]time.Duration{
		workspaceInvitationTTL1Hour: time.Hour, workspaceInvitationTTL24Hours: 24 * time.Hour,
		workspaceInvitationTTL72Hours: 72 * time.Hour, workspaceInvitationTTL7Days: 7 * 24 * time.Hour,
	}
	for policy, want := range wantDurations {
		got, ok := workspaceInvitationTTLDuration(policy)
		if !ok || got != want {
			t.Fatalf("TTL policy %s drifted: got=%s ok=%t", policy, got, ok)
		}
	}
	if _, ok := workspaceInvitationTTLDuration("30d"); ok {
		t.Fatal("arbitrary invitation TTL was accepted")
	}
	for _, roleKey := range []string{
		localIdentityRoleWorkspaceReader,
		localIdentityRoleWorkspaceBuilder,
		localIdentityRoleWorkspaceReviewer,
	} {
		definition, exists := builtInLocalIdentityRole(roleKey)
		if !exists || !workspaceInvitationRoleEligible(definition) {
			t.Fatalf("eligible invitation role was rejected: %s", roleKey)
		}
	}
	administrator, _ := builtInLocalIdentityRole(localIdentityRoleWorkspaceAdmin)
	if workspaceInvitationRoleEligible(administrator) {
		t.Fatal("workspace_admin became invitation-eligible")
	}

	failures := []struct {
		err  error
		code string
	}{
		{errWorkspaceInvitationAdminUnavailable, WorkspaceInvitationFailureAdminUnavailable},
		{errWorkspaceInvitationCursorInvalid, WorkspaceInvitationFailureCursorInvalid},
		{errWorkspaceInvitationRoleIneligible, WorkspaceInvitationFailureRoleIneligible},
		{errWorkspaceInvitationRoleCatalogMismatch, WorkspaceInvitationFailureRoleCatalogMismatch},
		{errWorkspaceInvitationVersionConflict, WorkspaceInvitationFailureVersionConflict},
		{errWorkspaceInvitationTransitionInvalid, WorkspaceInvitationFailureTransitionInvalid},
		{errWorkspaceInvitationInvalid, WorkspaceInvitationFailureInvalid},
		{errWorkspaceInvitationNotClaimable, WorkspaceInvitationFailureNotClaimable},
		{errWorkspaceInvitationAccountIneligible, WorkspaceInvitationFailureAccountIneligible},
		{errWorkspaceInvitationMembershipConflict, WorkspaceInvitationFailureMembershipConflict},
		{errWorkspaceInvitationStoreUnavailable, WorkspaceInvitationFailureStoreUnavailable},
		{errLocalIdentityRecentAuthentication, LocalIdentityFailureRecentAuthentication},
		{errLocalIdentityMembershipDenied, LocalIdentityFailureMembershipDenied},
		{errLocalIdentityPermissionDenied, LocalIdentityFailurePermissionDenied},
	}
	for _, failure := range failures {
		if got := workspaceInvitationFailureCode(failure.err); got != failure.code {
			t.Fatalf("failure mapping drifted: got=%s want=%s", got, failure.code)
		}
	}
}
