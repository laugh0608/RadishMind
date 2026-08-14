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
  type ReactFlowInstance,
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

type WorkflowNodeDesignerViewportMode = "focus" | "graph";

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
  retrieval: 216,
  model: 432,
  policy: 648,
  preview: 864,
  output: 1080,
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

function workflowNodeDesignerCompactFocusPosition(index: number): { x: number; y: number } {
  return {
    x: (index % 2) * 216,
    y: Math.floor(index / 2) * 180,
  };
}

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
  const [flowInstance, setFlowInstance] = useState<ReactFlowInstance<
    WorkflowNodeDesignerNode,
    WorkflowNodeDesignerEdge
  > | null>(null);
  const [viewportMode, setViewportMode] = useState<WorkflowNodeDesignerViewportMode>("focus");
  const [narrowViewport, setNarrowViewport] = useState(
    () => typeof window !== "undefined" && window.matchMedia("(max-width: 760px)").matches,
  );
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
  const topologyKey = `${draft.nodes.map((node) => node.nodeId).join("|")}::${draft.edges
    .map((edge) => `${edge.fromNodeId}>${edge.toNodeId}`)
    .join("|")}`;
  const canvasNodeIds = useMemo(() => draft.nodes.map((node) => node.nodeId), [topologyKey]);
  const neighborhoodNodeIds = useMemo(
    () => workflowNodeDesignerNeighborhoodNodeIds(draft, selectedNodeId),
    [selectedNodeId, topologyKey],
  );
  const validationFocusNodeIds = useMemo(
    () => workflowNodeDesignerValidationViewportNodeIds(draft, validationFocus),
    [topologyKey, validationFocus],
  );
  const focusedCanvasNodeIds = validationFocusNodeIds.length > 0 ? validationFocusNodeIds : neighborhoodNodeIds;
  const fitCanvasNodes = useCallback(
    (
      instance: ReactFlowInstance<WorkflowNodeDesignerNode, WorkflowNodeDesignerEdge>,
      targetNodeIds: string[],
      mode: WorkflowNodeDesignerViewportMode,
    ) => {
      window.requestAnimationFrame(() => {
        void instance.fitView({
          nodes: targetNodeIds.map((id) => ({ id })),
          padding: mode === "graph" ? 0.12 : narrowViewport ? 0.2 : 0.16,
          minZoom: 0.24,
          maxZoom: mode === "graph" ? 0.72 : narrowViewport ? 0.9 : 0.96,
          duration: 180,
        });
      });
    },
    [narrowViewport],
  );
  const reframeCanvasViewport = useCallback(
    (instance: ReactFlowInstance<WorkflowNodeDesignerNode, WorkflowNodeDesignerEdge>) => {
      const targetNodeIds = viewportMode === "graph" ? canvasNodeIds : focusedCanvasNodeIds;
      fitCanvasNodes(instance, targetNodeIds, viewportMode);
    },
    [canvasNodeIds, fitCanvasNodes, focusedCanvasNodeIds, viewportMode],
  );

  const onFlowInit = useCallback(
    (instance: ReactFlowInstance<WorkflowNodeDesignerNode, WorkflowNodeDesignerEdge>) => {
      setFlowInstance(instance);
      fitCanvasNodes(instance, focusedCanvasNodeIds, "focus");
    },
    [fitCanvasNodes, focusedCanvasNodeIds],
  );

  useEffect(() => {
    const narrowViewport = window.matchMedia("(max-width: 760px)");
    const handleViewportModeChange = () => setNarrowViewport(narrowViewport.matches);
    narrowViewport.addEventListener("change", handleViewportModeChange);
    return () => narrowViewport.removeEventListener("change", handleViewportModeChange);
  }, []);

  useEffect(() => {
    if (flowInstance) {
      reframeCanvasViewport(flowInstance);
    }
  }, [flowInstance, reframeCanvasViewport, topologyKey]);

  useEffect(() => {
    if (!flowInstance) {
      return;
    }
    let frameId = 0;
    const handleResize = () => {
      window.cancelAnimationFrame(frameId);
      frameId = window.requestAnimationFrame(() => reframeCanvasViewport(flowInstance));
    };
    window.addEventListener("resize", handleResize);
    return () => {
      window.cancelAnimationFrame(frameId);
      window.removeEventListener("resize", handleResize);
    };
  }, [flowInstance, reframeCanvasViewport]);

  const onNodesChange = useCallback(
    (changes: NodeChange<WorkflowNodeDesignerNode>[]) => {
      const allowedChanges = changes.filter((change) => {
        if (change.type === "remove") {
          return false;
        }
        if (!editingDisabled) {
          return true;
        }
        return change.type === "dimensions" || change.type === "select";
      });
      if (allowedChanges.length === 0) {
        return;
      }
      setNodes((currentNodes) => applyNodeChanges(allowedChanges, currentNodes));
    },
    [editingDisabled],
  );

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
      setViewportMode("focus");
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
    setViewportMode("focus");
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
    setViewportMode("focus");
    setValidationFocus(null);
    setInteractionFeedback({
      tone: "neutral",
      message: "Validation overlay focus cleared; canvas selection remains UI-only.",
    });
  }, []);

  const focusSelectedNode = useCallback(() => {
    setViewportMode("focus");
    setValidationFocus(null);
    if (flowInstance) {
      fitCanvasNodes(flowInstance, neighborhoodNodeIds, "focus");
    }
  }, [fitCanvasNodes, flowInstance, neighborhoodNodeIds]);

  const fitWholeGraph = useCallback(() => {
    setViewportMode("graph");
    if (flowInstance) {
      fitCanvasNodes(flowInstance, canvasNodeIds, "graph");
    }
  }, [canvasNodeIds, fitCanvasNodes, flowInstance]);

  const onNodeClick = useCallback((_: unknown, node: WorkflowNodeDesignerNode) => {
    selectNode(node.data.draftNodeId);
  }, [selectNode]);

  const onNodeDragStop = useCallback(
    (_: unknown, node: WorkflowNodeDesignerNode) => {
      if (editingDisabled) {
        setInteractionFeedback({
          tone: "bad",
          message: "Canvas position update rejected: local draft edits are locked.",
        });
        return;
      }
      onUpdateNodeDesignerPosition(node.data.draftNodeId, node.position.x, node.position.y);
      setInteractionFeedback({
        tone: "good",
        message: `Saved canvas position for ${node.data.label} as active draft layout metadata.`,
      });
    },
    [editingDisabled, onUpdateNodeDesignerPosition],
  );
  const onRemoveDraftEdge = useCallback(
    (edgeId: string) => {
      if (editingDisabled) {
        setInteractionFeedback({
          tone: "bad",
          message: "Edge removal rejected: local draft edits are locked.",
        });
        return false;
      }
      const removed = onRemoveEdge(edgeId);
      setInteractionFeedback({
        tone: removed ? "good" : "bad",
        message: removed
          ? `Removed draft edge: ${edgeId}. Validation inspector will recompute graph findings from the active draft.`
          : `Edge removal rejected: ${edgeId} is not in the active draft.`,
      });
      return removed;
    },
    [editingDisabled, onRemoveEdge],
  );
  const onRemoveDraftNode = useCallback(
    (nodeId: string) => {
      if (editingDisabled) {
        setInteractionFeedback({
          tone: "bad",
          message: "Node removal rejected: local draft edits are locked.",
        });
        return;
      }
      if (!canRemoveNode(nodeId)) {
        setInteractionFeedback({
          tone: "bad",
          message: `Node removal rejected: ${nodeId} is protected by the active draft structure.`,
        });
        return;
      }
      onRemoveNode(nodeId);
    },
    [canRemoveNode, editingDisabled, onRemoveNode],
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
        hidden: viewportMode === "focus" && !focusedCanvasNodeIds.includes(node.data.draftNodeId),
        position:
          narrowViewport && viewportMode === "focus" && focusedCanvasNodeIds.includes(node.data.draftNodeId)
            ? workflowNodeDesignerCompactFocusPosition(focusedCanvasNodeIds.indexOf(node.data.draftNodeId))
            : node.position,
        selected: node.data.draftNodeId === selectedNodeId,
        data: {
          ...node.data,
          validationFocus: workflowNodeDesignerValidationFocusState(
            validationFocus?.nodeIds.includes(node.data.draftNodeId) ?? false,
          ),
          validationSeverity: validationFocus?.severity,
        },
      })),
    [focusedCanvasNodeIds, narrowViewport, nodes, selectedNodeId, validationFocus, viewportMode],
  );
  const displayedEdges = useMemo(
    () =>
      edges.map((edge) => ({
        ...edge,
        hidden:
          viewportMode === "focus" &&
          (!focusedCanvasNodeIds.includes(edge.source) || !focusedCanvasNodeIds.includes(edge.target)),
        data: {
          ...(edge.data ?? { edgeKind: "data_edge", conditionSummary: "" }),
          validationFocus: workflowNodeDesignerValidationFocusState(validationFocus?.edgeIds.includes(edge.id) ?? false),
          validationSeverity: validationFocus?.severity,
        },
      })),
    [edges, focusedCanvasNodeIds, validationFocus, viewportMode],
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
    : narrowViewport && viewportMode === "focus"
      ? "Focused layout is transient; use Fit graph before changing saved positions, or edit the visible typed ports and Inspector."
      : "Drag nodes, connect typed ports, or edit inspector fields on the active draft.";
  const inspectorContent = selectedNode ? (
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
      onRemoveNode={onRemoveDraftNode}
    />
  ) : (
    <div className="workflow-node-designer-empty">
      <strong>Node unavailable</strong>
      <p>The list editor remains available and no sample fallback has been applied.</p>
    </div>
  );

  return (
    <section className="workflow-node-designer" aria-label="Workflow node designer canvas">
      <div className="workflow-node-designer-toolbar">
        <div>
          <p className="eyebrow">Node Designer Canvas</p>
          <h5>{draft.label}</h5>
        </div>
        <div className="workflow-node-designer-status">
          <span>{draft.localOnlyInteraction}</span>
          <strong>{draft.nodes.length} nodes / {draft.edges.length} draft edges</strong>
        </div>
      </div>

      <details className="workflow-node-designer-mapping-summary">
        <summary>
          <span>Persistence mapping</span>
          <strong>{mappedLayoutCount} positioned nodes</strong>
          <small>Attributes and endpoints save; viewport, selection, and edge kind do not</small>
        </summary>
        <div className="workflow-node-designer-mapping-details" aria-label="Workflow node designer saved draft mapping">
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
            <p>Visual edge kind labels are derived from node lane, risk, policy, and audit context.</p>
          </article>
        </div>
      </details>

      <div className="workflow-node-designer-interaction-bar" aria-label="Workflow node designer interaction state">
        <div
          className={`workflow-node-designer-feedback ${interactionFeedback.tone} ${
            editingDisabled ? "editing-locked" : "editing-enabled"
          }`}
          data-editing-state={editingDisabled ? "locked" : "enabled"}
          role="status"
          aria-live="polite"
        >
          <span>{editingStateLabel}</span>
          <strong>{interactionFeedback.message}</strong>
          <p>{editingStateSummary}</p>
        </div>
        <div className="workflow-node-designer-node-switcher" aria-label="Select workflow draft node">
          <span>Node navigator</span>
          <div className="workflow-node-designer-node-switcher-list">
            {draft.nodes.map((node) => (
              <button
                key={node.nodeId}
                type="button"
                className={node.nodeId === selectedNodeId ? "selected is-selected" : "is-not-selected"}
                data-node-status={node.readiness}
                data-selection-state={node.nodeId === selectedNodeId ? "selected" : "available"}
                aria-pressed={node.nodeId === selectedNodeId}
                aria-label={`${node.label}; ${node.nodeType}; status ${node.readiness}; ${
                  node.nodeId === selectedNodeId ? "currently selected" : "select node"
                }`}
                onClick={() => selectNode(node.nodeId)}
              >
                <strong>{node.label}</strong>
                <small>{node.nodeType}</small>
              </button>
            ))}
          </div>
          <label className="workflow-node-designer-mobile-node-select">
            <span>Inspect node</span>
            <select value={selectedNodeId} onChange={(event) => selectNode(event.currentTarget.value)}>
              {draft.nodes.map((node) => (
                <option key={node.nodeId} value={node.nodeId}>
                  {node.label} · {node.nodeType}
                </option>
              ))}
            </select>
          </label>
        </div>
      </div>

      <div className="workflow-node-designer-shell">
        <div className="workflow-node-designer-canvas-stage">
          <div className="workflow-node-designer-canvas-toolbar" aria-label="Workflow canvas viewport controls">
            <div>
              <span>Canvas view</span>
              <strong>
                {validationFocus
                  ? `${validationFocus.label} · ${validationFocusNodeIds.length} target nodes`
                  : viewportMode === "graph"
                    ? `Full graph · ${draft.nodes.length} nodes`
                    : `Focused neighborhood · ${focusedCanvasNodeIds.length} nodes`}
              </strong>
            </div>
            <div>
              <button type="button" aria-pressed={viewportMode === "focus"} onClick={focusSelectedNode}>
                Focus node
              </button>
              <button type="button" aria-pressed={viewportMode === "graph"} onClick={fitWholeGraph}>
                Fit graph
              </button>
            </div>
          </div>
          <div
            className={`workflow-node-designer-canvas ${editingDisabled ? "locked" : "editable"}`}
            data-editing-state={editingDisabled ? "locked" : "enabled"}
            aria-label={`Workflow node designer canvas; ${editingStateLabel.toLowerCase()}`}
          >
            <ReactFlow<WorkflowNodeDesignerNode, WorkflowNodeDesignerEdge>
              nodes={displayedNodes}
              edges={displayedEdges}
              nodeTypes={nodeTypes}
              edgeTypes={edgeTypes}
              nodesDraggable={!editingDisabled && (!narrowViewport || viewportMode === "graph")}
              nodesConnectable={!editingDisabled}
              deleteKeyCode={null}
              elementsSelectable
              minZoom={0.24}
              maxZoom={1.4}
              onInit={onFlowInit}
              onNodesChange={onNodesChange}
              onNodeClick={onNodeClick}
              onNodeDragStop={onNodeDragStop}
              onConnect={onConnect}
              isValidConnection={isValidConnection}
            >
              <Background gap={24} size={1} />
              {viewportMode === "graph" ? <MiniMap pannable zoomable nodeStrokeWidth={3} /> : null}
              <Controls showFitView={false} showInteractive={false} />
            </ReactFlow>
          </div>
        </div>

        {narrowViewport ? (
          <details
            className="workflow-node-designer-inspector is-disclosure"
            aria-label="Workflow node designer inspector"
          >
            <summary>
              <span>Inspector</span>
              <strong>{selectedNode?.label ?? "Node unavailable"}</strong>
            </summary>
            <div className="workflow-node-designer-inspector-content">
              {inspectorContent}
            </div>
          </details>
        ) : (
          <aside className="workflow-node-designer-inspector" aria-label="Workflow node designer inspector">
            {inspectorContent}
          </aside>
        )}
      </div>

      <details
        className="workflow-node-designer-validation-navigation"
        aria-label="Workflow node designer validation overlay navigation"
      >
        <summary className="workflow-node-designer-validation-navigation-heading">
          <div>
            <span>Validation overlay</span>
            <strong>
              {validationInspector.validationStatus} / {validationNavigationItems.length} findings
            </strong>
          </div>
          <small>Expand findings</small>
        </summary>
        <div className="workflow-node-designer-validation-navigation-body">
          <div className="workflow-node-designer-validation-navigation-actions">
            <span>Finding focus is independent from node selection.</span>
            <button type="button" disabled={!validationFocus} onClick={clearValidationFocus}>
              Clear focus
            </button>
          </div>
          <div className="workflow-node-designer-validation-navigation-list">
            {validationNavigationItems.map((item) => {
              const isFocused = validationFocus?.checkId === item.checkId;
              return (
                <button
                  key={item.checkId}
                  type="button"
                  className={`${item.status} status-${item.status} ${item.severity} severity-${item.severity} ${
                    isFocused ? "focused is-validation-focused" : "is-not-validation-focused"
                  }`}
                  data-validation-status={item.status}
                  data-validation-severity={item.severity}
                  data-validation-focus={isFocused ? "focused" : "none"}
                  aria-pressed={isFocused}
                  aria-label={`${item.label}; status ${item.status}; severity ${item.severity}; ${item.targetSummary}; ${
                    isFocused ? "validation focus active" : "focus validation finding"
                  }`}
                  onClick={() => focusValidationFinding(item)}
                >
                  <span>{item.status}</span>
                  <strong>{item.label}</strong>
                  <small>{item.targetSummary}</small>
                </button>
              );
            })}
          </div>
        </div>
      </details>
    </section>
  );
}

