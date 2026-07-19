const DEV_SOURCE = "dev-application-session-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const SESSION_SCHEMA_VERSION = "application_session.v1";
const TURN_SCHEMA_VERSION = "application_session_turn.v1";
const AUTHORITY_SCHEMA_VERSION = "application_runtime_authority.v1";
const ANSWER_SCHEMA_VERSION = "workflow_rag_application_answer.v1";
const APPLICATION_ID_PATTERN = /^app_[a-z0-9]{16}$/u;
const SESSION_ID_PATTERN = /^appsess_[a-z2-7]{16}$/u;
const TURN_ID_PATTERN = /^appturn_[a-z2-7]{16}$/u;
const RUN_ID_PATTERN = /^run_[a-z0-9]{16,64}$/u;
const RAG_ASSIGNMENT_ID_PATTERN = /^wragra_[a-z2-7]{16}$/u;
const RAG_BINDING_ID_PATTERN = /^wragb_[a-z2-7]{16}$/u;
const RAG_SNAPSHOT_ID_PATTERN = /^rags_[a-z2-7]{16}$/u;
const REF_PATTERN = /^[A-Za-z0-9][A-Za-z0-9_.:/-]{0,159}$/u;
const DIGEST_PATTERN = /^sha256:[a-f0-9]{64}$/u;
const FRAGMENT_REF_PATTERN = /^[a-z][a-z0-9_]{2,63}$/u;
const FORBIDDEN_RESPONSE_FIELDS = new Set([
  "authorization", "credential", "credentials", "token", "secret", "headers", "header", "cookie", "dsn",
  "raw_request", "raw_response", "prompt", "messages", "provider_raw_envelope", "fragment_content", "input_text",
]);

export type ApplicationInteractionExecutionProfile =
  | "workflow_definition_executor_v1"
  | "application_rag_invocation_v1";

export type ApplicationInteractionSessionConfig = {
  mode: "offline" | "dev_application_session_http";
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  subjectRef: string;
};

export type ApplicationInteractionAuthority = {
  schemaVersion: typeof AUTHORITY_SCHEMA_VERSION;
  executionProfile: ApplicationInteractionExecutionProfile;
  applicationId: string;
  applicationRecordVersion: number;
  applicationLifecycle: "active";
  workflowDefinition: null | {
    definitionId: string;
    definitionVersion: number;
    definitionDigest: string;
    activationPointerVersion: number;
    candidateId: string;
    candidateReviewVersion: number;
  };
  applicationRAG: null | {
    assignmentId: string;
    assignmentVersion: number;
    assignmentDigest: string;
    publishCandidateId: string;
    publishReviewVersion: number;
    draftId: string;
    draftVersion: number;
    draftDigest: string;
    bindingId: string;
    bindingVersion: 1;
    bindingDigest: string;
    snapshotId: string;
    snapshotVersion: number;
    snapshotDigest: string;
    retrievalProfileId: string;
    retrievalProfileVersion: number;
    retrievalProfileDigest: string;
  };
  authorityDigest: string;
};

export type ApplicationInteractionSession = {
  schemaVersion: typeof SESSION_SCHEMA_VERSION;
  sessionId: string;
  tenantRef: string;
  workspaceId: string;
  applicationId: string;
  ownerSubjectRef: string;
  state: "active" | "closed";
  recordVersion: number;
  executionProfile: ApplicationInteractionExecutionProfile;
  definitionId: string;
  authority: ApplicationInteractionAuthority;
  contentRetention: "metadata_only";
  turnCount: number;
  lastTurnId: string;
  createdAt: string;
  updatedAt: string;
  closedAt: string;
  requestId: string;
  auditRef: string;
};

export type ApplicationInteractionRunRef = {
  runId: string;
  schemaVersion: "workflow_run_record.v4" | "workflow_run_record.v5";
};

export type ApplicationInteractionTurn = {
  schemaVersion: typeof TURN_SCHEMA_VERSION;
  turnId: string;
  sessionId: string;
  sequence: number;
  clientTurnKey: string;
  tenantRef: string;
  workspaceId: string;
  applicationId: string;
  ownerSubjectRef: string;
  executionProfile: ApplicationInteractionExecutionProfile;
  authority: ApplicationInteractionAuthority;
  status: "running" | "succeeded" | "failed" | "canceled" | "outcome_unknown";
  inputDigest: string;
  inputBytes: number;
  runRef: ApplicationInteractionRunRef | null;
  failureCode: string;
  failureSummary: string;
  startedAt: string;
  completedAt: string;
  actorRef: string;
  requestId: string;
  auditRef: string;
};

export type ApplicationInteractionAnswer = {
  schemaVersion: typeof ANSWER_SCHEMA_VERSION;
  answer: string;
  citations: Array<{ fragmentRef: string; claimSummary: string }>;
  limitations: string[];
  confidence: "low" | "medium" | "high";
};

export type ApplicationInteractionSessionListResult = {
  status: "offline" | "ready" | "failed";
  sessions: ApplicationInteractionSession[];
  nextCursor: string;
  failureCode: string;
  requestId: string;
  auditRef: string;
  summary: string;
};

export type ApplicationInteractionSessionResult = {
  status: "offline" | "ready" | "failed" | "version_conflict";
  session: ApplicationInteractionSession | null;
  failureCode: string;
  currentRecordVersion: number;
  currentState: string;
  requestId: string;
  auditRef: string;
  summary: string;
};

