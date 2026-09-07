import assert from "node:assert/strict";
import test from "node:test";

import {
  WorkflowTemplateRequestCoordinator,
  createWorkflowTemplateCandidate,
  decideWorkflowTemplateListing,
  deriveWorkflowTemplateDraft,
  listWorkflowTemplateCandidates,
  listWorkflowTemplateVersions,
  reviewWorkflowTemplateCandidate,
  type WorkflowTemplateCatalogConfig,
} from "../src/features/control-plane-read/workflowTemplateCatalogConsumer.ts";
import {
  saveWorkflowDraftDevRecord,
  type WorkflowSavedDraftConsumerConfig,
} from "../src/features/control-plane-read/savedWorkflowDraftConsumer.ts";

const digest = `sha256:${"a".repeat(64)}`;
const definitionDigest = `sha256:${"b".repeat(64)}`;
const applicationId = "app_aaaaaaaaaaaaaaaa";
const live: WorkflowTemplateCatalogConfig = {
  mode: "dev_workflow_template_http",
  baseUrl: "http://platform.test",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  subjectRef: "subject_demo_user",
};
const offline: WorkflowTemplateCatalogConfig = { ...live, mode: "offline" };

test("offline workflow template source performs zero HTTP requests", async () => {
  let requests = 0;
  globalThis.fetch = async () => { requests += 1; throw new Error("unexpected request"); };
  assert.deepEqual((await listWorkflowTemplateCandidates(offline)).records, []);
  assert.equal((await createWorkflowTemplateCandidate(offline, candidateInput())).failureCode, "workflow_template_offline");
  assert.equal((await deriveWorkflowTemplateDraft(offline, "template_support", {
    expectedPointerVersion: 1,
    templateVersion: 1,
    targetApplicationId: applicationId,
    draftId: "draft_template_support",
    name: "Support draft",
    confirmed: true,
  })).failureCode, "workflow_template_offline");
  assert.equal(requests, 0);
});

test("candidate list accepts exact records and rejects scope, duplicate, digest, cursor, unknown, and sensitive drift", async () => {
  globalThis.fetch = async () => json(candidatePage([candidateDocument()]));
  assert.equal((await listWorkflowTemplateCandidates(live)).records[0]?.candidateId, "candidate_support");

  const cases: unknown[] = [
    { ...candidatePage([]), workspace_id: "workspace_other" },
    candidatePage([candidateDocument(), candidateDocument()]),
    candidatePage([{ ...candidateDocument(), source_definition_digest: "sha256:bad" }]),
    { ...candidatePage([]), next_cursor: "not+a+base64url+cursor" },
    candidatePage([{ ...candidateDocument(), unexpected: true }]),
    { ...candidatePage([]), authorization: "forbidden" },
  ];
  for (const body of cases) {
    globalThis.fetch = async () => json(body);
    await assert.rejects(() => listWorkflowTemplateCandidates(live), /workflow template/u);
  }
});

test("candidate create and review send exact bodies, mutation-only URLs, and least scopes", async () => {
  const captured: Array<{ url: string; headers: Headers; body: unknown }> = [];
  globalThis.fetch = async (input, init) => {
    captured.push({ url: String(input), headers: new Headers(init?.headers), body: JSON.parse(String(init?.body)) });
    return json(operationEnvelope({ candidate: candidateDocument() }));
  };
  await createWorkflowTemplateCandidate(live, candidateInput());
  await reviewWorkflowTemplateCandidate(live, "candidate_support", {
    expectedReviewVersion: 0,
    decision: "approve",
    reason: "Reviewed portable metadata.",
  });
  assert.equal(captured[0]?.url, "http://platform.test/v1/user-workspace/workflow-template-candidates");
  assert.equal(captured[0]?.headers.get("X-RadishMind-Dev-Read-Scopes"), "workflow_definitions:read,workflow_definitions:write");
  assert.deepEqual(captured[0]?.body, {
    candidate_id: "candidate_support", template_id: "template_support", source_application_id: applicationId,
    source_definition_id: "definition_support", source_definition_version: 1, title: "Support workflow",
    summary: "Portable support workflow.", usage_notes: "Bind a reviewed provider profile.", labels: ["reviewed", "support"],
  });
  assert.equal(captured[1]?.url, "http://platform.test/v1/user-workspace/workflow-template-candidates/candidate_support/decisions");
  assert.equal(captured[1]?.headers.get("X-RadishMind-Dev-Read-Scopes"), "workflow_definitions:read,workflow_definitions:review");
  assert.deepEqual(captured[1]?.body, { expected_review_version: 0, decision: "approve", reason: "Reviewed portable metadata." });
});

