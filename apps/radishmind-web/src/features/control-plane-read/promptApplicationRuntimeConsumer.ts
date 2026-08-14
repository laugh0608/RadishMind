import type { ApplicationPublishPromptTemplateRef } from "./applicationPublishCandidateConsumer.ts";

const DEV_SOURCE = "dev-prompt-application-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";

export type PromptApplicationRuntimeConfig = {
  mode: "offline" | "dev_prompt_application_http";
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  subjectRef: string;
};

export type PromptApplicationRuntimeAction = "activate" | "replace" | "revoke";

export type PromptApplicationRuntimeAssignment = {
  assignmentId: string;
  assignmentVersion: number;
  state: "active" | "revoked";
  candidateId: string;
  candidateReviewVersion: number;
  draftId: string;
  draftVersion: number;
  draftDigest: string;
  promptTemplateRef: ApplicationPublishPromptTemplateRef;
  assignmentDigest: string;
  activatedAt: string;
  updatedAt: string;
  revokedAt: string | null;
  updatedByActorRef: string;
};

export type PromptApplicationRuntimeEvent = {
  eventId: string;
  eventSequence: number;
  action: PromptApplicationRuntimeAction;
  expectedAssignmentVersion: number;
  resultingAssignmentVersion: number;
  candidateId: string;
  promptTemplateRef: ApplicationPublishPromptTemplateRef;
  assignmentDigest: string;
  occurredAt: string;
  actorRef: string;
};

export type PromptApplicationRuntimeState = {
  status: "offline" | "idle" | "loading" | "ready" | "not_found" | "version_conflict" | "blocked" | "failed";
  assignment: PromptApplicationRuntimeAssignment | null;
  events: PromptApplicationRuntimeEvent[];
  failureCode: string;
  currentAssignmentVersion: number;
  currentState: string;
  requestId: string;
  auditRef: string;
  summary: string;
};

type PromptTemplateRefDocument = {
  template_id: string;
  template_version: number;
  template_digest: string;
};

type AssignmentDocument = {
  schema_version: string;
  assignment_id: string;
  tenant_ref: string;
  workspace_id: string;
  application_id: string;
  owner_subject_ref: string;
  assignment_version: number;
  state: string;
  candidate_id: string;
  candidate_review_version: number;
  draft_id: string;
  draft_version: number;
  draft_digest: string;
  prompt_template_ref: PromptTemplateRefDocument;
  assignment_digest: string;
  activated_at: string;
  updated_at: string;
  revoked_at: string | null;
  activated_by_actor_ref: string;
  updated_by_actor_ref: string;
  request_id: string;
  audit_ref: string;
};

type EventDocument = {
  schema_version: string;
  event_id: string;
  assignment_id: string;
  tenant_ref: string;
  workspace_id: string;
  application_id: string;
  owner_subject_ref: string;
  event_sequence: number;
  action: string;
  expected_assignment_version: number;
  resulting_assignment_version: number;
  candidate_id: string;
  candidate_review_version: number;
  draft_id: string;
  draft_version: number;
  draft_digest: string;
  prompt_template_ref: PromptTemplateRefDocument;
  assignment_digest: string;
  occurred_at: string;
  actor_ref: string;
  request_id: string;
  audit_ref: string;
};

type RuntimeEnvelope = {
  request_id: string;
  tenant_ref: string;
  workspace_id: string;
  application_id: string;
  assignment: AssignmentDocument | null;
  events: EventDocument[];
  failure_code: string | null;
  current_assignment_version: number;
  current_state: string;
  audit_ref: string;
};

export function readPromptApplicationRuntimeConfig(): PromptApplicationRuntimeConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  return {
    mode: env.VITE_RADISHMIND_PROMPT_APPLICATION_SOURCE?.trim() === DEV_SOURCE
      ? "dev_prompt_application_http"
      : "offline",
    baseUrl: (
      env.VITE_RADISHMIND_PROMPT_APPLICATION_BASE_URL ??
      env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ??
      DEFAULT_BASE_URL
    ).trim().replace(/\/$/u, ""),
    tenantRef: env.VITE_RADISHMIND_DEV_READ_TENANT_REF?.trim() || "tenant_demo",
    workspaceId: env.VITE_RADISHMIND_PROMPT_APPLICATION_WORKSPACE_ID?.trim() || "workspace_demo",
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || "subject_demo_user",
  };
}

export function initialPromptApplicationRuntimeState(
  config: PromptApplicationRuntimeConfig,
): PromptApplicationRuntimeState {
  return {
    status: config.mode === "offline" ? "offline" : "idle",
    assignment: null,
    events: [],
    failureCode: config.mode === "offline" ? "prompt_application_runtime_http_disabled" : "",
    currentAssignmentVersion: 0,
    currentState: "",
    requestId: "",
    auditRef: "",
    summary: config.mode === "offline"
      ? "Prompt Runtime Assignment owner is offline."
      : "Load the current exact Prompt runtime authority before making a decision.",
  };
}

