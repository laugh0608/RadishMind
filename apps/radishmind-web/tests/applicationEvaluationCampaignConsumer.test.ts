import assert from "node:assert/strict";
import test from "node:test";

import {
  APPLICATION_EVALUATION_PROFILES,
  applicationEvaluationPlanTemplate,
  createApplicationEvaluationPlan,
  executeApplicationEvaluationCampaign,
  listApplicationEvaluationCampaigns,
  listApplicationEvaluationPlans,
  materializeApplicationEvaluationHandoff,
  parseApplicationEvaluationPlanDraft,
  previewApplicationEvaluationPair,
  readApplicationEvaluationPlanVersion,
  reviseApplicationEvaluationPlan,
  serializeApplicationEvaluationItems,
  serializeApplicationEvaluationTarget,
  type ApplicationEvaluationCampaign,
  type ApplicationEvaluationConfig,
  type ApplicationEvaluationPlan,
  type ApplicationEvaluationPlanVersion,
} from "../src/features/control-plane-read/applicationEvaluationCampaignConsumer.ts";

const digest = `sha256:${"a".repeat(64)}`;
const config: ApplicationEvaluationConfig = {
  mode: "dev_application_evaluation_http",
  baseUrl: "http://platform.test",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  environment: "test",
  applicationId: "app_flow_copilot",
  subjectRef: "subject_demo_user",
};

