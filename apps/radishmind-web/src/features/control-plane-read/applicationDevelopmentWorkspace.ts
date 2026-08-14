export const APPLICATION_DEVELOPMENT_STAGE_IDS = [
  "configure_build",
  "human_promotion",
  "controlled_test",
  "evidence_review",
  "release_readiness",
] as const;

export type ApplicationDevelopmentStageId = typeof APPLICATION_DEVELOPMENT_STAGE_IDS[number];
export type ApplicationRuntimeReviewSurface = "run" | "request" | "evidence";
export type WorkflowReviewSurface = "runs" | "comparison" | "cases" | "release";
export type ApplicationDevelopmentWorkspaceStatus = "active" | "archived" | "unavailable";
export type ApplicationDevelopmentStageAvailability = "available" | "read_only" | "blocked";
export type ApplicationDevelopmentSurfaceKind =
  | "workflow_rag"
  | "prompt_application"
  | "agent_copilot"
  | "unsupported";

export type ApplicationDevelopmentWorkspaceInput = {
  applicationId: string;
  displayName: string;
  applicationKind: string;
  lifecycleState: "active" | "archived" | "";
  recordVersion: number;
  updatedAt: string;
  ownerSubjectRef: string;
  workspaceId: string;
  source: "application_catalog" | "offline_read_model";
};

export type ApplicationDevelopmentStage = {
  stageId: ApplicationDevelopmentStageId;
  label: string;
  summary: string;
  anchor: string;
  availability: ApplicationDevelopmentStageAvailability;
};

export type ApplicationDevelopmentWorkspaceContext = {
  status: ApplicationDevelopmentWorkspaceStatus;
  applicationId: string;
  displayName: string;
  applicationKind: string;
  surfaceKind: ApplicationDevelopmentSurfaceKind;
  lifecycleState: "active" | "archived" | "unavailable";
  recordVersion: number;
  updatedAt: string;
  ownerSubjectRef: string;
  workspaceId: string;
  source: ApplicationDevelopmentWorkspaceInput["source"];
  generationKey: string;
  applicationActive: boolean;
  stages: ApplicationDevelopmentStage[];
};

const SCOPE_REF_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._:-]{2,127}$/u;

type ApplicationDevelopmentStageDefinition = Omit<ApplicationDevelopmentStage, "availability"> & {
  aliases: readonly string[];
};

const STAGE_DEFINITIONS: ReadonlyArray<ApplicationDevelopmentStageDefinition> = [
  {
    stageId: "configure_build",
    label: "Configure / Build",
    summary: "Review configuration drafts and the application kind's exact source authority.",
    anchor: "application-configuration-draft",
    aliases: [
      "workflow-rag-snapshot-panel",
      "workspace-workflow-definitions",
      "prompt-application-template-workspace",
      "agent-copilot-profile-workspace",
    ],
  },
  {
    stageId: "human_promotion",
    label: "Human Promotion",
    summary: "Review immutable candidates, activation, assignment, drift, and blockers.",
    anchor: "application-publish-review",
    aliases: [
      "workflow-rag-promotion-review",
      "workflow-definition-promotion",
      "prompt-application-runtime-assignment",
      "agent-copilot-runtime-assignment",
    ],
  },
  {
    stageId: "controlled_test",
    label: "Controlled Test",
    summary: "Run the selected kind's exact Workflow v5, RAG v4, Prompt Application v6, or Agent Copilot v7 profile.",
    anchor: "application-interaction-session",
    aliases: [
      "application-api-integration",
      "workspace-api-keys",
      "application-rag-invocation",
      "prompt-application-invocation",
      "prompt-application-session",
      "agent-copilot-invocation",
      "agent-copilot-session",
      "model-gateway-playground",
    ],
  },
  {
    stageId: "evidence_review",
    label: "Run / Evaluation Review",
    summary: "Inspect durable runs, comparison, evaluation, request, and operations evidence.",
    anchor: "workspace-run-history",
    aliases: [
      "application-evaluation-plan",
      "application-evaluation-campaign",
      "application-evaluation-pair",
      "application-evaluation-handoff",
      "workflow-run-comparison",
      "workflow-evaluation-cases",
      "workflow-evaluation-release-review",
      "workflow-rag-evaluation-panel",
      "application-operations",
      "model-gateway-request-history",
    ],
  },
  {
    stageId: "release_readiness",
    label: "Release Readiness",
    summary: "Review source coverage and blockers without creating a publish decision.",
    anchor: "application-development-workspace-readiness",
    aliases: [],
  },
];

