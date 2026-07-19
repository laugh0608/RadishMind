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
  return { request_id: "request_compare", workspace_id: "workspace_demo", application_id: "app_demo", failure_code: null, failure_summary: "", audit_ref: "audit_compare", comparison: { schema_version: "workflow_run_comparison.v1", classification: "regression", comparison_state: "comparable", baseline: run("run_base", "succeeded"), candidate: run("run_candidate", "failed"), draft_changed: false, execution_source_changed: false, provider_changed: false, model_changed: false, status_changed: true, failure_changed: true, duration_delta_ms: 20, provider_call_delta: 0, nodes: [{ node_id: "node_model", node_type: "llm", change: "changed", baseline_status: "succeeded", candidate_status: "failed", baseline_duration_ms: 50, candidate_duration_ms: 70, duration_delta_ms: 20 }], findings: [{ code: "status_regressed", severity: "review_required" }], recommended_review_action: "check_gateway_capacity" }, ...overrides };
}

const digest = `sha256:${"a".repeat(64)}`;
function ragEnvelope() {
  const body = envelope();
  body.comparison.schema_version = "workflow_run_comparison.v2" as "workflow_run_comparison.v1";
  body.comparison.classification = "unchanged";
  for (const value of [body.comparison.baseline, body.comparison.candidate]) {
    value.schema_version = "workflow_run_record.v3";
    (value.side_effects as Record<string, number>).retrieval_calls = 1;
  }
  return {
    ...body,
    comparison: {
      ...body.comparison,
      retrieval: {
        run_profile: "workflow_rag_retrieval.v1", snapshot_id: "rags_aaaaaaaaaaaaaaaa", snapshot_version: 1,
        snapshot_digest: digest, rag_ref: "workflow.rag.official.v1", profile_id: "workflow.rag.lexical-ngram-dev.v1",
        profile_version: 1, profile_digest: digest, query_digest: digest, query_bytes: 27, retrieval_node_id: "node_retrieval",
        baseline_attempt_status: "succeeded", candidate_attempt_status: "succeeded", baseline_candidate_count: 2,
        candidate_candidate_count: 2, candidate_count_delta: 0, baseline_selected_count: 1, candidate_selected_count: 1,
        baseline_citation_count: 1, candidate_citation_count: 1, context_bytes_delta: 0, latency_delta_ms: -2,
        evidence_changed: false, ranking_changed: false, citation_changed: false, citation_added_refs: [], citation_removed_refs: [],
        fragments: [{ fragment_ref: "official_guide", content_digest: digest, source_type: "manual", is_official: true, baseline_rank: 1, candidate_rank: 1, change: "unchanged" }],
      },
    },
  };
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

test("workflow run comparison maps metadata-only RAG v2 evidence", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => new Response(JSON.stringify(ragEnvelope()), { status: 200 });
  try {
    const result = await compareWorkflowRuns("app_demo", "run_base", "run_candidate", live);
    assert.equal(result.schemaVersion, "workflow_run_comparison.v2");
    assert.equal(result.retrieval?.runProfile, "workflow_rag_retrieval.v1");
    assert.equal(result.retrieval?.fragments[0]?.fragmentRef, "official_guide");
    assert.equal(result.baseline.sideEffects.retrievalCalls, 1);
  } finally { globalThis.fetch = originalFetch; }
});

