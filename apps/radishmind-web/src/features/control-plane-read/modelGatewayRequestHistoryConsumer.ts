export type ModelGatewayRequestHistoryConfig = {
  mode: "offline" | "dev_gateway_request_history_http";
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  consumerRef: string;
  applicationId: string;
  subjectRef: string;
};

export type GatewayRequestHistoryFilter = {
  route: string;
  protocol: "" | "openai-chat-completions" | "openai-responses" | "anthropic-messages";
  provider: string;
  profile: string;
  model: string;
  status: "" | "started" | "succeeded" | "failed" | "canceled";
  failureBoundary: string;
  usageAvailability: "" | "reported" | "not_reported" | "not_applicable";
  startedFrom: string;
  startedTo: string;
};

export type GatewayRequestCostAvailability =
  | "estimated"
  | "usage_not_reported"
  | "price_not_configured"
  | "price_unavailable"
  | "not_applicable"
  | "legacy_not_captured";

export type GatewayRequestCostEstimate = {
  schemaVersion: "gateway_request_cost_estimate.v1";
  availability: GatewayRequestCostAvailability;
  reason: string;
  currency: "USD" | "";
  estimatedCostMicros: number | null;
  tokenUnit: number | null;
  inputPriceMicrosPerTokenUnit: number | null;
  outputPriceMicrosPerTokenUnit: number | null;
  pricingPolicyId: string;
  pricingPolicyVersion: number | null;
  pricingPolicyDigest: string;
  roundingMode: "half_up_to_currency_micro" | "";
};

export type GatewayRequestHistorySummary = {
  schemaVersion: "gateway_request_record.v1" | "gateway_request_record.v2";
  recordVersion: number;
  storeMode: "memory_dev" | "sqlite_dev" | "postgres_dev_test";
  requestId: string;
  auditRef: string;
  route: string;
  protocol: string;
  stream: boolean;
  status: "started" | "succeeded" | "failed" | "canceled";
  startedAt: string;
  completedAt: string;
  durationMs: number;
  providerDurationMs: number;
  providerDurationAvailable: boolean;
  selectionSource: string;
  selectedProvider: string;
  selectedProfile: string;
  selectedModel: string;
  providerRouteConfigurationId: string;
  providerRouteGeneration: number;
  providerRouteSnapshotDigest: string;
  httpStatusCode: number;
  failureCode: string;
  failureBoundary: string;
  usageAvailability: "reported" | "not_reported" | "not_applicable";
  usageSource: string;
  inputTokens: number;
  outputTokens: number;
  totalTokens: number;
  costEstimate: GatewayRequestCostEstimate;
  staleStarted: boolean;
};

export type GatewayRequestHistoryDetail = GatewayRequestHistorySummary & {
  tenantRef: string;
  workspaceId: string;
  consumerRef: string;
  applicationId: string;
  subjectRef: string;
  gatewayDurationMs: number;
  gatewayDurationAvailable: boolean;
};

export type GatewayRequestHistoryState = {
  status: "offline" | "loading" | "ready" | "empty" | "failed";
  requests: GatewayRequestHistorySummary[];
  nextCursor: string;
  hasMore: boolean;
  requestId: string;
  auditRef: string;
  failureCode: string;
  failureSummary: string;
};

type GatewayRequestSummaryDocument = {
  schema_version: "gateway_request_record.v1" | "gateway_request_record.v2";
  record_version: number;
  store_mode: "memory_dev" | "sqlite_dev" | "postgres_dev_test";
  request_id: string;
  audit_ref: string;
  route: string;
  protocol: string;
  stream: boolean;
  status: GatewayRequestHistorySummary["status"];
  started_at: string;
  completed_at: string;
  duration_ms: number;
  provider_duration_ms: number;
  provider_duration_available: boolean;
  selection_source: string;
  selected_provider: string;
  selected_profile: string;
  selected_model: string;
  provider_route_configuration_id?: string;
  provider_route_generation?: number;
  provider_route_snapshot_digest?: string;
  http_status_code: number;
  failure_code: string;
  failure_boundary: string;
  usage_availability: GatewayRequestHistorySummary["usageAvailability"];
  usage: GatewayRequestUsageDocument;
  cost_estimate: GatewayRequestCostEstimateDocument;
  stale_started: boolean;
};

