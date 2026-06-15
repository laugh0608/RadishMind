import type { WorkflowDefinitionSummary } from "../../../../../contracts/typescript/control-plane-read-api";
import {
  buildWorkflowApplicationDetailViewModel,
  type WorkflowApplicationDetailViewModel,
} from "./workflowApplicationDetail";
import {
  buildWorkflowBlockedActionPreviewViewModel,
  type WorkflowBlockedActionPreviewViewModel,
} from "./workflowBlockedActionPreview";
import {
  buildWorkflowConfirmationPlaceholderViewModel,
  type WorkflowConfirmationPlaceholderViewModel,
} from "./workflowConfirmationPlaceholder";
import {
  buildWorkflowDefinitionDetailViewModel,
  type WorkflowDefinitionDetailViewModel,
} from "./workflowDefinitionDetail";
import {
  buildWorkflowDraftDesignerViewModel,
  type WorkflowDraftDesignerDefinitionDetailSource,
  type WorkflowDraftDesignerDraft,
  type WorkflowDraftDesignerViewModel,
} from "./workflowDraftDesigner";
import {
  buildWorkflowDraftValidationInspectorViewModel,
  type WorkflowDraftValidationInspectorViewModel,
} from "./workflowDraftValidationInspector";
import {
  buildWorkflowExecutionPlanPreviewViewModel,
  type WorkflowExecutionPlanPreviewViewModel,
} from "./workflowExecutionPlanPreview";
import {
  buildWorkflowRunDetailViewModel,
  type WorkflowRunDetailViewModel,
} from "./workflowRunDetail";
import {
  buildWorkflowRuntimeReadinessInspectorViewModel,
  type WorkflowRuntimeReadinessInspectorViewModel,
} from "./workflowRuntimeReadinessInspector";
import {
  buildWorkflowReviewHandoffViewModel,
  type WorkflowReviewHandoffViewModel,
} from "./workflowReviewHandoff";
import {
  buildWorkflowScenarioInspectorViewModel,
  type WorkflowScenarioInspectorViewModel,
} from "./workflowScenarioInspector";
import {
  buildWorkflowSurfaceOverviewViewModel,
  type WorkflowSurfaceOverviewViewModel,
} from "./workflowSurfaceOverview";
import {
  buildWorkflowUserWorkspaceHomeViewModel,
  type WorkflowUserWorkspaceHomeViewModel,
} from "./workflowUserWorkspaceHome";
import {
  buildWorkflowWorkspaceReviewViewModel,
  type WorkflowWorkspaceReviewViewModel,
} from "./workflowWorkspaceReview";
import type {
  WorkspaceApplicationRow,
  WorkspaceApplicationsViewModel,
} from "./workspaceApplications";
import type { WorkspaceApiKeysViewModel } from "./workspaceApiKeys";
import type { WorkspaceRunHistoryViewModel, WorkspaceRunRecordRow } from "./workspaceRunHistory";
import type { WorkspaceUsageQuotaViewModel } from "./workspaceUsageQuota";
import type {
  WorkspaceWorkflowDefinitionRow,
  WorkspaceWorkflowDefinitionsViewModel,
} from "./workspaceWorkflowDefinitions";

export type WorkflowWorkspaceSelectionState = {
  applicationRef: string | null;
  workflowDefinitionId: string | null;
  runId: string | null;
  draftId: string | null;
  scenarioId: string | null;
};

export type WorkflowWorkspaceContextSource = {
  workspaceApplications: WorkspaceApplicationsViewModel;
  workspaceApiKeys: WorkspaceApiKeysViewModel;
  workspaceUsageQuota: WorkspaceUsageQuotaViewModel;
  workspaceWorkflowDefinitions: WorkspaceWorkflowDefinitionsViewModel;
  workspaceRunHistory: WorkspaceRunHistoryViewModel;
  localWorkflowDrafts?: WorkflowDraftDesignerDraft[];
  selection: WorkflowWorkspaceSelectionState;
};

export type WorkflowWorkspaceSelectionPatch = {
  applicationRef: string | null;
  workflowDefinitionId: string | null;
  runId: string | null;
  draftId: string | null;
  scenarioId: string | null;
};

