const DEV_SOURCE = "dev-workflow-rag-application-runtime-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const ASSIGNMENT_SCHEMA_VERSION = "workflow_rag_application_runtime_assignment.v1";
const ANSWER_SCHEMA_VERSION = "workflow_rag_application_answer.v1";
const APPLICATION_ID_PATTERN = /^app_[a-z0-9]{16}$/u;
const ASSIGNMENT_ID_PATTERN = /^wragra_[a-z2-7]{16}$/u;
const API_KEY_ID_PATTERN = /^key_[a-z2-7]{16}$/u;
const TOKEN_PATTERN = /^rmd_dev_key_[a-z2-7]{16}\.[A-Za-z0-9_-]{43}$/u;
const BINDING_ID_PATTERN = /^wragb_[a-z2-7]{16}$/u;
const RUN_ID_PATTERN = /^run_[a-z0-9]{16,64}$/u;
const REF_PATTERN = /^[A-Za-z0-9][A-Za-z0-9_.:/-]{2,159}$/u;
const FRAGMENT_REF_PATTERN = /^[a-z][a-z0-9_]{2,63}$/u;
const DIGEST_PATTERN = /^sha256:[a-f0-9]{64}$/u;
const FORBIDDEN_RESPONSE_FIELDS = new Set([
  "authorization", "credential", "token", "secret", "headers", "cookie", "dsn", "endpoint", "url", "uri",
  "raw_request", "raw_response", "prompt", "messages", "provider_raw_envelope", "fragment_content", "content",
]);

export type WorkflowRAGApplicationRuntimeConfig = {
  mode: "offline" | "dev_workflow_rag_application_runtime_http";
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  subjectRef: string;
};

export type WorkflowRAGApplicationRuntimeAssignment = {
  schemaVersion: typeof ASSIGNMENT_SCHEMA_VERSION;
  assignmentId: string;
  recordVersion: number;
  assignmentDigest: string;
  tenantRef: string;
  workspaceId: string;
  applicationId: string;
  ownerSubjectRef: string;
  state: "active" | "revoked";
  publishCandidateId: string;
  publishReviewVersion: number;
  draftId: string;
  draftVersion: number;
  draftDigest: string;
  bindingRef: { bindingId: string; bindingVersion: 1; bindingDigest: string };
  createdAt: string;
  updatedAt: string;
  updatedByActorRef: string;
  requestId: string;
  auditRef: string;
};

export type WorkflowRAGApplicationRuntimeResult = {
  status: "offline" | "ready" | "not_found" | "version_conflict" | "failed";
  assignment: WorkflowRAGApplicationRuntimeAssignment | null;
  failureCode: string;
  currentRecordVersion: number;
  currentState: string;
  requestId: string;
  auditRef: string;
  summary: string;
};

export type WorkflowRAGApplicationAnswer = {
  schemaVersion: typeof ANSWER_SCHEMA_VERSION;
  answer: string;
  citations: Array<{ fragmentRef: string; claimSummary: string }>;
  limitations: string[];
  confidence: "low" | "medium" | "high";
};

export type WorkflowRAGApplicationInvocationResult = {
  status: "offline" | "succeeded" | "failed";
  runId: string;
  runStatus: "" | "running" | "succeeded" | "failed" | "canceled";
  answer: WorkflowRAGApplicationAnswer | null;
  failureCode: string;
  failureSummary: string;
  requestId: string;
  auditRef: string;
  summary: string;
};

type AssignmentDocument = {
  schema_version: string;
  assignment_id: string;
  record_version: number;
  assignment_digest: string;
  tenant_ref: string;
  workspace_id: string;
  application_id: string;
  owner_subject_ref: string;
  state: string;
  publish_candidate_id: string;
  publish_review_version: number;
  publish_candidate_state: string;
  draft_id: string;
  draft_version: number;
  draft_digest: string;
  binding_ref: { binding_id: string; binding_version: number; binding_digest: string };
  created_at: string;
  updated_at: string;
  created_by_actor_ref: string;
  updated_by_actor_ref: string;
  request_id: string;
  audit_ref: string;
};

type RuntimeEnvelope = {
  request_id: string;
  tenant_ref: string;
  workspace_id: string;
  application_id: string;
  assignment: AssignmentDocument | null;
  events: unknown[];
  audits: unknown[];
  failure_code: string | null;
  current_record_version: number;
  current_state: string;
  audit_ref: string;
};

type AnswerDocument = {
  schema_version: string;
  answer: string;
  citations: Array<{ fragment_ref: string; claim_summary: string }>;
  limitations: string[];
  confidence: string;
};

