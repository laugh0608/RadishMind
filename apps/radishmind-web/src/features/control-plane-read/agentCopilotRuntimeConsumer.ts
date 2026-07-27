import type { AgentCopilotProfileVersionSummary } from "./agentCopilotProfileConsumer.ts";

const DEV_SOURCE = "dev-agent-copilot-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const DIGEST = /^sha256:[a-f0-9]{64}$/u;
const FORBIDDEN_RESPONSE_KEYS = new Set([
  "authorization", "api_key", "credential", "endpoint", "headers", "cookie", "dsn",
  "raw_request", "raw_response", "provider_raw_response", "input", "output", "prompt", "messages",
]);

export type AgentCopilotRuntimeConfig = {
  mode: "offline" | "dev_agent_copilot_http";
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  subjectRef: string;
};
export type AgentCopilotRuntimeAssignment = {
  assignmentId: string;
  assignmentVersion: number;
  state: "active" | "revoked";
  candidateId: string;
  candidateReviewVersion: number;
  draftId: string;
  draftVersion: number;
  draftDigest: string;
  profile: Pick<AgentCopilotProfileVersionSummary, "profileId" | "profileVersion" | "profileDigest" | "policyDigest">;
  assignmentDigest: string;
  activatedAt: string;
  updatedAt: string;
  revokedAt: string;
};
export type AgentCopilotRuntimeEvent = {
  eventId: string;
  eventSequence: number;
  action: "activate" | "replace" | "revoke";
  expectedAssignmentVersion: number;
  resultingAssignmentVersion: number;
  candidateId: string;
  occurredAt: string;
};
export type AgentCopilotRuntimeState = {
  status: "offline" | "idle" | "ready" | "not_found" | "version_conflict" | "blocked" | "failed";
  assignment: AgentCopilotRuntimeAssignment | null;
  events: AgentCopilotRuntimeEvent[];
  currentAssignmentVersion: number;
  currentState: string;
  failureCode: string;
  summary: string;
};

type Document = Record<string, any>;

export function readAgentCopilotRuntimeConfig(): AgentCopilotRuntimeConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  return {
    mode: env.VITE_RADISHMIND_AGENT_COPILOT_SOURCE?.trim() === DEV_SOURCE
      ? "dev_agent_copilot_http"
      : "offline",
    baseUrl: (
      env.VITE_RADISHMIND_AGENT_COPILOT_BASE_URL ??
      env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ??
      DEFAULT_BASE_URL
    ).trim().replace(/\/$/u, ""),
    tenantRef: env.VITE_RADISHMIND_DEV_READ_TENANT_REF?.trim() || "tenant_demo",
    workspaceId: env.VITE_RADISHMIND_AGENT_COPILOT_WORKSPACE_ID?.trim() || "workspace_demo",
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || "subject_demo_user",
  };
}

export function initialAgentCopilotRuntimeState(config: AgentCopilotRuntimeConfig): AgentCopilotRuntimeState {
  return {
    status: config.mode === "offline" ? "offline" : "idle",
    assignment: null,
    events: [],
    currentAssignmentVersion: 0,
    currentState: "none",
    failureCode: config.mode === "offline" ? "agent_copilot_runtime_http_disabled" : "",
    summary: config.mode === "offline"
      ? "Agent Copilot runtime Web 未启用。"
      : "加载当前 assignment，或对 approved candidate 作显式决定。",
  };
}

export async function readAgentCopilotRuntime(
  config: AgentCopilotRuntimeConfig,
  applicationId: string,
): Promise<AgentCopilotRuntimeState> {
  return requestRuntime(config, applicationId, "GET", "events", null, "read");
}

export async function decideAgentCopilotRuntime(
  config: AgentCopilotRuntimeConfig,
  applicationId: string,
  expectedAssignmentVersion: number,
  action: "activate" | "replace" | "revoke",
  candidateId: string,
): Promise<AgentCopilotRuntimeState> {
  return requestRuntime(config, applicationId, "POST", "decisions", {
    workspace_id: config.workspaceId,
    expected_assignment_version: expectedAssignmentVersion,
    action,
    candidate_id: action === "revoke" ? "" : candidateId.trim(),
  }, "write");
}

