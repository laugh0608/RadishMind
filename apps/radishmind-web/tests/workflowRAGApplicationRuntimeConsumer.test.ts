import assert from "node:assert/strict";
import test from "node:test";

import {
  decideWorkflowRAGApplicationRuntimeAssignment,
  invokeWorkflowRAGApplication,
  readWorkflowRAGApplicationRuntimeAssignment,
  type WorkflowRAGApplicationRuntimeConfig,
} from "../src/features/control-plane-read/workflowRAGApplicationRuntimeConsumer.ts";
import {
  createWorkflowRAGApplicationCredentialHandoffDetail,
} from "../src/features/control-plane-read/workflowRAGApplicationRuntimeEvents.ts";

const config: WorkflowRAGApplicationRuntimeConfig = {
  mode: "dev_workflow_rag_application_runtime_http",
  baseUrl: "http://platform.test",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  subjectRef: "subject_demo_user",
};
const applicationId = "app_abcdefghijklmnop";
const apiKeyId = "key_abcdefghijklmnop";
const token = `rmd_dev_key_abcdefghijklmnop.${"A".repeat(43)}`;

test("application RAG runtime stays offline without fetching", async () => {
  let requests = 0;
  globalThis.fetch = async () => { requests += 1; throw new Error("offline request"); };
  const offline = { ...config, mode: "offline" as const };
  await readWorkflowRAGApplicationRuntimeAssignment(offline, applicationId);
  await decideWorkflowRAGApplicationRuntimeAssignment(offline, {
    applicationId, expectedRecordVersion: 0, decision: "activate", publishCandidateId: "candidate-approved-0001", reason: "Approved runtime selection.",
  });
  await invokeWorkflowRAGApplication(offline, { applicationId, apiKeyId, token, text: "What is the reviewed answer?" });
  assert.equal(requests, 0);
});

test("assignment read and decision keep exact management scope and CAS", async () => {
  const requests: Array<{ url: string; headers: Headers; body: unknown }> = [];
  globalThis.fetch = async (input, init) => {
    requests.push({ url: String(input), headers: new Headers(init?.headers), body: init?.body ? JSON.parse(String(init.body)) : null });
    return jsonResponse(runtimeEnvelope());
  };
  const read = await readWorkflowRAGApplicationRuntimeAssignment(config, applicationId);
  assert.equal(read.assignment?.assignmentId, "wragra_abcdefghijklmnop");
  assert.equal(requests[0]?.headers.get("X-RadishMind-Dev-Read-Scopes"), "workflow_rag_runtime:read");
  assert.equal(requests[0]?.headers.get("X-RadishMind-Dev-Workflow-Workspace"), "workspace_demo");
  assert.equal(requests[0]?.headers.get("X-RadishMind-Dev-Workflow-Application"), "app_abcdefghijklmnop");

  const decision = await decideWorkflowRAGApplicationRuntimeAssignment(config, {
    applicationId, expectedRecordVersion: 1, decision: "replace", publishCandidateId: "candidate-approved-0002", reason: "Replace with reviewed candidate evidence.",
  });
  assert.equal(decision.status, "ready");
  assert.equal(requests[1]?.headers.get("X-RadishMind-Dev-Read-Scopes"), "workflow_rag_runtime:write");
  assert.deepEqual(requests[1]?.body, {
    workspace_id: "workspace_demo", expected_record_version: 1, decision: "replace",
    publish_candidate_id: "candidate-approved-0002", reason: "Replace with reviewed candidate evidence.",
  });
});

test("assignment consumer preserves version conflict and rejects response leakage", async () => {
  globalThis.fetch = async () => jsonResponse({ ...runtimeEnvelope(), assignment: null, failure_code: "workflow_rag_runtime_assignment_version_conflict", current_record_version: 3, current_state: "active" });
  const conflict = await decideWorkflowRAGApplicationRuntimeAssignment(config, {
    applicationId, expectedRecordVersion: 1, decision: "revoke", publishCandidateId: "", reason: "Revoke after reviewer decision.",
  });
  assert.equal(conflict.status, "version_conflict");
  assert.equal(conflict.currentRecordVersion, 3);

  globalThis.fetch = async () => jsonResponse({ ...runtimeEnvelope(), token: "forbidden" });
  const leaked = await readWorkflowRAGApplicationRuntimeAssignment(config, applicationId);
  assert.equal(leaked.failureCode, "workflow_rag_runtime_store_unavailable");
});

