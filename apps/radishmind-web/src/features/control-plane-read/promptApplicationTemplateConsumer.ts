const DEV_SOURCE = "dev-prompt-application-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const TEMPLATE_PATH = "/v1/user-workspace/prompt-application-templates";
const DRAFT_SCHEMA = "prompt_application_template_draft.v1";
const VERSION_SCHEMA = "prompt_application_template_version.v1";
const TEMPLATE_ID = /^ptpl_[a-z2-7]{16}$/u;
const APPLICATION_ID = /^app_[a-z2-7]{16}$/u;
const LOCAL_USER_ACTOR_PATTERN = /^user:usr_[a-f0-9]{32}$/u;
const REFERENCE = /^[A-Za-z0-9][A-Za-z0-9_.:/-]{0,159}$/u;
const DIGEST = /^sha256:[a-f0-9]{64}$/u;
const VARIABLE_NAME = /^[A-Za-z][A-Za-z0-9_]{0,63}$/u;
const SECRET_MATERIAL = /authorization:|bearer\s|api[_-]?key\s*[:=]|x-radishmind-dev-|cookie:|password\s*=|secret\s*=|token\s*=|sk-[a-z0-9]|(?:postgres(?:ql)?|mysql|mongodb):\/\//iu;
const FORBIDDEN_RESPONSE_KEYS = new Set([
  "authorization", "api_key", "credential", "endpoint", "headers", "cookie", "dsn",
  "raw_request", "raw_response", "provider_raw_response", "input", "output",
]);

const ENVELOPE_KEYS = [
  "request_id", "workspace_id", "application_id", "draft", "version", "failure_code",
  "current_draft_version", "current_template_version", "validation_summary", "audit_ref",
] as const;
const LIST_ENVELOPE_KEYS = [
  "request_id", "workspace_id", "application_id", "draft_summaries", "failure_code", "audit_ref",
] as const;
const VERSION_LIST_ENVELOPE_KEYS = [
  "request_id", "workspace_id", "application_id", "template_id", "version_summaries",
  "failure_code", "audit_ref",
] as const;
const DRAFT_INPUT_KEYS = [
  "schema_version", "template_id", "workspace_id", "application_id", "template_name",
  "description", "messages", "variables", "output_contract",
] as const;
const DRAFT_KEYS = [
  ...DRAFT_INPUT_KEYS, "tenant_ref", "owner_subject_ref", "draft_version", "template_digest",
  "validation_summary", "created_at", "updated_at", "created_by_actor_ref",
  "updated_by_actor_ref", "request_id", "audit_ref",
] as const;
const VERSION_KEYS = [
  "schema_version", "template_id", "template_version", "source_draft_version", "tenant_ref",
  "workspace_id", "application_id", "owner_subject_ref", "template_name", "description",
  "messages", "variables", "output_contract", "template_digest", "created_at",
  "created_by_actor_ref", "request_id", "audit_ref",
] as const;
const DRAFT_SUMMARY_KEYS = [
  "schema_version", "template_id", "application_id", "template_name", "description",
  "draft_version", "template_digest", "validation_state", "message_roles", "variable_names",
  "output_kind", "updated_at", "updated_by_actor_ref",
] as const;
const VERSION_SUMMARY_KEYS = [
  "schema_version", "template_id", "template_version", "source_draft_version", "template_name",
  "template_digest", "output_kind", "created_at", "created_by_actor_ref",
] as const;

export type PromptTemplateRole = "system" | "developer" | "user";
export type PromptTemplateVariableType = "string" | "integer" | "number" | "boolean" | "string_list";
export type PromptTemplateOutputKind = "text" | "json_object";
export type PromptTemplateMessage = { role: PromptTemplateRole; content: string };
export type PromptTemplateVariable = {
  name: string;
  type: PromptTemplateVariableType;
  required: boolean;
  description: string;
  defaultValue?: string | number | boolean | string[];
};
export type PromptTemplateJSONSchema = {
  type: "object" | "array" | "string" | "integer" | "number" | "boolean";
  properties?: Record<string, PromptTemplateJSONSchema>;
  required?: string[];
  additionalProperties: boolean;
  items?: PromptTemplateJSONSchema;
};
export type PromptTemplateOutputContract = {
  kind: PromptTemplateOutputKind;
  allowEmpty: boolean;
  maxBytes: number;
  jsonSchema?: PromptTemplateJSONSchema;
};
export type PromptTemplateDraftInput = {
  templateId: string;
  workspaceId: string;
  applicationId: string;
  templateName: string;
  description: string;
  messages: PromptTemplateMessage[];
  variables: PromptTemplateVariable[];
  outputContract: PromptTemplateOutputContract;
};
export type PromptTemplateFinding = { code: string; field: string; summary: string };
export type PromptTemplateValidation = {
  state: "valid" | "invalid";
  isValid: boolean;
  findings: PromptTemplateFinding[];
};
export type PromptTemplateDraft = PromptTemplateDraftInput & {
  tenantRef: string;
  ownerSubjectRef: string;
  draftVersion: number;
  templateDigest: string;
  validation: PromptTemplateValidation;
  createdAt: string;
  updatedAt: string;
  createdByActorRef: string;
  updatedByActorRef: string;
  requestId: string;
  auditRef: string;
};
export type PromptTemplateVersion = PromptTemplateDraftInput & {
  tenantRef: string;
  ownerSubjectRef: string;
  templateVersion: number;
  sourceDraftVersion: number;
  templateDigest: string;
  createdAt: string;
  createdByActorRef: string;
  requestId: string;
  auditRef: string;
};
export type PromptTemplateDraftSummary = {
  templateId: string;
  applicationId: string;
  templateName: string;
  description: string;
  draftVersion: number;
  templateDigest: string;
  validationState: "valid";
  messageRoles: PromptTemplateRole[];
  variableNames: string[];
  outputKind: PromptTemplateOutputKind;
  updatedAt: string;
  updatedByActorRef: string;
};
export type PromptTemplateVersionSummary = {
  templateId: string;
  templateVersion: number;
  sourceDraftVersion: number;
  templateName: string;
  templateDigest: string;
  outputKind: PromptTemplateOutputKind;
  createdAt: string;
  createdByActorRef: string;
};
export type PromptTemplateConfig = {
  mode: "offline" | "dev_prompt_application_http";
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  subjectRef: string;
  authMode?: "dev_headers" | "local_session_dev_test";
};
export type PromptTemplateOperation = {
  status: "offline" | "idle" | "valid" | "invalid" | "saved" | "restored" | "versioned" | "version_conflict" | "scope_denied" | "failed";
  draft: PromptTemplateDraft | null;
  version: PromptTemplateVersion | null;
  validation: PromptTemplateValidation;
  failureCode: string;
  currentDraftVersion: number;
  currentTemplateVersion: number;
  summary: string;
};
export type PromptTemplateListResult = {
  status: "offline" | "ready" | "empty" | "failed";
  summaries: PromptTemplateDraftSummary[];
  failureCode: string;
  summary: string;
};
export type PromptTemplateVersionListResult = {
  status: "offline" | "ready" | "empty" | "failed";
  summaries: PromptTemplateVersionSummary[];
  failureCode: string;
  summary: string;
};
export type PromptTemplatePreview = {
  status: "valid" | "invalid";
  messages: PromptTemplateMessage[];
  findings: PromptTemplateFinding[];
};

