import { useCallback, useEffect, useMemo, useState } from "react";
import {
  Background,
  BaseEdge,
  Controls,
  Handle,
  MarkerType,
  MiniMap,
  Position,
  ReactFlow,
  applyNodeChanges,
  getSmoothStepPath,
  type Connection,
  type Edge,
  type EdgeProps,
  type Node,
  type NodeChange,
  type NodeProps,
  type NodeTypes,
} from "@xyflow/react";

import type {
  WorkflowDraftDesignerDraft,
  WorkflowDraftDesignerEdge,
  WorkflowDraftDesignerNode,
} from "./workflowDraftDesigner";
import type {
  WorkflowDraftValidationInspectorViewModel,
  WorkflowDraftValidationSeverity,
  WorkflowDraftValidationStatus,
} from "./workflowDraftValidationInspector";

type WorkflowNodeDesignerEdgeKind = "data_edge" | "control_edge" | "guard_edge" | "audit_edge";

type WorkflowNodeDesignerNodeData = {
  draftNodeId: string;
  label: string;
  nodeType: WorkflowDraftDesignerNode["nodeType"];
  lane: WorkflowDraftDesignerNode["lane"];
  readiness: WorkflowDraftDesignerNode["readiness"];
  inputSummary: string;
  outputSummary: string;
  providerRef: string;
  toolRef: string;
  ragRef: string;
  riskLevel: WorkflowDraftDesignerNode["riskLevel"];
  requiresConfirmation: boolean;
  previewOnlyReason: string;
  protectedNode: boolean;
  validationFocus?: "none" | "focused";
  validationSeverity?: WorkflowDraftValidationSeverity;
};

type WorkflowNodeDesignerEdgeData = {
  edgeKind: WorkflowNodeDesignerEdgeKind;
  conditionSummary: string;
  validationFocus?: "none" | "focused";
  validationSeverity?: WorkflowDraftValidationSeverity;
};

type WorkflowNodeDesignerFeedback = {
  tone: "neutral" | "good" | "bad";
  message: string;
};

type WorkflowNodeDesignerValidationFocus = {
  checkId: string;
  label: string;
  severity: WorkflowDraftValidationSeverity;
  nodeIds: string[];
  edgeIds: string[];
};

type WorkflowNodeDesignerValidationNavItem = {
  checkId: string;
  label: string;
  status: WorkflowDraftValidationStatus;
  severity: WorkflowDraftValidationSeverity;
  summary: string;
  targetNodeIds: string[];
  targetEdgeIds: string[];
  targetSummary: string;
};

type WorkflowNodeDesignerNode = Node<WorkflowNodeDesignerNodeData, "workflowDraftNode">;
type WorkflowNodeDesignerEdge = Edge<WorkflowNodeDesignerEdgeData, "workflowDraftEdge">;

type WorkflowNodeDesignerProps = {
  draft: WorkflowDraftDesignerDraft;
  validationInspector: WorkflowDraftValidationInspectorViewModel;
  editingDisabled: boolean;
  canRemoveNode: (nodeId: string) => boolean;
  onUpdateNodeLabel: (nodeId: string, label: string) => void;
  onUpdateNodeInputSummary: (nodeId: string, inputSummary: string) => void;
  onUpdateNodeOutputSummary: (nodeId: string, outputSummary: string) => void;
  onUpdateNodeProviderRef: (nodeId: string, providerRef: string) => void;
  onUpdateNodeToolRef: (nodeId: string, toolRef: string) => void;
  onUpdateNodeRagRef: (nodeId: string, ragRef: string) => void;
  onUpdateNodeOutputMapping: (nodeId: string, outputMappingSummary: string) => void;
  onUpdateNodeDesignerPosition: (nodeId: string, x: number, y: number) => void;
  onAddEdge: (fromNodeId: string, toNodeId: string) => boolean;
  onRemoveEdge: (edgeId: string) => boolean;
  onRemoveNode: (nodeId: string) => void;
};

const LANE_X: Record<WorkflowDraftDesignerNode["lane"], number> = {
  context: 0,
  model: 260,
  policy: 520,
  preview: 780,
  output: 1040,
};

