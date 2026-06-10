import type { WorkflowBlockedActionPreviewViewModel } from "./workflowBlockedActionPreview";
import type { WorkflowConfirmationPlaceholderViewModel } from "./workflowConfirmationPlaceholder";
import type {
  WorkflowRuntimeReadinessBlocker,
  WorkflowRuntimeReadinessInspectorViewModel,
} from "./workflowRuntimeReadinessInspector";
import type { WorkflowScenarioInspectorViewModel } from "./workflowScenarioInspector";
import type { WorkflowSurfaceOverviewViewModel } from "./workflowSurfaceOverview";
import type {
  WorkflowUserWorkspaceHomeRouteEvidence,
  WorkflowUserWorkspaceHomeStatus,
  WorkflowUserWorkspaceHomeViewModel,
} from "./workflowUserWorkspaceHome";
import type {
  WorkflowWorkspaceReviewBlockedCapabilityGroup,
  WorkflowWorkspaceReviewStage,
  WorkflowWorkspaceReviewViewModel,
} from "./workflowWorkspaceReview";

export type WorkflowReviewHandoffStatus = WorkflowUserWorkspaceHomeStatus;

export type WorkflowReviewHandoffRecipient = {
  recipientId: string;
  label: string;
  role: string;
  status: WorkflowReviewHandoffStatus;
  handoffNeed: string;
  evidenceRefs: string[];
};

export type WorkflowReviewHandoffFinding = {
  findingId: string;
  label: string;
  sourceSurface:
    | "scenario"
    | "review"
    | "plan"
    | "readiness"
    | "blocked_action"
    | "confirmation"
    | "stop_line";
  status: WorkflowReviewHandoffStatus;
  summary: string;
  evidenceRef: string;
  humanReviewQuestion: string;
};

export type WorkflowReviewHandoffEvidence = {
  evidenceId: string;
  label: string;
  sourceSurface: "home" | "review" | "scenario" | "overview" | "blocked_action" | "confirmation";
  routeOrPageId: string;
  requestId: string;
  auditRef: string;
  status: WorkflowReviewHandoffStatus;
  summary: string;
};

export type WorkflowReviewHandoffDecisionBlocker = {
  blockerId: string;
  label: string;
  sourceSurface: string;
  status: "blocked";
  missingPrerequisite: string;
  summary: string;
  auditRefs: string[];
};

export type WorkflowReviewHandoffBoundaryLock = {
  boundaryId: string;
  label: string;
  status: "locked";
  summary: string;
};

export type WorkflowReviewHandoffSource = {
  workflowUserWorkspaceHome: WorkflowUserWorkspaceHomeViewModel;
  workflowWorkspaceReview: WorkflowWorkspaceReviewViewModel;
  workflowSurfaceOverview: WorkflowSurfaceOverviewViewModel;
  workflowScenarioInspector: WorkflowScenarioInspectorViewModel;
  workflowRuntimeReadinessInspector: WorkflowRuntimeReadinessInspectorViewModel;
  workflowBlockedActionPreview: WorkflowBlockedActionPreviewViewModel;
  workflowConfirmationPlaceholder: WorkflowConfirmationPlaceholderViewModel;
};

export type WorkflowReviewHandoffViewModel = {
  pageId: "workflow-review-handoff-offline";
  sourcePageIds: string[];
  handoffMode: "offline_read_only_advisory";
  handoffPackageId: string;
  tenantRef: string;
  applicationId: string;
  workflowDefinitionId: string;
  runId: string;
  draftId: string;
  scenarioId: string;
  requestId: string;
  auditRef: string;
  handoffNarrative: string;
  recipients: WorkflowReviewHandoffRecipient[];
  keyFindings: WorkflowReviewHandoffFinding[];
  evidenceChecklist: WorkflowReviewHandoffEvidence[];
  decisionBlockers: WorkflowReviewHandoffDecisionBlocker[];
  boundaryLocks: WorkflowReviewHandoffBoundaryLock[];
  canRenderReviewHandoff: boolean;
  canInspectHandoffLocally: true;
  canRequestLiveBackend: false;
  canExportHandoff: false;
  canSendHandoff: false;
  canPersistHandoff: false;
  canPersistReview: false;
  canPublishWorkflow: false;
  canStartRuntime: false;
  canSubmitConfirmationDecision: false;
  canWriteBusinessTruth: false;
  canReplayRun: false;
  canAttachDatabase: false;
  canEnableRadishAuth: false;
  canImplementRepositoryAdapter: false;
};

