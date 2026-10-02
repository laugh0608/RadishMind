import assert from "node:assert/strict";
import test from "node:test";

import {
  buildWorkflowRAGRetrievalDraft,
  evaluateWorkflowRAGExecutionEligibility,
  executeWorkflowRAGRetrieval,
  initialWorkflowRAGExecutionState,
  validateWorkflowRAGExecutionInput,
} from "../src/features/control-plane-read/workflowRAGExecutionConsumer.ts";
import type { WorkflowRAGSnapshotConfig } from "../src/features/control-plane-read/workflowRAGSnapshotConsumer.ts";
import { saveWorkflowDraftDevRecord } from "../src/features/control-plane-read/savedWorkflowDraftConsumer.ts";
import { workflowSavedDraftRequestedCapabilities } from "../src/features/control-plane-read/workflowSavedDraftCapabilityPolicy.ts";

const digestA = `sha256:${"a".repeat(64)}`;
const digestB = `sha256:${"b".repeat(64)}`;
const digestC = `sha256:${"c".repeat(64)}`;
const config: WorkflowRAGSnapshotConfig = {
  mode: "dev_workflow_rag_http",
  baseUrl: "http://platform.test",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  subjectRef: "subject_demo_user",
  authMode: "dev_headers",
  scopes: new Set(["workflow_rag:execute", "workflow_runs:execute", "workflow_drafts:read", "workflow_rag_snapshots:read"]),
};

test("RAG draft builder creates the exact four-stage graph without enabling executor v0 retrieval", () => {
  const draft = boundDraft();
  assert.equal(draft.executionProfile, "rag_retrieval_v1");
  assert.deepEqual(draft.nodes.map((node) => node.nodeType), ["prompt", "rag_retrieval", "llm", "output"]);
  assert.deepEqual(draft.edges.map((edge) => [edge.fromNodeId, edge.toNodeId, edge.conditionSummary]), [
    ["node_rag_prompt", "node_rag_retrieval", ""],
    ["node_rag_retrieval", "node_rag_model", ""],
    ["node_rag_model", "node_rag_output", ""],
  ]);
  assert.equal(draft.nodes.every((node) => !node.toolRef && !node.requiresConfirmation), true);
  assert.deepEqual(workflowSavedDraftRequestedCapabilities(draft), []);
});

test("RAG draft builder binds the selected application scope", () => {
  const draft = buildWorkflowRAGRetrievalDraft(sourceDraft(), 2, "app_selected_catalog");
  assert.equal(draft.applicationRef, "app_selected_catalog");
  assert.equal(draft.draftId, "draft_app_selected_catalog_rag_v1_02");
});

test("saved draft persistence carries the exact rag_ref and RAG execution metadata", async () => {
  const originalFetch = globalThis.fetch;
  let body: any = null;
  let headers = new Headers();
  globalThis.fetch = async (_input, init) => {
    body = JSON.parse(String(init?.body));
    headers = new Headers(init?.headers);
    return response({ request_id: "request_save_rag", workspace_id: "workspace_demo", application_id: "app_flow_copilot", draft: null, failure_code: "draft_store_unavailable", current_draft_version: 0, current_lifecycle_version: 0, current_lifecycle_state: "", validation_summary: { validation_state: "unavailable", valid_for_review: false, findings: [] }, blocked_capabilities: [], audit_ref: "audit_save_rag" });
  };
  try {
    await saveWorkflowDraftDevRecord(boundDraft(), { mode: "dev_saved_draft_http", baseUrl: "http://platform.test", tenantRef: "tenant_demo", workspaceId: "workspace_demo", subjectRef: "subject_demo_user" }, 0);
    assert.deepEqual(body.draft.rag_refs, ["workflow.rag.product_docs.v2"]);
    assert.equal(body.draft.nodes.find((node: any) => node.node_type === "rag_retrieval").rag_ref, "workflow.rag.product_docs.v2");
    assert.deepEqual(body.draft.additional_fields.rag_retrieval_v1, { version: 1, execution_route: "retrieval_executions", side_effect_policy: "retrieval_and_provider_once" });
    assert.deepEqual(body.draft.edges.map((edge: any) => edge.condition_summary), ["", "", ""]);
    assert.deepEqual(body.draft.requested_capabilities, []);
    assert.equal(headers.get("X-RadishMind-Active-Workspace"), "workspace_demo");
    assert.equal(headers.get("X-RadishMind-Dev-Read-Membership-Permissions"), "workflow_drafts:read,workflow_drafts:write");
    assert.equal(body.expected_lifecycle_version, 0);
  } finally { globalThis.fetch = originalFetch; }
});

