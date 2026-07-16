import assert from "node:assert/strict";
import test from "node:test";

import {
  createWorkflowHTTPToolActionPlan,
  decideWorkflowHTTPToolActionPlan,
  evaluateWorkflowHTTPToolActionEligibility,
  initialWorkflowHTTPToolActionConsumerState,
  readWorkflowHTTPToolActionPlan,
  validateWorkflowHTTPToolPublicArguments,
  workflowHTTPToolActionPermissions,
  type WorkflowHTTPToolActionPlan,
} from "../src/features/control-plane-read/workflowHTTPToolActionConsumer.ts";
import type { WorkflowDraftDesignerDraft } from "../src/features/control-plane-read/workflowDraftDesigner.ts";

const config = {
  mode: "dev_workflow_http_tool_http" as const,
  baseUrl: "http://127.0.0.1:7000",
  workspaceId: "workspace_demo",
  tenantRef: "tenant_demo",
  subjectRef: "subject_demo_user",
  scopeGrants: [
    "workflow_drafts:read",
    "workflow_tool_actions:plan",
    "workflow_tool_actions:read",
    "workflow_tool_actions:confirm",
  ],
};

test("Batch A permissions separate plan, read, confirm, and unavailable execution", () => {
  const permissions = workflowHTTPToolActionPermissions(config);
  assert.equal(permissions.plan.available, true);
  assert.equal(permissions.read.available, true);
  assert.equal(permissions.confirm.available, true);
  assert.equal(permissions.execute.available, false);
  assert.equal(permissions.execute.phase, "batch_b");
  assert.deepEqual(permissions.execute.requiredGrants, ["workflow_tool_actions:execute"]);

  const restricted = workflowHTTPToolActionPermissions({ ...config, scopeGrants: ["workflow_tool_actions:read"] });
  assert.equal(restricted.plan.available, false);
  assert.equal(restricted.read.available, true);
  assert.equal(restricted.confirm.available, false);
});

test("HTTP Tool action eligibility requires an exact saved clean single-chain draft", () => {
  const draft = eligibleDraft();
  const saved = savedDraftState();
  const eligible = evaluateWorkflowHTTPToolActionEligibility(draft, saved, false);
  assert.equal(eligible.eligible, true);
  assert.equal(eligible.draftVersion, 4);
  assert.equal(eligible.nodeId, "node_http_tool");

  const dirty = evaluateWorkflowHTTPToolActionEligibility(draft, saved, true);
  assert.equal(dirty.eligible, false);
  assert.equal(dirty.reasons.some((reason) => reason.code === "workflow_tool_unsaved_local_changes"), true);

  const previewReference = {
    ...draft,
    nodes: draft.nodes.map((node) => node.nodeId === "node_http_tool"
      ? { ...node, toolRef: "tool:workflow-preview-readonly" }
      : node),
  };
  const preview = evaluateWorkflowHTTPToolActionEligibility(previewReference, saved, false);
  assert.equal(preview.eligible, false);
  assert.equal(preview.reasons.some((reason) => reason.code === "workflow_tool_exact_version_required"), true);
});

test("public arguments accept only the registered resource and locale fields", () => {
  const valid = validateWorkflowHTTPToolPublicArguments({ resourceKey: "catalog/review:item-1", locale: "zh-CN" });
  assert.equal(valid.valid, true);
  assert.deepEqual(valid.value, { resourceKey: "catalog/review:item-1", locale: "zh-CN" });

  assert.equal(validateWorkflowHTTPToolPublicArguments({ resourceKey: "https://service.invalid/item" }).valid, false);
  assert.equal(validateWorkflowHTTPToolPublicArguments({ resourceKey: "reviewed-item", locale: "not a locale" }).valid, false);
  assert.equal(validateWorkflowHTTPToolPublicArguments({ resourceKey: "authorization=Bearer-secret" }).valid, false);
});

