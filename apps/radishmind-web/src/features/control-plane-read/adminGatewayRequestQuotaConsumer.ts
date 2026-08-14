const DEV_SOURCE = "dev-admin-gateway-request-quota-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const IDENTIFIER = /^[A-Za-z0-9][A-Za-z0-9._:-]{0,255}$/u;
const WORKSPACE_IDENTIFIER = /^[A-Za-z0-9][A-Za-z0-9_.:/-]{0,159}$/u;
const UTC_DATE = /^\d{4}-\d{2}-\d{2}$/u;
const SCHEMA_VERSION = "gateway_request_quota_v1";
const PERIOD = "calendar_day_utc";
const MINIMUM_REQUEST_LIMIT = 1;
const MAXIMUM_REQUEST_LIMIT = 1_000_000;

const ENVELOPE_KEYS = [
  "request_id",
  "tenant_ref",
  "workspace_id",
  "environment",
  "application_id",
  "policy",
  "usage",
  "failure_code",
  "audit_ref",
] as const;

const POLICY_KEYS = [
  "schema_version",
  "policy_id",
  "tenant_ref",
  "workspace_id",
  "environment",
  "application_id",
  "period",
  "request_limit",
  "record_version",
  "created_at",
  "updated_at",
  "created_by",
  "updated_by",
  "last_request_id",
  "last_audit_ref",
] as const;

const USAGE_KEYS = [
  "schema_version",
  "tenant_ref",
  "workspace_id",
  "environment",
  "application_id",
  "period",
  "period_start",
  "policy_id",
  "policy_version",
  "request_limit",
  "admitted_request_count",
  "remaining_request_count",
  "updated_at",
] as const;

const FORBIDDEN_RESPONSE_FIELDS = new Set([
  "authorization",
  "api_key",
  "api_key_id",
  "credential",
  "credential_digest",
  "secret",
  "endpoint",
  "base_url",
  "dsn",
  "headers",
  "cookie",
  "raw_request",
  "raw_response",
  "provider_raw_response",
  "prompt",
  "messages",
  "input",
  "output",
  "sql",
]);

export const ADMIN_GATEWAY_REQUEST_QUOTA_FAILURE_CODES = [
  "gateway_quota_disabled",
  "gateway_quota_scope_denied",
  "gateway_quota_environment_forbidden",
  "gateway_quota_payload_invalid",
  "gateway_quota_policy_not_found",
  "gateway_quota_policy_version_conflict",
  "gateway_quota_attempt_conflict",
  "gateway_quota_exceeded",
  "gateway_quota_store_unavailable",
] as const;

export type AdminGatewayRequestQuotaFailureCode =
  typeof ADMIN_GATEWAY_REQUEST_QUOTA_FAILURE_CODES[number];
export type AdminGatewayRequestQuotaEnvironment = "development" | "test";

export type AdminGatewayRequestQuotaConfig = {
  mode: "offline" | "dev_admin_gateway_request_quota_http";
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  environment: AdminGatewayRequestQuotaEnvironment;
  applicationId: string;
  subjectRef: string;
};

export type AdminGatewayRequestQuotaPolicy = {
  schemaVersion: typeof SCHEMA_VERSION;
  policyId: string;
  tenantRef: string;
  workspaceId: string;
  environment: AdminGatewayRequestQuotaEnvironment;
  applicationId: string;
  period: typeof PERIOD;
  requestLimit: number;
  recordVersion: number;
  createdAt: string;
  updatedAt: string;
  createdBy: string;
  updatedBy: string;
  lastRequestId: string;
  lastAuditRef: string;
};

export type AdminGatewayRequestQuotaUsage = {
  schemaVersion: typeof SCHEMA_VERSION;
  tenantRef: string;
  workspaceId: string;
  environment: AdminGatewayRequestQuotaEnvironment;
  applicationId: string;
  period: typeof PERIOD;
  periodStart: string;
  policyId: string;
  policyVersion: number;
  requestLimit: number;
  admittedRequestCount: number;
  remainingRequestCount: number;
  updatedAt: string;
};

