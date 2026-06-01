export const CONTROL_PLANE_READ_ROUTES = {
  tenantSummary: "/v1/control-plane/tenants/{tenant_ref}/summary",
  applications: "/v1/user-workspace/applications",
  apiKeys: "/v1/user-workspace/api-keys",
  quotaSummary: "/v1/user-workspace/usage/quota-summary",
  workflowDefinitions: "/v1/user-workspace/workflow-definitions",
  runs: "/v1/user-workspace/runs",
  audit: "/v1/control-plane/audit",
} as const;

export const CONTROL_PLANE_READ_ROUTE_IDS = [
  "tenant-summary-route",
  "application-summary-list-route",
  "api-key-summary-list-route",
  "quota-summary-route",
  "workflow-definition-summary-list-route",
  "run-record-summary-list-route",
  "audit-summary-list-route",
] as const;

export const CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS = [
  "raw_secret_value",
  "api_key_value",
  "api_key_hash",
  "authorization_header",
  "bearer_token",
  "cookie_value",
  "raw_request_body_dump",
  "raw_tool_payload",
  "business_writeback_payload",
  "full_prompt_dump_with_secret",
] as const;

export type ControlPlaneReadRouteId = (typeof CONTROL_PLANE_READ_ROUTE_IDS)[number];

export type ControlPlaneReadPath =
  (typeof CONTROL_PLANE_READ_ROUTES)[keyof typeof CONTROL_PLANE_READ_ROUTES];

export type ControlPlaneReadMethod = "GET";

export type ControlPlaneReadModel =
  | "tenant-summary"
  | "application-summary"
  | "api-key-summary"
  | "quota-summary"
  | "workflow-definition-summary"
  | "run-record-summary"
  | "audit-summary";

export type ControlPlaneReadScope =
  | "tenant:read"
  | "applications:read"
  | "api_keys:read"
  | "usage:read"
  | "runs:read"
  | "audit:read";

export type ControlPlaneReadFailureCode =
  | "identity_context_missing"
  | "tenant_binding_missing"
  | "scope_denied"
  | "tenant_not_found"
  | "quota_policy_missing"
  | "invalid_filter"
  | "read_store_unavailable"
  | "read_store_contract_mismatch"
  | "database_read_disabled"
  | "auth_context_contract_mismatch";

export type ControlPlaneReadPagination = "single_resource" | "cursor_required";

export type ControlPlaneReadConsumerSource = "offline_fixture" | "dev_live_http";

export type ControlPlaneReadAllowedFilter =
  | "application_kind"
  | "owner_subject_ref"
  | "last_run_status"
  | "state"
  | "scope"
  | "period"
  | "application_ref"
  | "definition_status"
  | "risk_level"
  | "workflow_definition_id"
  | "status"
  | "failure_code"
  | "event_kind"
  | "resource_ref"
  | "actor_subject_ref";

export type ControlPlaneReadListParameters = {
  limit?: number;
  cursor?: string;
  sort?: string;
  tenant_ref?: string;
};

export type ApplicationSummaryFilters = ControlPlaneReadListParameters & {
  application_kind?: string;
  owner_subject_ref?: string;
  last_run_status?: string;
};

export type APIKeySummaryFilters = ControlPlaneReadListParameters & {
  state?: string;
  owner_subject_ref?: string;
  scope?: string;
};

export type QuotaSummaryFilters = {
  period?: string;
};

export type WorkflowDefinitionSummaryFilters = ControlPlaneReadListParameters & {
  application_ref?: string;
  definition_status?: string;
  risk_level?: string;
};

export type RunRecordSummaryFilters = ControlPlaneReadListParameters & {
  application_ref?: string;
  workflow_definition_id?: string;
  status?: string;
  failure_code?: string;
};

export type AuditSummaryFilters = ControlPlaneReadListParameters & {
  event_kind?: string;
  resource_ref?: string;
  actor_subject_ref?: string;
  failure_code?: string;
};

export type ControlPlaneReadRouteDefinition = {
  routeId: ControlPlaneReadRouteId;
  method: ControlPlaneReadMethod;
  path: ControlPlaneReadPath;
  readModel: ControlPlaneReadModel;
  requiredScope: ControlPlaneReadScope;
  pagination: ControlPlaneReadPagination;
  allowedFilters: readonly ControlPlaneReadAllowedFilter[];
  tenantScoped: true;
  authRequired: true;
  fakeStoreBacked: true;
  databaseBacked: false;
  canMutate: false;
};

