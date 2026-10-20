import type { WorkflowProjectionCopy, WorkflowProjectionMessage } from "./workflowProjectionCopy.ts";
import {
  CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS,
  CONTROL_PLANE_READ_ROUTE_DEFINITIONS,
  CONTROL_PLANE_READ_ROUTES,
  controlPlaneReadResponseHasForbiddenOutput,
} from "../../../../../contracts/typescript/control-plane-read-api";
import {
  buildWorkflowDraftDesignerViewModel,
  type WorkflowDraftDesignerBlockedCapability,
  type WorkflowDraftDesignerDraft,
  type WorkflowDraftDesignerNode,
} from "./workflowDraftDesigner";
import {
  buildWorkflowDraftValidationInspectorViewModel,
  type WorkflowDraftValidationInspectorViewModel,
  type WorkflowDraftValidationStatus,
} from "./workflowDraftValidationInspector";

export type WorkflowExecutionPlanStageKind = "context" | "model" | "policy" | "preview" | "output" | "audit";
export type WorkflowExecutionPlanStatus = "ready" | "review_required" | "blocked";

export type WorkflowExecutionPlanSummary = WorkflowProjectionCopy & {
  label: string;
  value: string;
  status: WorkflowExecutionPlanStatus;
  summary: string;
};

export type WorkflowExecutionPlanStage = WorkflowProjectionCopy & {
  stageId: string;
  order: number;
  label: string;
  stageKind: WorkflowExecutionPlanStageKind;
  status: WorkflowExecutionPlanStatus;
  nodeIds: string[];
  summary: string;
  blockedReason: string;
};

export type WorkflowExecutionPlanNodeMapping = WorkflowProjectionCopy & {
  nodeId: string;
  label: string;
  stageId: string;
  stageOrder: number;
  nodeType: WorkflowDraftDesignerNode["nodeType"];
  providerProfileRef: string;
  executionMode: "offline_preview_only";
  requiresConfirmation: boolean;
  inputSummary: string;
  outputSummary: string;
};

export type WorkflowExecutionPlanProviderRequirement = WorkflowProjectionCopy & {
  requirementId: string;
  label: string;
  providerProfileRef: string;
  nodeIds: string[];
  status: "defined_not_connected" | "blocked";
  missingPrerequisite: string;
  summary: string;
};

export type WorkflowExecutionPlanGate = WorkflowProjectionCopy & {
  gateId: string;
  label: string;
  gateKind: "policy" | "confirmation" | "audit" | "runtime";
  status: WorkflowExecutionPlanStatus;
  requiredBeforeStageId: string;
  summary: string;
  auditRef: string;
};

export type WorkflowExecutionPlanBlockedReason = WorkflowProjectionCopy & {
  reasonId: string;
  label: string;
  blockedCapability: "runtime" | "publish" | "writeback" | "replay";
  status: "blocked";
  missingPrerequisite: string;
  summary: string;
  auditRef: string;
};

export type WorkflowExecutionPlanAuditMetadata = {
  sourceRouteId: "workflow-definition-summary-list-route";
  draftRouteId: "workflow-execution-plan-preview-offline-draft";
  validationRouteId: "workflow-draft-validation-inspector-offline-draft";
  requestId: string;
  auditRef: string;
  selectedDraftId: string;
};

export type WorkflowExecutionPlanPreviewViewModel = WorkflowProjectionCopy & {
  pageId: "workflow-execution-plan-preview-offline";
  sourcePageId: "workflow-draft-validation-inspector-offline";
  sourceRouteId: "workflow-definition-summary-list-route";
  draftRouteId: "workflow-execution-plan-preview-offline-draft";
  routePath: typeof CONTROL_PLANE_READ_ROUTES.workflowDefinitions;
  requestId: string;
  auditRef: string;
  selectedDraftId: string;
  validationStatus: WorkflowDraftValidationStatus;
  summary: WorkflowExecutionPlanSummary[];
  stageOrder: WorkflowExecutionPlanStage[];
  nodeStageMappings: WorkflowExecutionPlanNodeMapping[];
  providerProfileRequirements: WorkflowExecutionPlanProviderRequirement[];
  confirmationAuditGates: WorkflowExecutionPlanGate[];
  blockedPlanReasons: WorkflowExecutionPlanBlockedReason[];
  auditMetadata: WorkflowExecutionPlanAuditMetadata;
  forbiddenProjectionBlocked: boolean;
  canRenderExecutionPlanPreview: boolean;
  canPreviewPlanLocally: true;
  canRequestLiveBackend: false;
  canPersistExecutionPlan: false;
  canPublishWorkflow: false;
  canStartRun: false;
  canExecuteWorkflow: false;
  canSubmitConfirmationDecision: false;
  canWriteBusinessTruth: false;
  canReplayRun: false;
};

