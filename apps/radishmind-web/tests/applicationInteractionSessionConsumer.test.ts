import assert from "node:assert/strict";
import test from "node:test";

import {
  applicationInteractionResponseMatchesScope,
  closeApplicationInteractionSession,
  createApplicationInteractionSession,
  executeApplicationInteractionTurn,
  listApplicationInteractionSessions,
  listApplicationInteractionTurns,
  type ApplicationInteractionSessionConfig,
} from "../src/features/control-plane-read/applicationInteractionSessionConsumer.ts";

const config: ApplicationInteractionSessionConfig = {
  mode: "dev_application_session_http",
  baseUrl: "http://platform.test",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  subjectRef: "subject_demo_user",
};
const applicationId = "app_abcdefghijklmnop";
const sessionId = "appsess_abcdefghijklmnop";

test("application session consumer stays offline with zero requests", async () => {
  let requests = 0;
  globalThis.fetch = async () => { requests += 1; throw new Error("offline request"); };
  const offline = { ...config, mode: "offline" as const };
  const listed = await listApplicationInteractionSessions(offline, applicationId);
  const created = await createApplicationInteractionSession(offline, {
    applicationId, executionProfile: "application_rag_invocation_v1", definitionId: "",
  });
  assert.equal(listed.status, "offline");
  assert.equal(created.status, "offline");
  assert.equal(requests, 0);
});

test("session list and create use exact application scope and explicit profile", async () => {
  const requests: Array<{ url: string; headers: Headers; body: unknown }> = [];
  globalThis.fetch = async (input, init) => {
    requests.push({ url: String(input), headers: new Headers(init?.headers), body: init?.body ? JSON.parse(String(init.body)) : null });
    return requests.length === 1
      ? jsonResponse(sessionListEnvelope([workflowSession()]))
      : jsonResponse(sessionEnvelope(workflowSession()));
  };

  const listed = await listApplicationInteractionSessions(config, applicationId);
  assert.equal(listed.status, "ready");
  assert.equal(listed.sessions[0]?.authority.workflowDefinition?.definitionVersion, 3);
  assert.equal(requests[0]?.headers.get("X-RadishMind-Dev-Read-Scopes"), "application_sessions:read");
  assert.match(requests[0]?.url ?? "", /workspace_id=workspace_demo/u);
  assert.match(requests[0]?.url ?? "", /application_id=app_abcdefghijklmnop/u);

  const created = await createApplicationInteractionSession(config, {
    applicationId, executionProfile: "workflow_definition_executor_v1", definitionId: "definition_support_flow",
  });
  assert.equal(created.status, "ready");
  assert.equal(requests[1]?.headers.get("X-RadishMind-Dev-Read-Scopes"), "application_sessions:write");
  assert.deepEqual(requests[1]?.body, {
    workspace_id: "workspace_demo",
    application_id: applicationId,
    execution_profile: "workflow_definition_executor_v1",
    definition_id: "definition_support_flow",
  });
});

test("session consumer rejects schema, scope, and sensitive response drift", async () => {
  const invalidDocuments = [
    { ...sessionListEnvelope([workflowSession()]), unexpected: true },
    { ...sessionListEnvelope([workflowSession()]), application_id: "app_ponmlkjihgfedcba" },
    { ...sessionListEnvelope([workflowSession()]), token: "forbidden" },
    sessionListEnvelope([{ ...workflowSession(), input_text: "must not echo" }]),
    sessionListEnvelope([{ ...workflowSession(), owner_subject_ref: "subject_wrong_user" }]),
  ];
  for (const document of invalidDocuments) {
    globalThis.fetch = async () => jsonResponse(document);
    const result = await listApplicationInteractionSessions(config, applicationId);
    assert.equal(result.failureCode, "application_session_store_contract_mismatch");
  }
});

