import type { WorkflowApplicationDetailViewModel } from "./workflowApplicationDetail";
import type { WorkflowDefinitionDetailViewModel } from "./workflowDefinitionDetail";
import type { WorkflowDraftDesignerDraft } from "./workflowDraftDesigner";
import type { WorkflowDraftValidationInspectorViewModel } from "./workflowDraftValidationInspector";
import type { WorkflowExecutionPlanPreviewViewModel } from "./workflowExecutionPlanPreview";
import type { WorkflowRuntimeReadinessInspectorViewModel } from "./workflowRuntimeReadinessInspector";
import type { WorkflowRunDetailViewModel } from "./workflowRunDetail";
import type {
  WorkflowScenarioBlockedReason,
  WorkflowScenarioInspectorViewModel,
  WorkflowScenarioStopLine,
} from "./workflowScenarioInspector";
import type {
  WorkflowSurfaceOverviewBlockedCapability,
  WorkflowSurfaceOverviewStatus,
  WorkflowSurfaceOverviewStopLine,
  WorkflowSurfaceOverviewViewModel,
} from "./workflowSurfaceOverview";

export type WorkflowWorkspaceReviewStatus = WorkflowSurfaceOverviewStatus | "locked";

export type WorkflowWorkspaceReviewSource = {
  applicationDetail: WorkflowApplicationDetailViewModel;
  definitionDetail: WorkflowDefinitionDetailViewModel;
  runDetail: WorkflowRunDetailViewModel;
  selectedDraft: WorkflowDraftDesignerDraft;
  validationInspector: WorkflowDraftValidationInspectorViewModel;
  executionPlanPreview: WorkflowExecutionPlanPreviewViewModel;
  runtimeReadinessInspector: WorkflowRuntimeReadinessInspectorViewModel;
  surfaceOverview: WorkflowSurfaceOverviewViewModel;
  scenarioInspector: WorkflowScenarioInspectorViewModel;
};

export type WorkflowWorkspaceReviewContextCard = {
  contextId: string;
  label: string;
  primaryRef: string;
  secondaryRef: string;
  status: WorkflowWorkspaceReviewStatus;
  summary: string;
  auditRef: string;
};

export type WorkflowWorkspaceReviewStage = {
  stageId: string;
  order: number;
  label: string;
  sourceSurface:
    | "scenario"
    | "application"
    | "definition"
    | "draft"
    | "validation"
    | "plan"
    | "readiness"
    | "blocked_capability"
    | "stop_line";
  primaryRef: string;
  status: WorkflowWorkspaceReviewStatus;
  blockedCount: number;
  summary: string;
  reviewQuestion: string;
  auditRef: string;
};

export type WorkflowWorkspaceReviewRelation = {
  relationId: string;
  label: string;
  sourceRef: string;
  targetRef: string;
  status: WorkflowWorkspaceReviewStatus;
  summary: string;
  auditRef: string;
};

export type WorkflowWorkspaceReviewBlockedCapabilityGroup = {
  groupId: string;
  label: string;
  sourceSurface: string;
  status: "blocked";
  count: number;
  missingPrerequisites: string[];
  exampleSummary: string;
  auditRefs: string[];
};

export type WorkflowWorkspaceReviewStopLine = {
  stopLineId: string;
  label: string;
  sourceSurface: "overview" | "scenario";
  status: "locked";
  summary: string;
};

export type WorkflowWorkspaceReviewViewModel = {
  pageId: "workflow-workspace-review-offline";
  sourcePageIds: string[];
  reviewMode: "offline_read_only_advisory";
  applicationId: string;
  workflowDefinitionId: string;
  runId: string;
  draftId: string;
  scenarioId: string;
  requestId: string;
  auditRef: string;
  reviewNarrative: string;
  contextCards: WorkflowWorkspaceReviewContextCard[];
  reviewStages: WorkflowWorkspaceReviewStage[];
  relationMap: WorkflowWorkspaceReviewRelation[];
  blockedCapabilityGroups: WorkflowWorkspaceReviewBlockedCapabilityGroup[];
  stopLines: WorkflowWorkspaceReviewStopLine[];
  canRenderWorkspaceReview: boolean;
  canInspectWorkspaceLocally: true;
  canRequestLiveBackend: false;
  canMutateBuilder: false;
  canPersistReview: false;
  canPersistDraft: false;
  canPersistExecutionPlan: false;
  canPersistRuntimeReadiness: false;
  canPublishWorkflow: false;
  canStartRuntime: false;
  canSubmitConfirmationDecision: false;
  canWriteBusinessTruth: false;
  canReplayRun: false;
  canAttachDatabase: false;
  canEnableRadishOidc: false;
  canImplementRepositoryAdapter: false;
};

