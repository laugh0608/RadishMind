import assert from "node:assert/strict";
import test from "node:test";

import {
  WorkflowDefinitionPromotionConflict,
  createWorkflowDefinitionCandidate,
  decideWorkflowDefinitionCandidate,
  deriveWorkflowDraftFromDefinitionVersion,
  listWorkflowDefinitionCandidates,
  startWorkflowDefinitionRun,
  type WorkflowDefinitionPromotionConfig,
  type WorkflowDefinitionVersion,
} from "../src/features/control-plane-read/workflowDefinitionPromotionConsumer.ts";

const live: WorkflowDefinitionPromotionConfig = { mode: "dev_workflow_definition_promotion_http", baseUrl: "http://platform.test", tenantRef: "tenant_demo", workspaceId: "workspace_demo", subjectRef: "subject_demo_user" };
const offline: WorkflowDefinitionPromotionConfig = { ...live, mode: "offline" };
const applicationId = "app_definition_demo";
const digest = `sha256:${"a".repeat(64)}`;

test("workflow definition promotion remains offline with zero requests", async () => {
  let requests = 0;
  globalThis.fetch = async () => { requests += 1; throw new Error("unexpected"); };
  assert.deepEqual(await listWorkflowDefinitionCandidates(offline, applicationId), []);
  await assert.rejects(() => createWorkflowDefinitionCandidate(offline, applicationId, { candidateId: "candidate_demo", definitionId: "definition_demo", draftId: "draft_demo", expectedDraftVersion: 1 }), /offline mode/u);
  assert.equal(requests, 0);
});

test("candidate create sends exact scope and strict authority fields", async () => {
  let captured: { headers: Headers; body: unknown } | undefined;
  globalThis.fetch = async (_input, init) => {
    captured = { headers: new Headers(init?.headers), body: JSON.parse(String(init?.body)) };
    return json(releaseEnvelope({ candidate: candidateDocument() }));
  };
  const candidate = await createWorkflowDefinitionCandidate(live, applicationId, { candidateId: "candidate_demo", definitionId: "definition_demo", draftId: "draft_demo", expectedDraftVersion: 3 });
  assert.equal(candidate.definitionDigest, digest);
  assert.equal(captured?.headers.get("X-RadishMind-Dev-Read-Scopes"), "workflow_definitions:write");
  assert.equal(captured?.headers.get("X-RadishMind-Dev-Workflow-Application"), applicationId);
  assert.deepEqual(captured?.body, { candidate_id: "candidate_demo", definition_id: "definition_demo", draft_id: "draft_demo", expected_draft_version: 3 });
});

test("strict candidate list rejects scope, schema, and sensitive drift", async () => {
  globalThis.fetch = async () => json({ request_id: "request_list", workspace_id: "workspace_demo", application_id: applicationId, candidates: [candidateDocument()], failure_code: null, audit_ref: "audit_list" });
  assert.equal((await listWorkflowDefinitionCandidates(live, applicationId))[0]?.candidateId, "candidate_demo");
  globalThis.fetch = async () => json({ request_id: "request_list", workspace_id: "workspace_other", application_id: applicationId, candidates: [], failure_code: null, audit_ref: "audit_list" });
  await assert.rejects(() => listWorkflowDefinitionCandidates(live, applicationId), /candidate list failed/u);
  globalThis.fetch = async () => json({ request_id: "request_list", workspace_id: "workspace_demo", application_id: applicationId, candidates: [{ ...candidateDocument(), token: "forbidden" }], failure_code: null, audit_ref: "audit_list" });
  await assert.rejects(() => listWorkflowDefinitionCandidates(live, applicationId), /candidate list failed/u);
});

test("review CAS conflict remains explicit without fallback", async () => {
  globalThis.fetch = async () => json(releaseEnvelope({ failure_code: "workflow_definition_version_conflict", current_review_version: 2 }));
  await assert.rejects(
    () => decideWorkflowDefinitionCandidate(live, applicationId, "candidate_demo", { expectedReviewVersion: 0, decision: "approve", reason: "Reviewed." }),
    (error: unknown) => error instanceof WorkflowDefinitionPromotionConflict && error.currentReviewVersion === 2,
  );
});

