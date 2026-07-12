import {
  initialApplicationModelCatalogState,
  loadApplicationModelCatalog,
  readApplicationApiIntegrationConfig,
  type ApplicationApiProtocol,
  type ApplicationModelCatalogItem,
  type ApplicationModelCatalogState,
} from "./applicationApiIntegrationConsumer.ts";

const APPLICATION_DRAFT_SCHEMA_VERSION = "application_configuration_draft.v1";
const DEV_SOURCE = "dev-application-draft-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const FORBIDDEN_DRAFT_RESPONSE_FIELDS = new Set([
  "authorization", "api_key", "key_hash", "secret", "credential", "endpoint", "headers", "cookie",
  "raw_request", "raw_response", "input", "output", "prompt", "messages",
]);

export type ApplicationConfigurationDraftConfig = {
  mode: "offline" | "dev_application_draft_http";
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  subjectRef: string;
};

export type ApplicationConfigurationBaseline = {
  applicationId: string;
  displayName: string;
  applicationKind: string;
  updatedAt: string;
};

export type ApplicationConfigurationDraft = {
  draftId: string;
  workspaceId: string;
  applicationId: string;
  baseApplicationUpdatedAt: string;
  schemaVersion: typeof APPLICATION_DRAFT_SCHEMA_VERSION;
  displayName: string;
  description: string;
  applicationKind: string;
  defaultProtocol: ApplicationApiProtocol;
  defaultModel: string;
  allowedProtocols: ApplicationApiProtocol[];
};

export type ApplicationConfigurationDraftFinding = {
  code: string;
  field: string;
  summary: string;
};

export type ApplicationConfigurationDraftValidation = {
  state: "valid" | "invalid";
  isValid: boolean;
  findings: ApplicationConfigurationDraftFinding[];
};

export type ApplicationConfigurationDraftOperationState = {
  status: "offline" | "unsaved" | "validating" | "invalid" | "valid" | "saving" | "saved" | "loading" | "restored" | "version_conflict" | "store_failure" | "scope_denied";
  summary: string;
  failureCode: string;
  currentDraftVersion: number;
  validation: ApplicationConfigurationDraftValidation;
  requestId: string;
  auditRef: string;
};

export type ApplicationConfigurationDraftSummary = {
  draftId: string;
  applicationId: string;
  draftVersion: number;
  displayName: string;
  applicationKind: string;
  defaultProtocol: ApplicationApiProtocol;
  defaultModel: string;
  validationState: string;
  updatedAt: string;
  updatedByActorRef: string;
};

export type ApplicationConfigurationDraftListState = {
  status: "offline" | "idle" | "loading" | "ready" | "empty" | "failed";
  summaries: ApplicationConfigurationDraftSummary[];
  failureCode: string;
  summary: string;
};

export type ApplicationConfigurationDiff = {
  field: "display_name" | "application_kind" | "default_protocol" | "default_model";
  before: string;
  after: string;
  changed: boolean;
};

type DraftDocument = {
  draft_id: string;
  workspace_id: string;
  application_id: string;
  base_application_updated_at: string;
  schema_version: string;
  display_name: string;
  description: string;
  application_kind: string;
  default_protocol: string;
  default_model: string;
  allowed_protocols: string[];
  draft_version: number;
  validation_summary: ValidationDocument;
  created_at: string;
  updated_at: string;
  created_by_actor_ref: string;
  updated_by_actor_ref: string;
  request_id: string;
  audit_ref: string;
};

type ValidationDocument = { state: string; is_valid: boolean; findings: Array<{ code: string; field: string; summary: string }> };
type DraftEnvelope = {
  request_id: string;
  workspace_id: string;
  application_id: string;
  draft: DraftDocument | null;
  failure_code: string | null;
  current_draft_version: number;
  validation_summary: ValidationDocument;
  audit_ref: string;
};
type DraftListEnvelope = {
  request_id: string;
  workspace_id: string;
  application_id: string;
  draft_summaries: Array<{
    draft_id: string; application_id: string; draft_version: number; display_name: string; application_kind: string;
    default_protocol: string; default_model: string; validation_state: string; updated_at: string; updated_by_actor_ref: string;
  }>;
  failure_code: string | null;
  audit_ref: string;
};

