import {
  CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS,
  CONTROL_PLANE_READ_ROUTE_DEFINITIONS,
  CONTROL_PLANE_READ_ROUTES,
  controlPlaneReadResponseHasForbiddenOutput,
  toControlPlaneReadCollectionViewModel,
  type AuditSummary,
  type ControlPlaneReadCollectionViewModel,
  type ControlPlaneReadResponseByRoute,
} from "../../../../../contracts/typescript/control-plane-read-api";

export type AdminAuditLogStateId =
  | "ready"
  | "empty"
  | "denied"
  | "stale"
  | "partial_failure"
  | "forbidden_projection";

export type AdminAuditLogMetric = {
  label: string;
  value: string;
  detail: string;
};

export type AdminAuditEventRow = {
  auditRef: string;
  actorSubjectRef: string;
  eventKind: string;
  resourceRef: string;
  decision: string;
  failureCode: string;
  traceId: string;
  recordedAt: string;
};

export type AdminAuditLogStatePreview = {
  id: AdminAuditLogStateId;
  label: string;
  status: string;
  summary: string;
  itemCount: number;
  failureCode: string;
};

export type AdminAuditLogViewModel = {
  pageId: "admin-audit-log";
  routeId: "audit-summary-list-route";
  routePath: typeof CONTROL_PLANE_READ_ROUTES.audit;
  readModel: "audit-summary";
  requiredScope: "audit:read";
  collection: ControlPlaneReadCollectionViewModel;
  auditEvents: AdminAuditEventRow[];
  metrics: AdminAuditLogMetric[];
  statePreviews: AdminAuditLogStatePreview[];
  auditRef: string;
  requestId: string;
  nextCursor: string | null;
  forbiddenProjectionBlocked: boolean;
  canRenderAuditLog: boolean;
  canRequestLiveBackend: boolean;
  canMutate: false;
  canDeleteAuditRecord: false;
  canEditAuditRecord: false;
  canExportRawPayload: false;
  canRevealSecrets: false;
};

export function buildAdminAuditLogViewModel(
  collectionOverride?: ControlPlaneReadCollectionViewModel,
): AdminAuditLogViewModel {
  const route = CONTROL_PLANE_READ_ROUTE_DEFINITIONS["audit-summary-list-route"];
  const collection =
    collectionOverride ??
    toControlPlaneReadCollectionViewModel("audit-summary-list-route", buildAuditLogEnvelope());
  const auditEvents = collection.items.map((item) => toAuditEventRow(item as AuditSummary));
  const forbiddenProjectionBlocked = controlPlaneReadResponseHasForbiddenOutput({
    items: [{ [CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS[5]]: "blocked" }],
  });

  return {
    pageId: "admin-audit-log",
    routeId: "audit-summary-list-route",
    routePath: CONTROL_PLANE_READ_ROUTES.audit,
    readModel: "audit-summary",
    requiredScope: "audit:read",
    collection,
    auditEvents,
    metrics: buildMetrics(collection, auditEvents),
    statePreviews: buildStatePreviews(collection, forbiddenProjectionBlocked),
    auditRef: collection.auditRef,
    requestId: collection.requestId,
    nextCursor: collection.nextCursor,
    forbiddenProjectionBlocked,
    canRenderAuditLog:
      route.path === CONTROL_PLANE_READ_ROUTES.audit &&
      route.canMutate === false &&
      collection.canRenderItems &&
      !controlPlaneReadResponseHasForbiddenOutput(collection),
    canRequestLiveBackend: collection.devLiveHttpEnabled,
    canMutate: false,
    canDeleteAuditRecord: false,
    canEditAuditRecord: false,
    canExportRawPayload: false,
    canRevealSecrets: false,
  };
}

