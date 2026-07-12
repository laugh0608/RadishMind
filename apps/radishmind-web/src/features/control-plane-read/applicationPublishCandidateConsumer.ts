import type { ApplicationApiProtocol } from "./applicationApiIntegrationConsumer.ts";

const APPLICATION_PUBLISH_SCHEMA_VERSION = "application_publish_candidate.v1";
const DEV_SOURCE = "dev-application-publish-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const FORBIDDEN_RESPONSE_FIELDS = new Set([
  "authorization", "api_key", "key_hash", "secret", "credential", "endpoint", "headers", "cookie",
  "raw_request", "raw_response", "input", "output", "prompt", "messages", "dsn",
]);

export type ApplicationPublishCandidateConfig = {
  mode: "offline" | "dev_application_publish_http";
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  subjectRef: string;
};

export type ApplicationPublishConfiguration = {
  displayName: string;
  description: string;
  applicationKind: string;
  defaultProtocol: ApplicationApiProtocol;
  defaultModel: string;
  allowedProtocols: ApplicationApiProtocol[];
};

export type ApplicationPublishReview = {
  reviewVersion: number;
  decision: ApplicationPublishDecision;
  reason: string;
  state: ApplicationPublishCandidateState;
  reviewedAt: string;
  reviewerRef: string;
  requestId: string;
  auditRef: string;
};

export type ApplicationPublishDecision = "approve" | "reject" | "request_changes" | "withdraw";
export type ApplicationPublishCandidateState = "pending_review" | "approved" | "rejected" | "changes_requested" | "withdrawn";

export type ApplicationPromotionEligibility = {
  eligible: false;
  status: "promotion_blocked";
  blockers: Array<{ code: string; summary: string }>;
};

export type ApplicationPublishCandidate = {
  schemaVersion: typeof APPLICATION_PUBLISH_SCHEMA_VERSION;
  candidateId: string;
  workspaceId: string;
  applicationId: string;
  draftId: string;
  draftVersion: number;
  draftDigest: string;
  baseApplicationUpdatedAt: string;
  configuration: ApplicationPublishConfiguration;
  evidenceRequestIds: string[];
  candidateState: ApplicationPublishCandidateState;
  reviewVersion: number;
  reviews: ApplicationPublishReview[];
  promotionEligibility: ApplicationPromotionEligibility;
  createdAt: string;
  updatedAt: string;
  createdByActorRef: string;
  updatedByActorRef: string;
  requestId: string;
  auditRef: string;
};

export type ApplicationPublishCandidateSummary = {
  candidateId: string;
  applicationId: string;
  draftId: string;
  draftVersion: number;
  draftDigest: string;
  candidateState: ApplicationPublishCandidateState;
  reviewVersion: number;
  promotionStatus: "promotion_blocked";
  promotionBlockers: number;
  createdAt: string;
  updatedAt: string;
  updatedByActorRef: string;
};

export type ApplicationPublishOperationState = {
  status: "offline" | "idle" | "loading" | "creating" | "created" | "loaded" | "reviewing" | "reviewed" | "review_version_conflict" | "immutable_conflict" | "scope_denied" | "failed";
  summary: string;
  failureCode: string;
  currentReviewVersion: number;
  currentCandidateState: string;
  currentDraftVersion: number;
  requestId: string;
  auditRef: string;
};

export type ApplicationPublishCandidateListState = {
  status: "offline" | "idle" | "loading" | "ready" | "empty" | "failed";
  summaries: ApplicationPublishCandidateSummary[];
  failureCode: string;
  summary: string;
};

export type ApplicationPublishWorkspaceMemory = {
  applicationId: string;
  candidateId: string;
  selectedDraftId: string;
  evidenceText: string;
  reviewReason: string;
};

