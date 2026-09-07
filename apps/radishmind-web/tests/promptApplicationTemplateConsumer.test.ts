import assert from "node:assert/strict";
import test from "node:test";

import {
  createPromptTemplateDraftInput,
  listPromptTemplateDrafts,
  mergePromptTemplateValidationOperation,
  readPromptTemplateVersion,
  renderPromptTemplatePreview,
  savePromptTemplateDraft,
  validatePromptTemplateLocally,
  type PromptTemplateConfig,
} from "../src/features/control-plane-read/promptApplicationTemplateConsumer.ts";

const config: PromptTemplateConfig = {
  mode: "dev_prompt_application_http",
  baseUrl: "http://platform.test",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  subjectRef: "subject_demo_user",
};
const applicationId = "app_aaaaaaaaaaaaaaaa";
const templateId = "ptpl_aaaaaaaaaaaaaaaa";

test("Prompt Template owner stays offline without fetching", async () => {
  let calls = 0;
  globalThis.fetch = async () => {
    calls += 1;
    throw new Error("unexpected fetch");
  };
  const offline = { ...config, mode: "offline" as const };
  const input = createPromptTemplateDraftInput(offline, applicationId);
  assert.equal((await savePromptTemplateDraft(offline, input, 0)).status, "offline");
  assert.equal((await listPromptTemplateDrafts(offline, applicationId)).status, "offline");
  assert.equal(calls, 0);
});

test("Prompt Template local validation and preview are deterministic and fail closed", () => {
  const input = createPromptTemplateDraftInput(config, applicationId);
  input.templateId = templateId;
  assert.equal(validatePromptTemplateLocally(input).isValid, true);

  const preview = renderPromptTemplatePreview(input, {
    question: "如何审查发布？",
    tone: "简洁",
  });
  assert.equal(preview.status, "valid");
  assert.equal(preview.messages[1]?.content, "问题：如何审查发布？");

  const invalidNull = renderPromptTemplatePreview(input, {
    question: null,
    tone: "简洁",
  });
  assert.equal(invalidNull.status, "invalid");
  assert.equal(invalidNull.findings[0]?.code, "prompt_template_variable_invalid");

  input.messages[0]!.content = "Authorization: Bearer hidden {{ tone }}";
  assert.equal(
    validatePromptTemplateLocally(input).findings.some(
      (finding) => finding.code === "prompt_template_secret_material_forbidden",
    ),
    true,
  );
});

test("Prompt Template save sends exact source and CAS scope", async () => {
  const input = createPromptTemplateDraftInput(config, applicationId);
  input.templateId = templateId;
  let captured: { headers: Headers; body: any } | undefined;
  globalThis.fetch = async (_url, init) => {
    captured = {
      headers: new Headers(init?.headers),
      body: JSON.parse(String(init?.body)),
    };
    return jsonResponse({
      ...emptyEnvelope(),
      failure_code: "prompt_template_version_conflict",
      current_draft_version: 2,
    });
  };

  const result = await savePromptTemplateDraft(config, input, 1);
  assert.equal(result.status, "version_conflict");
  assert.equal(result.currentDraftVersion, 2);
  assert.equal(captured?.headers.get("X-RadishMind-Dev-Read-Scopes"), "prompt_application_templates:write");
  assert.equal(captured?.headers.get("X-RadishMind-Active-Workspace"), "workspace_demo");
  assert.equal(captured?.headers.get("X-RadishMind-Dev-Read-Membership-Permissions"), "prompt_application_templates:write");
  assert.equal(captured?.headers.get("X-RadishMind-Dev-Prompt-Template-Application"), applicationId);
  assert.equal(captured?.body.expected_draft_version, 1);
  assert.equal(captured?.body.template.schema_version, "prompt_application_template_draft.v1");
  assert.equal(captured?.body.template.messages[0].content, "请使用 {{ tone }} 的语气。");
  assert.equal("template_digest" in captured!.body.template, false);
});

