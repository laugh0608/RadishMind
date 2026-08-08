import type {
  ApplicationDevelopmentStageId,
  ApplicationDevelopmentSurfaceKind,
  ApplicationDevelopmentWorkspaceStatus,
} from "./applicationDevelopmentWorkspace.ts";

export type PromptAgentTypeWorkspaceSurface =
  | "source"
  | "configuration"
  | "candidate"
  | "assignment"
  | "access"
  | "controlled_use"
  | "session"
  | "evaluation";

export type PromptAgentTypeTaskAvailability = "available" | "read_only" | "blocked";

export type PromptAgentTypeWorkspaceTask = {
  surface: PromptAgentTypeWorkspaceSurface;
  anchor: string;
  label: string;
  summary: string;
  number: string;
  availability: PromptAgentTypeTaskAvailability;
};

type TaskDefinition = Omit<PromptAgentTypeWorkspaceTask, "availability"> & {
  archivedAvailability: Exclude<PromptAgentTypeTaskAvailability, "available">;
};

const COMMON_REVIEW_ANCHORS = new Set([
  "workspace-run-history",
  "workflow-run-comparison",
  "workflow-evaluation-cases",
  "workflow-evaluation-release-review",
]);

const PROMPT_TASKS: readonly TaskDefinition[] = [
  task("source", "prompt-application-template-workspace", "Template", "Draft, preview, immutable version", "01", "read_only"),
  task("configuration", "application-configuration-draft", "Configuration", "Bind the exact Template version", "02", "read_only"),
  task("candidate", "application-publish-review", "Candidate", "Human review, no automatic activation", "03", "read_only"),
  task("assignment", "prompt-application-runtime-assignment", "Assignment", "Explicit activate, replace, or revoke", "04", "read_only"),
  task("access", "application-api-integration", "Access", "Existing API Integration and Key owner", "05", "read_only"),
  task("controlled_use", "prompt-application-invocation", "Invocation", "One exact-authority provider call", "06", "blocked"),
  task("session", "prompt-application-session", "Session", "Metadata-only Session v2", "07", "blocked"),
  task("evaluation", "workspace-run-history", "Evaluation", "Run v6, Comparison v5, human evidence", "08", "read_only"),
];

const AGENT_TASKS: readonly TaskDefinition[] = [
  task("source", "agent-copilot-profile-workspace", "Profile", "Structured advisory source and version", "01", "read_only"),
  task("configuration", "application-configuration-draft", "Configuration", "Bind the exact Profile version", "02", "read_only"),
  task("candidate", "application-publish-review", "Candidate", "Human review, no automatic activation", "03", "read_only"),
  task("assignment", "agent-copilot-runtime-assignment", "Assignment", "Explicit activate, replace, or revoke", "04", "read_only"),
  task("access", "workspace-api-keys", "Access", "Existing scoped Key owner", "05", "read_only"),
  task("controlled_use", "agent-copilot-invocation", "Suggestion", "One Session v3 advisory suggestion", "06", "blocked"),
  task("evaluation", "workspace-run-history", "Evaluation", "Run v7, Comparison v6, human evidence", "07", "read_only"),
];

export function promptAgentTypeWorkspaceTasks(
  surfaceKind: ApplicationDevelopmentSurfaceKind,
  status: ApplicationDevelopmentWorkspaceStatus,
): PromptAgentTypeWorkspaceTask[] {
  const definitions = surfaceKind === "prompt_application"
    ? PROMPT_TASKS
    : surfaceKind === "agent_copilot"
      ? AGENT_TASKS
      : [];
  return definitions.map(({ archivedAvailability, ...definition }) => ({
    ...definition,
    availability: status === "active" ? "available" : archivedAvailability,
  }));
}

export function promptAgentTypeWorkspaceSurfaceForHash(
  hash: string,
  surfaceKind: ApplicationDevelopmentSurfaceKind,
  activeStage: ApplicationDevelopmentStageId | null,
): PromptAgentTypeWorkspaceSurface | null {
  if (surfaceKind !== "prompt_application" && surfaceKind !== "agent_copilot") return null;
  const anchor = hash.trim().replace(/^#/u, "");
  if (COMMON_REVIEW_ANCHORS.has(anchor)) return "evaluation";
  if (anchor === "application-configuration-draft") return "configuration";
  if (anchor === "application-publish-review") return "candidate";
  if (anchor === "application-api-integration" || anchor === "workspace-api-keys") return "access";

  if (surfaceKind === "prompt_application") {
    if (anchor === "prompt-application-template-workspace") return "source";
    if (anchor === "prompt-application-runtime-assignment") return "assignment";
    if (anchor === "prompt-application-invocation") return "controlled_use";
    if (anchor === "prompt-application-session" || anchor === "application-interaction-session") return "session";
  } else {
    if (anchor === "agent-copilot-profile-workspace") return "source";
    if (anchor === "agent-copilot-runtime-assignment") return "assignment";
    if (anchor === "agent-copilot-invocation" || anchor === "agent-copilot-session" || anchor === "application-interaction-session") {
      return "controlled_use";
    }
  }

  if (activeStage === "configure_build") return "source";
  if (activeStage === "human_promotion") return "candidate";
  if (activeStage === "controlled_test") return "controlled_use";
  if (activeStage === "evidence_review") return "evaluation";
  return null;
}

export function promptAgentTypeWorkspaceOwnsHash(
  hash: string,
  surfaceKind: ApplicationDevelopmentSurfaceKind,
): boolean {
  return promptAgentTypeWorkspaceSurfaceForHash(hash, surfaceKind, null) !== null;
}

function task(
  surface: PromptAgentTypeWorkspaceSurface,
  anchor: string,
  label: string,
  summary: string,
  number: string,
  archivedAvailability: Exclude<PromptAgentTypeTaskAvailability, "available">,
): TaskDefinition {
  return { surface, anchor, label, summary, number, archivedAvailability };
}
