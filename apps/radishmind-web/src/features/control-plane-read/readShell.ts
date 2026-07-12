import {
  CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS,
  CONTROL_PLANE_READ_ROUTE_DEFINITIONS,
  CONTROL_PLANE_READ_ROUTES,
  controlPlaneReadResponseHasForbiddenOutput,
  isControlPlaneReadEnvelope,
  listControlPlaneReadRouteCatalog,
  toControlPlaneReadCollectionViewModel,
  type ControlPlaneReadCollectionViewModel,
  type ControlPlaneReadEnvelope,
  type ControlPlaneReadResponseByRoute,
  type ControlPlaneReadRouteCatalogViewModel,
  type ControlPlaneReadRouteId,
  type ControlPlaneReadSummaryItem,
} from "../../../../../contracts/typescript/control-plane-read-api";

export type ControlPlaneReadRouteCard = {
  routeId: ControlPlaneReadRouteId;
  label: string;
  surface: "admin" | "workspace";
  path: string;
  requiredScope: string;
  readModel: string;
  pagination: string;
};

export type ControlPlaneReadStatePreview = {
  id: "loading" | "ready" | "empty" | "denied" | "stale" | "partial_failure" | "forbidden_projection";
  label: string;
  status: string;
  tone: "good" | "bad" | "neutral";
  summary: string;
  itemCount: number;
  failureCode: string | null;
  auditRef: string;
};

export type ControlPlaneReadShellViewModel = {
  catalog: ControlPlaneReadRouteCatalogViewModel;
  routeCards: ControlPlaneReadRouteCard[];
  statePreviews: ControlPlaneReadStatePreview[];
  readyPreview: ControlPlaneReadCollectionViewModel;
  deniedPreview: ControlPlaneReadCollectionViewModel;
  forbiddenProjectionBlocked: boolean;
  forbiddenOutputKeys: readonly string[];
  lockedCapabilities: string[];
  usesCanonicalRoutes: boolean;
  offlineOnly: true;
};

export function buildControlPlaneReadShellViewModel(): ControlPlaneReadShellViewModel {
  const catalog = listControlPlaneReadRouteCatalog();
  const readyEnvelope = buildReadyEnvelope();
  const deniedEnvelope = buildDeniedEnvelope();
  const readyPreview = toControlPlaneReadCollectionViewModel("tenant-summary-route", readyEnvelope);
  const deniedPreview = toControlPlaneReadCollectionViewModel("tenant-summary-route", deniedEnvelope);
  const forbiddenProjectionBlocked = controlPlaneReadResponseHasForbiddenOutput({
    items: [
      {
        [CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS[0]]: "redacted",
      },
    ],
  });

  return {
    catalog,
    routeCards: catalog.routes.map((route) => ({
      routeId: route.routeId,
      label: route.routeId.replace(/-/g, " "),
      surface: route.path.startsWith(CONTROL_PLANE_READ_ROUTES.audit) ||
        route.path.startsWith("/v1/control-plane")
        ? "admin"
        : "workspace",
      path: route.path,
      requiredScope: route.requiredScope,
      readModel: route.readModel,
      pagination: route.pagination,
    })),
    statePreviews: buildStatePreviews(readyPreview, deniedPreview, forbiddenProjectionBlocked),
    readyPreview,
    deniedPreview,
    forbiddenProjectionBlocked,
    forbiddenOutputKeys: CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS,
    lockedCapabilities: ["writeback locked", "unrestricted executor locked", "database detached", "auth pending"],
    usesCanonicalRoutes:
      CONTROL_PLANE_READ_ROUTE_DEFINITIONS["tenant-summary-route"].path === CONTROL_PLANE_READ_ROUTES.tenantSummary &&
      isControlPlaneReadEnvelope(readyEnvelope),
    offlineOnly: true,
  };
}

function buildReadyEnvelope(): ControlPlaneReadResponseByRoute["tenant-summary-route"] {
  return {
    request_id: "ui-shell-ready-preview",
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
    audit_ref: "audit_demo_ready",
  };
}

function buildDeniedEnvelope(): ControlPlaneReadEnvelope<ControlPlaneReadSummaryItem> {
  return {
    request_id: "ui-shell-denied-preview",
    tenant_ref: "tenant_demo",
    items: [],
    next_cursor: null,
    failure_code: "scope_denied",
    audit_ref: "audit_demo_denied",
  };
}

function buildStatePreviews(
  readyPreview: ControlPlaneReadCollectionViewModel,
  deniedPreview: ControlPlaneReadCollectionViewModel,
  forbiddenProjectionBlocked: boolean,
): ControlPlaneReadStatePreview[] {
  return [
    {
      id: "loading",
      label: "Loading",
      status: "pending",
      tone: "neutral",
      summary: "Read route request is in flight; no action path is exposed.",
      itemCount: 0,
      failureCode: null,
      auditRef: "pending",
    },
    {
      id: "ready",
      label: "Ready",
      status: readyPreview.statusLabel,
      tone: "good",
      summary: "Envelope passed the shared read model and can render sanitized summary items.",
      itemCount: readyPreview.itemCount,
      failureCode: readyPreview.failureCode,
      auditRef: readyPreview.auditRef,
    },
    {
      id: "empty",
      label: "Empty",
      status: "ready",
      tone: "neutral",
      summary: "A successful read can render zero items without implying missing authorization.",
      itemCount: 0,
      failureCode: null,
      auditRef: "audit_demo_empty",
    },
    {
      id: "denied",
      label: "Denied",
      status: deniedPreview.statusLabel,
      tone: "bad",
      summary: "Failure views preserve failure code and audit reference while rendering no items.",
      itemCount: deniedPreview.itemCount,
      failureCode: deniedPreview.failureCode,
      auditRef: deniedPreview.auditRef,
    },
    {
      id: "stale",
      label: "Stale",
      status: "stale",
      tone: "neutral",
      summary: "Stale state can keep the previous read-only snapshot and diagnostic timestamp.",
      itemCount: readyPreview.itemCount,
      failureCode: null,
      auditRef: readyPreview.auditRef,
    },
    {
      id: "partial_failure",
      label: "Partial failure",
      status: "degraded",
      tone: "bad",
      summary: "A route can fail without enabling fallback write, replay, or materialized result access.",
      itemCount: 0,
      failureCode: "read_store_unavailable",
      auditRef: "audit_demo_partial_failure",
    },
    {
      id: "forbidden_projection",
      label: "Forbidden projection",
      status: forbiddenProjectionBlocked ? "blocked" : "clear",
      tone: forbiddenProjectionBlocked ? "bad" : "good",
      summary: "Sensitive keys are stopped before any summary item is rendered.",
      itemCount: 0,
      failureCode: forbiddenProjectionBlocked ? "read_store_contract_mismatch" : null,
      auditRef: "audit_demo_forbidden_projection",
    },
  ];
}
