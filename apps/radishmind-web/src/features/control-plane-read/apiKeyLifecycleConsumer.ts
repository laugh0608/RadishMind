import { CONTROL_PLANE_READ_ROUTES } from "../../../../../contracts/typescript/control-plane-read-api.ts";

const API_KEY_COLLECTION_PATH = CONTROL_PLANE_READ_ROUTES.apiKeys;
const API_KEY_RECORD_SCHEMA_VERSION = "api_key_record.v1";
const DEV_SOURCE = "dev-api-key-lifecycle-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const API_KEY_ID_PATTERN = /^key_[a-z2-7]{16}$/u;
const APPLICATION_ID_PATTERN = /^app_[a-z0-9]{16}$/u;
const TOKEN_PATTERN = /^rmd_dev_key_[a-z2-7]{16}\.[A-Za-z0-9_-]{43}$/u;
const SCOPE_ID_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._:-]{2,127}$/u;
const ALLOWED_SCOPES = ["models:read", "chat:invoke", "responses:invoke", "messages:invoke"] as const;
const EFFECTIVE_STATES = ["active", "expired", "revoked"] as const;
const RECORD_KEYS = [
  "schema_version", "api_key_id", "tenant_ref", "workspace_id", "application_id", "owner_subject_ref",
  "display_name", "scopes", "lifecycle_state", "effective_state", "record_version", "created_at",
  "expires_at", "last_used_at", "revoked_at", "created_by_actor_ref", "revoked_by_actor_ref",
  "request_id", "audit_ref",
] as const;
const SUMMARY_KEYS = [
  "api_key_id", "tenant_ref", "workspace_id", "application_id", "owner_subject_ref", "display_name",
  "scopes", "state", "lifecycle_state", "effective_state", "record_version", "created_at", "expires_at",
  "last_used_at", "revoked_at",
] as const;
const OPERATION_KEYS = [
  "request_id", "tenant_ref", "workspace_id", "record", "failure_code", "current_record_version",
  "current_effective_state", "audit_ref",
] as const;
const ISSUE_OPERATION_KEYS = [...OPERATION_KEYS, "credential"] as const;
const LIST_KEYS = ["request_id", "tenant_ref", "items", "next_cursor", "failure_code", "audit_ref"] as const;
const CREDENTIAL_KEYS = ["token"] as const;
const FORBIDDEN_RESPONSE_FIELDS = new Set([
  "authorization", "credential_digest", "key_hash", "secret", "headers", "cookie", "dsn", "endpoint",
  "base_url", "raw_request", "raw_response", "input", "output", "prompt", "messages", "stack_trace",
]);

export type APIKeyLifecycleMode = "offline" | "dev_api_key_lifecycle_http";
export type APIKeyLifecycleAuthMode = "dev_headers" | "signed_test_token" | "radish_oidc_integration_test";
export type APIKeyScope = typeof ALLOWED_SCOPES[number];
export type APIKeyEffectiveState = typeof EFFECTIVE_STATES[number];

export type APIKeyLifecycleConfig = {
  mode: APIKeyLifecycleMode;
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  subjectRef: string;
  authMode: APIKeyLifecycleAuthMode;
};

export type APIKeyRecord = {
  schemaVersion: typeof API_KEY_RECORD_SCHEMA_VERSION;
  apiKeyId: string;
  tenantRef: string;
  workspaceId: string;
  applicationId: string;
  ownerSubjectRef: string;
  displayName: string;
  scopes: APIKeyScope[];
  lifecycleState: "active" | "revoked";
  effectiveState: APIKeyEffectiveState;
  recordVersion: number;
  createdAt: string;
  expiresAt: string;
  lastUsedAt: string | null;
  revokedAt: string | null;
  createdByActorRef: string;
  revokedByActorRef: string | null;
  requestId: string;
  auditRef: string;
};

export type APIKeyIssueInput = {
  applicationId: string;
  displayName: string;
  scopes: APIKeyScope[];
  expiresInDays: number;
};

export type APIKeyListResult = {
  status: "offline" | "ready" | "empty" | "failed";
  records: APIKeyRecord[];
  nextCursor: string;
  failureCode: string;
  requestId: string;
  auditRef: string;
  summary: string;
};

