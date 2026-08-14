const DEV_SOURCE = "dev-admin-gateway-model-pricing-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const IDENTIFIER = /^[A-Za-z0-9][A-Za-z0-9._:-]{0,255}$/u;
const WORKSPACE_IDENTIFIER = /^[A-Za-z0-9][A-Za-z0-9_.:/-]{0,159}$/u;
const POLICY_ID = /^gmp_[a-f0-9]{24}$/u;
const DIGEST = /^sha256:[a-f0-9]{64}$/u;
const SCHEMA_VERSION = "gateway_model_pricing_policy.v1";
const CURRENCY = "USD";
const TOKEN_UNIT = 1_000_000;

const ENVELOPE_KEYS = [
  "request_id",
  "tenant_ref",
  "workspace_id",
  "environment",
  "provider_id",
  "profile_id",
  "model_id",
  "policy",
  "current_version",
  "failure_code",
  "audit_ref",
] as const;

const POLICY_KEYS = [
  "schema_version",
  "policy_id",
  "record_version",
  "tenant_ref",
  "workspace_id",
  "environment",
  "provider_id",
  "profile_id",
  "model_id",
  "currency",
  "token_unit",
  "input_price_micros_per_token_unit",
  "output_price_micros_per_token_unit",
  "policy_digest",
  "reason",
  "updated_at",
  "updated_by_actor_ref",
  "request_id",
  "audit_ref",
] as const;

const FORBIDDEN_RESPONSE_FIELDS = new Set([
  "authorization",
  "api_key",
  "credential",
  "secret",
  "endpoint",
  "base_url",
  "dsn",
  "headers",
  "cookie",
  "invoice",
  "contract",
  "raw_request",
  "raw_response",
  "provider_raw_response",
  "prompt",
  "messages",
  "input",
  "output",
  "sql",
]);

export const ADMIN_GATEWAY_MODEL_PRICING_FAILURE_CODES = [
  "gateway_pricing_disabled",
  "gateway_pricing_scope_denied",
  "gateway_pricing_environment_forbidden",
  "gateway_pricing_payload_invalid",
  "gateway_pricing_policy_not_found",
  "gateway_pricing_policy_version_conflict",
  "gateway_pricing_policy_scope_conflict",
  "gateway_pricing_store_unavailable",
] as const;

export type AdminGatewayModelPricingFailureCode =
  typeof ADMIN_GATEWAY_MODEL_PRICING_FAILURE_CODES[number];
export type AdminGatewayModelPricingEnvironment = "development" | "test";

export type AdminGatewayModelPricingScope = {
  providerId: string;
  profileId: string;
  modelId: string;
};

export type AdminGatewayModelPricingConfig = {
  mode: "offline" | "dev_admin_gateway_model_pricing_http";
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  environment: AdminGatewayModelPricingEnvironment;
  subjectRef: string;
  initialScope: AdminGatewayModelPricingScope;
};

export type AdminGatewayModelPricingPolicy = {
  schemaVersion: typeof SCHEMA_VERSION;
  policyId: string;
  recordVersion: number;
  tenantRef: string;
  workspaceId: string;
  environment: AdminGatewayModelPricingEnvironment;
  providerId: string;
  profileId: string;
  modelId: string;
  currency: typeof CURRENCY;
  tokenUnit: typeof TOKEN_UNIT;
  inputPriceMicrosPerTokenUnit: number;
  outputPriceMicrosPerTokenUnit: number;
  policyDigest: string;
  reason: string;
  updatedAt: string;
  updatedByActorRef: string;
  requestId: string;
  auditRef: string;
};

export type AdminGatewayModelPricingEnvelope = {
  requestId: string;
  tenantRef: string;
  workspaceId: string;
  environment: AdminGatewayModelPricingEnvironment;
  scope: AdminGatewayModelPricingScope;
  policy: AdminGatewayModelPricingPolicy | null;
  currentVersion: number;
  failureCode: AdminGatewayModelPricingFailureCode | null;
  auditRef: string;
};

type Document = Record<string, unknown>;

