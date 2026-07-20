import { CONTROL_PLANE_READ_ROUTES } from "../../../../../contracts/typescript/control-plane-read-api.ts";

const APPLICATION_CATALOG_SCHEMA_VERSION = "application_catalog_record.v1";
const APPLICATION_CATALOG_COLLECTION_PATH = CONTROL_PLANE_READ_ROUTES.applications;
const DEV_SOURCE = "dev-application-catalog-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const APPLICATION_ID_PATTERN = /^app_[a-z0-9]{16}$/u;
const SCOPE_ID_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._:-]{2,127}$/u;
const APPLICATION_KINDS = ["workflow_copilot", "docs_qa", "agent", "prompt_application"] as const;
const RECORD_KEYS = [
  "schema_version", "application_id", "tenant_ref", "workspace_id", "owner_subject_ref", "display_name",
  "description", "application_kind", "lifecycle_state", "record_version", "created_at", "updated_at",
  "archived_at", "created_by_actor_ref", "updated_by_actor_ref", "request_id", "audit_ref",
] as const;
const SUMMARY_KEYS = [
  "application_ref", "tenant_ref", "workspace_id", "application_kind", "display_name", "description",
  "owner_subject_ref", "latest_workflow_definition_ref", "last_run_status", "lifecycle_state", "record_version",
  "created_at", "updated_at", "archived_at",
] as const;
const OPERATION_ENVELOPE_KEYS = [
  "request_id", "tenant_ref", "workspace_id", "record", "failure_code", "current_record_version",
  "current_lifecycle_state", "audit_ref",
] as const;
const LIST_ENVELOPE_KEYS = ["request_id", "tenant_ref", "items", "next_cursor", "failure_code", "audit_ref"] as const;
const FORBIDDEN_RESPONSE_FIELDS = new Set([
  "authorization", "api_key", "key_hash", "secret", "credential", "endpoint", "headers", "cookie", "dsn",
  "raw_request", "raw_response", "input", "output", "prompt", "messages",
]);

export type ApplicationCatalogMode = "offline" | "dev_application_catalog_http";
export type ApplicationCatalogLifecycleState = "active" | "archived";
export type ApplicationCatalogKind = typeof APPLICATION_KINDS[number];
export type ApplicationCatalogAuthMode = "dev_headers" | "signed_test_token" | "radish_oidc_integration_test";

export type ApplicationCatalogConfig = {
  mode: ApplicationCatalogMode;
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  subjectRef: string;
  authMode: ApplicationCatalogAuthMode;
};

export type ApplicationCatalogRecord = {
  schemaVersion: typeof APPLICATION_CATALOG_SCHEMA_VERSION;
  applicationId: string;
  tenantRef: string;
  workspaceId: string;
  ownerSubjectRef: string;
  displayName: string;
  description: string;
  applicationKind: ApplicationCatalogKind;
  lifecycleState: ApplicationCatalogLifecycleState;
  recordVersion: number;
  createdAt: string;
  updatedAt: string;
  archivedAt: string | null;
  createdByActorRef: string;
  updatedByActorRef: string;
  requestId: string;
  auditRef: string;
};

export type ApplicationCatalogMutableFields = {
  displayName: string;
  description: string;
  applicationKind: ApplicationCatalogKind;
};

export type ApplicationCatalogListResult = {
  status: "offline" | "ready" | "empty" | "failed";
  records: ApplicationCatalogRecord[];
  nextCursor: string;
  failureCode: string;
  requestId: string;
  auditRef: string;
  summary: string;
};

export type ApplicationCatalogOperationResult = {
  status: "offline" | "created" | "loaded" | "updated" | "archived" | "version_conflict" | "record_archived" | "scope_denied" | "failed";
  record: ApplicationCatalogRecord | null;
  failureCode: string;
  currentRecordVersion: number;
  currentLifecycleState: string;
  requestId: string;
  auditRef: string;
  summary: string;
};

type ApplicationCatalogRecordDocument = {
  schema_version: string;
  application_id: string;
  tenant_ref: string;
  workspace_id: string;
  owner_subject_ref: string;
  display_name: string;
  description: string;
  application_kind: string;
  lifecycle_state: string;
  record_version: number;
  created_at: string;
  updated_at: string;
  archived_at: string | null;
  created_by_actor_ref: string;
  updated_by_actor_ref: string;
  request_id: string;
  audit_ref: string;
};

