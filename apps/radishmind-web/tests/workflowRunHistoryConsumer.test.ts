import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";
import { legacyActionSafetyFixture } from "./actionSafetyFixture.ts";

import {
  EMPTY_WORKFLOW_RUN_HISTORY_FILTER,
  initialWorkflowRunHistoryState,
  isWorkflowRunComparisonEligible,
  listWorkflowRunHistory,
  resolveWorkflowRunHistoryConfig,
} from "../src/features/control-plane-read/workflowRunHistoryConsumer.ts";
import type { WorkflowExecutorConsumerConfig } from "../src/features/control-plane-read/workflowExecutorConsumer.ts";
import { parseWorkflowRunRecordDocument } from "../src/features/control-plane-read/workflowRunRecordConsumer.ts";

const offline: WorkflowExecutorConsumerConfig = { mode: "disabled", baseUrl: "http://127.0.0.1:7000", workspaceId: "workspace_demo", tenantRef: "tenant_demo", subjectRef: "subject_demo" };
const live: WorkflowExecutorConsumerConfig = { ...offline, mode: "dev_workflow_executor_http" };

test("workflow run history uses an independent read capability and keeps executor compatibility", () => {
  assert.deepEqual(
    resolveWorkflowRunHistoryConfig({
      VITE_RADISHMIND_WORKFLOW_RUN_HISTORY_SOURCE: "dev-workflow-run-history-http",
      VITE_RADISHMIND_WORKFLOW_RUN_HISTORY_BASE_URL: "http://127.0.0.1:7100/",
      VITE_RADISHMIND_WORKFLOW_RUN_HISTORY_WORKSPACE_ID: "workspace_prompt",
      VITE_RADISHMIND_DEV_READ_TENANT_REF: "tenant_prompt",
      VITE_RADISHMIND_DEV_READ_SUBJECT_REF: "subject_prompt",
    }),
    {
      mode: "dev_workflow_executor_http",
      baseUrl: "http://127.0.0.1:7100",
      workspaceId: "workspace_prompt",
      tenantRef: "tenant_prompt",
      subjectRef: "subject_prompt",
      diagnosticsDevEnabled: false,
    },
  );
  assert.equal(
    resolveWorkflowRunHistoryConfig({
      VITE_RADISHMIND_WORKFLOW_EXECUTOR_SOURCE: "dev-workflow-executor-http",
    }).mode,
    "dev_workflow_executor_http",
  );
  assert.equal(
    resolveWorkflowRunHistoryConfig({
      VITE_RADISHMIND_WORKFLOW_RUN_HISTORY_SOURCE: "disabled",
      VITE_RADISHMIND_WORKFLOW_EXECUTOR_SOURCE: "dev-workflow-executor-http",
    }).mode,
    "disabled",
  );
});

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
  try { await assert.rejects(() => listWorkflowRunHistory("app_demo", live, EMPTY_WORKFLOW_RUN_HISTORY_FILTER), /incompatible side effect/); }
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

test("workflow run history maps v2 confirmation, attempt, and outcome evidence", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => new Response(JSON.stringify({
    request_id: "request_tool_history", workspace_id: "workspace_demo", application_id: "app_demo",
    runs: [{
      schema_version: "workflow_run_record.v2", record_version: 2,
      run_id: "run_0123456789abcdef", plan_id: "wtap_abcdefghijklmnop",
      confirmation_id: "wtcd_abcdefghijklmnop", tool_attempt_status: "outcome_unknown",
      draft_id: "draft_tool", draft_version: 4, workspace_id: "workspace_demo", application_id: "app_demo",
      status: "outcome_unknown", failure_code: "workflow_tool_outcome_unknown",
      started_at: "2026-07-17T10:00:00Z", completed_at: "2026-07-17T10:00:30Z", duration_ms: 30000,
      selected_provider: "mock", selected_profile: "provider_profile_mock", selected_model: "mock",
      request_id: "request_tool_run", audit_ref: "audit_tool_run", stale_running: false,
      failure_boundary: "tool_transport", failed_node_id: "node_http_tool",
      last_completed_node_id: "node_prompt", gateway_failure_category: "none",
      tool_failure_category: "outcome_unknown", recommended_review_action: "review_tool_outcome",
      action_safety: legacyActionSafetyFixture("workflow_run", "run_0123456789abcdef", 2),
      side_effects: { provider_calls: 0, tool_calls: 1, confirmation_calls: 1, business_writes: 0, replay_writes: 0 },
    }],
    next_cursor: "", has_more: false, failure_code: null, failure_summary: "", audit_ref: "audit_tool_history",
  }), { status: 200 });
  try {
    const result = await listWorkflowRunHistory("app_demo", live, { ...EMPTY_WORKFLOW_RUN_HISTORY_FILTER, status: "outcome_unknown" });
    const run = result.runs[0];
    assert.equal(run?.schemaVersion, "workflow_run_record.v2");
    assert.equal(run?.planId, "wtap_abcdefghijklmnop");
    assert.equal(run?.confirmationId, "wtcd_abcdefghijklmnop");
    assert.equal(run?.toolAttemptStatus, "outcome_unknown");
    assert.equal(run?.toolFailureCategory, "outcome_unknown");
    assert.equal(run?.sideEffects.toolCalls, 1);
    assert.equal(run?.sideEffects.businessWrites, 0);
  } finally { globalThis.fetch = originalFetch; }
});