export type ApplicationInteractionTurnListResult = {
  status: "offline" | "ready" | "failed";
  turns: ApplicationInteractionTurn[];
  failureCode: string;
  requestId: string;
  auditRef: string;
  summary: string;
};

export type ApplicationInteractionTurnResult = {
  status: "offline" | "succeeded" | "failed" | "canceled";
  session: ApplicationInteractionSession | null;
  turn: ApplicationInteractionTurn | null;
  advisoryOutput: string;
  answer: ApplicationInteractionAnswer | null;
  failureCode: string;
  failureSummary: string;
  idempotentReplay: boolean;
  requestId: string;
  auditRef: string;
  summary: string;
};

type SessionEnvelope = {
  request_id: string;
  tenant_ref: string;
  workspace_id: string;
  application_id: string;
  session: unknown;
  failure_code: string | null;
  current_record_version: number;
  current_state: string;
  idempotent_replay: boolean;
  audit_ref: string;
};

type SessionListEnvelope = {
  request_id: string;
  tenant_ref: string;
  workspace_id: string;
  application_id: string;
  items: unknown[];
  next_cursor: string | null;
  failure_code: string | null;
  audit_ref: string;
};

type TurnListEnvelope = {
  request_id: string;
  tenant_ref: string;
  workspace_id: string;
  application_id: string;
  session_id: string;
  items: unknown[];
  failure_code: string | null;
  audit_ref: string;
};

type TurnEnvelope = {
  request_id: string;
  tenant_ref: string;
  workspace_id: string;
  application_id: string;
  session_id: string;
  session: unknown;
  turn: unknown;
  advisory_output?: string;
  answer?: unknown;
  failure_code: string | null;
  failure_summary: string;
  idempotent_replay: boolean;
  audit_ref: string;
};

export function readApplicationInteractionSessionConfig(): ApplicationInteractionSessionConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  return {
    mode: env.VITE_RADISHMIND_APPLICATION_SESSION_SOURCE?.trim() === DEV_SOURCE
      ? "dev_application_session_http"
      : "offline",
    baseUrl: normalizeBaseUrl(
      env.VITE_RADISHMIND_APPLICATION_SESSION_BASE_URL ??
      env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ??
      DEFAULT_BASE_URL,
    ),
    tenantRef: env.VITE_RADISHMIND_DEV_READ_TENANT_REF?.trim() || "tenant_demo",
    workspaceId: env.VITE_RADISHMIND_APPLICATION_SESSION_WORKSPACE_ID?.trim() || "workspace_demo",
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || "subject_demo_user",
  };
}

export function initialApplicationInteractionSessionListResult(
  config: ApplicationInteractionSessionConfig,
): ApplicationInteractionSessionListResult {
  return config.mode === "offline"
    ? { status: "offline", sessions: [], nextCursor: "", failureCode: "application_session_http_disabled", requestId: "", auditRef: "", summary: "Offline mode sends zero application session requests." }
    : { status: "ready", sessions: [], nextCursor: "", failureCode: "", requestId: "", auditRef: "", summary: "Load sessions for the selected application." };
}

export async function listApplicationInteractionSessions(
  config: ApplicationInteractionSessionConfig,
  applicationId: string,
  signal?: AbortSignal,
): Promise<ApplicationInteractionSessionListResult> {
  if (config.mode === "offline") return initialApplicationInteractionSessionListResult(config);
  if (!validScope(config, applicationId)) return failedSessionList("application_session_payload_invalid");
  const requestId = createRequestId("application-session-list");
  const query = new URLSearchParams({ workspace_id: config.workspaceId, application_id: applicationId, limit: "100" });
  try {
    const response = await fetch(`${config.baseUrl}/v1/user-workspace/application-sessions?${query}`, {
      headers: managementHeaders(config, applicationId, requestId, "application_sessions:read"), signal,
    });
    const document: unknown = await response.json();
    if (!isSessionListEnvelope(document, config, applicationId)) return failedSessionList("application_session_store_contract_mismatch", requestId);
    if (!response.ok || document.failure_code) return failedSessionList(document.failure_code || "application_session_store_unavailable", document.request_id, document.audit_ref);
    const sessions = document.items.map((item) => parseSession(item, config, applicationId));
    if (sessions.some((session) => session === null)) return failedSessionList("application_session_store_contract_mismatch", document.request_id, document.audit_ref);
    return { status: "ready", sessions: sessions as ApplicationInteractionSession[], nextCursor: document.next_cursor ?? "", failureCode: "", requestId: document.request_id, auditRef: document.audit_ref, summary: `Loaded ${sessions.length} metadata-only sessions.` };
  } catch (error) {
    return failedSessionList(isAbort(error) ? "application_session_request_canceled" : "application_session_store_unavailable", requestId);
  }
}

export async function createApplicationInteractionSession(
  config: ApplicationInteractionSessionConfig,
  input: { applicationId: string; executionProfile: ApplicationInteractionExecutionProfile; definitionId: string },
  signal?: AbortSignal,
): Promise<ApplicationInteractionSessionResult> {
  if (config.mode === "offline") return offlineSessionResult();
  const definitionId = input.definitionId.trim();
  if (!validScope(config, input.applicationId) || !validProfileBinding(input.executionProfile, definitionId)) {
    return failedSession("application_session_payload_invalid");
  }
  const requestId = createRequestId("application-session-create");
  const body: Record<string, unknown> = {
    workspace_id: config.workspaceId,
    application_id: input.applicationId,
    execution_profile: input.executionProfile,
  };
  if (input.executionProfile === "workflow_definition_executor_v1") body.definition_id = definitionId;
  return fetchSession(config, input.applicationId, requestId, `${config.baseUrl}/v1/user-workspace/application-sessions`, {
    method: "POST", headers: { ...managementHeaders(config, input.applicationId, requestId, "application_sessions:write"), "Content-Type": "application/json" }, body: JSON.stringify(body), signal,
  });
}