function buildAuditLogEnvelope(): ControlPlaneReadResponseByRoute["audit-summary-list-route"] {
  return {
    request_id: "req_admin_audit_log_demo",
    tenant_ref: "tenant_demo",
    items: [
      {
        audit_ref: "audit_admin_policy_read_20260601_001",
        tenant_ref: "tenant_demo",
        actor_subject_ref: "subject_admin_ops",
        event_kind: "control_plane.read",
        resource_ref: "tenant_demo",
        decision: "allowed",
        failure_code: null,
        trace_id: "trace_audit_admin_policy_read_001",
        recorded_at: "2026-06-01T09:14:00Z",
      },
      {
        audit_ref: "audit_cross_tenant_denied_20260601_002",
        tenant_ref: "tenant_demo",
        actor_subject_ref: "subject_workspace_reader",
        event_kind: "control_plane.denied",
        resource_ref: "tenant_other",
        decision: "denied",
        failure_code: "scope_denied",
        trace_id: "trace_audit_cross_tenant_denied_002",
        recorded_at: "2026-06-01T09:10:00Z",
      },
    ],
    next_cursor: "cursor_admin_audit_log_next",
    failure_code: null,
    audit_ref: "audit_admin_audit_log_demo",
  };
}

function toAuditEventRow(auditEvent: AuditSummary): AdminAuditEventRow {
  return {
    auditRef: auditEvent.audit_ref,
    actorSubjectRef: auditEvent.actor_subject_ref,
    eventKind: auditEvent.event_kind,
    resourceRef: auditEvent.resource_ref,
    decision: auditEvent.decision,
    failureCode: auditEvent.failure_code ?? "none",
    traceId: auditEvent.trace_id,
    recordedAt: auditEvent.recorded_at,
  };
}

function buildMetrics(
  collection: ControlPlaneReadCollectionViewModel,
  auditEvents: AdminAuditEventRow[],
): AdminAuditLogMetric[] {
  const deniedCount = auditEvents.filter((event) => event.decision === "denied").length;
  const failureCount = auditEvents.filter((event) => event.failureCode !== "none").length;

  return [
    {
      label: "Audit events",
      value: String(collection.itemCount),
      detail: collection.tenantRef,
    },
    {
      label: "Cursor",
      value: collection.nextCursor ? "available" : "none",
      detail: collection.nextCursor ?? "single page",
    },
    {
      label: "Denied",
      value: String(deniedCount),
      detail: "decision summary only",
    },
    {
      label: "Failures",
      value: String(failureCount),
      detail: "failure code display only",
    },
  ];
}

function buildStatePreviews(
  collection: ControlPlaneReadCollectionViewModel,
  forbiddenProjectionBlocked: boolean,
): AdminAuditLogStatePreview[] {
  return [
    {
      id: "ready",
      label: "Ready",
      status: collection.statusLabel,
      summary: "Audit summaries render from the validated read consumer view model.",
      itemCount: collection.itemCount,
      failureCode: collection.failureCode ?? "none",
    },
    {
      id: "empty",
      label: "Empty",
      status: "empty",
      summary: "The page keeps route metadata visible when no audit events are returned.",
      itemCount: 0,
      failureCode: "none",
    },
    {
      id: "denied",
      label: "Denied",
      status: "scope_denied",
      summary: "Audit scope denial renders without exposing partial audit rows.",
      itemCount: 0,
      failureCode: "scope_denied",
    },
    {
      id: "stale",
      label: "Stale",
      status: "stale",
      summary: "Cached audit summaries stay visibly separated from live audit storage.",
      itemCount: collection.itemCount,
      failureCode: "read_store_unavailable",
    },
    {
      id: "partial_failure",
      label: "Partial failure",
      status: "partial_failure",
      summary: "Cursor and failure metadata stay explicit when a later audit page cannot be read.",
      itemCount: collection.itemCount,
      failureCode: "read_store_unavailable",
    },
    {
      id: "forbidden_projection",
      label: "Forbidden projection",
      status: forbiddenProjectionBlocked ? "blocked" : "clear",
      summary: "Cookie and secret projections are blocked before audit rows can render them.",
      itemCount: 0,
      failureCode: forbiddenProjectionBlocked ? "read_store_contract_mismatch" : "none",
    },
  ];
}
