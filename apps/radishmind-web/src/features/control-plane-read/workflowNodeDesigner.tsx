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
};

type WorkflowNodeDesignerEdgeData = {
  edgeKind: WorkflowNodeDesignerEdgeKind;
  conditionSummary: string;
};

type WorkflowNodeDesignerNode = Node<WorkflowNodeDesignerNodeData, "workflowDraftNode">;
type WorkflowNodeDesignerEdge = Edge<WorkflowNodeDesignerEdgeData, "workflowDraftEdge">;

type WorkflowNodeDesignerProps = {
  draft: WorkflowDraftDesignerDraft;
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
  const [nodes, setNodes] = useState(initialNodes);
  const [selectedNodeId, setSelectedNodeId] = useState(draft.nodes[0]?.nodeId ?? "");
  const [connectionFeedback, setConnectionFeedback] = useState("Connect typed ports to add controlled draft edges.");

  useEffect(() => {
    setNodes(initialNodes);
    setSelectedNodeId((current) =>
      draft.nodes.some((node) => node.nodeId === current) ? current : draft.nodes[0]?.nodeId ?? "",
    );
  }, [draft, initialNodes]);

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
        setConnectionFeedback("Connection rejected: local draft edits are locked while a saved draft operation is pending.");
        return;
      }
      const valid = validateWorkflowNodeDesignerConnection(connection, draft);
      if (!valid || !connection.source || !connection.target) {
        setConnectionFeedback("Connection rejected: typed ports require distinct source and target draft nodes.");
        return;
      }
      const source = draft.nodes.find((node) => node.nodeId === connection.source);
      const target = draft.nodes.find((node) => node.nodeId === connection.target);
      const edgeKind = source && target ? deriveWorkflowNodeDesignerEdgeKind(source, target) : "data_edge";
      const added = onAddEdge(connection.source, connection.target);
      if (!added) {
        setConnectionFeedback("Connection rejected: active draft endpoints are unavailable or already connected.");
        return;
      }
      setConnectionFeedback(
        `Added draft edge: ${connection.source} to ${connection.target} is tracked as ${edgeKind}.`,
      );
    },
    [draft, editingDisabled, onAddEdge],
  );

  const onNodeClick = useCallback((_: unknown, node: WorkflowNodeDesignerNode) => {
    setSelectedNodeId(node.data.draftNodeId);
  }, []);

  const onNodeDragStop = useCallback(
    (_: unknown, node: WorkflowNodeDesignerNode) => {
      onUpdateNodeDesignerPosition(node.data.draftNodeId, node.position.x, node.position.y);
    },
    [onUpdateNodeDesignerPosition],
  );
  const onRemoveDraftEdge = useCallback(
    (edgeId: string) => {
      const removed = onRemoveEdge(edgeId);
      setConnectionFeedback(
        removed
          ? `Removed draft edge: ${edgeId}. Validation inspector will recompute graph findings from the active draft.`
          : `Edge removal rejected: ${edgeId} is not in the active draft.`,
      );
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
  const layoutPersistenceLabel =
    draft.designerLayout.persistence === "saved_draft_metadata" ? "restored saved draft layout" : "active draft layout";
  const layoutPersistenceSummary =
    draft.designerLayout.persistence === "saved_draft_metadata"
      ? "Node positions were restored from saved draft layout metadata; viewport and selection remain transient."
      : "Save Draft writes sanitized node positions as saved draft layout metadata; viewport and selection remain transient.";

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

      <div className="workflow-node-designer-shell">
        <div className="workflow-node-designer-canvas">
          <ReactFlow
            nodes={nodes}
            edges={edges}
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
              connectionFeedback={connectionFeedback}
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
    <div className={`workflow-node-designer-node ${selected ? "selected" : ""} ${data.readiness}`}>
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
  return (
    <>
      <BaseEdge
        id={props.id}
        path={edgePath}
        markerEnd={props.markerEnd}
        className={`workflow-node-designer-edge-path ${EDGE_KIND_CLASS[edgeKind]}`}
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
  connectionFeedback,
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
  connectionFeedback: string;
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
        <p>{connectionFeedback}</p>
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
