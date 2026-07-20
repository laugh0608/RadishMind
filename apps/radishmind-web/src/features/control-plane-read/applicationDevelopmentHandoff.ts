import type {
  ApplicationDevelopmentStageId,
  ApplicationDevelopmentWorkspaceContext,
} from "./applicationDevelopmentWorkspace.ts";

export const APPLICATION_DEVELOPMENT_HANDOFF_REF_KINDS = [
  "draft",
  "candidate",
  "definition",
  "binding",
  "assignment",
  "session",
  "run",
  "request",
  "evaluation",
] as const;

export type ApplicationDevelopmentHandoffRefKind =
  typeof APPLICATION_DEVELOPMENT_HANDOFF_REF_KINDS[number];

export type ApplicationDevelopmentHandoff = {
  handoffId: string;
  applicationId: string;
  workspaceGenerationKey: string;
  sourceStage: ApplicationDevelopmentStageId;
  targetStage: ApplicationDevelopmentStageId;
  targetAnchor: string;
  refKind: ApplicationDevelopmentHandoffRefKind;
  refId: string;
};

export type ApplicationDevelopmentHandoffState = {
  workspaceGenerationKey: string;
  nextSequence: number;
  pending: ApplicationDevelopmentHandoff | null;
};

export type ApplicationDevelopmentHandoffInput = {
  applicationId: string;
  sourceStage: ApplicationDevelopmentStageId;
  refKind: ApplicationDevelopmentHandoffRefKind;
  refId: string;
};

const REF_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._:-]{1,159}$/u;

const DESTINATIONS: Record<
  ApplicationDevelopmentHandoffRefKind,
  { targetStage: ApplicationDevelopmentStageId; targetAnchor: string }
> = {
  draft: { targetStage: "human_promotion", targetAnchor: "application-publish-review" },
  candidate: { targetStage: "human_promotion", targetAnchor: "application-publish-review" },
  definition: { targetStage: "controlled_test", targetAnchor: "application-interaction-session" },
  binding: { targetStage: "human_promotion", targetAnchor: "workflow-rag-promotion-review" },
  assignment: { targetStage: "controlled_test", targetAnchor: "application-rag-invocation" },
  session: { targetStage: "controlled_test", targetAnchor: "application-interaction-session" },
  run: { targetStage: "evidence_review", targetAnchor: "workspace-run-history" },
  request: { targetStage: "evidence_review", targetAnchor: "model-gateway-request-history" },
  evaluation: { targetStage: "evidence_review", targetAnchor: "workflow-rag-evaluation-panel" },
};

export function initialApplicationDevelopmentHandoffState(
  context: ApplicationDevelopmentWorkspaceContext,
): ApplicationDevelopmentHandoffState {
  return {
    workspaceGenerationKey: context.generationKey,
    nextSequence: 1,
    pending: null,
  };
}

export function issueApplicationDevelopmentHandoff(
  state: ApplicationDevelopmentHandoffState,
  context: ApplicationDevelopmentWorkspaceContext,
  input: ApplicationDevelopmentHandoffInput,
): ApplicationDevelopmentHandoffState {
  assertCurrentGeneration(state, context);
  const applicationId = input.applicationId.trim();
  const refId = input.refId.trim();
  if (!context.applicationId || applicationId !== context.applicationId) {
    throw new Error("Application development handoff scope does not match the current workspace.");
  }
  if (!APPLICATION_DEVELOPMENT_HANDOFF_REF_KINDS.includes(input.refKind) || !REF_PATTERN.test(refId)) {
    throw new Error("Application development handoff reference is invalid.");
  }
  const destination = DESTINATIONS[input.refKind];
  const handoff: ApplicationDevelopmentHandoff = {
    handoffId: `workspace-handoff:${state.nextSequence}:${input.refKind}:${refId}`,
    applicationId,
    workspaceGenerationKey: context.generationKey,
    sourceStage: input.sourceStage,
    targetStage: destination.targetStage,
    targetAnchor: destination.targetAnchor,
    refKind: input.refKind,
    refId,
  };
  return {
    workspaceGenerationKey: context.generationKey,
    nextSequence: state.nextSequence + 1,
    pending: handoff,
  };
}

export function consumeApplicationDevelopmentHandoff(
  state: ApplicationDevelopmentHandoffState,
  context: ApplicationDevelopmentWorkspaceContext,
  targetStage: ApplicationDevelopmentStageId,
  handoffId: string,
): { state: ApplicationDevelopmentHandoffState; handoff: ApplicationDevelopmentHandoff | null } {
  assertCurrentGeneration(state, context);
  const pending = state.pending;
  if (!pending || pending.handoffId !== handoffId || pending.targetStage !== targetStage) {
    return { state, handoff: null };
  }
  return {
    state: { ...state, pending: null },
    handoff: pending,
  };
}

export function clearApplicationDevelopmentHandoff(
  state: ApplicationDevelopmentHandoffState,
  context: ApplicationDevelopmentWorkspaceContext,
): ApplicationDevelopmentHandoffState {
  assertCurrentGeneration(state, context);
  if (!state.pending) return state;
  return { ...state, pending: null };
}

function assertCurrentGeneration(
  state: ApplicationDevelopmentHandoffState,
  context: ApplicationDevelopmentWorkspaceContext,
): void {
  if (state.workspaceGenerationKey !== context.generationKey) {
    throw new Error("Application development handoff generation is stale.");
  }
}