test("execution eligibility requires every grant, an exact saved version, and the exact topology", () => {
  const draft = boundDraft();
  assert.equal(evaluateWorkflowRAGExecutionEligibility(draft, savedState(), false, config).eligible, true);
  assert.equal(
    evaluateWorkflowRAGExecutionEligibility(
      draft,
      { ...savedState(), currentLifecycleState: "archived" },
      false,
      config,
    ).eligible,
    false,
  );
  assert.equal(evaluateWorkflowRAGExecutionEligibility(draft, savedState(), true, config).reasons.some((reason) => reason.code === "unsaved_local_changes"), true);

  const missingScope = { ...config, scopes: new Set(["workflow_rag:execute", "workflow_runs:execute", "workflow_drafts:read"] as const) };
  assert.equal(evaluateWorkflowRAGExecutionEligibility(draft, savedState(), false, missingScope).reasons.some((reason) => reason.code === "rag_execution_scope_denied"), true);

  const invalidGraph = { ...draft, edges: [...draft.edges, { ...draft.edges[0]!, edgeId: "edge_extra" }] };
  assert.equal(evaluateWorkflowRAGExecutionEligibility(invalidGraph, savedState(), false, config).reasons.some((reason) => reason.code === "rag_execution_edges_invalid"), true);

  const unbound = { ...draft, nodes: draft.nodes.map((node) => node.nodeType === "rag_retrieval" ? { ...node, ragRef: "" } : node) };
  assert.equal(evaluateWorkflowRAGExecutionEligibility(unbound, savedState(), false, config).reasons.some((reason) => reason.code === "rag_ref_invalid"), true);
});