type GatewayRequestUsageDocument = {
  availability: GatewayRequestHistorySummary["usageAvailability"];
  source: string;
  input_tokens: number;
  output_tokens: number;
  total_tokens: number;
};

type GatewayRequestCostEstimateDocument = {
  schema_version: "gateway_request_cost_estimate.v1";
  availability: GatewayRequestCostAvailability;
  reason: string;
  currency?: "USD";
  estimated_cost_micros?: number;
  token_unit?: number;
  input_price_micros_per_token_unit?: number;
  output_price_micros_per_token_unit?: number;
  pricing_policy_id?: string;
  pricing_policy_version?: number;
  pricing_policy_digest?: string;
  rounding_mode?: "half_up_to_currency_micro";
};

type GatewayRequestDetailDocument = Omit<GatewayRequestSummaryDocument, "usage_availability"> & {
  tenant_ref: string;
  workspace_id: string;
  consumer_ref: string;
  application_id?: string;
  subject_ref: string;
  gateway_duration_ms: number;
  gateway_duration_available: boolean;
};

type GatewayRequestListEnvelope = {
  request_id: string;
  tenant_ref: string;
  workspace_id: string;
  consumer_ref: string;
  application_id?: string;
  requests: GatewayRequestSummaryDocument[];
  next_cursor: string;
  has_more: boolean;
  failure_code: string | null;
  failure_summary: string;
  audit_ref: string;
};

type GatewayRequestReadEnvelope = {
  request_id: string;
  tenant_ref: string;
  workspace_id: string;
  consumer_ref: string;
  application_id?: string;
  request: GatewayRequestDetailDocument | null;
  failure_code: string | null;
  failure_summary: string;
  audit_ref: string;
};

const DEV_SOURCE = "dev-gateway-request-history-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const FORBIDDEN_FIELDS = new Set([
  "prompt", "messages", "instructions", "input", "response", "response_body", "stream_delta",
  "authorization", "api_key", "credential", "secret", "endpoint", "base_url", "dsn", "provider_raw_envelope",
  "raw_error", "stderr", "stack_trace", "tool_payload", "tool_result", "cookie", "headers",
]);

export const EMPTY_GATEWAY_REQUEST_HISTORY_FILTER: GatewayRequestHistoryFilter = {
  route: "",
  protocol: "",
  provider: "",
  profile: "",
  model: "",
  status: "",
  failureBoundary: "",
  usageAvailability: "",
  startedFrom: "",
  startedTo: "",
};

export function readModelGatewayRequestHistoryConfig(): ModelGatewayRequestHistoryConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  return {
    mode: env.VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_SOURCE?.trim() === DEV_SOURCE
      ? "dev_gateway_request_history_http"
      : "offline",
    baseUrl: normalizeBaseUrl(
      env.VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_BASE_URL ??
        env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ??
        DEFAULT_BASE_URL,
    ),
    tenantRef: env.VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_TENANT_REF?.trim() || "tenant_demo",
    workspaceId: env.VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_WORKSPACE_ID?.trim() || "workspace_demo",
    consumerRef: env.VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_CONSUMER_REF?.trim() || "consumer_web_dev",
    applicationId: env.VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_APPLICATION_ID?.trim() || "",
    subjectRef: env.VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_SUBJECT_REF?.trim() || "subject_web_dev",
  };
}

export function initialGatewayRequestHistoryState(config: ModelGatewayRequestHistoryConfig): GatewayRequestHistoryState {
  if (config.mode === "dev_gateway_request_history_http") {
    return emptyHistoryState("loading", "gateway-request-history-loading", "audit_gateway_request_history_loading", "");
  }
  return emptyHistoryState(
    "offline",
    "gateway-request-history-offline",
    "audit_gateway_request_history_offline",
    "Offline evidence mode does not request Gateway history.",
  );
}

