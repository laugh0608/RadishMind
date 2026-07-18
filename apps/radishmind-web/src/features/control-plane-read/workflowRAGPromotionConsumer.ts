import { buildWorkflowRAGRequestHeaders } from "./workflowRAGSnapshotConsumer.ts";

const DEV_SOURCE = "dev-workflow-rag-promotion-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const COLLECTION_PATH = "/v1/user-workspace/workflow-rag-knowledge-promotion-candidates";
const CANDIDATE_SCHEMA = "workflow_rag_knowledge_promotion_candidate.v1";
const DECISION_SCHEMA = "workflow_rag_knowledge_promotion_decision.v1";
const BINDING_SCHEMA = "workflow_rag_application_binding.v1";
const CANDIDATE_ID = /^wragp_[a-z2-7]{16}$/u;
const DECISION_ID = /^wragpd_[a-z2-7]{16}$/u;
const BINDING_ID = /^wragb_[a-z2-7]{16}$/u;
const DATASET_ID = /^wragd_[a-z0-9_]{3,47}$/u;
const REVIEW_ID = /^wragr_[a-z2-7]{16}$/u;
const SNAPSHOT_ID = /^rags_[a-z2-7]{16}$/u;
const SCOPE_ID = /^[A-Za-z0-9][A-Za-z0-9_.:-]{2,119}$/u;
const REFERENCE = /^[A-Za-z0-9][A-Za-z0-9_.:/-]{2,159}$/u;
const DIGEST = /^sha256:[a-f0-9]{64}$/u;
const RAG_REF = /^workflow\.rag\.[a-z][a-z0-9_]{2,47}\.v[1-9][0-9]*$/u;
const PROFILE_ID = /^[a-z][a-z0-9._-]{2,79}$/u;
const FAILURE = /^(|workflow_rag_[a-z_]{3,80})$/u;
const STATES = ["pending", "deferred", "approved", "rejected", "canceled"] as const;
const DECISIONS = ["approve", "reject", "defer", "cancel"] as const;
const SCOPES = [
  "workflow_rag_promotions:read", "workflow_rag_promotions:write", "workflow_rag_promotions:review",
  "workflow_rag_evaluation_datasets:read", "workflow_rag_snapshots:read", "application_drafts:read",
] as const;
const ENVELOPE_KEYS = ["request_id", "tenant_ref", "workspace_id", "application_id", "candidate", "decisions", "binding", "eligibility", "failure_code", "current_record_version", "current_state", "audit_ref"] as const;
const LIST_KEYS = ["request_id", "tenant_ref", "workspace_id", "application_id", "items", "next_cursor", "failure_code", "audit_ref"] as const;
const CANDIDATE_KEYS = ["schema_version", "candidate_id", "candidate_digest", "tenant_ref", "workspace_id", "application_id", "owner_subject_ref", "evidence", "candidate_state", "record_version", "binding_ref", "created_at", "updated_at", "created_by_actor_ref", "updated_by_actor_ref", "request_id", "audit_ref"] as const;
const SUMMARY_KEYS = ["candidate_id", "dataset", "candidate_review_id", "source_draft", "candidate_state", "record_version", "binding_ref", "eligibility_status", "blocker_count", "created_at", "updated_at"] as const;
const EVIDENCE_KEYS = ["dataset", "candidate_review_id", "baseline_snapshot", "candidate_snapshot", "profile", "source_draft"] as const;
const DATASET_KEYS = ["dataset_id", "dataset_version", "dataset_digest"] as const;
const SNAPSHOT_KEYS = ["tenant_ref", "workspace_id", "application_id", "snapshot_id", "snapshot_version", "snapshot_digest", "rag_ref"] as const;
const PROFILE_KEYS = ["profile_id", "profile_version", "profile_digest"] as const;
const SOURCE_DRAFT_KEYS = ["draft_id", "draft_version", "draft_digest", "base_application_updated_at"] as const;
const BINDING_REF_KEYS = ["binding_id", "binding_version", "binding_digest"] as const;
const DECISION_KEYS = ["schema_version", "decision_id", "candidate_id", "candidate_digest", "decision", "reason", "from_state", "to_state", "before_record_version", "after_record_version", "actor_ref", "occurred_at", "request_id", "audit_ref"] as const;
const BINDING_KEYS = ["schema_version", "binding_id", "binding_version", "binding_digest", "candidate_id", "candidate_digest", "approved_decision_id", "approved_record_version", "tenant_ref", "workspace_id", "application_id", "owner_subject_ref", "evidence", "issued_at", "issued_by_actor_ref", "request_id", "audit_ref"] as const;
const ELIGIBILITY_KEYS = ["eligible", "status", "blockers"] as const;
const FORBIDDEN_KEYS = new Set(["query", "query_text", "fragment", "fragment_content", "content", "samples", "metrics", "model_response", "response", "prompt", "messages", "credential", "secret", "authorization", "endpoint", "headers", "cookie", "dsn", "raw_request", "raw_response"]);