export function readAdminGatewayModelPricingConfig(
  context: { tenantRef: string; workspaceId: string },
): AdminGatewayModelPricingConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  const configuredEnvironment = env.VITE_RADISHMIND_ADMIN_GATEWAY_PRICING_ENVIRONMENT?.trim();
  return {
    mode: env.VITE_RADISHMIND_ADMIN_GATEWAY_PRICING_SOURCE?.trim() === DEV_SOURCE
      ? "dev_admin_gateway_model_pricing_http"
      : "offline",
    baseUrl: normalizeBaseUrl(
      env.VITE_RADISHMIND_ADMIN_GATEWAY_PRICING_BASE_URL ??
      env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ??
      DEFAULT_BASE_URL,
    ),
    tenantRef: context.tenantRef.trim(),
    workspaceId: context.workspaceId.trim(),
    environment: configuredEnvironment === "development" ? "development" : "test",
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || "subject_demo_user",
    initialScope: {
      providerId: env.VITE_RADISHMIND_ADMIN_GATEWAY_PRICING_PROVIDER_ID?.trim() || "mock",
      profileId: env.VITE_RADISHMIND_ADMIN_GATEWAY_PRICING_PROFILE_ID?.trim() || "mock-dev",
      modelId: env.VITE_RADISHMIND_ADMIN_GATEWAY_PRICING_MODEL_ID?.trim() || "mock-model",
    },
  };
}

export async function readAdminGatewayModelPricing(
  config: AdminGatewayModelPricingConfig,
  scope: AdminGatewayModelPricingScope,
): Promise<AdminGatewayModelPricingEnvelope> {
  return requestAdminGatewayModelPricing(config, scope, "GET", null);
}

export async function putAdminGatewayModelPricing(
  config: AdminGatewayModelPricingConfig,
  scope: AdminGatewayModelPricingScope,
  update: {
    expectedVersion: number;
    inputPriceMicrosPerTokenUnit: number;
    outputPriceMicrosPerTokenUnit: number;
    reason: string;
  },
): Promise<AdminGatewayModelPricingEnvelope> {
  if (!isValidAdminGatewayModelPricingScope(scope) || !isValidVersion(update.expectedVersion) ||
    !isValidRate(update.inputPriceMicrosPerTokenUnit) || !isValidRate(update.outputPriceMicrosPerTokenUnit) ||
    !isValidAdminGatewayModelPricingReason(update.reason)) {
    throw new Error("Admin Gateway pricing update is invalid.");
  }
  return requestAdminGatewayModelPricing(config, scope, "PUT", {
    expected_version: update.expectedVersion,
    provider_id: scope.providerId,
    profile_id: scope.profileId,
    model_id: scope.modelId,
    currency: CURRENCY,
    input_price_micros_per_token_unit: update.inputPriceMicrosPerTokenUnit,
    output_price_micros_per_token_unit: update.outputPriceMicrosPerTokenUnit,
    reason: update.reason.trim(),
  });
}

export function isValidAdminGatewayModelPricingScope(scope: AdminGatewayModelPricingScope): boolean {
  return [scope.providerId, scope.profileId, scope.modelId]
    .every((value) => IDENTIFIER.test(value.trim()));
}

export function isValidAdminGatewayModelPricingRate(value: number): boolean {
  return isValidRate(value);
}

export function isValidAdminGatewayModelPricingReason(value: string): boolean {
  const reason = value.trim();
  if (!reason || reason.length > 512 || /[\r\n\0]/u.test(reason)) return false;
  return ![
    "authorization:", "api_key", "apikey", "password", "secret", "credential",
    "postgres://", "postgresql://", "mysql://", "mongodb://", "https://", "http://",
  ].some((forbidden) => reason.toLowerCase().includes(forbidden));
}

async function requestAdminGatewayModelPricing(
  config: AdminGatewayModelPricingConfig,
  rawScope: AdminGatewayModelPricingScope,
  method: "GET" | "PUT",
  body: Record<string, unknown> | null,
): Promise<AdminGatewayModelPricingEnvelope> {
  const scope = normalizeScope(rawScope);
  assertValidConfig(config);
  if (!isValidAdminGatewayModelPricingScope(scope)) {
    throw new Error("Admin Gateway pricing scope is invalid.");
  }
  if (config.mode === "offline") return offlineEnvelope(config, scope);

  const requestId = createRequestId(method === "GET" ? "read" : "write");
  const query = method === "GET"
    ? `?${new URLSearchParams({ provider_id: scope.providerId, profile_id: scope.profileId, model_id: scope.modelId })}`
    : "";
  const response = await fetch(`${config.baseUrl}/v1/admin/gateway-model-pricing-policy${query}`, {
    method,
    headers: pricingHeaders(config, requestId, method),
    body: body === null ? undefined : JSON.stringify(body),
  });
  let value: unknown;
  try {
    value = await response.json();
  } catch {
    throw new Error(`Admin Gateway pricing returned an invalid HTTP ${response.status} envelope.`);
  }
  assertNoForbiddenFields(value);
  if (!isEnvelopeDocument(value, config, scope)) {
    throw new Error(`Admin Gateway pricing returned an invalid HTTP ${response.status} envelope.`);
  }
  return mapEnvelope(value);
}

