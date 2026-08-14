import assert from "node:assert/strict";
import test from "node:test";

import {
  createAgentCopilotProfileInput,
  readAgentCopilotProfileVersion,
  saveAgentCopilotProfile,
  validateAgentCopilotProfileLocally,
  type AgentCopilotProfileConfig,
  type AgentCopilotProfileInput,
} from "../src/features/control-plane-read/agentCopilotProfileConsumer.ts";

const config: AgentCopilotProfileConfig = {
  mode: "dev_agent_copilot_http",
  baseUrl: "http://platform.test",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  subjectRef: "subject_demo_user",
};
const applicationId = "app_aaaaaaaaaaaaaaaa";

test("Agent Profile local review is deterministic and rejects safety relaxation or secret material", () => {
  const input = validInput();
  assert.deepEqual(validateAgentCopilotProfileLocally(input), validateAgentCopilotProfileLocally(input));
  assert.equal(validateAgentCopilotProfileLocally(input).isValid, true);

  const relaxed = structuredClone(input) as AgentCopilotProfileInput;
  (relaxed.toolHintsPolicy as { allowToolCalls: boolean }).allowToolCalls = true;
  assert.equal(validateAgentCopilotProfileLocally(relaxed).findings[0]?.code, "agent_copilot_profile_policy_invalid");

  const secret = { ...input, description: "Authorization: Bearer secret-value" };
  assert.equal(validateAgentCopilotProfileLocally(secret).findings[0]?.code, "agent_copilot_profile_secret_material_forbidden");
});

test("Agent Profile save carries exact CAS source and immutable policy fields", async () => {
  const input = validInput();
  let request: { headers: Headers; body: any } | null = null;
  globalThis.fetch = async (_url, init) => {
    request = { headers: new Headers(init?.headers), body: JSON.parse(String(init?.body)) };
    return jsonResponse(profileEnvelope({ draft: profileDocument(input) }));
  };
  const result = await saveAgentCopilotProfile(config, input, 0);
  assert.equal(result.status, "saved");
  assert.equal(result.draft?.profileId, input.profileId);
  assert.equal(request?.headers.get("X-RadishMind-Dev-Read-Scopes"), "agent_copilot_profiles:write");
  assert.equal(request?.headers.get("X-RadishMind-Active-Workspace"), "workspace_demo");
  assert.equal(request?.headers.get("X-RadishMind-Dev-Read-Membership-Permissions"), "agent_copilot_profiles:write");
  assert.equal(request?.body.expected_draft_version, 0);
  assert.deepEqual(request?.body.profile.tool_hints_policy, {
    allow_retrieval: false,
    allow_tool_calls: false,
    allow_image_reasoning: false,
  });
  assert.deepEqual(request?.body.profile.risk_policy, {
    mode: "advisory",
    requires_confirmation_for_actions: true,
    confirmation_action_kinds: ["candidate_edit", "candidate_operation", "ghost_completion"],
  });
});

test("exact Profile version source requires read_source and rejects sensitive response drift", async () => {
  const input = validInput();
  let scope = "";
  globalThis.fetch = async (_url, init) => {
    scope = new Headers(init?.headers).get("X-RadishMind-Dev-Read-Scopes") ?? "";
    return jsonResponse(profileEnvelope({ version: profileDocument(input, true) }));
  };
  const result = await readAgentCopilotProfileVersion(config, applicationId, input.profileId, 1);
  assert.equal(scope, "agent_copilot_profiles:read_source");
  assert.equal(result.version?.profileVersion, 1);

  globalThis.fetch = async () => jsonResponse({
    ...profileEnvelope({ version: profileDocument(input, true) }),
    provider_raw_response: "forbidden",
  });
  assert.equal(
    (await readAgentCopilotProfileVersion(config, applicationId, input.profileId, 1)).failureCode,
    "agent_copilot_profile_store_unavailable",
  );
});

function validInput(): AgentCopilotProfileInput {
  return {
    ...createAgentCopilotProfileInput(config, applicationId),
    profileId: "acpf_aaaaaaaaaaaaaaaa",
  };
}

function profileDocument(input: AgentCopilotProfileInput, immutable = false): Record<string, unknown> {
  const profile: Record<string, unknown> = {
    schema_version: input.schemaVersion,
    profile_id: input.profileId,
    workspace_id: input.workspaceId,
    application_id: input.applicationId,
    profile_name: input.profileName,
    description: input.description,
    project: input.project,
    allowed_tasks: input.allowedTasks,
    default_locale: input.defaultLocale,
    allowed_locales: input.allowedLocales,
    context_policy: {
      allowed_fields: input.contextPolicy.allowedFields,
      max_bytes: input.contextPolicy.maxBytes,
      require_task_context: input.contextPolicy.requireTaskContext,
    },
    artifact_policy: {
      allowed_kinds: input.artifactPolicy.allowedKinds,
      allowed_roles: input.artifactPolicy.allowedRoles,
      max_count: input.artifactPolicy.maxCount,
      max_item_bytes: input.artifactPolicy.maxItemBytes,
      max_total_bytes: input.artifactPolicy.maxTotalBytes,
    },
    response_policy: {
      allowed_action_kinds: input.responsePolicy.allowedActionKinds,
      max_answers: input.responsePolicy.maxAnswers,
      max_issues: input.responsePolicy.maxIssues,
      max_actions: input.responsePolicy.maxActions,
      max_citations: input.responsePolicy.maxCitations,
      max_visible_text_bytes: input.responsePolicy.maxVisibleTextBytes,
    },
    risk_policy: {
      mode: "advisory",
      requires_confirmation_for_actions: true,
      confirmation_action_kinds: input.riskPolicy.confirmationActionKinds,
    },
    tool_hints_policy: {
      allow_retrieval: false,
      allow_tool_calls: false,
      allow_image_reasoning: false,
    },
    profile_digest: `sha256:${"a".repeat(64)}`,
    policy_digest: `sha256:${"b".repeat(64)}`,
  };
  if (immutable) {
    return {
      ...profile,
      schema_version: "agent_copilot_profile_version.v1",
      profile_version: 1,
      source_draft_version: 1,
      published_at: "2026-07-25T10:00:00Z",
      published_by_actor_ref: "subject_demo_user",
    };
  }
  return {
    ...profile,
    draft_version: 1,
    validation_summary: validValidation(),
    updated_at: "2026-07-25T10:00:00Z",
    updated_by_actor_ref: "subject_demo_user",
  };
}

function profileEnvelope(partial: Record<string, unknown>): Record<string, unknown> {
  return {
    request_id: "agent-profile-test",
    workspace_id: "workspace_demo",
    application_id: applicationId,
    draft: null,
    version: null,
    validation_summary: validValidation(),
    current_draft_version: 1,
    current_profile_version: 1,
    failure_code: null,
    failure_summary: "",
    audit_ref: "audit_agent_profile_test",
    ...partial,
  };
}

function validValidation(): Record<string, unknown> {
  return { is_valid: true, findings: [] };
}

function jsonResponse(body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
}
