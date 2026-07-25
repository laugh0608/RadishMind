const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
type Document = Record<string, unknown>;
const FORBIDDEN_SESSION_RESPONSE_KEYS = new Set([
  "authorization", "api_key", "credential", "headers", "cookie", "dsn",
  "raw_request", "raw_response", "provider_raw_response", "variables",
  "input_text", "condition_values", "messages",
]);

export type PromptApplicationSessionConfig = {
  mode: "offline" | "dev_application_session_http";
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  subjectRef: string;
};

export type PromptApplicationSession = {
  sessionId: string;
  applicationId: string;
  state: "active" | "closed";
  recordVersion: number;
  assignmentId: string;
  assignmentVersion: number;
  templateId: string;
  templateVersion: number;
  turnCount: number;
  lastTurnId: string | null;
  createdAt: string;
  updatedAt: string;
};

export type PromptApplicationSessionTurn = {
  turnId: string;
  sequence: number;
  clientTurnKey: string;
  status: "running" | "succeeded" | "failed" | "canceled" | "outcome_unknown";
  runId: string;
  failureCode: string;
  failureSummary: string;
  startedAt: string;
  completedAt: string | null;
};

export type PromptApplicationSessionResult = {
  status: "offline" | "ready" | "succeeded" | "replayed" | "blocked" | "failed";
  session: PromptApplicationSession | null;
  turn: PromptApplicationSessionTurn | null;
  output: string;
  failureCode: string;
  failureSummary: string;
  idempotentReplay: boolean;
  summary: string;
};

export function readPromptApplicationSessionConfig(): PromptApplicationSessionConfig {
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

export function initialPromptApplicationSessionResult(
  config: PromptApplicationSessionConfig,
): PromptApplicationSessionResult {
  return {
    status: config.mode === "offline" ? "offline" : "ready",
    session: null,
    turn: null,
    output: "",
    failureCode: config.mode === "offline" ? "application_session_http_disabled" : "",
    failureSummary: "",
    idempotentReplay: false,
    summary: config.mode === "offline"
      ? "Prompt Session v2 owner is offline."
      : "Create a Prompt Application Session v2 from the current exact runtime authority.",
  };
}

export async function createPromptApplicationSession(
  config: PromptApplicationSessionConfig,
  applicationId: string,
): Promise<PromptApplicationSessionResult> {
  if (config.mode === "offline") return initialPromptApplicationSessionResult(config);
  return requestSession(config, applicationId, "/v1/user-workspace/application-sessions", {
    workspace_id: config.workspaceId,
    application_id: applicationId,
    execution_profile: "prompt_application_invocation_v1",
  }, "application_sessions:write", null);
}

export async function executePromptApplicationSessionTurn(
  config: PromptApplicationSessionConfig,
  session: PromptApplicationSession,
  variables: Record<string, string | number | boolean | string[]>,
  clientTurnKey: string,
): Promise<PromptApplicationSessionResult> {
  if (config.mode === "offline") return initialPromptApplicationSessionResult(config);
  if (session.state !== "active" || !/^[A-Za-z0-9][A-Za-z0-9._:-]{7,159}$/u.test(clientTurnKey)) {
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
      client_turn_key: clientTurnKey,
      input_text: "",
      condition_values: {},
      model: "",
      variables,
    },
    "application_sessions:execute",
    session.sessionId,
  );
}

async function requestSession(
  config: PromptApplicationSessionConfig,
  applicationId: string,
  path: string,
  body: unknown,
  scope: string,
  expectedSessionId: string | null,
): Promise<PromptApplicationSessionResult> {
  const requestId = createRequestId("prompt-session");
  try {
    const response = await fetch(`${config.baseUrl}${path}`, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        "X-Request-Id": requestId,
        "X-RadishMind-Dev-Read-Identity": `prompt-application-session-web:${config.subjectRef}`,
        "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
        "X-RadishMind-Dev-Read-Subject": config.subjectRef,
        "X-RadishMind-Dev-Workflow-Workspace": config.workspaceId,
        "X-RadishMind-Dev-Workflow-Application": applicationId,
        "X-RadishMind-Dev-Read-Scopes": scope,
      },
      body: JSON.stringify(body),
    });
    const value: unknown = await response.json();
    if (!response.ok || !isSessionEnvelope(value, config, applicationId, expectedSessionId)) {
      return failed("application_session_response_invalid");
    }
    const failureCode = nullableString(value.failure_code);
    const session = value.session === null ? null : mapSession(value.session as Document);
    const turn = expectedSessionId && value.turn !== null ? mapTurn(value.turn as Document) : null;
    const replay = value.idempotent_replay as boolean;
    const output = replay ? "" : nullableString(value.prompt_output);
    return {
      status: failureCode ? "blocked" : replay ? "replayed" : turn ? "succeeded" : "ready",
      session,
      turn,
      output,
      failureCode,
      failureSummary: nullableString(value.failure_summary),
      idempotentReplay: replay,
      summary: failureCode
        ? `Prompt Session v2 失败关闭：${failureCode}。`
        : turn
          ? `Session turn #${turn.sequence} 已记录。`
          : `Prompt Session v2 ${session?.sessionId ?? ""} 已创建。`,
    };
  } catch {
    return failed("application_session_store_unavailable");
  }
}

