const ARTIFACT_SCHEMA_VERSION = "application_result_artifact.v1";
const SUMMARY_SCHEMA_VERSION = "application_result_artifact_summary.v2";
const LIFECYCLE_SCHEMA_VERSION = "application_result_artifact_lifecycle.v1";
const LIFECYCLE_EVENT_SCHEMA_VERSION = "application_result_artifact_lifecycle_event.v1";
const EXPORT_SCHEMA_VERSION = "application_result_artifact_export.v1";
const APPLICATION_ID_PATTERN = /^app_[a-z0-9]{16}$/u;
const SESSION_ID_PATTERN = /^appsess_[a-z2-7]{16}$/u;
const ARTIFACT_ID_PATTERN = /^appres_[a-z2-7]{16}$/u;
const TURN_ID_PATTERN = /^appturn_[a-z2-7]{16}$/u;
const RUN_ID_PATTERN = /^run_[a-z0-9]{16,64}$/u;
const REF_PATTERN = /^[A-Za-z0-9][A-Za-z0-9_.:/-]{0,159}$/u;
const DIGEST_PATTERN = /^sha256:[a-f0-9]{64}$/u;

type Document = Record<string, unknown>;

export type ApplicationResultArtifactConfig = {
  mode: "offline" | "dev_application_session_http";
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  subjectRef: string;
};

export type ApplicationResultArtifactLifecycleState = "active" | "archived";
export type ApplicationResultArtifactContentType = "text/markdown" | "application/json";
export type ApplicationResultArtifactExecutionProfile =
  | "workflow_definition_executor_v1"
  | "workflow_definition_executor_v2"
  | "application_rag_invocation_v1"
  | "prompt_application_invocation_v1"
  | "agent_copilot_suggestion_v1";

export type ApplicationResultArtifactRunRef = {
  schemaVersion:
    | "workflow_run_record.v4"
    | "workflow_run_record.v5"
    | "workflow_run_record.v6"
    | "workflow_run_record.v7"
    | "workflow_run_record.v8";
  runId: string;
};

export type ApplicationResultArtifactSummary = {
  schemaVersion: typeof SUMMARY_SCHEMA_VERSION;
  artifactId: string;
  recordVersion: 1;
  tenantRef: string;
  workspaceId: string;
  applicationId: string;
  ownerSubjectRef: string;
  sessionId: string;
  turnId: string;
  clientTurnKey: string;
  executionProfile: string;
  runRef: ApplicationResultArtifactRunRef;
  contentType: ApplicationResultArtifactContentType;
  contentBytes: number;
  contentDigest: string;
  createdAt: string;
  lifecycleState: ApplicationResultArtifactLifecycleState;
  lifecycleVersion: number;
  archivedAt: string | null;
  lifecycleUpdatedAt: string;
};

export type ApplicationResultArtifact = Omit<
  ApplicationResultArtifactSummary,
  "schemaVersion" | "lifecycleState" | "lifecycleVersion" | "archivedAt" | "lifecycleUpdatedAt"
> & {
  schemaVersion: typeof ARTIFACT_SCHEMA_VERSION;
  content: string;
  createdByActorRef: string;
  requestId: string;
  auditRef: string;
};

export type ApplicationResultArtifactLifecycle = {
  schemaVersion: typeof LIFECYCLE_SCHEMA_VERSION;
  tenantRef: string;
  workspaceId: string;
  applicationId: string;
  ownerSubjectRef: string;
  artifactId: string;
  lifecycleState: ApplicationResultArtifactLifecycleState;
  lifecycleVersion: number;
  archivedAt: string | null;
  updatedAt: string;
  updatedByActorRef: string;
  requestId: string;
  auditRef: string;
};

export type ApplicationResultArtifactLifecycleEvent = {
  schemaVersion: typeof LIFECYCLE_EVENT_SCHEMA_VERSION;
  tenantRef: string;
  workspaceId: string;
  applicationId: string;
  ownerSubjectRef: string;
  artifactId: string;
  lifecycleVersion: number;
  fromState: ApplicationResultArtifactLifecycleState;
  toState: ApplicationResultArtifactLifecycleState;
  transitionKind: "archived" | "unarchived";
  occurredAt: string;
  actorRef: string;
  requestId: string;
  auditRef: string;
};

export type ApplicationResultArtifactExport = {
  schemaVersion: typeof EXPORT_SCHEMA_VERSION;
  artifact: ApplicationResultArtifact;
  lifecycle: ApplicationResultArtifactLifecycle;
  exportedAt: string;
  exportedByActorRef: string;
  requestId: string;
  auditRef: string;
  exportDigest: string;
};

export type ApplicationResultArtifactListResult = {
  status: "offline" | "ready" | "failed";
  items: ApplicationResultArtifactSummary[];
  nextCursor: string;
  failureCode: string;
  requestId: string;
  auditRef: string;
  summary: string;
};

export type ApplicationResultArtifactReadResult = {
  status: "offline" | "ready" | "failed";
  artifact: ApplicationResultArtifact | null;
  lifecycle: ApplicationResultArtifactLifecycle | null;
  failureCode: string;
  requestId: string;
  auditRef: string;
  summary: string;
};

export type ApplicationResultArtifactTransitionResult = {
  status: "offline" | "ready" | "version_conflict" | "state_conflict" | "failed";
  lifecycle: ApplicationResultArtifactLifecycle | null;
  event: ApplicationResultArtifactLifecycleEvent | null;
  failureCode: string;
  currentLifecycleVersion: number;
  currentLifecycleState: ApplicationResultArtifactLifecycleState | "";
  requestId: string;
  auditRef: string;
  summary: string;
};

export type ApplicationResultArtifactExportResult = {
  status: "offline" | "ready" | "failed";
  exportDocument: ApplicationResultArtifactExport | null;
  failureCode: string;
  requestId: string;
  auditRef: string;
  summary: string;
};