export type AdminGatewayRequestQuotaEnvelope = {
  requestId: string;
  tenantRef: string;
  workspaceId: string;
  environment: AdminGatewayRequestQuotaEnvironment;
  applicationId: string;
  policy: AdminGatewayRequestQuotaPolicy | null;
  usage: AdminGatewayRequestQuotaUsage | null;
  failureCode: AdminGatewayRequestQuotaFailureCode | null;
  auditRef: string;
};

type Document = Record<string, unknown>;

export function readAdminGatewayRequestQuotaConfig(
  context: { tenantRef: string; workspaceId: string; applicationId: string },
): AdminGatewayRequestQuotaConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  const configuredEnvironment = env.VITE_RADISHMIND_ADMIN_GATEWAY_QUOTA_ENVIRONMENT?.trim();
  return {
    mode: env.VITE_RADISHMIND_ADMIN_GATEWAY_QUOTA_SOURCE?.trim() === DEV_SOURCE
      ? "dev_admin_gateway_request_quota_http"
      : "offline",
    baseUrl: normalizeBaseUrl(
      env.VITE_RADISHMIND_ADMIN_GATEWAY_QUOTA_BASE_URL ??
      env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ??
      DEFAULT_BASE_URL,
    ),
    tenantRef: context.tenantRef.trim(),
    workspaceId: context.workspaceId.trim(),
    environment: configuredEnvironment === "development" ? "development" : "test",
    applicationId: context.applicationId.trim(),
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || "subject_demo_user",
  };
}

export async function readAdminGatewayRequestQuota(
  config: AdminGatewayRequestQuotaConfig,
): Promise<AdminGatewayRequestQuotaEnvelope> {
  return requestAdminGatewayRequestQuota(config, "GET", null);
}

export async function putAdminGatewayRequestQuota(
  config: AdminGatewayRequestQuotaConfig,
  expectedVersion: number,
  requestLimit: number,
): Promise<AdminGatewayRequestQuotaEnvelope> {
  if (!Number.isSafeInteger(expectedVersion) || expectedVersion < 0 || !isValidAdminGatewayRequestQuotaLimit(requestLimit)) {
    throw new Error("Admin Gateway quota update is invalid.");
  }
  return requestAdminGatewayRequestQuota(config, "PUT", {
    expected_version: expectedVersion,
    request_limit: requestLimit,
  });
}

export function isValidAdminGatewayRequestQuotaLimit(value: number): boolean {
  return Number.isSafeInteger(value) && value >= MINIMUM_REQUEST_LIMIT && value <= MAXIMUM_REQUEST_LIMIT;
}

async function requestAdminGatewayRequestQuota(
  config: AdminGatewayRequestQuotaConfig,
  method: "GET" | "PUT",
  body: { expected_version: number; request_limit: number } | null,
): Promise<AdminGatewayRequestQuotaEnvelope> {
  assertValidConfig(config);
  if (config.mode === "offline") {
    return offlineEnvelope(config);
  }
  const requestId = createRequestId(method === "GET" ? "read" : "write");
  const response = await fetch(
    `${config.baseUrl}/v1/admin/gateway-request-quotas/${encodeURIComponent(config.applicationId)}`,
    {
      method,
      headers: quotaHeaders(config, requestId, method),
      body: body === null ? undefined : JSON.stringify(body),
    },
  );
  let value: unknown;
  try {
    value = await response.json();
  } catch {
    throw new Error(`Admin Gateway quota returned an invalid HTTP ${response.status} envelope.`);
  }
  assertNoForbiddenFields(value);
  if (!isEnvelopeDocument(value, config)) {
    throw new Error(`Admin Gateway quota returned an invalid HTTP ${response.status} envelope.`);
  }
  return mapEnvelope(value);
}

function quotaHeaders(
  config: AdminGatewayRequestQuotaConfig,
  requestId: string,
  method: "GET" | "PUT",
): Record<string, string> {
  const permission = method === "GET" ? "admin_gateway_quotas:read" : "admin_gateway_quotas:write";
  return {
    Accept: "application/json",
    ...(method === "PUT" ? { "Content-Type": "application/json" } : {}),
    "X-Request-Id": requestId,
    "X-RadishMind-Dev-Read-Identity": "admin-gateway-quota-web",
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Read-Scopes": permission,
    "X-RadishMind-Dev-Read-Audit": `audit_${requestId}`,
    "X-RadishMind-Active-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Read-Membership-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Read-Membership-Permissions": permission,
    "X-RadishMind-Dev-Gateway-Quota-Environment": config.environment,
  };
}

