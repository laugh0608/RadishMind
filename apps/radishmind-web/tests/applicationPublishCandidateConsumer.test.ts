import assert from "node:assert/strict";
import test from "node:test";

import {
  createApplicationPublishCandidate,
  createApplicationPublishWorkspaceMemory,
  listApplicationPublishCandidates,
  parseApplicationPublishEvidence,
  readApplicationPublishCandidate,
  reviewApplicationPublishCandidate,
  validateApplicationPublishReview,
  type ApplicationPublishCandidateConfig,
} from "../src/features/control-plane-read/applicationPublishCandidateConsumer.ts";

const devConfig: ApplicationPublishCandidateConfig = {
  mode: "dev_application_publish_http",
  baseUrl: "http://platform.test",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  subjectRef: "subject_demo_user",
};

test("application publish workspace stays offline without fetching", async () => {
  let fetchCount = 0;
  globalThis.fetch = async () => { fetchCount += 1; throw new Error("offline fetch"); };
  const offline = { ...devConfig, mode: "offline" as const };
  await createApplicationPublishCandidate(offline, "app_flow_copilot", "candidate-1", "draft-1", 1, []);
  await listApplicationPublishCandidates(offline, "app_flow_copilot");
  await readApplicationPublishCandidate(offline, "app_flow_copilot", "candidate-1");
  await reviewApplicationPublishCandidate(offline, "app_flow_copilot", "candidate-1", 0, "approve", "Reviewed offline only.");
  assert.equal(fetchCount, 0);
});

test("candidate create sends only binding fields and exact application scope", async () => {
  let captured: { url: string; headers: Headers; body: any } | undefined;
  globalThis.fetch = async (input, init) => {
    captured = { url: String(input), headers: new Headers(init?.headers), body: JSON.parse(String(init?.body)) };
    return jsonResponse(candidateEnvelope());
  };
  const result = await createApplicationPublishCandidate(devConfig, "app_flow_copilot", "candidate-app-flow-v1", "app-config-app-flow", 3, ["playground-request-0001"]);
  assert.equal(result.state.status, "created");
  assert.equal(result.candidate?.draftDigest, `sha256:${"a".repeat(64)}`);
  assert.deepEqual(captured?.body, { candidate_id: "candidate-app-flow-v1", draft_id: "app-config-app-flow", expected_draft_version: 3, evidence_request_ids: ["playground-request-0001"] });
  assert.equal(captured?.headers.get("X-RadishMind-Dev-Application-Publish-Workspace"), "workspace_demo");
  assert.equal(captured?.headers.get("X-RadishMind-Dev-Application-Publish-Application"), "app_flow_copilot");
  assert.equal(captured?.headers.get("X-RadishMind-Dev-Read-Scopes"), "application_publish_candidates:write,workflow_rag_promotions:read");
  assert.equal("configuration" in captured!.body, false);
});

test("review maps CAS conflict and keeps current review metadata", async () => {
  globalThis.fetch = async () => jsonResponse({ ...candidateEnvelope(), candidate: null, failure_code: "publish_candidate_review_version_conflict", current_review_version: 2, current_candidate_state: "approved" });
  const result = await reviewApplicationPublishCandidate(devConfig, "app_flow_copilot", "candidate-app-flow-v1", 1, "reject", "Stale review must not overwrite approval.");
  assert.equal(result.state.status, "review_version_conflict");
  assert.equal(result.state.currentReviewVersion, 2);
  assert.equal(result.state.currentCandidateState, "approved");
  assert.equal(result.candidate, null);
});

test("candidate list and detail reject scope drift and forbidden material", async () => {
  globalThis.fetch = async () => jsonResponse({ ...candidateEnvelope(), application_id: "app_docs_assistant" });
  const scopeDrift = await readApplicationPublishCandidate(devConfig, "app_flow_copilot", "candidate-app-flow-v1");
  assert.equal(scopeDrift.state.failureCode, "publish_candidate_store_unavailable");

  globalThis.fetch = async () => jsonResponse({ ...candidateListEnvelope(), raw_request: "forbidden" });
  const forbidden = await listApplicationPublishCandidates(devConfig, "app_flow_copilot");
  assert.equal(forbidden.status, "failed");
  assert.equal(forbidden.failureCode, "publish_candidate_store_unavailable");

  const invalidProtocol = candidateEnvelope();
  invalidProtocol.candidate.configuration.default_protocol = "unknown";
  globalThis.fetch = async () => jsonResponse(invalidProtocol);
  const invalidDetail = await readApplicationPublishCandidate(devConfig, "app_flow_copilot", "candidate-app-flow-v1");
  assert.equal(invalidDetail.state.failureCode, "publish_candidate_store_unavailable");

  const invalidDigest = candidateListEnvelope();
  invalidDigest.candidate_summaries[0].draft_digest = "not-a-digest";
  globalThis.fetch = async () => jsonResponse(invalidDigest);
  const invalidList = await listApplicationPublishCandidates(devConfig, "app_flow_copilot");
  assert.equal(invalidList.failureCode, "publish_candidate_store_unavailable");
});