export type WorkflowRAGPromotionScope = typeof SCOPES[number];
export type WorkflowRAGPromotionState = typeof STATES[number];
export type WorkflowRAGPromotionDecision = typeof DECISIONS[number];

export type WorkflowRAGPromotionConfig = {
  mode: "offline" | "dev_workflow_rag_promotion_http";
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  subjectRef: string;
  authMode: "dev_headers" | "signed_test_token" | "radish_oidc_integration_test";
  scopes: ReadonlySet<WorkflowRAGPromotionScope>;
};

export type WorkflowRAGApplicationBindingRef = { bindingId: string; bindingVersion: number; bindingDigest: string };
export type WorkflowRAGPromotionDatasetBinding = { datasetId: string; datasetVersion: number; datasetDigest: string };
export type WorkflowRAGPromotionSnapshotBinding = { tenantRef: string; workspaceId: string; applicationId: string; snapshotId: string; snapshotVersion: number; snapshotDigest: string; ragRef: string };
export type WorkflowRAGPromotionProfileBinding = { profileId: string; profileVersion: number; profileDigest: string };
export type WorkflowRAGPromotionSourceDraft = { draftId: string; draftVersion: number; draftDigest: string; baseApplicationUpdatedAt: string };
export type WorkflowRAGPromotionEvidence = {
  dataset: WorkflowRAGPromotionDatasetBinding;
  candidateReviewId: string;
  baselineSnapshot: WorkflowRAGPromotionSnapshotBinding;
  candidateSnapshot: WorkflowRAGPromotionSnapshotBinding;
  profile: WorkflowRAGPromotionProfileBinding;
  sourceDraft: WorkflowRAGPromotionSourceDraft;
};

export type WorkflowRAGPromotionCandidate = {
  candidateId: string;
  candidateDigest: string;
  applicationId: string;
  evidence: WorkflowRAGPromotionEvidence;
  candidateState: WorkflowRAGPromotionState;
  recordVersion: number;
  bindingRef: WorkflowRAGApplicationBindingRef | null;
  createdAt: string;
  updatedAt: string;
  createdByActorRef: string;
  updatedByActorRef: string;
};

export type WorkflowRAGPromotionDecisionRecord = {
  decisionId: string; decision: WorkflowRAGPromotionDecision; reason: string; fromState: WorkflowRAGPromotionState; toState: WorkflowRAGPromotionState;
  beforeRecordVersion: number; afterRecordVersion: number; actorRef: string; occurredAt: string;
};

export type WorkflowRAGApplicationBinding = WorkflowRAGApplicationBindingRef & {
  candidateId: string; candidateDigest: string; approvedDecisionId: string; approvedRecordVersion: number;
  evidence: WorkflowRAGPromotionEvidence; issuedAt: string; issuedByActorRef: string;
};

export type WorkflowRAGPromotionEligibility = { eligible: boolean; status: "eligible" | "blocked"; blockers: string[] };
export type WorkflowRAGPromotionDetail = {
  candidate: WorkflowRAGPromotionCandidate;
  decisions: WorkflowRAGPromotionDecisionRecord[];
  binding: WorkflowRAGApplicationBinding | null;
  eligibility: WorkflowRAGPromotionEligibility;
};

export type WorkflowRAGPromotionSummary = {
  candidateId: string; dataset: WorkflowRAGPromotionDatasetBinding; candidateReviewId: string; sourceDraft: WorkflowRAGPromotionSourceDraft;
  candidateState: WorkflowRAGPromotionState; recordVersion: number; bindingRef: WorkflowRAGApplicationBindingRef | null;
  eligibilityStatus: "eligible" | "blocked"; blockerCount: number; createdAt: string; updatedAt: string;
};