export async function closeApplicationInteractionSession(
  config: ApplicationInteractionSessionConfig,
  session: ApplicationInteractionSession,
  signal?: AbortSignal,
): Promise<ApplicationInteractionSessionResult> {
  if (config.mode === "offline") return offlineSessionResult();
  if (!validScope(config, session.applicationId) || !SESSION_ID_PATTERN.test(session.sessionId) || !isPositiveInteger(session.recordVersion)) {
    return failedSession("application_session_payload_invalid");
  }
  const requestId = createRequestId("application-session-close");
  return fetchSession(config, session.applicationId, requestId, `${config.baseUrl}/v1/user-workspace/application-sessions/${encodeURIComponent(session.sessionId)}/close`, {
    method: "POST", headers: { ...managementHeaders(config, session.applicationId, requestId, "application_sessions:write"), "Content-Type": "application/json" },
    body: JSON.stringify({ workspace_id: config.workspaceId, application_id: session.applicationId, expected_version: session.recordVersion }), signal,
  });
}

export async function listApplicationInteractionTurns(
  config: ApplicationInteractionSessionConfig,
  session: ApplicationInteractionSession,
  signal?: AbortSignal,
): Promise<ApplicationInteractionTurnListResult> {
  if (config.mode === "offline") return { status: "offline", turns: [], failureCode: "application_session_http_disabled", requestId: "", auditRef: "", summary: "Offline mode sends zero application session requests." };
  if (!validScope(config, session.applicationId) || !SESSION_ID_PATTERN.test(session.sessionId)) return failedTurnList("application_session_payload_invalid");
  const requestId = createRequestId("application-turn-list");
  const query = new URLSearchParams({ workspace_id: config.workspaceId, application_id: session.applicationId });
  try {
    const response = await fetch(`${config.baseUrl}/v1/user-workspace/application-sessions/${encodeURIComponent(session.sessionId)}/turns?${query}`, {
      headers: managementHeaders(config, session.applicationId, requestId, "application_sessions:read"), signal,
    });
    const document: unknown = await response.json();
    if (!isTurnListEnvelope(document, config, session.applicationId, session.sessionId)) return failedTurnList("application_session_store_contract_mismatch", requestId);
    if (!response.ok || document.failure_code) return failedTurnList(document.failure_code || "application_session_store_unavailable", document.request_id, document.audit_ref);
    const turns = document.items.map((item) => parseTurn(item, config, session.applicationId, session.sessionId, session.executionProfile));
    if (turns.some((turn) => turn === null)) return failedTurnList("application_session_store_contract_mismatch", document.request_id, document.audit_ref);
    return { status: "ready", turns: turns as ApplicationInteractionTurn[], failureCode: "", requestId: document.request_id, auditRef: document.audit_ref, summary: `Loaded ${turns.length} persisted turn metadata records without transcript content.` };
  } catch (error) {
    return failedTurnList(isAbort(error) ? "application_session_request_canceled" : "application_session_store_unavailable", requestId);
  }
}