test("invocation sends only Bearer and bounded input then maps transient answer", async () => {
  let captured: { headers: Headers; body: unknown } | undefined;
  globalThis.fetch = async (_input, init) => {
    captured = { headers: new Headers(init?.headers), body: JSON.parse(String(init?.body)) };
    return jsonResponse(invocationEnvelope());
  };
  const result = await invokeWorkflowRAGApplication(config, { applicationId, apiKeyId, token, text: "What is the reviewed answer?" });
  assert.equal(result.status, "succeeded");
  assert.equal(result.runId, "run_abcdefghijklmnop");
  assert.equal(result.answer?.citations[0]?.fragmentRef, "official_manual");
  assert.equal(captured?.headers.get("Authorization"), `Bearer ${token}`);
  assert.equal(captured?.headers.has("X-RadishMind-Dev-Read-Identity"), false);
  assert.deepEqual(captured?.body, { input: "What is the reviewed answer?" });
});

test("invocation rejects forbidden input and invalid response without retry", async () => {
  let requests = 0;
  globalThis.fetch = async () => { requests += 1; return jsonResponse({ ...invocationEnvelope(), raw_response: "forbidden" }); };
  const invalidInput = await invokeWorkflowRAGApplication(config, { applicationId, apiKeyId, token, text: "Authorization: Bearer hidden" });
  assert.equal(invalidInput.failureCode, "workflow_rag_runtime_secret_material_forbidden");
  assert.equal(requests, 0);
  const invalidResponse = await invokeWorkflowRAGApplication(config, { applicationId, apiKeyId, token, text: "What is the reviewed answer?" });
  assert.equal(invalidResponse.failureCode, "workflow_rag_runtime_store_unavailable");
  assert.equal(requests, 1);
});

test("one-time application RAG handoff validates exact application and credential", () => {
  assert.deepEqual(createWorkflowRAGApplicationCredentialHandoffDetail(applicationId, apiKeyId, token), { applicationId, apiKeyId, token });
  assert.throws(() => createWorkflowRAGApplicationCredentialHandoffDetail(applicationId, apiKeyId, `${token}x`));
  assert.throws(() => createWorkflowRAGApplicationCredentialHandoffDetail("app_wrong", apiKeyId, token));
});

function runtimeEnvelope() {
  return {
    request_id: "workflow-rag-runtime-read-0001", tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: applicationId,
    assignment: {
      schema_version: "workflow_rag_application_runtime_assignment.v1", assignment_id: "wragra_abcdefghijklmnop", record_version: 1,
      assignment_digest: `sha256:${"a".repeat(64)}`, tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: applicationId,
      owner_subject_ref: "subject_demo_user", state: "active", publish_candidate_id: "candidate-approved-0001", publish_review_version: 1,
      publish_candidate_state: "approved", draft_id: "application-draft-0001", draft_version: 2, draft_digest: `sha256:${"b".repeat(64)}`,
      binding_ref: { binding_id: "wragb_abcdefghijklmnop", binding_version: 1, binding_digest: `sha256:${"c".repeat(64)}` },
      created_at: "2026-07-19T10:00:00Z", updated_at: "2026-07-19T10:00:00Z", created_by_actor_ref: "subject_demo_user",
      updated_by_actor_ref: "subject_demo_user", request_id: "workflow-rag-runtime-read-0001", audit_ref: "audit_workflow-rag-runtime-read-0001",
    },
    events: [], audits: [], failure_code: null, current_record_version: 1, current_state: "active", audit_ref: "audit_workflow-rag-runtime-read-0001",
  };
}

function invocationEnvelope() {
  return {
    request_id: "workflow-rag-application-invocation-0001", tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: applicationId,
    run: { schema_version: "workflow_run_record.v4", run_id: "run_abcdefghijklmnop", status: "succeeded" },
    answer: {
      schema_version: "workflow_rag_application_answer.v1", answer: "The reviewed answer is supported by the official manual.",
      citations: [{ fragment_ref: "official_manual", claim_summary: "The official manual supports this answer." }], limitations: [], confidence: "high",
    },
    failure_code: null, failure_summary: "", audit_ref: "audit_workflow-rag-application-invocation-0001",
  };
}

function jsonResponse(value: unknown): Response {
  return new Response(JSON.stringify(value), { status: 200, headers: { "Content-Type": "application/json" } });
}