test("definition run maps strict v5 evidence and transient advisory output", async () => {
  globalThis.fetch = async () => json({ request_id: "request_run", workspace_id: "workspace_demo", application_id: applicationId, run: runV5(), advisory_output: "Transient advisory output.", failure_code: null, failure_summary: "", audit_ref: "audit_run" });
  const result = await startWorkflowDefinitionRun(live, applicationId, { definitionId: "definition_demo", expectedPointerVersion: 1, expectedDefinitionVersion: 1, expectedDefinitionDigest: digest, inputText: "Bounded one-time input.", conditionValues: {}, model: "" });
  assert.equal(result.record.schemaVersion, "workflow_run_record.v5");
  assert.equal(result.record.definitionAuthority?.activationPointerVersion, 1);
  assert.equal(result.record.output, "");
  assert.equal(result.advisoryOutput, "Transient advisory output.");
  globalThis.fetch = async () => json({ request_id: "request_run", workspace_id: "workspace_demo", application_id: applicationId, run: runV5(), advisory_output: "ok", raw_response: "forbidden", failure_code: null, failure_summary: "", audit_ref: "audit_run" });
  await assert.rejects(() => startWorkflowDefinitionRun(live, applicationId, { definitionId: "definition_demo", expectedPointerVersion: 1, expectedDefinitionVersion: 1, expectedDefinitionDigest: digest, inputText: "Bounded.", conditionValues: {}, model: "" }), /invalid or sensitive/u);
});

test("derived draft preserves exact immutable base version provenance", () => {
  const document = versionDocument();
  const version: WorkflowDefinitionVersion = {
    definitionId: document.definition_id, version: document.version, definitionDigest: document.definition_digest,
    candidateId: document.candidate_id, candidateReviewVersion: document.candidate_review_version,
    sourceDraftId: document.source_draft_id, sourceDraftVersion: document.source_draft_version,
    sourceDraftDigest: document.source_draft_digest, snapshot: {
      schemaVersion: "saved_workflow_draft.v1", name: "Definition demo", description: "Reviewed definition.",
      nodes: snapshotDocument().nodes.map((node) => ({ nodeId: String(node.node_id), nodeType: node.node_type as "prompt" | "llm" | "output", label: String(node.label), inputSummary: String(node.input_summary), outputSummary: String(node.output_summary), inputContractRef: String(node.input_contract_ref), outputContractRef: String(node.output_contract_ref), inputContractFields: node.input_contract_fields as string[], outputContractFields: node.output_contract_fields as string[], outputMappingSummary: String(node.output_mapping_summary), providerRef: String(node.provider_ref), toolRef: "", ragRef: "", riskLevel: "low", requiresConfirmation: false })),
      edges: snapshotDocument().edges.map((edge) => ({ edgeId: String(edge.edge_id), fromNodeId: String(edge.from_node_id), toNodeId: String(edge.to_node_id), conditionSummary: String(edge.condition_summary) })),
      inputContract: { contractId: "contract_input", requiredFields: ["input_text"], summary: "input" }, outputContract: { contractId: "contract_output", requiredFields: ["answer"], summary: "output" }, providerRefs: ["provider:mock"], toolRefs: [], ragRefs: [], requestedCapabilities: [], executionProfile: "workflow_definition_executor_v1",
    }, activationEligible: true, eligibilityBlockers: [], createdAt: document.created_at, createdByActorRef: document.created_by_actor_ref, requestId: document.request_id, auditRef: document.audit_ref,
  };
  const draft = deriveWorkflowDraftFromDefinitionVersion(version, applicationId, 2);
  assert.equal(draft.workflowDefinitionId, "definition_demo");
  assert.equal(draft.baseDefinitionVersion, 1);
  assert.equal(draft.localOnlyInteraction, "local_edit");
});

