import type { TFunction } from "i18next";
import type { WorkflowExecutorConsumerState } from "./workflowExecutorConsumer.ts";

export function executorStateMessage(t: TFunction<"workflow">, state: WorkflowExecutorConsumerState): string {
  switch (state.status) {
    case "disabled": return t($ => $.executor.stateDisabled);
    case "idle": return t($ => $.executor.stateIdle);
    case "starting": return t($ => $.executor.stateStarting);
    case "reading": return t($ => $.executor.stateReading);
    case "succeeded": return t($ => $.executor.stateSucceeded);
    case "failed": return t($ => $.executor.stateFailed, { code: state.failureCode ?? "workflow_run_unavailable" });
  }
}

export function executorBlockerMessage(t: TFunction<"workflow">, code: string): string {
  switch (code) {
    case "saved_draft_version_unavailable": return t($ => $.executor.saved_draft_version_unavailable);
    case "unsaved_local_changes": return t($ => $.executor.unsaved_local_changes);
    case "executor_profile_missing": return t($ => $.executor.executor_profile_missing);
    case "executor_graph_budget": return t($ => $.executor.executor_graph_budget);
    case "executor_node_id_invalid": return t($ => $.executor.executor_node_id_invalid);
    case "executor_node_type_blocked": return t($ => $.executor.executor_node_type_blocked);
    case "executor_node_risk_blocked": return t($ => $.executor.executor_node_risk_blocked);
    case "executor_node_roles_invalid": return t($ => $.executor.executor_node_roles_invalid);
    case "executor_edge_invalid": return t($ => $.executor.executor_edge_invalid);
    case "executor_condition_route_invalid": return t($ => $.executor.executor_condition_route_invalid);
    case "executor_condition_source_invalid": return t($ => $.executor.executor_condition_source_invalid);
    case "executor_root_terminal_invalid": return t($ => $.executor.executor_root_terminal_invalid);
    case "executor_cycle": return t($ => $.executor.executor_cycle);
    case "executor_unreachable_node": return t($ => $.executor.executor_unreachable_node);
    default: return t($ => $.executor.unknownBlocker, { code });
  }
}