export type WorkflowRAGPromotionListResult = { status: "offline" | "scope_denied" | "ready" | "empty" | "failed"; summaries: WorkflowRAGPromotionSummary[]; nextCursor: string; failureCode: string; summary: string };
export type WorkflowRAGPromotionOperationResult = { status: "offline" | "scope_denied" | "created" | "loaded" | "decided" | "record_version_conflict" | "failed"; detail: WorkflowRAGPromotionDetail | null; failureCode: string; currentRecordVersion: number; currentState: string; summary: string };

type Document = Record<string, unknown>;

export function readWorkflowRAGPromotionConfig(): WorkflowRAGPromotionConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  return {
    mode: env.VITE_RADISHMIND_WORKFLOW_RAG_PROMOTION_SOURCE?.trim() === DEV_SOURCE ? "dev_workflow_rag_promotion_http" : "offline",
    baseUrl: normalizeBaseUrl(env.VITE_RADISHMIND_WORKFLOW_RAG_PROMOTION_BASE_URL ?? env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ?? DEFAULT_BASE_URL),
    tenantRef: env.VITE_RADISHMIND_DEV_READ_TENANT_REF?.trim() || "tenant_demo",
    workspaceId: env.VITE_RADISHMIND_WORKFLOW_RAG_WORKSPACE_ID?.trim() || "workspace_demo",
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || "subject_demo_user",
    authMode: normalizeAuthMode(env.VITE_RADISHMIND_READ_AUTH_MODE),
    scopes: new Set((env.VITE_RADISHMIND_WORKFLOW_RAG_PROMOTION_SCOPES ?? "").split(",").map((value) => value.trim()).filter(isPromotionScope)),
  };
}

export function initialWorkflowRAGPromotionListResult(config: WorkflowRAGPromotionConfig): WorkflowRAGPromotionListResult {
  return config.mode === "offline" ? listBoundary("offline") : { status: "empty", summaries: [], nextCursor: "", failureCode: "", summary: "Load knowledge promotion candidates for this application." };
}

export function initialWorkflowRAGPromotionOperationResult(config: WorkflowRAGPromotionConfig): WorkflowRAGPromotionOperationResult {
  return config.mode === "offline" ? operationBoundary("offline") : { status: "loaded", detail: null, failureCode: "", currentRecordVersion: 0, currentState: "", summary: "Select evidence and an exact source draft before creating a candidate." };
}

export async function listWorkflowRAGPromotionCandidates(config: WorkflowRAGPromotionConfig, applicationId: string, cursor = ""): Promise<WorkflowRAGPromotionListResult> {
  const boundary = boundaryFor(config, applicationId, ["workflow_rag_promotions:read"]);
  if (boundary) return listBoundary(boundary);
  if (cursor && !validCursor(cursor)) return failedList("workflow_rag_promotion_payload_invalid");
  const query = new URLSearchParams({ workspace_id: config.workspaceId, application_id: applicationId, limit: "100" });
  if (cursor) query.set("cursor", cursor);
  try {
    const value = await fetchDocument(`${config.baseUrl}${COLLECTION_PATH}?${query}`, { headers: headers(config, applicationId, ["workflow_rag_promotions:read"], "promotion-list") });
    if (!isListEnvelope(value, config, applicationId)) return failedList();
    if (value.failure_code) return failedList(String(value.failure_code));
    const summaries = (value.items as Document[]).map(mapSummary);
    return { status: summaries.length ? "ready" : "empty", summaries, nextCursor: typeof value.next_cursor === "string" ? value.next_cursor : "", failureCode: "", summary: `Loaded ${summaries.length} knowledge promotion candidates.` };
  } catch { return failedList(); }
}

export async function readWorkflowRAGPromotionCandidate(config: WorkflowRAGPromotionConfig, applicationId: string, candidateId: string): Promise<WorkflowRAGPromotionOperationResult> {
  const boundary = boundaryFor(config, applicationId, ["workflow_rag_promotions:read"]);
  if (boundary) return operationBoundary(boundary);
  if (!CANDIDATE_ID.test(candidateId)) return failedOperation("workflow_rag_promotion_payload_invalid");
  const query = new URLSearchParams({ workspace_id: config.workspaceId, application_id: applicationId });
  try {
    return mapEnvelope(await fetchDocument(`${config.baseUrl}${COLLECTION_PATH}/${encodeURIComponent(candidateId)}?${query}`, { headers: headers(config, applicationId, ["workflow_rag_promotions:read"], "promotion-read") }), config, applicationId, "loaded");
  } catch { return failedOperation(); }
}

