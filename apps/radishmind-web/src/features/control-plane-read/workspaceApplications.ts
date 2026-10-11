import type { WorkflowProjectionCopy } from "./workflowProjectionCopy.ts";
import {
  CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS,
  CONTROL_PLANE_READ_ROUTE_DEFINITIONS,
  CONTROL_PLANE_READ_ROUTES,
  controlPlaneReadResponseHasForbiddenOutput,
  toControlPlaneReadCollectionViewModel,
  type ApplicationSummary,
  type ControlPlaneReadCollectionViewModel,
  type ControlPlaneReadResponseByRoute,
} from "../../../../../contracts/typescript/control-plane-read-api.ts";
import type { ApplicationCatalogRecord } from "./applicationCatalogConsumer.ts";

export type WorkspaceApplicationsStateId =
  | "ready"
  | "empty"
  | "denied"
  | "stale"
  | "partial_failure"
  | "forbidden_projection";

export type WorkspaceApplicationsMetric = WorkflowProjectionCopy & {
  id: "applications" | "cursor" | "blockedRuns" | "audit";
  label: string;
  value: string;
  detail: string;
};

export type WorkspaceApplicationRow = {
  applicationRef: string;
  displayName: string;
  applicationKind: string;
  ownerSubjectRef: string;
  latestWorkflowDefinitionRef: string;
  lastRunStatus: string;
  updatedAt: string;
  workspaceId?: string;
  description?: string;
  lifecycleState?: "active" | "archived";
  recordVersion?: number;
  createdAt?: string;
  archivedAt?: string | null;
};

export type WorkspaceApplicationCatalogProjection = WorkflowProjectionCopy & {
  records: ApplicationCatalogRecord[];
  summary: string;
};

export type WorkspaceApplicationsStatePreview = WorkflowProjectionCopy & {
  id: WorkspaceApplicationsStateId;
  label: string;
  status: string;
  summary: string;
  itemCount: number;
  failureCode: string;
};

export type WorkspaceApplicationsViewModel = {
  pageId: "workspace-applications";
  routeId: "application-summary-list-route";
  routePath: typeof CONTROL_PLANE_READ_ROUTES.applications;
  readModel: "application-summary";
  requiredScope: "applications:read";
  collection: ControlPlaneReadCollectionViewModel;
  applications: WorkspaceApplicationRow[];
  metrics: WorkspaceApplicationsMetric[];
  statePreviews: WorkspaceApplicationsStatePreview[];
  auditRef: string;
  requestId: string;
  nextCursor: string | null;
  forbiddenProjectionBlocked: boolean;
  canRenderApplications: boolean;
  canRequestLiveBackend: boolean;
  canMutate: false;
};

