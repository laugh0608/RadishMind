import { CONTROL_PLANE_READ_ROUTES } from "../../../../../contracts/typescript/control-plane-read-api.ts";
import type { WorkflowDraftDesignerDraft } from "./workflowDraftDesigner.ts";
import {
  parseWorkflowRunRecordDocument,
  type WorkflowRunRecord,
} from "./workflowRunRecordConsumer.ts";

const DEV_SOURCE = "dev-workflow-definition-promotion-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const WORKFLOW_DEFINITIONS_ROUTE = CONTROL_PLANE_READ_ROUTES.workflowDefinitions;
const ID_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._:-]{0,159}$/u;
const REF_PATTERN = /^[A-Za-z0-9][A-Za-z0-9_.:/-]{2,239}$/u;
const DIGEST_PATTERN = /^sha256:[a-f0-9]{64}$/u;
const FORBIDDEN_RESPONSE_KEYS = new Set([
  "input_text", "condition_values", "prompt", "messages", "answer", "credential",
  "token", "authorization", "cookie", "header", "headers", "raw_request", "raw_response",
  "provider_raw_envelope", "endpoint", "url", "uri",
]);

export type WorkflowDefinitionPromotionConfig = {
  mode: "offline" | "dev_workflow_definition_promotion_http";
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  subjectRef: string;
};

export type WorkflowDefinitionSnapshot = {
  schemaVersion: "saved_workflow_draft.v1";
  name: string;
  description: string;
  nodes: Array<{
    nodeId: string;
    nodeType: "prompt" | "llm" | "condition" | "output";
    label: string;
    inputSummary: string;
    outputSummary: string;
    inputContractRef: string;
    outputContractRef: string;
    inputContractFields: string[];
    outputContractFields: string[];
    outputMappingSummary: string;
    providerRef: string;
    toolRef: string;
    ragRef: string;
    riskLevel: string;
    requiresConfirmation: boolean;
  }>;
  edges: Array<{ edgeId: string; fromNodeId: string; toNodeId: string; conditionSummary: string }>;
  inputContract: { contractId: string; requiredFields: string[]; summary: string };
  outputContract: { contractId: string; requiredFields: string[]; summary: string };
  providerRefs: string[];
  toolRefs: string[];
  ragRefs: string[];
  requestedCapabilities: string[];
  executionProfile: "workflow_definition_executor_v1";
};

export type WorkflowDefinitionReview = {
  reviewVersion: number;
  decision: "approve" | "reject";
  reason: string;
  reviewerRef: string;
  reviewedAt: string;
  requestId: string;
  auditRef: string;
};

export type WorkflowDefinitionCandidate = {
  candidateId: string;
  definitionId: string;
  sourceDraftId: string;
  sourceDraftVersion: number;
  sourceDraftDigest: string;
  definitionDigest: string;
  snapshot: WorkflowDefinitionSnapshot;
  activationEligible: boolean;
  eligibilityBlockers: string[];
  state: "pending" | "approved" | "rejected" | "superseded";
  reviewVersion: number;
  reviews: WorkflowDefinitionReview[];
  createdAt: string;
  updatedAt: string;
  requestId: string;
  auditRef: string;
};

export type WorkflowDefinitionVersion = {
  definitionId: string;
  version: number;
  definitionDigest: string;
  candidateId: string;
  candidateReviewVersion: number;
  sourceDraftId: string;
  sourceDraftVersion: number;
  sourceDraftDigest: string;
  snapshot: WorkflowDefinitionSnapshot;
  activationEligible: boolean;
  eligibilityBlockers: string[];
  createdAt: string;
  createdByActorRef: string;
  requestId: string;
  auditRef: string;
};

export type WorkflowDefinitionActivation = {
  definitionId: string;
  pointerVersion: number;
  state: "active" | "inactive";
  activeVersion: number;
  activeDefinitionDigest: string;
  events: Array<{
    eventId: string;
    decision: "activate" | "replace" | "deactivate";
    reason: string;
    beforePointerVersion: number;
    afterPointerVersion: number;
    beforeActiveVersion: number;
    afterActiveVersion: number;
    actorRef: string;
    createdAt: string;
    requestId: string;
    auditRef: string;
  }>;
  updatedAt: string;
  updatedByActorRef: string;
  requestId: string;
  auditRef: string;
};

export class WorkflowDefinitionPromotionConflict extends Error {
  readonly failureCode: string;
  readonly currentReviewVersion: number;
  readonly currentPointerVersion: number;