export async function createWorkflowRAGPromotionCandidate(
  config: WorkflowRAGPromotionConfig,
  applicationId: string,
  dataset: WorkflowRAGPromotionDatasetBinding,
  candidateReviewId: string,
  sourceDraft: Pick<WorkflowRAGPromotionSourceDraft, "draftId" | "draftVersion">,
): Promise<WorkflowRAGPromotionOperationResult> {
  const required: WorkflowRAGPromotionScope[] = ["workflow_rag_promotions:write", "workflow_rag_evaluation_datasets:read", "workflow_rag_snapshots:read", "application_drafts:read"];
  const boundary = boundaryFor(config, applicationId, required);
  if (boundary) return operationBoundary(boundary);
  if (!isDatasetInput(dataset) || !REVIEW_ID.test(candidateReviewId) || !REFERENCE.test(sourceDraft.draftId) || !integer(sourceDraft.draftVersion, 1)) return failedOperation("workflow_rag_promotion_payload_invalid");
  return write(config, applicationId, COLLECTION_PATH, required, "promotion-create", {
    workspace_id: config.workspaceId, application_id: applicationId,
    dataset_id: dataset.datasetId, dataset_version: dataset.datasetVersion, dataset_digest: dataset.datasetDigest,
    candidate_review_id: candidateReviewId, draft_id: sourceDraft.draftId, expected_draft_version: sourceDraft.draftVersion,
  }, "created");
}

export async function decideWorkflowRAGPromotionCandidate(
  config: WorkflowRAGPromotionConfig,
  applicationId: string,
  candidateId: string,
  expectedRecordVersion: number,
  decision: WorkflowRAGPromotionDecision,
  reason: string,
): Promise<WorkflowRAGPromotionOperationResult> {
  const required: WorkflowRAGPromotionScope[] = ["workflow_rag_promotions:review"];
  const boundary = boundaryFor(config, applicationId, required);
  if (boundary) return operationBoundary(boundary);
  const normalizedReason = reason.trim();
  if (!CANDIDATE_ID.test(candidateId) || !integer(expectedRecordVersion, 1) || !DECISIONS.includes(decision) || normalizedReason.length < 4 || normalizedReason.length > 500 || containsSecretMaterial(normalizedReason)) return failedOperation(containsSecretMaterial(normalizedReason) ? "workflow_rag_promotion_secret_material_forbidden" : "workflow_rag_promotion_payload_invalid");
  return write(config, applicationId, `${COLLECTION_PATH}/${encodeURIComponent(candidateId)}/decisions`, required, "promotion-decision", {
    workspace_id: config.workspaceId, application_id: applicationId, expected_record_version: expectedRecordVersion, decision, reason: normalizedReason,
  }, "decided");
}

export function workflowRAGPromotionDecisionAllowed(state: WorkflowRAGPromotionState, decision: WorkflowRAGPromotionDecision): boolean {
  if (state === "pending") return true;
  if (state === "deferred") return decision !== "defer";
  return state === "approved" && decision === "cancel";
}

async function write(config: WorkflowRAGPromotionConfig, applicationId: string, path: string, scopes: WorkflowRAGPromotionScope[], operation: string, body: Document, success: "created" | "decided") {
  try {
    const value = await fetchDocument(`${config.baseUrl}${path}`, { method: "POST", headers: { ...headers(config, applicationId, scopes, operation), "Content-Type": "application/json" }, body: JSON.stringify(body) });
    return mapEnvelope(value, config, applicationId, success);
  } catch { return failedOperation(); }
}

function mapEnvelope(value: unknown, config: WorkflowRAGPromotionConfig, applicationId: string, success: "created" | "loaded" | "decided"): WorkflowRAGPromotionOperationResult {
  if (!isEnvelope(value, config, applicationId)) return failedOperation();
  const failureCode = typeof value.failure_code === "string" ? value.failure_code : "";
  const status = failureCode === "workflow_rag_promotion_record_version_conflict" ? "record_version_conflict" : failureCode ? "failed" : success;
  const detail = value.candidate === null ? null : {
    candidate: mapCandidate(value.candidate as Document), decisions: (value.decisions as Document[]).map(mapDecision),
    binding: value.binding === null ? null : mapBinding(value.binding as Document), eligibility: mapEligibility(value.eligibility as Document),
  };
  return { status, detail, failureCode, currentRecordVersion: value.current_record_version as number, currentState: value.current_state as string, summary: failureCode ? `Knowledge promotion operation failed: ${failureCode}.` : success === "created" ? "Knowledge promotion candidate created for explicit human review." : success === "decided" ? "Append-only promotion decision recorded." : "Knowledge promotion evidence and current eligibility loaded." };
}