const DEFAULT_DRAFT = buildWorkflowDraftDesignerViewModel().drafts[0]!;

export function buildWorkflowExecutionPlanPreviewViewModel(
  draft: WorkflowDraftDesignerDraft = DEFAULT_DRAFT,
  validationInspector: WorkflowDraftValidationInspectorViewModel = buildWorkflowDraftValidationInspectorViewModel(draft),
): WorkflowExecutionPlanPreviewViewModel {
  const route = CONTROL_PLANE_READ_ROUTE_DEFINITIONS["workflow-definition-summary-list-route"];
  const routePath = CONTROL_PLANE_READ_ROUTES.workflowDefinitions;
  const stageOrder = buildStageOrder(draft);
  const nodeStageMappings = buildNodeStageMappings(draft, stageOrder);
  const providerProfileRequirements = buildProviderProfileRequirements(draft, nodeStageMappings);
  const confirmationAuditGates = buildConfirmationAuditGates(draft, validationInspector);
  const blockedPlanReasons = buildBlockedPlanReasons(draft.blockedCapabilities);
  const forbiddenProjectionBlocked = controlPlaneReadResponseHasForbiddenOutput({
    executionPlanPreview: { [CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS[7]]: "blocked" },
  });

  return {
    pageId: "workflow-execution-plan-preview-offline",
    sourcePageId: "workflow-draft-validation-inspector-offline",
    sourceRouteId: "workflow-definition-summary-list-route",
    draftRouteId: "workflow-execution-plan-preview-offline-draft",
    routePath,
    requestId: draft.routeMetadata.requestId,
    auditRef: draft.routeMetadata.auditRef,
    selectedDraftId: draft.draftId,
    validationStatus: validationInspector.validationStatus,
    summary: buildSummary(
      draft,
      validationInspector.validationStatus,
      stageOrder,
      providerProfileRequirements,
      confirmationAuditGates,
      blockedPlanReasons,
    ),
    stageOrder,
    nodeStageMappings,
    providerProfileRequirements,
    confirmationAuditGates,
    blockedPlanReasons,
    auditMetadata: {
      sourceRouteId: "workflow-definition-summary-list-route",
      draftRouteId: "workflow-execution-plan-preview-offline-draft",
      validationRouteId: "workflow-draft-validation-inspector-offline-draft",
      requestId: draft.routeMetadata.requestId,
      auditRef: draft.routeMetadata.auditRef,
      selectedDraftId: draft.draftId,
    },
    forbiddenProjectionBlocked,
    canRenderExecutionPlanPreview:
      route.path === routePath &&
      route.canMutate === false &&
      stageOrder.length >= 6 &&
      nodeStageMappings.length === draft.nodes.length &&
      providerProfileRequirements.length >= 3 &&
      confirmationAuditGates.length >= 4 &&
      blockedPlanReasons.length >= 4,
    canPreviewPlanLocally: true,
    canRequestLiveBackend: false,
    canPersistExecutionPlan: false,
    canPublishWorkflow: false,
    canStartRun: false,
    canExecuteWorkflow: false,
    canSubmitConfirmationDecision: false,
    canWriteBusinessTruth: false,
    canReplayRun: false,
  };
}

