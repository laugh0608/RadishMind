export type WorkflowSavedDraftVersionState = {
  status: string;
  currentDraftVersion: number;
};

export type WorkflowSavedDraftVersionResolution = {
  operation: "save" | "read" | "validate";
  failureCode: string | null;
  envelopeDraftVersion: number;
  previousDraftVersion: number;
};

export function workflowSavedDraftConflictRequiresResolution(state: WorkflowSavedDraftVersionState): boolean {
  return state.status === "version_conflict";
}

export function nextWorkflowSavedDraftExpectedVersion(state: WorkflowSavedDraftVersionState): number | null {
  if (workflowSavedDraftConflictRequiresResolution(state)) {
    return null;
  }
  return normalizeWorkflowSavedDraftVersion(state.currentDraftVersion);
}

export function resolveWorkflowSavedDraftVersion({
  operation,
  failureCode,
  envelopeDraftVersion,
  previousDraftVersion,
}: WorkflowSavedDraftVersionResolution): number {
  if (operation === "validate") {
    return normalizeWorkflowSavedDraftVersion(previousDraftVersion);
  }
  if (failureCode && failureCode !== "draft_version_conflict") {
    return normalizeWorkflowSavedDraftVersion(previousDraftVersion);
  }
  return normalizeWorkflowSavedDraftVersion(envelopeDraftVersion);
}

function normalizeWorkflowSavedDraftVersion(version: number): number {
  return Number.isSafeInteger(version) && version > 0 ? version : 0;
}