type CandidateDocument = {
  schema_version: string;
  candidate_id: string;
  workspace_id: string;
  application_id: string;
  draft_id: string;
  draft_version: number;
  draft_digest: string;
  base_application_updated_at: string;
  configuration: {
    display_name: string; description: string; application_kind: string; default_protocol: string;
    default_model: string; allowed_protocols: string[];
  };
  evidence_request_ids: string[];
  candidate_state: string;
  review_version: number;
  reviews: Array<{
    review_version: number; decision: string; reason: string; state: string; reviewed_at: string;
    reviewer_ref: string; request_id: string; audit_ref: string;
  }>;
  promotion_eligibility: { eligible: boolean; status: string; blockers: Array<{ code: string; summary: string }> };
  created_at: string; updated_at: string; created_by_actor_ref: string; updated_by_actor_ref: string;
  request_id: string; audit_ref: string;
};

type CandidateEnvelope = {
  request_id: string;
  workspace_id: string;
  application_id: string;
  candidate: CandidateDocument | null;
  failure_code: string | null;
  current_review_version: number;
  current_candidate_state: string;
  current_draft_version: number;
  audit_ref: string;
};

type CandidateListEnvelope = {
  request_id: string;
  workspace_id: string;
  application_id: string;
  candidate_summaries: Array<{
    candidate_id: string; application_id: string; draft_id: string; draft_version: number; draft_digest: string;
    candidate_state: string; review_version: number; promotion_status: string; promotion_blockers: number;
    created_at: string; updated_at: string; updated_by_actor_ref: string;
  }>;
  failure_code: string | null;
  audit_ref: string;
};

export function readApplicationPublishCandidateConfig(): ApplicationPublishCandidateConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  return {
    mode: env.VITE_RADISHMIND_APPLICATION_PUBLISH_SOURCE?.trim() === DEV_SOURCE ? "dev_application_publish_http" : "offline",
    baseUrl: (env.VITE_RADISHMIND_APPLICATION_PUBLISH_BASE_URL ?? env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ?? DEFAULT_BASE_URL).trim().replace(/\/$/u, ""),
    tenantRef: env.VITE_RADISHMIND_DEV_READ_TENANT_REF?.trim() || "tenant_demo",
    workspaceId: env.VITE_RADISHMIND_APPLICATION_PUBLISH_WORKSPACE_ID?.trim() || "workspace_demo",
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || "subject_demo_user",
  };
}

export function initialApplicationPublishOperationState(config: ApplicationPublishCandidateConfig): ApplicationPublishOperationState {
  return {
    status: config.mode === "offline" ? "offline" : "idle",
    summary: config.mode === "offline" ? "Offline publish review stays in component memory and sends no requests." : "Select a saved valid application draft to create a publish candidate.",
    failureCode: "", currentReviewVersion: 0, currentCandidateState: "", currentDraftVersion: 0, requestId: "", auditRef: "",
  };
}

export function initialApplicationPublishListState(config: ApplicationPublishCandidateConfig): ApplicationPublishCandidateListState {
  return config.mode === "offline"
    ? { status: "offline", summaries: [], failureCode: "application_publish_http_disabled", summary: "Offline mode does not load publish candidates." }
    : { status: "idle", summaries: [], failureCode: "", summary: "Load publish candidates for the selected application." };
}

export function createApplicationPublishWorkspaceMemory(applicationId: string, suffix = "new"): ApplicationPublishWorkspaceMemory {
  const normalizedApplicationId = applicationId.trim();
  return { applicationId: normalizedApplicationId, candidateId: `publish-${normalizedApplicationId}-${suffix}`, selectedDraftId: "", evidenceText: "", reviewReason: "" };
}

export function parseApplicationPublishEvidence(raw: string): { requestIds: string[]; failureCode: string } {
  const requestIds = [...new Set(raw.split(/[\s,]+/u).map((value) => value.trim()).filter(Boolean))].sort();
  if (requestIds.length > 20 || requestIds.some((requestId) => !/^[A-Za-z0-9][A-Za-z0-9._:-]{7,159}$/u.test(requestId))) {
    return { requestIds: [], failureCode: "publish_candidate_payload_invalid" };
  }
  if (requestIds.some(containsSecretMaterial)) return { requestIds: [], failureCode: "publish_candidate_secret_material_forbidden" };
  return { requestIds, failureCode: "" };
}