export type ControlPlaneReadEnvelope<TItem extends ControlPlaneReadSummaryItem> = {
  request_id: string;
  tenant_ref: string;
  items: TItem[];
  next_cursor: string | null;
  failure_code: ControlPlaneReadFailureCode | null;
  audit_ref: string;
};

export type TenantSummary = {
  tenant_ref: string;
  tenant_display_name: string;
  tenant_state: string;
  plan_ref: string;
  quota_summary_ref: string;
  deployment_status_ref: string;
  audit_summary_ref: string;
};

export type ApplicationSummary = {
  application_ref: string;
  tenant_ref: string;
  application_kind: string;
  display_name: string;
  owner_subject_ref: string;
  latest_workflow_definition_ref: string;
  last_run_status: string;
  updated_at: string;
};

export type APIKeySummary = {
  api_key_id: string;
  tenant_ref: string;
  owner_subject_ref: string;
  scopes: string[];
  state: string;
  created_at: string;
  expires_at: string | null;
  last_used_at: string | null;
};

export type QuotaSummary = {
  quota_id: string;
  tenant_ref: string;
  period: string;
  request_limit: number;
  token_limit: number;
  cost_limit: number;
  usage_snapshot: {
    request_count: number;
    token_count: number;
    estimated_cost: number;
  };
  over_quota_failure_code: string;
};

export type WorkflowDefinitionSummary = {
  workflow_definition_id: string;
  tenant_ref: string;
  application_ref: string;
  version: number;
  definition_status: string;
  node_count: number;
  risk_level: string;
  requires_confirmation_capable: boolean;
  updated_at: string;
};

export type RunRecordSummary = {
  run_id: string;
  tenant_ref: string;
  workflow_definition_id: string;
  application_ref: string;
  status: string;
  failure_code: string | null;
  cost_summary: {
    estimated_cost: number;
    currency: string;
  };
  trace_id: string;
  started_at: string;
  completed_at: string | null;
};

export type AuditSummary = {
  audit_ref: string;
  tenant_ref: string;
  actor_subject_ref: string;
  event_kind: string;
  resource_ref: string;
  decision: "allowed" | "denied" | string;
  failure_code: string | null;
  trace_id: string;
  recorded_at: string;
};

export type ControlPlaneReadSummaryItem =
  | TenantSummary
  | ApplicationSummary
  | APIKeySummary
  | QuotaSummary
  | WorkflowDefinitionSummary
  | RunRecordSummary
  | AuditSummary;

export type ControlPlaneReadResponseByRoute = {
  "tenant-summary-route": ControlPlaneReadEnvelope<TenantSummary>;
  "application-summary-list-route": ControlPlaneReadEnvelope<ApplicationSummary>;
  "api-key-summary-list-route": ControlPlaneReadEnvelope<APIKeySummary>;
  "quota-summary-route": ControlPlaneReadEnvelope<QuotaSummary>;
  "workflow-definition-summary-list-route": ControlPlaneReadEnvelope<WorkflowDefinitionSummary>;
  "run-record-summary-list-route": ControlPlaneReadEnvelope<RunRecordSummary>;
  "audit-summary-list-route": ControlPlaneReadEnvelope<AuditSummary>;
};

