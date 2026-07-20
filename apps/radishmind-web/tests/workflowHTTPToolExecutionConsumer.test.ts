import assert from "node:assert/strict";
import test from "node:test";

import {
  executeWorkflowHTTPToolActionPlan,
  initialWorkflowHTTPToolExecutionState,
} from "../src/features/control-plane-read/workflowHTTPToolExecutionConsumer.ts";
import type {
  WorkflowHTTPToolActionConsumerConfig,
  WorkflowHTTPToolActionPlan,
} from "../src/features/control-plane-read/workflowHTTPToolActionConsumer.ts";

const baseConfig: WorkflowHTTPToolActionConsumerConfig = {
  mode: "dev_workflow_http_tool_http",
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
const executionConfig: WorkflowHTTPToolActionConsumerConfig = {
  ...baseConfig,
  scopeGrants: [
    ...baseConfig.scopeGrants,
    "workflow_tool_actions:execute",
    "workflow_runs:execute",
  ],
};

test("Workflow HTTP Tool execution stays offline and scope denial performs zero fetches", async (t) => {
  let fetchCount = 0;
  const originalFetch = globalThis.fetch;
  t.after(() => { globalThis.fetch = originalFetch; });
  globalThis.fetch = async () => {
    fetchCount += 1;
    throw new Error("execution must remain closed");
  };

  const disabled = { ...baseConfig, mode: "disabled" as const };
  assert.equal(initialWorkflowHTTPToolExecutionState(disabled).status, "disabled");
  assert.equal((await executeWorkflowHTTPToolActionPlan(disabled, approvedPlan(), boundedInput())).status, "disabled");
  const denied = await executeWorkflowHTTPToolActionPlan(baseConfig, approvedPlan(), boundedInput());
  assert.equal(denied.failureCode, "workflow_run_scope_denied");
  assert.equal(fetchCount, 0);
});

test("approved plan executes once with exact scopes and maps a redacted v2 terminal record", async (t) => {
  const requests: Array<{ url: string; init: RequestInit }> = [];
  const originalFetch = globalThis.fetch;
  t.after(() => { globalThis.fetch = originalFetch; });
  globalThis.fetch = async (input, init = {}) => {
    requests.push({ url: String(input), init });
    return jsonResponse(successEnvelope());
  };
  const rawInput = "private execution input must remain request-only";

  const result = await executeWorkflowHTTPToolActionPlan(
    executionConfig,
    approvedPlan(),
    { inputText: rawInput, model: "mock", temperature: 0 },
  );

  assert.equal(result.status, "succeeded");
  assert.equal(result.actionPlan?.status, "consumed");
  assert.equal(result.run?.schemaVersion, "workflow_run_record.v2");
  assert.equal(result.run?.toolAttempt?.status, "succeeded");
  assert.equal(result.run?.sideEffects.toolCalls, 1);
  assert.equal(result.run?.sideEffects.confirmationCalls, 1);
  assert.equal(result.run?.sideEffects.businessWrites, 0);
  assert.equal(JSON.stringify(result).includes(rawInput), false);
  assert.equal(requests.length, 1);
  assert.equal(requests[0]?.url.endsWith("/v1/user-workspace/workflow-tool-action-plans/wtap_abcdefghijklmnop/executions"), true);
  assert.deepEqual(JSON.parse(String(requests[0]?.init.body)), {
    workspace_id: "workspace_demo",
    application_id: "app_flow_copilot",
    expected_record_version: 2,
    input_text: rawInput,
    model: "mock",
    temperature: 0,
  });
  const headers = requests[0]?.init.headers as Record<string, string>;
  assert.equal(
    headers["X-RadishMind-Dev-Read-Scopes"],
    "workflow_tool_actions:execute,workflow_runs:execute,workflow_drafts:read",
  );
});

test("unsafe execution evidence is rejected and transport failure is never retried", async (t) => {
  const originalFetch = globalThis.fetch;
  t.after(() => { globalThis.fetch = originalFetch; });
  const unsafe = successEnvelope() as any;
  unsafe.run.tool_attempt.output_projection.url = "https://private.invalid";
  globalThis.fetch = async () => jsonResponse(unsafe);
  const rejected = await executeWorkflowHTTPToolActionPlan(executionConfig, approvedPlan(), boundedInput());
  assert.equal(rejected.failureCode, "workflow_tool_store_contract_mismatch");
  assert.equal(rejected.run, null);

  let attempts = 0;
  globalThis.fetch = async () => {
    attempts += 1;
    throw new Error("network unavailable");
  };
  const unavailable = await executeWorkflowHTTPToolActionPlan(executionConfig, approvedPlan(), boundedInput());
  assert.equal(unavailable.failureCode, "workflow_tool_store_unavailable");
  assert.equal(attempts, 1);
});

function boundedInput() {
  return { inputText: "Review the approved resource.", model: "mock", temperature: 0 };
}

function approvedPlan(): WorkflowHTTPToolActionPlan {
  return {
    schemaVersion: "workflow_http_tool_action_plan.v1",
    planId: "wtap_abcdefghijklmnop",
    recordVersion: 2,
    tenantRef: "tenant_demo",
    workspaceId: "workspace_demo",
    applicationId: "app_flow_copilot",
    draftId: "draft_http_tool_review",
    draftVersion: 4,
    nodeId: "node_http_tool",
    toolId: "workflow.http.reviewed-json-read.v1",
    toolVersion: 1,
    definitionDigest: digest("a"),
    profileId: "workflow_http_profile_reviewed_json_read_dev_v1",
    profileVersion: 1,
    profileDigest: digest("b"),
    method: "GET",
    targetPolicyKey: "reviewed_json_read_dev",
    publicArguments: { resourceKey: "catalog/review:item-1", locale: "zh-CN" },
    outputFields: ["resource_key", "title", "summary", "updated_at"],
    outputSchemaDigest: digest("d"),
    credentialPolicy: "none",
    timeoutMs: 5000,
    maxResponseBytes: 65536,
    maxOutputBytes: 16384,
    plannedByActorRef: "subject_demo_user",
    createdAt: "2026-07-17T02:00:00Z",
    expiresAt: "2026-07-17T02:15:00Z",
    toolPlanDigest: digest("c"),
    status: "approved",
    lastDecisionByActorRef: "subject_demo_user",
    lastDecisionAt: "2026-07-17T02:05:00Z",
    auditRef: "audit_workflow_tool_plan",
  };
}

function actionPlanDocument() {
  const plan = approvedPlan();
  return {
    schema_version: plan.schemaVersion,
    plan_id: plan.planId,
    record_version: 3,
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
    status: "consumed",
    last_decision_by_actor_ref: plan.lastDecisionByActorRef,
    last_decision_at: plan.lastDecisionAt,
    audit_ref: plan.auditRef,
  };
}

function successEnvelope() {
  return {
    request_id: "request_workflow_tool_execution",
    workspace_id: "workspace_demo",
    application_id: "app_flow_copilot",
    action_plan: actionPlanDocument(),
    run: {
      schema_version: "workflow_run_record.v2",
      record_version: 2,
      run_id: "run_0123456789abcdef",
      plan_id: "wtap_abcdefghijklmnop",
      confirmation_id: "wtcd_abcdefghijklmnop",
      tenant_ref: "tenant_demo",
      workspace_id: "workspace_demo",
      application_id: "app_flow_copilot",
      draft_id: "draft_http_tool_review",
      draft_version: 4,
      status: "succeeded",
      failure_code: null,
      failure_summary: "",
      started_at: "2026-07-17T02:05:01Z",
      completed_at: "2026-07-17T02:05:02Z",
      input_bytes: 29,
      requested_model: "mock",
      selected_provider: "mock",
      selected_profile: "provider_profile_mock",
      selected_model: "mock",
      upstream_model: "mock",
      selection_source: "configured_default",
      nodes: [{
        node_id: "node_http_tool",
        node_type: "http_tool",
        label: "Reviewed JSON read",
        status: "succeeded",
        started_at: "2026-07-17T02:05:01Z",
        completed_at: "2026-07-17T02:05:02Z",
        duration_ms: 100,
        predecessor_node_ids: [],
        provider_ref: "",
        output_preview: "reviewed item projected",
        failure_code: null,
      }],
      tool_attempt: {
        attempt_id: "wtea_abcdefghijklmnop",
        node_id: "node_http_tool",
        tool_id: "workflow.http.reviewed-json-read.v1",
        definition_digest: digest("a"),
        profile_id: "workflow_http_profile_reviewed_json_read_dev_v1",
        profile_digest: digest("b"),
        tool_plan_digest: digest("c"),
        confirmation_id: "wtcd_abcdefghijklmnop",
        status: "succeeded",
        claimed_at: "2026-07-17T02:05:01Z",
        completed_at: "2026-07-17T02:05:02Z",
        http_status_class: "2xx",
        response_bytes: 128,
        duration_ms: 100,
        output_projection: {
          resource_key: "catalog/review:item-1",
          title: "Reviewed item",
          summary: "Safe projected response",
          updated_at: "2026-07-17T02:00:00Z",
        },
        failure_code: null,
      },
      output: "reviewable result",
      request_id: "request_workflow_tool_execution",
      audit_ref: "audit_workflow_tool_execution",
      actor_ref: "subject_demo_user",
      side_effects: {
        provider_calls: 1,
        tool_calls: 1,
        confirmation_calls: 1,
        business_writes: 0,
        replay_writes: 0,
      },
      diagnostic: {
        failure_boundary: null,
        failure_stage: "",
        failed_node_id: null,
        last_completed_node_id: "node_http_tool",
        terminal_write_state: "stored",
        gateway_failure_category: "none",
        tool_failure_category: "none",
        summary: "",
        recommended_review_action: "",
        observed_at: "2026-07-17T02:05:02Z",
      },
    },
    failure_code: null,
    failure_summary: "",
    audit_ref: "audit_workflow_tool_execution",
  };
}

function jsonResponse(value: unknown): Response {
  return new Response(JSON.stringify(value), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
}

function digest(character: string): string {
  return `sha256:${character.repeat(64)}`;
}
