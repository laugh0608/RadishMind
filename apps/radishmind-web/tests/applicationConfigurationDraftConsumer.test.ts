import assert from "node:assert/strict";
import test from "node:test";

import {
  compareApplicationConfigurationDraft,
  createApplicationConfigurationDraft,
  initialApplicationConfigurationDraftListState,
  initialApplicationConfigurationDraftState,
  listApplicationConfigurationDrafts,
  readApplicationConfigurationDraft,
  saveApplicationConfigurationDraft,
  validateApplicationConfigurationDraft,
  validateApplicationConfigurationDraftRemote,
  type ApplicationConfigurationDraftConfig,
} from "../src/features/control-plane-read/applicationConfigurationDraftConsumer.ts";
import { createApplicationApiIntegrationDraftHandoffDetail } from "../src/features/control-plane-read/applicationApiIntegrationEvents.ts";

const offline: ApplicationConfigurationDraftConfig = {
  mode: "offline", baseUrl: "http://127.0.0.1:7000", tenantRef: "tenant_demo", workspaceId: "workspace_demo", subjectRef: "subject_demo_user",
};
const live: ApplicationConfigurationDraftConfig = { ...offline, mode: "dev_application_draft_http" };
const baseline = { applicationId: "app_docs_assistant", displayName: "Radish Docs Assistant", applicationKind: "docs_qa", updatedAt: "2026-05-31T09:40:00Z" };
const models = [{ id: "profile:local-dev", ownedBy: "radishmind", protocols: ["responses", "messages"] as const }];

