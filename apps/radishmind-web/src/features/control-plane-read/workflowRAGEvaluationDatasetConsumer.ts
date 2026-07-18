import { buildWorkflowRAGRequestHeaders } from "./workflowRAGSnapshotConsumer.ts";

const DEV_SOURCE = "dev-workflow-rag-evaluation-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const COLLECTION_PATH = "/v1/user-workspace/workflow-rag-evaluation-datasets";
const RESOURCE_SCHEMA = "workflow_rag_evaluation_dataset_resource.v1";
const DATASET_SCHEMA = "workflow_rag_evaluation_dataset.v1";
const REVIEW_SCHEMA = "workflow_rag_candidate_snapshot_review.v1";
const QUALITY_SCHEMA = "workflow_rag_quality_review.v1";
const DATASET_ID = /^wragd_[a-z0-9_]{3,47}$/u;
const REVIEW_ID = /^wragr_[a-z2-7]{16}$/u;
const SNAPSHOT_ID = /^rags_[a-z2-7]{16}$/u;
const KEY = /^[a-z][a-z0-9_]{2,47}$/u;
const SAMPLE_ID = /^[a-z][a-z0-9_]{2,47}$/u;
const FRAGMENT_REF = /^[a-z][a-z0-9_]{2,63}$/u;
const SCOPE_ID = /^[A-Za-z0-9][A-Za-z0-9_.:-]{2,119}$/u;
const REFERENCE = /^[A-Za-z0-9][A-Za-z0-9_.:/-]{2,159}$/u;
const DIGEST = /^sha256:[a-f0-9]{64}$/u;
const RAG_REF = /^workflow\.rag\.[a-z][a-z0-9_]{2,47}\.v[1-9][0-9]*$/u;
const FAILURE = /^(|workflow_rag_[a-z_]{3,80})$/u;
const CLASSIFICATIONS = ["synthetic_public", "workspace_internal"] as const;
const LIFECYCLES = ["active", "archived"] as const;
const CONCLUSIONS = ["improved", "unchanged", "regressed", "needs_review"] as const;
const FINDING_CODES = ["duplicate_fragment_content", "official_evidence_absent", "fragment_expectation_uncovered", "official_expectation_coverage_partial"] as const;
const SOURCE_TYPES = ["document", "wiki", "faq", "forum", "manual"] as const;
const RESOURCE_KEYS = ["schema_version", "dataset_id", "tenant_ref", "workspace_id", "application_id", "dataset_key", "display_name", "lifecycle_state", "content_classification", "latest_version", "latest_digest", "baseline_snapshot", "sample_count", "created_at", "updated_at", "archived_at"] as const;
const VERSION_KEYS = ["schema_version", "dataset_key", "display_name", "lifecycle_state", "created_at", "created_by_actor_ref", "request_id", "audit_ref", "dataset"] as const;
const DATASET_KEYS = ["schema_version", "dataset_id", "dataset_version", "dataset_digest", "content_classification", "snapshot", "profile", "thresholds", "review", "samples"] as const;
const SNAPSHOT_KEYS = ["tenant_ref", "workspace_id", "application_id", "snapshot_id", "snapshot_version", "snapshot_digest", "rag_ref"] as const;
const PROFILE_KEYS = ["profile_id", "profile_version", "profile_digest"] as const;
const METRIC_KEYS = ["hit_at_k", "expected_recall_at_k", "required_official_recall_at_k", "mean_reciprocal_rank", "no_evidence_accuracy", "sample_pass_rate"] as const;
const SAMPLE_KEYS = ["sample_id", "query_text", "expectation", "expected_citation_refs", "required_official_refs", "top_k", "review_note"] as const;
const REVIEW_METADATA_KEYS = ["reviewer_ref", "reviewed_at", "review_summary"] as const;
const OPERATION_KEYS = ["request_id", "tenant_ref", "workspace_id", "application_id", "resource", "version", "review", "failure_code", "current_latest_version", "current_lifecycle_state", "audit_ref"] as const;
const LIST_KEYS = ["request_id", "tenant_ref", "workspace_id", "application_id", "items", "next_cursor", "failure_code", "audit_ref"] as const;
const CANDIDATE_REVIEW_KEYS = ["schema_version", "review_id", "tenant_ref", "workspace_id", "application_id", "dataset", "profile", "baseline_snapshot", "candidate_snapshot", "baseline_lifecycle", "candidate_lifecycle", "conclusion", "baseline", "candidate", "metric_delta", "samples", "added_finding_codes", "removed_finding_codes", "created_at", "created_by_actor_ref", "request_id", "audit_ref"] as const;
const QUALITY_KEYS = ["schema_version", "dataset", "snapshot", "profile", "status", "sample_count", "thresholds", "metrics", "knowledge_summary", "samples", "findings"] as const;
const QUALITY_SAMPLE_KEYS = ["sample_id", "expectation", "query_digest", "query_bytes", "top_k", "failure_code", "selected", "expected_ref_count", "expected_hit_count", "required_official_ref_count", "required_official_hit_count", "first_expected_rank", "passed"] as const;
const SELECTED_KEYS = ["fragment_ref", "rank", "content_digest", "source_type", "is_official"] as const;
const KNOWLEDGE_KEYS = ["fragment_count", "official_fragment_count", "expected_fragment_count", "source_type_distribution"] as const;
const DISTRIBUTION_KEYS = ["source_type", "count"] as const;
const FINDING_KEYS = ["code", "severity", "fragment_refs"] as const;
const COMPARISON_KEYS = ["sample_id", "baseline_passed", "candidate_passed", "baseline_expected_hit_count", "candidate_expected_hit_count", "baseline_required_official_hit_count", "candidate_required_official_hit_count", "baseline_first_expected_rank", "candidate_first_expected_rank", "change"] as const;
const REVIEW_FORBIDDEN_KEYS = new Set(["query_text", "review_note", "content", "fragment_content", "prompt", "messages", "credential", "authorization", "raw_request", "raw_response"]);

export type WorkflowRAGEvaluationMode = "offline" | "dev_workflow_rag_evaluation_http";
export type WorkflowRAGEvaluationClassification = typeof CLASSIFICATIONS[number];
export type WorkflowRAGEvaluationLifecycle = typeof LIFECYCLES[number];
export type WorkflowRAGEvaluationScope = "workflow_rag_evaluation_datasets:read" | "workflow_rag_evaluation_datasets:read_content" | "workflow_rag_evaluation_datasets:write" | "workflow_rag_evaluation_datasets:review" | "workflow_rag_evaluation_datasets:archive" | "workflow_rag_snapshots:read";

