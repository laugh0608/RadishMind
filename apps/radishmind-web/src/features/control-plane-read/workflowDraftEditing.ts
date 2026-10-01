import type { WorkflowDraftDesignerDraft, WorkflowDraftDesignerNode, WorkflowDraftDesignerEdge, WorkflowDraftDesignerLayout } from "./workflowDraftDesigner.ts";
export type WorkflowDraftNodeMoveDirection = "up" | "down";

type WorkflowDraftNodeTypeOption = {
  nodeType: WorkflowDraftDesignerNode["nodeType"];
  lane: WorkflowDraftDesignerNode["lane"];
  label: string;
  summary: string;
};

export const WORKFLOW_DRAFT_NODE_TYPE_OPTIONS: WorkflowDraftNodeTypeOption[] = [
  {
    nodeType: "prompt",
    lane: "context",
    label: "Context",
    summary: "Collects sanitized workspace, selection, and diagnostic context.",
  },
  {
    nodeType: "llm",
    lane: "model",
    label: "Model",
    summary: "Adds advisory reasoning without direct execution.",
  },
  {
    nodeType: "rag_retrieval",
    lane: "retrieval",
    label: "RAG Retrieval",
    summary: "Binds one exact immutable application knowledge snapshot version.",
  },
  {
    nodeType: "condition",
    lane: "policy",
    label: "Policy",
    summary: "Keeps risk and confirmation gates explicit.",
  },
  {
    nodeType: "http_tool",
    lane: "preview",
    label: "Preview",
    summary: "Models tool preview metadata while execution stays blocked.",
  },
  {
    nodeType: "output",
    lane: "output",
    label: "Output",
    summary: "Adds reviewable output or audit projection nodes.",
  },
];

function buildLocalWorkflowDraftNode(
  draft: WorkflowDraftDesignerDraft,
  nodeType: WorkflowDraftDesignerNode["nodeType"],
): WorkflowDraftDesignerNode {
  const option = workflowDraftNodeTypeOption(nodeType);
  const nodeNumber = nextWorkflowDraftNodeNumber(draft, nodeType);
  const nodeNumberLabel = String(nodeNumber).padStart(2, "0");
  const requiresConfirmation = nodeType === "condition" || nodeType === "http_tool";
  return {
    nodeId: uniqueWorkflowDraftNodeId(draft, nodeType, nodeNumber),
    label: `${option.label} ${nodeNumberLabel}`,
    nodeType,
    lane: option.lane,
    readiness: requiresConfirmation ? "review_required" : "ready",
    inputSummary: workflowDraftNodeInputSummary(option),
    outputSummary: workflowDraftNodeOutputSummary(option),
    providerRef: workflowDraftNodeProviderRef(option.nodeType),
    toolRef: option.nodeType === "http_tool" ? "tool:workflow-preview-readonly" : "",
    ragRef: "",
    inputContractFields: workflowDraftContractFieldsForNode(option.nodeType, "input"),
    outputContractFields: workflowDraftContractFieldsForNode(option.nodeType, "output"),
    outputMappingSummary: workflowDraftNodeOutputMappingSummary(option),
    riskLevel: requiresConfirmation ? "medium" : "low",
    requiresConfirmation,
    previewOnlyReason: "Local structure edit only; workflow execution remains blocked.",
  };
}

function parseWorkflowDraftContractFields(fieldsText: string): string[] {
  const seen = new Set<string>();
  return fieldsText
    .split(/[\n,]+/)
    .map((field) => field.toLowerCase().replace(/[^a-z0-9]+/g, "_").replace(/^_+|_+$/g, "").slice(0, 80))
    .filter((field) => {
      if (!field || seen.has(field)) {
        return false;
      }
      seen.add(field);
      return true;
    });
}

function workflowDraftWithStructureEdits(
  draft: WorkflowDraftDesignerDraft,
  nodes: WorkflowDraftDesignerNode[],
): WorkflowDraftDesignerDraft {
  return {
    ...draft,
    nodes,
    edges: rebuildWorkflowDraftEdges(nodes, draft.edges),
    designerLayout: workflowDraftLayoutForNodes(draft.designerLayout, nodes),
    localOnlyInteraction: "local_edit",
  };
}

function workflowDraftLayoutForNodes(
  layout: WorkflowDraftDesignerLayout,
  nodes: WorkflowDraftDesignerNode[],
): WorkflowDraftDesignerLayout {
  const nodeIds = new Set(nodes.map((node) => node.nodeId));
  return {
    source: "workflow_node_designer",
    persistence: "ui_only",
    nodePositions: layout.nodePositions.filter((position) => nodeIds.has(position.nodeId)),
  };
}

