import type { TFunction } from "i18next";
import type { WorkflowRAGPromotionOperationResult } from "./workflowRAGPromotionConsumer.ts";

export function workflowRAGPromotionStatus(t: TFunction<"workflow">, code: string): string {
  switch (code) {
    case "offline": return t($ => $.ragPromotion.status.offline);
    case "loaded": return t($ => $.ragPromotion.status.loaded);
    case "created": return t($ => $.ragPromotion.status.created);
    case "decided": return t($ => $.ragPromotion.status.decided);
    case "failed": return t($ => $.ragPromotion.status.failed);
    case "scope_denied": return t($ => $.ragPromotion.status.scope_denied);
    case "record_version_conflict": return t($ => $.ragPromotion.status.record_version_conflict);
    case "pending": return t($ => $.ragPromotion.status.pending);
    case "deferred": return t($ => $.ragPromotion.status.deferred);
    case "approved": return t($ => $.ragPromotion.status.approved);
    case "rejected": return t($ => $.ragPromotion.status.rejected);
    case "canceled": return t($ => $.ragPromotion.status.canceled);
    case "eligible": return t($ => $.ragPromotion.status.eligible);
    case "blocked": return t($ => $.ragPromotion.status.blocked);
    case "approve": return t($ => $.ragPromotion.status.approve);
    case "reject": return t($ => $.ragPromotion.status.reject);
    case "defer": return t($ => $.ragPromotion.status.defer);
    case "cancel": return t($ => $.ragPromotion.status.cancel);
    case "improved": return t($ => $.ragPromotion.status.improved);
    case "unchanged": return t($ => $.ragPromotion.status.unchanged);
    case "regressed": return t($ => $.ragPromotion.status.regressed);
    case "needs_review": return t($ => $.ragPromotion.status.needs_review);
    case "passed": return t($ => $.ragPromotion.status.passed);
    case "valid": return t($ => $.ragPromotion.status.valid);
    case "invalid": return t($ => $.ragPromotion.status.invalid);
    case "unknown": return t($ => $.ragPromotion.status.unknown);
    default: return t($ => $.ragPromotion.status.unknown);
  }
}

export function workflowRAGPromotionFailure(t: TFunction<"workflow">, code: string): string {
  switch (code) {
    case "workflow_rag_promotion_scope_denied": return t($ => $.ragPromotion.failure.workflow_rag_promotion_scope_denied);
    case "workflow_rag_promotion_payload_invalid": return t($ => $.ragPromotion.failure.workflow_rag_promotion_payload_invalid);
    case "workflow_rag_promotion_secret_material_forbidden": return t($ => $.ragPromotion.failure.workflow_rag_promotion_secret_material_forbidden);
    case "workflow_rag_promotion_not_found": return t($ => $.ragPromotion.failure.workflow_rag_promotion_not_found);
    case "workflow_rag_promotion_dataset_changed": return t($ => $.ragPromotion.failure.workflow_rag_promotion_dataset_changed);
    case "workflow_rag_promotion_dataset_archived": return t($ => $.ragPromotion.failure.workflow_rag_promotion_dataset_archived);
    case "workflow_rag_promotion_review_invalid": return t($ => $.ragPromotion.failure.workflow_rag_promotion_review_invalid);
    case "workflow_rag_promotion_review_not_eligible": return t($ => $.ragPromotion.failure.workflow_rag_promotion_review_not_eligible);
    case "workflow_rag_promotion_snapshot_changed": return t($ => $.ragPromotion.failure.workflow_rag_promotion_snapshot_changed);
    case "workflow_rag_promotion_snapshot_archived": return t($ => $.ragPromotion.failure.workflow_rag_promotion_snapshot_archived);
    case "workflow_rag_promotion_profile_changed": return t($ => $.ragPromotion.failure.workflow_rag_promotion_profile_changed);
    case "workflow_rag_promotion_draft_changed": return t($ => $.ragPromotion.failure.workflow_rag_promotion_draft_changed);
    case "workflow_rag_promotion_draft_invalid": return t($ => $.ragPromotion.failure.workflow_rag_promotion_draft_invalid);
    case "workflow_rag_promotion_application_archived": return t($ => $.ragPromotion.failure.workflow_rag_promotion_application_archived);
    case "workflow_rag_promotion_record_version_conflict": return t($ => $.ragPromotion.failure.workflow_rag_promotion_record_version_conflict);
    case "workflow_rag_promotion_transition_invalid": return t($ => $.ragPromotion.failure.workflow_rag_promotion_transition_invalid);
    case "workflow_rag_promotion_store_unavailable": return t($ => $.ragPromotion.failure.workflow_rag_promotion_store_unavailable);
    case "workflow_rag_promotion_store_contract_mismatch": return t($ => $.ragPromotion.failure.workflow_rag_promotion_store_contract_mismatch);
    case "workflow_rag_promotion_write_disabled": return t($ => $.ragPromotion.failure.workflow_rag_promotion_write_disabled);
    case "workflow_rag_promotion_not_approved": return t($ => $.ragPromotion.failure.workflow_rag_promotion_not_approved);
    case "workflow_rag_promotion_http_disabled": return t($ => $.ragPromotion.failure.workflow_rag_promotion_http_disabled);
    case "workflow_rag_binding_not_eligible": return t($ => $.ragPromotion.failure.workflow_rag_binding_not_eligible);
    default: return t($ => $.ragPromotion.failure.unknown);
  }
}

export function workflowRAGPromotionFeedback(t: TFunction<"workflow">, result: WorkflowRAGPromotionOperationResult): string {
  if (result.status === "loaded" && !result.detail) return t($ => $.ragPromotion.feedback.idle);
  switch (result.status) {
    case "created": return t($ => $.ragPromotion.feedback.created);
    case "loaded": return t($ => $.ragPromotion.feedback.loaded);
    case "decided": return t($ => $.ragPromotion.feedback.decided);
    case "failed": return t($ => $.ragPromotion.feedback.failed);
    case "record_version_conflict": return t($ => $.ragPromotion.feedback.record_version_conflict);
    case "offline": return t($ => $.ragPromotion.feedback.offline);
    case "scope_denied": return t($ => $.ragPromotion.feedback.scope_denied);
  }
}
