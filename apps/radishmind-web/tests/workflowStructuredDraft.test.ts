import assert from "node:assert/strict";
import test from "node:test";
import { buildWorkflowStructuredInputDraft } from "../src/features/control-plane-read/workflowDraftCreation.ts";
import { validateWorkflowDraftInputContract, workflowDraftInputContractDocument } from "../src/features/control-plane-read/workflowDraftInputContract.ts";
import { isSavedWorkflowDraftV2Content } from "../src/features/control-plane-read/workflowSavedDraftV2Validation.ts";
import { buildDerivedWorkflowDraft, cloneWorkflowDraftForEditing } from "../src/features/control-plane-read/workflowSavedDraftDerivation.ts";
import { compareWorkflowSavedDraftRevision } from "../src/features/control-plane-read/workflowSavedDraftRevisionComparison.ts";
import { saveWorkflowDraftDevRecord, workflowDraftFromSavedWorkflowDraftDocument, type SavedWorkflowDraftDocument, type WorkflowSavedDraftConsumerConfig } from "../src/features/control-plane-read/savedWorkflowDraftConsumer.ts";
import { readWorkflowSavedDraftRevision, restoreWorkflowSavedDraftRevision } from "../src/features/control-plane-read/workflowSavedDraftRevisionConsumer.ts";
import { evaluateWorkflowExecutorEligibility } from "../src/features/control-plane-read/workflowExecutorConsumer.ts";
import { evaluateWorkflowDefinitionCandidateCompatibility, deriveWorkflowDraftFromDefinitionVersion, type WorkflowDefinitionVersion } from "../src/features/control-plane-read/workflowDefinitionPromotionConsumer.ts";
import type { WorkflowDraftDesignerDraft } from "../src/features/control-plane-read/workflowDraftDesigner.ts";

const config: WorkflowSavedDraftConsumerConfig = { mode: "dev_saved_draft_http", baseUrl: "http://platform.test", workspaceId: "workspace_demo", tenantRef: "tenant_demo", subjectRef: "subject_test" };
const digest = `sha256:${"a".repeat(64)}`;
const time = "2026-10-21T01:00:00Z";
function draft() {
  const source: WorkflowDraftDesignerDraft = { draftId: "draft_source", templateRef: "wf_source", label: "Source", applicationRef: "app_test", workflowDefinitionId: "wf_source", providerProfileRef: "provider:mock", summary: "Source", nodes: [], edges: [], designerLayout: { source: "workflow_node_designer", persistence: "ui_only", nodePositions: [] }, readiness: [], risks: [], blockedCapabilities: [], routeMetadata: { sourceRouteId: "workflow-definition-summary-list-route", draftRouteId: "workflow-draft-designer-offline-draft", routePath: "/v1/user-workspace/workflow-definitions", requestId: "request", auditRef: "audit" }, localOnlyInteraction: "local_edit" };
  const value = buildWorkflowStructuredInputDraft(source, 1, "app_test");
  assert.equal(value.schemaVersion, "saved_workflow_draft.v2");
  if (value.schemaVersion !== "saved_workflow_draft.v2") throw new Error("v2 required");
  value.inputContract = { contractId: "diagnostics_v1", summary: "Ordered metadata", fields: [
    { name: "context", valueType: "string", required: true, label: "Context", description: "Bounded context" },
    { name: "attempts", valueType: "integer", required: false, label: "Attempts", description: "Optional attempts" },
    { name: "threshold", valueType: "number", required: true, label: "Threshold", description: "Finite number" },
    { name: "confirmed", valueType: "boolean", required: true, label: "Confirmed", description: "Explicit boolean" },
  ] };
  return value;
}
function document(value = draft()): SavedWorkflowDraftDocument {
  return { draft_id: value.draftId, workspace_id: config.workspaceId, application_id: value.applicationRef, source_definition_id: value.workflowDefinitionId, base_definition_version: 0, schema_version: value.schemaVersion, draft_status: "valid_for_review", name: value.label, description: value.summary,
    nodes: value.nodes.map(node => ({ node_id: node.nodeId, node_type: node.nodeType, label: node.label, input_summary: node.inputSummary, output_summary: node.outputSummary, input_contract_ref: value.inputContract.contractId, output_contract_ref: "output", input_contract_fields: node.inputContractFields, output_contract_fields: node.outputContractFields, output_mapping_summary: node.outputMappingSummary, provider_ref: node.providerRef, tool_ref: "", rag_ref: "", risk_level: "low", requires_confirmation: false })),
    edges: value.edges.map(edge => ({ edge_id: edge.edgeId, from_node_id: edge.fromNodeId, to_node_id: edge.toNodeId, condition_summary: edge.conditionSummary })),
    input_contract: { ...workflowDraftInputContractDocument(value.inputContract), contract_digest: digest }, output_contract: { contract_id: "output", required_fields: ["answer_summary"], summary: "Output" }, provider_refs: ["provider:mock"], tool_refs: [], rag_refs: [], requested_capabilities: [], draft_version: 2, lifecycle_state: "active", lifecycle_version: 1, archived_at: null, library_updated_at: time, lifecycle_updated_by_actor_ref: "subject_test", provenance_kind: "workflow_definition", created_at: time, updated_at: time, created_by_actor_ref: "subject_test", updated_by_actor_ref: "subject_test", validation_summary: { validation_state: "valid_for_review", valid_for_review: true, findings: [] }, blocked_capability_summary: [], request_audit_metadata: { request_id: "request", audit_ref: "audit", actor_ref: "subject_test" }, sample_or_unsaved_draft_status: "saved_dev_record" };
}
function envelope(saved: SavedWorkflowDraftDocument) { return { request_id: "request", workspace_id: config.workspaceId, application_id: saved.application_id, draft: saved, failure_code: null, current_draft_version: saved.draft_version, current_lifecycle_version: 1, current_lifecycle_state: "active", validation_summary: saved.validation_summary, blocked_capabilities: [], audit_ref: "audit" }; }
function json(value: unknown) { return new Response(JSON.stringify(value), { status: 200 }); }

