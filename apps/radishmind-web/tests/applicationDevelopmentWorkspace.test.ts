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
import {
  clearApplicationDevelopmentHandoff,
  consumeApplicationDevelopmentHandoff,
  initialApplicationDevelopmentHandoffState,
  issueApplicationDevelopmentHandoff,
} from "../src/features/control-plane-read/applicationDevelopmentHandoff.ts";
import {
  APPLICATION_DEVELOPMENT_CONTRIBUTION_IDS,
  applyApplicationDevelopmentEvidence,
  buildApplicationDevelopmentReadinessViewModel,
  initialApplicationDevelopmentEvidenceState,
  type ApplicationDevelopmentContributionId,
  type ApplicationDevelopmentEvidenceInput,
} from "../src/features/control-plane-read/applicationDevelopmentReadiness.ts";

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

test("feature-scoped handoff keeps only one exact ref and consumes it at the owning stage", () => {
  const context = buildApplicationDevelopmentWorkspaceContext(activeApplication);
  const initial = initialApplicationDevelopmentHandoffState(context);
  const issued = issueApplicationDevelopmentHandoff(initial, context, {
    applicationId: context.applicationId,
    sourceStage: "controlled_test",
    refKind: "run",
    refId: "run-definition-v5-001",
  });

  assert.deepEqual(issued.pending, {
    handoffId: "workspace-handoff:1:run:run-definition-v5-001",
    applicationId: context.applicationId,
    workspaceGenerationKey: context.generationKey,
    sourceStage: "controlled_test",
    targetStage: "evidence_review",
    targetAnchor: "workspace-run-history",
    refKind: "run",
    refId: "run-definition-v5-001",
  });
  assert.equal(
    consumeApplicationDevelopmentHandoff(issued, context, "controlled_test", issued.pending?.handoffId ?? "").handoff,
    null,
  );
  const consumed = consumeApplicationDevelopmentHandoff(
    issued,
    context,
    "evidence_review",
    issued.pending?.handoffId ?? "",
  );
  assert.equal(consumed.handoff?.refId, "run-definition-v5-001");
  assert.equal(consumed.state.pending, null);
});

test("handoff rejects scope drift, unsafe refs, and stale workspace generations", () => {
  const context = buildApplicationDevelopmentWorkspaceContext(activeApplication);
  const initial = initialApplicationDevelopmentHandoffState(context);
  assert.throws(
    () => issueApplicationDevelopmentHandoff(initial, context, {
      applicationId: "app_other_scope",
      sourceStage: "configure_build",
      refKind: "draft",
      refId: "draft-001",
    }),
    /scope does not match/,
  );
  assert.throws(
    () => issueApplicationDevelopmentHandoff(initial, context, {
      applicationId: context.applicationId,
      sourceStage: "configure_build",
      refKind: "draft",
      refId: "unsafe ref",
    }),
    /reference is invalid/,
  );
  const revised = buildApplicationDevelopmentWorkspaceContext({ ...activeApplication, recordVersion: 4 });
  assert.throws(
    () => issueApplicationDevelopmentHandoff(initial, revised, {
      applicationId: revised.applicationId,
      sourceStage: "configure_build",
      refKind: "draft",
      refId: "draft-001",
    }),
    /generation is stale/,
  );
});

test("new handoffs replace older refs and workspace leave clears the pending selection", () => {
  const context = buildApplicationDevelopmentWorkspaceContext(activeApplication);
  const first = issueApplicationDevelopmentHandoff(initialApplicationDevelopmentHandoffState(context), context, {
    applicationId: context.applicationId,
    sourceStage: "configure_build",
    refKind: "draft",
    refId: "draft-001",
  });
  const second = issueApplicationDevelopmentHandoff(first, context, {
    applicationId: context.applicationId,
    sourceStage: "controlled_test",
    refKind: "run",
    refId: "run-002",
  });

  assert.equal(second.pending?.refId, "run-002");
  assert.equal(second.nextSequence, 3);
  assert.equal(clearApplicationDevelopmentHandoff(second, context).pending, null);
});

test("readiness starts incomplete for an active Application and rolls up all seven source groups", () => {
  const context = buildApplicationDevelopmentWorkspaceContext(activeApplication);
  const state = initialApplicationDevelopmentEvidenceState(context);
  const view = buildApplicationDevelopmentReadinessViewModel(state);

  assert.equal(view.status, "review_incomplete");
  assert.equal(view.sources.length, 7);
  assert.equal(view.sources.find((source) => source.sourceGroupId === "application")?.status, "available");
  assert.equal(view.canPersistReadiness, false);
  assert.equal(view.canPublish, false);
});

