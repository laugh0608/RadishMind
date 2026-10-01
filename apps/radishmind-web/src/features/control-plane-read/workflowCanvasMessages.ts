import type { TFunction } from "i18next";

export type WorkflowCanvasMessage =
  | { code: "connectHint" }
  | { code: "connectLocked" }
  | { code: "connectInvalid" }
  | { code: "connectUnavailable" }
  | { code: "edgeAdded"; from: string; to: string; kind: string }
  | { code: "nodeSelected"; label: string; id: string }
  | { code: "findingFocused"; check: string; nodes: number; edges: number }
  | { code: "focusCleared" }
  | { code: "positionLocked" }
  | { code: "positionUpdated"; label: string }
  | { code: "removeEdgeLocked" }
  | { code: "edgeRemoved"; id: string }
  | { code: "edgeMissing"; id: string }
  | { code: "removeNodeLocked" }
  | { code: "nodeProtected"; id: string };

export function workflowCanvasMessage(t: TFunction<"workflow">, message: WorkflowCanvasMessage): string {
  switch (message.code) {
    case "connectHint": return t($ => $.canvas.connectHint, message);
    case "connectLocked": return t($ => $.canvas.connectLocked, message);
    case "connectInvalid": return t($ => $.canvas.connectInvalid, message);
    case "connectUnavailable": return t($ => $.canvas.connectUnavailable, message);
    case "edgeAdded": return t($ => $.canvas.edgeAdded, message);
    case "nodeSelected": return t($ => $.canvas.nodeSelected, message);
    case "findingFocused": return t($ => $.canvas.findingFocused, message);
    case "focusCleared": return t($ => $.canvas.focusCleared, message);
    case "positionLocked": return t($ => $.canvas.positionLocked, message);
    case "positionUpdated": return t($ => $.canvas.positionUpdated, message);
    case "removeEdgeLocked": return t($ => $.canvas.removeEdgeLocked, message);
    case "edgeRemoved": return t($ => $.canvas.edgeRemoved, message);
    case "edgeMissing": return t($ => $.canvas.edgeMissing, message);
    case "removeNodeLocked": return t($ => $.canvas.removeNodeLocked, message);
    case "nodeProtected": return t($ => $.canvas.nodeProtected, message);
  }
}

export function workflowValidationLabel(t: TFunction<"workflow">, checkId: string): string {
  switch (checkId) {
    case "executor_v0_profile": return t($ => $.canvas.checkExecutorV0Profile);
    case "executor_v0_node_roles": return t($ => $.canvas.checkExecutorV0NodeRoles);
    case "executor_v0_low_risk": return t($ => $.canvas.checkExecutorV0LowRisk);
    case "executor_v0_topology": return t($ => $.canvas.checkExecutorV0Topology);
    case "executor_v0_external_side_effects": return t($ => $.canvas.checkExecutorV0ExternalSideEffects);
    case "entry_context_lane": return t($ => $.canvas.checkEntryContextLane);
    case "model_reasoning_lane": return t($ => $.canvas.checkModelReasoningLane);
    case "policy_gate_path": return t($ => $.canvas.checkPolicyGatePath);
    case "output_audit_path": return t($ => $.canvas.checkOutputAuditPath);
    case "orphan_node_scan": return t($ => $.canvas.checkOrphanNodeScan);
    case "input_contract_fields": return t($ => $.canvas.checkInputContractFields);
    case "output_contract_fields": return t($ => $.canvas.checkOutputContractFields);
    default: return checkId;
  }
}
