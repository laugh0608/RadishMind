import assert from "node:assert/strict";
import test from "node:test";

import {
  buildWorkflowExecutorV0Draft,
  evaluateWorkflowExecutorEligibility,
  readWorkflowRunDevRecord,
  startWorkflowRunDevRecord,
} from "../src/features/control-plane-read/workflowExecutorConsumer.ts";
import { workflowSavedDraftRequestedCapabilities } from "../src/features/control-plane-read/workflowSavedDraftCapabilityPolicy.ts";

const executorConfig = {
  mode: "dev_workflow_executor_http" as const,
  baseUrl: "http://127.0.0.1:7000",
  workspaceId: "workspace_demo",
  tenantRef: "tenant_demo",
  subjectRef: "subject_demo_user",
};

test("executor v0 draft builder produces a saved-draft-compatible bounded graph", () => {
  const draft = buildWorkflowExecutorV0Draft(sourceDraft(), 2);

  assert.equal(draft.draftId, "draft_app_flow_copilot_executor_v0_02");
  assert.equal(draft.executionProfile, "executor_v0");
  assert.deepEqual(draft.nodes.map((node) => node.nodeType), ["prompt", "llm", "output"]);
  assert.equal(draft.edges.length, 2);
  assert.equal(draft.nodes.every((node) => !node.toolRef && !node.ragRef && !node.requiresConfirmation), true);
  assert.deepEqual(workflowSavedDraftRequestedCapabilities(draft), []);
  assert.deepEqual(
    workflowSavedDraftRequestedCapabilities({ ...draft, executionProfile: "review_only" }),
    ["publish", "run", "confirmation_decision", "writeback", "replay"],
  );
});

test("executor eligibility requires the exact saved clean graph", () => {
  const draft = buildWorkflowExecutorV0Draft(sourceDraft(), 1);
  const savedState = {
    status: "saved_dev_record" as const,
    mode: "dev_saved_draft_http" as const,
    sourceLabel: "saved dev record",
    summary: "saved",
    failureCode: null,
    currentDraftVersion: 3,
    conflictDraftVersion: null,
    auditRef: "audit_saved",
    requestId: "req_saved",
  };

  const eligible = evaluateWorkflowExecutorEligibility(draft, savedState, false);
  assert.equal(eligible.eligible, true);
  assert.equal(eligible.savedDraftVersion, 3);

  const dirty = evaluateWorkflowExecutorEligibility(draft, savedState, true);
  assert.equal(dirty.eligible, false);
  assert.equal(dirty.reasons.some((reason) => reason.code === "unsaved_local_changes"), true);

  const unsafeDraft = {
    ...draft,
    nodes: draft.nodes.map((node, index) => index === 1 ? { ...node, toolRef: "tool:blocked" } : node),
  };
  const unsafe = evaluateWorkflowExecutorEligibility(unsafeDraft, savedState, false);
  assert.equal(unsafe.eligible, false);
  assert.equal(unsafe.reasons.some((reason) => reason.code === "executor_node_risk_blocked"), true);

  const cyclicDraft = {
    ...draft,
    edges: [
      ...draft.edges,
      {
        edgeId: "edge_output_prompt",
        fromNodeId: "node_executor_output",
        toNodeId: "node_executor_prompt",
        edgeKind: "context" as const,
        conditionSummary: "cycle",
      },
    ],
  };
  const cyclic = evaluateWorkflowExecutorEligibility(cyclicDraft, savedState, false);
  assert.equal(cyclic.eligible, false);
  assert.equal(
    cyclic.reasons.some((reason) =>
      reason.code === "executor_cycle" || reason.code === "executor_root_terminal_invalid"),
    true,
  );
});

