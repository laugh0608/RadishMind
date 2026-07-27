import assert from "node:assert/strict";
import test from "node:test";

import {
  decideAgentCopilotRuntime,
  readAgentCopilotRuntime,
  type AgentCopilotRuntimeConfig,
} from "../src/features/control-plane-read/agentCopilotRuntimeConsumer.ts";

const config: AgentCopilotRuntimeConfig = {
  mode: "dev_agent_copilot_http",
  baseUrl: "http://platform.test",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  subjectRef: "subject_demo_user",
};
const applicationId = "app_aaaaaaaaaaaaaaaa";

test("Agent runtime reads exact assignment lineage and append-only events", async () => {
  globalThis.fetch = async () => jsonResponse(runtimeEnvelope());
  const result = await readAgentCopilotRuntime(config, applicationId);
  assert.equal(result.status, "ready");
  assert.equal(result.assignment?.profile.profileVersion, 1);
  assert.equal(result.events[0]?.resultingAssignmentVersion, 1);
});

test("Agent runtime decision sends only CAS fields and preserves a version conflict", async () => {
  let body: any;
  let headers: Headers | undefined;
  globalThis.fetch = async (_url, init) => {
    body = JSON.parse(String(init?.body));
    headers = new Headers(init?.headers);
    return jsonResponse({
      ...runtimeEnvelope(),
      assignment: null,
      events: [],
      current_assignment_version: 2,
      current_state: "active",
      failure_code: "agent_copilot_runtime_assignment_version_conflict",
    });
  };
  const result = await decideAgentCopilotRuntime(config, applicationId, 1, "replace", "candidate-agent-v2");
  assert.deepEqual(body, {
    workspace_id: "workspace_demo",
    expected_assignment_version: 1,
    action: "replace",
    candidate_id: "candidate-agent-v2",
  });
  assert.equal(headers?.get("X-RadishMind-Active-Workspace"), "workspace_demo");
  assert.equal(headers?.get("X-RadishMind-Dev-Read-Membership-Workspace"), "workspace_demo");
  assert.equal(headers?.get("X-RadishMind-Dev-Read-Membership-Permissions"), "agent_copilot_runtime:write");
  assert.equal(result.status, "version_conflict");
  assert.equal(result.currentAssignmentVersion, 2);
});

test("Agent runtime rejects scope drift, broken event continuity, and sensitive material", async () => {
  globalThis.fetch = async () => jsonResponse({ ...runtimeEnvelope(), application_id: "app_bbbbbbbbbbbbbbbb" });
  assert.equal((await readAgentCopilotRuntime(config, applicationId)).status, "failed");

  const broken = runtimeEnvelope();
  (broken.events as Array<Record<string, unknown>>).push({
    ...(broken.events as Array<Record<string, unknown>>)[0],
    event_id: "acrae_bbbbbbbbbbbbbbbb",
    event_sequence: 3,
  });
  globalThis.fetch = async () => jsonResponse(broken);
  assert.equal((await readAgentCopilotRuntime(config, applicationId)).status, "failed");

  globalThis.fetch = async () => jsonResponse({ ...runtimeEnvelope(), raw_response: "forbidden" });
  assert.equal((await readAgentCopilotRuntime(config, applicationId)).status, "failed");
});

function runtimeEnvelope(): Record<string, any> {
  return {
    request_id: "agent-runtime-test",
    workspace_id: "workspace_demo",
    application_id: applicationId,
    assignment: {
      schema_version: "agent_copilot_runtime_assignment.v1",
      assignment_id: "acra_aaaaaaaaaaaaaaaa",
      assignment_version: 1,
      workspace_id: "workspace_demo",
      application_id: applicationId,
      state: "active",
      candidate_id: "candidate-agent-v1",
      candidate_review_version: 1,
      draft_id: "draft-agent-v4",
      draft_version: 2,
      draft_digest: `sha256:${"a".repeat(64)}`,
      agent_copilot_profile_ref: {
        profile_id: "acpf_aaaaaaaaaaaaaaaa",
        profile_version: 1,
        profile_digest: `sha256:${"b".repeat(64)}`,
        policy_digest: `sha256:${"c".repeat(64)}`,
      },
      assignment_digest: `sha256:${"d".repeat(64)}`,
      activated_at: "2026-07-25T10:00:00Z",
      updated_at: "2026-07-25T10:00:00Z",
      revoked_at: null,
    },
    events: [{
      schema_version: "agent_copilot_runtime_assignment_event.v1",
      event_id: "acrae_aaaaaaaaaaaaaaaa",
      workspace_id: "workspace_demo",
      application_id: applicationId,
      event_sequence: 1,
      action: "activate",
      expected_assignment_version: 0,
      resulting_assignment_version: 1,
      candidate_id: "candidate-agent-v1",
      occurred_at: "2026-07-25T10:00:00Z",
    }],
    current_assignment_version: 1,
    current_state: "active",
    failure_code: null,
    failure_summary: "",
    audit_ref: "audit_agent_runtime_test",
  };
}

function jsonResponse(body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
}