test("offline, scope denial, and invalid input perform zero execution requests", async () => {
  const originalFetch = globalThis.fetch;
  let calls = 0;
  globalThis.fetch = async () => { calls += 1; throw new Error("must remain offline"); };
  try {
    const draft = boundDraft();
    const offline = { ...config, mode: "offline" as const };
    assert.equal(initialWorkflowRAGExecutionState(offline).status, "offline");
    const deniedEligibility = evaluateWorkflowRAGExecutionEligibility(draft, savedState(), false, offline);
    assert.equal((await executeWorkflowRAGRetrieval(offline, draft, deniedEligibility, validInput())).status, "failed");

    const eligible = evaluateWorkflowRAGExecutionEligibility(draft, savedState(), false, config);
    assert.equal((await executeWorkflowRAGRetrieval(config, draft, eligible, { ...validInput(), inputText: " " })).failureCode, "workflow_rag_query_invalid");
    assert.equal(validateWorkflowRAGExecutionInput({ ...validInput(), inputText: "x".repeat(4097) }), "workflow_rag_query_invalid");
    assert.equal(calls, 0);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("execution posts only the bounded request and maps metadata-only v3 with a transient answer", async () => {
  const originalFetch = globalThis.fetch;
  let captured: { url: string; headers: Headers; body: Record<string, unknown> } | null = null;
  globalThis.fetch = async (input, init) => {
    captured = { url: String(input), headers: new Headers(init?.headers), body: JSON.parse(String(init?.body)) };
    return response(executionEnvelope());
  };
  try {
    const draft = boundDraft();
    const eligibility = evaluateWorkflowRAGExecutionEligibility(draft, savedState(), false, config);
    const result = await executeWorkflowRAGRetrieval(config, draft, eligibility, validInput());
    assert.equal(result.status, "succeeded", JSON.stringify(result));
    assert.equal(result.record?.schemaVersion, "workflow_run_record.v3");
    assert.equal(result.record?.retrievalAttempt?.selectedFragments[0]?.fragmentRef, "official_guide");
    assert.equal(result.answer?.citations[0]?.fragmentRef, "official_guide");
    assert.equal(result.record?.output, "");
    assert.equal(result.record?.sideEffects.retrievalCalls, 1);
    assert.equal(result.record?.sideEffects.providerCalls, 1);
    assert.match(captured!.url, /\/workflow-drafts\/draft_app_flow_copilot_rag_v1_01\/retrieval-executions$/u);
    assert.equal(captured!.headers.get("X-RadishMind-Dev-Read-Scopes"), "workflow_rag:execute,workflow_runs:execute,workflow_drafts:read,workflow_rag_snapshots:read");
    assert.equal(captured!.headers.get("X-RadishMind-Active-Workspace"), "workspace_demo");
    assert.equal(captured!.headers.get("X-RadishMind-Dev-Read-Membership-Workspace"), "workspace_demo");
    assert.equal(captured!.headers.get("X-RadishMind-Dev-Read-Membership-Permissions"), "workflow_rag:execute,workflow_runs:execute,workflow_drafts:read,workflow_rag_snapshots:read");
    assert.deepEqual(captured!.body, { workspace_id: "workspace_demo", application_id: "app_flow_copilot", draft_version: 4, input_text: "Explain the supported boundary.", model: "mock-rag", temperature: 0.2 });
    for (const forbidden of ["fragment", "rank", "citation", "snapshot_digest", "profile_digest", "rag_ref"]) assert.equal(forbidden in captured!.body, false);
    assert.equal(JSON.stringify(result.record).includes("Explain the supported boundary"), false);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("consumer rejects unselected, duplicate, malformed, and leaking model evidence without retry", async () => {
  const originalFetch = globalThis.fetch;
  const draft = boundDraft();
  const eligibility = evaluateWorkflowRAGExecutionEligibility(draft, savedState(), false, config);
  let calls = 0;
  try {
    for (const mutate of [
      (body: any) => { body.retrieval_answer.citations[0].fragment_ref = "unselected_fragment"; },
      (body: any) => { body.retrieval_answer.citations.push({ ...body.retrieval_answer.citations[0] }); },
      (body: any) => { body.retrieval_answer.confidence = "certain"; },
      (body: any) => { body.run.raw_response = "provider material"; },
      (body: any) => { body.action_safety = { effective_level: "write_allowed_by_policy" }; },
      (body: any) => { delete body.action_safety; },
    ]) {
      const body = executionEnvelope() as any;
      mutate(body);
      globalThis.fetch = async () => { calls += 1; return response(body); };
      const result = await executeWorkflowRAGRetrieval(config, draft, eligibility, validInput());
      assert.equal(result.status, "failed");
      assert.equal(result.answer, null);
      assert.equal(result.record, null);
    }
    assert.equal(calls, 6);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

function boundDraft() {
  const draft = buildWorkflowRAGRetrievalDraft(sourceDraft(), 1);
  return { ...draft, nodes: draft.nodes.map((node) => node.nodeType === "rag_retrieval" ? { ...node, ragRef: "workflow.rag.product_docs.v2" } : node) };
}

function sourceDraft() {
  return {
    draftId: "draft_source", templateRef: "wf_source", label: "RadishFlow advisory", applicationRef: "app_flow_copilot",
    workflowDefinitionId: "wf_radishflow_copilot_latest", providerProfileRef: "provider:mock", summary: "source", nodes: [], edges: [],
    designerLayout: { source: "workflow_node_designer" as const, persistence: "ui_only" as const, nodePositions: [] },
    readiness: [], risks: [], blockedCapabilities: [],
    routeMetadata: { sourceRouteId: "workflow-definition-summary-list-route" as const, draftRouteId: "workflow-draft-designer-offline-draft" as const, routePath: "/v1/user-workspace/workflow-definitions" as const, requestId: "req_source", auditRef: "audit_source" },
    localOnlyInteraction: "inspect_only" as const, executionProfile: "review_only" as const,
  };
}

function savedState() {
  return { status: "saved_dev_record" as const, mode: "dev_saved_draft_http" as const, sourceLabel: "saved", summary: "saved", failureCode: null, currentDraftVersion: 4, currentLifecycleVersion: 1, currentLifecycleState: "active" as const, conflictDraftVersion: null, auditRef: "audit_saved", requestId: "request_saved" };
}

function validInput() { return { inputText: "Explain the supported boundary.", model: "mock-rag", temperature: 0.2 }; }

function executionEnvelope() {
  return {
    action_safety: null,
    request_id: "request_run_v3", workspace_id: "workspace_demo", application_id: "app_flow_copilot", failure_code: null, failure_summary: "", audit_ref: "audit_run_v3",
    run: {
      schema_version: "workflow_run_record.v3", record_version: 2, run_id: "run_aaaaaaaaaaaaaaaa", tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: "app_flow_copilot",
      draft_id: "draft_app_flow_copilot_rag_v1_01", draft_version: 4, draft_digest: digestA, status: "succeeded", failure_code: "", failure_summary: "", started_at: "2026-07-18T08:00:00Z", completed_at: "2026-07-18T08:00:01Z",
      snapshot: { snapshot_id: "rags_abcdefghijklmnop", snapshot_version: 2, snapshot_digest: digestB, rag_ref: "workflow.rag.product_docs.v2" },
      retrieval_attempt: { node_id: "node_rag_retrieval", status: "succeeded", profile_id: "workflow.rag.lexical-ngram-dev.v1", profile_version: 1, profile_digest: digestC, query_digest: digestA, query_bytes: 31, candidate_count: 1, selected_fragments: [{ fragment_ref: "official_guide", content_digest: digestB, rank: 1, source_type: "manual", is_official: true, excerpt_truncated: false }], retrieval_latency_ms: 2, context_bytes: 92, citation_refs: ["official_guide"] },
      answer: null, selected_provider: "mock", selected_model: "mock-rag", request_id: "request_run_v3", audit_ref: "audit_run_v3", actor_ref: "subject_demo_user",
      side_effects: { retrieval_calls: 1, provider_calls: 1, tool_calls: 0, confirmation_calls: 0, business_writes: 0, replay_writes: 0 },
      diagnostic: { failure_boundary: "none", retrieval_failure_category: "none" },
    },
    retrieval_answer: { schema_version: "workflow_rag_answer.v1", answer: "The evidence supports this advisory answer.", citations: [{ fragment_ref: "official_guide", claim_summary: "Official guidance supports the answer." }], limitations: ["Limited to the selected snapshot."], confidence: "medium" },
  };
}

function response(value: unknown) { return new Response(JSON.stringify(value), { status: 200, headers: { "Content-Type": "application/json" } }); }

test("retained RAG eligibility and outcome messages render in both languages without changing evidence", async () => {
  const { createUiI18n, initializeUiI18n } = await import("../src/i18n/instance.ts");
  const { workflowRAGExecution: en } = await import("../src/i18n/locales/en-US/workflowRAGExecution.ts");
  const { workflowRAGExecution: zh } = await import("../src/i18n/locales/zh-CN/workflowRAGExecution.ts");
  const { workflowRAGExecutionReasonMessage, workflowRAGExecutionFeedback, workflowRAGExecutionStatus, workflowRAGExecutionFailure, workflowRAGConfidenceLabel } = await import("../src/features/control-plane-read/workflowRAGExecutionMessages.ts");
  const i18n = createUiI18n(); await initializeUiI18n(i18n, "en-US");
  i18n.addResourceBundle("en-US", "workflow", { ragExecution: en });
  i18n.addResourceBundle("zh-CN", "workflow", { ragExecution: zh });
  const denied = evaluateWorkflowRAGExecutionEligibility(sourceDraft(), savedState(), true, { ...config, mode: "offline", scopes: new Set() });
  const topology = evaluateWorkflowRAGExecutionEligibility({ ...boundDraft(), edges: [], blockedCapabilities: ["tool"] }, { ...savedState(), currentLifecycleState: "archived" }, false, config);
  const reasons = [...denied.reasons, ...topology.reasons];
  assert.equal(new Set(reasons.map(reason => reason.code)).size, 10);
  const before = JSON.stringify(reasons);
  const render = () => {
    const t = i18n.getFixedT(null, "workflow");
    return [
      ...reasons.map(reason => workflowRAGExecutionReasonMessage(t, reason)),
      ...(["offline", "ready", "executing", "succeeded", "localRejected", "failed", "unavailable"] as const).map(message => workflowRAGExecutionFeedback(t, message)),
      ...Object.keys(en.status).map(status => workflowRAGExecutionStatus(t, status)),
      ...Object.keys(en.failure).map(code => workflowRAGExecutionFailure(t, code)),
      ...(["low", "medium", "high"] as const).map(level => workflowRAGConfidenceLabel(t, level)),
    ];
  };
  const english = render();
  assert.equal(workflowRAGExecutionFailure(i18n.getFixedT(null, "workflow"), "workflow_rag_no_evidence"), "No matching evidence met the retrieval threshold. Review the question and selected snapshot.");
  assert.match(workflowRAGExecutionReasonMessage(i18n.getFixedT(null, "workflow"), { code: "rag_execution_scope_denied", scope: "workflow_rag:execute" }), /workflow_rag:execute/);
  await i18n.changeLanguage("zh-CN");
  assert.equal(workflowRAGExecutionFailure(i18n.getFixedT(null, "workflow"), "workflow_rag_no_evidence"), "没有匹配证据达到检索阈值，请检查问题与所选快照。");
  assert.equal(workflowRAGExecutionFailure(i18n.getFixedT(null, "workflow"), "unknown_server_failure"), "操作失败，请根据错误码与请求引用检查原因。");
  render().forEach((message, index) => {
    assert.ok(message && english[index]);
    assert.notEqual(message, english[index]);
    assert.doesNotMatch(message + english[index], /\{\{|ragExecution\.|This message is unavailable|此消息暂不可用/);
  });
  assert.equal(JSON.stringify(reasons), before);
});

test("RAG response wait and language changes preserve exact request, answer and invocation count", async () => {
  const { createUiI18n, initializeUiI18n } = await import("../src/i18n/instance.ts");
  const i18n = createUiI18n(); await initializeUiI18n(i18n, "en-US");
  const originalFetch = globalThis.fetch;
  let release!: () => void;
  const pending = new Promise<void>(resolve => { release = resolve; });
  let calls = 0;
  let payload: unknown;
  globalThis.fetch = async (_input, init) => {
    calls++;
    payload = JSON.parse(String(init?.body));
    await pending;
    return response(executionEnvelope());
  };
  try {
    const draft = boundDraft();
    const beforeDraft = JSON.stringify(draft);
    const resultPending = executeWorkflowRAGRetrieval(config, draft, evaluateWorkflowRAGExecutionEligibility(draft, savedState(), false, config), validInput());
    await i18n.changeLanguage("zh-CN");
    assert.equal(calls, 1);
    assert.deepEqual(payload, { workspace_id: "workspace_demo", application_id: "app_flow_copilot", draft_version: 4, input_text: "Explain the supported boundary.", model: "mock-rag", temperature: 0.2 });
    release();
    const result = await resultPending;
    assert.equal(result.message, "succeeded");
    assert.equal(result.answer?.answer, "The evidence supports this advisory answer.");
    assert.equal(result.answer?.citations[0]?.fragmentRef, "official_guide");
    const beforeResult = JSON.stringify(result);
    await i18n.changeLanguage("en-US");
    assert.equal(JSON.stringify(result), beforeResult);
    assert.equal(JSON.stringify(draft), beforeDraft);
    assert.equal(calls, 1);
  } finally { release(); globalThis.fetch = originalFetch; }
});

test("RAG local, transport and server failures retain distinct stable feedback without exposing raw diagnostics", async () => {
  const originalFetch = globalThis.fetch;
  const draft = boundDraft();
  const eligibility = evaluateWorkflowRAGExecutionEligibility(draft, savedState(), false, config);
  let calls = 0;
  try {
    globalThis.fetch = async () => { calls++; throw new Error("private raw transport detail"); };
    assert.equal((await executeWorkflowRAGRetrieval(config, draft, eligibility, { ...validInput(), inputText: "" })).message, "localRejected");
    assert.equal(calls, 0);
    const unavailable = await executeWorkflowRAGRetrieval(config, draft, eligibility, validInput());
    assert.equal(unavailable.message, "unavailable");
    assert.equal(calls, 1);
    assert.ok(!JSON.stringify(unavailable).includes("private raw"));
    for (const code of ["workflow_rag_no_evidence", "workflow_rag_snapshot_scope_denied", "unknown_server_failure"]) {
      globalThis.fetch = async () => response({ ...executionEnvelope(), run: null, retrieval_answer: null, failure_code: code, failure_summary: "untranslated server diagnostic" });
      const failed = await executeWorkflowRAGRetrieval(config, draft, eligibility, validInput());
      assert.equal(failed.message, "failed");
      assert.equal(failed.failureCode, code);
      assert.equal(failed.status, code === "workflow_rag_snapshot_scope_denied" ? "scope_denied" : "failed");
      assert.equal(failed.requestId, "request_run_v3");
      assert.ok(!JSON.stringify(failed).includes("untranslated server diagnostic"));
    }
  } finally { globalThis.fetch = originalFetch; }
});
