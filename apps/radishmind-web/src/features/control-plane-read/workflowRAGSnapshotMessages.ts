import type { TFunction } from "i18next";
import type { WorkflowRAGSnapshotFindingMessage } from "./workflowRAGSnapshotFeedback.ts";

export function workflowRAGSnapshotFindingMessage(t: TFunction<"workflow">, message: WorkflowRAGSnapshotFindingMessage): string {
  switch (message.key) {
    case "fileCount": return t($ => $.ragSnapshot.findings.fileCount, message.values);
    case "rawBudget": return t($ => $.ragSnapshot.findings.rawBudget);
    case "unsupportedFile": return t($ => $.ragSnapshot.findings.unsupportedFile, message.values);
    case "fileBudget": return t($ => $.ragSnapshot.findings.fileBudget, message.values);
    case "utf8": return t($ => $.ragSnapshot.findings.utf8, message.values);
    case "invalidContent": return t($ => $.ragSnapshot.findings.invalidContent, message.values);
    case "fileSecret": return t($ => $.ragSnapshot.findings.fileSecret, message.values);
    case "duplicateSource": return t($ => $.ragSnapshot.findings.duplicateSource, message.values);
    case "noContent": return t($ => $.ragSnapshot.findings.noContent, message.values);
    case "fragmentBudget": return t($ => $.ragSnapshot.findings.fragmentBudget, message.values);
    case "duplicateContent": return t($ => $.ragSnapshot.findings.duplicateContent, message.values);
    case "snapshotBudget": return t($ => $.ragSnapshot.findings.snapshotBudget);
    case "noFragments": return t($ => $.ragSnapshot.findings.noFragments);
    case "metadata": return t($ => $.ragSnapshot.findings.metadata);
    case "fragmentCount": return t($ => $.ragSnapshot.findings.fragmentCount, message.values);
    case "missingSource": return t($ => $.ragSnapshot.findings.missingSource);
    case "duplicateRef": return t($ => $.ragSnapshot.findings.duplicateRef, message.values);
    case "contentBudget": return t($ => $.ragSnapshot.findings.contentBudget);
    case "replacementInvalid": return t($ => $.ragSnapshot.findings.replacementInvalid);
    case "fragmentSecret": return t($ => $.ragSnapshot.findings.fragmentSecret);
    case "fragmentInvalid": return t($ => $.ragSnapshot.findings.fragmentInvalid);
  }
}
