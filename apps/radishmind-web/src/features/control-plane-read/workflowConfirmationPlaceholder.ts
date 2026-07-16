import {
  CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS,
  CONTROL_PLANE_READ_ROUTE_DEFINITIONS,
  CONTROL_PLANE_READ_ROUTES,
  controlPlaneReadResponseHasForbiddenOutput,
} from "../../../../../contracts/typescript/control-plane-read-api";
import {
  buildWorkflowBlockedActionPreviewViewModel,
  type WorkflowBlockedActionPreviewViewModel,
} from "./workflowBlockedActionPreview";

export type WorkflowConfirmationDecisionField = {
  fieldId: string;
  label: string;
  required: boolean;
  source: "archived_legacy_confirmation_contract" | "audit_context" | "policy_context";
};

export type WorkflowConfirmationPlaceholderPrerequisite = {
  prerequisiteId: string;
  label: string;
  status: "missing" | "defined_not_connected" | "blocked";
  summary: string;
  auditRef: string;
};

export type WorkflowConfirmationPlaceholderAuditMetadata = {
  sourceRouteId: "run-record-summary-list-route";
  draftRouteId: "confirmation-placeholder-read-draft";
  requestId: string;
  auditRef: string;
  policyRef: string;
  traceRef: string;
};

export type WorkflowConfirmationPlaceholderViewModel = {
  pageId: "workflow-confirmation-placeholder-read";
  sourcePageId: "workflow-blocked-action-preview-read";
  sourceRouteId: "run-record-summary-list-route";
  draftRouteId: "confirmation-placeholder-read-draft";
  routePath: typeof CONTROL_PLANE_READ_ROUTES.runs;
  requestId: string;
  auditRef: string;
  confirmationPlaceholderId: string;
  requiredActionRef: string;
  requiredRunRef: string;
  workflowDefinitionId: string;
  nodeExecutionRef: string;
  toolRef: string;
  actionKind: string;
  riskLevel: "low" | "medium" | "high";
  riskSummary: string;
  policyReason: string;
  requiredDecisionShape: string[];
  decisionFields: WorkflowConfirmationDecisionField[];
  humanReviewRequired: true;
  disabledReason: string;
  prerequisites: WorkflowConfirmationPlaceholderPrerequisite[];
  auditMetadata: WorkflowConfirmationPlaceholderAuditMetadata;
  legacyContractStatus: "superseded_archived";
  historicalReadAllowed: true;
  newSubmissionAllowed: false;
  runtimeAuthoritative: false;
  supersededBy: Array<
    "workflow_http_tool_action_plan.v1" | "workflow_http_tool_confirmation_decision.v1"
  >;
  forbiddenProjectionBlocked: boolean;
  canRenderConfirmationPlaceholder: boolean;
  canRequestLiveBackend: false;
  canMutate: false;
  canSubmitConfirmationDecision: false;
  canApproveDecision: false;
  canRejectDecision: false;
  canDeferDecision: false;
  canPersistDecision: false;
  canUnlockExecution: false;
  canExecuteTool: false;
  canWriteBusinessTruth: false;
  canReplayRun: false;
};