export async function listGatewayRequestHistory(
  config: ModelGatewayRequestHistoryConfig,
  filter: GatewayRequestHistoryFilter,
  cursor = "",
  previousRequests: GatewayRequestHistorySummary[] = [],
): Promise<GatewayRequestHistoryState> {
  if (config.mode !== "dev_gateway_request_history_http") {
    return initialGatewayRequestHistoryState(config);
  }
  const query = scopedQuery(config);
  query.set("limit", "25");
  if (cursor) query.set("cursor", cursor);
  appendFilter(query, filter);
  const response = await fetch(`${config.baseUrl}/v1/model-gateway/requests?${query}`, {
    headers: gatewayRequestHistoryHeaders(config, "list"),
  });
  const body: unknown = await response.json();
  assertNoForbiddenFields(body);
  if (!response.ok || !isGatewayRequestListEnvelope(body, config)) {
    throw new Error(`Gateway request history route failed with HTTP ${response.status}`);
  }
  if (body.failure_code) {
    return emptyHistoryState("failed", body.request_id, body.audit_ref, body.failure_summary, body.failure_code);
  }
  const requests = body.requests.map(mapGatewayRequestSummary);
  return {
    status: previousRequests.length + requests.length > 0 ? "ready" : "empty",
    requests: [...previousRequests, ...requests],
    nextCursor: body.next_cursor,
    hasMore: body.has_more,
    requestId: body.request_id,
    auditRef: body.audit_ref,
    failureCode: "",
    failureSummary: "",
  };
}

export async function readGatewayRequestHistoryDetail(
  config: ModelGatewayRequestHistoryConfig,
  requestId: string,
): Promise<GatewayRequestHistoryDetail> {
  if (config.mode !== "dev_gateway_request_history_http") {
    throw new Error("Gateway request history detail is unavailable in offline mode.");
  }
  const query = scopedQuery(config);
  const response = await fetch(
    `${config.baseUrl}/v1/model-gateway/requests/${encodeURIComponent(requestId)}?${query}`,
    { headers: gatewayRequestHistoryHeaders(config, "detail") },
  );
  const body: unknown = await response.json();
  assertNoForbiddenFields(body);
  if (!response.ok || !isGatewayRequestReadEnvelope(body, config)) {
    throw new Error(`Gateway request detail route failed with HTTP ${response.status}`);
  }
  if (body.failure_code || !body.request) {
    throw new Error(`${body.failure_code || "gateway_request_record_not_found"}: ${body.failure_summary}`);
  }
  return mapGatewayRequestDetail(body.request);
}

function scopedQuery(config: ModelGatewayRequestHistoryConfig): URLSearchParams {
  const query = new URLSearchParams({ workspace_id: config.workspaceId, consumer_ref: config.consumerRef });
  if (config.applicationId) query.set("application_id", config.applicationId);
  return query;
}

function appendFilter(query: URLSearchParams, filter: GatewayRequestHistoryFilter) {
  const exactValues: Array<[string, string]> = [
    ["route", filter.route], ["protocol", filter.protocol], ["provider", filter.provider],
    ["profile", filter.profile], ["model", filter.model], ["status", filter.status],
    ["failure_boundary", filter.failureBoundary], ["usage_availability", filter.usageAvailability],
  ];
  for (const [key, value] of exactValues) if (value.trim()) query.set(key, value.trim());
  if (filter.startedFrom) query.set("started_from", new Date(filter.startedFrom).toISOString());
  if (filter.startedTo) query.set("started_to", new Date(filter.startedTo).toISOString());
}

function gatewayRequestHistoryHeaders(config: ModelGatewayRequestHistoryConfig, operation: string): HeadersInit {
  const headers: Record<string, string> = {
    Accept: "application/json",
    "X-Request-Id": `dev-gateway-request-history-${operation}`,
    "X-RadishMind-Dev-Gateway-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Gateway-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Gateway-Consumer": config.consumerRef,
    "X-RadishMind-Dev-Gateway-Subject": config.subjectRef,
    "X-RadishMind-Dev-Gateway-Scopes": "gateway_requests:read",
    "X-RadishMind-Dev-Gateway-Audit": `audit_dev_gateway_request_history_${operation}`,
  };
  if (config.applicationId) headers["X-RadishMind-Dev-Gateway-Application"] = config.applicationId;
  return headers;
}

