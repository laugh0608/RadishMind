import assert from "node:assert/strict";
import test from "node:test";

import {
  invokePromptApplication,
  parsePromptApplicationVariables,
} from "../src/features/control-plane-read/promptApplicationInvocationConsumer.ts";
import { createPromptApplicationCredentialHandoffDetail } from "../src/features/control-plane-read/promptApplicationInvocationEvents.ts";

const token = `rmd_dev_key_aaaaaaaaaaaaaaaa.${"A".repeat(43)}`;
const applicationId = "app_aaaaaaaaaaaaaaaa";

test("Prompt invocation validates transient variables and one-time handoff scope", () => {
  assert.deepEqual(parsePromptApplicationVariables('{"question":"审查发布","tone":"清晰"}'), {
    isValid: true,
    failureCode: "",
    variables: { question: "审查发布", tone: "清晰" },
  });
  assert.equal(parsePromptApplicationVariables('{"question":null}').isValid, false);
  assert.equal(parsePromptApplicationVariables('{"question":"Authorization: Bearer hidden"}').isValid, false);
  assert.deepEqual(
    createPromptApplicationCredentialHandoffDetail(applicationId, "key_aaaaaaaaaaaaaaaa", token),
    { applicationId, apiKeyId: "key_aaaaaaaaaaaaaaaa", token },
  );
  assert.throws(
    () => createPromptApplicationCredentialHandoffDetail("app_other", "key_aaaaaaaaaaaaaaaa", token),
    /invalid/,
  );
});

test("Prompt invocation sends only Bearer plus bounded variables and maps Run v6", async () => {
  let captured: { headers: Headers; body: any } | undefined;
  globalThis.fetch = async (_input, init) => {
    captured = {
      headers: new Headers(init?.headers),
      body: JSON.parse(String(init?.body)),
    };
    return jsonResponse(successEnvelope());
  };

  const result = await invokePromptApplication(
    { baseUrl: "http://platform.test" },
    token,
    { question: "如何审查？", tone: "清晰" },
    "prompt-client-key-0001",
  );
  assert.equal(result.status, "succeeded");
  assert.equal(result.output, "请检查模板、草案与候选摘要。");
  assert.equal(result.run?.templateVersion, 2);
  assert.equal(result.run?.providerCalls, 1);
  assert.equal(captured?.headers.get("Authorization"), `Bearer ${token}`);
  assert.equal(captured?.headers.has("X-RadishMind-Dev-Read-Identity"), false);
  assert.deepEqual(captured?.body, {
    variables: { question: "如何审查？", tone: "清晰" },
    client_invocation_key: "prompt-client-key-0001",
  });
});

test("Prompt invocation replay stays metadata-only and invalid response does not retry", async () => {
  let calls = 0;
  const replay = successEnvelope();
  replay.output = "";
  replay.idempotent_replay = true;
  globalThis.fetch = async () => {
    calls += 1;
    return jsonResponse(replay);
  };
  const replayed = await invokePromptApplication(
    { baseUrl: "http://platform.test" },
    token,
    { question: "如何审查？" },
    "prompt-client-key-0001",
  );
  assert.equal(replayed.status, "replayed");
  assert.equal(replayed.output, "");

  const leaked = successEnvelope();
  (leaked.run as any).messages = [{ role: "user", content: "forbidden" }];
  globalThis.fetch = async () => {
    calls += 1;
    return jsonResponse(leaked);
  };
  const rejected = await invokePromptApplication(
    { baseUrl: "http://platform.test" },
    token,
    { question: "如何审查？" },
    "prompt-client-key-0002",
  );
  assert.equal(rejected.failureCode, "prompt_invocation_response_invalid");
  assert.equal(calls, 2);
});

function successEnvelope() {
  const digest = (character: string) => `sha256:${character.repeat(64)}`;
  return {
    request_id: "prompt-invocation-request",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: applicationId,
    run: {
      schema_version: "workflow_run_record.v6",
      record_version: 1,
      run_id: "run_aaaaaaaaaaaaaaaa",
      tenant_ref: "tenant_demo",
      workspace_id: "workspace_demo",
      application_id: applicationId,
      execution_kind: "prompt_application_invocation",
      execution_source_kind: "prompt_application_template",
      execution_source_id: "ptra_aaaaaaaaaaaaaaaa",
      execution_source_version: 1,
      execution_profile: "prompt_application_invocation_v1",
      authority: {
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
          prompt_template_ref: {
            template_id: "ptpl_aaaaaaaaaaaaaaaa",
            template_version: 2,
            template_digest: digest("c"),
          },
          default_protocol: "responses",
          default_model: "radishmind-local-dev",
          protocol_policy_digest: digest("d"),
          model_eligibility_digest: digest("e"),
        },
        authority_digest: digest("f"),
      },
      input_digest: digest("1"),
      input_bytes: 48,
      variable_names: ["question", "tone"],
      variable_names_digest: digest("2"),
      requested_protocol: "responses",
      selected_protocol: "responses",
      requested_model: "radishmind-local-dev",
      selected_provider: "mock",
      selected_profile: "default",
      selected_model: "radishmind-local-dev",
      upstream_model: "radishmind-local-dev",
      selection_source: "application_runtime_authority",
      status: "succeeded",
      failure_code: "",
      failure_summary: "",
      started_at: "2026-07-25T10:00:00Z",
      completed_at: "2026-07-25T10:00:01Z",
      output: "",
      usage: { state: "provider_reported", input_tokens: 10, output_tokens: 8, total_tokens: 18 },
      side_effects: { retrieval_calls: 0, provider_calls: 1, tool_calls: 0, confirmation_calls: 0, business_writes: 0, replay_writes: 0 },
      diagnostic: {
        failure_boundary: "",
        failure_stage: "",
        terminal_write_state: "stored",
        gateway_failure_category: "none",
        summary: "",
        recommended_review_action: "",
        observed_at: "2026-07-25T10:00:01Z",
      },
      request_id: "prompt-invocation-request",
      audit_ref: "audit-prompt-invocation-request",
      actor_ref: "subject_demo_user",
    },
    output: "请检查模板、草案与候选摘要。",
    failure_code: null,
    failure_summary: "",
    idempotent_replay: false,
    audit_ref: "audit-prompt-invocation-request",
  };
}

function jsonResponse(body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
}
