import assert from "node:assert/strict";
import test from "node:test";

import {
  decidePromptApplicationRuntime,
  readPromptApplicationRuntime,
  type PromptApplicationRuntimeConfig,
} from "../src/features/control-plane-read/promptApplicationRuntimeConsumer.ts";

const config: PromptApplicationRuntimeConfig = {
  mode: "dev_prompt_application_http",
  baseUrl: "http://platform.test",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  subjectRef: "subject_demo_user",
};
const applicationId = "app_aaaaaaaaaaaaaaaa";

test("Prompt runtime stays offline without fetching", async () => {
  let calls = 0;
  globalThis.fetch = async () => {
    calls += 1;
    throw new Error("unexpected fetch");
  };
  const offline = { ...config, mode: "offline" as const };
  assert.equal((await readPromptApplicationRuntime(offline, applicationId)).status, "offline");
  assert.equal((await decidePromptApplicationRuntime(offline, applicationId, 0, "activate", "candidate-1")).status, "offline");
  assert.equal(calls, 0);
});

test("Prompt runtime reads exact authority and append-only events", async () => {
  let captured: { url: string; headers: Headers } | undefined;
  globalThis.fetch = async (input, init) => {
    captured = { url: String(input), headers: new Headers(init?.headers) };
    return jsonResponse(runtimeEnvelope());
  };

  const result = await readPromptApplicationRuntime(config, applicationId, true);
  assert.equal(result.status, "ready");
  assert.equal(result.assignment?.assignmentVersion, 1);
  assert.equal(result.assignment?.promptTemplateRef.templateVersion, 2);
  assert.equal(result.events[0]?.action, "activate");
  assert.equal(captured?.url.endsWith("/prompt-runtime-assignment/events?workspace_id=workspace_demo"), true);
  assert.equal(captured?.headers.get("X-RadishMind-Dev-Read-Scopes"), "prompt_application_runtime:read");
  assert.equal(captured?.headers.get("X-RadishMind-Dev-Prompt-Runtime-Application"), applicationId);
});

test("Prompt runtime decision sends only CAS fields and preserves conflict state", async () => {
  let body: any;
  globalThis.fetch = async (_input, init) => {
    body = JSON.parse(String(init?.body));
    const envelope = runtimeEnvelope();
    envelope.assignment = null;
    envelope.events = [];
    envelope.failure_code = "prompt_runtime_assignment_version_conflict";
    envelope.current_assignment_version = 3;
    envelope.current_state = "active";
    return jsonResponse(envelope);
  };

  const result = await decidePromptApplicationRuntime(config, applicationId, 2, "replace", "candidate-prompt-v2");
  assert.deepEqual(body, {
    workspace_id: "workspace_demo",
    expected_assignment_version: 2,
    action: "replace",
    candidate_id: "candidate-prompt-v2",
  });
  assert.equal(result.status, "version_conflict");
  assert.equal(result.currentAssignmentVersion, 3);
});

test("Prompt runtime rejects scope drift and broken event continuity", async () => {
  const drift = runtimeEnvelope();
  drift.application_id = "app_bbbbbbbbbbbbbbbb";
  globalThis.fetch = async () => jsonResponse(drift);
  assert.equal((await readPromptApplicationRuntime(config, applicationId)).failureCode, "prompt_runtime_store_unavailable");

  const broken = runtimeEnvelope();
  broken.events[0].event_sequence = 2;
  globalThis.fetch = async () => jsonResponse(broken);
  assert.equal((await readPromptApplicationRuntime(config, applicationId)).status, "failed");
});

function runtimeEnvelope() {
  const templateRef = {
    template_id: "ptpl_aaaaaaaaaaaaaaaa",
    template_version: 2,
    template_digest: `sha256:${"b".repeat(64)}`,
  };
  const assignment = {
    schema_version: "prompt_application_runtime_assignment.v1",
    assignment_id: "ptra_aaaaaaaaaaaaaaaa",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: applicationId,
    owner_subject_ref: "subject_demo_user",
    assignment_version: 1,
    state: "active",
    candidate_id: "candidate-prompt-v1",
    candidate_review_version: 1,
    draft_id: "draft-prompt-v1",
    draft_version: 2,
    draft_digest: `sha256:${"a".repeat(64)}`,
    prompt_template_ref: templateRef,
    assignment_digest: `sha256:${"c".repeat(64)}`,
    activated_at: "2026-07-25T10:00:00Z",
    updated_at: "2026-07-25T10:00:00Z",
    revoked_at: null,
    activated_by_actor_ref: "subject_demo_user",
    updated_by_actor_ref: "subject_demo_user",
    request_id: "prompt-runtime-request",
    audit_ref: "audit-prompt-runtime-request",
  };
  return {
    request_id: "prompt-runtime-request",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: applicationId,
    assignment,
    events: [{
      schema_version: "prompt_application_runtime_assignment_event.v1",
      event_id: "ptrae_aaaaaaaaaaaaaaaa",
      assignment_id: assignment.assignment_id,
      tenant_ref: "tenant_demo",
      workspace_id: "workspace_demo",
      application_id: applicationId,
      owner_subject_ref: "subject_demo_user",
      event_sequence: 1,
      action: "activate",
      expected_assignment_version: 0,
      resulting_assignment_version: 1,
      candidate_id: assignment.candidate_id,
      candidate_review_version: 1,
      draft_id: assignment.draft_id,
      draft_version: 2,
      draft_digest: assignment.draft_digest,
      prompt_template_ref: templateRef,
      assignment_digest: assignment.assignment_digest,
      occurred_at: "2026-07-25T10:00:00Z",
      actor_ref: "subject_demo_user",
      request_id: "prompt-runtime-request",
      audit_ref: "audit-prompt-runtime-request",
    }],
    failure_code: null,
    current_assignment_version: 1,
    current_state: "active",
    audit_ref: "audit-prompt-runtime-request",
  };
}

function jsonResponse(body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
}
