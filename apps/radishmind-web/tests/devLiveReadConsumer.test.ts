import assert from "node:assert/strict";
import test from "node:test";

import {
  initialControlPlaneReadDevLiveLoadState,
  loadControlPlaneReadDevLiveCollections,
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

test("Control Plane read consumer scopes catalog and API key lifecycle routes to the active workspace", async () => {
  const originalFetch = globalThis.fetch;
  const urls: string[] = [];
  globalThis.fetch = async (input) => {
    urls.push(String(input));
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
      applicationCatalogEnabled: true,
      apiKeyLifecycleEnabled: true,
      workspaceId: "workspace_browser",
    });
    assert.equal(urls.some((url) => url.endsWith("/v1/user-workspace/applications?workspace_id=workspace_browser&lifecycle_state=active&limit=100")), true);
    assert.equal(urls.some((url) => url.endsWith("/v1/user-workspace/api-keys?workspace_id=workspace_browser&limit=100")), true);
    assert.equal(urls.some((url) => url.endsWith("/v1/user-workspace/api-keys")), false);
  } finally {
    globalThis.fetch = originalFetch;
  }
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