test("workflow run history maps v5 definition authority and exact source filters", async () => {
  const originalFetch = globalThis.fetch;
  const digest = `sha256:${"a".repeat(64)}`;
  globalThis.fetch = async (input) => {
    const url = new URL(String(input));
    assert.equal(url.searchParams.get("execution_source_kind"), "workflow_definition");
    assert.equal(url.searchParams.get("execution_source_id"), "definition_demo");
    assert.equal(url.searchParams.get("execution_source_version"), "2");
    return new Response(JSON.stringify({
      request_id: "request_definition_history", workspace_id: "workspace_demo", application_id: "app_demo",
      runs: [{ schema_version: "workflow_run_record.v5", record_version: 2, run_id: "run_definition_history",
        draft_id: "", draft_version: 0, execution_kind: "workflow_definition_execution",
        execution_source_kind: "workflow_definition", execution_source_id: "definition_demo", execution_source_version: 2,
        execution_profile: "workflow_definition_executor_v1", definition_digest: digest, activation_pointer_version: 3,
        source_draft_id: "draft_definition_source", source_draft_version: 4, source_draft_digest: digest,
        workspace_id: "workspace_demo", application_id: "app_demo", status: "succeeded", failure_code: "",
        started_at: "2026-07-19T10:00:00Z", completed_at: "2026-07-19T10:00:01Z", duration_ms: 1000,
        selected_provider: "mock", selected_profile: "", selected_model: "mock", request_id: "request_run", audit_ref: "audit_run",
        stale_running: false, side_effects: { retrieval_calls: 0, provider_calls: 1, tool_calls: 0, confirmation_calls: 0, business_writes: 0, replay_writes: 0 } }],
      next_cursor: "", has_more: false, failure_code: null, failure_summary: "", audit_ref: "audit_definition_history",
    }), { status: 200 });
  };
  try {
    const result = await listWorkflowRunHistory("app_demo", live, { ...EMPTY_WORKFLOW_RUN_HISTORY_FILTER, executionSourceKind: "workflow_definition", executionSourceId: "definition_demo", executionSourceVersion: 2 });
    const run = result.runs[0];
    assert.equal(run?.schemaVersion, "workflow_run_record.v5");
    assert.equal(run?.executionProfile, "workflow_definition_executor_v1");
    assert.equal(run?.activationPointerVersion, 3);
    assert.equal(run?.sourceDraftId, "draft_definition_source");
  } finally { globalThis.fetch = originalFetch; }
});

