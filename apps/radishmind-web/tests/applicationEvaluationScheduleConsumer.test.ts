import assert from "node:assert/strict";
import test from "node:test";

import {
  applicationEvaluationPlanTemplate,
  type ApplicationEvaluationConfig,
  type ApplicationEvaluationPlan,
  type ApplicationEvaluationPlanVersion,
} from "../src/features/control-plane-read/applicationEvaluationCampaignConsumer.ts";
import {
  createApplicationEvaluationSchedule,
  listApplicationEvaluationSchedules,
  readApplicationEvaluationScheduleOccurrence,
  readApplicationEvaluationScheduleVersion,
  reviseApplicationEvaluationSchedule,
  transitionApplicationEvaluationSchedule,
  type ApplicationEvaluationSchedule,
} from "../src/features/control-plane-read/applicationEvaluationScheduleConsumer.ts";

const digest = `sha256:${"a".repeat(64)}`;
const scheduleId = "aesch_aaaaaaaaaaaaaaaa";
const planId = "aeplan_aaaaaaaaaaaaaaaa";
const apiKeyId = "key_aaaaaaaaaaaaaaaa";
const campaignId = "aecamp_aaaaaaaaaaaaaaaa";
const config: ApplicationEvaluationConfig = {
  mode: "dev_application_evaluation_http",
  baseUrl: "http://platform.test",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  environment: "test",
  applicationId: "app_flow_copilot",
  subjectRef: "subject_demo_user",
};