export function validateApplicationPublishReview(decision: ApplicationPublishDecision, reason: string): string {
  if (!(["approve", "reject", "request_changes", "withdraw"] as string[]).includes(decision) || reason.trim().length < 4 || reason.trim().length > 500) return "publish_candidate_payload_invalid";
  return containsSecretMaterial(reason) ? "publish_candidate_secret_material_forbidden" : "";
}

export async function createApplicationPublishCandidate(
  config: ApplicationPublishCandidateConfig,
  applicationId: string,
  candidateId: string,
  draftId: string,
  expectedDraftVersion: number,
  evidenceRequestIds: string[],
): Promise<{ candidate: ApplicationPublishCandidate | null; state: ApplicationPublishOperationState }> {
  if (config.mode === "offline") return { candidate: null, state: initialApplicationPublishOperationState(config) };
  return writeCandidate(config, applicationId, "/v1/user-workspace/application-publish-candidates", {
    candidate_id: candidateId.trim(), draft_id: draftId.trim(), expected_draft_version: expectedDraftVersion,
    evidence_request_ids: evidenceRequestIds,
  }, "created", "write");
}

export async function reviewApplicationPublishCandidate(
  config: ApplicationPublishCandidateConfig,
  applicationId: string,
  candidateId: string,
  expectedReviewVersion: number,
  decision: ApplicationPublishDecision,
  reason: string,
): Promise<{ candidate: ApplicationPublishCandidate | null; state: ApplicationPublishOperationState }> {
  if (config.mode === "offline") return { candidate: null, state: initialApplicationPublishOperationState(config) };
  return writeCandidate(config, applicationId, `/v1/user-workspace/application-publish-candidates/${encodeURIComponent(candidateId)}/reviews`, {
    expected_review_version: expectedReviewVersion, decision, reason: reason.trim(),
  }, "reviewed", "review");
}

export async function readApplicationPublishCandidate(
  config: ApplicationPublishCandidateConfig,
  applicationId: string,
  candidateId: string,
): Promise<{ candidate: ApplicationPublishCandidate | null; state: ApplicationPublishOperationState }> {
  if (config.mode === "offline") return { candidate: null, state: initialApplicationPublishOperationState(config) };
  const requestId = createRequestId("app-publish-read");
  try {
    const response = await fetch(`${config.baseUrl}/v1/user-workspace/application-publish-candidates/${encodeURIComponent(candidateId)}?workspace_id=${encodeURIComponent(config.workspaceId)}&application_id=${encodeURIComponent(applicationId)}`, { headers: candidateHeaders(config, applicationId, requestId, "read") });
    const document: unknown = await response.json();
    if (!response.ok || !isCandidateEnvelope(document, config, applicationId)) throw new Error("invalid publish candidate response");
    return mapCandidateEnvelope(document, "loaded");
  } catch {
    return failedCandidateOperation("publish_candidate_store_unavailable");
  }
}

export async function listApplicationPublishCandidates(config: ApplicationPublishCandidateConfig, applicationId: string): Promise<ApplicationPublishCandidateListState> {
  if (config.mode === "offline") return initialApplicationPublishListState(config);
  const requestId = createRequestId("app-publish-list");
  try {
    const response = await fetch(`${config.baseUrl}/v1/user-workspace/application-publish-candidates?workspace_id=${encodeURIComponent(config.workspaceId)}&application_id=${encodeURIComponent(applicationId)}`, { headers: candidateHeaders(config, applicationId, requestId, "read") });
    const document: unknown = await response.json();
    if (!response.ok || !isCandidateListEnvelope(document, config, applicationId)) throw new Error("invalid publish candidate list response");
    if (document.failure_code) return { status: "failed", summaries: [], failureCode: document.failure_code, summary: "Publish candidates could not be loaded." };
    const summaries = document.candidate_summaries.map(mapCandidateSummary);
    return { status: summaries.length ? "ready" : "empty", summaries, failureCode: "", summary: summaries.length ? `Loaded ${summaries.length} publish candidates.` : "No publish candidates exist for this application." };
  } catch {
    return { status: "failed", summaries: [], failureCode: "publish_candidate_store_unavailable", summary: "Publish candidates could not be loaded." };
  }
}

