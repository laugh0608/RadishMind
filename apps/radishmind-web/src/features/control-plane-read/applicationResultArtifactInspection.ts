import type {
  ApplicationResultArtifactLifecycle,
  ApplicationResultArtifactReadResult,
  ApplicationResultArtifactSummary,
} from "./applicationResultArtifactConsumer.ts";

export type ApplicationResultArtifactInspection = {
  generation: number;
  summary: ApplicationResultArtifactSummary | null;
  readResult: ApplicationResultArtifactReadResult | null;
};

type InspectionAction =
  | { type: "begin"; generation: number }
  | { type: "select"; generation: number; summary: ApplicationResultArtifactSummary | null }
  | { type: "read"; generation: number; result: ApplicationResultArtifactReadResult }
  | { type: "transition"; generation: number; lifecycle: ApplicationResultArtifactLifecycle };

export function reduceApplicationResultArtifactInspection(
  state: ApplicationResultArtifactInspection,
  action: InspectionAction,
): ApplicationResultArtifactInspection {
  if (action.generation < state.generation) return state;
  if (action.type === "select") {
    return { generation: action.generation, summary: action.summary, readResult: null };
  }
  if (action.type === "begin") return { ...state, generation: action.generation };
  if (action.generation !== state.generation || !state.summary) return state;
  if (action.type === "read") {
    const { artifact, lifecycle } = action.result;
    if (artifact && (artifact.artifactId !== state.summary.artifactId ||
      artifact.applicationId !== state.summary.applicationId || artifact.sessionId !== state.summary.sessionId)) return state;
    if (lifecycle && (lifecycle.artifactId !== state.summary.artifactId ||
      lifecycle.applicationId !== state.summary.applicationId)) return state;
    return { ...state, readResult: action.result };
  }
  if (!state.readResult?.artifact || action.lifecycle.artifactId !== state.summary.artifactId ||
    action.lifecycle.applicationId !== state.summary.applicationId) return state;
  return {
    ...state,
    summary: {
      ...state.summary,
      lifecycleState: action.lifecycle.lifecycleState,
      lifecycleVersion: action.lifecycle.lifecycleVersion,
      archivedAt: action.lifecycle.archivedAt,
      lifecycleUpdatedAt: action.lifecycle.updatedAt,
    },
    readResult: { ...state.readResult, lifecycle: action.lifecycle },
  };
}
