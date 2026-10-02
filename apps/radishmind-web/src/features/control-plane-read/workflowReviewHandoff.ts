import type { WorkflowProjectionCopy } from "./workflowProjectionCopy.ts";
import type { WorkflowBlockedActionPreviewViewModel } from "./workflowBlockedActionPreview";
import type { WorkflowConfirmationPlaceholderViewModel } from "./workflowConfirmationPlaceholder";
import type { WorkflowDraftDesignerDraft } from "./workflowDraftDesigner";
import type { WorkflowDraftValidationInspectorViewModel } from "./workflowDraftValidationInspector";
import type { WorkflowExecutionPlanPreviewViewModel } from "./workflowExecutionPlanPreview";
import type { WorkflowSavedDraftConflictReviewSummary } from "./savedWorkflowDraftConsumer";
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

export type WorkflowReviewHandoffRecipient = WorkflowProjectionCopy & {
  recipientId: string;
  label: string;
  role: string;
  status: WorkflowReviewHandoffStatus;
  handoffNeed: string;
  evidenceRefs: string[];
};

export type WorkflowReviewHandoffFinding = WorkflowProjectionCopy & {
  findingId: string;
  label: string;
  sourceSurface:
    | "scenario"
    | "review"
    | "validation"
    | "saved_draft_conflict"
    | "node_designer"
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

export type WorkflowReviewHandoffEvidence = WorkflowProjectionCopy & {
  evidenceId: string;
  label: string;
  sourceSurface:
    | "home"
    | "review"
    | "scenario"
    | "overview"
    | "validation"
    | "saved_draft_conflict"
    | "node_designer"
    | "plan"
    | "readiness"
    | "blocked_action"
    | "confirmation";
  routeOrPageId: string;
  requestId: string;
  auditRef: string;
  status: WorkflowReviewHandoffStatus;
  summary: string;
};

export type WorkflowReviewHandoffDecisionBlocker = WorkflowProjectionCopy & {
  blockerId: string;
  label: string;
  sourceSurface: string;
  status: "blocked";
  missingPrerequisite: string;
  summary: string;
  auditRefs: string[];
};

export type WorkflowReviewHandoffBoundaryLock = WorkflowProjectionCopy & {
  boundaryId: string;
  label: string;
  status: "locked";
  summary: string;
};

export type WorkflowReviewHandoffActiveDraftReviewSection = WorkflowProjectionCopy & {
  sectionId: "active_draft_validation" | "active_draft_execution_plan" | "active_draft_runtime_readiness";
  label: string;
  sourceSurface: "validation" | "plan" | "readiness";
  status: WorkflowReviewHandoffStatus;
  primaryRef: string;
  requestId: string;
  auditRef: string;
  blockerCount: number;
  summary: string;
  reviewerQuestion: string;
  evidenceRefs: string[];
};

export type WorkflowReviewHandoffActiveDraftReviewRecord = {
  recordId: "active_draft_review_record";
  recordMode: "active_draft_advisory_only";
  draftId: string;
  validationStatus: WorkflowReviewHandoffStatus;
  planPreviewStatus: WorkflowReviewHandoffStatus;
  runtimeReadinessStatus: "blocked";
  sections: WorkflowReviewHandoffActiveDraftReviewSection[];
  canRenderActiveDraftReviewRecord: boolean;
  canPersistRecord: false;
  canExportRecord: false;
  canSendRecord: false;
  canStartRuntime: false;
  canSubmitConfirmationDecision: false;
  canWriteBusinessTruth: false;
};

export type WorkflowReviewHandoffNodeDesignerReviewSection = WorkflowProjectionCopy & {
  sectionId:
    | "node_designer_canvas_layout"
    | "node_designer_validation_overlay"
    | "node_designer_inspector_state"
    | "node_designer_saved_draft_mapping";
  label: string;
  sourceSurface: "node_designer";
  status: WorkflowReviewHandoffStatus;
  primaryRef: string;
  requestId: string;
  auditRef: string;
  itemCount: number;
  summary: string;
  reviewerQuestion: string;
  evidenceRefs: string[];
};

export type WorkflowReviewHandoffNodeDesignerGraphFinding = WorkflowProjectionCopy & {
  findingId: string;
  label: string;
  sourceCheckId: string;
  targetKind: "node" | "edge" | "graph";
  status: WorkflowReviewHandoffStatus;
  severity: "info" | "warning" | "blocking";
  targetRefs: string[];
  targetSummary: string;
  handoffPath: string;
  handoffPathRefs: string[];
  summary: string;
  reviewerQuestion: string;
  evidenceRefs: string[];
};

export type WorkflowReviewHandoffNodeDesignerReviewRecord = {
  recordId: "node_designer_review_record";
  recordMode: "node_designer_advisory_only";
  draftId: string;
  positionedNodeCount: number;
  defaultLayoutNodeCount: number;
  derivedEdgeCount: number;
  validationOverlayCount: number;
  inspectorFieldCount: number;
  graphReviewFindings: WorkflowReviewHandoffNodeDesignerGraphFinding[];
  nodeTargetedFindingCount: number;
  edgeTargetedFindingCount: number;
  graphLevelFindingCount: number;
  sections: WorkflowReviewHandoffNodeDesignerReviewSection[];
  canRenderNodeDesignerReviewRecord: boolean;
  canPersistLayout: false;
  canPersistEdgeKind: false;
  canPersistOverlay: false;
  canPersistInspectorState: false;
  canExportRecord: false;
  canSendRecord: false;
  canStartRuntime: false;
  canSubmitConfirmationDecision: false;
  canWriteBusinessTruth: false;
};

export type WorkflowReviewHandoffSource = {
  activeWorkflowDraft: WorkflowDraftDesignerDraft;
  savedDraftConflictReviewSummary?: WorkflowSavedDraftConflictReviewSummary | null;
  workflowUserWorkspaceHome: WorkflowUserWorkspaceHomeViewModel;
  workflowWorkspaceReview: WorkflowWorkspaceReviewViewModel;
  workflowSurfaceOverview: WorkflowSurfaceOverviewViewModel;
  workflowScenarioInspector: WorkflowScenarioInspectorViewModel;
  workflowDraftValidationInspector: WorkflowDraftValidationInspectorViewModel;
  workflowExecutionPlanPreview: WorkflowExecutionPlanPreviewViewModel;
  workflowRuntimeReadinessInspector: WorkflowRuntimeReadinessInspectorViewModel;
  workflowBlockedActionPreview: WorkflowBlockedActionPreviewViewModel;
  workflowConfirmationPlaceholder: WorkflowConfirmationPlaceholderViewModel;
};

export type WorkflowReviewHandoffViewModel = WorkflowProjectionCopy & {
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
  activeDraftReviewRecord: WorkflowReviewHandoffActiveDraftReviewRecord;
  savedDraftConflictReviewSummary: WorkflowSavedDraftConflictReviewSummary | null;
  nodeDesignerReviewRecord: WorkflowReviewHandoffNodeDesignerReviewRecord;
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
  const activeDraftReviewRecord = buildActiveDraftReviewRecord(source);
  const savedDraftConflictReviewSummary = source.savedDraftConflictReviewSummary ?? null;
  const nodeDesignerReviewRecord = buildNodeDesignerReviewRecord(source);
  const recipients = buildRecipients(source);
  const keyFindings = buildKeyFindings(source, nodeDesignerReviewRecord);
  const evidenceChecklist = buildEvidenceChecklist(source, nodeDesignerReviewRecord);
  const decisionBlockers = buildDecisionBlockers(source);
  const boundaryLocks = buildBoundaryLocks(source);

  return {
    pageId: "workflow-review-handoff-offline",
    sourcePageIds: [
      source.workflowUserWorkspaceHome.pageId,
      source.workflowWorkspaceReview.pageId,
      source.workflowSurfaceOverview.pageId,
      source.workflowScenarioInspector.pageId,
      source.workflowDraftValidationInspector.pageId,
      source.workflowExecutionPlanPreview.pageId,
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
    handoffNarrative: buildHandoffNarrative(
      source,
      activeDraftReviewRecord,
      savedDraftConflictReviewSummary,
      nodeDesignerReviewRecord,
      decisionBlockers,
      boundaryLocks,
    ),
    handoffNarrativeMessage: {
      key: "handoffNarrative", values: {
        scenario: source.workflowScenarioInspector.selectedScenario.labelMessage ?? source.workflowScenarioInspector.selectedScenario.label, active: activeDraftReviewRecord.sections.length, conflicts: savedDraftConflictReviewSummary ? 1 : 0, conflictCode: savedDraftConflictReviewSummary ? ` (${savedDraftConflictReviewSummary.failureCode})` : "", canvas: nodeDesignerReviewRecord.sections.length, stages: source.workflowWorkspaceReview.reviewStages.length, blockers: decisionBlockers.length, routes: source.workflowUserWorkspaceHome.routeEvidence.length, boundaries: boundaryLocks.length
      }
    },
    activeDraftReviewRecord,
    savedDraftConflictReviewSummary,
    nodeDesignerReviewRecord,
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
      source.workflowDraftValidationInspector.canRenderDraftValidationInspector &&
      source.workflowExecutionPlanPreview.canRenderExecutionPlanPreview &&
      source.workflowRuntimeReadinessInspector.canRenderRuntimeReadinessInspector &&
      source.workflowBlockedActionPreview.canRenderBlockedActionPreview &&
      source.workflowConfirmationPlaceholder.canRenderConfirmationPlaceholder &&
      activeDraftReviewRecord.canRenderActiveDraftReviewRecord &&
      nodeDesignerReviewRecord.canRenderNodeDesignerReviewRecord &&
      recipients.length === 4 &&
      keyFindings.length >= 8 &&
      evidenceChecklist.length >= 12 &&
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
  activeDraftReviewRecord: WorkflowReviewHandoffActiveDraftReviewRecord,
  savedDraftConflictReviewSummary: WorkflowSavedDraftConflictReviewSummary | null,
  nodeDesignerReviewRecord: WorkflowReviewHandoffNodeDesignerReviewRecord,
  decisionBlockers: WorkflowReviewHandoffDecisionBlocker[],
  boundaryLocks: WorkflowReviewHandoffBoundaryLock[],
): string {
  const conflictReviewClause = savedDraftConflictReviewSummary
    ? `, 1 saved draft conflict review for ${savedDraftConflictReviewSummary.failureCode}`
    : "";
  return `${source.workflowScenarioInspector.selectedScenario.label} is packaged for human review with ${activeDraftReviewRecord.sections.length} active draft review sections${conflictReviewClause}, ${nodeDesignerReviewRecord.sections.length} node designer review sections, ${source.workflowWorkspaceReview.reviewStages.length} review stages, ${decisionBlockers.length} decision blockers, ${source.workflowUserWorkspaceHome.routeEvidence.length} route evidence entries, and ${boundaryLocks.length} locked boundaries.`;
}

function buildActiveDraftReviewRecord(
  source: WorkflowReviewHandoffSource,
): WorkflowReviewHandoffActiveDraftReviewRecord {
  const sections = buildActiveDraftReviewSections(source);
  const validationStatus = validationStatusToHandoffStatus(
    source.workflowDraftValidationInspector.validationStatus,
  );
  const planPreviewStatus = source.workflowExecutionPlanPreview.canRenderExecutionPlanPreview
    ? "review_required"
    : "blocked";
  const draftIdsMatch =
    source.workflowWorkspaceReview.draftId === source.workflowDraftValidationInspector.inspectedDraftId &&
    source.workflowDraftValidationInspector.inspectedDraftId ===
      source.workflowExecutionPlanPreview.selectedDraftId &&
    source.workflowExecutionPlanPreview.selectedDraftId ===
      source.workflowRuntimeReadinessInspector.selectedDraftId;

  return {
    recordId: "active_draft_review_record",
    recordMode: "active_draft_advisory_only",
    draftId: source.workflowWorkspaceReview.draftId,
    validationStatus,
    planPreviewStatus,
    runtimeReadinessStatus: "blocked",
    sections,
    canRenderActiveDraftReviewRecord:
      draftIdsMatch &&
      source.workflowDraftValidationInspector.canRenderDraftValidationInspector &&
      source.workflowExecutionPlanPreview.canRenderExecutionPlanPreview &&
      source.workflowRuntimeReadinessInspector.canRenderRuntimeReadinessInspector &&
      sections.length === 3 &&
      sections.every((section) => section.requestId.length > 0 && section.auditRef.length > 0),
    canPersistRecord: false,
    canExportRecord: false,
    canSendRecord: false,
    canStartRuntime: false,
    canSubmitConfirmationDecision: false,
    canWriteBusinessTruth: false,
  };
}

function buildActiveDraftReviewSections(
  source: WorkflowReviewHandoffSource,
): WorkflowReviewHandoffActiveDraftReviewSection[] {
  const validationInspector = source.workflowDraftValidationInspector;
  const executionPlanPreview = source.workflowExecutionPlanPreview;
  const runtimeReadinessInspector = source.workflowRuntimeReadinessInspector;
  const validationBlockedCount =
    validationInspector.structuralChecks.filter((check) => check.status === "blocked").length +
    validationInspector.contractChecks.filter((check) => check.status !== "passed").length +
    validationInspector.blockedCapabilityChecks.length;

  return [
    {
      sectionId: "active_draft_validation",
      label: "Active draft validation",
      labelMessage: { key: "active_draft_validation_label" },
      sourceSurface: "validation",
      status: validationStatusToHandoffStatus(validationInspector.validationStatus),
      primaryRef: validationInspector.inspectedDraftId,
      requestId: validationInspector.requestId,
      auditRef: validationInspector.auditRef,
      blockerCount: validationBlockedCount,
      summary: `Validation inspects ${validationInspector.structuralChecks.length} structural checks, ${validationInspector.contractChecks.length} contract checks, and ${validationInspector.blockedCapabilityChecks.length} blocked capability checks for the active draft.`,
      summaryMessage: {
        key: "active_draft_validation_summary", values: {
          value1: validationInspector.structuralChecks.length, value2: validationInspector.contractChecks.length, value3: validationInspector.blockedCapabilityChecks.length
        }
      },
      reviewerQuestion: "Which structural, contract, or blocked capability findings need review before any future implementation gate?",
      reviewerQuestionMessage: { key: "active_draft_validation_reviewerQuestion" },
      evidenceRefs: [
        ...validationInspector.structuralChecks.map((check) => check.checkId),
        ...validationInspector.contractChecks.map((check) => check.checkId),
        ...validationInspector.blockedCapabilityChecks.map((check) => check.checkId),
      ].slice(0, 8),
    },
    {
      sectionId: "active_draft_execution_plan",
      label: "Active draft execution plan preview",
      labelMessage: { key: "active_draft_execution_plan_label" },
      sourceSurface: "plan",
      status: executionPlanPreview.canRenderExecutionPlanPreview ? "review_required" : "blocked",
      primaryRef: executionPlanPreview.selectedDraftId,
      requestId: executionPlanPreview.requestId,
      auditRef: executionPlanPreview.auditRef,
      blockerCount: executionPlanPreview.blockedPlanReasons.length,
      summary: `Plan preview orders ${executionPlanPreview.stageOrder.length} offline stages, ${executionPlanPreview.providerProfileRequirements.length} provider requirements, and ${executionPlanPreview.confirmationAuditGates.length} gates without creating an executable plan.`,
      summaryMessage: {
        key: "active_draft_execution_plan_summary", values: {
          value1: executionPlanPreview.stageOrder.length, value2: executionPlanPreview.providerProfileRequirements.length, value3: executionPlanPreview.confirmationAuditGates.length
        }
      },
      reviewerQuestion: "Does the previewed stage order explain future execution intent while keeping runtime and writeback blocked?",
      reviewerQuestionMessage: { key: "active_draft_execution_plan_reviewerQuestion" },
      evidenceRefs: [
        ...executionPlanPreview.stageOrder.map((stage) => stage.stageId),
        ...executionPlanPreview.providerProfileRequirements.map((requirement) => requirement.requirementId),
        ...executionPlanPreview.confirmationAuditGates.map((gate) => gate.gateId),
      ].slice(0, 8),
    },
    {
      sectionId: "active_draft_runtime_readiness",
      label: "Active draft runtime readiness",
      labelMessage: { key: "active_draft_runtime_readiness_label" },
      sourceSurface: "readiness",
      status: "blocked",
      primaryRef: runtimeReadinessInspector.selectedDraftId,
      requestId: runtimeReadinessInspector.requestId,
      auditRef: runtimeReadinessInspector.auditRef,
      blockerCount: runtimeReadinessInspector.readinessBlockers.length,
      summary: `Runtime readiness keeps ${runtimeReadinessInspector.runtimePrerequisites.length} prerequisites and ${runtimeReadinessInspector.implementationGates.length} implementation gates visible while runtime start stays blocked.`,
      summaryMessage: {
        key: "active_draft_runtime_readiness_summary", values: {
          value1: runtimeReadinessInspector.runtimePrerequisites.length, value2: runtimeReadinessInspector.implementationGates.length
        }
      },
      reviewerQuestion: "Which executor, store, auth, confirmation, writeback, or replay prerequisites still block runtime readiness?",
      reviewerQuestionMessage: { key: "active_draft_runtime_readiness_reviewerQuestion" },
      evidenceRefs: [
        ...runtimeReadinessInspector.runtimePrerequisites.map((prerequisite) => prerequisite.prerequisiteId),
        ...runtimeReadinessInspector.implementationGates.map((gate) => gate.gateId),
      ].slice(0, 8),
    },
  ];
}

function buildNodeDesignerReviewRecord(
  source: WorkflowReviewHandoffSource,
): WorkflowReviewHandoffNodeDesignerReviewRecord {
  const draft = source.activeWorkflowDraft;
  const positionedNodeIds = new Set(draft.designerLayout.nodePositions.map((position) => position.nodeId));
  const positionedNodeCount = draft.nodes.filter((node) => positionedNodeIds.has(node.nodeId)).length;
  const defaultLayoutNodeCount = Math.max(0, draft.nodes.length - positionedNodeCount);
  const layoutPersistenceLabel =
    draft.designerLayout.persistence === "saved_draft_metadata" ? "saved draft metadata" : "active draft session";
  const validationOverlayCount = countNodeDesignerValidationOverlayItems(source);
  const inspectorFieldCount = countNodeDesignerInspectorFields(draft);
  const graphReviewFindings = buildNodeDesignerGraphReviewFindings(source);
  const sections = buildNodeDesignerReviewSections(
    source,
    layoutPersistenceLabel,
    positionedNodeCount,
    defaultLayoutNodeCount,
    validationOverlayCount,
    inspectorFieldCount,
  );
  const draftIdsMatch =
    draft.draftId === source.workflowWorkspaceReview.draftId &&
    draft.draftId === source.workflowDraftValidationInspector.inspectedDraftId &&
    draft.draftId === source.workflowExecutionPlanPreview.selectedDraftId &&
    draft.draftId === source.workflowRuntimeReadinessInspector.selectedDraftId;

  return {
    recordId: "node_designer_review_record",
    recordMode: "node_designer_advisory_only",
    draftId: draft.draftId,
    positionedNodeCount,
    defaultLayoutNodeCount,
    derivedEdgeCount: draft.edges.length,
    validationOverlayCount,
    inspectorFieldCount,
    graphReviewFindings,
    nodeTargetedFindingCount: countNodeDesignerGraphReviewFindings(graphReviewFindings, "node"),
    edgeTargetedFindingCount: countNodeDesignerGraphReviewFindings(graphReviewFindings, "edge"),
    graphLevelFindingCount: countNodeDesignerGraphReviewFindings(graphReviewFindings, "graph"),
    sections,
    canRenderNodeDesignerReviewRecord:
      draftIdsMatch &&
      draft.nodes.length > 0 &&
      graphReviewFindings.length > 0 &&
      sections.length === 4 &&
      sections.every((section) => section.requestId.length > 0 && section.auditRef.length > 0),
    canPersistLayout: false,
    canPersistEdgeKind: false,
    canPersistOverlay: false,
    canPersistInspectorState: false,
    canExportRecord: false,
    canSendRecord: false,
    canStartRuntime: false,
    canSubmitConfirmationDecision: false,
    canWriteBusinessTruth: false,
  };
}

function buildNodeDesignerReviewSections(
  source: WorkflowReviewHandoffSource,
  layoutPersistenceLabel: string,
  positionedNodeCount: number,
  defaultLayoutNodeCount: number,
  validationOverlayCount: number,
  inspectorFieldCount: number,
): WorkflowReviewHandoffNodeDesignerReviewSection[] {
  const draft = source.activeWorkflowDraft;
  const validationEvidenceRefs = nodeDesignerValidationEvidenceRefs(source);
  const inspectorEvidenceRefs = draft.nodes.flatMap((node) => [
    node.nodeId,
    node.providerRef || `provider_ref_empty_${node.nodeId}`,
    node.toolRef || `tool_ref_empty_${node.nodeId}`,
    node.ragRef || `rag_ref_empty_${node.nodeId}`,
  ]);

  return [
    {
      sectionId: "node_designer_canvas_layout",
      label: "Node designer canvas layout",
      labelMessage: { key: "node_designer_canvas_layout_label" },
      sourceSurface: "node_designer",
      status: "review_required",
      primaryRef: draft.draftId,
      requestId: draft.routeMetadata.requestId,
      auditRef: draft.routeMetadata.auditRef,
      itemCount: draft.nodes.length,
      summary: `Node Designer presents ${draft.nodes.length} active draft nodes with ${positionedNodeCount} ${layoutPersistenceLabel} positions and ${defaultLayoutNodeCount} default lane-derived positions.`,
      summaryMessage: {
        key: "node_designer_canvas_layout_summary", values: {
          value1: draft.nodes.length, value2: positionedNodeCount, value3: {
            key: source.activeWorkflowDraft.designerLayout.persistence === "saved_draft_metadata" ? "layoutSaved" : "layoutSession"
          }, value4: defaultLayoutNodeCount
        }
      },
      reviewerQuestion: "Does the visual layout help review the draft without implying runtime order or persisted schema state?",
      reviewerQuestionMessage: { key: "node_designer_canvas_layout_reviewerQuestion" },
      evidenceRefs: Array.from(
        new Set([
          ...draft.designerLayout.nodePositions.map((position) => position.nodeId),
          ...draft.nodes.map((node) => node.nodeId),
        ]),
      ).slice(0, 8),
    },
    {
      sectionId: "node_designer_validation_overlay",
      label: "Node designer validation overlay",
      labelMessage: { key: "node_designer_validation_overlay_label" },
      sourceSurface: "node_designer",
      status: validationStatusToHandoffStatus(source.workflowDraftValidationInspector.validationStatus),
      primaryRef: source.workflowDraftValidationInspector.inspectedDraftId,
      requestId: source.workflowDraftValidationInspector.requestId,
      auditRef: source.workflowDraftValidationInspector.auditRef,
      itemCount: validationOverlayCount,
      summary: `Canvas overlay review carries ${validationOverlayCount} validation, contract, and blocked capability items from the active draft inspector.`,
      summaryMessage: { key: "node_designer_validation_overlay_summary", values: { value1: validationOverlayCount } },
      reviewerQuestion: "Which overlay findings should the reviewer inspect before accepting the draft as reviewable?",
      reviewerQuestionMessage: { key: "node_designer_validation_overlay_reviewerQuestion" },
      evidenceRefs: validationEvidenceRefs.slice(0, 8),
    },
    {
      sectionId: "node_designer_inspector_state",
      label: "Node designer inspector state",
      labelMessage: { key: "node_designer_inspector_state_label" },
      sourceSurface: "node_designer",
      status: "review_required",
      primaryRef: draft.draftId,
      requestId: draft.routeMetadata.requestId,
      auditRef: draft.routeMetadata.auditRef,
      itemCount: inspectorFieldCount,
      summary: `Inspector handoff covers labels, summaries, provider / tool / RAG refs, contract fields, output mappings, risk markers, and confirmation markers for ${draft.nodes.length} nodes.`,
      summaryMessage: { key: "node_designer_inspector_state_summary", values: { value1: draft.nodes.length } },
      reviewerQuestion: "Do node inspector attributes explain provider, tool, RAG, contract, risk, and confirmation intent clearly enough for review?",
      reviewerQuestionMessage: { key: "node_designer_inspector_state_reviewerQuestion" },
      evidenceRefs: inspectorEvidenceRefs.slice(0, 8),
    },
    {
      sectionId: "node_designer_saved_draft_mapping",
      label: "Node designer saved draft mapping",
      labelMessage: { key: "node_designer_saved_draft_mapping_label" },
      sourceSurface: "node_designer",
      status: "offline_only",
      primaryRef: draft.draftId,
      requestId: draft.routeMetadata.requestId,
      auditRef: draft.routeMetadata.auditRef,
      itemCount: draft.edges.length,
      summary: `Saved draft mapping review keeps node attributes, contract fields, edge endpoints, condition summaries, and controlled layout metadata distinct from derived edge kind.`,
      summaryMessage: { key: "node_designer_saved_draft_mapping_summary" },
      reviewerQuestion: "Does the mapping make clear which canvas details are persisted and which remain advisory UI state?",
      reviewerQuestionMessage: { key: "node_designer_saved_draft_mapping_reviewerQuestion" },
      evidenceRefs: [
        "node_designer_saved_draft_mapping_v1",
        "node_designer_saved_draft_mapping_implementation_v1",
        ...draft.edges.map((edge) => edge.edgeId),
      ].slice(0, 8),
    },
  ];
}

function buildNodeDesignerGraphReviewFindings(
  source: WorkflowReviewHandoffSource,
): WorkflowReviewHandoffNodeDesignerGraphFinding[] {
  const draft = source.activeWorkflowDraft;
  const validationInspector = source.workflowDraftValidationInspector;
  const nodeIds = new Set(draft.nodes.map((node) => node.nodeId));
  const structuralFindings = validationInspector.structuralChecks
    .filter((check) => check.status !== "passed")
    .map<WorkflowReviewHandoffNodeDesignerGraphFinding>((check) => {
      const targetNodeIds = check.evidenceRefs.filter((nodeId) => nodeIds.has(nodeId));
      const targetEdgeIds = nodeDesignerEdgeIdsForTargetNodes(draft, targetNodeIds);
      const targetKind = targetEdgeIds.length > 0 ? "edge" : nodeDesignerTargetKind(targetNodeIds, targetEdgeIds);
      const targetRefs = nodeDesignerTargetRefs(targetKind, targetNodeIds, targetEdgeIds);
      return {
        findingId: `node_designer_graph_review_${check.checkId}`,
        label: check.label,
        labelMessage: check.labelMessage,
        sourceCheckId: check.checkId,
        targetKind,
        status: validationStatusToHandoffStatus(check.status),
        severity: check.severity,
        targetRefs,
        targetSummary: nodeDesignerTargetSummary(targetKind, targetNodeIds, targetEdgeIds),
        targetSummaryMessage: targetKind === "edge" ? {
          key: "targetEdges", values: {
            edges: targetEdgeIds.length, nodes: targetNodeIds.length
          }
        } : targetKind === "node" ? {
          key: "targetNodes", values: {
            nodes: targetNodeIds.length
          }
        } : {
          key: "targetGraph"
        },
        handoffPath: nodeDesignerGraphReviewHandoffPath(targetKind),
        handoffPathMessage: { key: targetKind === "node" ? "pathNode" : targetKind === "edge" ? "pathEdge" : "pathGraph" },
        handoffPathRefs: nodeDesignerGraphReviewHandoffPathRefs(targetKind, targetRefs),
        summary: check.summary,
        summaryMessage: check.summaryMessage,
        reviewerQuestion:
          "Does this graph finding identify the node or edge context a reviewer should inspect before handoff?",
        reviewerQuestionMessage: { key: "graphStructuralQuestion" },
        evidenceRefs: [check.checkId, ...targetRefs].slice(0, 8),
      };
    });
  const contractFindings = validationInspector.contractChecks
    .filter((check) => check.status !== "passed")
    .map<WorkflowReviewHandoffNodeDesignerGraphFinding>((check) => {
      const targetNodeIds = nodeDesignerContractTargetNodeIds(draft, check.checkId);
      const targetEdgeIds = nodeDesignerEdgeIdsForTargetNodes(draft, targetNodeIds);
      const targetKind = targetNodeIds.length > 0 ? "node" : nodeDesignerTargetKind(targetNodeIds, targetEdgeIds);
      const targetRefs = nodeDesignerTargetRefs(targetKind, targetNodeIds, targetEdgeIds);
      return {
        findingId: `node_designer_graph_review_${check.checkId}`,
        label: check.label,
        labelMessage: check.labelMessage,
        sourceCheckId: check.checkId,
        targetKind,
        status: validationStatusToHandoffStatus(check.status),
        severity: check.severity,
        targetRefs,
        targetSummary: nodeDesignerTargetSummary(targetKind, targetNodeIds, targetEdgeIds),
        targetSummaryMessage: targetKind === "edge" ? {
          key: "targetEdges", values: {
            edges: targetEdgeIds.length, nodes: targetNodeIds.length
          }
        } : targetKind === "node" ? {
          key: "targetNodes", values: {
            nodes: targetNodeIds.length
          }
        } : {
          key: "targetGraph"
        },
        handoffPath: nodeDesignerGraphReviewHandoffPath(targetKind),
        handoffPathMessage: { key: targetKind === "node" ? "pathNode" : targetKind === "edge" ? "pathEdge" : "pathGraph" },
        handoffPathRefs: nodeDesignerGraphReviewHandoffPathRefs(targetKind, targetRefs),
        summary: `${check.summary} Missing fields: ${
          check.missingFields.length > 0 ? check.missingFields.join(", ") : "none"
        }.`,
        summaryMessage: {
          key: "graphContractSummary", values: {
            value1: check.summaryMessage ?? check.summary, value2: check.missingFields.length > 0 ? check.missingFields.join(", ") : {
              status: "none"
            }
          }
        },
        reviewerQuestion:
          "Which node contract fields should remain visible before this draft can be reviewed as complete?",
        reviewerQuestionMessage: { key: "graphContractQuestion" },
        evidenceRefs: [check.checkId, ...check.missingFields, ...targetRefs].slice(0, 8),
      };
    });
  const blockedCapabilityFindings = validationInspector.blockedCapabilityChecks.map<WorkflowReviewHandoffNodeDesignerGraphFinding>((check) => ({
    findingId: `node_designer_graph_review_${check.checkId}`,
    label: check.label,
    labelMessage: check.labelMessage,
    sourceCheckId: check.checkId,
    targetKind: "graph" as const,
    status: "blocked" as const,
    severity: check.severity,
    targetRefs: [check.capabilityId],
    targetSummary: `Graph-level blocked capability: ${check.capabilityId}`,
    targetSummaryMessage: { key: "targetCapability", values: { capability: check.capabilityId } },
    handoffPath: nodeDesignerGraphReviewHandoffPath("graph"),
    handoffPathMessage: { key: "pathGraph" },
    handoffPathRefs: nodeDesignerGraphReviewHandoffPathRefs("graph", [check.capabilityId]),
    summary: check.summary,
    summaryMessage: check.summaryMessage,
    reviewerQuestion: "Which missing prerequisite keeps this graph-level capability blocked for the handoff?",
    reviewerQuestionMessage: { key: "graphCapabilityQuestion" },
    evidenceRefs: [check.checkId, check.capabilityId, check.auditRef],
  }));

  return [...structuralFindings, ...contractFindings, ...blockedCapabilityFindings];
}

function countNodeDesignerValidationOverlayItems(source: WorkflowReviewHandoffSource): number {
  const validationInspector = source.workflowDraftValidationInspector;
  return (
    validationInspector.structuralChecks.filter((check) => check.status !== "passed").length +
    validationInspector.contractChecks.filter((check) => check.status !== "passed").length +
    validationInspector.blockedCapabilityChecks.length
  );
}

function countNodeDesignerInspectorFields(draft: WorkflowDraftDesignerDraft): number {
  return draft.nodes.reduce(
    (total, node) =>
      total +
      9 +
      node.inputContractFields.length +
      node.outputContractFields.length +
      (node.requiresConfirmation ? 1 : 0),
    0,
  );
}

function nodeDesignerValidationEvidenceRefs(source: WorkflowReviewHandoffSource): string[] {
  const validationInspector = source.workflowDraftValidationInspector;
  return [
    ...validationInspector.structuralChecks
      .filter((check) => check.status !== "passed")
      .map((check) => check.checkId),
    ...validationInspector.contractChecks
      .filter((check) => check.status !== "passed")
      .map((check) => check.checkId),
    ...validationInspector.blockedCapabilityChecks.map((check) => check.checkId),
  ];
}

function countNodeDesignerGraphReviewFindings(
  findings: WorkflowReviewHandoffNodeDesignerGraphFinding[],
  targetKind: WorkflowReviewHandoffNodeDesignerGraphFinding["targetKind"],
): number {
  return findings.filter((finding) => finding.targetKind === targetKind).length;
}

function nodeDesignerContractTargetNodeIds(
  draft: WorkflowDraftDesignerDraft,
  checkId: string,
): string[] {
  if (checkId === "input_contract_fields") {
    return draft.nodes
      .filter((node) => node.lane === "context" || node.inputContractFields.length > 0)
      .map((node) => node.nodeId);
  }
  if (checkId === "output_contract_fields") {
    return draft.nodes
      .filter(
        (node) =>
          node.lane === "output" ||
          node.outputContractFields.length > 0 ||
          node.outputMappingSummary.trim().length > 0,
      )
      .map((node) => node.nodeId);
  }
  return [];
}

function nodeDesignerEdgeIdsForTargetNodes(
  draft: WorkflowDraftDesignerDraft,
  targetNodeIds: string[],
): string[] {
  const targetNodeIdSet = new Set(targetNodeIds);
  return draft.edges
    .filter((edge) => targetNodeIdSet.has(edge.fromNodeId) || targetNodeIdSet.has(edge.toNodeId))
    .map((edge) => edge.edgeId);
}

function nodeDesignerTargetKind(
  nodeIds: string[],
  edgeIds: string[],
): WorkflowReviewHandoffNodeDesignerGraphFinding["targetKind"] {
  if (edgeIds.length > 0) {
    return "edge";
  }
  if (nodeIds.length > 0) {
    return "node";
  }
  return "graph";
}

function nodeDesignerTargetRefs(
  targetKind: WorkflowReviewHandoffNodeDesignerGraphFinding["targetKind"],
  nodeIds: string[],
  edgeIds: string[],
): string[] {
  if (targetKind === "edge") {
    return [...edgeIds, ...nodeIds].slice(0, 8);
  }
  if (targetKind === "node") {
    return nodeIds.slice(0, 8);
  }
  return ["graph_level_review"];
}

function nodeDesignerTargetSummary(
  targetKind: WorkflowReviewHandoffNodeDesignerGraphFinding["targetKind"],
  nodeIds: string[],
  edgeIds: string[],
): string {
  if (targetKind === "edge") {
    return `${edgeIds.length} related edges / ${nodeIds.length} related nodes`;
  }
  if (targetKind === "node") {
    return `${nodeIds.length} related nodes`;
  }
  return "Graph-level review item";
}

function nodeDesignerGraphReviewHandoffPath(
  targetKind: WorkflowReviewHandoffNodeDesignerGraphFinding["targetKind"],
): string {
  if (targetKind === "node") {
    return "validation overlay / node inspector / saved draft mapping";
  }
  if (targetKind === "edge") {
    return "validation overlay / connected edge review / draft edge summary";
  }
  return "validation overlay / runtime readiness / decision blockers";
}

function nodeDesignerGraphReviewHandoffPathRefs(
  targetKind: WorkflowReviewHandoffNodeDesignerGraphFinding["targetKind"],
  targetRefs: string[],
): string[] {
  const sectionRefs =
    targetKind === "node"
      ? ["node_designer_validation_overlay", "node_designer_inspector_state"]
      : targetKind === "edge"
        ? ["node_designer_validation_overlay", "node_designer_saved_draft_mapping"]
        : ["node_designer_validation_overlay", "runtime_readiness", "decision_blockers"];

  return [...sectionRefs, ...targetRefs].slice(0, 8);
}

function buildRecipients(source: WorkflowReviewHandoffSource): WorkflowReviewHandoffRecipient[] {
  return [
    {
      recipientId: "workflow_owner",
      label: "Workflow owner",
      labelMessage: { key: "workflow_owner_label" },
      role: "Application and draft review",
      roleMessage: { key: "workflow_owner_role" },
      status: "review_required",
      handoffNeed: "Confirm the selected application, workflow definition, draft, run, and scenario belong together.",
      handoffNeedMessage: { key: "workflow_owner_handoffNeed" },
      evidenceRefs: [
        source.workflowWorkspaceReview.applicationId,
        source.workflowWorkspaceReview.workflowDefinitionId,
        source.workflowWorkspaceReview.draftId,
      ],
    },
    {
      recipientId: "policy_reviewer",
      label: "Policy reviewer",
      labelMessage: { key: "policy_reviewer_label" },
      role: "Risk and confirmation review",
      roleMessage: { key: "policy_reviewer_role" },
      status: "blocked",
      handoffNeed: "Review the blocked tool action, confirmation placeholder, and human review requirement.",
      handoffNeedMessage: { key: "policy_reviewer_handoffNeed" },
      evidenceRefs: [
        source.workflowBlockedActionPreview.toolActionId,
        source.workflowConfirmationPlaceholder.confirmationPlaceholderId,
      ],
    },
    {
      recipientId: "runtime_owner",
      label: "Runtime owner",
      labelMessage: { key: "runtime_owner_label" },
      role: "Implementation gate review",
      roleMessage: { key: "runtime_owner_role" },
      status: "blocked",
      handoffNeed: "Review executor, durable store, auth/store, writeback, and replay gates before any future runtime task.",
      handoffNeedMessage: { key: "runtime_owner_handoffNeed" },
      evidenceRefs: source.workflowRuntimeReadinessInspector.implementationGates.map((gate) => gate.gateId).slice(0, 4),
    },
    {
      recipientId: "control_plane_reviewer",
      label: "Control plane reviewer",
      labelMessage: { key: "control_plane_reviewer_label" },
      role: "Read-side evidence review",
      roleMessage: { key: "control_plane_reviewer_role" },
      status: "offline_only",
      handoffNeed: "Confirm the handoff is backed by read-side routes and does not rely on production API state.",
      handoffNeedMessage: { key: "control_plane_reviewer_handoffNeed" },
      evidenceRefs: source.workflowUserWorkspaceHome.routeEvidence.map((evidence) => evidence.evidenceId).slice(0, 4),
    },
  ];
}

function buildKeyFindings(
  source: WorkflowReviewHandoffSource,
  nodeDesignerReviewRecord: WorkflowReviewHandoffNodeDesignerReviewRecord,
): WorkflowReviewHandoffFinding[] {
  const scenarioStage = requireStage(source.workflowWorkspaceReview.reviewStages, "stage_scenario_context");
  const validationStage = requireStage(source.workflowWorkspaceReview.reviewStages, "stage_draft_validation");
  const planStage = requireStage(source.workflowWorkspaceReview.reviewStages, "stage_execution_plan");
  const readinessStage = requireStage(source.workflowWorkspaceReview.reviewStages, "stage_runtime_readiness");
  const stopLineStage = requireStage(source.workflowWorkspaceReview.reviewStages, "stage_stop_lines");
  const savedDraftConflictReviewSummary = source.savedDraftConflictReviewSummary ?? null;
  const savedDraftConflictFindings: WorkflowReviewHandoffFinding[] = savedDraftConflictReviewSummary
    ? [
        {
          findingId: "saved_draft_conflict_review",
          label: "Saved draft conflict review",
          labelMessage: { key: "saved_draft_conflict_review_label" },
          sourceSurface: "saved_draft_conflict",
          status:
            savedDraftConflictReviewSummary.status === "local_draft_continued"
              ? "review_required"
              : "blocked",
          summary: `${savedDraftConflictReviewSummary.summary} ${savedDraftConflictReviewSummary.localDraftPreservationSummary} Metadata state is ${savedDraftConflictReviewSummary.savedMetadataState}; open state is ${savedDraftConflictReviewSummary.openActionState}. Saved draft validation is ${savedDraftConflictReviewSummary.savedValidationState}; blocked capability count is ${
            savedDraftConflictReviewSummary.savedBlockedCapabilityCount ?? "not_loaded"
          }.`,
          summaryMessage: {
            key: "saved_draft_conflict_review_summary", values: {
              value1: (savedDraftConflictReviewSummary.summaryMessage ?? savedDraftConflictReviewSummary.summary), value2: (savedDraftConflictReviewSummary.localDraftPreservationSummaryMessage ?? savedDraftConflictReviewSummary.localDraftPreservationSummary), value3: {
                status: savedDraftConflictReviewSummary.savedMetadataState
              }, value4: {
                status: savedDraftConflictReviewSummary.openActionState
              }, value5: {
                status: savedDraftConflictReviewSummary.savedValidationState
              }, value6: savedDraftConflictReviewSummary.savedBlockedCapabilityCount ?? { status: "not_loaded" }
            }
          },
          evidenceRef: savedDraftConflictReviewSummary.reviewId,
          humanReviewQuestion: `${savedDraftConflictReviewSummary.reviewerQuestion} ${savedDraftConflictReviewSummary.nextReviewerStep}`,
          humanReviewQuestionMessage: {
            key: "saved_draft_conflict_review_humanReviewQuestion", values: {
              value1: (savedDraftConflictReviewSummary.reviewerQuestionMessage ?? savedDraftConflictReviewSummary.reviewerQuestion), value2: (savedDraftConflictReviewSummary.nextReviewerStepMessage ?? savedDraftConflictReviewSummary.nextReviewerStep)
            }
          },
        },
      ]
    : [];

  return [
    {
      findingId: "scenario_scope",
      label: "Scenario scope",
      labelMessage: { key: "scenario_scope_label" },
      sourceSurface: "scenario",
      status: scenarioStage.status,
      summary: source.workflowScenarioInspector.selectedScenario.intent,
      summaryMessage: source.workflowScenarioInspector.selectedScenario.intentMessage,
      evidenceRef: source.workflowScenarioInspector.selectedScenario.scenarioId,
      humanReviewQuestion: scenarioStage.reviewQuestion,
      humanReviewQuestionMessage: scenarioStage.reviewQuestionMessage,
    },
    {
      findingId: "review_chain",
      label: "Review chain",
      labelMessage: { key: "review_chain_label" },
      sourceSurface: "review",
      status: "offline_only",
      summary: source.workflowWorkspaceReview.reviewNarrative,
      summaryMessage: source.workflowWorkspaceReview.reviewNarrativeMessage,
      evidenceRef: source.workflowWorkspaceReview.pageId,
      humanReviewQuestion: "Does the selected context explain the current application, definition, run, draft, and scenario?",
      humanReviewQuestionMessage: { key: "review_chain_humanReviewQuestion" },
    },
    {
      findingId: "draft_validation",
      label: "Draft validation",
      labelMessage: { key: "draft_validation_label" },
      sourceSurface: "validation",
      status: validationStatusToHandoffStatus(source.workflowDraftValidationInspector.validationStatus),
      summary: `Active draft validation is ${source.workflowDraftValidationInspector.validationStatus} with ${source.workflowDraftValidationInspector.blockedCapabilityChecks.length} blocked capability checks.`,
      summaryMessage: {
        key: "draft_validation_summary", values: {
          value1: {
            status: source.workflowDraftValidationInspector.validationStatus
          }, value2: source.workflowDraftValidationInspector.blockedCapabilityChecks.length
        }
      },
      evidenceRef: source.workflowDraftValidationInspector.inspectedDraftId,
      humanReviewQuestion: validationStage.reviewQuestion,
      humanReviewQuestionMessage: validationStage.reviewQuestionMessage,
    },
    ...savedDraftConflictFindings,
    {
      findingId: "execution_plan_preview",
      label: "Execution plan preview",
      labelMessage: { key: "execution_plan_preview_label" },
      sourceSurface: "plan",
      status: source.workflowExecutionPlanPreview.canRenderExecutionPlanPreview ? "review_required" : "blocked",
      summary: `Active draft plan preview has ${source.workflowExecutionPlanPreview.stageOrder.length} stages, ${source.workflowExecutionPlanPreview.providerProfileRequirements.length} provider requirements, and ${source.workflowExecutionPlanPreview.blockedPlanReasons.length} blocked reasons.`,
      summaryMessage: {
        key: "execution_plan_preview_summary", values: {
          value1: source.workflowExecutionPlanPreview.stageOrder.length, value2: source.workflowExecutionPlanPreview.providerProfileRequirements.length, value3: source.workflowExecutionPlanPreview.blockedPlanReasons.length
        }
      },
      evidenceRef: source.workflowExecutionPlanPreview.selectedDraftId,
      humanReviewQuestion: planStage.reviewQuestion,
      humanReviewQuestionMessage: planStage.reviewQuestionMessage,
    },
    {
      findingId: "node_designer_review",
      label: "Node designer review",
      labelMessage: { key: "node_designer_review_label" },
      sourceSurface: "node_designer",
      status: nodeDesignerReviewRecord.canRenderNodeDesignerReviewRecord ? "review_required" : "blocked",
      summary: `Node Designer handoff carries ${nodeDesignerReviewRecord.sections.length} canvas review sections, ${nodeDesignerReviewRecord.positionedNodeCount} UI-only positions, ${nodeDesignerReviewRecord.defaultLayoutNodeCount} default positions, ${nodeDesignerReviewRecord.validationOverlayCount} overlay items, and ${nodeDesignerReviewRecord.graphReviewFindings.length} graph review findings.`,
      summaryMessage: {
        key: "node_designer_review_summary", values: {
          value1: nodeDesignerReviewRecord.sections.length, value2: nodeDesignerReviewRecord.positionedNodeCount, value3: nodeDesignerReviewRecord.defaultLayoutNodeCount, value4: nodeDesignerReviewRecord.validationOverlayCount, value5: nodeDesignerReviewRecord.graphReviewFindings.length
        }
      },
      evidenceRef: nodeDesignerReviewRecord.recordId,
      humanReviewQuestion: "Does the canvas review record make visual layout, overlay, inspector state, and saved draft mapping boundaries clear?",
      humanReviewQuestionMessage: { key: "node_designer_review_humanReviewQuestion" },
    },
    {
      findingId: "node_designer_graph_review",
      label: "Node designer graph review",
      labelMessage: { key: "node_designer_graph_review_label" },
      sourceSurface: "node_designer",
      status: nodeDesignerReviewRecord.canRenderNodeDesignerReviewRecord ? "review_required" : "blocked",
      summary: `Graph review groups ${nodeDesignerReviewRecord.nodeTargetedFindingCount} node-targeted, ${nodeDesignerReviewRecord.edgeTargetedFindingCount} edge-targeted, and ${nodeDesignerReviewRecord.graphLevelFindingCount} graph-level findings from validation overlay detail.`,
      summaryMessage: {
        key: "node_designer_graph_review_summary", values: {
          value1: nodeDesignerReviewRecord.nodeTargetedFindingCount, value2: nodeDesignerReviewRecord.edgeTargetedFindingCount, value3: nodeDesignerReviewRecord.graphLevelFindingCount
        }
      },
      evidenceRef: "node_designer_graph_review_findings",
      humanReviewQuestion: "Can the reviewer tell which nodes, edges, or graph-level blockers need attention before handoff?",
      humanReviewQuestionMessage: { key: "node_designer_graph_review_humanReviewQuestion" },
    },
    {
      findingId: "runtime_readiness",
      label: "Runtime readiness",
      labelMessage: { key: "runtime_readiness_label" },
      sourceSurface: "readiness",
      status: readinessStage.status,
      summary: readinessStage.summary,
      summaryMessage: readinessStage.summaryMessage,
      evidenceRef: source.workflowRuntimeReadinessInspector.readinessRouteId,
      humanReviewQuestion: readinessStage.reviewQuestion,
      humanReviewQuestionMessage: readinessStage.reviewQuestionMessage,
    },
    {
      findingId: "blocked_action",
      label: "Blocked action",
      labelMessage: { key: "blocked_action_label" },
      sourceSurface: "blocked_action",
      status: "blocked",
      summary: source.workflowBlockedActionPreview.policyReason,
      summaryMessage: source.workflowBlockedActionPreview.policyReasonMessage,
      evidenceRef: source.workflowBlockedActionPreview.toolActionId,
      humanReviewQuestion: "Which missing prerequisites prevent this candidate action from executing?",
      humanReviewQuestionMessage: { key: "blocked_action_humanReviewQuestion" },
    },
    {
      findingId: "confirmation_placeholder",
      label: "Confirmation placeholder",
      labelMessage: { key: "confirmation_placeholder_label" },
      sourceSurface: "confirmation",
      status: "blocked",
      summary: source.workflowConfirmationPlaceholder.disabledReason,
      summaryMessage: source.workflowConfirmationPlaceholder.disabledReasonMessage,
      evidenceRef: source.workflowConfirmationPlaceholder.confirmationPlaceholderId,
      humanReviewQuestion: "Which decision fields are visible, and why can no decision be submitted?",
      humanReviewQuestionMessage: { key: "confirmation_placeholder_humanReviewQuestion" },
    },
    {
      findingId: "stop_lines",
      label: "Stop lines",
      labelMessage: { key: "stop_lines_label" },
      sourceSurface: "stop_line",
      status: stopLineStage.status,
      summary: stopLineStage.summary,
      summaryMessage: stopLineStage.summaryMessage,
      evidenceRef: `${source.workflowWorkspaceReview.stopLines.length} locked stop lines`,
      evidenceRefMessage: { key: "lockedStopLineCount", values: { count: source.workflowWorkspaceReview.stopLines.length } },
      humanReviewQuestion: stopLineStage.reviewQuestion,
      humanReviewQuestionMessage: stopLineStage.reviewQuestionMessage,
    },
  ];
}

function buildEvidenceChecklist(
  source: WorkflowReviewHandoffSource,
  nodeDesignerReviewRecord: WorkflowReviewHandoffNodeDesignerReviewRecord,
): WorkflowReviewHandoffEvidence[] {
  const routeEvidence = source.workflowUserWorkspaceHome.routeEvidence.map((evidence) =>
    evidenceFromRoute(evidence),
  );
  const savedDraftConflictReviewSummary = source.savedDraftConflictReviewSummary ?? null;
  const savedDraftConflictEvidence: WorkflowReviewHandoffEvidence[] = savedDraftConflictReviewSummary
    ? [
        {
          evidenceId: "saved_draft_conflict_review",
          label: "Saved draft conflict review",
          labelMessage: { key: "saved_draft_conflict_review_label_2" },
          sourceSurface: "saved_draft_conflict",
          routeOrPageId: "workflow-draft-designer",
          requestId: savedDraftConflictReviewSummary.requestId,
          auditRef: savedDraftConflictReviewSummary.auditRef,
          status:
            savedDraftConflictReviewSummary.status === "local_draft_continued"
              ? "review_required"
              : "blocked",
          summary: `${savedDraftConflictReviewSummary.failureCode} keeps local draft ${savedDraftConflictReviewSummary.draftId} separate from saved version ${savedDraftConflictReviewSummary.savedDraftVersion}; metadata state is ${savedDraftConflictReviewSummary.savedMetadataState}; open state is ${savedDraftConflictReviewSummary.openActionState}; auto overwrite and auto merge stay disabled.`,
          summaryMessage: {
            key: "saved_draft_conflict_review_summary_2", values: {
              value1: savedDraftConflictReviewSummary.failureCode, value2: savedDraftConflictReviewSummary.draftId, value3: savedDraftConflictReviewSummary.savedDraftVersion, value4: {
                status: savedDraftConflictReviewSummary.savedMetadataState
              }, value5: {
                status: savedDraftConflictReviewSummary.openActionState
              }
            }
          },
        },
      ]
    : [];

  return [
    ...routeEvidence,
    {
      evidenceId: "review_workspace",
      label: "Review workspace",
      labelMessage: { key: "review_workspace_label" },
      sourceSurface: "review",
      routeOrPageId: source.workflowWorkspaceReview.pageId,
      requestId: source.workflowWorkspaceReview.requestId,
      auditRef: source.workflowWorkspaceReview.auditRef,
      status: source.workflowWorkspaceReview.canRenderWorkspaceReview ? "offline_only" : "blocked",
      summary: "Review workspace supplies selected context, stage order, relations, blockers, and stop lines.",
      summaryMessage: { key: "review_workspace_summary" },
    },
    {
      evidenceId: "scenario_inspector",
      label: "Scenario inspector",
      labelMessage: { key: "scenario_inspector_label" },
      sourceSurface: "scenario",
      routeOrPageId: source.workflowScenarioInspector.pageId,
      requestId: source.workflowScenarioInspector.selectedScenarioId,
      auditRef: source.workflowScenarioInspector.relationMap[0]?.auditRef ?? source.workflowScenarioInspector.scenarioMode,
      status: source.workflowScenarioInspector.canRenderScenarioInspector ? "offline_only" : "blocked",
      summary: "Scenario inspector supplies the advisory intent, input contract, expected output, and blocked reasons.",
      summaryMessage: { key: "scenario_inspector_summary" },
    },
    {
      evidenceId: "active_draft_validation_inspector",
      label: "Active draft validation inspector",
      labelMessage: { key: "active_draft_validation_inspector_label" },
      sourceSurface: "validation",
      routeOrPageId: source.workflowDraftValidationInspector.draftRouteId,
      requestId: source.workflowDraftValidationInspector.requestId,
      auditRef: source.workflowDraftValidationInspector.auditRef,
      status: validationStatusToHandoffStatus(source.workflowDraftValidationInspector.validationStatus),
      summary: "Validation inspector supplies active draft structural, contract, and blocked capability findings.",
      summaryMessage: { key: "active_draft_validation_inspector_summary" },
    },
    ...savedDraftConflictEvidence,
    {
      evidenceId: "active_draft_execution_plan_preview",
      label: "Active draft execution plan preview",
      labelMessage: { key: "active_draft_execution_plan_preview_label" },
      sourceSurface: "plan",
      routeOrPageId: source.workflowExecutionPlanPreview.draftRouteId,
      requestId: source.workflowExecutionPlanPreview.requestId,
      auditRef: source.workflowExecutionPlanPreview.auditRef,
      status: source.workflowExecutionPlanPreview.canRenderExecutionPlanPreview ? "review_required" : "blocked",
      summary: "Execution plan preview supplies active draft stage order, provider requirements, gates, and blocked reasons.",
      summaryMessage: { key: "active_draft_execution_plan_preview_summary" },
    },
    {
      evidenceId: "active_draft_runtime_readiness_inspector",
      label: "Active draft runtime readiness inspector",
      labelMessage: { key: "active_draft_runtime_readiness_inspector_label" },
      sourceSurface: "readiness",
      routeOrPageId: source.workflowRuntimeReadinessInspector.readinessRouteId,
      requestId: source.workflowRuntimeReadinessInspector.requestId,
      auditRef: source.workflowRuntimeReadinessInspector.auditRef,
      status: "blocked",
      summary: "Runtime readiness inspector supplies active draft prerequisites, blockers, and implementation gates.",
      summaryMessage: { key: "active_draft_runtime_readiness_inspector_summary" },
    },
    {
      evidenceId: "node_designer_review_handoff",
      label: "Node designer review handoff",
      labelMessage: { key: "node_designer_review_handoff_label" },
      sourceSurface: "node_designer",
      routeOrPageId: "workflow-node-designer",
      requestId: source.activeWorkflowDraft.routeMetadata.requestId,
      auditRef: source.activeWorkflowDraft.routeMetadata.auditRef,
      status: nodeDesignerReviewRecord.canRenderNodeDesignerReviewRecord ? "review_required" : "blocked",
      summary:
        "Node Designer supplies canvas layout, validation overlay, inspector state, and saved draft mapping review without persistence or runtime unlock.",
      summaryMessage: { key: "node_designer_review_handoff_summary" },
    },
    {
      evidenceId: "blocked_action_preview",
      label: "Blocked action preview",
      labelMessage: { key: "blocked_action_preview_label" },
      sourceSurface: "blocked_action",
      routeOrPageId: source.workflowBlockedActionPreview.draftRouteId,
      requestId: source.workflowBlockedActionPreview.requestId,
      auditRef: source.workflowBlockedActionPreview.auditRef,
      status: "blocked",
      summary: "Blocked action preview explains the candidate action and missing prerequisites without execution.",
      summaryMessage: { key: "blocked_action_preview_summary" },
    },
    {
      evidenceId: "confirmation_placeholder",
      label: "Confirmation placeholder",
      labelMessage: { key: "confirmation_placeholder_label_2" },
      sourceSurface: "confirmation",
      routeOrPageId: source.workflowConfirmationPlaceholder.draftRouteId,
      requestId: source.workflowConfirmationPlaceholder.requestId,
      auditRef: source.workflowConfirmationPlaceholder.auditRef,
      status: "blocked",
      summary: "Confirmation placeholder exposes future decision shape without accepting a decision.",
      summaryMessage: { key: "confirmation_placeholder_summary" },
    },
  ];
}

function evidenceFromRoute(evidence: WorkflowUserWorkspaceHomeRouteEvidence): WorkflowReviewHandoffEvidence {
  return {
    evidenceId: `route_${evidence.evidenceId}`,
    label: evidence.label,
    labelMessage: evidence.labelMessage,
    sourceSurface: "home",
    routeOrPageId: evidence.routeId,
    requestId: evidence.requestId,
    auditRef: evidence.auditRef,
    status: evidence.status,
    summary: evidence.summary,
    summaryMessage: evidence.summaryMessage,
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
    labelMessage: group.labelMessage,
    sourceSurface: group.sourceSurface,
    status: "blocked",
    missingPrerequisite: group.missingPrerequisites.join(", "),
    missingPrerequisiteMessage: group.missingPrerequisiteMessage,
    summary: group.exampleSummary,
    summaryMessage: group.exampleSummaryMessage,
    auditRefs: group.auditRefs,
  };
}

function blockerFromRuntimeBlocker(
  blocker: WorkflowRuntimeReadinessBlocker,
): WorkflowReviewHandoffDecisionBlocker {
  return {
    blockerId: `runtime_${blocker.blockerId}`,
    label: blocker.label,
    labelMessage: blocker.labelMessage,
    sourceSurface: blocker.area,
    status: "blocked",
    missingPrerequisite: blocker.missingPrerequisite,
    missingPrerequisiteMessage: blocker.missingPrerequisiteMessage,
    summary: blocker.summary,
    summaryMessage: blocker.summaryMessage,
    auditRefs: [blocker.auditRef],
  };
}

function buildBoundaryLocks(source: WorkflowReviewHandoffSource): WorkflowReviewHandoffBoundaryLock[] {
  const stopLineLocks = source.workflowUserWorkspaceHome.stopLines.slice(0, 8).map((stopLine) => ({
    boundaryId: stopLine.stopLineId,
    label: stopLine.label,
    labelMessage: stopLine.labelMessage,
    status: "locked" as const,
    summary: stopLine.summary,
    summaryMessage: stopLine.summaryMessage,
  }));
  const explicitLocks: WorkflowReviewHandoffBoundaryLock[] = [
    {
      boundaryId: "handoff_not_persisted",
      label: "Handoff persistence",
      labelMessage: { key: "handoff_not_persisted_label" },
      status: "locked",
      summary: "The handoff package is rendered from current offline view models and is not saved or exported.",
      summaryMessage: { key: "handoff_not_persisted_summary" },
    },
    {
      boundaryId: "handoff_no_confirmation_submission",
      label: "Confirmation submission",
      labelMessage: { key: "handoff_no_confirmation_submission_label" },
      status: "locked",
      summary: "Human review need is visible, but no approve, reject, defer, or submit path is connected.",
      summaryMessage: { key: "handoff_no_confirmation_submission_summary" },
    },
    {
      boundaryId: "handoff_no_runtime_unlock",
      label: "Runtime unlock",
      labelMessage: { key: "handoff_no_runtime_unlock_label" },
      status: "locked",
      summary: "The handoff does not unlock workflow execution, tool execution, writeback, replay, or resume.",
      summaryMessage: { key: "handoff_no_runtime_unlock_summary" },
    },
    {
      boundaryId: "node_designer_no_persisted_runtime_state",
      label: "Node Designer state",
      labelMessage: { key: "node_designer_no_persisted_runtime_state_label" },
      status: "locked",
      summary:
        "Node Designer layout, derived edge kind, validation overlay, and inspector state remain review context only; they do not create persisted runtime state.",
      summaryMessage: { key: "node_designer_no_persisted_runtime_state_summary" },
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

function validationStatusToHandoffStatus(
  status: WorkflowDraftValidationInspectorViewModel["validationStatus"],
): WorkflowReviewHandoffStatus {
  if (status === "passed") {
    return "ready";
  }
  if (status === "blocked") {
    return "blocked";
  }
  return "review_required";
}