export function buildWorkflowWorkspaceReviewViewModel(
  source: WorkflowWorkspaceReviewSource,
): WorkflowWorkspaceReviewViewModel {
  const blockedCapabilityGroups = buildBlockedCapabilityGroups(source);
  const stopLines = buildStopLines(source.surfaceOverview.stopLines, source.scenarioInspector.stopLines);
  const contextCards = buildContextCards(source);
  const reviewStages = buildReviewStages(source, blockedCapabilityGroups, stopLines);
  const relationMap = buildRelationMap(source, blockedCapabilityGroups, stopLines);

  return {
    pageId: "workflow-workspace-review-offline",
    sourcePageIds: [
      source.applicationDetail.pageId,
      source.definitionDetail.pageId,
      source.runDetail.pageId,
      "workflow-draft-designer-offline",
      source.validationInspector.pageId,
      source.executionPlanPreview.pageId,
      source.runtimeReadinessInspector.pageId,
      source.surfaceOverview.pageId,
      source.scenarioInspector.pageId,
    ],
    reviewMode: "offline_read_only_advisory",
    applicationId: source.applicationDetail.applicationId,
    workflowDefinitionId: source.definitionDetail.workflowDefinitionId,
    runId: source.runDetail.runId,
    draftId: source.selectedDraft.draftId,
    scenarioId: source.scenarioInspector.selectedScenarioId,
    requestId: source.executionPlanPreview.requestId,
    auditRef: source.runtimeReadinessInspector.auditRef,
    reviewNarrative: buildReviewNarrative(source, blockedCapabilityGroups, stopLines),
    contextCards,
    reviewStages,
    relationMap,
    blockedCapabilityGroups,
    stopLines,
    canRenderWorkspaceReview:
      source.applicationDetail.canRenderApplicationDetail &&
      source.definitionDetail.canRenderDefinitionDetail &&
      source.runDetail.canRenderRunDetail &&
      source.validationInspector.canRenderDraftValidationInspector &&
      source.executionPlanPreview.canRenderExecutionPlanPreview &&
      source.runtimeReadinessInspector.canRenderRuntimeReadinessInspector &&
      source.surfaceOverview.canRenderSurfaceOverview &&
      source.scenarioInspector.canRenderScenarioInspector &&
      contextCards.length === 5 &&
      reviewStages.length >= 8 &&
      relationMap.length >= 8 &&
      blockedCapabilityGroups.length >= 5 &&
      stopLines.length >= 10,
    canInspectWorkspaceLocally: true,
    canRequestLiveBackend: false,
    canMutateBuilder: false,
    canPersistReview: false,
    canPersistDraft: false,
    canPersistExecutionPlan: false,
    canPersistRuntimeReadiness: false,
    canPublishWorkflow: false,
    canStartRuntime: false,
    canSubmitConfirmationDecision: false,
    canWriteBusinessTruth: false,
    canReplayRun: false,
    canAttachDatabase: false,
    canEnableRadishOidc: false,
    canImplementRepositoryAdapter: false,
  };
}

function buildReviewNarrative(
  source: WorkflowWorkspaceReviewSource,
  blockedCapabilityGroups: WorkflowWorkspaceReviewBlockedCapabilityGroup[],
  stopLines: WorkflowWorkspaceReviewStopLine[],
): string {
  return `${source.scenarioInspector.selectedScenario.label} is reviewed against ${source.selectedDraft.draftId}, ${source.executionPlanPreview.stageOrder.length} preview stages, ${source.runtimeReadinessInspector.readinessBlockers.length} readiness blockers, ${blockedCapabilityGroups.length} blocked capability groups, and ${stopLines.length} locked stop lines.`;
}