export type WorkflowWorkspaceContextViewModel = {
  selectedApplication: WorkspaceApplicationRow;
  selectedWorkflowDefinition: WorkspaceWorkflowDefinitionRow;
  selectedRun: WorkspaceRunRecordRow;
  selectedWorkflowDraft: WorkflowDraftDesignerDraft;
  workflowDefinitionsForSelectedApplication: WorkspaceWorkflowDefinitionRow[];
  runsForSelectedContext: WorkspaceRunRecordRow[];
  workflowApplicationDetail: WorkflowApplicationDetailViewModel;
  workflowDefinitionDetail: WorkflowDefinitionDetailViewModel;
  workflowRunDetail: WorkflowRunDetailViewModel;
  workflowBlockedActionPreview: WorkflowBlockedActionPreviewViewModel;
  workflowConfirmationPlaceholder: WorkflowConfirmationPlaceholderViewModel;
  workflowDraftDesigner: WorkflowDraftDesignerViewModel;
  workflowDraftValidationInspector: WorkflowDraftValidationInspectorViewModel;
  workflowExecutionPlanPreview: WorkflowExecutionPlanPreviewViewModel;
  workflowRuntimeReadinessInspector: WorkflowRuntimeReadinessInspectorViewModel;
  workflowSurfaceOverview: WorkflowSurfaceOverviewViewModel;
  workflowScenarioInspector: WorkflowScenarioInspectorViewModel;
  workflowWorkspaceReview: WorkflowWorkspaceReviewViewModel;
  workflowUserWorkspaceHome: WorkflowUserWorkspaceHomeViewModel;
  workflowReviewHandoff: WorkflowReviewHandoffViewModel;
};