test("workflow run history maps metadata-only structured Definition v8", async () => {
  const originalFetch = globalThis.fetch;
  const definitionDigest = `sha256:${"d".repeat(64)}`;
  const contractDigest = `sha256:${"e".repeat(64)}`;
  globalThis.fetch = async () => new Response(JSON.stringify({
    request_id: "request_structured_definition_history", workspace_id: "workspace_demo", application_id: "app_demo",
    runs: [{ schema_version: "workflow_run_record.v8", record_version: 8, run_id: "run_structured_history",
      draft_id: "", draft_version: 0, execution_kind: "workflow_definition_execution",
      execution_source_kind: "workflow_definition", execution_source_id: "definition_structured",
      execution_source_version: 1, execution_profile: "workflow_definition_executor_v2",
      input_contract_id: "contract_structured", input_contract_digest: contractDigest,
      definition_digest: definitionDigest, activation_pointer_version: 3,
      source_draft_id: "draft_structured_source", source_draft_version: 1, source_draft_digest: definitionDigest,
      workspace_id: "workspace_demo", application_id: "app_demo", status: "succeeded", failure_code: "",
      started_at: "2026-08-11T10:00:00Z", completed_at: "2026-08-11T10:00:01Z", duration_ms: 1000,
      selected_provider: "mock", selected_profile: "", selected_model: "mock",
      request_id: "request_structured_run", audit_ref: "audit_structured_run", stale_running: false,
      side_effects: { provider_calls: 1, tool_calls: 0, confirmation_calls: 0, business_writes: 0, replay_writes: 0 } }],
    next_cursor: "", has_more: false, failure_code: null, failure_summary: "", audit_ref: "audit_structured_definition_history",
  }), { status: 200 });
  try {
    const result = await listWorkflowRunHistory("app_demo", live, EMPTY_WORKFLOW_RUN_HISTORY_FILTER);
    const run = result.runs[0];
    assert.equal(run?.schemaVersion, "workflow_run_record.v8");
    assert.equal(run?.executionProfile, "workflow_definition_executor_v2");
    assert.equal(run?.inputContractId, "contract_structured");
    assert.equal(run?.inputContractDigest, contractDigest);
    assert.equal(run?.definitionDigest, definitionDigest);
  } finally { globalThis.fetch = originalFetch; }
});

test("workflow run history and detail recognize v9 as side-effectful Definition evidence", async () => {
  const originalFetch = globalThis.fetch;
  const definitionDigest = `sha256:${"b".repeat(64)}`;
  globalThis.fetch = async () => new Response(JSON.stringify({
    request_id: "request_definition_tool_history", workspace_id: "workspace_demo", application_id: "app_demo",
    runs: [{ schema_version: "workflow_run_record.v9", record_version: 2, run_id: "run_0123456789abcdef",
      plan_id: "wtap_abcdefghijklmnop", confirmation_id: "wtcd_abcdefghijklmnop", tool_attempt_status: "succeeded",
      draft_id: "", draft_version: 0, execution_kind: "workflow_definition_http_tool_execution",
      execution_source_kind: "workflow_definition", execution_source_id: "definition_demo", execution_source_version: 3,
      execution_profile: "workflow_definition_http_tool_v1", definition_digest: definitionDigest,
      activation_pointer_version: 4, source_draft_id: "draft_definition_source", source_draft_version: 3,
      source_draft_digest: definitionDigest, workspace_id: "workspace_demo", application_id: "app_demo",
      status: "succeeded", failure_code: "", started_at: "2026-08-15T10:00:00Z",
      completed_at: "2026-08-15T10:00:02Z", duration_ms: 2000, selected_provider: "mock",
      selected_profile: "default", selected_model: "mock", request_id: "request_definition_tool_run",
      audit_ref: "audit_definition_tool_run", stale_running: false, failure_boundary: "",
      failed_node_id: "", last_completed_node_id: "node_output", gateway_failure_category: "none",
      tool_failure_category: "none", recommended_review_action: "",
      action_safety: legacyActionSafetyFixture("workflow_run", "run_0123456789abcdef", 2),
      side_effects: { provider_calls: 1, tool_calls: 1, confirmation_calls: 1, business_writes: 0, replay_writes: 0 } }],
    next_cursor: "", has_more: false, failure_code: null, failure_summary: "", audit_ref: "audit_definition_tool_history",
  }), { status: 200 });
  try {
    const result = await listWorkflowRunHistory("app_demo", live, EMPTY_WORKFLOW_RUN_HISTORY_FILTER);
    const run = result.runs[0]!;
    assert.equal(run.schemaVersion, "workflow_run_record.v9");
    assert.equal(run.executionProfile, "workflow_definition_http_tool_v1");
    assert.equal(run.planId, "wtap_abcdefghijklmnop");
    assert.equal(run.sideEffects.toolCalls, 1);
    assert.equal(run.activationPointerVersion, 4);
    assert.equal(isWorkflowRunComparisonEligible(run), false);
  } finally { globalThis.fetch = originalFetch; }

  const fixture = JSON.parse(readFileSync(
    new URL("../../../scripts/checks/fixtures/workflow-http-tool-contracts-v1.json", import.meta.url),
    "utf8",
  ));
  const detail = fixture.positive.run_record_v9;
  assert.equal(parseWorkflowRunRecordDocument(detail)?.schemaVersion, "workflow_run_record.v9");
  assert.equal(parseWorkflowRunRecordDocument({ ...detail, raw_output: "private" }), null);
  assert.equal(parseWorkflowRunRecordDocument({
    ...detail,
    tool_attempt: { ...detail.tool_attempt, output_projection: { summary: "https://private.invalid/raw" } },
  }), null);
});