type Document = Record<string, unknown>;

export function mergePromptTemplateValidationOperation(
  current: PromptTemplateOperation,
  validation: PromptTemplateOperation,
): PromptTemplateOperation {
  return {
    ...validation,
    draft: current.draft,
    version: current.version,
    currentDraftVersion: current.currentDraftVersion,
    currentTemplateVersion: current.currentTemplateVersion,
  };
}

export function readPromptTemplateConfig(): PromptTemplateConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  return {
    mode: env.VITE_RADISHMIND_PROMPT_APPLICATION_SOURCE?.trim() === DEV_SOURCE
      ? "dev_prompt_application_http"
      : "offline",
    baseUrl: normalizeBaseUrl(
      env.VITE_RADISHMIND_PROMPT_APPLICATION_BASE_URL ??
      env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ??
      DEFAULT_BASE_URL,
    ),
    tenantRef: env.VITE_RADISHMIND_DEV_READ_TENANT_REF?.trim() || "tenant_demo",
    workspaceId: env.VITE_RADISHMIND_PROMPT_APPLICATION_WORKSPACE_ID?.trim() || "workspace_demo",
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || "subject_demo_user",
    authMode: env.VITE_RADISHMIND_READ_AUTH_MODE?.trim() === "local_session_dev_test"
      ? "local_session_dev_test"
      : "dev_headers",
  };
}

export function createPromptTemplateDraftInput(
  config: PromptTemplateConfig,
  applicationId: string,
): PromptTemplateDraftInput {
  return {
    templateId: createPromptTemplateId(),
    workspaceId: config.workspaceId,
    applicationId,
    templateName: "支持问题摘要",
    description: "按指定语气概括支持问题",
    messages: [
      { role: "system", content: "请使用 {{ tone }} 的语气。" },
      { role: "user", content: "问题：{{ question }}" },
    ],
    variables: [
      { name: "question", type: "string", required: true, description: "用户问题" },
      { name: "tone", type: "string", required: false, description: "回答语气", defaultValue: "清晰" },
    ],
    outputContract: { kind: "text", allowEmpty: false, maxBytes: 4096 },
  };
}

