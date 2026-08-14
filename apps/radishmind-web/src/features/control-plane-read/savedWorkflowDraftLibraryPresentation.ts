import type {
  WorkflowSavedDraftLifecycleState,
  WorkflowSavedDraftSummary,
} from "./savedWorkflowDraftConsumer.ts";

export type WorkflowSavedDraftLibraryActions = {
  primaryLabel: "打开草案" | "只读审查";
  readOnly: boolean;
  lifecycleLabel: "归档" | "解除归档";
  lifecycleTarget: WorkflowSavedDraftLifecycleState;
  lifecycleConfirmationRequired: boolean;
};

export function workflowSavedDraftLibraryActions(
  summary: Pick<WorkflowSavedDraftSummary, "lifecycleState">,
): WorkflowSavedDraftLibraryActions {
  if (summary.lifecycleState === "archived") {
    return {
      primaryLabel: "只读审查",
      readOnly: true,
      lifecycleLabel: "解除归档",
      lifecycleTarget: "active",
      lifecycleConfirmationRequired: false,
    };
  }
  return {
    primaryLabel: "打开草案",
    readOnly: false,
    lifecycleLabel: "归档",
    lifecycleTarget: "archived",
    lifecycleConfirmationRequired: true,
  };
}

export function workflowSavedDraftArchiveConfirmationSummary(draftId: string): string {
  return `归档 ${draftId} 后，草案将离开活动工作集；当前内容、不可变历史和比较仍可只读审查。`;
}