  constructor(
    failureCode: string,
    currentReviewVersion: number,
    currentPointerVersion: number,
  ) {
    super(`${failureCode}: refresh the authoritative record before deciding again.`);
    this.name = "WorkflowDefinitionPromotionConflict";
    this.failureCode = failureCode;
    this.currentReviewVersion = currentReviewVersion;
    this.currentPointerVersion = currentPointerVersion;
  }
}

type CandidateDocument = {
  schema_version: "workflow_definition_release_candidate.v1";
  candidate_id: string;
  definition_id: string;
  source_draft_id: string;
  source_draft_version: number;
  source_draft_digest: string;
  definition_digest: string;
  snapshot: SnapshotDocument;
  activation_eligible: boolean;
  eligibility_blockers: string[];
  state: WorkflowDefinitionCandidate["state"];
  review_version: number;
  reviews: ReviewDocument[];
  created_at: string;
  updated_at: string;
  created_by_actor_ref: string;
  updated_by_actor_ref: string;
  request_id: string;
  audit_ref: string;
};

type VersionDocument = {
  schema_version: "workflow_definition_version.v1";
  definition_id: string;
  version: number;
  definition_digest: string;
  candidate_id: string;
  candidate_review_version: number;
  source_draft_id: string;
  source_draft_version: number;
  source_draft_digest: string;
  snapshot: SnapshotDocument;
  activation_eligible: boolean;
  eligibility_blockers: string[];
  created_at: string;
  created_by_actor_ref: string;
  request_id: string;
  audit_ref: string;
};

type ActivationDocument = {
  schema_version: "workflow_definition_activation.v1";
  definition_id: string;
  pointer_version: number;
  state: WorkflowDefinitionActivation["state"];
  active_version: number;
  active_definition_digest: string;
  events: ActivationEventDocument[];
  updated_at: string;
  updated_by_actor_ref: string;
  request_id: string;
  audit_ref: string;
};

type ReviewDocument = {
  schema_version: "workflow_definition_release_decision.v1";
  review_version: number;
  decision: "approve" | "reject";
  reason: string;
  reviewer_ref: string;
  reviewed_at: string;
  request_id: string;
  audit_ref: string;
};

type ActivationEventDocument = {
  schema_version: "workflow_definition_activation_event.v1";
  event_id: string;
  definition_id: string;
  decision: "activate" | "replace" | "deactivate";
  reason: string;
  before_pointer_version: number;
  after_pointer_version: number;
  before_active_version: number;
  after_active_version: number;
  actor_ref: string;
  created_at: string;
  request_id: string;
  audit_ref: string;
};

type SnapshotDocument = {
  schema_version: "saved_workflow_draft.v1";
  name: string;
  description: string;
  nodes: Array<Record<string, unknown>>;
  edges: Array<Record<string, unknown>>;
  input_contract: Record<string, unknown>;
  output_contract: Record<string, unknown>;
  provider_refs: string[];
  tool_refs: string[];
  rag_refs: string[];
  requested_capabilities: string[];
  execution_profile: "workflow_definition_executor_v1";
};

type ReleaseEnvelope = {
  request_id: string;
  workspace_id: string;
  application_id: string;
  candidate: CandidateDocument | null;
  version: VersionDocument | null;
  activation: ActivationDocument | null;
  failure_code: string | null;
  current_review_version: number;
  current_pointer_version: number;
  audit_ref: string;
};

export function readWorkflowDefinitionPromotionConfig(): WorkflowDefinitionPromotionConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  return {
    mode: env.VITE_RADISHMIND_WORKFLOW_DEFINITION_PROMOTION_SOURCE?.trim() === DEV_SOURCE
      ? "dev_workflow_definition_promotion_http"
      : "offline",
    baseUrl: normalizeBaseUrl(env.VITE_RADISHMIND_WORKFLOW_DEFINITION_PROMOTION_BASE_URL ?? env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ?? DEFAULT_BASE_URL),
    tenantRef: env.VITE_RADISHMIND_DEV_READ_TENANT_REF?.trim() || "tenant_demo",
    workspaceId: env.VITE_RADISHMIND_WORKFLOW_DEFINITION_PROMOTION_WORKSPACE_ID?.trim() || "workspace_demo",
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || "subject_demo_user",
  };
}

export async function listWorkflowDefinitionCandidates(config: WorkflowDefinitionPromotionConfig, applicationId: string): Promise<WorkflowDefinitionCandidate[]> {
  if (config.mode === "offline") return [];
  const body = await readJSON(`${config.baseUrl}/v1/user-workspace/workflow-definition-candidates?${scopeQuery(config, applicationId)}`, config, applicationId, "workflow_definitions:read");
  if (!isCandidateListEnvelope(body, config, applicationId) || body.failure_code) throw responseError(body, "candidate list");
  return body.candidates.map(mapCandidate);
}

