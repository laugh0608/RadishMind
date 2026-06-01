import {
  CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS,
  CONTROL_PLANE_READ_ROUTE_DEFINITIONS,
  CONTROL_PLANE_READ_ROUTES,
  controlPlaneReadResponseHasForbiddenOutput,
  toControlPlaneReadCollectionViewModel,
  type ControlPlaneReadCollectionViewModel,
  type ControlPlaneReadResponseByRoute,
  type QuotaSummary,
} from "../../../../../contracts/typescript/control-plane-read-api";

export type WorkspaceUsageQuotaStateId =
  | "ready"
  | "empty"
  | "denied"
  | "stale"
  | "partial_failure"
  | "forbidden_projection";

export type WorkspaceUsageQuotaLimit = {
  label: string;
  value: string;
  used: string;
  detail: string;
};

export type WorkspaceUsageQuotaSnapshot = {
  label: string;
  value: string;
  detail: string;
};

export type WorkspaceUsageQuotaStatePreview = {
  id: WorkspaceUsageQuotaStateId;
  label: string;
  status: string;
  summary: string;
  itemCount: number;
  failureCode: string;
};

export type WorkspaceUsageQuotaViewModel = {
  pageId: "workspace-usage-quota";
  routeId: "quota-summary-route";
  routePath: typeof CONTROL_PLANE_READ_ROUTES.quotaSummary;
  readModel: "quota-summary";
  requiredScope: "usage:read";
  collection: ControlPlaneReadCollectionViewModel;
  quota: QuotaSummary | null;
  limits: WorkspaceUsageQuotaLimit[];
  usageSnapshot: WorkspaceUsageQuotaSnapshot[];
  statePreviews: WorkspaceUsageQuotaStatePreview[];
  auditRef: string;
  requestId: string;
  overQuotaFailureCode: string;
  forbiddenProjectionBlocked: boolean;
  canRenderQuota: boolean;
  canRequestLiveBackend: boolean;
  canMutate: false;
  canEnforceQuota: false;
  canWriteBillingLedger: false;
};

export function buildWorkspaceUsageQuotaViewModel(
  collectionOverride?: ControlPlaneReadCollectionViewModel,
): WorkspaceUsageQuotaViewModel {
  const route = CONTROL_PLANE_READ_ROUTE_DEFINITIONS["quota-summary-route"];
  const collection =
    collectionOverride ?? toControlPlaneReadCollectionViewModel("quota-summary-route", buildQuotaEnvelope());
  const quota = collection.items[0] as QuotaSummary | undefined;
  const forbiddenProjectionBlocked = controlPlaneReadResponseHasForbiddenOutput({
    items: [{ [CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS[8]]: "blocked" }],
  });

  return {
    pageId: "workspace-usage-quota",
    routeId: "quota-summary-route",
    routePath: CONTROL_PLANE_READ_ROUTES.quotaSummary,
    readModel: "quota-summary",
    requiredScope: "usage:read",
    collection,
    quota: quota ?? null,
    limits: quota ? buildLimits(quota) : [],
    usageSnapshot: quota ? buildUsageSnapshot(quota) : [],
    statePreviews: buildStatePreviews(collection, forbiddenProjectionBlocked),
    auditRef: collection.auditRef,
    requestId: collection.requestId,
    overQuotaFailureCode: quota?.over_quota_failure_code ?? "quota_policy_missing",
    forbiddenProjectionBlocked,
    canRenderQuota:
      route.path === CONTROL_PLANE_READ_ROUTES.quotaSummary &&
      route.canMutate === false &&
      collection.canRenderItems &&
      !controlPlaneReadResponseHasForbiddenOutput(collection),
    canRequestLiveBackend: collection.devLiveHttpEnabled,
    canMutate: false,
    canEnforceQuota: false,
    canWriteBillingLedger: false,
  };
}