function isEnvelopeDocument(value: unknown, config: AdminGatewayRequestQuotaConfig): value is Document {
  if (!isExactDocument(value, ENVELOPE_KEYS)) return false;
  const failureCode = nullableFailureCode(value.failure_code);
  if (failureCode === undefined || value.environment !== config.environment ||
    value.application_id !== config.applicationId || !validIdentifier(value.request_id) ||
    !matchesContextOrEmpty(value.tenant_ref, config.tenantRef) ||
    !matchesContextOrEmpty(value.workspace_id, config.workspaceId) ||
    !(value.audit_ref === "" || validIdentifier(value.audit_ref))) {
    return false;
  }
  if (failureCode !== null) {
    return value.policy === null && value.usage === null;
  }
  return value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId &&
    isPolicyDocument(value.policy, config) && isUsageDocument(value.usage, config, value.policy);
}

function isPolicyDocument(value: unknown, config: AdminGatewayRequestQuotaConfig): value is Document {
  if (!isExactDocument(value, POLICY_KEYS)) return false;
  return value.schema_version === SCHEMA_VERSION && validIdentifier(value.policy_id) &&
    value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId &&
    value.environment === config.environment && value.application_id === config.applicationId &&
    value.period === PERIOD && validLimitValue(value.request_limit) && validPositiveInteger(value.record_version) &&
    validTimestamp(value.created_at) && validTimestamp(value.updated_at) &&
    validIdentifier(value.created_by) && validIdentifier(value.updated_by) &&
    validIdentifier(value.last_request_id) && validIdentifier(value.last_audit_ref);
}

function isUsageDocument(
  value: unknown,
  config: AdminGatewayRequestQuotaConfig,
  policy: Document,
): value is Document {
  if (!isExactDocument(value, USAGE_KEYS)) return false;
  return value.schema_version === SCHEMA_VERSION && value.tenant_ref === config.tenantRef &&
    value.workspace_id === config.workspaceId && value.environment === config.environment &&
    value.application_id === config.applicationId && value.period === PERIOD &&
    typeof value.period_start === "string" && UTC_DATE.test(value.period_start) &&
    value.policy_id === policy.policy_id && value.policy_version === policy.record_version &&
    value.request_limit === policy.request_limit && validNonNegativeInteger(value.admitted_request_count) &&
    validNonNegativeInteger(value.remaining_request_count) &&
    value.remaining_request_count === Math.max(
      Number(value.request_limit) - Number(value.admitted_request_count),
      0,
    ) && validTimestamp(value.updated_at);
}

function mapEnvelope(value: Document): AdminGatewayRequestQuotaEnvelope {
  return {
    requestId: String(value.request_id),
    tenantRef: String(value.tenant_ref),
    workspaceId: String(value.workspace_id),
    environment: value.environment as AdminGatewayRequestQuotaEnvironment,
    applicationId: String(value.application_id),
    policy: value.policy === null ? null : mapPolicy(value.policy as Document),
    usage: value.usage === null ? null : mapUsage(value.usage as Document),
    failureCode: value.failure_code as AdminGatewayRequestQuotaFailureCode | null,
    auditRef: String(value.audit_ref),
  };
}

function mapPolicy(value: Document): AdminGatewayRequestQuotaPolicy {
  return {
    schemaVersion: SCHEMA_VERSION,
    policyId: String(value.policy_id),
    tenantRef: String(value.tenant_ref),
    workspaceId: String(value.workspace_id),
    environment: value.environment as AdminGatewayRequestQuotaEnvironment,
    applicationId: String(value.application_id),
    period: PERIOD,
    requestLimit: Number(value.request_limit),
    recordVersion: Number(value.record_version),
    createdAt: String(value.created_at),
    updatedAt: String(value.updated_at),
    createdBy: String(value.created_by),
    updatedBy: String(value.updated_by),
    lastRequestId: String(value.last_request_id),
    lastAuditRef: String(value.last_audit_ref),
  };
}