export function buildWorkflowWorkspaceContextViewModel(
  source: WorkflowWorkspaceContextSource,
): WorkflowWorkspaceContextViewModel {
  const selectedApplication = selectApplication(
    source.workspaceApplications,
    source.selection.applicationRef,
  );
  const workflowDefinitionsForSelectedApplication = workflowDefinitionsForApplication(
    source.workspaceWorkflowDefinitions,
    selectedApplication,
  );
  const selectedWorkflowDefinition = selectWorkflowDefinition(
    workflowDefinitionsForSelectedApplication,
    selectedApplication,
    source.selection.workflowDefinitionId,
  );
  const workflowDefinitionDetailsById = buildWorkflowDefinitionDetailsById(
    source.workspaceWorkflowDefinitions,
  );
  const workflowDefinitionDetail =
    workflowDefinitionDetailsById[selectedWorkflowDefinition.workflowDefinitionId] ??
    buildWorkflowDefinitionDetailViewModel(
      toWorkflowDefinitionSummary(
        selectedWorkflowDefinition,
        source.workspaceWorkflowDefinitions.collection.tenantRef,
      ),
    );
  const workflowApplicationDetail = buildWorkflowApplicationDetailViewModel(selectedApplication, {
    tenantRef: source.workspaceApplications.collection.tenantRef,
    requestId: source.workspaceApplications.requestId,
    auditRef: source.workspaceApplications.auditRef,
  });
  const runsForSelectedContext = runsForApplicationAndDefinition(
    source.workspaceRunHistory,
    selectedApplication,
    selectedWorkflowDefinition,
  );
  const selectedRun = selectRun(runsForSelectedContext, source.selection.runId);
  const workflowRunDetail = buildWorkflowRunDetailViewModel(selectedRun);
  const workflowBlockedActionPreview = buildWorkflowBlockedActionPreviewViewModel(
    workflowDefinitionDetail.blockedActionPreview,
    workflowRunDetail.blockedReplayPreview,
    {
      runId: workflowRunDetail.runId,
      workflowDefinitionId: workflowDefinitionDetail.workflowDefinitionId,
      requestId: workflowRunDetail.requestId,
      auditRef: workflowRunDetail.auditRef,
    },
  );
  const workflowConfirmationPlaceholder = buildWorkflowConfirmationPlaceholderViewModel(
    workflowBlockedActionPreview,
  );
  const workflowDraftDesigner = buildWorkflowDraftDesignerViewModel({
    workflowDefinitions: source.workspaceWorkflowDefinitions.workflowDefinitions,
    localDrafts: source.localWorkflowDrafts,
    tenantRef: source.workspaceWorkflowDefinitions.collection.tenantRef,
    detailSourcesByWorkflowDefinitionId: toDraftDesignerDetailSources(workflowDefinitionDetailsById),
    confirmationPlaceholder: workflowConfirmationPlaceholder,
    sourceRequestId: source.workspaceWorkflowDefinitions.requestId,
    sourceAuditRef: source.workspaceWorkflowDefinitions.auditRef,
  });
  const selectedWorkflowDraft = selectWorkflowDraft(
    workflowDraftDesigner,
    selectedWorkflowDefinition,
    source.selection.draftId,
  );
  const workflowDraftValidationInspector = buildWorkflowDraftValidationInspectorViewModel(
    selectedWorkflowDraft,
  );
  const workflowExecutionPlanPreview = buildWorkflowExecutionPlanPreviewViewModel(
    selectedWorkflowDraft,
    workflowDraftValidationInspector,
  );
  const workflowRuntimeReadinessInspector = buildWorkflowRuntimeReadinessInspectorViewModel(
    workflowExecutionPlanPreview,
  );
  const overviewSource = {
    applicationDetail: workflowApplicationDetail,
    definitionDetail: workflowDefinitionDetail,
    runDetail: workflowRunDetail,
    selectedDraft: selectedWorkflowDraft,
    validationInspector: workflowDraftValidationInspector,
    executionPlanPreview: workflowExecutionPlanPreview,
    runtimeReadinessInspector: workflowRuntimeReadinessInspector,
  };
  const workflowSurfaceOverview = buildWorkflowSurfaceOverviewViewModel(overviewSource);
  const workflowScenarioInspector = buildWorkflowScenarioInspectorViewModel(
    overviewSource,
    source.selection.scenarioId,
  );
  const workflowWorkspaceReview = buildWorkflowWorkspaceReviewViewModel({
    ...overviewSource,
    surfaceOverview: workflowSurfaceOverview,
    scenarioInspector: workflowScenarioInspector,
  });
  const workflowUserWorkspaceHome = buildWorkflowUserWorkspaceHomeViewModel({
    workspaceApplications: source.workspaceApplications,
    workspaceApiKeys: source.workspaceApiKeys,
    workspaceUsageQuota: source.workspaceUsageQuota,
    workspaceWorkflowDefinitions: source.workspaceWorkflowDefinitions,
    workspaceRunHistory: source.workspaceRunHistory,
    workflowWorkspaceReview,
    workflowSurfaceOverview,
    workflowScenarioInspector,
  });
  const workflowReviewHandoff = buildWorkflowReviewHandoffViewModel({
    workflowUserWorkspaceHome,
    workflowWorkspaceReview,
    workflowSurfaceOverview,
    workflowScenarioInspector,
    workflowRuntimeReadinessInspector,
    workflowBlockedActionPreview,
    workflowConfirmationPlaceholder,
  });

  return {
    selectedApplication,
    selectedWorkflowDefinition,
    selectedRun,
    selectedWorkflowDraft,
    workflowDefinitionsForSelectedApplication,
    runsForSelectedContext,
    workflowApplicationDetail,
    workflowDefinitionDetail,
    workflowRunDetail,
    workflowBlockedActionPreview,
    workflowConfirmationPlaceholder,
    workflowDraftDesigner,
    workflowDraftValidationInspector,
    workflowExecutionPlanPreview,
    workflowRuntimeReadinessInspector,
    workflowSurfaceOverview,
    workflowScenarioInspector,
    workflowWorkspaceReview,
    workflowUserWorkspaceHome,
    workflowReviewHandoff,
  };
}

export function selectionForApplication(
  applicationRef: string,
  source: Pick<WorkflowWorkspaceContextSource, "workspaceApplications" | "workspaceWorkflowDefinitions" | "workspaceRunHistory">,
): WorkflowWorkspaceSelectionPatch {
  const nextApplication = source.workspaceApplications.applications.find(
    (application) => application.applicationRef === applicationRef,
  );
  const nextDefinition =
    source.workspaceWorkflowDefinitions.workflowDefinitions.find(
      (workflowDefinition) =>
        workflowDefinition.applicationRef === applicationRef &&
        workflowDefinition.workflowDefinitionId === nextApplication?.latestWorkflowDefinitionRef,
    ) ??
    source.workspaceWorkflowDefinitions.workflowDefinitions.find(
      (workflowDefinition) => workflowDefinition.applicationRef === applicationRef,
    );
  const nextRun =
    source.workspaceRunHistory.runs.find(
      (run) =>
        run.applicationRef === applicationRef &&
        (!nextDefinition || run.workflowDefinitionId === nextDefinition.workflowDefinitionId),
    ) ?? source.workspaceRunHistory.runs.find((run) => run.applicationRef === applicationRef);

  return {
    applicationRef,
    workflowDefinitionId: nextDefinition?.workflowDefinitionId ?? null,
    runId: nextRun?.runId ?? null,
    draftId: null,
    scenarioId: null,
  };
}