export async function createWorkflowDefinitionCandidate(config: WorkflowDefinitionPromotionConfig, applicationId: string, input: { candidateId: string; definitionId: string; draftId: string; expectedDraftVersion: number }): Promise<WorkflowDefinitionCandidate> {
  assertLive(config);
  return requireCandidate(await writeRelease(config, applicationId, "/v1/user-workspace/workflow-definition-candidates", "workflow_definitions:write", {
    candidate_id: input.candidateId,
    definition_id: input.definitionId,
    draft_id: input.draftId,
    expected_draft_version: input.expectedDraftVersion,
  }));
}

export async function decideWorkflowDefinitionCandidate(config: WorkflowDefinitionPromotionConfig, applicationId: string, candidateId: string, input: { expectedReviewVersion: number; decision: "approve" | "reject"; reason: string }): Promise<{ candidate: WorkflowDefinitionCandidate; version: WorkflowDefinitionVersion | null }> {
  assertLive(config);
  const envelope = await writeRelease(config, applicationId, `/v1/user-workspace/workflow-definition-candidates/${encodeURIComponent(candidateId)}/decisions`, "workflow_definitions:review", {
    expected_review_version: input.expectedReviewVersion,
    decision: input.decision,
    reason: input.reason,
  });
  return { candidate: requireCandidate(envelope), version: envelope.version ? mapVersion(envelope.version) : null };
}

export async function listWorkflowDefinitionVersions(config: WorkflowDefinitionPromotionConfig, applicationId: string, definitionId: string): Promise<WorkflowDefinitionVersion[]> {
  if (config.mode === "offline") return [];
  const body = await readJSON(`${config.baseUrl}${WORKFLOW_DEFINITIONS_ROUTE}/${encodeURIComponent(definitionId)}/versions?${scopeQuery(config, applicationId)}`, config, applicationId, "workflow_definitions:read");
  if (!isVersionListEnvelope(body, config, applicationId, definitionId) || body.failure_code) throw responseError(body, "version list");
  return body.versions.map(mapVersion);
}

export async function readWorkflowDefinitionActivation(config: WorkflowDefinitionPromotionConfig, applicationId: string, definitionId: string): Promise<WorkflowDefinitionActivation | null> {
  if (config.mode === "offline") return null;
  const body = await readRelease(config, applicationId, `${WORKFLOW_DEFINITIONS_ROUTE}/${encodeURIComponent(definitionId)}/activation`, "workflow_definitions:read");
  if (body.failure_code === "workflow_definition_not_found") return null;
  if (body.failure_code) throw responseError(body, "activation read");
  return body.activation ? mapActivation(body.activation) : null;
}

export async function decideWorkflowDefinitionActivation(config: WorkflowDefinitionPromotionConfig, applicationId: string, definitionId: string, input: { expectedPointerVersion: number; decision: "activate" | "replace" | "deactivate"; version: number; reason: string }): Promise<WorkflowDefinitionActivation> {
  assertLive(config);
  const body = await writeRelease(config, applicationId, `${WORKFLOW_DEFINITIONS_ROUTE}/${encodeURIComponent(definitionId)}/activation-decisions`, "workflow_definitions:activate", {
    expected_pointer_version: input.expectedPointerVersion,
    decision: input.decision,
    version: input.version,
    reason: input.reason,
  });
  if (!body.activation) throw responseError(body, "activation decision");
  return mapActivation(body.activation);
}

export async function startWorkflowDefinitionRun(config: WorkflowDefinitionPromotionConfig, applicationId: string, input: { definitionId: string; expectedPointerVersion: number; expectedDefinitionVersion: number; expectedDefinitionDigest: string; inputText: string; conditionValues: Record<string, boolean>; model: string }): Promise<{ record: WorkflowRunRecord; advisoryOutput: string }> {
  assertLive(config);
  const response = await fetch(`${config.baseUrl}/v1/user-workspace/workflow-definition-runs`, {
    method: "POST",
    headers: { ...headers(config, applicationId, "workflow_runs:execute,workflow_definitions:read"), "Content-Type": "application/json" },
    body: JSON.stringify({
      workspace_id: config.workspaceId,
      application_id: applicationId,
      definition_id: input.definitionId,
      expected_pointer_version: input.expectedPointerVersion,
      expected_definition_version: input.expectedDefinitionVersion,
      expected_definition_digest: input.expectedDefinitionDigest,
      input_text: input.inputText,
      condition_values: input.conditionValues,
      model: input.model,
      temperature: null,
    }),
  });
  const body: unknown = await response.json();
  if (!isRunEnvelope(body, config, applicationId)) throw new Error("workflow definition run returned an invalid or sensitive envelope");
  if (!response.ok || body.failure_code || !body.run) throw responseError(body, "definition run");
  const record = parseWorkflowRunRecordDocument(body.run);
  if (!record || record.schemaVersion !== "workflow_run_record.v5" || record.executionSourceKind !== "workflow_definition") {
    throw new Error("workflow definition run did not return strict workflow_run_record.v5 evidence");
  }
  return { record, advisoryOutput: body.advisory_output ?? "" };
}