function workflowDraftLayoutWithNodePosition(
  draft: WorkflowDraftDesignerDraft,
  nodeId: string,
  x: number,
  y: number,
): WorkflowDraftDesignerLayout {
  const nodeIds = new Set(draft.nodes.map((node) => node.nodeId));
  const nextPosition = {
    nodeId,
    x: workflowDraftDesignerCoordinate(x),
    y: workflowDraftDesignerCoordinate(y),
  };
  const positions = draft.designerLayout.nodePositions
    .filter((position) => nodeIds.has(position.nodeId) && position.nodeId !== nodeId);
  return {
    source: "workflow_node_designer",
    persistence: "ui_only",
    nodePositions: [...positions, nextPosition],
  };
}

function workflowDraftDesignerCoordinate(value: number): number {
  if (!Number.isFinite(value)) {
    return 0;
  }
  return Math.max(-10000, Math.min(10000, Math.round(value)));
}

function insertWorkflowDraftNode(
  nodes: WorkflowDraftDesignerNode[],
  nextNode: WorkflowDraftDesignerNode,
): WorkflowDraftDesignerNode[] {
  if (nextNode.lane === "output") {
    return [...nodes, nextNode];
  }
  const firstOutputIndex = nodes.findIndex((node) => node.lane === "output");
  if (firstOutputIndex === -1) {
    return [...nodes, nextNode];
  }
  return [...nodes.slice(0, firstOutputIndex), nextNode, ...nodes.slice(firstOutputIndex)];
}

function canMoveWorkflowDraftNode(
  draft: WorkflowDraftDesignerDraft,
  nodeId: string,
  direction: WorkflowDraftNodeMoveDirection,
): boolean {
  const nodeIndex = draft.nodes.findIndex((node) => node.nodeId === nodeId);
  if (nodeIndex === -1) {
    return false;
  }
  return direction === "up" ? nodeIndex > 0 : nodeIndex < draft.nodes.length - 1;
}

function moveWorkflowDraftNode(
  nodes: WorkflowDraftDesignerNode[],
  nodeId: string,
  direction: WorkflowDraftNodeMoveDirection,
): WorkflowDraftDesignerNode[] {
  const nodeIndex = nodes.findIndex((node) => node.nodeId === nodeId);
  const nextIndex = direction === "up" ? nodeIndex - 1 : nodeIndex + 1;
  if (nodeIndex === -1 || nextIndex < 0 || nextIndex >= nodes.length) {
    return nodes;
  }
  const reorderedNodes = [...nodes];
  const movedNode = reorderedNodes[nodeIndex]!;
  reorderedNodes[nodeIndex] = reorderedNodes[nextIndex]!;
  reorderedNodes[nextIndex] = movedNode;
  return reorderedNodes;
}

export function canRemoveWorkflowDraftNode(draft: WorkflowDraftDesignerDraft, nodeId: string): boolean {
  const node = draft.nodes.find((candidate) => candidate.nodeId === nodeId);
  if (!node || draft.nodes.length <= 3) {
    return false;
  }
  const remainingNodes = draft.nodes.filter((candidate) => candidate.nodeId !== nodeId);
  if (!hasWorkflowDraftLane(remainingNodes, "context") || !hasWorkflowDraftLane(remainingNodes, "model")) {
    return false;
  }
  if (countWorkflowDraftLane(remainingNodes, "output") < 1) {
    return false;
  }
  return rebuildWorkflowDraftEdges(remainingNodes, draft.edges).length >= 3;
}

function rebuildWorkflowDraftEdges(
  nodes: WorkflowDraftDesignerNode[],
  previousEdges: WorkflowDraftDesignerEdge[],
): WorkflowDraftDesignerEdge[] {
  const rebuiltEdges = nodes.slice(1).map((node, index) =>
    buildWorkflowDraftEdge(nodes[index]!, node, previousEdges),
  );
  if (rebuiltEdges.some((edge) => edge.edgeKind === "audit")) {
    return rebuiltEdges;
  }
  const outputNodes = nodes.filter((node) => node.lane === "output");
  if (outputNodes.length < 2) {
    return rebuiltEdges;
  }
  return [
    ...rebuiltEdges,
    buildWorkflowDraftEdge(
      outputNodes[outputNodes.length - 2]!,
      outputNodes[outputNodes.length - 1]!,
      previousEdges,
      "audit",
    ),
  ];
}

