import type { WorkflowDraftDesignerDraft } from "./workflowDraftDesigner.ts";
import {
  workflowDraftFromSavedWorkflowDraftDocument,
  type SavedWorkflowDraftDocument,
} from "./savedWorkflowDraftConsumer.ts";

const DEV_SOURCE = "dev-workflow-template-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const DEFAULT_WORKSPACE_ID = "workspace_demo";
const DEFAULT_TENANT_REF = "tenant_demo";
const DEFAULT_SUBJECT_REF = "subject_demo_user";
const ID_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._:-]{0,159}$/u;
const APPLICATION_ID_PATTERN = /^app_[a-z2-7]{16}$/u;
const DIGEST_PATTERN = /^sha256:[0-9a-f]{64}$/u;
const CURSOR_PATTERN = /^[A-Za-z0-9_-]+$/u;
const FORBIDDEN_RESPONSE_KEYS = new Set([
  "authorization", "api_key", "credential", "secret", "token", "cookie", "headers",
  "raw_request", "raw_response", "provider_raw_response", "prompt", "messages", "input", "output",
  "endpoint", "base_url", "dsn",
]);

export type WorkflowTemplateCatalogConfig = {
  mode: "offline" | "dev_workflow_template_http";
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  subjectRef: string;
};

export type WorkflowTemplatePortability = {
  executionProfile: "workflow_definition_executor_v1";
  nodeKinds: Array<"prompt" | "llm" | "condition" | "output">;
  providerRefs: string[];
  riskLevel: "low" | "medium" | "high";
  portable: true;
  blockers: [];
};

export type WorkflowTemplateDecision = {
  reviewVersion: 1;
  decision: "approve" | "reject" | "request_changes" | "withdraw";
  reason: string;
  reviewerRef: string;
  decidedAt: string;
  requestId: string;
  auditRef: string;
};

export type WorkflowTemplateCandidate = {
  candidateId: string;
  templateId: string;
  state: "pending" | "approved" | "rejected" | "changes_requested" | "withdrawn";
  reviewVersion: number;
  sourceApplicationId: string;
  sourceOwnerSubjectRef: string;
  sourceDefinitionId: string;
  sourceDefinitionVersion: number;
  sourceDefinitionDigest: string;
  title: string;
  summary: string;
  usageNotes: string;
  labels: string[];
  portability: WorkflowTemplatePortability;
  decisions: WorkflowTemplateDecision[];
  createdAt: string;
  updatedAt: string;
  requestId: string;
  auditRef: string;
};

export type WorkflowTemplateVersion = Omit<WorkflowTemplateCandidate,
  "state" | "reviewVersion" | "decisions" | "createdAt" | "updatedAt"> & {
  version: number;
  templateDigest: string;
  candidateReviewVersion: 1;
  createdAt: string;
};

export type WorkflowTemplateListingEvent = {
  eventId: string;
  templateId: string;
  decision: "list" | "replace" | "unlist";
  reason: string;
  beforePointerVersion: number;
  afterPointerVersion: number;
  beforeListedVersion: number;
  afterListedVersion: number;
  actorRef: string;
  createdAt: string;
  requestId: string;
  auditRef: string;
};

export type WorkflowTemplateLineage = {
  templateId: string;
  tenantRef: string;
  workspaceId: string;
  pointerVersion: number;
  lifecycle: "listed" | "unlisted";
  listedVersion: number;
  listedDigest: string;
  events: WorkflowTemplateListingEvent[];
  createdAt: string;
  updatedAt: string;
  requestId: string;
  auditRef: string;
};

export type WorkflowTemplatePage<T> = {
  records: T[];
  nextCursor: string;
  failureCode: string | null;
  requestId: string;
  auditRef: string;
};

export type WorkflowTemplateOperationResult = {
  candidate: WorkflowTemplateCandidate | null;
  version: WorkflowTemplateVersion | null;
  lineage: WorkflowTemplateLineage | null;
  draft: WorkflowDraftDesignerDraft | null;
  draftAuthority: {
    draftId: string;
    draftVersion: number;
    lifecycleVersion: number;
    lifecycleState: "active";
    targetApplicationId: string;
  } | null;
  failureCode: string | null;
  currentReviewVersion: number;
  currentPointerVersion: number;
  requestId: string;
  auditRef: string;
};

export type WorkflowTemplateScopeTicket = Readonly<{
  generation: number;
  scopeKey: string;
  signal: AbortSignal;
}>;

export class WorkflowTemplateRequestCoordinator {
  #generation = 0;
  #scopeKey = "";
  #controller = new AbortController();

  reset(scopeKey: string): WorkflowTemplateScopeTicket {
    this.#controller.abort();
    this.#controller = new AbortController();
    this.#scopeKey = scopeKey;
    this.#generation += 1;
    return this.current();
  }