export type WorkflowRAGEvaluationConfig = {
  mode: WorkflowRAGEvaluationMode;
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  subjectRef: string;
  authMode: "dev_headers" | "signed_test_token" | "radish_oidc_integration_test";
  scopes: ReadonlySet<WorkflowRAGEvaluationScope>;
};

export type WorkflowRAGEvaluationSnapshotBinding = {
  tenantRef: string; workspaceId: string; applicationId: string; snapshotId: string; snapshotVersion: number; snapshotDigest: string; ragRef: string;
};

export type WorkflowRAGEvaluationMetrics = {
  hitAtK: number; expectedRecallAtK: number; requiredOfficialRecallAtK: number; meanReciprocalRank: number; noEvidenceAccuracy: number; samplePassRate: number;
};

export type WorkflowRAGEvaluationSampleInput = {
  sampleId: string; queryText: string; expectation: "evidence_required" | "no_evidence"; expectedCitationRefs: string[]; requiredOfficialRefs: string[]; topK: number; reviewNote: string;
};

export type WorkflowRAGEvaluationDatasetInput = {
  datasetKey: string; displayName: string; contentClassification: WorkflowRAGEvaluationClassification; baselineSnapshot: WorkflowRAGEvaluationSnapshotBinding;
  thresholds: WorkflowRAGEvaluationMetrics; reviewSummary: string; samples: WorkflowRAGEvaluationSampleInput[];
};

export type WorkflowRAGEvaluationDatasetResource = {
  datasetId: string; datasetKey: string; displayName: string; lifecycleState: WorkflowRAGEvaluationLifecycle; contentClassification: WorkflowRAGEvaluationClassification;
  latestVersion: number; latestDigest: string; baselineSnapshot: WorkflowRAGEvaluationSnapshotBinding; sampleCount: number; createdAt: string; updatedAt: string; archivedAt: string | null;
};

export type WorkflowRAGEvaluationDatasetVersion = {
  datasetKey: string; displayName: string; lifecycleState: WorkflowRAGEvaluationLifecycle; createdAt: string; createdByActorRef: string; requestId: string; auditRef: string;
  datasetId: string; datasetVersion: number; datasetDigest: string; contentClassification: WorkflowRAGEvaluationClassification; snapshot: WorkflowRAGEvaluationSnapshotBinding;
  thresholds: WorkflowRAGEvaluationMetrics; reviewSummary: string; reviewerRef: string; reviewedAt: string; samples: WorkflowRAGEvaluationSampleInput[];
};

export type WorkflowRAGCandidateReviewSummary = {
  reviewId: string; datasetId: string; datasetVersion: number; datasetDigest: string; baselineSnapshot: WorkflowRAGEvaluationSnapshotBinding; candidateSnapshot: WorkflowRAGEvaluationSnapshotBinding;
  baselineLifecycle: WorkflowRAGEvaluationLifecycle; candidateLifecycle: WorkflowRAGEvaluationLifecycle; conclusion: typeof CONCLUSIONS[number]; baselineStatus: string; candidateStatus: string;
  baselineMetrics: WorkflowRAGEvaluationMetrics; candidateMetrics: WorkflowRAGEvaluationMetrics; metricDelta: WorkflowRAGEvaluationMetrics; sampleChanges: Array<{ sampleId: string; change: "improved" | "unchanged" | "regressed" }>;
  addedFindingCodes: string[]; removedFindingCodes: string[]; createdAt: string; createdByActorRef: string; requestId: string; auditRef: string;
};

export type WorkflowRAGEvaluationListResult = { status: "offline" | "scope_denied" | "ready" | "empty" | "failed"; resources: WorkflowRAGEvaluationDatasetResource[]; nextCursor: string; failureCode: string; summary: string };
export type WorkflowRAGCandidateReviewListResult = { status: "offline" | "scope_denied" | "ready" | "empty" | "failed"; reviews: WorkflowRAGCandidateReviewSummary[]; nextCursor: string; failureCode: string; summary: string };
export type WorkflowRAGEvaluationOperationResult = { status: "offline" | "scope_denied" | "created" | "loaded" | "versioned" | "archived" | "reviewed" | "version_conflict" | "failed"; resource: WorkflowRAGEvaluationDatasetResource | null; version: WorkflowRAGEvaluationDatasetVersion | null; review: WorkflowRAGCandidateReviewSummary | null; failureCode: string; currentLatestVersion: number; currentLifecycleState: string; summary: string };

type RecordDocument = Record<string, unknown>;

export function readWorkflowRAGEvaluationConfig(): WorkflowRAGEvaluationConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  return {
    mode: env.VITE_RADISHMIND_WORKFLOW_RAG_EVALUATION_SOURCE?.trim() === DEV_SOURCE ? "dev_workflow_rag_evaluation_http" : "offline",
    baseUrl: normalizeBaseUrl(env.VITE_RADISHMIND_WORKFLOW_RAG_EVALUATION_BASE_URL ?? env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ?? DEFAULT_BASE_URL),
    tenantRef: env.VITE_RADISHMIND_DEV_READ_TENANT_REF?.trim() || "tenant_demo",
    workspaceId: env.VITE_RADISHMIND_WORKFLOW_RAG_WORKSPACE_ID?.trim() || "workspace_demo",
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || "subject_demo_user",
    authMode: normalizeAuthMode(env.VITE_RADISHMIND_READ_AUTH_MODE),
    scopes: new Set((env.VITE_RADISHMIND_WORKFLOW_RAG_EVALUATION_SCOPES ?? "").split(",").map((value) => value.trim()).filter(isEvaluationScope)),
  };
}

