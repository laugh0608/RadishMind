import assert from "node:assert/strict";
import test from "node:test";

import {
  EMPTY_WORKFLOW_RUN_HISTORY_FILTER,
  isWorkflowRunComparisonCompatible,
  isWorkflowRunComparisonEligible,
  listWorkflowRunHistory,
  readWorkflowRunHistoryDetail,
} from "../src/features/control-plane-read/workflowRunHistoryConsumer.ts";
import type { WorkflowExecutorConsumerConfig } from "../src/features/control-plane-read/workflowExecutorConsumer.ts";

const digestA = `sha256:${"a".repeat(64)}`;
const digestB = `sha256:${"b".repeat(64)}`;
const digestC = `sha256:${"c".repeat(64)}`;
const live: WorkflowExecutorConsumerConfig = {
  mode: "dev_workflow_executor_http", baseUrl: "http://platform.test", workspaceId: "workspace_demo",
  tenantRef: "tenant_demo", subjectRef: "subject_demo_user",
};

test("Run History maps the complete metadata-only v3 summary", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => jsonResponse(listEnvelope());
  try {
    const state = await listWorkflowRunHistory("app_flow_copilot", live, EMPTY_WORKFLOW_RUN_HISTORY_FILTER);
    const run = state.runs[0];
    assert.equal(state.status, "ready");
    assert.equal(run?.schemaVersion, "workflow_run_record.v3");
    assert.equal(run?.snapshotVersion, 2);
    assert.equal(run?.retrievalProfileId, "workflow.rag.lexical-ngram-dev.v1");
    assert.equal(run?.selectedFragments[0]?.fragmentRef, "official_guide");
    assert.deepEqual(run?.citationRefs, ["official_guide"]);
    assert.equal(run?.sideEffects.retrievalCalls, 1);
    assert.equal(run?.sideEffects.providerCalls, 1);
    assert.equal(JSON.stringify(run).includes("private fragment body"), false);
  } finally { globalThis.fetch = originalFetch; }
});

test("authorized v3 detail resolves a bounded preview through the exact history route", async () => {
  const originalFetch = globalThis.fetch;
  const requests: Array<{ url: string; headers: Headers }> = [];
  globalThis.fetch = async (input, init) => {
    requests.push({ url: String(input), headers: new Headers(init?.headers) });
    return requests.length === 1 ? jsonResponse(listEnvelope()) : jsonResponse(detailEnvelope("精确不可变快照片段预览"));
  };
  try {
    const history = await listWorkflowRunHistory("app_flow_copilot", live, EMPTY_WORKFLOW_RUN_HISTORY_FILTER);
    const detail = await readWorkflowRunHistoryDetail(history.runs[0]!, "app_flow_copilot", live, true);
    assert.equal(detail?.schemaVersion, "workflow_run_record.v3");
    assert.equal(detail?.retrievalFragmentPreviews[0]?.preview, "精确不可变快照片段预览");
    assert.match(requests[1]!.url, /include_retrieval_fragment_previews=true/u);
    assert.equal(requests[1]!.headers.get("X-RadishMind-Dev-Read-Scopes"), "workflow_runs:read,workflow_rag_snapshots:read");
  } finally { globalThis.fetch = originalFetch; }
});

test("history rejects nested fragment bodies, unknown previews, and previews over 512 characters", async () => {
  const originalFetch = globalThis.fetch;
  try {
    const leaking = listEnvelope() as any;
    leaking.runs[0].selected_fragments[0].fragment_content = "private fragment body";
    globalThis.fetch = async () => jsonResponse(leaking);
    await assert.rejects(() => listWorkflowRunHistory("app_flow_copilot", live, EMPTY_WORKFLOW_RUN_HISTORY_FILTER), /HTTP 200/u);

    const validHistory = listEnvelope();
    globalThis.fetch = async () => jsonResponse(validHistory);
    const history = await listWorkflowRunHistory("app_flow_copilot", live, EMPTY_WORKFLOW_RUN_HISTORY_FILTER);

    globalThis.fetch = async () => jsonResponse(detailEnvelope("x".repeat(513)));
    await assert.rejects(() => readWorkflowRunHistoryDetail(history.runs[0]!, "app_flow_copilot", live, true), /incompatible retrieval previews/u);

    const unknownRef = detailEnvelope("bounded") as any;
    unknownRef.retrieval_fragment_previews[0].fragment_ref = "not_selected";
    globalThis.fetch = async () => jsonResponse(unknownRef);
    await assert.rejects(() => readWorkflowRunHistoryDetail(history.runs[0]!, "app_flow_copilot", live, true), /incompatible retrieval previews/u);
  } finally { globalThis.fetch = originalFetch; }
});