test("Prompt Template local Session transport keeps product binding and accepts the authenticated owner", async () => {
  const localConfig = { ...config, authMode: "local_session_dev_test" as const };
  const envelope = versionEnvelope();
  envelope.version.owner_subject_ref = "user:usr_6f18e9a408a2ae7ec3f637b331eb5a98";
  let captured: RequestInit | undefined;
  globalThis.fetch = async (_url, init) => {
    captured = init;
    return jsonResponse(envelope);
  };

  const result = await readPromptTemplateVersion(localConfig, applicationId, templateId, 1);
  const headers = new Headers(captured?.headers);
  assert.equal(result.status, "versioned");
  assert.equal(captured?.credentials, "include");
  assert.equal(captured?.cache, "no-store");
  assert.equal(headers.get("X-RadishMind-Active-Tenant"), "tenant_demo");
  assert.equal(headers.get("X-RadishMind-Dev-Prompt-Template-Application"), applicationId);
  assert.equal(headers.has("X-RadishMind-Dev-Read-Identity"), false);
  assert.equal(headers.has("X-RadishMind-Dev-Read-Membership-Workspace"), false);
  assert.equal(headers.has("Authorization"), false);
});

test("Prompt Template validation preserves the persisted CAS lineage", () => {
  const current = {
    ...emptyOperation(),
    status: "saved" as const,
    currentDraftVersion: 3,
    currentTemplateVersion: 2,
    draft: {} as never,
    version: {} as never,
  };
  const validation = {
    ...emptyOperation(),
    status: "valid" as const,
    summary: "validated",
  };

  const merged = mergePromptTemplateValidationOperation(current, validation);
  assert.equal(merged.status, "valid");
  assert.equal(merged.summary, "validated");
  assert.equal(merged.currentDraftVersion, 3);
  assert.equal(merged.currentTemplateVersion, 2);
  assert.equal(merged.draft, current.draft);
  assert.equal(merged.version, current.version);
});

test("Prompt Template version source requires read_source and rejects response drift", async () => {
  let scopes = "";
  globalThis.fetch = async (_url, init) => {
    scopes = new Headers(init?.headers).get("X-RadishMind-Dev-Read-Scopes") ?? "";
    return jsonResponse(versionEnvelope());
  };
  const result = await readPromptTemplateVersion(config, applicationId, templateId, 1);
  assert.equal(result.status, "versioned");
  assert.equal(result.version?.templateDigest, `sha256:${"a".repeat(64)}`);
  assert.equal(result.version?.messages[1]?.content, "问题：{{ question }}");
  assert.equal(scopes, "prompt_application_templates:read_source");

  globalThis.fetch = async () => jsonResponse({ ...versionEnvelope(), raw_response: "forbidden" });
  const rejected = await readPromptTemplateVersion(config, applicationId, templateId, 1);
  assert.equal(rejected.status, "failed");
  assert.equal(rejected.version, null);
});

function emptyEnvelope() {
  return {
    request_id: "prompt-template-request",
    workspace_id: "workspace_demo",
    application_id: applicationId,
    draft: null,
    version: null,
    failure_code: null,
    current_draft_version: 0,
    current_template_version: 0,
    validation_summary: { state: "valid", is_valid: true, findings: [] },
    audit_ref: "audit-prompt-template-request",
  };
}

function emptyOperation() {
  return {
    status: "idle" as const,
    draft: null,
    version: null,
    validation: { state: "valid" as const, isValid: true, findings: [] },
    failureCode: "",
    currentDraftVersion: 0,
    currentTemplateVersion: 0,
    summary: "",
  };
}

function versionEnvelope() {
  return {
    ...emptyEnvelope(),
    current_draft_version: 2,
    current_template_version: 1,
    version: {
      schema_version: "prompt_application_template_version.v1",
      template_id: templateId,
      template_version: 1,
      source_draft_version: 2,
      tenant_ref: "tenant_demo",
      workspace_id: "workspace_demo",
      application_id: applicationId,
      owner_subject_ref: "subject_demo_user",
      template_name: "支持问题摘要",
      description: "按指定语气概括支持问题",
      messages: [
        { role: "system", content: "请使用 {{ tone }} 的语气。" },
        { role: "user", content: "问题：{{ question }}" },
      ],
      variables: [
        { name: "question", type: "string", required: true, description: "用户问题" },
        { name: "tone", type: "string", required: false, description: "回答语气", default_value: "清晰" },
      ],
      output_contract: { kind: "text", allow_empty: false, max_bytes: 4096 },
      template_digest: `sha256:${"a".repeat(64)}`,
      created_at: "2026-07-25T10:00:00Z",
      created_by_actor_ref: "subject_demo_user",
      request_id: "prompt-template-request",
      audit_ref: "audit-prompt-template-request",
    },
  };
}

function jsonResponse(body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
}