export function validateWorkflowRAGEvaluationDatasetInput(config: WorkflowRAGEvaluationConfig, applicationId: string, input: WorkflowRAGEvaluationDatasetInput): string {
  if (!KEY.test(input.datasetKey.trim()) || input.displayName.trim().length < 2 || input.displayName.trim().length > 120 ||
    !CLASSIFICATIONS.includes(input.contentClassification) || !isSnapshotBindingInput(input.baselineSnapshot, config, applicationId) ||
    !validMetricsInput(input.thresholds) || input.reviewSummary.trim().length < 1 || input.reviewSummary.trim().length > 512 ||
    input.samples.length < 1 || input.samples.length > 200 || containsSecretMaterial(input.displayName) || containsSecretMaterial(input.reviewSummary)) return "workflow_rag_evaluation_payload_invalid";
  const sampleIds = new Set<string>();
  for (const sample of input.samples) {
    const queryBytes = new TextEncoder().encode(sample.queryText.trim()).length;
    if (!SAMPLE_ID.test(sample.sampleId.trim()) || sampleIds.has(sample.sampleId.trim()) || queryBytes < 1 || queryBytes > 4096 ||
      (sample.expectation !== "evidence_required" && sample.expectation !== "no_evidence") || !uniqueFragmentRefs(sample.expectedCitationRefs) ||
      !uniqueFragmentRefs(sample.requiredOfficialRefs) || sample.topK < 1 || sample.topK > 8 || !Number.isInteger(sample.topK) ||
      sample.reviewNote.trim().length < 1 || sample.reviewNote.trim().length > 512 || containsSecretMaterial(sample.queryText) || containsSecretMaterial(sample.reviewNote) ||
      (sample.expectation === "evidence_required" && sample.expectedCitationRefs.length < 1) ||
      (sample.expectation === "no_evidence" && (sample.expectedCitationRefs.length !== 0 || sample.requiredOfficialRefs.length !== 0))) return "workflow_rag_evaluation_payload_invalid";
    sampleIds.add(sample.sampleId.trim());
  }
  return "";
}

export async function listWorkflowRAGEvaluationDatasets(config: WorkflowRAGEvaluationConfig, applicationId: string, lifecycle: WorkflowRAGEvaluationLifecycle, cursor = ""): Promise<WorkflowRAGEvaluationListResult> {
  const boundary = boundaryFor(config, applicationId, ["workflow_rag_evaluation_datasets:read"]);
  if (boundary) return listBoundary(boundary);
  if (!LIFECYCLES.includes(lifecycle) || (cursor && !KEY.test(cursor))) return failedList("workflow_rag_evaluation_payload_invalid");
  const query = new URLSearchParams({ workspace_id: config.workspaceId, application_id: applicationId, lifecycle_state: lifecycle, limit: "100" });
  if (cursor) query.set("cursor", cursor);
  try {
    const value: unknown = await fetchJSON(`${config.baseUrl}${COLLECTION_PATH}?${query}`, { headers: headers(config, applicationId, ["workflow_rag_evaluation_datasets:read"], "list") });
    if (!isListEnvelope(value, config, applicationId, lifecycle)) return failedList();
    if (value.failure_code) return failedList(String(value.failure_code));
    const resources = (value.items as RecordDocument[]).map(mapResource);
    return { status: resources.length ? "ready" : "empty", resources, nextCursor: value.next_cursor as string ?? "", failureCode: "", summary: `Loaded ${resources.length} ${lifecycle} evaluation datasets.` };
  } catch { return failedList(); }
}

export async function createWorkflowRAGEvaluationDataset(config: WorkflowRAGEvaluationConfig, applicationId: string, input: WorkflowRAGEvaluationDatasetInput): Promise<WorkflowRAGEvaluationOperationResult> {
  const boundary = boundaryFor(config, applicationId, ["workflow_rag_evaluation_datasets:write", "workflow_rag_snapshots:read"]);
  if (boundary) return operationBoundary(boundary);
  const failure = validateWorkflowRAGEvaluationDatasetInput(config, applicationId, input);
  if (failure) return failedOperation(failure);
  return writeOperation(config, applicationId, COLLECTION_PATH, ["workflow_rag_evaluation_datasets:write", "workflow_rag_snapshots:read"], "create", mapCreateBody(config, applicationId, input), "created");
}

export async function readWorkflowRAGEvaluationDataset(config: WorkflowRAGEvaluationConfig, applicationId: string, datasetId: string, datasetVersion: number): Promise<WorkflowRAGEvaluationOperationResult> {
  const boundary = boundaryFor(config, applicationId, ["workflow_rag_evaluation_datasets:read", "workflow_rag_evaluation_datasets:read_content"]);
  if (boundary) return operationBoundary(boundary);
  if (!DATASET_ID.test(datasetId) || !integer(datasetVersion, 1)) return failedOperation("workflow_rag_evaluation_payload_invalid");
  const query = new URLSearchParams({ workspace_id: config.workspaceId, application_id: applicationId, dataset_version: String(datasetVersion) });
  try {
    const value = await fetchJSON(`${config.baseUrl}${COLLECTION_PATH}/${encodeURIComponent(datasetId)}?${query}`, { headers: headers(config, applicationId, ["workflow_rag_evaluation_datasets:read", "workflow_rag_evaluation_datasets:read_content"], "read") });
    return mapOperation(value, config, applicationId, "loaded", "version");
  } catch { return failedOperation(); }
}

export async function versionWorkflowRAGEvaluationDataset(config: WorkflowRAGEvaluationConfig, applicationId: string, datasetId: string, expectedLatestVersion: number, input: WorkflowRAGEvaluationDatasetInput): Promise<WorkflowRAGEvaluationOperationResult> {
  const boundary = boundaryFor(config, applicationId, ["workflow_rag_evaluation_datasets:write", "workflow_rag_snapshots:read"]);
  if (boundary) return operationBoundary(boundary);
  const failure = validateWorkflowRAGEvaluationDatasetInput(config, applicationId, input);
  if (!DATASET_ID.test(datasetId) || !integer(expectedLatestVersion, 1) || failure) return failedOperation(failure || "workflow_rag_evaluation_payload_invalid");
  const body = mapCreateBody(config, applicationId, input) as RecordDocument;
  delete body.dataset_key;
  body.expected_latest_version = expectedLatestVersion;
  return writeOperation(config, applicationId, `${COLLECTION_PATH}/${encodeURIComponent(datasetId)}/versions`, ["workflow_rag_evaluation_datasets:write", "workflow_rag_snapshots:read"], "version", body, "versioned");
}

