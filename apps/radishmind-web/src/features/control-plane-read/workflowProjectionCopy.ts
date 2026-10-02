import type { TFunction } from "i18next";
import type { workflowInspectionProjection } from "../../i18n/locales/en-US/workflowInspectionProjection.ts";
import type { workflowHandoffProjection } from "../../i18n/locales/en-US/workflowHandoffProjection.ts";

import { workflowDraftStatusLabel } from "./workflowDraftMessages.ts";

type WorkflowProjectionResources = typeof workflowInspectionProjection & typeof workflowHandoffProjection;

export type WorkflowProjectionValue = string | number | WorkflowProjectionMessage;
type MessageParameters<Text extends string> = Text extends `${string}{{${infer Name}}}${infer Rest}`
  ? Name | MessageParameters<Rest> : never;
type ProjectionTemplateMessage = {
  [Key in keyof WorkflowProjectionResources]: { key: Key } &
    (MessageParameters<WorkflowProjectionResources[Key]> extends never
      ? { values?: never }
      : { values: Record<MessageParameters<WorkflowProjectionResources[Key]>, WorkflowProjectionValue> });
}[keyof WorkflowProjectionResources];
export type WorkflowProjectionMessage =
  | ProjectionTemplateMessage
  | { status: string }
  | { parts: WorkflowProjectionValue[]; separator: string };

export type WorkflowProjectionTextField =
  | "label" | "summary" | "missingPrerequisite" | "currentEvidence" | "blockedReason"
  | "reviewerQuestion" | "humanReviewQuestion" | "reviewQuestion" | "handoffNeed" | "role"
  | "targetSummary" | "handoffPath" | "handoffNarrative" | "reviewNarrative"
  | "localDraftPreservationSummary" | "nextReviewerStep" | "openUnavailableReason"
  | "intent" | "reason" | "policyReason" | "disabledReason" | "exampleSummary" | "evidenceRef";

// Optional, local presentation metadata. Raw text and capability decisions stay independent
// of UI locale; unannotated user content and server diagnostics are never guessed from prose.
export type WorkflowProjectionCopy = {
  [Field in `${WorkflowProjectionTextField}Message`]?: WorkflowProjectionMessage;
};

export function workflowProjectionMessage(t: TFunction<"workflow">, message: WorkflowProjectionMessage): string {
  if ("status" in message) return workflowProjectionStatusLabel(t, message.status);
  const render = (value: WorkflowProjectionValue): string | number =>
    typeof value === "object" ? workflowProjectionMessage(t, value) : value;
  if ("parts" in message) return message.parts.map(render).join(message.separator);
  const values = Object.fromEntries(Object.entries(message.values ?? {}).map(([key, value]) => [key, render(value)]));
  // Parameter names are checked by ProjectionTemplateMessage before recursive rendering.
  return String(t($ => $.projection[message.key] as string, values));
}

export function workflowProjectionText<Field extends WorkflowProjectionTextField>(
  t: TFunction<"workflow">,
  source: WorkflowProjectionCopy & Record<Field, string | null>,
  field: Field,
): string {
  const message = source[`${field}Message`];
  return message ? workflowProjectionMessage(t, message) : source[field] ?? "";
}

export function workflowProjectionStatusLabel(t: TFunction<"workflow">, status: string): string {
  switch (status) {
    case "satisfied": return t($ => $.projection.status_satisfied);
    case "critical": return t($ => $.projection.status_critical);
    case "defined_not_connected": return t($ => $.projection.status_defined_not_connected);
    case "offline_preview_only": return t($ => $.projection.status_offline_preview_only);
    case "offline_only": return t($ => $.projection.status_offline_only);
    case "offline_read_only_advisory": return t($ => $.projection.status_offline_read_only_advisory);
    case "executor": return t($ => $.projection.status_executor);
    case "provider": return t($ => $.projection.status_provider);
    case "confirmation": return t($ => $.projection.status_confirmation);
    case "store": return t($ => $.projection.status_store);
    case "writeback": return t($ => $.projection.status_writeback);
    case "replay": return t($ => $.projection.status_replay);
    case "auth_store": return t($ => $.projection.status_auth_store);
    case "publish": return t($ => $.projection.status_publish);
    case "runtime": return t($ => $.projection.status_runtime);
    case "implementation_trigger": return t($ => $.projection.status_implementation_trigger);
    case "auth": return t($ => $.projection.status_auth);
    case "runtime_start": return t($ => $.projection.status_runtime_start);
    case "live_backend": return t($ => $.projection.status_live_backend);
    case "home": return t($ => $.projection.status_home);
    case "review": return t($ => $.projection.status_review);
    case "scenario": return t($ => $.projection.status_scenario);
    case "overview": return t($ => $.projection.status_overview);
    case "validation": return t($ => $.projection.status_validation);
    case "saved_draft_conflict": return t($ => $.projection.status_saved_draft_conflict);
    case "node_designer": return t($ => $.projection.status_node_designer);
    case "plan": return t($ => $.projection.status_plan);
    case "readiness": return t($ => $.projection.status_readiness);
    case "blocked_action": return t($ => $.projection.status_blocked_action);
    case "stop_line": return t($ => $.projection.status_stop_line);
    case "application": return t($ => $.projection.status_application);
    case "definition": return t($ => $.projection.status_definition);
    case "draft": return t($ => $.projection.status_draft);
    case "run": return t($ => $.projection.status_run);
    case "node": return t($ => $.projection.status_node);
    case "edge": return t($ => $.projection.status_edge);
    case "graph": return t($ => $.projection.status_graph);
    default: return workflowDraftStatusLabel(t, status);
  }
}
