import type { TFunction } from "i18next";
import type { WorkflowHTTPToolActionConsumerState, WorkflowHTTPToolPublicArgumentsValidation } from "./workflowHTTPToolActionConsumer.ts";
import type { WorkflowHTTPToolExecutionState } from "./workflowHTTPToolExecutionConsumer.ts";

export function workflowHTTPToolStatus(t: TFunction<"workflow">, code: string): string {
  switch (code) {
    case "disabled": return t($ => $.httpTool.status.disabled);
    case "idle": return t($ => $.httpTool.status.idle);
    case "creating": return t($ => $.httpTool.status.creating);
    case "reading": return t($ => $.httpTool.status.reading);
    case "deciding": return t($ => $.httpTool.status.deciding);
    case "ready": return t($ => $.httpTool.status.ready);
    case "conflict_refreshed": return t($ => $.httpTool.status.conflict_refreshed);
    case "failed": return t($ => $.httpTool.status.failed);
    case "restoring": return t($ => $.httpTool.status.restoring);
    case "executing": return t($ => $.httpTool.status.executing);
    case "succeeded": return t($ => $.httpTool.status.succeeded);
    case "outcome_unknown": return t($ => $.httpTool.status.outcome_unknown);
    case "pending": return t($ => $.httpTool.status.pending);
    case "deferred": return t($ => $.httpTool.status.deferred);
    case "approved": return t($ => $.httpTool.status.approved);
    case "rejected": return t($ => $.httpTool.status.rejected);
    case "canceled": return t($ => $.httpTool.status.canceled);
    case "expired": return t($ => $.httpTool.status.expired);
    case "invalidated": return t($ => $.httpTool.status.invalidated);
    case "consumed": return t($ => $.httpTool.status.consumed);
    case "approve": return t($ => $.httpTool.status.approve);
    case "reject": return t($ => $.httpTool.status.reject);
    case "defer": return t($ => $.httpTool.status.defer);
    case "cancel": return t($ => $.httpTool.status.cancel);
    case "expire": return t($ => $.httpTool.status.expire);
    case "invalidate": return t($ => $.httpTool.status.invalidate);
    default: return t($ => $.httpTool.status.unknown);
  }
}