function releaseEnvelope(overrides: Record<string, unknown> = {}) { return { request_id: "request_release", workspace_id: "workspace_demo", application_id: applicationId, candidate: null, version: null, activation: null, failure_code: null, current_review_version: 0, current_pointer_version: 0, audit_ref: "audit_release", ...overrides }; }
function candidateDocument() { return { schema_version: "workflow_definition_release_candidate.v1", candidate_id: "candidate_demo", definition_id: "definition_demo", source_draft_id: "draft_demo", source_draft_version: 3, source_draft_digest: digest, definition_digest: digest, snapshot: snapshotDocument(), activation_eligible: true, eligibility_blockers: [], state: "pending", review_version: 0, reviews: [], created_at: "2026-07-19T10:00:00Z", updated_at: "2026-07-19T10:00:00Z", created_by_actor_ref: "subject_demo_user", updated_by_actor_ref: "subject_demo_user", request_id: "request_candidate", audit_ref: "audit_candidate" }; }
function versionDocument() { return { schema_version: "workflow_definition_version.v1" as const, definition_id: "definition_demo", version: 1, definition_digest: digest, candidate_id: "candidate_demo", candidate_review_version: 1, source_draft_id: "draft_demo", source_draft_version: 3, source_draft_digest: digest, snapshot: snapshotDocument(), activation_eligible: true, eligibility_blockers: [], created_at: "2026-07-19T10:01:00Z", created_by_actor_ref: "subject_demo_user", request_id: "request_version", audit_ref: "audit_version" }; }
function snapshotDocument() {
  const node = (id: string, type: string) => ({ node_id: id, node_type: type, label: id, input_summary: "bounded metadata", output_summary: "bounded metadata", input_contract_ref: "contract_input", output_contract_ref: "contract_output", input_contract_fields: ["input_text"], output_contract_fields: ["answer"], output_mapping_summary: "bounded mapping", provider_ref: type === "llm" ? "provider:mock" : "", tool_ref: "", rag_ref: "", risk_level: "low", requires_confirmation: false });
  return { schema_version: "saved_workflow_draft.v1", name: "Definition demo", description: "Reviewed definition.", nodes: [node("node_prompt", "prompt"), node("node_model", "llm"), node("node_output", "output")], edges: [{ edge_id: "edge_prompt_model", from_node_id: "node_prompt", to_node_id: "node_model", condition_summary: "always" }, { edge_id: "edge_model_output", from_node_id: "node_model", to_node_id: "node_output", condition_summary: "always" }], input_contract: { contract_id: "contract_input", required_fields: ["input_text"], summary: "input" }, output_contract: { contract_id: "contract_output", required_fields: ["answer"], summary: "output" }, provider_refs: ["provider:mock"], tool_refs: [], rag_refs: [], requested_capabilities: [], execution_profile: "workflow_definition_executor_v1" };
}
function runV5() { const node = (id: string, type: string) => ({ node_id: id, node_type: type, label: id, status: "succeeded", started_at: "2026-07-19T10:02:00Z", completed_at: "2026-07-19T10:02:01Z", duration_ms: 10, predecessor_node_ids: [], provider_ref: type === "llm" ? "provider:mock" : "", output_preview: "", failure_code: "" }); return { schema_version: "workflow_run_record.v5", record_version: 2, run_id: "run_definition_demo", draft_id: "", draft_version: 0, workspace_id: "workspace_demo", application_id: applicationId, execution_kind: "workflow_definition_execution", execution_source_kind: "workflow_definition", execution_source_id: "definition_demo", execution_source_version: 1, execution_profile: "workflow_definition_executor_v1", input_digest: digest, definition_authority: { definition_id: "definition_demo", definition_version: 1, definition_digest: digest, activation_pointer_version: 1, candidate_id: "candidate_demo", candidate_review_version: 1, source_draft_id: "draft_demo", source_draft_version: 3, source_draft_digest: digest, application_record_version: 1, application_lifecycle: "active" }, status: "succeeded", failure_code: "", failure_summary: "", started_at: "2026-07-19T10:02:00Z", completed_at: "2026-07-19T10:02:01Z", input_bytes: 23, condition_node_ids: [], requested_model: "", selected_provider: "mock", selected_profile: "", selected_model: "mock", upstream_model: "mock", selection_source: "mock", nodes: [node("node_prompt", "prompt"), node("node_model", "llm"), node("node_output", "output")], output: "", request_id: "request_run", audit_ref: "audit_run", actor_ref: "subject_demo_user", side_effects: { retrieval_calls: 0, provider_calls: 1, tool_calls: 0, confirmation_calls: 0, business_writes: 0, replay_writes: 0 }, diagnostic: { failure_boundary: "", failure_stage: "", failed_node_id: "", last_completed_node_id: "node_output", terminal_write_state: "stored", gateway_failure_category: "none", tool_failure_category: "none", retrieval_failure_category: "none", summary: "", recommended_review_action: "", observed_at: "2026-07-19T10:02:01Z" } }; }
function json(value: unknown): Promise<Response> { return Promise.resolve(new Response(JSON.stringify(value), { status: 200, headers: { "Content-Type": "application/json" } })); }
