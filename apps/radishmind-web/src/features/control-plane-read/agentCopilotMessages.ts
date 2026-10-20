import type { TFunction } from "i18next";
import type { AgentCopilotProfileFinding, AgentCopilotProfileOperation } from "./agentCopilotProfileConsumer.ts";
import type { AgentCopilotSessionResult } from "./agentCopilotSessionConsumer.ts";

// Select messages from stable codes and fields, never from translated diagnostics.
const findingCodes = {
  scope: "agent_copilot_profile_payload_invalid",
  profile_name: "agent_copilot_profile_payload_invalid",
  description: "agent_copilot_profile_payload_invalid",
  allowed_tasks: "agent_copilot_profile_project_task_invalid",
  allowed_locales: "agent_copilot_profile_policy_invalid",
  safety: "agent_copilot_profile_policy_invalid",
  profile: "agent_copilot_profile_secret_material_forbidden",
} as const;

export function profileFindingMessage(t: TFunction<"agent">, finding: AgentCopilotProfileFinding): string {
  if (Object.hasOwn(findingCodes, finding.field)) {
    const field = finding.field as keyof typeof findingCodes;
    if (findingCodes[field] === finding.code) return t($ => $.findings[field]);
  }
  return t($ => $.findings.unknown);
}

export function profileOperationMessage(t: TFunction<"agent">, operation: AgentCopilotProfileOperation, dirty: boolean): string {
  if (operation.status === "idle" && dirty) return t($ => $.profileStatus.dirty);
  return t($ => $.profileStatus[operation.status], {
    version: operation.status === "versioned" ? operation.currentProfileVersion : operation.currentDraftVersion,
  });
}

export function agentSessionMessage(t: TFunction<"agent">, result: AgentCopilotSessionResult): string {
  if (result.failureCode === "application_session_request_canceled") return t($ => $.session.canceled);
  if (result.status === "ready" && result.session) return t($ => $.session.created, { id: result.session.sessionId });
  return t($ => $.session[result.status], { sequence: result.turn?.sequence ?? 0 });
}