const EDGE_KIND_CLASS: Record<WorkflowNodeDesignerEdgeKind, string> = {
  data_edge: "data",
  control_edge: "control",
  guard_edge: "guard",
  audit_edge: "audit",
};

const nodeTypes = {
  workflowDraftNode: WorkflowNodeDesignerNodeCard,
} satisfies NodeTypes;

const edgeTypes = {
  workflowDraftEdge: WorkflowNodeDesignerEdgePath,
};

export function WorkflowNodeDesigner({
  draft,
  validationInspector,
  editingDisabled,
  canRemoveNode,
  onUpdateNodeLabel,
  onUpdateNodeInputSummary,
  onUpdateNodeOutputSummary,
  onUpdateNodeProviderRef,
  onUpdateNodeToolRef,
  onUpdateNodeRagRef,
  onUpdateNodeOutputMapping,
  onUpdateNodeDesignerPosition,
  onAddEdge,
  onRemoveEdge,
  onRemoveNode,
}: WorkflowNodeDesignerProps) {
  const initialNodes = useMemo(() => buildWorkflowNodeDesignerNodes(draft, canRemoveNode), [draft, canRemoveNode]);
  const edges = useMemo(() => buildWorkflowNodeDesignerEdges(draft), [draft]);
  const validationNavigationItems = useMemo(
    () => buildWorkflowNodeDesignerValidationNavigation(draft, validationInspector),
    [draft, validationInspector],
  );
  const [nodes, setNodes] = useState(initialNodes);
  const [selectedNodeId, setSelectedNodeId] = useState(draft.nodes[0]?.nodeId ?? "");
  const [validationFocus, setValidationFocus] = useState<WorkflowNodeDesignerValidationFocus | null>(null);
  const [interactionFeedback, setInteractionFeedback] = useState<WorkflowNodeDesignerFeedback>({
    tone: "neutral",
    message: "Connect typed ports to add controlled draft edges.",
  });

  useEffect(() => {
    setNodes(initialNodes);
    setSelectedNodeId((current) =>
      draft.nodes.some((node) => node.nodeId === current) ? current : draft.nodes[0]?.nodeId ?? "",
    );
  }, [draft, initialNodes]);

  useEffect(() => {
    setValidationFocus((current) => {
      if (!current) {
        return null;
      }
      const nextItem = validationNavigationItems.find((item) => item.checkId === current.checkId);
      if (!nextItem) {
        return null;
      }
      return {
        checkId: nextItem.checkId,
        label: nextItem.label,
        severity: nextItem.severity,
        nodeIds: nextItem.targetNodeIds,
        edgeIds: nextItem.targetEdgeIds,
      };
    });
  }, [validationNavigationItems]);

  const selectedNode = draft.nodes.find((node) => node.nodeId === selectedNodeId) ?? draft.nodes[0];
  const selectedNodeRemovable = selectedNode ? canRemoveNode(selectedNode.nodeId) : false;

  const onNodesChange = useCallback((changes: NodeChange<WorkflowNodeDesignerNode>[]) => {
    setNodes((currentNodes) => applyNodeChanges(changes, currentNodes));
  }, []);

  const isValidConnection = useCallback(
    (connection: Connection | WorkflowNodeDesignerEdge) => validateWorkflowNodeDesignerConnection(connection, draft),
    [draft],
  );

  const onConnect = useCallback(
    (connection: Connection) => {
      if (editingDisabled) {
        setInteractionFeedback({
          tone: "bad",
          message: "Connection rejected: local draft edits are locked while a saved draft operation is pending.",
        });
        return;
      }
      const valid = validateWorkflowNodeDesignerConnection(connection, draft);
      if (!valid || !connection.source || !connection.target) {
        setInteractionFeedback({
          tone: "bad",
          message: "Connection rejected: typed ports require distinct source and target draft nodes with no duplicate pair.",
        });
        return;
      }
      const source = draft.nodes.find((node) => node.nodeId === connection.source);
      const target = draft.nodes.find((node) => node.nodeId === connection.target);
      const edgeKind = source && target ? deriveWorkflowNodeDesignerEdgeKind(source, target) : "data_edge";
      const added = onAddEdge(connection.source, connection.target);
      if (!added) {
        setInteractionFeedback({
          tone: "bad",
          message: "Connection rejected: active draft endpoints are unavailable or already connected.",
        });
        return;
      }
      setInteractionFeedback({
        tone: "good",
        message: `Added draft edge: ${connection.source} to ${connection.target} is tracked as ${edgeKind}.`,
      });
    },
    [draft, editingDisabled, onAddEdge],
  );

  const selectNode = useCallback(
    (nodeId: string) => {
      const nextNode = draft.nodes.find((draftNode) => draftNode.nodeId === nodeId);
      if (!nextNode) {
        return;
      }
      setValidationFocus(null);
      setSelectedNodeId(nextNode.nodeId);
      setInteractionFeedback({
        tone: "neutral",
        message: `Selected node: ${nextNode.label} (${nextNode.nodeId}).`,
      });
    },
    [draft.nodes],
  );

  const focusValidationFinding = useCallback((item: WorkflowNodeDesignerValidationNavItem) => {
    const firstNodeId = item.targetNodeIds[0];
    setValidationFocus({
      checkId: item.checkId,
      label: item.label,
      severity: item.severity,
      nodeIds: item.targetNodeIds,
      edgeIds: item.targetEdgeIds,
    });
    if (firstNodeId) {
      setSelectedNodeId(firstNodeId);
    }
    setInteractionFeedback({
      tone: workflowNodeDesignerFeedbackToneForValidation(item.status, item.severity),
      message: firstNodeId
        ? `Focused validation finding: ${item.label} targets ${item.targetSummary}.`
        : `Focused validation finding: ${item.label}; no graph target is available.`,
    });
  }, []);

  const clearValidationFocus = useCallback(() => {
    setValidationFocus(null);
    setInteractionFeedback({
      tone: "neutral",
      message: "Validation overlay focus cleared; canvas selection remains UI-only.",
    });
  }, []);

  const onNodeClick = useCallback((_: unknown, node: WorkflowNodeDesignerNode) => {
    selectNode(node.data.draftNodeId);
  }, [selectNode]);

  const onNodeDragStop = useCallback(
    (_: unknown, node: WorkflowNodeDesignerNode) => {
      onUpdateNodeDesignerPosition(node.data.draftNodeId, node.position.x, node.position.y);
      setInteractionFeedback({
        tone: "good",
        message: `Saved canvas position for ${node.data.label} as active draft layout metadata.`,
      });
    },
    [onUpdateNodeDesignerPosition],
  );
  const onRemoveDraftEdge = useCallback(
    (edgeId: string) => {
      const removed = onRemoveEdge(edgeId);
      setInteractionFeedback({
        tone: removed ? "good" : "bad",
        message: removed
          ? `Removed draft edge: ${edgeId}. Validation inspector will recompute graph findings from the active draft.`
          : `Edge removal rejected: ${edgeId} is not in the active draft.`,
      });
      return removed;
    },
    [onRemoveEdge],
  );

  const mappedLayoutCount = draft.designerLayout.nodePositions.filter((position) =>
    draft.nodes.some((node) => node.nodeId === position.nodeId),
  ).length;
  const selectedNodeEdges = selectedNode
    ? draft.edges.filter((edge) => edge.fromNodeId === selectedNode.nodeId || edge.toNodeId === selectedNode.nodeId)
    : [];
  const displayedNodes = useMemo(
    () =>
      nodes.map((node) => ({
        ...node,
        selected: node.data.draftNodeId === selectedNodeId,
        data: {
          ...node.data,
          validationFocus: workflowNodeDesignerValidationFocusState(
            validationFocus?.nodeIds.includes(node.data.draftNodeId) ?? false,
          ),
          validationSeverity: validationFocus?.severity,
        },
      })),
    [nodes, selectedNodeId, validationFocus],
  );
  const displayedEdges = useMemo(
    () =>
      edges.map((edge) => ({
        ...edge,
        data: {
          ...(edge.data ?? { edgeKind: "data_edge", conditionSummary: "" }),
          validationFocus: workflowNodeDesignerValidationFocusState(validationFocus?.edgeIds.includes(edge.id) ?? false),
          validationSeverity: validationFocus?.severity,
        },
      })),
    [edges, validationFocus],
  );
  const layoutPersistenceLabel =
    draft.designerLayout.persistence === "saved_draft_metadata" ? "restored saved draft layout" : "active draft layout";
  const layoutPersistenceSummary =
    draft.designerLayout.persistence === "saved_draft_metadata"
      ? "Node positions were restored from saved draft layout metadata; viewport and selection remain transient."
      : "Save Draft writes sanitized node positions as saved draft layout metadata; viewport and selection remain transient.";
  const editingStateLabel = editingDisabled ? "Editing locked" : "Editing enabled";
  const editingStateSummary = editingDisabled
    ? "Saved draft operation pending; canvas selection remains available without local mutation."
    : "Drag nodes, connect typed ports, or edit inspector fields on the active draft.";

  return (
    <section className="workflow-node-designer" aria-label="Workflow node designer canvas">
      <div className="workflow-node-designer-toolbar">
        <div>
          <p className="eyebrow">Node Designer Canvas</p>
          <h5>{draft.label}</h5>
        </div>
        <div className="workflow-node-designer-status">
          <span>{draft.localOnlyInteraction}</span>
          <strong>{draft.nodes.length} nodes / {draft.edges.length} derived edges</strong>
        </div>
      </div>

      <div className="workflow-node-designer-mapping-summary" aria-label="Workflow node designer saved draft mapping">
        <article>
          <span>Saved draft mapping</span>
          <strong>Attributes and edge endpoints</strong>
          <p>Save Draft writes node attributes, contract fields, edge endpoints, and condition summaries.</p>
        </article>
        <article>
          <span>Layout metadata</span>
          <strong>{mappedLayoutCount} positioned nodes</strong>
          <p>{layoutPersistenceLabel}: {layoutPersistenceSummary}</p>
        </article>
        <article>
          <span>Derived edge kind</span>
          <strong>Not persisted</strong>
          <p>Canvas edge colors are derived from node lane, risk, policy, and audit context.</p>
        </article>
      </div>

      <div className="workflow-node-designer-interaction-bar" aria-label="Workflow node designer interaction state">
        <div
          className={`workflow-node-designer-feedback ${interactionFeedback.tone}`}
          role="status"
          aria-live="polite"
        >
          <span>{editingStateLabel}</span>
          <strong>{interactionFeedback.message}</strong>
          <p>{editingStateSummary}</p>
        </div>
        <div className="workflow-node-designer-node-switcher" aria-label="Select workflow draft node">
          <span>Selected node</span>
          <div className="workflow-node-designer-node-switcher-list">
            {draft.nodes.map((node) => (
              <button
                key={node.nodeId}
                type="button"
                className={node.nodeId === selectedNodeId ? "selected" : ""}
                aria-pressed={node.nodeId === selectedNodeId}
                onClick={() => selectNode(node.nodeId)}
              >
                <strong>{node.label}</strong>
                <small>{node.nodeType}</small>
              </button>
            ))}
          </div>
        </div>
      </div>

      <div
        className="workflow-node-designer-validation-navigation"
        aria-label="Workflow node designer validation overlay navigation"
      >
        <div className="workflow-node-designer-validation-navigation-heading">
          <div>
            <span>Validation overlay</span>
            <strong>
              {validationInspector.validationStatus} / {validationNavigationItems.length} findings
            </strong>
          </div>
          <button type="button" disabled={!validationFocus} onClick={clearValidationFocus}>
            Clear focus
          </button>
        </div>
        <div className="workflow-node-designer-validation-navigation-list">
          {validationNavigationItems.map((item) => (
            <button
              key={item.checkId}
              type="button"
              className={`${item.status} ${item.severity} ${
                validationFocus?.checkId === item.checkId ? "focused" : ""
              }`}
              aria-pressed={validationFocus?.checkId === item.checkId}
              onClick={() => focusValidationFinding(item)}
            >
              <span>{item.status}</span>
              <strong>{item.label}</strong>
              <small>{item.targetSummary}</small>
            </button>
          ))}
        </div>
      </div>

      <div className="workflow-node-designer-shell">
        <div className={`workflow-node-designer-canvas ${editingDisabled ? "locked" : "editable"}`}>
          <ReactFlow
            nodes={displayedNodes}
            edges={displayedEdges}
            nodeTypes={nodeTypes}
            edgeTypes={edgeTypes}
            nodesDraggable={!editingDisabled}
            nodesConnectable={!editingDisabled}
            elementsSelectable
            fitView
            minZoom={0.45}
            maxZoom={1.4}
            onNodesChange={onNodesChange}
            onNodeClick={onNodeClick}
            onNodeDragStop={onNodeDragStop}
            onConnect={onConnect}
            isValidConnection={isValidConnection}
          >
            <Background gap={24} size={1} />
            <MiniMap pannable zoomable nodeStrokeWidth={3} />
            <Controls showInteractive={false} />
          </ReactFlow>
        </div>

        <aside className="workflow-node-designer-inspector" aria-label="Workflow node designer inspector">
          {selectedNode ? (
            <WorkflowNodeDesignerInspector
              node={selectedNode}
              editingDisabled={editingDisabled}
              canDelete={selectedNodeRemovable}
              interactionFeedback={interactionFeedback}
              edges={selectedNodeEdges}
              onUpdateNodeLabel={onUpdateNodeLabel}
              onUpdateNodeInputSummary={onUpdateNodeInputSummary}
              onUpdateNodeOutputSummary={onUpdateNodeOutputSummary}
              onUpdateNodeProviderRef={onUpdateNodeProviderRef}
              onUpdateNodeToolRef={onUpdateNodeToolRef}
              onUpdateNodeRagRef={onUpdateNodeRagRef}
              onUpdateNodeOutputMapping={onUpdateNodeOutputMapping}
              onRemoveEdge={onRemoveDraftEdge}
              onRemoveNode={onRemoveNode}
            />
          ) : (
            <div className="workflow-node-designer-empty">
              <strong>Node unavailable</strong>
              <p>The list editor remains available and no sample fallback has been applied.</p>
            </div>
          )}
        </aside>
      </div>
    </section>
  );
}

