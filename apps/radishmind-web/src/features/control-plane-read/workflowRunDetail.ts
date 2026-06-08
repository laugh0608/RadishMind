import {
  CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS,
  CONTROL_PLANE_READ_ROUTE_DEFINITIONS,
  CONTROL_PLANE_READ_ROUTES,
  controlPlaneReadResponseHasForbiddenOutput,
} from "../../../../../contracts/typescript/control-plane-read-api";
import type { WorkspaceRunRecordRow } from "./workspaceRunHistory";

export type WorkflowRunDetailTimelineEvent = {
  eventId: string;
  label: string;
  state: "queued" | "running" | "tool_preview_blocked" | "completed" | "failed";
  recordedAt: string;
  summary: string;
  auditRef: string;
};

export type WorkflowRunDetailSummary = {
  label: string;
  summary: string;
  fields: string[];
};

export type WorkflowRunDetailGuardPreview = {
  guardId: string;
  label: string;
  status: "blocked";
  reason: string;
  missingPrerequisite: string;
  auditRef: string;
};

export type WorkflowRunDetailViewModel = {
  pageId: "workflow-run-detail-read";
  sourcePageId: "workspace-run-history";
  sourceRouteId: "run-record-summary-list-route";
  draftRouteId: "run-detail-read-draft";
  routePath: typeof CONTROL_PLANE_READ_ROUTES.runs;
  requestId: string;
  auditRef: string;
  runId: string;
  workflowDefinitionId: string;
  applicationRef: string;
  status: string;
  failureCode: string;
  traceId: string;
  startedAt: string;
  completedAt: string;
  stateTimeline: WorkflowRunDetailTimelineEvent[];
  inputSummary: WorkflowRunDetailSummary;
  outputSummary: WorkflowRunDetailSummary;
  costSummary: WorkflowRunDetailSummary;
  tokenSummary: WorkflowRunDetailSummary;
  auditRefs: string[];
  blockedResultPreview: WorkflowRunDetailGuardPreview;
  blockedReplayPreview: WorkflowRunDetailGuardPreview;
  forbiddenProjectionBlocked: boolean;
  canRenderRunDetail: boolean;
  canRequestLiveBackend: false;
  canMutate: false;
  canStartRun: false;
  canCancelRun: false;
  canResumeRun: false;
  canReplayRun: false;
  canMaterializeResult: false;
  canWriteBusinessTruth: false;
};

const DEFAULT_RUN_SUMMARY: WorkspaceRunRecordRow = {
  runId: "run_radishflow_copilot_20260531_001",
  workflowDefinitionId: "wf_radishflow_copilot_latest",
  applicationRef: "app_flow_copilot",
  status: "succeeded",
  failureCode: "none",
  estimatedCost: "USD 0.08",
  traceId: "trace_run_radishflow_copilot_001",
  startedAt: "2026-05-31T10:31:00Z",
  completedAt: "2026-05-31T10:31:03Z",
};

