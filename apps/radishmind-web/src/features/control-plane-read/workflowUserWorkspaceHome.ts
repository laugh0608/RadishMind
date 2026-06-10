import type { WorkspaceApiKeysViewModel } from "./workspaceApiKeys";
import type { WorkspaceApplicationsViewModel } from "./workspaceApplications";
import type { WorkspaceRunHistoryViewModel } from "./workspaceRunHistory";
import type { WorkspaceUsageQuotaViewModel } from "./workspaceUsageQuota";
import type { WorkspaceWorkflowDefinitionsViewModel } from "./workspaceWorkflowDefinitions";
import type { WorkflowScenarioInspectorViewModel } from "./workflowScenarioInspector";
import type {
  WorkflowSurfaceOverviewMetric,
  WorkflowSurfaceOverviewStatus,
  WorkflowSurfaceOverviewViewModel,
} from "./workflowSurfaceOverview";
import type {
  WorkflowWorkspaceReviewContextCard,
  WorkflowWorkspaceReviewStage,
  WorkflowWorkspaceReviewStatus,
  WorkflowWorkspaceReviewStopLine,
  WorkflowWorkspaceReviewViewModel,
} from "./workflowWorkspaceReview";

export type WorkflowUserWorkspaceHomeStatus =
  | WorkflowSurfaceOverviewStatus
  | WorkflowWorkspaceReviewStatus
  | "locked";

export type WorkflowUserWorkspaceHomeSource = {
  workspaceApplications: WorkspaceApplicationsViewModel;
  workspaceApiKeys: WorkspaceApiKeysViewModel;
  workspaceUsageQuota: WorkspaceUsageQuotaViewModel;
  workspaceWorkflowDefinitions: WorkspaceWorkflowDefinitionsViewModel;
  workspaceRunHistory: WorkspaceRunHistoryViewModel;
  workflowWorkspaceReview: WorkflowWorkspaceReviewViewModel;
  workflowSurfaceOverview: WorkflowSurfaceOverviewViewModel;
  workflowScenarioInspector: WorkflowScenarioInspectorViewModel;
};

export type WorkflowUserWorkspaceHomeMetric = {
  metricId: string;
  label: string;
  value: string;
  status: WorkflowUserWorkspaceHomeStatus;
  summary: string;
};

export type WorkflowUserWorkspaceHomeApplication = {
  applicationRef: string;
  displayName: string;
  applicationKind: string;
  workflowDefinitionId: string;
  latestRunStatus: string;
  selected: boolean;
  status: WorkflowUserWorkspaceHomeStatus;
  summary: string;
  auditRef: string;
};

export type WorkflowUserWorkspaceHomeRun = {
  runId: string;
  applicationRef: string;
  workflowDefinitionId: string;
  status: WorkflowUserWorkspaceHomeStatus;
  failureCode: string;
  cost: string;
  traceId: string;
  selected: boolean;
  summary: string;
};

export type WorkflowUserWorkspaceHomeReadiness = {
  readinessId: string;
  label: string;
  value: string;
  status: WorkflowUserWorkspaceHomeStatus;
  sourceSurface: "overview" | "review" | "scenario";
  summary: string;
  auditRef: string;
};

export type WorkflowUserWorkspaceHomeRouteEvidence = {
  evidenceId: string;
  label: string;
  routeId: string;
  requestId: string;
  auditRef: string;
  status: WorkflowUserWorkspaceHomeStatus;
  summary: string;
};