export async function archiveWorkflowRAGEvaluationDataset(config: WorkflowRAGEvaluationConfig, applicationId: string, datasetId: string, expectedLatestVersion: number): Promise<WorkflowRAGEvaluationOperationResult> {
  const boundary = boundaryFor(config, applicationId, ["workflow_rag_evaluation_datasets:archive"]);
  if (boundary) return operationBoundary(boundary);
  if (!DATASET_ID.test(datasetId) || !integer(expectedLatestVersion, 1)) return failedOperation("workflow_rag_evaluation_payload_invalid");
  return writeOperation(config, applicationId, `${COLLECTION_PATH}/${encodeURIComponent(datasetId)}/archive`, ["workflow_rag_evaluation_datasets:archive"], "archive", { workspace_id: config.workspaceId, application_id: applicationId, expected_latest_version: expectedLatestVersion }, "archived");
}

export async function createWorkflowRAGCandidateReview(config: WorkflowRAGEvaluationConfig, applicationId: string, datasetId: string, datasetVersion: number, datasetDigest: string, candidateSnapshot: WorkflowRAGEvaluationSnapshotBinding): Promise<WorkflowRAGEvaluationOperationResult> {
  const required: WorkflowRAGEvaluationScope[] = ["workflow_rag_evaluation_datasets:review", "workflow_rag_evaluation_datasets:read", "workflow_rag_snapshots:read"];
  const boundary = boundaryFor(config, applicationId, required);
  if (boundary) return operationBoundary(boundary);
  if (!DATASET_ID.test(datasetId) || !integer(datasetVersion, 1) || !DIGEST.test(datasetDigest) || !isSnapshotBindingInput(candidateSnapshot, config, applicationId)) return failedOperation("workflow_rag_evaluation_payload_invalid");
  return writeOperation(config, applicationId, `${COLLECTION_PATH}/${encodeURIComponent(datasetId)}/candidate-reviews`, required, "candidate-review", { workspace_id: config.workspaceId, application_id: applicationId, dataset_version: datasetVersion, dataset_digest: datasetDigest, candidate_snapshot: mapSnapshotBinding(candidateSnapshot) }, "reviewed");
}

export async function listWorkflowRAGCandidateReviews(config: WorkflowRAGEvaluationConfig, applicationId: string, datasetId: string, cursor = ""): Promise<WorkflowRAGCandidateReviewListResult> {
  const boundary = boundaryFor(config, applicationId, ["workflow_rag_evaluation_datasets:read"]);
  if (boundary) return reviewListBoundary(boundary);
  if (!DATASET_ID.test(datasetId)) return failedReviewList("workflow_rag_evaluation_payload_invalid");
  const query = new URLSearchParams({ workspace_id: config.workspaceId, application_id: applicationId, limit: "100" });
  if (cursor) query.set("cursor", cursor);
  try {
    const value = await fetchJSON(`${config.baseUrl}${COLLECTION_PATH}/${encodeURIComponent(datasetId)}/candidate-reviews?${query}`, { headers: headers(config, applicationId, ["workflow_rag_evaluation_datasets:read"], "review-list") });
    if (!isReviewListEnvelope(value, config, applicationId, datasetId)) return failedReviewList();
    if (value.failure_code) return failedReviewList(String(value.failure_code));
    const reviews = (value.items as RecordDocument[]).map(mapCandidateReview);
    return { status: reviews.length ? "ready" : "empty", reviews, nextCursor: value.next_cursor as string ?? "", failureCode: "", summary: `Loaded ${reviews.length} candidate reviews.` };
  } catch { return failedReviewList(); }
}

export async function readWorkflowRAGCandidateReview(config: WorkflowRAGEvaluationConfig, applicationId: string, datasetId: string, reviewId: string): Promise<WorkflowRAGEvaluationOperationResult> {
  const boundary = boundaryFor(config, applicationId, ["workflow_rag_evaluation_datasets:read"]);
  if (boundary) return operationBoundary(boundary);
  if (!DATASET_ID.test(datasetId) || !REVIEW_ID.test(reviewId)) return failedOperation("workflow_rag_evaluation_payload_invalid");
  const query = new URLSearchParams({ workspace_id: config.workspaceId, application_id: applicationId });
  try {
    const value = await fetchJSON(`${config.baseUrl}${COLLECTION_PATH}/${encodeURIComponent(datasetId)}/candidate-reviews/${encodeURIComponent(reviewId)}?${query}`, { headers: headers(config, applicationId, ["workflow_rag_evaluation_datasets:read"], "review-read") });
    return mapOperation(value, config, applicationId, "reviewed", "review");
  } catch { return failedOperation(); }
}

async function writeOperation(config: WorkflowRAGEvaluationConfig, applicationId: string, path: string, scopes: WorkflowRAGEvaluationScope[], operation: string, body: unknown, success: WorkflowRAGEvaluationOperationResult["status"]): Promise<WorkflowRAGEvaluationOperationResult> {
  try {
    const value = await fetchJSON(`${config.baseUrl}${path}`, { method: "POST", headers: { ...headers(config, applicationId, scopes, operation), "Content-Type": "application/json" }, body: JSON.stringify(body) });
    return mapOperation(value, config, applicationId, success, success === "reviewed" ? "review" : success === "archived" ? "resource" : "version");
  } catch { return failedOperation(); }
}

function mapOperation(value: unknown, config: WorkflowRAGEvaluationConfig, applicationId: string, success: WorkflowRAGEvaluationOperationResult["status"], required: "resource" | "version" | "review"): WorkflowRAGEvaluationOperationResult {
  if (!isOperationEnvelope(value, config, applicationId)) return failedOperation();
  if (value.failure_code) {
    const code = String(value.failure_code);
    return { status: code === "workflow_rag_evaluation_version_conflict" ? "version_conflict" : code === "workflow_rag_evaluation_scope_denied" ? "scope_denied" : "failed", resource: null, version: null, review: null, failureCode: code, currentLatestVersion: Number(value.current_latest_version), currentLifecycleState: String(value.current_lifecycle_state), summary: "Evaluation dataset operation failed without fallback." };
  }
  if ((required === "resource" && value.resource === null) || (required === "version" && value.version === null) || (required === "review" && value.review === null)) return failedOperation();
  return { status: success, resource: value.resource ? mapResource(value.resource as RecordDocument) : null, version: value.version ? mapVersion(value.version as RecordDocument) : null, review: value.review ? mapCandidateReview(value.review as RecordDocument) : null, failureCode: "", currentLatestVersion: Number(value.current_latest_version), currentLifecycleState: String(value.current_lifecycle_state), summary: `Evaluation dataset ${success}.` };
}