export function validatePromptTemplateLocally(input: PromptTemplateDraftInput): PromptTemplateValidation {
  const findings: PromptTemplateFinding[] = [];
  const add = (code: string, field: string, summary: string) => findings.push({ code, field, summary });
  if (!TEMPLATE_ID.test(input.templateId) || !REFERENCE.test(input.workspaceId) || !APPLICATION_ID.test(input.applicationId)) {
    add("prompt_template_payload_invalid", "scope", "Template、workspace 或 application 标识不符合契约。");
  }
  if (input.templateName.trim().length < 2 || input.templateName.trim().length > 80) {
    add("prompt_template_payload_invalid", "template_name", "模板名称必须为 2 至 80 个字符。");
  }
  if (input.description.trim().length > 512) {
    add("prompt_template_payload_invalid", "description", "模板描述不得超过 512 个字符。");
  }
  if (SECRET_MATERIAL.test(input.templateName) || SECRET_MATERIAL.test(input.description)) {
    add("prompt_template_secret_material_forbidden", "metadata", "模板元数据包含凭据样式材料。");
  }
  if (input.messages.length < 1 || input.messages.length > 16) {
    add("prompt_template_payload_invalid", "messages", "消息数量必须为 1 至 16。");
  }
  if (input.variables.length > 64) {
    add("prompt_template_variable_invalid", "variables", "变量数量不得超过 64。");
  }
  const variables = new Map<string, PromptTemplateVariable>();
  for (const [index, variable] of input.variables.entries()) {
    const field = `variables[${index}]`;
    if (!VARIABLE_NAME.test(variable.name) || !isVariableType(variable.type)) {
      add("prompt_template_variable_invalid", field, "变量名称或类型不符合契约。");
      continue;
    }
    if (variables.has(variable.name)) add("prompt_template_variable_invalid", `${field}.name`, "变量名称重复。");
    variables.set(variable.name, variable);
    if (variable.description.length > 512 || SECRET_MATERIAL.test(variable.description)) {
      add(SECRET_MATERIAL.test(variable.description) ? "prompt_template_secret_material_forbidden" : "prompt_template_variable_invalid", `${field}.description`, "变量描述不符合预算或敏感材料边界。");
    }
    if (variable.required && variable.defaultValue !== undefined) {
      add("prompt_template_variable_invalid", `${field}.default_value`, "必填变量不能声明默认值。");
    } else if (variable.defaultValue !== undefined && !valueMatchesType(variable.defaultValue, variable.type)) {
      add("prompt_template_variable_invalid", `${field}.default_value`, "默认值与变量类型不匹配。");
    }
  }
  const referenced = new Set<string>();
  let totalBytes = 0;
  for (const [index, message] of input.messages.entries()) {
    const field = `messages[${index}]`;
    totalBytes += utf8Bytes(message.content);
    if (!isRole(message.role) || !message.content || utf8Bytes(message.content) > 16 * 1024) {
      add("prompt_template_payload_invalid", field, "消息角色、内容或长度不符合契约。");
    }
    if (SECRET_MATERIAL.test(message.content)) {
      add("prompt_template_secret_material_forbidden", `${field}.content`, "消息包含凭据样式材料。");
    }
    const parsed = parseTemplate(message.content);
    if (!parsed.valid) {
      add("prompt_template_syntax_invalid", `${field}.content`, "消息包含不受支持的模板语法。");
      continue;
    }
    for (const name of parsed.variables) {
      referenced.add(name);
      if (!variables.has(name)) add("prompt_template_variable_invalid", `${field}.content`, `变量 ${name} 未声明。`);
    }
  }
  if (totalBytes > 64 * 1024) add("prompt_template_payload_invalid", "messages", "模板源码超过 64 KiB。");
  for (const name of variables.keys()) {
    if (!referenced.has(name)) add("prompt_template_variable_invalid", `variables.${name}`, "已声明变量未被消息引用。");
  }
  if (!isOutputContract(input.outputContract)) {
    add("prompt_template_output_contract_invalid", "output_contract", "输出契约不符合当前受限 schema。");
  }
  return { state: findings.length ? "invalid" : "valid", isValid: findings.length === 0, findings };
}

export function renderPromptTemplatePreview(
  input: PromptTemplateDraftInput,
  values: Record<string, unknown>,
): PromptTemplatePreview {
  const validation = validatePromptTemplateLocally(input);
  if (!validation.isValid) return { status: "invalid", messages: [], findings: validation.findings };
  const findings: PromptTemplateFinding[] = [];
  const canonical = new Map<string, string>();
  const declarations = new Map(input.variables.map((variable) => [variable.name, variable]));
  for (const name of Object.keys(values)) {
    if (!declarations.has(name)) {
      findings.push({ code: "prompt_template_variable_invalid", field: `input.${name}`, summary: "输入包含未声明变量。" });
    }
  }
  for (const variable of input.variables) {
    const value = Object.hasOwn(values, variable.name)
      ? values[variable.name]
      : variable.defaultValue;
    if (value === undefined) {
      if (variable.required) findings.push({ code: "prompt_template_variable_invalid", field: `input.${variable.name}`, summary: "缺少必填变量。" });
      canonical.set(variable.name, "");
      continue;
    }
    if (!valueMatchesType(value, variable.type)) {
      findings.push({ code: "prompt_template_variable_invalid", field: `input.${variable.name}`, summary: "合成值与变量类型不匹配。" });
      continue;
    }
    const rendered = canonicalVariableValue(value, variable.type);
    if (SECRET_MATERIAL.test(rendered)) {
      findings.push({ code: "prompt_template_secret_material_forbidden", field: `input.${variable.name}`, summary: "合成值包含凭据样式材料。" });
    }
    canonical.set(variable.name, rendered);
  }
  if (findings.length) return { status: "invalid", messages: [], findings };
  const messages = input.messages.map((message) => ({
    role: message.role,
    content: message.content.replace(/\{\{\s*([A-Za-z][A-Za-z0-9_]*)\s*\}\}/gu, (_match, name: string) => canonical.get(name) ?? ""),
  }));
  if (messages.reduce((total, message) => total + utf8Bytes(message.content), 0) > 128 * 1024) {
    return { status: "invalid", messages: [], findings: [{ code: "prompt_template_variable_invalid", field: "input", summary: "渲染结果超过 128 KiB。" }] };
  }
  return { status: "valid", messages, findings: [] };
}

export async function validatePromptTemplateRemote(
  config: PromptTemplateConfig,
  input: PromptTemplateDraftInput,
): Promise<PromptTemplateOperation> {
  return writeTemplate(config, input.applicationId, `${TEMPLATE_PATH}/validate`, { template: draftInputPayload(input) }, "valid");
}

export async function savePromptTemplateDraft(
  config: PromptTemplateConfig,
  input: PromptTemplateDraftInput,
  expectedDraftVersion: number,
): Promise<PromptTemplateOperation> {
  return writeTemplate(config, input.applicationId, TEMPLATE_PATH, {
    expected_draft_version: expectedDraftVersion,
    template: draftInputPayload(input),
  }, "saved");
}

export async function readPromptTemplateDraft(
  config: PromptTemplateConfig,
  applicationId: string,
  templateId: string,
): Promise<PromptTemplateOperation> {
  const query = templateQuery(config, applicationId);
  return readTemplate(config, applicationId, `${TEMPLATE_PATH}/${encodeURIComponent(templateId)}?${query}`, "restored");
}