export type WorkflowUserWorkspaceHomeViewModel = {
  pageId: "workflow-user-workspace-home-offline";
  sourcePageIds: string[];
  homeMode: "offline_read_only_advisory";
  tenantRef: string;
  applicationId: string;
  workflowDefinitionId: string;
  runId: string;
  draftId: string;
  scenarioId: string;
  requestId: string;
  auditRef: string;
  homeNarrative: string;
  metrics: WorkflowUserWorkspaceHomeMetric[];
  applicationPortfolio: WorkflowUserWorkspaceHomeApplication[];
  currentReviewStages: WorkflowWorkspaceReviewStage[];
  selectedContext: WorkflowWorkspaceReviewContextCard[];
  readinessRollup: WorkflowUserWorkspaceHomeReadiness[];
  recentRuns: WorkflowUserWorkspaceHomeRun[];
  routeEvidence: WorkflowUserWorkspaceHomeRouteEvidence[];
  stopLines: WorkflowWorkspaceReviewStopLine[];
  canRenderUserWorkspaceHome: boolean;
  canInspectWorkspaceLocally: true;
  canRequestLiveBackend: false;
  canMutateApplications: false;
  canCreateApiKey: false;
  canEnforceQuota: false;
  canMutateBuilder: false;
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

export function buildWorkflowUserWorkspaceHomeViewModel(
  source: WorkflowUserWorkspaceHomeSource,
): WorkflowUserWorkspaceHomeViewModel {
  const applicationPortfolio = buildApplicationPortfolio(source);
  const recentRuns = buildRecentRuns(source);
  const readinessRollup = buildReadinessRollup(source);
  const routeEvidence = buildRouteEvidence(source);
  const stopLines = source.workflowWorkspaceReview.stopLines.slice(0, 8);
  const currentReviewStages = source.workflowWorkspaceReview.reviewStages.slice(0, 4);
  const selectedContext = source.workflowWorkspaceReview.contextCards;

  return {
    pageId: "workflow-user-workspace-home-offline",
    sourcePageIds: [
      source.workspaceApplications.pageId,
      source.workspaceApiKeys.pageId,
      source.workspaceUsageQuota.pageId,
      source.workspaceWorkflowDefinitions.pageId,
      source.workspaceRunHistory.pageId,
      source.workflowSurfaceOverview.pageId,
      source.workflowScenarioInspector.pageId,
      source.workflowWorkspaceReview.pageId,
    ],
    homeMode: "offline_read_only_advisory",
    tenantRef: source.workspaceApplications.collection.tenantRef,
    applicationId: source.workflowWorkspaceReview.applicationId,
    workflowDefinitionId: source.workflowWorkspaceReview.workflowDefinitionId,
    runId: source.workflowWorkspaceReview.runId,
    draftId: source.workflowWorkspaceReview.draftId,
    scenarioId: source.workflowWorkspaceReview.scenarioId,
    requestId: source.workflowWorkspaceReview.requestId,
    auditRef: source.workflowWorkspaceReview.auditRef,
    homeNarrative: buildHomeNarrative(source),
    metrics: buildMetrics(source),
    applicationPortfolio,
    currentReviewStages,
    selectedContext,
    readinessRollup,
    recentRuns,
    routeEvidence,
    stopLines,
    canRenderUserWorkspaceHome:
      source.workspaceApplications.canRenderApplications &&
      source.workspaceApiKeys.canRenderApiKeys &&
      source.workspaceUsageQuota.canRenderQuota &&
      source.workspaceWorkflowDefinitions.canRenderWorkflowDefinitions &&
      source.workspaceRunHistory.canRenderRuns &&
      source.workflowSurfaceOverview.canRenderSurfaceOverview &&
      source.workflowScenarioInspector.canRenderScenarioInspector &&
      source.workflowWorkspaceReview.canRenderWorkspaceReview &&
      applicationPortfolio.length > 0 &&
      recentRuns.length > 0 &&
      readinessRollup.length >= 6 &&
      routeEvidence.length >= 5 &&
      stopLines.length >= 6,
    canInspectWorkspaceLocally: true,
    canRequestLiveBackend: false,
    canMutateApplications: false,
    canCreateApiKey: false,
    canEnforceQuota: false,
    canMutateBuilder: false,
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

function buildHomeNarrative(source: WorkflowUserWorkspaceHomeSource): string {
  return `${source.workspaceApplications.applications.length} applications, ${source.workspaceWorkflowDefinitions.workflowDefinitions.length} workflow definitions, ${source.workspaceRunHistory.runs.length} recent runs, ${source.workflowWorkspaceReview.blockedCapabilityGroups.length} blocked capability groups, and ${source.workflowWorkspaceReview.stopLines.length} locked stop lines are summarized for the selected workspace context.`;
}

function buildMetrics(source: WorkflowUserWorkspaceHomeSource): WorkflowUserWorkspaceHomeMetric[] {
  const activeApiKeys = source.workspaceApiKeys.apiKeys.filter((apiKey) => apiKey.state === "active").length;
  const failedRuns = source.workspaceRunHistory.runs.filter((run) => run.failureCode !== "none").length;
  const blockedGroups = source.workflowWorkspaceReview.blockedCapabilityGroups.length;

  return [
    {
      metricId: "applications",
      label: "Applications",
      value: String(source.workspaceApplications.applications.length),
      status: source.workspaceApplications.canRenderApplications ? "ready" : "blocked",
      summary: "Application summaries anchor the user workspace without lifecycle mutation.",
    },
    {
      metricId: "workflow_definitions",
      label: "Workflow definitions",
      value: String(source.workspaceWorkflowDefinitions.workflowDefinitions.length),
      status: source.workspaceWorkflowDefinitions.canRenderWorkflowDefinitions ? "ready" : "blocked",
      summary: "Definition summaries stay read-only and point into the workflow review context.",
    },
    {
      metricId: "api_keys",
      label: "API keys",
      value: `${activeApiKeys}/${source.workspaceApiKeys.apiKeys.length}`,
      status: source.workspaceApiKeys.canRenderApiKeys ? "offline_only" : "blocked",
      summary: "Key state is shown without key value, secret material, lifecycle mutation, or gateway enforcement.",
    },
    {
      metricId: "quota",
      label: "Quota",
      value: quotaSummary(source.workspaceUsageQuota),
      status: source.workspaceUsageQuota.canRenderQuota ? "offline_only" : "blocked",
      summary: "Usage and quota are summarized without enforcement, rate limit, cost record writes, or cost writeback.",
    },
    {
      metricId: "recent_runs",
      label: "Recent runs",
      value: `${source.workspaceRunHistory.runs.length} / ${failedRuns} failed`,
      status: failedRuns > 0 ? "review_required" : "ready",
      summary: "Run status, trace, failure code, and cost summary are visible while replay and resume stay locked.",
    },
    {
      metricId: "blocked_capabilities",
      label: "Blocked capabilities",
      value: String(blockedGroups),
      status: "blocked",
      summary: "Review blockers explain why publish, execution, confirmation, writeback, and replay are unavailable.",
    },
    {
      metricId: "stop_lines",
      label: "Stop lines",
      value: String(source.workflowWorkspaceReview.stopLines.length),
      status: "locked",
      summary: "Locked stop lines keep live backend, persistence, auth, database, executor, and replay out of scope.",
    },
  ];
}

function buildApplicationPortfolio(
  source: WorkflowUserWorkspaceHomeSource,
): WorkflowUserWorkspaceHomeApplication[] {
  return source.workspaceApplications.applications.map((application) => ({
    applicationRef: application.applicationRef,
    displayName: application.displayName,
    applicationKind: application.applicationKind,
    workflowDefinitionId: application.latestWorkflowDefinitionRef,
    latestRunStatus: application.lastRunStatus,
    selected: application.applicationRef === source.workflowWorkspaceReview.applicationId,
    status:
      application.applicationRef === source.workflowWorkspaceReview.applicationId
        ? "ready"
        : application.lastRunStatus === "blocked"
          ? "review_required"
          : "offline_only",
    summary: `${application.ownerSubjectRef} owns the application; latest workflow ${application.latestWorkflowDefinitionRef} is available as read-side context.`,
    auditRef: source.workspaceApplications.auditRef,
  }));
}

function buildRecentRuns(source: WorkflowUserWorkspaceHomeSource): WorkflowUserWorkspaceHomeRun[] {
  return source.workspaceRunHistory.runs.slice(0, 4).map((run) => ({
    runId: run.runId,
    applicationRef: run.applicationRef,
    workflowDefinitionId: run.workflowDefinitionId,
    status:
      run.runId === source.workflowWorkspaceReview.runId
        ? "ready"
        : run.failureCode !== "none"
          ? "review_required"
          : "offline_only",
    failureCode: run.failureCode,
    cost: run.estimatedCost,
    traceId: run.traceId,
    selected: run.runId === source.workflowWorkspaceReview.runId,
    summary: `${run.status} run started ${run.startedAt}; replay, resume, and materialized result reader remain disabled.`,
  }));
}

function buildReadinessRollup(source: WorkflowUserWorkspaceHomeSource): WorkflowUserWorkspaceHomeReadiness[] {
  const overviewItems = source.workflowSurfaceOverview.summary.map((metric) =>
    readinessFromOverviewMetric(metric, source.workflowSurfaceOverview.auditRef),
  );
  const scenarioItems = source.workflowScenarioInspector.summary.slice(0, 3).map((summary) => ({
    readinessId: `scenario_${summary.label.toLowerCase().replaceAll(" ", "_")}`,
    label: summary.label,
    value: summary.value,
    status: summary.status,
    sourceSurface: "scenario" as const,
    summary: summary.summary,
    auditRef: source.workflowScenarioInspector.relationMap[0]?.auditRef ?? source.workflowScenarioInspector.scenarioMode,
  }));
  const reviewItems = source.workflowWorkspaceReview.blockedCapabilityGroups.slice(0, 3).map((group) => ({
    readinessId: `review_${group.groupId}`,
    label: group.label,
    value: `${group.count} blocked`,
    status: "blocked" as const,
    sourceSurface: "review" as const,
    summary: group.exampleSummary,
    auditRef: group.auditRefs[0] ?? source.workflowWorkspaceReview.auditRef,
  }));

  return [...overviewItems, ...scenarioItems, ...reviewItems];
}

function readinessFromOverviewMetric(
  metric: WorkflowSurfaceOverviewMetric,
  auditRef: string,
): WorkflowUserWorkspaceHomeReadiness {
  return {
    readinessId: `overview_${metric.metricId}`,
    label: metric.label,
    value: metric.value,
    status: metric.status,
    sourceSurface: "overview",
    summary: metric.summary,
    auditRef,
  };
}

function buildRouteEvidence(source: WorkflowUserWorkspaceHomeSource): WorkflowUserWorkspaceHomeRouteEvidence[] {
  return [
    {
      evidenceId: "applications_route",
      label: "Applications",
      routeId: source.workspaceApplications.routeId,
      requestId: source.workspaceApplications.requestId,
      auditRef: source.workspaceApplications.auditRef,
      status: source.workspaceApplications.canRenderApplications ? "ready" : "blocked",
      summary: "Application summary list is consumed as read-side workspace context.",
    },
    {
      evidenceId: "workflow_definitions_route",
      label: "Workflow definitions",
      routeId: source.workspaceWorkflowDefinitions.routeId,
      requestId: source.workspaceWorkflowDefinitions.requestId,
      auditRef: source.workspaceWorkflowDefinitions.auditRef,
      status: source.workspaceWorkflowDefinitions.canRenderWorkflowDefinitions ? "ready" : "blocked",
      summary: "Definition summaries connect applications to workflow detail and draft inspection.",
    },
    {
      evidenceId: "run_history_route",
      label: "Run history",
      routeId: source.workspaceRunHistory.routeId,
      requestId: source.workspaceRunHistory.requestId,
      auditRef: source.workspaceRunHistory.auditRef,
      status: source.workspaceRunHistory.canRenderRuns ? "ready" : "blocked",
      summary: "Run history supplies status, trace, failure code, and cost summary without executor access.",
    },
    {
      evidenceId: "api_keys_route",
      label: "API keys",
      routeId: source.workspaceApiKeys.routeId,
      requestId: source.workspaceApiKeys.requestId,
      auditRef: source.workspaceApiKeys.auditRef,
      status: source.workspaceApiKeys.canRenderApiKeys ? "offline_only" : "blocked",
      summary: "API key summaries expose scope and state without key values or lifecycle mutation.",
    },
    {
      evidenceId: "quota_route",
      label: "Usage quota",
      routeId: source.workspaceUsageQuota.routeId,
      requestId: source.workspaceUsageQuota.requestId,
      auditRef: source.workspaceUsageQuota.auditRef,
      status: source.workspaceUsageQuota.canRenderQuota ? "offline_only" : "blocked",
      summary: "Quota summary is displayed without enforcement, rate limit, or cost record writes.",
    },
    {
      evidenceId: "review_workspace",
      label: "Workflow review",
      routeId: source.workflowWorkspaceReview.pageId,
      requestId: source.workflowWorkspaceReview.requestId,
      auditRef: source.workflowWorkspaceReview.auditRef,
      status: source.workflowWorkspaceReview.canRenderWorkspaceReview ? "offline_only" : "blocked",
      summary: "Review workspace rolls up context, stages, blockers, and stop lines for the selected scenario.",
    },
  ];
}

function quotaSummary(usageQuota: WorkspaceUsageQuotaViewModel): string {
  const costLimit = usageQuota.limits.find((limit) => limit.label === "Cost");
  if (costLimit) {
    return `${costLimit.used} / ${costLimit.value}`;
  }
  return usageQuota.quota?.quota_id ?? "not available";
}