export function readApplicationConfigurationDraftConfig(): ApplicationConfigurationDraftConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  return {
    mode: env.VITE_RADISHMIND_APPLICATION_DRAFT_SOURCE?.trim() === DEV_SOURCE ? "dev_application_draft_http" : "offline",
    baseUrl: (env.VITE_RADISHMIND_APPLICATION_DRAFT_BASE_URL ?? env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ?? DEFAULT_BASE_URL).trim().replace(/\/$/u, ""),
    tenantRef: env.VITE_RADISHMIND_DEV_READ_TENANT_REF?.trim() || "tenant_demo",
    workspaceId: env.VITE_RADISHMIND_APPLICATION_DRAFT_WORKSPACE_ID?.trim() || "workspace_demo",
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || "subject_demo_user",
  };
}

export function createApplicationConfigurationDraft(
  config: ApplicationConfigurationDraftConfig,
  baseline: ApplicationConfigurationBaseline,
): ApplicationConfigurationDraft {
  return {
    draftId: `app-config-${baseline.applicationId}`,
    workspaceId: config.workspaceId,
    applicationId: baseline.applicationId,
    baseApplicationUpdatedAt: baseline.updatedAt,
    schemaVersion: APPLICATION_DRAFT_SCHEMA_VERSION,
    displayName: baseline.displayName,
    description: "",
    applicationKind: baseline.applicationKind,
    defaultProtocol: "responses",
    defaultModel: "",
    allowedProtocols: ["chat_completions", "responses", "messages"],
  };
}

export function initialApplicationConfigurationDraftState(config: ApplicationConfigurationDraftConfig): ApplicationConfigurationDraftOperationState {
  return {
    status: config.mode === "offline" ? "offline" : "unsaved",
    summary: config.mode === "offline" ? "Offline draft editing stays in component memory and sends no requests." : "Edit and validate this application configuration before saving.",
    failureCode: "", currentDraftVersion: 0, validation: { state: "invalid", isValid: false, findings: [] }, requestId: "", auditRef: "",
  };
}

export function initialApplicationConfigurationDraftListState(config: ApplicationConfigurationDraftConfig): ApplicationConfigurationDraftListState {
  return config.mode === "offline"
    ? { status: "offline", summaries: [], failureCode: "application_draft_http_disabled", summary: "Offline mode does not load saved application drafts." }
    : { status: "idle", summaries: [], failureCode: "", summary: "Load saved drafts for the selected application." };
}

export function initialApplicationDraftModelCatalog(config: ApplicationConfigurationDraftConfig, applicationId: string): ApplicationModelCatalogState {
  const integration = readApplicationApiIntegrationConfig();
  return initialApplicationModelCatalogState(config.mode === "offline" ? { ...integration, mode: "offline" } : integration, applicationId);
}

export async function loadApplicationDraftModelCatalog(applicationId: string, signal?: AbortSignal): Promise<ApplicationModelCatalogState> {
  return loadApplicationModelCatalog(readApplicationApiIntegrationConfig(), applicationId, signal);
}

export function validateApplicationConfigurationDraft(
  draft: ApplicationConfigurationDraft,
  models: ApplicationModelCatalogItem[],
): ApplicationConfigurationDraftValidation {
  const findings: ApplicationConfigurationDraftFinding[] = [];
  const add = (code: string, field: string, summary: string) => findings.push({ code, field, summary });
  if (!/^[A-Za-z0-9][A-Za-z0-9._:-]{0,159}$/u.test(draft.draftId) || !/^[A-Za-z0-9][A-Za-z0-9._:-]{0,159}$/u.test(draft.applicationId)) add("application_draft_payload_invalid", "scope", "Draft and application identifiers must be stable safe values.");
  if (draft.displayName.trim().length < 2 || draft.displayName.trim().length > 120) add("application_draft_payload_invalid", "display_name", "Display name must contain 2 to 120 characters.");
  if (draft.description.trim().length > 1000) add("application_draft_payload_invalid", "description", "Description must not exceed 1000 characters.");
  if (!["workflow_copilot", "docs_qa", "agent", "prompt_application"].includes(draft.applicationKind)) add("application_draft_payload_invalid", "application_kind", "Application kind is unsupported.");
  if (draft.allowedProtocols.length === 0 || new Set(draft.allowedProtocols).size !== draft.allowedProtocols.length || !draft.allowedProtocols.includes(draft.defaultProtocol)) add("application_draft_payload_invalid", "allowed_protocols", "Allowed protocols must be unique and include the default protocol.");
  if (!draft.defaultModel.trim()) add("application_draft_payload_invalid", "default_model", "Select a validated model.");
  const selectedModel = models.find((model) => model.id === draft.defaultModel);
  if (draft.defaultModel && !selectedModel) add("application_draft_model_unavailable", "default_model", "Default model is not present in the current validated catalog.");
  if (selectedModel && !selectedModel.protocols.includes(draft.defaultProtocol)) add("application_draft_protocol_incompatible", "default_protocol", "Default protocol is not supported by the selected model.");
  for (const [field, value] of [["display_name", draft.displayName], ["description", draft.description], ["default_model", draft.defaultModel]] as const) {
    if (/authorization:|bearer\s|api[_-]?key=|x-radishmind-dev-|sk-/iu.test(value)) add("application_draft_secret_material_forbidden", field, "Secret or internal caller material is forbidden in application drafts.");
  }
  return { state: findings.length === 0 ? "valid" : "invalid", isValid: findings.length === 0, findings };
}

