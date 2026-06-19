import {
  CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS,
  CONTROL_PLANE_READ_ROUTE_DEFINITIONS,
  CONTROL_PLANE_READ_ROUTES,
  controlPlaneReadResponseHasForbiddenOutput,
} from "../../../../../contracts/typescript/control-plane-read-api";
import type { WorkflowDefinitionBlockedActionPreview } from "./workflowDefinitionDetail";
import type { WorkflowRunDetailGuardPreview } from "./workflowRunDetail";

export type WorkflowBlockedActionRequirement = {
  requirementId: string;
  label: string;
  status: "missing" | "defined_not_connected" | "blocked";
  summary: string;
};

export type WorkflowBlockedActionAuditStep = {
  stepId: string;
  label: string;
  status: "observed" | "blocked";
  auditRef: string;
  summary: string;
};

export type WorkflowConfirmationPlaceholderPreview = {
  confirmationPlaceholderId: string;
  requiredActionRef: string;
  riskSummary: string;
  requiredDecisionShape: string[];
  humanReviewRequired: true;
  disabledReason: string;
  auditRef: string;
};

export type WorkflowBlockedActionPreviewContext = {
  runId: string;
  workflowDefinitionId: string;
  requestId?: string;
  auditRef?: string;
};

export type WorkflowBlockedActionPreviewViewModel = {
  pageId: "workflow-blocked-action-preview-read";
  sourcePageId: "workflow-run-detail-read";
  sourceRouteId: "run-record-summary-list-route";
  draftRouteId: "tool-action-preview-read-draft";
  confirmationDraftRouteId: "confirmation-placeholder-read-draft";
  routePath: typeof CONTROL_PLANE_READ_ROUTES.runs;
  requestId: string;
  auditRef: string;
  runId: string;
  workflowDefinitionId: string;
  nodeExecutionRef: string;
  toolActionId: string;
  toolRef: string;
  actionKind: string;
  riskLevel: "low" | "medium" | "high";
  requiresConfirmation: true;
  policyReason: string;
  blockedState: WorkflowDefinitionBlockedActionPreview["blockedState"];
  missingPrerequisites: WorkflowBlockedActionRequirement[];
  auditTrail: WorkflowBlockedActionAuditStep[];
  confirmationPlaceholder: WorkflowConfirmationPlaceholderPreview;
  relatedRunGuard: WorkflowRunDetailGuardPreview;
  forbiddenProjectionBlocked: boolean;
  canRenderBlockedActionPreview: boolean;
  canRequestLiveBackend: false;
  canMutate: false;
  canExecuteTool: false;
  canSubmitConfirmationDecision: false;
  canUnlockExecution: false;
  canWriteBusinessTruth: false;
  canReplayRun: false;
};

const DEFAULT_BLOCKED_ACTION_PREVIEW: WorkflowDefinitionBlockedActionPreview = {
  toolActionId: "tool_action_preview_reconnect_stream",
  nodeId: "node_policy_gate",
  toolRef: "radishflow.candidate_action",
  actionKind: "suggest_flowsheet_edit",
  riskLevel: "medium",
  requiresConfirmation: true,
  policyReason: "Candidate action remains blocked until a future confirmation flow and executor are implemented.",
  blockedState: "blocked_executor_not_available",
  auditRef: "audit_tool_action_preview_demo",
};

const DEFAULT_RELATED_RUN_GUARD: WorkflowRunDetailGuardPreview = {
  guardId: "guard_replay_resume",
  label: "Replay and resume",
  status: "blocked",
  reason: "Run replay and resume remain disabled until executor, confirmation, and durable run store gates exist.",
  missingPrerequisite: "workflow executor implementation gate",
  auditRef: "audit_run_detail_blocked_replay_demo",
};

const DEFAULT_CONTEXT: Required<WorkflowBlockedActionPreviewContext> = {
  runId: "run_radishflow_copilot_20260531_001",
  workflowDefinitionId: "wf_radishflow_copilot_latest",
  requestId: "req_workflow_blocked_action_preview_demo",
  auditRef: "audit_workflow_blocked_action_preview_demo",
};