export function selectionForWorkflowDefinition(
  workflowDefinitionId: string,
  source: Pick<WorkflowWorkspaceContextSource, "workspaceWorkflowDefinitions" | "workspaceRunHistory">,
): WorkflowWorkspaceSelectionPatch {
  const nextDefinition = source.workspaceWorkflowDefinitions.workflowDefinitions.find(
    (workflowDefinition) => workflowDefinition.workflowDefinitionId === workflowDefinitionId,
  );
  const nextRun = source.workspaceRunHistory.runs.find(
    (run) =>
      run.workflowDefinitionId === workflowDefinitionId &&
      (!nextDefinition || run.applicationRef === nextDefinition.applicationRef),
  );

  return {
    applicationRef: nextDefinition?.applicationRef ?? null,
    workflowDefinitionId,
    runId: nextRun?.runId ?? null,
    draftId: null,
    scenarioId: null,
  };
}

export function selectionForRun(
  runId: string,
  source: Pick<WorkflowWorkspaceContextSource, "workspaceRunHistory">,
): WorkflowWorkspaceSelectionPatch {
  const nextRun = source.workspaceRunHistory.runs.find((run) => run.runId === runId);

  return {
    applicationRef: nextRun?.applicationRef ?? null,
    workflowDefinitionId: nextRun?.workflowDefinitionId ?? null,
    runId,
    draftId: null,
    scenarioId: null,
  };
}

export function selectionForDraft(
  draftId: string,
  workflowDraftDesigner: WorkflowDraftDesignerViewModel,
  source: Pick<WorkflowWorkspaceContextSource, "workspaceRunHistory">,
): WorkflowWorkspaceSelectionPatch {
  const nextDraft = workflowDraftDesigner.drafts.find((draft) => draft.draftId === draftId);
  const nextRun = source.workspaceRunHistory.runs.find(
    (run) =>
      run.applicationRef === nextDraft?.applicationRef &&
      run.workflowDefinitionId === nextDraft?.workflowDefinitionId,
  );

  return {
    applicationRef: nextDraft?.applicationRef ?? null,
    workflowDefinitionId: nextDraft?.workflowDefinitionId ?? null,
    runId: nextRun?.runId ?? null,
    draftId,
    scenarioId: null,
  };
}

function selectApplication(
  workspaceApplications: WorkspaceApplicationsViewModel,
  selectedApplicationRef: string | null,
): WorkspaceApplicationRow {
  const targetApplicationRef =
    selectedApplicationRef ?? workspaceApplications.applications[0]?.applicationRef;
  return (
    workspaceApplications.applications.find(
      (application) => application.applicationRef === targetApplicationRef,
    ) ?? workspaceApplications.applications[0]!
  );
}

function workflowDefinitionsForApplication(
  workspaceWorkflowDefinitions: WorkspaceWorkflowDefinitionsViewModel,
  selectedApplication: WorkspaceApplicationRow,
): WorkspaceWorkflowDefinitionRow[] {
  const filteredDefinitions = workspaceWorkflowDefinitions.workflowDefinitions.filter(
    (workflowDefinition) => workflowDefinition.applicationRef === selectedApplication.applicationRef,
  );
  return filteredDefinitions.length > 0
    ? filteredDefinitions
    : workspaceWorkflowDefinitions.workflowDefinitions;
}

function selectWorkflowDefinition(
  workflowDefinitions: WorkspaceWorkflowDefinitionRow[],
  selectedApplication: WorkspaceApplicationRow,
  selectedWorkflowDefinitionId: string | null,
): WorkspaceWorkflowDefinitionRow {
  const targetWorkflowDefinitionId =
    selectedWorkflowDefinitionId ??
    selectedApplication.latestWorkflowDefinitionRef ??
    workflowDefinitions[0]?.workflowDefinitionId;
  return (
    workflowDefinitions.find(
      (workflowDefinition) => workflowDefinition.workflowDefinitionId === targetWorkflowDefinitionId,
    ) ??
    workflowDefinitions.find(
      (workflowDefinition) =>
        workflowDefinition.workflowDefinitionId === selectedApplication.latestWorkflowDefinitionRef,
    ) ??
    workflowDefinitions[0]!
  );
}

