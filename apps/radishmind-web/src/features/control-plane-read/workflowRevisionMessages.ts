import type { TFunction } from "i18next";
import type { WorkflowSavedDraftRevisionHistoryState } from "./workflowSavedDraftRevisionConsumer.ts";
import type { WorkflowSavedDraftRevisionChange } from "./workflowSavedDraftRevisionComparison.ts";

export type WorkflowRevisionOperation =
  | { code: "reading" | "ready" | "restoring" | "restored"; version: number }
  | { code: "read_failed" }
  | { code: "restore_failed"; failureCode: string };

export function workflowRevisionOperationMessage(t: TFunction<"workflow">, operation: WorkflowRevisionOperation): string {
  switch (operation.code) {
    case "reading": return t($ => $.revision.revisionDetailsLoading, { version: operation.version });
    case "ready": return t($ => $.revision.revisionDetailsReady, { version: operation.version });
    case "restoring": return t($ => $.revision.revisionRestoring, { version: operation.version });
    case "restored": return t($ => $.revision.revisionRestored, { version: operation.version });
    case "read_failed": return t($ => $.revision.revisionDetailsFailed);
    case "restore_failed": return t($ => $.revision.revisionRestoreFailed, { code: operation.failureCode });
  }
}

export function workflowRevisionHistoryMessage(t: TFunction<"workflow">, history: WorkflowSavedDraftRevisionHistoryState): string {
  switch (history.status) {
    case "disabled": return t($ => $.revision.revisionDisabled);
    case "idle": return t($ => $.revision.revisionIdle);
    case "loading": return t($ => $.revision.revisionLoading);
    case "ready": return t($ => $.revision.revisionReady, { count: history.revisions.length });
    case "empty": return t($ => $.revision.revisionEmpty);
    case "failed": return t($ => $.revision.revisionFailed, { code: history.failureCode ?? "draft_revision_history_failed" });
  }
}

export function workflowRevisionChangeMessage(t: TFunction<"workflow">, change: WorkflowSavedDraftRevisionChange): string {
  switch (change.kind) {
    case "metadata": return change.before !== undefined && change.after !== undefined
      ? t($ => $.revision.metadataDifference, { before: change.before || t($ => $.revision.unset), after: change.after || t($ => $.revision.unset) })
      : t($ => $.revision.changeMetadata);
    case "layout": return t($ => $.revision.changeLayout);
    case "validation": return t($ => $.revision.changeValidation);
    case "provenance": return t($ => $.revision.changeProvenance);
    case "node_added": return t($ => $.revision.changeNodeAdded);
    case "node_removed": return t($ => $.revision.changeNodeRemoved);
    case "node_changed": return t($ => $.revision.changeNodeChanged);
    case "edge_added": return t($ => $.revision.changeEdgeAdded);
    case "edge_removed": return t($ => $.revision.changeEdgeRemoved);
    case "edge_changed": return t($ => $.revision.changeEdgeChanged);
  }
}
