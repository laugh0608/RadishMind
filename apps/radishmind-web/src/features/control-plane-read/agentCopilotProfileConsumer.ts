const DEV_SOURCE = "dev-agent-copilot-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const PROFILE_PATH = "/v1/user-workspace/agent-copilot-profiles";
const PROFILE_SCHEMA = "agent_copilot_profile_draft.v1";
const PROFILE_ID = /^acpf_[a-z2-7]{16}$/u;
const DIGEST = /^sha256:[a-f0-9]{64}$/u;
const REFERENCE = /^[A-Za-z0-9][A-Za-z0-9._:-]{2,127}$/u;
const FORBIDDEN_RESPONSE_KEYS = new Set([
  "authorization", "api_key", "credential", "endpoint", "headers", "cookie", "dsn",
  "raw_request", "raw_response", "provider_raw_response", "input", "output", "prompt", "messages",
]);

export type AgentCopilotProject = "radishflow" | "radish";
export type AgentCopilotContextPolicy = {
  allowedFields: string[];
  maxBytes: number;
  requireTaskContext: boolean;
};
export type AgentCopilotArtifactPolicy = {
  allowedKinds: string[];
  allowedRoles: string[];
  maxCount: number;
  maxItemBytes: number;
  maxTotalBytes: number;
};
export type AgentCopilotResponsePolicy = {
  allowedActionKinds: string[];
  maxAnswers: number;
  maxIssues: number;
  maxActions: number;
  maxCitations: number;
  maxVisibleTextBytes: number;
};
export type AgentCopilotRiskPolicy = {
  mode: "advisory";
  requiresConfirmationForActions: true;
  confirmationActionKinds: string[];
};
export type AgentCopilotProfileInput = {
  schemaVersion: typeof PROFILE_SCHEMA;
  profileId: string;
  workspaceId: string;
  applicationId: string;
  profileName: string;
  description: string;
  project: AgentCopilotProject;
  allowedTasks: string[];
  defaultLocale: string;
  allowedLocales: string[];
  contextPolicy: AgentCopilotContextPolicy;
  artifactPolicy: AgentCopilotArtifactPolicy;
  responsePolicy: AgentCopilotResponsePolicy;
  riskPolicy: AgentCopilotRiskPolicy;
  toolHintsPolicy: {
    allowRetrieval: false;
    allowToolCalls: false;
    allowImageReasoning: false;
  };
};
export type AgentCopilotProfileFinding = { code: string; field: string; summary: string };
export type AgentCopilotProfileValidation = {
  state: "valid" | "invalid";
  isValid: boolean;
  findings: AgentCopilotProfileFinding[];
};
export type AgentCopilotProfileDraft = AgentCopilotProfileInput & {
  draftVersion: number;
  profileDigest: string;
  policyDigest: string;
  validation: AgentCopilotProfileValidation;
  updatedAt: string;
  updatedByActorRef: string;
};
export type AgentCopilotProfileVersion = AgentCopilotProfileInput & {
  profileVersion: number;
  sourceDraftVersion: number;
  profileDigest: string;
  policyDigest: string;
  publishedAt: string;
  publishedByActorRef: string;
};
export type AgentCopilotProfileSummary = {
  profileId: string;
  profileName: string;
  project: AgentCopilotProject;
  allowedTasks: string[];
  defaultLocale: string;
  draftVersion: number;
  profileDigest: string;
  policyDigest: string;
  validationState: string;
  updatedAt: string;
};
export type AgentCopilotProfileVersionSummary = {
  profileId: string;
  profileVersion: number;
  sourceDraftVersion: number;
  profileName: string;
  project: AgentCopilotProject;
  defaultLocale: string;
  profileDigest: string;
  policyDigest: string;
  publishedAt: string;
};
export type AgentCopilotProfileConfig = {
  mode: "offline" | "dev_agent_copilot_http";
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  subjectRef: string;
};
export type AgentCopilotProfileOperation = {
  status: "offline" | "idle" | "valid" | "invalid" | "saved" | "restored" | "versioned" | "version_conflict" | "failed";
  draft: AgentCopilotProfileDraft | null;
  version: AgentCopilotProfileVersion | null;
  validation: AgentCopilotProfileValidation;
  currentDraftVersion: number;
  currentProfileVersion: number;
  failureCode: string;
  summary: string;
};
export type AgentCopilotProfileList = {
  status: "offline" | "ready" | "empty" | "failed";
  summaries: AgentCopilotProfileSummary[];
  failureCode: string;
  summary: string;
};
export type AgentCopilotProfileVersionList = {
  status: "offline" | "ready" | "empty" | "failed";
  summaries: AgentCopilotProfileVersionSummary[];
  failureCode: string;
  summary: string;
};