test("turn execution keeps input request-only and returns transient v5 output", async () => {
  let captured: { headers: Headers; body: Record<string, unknown> } | undefined;
  globalThis.fetch = async (_input, init) => {
    captured = { headers: new Headers(init?.headers), body: JSON.parse(String(init?.body)) };
    return jsonResponse(turnEnvelope(workflowSession(2), workflowTurn()));
  };
  const result = await executeApplicationInteractionTurn(config, {
    session: parsedWorkflowSession(),
    clientTurnKey: "web_turn_001",
    inputText: "Produce a bounded advisory recommendation.",
  });
  assert.equal(result.status, "succeeded");
  assert.equal(result.turn?.runRef?.schemaVersion, "workflow_run_record.v5");
  assert.equal(result.advisoryOutput, "Review this bounded advisory recommendation.");
  assert.equal(captured?.headers.get("X-RadishMind-Dev-Read-Scopes"), "application_sessions:execute");
  assert.deepEqual(captured?.body, {
    workspace_id: "workspace_demo",
    application_id: applicationId,
    expected_session_version: 1,
    client_turn_key: "web_turn_001",
    input_text: "Produce a bounded advisory recommendation.",
    condition_values: {},
    model: "",
    temperature: null,
  });
});

test("RAG turn metadata requires v4 and maps answer only from current response", async () => {
  globalThis.fetch = async () => jsonResponse(turnEnvelope(ragSession(2), ragTurn(), {
    advisory_output: "",
    answer: {
      schema_version: "workflow_rag_application_answer.v1",
      answer: "The reviewed evidence supports this response.",
      citations: [{ fragment_ref: "official_manual", claim_summary: "The manual contains the reviewed evidence." }],
      limitations: ["This is advisory."],
      confidence: "high",
    },
  }));
  const result = await executeApplicationInteractionTurn(config, {
    session: parsedRAGSession(), clientTurnKey: "web_turn_rag_001", inputText: "What does the reviewed evidence say?",
  });
  assert.equal(result.status, "succeeded");
  assert.equal(result.turn?.runRef?.schemaVersion, "workflow_run_record.v4");
  assert.equal(result.answer?.citations[0]?.fragmentRef, "official_manual");
  assert.equal(result.advisoryOutput, "");
});

test("turn consumer rejects secret input before fetch and invalid terminal contract", async () => {
  let requests = 0;
  globalThis.fetch = async () => {
    requests += 1;
    return jsonResponse(turnEnvelope(workflowSession(2), { ...workflowTurn(), run_ref: { run_id: "run_abcdefghijklmnop", schema_version: "workflow_run_record.v4" } }));
  };
  const forbidden = await executeApplicationInteractionTurn(config, {
    session: parsedWorkflowSession(), clientTurnKey: "web_turn_secret", inputText: "Authorization: Bearer hidden",
  });
  assert.equal(forbidden.failureCode, "application_session_secret_material_forbidden");
  assert.equal(requests, 0);
  const forbiddenModel = await executeApplicationInteractionTurn(config, {
    session: parsedWorkflowSession(), clientTurnKey: "web_turn_secret_model", inputText: "Safe bounded input.", model: "token=hidden",
  });
  assert.equal(forbiddenModel.failureCode, "application_session_secret_material_forbidden");
  assert.equal(requests, 0);
  const mismatch = await executeApplicationInteractionTurn(config, {
    session: parsedWorkflowSession(), clientTurnKey: "web_turn_contract", inputText: "Safe bounded input.",
  });
  assert.equal(mismatch.failureCode, "application_session_store_contract_mismatch");
  assert.equal(requests, 1);
});

test("turn history and close preserve metadata-only CAS", async () => {
  let operation = "turns";
  globalThis.fetch = async () => operation === "turns"
    ? jsonResponse(turnListEnvelope([workflowTurn()]))
    : jsonResponse(sessionEnvelope({ ...workflowSession(2), state: "closed", closed_at: "2026-07-19T11:00:00Z" }));
  const session = parsedWorkflowSession();
  const turns = await listApplicationInteractionTurns(config, session);
  assert.equal(turns.status, "ready");
  assert.equal(turns.turns[0]?.inputBytes, 41);
  assert.equal(Object.hasOwn(turns.turns[0] ?? {}, "inputText"), false);
  operation = "close";
  const closed = await closeApplicationInteractionSession(config, session);
  assert.equal(closed.session?.state, "closed");
  assert.equal(closed.session?.recordVersion, 2);
});