async function writeCandidate(
  config: ApplicationPublishCandidateConfig,
  applicationId: string,
  path: string,
  body: unknown,
  successStatus: "created" | "reviewed",
  operation: "write" | "review",
): Promise<{ candidate: ApplicationPublishCandidate | null; state: ApplicationPublishOperationState }> {
  const requestId = createRequestId("app-publish-write");
  try {
    const response = await fetch(`${config.baseUrl}${path}`, { method: "POST", headers: { ...candidateHeaders(config, applicationId, requestId, operation), "Content-Type": "application/json" }, body: JSON.stringify(body) });
    const document: unknown = await response.json();
    if (!response.ok || !isCandidateEnvelope(document, config, applicationId)) throw new Error("invalid publish candidate response");
    return mapCandidateEnvelope(document, successStatus);
  } catch {
    return failedCandidateOperation("publish_candidate_store_unavailable");
  }
}

function mapCandidateEnvelope(document: CandidateEnvelope, successStatus: "created" | "loaded" | "reviewed") {
  const status = document.failure_code === "publish_candidate_review_version_conflict" ? "review_version_conflict"
    : document.failure_code === "publish_candidate_immutable_conflict" ? "immutable_conflict"
      : document.failure_code === "publish_candidate_scope_denied" ? "scope_denied"
        : document.failure_code ? "failed" : successStatus;
  return {
    candidate: document.candidate ? mapCandidate(document.candidate) : null,
    state: {
      status,
      summary: document.failure_code
        ? "Publish candidate operation failed without changing stored state."
        : successStatus === "created"
          ? "Immutable publish candidate created."
          : successStatus === "loaded"
            ? "Publish candidate loaded with current eligibility."
            : "Publish candidate review recorded.",
      failureCode: document.failure_code ?? "", currentReviewVersion: document.current_review_version,
      currentCandidateState: document.current_candidate_state, currentDraftVersion: document.current_draft_version,
      requestId: document.request_id, auditRef: document.audit_ref,
    } satisfies ApplicationPublishOperationState,
  };
}

function failedCandidateOperation(failureCode: string) {
  return {
    candidate: null,
    state: { status: "failed", summary: "Publish candidate store is unavailable.", failureCode, currentReviewVersion: 0, currentCandidateState: "", currentDraftVersion: 0, requestId: "", auditRef: "" } satisfies ApplicationPublishOperationState,
  };
}

function candidateHeaders(config: ApplicationPublishCandidateConfig, applicationId: string, requestId: string, operation: "read" | "write" | "review"): Record<string, string> {
  const scope = operation === "read" ? "application_publish_candidates:read" : operation === "review" ? "application_publish_candidates:review" : "application_publish_candidates:write";
  return {
    Accept: "application/json", "X-Request-Id": requestId,
    "X-RadishMind-Dev-Read-Identity": "radishmind-web-application-publish-dev",
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Read-Scopes": scope,
    "X-RadishMind-Dev-Read-Audit": `audit-${requestId}`,
    "X-RadishMind-Dev-Application-Publish-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Application-Publish-Application": applicationId,
  };
}

function mapCandidate(document: CandidateDocument): ApplicationPublishCandidate {
  return {
    schemaVersion: APPLICATION_PUBLISH_SCHEMA_VERSION, candidateId: document.candidate_id,
    workspaceId: document.workspace_id, applicationId: document.application_id, draftId: document.draft_id,
    draftVersion: document.draft_version, draftDigest: document.draft_digest,
    baseApplicationUpdatedAt: document.base_application_updated_at,
    configuration: {
      displayName: document.configuration.display_name, description: document.configuration.description,
      applicationKind: document.configuration.application_kind, defaultProtocol: document.configuration.default_protocol as ApplicationApiProtocol,
      defaultModel: document.configuration.default_model, allowedProtocols: document.configuration.allowed_protocols as ApplicationApiProtocol[],
    },
    evidenceRequestIds: [...document.evidence_request_ids], candidateState: document.candidate_state as ApplicationPublishCandidateState,
    reviewVersion: document.review_version,
    reviews: document.reviews.map((review) => ({ reviewVersion: review.review_version, decision: review.decision as ApplicationPublishDecision, reason: review.reason, state: review.state as ApplicationPublishCandidateState, reviewedAt: review.reviewed_at, reviewerRef: review.reviewer_ref, requestId: review.request_id, auditRef: review.audit_ref })),
    promotionEligibility: { eligible: false, status: "promotion_blocked", blockers: document.promotion_eligibility.blockers.map((blocker) => ({ ...blocker })) },
    createdAt: document.created_at, updatedAt: document.updated_at, createdByActorRef: document.created_by_actor_ref,
    updatedByActorRef: document.updated_by_actor_ref, requestId: document.request_id, auditRef: document.audit_ref,
  };
}