type Document = Record<string, any>;

export const AGENT_COPILOT_TASKS: Record<AgentCopilotProject, readonly string[]> = {
  radishflow: [
    "explain_diagnostics",
    "suggest_flowsheet_edits",
    "suggest_ghost_completion",
    "summarize_selection",
    "explain_control_plane_state",
    "inspect_canvas_snapshot",
  ],
  radish: [
    "answer_docs_question",
    "summarize_doc_or_thread",
    "suggest_forum_metadata",
    "explain_console_capability",
    "interpret_attachment",
  ],
};

export function readAgentCopilotProfileConfig(): AgentCopilotProfileConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  return {
    mode: env.VITE_RADISHMIND_AGENT_COPILOT_SOURCE?.trim() === DEV_SOURCE
      ? "dev_agent_copilot_http"
      : "offline",
    baseUrl: normalizeBaseUrl(
      env.VITE_RADISHMIND_AGENT_COPILOT_BASE_URL ??
      env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ??
      DEFAULT_BASE_URL,
    ),
    tenantRef: env.VITE_RADISHMIND_DEV_READ_TENANT_REF?.trim() || "tenant_demo",
    workspaceId: env.VITE_RADISHMIND_AGENT_COPILOT_WORKSPACE_ID?.trim() || "workspace_demo",
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || "subject_demo_user",
  };
}

export function createAgentCopilotProfileInput(
  config: AgentCopilotProfileConfig,
  applicationId: string,
): AgentCopilotProfileInput {
  return {
    schemaVersion: PROFILE_SCHEMA,
    profileId: createProfileId(),
    workspaceId: config.workspaceId,
    applicationId,
    profileName: "RadishFlow diagnostics advisor",
    description: "Review diagnostics and return advisory candidate actions.",
    project: "radishflow",
    allowedTasks: ["explain_diagnostics", "suggest_flowsheet_edits"],
    defaultLocale: "zh-CN",
    allowedLocales: ["en-US", "zh-CN"],
    contextPolicy: {
      allowedFields: ["selected_unit_ids", "diagnostics"],
      maxBytes: 65536,
      requireTaskContext: true,
    },
    artifactPolicy: {
      allowedKinds: ["json", "text"],
      allowedRoles: ["primary", "supporting"],
      maxCount: 8,
      maxItemBytes: 65536,
      maxTotalBytes: 131072,
    },
    responsePolicy: {
      allowedActionKinds: ["candidate_edit", "read_only_check"],
      maxAnswers: 8,
      maxIssues: 16,
      maxActions: 8,
      maxCitations: 16,
      maxVisibleTextBytes: 8192,
    },
    riskPolicy: {
      mode: "advisory",
      requiresConfirmationForActions: true,
      confirmationActionKinds: ["candidate_edit", "candidate_operation", "ghost_completion"],
    },
    toolHintsPolicy: {
      allowRetrieval: false,
      allowToolCalls: false,
      allowImageReasoning: false,
    },
  };
}

