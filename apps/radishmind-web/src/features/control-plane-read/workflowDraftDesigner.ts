import {
  CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS,
  CONTROL_PLANE_READ_ROUTE_DEFINITIONS,
  CONTROL_PLANE_READ_ROUTES,
  controlPlaneReadResponseHasForbiddenOutput,
  type WorkflowDefinitionSummary,
} from "../../../../../contracts/typescript/control-plane-read-api";
import {
  buildWorkflowDefinitionDetailViewModel,
  type WorkflowDefinitionBlockedActionPreview,
  type WorkflowDefinitionDetailEdge,
  type WorkflowDefinitionDetailNode,
} from "./workflowDefinitionDetail";
import { buildWorkflowConfirmationPlaceholderViewModel, type WorkflowConfirmationPlaceholderViewModel } from "./workflowConfirmationPlaceholder";
import type { WorkspaceWorkflowDefinitionRow } from "./workspaceWorkflowDefinitions";

export type WorkflowDraftDesignerTemplate = {
  draftId: string;
  label: string;
  applicationRef: string;
  workflowDefinitionId: string;
  workflowKind: "workflow_copilot" | "docs_qa" | "workflow_template";
  providerProfileRef: string;
  summary: string;
  riskLevel: "low" | "medium" | "high";
  nodeCount: number;
  status: "ready_for_review" | "needs_policy_review" | "blocked_missing_runtime";
};

export type WorkflowDraftDesignerNode = {
  nodeId: string;
  label: string;
  nodeType: WorkflowDefinitionDetailNode["nodeType"];
  lane: "context" | "model" | "policy" | "preview" | "output";
  readiness: "ready" | "review_required" | "blocked";
  inputSummary: string;
  outputSummary: string;
  providerRef: string;
  toolRef: string;
  ragRef: string;
  inputContractFields: string[];
  outputContractFields: string[];
  outputMappingSummary: string;
  riskLevel: "low" | "medium" | "high";
  requiresConfirmation: boolean;
  previewOnlyReason: string;
};

export type WorkflowDraftDesignerEdge = {
  edgeId: string;
  fromNodeId: string;
  toNodeId: string;
  edgeKind: "context" | "policy" | "preview" | "audit";
  conditionSummary: string;
};

export type WorkflowDraftDesignerLayoutPosition = {
  nodeId: string;
  x: number;
  y: number;
};

export type WorkflowDraftDesignerLayout = {
  source: "workflow_node_designer";
  persistence: "ui_only" | "saved_draft_metadata";
  nodePositions: WorkflowDraftDesignerLayoutPosition[];
};

export type WorkflowDraftDesignerReadiness = {
  checkId: string;
  label: string;
  status: "ready" | "review_required" | "blocked";
  summary: string;
};

export type WorkflowDraftDesignerRisk = {
  riskId: string;
  label: string;
  riskLevel: "low" | "medium" | "high";
  requiresConfirmation: boolean;
  summary: string;
};

export type WorkflowDraftDesignerBlockedCapability = {
  capabilityId: string;
  label: string;
  status: "blocked";
  missingPrerequisite: string;
  summary: string;
  auditRef: string;
};

export type WorkflowDraftDesignerRouteMetadata = {
  sourceRouteId: "workflow-definition-summary-list-route";
  draftRouteId: "workflow-draft-designer-offline-draft";
  routePath: typeof CONTROL_PLANE_READ_ROUTES.workflowDefinitions;
  requestId: string;
  auditRef: string;
};

export type WorkflowDraftDesignerDraft = {
  draftId: string;
  templateRef: string;
  label: string;
  applicationRef: string;
  workflowDefinitionId: string;
  providerProfileRef: string;
  summary: string;
  nodes: WorkflowDraftDesignerNode[];
  edges: WorkflowDraftDesignerEdge[];
  designerLayout: WorkflowDraftDesignerLayout;
  readiness: WorkflowDraftDesignerReadiness[];
  risks: WorkflowDraftDesignerRisk[];
  blockedCapabilities: WorkflowDraftDesignerBlockedCapability[];
  routeMetadata: WorkflowDraftDesignerRouteMetadata;
  localOnlyInteraction: "inspect_only" | "local_edit";
};

export type WorkflowDraftDesignerSource = {
  workflowDefinitions?: WorkspaceWorkflowDefinitionRow[];
  localDrafts?: WorkflowDraftDesignerDraft[];
  tenantRef?: string;
  detailSourcesByWorkflowDefinitionId?: Record<string, WorkflowDraftDesignerDefinitionDetailSource>;
  detailNodes?: WorkflowDefinitionDetailNode[];
  detailEdges?: WorkflowDefinitionDetailEdge[];
  blockedActionPreview?: WorkflowDefinitionBlockedActionPreview;
  confirmationPlaceholder?: WorkflowConfirmationPlaceholderViewModel;
  sourceRequestId?: string;
  sourceAuditRef?: string;
};