function buildWorkflowNodeDesignerNodes(
  draft: WorkflowDraftDesignerDraft,
  canRemoveNode: (nodeId: string) => boolean,
): WorkflowNodeDesignerNode[] {
  const laneCounts = new Map<WorkflowDraftDesignerNode["lane"], number>();
  const positionsByNodeId = new Map(
    draft.designerLayout.nodePositions.map((position) => [position.nodeId, position]),
  );
  return draft.nodes.map((node, index) => {
    const laneIndex = laneCounts.get(node.lane) ?? 0;
    laneCounts.set(node.lane, laneIndex + 1);
    const savedPosition = positionsByNodeId.get(node.nodeId);
    return {
      id: node.nodeId,
      type: "workflowDraftNode",
      position: savedPosition
        ? { x: savedPosition.x, y: savedPosition.y }
        : {
            x: LANE_X[node.lane],
            y: laneIndex * 190 + (index % 2) * 18,
          },
      data: {
        draftNodeId: node.nodeId,
        label: node.label,
        nodeType: node.nodeType,
        lane: node.lane,
        readiness: node.readiness,
        inputSummary: node.inputSummary,
        outputSummary: node.outputSummary,
        providerRef: node.providerRef,
        toolRef: node.toolRef,
        ragRef: node.ragRef,
        riskLevel: node.riskLevel,
        requiresConfirmation: node.requiresConfirmation,
        previewOnlyReason: node.previewOnlyReason,
        protectedNode: !canRemoveNode(node.nodeId),
      },
    };
  });
}

