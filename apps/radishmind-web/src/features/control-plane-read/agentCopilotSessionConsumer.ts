import {
  parseApplicationResultArtifactSummary,
  type ApplicationResultArtifactSummary,
} from "./applicationResultArtifactConsumer.ts";

const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const EXECUTION_PROFILE = "agent_copilot_suggestion_v1";
const FORBIDDEN_KEYS = new Set([
  "authorization", "api_key", "credential", "headers", "cookie", "dsn",
  "raw_request", "raw_response", "provider_raw_response",
]);

type Document = Record<string, any>;

export type AgentCopilotSessionConfig = {
  mode: "offline" | "dev_application_session_http";
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  subjectRef: string;
};
export type AgentCopilotSession = {
  sessionId: string;
  applicationId: string;
  state: "active" | "closed";
  recordVersion: number;
  assignmentId: string;
  assignmentVersion: number;
  profileId: string;
  profileVersion: number;
  project: string;
  turnCount: number;
};
export type AgentCopilotSessionTurn = {
  turnId: string;
  sequence: number;
  status: "running" | "succeeded" | "failed" | "canceled" | "outcome_unknown";
  runId: string;
  failureCode: string;
  failureSummary: string;
};
export type AgentCopilotResponse = {
  status: "ok" | "partial";
  project: string;
  task: string;
  summary: string;
  answers: Array<{ text: string }>;
  issues: Array<{ message: string; severity: string }>;
  proposedActions: Array<{ kind: string; title: string; riskLevel: string; requiresConfirmation: boolean }>;
  citations: Array<{ id: string; kind: string; label: string }>;
  riskLevel: string;
  requiresConfirmation: boolean;
};
export type AgentCopilotSessionResult = {
  status: "offline" | "ready" | "succeeded" | "replayed" | "blocked" | "failed";
  session: AgentCopilotSession | null;
  turn: AgentCopilotSessionTurn | null;
  response: AgentCopilotResponse | null;
  resultArtifact: ApplicationResultArtifactSummary | null;
  resultArtifactFailureCode: string;
  failureCode: string;
  failureSummary: string;
  idempotentReplay: boolean;
  summary: string;
};
export type AgentCopilotTurnInput = {
  task: string;
  locale: string;
  conversationId: string;
  artifacts: Array<{
    kind: string;
    role: string;
    name: string;
    mime_type: string;
    uri?: string;
    content?: unknown;
    metadata?: Record<string, unknown>;
  }>;
  context: Record<string, unknown>;
  saveResult?: boolean;
  clientTurnKey: string;
};

export function readAgentCopilotSessionConfig(): AgentCopilotSessionConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  return {
    mode: env.VITE_RADISHMIND_APPLICATION_SESSION_SOURCE?.trim() === "dev-application-session-http"
      ? "dev_application_session_http"
      : "offline",
    baseUrl: (
      env.VITE_RADISHMIND_APPLICATION_SESSION_BASE_URL ??
      env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ??
      DEFAULT_BASE_URL
    ).trim().replace(/\/$/u, ""),
    tenantRef: env.VITE_RADISHMIND_DEV_READ_TENANT_REF?.trim() || "tenant_demo",
    workspaceId: env.VITE_RADISHMIND_APPLICATION_SESSION_WORKSPACE_ID?.trim() || "workspace_demo",
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || "subject_demo_user",
  };
}

export function initialAgentCopilotSessionResult(
  config: AgentCopilotSessionConfig,
): AgentCopilotSessionResult {
  return {
    status: config.mode === "offline" ? "offline" : "ready",
    session: null,
    turn: null,
    response: null,
    resultArtifact: null,
    resultArtifactFailureCode: "",
    failureCode: config.mode === "offline" ? "application_session_http_disabled" : "",
    failureSummary: "",
    idempotentReplay: false,
    summary: config.mode === "offline"
      ? "Agent Copilot Session v3 owner is offline."
      : "Create a Session v3 from the exact active Agent Copilot assignment.",
  };
}