export async function readPromptApplicationRuntime(
  config: PromptApplicationRuntimeConfig,
  applicationId: string,
  includeEvents = true,
): Promise<PromptApplicationRuntimeState> {
  if (config.mode === "offline") return initialPromptApplicationRuntimeState(config);
  const suffix = includeEvents ? "/events" : "";
  return requestRuntime(
    config,
    applicationId,
    `/v1/user-workspace/applications/${encodeURIComponent(applicationId)}/prompt-runtime-assignment${suffix}?workspace_id=${encodeURIComponent(config.workspaceId)}`,
    "GET",
  );
}

export async function decidePromptApplicationRuntime(
  config: PromptApplicationRuntimeConfig,
  applicationId: string,
  expectedAssignmentVersion: number,
  action: PromptApplicationRuntimeAction,
  candidateId: string,
): Promise<PromptApplicationRuntimeState> {
  if (config.mode === "offline") return initialPromptApplicationRuntimeState(config);
  return requestRuntime(
    config,
    applicationId,
    `/v1/user-workspace/applications/${encodeURIComponent(applicationId)}/prompt-runtime-assignment/decisions`,
    "POST",
    {
      workspace_id: config.workspaceId,
      expected_assignment_version: expectedAssignmentVersion,
      action,
      candidate_id: action === "revoke" ? "" : candidateId.trim(),
    },
  );
}

async function requestRuntime(
  config: PromptApplicationRuntimeConfig,
  applicationId: string,
  path: string,
  method: "GET" | "POST",
  body?: unknown,
): Promise<PromptApplicationRuntimeState> {
  const requestId = createRequestId("prompt-runtime");
  try {
    const response = await fetch(`${config.baseUrl}${path}`, {
      method,
      headers: {
        Accept: "application/json",
        ...(method === "POST" ? { "Content-Type": "application/json" } : {}),
        "X-Request-Id": requestId,
        "X-RadishMind-Dev-Read-Identity": "radishmind-web-prompt-runtime-dev",
        "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
        "X-RadishMind-Dev-Read-Subject": config.subjectRef,
        "X-RadishMind-Dev-Read-Scopes": method === "POST"
          ? "prompt_application_runtime:read,prompt_application_runtime:write"
          : "prompt_application_runtime:read",
        "X-RadishMind-Dev-Read-Audit": `audit-${requestId}`,
        ...(method === "POST" ? {
          "X-RadishMind-Active-Workspace": config.workspaceId,
          "X-RadishMind-Dev-Read-Membership-Workspace": config.workspaceId,
          "X-RadishMind-Dev-Read-Membership-Permissions": "prompt_application_runtime:write",
        } : {}),
        "X-RadishMind-Dev-Prompt-Runtime-Workspace": config.workspaceId,
        "X-RadishMind-Dev-Prompt-Runtime-Application": applicationId,
      },
      ...(body === undefined ? {} : { body: JSON.stringify(body) }),
    });
    const document: unknown = await response.json();
    if (!response.ok || !isRuntimeEnvelope(document, config, applicationId)) {
      throw new Error("invalid Prompt runtime response");
    }
    return mapRuntimeEnvelope(document);
  } catch {
    return {
      ...initialPromptApplicationRuntimeState(config),
      status: "failed",
      failureCode: "prompt_runtime_store_unavailable",
      summary: "Prompt Runtime Assignment could not be loaded or changed.",
    };
  }
}

function mapRuntimeEnvelope(document: RuntimeEnvelope): PromptApplicationRuntimeState {
  const failureCode = document.failure_code ?? "";
  const status = failureCode === "prompt_runtime_assignment_not_found"
    ? "not_found"
    : failureCode === "prompt_runtime_assignment_version_conflict"
      ? "version_conflict"
      : failureCode
        ? "blocked"
        : "ready";
  return {
    status,
    assignment: document.assignment ? mapAssignment(document.assignment) : null,
    events: document.events.map(mapEvent),
    failureCode,
    currentAssignmentVersion: document.current_assignment_version,
    currentState: document.current_state,
    requestId: document.request_id,
    auditRef: document.audit_ref,
    summary: failureCode
      ? "Runtime owner closed the operation without changing authority."
      : document.assignment
        ? `Loaded ${document.assignment.state} assignment v${document.assignment.assignment_version}.`
        : "No Prompt Runtime Assignment exists.",
  };
}

function mapAssignment(document: AssignmentDocument): PromptApplicationRuntimeAssignment {
  return {
    assignmentId: document.assignment_id,
    assignmentVersion: document.assignment_version,
    state: document.state as PromptApplicationRuntimeAssignment["state"],
    candidateId: document.candidate_id,
    candidateReviewVersion: document.candidate_review_version,
    draftId: document.draft_id,
    draftVersion: document.draft_version,
    draftDigest: document.draft_digest,
    promptTemplateRef: mapPromptTemplateRef(document.prompt_template_ref),
    assignmentDigest: document.assignment_digest,
    activatedAt: document.activated_at,
    updatedAt: document.updated_at,
    revokedAt: document.revoked_at,
    updatedByActorRef: document.updated_by_actor_ref,
  };
}

