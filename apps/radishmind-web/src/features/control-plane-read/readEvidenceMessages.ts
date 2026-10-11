import type { TFunction } from "i18next";
import { workflowProjectionStatusLabel } from "./workflowProjectionCopy.ts";
const states = {
  "blocked": "blocked",
  "selected": "selected",
  "confirmation": "confirmation",
  "advisory": "advisory",
  "required": "required",
  "optional": "optional",
  "capable": "capable",
  "read-only": "read_only",
  "none": "none",
  "clear": "clear",
  "pending": "pending",
  "degraded": "degraded",
  "tenant_binding_missing": "tenant_binding_missing",
  "blocked_review": "blocked_review",
  "offline_read_only_advisory": "offline_read_only_advisory",
  "offline_scenario_advisory": "offline_scenario_advisory",
  "radishflow_candidate_action": "radishflow_candidate_action",
  "radish_docs_answer": "radish_docs_answer",
  "workflow_advisory_review": "workflow_advisory_review",
  "admin": "admin",
  "workspace": "workspace",
  "single_resource": "single_resource",
  "cursor_required": "cursor_required"
} as const;
export function readEvidenceStatus(t: TFunction<"workflow">, value: string): string {
 return Object.hasOwn(states, value) ? t($ => $.readEvidence.states[states[value as keyof typeof states]]) : workflowProjectionStatusLabel(t, value);
}