function isEnvelope(value: unknown, config: WorkflowRAGPromotionConfig, applicationId: string): value is Document {
  if (!isRecord(value) || containsForbiddenMaterial(value) || !hasOnlyKeys(value, ENVELOPE_KEYS) || !scopeMatches(value, config, applicationId) || !nullableFailure(value.failure_code) || !integer(value.current_record_version, 0) || typeof value.current_state !== "string" || !isEligibility(value.eligibility) || !Array.isArray(value.decisions) || !value.decisions.every(isDecision)) return false;
  if (value.candidate === null) return value.binding === null && value.decisions.length === 0;
  if (!isCandidate(value.candidate, config, applicationId) || !(value.binding === null || isBinding(value.binding, config, applicationId))) return false;
  return value.decisions.every((decision) => (decision as Document).candidate_id === (value.candidate as Document).candidate_id) && (value.binding === null || (value.binding as Document).candidate_id === (value.candidate as Document).candidate_id);
}

function isListEnvelope(value: unknown, config: WorkflowRAGPromotionConfig, applicationId: string): value is Document {
  return isRecord(value) && !containsForbiddenMaterial(value) && hasOnlyKeys(value, LIST_KEYS) && scopeMatches(value, config, applicationId) && nullableFailure(value.failure_code) && (value.next_cursor === null || typeof value.next_cursor === "string" && validCursor(value.next_cursor)) && Array.isArray(value.items) && value.items.every(isSummary);
}

function isCandidate(value: unknown, config: WorkflowRAGPromotionConfig, applicationId: string): value is Document {
  return isRecord(value) && hasOnlyKeys(value, CANDIDATE_KEYS) && value.schema_version === CANDIDATE_SCHEMA && CANDIDATE_ID.test(String(value.candidate_id)) && DIGEST.test(String(value.candidate_digest)) && scopeMatches(value, config, applicationId) && value.owner_subject_ref === config.subjectRef && isEvidence(value.evidence, config, applicationId) && STATES.includes(value.candidate_state as WorkflowRAGPromotionState) && integer(value.record_version, 1) && (value.binding_ref === null || isBindingRef(value.binding_ref)) && timestamps(value.created_at, value.updated_at) && isReference(value.created_by_actor_ref) && isReference(value.updated_by_actor_ref) && isReference(value.request_id) && isReference(value.audit_ref);
}

function isSummary(value: unknown): value is Document {
  return isRecord(value) && hasOnlyKeys(value, SUMMARY_KEYS) && CANDIDATE_ID.test(String(value.candidate_id)) && isDataset(value.dataset) && REVIEW_ID.test(String(value.candidate_review_id)) && isSourceDraft(value.source_draft) && STATES.includes(value.candidate_state as WorkflowRAGPromotionState) && integer(value.record_version, 1) && (value.binding_ref === null || isBindingRef(value.binding_ref)) && (value.eligibility_status === "eligible" || value.eligibility_status === "blocked") && integer(value.blocker_count, 0) && timestamps(value.created_at, value.updated_at);
}

function isEvidence(value: unknown, config: WorkflowRAGPromotionConfig, applicationId: string): value is Document {
  return isRecord(value) && hasOnlyKeys(value, EVIDENCE_KEYS) && isDataset(value.dataset) && REVIEW_ID.test(String(value.candidate_review_id)) && isSnapshot(value.baseline_snapshot, config, applicationId) && isSnapshot(value.candidate_snapshot, config, applicationId) && isProfile(value.profile) && isSourceDraft(value.source_draft);
}