export type WorkflowDraftDesignerDefinitionDetailSource = {
  nodes: WorkflowDefinitionDetailNode[];
  edges: WorkflowDefinitionDetailEdge[];
  blockedActionPreview: WorkflowDefinitionBlockedActionPreview;
};

export type WorkflowDraftDesignerViewModel = {
  pageId: "workflow-draft-designer-offline";
  sourcePageId: "workspace-workflow-definitions";
  sourceRouteId: "workflow-definition-summary-list-route";
  draftRouteId: "workflow-draft-designer-offline-draft";
  routePath: typeof CONTROL_PLANE_READ_ROUTES.workflowDefinitions;
  requestId: string;
  auditRef: string;
  defaultDraftId: string;
  templates: WorkflowDraftDesignerTemplate[];
  drafts: WorkflowDraftDesignerDraft[];
  forbiddenProjectionBlocked: boolean;
  canRenderDraftDesigner: boolean;
  canInspectDraftLocally: true;
  canSwitchDraftLocally: true;
  canRequestLiveBackend: false;
  canPersistDraft: false;
  canPublishWorkflow: false;
  canStartRun: false;
  canExecuteWorkflow: false;
  canSubmitConfirmationDecision: false;
  canWriteBusinessTruth: false;
  canReplayRun: false;
};

const DEFAULT_WORKFLOW_DEFINITIONS: WorkspaceWorkflowDefinitionRow[] = [
  {
    workflowDefinitionId: "wf_radishflow_copilot_latest",
    applicationRef: "app_flow_copilot",
    version: 3,
    definitionStatus: "published",
    nodeCount: 8,
    riskLevel: "medium",
    requiresConfirmationCapable: true,
    updatedAt: "2026-05-31T10:25:00Z",
  },
  {
    workflowDefinitionId: "wf_radish_docs_qa_draft",
    applicationRef: "app_docs_assistant",
    version: 1,
    definitionStatus: "draft",
    nodeCount: 4,
    riskLevel: "low",
    requiresConfirmationCapable: false,
    updatedAt: "2026-05-31T09:05:00Z",
  },
];
const DEFAULT_DRAFT_ID = "draft_wf_radishflow_copilot_latest_offline";

export function buildWorkflowDraftDesignerViewModel(
  source: WorkflowDraftDesignerSource = {},
): WorkflowDraftDesignerViewModel {
  const route = CONTROL_PLANE_READ_ROUTE_DEFINITIONS["workflow-definition-summary-list-route"];
  const routePath = CONTROL_PLANE_READ_ROUTES.workflowDefinitions;
  const detail = buildWorkflowDefinitionDetailViewModel();
  const confirmationPlaceholder = source.confirmationPlaceholder ?? buildWorkflowConfirmationPlaceholderViewModel();
  const workflowDefinitions = source.workflowDefinitions?.length
    ? source.workflowDefinitions
    : DEFAULT_WORKFLOW_DEFINITIONS;
  const templates = workflowDefinitions.map((workflowDefinition) =>
    buildTemplate(workflowDefinition, confirmationPlaceholder),
  );
  const requestId = source.sourceRequestId ?? "req_workflow_draft_designer_offline_demo";
  const auditRef = source.sourceAuditRef ?? "audit_workflow_draft_designer_offline_demo";
  const baseDrafts = templates.map((template, index) => {
    const workflowDefinition = workflowDefinitions[index];
    const detailSource = workflowDefinition
      ? detailSourceForWorkflowDefinition(workflowDefinition, source)
      : {
          nodes: detail.nodes,
          edges: detail.edges,
          blockedActionPreview: detail.blockedActionPreview,
        };
    return buildDraft(
      template,
      buildDesignerNodes(detailSource.nodes),
      buildDesignerEdges(detailSource.edges),
      detailSource.blockedActionPreview,
      confirmationPlaceholder,
      requestId,
      auditRef,
    );
  });
  const localDrafts = (source.localDrafts ?? []).filter((draft) =>
    workflowDefinitions.some(
      (workflowDefinition) => workflowDefinition.workflowDefinitionId === draft.workflowDefinitionId,
    ),
  );
  const drafts = [...baseDrafts, ...localDrafts];
  const forbiddenProjectionBlocked = controlPlaneReadResponseHasForbiddenOutput({
    draft: { [CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS[7]]: "blocked" },
  });

  return {
    pageId: "workflow-draft-designer-offline",
    sourcePageId: "workspace-workflow-definitions",
    sourceRouteId: "workflow-definition-summary-list-route",
    draftRouteId: "workflow-draft-designer-offline-draft",
    routePath,
    requestId,
    auditRef,
    defaultDraftId: drafts[0]?.draftId ?? DEFAULT_DRAFT_ID,
    templates,
    drafts,
    forbiddenProjectionBlocked,
    canRenderDraftDesigner:
      route.path === routePath &&
      route.canMutate === false &&
      templates.length >= 2 &&
      baseDrafts.length === templates.length &&
      baseDrafts.every(
        (draft, index) => draft.nodes.length === templates[index]?.nodeCount && draft.edges.length >= 3,
      ) &&
      localDrafts.every((draft) => draft.nodes.length > 0 && draft.edges.length >= 3),
    canInspectDraftLocally: true,
    canSwitchDraftLocally: true,
    canRequestLiveBackend: false,
    canPersistDraft: false,
    canPublishWorkflow: false,
    canStartRun: false,
    canExecuteWorkflow: false,
    canSubmitConfirmationDecision: false,
    canWriteBusinessTruth: false,
    canReplayRun: false,
  };
}