export function compareApplicationConfigurationDraft(baseline: ApplicationConfigurationBaseline, draft: ApplicationConfigurationDraft): ApplicationConfigurationDiff[] {
  return [
    diff("display_name", baseline.displayName, draft.displayName),
    diff("application_kind", baseline.applicationKind, draft.applicationKind),
    diff("default_protocol", "not configured in read model", draft.defaultProtocol),
    diff("default_model", "not configured in read model", draft.defaultModel || "not selected"),
  ];
}

export async function validateApplicationConfigurationDraftRemote(config: ApplicationConfigurationDraftConfig, draft: ApplicationConfigurationDraft): Promise<ApplicationConfigurationDraftOperationState> {
  return writeDraftRequest(config, draft, "/v1/user-workspace/application-drafts/validate", { draft: draftPayload(draft) }, "valid");
}

export async function saveApplicationConfigurationDraft(config: ApplicationConfigurationDraftConfig, draft: ApplicationConfigurationDraft, expectedVersion: number): Promise<ApplicationConfigurationDraftOperationState> {
  return writeDraftRequest(config, draft, "/v1/user-workspace/application-drafts", { expected_draft_version: expectedVersion, draft: draftPayload(draft) }, "saved");
}

export async function listApplicationConfigurationDrafts(config: ApplicationConfigurationDraftConfig, applicationId: string): Promise<ApplicationConfigurationDraftListState> {
  if (config.mode !== "dev_application_draft_http") return initialApplicationConfigurationDraftListState(config);
  const requestId = createRequestId("app-draft-list");
  try {
    const response = await fetch(`${config.baseUrl}/v1/user-workspace/application-drafts?workspace_id=${encodeURIComponent(config.workspaceId)}&application_id=${encodeURIComponent(applicationId)}`, { headers: draftHeaders(config, applicationId, requestId, "read") });
    const document: unknown = await response.json();
    if (!response.ok || !isDraftListEnvelope(document, config, applicationId)) throw new Error("invalid application draft list response");
    if (document.failure_code) return { status: "failed", summaries: [], failureCode: document.failure_code, summary: "Saved application drafts could not be loaded." };
    const summaries = document.draft_summaries.map(mapDraftSummary);
    return { status: summaries.length ? "ready" : "empty", summaries, failureCode: "", summary: summaries.length ? `Loaded ${summaries.length} saved application drafts.` : "No saved drafts exist for this application." };
  } catch {
    return { status: "failed", summaries: [], failureCode: "application_draft_store_unavailable", summary: "Saved application drafts could not be loaded." };
  }
}

export async function readApplicationConfigurationDraft(config: ApplicationConfigurationDraftConfig, applicationId: string, draftId: string): Promise<{ draft: ApplicationConfigurationDraft | null; state: ApplicationConfigurationDraftOperationState }> {
  if (config.mode !== "dev_application_draft_http") return { draft: null, state: initialApplicationConfigurationDraftState(config) };
  const requestId = createRequestId("app-draft-read");
  try {
    const response = await fetch(`${config.baseUrl}/v1/user-workspace/application-drafts/${encodeURIComponent(draftId)}?workspace_id=${encodeURIComponent(config.workspaceId)}&application_id=${encodeURIComponent(applicationId)}`, { headers: draftHeaders(config, applicationId, requestId, "read") });
    const document: unknown = await response.json();
    if (!response.ok || !isDraftEnvelope(document, config, applicationId)) throw new Error("invalid application draft response");
    if (document.failure_code || !document.draft) return { draft: null, state: operationStateFromEnvelope(document, "store_failure") };
    return { draft: mapDraft(document.draft), state: operationStateFromEnvelope(document, "restored") };
  } catch {
    return { draft: null, state: failedOperationState("application_draft_store_unavailable") };
  }
}

