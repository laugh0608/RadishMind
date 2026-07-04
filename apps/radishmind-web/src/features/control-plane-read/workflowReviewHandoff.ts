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

export type WorkflowReviewHandoffEvidence = {
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

export type WorkflowReviewHandoffActiveDraftReviewSection = {
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

export type WorkflowReviewHandoffNodeDesignerReviewSection = {
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

export type WorkflowReviewHandoffNodeDesignerGraphFinding = {
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
      sourceSurface: "validation",
      status: validationStatusToHandoffStatus(validationInspector.validationStatus),
      primaryRef: validationInspector.inspectedDraftId,
      requestId: validationInspector.requestId,
      auditRef: validationInspector.auditRef,
      blockerCount: validationBlockedCount,
      summary: `Validation inspects ${validationInspector.structuralChecks.length} structural checks, ${validationInspector.contractChecks.length} contract checks, and ${validationInspector.blockedCapabilityChecks.length} blocked capability checks for the active draft.`,
      reviewerQuestion: "Which structural, contract, or blocked capability findings need review before any future implementation gate?",
      evidenceRefs: [
        ...validationInspector.structuralChecks.map((check) => check.checkId),
        ...validationInspector.contractChecks.map((check) => check.checkId),
        ...validationInspector.blockedCapabilityChecks.map((check) => check.checkId),
      ].slice(0, 8),
    },
    {
      sectionId: "active_draft_execution_plan",
      label: "Active draft execution plan preview",
      sourceSurface: "plan",
      status: executionPlanPreview.canRenderExecutionPlanPreview ? "review_required" : "blocked",
      primaryRef: executionPlanPreview.selectedDraftId,
      requestId: executionPlanPreview.requestId,
      auditRef: executionPlanPreview.auditRef,
      blockerCount: executionPlanPreview.blockedPlanReasons.length,
      summary: `Plan preview orders ${executionPlanPreview.stageOrder.length} offline stages, ${executionPlanPreview.providerProfileRequirements.length} provider requirements, and ${executionPlanPreview.confirmationAuditGates.length} gates without creating an executable plan.`,
      reviewerQuestion: "Does the previewed stage order explain future execution intent while keeping runtime and writeback blocked?",
      evidenceRefs: [
        ...executionPlanPreview.stageOrder.map((stage) => stage.stageId),
        ...executionPlanPreview.providerProfileRequirements.map((requirement) => requirement.requirementId),
        ...executionPlanPreview.confirmationAuditGates.map((gate) => gate.gateId),
      ].slice(0, 8),
    },
    {
      sectionId: "active_draft_runtime_readiness",
      label: "Active draft runtime readiness",
      sourceSurface: "readiness",
      status: "blocked",
      primaryRef: runtimeReadinessInspector.selectedDraftId,
      requestId: runtimeReadinessInspector.requestId,
      auditRef: runtimeReadinessInspector.auditRef,
      blockerCount: runtimeReadinessInspector.readinessBlockers.length,
      summary: `Runtime readiness keeps ${runtimeReadinessInspector.runtimePrerequisites.length} prerequisites and ${runtimeReadinessInspector.implementationGates.length} implementation gates visible while runtime start stays blocked.`,
      reviewerQuestion: "Which executor, store, auth, confirmation, writeback, or replay prerequisites still block runtime readiness?",
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
      sourceSurface: "node_designer",
      status: "review_required",
      primaryRef: draft.draftId,
      requestId: draft.routeMetadata.requestId,
      auditRef: draft.routeMetadata.auditRef,
      itemCount: draft.nodes.length,
      summary: `Node Designer presents ${draft.nodes.length} active draft nodes with ${positionedNodeCount} ${layoutPersistenceLabel} positions and ${defaultLayoutNodeCount} default lane-derived positions.`,
      reviewerQuestion: "Does the visual layout help review the draft without implying runtime order or persisted schema state?",
      evidenceRefs: [
        ...draft.designerLayout.nodePositions.map((position) => position.nodeId),
        ...draft.nodes.map((node) => node.nodeId),
      ].slice(0, 8),
    },
    {
      sectionId: "node_designer_validation_overlay",
      label: "Node designer validation overlay",
      sourceSurface: "node_designer",
      status: validationStatusToHandoffStatus(source.workflowDraftValidationInspector.validationStatus),
      primaryRef: source.workflowDraftValidationInspector.inspectedDraftId,
      requestId: source.workflowDraftValidationInspector.requestId,
      auditRef: source.workflowDraftValidationInspector.auditRef,
      itemCount: validationOverlayCount,
      summary: `Canvas overlay review carries ${validationOverlayCount} validation, contract, and blocked capability items from the active draft inspector.`,
      reviewerQuestion: "Which overlay findings should the reviewer inspect before accepting the draft as reviewable?",
      evidenceRefs: validationEvidenceRefs.slice(0, 8),
    },
    {
      sectionId: "node_designer_inspector_state",
      label: "Node designer inspector state",
      sourceSurface: "node_designer",
      status: "review_required",
      primaryRef: draft.draftId,
      requestId: draft.routeMetadata.requestId,
      auditRef: draft.routeMetadata.auditRef,
      itemCount: inspectorFieldCount,
      summary: `Inspector handoff covers labels, summaries, provider / tool / RAG refs, contract fields, output mappings, risk markers, and confirmation markers for ${draft.nodes.length} nodes.`,
      reviewerQuestion: "Do node inspector attributes explain provider, tool, RAG, contract, risk, and confirmation intent clearly enough for review?",
      evidenceRefs: inspectorEvidenceRefs.slice(0, 8),
    },
    {
      sectionId: "node_designer_saved_draft_mapping",
      label: "Node designer saved draft mapping",
      sourceSurface: "node_designer",
      status: "offline_only",
      primaryRef: draft.draftId,
      requestId: draft.routeMetadata.requestId,
      auditRef: draft.routeMetadata.auditRef,
      itemCount: draft.edges.length,
      summary: `Saved draft mapping review keeps node attributes, contract fields, edge endpoints, condition summaries, and controlled layout metadata distinct from derived edge kind.`,
      reviewerQuestion: "Does the mapping make clear which canvas details are persisted and which remain advisory UI state?",
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
    .map((check) => {
      const targetNodeIds = check.evidenceRefs.filter((nodeId) => nodeIds.has(nodeId));
      const targetEdgeIds = nodeDesignerEdgeIdsForTargetNodes(draft, targetNodeIds);
      const targetKind = targetEdgeIds.length > 0 ? "edge" : nodeDesignerTargetKind(targetNodeIds, targetEdgeIds);
      const targetRefs = nodeDesignerTargetRefs(targetKind, targetNodeIds, targetEdgeIds);
      return {
        findingId: `node_designer_graph_review_${check.checkId}`,
        label: check.label,
        sourceCheckId: check.checkId,
        targetKind,
        status: validationStatusToHandoffStatus(check.status),
        severity: check.severity,
        targetRefs,
        targetSummary: nodeDesignerTargetSummary(targetKind, targetNodeIds, targetEdgeIds),
        handoffPath: nodeDesignerGraphReviewHandoffPath(targetKind),
        handoffPathRefs: nodeDesignerGraphReviewHandoffPathRefs(targetKind, targetRefs),
        summary: check.summary,
        reviewerQuestion:
          "Does this graph finding identify the node or edge context a reviewer should inspect before handoff?",
        evidenceRefs: [check.checkId, ...targetRefs].slice(0, 8),
      };
    });
  const contractFindings = validationInspector.contractChecks
    .filter((check) => check.status !== "passed")
    .map((check) => {
      const targetNodeIds = nodeDesignerContractTargetNodeIds(draft, check.checkId);
      const targetEdgeIds = nodeDesignerEdgeIdsForTargetNodes(draft, targetNodeIds);
      const targetKind = targetNodeIds.length > 0 ? "node" : nodeDesignerTargetKind(targetNodeIds, targetEdgeIds);
      const targetRefs = nodeDesignerTargetRefs(targetKind, targetNodeIds, targetEdgeIds);
      return {
        findingId: `node_designer_graph_review_${check.checkId}`,
        label: check.label,
        sourceCheckId: check.checkId,
        targetKind,
        status: validationStatusToHandoffStatus(check.status),
        severity: check.severity,
        targetRefs,
        targetSummary: nodeDesignerTargetSummary(targetKind, targetNodeIds, targetEdgeIds),
        handoffPath: nodeDesignerGraphReviewHandoffPath(targetKind),
        handoffPathRefs: nodeDesignerGraphReviewHandoffPathRefs(targetKind, targetRefs),
        summary: `${check.summary} Missing fields: ${
          check.missingFields.length > 0 ? check.missingFields.join(", ") : "none"
        }.`,
        reviewerQuestion:
          "Which node contract fields should remain visible before this draft can be reviewed as complete?",
        evidenceRefs: [check.checkId, ...check.missingFields, ...targetRefs].slice(0, 8),
      };
    });
  const blockedCapabilityFindings = validationInspector.blockedCapabilityChecks.map((check) => ({
    findingId: `node_designer_graph_review_${check.checkId}`,
    label: check.label,
    sourceCheckId: check.checkId,
    targetKind: "graph" as const,
    status: "blocked" as const,
    severity: check.severity,
    targetRefs: [check.capabilityId],
    targetSummary: `Graph-level blocked capability: ${check.capabilityId}`,
    handoffPath: nodeDesignerGraphReviewHandoffPath("graph"),
    handoffPathRefs: nodeDesignerGraphReviewHandoffPathRefs("graph", [check.capabilityId]),
    summary: check.summary,
    reviewerQuestion: "Which missing prerequisite keeps this graph-level capability blocked for the handoff?",
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
          sourceSurface: "saved_draft_conflict",
          status:
            savedDraftConflictReviewSummary.status === "local_draft_continued"
              ? "review_required"
              : "blocked",
          summary: `${savedDraftConflictReviewSummary.summary} ${savedDraftConflictReviewSummary.localDraftPreservationSummary} Metadata state is ${savedDraftConflictReviewSummary.savedMetadataState}; restore state is ${savedDraftConflictReviewSummary.restoreActionState}. Saved draft validation is ${savedDraftConflictReviewSummary.savedValidationState}; blocked capability count is ${
            savedDraftConflictReviewSummary.savedBlockedCapabilityCount ?? "not_loaded"
          }.`,
          evidenceRef: savedDraftConflictReviewSummary.reviewId,
          humanReviewQuestion: `${savedDraftConflictReviewSummary.reviewerQuestion} ${savedDraftConflictReviewSummary.nextReviewerStep}`,
        },
      ]
    : [];

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
      sourceSurface: "validation",
      status: validationStatusToHandoffStatus(source.workflowDraftValidationInspector.validationStatus),
      summary: `Active draft validation is ${source.workflowDraftValidationInspector.validationStatus} with ${source.workflowDraftValidationInspector.blockedCapabilityChecks.length} blocked capability checks.`,
      evidenceRef: source.workflowDraftValidationInspector.inspectedDraftId,
      humanReviewQuestion: validationStage.reviewQuestion,
    },
    ...savedDraftConflictFindings,
    {
      findingId: "execution_plan_preview",
      label: "Execution plan preview",
      sourceSurface: "plan",
      status: source.workflowExecutionPlanPreview.canRenderExecutionPlanPreview ? "review_required" : "blocked",
      summary: `Active draft plan preview has ${source.workflowExecutionPlanPreview.stageOrder.length} stages, ${source.workflowExecutionPlanPreview.providerProfileRequirements.length} provider requirements, and ${source.workflowExecutionPlanPreview.blockedPlanReasons.length} blocked reasons.`,
      evidenceRef: source.workflowExecutionPlanPreview.selectedDraftId,
      humanReviewQuestion: planStage.reviewQuestion,
    },
    {
      findingId: "node_designer_review",
      label: "Node designer review",
      sourceSurface: "node_designer",
      status: nodeDesignerReviewRecord.canRenderNodeDesignerReviewRecord ? "review_required" : "blocked",
      summary: `Node Designer handoff carries ${nodeDesignerReviewRecord.sections.length} canvas review sections, ${nodeDesignerReviewRecord.positionedNodeCount} UI-only positions, ${nodeDesignerReviewRecord.defaultLayoutNodeCount} default positions, ${nodeDesignerReviewRecord.validationOverlayCount} overlay items, and ${nodeDesignerReviewRecord.graphReviewFindings.length} graph review findings.`,
      evidenceRef: nodeDesignerReviewRecord.recordId,
      humanReviewQuestion: "Does the canvas review record make visual layout, overlay, inspector state, and saved draft mapping boundaries clear?",
    },
    {
      findingId: "node_designer_graph_review",
      label: "Node designer graph review",
      sourceSurface: "node_designer",
      status: nodeDesignerReviewRecord.canRenderNodeDesignerReviewRecord ? "review_required" : "blocked",
      summary: `Graph review groups ${nodeDesignerReviewRecord.nodeTargetedFindingCount} node-targeted, ${nodeDesignerReviewRecord.edgeTargetedFindingCount} edge-targeted, and ${nodeDesignerReviewRecord.graphLevelFindingCount} graph-level findings from validation overlay detail.`,
      evidenceRef: "node_designer_graph_review_findings",
      humanReviewQuestion: "Can the reviewer tell which nodes, edges, or graph-level blockers need attention before handoff?",
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
          sourceSurface: "saved_draft_conflict",
          routeOrPageId: "workflow-draft-designer",
          requestId: savedDraftConflictReviewSummary.requestId,
          auditRef: savedDraftConflictReviewSummary.auditRef,
          status:
            savedDraftConflictReviewSummary.status === "local_draft_continued"
              ? "review_required"
              : "blocked",
          summary: `${savedDraftConflictReviewSummary.failureCode} keeps local draft ${savedDraftConflictReviewSummary.draftId} separate from saved version ${savedDraftConflictReviewSummary.savedDraftVersion}; metadata state is ${savedDraftConflictReviewSummary.savedMetadataState}; restore state is ${savedDraftConflictReviewSummary.restoreActionState}; auto overwrite and auto merge stay disabled.`,
        },
      ]
    : [];

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
      evidenceId: "active_draft_validation_inspector",
      label: "Active draft validation inspector",
      sourceSurface: "validation",
      routeOrPageId: source.workflowDraftValidationInspector.draftRouteId,
      requestId: source.workflowDraftValidationInspector.requestId,
      auditRef: source.workflowDraftValidationInspector.auditRef,
      status: validationStatusToHandoffStatus(source.workflowDraftValidationInspector.validationStatus),
      summary: "Validation inspector supplies active draft structural, contract, and blocked capability findings.",
    },
    ...savedDraftConflictEvidence,
    {
      evidenceId: "active_draft_execution_plan_preview",
      label: "Active draft execution plan preview",
      sourceSurface: "plan",
      routeOrPageId: source.workflowExecutionPlanPreview.draftRouteId,
      requestId: source.workflowExecutionPlanPreview.requestId,
      auditRef: source.workflowExecutionPlanPreview.auditRef,
      status: source.workflowExecutionPlanPreview.canRenderExecutionPlanPreview ? "review_required" : "blocked",
      summary: "Execution plan preview supplies active draft stage order, provider requirements, gates, and blocked reasons.",
    },
    {
      evidenceId: "active_draft_runtime_readiness_inspector",
      label: "Active draft runtime readiness inspector",
      sourceSurface: "readiness",
      routeOrPageId: source.workflowRuntimeReadinessInspector.readinessRouteId,
      requestId: source.workflowRuntimeReadinessInspector.requestId,
      auditRef: source.workflowRuntimeReadinessInspector.auditRef,
      status: "blocked",
      summary: "Runtime readiness inspector supplies active draft prerequisites, blockers, and implementation gates.",
    },
    {
      evidenceId: "node_designer_review_handoff",
      label: "Node designer review handoff",
      sourceSurface: "node_designer",
      routeOrPageId: "workflow-node-designer",
      requestId: source.activeWorkflowDraft.routeMetadata.requestId,
      auditRef: source.activeWorkflowDraft.routeMetadata.auditRef,
      status: nodeDesignerReviewRecord.canRenderNodeDesignerReviewRecord ? "review_required" : "blocked",
      summary:
        "Node Designer supplies canvas layout, validation overlay, inspector state, and saved draft mapping review without persistence or runtime unlock.",
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
    {
      boundaryId: "node_designer_no_persisted_runtime_state",
      label: "Node Designer state",
      status: "locked",
      summary:
        "Node Designer layout, derived edge kind, validation overlay, and inspector state remain review context only; they do not create persisted runtime state.",
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