export async function executeApplicationInteractionTurn(
  config: ApplicationInteractionSessionConfig,
  input: {
    session: ApplicationInteractionSession;
    clientTurnKey: string;
    inputText: string;
    conditionValues?: Record<string, boolean>;
    model?: string;
    temperature?: number | null;
  },
  signal?: AbortSignal,
): Promise<ApplicationInteractionTurnResult> {
  if (config.mode === "offline") return offlineTurnResult();
  const text = input.inputText.trim();
  const conditionValues = input.conditionValues ?? {};
  const model = input.model?.trim() ?? "";
  const temperature = input.temperature ?? null;
  if (!validScope(config, input.session.applicationId) || input.session.state !== "active" ||
    !SESSION_ID_PATTERN.test(input.session.sessionId) || !REF_PATTERN.test(input.clientTurnKey) ||
    text.length === 0 || new TextEncoder().encode(text).length > 8192 || containsSensitiveText(text) ||
    Object.keys(conditionValues).length > 16 || !Object.entries(conditionValues).every(([key, value]) => REF_PATTERN.test(key) && typeof value === "boolean") ||
    model.length > 256 || model.includes("://") || containsSensitiveText(model) ||
    (input.session.executionProfile === "application_rag_invocation_v1" && (Object.keys(conditionValues).length > 0 || model !== "" || temperature !== null))) {
    return failedTurn(containsSensitiveText(text) || containsSensitiveText(model) ? "application_session_secret_material_forbidden" : "application_session_payload_invalid");
  }
  const requestId = createRequestId("application-session-turn");
  try {
    const response = await fetch(`${config.baseUrl}/v1/user-workspace/application-sessions/${encodeURIComponent(input.session.sessionId)}/turns`, {
      method: "POST",
      headers: { ...managementHeaders(config, input.session.applicationId, requestId, "application_sessions:execute"), "Content-Type": "application/json" },
      body: JSON.stringify({
        workspace_id: config.workspaceId,
        application_id: input.session.applicationId,
        expected_session_version: input.session.recordVersion,
        client_turn_key: input.clientTurnKey,
        input_text: text,
        condition_values: conditionValues,
        model,
        temperature,
      }),
      signal,
    });
    const document: unknown = await response.json();
    if (!isTurnEnvelope(document, config, input.session.applicationId, input.session.sessionId)) return failedTurn("application_session_store_contract_mismatch", requestId);
    const session = document.session === null ? null : parseSession(document.session, config, input.session.applicationId);
    const turn = document.turn === null ? null : parseTurn(document.turn, config, input.session.applicationId, input.session.sessionId, input.session.executionProfile);
    const answer = document.answer === undefined || document.answer === null ? null : parseAnswer(document.answer);
    if ((document.session !== null && !session) || (document.turn !== null && !turn) || (document.answer !== undefined && document.answer !== null && !answer)) {
      return failedTurn("application_session_store_contract_mismatch", document.request_id);
    }
    if (!response.ok || document.failure_code || !session || !turn) {
      return { status: turn?.status === "canceled" ? "canceled" : "failed", session, turn, advisoryOutput: "", answer: null, failureCode: document.failure_code || "application_session_store_unavailable", failureSummary: document.failure_summary, idempotentReplay: document.idempotent_replay, requestId: document.request_id, auditRef: document.audit_ref, summary: document.failure_summary || document.failure_code || "Application session turn failed closed." };
    }
    if (turn.status !== "succeeded" || (!document.advisory_output && !answer) || (document.advisory_output && answer)) return failedTurn("application_session_store_contract_mismatch", document.request_id);
    return { status: "succeeded", session, turn, advisoryOutput: document.advisory_output ?? "", answer, failureCode: "", failureSummary: "", idempotentReplay: document.idempotent_replay, requestId: document.request_id, auditRef: document.audit_ref, summary: "Turn succeeded. Request and response content remain only in current component memory." };
  } catch (error) {
    return failedTurn(isAbort(error) ? "application_session_request_canceled" : "application_session_store_unavailable", requestId);
  }
}

export function applicationInteractionResponseMatchesScope(
  expected: { generation: number; applicationId: string; sessionId: string; clientTurnKey: string },
  observed: { generation: number; applicationId: string; sessionId: string; clientTurnKey: string },
): boolean {
  return expected.generation === observed.generation && expected.applicationId === observed.applicationId &&
    expected.sessionId === observed.sessionId && expected.clientTurnKey === observed.clientTurnKey;
}

async function fetchSession(
  config: ApplicationInteractionSessionConfig,
  applicationId: string,
  requestId: string,
  url: string,
  init: RequestInit,
): Promise<ApplicationInteractionSessionResult> {
  try {
    const response = await fetch(url, init);
    const document: unknown = await response.json();
    if (!isSessionEnvelope(document, config, applicationId)) return failedSession("application_session_store_contract_mismatch", requestId);
    const session = document.session === null ? null : parseSession(document.session, config, applicationId);
    if (document.session !== null && !session) return failedSession("application_session_store_contract_mismatch", document.request_id, document.audit_ref);
    if (!response.ok || document.failure_code || !session) {
      return { status: document.failure_code === "application_session_version_conflict" ? "version_conflict" : "failed", session: null, failureCode: document.failure_code || "application_session_store_unavailable", currentRecordVersion: document.current_record_version, currentState: document.current_state, requestId: document.request_id, auditRef: document.audit_ref, summary: sessionFailureSummary(document.failure_code || "application_session_store_unavailable") };
    }
    return { status: "ready", session, failureCode: "", currentRecordVersion: session.recordVersion, currentState: session.state, requestId: document.request_id, auditRef: document.audit_ref, summary: `Session ${session.sessionId} is ${session.state}.` };
  } catch (error) {
    return failedSession(isAbort(error) ? "application_session_request_canceled" : "application_session_store_unavailable", requestId);
  }
}

function parseSession(value: unknown, config: ApplicationInteractionSessionConfig, applicationId: string): ApplicationInteractionSession | null {
  if (!isExactRecord(value, ["schema_version", "session_id", "tenant_ref", "workspace_id", "application_id", "owner_subject_ref", "state", "record_version", "profile_binding", "authority", "content_retention", "turn_count", "last_turn_id", "created_at", "updated_at", "closed_at", "created_by_actor_ref", "updated_by_actor_ref", "request_id", "audit_ref"]) ||
    containsForbiddenField(value) || value.schema_version !== SESSION_SCHEMA_VERSION || !SESSION_ID_PATTERN.test(String(value.session_id)) ||
    value.tenant_ref !== config.tenantRef || value.workspace_id !== config.workspaceId || value.application_id !== applicationId ||
    value.owner_subject_ref !== config.subjectRef || (value.state !== "active" && value.state !== "closed") || !isPositiveInteger(value.record_version) ||
    !isRecord(value.profile_binding) || !isNonNegativeInteger(value.turn_count) ||
    !(value.last_turn_id === null || TURN_ID_PATTERN.test(String(value.last_turn_id))) || value.content_retention !== "metadata_only" ||
    !isTimestamp(value.created_at) || !isTimestamp(value.updated_at) || !(value.closed_at === null || isTimestamp(value.closed_at)) ||
    (value.state === "active" && value.closed_at !== null) || (value.state === "closed" && value.closed_at === null) ||
    !REF_PATTERN.test(String(value.created_by_actor_ref)) || !REF_PATTERN.test(String(value.updated_by_actor_ref)) ||
    !REF_PATTERN.test(String(value.request_id)) || !REF_PATTERN.test(String(value.audit_ref))) return null;
  const binding = parseProfileBinding(value.profile_binding);
  const authority = parseAuthority(value.authority, applicationId);
  if (!binding || !authority || binding.executionProfile !== authority.executionProfile) return null;
  return {
    schemaVersion: SESSION_SCHEMA_VERSION, sessionId: String(value.session_id), tenantRef: config.tenantRef,
    workspaceId: config.workspaceId, applicationId, ownerSubjectRef: String(value.owner_subject_ref), state: value.state,
    recordVersion: Number(value.record_version), executionProfile: binding.executionProfile, definitionId: binding.definitionId,
    authority, contentRetention: "metadata_only", turnCount: Number(value.turn_count), lastTurnId: value.last_turn_id === null ? "" : String(value.last_turn_id),
    createdAt: String(value.created_at), updatedAt: String(value.updated_at), closedAt: value.closed_at === null ? "" : String(value.closed_at),
    requestId: String(value.request_id), auditRef: String(value.audit_ref),
  };
}