function buildSummary(
  draft: WorkflowDraftDesignerDraft,
  validationStatus: WorkflowDraftValidationStatus,
  stageOrder: WorkflowExecutionPlanStage[],
  providerProfileRequirements: WorkflowExecutionPlanProviderRequirement[],
  confirmationAuditGates: WorkflowExecutionPlanGate[],
  blockedPlanReasons: WorkflowExecutionPlanBlockedReason[],
): WorkflowExecutionPlanSummary[] {
  return [
    {
      label: "Selected draft",
      labelMessage: { key: "planDraft" },
      value: draft.draftId,
      status: validationStatus === "blocked" ? "blocked" : "review_required",
      summary: "Execution plan preview is derived from the selected offline draft and validation inspector.",
      summaryMessage: { key: "planDraftSummary" },
    },
    {
      label: "Stage order",
      labelMessage: { key: "stageOrder" },
      value: String(stageOrder.length),
      status: stageOrder.some((stage) => stage.status === "blocked") ? "blocked" : "review_required",
      summary: "Stages show future execution order without creating an executable runtime plan.",
      summaryMessage: { key: "planStagesSummary" },
    },
    {
      label: "Provider requirements",
      labelMessage: { key: "providerRequirements" },
      value: String(providerProfileRequirements.length),
      status: "review_required",
      summary: "Provider profile and tool adapter requirements stay visible before any future runtime work.",
      summaryMessage: { key: "planProviderSummary" },
    },
    {
      label: "Blocked reasons",
      labelMessage: { key: "blockedReasons" },
      value: String(blockedPlanReasons.length),
      status: "blocked",
      summary: "Runtime, publish, writeback, and replay remain blocked by explicit prerequisites.",
      summaryMessage: { key: "planBlockedSummary" },
    },
    {
      label: "Gates",
      labelMessage: { key: "gates" },
      value: String(confirmationAuditGates.length),
      status: "review_required",
      summary: "Policy, confirmation, audit, and runtime gates are rendered as metadata only.",
      summaryMessage: { key: "planGatesSummary" },
    },
  ];
}

function buildStageOrder(draft: WorkflowDraftDesignerDraft): WorkflowExecutionPlanStage[] {
  return [
    buildStage(draft, "stage_context_collection", 1, "Context collection", "context"),
    buildStage(draft, "stage_model_reasoning", 2, "Model reasoning", "model"),
    buildStage(draft, "stage_policy_gate", 3, "Policy gate", "policy"),
    buildStage(draft, "stage_tool_preview", 4, "Tool preview", "preview"),
    buildStage(draft, "stage_advisory_output", 5, "Advisory output", "output"),
    buildStage(draft, "stage_audit_projection", 6, "Audit projection", "audit"),
  ];
}

function buildStage(
  draft: WorkflowDraftDesignerDraft,
  stageId: string,
  order: number,
  label: string,
  stageKind: WorkflowExecutionPlanStageKind,
): WorkflowExecutionPlanStage {
  const nodes = nodesForStage(draft.nodes, stageKind);
  const status = stageStatus(stageKind, nodes);
  return {
    stageId,
    order,
    label,
    labelMessage: { key: `stageLabel_${stageKind}` },
    stageKind,
    status,
    nodeIds: nodes.map((node) => node.nodeId),
    summary: stageSummary(stageKind, nodes),
    summaryMessage: stageSummaryMessage(stageKind, nodes),
    blockedReason: stageBlockedReason(stageKind),
    blockedReasonMessage: stageBlockedReasonMessage(stageKind),
  };
}

function stageSummaryMessage(stageKind: WorkflowExecutionPlanStageKind, nodes: WorkflowDraftDesignerNode[]): WorkflowProjectionMessage | undefined {
  if (stageKind === "preview") return { key: "stagePreview" };
  if (stageKind === "audit") return { key: "stageAudit" };
  return nodes.map(node => node.label).join(", ").length > 0 ? undefined : { key: "stageEmpty" };
}

function stageBlockedReasonMessage(stageKind: WorkflowExecutionPlanStageKind): WorkflowProjectionMessage {
  switch (stageKind) {
    case "preview": return { key: "stageBlockedPreview" };
    case "policy": return { key: "stageBlockedPolicy" };
    case "audit": return { key: "stageBlockedAudit" };
    default: return { key: "stageLocalAllowed" };
  }
}

