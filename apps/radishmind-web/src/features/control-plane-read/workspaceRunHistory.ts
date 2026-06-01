import {
  CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS,
  CONTROL_PLANE_READ_ROUTE_DEFINITIONS,
  CONTROL_PLANE_READ_ROUTES,
  controlPlaneReadResponseHasForbiddenOutput,
  toControlPlaneReadCollectionViewModel,
  type ControlPlaneReadCollectionViewModel,
  type ControlPlaneReadResponseByRoute,
  type RunRecordSummary,
} from "../../../../../contracts/typescript/control-plane-read-api";

export type WorkspaceRunHistoryStateId =
  | "ready"
  | "empty"
  | "denied"
  | "stale"
  | "partial_failure"
  | "forbidden_projection";

export type WorkspaceRunHistoryMetric = {
  label: string;
  value: string;
  detail: string;
};

export type WorkspaceRunRecordRow = {
  runId: string;
  workflowDefinitionId: string;
  applicationRef: string;
  status: string;
  failureCode: string;
  estimatedCost: string;
  traceId: string;
  startedAt: string;
  completedAt: string;
};

export type WorkspaceRunHistoryStatePreview = {
  id: WorkspaceRunHistoryStateId;
  label: string;
  status: string;
  summary: string;
  itemCount: number;
  failureCode: string;
};

export type WorkspaceRunHistoryViewModel = {
  pageId: "workspace-run-history";
  routeId: "run-record-summary-list-route";
  routePath: typeof CONTROL_PLANE_READ_ROUTES.runs;
  readModel: "run-record-summary";
  requiredScope: "runs:read";
  collection: ControlPlaneReadCollectionViewModel;
  runs: WorkspaceRunRecordRow[];
  metrics: WorkspaceRunHistoryMetric[];
  statePreviews: WorkspaceRunHistoryStatePreview[];
  auditRef: string;
  requestId: string;
  nextCursor: string | null;
  forbiddenProjectionBlocked: boolean;
  canRenderRuns: boolean;
  canRequestLiveBackend: boolean;
  canMutate: false;
  canStartRun: false;
  canCancelRun: false;
  canResumeRun: false;
  canReplayRun: false;
  canMaterializeResult: false;
  canWriteBusinessTruth: false;
};

export function buildWorkspaceRunHistoryViewModel(
  collectionOverride?: ControlPlaneReadCollectionViewModel,
): WorkspaceRunHistoryViewModel {
  const route = CONTROL_PLANE_READ_ROUTE_DEFINITIONS["run-record-summary-list-route"];
  const collection =
    collectionOverride ??
    toControlPlaneReadCollectionViewModel("run-record-summary-list-route", buildRunHistoryEnvelope());
  const runs = collection.items.map((item) => toRunRecordRow(item as RunRecordSummary));
  const forbiddenProjectionBlocked = controlPlaneReadResponseHasForbiddenOutput({
    items: [{ [CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS[8]]: "blocked" }],
  });

  return {
    pageId: "workspace-run-history",
    routeId: "run-record-summary-list-route",
    routePath: CONTROL_PLANE_READ_ROUTES.runs,
    readModel: "run-record-summary",
    requiredScope: "runs:read",
    collection,
    runs,
    metrics: buildMetrics(collection, runs),
    statePreviews: buildStatePreviews(collection, forbiddenProjectionBlocked),
    auditRef: collection.auditRef,
    requestId: collection.requestId,
    nextCursor: collection.nextCursor,
    forbiddenProjectionBlocked,
    canRenderRuns:
      route.path === CONTROL_PLANE_READ_ROUTES.runs &&
      route.canMutate === false &&
      collection.canRenderItems &&
      !controlPlaneReadResponseHasForbiddenOutput(collection),
    canRequestLiveBackend: collection.devLiveHttpEnabled,
    canMutate: false,
    canStartRun: false,
    canCancelRun: false,
    canResumeRun: false,
    canReplayRun: false,
    canMaterializeResult: false,
    canWriteBusinessTruth: false,
  };
}