export type ApplicationResultArtifactRequestScope = {
  generation: number;
  applicationId: string;
  sessionId: string;
  lifecycleState: ApplicationResultArtifactLifecycleState;
  artifactId: string;
};

export type ApplicationResultArtifactLibraryRequestScope = {
  generation: number;
  applicationId: string;
  lifecycleState: ApplicationResultArtifactLifecycleState;
  executionProfile: ApplicationResultArtifactExecutionProfile | "";
  contentType: ApplicationResultArtifactContentType | "";
  cursor: string;
  sessionId: string;
  artifactId: string;
};

export function initialApplicationResultArtifactListResult(
  config: ApplicationResultArtifactConfig,
): ApplicationResultArtifactListResult {
  return config.mode === "offline"
    ? failedList("application_session_http_disabled", "offline")
    : {
        status: "ready",
        items: [],
        nextCursor: "",
        failureCode: "",
        requestId: "",
        auditRef: "",
        summary: "选择 Session 后读取显式保存的 active 结果资产。",
      };
}

export async function listApplicationResultArtifacts(
  config: ApplicationResultArtifactConfig,
  input: {
    applicationId: string;
    sessionId: string;
    lifecycleState?: ApplicationResultArtifactLifecycleState;
    limit?: number;
    cursor?: string;
  },
  signal?: AbortSignal,
): Promise<ApplicationResultArtifactListResult> {
  if (config.mode === "offline") return initialApplicationResultArtifactListResult(config);
  const lifecycleState = input.lifecycleState ?? "active";
  const limit = input.limit ?? 50;
  if (!validScope(config, input.applicationId, input.sessionId) || !validLifecycleState(lifecycleState) ||
    !Number.isInteger(limit) || limit < 1 || limit > 100 || (input.cursor?.length ?? 0) > 4096) {
    return failedList("application_result_artifact_payload_invalid");
  }
  const requestId = createRequestId("application-result-artifact-list");
  const query = new URLSearchParams({
    workspace_id: config.workspaceId,
    application_id: input.applicationId,
    lifecycle_state: lifecycleState,
    limit: String(limit),
  });
  if (input.cursor) query.set("cursor", input.cursor);
  try {
    const response = await fetch(
      `${config.baseUrl}/v1/user-workspace/application-sessions/${encodeURIComponent(input.sessionId)}/result-artifacts?${query}`,
      { headers: artifactHeaders(config, input.applicationId, requestId, false), signal },
    );
    const value: unknown = await response.json();
    if (!isListEnvelope(value, config, input.applicationId, input.sessionId)) {
      return failedList("application_result_artifact_store_contract_mismatch", requestId);
    }
    const failureCode = nullableString(value.failure_code);
    if (!response.ok || failureCode) {
      return failedList(failureCode || "application_result_artifact_store_unavailable", String(value.request_id), String(value.audit_ref));
    }
    const items = value.items.map((item) =>
      parseApplicationResultArtifactSummary(item, config, input.applicationId, input.sessionId)
    );
    if (items.some((item) => item === null) || items.some((item) => item?.lifecycleState !== lifecycleState)) {
      return failedList("application_result_artifact_store_contract_mismatch", String(value.request_id), String(value.audit_ref));
    }
    return {
      status: "ready",
      items: items as ApplicationResultArtifactSummary[],
      nextCursor: nullableString(value.next_cursor),
      failureCode: "",
      requestId: String(value.request_id),
      auditRef: String(value.audit_ref),
      summary: `已读取 ${items.length} 条 ${lifecycleState} 结果资产元数据；正文仍未读取。`,
    };
  } catch (error) {
    return failedList(isAbort(error) ? "application_result_artifact_request_canceled" : "application_result_artifact_store_unavailable", requestId);
  }
}

export async function listApplicationResultArtifactsByApplication(
  config: ApplicationResultArtifactConfig,
  input: {
    applicationId: string;
    lifecycleState?: ApplicationResultArtifactLifecycleState;
    executionProfile?: ApplicationResultArtifactExecutionProfile | "";
    contentType?: ApplicationResultArtifactContentType | "";
    limit?: number;
    cursor?: string;
  },
  signal?: AbortSignal,
): Promise<ApplicationResultArtifactListResult> {
  if (config.mode === "offline") return failedList("application_session_http_disabled", "offline");
  const lifecycleState = input.lifecycleState ?? "active";
  const executionProfile = input.executionProfile ?? "";
  const contentType = input.contentType ?? "";
  const limit = input.limit ?? 50;
  if (!validApplicationScope(config, input.applicationId) || !validLifecycleState(lifecycleState) ||
    !validOptionalExecutionProfile(executionProfile) || !validOptionalContentType(contentType) ||
    !Number.isInteger(limit) || limit < 1 || limit > 100 || (input.cursor?.length ?? 0) > 4096) {
    return failedList("application_result_artifact_payload_invalid");
  }
  const requestId = createRequestId("application-result-artifact-application-list");
  const query = new URLSearchParams({
    workspace_id: config.workspaceId,
    lifecycle_state: lifecycleState,
    limit: String(limit),
  });
  if (executionProfile) query.set("execution_profile", executionProfile);
  if (contentType) query.set("content_type", contentType);
  if (input.cursor) query.set("cursor", input.cursor);
  try {
    const response = await fetch(
      `${config.baseUrl}/v1/user-workspace/applications/${encodeURIComponent(input.applicationId)}/result-artifacts?${query}`,
      { headers: artifactHeaders(config, input.applicationId, requestId, false), signal },
    );
    const value: unknown = await response.json();
    if (!isApplicationListEnvelope(value, config, input.applicationId)) {
      return failedList("application_result_artifact_store_contract_mismatch", requestId);
    }
    const failureCode = nullableString(value.failure_code);
    if (!response.ok || failureCode) {
      return failedList(failureCode || "application_result_artifact_store_unavailable", String(value.request_id), String(value.audit_ref));
    }
    const items = value.items.map((item) =>
      parseApplicationResultArtifactSummary(item, config, input.applicationId, "")
    );
    if (items.some((item) => item === null) || items.some((item) =>
      item?.lifecycleState !== lifecycleState || executionProfile && item.executionProfile !== executionProfile ||
      contentType && item.contentType !== contentType
    )) {
      return failedList("application_result_artifact_store_contract_mismatch", String(value.request_id), String(value.audit_ref));
    }
    return {
      status: "ready",
      items: items as ApplicationResultArtifactSummary[],
      nextCursor: nullableString(value.next_cursor),
      failureCode: "",
      requestId: String(value.request_id),
      auditRef: String(value.audit_ref),
      summary: `已读取当前 Application 的 ${items.length} 条 ${lifecycleState} 结果资产元数据；正文仍未读取。`,
    };
  } catch (error) {
    return failedList(isAbort(error) ? "application_result_artifact_request_canceled" : "application_result_artifact_store_unavailable", requestId);
  }
}