test("workflow run comparison maps the application RAG v4 review profile", async () => {
  const originalFetch = globalThis.fetch;
  const body = ragEnvelope();
  for (const value of [body.comparison.baseline, body.comparison.candidate]) {
    value.schema_version = "workflow_run_record.v4";
    value.draft_id = "";
    value.draft_version = 0;
    Object.assign(value, { execution_kind: "application_rag_invocation", execution_source_kind: "application_configuration_draft", execution_source_id: "appdraft_aaaaaaaaaaaaaaaa", execution_source_version: 2 });
  }
  body.comparison.schema_version = "workflow_run_comparison.v3" as "workflow_run_comparison.v1";
  body.comparison.retrieval.run_profile = "workflow_rag_application_invocation.v1";
  body.comparison.retrieval.retrieval_node_id = "application_rag_retrieval";
  const authority = (suffix: "a" | "b") => ({
    assignment_id: `wragra_${suffix.repeat(16)}`, assignment_version: suffix === "a" ? 1 : 2, assignment_digest: digest,
    publish_candidate_id: `candidate_${suffix}`, publish_review_version: 1, draft_id: "appdraft_aaaaaaaaaaaaaaaa", draft_version: 2, draft_digest: digest,
    binding_ref: { binding_id: `wragb_${suffix.repeat(16)}`, binding_version: 1, binding_digest: digest },
    snapshot_id: `rags_${suffix.repeat(16)}`, snapshot_version: suffix === "a" ? 1 : 2, snapshot_digest: digest,
    rag_ref: "workflow.rag.official.v1", profile_id: "workflow.rag.lexical-ngram-dev.v1", profile_version: 1, profile_digest: digest,
    configured_protocol: "responses", configured_model: "mock-rag",
  });
  const retrieval = body.comparison.retrieval as Record<string, unknown>;
  for (const key of ["snapshot_id", "snapshot_version", "snapshot_digest", "rag_ref", "profile_id", "profile_version", "profile_digest"]) delete retrieval[key];
  Object.assign(retrieval, { baseline_authority: authority("a"), candidate_authority: authority("b"), authority_changed: true });
  globalThis.fetch = async () => new Response(JSON.stringify(body), { status: 200 });
  try {
    const result = await compareWorkflowRuns("app_demo", "run_base", "run_candidate", live);
    assert.equal(result.retrieval?.runProfile, "workflow_rag_application_invocation.v1");
    assert.equal(result.schemaVersion, "workflow_run_comparison.v3");
    assert.equal(result.retrieval?.candidateAuthority?.snapshotVersion, 2);
    assert.equal(result.retrieval?.authorityChanged, true);
    assert.equal(result.baseline.schemaVersion, "workflow_run_record.v4");
    assert.equal(result.baseline.executionSourceId, "appdraft_aaaaaaaaaaaaaaaa");
    assert.equal(result.baseline.sideEffects.retrievalCalls, 1);
  } finally { globalThis.fetch = originalFetch; }
});

test("workflow run comparison rejects RAG leakage and unknown citation references", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => { const body = ragEnvelope(); (body.comparison.retrieval as Record<string, unknown>).raw_query = "private"; return new Response(JSON.stringify(body), { status: 200 }); };
  try { await assert.rejects(() => compareWorkflowRuns("app_demo", "run_base", "run_candidate", live), /route failed/); }
  finally { globalThis.fetch = originalFetch; }
  globalThis.fetch = async () => { const body = ragEnvelope(); (body.comparison.retrieval as { citation_added_refs: string[] }).citation_added_refs = ["not_selected"]; return new Response(JSON.stringify(body), { status: 200 }); };
  try { await assert.rejects(() => compareWorkflowRuns("app_demo", "run_base", "run_candidate", live), /route failed/); }
  finally { globalThis.fetch = originalFetch; }
});

test("workflow run comparison rejects unknown fields and response scope drift", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => { const body = ragEnvelope(); (body.comparison.retrieval as Record<string, unknown>).extra = "unknown"; return new Response(JSON.stringify(body), { status: 200 }); };
  try { await assert.rejects(() => compareWorkflowRuns("app_demo", "run_base", "run_candidate", live), /route failed/); }
  finally { globalThis.fetch = originalFetch; }
  globalThis.fetch = async () => new Response(JSON.stringify({ ...ragEnvelope(), application_id: "app_other" }), { status: 200 });
  try { await assert.rejects(() => compareWorkflowRuns("app_demo", "run_base", "run_candidate", live), /scope drifted/); }
  finally { globalThis.fetch = originalFetch; }
});