function buildContextCards(source: WorkflowWorkspaceReviewSource): WorkflowWorkspaceReviewContextCard[] {
  const selectedScenario = source.scenarioInspector.selectedScenario;

  return [
    {
      contextId: "selected_application",
      label: "Selected application",
      primaryRef: source.applicationDetail.displayName,
      secondaryRef: source.applicationDetail.applicationId,
      status: source.applicationDetail.canRenderApplicationDetail ? "ready" : "blocked",
      summary: `${source.applicationDetail.providerProfileRef} is visible as read-side metadata; application mutation and runtime start stay blocked.`,
      auditRef: source.applicationDetail.auditRef,
    },
    {
      contextId: "selected_definition",
      label: "Workflow definition",
      primaryRef: source.definitionDetail.workflowDefinitionId,
      secondaryRef: `v${source.definitionDetail.version} / ${source.definitionDetail.riskLevel}`,
      status:
        source.applicationDetail.latestWorkflowDefinitionRef === source.definitionDetail.workflowDefinitionId
          ? "ready"
          : "review_required",
      summary: `${source.definitionDetail.nodes.length} nodes and ${source.definitionDetail.edges.length} edges define the current read-only workflow shape.`,
      auditRef: source.definitionDetail.auditRef,
    },
    {
      contextId: "selected_run",
      label: "Run record",
      primaryRef: source.runDetail.runId,
      secondaryRef: `${source.runDetail.status} / ${source.runDetail.failureCode}`,
      status: source.applicationDetail.latestRunRef === source.runDetail.runId ? "ready" : "review_required",
      summary: `The run is a sanitized existing record with trace ${source.runDetail.traceId}; replay and materialized result reader are blocked.`,
      auditRef: source.runDetail.auditRef,
    },
    {
      contextId: "selected_draft",
      label: "Draft",
      primaryRef: source.selectedDraft.draftId,
      secondaryRef: `${source.selectedDraft.nodes.length} nodes / ${source.selectedDraft.edges.length} edges`,
      status: "offline_only",
      summary: "The selected draft can be inspected and switched locally, but cannot be saved, published, executed, or persisted.",
      auditRef: source.selectedDraft.routeMetadata.auditRef,
    },
    {
      contextId: "selected_scenario",
      label: "Scenario",
      primaryRef: selectedScenario.label,
      secondaryRef: selectedScenario.scenarioKind,
      status: selectedScenario.requiresConfirmation ? "review_required" : "offline_only",
      summary: selectedScenario.intent,
      auditRef: source.scenarioInspector.relationMap[0]?.auditRef ?? source.scenarioInspector.selectedScenarioId,
    },
  ];
}

function buildReviewStages(
  source: WorkflowWorkspaceReviewSource,
  blockedCapabilityGroups: WorkflowWorkspaceReviewBlockedCapabilityGroup[],
  stopLines: WorkflowWorkspaceReviewStopLine[],
): WorkflowWorkspaceReviewStage[] {
  const selectedScenario = source.scenarioInspector.selectedScenario;

  return [
    {
      stageId: "stage_scenario_context",
      order: 1,
      label: "Scenario context",
      sourceSurface: "scenario",
      primaryRef: selectedScenario.scenarioId,
      status: selectedScenario.requiresConfirmation ? "review_required" : "offline_only",
      blockedCount: source.scenarioInspector.blockedReasons.length,
      summary: selectedScenario.triggerSummary,
      reviewQuestion: "What user-facing scenario is being reviewed in this workspace?",
      auditRef: source.scenarioInspector.relationMap[0]?.auditRef ?? source.runtimeReadinessInspector.auditRef,
    },
    {
      stageId: "stage_application_definition",
      order: 2,
      label: "Application and definition",
      sourceSurface: "application",
      primaryRef: source.applicationDetail.applicationId,
      status:
        source.applicationDetail.latestWorkflowDefinitionRef === source.definitionDetail.workflowDefinitionId
          ? "ready"
          : "review_required",
      blockedCount: source.applicationDetail.blockedCapabilities.length,
      summary: "Application detail anchors the selected workflow definition without exposing lifecycle mutation.",
      reviewQuestion: "Does the selected application point at the definition under review?",
      auditRef: source.applicationDetail.auditRef,
    },
    {
      stageId: "stage_definition_to_draft",
      order: 3,
      label: "Definition to draft",
      sourceSurface: "draft",
      primaryRef: source.selectedDraft.draftId,
      status: "offline_only",
      blockedCount: source.selectedDraft.blockedCapabilities.length,
      summary: "Draft nodes and edges are derived from offline definition metadata and stay inspect-only.",
      reviewQuestion: "Which draft carries the scenario through the workflow graph?",
      auditRef: source.selectedDraft.routeMetadata.auditRef,
    },
    {
      stageId: "stage_draft_validation",
      order: 4,
      label: "Draft validation",
      sourceSurface: "validation",
      primaryRef: source.validationInspector.inspectedDraftId,
      status: validationStatusToReviewStatus(source.validationInspector.validationStatus),
      blockedCount: source.validationInspector.blockedCapabilityChecks.length,
      summary: "Validation explains structure, contract fields, and blocked capability checks before runtime work.",
      reviewQuestion: "Which checks still need human review or future implementation gates?",
      auditRef: source.validationInspector.auditRef,
    },
    {
      stageId: "stage_execution_plan",
      order: 5,
      label: "Execution plan preview",
      sourceSurface: "plan",
      primaryRef: source.executionPlanPreview.selectedDraftId,
      status: source.executionPlanPreview.canRenderExecutionPlanPreview ? "offline_only" : "blocked",
      blockedCount: source.executionPlanPreview.blockedPlanReasons.length,
      summary: "Plan preview orders future stages, provider requirements, and confirmation/audit gates without persistence.",
      reviewQuestion: "How would the draft be ordered if a future executor existed?",
      auditRef: source.executionPlanPreview.auditRef,
    },
    {
      stageId: "stage_runtime_readiness",
      order: 6,
      label: "Runtime readiness",
      sourceSurface: "readiness",
      primaryRef: source.runtimeReadinessInspector.readinessRouteId,
      status: "blocked",
      blockedCount: source.runtimeReadinessInspector.readinessBlockers.length,
      summary: "Readiness keeps executor, durable store, confirmation, auth/store, writeback, and replay gates blocked.",
      reviewQuestion: "Why can this workflow not start, persist, confirm, write back, or replay?",
      auditRef: source.runtimeReadinessInspector.auditRef,
    },
    {
      stageId: "stage_blocked_capability_rollup",
      order: 7,
      label: "Blocked capability rollup",
      sourceSurface: "blocked_capability",
      primaryRef: `${blockedCapabilityGroups.length} groups`,
      status: "blocked",
      blockedCount: blockedCapabilityGroups.reduce((total, group) => total + group.count, 0),
      summary: "Blocked capabilities are grouped across application, draft, plan, readiness, run, and scenario evidence.",
      reviewQuestion: "Which missing prerequisites explain every disabled action path?",
      auditRef: source.surfaceOverview.auditRef,
    },
    {
      stageId: "stage_stop_lines",
      order: 8,
      label: "Stop lines",
      sourceSurface: "stop_line",
      primaryRef: `${stopLines.length} locked stop lines`,
      status: "locked",
      blockedCount: stopLines.length,
      summary: "Stop lines close the review by keeping live backend, mutation, execution, persistence, auth/db, and replay out of scope.",
      reviewQuestion: "Which boundaries must remain locked before a future implementation task can start?",
      auditRef: source.runtimeReadinessInspector.auditRef,
    },
  ];
}