function buildWorkflowNodeDesignerEdges(draft: WorkflowDraftDesignerDraft): WorkflowNodeDesignerEdge[] {
  return draft.edges.map((edge) => {
    const source = draft.nodes.find((node) => node.nodeId === edge.fromNodeId);
    const target = draft.nodes.find((node) => node.nodeId === edge.toNodeId);
    const edgeKind = source && target ? deriveWorkflowNodeDesignerEdgeKind(source, target, edge.edgeKind) : "data_edge";
    return {
      id: edge.edgeId,
      source: edge.fromNodeId,
      target: edge.toNodeId,
      type: "workflowDraftEdge",
      markerEnd: {
        type: MarkerType.ArrowClosed,
      },
      data: {
        edgeKind,
        conditionSummary: edge.conditionSummary,
      },
      className: `workflow-node-designer-edge ${EDGE_KIND_CLASS[edgeKind]}`,
    };
  });
}

function buildWorkflowNodeDesignerValidationNavigation(
  draft: WorkflowDraftDesignerDraft,
  validationInspector: WorkflowDraftValidationInspectorViewModel,
): WorkflowNodeDesignerValidationNavItem[] {
  const nodeIds = new Set(draft.nodes.map((node) => node.nodeId));
  const structuralItems = validationInspector.structuralChecks.map((check) => {
    const targetNodeIds = check.evidenceRefs.filter((nodeId) => nodeIds.has(nodeId));
    const targetEdgeIds = workflowNodeDesignerEdgeIdsForTargets(draft, targetNodeIds);
    return {
      checkId: check.checkId,
      label: check.label,
      status: check.status,
      severity: check.severity,
      summary: check.summary,
      targetNodeIds,
      targetEdgeIds,
      targetSummary: workflowNodeDesignerValidationTargetSummary(targetNodeIds, targetEdgeIds),
    };
  });
  const contractItems = validationInspector.contractChecks.map((check) => {
    const targetNodeIds = workflowNodeDesignerContractTargetNodeIds(draft, check.checkId);
    const targetEdgeIds = workflowNodeDesignerEdgeIdsForTargets(draft, targetNodeIds);
    return {
      checkId: check.checkId,
      label: check.label,
      status: check.status,
      severity: check.severity,
      summary: check.summary,
      targetNodeIds,
      targetEdgeIds,
      targetSummary: workflowNodeDesignerValidationTargetSummary(targetNodeIds, targetEdgeIds),
    };
  });
  return [...structuralItems, ...contractItems];
}