test("disabled source performs zero fetches for create, detail, and decisions", async (t) => {
  let fetchCount = 0;
  const originalFetch = globalThis.fetch;
  t.after(() => { globalThis.fetch = originalFetch; });
  globalThis.fetch = async () => {
    fetchCount += 1;
    throw new Error("disabled source must not fetch");
  };
  const disabled = { ...config, mode: "disabled" as const };
  const plan = mappedPlan();

  const initial = initialWorkflowHTTPToolActionConsumerState(disabled);
  const created = await createWorkflowHTTPToolActionPlan(disabled, {
    draftId: plan.draftId,
    applicationId: plan.applicationId,
    draftVersion: plan.draftVersion,
    nodeId: plan.nodeId,
    publicArguments: plan.publicArguments,
  });
  const read = await readWorkflowHTTPToolActionPlan(disabled, plan);
  const decided = await decideWorkflowHTTPToolActionPlan(disabled, plan, "approve");

  assert.equal(initial.status, "disabled");
  assert.equal(created.status, "disabled");
  assert.equal(read.status, "disabled");
  assert.equal(decided.status, "disabled");
  assert.equal(fetchCount, 0);
});

test("missing operation grants fail before any HTTP request", async (t) => {
  let fetchCount = 0;
  const originalFetch = globalThis.fetch;
  t.after(() => { globalThis.fetch = originalFetch; });
  globalThis.fetch = async () => {
    fetchCount += 1;
    throw new Error("missing grants must not fetch");
  };
  const restricted = { ...config, scopeGrants: ["workflow_tool_actions:read"] };
  const plan = mappedPlan();
  const created = await createWorkflowHTTPToolActionPlan(restricted, {
    draftId: plan.draftId,
    applicationId: plan.applicationId,
    draftVersion: plan.draftVersion,
    nodeId: plan.nodeId,
    publicArguments: plan.publicArguments,
  });
  const decided = await decideWorkflowHTTPToolActionPlan(restricted, plan, "approve");
  assert.equal(created.failureCode, "workflow_tool_action_scope_denied");
  assert.equal(decided.failureCode, "workflow_tool_action_scope_denied");
  assert.equal(fetchCount, 0);
});

test("create sends only exact saved scope and public arguments, then maps a redacted plan", async (t) => {
  const requests: Array<{ url: string; init: RequestInit }> = [];
  const originalFetch = globalThis.fetch;
  t.after(() => { globalThis.fetch = originalFetch; });
  globalThis.fetch = async (input, init = {}) => {
    requests.push({ url: String(input), init });
    return jsonResponse(successEnvelope(actionPlanDocument()));
  };

  const state = await createWorkflowHTTPToolActionPlan(config, {
    draftId: "draft_http_tool_review",
    applicationId: "app_flow_copilot",
    draftVersion: 4,
    nodeId: "node_http_tool",
    publicArguments: { resourceKey: "catalog/review:item-1", locale: "zh-CN" },
  });

  assert.equal(state.status, "ready");
  assert.equal(state.actionPlan?.toolId, "workflow.http.reviewed-json-read.v1");
  assert.equal(state.actionPlan?.targetPolicyKey, "reviewed_json_read_dev");
  assert.equal(requests.length, 1);
  assert.equal(requests[0]?.url.endsWith("/v1/user-workspace/workflow-drafts/draft_http_tool_review/tool-action-plans"), true);
  const body = JSON.parse(String(requests[0]?.init.body));
  assert.deepEqual(body, {
    workspace_id: "workspace_demo",
    application_id: "app_flow_copilot",
    draft_version: 4,
    node_id: "node_http_tool",
    public_arguments: { resource_key: "catalog/review:item-1", locale: "zh-CN" },
  });
  assert.equal("tool_id" in body, false);
  assert.equal("profile_id" in body, false);
  const headers = requests[0]?.init.headers as Record<string, string>;
  assert.equal(headers["X-RadishMind-Dev-Read-Scopes"], "workflow_drafts:read,workflow_tool_actions:plan");
});

test("detail reload reads the durable scoped plan without a list route", async (t) => {
  const requests: Array<{ url: string; init: RequestInit }> = [];
  const originalFetch = globalThis.fetch;
  t.after(() => { globalThis.fetch = originalFetch; });
  globalThis.fetch = async (input, init = {}) => {
    requests.push({ url: String(input), init });
    return jsonResponse(successEnvelope(actionPlanDocument()));
  };

  const state = await readWorkflowHTTPToolActionPlan(config, mappedPlan());
  assert.equal(state.status, "ready");
  assert.equal(state.actionPlan?.recordVersion, 1);
  assert.match(requests[0]!.url, /workflow-tool-action-plans\/wtap_abcdefghijklmnop\?workspace_id=workspace_demo&application_id=app_flow_copilot$/u);
  assert.equal(requests[0]?.init.method, "GET");
  const headers = requests[0]?.init.headers as Record<string, string>;
  assert.equal(headers["X-RadishMind-Dev-Read-Scopes"], "workflow_tool_actions:read");
});