export function buildWorkspaceApplicationsViewModel(
  collectionOverride?: ControlPlaneReadCollectionViewModel,
  catalogProjection?: WorkspaceApplicationCatalogProjection,
): WorkspaceApplicationsViewModel {
  const route = CONTROL_PLANE_READ_ROUTE_DEFINITIONS["application-summary-list-route"];
  const collection =
    collectionOverride ??
    toControlPlaneReadCollectionViewModel("application-summary-list-route", buildApplicationsEnvelope());
  if (catalogProjection) {
    const applications = catalogProjection.records.map(toCatalogApplicationRow);
    const catalogCollection: ControlPlaneReadCollectionViewModel = {
      ...collection,
      source: "dev_live_http",
      items: [],
      itemCount: applications.length,
      nextCursor: null,
      failureCode: null,
      statusLabel: "ready",
      canRenderItems: true,
      canFetchNextPage: false,
      devLiveHttpEnabled: true,
    };
    return {
      pageId: "workspace-applications",
      routeId: "application-summary-list-route",
      routePath: CONTROL_PLANE_READ_ROUTES.applications,
      readModel: "application-summary",
      requiredScope: "applications:read",
      collection: catalogCollection,
      applications,
      metrics: buildMetrics(catalogCollection, applications),
      statePreviews: buildStatePreviews(catalogCollection, false),
      auditRef: "audit_application_catalog_snapshot",
      requestId: "request_application_catalog_snapshot",
      nextCursor: null,
      forbiddenProjectionBlocked: false,
      canRenderApplications: true,
      canRequestLiveBackend: true,
      canMutate: false,
    };
  }
  const applications = collection.items.map((item) => toApplicationRow(item as ApplicationSummary));
  const forbiddenProjectionBlocked = controlPlaneReadResponseHasForbiddenOutput({
    items: [{ [CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS[7]]: "blocked" }],
  });

  return {
    pageId: "workspace-applications",
    routeId: "application-summary-list-route",
    routePath: CONTROL_PLANE_READ_ROUTES.applications,
    readModel: "application-summary",
    requiredScope: "applications:read",
    collection,
    applications,
    metrics: buildMetrics(collection, applications),
    statePreviews: buildStatePreviews(collection, forbiddenProjectionBlocked),
    auditRef: collection.auditRef,
    requestId: collection.requestId,
    nextCursor: collection.nextCursor,
    forbiddenProjectionBlocked,
    canRenderApplications:
      route.path === CONTROL_PLANE_READ_ROUTES.applications &&
      route.canMutate === false &&
      collection.canRenderItems &&
      !controlPlaneReadResponseHasForbiddenOutput(collection),
    canRequestLiveBackend: collection.devLiveHttpEnabled,
    canMutate: false,
  };
}

function toCatalogApplicationRow(record: ApplicationCatalogRecord): WorkspaceApplicationRow {
  return {
    applicationRef: record.applicationId,
    displayName: record.displayName,
    applicationKind: record.applicationKind,
    ownerSubjectRef: record.ownerSubjectRef,
    latestWorkflowDefinitionRef: "",
    lastRunStatus: "not_available",
    updatedAt: record.updatedAt,
    workspaceId: record.workspaceId,
    description: record.description,
    lifecycleState: record.lifecycleState,
    recordVersion: record.recordVersion,
    createdAt: record.createdAt,
    archivedAt: record.archivedAt,
  };
}

function buildApplicationsEnvelope(): ControlPlaneReadResponseByRoute["application-summary-list-route"] {
  return {
    request_id: "req_workspace_applications_demo",
    tenant_ref: "tenant_demo",
    items: [
      {
        application_ref: "app_flow_copilot",
        tenant_ref: "tenant_demo",
        application_kind: "workflow_copilot",
        display_name: "RadishFlow Copilot",
        owner_subject_ref: "subject_platform_ops",
        latest_workflow_definition_ref: "wf_radishflow_copilot_latest",
        last_run_status: "succeeded",
        updated_at: "2026-05-31T10:20:00Z",
      },
      {
        application_ref: "app_docs_assistant",
        tenant_ref: "tenant_demo",
        application_kind: "docs_qa",
        display_name: "Radish Docs Assistant",
        owner_subject_ref: "subject_docs_team",
        latest_workflow_definition_ref: "wf_radish_docs_qa_draft",
        last_run_status: "blocked",
        updated_at: "2026-05-31T09:40:00Z",
      },
    ],
    next_cursor: "cursor_workspace_applications_next",
    failure_code: null,
    audit_ref: "audit_workspace_applications_demo",
  };
}

function toApplicationRow(application: ApplicationSummary): WorkspaceApplicationRow {
  return {
    applicationRef: application.application_ref,
    displayName: application.display_name,
    applicationKind: application.application_kind,
    ownerSubjectRef: application.owner_subject_ref,
    latestWorkflowDefinitionRef: application.latest_workflow_definition_ref,
    lastRunStatus: application.last_run_status,
    updatedAt: application.updated_at,
  };
}

