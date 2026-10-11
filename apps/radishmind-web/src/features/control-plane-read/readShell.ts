import type { WorkflowProjectionCopy } from "./workflowProjectionCopy.ts";
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
} from "../../../../../contracts/typescript/control-plane-read-api.ts";

export type ControlPlaneReadRouteCard = WorkflowProjectionCopy & {
  routeId: ControlPlaneReadRouteId;
  label: string;
  surface: "admin" | "workspace";
  path: string;
  requiredScope: string;
  readModel: string;
  pagination: string;
};

export type ControlPlaneReadStatePreview = WorkflowProjectionCopy & {
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
    lockedCapabilities: ["writeback locked", "unrestricted executor locked", "production database locked", "production auth pending"],
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
      labelMessage: { key: "readShell_loading_label" },
      status: "pending",
      tone: "neutral",
      summary: "Read route request is in flight; no action path is exposed.",
      summaryMessage: { key: "readShell_loading_summary" },
      itemCount: 0,
      failureCode: null,
      auditRef: "pending",
    },
    {
      id: "ready",
      label: "Ready",
      labelMessage: { key: "readShell_ready_label" },
      status: readyPreview.statusLabel,
      tone: "good",
      summary: "Envelope passed the shared read model and can render sanitized summary items.",
      summaryMessage: { key: "readShell_ready_summary" },
      itemCount: readyPreview.itemCount,
      failureCode: readyPreview.failureCode,
      auditRef: readyPreview.auditRef,
    },
    {
      id: "empty",
      label: "Empty",
      labelMessage: { key: "readShell_empty_label" },
      status: "ready",
      tone: "neutral",
      summary: "A successful read can render zero items without implying missing authorization.",
      summaryMessage: { key: "readShell_empty_summary" },
      itemCount: 0,
      failureCode: null,
      auditRef: "audit_demo_empty",
    },
    {
      id: "denied",
      label: "Denied",
      labelMessage: { key: "readShell_denied_label" },
      status: deniedPreview.statusLabel,
      tone: "bad",
      summary: "Failure views preserve failure code and audit reference while rendering no items.",
      summaryMessage: { key: "readShell_denied_summary" },
      itemCount: deniedPreview.itemCount,
      failureCode: deniedPreview.failureCode,
      auditRef: deniedPreview.auditRef,
    },
    {
      id: "stale",
      label: "Stale",
      labelMessage: { key: "readShell_stale_label" },
      status: "stale",
      tone: "neutral",
      summary: "Stale state can keep the previous read-only snapshot and diagnostic timestamp.",
      summaryMessage: { key: "readShell_stale_summary" },
      itemCount: readyPreview.itemCount,
      failureCode: null,
      auditRef: readyPreview.auditRef,
    },
    {
      id: "partial_failure",
      label: "Partial failure",
      labelMessage: { key: "readShell_partial_failure_label" },
      status: "degraded",
      tone: "bad",
      summary: "A route can fail without enabling fallback write, replay, or materialized result access.",
      summaryMessage: { key: "readShell_partial_failure_summary" },
      itemCount: 0,
      failureCode: "read_store_unavailable",
      auditRef: "audit_demo_partial_failure",
    },
    {
      id: "forbidden_projection",
      label: "Forbidden projection",
      labelMessage: { key: "readShell_forbidden_projection_label" },
      status: forbiddenProjectionBlocked ? "blocked" : "clear",
      tone: forbiddenProjectionBlocked ? "bad" : "good",
      summary: "Sensitive keys are stopped before any summary item is rendered.",
      summaryMessage: { key: "readShell_forbidden_projection_summary" },
      itemCount: 0,
      failureCode: forbiddenProjectionBlocked ? "read_store_contract_mismatch" : null,
      auditRef: "audit_demo_forbidden_projection",
    },
  ];
}