function workflowNodeDesignerContractTargetNodeIds(
  draft: WorkflowDraftDesignerDraft,
  checkId: string,
): string[] {
  if (checkId === "input_contract_fields") {
    return draft.nodes
      .filter((node) => node.lane === "context" || node.inputContractFields.length > 0)
      .map((node) => node.nodeId);
  }
  if (checkId === "output_contract_fields") {
    return draft.nodes
      .filter(
        (node) =>
          node.lane === "output" ||
          node.outputContractFields.length > 0 ||
          node.outputMappingSummary.trim().length > 0,
      )
      .map((node) => node.nodeId);
  }
  return [];
}

function workflowNodeDesignerEdgeIdsForTargets(
  draft: WorkflowDraftDesignerDraft,
  targetNodeIds: string[],
): string[] {
  const targetNodeIdSet = new Set(targetNodeIds);
  return draft.edges
    .filter((edge) => targetNodeIdSet.has(edge.fromNodeId) || targetNodeIdSet.has(edge.toNodeId))
    .map((edge) => edge.edgeId);
}

function workflowNodeDesignerValidationTargetSummary(nodeIds: string[], edgeIds: string[]): string {
  if (nodeIds.length === 0 && edgeIds.length === 0) {
    return "No graph target";
  }
  return `${nodeIds.length} nodes / ${edgeIds.length} edges`;
}

