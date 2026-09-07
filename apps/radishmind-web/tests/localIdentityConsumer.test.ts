import assert from "node:assert/strict";
import test from "node:test";

import {
  LocalIdentityConsumerError,
  authenticateLocalIdentity,
  localIdentityConsumerConfigFromEnvironment,
  localIdentityReturnTarget,
  probeLocalIdentitySession,
  readLocalIdentityAccountProfile,
  readLocalIdentityCSRFCookie,
  revokeLocalIdentityExternalIdentity,
  startLocalIdentityOIDC,
  type LocalIdentityConsumerConfig,
} from "../src/features/local-identity/localIdentityConsumer.ts";

const config: LocalIdentityConsumerConfig = {
  mode: "local_identity_dev",
  baseUrl: "http://platform.test",
};

test("local identity Web consumer is opt-in and normalizes unsafe base URLs", () => {
  assert.deepEqual(localIdentityConsumerConfigFromEnvironment({}), {
    mode: "disabled",
    baseUrl: "http://127.0.0.1:7000",
  });
  assert.deepEqual(localIdentityConsumerConfigFromEnvironment({
    VITE_RADISHMIND_LOCAL_IDENTITY_MODE: "local_identity_dev",
    VITE_RADISHMIND_LOCAL_IDENTITY_BASE_URL: "https://mind.example.test/platform/",
  }), {
    mode: "local_identity_dev",
    baseUrl: "https://mind.example.test/platform",
  });
  assert.equal(localIdentityConsumerConfigFromEnvironment({
    VITE_RADISHMIND_LOCAL_IDENTITY_BASE_URL: "http://remote.example.test",
  }).baseUrl, "http://127.0.0.1:7000");
});

test("local identity return targets keep only the server-reviewed path and query", () => {
  assert.equal(localIdentityReturnTarget({ pathname: "/workspace", search: "?view=active" }), "/workspace?view=active");
  assert.equal(localIdentityReturnTarget({ pathname: "//remote.example.test", search: "" }), "/");
});

