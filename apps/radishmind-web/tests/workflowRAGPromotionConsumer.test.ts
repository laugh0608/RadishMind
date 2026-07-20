import assert from "node:assert/strict";
import test from "node:test";

import {
  createWorkflowRAGPromotionCandidate,
  decideWorkflowRAGPromotionCandidate,
  listWorkflowRAGPromotionCandidates,
  readWorkflowRAGPromotionCandidate,
  workflowRAGPromotionDecisionAllowed,
  type WorkflowRAGPromotionConfig,
} from "../src/features/control-plane-read/workflowRAGPromotionConsumer.ts";

const allScopes = new Set([
  "workflow_rag_promotions:read", "workflow_rag_promotions:write", "workflow_rag_promotions:review",
  "workflow_rag_evaluation_datasets:read", "workflow_rag_snapshots:read", "application_drafts:read",
] as const);
const live: WorkflowRAGPromotionConfig = { mode: "dev_workflow_rag_promotion_http", baseUrl: "http://platform.test", tenantRef: "tenant_demo", workspaceId: "workspace_demo", subjectRef: "subject_demo_user", authMode: "dev_headers", scopes: allScopes };

test("knowledge promotion offline and scope boundaries send zero requests", async () => {
  const originalFetch = globalThis.fetch;
  let calls = 0;
  globalThis.fetch = async () => { calls += 1; throw new Error("unexpected request"); };
  try {
    const offline = { ...live, mode: "offline" as const };
    assert.equal((await listWorkflowRAGPromotionCandidates(offline, "app_flow_copilot")).status, "offline");
    assert.equal((await readWorkflowRAGPromotionCandidate(offline, "app_flow_copilot", "wragp_aaaaaaaaaaaaaaaa")).status, "offline");
    assert.equal((await createWorkflowRAGPromotionCandidate(offline, "app_flow_copilot", datasetBinding(), "wragr_aaaaaaaaaaaaaaaa", { draftId: "app-config-flow", draftVersion: 1 })).status, "offline");
    assert.equal((await decideWorkflowRAGPromotionCandidate(offline, "app_flow_copilot", "wragp_aaaaaaaaaaaaaaaa", 1, "approve", "Explicit evidence approval.")).status, "offline");
    assert.equal((await listWorkflowRAGPromotionCandidates({ ...live, scopes: new Set() }, "app_flow_copilot")).status, "scope_denied");
    assert.equal(calls, 0);
  } finally { globalThis.fetch = originalFetch; }
});

test("promotion create sends exact authority refs and dedicated scopes", async () => {
  const originalFetch = globalThis.fetch;
  let captured: { url: string; headers: Headers; body: Record<string, unknown> } | undefined;
  globalThis.fetch = async (input, init) => {
    captured = { url: String(input), headers: new Headers(init?.headers), body: JSON.parse(String(init?.body)) };
    return jsonResponse(promotionEnvelope());
  };
  try {
    const result = await createWorkflowRAGPromotionCandidate(live, "app_flow_copilot", datasetBinding(), "wragr_aaaaaaaaaaaaaaaa", { draftId: "app-config-flow", draftVersion: 1 });
    assert.equal(result.status, "created");
    assert.equal(result.detail?.candidate.evidence.candidateReviewId, "wragr_aaaaaaaaaaaaaaaa");
    assert.equal(captured?.url, "http://platform.test/v1/user-workspace/workflow-rag-knowledge-promotion-candidates");
    assert.deepEqual(captured?.body, {
      workspace_id: "workspace_demo", application_id: "app_flow_copilot", dataset_id: "wragd_flow_quality", dataset_version: 2,
      dataset_digest: digest("d"), candidate_review_id: "wragr_aaaaaaaaaaaaaaaa", draft_id: "app-config-flow", expected_draft_version: 1,
    });
    assert.equal(captured?.headers.get("X-RadishMind-Dev-Read-Scopes"), "workflow_rag_promotions:write,workflow_rag_evaluation_datasets:read,workflow_rag_snapshots:read,application_drafts:read");
    assert.equal("baseline_snapshot" in captured!.body, false);
  } finally { globalThis.fetch = originalFetch; }
});