function workflowNodeDesignerFeedbackToneForValidation(
  status: WorkflowDraftValidationStatus,
  severity: WorkflowDraftValidationSeverity,
): WorkflowNodeDesignerFeedback["tone"] {
  if (severity === "blocking") {
    return "bad";
  }
  if (status === "passed") {
    return "good";
  }
  return "neutral";
}

function workflowNodeDesignerValidationFocusState(isFocused: boolean): "focused" | "none" {
  return isFocused ? "focused" : "none";
}

function deriveWorkflowNodeDesignerEdgeKind(
  source: WorkflowDraftDesignerNode,
  target: WorkflowDraftDesignerNode,
  draftEdgeKind?: string,
): WorkflowNodeDesignerEdgeKind {
  if (draftEdgeKind === "audit" || source.lane === "output" || target.lane === "output") {
    return "audit_edge";
  }
  if (draftEdgeKind === "policy" || source.requiresConfirmation || target.requiresConfirmation) {
    return "guard_edge";
  }
  if (draftEdgeKind === "preview" || source.nodeType === "condition" || target.nodeType === "condition") {
    return "control_edge";
  }
  return "data_edge";
}

function validateWorkflowNodeDesignerConnection(
  connection: Connection | WorkflowNodeDesignerEdge,
  draft: WorkflowDraftDesignerDraft,
): boolean {
  if (!connection.source || !connection.target || connection.source === connection.target) {
    return false;
  }
  const source = draft.nodes.find((node) => node.nodeId === connection.source);
  const target = draft.nodes.find((node) => node.nodeId === connection.target);
  if (!source || !target) {
    return false;
  }
  return !draft.edges.some((edge) => edge.fromNodeId === source.nodeId && edge.toNodeId === target.nodeId);
}