type InvocationEnvelope = {
  request_id: string;
  tenant_ref: string;
  workspace_id: string;
  application_id: string;
  run: Record<string, unknown> | null;
  answer: AnswerDocument | null;
  failure_code: string | null;
  failure_summary: string;
  audit_ref: string;
};

export function readWorkflowRAGApplicationRuntimeConfig(): WorkflowRAGApplicationRuntimeConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  return {
    mode: env.VITE_RADISHMIND_WORKFLOW_RAG_APPLICATION_RUNTIME_SOURCE?.trim() === DEV_SOURCE
      ? "dev_workflow_rag_application_runtime_http"
      : "offline",
    baseUrl: normalizeBaseUrl(
      env.VITE_RADISHMIND_WORKFLOW_RAG_APPLICATION_RUNTIME_BASE_URL ??
      env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ??
      DEFAULT_BASE_URL,
    ),
    tenantRef: env.VITE_RADISHMIND_DEV_READ_TENANT_REF?.trim() || "tenant_demo",
    workspaceId: env.VITE_RADISHMIND_WORKFLOW_RAG_APPLICATION_RUNTIME_WORKSPACE_ID?.trim() || "workspace_demo",
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || "subject_demo_user",
  };
}

export function initialWorkflowRAGApplicationRuntimeResult(
  config: WorkflowRAGApplicationRuntimeConfig,
): WorkflowRAGApplicationRuntimeResult {
  return config.mode === "offline"
    ? { status: "offline", assignment: null, failureCode: "workflow_rag_application_runtime_http_disabled", currentRecordVersion: 0, currentState: "", requestId: "", auditRef: "", summary: "Offline mode sends no runtime assignment requests." }
    : { status: "not_found", assignment: null, failureCode: "", currentRecordVersion: 0, currentState: "", requestId: "", auditRef: "", summary: "Load the current application RAG runtime assignment." };
}

export function initialWorkflowRAGApplicationInvocationResult(
  config: WorkflowRAGApplicationRuntimeConfig,
): WorkflowRAGApplicationInvocationResult {
  return config.mode === "offline"
    ? { status: "offline", runId: "", runStatus: "", answer: null, failureCode: "workflow_rag_application_runtime_http_disabled", failureSummary: "", requestId: "", auditRef: "", summary: "Offline mode sends no application RAG invocation requests." }
    : { status: "failed", runId: "", runStatus: "", answer: null, failureCode: "", failureSummary: "", requestId: "", auditRef: "", summary: "Use a one-time API key with application_rag:invoke to call the active assignment." };
}

export async function readWorkflowRAGApplicationRuntimeAssignment(
  config: WorkflowRAGApplicationRuntimeConfig,
  applicationId: string,
): Promise<WorkflowRAGApplicationRuntimeResult> {
  if (config.mode === "offline") return initialWorkflowRAGApplicationRuntimeResult(config);
  if (!APPLICATION_ID_PATTERN.test(applicationId)) return failedRuntime("workflow_rag_runtime_payload_invalid");
  const requestId = createRequestId("workflow-rag-runtime-read");
  const query = new URLSearchParams({ workspace_id: config.workspaceId });
  return fetchRuntime(config, applicationId, requestId,
    `${config.baseUrl}/v1/user-workspace/applications/${encodeURIComponent(applicationId)}/workflow-rag-runtime-assignment?${query}`,
    { headers: managementHeaders(config, applicationId, requestId, "read") });
}

export async function decideWorkflowRAGApplicationRuntimeAssignment(
  config: WorkflowRAGApplicationRuntimeConfig,
  input: {
    applicationId: string;
    expectedRecordVersion: number;
    decision: "activate" | "replace" | "revoke";
    publishCandidateId: string;
    reason: string;
  },
): Promise<WorkflowRAGApplicationRuntimeResult> {
  if (config.mode === "offline") return initialWorkflowRAGApplicationRuntimeResult(config);
  const reason = input.reason.trim();
  if (!APPLICATION_ID_PATTERN.test(input.applicationId) || !Number.isInteger(input.expectedRecordVersion) ||
    input.expectedRecordVersion < 0 || reason.length < 4 || reason.length > 500 || containsSensitiveText(reason) ||
    !(["activate", "replace", "revoke"] as string[]).includes(input.decision) ||
    (input.decision !== "revoke" && !REF_PATTERN.test(input.publishCandidateId)) ||
    (input.decision === "revoke" && input.publishCandidateId.trim() !== "")) {
    return failedRuntime(containsSensitiveText(reason)
      ? "workflow_rag_runtime_secret_material_forbidden"
      : "workflow_rag_runtime_payload_invalid");
  }
  const requestId = createRequestId("workflow-rag-runtime-decision");
  const body = {
    workspace_id: config.workspaceId,
    expected_record_version: input.expectedRecordVersion,
    decision: input.decision,
    publish_candidate_id: input.publishCandidateId.trim(),
    reason,
  };
  return fetchRuntime(config, input.applicationId, requestId,
    `${config.baseUrl}/v1/user-workspace/applications/${encodeURIComponent(input.applicationId)}/workflow-rag-runtime-assignment/decisions`,
    { method: "POST", headers: { ...managementHeaders(config, input.applicationId, requestId, "write"), "Content-Type": "application/json" }, body: JSON.stringify(body) });
}