function buildRelationMap(
  source: WorkflowWorkspaceReviewSource,
  blockedCapabilityGroups: WorkflowWorkspaceReviewBlockedCapabilityGroup[],
  stopLines: WorkflowWorkspaceReviewStopLine[],
): WorkflowWorkspaceReviewRelation[] {
  const selectedScenario = source.scenarioInspector.selectedScenario;

  return [
    {
      relationId: "scenario_to_application",
      label: "Scenario to application",
      sourceRef: selectedScenario.scenarioId,
      targetRef: source.applicationDetail.applicationId,
      status: selectedScenario.applicationRef === source.applicationDetail.applicationId ? "ready" : "review_required",
      summary: "The selected scenario is interpreted inside the currently selected application context.",
      auditRef: source.applicationDetail.auditRef,
    },
    {
      relationId: "application_to_definition",
      label: "Application to definition",
      sourceRef: source.applicationDetail.applicationId,
      targetRef: source.definitionDetail.workflowDefinitionId,
      status:
        source.applicationDetail.latestWorkflowDefinitionRef === source.definitionDetail.workflowDefinitionId
          ? "ready"
          : "review_required",
      summary: "Application detail points at the workflow definition under review while lifecycle writes stay unavailable.",
      auditRef: source.applicationDetail.auditRef,
    },
    {
      relationId: "definition_to_draft",
      label: "Definition to draft",
      sourceRef: source.definitionDetail.workflowDefinitionId,
      targetRef: source.selectedDraft.draftId,
      status: "offline_only",
      summary: "The selected draft projects definition nodes and edges into an inspect-only workspace view.",
      auditRef: source.selectedDraft.routeMetadata.auditRef,
    },
    {
      relationId: "draft_to_validation",
      label: "Draft to validation",
      sourceRef: source.selectedDraft.draftId,
      targetRef: source.validationInspector.draftRouteId,
      status: validationStatusToReviewStatus(source.validationInspector.validationStatus),
      summary: "Validation keeps structural checks, contract checks, and blocked capability checks beside the draft.",
      auditRef: source.validationInspector.auditRef,
    },
    {
      relationId: "validation_to_plan",
      label: "Validation to plan",
      sourceRef: source.validationInspector.inspectedDraftId,
      targetRef: source.executionPlanPreview.draftRouteId,
      status: source.executionPlanPreview.canRenderExecutionPlanPreview ? "offline_only" : "blocked",
      summary: "Execution plan preview consumes validation evidence without creating a persistent executable plan.",
      auditRef: source.executionPlanPreview.auditRef,
    },
    {
      relationId: "plan_to_readiness",
      label: "Plan to readiness",
      sourceRef: source.executionPlanPreview.draftRouteId,
      targetRef: source.runtimeReadinessInspector.readinessRouteId,
      status: "blocked",
      summary: "Runtime readiness explains why the previewed plan cannot start or persist.",
      auditRef: source.runtimeReadinessInspector.auditRef,
    },
    {
      relationId: "readiness_to_blocked_capabilities",
      label: "Readiness to blocked capabilities",
      sourceRef: source.runtimeReadinessInspector.readinessRouteId,
      targetRef: `${blockedCapabilityGroups.length} capability groups`,
      status: "blocked",
      summary: "Readiness blockers are rolled up with application, draft, plan, run, and scenario blockers.",
      auditRef: source.runtimeReadinessInspector.auditRef,
    },
    {
      relationId: "blocked_capabilities_to_stop_lines",
      label: "Blocked capabilities to stop lines",
      sourceRef: `${blockedCapabilityGroups.length} capability groups`,
      targetRef: `${stopLines.length} stop lines`,
      status: "locked",
      summary: "Every disabled path is closed by explicit offline-only, no-mutation, no-runtime, no-writeback, or auth/store stop lines.",
      auditRef: source.surfaceOverview.auditRef,
    },
  ];
}