export function validateAgentCopilotProfileLocally(
  input: AgentCopilotProfileInput,
): AgentCopilotProfileValidation {
  const findings: AgentCopilotProfileFinding[] = [];
  const add = (code: string, field: string, summary: string) => findings.push({ code, field, summary });
  if (!PROFILE_ID.test(input.profileId) || !REFERENCE.test(input.workspaceId) || !/^app_[a-z2-7]{16}$/u.test(input.applicationId)) {
    add("agent_copilot_profile_payload_invalid", "scope", "Profile、workspace 或 application 标识不符合契约。");
  }
  if (input.profileName.trim().length < 2 || input.profileName.trim().length > 80) {
    add("agent_copilot_profile_payload_invalid", "profile_name", "Profile 名称必须为 2 至 80 个字符。");
  }
  if (input.description.trim().length > 512) {
    add("agent_copilot_profile_payload_invalid", "description", "Profile 描述不得超过 512 个字符。");
  }
  const canonicalTasks = AGENT_COPILOT_TASKS[input.project] ?? [];
  if (!input.allowedTasks.length || new Set(input.allowedTasks).size !== input.allowedTasks.length ||
      input.allowedTasks.some((task) => !canonicalTasks.includes(task))) {
    add("agent_copilot_profile_project_task_invalid", "allowed_tasks", "任务必须来自所选 project 的 canonical task 集合且不能重复。");
  }
  if (!input.allowedLocales.includes(input.defaultLocale) || new Set(input.allowedLocales).size !== input.allowedLocales.length) {
    add("agent_copilot_profile_policy_invalid", "allowed_locales", "允许语言必须唯一并包含默认语言。");
  }
  if (input.riskPolicy.mode !== "advisory" || !input.riskPolicy.requiresConfirmationForActions ||
      input.toolHintsPolicy.allowRetrieval || input.toolHintsPolicy.allowToolCalls ||
      input.toolHintsPolicy.allowImageReasoning) {
    add("agent_copilot_profile_policy_invalid", "safety", "首版必须保持 advisory、候选动作确认且全部 tool hint 为 false。");
  }
  if (containsSecretMaterial(JSON.stringify(input))) {
    add("agent_copilot_profile_secret_material_forbidden", "profile", "Profile 包含凭据、endpoint、DSN 或运行配置样式材料。");
  }
  return { state: findings.length ? "invalid" : "valid", isValid: findings.length === 0, findings };
}

export async function validateAgentCopilotProfileRemote(
  config: AgentCopilotProfileConfig,
  input: AgentCopilotProfileInput,
): Promise<AgentCopilotProfileOperation> {
  return writeProfile(config, input.applicationId, `${PROFILE_PATH}/validate`, {
    profile: profilePayload(input),
  }, "valid", ["agent_copilot_profiles:write"]);
}

export async function saveAgentCopilotProfile(
  config: AgentCopilotProfileConfig,
  input: AgentCopilotProfileInput,
  expectedDraftVersion: number,
): Promise<AgentCopilotProfileOperation> {
  return writeProfile(config, input.applicationId, PROFILE_PATH, {
    expected_draft_version: expectedDraftVersion,
    profile: profilePayload(input),
  }, "saved", ["agent_copilot_profiles:write"]);
}

export async function readAgentCopilotProfile(
  config: AgentCopilotProfileConfig,
  applicationId: string,
  profileId: string,
): Promise<AgentCopilotProfileOperation> {
  if (config.mode === "offline") return offlineOperation();
  return readProfile(
    config,
    applicationId,
    `${PROFILE_PATH}/${encodeURIComponent(profileId)}?${profileQuery(config, applicationId)}`,
    "restored",
    ["agent_copilot_profiles:read_source"],
  );
}

export async function listAgentCopilotProfiles(
  config: AgentCopilotProfileConfig,
  applicationId: string,
): Promise<AgentCopilotProfileList> {
  if (config.mode === "offline") return offlineList();
  try {
    const value = await fetchDocument(
      `${config.baseUrl}${PROFILE_PATH}?${profileQuery(config, applicationId)}`,
      { headers: profileHeaders(config, applicationId, ["agent_copilot_profiles:read"], "list") },
    );
    if (!isEnvelope(value, config, applicationId) || !Array.isArray(value.draft_summaries)) return failedList();
    if (value.failure_code) return failedList(String(value.failure_code));
    const summaries = value.draft_summaries.map(mapSummary);
    return {
      status: summaries.length ? "ready" : "empty",
      summaries,
      failureCode: "",
      summary: summaries.length ? `已加载 ${summaries.length} 个 Agent Profile 草案。` : "当前应用没有 Agent Profile 草案。",
    };
  } catch {
    return failedList();
  }
}

