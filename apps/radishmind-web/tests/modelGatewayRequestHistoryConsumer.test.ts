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
    assert.equal(url.searchParams.get("cursor"), "cursor_previous");
    const headers = new Headers(init?.headers);
    assert.equal(headers.get("X-RadishMind-Dev-Gateway-Tenant"), "tenant_demo");
    assert.equal(headers.get("X-RadishMind-Dev-Gateway-Scopes"), "gateway_requests:read");
    assert.equal(headers.get("X-RadishMind-Dev-Gateway-Application"), "application_demo");
    return jsonResponse({
      request_id: "request_list",
      requests: [summaryDocument()],
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
      { ...EMPTY_GATEWAY_REQUEST_HISTORY_FILTER, protocol: "openai-responses", status: "failed", usageAvailability: "not_reported" },
      "cursor_previous",
    );
    assert.equal(result.status, "ready");
    assert.equal(result.requests[0]?.requestId, "request_gateway_1");
    assert.equal(result.requests[0]?.providerDurationAvailable, true);
    assert.equal(result.hasMore, true);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Gateway request history preserves stable API failures without offline fallback", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => jsonResponse({
    request_id: "request_denied",
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
      request: {
        ...detailSummary,
        tenant_ref: "tenant_demo",
        workspace_id: "workspace_demo",
        consumer_ref: "consumer_web_dev",
        application_id: "application_demo",
        subject_ref: "subject_web_dev",
        gateway_duration_ms: 90,
        gateway_duration_available: true,
        usage: { availability: "not_reported", source: "", input_tokens: 0, output_tokens: 0, total_tokens: 0 },
      },
      failure_code: null,
      failure_summary: "",
      audit_ref: "audit_detail",
    });
  };
  try {
    const detail = await readGatewayRequestHistoryDetail(live, "request_gateway_1");
    assert.equal(detail.gatewayDurationMs, 90);
    assert.equal(detail.usageAvailability, "not_reported");
    assert.equal(detail.totalTokens, 0);
    assert.equal(detail.consumerRef, "consumer_web_dev");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Gateway request history rejects forbidden fields at any response depth", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => jsonResponse({
    request_id: "request_list",
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

function summaryDocument() {
  return {
    schema_version: "gateway_request_record.v1",
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
    http_status_code: 502,
    failure_code: "GATEWAY_PROVIDER_FAILED",
    failure_boundary: "provider",
    usage_availability: "not_reported",
    stale_started: false,
  };
}

function jsonResponse(body: unknown): Response {
  return new Response(JSON.stringify(body), { status: 200, headers: { "Content-Type": "application/json" } });
}
