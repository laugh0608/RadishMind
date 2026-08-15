import assert from "node:assert/strict";
import test from "node:test";

import {
  EMPTY_GATEWAY_REQUEST_HISTORY_FILTER,
  initialGatewayRequestHistoryState,
  listGatewayRequestHistory,
  readGatewayRequestHistoryDetail,
  type ModelGatewayRequestHistoryConfig,
} from "../src/features/control-plane-read/modelGatewayRequestHistoryConsumer.ts";

const offline: ModelGatewayRequestHistoryConfig = {
  mode: "offline",
  baseUrl: "http://127.0.0.1:7000",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  consumerRef: "consumer_web_dev",
  applicationId: "application_demo",
  subjectRef: "subject_web_dev",
};
const live: ModelGatewayRequestHistoryConfig = { ...offline, mode: "dev_gateway_request_history_http" };

test("Gateway request history stays offline without fetching", async () => {
  let called = false;
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => { called = true; throw new Error("unexpected fetch"); };
  try {
    assert.equal(initialGatewayRequestHistoryState(offline).status, "offline");
    assert.equal((await listGatewayRequestHistory(offline, EMPTY_GATEWAY_REQUEST_HISTORY_FILTER)).status, "offline");
    assert.equal(called, false);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Gateway request history maps scoped summaries, filters, pagination, and caller headers", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async (input, init) => {
    const url = new URL(String(input));
    assert.equal(url.pathname, "/v1/model-gateway/requests");
    assert.equal(url.searchParams.get("workspace_id"), "workspace_demo");
    assert.equal(url.searchParams.get("consumer_ref"), "consumer_web_dev");
    assert.equal(url.searchParams.get("application_id"), "application_demo");
    assert.equal(url.searchParams.get("protocol"), "openai-responses");
    assert.equal(url.searchParams.get("status"), "failed");
    assert.equal(url.searchParams.get("usage_availability"), "not_reported");
    assert.equal(url.searchParams.get("fallback_used"), "true");
    assert.equal(url.searchParams.get("terminal_provider"), "mock-backup");
    assert.equal(url.searchParams.get("terminal_profile"), "backup-dev");
    assert.equal(url.searchParams.get("cursor"), "cursor_previous");
    const headers = new Headers(init?.headers);
    assert.equal(headers.get("X-RadishMind-Dev-Gateway-Tenant"), "tenant_demo");
    assert.equal(headers.get("X-RadishMind-Dev-Gateway-Scopes"), "gateway_requests:read");
    assert.equal(headers.get("X-RadishMind-Dev-Gateway-Application"), "application_demo");
    return jsonResponse({
      request_id: "request_list",
      ...historyEnvelopeScope(),
      requests: [{ ...summaryDocument(), store_mode: "sqlite_dev" }],
      next_cursor: "cursor_next",
      has_more: true,
      failure_code: null,
      failure_summary: "",
      audit_ref: "audit_list",
    });
  };
  try {
    const result = await listGatewayRequestHistory(
      live,
      {
        ...EMPTY_GATEWAY_REQUEST_HISTORY_FILTER,
        protocol: "openai-responses",
        status: "failed",
        usageAvailability: "not_reported",
        fallbackUsed: "true",
        terminalProvider: "mock-backup",
        terminalProfile: "backup-dev",
      },
      "cursor_previous",
    );
    assert.equal(result.status, "ready");
    assert.equal(result.requests[0]?.requestId, "request_gateway_1");
    assert.equal(result.requests[0]?.storeMode, "sqlite_dev");
    assert.equal(result.requests[0]?.providerDurationAvailable, true);
    assert.equal(result.requests[0]?.providerRouteConfigurationId, "gateway-default");
    assert.equal(result.requests[0]?.providerRouteGeneration, 3);
    assert.equal(result.requests[0]?.providerRouteSnapshotDigest, `sha256:${"d".repeat(64)}`);
    assert.equal(result.requests[0]?.costEstimate.availability, "usage_not_reported");
    assert.equal(result.hasMore, true);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Gateway request history preserves stable API failures without offline fallback", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => jsonResponse({
    request_id: "request_denied",
    ...historyEnvelopeScope(),
    requests: [],
    next_cursor: "",
    has_more: false,
    failure_code: "gateway_request_scope_denied",
    failure_summary: "Gateway request history scope is denied.",
    audit_ref: "audit_denied",
  });
  try {
    const result = await listGatewayRequestHistory(live, EMPTY_GATEWAY_REQUEST_HISTORY_FILTER);
    assert.equal(result.status, "failed");
    assert.equal(result.failureCode, "gateway_request_scope_denied");
    assert.deepEqual(result.requests, []);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Gateway request history maps sanitized detail and usage availability", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async (input, init) => {
    const url = new URL(String(input));
    assert.equal(url.pathname, "/v1/model-gateway/requests/request_gateway_1");
    assert.equal(new Headers(init?.headers).get("X-RadishMind-Dev-Gateway-Audit"), "audit_dev_gateway_request_history_detail");
    const summary = summaryDocument();
    const { usage_availability: _usageAvailability, ...detailSummary } = summary;
    return jsonResponse({
      request_id: "request_detail",
      ...historyEnvelopeScope(),
      request: {
        ...detailSummary,
        tenant_ref: "tenant_demo",
        workspace_id: "workspace_demo",
        consumer_ref: "consumer_web_dev",
        application_id: "application_demo",
        subject_ref: "subject_web_dev",
        gateway_duration_ms: 90,
        gateway_duration_available: true,
        usage: {
          availability: "reported",
          source: "gemini_usage_metadata",
          input_tokens: 34,
          output_tokens: 11,
          total_tokens: 45,
        },
        cost_estimate: estimatedCostDocument(17),
      },
      failure_code: null,
      failure_summary: "",
      audit_ref: "audit_detail",
    });
  };
  try {
    const detail = await readGatewayRequestHistoryDetail(live, "request_gateway_1");
    assert.equal(detail.gatewayDurationMs, 90);
    assert.equal(detail.usageAvailability, "reported");
    assert.equal(detail.usageSource, "gemini_usage_metadata");
    assert.equal(detail.totalTokens, 45);
    assert.equal(detail.consumerRef, "consumer_web_dev");
    assert.equal(detail.providerRouteGeneration, 3);
    assert.equal(detail.schemaVersion, "gateway_request_record.v2");
    assert.equal(detail.costEstimate.estimatedCostMicros, 17);
    assert.equal(detail.costEstimate.pricingPolicyVersion, 3);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Gateway request history strictly maps v3 attempt lineage and partial cost coverage", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async (input) => {
    const detailRequest = new URL(String(input)).pathname.includes("request_gateway_v3");
    return detailRequest
      ? jsonResponse({
        request_id: "request_detail_v3",
        ...historyEnvelopeScope(),
        request: v3DetailDocument(),
        failure_code: null,
        failure_summary: "",
        audit_ref: "audit_detail_v3",
      })
      : jsonResponse({
        request_id: "request_list_v3",
        ...historyEnvelopeScope(),
        requests: [v3SummaryDocument()],
        next_cursor: "",
        has_more: false,
        failure_code: null,
        failure_summary: "",
        audit_ref: "audit_list_v3",
      });
  };
  try {
    const listed = await listGatewayRequestHistory(live, EMPTY_GATEWAY_REQUEST_HISTORY_FILTER);
    assert.equal(listed.requests[0]?.schemaVersion, "gateway_request_record.v3");
    assert.equal(listed.requests[0]?.attemptCount, 2);
    assert.equal(listed.requests[0]?.fallbackUsed, true);
    assert.equal(listed.requests[0]?.terminalProfile, "backup-dev");
    assert.equal(listed.requests[0]?.attemptCostSummary?.coverage, "partial");

    const detail = await readGatewayRequestHistoryDetail(live, "request_gateway_v3");
    assert.equal(detail.attemptPhase, "terminal");
    assert.equal(detail.attemptPlan?.executionMode, "sequential_fallback");
    assert.equal(detail.attemptPlan?.targets[0]?.pricingAvailability, "configured");
    assert.equal(detail.attemptPlan?.targets[1]?.pricingAvailability, "not_configured");
    assert.equal(detail.providerAttempts[0]?.failure?.fallbackDisposition, "eligible");
    assert.equal(detail.providerAttempts[1]?.status, "succeeded");
    assert.equal(detail.providerAttempts[1]?.totalTokens, 14);
    assert.equal(detail.terminalAttemptId, "request_gateway_v3.pa2");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Gateway request history derives omitted v3 terminal projection from the terminal attempt", async () => {
  const originalFetch = globalThis.fetch;
  const detail = v3DetailDocument();
  delete detail.terminal_provider;
  delete detail.terminal_profile;
  globalThis.fetch = async () => jsonResponse({
    request_id: "request_detail_v3_raw",
    ...historyEnvelopeScope(),
    request: detail,
    failure_code: null,
    failure_summary: "",
    audit_ref: "audit_detail_v3_raw",
  });
  try {
    const mapped = await readGatewayRequestHistoryDetail(live, "request_gateway_v3");
    assert.equal(mapped.terminalProvider, "mock-backup");
    assert.equal(mapped.terminalProfile, "backup-dev");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Gateway request history rejects a v3 terminal projection that conflicts with the terminal attempt", async () => {
  const originalFetch = globalThis.fetch;
  const detail = v3DetailDocument();
  detail.terminal_provider = "unexpected-provider";
  globalThis.fetch = async () => jsonResponse({
    request_id: "request_detail_v3_conflict",
    ...historyEnvelopeScope(),
    request: detail,
    failure_code: null,
    failure_summary: "",
    audit_ref: "audit_detail_v3_conflict",
  });
  try {
    await assert.rejects(
      () => readGatewayRequestHistoryDetail(live, "request_gateway_v3"),
      /Gateway request detail route failed/u,
    );
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Gateway request history rejects inconsistent v3 terminal lineage", async () => {
  const originalFetch = globalThis.fetch;
  const detail = v3DetailDocument();
  detail.terminal_attempt_id = "request_gateway_v3.pa1";
  globalThis.fetch = async () => jsonResponse({
    request_id: "request_detail_v3_invalid",
    ...historyEnvelopeScope(),
    request: detail,
    failure_code: null,
    failure_summary: "",
    audit_ref: "audit_detail_v3_invalid",
  });
  try {
    await assert.rejects(
      () => readGatewayRequestHistoryDetail(live, "request_gateway_v3"),
      /Gateway request detail route failed/u,
    );
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Gateway request history rejects forbidden fields at any response depth", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => jsonResponse({
    request_id: "request_list",
    ...historyEnvelopeScope(),
    requests: [{ ...summaryDocument(), debug: { prompt: "must-not-cross-consumer" } }],
    next_cursor: "",
    has_more: false,
    failure_code: null,
    failure_summary: "",
    audit_ref: "audit_list",
  });
  try {
    await assert.rejects(
      () => listGatewayRequestHistory(live, EMPTY_GATEWAY_REQUEST_HISTORY_FILTER),
      /forbidden field response\.requests\[0\]\.debug\.prompt/,
    );
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Gateway request history rejects partial Provider route lineage", async () => {
  const originalFetch = globalThis.fetch;
  const { provider_route_snapshot_digest: _snapshotDigest, ...partial } = summaryDocument();
  globalThis.fetch = async () => jsonResponse({
    request_id: "request_list",
    ...historyEnvelopeScope(),
    requests: [partial],
    next_cursor: "",
    has_more: false,
    failure_code: null,
    failure_summary: "",
    audit_ref: "audit_list",
  });
  try {
    await assert.rejects(
      () => listGatewayRequestHistory(live, EMPTY_GATEWAY_REQUEST_HISTORY_FILTER),
      /Gateway request history route failed/,
    );
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Gateway request history rejects inconsistent reported usage", async () => {
  const originalFetch = globalThis.fetch;
  const summary = summaryDocument();
  globalThis.fetch = async () => jsonResponse({
    request_id: "request_list",
    ...historyEnvelopeScope(),
    requests: [{
      ...summary,
      usage_availability: "reported",
      usage: {
        availability: "reported",
        source: "openai_compatible_usage",
        input_tokens: 10,
        output_tokens: 4,
        total_tokens: 99,
      },
    }],
    next_cursor: "",
    has_more: false,
    failure_code: null,
    failure_summary: "",
    audit_ref: "audit_list",
  });
  try {
    await assert.rejects(
      () => listGatewayRequestHistory(live, EMPTY_GATEWAY_REQUEST_HISTORY_FILTER),
      /Gateway request history route failed/,
    );
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Gateway request history accepts all six cost states and v1 only as legacy evidence", async () => {
  const originalFetch = globalThis.fetch;
  const states = [
    summaryWithCost("request_estimated", estimatedCostDocument(17), "reported"),
    summaryWithCost("request_usage_missing", unavailableCostDocument("usage_not_reported", "provider_usage_not_reported"), "not_reported"),
    summaryWithCost("request_price_missing", unavailableCostDocument("price_not_configured", "pricing_policy_not_configured"), "reported"),
    summaryWithCost("request_price_unavailable", unavailableCostDocument("price_unavailable", "pricing_snapshot_unavailable"), "reported"),
    summaryWithCost("request_not_applicable", unavailableCostDocument("not_applicable", "provider_not_attempted"), "not_applicable"),
    {
      ...summaryWithCost("request_legacy", unavailableCostDocument("legacy_not_captured", "legacy_record_without_cost_snapshot"), "reported"),
      schema_version: "gateway_request_record.v1",
    },
  ];
  globalThis.fetch = async () => jsonResponse({
    request_id: "request_cost_states",
    ...historyEnvelopeScope(),
    requests: states,
    next_cursor: "",
    has_more: false,
    failure_code: null,
    failure_summary: "",
    audit_ref: "audit_cost_states",
  });
  try {
    const result = await listGatewayRequestHistory(live, EMPTY_GATEWAY_REQUEST_HISTORY_FILTER);
    assert.deepEqual(result.requests.map((request) => request.costEstimate.availability), [
      "estimated", "usage_not_reported", "price_not_configured", "price_unavailable", "not_applicable", "legacy_not_captured",
    ]);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Gateway request history rejects benign extra cost fields and scope drift", async () => {
  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async () => jsonResponse({
      request_id: "request_extra_cost",
      ...historyEnvelopeScope(),
      requests: [{
        ...summaryDocument(),
        cost_estimate: { ...unavailableCostDocument("usage_not_reported", "provider_usage_not_reported"), debug_hint: "unexpected" },
      }],
      next_cursor: "",
      has_more: false,
      failure_code: null,
      failure_summary: "",
      audit_ref: "audit_extra_cost",
    });
    await assert.rejects(() => listGatewayRequestHistory(live, EMPTY_GATEWAY_REQUEST_HISTORY_FILTER), /route failed/);

    globalThis.fetch = async () => jsonResponse({
      request_id: "request_scope_drift",
      ...historyEnvelopeScope(),
      workspace_id: "workspace_other",
      requests: [],
      next_cursor: "",
      has_more: false,
      failure_code: null,
      failure_summary: "",
      audit_ref: "audit_scope_drift",
    });
    await assert.rejects(() => listGatewayRequestHistory(live, EMPTY_GATEWAY_REQUEST_HISTORY_FILTER), /route failed/);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

function summaryDocument() {
  return {
    schema_version: "gateway_request_record.v2",
    record_version: 3,
    store_mode: "postgres_dev_test",
    request_id: "request_gateway_1",
    audit_ref: "audit_gateway_1",
    route: "POST /v1/responses",
    protocol: "openai-responses",
    stream: false,
    status: "failed",
    started_at: "2026-07-12T04:00:00Z",
    completed_at: "2026-07-12T04:00:00.120Z",
    duration_ms: 120,
    provider_duration_ms: 70,
    provider_duration_available: true,
    selection_source: "explicit_model",
    selected_provider: "mock",
    selected_profile: "mock-dev",
    selected_model: "mock-model",
    provider_route_configuration_id: "gateway-default",
    provider_route_generation: 3,
    provider_route_snapshot_digest: `sha256:${"d".repeat(64)}`,
    http_status_code: 502,
    failure_code: "GATEWAY_PROVIDER_FAILED",
    failure_boundary: "provider",
    usage_availability: "not_reported",
    usage: {
      availability: "not_reported",
      source: "",
      input_tokens: 0,
      output_tokens: 0,
      total_tokens: 0,
    },
    cost_estimate: unavailableCostDocument("usage_not_reported", "provider_usage_not_reported"),
    attempt_count: 0,
    fallback_allowed: false,
    fallback_used: false,
    stale_started: false,
  };
}

function estimatedCostDocument(estimatedCostMicros: number) {
  return {
    schema_version: "gateway_request_cost_estimate.v1",
    availability: "estimated",
    reason: "",
    currency: "USD",
    estimated_cost_micros: estimatedCostMicros,
    token_unit: 1_000_000,
    input_price_micros_per_token_unit: 1_000_000,
    output_price_micros_per_token_unit: 3_000_000,
    pricing_policy_id: `gmp_${"a".repeat(24)}`,
    pricing_policy_version: 3,
    pricing_policy_digest: `sha256:${"c".repeat(64)}`,
    rounding_mode: "half_up_to_currency_micro",
  };
}

function v3SummaryDocument() {
  return {
    ...summaryDocument(),
    schema_version: "gateway_request_record.v3",
    request_id: "request_gateway_v3",
    status: "succeeded",
    http_status_code: 200,
    failure_code: "",
    failure_boundary: "",
    selected_provider: "mock-primary",
    selected_profile: "primary-dev",
    attempt_count: 2,
    fallback_allowed: true,
    fallback_used: true,
    terminal_provider: "mock-backup",
    terminal_profile: "backup-dev",
    provider_attempt_cost_summary: attemptCostSummaryDocument(),
  };
}

function v3DetailDocument() {
  const { usage_availability: _usageAvailability, ...summary } = v3SummaryDocument();
  return {
    ...summary,
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    consumer_ref: "consumer_web_dev",
    application_id: "application_demo",
    subject_ref: "subject_web_dev",
    gateway_duration_ms: 120,
    gateway_duration_available: true,
    provider_attempt_plan: {
      schema_version: "gateway_provider_attempt_plan.v1",
      root_request_id: "request_gateway_v3",
      route: "POST /v1/responses",
      protocol: "openai-responses",
      requested_model: "mock-model",
      configuration_id: "gateway-default",
      route_generation: 3,
      route_snapshot_digest: `sha256:${"d".repeat(64)}`,
      execution_mode: "sequential_fallback",
      fallback_mode: "allow_configured",
      fallback_allowed: true,
      max_attempts: 2,
      targets: [
        attemptPlanTarget(1, "primary", "mock-primary", "primary-dev", "primary-upstream", "e"),
        attemptPlanTarget(2, "backup", "mock-backup", "backup-dev", "backup-upstream", "f"),
      ],
    },
    provider_attempt_phase: "terminal",
    terminal_attempt_id: "request_gateway_v3.pa2",
    provider_attempts: [
      {
        ...attemptRecord(1, "primary", "mock-primary", "primary-dev", "primary-upstream", "e"),
        status: "failed",
        completed_at: "2026-07-12T04:00:00.050Z",
        duration_ms: 50,
        quota_admission_id: "quota-primary",
        failure: {
          schema_version: "gateway_provider_attempt_failure.v1",
          failure_class: "provider_temporarily_unavailable",
          fallback_disposition: "eligible",
          provider_response_started: false,
          outcome: "failed",
          code: "PROVIDER_TEMPORARILY_UNAVAILABLE",
          http_status_class: "5xx",
        },
        failure_boundary: "provider",
        usage: unavailableUsageDocument(),
        cost_estimate: unavailableCostDocument("usage_not_reported", "provider_usage_not_reported"),
      },
      {
        ...attemptRecord(2, "backup", "mock-backup", "backup-dev", "backup-upstream", "f"),
        status: "succeeded",
        completed_at: "2026-07-12T04:00:00.100Z",
        duration_ms: 40,
        quota_admission_id: "quota-backup",
        usage: {
          availability: "reported",
          source: "openai_compatible_usage",
          input_tokens: 10,
          output_tokens: 4,
          total_tokens: 14,
        },
        cost_estimate: estimatedCostDocument(17),
      },
    ],
  };
}

function attemptPlanTarget(
  ordinal: number,
  profileId: string,
  providerId: string,
  runtimeProfile: string,
  upstreamModel: string,
  digestCharacter: string,
) {
  return {
    attempt_id: `request_gateway_v3.pa${ordinal}`,
    ordinal,
    provider_profile_id: profileId,
    provider_id: providerId,
    runtime_profile: runtimeProfile,
    selected_model: "mock-model",
    upstream_model: upstreamModel,
    inventory_digest: `sha256:${digestCharacter.repeat(64)}`,
    pricing_snapshot: ordinal === 1 ? {
      availability: "configured",
      currency: "USD",
      token_unit: 1_000_000,
      input_price_micros_per_token_unit: 1_000_000,
      output_price_micros_per_token_unit: 3_000_000,
      pricing_policy_id: `gmp_${"a".repeat(24)}`,
      pricing_policy_version: 3,
      pricing_policy_digest: `sha256:${"c".repeat(64)}`,
      integrity_digest: "b".repeat(64),
    } : { availability: "not_configured", reason: "pricing_policy_not_configured" },
  };
}

function attemptRecord(
  ordinal: number,
  profileId: string,
  providerId: string,
  runtimeProfile: string,
  upstreamModel: string,
  digestCharacter: string,
) {
  return {
    schema_version: "gateway_provider_attempt_record.v1",
    attempt_id: `request_gateway_v3.pa${ordinal}`,
    ordinal,
    configured_profile_id: profileId,
    provider_id: providerId,
    runtime_profile: runtimeProfile,
    selected_model: "mock-model",
    upstream_model: upstreamModel,
    route_generation: 3,
    route_snapshot_digest: `sha256:${"d".repeat(64)}`,
    inventory_digest: `sha256:${digestCharacter.repeat(64)}`,
    started_at: `2026-07-12T04:00:00.0${ordinal}0Z`,
  };
}

function attemptCostSummaryDocument() {
  return {
    schema_version: "gateway_request_attempt_cost_summary.v1",
    known_cost_micros: 17,
    coverage: "partial",
    estimated_attempt_count: 1,
    unknown_attempt_count: 1,
  };
}

function unavailableUsageDocument() {
  return { availability: "not_reported", source: "", input_tokens: 0, output_tokens: 0, total_tokens: 0 };
}

function unavailableCostDocument(availability: string, reason: string) {
  return { schema_version: "gateway_request_cost_estimate.v1", availability, reason };
}

function summaryWithCost(requestId: string, costEstimate: Record<string, unknown>, usageAvailability: "reported" | "not_reported" | "not_applicable") {
  const usage = usageAvailability === "reported"
    ? { availability: "reported", source: "openai_compatible_usage", input_tokens: 10, output_tokens: 4, total_tokens: 14 }
    : { availability: usageAvailability, source: "", input_tokens: 0, output_tokens: 0, total_tokens: 0 };
  return { ...summaryDocument(), request_id: requestId, usage_availability: usageAvailability, usage, cost_estimate: costEstimate };
}

function historyEnvelopeScope() {
  return {
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    consumer_ref: "consumer_web_dev",
    application_id: "application_demo",
  };
}

function jsonResponse(body: unknown): Response {
  return new Response(JSON.stringify(body), { status: 200, headers: { "Content-Type": "application/json" } });
}