export function deriveWorkflowDraftFromDefinitionVersion(version: WorkflowDefinitionVersion, applicationId: string, draftNumber: number): WorkflowDraftDesignerDraft {
  const number = String(Math.max(1, draftNumber)).padStart(2, "0");
  const draftId = `draft_${safeIdPart(version.definitionId)}_v${version.version}_${number}`;
  return {
    draftId,
    templateRef: `${version.definitionId}:v${version.version}`,
    label: `${version.snapshot.name} derived ${number}`,
    applicationRef: applicationId,
    workflowDefinitionId: version.definitionId,
    baseDefinitionVersion: version.version,
    providerProfileRef: version.snapshot.providerRefs[0] ?? "",
    summary: `Derived from immutable ${version.definitionId} v${version.version}; edits require a new candidate, review, and activation.`,
    nodes: version.snapshot.nodes.map((node) => ({
      ...node,
      riskLevel: node.riskLevel === "high" ? "high" : node.riskLevel === "medium" ? "medium" : "low",
      lane: node.nodeType === "llm" ? "model" : node.nodeType === "condition" ? "policy" : node.nodeType === "output" ? "output" : "context",
      readiness: node.requiresConfirmation ? "review_required" : "ready",
      previewOnlyReason: "Derived editable draft; the source definition version remains immutable.",
    })),
    edges: version.snapshot.edges.map((edge) => ({ ...edge, edgeKind: "context" })),
    designerLayout: { source: "workflow_node_designer", persistence: "ui_only", nodePositions: version.snapshot.nodes.map((node, index) => ({ nodeId: node.nodeId, x: index * 320, y: 0 })) },
    readiness: [{ checkId: "definition_provenance", label: "Definition provenance", status: "ready", summary: `Exact source ${version.definitionId} v${version.version} · ${version.definitionDigest}.` }],
    risks: [],
    blockedCapabilities: [],
    routeMetadata: { sourceRouteId: "workflow-definition-summary-list-route", draftRouteId: "workflow-draft-designer-offline-draft", routePath: WORKFLOW_DEFINITIONS_ROUTE, requestId: version.requestId, auditRef: version.auditRef },
    localOnlyInteraction: "local_edit",
    executionProfile: "executor_v0",
  };
}

async function readRelease(config: WorkflowDefinitionPromotionConfig, applicationId: string, path: string, scope: string): Promise<ReleaseEnvelope> {
  const body = await readJSON(`${config.baseUrl}${path}?${scopeQuery(config, applicationId)}`, config, applicationId, scope);
  if (!isReleaseEnvelope(body, config, applicationId)) throw new Error("workflow definition release returned an invalid or sensitive envelope");
  return body;
}

async function writeRelease(config: WorkflowDefinitionPromotionConfig, applicationId: string, path: string, scope: string, payload: unknown): Promise<ReleaseEnvelope> {
  const response = await fetch(`${config.baseUrl}${path}`, { method: "POST", headers: { ...headers(config, applicationId, scope), "Content-Type": "application/json" }, body: JSON.stringify(payload) });
  const body: unknown = await response.json();
  if (!isReleaseEnvelope(body, config, applicationId)) throw new Error("workflow definition release returned an invalid or sensitive envelope");
  if (!response.ok || body.failure_code) {
    if (body.failure_code?.includes("conflict")) throw new WorkflowDefinitionPromotionConflict(body.failure_code, body.current_review_version, body.current_pointer_version);
    throw responseError(body, "workflow definition write");
  }
  return body;
}

async function readJSON(url: string, config: WorkflowDefinitionPromotionConfig, applicationId: string, scope: string): Promise<unknown> {
  const response = await fetch(url, { headers: headers(config, applicationId, scope) });
  const body: unknown = await response.json();
  if (!response.ok) throw responseError(body, `HTTP ${response.status}`);
  return body;
}

