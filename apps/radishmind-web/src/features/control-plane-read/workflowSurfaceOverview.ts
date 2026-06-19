import type { WorkflowApplicationDetailViewModel } from "./workflowApplicationDetail";
import type { WorkflowDefinitionDetailViewModel } from "./workflowDefinitionDetail";
import type { WorkflowDraftDesignerDraft } from "./workflowDraftDesigner";
import type { WorkflowDraftValidationInspectorViewModel } from "./workflowDraftValidationInspector";
import type { WorkflowExecutionPlanPreviewViewModel } from "./workflowExecutionPlanPreview";
import type { WorkflowRuntimeReadinessInspectorViewModel } from "./workflowRuntimeReadinessInspector";
import type { WorkflowRunDetailViewModel } from "./workflowRunDetail";

export type WorkflowSurfaceOverviewStatus = "ready" | "offline_only" | "review_required" | "blocked";

export type WorkflowSurfaceOverviewMetric = {
  metricId: string;
  label: string;
  value: string;
  status: WorkflowSurfaceOverviewStatus;
  summary: string;
};

export type WorkflowSurfaceOverviewRelation = {
  relationId: string;
  label: string;
  sourceRef: string;
  targetRef: string;
  status: WorkflowSurfaceOverviewStatus;
  summary: string;
  auditRef: string;
};

export type WorkflowSurfaceOverviewBlockedCapability = {
  capabilityId: string;
  label: string;
  sourceSurface: "application" | "draft" | "plan" | "readiness" | "run";
  status: "blocked";
  missingPrerequisite: string;
  summary: string;
  auditRef: string;
};

export type WorkflowSurfaceOverviewStopLine = {
  stopLineId: string;
  label: string;
  status: "locked";
  summary: string;
};

export type WorkflowSurfaceOverviewSource = {
  applicationDetail: WorkflowApplicationDetailViewModel;
  definitionDetail: WorkflowDefinitionDetailViewModel;
  runDetail: WorkflowRunDetailViewModel;
  selectedDraft: WorkflowDraftDesignerDraft;
  validationInspector: WorkflowDraftValidationInspectorViewModel;
  executionPlanPreview: WorkflowExecutionPlanPreviewViewModel;
  runtimeReadinessInspector: WorkflowRuntimeReadinessInspectorViewModel;
};

export type WorkflowSurfaceOverviewViewModel = {
  pageId: "workflow-surface-overview-offline";
  sourcePageIds: string[];
  overviewMode: "offline_read_only_advisory";
  applicationId: string;
  workflowDefinitionId: string;
  selectedDraftId: string;
  latestRunId: string;
  requestId: string;
  auditRef: string;
  summary: WorkflowSurfaceOverviewMetric[];
  relationMap: WorkflowSurfaceOverviewRelation[];
  blockedCapabilities: WorkflowSurfaceOverviewBlockedCapability[];
  stopLines: WorkflowSurfaceOverviewStopLine[];
  canRenderSurfaceOverview: boolean;
  canInspectOverviewLocally: true;
  canRequestLiveBackend: false;
  canMutateBuilder: false;
  canPersistDraft: false;
  canPersistExecutionPlan: false;
  canPersistRuntimeReadiness: false;
  canPublishWorkflow: false;
  canStartRuntime: false;
  canSubmitConfirmationDecision: false;
  canWriteBusinessTruth: false;
  canReplayRun: false;
  canAttachDatabase: false;
  canEnableRadishAuth: false;
  canImplementRepositoryAdapter: false;
};