export function buildWorkflowReviewHandoffViewModel(
  source: WorkflowReviewHandoffSource,
): WorkflowReviewHandoffViewModel {
  const recipients = buildRecipients(source);
  const keyFindings = buildKeyFindings(source);
  const evidenceChecklist = buildEvidenceChecklist(source);
  const decisionBlockers = buildDecisionBlockers(source);
  const boundaryLocks = buildBoundaryLocks(source);

  return {
    pageId: "workflow-review-handoff-offline",
    sourcePageIds: [
      source.workflowUserWorkspaceHome.pageId,
      source.workflowWorkspaceReview.pageId,
      source.workflowSurfaceOverview.pageId,
      source.workflowScenarioInspector.pageId,
      source.workflowRuntimeReadinessInspector.pageId,
      source.workflowBlockedActionPreview.pageId,
      source.workflowConfirmationPlaceholder.pageId,
    ],
    handoffMode: "offline_read_only_advisory",
    handoffPackageId: buildHandoffPackageId(source),
    tenantRef: source.workflowUserWorkspaceHome.tenantRef,
    applicationId: source.workflowWorkspaceReview.applicationId,
    workflowDefinitionId: source.workflowWorkspaceReview.workflowDefinitionId,
    runId: source.workflowWorkspaceReview.runId,
    draftId: source.workflowWorkspaceReview.draftId,
    scenarioId: source.workflowWorkspaceReview.scenarioId,
    requestId: source.workflowWorkspaceReview.requestId,
    auditRef: source.workflowWorkspaceReview.auditRef,
    handoffNarrative: buildHandoffNarrative(source, decisionBlockers, boundaryLocks),
    recipients,
    keyFindings,
    evidenceChecklist,
    decisionBlockers,
    boundaryLocks,
    canRenderReviewHandoff:
      source.workflowUserWorkspaceHome.canRenderUserWorkspaceHome &&
      source.workflowWorkspaceReview.canRenderWorkspaceReview &&
      source.workflowSurfaceOverview.canRenderSurfaceOverview &&
      source.workflowScenarioInspector.canRenderScenarioInspector &&
      source.workflowRuntimeReadinessInspector.canRenderRuntimeReadinessInspector &&
      source.workflowBlockedActionPreview.canRenderBlockedActionPreview &&
      source.workflowConfirmationPlaceholder.canRenderConfirmationPlaceholder &&
      recipients.length === 4 &&
      keyFindings.length >= 7 &&
      evidenceChecklist.length >= 8 &&
      decisionBlockers.length >= 6 &&
      boundaryLocks.length >= 8,
    canInspectHandoffLocally: true,
    canRequestLiveBackend: false,
    canExportHandoff: false,
    canSendHandoff: false,
    canPersistHandoff: false,
    canPersistReview: false,
    canPublishWorkflow: false,
    canStartRuntime: false,
    canSubmitConfirmationDecision: false,
    canWriteBusinessTruth: false,
    canReplayRun: false,
    canAttachDatabase: false,
    canEnableRadishAuth: false,
    canImplementRepositoryAdapter: false,
  };
}

function buildHandoffPackageId(source: WorkflowReviewHandoffSource): string {
  return `handoff_${source.workflowWorkspaceReview.applicationId}_${source.workflowWorkspaceReview.scenarioId}`;
}