function nodesForStage(
  nodes: WorkflowDraftDesignerNode[],
  stageKind: WorkflowExecutionPlanStageKind,
): WorkflowDraftDesignerNode[] {
  if (stageKind === "audit") {
    return nodes.filter((node) => node.nodeId.includes("audit"));
  }
  if (stageKind === "output") {
    return nodes.filter((node) => node.lane === "output" && !node.nodeId.includes("audit"));
  }
  return nodes.filter((node) => node.lane === stageKind);
}

function stageStatus(
  stageKind: WorkflowExecutionPlanStageKind,
  nodes: WorkflowDraftDesignerNode[],
): WorkflowExecutionPlanStatus {
  if (stageKind === "preview") {
    return "blocked";
  }
  if (stageKind === "policy" || nodes.some((node) => node.requiresConfirmation)) {
    return "review_required";
  }
  return "ready";
}

function stageSummary(
  stageKind: WorkflowExecutionPlanStageKind,
  nodes: WorkflowDraftDesignerNode[],
): string {
  if (stageKind === "preview") {
    return "Preview-only tool action metadata is visible, but executor and tool adapter calls stay blocked.";
  }
  if (stageKind === "audit") {
    return "Audit projection is included as sanitized metadata for future run trace review.";
  }
  const labels = nodes.map((node) => node.label).join(", ");
  return labels.length > 0 ? labels : "No node is assigned to this stage in the offline draft.";
}

function stageBlockedReason(stageKind: WorkflowExecutionPlanStageKind): string {
  if (stageKind === "preview") {
    return "workflow executor and tool adapter gate missing";
  }
  if (stageKind === "policy") {
    return "confirmation decision and execution unlock are not connected";
  }
  if (stageKind === "audit") {
    return "durable audit and result stores are not implemented";
  }
  return "not blocked for local read-only preview";
}

function buildNodeStageMappings(
  draft: WorkflowDraftDesignerDraft,
  stageOrder: WorkflowExecutionPlanStage[],
): WorkflowExecutionPlanNodeMapping[] {
  return draft.nodes.map((node) => {
    const stage = stageOrder.find((candidate) => candidate.nodeIds.includes(node.nodeId)) ?? stageOrder[0]!;
    return {
      nodeId: node.nodeId,
      label: node.label,
      labelMessage: node.labelMessage,
      stageId: stage.stageId,
      stageOrder: stage.order,
      nodeType: node.nodeType,
      providerProfileRef: providerProfileForNode(draft.providerProfileRef, node),
      executionMode: "offline_preview_only",
      requiresConfirmation: node.requiresConfirmation,
      inputSummary: node.inputSummary,
      outputSummary: node.outputSummary,
    };
  });
}

function providerProfileForNode(providerProfileRef: string, node: WorkflowDraftDesignerNode): string {
  if (node.providerRef.trim()) {
    return node.providerRef.trim();
  }
  if (node.nodeType === "http_tool") {
    return node.toolRef.trim() || "tool-adapter:radishflow.candidate-action";
  }
  if (node.nodeType === "condition") {
    return "policy:confirmation-gated";
  }
  return providerProfileRef;
}

