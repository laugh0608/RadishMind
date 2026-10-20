import { readPromptApplicationRuntime, type PromptApplicationRuntimeAssignment, type PromptApplicationRuntimeConfig } from "./promptApplicationRuntimeConsumer.ts";
import { readPromptTemplateVersion, type PromptTemplateConfig, type PromptTemplateVersion } from "./promptApplicationTemplateConsumer.ts";

export type PromptPlanReference = {
  key: string;
  assignmentVersion: number;
  source: PromptTemplateVersion;
};
export type PromptPlanReferenceRead =
  | { status: "ready"; reference: PromptPlanReference }
  | { status: "unavailable"; code: string }
  | { status: "cancelled" };

export async function readPromptPlanReference(
  runtimeConfig: PromptApplicationRuntimeConfig,
  templateConfig: PromptTemplateConfig,
  applicationId: string,
  isCurrent: () => boolean,
): Promise<PromptPlanReferenceRead> {
  if (!isCurrent()) return { status: "cancelled" };
  if (runtimeConfig.tenantRef !== templateConfig.tenantRef || runtimeConfig.workspaceId !== templateConfig.workspaceId ||
    runtimeConfig.subjectRef !== templateConfig.subjectRef || runtimeConfig.authMode !== templateConfig.authMode) {
    return { status: "unavailable", code: "prompt_plan_reference_scope_mismatch" };
  }
  if (runtimeConfig.mode === "offline" || templateConfig.mode === "offline") return { status: "unavailable", code: "prompt_plan_reference_offline" };
  const first = await readPromptApplicationRuntime(runtimeConfig, applicationId, false);
  if (!isCurrent()) return { status: "cancelled" };
  if (first.status !== "ready" || first.assignment?.state !== "active") {
    return { status: "unavailable", code: first.failureCode || "prompt_plan_reference_inactive" };
  }
  const assignment = first.assignment;
  const ref = assignment.promptTemplateRef;
  const template = await readPromptTemplateVersion(templateConfig, applicationId, ref.templateId, ref.templateVersion, true);
  if (!isCurrent()) return { status: "cancelled" };
  const source = template.version;
  if (template.failureCode || !source) return { status: "unavailable", code: template.failureCode || "prompt_plan_reference_invalid" };
  if (source.tenantRef !== templateConfig.tenantRef || source.workspaceId !== templateConfig.workspaceId || source.applicationId !== applicationId ||
    source.templateId !== ref.templateId || source.templateVersion !== ref.templateVersion || source.templateDigest !== ref.templateDigest) {
    return { status: "unavailable", code: "prompt_plan_reference_mismatch" };
  }
  const last = await readPromptApplicationRuntime(runtimeConfig, applicationId, false);
  if (!isCurrent()) return { status: "cancelled" };
  if (last.status !== "ready" || last.assignment?.state !== "active" || assignmentKey(last.assignment) !== assignmentKey(assignment)) {
    return { status: "unavailable", code: "prompt_plan_reference_changed" };
  }
  return { status: "ready", reference: { key: assignmentKey(assignment), assignmentVersion: assignment.assignmentVersion, source } };
}

function assignmentKey(assignment: PromptApplicationRuntimeAssignment): string {
  const template = assignment.promptTemplateRef;
  return JSON.stringify([assignment.assignmentId, assignment.assignmentVersion, assignment.assignmentDigest,
    template.templateId, template.templateVersion, template.templateDigest]);
}