function buildWorkflowDraftEdge(
  fromNode: WorkflowDraftDesignerNode,
  toNode: WorkflowDraftDesignerNode,
  previousEdges: WorkflowDraftDesignerEdge[],
  forcedEdgeKind?: WorkflowDraftDesignerEdge["edgeKind"],
): WorkflowDraftDesignerEdge {
  const previousEdge = previousEdges.find(
    (edge) => edge.fromNodeId === fromNode.nodeId && edge.toNodeId === toNode.nodeId,
  );
  const edgeKind = forcedEdgeKind ?? workflowDraftEdgeKindForConnection(fromNode, toNode);
  return {
    edgeId: previousEdge?.edgeId ?? workflowDraftEdgeId(fromNode.nodeId, toNode.nodeId, edgeKind),
    fromNodeId: fromNode.nodeId,
    toNodeId: toNode.nodeId,
    edgeKind,
    conditionSummary:
      workflowDraftNonEmptyConditionSummary(
        previousEdge?.conditionSummary,
        workflowDraftEdgeConditionSummary(fromNode, toNode, edgeKind),
      ),
  };
}

function buildWorkflowDraftEdgeForConnection(
  draft: WorkflowDraftDesignerDraft,
  fromNodeId: string,
  toNodeId: string,
): WorkflowDraftDesignerEdge | null {
  if (fromNodeId === toNodeId) {
    return null;
  }
  const fromNode = draft.nodes.find((node) => node.nodeId === fromNodeId);
  const toNode = draft.nodes.find((node) => node.nodeId === toNodeId);
  if (!fromNode || !toNode) {
    return null;
  }
  if (draft.edges.some((edge) => edge.fromNodeId === fromNodeId && edge.toNodeId === toNodeId)) {
    return null;
  }
  return buildWorkflowDraftEdge(fromNode, toNode, draft.edges);
}

function workflowDraftEdgeKindForConnection(
  fromNode: WorkflowDraftDesignerNode,
  toNode: WorkflowDraftDesignerNode,
): WorkflowDraftDesignerEdge["edgeKind"] {
  if (toNode.lane === "output" && (fromNode.lane === "output" || workflowDraftNodeLooksLikeAudit(toNode))) {
    return "audit";
  }
  if (toNode.lane === "preview" || fromNode.lane === "preview") {
    return "preview";
  }
  if (toNode.lane === "policy" || fromNode.lane === "policy") {
    return "policy";
  }
  return "context";
}

function workflowDraftEdgeConditionSummary(
  fromNode: WorkflowDraftDesignerNode,
  toNode: WorkflowDraftDesignerNode,
  edgeKind: WorkflowDraftDesignerEdge["edgeKind"],
): string {
  if (edgeKind === "audit") {
    return "Sanitized output metadata remains visible in the audit path after local graph editing.";
  }
  if (edgeKind === "preview") {
    return "Preview-only metadata flows forward while execution stays blocked.";
  }
  if (edgeKind === "policy") {
    return "Risk-bearing output remains behind policy and confirmation review markers.";
  }
  return `${fromNode.label} passes sanitized context to ${toNode.label}.`;
}

function workflowDraftReviewableEdgeConditionSummary(
  draft: WorkflowDraftDesignerDraft,
  edge: WorkflowDraftDesignerEdge,
  conditionSummary: string,
): string {
  const fromNode = draft.nodes.find((node) => node.nodeId === edge.fromNodeId);
  const toNode = draft.nodes.find((node) => node.nodeId === edge.toNodeId);
  const fallback =
    fromNode && toNode
      ? workflowDraftEdgeConditionSummary(fromNode, toNode, edge.edgeKind)
      : "Draft edge keeps a reviewable condition summary after local graph editing.";
  return workflowDraftNonEmptyConditionSummary(conditionSummary, fallback);
}

function workflowDraftNonEmptyConditionSummary(value: string | undefined, fallback: string): string {
  const normalized = value?.trim();
  return normalized ? normalized : fallback;
}

function uniqueWorkflowDraftNodeId(
  draft: WorkflowDraftDesignerDraft,
  nodeType: WorkflowDraftDesignerNode["nodeType"],
  initialNumber: number,
): string {
  const draftKey = workflowDraftSafeKey(draft.draftId, 32);
  let nodeNumber = initialNumber;
  let candidate = "";
  const existingNodeIds = new Set(draft.nodes.map((node) => node.nodeId));
  do {
    candidate = `node_${draftKey}_${nodeType}_${String(nodeNumber).padStart(2, "0")}`;
    nodeNumber += 1;
  } while (existingNodeIds.has(candidate));
  return candidate;
}