function headers(config: WorkflowDefinitionPromotionConfig, applicationId: string, scopes: string): HeadersInit {
  const requestId = `web-workflow-definition-${Date.now().toString(36)}`;
  return { Accept: "application/json", "X-Request-Id": requestId, "X-RadishMind-Dev-Read-Identity": "workflow-definition-promotion-web", "X-RadishMind-Dev-Read-Tenant": config.tenantRef, "X-RadishMind-Dev-Read-Subject": config.subjectRef, "X-RadishMind-Dev-Read-Scopes": scopes, "X-RadishMind-Dev-Read-Audit": `audit_${requestId}`, "X-RadishMind-Dev-Workflow-Workspace": config.workspaceId, "X-RadishMind-Dev-Workflow-Application": applicationId };
}

function scopeQuery(config: WorkflowDefinitionPromotionConfig, applicationId: string): string {
  return new URLSearchParams({ workspace_id: config.workspaceId, application_id: applicationId }).toString();
}

function assertLive(config: WorkflowDefinitionPromotionConfig): asserts config is WorkflowDefinitionPromotionConfig {
  if (config.mode !== "dev_workflow_definition_promotion_http") throw new Error("workflow definition promotion is unavailable in offline mode");
}

function isReleaseEnvelope(value: unknown, config: WorkflowDefinitionPromotionConfig, applicationId: string): value is ReleaseEnvelope {
  const keys = ["request_id", "workspace_id", "application_id", "candidate", "version", "activation", "failure_code", "current_review_version", "current_pointer_version", "audit_ref"];
  if (!strictObject(value, keys) || containsForbiddenResponseKey(value)) return false;
  const item = value as Partial<ReleaseEnvelope>;
  return item.workspace_id === config.workspaceId && item.application_id === applicationId && typeof item.request_id === "string" && typeof item.audit_ref === "string" &&
    (item.failure_code === null || typeof item.failure_code === "string") && Number.isInteger(item.current_review_version) && Number.isInteger(item.current_pointer_version) &&
    (item.candidate === null || isCandidateDocument(item.candidate)) && (item.version === null || isVersionDocument(item.version)) && (item.activation === null || isActivationDocument(item.activation));
}

function isCandidateListEnvelope(value: unknown, config: WorkflowDefinitionPromotionConfig, applicationId: string): value is { request_id: string; workspace_id: string; application_id: string; candidates: CandidateDocument[]; failure_code: string | null; audit_ref: string } {
  if (!strictObject(value, ["request_id", "workspace_id", "application_id", "candidates", "failure_code", "audit_ref"]) || containsForbiddenResponseKey(value)) return false;
  const item = value as Record<string, unknown>;
  return item.workspace_id === config.workspaceId && item.application_id === applicationId && typeof item.request_id === "string" && typeof item.audit_ref === "string" && (item.failure_code === null || typeof item.failure_code === "string") && Array.isArray(item.candidates) && item.candidates.every(isCandidateDocument);
}

function isVersionListEnvelope(value: unknown, config: WorkflowDefinitionPromotionConfig, applicationId: string, definitionId: string): value is { request_id: string; workspace_id: string; application_id: string; definition_id: string; versions: VersionDocument[]; failure_code: string | null; audit_ref: string } {
  if (!strictObject(value, ["request_id", "workspace_id", "application_id", "definition_id", "versions", "failure_code", "audit_ref"]) || containsForbiddenResponseKey(value)) return false;
  const item = value as Record<string, unknown>;
  return item.workspace_id === config.workspaceId && item.application_id === applicationId && item.definition_id === definitionId && typeof item.request_id === "string" && typeof item.audit_ref === "string" && (item.failure_code === null || typeof item.failure_code === "string") && Array.isArray(item.versions) && item.versions.every(isVersionDocument);
}

function isRunEnvelope(value: unknown, config: WorkflowDefinitionPromotionConfig, applicationId: string): value is { request_id: string; workspace_id: string; application_id: string; run: unknown | null; advisory_output?: string; failure_code: string | null; failure_summary: string; audit_ref: string } {
  const keys = value && typeof value === "object" && "advisory_output" in value
    ? ["request_id", "workspace_id", "application_id", "run", "advisory_output", "failure_code", "failure_summary", "audit_ref"]
    : ["request_id", "workspace_id", "application_id", "run", "failure_code", "failure_summary", "audit_ref"];
  if (!strictObject(value, keys) || containsForbiddenResponseKey(value)) return false;
  const item = value as Record<string, unknown>;
  return item.workspace_id === config.workspaceId && item.application_id === applicationId && typeof item.request_id === "string" && typeof item.audit_ref === "string" && typeof item.failure_summary === "string" && (item.advisory_output === undefined || typeof item.advisory_output === "string") && (item.failure_code === null || typeof item.failure_code === "string") && (item.run === null || typeof item.run === "object");
}