export function buildWorkflowRunDetailViewModel(
  run: WorkspaceRunRecordRow = DEFAULT_RUN_SUMMARY,
): WorkflowRunDetailViewModel {
  const route = CONTROL_PLANE_READ_ROUTE_DEFINITIONS["run-record-summary-list-route"];
  const routePath = CONTROL_PLANE_READ_ROUTES.runs;
  const stateTimeline = buildStateTimeline(run);
  const auditRefs = [
    "audit_run_detail_started_demo",
    "audit_run_detail_model_response_demo",
    "audit_run_detail_blocked_result_demo",
  ];
  const forbiddenProjectionBlocked = controlPlaneReadResponseHasForbiddenOutput({
    detail: { [CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS[9]]: "blocked" },
  });

  return {
    pageId: "workflow-run-detail-read",
    sourcePageId: "workspace-run-history",
    sourceRouteId: "run-record-summary-list-route",
    draftRouteId: "run-detail-read-draft",
    routePath,
    requestId: "req_workflow_run_detail_demo",
    auditRef: "audit_workflow_run_detail_demo",
    runId: run.runId,
    workflowDefinitionId: run.workflowDefinitionId,
    applicationRef: run.applicationRef,
    status: run.status,
    failureCode: run.failureCode,
    traceId: run.traceId,
    startedAt: run.startedAt,
    completedAt: run.completedAt,
    stateTimeline,
    inputSummary: {
      label: "Input summary",
      summary: "Tenant, application, workflow definition, selected artifacts, and operator intent are shown as sanitized run inputs.",
      fields: ["tenant_ref", "application_ref", "workflow_definition_id", "selection_summary"],
    },
    outputSummary: {
      label: "Output summary",
      summary: "The run output remains advisory and read-only; raw prompts, raw tool payloads, and writeback payloads stay blocked.",
      fields: ["answer_summary", "candidate_actions", "risk_summary", "audit_refs"],
    },
    costSummary: {
      label: "Cost snapshot",
      summary: "Estimated committed summary only; no billing ledger or quota enforcement is implemented by this surface.",
      fields: [run.estimatedCost, "billing_not_ready", "quota_enforcement_not_ready"],
    },
    tokenSummary: {
      label: "Token snapshot",
      summary: "Static read-side token estimate for product shape; no durable result store or materialized result reader is available.",
      fields: ["input_tokens_estimate: 1840", "output_tokens_estimate: 620", "total_tokens_estimate: 2460"],
    },
    auditRefs,
    blockedResultPreview: {
      guardId: "guard_materialized_result_reader",
      label: "Materialized result reader",
      status: "blocked",
      reason: "Only summarized run output is available before durable result storage is implemented.",
      missingPrerequisite: "durable result store implementation gate",
      auditRef: "audit_run_detail_blocked_result_demo",
    },
    blockedReplayPreview: {
      guardId: "guard_replay_resume",
      label: "Replay and resume",
      status: "blocked",
      reason: "Run replay and resume remain disabled until executor, confirmation, and durable run store gates exist.",
      missingPrerequisite: "workflow executor implementation gate",
      auditRef: "audit_run_detail_blocked_replay_demo",
    },
    forbiddenProjectionBlocked,
    canRenderRunDetail:
      route.path === routePath &&
      route.canMutate === false &&
      stateTimeline.length >= 3 &&
      run.runId.length > 0,
    canRequestLiveBackend: false,
    canMutate: false,
    canStartRun: false,
    canCancelRun: false,
    canResumeRun: false,
    canReplayRun: false,
    canMaterializeResult: false,
    canWriteBusinessTruth: false,
  };
}

function buildStateTimeline(run: WorkspaceRunRecordRow): WorkflowRunDetailTimelineEvent[] {
  return [
    {
      eventId: "event_run_queued",
      label: "Queued",
      state: "queued",
      recordedAt: run.startedAt,
      summary: "The run was accepted into the read-side fixture as an existing audited record.",
      auditRef: "audit_run_detail_queued_demo",
    },
    {
      eventId: "event_model_reasoning",
      label: "Model reasoning",
      state: "running",
      recordedAt: run.startedAt,
      summary: "Model-facing reasoning is represented only by sanitized summary fields.",
      auditRef: "audit_run_detail_model_response_demo",
    },
    {
      eventId: "event_tool_preview_blocked",
      label: "Tool preview blocked",
      state: "tool_preview_blocked",
      recordedAt: run.completedAt,
      summary: "Candidate tool execution stays blocked because confirmation and executor gates are not implemented.",
      auditRef: "audit_run_detail_tool_preview_demo",
    },
    {
      eventId: "event_run_terminal",
      label: "Terminal state",
      state: run.status === "failed" ? "failed" : "completed",
      recordedAt: run.completedAt,
      summary:
        run.failureCode === "none"
          ? "The committed summary marks this run as completed without exposing a materialized result reader."
          : `The committed summary exposes failure code ${run.failureCode} without replay or resume controls.`,
      auditRef: "audit_run_detail_terminal_demo",
    },
  ];
}