function buildProviderProfileRequirements(
  draft: WorkflowDraftDesignerDraft,
  nodeStageMappings: WorkflowExecutionPlanNodeMapping[],
): WorkflowExecutionPlanProviderRequirement[] {
  const modelNodeIds = nodeStageMappings
    .filter((mapping) => mapping.nodeType === "llm")
    .map((mapping) => mapping.nodeId);
  const toolNodeIds = nodeStageMappings
    .filter((mapping) => mapping.nodeType === "http_tool")
    .map((mapping) => mapping.nodeId);
  const policyNodeIds = nodeStageMappings
    .filter((mapping) => mapping.nodeType === "condition")
    .map((mapping) => mapping.nodeId);
  return [
    {
      requirementId: "provider_profile_model_stage",
      label: "Model provider profile",
      labelMessage: { key: "provider_profile_model_stage_label" },
      providerProfileRef: providerRequirementRef(nodeStageMappings, modelNodeIds, draft.providerProfileRef),
      nodeIds: modelNodeIds,
      status: "defined_not_connected",
      missingPrerequisite: "workflow runtime provider binding task card",
      missingPrerequisiteMessage: { key: "provider_profile_model_stage_missingPrerequisite" },
      summary: "The model profile is shown as a future requirement; no model runtime call is made by this preview.",
      summaryMessage: { key: "provider_profile_model_stage_summary" },
    },
    {
      requirementId: "tool_adapter_preview_stage",
      label: "Tool adapter",
      labelMessage: { key: "tool_adapter_preview_stage_label" },
      providerProfileRef: providerRequirementRef(
        nodeStageMappings,
        toolNodeIds,
        "tool-adapter:radishflow.candidate-action",
      ),
      nodeIds: toolNodeIds,
      status: "blocked",
      missingPrerequisite: "tool executor adapter implementation gate",
      missingPrerequisiteMessage: { key: "tool_adapter_preview_stage_missingPrerequisite" },
      summary: "Tool action preview nodes cannot execute until the executor and confirmation gates exist.",
      summaryMessage: { key: "tool_adapter_preview_stage_summary" },
    },
    {
      requirementId: "policy_confirmation_profile",
      label: "Policy and confirmation profile",
      labelMessage: { key: "policy_confirmation_profile_label" },
      providerProfileRef: "policy:confirmation-gated",
      nodeIds: policyNodeIds,
      status: "defined_not_connected",
      missingPrerequisite: "confirmation decision store and policy audit gate",
      missingPrerequisiteMessage: { key: "policy_confirmation_profile_missingPrerequisite" },
      summary: "Policy and confirmation nodes remain visible but disconnected from execution unlock.",
      summaryMessage: { key: "policy_confirmation_profile_summary" },
    },
  ];
}

function providerRequirementRef(
  nodeStageMappings: WorkflowExecutionPlanNodeMapping[],
  nodeIds: string[],
  fallbackRef: string,
): string {
  const matchedRefs = nodeStageMappings
    .filter((mapping) => nodeIds.includes(mapping.nodeId))
    .map((mapping) => mapping.providerProfileRef)
    .filter((ref) => ref.trim());
  return matchedRefs[0] ?? fallbackRef;
}

function buildConfirmationAuditGates(
  draft: WorkflowDraftDesignerDraft,
  validationInspector: WorkflowDraftValidationInspectorViewModel,
): WorkflowExecutionPlanGate[] {
  return [
    {
      gateId: "gate_validation_before_plan",
      label: "Validation inspector gate",
      labelMessage: { key: "gate_validation_before_plan_label" },
      gateKind: "policy",
      status: validationInspector.validationStatus === "blocked" ? "blocked" : "review_required",
      requiredBeforeStageId: "stage_policy_gate",
      summary: "The offline validation inspector must remain visible before a future executable plan can exist.",
      summaryMessage: { key: "gate_validation_before_plan_summary" },
      auditRef: validationInspector.auditRef,
    },
    {
      gateId: "gate_confirmation_before_tool",
      label: "Confirmation before tool preview",
      labelMessage: { key: "gate_confirmation_before_tool_label" },
      gateKind: "confirmation",
      status: "blocked",
      requiredBeforeStageId: "stage_tool_preview",
      summary: "Human confirmation shape is displayed, but no decision can be submitted or persisted.",
      summaryMessage: { key: "gate_confirmation_before_tool_summary" },
      auditRef: findAuditRef(draft.blockedCapabilities, "blocked_confirmation_decision"),
    },
    {
      gateId: "gate_audit_projection",
      label: "Audit projection",
      labelMessage: { key: "gate_audit_projection_label" },
      gateKind: "audit",
      status: "review_required",
      requiredBeforeStageId: "stage_audit_projection",
      summary: "Plan preview includes request and audit refs without exposing raw prompt, token, or tool payloads.",
      summaryMessage: { key: "gate_audit_projection_summary" },
      auditRef: draft.routeMetadata.auditRef,
    },
    {
      gateId: "gate_runtime_execution",
      label: "Runtime execution gate",
      labelMessage: { key: "gate_runtime_execution_label" },
      gateKind: "runtime",
      status: "blocked",
      requiredBeforeStageId: "stage_model_reasoning",
      summary: "Executor, durable run store, result materialization, and replay gates are not implemented.",
      summaryMessage: { key: "gate_runtime_execution_summary" },
      auditRef: findAuditRef(draft.blockedCapabilities, "blocked_runtime"),
    },
  ];
}