test("approval records a CAS decision and never calls an execution route", async (t) => {
  const requests: Array<{ url: string; init: RequestInit }> = [];
  const originalFetch = globalThis.fetch;
  t.after(() => { globalThis.fetch = originalFetch; });
  globalThis.fetch = async (input, init = {}) => {
    requests.push({ url: String(input), init });
    return jsonResponse(successEnvelope(approvedActionPlanDocument(), confirmationDecisionDocument()));
  };

  const state = await decideWorkflowHTTPToolActionPlan(config, mappedPlan(), "approve");
  assert.equal(state.status, "ready");
  assert.equal(state.actionPlan?.status, "approved");
  assert.equal(state.confirmationDecision?.outcome, "approve");
  assert.equal(requests.length, 1);
  assert.equal(requests[0]?.url.endsWith("/workflow-tool-action-plans/wtap_abcdefghijklmnop/decisions"), true);
  assert.equal(requests.some((request) => request.url.includes("/executions")), false);
  assert.deepEqual(JSON.parse(String(requests[0]?.init.body)), {
    workspace_id: "workspace_demo",
    application_id: "app_flow_copilot",
    expected_record_version: 1,
    decision: "approve",
  });
  const headers = requests[0]?.init.headers as Record<string, string>;
  assert.equal(headers["X-RadishMind-Dev-Read-Scopes"], "workflow_tool_actions:confirm");
});

test("CAS conflict automatically refreshes the durable detail before another decision", async (t) => {
  const requests: Array<{ url: string; init: RequestInit }> = [];
  const originalFetch = globalThis.fetch;
  t.after(() => { globalThis.fetch = originalFetch; });
  globalThis.fetch = async (input, init = {}) => {
    requests.push({ url: String(input), init });
    if (requests.length === 1) {
      return jsonResponse(failureEnvelope("workflow_tool_confirmation_stale"), 409);
    }
    return jsonResponse(successEnvelope({
      ...actionPlanDocument(),
      record_version: 2,
      status: "deferred",
      last_decision_by_actor_ref: "subject_other_reviewer",
      last_decision_at: "2026-07-16T02:05:00Z",
    }));
  };

  const state = await decideWorkflowHTTPToolActionPlan(config, mappedPlan(), "approve");
  assert.equal(state.status, "conflict_refreshed");
  assert.equal(state.failureCode, "workflow_tool_confirmation_stale");
  assert.equal(state.actionPlan?.recordVersion, 2);
  assert.equal(state.actionPlan?.status, "deferred");
  assert.equal(requests.length, 2);
  assert.equal(requests[1]?.init.method, "GET");
});

test("strict consumer rejects forbidden or additional response fields", async (t) => {
  const originalFetch = globalThis.fetch;
  t.after(() => { globalThis.fetch = originalFetch; });
  globalThis.fetch = async () => jsonResponse(successEnvelope({
    ...actionPlanDocument(),
    endpoint: "https://private.invalid",
  }));

  const state = await readWorkflowHTTPToolActionPlan(config, mappedPlan());
  assert.equal(state.status, "failed");
  assert.equal(state.failureCode, "workflow_tool_store_contract_mismatch");
  assert.equal(state.actionPlan, null);
  assert.equal(JSON.stringify(state).includes("private.invalid"), false);
});

test("strict consumer rejects output schema field drift", async (t) => {
  const originalFetch = globalThis.fetch;
  t.after(() => { globalThis.fetch = originalFetch; });
  globalThis.fetch = async () => jsonResponse(successEnvelope({
    ...actionPlanDocument(),
    output_fields: ["summary", "status"],
  }));

  const state = await readWorkflowHTTPToolActionPlan(config, mappedPlan());
  assert.equal(state.status, "failed");
  assert.equal(state.failureCode, "workflow_tool_store_contract_mismatch");
  assert.equal(state.actionPlan, null);
});