test("workflow run history and detail recognize strict metadata-only v6", async () => {
  const originalFetch = globalThis.fetch;
  const digest = `sha256:${"b".repeat(64)}`;
  globalThis.fetch = async (input) => {
    const url = new URL(String(input));
    assert.equal(url.searchParams.get("execution_source_kind"), "prompt_application_template");
    return new Response(JSON.stringify({
      request_id: "request_prompt_history", workspace_id: "workspace_demo", application_id: "app_demo",
      runs: [{ schema_version: "workflow_run_record.v6", record_version: 2, run_id: "run_abcdefghijklmnop",
        draft_id: "", draft_version: 0, execution_kind: "prompt_application_invocation",
        execution_source_kind: "prompt_application_template", execution_source_id: "ptpl_abcdefghijklmnop",
        execution_source_version: 1, execution_profile: "prompt_application_invocation_v1",
        runtime_assignment_id: "ptra_abcdefghijklmnop", runtime_assignment_version: 1,
        publish_candidate_id: "candidate_prompt_demo", publish_review_version: 1,
        authority_digest: digest, prompt_template_digest: digest, variable_names_digest: digest,
        requested_protocol: "openai_chat_completions", selected_protocol: "openai_chat_completions",
        usage_state: "provider_reported", input_tokens: 5, output_tokens: 7, total_tokens: 12,
        workspace_id: "workspace_demo", application_id: "app_demo", status: "succeeded", failure_code: "",
        started_at: "2026-07-24T10:00:00Z", completed_at: "2026-07-24T10:00:01Z", duration_ms: 1000,
        selected_provider: "mock", selected_profile: "default", selected_model: "profile:local-dev",
        request_id: "request_prompt_run", audit_ref: "audit_prompt_run", stale_running: false,
        side_effects: { provider_calls: 1, tool_calls: 0, confirmation_calls: 0, business_writes: 0, replay_writes: 0 } }],
      next_cursor: "", has_more: false, failure_code: null, failure_summary: "", audit_ref: "audit_prompt_history",
    }), { status: 200 });
  };
  try {
    const result = await listWorkflowRunHistory("app_demo", live, {
      ...EMPTY_WORKFLOW_RUN_HISTORY_FILTER,
      executionSourceKind: "prompt_application_template",
      executionSourceId: "ptpl_abcdefghijklmnop",
      executionSourceVersion: 1,
    });
    const run = result.runs[0];
    assert.equal(run?.schemaVersion, "workflow_run_record.v6");
    assert.equal(run?.executionProfile, "prompt_application_invocation_v1");
    assert.equal(run?.authorityDigest, digest);
    assert.equal(run?.usageState, "provider_reported");
    assert.equal(run?.totalTokens, 12);
  } finally { globalThis.fetch = originalFetch; }

  const detail = promptRunV6Document(digest);
  const parsed = parseWorkflowRunRecordDocument(detail);
  assert.equal(parsed?.schemaVersion, "workflow_run_record.v6");
  assert.equal(parsed?.promptApplicationAuthority?.assignmentId, "ptra_abcdefghijklmnop");
  assert.equal(parsed?.promptApplicationAuthority?.templateId, "ptpl_abcdefghijklmnop");
  assert.equal(parsed?.promptApplicationAuthority?.templateVersion, 1);
  assert.equal(parsed?.variableNamesDigest, digest);
  assert.equal(parsed?.promptUsage?.totalTokens, 12);
  assert.equal(parsed?.output, "");
  assert.equal(parseWorkflowRunRecordDocument({ ...detail, rendered_messages: [] }), null);
});