type ApplicationCatalogSummaryDocument = {
  application_ref: string;
  tenant_ref: string;
  workspace_id: string;
  application_kind: string;
  display_name: string;
  description: string;
  owner_subject_ref: string;
  latest_workflow_definition_ref: string;
  last_run_status: string;
  lifecycle_state: string;
  record_version: number;
  created_at: string;
  updated_at: string;
  archived_at: string | null;
};

type ApplicationCatalogOperationEnvelope = {
  request_id: string;
  tenant_ref: string;
  workspace_id: string;
  record: ApplicationCatalogRecordDocument | null;
  failure_code: string | null;
  current_record_version: number;
  current_lifecycle_state: string;
  audit_ref: string;
};

type ApplicationCatalogListEnvelope = {
  request_id: string;
  tenant_ref: string;
  items: ApplicationCatalogSummaryDocument[];
  next_cursor: string | null;
  failure_code: string | null;
  audit_ref: string;
};

export function readApplicationCatalogConfig(): ApplicationCatalogConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  return {
    mode: env.VITE_RADISHMIND_APPLICATION_CATALOG_SOURCE?.trim() === DEV_SOURCE
      ? "dev_application_catalog_http"
      : "offline",
    baseUrl: normalizeBaseUrl(
      env.VITE_RADISHMIND_APPLICATION_CATALOG_BASE_URL ??
      env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ??
      DEFAULT_BASE_URL,
    ),
    tenantRef: env.VITE_RADISHMIND_DEV_READ_TENANT_REF?.trim() || "tenant_demo",
    workspaceId: env.VITE_RADISHMIND_APPLICATION_CATALOG_WORKSPACE_ID?.trim() || "workspace_demo",
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || "subject_demo_user",
    authMode: normalizeAuthMode(env.VITE_RADISHMIND_READ_AUTH_MODE),
  };
}

export function validateApplicationCatalogMutableFields(fields: ApplicationCatalogMutableFields): string {
  const displayName = fields.displayName.trim();
  const description = fields.description.trim();
  if (displayName.length < 2 || displayName.length > 120 || description.length > 1000 || !APPLICATION_KINDS.includes(fields.applicationKind)) {
    return "application_catalog_payload_invalid";
  }
  if (containsSecretMaterial(displayName) || containsSecretMaterial(description)) {
    return "application_catalog_secret_material_forbidden";
  }
  return "";
}

export async function listApplicationCatalogRecords(
  config: ApplicationCatalogConfig,
  lifecycleState: ApplicationCatalogLifecycleState,
  cursor = "",
  applicationKind: ApplicationCatalogKind | "" = "",
): Promise<ApplicationCatalogListResult> {
  if (config.mode === "offline") {
    return offlineListResult();
  }
  const query = new URLSearchParams({
    workspace_id: config.workspaceId,
    lifecycle_state: lifecycleState,
    limit: "100",
  });
  if (cursor) query.set("cursor", cursor);
  if (applicationKind) query.set("application_kind", applicationKind);
  const requestId = createRequestId("application-catalog-list");
  try {
    const response = await fetch(`${config.baseUrl}${APPLICATION_CATALOG_COLLECTION_PATH}?${query}`, {
      headers: applicationCatalogHeaders(config, requestId, "read"),
    });
    const body: unknown = await response.json();
    if (!isApplicationCatalogListEnvelope(body, config, lifecycleState)) {
      return failedListResult("application_catalog_store_unavailable");
    }
    if (body.failure_code) {
      return failedListResult(body.failure_code, body.request_id, body.audit_ref);
    }
    const records = body.items.map(mapApplicationCatalogSummary);
    return {
      status: records.length ? "ready" : "empty",
      records,
      nextCursor: body.next_cursor ?? "",
      failureCode: "",
      requestId: body.request_id,
      auditRef: body.audit_ref,
      summary: records.length
        ? `Loaded ${records.length} ${lifecycleState} application catalog records.`
        : `No ${lifecycleState} applications exist in this workspace.`,
    };
  } catch {
    return failedListResult("application_catalog_store_unavailable");
  }
}