test("publish candidate v2 consumes the exact binding ref and dynamic blocker", async () => {
  const bindingDigest = `sha256:${"b".repeat(64)}`;
  const envelope = candidateEnvelope();
  envelope.candidate.schema_version = "application_publish_candidate.v2";
  envelope.candidate.configuration.workflow_rag_binding_ref = { binding_id: "wragb_aaaaaaaaaaaaaaaa", binding_version: 1, binding_digest: bindingDigest };
  envelope.candidate.promotion_eligibility.blockers.unshift({ code: "workflow_rag_promotion_dataset_archived", summary: "Workflow RAG binding is no longer eligible." });
  globalThis.fetch = async () => jsonResponse(envelope);
  const result = await readApplicationPublishCandidate(devConfig, "app_flow_copilot", "candidate-app-flow-v1");
  assert.equal(result.candidate?.schemaVersion, "application_publish_candidate.v2");
  assert.equal(result.candidate?.configuration.workflowRAGBindingRef?.bindingDigest, bindingDigest);
  assert.equal(result.candidate?.promotionEligibility.blockers.some((item) => item.code === "workflow_rag_promotion_dataset_archived"), true);
});

test("evidence and review validation reject secrets and normalize safe refs", () => {
  assert.deepEqual(parseApplicationPublishEvidence("playground-request-0002\nplayground-request-0001,playground-request-0001"), {
    requestIds: ["playground-request-0001", "playground-request-0002"], failureCode: "",
  });
  assert.equal(parseApplicationPublishEvidence("Authorization: Bearer hidden").failureCode, "publish_candidate_payload_invalid");
  assert.equal(validateApplicationPublishReview("approve", "Authorization: Bearer hidden"), "publish_candidate_secret_material_forbidden");
  assert.equal(validateApplicationPublishReview("approve", "Reviewed configuration and history references."), "");
});

test("application switching creates isolated in-memory publish state", () => {
  const first = { ...createApplicationPublishWorkspaceMemory("app_flow_copilot", "one"), selectedDraftId: "draft-flow", evidenceText: "playground-request-0001", reviewReason: "Flow review" };
  const switched = createApplicationPublishWorkspaceMemory("app_docs_assistant", "two");
  assert.equal(first.applicationId, "app_flow_copilot");
  assert.deepEqual(switched, { applicationId: "app_docs_assistant", candidateId: "publish-app_docs_assistant-two", selectedDraftId: "", evidenceText: "", reviewReason: "" });
});

function candidateEnvelope() {
  return {
    request_id: "app-publish-request-0001", workspace_id: "workspace_demo", application_id: "app_flow_copilot",
    candidate: {
      schema_version: "application_publish_candidate.v1", candidate_id: "candidate-app-flow-v1",
      workspace_id: "workspace_demo", application_id: "app_flow_copilot", draft_id: "app-config-app-flow",
      draft_version: 3, draft_digest: `sha256:${"a".repeat(64)}`, base_application_updated_at: "2026-05-31T10:20:00Z",
      configuration: { display_name: "RadishFlow Copilot", description: "Sanitized candidate.", application_kind: "workflow_copilot", default_protocol: "responses", default_model: "radishmind-local-dev", allowed_protocols: ["responses"] },
      evidence_request_ids: ["playground-request-0001"], candidate_state: "pending_review", review_version: 0, reviews: [],
      promotion_eligibility: { eligible: false, status: "promotion_blocked", blockers: [{ code: "promotion_disabled", summary: "Promotion runtime is disabled." }] },
      created_at: "2026-07-12T10:00:00Z", updated_at: "2026-07-12T10:00:00Z",
      created_by_actor_ref: "subject_demo_user", updated_by_actor_ref: "subject_demo_user",
      request_id: "app-publish-request-0001", audit_ref: "audit-app-publish-request-0001",
    },
    failure_code: null, current_review_version: 0, current_candidate_state: "pending_review", current_draft_version: 0,
    audit_ref: "audit-app-publish-request-0001",
  };
}

function candidateListEnvelope() {
  return {
    request_id: "app-publish-list-0001", workspace_id: "workspace_demo", application_id: "app_flow_copilot",
    candidate_summaries: [{ candidate_id: "candidate-app-flow-v1", application_id: "app_flow_copilot", draft_id: "app-config-app-flow", draft_version: 3, draft_digest: `sha256:${"a".repeat(64)}`, candidate_state: "pending_review", review_version: 0, promotion_status: "promotion_blocked", promotion_blockers: 5, created_at: "2026-07-12T10:00:00Z", updated_at: "2026-07-12T10:00:00Z", updated_by_actor_ref: "subject_demo_user" }],
    failure_code: null, audit_ref: "audit-app-publish-list-0001",
  };
}

function jsonResponse(document: unknown): Response {
  return new Response(JSON.stringify(document), { status: 200, headers: { "Content-Type": "application/json" } });
}
