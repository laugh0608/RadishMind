import type { WorkflowProjectionCopy } from "./workflowProjectionCopy.ts";
import {
  CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS,
  CONTROL_PLANE_READ_ROUTE_DEFINITIONS,
  CONTROL_PLANE_READ_ROUTES,
  controlPlaneReadResponseHasForbiddenOutput,
  toControlPlaneReadCollectionViewModel,
  type ControlPlaneReadCollectionViewModel,
  type ControlPlaneReadResponseByRoute,
  type WorkflowDefinitionSummary,
} from "../../../../../contracts/typescript/control-plane-read-api.ts";

export type WorkspaceWorkflowDefinitionsStateId =
  | "ready"
  | "empty"
  | "denied"
  | "stale"
  | "partial_failure"
  | "forbidden_projection";

export type WorkspaceWorkflowDefinitionsMetric = WorkflowProjectionCopy & {
  label: string;
  value: string;
  detail: string;
};

export type WorkspaceWorkflowDefinitionRow = {
  workflowDefinitionId: string;
  applicationRef: string;
  version: number;
  definitionStatus: string;
  nodeCount: number;
  riskLevel: string;
  requiresConfirmationCapable: boolean;
  updatedAt: string;
};

export type WorkspaceWorkflowDefinitionsStatePreview = WorkflowProjectionCopy & {
  id: WorkspaceWorkflowDefinitionsStateId;
  label: string;
  status: string;
  summary: string;
  itemCount: number;
  failureCode: string;
};

export type WorkspaceWorkflowDefinitionsViewModel = {
  pageId: "workspace-workflow-definitions";
  routeId: "workflow-definition-summary-list-route";
  routePath: typeof CONTROL_PLANE_READ_ROUTES.workflowDefinitions;
  readModel: "workflow-definition-summary";
  requiredScope: "applications:read";
  collection: ControlPlaneReadCollectionViewModel;
  workflowDefinitions: WorkspaceWorkflowDefinitionRow[];
  metrics: WorkspaceWorkflowDefinitionsMetric[];
  statePreviews: WorkspaceWorkflowDefinitionsStatePreview[];
  auditRef: string;
  requestId: string;
  nextCursor: string | null;
  forbiddenProjectionBlocked: boolean;
  canRenderWorkflowDefinitions: boolean;
  canRequestLiveBackend: boolean;
  canMutate: false;
  canCreateWorkflow: false;
  canEditWorkflow: false;
  canRunWorkflow: false;
  canConfirmWorkflowAction: false;
};

export function buildWorkspaceWorkflowDefinitionsViewModel(
  collectionOverride?: ControlPlaneReadCollectionViewModel,
): WorkspaceWorkflowDefinitionsViewModel {
  const route = CONTROL_PLANE_READ_ROUTE_DEFINITIONS["workflow-definition-summary-list-route"];
  const collection =
    collectionOverride ??
    toControlPlaneReadCollectionViewModel(
      "workflow-definition-summary-list-route",
      buildWorkflowDefinitionsEnvelope(),
    );
  const workflowDefinitions = collection.items.map((item) =>
    toWorkflowDefinitionRow(item as WorkflowDefinitionSummary),
  );
  const forbiddenProjectionBlocked = controlPlaneReadResponseHasForbiddenOutput({
    items: [{ [CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS[6]]: "blocked" }],
  });

  return {
    pageId: "workspace-workflow-definitions",
    routeId: "workflow-definition-summary-list-route",
    routePath: CONTROL_PLANE_READ_ROUTES.workflowDefinitions,
    readModel: "workflow-definition-summary",
    requiredScope: "applications:read",
    collection,
    workflowDefinitions,
    metrics: buildMetrics(collection, workflowDefinitions),
    statePreviews: buildStatePreviews(collection, forbiddenProjectionBlocked),
    auditRef: collection.auditRef,
    requestId: collection.requestId,
    nextCursor: collection.nextCursor,
    forbiddenProjectionBlocked,
    canRenderWorkflowDefinitions:
      route.path === CONTROL_PLANE_READ_ROUTES.workflowDefinitions &&
      route.canMutate === false &&
      collection.canRenderItems &&
      !controlPlaneReadResponseHasForbiddenOutput(collection),
    canRequestLiveBackend: collection.devLiveHttpEnabled,
    canMutate: false,
    canCreateWorkflow: false,
    canEditWorkflow: false,
    canRunWorkflow: false,
    canConfirmWorkflowAction: false,
  };
}

function buildWorkflowDefinitionsEnvelope(): ControlPlaneReadResponseByRoute["workflow-definition-summary-list-route"] {
  return {
    request_id: "req_workspace_workflow_definitions_demo",
    tenant_ref: "tenant_demo",
    items: [
      {
        workflow_definition_id: "wf_radishflow_copilot_latest",
        tenant_ref: "tenant_demo",
        application_ref: "app_flow_copilot",
        version: 3,
        definition_status: "published",
        node_count: 8,
        risk_level: "medium",
        requires_confirmation_capable: true,
        updated_at: "2026-05-31T10:25:00Z",
      },
      {
        workflow_definition_id: "wf_radish_docs_qa_draft",
        tenant_ref: "tenant_demo",
        application_ref: "app_docs_assistant",
        version: 1,
        definition_status: "draft",
        node_count: 4,
        risk_level: "low",
        requires_confirmation_capable: false,
        updated_at: "2026-05-31T09:05:00Z",
      },
    ],
    next_cursor: "cursor_workspace_workflow_definitions_next",
    failure_code: null,
    audit_ref: "audit_workspace_workflow_definitions_demo",
  };
}