test("v2 save and exact read preserve ordered field metadata without sending a fabricated digest", async (t) => {
  const value = draft();
  value.inputContract.contractDigest = digest;
  const saved = document(value);
  const original = globalThis.fetch; t.after(() => { globalThis.fetch = original; });
  globalThis.fetch = async (_, init) => {
    const body = JSON.parse(String(init?.body));
    assert.equal(body.draft.schema_version, "saved_workflow_draft.v2");
    assert.deepEqual(body.draft.input_contract, workflowDraftInputContractDocument(value.inputContract));
    assert.equal("contract_digest" in body.draft.input_contract, false);
    assert.equal("required_fields" in body.draft.input_contract, false);
    assert.equal(body.draft.additional_fields.executor_v0, undefined);
    return json(envelope(saved));
  };
  const state = await saveWorkflowDraftDevRecord(value, config, 1, 1);
  assert.equal(state.status, "saved_dev_record");
  const read = workflowDraftFromSavedWorkflowDraftDocument(saved);
  assert.deepEqual(read.inputContract, value.inputContract);
  assert.equal(read.schemaVersion, value.schemaVersion);
  assert.equal(read.executionProfile, "executor_v2");
  assert.equal(evaluateWorkflowDefinitionCandidateCompatibility(read).executionProfile, "workflow_definition_executor_v2");
  assert.equal(evaluateWorkflowExecutorEligibility(read, state, false).eligible, false);
});

test("v2 contract validation catches distinct metadata errors and preserves original input", () => {
  const contract = draft().inputContract;
  assert.deepEqual(validateWorkflowDraftInputContract(contract), []);
  const invalid = { ...contract, contractId: "bad id", summary: "x".repeat(4001), fields: [
    { ...contract.fields[0]!, name: "Bad Name", label: "", description: "x".repeat(4001) },
    { ...contract.fields[1]!, name: "Bad Name" },
  ] };
  assert.deepEqual(new Set(validateWorkflowDraftInputContract(invalid).map(issue => issue.code)), new Set(["identity", "summary", "name", "label", "description", "duplicate"]));
  assert.throws(() => workflowDraftInputContractDocument(invalid), /contract_invalid/);
  assert.equal(invalid.fields[0]!.name, "Bad Name");
  assert.equal(validateWorkflowDraftInputContract({ ...contract, fields: [] })[0]?.code, "count");
  assert.equal(validateWorkflowDraftInputContract({ ...contract, fields: Array.from({ length: 17 }, (_, index) => ({ ...contract.fields[0]!, name: `field_${index}` })) })[0]?.code, "count");
  assert.equal(validateWorkflowDraftInputContract({ ...contract, contractId: "api_key", summary: "password=synthetic", fields: [{ ...contract.fields[0]!, description: "bearer synthetic", label: "密".repeat(54) }] }).filter(issue => issue.code === "sensitive").length, 3);
  assert.equal(validateWorkflowDraftInputContract({ ...contract, fields: [{ ...contract.fields[0]!, valueType: "object", required: "true" }] } as unknown as typeof contract).length, 2);
});