test("listing CAS failure preserves current pointer authority without retry", async () => {
  let requests = 0;
  globalThis.fetch = async () => {
    requests += 1;
    return json(operationEnvelope({ failure_code: "workflow_template_pointer_version_conflict", current_pointer_version: 4 }));
  };
  const result = await decideWorkflowTemplateListing(live, "template_support", {
    expectedPointerVersion: 3,
    decision: "replace",
    version: 2,
    reason: "Replace with reviewed version.",
  });
  assert.equal(result.failureCode, "workflow_template_pointer_version_conflict");
  assert.equal(result.currentPointerVersion, 4);
  assert.equal(requests, 1);
});

test("version list binds template scope and rejects duplicate versions", async () => {
  globalThis.fetch = async (input) => {
    assert.equal(String(input), "http://platform.test/v1/user-workspace/workflow-templates/template_support/versions?workspace_id=workspace_demo&limit=25");
    return json(versionPage([versionDocument()]));
  };
  assert.equal((await listWorkflowTemplateVersions(live, "template_support", { limit: 25 })).records[0]?.templateDigest, digest);
  globalThis.fetch = async () => json(versionPage([versionDocument(), versionDocument()]));
  await assert.rejects(() => listWorkflowTemplateVersions(live, "template_support"), /duplicate/u);
});

test("derive validates exact v2 provenance and hands off server draft/version authority", async () => {
  let capturedBody: unknown;
  globalThis.fetch = async (_input, init) => {
    capturedBody = JSON.parse(String(init?.body));
    return json(operationEnvelope({ version: versionDocument(), lineage: listedLineage(), draft: derivedDraftDocument() }));
  };
  const result = await deriveWorkflowTemplateDraft(live, "template_support", {
    expectedPointerVersion: 1,
    templateVersion: 1,
    targetApplicationId: applicationId,
    draftId: "draft_template_support",
    name: "Support draft",
    confirmed: true,
  });
  assert.deepEqual(capturedBody, {
    expected_pointer_version: 1, template_version: 1, target_application_id: applicationId,
    draft_id: "draft_template_support", name: "Support draft", confirmed: true,
  });
  assert.equal(result.draftAuthority?.draftId, "draft_template_support");
  assert.equal(result.draftAuthority?.draftVersion, 1);
  assert.equal(result.draft?.derivation?.version, 2);
  assert.equal(result.draft?.derivation?.sourceKind, "workspace_workflow_template");
  assert.equal(result.draft?.localOnlyInteraction, "inspect_only");
  assert.deepEqual(result.draft?.requestedCapabilities, []);
});

test("template-derived draft save preserves only strict derivation_v2 metadata", async () => {
  globalThis.fetch = async () => json(operationEnvelope({ version: versionDocument(), lineage: listedLineage(), draft: derivedDraftDocument() }));
  const derived = await deriveWorkflowTemplateDraft(live, "template_support", {
    expectedPointerVersion: 1, templateVersion: 1, targetApplicationId: applicationId,
    draftId: "draft_template_support", name: "Support draft", confirmed: true,
  });
  assert.ok(derived.draft);

  let savedBody: Record<string, unknown> | undefined;
  globalThis.fetch = async (_input, init) => {
    savedBody = JSON.parse(String(init?.body)) as Record<string, unknown>;
    return json({
      request_id: "request_save", workspace_id: "workspace_demo", application_id: applicationId,
      draft: derivedDraftDocument(), failure_code: null, current_draft_version: 2,
      current_lifecycle_version: 1, current_lifecycle_state: "active",
      validation_summary: { validation_state: "valid_for_review", valid_for_review: true, findings: [] },
      blocked_capabilities: [], audit_ref: "audit_save",
    });
  };
  const savedConfig: WorkflowSavedDraftConsumerConfig = {
    mode: "dev_saved_draft_http", baseUrl: "http://platform.test", workspaceId: "workspace_demo",
    tenantRef: "tenant_demo", subjectRef: "subject_demo_user",
  };
  await saveWorkflowDraftDevRecord(derived.draft, savedConfig, 1, 1);
  const payload = (savedBody?.draft ?? {}) as Record<string, unknown>;
  const additionalFields = (payload.additional_fields ?? {}) as Record<string, unknown>;
  assert.deepEqual(payload.requested_capabilities, []);
  assert.equal(Object.hasOwn(additionalFields, "derivation_v1"), false);
  assert.deepEqual(additionalFields.derivation_v2, derivedDraftDocument().additional_fields.derivation_v2);
});