export async function listPromptTemplateDrafts(
  config: PromptTemplateConfig,
  applicationId: string,
): Promise<PromptTemplateListResult> {
  if (config.mode === "offline") return offlineList();
  try {
    const value = await fetchDocument(`${config.baseUrl}${TEMPLATE_PATH}?${templateQuery(config, applicationId)}`, {
      ...templateRequestInit(config),
      headers: templateHeaders(config, applicationId, ["prompt_application_templates:read"], "list"),
    });
    if (!isDraftListEnvelope(value, config, applicationId)) return failedList();
    const failureCode = nullableString(value.failure_code);
    if (failureCode) return failedList(failureCode);
    const summaries = (value.draft_summaries as Document[]).map(mapDraftSummary);
    return {
      status: summaries.length ? "ready" : "empty",
      summaries,
      failureCode: "",
      summary: summaries.length ? `已加载 ${summaries.length} 个模板草案。` : "当前应用没有模板草案。",
    };
  } catch {
    return failedList();
  }
}

export async function createPromptTemplateVersion(
  config: PromptTemplateConfig,
  applicationId: string,
  templateId: string,
  sourceDraftVersion: number,
): Promise<PromptTemplateOperation> {
  return writeTemplate(config, applicationId, `${TEMPLATE_PATH}/${encodeURIComponent(templateId)}/versions`, {
    workspace_id: config.workspaceId,
    application_id: applicationId,
    source_draft_version: sourceDraftVersion,
  }, "versioned", ["prompt_application_templates:version"]);
}

export async function readPromptTemplateVersion(
  config: PromptTemplateConfig,
  applicationId: string,
  templateId: string,
  templateVersion: number,
): Promise<PromptTemplateOperation> {
  const query = templateQuery(config, applicationId);
  return readTemplate(
    config,
    applicationId,
    `${TEMPLATE_PATH}/${encodeURIComponent(templateId)}/versions/${templateVersion}?${query}`,
    "versioned",
  );
}

export async function listPromptTemplateVersions(
  config: PromptTemplateConfig,
  applicationId: string,
  templateId: string,
): Promise<PromptTemplateVersionListResult> {
  if (config.mode === "offline") return offlineVersionList();
  try {
    const value = await fetchDocument(
      `${config.baseUrl}${TEMPLATE_PATH}/${encodeURIComponent(templateId)}/versions?${templateQuery(config, applicationId)}`,
      {
        ...templateRequestInit(config),
        headers: templateHeaders(config, applicationId, ["prompt_application_templates:read"], "version-list"),
      },
    );
    if (!isVersionListEnvelope(value, config, applicationId, templateId)) return failedVersionList();
    const failureCode = nullableString(value.failure_code);
    if (failureCode) return failedVersionList(failureCode);
    const summaries = (value.version_summaries as Document[]).map(mapVersionSummary);
    return {
      status: summaries.length ? "ready" : "empty",
      summaries,
      failureCode: "",
      summary: summaries.length ? `已加载 ${summaries.length} 个不可变版本。` : "当前模板没有不可变版本。",
    };
  } catch {
    return failedVersionList();
  }
}

