import assert from "node:assert/strict";
import test from "node:test";

import {
  archiveWorkflowRAGEvaluationDataset,
  createWorkflowRAGCandidateReview,
  createWorkflowRAGEvaluationDataset,
  listWorkflowRAGCandidateReviews,
  listWorkflowRAGEvaluationDatasets,
  readWorkflowRAGCandidateReview,
  readWorkflowRAGEvaluationDataset,
  versionWorkflowRAGEvaluationDataset,
  type WorkflowRAGEvaluationConfig,
  type WorkflowRAGEvaluationDatasetInput,
} from "../src/features/control-plane-read/workflowRAGEvaluationDatasetConsumer.ts";

const applicationId = "app_flow_copilot";
const datasetId = "wragd_aaaaaaaaaaaaaaaa";
const reviewId = "wragr_bbbbbbbbbbbbbbbb";
const baselineId = "rags_aaaaaaaaaaaaaaaa";
const candidateId = "rags_bbbbbbbbbbbbbbbb";
const digestA = `sha256:${"a".repeat(64)}`;
const digestB = `sha256:${"b".repeat(64)}`;
const digestC = `sha256:${"c".repeat(64)}`;

const config: WorkflowRAGEvaluationConfig = {
  mode: "dev_workflow_rag_evaluation_http",
  baseUrl: "http://platform.test",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  subjectRef: "subject_demo_user",
  authMode: "dev_headers",
  scopes: new Set([
    "workflow_rag_evaluation_datasets:read",
    "workflow_rag_evaluation_datasets:read_content",
    "workflow_rag_evaluation_datasets:write",
    "workflow_rag_evaluation_datasets:review",
    "workflow_rag_evaluation_datasets:archive",
    "workflow_rag_snapshots:read",
  ]),
};

test("offline and missing evaluation scopes send zero requests", async () => {
  const originalFetch = globalThis.fetch;
  let calls = 0;
  globalThis.fetch = async () => { calls += 1; throw new Error("boundary must not fetch"); };
  try {
    const offline = { ...config, mode: "offline" as const };
    assert.equal((await listWorkflowRAGEvaluationDatasets(offline, applicationId, "active")).status, "offline");
    assert.equal((await createWorkflowRAGEvaluationDataset(offline, applicationId, datasetInput())).status, "offline");
    assert.equal((await readWorkflowRAGEvaluationDataset(offline, applicationId, datasetId, 1)).status, "offline");
    assert.equal((await createWorkflowRAGCandidateReview(offline, applicationId, datasetId, 1, digestA, candidateBinding())).status, "offline");

    const noScopes = { ...config, scopes: new Set<never>() };
    assert.equal((await listWorkflowRAGEvaluationDatasets(noScopes, applicationId, "active")).status, "scope_denied");
    assert.equal((await readWorkflowRAGEvaluationDataset(noScopes, applicationId, datasetId, 1)).status, "scope_denied");
    assert.equal((await archiveWorkflowRAGEvaluationDataset(noScopes, applicationId, datasetId, 1)).status, "scope_denied");
    assert.equal(calls, 0);
  } finally { globalThis.fetch = originalFetch; }
});

test("create sends complete replacement input and exact write plus snapshot scopes", async () => {
  const originalFetch = globalThis.fetch;
  let captured: { url: string; headers: Headers; body: Record<string, unknown> } | null = null;
  globalThis.fetch = async (input, init) => {
    captured = { url: String(input), headers: new Headers(init?.headers), body: JSON.parse(String(init?.body)) };
    return jsonResponse(operationEnvelope());
  };
  try {
    const result = await createWorkflowRAGEvaluationDataset(config, applicationId, datasetInput());
    assert.equal(result.status, "created");
    assert.equal(result.version?.samples[0]?.queryText, "approved pressure procedure");
    assert.equal(captured?.url, "http://platform.test/v1/user-workspace/workflow-rag-evaluation-datasets");
    assert.equal(captured?.headers.get("X-RadishMind-Dev-Read-Scopes"), "workflow_rag_evaluation_datasets:write,workflow_rag_snapshots:read");
    assert.deepEqual(captured?.body, {
      workspace_id: "workspace_demo",
      application_id: applicationId,
      dataset_key: "pressure_quality",
      display_name: "Pressure quality",
      content_classification: "synthetic_public",
      baseline_snapshot: snapshotBindingDocument(baselineId, "workflow.rag.baseline_docs.v1", digestB),
      thresholds: metricsDocument(1),
      review_summary: "Human reviewed baseline evidence.",
      samples: [{ sample_id: "pressure_sample", query_text: "approved pressure procedure", expectation: "evidence_required", expected_citation_refs: ["official_guide"], required_official_refs: ["official_guide"], top_k: 1, review_note: "Official guide required." }],
    });
  } finally { globalThis.fetch = originalFetch; }
});