test("application switch, session switch, cancel, and late response invalidate scope", () => {
  const expected = { generation: 4, applicationId, sessionId, clientTurnKey: "web_turn_001" };
  assert.equal(applicationInteractionResponseMatchesScope(expected, { ...expected }), true);
  assert.equal(applicationInteractionResponseMatchesScope(expected, { ...expected, generation: 5 }), false);
  assert.equal(applicationInteractionResponseMatchesScope(expected, { ...expected, applicationId: "app_ponmlkjihgfedcba" }), false);
  assert.equal(applicationInteractionResponseMatchesScope(expected, { ...expected, sessionId: "appsess_ponmlkjihgfedcba" }), false);
  assert.equal(applicationInteractionResponseMatchesScope(expected, { ...expected, clientTurnKey: "web_turn_002" }), false);
});

function parsedWorkflowSession() {
  return {
    schemaVersion: "application_session.v1" as const,
    sessionId,
    tenantRef: "tenant_demo",
    workspaceId: "workspace_demo",
    applicationId,
    ownerSubjectRef: "subject_demo_user",
    state: "active" as const,
    recordVersion: 1,
    executionProfile: "workflow_definition_executor_v1" as const,
    definitionId: "definition_support_flow",
    authority: {
      schemaVersion: "application_runtime_authority.v1" as const,
      executionProfile: "workflow_definition_executor_v1" as const,
      applicationId,
      applicationRecordVersion: 2,
      applicationLifecycle: "active" as const,
      workflowDefinition: {
        definitionId: "definition_support_flow", definitionVersion: 3, definitionDigest: digest("b"), activationPointerVersion: 2,
        candidateId: "candidate_definition_003", candidateReviewVersion: 1,
      },
      applicationRAG: null,
      authorityDigest: digest("a"),
    },
    contentRetention: "metadata_only" as const,
    turnCount: 0,
    lastTurnId: "",
    createdAt: "2026-07-19T10:00:00Z",
    updatedAt: "2026-07-19T10:00:00Z",
    closedAt: "",
    requestId: "application-session-create-0001",
    auditRef: "audit_application-session-create-0001",
  };
}

function parsedRAGSession() {
  const raw = ragSession();
  return {
    ...parsedWorkflowSession(),
    executionProfile: "application_rag_invocation_v1" as const,
    definitionId: "",
    authority: {
      schemaVersion: "application_runtime_authority.v1" as const,
      executionProfile: "application_rag_invocation_v1" as const,
      applicationId,
      applicationRecordVersion: 2,
      applicationLifecycle: "active" as const,
      workflowDefinition: null,
      applicationRAG: {
        assignmentId: raw.authority.application_rag.assignment_id,
        assignmentVersion: 1,
        assignmentDigest: digest("c"),
        publishCandidateId: "publish_candidate_001",
        publishReviewVersion: 1,
        draftId: "application_draft_001",
        draftVersion: 2,
        draftDigest: digest("d"),
        bindingId: "wragb_abcdefghijklmnop",
        bindingVersion: 1 as const,
        bindingDigest: digest("e"),
        snapshotId: "rags_abcdefghijklmnop",
        snapshotVersion: 1,
        snapshotDigest: digest("f"),
        retrievalProfileId: "retrieval_profile_001",
        retrievalProfileVersion: 1,
        retrievalProfileDigest: digest("1"),
      },
      authorityDigest: digest("2"),
    },
  };
}

function workflowSession(recordVersion = 1) {
  return {
    schema_version: "application_session.v1", session_id: sessionId, tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: applicationId,
    owner_subject_ref: "subject_demo_user", state: "active", record_version: recordVersion,
    profile_binding: { execution_profile: "workflow_definition_executor_v1", definition_id: "definition_support_flow" },
    authority: workflowAuthority(), content_retention: "metadata_only", turn_count: recordVersion - 1,
    last_turn_id: recordVersion === 1 ? null : "appturn_abcdefghijklmnop", created_at: "2026-07-19T10:00:00Z", updated_at: "2026-07-19T10:00:00Z",
    closed_at: null, created_by_actor_ref: "subject_demo_user", updated_by_actor_ref: "subject_demo_user",
    request_id: "application-session-create-0001", audit_ref: "audit_application-session-create-0001",
  };
}