function detailSourceForWorkflowDefinition(
  workflowDefinition: WorkspaceWorkflowDefinitionRow,
  source: WorkflowDraftDesignerSource,
): WorkflowDraftDesignerDefinitionDetailSource {
  const explicitDetailSource =
    source.detailSourcesByWorkflowDefinitionId?.[workflowDefinition.workflowDefinitionId];
  if (explicitDetailSource) {
    return explicitDetailSource;
  }

  if (
    source.workflowDefinitions?.length === 1 &&
    source.detailNodes &&
    source.detailEdges &&
    source.blockedActionPreview
  ) {
    return {
      nodes: source.detailNodes,
      edges: source.detailEdges,
      blockedActionPreview: source.blockedActionPreview,
    };
  }

  const detail = buildWorkflowDefinitionDetailViewModel(
    toWorkflowDefinitionSummary(workflowDefinition, source.tenantRef ?? "tenant_demo"),
  );
  return {
    nodes: detail.nodes,
    edges: detail.edges,
    blockedActionPreview: detail.blockedActionPreview,
  };
}

function toWorkflowDefinitionSummary(
  workflowDefinition: WorkspaceWorkflowDefinitionRow,
  tenantRef: string,
): WorkflowDefinitionSummary {
  return {
    workflow_definition_id: workflowDefinition.workflowDefinitionId,
    tenant_ref: tenantRef,
    application_ref: workflowDefinition.applicationRef,
    version: workflowDefinition.version,
    definition_status: workflowDefinition.definitionStatus,
    node_count: workflowDefinition.nodeCount,
    risk_level: workflowDefinition.riskLevel,
    requires_confirmation_capable: workflowDefinition.requiresConfirmationCapable,
    updated_at: workflowDefinition.updatedAt,
  };
}

function buildTemplate(
  workflowDefinition: WorkspaceWorkflowDefinitionRow,
  confirmationPlaceholder: WorkflowConfirmationPlaceholderViewModel,
): WorkflowDraftDesignerTemplate {
  const riskLevel = normalizeRiskLevel(workflowDefinition.riskLevel);
  const isDocsAssistant = workflowDefinition.applicationRef === "app_docs_assistant";
  return {
    draftId: workflowDefinition.workflowDefinitionId === "wf_radishflow_copilot_latest"
      ? DEFAULT_DRAFT_ID
      : `draft_${workflowDefinition.workflowDefinitionId}_offline`,
    label: isDocsAssistant ? "Docs QA advisory draft" : "RadishFlow copilot draft",
    applicationRef: workflowDefinition.applicationRef,
    workflowDefinitionId: workflowDefinition.workflowDefinitionId,
    workflowKind: isDocsAssistant ? "docs_qa" : "workflow_copilot",
    providerProfileRef: isDocsAssistant ? "profile:radishmind-docs-qa" : "profile:radishmind-default-workflow",
    summary: isDocsAssistant
      ? "Question answering draft with retrieval, model reasoning, citation review, and advisory output lanes."
      : `Workflow copilot draft with policy gate, ${confirmationPlaceholder.confirmationPlaceholderId}, and blocked action preview lanes.`,
    riskLevel,
    nodeCount: workflowDefinition.nodeCount,
    status: workflowDefinition.requiresConfirmationCapable ? "needs_policy_review" : "ready_for_review",
  };
}