function parseTurn(value: unknown, config: ApplicationInteractionSessionConfig, applicationId: string, sessionId: string, profile: ApplicationInteractionExecutionProfile): ApplicationInteractionTurn | null {
  if (!isExactRecord(value, ["schema_version", "turn_id", "session_id", "sequence", "client_turn_key", "tenant_ref", "workspace_id", "application_id", "owner_subject_ref", "execution_profile", "authority", "status", "input_digest", "input_bytes", "run_ref", "failure_code", "failure_summary", "started_at", "completed_at", "actor_ref", "request_id", "audit_ref"]) ||
    containsForbiddenField(value) || value.schema_version !== TURN_SCHEMA_VERSION || !TURN_ID_PATTERN.test(String(value.turn_id)) || value.session_id !== sessionId ||
    !isPositiveInteger(value.sequence) || !REF_PATTERN.test(String(value.client_turn_key)) || value.tenant_ref !== config.tenantRef ||
    value.workspace_id !== config.workspaceId || value.application_id !== applicationId || value.owner_subject_ref !== config.subjectRef ||
    value.execution_profile !== profile || !isTurnStatus(value.status) || !DIGEST_PATTERN.test(String(value.input_digest)) ||
    !isPositiveInteger(value.input_bytes) || Number(value.input_bytes) > 8192 || typeof value.failure_code !== "string" || value.failure_code.length > 160 ||
    typeof value.failure_summary !== "string" || value.failure_summary.length > 256 || !isTimestamp(value.started_at) ||
    !(value.completed_at === null || isTimestamp(value.completed_at)) || !REF_PATTERN.test(String(value.actor_ref)) ||
    !REF_PATTERN.test(String(value.request_id)) || !REF_PATTERN.test(String(value.audit_ref))) return null;
  const authority = parseAuthority(value.authority, applicationId);
  const runRef = value.run_ref === null ? null : parseRunRef(value.run_ref, profile);
  if (!authority || authority.executionProfile !== profile || (value.run_ref !== null && !runRef) ||
    (value.status === "running" && (runRef !== null || value.failure_code !== "" || value.failure_summary !== "" || value.completed_at !== null)) ||
    (value.status !== "running" && value.completed_at === null) ||
    (value.status === "succeeded" && (!runRef || value.failure_code !== "" || value.failure_summary !== "")) ||
    (["failed", "canceled", "outcome_unknown"].includes(value.status) && value.failure_code.length === 0)) return null;
  return {
    schemaVersion: TURN_SCHEMA_VERSION, turnId: String(value.turn_id), sessionId, sequence: Number(value.sequence), clientTurnKey: String(value.client_turn_key),
    tenantRef: config.tenantRef, workspaceId: config.workspaceId, applicationId, ownerSubjectRef: String(value.owner_subject_ref), executionProfile: profile,
    authority, status: value.status, inputDigest: String(value.input_digest), inputBytes: Number(value.input_bytes), runRef,
    failureCode: value.failure_code, failureSummary: value.failure_summary, startedAt: String(value.started_at),
    completedAt: value.completed_at === null ? "" : String(value.completed_at), actorRef: String(value.actor_ref), requestId: String(value.request_id), auditRef: String(value.audit_ref),
  };
}

function parseAuthority(value: unknown, applicationId: string): ApplicationInteractionAuthority | null {
  if (!isExactRecord(value, ["schema_version", "execution_profile", "application_id", "application_record_version", "application_lifecycle", "workflow_definition", "application_rag", "authority_digest"]) ||
    containsForbiddenField(value) || value.schema_version !== AUTHORITY_SCHEMA_VERSION || !isExecutionProfile(value.execution_profile) ||
    value.application_id !== applicationId || !isPositiveInteger(value.application_record_version) || value.application_lifecycle !== "active" ||
    !DIGEST_PATTERN.test(String(value.authority_digest))) return null;
  const workflowDefinition = value.workflow_definition === null ? null : parseWorkflowDefinitionAuthority(value.workflow_definition);
  const applicationRAG = value.application_rag === null ? null : parseApplicationRAGAuthority(value.application_rag);
  if ((value.execution_profile === "workflow_definition_executor_v1" && (!workflowDefinition || applicationRAG)) ||
    (value.execution_profile === "application_rag_invocation_v1" && (workflowDefinition || !applicationRAG))) return null;
  return { schemaVersion: AUTHORITY_SCHEMA_VERSION, executionProfile: value.execution_profile, applicationId, applicationRecordVersion: Number(value.application_record_version), applicationLifecycle: "active", workflowDefinition, applicationRAG, authorityDigest: String(value.authority_digest) };
}