function buildQuotaEnvelope(): ControlPlaneReadResponseByRoute["quota-summary-route"] {
  return {
    request_id: "req_workspace_usage_quota_demo",
    tenant_ref: "tenant_demo",
    items: [
      {
        quota_id: "quota_demo_current",
        tenant_ref: "tenant_demo",
        period: "monthly",
        request_limit: 10000,
        token_limit: 1000000,
        cost_limit: 100,
        usage_snapshot: {
          request_count: 12,
          token_count: 3456,
          estimated_cost: 0.42,
        },
        over_quota_failure_code: "quota_exceeded",
      },
    ],
    next_cursor: null,
    failure_code: null,
    audit_ref: "audit_workspace_usage_quota_demo",
  };
}

function buildLimits(quota: QuotaSummary): WorkspaceUsageQuotaLimit[] {
  return [
    {
      label: "Requests",
      value: quota.request_limit.toLocaleString("en-US"),
      used: quota.usage_snapshot.request_count.toLocaleString("en-US"),
      detail: `${usagePercent(quota.usage_snapshot.request_count, quota.request_limit)} used`,
    },
    {
      label: "Tokens",
      value: quota.token_limit.toLocaleString("en-US"),
      used: quota.usage_snapshot.token_count.toLocaleString("en-US"),
      detail: `${usagePercent(quota.usage_snapshot.token_count, quota.token_limit)} used`,
    },
    {
      label: "Cost",
      value: formatCost(quota.cost_limit),
      used: formatCost(quota.usage_snapshot.estimated_cost),
      detail: `${usagePercent(quota.usage_snapshot.estimated_cost, quota.cost_limit)} used`,
    },
  ];
}

function buildUsageSnapshot(quota: QuotaSummary): WorkspaceUsageQuotaSnapshot[] {
  return [
    {
      label: "Quota",
      value: quota.quota_id,
      detail: quota.tenant_ref,
    },
    {
      label: "Period",
      value: quota.period,
      detail: "read-only quota window",
    },
    {
      label: "Failure code",
      value: quota.over_quota_failure_code,
      detail: "reported only, not enforced by this page",
    },
    {
      label: "Audit",
      value: "available",
      detail: "request and audit refs remain visible",
    },
  ];
}

function buildStatePreviews(
  collection: ControlPlaneReadCollectionViewModel,
  forbiddenProjectionBlocked: boolean,
): WorkspaceUsageQuotaStatePreview[] {
  return [
    {
      id: "ready",
      label: "Ready",
      status: collection.statusLabel,
      summary: "Quota summary renders from the offline consumer view model without enforcement.",
      itemCount: collection.itemCount,
      failureCode: collection.failureCode ?? "none",
    },
    {
      id: "empty",
      label: "Empty",
      status: "empty",
      summary: "The page keeps route metadata visible when no quota policy summary is returned.",
      itemCount: 0,
      failureCode: "none",
    },
    {
      id: "denied",
      label: "Denied",
      status: "scope_denied",
      summary: "Usage scope denial renders without exposing partial quota or usage rows.",
      itemCount: 0,
      failureCode: "scope_denied",
    },
    {
      id: "stale",
      label: "Stale",
      status: "stale",
      summary: "Cached quota summaries stay visibly separated from live billing or gateway enforcement data.",
      itemCount: collection.itemCount,
      failureCode: "read_store_unavailable",
    },
    {
      id: "partial_failure",
      label: "Partial failure",
      status: "partial_failure",
      summary: "Failure metadata remains explicit when quota policy details cannot be fully read.",
      itemCount: collection.itemCount,
      failureCode: "quota_policy_missing",
    },
    {
      id: "forbidden_projection",
      label: "Forbidden projection",
      status: forbiddenProjectionBlocked ? "blocked" : "clear",
      summary: "Billing writeback payload projections are blocked by the shared forbidden output guard.",
      itemCount: 0,
      failureCode: forbiddenProjectionBlocked ? "read_store_contract_mismatch" : "none",
    },
  ];
}

function usagePercent(used: number, limit: number): string {
  if (limit <= 0) {
    return "0.0%";
  }
  return `${((used / limit) * 100).toFixed(1)}%`;
}

function formatCost(value: number): string {
  return `$${value.toFixed(2)}`;
}