export async function createAgentCopilotSession(
  config: AgentCopilotSessionConfig,
  applicationId: string,
  signal?: AbortSignal,
): Promise<AgentCopilotSessionResult> {
  if (config.mode === "offline") return initialAgentCopilotSessionResult(config);
  return requestSession(
    config,
    applicationId,
    "/v1/user-workspace/application-sessions",
    {
      workspace_id: config.workspaceId,
      application_id: applicationId,
      execution_profile: EXECUTION_PROFILE,
    },
    "application_sessions:write",
    null,
    signal,
  );
}

export async function executeAgentCopilotSessionTurn(
  config: AgentCopilotSessionConfig,
  session: AgentCopilotSession,
  input: AgentCopilotTurnInput,
  signal?: AbortSignal,
): Promise<AgentCopilotSessionResult> {
  if (config.mode === "offline") return initialAgentCopilotSessionResult(config);
  if (session.state !== "active" || !/^[A-Za-z0-9][A-Za-z0-9._:-]{7,159}$/u.test(input.clientTurnKey)) {
    return failed("application_session_payload_invalid");
  }
  return requestSession(
    config,
    session.applicationId,
    `/v1/user-workspace/application-sessions/${encodeURIComponent(session.sessionId)}/turns`,
    {
      workspace_id: config.workspaceId,
      application_id: session.applicationId,
      expected_session_version: session.recordVersion,
      client_turn_key: input.clientTurnKey,
      input_text: "",
      condition_values: {},
      model: "",
      temperature: null,
      task: input.task,
      locale: input.locale,
      conversation_id: input.conversationId || undefined,
      artifacts: input.artifacts,
      context: input.context,
      ...(input.saveResult === true ? { save_result: true } : {}),
    },
    "application_sessions:execute",
    session.sessionId,
    signal,
  );
}

async function requestSession(
  config: AgentCopilotSessionConfig,
  applicationId: string,
  path: string,
  body: unknown,
  scope: string,
  expectedSessionId: string | null,
  signal?: AbortSignal,
): Promise<AgentCopilotSessionResult> {
  const requestId = createRequestId("agent-session");
  try {
    const response = await fetch(`${config.baseUrl}${path}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...sessionHeaders(config, applicationId, requestId, scope),
      },
      body: JSON.stringify(body),
      signal,
    });
    const value: unknown = await response.json();
    if (!response.ok || !isEnvelope(value, config, applicationId, expectedSessionId)) {
      return failed("application_session_response_invalid");
    }
    const failureCode = nullableString(value.failure_code);
    const session = value.session === null ? null : mapSession(value.session);
    const turn = expectedSessionId && value.turn !== null ? mapTurn(value.turn) : null;
    const replay = value.idempotent_replay === true;
    const agentResponse = !replay && isAgentResponse(value.agent_response)
      ? mapAgentResponse(value.agent_response)
      : null;
    const resultArtifact = value.result_artifact === undefined || value.result_artifact === null
      ? null
      : parseApplicationResultArtifactSummary(value.result_artifact, config, applicationId, expectedSessionId ?? "");
    const resultArtifactFailureCode = nullableString(value.result_artifact_failure_code);
    if ((value.result_artifact !== undefined && value.result_artifact !== null && !resultArtifact) ||
      (resultArtifact && resultArtifactFailureCode) ||
      (resultArtifactFailureCode && (!turn || turn.status !== "succeeded" || Boolean(failureCode))) ||
      (resultArtifact && (!turn || resultArtifact.turnId !== turn.turnId || resultArtifact.runRef.runId !== turn.runId))) {
      return failed("application_session_response_invalid");
    }
    return {
      status: failureCode ? "blocked" : replay ? "replayed" : turn ? "succeeded" : "ready",
      session,
      turn,
      response: agentResponse,
      resultArtifact,
      resultArtifactFailureCode,
      failureCode,
      failureSummary: nullableString(value.failure_summary),
      idempotentReplay: replay,
      summary: failureCode
        ? `Agent Copilot Session v3 失败关闭：${failureCode}。`
        : turn
          ? `Agent Copilot turn #${turn.sequence} 已记录。`
          : `Agent Copilot Session v3 ${session?.sessionId ?? ""} 已创建。`,
    };
  } catch (error) {
    return failed(error instanceof DOMException && error.name === "AbortError"
      ? "application_session_request_canceled"
      : "application_session_store_unavailable");
  }
}

