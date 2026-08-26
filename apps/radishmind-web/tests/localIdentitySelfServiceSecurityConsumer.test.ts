import assert from "node:assert/strict";
import test from "node:test";

import {
  LocalIdentitySelfServiceSecurityError,
  listLocalIdentitySelfServiceSessions,
  localIdentitySelfServiceSecurityFailureKind,
  revokeLocalIdentitySelfServiceSession,
  revokeOtherLocalIdentitySelfServiceSessions,
  rotateLocalIdentitySelfServiceCredential,
} from "../src/features/local-identity/localIdentitySelfServiceSecurityConsumer.ts";
import type { LocalIdentityConsumerConfig } from "../src/features/local-identity/localIdentityConsumer.ts";

const config: LocalIdentityConsumerConfig = {
  mode: "local_identity_dev",
  baseUrl: "http://platform.test",
};

test("self-service session directory sends one filter-bound request and parses canonical sessions", async () => {
  const originalFetch = globalThis.fetch;
  const controller = new AbortController();
  let captured: { url: string; method?: string; credentials?: RequestCredentials; signal?: AbortSignal | null; headers: Headers } | null = null;
  globalThis.fetch = async (input, init) => {
    captured = {
      url: String(input),
      method: init?.method,
      credentials: init?.credentials,
      signal: init?.signal,
      headers: new Headers(init?.headers),
    };
    return jsonResponse(sessionPageDocument());
  };
  try {
    const page = await listLocalIdentitySelfServiceSessions(config, {
      state: "all",
      limit: 100,
      cursor: "cursor_aaaaaaaaaaaaaaaa",
    }, controller.signal);
    assert.equal(captured?.url, "http://platform.test/v1/auth/sessions?state=all&limit=100&cursor=cursor_aaaaaaaaaaaaaaaa");
    assert.equal(captured?.method, "GET");
    assert.equal(captured?.credentials, "include");
    assert.equal(captured?.signal, controller.signal);
    assert.equal(captured?.headers.get("Authorization"), null);
    assert.equal(captured?.headers.get("X-RadishMind-CSRF-Token"), null);
    assert.equal(page.sessions[0]?.currentSession, true);
    assert.equal(page.sessions[1]?.effectiveState, "revoked");
    assert.equal(page.nextCursor, "cursor_bbbbbbbbbbbbbbbb");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("session directory rejects unsafe query input and any schema or sensitive-field drift", async () => {
  await assert.rejects(
    () => listLocalIdentitySelfServiceSessions(config, { state: "all", limit: 101 }),
    hasCode("local_identity_input_invalid"),
  );

  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async () => jsonResponse({
      ...sessionPageDocument(),
      sessions: [{
        ...sessionPageDocument().sessions[0],
        authentication_source_ref: "credential:must-not-cross",
      }],
    });
    await assert.rejects(
      () => listLocalIdentitySelfServiceSessions(config, { state: "all" }),
      hasCode("local_identity_response_invalid"),
    );

    globalThis.fetch = async () => jsonResponse({
      ...sessionPageDocument(),
      extra_projection: true,
    });
    await assert.rejects(
      () => listLocalIdentitySelfServiceSessions(config, { state: "all" }),
      hasCode("local_identity_response_invalid"),
    );
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("exact revoke sends CAS plus confirmation and accepts only a revoked canonical target", async () => {
  await withCSRFCookie(async () => {
    const originalFetch = globalThis.fetch;
    let captured: { url: string; headers: Headers; body: unknown; credentials?: RequestCredentials } | null = null;
    globalThis.fetch = async (input, init) => {
      captured = {
        url: String(input),
        headers: new Headers(init?.headers),
        body: JSON.parse(String(init?.body)),
        credentials: init?.credentials,
      };
      return jsonResponse({
        schema_version: "local_identity_self_service_session_revocation.v1",
        session: {
          ...sessionPageDocument().sessions[0],
          effective_state: "revoked",
          record_version: 4,
          revoked_at: "2026-08-26T09:30:00Z",
        },
        current_session_revoked: true,
      });
    };
    try {
      const result = await revokeLocalIdentitySelfServiceSession(config, {
        sessionId: "ses_aaaaaaaaaaaaaaaa",
        expectedRecordVersion: 3,
      });
      assert.equal(captured?.url, "http://platform.test/v1/auth/sessions/ses_aaaaaaaaaaaaaaaa/revoke");
      assert.equal(captured?.headers.get("X-RadishMind-CSRF-Token"), "csrf-proof");
      assert.equal(captured?.credentials, "include");
      assert.deepEqual(captured?.body, { expected_record_version: 3, confirmed: true });
      assert.equal(result.currentSessionRevoked, true);
    } finally {
      globalThis.fetch = originalFetch;
    }
  });
});

test("bulk revoke and credential rotation use their aggregate routes without adding client authority", async () => {
  await withCSRFCookie(async () => {
    const originalFetch = globalThis.fetch;
    const captured: Array<{ url: string; body: unknown }> = [];
    globalThis.fetch = async (input, init) => {
      const url = String(input);
      captured.push({ url, body: JSON.parse(String(init?.body)) });
      if (url.endsWith("/revoke-others")) {
        return jsonResponse({
          schema_version: "local_identity_self_service_bulk_session_revocation.v1",
          revoked_count: 2,
        });
      }
      return jsonResponse({
        schema_version: "local_identity_self_service_credential_rotation.v1",
        policy_version: "local_password_policy.v1",
        revoked_session_count: 3,
        current_session_revoked: false,
      });
    };
    try {
      const bulk = await revokeOtherLocalIdentitySelfServiceSessions(config);
      const rotation = await rotateLocalIdentitySelfServiceCredential(config, {
        currentPassword: "current password value",
        newPassword: "replacement password value",
      });
      assert.equal(bulk.revokedCount, 2);
      assert.equal(rotation.currentSessionRevoked, false);
      assert.deepEqual(captured, [
        {
          url: "http://platform.test/v1/auth/sessions/revoke-others",
          body: { confirmed: true },
        },
        {
          url: "http://platform.test/v1/auth/local/credential/rotate",
          body: {
            current_password: "current password value",
            new_password: "replacement password value",
            session_impact_confirmed: true,
          },
        },
      ]);
    } finally {
      globalThis.fetch = originalFetch;
    }
  });
});

test("self-service errors accept only the declared owner or local transport boundary", async () => {
  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async () => jsonResponse(errorDocument(
      "local_identity_session_version_conflict",
      "local_identity_self_service_security",
    ), 409);
    await assert.rejects(
      () => listLocalIdentitySelfServiceSessions(config, { state: "all" }),
      (error: unknown) => hasCode("local_identity_session_version_conflict")(error) &&
        localIdentitySelfServiceSecurityFailureKind(error) === "conflict",
    );

    globalThis.fetch = async () => jsonResponse(errorDocument(
      "LOCAL_IDENTITY_AUTHENTICATION_REQUIRED",
      "local_identity",
    ), 401);
    await assert.rejects(
      () => listLocalIdentitySelfServiceSessions(config, { state: "all" }),
      (error: unknown) => hasCode("LOCAL_IDENTITY_AUTHENTICATION_REQUIRED")(error) &&
        localIdentitySelfServiceSecurityFailureKind(error) === "authentication_required",
    );

    globalThis.fetch = async () => jsonResponse(errorDocument(
      "local_identity_session_version_conflict",
      "local_identity",
    ), 409);
    await assert.rejects(
      () => listLocalIdentitySelfServiceSessions(config, { state: "all" }),
      hasCode("local_identity_response_invalid"),
    );
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("failure classification covers denied, recent authentication, credential, unavailable and invalid responses", () => {
  const failure = (code: string) => new LocalIdentitySelfServiceSecurityError(409, code, "stable failure");
  assert.equal(localIdentitySelfServiceSecurityFailureKind(failure("local_identity_session_scope_denied")), "denied");
  assert.equal(localIdentitySelfServiceSecurityFailureKind(failure("local_identity_session_recent_authentication_required")), "recent_authentication");
  assert.equal(localIdentitySelfServiceSecurityFailureKind(failure("local_identity_credential_unavailable")), "credential_unavailable");
  assert.equal(localIdentitySelfServiceSecurityFailureKind(failure("local_identity_credential_current_invalid")), "credential_invalid");
  assert.equal(localIdentitySelfServiceSecurityFailureKind(failure("local_identity_credential_policy_rejected")), "credential_policy");
  assert.equal(localIdentitySelfServiceSecurityFailureKind(failure("local_identity_service_unavailable")), "unavailable");
  assert.equal(localIdentitySelfServiceSecurityFailureKind(failure("local_identity_response_invalid")), "invalid_response");
});

function sessionPageDocument() {
  return {
    sessions: [
      {
        schema_version: "local_identity_self_service_session_summary.v1",
        session_id: "ses_aaaaaaaaaaaaaaaa",
        authentication_method: "local_password",
        effective_state: "active",
        record_version: 3,
        current_session: true,
        created_at: "2026-08-26T08:00:00Z",
        last_verified_at: "2026-08-26T09:00:00Z",
        expires_at: "2026-08-26T20:00:00Z",
      },
      {
        schema_version: "local_identity_self_service_session_summary.v1",
        session_id: "ses_bbbbbbbbbbbbbbbb",
        authentication_method: "oidc",
        effective_state: "revoked",
        record_version: 2,
        current_session: false,
        created_at: "2026-08-25T08:00:00Z",
        last_verified_at: "2026-08-25T09:00:00Z",
        expires_at: "2026-08-25T20:00:00Z",
        revoked_at: "2026-08-25T10:00:00Z",
      },
    ],
    snapshot_at: "2026-08-26T09:15:00Z",
    next_cursor: "cursor_bbbbbbbbbbbbbbbb",
  };
}

function errorDocument(code: string, failureBoundary: string) {
  return {
    error: {
      message: "stable failure",
      type: "invalid_request_error",
      code,
      request_id: "req_aaaaaaaaaaaaaaaa",
      route: "/v1/auth/sessions",
      failure_boundary: failureBoundary,
    },
  };
}

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function hasCode(code: string): (error: unknown) => boolean {
  return (error: unknown) => error instanceof LocalIdentitySelfServiceSecurityError && error.code === code;
}

async function withCSRFCookie(run: () => Promise<void>): Promise<void> {
  const originalDocument = Object.getOwnPropertyDescriptor(globalThis, "document");
  Object.defineProperty(globalThis, "document", {
    configurable: true,
    value: { cookie: "radishmind_csrf_dev=csrf-proof" },
  });
  try {
    await run();
  } finally {
    if (originalDocument) Object.defineProperty(globalThis, "document", originalDocument);
    else Reflect.deleteProperty(globalThis, "document");
  }
}
