const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const TOKEN_PATTERN = /^rmd_dev_key_[a-z2-7]{16}\.[A-Za-z0-9_-]{43}$/u;
const CLIENT_KEY_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._:-]{7,159}$/u;
const VARIABLE_NAME_PATTERN = /^[A-Za-z][A-Za-z0-9_]{0,63}$/u;
const DIGEST_PATTERN = /^sha256:[a-f0-9]{64}$/u;
const SECRET_MATERIAL = /authorization:|bearer\s|api[_-]?key\s*[:=]|cookie:|password\s*=|secret\s*=|token\s*=|sk-[a-z0-9]|(?:postgres(?:ql)?|mysql|mongodb):\/\//iu;
const ENVELOPE_KEYS = [
  "request_id", "tenant_ref", "workspace_id", "application_id", "run", "output",
  "failure_code", "failure_summary", "idempotent_replay", "audit_ref",
] as const;
const RUN_KEYS = [
  "schema_version", "record_version", "run_id", "tenant_ref", "workspace_id", "application_id",
  "execution_kind", "execution_source_kind", "execution_source_id", "execution_source_version",
  "execution_profile", "authority", "input_digest", "input_bytes", "variable_names",
  "variable_names_digest", "requested_protocol", "selected_protocol", "requested_model",
  "selected_provider", "selected_profile", "selected_model", "upstream_model", "selection_source",
  "status", "failure_code", "failure_summary", "started_at", "completed_at", "output",
  "usage", "side_effects", "diagnostic", "request_id", "audit_ref", "actor_ref",
] as const;

type Document = Record<string, unknown>;

export type PromptApplicationInvocationConfig = {
  baseUrl: string;
};

export type PromptApplicationInvocationRun = {
  runId: string;
  recordVersion: number;
  status: string;
  assignmentId: string;
  assignmentVersion: number;
  candidateId: string;
  draftId: string;
  templateId: string;
  templateVersion: number;
  selectedProtocol: string;
  selectedModel: string;
  providerCalls: number;
  failureCode: string;
  failureSummary: string;
  startedAt: string;
  completedAt: string;
  requestId: string;
  auditRef: string;
};

export type PromptApplicationInvocationResult = {
  status: "idle" | "running" | "succeeded" | "replayed" | "blocked" | "failed";
  run: PromptApplicationInvocationRun | null;
  output: string;
  failureCode: string;
  failureSummary: string;
  idempotentReplay: boolean;
  requestId: string;
  auditRef: string;
  summary: string;
};

export type PromptApplicationInvocationInputValidation = {
  isValid: boolean;
  failureCode: string;
  variables: Record<string, string | number | boolean | string[]>;
};

export function readPromptApplicationInvocationConfig(): PromptApplicationInvocationConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  return {
    baseUrl: (
      env.VITE_RADISHMIND_PROMPT_APPLICATION_BASE_URL ??
      env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ??
      DEFAULT_BASE_URL
    ).trim().replace(/\/$/u, ""),
  };
}

export function initialPromptApplicationInvocationResult(): PromptApplicationInvocationResult {
  return {
    status: "idle",
    run: null,
    output: "",
    failureCode: "",
    failureSummary: "",
    idempotentReplay: false,
    requestId: "",
    auditRef: "",
    summary: "接收一次性 API key 后，可执行一次受控 Prompt Application 调用。",
  };
}

export function parsePromptApplicationVariables(raw: string): PromptApplicationInvocationInputValidation {
  try {
    if (new TextEncoder().encode(raw).byteLength > 64 * 1024 || SECRET_MATERIAL.test(raw)) {
      return invalidInput("prompt_invocation_input_invalid");
    }
    const value: unknown = JSON.parse(raw);
    if (!isRecord(value) || Object.keys(value).length > 64) return invalidInput("prompt_invocation_input_invalid");
    const variables: Record<string, string | number | boolean | string[]> = {};
    for (const [name, item] of Object.entries(value)) {
      if (!VARIABLE_NAME_PATTERN.test(name) || !isVariableValue(item)) {
        return invalidInput("prompt_invocation_input_invalid");
      }
      variables[name] = Array.isArray(item) ? [...item] : item;
    }
    return { isValid: true, failureCode: "", variables };
  } catch {
    return invalidInput("prompt_invocation_input_invalid");
  }
}