export function buildWorkflowSurfaceOverviewViewModel(
  source: WorkflowSurfaceOverviewSource,
): WorkflowSurfaceOverviewViewModel {
  const blockedCapabilities = buildBlockedCapabilities(source);
  const relationMap = buildRelationMap(source, blockedCapabilities);
  const stopLines = buildStopLines();

  return {
    pageId: "workflow-surface-overview-offline",
    sourcePageIds: [
      source.applicationDetail.pageId,
      source.definitionDetail.pageId,
      source.runDetail.pageId,
      "workflow-draft-designer-offline",
      source.validationInspector.pageId,
      source.executionPlanPreview.pageId,
      source.runtimeReadinessInspector.pageId,
    ],
    overviewMode: "offline_read_only_advisory",
    applicationId: source.applicationDetail.applicationId,
    workflowDefinitionId: source.definitionDetail.workflowDefinitionId,
    selectedDraftId: source.selectedDraft.draftId,
    latestRunId: source.runDetail.runId,
    requestId: source.executionPlanPreview.requestId,
    auditRef: source.runtimeReadinessInspector.auditRef,
    summary: buildSummary(source, blockedCapabilities),
    relationMap,
    blockedCapabilities,
    stopLines,
    canRenderSurfaceOverview:
      source.applicationDetail.canRenderApplicationDetail &&
      source.definitionDetail.canRenderDefinitionDetail &&
      source.runDetail.canRenderRunDetail &&
      source.validationInspector.canRenderDraftValidationInspector &&
      source.executionPlanPreview.canRenderExecutionPlanPreview &&
      source.runtimeReadinessInspector.canRenderRuntimeReadinessInspector &&
      source.selectedDraft.nodes.length > 0 &&
      relationMap.length >= 7 &&
      blockedCapabilities.length >= 12,
    canInspectOverviewLocally: true,
    canRequestLiveBackend: false,
    canMutateBuilder: false,
    canPersistDraft: false,
    canPersistExecutionPlan: false,
    canPersistRuntimeReadiness: false,
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

function buildSummary(
  source: WorkflowSurfaceOverviewSource,
  blockedCapabilities: WorkflowSurfaceOverviewBlockedCapability[],
): WorkflowSurfaceOverviewMetric[] {
  return [
    {
      metricId: "application",
      label: "Application",
      value: source.applicationDetail.displayName,
      status: source.applicationDetail.canRenderApplicationDetail ? "ready" : "blocked",
      summary: `${source.applicationDetail.applicationId} uses ${source.applicationDetail.providerProfileRef} and stays read-only.`,
    },
    {
      metricId: "selected_draft",
      label: "Selected draft",
      value: source.selectedDraft.draftId,
      status: "offline_only",
      summary: `${source.selectedDraft.nodes.length} nodes and ${source.selectedDraft.edges.length} edges are available for local inspection.`,
    },
    {
      metricId: "validation",
      label: "Validation",
      value: source.validationInspector.validationStatus,
      status: validationStatusToOverviewStatus(source.validationInspector.validationStatus),
      summary: "Structural, contract, and blocked capability checks remain visible before any future runtime work.",
    },
    {
      metricId: "execution_plan",
      label: "Execution plan",
      value: `${source.executionPlanPreview.stageOrder.length} stages`,
      status: source.executionPlanPreview.canRenderExecutionPlanPreview ? "offline_only" : "blocked",
      summary: `${source.executionPlanPreview.nodeStageMappings.length} node-to-stage mappings are previewed without persistence.`,
    },
    {
      metricId: "runtime_readiness",
      label: "Runtime readiness",
      value: `${source.runtimeReadinessInspector.readinessBlockers.length} blockers`,
      status: "blocked",
      summary: "Executor, durable store, auth/repository, confirmation, writeback, and replay remain gated.",
    },
    {
      metricId: "blocked_capabilities",
      label: "Blocked capabilities",
      value: String(blockedCapabilities.length),
      status: "blocked",
      summary: "The overview groups blocked capability evidence from application, draft, plan, readiness, and run surfaces.",
    },
  ];
}

function buildRelationMap(
  source: WorkflowSurfaceOverviewSource,
  blockedCapabilities: WorkflowSurfaceOverviewBlockedCapability[],
): WorkflowSurfaceOverviewRelation[] {
  return [
    {
      relationId: "application_to_definition",
      label: "Application to definition",
      sourceRef: source.applicationDetail.applicationId,
      targetRef: source.applicationDetail.latestWorkflowDefinitionRef,
      status:
        source.applicationDetail.latestWorkflowDefinitionRef === source.definitionDetail.workflowDefinitionId
          ? "ready"
          : "review_required",
      summary: "Application detail points at the workflow definition detail surface without allowing lifecycle mutation.",
      auditRef: source.applicationDetail.auditRef,
    },
    {
      relationId: "definition_to_draft",
      label: "Definition to draft",
      sourceRef: source.definitionDetail.workflowDefinitionId,
      targetRef: source.selectedDraft.draftId,
      status: "offline_only",
      summary: "The selected draft is derived from committed definition metadata and can only be inspected locally.",
      auditRef: source.selectedDraft.routeMetadata.auditRef,
    },
    {
      relationId: "draft_to_validation",
      label: "Draft to validation",
      sourceRef: source.selectedDraft.draftId,
      targetRef: source.validationInspector.draftRouteId,
      status: validationStatusToOverviewStatus(source.validationInspector.validationStatus),
      summary: "Validation explains draft structure, contract fields, and blocked capability checks without storing results.",
      auditRef: source.validationInspector.auditRef,
    },
    {
      relationId: "validation_to_plan",
      label: "Validation to plan",
      sourceRef: source.validationInspector.inspectedDraftId,
      targetRef: source.executionPlanPreview.draftRouteId,
      status: source.executionPlanPreview.canRenderExecutionPlanPreview ? "offline_only" : "blocked",
      summary: "Execution plan preview adds stage order, provider requirements, and gates without creating a runtime plan.",
      auditRef: source.executionPlanPreview.auditRef,
    },
    {
      relationId: "plan_to_readiness",
      label: "Plan to readiness",
      sourceRef: source.executionPlanPreview.draftRouteId,
      targetRef: source.runtimeReadinessInspector.readinessRouteId,
      status: "blocked",
      summary: "Runtime readiness consumes the plan preview and keeps implementation gates blocked.",
      auditRef: source.runtimeReadinessInspector.auditRef,
    },
    {
      relationId: "application_to_latest_run",
      label: "Application to latest run",
      sourceRef: source.applicationDetail.latestRunRef,
      targetRef: source.runDetail.runId,
      status: source.applicationDetail.latestRunRef === source.runDetail.runId ? "ready" : "review_required",
      summary: "Run detail is a sanitized existing record; replay, resume, and result materialization stay disabled.",
      auditRef: source.runDetail.auditRef,
    },
    {
      relationId: "blocked_capability_rollup",
      label: "Blocked capability rollup",
      sourceRef: source.runtimeReadinessInspector.pageId,
      targetRef: `${blockedCapabilities.length} blocked capabilities`,
      status: "blocked",
      summary: "Blocked capabilities remain advisory evidence and do not expose execute, confirm, writeback, or replay controls.",
      auditRef: source.runtimeReadinessInspector.auditRef,
    },
  ];
}

function buildBlockedCapabilities(
  source: WorkflowSurfaceOverviewSource,
): WorkflowSurfaceOverviewBlockedCapability[] {
  const applicationCapabilities = source.applicationDetail.blockedCapabilities.map((capability) => ({
    capabilityId: `application_${capability.capabilityId}`,
    label: capability.label,
    sourceSurface: "application" as const,
    status: "blocked" as const,
    missingPrerequisite: capability.missingPrerequisite,
    summary: capability.reason,
    auditRef: capability.auditRef,
  }));
  const draftCapabilities = source.selectedDraft.blockedCapabilities.map((capability) => ({
    capabilityId: `draft_${capability.capabilityId}`,
    label: capability.label,
    sourceSurface: "draft" as const,
    status: "blocked" as const,
    missingPrerequisite: capability.missingPrerequisite,
    summary: capability.summary,
    auditRef: capability.auditRef,
  }));
  const planCapabilities = source.executionPlanPreview.blockedPlanReasons.map((reason) => ({
    capabilityId: `plan_${reason.reasonId}`,
    label: reason.label,
    sourceSurface: "plan" as const,
    status: "blocked" as const,
    missingPrerequisite: reason.missingPrerequisite,
    summary: reason.summary,
    auditRef: reason.auditRef,
  }));
  const readinessCapabilities = source.runtimeReadinessInspector.readinessBlockers.map((blocker) => ({
    capabilityId: `readiness_${blocker.blockerId}`,
    label: blocker.label,
    sourceSurface: "readiness" as const,
    status: "blocked" as const,
    missingPrerequisite: blocker.missingPrerequisite,
    summary: blocker.summary,
    auditRef: blocker.auditRef,
  }));
  const runCapabilities: WorkflowSurfaceOverviewBlockedCapability[] = [
    {
      capabilityId: `run_${source.runDetail.blockedResultPreview.guardId}`,
      label: source.runDetail.blockedResultPreview.label,
      sourceSurface: "run",
      status: "blocked",
      missingPrerequisite: source.runDetail.blockedResultPreview.missingPrerequisite,
      summary: source.runDetail.blockedResultPreview.reason,
      auditRef: source.runDetail.blockedResultPreview.auditRef,
    },
    {
      capabilityId: `run_${source.runDetail.blockedReplayPreview.guardId}`,
      label: source.runDetail.blockedReplayPreview.label,
      sourceSurface: "run",
      status: "blocked",
      missingPrerequisite: source.runDetail.blockedReplayPreview.missingPrerequisite,
      summary: source.runDetail.blockedReplayPreview.reason,
      auditRef: source.runDetail.blockedReplayPreview.auditRef,
    },
  ];

  return [
    ...applicationCapabilities,
    ...draftCapabilities,
    ...planCapabilities,
    ...readinessCapabilities,
    ...runCapabilities,
  ];
}

function buildStopLines(): WorkflowSurfaceOverviewStopLine[] {
  return [
    {
      stopLineId: "offline_only",
      label: "Offline only",
      status: "locked",
      summary: "The overview consumes fixture-derived view models and never requests a live backend.",
    },
    {
      stopLineId: "builder_mutation",
      label: "Builder mutation",
      status: "locked",
      summary: "Draft selection is local inspection only; no create, edit, save, publish, or lifecycle mutation exists.",
    },
    {
      stopLineId: "runtime_execution",
      label: "Runtime execution",
      status: "locked",
      summary: "Workflow executor, node executor, tool executor, and agent loop remain outside this UI.",
    },
    {
      stopLineId: "confirmation_decision",
      label: "Confirmation decision",
      status: "locked",
      summary: "Human review shape is visible, but no approve, reject, defer, persist, or execution unlock path exists.",
    },
    {
      stopLineId: "writeback_replay",
      label: "Writeback and replay",
      status: "locked",
      summary: "Candidate actions stay advisory; business writeback, run replay, and run resume remain blocked.",
    },
    {
      stopLineId: "auth_database_repository",
      label: "Auth, database, repository",
      status: "locked",
      summary: "Radish auth, database attach, store selector, and repository adapter work need separate implementation triggers.",
    },
  ];
}

function validationStatusToOverviewStatus(status: string): WorkflowSurfaceOverviewStatus {
  if (status === "passed") {
    return "ready";
  }
  if (status === "blocked") {
    return "blocked";
  }
  return "review_required";
}