export async function readApplicationResultArtifact(
  config: ApplicationResultArtifactConfig,
  input: { applicationId: string; sessionId: string; artifactId: string },
  signal?: AbortSignal,
): Promise<ApplicationResultArtifactReadResult> {
  if (config.mode === "offline") return failedRead("application_session_http_disabled", "offline");
  if (!validScope(config, input.applicationId, input.sessionId) || !ARTIFACT_ID_PATTERN.test(input.artifactId)) {
    return failedRead("application_result_artifact_payload_invalid");
  }
  const requestId = createRequestId("application-result-artifact-read");
  const query = new URLSearchParams({ workspace_id: config.workspaceId, application_id: input.applicationId });
  try {
    const response = await fetch(
      `${config.baseUrl}/v1/user-workspace/application-sessions/${encodeURIComponent(input.sessionId)}/result-artifacts/${encodeURIComponent(input.artifactId)}?${query}`,
      { headers: artifactHeaders(config, input.applicationId, requestId, false), signal },
    );
    const value: unknown = await response.json();
    if (!isReadEnvelope(value, config, input.applicationId, input.sessionId)) {
      return failedRead("application_result_artifact_store_contract_mismatch", requestId);
    }
    const failureCode = nullableString(value.failure_code);
    const artifact = value.artifact === null ? null : parseApplicationResultArtifact(value.artifact, config, input.applicationId, input.sessionId);
    const lifecycle = value.lifecycle === null ? null : parseApplicationResultArtifactLifecycle(value.lifecycle, config, input.applicationId, input.artifactId);
    if (!response.ok || failureCode) {
      return failedRead(failureCode || "application_result_artifact_store_unavailable", String(value.request_id), String(value.audit_ref));
    }
    if (!artifact || !lifecycle || artifact.artifactId !== input.artifactId || lifecycle.artifactId !== artifact.artifactId) {
      return failedRead("application_result_artifact_store_contract_mismatch", String(value.request_id), String(value.audit_ref));
    }
    return {
      status: "ready",
      artifact,
      lifecycle,
      failureCode: "",
      requestId: String(value.request_id),
      auditRef: String(value.audit_ref),
      summary: `已精确读取结果资产 ${artifact.artifactId}；正文仅保留在当前组件内存。`,
    };
  } catch (error) {
    return failedRead(isAbort(error) ? "application_result_artifact_request_canceled" : "application_result_artifact_store_unavailable", requestId);
  }
}

export async function exportApplicationResultArtifact(
  config: ApplicationResultArtifactConfig,
  input: { applicationId: string; artifactId: string },
  signal?: AbortSignal,
): Promise<ApplicationResultArtifactExportResult> {
  if (config.mode === "offline") return failedExport("application_session_http_disabled", "offline");
  if (!validApplicationScope(config, input.applicationId) || !ARTIFACT_ID_PATTERN.test(input.artifactId)) {
    return failedExport("application_result_artifact_payload_invalid");
  }
  const requestId = createRequestId("application-result-artifact-export");
  const query = new URLSearchParams({ workspace_id: config.workspaceId });
  try {
    const response = await fetch(
      `${config.baseUrl}/v1/user-workspace/applications/${encodeURIComponent(input.applicationId)}/result-artifacts/${encodeURIComponent(input.artifactId)}/export?${query}`,
      { headers: artifactExportHeaders(config, input.applicationId, requestId), signal },
    );
    const value: unknown = await response.json();
    if (!isExportEnvelope(value, config, input.applicationId)) {
      return failedExport("application_result_artifact_store_contract_mismatch", requestId);
    }
    const failureCode = nullableString(value.failure_code);
    if (!response.ok || failureCode) {
      return failedExport(failureCode || "application_result_artifact_store_unavailable", String(value.request_id), String(value.audit_ref));
    }
    const exportDocument = value.export === null
      ? null
      : await parseApplicationResultArtifactExport(value.export, config, input.applicationId);
    if (!exportDocument || exportDocument.artifact.artifactId !== input.artifactId ||
      exportDocument.requestId !== value.request_id || exportDocument.auditRef !== value.audit_ref) {
      return failedExport("application_result_artifact_store_contract_mismatch", String(value.request_id), String(value.audit_ref));
    }
    return {
      status: "ready",
      exportDocument,
      failureCode: "",
      requestId: String(value.request_id),
      auditRef: String(value.audit_ref),
      summary: `结果资产 ${input.artifactId} 已通过 content 与 export digest 校验，可显式下载。`,
    };
  } catch (error) {
    return failedExport(isAbort(error) ? "application_result_artifact_request_canceled" : "application_result_artifact_store_unavailable", requestId);
  }
}