function isCandidateDocument(value: unknown): value is CandidateDocument {
  const keys = ["schema_version", "candidate_id", "definition_id", "source_draft_id", "source_draft_version", "source_draft_digest", "definition_digest", "snapshot", "activation_eligible", "eligibility_blockers", "state", "review_version", "reviews", "created_at", "updated_at", "created_by_actor_ref", "updated_by_actor_ref", "request_id", "audit_ref"];
  if (!strictObject(value, keys)) return false;
  const item = value as Record<string, unknown>;
  return item.schema_version === "workflow_definition_release_candidate.v1" && id(item.candidate_id) && id(item.definition_id) && id(item.source_draft_id) && positive(item.source_draft_version) && digest(item.source_draft_digest) && digest(item.definition_digest) && isSnapshotDocument(item.snapshot) && typeof item.activation_eligible === "boolean" && strings(item.eligibility_blockers) && ["pending", "approved", "rejected", "superseded"].includes(String(item.state)) && nonnegative(item.review_version) && Array.isArray(item.reviews) && item.reviews.every(isReviewDocument) && timestamp(item.created_at) && timestamp(item.updated_at) && ref(item.created_by_actor_ref) && ref(item.updated_by_actor_ref) && ref(item.request_id) && ref(item.audit_ref);
}

function isVersionDocument(value: unknown): value is VersionDocument {
  const keys = ["schema_version", "definition_id", "version", "definition_digest", "candidate_id", "candidate_review_version", "source_draft_id", "source_draft_version", "source_draft_digest", "snapshot", "activation_eligible", "eligibility_blockers", "created_at", "created_by_actor_ref", "request_id", "audit_ref"];
  if (!strictObject(value, keys)) return false;
  const item = value as Record<string, unknown>;
  return item.schema_version === "workflow_definition_version.v1" && id(item.definition_id) && positive(item.version) && digest(item.definition_digest) && id(item.candidate_id) && positive(item.candidate_review_version) && id(item.source_draft_id) && positive(item.source_draft_version) && digest(item.source_draft_digest) && isSnapshotDocument(item.snapshot) && typeof item.activation_eligible === "boolean" && strings(item.eligibility_blockers) && timestamp(item.created_at) && ref(item.created_by_actor_ref) && ref(item.request_id) && ref(item.audit_ref);
}

function isActivationDocument(value: unknown): value is ActivationDocument {
  const keys = ["schema_version", "definition_id", "pointer_version", "state", "active_version", "active_definition_digest", "events", "updated_at", "updated_by_actor_ref", "request_id", "audit_ref"];
  if (!strictObject(value, keys)) return false;
  const item = value as Record<string, unknown>;
  return item.schema_version === "workflow_definition_activation.v1" && id(item.definition_id) && nonnegative(item.pointer_version) && ["active", "inactive"].includes(String(item.state)) && nonnegative(item.active_version) && (item.active_definition_digest === "" || digest(item.active_definition_digest)) && Array.isArray(item.events) && item.events.every(isActivationEventDocument) && typeof item.updated_at === "string" && typeof item.updated_by_actor_ref === "string" && typeof item.request_id === "string" && typeof item.audit_ref === "string";
}

function isReviewDocument(value: unknown): value is ReviewDocument {
  if (!strictObject(value, ["schema_version", "review_version", "decision", "reason", "reviewer_ref", "reviewed_at", "request_id", "audit_ref"])) return false;
  const item = value as Record<string, unknown>;
  return item.schema_version === "workflow_definition_release_decision.v1" && positive(item.review_version) && ["approve", "reject"].includes(String(item.decision)) && typeof item.reason === "string" && ref(item.reviewer_ref) && timestamp(item.reviewed_at) && ref(item.request_id) && ref(item.audit_ref);
}

function isActivationEventDocument(value: unknown): value is ActivationEventDocument {
  if (!strictObject(value, ["schema_version", "event_id", "definition_id", "decision", "reason", "before_pointer_version", "after_pointer_version", "before_active_version", "after_active_version", "actor_ref", "created_at", "request_id", "audit_ref"])) return false;
  const item = value as Record<string, unknown>;
  return item.schema_version === "workflow_definition_activation_event.v1" && id(item.event_id) && id(item.definition_id) && ["activate", "replace", "deactivate"].includes(String(item.decision)) && typeof item.reason === "string" && nonnegative(item.before_pointer_version) && positive(item.after_pointer_version) && nonnegative(item.before_active_version) && nonnegative(item.after_active_version) && ref(item.actor_ref) && timestamp(item.created_at) && ref(item.request_id) && ref(item.audit_ref);
}