function isDataset(value: unknown): value is Document { return isRecord(value) && hasOnlyKeys(value, DATASET_KEYS) && DATASET_ID.test(String(value.dataset_id)) && integer(value.dataset_version, 1) && DIGEST.test(String(value.dataset_digest)); }
function isDatasetInput(value: WorkflowRAGPromotionDatasetBinding): boolean { return DATASET_ID.test(value.datasetId) && integer(value.datasetVersion, 1) && DIGEST.test(value.datasetDigest); }
function isSnapshot(value: unknown, config: WorkflowRAGPromotionConfig, applicationId: string): value is Document { return isRecord(value) && hasOnlyKeys(value, SNAPSHOT_KEYS) && value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId && value.application_id === applicationId && SNAPSHOT_ID.test(String(value.snapshot_id)) && integer(value.snapshot_version, 1) && DIGEST.test(String(value.snapshot_digest)) && RAG_REF.test(String(value.rag_ref)); }
function isProfile(value: unknown): value is Document { return isRecord(value) && hasOnlyKeys(value, PROFILE_KEYS) && PROFILE_ID.test(String(value.profile_id)) && integer(value.profile_version, 1) && DIGEST.test(String(value.profile_digest)); }
function isSourceDraft(value: unknown): value is Document { return isRecord(value) && hasOnlyKeys(value, SOURCE_DRAFT_KEYS) && REFERENCE.test(String(value.draft_id)) && integer(value.draft_version, 1) && DIGEST.test(String(value.draft_digest)) && isTimestamp(value.base_application_updated_at); }
function isBindingRef(value: unknown): value is Document { return isRecord(value) && hasOnlyKeys(value, BINDING_REF_KEYS) && BINDING_ID.test(String(value.binding_id)) && value.binding_version === 1 && DIGEST.test(String(value.binding_digest)); }
function isDecision(value: unknown): value is Document { return isRecord(value) && hasOnlyKeys(value, DECISION_KEYS) && value.schema_version === DECISION_SCHEMA && DECISION_ID.test(String(value.decision_id)) && CANDIDATE_ID.test(String(value.candidate_id)) && DIGEST.test(String(value.candidate_digest)) && DECISIONS.includes(value.decision as WorkflowRAGPromotionDecision) && typeof value.reason === "string" && value.reason.trim().length >= 4 && value.reason.length <= 500 && !containsSecretMaterial(value.reason) && STATES.includes(value.from_state as WorkflowRAGPromotionState) && STATES.includes(value.to_state as WorkflowRAGPromotionState) && integer(value.before_record_version, 1) && value.after_record_version === (value.before_record_version as number) + 1 && isReference(value.actor_ref) && isTimestamp(value.occurred_at) && isReference(value.request_id) && isReference(value.audit_ref); }
function isBinding(value: unknown, config: WorkflowRAGPromotionConfig, applicationId: string): value is Document { return isRecord(value) && hasOnlyKeys(value, BINDING_KEYS) && value.schema_version === BINDING_SCHEMA && isBindingRef(pickBindingRef(value)) && CANDIDATE_ID.test(String(value.candidate_id)) && DIGEST.test(String(value.candidate_digest)) && DECISION_ID.test(String(value.approved_decision_id)) && integer(value.approved_record_version, 2) && scopeMatches(value, config, applicationId) && value.owner_subject_ref === config.subjectRef && isEvidence(value.evidence, config, applicationId) && isTimestamp(value.issued_at) && isReference(value.issued_by_actor_ref) && isReference(value.request_id) && isReference(value.audit_ref); }
function isEligibility(value: unknown): value is Document { return isRecord(value) && hasOnlyKeys(value, ELIGIBILITY_KEYS) && typeof value.eligible === "boolean" && (value.status === "eligible" || value.status === "blocked") && value.eligible === (value.status === "eligible") && Array.isArray(value.blockers) && value.blockers.every((blocker) => isRecord(blocker) && hasOnlyKeys(blocker, ["code"]) && FAILURE.test(String(blocker.code)) && String(blocker.code).length > 0); }