export async function transitionApplicationResultArtifactLifecycle(
  config: ApplicationResultArtifactConfig,
  input: {
    applicationId: string;
    sessionId: string;
    artifactId: string;
    expectedLifecycleVersion: number;
    targetState: ApplicationResultArtifactLifecycleState;
  },
  signal?: AbortSignal,
): Promise<ApplicationResultArtifactTransitionResult> {
  if (config.mode === "offline") return failedTransition("application_session_http_disabled", "offline");
  if (!validScope(config, input.applicationId, input.sessionId) || !ARTIFACT_ID_PATTERN.test(input.artifactId) ||
    !Number.isInteger(input.expectedLifecycleVersion) || input.expectedLifecycleVersion < 1 || !validLifecycleState(input.targetState)) {
    return failedTransition("application_result_artifact_payload_invalid");
  }
  const requestId = createRequestId(`application-result-artifact-${input.targetState}`);
  const action = input.targetState === "archived" ? "archive" : "unarchive";
  try {
    const response = await fetch(
      `${config.baseUrl}/v1/user-workspace/application-sessions/${encodeURIComponent(input.sessionId)}/result-artifacts/${encodeURIComponent(input.artifactId)}/${action}`,
      {
        method: "POST",
        headers: { ...artifactHeaders(config, input.applicationId, requestId, true), "Content-Type": "application/json" },
        body: JSON.stringify({
          workspace_id: config.workspaceId,
          application_id: input.applicationId,
          expected_lifecycle_version: input.expectedLifecycleVersion,
        }),
        signal,
      },
    );
    const value: unknown = await response.json();
    if (!isTransitionEnvelope(value, config, input.applicationId, input.sessionId)) {
      return failedTransition("application_result_artifact_store_contract_mismatch", requestId);
    }
    const failureCode = nullableString(value.failure_code);
    const currentState = validLifecycleState(value.current_lifecycle_state) ? value.current_lifecycle_state : "";
    const currentVersion = integer(value.current_lifecycle_version, 0) ? value.current_lifecycle_version : 0;
    if (!response.ok || failureCode) {
      return {
        ...failedTransition(failureCode || "application_result_artifact_store_unavailable", String(value.request_id), String(value.audit_ref)),
        status: failureCode === "application_result_artifact_lifecycle_version_conflict"
          ? "version_conflict"
          : failureCode === "application_result_artifact_lifecycle_state_conflict" ? "state_conflict" : "failed",
        currentLifecycleVersion: currentVersion,
        currentLifecycleState: currentState,
      };
    }
    const lifecycle = value.lifecycle === null ? null : parseApplicationResultArtifactLifecycle(value.lifecycle, config, input.applicationId, input.artifactId);
    const event = value.event === null ? null : parseApplicationResultArtifactLifecycleEvent(value.event, config, input.applicationId, input.artifactId);
    if (!lifecycle || !event || lifecycle.lifecycleState !== input.targetState ||
      lifecycle.lifecycleVersion !== input.expectedLifecycleVersion + 1 || event.toState !== input.targetState ||
      event.lifecycleVersion !== lifecycle.lifecycleVersion || currentVersion !== lifecycle.lifecycleVersion || currentState !== lifecycle.lifecycleState) {
      return failedTransition("application_result_artifact_store_contract_mismatch", String(value.request_id), String(value.audit_ref));
    }
    return {
      status: "ready",
      lifecycle,
      event,
      failureCode: "",
      currentLifecycleVersion: lifecycle.lifecycleVersion,
      currentLifecycleState: lifecycle.lifecycleState,
      requestId: String(value.request_id),
      auditRef: String(value.audit_ref),
      summary: input.targetState === "archived" ? "结果资产已归档；正文和来源保持不变。" : "结果资产已恢复为 active。",
    };
  } catch (error) {
    return failedTransition(isAbort(error) ? "application_result_artifact_request_canceled" : "application_result_artifact_store_unavailable", requestId);
  }
}

export function applicationResultArtifactResponseMatchesScope(
  expected: ApplicationResultArtifactRequestScope,
  observed: ApplicationResultArtifactRequestScope,
): boolean {
  return expected.generation === observed.generation && expected.applicationId === observed.applicationId &&
    expected.sessionId === observed.sessionId && expected.lifecycleState === observed.lifecycleState &&
    expected.artifactId === observed.artifactId;
}

export function applicationResultArtifactLibraryResponseMatchesScope(
  expected: ApplicationResultArtifactLibraryRequestScope,
  observed: ApplicationResultArtifactLibraryRequestScope,
): boolean {
  return expected.generation === observed.generation && expected.applicationId === observed.applicationId &&
    expected.lifecycleState === observed.lifecycleState && expected.executionProfile === observed.executionProfile &&
    expected.contentType === observed.contentType && expected.cursor === observed.cursor &&
    expected.sessionId === observed.sessionId && expected.artifactId === observed.artifactId;
}

export function applicationResultArtifactExportFilename(exportDocument: ApplicationResultArtifactExport): string {
  return `radishmind-${exportDocument.artifact.artifactId}-lifecycle-v${exportDocument.lifecycle.lifecycleVersion}.json`;
}

export function serializeApplicationResultArtifactExport(exportDocument: ApplicationResultArtifactExport): string {
  return `${JSON.stringify(applicationResultArtifactExportWireDocument(exportDocument), null, 2)}\n`;
}