function isOperationEnvelope(value: unknown, config: WorkflowRAGEvaluationConfig, applicationId: string): value is RecordDocument {
  if (!record(value) || !exact(value, OPERATION_KEYS) || !nonEmpty(value.request_id) || value.tenant_ref !== config.tenantRef || value.workspace_id !== config.workspaceId || value.application_id !== applicationId ||
    !(value.failure_code === null || nonEmpty(value.failure_code)) || !integer(value.current_latest_version, 0) || typeof value.current_lifecycle_state !== "string" || !nonEmpty(value.audit_ref)) return false;
  if (value.failure_code !== null) return value.resource === null && value.version === null && value.review === null;
  return (value.resource === null || isResource(value.resource, config, applicationId)) && (value.version === null || isVersion(value.version, config, applicationId)) && (value.review === null || isCandidateReview(value.review, config, applicationId));
}

function isListEnvelope(value: unknown, config: WorkflowRAGEvaluationConfig, applicationId: string, lifecycle: WorkflowRAGEvaluationLifecycle): value is RecordDocument {
  return record(value) && exact(value, LIST_KEYS) && nonEmpty(value.request_id) && value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId && value.application_id === applicationId &&
    Array.isArray(value.items) && value.items.every((item) => isResource(item, config, applicationId, lifecycle)) && (value.next_cursor === null || KEY.test(String(value.next_cursor))) &&
    (value.failure_code === null || nonEmpty(value.failure_code)) && nonEmpty(value.audit_ref) && !containsSecretMaterial(JSON.stringify(value));
}

function isReviewListEnvelope(value: unknown, config: WorkflowRAGEvaluationConfig, applicationId: string, datasetId: string): value is RecordDocument {
  return record(value) && exact(value, LIST_KEYS) && nonEmpty(value.request_id) && value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId && value.application_id === applicationId &&
    Array.isArray(value.items) && value.items.every((item) => {
      if (!isCandidateReview(item, config, applicationId)) return false;
      return record(item.dataset) && item.dataset.dataset_id === datasetId;
    }) &&
    (value.next_cursor === null || (typeof value.next_cursor === "string" && value.next_cursor.includes("|"))) && (value.failure_code === null || nonEmpty(value.failure_code)) && nonEmpty(value.audit_ref);
}

function isResource(value: unknown, config: WorkflowRAGEvaluationConfig, applicationId: string, lifecycle?: WorkflowRAGEvaluationLifecycle): value is RecordDocument {
  return record(value) && exact(value, RESOURCE_KEYS) && value.schema_version === RESOURCE_SCHEMA && DATASET_ID.test(String(value.dataset_id)) && value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId && value.application_id === applicationId &&
    KEY.test(String(value.dataset_key)) && nonEmpty(value.display_name) && LIFECYCLES.includes(value.lifecycle_state as WorkflowRAGEvaluationLifecycle) && (!lifecycle || value.lifecycle_state === lifecycle) &&
    CLASSIFICATIONS.includes(value.content_classification as WorkflowRAGEvaluationClassification) && integer(value.latest_version, 1) && DIGEST.test(String(value.latest_digest)) &&
    isSnapshotBinding(value.baseline_snapshot, config, applicationId) && integer(value.sample_count, 1, 200) && timestamp(value.created_at) && timestamp(value.updated_at) &&
    (value.lifecycle_state === "active" ? value.archived_at === null : timestamp(value.archived_at));
}

function isVersion(value: unknown, config: WorkflowRAGEvaluationConfig, applicationId: string): value is RecordDocument {
  return record(value) && exact(value, VERSION_KEYS) && value.schema_version === RESOURCE_SCHEMA && KEY.test(String(value.dataset_key)) && nonEmpty(value.display_name) &&
    LIFECYCLES.includes(value.lifecycle_state as WorkflowRAGEvaluationLifecycle) && timestamp(value.created_at) && REFERENCE.test(String(value.created_by_actor_ref)) &&
    REFERENCE.test(String(value.request_id)) && REFERENCE.test(String(value.audit_ref)) && isDataset(value.dataset, config, applicationId);
}

function isDataset(value: unknown, config: WorkflowRAGEvaluationConfig, applicationId: string): value is RecordDocument {
  return record(value) && exact(value, DATASET_KEYS) && value.schema_version === DATASET_SCHEMA && DATASET_ID.test(String(value.dataset_id)) && integer(value.dataset_version, 1) && DIGEST.test(String(value.dataset_digest)) &&
    CLASSIFICATIONS.includes(value.content_classification as WorkflowRAGEvaluationClassification) && isSnapshotBinding(value.snapshot, config, applicationId) && isProfile(value.profile) &&
    isMetrics(value.thresholds, 0, 1) && record(value.review) && exact(value.review, REVIEW_METADATA_KEYS) && REFERENCE.test(String(value.review.reviewer_ref)) && timestamp(value.review.reviewed_at) && nonEmpty(value.review.review_summary) &&
    Array.isArray(value.samples) && value.samples.length >= 1 && value.samples.length <= 200 && value.samples.every(isDatasetSample);
}

function isDatasetSample(value: unknown): value is RecordDocument {
  return record(value) && exact(value, SAMPLE_KEYS) && SAMPLE_ID.test(String(value.sample_id)) && nonEmpty(value.query_text) && new TextEncoder().encode(String(value.query_text)).length <= 4096 &&
    (value.expectation === "evidence_required" || value.expectation === "no_evidence") && fragmentRefs(value.expected_citation_refs) && fragmentRefs(value.required_official_refs) && integer(value.top_k, 1, 8) && nonEmpty(value.review_note) && String(value.review_note).length <= 512 &&
    !containsSecretMaterial(String(value.query_text)) && !containsSecretMaterial(String(value.review_note));
}