function buildBlockedCapabilityGroups(
  source: WorkflowWorkspaceReviewSource,
): WorkflowWorkspaceReviewBlockedCapabilityGroup[] {
  const candidates = [
    ...source.surfaceOverview.blockedCapabilities.map((capability) =>
      blockedCapabilityFromOverview(capability),
    ),
    ...source.scenarioInspector.blockedReasons.map((reason) => blockedCapabilityFromScenario(reason)),
  ];
  const grouped = new Map<string, WorkflowWorkspaceReviewBlockedCapabilityGroup>();

  for (const candidate of candidates) {
    const existing = grouped.get(candidate.sourceSurface);
    if (!existing) {
      grouped.set(candidate.sourceSurface, {
        groupId: `blocked_${candidate.sourceSurface}`,
        label: `${candidate.sourceSurface} blockers`,
        sourceSurface: candidate.sourceSurface,
        status: "blocked",
        count: 1,
        missingPrerequisites: [candidate.missingPrerequisite],
        exampleSummary: candidate.summary,
        auditRefs: [candidate.auditRef],
      });
      continue;
    }

    existing.count += 1;
    existing.missingPrerequisites = appendUnique(existing.missingPrerequisites, candidate.missingPrerequisite).slice(0, 4);
    existing.auditRefs = appendUnique(existing.auditRefs, candidate.auditRef).slice(0, 4);
  }

  return [...grouped.values()];
}

function blockedCapabilityFromOverview(capability: WorkflowSurfaceOverviewBlockedCapability) {
  return {
    sourceSurface: capability.sourceSurface,
    missingPrerequisite: capability.missingPrerequisite,
    summary: capability.summary,
    auditRef: capability.auditRef,
  };
}

function blockedCapabilityFromScenario(reason: WorkflowScenarioBlockedReason) {
  return {
    sourceSurface: reason.sourceSurface,
    missingPrerequisite: reason.missingPrerequisite,
    summary: reason.summary,
    auditRef: reason.auditRef,
  };
}

function buildStopLines(
  overviewStopLines: WorkflowSurfaceOverviewStopLine[],
  scenarioStopLines: WorkflowScenarioStopLine[],
): WorkflowWorkspaceReviewStopLine[] {
  return [
    ...overviewStopLines.map((stopLine) => ({
      stopLineId: `overview_${stopLine.stopLineId}`,
      label: stopLine.label,
      sourceSurface: "overview" as const,
      status: "locked" as const,
      summary: stopLine.summary,
    })),
    ...scenarioStopLines.map((stopLine) => ({
      stopLineId: `scenario_${stopLine.stopLineId}`,
      label: stopLine.label,
      sourceSurface: "scenario" as const,
      status: "locked" as const,
      summary: stopLine.summary,
    })),
  ];
}

function validationStatusToReviewStatus(status: string): WorkflowWorkspaceReviewStatus {
  if (status === "passed") {
    return "ready";
  }
  if (status === "blocked") {
    return "blocked";
  }
  return "review_required";
}

function appendUnique(values: string[], nextValue: string): string[] {
  if (values.includes(nextValue)) {
    return values;
  }
  return [...values, nextValue];
}