function buildHandoffNarrative(
  source: WorkflowReviewHandoffSource,
  decisionBlockers: WorkflowReviewHandoffDecisionBlocker[],
  boundaryLocks: WorkflowReviewHandoffBoundaryLock[],
): string {
  return `${source.workflowScenarioInspector.selectedScenario.label} is packaged for human review with ${source.workflowWorkspaceReview.reviewStages.length} review stages, ${decisionBlockers.length} decision blockers, ${source.workflowUserWorkspaceHome.routeEvidence.length} route evidence entries, and ${boundaryLocks.length} locked boundaries.`;
}

function buildRecipients(source: WorkflowReviewHandoffSource): WorkflowReviewHandoffRecipient[] {
  return [
    {
      recipientId: "workflow_owner",
      label: "Workflow owner",
      role: "Application and draft review",
      status: "review_required",
      handoffNeed: "Confirm the selected application, workflow definition, draft, run, and scenario belong together.",
      evidenceRefs: [
        source.workflowWorkspaceReview.applicationId,
        source.workflowWorkspaceReview.workflowDefinitionId,
        source.workflowWorkspaceReview.draftId,
      ],
    },
    {
      recipientId: "policy_reviewer",
      label: "Policy reviewer",
      role: "Risk and confirmation review",
      status: "blocked",
      handoffNeed: "Review the blocked tool action, confirmation placeholder, and human review requirement.",
      evidenceRefs: [
        source.workflowBlockedActionPreview.toolActionId,
        source.workflowConfirmationPlaceholder.confirmationPlaceholderId,
      ],
    },
    {
      recipientId: "runtime_owner",
      label: "Runtime owner",
      role: "Implementation gate review",
      status: "blocked",
      handoffNeed: "Review executor, durable store, auth/store, writeback, and replay gates before any future runtime task.",
      evidenceRefs: source.workflowRuntimeReadinessInspector.implementationGates.map((gate) => gate.gateId).slice(0, 4),
    },
    {
      recipientId: "control_plane_reviewer",
      label: "Control plane reviewer",
      role: "Read-side evidence review",
      status: "offline_only",
      handoffNeed: "Confirm the handoff is backed by read-side routes and does not rely on production API state.",
      evidenceRefs: source.workflowUserWorkspaceHome.routeEvidence.map((evidence) => evidence.evidenceId).slice(0, 4),
    },
  ];
}