function WorkflowNodeDesignerNodeCard({ data, selected }: NodeProps<WorkflowNodeDesignerNode>) {
  return (
    <div
      className={`workflow-node-designer-node ${selected ? "selected" : ""} ${data.readiness} ${
        data.validationFocus === "focused" ? `validation-focused ${data.validationSeverity ?? ""}` : ""
      }`}
    >
      <Handle id={`${data.draftNodeId}:input`} type="target" position={Position.Left} />
      <div className="workflow-node-designer-node-header">
        <span>{data.lane}</span>
        <strong>{data.label}</strong>
      </div>
      <p>{data.outputSummary}</p>
      <dl>
        <div>
          <dt>Type</dt>
          <dd>{data.nodeType}</dd>
        </div>
        <div>
          <dt>Risk</dt>
          <dd>{data.requiresConfirmation ? "requires confirmation" : data.riskLevel}</dd>
        </div>
        <div>
          <dt>Ref</dt>
          <dd>{data.providerRef || data.toolRef || data.ragRef || "draft local"}</dd>
        </div>
        <div>
          <dt>Guard</dt>
          <dd>{data.protectedNode ? "protected" : data.readiness}</dd>
        </div>
      </dl>
      <Handle id={`${data.draftNodeId}:output`} type="source" position={Position.Right} />
    </div>
  );
}

function WorkflowNodeDesignerEdgePath(props: EdgeProps<WorkflowNodeDesignerEdge>) {
  const [edgePath, labelX, labelY] = getSmoothStepPath({
    sourceX: props.sourceX,
    sourceY: props.sourceY,
    sourcePosition: props.sourcePosition,
    targetX: props.targetX,
    targetY: props.targetY,
    targetPosition: props.targetPosition,
    borderRadius: 18,
  });
  const edgeKind = props.data?.edgeKind ?? "data_edge";
  const validationFocusClass =
    props.data?.validationFocus === "focused" ? `validation-focused ${props.data.validationSeverity ?? ""}` : "";
  return (
    <>
      <BaseEdge
        id={props.id}
        path={edgePath}
        markerEnd={props.markerEnd}
        className={`workflow-node-designer-edge-path ${EDGE_KIND_CLASS[edgeKind]} ${validationFocusClass}`}
      />
      <text className="workflow-node-designer-edge-label" x={labelX} y={labelY} textAnchor="middle">
        {edgeKind}
      </text>
    </>
  );
}