export async function invokeWorkflowRAGApplication(
  config: WorkflowRAGApplicationRuntimeConfig,
  input: { applicationId: string; apiKeyId: string; token: string; text: string },
): Promise<WorkflowRAGApplicationInvocationResult> {
  if (config.mode === "offline") return initialWorkflowRAGApplicationInvocationResult(config);
  const text = input.text.trim();
  if (!APPLICATION_ID_PATTERN.test(input.applicationId) || !API_KEY_ID_PATTERN.test(input.apiKeyId) ||
    !TOKEN_PATTERN.test(input.token) || !input.token.includes(input.apiKeyId) || text.length === 0 ||
    new TextEncoder().encode(text).length > 4096 || containsSensitiveText(text)) {
    return failedInvocation(containsSensitiveText(text)
      ? "workflow_rag_runtime_secret_material_forbidden"
      : "workflow_rag_runtime_payload_invalid");
  }
  const requestId = createRequestId("workflow-rag-application-invocation");
  try {
    const response = await fetch(`${config.baseUrl}/v1/application-rag/invocations`, {
      method: "POST",
      headers: { "Authorization": `Bearer ${input.token}`, "Content-Type": "application/json", "X-Request-Id": requestId },
      body: JSON.stringify({ input: text }),
    });
    const document: unknown = await response.json();
    if (!isInvocationEnvelope(document, config, input.applicationId)) return failedInvocation("workflow_rag_runtime_store_unavailable");
    if (document.failure_code) {
      return {
        status: "failed", runId: document.run && typeof document.run.run_id === "string" ? document.run.run_id : "",
        runStatus: document.run && isRunStatus(document.run.status) ? document.run.status : "", answer: null,
        failureCode: document.failure_code, failureSummary: document.failure_summary, requestId: document.request_id,
        auditRef: document.audit_ref, summary: document.failure_summary || document.failure_code,
      };
    }
    const answer = parseAnswer(document.answer);
    const runId = document.run && typeof document.run.run_id === "string" ? document.run.run_id : "";
    if (!response.ok || !answer || !RUN_ID_PATTERN.test(runId) || document.run?.schema_version !== "workflow_run_record.v4" || document.run.status !== "succeeded") {
      return failedInvocation("workflow_rag_runtime_store_unavailable");
    }
    return { status: "succeeded", runId, runStatus: "succeeded", answer, failureCode: "", failureSummary: "", requestId: document.request_id, auditRef: document.audit_ref, summary: "Application RAG invocation succeeded. The answer remains in current component memory only." };
  } catch {
    return failedInvocation("workflow_rag_runtime_store_unavailable");
  }
}

async function fetchRuntime(
  config: WorkflowRAGApplicationRuntimeConfig,
  applicationId: string,
  requestId: string,
  url: string,
  init: RequestInit,
): Promise<WorkflowRAGApplicationRuntimeResult> {
  try {
    const response = await fetch(url, init);
    const document: unknown = await response.json();
    if (!response.ok || !isRuntimeEnvelope(document, config, applicationId)) return failedRuntime("workflow_rag_runtime_store_unavailable");
    if (document.failure_code) {
      return {
        status: document.failure_code === "workflow_rag_runtime_assignment_not_found" ? "not_found" :
          document.failure_code === "workflow_rag_runtime_assignment_version_conflict" ? "version_conflict" : "failed",
        assignment: null, failureCode: document.failure_code, currentRecordVersion: document.current_record_version,
        currentState: document.current_state, requestId: document.request_id, auditRef: document.audit_ref,
        summary: runtimeFailureSummary(document.failure_code),
      };
    }
    const assignment = parseAssignment(document.assignment);
    if (!assignment) return failedRuntime("workflow_rag_runtime_store_contract_mismatch");
    return { status: "ready", assignment, failureCode: "", currentRecordVersion: assignment.recordVersion, currentState: assignment.state, requestId: document.request_id, auditRef: document.audit_ref, summary: `Runtime assignment ${assignment.assignmentId} is ${assignment.state}.` };
  } catch {
    return failedRuntime("workflow_rag_runtime_store_unavailable", requestId);
  }
}