function pricingHeaders(
  config: AdminGatewayModelPricingConfig,
  requestId: string,
  method: "GET" | "PUT",
): Record<string, string> {
  const permission = method === "GET" ? "admin_gateway_pricing:read" : "admin_gateway_pricing:write";
  return {
    Accept: "application/json",
    ...(method === "PUT" ? { "Content-Type": "application/json" } : {}),
    "X-Request-Id": requestId,
    "X-RadishMind-Dev-Read-Identity": "admin-gateway-pricing-web",
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Read-Scopes": permission,
    "X-RadishMind-Dev-Read-Audit": `audit_${requestId}`,
    "X-RadishMind-Active-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Read-Membership-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Read-Membership-Permissions": permission,
    "X-RadishMind-Dev-Gateway-Pricing-Environment": config.environment,
  };
}

function isEnvelopeDocument(
  value: unknown,
  config: AdminGatewayModelPricingConfig,
  scope: AdminGatewayModelPricingScope,
): value is Document {
  if (!isExactDocument(value, ENVELOPE_KEYS)) return false;
  const failureCode = nullableFailureCode(value.failure_code);
  const contextMatches = failureCode === null
    ? value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId
    : matchesContextOrEmpty(value.tenant_ref, config.tenantRef) && matchesContextOrEmpty(value.workspace_id, config.workspaceId);
  if (failureCode === undefined || !validIdentifier(value.request_id) || !contextMatches ||
    value.environment !== config.environment || value.provider_id !== scope.providerId ||
    value.profile_id !== scope.profileId || value.model_id !== scope.modelId ||
    !isValidVersion(value.current_version) || !(value.audit_ref === "" || validIdentifier(value.audit_ref))) {
    return false;
  }
  if (failureCode !== null) {
    return value.policy === null &&
      (failureCode === "gateway_pricing_policy_version_conflict" ? value.current_version >= 1 : true);
  }
  return isPolicyDocument(value.policy, config, scope) &&
    value.current_version === value.policy.record_version;
}

function isPolicyDocument(
  value: unknown,
  config: AdminGatewayModelPricingConfig,
  scope: AdminGatewayModelPricingScope,
): value is Document {
  if (!isExactDocument(value, POLICY_KEYS)) return false;
  return value.schema_version === SCHEMA_VERSION && typeof value.policy_id === "string" && POLICY_ID.test(value.policy_id) &&
    isPositiveVersion(value.record_version) && value.tenant_ref === config.tenantRef &&
    value.workspace_id === config.workspaceId && value.environment === config.environment &&
    value.provider_id === scope.providerId && value.profile_id === scope.profileId && value.model_id === scope.modelId &&
    value.currency === CURRENCY && value.token_unit === TOKEN_UNIT &&
    isValidRate(value.input_price_micros_per_token_unit) && isValidRate(value.output_price_micros_per_token_unit) &&
    typeof value.policy_digest === "string" && DIGEST.test(value.policy_digest) &&
    typeof value.reason === "string" && isValidAdminGatewayModelPricingReason(value.reason) &&
    validTimestamp(value.updated_at) && validIdentifier(value.updated_by_actor_ref) &&
    validIdentifier(value.request_id) && validIdentifier(value.audit_ref);
}

function mapEnvelope(value: Document): AdminGatewayModelPricingEnvelope {
  return {
    requestId: value.request_id as string,
    tenantRef: value.tenant_ref as string,
    workspaceId: value.workspace_id as string,
    environment: value.environment as AdminGatewayModelPricingEnvironment,
    scope: {
      providerId: value.provider_id as string,
      profileId: value.profile_id as string,
      modelId: value.model_id as string,
    },
    policy: value.policy === null ? null : mapPolicy(value.policy as Document),
    currentVersion: value.current_version as number,
    failureCode: value.failure_code as AdminGatewayModelPricingFailureCode | null,
    auditRef: value.audit_ref as string,
  };
}

