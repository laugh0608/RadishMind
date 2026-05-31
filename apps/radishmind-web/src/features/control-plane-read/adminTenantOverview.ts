import {
  CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS,
  CONTROL_PLANE_READ_ROUTE_DEFINITIONS,
  CONTROL_PLANE_READ_ROUTES,
  controlPlaneReadResponseHasForbiddenOutput,
  toControlPlaneReadCollectionViewModel,
  type ControlPlaneReadCollectionViewModel,
  type ControlPlaneReadResponseByRoute,
  type TenantSummary,
} from "../../../../../contracts/typescript/control-plane-read-api";

export type AdminTenantOverviewStateId = "ready" | "empty" | "denied" | "stale" | "forbidden_projection";

export type AdminTenantOverviewFact = {
  label: string;
  value: string;
  detail: string;
};

export type AdminTenantOverviewStatePreview = {
  id: AdminTenantOverviewStateId;
  label: string;
  status: string;
  summary: string;
  itemCount: number;
  failureCode: string;
};

export type AdminTenantOverviewViewModel = {
  pageId: "admin-tenant-overview";
  routeId: "tenant-summary-route";
  routePath: typeof CONTROL_PLANE_READ_ROUTES.tenantSummary;
  readModel: "tenant-summary";
  requiredScope: "tenant:read";
  tenant: TenantSummary | null;
  collection: ControlPlaneReadCollectionViewModel;
  facts: AdminTenantOverviewFact[];
  statePreviews: AdminTenantOverviewStatePreview[];
  auditRef: string;
  requestId: string;
  forbiddenProjectionBlocked: boolean;
  canRenderTenant: boolean;
  canRequestLiveBackend: false;
  canMutate: false;
};

export function buildAdminTenantOverviewViewModel(): AdminTenantOverviewViewModel {
  const route = CONTROL_PLANE_READ_ROUTE_DEFINITIONS["tenant-summary-route"];
  const collection = toControlPlaneReadCollectionViewModel("tenant-summary-route", buildTenantSummaryEnvelope());
  const tenant = collection.items[0] as TenantSummary | undefined;
  const forbiddenProjectionBlocked = controlPlaneReadResponseHasForbiddenOutput({
    items: [{ [CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS[1]]: "blocked" }],
  });

  return {
    pageId: "admin-tenant-overview",
    routeId: "tenant-summary-route",
    routePath: CONTROL_PLANE_READ_ROUTES.tenantSummary,
    readModel: "tenant-summary",
    requiredScope: "tenant:read",
    tenant: tenant ?? null,
    collection,
    facts: tenant ? buildTenantFacts(tenant) : [],
    statePreviews: buildStatePreviews(collection, forbiddenProjectionBlocked),
    auditRef: collection.auditRef,
    requestId: collection.requestId,
    forbiddenProjectionBlocked,
    canRenderTenant:
      route.path === CONTROL_PLANE_READ_ROUTES.tenantSummary &&
      route.canMutate === false &&
      collection.canRenderItems &&
      !controlPlaneReadResponseHasForbiddenOutput(collection),
    canRequestLiveBackend: false,
    canMutate: false,
  };
}

function buildTenantSummaryEnvelope(): ControlPlaneReadResponseByRoute["tenant-summary-route"] {
  return {
    request_id: "req_admin_tenant_overview_demo",
    tenant_ref: "tenant_demo",
    items: [
      {
        tenant_ref: "tenant_demo",
        tenant_display_name: "Demo tenant",
        tenant_state: "active",
        plan_ref: "plan_internal",
        quota_summary_ref: "quota_demo_current",
        deployment_status_ref: "deployment_local_read_only",
        audit_summary_ref: "audit_demo_latest",
      },
    ],
    next_cursor: null,
    failure_code: null,
    audit_ref: "audit_admin_tenant_overview_demo",
  };
}

function buildTenantFacts(tenant: TenantSummary): AdminTenantOverviewFact[] {
  return [
    {
      label: "Tenant state",
      value: tenant.tenant_state,
      detail: tenant.tenant_ref,
    },
    {
      label: "Plan",
      value: tenant.plan_ref,
      detail: "read-only plan reference",
    },
    {
      label: "Quota",
      value: tenant.quota_summary_ref,
      detail: "quota summary route remains separate",
    },
    {
      label: "Deployment",
      value: tenant.deployment_status_ref,
      detail: "status reference only",
    },
    {
      label: "Audit",
      value: tenant.audit_summary_ref,
      detail: "latest audit summary reference",
    },
  ];
}

function buildStatePreviews(
  collection: ControlPlaneReadCollectionViewModel,
  forbiddenProjectionBlocked: boolean,
): AdminTenantOverviewStatePreview[] {
  return [
    {
      id: "ready",
      label: "Ready",
      status: collection.statusLabel,
      summary: "Tenant summary can render from the offline consumer view model.",
      itemCount: collection.itemCount,
      failureCode: collection.failureCode ?? "none",
    },
    {
      id: "empty",
      label: "Empty",
      status: "empty",
      summary: "The page keeps the tenant frame visible when no summary item is available.",
      itemCount: 0,
      failureCode: "none",
    },
    {
      id: "denied",
      label: "Denied",
      status: "scope_denied",
      summary: "The page renders a denial state instead of exposing partial tenant data.",
      itemCount: 0,
      failureCode: "scope_denied",
    },
    {
      id: "stale",
      label: "Stale",
      status: "stale",
      summary: "Cached contract data remains visibly separated from live control-plane data.",
      itemCount: collection.itemCount,
      failureCode: "read_store_unavailable",
    },
    {
      id: "forbidden_projection",
      label: "Forbidden projection",
      status: forbiddenProjectionBlocked ? "blocked" : "clear",
      summary: "Sensitive output keys are blocked before the tenant overview can render them.",
      itemCount: 0,
      failureCode: forbiddenProjectionBlocked ? "forbidden_projection" : "none",
    },
  ];
}