  current(): WorkflowTemplateScopeTicket {
    return { generation: this.#generation, scopeKey: this.#scopeKey, signal: this.#controller.signal };
  }

  accepts(ticket: WorkflowTemplateScopeTicket): boolean {
    return !ticket.signal.aborted && ticket.generation === this.#generation && ticket.scopeKey === this.#scopeKey;
  }

  abort(): void {
    this.#controller.abort();
  }
}

export function readWorkflowTemplateCatalogConfig(): WorkflowTemplateCatalogConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  return {
    mode: env.VITE_RADISHMIND_WORKFLOW_TEMPLATE_SOURCE?.trim() === DEV_SOURCE
      ? "dev_workflow_template_http"
      : "offline",
    baseUrl: normalizeBaseUrl(
      env.VITE_RADISHMIND_WORKFLOW_TEMPLATE_BASE_URL ??
      env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ??
      DEFAULT_BASE_URL,
    ),
    workspaceId: env.VITE_RADISHMIND_WORKFLOW_TEMPLATE_WORKSPACE_ID?.trim() || DEFAULT_WORKSPACE_ID,
    tenantRef: env.VITE_RADISHMIND_DEV_READ_TENANT_REF?.trim() || DEFAULT_TENANT_REF,
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || DEFAULT_SUBJECT_REF,
  };
}

export async function listWorkflowTemplateCandidates(
  config: WorkflowTemplateCatalogConfig,
  query: { state?: WorkflowTemplateCandidate["state"]; limit?: number; cursor?: string; signal?: AbortSignal } = {},
): Promise<WorkflowTemplatePage<WorkflowTemplateCandidate>> {
  if (config.mode === "offline") return offlinePage();
  const parameters = listParameters(config.workspaceId, query.limit, query.cursor);
  if (query.state) parameters.set("state", query.state);
  const body = await requestJSON(config, `/v1/user-workspace/workflow-template-candidates?${parameters}`, "read", { signal: query.signal });
  return parseCandidatePage(body, config);
}

export async function readWorkflowTemplateCandidate(
  config: WorkflowTemplateCatalogConfig,
  candidateId: string,
  signal?: AbortSignal,
): Promise<WorkflowTemplateOperationResult> {
  requireId(candidateId, "candidate id");
  return requestOperation(config, `/v1/user-workspace/workflow-template-candidates/${encodeURIComponent(candidateId)}?workspace_id=${encodeURIComponent(config.workspaceId)}`, "read", { signal });
}

export async function createWorkflowTemplateCandidate(
  config: WorkflowTemplateCatalogConfig,
  input: {
    candidateId: string; templateId: string; sourceApplicationId: string; sourceDefinitionId: string;
    sourceDefinitionVersion: number; title: string; summary: string; usageNotes: string; labels: string[];
  },
  signal?: AbortSignal,
): Promise<WorkflowTemplateOperationResult> {
  return requestOperation(config, "/v1/user-workspace/workflow-template-candidates", "create", {
    method: "POST", signal, body: JSON.stringify({
      candidate_id: input.candidateId, template_id: input.templateId,
      source_application_id: input.sourceApplicationId, source_definition_id: input.sourceDefinitionId,
      source_definition_version: input.sourceDefinitionVersion, title: input.title, summary: input.summary,
      usage_notes: input.usageNotes, labels: input.labels,
    }),
  });
}

export async function reviewWorkflowTemplateCandidate(
  config: WorkflowTemplateCatalogConfig,
  candidateId: string,
  input: { expectedReviewVersion: number; decision: WorkflowTemplateDecision["decision"]; reason: string },
  signal?: AbortSignal,
): Promise<WorkflowTemplateOperationResult> {
  requireId(candidateId, "candidate id");
  return requestOperation(config, `/v1/user-workspace/workflow-template-candidates/${encodeURIComponent(candidateId)}/decisions`, "review", {
    method: "POST", signal, body: JSON.stringify({
      expected_review_version: input.expectedReviewVersion, decision: input.decision, reason: input.reason,
    }),
  });
}

export async function listWorkflowTemplates(
  config: WorkflowTemplateCatalogConfig,
  query: { limit?: number; cursor?: string; signal?: AbortSignal } = {},
): Promise<WorkflowTemplatePage<WorkflowTemplateLineage>> {
  if (config.mode === "offline") return offlinePage();
  const parameters = listParameters(config.workspaceId, query.limit, query.cursor);
  const body = await requestJSON(config, `/v1/user-workspace/workflow-templates?${parameters}`, "read", { signal: query.signal });
  return parseLineagePage(body, config);
}

export async function readWorkflowTemplate(
  config: WorkflowTemplateCatalogConfig,
  templateId: string,
  signal?: AbortSignal,
): Promise<WorkflowTemplateOperationResult> {
  requireId(templateId, "template id");
  return requestOperation(config, `/v1/user-workspace/workflow-templates/${encodeURIComponent(templateId)}?workspace_id=${encodeURIComponent(config.workspaceId)}`, "read", { signal });
}

export async function listWorkflowTemplateVersions(
  config: WorkflowTemplateCatalogConfig,
  templateId: string,
  query: { limit?: number; cursor?: string; signal?: AbortSignal } = {},
): Promise<WorkflowTemplatePage<WorkflowTemplateVersion>> {
  if (config.mode === "offline") return offlinePage();
  requireId(templateId, "template id");
  const parameters = listParameters(config.workspaceId, query.limit, query.cursor);
  const body = await requestJSON(config, `/v1/user-workspace/workflow-templates/${encodeURIComponent(templateId)}/versions?${parameters}`, "read", { signal: query.signal });
  return parseVersionPage(body, config, templateId);
}