function isCandidateReview(value: unknown, config: WorkflowRAGEvaluationConfig, applicationId: string): value is RecordDocument {
  if (!record(value) || !exact(value, CANDIDATE_REVIEW_KEYS) || containsReviewForbidden(value) || value.schema_version !== REVIEW_SCHEMA || !REVIEW_ID.test(String(value.review_id)) ||
    value.tenant_ref !== config.tenantRef || value.workspace_id !== config.workspaceId || value.application_id !== applicationId || !isDatasetBinding(value.dataset) || !isProfile(value.profile) ||
    !isSnapshotBinding(value.baseline_snapshot, config, applicationId) || !isSnapshotBinding(value.candidate_snapshot, config, applicationId) ||
    !LIFECYCLES.includes(value.baseline_lifecycle as WorkflowRAGEvaluationLifecycle) || !LIFECYCLES.includes(value.candidate_lifecycle as WorkflowRAGEvaluationLifecycle) ||
    !CONCLUSIONS.includes(value.conclusion as typeof CONCLUSIONS[number]) || !isQualityReview(value.baseline, config, applicationId) || !isQualityReview(value.candidate, config, applicationId) ||
    !isMetrics(value.metric_delta, -1, 1) || !Array.isArray(value.samples) || value.samples.length < 1 || !value.samples.every(isComparison) ||
    !findingCodes(value.added_finding_codes) || !findingCodes(value.removed_finding_codes) || !timestamp(value.created_at) || !REFERENCE.test(String(value.created_by_actor_ref)) || !REFERENCE.test(String(value.request_id)) || !REFERENCE.test(String(value.audit_ref))) return false;
  return true;
}

function isQualityReview(value: unknown, config: WorkflowRAGEvaluationConfig, applicationId: string): value is RecordDocument {
  return record(value) && exact(value, QUALITY_KEYS) && value.schema_version === QUALITY_SCHEMA && isDatasetBinding(value.dataset) && isSnapshotBinding(value.snapshot, config, applicationId) && isProfile(value.profile) &&
    (value.status === "passed" || value.status === "needs_review" || value.status === "failed") && integer(value.sample_count, 1, 200) && isMetrics(value.thresholds, 0, 1) && isMetrics(value.metrics, 0, 1) &&
    record(value.knowledge_summary) && exact(value.knowledge_summary, KNOWLEDGE_KEYS) && integer(value.knowledge_summary.fragment_count, 1, 256) && integer(value.knowledge_summary.official_fragment_count, 0, 256) && integer(value.knowledge_summary.expected_fragment_count, 0, 256) &&
    Array.isArray(value.knowledge_summary.source_type_distribution) && value.knowledge_summary.source_type_distribution.every((item) => record(item) && exact(item, DISTRIBUTION_KEYS) && SOURCE_TYPES.includes(item.source_type as typeof SOURCE_TYPES[number]) && integer(item.count, 1, 256)) &&
    Array.isArray(value.samples) && value.samples.length === value.sample_count && value.samples.every(isQualitySample) && Array.isArray(value.findings) && value.findings.every(isFinding);
}

function isQualitySample(value: unknown): value is RecordDocument {
  return record(value) && exact(value, QUALITY_SAMPLE_KEYS) && SAMPLE_ID.test(String(value.sample_id)) && (value.expectation === "evidence_required" || value.expectation === "no_evidence") && DIGEST.test(String(value.query_digest)) &&
    integer(value.query_bytes, 1, 4096) && integer(value.top_k, 1, 8) && typeof value.failure_code === "string" && FAILURE.test(value.failure_code) && Array.isArray(value.selected) && value.selected.length <= Number(value.top_k) &&
    value.selected.every((item, index) => record(item) && exact(item, SELECTED_KEYS) && FRAGMENT_REF.test(String(item.fragment_ref)) && item.rank === index + 1 && DIGEST.test(String(item.content_digest)) && SOURCE_TYPES.includes(item.source_type as typeof SOURCE_TYPES[number]) && typeof item.is_official === "boolean") &&
    integer(value.expected_ref_count, 0, 256) && integer(value.expected_hit_count, 0, 256) && integer(value.required_official_ref_count, 0, 256) && integer(value.required_official_hit_count, 0, 256) && integer(value.first_expected_rank, 0, 8) && typeof value.passed === "boolean";
}

function isFinding(value: unknown): value is RecordDocument { return record(value) && exact(value, FINDING_KEYS) && FINDING_CODES.includes(value.code as typeof FINDING_CODES[number]) && (value.severity === "review_required" || value.severity === "info") && fragmentRefs(value.fragment_refs); }
function isComparison(value: unknown): value is RecordDocument { return record(value) && exact(value, COMPARISON_KEYS) && SAMPLE_ID.test(String(value.sample_id)) && typeof value.baseline_passed === "boolean" && typeof value.candidate_passed === "boolean" && integer(value.baseline_expected_hit_count, 0) && integer(value.candidate_expected_hit_count, 0) && integer(value.baseline_required_official_hit_count, 0) && integer(value.candidate_required_official_hit_count, 0) && integer(value.baseline_first_expected_rank, 0, 8) && integer(value.candidate_first_expected_rank, 0, 8) && (value.change === "improved" || value.change === "unchanged" || value.change === "regressed"); }
function isDatasetBinding(value: unknown): value is RecordDocument { return record(value) && exact(value, ["dataset_id", "dataset_version", "dataset_digest"]) && DATASET_ID.test(String(value.dataset_id)) && integer(value.dataset_version, 1) && DIGEST.test(String(value.dataset_digest)); }
function isProfile(value: unknown): value is RecordDocument { return record(value) && exact(value, PROFILE_KEYS) && value.profile_id === "workflow.rag.lexical-ngram-dev.v1" && value.profile_version === 1 && DIGEST.test(String(value.profile_digest)); }
function isMetrics(value: unknown, minimum: number, maximum: number): value is RecordDocument { return record(value) && exact(value, METRIC_KEYS) && METRIC_KEYS.every((key) => typeof value[key] === "number" && Number.isFinite(value[key]) && Number(value[key]) >= minimum && Number(value[key]) <= maximum); }
function isSnapshotBinding(value: unknown, config: WorkflowRAGEvaluationConfig, applicationId: string): value is RecordDocument { return record(value) && exact(value, SNAPSHOT_KEYS) && value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId && value.application_id === applicationId && SNAPSHOT_ID.test(String(value.snapshot_id)) && integer(value.snapshot_version, 1) && DIGEST.test(String(value.snapshot_digest)) && RAG_REF.test(String(value.rag_ref)); }

