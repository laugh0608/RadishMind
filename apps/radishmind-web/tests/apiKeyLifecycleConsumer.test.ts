import assert from "node:assert/strict";
import test from "node:test";

import {
  issueAPIKey,
  listAPIKeyRecords,
  readAPIKeyRecord,
  replaceAPIKeyListRecord,
  revokeAPIKey,
  validateAPIKeyIssueInput,
  type APIKeyIssueInput,
  type APIKeyLifecycleConfig,
} from "../src/features/control-plane-read/apiKeyLifecycleConsumer.ts";

const token = `rmd_dev_key_aaaaaaaaaaaaaaaa.${"A".repeat(43)}`;
const offline: APIKeyLifecycleConfig = {
  mode: "offline",
  baseUrl: "http://127.0.0.1:7000",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  subjectRef: "subject_demo_user",
  authMode: "dev_headers",
};
const live: APIKeyLifecycleConfig = { ...offline, mode: "dev_api_key_lifecycle_http" };

test("API key lifecycle stays offline without fetching or simulating credentials", async () => {
  const originalFetch = globalThis.fetch;
  let called = false;
  globalThis.fetch = async () => { called = true; throw new Error("unexpected fetch"); };
  try {
    assert.equal((await listAPIKeyRecords(offline, "app_aaaaaaaaaaaaaaaa")).status, "offline");
    const issued = await issueAPIKey(offline, issueInput());
    assert.equal(issued.status, "offline");
    assert.equal(issued.credentialToken, "");
    assert.equal(called, false);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("issue accepts the one-time credential only with no-store and strict management scopes", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async (url, init) => {
    assert.equal(String(url), "http://127.0.0.1:7000/v1/user-workspace/api-keys");
    assert.equal(init?.method, "POST");
    const headers = new Headers(init?.headers);
    assert.equal(headers.get("X-RadishMind-Dev-Read-Scopes"), "api_keys:write,api_keys:read");
    assert.equal(headers.get("X-RadishMind-Active-Workspace"), "workspace_demo");
    assert.equal(headers.get("X-RadishMind-Dev-Read-Membership-Workspace"), "workspace_demo");
    assert.equal(headers.get("X-RadishMind-Dev-Read-Membership-Permissions"), "api_keys:write");
    assert.equal(headers.has("Authorization"), false);
    const body = JSON.parse(String(init?.body)) as Record<string, unknown>;
    assert.deepEqual(body.scopes, ["models:read", "responses:invoke"]);
    assert.equal(body.application_id, "app_aaaaaaaaaaaaaaaa");
    return jsonResponse(issueEnvelope(), 201, { "Cache-Control": "private, no-store" });
  };
  try {
    const issued = await issueAPIKey(live, issueInput());
    assert.equal(issued.status, "issued");
    assert.equal(issued.record?.apiKeyId, "key_aaaaaaaaaaaaaaaa");
    assert.equal(issued.credentialToken, token);
    assert.equal(JSON.stringify(issued.record).includes(token), false);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("issue rejects a missing no-store header and an invalid credential without exposing it", async () => {
  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async () => jsonResponse(issueEnvelope(), 201);
    const missingNoStore = await issueAPIKey(live, issueInput());
    assert.equal(missingNoStore.failureCode, "api_key_store_unavailable");
    assert.equal(missingNoStore.credentialToken, "");

    globalThis.fetch = async () => jsonResponse({
      ...issueEnvelope(),
      credential: { token: "rmd_dev_invalid" },
    }, 201, { "Cache-Control": "no-store" });
    const invalidCredential = await issueAPIKey(live, issueInput());
    assert.equal(invalidCredential.failureCode, "api_key_store_unavailable");
    assert.equal(invalidCredential.credentialToken, "");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("list, detail, and revoke preserve application scope without credential material", async () => {
  const originalFetch = globalThis.fetch;
  let calls = 0;
  globalThis.fetch = async (url, init) => {
    calls += 1;
    const headers = new Headers(init?.headers);
    if (calls === 1) {
      const requestURL = new URL(String(url));
      assert.equal(requestURL.pathname, "/v1/user-workspace/api-keys");
      assert.equal(requestURL.searchParams.get("application_id"), "app_aaaaaaaaaaaaaaaa");
      assert.equal(headers.get("X-RadishMind-Dev-Read-Scopes"), "api_keys:read");
      return jsonResponse(listEnvelope());
    }
    if (calls === 2) {
      assert.equal(new URL(String(url)).pathname, "/v1/user-workspace/api-keys/key_aaaaaaaaaaaaaaaa");
      return jsonResponse(operationEnvelope());
    }
    assert.equal(new URL(String(url)).pathname, "/v1/user-workspace/api-keys/key_aaaaaaaaaaaaaaaa/revoke");
    assert.equal(headers.get("X-RadishMind-Dev-Read-Scopes"), "api_keys:revoke,api_keys:read");
    assert.equal(headers.get("X-RadishMind-Active-Workspace"), "workspace_demo");
    assert.equal(headers.get("X-RadishMind-Dev-Read-Membership-Workspace"), "workspace_demo");
    assert.equal(headers.get("X-RadishMind-Dev-Read-Membership-Permissions"), "api_keys:revoke");
    assert.deepEqual(JSON.parse(String(init?.body)), { workspace_id: "workspace_demo", expected_version: 1 });
    return jsonResponse(operationEnvelope({ revoked: true, version: 2 }));
  };
  try {
    const list = await listAPIKeyRecords(live, "app_aaaaaaaaaaaaaaaa");
    assert.equal(list.status, "ready");
    assert.equal(list.records[0]?.applicationId, "app_aaaaaaaaaaaaaaaa");
    assert.equal(JSON.stringify(list).includes("credential"), false);

    const detail = await readAPIKeyRecord(live, "key_aaaaaaaaaaaaaaaa");
    assert.equal(detail.status, "loaded");
    assert.equal(detail.credentialToken, "");

    const revoked = await revokeAPIKey(live, "key_aaaaaaaaaaaaaaaa", 1);
    assert.equal(revoked.status, "revoked");
    assert.equal(revoked.record?.effectiveState, "revoked");
    assert.equal(revoked.record?.recordVersion, 2);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("signed issue selects the active workspace without dev membership proof", async () => {
  const originalFetch = globalThis.fetch;
  const tokenGlobal = globalThis as typeof globalThis & {
    __RADISHMIND_CONTROL_PLANE_SIGNED_TEST_TOKEN__?: () => string;
  };
  const originalProvider = tokenGlobal.__RADISHMIND_CONTROL_PLANE_SIGNED_TEST_TOKEN__;
  let capturedHeaders = new Headers();
  tokenGlobal.__RADISHMIND_CONTROL_PLANE_SIGNED_TEST_TOKEN__ = () => "signed-api-key-token-in-memory";
  globalThis.fetch = async (_input, init) => {
    capturedHeaders = new Headers(init?.headers);
    return jsonResponse(issueEnvelope(), 201, { "Cache-Control": "private, no-store" });
  };
  try {
    const result = await issueAPIKey({ ...live, authMode: "signed_test_token" }, issueInput());
    assert.equal(result.status, "issued");
    assert.equal(capturedHeaders.get("Authorization"), "Bearer signed-api-key-token-in-memory");
    assert.equal(capturedHeaders.get("X-RadishMind-Active-Workspace"), "workspace_demo");
    assert.equal(capturedHeaders.has("X-RadishMind-Dev-Read-Membership-Workspace"), false);
    assert.equal(capturedHeaders.has("X-RadishMind-Dev-Read-Membership-Permissions"), false);
  } finally {
    globalThis.fetch = originalFetch;
    tokenGlobal.__RADISHMIND_CONTROL_PLANE_SIGNED_TEST_TOKEN__ = originalProvider;
  }
});

test("local Session API key transport includes cookies and no dev or bearer proof", async () => {
  const originalFetch = globalThis.fetch;
  const requests: Array<{ method: string; headers: Headers; credentials: RequestCredentials | undefined }> = [];
  globalThis.fetch = async (_input, init) => {
    const method = init?.method ?? "GET";
    requests.push({ method, headers: new Headers(init?.headers), credentials: init?.credentials });
    const document = method === "GET" ? listEnvelope() : issueEnvelope();
    const localActor = "user:usr_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa";
    if ("items" in document) document.items[0]!.owner_subject_ref = localActor;
    if ("record" in document) {
      document.record.owner_subject_ref = localActor;
      document.record.created_by_actor_ref = localActor;
    }
    return method === "GET"
      ? jsonResponse(document)
      : jsonResponse(document, 201, { "Cache-Control": "private, no-store" });
  };
  try {
    const result = await issueAPIKey({ ...live, authMode: "local_session_dev_test" }, issueInput());
    const listed = await listAPIKeyRecords({ ...live, authMode: "local_session_dev_test" }, "app_aaaaaaaaaaaaaaaa");
    const capturedHeaders = requests[0]?.headers ?? new Headers();
    assert.equal(result.status, "issued");
    assert.equal(listed.status, "ready");
    assert.equal(requests.every(({ credentials }) => credentials === "include"), true);
    assert.equal(capturedHeaders.get("X-RadishMind-Active-Tenant"), "tenant_demo");
    assert.equal(capturedHeaders.get("X-RadishMind-Active-Workspace"), "workspace_demo");
    assert.equal(requests[1]?.headers.get("X-RadishMind-Active-Tenant"), "tenant_demo");
    assert.equal(requests[1]?.headers.has("X-RadishMind-Active-Workspace"), false);
    assert.equal(capturedHeaders.has("Authorization"), false);
    assert.equal(capturedHeaders.has("X-RadishMind-Dev-Read-Identity"), false);
    assert.equal(capturedHeaders.has("X-RadishMind-Dev-Read-Tenant"), false);
    assert.equal(capturedHeaders.has("X-RadishMind-Dev-Read-Subject"), false);
    assert.equal(capturedHeaders.has("X-RadishMind-Dev-Read-Scopes"), false);
    assert.equal(capturedHeaders.has("X-RadishMind-Dev-Read-Audit"), false);
    assert.equal(capturedHeaders.has("X-RadishMind-Dev-Read-Membership-Workspace"), false);
    assert.equal(capturedHeaders.has("X-RadishMind-Dev-Read-Membership-Permissions"), false);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("strict response validation rejects unknown fields, scope drift, and sensitive output", async () => {
  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async () => jsonResponse({ ...listEnvelope(), debug: "unexpected" });
    assert.equal((await listAPIKeyRecords(live, "app_aaaaaaaaaaaaaaaa")).failureCode, "api_key_store_unavailable");

    const drift = listEnvelope();
    drift.items[0]!.owner_subject_ref = "subject_other";
    globalThis.fetch = async () => jsonResponse(drift);
    assert.equal((await listAPIKeyRecords(live, "app_aaaaaaaaaaaaaaaa")).failureCode, "api_key_store_unavailable");

    const sensitive = listEnvelope() as ReturnType<typeof listEnvelope> & { authorization: string };
    sensitive.authorization = "Bearer hidden";
    globalThis.fetch = async () => jsonResponse(sensitive);
    assert.equal((await listAPIKeyRecords(live, "app_aaaaaaaaaaaaaaaa")).failureCode, "api_key_store_unavailable");

    globalThis.fetch = async () => jsonResponse({ ...operationEnvelope(), credential: { token } });
    assert.equal((await readAPIKeyRecord(live, "key_aaaaaaaaaaaaaaaa")).failureCode, "api_key_store_unavailable");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("revoke maps the stored CAS conflict and keeps the current version", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => jsonResponse({
    ...operationEnvelope(),
    record: null,
    failure_code: "api_key_version_conflict",
    current_record_version: 3,
    current_effective_state: "active",
  }, 409);
  try {
    const conflict = await revokeAPIKey(live, "key_aaaaaaaaaaaaaaaa", 1);
    assert.equal(conflict.status, "version_conflict");
    assert.equal(conflict.currentRecordVersion, 3);
    assert.equal(conflict.credentialToken, "");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("issue input validation rejects invalid scope, expiry, identifiers, and sensitive display names", () => {
  assert.equal(validateAPIKeyIssueInput(issueInput()), "");
  assert.equal(
    validateAPIKeyIssueInput({ ...issueInput(), scopes: ["agent_copilot:invoke"] }),
    "",
  );
  assert.equal(validateAPIKeyIssueInput({ ...issueInput(), applicationId: "app_invalid" }), "api_key_payload_invalid");
  assert.equal(validateAPIKeyIssueInput({ ...issueInput(), scopes: [] }), "api_key_payload_invalid");
  assert.equal(validateAPIKeyIssueInput({ ...issueInput(), expiresInDays: 91 }), "api_key_payload_invalid");
  assert.equal(validateAPIKeyIssueInput({ ...issueInput(), displayName: "Authorization: Bearer hidden" }), "api_key_secret_material_forbidden");
});

test("exact API key detail replaces the matching list projection without changing pagination state", async () => {
  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async () => jsonResponse(listEnvelope());
    const list = await listAPIKeyRecords(live, "app_aaaaaaaaaaaaaaaa");
    const exactRecord = {
      ...list.records[0]!,
      lastUsedAt: "2026-07-30T10:22:42Z",
      requestId: "api-key-exact-read-request",
      auditRef: "audit-api-key-exact-read-request",
    };
    const replaced = replaceAPIKeyListRecord(list, exactRecord);

    assert.equal(replaced.records[0]?.lastUsedAt, "2026-07-30T10:22:42Z");
    assert.equal(replaced.nextCursor, list.nextCursor);
    assert.equal(replaced.status, list.status);
    assert.equal(replaceAPIKeyListRecord(list, { ...exactRecord, apiKeyId: "key_bbbbbbbbbbbbbbbb" }), list);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

function issueInput(): APIKeyIssueInput {
  return {
    applicationId: "app_aaaaaaaaaaaaaaaa",
    displayName: "Browser development key",
    scopes: ["responses:invoke", "models:read"],
    expiresInDays: 30,
  };
}

function recordDocument(options: { revoked?: boolean; version?: number } = {}) {
  const revoked = options.revoked ?? false;
  const version = options.version ?? 1;
  return {
    schema_version: "api_key_record.v1",
    api_key_id: "key_aaaaaaaaaaaaaaaa",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: "app_aaaaaaaaaaaaaaaa",
    owner_subject_ref: "subject_demo_user",
    display_name: "Browser development key",
    scopes: ["models:read", "responses:invoke"],
    lifecycle_state: revoked ? "revoked" : "active",
    effective_state: revoked ? "revoked" : "active",
    record_version: version,
    created_at: "2026-07-15T01:00:00Z",
    expires_at: "2026-08-14T01:00:00Z",
    last_used_at: null,
    revoked_at: revoked ? "2026-07-15T01:10:00Z" : null,
    created_by_actor_ref: "subject_demo_user",
    revoked_by_actor_ref: revoked ? "subject_demo_user" : null,
    request_id: "api-key-operation-request-0001",
    audit_ref: "audit-api-key-operation-request-0001",
  };
}

function operationEnvelope(options: { revoked?: boolean; version?: number } = {}) {
  const record = recordDocument(options);
  return {
    request_id: "api-key-operation-response-0001",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    record,
    failure_code: null,
    current_record_version: record.record_version,
    current_effective_state: record.effective_state,
    audit_ref: "audit-api-key-operation-response-0001",
  };
}

function issueEnvelope() {
  return { ...operationEnvelope(), credential: { token } };
}

function listEnvelope() {
  const record = recordDocument();
  return {
    request_id: "api-key-list-response-0001",
    tenant_ref: "tenant_demo",
    items: [{
      api_key_id: record.api_key_id,
      tenant_ref: record.tenant_ref,
      workspace_id: record.workspace_id,
      application_id: record.application_id,
      owner_subject_ref: record.owner_subject_ref,
      display_name: record.display_name,
      scopes: record.scopes,
      state: record.effective_state,
      lifecycle_state: record.lifecycle_state,
      effective_state: record.effective_state,
      record_version: record.record_version,
      created_at: record.created_at,
      expires_at: record.expires_at,
      last_used_at: record.last_used_at,
      revoked_at: record.revoked_at,
    }],
    next_cursor: null,
    failure_code: null,
    audit_ref: "audit-api-key-list-response-0001",
  };
}

function jsonResponse(body: unknown, status = 200, headers: Record<string, string> = {}): Response {
  return new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json", ...headers } });
}