test("RAG comparison selection requires an exact immutable retrieval binding", async () => {
  const originalFetch = globalThis.fetch;
  const body = listEnvelope();
  const compatible = structuredClone(body.runs[0]!);
  compatible.run_id = "run_bbbbbbbbbbbbbbbb";
  const differentQuery = structuredClone(body.runs[0]!);
  differentQuery.run_id = "run_cccccccccccccccc";
  differentQuery.query_digest = digestB;
  body.runs.push(compatible, differentQuery);
  globalThis.fetch = async () => jsonResponse(body);
  try {
    const history = await listWorkflowRunHistory("app_flow_copilot", live, EMPTY_WORKFLOW_RUN_HISTORY_FILTER);
    assert.equal(isWorkflowRunComparisonEligible(history.runs[0]!), true);
    assert.equal(isWorkflowRunComparisonCompatible(history.runs[0], history.runs[1]!), true);
    assert.equal(isWorkflowRunComparisonCompatible(history.runs[0], history.runs[2]!), false);
  } finally { globalThis.fetch = originalFetch; }
});

function listEnvelope() {
  return {
    request_id: "request_history_v3", workspace_id: "workspace_demo", application_id: "app_flow_copilot",
    runs: [{
      schema_version: "workflow_run_record.v3", record_version: 2, run_id: "run_aaaaaaaaaaaaaaaa", draft_id: "draft_rag_execution", draft_version: 4, draft_digest: digestA,
      workspace_id: "workspace_demo", application_id: "app_flow_copilot", status: "succeeded", failure_code: "", started_at: "2026-07-18T08:00:00Z", completed_at: "2026-07-18T08:00:01Z", duration_ms: 1000,
      selected_provider: "mock", selected_profile: "workflow.rag.lexical-ngram-dev.v1", selected_model: "mock-rag", request_id: "request_run_v3", audit_ref: "audit_run_v3", stale_running: false,
      failure_boundary: "", failed_node_id: "", last_completed_node_id: "node_rag_model", gateway_failure_category: "none", tool_failure_category: "none", recommended_review_action: "",
      snapshot_id: "rags_abcdefghijklmnop", snapshot_version: 2, snapshot_digest: digestB, rag_ref: "workflow.rag.product_docs.v2",
      retrieval_node_id: "node_rag_retrieval", retrieval_attempt_status: "succeeded", retrieval_profile_id: "workflow.rag.lexical-ngram-dev.v1", retrieval_profile_version: 1, retrieval_profile_digest: digestC,
      query_digest: digestA, query_bytes: 31, candidate_count: 1, selected_fragments: [{ fragment_ref: "official_guide", content_digest: digestB, rank: 1, source_type: "manual", is_official: true, excerpt_truncated: false }], citation_refs: ["official_guide"], retrieval_latency_ms: 2, retrieval_context_bytes: 92, retrieval_failure_category: "none",
      side_effects: { retrieval_calls: 1, provider_calls: 1, tool_calls: 0, confirmation_calls: 0, business_writes: 0, replay_writes: 0 },
    }],
    next_cursor: "", has_more: false, failure_code: null, failure_summary: "", audit_ref: "audit_history_v3",
  };
}

function detailEnvelope(preview: string) {
  return {
    request_id: "request_detail_v3", workspace_id: "workspace_demo", application_id: "app_flow_copilot", run: runRecord(), failure_code: null, failure_summary: "", audit_ref: "audit_detail_v3",
    retrieval_fragment_previews: [{ fragment_ref: "official_guide", preview, truncated: false }],
  };
}

function runRecord() {
  return {
    schema_version: "workflow_run_record.v3", record_version: 2, run_id: "run_aaaaaaaaaaaaaaaa", tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: "app_flow_copilot", draft_id: "draft_rag_execution", draft_version: 4, draft_digest: digestA,
    status: "succeeded", failure_code: "", failure_summary: "", started_at: "2026-07-18T08:00:00Z", completed_at: "2026-07-18T08:00:01Z",
    snapshot: { snapshot_id: "rags_abcdefghijklmnop", snapshot_version: 2, snapshot_digest: digestB, rag_ref: "workflow.rag.product_docs.v2" },
    retrieval_attempt: { node_id: "node_rag_retrieval", status: "succeeded", profile_id: "workflow.rag.lexical-ngram-dev.v1", profile_version: 1, profile_digest: digestC, query_digest: digestA, query_bytes: 31, candidate_count: 1, selected_fragments: [{ fragment_ref: "official_guide", content_digest: digestB, rank: 1, source_type: "manual", is_official: true, excerpt_truncated: false }], retrieval_latency_ms: 2, context_bytes: 92, citation_refs: ["official_guide"] },
    answer: null, selected_provider: "mock", selected_model: "mock-rag", request_id: "request_run_v3", audit_ref: "audit_run_v3", actor_ref: "subject_demo_user",
    side_effects: { retrieval_calls: 1, provider_calls: 1, tool_calls: 0, confirmation_calls: 0, business_writes: 0, replay_writes: 0 }, diagnostic: { failure_boundary: "none", retrieval_failure_category: "none" },
  };
}

function jsonResponse(value: unknown) { return new Response(JSON.stringify(value), { status: 200, headers: { "Content-Type": "application/json" } }); }
