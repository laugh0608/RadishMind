export type ControlledUseOwner = "prompt_invocation" | "prompt_session" | "agent_session";

export type ControlledUseFailureGuidance = {
  title: string;
  summary: string;
  assignmentAnchor: "prompt-application-runtime-assignment" | "agent-copilot-runtime-assignment";
  assignmentLabel: string;
  sideEffectSummary: string;
};

const PROMPT_INVOCATION_FAILURES: Record<string, Pick<ControlledUseFailureGuidance, "title" | "summary">> = {
  prompt_runtime_assignment_not_found: {
    title: "No active Prompt assignment",
    summary: "This application has no exact active Prompt runtime assignment. An approved candidate must be activated explicitly before controlled use.",
  },
  prompt_runtime_candidate_ineligible: {
    title: "Prompt assignment is no longer eligible",
    summary: "The assigned candidate or one of its exact authority references is not currently eligible. Review the candidate evidence before replacing or revoking the assignment.",
  },
  prompt_runtime_authority_changed: {
    title: "Prompt authority changed",
    summary: "The current assignment no longer matches the application, candidate, configuration draft, or template authority reviewed at activation.",
  },
};

const SESSION_FAILURES: Record<string, { titleSuffix: string; summary: string }> = {
  application_session_authority_not_found: {
    titleSuffix: "assignment is missing",
    summary: "No exact active runtime authority is available for this application and execution profile.",
  },
  application_session_authority_changed: {
    titleSuffix: "authority changed",
    summary: "The current assignment no longer matches one or more reviewed application, candidate, configuration, or source-version references.",
  },
  application_session_profile_ineligible: {
    titleSuffix: "profile is ineligible",
    summary: "The current application and execution profile no longer form an eligible exact runtime authority.",
  },
};

const SIDE_EFFECT_SUMMARY = "The server blocked this attempt before delegated Gateway or provider execution. No provider call was made.";

export function controlledUseFailureGuidance(
  owner: ControlledUseOwner,
  failureCode: string,
): ControlledUseFailureGuidance | null {
  if (owner === "prompt_invocation") {
    const failure = PROMPT_INVOCATION_FAILURES[failureCode];
    return failure ? {
      ...failure,
      assignmentAnchor: "prompt-application-runtime-assignment",
      assignmentLabel: "Open Prompt Assignment",
      sideEffectSummary: SIDE_EFFECT_SUMMARY,
    } : null;
  }

  const failure = SESSION_FAILURES[failureCode];
  if (!failure) return null;
  const prompt = owner === "prompt_session";
  const typeLabel = prompt ? "Prompt" : "Agent Copilot";
  return {
    title: `${typeLabel} ${failure.titleSuffix}`,
    summary: failure.summary,
    assignmentAnchor: prompt
      ? "prompt-application-runtime-assignment"
      : "agent-copilot-runtime-assignment",
    assignmentLabel: prompt ? "Open Prompt Assignment" : "Open Agent Assignment",
    sideEffectSummary: SIDE_EFFECT_SUMMARY,
  };
}