function buildMetrics(
  collection: ControlPlaneReadCollectionViewModel,
  applications: WorkspaceApplicationRow[],
): WorkspaceApplicationsMetric[] {
  const blockedCount = applications.filter((application) => application.lastRunStatus === "blocked").length;
  return [
    {
      id: "applications",
      label: "Applications",
      labelMessage: { key: "workspaceApplications_applications_label" },
      value: String(collection.itemCount),
      detail: collection.tenantRef,
    },
    {
      id: "cursor",
      label: "Cursor",
      labelMessage: { key: "workspaceApplications_cursor_label" },
      value: collection.nextCursor ? "available" : "none",
      valueMessage: collection.nextCursor ? { key: "workspaceApplications_cursor_value_yes" } : { key: "workspaceApplications_cursor_value_no" },
      detail: collection.nextCursor ?? "single page",
    },
    {
      id: "blockedRuns",
      label: "Blocked runs",
      labelMessage: { key: "workspaceApplications_blockedRuns_label" },
      value: String(blockedCount),
      detail: "read-only status reference",
      detailMessage: { key: "workspaceApplications_blockedRuns_detail" },
    },
    {
      id: "audit",
      label: "Audit",
      labelMessage: { key: "workspaceApplications_audit_label" },
      value: collection.auditRef,
      detail: "latest list read audit reference",
      detailMessage: { key: "workspaceApplications_audit_detail" },
    },
  ];
}

function buildStatePreviews(
  collection: ControlPlaneReadCollectionViewModel,
  forbiddenProjectionBlocked: boolean,
): WorkspaceApplicationsStatePreview[] {
  return [
    {
      id: "ready",
      label: "Ready",
      labelMessage: { key: "workspaceApplications_ready_label" },
      status: collection.statusLabel,
      summary: "Application summaries render from the offline consumer view model.",
      summaryMessage: { key: "workspaceApplications_ready_summary" },
      itemCount: collection.itemCount,
      failureCode: collection.failureCode ?? "none",
    },
    {
      id: "empty",
      label: "Empty",
      labelMessage: { key: "workspaceApplications_empty_label" },
      status: "empty",
      summary: "The list keeps filters and route metadata visible when no applications are returned.",
      summaryMessage: { key: "workspaceApplications_empty_summary" },
      itemCount: 0,
      failureCode: "none",
    },
    {
      id: "denied",
      label: "Denied",
      labelMessage: { key: "workspaceApplications_denied_label" },
      status: "scope_denied",
      summary: "The page renders a denial state instead of exposing partial application rows.",
      summaryMessage: { key: "workspaceApplications_denied_summary" },
      itemCount: 0,
      failureCode: "scope_denied",
    },
    {
      id: "stale",
      label: "Stale",
      labelMessage: { key: "workspaceApplications_stale_label" },
      status: "stale",
      summary: "Cached application summaries remain visibly separated from live workspace data.",
      summaryMessage: { key: "workspaceApplications_stale_summary" },
      itemCount: collection.itemCount,
      failureCode: "read_store_unavailable",
    },
    {
      id: "partial_failure",
      label: "Partial failure",
      labelMessage: { key: "workspaceApplications_partial_failure_label" },
      status: "partial_failure",
      summary: "Cursor availability and failure metadata stay explicit when a later page cannot be read.",
      summaryMessage: { key: "workspaceApplications_partial_failure_summary" },
      itemCount: collection.itemCount,
      failureCode: "read_store_unavailable",
    },
    {
      id: "forbidden_projection",
      label: "Forbidden projection",
      labelMessage: { key: "workspaceApplications_forbidden_projection_label" },
      status: forbiddenProjectionBlocked ? "blocked" : "clear",
      summary: "Sensitive raw tool payload fields are blocked before application rows can render them.",
      summaryMessage: { key: "workspaceApplications_forbidden_projection_summary" },
      itemCount: 0,
      failureCode: forbiddenProjectionBlocked ? "forbidden_projection" : "none",
    },
  ];
}