test("workflow run history and detail recognize strict metadata-only Agent Copilot v7", async () => {
  const originalFetch = globalThis.fetch;
  const digest = `sha256:${"c".repeat(64)}`;
  globalThis.fetch = async (input) => {
    const url = new URL(String(input));
    assert.equal(url.searchParams.get("execution_source_kind"), "agent_copilot_profile");
    return new Response(JSON.stringify({
      request_id: "request_agent_history", workspace_id: "workspace_demo", application_id: "app_demo",
      runs: [{
        schema_version: "workflow_run_record.v7", record_version: 2, run_id: "run_agent_history",
        draft_id: "", draft_version: 0, execution_kind: "agent_copilot_suggestion",
        execution_source_kind: "agent_copilot_profile", execution_source_id: "acpf_aaaaaaaaaaaaaaaa",
        execution_source_version: 1, execution_profile: "agent_copilot_suggestion_v1",
        runtime_assignment_id: "acra_aaaaaaaaaaaaaaaa", runtime_assignment_version: 1,
        publish_candidate_id: "candidate_agent_demo", publish_review_version: 1,
        authority_digest: digest, profile_digest: digest, policy_digest: digest, allowed_tasks_digest: digest,
        project: "radishflow", task: "suggest_flowsheet_edits", locale: "zh-CN",
        response_status: "ok", response_digest: digest, answer_count: 1, issue_count: 1,
        action_count: 1, citation_count: 0, risk_level: "medium", requires_confirmation: true,
        requested_protocol: "openai_chat_completions", selected_protocol: "openai_chat_completions",
        usage_state: "provider_reported", input_tokens: 5, output_tokens: 7, total_tokens: 12,
        workspace_id: "workspace_demo", application_id: "app_demo", status: "succeeded", failure_code: "",
        started_at: "2026-07-25T10:00:00Z", completed_at: "2026-07-25T10:00:01Z", duration_ms: 1000,
        selected_provider: "mock", selected_profile: "default", selected_model: "profile:local-dev",
        request_id: "request_agent_run", audit_ref: "audit_agent_run", stale_running: false,
        action_safety: legacyActionSafetyFixture("workflow_run", "run_agent_history", 2),
        side_effects: { provider_calls: 1, tool_calls: 0, confirmation_calls: 0, business_writes: 0, replay_writes: 0 },
      }],
      next_cursor: "", has_more: false, failure_code: null, failure_summary: "", audit_ref: "audit_agent_history",
    }), { status: 200 });
  };
  try {
    const result = await listWorkflowRunHistory("app_demo", live, {
      ...EMPTY_WORKFLOW_RUN_HISTORY_FILTER,
      executionSourceKind: "agent_copilot_profile",
      executionSourceId: "acpf_aaaaaaaaaaaaaaaa",
      executionSourceVersion: 1,
    });
    const run = result.runs[0];
    assert.equal(run?.schemaVersion, "workflow_run_record.v7");
    assert.equal(run?.executionProfile, "agent_copilot_suggestion_v1");
    assert.equal(run?.profileDigest, digest);
    assert.equal(run?.task, "suggest_flowsheet_edits");
    assert.equal(run?.requiresConfirmation, true);
  } finally { globalThis.fetch = originalFetch; }

  const detail = agentRunV7Document(digest);
  const parsed = parseWorkflowRunRecordDocument(detail);
  assert.equal(parsed?.schemaVersion, "workflow_run_record.v7");
  assert.equal(parsed?.agentCopilotAuthority?.assignmentId, "acra_aaaaaaaaaaaaaaaa");
  assert.equal(parsed?.agentCopilotAuthority?.profileId, "acpf_aaaaaaaaaaaaaaaa");
  assert.equal(parsed?.agentTask, "suggest_flowsheet_edits");
  assert.equal(parsed?.agentResponseStatus, "ok");
  assert.equal(parsed?.agentRequiresConfirmation, true);
  assert.equal(parsed?.output, "");
  assert.equal(parseWorkflowRunRecordDocument({ ...detail, context: {} }), null);
});