function isEnvelope(
  value: unknown,
  config: AgentCopilotSessionConfig,
  applicationId: string,
  expectedSessionId: string | null,
): value is Document {
  if (!isRecord(value) || containsForbidden(value) || value.tenant_ref !== config.tenantRef ||
      value.workspace_id !== config.workspaceId || value.application_id !== applicationId ||
      (value.failure_code !== null && typeof value.failure_code !== "string") ||
      typeof value.idempotent_replay !== "boolean" ||
      (value.session !== null && !isSession(value.session, config, applicationId))) return false;
  if (!expectedSessionId) {
    return Object.hasOwn(value, "current_record_version") && Object.hasOwn(value, "current_state") &&
      !Object.hasOwn(value, "turn") && !Object.hasOwn(value, "agent_response") &&
      !Object.hasOwn(value, "result_artifact") && !Object.hasOwn(value, "result_artifact_failure_code");
  }
  return value.session_id === expectedSessionId && value.session !== null &&
    (value.turn === null || isTurn(value.turn, config, applicationId, expectedSessionId)) &&
    (value.agent_response === undefined || value.agent_response === null || isAgentResponse(value.agent_response)) &&
    (value.result_artifact === undefined || value.result_artifact === null || isRecord(value.result_artifact)) &&
    (value.result_artifact_failure_code === undefined || typeof value.result_artifact_failure_code === "string" && /^[a-z][a-z0-9_]{2,127}$/u.test(value.result_artifact_failure_code)) &&
    value.prompt_output === undefined && value.answer === undefined && value.advisory_output === undefined &&
    (!value.idempotent_replay || value.agent_response === undefined || value.agent_response === null);
}

function isSession(value: unknown, config: AgentCopilotSessionConfig, applicationId: string): value is Document {
  return isRecord(value) && value.schema_version === "application_session.v3" &&
    value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId &&
    value.application_id === applicationId && /^appsess_[a-z2-7]{16}$/u.test(String(value.session_id)) &&
    (value.state === "active" || value.state === "closed") &&
    Number.isInteger(value.record_version) && value.record_version >= 1 &&
    isRecord(value.profile_binding) && value.profile_binding.execution_profile === EXECUTION_PROFILE &&
    isAuthority(value.authority, applicationId) && value.content_retention === "metadata_only";
}

function isTurn(
  value: unknown,
  config: AgentCopilotSessionConfig,
  applicationId: string,
  sessionId: string,
): value is Document {
  return isRecord(value) && value.schema_version === "application_session_turn.v3" &&
    value.session_id === sessionId && value.tenant_ref === config.tenantRef &&
    value.workspace_id === config.workspaceId && value.application_id === applicationId &&
    /^appturn_[a-z2-7]{16}$/u.test(String(value.turn_id)) &&
    value.execution_profile === EXECUTION_PROFILE && isAuthority(value.authority, applicationId) &&
    (value.run_ref === null || isRecord(value.run_ref) &&
      value.run_ref.schema_version === "workflow_run_record.v7" && typeof value.run_ref.run_id === "string");
}

function isAuthority(value: unknown, applicationId: string): value is Document {
  return isRecord(value) && value.schema_version === "application_runtime_authority.v3" &&
    value.application_id === applicationId && value.execution_profile === EXECUTION_PROFILE &&
    isRecord(value.agent_copilot) &&
    /^acra_[a-z2-7]{16}$/u.test(String(value.agent_copilot.assignment_id)) &&
    isRecord(value.agent_copilot.agent_copilot_profile_ref) &&
    /^acpf_[a-z2-7]{16}$/u.test(String(value.agent_copilot.agent_copilot_profile_ref.profile_id));
}