function mapResource(value: RecordDocument): WorkflowRAGEvaluationDatasetResource { return { datasetId: String(value.dataset_id), datasetKey: String(value.dataset_key), displayName: String(value.display_name), lifecycleState: value.lifecycle_state as WorkflowRAGEvaluationLifecycle, contentClassification: value.content_classification as WorkflowRAGEvaluationClassification, latestVersion: Number(value.latest_version), latestDigest: String(value.latest_digest), baselineSnapshot: mapSnapshot(value.baseline_snapshot as RecordDocument), sampleCount: Number(value.sample_count), createdAt: String(value.created_at), updatedAt: String(value.updated_at), archivedAt: value.archived_at as string | null }; }
function mapVersion(value: RecordDocument): WorkflowRAGEvaluationDatasetVersion { const dataset = value.dataset as RecordDocument; const review = dataset.review as RecordDocument; return { datasetKey: String(value.dataset_key), displayName: String(value.display_name), lifecycleState: value.lifecycle_state as WorkflowRAGEvaluationLifecycle, createdAt: String(value.created_at), createdByActorRef: String(value.created_by_actor_ref), requestId: String(value.request_id), auditRef: String(value.audit_ref), datasetId: String(dataset.dataset_id), datasetVersion: Number(dataset.dataset_version), datasetDigest: String(dataset.dataset_digest), contentClassification: dataset.content_classification as WorkflowRAGEvaluationClassification, snapshot: mapSnapshot(dataset.snapshot as RecordDocument), thresholds: mapMetrics(dataset.thresholds as RecordDocument), reviewSummary: String(review.review_summary), reviewerRef: String(review.reviewer_ref), reviewedAt: String(review.reviewed_at), samples: (dataset.samples as RecordDocument[]).map(mapSample) }; }
function mapCandidateReview(value: RecordDocument): WorkflowRAGCandidateReviewSummary { const dataset = value.dataset as RecordDocument; const baseline = value.baseline as RecordDocument; const candidate = value.candidate as RecordDocument; return { reviewId: String(value.review_id), datasetId: String(dataset.dataset_id), datasetVersion: Number(dataset.dataset_version), datasetDigest: String(dataset.dataset_digest), baselineSnapshot: mapSnapshot(value.baseline_snapshot as RecordDocument), candidateSnapshot: mapSnapshot(value.candidate_snapshot as RecordDocument), baselineLifecycle: value.baseline_lifecycle as WorkflowRAGEvaluationLifecycle, candidateLifecycle: value.candidate_lifecycle as WorkflowRAGEvaluationLifecycle, conclusion: value.conclusion as typeof CONCLUSIONS[number], baselineStatus: String(baseline.status), candidateStatus: String(candidate.status), baselineMetrics: mapMetrics(baseline.metrics as RecordDocument), candidateMetrics: mapMetrics(candidate.metrics as RecordDocument), metricDelta: mapMetrics(value.metric_delta as RecordDocument), sampleChanges: (value.samples as RecordDocument[]).map((sample) => ({ sampleId: String(sample.sample_id), change: sample.change as "improved" | "unchanged" | "regressed" })), addedFindingCodes: [...value.added_finding_codes as string[]], removedFindingCodes: [...value.removed_finding_codes as string[]], createdAt: String(value.created_at), createdByActorRef: String(value.created_by_actor_ref), requestId: String(value.request_id), auditRef: String(value.audit_ref) }; }
function mapSnapshot(value: RecordDocument): WorkflowRAGEvaluationSnapshotBinding { return { tenantRef: String(value.tenant_ref), workspaceId: String(value.workspace_id), applicationId: String(value.application_id), snapshotId: String(value.snapshot_id), snapshotVersion: Number(value.snapshot_version), snapshotDigest: String(value.snapshot_digest), ragRef: String(value.rag_ref) }; }
function mapMetrics(value: RecordDocument): WorkflowRAGEvaluationMetrics { return { hitAtK: Number(value.hit_at_k), expectedRecallAtK: Number(value.expected_recall_at_k), requiredOfficialRecallAtK: Number(value.required_official_recall_at_k), meanReciprocalRank: Number(value.mean_reciprocal_rank), noEvidenceAccuracy: Number(value.no_evidence_accuracy), samplePassRate: Number(value.sample_pass_rate) }; }
function mapSample(value: RecordDocument): WorkflowRAGEvaluationSampleInput { return { sampleId: String(value.sample_id), queryText: String(value.query_text), expectation: value.expectation as "evidence_required" | "no_evidence", expectedCitationRefs: [...value.expected_citation_refs as string[]], requiredOfficialRefs: [...value.required_official_refs as string[]], topK: Number(value.top_k), reviewNote: String(value.review_note) }; }

function mapCreateBody(config: WorkflowRAGEvaluationConfig, applicationId: string, input: WorkflowRAGEvaluationDatasetInput): RecordDocument { return { workspace_id: config.workspaceId, application_id: applicationId, dataset_key: input.datasetKey.trim(), display_name: input.displayName.trim(), content_classification: input.contentClassification, baseline_snapshot: mapSnapshotBinding(input.baselineSnapshot), thresholds: mapMetricsBody(input.thresholds), review_summary: input.reviewSummary.trim(), samples: input.samples.map((sample) => ({ sample_id: sample.sampleId.trim(), query_text: sample.queryText.trim(), expectation: sample.expectation, expected_citation_refs: sample.expectedCitationRefs, required_official_refs: sample.requiredOfficialRefs, top_k: sample.topK, review_note: sample.reviewNote.trim() })) }; }
function mapSnapshotBinding(value: WorkflowRAGEvaluationSnapshotBinding): RecordDocument { return { tenant_ref: value.tenantRef, workspace_id: value.workspaceId, application_id: value.applicationId, snapshot_id: value.snapshotId, snapshot_version: value.snapshotVersion, snapshot_digest: value.snapshotDigest, rag_ref: value.ragRef }; }
function mapMetricsBody(value: WorkflowRAGEvaluationMetrics): RecordDocument { return { hit_at_k: value.hitAtK, expected_recall_at_k: value.expectedRecallAtK, required_official_recall_at_k: value.requiredOfficialRecallAtK, mean_reciprocal_rank: value.meanReciprocalRank, no_evidence_accuracy: value.noEvidenceAccuracy, sample_pass_rate: value.samplePassRate }; }