function mapPolicy(value: Document): AdminGatewayModelPricingPolicy {
  return {
    schemaVersion: value.schema_version as typeof SCHEMA_VERSION,
    policyId: value.policy_id as string,
    recordVersion: value.record_version as number,
    tenantRef: value.tenant_ref as string,
    workspaceId: value.workspace_id as string,
    environment: value.environment as AdminGatewayModelPricingEnvironment,
    providerId: value.provider_id as string,
    profileId: value.profile_id as string,
    modelId: value.model_id as string,
    currency: value.currency as typeof CURRENCY,
    tokenUnit: value.token_unit as typeof TOKEN_UNIT,
    inputPriceMicrosPerTokenUnit: value.input_price_micros_per_token_unit as number,
    outputPriceMicrosPerTokenUnit: value.output_price_micros_per_token_unit as number,
    policyDigest: value.policy_digest as string,
    reason: value.reason as string,
    updatedAt: value.updated_at as string,
    updatedByActorRef: value.updated_by_actor_ref as string,
    requestId: value.request_id as string,
    auditRef: value.audit_ref as string,
  };
}

function offlineEnvelope(
  config: AdminGatewayModelPricingConfig,
  scope: AdminGatewayModelPricingScope,
): AdminGatewayModelPricingEnvelope {
  return {
    requestId: "admin-gateway-pricing-offline",
    tenantRef: config.tenantRef,
    workspaceId: config.workspaceId,
    environment: config.environment,
    scope,
    policy: null,
    currentVersion: 0,
    failureCode: "gateway_pricing_disabled",
    auditRef: "audit_admin_gateway_pricing_offline",
  };
}

function assertValidConfig(config: AdminGatewayModelPricingConfig) {
  if (!validIdentifier(config.tenantRef) || !validWorkspaceIdentifier(config.workspaceId) ||
    !validIdentifier(config.subjectRef) || !["development", "test"].includes(config.environment) ||
    !/^https?:\/\/[^\s]+$/u.test(config.baseUrl)) {
    throw new Error("Admin Gateway pricing configuration is invalid.");
  }
}

function normalizeScope(scope: AdminGatewayModelPricingScope): AdminGatewayModelPricingScope {
  return {
    providerId: scope.providerId.trim(),
    profileId: scope.profileId.trim(),
    modelId: scope.modelId.trim(),
  };
}

function nullableFailureCode(value: unknown): AdminGatewayModelPricingFailureCode | null | undefined {
  if (value === null) return null;
  return typeof value === "string" && ADMIN_GATEWAY_MODEL_PRICING_FAILURE_CODES.includes(
    value as AdminGatewayModelPricingFailureCode,
  ) ? value as AdminGatewayModelPricingFailureCode : undefined;
}

function assertNoForbiddenFields(value: unknown, path = "response") {
  if (Array.isArray(value)) {
    value.forEach((item, index) => assertNoForbiddenFields(item, `${path}[${index}]`));
    return;
  }
  if (!isDocument(value)) return;
  for (const [key, nested] of Object.entries(value)) {
    if (FORBIDDEN_RESPONSE_FIELDS.has(key.toLowerCase())) {
      throw new Error(`Admin Gateway pricing response contains forbidden field ${path}.${key}.`);
    }
    assertNoForbiddenFields(nested, `${path}.${key}`);
  }
}

function isExactDocument<const T extends readonly string[]>(value: unknown, keys: T): value is Document {
  if (!isDocument(value)) return false;
  const actual = Object.keys(value).sort();
  const expected = [...keys].sort();
  return actual.length === expected.length && actual.every((key, index) => key === expected[index]);
}

function isDocument(value: unknown): value is Document {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function validIdentifier(value: unknown): value is string {
  return typeof value === "string" && IDENTIFIER.test(value);
}

function validWorkspaceIdentifier(value: unknown): value is string {
  return typeof value === "string" && WORKSPACE_IDENTIFIER.test(value);
}

function matchesContextOrEmpty(value: unknown, expected: string): boolean {
  return value === "" || value === expected;
}

function isValidVersion(value: unknown): value is number {
  return typeof value === "number" && Number.isSafeInteger(value) && value >= 0;
}

function isPositiveVersion(value: unknown): value is number {
  return isValidVersion(value) && value >= 1;
}

function isValidRate(value: unknown): value is number {
  return typeof value === "number" && Number.isSafeInteger(value) && value >= 0;
}

function validTimestamp(value: unknown): value is string {
  return typeof value === "string" && value.length > 0 && Number.isFinite(Date.parse(value));
}

function normalizeBaseUrl(value: string): string {
  return value.trim().replace(/\/$/u, "");
}

function createRequestId(operation: string): string {
  return `admin-gateway-pricing-${operation}-${Date.now().toString(36)}`;
}