async function writeDraftRequest(config: ApplicationConfigurationDraftConfig, draft: ApplicationConfigurationDraft, path: string, body: unknown, successStatus: "valid" | "saved"): Promise<ApplicationConfigurationDraftOperationState> {
  if (config.mode !== "dev_application_draft_http") return initialApplicationConfigurationDraftState(config);
  const requestId = createRequestId("app-draft-write");
  try {
    const response = await fetch(`${config.baseUrl}${path}`, { method: "POST", headers: { ...draftHeaders(config, draft.applicationId, requestId, "write"), "Content-Type": "application/json" }, body: JSON.stringify(body) });
    const document: unknown = await response.json();
    if (!response.ok || !isDraftEnvelope(document, config, draft.applicationId)) throw new Error("invalid application draft response");
    return operationStateFromEnvelope(document, successStatus);
  } catch {
    return failedOperationState("application_draft_store_unavailable");
  }
}

function draftPayload(draft: ApplicationConfigurationDraft) {
  return {
    draft_id: draft.draftId, workspace_id: draft.workspaceId, application_id: draft.applicationId,
    base_application_updated_at: draft.baseApplicationUpdatedAt, schema_version: draft.schemaVersion,
    display_name: draft.displayName, description: draft.description, application_kind: draft.applicationKind,
    default_protocol: draft.defaultProtocol, default_model: draft.defaultModel, allowed_protocols: draft.allowedProtocols,
  };
}

function draftHeaders(config: ApplicationConfigurationDraftConfig, applicationId: string, requestId: string, operation: "read" | "write"): HeadersInit {
  return {
    Accept: "application/json", "X-Request-Id": requestId,
    "X-RadishMind-Dev-Read-Identity": "radishmind-web-application-draft",
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Read-Scopes": operation === "write" ? "application_drafts:read,application_drafts:write" : "application_drafts:read",
    "X-RadishMind-Dev-Read-Audit": `audit_${requestId}_application_draft`,
    "X-RadishMind-Dev-Application-Draft-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Application-Draft-Application": applicationId,
  };
}

function operationStateFromEnvelope(document: DraftEnvelope, successStatus: "valid" | "saved" | "restored" | "store_failure"): ApplicationConfigurationDraftOperationState {
  const validation = mapValidation(document.validation_summary);
  const failureCode = document.failure_code ?? "";
  let status: ApplicationConfigurationDraftOperationState["status"] = successStatus;
  if (!failureCode && !validation.isValid) status = "invalid";
  else if (failureCode === "application_draft_version_conflict") status = "version_conflict";
  else if (failureCode === "application_draft_scope_denied") status = "scope_denied";
  else if (failureCode) status = failureCode.includes("payload") || failureCode.includes("secret") ? "invalid" : "store_failure";
  return { status, summary: failureCode ? `Application draft operation failed: ${failureCode}.` : status === "saved" ? `Saved application draft version ${document.current_draft_version}.` : status === "restored" ? `Restored application draft version ${document.current_draft_version}.` : "Application configuration is valid for dev/test review.", failureCode, currentDraftVersion: document.current_draft_version, validation, requestId: document.request_id, auditRef: document.audit_ref };
}

function failedOperationState(failureCode: string): ApplicationConfigurationDraftOperationState {
  return { status: "store_failure", summary: "Application draft store is unavailable.", failureCode, currentDraftVersion: 0, validation: { state: "invalid", isValid: false, findings: [] }, requestId: "", auditRef: "" };
}