async function writeTemplate(
  config: PromptTemplateConfig,
  applicationId: string,
  path: string,
  body: unknown,
  success: "valid" | "saved" | "versioned",
  scopes: string[] = ["prompt_application_templates:write"],
): Promise<PromptTemplateOperation> {
  if (config.mode === "offline") return offlineOperation();
  try {
    const value = await fetchDocument(`${config.baseUrl}${path}`, {
      ...templateRequestInit(config),
      method: "POST",
      headers: { ...templateHeaders(config, applicationId, scopes, success), "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    return mapEnvelope(value, config, applicationId, success);
  } catch {
    return failedOperation();
  }
}

async function readTemplate(
  config: PromptTemplateConfig,
  applicationId: string,
  path: string,
  success: "restored" | "versioned",
): Promise<PromptTemplateOperation> {
  if (config.mode === "offline") return offlineOperation();
  try {
    const value = await fetchDocument(`${config.baseUrl}${path}`, {
      ...templateRequestInit(config),
      headers: templateHeaders(config, applicationId, ["prompt_application_templates:read_source"], success),
    });
    return mapEnvelope(value, config, applicationId, success);
  } catch {
    return failedOperation();
  }
}

function mapEnvelope(
  value: unknown,
  config: PromptTemplateConfig,
  applicationId: string,
  success: "valid" | "saved" | "restored" | "versioned",
): PromptTemplateOperation {
  if (!isTemplateEnvelope(value, config, applicationId)) return failedOperation();
  const failureCode = nullableString(value.failure_code);
  const validation = mapValidation(value.validation_summary as Document);
  const status = failureCode === "prompt_template_version_conflict"
    ? "version_conflict"
    : failureCode === "prompt_template_scope_denied"
      ? "scope_denied"
      : failureCode
        ? "failed"
        : !validation.isValid
          ? "invalid"
          : success;
  return {
    status,
    draft: value.draft === null ? null : mapDraft(value.draft as Document),
    version: value.version === null ? null : mapVersion(value.version as Document),
    validation,
    failureCode,
    currentDraftVersion: value.current_draft_version as number,
    currentTemplateVersion: value.current_template_version as number,
    summary: failureCode
      ? `Prompt Template 操作失败：${failureCode}。`
      : success === "saved"
        ? `模板草案已保存为 v${value.current_draft_version as number}。`
        : success === "versioned"
          ? `不可变模板版本 v${value.current_template_version as number} 已就绪。`
          : success === "restored"
            ? `模板草案 v${value.current_draft_version as number} 已恢复。`
            : validation.isValid ? "服务端确定性校验通过。" : "服务端校验返回阻塞项。",
  };
}

function isTemplateEnvelope(value: unknown, config: PromptTemplateConfig, applicationId: string): value is Document {
  return isRecord(value) && !containsForbiddenResponse(value) && hasExactKeys(value, ENVELOPE_KEYS) &&
    scopeMatches(value, config, applicationId) && isReference(value.request_id) && isReference(value.audit_ref) &&
    nullableFailure(value.failure_code) && integer(value.current_draft_version, 0) &&
    integer(value.current_template_version, 0) && isValidation(value.validation_summary) &&
    (value.draft === null || isDraft(value.draft, config, applicationId)) &&
    (value.version === null || isVersion(value.version, config, applicationId));
}

function isDraftListEnvelope(value: unknown, config: PromptTemplateConfig, applicationId: string): value is Document {
  return isRecord(value) && !containsForbiddenResponse(value) && hasExactKeys(value, LIST_ENVELOPE_KEYS) &&
    scopeMatches(value, config, applicationId) && isReference(value.request_id) && isReference(value.audit_ref) &&
    nullableFailure(value.failure_code) && Array.isArray(value.draft_summaries) &&
    value.draft_summaries.every((summary) => isDraftSummary(summary, applicationId));
}

function isVersionListEnvelope(
  value: unknown,
  config: PromptTemplateConfig,
  applicationId: string,
  templateId: string,
): value is Document {
  return isRecord(value) && !containsForbiddenResponse(value) && hasExactKeys(value, VERSION_LIST_ENVELOPE_KEYS) &&
    scopeMatches(value, config, applicationId) && value.template_id === templateId &&
    isReference(value.request_id) && isReference(value.audit_ref) && nullableFailure(value.failure_code) &&
    Array.isArray(value.version_summaries) && value.version_summaries.every(isVersionSummary);
}

function isDraft(value: unknown, config: PromptTemplateConfig, applicationId: string): value is Document {
  return isRecord(value) && hasExactKeys(value, DRAFT_KEYS) && value.schema_version === DRAFT_SCHEMA &&
    value.tenant_ref === config.tenantRef && matchesPromptTemplateOwner(value.owner_subject_ref, config) &&
    value.workspace_id === config.workspaceId && value.application_id === applicationId &&
    isDraftInput(value, config, applicationId) && integer(value.draft_version, 1) &&
    isDigest(value.template_digest) && isValidation(value.validation_summary) &&
    (value.validation_summary as Document).state === "valid" &&
    isTimestamp(value.created_at) && isTimestamp(value.updated_at) &&
    isReference(value.created_by_actor_ref) && isReference(value.updated_by_actor_ref) &&
    isReference(value.request_id) && isReference(value.audit_ref);
}

function isVersion(value: unknown, config: PromptTemplateConfig, applicationId: string): value is Document {
  return isRecord(value) && hasExactKeys(value, VERSION_KEYS) && value.schema_version === VERSION_SCHEMA &&
    value.tenant_ref === config.tenantRef && matchesPromptTemplateOwner(value.owner_subject_ref, config) &&
    value.workspace_id === config.workspaceId && value.application_id === applicationId &&
    TEMPLATE_ID.test(String(value.template_id)) && integer(value.template_version, 1) &&
    integer(value.source_draft_version, 1) && validTemplateMetadata(value) &&
    isSource(value) && isDigest(value.template_digest) && isTimestamp(value.created_at) &&
    isReference(value.created_by_actor_ref) && isReference(value.request_id) && isReference(value.audit_ref);
}

function isDraftInput(value: Document, config: PromptTemplateConfig, applicationId: string): boolean {
  return value.schema_version === DRAFT_SCHEMA && TEMPLATE_ID.test(String(value.template_id)) &&
    value.workspace_id === config.workspaceId && value.application_id === applicationId &&
    validTemplateMetadata(value) && isSource(value);
}

function isDraftSummary(value: unknown, applicationId: string): value is Document {
  return isRecord(value) && hasExactKeys(value, DRAFT_SUMMARY_KEYS) && value.schema_version === DRAFT_SCHEMA &&
    TEMPLATE_ID.test(String(value.template_id)) && value.application_id === applicationId &&
    validTemplateMetadata(value) && integer(value.draft_version, 1) && isDigest(value.template_digest) &&
    value.validation_state === "valid" && Array.isArray(value.message_roles) &&
    value.message_roles.length >= 1 && value.message_roles.length <= 16 && value.message_roles.every(isRole) &&
    Array.isArray(value.variable_names) && value.variable_names.length <= 64 &&
    value.variable_names.every((name) => typeof name === "string" && VARIABLE_NAME.test(name)) &&
    isOutputKind(value.output_kind) && isTimestamp(value.updated_at) && isReference(value.updated_by_actor_ref);
}

function isVersionSummary(value: unknown): value is Document {
  return isRecord(value) && hasExactKeys(value, VERSION_SUMMARY_KEYS) && value.schema_version === VERSION_SCHEMA &&
    TEMPLATE_ID.test(String(value.template_id)) && integer(value.template_version, 1) &&
    integer(value.source_draft_version, 1) && typeof value.template_name === "string" &&
    value.template_name.trim().length >= 2 && value.template_name.trim().length <= 80 &&
    isDigest(value.template_digest) && isOutputKind(value.output_kind) &&
    isTimestamp(value.created_at) && isReference(value.created_by_actor_ref);
}

function isSource(value: Document): boolean {
  return Array.isArray(value.messages) && value.messages.length >= 1 && value.messages.length <= 16 &&
    value.messages.every(isMessage) && Array.isArray(value.variables) && value.variables.length <= 64 &&
    value.variables.every(isVariable) && isOutputContractDocument(value.output_contract);
}

function isMessage(value: unknown): boolean {
  return isRecord(value) && hasExactKeys(value, ["role", "content"]) && isRole(value.role) &&
    typeof value.content === "string" && value.content.length > 0 && utf8Bytes(value.content) <= 16 * 1024;
}

function isVariable(value: unknown): boolean {
  if (!isRecord(value)) return false;
  const keys = value.default_value === undefined
    ? ["name", "type", "required", "description"]
    : ["name", "type", "required", "description", "default_value"];
  return hasExactKeys(value, keys) && typeof value.name === "string" && VARIABLE_NAME.test(value.name) &&
    isVariableType(value.type) && typeof value.required === "boolean" &&
    typeof value.description === "string" && value.description.length <= 512 &&
    (value.default_value === undefined || !value.required && valueMatchesType(value.default_value, value.type));
}

function isOutputContractDocument(value: unknown): boolean {
  if (!isRecord(value) || !isOutputKind(value.kind) || typeof value.allow_empty !== "boolean" ||
    !integer(value.max_bytes, 1) || (value.max_bytes as number) > 65536) return false;
  if (value.kind === "text") return hasExactKeys(value, ["kind", "allow_empty", "max_bytes"]);
  return hasExactKeys(value, ["kind", "allow_empty", "max_bytes", "json_schema"]) && isJSONSchema(value.json_schema);
}

function isJSONSchema(value: unknown, depth = 0): boolean {
  if (!isRecord(value) || depth > 8 || !["object", "array", "string", "integer", "number", "boolean"].includes(String(value.type)) ||
    typeof value.additionalProperties !== "boolean") return false;
  const allowed = new Set(["type", "properties", "required", "additionalProperties", "items"]);
  if (Object.keys(value).some((key) => !allowed.has(key))) return false;
  if (value.properties !== undefined && (!isRecord(value.properties) ||
    !Object.entries(value.properties).every(([name, nested]) => VARIABLE_NAME.test(name) && isJSONSchema(nested, depth + 1)))) return false;
  if (value.required !== undefined && (!Array.isArray(value.required) ||
    !value.required.every((item) => typeof item === "string" && item.length > 0 && item.length <= 64) ||
    new Set(value.required).size !== value.required.length)) return false;
  return value.items === undefined || isJSONSchema(value.items, depth + 1);
}

function isValidation(value: unknown): boolean {
  return isRecord(value) && hasExactKeys(value, ["state", "is_valid", "findings"]) &&
    (value.state === "valid" || value.state === "invalid") && value.is_valid === (value.state === "valid") &&
    Array.isArray(value.findings) && value.findings.every((finding) =>
      isRecord(finding) && hasExactKeys(finding, ["code", "field", "summary"]) &&
      isReference(finding.code) && typeof finding.field === "string" && typeof finding.summary === "string"
    );
}

function mapDraft(value: Document): PromptTemplateDraft {
  return {
    ...mapDraftInput(value),
    tenantRef: value.tenant_ref as string,
    ownerSubjectRef: value.owner_subject_ref as string,
    draftVersion: value.draft_version as number,
    templateDigest: value.template_digest as string,
    validation: mapValidation(value.validation_summary as Document),
    createdAt: value.created_at as string,
    updatedAt: value.updated_at as string,
    createdByActorRef: value.created_by_actor_ref as string,
    updatedByActorRef: value.updated_by_actor_ref as string,
    requestId: value.request_id as string,
    auditRef: value.audit_ref as string,
  };
}

function mapVersion(value: Document): PromptTemplateVersion {
  return {
    ...mapDraftInput(value),
    tenantRef: value.tenant_ref as string,
    ownerSubjectRef: value.owner_subject_ref as string,
    templateVersion: value.template_version as number,
    sourceDraftVersion: value.source_draft_version as number,
    templateDigest: value.template_digest as string,
    createdAt: value.created_at as string,
    createdByActorRef: value.created_by_actor_ref as string,
    requestId: value.request_id as string,
    auditRef: value.audit_ref as string,
  };
}

function mapDraftInput(value: Document): PromptTemplateDraftInput {
  return {
    templateId: value.template_id as string,
    workspaceId: value.workspace_id as string,
    applicationId: value.application_id as string,
    templateName: value.template_name as string,
    description: value.description as string,
    messages: (value.messages as Document[]).map((message) => ({
      role: message.role as PromptTemplateRole,
      content: message.content as string,
    })),
    variables: (value.variables as Document[]).map((variable) => ({
      name: variable.name as string,
      type: variable.type as PromptTemplateVariableType,
      required: variable.required as boolean,
      description: variable.description as string,
      ...(variable.default_value === undefined ? {} : { defaultValue: cloneJSONValue(variable.default_value) as PromptTemplateVariable["defaultValue"] }),
    })),
    outputContract: mapOutputContract(value.output_contract as Document),
  };
}

function mapOutputContract(value: Document): PromptTemplateOutputContract {
  return {
    kind: value.kind as PromptTemplateOutputKind,
    allowEmpty: value.allow_empty as boolean,
    maxBytes: value.max_bytes as number,
    ...(value.json_schema === undefined ? {} : { jsonSchema: mapJSONSchema(value.json_schema as Document) }),
  };
}

function mapJSONSchema(value: Document): PromptTemplateJSONSchema {
  const properties = value.properties === undefined
    ? undefined
    : Object.fromEntries(Object.entries(value.properties as Document).map(([name, nested]) => [name, mapJSONSchema(nested as Document)]));
  return {
    type: value.type as PromptTemplateJSONSchema["type"],
    additionalProperties: value.additionalProperties as boolean,
    ...(properties ? { properties } : {}),
    ...(value.required === undefined ? {} : { required: [...value.required as string[]] }),
    ...(value.items === undefined ? {} : { items: mapJSONSchema(value.items as Document) }),
  };
}

function mapValidation(value: Document): PromptTemplateValidation {
  return {
    state: value.state as PromptTemplateValidation["state"],
    isValid: value.is_valid as boolean,
    findings: (value.findings as Document[]).map((finding) => ({
      code: finding.code as string,
      field: finding.field as string,
      summary: finding.summary as string,
    })),
  };
}

function mapDraftSummary(value: Document): PromptTemplateDraftSummary {
  return {
    templateId: value.template_id as string,
    applicationId: value.application_id as string,
    templateName: value.template_name as string,
    description: value.description as string,
    draftVersion: value.draft_version as number,
    templateDigest: value.template_digest as string,
    validationState: "valid",
    messageRoles: [...value.message_roles as PromptTemplateRole[]],
    variableNames: [...value.variable_names as string[]],
    outputKind: value.output_kind as PromptTemplateOutputKind,
    updatedAt: value.updated_at as string,
    updatedByActorRef: value.updated_by_actor_ref as string,
  };
}

function mapVersionSummary(value: Document): PromptTemplateVersionSummary {
  return {
    templateId: value.template_id as string,
    templateVersion: value.template_version as number,
    sourceDraftVersion: value.source_draft_version as number,
    templateName: value.template_name as string,
    templateDigest: value.template_digest as string,
    outputKind: value.output_kind as PromptTemplateOutputKind,
    createdAt: value.created_at as string,
    createdByActorRef: value.created_by_actor_ref as string,
  };
}

function draftInputPayload(input: PromptTemplateDraftInput): Document {
  return {
    schema_version: DRAFT_SCHEMA,
    template_id: input.templateId.trim(),
    workspace_id: input.workspaceId.trim(),
    application_id: input.applicationId.trim(),
    template_name: input.templateName.trim(),
    description: input.description.trim(),
    messages: input.messages.map((message) => ({ role: message.role, content: message.content })),
    variables: input.variables.map((variable) => ({
      name: variable.name.trim(),
      type: variable.type,
      required: variable.required,
      description: variable.description.trim(),
      ...(variable.defaultValue === undefined ? {} : { default_value: cloneJSONValue(variable.defaultValue) }),
    })),
    output_contract: {
      kind: input.outputContract.kind,
      allow_empty: input.outputContract.allowEmpty,
      max_bytes: input.outputContract.maxBytes,
      ...(input.outputContract.jsonSchema === undefined ? {} : { json_schema: jsonSchemaPayload(input.outputContract.jsonSchema) }),
    },
  };
}

function jsonSchemaPayload(value: PromptTemplateJSONSchema): Document {
  return {
    type: value.type,
    additionalProperties: value.additionalProperties,
    ...(value.properties === undefined ? {} : {
      properties: Object.fromEntries(Object.entries(value.properties).map(([name, nested]) => [name, jsonSchemaPayload(nested)])),
    }),
    ...(value.required === undefined ? {} : { required: [...value.required] }),
    ...(value.items === undefined ? {} : { items: jsonSchemaPayload(value.items) }),
  };
}

function templateHeaders(
  config: PromptTemplateConfig,
  applicationId: string,
  scopes: string[],
  operation: string,
): Record<string, string> {
  const requestId = createRequestId(`prompt-template-${operation}`);
  const mutationPermissions = scopes.filter((scope) =>
    scope === "prompt_application_templates:write" || scope === "prompt_application_templates:version"
  );
  if (config.authMode === "local_session_dev_test") {
    return {
      Accept: "application/json",
      "X-Request-Id": requestId,
      "X-RadishMind-Active-Tenant": config.tenantRef,
      ...(mutationPermissions.length === 0 ? {} : { "X-RadishMind-Active-Workspace": config.workspaceId }),
      "X-RadishMind-Dev-Prompt-Template-Workspace": config.workspaceId,
      "X-RadishMind-Dev-Prompt-Template-Application": applicationId,
    };
  }
  return {
    Accept: "application/json",
    "X-Request-Id": requestId,
    "X-RadishMind-Dev-Read-Identity": "radishmind-web-prompt-template",
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Read-Scopes": scopes.join(","),
    "X-RadishMind-Dev-Read-Audit": `audit_${requestId}`,
    ...(mutationPermissions.length === 0 ? {} : {
      "X-RadishMind-Active-Workspace": config.workspaceId,
      "X-RadishMind-Dev-Read-Membership-Workspace": config.workspaceId,
      "X-RadishMind-Dev-Read-Membership-Permissions": mutationPermissions.join(","),
    }),
    "X-RadishMind-Dev-Prompt-Template-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Prompt-Template-Application": applicationId,
  };
}

function templateRequestInit(config: PromptTemplateConfig): Pick<RequestInit, "credentials" | "cache"> {
  return {
    credentials: config.authMode === "local_session_dev_test" ? "include" : "omit",
    cache: "no-store",
  };
}

function matchesPromptTemplateOwner(value: unknown, config: PromptTemplateConfig): value is string {
  return typeof value === "string" && (config.authMode === "local_session_dev_test"
    ? LOCAL_USER_ACTOR_PATTERN.test(value)
    : value === config.subjectRef);
}

function templateQuery(config: PromptTemplateConfig, applicationId: string): string {
  return new URLSearchParams({ workspace_id: config.workspaceId, application_id: applicationId }).toString();
}

async function fetchDocument(url: string, init: RequestInit): Promise<unknown> {
  const response = await fetch(url, init);
  const value: unknown = await response.json();
  if (!response.ok) throw new Error("Prompt Application HTTP request failed.");
  return value;
}

function offlineOperation(): PromptTemplateOperation {
  return {
    status: "offline",
    draft: null,
    version: null,
    validation: { state: "invalid", isValid: false, findings: [] },
    failureCode: "prompt_application_http_disabled",
    currentDraftVersion: 0,
    currentTemplateVersion: 0,
    summary: "离线模式只保留当前组件内的临时编辑，不发送请求。",
  };
}

function failedOperation(failureCode = "prompt_template_store_unavailable"): PromptTemplateOperation {
  return {
    ...offlineOperation(),
    status: "failed",
    failureCode,
    summary: `Prompt Template owner 不可用：${failureCode}。`,
  };
}

function offlineList(): PromptTemplateListResult {
  return { status: "offline", summaries: [], failureCode: "prompt_application_http_disabled", summary: "离线模式不读取模板 owner。" };
}

function failedList(failureCode = "prompt_template_store_unavailable"): PromptTemplateListResult {
  return { status: "failed", summaries: [], failureCode, summary: "模板草案列表读取失败。" };
}

function offlineVersionList(): PromptTemplateVersionListResult {
  return { status: "offline", summaries: [], failureCode: "prompt_application_http_disabled", summary: "离线模式不读取模板版本。" };
}

function failedVersionList(failureCode = "prompt_template_store_unavailable"): PromptTemplateVersionListResult {
  return { status: "failed", summaries: [], failureCode, summary: "模板版本列表读取失败。" };
}

function parseTemplate(content: string): { valid: boolean; variables: string[] } {
  const variables: string[] = [];
  let cursor = 0;
  const pattern = /\{\{\s*([A-Za-z][A-Za-z0-9_]*)\s*\}\}/gu;
  for (const match of content.matchAll(pattern)) {
    const index = match.index ?? 0;
    const literal = content.slice(cursor, index);
    if (literal.includes("{{") || literal.includes("}}")) return { valid: false, variables: [] };
    variables.push(match[1]);
    cursor = index + match[0].length;
  }
  const tail = content.slice(cursor);
  return { valid: !tail.includes("{{") && !tail.includes("}}"), variables };
}

function valueMatchesType(value: unknown, type: PromptTemplateVariableType): boolean {
  if (type === "string") return typeof value === "string" && utf8Bytes(value) <= 16 * 1024;
  if (type === "integer") return typeof value === "number" && Number.isSafeInteger(value);
  if (type === "number") return typeof value === "number" && Number.isFinite(value);
  if (type === "boolean") return typeof value === "boolean";
  return Array.isArray(value) && value.length <= 128 &&
    value.every((item) => typeof item === "string" && utf8Bytes(item) <= 16 * 1024) &&
    utf8Bytes(JSON.stringify(value)) <= 32 * 1024;
}

function canonicalVariableValue(value: unknown, type: PromptTemplateVariableType): string {
  if (type === "string") return value as string;
  if (type === "boolean") return value ? "true" : "false";
  return JSON.stringify(value);
}

function isOutputContract(value: PromptTemplateOutputContract): boolean {
  if (!isOutputKind(value.kind) || !Number.isInteger(value.maxBytes) || value.maxBytes < 1 || value.maxBytes > 65536) return false;
  return value.kind === "text" ? value.jsonSchema === undefined : value.jsonSchema !== undefined && isJSONSchema(jsonSchemaPayload(value.jsonSchema));
}

function validTemplateMetadata(value: Document): boolean {
  return typeof value.template_name === "string" && value.template_name.trim().length >= 2 &&
    value.template_name.trim().length <= 80 && typeof value.description === "string" &&
    value.description.length <= 512;
}

function scopeMatches(value: Document, config: PromptTemplateConfig, applicationId: string): boolean {
  return value.workspace_id === config.workspaceId && value.application_id === applicationId;
}

function containsForbiddenResponse(value: unknown): boolean {
  if (Array.isArray(value)) return value.some(containsForbiddenResponse);
  if (!isRecord(value)) return false;
  return Object.entries(value).some(([key, nested]) =>
    FORBIDDEN_RESPONSE_KEYS.has(key.toLowerCase()) || containsForbiddenResponse(nested)
  );
}

function hasExactKeys(value: Document, expected: readonly string[]): boolean {
  const keys = Object.keys(value).sort();
  const target = [...expected].sort();
  return keys.length === target.length && keys.every((key, index) => key === target[index]);
}

function isRecord(value: unknown): value is Document {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function isRole(value: unknown): value is PromptTemplateRole {
  return value === "system" || value === "developer" || value === "user";
}

function isVariableType(value: unknown): value is PromptTemplateVariableType {
  return value === "string" || value === "integer" || value === "number" ||
    value === "boolean" || value === "string_list";
}

function isOutputKind(value: unknown): value is PromptTemplateOutputKind {
  return value === "text" || value === "json_object";
}

function isReference(value: unknown): value is string {
  return typeof value === "string" && REFERENCE.test(value);
}

function isDigest(value: unknown): value is string {
  return typeof value === "string" && DIGEST.test(value);
}

function isTimestamp(value: unknown): value is string {
  return typeof value === "string" && value.trim().length > 0 && !Number.isNaN(Date.parse(value));
}

function integer(value: unknown, minimum: number): boolean {
  return Number.isInteger(value) && (value as number) >= minimum;
}

function nullableFailure(value: unknown): boolean {
  return value === null || typeof value === "string" && /^prompt_[a-z_]{3,100}$/u.test(value);
}

function nullableString(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function normalizeBaseUrl(value: string): string {
  return value.trim().replace(/\/$/u, "");
}

function utf8Bytes(value: string): number {
  return new TextEncoder().encode(value).byteLength;
}

function createRequestId(prefix: string): string {
  const suffix = globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random()}`;
  return `${prefix}-${suffix.replaceAll("-", "").slice(0, 16)}`;
}

function createPromptTemplateId(): string {
  const alphabet = "abcdefghijklmnopqrstuvwxyz234567";
  const bytes = new Uint8Array(16);
  if (globalThis.crypto?.getRandomValues) {
    globalThis.crypto.getRandomValues(bytes);
  } else {
    const seed = `${Date.now()}${Math.random()}`;
    for (let index = 0; index < bytes.length; index += 1) bytes[index] = seed.charCodeAt(index % seed.length);
  }
  return `ptpl_${[...bytes].map((value) => alphabet[value % alphabet.length]).join("")}`;
}

function cloneJSONValue(value: unknown): unknown {
  return JSON.parse(JSON.stringify(value)) as unknown;
}