async function requestRuntime(
  config: AgentCopilotRuntimeConfig,
  applicationId: string,
  method: "GET" | "POST",
  suffix: "events" | "decisions",
  body: unknown,
  operation: "read" | "write",
): Promise<AgentCopilotRuntimeState> {
  if (config.mode === "offline") return initialAgentCopilotRuntimeState(config);
  const requestId = createRequestId(`agent-runtime-${operation}`);
  const query = method === "GET" ? `?workspace_id=${encodeURIComponent(config.workspaceId)}` : "";
  try {
    const response = await fetch(
      `${config.baseUrl}/v1/user-workspace/applications/${encodeURIComponent(applicationId)}/agent-copilot-runtime-assignment/${suffix}${query}`,
      {
        method,
        headers: {
          ...runtimeHeaders(config, applicationId, requestId, operation),
          ...(method === "POST" ? { "Content-Type": "application/json" } : {}),
        },
        body: body === null ? undefined : JSON.stringify(body),
      },
    );
    const value: unknown = await response.json();
    if (!response.ok || !isEnvelope(value, config, applicationId)) throw new Error("invalid runtime response");
    return mapRuntime(value);
  } catch {
    return {
      ...initialAgentCopilotRuntimeState(config),
      status: "failed",
      failureCode: "agent_copilot_runtime_store_unavailable",
      summary: "Agent Copilot runtime owner 不可用；未回退到内存数据。",
    };
  }
}

function mapRuntime(value: Document): AgentCopilotRuntimeState {
  const failureCode = typeof value.failure_code === "string" ? value.failure_code : "";
  const assignment = isAssignment(value.assignment) ? mapAssignment(value.assignment) : null;
  const events = Array.isArray(value.events) && value.events.every(isEvent)
    ? value.events.map(mapEvent)
    : [];
  return {
    status: failureCode === "agent_copilot_runtime_assignment_not_found"
      ? "not_found"
      : failureCode === "agent_copilot_runtime_assignment_version_conflict"
        ? "version_conflict"
        : failureCode
          ? failureCode.includes("ineligible") || failureCode.includes("authority") ? "blocked" : "failed"
          : "ready",
    assignment,
    events,
    currentAssignmentVersion: Number(value.current_assignment_version),
    currentState: String(value.current_state),
    failureCode,
    summary: failureCode
      ? `Agent Copilot assignment 操作失败：${failureCode}。`
      : assignment
        ? `当前 assignment v${assignment.assignmentVersion} 为 ${assignment.state}。`
        : "当前没有 Agent Copilot assignment。",
  };
}

function mapAssignment(value: Document): AgentCopilotRuntimeAssignment {
  return {
    assignmentId: value.assignment_id,
    assignmentVersion: value.assignment_version,
    state: value.state,
    candidateId: value.candidate_id,
    candidateReviewVersion: value.candidate_review_version,
    draftId: value.draft_id,
    draftVersion: value.draft_version,
    draftDigest: value.draft_digest,
    profile: {
      profileId: value.agent_copilot_profile_ref.profile_id,
      profileVersion: value.agent_copilot_profile_ref.profile_version,
      profileDigest: value.agent_copilot_profile_ref.profile_digest,
      policyDigest: value.agent_copilot_profile_ref.policy_digest,
    },
    assignmentDigest: value.assignment_digest,
    activatedAt: value.activated_at,
    updatedAt: value.updated_at,
    revokedAt: value.revoked_at ?? "",
  };
}

function mapEvent(value: Document): AgentCopilotRuntimeEvent {
  return {
    eventId: value.event_id,
    eventSequence: value.event_sequence,
    action: value.action,
    expectedAssignmentVersion: value.expected_assignment_version,
    resultingAssignmentVersion: value.resulting_assignment_version,
    candidateId: value.candidate_id,
    occurredAt: value.occurred_at,
  };
}