test("Application configuration draft offline mode sends zero requests", async () => {
  const originalFetch = globalThis.fetch;
  let calls = 0;
  globalThis.fetch = async () => { calls += 1; throw new Error("unexpected fetch"); };
  try {
    const draft = createApplicationConfigurationDraft(offline, baseline);
    draft.defaultModel = "profile:local-dev";
    assert.equal(initialApplicationConfigurationDraftState(offline).status, "offline");
    assert.equal(initialApplicationConfigurationDraftListState(offline).status, "offline");
    assert.equal((await validateApplicationConfigurationDraftRemote(offline, draft)).status, "offline");
    assert.equal((await saveApplicationConfigurationDraft(offline, draft, 0)).status, "offline");
    assert.equal((await listApplicationConfigurationDrafts(offline, baseline.applicationId)).status, "offline");
    assert.equal((await readApplicationConfigurationDraft(offline, baseline.applicationId, draft.draftId)).draft, null);
    assert.equal(calls, 0);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Application configuration validation covers model compatibility and secret rejection", () => {
  const draft = createApplicationConfigurationDraft(live, baseline);
  draft.defaultModel = "profile:local-dev";
  draft.defaultProtocol = "responses";
  assert.equal(validateApplicationConfigurationDraft(draft, models.map((model) => ({ ...model, protocols: [...model.protocols] }))).isValid, true);

  draft.defaultProtocol = "chat_completions";
  let validation = validateApplicationConfigurationDraft(draft, models.map((model) => ({ ...model, protocols: [...model.protocols] })));
  assert.equal(validation.findings.some((finding) => finding.code === "application_draft_protocol_incompatible"), true);

  draft.defaultProtocol = "responses";
  draft.description = "Authorization: Bearer secret-value";
  validation = validateApplicationConfigurationDraft(draft, models.map((model) => ({ ...model, protocols: [...model.protocols] })));
  assert.equal(validation.findings.some((finding) => finding.code === "application_draft_secret_material_forbidden"), true);
  assert.equal(JSON.stringify(validation).includes("secret-value"), false);
});

test("Application draft save carries application scope and maps version conflict", async () => {
  const originalFetch = globalThis.fetch;
  const draft = createApplicationConfigurationDraft(live, baseline);
  draft.defaultModel = "profile:local-dev";
  try {
    globalThis.fetch = async (url, init) => {
      assert.equal(String(url), "http://127.0.0.1:7000/v1/user-workspace/application-drafts");
      const headers = new Headers(init?.headers);
      assert.equal(headers.get("X-RadishMind-Dev-Application-Draft-Workspace"), "workspace_demo");
      assert.equal(headers.get("X-RadishMind-Dev-Application-Draft-Application"), "app_docs_assistant");
      assert.equal(headers.get("X-RadishMind-Dev-Read-Scopes"), "application_drafts:read,application_drafts:write");
      const body = JSON.parse(String(init?.body));
      assert.equal(body.expected_draft_version, 1);
      assert.equal(body.draft.application_id, "app_docs_assistant");
      return jsonResponse(draftEnvelope({ failureCode: "application_draft_version_conflict", version: 2 }));
    };
    const state = await saveApplicationConfigurationDraft(live, draft, 1);
    assert.equal(state.status, "version_conflict");
    assert.equal(state.currentDraftVersion, 2);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Application draft list and restore reject response scope drift", async () => {
  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async () => jsonResponse({
      request_id: "list-request", workspace_id: "workspace_demo", application_id: "app_flow_copilot",
      draft_summaries: [], failure_code: null, audit_ref: "audit-list",
    });
    const list = await listApplicationConfigurationDrafts(live, "app_docs_assistant");
    assert.equal(list.status, "failed");
    assert.equal(list.summaries.length, 0);

    globalThis.fetch = async () => jsonResponse(draftEnvelope({ applicationId: "app_flow_copilot" }));
    const restored = await readApplicationConfigurationDraft(live, "app_docs_assistant", "app-config-app-docs-assistant");
    assert.equal(restored.draft, null);
    assert.equal(restored.state.status, "store_failure");

    globalThis.fetch = async () => jsonResponse({ ...draftEnvelope(), metadata: { api_key: "sk-private" } });
    const forbidden = await readApplicationConfigurationDraft(live, "app_docs_assistant", "app-config-app-docs-assistant");
    assert.equal(forbidden.draft, null);
    assert.equal(JSON.stringify(forbidden).includes("sk-private"), false);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Application draft v2 restores and saves an exact ref-only RAG binding", async () => {
  const originalFetch = globalThis.fetch;
  const bindingDigest = `sha256:${"b".repeat(64)}`;
  const draftDigest = `sha256:${"d".repeat(64)}`;
  let calls = 0;
  try {
    globalThis.fetch = async (_input, init) => {
      calls += 1;
      if (calls === 1) return jsonResponse({
        ...draftEnvelope({ version: 2 }),
        draft: {
          draft_id: "app-config-app-docs-assistant", workspace_id: "workspace_demo", application_id: "app_docs_assistant",
          base_application_updated_at: baseline.updatedAt, schema_version: "application_configuration_draft.v2", display_name: baseline.displayName,
          description: "Bound draft", application_kind: baseline.applicationKind, default_protocol: "responses", default_model: "profile:local-dev",
          allowed_protocols: ["responses"], workflow_rag_binding_ref: { binding_id: "wragb_aaaaaaaaaaaaaaaa", binding_version: 1, binding_digest: bindingDigest },
          draft_version: 2, draft_digest: draftDigest, validation_summary: { state: "valid", is_valid: true, findings: [] },
          created_at: baseline.updatedAt, updated_at: baseline.updatedAt, created_by_actor_ref: "subject_demo_user", updated_by_actor_ref: "subject_demo_user",
          request_id: "app-draft-read", audit_ref: "audit-app-draft-read",
        },
      });
      const headers = new Headers(init?.headers);
      assert.equal(headers.get("X-RadishMind-Dev-Read-Scopes"), "application_drafts:read,application_drafts:write,workflow_rag_promotions:bind");
      const body = JSON.parse(String(init?.body));
      assert.deepEqual(body.draft.workflow_rag_binding_ref, { binding_id: "wragb_aaaaaaaaaaaaaaaa", binding_version: 1, binding_digest: bindingDigest });
      return jsonResponse(draftEnvelope({ version: 3 }));
    };
    const restored = await readApplicationConfigurationDraft(live, baseline.applicationId, "app-config-app-docs-assistant");
    assert.equal(restored.draft?.schemaVersion, "application_configuration_draft.v2");
    assert.equal(restored.draft?.draftDigest, draftDigest);
    assert.equal(restored.draft?.workflowRAGBindingRef?.bindingId, "wragb_aaaaaaaaaaaaaaaa");
    const saved = await saveApplicationConfigurationDraft(live, restored.draft!, 2);
    assert.equal(saved.status, "saved");
  } finally { globalThis.fetch = originalFetch; }
});

test("Application switching creates isolated state and handoff keeps exact public fields", () => {
  const first = createApplicationConfigurationDraft(live, baseline);
  first.description = "Unsaved first application state";
  first.defaultModel = "profile:local-dev";
  const second = createApplicationConfigurationDraft(live, { applicationId: "app_flow_copilot", displayName: "RadishFlow Copilot", applicationKind: "workflow_copilot", updatedAt: "2026-05-31T10:20:00Z" });
  assert.equal(second.applicationId, "app_flow_copilot");
  assert.equal(second.description, "");
  assert.equal(second.defaultModel, "");
  assert.equal(JSON.stringify(second).includes(first.description), false);

  const differences = compareApplicationConfigurationDraft(baseline, first);
  assert.equal(differences.some((item) => item.field === "default_model" && item.changed), true);
  assert.deepEqual(createApplicationApiIntegrationDraftHandoffDetail(" app_docs_assistant ", "responses", " profile:local-dev "), {
    applicationId: "app_docs_assistant", protocol: "responses", model: "profile:local-dev",
  });
});

function draftEnvelope(options: { failureCode?: string | null; version?: number; applicationId?: string } = {}) {
  const applicationId = options.applicationId ?? "app_docs_assistant";
  return {
    request_id: "app-draft-request", workspace_id: "workspace_demo", application_id: applicationId,
    draft: null, failure_code: options.failureCode ?? null, current_draft_version: options.version ?? 0,
    validation_summary: { state: "valid", is_valid: true, findings: [] }, audit_ref: "audit-app-draft-request",
  };
}

function jsonResponse(body: unknown): Response {
  return new Response(JSON.stringify(body), { status: 200, headers: { "Content-Type": "application/json" } });
}
