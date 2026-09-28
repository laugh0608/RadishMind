import type { TFunction } from "i18next";
import type { WorkflowSavedDraftConsumerState, WorkflowSavedDraftListState, WorkflowSavedDraftLifecycleOperationState } from "./savedWorkflowDraftConsumer.ts";
import type { WorkflowDraftDesignerNode } from "./workflowDraftDesigner.ts";

// Stable protocol/status identifiers are mapped only at the display boundary.
export function workflowDraftStatusLabel(t: TFunction<"workflow">, status: string): string {
  switch (status) {
    case "active": return t($ => $.draft.statusActive);
    case "archived": return t($ => $.draft.statusArchived);
    case "unknown": return t($ => $.draft.statusUnknown);
    case "local_edit": return t($ => $.draft.statusLocalEdit);
    case "inspect_only": return t($ => $.draft.statusInspectOnly);
    case "select_only": return t($ => $.draft.statusSelectOnly);
    case "sample": return t($ => $.draft.statusSample);
    case "unsaved_local": return t($ => $.draft.statusUnsavedLocal);
    case "saving": return t($ => $.draft.statusSaving);
    case "validating": return t($ => $.draft.statusValidating);
    case "reading": return t($ => $.draft.statusReading);
    case "saved_dev_record": return t($ => $.draft.statusSavedDevRecord);
    case "validation_ready": return t($ => $.draft.statusValidationReady);
    case "version_conflict": return t($ => $.draft.statusVersionConflict);
    case "conflict_local_continued": return t($ => $.draft.statusConflictLocalContinued);
    case "save_failed": return t($ => $.draft.statusSaveFailed);
    case "read_failed": return t($ => $.draft.statusReadFailed);
    case "validation_failed": return t($ => $.draft.statusValidationFailed);
    case "loading": return t($ => $.draft.statusLoading);
    case "ready": return t($ => $.draft.statusReady);
    case "empty": return t($ => $.draft.statusEmpty);
    case "list_failed": return t($ => $.draft.statusListFailed);
    case "open_failed": return t($ => $.draft.statusOpenFailed);
    case "loaded": return t($ => $.draft.statusLoaded);
    case "refreshing": return t($ => $.draft.statusRefreshing);
    case "failed": return t($ => $.draft.statusFailed);
    case "disabled": return t($ => $.draft.statusDisabled);
    case "missing": return t($ => $.draft.statusMissing);
    case "needs_review": return t($ => $.draft.statusNeedsReview);
    case "local_draft_continued": return t($ => $.draft.statusLocalDraftContinued);
    case "open_available": return t($ => $.draft.statusOpenAvailable);
    case "open_requires_saved_list": return t($ => $.draft.statusOpenRequiresSavedList);
    case "valid_for_review": return t($ => $.draft.statusValidForReview);
    case "invalid_draft": return t($ => $.draft.statusInvalidDraft);
    case "blocked_capability": return t($ => $.draft.statusBlockedCapability);
    case "schema_unsupported": return t($ => $.draft.statusSchemaUnsupported);
    case "unversioned": return t($ => $.draft.statusUnversioned);
    case "workflow_definition": return t($ => $.draft.statusWorkflowDefinition);
    case "saved_draft_derivation": return t($ => $.draft.statusSavedDraftDerivation);
    case "workspace_template_derivation": return t($ => $.draft.statusWorkspaceTemplateDerivation);
    case "context": return t($ => $.draft.statusContext);
    case "model": return t($ => $.draft.statusModel);
    case "retrieval": return t($ => $.draft.statusRetrieval);
    case "policy": return t($ => $.draft.statusPolicy);
    case "preview": return t($ => $.draft.statusPreview);
    case "output": return t($ => $.draft.statusOutput);
    case "low": return t($ => $.draft.statusLow);
    case "medium": return t($ => $.draft.statusMedium);
    case "high": return t($ => $.draft.statusHigh);
    case "blocked": return t($ => $.draft.statusBlocked);
    case "review_required": return t($ => $.draft.statusReviewRequired);
    case "passed": return t($ => $.draft.statusPassed);
    case "warning": return t($ => $.draft.statusWarning);
    case "blocking": return t($ => $.draft.statusBlocking);
    case "info": return t($ => $.draft.statusInfo);
    case "none": return t($ => $.draft.statusNone);
    case "not_loaded": return t($ => $.draft.statusNotLoaded);
    case "saved": return t($ => $.draft.statusSaved);
    case "restored": return t($ => $.draft.statusRestored);
    case "backfilled_current": return t($ => $.draft.statusBackfilledCurrent);
    case "data_edge": return t($ => $.draft.statusDataEdge);
    case "control_edge": return t($ => $.draft.statusControlEdge);
    case "guard_edge": return t($ => $.draft.statusGuardEdge);
    case "audit_edge": return t($ => $.draft.statusAuditEdge);
    case "data": return t($ => $.draft.statusData);
    case "control": return t($ => $.draft.statusControl);
    case "audit": return t($ => $.draft.statusAudit);
    case "prompt": return t($ => $.draft.statusPrompt);
    case "llm": return t($ => $.draft.statusLlm);
    case "condition": return t($ => $.draft.statusCondition);
    case "http_tool": return t($ => $.draft.statusHttpTool);
    case "rag_retrieval": return t($ => $.draft.statusRagRetrieval);
    case "ready_for_review": return t($ => $.draft.status_ready_for_review);
    case "needs_policy_review": return t($ => $.draft.status_needs_policy_review);
    case "blocked_missing_runtime": return t($ => $.draft.status_blocked_missing_runtime);
    default: return status;
  }
}