function mapEvent(document: EventDocument): PromptApplicationRuntimeEvent {
  return {
    eventId: document.event_id,
    eventSequence: document.event_sequence,
    action: document.action as PromptApplicationRuntimeAction,
    expectedAssignmentVersion: document.expected_assignment_version,
    resultingAssignmentVersion: document.resulting_assignment_version,
    candidateId: document.candidate_id,
    promptTemplateRef: mapPromptTemplateRef(document.prompt_template_ref),
    assignmentDigest: document.assignment_digest,
    occurredAt: document.occurred_at,
    actorRef: document.actor_ref,
  };
}

function isRuntimeEnvelope(
  value: unknown,
  config: PromptApplicationRuntimeConfig,
  applicationId: string,
): value is RuntimeEnvelope {
  if (!isRecord(value) || value.tenant_ref !== config.tenantRef || value.workspace_id !== config.workspaceId ||
    value.application_id !== applicationId || !isNonEmptyString(value.request_id) || !isNonEmptyString(value.audit_ref) ||
    !Number.isInteger(value.current_assignment_version) || value.current_assignment_version < 0 ||
    typeof value.current_state !== "string" || (value.failure_code !== null && !isNonEmptyString(value.failure_code)) ||
    !Array.isArray(value.events) || !value.events.every((event) => isEvent(event, config, applicationId)) ||
    (value.assignment !== null && !isAssignment(value.assignment, config, applicationId))) return false;
  if (value.assignment && (
    value.current_assignment_version !== value.assignment.assignment_version ||
    value.current_state !== value.assignment.state
  )) return false;
  return value.events.every((event, index) =>
    event.event_sequence === index + 1 &&
    event.resulting_assignment_version === event.event_sequence
  );
}

function isAssignment(value: unknown, config: PromptApplicationRuntimeConfig, applicationId: string): value is AssignmentDocument {
  return isRecord(value) && value.schema_version === "prompt_application_runtime_assignment.v1" &&
    value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId && value.application_id === applicationId &&
    /^ptra_[a-z2-7]{16}$/u.test(String(value.assignment_id)) &&
    Number.isInteger(value.assignment_version) && value.assignment_version >= 1 &&
    (value.state === "active" || value.state === "revoked") &&
    isNonEmptyString(value.candidate_id) && Number.isInteger(value.candidate_review_version) && value.candidate_review_version >= 1 &&
    isNonEmptyString(value.draft_id) && Number.isInteger(value.draft_version) && value.draft_version >= 1 &&
    isDigest(value.draft_digest) && isPromptTemplateRef(value.prompt_template_ref) && isDigest(value.assignment_digest) &&
    isNonEmptyString(value.activated_at) && isNonEmptyString(value.updated_at) &&
    (value.revoked_at === null || isNonEmptyString(value.revoked_at)) &&
    isNonEmptyString(value.updated_by_actor_ref);
}

function isEvent(value: unknown, config: PromptApplicationRuntimeConfig, applicationId: string): value is EventDocument {
  return isRecord(value) && value.schema_version === "prompt_application_runtime_assignment_event.v1" &&
    value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId && value.application_id === applicationId &&
    /^ptrae_[a-z2-7]{16}$/u.test(String(value.event_id)) && /^ptra_[a-z2-7]{16}$/u.test(String(value.assignment_id)) &&
    Number.isInteger(value.event_sequence) && value.event_sequence >= 1 &&
    (value.action === "activate" || value.action === "replace" || value.action === "revoke") &&
    Number.isInteger(value.expected_assignment_version) && value.expected_assignment_version >= 0 &&
    Number.isInteger(value.resulting_assignment_version) && value.resulting_assignment_version >= 1 &&
    isNonEmptyString(value.candidate_id) && isPromptTemplateRef(value.prompt_template_ref) &&
    isDigest(value.assignment_digest) && isNonEmptyString(value.occurred_at) && isNonEmptyString(value.actor_ref);
}

function mapPromptTemplateRef(value: PromptTemplateRefDocument): ApplicationPublishPromptTemplateRef {
  return {
    templateId: value.template_id,
    templateVersion: value.template_version,
    templateDigest: value.template_digest,
  };
}

function isPromptTemplateRef(value: unknown): value is PromptTemplateRefDocument {
  return isRecord(value) && /^ptpl_[a-z2-7]{16}$/u.test(String(value.template_id)) &&
    Number.isInteger(value.template_version) && value.template_version >= 1 &&
    isDigest(value.template_digest);
}

function isDigest(value: unknown): value is string {
  return typeof value === "string" && /^sha256:[a-f0-9]{64}$/u.test(value);
}

function isNonEmptyString(value: unknown): value is string {
  return typeof value === "string" && value.trim().length > 0;
}

function isRecord(value: unknown): value is Record<string, any> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function createRequestId(prefix: string): string {
  return `${prefix}-${Date.now()}-${(globalThis.crypto?.randomUUID?.() ?? Math.random().toString(16).slice(2)).replaceAll("-", "").slice(0, 12)}`;
}