function parseAssignment(value: AssignmentDocument | null): WorkflowRAGApplicationRuntimeAssignment | null {
  if (!value || !isRecord(value) || Object.keys(value).length !== 22 || containsForbiddenField(value) ||
    value.schema_version !== ASSIGNMENT_SCHEMA_VERSION || !ASSIGNMENT_ID_PATTERN.test(value.assignment_id) ||
    !isPositiveInteger(value.record_version) || !DIGEST_PATTERN.test(value.assignment_digest) ||
    !REF_PATTERN.test(value.tenant_ref) || !REF_PATTERN.test(value.workspace_id) ||
    !APPLICATION_ID_PATTERN.test(value.application_id) || !REF_PATTERN.test(value.owner_subject_ref) ||
    (value.state !== "active" && value.state !== "revoked") || !REF_PATTERN.test(value.publish_candidate_id) ||
    !isPositiveInteger(value.publish_review_version) || value.publish_candidate_state !== "approved" ||
    !REF_PATTERN.test(value.draft_id) || !isPositiveInteger(value.draft_version) || !DIGEST_PATTERN.test(value.draft_digest) ||
    !isBindingRef(value.binding_ref) || !isTimestamp(value.created_at) || !isTimestamp(value.updated_at) ||
    !REF_PATTERN.test(value.created_by_actor_ref) || !REF_PATTERN.test(value.updated_by_actor_ref) ||
    !REF_PATTERN.test(value.request_id) || !REF_PATTERN.test(value.audit_ref)) return null;
  return {
    schemaVersion: ASSIGNMENT_SCHEMA_VERSION, assignmentId: value.assignment_id, recordVersion: value.record_version,
    assignmentDigest: value.assignment_digest, tenantRef: value.tenant_ref, workspaceId: value.workspace_id,
    applicationId: value.application_id, ownerSubjectRef: value.owner_subject_ref, state: value.state,
    publishCandidateId: value.publish_candidate_id, publishReviewVersion: value.publish_review_version,
    draftId: value.draft_id, draftVersion: value.draft_version, draftDigest: value.draft_digest,
    bindingRef: { bindingId: value.binding_ref.binding_id, bindingVersion: 1, bindingDigest: value.binding_ref.binding_digest },
    createdAt: value.created_at, updatedAt: value.updated_at, updatedByActorRef: value.updated_by_actor_ref,
    requestId: value.request_id, auditRef: value.audit_ref,
  };
}

function parseAnswer(value: AnswerDocument | null): WorkflowRAGApplicationAnswer | null {
  if (!value || !isRecord(value) || Object.keys(value).length !== 5 || containsForbiddenField(value) ||
    value.schema_version !== ANSWER_SCHEMA_VERSION || typeof value.answer !== "string" || value.answer.length < 1 || value.answer.length > 16384 ||
    !Array.isArray(value.citations) || value.citations.length < 1 || value.citations.length > 8 ||
    !Array.isArray(value.limitations) || value.limitations.length > 8 || !value.limitations.every((item) => typeof item === "string" && item.length >= 1 && item.length <= 512) ||
    (value.confidence !== "low" && value.confidence !== "medium" && value.confidence !== "high")) return null;
  const citations = value.citations.map((citation) => {
    if (!isRecord(citation) || Object.keys(citation).length !== 2 || !FRAGMENT_REF_PATTERN.test(citation.fragment_ref) ||
      typeof citation.claim_summary !== "string" || citation.claim_summary.length < 1 || citation.claim_summary.length > 512) return null;
    return { fragmentRef: citation.fragment_ref, claimSummary: citation.claim_summary };
  });
  if (citations.some((citation) => citation === null) || new Set(citations.map((citation) => citation?.fragmentRef)).size !== citations.length) return null;
  return { schemaVersion: ANSWER_SCHEMA_VERSION, answer: value.answer, citations: citations as Array<{ fragmentRef: string; claimSummary: string }>, limitations: [...value.limitations], confidence: value.confidence };
}

function isRuntimeEnvelope(value: unknown, config: WorkflowRAGApplicationRuntimeConfig, applicationId: string): value is RuntimeEnvelope {
  return isRecord(value) && Object.keys(value).length === 11 && !containsForbiddenField(value) &&
    REF_PATTERN.test(String(value.request_id)) && value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId &&
    value.application_id === applicationId && (value.assignment === null || isRecord(value.assignment)) &&
    Array.isArray(value.events) && Array.isArray(value.audits) &&
    (value.failure_code === null || typeof value.failure_code === "string") &&
    typeof value.current_record_version === "number" && Number.isInteger(value.current_record_version) && value.current_record_version >= 0 &&
    typeof value.current_state === "string" && REF_PATTERN.test(String(value.audit_ref));
}