function buildRunHistoryEnvelope(): ControlPlaneReadResponseByRoute["run-record-summary-list-route"] {
  return {
    request_id: "req_workspace_run_history_demo",
    tenant_ref: "tenant_demo",
    items: [
      {
        run_id: "run_radishflow_copilot_20260531_001",
        tenant_ref: "tenant_demo",
        workflow_definition_id: "wf_radishflow_copilot_latest",
        application_ref: "app_flow_copilot",
        status: "succeeded",
        failure_code: null,
        cost_summary: {
          estimated_cost: 0.08,
          currency: "USD",
        },
        trace_id: "trace_run_radishflow_copilot_001",
        started_at: "2026-05-31T10:31:00Z",
        completed_at: "2026-05-31T10:31:03Z",
      },
      {
        run_id: "run_radish_docs_qa_20260531_002",
        tenant_ref: "tenant_demo",
        workflow_definition_id: "wf_radish_docs_qa_draft",
        application_ref: "app_docs_assistant",
        status: "failed",
        failure_code: "read_store_unavailable",
        cost_summary: {
          estimated_cost: 0.02,
          currency: "USD",
        },
        trace_id: "trace_run_radish_docs_qa_002",
        started_at: "2026-05-31T10:12:00Z",
        completed_at: "2026-05-31T10:12:01Z",
      },
    ],
    next_cursor: "cursor_workspace_run_history_next",
    failure_code: null,
    audit_ref: "audit_workspace_run_history_demo",
  };
}

function toRunRecordRow(run: RunRecordSummary): WorkspaceRunRecordRow {
  return {
    runId: run.run_id,
    workflowDefinitionId: run.workflow_definition_id,
    applicationRef: run.application_ref,
    status: run.status,
    failureCode: run.failure_code ?? "none",
    estimatedCost: formatCost(run.cost_summary.estimated_cost, run.cost_summary.currency),
    traceId: run.trace_id,
    startedAt: run.started_at,
    completedAt: run.completed_at ?? "still running",
  };
}

function buildMetrics(
  collection: ControlPlaneReadCollectionViewModel,
  runs: WorkspaceRunRecordRow[],
): WorkspaceRunHistoryMetric[] {
  const succeededCount = runs.filter((run) => run.status === "succeeded").length;
  const failedCount = runs.filter((run) => run.failureCode !== "none").length;

  return [
    {
      label: "Runs",
      value: String(collection.itemCount),
      detail: collection.tenantRef,
    },
    {
      label: "Cursor",
      value: collection.nextCursor ? "available" : "none",
      detail: collection.nextCursor ?? "single page",
    },
    {
      label: "Succeeded",
      value: String(succeededCount),
      detail: "read-only status summary",
    },
    {
      label: "Failures",
      value: String(failedCount),
      detail: "failure code display only",
    },
  ];
}

function buildStatePreviews(
  collection: ControlPlaneReadCollectionViewModel,
  forbiddenProjectionBlocked: boolean,
): WorkspaceRunHistoryStatePreview[] {
  return [
    {
      id: "ready",
      label: "Ready",
      status: collection.statusLabel,
      summary: "Run record summaries render from the offline consumer view model.",
      itemCount: collection.itemCount,
      failureCode: collection.failureCode ?? "none",
    },
    {
      id: "empty",
      label: "Empty",
      status: "empty",
      summary: "The page keeps route metadata visible when no run records are returned.",
      itemCount: 0,
      failureCode: "none",
    },
    {
      id: "denied",
      label: "Denied",
      status: "scope_denied",
      summary: "Run scope denial renders without exposing partial run rows.",
      itemCount: 0,
      failureCode: "scope_denied",
    },
    {
      id: "stale",
      label: "Stale",
      status: "stale",
      summary: "Cached run history stays visibly separated from live workflow executor data.",
      itemCount: collection.itemCount,
      failureCode: "read_store_unavailable",
    },
    {
      id: "partial_failure",
      label: "Partial failure",
      status: "partial_failure",
      summary: "Cursor and failure metadata stay explicit when a later run page cannot be read.",
      itemCount: collection.itemCount,
      failureCode: "read_store_unavailable",
    },
    {
      id: "forbidden_projection",
      label: "Forbidden projection",
      status: forbiddenProjectionBlocked ? "blocked" : "clear",
      summary: "Business writeback projections are blocked before run rows can render them.",
      itemCount: 0,
      failureCode: forbiddenProjectionBlocked ? "read_store_contract_mismatch" : "none",
    },
  ];
}

function formatCost(value: number, currency: string): string {
  return `${currency} ${value.toFixed(2)}`;
}