function isSnapshotDocument(value: unknown): value is SnapshotDocument {
  if (!strictObject(value, ["schema_version", "name", "description", "nodes", "edges", "input_contract", "output_contract", "provider_refs", "tool_refs", "rag_refs", "requested_capabilities", "execution_profile"])) return false;
  const item = value as Record<string, unknown>;
  return item.schema_version === "saved_workflow_draft.v1" && typeof item.name === "string" && typeof item.description === "string" && Array.isArray(item.nodes) && item.nodes.length > 0 && item.nodes.every(isSnapshotNode) && Array.isArray(item.edges) && item.edges.every(isSnapshotEdge) && isContract(item.input_contract) && isContract(item.output_contract) && strings(item.provider_refs) && strings(item.tool_refs) && strings(item.rag_refs) && strings(item.requested_capabilities) && item.execution_profile === "workflow_definition_executor_v1";
}

function isSnapshotNode(value: unknown): value is Record<string, unknown> {
  const keys = ["node_id", "node_type", "label", "input_summary", "output_summary", "input_contract_ref", "output_contract_ref", "input_contract_fields", "output_contract_fields", "output_mapping_summary", "provider_ref", "tool_ref", "rag_ref", "risk_level", "requires_confirmation"];
  if (!strictObject(value, keys)) return false;
  const item = value as Record<string, unknown>;
  return id(item.node_id) && ["prompt", "llm", "condition", "output"].includes(String(item.node_type)) && keys.slice(2, 9).every((key) => key.endsWith("fields") ? strings(item[key]) : typeof item[key] === "string") && typeof item.output_mapping_summary === "string" && typeof item.provider_ref === "string" && typeof item.tool_ref === "string" && typeof item.rag_ref === "string" && typeof item.risk_level === "string" && typeof item.requires_confirmation === "boolean";
}

function isSnapshotEdge(value: unknown): value is Record<string, unknown> {
  if (!strictObject(value, ["edge_id", "from_node_id", "to_node_id", "condition_summary"])) return false;
  const item = value as Record<string, unknown>;
  return id(item.edge_id) && id(item.from_node_id) && id(item.to_node_id) && typeof item.condition_summary === "string";
}

function isContract(value: unknown): value is Record<string, unknown> {
  return strictObject(value, ["contract_id", "required_fields", "summary"]) && typeof (value as Record<string, unknown>).contract_id === "string" && strings((value as Record<string, unknown>).required_fields) && typeof (value as Record<string, unknown>).summary === "string";
}