export function parseApplicationResultArtifactSummary(
  value: unknown,
  config: ApplicationResultArtifactConfig,
  applicationId: string,
  sessionId: string,
): ApplicationResultArtifactSummary | null {
  const keys = [
    "schema_version", "artifact_id", "record_version", "tenant_ref", "workspace_id", "application_id",
    "owner_subject_ref", "session_id", "turn_id", "client_turn_key", "execution_profile", "run_ref",
    "content_type", "content_bytes", "content_digest", "created_at", "lifecycle_state", "lifecycle_version",
    "archived_at", "lifecycle_updated_at",
  ];
  if (!isExactDocument(value, keys) || value.schema_version !== SUMMARY_SCHEMA_VERSION || value.record_version !== 1 ||
    !summaryScopeMatches(value, config, applicationId, sessionId) || !ARTIFACT_ID_PATTERN.test(String(value.artifact_id)) ||
    !TURN_ID_PATTERN.test(String(value.turn_id)) || !REF_PATTERN.test(String(value.client_turn_key)) ||
    !validContentType(value.content_type) || !integer(value.content_bytes, 1) || value.content_bytes > 65536 ||
    !DIGEST_PATTERN.test(String(value.content_digest)) || !isTimestamp(value.created_at) ||
    !validLifecycleState(value.lifecycle_state) || !integer(value.lifecycle_version, 1) ||
    !validArchivedAt(value.lifecycle_state, value.archived_at) || !isTimestamp(value.lifecycle_updated_at)) return null;
  const runRef = parseRunRef(value.run_ref, value.execution_profile);
  if (!runRef) return null;
  return {
    schemaVersion: SUMMARY_SCHEMA_VERSION,
    artifactId: String(value.artifact_id),
    recordVersion: 1,
    tenantRef: String(value.tenant_ref),
    workspaceId: String(value.workspace_id),
    applicationId: String(value.application_id),
    ownerSubjectRef: String(value.owner_subject_ref),
    sessionId: String(value.session_id),
    turnId: String(value.turn_id),
    clientTurnKey: String(value.client_turn_key),
    executionProfile: String(value.execution_profile),
    runRef,
    contentType: value.content_type,
    contentBytes: value.content_bytes,
    contentDigest: String(value.content_digest),
    createdAt: String(value.created_at),
    lifecycleState: value.lifecycle_state,
    lifecycleVersion: value.lifecycle_version,
    archivedAt: value.archived_at === null ? null : String(value.archived_at),
    lifecycleUpdatedAt: String(value.lifecycle_updated_at),
  };
}

function parseApplicationResultArtifact(
  value: unknown,
  config: ApplicationResultArtifactConfig,
  applicationId: string,
  sessionId: string,
): ApplicationResultArtifact | null {
  const keys = [
    "schema_version", "artifact_id", "record_version", "tenant_ref", "workspace_id", "application_id",
    "owner_subject_ref", "session_id", "turn_id", "client_turn_key", "execution_profile", "run_ref",
    "content_type", "content", "content_bytes", "content_digest", "created_at", "created_by_actor_ref",
    "request_id", "audit_ref",
  ];
  if (!isExactDocument(value, keys) || value.schema_version !== ARTIFACT_SCHEMA_VERSION || value.record_version !== 1 ||
    !summaryScopeMatches(value, config, applicationId, sessionId) || !ARTIFACT_ID_PATTERN.test(String(value.artifact_id)) ||
    !TURN_ID_PATTERN.test(String(value.turn_id)) || !REF_PATTERN.test(String(value.client_turn_key)) ||
    !validContentType(value.content_type) || typeof value.content !== "string" || value.content.trim() === "" ||
    !integer(value.content_bytes, 1) || value.content_bytes > 65536 || new TextEncoder().encode(value.content).length !== value.content_bytes ||
    !DIGEST_PATTERN.test(String(value.content_digest)) || !isTimestamp(value.created_at) ||
    !REF_PATTERN.test(String(value.created_by_actor_ref)) || !REF_PATTERN.test(String(value.request_id)) ||
    !REF_PATTERN.test(String(value.audit_ref))) return null;
  if (value.content_type === "application/json") {
    try {
      JSON.parse(value.content);
    } catch {
      return null;
    }
  }
  const runRef = parseRunRef(value.run_ref, value.execution_profile);
  if (!runRef) return null;
  return {
    schemaVersion: ARTIFACT_SCHEMA_VERSION,
    artifactId: String(value.artifact_id),
    recordVersion: 1,
    tenantRef: String(value.tenant_ref),
    workspaceId: String(value.workspace_id),
    applicationId: String(value.application_id),
    ownerSubjectRef: String(value.owner_subject_ref),
    sessionId: String(value.session_id),
    turnId: String(value.turn_id),
    clientTurnKey: String(value.client_turn_key),
    executionProfile: String(value.execution_profile),
    runRef,
    contentType: value.content_type,
    content: value.content,
    contentBytes: value.content_bytes,
    contentDigest: String(value.content_digest),
    createdAt: String(value.created_at),
    createdByActorRef: String(value.created_by_actor_ref),
    requestId: String(value.request_id),
    auditRef: String(value.audit_ref),
  };
}

function parseApplicationResultArtifactLifecycle(
  value: unknown,
  config: ApplicationResultArtifactConfig,
  applicationId: string,
  artifactId: string,
): ApplicationResultArtifactLifecycle | null {
  const keys = [
    "schema_version", "tenant_ref", "workspace_id", "application_id", "owner_subject_ref", "artifact_id",
    "lifecycle_state", "lifecycle_version", "archived_at", "updated_at", "updated_by_actor_ref", "request_id", "audit_ref",
  ];
  if (!isExactDocument(value, keys) || value.schema_version !== LIFECYCLE_SCHEMA_VERSION ||
    !lifecycleScopeMatches(value, config, applicationId, artifactId) || !validLifecycleState(value.lifecycle_state) ||
    !integer(value.lifecycle_version, 1) || !validArchivedAt(value.lifecycle_state, value.archived_at) ||
    !isTimestamp(value.updated_at) || !REF_PATTERN.test(String(value.updated_by_actor_ref)) ||
    !REF_PATTERN.test(String(value.request_id)) || !REF_PATTERN.test(String(value.audit_ref))) return null;
  return {
    schemaVersion: LIFECYCLE_SCHEMA_VERSION,
    tenantRef: String(value.tenant_ref),
    workspaceId: String(value.workspace_id),
    applicationId: String(value.application_id),
    ownerSubjectRef: String(value.owner_subject_ref),
    artifactId: String(value.artifact_id),
    lifecycleState: value.lifecycle_state,
    lifecycleVersion: value.lifecycle_version,
    archivedAt: value.archived_at === null ? null : String(value.archived_at),
    updatedAt: String(value.updated_at),
    updatedByActorRef: String(value.updated_by_actor_ref),
    requestId: String(value.request_id),
    auditRef: String(value.audit_ref),
  };
}

