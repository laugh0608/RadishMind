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

export type GatewayRequestHistorySummary = {
  schemaVersion: "gateway_request_record.v1";
  recordVersion: number;
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
  httpStatusCode: number;
  failureCode: string;
  failureBoundary: string;
  usageAvailability: "reported" | "not_reported" | "not_applicable";
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
  usageSource: string;
  inputTokens: number;
  outputTokens: number;
  totalTokens: number;
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
  schema_version: "gateway_request_record.v1";
  record_version: number;
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
  http_status_code: number;
  failure_code: string;
  failure_boundary: string;
  usage_availability: GatewayRequestHistorySummary["usageAvailability"];
  stale_started: boolean;
};

type GatewayRequestDetailDocument = Omit<GatewayRequestSummaryDocument, "usage_availability"> & {
  tenant_ref: string;
  workspace_id: string;
  consumer_ref: string;
  application_id?: string;
  subject_ref: string;
  gateway_duration_ms: number;
  gateway_duration_available: boolean;
  usage: {
    availability: GatewayRequestHistorySummary["usageAvailability"];
    source: string;
    input_tokens: number;
    output_tokens: number;
    total_tokens: number;
  };
};

type GatewayRequestListEnvelope = {
  request_id: string;
  requests: GatewayRequestSummaryDocument[];
  next_cursor: string;
  has_more: boolean;
  failure_code: string | null;
  failure_summary: string;
  audit_ref: string;
};

type GatewayRequestReadEnvelope = {
  request_id: string;
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
  if (!response.ok || !isGatewayRequestListEnvelope(body)) {
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
  if (!response.ok || !isGatewayRequestReadEnvelope(body)) {
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
    httpStatusCode: value.http_status_code,
    failureCode: value.failure_code,
    failureBoundary: value.failure_boundary,
    usageAvailability: value.usage_availability,
    staleStarted: value.stale_started,
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
    usageSource: value.usage.source,
    inputTokens: value.usage.input_tokens,
    outputTokens: value.usage.output_tokens,
    totalTokens: value.usage.total_tokens,
  };
}

function isGatewayRequestListEnvelope(value: unknown): value is GatewayRequestListEnvelope {
  if (!isRecord(value)) return false;
  return typeof value.request_id === "string" && Array.isArray(value.requests) &&
    value.requests.every(isGatewayRequestSummaryDocument) && typeof value.next_cursor === "string" &&
    typeof value.has_more === "boolean" && isNullableString(value.failure_code) &&
    typeof value.failure_summary === "string" && typeof value.audit_ref === "string";
}

function isGatewayRequestReadEnvelope(value: unknown): value is GatewayRequestReadEnvelope {
  if (!isRecord(value)) return false;
  return typeof value.request_id === "string" && (value.request === null || isGatewayRequestDetailDocument(value.request)) &&
    isNullableString(value.failure_code) && typeof value.failure_summary === "string" && typeof value.audit_ref === "string";
}

function isGatewayRequestSummaryDocument(value: unknown): value is GatewayRequestSummaryDocument {
  if (!isRecord(value)) return false;
  return value.schema_version === "gateway_request_record.v1" && isNonNegativeInteger(value.record_version) &&
    stringFields(value, ["request_id", "audit_ref", "route", "protocol", "started_at", "completed_at", "selection_source", "selected_provider", "selected_profile", "selected_model", "failure_code", "failure_boundary"]) &&
    typeof value.stream === "boolean" && ["started", "succeeded", "failed", "canceled"].includes(String(value.status)) &&
    isNonNegativeInteger(value.duration_ms) && isNonNegativeInteger(value.provider_duration_ms) &&
    typeof value.provider_duration_available === "boolean" && isNonNegativeInteger(value.http_status_code) &&
    ["reported", "not_reported", "not_applicable"].includes(String(value.usage_availability)) && typeof value.stale_started === "boolean";
}

function isGatewayRequestDetailDocument(value: unknown): value is GatewayRequestDetailDocument {
  if (!isRecord(value) || !isRecord(value.usage)) return false;
  return isGatewayRequestSummaryDocument({ ...value, usage_availability: value.usage.availability }) &&
    stringFields(value, ["tenant_ref", "workspace_id", "consumer_ref", "subject_ref"]) &&
    (value.application_id === undefined || typeof value.application_id === "string") &&
    isNonNegativeInteger(value.gateway_duration_ms) && typeof value.gateway_duration_available === "boolean" &&
    stringFields(value.usage, ["availability", "source"]) &&
    ["reported", "not_reported", "not_applicable"].includes(String(value.usage.availability)) &&
    isNonNegativeInteger(value.usage.input_tokens) && isNonNegativeInteger(value.usage.output_tokens) &&
    isNonNegativeInteger(value.usage.total_tokens);
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