function workflowDraftEdgeId(
  fromNodeId: string,
  toNodeId: string,
  edgeKind: WorkflowDraftDesignerEdge["edgeKind"],
): string {
  return `edge_${workflowDraftSafeKey(fromNodeId, 36)}_to_${workflowDraftSafeKey(toNodeId, 36)}_${edgeKind}`;
}

function workflowDraftSafeKey(value: string, maxLength: number): string {
  const normalized = value.toLowerCase().replace(/[^a-z0-9]+/g, "_").replace(/^_+|_+$/g, "");
  return (normalized || "local").slice(0, maxLength);
}

function nextWorkflowDraftNodeNumber(
  draft: WorkflowDraftDesignerDraft,
  nodeType: WorkflowDraftDesignerNode["nodeType"],
): number {
  return draft.nodes.filter((node) => node.nodeType === nodeType).length + 1;
}

function workflowDraftNodeTypeOption(
  nodeType: WorkflowDraftDesignerNode["nodeType"],
): WorkflowDraftNodeTypeOption {
  return WORKFLOW_DRAFT_NODE_TYPE_OPTIONS.find((option) => option.nodeType === nodeType) ??
    WORKFLOW_DRAFT_NODE_TYPE_OPTIONS[0]!;
}

function workflowDraftNodeInputSummary(option: WorkflowDraftNodeTypeOption): string {
  if (option.nodeType === "prompt") {
    return "Tenant ref, application ref, selection summary, and diagnostic summary.";
  }
  if (option.nodeType === "llm") {
    return "Sanitized prompt context, answer contract, and provider profile reference.";
  }
  if (option.nodeType === "condition") {
    return "Candidate action shape, risk level, and confirmation policy marker.";
  }
  if (option.nodeType === "http_tool") {
    return "Sanitized candidate action payload without raw tool request body.";
  }
  return "Answer summary, risk summary, audit refs, and review context.";
}

function workflowDraftNodeOutputSummary(option: WorkflowDraftNodeTypeOption): string {
  if (option.nodeType === "prompt") {
    return "Sanitized context packet for advisory reasoning.";
  }
  if (option.nodeType === "llm") {
    return "Advisory answer, candidate actions, risk summary, and audit refs.";
  }
  if (option.nodeType === "condition") {
    return "Review-required branch metadata without execution unlock.";
  }
  if (option.nodeType === "http_tool") {
    return "Preview-only action metadata and audit reference.";
  }
  return "Read-only advisory output or sanitized audit projection.";
}

function workflowDraftNodeProviderRef(nodeType: WorkflowDraftDesignerNode["nodeType"]): string {
  if (nodeType === "llm") {
    return "profile:radishmind-default-workflow";
  }
  if (nodeType === "condition") {
    return "policy:confirmation-gated";
  }
  return "";
}

function workflowDraftContractFieldsForNode(
  nodeType: WorkflowDraftDesignerNode["nodeType"],
  contractKind: "input" | "output",
): string[] {
  if (contractKind === "input") {
    if (nodeType === "prompt") {
      return ["tenant_ref", "application_ref", "selection_summary", "diagnostic_summary"];
    }
    if (nodeType === "llm") {
      return ["prompt_context", "answer_contract", "provider_profile_ref"];
    }
    if (nodeType === "condition") {
      return ["candidate_action", "risk_level", "confirmation_policy"];
    }
    if (nodeType === "http_tool") {
      return ["candidate_action", "audit_refs"];
    }
    return ["answer_summary", "risk_summary", "audit_refs"];
  }
  if (nodeType === "prompt") {
    return ["prompt_context"];
  }
  if (nodeType === "llm") {
    return ["answer_summary", "candidate_actions", "risk_summary", "audit_refs"];
  }
  if (nodeType === "condition") {
    return ["policy_result", "requires_confirmation"];
  }
  if (nodeType === "http_tool") {
    return ["preview_action_metadata", "audit_refs"];
  }
  return ["answer_summary", "risk_summary", "audit_refs"];
}

function workflowDraftNodeOutputMappingSummary(option: WorkflowDraftNodeTypeOption): string {
  if (option.nodeType === "llm") {
    return "Map advisory answer, candidate actions, risk summary, and audit refs into reviewable output fields.";
  }
  if (option.nodeType === "condition") {
    return "Map policy result into review-required branch metadata without unlocking execution.";
  }
  if (option.nodeType === "http_tool") {
    return "Map preview-only action metadata into audit-visible candidate action fields.";
  }
  if (option.nodeType === "output") {
    return "Map advisory fields into the read-only workspace review surface.";
  }
  return "Map sanitized context fields into the next draft node contract.";
}