function eligibleDraft(): WorkflowDraftDesignerDraft {
  return {
    draftId: "draft_http_tool_review",
    templateRef: "wf_http_tool_review",
    label: "HTTP Tool review",
    applicationRef: "app_flow_copilot",
    workflowDefinitionId: "wf_http_tool_review",
    providerProfileRef: "provider_mock",
    summary: "review",
    nodes: [
      node("node_prompt", "prompt", "", false, "low"),
      node("node_http_tool", "http_tool", "workflow.http.reviewed-json-read.v1", true, "medium"),
      node("node_llm", "llm", "", false, "low"),
      node("node_output", "output", "", false, "low"),
    ],
    edges: [
      edge("edge_prompt_tool", "node_prompt", "node_http_tool"),
      edge("edge_tool_llm", "node_http_tool", "node_llm"),
      edge("edge_llm_output", "node_llm", "node_output"),
    ],
    designerLayout: { source: "workflow_node_designer", persistence: "saved_draft_metadata", nodePositions: [] },
    readiness: [],
    risks: [],
    blockedCapabilities: [],
    routeMetadata: {
      sourceRouteId: "workflow-definition-summary-list-route",
      draftRouteId: "workflow-draft-designer-offline-draft",
      routePath: "/v1/user-workspace/workflow-definitions",
      requestId: "request_saved_draft",
      auditRef: "audit_saved_draft",
    },
    localOnlyInteraction: "inspect_only",
    executionProfile: "review_only",
  };
}

function node(
  nodeId: string,
  nodeType: "prompt" | "http_tool" | "llm" | "output",
  toolRef: string,
  requiresConfirmation: boolean,
  riskLevel: "low" | "medium",
) {
  return {
    nodeId,
    label: nodeId,
    nodeType,
    lane: nodeType === "prompt" ? "context" as const : nodeType === "llm" ? "model" as const : nodeType === "http_tool" ? "preview" as const : "output" as const,
    readiness: "ready" as const,
    inputSummary: "input",
    outputSummary: "output",
    providerRef: nodeType === "llm" ? "provider_mock" : "",
    toolRef,
    ragRef: "",
    inputContractFields: ["input"],
    outputContractFields: ["output"],
    outputMappingSummary: "mapping",
    riskLevel,
    requiresConfirmation,
    previewOnlyReason: "review",
  };
}

function edge(edgeId: string, fromNodeId: string, toNodeId: string) {
  return { edgeId, fromNodeId, toNodeId, edgeKind: "context" as const, conditionSummary: "next" };
}

function savedDraftState() {
  return {
    status: "saved_dev_record" as const,
    mode: "dev_saved_draft_http" as const,
    sourceLabel: "saved",
    summary: "saved",
    failureCode: null,
    currentDraftVersion: 4,
    conflictDraftVersion: null,
    auditRef: "audit_saved_draft",
    requestId: "request_saved_draft",
  };
}

function mappedPlan(): WorkflowHTTPToolActionPlan {
  return {
    schemaVersion: "workflow_http_tool_action_plan.v1",
    planId: "wtap_abcdefghijklmnop",
    recordVersion: 1,
    tenantRef: "tenant_demo",
    workspaceId: "workspace_demo",
    applicationId: "app_flow_copilot",
    draftId: "draft_http_tool_review",
    draftVersion: 4,
    nodeId: "node_http_tool",
    toolId: "workflow.http.reviewed-json-read.v1",
    toolVersion: 1,
    definitionDigest: "sha256:a3807ab3eed72b8f712ca4bf1f5cbc442fe961271779ff00ae19230ad3573294",
    profileId: "workflow_http_profile_reviewed_json_read_dev_v1",
    profileVersion: 1,
    profileDigest: "sha256:ce8517a547848768de11a9fae6ad903134ee8c5cfcc5a4b87436c26129918001",
    method: "GET",
    targetPolicyKey: "reviewed_json_read_dev",
    publicArguments: { resourceKey: "catalog/review:item-1", locale: "zh-CN" },
    outputFields: ["resource_key", "title", "summary", "updated_at"],
    outputSchemaDigest: "sha256:b36f96fa9eb6f095782981e1177bbf8c96f02e807524ed32c12772771fab82be",
    credentialPolicy: "none",
    timeoutMs: 5000,
    maxResponseBytes: 65536,
    maxOutputBytes: 16384,
    plannedByActorRef: "subject_demo_user",
    createdAt: "2026-07-16T02:00:00Z",
    expiresAt: "2026-07-16T02:15:00Z",
    toolPlanDigest: digest("c"),
    status: "pending",
    lastDecisionByActorRef: null,
    lastDecisionAt: null,
    auditRef: "audit_workflow_tool_plan",
  };
}