function parseApplicationResultArtifactLifecycleEvent(
  value: unknown,
  config: ApplicationResultArtifactConfig,
  applicationId: string,
  artifactId: string,
): ApplicationResultArtifactLifecycleEvent | null {
  const keys = [
    "schema_version", "tenant_ref", "workspace_id", "application_id", "owner_subject_ref", "artifact_id",
    "lifecycle_version", "from_state", "to_state", "transition_kind", "occurred_at", "actor_ref", "request_id", "audit_ref",
  ];
  if (!isExactDocument(value, keys) || value.schema_version !== LIFECYCLE_EVENT_SCHEMA_VERSION ||
    !lifecycleScopeMatches(value, config, applicationId, artifactId) || !integer(value.lifecycle_version, 2) ||
    !validLifecycleState(value.from_state) || !validLifecycleState(value.to_state) || value.from_state === value.to_state ||
    (value.transition_kind !== "archived" && value.transition_kind !== "unarchived") ||
    (value.transition_kind === "archived" && (value.from_state !== "active" || value.to_state !== "archived")) ||
    (value.transition_kind === "unarchived" && (value.from_state !== "archived" || value.to_state !== "active")) ||
    !isTimestamp(value.occurred_at) || !REF_PATTERN.test(String(value.actor_ref)) ||
    !REF_PATTERN.test(String(value.request_id)) || !REF_PATTERN.test(String(value.audit_ref))) return null;
  return {
    schemaVersion: LIFECYCLE_EVENT_SCHEMA_VERSION,
    tenantRef: String(value.tenant_ref),
    workspaceId: String(value.workspace_id),
    applicationId: String(value.application_id),
    ownerSubjectRef: String(value.owner_subject_ref),
    artifactId: String(value.artifact_id),
    lifecycleVersion: value.lifecycle_version,
    fromState: value.from_state,
    toState: value.to_state,
    transitionKind: value.transition_kind,
    occurredAt: String(value.occurred_at),
    actorRef: String(value.actor_ref),
    requestId: String(value.request_id),
    auditRef: String(value.audit_ref),
  };
}

function isListEnvelope(
  value: unknown,
  config: ApplicationResultArtifactConfig,
  applicationId: string,
  sessionId: string,
): value is Document & { items: unknown[] } {
  return isExactDocument(value, [
    "request_id", "tenant_ref", "workspace_id", "application_id", "session_id", "items", "next_cursor", "failure_code", "audit_ref",
  ]) && envelopeScopeMatches(value, config, applicationId, sessionId) && Array.isArray(value.items) &&
    (value.next_cursor === null || typeof value.next_cursor === "string" && value.next_cursor.length <= 4096) &&
    (value.failure_code === null || typeof value.failure_code === "string");
}

function isApplicationListEnvelope(
  value: unknown,
  config: ApplicationResultArtifactConfig,
  applicationId: string,
): value is Document & { items: unknown[] } {
  return isExactDocument(value, [
    "request_id", "tenant_ref", "workspace_id", "application_id", "items", "next_cursor", "failure_code", "audit_ref",
  ]) && applicationEnvelopeScopeMatches(value, config, applicationId) && Array.isArray(value.items) &&
    (value.next_cursor === null || typeof value.next_cursor === "string" && value.next_cursor.length <= 4096) &&
    (value.failure_code === null || typeof value.failure_code === "string");
}

function isReadEnvelope(
  value: unknown,
  config: ApplicationResultArtifactConfig,
  applicationId: string,
  sessionId: string,
): value is Document {
  return isExactDocument(value, [
    "request_id", "tenant_ref", "workspace_id", "application_id", "session_id", "artifact", "lifecycle", "failure_code", "audit_ref",
  ]) && envelopeScopeMatches(value, config, applicationId, sessionId) &&
    (value.artifact === null || isRecord(value.artifact)) && (value.lifecycle === null || isRecord(value.lifecycle)) &&
    (value.failure_code === null || typeof value.failure_code === "string");
}

function isTransitionEnvelope(
  value: unknown,
  config: ApplicationResultArtifactConfig,
  applicationId: string,
  sessionId: string,
): value is Document {
  return isExactDocument(value, [
    "request_id", "tenant_ref", "workspace_id", "application_id", "session_id", "lifecycle", "event", "failure_code",
    "current_lifecycle_version", "current_lifecycle_state", "audit_ref",
  ]) && envelopeScopeMatches(value, config, applicationId, sessionId) &&
    (value.lifecycle === null || isRecord(value.lifecycle)) && (value.event === null || isRecord(value.event)) &&
    (value.failure_code === null || typeof value.failure_code === "string") && integer(value.current_lifecycle_version, 0) &&
    (value.current_lifecycle_state === "" || validLifecycleState(value.current_lifecycle_state));
}

function isExportEnvelope(
  value: unknown,
  config: ApplicationResultArtifactConfig,
  applicationId: string,
): value is Document {
  return isExactDocument(value, [
    "request_id", "tenant_ref", "workspace_id", "application_id", "export", "failure_code", "audit_ref",
  ]) && applicationEnvelopeScopeMatches(value, config, applicationId) &&
    (value.export === null || isRecord(value.export)) &&
    (value.failure_code === null || typeof value.failure_code === "string");
}