export function buildApplicationDevelopmentWorkspaceContext(
  input: ApplicationDevelopmentWorkspaceInput,
): ApplicationDevelopmentWorkspaceContext {
  const applicationId = input.applicationId.trim();
  const displayName = input.displayName.trim();
  const surfaceKind = applicationDevelopmentSurfaceKind(input.applicationKind);
  const applicationAvailable = SCOPE_REF_PATTERN.test(applicationId) && displayName.length > 0 &&
    surfaceKind !== "unsupported";
  const lifecycleKnown = input.lifecycleState === "active" || input.lifecycleState === "archived";
  const status: ApplicationDevelopmentWorkspaceStatus = !applicationAvailable || !lifecycleKnown
    ? "unavailable"
    : input.lifecycleState === "archived"
      ? "archived"
      : "active";
  const scopeAvailable = status !== "unavailable";
  const lifecycleState = status === "unavailable" ? "unavailable" : status;
  const recordVersion = Number.isInteger(input.recordVersion) && input.recordVersion > 0
    ? input.recordVersion
    : 0;
  const updatedAt = input.updatedAt.trim();

  return {
    status,
    applicationId: scopeAvailable ? applicationId : "",
    displayName: scopeAvailable ? displayName : "No selected application",
    applicationKind: scopeAvailable ? input.applicationKind.trim() || "not_available" : "not_available",
    surfaceKind: scopeAvailable ? surfaceKind : "unsupported",
    lifecycleState,
    recordVersion,
    updatedAt,
    ownerSubjectRef: scopeAvailable ? input.ownerSubjectRef.trim() : "",
    workspaceId: scopeAvailable ? input.workspaceId.trim() : "",
    source: input.source,
    generationKey: applicationDevelopmentGenerationKey(
      scopeAvailable ? applicationId : "unavailable",
      lifecycleState,
      recordVersion,
      updatedAt,
    ),
    applicationActive: status === "active",
    stages: STAGE_DEFINITIONS.map((stage) => ({
      stageId: stage.stageId,
      label: stage.label,
      summary: stage.summary,
      anchor: stage.anchor,
      availability: stageAvailability(stage.stageId, status),
    })),
  };
}

export function applicationDevelopmentSurfaceKind(applicationKind: string): ApplicationDevelopmentSurfaceKind {
  switch (applicationKind.trim()) {
    case "workflow_copilot":
    case "docs_qa":
      return "workflow_rag";
    case "prompt_application":
      return "prompt_application";
    case "agent":
      return "agent_copilot";
    default:
      return "unsupported";
  }
}

export function applicationDevelopmentStageForHash(hash: string): ApplicationDevelopmentStageId | null {
  const anchor = hash.trim().replace(/^#/u, "");
  return STAGE_DEFINITIONS.find((stage) => stage.anchor === anchor || stage.aliases.includes(anchor))?.stageId ?? null;
}

export function applicationDevelopmentHashTargetsOwnerSurface(hash: string): boolean {
  const anchor = hash.trim().replace(/^#/u, "");
  if (applicationRuntimeReviewSurfaceForHash(anchor)) return false;
  if (workflowReviewSurfaceForHash(anchor)) return false;
  return STAGE_DEFINITIONS.some((stage) => stage.aliases.includes(anchor));
}

export function applicationRuntimeReviewSurfaceForHash(hash: string): ApplicationRuntimeReviewSurface | null {
  const anchor = hash.trim().replace(/^#/u, "");
  if (anchor === "model-gateway-playground") return "run";
  if (anchor === "model-gateway-request-history") return "request";
  if (anchor === "application-operations") return "evidence";
  return null;
}

export function workflowReviewSurfaceForHash(hash: string): WorkflowReviewSurface | null {
  const anchor = hash.trim().replace(/^#/u, "");
  if (anchor === "workspace-run-history") return "runs";
  if (anchor === "workflow-run-comparison") return "comparison";
  if (anchor === "workflow-evaluation-cases") return "cases";
  if (anchor === "workflow-evaluation-release-review") return "release";
  return null;
}

function applicationDevelopmentGenerationKey(
  applicationId: string,
  lifecycleState: ApplicationDevelopmentWorkspaceContext["lifecycleState"],
  recordVersion: number,
  updatedAt: string,
): string {
  return ["application-development", applicationId, lifecycleState, String(recordVersion), updatedAt || "no-revision-time"].join(":");
}

function stageAvailability(
  stageId: ApplicationDevelopmentStageId,
  status: ApplicationDevelopmentWorkspaceStatus,
): ApplicationDevelopmentStageAvailability {
  if (status === "unavailable") return "blocked";
  if (status === "active") return "available";
  return "read_only";
}
