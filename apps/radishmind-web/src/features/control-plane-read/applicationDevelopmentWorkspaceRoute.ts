import {
  applicationDevelopmentStageForHash,
  type ApplicationDevelopmentStageId,
  type ApplicationDevelopmentWorkspaceContext,
} from "./applicationDevelopmentWorkspace.ts";

export type ApplicationDevelopmentRouteReason = "initialize" | "hash_change" | "scope_restore" | "workspace_leave";

export type ApplicationDevelopmentRouteState = {
  workspaceGenerationKey: string;
  activeStage: ApplicationDevelopmentStageId | null;
  routeGeneration: number;
  surfaceKey: string;
};

export function initialApplicationDevelopmentRouteState(
  context: ApplicationDevelopmentWorkspaceContext,
  hash: string,
): ApplicationDevelopmentRouteState {
  return routeState(context.generationKey, stageForWorkspaceHash(hash), 1);
}

export function transitionApplicationDevelopmentRoute(
  current: ApplicationDevelopmentRouteState,
  context: ApplicationDevelopmentWorkspaceContext,
  hash: string,
  reason: ApplicationDevelopmentRouteReason = "hash_change",
): ApplicationDevelopmentRouteState {
  const activeStage = reason === "workspace_leave" ? null : stageForWorkspaceHash(hash);
  const scopeChanged = current.workspaceGenerationKey !== context.generationKey;
  const stageChanged = current.activeStage !== activeStage;
  if (!scopeChanged && !stageChanged && reason !== "scope_restore") return current;
  return routeState(context.generationKey, activeStage, current.routeGeneration + 1);
}

export function applicationDevelopmentRouteAcceptsResponse(
  expectedSurfaceKey: string,
  current: ApplicationDevelopmentRouteState,
): boolean {
  return expectedSurfaceKey === current.surfaceKey && current.activeStage !== null;
}

function stageForWorkspaceHash(hash: string): ApplicationDevelopmentStageId | null {
  const normalized = hash.trim().replace(/^#/u, "");
  if (!normalized || normalized === "application-development-workspace" || normalized === "workspace-applications") {
    return "configure_build";
  }
  return applicationDevelopmentStageForHash(normalized);
}

function routeState(
  workspaceGenerationKey: string,
  activeStage: ApplicationDevelopmentStageId | null,
  routeGeneration: number,
): ApplicationDevelopmentRouteState {
  return {
    workspaceGenerationKey,
    activeStage,
    routeGeneration,
    surfaceKey: `${workspaceGenerationKey}:route:${routeGeneration}:${activeStage ?? "inactive"}`,
  };
}
