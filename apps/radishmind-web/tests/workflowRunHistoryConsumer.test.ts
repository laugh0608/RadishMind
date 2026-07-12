import assert from "node:assert/strict";
import test from "node:test";

import {
  EMPTY_WORKFLOW_RUN_HISTORY_FILTER,
  initialWorkflowRunHistoryState,
  listWorkflowRunHistory,
} from "../src/features/control-plane-read/workflowRunHistoryConsumer.ts";
import type { WorkflowExecutorConsumerConfig } from "../src/features/control-plane-read/workflowExecutorConsumer.ts";

const offline: WorkflowExecutorConsumerConfig = { mode: "disabled", baseUrl: "http://127.0.0.1:7000", workspaceId: "workspace_demo", tenantRef: "tenant_demo", subjectRef: "subject_demo" };
const live: WorkflowExecutorConsumerConfig = { ...offline, mode: "dev_workflow_executor_http" };

test("workflow run history stays offline by default without fetching", async () => {
  let called = false;
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => { called = true; throw new Error("unexpected fetch"); };
  try {
    assert.equal(initialWorkflowRunHistoryState(offline).status, "offline");
    assert.equal((await listWorkflowRunHistory("app_demo", offline, EMPTY_WORKFLOW_RUN_HISTORY_FILTER)).status, "offline");
    assert.equal(called, false);
  } finally { globalThis.fetch = originalFetch; }
});

test("workflow run history maps scoped page and preserves zero forbidden side effects", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async (input) => {
    const url = String(input);
    assert.match(url, /\/v1\/user-workspace\/workflow-runs\?/);
    assert.doesNotMatch(url, /\/v1\/user-workspace\/runs/);
    return new Response(JSON.stringify({
      request_id: "request_list", workspace_id: "workspace_demo", application_id: "app_demo",
      runs: [{ schema_version: "workflow_run_record.v0", record_version: 2, run_id: "run_real", draft_id: "draft_real", draft_version: 1, workspace_id: "workspace_demo", application_id: "app_demo", status: "succeeded", failure_code: "", started_at: "2026-07-11T10:00:00Z", completed_at: "2026-07-11T10:00:01Z", duration_ms: 1000, selected_provider: "mock", selected_profile: "", selected_model: "mock", request_id: "request_run", audit_ref: "audit_run", stale_running: false, side_effects: { provider_calls: 1, tool_calls: 0, confirmation_calls: 0, business_writes: 0, replay_writes: 0 } }],
      next_cursor: "cursor_next", has_more: true, failure_code: null, failure_summary: "", audit_ref: "audit_list",
    }), { status: 200, headers: { "Content-Type": "application/json" } });
  };
  try {
    const result = await listWorkflowRunHistory("app_demo", live, EMPTY_WORKFLOW_RUN_HISTORY_FILTER);
    assert.equal(result.status, "ready"); assert.equal(result.runs[0]?.runId, "run_real"); assert.equal(result.runs[0]?.sideEffects.toolCalls, 0); assert.equal(result.hasMore, true);
  } finally { globalThis.fetch = originalFetch; }
});

test("workflow run history rejects forbidden side effect counts", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => new Response(JSON.stringify({ request_id: "request", workspace_id: "workspace_demo", application_id: "app_demo", runs: [{ schema_version: "workflow_run_record.v0", record_version: 2, run_id: "run_bad", draft_id: "draft", draft_version: 1, workspace_id: "workspace_demo", application_id: "app_demo", status: "succeeded", failure_code: "", started_at: "2026-07-11T10:00:00Z", completed_at: "2026-07-11T10:00:01Z", duration_ms: 1000, selected_provider: "mock", selected_profile: "", selected_model: "mock", request_id: "request", audit_ref: "audit", stale_running: false, side_effects: { provider_calls: 1, tool_calls: 1, confirmation_calls: 0, business_writes: 0, replay_writes: 0 } }], next_cursor: "", has_more: false, failure_code: null, failure_summary: "", audit_ref: "audit" }), { status: 200 });
  try { await assert.rejects(() => listWorkflowRunHistory("app_demo", live, EMPTY_WORKFLOW_RUN_HISTORY_FILTER), /forbidden side effect/); }
  finally { globalThis.fetch = originalFetch; }
});

test("workflow run history maps v1 diagnostics and sends exact failure filters", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async (input) => {
    const url = new URL(String(input));
    assert.equal(url.searchParams.get("failure_code"), "workflow_run_gateway_failed");
    assert.equal(url.searchParams.get("failure_boundary"), "gateway");
    assert.equal(url.searchParams.get("provider"), "mock");
    assert.equal(url.searchParams.get("model"), "mock-model");
    assert.equal(url.searchParams.get("stale_running"), "true");
    return new Response(JSON.stringify({
      request_id: "request_diag", workspace_id: "workspace_demo", application_id: "app_demo",
      runs: [{ schema_version: "workflow_run_record.v1", record_version: 4, run_id: "run_diag", draft_id: "draft_diag", draft_version: 2, workspace_id: "workspace_demo", application_id: "app_demo", status: "failed", failure_code: "workflow_run_gateway_failed", failure_boundary: "gateway", failed_node_id: "node_model", last_completed_node_id: "node_prompt", gateway_failure_category: "timeout", recommended_review_action: "check_gateway_capacity", started_at: "2026-07-11T10:00:00Z", completed_at: "2026-07-11T10:00:01Z", duration_ms: 1000, selected_provider: "mock", selected_profile: "", selected_model: "mock-model", request_id: "request_run", audit_ref: "audit_run", stale_running: true, side_effects: { provider_calls: 1, tool_calls: 0, confirmation_calls: 0, business_writes: 0, replay_writes: 0 } }],
      next_cursor: "", has_more: false, failure_code: null, failure_summary: "", audit_ref: "audit_diag",
    }), { status: 200 });
  };
  try {
    const result = await listWorkflowRunHistory("app_demo", live, { ...EMPTY_WORKFLOW_RUN_HISTORY_FILTER, failureCode: "workflow_run_gateway_failed", failureBoundary: "gateway", provider: "mock", model: "mock-model", staleRunning: "true" });
    assert.equal(result.runs[0]?.failureBoundary, "gateway");
    assert.equal(result.runs[0]?.gatewayFailureCategory, "timeout");
    assert.equal(result.runs[0]?.failedNodeId, "node_model");
  } finally { globalThis.fetch = originalFetch; }
});