test("derive fails closed on scope, nested unknown, sensitive, and provenance drift", async () => {
  const base = derivedDraftDocument();
  const cases = [
    { ...base, workspace_id: "workspace_other" },
    { ...base, nodes: [{ ...(base.nodes[0] as object), unexpected: true }, ...base.nodes.slice(1)] },
    { ...base, raw_response: "forbidden" },
    { ...base, additional_fields: { derivation_v2: { ...(base.additional_fields.derivation_v2 as object), template_digest: definitionDigest } } },
  ];
  for (const draft of cases) {
    globalThis.fetch = async () => json(operationEnvelope({ version: versionDocument(), lineage: listedLineage(), draft }));
    await assert.rejects(() => deriveWorkflowTemplateDraft(live, "template_support", {
      expectedPointerVersion: 1, templateVersion: 1, targetApplicationId: applicationId,
      draftId: "draft_template_support", name: "Support draft", confirmed: true,
    }), /workflow template/u);
  }
});

test("scope coordinator aborts old requests and rejects late responses", () => {
  const coordinator = new WorkflowTemplateRequestCoordinator();
  const first = coordinator.reset("workspace_demo:app_a:subject_a");
  assert.equal(first.signal.aborted, false);
  const second = coordinator.reset("workspace_demo:app_b:subject_a");
  assert.equal(first.signal.aborted, true);
  assert.equal(coordinator.accepts(first), false);
  assert.equal(coordinator.accepts(second), true);
  coordinator.abort();
  assert.equal(coordinator.accepts(second), false);
});

function candidateInput() {
  return {
    candidateId: "candidate_support", templateId: "template_support", sourceApplicationId: applicationId,
    sourceDefinitionId: "definition_support", sourceDefinitionVersion: 1, title: "Support workflow",
    summary: "Portable support workflow.", usageNotes: "Bind a reviewed provider profile.", labels: ["reviewed", "support"],
  };
}

function portabilityDocument() {
  return { execution_profile: "workflow_definition_executor_v1", node_kinds: ["prompt", "llm", "output"], provider_refs: ["profile:mock"], risk_level: "low", portable: true, blockers: [] };
}

function candidateDocument() {
  return {
    schema_version: "workspace_workflow_template_candidate.v1", candidate_id: "candidate_support", template_id: "template_support",
    state: "pending", review_version: 0, source_application_id: applicationId, source_owner_subject_ref: "subject_demo_user",
    source_definition_id: "definition_support", source_definition_version: 1, source_definition_digest: definitionDigest,
    title: "Support workflow", summary: "Portable support workflow.", usage_notes: "Bind a reviewed provider profile.", labels: ["reviewed", "support"],
    portability: portabilityDocument(), decisions: [], created_at: "2026-08-29T09:00:00Z", updated_at: "2026-08-29T09:00:00Z",
    created_by_actor_ref: "subject_demo_user", updated_by_actor_ref: "subject_demo_user", request_id: "request_candidate", audit_ref: "audit_candidate",
  };
}

function versionDocument() {
  return {
    schema_version: "workspace_workflow_template_version.v1", template_id: "template_support", version: 1, template_digest: digest,
    candidate_id: "candidate_support", candidate_review_version: 1, source_application_id: applicationId, source_owner_subject_ref: "subject_demo_user",
    source_definition_id: "definition_support", source_definition_version: 1, source_definition_digest: definitionDigest,
    title: "Support workflow", summary: "Portable support workflow.", usage_notes: "Bind a reviewed provider profile.", labels: ["reviewed", "support"],
    portability: portabilityDocument(), created_at: "2026-08-29T09:01:00Z", created_by_actor_ref: "subject_demo_user",
    request_id: "request_version", audit_ref: "audit_version",
  };
}