function mapCandidateSummary(document: CandidateListEnvelope["candidate_summaries"][number]): ApplicationPublishCandidateSummary {
  return {
    candidateId: document.candidate_id, applicationId: document.application_id, draftId: document.draft_id,
    draftVersion: document.draft_version, draftDigest: document.draft_digest,
    candidateState: document.candidate_state as ApplicationPublishCandidateState, reviewVersion: document.review_version,
    promotionStatus: "promotion_blocked", promotionBlockers: document.promotion_blockers,
    createdAt: document.created_at, updatedAt: document.updated_at, updatedByActorRef: document.updated_by_actor_ref,
  };
}

function isCandidateEnvelope(value: unknown, config: ApplicationPublishCandidateConfig, applicationId: string): value is CandidateEnvelope {
  return isRecord(value) && !containsForbiddenCandidateResponse(value) && value.workspace_id === config.workspaceId && value.application_id === applicationId &&
    isNonEmptyString(value.request_id) && isNonEmptyString(value.audit_ref) && (value.failure_code === null || isNonEmptyString(value.failure_code)) &&
    Number.isInteger(value.current_review_version) && value.current_review_version >= 0 && typeof value.current_candidate_state === "string" &&
    Number.isInteger(value.current_draft_version) && value.current_draft_version >= 0 &&
    (value.candidate === null || isCandidateDocument(value.candidate, config, applicationId) &&
      value.current_review_version === value.candidate.review_version && value.current_candidate_state === value.candidate.candidate_state);
}

function isCandidateListEnvelope(value: unknown, config: ApplicationPublishCandidateConfig, applicationId: string): value is CandidateListEnvelope {
  return isRecord(value) && !containsForbiddenCandidateResponse(value) && value.workspace_id === config.workspaceId && value.application_id === applicationId &&
    isNonEmptyString(value.request_id) && isNonEmptyString(value.audit_ref) && (value.failure_code === null || isNonEmptyString(value.failure_code)) && Array.isArray(value.candidate_summaries) &&
    value.candidate_summaries.every((summary) => isRecord(summary) && summary.application_id === applicationId && isNonEmptyString(summary.candidate_id) &&
      isNonEmptyString(summary.draft_id) && Number.isInteger(summary.draft_version) && summary.draft_version > 0 && isDigest(summary.draft_digest) &&
      isCandidateState(summary.candidate_state) && Number.isInteger(summary.review_version) && summary.review_version >= 0 && summary.promotion_status === "promotion_blocked" &&
      Number.isInteger(summary.promotion_blockers) && summary.promotion_blockers >= 0 && isNonEmptyString(summary.created_at) &&
      isNonEmptyString(summary.updated_at) && isNonEmptyString(summary.updated_by_actor_ref));
}