function isEnvelope(value: unknown, config: AgentCopilotRuntimeConfig, applicationId: string): value is Document {
  if (!isRecord(value) || containsForbiddenResponse(value) || value.workspace_id !== config.workspaceId ||
    value.application_id !== applicationId ||
    typeof value.request_id !== "string" || typeof value.audit_ref !== "string" ||
    (value.failure_code !== null && typeof value.failure_code !== "string") ||
    !Number.isInteger(value.current_assignment_version) ||
    typeof value.current_state !== "string" ||
    (value.assignment !== null && !isAssignment(value.assignment)) ||
    !Array.isArray(value.events) || !value.events.every(isEvent)) return false;
  if (value.assignment && (value.assignment.workspace_id !== config.workspaceId ||
      value.assignment.application_id !== applicationId)) return false;
  if (value.events.some((event: Document) =>
    event.workspace_id !== config.workspaceId || event.application_id !== applicationId
  )) return false;
  const sequences = value.events.map((event: Document) => event.event_sequence);
  return sequences.every((sequence: number, index: number) => index === 0 || sequence === sequences[index - 1] + 1);
}

function isAssignment(value: unknown): value is Document {
  return isRecord(value) &&
    value.schema_version === "agent_copilot_runtime_assignment.v1" &&
    /^acra_[a-z2-7]{16}$/u.test(String(value.assignment_id)) &&
    Number.isInteger(value.assignment_version) && value.assignment_version >= 1 &&
    (value.state === "active" || value.state === "revoked") &&
    typeof value.candidate_id === "string" &&
    isProfileRef(value.agent_copilot_profile_ref) &&
    DIGEST.test(String(value.assignment_digest));
}

function isEvent(value: unknown): value is Document {
  return isRecord(value) &&
    value.schema_version === "agent_copilot_runtime_assignment_event.v1" &&
    /^acrae_[a-z2-7]{16}$/u.test(String(value.event_id)) &&
    Number.isInteger(value.event_sequence) &&
    (value.action === "activate" || value.action === "replace" || value.action === "revoke");
}

function isProfileRef(value: unknown): value is Document {
  return isRecord(value) &&
    /^acpf_[a-z2-7]{16}$/u.test(String(value.profile_id)) &&
    Number.isInteger(value.profile_version) && value.profile_version >= 1 &&
    DIGEST.test(String(value.profile_digest)) &&
    DIGEST.test(String(value.policy_digest));
}

function runtimeHeaders(
  config: AgentCopilotRuntimeConfig,
  applicationId: string,
  requestId: string,
  operation: "read" | "write",
): Record<string, string> {
  return {
    Accept: "application/json",
    "X-Request-Id": requestId,
    "X-RadishMind-Dev-Read-Identity": "radishmind-web-agent-copilot-runtime",
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Read-Scopes": operation === "write"
      ? "agent_copilot_runtime:read,agent_copilot_runtime:write"
      : "agent_copilot_runtime:read",
    "X-RadishMind-Dev-Read-Audit": `audit_${requestId}_agent_runtime`,
    ...(operation === "write" ? {
      "X-RadishMind-Active-Workspace": config.workspaceId,
      "X-RadishMind-Dev-Read-Membership-Workspace": config.workspaceId,
      "X-RadishMind-Dev-Read-Membership-Permissions": "agent_copilot_runtime:write",
    } : {}),
    "X-RadishMind-Dev-Agent-Copilot-Runtime-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Agent-Copilot-Runtime-Application": applicationId,
  };
}

function createRequestId(prefix: string): string {
  return `${prefix}-${Date.now()}-${(globalThis.crypto?.randomUUID?.() ?? Math.random().toString(16).slice(2)).replaceAll("-", "").slice(0, 12)}`;
}

function isRecord(value: unknown): value is Document {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function containsForbiddenResponse(value: unknown): boolean {
  if (Array.isArray(value)) return value.some(containsForbiddenResponse);
  if (!isRecord(value)) return false;
  return Object.entries(value).some(([key, nested]) =>
    FORBIDDEN_RESPONSE_KEYS.has(key.toLowerCase()) || containsForbiddenResponse(nested)
  );
}