export async function readWorkflowTemplateVersion(
  config: WorkflowTemplateCatalogConfig,
  templateId: string,
  version: number,
  signal?: AbortSignal,
): Promise<WorkflowTemplateOperationResult> {
  requireId(templateId, "template id");
  if (!positiveInteger(version)) throw new Error("template version is invalid");
  return requestOperation(config, `/v1/user-workspace/workflow-templates/${encodeURIComponent(templateId)}/versions/${version}?workspace_id=${encodeURIComponent(config.workspaceId)}`, "read", { signal });
}

export async function decideWorkflowTemplateListing(
  config: WorkflowTemplateCatalogConfig,
  templateId: string,
  input: { expectedPointerVersion: number; decision: WorkflowTemplateListingEvent["decision"]; version: number; reason: string },
  signal?: AbortSignal,
): Promise<WorkflowTemplateOperationResult> {
  requireId(templateId, "template id");
  return requestOperation(config, `/v1/user-workspace/workflow-templates/${encodeURIComponent(templateId)}/listing-decisions`, "listing", {
    method: "POST", signal, body: JSON.stringify({
      expected_pointer_version: input.expectedPointerVersion, decision: input.decision,
      version: input.version, reason: input.reason,
    }),
  });
}

export async function deriveWorkflowTemplateDraft(
  config: WorkflowTemplateCatalogConfig,
  templateId: string,
  input: {
    expectedPointerVersion: number; templateVersion: number; targetApplicationId: string;
    draftId: string; name: string; confirmed: boolean;
  },
  signal?: AbortSignal,
): Promise<WorkflowTemplateOperationResult> {
  requireId(templateId, "template id");
  return requestOperation(config, `/v1/user-workspace/workflow-templates/${encodeURIComponent(templateId)}/derivations`, "derive", {
    method: "POST", signal, body: JSON.stringify({
      expected_pointer_version: input.expectedPointerVersion, template_version: input.templateVersion,
      target_application_id: input.targetApplicationId, draft_id: input.draftId, name: input.name,
      confirmed: input.confirmed,
    }),
  }, { templateId, templateVersion: input.templateVersion, targetApplicationId: input.targetApplicationId, draftId: input.draftId });
}

type Access = "read" | "create" | "review" | "listing" | "derive";

async function requestOperation(
  config: WorkflowTemplateCatalogConfig,
  path: string,
  access: Access,
  init: RequestInit = {},
  expectedDerivation?: { templateId: string; templateVersion: number; targetApplicationId: string; draftId: string },
): Promise<WorkflowTemplateOperationResult> {
  if (config.mode === "offline") return offlineOperation();
  const body = await requestJSON(config, path, access, init);
  return parseOperation(body, config, expectedDerivation);
}