test("promotion detail maps immutable binding and metadata-only evidence", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => jsonResponse(promotionEnvelope({ approved: true }));
  try {
    const result = await readWorkflowRAGPromotionCandidate(live, "app_flow_copilot", "wragp_aaaaaaaaaaaaaaaa");
    assert.equal(result.status, "loaded");
    assert.equal(result.detail?.candidate.candidateState, "approved");
    assert.equal(result.detail?.binding?.bindingId, "wragb_aaaaaaaaaaaaaaaa");
    assert.equal(result.detail?.eligibility.eligible, true);
    assert.equal(JSON.stringify(result).includes("query_text"), false);
  } finally { globalThis.fetch = originalFetch; }
});

test("promotion decision preserves CAS conflict metadata without fallback", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => jsonResponse({ ...emptyEnvelope(), failure_code: "workflow_rag_promotion_record_version_conflict", current_record_version: 3, current_state: "approved" });
  try {
    const result = await decideWorkflowRAGPromotionCandidate(live, "app_flow_copilot", "wragp_aaaaaaaaaaaaaaaa", 2, "cancel", "Cancel stale binding eligibility.");
    assert.equal(result.status, "record_version_conflict");
    assert.equal(result.currentRecordVersion, 3);
    assert.equal(result.currentState, "approved");
    assert.equal(result.detail, null);
  } finally { globalThis.fetch = originalFetch; }
});

test("promotion list and detail reject scope, schema, and forbidden material", async () => {
  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async () => jsonResponse({ ...promotionListEnvelope(), application_id: "app_docs_assistant" });
    assert.equal((await listWorkflowRAGPromotionCandidates(live, "app_flow_copilot")).status, "failed");

    globalThis.fetch = async () => jsonResponse({ ...promotionEnvelope(), query_text: "must never appear" });
    assert.equal((await readWorkflowRAGPromotionCandidate(live, "app_flow_copilot", "wragp_aaaaaaaaaaaaaaaa")).failureCode, "workflow_rag_promotion_store_unavailable");

    const drift = promotionEnvelope();
    drift.candidate.evidence.profile.profile_digest = "invalid";
    globalThis.fetch = async () => jsonResponse(drift);
    assert.equal((await readWorkflowRAGPromotionCandidate(live, "app_flow_copilot", "wragp_aaaaaaaaaaaaaaaa")).detail, null);
  } finally { globalThis.fetch = originalFetch; }
});

test("promotion state transitions stay explicit", () => {
  assert.equal(workflowRAGPromotionDecisionAllowed("pending", "approve"), true);
  assert.equal(workflowRAGPromotionDecisionAllowed("deferred", "defer"), false);
  assert.equal(workflowRAGPromotionDecisionAllowed("approved", "cancel"), true);
  assert.equal(workflowRAGPromotionDecisionAllowed("approved", "approve"), false);
  assert.equal(workflowRAGPromotionDecisionAllowed("canceled", "cancel"), false);
});

function promotionListEnvelope() {
  return {
    request_id: "promotion-list-request", tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: "app_flow_copilot",
    items: [{ candidate_id: "wragp_aaaaaaaaaaaaaaaa", dataset: datasetDocument(), candidate_review_id: "wragr_aaaaaaaaaaaaaaaa", source_draft: sourceDraft(), candidate_state: "pending", record_version: 1, binding_ref: null, eligibility_status: "blocked", blocker_count: 1, created_at: timestamp(), updated_at: timestamp() }],
    next_cursor: null, failure_code: null, audit_ref: "audit-promotion-list-request",
  };
}

