import assert from "node:assert/strict";
import test from "node:test";

import { compareWorkflowRuns } from "../src/features/control-plane-read/workflowRunComparisonConsumer.ts";
import type { WorkflowExecutorConsumerConfig } from "../src/features/control-plane-read/workflowExecutorConsumer.ts";

const offline: WorkflowExecutorConsumerConfig = { mode: "disabled", baseUrl: "http://127.0.0.1:7000", workspaceId: "workspace_demo", tenantRef: "tenant_demo", subjectRef: "subject_demo" };
const live: WorkflowExecutorConsumerConfig = { ...offline, mode: "dev_workflow_executor_http" };

function run(runId: string, status: "succeeded" | "failed") {
  return { run_id: runId, schema_version: "workflow_run_record.v1", draft_id: "draft_real", draft_version: 2, status, failure_code: status === "failed" ? "workflow_run_gateway_failed" : "", failure_boundary: status === "failed" ? "gateway" : "", gateway_failure_category: status === "failed" ? "timeout" : "none", selected_provider: "mock", selected_profile: "default", selected_model: "mock", duration_ms: status === "failed" ? 120 : 100, stale_running: false, request_id: `request_${runId}`, audit_ref: `audit_${runId}`, side_effects: { provider_calls: 1, tool_calls: 0, confirmation_calls: 0, business_writes: 0, replay_writes: 0 } };
}

function envelope(overrides: Record<string, unknown> = {}) {
  return { request_id: "request_compare", workspace_id: "workspace_demo", application_id: "app_demo", failure_code: null, failure_summary: "", audit_ref: "audit_compare", comparison: { schema_version: "workflow_run_comparison.v1", classification: "regression", comparison_state: "comparable", baseline: run("run_base", "succeeded"), candidate: run("run_candidate", "failed"), draft_changed: false, provider_changed: false, model_changed: false, status_changed: true, failure_changed: true, duration_delta_ms: 20, provider_call_delta: 0, nodes: [{ node_id: "node_model", node_type: "llm", change: "changed", baseline_status: "succeeded", candidate_status: "failed", baseline_duration_ms: 50, candidate_duration_ms: 70, duration_delta_ms: 20 }], findings: [{ code: "status_regressed", severity: "review_required" }], recommended_review_action: "check_gateway_capacity" }, ...overrides };
}

test("workflow run comparison rejects offline and same-run selection without fetching", async () => {
  let calls = 0; const originalFetch = globalThis.fetch; globalThis.fetch = async () => { calls++; throw new Error("unexpected fetch"); };
  try {
    await assert.rejects(() => compareWorkflowRuns("app_demo", "run_a", "run_b", offline), /offline mode/);
    await assert.rejects(() => compareWorkflowRuns("app_demo", "run_a", "run_a", live), /different workflow runs/);
    assert.equal(calls, 0);
  } finally { globalThis.fetch = originalFetch; }
});

test("workflow run comparison maps scoped nested resource and zero side effects", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async (input, init) => {
    const url = new URL(String(input));
    assert.equal(url.pathname, "/v1/user-workspace/workflow-runs/run_candidate/comparison");
    assert.equal(url.searchParams.get("baseline_run_id"), "run_base");
    assert.equal(url.searchParams.get("workspace_id"), "workspace_demo");
    assert.equal(new Headers(init?.headers).get("X-RadishMind-Dev-Read-Scopes"), "workflow_runs:read");
    return new Response(JSON.stringify(envelope()), { status: 200 });
  };
  try {
    const result = await compareWorkflowRuns("app_demo", "run_base", "run_candidate", live);
    assert.equal(result.classification, "regression"); assert.equal(result.nodes[0]?.change, "changed"); assert.equal(result.candidate.sideEffects.toolCalls, 0);
  } finally { globalThis.fetch = originalFetch; }
});

test("workflow run comparison rejects forbidden fields and side effects", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => new Response(JSON.stringify(envelope({ provider_raw_envelope: { secret: true } })), { status: 200 });
  try { await assert.rejects(() => compareWorkflowRuns("app_demo", "run_base", "run_candidate", live), /route failed/); }
  finally { globalThis.fetch = originalFetch; }

  globalThis.fetch = async () => { const body = envelope(); (body.comparison.candidate.side_effects as { tool_calls: number }).tool_calls = 1; return new Response(JSON.stringify(body), { status: 200 }); };
  try { await assert.rejects(() => compareWorkflowRuns("app_demo", "run_base", "run_candidate", live), /forbidden side effect/); }
  finally { globalThis.fetch = originalFetch; }
});