function mapCandidate(value: Document): WorkflowRAGPromotionCandidate { return { candidateId: value.candidate_id as string, candidateDigest: value.candidate_digest as string, applicationId: value.application_id as string, evidence: mapEvidence(value.evidence as Document), candidateState: value.candidate_state as WorkflowRAGPromotionState, recordVersion: value.record_version as number, bindingRef: value.binding_ref === null ? null : mapBindingRef(value.binding_ref as Document), createdAt: value.created_at as string, updatedAt: value.updated_at as string, createdByActorRef: value.created_by_actor_ref as string, updatedByActorRef: value.updated_by_actor_ref as string }; }
function mapSummary(value: Document): WorkflowRAGPromotionSummary { return { candidateId: value.candidate_id as string, dataset: mapDataset(value.dataset as Document), candidateReviewId: value.candidate_review_id as string, sourceDraft: mapSourceDraft(value.source_draft as Document), candidateState: value.candidate_state as WorkflowRAGPromotionState, recordVersion: value.record_version as number, bindingRef: value.binding_ref === null ? null : mapBindingRef(value.binding_ref as Document), eligibilityStatus: value.eligibility_status as "eligible" | "blocked", blockerCount: value.blocker_count as number, createdAt: value.created_at as string, updatedAt: value.updated_at as string }; }
function mapEvidence(value: Document): WorkflowRAGPromotionEvidence { return { dataset: mapDataset(value.dataset as Document), candidateReviewId: value.candidate_review_id as string, baselineSnapshot: mapSnapshot(value.baseline_snapshot as Document), candidateSnapshot: mapSnapshot(value.candidate_snapshot as Document), profile: { profileId: (value.profile as Document).profile_id as string, profileVersion: (value.profile as Document).profile_version as number, profileDigest: (value.profile as Document).profile_digest as string }, sourceDraft: mapSourceDraft(value.source_draft as Document) }; }
function mapDataset(value: Document): WorkflowRAGPromotionDatasetBinding { return { datasetId: value.dataset_id as string, datasetVersion: value.dataset_version as number, datasetDigest: value.dataset_digest as string }; }
function mapSnapshot(value: Document): WorkflowRAGPromotionSnapshotBinding { return { tenantRef: value.tenant_ref as string, workspaceId: value.workspace_id as string, applicationId: value.application_id as string, snapshotId: value.snapshot_id as string, snapshotVersion: value.snapshot_version as number, snapshotDigest: value.snapshot_digest as string, ragRef: value.rag_ref as string }; }
function mapSourceDraft(value: Document): WorkflowRAGPromotionSourceDraft { return { draftId: value.draft_id as string, draftVersion: value.draft_version as number, draftDigest: value.draft_digest as string, baseApplicationUpdatedAt: value.base_application_updated_at as string }; }
function mapBindingRef(value: Document): WorkflowRAGApplicationBindingRef { return { bindingId: value.binding_id as string, bindingVersion: value.binding_version as number, bindingDigest: value.binding_digest as string }; }
function mapDecision(value: Document): WorkflowRAGPromotionDecisionRecord { return { decisionId: value.decision_id as string, decision: value.decision as WorkflowRAGPromotionDecision, reason: value.reason as string, fromState: value.from_state as WorkflowRAGPromotionState, toState: value.to_state as WorkflowRAGPromotionState, beforeRecordVersion: value.before_record_version as number, afterRecordVersion: value.after_record_version as number, actorRef: value.actor_ref as string, occurredAt: value.occurred_at as string }; }
function mapBinding(value: Document): WorkflowRAGApplicationBinding { return { ...mapBindingRef(value), candidateId: value.candidate_id as string, candidateDigest: value.candidate_digest as string, approvedDecisionId: value.approved_decision_id as string, approvedRecordVersion: value.approved_record_version as number, evidence: mapEvidence(value.evidence as Document), issuedAt: value.issued_at as string, issuedByActorRef: value.issued_by_actor_ref as string }; }
function mapEligibility(value: Document): WorkflowRAGPromotionEligibility { return { eligible: value.eligible as boolean, status: value.status as "eligible" | "blocked", blockers: (value.blockers as Document[]).map((blocker) => blocker.code as string) }; }

