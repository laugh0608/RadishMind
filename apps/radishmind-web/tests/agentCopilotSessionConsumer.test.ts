import assert from "node:assert/strict";
import test from "node:test";

import {
  createAgentCopilotSession,
  executeAgentCopilotSessionTurn,
  type AgentCopilotSession,
  type AgentCopilotSessionConfig,
} from "../src/features/control-plane-read/agentCopilotSessionConsumer.ts";

const config: AgentCopilotSessionConfig = {
  mode: "dev_application_session_http",
  baseUrl: "http://platform.test",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  subjectRef: "subject_demo_user",
};
const applicationId = "app_aaaaaaaaaaaaaaaa";

test("Agent Session v3 create binds the exact execution profile", async () => {
  let body: any;
  globalThis.fetch = async (_url, init) => {
    body = JSON.parse(String(init?.body));
    return jsonResponse(createEnvelope());
  };
  const result = await createAgentCopilotSession(config, applicationId);
  assert.equal(result.status, "ready");
  assert.equal(result.session?.profileId, "acpf_aaaaaaaaaaaaaaaa");
  assert.deepEqual(body, {
    workspace_id: "workspace_demo",
    application_id: applicationId,
    execution_profile: "agent_copilot_suggestion_v1",
  });
});

test("Agent turn keeps context request-only and maps one transient advisory response", async () => {
  let capturedBody = "";
  globalThis.fetch = async (_url, init) => {
    capturedBody = String(init?.body);
    return jsonResponse(turnEnvelope());
  };
  const result = await executeAgentCopilotSessionTurn(config, session(), {
    task: "suggest_flowsheet_edits",
    locale: "zh-CN",
    conversationId: "",
    artifacts: [],
    context: { selected_unit_ids: ["unit-1"], diagnostics: [{ code: "not_converged" }] },
    clientTurnKey: "agent-turn-test-001",
  });
  assert.equal(result.status, "succeeded");
  assert.equal(result.turn?.runId, "run_aaaaaaaaaaaaaaaa");
  assert.equal(result.response?.proposedActions[0]?.requiresConfirmation, true);
  assert.match(capturedBody, /selected_unit_ids/u);
  assert.doesNotMatch(JSON.stringify(result), /selected_unit_ids|not_converged/u);
});

test("Agent turn drops replay output and rejects safety relaxation or sensitive response material", async () => {
  globalThis.fetch = async () => jsonResponse({ ...turnEnvelope(), idempotent_replay: true, agent_response: null });
  const replay = await executeAgentCopilotSessionTurn(config, session(), turnInput());
  assert.equal(replay.status, "replayed");
  assert.equal(replay.response, null);

  const relaxed = turnEnvelope();
  relaxed.agent_response.requires_confirmation = false;
  globalThis.fetch = async () => jsonResponse(relaxed);
  assert.equal((await executeAgentCopilotSessionTurn(config, session(), turnInput())).failureCode, "application_session_response_invalid");

  globalThis.fetch = async () => jsonResponse({ ...turnEnvelope(), raw_response: "forbidden" });
  assert.equal((await executeAgentCopilotSessionTurn(config, session(), turnInput())).failureCode, "application_session_response_invalid");
});

test("Agent turn maps cancellation without retry", async () => {
  let calls = 0;
  globalThis.fetch = async () => {
    calls += 1;
    throw new DOMException("canceled", "AbortError");
  };
  const result = await executeAgentCopilotSessionTurn(config, session(), turnInput());
  assert.equal(result.failureCode, "application_session_request_canceled");
  assert.equal(calls, 1);
});

function authority(): Record<string, any> {
  return {
    schema_version: "application_runtime_authority.v3",
    execution_profile: "agent_copilot_suggestion_v1",
    application_id: applicationId,
    application_record_version: 1,
    application_lifecycle: "active",
    agent_copilot: {
      assignment_id: "acra_aaaaaaaaaaaaaaaa",
      assignment_version: 1,
      project: "radishflow",
      agent_copilot_profile_ref: {
        profile_id: "acpf_aaaaaaaaaaaaaaaa",
        profile_version: 1,
      },
    },
  };
}

function sessionDocument(recordVersion = 1, turnCount = 0): Record<string, any> {
  return {
    schema_version: "application_session.v3",
    session_id: "appsess_aaaaaaaaaaaaaaaa",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: applicationId,
    state: "active",
    record_version: recordVersion,
    profile_binding: { execution_profile: "agent_copilot_suggestion_v1" },
    authority: authority(),
    content_retention: "metadata_only",
    turn_count: turnCount,
  };
}

function createEnvelope(): Record<string, any> {
  return {
    request_id: "agent-session-create",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: applicationId,
    session: sessionDocument(),
    current_record_version: 1,
    current_state: "active",
    idempotent_replay: false,
    failure_code: null,
    failure_summary: "",
    audit_ref: "audit_agent_session_create",
  };
}

function turnEnvelope(): Record<string, any> {
  return {
    request_id: "agent-session-turn",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: applicationId,
    session_id: "appsess_aaaaaaaaaaaaaaaa",
    session: sessionDocument(2, 1),
    turn: {
      schema_version: "application_session_turn.v3",
      turn_id: "appturn_aaaaaaaaaaaaaaaa",
      session_id: "appsess_aaaaaaaaaaaaaaaa",
      tenant_ref: "tenant_demo",
      workspace_id: "workspace_demo",
      application_id: applicationId,
      execution_profile: "agent_copilot_suggestion_v1",
      authority: authority(),
      sequence: 1,
      status: "succeeded",
      run_ref: { schema_version: "workflow_run_record.v7", run_id: "run_aaaaaaaaaaaaaaaa" },
      failure_code: "",
      failure_summary: "",
    },
    agent_response: {
      schema_version: 1,
      status: "ok",
      project: "radishflow",
      task: "suggest_flowsheet_edits",
      summary: "Review the candidate edit.",
      answers: [{ text: "A bounded edit is available." }],
      issues: [],
      proposed_actions: [{
        kind: "candidate_edit",
        title: "Adjust the selected unit",
        risk_level: "medium",
        requires_confirmation: true,
      }],
      citations: [],
      risk_level: "medium",
      requires_confirmation: true,
    },
    idempotent_replay: false,
    failure_code: null,
    failure_summary: "",
    audit_ref: "audit_agent_session_turn",
  };
}

function session(): AgentCopilotSession {
  return {
    sessionId: "appsess_aaaaaaaaaaaaaaaa",
    applicationId,
    state: "active",
    recordVersion: 1,
    assignmentId: "acra_aaaaaaaaaaaaaaaa",
    assignmentVersion: 1,
    profileId: "acpf_aaaaaaaaaaaaaaaa",
    profileVersion: 1,
    project: "radishflow",
    turnCount: 0,
  };
}

function turnInput() {
  return {
    task: "suggest_flowsheet_edits",
    locale: "zh-CN",
    conversationId: "",
    artifacts: [],
    context: { selected_unit_ids: ["unit-1"] },
    clientTurnKey: "agent-turn-test-001",
  };
}

function jsonResponse(body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
}
