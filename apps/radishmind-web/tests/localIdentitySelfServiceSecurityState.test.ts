import assert from "node:assert/strict";
import test from "node:test";

import {
  localIdentitySelfServiceSecurityResponseMatchesScope,
  localIdentitySelfServiceSecurityScope,
  localIdentitySelfServiceSecurityScopeKey,
  mergeLocalIdentitySelfServiceSessions,
  projectLocalIdentitySelfServiceSessions,
} from "../src/features/local-identity/localIdentitySelfServiceSecurityState.ts";
import { LocalIdentitySelfServiceSecurityError } from "../src/features/local-identity/localIdentitySelfServiceSecurityConsumer.ts";
import type { LocalIdentityAccountProfile } from "../src/features/local-identity/localIdentityConsumer.ts";

test("security scope rotates for actor, current session, or invalidation generation changes", () => {
  const profile = accountProfile();
  const first = localIdentitySelfServiceSecurityScope(profile, 4);
  assert.equal(localIdentitySelfServiceSecurityScopeKey(profile, 4), "usr_aaaaaaaaaaaaaaaa:ses_aaaaaaaaaaaaaaaa:4");
  assert.equal(localIdentitySelfServiceSecurityResponseMatchesScope(first, { ...first }), true);
  assert.equal(localIdentitySelfServiceSecurityResponseMatchesScope(first, { ...first, generation: 5 }), false);
  assert.equal(localIdentitySelfServiceSecurityResponseMatchesScope(first, {
    ...first,
    currentSessionId: "ses_bbbbbbbbbbbbbbbb",
  }), false);
  assert.equal(localIdentitySelfServiceSecurityResponseMatchesScope(first, {
    ...first,
    userId: "usr_bbbbbbbbbbbbbbbb",
  }), false);
});

test("session projection keeps one current owner, other active rows, history, and local credential impact", () => {
  const sessions = [
    session("ses_aaaaaaaaaaaaaaaa", "active", true, "local_password"),
    session("ses_bbbbbbbbbbbbbbbb", "active", false, "oidc"),
    session("ses_cccccccccccccccc", "active", false, "local_password"),
    session("ses_dddddddddddddddd", "expired", false, "local_password"),
    session("ses_eeeeeeeeeeeeeeee", "revoked", false, "oidc"),
  ];
  const projection = projectLocalIdentitySelfServiceSessions(sessions, "ses_aaaaaaaaaaaaaaaa");
  assert.equal(projection.currentSession?.sessionId, "ses_aaaaaaaaaaaaaaaa");
  assert.deepEqual(projection.otherActiveSessions.map((item) => item.sessionId), [
    "ses_bbbbbbbbbbbbbbbb",
    "ses_cccccccccccccccc",
  ]);
  assert.deepEqual(projection.endedSessions.map((item) => item.effectiveState), ["expired", "revoked"]);
  assert.deepEqual(projection.activeLocalPasswordSessions.map((item) => item.sessionId), [
    "ses_aaaaaaaaaaaaaaaa",
    "ses_cccccccccccccccc",
  ]);
});

test("session pagination and current actor drift fail closed", () => {
  const current = session("ses_aaaaaaaaaaaaaaaa", "active", true, "local_password");
  assert.throws(
    () => mergeLocalIdentitySelfServiceSessions([current], [current]),
    isInvalidResponse,
  );
  assert.throws(
    () => projectLocalIdentitySelfServiceSessions([
      { ...current, currentSession: false },
    ], "ses_aaaaaaaaaaaaaaaa"),
    isInvalidResponse,
  );
  assert.throws(
    () => projectLocalIdentitySelfServiceSessions([
      current,
      session("ses_bbbbbbbbbbbbbbbb", "active", true, "oidc"),
    ], "ses_aaaaaaaaaaaaaaaa"),
    isInvalidResponse,
  );
});

function accountProfile(): LocalIdentityAccountProfile {
  return {
    account: { userId: "usr_aaaaaaaaaaaaaaaa", displayName: "Local User", lifecycleState: "active" },
    session: {
      sessionId: "ses_aaaaaaaaaaaaaaaa",
      authenticationMethod: "local_password",
      expiresAt: "2026-08-26T20:00:00Z",
    },
    externalIdentities: [],
    roleAssignments: [],
    workspaceMemberships: [],
    capabilities: { oidcEnabled: false, recentAuthentication: true, hasActiveLocalCredential: true },
  };
}

function session(
  sessionId: string,
  effectiveState: "active" | "expired" | "revoked",
  currentSession: boolean,
  authenticationMethod: "local_password" | "oidc",
) {
  return {
    schemaVersion: "local_identity_self_service_session_summary.v1" as const,
    sessionId,
    authenticationMethod,
    effectiveState,
    recordVersion: 1,
    currentSession,
    createdAt: "2026-08-26T08:00:00Z",
    lastVerifiedAt: "2026-08-26T09:00:00Z",
    expiresAt: "2026-08-26T20:00:00Z",
    ...(effectiveState === "revoked" ? { revokedAt: "2026-08-26T10:00:00Z" } : {}),
  };
}

function isInvalidResponse(error: unknown): boolean {
  return error instanceof LocalIdentitySelfServiceSecurityError && error.code === "local_identity_response_invalid";
}