test("session probe sends only credentialed cookie transport and parses the sanitized document", async () => {
  const originalFetch = globalThis.fetch;
  let captured: { credentials?: RequestCredentials; cache?: RequestCache; headers: Headers } | null = null;
  globalThis.fetch = async (_input, init) => {
    captured = { credentials: init?.credentials, cache: init?.cache, headers: new Headers(init?.headers) };
    return jsonResponse(authenticationDocument());
  };
  try {
    const authentication = await probeLocalIdentitySession(config);
    assert.equal(authentication?.account.displayName, "Local User");
    assert.equal(captured?.credentials, "include");
    assert.equal(captured?.cache, "no-store");
    assert.equal(captured?.headers.get("Authorization"), null);
    assert.equal(captured?.headers.get("X-RadishMind-CSRF-Token"), null);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("local registration uses the bootstrap CSRF proof without persisting session material", async () => {
  const originalFetch = globalThis.fetch;
  let captured: { url: string; headers: Headers; body: any } | null = null;
  globalThis.fetch = async (input, init) => {
    captured = {
      url: String(input),
      headers: new Headers(init?.headers),
      body: JSON.parse(String(init?.body)),
    };
    return jsonResponse(authenticationDocument(), 201);
  };
  try {
    await authenticateLocalIdentity(config, {
      intent: "register",
      loginIdentifier: "local@example.com",
      displayName: "Local User",
      password: "a sufficiently long password",
      returnTo: "/workspace",
    });
    assert.equal(captured?.url, "http://platform.test/v1/auth/local/register");
    assert.equal(captured?.headers.get("X-RadishMind-CSRF-Token"), "bootstrap");
    assert.deepEqual(captured?.body, {
      login_identifier: "local@example.com",
      display_name: "Local User",
      password: "a sufficiently long password",
      return_to: "/workspace",
    });
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("OIDC login start sends only the reviewed bootstrap request and validates the authorization URL", async () => {
  const originalFetch = globalThis.fetch;
  let captured: { url: string; headers: Headers; body: any } | null = null;
  globalThis.fetch = async (input, init) => {
    captured = { url: String(input), headers: new Headers(init?.headers), body: JSON.parse(String(init?.body)) };
    return jsonResponse({
      authorization_url: "https://radish.example.test/authorize?state=opaque",
      expires_at: "2026-08-24T12:00:00Z",
    });
  };
  try {
    const authorization = await startLocalIdentityOIDC(config, "login", "/workspace?view=active");
    assert.equal(authorization.authorizationUrl, "https://radish.example.test/authorize?state=opaque");
    assert.equal(captured?.url, "http://platform.test/v1/auth/oidc/start");
    assert.equal(captured?.headers.get("X-RadishMind-CSRF-Token"), "bootstrap");
    assert.deepEqual(captured?.body, { intent: "login", return_to: "/workspace?view=active" });
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("account profile accepts local grants but rejects upstream identity claims", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => jsonResponse(accountProfileDocument());
  try {
    const profile = await readLocalIdentityAccountProfile(config);
    assert.equal(profile.externalIdentities[0]?.providerRef, "radish_oidc");
    assert.equal(profile.roleAssignments[0]?.roleKey, "workspace_reader");
    assert.deepEqual(profile.roleAssignments[0]?.permissionGrants, ["applications:read"]);

    globalThis.fetch = async () => jsonResponse({
      ...accountProfileDocument(),
      subject: "upstream-subject-must-not-cross",
    });
    await assert.rejects(
      () => readLocalIdentityAccountProfile(config),
      (error: unknown) => error instanceof LocalIdentityConsumerError &&
        error.code === "local_identity_response_invalid",
    );
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("external identity revoke requires exactly one reviewed CSRF cookie and sends CAS", async () => {
  assert.equal(readLocalIdentityCSRFCookie("other=1; radishmind_csrf_dev=proof-123"), "proof-123");
  assert.throws(
    () => readLocalIdentityCSRFCookie("radishmind_csrf_dev=one; __Host-radishmind_csrf=two"),
    /CSRF proof is unavailable/u,
  );
  const originalFetch = globalThis.fetch;
  const originalDocument = Object.getOwnPropertyDescriptor(globalThis, "document");
  let captured: { url: string; headers: Headers; body: any } | null = null;
  Object.defineProperty(globalThis, "document", { configurable: true, value: { cookie: "radishmind_csrf_dev=proof-123" } });
  globalThis.fetch = async (input, init) => {
    captured = { url: String(input), headers: new Headers(init?.headers), body: JSON.parse(String(init?.body)) };
    return new Response(null, { status: 204 });
  };
  try {
    await revokeLocalIdentityExternalIdentity(config, "xid_aaaaaaaaaaaaaaaa", 3);
    assert.equal(captured?.url, "http://platform.test/v1/auth/external-identities/xid_aaaaaaaaaaaaaaaa/revoke");
    assert.equal(captured?.headers.get("X-RadishMind-CSRF-Token"), "proof-123");
    assert.deepEqual(captured?.body, { expected_record_version: 3 });
  } finally {
    globalThis.fetch = originalFetch;
    if (originalDocument) Object.defineProperty(globalThis, "document", originalDocument);
    else Reflect.deleteProperty(globalThis, "document");
  }
});

test("local identity error documents reject additional or cross-boundary fields", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => jsonResponse({
    error: {
      message: "an active local session is required",
      type: "authentication_error",
      code: "LOCAL_IDENTITY_AUTHENTICATION_REQUIRED",
      request_id: "req_aaaaaaaaaaaaaaaa",
      route: "/v1/auth/account",
      failure_boundary: "local_identity",
      detail: "must not be accepted",
    },
  }, 401);
  try {
    await assert.rejects(
      () => readLocalIdentityAccountProfile(config),
      (error: unknown) => error instanceof LocalIdentityConsumerError &&
        error.code === "local_identity_response_invalid",
    );
  } finally {
    globalThis.fetch = originalFetch;
  }
});

function authenticationDocument() {
  return {
    account: {
      user_id: "usr_aaaaaaaaaaaaaaaa",
      display_name: "Local User",
      lifecycle_state: "active",
    },
    session: {
      session_id: "ses_aaaaaaaaaaaaaaaa",
      authentication_method: "local_password",
      expires_at: "2026-08-24T12:00:00Z",
    },
    return_to: "/workspace",
  };
}

function accountProfileDocument() {
  return {
    account: authenticationDocument().account,
    session: authenticationDocument().session,
    external_identities: [{
      binding_id: "xid_aaaaaaaaaaaaaaaa",
      provider_ref: "radish_oidc",
      lifecycle_state: "active",
      record_version: 1,
      created_at: "2026-08-23T12:00:00Z",
      updated_at: "2026-08-23T12:00:00Z",
      can_revoke: true,
    }],
    role_assignments: [{
      assignment_id: "rla_aaaaaaaaaaaaaaaa",
      tenant_ref: "tenant_demo",
      workspace_id: "workspace_demo",
      role_key: "workspace_reader",
      permission_grants: ["applications:read"],
      lifecycle_state: "active",
      record_version: 1,
    }],
    workspace_memberships: [{
      membership_id: "mbr_aaaaaaaaaaaaaaaa",
      tenant_ref: "tenant_demo",
      workspace_id: "workspace_demo",
      lifecycle_state: "active",
      record_version: 1,
    }],
    capabilities: {
      oidc_enabled: true,
      recent_authentication: true,
      has_active_local_credential: true,
    },
  };
}

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}
