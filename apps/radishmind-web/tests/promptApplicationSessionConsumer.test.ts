import assert from "node:assert/strict";
import test from "node:test";

import {
  createPromptApplicationSession,
  executePromptApplicationSessionTurn,
  listPromptApplicationSessions,
  type PromptApplicationSessionConfig,
} from "../src/features/control-plane-read/promptApplicationSessionConsumer.ts";

const config: PromptApplicationSessionConfig = {
  mode: "dev_application_session_http",
  baseUrl: "http://platform.test",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  subjectRef: "subject_demo_user",
};
const applicationId = "app_aaaaaaaaaaaaaaaa";

test("Prompt Session v2 creates explicit profile and sends transient variables only on turn", async () => {
  const requests: Array<{ url: string; headers: Headers; body: any }> = [];
  globalThis.fetch = async (input, init) => {
    requests.push({
      url: String(input),
      headers: new Headers(init?.headers),
      body: JSON.parse(String(init?.body)),
    });
    return requests.length === 1 ? jsonResponse(sessionEnvelope()) : jsonResponse(turnEnvelope());
  };

  const created = await createPromptApplicationSession(config, applicationId);
  assert.equal(created.status, "ready");
  assert.equal(created.session?.templateVersion, 2);
  assert.deepEqual(requests[0]?.body, {
    workspace_id: "workspace_demo",
    application_id: applicationId,
    execution_profile: "prompt_application_invocation_v1",
  });
  assert.equal(requests[0]?.headers.get("X-RadishMind-Dev-Read-Scopes"), "application_sessions:write");

  const turn = await executePromptApplicationSessionTurn(
    config,
    created.session!,
    { question: "如何审查？", tone: "清晰" },
    "prompt-turn-client-0001",
  );
  assert.equal(turn.status, "succeeded");
  assert.equal(turn.output, "请检查模板、草案与候选。");
  assert.equal(turn.turn?.runId, "run_aaaaaaaaaaaaaaaa");
  assert.deepEqual(requests[1]?.body, {
    workspace_id: "workspace_demo",
    application_id: applicationId,
    expected_session_version: 1,
    client_turn_key: "prompt-turn-client-0001",
    input_text: "",
    condition_values: {},
    model: "",
    variables: { question: "如何审查？", tone: "清晰" },
  });
  assert.equal(requests[1]?.headers.get("X-RadishMind-Dev-Read-Scopes"), "application_sessions:execute");
});

test("Prompt Session v2 lists only strict metadata records for stage recovery", async () => {
  let requestUrl = "";
  let requestHeaders = new Headers();
  globalThis.fetch = async (input, init) => {
    requestUrl = String(input);
    requestHeaders = new Headers(init?.headers);
    return jsonResponse({
      request_id: "prompt-session-list-request",
      tenant_ref: "tenant_demo",
      workspace_id: "workspace_demo",
      application_id: applicationId,
      items: [{ ...sessionDocument(), record_version: 2, turn_count: 1, last_turn_id: "appturn_aaaaaaaaaaaaaaaa" }],
      next_cursor: null,
      failure_code: null,
      audit_ref: "audit-prompt-session-list",
    });
  };
  const listed = await listPromptApplicationSessions(config, applicationId);
  assert.equal(listed.status, "ready");
  assert.equal(listed.sessions[0]?.recordVersion, 2);
  assert.equal(listed.sessions[0]?.turnCount, 1);
  const url = new URL(requestUrl);
  assert.equal(url.searchParams.get("execution_profile"), "prompt_application_invocation_v1");
  assert.equal(url.searchParams.get("state"), "active");
  assert.equal(requestHeaders.get("X-RadishMind-Dev-Read-Scopes"), "application_sessions:read");

  globalThis.fetch = async () => jsonResponse({
    request_id: "prompt-session-list-failure",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: applicationId,
    items: [],
    next_cursor: null,
    failure_code: "application_session_store_unavailable",
    audit_ref: "audit-prompt-session-list-failure",
  }, 503);
  assert.equal(
    (await listPromptApplicationSessions(config, applicationId)).failureCode,
    "application_session_store_unavailable",
  );
});

test("Prompt Session v2 list rejects transcript material", async () => {
  globalThis.fetch = async () => jsonResponse({
    request_id: "prompt-session-list-request",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: applicationId,
    items: [{ ...sessionDocument(), variables: { question: "不应持久化" } }],
    next_cursor: null,
    failure_code: null,
    audit_ref: "audit-prompt-session-list",
  });
  assert.equal(
    (await listPromptApplicationSessions(config, applicationId)).failureCode,
    "application_session_response_invalid",
  );
});

test("Prompt Session v2 list rejects unknown envelope fields", async () => {
  globalThis.fetch = async () => jsonResponse({
    request_id: "prompt-session-list-request",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: applicationId,
    items: [sessionDocument()],
    next_cursor: null,
    failure_code: null,
    audit_ref: "audit-prompt-session-list",
    unexpected_projection: "must fail closed",
  });
  assert.equal(
    (await listPromptApplicationSessions(config, applicationId)).failureCode,
    "application_session_response_invalid",
  );
});