function mapGatewayRequestSummary(value: GatewayRequestSummaryDocument): GatewayRequestHistorySummary {
  return {
    schemaVersion: value.schema_version,
    recordVersion: value.record_version,
    storeMode: value.store_mode,
    requestId: value.request_id,
    auditRef: value.audit_ref,
    route: value.route,
    protocol: value.protocol,
    stream: value.stream,
    status: value.status,
    startedAt: value.started_at,
    completedAt: value.completed_at,
    durationMs: value.duration_ms,
    providerDurationMs: value.provider_duration_ms,
    providerDurationAvailable: value.provider_duration_available,
    selectionSource: value.selection_source,
    selectedProvider: value.selected_provider,
    selectedProfile: value.selected_profile,
    selectedModel: value.selected_model,
    providerRouteConfigurationId: value.provider_route_configuration_id ?? "",
    providerRouteGeneration: value.provider_route_generation ?? 0,
    providerRouteSnapshotDigest: value.provider_route_snapshot_digest ?? "",
    httpStatusCode: value.http_status_code,
    failureCode: value.failure_code,
    failureBoundary: value.failure_boundary,
    usageAvailability: value.usage_availability,
    usageSource: value.usage.source,
    inputTokens: value.usage.input_tokens,
    outputTokens: value.usage.output_tokens,
    totalTokens: value.usage.total_tokens,
    costEstimate: mapGatewayRequestCostEstimate(value.cost_estimate),
    staleStarted: value.stale_started,
  };
}

function mapGatewayRequestCostEstimate(value: GatewayRequestCostEstimateDocument): GatewayRequestCostEstimate {
  return {
    schemaVersion: value.schema_version,
    availability: value.availability,
    reason: value.reason,
    currency: value.currency ?? "",
    estimatedCostMicros: value.estimated_cost_micros ?? null,
    tokenUnit: value.token_unit ?? null,
    inputPriceMicrosPerTokenUnit: value.input_price_micros_per_token_unit ?? null,
    outputPriceMicrosPerTokenUnit: value.output_price_micros_per_token_unit ?? null,
    pricingPolicyId: value.pricing_policy_id ?? "",
    pricingPolicyVersion: value.pricing_policy_version ?? null,
    pricingPolicyDigest: value.pricing_policy_digest ?? "",
    roundingMode: value.rounding_mode ?? "",
  };
}

function mapGatewayRequestDetail(value: GatewayRequestDetailDocument): GatewayRequestHistoryDetail {
  return {
    ...mapGatewayRequestSummary({ ...value, usage_availability: value.usage.availability }),
    tenantRef: value.tenant_ref,
    workspaceId: value.workspace_id,
    consumerRef: value.consumer_ref,
    applicationId: value.application_id ?? "",
    subjectRef: value.subject_ref,
    gatewayDurationMs: value.gateway_duration_ms,
    gatewayDurationAvailable: value.gateway_duration_available,
  };
}

function isGatewayRequestListEnvelope(
  value: unknown,
  config: ModelGatewayRequestHistoryConfig,
): value is GatewayRequestListEnvelope {
  if (!isRecord(value)) return false;
  return hasOnlyKeys(value, [
    "request_id", "tenant_ref", "workspace_id", "consumer_ref", "application_id", "requests",
    "next_cursor", "has_more", "failure_code", "failure_summary", "audit_ref",
  ]) && typeof value.request_id === "string" && value.tenant_ref === config.tenantRef &&
    value.workspace_id === config.workspaceId && value.consumer_ref === config.consumerRef &&
    matchesOptionalApplication(value.application_id, config.applicationId) && Array.isArray(value.requests) &&
    value.requests.every((request) => isGatewayRequestSummaryDocument(request)) && typeof value.next_cursor === "string" &&
    typeof value.has_more === "boolean" && isNullableString(value.failure_code) &&
    typeof value.failure_summary === "string" && typeof value.audit_ref === "string";
}

function isGatewayRequestReadEnvelope(
  value: unknown,
  config: ModelGatewayRequestHistoryConfig,
): value is GatewayRequestReadEnvelope {
  if (!isRecord(value)) return false;
  return hasOnlyKeys(value, [
    "request_id", "tenant_ref", "workspace_id", "consumer_ref", "application_id", "request",
    "failure_code", "failure_summary", "audit_ref",
  ]) && typeof value.request_id === "string" && value.tenant_ref === config.tenantRef &&
    value.workspace_id === config.workspaceId && value.consumer_ref === config.consumerRef &&
    matchesOptionalApplication(value.application_id, config.applicationId) &&
    (value.request === null || isGatewayRequestDetailDocument(value.request)) &&
    isNullableString(value.failure_code) && typeof value.failure_summary === "string" && typeof value.audit_ref === "string";
}

