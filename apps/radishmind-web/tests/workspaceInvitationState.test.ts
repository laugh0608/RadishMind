import assert from "node:assert/strict";
import test from "node:test";
import { createInvitationRequestScope, isWorkspaceInvitationChangedEvent, mergeWorkspaceInvitationPages,
  workspaceInvitationAuthorityKey } from "../src/features/local-identity/workspaceInvitationState.ts";
import { availableIdentityWorkspaces } from "../src/features/local-identity/localIdentityWorkspaceAccess.ts";
import type { LocalIdentityAccountProfile } from "../src/features/local-identity/localIdentityConsumer.ts";
import type { WorkspaceInvitationPage } from "../src/features/local-identity/workspaceInvitationConsumer.ts";

const profile: LocalIdentityAccountProfile = {
  account: { userId: "usr_aaaaaaaaaaaaaaaa", lifecycleState: "active", displayName: "Claimant" },
  session: { sessionId: "ses_aaaaaaaaaaaaaaaa", expiresAt: "2026-09-09T00:00:00Z", authenticationMethod: "local_password" },
  externalIdentities: [], roleAssignments: [], workspaceMemberships: [],
  capabilities: { oidcEnabled: false, recentAuthentication: true, hasActiveLocalCredential: true },
};

test("shared invitation request scope rejects superseded, cancelled, disposed and cross-authority responses", () => {
  let authority = "actor-a:session-a:workspace-a:0";
  const scope = createInvitationRequestScope(() => authority);
  const first = scope.begin();
  assert.equal(scope.isCurrent(first), true);
  const second = scope.begin();
  assert.equal(first.signal.aborted, true);
  assert.equal(scope.isCurrent(first), false);
  assert.equal(scope.isCurrent(second), true);
  authority = "actor-a:session-a:workspace-a:1";
  assert.equal(scope.isCurrent(second), false, "Cross-tab authority changes must invalidate even before React renders.");
  const third = scope.begin();
  scope.invalidate();
  assert.equal(third.signal.aborted, true);
  assert.equal(scope.isCurrent(third), false);
  const fourth = scope.begin();
  scope.dispose();
  assert.equal(scope.isCurrent(fourth), false);
  assert.equal(scope.begin().signal.aborted, true);
  scope.activate();
  const fifth = scope.begin();
  assert.equal(scope.isCurrent(fifth), true);
  assert.equal(scope.isCurrent(fourth), false, "Strict Mode reactivation must not resurrect old work.");
});

test("late invitation completion cannot cross account, session, scope, route or recent-auth boundaries", async () => {
  const original = workspaceInvitationAuthorityKey(profile, "tenant-a/workspace-a/admin-invitations");
  for (const candidate of [
    { ...profile, account: { ...profile.account, userId: "usr_bbbbbbbbbbbbbbbb" } },
    { ...profile, account: { ...profile.account, lifecycleState: "disabled" as const } },
    { ...profile, session: { ...profile.session, sessionId: "ses_bbbbbbbbbbbbbbbb" } },
    { ...profile, session: { ...profile.session, expiresAt: "2026-09-10T00:00:00Z" } },
    { ...profile, capabilities: { ...profile.capabilities, recentAuthentication: false } },
  ]) assert.notEqual(workspaceInvitationAuthorityKey(candidate, "tenant-a/workspace-a/admin-invitations"), original);
  for (const target of ["tenant-b/workspace-a/admin-invitations", "tenant-a/workspace-b/admin-invitations", "tenant-a/workspace-a/claim"]) {
    assert.notEqual(workspaceInvitationAuthorityKey(profile, target), original);
  }
  let authority = original;
  const scope = createInvitationRequestScope(() => authority);
  const request = scope.begin();
  let resolve!: () => void;
  const pending = new Promise<void>((done) => { resolve = done; });
  let applied = false;
  const completion = pending.then(() => { if (scope.isCurrent(request)) applied = true; });
  authority = "new-actor";
  resolve(); await completion;
  assert.equal(applied, false, "Ignoring AbortSignal must not bypass the authority guard.");
});

test("invitation broadcast accepts metadata-only invalidation and never accepts resource or secret payloads", () => {
  assert.equal(isWorkspaceInvitationChangedEvent({ kind: "invitations_changed", version: 1 }), true);
  for (const value of [null, [], "invitations_changed", { kind: "invitations_changed" },
    { kind: "invitations_changed", version: 2 }, { kind: "session_changed", version: 1 },
    { kind: "invitations_changed", version: 1, invitation_code: "secret" },
    { kind: "invitations_changed", version: 1, workspace_id: "workspace-a" }]) {
    assert.equal(isWorkspaceInvitationChangedEvent(value), false);
  }
});

test("invitation pagination rejects changed snapshots, repeated records, scope changes and cursor loops", () => {
  const page = { requestId: "request:one", tenantRef: "tenant-a", workspaceId: "workspace-a", asOf: "2026-09-08T00:00:00Z",
    invitations: [{ invitationId: "wsi_a" }], nextCursor: "cursor-one" } as WorkspaceInvitationPage;
  const next = { ...page, requestId: "request:two", invitations: [{ invitationId: "wsi_b" }], nextCursor: "" } as WorkspaceInvitationPage;
  assert.deepEqual(mergeWorkspaceInvitationPages(page, next).invitations.map((entry) => entry.invitationId), ["wsi_a", "wsi_b"]);
  for (const changed of [{ ...next, asOf: "2026-09-08T01:00:00Z" }, { ...next, tenantRef: "tenant-b" },
    { ...next, workspaceId: "workspace-b" }, { ...next, invitations: page.invitations },
    { ...next, invitations: [...next.invitations, ...next.invitations] }, { ...next, nextCursor: page.nextCursor }]) {
    assert.throws(() => mergeWorkspaceInvitationPages(page, changed), { code: "workspace_invitation_response_invalid" });
  }
});

test("available workspace projection uses active unexpired membership and never selects a workspace", () => {
  const current = { ...profile, workspaceMemberships: [
    { membershipId: "mbr_1", tenantRef: "tenant-a", workspaceId: "workspace-a", lifecycleState: "active" as const, recordVersion: 1 },
    { membershipId: "mbr_2", tenantRef: "tenant-a", workspaceId: "workspace-a", lifecycleState: "active" as const, recordVersion: 1 },
    { membershipId: "mbr_3", tenantRef: "tenant-a", workspaceId: "workspace-b", lifecycleState: "revoked" as const, recordVersion: 2 },
    { membershipId: "mbr_4", tenantRef: "tenant-a", workspaceId: "workspace-c", lifecycleState: "active" as const, recordVersion: 1, expiresAt: "2026-09-08T00:00:00Z" },
  ] };
  const snapshot = structuredClone(current);
  assert.deepEqual(availableIdentityWorkspaces(current, Date.parse("2026-09-08T00:00:00Z")).map((entry) => entry.workspaceId), ["workspace-a"]);
  assert.deepEqual(current, snapshot);
});