function WorkflowNodeDesignerInspector({
  node,
  editingDisabled,
  canDelete,
  interactionFeedback,
  edges,
  onUpdateNodeLabel,
  onUpdateNodeInputSummary,
  onUpdateNodeOutputSummary,
  onUpdateNodeProviderRef,
  onUpdateNodeToolRef,
  onUpdateNodeRagRef,
  onUpdateNodeOutputMapping,
  onRemoveEdge,
  onRemoveNode,
}: {
  node: WorkflowDraftDesignerNode;
  editingDisabled: boolean;
  canDelete: boolean;
  interactionFeedback: WorkflowNodeDesignerFeedback;
  edges: WorkflowDraftDesignerEdge[];
  onUpdateNodeLabel: (nodeId: string, label: string) => void;
  onUpdateNodeInputSummary: (nodeId: string, inputSummary: string) => void;
  onUpdateNodeOutputSummary: (nodeId: string, outputSummary: string) => void;
  onUpdateNodeProviderRef: (nodeId: string, providerRef: string) => void;
  onUpdateNodeToolRef: (nodeId: string, toolRef: string) => void;
  onUpdateNodeRagRef: (nodeId: string, ragRef: string) => void;
  onUpdateNodeOutputMapping: (nodeId: string, outputMappingSummary: string) => void;
  onRemoveEdge: (edgeId: string) => boolean;
  onRemoveNode: (nodeId: string) => void;
}) {
  return (
    <>
      <div className="workflow-node-designer-inspector-heading">
        <span>{node.nodeType}</span>
        <strong>{node.nodeId}</strong>
        <p className={`workflow-node-designer-inspector-feedback ${interactionFeedback.tone}`}>
          {interactionFeedback.message}
        </p>
      </div>
      <label>
        <span>Label</span>
        <input
          type="text"
          value={node.label}
          maxLength={160}
          disabled={editingDisabled}
          onChange={(event) => onUpdateNodeLabel(node.nodeId, event.currentTarget.value)}
        />
      </label>
      <label>
        <span>Provider ref</span>
        <input
          type="text"
          value={node.providerRef}
          maxLength={240}
          disabled={editingDisabled}
          onChange={(event) => onUpdateNodeProviderRef(node.nodeId, event.currentTarget.value)}
        />
      </label>
      <label>
        <span>Tool ref</span>
        <input
          type="text"
          value={node.toolRef}
          maxLength={240}
          disabled={editingDisabled}
          onChange={(event) => onUpdateNodeToolRef(node.nodeId, event.currentTarget.value)}
        />
      </label>
      <label>
        <span>RAG ref</span>
        <input
          type="text"
          value={node.ragRef}
          maxLength={240}
          disabled={editingDisabled}
          onChange={(event) => onUpdateNodeRagRef(node.nodeId, event.currentTarget.value)}
        />
      </label>
      <label>
        <span>Input summary</span>
        <textarea
          value={node.inputSummary}
          maxLength={4000}
          rows={3}
          disabled={editingDisabled}
          onChange={(event) => onUpdateNodeInputSummary(node.nodeId, event.currentTarget.value)}
        />
      </label>
      <label>
        <span>Output summary</span>
        <textarea
          value={node.outputSummary}
          maxLength={4000}
          rows={3}
          disabled={editingDisabled}
          onChange={(event) => onUpdateNodeOutputSummary(node.nodeId, event.currentTarget.value)}
        />
      </label>
      <label>
        <span>Output mapping</span>
        <textarea
          value={node.outputMappingSummary}
          maxLength={4000}
          rows={3}
          disabled={editingDisabled}
          onChange={(event) => onUpdateNodeOutputMapping(node.nodeId, event.currentTarget.value)}
        />
      </label>
      <div className="workflow-node-designer-edge-actions" aria-label="Selected node draft edges">
        <span>Connected edges</span>
        {edges.length === 0 ? (
          <p>No draft edge is connected to this node.</p>
        ) : (
          edges.map((edge) => (
            <div key={edge.edgeId} className="workflow-node-designer-edge-action">
              <div className="workflow-node-designer-edge-action-main">
                <strong>
                  {edge.fromNodeId} to {edge.toNodeId}
                </strong>
                <small>
                  {edge.edgeKind} / {edge.edgeId}
                </small>
              </div>
              <button type="button" disabled={editingDisabled} onClick={() => onRemoveEdge(edge.edgeId)}>
                Remove edge
              </button>
            </div>
          ))
        )}
      </div>
      <button type="button" disabled={editingDisabled || !canDelete} onClick={() => onRemoveNode(node.nodeId)}>
        Remove node
      </button>
    </>
  );
}