function buildKeyFindings(source: WorkflowReviewHandoffSource): WorkflowReviewHandoffFinding[] {
  const scenarioStage = requireStage(source.workflowWorkspaceReview.reviewStages, "stage_scenario_context");
  const validationStage = requireStage(source.workflowWorkspaceReview.reviewStages, "stage_draft_validation");
  const planStage = requireStage(source.workflowWorkspaceReview.reviewStages, "stage_execution_plan");
  const readinessStage = requireStage(source.workflowWorkspaceReview.reviewStages, "stage_runtime_readiness");
  const stopLineStage = requireStage(source.workflowWorkspaceReview.reviewStages, "stage_stop_lines");

  return [
    {
      findingId: "scenario_scope",
      label: "Scenario scope",
      sourceSurface: "scenario",
      status: scenarioStage.status,
      summary: source.workflowScenarioInspector.selectedScenario.intent,
      evidenceRef: source.workflowScenarioInspector.selectedScenario.scenarioId,
      humanReviewQuestion: scenarioStage.reviewQuestion,
    },
    {
      findingId: "review_chain",
      label: "Review chain",
      sourceSurface: "review",
      status: "offline_only",
      summary: source.workflowWorkspaceReview.reviewNarrative,
      evidenceRef: source.workflowWorkspaceReview.pageId,
      humanReviewQuestion: "Does the selected context explain the current application, definition, run, draft, and scenario?",
    },
    {
      findingId: "draft_validation",
      label: "Draft validation",
      sourceSurface: "review",
      status: validationStage.status,
      summary: validationStage.summary,
      evidenceRef: validationStage.primaryRef,
      humanReviewQuestion: validationStage.reviewQuestion,
    },
    {
      findingId: "execution_plan_preview",
      label: "Execution plan preview",
      sourceSurface: "plan",
      status: planStage.status,
      summary: planStage.summary,
      evidenceRef: source.workflowSurfaceOverview.selectedDraftId,
      humanReviewQuestion: planStage.reviewQuestion,
    },
    {
      findingId: "runtime_readiness",
      label: "Runtime readiness",
      sourceSurface: "readiness",
      status: readinessStage.status,
      summary: readinessStage.summary,
      evidenceRef: source.workflowRuntimeReadinessInspector.readinessRouteId,
      humanReviewQuestion: readinessStage.reviewQuestion,
    },
    {
      findingId: "blocked_action",
      label: "Blocked action",
      sourceSurface: "blocked_action",
      status: "blocked",
      summary: source.workflowBlockedActionPreview.policyReason,
      evidenceRef: source.workflowBlockedActionPreview.toolActionId,
      humanReviewQuestion: "Which missing prerequisites prevent this candidate action from executing?",
    },
    {
      findingId: "confirmation_placeholder",
      label: "Confirmation placeholder",
      sourceSurface: "confirmation",
      status: "blocked",
      summary: source.workflowConfirmationPlaceholder.disabledReason,
      evidenceRef: source.workflowConfirmationPlaceholder.confirmationPlaceholderId,
      humanReviewQuestion: "Which decision fields are visible, and why can no decision be submitted?",
    },
    {
      findingId: "stop_lines",
      label: "Stop lines",
      sourceSurface: "stop_line",
      status: stopLineStage.status,
      summary: stopLineStage.summary,
      evidenceRef: `${source.workflowWorkspaceReview.stopLines.length} locked stop lines`,
      humanReviewQuestion: stopLineStage.reviewQuestion,
    },
  ];
}

function buildEvidenceChecklist(source: WorkflowReviewHandoffSource): WorkflowReviewHandoffEvidence[] {
  const routeEvidence = source.workflowUserWorkspaceHome.routeEvidence.map((evidence) =>
    evidenceFromRoute(evidence),
  );

  return [
    ...routeEvidence,
    {
      evidenceId: "review_workspace",
      label: "Review workspace",
      sourceSurface: "review",
      routeOrPageId: source.workflowWorkspaceReview.pageId,
      requestId: source.workflowWorkspaceReview.requestId,
      auditRef: source.workflowWorkspaceReview.auditRef,
      status: source.workflowWorkspaceReview.canRenderWorkspaceReview ? "offline_only" : "blocked",
      summary: "Review workspace supplies selected context, stage order, relations, blockers, and stop lines.",
    },
    {
      evidenceId: "scenario_inspector",
      label: "Scenario inspector",
      sourceSurface: "scenario",
      routeOrPageId: source.workflowScenarioInspector.pageId,
      requestId: source.workflowScenarioInspector.selectedScenarioId,
      auditRef: source.workflowScenarioInspector.relationMap[0]?.auditRef ?? source.workflowScenarioInspector.scenarioMode,
      status: source.workflowScenarioInspector.canRenderScenarioInspector ? "offline_only" : "blocked",
      summary: "Scenario inspector supplies the advisory intent, input contract, expected output, and blocked reasons.",
    },
    {
      evidenceId: "blocked_action_preview",
      label: "Blocked action preview",
      sourceSurface: "blocked_action",
      routeOrPageId: source.workflowBlockedActionPreview.draftRouteId,
      requestId: source.workflowBlockedActionPreview.requestId,
      auditRef: source.workflowBlockedActionPreview.auditRef,
      status: "blocked",
      summary: "Blocked action preview explains the candidate action and missing prerequisites without execution.",
    },
    {
      evidenceId: "confirmation_placeholder",
      label: "Confirmation placeholder",
      sourceSurface: "confirmation",
      routeOrPageId: source.workflowConfirmationPlaceholder.draftRouteId,
      requestId: source.workflowConfirmationPlaceholder.requestId,
      auditRef: source.workflowConfirmationPlaceholder.auditRef,
      status: "blocked",
      summary: "Confirmation placeholder exposes future decision shape without accepting a decision.",
    },
  ];
}

