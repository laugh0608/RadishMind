import type { TFunction } from "i18next";
import type { WorkflowDefinitionCandidateCompatibility } from "./workflowDefinitionPromotionConsumer.ts";
import { workflowDraftStatusLabel } from "./workflowDraftMessages.ts";

export type WorkflowPromotionMessage =
  | { code: "created"; id: string; version: number }
  | { code: "reviewed"; decision: "approve" | "reject" }
  | { code: "activated"; decision: "activate" | "replace" | "deactivate"; version: number }
  | { code: "runFinished"; id: string; schema: string }
  | { code: "failed" | "inputInvalid"; failureCode: string }
  | { code: "conflict"; failureCode: string; review: number; pointer: number }
  | { code: "exactDraftRequired" }
  | { code: "incompatible"; compatibility: WorkflowDefinitionCandidateCompatibility };

export function workflowPromotionMessage(t: TFunction<"workflow">, message: WorkflowPromotionMessage): string {
  switch (message.code) {
    case "created": return t($ => $.promotion.created, message);
    case "reviewed": return t($ => $.promotion.reviewed, { decision: workflowDraftStatusLabel(t, message.decision) });
    case "activated": return t($ => $.promotion.activated, { decision: workflowDraftStatusLabel(t, message.decision), version: message.version });
    case "runFinished": return t($ => $.promotion.runFinished, message);
    case "failed": return t($ => $.promotion.failed, { code: message.failureCode });
    case "inputInvalid": return t($ => $.promotion.inputInvalid, { code: message.failureCode });
    case "conflict": return t($ => $.promotion.conflict, { code: message.failureCode, review: message.review, pointer: message.pointer });
    case "exactDraftRequired": return t($ => $.promotion.exactDraftRequired);
    case "incompatible": return message.compatibility.handoffAnchor
      ? t($ => $.promotion.incompatibleRag)
      : message.compatibility.executionProfile === "workflow_definition_http_tool_v1"
        ? t($ => $.promotion.incompatibleTool)
        : t($ => $.promotion.incompatibleOther, { types: message.compatibility.unsupportedNodeTypes.join(", ") });
  }
}

export type WorkflowTemplateAction = "create" | "review" | "listing" | "derive";
export type WorkflowTemplateMessage =
  | { code: "offline" | "loading" | "versionFailed" }
  | { code: "ready"; candidates: number; templates: number }
  | { code: "loadFailed"; failureCode: string }
  | { code: "working"; action: WorkflowTemplateAction }
  | { code: "failed"; action: WorkflowTemplateAction; failureCode: string; review: number; pointer: number }
  | { code: "derived"; id: string; version: number }
  | { code: "completed"; action: WorkflowTemplateAction; audit: string };

function templateActionLabel(t: TFunction<"workflow">, action: WorkflowTemplateAction): string {
  switch (action) {
    case "create": return t($ => $.template.create);
    case "review": return t($ => $.template.review);
    case "listing": return t($ => $.template.listing);
    case "derive": return t($ => $.template.derive);
  }
}

export function workflowTemplateMessage(t: TFunction<"workflow">, message: WorkflowTemplateMessage): string {
  switch (message.code) {
    case "offline": return t($ => $.template.offline);
    case "loading": return t($ => $.template.loading);
    case "versionFailed": return t($ => $.template.versionFailed);
    case "ready": return t($ => $.template.ready, message);
    case "loadFailed": return t($ => $.template.loadFailed, { code: message.failureCode });
    case "working": return t($ => $.template.working, { action: templateActionLabel(t, message.action) });
    case "failed": return t($ => $.template.failed, { action: templateActionLabel(t, message.action), code: message.failureCode, review: message.review, pointer: message.pointer });
    case "derived": return t($ => $.template.derived, message);
    case "completed": return t($ => $.template.completed, { action: templateActionLabel(t, message.action), audit: message.audit });
  }
}