export async function createAgentCopilotProfileVersion(
  config: AgentCopilotProfileConfig,
  applicationId: string,
  profileId: string,
  sourceDraftVersion: number,
): Promise<AgentCopilotProfileOperation> {
  return writeProfile(
    config,
    applicationId,
    `${PROFILE_PATH}/${encodeURIComponent(profileId)}/versions`,
    {
      workspace_id: config.workspaceId,
      application_id: applicationId,
      source_draft_version: sourceDraftVersion,
    },
    "versioned",
    ["agent_copilot_profiles:version"],
  );
}

export async function listAgentCopilotProfileVersions(
  config: AgentCopilotProfileConfig,
  applicationId: string,
  profileId: string,
): Promise<AgentCopilotProfileVersionList> {
  if (config.mode === "offline") return offlineVersionList();
  try {
    const value = await fetchDocument(
      `${config.baseUrl}${PROFILE_PATH}/${encodeURIComponent(profileId)}/versions?${profileQuery(config, applicationId)}`,
      { headers: profileHeaders(config, applicationId, ["agent_copilot_profiles:read"], "version-list") },
    );
    if (!isEnvelope(value, config, applicationId) || value.profile_id !== profileId ||
        !Array.isArray(value.version_summaries)) return failedVersionList();
    if (value.failure_code) return failedVersionList(String(value.failure_code));
    const summaries = value.version_summaries.map(mapVersionSummary);
    return {
      status: summaries.length ? "ready" : "empty",
      summaries,
      failureCode: "",
      summary: summaries.length ? `已加载 ${summaries.length} 个不可变 Profile 版本。` : "当前 Profile 没有不可变版本。",
    };
  } catch {
    return failedVersionList();
  }
}

export async function readAgentCopilotProfileVersion(
  config: AgentCopilotProfileConfig,
  applicationId: string,
  profileId: string,
  profileVersion: number,
): Promise<AgentCopilotProfileOperation> {
  if (config.mode === "offline" || !Number.isInteger(profileVersion) || profileVersion < 1) {
    return offlineOperation();
  }
  try {
    const value = await fetchDocument(
      `${config.baseUrl}${PROFILE_PATH}/${encodeURIComponent(profileId)}/versions/${profileVersion}?${profileQuery(config, applicationId)}`,
      { headers: profileHeaders(config, applicationId, ["agent_copilot_profiles:read_source"], "version-read") },
    );
    if (!isEnvelope(value, config, applicationId)) return failedOperation();
    return mapOperation(value, "versioned");
  } catch {
    return failedOperation();
  }
}