function buildDraft(
  template: WorkflowDraftDesignerTemplate,
  nodes: WorkflowDraftDesignerNode[],
  edges: WorkflowDraftDesignerEdge[],
  blockedActionPreview: WorkflowDefinitionBlockedActionPreview,
  confirmationPlaceholder: WorkflowConfirmationPlaceholderViewModel,
  requestId: string,
  auditRef: string,
): WorkflowDraftDesignerDraft {
  return {
    draftId: template.draftId,
    templateRef: template.workflowDefinitionId,
    label: template.label,
    applicationRef: template.applicationRef,
    workflowDefinitionId: template.workflowDefinitionId,
    providerProfileRef: template.providerProfileRef,
    summary: template.summary,
    nodes,
    edges,
    designerLayout: {
      source: "workflow_node_designer",
      persistence: "ui_only",
      nodePositions: [],
    },
    readiness: buildReadiness(template),
    risks: buildRisks(template, blockedActionPreview, confirmationPlaceholder),
    blockedCapabilities: buildBlockedCapabilities(template, confirmationPlaceholder),
    routeMetadata: {
      sourceRouteId: "workflow-definition-summary-list-route",
      draftRouteId: "workflow-draft-designer-offline-draft",
      routePath: CONTROL_PLANE_READ_ROUTES.workflowDefinitions,
      requestId,
      auditRef,
    },
    localOnlyInteraction: "inspect_only",
  };
}

function buildDesignerNodes(detailNodes: WorkflowDefinitionDetailNode[]): WorkflowDraftDesignerNode[] {
  return detailNodes.map((node) => ({
    nodeId: node.nodeId,
    label: node.label,
    nodeType: node.nodeType,
    lane: toLane(node),
    readiness: node.requiresConfirmation ? "review_required" : "ready",
    inputSummary: node.inputSummary,
    outputSummary: node.outputSummary,
    providerRef: defaultWorkflowDraftProviderRef(node.nodeType),
    toolRef: node.nodeType === "http_tool" ? "tool:workflow-preview-readonly" : "",
    ragRef: "",
    inputContractFields: deriveWorkflowDraftContractFields(node.inputSummary, "input"),
    outputContractFields: deriveWorkflowDraftContractFields(node.outputSummary, "output"),
    outputMappingSummary: defaultWorkflowDraftOutputMappingSummary(node.nodeType),
    riskLevel: node.riskLevel,
    requiresConfirmation: node.requiresConfirmation,
    previewOnlyReason: node.requiresConfirmation
      ? "The node is visible for local inspection, with future execution gated by confirmation and executor prerequisites."
      : "The node is visible for local inspection only; no runtime request is made.",
  }));
}

function defaultWorkflowDraftProviderRef(nodeType: WorkflowDefinitionDetailNode["nodeType"]): string {
  if (nodeType === "llm") {
    return "profile:radishmind-default-workflow";
  }
  if (nodeType === "condition") {
    return "policy:confirmation-gated";
  }
  return "";
}

function defaultWorkflowDraftOutputMappingSummary(nodeType: WorkflowDefinitionDetailNode["nodeType"]): string {
  if (nodeType === "llm") {
    return "Map advisory answer, candidate actions, risk summary, and audit refs into reviewable output fields.";
  }
  if (nodeType === "http_tool") {
    return "Map preview-only action metadata into audit-visible candidate action fields.";
  }
  if (nodeType === "condition") {
    return "Map policy result into review-required branch metadata without unlocking execution.";
  }
  if (nodeType === "output") {
    return "Map final advisory fields into the read-only workspace review surface.";
  }
  return "Map sanitized context fields into the next draft node contract.";
}

function deriveWorkflowDraftContractFields(
  summary: string,
  contractKind: "input" | "output",
): string[] {
  const summaryText = summary.toLowerCase();
  const candidates =
    contractKind === "input"
      ? ["tenant_ref", "application_ref", "workspace_id", "selection_summary", "diagnostic_summary"]
      : ["answer_summary", "candidate_actions", "risk_summary", "audit_refs", "citation_summary"];
  const matched = candidates.filter((field) => field.split("_").every((word) => summaryText.includes(word)));
  return matched.length > 0 ? matched : contractKind === "input" ? ["workspace_id"] : ["audit_refs"];
}