function parseWorkflowDefinitionAuthority(value: unknown): NonNullable<ApplicationInteractionAuthority["workflowDefinition"]> | null {
  if (!isExactRecord(value, ["definition_id", "definition_version", "definition_digest", "activation_pointer_version", "candidate_id", "candidate_review_version"]) ||
    !REF_PATTERN.test(String(value.definition_id)) || !isPositiveInteger(value.definition_version) || !DIGEST_PATTERN.test(String(value.definition_digest)) ||
    !isPositiveInteger(value.activation_pointer_version) || !REF_PATTERN.test(String(value.candidate_id)) || !isPositiveInteger(value.candidate_review_version)) return null;
  return { definitionId: String(value.definition_id), definitionVersion: Number(value.definition_version), definitionDigest: String(value.definition_digest), activationPointerVersion: Number(value.activation_pointer_version), candidateId: String(value.candidate_id), candidateReviewVersion: Number(value.candidate_review_version) };
}

function parseApplicationRAGAuthority(value: unknown): NonNullable<ApplicationInteractionAuthority["applicationRAG"]> | null {
  if (!isExactRecord(value, ["assignment_id", "assignment_version", "assignment_digest", "publish_candidate_id", "publish_review_version", "draft_id", "draft_version", "draft_digest", "binding_ref", "snapshot_id", "snapshot_version", "snapshot_digest", "retrieval_profile_id", "retrieval_profile_version", "retrieval_profile_digest"]) ||
    !RAG_ASSIGNMENT_ID_PATTERN.test(String(value.assignment_id)) || !isPositiveInteger(value.assignment_version) || !DIGEST_PATTERN.test(String(value.assignment_digest)) ||
    !REF_PATTERN.test(String(value.publish_candidate_id)) || !isPositiveInteger(value.publish_review_version) || !REF_PATTERN.test(String(value.draft_id)) ||
    !isPositiveInteger(value.draft_version) || !DIGEST_PATTERN.test(String(value.draft_digest)) || !isExactRecord(value.binding_ref, ["binding_id", "binding_version", "binding_digest"]) ||
    !RAG_BINDING_ID_PATTERN.test(String(value.binding_ref.binding_id)) || value.binding_ref.binding_version !== 1 || !DIGEST_PATTERN.test(String(value.binding_ref.binding_digest)) ||
    !RAG_SNAPSHOT_ID_PATTERN.test(String(value.snapshot_id)) || !isPositiveInteger(value.snapshot_version) || !DIGEST_PATTERN.test(String(value.snapshot_digest)) ||
    !REF_PATTERN.test(String(value.retrieval_profile_id)) || !isPositiveInteger(value.retrieval_profile_version) || !DIGEST_PATTERN.test(String(value.retrieval_profile_digest))) return null;
  return {
    assignmentId: String(value.assignment_id), assignmentVersion: Number(value.assignment_version), assignmentDigest: String(value.assignment_digest),
    publishCandidateId: String(value.publish_candidate_id), publishReviewVersion: Number(value.publish_review_version), draftId: String(value.draft_id),
    draftVersion: Number(value.draft_version), draftDigest: String(value.draft_digest), bindingId: String(value.binding_ref.binding_id), bindingVersion: 1,
    bindingDigest: String(value.binding_ref.binding_digest), snapshotId: String(value.snapshot_id), snapshotVersion: Number(value.snapshot_version),
    snapshotDigest: String(value.snapshot_digest), retrievalProfileId: String(value.retrieval_profile_id), retrievalProfileVersion: Number(value.retrieval_profile_version),
    retrievalProfileDigest: String(value.retrieval_profile_digest),
  };
}

function parseProfileBinding(value: Record<string, unknown>): { executionProfile: ApplicationInteractionExecutionProfile; definitionId: string } | null {
  if (!isExecutionProfile(value.execution_profile)) return null;
  if (value.execution_profile === "workflow_definition_executor_v1" && isExactRecord(value, ["execution_profile", "definition_id"]) && REF_PATTERN.test(String(value.definition_id))) {
    return { executionProfile: value.execution_profile, definitionId: String(value.definition_id) };
  }
  if (value.execution_profile === "application_rag_invocation_v1" && isExactRecord(value, ["execution_profile"])) return { executionProfile: value.execution_profile, definitionId: "" };
  return null;
}

function parseRunRef(value: unknown, profile: ApplicationInteractionExecutionProfile): ApplicationInteractionRunRef | null {
  if (!isExactRecord(value, ["run_id", "schema_version"]) || !RUN_ID_PATTERN.test(String(value.run_id))) return null;
  const expectedSchema = profile === "workflow_definition_executor_v1" ? "workflow_run_record.v5" : "workflow_run_record.v4";
  return value.schema_version === expectedSchema ? { runId: String(value.run_id), schemaVersion: expectedSchema } : null;
}