export type APIKeyOperationResult = {
  status: "offline" | "issued" | "loaded" | "revoked" | "version_conflict" | "scope_denied" | "failed";
  record: APIKeyRecord | null;
  credentialToken: string;
  failureCode: string;
  currentRecordVersion: number;
  currentEffectiveState: string;
  requestId: string;
  auditRef: string;
  summary: string;
};

type APIKeyRecordDocument = {
  schema_version: string;
  api_key_id: string;
  tenant_ref: string;
  workspace_id: string;
  application_id: string;
  owner_subject_ref: string;
  display_name: string;
  scopes: string[];
  lifecycle_state: string;
  effective_state: string;
  record_version: number;
  created_at: string;
  expires_at: string;
  last_used_at: string | null;
  revoked_at: string | null;
  created_by_actor_ref: string;
  revoked_by_actor_ref: string | null;
  request_id: string;
  audit_ref: string;
};

type APIKeySummaryDocument = {
  api_key_id: string;
  tenant_ref: string;
  workspace_id: string;
  application_id: string;
  owner_subject_ref: string;
  display_name: string;
  scopes: string[];
  state: string;
  lifecycle_state: string;
  effective_state: string;
  record_version: number;
  created_at: string;
  expires_at: string;
  last_used_at: string | null;
  revoked_at: string | null;
};

type APIKeyOperationEnvelope = {
  request_id: string;
  tenant_ref: string;
  workspace_id: string;
  record: APIKeyRecordDocument | null;
  credential?: { token: string };
  failure_code: string | null;
  current_record_version: number;
  current_effective_state: string;
  audit_ref: string;
};

type APIKeyListEnvelope = {
  request_id: string;
  tenant_ref: string;
  items: APIKeySummaryDocument[];
  next_cursor: string | null;
  failure_code: string | null;
  audit_ref: string;
};

export function readAPIKeyLifecycleConfig(): APIKeyLifecycleConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  return {
    mode: env.VITE_RADISHMIND_API_KEY_LIFECYCLE_SOURCE?.trim() === DEV_SOURCE
      ? "dev_api_key_lifecycle_http"
      : "offline",
    baseUrl: normalizeBaseUrl(
      env.VITE_RADISHMIND_API_KEY_LIFECYCLE_BASE_URL ??
      env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ??
      DEFAULT_BASE_URL,
    ),
    tenantRef: env.VITE_RADISHMIND_DEV_READ_TENANT_REF?.trim() || "tenant_demo",
    workspaceId: env.VITE_RADISHMIND_API_KEY_LIFECYCLE_WORKSPACE_ID?.trim() || "workspace_demo",
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || "subject_demo_user",
    authMode: normalizeAuthMode(env.VITE_RADISHMIND_READ_AUTH_MODE),
  };
}

export function validateAPIKeyIssueInput(input: APIKeyIssueInput): string {
  const displayName = input.displayName.trim();
  if (!APPLICATION_ID_PATTERN.test(input.applicationId) || displayName.length < 2 || displayName.length > 80 ||
    !Number.isInteger(input.expiresInDays) || input.expiresInDays < 1 || input.expiresInDays > 90 ||
    input.scopes.length === 0 || input.scopes.some((scope) => !ALLOWED_SCOPES.includes(scope)) ||
    new Set(input.scopes).size !== input.scopes.length) {
    return "api_key_payload_invalid";
  }
  if (containsSensitiveText(displayName)) return "api_key_secret_material_forbidden";
  return "";
}

export async function listAPIKeyRecords(
  config: APIKeyLifecycleConfig,
  applicationId: string,
  cursor = "",
  effectiveState: APIKeyEffectiveState | "" = "",
): Promise<APIKeyListResult> {
  if (config.mode === "offline") return offlineListResult();
  if (!APPLICATION_ID_PATTERN.test(applicationId) || (effectiveState && !EFFECTIVE_STATES.includes(effectiveState))) {
    return failedListResult("api_key_payload_invalid");
  }
  const query = new URLSearchParams({ workspace_id: config.workspaceId, application_id: applicationId, limit: "100" });
  if (cursor) query.set("cursor", cursor);
  if (effectiveState) query.set("effective_state", effectiveState);
  const requestId = createRequestId("api-key-list");
  try {
    const response = await fetch(`${config.baseUrl}${API_KEY_COLLECTION_PATH}?${query}`, {
      headers: apiKeyManagementHeaders(config, requestId, "read"),
    });
    const document: unknown = await response.json();
    if (!isAPIKeyListEnvelope(document, config, applicationId)) return failedListResult("api_key_store_unavailable");
    if (document.failure_code) return failedListResult(document.failure_code, document.request_id, document.audit_ref);
    const records = document.items.map(mapAPIKeySummary);
    return {
      status: records.length ? "ready" : "empty",
      records,
      nextCursor: document.next_cursor ?? "",
      failureCode: "",
      requestId: document.request_id,
      auditRef: document.audit_ref,
      summary: records.length ? `Loaded ${records.length} API key records for the selected application.` : "No API keys exist for the selected application.",
    };
  } catch {
    return failedListResult("api_key_store_unavailable");
  }
}