export function workflowHTTPToolFailure(t: TFunction<"workflow">, code: string): string {
  switch (code) {
    case "workflow_tool_saved_draft_required": return t($ => $.httpTool.failureText.workflow_tool_saved_draft_required);
    case "workflow_tool_unsaved_local_changes": return t($ => $.httpTool.failureText.workflow_tool_unsaved_local_changes);
    case "workflow_tool_node_count_invalid": return t($ => $.httpTool.failureText.workflow_tool_node_count_invalid);
    case "workflow_tool_exact_version_required": return t($ => $.httpTool.failureText.workflow_tool_exact_version_required);
    case "workflow_tool_confirmation_boundary_invalid": return t($ => $.httpTool.failureText.workflow_tool_confirmation_boundary_invalid);
    case "workflow_tool_graph_ineligible": return t($ => $.httpTool.failureText.workflow_tool_graph_ineligible);
    case "workflow_definition_http_tool_authority_invalid": return t($ => $.httpTool.failureText.workflow_definition_http_tool_authority_invalid);
    case "workflow_tool_action_scope_denied": return t($ => $.httpTool.failureText.workflow_tool_action_scope_denied);
    case "workflow_run_scope_denied": return t($ => $.httpTool.failureText.workflow_run_scope_denied);
    case "workflow_run_input_invalid": return t($ => $.httpTool.failureText.workflow_run_input_invalid);
    case "workflow_tool_not_registered": return t($ => $.httpTool.failureText.workflow_tool_not_registered);
    case "workflow_tool_profile_disabled": return t($ => $.httpTool.failureText.workflow_tool_profile_disabled);
    case "workflow_tool_target_denied": return t($ => $.httpTool.failureText.workflow_tool_target_denied);
    case "workflow_tool_arguments_invalid": return t($ => $.httpTool.failureText.workflow_tool_arguments_invalid);
    case "workflow_tool_confirmation_required": return t($ => $.httpTool.failureText.workflow_tool_confirmation_required);
    case "workflow_tool_confirmation_rejected": return t($ => $.httpTool.failureText.workflow_tool_confirmation_rejected);
    case "workflow_tool_confirmation_expired": return t($ => $.httpTool.failureText.workflow_tool_confirmation_expired);
    case "workflow_tool_confirmation_stale": return t($ => $.httpTool.failureText.workflow_tool_confirmation_stale);
    case "workflow_tool_confirmation_mismatch": return t($ => $.httpTool.failureText.workflow_tool_confirmation_mismatch);
    case "workflow_tool_confirmation_invalidated": return t($ => $.httpTool.failureText.workflow_tool_confirmation_invalidated);
    case "workflow_tool_action_canceled": return t($ => $.httpTool.failureText.workflow_tool_action_canceled);
    case "workflow_tool_action_consumed": return t($ => $.httpTool.failureText.workflow_tool_action_consumed);
    case "workflow_tool_transport_failed": return t($ => $.httpTool.failureText.workflow_tool_transport_failed);
    case "workflow_tool_timeout": return t($ => $.httpTool.failureText.workflow_tool_timeout);
    case "workflow_tool_response_status_invalid": return t($ => $.httpTool.failureText.workflow_tool_response_status_invalid);
    case "workflow_tool_response_too_large": return t($ => $.httpTool.failureText.workflow_tool_response_too_large);
    case "workflow_tool_response_invalid": return t($ => $.httpTool.failureText.workflow_tool_response_invalid);
    case "workflow_tool_outcome_unknown": return t($ => $.httpTool.failureText.workflow_tool_outcome_unknown);
    case "workflow_tool_store_unavailable": return t($ => $.httpTool.failureText.workflow_tool_store_unavailable);
    case "workflow_tool_store_contract_mismatch": return t($ => $.httpTool.failureText.workflow_tool_store_contract_mismatch);
    case "workflow_tool_action_draft_not_found": return t($ => $.httpTool.failureText.workflow_tool_action_draft_not_found);
    case "workflow_tool_action_draft_ineligible": return t($ => $.httpTool.failureText.workflow_tool_action_draft_ineligible);
    case "workflow_tool_action_definition_not_found": return t($ => $.httpTool.failureText.workflow_tool_action_definition_not_found);
    case "workflow_tool_action_definition_inactive": return t($ => $.httpTool.failureText.workflow_tool_action_definition_inactive);
    case "workflow_tool_action_definition_drift": return t($ => $.httpTool.failureText.workflow_tool_action_definition_drift);
    case "workflow_tool_action_definition_ineligible": return t($ => $.httpTool.failureText.workflow_tool_action_definition_ineligible);
    case "workflow_tool_action_plan_not_found": return t($ => $.httpTool.failureText.workflow_tool_action_plan_not_found);
    default: return t($ => $.httpTool.failureText.unknown);
  }
}

export function workflowHTTPToolActionFeedback(t: TFunction<"workflow">, state: WorkflowHTTPToolActionConsumerState): string {
  switch (state.status) {
    case "disabled": return t($ => $.httpTool.actionDisabled);
    case "idle": return t($ => $.httpTool.actionIdle);
    case "creating": return t($ => $.httpTool.actionCreating);
    case "reading": return t($ => $.httpTool.actionReading);
    case "deciding": return t($ => $.httpTool.actionDeciding);
    case "conflict_refreshed": return t($ => $.httpTool.actionConflict);
    case "failed": return workflowHTTPToolFailure(t, state.failureCode);
    case "ready": return t($ => $.httpTool.actionReady, { status: workflowHTTPToolStatus(t, state.actionPlan?.status ?? "unknown") });
  }
}

export function workflowHTTPToolExecutionFeedback(t: TFunction<"workflow">, state: WorkflowHTTPToolExecutionState): string {
  switch (state.status) {
    case "disabled": return t($ => $.httpTool.actionDisabled);
    case "idle": return t($ => $.httpTool.executionIdle);
    case "restoring": return t($ => $.httpTool.executionRestoring);
    case "executing": return t($ => $.httpTool.executionRunning);
    case "succeeded": return t($ => $.httpTool.executionSuccess);
    case "outcome_unknown": return t($ => $.httpTool.unknownOutcome);
    case "failed": return workflowHTTPToolFailure(t, state.failureCode);
  }
}

export function workflowHTTPToolArgumentsFeedback(t: TFunction<"workflow">, validation: WorkflowHTTPToolPublicArgumentsValidation): string {
  switch (validation.reason) {
    case "resource": return t($ => $.httpTool.argumentResource);
    case "locale": return t($ => $.httpTool.argumentLocale);
    case "sensitive": return t($ => $.httpTool.argumentSensitive);
    case "": return "";
  }
}