function promptRunV6Document(digest: string) {
  return {
    schema_version: "workflow_run_record.v6", record_version: 2, run_id: "run_abcdefghijklmnop",
    tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: "app_demo",
    execution_kind: "prompt_application_invocation", execution_source_kind: "prompt_application_template",
    execution_source_id: "ptpl_abcdefghijklmnop", execution_source_version: 1,
    execution_profile: "prompt_application_invocation_v1",
    authority: {
      schema_version: "application_runtime_authority.v2", execution_profile: "prompt_application_invocation_v1",
      application_id: "app_demo", application_record_version: 1, application_lifecycle: "active",
      prompt_application: {
        assignment_id: "ptra_abcdefghijklmnop", assignment_version: 1, assignment_digest: digest,
        publish_candidate_id: "candidate_prompt_demo", publish_review_version: 1,
        draft_id: "draft_prompt_demo", draft_version: 2, draft_digest: digest,
        prompt_template_ref: { template_id: "ptpl_abcdefghijklmnop", template_version: 1, template_digest: digest },
        default_protocol: "openai_chat_completions", default_model: "profile:local-dev",
        protocol_policy_digest: digest, model_eligibility_digest: digest,
      },
      authority_digest: digest,
    },
    input_digest: digest, input_bytes: 64, variable_names: ["question", "tone"], variable_names_digest: digest,
    requested_protocol: "openai_chat_completions", selected_protocol: "openai_chat_completions",
    requested_model: "profile:local-dev", selected_provider: "mock", selected_profile: "default",
    selected_model: "profile:local-dev", upstream_model: "profile:local-dev", selection_source: "test_selection",
    status: "succeeded", failure_code: "", failure_summary: "", started_at: "2026-07-24T10:00:00Z",
    completed_at: "2026-07-24T10:00:01Z", output: "",
    usage: { state: "provider_reported", input_tokens: 5, output_tokens: 7, total_tokens: 12 },
    side_effects: { provider_calls: 1, tool_calls: 0, confirmation_calls: 0, business_writes: 0, replay_writes: 0 },
    diagnostic: { failure_boundary: "", failure_stage: "", terminal_write_state: "stored",
      gateway_failure_category: "none", summary: "", recommended_review_action: "", observed_at: "2026-07-24T10:00:01Z" },
    request_id: "request_prompt_run", audit_ref: "audit_prompt_run", actor_ref: "subject_demo",
  };
}

function agentRunV7Document(digest: string) {
  return {
    schema_version: "workflow_run_record.v7", record_version: 2, run_id: "run_abcdefghijklmnop",
    tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: "app_demo",
    execution_kind: "agent_copilot_suggestion", execution_source_kind: "agent_copilot_profile",
    execution_source_id: "acpf_aaaaaaaaaaaaaaaa", execution_source_version: 1,
    execution_profile: "agent_copilot_suggestion_v1",
    authority: {
      schema_version: "application_runtime_authority.v3", execution_profile: "agent_copilot_suggestion_v1",
      application_id: "app_demo", application_record_version: 1, application_lifecycle: "active",
      agent_copilot: {
        assignment_id: "acra_aaaaaaaaaaaaaaaa", assignment_version: 1, assignment_digest: digest,
        publish_candidate_id: "candidate_agent_demo", publish_review_version: 1,
        draft_id: "draft_agent_demo", draft_version: 2, draft_digest: digest,
        agent_copilot_profile_ref: {
          profile_id: "acpf_aaaaaaaaaaaaaaaa", profile_version: 1,
          profile_digest: digest, policy_digest: digest,
        },
        project: "radishflow", allowed_tasks_digest: digest,
        default_protocol: "openai_chat_completions", default_model: "profile:local-dev",
        protocol_policy_digest: digest, model_eligibility_digest: digest,
      },
      authority_digest: digest,
    },
    project: "radishflow", task: "suggest_flowsheet_edits", locale: "zh-CN",
    input_digest: digest, input_bytes: 128, context_bytes: 64, artifact_count: 0, artifact_bytes: 0,
    requested_protocol: "openai_chat_completions", selected_protocol: "openai_chat_completions",
    requested_model: "profile:local-dev", selected_provider: "mock", selected_profile: "default",
    selected_model: "profile:local-dev", upstream_model: "profile:local-dev", selection_source: "test_selection",
    response_status: "ok", response_digest: digest, answer_count: 1, issue_count: 1,
    action_count: 1, citation_count: 0, risk_level: "medium", requires_confirmation: true,
    status: "succeeded", failure_code: "", failure_summary: "", started_at: "2026-07-25T10:00:00Z",
    completed_at: "2026-07-25T10:00:01Z", output: "",
    usage: { state: "provider_reported", input_tokens: 5, output_tokens: 7, total_tokens: 12 },
    side_effects: { retrieval_calls: 0, provider_calls: 1, tool_calls: 0, confirmation_calls: 0, business_writes: 0, replay_writes: 0 },
    diagnostic: { failure_boundary: "", failure_stage: "", terminal_write_state: "stored",
      gateway_failure_category: "none", summary: "", recommended_review_action: "", observed_at: "2026-07-25T10:00:01Z" },
    request_id: "request_agent_run", audit_ref: "audit_agent_run", actor_ref: "subject_demo",
  };
}
