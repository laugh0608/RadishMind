import assert from "node:assert/strict";
import test from "node:test";

import {
  isValidAdminGatewayRequestQuotaLimit,
  putAdminGatewayRequestQuota,
  readAdminGatewayRequestQuota,
  type AdminGatewayRequestQuotaConfig,
} from "../src/features/control-plane-read/adminGatewayRequestQuotaConsumer.ts";

const config: AdminGatewayRequestQuotaConfig = {
  mode: "dev_admin_gateway_request_quota_http",
  baseUrl: "http://platform.test",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  environment: "test",
  applicationId: "app_flow_copilot",
  subjectRef: "subject_demo_user",
};

test("Admin Gateway quota consumer stays offline without a request", async () => {
  const originalFetch = globalThis.fetch;
  let fetchCount = 0;
  globalThis.fetch = async () => {
    fetchCount += 1;
    throw new Error("unexpected fetch");
  };
  try {
    const envelope = await readAdminGatewayRequestQuota({ ...config, mode: "offline" });
    assert.equal(envelope.failureCode, "gateway_quota_disabled");
    assert.equal(envelope.policy, null);
    assert.equal(envelope.usage, null);
    assert.equal(fetchCount, 0);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Admin Gateway quota GET sends the exact workspace, environment, and read permission", async () => {
  const originalFetch = globalThis.fetch;
  let captured: { url: string; method: string; headers: Headers } | null = null;
  globalThis.fetch = async (input, init) => {
    captured = { url: String(input), method: String(init?.method), headers: new Headers(init?.headers) };
    return jsonResponse(successEnvelope());
  };
  try {
    const envelope = await readAdminGatewayRequestQuota(config);
    assert.equal(envelope.policy?.recordVersion, 3);
    assert.equal(envelope.usage?.admittedRequestCount, 80);
    assert.equal(envelope.usage?.remainingRequestCount, 20);
    assert.equal(captured?.url, "http://platform.test/v1/admin/gateway-request-quotas/app_flow_copilot");
    assert.equal(captured?.method, "GET");
    assert.equal(captured?.headers.get("X-RadishMind-Dev-Read-Scopes"), "admin_gateway_quotas:read");
    assert.equal(captured?.headers.get("X-RadishMind-Active-Workspace"), "workspace_demo");
    assert.equal(
      captured?.headers.get("X-RadishMind-Dev-Read-Membership-Permissions"),
      "admin_gateway_quotas:read",
    );
    assert.equal(captured?.headers.get("X-RadishMind-Dev-Gateway-Quota-Environment"), "test");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Admin Gateway quota PUT sends positive integer CAS and preserves a conflict envelope", async () => {
  const originalFetch = globalThis.fetch;
  let captured: { method: string; headers: Headers; body: unknown } | null = null;
  globalThis.fetch = async (_input, init) => {
    captured = {
      method: String(init?.method),
      headers: new Headers(init?.headers),
      body: JSON.parse(String(init?.body)),
    };
    return jsonResponse(failureEnvelope("gateway_quota_policy_version_conflict"), 409);
  };
  try {
    const envelope = await putAdminGatewayRequestQuota(config, 3, 140);
    assert.equal(envelope.failureCode, "gateway_quota_policy_version_conflict");
    assert.equal(envelope.policy, null);
    assert.deepEqual(captured?.body, { expected_version: 3, request_limit: 140 });
    assert.equal(captured?.method, "PUT");
    assert.equal(captured?.headers.get("X-RadishMind-Dev-Read-Scopes"), "admin_gateway_quotas:write");
    assert.equal(
      captured?.headers.get("X-RadishMind-Dev-Read-Membership-Permissions"),
      "admin_gateway_quotas:write",
    );
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Admin Gateway quota consumer preserves fail-closed management states", async () => {
  const originalFetch = globalThis.fetch;
  try {
    for (const failureCode of [
      "gateway_quota_scope_denied",
      "gateway_quota_environment_forbidden",
      "gateway_quota_policy_not_found",
      "gateway_quota_store_unavailable",
    ]) {
      globalThis.fetch = async () => jsonResponse(failureEnvelope(failureCode), 403);
      const envelope = await readAdminGatewayRequestQuota(config);
      assert.equal(envelope.failureCode, failureCode);
      assert.equal(envelope.policy, null);
      assert.equal(envelope.usage, null);
    }
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Admin Gateway quota consumer rejects forbidden, scope-drifted, and inconsistent responses", async () => {
  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async () => jsonResponse({
      ...successEnvelope(),
      debug: { endpoint: "https://forbidden.test" },
    });
    await assert.rejects(() => readAdminGatewayRequestQuota(config), /forbidden field/);

    globalThis.fetch = async () => jsonResponse({ ...successEnvelope(), workspace_id: "workspace_other" });
    await assert.rejects(() => readAdminGatewayRequestQuota(config), /invalid HTTP 200 envelope/);

    const inconsistent = successEnvelope();
    inconsistent.usage.remaining_request_count = 21;
    globalThis.fetch = async () => jsonResponse(inconsistent);
    await assert.rejects(() => readAdminGatewayRequestQuota(config), /invalid HTTP 200 envelope/);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Admin Gateway quota limit validation accepts only the documented positive integer range", async () => {
  assert.equal(isValidAdminGatewayRequestQuotaLimit(1), true);
  assert.equal(isValidAdminGatewayRequestQuotaLimit(1_000_000), true);
  for (const invalid of [0, -1, 1.5, 1_000_001, Number.NaN]) {
    assert.equal(isValidAdminGatewayRequestQuotaLimit(invalid), false);
  }
  await assert.rejects(() => putAdminGatewayRequestQuota(config, 3, 0), /update is invalid/);
});

function successEnvelope() {
  return {
    request_id: "admin-gateway-quota-read-test",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    environment: "test",
    application_id: "app_flow_copilot",
    policy: {
      schema_version: "gateway_request_quota_v1",
      policy_id: "quota_0123456789abcdef01234567",
      tenant_ref: "tenant_demo",
      workspace_id: "workspace_demo",
      environment: "test",
      application_id: "app_flow_copilot",
      period: "calendar_day_utc",
      request_limit: 100,
      record_version: 3,
      created_at: "2026-08-09T00:00:00Z",
      updated_at: "2026-08-09T01:00:00Z",
      created_by: "subject_platform_ops",
      updated_by: "subject_platform_ops",
      last_request_id: "quota-policy-put-test",
      last_audit_ref: "audit_quota_policy_put_test",
    },
    usage: {
      schema_version: "gateway_request_quota_v1",
      tenant_ref: "tenant_demo",
      workspace_id: "workspace_demo",
      environment: "test",
      application_id: "app_flow_copilot",
      period: "calendar_day_utc",
      period_start: "2026-08-09",
      policy_id: "quota_0123456789abcdef01234567",
      policy_version: 3,
      request_limit: 100,
      admitted_request_count: 80,
      remaining_request_count: 20,
      updated_at: "2026-08-09T01:00:00Z",
    },
    failure_code: null,
    audit_ref: "audit_admin_gateway_quota_read_test",
  };
}

function failureEnvelope(failureCode: string) {
  return {
    request_id: "admin-gateway-quota-write-test",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    environment: "test",
    application_id: "app_flow_copilot",
    policy: null,
    usage: null,
    failure_code: failureCode,
    audit_ref: "audit_admin_gateway_quota_write_test",
  };
}

function jsonResponse(value: unknown, status = 200) {
  return new Response(JSON.stringify(value), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}
