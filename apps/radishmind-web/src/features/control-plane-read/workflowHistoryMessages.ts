import type { TFunction } from "i18next";

export type WorkflowHistoryMessage =
  | { code: "handoffOffline" | "handoffLoading" | "handoffLoaded" | "handoffUnavailable" | "handoffFailed"; id: string }
  | { code: "diagnosticNeedsDraft" | "diagnosticWorking" }
  | { code: "diagnosticRecorded"; scenario: string }
  | { code: "diagnosticFailed"; failureCode: string };

export function workflowHistoryMessage(t: TFunction<"workflow">, message: WorkflowHistoryMessage): string {
  switch (message.code) {
    case "handoffOffline": return t($ => $.history.handoffOffline, message);
    case "handoffLoading": return t($ => $.history.handoffLoading, message);
    case "handoffLoaded": return t($ => $.history.handoffLoaded, message);
    case "handoffUnavailable": return t($ => $.history.handoffUnavailable, message);
    case "handoffFailed": return t($ => $.history.handoffFailed, message);
    case "diagnosticNeedsDraft": return t($ => $.history.diagnosticNeedsDraft, message);
    case "diagnosticWorking": return t($ => $.history.diagnosticWorking, message);
    case "diagnosticRecorded": return t($ => $.history.diagnosticRecorded, message);
    case "diagnosticFailed": return t($ => $.history.diagnosticFailed, { code: message.failureCode });
  }
}