test("v2 decoding rejects unknown fields, mixed schemas, broken metadata and unsupported node types", () => {
  const valid = document();
  assert.equal(isSavedWorkflowDraftV2Content(valid), true);
  const mutations = [
    { ...valid, extra: true }, { ...valid, schema_version: "saved_workflow_draft.v3" },
    { ...valid, input_contract: { ...valid.input_contract, required_fields: [] } },
    { ...valid, input_contract: { ...valid.input_contract, contract_digest: "invalid" } },
    { ...valid, output_contract: { ...valid.output_contract, fields: [] } },
    { ...valid, nodes: [{ ...valid.nodes[0], node_type: "http_tool" }] },
    { ...valid, nodes: [{ ...valid.nodes[0], extra: true }] },
    { ...valid, edges: [{ ...valid.edges[0], extra: true }] },
    { ...valid, provider_refs: [4] }, { ...valid, additional_fields: [] },
  ];
  for (const broken of mutations) {
    assert.equal(isSavedWorkflowDraftV2Content(broken), false);
    assert.throws(() => workflowDraftFromSavedWorkflowDraftDocument(broken as SavedWorkflowDraftDocument));
  }
  assert.throws(() => workflowDraftFromSavedWorkflowDraftDocument({ ...valid, schema_version: "saved_workflow_draft.v1" }));
});

test("v2 clone, derivation and revision comparison keep independent ordered contracts", () => {
  const original = workflowDraftFromSavedWorkflowDraftDocument(document());
  const cloned = cloneWorkflowDraftForEditing(original);
  const derived = buildDerivedWorkflowDraft(original, 2, []);
  assert.deepEqual(derived.inputContract, original.inputContract);
  assert.notEqual(derived.inputContract, original.inputContract);
  cloned.inputContract!.fields.reverse();
  cloned.inputContract!.fields[0]!.label = "Changed";
  assert.notDeepEqual(cloned.inputContract, original.inputContract);
  assert.equal(original.inputContract!.fields[0]!.name, "context");
  assert.equal(compareWorkflowSavedDraftRevision(original, cloned).changes[0]?.subject, "input contract");
  const version = { definitionId: "definition_test", version: 2, definitionDigest: digest, requestId: "request", auditRef: "audit", snapshot: { schemaVersion: "saved_workflow_draft.v2", name: original.label, providerRefs: ["provider:mock"], nodes: original.nodes, edges: original.edges, inputContract: original.inputContract } } as unknown as WorkflowDefinitionVersion;
  const fromVersion = deriveWorkflowDraftFromDefinitionVersion(version, "app_test", 1);
  assert.equal(fromVersion.baseDefinitionVersion, 2);
  assert.deepEqual(fromVersion.inputContract, original.inputContract);
  assert.notEqual(fromVersion.inputContract!.fields[0], original.inputContract!.fields[0]);
});

test("v2 historical read and restore preserve contracts and reject mismatched exact revisions", async (t) => {
  const value = draft(); const saved = document(value);
  const original = globalThis.fetch; t.after(() => { globalThis.fetch = original; });
  globalThis.fetch = async () => json({ ...envelope(saved), revision: { schema_version: "saved_workflow_draft_revision.v1", draft: saved, revision_kind: "saved", restored_from_version: 0 } });
  const read = await readWorkflowSavedDraftRevision(value, 2, config);
  assert.deepEqual(read.draft.inputContract, workflowDraftFromSavedWorkflowDraftDocument(saved).inputContract);
  const restored = await restoreWorkflowSavedDraftRevision(value, 1, 1, 1, config);
  assert.deepEqual(restored.draft?.inputContract, read.draft.inputContract);
  await assert.rejects(() => readWorkflowSavedDraftRevision(value, 1, config), /scope_invalid/);
  globalThis.fetch = async () => json({ ...envelope(saved), draft: { ...saved, application_id: "another_app" } });
  await assert.rejects(() => restoreWorkflowSavedDraftRevision(value, 1, 1, 1, config), /scope_invalid/);
});