export async function invokePromptApplication(
  config: PromptApplicationInvocationConfig,
  credential: string,
  variables: Record<string, string | number | boolean | string[]>,
  clientInvocationKey: string,
  signal?: AbortSignal,
): Promise<PromptApplicationInvocationResult> {
  if (!TOKEN_PATTERN.test(credential.trim()) ||
    !CLIENT_KEY_PATTERN.test(clientInvocationKey.trim()) ||
    Object.keys(variables).length > 64) {
    return failedResult("prompt_invocation_input_invalid", "调用输入或一次性 credential 不符合约束。");
  }
  const requestId = createRequestId("prompt-invocation");
  try {
    const response = await fetch(`${config.baseUrl}/v1/prompt-applications/invocations`, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        Authorization: `Bearer ${credential.trim()}`,
        "X-Request-Id": requestId,
      },
      body: JSON.stringify({
        variables,
        client_invocation_key: clientInvocationKey.trim(),
      }),
      signal,
    });
    const value: unknown = await response.json();
    if (!response.ok || !isInvocationEnvelope(value)) {
      return failedResult("prompt_invocation_response_invalid", "调用响应不符合 Prompt Application 契约。");
    }
    const failureCode = nullableString(value.failure_code);
    const run = value.run === null ? null : mapRun(value.run as Document);
    const replay = value.idempotent_replay as boolean;
    return {
      status: failureCode
        ? "blocked"
        : replay
          ? "replayed"
          : run?.status === "succeeded"
            ? "succeeded"
            : "failed",
      run,
      output: replay ? "" : value.output as string,
      failureCode,
      failureSummary: value.failure_summary as string,
      idempotentReplay: replay,
      requestId: value.request_id as string,
      auditRef: value.audit_ref as string,
      summary: failureCode
        ? `Prompt invocation 失败关闭：${failureCode}。`
        : replay
          ? "幂等重试只返回终态 metadata，不恢复 output。"
          : `Prompt invocation 已记录 ${run?.runId ?? "terminal metadata"}。`,
    };
  } catch (error) {
    if (error instanceof DOMException && error.name === "AbortError") {
      return failedResult("prompt_invocation_canceled", "调用已取消，客户端不会自动重试。");
    }
    return failedResult("prompt_invocation_transport_failed", "调用传输失败，客户端不会自动重试或改键绕过。");
  }
}

function isInvocationEnvelope(value: unknown): value is Document {
  if (!isRecord(value) || !hasExactKeys(value, ENVELOPE_KEYS) ||
    !isReference(value.request_id) || !isReference(value.tenant_ref) ||
    !isReference(value.workspace_id) || !/^app_[a-z2-7]{16}$/u.test(String(value.application_id)) ||
    !isReference(value.audit_ref) || (value.failure_code !== null && !isReference(value.failure_code)) ||
    typeof value.failure_summary !== "string" || typeof value.idempotent_replay !== "boolean" ||
    typeof value.output !== "string" || new TextEncoder().encode(value.output).byteLength > 65536 ||
    (value.run !== null && !isPromptRun(value.run, value))) return false;
  if (value.idempotent_replay && value.output !== "") return false;
  if (value.failure_code !== null && value.output !== "") return false;
  return !containsForbiddenResponse(value, new Set(["output"]));
}

function isPromptRun(value: unknown, envelope: Document): value is Document {
  if (!isRecord(value) || !hasExactKeys(value, RUN_KEYS) ||
    value.schema_version !== "workflow_run_record.v6" || !integer(value.record_version, 1) ||
    value.tenant_ref !== envelope.tenant_ref || value.workspace_id !== envelope.workspace_id ||
    value.application_id !== envelope.application_id || value.execution_kind !== "prompt_application_invocation" ||
    value.execution_source_kind !== "prompt_application_template" ||
    value.execution_profile !== "prompt_application_invocation_v1" ||
    !isReference(value.run_id) || !isReference(value.execution_source_id) ||
    !integer(value.execution_source_version, 1) || !isDigest(value.input_digest) ||
    !integer(value.input_bytes, 0) || !Array.isArray(value.variable_names) ||
    !value.variable_names.every((name) => typeof name === "string" && VARIABLE_NAME_PATTERN.test(name)) ||
    !isDigest(value.variable_names_digest) || value.output !== "" ||
    !isReference(value.request_id) || !isReference(value.audit_ref) || !isReference(value.actor_ref) ||
    !isAuthority(value.authority, envelope.application_id as string) ||
    !isUsage(value.usage) || !isSideEffects(value.side_effects) || !isDiagnostic(value.diagnostic)) return false;
  return typeof value.status === "string" && typeof value.failure_code === "string" &&
    typeof value.failure_summary === "string" && typeof value.started_at === "string" &&
    typeof value.completed_at === "string" && typeof value.selected_protocol === "string" &&
    typeof value.selected_model === "string";
}

function isAuthority(value: unknown, applicationId: string): value is Document {
  if (!isRecord(value) || value.schema_version !== "application_runtime_authority.v2" ||
    value.execution_profile !== "prompt_application_invocation_v1" ||
    value.application_id !== applicationId || !integer(value.application_record_version, 1) ||
    value.application_lifecycle !== "active" || !isDigest(value.authority_digest) ||
    !isRecord(value.prompt_application)) return false;
  const prompt = value.prompt_application;
  return /^ptra_[a-z2-7]{16}$/u.test(String(prompt.assignment_id)) &&
    integer(prompt.assignment_version, 1) && isDigest(prompt.assignment_digest) &&
    isReference(prompt.publish_candidate_id) && integer(prompt.publish_review_version, 1) &&
    isReference(prompt.draft_id) && integer(prompt.draft_version, 1) && isDigest(prompt.draft_digest) &&
    isRecord(prompt.prompt_template_ref) &&
    /^ptpl_[a-z2-7]{16}$/u.test(String(prompt.prompt_template_ref.template_id)) &&
    integer(prompt.prompt_template_ref.template_version, 1) &&
    isDigest(prompt.prompt_template_ref.template_digest);
}