export async function issueAPIKey(
  config: APIKeyLifecycleConfig,
  input: APIKeyIssueInput,
): Promise<APIKeyOperationResult> {
  if (config.mode === "offline") return offlineOperationResult();
  const failureCode = validateAPIKeyIssueInput(input);
  if (failureCode) return localValidationFailure(failureCode);
  const requestId = createRequestId("api-key-issue");
  try {
    const response = await fetch(`${config.baseUrl}${API_KEY_COLLECTION_PATH}`, {
      method: "POST",
      headers: { ...apiKeyManagementHeaders(config, requestId, "issue"), "Content-Type": "application/json" },
      body: JSON.stringify({
        workspace_id: config.workspaceId,
        application_id: input.applicationId,
        display_name: input.displayName.trim(),
        scopes: [...input.scopes].sort(),
        expires_in_days: input.expiresInDays,
      }),
    });
    const document: unknown = await response.json();
    const cacheControl = response.headers.get("Cache-Control")?.toLowerCase() ?? "";
    return mapOperationEnvelope(document, config, "issued", true, cacheControl.includes("no-store"));
  } catch {
    return failedOperationResult("api_key_store_unavailable");
  }
}

export async function readAPIKeyRecord(
  config: APIKeyLifecycleConfig,
  apiKeyId: string,
): Promise<APIKeyOperationResult> {
  if (config.mode === "offline") return offlineOperationResult();
  if (!API_KEY_ID_PATTERN.test(apiKeyId)) return localValidationFailure("api_key_payload_invalid");
  const requestId = createRequestId("api-key-read");
  try {
    const response = await fetch(
      `${config.baseUrl}${API_KEY_COLLECTION_PATH}/${encodeURIComponent(apiKeyId)}?workspace_id=${encodeURIComponent(config.workspaceId)}`,
      { headers: apiKeyManagementHeaders(config, requestId, "read") },
    );
    return mapOperationEnvelope(await response.json(), config, "loaded", false, false);
  } catch {
    return failedOperationResult("api_key_store_unavailable");
  }
}

export async function revokeAPIKey(
  config: APIKeyLifecycleConfig,
  apiKeyId: string,
  expectedVersion: number,
): Promise<APIKeyOperationResult> {
  if (config.mode === "offline") return offlineOperationResult();
  if (!API_KEY_ID_PATTERN.test(apiKeyId) || !Number.isInteger(expectedVersion) || expectedVersion < 1) {
    return localValidationFailure("api_key_payload_invalid");
  }
  const requestId = createRequestId("api-key-revoke");
  try {
    const response = await fetch(`${config.baseUrl}${API_KEY_COLLECTION_PATH}/${encodeURIComponent(apiKeyId)}/revoke`, {
      method: "POST",
      headers: { ...apiKeyManagementHeaders(config, requestId, "revoke"), "Content-Type": "application/json" },
      body: JSON.stringify({ workspace_id: config.workspaceId, expected_version: expectedVersion }),
    });
    return mapOperationEnvelope(await response.json(), config, "revoked", false, false);
  } catch {
    return failedOperationResult("api_key_store_unavailable");
  }
}