function buildBlockedPlanReasons(
  blockedCapabilities: WorkflowDraftDesignerBlockedCapability[],
): WorkflowExecutionPlanBlockedReason[] {
  return [
    {
      reasonId: "blocked_runtime",
      label: "Runtime execution",
      labelMessage: { key: "blocked_runtime_label" },
      blockedCapability: "runtime",
      status: "blocked",
      missingPrerequisite: findMissingPrerequisite(blockedCapabilities, "blocked_runtime"),
      missingPrerequisiteMessage: blockedCapabilities.find(capability => capability.capabilityId === "blocked_runtime")?.missingPrerequisiteMessage ?? (blockedCapabilities.some(capability => capability.capabilityId === "blocked_runtime") ? undefined : { key: "findBlockedReason_missingPrerequisite" }),
      summary: "No workflow executor, node executor, tool executor, or agent loop is available.",
      summaryMessage: { key: "blocked_runtime_summary" },
      auditRef: findAuditRef(blockedCapabilities, "blocked_runtime"),
    },
    {
      reasonId: "blocked_publish",
      label: "Workflow publish",
      labelMessage: { key: "blocked_publish_label" },
      blockedCapability: "publish",
      status: "blocked",
      missingPrerequisite: findMissingPrerequisite(blockedCapabilities, "blocked_publish"),
      missingPrerequisiteMessage: blockedCapabilities.find(capability => capability.capabilityId === "blocked_publish")?.missingPrerequisiteMessage ?? (blockedCapabilities.some(capability => capability.capabilityId === "blocked_publish") ? undefined : { key: "findBlockedReason_missingPrerequisite" }),
      summary: "The preview is derived from an offline draft and cannot publish workflow versions.",
      summaryMessage: { key: "blocked_publish_summary" },
      auditRef: findAuditRef(blockedCapabilities, "blocked_publish"),
    },
    {
      reasonId: "blocked_writeback",
      label: "Business writeback",
      labelMessage: { key: "blocked_writeback_label" },
      blockedCapability: "writeback",
      status: "blocked",
      missingPrerequisite: "business writeback policy and confirmation execution gate",
      missingPrerequisiteMessage: { key: "blocked_writeback_missingPrerequisite" },
      summary: "Candidate actions remain advisory and cannot write to upstream project truth sources.",
      summaryMessage: { key: "blocked_writeback_summary" },
      auditRef: "audit_execution_plan_writeback_blocked",
    },
    {
      reasonId: "blocked_replay",
      label: "Run replay",
      labelMessage: { key: "blocked_replay_label" },
      blockedCapability: "replay",
      status: "blocked",
      missingPrerequisite: "durable run store, executor, and replay policy gate",
      missingPrerequisiteMessage: { key: "blocked_replay_missingPrerequisite" },
      summary: "Run replay and resume are not available from an offline execution plan preview.",
      summaryMessage: { key: "blocked_replay_summary" },
      auditRef: "audit_execution_plan_replay_blocked",
    },
  ];
}

function findMissingPrerequisite(
  blockedCapabilities: WorkflowDraftDesignerBlockedCapability[],
  capabilityId: string,
): string {
  return (
    blockedCapabilities.find((capability) => capability.capabilityId === capabilityId)?.missingPrerequisite ??
    "future implementation gate"
  );
}

function findAuditRef(blockedCapabilities: WorkflowDraftDesignerBlockedCapability[], capabilityId: string): string {
  return (
    blockedCapabilities.find((capability) => capability.capabilityId === capabilityId)?.auditRef ??
    `audit_execution_plan_${capabilityId}`
  );
}
