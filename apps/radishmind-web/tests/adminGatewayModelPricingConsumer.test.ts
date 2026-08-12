import assert from "node:assert/strict";
import test from "node:test";

import {
  isValidAdminGatewayModelPricingRate,
  isValidAdminGatewayModelPricingReason,
  putAdminGatewayModelPricing,
  readAdminGatewayModelPricing,
  type AdminGatewayModelPricingConfig,
  type AdminGatewayModelPricingScope,
} from "../src/features/control-plane-read/adminGatewayModelPricingConsumer.ts";

const scope: AdminGatewayModelPricingScope = {
  providerId: "mock",
  profileId: "mock-dev",
  modelId: "mock-model",
};

const config: AdminGatewayModelPricingConfig = {
  mode: "dev_admin_gateway_model_pricing_http",
  baseUrl: "http://platform.test",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  environment: "test",
  subjectRef: "subject_demo_user",
  initialScope: scope,
};

test("Admin Gateway pricing stays offline without a request", async () => {
  const originalFetch = globalThis.fetch;
  let calls = 0;
  globalThis.fetch = async () => { calls += 1; throw new Error("unexpected fetch"); };
  try {
    const envelope = await readAdminGatewayModelPricing({ ...config, mode: "offline" }, scope);
    assert.equal(envelope.failureCode, "gateway_pricing_disabled");
    assert.equal(envelope.policy, null);
    assert.equal(calls, 0);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Admin Gateway pricing GET sends exact selection scope and independent read permission", async () => {
  const originalFetch = globalThis.fetch;
  let captured: { url: URL; headers: Headers } | null = null;
  globalThis.fetch = async (input, init) => {
    captured = { url: new URL(String(input)), headers: new Headers(init?.headers) };
    return jsonResponse(successEnvelope());
  };
  try {
    const envelope = await readAdminGatewayModelPricing(config, scope);
    assert.equal(envelope.policy?.recordVersion, 3);
    assert.equal(envelope.policy?.inputPriceMicrosPerTokenUnit, 1_000_000);
    assert.equal(captured?.url.pathname, "/v1/admin/gateway-model-pricing-policy");
    assert.equal(captured?.url.searchParams.get("provider_id"), "mock");
    assert.equal(captured?.url.searchParams.get("profile_id"), "mock-dev");
    assert.equal(captured?.url.searchParams.get("model_id"), "mock-model");
    assert.equal(captured?.headers.get("X-RadishMind-Dev-Read-Scopes"), "admin_gateway_pricing:read");
    assert.equal(captured?.headers.get("X-RadishMind-Active-Workspace"), "workspace_demo");
    assert.equal(captured?.headers.get("X-RadishMind-Dev-Gateway-Pricing-Environment"), "test");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Admin Gateway pricing PUT sends one reviewed CAS revision and preserves conflict metadata", async () => {
  const originalFetch = globalThis.fetch;
  let captured: { body: unknown; headers: Headers } | null = null;
  globalThis.fetch = async (_input, init) => {
    captured = { body: JSON.parse(String(init?.body)), headers: new Headers(init?.headers) };
    return jsonResponse(failureEnvelope("gateway_pricing_policy_version_conflict", 4), 409);
  };
  try {
    const envelope = await putAdminGatewayModelPricing(config, scope, {
      expectedVersion: 3,
      inputPriceMicrosPerTokenUnit: 1_250_000,
      outputPriceMicrosPerTokenUnit: 3_500_000,
      reason: "reviewed development pricing evidence",
    });
    assert.equal(envelope.failureCode, "gateway_pricing_policy_version_conflict");
    assert.equal(envelope.currentVersion, 4);
    assert.deepEqual(captured?.body, {
      expected_version: 3,
      provider_id: "mock",
      profile_id: "mock-dev",
      model_id: "mock-model",
      currency: "USD",
      input_price_micros_per_token_unit: 1_250_000,
      output_price_micros_per_token_unit: 3_500_000,
      reason: "reviewed development pricing evidence",
    });
    assert.equal(captured?.headers.get("X-RadishMind-Dev-Read-Scopes"), "admin_gateway_pricing:write");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Admin Gateway pricing rejects forbidden and scope-drifted responses", async () => {
  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async () => jsonResponse({ ...successEnvelope(), debug: { invoice: "forbidden" } });
    await assert.rejects(() => readAdminGatewayModelPricing(config, scope), /forbidden field/);

    globalThis.fetch = async () => jsonResponse({ ...successEnvelope(), model_id: "other-model" });
    await assert.rejects(() => readAdminGatewayModelPricing(config, scope), /invalid HTTP 200 envelope/);

    const invalidPolicy = successEnvelope();
    invalidPolicy.policy.policy_digest = "sha256:bad";
    globalThis.fetch = async () => jsonResponse(invalidPolicy);
    await assert.rejects(() => readAdminGatewayModelPricing(config, scope), /invalid HTTP 200 envelope/);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Admin Gateway pricing preserves a verified-boundary denial without inventing context", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => jsonResponse({
    ...failureEnvelope("gateway_pricing_scope_denied", 0),
    tenant_ref: "",
    workspace_id: "",
  }, 403);
  try {
    const envelope = await readAdminGatewayModelPricing(config, scope);
    assert.equal(envelope.failureCode, "gateway_pricing_scope_denied");
    assert.equal(envelope.policy, null);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Admin Gateway pricing accepts zero price but rejects unsafe rates and sensitive reasons", async () => {
  assert.equal(isValidAdminGatewayModelPricingRate(0), true);
  assert.equal(isValidAdminGatewayModelPricingRate(Number.MAX_SAFE_INTEGER), true);
  assert.equal(isValidAdminGatewayModelPricingRate(Number.MAX_SAFE_INTEGER + 1), false);
  assert.equal(isValidAdminGatewayModelPricingRate(-1), false);
  assert.equal(isValidAdminGatewayModelPricingReason("reviewed development evidence"), true);
  assert.equal(isValidAdminGatewayModelPricingReason("secret=do-not-store"), false);
  await assert.rejects(() => putAdminGatewayModelPricing(config, scope, {
    expectedVersion: 3,
    inputPriceMicrosPerTokenUnit: -1,
    outputPriceMicrosPerTokenUnit: 3_000_000,
    reason: "reviewed development evidence",
  }), /update is invalid/);
});

function successEnvelope() {
  return {
    request_id: "admin-gateway-pricing-read-test",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    environment: "test",
    provider_id: "mock",
    profile_id: "mock-dev",
    model_id: "mock-model",
    policy: {
      schema_version: "gateway_model_pricing_policy.v1",
      policy_id: `gmp_${"a".repeat(24)}`,
      record_version: 3,
      tenant_ref: "tenant_demo",
      workspace_id: "workspace_demo",
      environment: "test",
      provider_id: "mock",
      profile_id: "mock-dev",
      model_id: "mock-model",
      currency: "USD",
      token_unit: 1_000_000,
      input_price_micros_per_token_unit: 1_000_000,
      output_price_micros_per_token_unit: 3_000_000,
      policy_digest: `sha256:${"b".repeat(64)}`,
      reason: "reviewed development pricing evidence",
      updated_at: "2026-08-12T01:00:00Z",
      updated_by_actor_ref: "subject_platform_ops",
      request_id: "pricing-policy-put-test",
      audit_ref: "audit_pricing_policy_put_test",
    },
    current_version: 3,
    failure_code: null,
    audit_ref: "audit_admin_gateway_pricing_read_test",
  };
}

function failureEnvelope(failureCode: string, currentVersion: number) {
  return {
    request_id: "admin-gateway-pricing-write-test",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    environment: "test",
    provider_id: "mock",
    profile_id: "mock-dev",
    model_id: "mock-model",
    policy: null,
    current_version: currentVersion,
    failure_code: failureCode,
    audit_ref: "audit_admin_gateway_pricing_write_test",
  };
}

function jsonResponse(value: unknown, status = 200) {
  return new Response(JSON.stringify(value), { status, headers: { "Content-Type": "application/json" } });
}