function toWorkflowDefinitionRow(
  workflowDefinition: WorkflowDefinitionSummary,
): WorkspaceWorkflowDefinitionRow {
  return {
    workflowDefinitionId: workflowDefinition.workflow_definition_id,
    applicationRef: workflowDefinition.application_ref,
    version: workflowDefinition.version,
    definitionStatus: workflowDefinition.definition_status,
    nodeCount: workflowDefinition.node_count,
    riskLevel: workflowDefinition.risk_level,
    requiresConfirmationCapable: workflowDefinition.requires_confirmation_capable,
    updatedAt: workflowDefinition.updated_at,
  };
}

function buildMetrics(
  collection: ControlPlaneReadCollectionViewModel,
  workflowDefinitions: WorkspaceWorkflowDefinitionRow[],
): WorkspaceWorkflowDefinitionsMetric[] {
  const publishedCount = workflowDefinitions.filter(
    (workflowDefinition) => workflowDefinition.definitionStatus === "published",
  ).length;
  const confirmationCapableCount = workflowDefinitions.filter(
    (workflowDefinition) => workflowDefinition.requiresConfirmationCapable,
  ).length;

  return [
    {
      label: "Definitions",
      labelMessage: { key: "workspaceWorkflowDefinitions_buildMetrics_label" },
      value: String(collection.itemCount),
      detail: collection.tenantRef,
    },
    {
      label: "Cursor",
      labelMessage: { key: "workspaceWorkflowDefinitions_buildMetrics_label_2" },
      value: collection.nextCursor ? "available" : "none",
      valueMessage: collection.nextCursor ? { key: "workspaceWorkflowDefinitions_buildMetrics_value_yes" } : { key: "workspaceWorkflowDefinitions_buildMetrics_value_no" },
      detail: collection.nextCursor ?? "single page",
      detailMessage: collection.nextCursor ? undefined : { key: "readSinglePage" },
    },
    {
      label: "Published",
      labelMessage: { key: "workspaceWorkflowDefinitions_buildMetrics_label_3" },
      value: String(publishedCount),
      detail: "read-only lifecycle state",
      detailMessage: { key: "workspaceWorkflowDefinitions_buildMetrics_detail" },
    },
    {
      label: "Confirmation capable",
      labelMessage: { key: "workspaceWorkflowDefinitions_buildMetrics_label_4" },
      value: String(confirmationCapableCount),
      detail: "capability marker only",
      detailMessage: { key: "workspaceWorkflowDefinitions_buildMetrics_detail_2" },
    },
  ];
}

function buildStatePreviews(
  collection: ControlPlaneReadCollectionViewModel,
  forbiddenProjectionBlocked: boolean,
): WorkspaceWorkflowDefinitionsStatePreview[] {
  return [
    {
      id: "ready",
      label: "Ready",
      labelMessage: { key: "workspaceWorkflowDefinitions_ready_label" },
      status: collection.statusLabel,
      summary: "Workflow definition summaries render from the offline consumer view model.",
      summaryMessage: { key: "workspaceWorkflowDefinitions_ready_summary" },
      itemCount: collection.itemCount,
      failureCode: collection.failureCode ?? "none",
    },
    {
      id: "empty",
      label: "Empty",
      labelMessage: { key: "workspaceWorkflowDefinitions_empty_label" },
      status: "empty",
      summary: "The page keeps route metadata visible when no workflow definitions are returned.",
      summaryMessage: { key: "workspaceWorkflowDefinitions_empty_summary" },
      itemCount: 0,
      failureCode: "none",
    },
    {
      id: "denied",
      label: "Denied",
      labelMessage: { key: "workspaceWorkflowDefinitions_denied_label" },
      status: "tenant_binding_missing",
      summary: "Tenant binding failure renders without exposing partial workflow definition rows.",
      summaryMessage: { key: "workspaceWorkflowDefinitions_denied_summary" },
      itemCount: 0,
      failureCode: "tenant_binding_missing",
    },
    {
      id: "stale",
      label: "Stale",
      labelMessage: { key: "workspaceWorkflowDefinitions_stale_label" },
      status: "stale",
      summary: "Cached definitions stay visibly separated from live workflow builder data.",
      summaryMessage: { key: "workspaceWorkflowDefinitions_stale_summary" },
      itemCount: collection.itemCount,
      failureCode: "read_store_unavailable",
    },
    {
      id: "partial_failure",
      label: "Partial failure",
      labelMessage: { key: "workspaceWorkflowDefinitions_partial_failure_label" },
      status: "partial_failure",
      summary: "Cursor and failure metadata stay explicit when a later definition page cannot be read.",
      summaryMessage: { key: "workspaceWorkflowDefinitions_partial_failure_summary" },
      itemCount: collection.itemCount,
      failureCode: "read_store_unavailable",
    },
    {
      id: "forbidden_projection",
      label: "Forbidden projection",
      labelMessage: { key: "workspaceWorkflowDefinitions_forbidden_projection_label" },
      status: forbiddenProjectionBlocked ? "blocked" : "clear",
      summary: "Raw tool payload projections are blocked before workflow definition rows can render them.",
      summaryMessage: { key: "workspaceWorkflowDefinitions_forbidden_projection_summary" },
      itemCount: 0,
      failureCode: forbiddenProjectionBlocked ? "read_store_contract_mismatch" : "none",
    },
  ];
}