test("readiness distinguishes no selected Application from an archived lifecycle blocker", () => {
  const unavailable = buildApplicationDevelopmentWorkspaceContext({
    ...activeApplication,
    applicationId: "",
    displayName: "",
    lifecycleState: "",
    recordVersion: 0,
  });
  const archived = buildApplicationDevelopmentWorkspaceContext({
    ...activeApplication,
    lifecycleState: "archived",
  });

  assert.equal(
    buildApplicationDevelopmentReadinessViewModel(initialApplicationDevelopmentEvidenceState(unavailable)).status,
    "review_not_started",
  );
  const archivedView = buildApplicationDevelopmentReadinessViewModel(initialApplicationDevelopmentEvidenceState(archived));
  assert.equal(archivedView.status, "review_blocked");
  assert.equal(archivedView.sources[0]?.blockers[0]?.code, "application_archived");
});

test("multi-owner groups stay incomplete until every contribution is available", () => {
  const context = buildApplicationDevelopmentWorkspaceContext(activeApplication);
  const state = applyApplicationDevelopmentEvidence(
    initialApplicationDevelopmentEvidenceState(context),
    context,
    evidenceInput(context, "configuration_draft", "draft-001"),
  );
  const source = buildApplicationDevelopmentReadinessViewModel(state).sources.find(
    (item) => item.sourceGroupId === "configuration_candidate",
  );

  assert.equal(source?.status, "incomplete");
  assert.equal(source?.coverage, "partial");
  assert.equal(source?.missingEvidence.some((item) => item.includes("publish candidate")), true);
});

test("readiness becomes reviewable only when every required owner contribution is complete", () => {
  const context = buildApplicationDevelopmentWorkspaceContext(activeApplication);
  const remaining = APPLICATION_DEVELOPMENT_CONTRIBUTION_IDS.filter(
    (contributionId) => contributionId !== "application_lifecycle",
  );
  const completed = remaining.reduce(
    (state, contributionId, index) => applyApplicationDevelopmentEvidence(
      state,
      context,
      evidenceInput(context, contributionId, `evidence-${index + 1}`),
    ),
    initialApplicationDevelopmentEvidenceState(context),
  );
  const view = buildApplicationDevelopmentReadinessViewModel(completed);

  assert.equal(view.status, "dev_test_evidence_reviewable");
  assert.equal(view.sources.every((source) => source.status === "available"), true);
  assert.equal(view.evidenceCount, 9);
});

test("owner failure and partial coverage block readiness without hiding other evidence", () => {
  const context = buildApplicationDevelopmentWorkspaceContext(activeApplication);
  const withFailure = applyApplicationDevelopmentEvidence(
    initialApplicationDevelopmentEvidenceState(context),
    context,
    {
      ...evidenceInput(context, "operations_coverage", "request-operations-001"),
      status: "partial_failure",
      coverage: "partial",
      missingEvidence: ["Workflow operations source is unavailable."],
      blockers: [{ code: "workflow_run_store_unavailable", summary: "Workflow operations evidence could not be loaded." }],
      failureCodes: ["workflow_run_store_unavailable"],
    },
  );
  const view = buildApplicationDevelopmentReadinessViewModel(withFailure);
  const operations = view.sources.find((source) => source.sourceGroupId === "operations");

  assert.equal(view.status, "review_blocked");
  assert.equal(operations?.status, "partial_failure");
  assert.equal(operations?.evidenceRefs[0]?.id, "request-operations-001");
  assert.deepEqual(operations?.failureCodes, ["workflow_run_store_unavailable"]);
});

test("readiness rejects evidence from another application or route generation", () => {
  const context = buildApplicationDevelopmentWorkspaceContext(activeApplication);
  const state = initialApplicationDevelopmentEvidenceState(context);
  assert.throws(
    () => applyApplicationDevelopmentEvidence(state, context, {
      ...evidenceInput(context, "configuration_draft", "draft-001"),
      applicationId: "app_other_scope",
    }),
    /scope does not match/,
  );
  assert.throws(
    () => applyApplicationDevelopmentEvidence(state, context, {
      ...evidenceInput(context, "configuration_draft", "draft-001"),
      workspaceGenerationKey: "stale-generation",
    }),
    /generation is stale/,
  );
});

function evidenceInput(
  context: ReturnType<typeof buildApplicationDevelopmentWorkspaceContext>,
  contributionId: ApplicationDevelopmentContributionId,
  refId: string,
): ApplicationDevelopmentEvidenceInput {
  const kindByContribution: Record<ApplicationDevelopmentContributionId, ApplicationDevelopmentEvidenceInput["evidenceRefs"][number]["kind"]> = {
    application_lifecycle: "application",
    configuration_draft: "draft",
    publish_candidate: "candidate",
    workflow_definition: "definition",
    rag_binding: "binding",
    rag_assignment: "assignment",
    controlled_run: "run",
    evaluation_review: "evaluation",
    operations_coverage: "request",
  };
  return {
    applicationId: context.applicationId,
    workspaceGenerationKey: context.generationKey,
    contributionId,
    status: "available",
    coverage: "complete",
    evidenceRefs: [{ kind: kindByContribution[contributionId], id: refId, version: 1 }],
    missingEvidence: [],
    blockers: [],
    failureCodes: [],
  };
}