function evidenceFromRoute(evidence: WorkflowUserWorkspaceHomeRouteEvidence): WorkflowReviewHandoffEvidence {
  return {
    evidenceId: `route_${evidence.evidenceId}`,
    label: evidence.label,
    sourceSurface: "home",
    routeOrPageId: evidence.routeId,
    requestId: evidence.requestId,
    auditRef: evidence.auditRef,
    status: evidence.status,
    summary: evidence.summary,
  };
}

function buildDecisionBlockers(source: WorkflowReviewHandoffSource): WorkflowReviewHandoffDecisionBlocker[] {
  const reviewBlockers = source.workflowWorkspaceReview.blockedCapabilityGroups.map((group) =>
    blockerFromReviewGroup(group),
  );
  const runtimeBlockers = source.workflowRuntimeReadinessInspector.readinessBlockers
    .slice(0, 3)
    .map((blocker) => blockerFromRuntimeBlocker(blocker));

  return [...reviewBlockers, ...runtimeBlockers].slice(0, 8);
}

function blockerFromReviewGroup(
  group: WorkflowWorkspaceReviewBlockedCapabilityGroup,
): WorkflowReviewHandoffDecisionBlocker {
  return {
    blockerId: `review_${group.groupId}`,
    label: group.label,
    sourceSurface: group.sourceSurface,
    status: "blocked",
    missingPrerequisite: group.missingPrerequisites.join(", "),
    summary: group.exampleSummary,
    auditRefs: group.auditRefs,
  };
}

function blockerFromRuntimeBlocker(
  blocker: WorkflowRuntimeReadinessBlocker,
): WorkflowReviewHandoffDecisionBlocker {
  return {
    blockerId: `runtime_${blocker.blockerId}`,
    label: blocker.label,
    sourceSurface: blocker.area,
    status: "blocked",
    missingPrerequisite: blocker.missingPrerequisite,
    summary: blocker.summary,
    auditRefs: [blocker.auditRef],
  };
}

function buildBoundaryLocks(source: WorkflowReviewHandoffSource): WorkflowReviewHandoffBoundaryLock[] {
  const stopLineLocks = source.workflowUserWorkspaceHome.stopLines.slice(0, 8).map((stopLine) => ({
    boundaryId: stopLine.stopLineId,
    label: stopLine.label,
    status: "locked" as const,
    summary: stopLine.summary,
  }));
  const explicitLocks: WorkflowReviewHandoffBoundaryLock[] = [
    {
      boundaryId: "handoff_not_persisted",
      label: "Handoff persistence",
      status: "locked",
      summary: "The handoff package is rendered from current offline view models and is not saved or exported.",
    },
    {
      boundaryId: "handoff_no_confirmation_submission",
      label: "Confirmation submission",
      status: "locked",
      summary: "Human review need is visible, but no approve, reject, defer, or submit path is connected.",
    },
    {
      boundaryId: "handoff_no_runtime_unlock",
      label: "Runtime unlock",
      status: "locked",
      summary: "The handoff does not unlock workflow execution, tool execution, writeback, replay, or resume.",
    },
  ];

  return [...stopLineLocks, ...explicitLocks].slice(0, 10);
}

function requireStage(
  stages: WorkflowWorkspaceReviewStage[],
  stageId: string,
): WorkflowWorkspaceReviewStage {
  return stages.find((stage) => stage.stageId === stageId) ?? stages[0]!;
}