function isGatewayRequestSummaryDocument(
  value: unknown,
  detailProjection = false,
): value is GatewayRequestSummaryDocument {
  if (!isRecord(value) || !isGatewayRequestUsageDocument(value.usage) ||
    !isGatewayRequestCostEstimateDocument(value.cost_estimate)) return false;
  const allowedKeys = [
    "schema_version", "record_version", "store_mode", "request_id", "audit_ref", "route", "protocol",
    "stream", "status", "started_at", "completed_at", "duration_ms", "provider_duration_ms",
    "provider_duration_available", "selection_source", "selected_provider", "selected_profile",
    "selected_model", "provider_route_configuration_id", "provider_route_generation",
    "provider_route_snapshot_digest", "http_status_code", "failure_code", "failure_boundary",
    "usage_availability", "usage", "cost_estimate", "stale_started",
    ...(detailProjection ? [
      "tenant_ref", "workspace_id", "consumer_ref", "application_id", "subject_ref",
      "gateway_duration_ms", "gateway_duration_available",
    ] : []),
  ];
  return hasOnlyKeys(value, allowedKeys) &&
    ["gateway_request_record.v1", "gateway_request_record.v2"].includes(String(value.schema_version)) &&
    (value.schema_version === "gateway_request_record.v1"
      ? value.cost_estimate.availability === "legacy_not_captured"
      : value.cost_estimate.availability !== "legacy_not_captured") &&
    costAvailabilityMatchesUsage(value.cost_estimate.availability, value.usage.availability) &&
    isNonNegativeInteger(value.record_version) &&
    ["memory_dev", "sqlite_dev", "postgres_dev_test"].includes(String(value.store_mode)) &&
    stringFields(value, ["request_id", "audit_ref", "route", "protocol", "started_at", "completed_at", "selection_source", "selected_provider", "selected_profile", "selected_model", "failure_code", "failure_boundary"]) &&
    typeof value.stream === "boolean" && ["started", "succeeded", "failed", "canceled"].includes(String(value.status)) &&
    isNonNegativeInteger(value.duration_ms) && isNonNegativeInteger(value.provider_duration_ms) &&
    typeof value.provider_duration_available === "boolean" && isNonNegativeInteger(value.http_status_code) &&
    ["reported", "not_reported", "not_applicable"].includes(String(value.usage_availability)) &&
    value.usage_availability === value.usage.availability &&
    typeof value.stale_started === "boolean" && isProviderRouteLineageDocument(value);
}

function isGatewayRequestCostEstimateDocument(value: unknown): value is GatewayRequestCostEstimateDocument {
  if (!isRecord(value) || value.schema_version !== "gateway_request_cost_estimate.v1" ||
    typeof value.availability !== "string" || ![
      "estimated", "usage_not_reported", "price_not_configured", "price_unavailable",
      "not_applicable", "legacy_not_captured",
    ].includes(value.availability) || typeof value.reason !== "string" || value.reason.length > 160) {
    return false;
  }
  const estimateKeys = [
    "currency", "estimated_cost_micros", "token_unit", "input_price_micros_per_token_unit",
    "output_price_micros_per_token_unit", "pricing_policy_id", "pricing_policy_version",
    "pricing_policy_digest", "rounding_mode",
  ];
  if (value.availability !== "estimated") {
    return hasOnlyKeys(value, ["schema_version", "availability", "reason"]) &&
      value.reason.length > 0 && estimateKeys.every((key) => value[key] === undefined);
  }
  return hasOnlyKeys(value, ["schema_version", "availability", "reason", ...estimateKeys]) &&
    value.reason === "" && value.currency === "USD" && value.token_unit === 1_000_000 &&
    isNonNegativeInteger(value.estimated_cost_micros) &&
    isNonNegativeInteger(value.input_price_micros_per_token_unit) &&
    isNonNegativeInteger(value.output_price_micros_per_token_unit) &&
    typeof value.pricing_policy_id === "string" && /^gmp_[a-f0-9]{24}$/u.test(value.pricing_policy_id) &&
    typeof value.pricing_policy_version === "number" && Number.isSafeInteger(value.pricing_policy_version) &&
    value.pricing_policy_version >= 1 && typeof value.pricing_policy_digest === "string" &&
    /^sha256:[a-f0-9]{64}$/u.test(value.pricing_policy_digest) &&
    value.rounding_mode === "half_up_to_currency_micro";
}