function workflowNodeDesignerNeighborhoodNodeIds(
  draft: WorkflowDraftDesignerDraft,
  selectedNodeId: string,
): string[] {
  const availableNodeIds = new Set(draft.nodes.map((node) => node.nodeId));
  const selectedNode = draft.nodes.find((node) => node.nodeId === selectedNodeId) ?? draft.nodes[0];
  if (!selectedNode) {
    return [];
  }
  const adjacentNodeIds = new Set<string>();
  for (const edge of draft.edges) {
    if (edge.fromNodeId === selectedNodeId && availableNodeIds.has(edge.toNodeId)) {
      adjacentNodeIds.add(edge.toNodeId);
    }
    if (edge.toNodeId === selectedNodeId && availableNodeIds.has(edge.fromNodeId)) {
      adjacentNodeIds.add(edge.fromNodeId);
    }
  }
  const nearestAdjacentNode = draft.nodes
    .filter((node) => adjacentNodeIds.has(node.nodeId))
    .sort((left, right) => {
      const leftDistance = Math.abs(LANE_X[left.lane] - LANE_X[selectedNode.lane]);
      const rightDistance = Math.abs(LANE_X[right.lane] - LANE_X[selectedNode.lane]);
      return leftDistance - rightDistance;
    })[0];
  return nearestAdjacentNode ? [selectedNode.nodeId, nearestAdjacentNode.nodeId] : [selectedNode.nodeId];
}