function isInvocationEnvelope(value: unknown, config: WorkflowRAGApplicationRuntimeConfig, applicationId: string): value is InvocationEnvelope {
  return isRecord(value) && Object.keys(value).length === 9 && !containsForbiddenField(value) &&
    REF_PATTERN.test(String(value.request_id)) && value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId &&
    value.application_id === applicationId && (value.run === null || isRecord(value.run)) &&
    (value.answer === null || isRecord(value.answer)) && (value.failure_code === null || typeof value.failure_code === "string") &&
    typeof value.failure_summary === "string" && REF_PATTERN.test(String(value.audit_ref));
}

function managementHeaders(config: WorkflowRAGApplicationRuntimeConfig, applicationId: string, requestId: string, operation: "read" | "write") {
  return {
    "X-Request-Id": requestId,
    "X-RadishMind-Dev-Read-Identity": `workflow-rag-application-runtime-web:${config.subjectRef}`,
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Workflow-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Workflow-Application": applicationId,
    "X-RadishMind-Dev-Read-Scopes": operation === "write" ? "workflow_rag_runtime:write" : "workflow_rag_runtime:read",
  };
}

function failedRuntime(failureCode: string, requestId = ""): WorkflowRAGApplicationRuntimeResult {
  return { status: "failed", assignment: null, failureCode, currentRecordVersion: 0, currentState: "", requestId, auditRef: "", summary: runtimeFailureSummary(failureCode) };
}

function failedInvocation(failureCode: string): WorkflowRAGApplicationInvocationResult {
  return { status: "failed", runId: "", runStatus: "", answer: null, failureCode, failureSummary: "", requestId: "", auditRef: "", summary: runtimeFailureSummary(failureCode) };
}

function runtimeFailureSummary(failureCode: string): string {
  const summaries: Record<string, string> = {
    workflow_rag_runtime_assignment_not_found: "No runtime assignment exists for this application.",
    workflow_rag_runtime_assignment_revoked: "The current runtime assignment is revoked.",
    workflow_rag_runtime_assignment_version_conflict: "The assignment changed. Refresh before deciding again.",
    workflow_rag_runtime_candidate_not_approved: "The selected publish candidate is not approved.",
    workflow_rag_runtime_candidate_superseded: "The selected publish candidate has been superseded.",
    workflow_rag_runtime_no_evidence: "The active candidate snapshot contains no evidence for this input.",
    workflow_rag_runtime_citation_invalid: "The provider answer did not preserve valid selected-fragment citations.",
    workflow_rag_runtime_secret_material_forbidden: "Sensitive material is not allowed in this operation.",
  };
  return summaries[failureCode] ?? `Application RAG operation failed: ${failureCode}.`;
}

function containsSensitiveText(value: string): boolean {
  return /(authorization\s*:|bearer\s+|api[_ -]?key|token\s*[=:]|password\s*[=:]|secret\s*[=:]|cookie\s*:|postgres(?:ql)?:\/\/|https?:\/\/)/iu.test(value);
}

function containsForbiddenField(value: unknown): boolean {
  if (Array.isArray(value)) return value.some(containsForbiddenField);
  if (!isRecord(value)) return false;
  return Object.entries(value).some(([key, nested]) => FORBIDDEN_RESPONSE_FIELDS.has(key.toLowerCase()) || containsForbiddenField(nested));
}

function isBindingRef(value: unknown): value is AssignmentDocument["binding_ref"] {
  return isRecord(value) && Object.keys(value).length === 3 && BINDING_ID_PATTERN.test(String(value.binding_id)) &&
    value.binding_version === 1 && DIGEST_PATTERN.test(String(value.binding_digest));
}

function isRecord(value: unknown): value is Record<string, any> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function isPositiveInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value > 0;
}

function isTimestamp(value: unknown): value is string {
  return typeof value === "string" && value.length >= 20 && Number.isFinite(Date.parse(value));
}

function isRunStatus(value: unknown): value is WorkflowRAGApplicationInvocationResult["runStatus"] {
  return value === "running" || value === "succeeded" || value === "failed" || value === "canceled";
}

function normalizeBaseUrl(value: string): string {
  return value.trim().replace(/\/$/u, "");
}

function createRequestId(prefix: string): string {
  const suffix = (globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random()}`).replaceAll("-", "").slice(0, 20);
  return `${prefix}-${suffix}`;
}