test("Application Evaluation Schedule list remains offline without a request", async () => {
  const originalFetch = globalThis.fetch;
  let calls = 0;
  globalThis.fetch = async () => { calls += 1; throw new Error("unexpected fetch"); };
  try {
    const result = await listApplicationEvaluationSchedules({ ...config, mode: "offline" });
    assert.equal(result.failureCode, "application_evaluation_write_disabled");
    assert.deepEqual(result.schedules, []);
    assert.equal(calls, 0);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Application Evaluation Schedule reads strict list and exact version with read scope", async () => {
  const originalFetch = globalThis.fetch;
  const requests: Array<{ url: URL; headers: Headers }> = [];
  globalThis.fetch = async (input, init) => {
    const url = new URL(String(input));
    requests.push({ url, headers: new Headers(init?.headers) });
    return jsonResponse(url.pathname.endsWith("/versions/3") ? scheduleEnvelope({ version: versionRecord() }) : scheduleListEnvelope());
  };
  try {
    const listed = await listApplicationEvaluationSchedules(config);
    const exact = await readApplicationEvaluationScheduleVersion(config, scheduleId, 3);
    assert.equal(listed.schedules[0]?.nextDueAt, "2026-09-01T02:30:00Z");
    assert.equal(exact.version?.authorization.revalidationPolicy, "every_occurrence");
    assert.equal(exact.version?.maxProviderAttempts, 1);
    assert.equal(requests[0]?.url.searchParams.get("limit"), "100");
    assert.equal(requests[1]?.url.pathname, `/v1/user-workspace/applications/app_flow_copilot/evaluation-schedules/${scheduleId}/versions/3`);
    for (const request of requests) {
      assert.equal(request.headers.get("X-RadishMind-Dev-Read-Scopes"), "application_evaluations:read");
      assert.equal(request.headers.get("X-RadishMind-Dev-Workflow-Workspace"), "workspace_demo");
      assert.equal(request.headers.get("X-RadishMind-Dev-Workflow-Application"), "app_flow_copilot");
    }
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Application Evaluation Schedule local Session transport is credentialed and never mixes dev auth headers", async () => {
  const originalFetch = globalThis.fetch;
  let captured: RequestInit | undefined;
  globalThis.fetch = async (_input, init) => {
    captured = init;
    return jsonResponse(scheduleListEnvelope());
  };
  try {
    await listApplicationEvaluationSchedules({ ...config, authMode: "local_session_dev_test" });
    const headers = new Headers(captured?.headers);
    assert.equal(captured?.credentials, "include");
    assert.equal(captured?.cache, "no-store");
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

test("Application Evaluation Schedule create and revision bind exact Prompt plan, quota, UTC cadence, and CAS", async () => {
  const originalFetch = globalThis.fetch;
  const requests: Array<{ url: URL; headers: Headers; body: Record<string, unknown> }> = [];
  globalThis.fetch = async (input, init) => {
    requests.push({
      url: new URL(String(input)),
      headers: new Headers(init?.headers),
      body: JSON.parse(String(init?.body)) as Record<string, unknown>,
    });
    return jsonResponse(scheduleEnvelope({ schedule: scheduleRecord(), version: versionRecord() }));
  };
  try {
    const plan = mappedPlan();
    const version = mappedPlanVersion();
    await createApplicationEvaluationSchedule(config, { plan, version, quotaAPIKeyId: apiKeyId, hour: 2, minute: 30 });
    await reviseApplicationEvaluationSchedule(config, mappedSchedule(), { plan, version, quotaAPIKeyId: apiKeyId, hour: 4, minute: 5 });
    assert.deepEqual(requests[0]?.body, {
      workspace_id: "workspace_demo",
      environment: "test",
      plan_id: planId,
      plan_version: 17,
      plan_digest: digest,
      expected_plan_record_version: 4,
      quota_api_key_id: apiKeyId,
      schedule: { rule: "daily_utc", hour: 2, minute: 30 },
      acknowledge_provider_consumption: true,
    });
    assert.equal(requests[1]?.body.expected_version, 8);
    assert.deepEqual(requests[1]?.body.schedule, { rule: "daily_utc", hour: 4, minute: 5 });
    assert.equal(requests[1]?.url.pathname.endsWith(`/${scheduleId}/revisions`), true);
    assert.equal(requests[0]?.headers.get("X-RadishMind-Dev-Read-Scopes"), "application_evaluations:execute,workflow_runs:execute");
    assert.equal(JSON.stringify(requests[0]?.body).includes("token"), false);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Application Evaluation Schedule lifecycle uses CAS and action-specific confirmations", async () => {
  const originalFetch = globalThis.fetch;
  const requests: Array<{ url: URL; body: Record<string, unknown> }> = [];
  globalThis.fetch = async (input, init) => {
    requests.push({ url: new URL(String(input)), body: JSON.parse(String(init?.body)) as Record<string, unknown> });
    return jsonResponse(scheduleEnvelope({ schedule: scheduleRecord() }));
  };
  try {
    const schedule = mappedSchedule();
    await transitionApplicationEvaluationSchedule(config, schedule, "activate");
    await transitionApplicationEvaluationSchedule(config, schedule, "pause");
    await transitionApplicationEvaluationSchedule(config, schedule, "resume");
    await transitionApplicationEvaluationSchedule(config, schedule, "archive");
    assert.deepEqual(requests.map((request) => request.url.pathname.split("/").at(-1)), ["activate", "pause", "resume", "archive"]);
    assert.deepEqual(requests.map((request) => ({
      expected: request.body.expected_version,
      provider: request.body.acknowledge_provider_consumption,
      future: request.body.acknowledge_no_future_occurrences,
    })), [
      { expected: 8, provider: true, future: false },
      { expected: 8, provider: false, future: false },
      { expected: 8, provider: true, future: false },
      { expected: 8, provider: false, future: true },
    ]);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Application Evaluation Schedule reads one exact occurrence and exposes only its existing Campaign handoff", async () => {
  const originalFetch = globalThis.fetch;
  let requestURL: URL | null = null;
  globalThis.fetch = async (input) => {
    requestURL = new URL(String(input));
    return jsonResponse(scheduleEnvelope({ occurrence: occurrenceRecord() }));
  };
  try {
    const result = await readApplicationEvaluationScheduleOccurrence(config, scheduleId, 3, "2026-08-31T02:30:00Z");
    assert.equal(result.occurrence?.campaignId, campaignId);
    assert.equal(result.occurrence?.clientCampaignKey, "scheduled_campaign_aaaaaaaaaaaaaaaaaaaaaaaa");
    assert.equal(decodeURIComponent(requestURL?.pathname ?? "").endsWith(`/${scheduleId}/occurrences/3/2026-08-31T02:30:00Z`), true);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Application Evaluation Schedule fails closed on extra, bearer-shaped, scope-drifted, and invalid occurrence responses", async () => {
  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async () => jsonResponse({ ...scheduleListEnvelope(), debug: true });
    await assert.rejects(() => listApplicationEvaluationSchedules(config), /contract mismatch/);

    globalThis.fetch = async () => jsonResponse({ ...scheduleListEnvelope(), access_token: "never" });
    await assert.rejects(() => listApplicationEvaluationSchedules(config), /forbidden field/);

    globalThis.fetch = async () => jsonResponse({ ...scheduleListEnvelope(), workspace_id: "workspace_other" });
    await assert.rejects(() => listApplicationEvaluationSchedules(config), /scope or failure contract mismatch/);

    const invalidOccurrence = occurrenceRecord();
    invalidOccurrence.state = "succeeded";
    invalidOccurrence.campaign_id = null;
    globalThis.fetch = async () => jsonResponse(scheduleEnvelope({ occurrence: invalidOccurrence }));
    await assert.rejects(
      () => readApplicationEvaluationScheduleOccurrence(config, scheduleId, 3, "2026-08-31T02:30:00Z"),
      /schedule envelope contract mismatch/,
    );
  } finally {
    globalThis.fetch = originalFetch;
  }
});

function scope() {
  return {
    request_id: "request_schedule",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    environment: "test",
    application_id: "app_flow_copilot",
    failure_code: null,
    failure_summary: "",
    audit_ref: "audit_schedule",
  };
}

function scheduleRecord() {
  return {
    schema_version: "application_evaluation_schedule.v1",
    schedule_id: scheduleId,
    record_version: 8,
    latest_schedule_version: 3,
    latest_schedule_digest: digest,
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    environment: "test",
    application_id: "app_flow_copilot",
    plan_id: planId,
    plan_version: 17,
    plan_digest: digest,
    execution_profile: "prompt_application_invocation_v1",
    quota_api_key_id: apiKeyId,
    authorization_model: "system_actor_schedule_scoped_delegation_v1",
    system_actor_ref: "system:application-evaluation-scheduler",
    delegated_by_user_ref: "subject_demo_user",
    lifecycle_state: "active",
    next_due_at: "2026-09-01T02:30:00Z",
    created_at: "2026-08-20T02:30:00Z",
    updated_at: "2026-08-31T02:31:00Z",
    created_by_actor_ref: "subject_demo_user",
    updated_by_actor_ref: "subject_demo_user",
    request_id: "request_schedule_record",
    audit_ref: "audit_schedule_record",
  };
}

function versionRecord() {
  return {
    schema_version: "application_evaluation_schedule_version.v1",
    schedule_id: scheduleId,
    schedule_version: 3,
    previous_schedule_version: 2,
    schedule_digest: digest,
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    environment: "test",
    application_id: "app_flow_copilot",
    plan_id: planId,
    plan_version: 17,
    plan_digest: digest,
    execution_profile: "prompt_application_invocation_v1",
    quota_api_key_id: apiKeyId,
    schedule: { rule: "daily_utc", hour: 2, minute: 30 },
    item_count: 1,
    max_provider_attempts: 1,
    missed_window_policy: "record_only_no_catch_up",
    overlap_policy: "skip_while_campaign_non_terminal",
    authorization: {
      model: "system_actor_schedule_scoped_delegation_v1",
      system_actor_ref: "system:application-evaluation-scheduler",
      delegated_by_user_ref: "subject_demo_user",
      required_permissions: ["application_evaluations:execute", "workflow_runs:execute"],
      revalidation_policy: "every_occurrence",
      api_key_ownership_policy: "delegated_user_current_owner",
      revocation_policy: "fail_closed_immediate",
    },
    created_at: "2026-08-31T02:20:00Z",
    created_by_actor_ref: "subject_demo_user",
    request_id: "request_schedule_version",
    audit_ref: "audit_schedule_version",
  };
}

function occurrenceRecord(): Record<string, unknown> {
  return {
    schema_version: "application_evaluation_schedule_occurrence.v1",
    record_version: 4,
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    environment: "test",
    application_id: "app_flow_copilot",
    schedule_id: scheduleId,
    schedule_version: 3,
    schedule_digest: digest,
    scheduled_for_utc: "2026-08-31T02:30:00Z",
    state: "succeeded",
    client_campaign_key: "scheduled_campaign_aaaaaaaaaaaaaaaaaaaaaaaa",
    campaign_id: campaignId,
    system_actor_ref: "system:application-evaluation-scheduler",
    delegated_by_user_ref: "subject_demo_user",
    claimed_at: "2026-08-31T02:30:01Z",
    failure_code: null,
    created_at: "2026-08-31T02:30:00Z",
    updated_at: "2026-08-31T02:31:00Z",
    completed_at: "2026-08-31T02:31:00Z",
    request_id: "request_schedule_occurrence",
    audit_ref: "audit_schedule_occurrence",
  };
}

function scheduleListEnvelope() {
  return { ...scope(), schedules: [scheduleRecord()], next_cursor: "", has_more: false };
}

function scheduleEnvelope(parts: { schedule?: unknown; version?: unknown; occurrence?: unknown } = {}) {
  return {
    ...scope(),
    schedule: parts.schedule ?? null,
    version: parts.version ?? null,
    occurrence: parts.occurrence ?? null,
    current_record_version: parts.occurrence ? 4 : parts.schedule ? 8 : 0,
    current_state: parts.occurrence ? "succeeded" : parts.schedule ? "active" : "",
  };
}

function mappedSchedule(): ApplicationEvaluationSchedule {
  return {
    schemaVersion: "application_evaluation_schedule.v1", scheduleId, recordVersion: 8, latestScheduleVersion: 3,
    latestScheduleDigest: digest, tenantRef: "tenant_demo", workspaceId: "workspace_demo", environment: "test",
    applicationId: "app_flow_copilot", planId, planVersion: 17, planDigest: digest,
    executionProfile: "prompt_application_invocation_v1", quotaAPIKeyId: apiKeyId,
    authorizationModel: "system_actor_schedule_scoped_delegation_v1",
    systemActorRef: "system:application-evaluation-scheduler", delegatedByUserRef: "subject_demo_user",
    lifecycleState: "active", nextDueAt: "2026-09-01T02:30:00Z", createdAt: "2026-08-20T02:30:00Z",
    updatedAt: "2026-08-31T02:31:00Z", createdByActorRef: "subject_demo_user", updatedByActorRef: "subject_demo_user",
    requestId: "request_schedule_record", auditRef: "audit_schedule_record",
  };
}

function mappedPlan(): ApplicationEvaluationPlan {
  return {
    schemaVersion: "application_evaluation_plan.v1", planId, recordVersion: 4, latestPlanVersion: 17,
    latestPlanDigest: digest, tenantRef: "tenant_demo", workspaceId: "workspace_demo", environment: "test",
    applicationId: "app_flow_copilot", name: "Prompt regression", executionProfile: "prompt_application_invocation_v1",
    itemCount: 1, lifecycleState: "active", createdAt: "2026-08-01T00:00:00Z", updatedAt: "2026-08-31T00:00:00Z",
    createdByActorRef: "subject_demo_user", updatedByActorRef: "subject_demo_user",
    requestId: "request_plan", auditRef: "audit_plan",
  };
}

function mappedPlanVersion(): ApplicationEvaluationPlanVersion {
  const template = applicationEvaluationPlanTemplate("prompt_application_invocation_v1");
  return {
    schemaVersion: "application_evaluation_plan_version.v1", planId, planVersion: 17, previousPlanVersion: 16,
    planDigest: digest, tenantRef: "tenant_demo", workspaceId: "workspace_demo", environment: "test",
    applicationId: "app_flow_copilot", name: template.name, executionProfile: template.executionProfile,
    target: template.target, items: template.items, createdAt: "2026-08-31T00:00:00Z",
    createdByActorRef: "subject_demo_user", requestId: "request_plan_version", auditRef: "audit_plan_version",
  };
}

function jsonResponse(value: unknown): Response {
  return new Response(JSON.stringify(value), { status: 200, headers: { "Content-Type": "application/json" } });
}