test("executor HTTP consumer maps terminal record without retaining raw input", async (t) => {
  const draft = buildWorkflowExecutorV0Draft(sourceDraft(), 1);
  const rawInput = "private input appears only in the outbound request";
  const requests: Array<{ url: string; init: RequestInit }> = [];
  const originalFetch = globalThis.fetch;
  t.after(() => {
    globalThis.fetch = originalFetch;
  });
  globalThis.fetch = async (input, init = {}) => {
    requests.push({ url: String(input), init });
    return new Response(JSON.stringify(successEnvelope()), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    });
  };

  const started = await startWorkflowRunDevRecord(draft, rawInput, {}, executorConfig, { model: "mock" });
  assert.equal(started.status, "succeeded");
  assert.equal(started.record?.output, "reviewable result");
  assert.equal(JSON.stringify(started).includes(rawInput), false);
  assert.equal(requests[0]?.url.endsWith(`/v1/user-workspace/workflow-drafts/${draft.draftId}/runs`), true);
  assert.equal(String(requests[0]?.init.body).includes(rawInput), true);
  const headers = requests[0]?.init.headers as Record<string, string>;
  assert.equal(headers["X-RadishMind-Dev-Read-Scopes"], "workflow_drafts:read,workflow_runs:execute,workflow_runs:read");

  const reloaded = await readWorkflowRunDevRecord(started.record!, executorConfig);
  assert.equal(reloaded.status, "succeeded");
  assert.equal(requests[1]?.url.includes("/v1/user-workspace/workflow-runs/run_executor_test?"), true);
});

function sourceDraft() {
  return {
    draftId: "draft_source",
    templateRef: "wf_source",
    label: "RadishFlow advisory",
    applicationRef: "app_flow_copilot",
    workflowDefinitionId: "wf_radishflow_copilot_latest",
    providerProfileRef: "provider:mock",
    summary: "source",
    nodes: [],
    edges: [],
    designerLayout: {
      source: "workflow_node_designer" as const,
      persistence: "ui_only" as const,
      nodePositions: [],
    },
    readiness: [],
    risks: [],
    blockedCapabilities: [],
    routeMetadata: {
      sourceRouteId: "workflow-definition-summary-list-route" as const,
      draftRouteId: "workflow-draft-designer-offline-draft" as const,
      routePath: "/v1/user-workspace/workflow-definitions" as const,
      requestId: "req_source",
      auditRef: "audit_source",
    },
    localOnlyInteraction: "inspect_only" as const,
    executionProfile: "review_only" as const,
  };
}

function successEnvelope() {
  return {
    request_id: "req_executor_test",
    workspace_id: "workspace_demo",
    application_id: "app_flow_copilot",
    failure_code: null,
    failure_summary: "",
    audit_ref: "audit_executor_test",
    run: {
      schema_version: "workflow_run_record.v0",
      run_id: "run_executor_test",
      draft_id: "draft_app_flow_copilot_executor_v0_01",
      draft_version: 1,
      workspace_id: "workspace_demo",
      application_id: "app_flow_copilot",
      status: "succeeded",
      failure_code: "",
      failure_summary: "",
      started_at: "2026-07-11T00:00:00Z",
      completed_at: "2026-07-11T00:00:01Z",
      input_bytes: 32,
      condition_node_ids: [],
      requested_model: "mock",
      selected_provider: "mock",
      selected_profile: "",
      selected_model: "mock",
      upstream_model: "mock",
      selection_source: "configured_default",
      nodes: [
        {
          node_id: "node_executor_prompt",
          node_type: "prompt",
          label: "Prepare advisory prompt",
          status: "succeeded",
          started_at: "2026-07-11T00:00:00Z",
          completed_at: "2026-07-11T00:00:00Z",
          duration_ms: 0,
          predecessor_node_ids: [],
          provider_ref: "",
          output_preview: "input packet accepted; raw input not retained",
          failure_code: "",
        },
      ],
      output: "reviewable result",
      request_id: "req_executor_test",
      audit_ref: "audit_executor_test",
      side_effects: {
        provider_calls: 1,
        tool_calls: 0,
        confirmation_calls: 0,
        business_writes: 0,
        replay_writes: 0,
      },
    },
  };
}