async function parseApplicationResultArtifactExport(
  value: unknown,
  config: ApplicationResultArtifactConfig,
  applicationId: string,
): Promise<ApplicationResultArtifactExport | null> {
  const keys = [
    "schema_version", "artifact", "lifecycle", "exported_at", "exported_by_actor_ref", "request_id", "audit_ref",
    "export_digest",
  ];
  if (!isExactDocument(value, keys) || value.schema_version !== EXPORT_SCHEMA_VERSION ||
    !isRecord(value.artifact) || !isRecord(value.lifecycle) || !isTimestamp(value.exported_at) ||
    value.exported_by_actor_ref !== config.subjectRef || !REF_PATTERN.test(String(value.request_id)) ||
    !REF_PATTERN.test(String(value.audit_ref)) || !DIGEST_PATTERN.test(String(value.export_digest))) return null;
  const artifactSessionId = String(value.artifact.session_id);
  if (!SESSION_ID_PATTERN.test(artifactSessionId)) return null;
  const artifact = parseApplicationResultArtifact(value.artifact, config, applicationId, artifactSessionId);
  const artifactId = artifact?.artifactId ?? "";
  const lifecycle = parseApplicationResultArtifactLifecycle(value.lifecycle, config, applicationId, artifactId);
  if (!artifact || !lifecycle || lifecycle.artifactId !== artifact.artifactId) return null;
  const exportDocument: ApplicationResultArtifactExport = {
    schemaVersion: EXPORT_SCHEMA_VERSION,
    artifact,
    lifecycle,
    exportedAt: String(value.exported_at),
    exportedByActorRef: String(value.exported_by_actor_ref),
    requestId: String(value.request_id),
    auditRef: String(value.audit_ref),
    exportDigest: String(value.export_digest),
  };
  const contentDigest = await sha256Text(`${artifact.contentType.trim()}\u0000${artifact.content}`);
  const exportDigest = await sha256Text(JSON.stringify(applicationResultArtifactExportWireDocument({
    ...exportDocument,
    exportDigest: "",
  })));
  if (!contentDigest || !exportDigest || `sha256:${contentDigest}` !== artifact.contentDigest ||
    `sha256:${exportDigest}` !== exportDocument.exportDigest) return null;
  return exportDocument;
}

function artifactHeaders(
  config: ApplicationResultArtifactConfig,
  applicationId: string,
  requestId: string,
  lifecycleMutation: boolean,
): Record<string, string> {
  const permissions = lifecycleMutation
    ? "application_sessions:read,application_result_artifacts:archive"
    : "application_sessions:read";
  return {
    Accept: "application/json",
    "X-Request-Id": requestId,
    "X-RadishMind-Dev-Read-Identity": `application-result-artifact-web:${config.subjectRef}`,
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Read-Scopes": permissions,
    "X-RadishMind-Dev-Workflow-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Workflow-Application": applicationId,
    "X-RadishMind-Active-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Read-Membership-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Read-Membership-Permissions": permissions,
  };
}

function artifactExportHeaders(
  config: ApplicationResultArtifactConfig,
  applicationId: string,
  requestId: string,
): Record<string, string> {
  const permissions = "application_sessions:read,application_result_artifacts:export";
  return {
    ...artifactHeaders(config, applicationId, requestId, false),
    "X-RadishMind-Dev-Read-Scopes": permissions,
    "X-RadishMind-Dev-Read-Membership-Permissions": permissions,
  };
}

function parseRunRef(value: unknown, profile: unknown): ApplicationResultArtifactRunRef | null {
  if (!isExactDocument(value, ["schema_version", "run_id"]) || !RUN_ID_PATTERN.test(String(value.run_id))) return null;
  const expected = new Map<string, ApplicationResultArtifactRunRef["schemaVersion"]>([
    ["workflow_definition_executor_v1", "workflow_run_record.v5"],
    ["workflow_definition_executor_v2", "workflow_run_record.v8"],
    ["application_rag_invocation_v1", "workflow_run_record.v4"],
    ["prompt_application_invocation_v1", "workflow_run_record.v6"],
    ["agent_copilot_suggestion_v1", "workflow_run_record.v7"],
  ]).get(String(profile));
  if (!expected || value.schema_version !== expected) return null;
  return { schemaVersion: expected, runId: String(value.run_id) };
}

function summaryScopeMatches(
  value: Document,
  config: ApplicationResultArtifactConfig,
  applicationId: string,
  sessionId: string,
): boolean {
  return value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId &&
    value.application_id === applicationId && value.owner_subject_ref === config.subjectRef &&
    SESSION_ID_PATTERN.test(String(value.session_id)) && (!sessionId || value.session_id === sessionId);
}

function lifecycleScopeMatches(
  value: Document,
  config: ApplicationResultArtifactConfig,
  applicationId: string,
  artifactId: string,
): boolean {
  return value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId &&
    value.application_id === applicationId && value.owner_subject_ref === config.subjectRef && value.artifact_id === artifactId;
}

function envelopeScopeMatches(
  value: Document,
  config: ApplicationResultArtifactConfig,
  applicationId: string,
  sessionId: string,
): boolean {
  return REF_PATTERN.test(String(value.request_id)) && value.tenant_ref === config.tenantRef &&
    value.workspace_id === config.workspaceId && value.application_id === applicationId && value.session_id === sessionId &&
    REF_PATTERN.test(String(value.audit_ref));
}

function applicationEnvelopeScopeMatches(
  value: Document,
  config: ApplicationResultArtifactConfig,
  applicationId: string,
): boolean {
  return REF_PATTERN.test(String(value.request_id)) && value.tenant_ref === config.tenantRef &&
    value.workspace_id === config.workspaceId && value.application_id === applicationId &&
    REF_PATTERN.test(String(value.audit_ref));
}

function validScope(config: ApplicationResultArtifactConfig, applicationId: string, sessionId: string): boolean {
  return validApplicationScope(config, applicationId) && SESSION_ID_PATTERN.test(sessionId);
}

function validApplicationScope(config: ApplicationResultArtifactConfig, applicationId: string): boolean {
  return APPLICATION_ID_PATTERN.test(applicationId) && REF_PATTERN.test(config.tenantRef) &&
    REF_PATTERN.test(config.workspaceId) && REF_PATTERN.test(config.subjectRef);
}

function validLifecycleState(value: unknown): value is ApplicationResultArtifactLifecycleState {
  return value === "active" || value === "archived";
}

