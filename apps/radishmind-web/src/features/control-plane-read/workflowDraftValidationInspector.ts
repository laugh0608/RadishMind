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
  type WorkflowDraftDesignerEdge,
  type WorkflowDraftDesignerNode,
} from "./workflowDraftDesigner";

export type WorkflowDraftValidationStatus = "passed" | "needs_review" | "blocked";
export type WorkflowDraftValidationSeverity = "info" | "warning" | "blocking";

export type WorkflowDraftValidationSummary = {
  label: string;
  value: string;
  status: WorkflowDraftValidationStatus;
  summary: string;
};

export type WorkflowDraftStructuralCheck = {
  checkId: string;
  label: string;
  status: WorkflowDraftValidationStatus;
  severity: WorkflowDraftValidationSeverity;
  summary: string;
  evidenceRefs: string[];
};

export type WorkflowDraftContractCheck = {
  checkId: string;
  label: string;
  status: WorkflowDraftValidationStatus;
  severity: WorkflowDraftValidationSeverity;
  requiredFields: string[];
  presentFields: string[];
  missingFields: string[];
  summary: string;
};

export type WorkflowDraftBlockedCapabilityCheck = {
  checkId: string;
  capabilityId: string;
  label: string;
  status: "blocked";
  severity: "blocking";
  missingPrerequisite: string;
  summary: string;
  auditRef: string;
};

export type WorkflowDraftValidationAuditMetadata = {
  sourceRouteId: "workflow-definition-summary-list-route";
  draftRouteId: "workflow-draft-validation-inspector-offline-draft";
  requestId: string;
  auditRef: string;
  inspectedDraftId: string;
};

export type WorkflowDraftValidationInspectorViewModel = {
  pageId: "workflow-draft-validation-inspector-offline";
  sourcePageId: "workflow-draft-designer-offline";
  sourceRouteId: "workflow-definition-summary-list-route";
  draftRouteId: "workflow-draft-validation-inspector-offline-draft";
  routePath: typeof CONTROL_PLANE_READ_ROUTES.workflowDefinitions;
  requestId: string;
  auditRef: string;
  inspectedDraftId: string;
  validationStatus: WorkflowDraftValidationStatus;
  summary: WorkflowDraftValidationSummary[];
  structuralChecks: WorkflowDraftStructuralCheck[];
  contractChecks: WorkflowDraftContractCheck[];
  blockedCapabilityChecks: WorkflowDraftBlockedCapabilityCheck[];
  auditMetadata: WorkflowDraftValidationAuditMetadata;
  forbiddenProjectionBlocked: boolean;
  canRenderDraftValidationInspector: boolean;
  canInspectDraftLocally: true;
  canRequestLiveBackend: false;
  canPersistDraft: false;
  canPublishWorkflow: false;
  canStartRun: false;
  canExecuteWorkflow: false;
  canSubmitConfirmationDecision: false;
  canWriteBusinessTruth: false;
  canReplayRun: false;
};

const REQUIRED_INPUT_FIELDS = ["tenant_ref", "application_ref", "selection_summary", "diagnostic_summary"];
const REQUIRED_OUTPUT_FIELDS = ["answer_summary", "candidate_actions", "risk_summary", "audit_refs"];