function mapDraft(document: DraftDocument): ApplicationConfigurationDraft {
  return { draftId: document.draft_id, workspaceId: document.workspace_id, applicationId: document.application_id, baseApplicationUpdatedAt: document.base_application_updated_at, schemaVersion: APPLICATION_DRAFT_SCHEMA_VERSION, displayName: document.display_name, description: document.description, applicationKind: document.application_kind, defaultProtocol: document.default_protocol as ApplicationApiProtocol, defaultModel: document.default_model, allowedProtocols: document.allowed_protocols as ApplicationApiProtocol[] };
}

function mapDraftSummary(document: DraftListEnvelope["draft_summaries"][number]): ApplicationConfigurationDraftSummary {
  return { draftId: document.draft_id, applicationId: document.application_id, draftVersion: document.draft_version, displayName: document.display_name, applicationKind: document.application_kind, defaultProtocol: document.default_protocol as ApplicationApiProtocol, defaultModel: document.default_model, validationState: document.validation_state, updatedAt: document.updated_at, updatedByActorRef: document.updated_by_actor_ref };
}

function mapValidation(document: ValidationDocument): ApplicationConfigurationDraftValidation {
  return { state: document.is_valid ? "valid" : "invalid", isValid: document.is_valid, findings: document.findings.map((finding) => ({ ...finding })) };
}

function isDraftEnvelope(value: unknown, config: ApplicationConfigurationDraftConfig, applicationId: string): value is DraftEnvelope {
  if (!isRecord(value) || containsForbiddenDraftResponseField(value) || value.workspace_id !== config.workspaceId || value.application_id !== applicationId || typeof value.request_id !== "string" || typeof value.current_draft_version !== "number" || !(value.failure_code === null || typeof value.failure_code === "string") || !isValidationDocument(value.validation_summary)) return false;
  return value.draft === null || isDraftDocument(value.draft, config, applicationId);
}

function isDraftListEnvelope(value: unknown, config: ApplicationConfigurationDraftConfig, applicationId: string): value is DraftListEnvelope {
  return isRecord(value) && !containsForbiddenDraftResponseField(value) && value.workspace_id === config.workspaceId && value.application_id === applicationId && typeof value.request_id === "string" && (value.failure_code === null || typeof value.failure_code === "string") && Array.isArray(value.draft_summaries) && value.draft_summaries.every((summary) => isRecord(summary) && summary.application_id === applicationId && typeof summary.draft_id === "string" && typeof summary.draft_version === "number" && typeof summary.display_name === "string" && typeof summary.default_protocol === "string" && typeof summary.default_model === "string");
}

function isDraftDocument(value: unknown, config: ApplicationConfigurationDraftConfig, applicationId: string): value is DraftDocument {
  return isRecord(value) && value.workspace_id === config.workspaceId && value.application_id === applicationId && value.schema_version === APPLICATION_DRAFT_SCHEMA_VERSION && typeof value.draft_id === "string" && typeof value.display_name === "string" && typeof value.description === "string" && typeof value.application_kind === "string" && typeof value.default_protocol === "string" && typeof value.default_model === "string" && Array.isArray(value.allowed_protocols) && value.allowed_protocols.every((protocol) => typeof protocol === "string") && typeof value.draft_version === "number" && isValidationDocument(value.validation_summary);
}

function isValidationDocument(value: unknown): value is ValidationDocument {
  return isRecord(value) && (value.state === "valid" || value.state === "invalid") && typeof value.is_valid === "boolean" && Array.isArray(value.findings) && value.findings.every((finding) => isRecord(finding) && typeof finding.code === "string" && typeof finding.field === "string" && typeof finding.summary === "string");
}

function diff(field: ApplicationConfigurationDiff["field"], before: string, after: string): ApplicationConfigurationDiff { return { field, before, after, changed: before !== after }; }
function createRequestId(prefix: string): string { return `${prefix}-${Date.now()}-${(globalThis.crypto?.randomUUID?.() ?? Math.random().toString(16).slice(2)).replaceAll("-", "").slice(0, 12)}`; }
function containsForbiddenDraftResponseField(value: unknown): boolean {
  if (Array.isArray(value)) return value.some(containsForbiddenDraftResponseField);
  if (!isRecord(value)) return false;
  return Object.entries(value).some(([key, nested]) => FORBIDDEN_DRAFT_RESPONSE_FIELDS.has(key.toLowerCase()) || containsForbiddenDraftResponseField(nested));
}
function isRecord(value: unknown): value is Record<string, any> { return Boolean(value) && typeof value === "object" && !Array.isArray(value); }