function mapOperationEnvelope(
  value: unknown,
  config: APIKeyLifecycleConfig,
  successStatus: "issued" | "loaded" | "revoked",
  issueResponse: boolean,
  noStore: boolean,
): APIKeyOperationResult {
  if (!isAPIKeyOperationEnvelope(value, config, issueResponse, noStore)) {
    return failedOperationResult("api_key_store_unavailable");
  }
  const status = value.failure_code === "api_key_version_conflict"
    ? "version_conflict"
    : value.failure_code === "api_key_scope_denied" || value.failure_code === "workspace_membership_unavailable"
      ? "scope_denied"
      : value.failure_code
        ? "failed"
        : successStatus;
  return {
    status,
    record: value.record ? mapAPIKeyRecord(value.record) : null,
    credentialToken: issueResponse && !value.failure_code ? value.credential?.token ?? "" : "",
    failureCode: value.failure_code ?? "",
    currentRecordVersion: value.current_record_version,
    currentEffectiveState: value.current_effective_state,
    requestId: value.request_id,
    auditRef: value.audit_ref,
    summary: value.failure_code
      ? status === "version_conflict"
        ? `Stored API key changed to version ${value.current_record_version}; refresh before retrying the revoke operation.`
        : "API key operation failed without changing stored state."
      : successStatus === "issued"
        ? "API key issued. The credential is available only in this response and current component memory."
        : successStatus === "revoked"
          ? "API key revoked with the expected record version."
          : "API key detail loaded without credential material.",
  };
}

function isAPIKeyOperationEnvelope(
  value: unknown,
  config: APIKeyLifecycleConfig,
  issueResponse: boolean,
  noStore: boolean,
): value is APIKeyOperationEnvelope {
  const allowedKeys = issueResponse ? ISSUE_OPERATION_KEYS : OPERATION_KEYS;
  if (!isRecord(value) || !hasOnlyKeys(value, allowedKeys) || containsForbiddenResponse(value)) return false;
  if (!isNonEmptyString(value.request_id) || value.tenant_ref !== config.tenantRef || value.workspace_id !== config.workspaceId ||
    !(value.failure_code === null || isNonEmptyString(value.failure_code)) || !isPositiveOrZeroInteger(value.current_record_version) ||
    typeof value.current_effective_state !== "string" || !isNonEmptyString(value.audit_ref)) return false;
  if (value.record === null) {
    return value.failure_code !== null && value.credential === undefined;
  }
  if (value.failure_code !== null || !isAPIKeyRecordDocument(value.record, config) ||
    value.current_record_version !== value.record.record_version || value.current_effective_state !== value.record.effective_state) return false;
  if (!issueResponse) return value.credential === undefined;
  return noStore && isRecord(value.credential) && hasOnlyKeys(value.credential, CREDENTIAL_KEYS) &&
    TOKEN_PATTERN.test(String(value.credential.token)) && value.credential.token.includes(value.record.api_key_id);
}

function isAPIKeyListEnvelope(
  value: unknown,
  config: APIKeyLifecycleConfig,
  applicationId: string,
): value is APIKeyListEnvelope {
  return isRecord(value) && hasOnlyKeys(value, LIST_KEYS) && !containsForbiddenResponse(value) &&
    isNonEmptyString(value.request_id) && value.tenant_ref === config.tenantRef && Array.isArray(value.items) &&
    value.items.every((item) => isAPIKeySummaryDocument(item, config, applicationId)) &&
    (value.next_cursor === null || isNonEmptyString(value.next_cursor)) &&
    (value.failure_code === null || isNonEmptyString(value.failure_code)) && isNonEmptyString(value.audit_ref);
}

function isAPIKeyRecordDocument(value: unknown, config: APIKeyLifecycleConfig): value is APIKeyRecordDocument {
  return isRecord(value) && hasOnlyKeys(value, RECORD_KEYS) && value.schema_version === API_KEY_RECORD_SCHEMA_VERSION &&
    isAPIKeySharedDocument(value, config) && isScopeIdentifier(value.created_by_actor_ref) &&
    (value.revoked_by_actor_ref === null || isScopeIdentifier(value.revoked_by_actor_ref)) &&
    isNonEmptyString(value.request_id) && isNonEmptyString(value.audit_ref);
}

function isAPIKeySummaryDocument(
  value: unknown,
  config: APIKeyLifecycleConfig,
  applicationId: string,
): value is APIKeySummaryDocument {
  return isRecord(value) && hasOnlyKeys(value, SUMMARY_KEYS) && isAPIKeySharedDocument(value, config) &&
    value.application_id === applicationId && value.state === value.effective_state;
}