export async function createApplicationCatalogRecord(
  config: ApplicationCatalogConfig,
  fields: ApplicationCatalogMutableFields,
): Promise<ApplicationCatalogOperationResult> {
  if (config.mode === "offline") return offlineOperationResult();
  const failureCode = validateApplicationCatalogMutableFields(fields);
  if (failureCode) return localValidationFailure(failureCode);
  return writeApplicationCatalogRecord(config, APPLICATION_CATALOG_COLLECTION_PATH, "POST", {
    workspace_id: config.workspaceId,
    display_name: fields.displayName.trim(),
    description: fields.description.trim(),
    application_kind: fields.applicationKind,
  }, "create", "created");
}

export async function readApplicationCatalogRecord(
  config: ApplicationCatalogConfig,
  applicationId: string,
): Promise<ApplicationCatalogOperationResult> {
  if (config.mode === "offline") return offlineOperationResult();
  if (!APPLICATION_ID_PATTERN.test(applicationId)) return localValidationFailure("application_catalog_payload_invalid");
  const requestId = createRequestId("application-catalog-read");
  try {
    const response = await fetch(
      `${config.baseUrl}${APPLICATION_CATALOG_COLLECTION_PATH}/${encodeURIComponent(applicationId)}?workspace_id=${encodeURIComponent(config.workspaceId)}`,
      { headers: applicationCatalogHeaders(config, requestId, "read") },
    );
    const body: unknown = await response.json();
    return mapOperationEnvelope(body, config, "loaded");
  } catch {
    return failedOperationResult("application_catalog_store_unavailable");
  }
}

export async function updateApplicationCatalogRecord(
  config: ApplicationCatalogConfig,
  applicationId: string,
  expectedVersion: number,
  fields: ApplicationCatalogMutableFields,
): Promise<ApplicationCatalogOperationResult> {
  if (config.mode === "offline") return offlineOperationResult();
  const failureCode = validateApplicationCatalogMutableFields(fields);
  if (!APPLICATION_ID_PATTERN.test(applicationId) || !Number.isInteger(expectedVersion) || expectedVersion < 1 || failureCode) {
    return localValidationFailure(failureCode || "application_catalog_payload_invalid");
  }
  return writeApplicationCatalogRecord(
    config,
    `${APPLICATION_CATALOG_COLLECTION_PATH}/${encodeURIComponent(applicationId)}`,
    "PUT",
    {
      workspace_id: config.workspaceId,
      expected_version: expectedVersion,
      display_name: fields.displayName.trim(),
      description: fields.description.trim(),
      application_kind: fields.applicationKind,
    },
    "update",
    "updated",
  );
}

export async function archiveApplicationCatalogRecord(
  config: ApplicationCatalogConfig,
  applicationId: string,
  expectedVersion: number,
): Promise<ApplicationCatalogOperationResult> {
  if (config.mode === "offline") return offlineOperationResult();
  if (!APPLICATION_ID_PATTERN.test(applicationId) || !Number.isInteger(expectedVersion) || expectedVersion < 1) {
    return localValidationFailure("application_catalog_payload_invalid");
  }
  return writeApplicationCatalogRecord(
    config,
    `${APPLICATION_CATALOG_COLLECTION_PATH}/${encodeURIComponent(applicationId)}/archive`,
    "POST",
    { workspace_id: config.workspaceId, expected_version: expectedVersion },
    "archive",
    "archived",
  );
}