function promotionEnvelope(options: { approved?: boolean } = {}) {
  const approved = options.approved ?? false;
  const bindingRef = approved ? bindingRefDocument() : null;
  const evidence = evidenceDocument();
  const candidate = {
    schema_version: "workflow_rag_knowledge_promotion_candidate.v1", candidate_id: "wragp_aaaaaaaaaaaaaaaa", candidate_digest: digest("a"),
    tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: "app_flow_copilot", owner_subject_ref: "subject_demo_user",
    evidence, candidate_state: approved ? "approved" : "pending", record_version: approved ? 2 : 1, binding_ref: bindingRef,
    created_at: timestamp(), updated_at: timestamp(), created_by_actor_ref: "subject_demo_user", updated_by_actor_ref: "subject_demo_user",
    request_id: "promotion-create-request", audit_ref: "audit-promotion-create-request",
  };
  const decision = approved ? [{ schema_version: "workflow_rag_knowledge_promotion_decision.v1", decision_id: "wragpd_aaaaaaaaaaaaaaaa", candidate_id: candidate.candidate_id, candidate_digest: candidate.candidate_digest, decision: "approve", reason: "Approved exact evidence binding.", from_state: "pending", to_state: "approved", before_record_version: 1, after_record_version: 2, actor_ref: "subject_demo_user", occurred_at: timestamp(), request_id: "promotion-decision-request", audit_ref: "audit-promotion-decision-request" }] : [];
  const binding = approved ? { schema_version: "workflow_rag_application_binding.v1", ...bindingRefDocument(), candidate_id: candidate.candidate_id, candidate_digest: candidate.candidate_digest, approved_decision_id: "wragpd_aaaaaaaaaaaaaaaa", approved_record_version: 2, tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: "app_flow_copilot", owner_subject_ref: "subject_demo_user", evidence, issued_at: timestamp(), issued_by_actor_ref: "subject_demo_user", request_id: "promotion-decision-request", audit_ref: "audit-promotion-decision-request" } : null;
  return { request_id: "promotion-envelope-request", tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: "app_flow_copilot", candidate, decisions: decision, binding, eligibility: { eligible: approved, status: approved ? "eligible" : "blocked", blockers: approved ? [] : [{ code: "workflow_rag_promotion_not_approved" }] }, failure_code: null, current_record_version: candidate.record_version, current_state: candidate.candidate_state, audit_ref: "audit-promotion-envelope-request" };
}

function emptyEnvelope() { return { request_id: "promotion-conflict-request", tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: "app_flow_copilot", candidate: null, decisions: [], binding: null, eligibility: { eligible: false, status: "blocked", blockers: [] }, failure_code: null, current_record_version: 0, current_state: "", audit_ref: "audit-promotion-conflict-request" }; }
function evidenceDocument() { return { dataset: datasetDocument(), candidate_review_id: "wragr_aaaaaaaaaaaaaaaa", baseline_snapshot: snapshotDocument("a", "baseline_docs"), candidate_snapshot: snapshotDocument("b", "candidate_docs"), profile: { profile_id: "workflow.rag.lexical-ngram-dev.v1", profile_version: 1, profile_digest: digest("c") }, source_draft: sourceDraft() }; }
function datasetBinding() { return { datasetId: "wragd_flow_quality", datasetVersion: 2, datasetDigest: digest("d") }; }
function datasetDocument() { return { dataset_id: "wragd_flow_quality", dataset_version: 2, dataset_digest: digest("d") }; }
function snapshotDocument(seed: string, key: string) { return { tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: "app_flow_copilot", snapshot_id: `rags_${seed.repeat(16)}`, snapshot_version: 1, snapshot_digest: digest(seed), rag_ref: `workflow.rag.${key}.v1` }; }
function sourceDraft() { return { draft_id: "app-config-flow", draft_version: 1, draft_digest: digest("e"), base_application_updated_at: timestamp() }; }
function bindingRefDocument() { return { binding_id: "wragb_aaaaaaaaaaaaaaaa", binding_version: 1, binding_digest: digest("f") }; }
function digest(seed: string) { return `sha256:${seed.repeat(64)}`; }
function timestamp() { return "2026-07-18T10:00:00Z"; }
function jsonResponse(document: unknown): Response { return new Response(JSON.stringify(document), { status: 200, headers: { "Content-Type": "application/json" } }); }