function isAPIKeySharedDocument(value: Record<string, unknown>, config: APIKeyLifecycleConfig): boolean {
  const scopes = value.scopes;
  const lifecycleState = value.lifecycle_state;
  const effectiveState = value.effective_state;
  const revokedAt = value.revoked_at;
  return API_KEY_ID_PATTERN.test(String(value.api_key_id)) && value.tenant_ref === config.tenantRef &&
    value.workspace_id === config.workspaceId && APPLICATION_ID_PATTERN.test(String(value.application_id)) &&
    value.owner_subject_ref === config.subjectRef && typeof value.display_name === "string" &&
    value.display_name.trim().length >= 2 && value.display_name.trim().length <= 80 && !containsSensitiveText(value.display_name) &&
    Array.isArray(scopes) && scopes.length > 0 && scopes.every((scope) => typeof scope === "string" && ALLOWED_SCOPES.includes(scope as APIKeyScope)) &&
    new Set(scopes).size === scopes.length && (lifecycleState === "active" || lifecycleState === "revoked") &&
    typeof effectiveState === "string" && EFFECTIVE_STATES.includes(effectiveState as APIKeyEffectiveState) &&
    isPositiveInteger(value.record_version) && isTimestamp(value.created_at) && isTimestamp(value.expires_at) &&
    isNullableTimestamp(value.last_used_at) && isNullableTimestamp(revokedAt) &&
    ((lifecycleState === "revoked" && effectiveState === "revoked" && revokedAt !== null) ||
      (lifecycleState === "active" && effectiveState !== "revoked" && revokedAt === null));
}

function mapAPIKeyRecord(document: APIKeyRecordDocument): APIKeyRecord {
  return {
    schemaVersion: API_KEY_RECORD_SCHEMA_VERSION,
    apiKeyId: document.api_key_id,
    tenantRef: document.tenant_ref,
    workspaceId: document.workspace_id,
    applicationId: document.application_id,
    ownerSubjectRef: document.owner_subject_ref,
    displayName: document.display_name,
    scopes: document.scopes as APIKeyScope[],
    lifecycleState: document.lifecycle_state as APIKeyRecord["lifecycleState"],
    effectiveState: document.effective_state as APIKeyEffectiveState,
    recordVersion: document.record_version,
    createdAt: document.created_at,
    expiresAt: document.expires_at,
    lastUsedAt: document.last_used_at,
    revokedAt: document.revoked_at,
    createdByActorRef: document.created_by_actor_ref,
    revokedByActorRef: document.revoked_by_actor_ref,
    requestId: document.request_id,
    auditRef: document.audit_ref,
  };
}

function mapAPIKeySummary(document: APIKeySummaryDocument): APIKeyRecord {
  return {
    schemaVersion: API_KEY_RECORD_SCHEMA_VERSION,
    apiKeyId: document.api_key_id,
    tenantRef: document.tenant_ref,
    workspaceId: document.workspace_id,
    applicationId: document.application_id,
    ownerSubjectRef: document.owner_subject_ref,
    displayName: document.display_name,
    scopes: document.scopes as APIKeyScope[],
    lifecycleState: document.lifecycle_state as APIKeyRecord["lifecycleState"],
    effectiveState: document.effective_state as APIKeyEffectiveState,
    recordVersion: document.record_version,
    createdAt: document.created_at,
    expiresAt: document.expires_at,
    lastUsedAt: document.last_used_at,
    revokedAt: document.revoked_at,
    createdByActorRef: document.owner_subject_ref,
    revokedByActorRef: document.lifecycle_state === "revoked" ? document.owner_subject_ref : null,
    requestId: "",
    auditRef: "",
  };
}

function apiKeyManagementHeaders(
  config: APIKeyLifecycleConfig,
  requestId: string,
  operation: "read" | "issue" | "revoke",
): Record<string, string> {
  if (config.authMode !== "dev_headers") {
    const tokenProvider = config.authMode === "signed_test_token"
      ? (globalThis as typeof globalThis & { __RADISHMIND_CONTROL_PLANE_SIGNED_TEST_TOKEN__?: () => string })
        .__RADISHMIND_CONTROL_PLANE_SIGNED_TEST_TOKEN__
      : (globalThis as typeof globalThis & { __RADISHMIND_CONTROL_PLANE_OIDC_INTEGRATION_TOKEN__?: () => string })
        .__RADISHMIND_CONTROL_PLANE_OIDC_INTEGRATION_TOKEN__;
    const token = tokenProvider?.().trim() ?? "";
    if (!token) throw new Error("API key lifecycle auth token is unavailable in browser memory");
    return { Accept: "application/json", "X-Request-Id": requestId, Authorization: `Bearer ${token}` };
  }
  const scopes = operation === "revoke" ? "api_keys:revoke,api_keys:read" :
    operation === "issue" ? "api_keys:write,api_keys:read" : "api_keys:read";
  return {
    Accept: "application/json",
    "X-Request-Id": requestId,
    "X-RadishMind-Dev-Read-Identity": "radishmind-web-api-key-lifecycle-dev",
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Read-Scopes": scopes,
    "X-RadishMind-Dev-Read-Audit": `audit-${requestId}`,
  };
}