export function buildWorkflowConfirmationPlaceholderViewModel(
  blockedActionPreview: WorkflowBlockedActionPreviewViewModel = buildWorkflowBlockedActionPreviewViewModel(),
): WorkflowConfirmationPlaceholderViewModel {
  const route = CONTROL_PLANE_READ_ROUTE_DEFINITIONS["run-record-summary-list-route"];
  const routePath = CONTROL_PLANE_READ_ROUTES.runs;
  const placeholder = blockedActionPreview.confirmationPlaceholder;
  const decisionFields = buildDecisionFields();
  const prerequisites = buildPrerequisites(blockedActionPreview);
  const forbiddenProjectionBlocked = controlPlaneReadResponseHasForbiddenOutput({
    detail: { [CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS[7]]: "blocked" },
  });

  return {
    pageId: "workflow-confirmation-placeholder-read",
    sourcePageId: "workflow-blocked-action-preview-read",
    sourceRouteId: "run-record-summary-list-route",
    draftRouteId: "confirmation-placeholder-read-draft",
    routePath,
    requestId: blockedActionPreview.requestId,
    auditRef: placeholder.auditRef,
    confirmationPlaceholderId: placeholder.confirmationPlaceholderId,
    requiredActionRef: placeholder.requiredActionRef,
    requiredRunRef: blockedActionPreview.runId,
    workflowDefinitionId: blockedActionPreview.workflowDefinitionId,
    nodeExecutionRef: blockedActionPreview.nodeExecutionRef,
    toolRef: blockedActionPreview.toolRef,
    actionKind: blockedActionPreview.actionKind,
    riskLevel: blockedActionPreview.riskLevel,
    riskSummary: placeholder.riskSummary,
    policyReason: blockedActionPreview.policyReason,
    requiredDecisionShape: placeholder.requiredDecisionShape,
    decisionFields,
    humanReviewRequired: true,
    disabledReason:
      "This run-bound confirmation placeholder is archived and cannot submit a decision. Use the independent pre-run HTTP Tool action plan review surface for new confirmations.",
    prerequisites,
    auditMetadata: {
      sourceRouteId: "run-record-summary-list-route",
      draftRouteId: "confirmation-placeholder-read-draft",
      requestId: blockedActionPreview.requestId,
      auditRef: placeholder.auditRef,
      policyRef: "workflow_http_tool_pre_run_confirmation_v1",
      traceRef: "trace_confirmation_placeholder_demo",
    },
    legacyContractStatus: "superseded_archived",
    historicalReadAllowed: true,
    newSubmissionAllowed: false,
    runtimeAuthoritative: false,
    supersededBy: [
      "workflow_http_tool_action_plan.v1",
      "workflow_http_tool_confirmation_decision.v1",
    ],
    forbiddenProjectionBlocked,
    canRenderConfirmationPlaceholder:
      route.path === routePath &&
      route.canMutate === false &&
      placeholder.humanReviewRequired &&
      placeholder.requiredDecisionShape.length > 0 &&
      decisionFields.length >= 4 &&
      prerequisites.length >= 4,
    canRequestLiveBackend: false,
    canMutate: false,
    canSubmitConfirmationDecision: false,
    canApproveDecision: false,
    canRejectDecision: false,
    canDeferDecision: false,
    canPersistDecision: false,
    canUnlockExecution: false,
    canExecuteTool: false,
    canWriteBusinessTruth: false,
    canReplayRun: false,
  };
}

function buildDecisionFields(): WorkflowConfirmationDecisionField[] {
  return [
    {
      fieldId: "decision",
      label: "Decision",
      required: true,
      source: "archived_legacy_confirmation_contract",
    },
    {
      fieldId: "actor_subject_ref",
      label: "Actor subject ref",
      required: true,
      source: "audit_context",
    },
    {
      fieldId: "reason",
      label: "Reason",
      required: true,
      source: "policy_context",
    },
    {
      fieldId: "audit_ref",
      label: "Audit ref",
      required: true,
      source: "audit_context",
    },
  ];
}

function buildPrerequisites(
  blockedActionPreview: WorkflowBlockedActionPreviewViewModel,
): WorkflowConfirmationPlaceholderPrerequisite[] {
  return [
    {
      prerequisiteId: "prereq_confirmation_flow",
      label: "Archived run-bound confirmation",
      status: "blocked",
      summary:
        "The legacy placeholder remains readable for historical context, but all new submissions are permanently disabled.",
      auditRef: blockedActionPreview.confirmationPlaceholder.auditRef,
    },
    {
      prerequisiteId: "prereq_executor_gate",
      label: "Executor gate",
      status: "blocked",
      summary: "Batch A does not register HTTP execution, so this archived placeholder cannot unlock a tool attempt.",
      auditRef: "audit_confirmation_executor_gate_blocked_demo",
    },
    {
      prerequisiteId: "prereq_decision_store",
      label: "Independent pre-run decision store",
      status: "blocked",
      summary:
        "The versioned pre-run decision resource is separate and must never accept this legacy run-bound placeholder shape.",
      auditRef: "audit_confirmation_decision_store_missing_demo",
    },
    {
      prerequisiteId: "prereq_writeback_policy",
      label: "Writeback policy",
      status: "blocked",
      summary: "Business writeback remains outside the current read-only workflow surface.",
      auditRef: "audit_confirmation_writeback_policy_blocked_demo",
    },
  ];
}