async function writeApplicationCatalogRecord(
  config: ApplicationCatalogConfig,
  path: string,
  method: "POST" | "PUT",
  body: unknown,
  operation: "create" | "update" | "archive",
  successStatus: "created" | "updated" | "archived",
): Promise<ApplicationCatalogOperationResult> {
  const requestId = createRequestId(`application-catalog-${operation}`);
  try {
    const response = await fetch(`${config.baseUrl}${path}`, {
      method,
      headers: { ...applicationCatalogHeaders(config, requestId, operation), "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    const document: unknown = await response.json();
    return mapOperationEnvelope(document, config, successStatus);
  } catch {
    return failedOperationResult("application_catalog_store_unavailable");
  }
}

function mapOperationEnvelope(
  value: unknown,
  config: ApplicationCatalogConfig,
  successStatus: "created" | "loaded" | "updated" | "archived",
): ApplicationCatalogOperationResult {
  if (!isApplicationCatalogOperationEnvelope(value, config)) {
    return failedOperationResult("application_catalog_store_unavailable");
  }
  const status = value.failure_code === "application_catalog_version_conflict"
    ? "version_conflict"
    : value.failure_code === "application_catalog_archived"
      ? "record_archived"
      : value.failure_code === "application_catalog_scope_denied" || value.failure_code === "workspace_membership_unavailable"
        ? "scope_denied"
        : value.failure_code
          ? "failed"
          : successStatus;
  return {
    status,
    record: value.record ? mapApplicationCatalogRecord(value.record) : null,
    failureCode: value.failure_code ?? "",
    currentRecordVersion: value.current_record_version,
    currentLifecycleState: value.current_lifecycle_state,
    requestId: value.request_id,
    auditRef: value.audit_ref,
    summary: value.failure_code
      ? status === "version_conflict"
        ? `Stored application changed to version ${value.current_record_version}; local edits were preserved.`
        : "Application catalog operation failed without changing stored state."
      : successStatus === "created"
        ? "Application created and selected from the authoritative catalog."
        : successStatus === "updated"
          ? "Application metadata updated with the expected record version."
          : successStatus === "archived"
            ? "Application archived; new configuration, publish, and invocation handoffs are disabled."
            : "Application catalog record loaded.",
  };
}

function isApplicationCatalogOperationEnvelope(
  value: unknown,
  config: ApplicationCatalogConfig,
): value is ApplicationCatalogOperationEnvelope {
  if (!isRecord(value) || !hasOnlyKeys(value, OPERATION_ENVELOPE_KEYS) || containsForbiddenResponse(value)) return false;
  if (!isNonEmptyString(value.request_id) || value.tenant_ref !== config.tenantRef || value.workspace_id !== config.workspaceId ||
    !(value.failure_code === null || isNonEmptyString(value.failure_code)) || typeof value.current_record_version !== "number" || !Number.isInteger(value.current_record_version) ||
    value.current_record_version < 0 || typeof value.current_lifecycle_state !== "string" || !isNonEmptyString(value.audit_ref)) return false;
  if (value.record === null) return value.failure_code !== null;
  return isApplicationCatalogRecordDocument(value.record, config) && value.failure_code === null &&
    value.current_record_version === value.record.record_version && value.current_lifecycle_state === value.record.lifecycle_state;
}

function isApplicationCatalogListEnvelope(
  value: unknown,
  config: ApplicationCatalogConfig,
  lifecycleState: ApplicationCatalogLifecycleState,
): value is ApplicationCatalogListEnvelope {
  return isRecord(value) && hasOnlyKeys(value, LIST_ENVELOPE_KEYS) && !containsForbiddenResponse(value) &&
    isNonEmptyString(value.request_id) && value.tenant_ref === config.tenantRef && Array.isArray(value.items) &&
    value.items.every((item) => isApplicationCatalogSummaryDocument(item, config, lifecycleState)) &&
    (value.next_cursor === null || isNonEmptyString(value.next_cursor)) &&
    (value.failure_code === null || isNonEmptyString(value.failure_code)) && isNonEmptyString(value.audit_ref);
}

function isApplicationCatalogRecordDocument(value: unknown, config: ApplicationCatalogConfig): value is ApplicationCatalogRecordDocument {
  return isRecord(value) && hasOnlyKeys(value, RECORD_KEYS) && value.schema_version === APPLICATION_CATALOG_SCHEMA_VERSION &&
    APPLICATION_ID_PATTERN.test(String(value.application_id)) && value.tenant_ref === config.tenantRef &&
    value.workspace_id === config.workspaceId && value.owner_subject_ref === config.subjectRef &&
    isApplicationCatalogMutableDocument(value) && isLifecycleState(value.lifecycle_state) &&
    typeof value.record_version === "number" && Number.isInteger(value.record_version) && value.record_version > 0 && isTimestamp(value.created_at) &&
    isTimestamp(value.updated_at) && isArchivedAt(value.archived_at, value.lifecycle_state) &&
    isScopeIdentifier(value.created_by_actor_ref) && isScopeIdentifier(value.updated_by_actor_ref) &&
    isNonEmptyString(value.request_id) && isNonEmptyString(value.audit_ref);
}

function isApplicationCatalogSummaryDocument(
  value: unknown,
  config: ApplicationCatalogConfig,
  lifecycleState: ApplicationCatalogLifecycleState,
): value is ApplicationCatalogSummaryDocument {
  return isRecord(value) && hasOnlyKeys(value, SUMMARY_KEYS) && APPLICATION_ID_PATTERN.test(String(value.application_ref)) &&
    value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId && value.owner_subject_ref === config.subjectRef &&
    isApplicationCatalogMutableDocument(value) && value.lifecycle_state === lifecycleState && typeof value.record_version === "number" && Number.isInteger(value.record_version) &&
    value.record_version > 0 && isTimestamp(value.created_at) && isTimestamp(value.updated_at) &&
    isArchivedAt(value.archived_at, value.lifecycle_state) && value.latest_workflow_definition_ref === "" &&
    value.last_run_status === "not_available";
}

function isApplicationCatalogMutableDocument(value: Record<string, unknown>): boolean {
  return typeof value.display_name === "string" && typeof value.description === "string" &&
    isApplicationKind(value.application_kind) && validateApplicationCatalogMutableFields({
      displayName: value.display_name,
      description: value.description,
      applicationKind: value.application_kind,
    }) === "";
}

function mapApplicationCatalogRecord(document: ApplicationCatalogRecordDocument): ApplicationCatalogRecord {
  return {
    schemaVersion: APPLICATION_CATALOG_SCHEMA_VERSION,
    applicationId: document.application_id,
    tenantRef: document.tenant_ref,
    workspaceId: document.workspace_id,
    ownerSubjectRef: document.owner_subject_ref,
    displayName: document.display_name,
    description: document.description,
    applicationKind: document.application_kind as ApplicationCatalogKind,
    lifecycleState: document.lifecycle_state as ApplicationCatalogLifecycleState,
    recordVersion: document.record_version,
    createdAt: document.created_at,
    updatedAt: document.updated_at,
    archivedAt: document.archived_at,
    createdByActorRef: document.created_by_actor_ref,
    updatedByActorRef: document.updated_by_actor_ref,
    requestId: document.request_id,
    auditRef: document.audit_ref,
  };
}

function mapApplicationCatalogSummary(document: ApplicationCatalogSummaryDocument): ApplicationCatalogRecord {
  return {
    schemaVersion: APPLICATION_CATALOG_SCHEMA_VERSION,
    applicationId: document.application_ref,
    tenantRef: document.tenant_ref,
    workspaceId: document.workspace_id,
    ownerSubjectRef: document.owner_subject_ref,
    displayName: document.display_name,
    description: document.description,
    applicationKind: document.application_kind as ApplicationCatalogKind,
    lifecycleState: document.lifecycle_state as ApplicationCatalogLifecycleState,
    recordVersion: document.record_version,
    createdAt: document.created_at,
    updatedAt: document.updated_at,
    archivedAt: document.archived_at,
    createdByActorRef: document.owner_subject_ref,
    updatedByActorRef: document.owner_subject_ref,
    requestId: "",
    auditRef: "",
  };
}

function applicationCatalogHeaders(
  config: ApplicationCatalogConfig,
  requestId: string,
  operation: "read" | "create" | "update" | "archive",
): Record<string, string> {
  if (config.authMode !== "dev_headers") {
    const tokenProvider = config.authMode === "signed_test_token"
      ? (globalThis as typeof globalThis & { __RADISHMIND_CONTROL_PLANE_SIGNED_TEST_TOKEN__?: () => string })
        .__RADISHMIND_CONTROL_PLANE_SIGNED_TEST_TOKEN__
      : (globalThis as typeof globalThis & { __RADISHMIND_CONTROL_PLANE_OIDC_INTEGRATION_TOKEN__?: () => string })
        .__RADISHMIND_CONTROL_PLANE_OIDC_INTEGRATION_TOKEN__;
    const token = tokenProvider?.().trim() ?? "";
    if (!token) throw new Error("application catalog auth token is unavailable in browser memory");
    return { Accept: "application/json", "X-Request-Id": requestId, Authorization: `Bearer ${token}` };
  }
  const scope = operation === "archive" ? "applications:archive,applications:read" :
    operation === "read" ? "applications:read" : "applications:write,applications:read";
  return {
    Accept: "application/json",
    "X-Request-Id": requestId,
    "X-RadishMind-Dev-Read-Identity": "radishmind-web-application-catalog-dev",
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Read-Scopes": scope,
    "X-RadishMind-Dev-Read-Audit": `audit-${requestId}`,
  };
}

function offlineListResult(): ApplicationCatalogListResult {
  return {
    status: "offline", records: [], nextCursor: "", failureCode: "application_catalog_http_disabled",
    requestId: "", auditRef: "", summary: "Offline fixture mode sends no application catalog requests and does not simulate writes.",
  };
}

function failedListResult(failureCode: string, requestId = "", auditRef = ""): ApplicationCatalogListResult {
  return {
    status: "failed", records: [], nextCursor: "", failureCode, requestId, auditRef,
    summary: "Application catalog records could not be loaded; no fixture or memory fallback was used.",
  };
}

function offlineOperationResult(): ApplicationCatalogOperationResult {
  return {
    status: "offline", record: null, failureCode: "application_catalog_http_disabled", currentRecordVersion: 0,
    currentLifecycleState: "", requestId: "", auditRef: "",
    summary: "Offline fixture mode is read-only and sends no application catalog write requests.",
  };
}

function localValidationFailure(failureCode: string): ApplicationCatalogOperationResult {
  return {
    status: "failed", record: null, failureCode, currentRecordVersion: 0, currentLifecycleState: "",
    requestId: "", auditRef: "", summary: "Application metadata was rejected before any request was sent.",
  };
}

function failedOperationResult(failureCode: string): ApplicationCatalogOperationResult {
  return {
    status: "failed", record: null, failureCode, currentRecordVersion: 0, currentLifecycleState: "",
    requestId: "", auditRef: "", summary: "Application catalog operation failed without a trusted response.",
  };
}

function normalizeBaseUrl(value: string): string {
  const normalized = value.trim() || DEFAULT_BASE_URL;
  return normalized.endsWith("/") ? normalized.slice(0, -1) : normalized;
}

function normalizeAuthMode(value: string | undefined): ApplicationCatalogAuthMode {
  const normalized = value?.trim();
  return normalized === "signed_test_token" || normalized === "radish_oidc_integration_test" ? normalized : "dev_headers";
}

function createRequestId(prefix: string): string {
  return `${prefix}-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
}

function containsForbiddenResponse(value: unknown): boolean {
  if (typeof value === "string") return containsSecretMaterial(value);
  if (Array.isArray(value)) return value.some(containsForbiddenResponse);
  if (!isRecord(value)) return false;
  return Object.entries(value).some(([key, nested]) => FORBIDDEN_RESPONSE_FIELDS.has(key.toLowerCase()) || containsForbiddenResponse(nested));
}

function containsSecretMaterial(value: string): boolean {
  return /authorization:|bearer\s|api[_-]?key\s*[:=]|x-radishmind-dev-|sk-[a-z0-9]|postgres(?:ql)?:\/\//iu.test(value.trim());
}

function hasOnlyKeys(value: Record<string, unknown>, allowed: readonly string[]): boolean {
  const allowedSet = new Set(allowed);
  return Object.keys(value).length === allowed.length && Object.keys(value).every((key) => allowedSet.has(key));
}

function isApplicationKind(value: unknown): value is ApplicationCatalogKind {
  return typeof value === "string" && APPLICATION_KINDS.includes(value as ApplicationCatalogKind);
}

function isLifecycleState(value: unknown): value is ApplicationCatalogLifecycleState {
  return value === "active" || value === "archived";
}

function isScopeIdentifier(value: unknown): value is string {
  return typeof value === "string" && SCOPE_ID_PATTERN.test(value);
}

function isTimestamp(value: unknown): value is string {
  return typeof value === "string" && value.length >= 20 && Number.isFinite(Date.parse(value));
}

function isArchivedAt(value: unknown, lifecycleState: unknown): boolean {
  return lifecycleState === "active" ? value === null : isTimestamp(value);
}

function isNonEmptyString(value: unknown): value is string {
  return typeof value === "string" && value.trim().length > 0;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