function parseAnswer(value: unknown): ApplicationInteractionAnswer | null {
  if (!isExactRecord(value, ["schema_version", "answer", "citations", "limitations", "confidence"]) || containsForbiddenFieldExceptAnswer(value) ||
    value.schema_version !== ANSWER_SCHEMA_VERSION || typeof value.answer !== "string" || value.answer.length < 1 || value.answer.length > 16384 ||
    !Array.isArray(value.citations) || value.citations.length < 1 || value.citations.length > 8 || !Array.isArray(value.limitations) || value.limitations.length > 8 ||
    !value.limitations.every((item) => typeof item === "string" && item.length >= 1 && item.length <= 512) ||
    (value.confidence !== "low" && value.confidence !== "medium" && value.confidence !== "high")) return null;
  const citations = value.citations.map((item) => {
    if (!isExactRecord(item, ["fragment_ref", "claim_summary"]) || !FRAGMENT_REF_PATTERN.test(String(item.fragment_ref)) || typeof item.claim_summary !== "string" || item.claim_summary.length < 1 || item.claim_summary.length > 512) return null;
    return { fragmentRef: String(item.fragment_ref), claimSummary: item.claim_summary };
  });
  if (citations.some((item) => item === null) || new Set(citations.map((item) => item?.fragmentRef)).size !== citations.length) return null;
  return { schemaVersion: ANSWER_SCHEMA_VERSION, answer: value.answer, citations: citations as Array<{ fragmentRef: string; claimSummary: string }>, limitations: [...value.limitations], confidence: value.confidence };
}

function isSessionEnvelope(value: unknown, config: ApplicationInteractionSessionConfig, applicationId: string): value is SessionEnvelope {
  return isExactRecord(value, ["request_id", "tenant_ref", "workspace_id", "application_id", "session", "failure_code", "current_record_version", "current_state", "idempotent_replay", "audit_ref"]) &&
    !containsForbiddenField(value) && envelopeScopeMatches(value, config, applicationId) && (value.session === null || isRecord(value.session)) &&
    (value.failure_code === null || typeof value.failure_code === "string") && isNonNegativeInteger(value.current_record_version) && typeof value.current_state === "string" &&
    typeof value.idempotent_replay === "boolean";
}

function isSessionListEnvelope(value: unknown, config: ApplicationInteractionSessionConfig, applicationId: string): value is SessionListEnvelope {
  return isExactRecord(value, ["request_id", "tenant_ref", "workspace_id", "application_id", "items", "next_cursor", "failure_code", "audit_ref"]) &&
    !containsForbiddenField(value) && envelopeScopeMatches(value, config, applicationId) && Array.isArray(value.items) &&
    (value.next_cursor === null || REF_PATTERN.test(String(value.next_cursor))) && (value.failure_code === null || typeof value.failure_code === "string");
}

function isTurnListEnvelope(value: unknown, config: ApplicationInteractionSessionConfig, applicationId: string, sessionId: string): value is TurnListEnvelope {
  return isExactRecord(value, ["request_id", "tenant_ref", "workspace_id", "application_id", "session_id", "items", "failure_code", "audit_ref"]) &&
    !containsForbiddenField(value) && envelopeScopeMatches(value, config, applicationId) && value.session_id === sessionId && Array.isArray(value.items) &&
    (value.failure_code === null || typeof value.failure_code === "string");
}

function isTurnEnvelope(value: unknown, config: ApplicationInteractionSessionConfig, applicationId: string, sessionId: string): value is TurnEnvelope {
  if (!isRecord(value)) return false;
  const allowed = new Set(["request_id", "tenant_ref", "workspace_id", "application_id", "session_id", "session", "turn", "advisory_output", "answer", "failure_code", "failure_summary", "idempotent_replay", "audit_ref"]);
  const required = ["request_id", "tenant_ref", "workspace_id", "application_id", "session_id", "session", "turn", "failure_code", "failure_summary", "idempotent_replay", "audit_ref"];
  return required.every((key) => Object.hasOwn(value, key)) && Object.keys(value).every((key) => allowed.has(key)) &&
    !containsForbiddenFieldExceptAnswer(value) && envelopeScopeMatches(value, config, applicationId) && value.session_id === sessionId &&
    (value.session === null || isRecord(value.session)) && (value.turn === null || isRecord(value.turn)) &&
    (value.advisory_output === undefined || typeof value.advisory_output === "string") && (value.answer === undefined || value.answer === null || isRecord(value.answer)) &&
    (value.failure_code === null || typeof value.failure_code === "string") && typeof value.failure_summary === "string" && typeof value.idempotent_replay === "boolean";
}

function envelopeScopeMatches(value: Record<string, unknown>, config: ApplicationInteractionSessionConfig, applicationId: string): boolean {
  return REF_PATTERN.test(String(value.request_id)) && value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId &&
    value.application_id === applicationId && REF_PATTERN.test(String(value.audit_ref));
}

function managementHeaders(config: ApplicationInteractionSessionConfig, applicationId: string, requestId: string, scope: string) {
  return {
    "X-Request-Id": requestId,
    "X-RadishMind-Dev-Read-Identity": `application-interaction-session-web:${config.subjectRef}`,
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Workflow-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Workflow-Application": applicationId,
    "X-RadishMind-Dev-Read-Scopes": scope,
  };
}

function validScope(config: ApplicationInteractionSessionConfig, applicationId: string): boolean {
  return APPLICATION_ID_PATTERN.test(applicationId) && REF_PATTERN.test(config.tenantRef) && REF_PATTERN.test(config.workspaceId) && REF_PATTERN.test(config.subjectRef);
}

