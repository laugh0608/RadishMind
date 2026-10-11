import type { TFunction } from "i18next";
import { workflowRAGEvaluation } from "../../i18n/locales/en-US/workflowRAGEvaluation.ts";
import type { WorkflowRAGEvaluationMetrics, WorkflowRAGEvaluationOperationResult } from "./workflowRAGEvaluationDatasetConsumer.ts";

export function ragEvaluationState(t: TFunction<"workflow">, state: string): string {
  return Object.hasOwn(workflowRAGEvaluation.states, state)
    ? t($ => $.ragEvaluation.states[state as keyof typeof workflowRAGEvaluation.states]) : state;
}
export function ragEvaluationMetric(t: TFunction<"workflow">, metric: keyof WorkflowRAGEvaluationMetrics): string {
  return t($ => $.ragEvaluation.metrics[metric]);
}
const failureMessages = {
  workflow_rag_evaluation_http_disabled: "offline",
  workflow_rag_evaluation_scope_denied: "scope_denied",
  workflow_rag_evaluation_payload_invalid: "invalid",
  workflow_rag_evaluation_store_unavailable: "unavailable",
  workflow_rag_evaluation_version_conflict: "version_conflict",
  workflow_rag_evaluation_not_found: "notFound",
  workflow_rag_evaluation_binding_invalid: "bindingInvalid",
  workflow_rag_evaluation_classification_mismatch: "classificationMismatch",
  workflow_rag_evaluation_archived: "archivedFailure",
  workflow_rag_evaluation_review_failed: "reviewFailed",
} as const;
export function ragEvaluationFailureCode(code: string): string {
  return /^[a-z][a-z0-9_]{2,95}$/u.test(code) ? code : "";
}
export function ragEvaluationMessage(t: TFunction<"workflow">, result: Pick<WorkflowRAGEvaluationOperationResult, "status" | "failureCode">): string {
  const key = Object.hasOwn(failureMessages, result.failureCode)
    ? failureMessages[result.failureCode as keyof typeof failureMessages]
    : result.failureCode ? "failed" : result.status;
  return t($ => $.ragEvaluation.messages[key]);
}