function isSessionEnvelope(
  value: unknown,
  config: PromptApplicationSessionConfig,
  applicationId: string,
  expectedSessionId: string | null,
): value is Document {
  if (!isRecord(value) || value.tenant_ref !== config.tenantRef || value.workspace_id !== config.workspaceId ||
    value.application_id !== applicationId || (value.failure_code !== null && typeof value.failure_code !== "string") ||
    typeof value.idempotent_replay !== "boolean" || containsForbiddenSessionResponse(value) ||
    (value.session !== null && !isSession(value.session, config, applicationId))) return false;
  if (!expectedSessionId) {
    return Object.hasOwn(value, "current_record_version") && Object.hasOwn(value, "current_state") &&
      !Object.hasOwn(value, "turn") && !Object.hasOwn(value, "prompt_output");
  }
  return value.session_id === expectedSessionId && value.session !== null &&
    (value.turn === null || isTurn(value.turn, config, applicationId, expectedSessionId)) &&
    (value.prompt_output === undefined || typeof value.prompt_output === "string") &&
    value.advisory_output === undefined && value.answer === undefined &&
    (!value.idempotent_replay || value.prompt_output === undefined || value.prompt_output === "");
}

function containsForbiddenSessionResponse(value: unknown): boolean {
  if (Array.isArray(value)) return value.some(containsForbiddenSessionResponse);
  if (!isRecord(value)) return false;
  return Object.entries(value).some(([key, nested]) =>
    FORBIDDEN_SESSION_RESPONSE_KEYS.has(key) || containsForbiddenSessionResponse(nested)
  );
}

function isSession(value: unknown, config: PromptApplicationSessionConfig, applicationId: string): value is Document {
  if (!isRecord(value) || value.schema_version !== "application_session.v2" ||
    value.tenant_ref !== config.tenantRef || value.workspace_id !== config.workspaceId || value.application_id !== applicationId ||
    !/^appsess_[a-z2-7]{16}$/u.test(String(value.session_id)) || (value.state !== "active" && value.state !== "closed") ||
    !integer(value.record_version, 1) || !isRecord(value.profile_binding) ||
    value.profile_binding.execution_profile !== "prompt_application_invocation_v1" ||
    !isPromptAuthority(value.authority, applicationId) || value.content_retention !== "metadata_only" ||
    !integer(value.turn_count, 0)) return false;
  return value.last_turn_id === null || /^appturn_[a-z2-7]{16}$/u.test(String(value.last_turn_id));
}

function isTurn(value: unknown, config: PromptApplicationSessionConfig, applicationId: string, sessionId: string): value is Document {
  return isRecord(value) && value.schema_version === "application_session_turn.v2" &&
    value.session_id === sessionId && value.tenant_ref === config.tenantRef &&
    value.workspace_id === config.workspaceId && value.application_id === applicationId &&
    /^appturn_[a-z2-7]{16}$/u.test(String(value.turn_id)) && integer(value.sequence, 1) &&
    value.execution_profile === "prompt_application_invocation_v1" &&
    isPromptAuthority(value.authority, applicationId) && typeof value.status === "string" &&
    isRecord(value.run_ref) && value.run_ref.schema_version === "workflow_run_record.v6" &&
    typeof value.run_ref.run_id === "string";
}

function isPromptAuthority(value: unknown, applicationId: string): value is Document {
  return isRecord(value) && value.schema_version === "application_runtime_authority.v2" &&
    value.application_id === applicationId && value.execution_profile === "prompt_application_invocation_v1" &&
    isRecord(value.prompt_application) && /^ptra_[a-z2-7]{16}$/u.test(String(value.prompt_application.assignment_id)) &&
    isRecord(value.prompt_application.prompt_template_ref) &&
    /^ptpl_[a-z2-7]{16}$/u.test(String(value.prompt_application.prompt_template_ref.template_id));
}

function mapSession(value: Document): PromptApplicationSession {
  const authority = value.authority as Document;
  const prompt = authority.prompt_application as Document;
  const template = prompt.prompt_template_ref as Document;
  return {
    sessionId: value.session_id as string,
    applicationId: value.application_id as string,
    state: value.state as PromptApplicationSession["state"],
    recordVersion: value.record_version as number,
    assignmentId: prompt.assignment_id as string,
    assignmentVersion: prompt.assignment_version as number,
    templateId: template.template_id as string,
    templateVersion: template.template_version as number,
    turnCount: value.turn_count as number,
    lastTurnId: value.last_turn_id as string | null,
    createdAt: value.created_at as string,
    updatedAt: value.updated_at as string,
  };
}

function mapTurn(value: Document): PromptApplicationSessionTurn {
  const runRef = value.run_ref as Document;
  return {
    turnId: value.turn_id as string,
    sequence: value.sequence as number,
    clientTurnKey: value.client_turn_key as string,
    status: value.status as PromptApplicationSessionTurn["status"],
    runId: runRef.run_id as string,
    failureCode: value.failure_code as string,
    failureSummary: value.failure_summary as string,
    startedAt: value.started_at as string,
    completedAt: value.completed_at as string | null,
  };
}

function failed(failureCode: string): PromptApplicationSessionResult {
  return {
    status: "failed",
    session: null,
    turn: null,
    output: "",
    failureCode,
    failureSummary: "",
    idempotentReplay: false,
    summary: `Prompt Session v2 不可用：${failureCode}。`,
  };
}

function nullableString(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function integer(value: unknown, minimum: number): boolean {
  return Number.isInteger(value) && (value as number) >= minimum;
}

function isRecord(value: unknown): value is Document {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function createRequestId(prefix: string): string {
  return `${prefix}-${Date.now()}-${(globalThis.crypto?.randomUUID?.() ?? Math.random().toString(16).slice(2)).replaceAll("-", "").slice(0, 12)}`;
}