test("Prompt Session v2 rejects v1 schema and replay output", async () => {
  const invalid = sessionEnvelope();
  invalid.session.schema_version = "application_session.v1";
  globalThis.fetch = async () => jsonResponse(invalid);
  assert.equal((await createPromptApplicationSession(config, applicationId)).failureCode, "application_session_response_invalid");

  const replay = turnEnvelope();
  replay.idempotent_replay = true;
  globalThis.fetch = async () => jsonResponse(replay);
  const session = mapFixtureSession();
  assert.equal(
    (await executePromptApplicationSessionTurn(config, session, { question: "审查" }, "prompt-turn-client-0002")).failureCode,
    "application_session_response_invalid",
  );
});

test("Prompt Session v2 rejects raw variable material in response metadata", async () => {
  const invalid = turnEnvelope();
  invalid.turn.variables = { question: "不应回流" };
  globalThis.fetch = async () => jsonResponse(invalid);
  assert.equal(
    (await executePromptApplicationSessionTurn(
      config,
      mapFixtureSession(),
      { question: "审查" },
      "prompt-turn-client-0003",
    )).failureCode,
    "application_session_response_invalid",
  );
});

function sessionEnvelope() {
  return {
    request_id: "prompt-session-request",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: applicationId,
    session: sessionDocument(),
    failure_code: null,
    current_record_version: 1,
    current_state: "active",
    idempotent_replay: false,
    audit_ref: "audit-prompt-session-request",
  };
}

function turnEnvelope() {
  const session = { ...sessionDocument(), record_version: 2, turn_count: 1, last_turn_id: "appturn_aaaaaaaaaaaaaaaa" };
  return {
    request_id: "prompt-turn-request",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: applicationId,
    session_id: session.session_id,
    session,
    turn: {
      schema_version: "application_session_turn.v2",
      turn_id: "appturn_aaaaaaaaaaaaaaaa",
      session_id: session.session_id,
      sequence: 1,
      client_turn_key: "prompt-turn-client-0001",
      tenant_ref: "tenant_demo",
      workspace_id: "workspace_demo",
      application_id: applicationId,
      owner_subject_ref: "subject_demo_user",
      execution_profile: "prompt_application_invocation_v1",
      authority: authority(),
      status: "succeeded",
      input_digest: `sha256:${"4".repeat(64)}`,
      input_bytes: 42,
      run_ref: { run_id: "run_aaaaaaaaaaaaaaaa", schema_version: "workflow_run_record.v6" },
      failure_code: "",
      failure_summary: "",
      started_at: "2026-07-25T10:00:00Z",
      completed_at: "2026-07-25T10:00:01Z",
      actor_ref: "subject_demo_user",
      request_id: "prompt-turn-request",
      audit_ref: "audit-prompt-turn-request",
    },
    prompt_output: "请检查模板、草案与候选。",
    failure_code: null,
    failure_summary: "",
    idempotent_replay: false,
    audit_ref: "audit-prompt-turn-request",
  };
}

function sessionDocument() {
  return {
    schema_version: "application_session.v2",
    session_id: "appsess_aaaaaaaaaaaaaaaa",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: applicationId,
    owner_subject_ref: "subject_demo_user",
    state: "active",
    record_version: 1,
    profile_binding: { execution_profile: "prompt_application_invocation_v1" },
    authority: authority(),
    content_retention: "metadata_only",
    turn_count: 0,
    last_turn_id: null,
    created_at: "2026-07-25T10:00:00Z",
    updated_at: "2026-07-25T10:00:00Z",
    closed_at: null,
    created_by_actor_ref: "subject_demo_user",
    updated_by_actor_ref: "subject_demo_user",
    request_id: "prompt-session-request",
    audit_ref: "audit-prompt-session-request",
  };
}

function authority() {
  const digest = (character: string) => `sha256:${character.repeat(64)}`;
  return {
    schema_version: "application_runtime_authority.v2",
    execution_profile: "prompt_application_invocation_v1",
    application_id: applicationId,
    application_record_version: 1,
    application_lifecycle: "active",
    prompt_application: {
      assignment_id: "ptra_aaaaaaaaaaaaaaaa",
      assignment_version: 1,
      assignment_digest: digest("a"),
      publish_candidate_id: "candidate-prompt-v1",
      publish_review_version: 1,
      draft_id: "draft-prompt-v1",
      draft_version: 2,
      draft_digest: digest("b"),
      prompt_template_ref: { template_id: "ptpl_aaaaaaaaaaaaaaaa", template_version: 2, template_digest: digest("c") },
      default_protocol: "responses",
      default_model: "radishmind-local-dev",
      protocol_policy_digest: digest("d"),
      model_eligibility_digest: digest("e"),
    },
    authority_digest: digest("f"),
  };
}

function mapFixtureSession() {
  return {
    sessionId: "appsess_aaaaaaaaaaaaaaaa",
    applicationId,
    state: "active" as const,
    recordVersion: 1,
    assignmentId: "ptra_aaaaaaaaaaaaaaaa",
    assignmentVersion: 1,
    templateId: "ptpl_aaaaaaaaaaaaaaaa",
    templateVersion: 2,
    turnCount: 0,
    lastTurnId: null,
    createdAt: "2026-07-25T10:00:00Z",
    updatedAt: "2026-07-25T10:00:00Z",
  };
}

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });
}