export function buildWorkflowBlockedActionPreviewViewModel(
  action: WorkflowDefinitionBlockedActionPreview = DEFAULT_BLOCKED_ACTION_PREVIEW,
  relatedRunGuard: WorkflowRunDetailGuardPreview = DEFAULT_RELATED_RUN_GUARD,
  context: WorkflowBlockedActionPreviewContext = DEFAULT_CONTEXT,
): WorkflowBlockedActionPreviewViewModel {
  const route = CONTROL_PLANE_READ_ROUTE_DEFINITIONS["run-record-summary-list-route"];
  const routePath = CONTROL_PLANE_READ_ROUTES.runs;
  const missingPrerequisites = buildMissingPrerequisites();
  const auditTrail = buildAuditTrail(action);
  const confirmationPlaceholder = buildConfirmationPlaceholder(action);
  const forbiddenProjectionBlocked = controlPlaneReadResponseHasForbiddenOutput({
    detail: { [CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS[7]]: "blocked" },
  });

  return {
    pageId: "workflow-blocked-action-preview-read",
    sourcePageId: "workflow-run-detail-read",
    sourceRouteId: "run-record-summary-list-route",
    draftRouteId: "tool-action-preview-read-draft",
    confirmationDraftRouteId: "confirmation-placeholder-read-draft",
    routePath,
    requestId: context.requestId ?? DEFAULT_CONTEXT.requestId,
    auditRef: context.auditRef ?? DEFAULT_CONTEXT.auditRef,
    runId: context.runId,
    workflowDefinitionId: context.workflowDefinitionId,
    nodeExecutionRef: action.nodeId,
    toolActionId: action.toolActionId,
    toolRef: action.toolRef,
    actionKind: action.actionKind,
    riskLevel: action.riskLevel,
    requiresConfirmation: action.requiresConfirmation,
    policyReason: action.policyReason,
    blockedState: action.blockedState,
    missingPrerequisites,
    auditTrail,
    confirmationPlaceholder,
    relatedRunGuard,
    forbiddenProjectionBlocked,
    canRenderBlockedActionPreview:
      route.path === routePath &&
      route.canMutate === false &&
      action.requiresConfirmation &&
      missingPrerequisites.length >= 3 &&
      confirmationPlaceholder.humanReviewRequired,
    canRequestLiveBackend: false,
    canMutate: false,
    canExecuteTool: false,
    canSubmitConfirmationDecision: false,
    canUnlockExecution: false,
    canWriteBusinessTruth: false,
    canReplayRun: false,
  };
}

function buildMissingPrerequisites(): WorkflowBlockedActionRequirement[] {
  return [
    {
      requirementId: "req_confirmation_flow_contract",
      label: "Confirmation flow contract",
      status: "defined_not_connected",
      summary: "The future confirmation shape is visible, but no decision path is connected to this surface.",
    },
    {
      requirementId: "req_runtime_gate",
      label: "Workflow executor gate",
      status: "missing",
      summary: "The workflow executor is not implemented, so node execution cannot be resumed from this preview.",
    },
    {
      requirementId: "req_tool_adapter_gate",
      label: "Tool executor adapter",
      status: "missing",
      summary: "No tool adapter is connected for the candidate action; only sanitized metadata is rendered.",
    },
    {
      requirementId: "req_business_writeback_policy",
      label: "Business writeback policy",
      status: "blocked",
      summary: "Writing into the upstream business truth source remains outside the current read-side boundary.",
    },
  ];
}

function buildAuditTrail(action: WorkflowDefinitionBlockedActionPreview): WorkflowBlockedActionAuditStep[] {
  return [
    {
      stepId: "audit_step_candidate_action_projected",
      label: "Candidate action projected",
      status: "observed",
      auditRef: action.auditRef,
      summary: "The workflow definition detail exposed a sanitized candidate action preview.",
    },
    {
      stepId: "audit_step_policy_evaluated",
      label: "Policy evaluated",
      status: "observed",
      auditRef: "audit_tool_action_policy_gate_demo",
      summary: "The policy gate marked the action as requiring human review before any future runtime path.",
    },
    {
      stepId: "audit_step_confirmation_placeholder_displayed",
      label: "Confirmation placeholder displayed",
      status: "blocked",
      auditRef: "audit_confirmation_placeholder_demo",
      summary: "The placeholder shows the required decision shape without accepting a decision.",
    },
    {
      stepId: "audit_step_executor_blocked",
      label: "Executor blocked",
      status: "blocked",
      auditRef: "audit_executor_blocked_demo",
      summary: "Executor, tool adapter, writeback, and replay capabilities remain unavailable.",
    },
  ];
}

function buildConfirmationPlaceholder(
  action: WorkflowDefinitionBlockedActionPreview,
): WorkflowConfirmationPlaceholderPreview {
  return {
    confirmationPlaceholderId: "confirmation_placeholder_tool_action_preview",
    requiredActionRef: action.toolActionId,
    riskSummary: "Medium-risk candidate action requires future human approval before any runtime can continue.",
    requiredDecisionShape: ["decision: approve | reject", "actor_subject_ref", "reason", "audit_ref"],
    humanReviewRequired: true,
    disabledReason: "Confirmation submission is disabled until confirmation flow, executor, and writeback gates exist.",
    auditRef: "audit_confirmation_placeholder_demo",
  };
}