async function requestJSON(config: WorkflowTemplateCatalogConfig, path: string, access: Access, init: RequestInit): Promise<unknown> {
  const requestId = `workflow-template-${access}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
  const response = await fetch(`${normalizeBaseUrl(config.baseUrl)}${path}`, {
    ...init,
    headers: templateHeaders(config, access, requestId),
  });
  const body: unknown = await response.json();
  if (hasForbiddenResponseKey(body)) throw new Error("workflow template response contains sensitive fields");
  if (!response.ok && !isRecord(body)) throw new Error(`workflow template route returned HTTP ${response.status}`);
  return body;
}

function templateHeaders(config: WorkflowTemplateCatalogConfig, access: Access, requestId: string): HeadersInit {
  const scopes = access === "create"
    ? "workflow_definitions:read,workflow_definitions:write"
    : access === "review"
      ? "workflow_definitions:read,workflow_definitions:review"
      : access === "listing"
        ? "workflow_definitions:read,workflow_definitions:activate"
        : access === "derive"
          ? "workflow_definitions:read,workflow_drafts:write"
          : "workflow_definitions:read";
  return {
    Accept: "application/json", "Content-Type": "application/json", "X-Request-Id": requestId,
    "X-RadishMind-Dev-Read-Identity": "dev-workflow-template-consumer",
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Read-Scopes": scopes,
    "X-RadishMind-Dev-Read-Audit": "audit_dev_workflow_template_consumer",
    "X-RadishMind-Active-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Read-Membership-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Read-Membership-Permissions": scopes,
    "X-RadishMind-Dev-Workflow-Workspace": config.workspaceId,
  };
}

function parseOperation(
  value: unknown,
  config: WorkflowTemplateCatalogConfig,
  expectedDerivation?: { templateId: string; templateVersion: number; targetApplicationId: string; draftId: string },
): WorkflowTemplateOperationResult {
  const keys = ["request_id", "workspace_id", "candidate", "version", "lineage", "draft", "failure_code", "current_review_version", "current_pointer_version", "audit_ref"];
  if (!exactKeys(value, keys)) throw new Error("workflow template operation returned an invalid envelope");
  const document = value as Record<string, unknown>;
  if (!nonEmpty(document.request_id) || document.workspace_id !== config.workspaceId || !nullableString(document.failure_code) ||
      !nonNegativeInteger(document.current_review_version) || !nonNegativeInteger(document.current_pointer_version) || !nonEmpty(document.audit_ref)) {
    throw new Error("workflow template operation returned a scope-drifted envelope");
  }
  const candidate = document.candidate === null ? null : parseCandidate(document.candidate);
  const version = document.version === null ? null : parseVersion(document.version);
  const lineage = document.lineage === null ? null : parseLineage(document.lineage, config);
  let draft: WorkflowDraftDesignerDraft | null = null;
  let draftAuthority: WorkflowTemplateOperationResult["draftAuthority"] = null;
  if (document.draft !== null) {
    if (!expectedDerivation || !isStrictTemplateDerivedDraft(document.draft, config, expectedDerivation, version)) {
      throw new Error("workflow template derive returned an invalid saved draft");
    }
    const saved = document.draft as SavedWorkflowDraftDocument;
    draft = workflowDraftFromSavedWorkflowDraftDocument(saved);
    draftAuthority = {
      draftId: saved.draft_id, draftVersion: saved.draft_version, lifecycleVersion: saved.lifecycle_version,
      lifecycleState: "active", targetApplicationId: saved.application_id,
    };
  }
  return {
    candidate, version, lineage, draft, draftAuthority,
    failureCode: document.failure_code as string | null,
    currentReviewVersion: document.current_review_version as number,
    currentPointerVersion: document.current_pointer_version as number,
    requestId: document.request_id as string, auditRef: document.audit_ref as string,
  };
}

function parseCandidatePage(value: unknown, config: WorkflowTemplateCatalogConfig): WorkflowTemplatePage<WorkflowTemplateCandidate> {
  const keys = ["request_id", "workspace_id", "candidates", "next_cursor", "failure_code", "audit_ref"];
  if (!exactKeys(value, keys)) throw new Error("workflow template candidate list returned an invalid envelope");
  const document = value as Record<string, unknown>;
  if (document.workspace_id !== config.workspaceId || !Array.isArray(document.candidates)) throw new Error("workflow template candidate list scope drift");
  return pageFromDocument(document, document.candidates.map(parseCandidate), (candidate) => candidate.candidateId);
}

function parseLineagePage(value: unknown, config: WorkflowTemplateCatalogConfig): WorkflowTemplatePage<WorkflowTemplateLineage> {
  const keys = ["request_id", "workspace_id", "templates", "next_cursor", "failure_code", "audit_ref"];
  if (!exactKeys(value, keys)) throw new Error("workflow template list returned an invalid envelope");
  const document = value as Record<string, unknown>;
  if (document.workspace_id !== config.workspaceId || !Array.isArray(document.templates)) throw new Error("workflow template list scope drift");
  return pageFromDocument(document, document.templates.map((item) => parseLineage(item, config)), (lineage) => lineage.templateId);
}

function parseVersionPage(value: unknown, config: WorkflowTemplateCatalogConfig, templateId: string): WorkflowTemplatePage<WorkflowTemplateVersion> {
  const keys = ["request_id", "workspace_id", "template_id", "versions", "next_cursor", "failure_code", "audit_ref"];
  if (!exactKeys(value, keys)) throw new Error("workflow template version list returned an invalid envelope");
  const document = value as Record<string, unknown>;
  if (document.workspace_id !== config.workspaceId || document.template_id !== templateId || !Array.isArray(document.versions)) throw new Error("workflow template version list scope drift");
  const versions = document.versions.map(parseVersion);
  if (versions.some((version) => version.templateId !== templateId)) throw new Error("workflow template version list record drift");
  return pageFromDocument(document, versions, (version) => `${version.templateId}:${version.version}`);
}

function pageFromDocument<T>(document: Record<string, unknown>, records: T[], identity: (record: T) => string): WorkflowTemplatePage<T> {
  if (!nonEmpty(document.request_id) || !nonEmpty(document.audit_ref) || !nullableString(document.failure_code) || !validCursor(document.next_cursor)) {
    throw new Error("workflow template list metadata is invalid");
  }
  const identities = records.map(identity);
  if (new Set(identities).size !== identities.length) throw new Error("workflow template list contains duplicate records");
  return { records, nextCursor: document.next_cursor as string, failureCode: document.failure_code as string | null, requestId: document.request_id as string, auditRef: document.audit_ref as string };
}

function parseCandidate(value: unknown): WorkflowTemplateCandidate {
  const keys = ["schema_version", "candidate_id", "template_id", "state", "review_version", "source_application_id", "source_owner_subject_ref", "source_definition_id", "source_definition_version", "source_definition_digest", "title", "summary", "usage_notes", "labels", "portability", "decisions", "created_at", "updated_at", "created_by_actor_ref", "updated_by_actor_ref", "request_id", "audit_ref"];
  if (!exactKeys(value, keys)) throw new Error("workflow template candidate is invalid");
  const item = value as Record<string, unknown>;
  const states = ["pending", "approved", "rejected", "changes_requested", "withdrawn"];
  if (item.schema_version !== "workspace_workflow_template_candidate.v1" || !validId(item.candidate_id) || !validId(item.template_id) ||
      !states.includes(String(item.state)) || !nonNegativeInteger(item.review_version) || Number(item.review_version) > 1 ||
      !validApplicationId(item.source_application_id) || !nonEmpty(item.source_owner_subject_ref) || !validId(item.source_definition_id) ||
      !positiveInteger(item.source_definition_version) || !validDigest(item.source_definition_digest) || !boundedString(item.title, 2, 120) ||
      !boundedString(item.summary, 4, 1000) || !boundedString(item.usage_notes, 0, 2000) || !validLabels(item.labels) ||
      !Array.isArray(item.decisions) || item.decisions.length > 1 || !timestamp(item.created_at) || !timestamp(item.updated_at) ||
      !nonEmpty(item.request_id) || !nonEmpty(item.audit_ref)) throw new Error("workflow template candidate fields are invalid");
  const decisions = item.decisions.map(parseDecision);
  if (decisions.length !== Number(item.review_version)) throw new Error("workflow template candidate review authority is invalid");
  return {
    candidateId: item.candidate_id as string, templateId: item.template_id as string,
    state: item.state as WorkflowTemplateCandidate["state"], reviewVersion: item.review_version as number,
    sourceApplicationId: item.source_application_id as string, sourceOwnerSubjectRef: item.source_owner_subject_ref as string,
    sourceDefinitionId: item.source_definition_id as string, sourceDefinitionVersion: item.source_definition_version as number,
    sourceDefinitionDigest: item.source_definition_digest as string, title: item.title as string, summary: item.summary as string,
    usageNotes: item.usage_notes as string, labels: item.labels as string[], portability: parsePortability(item.portability), decisions,
    createdAt: item.created_at as string, updatedAt: item.updated_at as string, requestId: item.request_id as string, auditRef: item.audit_ref as string,
  };
}

function parseVersion(value: unknown): WorkflowTemplateVersion {
  const keys = ["schema_version", "template_id", "version", "template_digest", "candidate_id", "candidate_review_version", "source_application_id", "source_owner_subject_ref", "source_definition_id", "source_definition_version", "source_definition_digest", "title", "summary", "usage_notes", "labels", "portability", "created_at", "created_by_actor_ref", "request_id", "audit_ref"];
  if (!exactKeys(value, keys)) throw new Error("workflow template version is invalid");
  const item = value as Record<string, unknown>;
  if (item.schema_version !== "workspace_workflow_template_version.v1" || !validId(item.template_id) || !positiveInteger(item.version) ||
      !validDigest(item.template_digest) || !validId(item.candidate_id) || item.candidate_review_version !== 1 ||
      !validApplicationId(item.source_application_id) || !nonEmpty(item.source_owner_subject_ref) || !validId(item.source_definition_id) ||
      !positiveInteger(item.source_definition_version) || !validDigest(item.source_definition_digest) || !boundedString(item.title, 2, 120) ||
      !boundedString(item.summary, 4, 1000) || !boundedString(item.usage_notes, 0, 2000) || !validLabels(item.labels) ||
      !timestamp(item.created_at) || !nonEmpty(item.request_id) || !nonEmpty(item.audit_ref)) throw new Error("workflow template version fields are invalid");
  return {
    candidateId: item.candidate_id as string, templateId: item.template_id as string, version: item.version as number,
    templateDigest: item.template_digest as string, candidateReviewVersion: 1,
    sourceApplicationId: item.source_application_id as string, sourceOwnerSubjectRef: item.source_owner_subject_ref as string,
    sourceDefinitionId: item.source_definition_id as string, sourceDefinitionVersion: item.source_definition_version as number,
    sourceDefinitionDigest: item.source_definition_digest as string, title: item.title as string, summary: item.summary as string,
    usageNotes: item.usage_notes as string, labels: item.labels as string[], portability: parsePortability(item.portability),
    createdAt: item.created_at as string, requestId: item.request_id as string, auditRef: item.audit_ref as string,
  };
}

function parseLineage(value: unknown, config: WorkflowTemplateCatalogConfig): WorkflowTemplateLineage {
  const keys = ["schema_version", "template_id", "tenant_ref", "workspace_id", "pointer_version", "lifecycle", "listed_version", "listed_digest", "events", "created_at", "updated_at", "created_by_actor_ref", "updated_by_actor_ref", "request_id", "audit_ref"];
  if (!exactKeys(value, keys)) throw new Error("workflow template lineage is invalid");
  const item = value as Record<string, unknown>;
  if (item.schema_version !== "workspace_workflow_template_lineage.v1" || !validId(item.template_id) || item.tenant_ref !== config.tenantRef ||
      item.workspace_id !== config.workspaceId || !nonNegativeInteger(item.pointer_version) || !["listed", "unlisted"].includes(String(item.lifecycle)) ||
      !nonNegativeInteger(item.listed_version) || typeof item.listed_digest !== "string" || !Array.isArray(item.events) ||
      !timestamp(item.created_at) || !timestamp(item.updated_at) || !nonEmpty(item.request_id) || !nonEmpty(item.audit_ref)) throw new Error("workflow template lineage fields are invalid");
  if ((item.lifecycle === "listed" && (!positiveInteger(item.listed_version) || !validDigest(item.listed_digest))) ||
      (item.lifecycle === "unlisted" && (item.listed_version !== 0 || item.listed_digest !== ""))) throw new Error("workflow template listing authority is invalid");
  const events = item.events.map(parseListingEvent);
  if (events.length > 0 && events.at(-1)?.afterPointerVersion !== item.pointer_version) throw new Error("workflow template lineage event authority is invalid");
  return {
    templateId: item.template_id as string, tenantRef: item.tenant_ref as string, workspaceId: item.workspace_id as string,
    pointerVersion: item.pointer_version as number, lifecycle: item.lifecycle as WorkflowTemplateLineage["lifecycle"],
    listedVersion: item.listed_version as number, listedDigest: item.listed_digest as string, events,
    createdAt: item.created_at as string, updatedAt: item.updated_at as string, requestId: item.request_id as string, auditRef: item.audit_ref as string,
  };
}

function parseDecision(value: unknown): WorkflowTemplateDecision {
  const keys = ["schema_version", "review_version", "decision", "reason", "reviewer_ref", "decided_at", "request_id", "audit_ref"];
  if (!exactKeys(value, keys)) throw new Error("workflow template decision is invalid");
  const item = value as Record<string, unknown>;
  if (item.schema_version !== "workspace_workflow_template_decision.v1" || item.review_version !== 1 ||
      !["approve", "reject", "request_changes", "withdraw"].includes(String(item.decision)) || !boundedString(item.reason, 4, 500) ||
      !nonEmpty(item.reviewer_ref) || !timestamp(item.decided_at) || !nonEmpty(item.request_id) || !nonEmpty(item.audit_ref)) throw new Error("workflow template decision fields are invalid");
  return { reviewVersion: 1, decision: item.decision as WorkflowTemplateDecision["decision"], reason: item.reason as string, reviewerRef: item.reviewer_ref as string, decidedAt: item.decided_at as string, requestId: item.request_id as string, auditRef: item.audit_ref as string };
}

function parseListingEvent(value: unknown): WorkflowTemplateListingEvent {
  const keys = ["schema_version", "event_id", "template_id", "decision", "reason", "before_pointer_version", "after_pointer_version", "before_listed_version", "after_listed_version", "actor_ref", "created_at", "request_id", "audit_ref"];
  if (!exactKeys(value, keys)) throw new Error("workflow template listing event is invalid");
  const item = value as Record<string, unknown>;
  if (item.schema_version !== "workspace_workflow_template_listing_event.v1" || !validId(item.event_id) || !validId(item.template_id) ||
      !["list", "replace", "unlist"].includes(String(item.decision)) || !boundedString(item.reason, 4, 500) ||
      !nonNegativeInteger(item.before_pointer_version) || !positiveInteger(item.after_pointer_version) ||
      Number(item.after_pointer_version) !== Number(item.before_pointer_version) + 1 || !nonNegativeInteger(item.before_listed_version) ||
      !nonNegativeInteger(item.after_listed_version) || !nonEmpty(item.actor_ref) || !timestamp(item.created_at) || !nonEmpty(item.request_id) || !nonEmpty(item.audit_ref)) throw new Error("workflow template listing event fields are invalid");
  return { eventId: item.event_id as string, templateId: item.template_id as string, decision: item.decision as WorkflowTemplateListingEvent["decision"], reason: item.reason as string, beforePointerVersion: item.before_pointer_version as number, afterPointerVersion: item.after_pointer_version as number, beforeListedVersion: item.before_listed_version as number, afterListedVersion: item.after_listed_version as number, actorRef: item.actor_ref as string, createdAt: item.created_at as string, requestId: item.request_id as string, auditRef: item.audit_ref as string };
}

function parsePortability(value: unknown): WorkflowTemplatePortability {
  const keys = ["execution_profile", "node_kinds", "provider_refs", "risk_level", "portable", "blockers"];
  if (!exactKeys(value, keys)) throw new Error("workflow template portability is invalid");
  const item = value as Record<string, unknown>;
  const kinds = ["prompt", "llm", "condition", "output"];
  if (item.execution_profile !== "workflow_definition_executor_v1" || !Array.isArray(item.node_kinds) || item.node_kinds.length < 1 ||
      item.node_kinds.some((kind) => !kinds.includes(String(kind))) || new Set(item.node_kinds).size !== item.node_kinds.length ||
      !Array.isArray(item.provider_refs) || item.provider_refs.some((ref) => typeof ref !== "string" || !/^profile:[A-Za-z0-9][A-Za-z0-9._:-]{0,151}$/u.test(ref)) ||
      new Set(item.provider_refs).size !== item.provider_refs.length || !["low", "medium", "high"].includes(String(item.risk_level)) ||
      item.portable !== true || !Array.isArray(item.blockers) || item.blockers.length !== 0) throw new Error("workflow template portability fields are invalid");
  return { executionProfile: "workflow_definition_executor_v1", nodeKinds: item.node_kinds as WorkflowTemplatePortability["nodeKinds"], providerRefs: item.provider_refs as string[], riskLevel: item.risk_level as WorkflowTemplatePortability["riskLevel"], portable: true, blockers: [] };
}

function isStrictTemplateDerivedDraft(
  value: unknown,
  config: WorkflowTemplateCatalogConfig,
  expected: { templateId: string; templateVersion: number; targetApplicationId: string; draftId: string },
  version: WorkflowTemplateVersion | null,
): boolean {
  const keys = ["draft_id", "workspace_id", "application_id", "source_definition_id", "base_definition_version", "schema_version", "draft_status", "name", "description", "nodes", "edges", "input_contract", "output_contract", "provider_refs", "tool_refs", "rag_refs", "requested_capabilities", "additional_fields", "draft_version", "lifecycle_state", "lifecycle_version", "archived_at", "library_updated_at", "lifecycle_updated_by_actor_ref", "provenance_kind", "created_at", "updated_at", "created_by_actor_ref", "updated_by_actor_ref", "validation_summary", "blocked_capability_summary", "request_audit_metadata", "sample_or_unsaved_draft_status"];
  if (!exactKeys(value, keys)) return false;
  const draft = value as Record<string, unknown>;
  if (draft.draft_id !== expected.draftId || draft.workspace_id !== config.workspaceId || draft.application_id !== expected.targetApplicationId ||
      draft.draft_version !== 1 || draft.lifecycle_state !== "active" || draft.lifecycle_version !== 1 || draft.archived_at !== null ||
      draft.provenance_kind !== "workspace_template_derivation" || draft.schema_version !== "saved_workflow_draft.v1" ||
      !positiveInteger(draft.base_definition_version) || !boundedString(draft.name, 2, 120) || typeof draft.description !== "string" ||
      !Array.isArray(draft.nodes) || !draft.nodes.every(isStrictSavedDraftNode) || duplicateBy(draft.nodes, "node_id") ||
      !Array.isArray(draft.edges) || !draft.edges.every(isStrictSavedDraftEdge) || duplicateBy(draft.edges, "edge_id") ||
      !isStrictSavedDraftContract(draft.input_contract) || !isStrictSavedDraftContract(draft.output_contract) ||
      !stringArray(draft.provider_refs) || !stringArray(draft.tool_refs) || !stringArray(draft.rag_refs) || !stringArray(draft.requested_capabilities) ||
      !timestamp(draft.library_updated_at) || typeof draft.lifecycle_updated_by_actor_ref !== "string" || !timestamp(draft.created_at) || !timestamp(draft.updated_at) ||
      !nonEmpty(draft.created_by_actor_ref) || !nonEmpty(draft.updated_by_actor_ref) || !isStrictValidationSummary(draft.validation_summary) ||
      !Array.isArray(draft.blocked_capability_summary) || !draft.blocked_capability_summary.every(isStrictBlockedCapability) ||
      !isStrictAuditMetadata(draft.request_audit_metadata) || typeof draft.sample_or_unsaved_draft_status !== "string" ||
      !exactKeys(draft.additional_fields, ["derivation_v2"])) return false;
  const metadata = (draft.additional_fields as Record<string, unknown>).derivation_v2;
  const metadataKeys = ["version", "source_kind", "template_id", "template_version", "template_digest", "source_definition_id", "source_definition_version", "source_definition_digest"];
  if (!exactKeys(metadata, metadataKeys)) return false;
  const derivation = metadata as Record<string, unknown>;
  return derivation.version === 2 && derivation.source_kind === "workspace_workflow_template" &&
    derivation.template_id === expected.templateId && derivation.template_version === expected.templateVersion &&
    validDigest(derivation.template_digest) && derivation.source_definition_id === draft.source_definition_id &&
    derivation.source_definition_version === draft.base_definition_version && validDigest(derivation.source_definition_digest) &&
    (version === null || (version.templateId === expected.templateId && version.version === expected.templateVersion &&
      version.templateDigest === derivation.template_digest && version.sourceDefinitionId === derivation.source_definition_id &&
      version.sourceDefinitionVersion === derivation.source_definition_version && version.sourceDefinitionDigest === derivation.source_definition_digest));
}

function isStrictSavedDraftNode(value: unknown): boolean {
  const keys = ["node_id", "node_type", "label", "input_summary", "output_summary", "input_contract_ref", "output_contract_ref", "input_contract_fields", "output_contract_fields", "output_mapping_summary", "provider_ref", "tool_ref", "rag_ref", "risk_level", "requires_confirmation"];
  if (!exactKeys(value, keys)) return false;
  const node = value as Record<string, unknown>;
  return validId(node.node_id) && ["prompt", "llm", "condition", "output"].includes(String(node.node_type)) && nonEmpty(node.label) &&
    typeof node.input_summary === "string" && typeof node.output_summary === "string" && nonEmpty(node.input_contract_ref) &&
    nonEmpty(node.output_contract_ref) && stringArray(node.input_contract_fields) && stringArray(node.output_contract_fields) &&
    typeof node.output_mapping_summary === "string" && typeof node.provider_ref === "string" && typeof node.tool_ref === "string" &&
    typeof node.rag_ref === "string" && ["low", "medium", "high"].includes(String(node.risk_level)) && typeof node.requires_confirmation === "boolean";
}

function isStrictSavedDraftEdge(value: unknown): boolean {
  if (!exactKeys(value, ["edge_id", "from_node_id", "to_node_id", "condition_summary"])) return false;
  const edge = value as Record<string, unknown>;
  return validId(edge.edge_id) && validId(edge.from_node_id) && validId(edge.to_node_id) && typeof edge.condition_summary === "string";
}

function isStrictSavedDraftContract(value: unknown): boolean {
  if (!exactKeys(value, ["contract_id", "summary"]) && !exactKeys(value, ["contract_id", "required_fields", "summary"])) return false;
  const contract = value as Record<string, unknown>;
  return validId(contract.contract_id) &&
    (contract.required_fields === undefined || stringArray(contract.required_fields)) &&
    typeof contract.summary === "string";
}

function isStrictValidationSummary(value: unknown): boolean {
  if (!exactKeys(value, ["validation_state", "valid_for_review", "findings"])) return false;
  const summary = value as Record<string, unknown>;
  return ["valid_for_review", "invalid_draft", "blocked_capability", "schema_unsupported"].includes(String(summary.validation_state)) &&
    typeof summary.valid_for_review === "boolean" && Array.isArray(summary.findings) && summary.findings.every((finding) => {
      if (!exactKeys(finding, ["code", "severity", "field", "summary", "evidence_id"])) return false;
      return Object.values(finding).every((item) => typeof item === "string");
    });
}

function isStrictBlockedCapability(value: unknown): boolean {
  if (!isRecord(value)) return false;
  const keys = Object.keys(value).sort().join(",");
  if (keys !== "capability_id" && keys !== "capability_id,missing_prerequisite" && keys !== "capability_id,summary" && keys !== "capability_id,missing_prerequisite,summary") return false;
  return nonEmpty(value.capability_id) && (value.missing_prerequisite === undefined || typeof value.missing_prerequisite === "string") &&
    (value.summary === undefined || typeof value.summary === "string");
}

function isStrictAuditMetadata(value: unknown): boolean {
  if (!exactKeys(value, ["request_id", "audit_ref", "actor_ref"])) return false;
  const audit = value as Record<string, unknown>;
  return nonEmpty(audit.request_id) && nonEmpty(audit.audit_ref) && nonEmpty(audit.actor_ref);
}

function duplicateBy(values: unknown[], key: string): boolean {
  const identities = values.map((value) => isRecord(value) ? value[key] : undefined);
  return new Set(identities).size !== identities.length;
}

function listParameters(workspaceId: string, limit?: number, cursor?: string): URLSearchParams {
  const parameters = new URLSearchParams({ workspace_id: workspaceId });
  if (limit !== undefined) {
    if (!Number.isInteger(limit) || limit < 1 || limit > 100) throw new Error("workflow template list limit is invalid");
    parameters.set("limit", String(limit));
  }
  if (cursor) {
    if (!validCursor(cursor)) throw new Error("workflow template list cursor is invalid");
    parameters.set("cursor", cursor);
  }
  return parameters;
}

function offlinePage<T>(): WorkflowTemplatePage<T> {
  return { records: [], nextCursor: "", failureCode: null, requestId: "workflow-template-offline", auditRef: "audit_workflow_template_offline" };
}

function offlineOperation(): WorkflowTemplateOperationResult {
  return { candidate: null, version: null, lineage: null, draft: null, draftAuthority: null, failureCode: "workflow_template_offline", currentReviewVersion: 0, currentPointerVersion: 0, requestId: "workflow-template-offline", auditRef: "audit_workflow_template_offline" };
}

function requireId(value: string, label: string): void { if (!ID_PATTERN.test(value)) throw new Error(`${label} is invalid`); }
function validId(value: unknown): value is string { return typeof value === "string" && ID_PATTERN.test(value); }
function validApplicationId(value: unknown): value is string { return typeof value === "string" && APPLICATION_ID_PATTERN.test(value); }
function validDigest(value: unknown): value is string { return typeof value === "string" && DIGEST_PATTERN.test(value); }
function validCursor(value: unknown): value is string { return value === "" || (typeof value === "string" && value.length <= 4096 && CURSOR_PATTERN.test(value)); }
function nonEmpty(value: unknown): value is string { return typeof value === "string" && value.trim().length > 0; }
function boundedString(value: unknown, min: number, max: number): value is string { return typeof value === "string" && [...value].length >= min && [...value].length <= max; }
function timestamp(value: unknown): value is string { return typeof value === "string" && value.length > 0 && Number.isFinite(Date.parse(value)); }
function positiveInteger(value: unknown): value is number { return Number.isInteger(value) && Number(value) >= 1; }
function nonNegativeInteger(value: unknown): value is number { return Number.isInteger(value) && Number(value) >= 0; }
function nullableString(value: unknown): value is string | null { return value === null || nonEmpty(value); }
function stringArray(value: unknown): value is string[] { return Array.isArray(value) && value.every((item) => typeof item === "string"); }
function validLabels(value: unknown): value is string[] { return Array.isArray(value) && value.length <= 8 && value.every((label) => typeof label === "string" && /^[a-z0-9][a-z0-9._:-]{0,39}$/u.test(label)) && new Set(value).size === value.length; }
function isRecord(value: unknown): value is Record<string, unknown> { return Boolean(value) && typeof value === "object" && !Array.isArray(value); }
function exactKeys(value: unknown, keys: readonly string[]): value is Record<string, unknown> { if (!isRecord(value)) return false; const actual = Object.keys(value).sort(); const expected = [...keys].sort(); return actual.length === expected.length && actual.every((key, index) => key === expected[index]); }
function hasForbiddenResponseKey(value: unknown): boolean { if (Array.isArray(value)) return value.some(hasForbiddenResponseKey); if (!isRecord(value)) return false; return Object.entries(value).some(([key, nested]) => FORBIDDEN_RESPONSE_KEYS.has(key.toLowerCase()) || hasForbiddenResponseKey(nested)); }
function normalizeBaseUrl(value: string): string { const normalized = value.trim() || DEFAULT_BASE_URL; return normalized.endsWith("/") ? normalized.slice(0, -1) : normalized; }