test("list stays metadata-only and exact detail requires read_content", async () => {
  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async () => jsonResponse(listEnvelope());
    const listed = await listWorkflowRAGEvaluationDatasets(config, applicationId, "active");
    assert.equal(listed.status, "ready");
    assert.equal(JSON.stringify(listed).includes("approved pressure procedure"), false);

    const leaked = listEnvelope() as ReturnType<typeof listEnvelope> & { query_text?: string };
    leaked.query_text = "must not be listed";
    globalThis.fetch = async () => jsonResponse(leaked);
    assert.equal((await listWorkflowRAGEvaluationDatasets(config, applicationId, "active")).failureCode, "workflow_rag_evaluation_store_unavailable");

    globalThis.fetch = async () => jsonResponse(operationEnvelope());
    const detail = await readWorkflowRAGEvaluationDataset(config, applicationId, datasetId, 1);
    assert.equal(detail.status, "loaded");
    assert.equal(detail.version?.samples[0]?.queryText, "approved pressure procedure");

    let calls = 0;
    globalThis.fetch = async () => { calls += 1; return jsonResponse(operationEnvelope()); };
    const metadataOnly = { ...config, scopes: new Set([...config.scopes].filter((scope) => scope !== "workflow_rag_evaluation_datasets:read_content")) };
    assert.equal((await readWorkflowRAGEvaluationDataset(metadataOnly, applicationId, datasetId, 1)).status, "scope_denied");
    assert.equal(calls, 0);
  } finally { globalThis.fetch = originalFetch; }
});

test("candidate review is metadata-only, exact scoped, and rejects query or unknown output", async () => {
  const originalFetch = globalThis.fetch;
  const requests: Array<{ headers: Headers; body: Record<string, unknown> }> = [];
  try {
    globalThis.fetch = async (_input, init) => {
      requests.push({ headers: new Headers(init?.headers), body: init?.body ? JSON.parse(String(init.body)) : {} });
      return jsonResponse(reviewOperationEnvelope());
    };
    const created = await createWorkflowRAGCandidateReview(config, applicationId, datasetId, 1, digestA, candidateBinding());
    assert.equal(created.status, "reviewed");
    assert.equal(created.review?.conclusion, "regressed");
    assert.equal(JSON.stringify(created).includes("approved pressure procedure"), false);
    assert.equal(requests[0]?.headers.get("X-RadishMind-Dev-Read-Scopes"), "workflow_rag_evaluation_datasets:review,workflow_rag_evaluation_datasets:read,workflow_rag_snapshots:read");
    assert.deepEqual(requests[0]?.body, { workspace_id: "workspace_demo", application_id: applicationId, dataset_version: 1, dataset_digest: digestA, candidate_snapshot: snapshotBindingDocument(candidateId, "workflow.rag.candidate_docs.v1", digestC) });

    const leaked = reviewOperationEnvelope();
    Object.assign(leaked.review!.candidate.samples[0]!, { query_text: "leaked query" });
    globalThis.fetch = async () => jsonResponse(leaked);
    assert.equal((await readWorkflowRAGCandidateReview(config, applicationId, datasetId, reviewId)).failureCode, "workflow_rag_evaluation_store_unavailable");

    const unknown = reviewOperationEnvelope();
    Object.assign(unknown.review!, { debug: "unexpected" });
    globalThis.fetch = async () => jsonResponse(unknown);
    assert.equal((await readWorkflowRAGCandidateReview(config, applicationId, datasetId, reviewId)).failureCode, "workflow_rag_evaluation_store_unavailable");

    const drift = reviewOperationEnvelope();
    drift.review!.application_id = "app_other";
    globalThis.fetch = async () => jsonResponse(drift);
    assert.equal((await readWorkflowRAGCandidateReview(config, applicationId, datasetId, reviewId)).failureCode, "workflow_rag_evaluation_store_unavailable");
  } finally { globalThis.fetch = originalFetch; }
});