function workflowNodeDesignerValidationViewportNodeIds(
  draft: WorkflowDraftDesignerDraft,
  validationFocus: WorkflowNodeDesignerValidationFocus | null,
): string[] {
  if (!validationFocus) {
    return [];
  }
  const targetNodeIds = new Set(validationFocus.nodeIds);
  const targetEdgeIds = new Set(validationFocus.edgeIds);
  for (const edge of draft.edges) {
    if (targetEdgeIds.has(edge.edgeId)) {
      targetNodeIds.add(edge.fromNodeId);
      targetNodeIds.add(edge.toNodeId);
    }
  }
  return draft.nodes.map((node) => node.nodeId).filter((nodeId) => targetNodeIds.has(nodeId));
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
      deletable: false,
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
      deletable: false,
      markerEnd: {
        type: MarkerType.ArrowClosed,
        color: "var(--rm-border-strong)",
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
  const validationFocused = data.validationFocus === "focused";
  const validationSeverity = data.validationSeverity ?? "none";
  return (
    <div
      className={`workflow-node-designer-node ${data.readiness} status-${data.readiness} ${
        selected ? "selected is-selected" : "is-not-selected"
      } ${
        validationFocused
          ? `validation-focused is-validation-focused ${data.validationSeverity ?? ""} severity-${validationSeverity}`
          : "is-not-validation-focused"
      }`}
      data-node-status={data.readiness}
      data-selection-state={selected ? "selected" : "available"}
      data-validation-focus={validationFocused ? "focused" : "none"}
      data-validation-severity={validationSeverity}
      role="group"
      aria-label={`${data.label}; node status ${data.readiness}; ${selected ? "selected" : "not selected"}; ${
        validationFocused ? `validation focus ${validationSeverity}` : "no validation focus"
      }`}
    >
      <Handle id={`${data.draftNodeId}:input`} type="target" position={Position.Left} />
      <div className="workflow-node-designer-node-header">
        <div>
          <span>{data.lane} · {data.nodeType}</span>
          <small>{data.readiness}</small>
        </div>
        <strong>{data.label}</strong>
      </div>
      <p>{data.outputSummary}</p>
      <dl>
        <div>
          <dt>Ref</dt>
          <dd>{data.providerRef || data.toolRef || data.ragRef || "draft local"}</dd>
        </div>
        <div>
          <dt>Guard</dt>
          <dd>{data.protectedNode ? "protected" : data.requiresConfirmation ? "confirmation" : data.riskLevel}</dd>
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
  const validationFocused = props.data?.validationFocus === "focused";
  const validationSeverity = props.data?.validationSeverity ?? "none";
  const validationFocusClass = validationFocused
    ? `validation-focused is-validation-focused ${props.data?.validationSeverity ?? ""} severity-${validationSeverity}`
    : "is-not-validation-focused";
  return (
    <>
      <BaseEdge
        id={props.id}
        path={edgePath}
        markerEnd={props.markerEnd}
        className={`workflow-node-designer-edge-path ${EDGE_KIND_CLASS[edgeKind]} ${validationFocusClass}`}
        data-edge-kind={edgeKind}
        data-validation-focus={validationFocused ? "focused" : "none"}
        data-validation-severity={validationSeverity}
        aria-label={`Workflow edge ${props.source} to ${props.target}; kind ${edgeKind}; ${
          validationFocused ? `validation focus ${validationSeverity}` : "no validation focus"
        }`}
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
      <dl className="workflow-node-designer-inspector-meta">
        <div>
          <dt>Reference</dt>
          <dd>{node.providerRef || node.toolRef || node.ragRef || "draft local"}</dd>
        </div>
        <div>
          <dt>Readiness</dt>
          <dd>{node.readiness}</dd>
        </div>
      </dl>
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
      <details className="workflow-node-designer-inspector-details">
        <summary>
          <span>Node attributes and contract</span>
          <small>References, summaries, and mapping</small>
        </summary>
        <div className="workflow-node-designer-inspector-details-body">
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
        </div>
      </details>
      <details className="workflow-node-designer-inspector-details">
        <summary>
          <span>Connected edges</span>
          <small>{edges.length} draft edges</small>
        </summary>
        <div className="workflow-node-designer-edge-actions" aria-label="Selected node draft edges">
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
      </details>
      <button type="button" disabled={editingDisabled || !canDelete} onClick={() => onRemoveNode(node.nodeId)}>
        Remove node
      </button>
    </>
  );
}