function isUsage(value: unknown): boolean {
  return isRecord(value) && hasExactKeys(value, ["state", "input_tokens", "output_tokens", "total_tokens"]) &&
    (value.state === "unavailable" || value.state === "provider_reported") &&
    integer(value.input_tokens, 0) && integer(value.output_tokens, 0) && integer(value.total_tokens, 0);
}

function isSideEffects(value: unknown): boolean {
  if (!isRecord(value)) return false;
  const keys = value.retrieval_calls === undefined
    ? ["provider_calls", "tool_calls", "confirmation_calls", "business_writes", "replay_writes"]
    : ["retrieval_calls", "provider_calls", "tool_calls", "confirmation_calls", "business_writes", "replay_writes"];
  return hasExactKeys(value, keys) && (value.retrieval_calls === undefined || value.retrieval_calls === 0) &&
    integer(value.provider_calls, 0) && (value.provider_calls as number) <= 1 &&
    value.tool_calls === 0 && value.confirmation_calls === 0 &&
    value.business_writes === 0 && value.replay_writes === 0;
}

function isDiagnostic(value: unknown): boolean {
  return isRecord(value) && hasExactKeys(value, [
    "failure_boundary", "failure_stage", "terminal_write_state", "gateway_failure_category",
    "summary", "recommended_review_action", "observed_at",
  ]) && typeof value.summary === "string" && typeof value.observed_at === "string";
}

function mapRun(value: Document): PromptApplicationInvocationRun {
  const authority = value.authority as Document;
  const prompt = authority.prompt_application as Document;
  const template = prompt.prompt_template_ref as Document;
  const sideEffects = value.side_effects as Document;
  return {
    runId: value.run_id as string,
    recordVersion: value.record_version as number,
    status: value.status as string,
    assignmentId: prompt.assignment_id as string,
    assignmentVersion: prompt.assignment_version as number,
    candidateId: prompt.publish_candidate_id as string,
    draftId: prompt.draft_id as string,
    templateId: template.template_id as string,
    templateVersion: template.template_version as number,
    selectedProtocol: value.selected_protocol as string,
    selectedModel: value.selected_model as string,
    providerCalls: sideEffects.provider_calls as number,
    failureCode: value.failure_code as string,
    failureSummary: value.failure_summary as string,
    startedAt: value.started_at as string,
    completedAt: value.completed_at as string,
    requestId: value.request_id as string,
    auditRef: value.audit_ref as string,
  };
}

function invalidInput(failureCode: string): PromptApplicationInvocationInputValidation {
  return { isValid: false, failureCode, variables: {} };
}

function failedResult(failureCode: string, summary: string): PromptApplicationInvocationResult {
  return {
    ...initialPromptApplicationInvocationResult(),
    status: "failed",
    failureCode,
    failureSummary: summary,
    summary,
  };
}

function isVariableValue(value: unknown): value is string | number | boolean | string[] {
  if (typeof value === "string") return new TextEncoder().encode(value).byteLength <= 16 * 1024 && !SECRET_MATERIAL.test(value);
  if (typeof value === "number") return Number.isFinite(value);
  if (typeof value === "boolean") return true;
  return Array.isArray(value) && value.length <= 128 &&
    value.every((item) => typeof item === "string" && new TextEncoder().encode(item).byteLength <= 16 * 1024 && !SECRET_MATERIAL.test(item));
}

function containsForbiddenResponse(value: unknown, allowedKeys: Set<string>): boolean {
  if (Array.isArray(value)) return value.some((item) => containsForbiddenResponse(item, allowedKeys));
  if (!isRecord(value)) return false;
  const forbidden = /authorization|api[_-]?key|credential|headers|cookie|dsn|raw_request|raw_response|provider_raw_response|messages|variables/iu;
  return Object.entries(value).some(([key, nested]) =>
    (!allowedKeys.has(key) && forbidden.test(key)) || containsForbiddenResponse(nested, allowedKeys)
  );
}

function hasExactKeys(value: Document, expected: readonly string[]): boolean {
  const actual = Object.keys(value).sort();
  const target = [...expected].sort();
  return actual.length === target.length && actual.every((key, index) => key === target[index]);
}

function isDigest(value: unknown): value is string {
  return typeof value === "string" && DIGEST_PATTERN.test(value);
}

function isReference(value: unknown): value is string {
  return typeof value === "string" && /^[A-Za-z0-9][A-Za-z0-9_.:/-]{0,159}$/u.test(value);
}

function integer(value: unknown, minimum: number): boolean {
  return Number.isInteger(value) && (value as number) >= minimum;
}

function nullableString(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function isRecord(value: unknown): value is Document {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function createRequestId(prefix: string): string {
  return `${prefix}-${Date.now()}-${(globalThis.crypto?.randomUUID?.() ?? Math.random().toString(16).slice(2)).replaceAll("-", "").slice(0, 12)}`;
}
