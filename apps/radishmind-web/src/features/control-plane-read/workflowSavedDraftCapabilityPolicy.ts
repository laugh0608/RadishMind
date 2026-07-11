import type { WorkflowDraftDesignerDraft } from "./workflowDraftDesigner.ts";

const REVIEW_ONLY_REQUESTED_CAPABILITIES = [
  "publish",
  "run",
  "confirmation_decision",
  "writeback",
  "replay",
];

export function workflowSavedDraftRequestedCapabilities(
  draft: WorkflowDraftDesignerDraft,
): string[] {
  return draft.executionProfile === "executor_v0"
    ? []
    : [...REVIEW_ONLY_REQUESTED_CAPABILITIES];
}