export function workflowDraftConsumerMessage(t: TFunction<"workflow">, state: WorkflowSavedDraftConsumerState): string {
  const options = { version: state.currentDraftVersion, lifecycle: workflowDraftStatusLabel(t, state.currentLifecycleState), validation: workflowDraftStatusLabel(t, state.sourceLabel), code: state.failureCode ?? t($ => $.draft.none) };
  switch (state.status) {
    case "sample": return t($ => $.draft.consumerSample, options);
    case "unsaved_local": return t($ => $.draft.consumerUnsavedLocal, options);
    case "saving": return t($ => $.draft.consumerSaving, options);
    case "validating": return t($ => $.draft.consumerValidating, options);
    case "reading": return t($ => $.draft.consumerReading, options);
    case "saved_dev_record": return t($ => $.draft.consumerSavedDevRecord, options);
    case "validation_ready": return t($ => $.draft.consumerValidationReady, options);
    case "version_conflict": return t($ => $.draft.consumerVersionConflict, options);
    case "conflict_local_continued": return t($ => $.draft.consumerConflictLocalContinued, options);
    case "save_failed": return t($ => $.draft.consumerSaveFailed, options);
    case "read_failed": return t($ => $.draft.consumerReadFailed, options);
    case "validation_failed": return t($ => $.draft.consumerValidationFailed, options);
  }
}

export function workflowDraftLibraryMessage(t: TFunction<"workflow">, state: WorkflowSavedDraftListState): string {
  const options = { count: state.summaries.length, code: state.failureCode ?? t($ => $.draft.none) };
  switch (state.status) {
    case "sample": return t($ => $.draft.librarySample, options);
    case "loading": return t($ => $.draft.libraryLoading, options);
    case "ready": return t($ => $.draft.libraryReady, options);
    case "empty": return t($ => $.draft.libraryEmpty, options);
    case "list_failed": return t($ => $.draft.libraryListFailed, options);
    case "open_failed": return t($ => $.draft.libraryOpenFailed, options);
  }
}

export function workflowDraftLifecycleMessage(t: TFunction<"workflow">, state: WorkflowSavedDraftLifecycleOperationState): string {
  const options = { id: state.draftId, code: state.failureCode ?? t($ => $.draft.none) };
  switch (state.status) {
    case "idle": return t($ => $.draft.lifecycleIdle, options);
    case "transitioning": return t($ => $.draft.lifecycleTransitioning, options);
    case "archived": return t($ => $.draft.lifecycleArchived, options);
    case "unarchived": return t($ => $.draft.lifecycleUnarchived, options);
    case "failed": return t($ => $.draft.lifecycleFailed, options);
  }
}

export function workflowDraftNodeTypeLabel(t: TFunction<"workflow">, nodeType: WorkflowDraftDesignerNode["nodeType"]): string {
  switch (nodeType) {
    case "prompt": return t($ => $.draft.palettePromptLabel);
    case "llm": return t($ => $.draft.paletteLlmLabel);
    case "rag_retrieval": return t($ => $.draft.paletteRagRetrievalLabel);
    case "condition": return t($ => $.draft.paletteConditionLabel);
    case "http_tool": return t($ => $.draft.paletteHttpToolLabel);
    case "output": return t($ => $.draft.paletteOutputLabel);
  }
}

export function workflowDraftNodeTypeSummary(t: TFunction<"workflow">, nodeType: WorkflowDraftDesignerNode["nodeType"]): string {
  switch (nodeType) {
    case "prompt": return t($ => $.draft.palettePromptSummary);
    case "llm": return t($ => $.draft.paletteLlmSummary);
    case "rag_retrieval": return t($ => $.draft.paletteRagRetrievalSummary);
    case "condition": return t($ => $.draft.paletteConditionSummary);
    case "http_tool": return t($ => $.draft.paletteHttpToolSummary);
    case "output": return t($ => $.draft.paletteOutputSummary);
  }
}
