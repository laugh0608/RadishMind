import assert from "node:assert/strict";
import test from "node:test";

import {
  applicationDevelopmentStageForHash,
  buildApplicationDevelopmentWorkspaceContext,
  type ApplicationDevelopmentWorkspaceInput,
} from "../src/features/control-plane-read/applicationDevelopmentWorkspace.ts";
import {
  applicationDevelopmentRouteAcceptsResponse,
  initialApplicationDevelopmentRouteState,
  transitionApplicationDevelopmentRoute,
} from "../src/features/control-plane-read/applicationDevelopmentWorkspaceRoute.ts";

const activeApplication: ApplicationDevelopmentWorkspaceInput = {
  applicationId: "app_flow_copilot",
  displayName: "RadishFlow Copilot",
  applicationKind: "workflow_copilot",
  lifecycleState: "active",
  recordVersion: 3,
  updatedAt: "2026-07-20T12:00:00Z",
  ownerSubjectRef: "subject_demo_user",
  workspaceId: "workspace_demo",
  source: "application_catalog",
};

test("application development context keeps one exact active scope", () => {
  const context = buildApplicationDevelopmentWorkspaceContext(activeApplication);
  assert.equal(context.status, "active");
  assert.equal(context.applicationId, "app_flow_copilot");
  assert.equal(context.applicationActive, true);
  assert.equal(context.stages.length, 5);
  assert.equal(context.stages.every((stage) => stage.availability === "available"), true);
  assert.match(context.generationKey, /app_flow_copilot:active:3/u);
});

test("application revision and lifecycle changes rotate the workspace generation", () => {
  const current = buildApplicationDevelopmentWorkspaceContext(activeApplication);
  const revised = buildApplicationDevelopmentWorkspaceContext({
    ...activeApplication,
    recordVersion: 4,
    updatedAt: "2026-07-20T12:05:00Z",
  });
  const archived = buildApplicationDevelopmentWorkspaceContext({
    ...activeApplication,
    lifecycleState: "archived",
    recordVersion: 4,
    updatedAt: "2026-07-20T12:06:00Z",
  });
  assert.notEqual(current.generationKey, revised.generationKey);
  assert.notEqual(revised.generationKey, archived.generationKey);
  assert.equal(archived.applicationActive, false);
  assert.equal(archived.stages.find((stage) => stage.stageId === "controlled_test")?.availability, "blocked");
  assert.equal(archived.stages.find((stage) => stage.stageId === "evidence_review")?.availability, "read_only");
});

test("missing application fails closed without fabricating a scope", () => {
  const context = buildApplicationDevelopmentWorkspaceContext({
    ...activeApplication,
    applicationId: "",
    displayName: "",
    lifecycleState: "",
    recordVersion: 0,
    updatedAt: "",
  });
  assert.equal(context.status, "unavailable");
  assert.equal(context.applicationId, "");
  assert.equal(context.displayName, "No selected application");
  assert.equal(context.stages.every((stage) => stage.availability === "blocked"), true);
});

test("unknown lifecycle fails closed even when an application reference is present", () => {
  const context = buildApplicationDevelopmentWorkspaceContext({
    ...activeApplication,
    lifecycleState: "",
  });
  assert.equal(context.status, "unavailable");
  assert.equal(context.applicationId, "");
  assert.equal(context.applicationActive, false);
  assert.equal(context.stages.every((stage) => stage.availability === "blocked"), true);
});

test("stage anchors resolve only known application development destinations", () => {
  assert.equal(applicationDevelopmentStageForHash("#application-configuration-draft"), "configure_build");
  assert.equal(applicationDevelopmentStageForHash("application-publish-review"), "human_promotion");
  assert.equal(applicationDevelopmentStageForHash("#application-interaction-session"), "controlled_test");
  assert.equal(applicationDevelopmentStageForHash("#workspace-run-history"), "evidence_review");
  assert.equal(applicationDevelopmentStageForHash("#application-development-workspace-readiness"), "release_readiness");
  assert.equal(applicationDevelopmentStageForHash("#admin-audit-log"), null);
});

test("existing panel and handoff anchors resolve to their owning stages", () => {
  assert.equal(applicationDevelopmentStageForHash("#workflow-rag-snapshot-panel"), "configure_build");
  assert.equal(applicationDevelopmentStageForHash("#workflow-rag-promotion-review"), "human_promotion");
  assert.equal(applicationDevelopmentStageForHash("#application-api-integration"), "controlled_test");
  assert.equal(applicationDevelopmentStageForHash("#workspace-api-keys"), "controlled_test");
  assert.equal(applicationDevelopmentStageForHash("#model-gateway-playground"), "controlled_test");
  assert.equal(applicationDevelopmentStageForHash("#workflow-rag-evaluation-panel"), "evidence_review");
  assert.equal(applicationDevelopmentStageForHash("#application-operations"), "evidence_review");
  assert.equal(applicationDevelopmentStageForHash("#model-gateway-request-history"), "evidence_review");
  assert.equal(applicationDevelopmentStageForHash("#admin-audit-log"), null);
});

test("route transitions rotate only when stage or application scope changes", () => {
  const context = buildApplicationDevelopmentWorkspaceContext(activeApplication);
  const configure = initialApplicationDevelopmentRouteState(context, "#application-configuration-draft");
  const controlled = transitionApplicationDevelopmentRoute(configure, context, "#application-api-integration");
  const sameStage = transitionApplicationDevelopmentRoute(controlled, context, "#workspace-api-keys");
  const revisedContext = buildApplicationDevelopmentWorkspaceContext({
    ...activeApplication,
    recordVersion: 4,
    updatedAt: "2026-07-20T12:05:00Z",
  });
  const revised = transitionApplicationDevelopmentRoute(sameStage, revisedContext, "#workspace-api-keys");

  assert.equal(configure.activeStage, "configure_build");
  assert.equal(controlled.activeStage, "controlled_test");
  assert.notEqual(configure.surfaceKey, controlled.surfaceKey);
  assert.equal(sameStage, controlled);
  assert.notEqual(revised.surfaceKey, controlled.surfaceKey);
});

test("workspace leave and explicit restore reject late route responses", () => {
  const context = buildApplicationDevelopmentWorkspaceContext(activeApplication);
  const evidence = initialApplicationDevelopmentRouteState(context, "#workspace-run-history");
  const expectedSurfaceKey = evidence.surfaceKey;
  const left = transitionApplicationDevelopmentRoute(evidence, context, "#admin-audit-log", "workspace_leave");
  const restored = transitionApplicationDevelopmentRoute(left, context, "#workspace-run-history", "scope_restore");

  assert.equal(applicationDevelopmentRouteAcceptsResponse(expectedSurfaceKey, evidence), true);
  assert.equal(left.activeStage, null);
  assert.equal(applicationDevelopmentRouteAcceptsResponse(expectedSurfaceKey, left), false);
  assert.equal(restored.activeStage, "evidence_review");
  assert.notEqual(restored.surfaceKey, expectedSurfaceKey);
  assert.equal(applicationDevelopmentRouteAcceptsResponse(expectedSurfaceKey, restored), false);
});