export function buildWorkflowDraftValidationInspectorViewModel(
  draft: WorkflowDraftDesignerDraft = buildWorkflowDraftDesignerViewModel().drafts[0]!,
): WorkflowDraftValidationInspectorViewModel {
  const route = CONTROL_PLANE_READ_ROUTE_DEFINITIONS["workflow-definition-summary-list-route"];
  const routePath = CONTROL_PLANE_READ_ROUTES.workflowDefinitions;
  const structuralChecks = buildStructuralChecks(draft.nodes, draft.edges);
  const contractChecks = buildContractChecks(draft.nodes);
  const blockedCapabilityChecks = buildBlockedCapabilityChecks(draft.blockedCapabilities);
  const validationStatus = deriveValidationStatus(structuralChecks, contractChecks, blockedCapabilityChecks);
  const forbiddenProjectionBlocked = controlPlaneReadResponseHasForbiddenOutput({
    validation: { [CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS[7]]: "blocked" },
  });

  return {
    pageId: "workflow-draft-validation-inspector-offline",
    sourcePageId: "workflow-draft-designer-offline",
    sourceRouteId: "workflow-definition-summary-list-route",
    draftRouteId: "workflow-draft-validation-inspector-offline-draft",
    routePath,
    requestId: draft.routeMetadata.requestId,
    auditRef: draft.routeMetadata.auditRef,
    inspectedDraftId: draft.draftId,
    validationStatus,
    summary: buildSummary(draft, validationStatus, structuralChecks, contractChecks, blockedCapabilityChecks),
    structuralChecks,
    contractChecks,
    blockedCapabilityChecks,
    auditMetadata: {
      sourceRouteId: "workflow-definition-summary-list-route",
      draftRouteId: "workflow-draft-validation-inspector-offline-draft",
      requestId: draft.routeMetadata.requestId,
      auditRef: draft.routeMetadata.auditRef,
      inspectedDraftId: draft.draftId,
    },
    forbiddenProjectionBlocked,
    canRenderDraftValidationInspector:
      route.path === routePath &&
      route.canMutate === false &&
      structuralChecks.length >= 5 &&
      contractChecks.length >= 2 &&
      blockedCapabilityChecks.length >= 4,
    canInspectDraftLocally: true,
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

function buildSummary(
  draft: WorkflowDraftDesignerDraft,
  validationStatus: WorkflowDraftValidationStatus,
  structuralChecks: WorkflowDraftStructuralCheck[],
  contractChecks: WorkflowDraftContractCheck[],
  blockedCapabilityChecks: WorkflowDraftBlockedCapabilityCheck[],
): WorkflowDraftValidationSummary[] {
  return [
    {
      label: "Draft",
      value: draft.draftId,
      status: validationStatus,
      summary: draft.summary,
    },
    {
      label: "Graph checks",
      value: `${countStatus(structuralChecks, "passed")}/${structuralChecks.length}`,
      status: structuralChecks.some((check) => check.status === "blocked") ? "blocked" : "needs_review",
      summary: "Structural checks cover lanes, edge continuity, policy routing, preview, output, and audit paths.",
    },
    {
      label: "Contract checks",
      value: `${countStatus(contractChecks, "passed")}/${contractChecks.length}`,
      status: contractChecks.every((check) => check.status === "passed") ? "passed" : "needs_review",
      summary: "Input and output contract fields stay explicit before any future runtime work.",
    },
    {
      label: "Blocked capabilities",
      value: String(blockedCapabilityChecks.length),
      status: "blocked",
      summary: "Persistence, publish, runtime, confirmation decision, writeback, and replay remain unavailable.",
    },
  ];
}

function buildStructuralChecks(
  nodes: WorkflowDraftDesignerNode[],
  edges: WorkflowDraftDesignerEdge[],
): WorkflowDraftStructuralCheck[] {
  const nodeIds = new Set(nodes.map((node) => node.nodeId));
  const connectedNodeIds = new Set(edges.flatMap((edge) => [edge.fromNodeId, edge.toNodeId]));
  const orphanNodeIds = [...nodeIds].filter((nodeId) => !connectedNodeIds.has(nodeId));
  const policyNodeIds = nodes.filter((node) => node.lane === "policy").map((node) => node.nodeId);
  const previewNodeIds = nodes.filter((node) => node.lane === "preview").map((node) => node.nodeId);
  const outputNodeIds = nodes.filter((node) => node.lane === "output").map((node) => node.nodeId);

  return [
    {
      checkId: "entry_context_lane",
      label: "Entry context lane",
      status: hasLane(nodes, "context") ? "passed" : "blocked",
      severity: hasLane(nodes, "context") ? "info" : "blocking",
      summary: "A draft must start from a context collection lane before model reasoning.",
      evidenceRefs: nodes.filter((node) => node.lane === "context").map((node) => node.nodeId),
    },
    {
      checkId: "model_reasoning_lane",
      label: "Model reasoning lane",
      status: hasLane(nodes, "model") ? "passed" : "blocked",
      severity: hasLane(nodes, "model") ? "info" : "blocking",
      summary: "A model lane is required for advisory reasoning and response construction.",
      evidenceRefs: nodes.filter((node) => node.lane === "model").map((node) => node.nodeId),
    },
    {
      checkId: "policy_gate_path",
      label: "Policy gate path",
      status: policyNodeIds.length > 0 && previewNodeIds.length > 0 ? "needs_review" : "blocked",
      severity: policyNodeIds.length > 0 && previewNodeIds.length > 0 ? "warning" : "blocking",
      summary: "Risk-bearing preview nodes must remain behind a policy gate and confirmation marker.",
      evidenceRefs: [...policyNodeIds, ...previewNodeIds],
    },
    {
      checkId: "output_audit_path",
      label: "Output and audit path",
      status: outputNodeIds.length >= 2 && edges.some((edge) => edge.edgeKind === "audit") ? "passed" : "blocked",
      severity: outputNodeIds.length >= 2 && edges.some((edge) => edge.edgeKind === "audit") ? "info" : "blocking",
      summary: "Advisory output and audit projection must remain visible in the draft.",
      evidenceRefs: outputNodeIds,
    },
    {
      checkId: "orphan_node_scan",
      label: "Orphan node scan",
      status: orphanNodeIds.length === 0 ? "passed" : "blocked",
      severity: orphanNodeIds.length === 0 ? "info" : "blocking",
      summary:
        orphanNodeIds.length === 0
          ? "Every projected node has at least one incoming or outgoing edge."
          : "One or more nodes are not connected to the draft graph.",
      evidenceRefs: orphanNodeIds.length === 0 ? [...nodeIds] : orphanNodeIds,
    },
  ];
}

function buildContractChecks(nodes: WorkflowDraftDesignerNode[]): WorkflowDraftContractCheck[] {
  const inputFields = collectFields(nodes, "inputSummary");
  const outputFields = collectFields(nodes, "outputSummary");
  return [
    {
      checkId: "input_contract_fields",
      label: "Input contract fields",
      status: inputFields.missingFields.length === 0 ? "passed" : "needs_review",
      severity: inputFields.missingFields.length === 0 ? "info" : "warning",
      requiredFields: REQUIRED_INPUT_FIELDS,
      presentFields: inputFields.presentFields,
      missingFields: inputFields.missingFields,
      summary: "The inspector checks that tenant, application, selection, and diagnostic context remain explicit.",
    },
    {
      checkId: "output_contract_fields",
      label: "Output contract fields",
      status: outputFields.missingFields.length === 0 ? "passed" : "needs_review",
      severity: outputFields.missingFields.length === 0 ? "info" : "warning",
      requiredFields: REQUIRED_OUTPUT_FIELDS,
      presentFields: outputFields.presentFields,
      missingFields: outputFields.missingFields,
      summary: "The inspector checks that advisory answers, candidate actions, risk summary, and audit refs remain explicit.",
    },
  ];
}

function buildBlockedCapabilityChecks(
  blockedCapabilities: WorkflowDraftDesignerBlockedCapability[],
): WorkflowDraftBlockedCapabilityCheck[] {
  return blockedCapabilities.map((capability) => ({
    checkId: `validation_${capability.capabilityId}`,
    capabilityId: capability.capabilityId,
    label: capability.label,
    status: "blocked",
    severity: "blocking",
    missingPrerequisite: capability.missingPrerequisite,
    summary: capability.summary,
    auditRef: capability.auditRef,
  }));
}

function deriveValidationStatus(
  structuralChecks: WorkflowDraftStructuralCheck[],
  contractChecks: WorkflowDraftContractCheck[],
  blockedCapabilityChecks: WorkflowDraftBlockedCapabilityCheck[],
): WorkflowDraftValidationStatus {
  if (structuralChecks.some((check) => check.status === "blocked")) {
    return "blocked";
  }
  if (blockedCapabilityChecks.length > 0 || contractChecks.some((check) => check.status !== "passed")) {
    return "needs_review";
  }
  return "passed";
}

function hasLane(nodes: WorkflowDraftDesignerNode[], lane: WorkflowDraftDesignerNode["lane"]): boolean {
  return nodes.some((node) => node.lane === lane);
}

function collectFields(
  nodes: WorkflowDraftDesignerNode[],
  key: "inputSummary" | "outputSummary",
): { presentFields: string[]; missingFields: string[] } {
  const explicitFields = nodes.flatMap((node) =>
    key === "inputSummary" ? node.inputContractFields : node.outputContractFields,
  );
  const explicitFieldSet = new Set(explicitFields.map((field) => field.toLowerCase()));
  const summaries = nodes
    .map((node) => {
      const contractFields = key === "inputSummary" ? node.inputContractFields : node.outputContractFields;
      return `${node[key]} ${contractFields.join(" ")} ${node.outputMappingSummary}`.toLowerCase();
    })
    .join(" ");
  const requiredFields = key === "inputSummary" ? REQUIRED_INPUT_FIELDS : REQUIRED_OUTPUT_FIELDS;
  const presentFields = requiredFields.filter(
    (field) => explicitFieldSet.has(field) || fieldIsPresent(summaries, field),
  );
  return {
    presentFields,
    missingFields: requiredFields.filter((field) => !presentFields.includes(field)),
  };
}

function fieldIsPresent(summary: string, field: string): boolean {
  const words = field.split("_");
  return words.every((word) => summary.includes(word));
}

function countStatus<T extends { status: WorkflowDraftValidationStatus }>(
  checks: T[],
  status: WorkflowDraftValidationStatus,
): number {
  return checks.filter((check) => check.status === status).length;
}