function validContentType(value: unknown): value is ApplicationResultArtifactSummary["contentType"] {
  return value === "text/markdown" || value === "application/json";
}

function validOptionalContentType(value: unknown): value is ApplicationResultArtifactContentType | "" {
  return value === "" || validContentType(value);
}

function validOptionalExecutionProfile(value: unknown): value is ApplicationResultArtifactExecutionProfile | "" {
  return value === "" || value === "workflow_definition_executor_v1" || value === "workflow_definition_executor_v2" ||
    value === "application_rag_invocation_v1" || value === "prompt_application_invocation_v1" ||
    value === "agent_copilot_suggestion_v1";
}

function validArchivedAt(state: ApplicationResultArtifactLifecycleState, value: unknown): boolean {
  return state === "active" ? value === null : isTimestamp(value);
}

function applicationResultArtifactExportWireDocument(exportDocument: ApplicationResultArtifactExport): Document {
  const artifact = exportDocument.artifact;
  const lifecycle = exportDocument.lifecycle;
  return {
    schema_version: exportDocument.schemaVersion,
    artifact: {
      schema_version: artifact.schemaVersion,
      artifact_id: artifact.artifactId,
      record_version: artifact.recordVersion,
      tenant_ref: artifact.tenantRef,
      workspace_id: artifact.workspaceId,
      application_id: artifact.applicationId,
      owner_subject_ref: artifact.ownerSubjectRef,
      session_id: artifact.sessionId,
      turn_id: artifact.turnId,
      client_turn_key: artifact.clientTurnKey,
      execution_profile: artifact.executionProfile,
      run_ref: {
        run_id: artifact.runRef.runId,
        schema_version: artifact.runRef.schemaVersion,
      },
      content_type: artifact.contentType,
      content: artifact.content,
      content_bytes: artifact.contentBytes,
      content_digest: artifact.contentDigest,
      created_at: artifact.createdAt,
      created_by_actor_ref: artifact.createdByActorRef,
      request_id: artifact.requestId,
      audit_ref: artifact.auditRef,
    },
    lifecycle: {
      schema_version: lifecycle.schemaVersion,
      tenant_ref: lifecycle.tenantRef,
      workspace_id: lifecycle.workspaceId,
      application_id: lifecycle.applicationId,
      owner_subject_ref: lifecycle.ownerSubjectRef,
      artifact_id: lifecycle.artifactId,
      lifecycle_state: lifecycle.lifecycleState,
      lifecycle_version: lifecycle.lifecycleVersion,
      archived_at: lifecycle.archivedAt,
      updated_at: lifecycle.updatedAt,
      updated_by_actor_ref: lifecycle.updatedByActorRef,
      request_id: lifecycle.requestId,
      audit_ref: lifecycle.auditRef,
    },
    exported_at: exportDocument.exportedAt,
    exported_by_actor_ref: exportDocument.exportedByActorRef,
    request_id: exportDocument.requestId,
    audit_ref: exportDocument.auditRef,
    export_digest: exportDocument.exportDigest,
  };
}

async function sha256Text(value: string): Promise<string | null> {
  if (!globalThis.crypto?.subtle) return null;
  const digest = await globalThis.crypto.subtle.digest("SHA-256", new TextEncoder().encode(value));
  return [...new Uint8Array(digest)].map((byte) => byte.toString(16).padStart(2, "0")).join("");
}

function isExactDocument(value: unknown, keys: string[]): value is Document {
  return isRecord(value) && Object.keys(value).length === keys.length && keys.every((key) => Object.hasOwn(value, key));
}

function isRecord(value: unknown): value is Document {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function integer(value: unknown, minimum: number): value is number {
  return Number.isInteger(value) && (value as number) >= minimum;
}

function isTimestamp(value: unknown): value is string {
  return typeof value === "string" && value.includes("T") && Number.isFinite(Date.parse(value));
}

function nullableString(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function isAbort(error: unknown): boolean {
  return error instanceof DOMException && error.name === "AbortError";
}

function failedList(
  failureCode: string,
  requestId = "",
  auditRef = "",
): ApplicationResultArtifactListResult {
  return {
    status: requestId === "offline" ? "offline" : "failed",
    items: [],
    nextCursor: "",
    failureCode,
    requestId: requestId === "offline" ? "" : requestId,
    auditRef,
    summary: `结果资产列表不可用：${failureCode}。`,
  };
}

function failedRead(
  failureCode: string,
  requestId = "",
  auditRef = "",
): ApplicationResultArtifactReadResult {
  return {
    status: requestId === "offline" ? "offline" : "failed",
    artifact: null,
    lifecycle: null,
    failureCode,
    requestId: requestId === "offline" ? "" : requestId,
    auditRef,
    summary: `结果资产读取不可用：${failureCode}。`,
  };
}

function failedExport(
  failureCode: string,
  requestId = "",
  auditRef = "",
): ApplicationResultArtifactExportResult {
  return {
    status: requestId === "offline" ? "offline" : "failed",
    exportDocument: null,
    failureCode,
    requestId: requestId === "offline" ? "" : requestId,
    auditRef,
    summary: `结果资产导出不可用：${failureCode}。`,
  };
}

function failedTransition(
  failureCode: string,
  requestId = "",
  auditRef = "",
): ApplicationResultArtifactTransitionResult {
  return {
    status: requestId === "offline" ? "offline" : "failed",
    lifecycle: null,
    event: null,
    failureCode,
    currentLifecycleVersion: 0,
    currentLifecycleState: "",
    requestId: requestId === "offline" ? "" : requestId,
    auditRef,
    summary: `结果资产生命周期更新不可用：${failureCode}。`,
  };
}

function createRequestId(prefix: string): string {
  return `${prefix}-${Date.now()}-${(globalThis.crypto?.randomUUID?.() ?? Math.random().toString(16).slice(2)).replaceAll("-", "").slice(0, 12)}`;
}