function mapCandidate(value: CandidateDocument): WorkflowDefinitionCandidate { return { candidateId: value.candidate_id, definitionId: value.definition_id, sourceDraftId: value.source_draft_id, sourceDraftVersion: value.source_draft_version, sourceDraftDigest: value.source_draft_digest, definitionDigest: value.definition_digest, snapshot: mapSnapshot(value.snapshot), activationEligible: value.activation_eligible, eligibilityBlockers: [...value.eligibility_blockers], state: value.state, reviewVersion: value.review_version, reviews: value.reviews.map((review) => ({ reviewVersion: review.review_version, decision: review.decision, reason: review.reason, reviewerRef: review.reviewer_ref, reviewedAt: review.reviewed_at, requestId: review.request_id, auditRef: review.audit_ref })), createdAt: value.created_at, updatedAt: value.updated_at, requestId: value.request_id, auditRef: value.audit_ref }; }
function mapVersion(value: VersionDocument): WorkflowDefinitionVersion { return { definitionId: value.definition_id, version: value.version, definitionDigest: value.definition_digest, candidateId: value.candidate_id, candidateReviewVersion: value.candidate_review_version, sourceDraftId: value.source_draft_id, sourceDraftVersion: value.source_draft_version, sourceDraftDigest: value.source_draft_digest, snapshot: mapSnapshot(value.snapshot), activationEligible: value.activation_eligible, eligibilityBlockers: [...value.eligibility_blockers], createdAt: value.created_at, createdByActorRef: value.created_by_actor_ref, requestId: value.request_id, auditRef: value.audit_ref }; }
function mapActivation(value: ActivationDocument): WorkflowDefinitionActivation { return { definitionId: value.definition_id, pointerVersion: value.pointer_version, state: value.state, activeVersion: value.active_version, activeDefinitionDigest: value.active_definition_digest, events: value.events.map((event) => ({ eventId: event.event_id, decision: event.decision, reason: event.reason, beforePointerVersion: event.before_pointer_version, afterPointerVersion: event.after_pointer_version, beforeActiveVersion: event.before_active_version, afterActiveVersion: event.after_active_version, actorRef: event.actor_ref, createdAt: event.created_at, requestId: event.request_id, auditRef: event.audit_ref })), updatedAt: value.updated_at, updatedByActorRef: value.updated_by_actor_ref, requestId: value.request_id, auditRef: value.audit_ref }; }
function mapSnapshot(value: SnapshotDocument): WorkflowDefinitionSnapshot { return { schemaVersion: value.schema_version, name: value.name, description: value.description, nodes: value.nodes.map((node) => ({ nodeId: String(node.node_id), nodeType: node.node_type as WorkflowDefinitionSnapshot["nodes"][number]["nodeType"], label: String(node.label), inputSummary: String(node.input_summary), outputSummary: String(node.output_summary), inputContractRef: String(node.input_contract_ref), outputContractRef: String(node.output_contract_ref), inputContractFields: [...node.input_contract_fields as string[]], outputContractFields: [...node.output_contract_fields as string[]], outputMappingSummary: String(node.output_mapping_summary), providerRef: String(node.provider_ref), toolRef: String(node.tool_ref), ragRef: String(node.rag_ref), riskLevel: String(node.risk_level), requiresConfirmation: Boolean(node.requires_confirmation) })), edges: value.edges.map((edge) => ({ edgeId: String(edge.edge_id), fromNodeId: String(edge.from_node_id), toNodeId: String(edge.to_node_id), conditionSummary: String(edge.condition_summary) })), inputContract: mapContract(value.input_contract), outputContract: mapContract(value.output_contract), providerRefs: [...value.provider_refs], toolRefs: [...value.tool_refs], ragRefs: [...value.rag_refs], requestedCapabilities: [...value.requested_capabilities], executionProfile: value.execution_profile }; }
function mapContract(value: Record<string, unknown>) { return { contractId: String(value.contract_id), requiredFields: [...value.required_fields as string[]], summary: String(value.summary) }; }
function requireCandidate(value: ReleaseEnvelope): WorkflowDefinitionCandidate { if (!value.candidate) throw responseError(value, "candidate"); return mapCandidate(value.candidate); }
function responseError(value: unknown, operation: string): Error { if (value && typeof value === "object") { const item = value as Record<string, unknown>; if (typeof item.failure_code === "string") return new Error(`${item.failure_code}: ${typeof item.failure_summary === "string" ? item.failure_summary : operation}`); } return new Error(`${operation} failed`); }
function strictObject(value: unknown, keys: string[]): value is Record<string, unknown> { return Boolean(value) && typeof value === "object" && !Array.isArray(value) && Object.keys(value as Record<string, unknown>).length === keys.length && Object.keys(value as Record<string, unknown>).every((key) => keys.includes(key)); }
function containsForbiddenResponseKey(value: unknown): boolean { if (Array.isArray(value)) return value.some(containsForbiddenResponseKey); if (!value || typeof value !== "object") return false; return Object.entries(value as Record<string, unknown>).some(([key, nested]) => FORBIDDEN_RESPONSE_KEYS.has(key.toLowerCase()) || containsForbiddenResponseKey(nested)); }
function id(value: unknown): value is string { return typeof value === "string" && ID_PATTERN.test(value); }
function ref(value: unknown): value is string { return typeof value === "string" && REF_PATTERN.test(value); }
function digest(value: unknown): value is string { return typeof value === "string" && DIGEST_PATTERN.test(value); }
function positive(value: unknown): value is number { return Number.isInteger(value) && Number(value) >= 1; }
function nonnegative(value: unknown): value is number { return Number.isInteger(value) && Number(value) >= 0; }
function strings(value: unknown): value is string[] { return Array.isArray(value) && value.every((item) => typeof item === "string"); }
function timestamp(value: unknown): value is string { return typeof value === "string" && !Number.isNaN(Date.parse(value)); }
function normalizeBaseUrl(value: string): string { return value.trim().replace(/\/+$/u, ""); }
function safeIdPart(value: string): string { return value.toLowerCase().replace(/[^a-z0-9]+/gu, "_").replace(/^_+|_+$/gu, "").slice(0, 48) || "definition"; }