test("review history, version CAS, and archive never call retrieval execution", async () => {
  const originalFetch = globalThis.fetch;
  const requests: Array<{ url: string; headers: Headers; body: Record<string, unknown> }> = [];
  try {
    globalThis.fetch = async (input, init) => {
      requests.push({ url: String(input), headers: new Headers(init?.headers), body: init?.body ? JSON.parse(String(init.body)) : {} });
      if (String(input).includes("candidate-reviews") && !init?.method) return jsonResponse(reviewListEnvelope());
      return jsonResponse(requests.length === 2 ? operationEnvelope({ version: 2 }) : operationEnvelope({ version: 2, lifecycle: "archived", includeVersion: false }));
    };
    const history = await listWorkflowRAGCandidateReviews(config, applicationId, datasetId);
    assert.equal(history.status, "ready");
    assert.equal(history.reviews[0]?.reviewId, reviewId);

    const versioned = await versionWorkflowRAGEvaluationDataset(config, applicationId, datasetId, 1, datasetInput());
    assert.equal(versioned.status, "versioned");
    assert.equal(requests[1]?.body.expected_latest_version, 1);
    assert.equal("dataset_key" in requests[1]!.body, false);

    const archived = await archiveWorkflowRAGEvaluationDataset(config, applicationId, datasetId, 2);
    assert.equal(archived.status, "archived");
    assert.equal(requests[2]?.headers.get("X-RadishMind-Dev-Read-Scopes"), "workflow_rag_evaluation_datasets:archive");
    assert.equal(requests.some((request) => request.url.includes("retrieval-executions")), false);
  } finally { globalThis.fetch = originalFetch; }
});

function datasetInput(): WorkflowRAGEvaluationDatasetInput {
  return {
    datasetKey: "pressure_quality",
    displayName: "Pressure quality",
    contentClassification: "synthetic_public",
    baselineSnapshot: { tenantRef: "tenant_demo", workspaceId: "workspace_demo", applicationId, snapshotId: baselineId, snapshotVersion: 1, snapshotDigest: digestB, ragRef: "workflow.rag.baseline_docs.v1" },
    thresholds: { hitAtK: 1, expectedRecallAtK: 1, requiredOfficialRecallAtK: 1, meanReciprocalRank: 1, noEvidenceAccuracy: 1, samplePassRate: 1 },
    reviewSummary: "Human reviewed baseline evidence.",
    samples: [{ sampleId: "pressure_sample", queryText: "approved pressure procedure", expectation: "evidence_required", expectedCitationRefs: ["official_guide"], requiredOfficialRefs: ["official_guide"], topK: 1, reviewNote: "Official guide required." }],
  };
}

function candidateBinding() { return { tenantRef: "tenant_demo", workspaceId: "workspace_demo", applicationId, snapshotId: candidateId, snapshotVersion: 1, snapshotDigest: digestC, ragRef: "workflow.rag.candidate_docs.v1" }; }
function operationEnvelope(options: { version?: number; lifecycle?: "active" | "archived"; includeVersion?: boolean } = {}) {
  const version = options.version ?? 1;
  const lifecycle = options.lifecycle ?? "active";
  return { request_id: "request_eval_operation", tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: applicationId, resource: resourceDocument(version, lifecycle), version: options.includeVersion === false ? null : versionDocument(version, lifecycle), review: null, failure_code: null as string | null, current_latest_version: version, current_lifecycle_state: lifecycle, audit_ref: "audit_eval_operation" };
}
function reviewOperationEnvelope() { return { request_id: "request_eval_review", tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: applicationId, resource: resourceDocument(), version: null, review: reviewDocument(), failure_code: null as string | null, current_latest_version: 1, current_lifecycle_state: "active", audit_ref: "audit_eval_review" }; }
function listEnvelope() { return { request_id: "request_eval_list", tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: applicationId, items: [resourceDocument()], next_cursor: null as string | null, failure_code: null as string | null, audit_ref: "audit_eval_list" }; }
function reviewListEnvelope() { return { request_id: "request_review_list", tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: applicationId, items: [reviewDocument()], next_cursor: null as string | null, failure_code: null as string | null, audit_ref: "audit_review_list" }; }