export type ControlPlaneReadRequestByRoute = {
  "tenant-summary-route": {
    route_id: "tenant-summary-route";
    method: "GET";
    path: typeof CONTROL_PLANE_READ_ROUTES.tenantSummary;
    path_params: { tenant_ref: string };
    query?: Record<string, never>;
  };
  "application-summary-list-route": {
    route_id: "application-summary-list-route";
    method: "GET";
    path: typeof CONTROL_PLANE_READ_ROUTES.applications;
    query?: ApplicationSummaryFilters;
  };
  "api-key-summary-list-route": {
    route_id: "api-key-summary-list-route";
    method: "GET";
    path: typeof CONTROL_PLANE_READ_ROUTES.apiKeys;
    query?: APIKeySummaryFilters;
  };
  "quota-summary-route": {
    route_id: "quota-summary-route";
    method: "GET";
    path: typeof CONTROL_PLANE_READ_ROUTES.quotaSummary;
    query?: QuotaSummaryFilters;
  };
  "workflow-definition-summary-list-route": {
    route_id: "workflow-definition-summary-list-route";
    method: "GET";
    path: typeof CONTROL_PLANE_READ_ROUTES.workflowDefinitions;
    query?: WorkflowDefinitionSummaryFilters;
  };
  "run-record-summary-list-route": {
    route_id: "run-record-summary-list-route";
    method: "GET";
    path: typeof CONTROL_PLANE_READ_ROUTES.runs;
    query?: RunRecordSummaryFilters;
  };
  "audit-summary-list-route": {
    route_id: "audit-summary-list-route";
    method: "GET";
    path: typeof CONTROL_PLANE_READ_ROUTES.audit;
    query?: AuditSummaryFilters;
  };
};

export type ControlPlaneReadRouteCatalogViewModel = {
  routes: ControlPlaneReadRouteDefinition[];
  allRoutesFakeStoreBacked: true;
  allRoutesRequireAuth: true;
  allRoutesReadOnly: true;
  databaseBacked: false;
  formalUiReady: false;
};

export type ControlPlaneReadCollectionViewModel = {
  routeId: ControlPlaneReadRouteId;
  readModel: ControlPlaneReadModel;
  source: ControlPlaneReadConsumerSource;
  requestId: string;
  tenantRef: string;
  items: ControlPlaneReadSummaryItem[];
  itemCount: number;
  nextCursor: string | null;
  failureCode: ControlPlaneReadFailureCode | null;
  auditRef: string;
  statusLabel: "ready" | "denied";
  canRenderItems: boolean;
  canFetchNextPage: boolean;
  noSideEffectsExpected: true;
  devLiveHttpEnabled: boolean;
  productionApiConsumer: false;
  canExecuteWorkflow: false;
  canWriteBusinessTruth: false;
  canRevealSecrets: false;
};

export const CONTROL_PLANE_READ_ROUTE_DEFINITIONS: Record<
  ControlPlaneReadRouteId,
  ControlPlaneReadRouteDefinition
> = {
  "tenant-summary-route": {
    routeId: "tenant-summary-route",
    method: "GET",
    path: CONTROL_PLANE_READ_ROUTES.tenantSummary,
    readModel: "tenant-summary",
    requiredScope: "tenant:read",
    pagination: "single_resource",
    allowedFilters: [],
    tenantScoped: true,
    authRequired: true,
    fakeStoreBacked: true,
    databaseBacked: false,
    canMutate: false,
  },
  "application-summary-list-route": {
    routeId: "application-summary-list-route",
    method: "GET",
    path: CONTROL_PLANE_READ_ROUTES.applications,
    readModel: "application-summary",
    requiredScope: "applications:read",
    pagination: "cursor_required",
    allowedFilters: ["application_kind", "owner_subject_ref", "last_run_status"],
    tenantScoped: true,
    authRequired: true,
    fakeStoreBacked: true,
    databaseBacked: false,
    canMutate: false,
  },
  "api-key-summary-list-route": {
    routeId: "api-key-summary-list-route",
    method: "GET",
    path: CONTROL_PLANE_READ_ROUTES.apiKeys,
    readModel: "api-key-summary",
    requiredScope: "api_keys:read",
    pagination: "cursor_required",
    allowedFilters: ["state", "owner_subject_ref", "scope"],
    tenantScoped: true,
    authRequired: true,
    fakeStoreBacked: true,
    databaseBacked: false,
    canMutate: false,
  },
  "quota-summary-route": {
    routeId: "quota-summary-route",
    method: "GET",
    path: CONTROL_PLANE_READ_ROUTES.quotaSummary,
    readModel: "quota-summary",
    requiredScope: "usage:read",
    pagination: "single_resource",
    allowedFilters: ["period"],
    tenantScoped: true,
    authRequired: true,
    fakeStoreBacked: true,
    databaseBacked: false,
    canMutate: false,
  },
  "workflow-definition-summary-list-route": {
    routeId: "workflow-definition-summary-list-route",
    method: "GET",
    path: CONTROL_PLANE_READ_ROUTES.workflowDefinitions,
    readModel: "workflow-definition-summary",
    requiredScope: "applications:read",
    pagination: "cursor_required",
    allowedFilters: ["application_ref", "definition_status", "risk_level"],
    tenantScoped: true,
    authRequired: true,
    fakeStoreBacked: true,
    databaseBacked: false,
    canMutate: false,
  },
  "run-record-summary-list-route": {
    routeId: "run-record-summary-list-route",
    method: "GET",
    path: CONTROL_PLANE_READ_ROUTES.runs,
    readModel: "run-record-summary",
    requiredScope: "runs:read",
    pagination: "cursor_required",
    allowedFilters: ["application_ref", "workflow_definition_id", "status", "failure_code"],
    tenantScoped: true,
    authRequired: true,
    fakeStoreBacked: true,
    databaseBacked: false,
    canMutate: false,
  },
  "audit-summary-list-route": {
    routeId: "audit-summary-list-route",
    method: "GET",
    path: CONTROL_PLANE_READ_ROUTES.audit,
    readModel: "audit-summary",
    requiredScope: "audit:read",
    pagination: "cursor_required",
    allowedFilters: ["event_kind", "resource_ref", "actor_subject_ref", "failure_code"],
    tenantScoped: true,
    authRequired: true,
    fakeStoreBacked: true,
    databaseBacked: false,
    canMutate: false,
  },
};

