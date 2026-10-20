import type { TFunction } from "i18next";
import type { WorkflowRAGExecutionReason, WorkflowRAGExecutionState } from "./workflowRAGExecutionConsumer.ts";

export function workflowRAGExecutionStatus(t: TFunction<"workflow">, status: string): string {
  switch (status) {
    case "offline": return t($ => $.ragExecution.status.offline);
    case "idle": return t($ => $.ragExecution.status.idle);
    case "executing": return t($ => $.ragExecution.status.executing);
    case "succeeded": return t($ => $.ragExecution.status.succeeded);
    case "failed": return t($ => $.ragExecution.status.failed);
    case "scope_denied": return t($ => $.ragExecution.status.scope_denied);
    case "loading": return t($ => $.ragExecution.status.loading);
    case "reading": return t($ => $.ragExecution.status.reading);
    case "ready": return t($ => $.ragExecution.status.ready);
    case "empty": return t($ => $.ragExecution.status.empty);
    case "loaded": return t($ => $.ragExecution.status.loaded);
    default: return t($ => $.ragExecution.status.unknown);
  }
}

export function workflowRAGExecutionFeedback(t: TFunction<"workflow">, message: WorkflowRAGExecutionState["message"]): string {
  switch (message) {
    case "offline": return t($ => $.ragExecution.feedback.offline);
    case "ready": return t($ => $.ragExecution.feedback.ready);
    case "executing": return t($ => $.ragExecution.feedback.executing);
    case "succeeded": return t($ => $.ragExecution.feedback.succeeded);
    case "localRejected": return t($ => $.ragExecution.feedback.localRejected);
    case "failed": return t($ => $.ragExecution.feedback.failed);
    case "unavailable": return t($ => $.ragExecution.feedback.unavailable);
  }
}

export function workflowRAGExecutionFailure(t: TFunction<"workflow">, code: string): string {
  switch (code) {
    case "workflow_rag_snapshot_not_found": return t($ => $.ragExecution.failure.workflow_rag_snapshot_not_found);
    case "workflow_rag_snapshot_scope_denied": return t($ => $.ragExecution.failure.workflow_rag_snapshot_scope_denied);
    case "workflow_rag_snapshot_archived": return t($ => $.ragExecution.failure.workflow_rag_snapshot_archived);
    case "workflow_rag_snapshot_version_conflict": return t($ => $.ragExecution.failure.workflow_rag_snapshot_version_conflict);
    case "workflow_rag_profile_disabled": return t($ => $.ragExecution.failure.workflow_rag_profile_disabled);
    case "workflow_rag_draft_ineligible": return t($ => $.ragExecution.failure.workflow_rag_draft_ineligible);
    case "workflow_rag_query_invalid": return t($ => $.ragExecution.failure.workflow_rag_query_invalid);
    case "workflow_rag_budget_exceeded": return t($ => $.ragExecution.failure.workflow_rag_budget_exceeded);
    case "workflow_rag_no_evidence": return t($ => $.ragExecution.failure.workflow_rag_no_evidence);
    case "workflow_rag_retrieval_failed": return t($ => $.ragExecution.failure.workflow_rag_retrieval_failed);
    case "workflow_rag_gateway_failed": return t($ => $.ragExecution.failure.workflow_rag_gateway_failed);
    case "workflow_rag_answer_invalid": return t($ => $.ragExecution.failure.workflow_rag_answer_invalid);
    case "workflow_rag_citation_invalid": return t($ => $.ragExecution.failure.workflow_rag_citation_invalid);
    case "workflow_rag_execution_interrupted": return t($ => $.ragExecution.failure.workflow_rag_execution_interrupted);
    case "workflow_rag_canceled": return t($ => $.ragExecution.failure.workflow_rag_canceled);
    case "workflow_rag_store_unavailable": return t($ => $.ragExecution.failure.workflow_rag_store_unavailable);
    case "workflow_rag_execution_http_disabled": return t($ => $.ragExecution.failure.workflow_rag_execution_http_disabled);
    default: return t($ => $.ragExecution.failure.unknown);
  }
}

export function workflowRAGConfidenceLabel(t: TFunction<"workflow">, level: "low" | "medium" | "high"): string {
  switch (level) {
    case "low": return t($ => $.ragExecution.confidenceLevels.low);
    case "medium": return t($ => $.ragExecution.confidenceLevels.medium);
    case "high": return t($ => $.ragExecution.confidenceLevels.high);
  }
}

export function workflowRAGExecutionReasonMessage(t: TFunction<"workflow">, reason: WorkflowRAGExecutionReason): string {
  switch (reason.code) {
    case "rag_execution_offline": return t($ => $.ragExecution.reasons.rag_execution_offline);
    case "rag_execution_scope_denied": return t($ => $.ragExecution.reasons.rag_execution_scope_denied, { scope: reason.scope });
    case "rag_execution_profile_required": return t($ => $.ragExecution.reasons.rag_execution_profile_required);
    case "unsaved_local_changes": return t($ => $.ragExecution.reasons.unsaved_local_changes);
    case "saved_draft_version_unavailable": return t($ => $.ragExecution.reasons.saved_draft_version_unavailable);
    case "blocked_capabilities_present": return t($ => $.ragExecution.reasons.blocked_capabilities_present);
    case "rag_execution_topology_invalid": return t($ => $.ragExecution.reasons.rag_execution_topology_invalid);
    case "rag_ref_invalid": return t($ => $.ragExecution.reasons.rag_ref_invalid);
    case "rag_execution_node_boundary_invalid": return t($ => $.ragExecution.reasons.rag_execution_node_boundary_invalid);
    case "rag_execution_edges_invalid": return t($ => $.ragExecution.reasons.rag_execution_edges_invalid);
  }
}