function ragSession(recordVersion = 1) {
  return {
    ...workflowSession(recordVersion),
    profile_binding: { execution_profile: "application_rag_invocation_v1" },
    authority: ragAuthority(),
  };
}

function workflowAuthority() {
  return {
    schema_version: "application_runtime_authority.v1", execution_profile: "workflow_definition_executor_v1", application_id: applicationId,
    application_record_version: 2, application_lifecycle: "active",
    workflow_definition: { definition_id: "definition_support_flow", definition_version: 3, definition_digest: digest("b"), activation_pointer_version: 2, candidate_id: "candidate_definition_003", candidate_review_version: 1 },
    application_rag: null, authority_digest: digest("a"),
  };
}

function ragAuthority() {
  return {
    schema_version: "application_runtime_authority.v1", execution_profile: "application_rag_invocation_v1", application_id: applicationId,
    application_record_version: 2, application_lifecycle: "active", workflow_definition: null,
    application_rag: {
      assignment_id: "wragra_abcdefghijklmnop", assignment_version: 1, assignment_digest: digest("c"), publish_candidate_id: "publish_candidate_001",
      publish_review_version: 1, draft_id: "application_draft_001", draft_version: 2, draft_digest: digest("d"),
      binding_ref: { binding_id: "wragb_abcdefghijklmnop", binding_version: 1, binding_digest: digest("e") },
      snapshot_id: "rags_abcdefghijklmnop", snapshot_version: 1, snapshot_digest: digest("f"), retrieval_profile_id: "retrieval_profile_001",
      retrieval_profile_version: 1, retrieval_profile_digest: digest("1"),
    },
    authority_digest: digest("2"),
  };
}

function workflowTurn() {
  return {
    schema_version: "application_session_turn.v1", turn_id: "appturn_abcdefghijklmnop", session_id: sessionId, sequence: 1,
    client_turn_key: "web_turn_001", tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: applicationId,
    owner_subject_ref: "subject_demo_user", execution_profile: "workflow_definition_executor_v1", authority: workflowAuthority(), status: "succeeded",
    input_digest: digest("3"), input_bytes: 41, run_ref: { run_id: "run_abcdefghijklmnop", schema_version: "workflow_run_record.v5" },
    failure_code: "", failure_summary: "", started_at: "2026-07-19T10:01:00Z", completed_at: "2026-07-19T10:01:01Z",
    actor_ref: "subject_demo_user", request_id: "application-session-turn-0001", audit_ref: "audit_application-session-turn-0001",
  };
}

function ragTurn() {
  return {
    ...workflowTurn(), client_turn_key: "web_turn_rag_001", execution_profile: "application_rag_invocation_v1", authority: ragAuthority(),
    run_ref: { run_id: "run_abcdefghijklmnop", schema_version: "workflow_run_record.v4" },
  };
}

function sessionEnvelope(session: unknown) {
  return {
    request_id: "application-session-create-0001", tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: applicationId,
    session, failure_code: null, current_record_version: 1, current_state: "active", idempotent_replay: false,
    audit_ref: "audit_application-session-create-0001",
  };
}

function sessionListEnvelope(items: unknown[]) {
  return {
    request_id: "application-session-list-0001", tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: applicationId,
    items, next_cursor: null, failure_code: null, audit_ref: "audit_application-session-list-0001",
  };
}

function turnListEnvelope(items: unknown[]) {
  return {
    request_id: "application-turn-list-0001", tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: applicationId,
    session_id: sessionId, items, failure_code: null, audit_ref: "audit_application-turn-list-0001",
  };
}

function turnEnvelope(session: unknown, turn: unknown, overrides: Record<string, unknown> = {}) {
  return {
    request_id: "application-session-turn-0001", tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: applicationId,
    session_id: sessionId, session, turn, advisory_output: "Review this bounded advisory recommendation.", failure_code: null,
    failure_summary: "", idempotent_replay: false, audit_ref: "audit_application-session-turn-0001", ...overrides,
  };
}

function digest(character: string) {
  return `sha256:${character.repeat(64)}`;
}

function jsonResponse(value: unknown): Response {
  return new Response(JSON.stringify(value), { status: 200, headers: { "Content-Type": "application/json" } });
}