function validProfileBinding(profile: ApplicationInteractionExecutionProfile, definitionId: string): boolean {
  return profile === "workflow_definition_executor_v1" ? REF_PATTERN.test(definitionId) : profile === "application_rag_invocation_v1" && definitionId === "";
}

function containsSensitiveText(value: string): boolean {
  return /(authorization\s*:|bearer\s+|api[_ -]?key|token\s*[=:]|password\s*[=:]|secret\s*[=:]|cookie\s*:|postgres(?:ql)?:\/\/)/iu.test(value);
}

function containsForbiddenField(value: unknown): boolean {
  if (Array.isArray(value)) return value.some(containsForbiddenField);
  if (!isRecord(value)) return false;
  return Object.entries(value).some(([key, nested]) => FORBIDDEN_RESPONSE_FIELDS.has(key.toLowerCase()) || containsForbiddenField(nested));
}

function containsForbiddenFieldExceptAnswer(value: unknown): boolean {
  if (Array.isArray(value)) return value.some(containsForbiddenFieldExceptAnswer);
  if (!isRecord(value)) return false;
  return Object.entries(value).some(([key, nested]) => key !== "answer" && (FORBIDDEN_RESPONSE_FIELDS.has(key.toLowerCase()) || containsForbiddenFieldExceptAnswer(nested)));
}

function isExactRecord(value: unknown, keys: string[]): value is Record<string, any> {
  return isRecord(value) && Object.keys(value).length === keys.length && keys.every((key) => Object.hasOwn(value, key));
}

function isRecord(value: unknown): value is Record<string, any> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function isExecutionProfile(value: unknown): value is ApplicationInteractionExecutionProfile {
  return value === "workflow_definition_executor_v1" || value === "application_rag_invocation_v1";
}

function isTurnStatus(value: unknown): value is ApplicationInteractionTurn["status"] {
  return value === "running" || value === "succeeded" || value === "failed" || value === "canceled" || value === "outcome_unknown";
}

function isPositiveInteger(value: unknown): boolean {
  return typeof value === "number" && Number.isInteger(value) && value > 0;
}

function isNonNegativeInteger(value: unknown): boolean {
  return typeof value === "number" && Number.isInteger(value) && value >= 0;
}

function isTimestamp(value: unknown): boolean {
  return typeof value === "string" && value.length >= 20 && Number.isFinite(Date.parse(value));
}

function isAbort(error: unknown): boolean {
  return error instanceof DOMException && error.name === "AbortError";
}

function offlineSessionResult(): ApplicationInteractionSessionResult {
  return { status: "offline", session: null, failureCode: "application_session_http_disabled", currentRecordVersion: 0, currentState: "", requestId: "", auditRef: "", summary: "Offline mode sends zero application session requests." };
}

function offlineTurnResult(): ApplicationInteractionTurnResult {
  return { status: "offline", session: null, turn: null, advisoryOutput: "", answer: null, failureCode: "application_session_http_disabled", failureSummary: "", idempotentReplay: false, requestId: "", auditRef: "", summary: "Offline mode sends zero application session requests." };
}

function failedSessionList(failureCode: string, requestId = "", auditRef = ""): ApplicationInteractionSessionListResult {
  return { status: "failed", sessions: [], nextCursor: "", failureCode, requestId, auditRef, summary: sessionFailureSummary(failureCode) };
}

function failedSession(failureCode: string, requestId = "", auditRef = ""): ApplicationInteractionSessionResult {
  return { status: failureCode === "application_session_version_conflict" ? "version_conflict" : "failed", session: null, failureCode, currentRecordVersion: 0, currentState: "", requestId, auditRef, summary: sessionFailureSummary(failureCode) };
}

function failedTurnList(failureCode: string, requestId = "", auditRef = ""): ApplicationInteractionTurnListResult {
  return { status: "failed", turns: [], failureCode, requestId, auditRef, summary: sessionFailureSummary(failureCode) };
}

function failedTurn(failureCode: string, requestId = ""): ApplicationInteractionTurnResult {
  return { status: failureCode === "application_session_request_canceled" ? "canceled" : "failed", session: null, turn: null, advisoryOutput: "", answer: null, failureCode, failureSummary: "", idempotentReplay: false, requestId, auditRef: "", summary: sessionFailureSummary(failureCode) };
}

function sessionFailureSummary(failureCode: string): string {
  const summaries: Record<string, string> = {
    application_session_http_disabled: "Application session HTTP is disabled; offline mode sends zero requests.",
    application_session_payload_invalid: "Application session input is invalid.",
    application_session_secret_material_forbidden: "Credential or secret material is forbidden in session input.",
    application_session_request_canceled: "The current turn request was canceled and its eventual response is ignored.",
    application_session_version_conflict: "The session changed. Reload current metadata before continuing.",
    application_session_authority_changed: "Application runtime authority changed before delegated execution.",
    application_session_store_contract_mismatch: "Application session response failed strict schema, scope, or sensitive-field validation.",
  };
  return summaries[failureCode] ?? `Application session operation failed: ${failureCode}.`;
}

function normalizeBaseUrl(value: string): string {
  return value.trim().replace(/\/$/u, "");
}

function createRequestId(prefix: string): string {
  const suffix = (globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random()}`).replaceAll("-", "").slice(0, 20);
  return `${prefix}-${suffix}`;
}