function isAgentResponse(value: unknown): value is Document {
  if (!isRecord(value) || value.schema_version !== 1 ||
    (value.status !== "ok" && value.status !== "partial") ||
    typeof value.project !== "string" || typeof value.task !== "string" ||
    typeof value.summary !== "string" || !Array.isArray(value.answers) ||
    !Array.isArray(value.issues) || !Array.isArray(value.proposed_actions) ||
    !Array.isArray(value.citations) || typeof value.requires_confirmation !== "boolean") {
    return false;
  }
  if (!["low", "medium", "high"].includes(String(value.risk_level)) ||
    !value.answers.every((answer: unknown) => isRecord(answer) && typeof answer.text === "string") ||
    !value.issues.every((issue: unknown) =>
      isRecord(issue) && typeof issue.message === "string" && typeof issue.severity === "string"
    ) ||
    !value.citations.every((citation: unknown) =>
      isRecord(citation) && typeof citation.id === "string" &&
      typeof citation.kind === "string" && typeof citation.label === "string"
    ) ||
    !value.proposed_actions.every((action: unknown) =>
      isRecord(action) && typeof action.kind === "string" && typeof action.title === "string" &&
      ["low", "medium", "high"].includes(String(action.risk_level)) &&
      action.requires_confirmation === true
    )) return false;
  return value.proposed_actions.length === 0
    ? value.requires_confirmation === false
    : value.requires_confirmation === true;
}

function mapSession(value: Document): AgentCopilotSession {
  const authority = value.authority.agent_copilot;
  const profile = authority.agent_copilot_profile_ref;
  return {
    sessionId: value.session_id,
    applicationId: value.application_id,
    state: value.state,
    recordVersion: value.record_version,
    assignmentId: authority.assignment_id,
    assignmentVersion: authority.assignment_version,
    profileId: profile.profile_id,
    profileVersion: profile.profile_version,
    project: authority.project,
    turnCount: value.turn_count,
  };
}

function mapTurn(value: Document): AgentCopilotSessionTurn {
  return {
    turnId: value.turn_id,
    sequence: value.sequence,
    status: value.status,
    runId: value.run_ref?.run_id ?? "",
    failureCode: value.failure_code,
    failureSummary: value.failure_summary,
  };
}

function mapAgentResponse(value: Document): AgentCopilotResponse {
  return {
    status: value.status,
    project: value.project,
    task: value.task,
    summary: value.summary,
    answers: value.answers.map((answer: Document) => ({ text: String(answer.text) })),
    issues: value.issues.map((issue: Document) => ({ message: String(issue.message), severity: String(issue.severity) })),
    proposedActions: value.proposed_actions.map((action: Document) => ({
      kind: String(action.kind),
      title: String(action.title),
      riskLevel: String(action.risk_level),
      requiresConfirmation: action.requires_confirmation === true,
    })),
    citations: value.citations.map((citation: Document) => ({
      id: String(citation.id),
      kind: String(citation.kind),
      label: String(citation.label),
    })),
    riskLevel: String(value.risk_level),
    requiresConfirmation: value.requires_confirmation === true,
  };
}

function sessionHeaders(
  config: AgentCopilotSessionConfig,
  applicationId: string,
  requestId: string,
  scope: string,
): Record<string, string> {
  const mutationPermission = scope === "application_sessions:write" || scope === "application_sessions:execute" ? scope : "";
  return {
    Accept: "application/json",
    "X-Request-Id": requestId,
    "X-RadishMind-Dev-Read-Identity": `agent-copilot-session-web:${config.subjectRef}`,
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Workflow-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Workflow-Application": applicationId,
    "X-RadishMind-Dev-Read-Scopes": scope,
    ...(mutationPermission ? {
      "X-RadishMind-Active-Workspace": config.workspaceId,
      "X-RadishMind-Dev-Read-Membership-Workspace": config.workspaceId,
      "X-RadishMind-Dev-Read-Membership-Permissions": mutationPermission,
    } : {}),
  };
}

function failed(failureCode: string): AgentCopilotSessionResult {
  return {
    status: "failed",
    session: null,
    turn: null,
    response: null,
    resultArtifact: null,
    resultArtifactFailureCode: "",
    failureCode,
    failureSummary: "",
    idempotentReplay: false,
    summary: `Agent Copilot Session v3 不可用：${failureCode}。`,
  };
}

function containsForbidden(value: unknown): boolean {
  if (Array.isArray(value)) return value.some(containsForbidden);
  if (!isRecord(value)) return false;
  return Object.entries(value).some(([key, nested]) => FORBIDDEN_KEYS.has(key) || containsForbidden(nested));
}

function nullableString(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function createRequestId(prefix: string): string {
  return `${prefix}-${Date.now()}-${(globalThis.crypto?.randomUUID?.() ?? Math.random().toString(16).slice(2)).replaceAll("-", "").slice(0, 12)}`;
}

function isRecord(value: unknown): value is Document {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}