function actionPlanDocument() {
  const plan = mappedPlan();
  return {
    schema_version: plan.schemaVersion,
    plan_id: plan.planId,
    record_version: plan.recordVersion,
    tenant_ref: plan.tenantRef,
    workspace_id: plan.workspaceId,
    application_id: plan.applicationId,
    draft_id: plan.draftId,
    draft_version: plan.draftVersion,
    node_id: plan.nodeId,
    tool_id: plan.toolId,
    tool_version: plan.toolVersion,
    definition_digest: plan.definitionDigest,
    profile_id: plan.profileId,
    profile_version: plan.profileVersion,
    profile_digest: plan.profileDigest,
    method: plan.method,
    target_policy_key: plan.targetPolicyKey,
    public_arguments: { resource_key: plan.publicArguments.resourceKey, locale: plan.publicArguments.locale },
    output_fields: plan.outputFields,
    output_schema_digest: plan.outputSchemaDigest,
    credential_policy: plan.credentialPolicy,
    timeout_ms: plan.timeoutMs,
    max_response_bytes: plan.maxResponseBytes,
    max_output_bytes: plan.maxOutputBytes,
    planned_by_actor_ref: plan.plannedByActorRef,
    created_at: plan.createdAt,
    expires_at: plan.expiresAt,
    tool_plan_digest: plan.toolPlanDigest,
    status: plan.status,
    last_decision_by_actor_ref: plan.lastDecisionByActorRef,
    last_decision_at: plan.lastDecisionAt,
    audit_ref: plan.auditRef,
  };
}

function approvedActionPlanDocument() {
  return {
    ...actionPlanDocument(),
    record_version: 2,
    status: "approved",
    last_decision_by_actor_ref: "subject_demo_user",
    last_decision_at: "2026-07-16T02:05:00Z",
  };
}

function confirmationDecisionDocument() {
  return {
    schema_version: "workflow_http_tool_confirmation_decision.v1",
    confirmation_id: "wtcd_abcdefghijklmnop",
    plan_id: "wtap_abcdefghijklmnop",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: "app_flow_copilot",
    draft_id: "draft_http_tool_review",
    draft_version: 4,
    node_id: "node_http_tool",
    tool_id: "workflow.http.reviewed-json-read.v1",
    tool_version: 1,
    tool_plan_digest: digest("c"),
    outcome: "approve",
    decided_by_actor_ref: "subject_demo_user",
    actor_source: "human",
    decided_at: "2026-07-16T02:05:00Z",
    reason_code: "workflow_tool_confirmation_approved",
    expected_record_version: 1,
    resulting_record_version: 2,
    audit_ref: "audit_workflow_tool_confirmation",
  };
}

function successEnvelope(actionPlan: object, confirmationDecision: object | null = null) {
  return {
    request_id: "request_workflow_tool_action",
    workspace_id: "workspace_demo",
    application_id: "app_flow_copilot",
    action_plan: actionPlan,
    confirmation_decision: confirmationDecision,
    failure_code: null,
    failure_summary: "",
    audit_ref: "audit_workflow_tool_action",
  };
}

function failureEnvelope(failureCode: string) {
  return {
    request_id: "request_workflow_tool_action_conflict",
    workspace_id: "workspace_demo",
    application_id: "app_flow_copilot",
    action_plan: null,
    confirmation_decision: null,
    failure_code: failureCode,
    failure_summary: "The displayed plan version is stale.",
    audit_ref: "audit_workflow_tool_action_conflict",
  };
}

function jsonResponse(value: unknown, status = 200): Response {
  return new Response(JSON.stringify(value), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function digest(character: string): string {
  return `sha256:${character.repeat(64)}`;
}