function headers(config: WorkflowRAGPromotionConfig, applicationId: string, scopes: WorkflowRAGPromotionScope[], operation: string) { return buildWorkflowRAGRequestHeaders(config, applicationId, scopes, operation); }
function boundaryFor(config: WorkflowRAGPromotionConfig, applicationId: string, scopes: WorkflowRAGPromotionScope[]): "offline" | "scope_denied" | "" { if (config.mode === "offline") return "offline"; if (!SCOPE_ID.test(applicationId) || scopes.some((scope) => !config.scopes.has(scope))) return "scope_denied"; return ""; }
function listBoundary(boundary: "offline" | "scope_denied"): WorkflowRAGPromotionListResult { return { status: boundary, summaries: [], nextCursor: "", failureCode: boundary === "offline" ? "workflow_rag_promotion_http_disabled" : "workflow_rag_promotion_scope_denied", summary: boundary === "offline" ? "Offline mode sends zero knowledge promotion requests." : "Required promotion scopes are unavailable; zero requests were sent." }; }
function operationBoundary(boundary: "offline" | "scope_denied"): WorkflowRAGPromotionOperationResult { return { status: boundary, detail: null, failureCode: boundary === "offline" ? "workflow_rag_promotion_http_disabled" : "workflow_rag_promotion_scope_denied", currentRecordVersion: 0, currentState: "", summary: boundary === "offline" ? "Offline mode sends zero knowledge promotion requests." : "Required promotion scopes are unavailable; zero requests were sent." }; }
function failedList(failureCode = "workflow_rag_promotion_store_unavailable"): WorkflowRAGPromotionListResult { return { status: "failed", summaries: [], nextCursor: "", failureCode: FAILURE.test(failureCode) ? failureCode : "workflow_rag_promotion_store_unavailable", summary: "Knowledge promotion list failed without trusted response or fallback." }; }
function failedOperation(failureCode = "workflow_rag_promotion_store_unavailable"): WorkflowRAGPromotionOperationResult { return { status: "failed", detail: null, failureCode: FAILURE.test(failureCode) ? failureCode : "workflow_rag_promotion_store_unavailable", currentRecordVersion: 0, currentState: "", summary: "Knowledge promotion operation failed without trusted response or fallback." }; }
async function fetchDocument(input: string, init: RequestInit): Promise<unknown> { const response = await fetch(input, init); if (!response.ok) throw new Error("knowledge promotion request failed"); return response.json(); }
function pickBindingRef(value: Document): Document { return { binding_id: value.binding_id, binding_version: value.binding_version, binding_digest: value.binding_digest }; }
function scopeMatches(value: Document, config: WorkflowRAGPromotionConfig, applicationId: string): boolean { return value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId && value.application_id === applicationId; }
function nullableFailure(value: unknown): boolean { return value === null || typeof value === "string" && FAILURE.test(value); }
function validCursor(value: string): boolean { const separator = value.lastIndexOf("|"); return separator > 0 && CANDIDATE_ID.test(value.slice(separator + 1)) && isTimestamp(value.slice(0, separator)); }
function timestamps(...values: unknown[]): boolean { return values.every(isTimestamp); }
function isTimestamp(value: unknown): value is string { return typeof value === "string" && value.length >= 20 && Number.isFinite(Date.parse(value)); }
function isReference(value: unknown): value is string { return typeof value === "string" && REFERENCE.test(value); }
function integer(value: unknown, minimum: number): value is number { return typeof value === "number" && Number.isInteger(value) && value >= minimum; }
function containsForbiddenMaterial(value: unknown): boolean { if (Array.isArray(value)) return value.some(containsForbiddenMaterial); if (!isRecord(value)) return false; return Object.entries(value).some(([key, nested]) => FORBIDDEN_KEYS.has(key.toLowerCase()) || containsForbiddenMaterial(nested)); }
function containsSecretMaterial(value: string): boolean { return /authorization:|bearer\s|api[_-]?key\s*[:=]|x-radishmind-dev-|cookie:|password\s*=|secret\s*=|token\s*=|sk-[a-z0-9]|-----begin private key-----|(?:postgres(?:ql)?|mysql|mongodb):\/\//iu.test(value); }
function isPromotionScope(value: string): value is WorkflowRAGPromotionScope { return SCOPES.includes(value as WorkflowRAGPromotionScope); }
function isRecord(value: unknown): value is Document { return typeof value === "object" && value !== null && !Array.isArray(value); }
function hasOnlyKeys(value: Document, allowed: readonly string[]): boolean { const expected = new Set(allowed); return Object.keys(value).length === expected.size && Object.keys(value).every((key) => expected.has(key)); }
function normalizeBaseUrl(value: string): string { const normalized = value.trim() || DEFAULT_BASE_URL; return normalized.endsWith("/") ? normalized.slice(0, -1) : normalized; }
function normalizeAuthMode(value: string | undefined): WorkflowRAGPromotionConfig["authMode"] { const normalized = value?.trim(); return normalized === "signed_test_token" || normalized === "radish_oidc_integration_test" ? normalized : "dev_headers"; }