async function writeProfile(
  config: AgentCopilotProfileConfig,
  applicationId: string,
  path: string,
  body: unknown,
  success: "valid" | "saved" | "versioned",
  scopes: string[],
): Promise<AgentCopilotProfileOperation> {
  if (config.mode === "offline") return offlineOperation();
  try {
    const value = await fetchDocument(`${config.baseUrl}${path}`, {
      method: "POST",
      headers: {
        ...profileHeaders(config, applicationId, scopes, success),
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    });
    if (!isEnvelope(value, config, applicationId)) return failedOperation();
    return mapOperation(value, success);
  } catch {
    return failedOperation();
  }
}

async function readProfile(
  config: AgentCopilotProfileConfig,
  applicationId: string,
  path: string,
  success: "restored",
  scopes: string[],
): Promise<AgentCopilotProfileOperation> {
  try {
    const value = await fetchDocument(`${config.baseUrl}${path}`, {
      headers: profileHeaders(config, applicationId, scopes, success),
    });
    if (!isEnvelope(value, config, applicationId)) return failedOperation();
    return mapOperation(value, success);
  } catch {
    return failedOperation();
  }
}

function mapOperation(
  value: Document,
  success: AgentCopilotProfileOperation["status"],
): AgentCopilotProfileOperation {
  const failureCode = typeof value.failure_code === "string" ? value.failure_code : "";
  const validation = mapValidation(value.validation_summary);
  const status = failureCode === "agent_copilot_profile_version_conflict"
    ? "version_conflict"
    : failureCode
      ? validation.findings.length ? "invalid" : "failed"
      : success;
  return {
    status,
    draft: isProfileDocument(value.draft) ? mapDraft(value.draft) : null,
    version: isProfileDocument(value.version) ? mapVersion(value.version) : null,
    validation,
    currentDraftVersion: integer(value.current_draft_version, 0) ? value.current_draft_version : 0,
    currentProfileVersion: integer(value.current_profile_version, 0) ? value.current_profile_version : 0,
    failureCode,
    summary: failureCode
      ? `Agent Profile 操作失败：${failureCode}。`
      : success === "saved"
        ? `Agent Profile 草案 v${value.current_draft_version} 已保存。`
        : success === "versioned"
          ? `不可变 Agent Profile v${value.current_profile_version} 已创建。`
          : success === "restored"
            ? "已恢复 Agent Profile 源码。"
            : "服务端确定性校验通过。",
  };
}

function profilePayload(input: AgentCopilotProfileInput): Document {
  return {
    schema_version: input.schemaVersion,
    profile_id: input.profileId,
    workspace_id: input.workspaceId,
    application_id: input.applicationId,
    profile_name: input.profileName,
    description: input.description,
    project: input.project,
    allowed_tasks: input.allowedTasks,
    default_locale: input.defaultLocale,
    allowed_locales: input.allowedLocales,
    context_policy: {
      allowed_fields: input.contextPolicy.allowedFields,
      max_bytes: input.contextPolicy.maxBytes,
      require_task_context: input.contextPolicy.requireTaskContext,
    },
    artifact_policy: {
      allowed_kinds: input.artifactPolicy.allowedKinds,
      allowed_roles: input.artifactPolicy.allowedRoles,
      max_count: input.artifactPolicy.maxCount,
      max_item_bytes: input.artifactPolicy.maxItemBytes,
      max_total_bytes: input.artifactPolicy.maxTotalBytes,
    },
    response_policy: {
      allowed_action_kinds: input.responsePolicy.allowedActionKinds,
      max_answers: input.responsePolicy.maxAnswers,
      max_issues: input.responsePolicy.maxIssues,
      max_actions: input.responsePolicy.maxActions,
      max_citations: input.responsePolicy.maxCitations,
      max_visible_text_bytes: input.responsePolicy.maxVisibleTextBytes,
    },
    risk_policy: {
      mode: input.riskPolicy.mode,
      requires_confirmation_for_actions: input.riskPolicy.requiresConfirmationForActions,
      confirmation_action_kinds: input.riskPolicy.confirmationActionKinds,
    },
    tool_hints_policy: {
      allow_retrieval: input.toolHintsPolicy.allowRetrieval,
      allow_tool_calls: input.toolHintsPolicy.allowToolCalls,
      allow_image_reasoning: input.toolHintsPolicy.allowImageReasoning,
    },
  };
}

function mapProfileInput(value: Document): AgentCopilotProfileInput {
  return {
    schemaVersion: value.schema_version,
    profileId: value.profile_id,
    workspaceId: value.workspace_id,
    applicationId: value.application_id,
    profileName: value.profile_name,
    description: value.description,
    project: value.project,
    allowedTasks: [...value.allowed_tasks],
    defaultLocale: value.default_locale,
    allowedLocales: [...value.allowed_locales],
    contextPolicy: {
      allowedFields: [...value.context_policy.allowed_fields],
      maxBytes: value.context_policy.max_bytes,
      requireTaskContext: value.context_policy.require_task_context,
    },
    artifactPolicy: {
      allowedKinds: [...value.artifact_policy.allowed_kinds],
      allowedRoles: [...value.artifact_policy.allowed_roles],
      maxCount: value.artifact_policy.max_count,
      maxItemBytes: value.artifact_policy.max_item_bytes,
      maxTotalBytes: value.artifact_policy.max_total_bytes,
    },
    responsePolicy: {
      allowedActionKinds: [...value.response_policy.allowed_action_kinds],
      maxAnswers: value.response_policy.max_answers,
      maxIssues: value.response_policy.max_issues,
      maxActions: value.response_policy.max_actions,
      maxCitations: value.response_policy.max_citations,
      maxVisibleTextBytes: value.response_policy.max_visible_text_bytes,
    },
    riskPolicy: {
      mode: value.risk_policy.mode,
      requiresConfirmationForActions: value.risk_policy.requires_confirmation_for_actions,
      confirmationActionKinds: [...value.risk_policy.confirmation_action_kinds],
    },
    toolHintsPolicy: {
      allowRetrieval: value.tool_hints_policy.allow_retrieval,
      allowToolCalls: value.tool_hints_policy.allow_tool_calls,
      allowImageReasoning: value.tool_hints_policy.allow_image_reasoning,
    },
  };
}

function mapDraft(value: Document): AgentCopilotProfileDraft {
  return {
    ...mapProfileInput(value),
    draftVersion: value.draft_version,
    profileDigest: value.profile_digest,
    policyDigest: value.policy_digest,
    validation: mapValidation(value.validation_summary),
    updatedAt: value.updated_at,
    updatedByActorRef: value.updated_by_actor_ref,
  };
}

function mapVersion(value: Document): AgentCopilotProfileVersion {
  return {
    ...mapProfileInput(value),
    profileVersion: value.profile_version,
    sourceDraftVersion: value.source_draft_version,
    profileDigest: value.profile_digest,
    policyDigest: value.policy_digest,
    publishedAt: value.published_at,
    publishedByActorRef: value.published_by_actor_ref,
  };
}

function mapSummary(value: Document): AgentCopilotProfileSummary {
  return {
    profileId: String(value.profile_id),
    profileName: String(value.profile_name),
    project: value.project,
    allowedTasks: Array.isArray(value.allowed_tasks) ? value.allowed_tasks.map(String) : [],
    defaultLocale: String(value.default_locale),
    draftVersion: Number(value.draft_version),
    profileDigest: String(value.profile_digest),
    policyDigest: String(value.policy_digest),
    validationState: String(value.validation_state),
    updatedAt: String(value.updated_at),
  };
}

function mapVersionSummary(value: Document): AgentCopilotProfileVersionSummary {
  return {
    profileId: String(value.profile_id),
    profileVersion: Number(value.profile_version),
    sourceDraftVersion: Number(value.source_draft_version),
    profileName: String(value.profile_name),
    project: value.project,
    defaultLocale: String(value.default_locale),
    profileDigest: String(value.profile_digest),
    policyDigest: String(value.policy_digest),
    publishedAt: String(value.published_at),
  };
}

function mapValidation(value: unknown): AgentCopilotProfileValidation {
  if (!isRecord(value) || !Array.isArray(value.findings)) {
    return { state: "invalid", isValid: false, findings: [] };
  }
  return {
    state: value.is_valid ? "valid" : "invalid",
    isValid: value.is_valid === true,
    findings: value.findings
      .filter(isRecord)
      .map((finding) => ({
        code: String(finding.code),
        field: String(finding.field),
        summary: String(finding.summary),
      })),
  };
}

function isEnvelope(value: unknown, config: AgentCopilotProfileConfig, applicationId: string): value is Document {
  return isRecord(value) && !containsForbiddenResponse(value) &&
    value.workspace_id === config.workspaceId && value.application_id === applicationId &&
    typeof value.request_id === "string" && typeof value.audit_ref === "string" &&
    (value.failure_code === null || typeof value.failure_code === "string");
}

function isProfileDocument(value: unknown): value is Document {
  return isRecord(value) && value.schema_version &&
    PROFILE_ID.test(String(value.profile_id)) &&
    typeof value.profile_name === "string" &&
    (value.project === "radishflow" || value.project === "radish") &&
    Array.isArray(value.allowed_tasks) &&
    isRecord(value.context_policy) &&
    isRecord(value.artifact_policy) &&
    isRecord(value.response_policy) &&
    isRecord(value.risk_policy) &&
    isRecord(value.tool_hints_policy) &&
    DIGEST.test(String(value.profile_digest)) &&
    DIGEST.test(String(value.policy_digest));
}

function profileHeaders(
  config: AgentCopilotProfileConfig,
  applicationId: string,
  scopes: string[],
  operation: string,
): Record<string, string> {
  const requestId = createRequestId(`agent-profile-${operation}`);
  return {
    Accept: "application/json",
    "X-Request-Id": requestId,
    "X-RadishMind-Dev-Read-Identity": "radishmind-web-agent-copilot",
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Read-Scopes": scopes.join(","),
    "X-RadishMind-Dev-Read-Audit": `audit_${requestId}_agent_copilot`,
    "X-RadishMind-Dev-Agent-Copilot-Profile-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Agent-Copilot-Profile-Application": applicationId,
  };
}

function offlineOperation(): AgentCopilotProfileOperation {
  return {
    status: "offline",
    draft: null,
    version: null,
    validation: { state: "invalid", isValid: false, findings: [] },
    currentDraftVersion: 0,
    currentProfileVersion: 0,
    failureCode: "agent_copilot_profile_http_disabled",
    summary: "Agent Copilot Web 未启用。",
  };
}

function failedOperation(): AgentCopilotProfileOperation {
  return {
    ...offlineOperation(),
    status: "failed",
    failureCode: "agent_copilot_profile_store_unavailable",
    summary: "Agent Copilot Profile owner 不可用；未回退到离线数据。",
  };
}

function offlineList(): AgentCopilotProfileList {
  return { status: "offline", summaries: [], failureCode: "agent_copilot_profile_http_disabled", summary: "Agent Copilot Web 未启用。" };
}

function failedList(failureCode = "agent_copilot_profile_store_unavailable"): AgentCopilotProfileList {
  return { status: "failed", summaries: [], failureCode, summary: "Agent Copilot Profile 草案无法加载。" };
}

function offlineVersionList(): AgentCopilotProfileVersionList {
  return { status: "offline", summaries: [], failureCode: "agent_copilot_profile_http_disabled", summary: "Agent Copilot Web 未启用。" };
}

function failedVersionList(failureCode = "agent_copilot_profile_store_unavailable"): AgentCopilotProfileVersionList {
  return { status: "failed", summaries: [], failureCode, summary: "Agent Copilot Profile 版本无法加载。" };
}

async function fetchDocument(url: string, init: RequestInit): Promise<unknown> {
  const response = await fetch(url, init);
  const value: unknown = await response.json();
  if (!response.ok) throw new Error("Agent Copilot Profile request failed.");
  return value;
}

function profileQuery(config: AgentCopilotProfileConfig, applicationId: string): string {
  return new URLSearchParams({ workspace_id: config.workspaceId, application_id: applicationId }).toString();
}

function createProfileId(): string {
  const alphabet = "abcdefghijklmnopqrstuvwxyz234567";
  const bytes = new Uint8Array(16);
  globalThis.crypto?.getRandomValues?.(bytes);
  return `acpf_${Array.from(bytes, (value, index) => alphabet[value % 32] ?? alphabet[index % 32]).join("")}`;
}

function createRequestId(prefix: string): string {
  return `${prefix}-${Date.now()}-${(globalThis.crypto?.randomUUID?.() ?? Math.random().toString(16).slice(2)).replaceAll("-", "").slice(0, 12)}`;
}

function containsForbiddenResponse(value: unknown): boolean {
  if (Array.isArray(value)) return value.some(containsForbiddenResponse);
  if (!isRecord(value)) return false;
  return Object.entries(value).some(([key, nested]) =>
    FORBIDDEN_RESPONSE_KEYS.has(key.toLowerCase()) || containsForbiddenResponse(nested)
  );
}

function containsSecretMaterial(value: string): boolean {
  return /authorization:|bearer\s|api[_-]?key\s*[:=]|cookie:|password\s*=|secret\s*=|token\s*=|(?:postgres(?:ql)?|mysql|mongodb):\/\/|https?:\/\//iu.test(value);
}

function normalizeBaseUrl(value: string): string {
  return value.trim().replace(/\/$/u, "");
}

function integer(value: unknown, minimum: number): value is number {
  return Number.isInteger(value) && Number(value) >= minimum;
}

function isRecord(value: unknown): value is Document {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}