function buildDesignerEdges(detailEdges: WorkflowDefinitionDetailEdge[]): WorkflowDraftDesignerEdge[] {
  return detailEdges.map((edge) => ({
    edgeId: edge.edgeId,
    fromNodeId: edge.fromNodeId,
    toNodeId: edge.toNodeId,
    edgeKind: toEdgeKind(edge),
    conditionSummary: edge.conditionSummary,
  }));
}

function buildReadiness(template: WorkflowDraftDesignerTemplate): WorkflowDraftDesignerReadiness[] {
  return [
    {
      checkId: "draft_structure",
      label: "Draft structure",
      status: "ready",
      summary: `${template.nodeCount} nodes are projected from committed offline workflow definition metadata.`,
    },
    {
      checkId: "route_binding",
      label: "Route binding",
      status: "ready",
      summary: "The designer reuses the workflow definition summary route and keeps the draft route local-only.",
    },
    {
      checkId: "policy_review",
      label: "Policy review",
      status: template.riskLevel === "low" ? "ready" : "review_required",
      summary: "Risk and confirmation markers are displayed before any future builder or runtime work is allowed.",
    },
    {
      checkId: "runtime_boundary",
      label: "Runtime boundary",
      status: "blocked",
      summary: "Executor, persistence, publish, confirmation decision, writeback, and replay remain unavailable.",
    },
  ];
}

function buildRisks(
  template: WorkflowDraftDesignerTemplate,
  blockedActionPreview: WorkflowDefinitionBlockedActionPreview,
  confirmationPlaceholder: WorkflowConfirmationPlaceholderViewModel,
): WorkflowDraftDesignerRisk[] {
  return [
    {
      riskId: "candidate_action_risk",
      label: "Candidate action risk",
      riskLevel: template.riskLevel,
      requiresConfirmation: blockedActionPreview.requiresConfirmation,
      summary: blockedActionPreview.policyReason,
    },
    {
      riskId: "human_review_requirement",
      label: "Human review requirement",
      riskLevel: confirmationPlaceholder.riskLevel,
      requiresConfirmation: confirmationPlaceholder.humanReviewRequired,
      summary: confirmationPlaceholder.disabledReason,
    },
  ];
}

function buildBlockedCapabilities(
  template: WorkflowDraftDesignerTemplate,
  confirmationPlaceholder: WorkflowConfirmationPlaceholderViewModel,
): WorkflowDraftDesignerBlockedCapability[] {
  return [
    {
      capabilityId: "blocked_draft_persistence",
      label: "Draft persistence",
      status: "blocked",
      missingPrerequisite: "workflow builder storage task card",
      summary: "The selected draft exists only as an offline view model and is not written to durable storage.",
      auditRef: `audit_${template.draftId}_persistence_blocked`,
    },
    {
      capabilityId: "blocked_publish",
      label: "Workflow publish",
      status: "blocked",
      missingPrerequisite: "workflow lifecycle mutation task card",
      summary: "Version publishing is outside the current read-side workflow product surface.",
      auditRef: `audit_${template.draftId}_publish_blocked`,
    },
    {
      capabilityId: "blocked_runtime",
      label: "Workflow runtime",
      status: "blocked",
      missingPrerequisite: "workflow executor task card",
      summary: "The designer cannot start a run or call node executors.",
      auditRef: `audit_${template.draftId}_runtime_blocked`,
    },
    {
      capabilityId: "blocked_confirmation_decision",
      label: "Confirmation decision",
      status: "blocked",
      missingPrerequisite: confirmationPlaceholder.confirmationPlaceholderId,
      summary: "Human review shape is visible, but no decision submission or execution unlock path exists.",
      auditRef: confirmationPlaceholder.auditRef,
    },
  ];
}

function normalizeRiskLevel(value: string): "low" | "medium" | "high" {
  if (value === "high" || value === "medium" || value === "low") {
    return value;
  }
  return "medium";
}

function toLane(node: WorkflowDefinitionDetailNode): WorkflowDraftDesignerNode["lane"] {
  if (node.nodeType === "prompt") {
    return "context";
  }
  if (node.nodeType === "llm") {
    return "model";
  }
  if (node.nodeType === "condition") {
    return "policy";
  }
  if (node.nodeType === "http_tool") {
    return "preview";
  }
  return "output";
}

function toEdgeKind(edge: WorkflowDefinitionDetailEdge): WorkflowDraftDesignerEdge["edgeKind"] {
  if (edge.edgeId.includes("policy") || edge.toNodeId.includes("policy")) {
    return "policy";
  }
  if (edge.toNodeId.includes("preview") || edge.fromNodeId.includes("preview")) {
    return "preview";
  }
  if (edge.toNodeId.includes("audit")) {
    return "audit";
  }
  return "context";
}
