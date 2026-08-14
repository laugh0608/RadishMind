import assert from "node:assert/strict";
import test from "node:test";

import {
  initialControlPlaneReadDevLiveLoadState,
  loadControlPlaneReadDevLiveCollection,
  loadControlPlaneReadDevLiveCollections,
  normalizeActiveWorkspaceId,
  type ControlPlaneReadDevLiveConfig,
} from "../src/features/control-plane-read/devLiveReadConsumer.ts";

const offline: ControlPlaneReadDevLiveConfig = {
  mode: "offline_fixture",
  baseUrl: "http://127.0.0.1:7000",
  tenantRef: "tenant_demo",
  subjectRef: "subject_demo_user",
};

const live: ControlPlaneReadDevLiveConfig = { ...offline, mode: "dev_live_http" };

test("Control Plane read consumer keeps offline mode at zero requests", async () => {
  const originalFetch = globalThis.fetch;
  let calls = 0;
  globalThis.fetch = async () => {
    calls += 1;
    throw new Error("offline mode must not fetch");
  };
  try {
    const collections = await loadControlPlaneReadDevLiveCollections(offline);
    assert.deepEqual(collections, {});
    assert.equal(calls, 0);
    assert.equal(initialControlPlaneReadDevLiveLoadState(offline).status, "idle");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Control Plane read consumer preserves sanitized non-2xx envelopes", async () => {
  const originalFetch = globalThis.fetch;
  let calls = 0;
  globalThis.fetch = async () => {
    calls += 1;
    return new Response(JSON.stringify({
      request_id: `denied-${calls}`,
      tenant_ref: "tenant_demo",
      items: [],
      next_cursor: null,
      failure_code: "scope_denied",
      audit_ref: `audit:denied-${calls}`,
    }), { status: 403, headers: { "Content-Type": "application/json" } });
  };
  try {
    const collections = await loadControlPlaneReadDevLiveCollections(live);
    assert.equal(calls, 7);
    assert.equal(Object.keys(collections).length, 7);
    for (const collection of Object.values(collections)) {
      assert.equal(collection?.failureCode, "scope_denied");
      assert.equal(collection?.items.length, 0);
    }
    assert.equal(JSON.stringify(collections).includes("Authorization"), false);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Control Plane read consumer sends active workspace and dev membership only to workspace routes", async () => {
  const originalFetch = globalThis.fetch;
  const requests: Array<{ url: string; headers: Headers }> = [];
  globalThis.fetch = async (input, init) => {
    requests.push({ url: String(input), headers: new Headers(init?.headers) });
    return new Response(JSON.stringify({
      request_id: "request-workspace-scope",
      tenant_ref: "tenant_demo",
      items: [],
      next_cursor: null,
      failure_code: null,
      audit_ref: "audit-workspace-scope",
    }), { status: 200, headers: { "Content-Type": "application/json" } });
  };
  try {
    await loadControlPlaneReadDevLiveCollections({
      ...live,
      workspaceId: "workspace_browser",
    });
    const workspaceRequests = requests.filter(({ url }) => url.includes("/v1/user-workspace/"));
    const adminRequests = requests.filter(({ url }) => !url.includes("/v1/user-workspace/"));
    assert.equal(workspaceRequests.length, 5);
    assert.equal(adminRequests.length, 2);
    assert.equal(workspaceRequests.every(({ headers }) =>
      headers.get("X-RadishMind-Active-Workspace") === "workspace_browser" &&
      headers.get("X-RadishMind-Dev-Read-Membership-Workspace") === "workspace_browser" &&
      headers.has("X-RadishMind-Dev-Read-Membership-Permissions")), true);
    assert.equal(adminRequests.every(({ headers }) =>
      !headers.has("X-RadishMind-Active-Workspace") &&
      !headers.has("X-RadishMind-Dev-Read-Membership-Workspace")), true);
    assert.equal(requests.some(({ url }) => url.includes("workspace_id=")), false);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Audit pagination sends only the reviewed cursor, limit, and sort query", async () => {
  const originalFetch = globalThis.fetch;
  let capturedUrl = "";
  globalThis.fetch = async (input) => {
    capturedUrl = String(input);
    return new Response(JSON.stringify({
      request_id: "request-audit-page-two",
      tenant_ref: "tenant_demo",
      items: [],
      next_cursor: null,
      failure_code: null,
      audit_ref: "audit-page-two",
    }), { status: 200, headers: { "Content-Type": "application/json" } });
  };
  try {
    const collection = await loadControlPlaneReadDevLiveCollection(
      live,
      "audit-summary-list-route",
      { cursor: "cursor.v1/next+page", limit: 50, sort: "recorded_at_desc" },
    );
    assert.equal(
      capturedUrl,
      "http://127.0.0.1:7000/v1/control-plane/audit?cursor=cursor.v1%2Fnext%2Bpage&limit=50&sort=recorded_at_desc",
    );
    assert.equal(collection.requestId, "request-audit-page-two");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Cursor pagination is audit-only and validates client bounds before fetch", async () => {
  let calls = 0;
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => {
    calls += 1;
    throw new Error("unexpected fetch");
  };
  try {
    await assert.rejects(
      loadControlPlaneReadDevLiveCollection(live, "application-summary-list-route", { cursor: "cursor" }),
      /does not accept cursor pagination/,
    );
    await assert.rejects(
      loadControlPlaneReadDevLiveCollection(live, "audit-summary-list-route", { cursor: " " }),
      /audit cursor is invalid/,
    );
    await assert.rejects(
      loadControlPlaneReadDevLiveCollection(live, "audit-summary-list-route", { limit: 101 }),
      /audit limit is invalid/,
    );
    assert.equal(calls, 0);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("active workspace selection is strict and normalized without persistence", () => {
  assert.equal(normalizeActiveWorkspaceId(" workspace_browser "), "workspace_browser");
  assert.equal(normalizeActiveWorkspaceId(""), null);
  assert.equal(normalizeActiveWorkspaceId("workspace browser"), null);
  assert.equal(normalizeActiveWorkspaceId("x".repeat(161)), null);
});

test("Control Plane read consumer rejects a non-envelope HTTP failure", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => new Response(JSON.stringify({ error: "raw upstream detail" }), {
    status: 401,
    headers: { "Content-Type": "application/json" },
  });
  try {
    await assert.rejects(
      loadControlPlaneReadDevLiveCollections(live),
      /returned HTTP 401 with a non read-side envelope/,
    );
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Control Plane read consumer uses an in-memory signed token without dev headers", async () => {
  const originalFetch = globalThis.fetch;
  const globalWithToken = globalThis as typeof globalThis & {
    __RADISHMIND_CONTROL_PLANE_SIGNED_TEST_TOKEN__?: () => string;
  };
  globalWithToken.__RADISHMIND_CONTROL_PLANE_SIGNED_TEST_TOKEN__ = () => "test-token-material";
  let calls = 0;
  globalThis.fetch = async (_input, init) => {
    calls += 1;
    const headers = new Headers(init?.headers);
    assert.equal(headers.get("Authorization"), "Bearer test-token-material");
    assert.equal(headers.has("X-RadishMind-Dev-Read-Identity"), false);
    assert.equal(headers.has("X-RadishMind-Dev-Read-Membership-Workspace"), false);
    return new Response(JSON.stringify({
      request_id: "request-signed",
      tenant_ref: "tenant_demo",
      items: [],
      next_cursor: null,
      failure_code: null,
      audit_ref: "audit-signed",
    }), { status: 200 });
  };
  try {
    await loadControlPlaneReadDevLiveCollections({
      mode: "dev_live_http",
      baseUrl: "http://127.0.0.1:7000",
      tenantRef: "tenant_demo",
      subjectRef: "subject_demo_user",
      authMode: "signed_test_token",
    });
    assert.equal(calls, 7);
  } finally {
    globalThis.fetch = originalFetch;
    delete globalWithToken.__RADISHMIND_CONTROL_PLANE_SIGNED_TEST_TOKEN__;
  }
});

test("Control Plane read consumer isolates an in-memory OIDC integration token", async () => {
  const originalFetch = globalThis.fetch;
  const globalWithToken = globalThis as typeof globalThis & {
    __RADISHMIND_CONTROL_PLANE_OIDC_INTEGRATION_TOKEN__?: () => string;
    __RADISHMIND_CONTROL_PLANE_SIGNED_TEST_TOKEN__?: () => string;
  };
  globalWithToken.__RADISHMIND_CONTROL_PLANE_OIDC_INTEGRATION_TOKEN__ = () => "oidc-memory-token";
  globalWithToken.__RADISHMIND_CONTROL_PLANE_SIGNED_TEST_TOKEN__ = () => "must-not-fallback";
  let calls = 0;
  globalThis.fetch = async (input, init) => {
    calls += 1;
    const headers = new Headers(init?.headers);
    assert.equal(headers.get("Authorization"), "Bearer oidc-memory-token");
    assert.equal(headers.has("X-RadishMind-Dev-Read-Identity"), false);
    const isAdminRoute = String(input).includes("/v1/control-plane/");
    return new Response(JSON.stringify({
      request_id: "request-oidc",
      tenant_ref: "tenant_demo",
      items: [],
      next_cursor: null,
      failure_code: isAdminRoute ? null : "workspace_membership_unavailable",
      audit_ref: "audit-oidc",
    }), { status: isAdminRoute ? 200 : 503 });
  };
  try {
    const collections = await loadControlPlaneReadDevLiveCollections({
      mode: "dev_live_http",
      baseUrl: "http://127.0.0.1:7000",
      tenantRef: "tenant_demo",
      subjectRef: "subject_demo_user",
      authMode: "radish_oidc_integration_test",
      storeMode: "postgres_dev_test",
    });
    assert.equal(calls, 7);
    assert.equal(collections["tenant-summary-route"]?.failureCode, null);
    assert.equal(collections["audit-summary-list-route"]?.failureCode, null);
    assert.equal(collections["application-summary-list-route"]?.failureCode, "workspace_membership_unavailable");
  } finally {
    globalThis.fetch = originalFetch;
    delete globalWithToken.__RADISHMIND_CONTROL_PLANE_OIDC_INTEGRATION_TOKEN__;
    delete globalWithToken.__RADISHMIND_CONTROL_PLANE_SIGNED_TEST_TOKEN__;
  }
});