export function listControlPlaneReadRouteCatalog(): ControlPlaneReadRouteCatalogViewModel {
  return {
    routes: CONTROL_PLANE_READ_ROUTE_IDS.map((routeId) => CONTROL_PLANE_READ_ROUTE_DEFINITIONS[routeId]),
    allRoutesFakeStoreBacked: true,
    allRoutesRequireAuth: true,
    allRoutesReadOnly: true,
    databaseBacked: false,
    formalUiReady: false,
  };
}

export function isControlPlaneReadEnvelope(
  value: unknown,
): value is ControlPlaneReadEnvelope<ControlPlaneReadSummaryItem> {
  if (!isRecord(value)) {
    return false;
  }
  return (
    typeof value.request_id === "string" &&
    typeof value.tenant_ref === "string" &&
    Array.isArray(value.items) &&
    (typeof value.next_cursor === "string" || value.next_cursor === null) &&
    (typeof value.failure_code === "string" || value.failure_code === null) &&
    typeof value.audit_ref === "string"
  );
}

export function toControlPlaneReadCollectionViewModel(
  routeId: ControlPlaneReadRouteId,
  response: ControlPlaneReadEnvelope<ControlPlaneReadSummaryItem>,
  options: { source?: ControlPlaneReadConsumerSource } = {},
): ControlPlaneReadCollectionViewModel {
  const route = CONTROL_PLANE_READ_ROUTE_DEFINITIONS[routeId];
  const denied = response.failure_code !== null;
  const source = options.source ?? "offline_fixture";
  return {
    routeId,
    readModel: route.readModel,
    source,
    requestId: response.request_id,
    tenantRef: response.tenant_ref,
    items: response.items,
    itemCount: response.items.length,
    nextCursor: response.next_cursor,
    failureCode: response.failure_code,
    auditRef: response.audit_ref,
    statusLabel: denied ? "denied" : "ready",
    canRenderItems: !denied && response.items.length > 0,
    canFetchNextPage: !denied && response.next_cursor !== null,
    noSideEffectsExpected: true,
    devLiveHttpEnabled: source === "dev_live_http",
    productionApiConsumer: false,
    canExecuteWorkflow: false,
    canWriteBusinessTruth: false,
    canRevealSecrets: false,
  };
}

export function controlPlaneReadResponseHasForbiddenOutput(value: unknown): boolean {
  return hasAnyKey(value, CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS);
}

function hasAnyKey(value: unknown, forbiddenKeys: readonly string[]): boolean {
  if (Array.isArray(value)) {
    return value.some((item) => hasAnyKey(item, forbiddenKeys));
  }
  if (!isRecord(value)) {
    return false;
  }
  for (const [key, nested] of Object.entries(value)) {
    if (forbiddenKeys.includes(key)) {
      return true;
    }
    if (hasAnyKey(nested, forbiddenKeys)) {
      return true;
    }
  }
  return false;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