function isCandidateDocument(value: unknown, config: ApplicationPublishCandidateConfig, applicationId: string): value is CandidateDocument {
  if (!isRecord(value) || value.schema_version !== APPLICATION_PUBLISH_SCHEMA_VERSION || value.workspace_id !== config.workspaceId || value.application_id !== applicationId ||
    !isNonEmptyString(value.candidate_id) || !isNonEmptyString(value.draft_id) || !Number.isInteger(value.draft_version) || value.draft_version < 1 || !isDigest(value.draft_digest) ||
    !isCandidateState(value.candidate_state) || !Number.isInteger(value.review_version) || value.review_version < 0 || !Array.isArray(value.evidence_request_ids) ||
    !value.evidence_request_ids.every(isEvidenceRequestId) || !isRecord(value.configuration) || !isNonEmptyString(value.configuration.display_name) ||
    typeof value.configuration.description !== "string" || !isNonEmptyString(value.configuration.application_kind) || !isApplicationProtocol(value.configuration.default_protocol) ||
    !isNonEmptyString(value.configuration.default_model) || !Array.isArray(value.configuration.allowed_protocols) || value.configuration.allowed_protocols.length === 0 ||
    !value.configuration.allowed_protocols.every(isApplicationProtocol) || !value.configuration.allowed_protocols.includes(value.configuration.default_protocol) ||
    !Array.isArray(value.reviews) || value.reviews.length !== value.review_version || !isRecord(value.promotion_eligibility) || value.promotion_eligibility.eligible !== false ||
    value.promotion_eligibility.status !== "promotion_blocked" || !Array.isArray(value.promotion_eligibility.blockers) || !isNonEmptyString(value.base_application_updated_at) ||
    !isNonEmptyString(value.created_at) || !isNonEmptyString(value.updated_at) || !isNonEmptyString(value.created_by_actor_ref) ||
    !isNonEmptyString(value.updated_by_actor_ref) || !isNonEmptyString(value.request_id) || !isNonEmptyString(value.audit_ref)) return false;
  return value.reviews.every((review, index) => isRecord(review) && review.review_version === index + 1 && isApplicationPublishDecision(review.decision) &&
    isNonEmptyString(review.reason) && isCandidateState(review.state) && isNonEmptyString(review.reviewed_at) && isNonEmptyString(review.reviewer_ref) &&
    isNonEmptyString(review.request_id) && isNonEmptyString(review.audit_ref)) &&
    value.promotion_eligibility.blockers.every((blocker) => isRecord(blocker) && isNonEmptyString(blocker.code) && isNonEmptyString(blocker.summary));
}

function isCandidateState(value: unknown): value is ApplicationPublishCandidateState {
  return value === "pending_review" || value === "approved" || value === "rejected" || value === "changes_requested" || value === "withdrawn";
}

function isApplicationPublishDecision(value: unknown): value is ApplicationPublishDecision {
  return value === "approve" || value === "reject" || value === "request_changes" || value === "withdraw";
}

function isApplicationProtocol(value: unknown): value is ApplicationApiProtocol {
  return value === "chat_completions" || value === "responses" || value === "messages";
}

function isDigest(value: unknown): value is string {
  return typeof value === "string" && /^sha256:[a-f0-9]{64}$/u.test(value);
}

function isEvidenceRequestId(value: unknown): value is string {
  return typeof value === "string" && /^[A-Za-z0-9][A-Za-z0-9._:-]{7,159}$/u.test(value);
}

function isNonEmptyString(value: unknown): value is string {
  return typeof value === "string" && value.trim().length > 0;
}

function containsForbiddenCandidateResponse(value: unknown): boolean {
  if (typeof value === "string") return containsSecretMaterial(value);
  if (Array.isArray(value)) return value.some(containsForbiddenCandidateResponse);
  if (!isRecord(value)) return false;
  return Object.entries(value).some(([key, nested]) => FORBIDDEN_RESPONSE_FIELDS.has(key.toLowerCase()) || containsForbiddenCandidateResponse(nested));
}

function containsSecretMaterial(value: string): boolean {
  return /authorization:|bearer\s|api[_-]?key=|x-radishmind-dev-|sk-[a-z0-9]|postgres(?:ql)?:\/\//iu.test(value.trim());
}

function createRequestId(prefix: string): string {
  return `${prefix}-${Date.now()}-${(globalThis.crypto?.randomUUID?.() ?? Math.random().toString(16).slice(2)).replaceAll("-", "").slice(0, 12)}`;
}

function isRecord(value: unknown): value is Record<string, any> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}