test("Application Evaluation list owners remain offline without a request", async () => {
  const originalFetch = globalThis.fetch;
  let calls = 0;
  globalThis.fetch = async () => { calls += 1; throw new Error("unexpected fetch"); };
  try {
    const plans = await listApplicationEvaluationPlans({ ...config, mode: "offline" });
    const campaigns = await listApplicationEvaluationCampaigns({ ...config, mode: "offline" });
    assert.equal(plans.failureCode, "application_evaluation_write_disabled");
    assert.deepEqual(plans.plans, []);
    assert.deepEqual(campaigns.campaigns, []);
    assert.equal(calls, 0);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Application Evaluation reads exact Plan scope, version, and permission headers", async () => {
  const originalFetch = globalThis.fetch;
  const requests: Array<{ url: URL; headers: Headers }> = [];
  globalThis.fetch = async (input, init) => {
    const url = new URL(String(input));
    requests.push({ url, headers: new Headers(init?.headers) });
    return jsonResponse(url.pathname.endsWith("/versions/1") ? planEnvelope() : planListEnvelope());
  };
  try {
    const listed = await listApplicationEvaluationPlans(config);
    const exact = await readApplicationEvaluationPlanVersion(config, "plan_demo", 1);
    assert.equal(listed.plans[0]?.latestPlanDigest, digest);
    assert.equal(exact.version?.items[0]?.workflowDefinition?.inputText, "Representative input");
    assert.equal(requests[0]?.url.searchParams.get("lifecycle_state"), "active");
    assert.equal(requests[1]?.url.pathname, "/v1/user-workspace/applications/app_flow_copilot/evaluation-plans/plan_demo/versions/1");
    for (const request of requests) {
      assert.equal(request.headers.get("X-RadishMind-Dev-Read-Scopes"), "application_evaluations:read");
      assert.equal(request.headers.get("X-RadishMind-Dev-Workflow-Workspace"), "workspace_demo");
      assert.equal(request.headers.get("X-RadishMind-Dev-Workflow-Application"), "app_flow_copilot");
      assert.equal(request.headers.get("X-RadishMind-Dev-Application-Evaluation-Environment"), "test");
    }
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Application Evaluation local Session transport includes cookies and omits every dev identity proof", async () => {
  const originalFetch = globalThis.fetch;
  let captured: RequestInit | undefined;
  globalThis.fetch = async (_input, init) => {
    captured = init;
    return jsonResponse(planListEnvelope());
  };
  try {
    await listApplicationEvaluationPlans({ ...config, authMode: "local_session_dev_test" });
    const headers = new Headers(captured?.headers);
    assert.equal(captured?.credentials, "include");
    assert.equal(captured?.cache, "no-store");
    assert.equal(headers.get("X-RadishMind-Active-Tenant"), "tenant_demo");
    assert.equal(headers.get("X-RadishMind-Active-Workspace"), "workspace_demo");
    assert.equal(headers.has("X-RadishMind-Dev-Read-Identity"), false);
    assert.equal(headers.has("X-RadishMind-Dev-Read-Tenant"), false);
    assert.equal(headers.has("X-RadishMind-Dev-Read-Subject"), false);
    assert.equal(headers.has("X-RadishMind-Dev-Read-Scopes"), false);
    assert.equal(headers.has("X-RadishMind-Dev-Read-Membership-Permissions"), false);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Application Evaluation create and revision send strict fixture bodies and CAS", async () => {
  const originalFetch = globalThis.fetch;
  const requests: Array<{ url: URL; headers: Headers; body: Record<string, unknown> }> = [];
  globalThis.fetch = async (input, init) => {
    requests.push({
      url: new URL(String(input)),
      headers: new Headers(init?.headers),
      body: JSON.parse(String(init?.body)) as Record<string, unknown>,
    });
    return jsonResponse(planEnvelope());
  };
  try {
    const draft = applicationEvaluationPlanTemplate("workflow_definition_executor_v1");
    await createApplicationEvaluationPlan(config, draft);
    await reviseApplicationEvaluationPlan(config, "plan_demo", 3, draft);
    assert.deepEqual(Object.keys(requests[0]?.body ?? {}).sort(), [
      "environment", "execution_profile", "items", "name", "target", "workspace_id",
    ]);
    assert.equal(requests[1]?.body.expected_version, 3);
    assert.equal(requests[1]?.url.pathname.endsWith("/plan_demo/revisions"), true);
    assert.equal(requests[1]?.headers.get("X-RadishMind-Dev-Read-Scopes"), "application_evaluations:write");
    assert.equal(JSON.stringify(requests[0]?.body).includes("secret"), false);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Application Evaluation campaign execution binds immutable authority, quota, and acknowledgements", async () => {
  const originalFetch = globalThis.fetch;
  let captured: { headers: Headers; body: Record<string, unknown> } | null = null;
  globalThis.fetch = async (_input, init) => {
    captured = { headers: new Headers(init?.headers), body: JSON.parse(String(init?.body)) as Record<string, unknown> };
    return jsonResponse(campaignEnvelope());
  };
  try {
    const result = await executeApplicationEvaluationCampaign(config, {
      plan: mappedPlan(),
      version: mappedVersion(),
      clientCampaignKey: "campaign_candidate",
      quotaAPIKeyId: "key_actor_owned",
    });
    assert.equal(result.campaign?.state, "succeeded");
    assert.deepEqual(captured?.body, {
      workspace_id: "workspace_demo",
      environment: "test",
      plan_id: "plan_demo",
      plan_version: 1,
      plan_digest: digest,
      expected_plan_record_version: 3,
      client_campaign_key: "campaign_candidate",
      quota_api_key_id: "key_actor_owned",
      acknowledge_sequential_execution: true,
      acknowledge_quota_consumption: true,
    });
    assert.equal(
      captured?.headers.get("X-RadishMind-Dev-Read-Scopes"),
      "application_evaluations:execute,workflow_definitions:read,workflow_runs:execute",
    );
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Application Evaluation pair preview and Handoff bind exact campaign versions", async () => {
  const originalFetch = globalThis.fetch;
  const requests: Array<{ url: URL; headers: Headers; body: Record<string, unknown> }> = [];
  globalThis.fetch = async (input, init) => {
    requests.push({
      url: new URL(String(input)),
      headers: new Headers(init?.headers),
      body: JSON.parse(String(init?.body)) as Record<string, unknown>,
    });
    return jsonResponse(pairEnvelope());
  };
  try {
    const baseline = mappedCampaign("campaign_baseline", 5);
    const candidate = mappedCampaign("campaign_candidate", 7);
    const preview = await previewApplicationEvaluationPair(config, baseline.campaignId, candidate.campaignId);
    const handoff = await materializeApplicationEvaluationHandoff(config, baseline, candidate);
    assert.equal(preview.review?.items[0]?.baselineRunId, "run_baseline");
    assert.equal(handoff.handoff?.caseRefs[0]?.version, 2);
    assert.deepEqual(requests[1]?.body, {
      workspace_id: "workspace_demo",
      environment: "test",
      baseline_campaign_id: "campaign_baseline",
      candidate_campaign_id: "campaign_candidate",
      expected_baseline_campaign_record_version: 5,
      expected_candidate_campaign_record_version: 7,
      acknowledge_evidence_materializing: true,
    });
    assert.equal(requests[0]?.headers.get("X-RadishMind-Dev-Read-Scopes"), "application_evaluations:read,workflow_runs:read");
    assert.equal(requests[1]?.headers.get("X-RadishMind-Dev-Read-Scopes"), "application_evaluations:read,workflow_evaluations:write,workflow_runs:read");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Application Evaluation fails closed on forbidden, extra, scope-drifted, and invalid nested responses", async () => {
  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async () => jsonResponse({ ...planListEnvelope(), secret: "never" });
    await assert.rejects(() => listApplicationEvaluationPlans(config), /forbidden field/);

    globalThis.fetch = async () => jsonResponse({ ...planListEnvelope(), debug: true });
    await assert.rejects(() => listApplicationEvaluationPlans(config), /contract mismatch/);

    globalThis.fetch = async () => jsonResponse({ ...planListEnvelope(), workspace_id: "workspace_other" });
    await assert.rejects(() => listApplicationEvaluationPlans(config), /contract mismatch/);

    const invalidPair = pairEnvelope();
    invalidPair.review.items[0].comparison = {};
    globalThis.fetch = async () => jsonResponse(invalidPair);
    await assert.rejects(() => previewApplicationEvaluationPair(config, "campaign_baseline", "campaign_candidate"), /pair envelope contract mismatch/);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Application Evaluation templates cover all five profiles and reject secret-bearing drafts", () => {
  for (const profile of APPLICATION_EVALUATION_PROFILES) {
    const template = applicationEvaluationPlanTemplate(profile);
    const parsed = parseApplicationEvaluationPlanDraft(
      template.name,
      profile,
      serializeApplicationEvaluationTarget(template.target),
      serializeApplicationEvaluationItems(template.items),
    );
    assert.equal(parsed.executionProfile, profile);
    assert.equal(parsed.items.length, 1);
    assert.equal([
      parsed.items[0]?.workflowDefinition,
      parsed.items[0]?.applicationRAG,
      parsed.items[0]?.promptApplication,
      parsed.items[0]?.agentCopilot,
    ].filter(Boolean).length, 1);
  }
  const template = applicationEvaluationPlanTemplate("prompt_application_invocation_v1");
  const secretItems = JSON.parse(serializeApplicationEvaluationItems(template.items)) as Array<Record<string, unknown>>;
  (secretItems[0]?.prompt_application as Record<string, unknown>).secret = "forbidden";
  assert.throws(() => parseApplicationEvaluationPlanDraft(
    template.name,
    template.executionProfile,
    serializeApplicationEvaluationTarget(template.target),
    JSON.stringify(secretItems),
  ), /forbidden field/);
});

test("Application Evaluation v2 decodes exact contracts and sends typed fixtures only", async () => {
  const originalFetch = globalThis.fetch;
  const bodies: Array<Record<string, unknown>> = [];
  globalThis.fetch = async (input, init) => {
    const url = new URL(String(input));
    if (init?.method === "POST") {
      bodies.push(JSON.parse(String(init.body)) as Record<string, unknown>);
    }
    return jsonResponse(url.pathname.endsWith("/versions/1") || init?.method === "POST"
      ? structuredPlanEnvelope()
      : { ...scope(), plans: [structuredPlanRecord()], next_cursor: "", has_more: false });
  };
  try {
    const listed = await listApplicationEvaluationPlans(config);
    const exact = await readApplicationEvaluationPlanVersion(config, "plan_structured", 1);
    assert.equal(listed.plans[0]?.schemaVersion, "application_evaluation_plan.v2");
    assert.equal(exact.version?.target.workflowDefinition?.inputContract?.contractId, "contract_structured");
    assert.deepEqual(exact.version?.items[0]?.workflowDefinition?.inputs, { customer_name: "Ada", retry_count: 2 });

    const draft = applicationEvaluationPlanTemplate("workflow_definition_executor_v2");
    draft.target = exact.version?.target ?? draft.target;
    draft.items = exact.version?.items ?? draft.items;
    await createApplicationEvaluationPlan(config, draft);
    const body = bodies[0] as { target: { workflow_definition: Record<string, unknown> }; items: Array<{ workflow_definition: Record<string, unknown> }> };
    assert.deepEqual(Object.keys(body.items[0]?.workflow_definition ?? {}), ["inputs"]);
    assert.equal(Object.hasOwn(body.target.workflow_definition, "input_contract"), true);
    assert.equal(JSON.stringify(body).includes("input_text"), false);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

function scope() {
  return {
    request_id: "request_evaluation",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    environment: "test",
    application_id: "app_flow_copilot",
    failure_code: null,
    failure_summary: "",
    audit_ref: "audit_evaluation",
  };
}

function planRecord() {
  return {
    schema_version: "application_evaluation_plan.v1",
    plan_id: "plan_demo",
    record_version: 3,
    latest_plan_version: 1,
    latest_plan_digest: digest,
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    environment: "test",
    application_id: "app_flow_copilot",
    name: "Workflow regression",
    execution_profile: "workflow_definition_executor_v1",
    item_count: 1,
    lifecycle_state: "active",
    created_at: "2026-08-10T00:00:00Z",
    updated_at: "2026-08-10T00:01:00Z",
    created_by_actor_ref: "subject_demo_user",
    updated_by_actor_ref: "subject_demo_user",
    request_id: "request_plan",
    audit_ref: "audit_plan",
  };
}

function planVersion() {
  return {
    schema_version: "application_evaluation_plan_version.v1",
    plan_id: "plan_demo",
    plan_version: 1,
    previous_plan_version: 0,
    plan_digest: digest,
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    environment: "test",
    application_id: "app_flow_copilot",
    name: "Workflow regression",
    execution_profile: "workflow_definition_executor_v1",
    target: { workflow_definition: {
      definition_id: "definition_demo",
      expected_pointer_version: 1,
      expected_definition_version: 1,
      expected_definition_digest: digest,
    } },
    items: [{
      item_key: "case_01",
      name: "Representative case",
      expected_classification: "unchanged",
      workflow_definition: { input_text: "Representative input", condition_values: {}, model: "", temperature: null },
      application_rag: null,
      prompt_application: null,
      agent_copilot: null,
    }],
    created_at: "2026-08-10T00:00:00Z",
    created_by_actor_ref: "subject_demo_user",
    request_id: "request_plan_version",
    audit_ref: "audit_plan_version",
  };
}

function structuredContractRecord() {
  return {
    contract_id: "contract_structured",
    fields: [
      { name: "customer_name", value_type: "string", required: true, label: "Customer", description: "Evaluation customer." },
      { name: "retry_count", value_type: "integer", required: false, label: "Retries", description: "Evaluation retry count." },
    ],
    summary: "Typed evaluation inputs.",
    contract_digest: digest,
  };
}

function structuredPlanRecord() {
  return {
    ...planRecord(),
    schema_version: "application_evaluation_plan.v2",
    plan_id: "plan_structured",
    execution_profile: "workflow_definition_executor_v2",
    name: "Structured workflow regression",
  };
}

function structuredPlanVersion() {
  return {
    ...planVersion(),
    schema_version: "application_evaluation_plan_version.v2",
    plan_id: "plan_structured",
    execution_profile: "workflow_definition_executor_v2",
    name: "Structured workflow regression",
    target: { workflow_definition: {
      definition_id: "definition_structured",
      expected_pointer_version: 2,
      expected_definition_version: 3,
      expected_definition_digest: digest,
      input_contract: structuredContractRecord(),
    } },
    items: [{
      item_key: "case_01",
      name: "Representative structured case",
      expected_classification: "unchanged",
      workflow_definition: { inputs: { customer_name: "Ada", retry_count: 2 } },
      application_rag: null,
      prompt_application: null,
      agent_copilot: null,
    }],
  };
}

function structuredPlanEnvelope() {
  return { ...scope(), plan: structuredPlanRecord(), version: structuredPlanVersion(), current_record_version: 3, current_state: "active" };
}

function campaignRecord(campaignId = "campaign_candidate", recordVersion = 7) {
  const runId = campaignId === "campaign_baseline" ? "run_baseline" : "run_candidate";
  return {
    schema_version: "application_evaluation_campaign.v1",
    campaign_id: campaignId,
    client_campaign_key: `${campaignId}_key`,
    record_version: recordVersion,
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    environment: "test",
    application_id: "app_flow_copilot",
    plan_id: "plan_demo",
    plan_version: 1,
    plan_digest: digest,
    execution_profile: "workflow_definition_executor_v1",
    quota_api_key_id: "key_actor_owned",
    authority: { execution_profile: "workflow_definition_executor_v1", authority_digest: digest, snapshot: { definition_id: "definition_demo" } },
    state: "succeeded",
    current_item_index: 1,
    succeeded_items: 1,
    failed_items: 0,
    failure_code: "",
    failure_summary: "",
    items: [{
      item_key: "case_01",
      run_id: runId,
      state: "succeeded",
      run_schema_version: "workflow_run_record.v5",
      run_profile: "workflow_definition_executor_v1",
      authority_digest: digest,
      failure_code: "",
      failure_boundary: "",
      started_at: "2026-08-10T00:02:00Z",
      completed_at: "2026-08-10T00:03:00Z",
    }],
    handoff: null,
    created_at: "2026-08-10T00:02:00Z",
    started_at: "2026-08-10T00:02:00Z",
    completed_at: "2026-08-10T00:03:00Z",
    created_by_actor_ref: "subject_demo_user",
    updated_by_actor_ref: "subject_demo_user",
    request_id: "request_campaign",
    audit_ref: "audit_campaign",
  };
}

function handoffRecord() {
  return {
    baseline_campaign_id: "campaign_baseline",
    candidate_campaign_id: "campaign_candidate",
    case_refs: [{ case_id: "case_evaluation_01", version: 2 }],
    suite_id: "suite_evaluation_01",
    state: "complete",
    audit_ref: "audit_handoff",
  };
}

function planListEnvelope() {
  return { ...scope(), plans: [planRecord()], next_cursor: "", has_more: false };
}

function planEnvelope() {
  return { ...scope(), plan: planRecord(), version: planVersion(), current_record_version: 3, current_state: "active" };
}

function campaignEnvelope() {
  return { ...scope(), campaign: campaignRecord(), idempotent_replay: false, current_record_version: 7, current_state: "succeeded" };
}

function pairEnvelope() {
  return {
    ...scope(),
    review: {
      plan_id: "plan_demo",
      plan_name: "Workflow regression",
      plan_version: 1,
      plan_digest: digest,
      execution_profile: "workflow_definition_executor_v1",
      baseline_campaign_id: "campaign_baseline",
      candidate_campaign_id: "campaign_candidate",
      expected_matches: 1,
      expected_mismatches: 0,
      items: [{
        item_key: "case_01",
        name: "Representative case",
        baseline_run_id: "run_baseline",
        candidate_run_id: "run_candidate",
        expected_classification: "unchanged",
        actual_classification: "unchanged",
        expectation_matched: true,
        comparison: null as Record<string, unknown> | null,
      }],
      existing_handoff: null,
    },
    candidate_campaign: campaignRecord(),
    handoff: handoffRecord(),
    idempotent_replay: false,
    current_baseline_record_version: 5,
    current_candidate_record_version: 7,
  };
}

function mappedPlan(): ApplicationEvaluationPlan {
  return {
    schemaVersion: "application_evaluation_plan.v1",
    planId: "plan_demo",
    recordVersion: 3,
    latestPlanVersion: 1,
    latestPlanDigest: digest,
    tenantRef: "tenant_demo",
    workspaceId: "workspace_demo",
    environment: "test",
    applicationId: "app_flow_copilot",
    name: "Workflow regression",
    executionProfile: "workflow_definition_executor_v1",
    itemCount: 1,
    lifecycleState: "active",
    createdAt: "2026-08-10T00:00:00Z",
    updatedAt: "2026-08-10T00:01:00Z",
    createdByActorRef: "subject_demo_user",
    updatedByActorRef: "subject_demo_user",
    requestId: "request_plan",
    auditRef: "audit_plan",
  };
}

function mappedVersion(): ApplicationEvaluationPlanVersion {
  const template = applicationEvaluationPlanTemplate("workflow_definition_executor_v1");
  return {
    schemaVersion: "application_evaluation_plan_version.v1",
    planId: "plan_demo",
    planVersion: 1,
    previousPlanVersion: 0,
    planDigest: digest,
    tenantRef: "tenant_demo",
    workspaceId: "workspace_demo",
    environment: "test",
    applicationId: "app_flow_copilot",
    name: template.name,
    executionProfile: template.executionProfile,
    target: template.target,
    items: template.items,
    createdAt: "2026-08-10T00:00:00Z",
    createdByActorRef: "subject_demo_user",
    requestId: "request_plan_version",
    auditRef: "audit_plan_version",
  };
}

function mappedCampaign(campaignId: string, recordVersion: number): ApplicationEvaluationCampaign {
  const source = campaignRecord(campaignId, recordVersion);
  return {
    schemaVersion: "application_evaluation_campaign.v1",
    campaignId,
    clientCampaignKey: source.client_campaign_key,
    recordVersion,
    tenantRef: "tenant_demo",
    workspaceId: "workspace_demo",
    environment: "test",
    applicationId: "app_flow_copilot",
    planId: "plan_demo",
    planVersion: 1,
    planDigest: digest,
    executionProfile: "workflow_definition_executor_v1",
    quotaAPIKeyId: "key_actor_owned",
    authority: { executionProfile: "workflow_definition_executor_v1", authorityDigest: digest },
    state: "succeeded",
    currentItemIndex: 1,
    succeededItems: 1,
    failedItems: 0,
    failureCode: "",
    failureSummary: "",
    items: [{
      itemKey: "case_01",
      runId: campaignId === "campaign_baseline" ? "run_baseline" : "run_candidate",
      state: "succeeded",
      runSchemaVersion: "workflow_run_record.v5",
      runProfile: "workflow_definition_executor_v1",
      authorityDigest: digest,
      failureCode: "",
      failureBoundary: "",
      startedAt: "2026-08-10T00:02:00Z",
      completedAt: "2026-08-10T00:03:00Z",
    }],
    handoff: null,
    createdAt: "2026-08-10T00:02:00Z",
    startedAt: "2026-08-10T00:02:00Z",
    completedAt: "2026-08-10T00:03:00Z",
    createdByActorRef: "subject_demo_user",
    updatedByActorRef: "subject_demo_user",
    requestId: "request_campaign",
    auditRef: "audit_campaign",
  };
}

function jsonResponse(value: unknown, status = 200) {
  return new Response(JSON.stringify(value), { status, headers: { "Content-Type": "application/json" } });
}