function isGatewayRequestDetailDocument(value: unknown): value is GatewayRequestDetailDocument {
  if (!isRecord(value) || !isRecord(value.usage)) return false;
  return isGatewayRequestSummaryDocument({ ...value, usage_availability: value.usage.availability }, true) &&
    stringFields(value, ["tenant_ref", "workspace_id", "consumer_ref", "subject_ref"]) &&
    (value.application_id === undefined || typeof value.application_id === "string") &&
    isNonNegativeInteger(value.gateway_duration_ms) && typeof value.gateway_duration_available === "boolean" &&
    isGatewayRequestUsageDocument(value.usage);
}

function isGatewayRequestUsageDocument(value: unknown): value is GatewayRequestUsageDocument {
  if (!isRecord(value) || !hasOnlyKeys(value, ["availability", "source", "input_tokens", "output_tokens", "total_tokens"]) ||
    typeof value.availability !== "string" || typeof value.source !== "string" ||
    !["reported", "not_reported", "not_applicable"].includes(String(value.availability)) ||
    !isNonNegativeInteger(value.input_tokens) || !isNonNegativeInteger(value.output_tokens) ||
    !isNonNegativeInteger(value.total_tokens)) {
    return false;
  }
  if (value.availability === "reported") {
    return [
      "openai_compatible_usage",
      "gemini_usage_metadata",
      "anthropic_usage",
      "huggingface_usage",
      "ollama_usage",
      "ollama_eval_counts",
    ].includes(value.source) && value.total_tokens === value.input_tokens + value.output_tokens;
  }
  return value.source === "" && value.input_tokens === 0 &&
    value.output_tokens === 0 && value.total_tokens === 0;
}

function isProviderRouteLineageDocument(value: GatewayRequestSummaryDocument | Record<string, unknown>): boolean {
  const configurationId = value.provider_route_configuration_id;
  const generation = value.provider_route_generation;
  const digest = value.provider_route_snapshot_digest;
  if (configurationId === undefined && generation === undefined && digest === undefined) return true;
  return typeof configurationId === "string" &&
    /^[A-Za-z0-9][A-Za-z0-9._:-]{0,159}$/u.test(configurationId) &&
    typeof generation === "number" && Number.isInteger(generation) && generation > 0 &&
    typeof digest === "string" && /^sha256:[a-f0-9]{64}$/u.test(digest);
}

function assertNoForbiddenFields(value: unknown, path = "response") {
  if (Array.isArray(value)) {
    value.forEach((item, index) => assertNoForbiddenFields(item, `${path}[${index}]`));
    return;
  }
  if (!isRecord(value)) return;
  for (const [key, nested] of Object.entries(value)) {
    if (FORBIDDEN_FIELDS.has(key.toLowerCase())) throw new Error(`Gateway request history contains forbidden field ${path}.${key}`);
    assertNoForbiddenFields(nested, `${path}.${key}`);
  }
}

function emptyHistoryState(
  status: GatewayRequestHistoryState["status"],
  requestId: string,
  auditRef: string,
  failureSummary: string,
  failureCode = "",
): GatewayRequestHistoryState {
  return { status, requests: [], nextCursor: "", hasMore: false, requestId, auditRef, failureCode, failureSummary };
}

function normalizeBaseUrl(value: string): string {
  return value.trim().replace(/\/$/, "");
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function isNullableString(value: unknown): value is string | null {
  return value === null || typeof value === "string";
}

function isNonNegativeInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value >= 0;
}

function stringFields(value: Record<string, unknown>, keys: string[]): boolean {
  return keys.every((key) => typeof value[key] === "string");
}

function hasOnlyKeys(value: Record<string, unknown>, allowedKeys: string[]): boolean {
  const allowed = new Set(allowedKeys);
  return Object.keys(value).every((key) => allowed.has(key));
}

function matchesOptionalApplication(value: unknown, expected: string): boolean {
  return expected ? value === expected : value === undefined || value === "";
}

function costAvailabilityMatchesUsage(
  costAvailability: GatewayRequestCostAvailability,
  usageAvailability: GatewayRequestHistorySummary["usageAvailability"],
): boolean {
  if (costAvailability === "legacy_not_captured") return true;
  if (costAvailability === "not_applicable") return usageAvailability === "not_applicable";
  if (costAvailability === "usage_not_reported") return usageAvailability === "not_reported";
  return usageAvailability === "reported";
}