function listedLineage() {
  return {
    schema_version: "workspace_workflow_template_lineage.v1", template_id: "template_support", tenant_ref: "tenant_demo", workspace_id: "workspace_demo",
    pointer_version: 1, lifecycle: "listed", listed_version: 1, listed_digest: digest,
    events: [{ schema_version: "workspace_workflow_template_listing_event.v1", event_id: "event_list_1", template_id: "template_support", decision: "list", reason: "List reviewed version.", before_pointer_version: 0, after_pointer_version: 1, before_listed_version: 0, after_listed_version: 1, actor_ref: "subject_demo_user", created_at: "2026-08-29T09:02:00Z", request_id: "request_list", audit_ref: "audit_list" }],
    created_at: "2026-08-29T09:01:00Z", updated_at: "2026-08-29T09:02:00Z", created_by_actor_ref: "subject_demo_user", updated_by_actor_ref: "subject_demo_user", request_id: "request_list", audit_ref: "audit_list",
  };
}

function derivedDraftDocument() {
  const node = (id: string, type: string, providerRef = "") => ({
    node_id: id, node_type: type, label: id, input_summary: "Bounded input.", output_summary: "Bounded output.",
    input_contract_ref: "contract_input", output_contract_ref: "contract_output", input_contract_fields: ["input_text"],
    output_contract_fields: ["answer"], output_mapping_summary: "Bounded mapping.", provider_ref: providerRef, tool_ref: "", rag_ref: "",
    risk_level: "low", requires_confirmation: false,
  });
  return {
    draft_id: "draft_template_support", workspace_id: "workspace_demo", application_id: applicationId,
    source_definition_id: "definition_support", base_definition_version: 1, schema_version: "saved_workflow_draft.v1", draft_status: "valid_for_review",
    name: "Support draft", description: "Portable support workflow.",
    nodes: [node("node_prompt", "prompt"), node("node_model", "llm", "profile:mock"), node("node_output", "output")],
    edges: [{ edge_id: "edge_prompt_model", from_node_id: "node_prompt", to_node_id: "node_model", condition_summary: "always" }, { edge_id: "edge_model_output", from_node_id: "node_model", to_node_id: "node_output", condition_summary: "always" }],
    input_contract: { contract_id: "contract_input", required_fields: ["input_text"], summary: "Bounded input." },
    output_contract: { contract_id: "contract_output", required_fields: ["answer"], summary: "Bounded output." },
    provider_refs: ["profile:mock"], tool_refs: [], rag_refs: [], requested_capabilities: [],
    additional_fields: { derivation_v2: { version: 2, source_kind: "workspace_workflow_template", template_id: "template_support", template_version: 1, template_digest: digest, source_definition_id: "definition_support", source_definition_version: 1, source_definition_digest: definitionDigest } },
    draft_version: 1, lifecycle_state: "active", lifecycle_version: 1, archived_at: null, library_updated_at: "2026-08-29T09:03:00Z",
    lifecycle_updated_by_actor_ref: "", provenance_kind: "workspace_template_derivation", created_at: "2026-08-29T09:03:00Z",
    updated_at: "2026-08-29T09:03:00Z", created_by_actor_ref: "subject_demo_user", updated_by_actor_ref: "subject_demo_user",
    validation_summary: { validation_state: "valid_for_review", valid_for_review: true, findings: [] }, blocked_capability_summary: [],
    request_audit_metadata: { request_id: "request_derive", audit_ref: "audit_derive", actor_ref: "subject_demo_user" }, sample_or_unsaved_draft_status: "",
  };
}

function operationEnvelope(overrides: Record<string, unknown> = {}) {
  return { request_id: "request_operation", workspace_id: "workspace_demo", candidate: null, version: null, lineage: null, draft: null, failure_code: null, current_review_version: 0, current_pointer_version: 0, audit_ref: "audit_operation", ...overrides };
}

function candidatePage(candidates: unknown[]) {
  return { request_id: "request_candidates", workspace_id: "workspace_demo", candidates, next_cursor: "", failure_code: null, audit_ref: "audit_candidates" };
}

function versionPage(versions: unknown[]) {
  return { request_id: "request_versions", workspace_id: "workspace_demo", template_id: "template_support", versions, next_cursor: "", failure_code: null, audit_ref: "audit_versions" };
}

function json(value: unknown): Promise<Response> {
  return Promise.resolve(new Response(JSON.stringify(value), { status: 200, headers: { "Content-Type": "application/json" } }));
}