function boundaryFor(config: WorkflowRAGEvaluationConfig, applicationId: string, scopes: WorkflowRAGEvaluationScope[]): "offline" | "scope_denied" | "" { if (config.mode === "offline") return "offline"; return !SCOPE_ID.test(applicationId) || scopes.some((scope) => !config.scopes.has(scope)) ? "scope_denied" : ""; }
function headers(config: WorkflowRAGEvaluationConfig, applicationId: string, scopes: WorkflowRAGEvaluationScope[], operation: string) { return buildWorkflowRAGRequestHeaders(config, applicationId, scopes, `rag-evaluation-${operation}`); }
function listBoundary(status: "offline" | "scope_denied"): WorkflowRAGEvaluationListResult { return { status, resources: [], nextCursor: "", failureCode: status === "offline" ? "workflow_rag_evaluation_http_disabled" : "workflow_rag_evaluation_scope_denied", summary: "Evaluation dataset boundary sent zero requests." }; }
function reviewListBoundary(status: "offline" | "scope_denied"): WorkflowRAGCandidateReviewListResult { return { status, reviews: [], nextCursor: "", failureCode: status === "offline" ? "workflow_rag_evaluation_http_disabled" : "workflow_rag_evaluation_scope_denied", summary: "Candidate review boundary sent zero requests." }; }
function operationBoundary(status: "offline" | "scope_denied"): WorkflowRAGEvaluationOperationResult { return { status, resource: null, version: null, review: null, failureCode: status === "offline" ? "workflow_rag_evaluation_http_disabled" : "workflow_rag_evaluation_scope_denied", currentLatestVersion: 0, currentLifecycleState: "", summary: "Evaluation dataset boundary sent zero requests." }; }
function failedList(code = "workflow_rag_evaluation_store_unavailable"): WorkflowRAGEvaluationListResult { return { status: code === "workflow_rag_evaluation_scope_denied" ? "scope_denied" : "failed", resources: [], nextCursor: "", failureCode: code, summary: "Evaluation dataset list failed without fallback." }; }
function failedReviewList(code = "workflow_rag_evaluation_store_unavailable"): WorkflowRAGCandidateReviewListResult { return { status: code === "workflow_rag_evaluation_scope_denied" ? "scope_denied" : "failed", reviews: [], nextCursor: "", failureCode: code, summary: "Candidate review list failed without fallback." }; }
function failedOperation(code = "workflow_rag_evaluation_store_unavailable"): WorkflowRAGEvaluationOperationResult { return { status: code === "workflow_rag_evaluation_scope_denied" ? "scope_denied" : "failed", resource: null, version: null, review: null, failureCode: code, currentLatestVersion: 0, currentLifecycleState: "", summary: "Evaluation dataset operation failed without trusted response or fallback." }; }
async function fetchJSON(input: string, init: RequestInit): Promise<unknown> { const response = await fetch(input, init); return response.json(); }
function isSnapshotBindingInput(value: WorkflowRAGEvaluationSnapshotBinding, config: WorkflowRAGEvaluationConfig, applicationId: string): boolean { return value.tenantRef === config.tenantRef && value.workspaceId === config.workspaceId && value.applicationId === applicationId && SNAPSHOT_ID.test(value.snapshotId) && integer(value.snapshotVersion, 1) && DIGEST.test(value.snapshotDigest) && RAG_REF.test(value.ragRef); }
function validMetricsInput(value: WorkflowRAGEvaluationMetrics): boolean { return Object.values(value).every((metric) => typeof metric === "number" && Number.isFinite(metric) && metric >= 0 && metric <= 1); }
function uniqueFragmentRefs(value: string[]): boolean { return value.length <= 256 && new Set(value).size === value.length && value.every((ref) => FRAGMENT_REF.test(ref)); }
function fragmentRefs(value: unknown): value is string[] { return Array.isArray(value) && uniqueFragmentRefs(value as string[]); }
function findingCodes(value: unknown): value is string[] { return Array.isArray(value) && new Set(value).size === value.length && value.every((code) => FINDING_CODES.includes(code as typeof FINDING_CODES[number])); }
function containsReviewForbidden(value: unknown): boolean { if (typeof value === "string") return containsSecretMaterial(value); if (Array.isArray(value)) return value.some(containsReviewForbidden); if (!record(value)) return false; return Object.entries(value).some(([key, nested]) => REVIEW_FORBIDDEN_KEYS.has(key.toLowerCase()) || containsReviewForbidden(nested)); }
function containsSecretMaterial(value: string): boolean { return /authorization:|bearer\s|api[_-]?key\s*[:=]|x-radishmind-dev-|cookie:|password\s*=|secret\s*=|token\s*=|sk-[a-z0-9]|-----begin private key-----|(?:postgres(?:ql)?|mysql|mongodb):\/\//iu.test(value); }
function integer(value: unknown, minimum: number, maximum = Number.MAX_SAFE_INTEGER): value is number { return typeof value === "number" && Number.isInteger(value) && value >= minimum && value <= maximum; }
function timestamp(value: unknown): value is string { return typeof value === "string" && value.length >= 20 && Number.isFinite(Date.parse(value)); }
function nonEmpty(value: unknown): value is string { return typeof value === "string" && value.trim().length > 0; }
function record(value: unknown): value is RecordDocument { return typeof value === "object" && value !== null && !Array.isArray(value); }
function exact(value: RecordDocument, keys: readonly string[]): boolean { const expected = new Set(keys); return Object.keys(value).length === keys.length && Object.keys(value).every((key) => expected.has(key)); }
function isEvaluationScope(value: string): value is WorkflowRAGEvaluationScope { return value === "workflow_rag_evaluation_datasets:read" || value === "workflow_rag_evaluation_datasets:read_content" || value === "workflow_rag_evaluation_datasets:write" || value === "workflow_rag_evaluation_datasets:review" || value === "workflow_rag_evaluation_datasets:archive" || value === "workflow_rag_snapshots:read"; }
function normalizeBaseUrl(value: string): string { const normalized = value.trim() || DEFAULT_BASE_URL; return normalized.endsWith("/") ? normalized.slice(0, -1) : normalized; }
function normalizeAuthMode(value: string | undefined): WorkflowRAGEvaluationConfig["authMode"] { return value?.trim() === "signed_test_token" || value?.trim() === "radish_oidc_integration_test" ? value.trim() as WorkflowRAGEvaluationConfig["authMode"] : "dev_headers"; }