function resourceDocument(version = 1, lifecycle: "active" | "archived" = "active") { return { schema_version: "workflow_rag_evaluation_dataset_resource.v1", dataset_id: datasetId, tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: applicationId, dataset_key: "pressure_quality", display_name: "Pressure quality", lifecycle_state: lifecycle, content_classification: "synthetic_public", latest_version: version, latest_digest: digestA, baseline_snapshot: snapshotBindingDocument(baselineId, "workflow.rag.baseline_docs.v1", digestB), sample_count: 1, created_at: "2026-07-18T08:00:00Z", updated_at: "2026-07-18T08:01:00Z", archived_at: lifecycle === "archived" ? "2026-07-18T08:02:00Z" : null }; }
function versionDocument(version = 1, lifecycle: "active" | "archived" = "active") { return { schema_version: "workflow_rag_evaluation_dataset_resource.v1", dataset_key: "pressure_quality", display_name: "Pressure quality", lifecycle_state: lifecycle, created_at: "2026-07-18T08:00:00Z", created_by_actor_ref: "subject_demo_user", request_id: "request_eval_operation", audit_ref: "audit_eval_operation", dataset: { schema_version: "workflow_rag_evaluation_dataset.v1", dataset_id: datasetId, dataset_version: version, dataset_digest: digestA, content_classification: "synthetic_public", snapshot: snapshotBindingDocument(baselineId, "workflow.rag.baseline_docs.v1", digestB), profile: profileDocument(), thresholds: metricsDocument(1), review: { reviewer_ref: "subject_demo_user", reviewed_at: "2026-07-18T08:00:00Z", review_summary: "Human reviewed baseline evidence." }, samples: [{ sample_id: "pressure_sample", query_text: "approved pressure procedure", expectation: "evidence_required", expected_citation_refs: ["official_guide"], required_official_refs: ["official_guide"], top_k: 1, review_note: "Official guide required." }] } }; }
function reviewDocument() { const baseline = qualityDocument("passed", baselineId, "workflow.rag.baseline_docs.v1", digestB, true); const candidate = qualityDocument("needs_review", candidateId, "workflow.rag.candidate_docs.v1", digestC, false); return { schema_version: "workflow_rag_candidate_snapshot_review.v1", review_id: reviewId, tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: applicationId, dataset: { dataset_id: datasetId, dataset_version: 1, dataset_digest: digestA }, profile: profileDocument(), baseline_snapshot: snapshotBindingDocument(baselineId, "workflow.rag.baseline_docs.v1", digestB), candidate_snapshot: snapshotBindingDocument(candidateId, "workflow.rag.candidate_docs.v1", digestC), baseline_lifecycle: "active", candidate_lifecycle: "active", conclusion: "regressed", baseline, candidate, metric_delta: metricsDocument(-1), samples: [{ sample_id: "pressure_sample", baseline_passed: true, candidate_passed: false, baseline_expected_hit_count: 1, candidate_expected_hit_count: 0, baseline_required_official_hit_count: 1, candidate_required_official_hit_count: 0, baseline_first_expected_rank: 1, candidate_first_expected_rank: 0, change: "regressed" }], added_finding_codes: ["official_evidence_absent"], removed_finding_codes: [], created_at: "2026-07-18T08:10:00Z", created_by_actor_ref: "subject_demo_user", request_id: "request_eval_review", audit_ref: "audit_eval_review" }; }
function qualityDocument(status: "passed" | "needs_review", snapshotId: string, ragRef: string, snapshotDigest: string, passed: boolean) { return { schema_version: "workflow_rag_quality_review.v1", dataset: { dataset_id: datasetId, dataset_version: 1, dataset_digest: digestA }, snapshot: snapshotBindingDocument(snapshotId, ragRef, snapshotDigest), profile: profileDocument(), status, sample_count: 1, thresholds: metricsDocument(1), metrics: metricsDocument(passed ? 1 : 0), knowledge_summary: { fragment_count: 1, official_fragment_count: passed ? 1 : 0, expected_fragment_count: passed ? 1 : 0, source_type_distribution: [{ source_type: passed ? "manual" : "wiki", count: 1 }] }, samples: [{ sample_id: "pressure_sample", expectation: "evidence_required", query_digest: digestA, query_bytes: 27, top_k: 1, failure_code: "", selected: passed ? [{ fragment_ref: "official_guide", rank: 1, content_digest: digestB, source_type: "manual", is_official: true }] : [{ fragment_ref: "candidate_note", rank: 1, content_digest: digestC, source_type: "wiki", is_official: false }], expected_ref_count: 1, expected_hit_count: passed ? 1 : 0, required_official_ref_count: 1, required_official_hit_count: passed ? 1 : 0, first_expected_rank: passed ? 1 : 0, passed }], findings: passed ? [] : [{ code: "official_evidence_absent", severity: "review_required", fragment_refs: [] }] }; }
function snapshotBindingDocument(snapshotId: string, ragRef: string, snapshotDigest: string) { return { tenant_ref: "tenant_demo", workspace_id: "workspace_demo", application_id: applicationId, snapshot_id: snapshotId, snapshot_version: 1, snapshot_digest: snapshotDigest, rag_ref: ragRef }; }
function profileDocument() { return { profile_id: "workflow.rag.lexical-ngram-dev.v1", profile_version: 1, profile_digest: digestB }; }
function metricsDocument(value: number) { return { hit_at_k: value, expected_recall_at_k: value, required_official_recall_at_k: value, mean_reciprocal_rank: value, no_evidence_accuracy: value, sample_pass_rate: value }; }
function jsonResponse(document: unknown): Response { return new Response(JSON.stringify(document), { status: 200, headers: { "Content-Type": "application/json" } }); }
