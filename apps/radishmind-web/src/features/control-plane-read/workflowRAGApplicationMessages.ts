import type { TFunction } from "i18next";
import type { WorkflowRAGApplicationInvocationResult, WorkflowRAGApplicationRuntimeResult } from "./workflowRAGApplicationRuntimeConsumer.ts";

export function workflowRAGApplicationStatus(t: TFunction<"workflow">, code: string): string {
  switch (code) {
    case "idle": return t($ => $.ragApplication.status.idle);
    case "offline": return t($ => $.ragApplication.status.offline);
    case "ready": return t($ => $.ragApplication.status.ready);
    case "active": return t($ => $.ragApplication.status.active);
    case "revoked": return t($ => $.ragApplication.status.revoked);
    case "not_found": return t($ => $.ragApplication.status.not_found);
    case "version_conflict": return t($ => $.ragApplication.status.version_conflict);
    case "failed": return t($ => $.ragApplication.status.failed);
    case "succeeded": return t($ => $.ragApplication.status.succeeded);
    case "loading": return t($ => $.ragApplication.status.loading);
    case "recording": return t($ => $.ragApplication.status.recording);
    case "invoking": return t($ => $.ragApplication.status.invoking);
    default: return t($ => $.ragApplication.status.unknown);
  }
}

export function workflowRAGApplicationFailure(t: TFunction<"workflow">, code: string): string {
  switch (code) {
    case "workflow_rag_runtime_assignment_not_found": return t($ => $.ragApplication.failure.workflow_rag_runtime_assignment_not_found);
    case "workflow_rag_runtime_assignment_revoked": return t($ => $.ragApplication.failure.workflow_rag_runtime_assignment_revoked);
    case "workflow_rag_runtime_assignment_version_conflict": return t($ => $.ragApplication.failure.workflow_rag_runtime_assignment_version_conflict);
    case "workflow_rag_runtime_candidate_not_approved": return t($ => $.ragApplication.failure.workflow_rag_runtime_candidate_not_approved);
    case "workflow_rag_runtime_candidate_superseded": return t($ => $.ragApplication.failure.workflow_rag_runtime_candidate_superseded);
    case "workflow_rag_runtime_configuration_changed": return t($ => $.ragApplication.failure.workflow_rag_runtime_configuration_changed);
    case "workflow_rag_runtime_configuration_invalid": return t($ => $.ragApplication.failure.workflow_rag_runtime_configuration_invalid);
    case "workflow_rag_runtime_binding_not_eligible": return t($ => $.ragApplication.failure.workflow_rag_runtime_binding_not_eligible);
    case "workflow_rag_runtime_application_archived": return t($ => $.ragApplication.failure.workflow_rag_runtime_application_archived);
    case "workflow_rag_runtime_scope_denied": return t($ => $.ragApplication.failure.workflow_rag_runtime_scope_denied);
    case "workflow_rag_runtime_payload_invalid": return t($ => $.ragApplication.failure.workflow_rag_runtime_payload_invalid);
    case "workflow_rag_runtime_secret_material_forbidden": return t($ => $.ragApplication.failure.workflow_rag_runtime_secret_material_forbidden);
    case "workflow_rag_runtime_no_evidence": return t($ => $.ragApplication.failure.workflow_rag_runtime_no_evidence);
    case "workflow_rag_runtime_budget_exceeded": return t($ => $.ragApplication.failure.workflow_rag_runtime_budget_exceeded);
    case "workflow_rag_runtime_gateway_failed": return t($ => $.ragApplication.failure.workflow_rag_runtime_gateway_failed);
    case "workflow_rag_runtime_answer_invalid": return t($ => $.ragApplication.failure.workflow_rag_runtime_answer_invalid);
    case "workflow_rag_runtime_citation_invalid": return t($ => $.ragApplication.failure.workflow_rag_runtime_citation_invalid);
    case "workflow_rag_runtime_store_unavailable": return t($ => $.ragApplication.failure.workflow_rag_runtime_store_unavailable);
    case "workflow_rag_runtime_store_contract_mismatch": return t($ => $.ragApplication.failure.workflow_rag_runtime_store_contract_mismatch);
    case "workflow_rag_runtime_write_disabled": return t($ => $.ragApplication.failure.workflow_rag_runtime_write_disabled);
    case "workflow_rag_runtime_transition_invalid": return t($ => $.ragApplication.failure.workflow_rag_runtime_transition_invalid);
    case "workflow_rag_application_runtime_http_disabled": return t($ => $.ragApplication.failure.workflow_rag_application_runtime_http_disabled);
    default: return t($ => $.ragApplication.failure.unknown);
  }
}

export function workflowRAGAssignmentFeedback(t: TFunction<"workflow">, result: WorkflowRAGApplicationRuntimeResult): string {
  if (result.failureCode === "workflow_rag_runtime_assignment_version_conflict") return t($ => $.ragApplication.assignment.conflict);
  if (result.failureCode) return workflowRAGApplicationFailure(t, result.failureCode);
  if (!result.assignment) return t($ => $.ragApplication.assignment.idle);
  return result.assignment.state === "active" ? t($ => $.ragApplication.assignment.active) : t($ => $.ragApplication.assignment.revoked);
}

export function workflowRAGInvocationFeedback(t: TFunction<"workflow">, result: WorkflowRAGApplicationInvocationResult): string {
  if (result.failureCode) return workflowRAGApplicationFailure(t, result.failureCode);
  return result.status === "succeeded" ? t($ => $.ragApplication.invocation.succeeded) : t($ => $.ragApplication.invocation.idle);
}
