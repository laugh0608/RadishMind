import assert from "node:assert/strict";
import test from "node:test";
import { applyWorkflowDraftEdit } from "../src/features/control-plane-read/workflowDraftEditing.ts";
import { buildWorkflowExecutorV0Draft } from "../src/features/control-plane-read/workflowExecutorConsumer.ts";

function draft() { return buildWorkflowExecutorV0Draft(sourceDraft(), 1); }

test("field edits retain identity and source content while normalizing explicit contract fields", () => {
  const source = draft();
  const original = structuredClone(source);
  const nodeId = source.nodes[1].nodeId;
  const edited = applyWorkflowDraftEdit(source, { type: "input_fields", nodeId, text: "query, query\nanswer value\n" });
  assert.deepEqual(edited.nodes[1].inputContractFields, ["query", "answer_value"]);
  assert.equal(edited.draftId, source.draftId);
  assert.equal(edited.applicationRef, source.applicationRef);
  assert.equal(edited.localOnlyInteraction, "local_edit");
  assert.deepEqual(source, original);
});

test("adding policy and tool nodes preserves confirmation markers and unique node identities", () => {
  let current = draft();
  for (const nodeType of ["condition", "http_tool", "http_tool"] as const) current = applyWorkflowDraftEdit(current, { type: "add_node", nodeType });
  const added = current.nodes.filter((node) => ["condition", "http_tool"].includes(node.nodeType));
  assert.equal(added.length, 3);
  assert.equal(added.every((node) => node.requiresConfirmation && node.riskLevel === "medium"), true);
  assert.equal(new Set(current.nodes.map((node) => node.nodeId)).size, current.nodes.length);
  assert.equal(current.nodes.at(-1)?.nodeType, "output");
});

test("invalid connections and minimum-graph removals leave the editor unchanged", () => {
  const source = draft();
  assert.equal(applyWorkflowDraftEdit(source, { type: "add_edge", fromNodeId: source.nodes[0].nodeId, toNodeId: source.nodes[0].nodeId }), source);
  assert.equal(applyWorkflowDraftEdit(source, { type: "add_edge", fromNodeId: source.edges[0].fromNodeId, toNodeId: source.edges[0].toNodeId }), source);
  assert.equal(applyWorkflowDraftEdit(source, { type: "remove_node", nodeId: source.nodes[0].nodeId }), source);
  assert.equal(applyWorkflowDraftEdit(source, { type: "move_node", nodeId: source.nodes[0].nodeId, direction: "up" }), source);
  assert.equal(applyWorkflowDraftEdit(source, { type: "node", nodeId: "missing", patch: { label: "ignored" } }), source);
});

test("explicit edge edits preserve reviewable conditions and remove only the selected edge", () => {
  const source = draft();
  const edgeId = source.edges[0].edgeId;
  const edited = applyWorkflowDraftEdit(source, { type: "edge_condition", edgeId, conditionSummary: "  reviewed condition  " });
  assert.equal(edited.edges[0].conditionSummary, "reviewed condition");
  const empty = applyWorkflowDraftEdit(edited, { type: "edge_condition", edgeId, conditionSummary: "  " });
  assert.ok(empty.edges[0].conditionSummary.trim());
  const removed = applyWorkflowDraftEdit(edited, { type: "remove_edge", edgeId });
  assert.deepEqual(removed.edges, [source.edges[1]]);
  assert.equal(applyWorkflowDraftEdit(removed, { type: "remove_edge", edgeId }), removed);
});

test("node positions stay bounded and removed nodes do not leave layout or edge references", () => {
  let current = draft();
  current = applyWorkflowDraftEdit(current, { type: "add_node", nodeType: "llm" });
  current = applyWorkflowDraftEdit(current, { type: "add_node", nodeType: "llm" });
  const extra = current.nodes[2].nodeId;
  current = applyWorkflowDraftEdit(current, { type: "position", nodeId: extra, x: 50000, y: -12.8 });
  assert.deepEqual(current.designerLayout.nodePositions.find((position) => position.nodeId === extra), { nodeId: extra, x: 10000, y: -13 });
  const removed = applyWorkflowDraftEdit(current, { type: "remove_node", nodeId: extra });
  assert.equal(removed.nodes.some((node) => node.nodeId === extra), false);
  assert.equal(removed.edges.some((edge) => edge.fromNodeId === extra || edge.toNodeId === extra), false);
  assert.equal(removed.designerLayout.nodePositions.some((position) => position.nodeId === extra), false);
});

function sourceDraft() {
  return {
    draftId: "draft_source",
    templateRef: "wf_source",
    label: "RadishFlow advisory",
    applicationRef: "app_flow_copilot",
    workflowDefinitionId: "wf_radishflow_copilot_latest",
    providerProfileRef: "provider:mock",
    summary: "source",
    nodes: [],
    edges: [],
    designerLayout: {
      source: "workflow_node_designer" as const,
      persistence: "ui_only" as const,
      nodePositions: [],
    },
    readiness: [],
    risks: [],
    blockedCapabilities: [],
    routeMetadata: {
      sourceRouteId: "workflow-definition-summary-list-route" as const,
      draftRouteId: "workflow-draft-designer-offline-draft" as const,
      routePath: "/v1/user-workspace/workflow-definitions" as const,
      requestId: "req_source",
      auditRef: "audit_source",
    },
    localOnlyInteraction: "inspect_only" as const,
    executionProfile: "review_only" as const,
  };
}