function mapUsage(value: Document): AdminGatewayRequestQuotaUsage {
  return {
    schemaVersion: SCHEMA_VERSION,
    tenantRef: String(value.tenant_ref),
    workspaceId: String(value.workspace_id),
    environment: value.environment as AdminGatewayRequestQuotaEnvironment,
    applicationId: String(value.application_id),
    period: PERIOD,
    periodStart: String(value.period_start),
    policyId: String(value.policy_id),
    policyVersion: Number(value.policy_version),
    requestLimit: Number(value.request_limit),
    admittedRequestCount: Number(value.admitted_request_count),
    remainingRequestCount: Number(value.remaining_request_count),
    updatedAt: String(value.updated_at),
  };
}

function offlineEnvelope(config: AdminGatewayRequestQuotaConfig): AdminGatewayRequestQuotaEnvelope {
  return {
    requestId: "offline_admin_gateway_quota",
    tenantRef: config.tenantRef,
    workspaceId: config.workspaceId,
    environment: config.environment,
    applicationId: config.applicationId,
    policy: null,
    usage: null,
    failureCode: "gateway_quota_disabled",
    auditRef: "offline_admin_gateway_quota",
  };
}

function assertValidConfig(config: AdminGatewayRequestQuotaConfig) {
  if (!validIdentifier(config.tenantRef) || !WORKSPACE_IDENTIFIER.test(config.workspaceId) ||
    !validIdentifier(config.applicationId) || !validIdentifier(config.subjectRef) ||
    (config.environment !== "development" && config.environment !== "test")) {
    throw new Error("Admin Gateway quota scope is invalid.");
  }
}

function assertNoForbiddenFields(value: unknown, path = "response") {
  if (Array.isArray(value)) {
    value.forEach((item, index) => assertNoForbiddenFields(item, `${path}[${index}]`));
    return;
  }
  if (!isDocument(value)) return;
  for (const [key, child] of Object.entries(value)) {
    if (FORBIDDEN_RESPONSE_FIELDS.has(key.toLowerCase())) {
      throw new Error(`Admin Gateway quota response contains forbidden field ${path}.${key}.`);
    }
    assertNoForbiddenFields(child, `${path}.${key}`);
  }
}

function isExactDocument(value: unknown, keys: readonly string[]): value is Document {
  if (!isDocument(value)) return false;
  const actualKeys = Object.keys(value).sort();
  const expectedKeys = [...keys].sort();
  return actualKeys.length === expectedKeys.length && actualKeys.every((key, index) => key === expectedKeys[index]);
}

function isDocument(value: unknown): value is Document {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function nullableFailureCode(value: unknown): AdminGatewayRequestQuotaFailureCode | null | undefined {
  if (value === null) return null;
  return typeof value === "string" && ADMIN_GATEWAY_REQUEST_QUOTA_FAILURE_CODES.includes(
    value as AdminGatewayRequestQuotaFailureCode,
  ) ? value as AdminGatewayRequestQuotaFailureCode : undefined;
}

function validIdentifier(value: unknown): value is string {
  return typeof value === "string" && IDENTIFIER.test(value);
}

function matchesContextOrEmpty(value: unknown, expected: string): boolean {
  return value === "" || value === expected;
}

function validLimitValue(value: unknown): boolean {
  return typeof value === "number" && isValidAdminGatewayRequestQuotaLimit(value);
}

function validPositiveInteger(value: unknown): boolean {
  return typeof value === "number" && Number.isSafeInteger(value) && value > 0;
}

function validNonNegativeInteger(value: unknown): boolean {
  return typeof value === "number" && Number.isSafeInteger(value) && value >= 0;
}

function validTimestamp(value: unknown): value is string {
  return typeof value === "string" && value.length <= 64 && !Number.isNaN(Date.parse(value));
}

function createRequestId(operation: "read" | "write"): string {
  return `admin-gateway-quota-${operation}-${Date.now().toString(36)}`;
}

function normalizeBaseUrl(baseUrl: string): string {
  const normalized = baseUrl.trim() || DEFAULT_BASE_URL;
  return normalized.endsWith("/") ? normalized.slice(0, -1) : normalized;
}