function runsForApplicationAndDefinition(
  workspaceRunHistory: WorkspaceRunHistoryViewModel,
  selectedApplication: WorkspaceApplicationRow,
  selectedWorkflowDefinition: WorkspaceWorkflowDefinitionRow,
): WorkspaceRunRecordRow[] {
  const definitionRuns = workspaceRunHistory.runs.filter(
    (run) =>
      run.applicationRef === selectedApplication.applicationRef &&
      run.workflowDefinitionId === selectedWorkflowDefinition.workflowDefinitionId,
  );
  if (definitionRuns.length > 0) {
    return definitionRuns;
  }
  const applicationRuns = workspaceRunHistory.runs.filter(
    (run) => run.applicationRef === selectedApplication.applicationRef,
  );
  return applicationRuns.length > 0 ? applicationRuns : workspaceRunHistory.runs;
}

function selectRun(
  runsForSelectedContext: WorkspaceRunRecordRow[],
  selectedRunId: string | null,
): WorkspaceRunRecordRow {
  const targetRunId = selectedRunId ?? runsForSelectedContext[0]?.runId;
  return runsForSelectedContext.find((run) => run.runId === targetRunId) ?? runsForSelectedContext[0]!;
}

function selectWorkflowDraft(
  workflowDraftDesigner: WorkflowDraftDesignerViewModel,
  selectedWorkflowDefinition: WorkspaceWorkflowDefinitionRow,
  selectedWorkflowDraftId: string | null,
): WorkflowDraftDesignerDraft {
  const definitionDraft = workflowDraftDesigner.drafts.find(
    (draft) => draft.workflowDefinitionId === selectedWorkflowDefinition.workflowDefinitionId,
  );
  const targetDraftId =
    selectedWorkflowDraftId ?? definitionDraft?.draftId ?? workflowDraftDesigner.defaultDraftId;
  return (
    workflowDraftDesigner.drafts.find((draft) => draft.draftId === targetDraftId) ??
    definitionDraft ??
    workflowDraftDesigner.drafts[0]!
  );
}

function buildWorkflowDefinitionDetailsById(
  workspaceWorkflowDefinitions: WorkspaceWorkflowDefinitionsViewModel,
): Record<string, WorkflowDefinitionDetailViewModel> {
  return Object.fromEntries(
    workspaceWorkflowDefinitions.workflowDefinitions.map((workflowDefinition) => {
      const detail = buildWorkflowDefinitionDetailViewModel(
        toWorkflowDefinitionSummary(
          workflowDefinition,
          workspaceWorkflowDefinitions.collection.tenantRef,
        ),
      );
      return [workflowDefinition.workflowDefinitionId, detail];
    }),
  );
}

function toDraftDesignerDetailSources(
  workflowDefinitionDetailsById: Record<string, WorkflowDefinitionDetailViewModel>,
): Record<string, WorkflowDraftDesignerDefinitionDetailSource> {
  return Object.fromEntries(
    Object.entries(workflowDefinitionDetailsById).map(([workflowDefinitionId, detail]) => [
      workflowDefinitionId,
      {
        nodes: detail.nodes,
        edges: detail.edges,
        blockedActionPreview: detail.blockedActionPreview,
      },
    ]),
  );
}

function toWorkflowDefinitionSummary(
  workflowDefinition: WorkspaceWorkflowDefinitionRow,
  tenantRef: string,
): WorkflowDefinitionSummary {
  return {
    workflow_definition_id: workflowDefinition.workflowDefinitionId,
    tenant_ref: tenantRef,
    application_ref: workflowDefinition.applicationRef,
    version: workflowDefinition.version,
    definition_status: workflowDefinition.definitionStatus,
    node_count: workflowDefinition.nodeCount,
    risk_level: workflowDefinition.riskLevel,
    requires_confirmation_capable: workflowDefinition.requiresConfirmationCapable,
    updated_at: workflowDefinition.updatedAt,
  };
}