function hasWorkflowDraftLane(
  nodes: WorkflowDraftDesignerNode[],
  lane: WorkflowDraftDesignerNode["lane"],
): boolean {
  return nodes.some((node) => node.lane === lane);
}

function countWorkflowDraftLane(
  nodes: WorkflowDraftDesignerNode[],
  lane: WorkflowDraftDesignerNode["lane"],
): number {
  return nodes.filter((node) => node.lane === lane).length;
}

function workflowDraftNodeLooksLikeAudit(node: WorkflowDraftDesignerNode): boolean {
  return `${node.nodeId} ${node.label}`.toLowerCase().includes("audit");
}


export type WorkflowDraftEdit =
  | { type: "label"; label: string }
  | { type: "summary"; summary: string }
  | { type: "node"; nodeId: string; patch: Partial<WorkflowDraftDesignerNode> }
  | { type: "input_fields" | "output_fields"; nodeId: string; text: string }
  | { type: "position"; nodeId: string; x: number; y: number }
  | { type: "edge_condition"; edgeId: string; conditionSummary: string }
  | { type: "add_edge"; fromNodeId: string; toNodeId: string }
  | { type: "remove_edge"; edgeId: string }
  | { type: "add_node"; nodeType: WorkflowDraftDesignerNode["nodeType"] }
  | { type: "move_node"; nodeId: string; direction: WorkflowDraftNodeMoveDirection }
  | { type: "remove_node"; nodeId: string };

// Invalid structural edits preserve object identity so the caller does not mark them dirty.
export function applyWorkflowDraftEdit(draft: WorkflowDraftDesignerDraft, edit: WorkflowDraftEdit): WorkflowDraftDesignerDraft {
  switch (edit.type) {
    case "label": return { ...draft, label: edit.label, localOnlyInteraction: "local_edit" };
    case "summary": return { ...draft, summary: edit.summary, localOnlyInteraction: "local_edit" };
    case "node":
    case "input_fields":
    case "output_fields": {
      if (!draft.nodes.some((node) => node.nodeId === edit.nodeId)) return draft;
      const patch = edit.type === "node" ? edit.patch : edit.type === "input_fields"
        ? { inputContractFields: parseWorkflowDraftContractFields(edit.text) }
        : { outputContractFields: parseWorkflowDraftContractFields(edit.text) };
      return { ...draft, localOnlyInteraction: "local_edit", nodes: draft.nodes.map((node) => node.nodeId === edit.nodeId ? { ...node, ...patch } : node) };
    }
    case "position":
      if (!draft.nodes.some((node) => node.nodeId === edit.nodeId)) return draft;
      return { ...draft, localOnlyInteraction: "local_edit", designerLayout: workflowDraftLayoutWithNodePosition(draft, edit.nodeId, edit.x, edit.y) };
    case "edge_condition":
      if (!draft.edges.some((edge) => edge.edgeId === edit.edgeId)) return draft;
      return { ...draft, localOnlyInteraction: "local_edit", edges: draft.edges.map((edge) => edge.edgeId === edit.edgeId
        ? { ...edge, conditionSummary: workflowDraftReviewableEdgeConditionSummary(draft, edge, edit.conditionSummary) } : edge) };
    case "add_edge": {
      const edge = buildWorkflowDraftEdgeForConnection(draft, edit.fromNodeId, edit.toNodeId);
      return edge ? { ...draft, localOnlyInteraction: "local_edit", edges: [...draft.edges, edge] } : draft;
    }
    case "remove_edge":
      return draft.edges.some((edge) => edge.edgeId === edit.edgeId)
        ? { ...draft, localOnlyInteraction: "local_edit", edges: draft.edges.filter((edge) => edge.edgeId !== edit.edgeId) } : draft;
    case "add_node": return workflowDraftWithStructureEdits(draft, insertWorkflowDraftNode(draft.nodes, buildLocalWorkflowDraftNode(draft, edit.nodeType)));
    case "move_node": return canMoveWorkflowDraftNode(draft, edit.nodeId, edit.direction)
      ? workflowDraftWithStructureEdits(draft, moveWorkflowDraftNode(draft.nodes, edit.nodeId, edit.direction)) : draft;
    case "remove_node": return canRemoveWorkflowDraftNode(draft, edit.nodeId)
      ? workflowDraftWithStructureEdits(draft, draft.nodes.filter((node) => node.nodeId !== edit.nodeId)) : draft;
  }
}