function offlineListResult(): APIKeyListResult {
  return {
    status: "offline", records: [], nextCursor: "", failureCode: "api_key_lifecycle_http_disabled",
    requestId: "", auditRef: "", summary: "Offline mode sends no API key lifecycle requests and does not simulate credentials.",
  };
}

function failedListResult(failureCode: string, requestId = "", auditRef = ""): APIKeyListResult {
  return {
    status: "failed", records: [], nextCursor: "", failureCode, requestId, auditRef,
    summary: "API key records could not be loaded; no fixture or memory fallback was used.",
  };
}

function offlineOperationResult(): APIKeyOperationResult {
  return {
    status: "offline", record: null, credentialToken: "", failureCode: "api_key_lifecycle_http_disabled",
    currentRecordVersion: 0, currentEffectiveState: "", requestId: "", auditRef: "",
    summary: "Offline mode sends no API key lifecycle write requests.",
  };
}

function localValidationFailure(failureCode: string): APIKeyOperationResult {
  return {
    status: "failed", record: null, credentialToken: "", failureCode, currentRecordVersion: 0,
    currentEffectiveState: "", requestId: "", auditRef: "",
    summary: "API key input was rejected before any request was sent.",
  };
}

function failedOperationResult(failureCode: string): APIKeyOperationResult {
  return {
    status: "failed", record: null, credentialToken: "", failureCode, currentRecordVersion: 0,
    currentEffectiveState: "", requestId: "", auditRef: "",
    summary: "API key operation failed without a trusted response.",
  };
}

function normalizeBaseUrl(value: string): string {
  const normalized = value.trim() || DEFAULT_BASE_URL;
  return normalized.endsWith("/") ? normalized.slice(0, -1) : normalized;
}

function normalizeAuthMode(value: string | undefined): APIKeyLifecycleAuthMode {
  const normalized = value?.trim();
  return normalized === "signed_test_token" || normalized === "radish_oidc_integration_test" ? normalized : "dev_headers";
}

function createRequestId(prefix: string): string {
  const randomPart = globalThis.crypto?.randomUUID?.().replaceAll("-", "") ?? Math.random().toString(16).slice(2);
  return `${prefix}-${Date.now()}-${randomPart.slice(0, 16)}`;
}

function containsForbiddenResponse(value: unknown): boolean {
  if (Array.isArray(value)) return value.some(containsForbiddenResponse);
  if (!isRecord(value)) return false;
  return Object.entries(value).some(([key, nested]) => FORBIDDEN_RESPONSE_FIELDS.has(key.toLowerCase()) || containsForbiddenResponse(nested));
}

function containsSensitiveText(value: string): boolean {
  return /(authorization\s*:|bearer\s+|rmd_dev_|api[_ -]?key\s*=|postgres(?:ql)?:\/\/|password\s*=|secret\s*=)/iu.test(value);
}

function hasOnlyKeys(value: Record<string, unknown>, allowed: readonly string[]): boolean {
  const allowedSet = new Set(allowed);
  return Object.keys(value).every((key) => allowedSet.has(key)) && allowed.every((key) => key in value || key === "credential");
}

function isTimestamp(value: unknown): value is string {
  return typeof value === "string" && value.length > 0 && Number.isFinite(Date.parse(value));
}

function isNullableTimestamp(value: unknown): value is string | null {
  return value === null || isTimestamp(value);
}

function isPositiveInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value > 0;
}

function isPositiveOrZeroInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value >= 0;
}

function isScopeIdentifier(value: unknown): value is string {
  return typeof value === "string" && SCOPE_ID_PATTERN.test(value);
}

function isNonEmptyString(value: unknown): value is string {
  return typeof value === "string" && value.trim().length > 0;
}

function isRecord(value: unknown): value is Record<string, any> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}
