import { useEffect, useMemo, useState, type KeyboardEvent } from "react";

import {
  buildAdminTenantOverviewViewModel,
  type AdminTenantOverviewFact,
  type AdminTenantOverviewStatePreview,
} from "../features/control-plane-read/adminTenantOverview";
import {
  buildAdminAuditLogViewModel,
  type AdminAuditEventRow,
  type AdminAuditLogMetric,
  type AdminAuditLogStatePreview,
} from "../features/control-plane-read/adminAuditLog";
import { buildAdminOperationsReviewViewModel } from "../features/control-plane-read/adminOperationsReview";
import { AdminOperationsReviewPanel } from "../features/control-plane-read/adminOperationsReviewPanel";
import { buildAdminProviderDeploymentReviewViewModel } from "../features/control-plane-read/adminProviderDeploymentReview";
import { AdminProviderDeploymentReviewPanel } from "../features/control-plane-read/adminProviderDeploymentReviewPanel";
import {
  initialControlPlaneReadDevLiveLoadState,
  loadControlPlaneReadDevLiveCollections,
  readControlPlaneReadDevLiveConfig,
  type ControlPlaneReadDevLiveLoadState,
} from "../features/control-plane-read/devLiveReadConsumer";
import {
  initialWorkflowSavedDraftConsumerState,
  initialWorkflowSavedDraftListState,
  listWorkflowDraftDevRecords,
  nextWorkflowSavedDraftExpectedVersion,
  readWorkflowDraftDevRecord,
  readWorkflowSavedDraftConsumerConfig,
  restoreWorkflowDraftDevRecord,
  saveWorkflowDraftDevRecord,
  validateWorkflowDraftDevRecord,
  type WorkflowSavedDraftListState,
  type WorkflowSavedDraftSummary,
  type WorkflowSavedDraftConsumerState,
} from "../features/control-plane-read/savedWorkflowDraftConsumer";
import { buildModelGatewayOverviewViewModel } from "../features/control-plane-read/modelGatewayOverview";
import { ModelGatewayOverviewPanel } from "../features/control-plane-read/modelGatewayOverviewPanel";
import { buildModelGatewayRouteEvidenceViewModel } from "../features/control-plane-read/modelGatewayRouteEvidence";
import { ModelGatewayRouteEvidencePanel } from "../features/control-plane-read/modelGatewayRouteEvidencePanel";
import { buildModelGatewayUsageAuditEvidenceViewModel } from "../features/control-plane-read/modelGatewayUsageAuditEvidence";
import { ModelGatewayUsageAuditEvidencePanel } from "../features/control-plane-read/modelGatewayUsageAuditEvidencePanel";
import { buildModelGatewayEvidenceReviewViewModel } from "../features/control-plane-read/modelGatewayEvidenceReview";
import { ModelGatewayEvidenceReviewPanel } from "../features/control-plane-read/modelGatewayEvidenceReviewPanel";
import {
  buildControlPlaneReadShellViewModel,
  type ControlPlaneReadRouteCard,
  type ControlPlaneReadStatePreview,
} from "../features/control-plane-read/readShell";
import {
  buildWorkspaceApplicationsViewModel,
  type WorkspaceApplicationRow,
  type WorkspaceApplicationsMetric,
  type WorkspaceApplicationsStatePreview,
} from "../features/control-plane-read/workspaceApplications";
import {
  type WorkflowApplicationBlockedCapabilityPreview,
  type WorkflowApplicationDetailViewModel,
  type WorkflowApplicationRiskSummary,
  type WorkflowApplicationRouteMetadata,
} from "../features/control-plane-read/workflowApplicationDetail";
import {
  buildWorkspaceApiKeysViewModel,
  type WorkspaceApiKeyRow,
  type WorkspaceApiKeysMetric,
  type WorkspaceApiKeysStatePreview,
} from "../features/control-plane-read/workspaceApiKeys";
import {
  buildWorkspaceUsageQuotaViewModel,
  type WorkspaceUsageQuotaLimit,
  type WorkspaceUsageQuotaSnapshot,
  type WorkspaceUsageQuotaStatePreview,
} from "../features/control-plane-read/workspaceUsageQuota";
import {
  buildWorkspaceWorkflowDefinitionsViewModel,
  type WorkspaceWorkflowDefinitionRow,
  type WorkspaceWorkflowDefinitionsMetric,
  type WorkspaceWorkflowDefinitionsStatePreview,
} from "../features/control-plane-read/workspaceWorkflowDefinitions";
import {
  type WorkflowDefinitionBlockedActionPreview,
  type WorkflowDefinitionDetailEdge,
  type WorkflowDefinitionDetailNode,
  type WorkflowDefinitionDetailSchemaSummary,
  type WorkflowDefinitionDetailViewModel,
} from "../features/control-plane-read/workflowDefinitionDetail";
import {
  type WorkflowDraftDesignerBlockedCapability,
  type WorkflowDraftDesignerDraft,
  type WorkflowDraftDesignerEdge,
  type WorkflowDraftDesignerNode,
  type WorkflowDraftDesignerReadiness,
  type WorkflowDraftDesignerRisk,
  type WorkflowDraftDesignerTemplate,
  type WorkflowDraftDesignerViewModel,
} from "../features/control-plane-read/workflowDraftDesigner";
import {
  type WorkflowDraftBlockedCapabilityCheck,
  type WorkflowDraftContractCheck,
  type WorkflowDraftStructuralCheck,
  type WorkflowDraftValidationInspectorViewModel,
  type WorkflowDraftValidationSummary,
} from "../features/control-plane-read/workflowDraftValidationInspector";
import {
  type WorkflowExecutionPlanBlockedReason,
  type WorkflowExecutionPlanGate,
  type WorkflowExecutionPlanNodeMapping,
  type WorkflowExecutionPlanPreviewViewModel,
  type WorkflowExecutionPlanProviderRequirement,
  type WorkflowExecutionPlanStage,
  type WorkflowExecutionPlanSummary,
} from "../features/control-plane-read/workflowExecutionPlanPreview";
import {
  type WorkflowRuntimeReadinessBlocker,
  type WorkflowRuntimeReadinessGate,
  type WorkflowRuntimeReadinessInspectorViewModel,
  type WorkflowRuntimeReadinessPrerequisite,
  type WorkflowRuntimeReadinessStatus,
  type WorkflowRuntimeReadinessSummary,
} from "../features/control-plane-read/workflowRuntimeReadinessInspector";
import { WorkflowReviewHandoffPanel } from "../features/control-plane-read/workflowReviewHandoffPanel";
import {
  type WorkflowSurfaceOverviewBlockedCapability,
  type WorkflowSurfaceOverviewMetric,
  type WorkflowSurfaceOverviewRelation,
  type WorkflowSurfaceOverviewStatus,
  type WorkflowSurfaceOverviewStopLine,
  type WorkflowSurfaceOverviewViewModel,
} from "../features/control-plane-read/workflowSurfaceOverview";
import { WorkflowWorkspaceReviewPanel } from "../features/control-plane-read/workflowWorkspaceReviewPanel";
import { WorkflowUserWorkspaceHomePanel } from "../features/control-plane-read/workflowUserWorkspaceHomePanel";
import {
  buildWorkflowWorkspaceContextViewModel,
  selectionForApplication,
  selectionForDraft,
  selectionForRun,
  selectionForWorkflowDefinition,
} from "../features/control-plane-read/workflowWorkspaceContext";
import {
  type WorkflowScenario,
  type WorkflowScenarioBlockedReason,
  type WorkflowScenarioExpectedOutput,
  type WorkflowScenarioInputField,
  type WorkflowScenarioInspectorViewModel,
  type WorkflowScenarioRelation,
  type WorkflowScenarioStatus,
  type WorkflowScenarioStopLine,
  type WorkflowScenarioSummary,
} from "../features/control-plane-read/workflowScenarioInspector";
import {
  buildWorkspaceRunHistoryViewModel,
  type WorkspaceRunHistoryMetric,
  type WorkspaceRunHistoryStatePreview,
  type WorkspaceRunRecordRow,
} from "../features/control-plane-read/workspaceRunHistory";
import {
  type WorkflowRunDetailGuardPreview,
  type WorkflowRunDetailSummary,
  type WorkflowRunDetailTimelineEvent,
  type WorkflowRunDetailViewModel,
} from "../features/control-plane-read/workflowRunDetail";
import {
  type WorkflowBlockedActionAuditStep,
  type WorkflowBlockedActionPreviewViewModel,
  type WorkflowBlockedActionRequirement,
  type WorkflowConfirmationPlaceholderPreview,
} from "../features/control-plane-read/workflowBlockedActionPreview";
import {
  type WorkflowConfirmationDecisionField,
  type WorkflowConfirmationPlaceholderPrerequisite,
  type WorkflowConfirmationPlaceholderViewModel,
} from "../features/control-plane-read/workflowConfirmationPlaceholder";
import type {
  ControlPlaneReadCollectionViewModel,
  ControlPlaneReadRouteId,
} from "../../../../contracts/typescript/control-plane-read-api";

const shell = buildControlPlaneReadShellViewModel();
const devLiveConfig = readControlPlaneReadDevLiveConfig();
const savedDraftConsumerConfig = readWorkflowSavedDraftConsumerConfig();

type ControlPlaneReadCollectionsByRoute = Partial<
  Record<ControlPlaneReadRouteId, ControlPlaneReadCollectionViewModel>
>;

type WorkflowDraftNodeMoveDirection = "up" | "down";

type WorkflowDraftNodeTypeOption = {
  nodeType: WorkflowDraftDesignerNode["nodeType"];
  lane: WorkflowDraftDesignerNode["lane"];
  label: string;
  summary: string;
};

const WORKFLOW_DRAFT_NODE_TYPE_OPTIONS: WorkflowDraftNodeTypeOption[] = [
  {
    nodeType: "prompt",
    lane: "context",
    label: "Context",
    summary: "Collects sanitized workspace, selection, and diagnostic context.",
  },
  {
    nodeType: "llm",
    lane: "model",
    label: "Model",
    summary: "Adds advisory reasoning without direct execution.",
  },
  {
    nodeType: "condition",
    lane: "policy",
    label: "Policy",
    summary: "Keeps risk and confirmation gates explicit.",
  },
  {
    nodeType: "http_tool",
    lane: "preview",
    label: "Preview",
    summary: "Models tool preview metadata while execution stays blocked.",
  },
  {
    nodeType: "output",
    lane: "output",
    label: "Output",
    summary: "Adds reviewable output or audit projection nodes.",
  },
];

export function App() {
  const [devLiveState, setDevLiveState] = useState<ControlPlaneReadDevLiveLoadState>(() =>
    initialControlPlaneReadDevLiveLoadState(devLiveConfig),
  );
  const [selectedApplicationRef, setSelectedApplicationRef] = useState<string | null>(null);
  const [selectedWorkflowDefinitionId, setSelectedWorkflowDefinitionId] = useState<string | null>(null);
  const [selectedRunId, setSelectedRunId] = useState<string | null>(null);
  const [selectedWorkflowDraftId, setSelectedWorkflowDraftId] = useState<string | null>(null);
  const [selectedWorkflowScenarioId, setSelectedWorkflowScenarioId] = useState<string | null>(null);
  const [savedDraftConsumerState, setSavedDraftConsumerState] = useState<WorkflowSavedDraftConsumerState>(() =>
    initialWorkflowSavedDraftConsumerState(savedDraftConsumerConfig),
  );
  const [savedDraftListState, setSavedDraftListState] = useState<WorkflowSavedDraftListState>(() =>
    initialWorkflowSavedDraftListState(savedDraftConsumerConfig),
  );
  const [workspaceCreatedDrafts, setWorkspaceCreatedDrafts] = useState<WorkflowDraftDesignerDraft[]>([]);
  const [editableWorkflowDraft, setEditableWorkflowDraft] = useState<WorkflowDraftDesignerDraft | null>(null);
  const [workflowDraftEditDirty, setWorkflowDraftEditDirty] = useState(false);

  useEffect(() => {
    if (devLiveConfig.mode !== "dev_live_http") {
      return;
    }
    let cancelled = false;
    setDevLiveState({
      status: "loading",
      mode: "dev_live_http",
      message: "Loading fake-store-backed read routes over dev HTTP.",
    });
    loadControlPlaneReadDevLiveCollections(devLiveConfig)
      .then((collections) => {
        if (cancelled) {
          return;
        }
        setDevLiveState({
          status: "ready",
          mode: "dev_live_http",
          message: "Dev live read consumer loaded fake-store-backed HTTP envelopes.",
          collections,
        });
      })
      .catch((error: unknown) => {
        if (cancelled) {
          return;
        }
        setDevLiveState({
          status: "failed",
          mode: "dev_live_http",
          message: error instanceof Error ? error.message : "Dev live read consumer failed.",
        });
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const liveCollections: ControlPlaneReadCollectionsByRoute =
    devLiveState.status === "ready" ? devLiveState.collections : {};
  const tenantOverview = useMemo(
    () => buildAdminTenantOverviewViewModel(liveCollections["tenant-summary-route"]),
    [liveCollections],
  );
  const adminAuditLog = useMemo(
    () => buildAdminAuditLogViewModel(liveCollections["audit-summary-list-route"]),
    [liveCollections],
  );
  const workspaceApplications = useMemo(
    () => buildWorkspaceApplicationsViewModel(liveCollections["application-summary-list-route"]),
    [liveCollections],
  );
  const workspaceApiKeys = useMemo(
    () => buildWorkspaceApiKeysViewModel(liveCollections["api-key-summary-list-route"]),
    [liveCollections],
  );
  const workspaceUsageQuota = useMemo(
    () => buildWorkspaceUsageQuotaViewModel(liveCollections["quota-summary-route"]),
    [liveCollections],
  );
  const workspaceWorkflowDefinitions = useMemo(
    () => buildWorkspaceWorkflowDefinitionsViewModel(liveCollections["workflow-definition-summary-list-route"]),
    [liveCollections],
  );
  const workspaceRunHistory = useMemo(
    () => buildWorkspaceRunHistoryViewModel(liveCollections["run-record-summary-list-route"]),
    [liveCollections],
  );
  const modelGatewayOverview = useMemo(
    () =>
      buildModelGatewayOverviewViewModel({
        readShell: shell,
        workspaceApiKeys,
        workspaceUsageQuota,
        workspaceRunHistory,
        adminAuditLog,
      }),
    [workspaceApiKeys, workspaceUsageQuota, workspaceRunHistory, adminAuditLog],
  );
  const modelGatewayRouteEvidence = useMemo(
    () => buildModelGatewayRouteEvidenceViewModel({ overview: modelGatewayOverview, readShell: shell }),
    [modelGatewayOverview],
  );
  const modelGatewayUsageAuditEvidence = useMemo(
    () =>
      buildModelGatewayUsageAuditEvidenceViewModel({
        overview: modelGatewayOverview,
        routeEvidence: modelGatewayRouteEvidence,
        workspaceApiKeys,
        workspaceUsageQuota,
        workspaceRunHistory,
        adminAuditLog,
      }),
    [
      modelGatewayOverview,
      modelGatewayRouteEvidence,
      workspaceApiKeys,
      workspaceUsageQuota,
      workspaceRunHistory,
      adminAuditLog,
    ],
  );
  const modelGatewayEvidenceReview = useMemo(
    () =>
      buildModelGatewayEvidenceReviewViewModel({
        overview: modelGatewayOverview,
        routeEvidence: modelGatewayRouteEvidence,
        usageAuditEvidence: modelGatewayUsageAuditEvidence,
      }),
    [modelGatewayOverview, modelGatewayRouteEvidence, modelGatewayUsageAuditEvidence],
  );
  const adminOperationsReview = useMemo(
    () =>
      buildAdminOperationsReviewViewModel({
        readShell: shell,
        tenantOverview,
        adminAuditLog,
        modelGatewayEvidenceReview,
      }),
    [tenantOverview, adminAuditLog, modelGatewayEvidenceReview],
  );
  const adminProviderDeploymentReview = useMemo(
    () =>
      buildAdminProviderDeploymentReviewViewModel({
        tenantOverview,
        adminAuditLog,
        modelGatewayRouteEvidence,
        modelGatewayEvidenceReview,
        adminOperationsReview,
      }),
    [tenantOverview, adminAuditLog, modelGatewayRouteEvidence, modelGatewayEvidenceReview, adminOperationsReview],
  );
  const workflowWorkspaceContext = useMemo(
    () =>
      buildWorkflowWorkspaceContextViewModel({
        workspaceApplications,
        workspaceApiKeys,
        workspaceUsageQuota,
        workspaceWorkflowDefinitions,
        workspaceRunHistory,
        localWorkflowDrafts: workspaceCreatedDrafts,
        activeWorkflowDraftOverride: editableWorkflowDraft,
        selection: {
          applicationRef: selectedApplicationRef,
          workflowDefinitionId: selectedWorkflowDefinitionId,
          runId: selectedRunId,
          draftId: selectedWorkflowDraftId,
          scenarioId: selectedWorkflowScenarioId,
        },
      }),
    [
      workspaceApplications,
      workspaceApiKeys,
      workspaceUsageQuota,
      workspaceWorkflowDefinitions,
      workspaceRunHistory,
      workspaceCreatedDrafts,
      editableWorkflowDraft,
      selectedApplicationRef,
      selectedWorkflowDefinitionId,
      selectedRunId,
      selectedWorkflowDraftId,
      selectedWorkflowScenarioId,
    ],
  );
  const {
    selectedApplication,
    selectedWorkflowDefinition,
    selectedRun,
    selectedWorkflowDraft,
    activeWorkflowDraft,
    workflowApplicationDetail,
    workflowDefinitionDetail,
    workflowRunDetail,
    workflowBlockedActionPreview,
    workflowConfirmationPlaceholder,
    workflowDraftDesigner,
    workflowDraftValidationInspector: activeWorkflowDraftValidationInspector,
    workflowExecutionPlanPreview: activeWorkflowExecutionPlanPreview,
    workflowRuntimeReadinessInspector: activeWorkflowRuntimeReadinessInspector,
    workflowSurfaceOverview,
    workflowScenarioInspector,
    workflowWorkspaceReview,
    workflowUserWorkspaceHome,
    workflowReviewHandoff,
  } = workflowWorkspaceContext;
  const createdWorkspaceDraftCountsByDefinition = useMemo(
    () =>
      workspaceCreatedDrafts.reduce<Record<string, number>>((counts, draft) => {
        counts[draft.workflowDefinitionId] = (counts[draft.workflowDefinitionId] ?? 0) + 1;
        return counts;
      }, {}),
    [workspaceCreatedDrafts],
  );

  useEffect(() => {
    setEditableWorkflowDraft(cloneWorkflowDraftForEditing(selectedWorkflowDraft));
    if (selectedWorkflowDraft.localOnlyInteraction === "local_edit") {
      setWorkflowDraftEditDirty(true);
      setSavedDraftConsumerState(workspaceDraftCreatedConsumerState(savedDraftConsumerConfig, selectedWorkflowDraft));
      return;
    }
    setSavedDraftConsumerState(initialWorkflowSavedDraftConsumerState(savedDraftConsumerConfig));
    setWorkflowDraftEditDirty(false);
  }, [selectedWorkflowDraft.draftId]);

  const markWorkflowDraftLocallyEdited = () => {
    setWorkflowDraftEditDirty(true);
    setSavedDraftConsumerState((state) => ({
      ...state,
      status: "unsaved_local",
      sourceLabel: "unsaved local",
      summary:
        state.mode === "dev_saved_draft_http"
          ? "Local draft has unsaved edits; validate or save through the dev-only saved draft route."
          : "Local draft has unsaved edits and remains in sample-only mode.",
      failureCode: null,
      conflictDraftVersion: null,
    }));
  };

  const handleWorkflowDraftLabelChange = (label: string) => {
    setEditableWorkflowDraft((draft) => ({
      ...(draft ?? cloneWorkflowDraftForEditing(selectedWorkflowDraft)),
      label,
      localOnlyInteraction: "local_edit",
    }));
    markWorkflowDraftLocallyEdited();
  };

  const handleWorkflowDraftSummaryChange = (summary: string) => {
    setEditableWorkflowDraft((draft) => ({
      ...(draft ?? cloneWorkflowDraftForEditing(selectedWorkflowDraft)),
      summary,
      localOnlyInteraction: "local_edit",
    }));
    markWorkflowDraftLocallyEdited();
  };

  const handleWorkflowDraftNodeLabelChange = (nodeId: string, label: string) => {
    setEditableWorkflowDraft((draft) => {
      const currentDraft = draft ?? cloneWorkflowDraftForEditing(selectedWorkflowDraft);
      return {
        ...currentDraft,
        localOnlyInteraction: "local_edit",
        nodes: currentDraft.nodes.map((node) => (node.nodeId === nodeId ? { ...node, label } : node)),
      };
    });
    markWorkflowDraftLocallyEdited();
  };

  const handleWorkflowDraftEdgeConditionChange = (edgeId: string, conditionSummary: string) => {
    setEditableWorkflowDraft((draft) => {
      const currentDraft = draft ?? cloneWorkflowDraftForEditing(selectedWorkflowDraft);
      return {
        ...currentDraft,
        localOnlyInteraction: "local_edit",
        edges: currentDraft.edges.map((edge) =>
          edge.edgeId === edgeId ? { ...edge, conditionSummary } : edge,
        ),
      };
    });
    markWorkflowDraftLocallyEdited();
  };

  const handleWorkflowDraftAddNode = (nodeType: WorkflowDraftDesignerNode["nodeType"]) => {
    setEditableWorkflowDraft((draft) => {
      const currentDraft = draft ?? cloneWorkflowDraftForEditing(selectedWorkflowDraft);
      const nextNode = buildLocalWorkflowDraftNode(currentDraft, nodeType);
      return workflowDraftWithStructureEdits(currentDraft, insertWorkflowDraftNode(currentDraft.nodes, nextNode));
    });
    markWorkflowDraftLocallyEdited();
  };

  const handleWorkflowDraftMoveNode = (nodeId: string, direction: WorkflowDraftNodeMoveDirection) => {
    if (!canMoveWorkflowDraftNode(activeWorkflowDraft, nodeId, direction)) {
      return;
    }
    setEditableWorkflowDraft((draft) => {
      const currentDraft = draft ?? cloneWorkflowDraftForEditing(selectedWorkflowDraft);
      return workflowDraftWithStructureEdits(
        currentDraft,
        moveWorkflowDraftNode(currentDraft.nodes, nodeId, direction),
      );
    });
    markWorkflowDraftLocallyEdited();
  };

  const handleWorkflowDraftRemoveNode = (nodeId: string) => {
    if (!canRemoveWorkflowDraftNode(activeWorkflowDraft, nodeId)) {
      return;
    }
    setEditableWorkflowDraft((draft) => {
      const currentDraft = draft ?? cloneWorkflowDraftForEditing(selectedWorkflowDraft);
      return workflowDraftWithStructureEdits(
        currentDraft,
        currentDraft.nodes.filter((node) => node.nodeId !== nodeId),
      );
    });
    markWorkflowDraftLocallyEdited();
  };

  const handleWorkflowDraftEditReset = () => {
    setEditableWorkflowDraft(cloneWorkflowDraftForEditing(selectedWorkflowDraft));
    if (selectedWorkflowDraft.localOnlyInteraction === "local_edit") {
      setWorkflowDraftEditDirty(true);
      setSavedDraftConsumerState(workspaceDraftCreatedConsumerState(savedDraftConsumerConfig, selectedWorkflowDraft));
      return;
    }
    setWorkflowDraftEditDirty(false);
    setSavedDraftConsumerState(initialWorkflowSavedDraftConsumerState(savedDraftConsumerConfig));
  };

  const applyWorkflowSelectionPatch = ({
    applicationRef,
    workflowDefinitionId,
    runId,
    draftId,
    scenarioId,
  }: {
    applicationRef: string | null;
    workflowDefinitionId: string | null;
    runId: string | null;
    draftId: string | null;
    scenarioId: string | null;
  }) => {
    setSelectedApplicationRef(applicationRef);
    setSelectedWorkflowDefinitionId(workflowDefinitionId);
    setSelectedRunId(runId);
    setSelectedWorkflowDraftId(draftId);
    setSelectedWorkflowScenarioId(scenarioId);
  };
  const handleSelectApplication = (applicationRef: string) => {
    applyWorkflowSelectionPatch(
      selectionForApplication(applicationRef, {
        workspaceApplications,
        workspaceWorkflowDefinitions,
        workspaceRunHistory,
      }),
    );
  };
  const handleSelectWorkflowDefinition = (workflowDefinitionId: string) => {
    applyWorkflowSelectionPatch(
      selectionForWorkflowDefinition(workflowDefinitionId, {
        workspaceWorkflowDefinitions,
        workspaceRunHistory,
      }),
    );
  };
  const handleSelectRun = (runId: string) => {
    applyWorkflowSelectionPatch(selectionForRun(runId, { workspaceRunHistory }));
  };
  const handleSelectWorkflowDraft = (draftId: string) => {
    applyWorkflowSelectionPatch(selectionForDraft(draftId, workflowDraftDesigner, { workspaceRunHistory }));
  };
  const handleCreateWorkspaceDraftFromDefinition = (workflowDefinitionId: string) => {
    const createdDraft = buildWorkspaceCreatedDraft(
      workflowDefinitionId,
      workflowDraftDesigner,
      workspaceCreatedDrafts,
    );
    if (!createdDraft) {
      return;
    }
    const nextRun = workspaceRunHistory.runs.find(
      (run) =>
        run.applicationRef === createdDraft.applicationRef &&
        run.workflowDefinitionId === createdDraft.workflowDefinitionId,
    );
    setWorkspaceCreatedDrafts((drafts) => [...drafts, createdDraft]);
    applyWorkflowSelectionPatch({
      applicationRef: createdDraft.applicationRef,
      workflowDefinitionId: createdDraft.workflowDefinitionId,
      runId: nextRun?.runId ?? null,
      draftId: createdDraft.draftId,
      scenarioId: null,
    });
    setEditableWorkflowDraft(cloneWorkflowDraftForEditing(createdDraft));
    setWorkflowDraftEditDirty(true);
    setSavedDraftConsumerState(workspaceDraftCreatedConsumerState(savedDraftConsumerConfig, createdDraft));
  };
  const refreshSavedWorkflowDraftList = (applicationRef: string) => {
    if (savedDraftConsumerConfig.mode !== "dev_saved_draft_http") {
      setSavedDraftListState(initialWorkflowSavedDraftListState(savedDraftConsumerConfig, applicationRef));
      return;
    }
    setSavedDraftListState((state) => ({
      ...state,
      status: "loading",
      mode: "dev_saved_draft_http",
      sourceLabel: "loading",
      summary: "Loading saved dev draft summaries for the selected application.",
      applicationRef,
      failureCode: null,
      summaries: [],
    }));
    listWorkflowDraftDevRecords(applicationRef, savedDraftConsumerConfig)
      .then(setSavedDraftListState)
      .catch((error: unknown) => {
        setSavedDraftListState((state) => ({
          ...state,
          status: "list_failed",
          sourceLabel: "list_failed",
          summary: error instanceof Error ? error.message : "Saved draft list failed.",
          applicationRef,
          failureCode: "dev_saved_draft_list_failed",
          summaries: [],
        }));
      });
  };
  useEffect(() => {
    if (savedDraftConsumerConfig.mode !== "dev_saved_draft_http") {
      setSavedDraftListState(initialWorkflowSavedDraftListState(savedDraftConsumerConfig, selectedApplication.applicationRef));
      return;
    }
    refreshSavedWorkflowDraftList(selectedApplication.applicationRef);
  }, [selectedApplication.applicationRef]);
  const handleRefreshSavedWorkflowDraftList = () => {
    refreshSavedWorkflowDraftList(selectedApplication.applicationRef);
  };
  const handleRestoreSavedWorkflowDraft = (summary: WorkflowSavedDraftSummary) => {
    if (savedDraftConsumerConfig.mode !== "dev_saved_draft_http") {
      return;
    }
    setSavedDraftConsumerState((state) => ({
      ...state,
      status: "reading",
      summary: `Restoring saved draft ${summary.draftId} through the dev-only read route.`,
      failureCode: null,
      conflictDraftVersion: null,
    }));
    restoreWorkflowDraftDevRecord(summary, savedDraftConsumerConfig)
      .then((result) => {
        setSavedDraftConsumerState(result.state);
        if (!result.draft) {
          setSavedDraftListState((state) => ({
            ...state,
            status: "restore_failed",
            sourceLabel: "restore_failed",
            summary: result.state.summary,
            failureCode: result.state.failureCode ?? "dev_saved_draft_restore_failed",
          }));
          return;
        }
        const restoredDraft = result.draft;
        const nextRun = workspaceRunHistory.runs.find(
          (run) =>
            run.applicationRef === restoredDraft.applicationRef &&
            run.workflowDefinitionId === restoredDraft.workflowDefinitionId,
        );
        setWorkspaceCreatedDrafts((drafts) => [
          ...drafts.filter((draft) => draft.draftId !== restoredDraft.draftId),
          restoredDraft,
        ]);
        applyWorkflowSelectionPatch({
          applicationRef: restoredDraft.applicationRef,
          workflowDefinitionId: restoredDraft.workflowDefinitionId,
          runId: nextRun?.runId ?? null,
          draftId: restoredDraft.draftId,
          scenarioId: null,
        });
        setEditableWorkflowDraft(cloneWorkflowDraftForEditing(restoredDraft));
        setWorkflowDraftEditDirty(false);
      })
      .catch((error: unknown) => {
        const message = error instanceof Error ? error.message : "Saved draft restore failed.";
        setSavedDraftConsumerState((state) => ({
          ...state,
          status: "read_failed",
          sourceLabel: "restore_failed",
          summary: message,
          failureCode: "dev_saved_draft_restore_failed",
          conflictDraftVersion: null,
        }));
        setSavedDraftListState((state) => ({
          ...state,
          status: "restore_failed",
          sourceLabel: "restore_failed",
          summary: message,
          failureCode: "dev_saved_draft_restore_failed",
        }));
      });
  };
  const handleValidateWorkflowDraft = () => {
    if (savedDraftConsumerConfig.mode !== "dev_saved_draft_http") {
      return;
    }
    setSavedDraftConsumerState((state) => ({
      ...state,
      status: "validating",
      summary: "Validating local draft through the dev-only saved draft route.",
      failureCode: null,
      conflictDraftVersion: null,
    }));
    validateWorkflowDraftDevRecord(activeWorkflowDraft, savedDraftConsumerConfig)
      .then(setSavedDraftConsumerState)
      .catch((error: unknown) => {
        setSavedDraftConsumerState((state) => ({
          ...state,
          status: "validation_failed",
          sourceLabel: "validation_failed",
          summary: error instanceof Error ? error.message : "Saved draft validation failed.",
          failureCode: "dev_saved_draft_consumer_failed",
          conflictDraftVersion: null,
        }));
      });
  };
  const handleSaveWorkflowDraft = () => {
    if (savedDraftConsumerConfig.mode !== "dev_saved_draft_http") {
      return;
    }
    const expectedDraftVersion = nextWorkflowSavedDraftExpectedVersion(savedDraftConsumerState);
    setSavedDraftConsumerState((state) => ({
      ...state,
      status: "saving",
      summary: "Saving local draft through the dev-only saved draft route.",
      failureCode: null,
      conflictDraftVersion: null,
    }));
    saveWorkflowDraftDevRecord(activeWorkflowDraft, savedDraftConsumerConfig, expectedDraftVersion)
      .then((nextState) => {
        setSavedDraftConsumerState(nextState);
        if (nextState.status === "saved_dev_record") {
          setWorkspaceCreatedDrafts((drafts) =>
            drafts.map((draft) =>
              draft.draftId === activeWorkflowDraft.draftId
                ? { ...activeWorkflowDraft, localOnlyInteraction: "inspect_only" }
                : draft,
            ),
          );
          setEditableWorkflowDraft((draft) =>
            draft === null ? null : { ...draft, localOnlyInteraction: "inspect_only" },
          );
          setWorkflowDraftEditDirty(false);
          refreshSavedWorkflowDraftList(activeWorkflowDraft.applicationRef);
        }
      })
      .catch((error: unknown) => {
        setSavedDraftConsumerState((state) => ({
          ...state,
          status: "save_failed",
          sourceLabel: "save_failed",
          summary: error instanceof Error ? error.message : "Saved draft save failed.",
          failureCode: "dev_saved_draft_consumer_failed",
          conflictDraftVersion: null,
        }));
      });
  };
  const handleReadWorkflowDraft = () => {
    if (savedDraftConsumerConfig.mode !== "dev_saved_draft_http") {
      return;
    }
    setSavedDraftConsumerState((state) => ({
      ...state,
      status: "reading",
      summary: "Reading local draft through the dev-only saved draft route.",
      failureCode: null,
      conflictDraftVersion: null,
    }));
    readWorkflowDraftDevRecord(activeWorkflowDraft, savedDraftConsumerConfig)
      .then(setSavedDraftConsumerState)
      .catch((error: unknown) => {
        setSavedDraftConsumerState((state) => ({
          ...state,
          status: "read_failed",
          sourceLabel: "read_failed",
          summary: error instanceof Error ? error.message : "Saved draft read failed.",
          failureCode: "dev_saved_draft_consumer_failed",
          conflictDraftVersion: null,
        }));
      });
  };

  return (
    <main className="product-shell">
      <aside className="product-nav" aria-label="Product areas">
        <div>
          <p className="eyebrow">RadishMind</p>
          <h1>Control Plane</h1>
          <p className="nav-summary">Read-only product surface for tenant, workspace, usage, workflow, and audit views.</p>
        </div>
        <nav className="nav-links" aria-label="Read shell sections">
          <div className="nav-link-group" aria-label="User workspace sections">
            <p className="nav-link-group-label">Workspace</p>
            <a href="#workflow-user-workspace-home">Workspace Home</a>
            <a href="#workspace-applications">Applications</a>
            <a href="#workspace-workflow-definitions">Workflows</a>
            <a href="#workspace-run-history">Run History</a>
            <a href="#workspace-api-keys">API Keys</a>
            <a href="#workspace-usage-quota">Usage Quota</a>
          </div>
          <div className="nav-link-group" aria-label="Model gateway sections">
            <p className="nav-link-group-label">Model Gateway</p>
            <a href="#model-gateway-overview">Gateway Overview</a>
            <a href="#model-gateway-route-evidence">Route Evidence</a>
            <a href="#model-gateway-usage-audit-evidence">Usage Evidence</a>
            <a href="#model-gateway-evidence-review">Evidence Review</a>
          </div>
          <div className="nav-link-group" aria-label="Workflow review sections">
            <p className="nav-link-group-label">Workflow Review</p>
            <a href="#workflow-application-detail">Application Detail</a>
            <a href="#workflow-draft-designer">Draft Designer</a>
            <a href="#workflow-draft-validation-inspector">Draft Validation</a>
            <a href="#workflow-execution-plan-preview">Execution Plan</a>
            <a href="#workflow-runtime-readiness-inspector">Runtime Readiness</a>
            <a href="#workflow-scenario-inspector">Scenario Inspector</a>
            <a href="#workflow-workspace-review">Review Workspace</a>
            <a href="#workflow-review-handoff">Review Handoff</a>
            <a href="#workflow-surface-overview">Workflow Overview</a>
            <a href="#workflow-blocked-action-preview">Blocked Action</a>
            <a href="#workflow-confirmation-placeholder">Confirmation</a>
          </div>
          <div className="nav-link-group" aria-label="Admin control plane sections">
            <p className="nav-link-group-label">Admin</p>
            <a href="#admin-operations-review">Operations Review</a>
            <a href="#admin-provider-deployment-review">Provider Deployment</a>
            <a href="#admin-tenant-overview">Tenant Overview</a>
            <a href="#admin-audit-log">Audit Log</a>
          </div>
          <div className="nav-link-group" aria-label="Contract and guard sections">
            <p className="nav-link-group-label">Contract</p>
            <a href="#routes">Route Catalog</a>
            <a href="#states">Shared States</a>
            <a href="#guard">Output Guard</a>
          </div>
        </nav>
        <div className="nav-locks" aria-label="Stop lines">
          {shell.lockedCapabilities.map((capability) => (
            <span key={capability}>{capability}</span>
          ))}
        </div>
      </aside>

      <section className="product-workspace" aria-label="Control plane read shell">
        <header className="workspace-header">
          <div>
            <p className="eyebrow">Shared Read Shell</p>
            <h2>Read catalog and status model</h2>
          </div>
          <div className="header-facts" aria-label="Read shell facts">
            <Fact label="Routes" value={String(shell.catalog.routes.length)} />
            <Fact label="Database" value={shell.catalog.databaseBacked ? "attached" : "detached"} />
            <Fact label="Writes" value={shell.catalog.allRoutesReadOnly ? "locked" : "enabled"} />
            <Fact label="Source" value={devLiveState.mode === "dev_live_http" ? devLiveState.status : "offline"} />
            <Fact label="Tenant page" value={tenantOverview.canRenderTenant ? "ready" : "blocked"} />
            <Fact label="Audit page" value={adminAuditLog.canRenderAuditLog ? "ready" : "blocked"} />
            <Fact label="App page" value={workspaceApplications.canRenderApplications ? "ready" : "blocked"} />
            <Fact
              label="App detail"
              value={workflowApplicationDetail.canRenderApplicationDetail ? "ready" : "blocked"}
            />
            <Fact label="Key page" value={workspaceApiKeys.canRenderApiKeys ? "ready" : "blocked"} />
            <Fact label="Quota page" value={workspaceUsageQuota.canRenderQuota ? "ready" : "blocked"} />
            <Fact
              label="Workflow page"
              value={workspaceWorkflowDefinitions.canRenderWorkflowDefinitions ? "ready" : "blocked"}
            />
            <Fact label="Run page" value={workspaceRunHistory.canRenderRuns ? "ready" : "blocked"} />
            <Fact
              label="Action guard"
              value={workflowBlockedActionPreview.canRenderBlockedActionPreview ? "ready" : "blocked"}
            />
            <Fact
              label="Confirm"
              value={workflowConfirmationPlaceholder.canRenderConfirmationPlaceholder ? "ready" : "blocked"}
            />
            <Fact
              label="Draft"
              value={workflowDraftDesigner.canRenderDraftDesigner ? "ready" : "blocked"}
            />
            <Fact
              label="Validate"
              value={activeWorkflowDraftValidationInspector.validationStatus}
            />
            <Fact
              label="Plan"
              value={activeWorkflowExecutionPlanPreview.canRenderExecutionPlanPreview ? "preview" : "blocked"}
            />
            <Fact
              label="Runtime"
              value={activeWorkflowRuntimeReadinessInspector.canRenderRuntimeReadinessInspector ? "blocked" : "missing"}
            />
            <Fact
              label="Overview"
              value={workflowSurfaceOverview.canRenderSurfaceOverview ? "offline" : "blocked"}
            />
            <Fact
              label="Scenario"
              value={workflowScenarioInspector.canRenderScenarioInspector ? "offline" : "blocked"}
            />
            <Fact
              label="Review"
              value={workflowWorkspaceReview.canRenderWorkspaceReview ? "offline" : "blocked"}
            />
            <Fact
              label="Home"
              value={workflowUserWorkspaceHome.canRenderUserWorkspaceHome ? "offline" : "blocked"}
            />
            <Fact
              label="Handoff"
              value={workflowReviewHandoff.canRenderReviewHandoff ? "offline" : "blocked"}
            />
            <Fact
              label="Gateway"
              value={modelGatewayOverview.canRenderModelGatewayOverview ? "offline" : "blocked"}
            />
            <Fact
              label="Gateway route"
              value={modelGatewayRouteEvidence.canRenderRouteEvidenceDetail ? "offline" : "blocked"}
            />
            <Fact
              label="Gateway usage"
              value={modelGatewayUsageAuditEvidence.canRenderUsageAuditEvidence ? "offline" : "blocked"}
            />
            <Fact
              label="Gateway review"
              value={modelGatewayEvidenceReview.canRenderEvidenceReview ? "offline" : "blocked"}
            />
            <Fact
              label="Admin review"
              value={adminOperationsReview.canRenderAdminOperationsReview ? "offline" : "blocked"}
            />
            <Fact
              label="Admin provider"
              value={adminProviderDeploymentReview.canRenderProviderDeploymentReview ? "offline" : "blocked"}
            />
          </div>
        </header>

        <LiveReadSourceStatus state={devLiveState} baseUrl={devLiveConfig.baseUrl} />
        <WorkflowUserWorkspaceHomePanel
          home={workflowUserWorkspaceHome}
          createdDraftCountsByWorkflowDefinition={createdWorkspaceDraftCountsByDefinition}
          savedDraftListState={savedDraftListState}
          onCreateDraftForWorkflowDefinition={handleCreateWorkspaceDraftFromDefinition}
          onRefreshSavedDrafts={handleRefreshSavedWorkflowDraftList}
          onRestoreSavedDraft={handleRestoreSavedWorkflowDraft}
        />
        <ModelGatewayOverviewPanel overview={modelGatewayOverview} />
        <ModelGatewayRouteEvidencePanel detail={modelGatewayRouteEvidence} />
        <ModelGatewayUsageAuditEvidencePanel evidence={modelGatewayUsageAuditEvidence} />
        <ModelGatewayEvidenceReviewPanel review={modelGatewayEvidenceReview} />
        <AdminOperationsReviewPanel review={adminOperationsReview} />
        <AdminProviderDeploymentReviewPanel review={adminProviderDeploymentReview} />
        <WorkflowWorkspaceReviewPanel review={workflowWorkspaceReview} />
        <WorkflowReviewHandoffPanel handoff={workflowReviewHandoff} />
        <WorkflowSurfaceOverviewPanel overview={workflowSurfaceOverview} />
        <WorkflowScenarioInspectorPanel
          inspector={workflowScenarioInspector}
          selectedScenarioId={workflowScenarioInspector.selectedScenarioId}
          onSelectScenario={setSelectedWorkflowScenarioId}
        />

        <section
          className="surface-band tenant-overview"
          id="admin-tenant-overview"
          aria-labelledby="admin-tenant-overview-title"
        >
          <div className="section-heading">
            <div>
              <p className="eyebrow">Admin Control Plane</p>
              <h3 id="admin-tenant-overview-title">Tenant overview</h3>
            </div>
            <StatusBadge tone={tenantOverview.canRenderTenant ? "good" : "bad"}>
              {tenantOverview.canRenderTenant ? "read-only ready" : "blocked"}
            </StatusBadge>
          </div>

          <div className="tenant-layout">
            <article className="tenant-summary">
              <div className="card-title-row">
                <div>
                  <p className="eyebrow">Tenant Summary Route</p>
                  <h4>{tenantOverview.tenant?.tenant_display_name ?? "No tenant summary"}</h4>
                </div>
                <StatusBadge tone="neutral">{tenantOverview.requiredScope}</StatusBadge>
              </div>
              <p className="route-path">{tenantOverview.routePath}</p>
              <dl className="tenant-meta">
                <div>
                  <dt>Route</dt>
                  <dd>{tenantOverview.routeId}</dd>
                </div>
                <div>
                  <dt>Model</dt>
                  <dd>{tenantOverview.readModel}</dd>
                </div>
                <div>
                  <dt>Request</dt>
                  <dd>{tenantOverview.requestId}</dd>
                </div>
                <div>
                  <dt>Audit</dt>
                  <dd>{tenantOverview.auditRef}</dd>
                </div>
              </dl>
            </article>

            <div className="tenant-facts" aria-label="Tenant overview facts">
              {tenantOverview.facts.map((fact) => (
                <TenantFact key={fact.label} fact={fact} />
              ))}
            </div>
          </div>

          <div className="tenant-states" aria-label="Tenant overview states">
            {tenantOverview.statePreviews.map((state) => (
              <TenantStatePreview key={state.id} state={state} />
            ))}
          </div>
        </section>

        <section className="surface-band admin-audit-log" id="admin-audit-log" aria-labelledby="admin-audit-log-title">
          <div className="section-heading">
            <div>
              <p className="eyebrow">Admin Control Plane</p>
              <h3 id="admin-audit-log-title">Audit log</h3>
            </div>
            <StatusBadge tone={adminAuditLog.canRenderAuditLog ? "good" : "bad"}>
              {adminAuditLog.canRenderAuditLog ? "read-only ready" : "blocked"}
            </StatusBadge>
          </div>

          <div className="audit-log-summary">
            <article className="audit-log-route">
              <div className="card-title-row">
                <div>
                  <p className="eyebrow">Audit Summary List Route</p>
                  <h4>{adminAuditLog.routeId}</h4>
                </div>
                <StatusBadge tone="neutral">{adminAuditLog.requiredScope}</StatusBadge>
              </div>
              <p className="route-path">{adminAuditLog.routePath}</p>
              <dl className="tenant-meta">
                <div>
                  <dt>Model</dt>
                  <dd>{adminAuditLog.readModel}</dd>
                </div>
                <div>
                  <dt>Request</dt>
                  <dd>{adminAuditLog.requestId}</dd>
                </div>
                <div>
                  <dt>Next cursor</dt>
                  <dd>{adminAuditLog.nextCursor ?? "none"}</dd>
                </div>
                <div>
                  <dt>Audit</dt>
                  <dd>{adminAuditLog.auditRef}</dd>
                </div>
              </dl>
            </article>

            <div className="audit-log-metrics" aria-label="Admin audit log metrics">
              {adminAuditLog.metrics.map((metric) => (
                <AuditLogMetric key={metric.label} metric={metric} />
              ))}
            </div>
          </div>

          <div className="audit-event-list" aria-label="Admin audit events">
            {adminAuditLog.auditEvents.map((event) => (
              <AuditEventRow key={event.auditRef} event={event} />
            ))}
          </div>

          <div className="audit-log-states" aria-label="Admin audit log states">
            {adminAuditLog.statePreviews.map((state) => (
              <AuditLogStatePreview key={state.id} state={state} />
            ))}
          </div>
        </section>

        <section
          className="surface-band workspace-applications"
          id="workspace-applications"
          aria-labelledby="workspace-applications-title"
        >
          <div className="section-heading">
            <div>
              <p className="eyebrow">User Workspace</p>
              <h3 id="workspace-applications-title">Applications</h3>
            </div>
            <StatusBadge tone={workspaceApplications.canRenderApplications ? "good" : "bad"}>
              {workspaceApplications.canRenderApplications ? "read-only ready" : "blocked"}
            </StatusBadge>
          </div>

          <div className="applications-summary">
            <article className="applications-route">
              <div className="card-title-row">
                <div>
                  <p className="eyebrow">Application Summary List Route</p>
                  <h4>{workspaceApplications.routeId}</h4>
                </div>
                <StatusBadge tone="neutral">{workspaceApplications.requiredScope}</StatusBadge>
              </div>
              <p className="route-path">{workspaceApplications.routePath}</p>
              <dl className="tenant-meta">
                <div>
                  <dt>Model</dt>
                  <dd>{workspaceApplications.readModel}</dd>
                </div>
                <div>
                  <dt>Request</dt>
                  <dd>{workspaceApplications.requestId}</dd>
                </div>
                <div>
                  <dt>Next cursor</dt>
                  <dd>{workspaceApplications.nextCursor ?? "none"}</dd>
                </div>
                <div>
                  <dt>Audit</dt>
                  <dd>{workspaceApplications.auditRef}</dd>
                </div>
              </dl>
            </article>

            <div className="applications-metrics" aria-label="Workspace application metrics">
              {workspaceApplications.metrics.map((metric) => (
                <ApplicationMetric key={metric.label} metric={metric} />
              ))}
            </div>
          </div>

          <div className="application-list" aria-label="Workspace applications">
            {workspaceApplications.applications.map((application) => (
              <ApplicationRow
                key={application.applicationRef}
                application={application}
                selected={application.applicationRef === selectedApplication.applicationRef}
                onSelectApplication={handleSelectApplication}
              />
            ))}
          </div>

          <WorkflowApplicationDetailPanel detail={workflowApplicationDetail} />

          <div className="application-states" aria-label="Workspace application states">
            {workspaceApplications.statePreviews.map((state) => (
              <ApplicationStatePreview key={state.id} state={state} />
            ))}
          </div>
        </section>

        <section className="surface-band workspace-api-keys" id="workspace-api-keys" aria-labelledby="workspace-api-keys-title">
          <div className="section-heading">
            <div>
              <p className="eyebrow">User Workspace</p>
              <h3 id="workspace-api-keys-title">API keys</h3>
            </div>
            <StatusBadge tone={workspaceApiKeys.canRenderApiKeys ? "good" : "bad"}>
              {workspaceApiKeys.canRenderApiKeys ? "read-only ready" : "blocked"}
            </StatusBadge>
          </div>

          <div className="api-keys-summary">
            <article className="api-keys-route">
              <div className="card-title-row">
                <div>
                  <p className="eyebrow">API Key Summary List Route</p>
                  <h4>{workspaceApiKeys.routeId}</h4>
                </div>
                <StatusBadge tone="neutral">{workspaceApiKeys.requiredScope}</StatusBadge>
              </div>
              <p className="route-path">{workspaceApiKeys.routePath}</p>
              <dl className="tenant-meta">
                <div>
                  <dt>Model</dt>
                  <dd>{workspaceApiKeys.readModel}</dd>
                </div>
                <div>
                  <dt>Request</dt>
                  <dd>{workspaceApiKeys.requestId}</dd>
                </div>
                <div>
                  <dt>Next cursor</dt>
                  <dd>{workspaceApiKeys.nextCursor ?? "none"}</dd>
                </div>
                <div>
                  <dt>Audit</dt>
                  <dd>{workspaceApiKeys.auditRef}</dd>
                </div>
              </dl>
            </article>

            <div className="api-keys-metrics" aria-label="Workspace API key metrics">
              {workspaceApiKeys.metrics.map((metric) => (
                <ApiKeyMetric key={metric.label} metric={metric} />
              ))}
            </div>
          </div>

          <div className="api-key-list" aria-label="Workspace API keys">
            {workspaceApiKeys.apiKeys.map((apiKey) => (
              <ApiKeyRow key={apiKey.apiKeyId} apiKey={apiKey} />
            ))}
          </div>

          <div className="api-key-states" aria-label="Workspace API key states">
            {workspaceApiKeys.statePreviews.map((state) => (
              <ApiKeyStatePreview key={state.id} state={state} />
            ))}
          </div>
        </section>

        <section
          className="surface-band workspace-usage-quota"
          id="workspace-usage-quota"
          aria-labelledby="workspace-usage-quota-title"
        >
          <div className="section-heading">
            <div>
              <p className="eyebrow">User Workspace</p>
              <h3 id="workspace-usage-quota-title">Usage quota</h3>
            </div>
            <StatusBadge tone={workspaceUsageQuota.canRenderQuota ? "good" : "bad"}>
              {workspaceUsageQuota.canRenderQuota ? "read-only ready" : "blocked"}
            </StatusBadge>
          </div>

          <div className="usage-quota-summary">
            <article className="usage-quota-route">
              <div className="card-title-row">
                <div>
                  <p className="eyebrow">Quota Summary Route</p>
                  <h4>{workspaceUsageQuota.routeId}</h4>
                </div>
                <StatusBadge tone="neutral">{workspaceUsageQuota.requiredScope}</StatusBadge>
              </div>
              <p className="route-path">{workspaceUsageQuota.routePath}</p>
              <dl className="tenant-meta">
                <div>
                  <dt>Model</dt>
                  <dd>{workspaceUsageQuota.readModel}</dd>
                </div>
                <div>
                  <dt>Period</dt>
                  <dd>{workspaceUsageQuota.quota?.period ?? "not available"}</dd>
                </div>
                <div>
                  <dt>Request</dt>
                  <dd>{workspaceUsageQuota.requestId}</dd>
                </div>
                <div>
                  <dt>Audit</dt>
                  <dd>{workspaceUsageQuota.auditRef}</dd>
                </div>
              </dl>
            </article>

            <div className="usage-quota-snapshot" aria-label="Workspace usage quota snapshot">
              {workspaceUsageQuota.usageSnapshot.map((snapshot) => (
                <UsageQuotaSnapshot key={snapshot.label} snapshot={snapshot} />
              ))}
            </div>
          </div>

          <div className="usage-quota-limits" aria-label="Workspace usage quota limits">
            {workspaceUsageQuota.limits.map((limit) => (
              <UsageQuotaLimit key={limit.label} limit={limit} />
            ))}
          </div>

          <div className="usage-quota-failure">
            <span>Over quota failure code</span>
            <strong>{workspaceUsageQuota.overQuotaFailureCode}</strong>
            <p>Displayed as read-side metadata only; enforcement, rate limit and cost record writes remain outside this page.</p>
          </div>

          <div className="usage-quota-states" aria-label="Workspace usage quota states">
            {workspaceUsageQuota.statePreviews.map((state) => (
              <UsageQuotaStatePreview key={state.id} state={state} />
            ))}
          </div>
        </section>

        <section
          className="surface-band workspace-workflow-definitions"
          id="workspace-workflow-definitions"
          aria-labelledby="workspace-workflow-definitions-title"
        >
          <div className="section-heading">
            <div>
              <p className="eyebrow">User Workspace</p>
              <h3 id="workspace-workflow-definitions-title">Workflow definitions</h3>
            </div>
            <StatusBadge tone={workspaceWorkflowDefinitions.canRenderWorkflowDefinitions ? "good" : "bad"}>
              {workspaceWorkflowDefinitions.canRenderWorkflowDefinitions ? "read-only ready" : "blocked"}
            </StatusBadge>
          </div>

          <div className="workflow-definitions-summary">
            <article className="workflow-definitions-route">
              <div className="card-title-row">
                <div>
                  <p className="eyebrow">Workflow Definition Summary List Route</p>
                  <h4>{workspaceWorkflowDefinitions.routeId}</h4>
                </div>
                <StatusBadge tone="neutral">{workspaceWorkflowDefinitions.requiredScope}</StatusBadge>
              </div>
              <p className="route-path">{workspaceWorkflowDefinitions.routePath}</p>
              <dl className="tenant-meta">
                <div>
                  <dt>Model</dt>
                  <dd>{workspaceWorkflowDefinitions.readModel}</dd>
                </div>
                <div>
                  <dt>Request</dt>
                  <dd>{workspaceWorkflowDefinitions.requestId}</dd>
                </div>
                <div>
                  <dt>Next cursor</dt>
                  <dd>{workspaceWorkflowDefinitions.nextCursor ?? "none"}</dd>
                </div>
                <div>
                  <dt>Audit</dt>
                  <dd>{workspaceWorkflowDefinitions.auditRef}</dd>
                </div>
              </dl>
            </article>

            <div className="workflow-definitions-metrics" aria-label="Workspace workflow definition metrics">
              {workspaceWorkflowDefinitions.metrics.map((metric) => (
                <WorkflowDefinitionMetric key={metric.label} metric={metric} />
              ))}
            </div>
          </div>

          <div className="workflow-definition-list" aria-label="Workspace workflow definitions">
            {workspaceWorkflowDefinitions.workflowDefinitions.map((workflowDefinition) => (
              <WorkflowDefinitionRow
                key={workflowDefinition.workflowDefinitionId}
                workflowDefinition={workflowDefinition}
                selected={workflowDefinition.workflowDefinitionId === selectedWorkflowDefinition.workflowDefinitionId}
                createdDraftCount={
                  createdWorkspaceDraftCountsByDefinition[workflowDefinition.workflowDefinitionId] ?? 0
                }
                onSelectWorkflowDefinition={handleSelectWorkflowDefinition}
                onCreateDraftForWorkflowDefinition={handleCreateWorkspaceDraftFromDefinition}
              />
            ))}
          </div>

          <WorkflowDefinitionDetailPanel detail={workflowDefinitionDetail} />
          <WorkflowDraftDesignerPanel
            designer={workflowDraftDesigner}
            selectedDraft={activeWorkflowDraft}
            selectedDraftId={selectedWorkflowDraft.draftId}
            savedDraftConsumerState={savedDraftConsumerState}
            draftEditDirty={workflowDraftEditDirty}
            onSelectDraft={handleSelectWorkflowDraft}
            onUpdateDraftLabel={handleWorkflowDraftLabelChange}
            onUpdateDraftSummary={handleWorkflowDraftSummaryChange}
            onUpdateNodeLabel={handleWorkflowDraftNodeLabelChange}
            onUpdateEdgeCondition={handleWorkflowDraftEdgeConditionChange}
            onAddNode={handleWorkflowDraftAddNode}
            onMoveNode={handleWorkflowDraftMoveNode}
            onRemoveNode={handleWorkflowDraftRemoveNode}
            onResetDraftEdits={handleWorkflowDraftEditReset}
            onValidateDraft={handleValidateWorkflowDraft}
            onSaveDraft={handleSaveWorkflowDraft}
            onReadDraft={handleReadWorkflowDraft}
          />
          <WorkflowDraftValidationInspectorPanel inspector={activeWorkflowDraftValidationInspector} />
          <WorkflowExecutionPlanPreviewPanel preview={activeWorkflowExecutionPlanPreview} />
          <WorkflowRuntimeReadinessInspectorPanel readiness={activeWorkflowRuntimeReadinessInspector} />

          <div className="workflow-definition-states" aria-label="Workspace workflow definition states">
            {workspaceWorkflowDefinitions.statePreviews.map((state) => (
              <WorkflowDefinitionStatePreview key={state.id} state={state} />
            ))}
          </div>
        </section>

        <section
          className="surface-band workspace-run-history"
          id="workspace-run-history"
          aria-labelledby="workspace-run-history-title"
        >
          <div className="section-heading">
            <div>
              <p className="eyebrow">User Workspace</p>
              <h3 id="workspace-run-history-title">Run history</h3>
            </div>
            <StatusBadge tone={workspaceRunHistory.canRenderRuns ? "good" : "bad"}>
              {workspaceRunHistory.canRenderRuns ? "read-only ready" : "blocked"}
            </StatusBadge>
          </div>

          <div className="run-history-summary">
            <article className="run-history-route">
              <div className="card-title-row">
                <div>
                  <p className="eyebrow">Run Record Summary List Route</p>
                  <h4>{workspaceRunHistory.routeId}</h4>
                </div>
                <StatusBadge tone="neutral">{workspaceRunHistory.requiredScope}</StatusBadge>
              </div>
              <p className="route-path">{workspaceRunHistory.routePath}</p>
              <dl className="tenant-meta">
                <div>
                  <dt>Model</dt>
                  <dd>{workspaceRunHistory.readModel}</dd>
                </div>
                <div>
                  <dt>Request</dt>
                  <dd>{workspaceRunHistory.requestId}</dd>
                </div>
                <div>
                  <dt>Next cursor</dt>
                  <dd>{workspaceRunHistory.nextCursor ?? "none"}</dd>
                </div>
                <div>
                  <dt>Audit</dt>
                  <dd>{workspaceRunHistory.auditRef}</dd>
                </div>
              </dl>
            </article>

            <div className="run-history-metrics" aria-label="Workspace run history metrics">
              {workspaceRunHistory.metrics.map((metric) => (
                <RunHistoryMetric key={metric.label} metric={metric} />
              ))}
            </div>
          </div>

          <div className="run-record-list" aria-label="Workspace run records">
            {workspaceRunHistory.runs.map((run) => (
              <RunRecordRow
                key={run.runId}
                run={run}
                selected={run.runId === selectedRun.runId}
                onSelectRun={handleSelectRun}
              />
            ))}
          </div>

          <WorkflowRunDetailPanel detail={workflowRunDetail} />
          <WorkflowBlockedActionPreviewPanel preview={workflowBlockedActionPreview} />
          <WorkflowConfirmationPlaceholderPanel placeholder={workflowConfirmationPlaceholder} />

          <div className="run-history-states" aria-label="Workspace run history states">
            {workspaceRunHistory.statePreviews.map((state) => (
              <RunHistoryStatePreview key={state.id} state={state} />
            ))}
          </div>
        </section>

        <section className="surface-band" id="routes" aria-labelledby="routes-title">
          <div className="section-heading">
            <div>
              <p className="eyebrow">Contract Binding</p>
              <h3 id="routes-title">Route catalog</h3>
            </div>
            <StatusBadge tone="neutral">offline contract</StatusBadge>
          </div>
          <div className="route-grid">
            {shell.routeCards.map((route) => (
              <RouteCard key={route.routeId} route={route} />
            ))}
          </div>
        </section>

        <section className="surface-band" id="states" aria-labelledby="states-title">
          <div className="section-heading">
            <div>
              <p className="eyebrow">State Model</p>
              <h3 id="states-title">Shared states</h3>
            </div>
            <StatusBadge tone="good">{shell.readyPreview.statusLabel}</StatusBadge>
          </div>
          <div className="state-grid">
            {shell.statePreviews.map((state) => (
              <StatePreview key={state.id} state={state} />
            ))}
          </div>
        </section>

        <section className="surface-band guard-band" id="guard" aria-labelledby="guard-title">
          <div className="section-heading">
            <div>
              <p className="eyebrow">Sensitive Output</p>
              <h3 id="guard-title">Forbidden output guard</h3>
            </div>
            <StatusBadge tone={shell.forbiddenProjectionBlocked ? "bad" : "good"}>
              {shell.forbiddenProjectionBlocked ? "blocked" : "clear"}
            </StatusBadge>
          </div>
          <div className="guard-layout">
            <div>
              <p className="metric-value">{shell.forbiddenOutputKeys.length}</p>
              <p className="metric-label">blocked output keys</p>
            </div>
            <div className="guard-list" aria-label="Forbidden output keys">
              {shell.forbiddenOutputKeys.map((key) => (
                <code key={key}>{key}</code>
              ))}
            </div>
          </div>
        </section>
      </section>
    </main>
  );
}

function WorkflowSurfaceOverviewPanel({ overview }: { overview: WorkflowSurfaceOverviewViewModel }) {
  return (
    <section
      className="surface-band workflow-surface-overview"
      id="workflow-surface-overview"
      aria-labelledby="workflow-surface-overview-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">Workflow Surface Overview</p>
          <h3 id="workflow-surface-overview-title">Application, draft, plan, readiness</h3>
        </div>
        <StatusBadge tone={overview.canRenderSurfaceOverview ? "neutral" : "bad"}>
          {overview.canRenderSurfaceOverview ? "offline advisory" : "blocked"}
        </StatusBadge>
      </div>

      <div className="workflow-surface-overview-summary-grid" aria-label="Workflow surface overview summary">
        {overview.summary.map((metric) => (
          <WorkflowSurfaceOverviewMetricCard key={metric.metricId} metric={metric} />
        ))}
      </div>

      <article className="workflow-surface-overview-card">
        <div className="workflow-surface-overview-row-main">
          <div>
            <p className="eyebrow">{overview.overviewMode}</p>
            <h5>{overview.applicationId}</h5>
          </div>
          <StatusBadge tone="neutral">inspect only</StatusBadge>
        </div>
        <dl className="workflow-run-guard-meta">
          <div>
            <dt>Workflow definition</dt>
            <dd>{overview.workflowDefinitionId}</dd>
          </div>
          <div>
            <dt>Selected draft</dt>
            <dd>{overview.selectedDraftId}</dd>
          </div>
          <div>
            <dt>Latest run</dt>
            <dd>{overview.latestRunId}</dd>
          </div>
          <div>
            <dt>Request</dt>
            <dd>{overview.requestId}</dd>
          </div>
          <div>
            <dt>Audit</dt>
            <dd>{overview.auditRef}</dd>
          </div>
        </dl>
      </article>

      <div className="workflow-surface-overview-relation-grid" aria-label="Workflow surface overview relationship map">
        {overview.relationMap.map((relation) => (
          <WorkflowSurfaceOverviewRelationCard key={relation.relationId} relation={relation} />
        ))}
      </div>

      <div
        className="workflow-surface-overview-blocked-grid"
        aria-label="Workflow surface overview blocked capabilities"
      >
        {overview.blockedCapabilities.map((capability) => (
          <WorkflowSurfaceOverviewBlockedCapabilityCard
            key={capability.capabilityId}
            capability={capability}
          />
        ))}
      </div>

      <div className="workflow-surface-overview-stopline-grid" aria-label="Workflow surface overview stop lines">
        {overview.stopLines.map((stopLine) => (
          <WorkflowSurfaceOverviewStopLineCard key={stopLine.stopLineId} stopLine={stopLine} />
        ))}
      </div>
    </section>
  );
}

function WorkflowSurfaceOverviewMetricCard({ metric }: { metric: WorkflowSurfaceOverviewMetric }) {
  return (
    <article className="workflow-surface-overview-card">
      <span>{metric.label}</span>
      <strong>{metric.value}</strong>
      <StatusBadge tone={workflowSurfaceOverviewTone(metric.status)}>{metric.status}</StatusBadge>
      <p>{metric.summary}</p>
    </article>
  );
}

function WorkflowSurfaceOverviewRelationCard({ relation }: { relation: WorkflowSurfaceOverviewRelation }) {
  return (
    <article className="workflow-surface-overview-relation">
      <div className="workflow-surface-overview-row-main">
        <div>
          <p className="eyebrow">{relation.relationId}</p>
          <h5>{relation.label}</h5>
        </div>
        <StatusBadge tone={workflowSurfaceOverviewTone(relation.status)}>{relation.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Source</dt>
          <dd>{relation.sourceRef}</dd>
        </div>
        <div>
          <dt>Target</dt>
          <dd>{relation.targetRef}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{relation.auditRef}</dd>
        </div>
      </dl>
      <p>{relation.summary}</p>
    </article>
  );
}

function WorkflowSurfaceOverviewBlockedCapabilityCard({
  capability,
}: {
  capability: WorkflowSurfaceOverviewBlockedCapability;
}) {
  return (
    <article className="workflow-surface-overview-blocked-capability">
      <div className="workflow-surface-overview-row-main">
        <div>
          <p className="eyebrow">{capability.sourceSurface}</p>
          <h5>{capability.label}</h5>
        </div>
        <StatusBadge tone="bad">{capability.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Capability</dt>
          <dd>{capability.capabilityId}</dd>
        </div>
        <div>
          <dt>Missing prerequisite</dt>
          <dd>{capability.missingPrerequisite}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{capability.auditRef}</dd>
        </div>
      </dl>
      <p>{capability.summary}</p>
    </article>
  );
}

function WorkflowSurfaceOverviewStopLineCard({ stopLine }: { stopLine: WorkflowSurfaceOverviewStopLine }) {
  return (
    <article className="workflow-surface-overview-stopline">
      <div className="workflow-surface-overview-row-main">
        <div>
          <p className="eyebrow">{stopLine.stopLineId}</p>
          <h5>{stopLine.label}</h5>
        </div>
        <StatusBadge tone="bad">{stopLine.status}</StatusBadge>
      </div>
      <p>{stopLine.summary}</p>
    </article>
  );
}

function workflowSurfaceOverviewTone(status: WorkflowSurfaceOverviewStatus): "good" | "bad" | "neutral" {
  if (status === "blocked") {
    return "bad";
  }
  if (status === "ready") {
    return "good";
  }
  return "neutral";
}

function WorkflowScenarioInspectorPanel({
  inspector,
  selectedScenarioId,
  onSelectScenario,
}: {
  inspector: WorkflowScenarioInspectorViewModel;
  selectedScenarioId: string;
  onSelectScenario: (scenarioId: string) => void;
}) {
  return (
    <section
      className="surface-band workflow-scenario-inspector"
      id="workflow-scenario-inspector"
      aria-labelledby="workflow-scenario-inspector-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">Workflow Scenario Inspector</p>
          <h3 id="workflow-scenario-inspector-title">Scenario, input, output, blockers</h3>
        </div>
        <StatusBadge tone={inspector.canRenderScenarioInspector ? "neutral" : "bad"}>
          {inspector.canRenderScenarioInspector ? "offline advisory" : "blocked"}
        </StatusBadge>
      </div>

      <div className="workflow-scenario-selector-grid" aria-label="Workflow scenario selector">
        {inspector.scenarios.map((scenario) => (
          <WorkflowScenarioSelectorCard
            key={scenario.scenarioId}
            scenario={scenario}
            selected={scenario.scenarioId === selectedScenarioId}
            onSelectScenario={onSelectScenario}
          />
        ))}
      </div>

      <div className="workflow-scenario-summary-grid" aria-label="Workflow scenario summary">
        {inspector.summary.map((summary) => (
          <WorkflowScenarioSummaryCard key={summary.label} summary={summary} />
        ))}
      </div>

      <WorkflowScenarioDetailCard scenario={inspector.selectedScenario} inspector={inspector} />

      <div className="workflow-scenario-input-grid" aria-label="Workflow scenario input contract">
        {inspector.selectedScenario.inputContract.map((input) => (
          <WorkflowScenarioInputCard key={input.fieldId} input={input} />
        ))}
      </div>

      <div className="workflow-scenario-output-grid" aria-label="Workflow scenario expected outputs">
        {inspector.selectedScenario.expectedOutputs.map((output) => (
          <WorkflowScenarioOutputCard key={output.outputId} output={output} />
        ))}
      </div>

      <div className="workflow-scenario-relation-grid" aria-label="Workflow scenario relationship map">
        {inspector.relationMap.map((relation) => (
          <WorkflowScenarioRelationCard key={relation.relationId} relation={relation} />
        ))}
      </div>

      <div className="workflow-scenario-blocked-grid" aria-label="Workflow scenario blocked reasons">
        {inspector.blockedReasons.map((reason) => (
          <WorkflowScenarioBlockedReasonCard key={reason.reasonId} reason={reason} />
        ))}
      </div>

      <div className="workflow-scenario-stopline-grid" aria-label="Workflow scenario stop lines">
        {inspector.stopLines.map((stopLine) => (
          <WorkflowScenarioStopLineCard key={stopLine.stopLineId} stopLine={stopLine} />
        ))}
      </div>
    </section>
  );
}

function WorkflowScenarioSelectorCard({
  scenario,
  selected,
  onSelectScenario,
}: {
  scenario: WorkflowScenario;
  selected: boolean;
  onSelectScenario: (scenarioId: string) => void;
}) {
  return (
    <article
      className={`workflow-scenario-selector-card selection-row${selected ? " selected" : ""}`}
      role="button"
      tabIndex={0}
      aria-pressed={selected}
      onClick={() => onSelectScenario(scenario.scenarioId)}
      onKeyDown={(event) => handleSelectionRowKeyDown(event, scenario.scenarioId, onSelectScenario)}
    >
      <div className="workflow-scenario-row-main">
        <div>
          <p className="eyebrow">{scenario.scenarioKind}</p>
          <h5>{scenario.label}</h5>
        </div>
        <StatusBadge tone={selected ? "neutral" : workflowScenarioTone(scenario.requiresConfirmation ? "blocked" : "review_required")}>
          {selected ? "selected" : scenario.requiresConfirmation ? "confirmation" : "advisory"}
        </StatusBadge>
      </div>
      <p>{scenario.triggerSummary}</p>
      <small>{scenario.scenarioId}</small>
    </article>
  );
}

function WorkflowScenarioSummaryCard({ summary }: { summary: WorkflowScenarioSummary }) {
  return (
    <article className="workflow-scenario-card">
      <span>{summary.label}</span>
      <strong>{summary.value}</strong>
      <StatusBadge tone={workflowScenarioTone(summary.status)}>{summary.status}</StatusBadge>
      <p>{summary.summary}</p>
    </article>
  );
}

function WorkflowScenarioDetailCard({
  scenario,
  inspector,
}: {
  scenario: WorkflowScenario;
  inspector: WorkflowScenarioInspectorViewModel;
}) {
  return (
    <article className="workflow-scenario-card">
      <div className="workflow-scenario-row-main">
        <div>
          <p className="eyebrow">{inspector.scenarioMode}</p>
          <h5>{scenario.intent}</h5>
        </div>
        <StatusBadge tone={workflowScenarioTone(scenario.requiresConfirmation ? "blocked" : "review_required")}>
          {scenario.requiresConfirmation ? "confirmation required" : "advisory only"}
        </StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Application</dt>
          <dd>{scenario.applicationRef}</dd>
        </div>
        <div>
          <dt>Workflow definition</dt>
          <dd>{scenario.workflowDefinitionId}</dd>
        </div>
        <div>
          <dt>Selected draft</dt>
          <dd>{scenario.draftId}</dd>
        </div>
        <div>
          <dt>Latest run</dt>
          <dd>{scenario.runId}</dd>
        </div>
        <div>
          <dt>Risk</dt>
          <dd>{scenario.riskLevel}</dd>
        </div>
        <div>
          <dt>Scenario</dt>
          <dd>{inspector.selectedScenarioId}</dd>
        </div>
      </dl>
      <p>{scenario.validationFocus}</p>
    </article>
  );
}

function WorkflowScenarioInputCard({ input }: { input: WorkflowScenarioInputField }) {
  return (
    <article className="workflow-scenario-input">
      <div className="workflow-scenario-row-main">
        <div>
          <p className="eyebrow">{input.fieldId}</p>
          <h5>{input.label}</h5>
        </div>
        <StatusBadge tone="neutral">{input.required ? "required" : "optional"}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Source</dt>
          <dd>{input.sourceRef}</dd>
        </div>
      </dl>
      <p>{input.summary}</p>
    </article>
  );
}

function WorkflowScenarioOutputCard({ output }: { output: WorkflowScenarioExpectedOutput }) {
  return (
    <article className="workflow-scenario-output">
      <div className="workflow-scenario-row-main">
        <div>
          <p className="eyebrow">{output.outputId}</p>
          <h5>{output.label}</h5>
        </div>
        <StatusBadge tone={workflowScenarioTone(output.status)}>{output.status}</StatusBadge>
      </div>
      <p>{output.summary}</p>
    </article>
  );
}

function WorkflowScenarioRelationCard({ relation }: { relation: WorkflowScenarioRelation }) {
  return (
    <article className="workflow-scenario-relation">
      <div className="workflow-scenario-row-main">
        <div>
          <p className="eyebrow">{relation.relationId}</p>
          <h5>{relation.label}</h5>
        </div>
        <StatusBadge tone={workflowScenarioTone(relation.status)}>{relation.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Source</dt>
          <dd>{relation.sourceRef}</dd>
        </div>
        <div>
          <dt>Target</dt>
          <dd>{relation.targetRef}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{relation.auditRef}</dd>
        </div>
      </dl>
      <p>{relation.summary}</p>
    </article>
  );
}

function WorkflowScenarioBlockedReasonCard({ reason }: { reason: WorkflowScenarioBlockedReason }) {
  return (
    <article className="workflow-scenario-blocked-reason">
      <div className="workflow-scenario-row-main">
        <div>
          <p className="eyebrow">{reason.sourceSurface}</p>
          <h5>{reason.label}</h5>
        </div>
        <StatusBadge tone="bad">{reason.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Missing prerequisite</dt>
          <dd>{reason.missingPrerequisite}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{reason.auditRef}</dd>
        </div>
      </dl>
      <p>{reason.summary}</p>
    </article>
  );
}

function WorkflowScenarioStopLineCard({ stopLine }: { stopLine: WorkflowScenarioStopLine }) {
  return (
    <article className="workflow-scenario-stopline">
      <div className="workflow-scenario-row-main">
        <div>
          <p className="eyebrow">{stopLine.stopLineId}</p>
          <h5>{stopLine.label}</h5>
        </div>
        <StatusBadge tone="bad">{stopLine.status}</StatusBadge>
      </div>
      <p>{stopLine.summary}</p>
    </article>
  );
}

function workflowScenarioTone(status: WorkflowScenarioStatus): "good" | "bad" | "neutral" {
  if (status === "blocked") {
    return "bad";
  }
  if (status === "ready") {
    return "good";
  }
  return "neutral";
}

function AuditLogMetric({ metric }: { metric: AdminAuditLogMetric }) {
  return (
    <article className="audit-log-metric">
      <span>{metric.label}</span>
      <strong>{metric.value}</strong>
      <p>{metric.detail}</p>
    </article>
  );
}

function AuditEventRow({ event }: { event: AdminAuditEventRow }) {
  return (
    <article className="audit-event-row">
      <div className="audit-event-row-main">
        <div>
          <p className="eyebrow">{event.eventKind}</p>
          <h4>{event.auditRef}</h4>
        </div>
        <StatusBadge tone={event.decision === "denied" ? "bad" : "good"}>{event.decision}</StatusBadge>
      </div>
      <dl className="audit-event-row-meta">
        <div>
          <dt>Actor</dt>
          <dd>{event.actorSubjectRef}</dd>
        </div>
        <div>
          <dt>Resource</dt>
          <dd>{event.resourceRef}</dd>
        </div>
        <div>
          <dt>Failure</dt>
          <dd>{event.failureCode}</dd>
        </div>
        <div>
          <dt>Trace</dt>
          <dd>{event.traceId}</dd>
        </div>
        <div>
          <dt>Recorded</dt>
          <dd>{event.recordedAt}</dd>
        </div>
      </dl>
    </article>
  );
}

function AuditLogStatePreview({ state }: { state: AdminAuditLogStatePreview }) {
  return (
    <article className="audit-log-state">
      <div>
        <strong>{state.label}</strong>
        <span>{state.status}</span>
      </div>
      <p>{state.summary}</p>
      <small>
        items {state.itemCount} / failure {state.failureCode}
      </small>
    </article>
  );
}

function RunHistoryMetric({ metric }: { metric: WorkspaceRunHistoryMetric }) {
  return (
    <article className="run-history-metric">
      <span>{metric.label}</span>
      <strong>{metric.value}</strong>
      <p>{metric.detail}</p>
    </article>
  );
}

function handleSelectionRowKeyDown(
  event: KeyboardEvent<HTMLElement>,
  selectionId: string,
  onSelect: (selectionId: string) => void,
) {
  if (event.key !== "Enter" && event.key !== " ") {
    return;
  }
  event.preventDefault();
  onSelect(selectionId);
}

function RunRecordRow({
  run,
  selected,
  onSelectRun,
}: {
  run: WorkspaceRunRecordRow;
  selected: boolean;
  onSelectRun: (runId: string) => void;
}) {
  return (
    <article
      className={`run-record-row selection-row${selected ? " selected" : ""}`}
      role="button"
      tabIndex={0}
      aria-pressed={selected}
      onClick={() => onSelectRun(run.runId)}
      onKeyDown={(event) => handleSelectionRowKeyDown(event, run.runId, onSelectRun)}
    >
      <div className="run-record-row-main">
        <div>
          <p className="eyebrow">{run.applicationRef}</p>
          <h4>{run.runId}</h4>
        </div>
        <StatusBadge tone={selected ? "neutral" : run.status === "failed" ? "bad" : "good"}>
          {selected ? "selected" : run.status}
        </StatusBadge>
      </div>
      <dl className="run-record-row-meta">
        <div>
          <dt>Workflow</dt>
          <dd>{run.workflowDefinitionId}</dd>
        </div>
        <div>
          <dt>Failure</dt>
          <dd>{run.failureCode}</dd>
        </div>
        <div>
          <dt>Cost</dt>
          <dd>{run.estimatedCost}</dd>
        </div>
        <div>
          <dt>Trace</dt>
          <dd>{run.traceId}</dd>
        </div>
        <div>
          <dt>Started</dt>
          <dd>{run.startedAt}</dd>
        </div>
        <div>
          <dt>Completed</dt>
          <dd>{run.completedAt}</dd>
        </div>
      </dl>
    </article>
  );
}

function WorkflowRunDetailPanel({ detail }: { detail: WorkflowRunDetailViewModel }) {
  return (
    <div className="workflow-run-detail" aria-label="Workflow run detail read surface">
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">Workflow Run Detail</p>
          <h4>{detail.runId}</h4>
        </div>
        <StatusBadge tone={detail.canRenderRunDetail ? "good" : "bad"}>
          {detail.canRenderRunDetail ? "detail ready" : "blocked"}
        </StatusBadge>
      </div>

      <div className="workflow-run-detail-summary-grid" aria-label="Workflow run detail summary">
        <article className="workflow-run-detail-card">
          <span>Route</span>
          <strong>{detail.draftRouteId}</strong>
          <p>{detail.routePath}</p>
        </article>
        <article className="workflow-run-detail-card">
          <span>Request</span>
          <strong>{detail.requestId}</strong>
          <p>{detail.auditRef}</p>
        </article>
        <article className="workflow-run-detail-card">
          <span>Status</span>
          <strong>{detail.status}</strong>
          <p>failure {detail.failureCode}</p>
        </article>
        <article className="workflow-run-detail-card">
          <span>Trace</span>
          <strong>{detail.traceId}</strong>
          <p>{detail.workflowDefinitionId}</p>
        </article>
      </div>

      <div className="workflow-run-detail-summary-grid" aria-label="Workflow run input and output summaries">
        <WorkflowRunDetailSummaryCard summary={detail.inputSummary} />
        <WorkflowRunDetailSummaryCard summary={detail.outputSummary} />
        <WorkflowRunDetailSummaryCard summary={detail.costSummary} />
        <WorkflowRunDetailSummaryCard summary={detail.tokenSummary} />
      </div>

      <div className="workflow-run-timeline" aria-label="Workflow run state timeline">
        {detail.stateTimeline.map((event) => (
          <WorkflowRunTimelineEventCard key={event.eventId} event={event} />
        ))}
      </div>

      <div className="workflow-run-guard-grid" aria-label="Workflow run blocked capability previews">
        <WorkflowRunGuardPreviewCard preview={detail.blockedResultPreview} />
        <WorkflowRunGuardPreviewCard preview={detail.blockedReplayPreview} />
      </div>

      <div className="workflow-run-audit-list" aria-label="Workflow run audit references">
        {detail.auditRefs.map((auditRef) => (
          <code key={auditRef}>{auditRef}</code>
        ))}
      </div>
    </div>
  );
}

function WorkflowRunDetailSummaryCard({ summary }: { summary: WorkflowRunDetailSummary }) {
  return (
    <article className="workflow-run-detail-card">
      <span>{summary.label}</span>
      <strong>{summary.fields.join(", ")}</strong>
      <p>{summary.summary}</p>
    </article>
  );
}

function WorkflowRunTimelineEventCard({ event }: { event: WorkflowRunDetailTimelineEvent }) {
  return (
    <article className="workflow-run-timeline-event">
      <div className="workflow-run-detail-row-main">
        <div>
          <p className="eyebrow">{event.state}</p>
          <h5>{event.label}</h5>
        </div>
        <StatusBadge tone={event.state === "failed" || event.state === "tool_preview_blocked" ? "bad" : "neutral"}>
          {event.recordedAt}
        </StatusBadge>
      </div>
      <p>{event.summary}</p>
      <small>{event.auditRef}</small>
    </article>
  );
}

function WorkflowRunGuardPreviewCard({ preview }: { preview: WorkflowRunDetailGuardPreview }) {
  return (
    <article className="workflow-run-guard">
      <div className="workflow-run-detail-row-main">
        <div>
          <p className="eyebrow">{preview.guardId}</p>
          <h5>{preview.label}</h5>
        </div>
        <StatusBadge tone="bad">{preview.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Missing prerequisite</dt>
          <dd>{preview.missingPrerequisite}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{preview.auditRef}</dd>
        </div>
      </dl>
      <p>{preview.reason}</p>
    </article>
  );
}

function WorkflowBlockedActionPreviewPanel({ preview }: { preview: WorkflowBlockedActionPreviewViewModel }) {
  return (
    <div
      className="workflow-blocked-action-preview"
      id="workflow-blocked-action-preview"
      aria-label="Workflow blocked action preview read surface"
    >
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">Blocked Action Preview</p>
          <h4>{preview.toolActionId}</h4>
        </div>
        <StatusBadge tone={preview.canRenderBlockedActionPreview ? "bad" : "neutral"}>{preview.blockedState}</StatusBadge>
      </div>

      <div className="workflow-blocked-action-summary-grid" aria-label="Workflow blocked action summary">
        <article className="workflow-blocked-action-card">
          <span>Tool</span>
          <strong>{preview.toolRef}</strong>
          <p>{preview.actionKind}</p>
        </article>
        <article className="workflow-blocked-action-card">
          <span>Route</span>
          <strong>{preview.draftRouteId}</strong>
          <p>{preview.routePath}</p>
        </article>
        <article className="workflow-blocked-action-card">
          <span>Request</span>
          <strong>{preview.requestId}</strong>
          <p>{preview.auditRef}</p>
        </article>
        <article className="workflow-blocked-action-card">
          <span>Risk</span>
          <strong>{preview.riskLevel}</strong>
          <p>{preview.requiresConfirmation ? "future human review required" : "read-only metadata"}</p>
        </article>
      </div>

      <article className="workflow-blocked-action-card">
        <div className="workflow-blocked-row-main">
          <div>
            <p className="eyebrow">{preview.workflowDefinitionId}</p>
            <h5>{preview.nodeExecutionRef}</h5>
          </div>
          <StatusBadge tone="bad">{preview.relatedRunGuard.status}</StatusBadge>
        </div>
        <dl className="workflow-run-guard-meta">
          <div>
            <dt>Run</dt>
            <dd>{preview.runId}</dd>
          </div>
          <div>
            <dt>Related guard</dt>
            <dd>{preview.relatedRunGuard.guardId}</dd>
          </div>
        </dl>
        <p>{preview.policyReason}</p>
      </article>

      <div className="workflow-blocked-requirement-grid" aria-label="Workflow blocked action missing prerequisites">
        {preview.missingPrerequisites.map((requirement) => (
          <WorkflowBlockedActionRequirementCard key={requirement.requirementId} requirement={requirement} />
        ))}
      </div>

      <WorkflowConfirmationPlaceholderCard placeholder={preview.confirmationPlaceholder} />

      <div className="workflow-blocked-audit-grid" aria-label="Workflow blocked action audit trail">
        {preview.auditTrail.map((step) => (
          <WorkflowBlockedActionAuditStepCard key={step.stepId} step={step} />
        ))}
      </div>
    </div>
  );
}

function WorkflowBlockedActionRequirementCard({ requirement }: { requirement: WorkflowBlockedActionRequirement }) {
  return (
    <article className="workflow-blocked-requirement">
      <div className="workflow-blocked-row-main">
        <div>
          <p className="eyebrow">{requirement.requirementId}</p>
          <h5>{requirement.label}</h5>
        </div>
        <StatusBadge tone={requirement.status === "defined_not_connected" ? "neutral" : "bad"}>
          {requirement.status}
        </StatusBadge>
      </div>
      <p>{requirement.summary}</p>
    </article>
  );
}

function WorkflowConfirmationPlaceholderCard({
  placeholder,
}: {
  placeholder: WorkflowConfirmationPlaceholderPreview;
}) {
  return (
    <article className="workflow-confirmation-placeholder">
      <div className="workflow-blocked-row-main">
        <div>
          <p className="eyebrow">{placeholder.confirmationPlaceholderId}</p>
          <h5>{placeholder.requiredActionRef}</h5>
        </div>
        <StatusBadge tone="bad">{placeholder.humanReviewRequired ? "review required" : "read-only"}</StatusBadge>
      </div>
      <p>{placeholder.riskSummary}</p>
      <div className="workflow-confirmation-shape" aria-label="Workflow confirmation placeholder decision shape">
        {placeholder.requiredDecisionShape.map((field) => (
          <code key={field}>{field}</code>
        ))}
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Disabled reason</dt>
          <dd>{placeholder.disabledReason}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{placeholder.auditRef}</dd>
        </div>
      </dl>
    </article>
  );
}

function WorkflowConfirmationPlaceholderPanel({
  placeholder,
}: {
  placeholder: WorkflowConfirmationPlaceholderViewModel;
}) {
  return (
    <div
      className="workflow-confirmation-placeholder-read"
      id="workflow-confirmation-placeholder"
      aria-label="Workflow confirmation placeholder read surface"
    >
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">Confirmation Placeholder</p>
          <h4>{placeholder.confirmationPlaceholderId}</h4>
        </div>
        <StatusBadge tone={placeholder.canRenderConfirmationPlaceholder ? "bad" : "neutral"}>
          {placeholder.humanReviewRequired ? "human review required later" : "read-only"}
        </StatusBadge>
      </div>

      <div className="workflow-confirmation-summary-grid" aria-label="Workflow confirmation placeholder summary">
        <article className="workflow-confirmation-card">
          <span>Action</span>
          <strong>{placeholder.requiredActionRef}</strong>
          <p>{placeholder.actionKind}</p>
        </article>
        <article className="workflow-confirmation-card">
          <span>Run</span>
          <strong>{placeholder.requiredRunRef}</strong>
          <p>{placeholder.workflowDefinitionId}</p>
        </article>
        <article className="workflow-confirmation-card">
          <span>Route</span>
          <strong>{placeholder.draftRouteId}</strong>
          <p>{placeholder.routePath}</p>
        </article>
        <article className="workflow-confirmation-card">
          <span>Risk</span>
          <strong>{placeholder.riskLevel}</strong>
          <p>{placeholder.toolRef}</p>
        </article>
      </div>

      <article className="workflow-confirmation-card">
        <div className="workflow-confirmation-row-main">
          <div>
            <p className="eyebrow">{placeholder.nodeExecutionRef}</p>
            <h5>{placeholder.requiredActionRef}</h5>
          </div>
          <StatusBadge tone="bad">submission disabled</StatusBadge>
        </div>
        <p>{placeholder.riskSummary}</p>
        <p>{placeholder.policyReason}</p>
        <dl className="workflow-run-guard-meta">
          <div>
            <dt>Disabled reason</dt>
            <dd>{placeholder.disabledReason}</dd>
          </div>
          <div>
            <dt>Audit</dt>
            <dd>{placeholder.auditRef}</dd>
          </div>
        </dl>
      </article>

      <div className="workflow-confirmation-shape" aria-label="Workflow confirmation placeholder decision shape">
        {placeholder.requiredDecisionShape.map((field) => (
          <code key={field}>{field}</code>
        ))}
      </div>

      <div className="workflow-confirmation-field-grid" aria-label="Workflow confirmation placeholder decision fields">
        {placeholder.decisionFields.map((field) => (
          <WorkflowConfirmationDecisionFieldCard key={field.fieldId} field={field} />
        ))}
      </div>

      <div
        className="workflow-confirmation-prerequisite-grid"
        aria-label="Workflow confirmation placeholder prerequisites"
      >
        {placeholder.prerequisites.map((prerequisite) => (
          <WorkflowConfirmationPrerequisiteCard key={prerequisite.prerequisiteId} prerequisite={prerequisite} />
        ))}
      </div>

      <article className="workflow-confirmation-card">
        <div className="workflow-confirmation-row-main">
          <div>
            <p className="eyebrow">{placeholder.auditMetadata.policyRef}</p>
            <h5>{placeholder.auditMetadata.traceRef}</h5>
          </div>
          <StatusBadge tone="neutral">{placeholder.auditMetadata.sourceRouteId}</StatusBadge>
        </div>
        <dl className="workflow-run-guard-meta">
          <div>
            <dt>Request</dt>
            <dd>{placeholder.auditMetadata.requestId}</dd>
          </div>
          <div>
            <dt>Draft route</dt>
            <dd>{placeholder.auditMetadata.draftRouteId}</dd>
          </div>
          <div>
            <dt>Audit</dt>
            <dd>{placeholder.auditMetadata.auditRef}</dd>
          </div>
        </dl>
      </article>
    </div>
  );
}

function WorkflowConfirmationDecisionFieldCard({ field }: { field: WorkflowConfirmationDecisionField }) {
  return (
    <article className="workflow-confirmation-field">
      <div className="workflow-confirmation-row-main">
        <div>
          <p className="eyebrow">{field.fieldId}</p>
          <h5>{field.label}</h5>
        </div>
        <StatusBadge tone={field.required ? "bad" : "neutral"}>{field.required ? "required later" : "optional"}</StatusBadge>
      </div>
      <p>{field.source}</p>
    </article>
  );
}

function WorkflowConfirmationPrerequisiteCard({
  prerequisite,
}: {
  prerequisite: WorkflowConfirmationPlaceholderPrerequisite;
}) {
  return (
    <article className="workflow-confirmation-prerequisite">
      <div className="workflow-confirmation-row-main">
        <div>
          <p className="eyebrow">{prerequisite.prerequisiteId}</p>
          <h5>{prerequisite.label}</h5>
        </div>
        <StatusBadge tone={prerequisite.status === "defined_not_connected" ? "neutral" : "bad"}>
          {prerequisite.status}
        </StatusBadge>
      </div>
      <p>{prerequisite.summary}</p>
      <small>{prerequisite.auditRef}</small>
    </article>
  );
}

function WorkflowBlockedActionAuditStepCard({ step }: { step: WorkflowBlockedActionAuditStep }) {
  return (
    <article className="workflow-blocked-audit-step">
      <div className="workflow-blocked-row-main">
        <div>
          <p className="eyebrow">{step.stepId}</p>
          <h5>{step.label}</h5>
        </div>
        <StatusBadge tone={step.status === "blocked" ? "bad" : "neutral"}>{step.status}</StatusBadge>
      </div>
      <p>{step.summary}</p>
      <small>{step.auditRef}</small>
    </article>
  );
}

function RunHistoryStatePreview({ state }: { state: WorkspaceRunHistoryStatePreview }) {
  return (
    <article className="run-history-state">
      <div>
        <strong>{state.label}</strong>
        <span>{state.status}</span>
      </div>
      <p>{state.summary}</p>
      <small>
        items {state.itemCount} / failure {state.failureCode}
      </small>
    </article>
  );
}

function WorkflowDefinitionMetric({ metric }: { metric: WorkspaceWorkflowDefinitionsMetric }) {
  return (
    <article className="workflow-definition-metric">
      <span>{metric.label}</span>
      <strong>{metric.value}</strong>
      <p>{metric.detail}</p>
    </article>
  );
}

function WorkflowDefinitionRow({
  workflowDefinition,
  selected,
  createdDraftCount,
  onSelectWorkflowDefinition,
  onCreateDraftForWorkflowDefinition,
}: {
  workflowDefinition: WorkspaceWorkflowDefinitionRow;
  selected: boolean;
  createdDraftCount: number;
  onSelectWorkflowDefinition: (workflowDefinitionId: string) => void;
  onCreateDraftForWorkflowDefinition: (workflowDefinitionId: string) => void;
}) {
  return (
    <article
      className={`workflow-definition-row selection-row${selected ? " selected" : ""}`}
      role="button"
      tabIndex={0}
      aria-pressed={selected}
      onClick={() => onSelectWorkflowDefinition(workflowDefinition.workflowDefinitionId)}
      onKeyDown={(event) =>
        handleSelectionRowKeyDown(event, workflowDefinition.workflowDefinitionId, onSelectWorkflowDefinition)
      }
    >
      <div className="workflow-definition-row-main">
        <div>
          <p className="eyebrow">{workflowDefinition.applicationRef}</p>
          <h4>{workflowDefinition.workflowDefinitionId}</h4>
        </div>
        <StatusBadge tone={selected ? "neutral" : workflowDefinition.definitionStatus === "published" ? "good" : "neutral"}>
          {selected ? "selected" : workflowDefinition.definitionStatus}
        </StatusBadge>
      </div>
      <dl className="workflow-definition-row-meta">
        <div>
          <dt>Version</dt>
          <dd>{workflowDefinition.version}</dd>
        </div>
        <div>
          <dt>Nodes</dt>
          <dd>{workflowDefinition.nodeCount}</dd>
        </div>
        <div>
          <dt>Risk</dt>
          <dd>{workflowDefinition.riskLevel}</dd>
        </div>
        <div>
          <dt>Confirmation</dt>
          <dd>{workflowDefinition.requiresConfirmationCapable ? "capable" : "not required"}</dd>
        </div>
        <div>
          <dt>Updated</dt>
          <dd>{workflowDefinition.updatedAt}</dd>
        </div>
      </dl>
      <div className="workflow-definition-row-actions">
        <span>{createdDraftCount} local drafts</span>
        <button
          type="button"
          onClick={(event) => {
            event.stopPropagation();
            onCreateDraftForWorkflowDefinition(workflowDefinition.workflowDefinitionId);
          }}
        >
          Create draft
        </button>
      </div>
    </article>
  );
}

function WorkflowDefinitionDetailPanel({ detail }: { detail: WorkflowDefinitionDetailViewModel }) {
  return (
    <div className="workflow-definition-detail" aria-label="Workflow definition detail read surface">
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">Workflow Definition Detail</p>
          <h4>{detail.workflowDefinitionId}</h4>
        </div>
        <StatusBadge tone={detail.canRenderDefinitionDetail ? "good" : "bad"}>
          {detail.canRenderDefinitionDetail ? "detail ready" : "blocked"}
        </StatusBadge>
      </div>

      <div className="workflow-detail-summary-grid" aria-label="Workflow definition detail summary">
        <article className="workflow-detail-summary-card">
          <span>Route</span>
          <strong>{detail.draftRouteId}</strong>
          <p>{detail.routePath}</p>
        </article>
        <article className="workflow-detail-summary-card">
          <span>Request</span>
          <strong>{detail.requestId}</strong>
          <p>{detail.auditRef}</p>
        </article>
        <article className="workflow-detail-summary-card">
          <span>Risk</span>
          <strong>{detail.riskLevel}</strong>
          <p>{detail.requiresConfirmationCapable ? "confirmation capable" : "no confirmation marker"}</p>
        </article>
        <article className="workflow-detail-summary-card">
          <span>Source</span>
          <strong>{detail.sourceRouteId}</strong>
          <p>{detail.applicationRef}</p>
        </article>
      </div>

      <div className="workflow-detail-schema-grid" aria-label="Workflow definition input and output summaries">
        <WorkflowDefinitionSchemaSummary summary={detail.inputSummary} />
        <WorkflowDefinitionSchemaSummary summary={detail.outputSummary} />
      </div>

      <div className="workflow-detail-node-list" aria-label="Workflow definition nodes">
        {detail.nodes.map((node) => (
          <WorkflowDefinitionDetailNodeCard key={node.nodeId} node={node} />
        ))}
      </div>

      <div className="workflow-detail-edge-list" aria-label="Workflow definition edges">
        {detail.edges.map((edge) => (
          <WorkflowDefinitionDetailEdgeCard key={edge.edgeId} edge={edge} />
        ))}
      </div>

      <WorkflowDefinitionBlockedActionPreviewCard preview={detail.blockedActionPreview} />
    </div>
  );
}

function WorkflowDefinitionSchemaSummary({ summary }: { summary: WorkflowDefinitionDetailSchemaSummary }) {
  return (
    <article className="workflow-detail-schema-card">
      <span>{summary.label}</span>
      <strong>{summary.fields.join(", ")}</strong>
      <p>{summary.summary}</p>
    </article>
  );
}

function WorkflowDefinitionDetailNodeCard({ node }: { node: WorkflowDefinitionDetailNode }) {
  return (
    <article className="workflow-detail-node">
      <div className="workflow-detail-row-main">
        <div>
          <p className="eyebrow">{node.nodeType}</p>
          <h5>{node.label}</h5>
        </div>
        <StatusBadge tone={node.requiresConfirmation ? "neutral" : "good"}>
          {node.requiresConfirmation ? "confirmation marker" : "read-only"}
        </StatusBadge>
      </div>
      <dl className="workflow-detail-node-meta">
        <div>
          <dt>Input</dt>
          <dd>{node.inputSummary}</dd>
        </div>
        <div>
          <dt>Output</dt>
          <dd>{node.outputSummary}</dd>
        </div>
        <div>
          <dt>Risk</dt>
          <dd>{node.riskLevel}</dd>
        </div>
      </dl>
    </article>
  );
}

function WorkflowDefinitionDetailEdgeCard({ edge }: { edge: WorkflowDefinitionDetailEdge }) {
  return (
    <article className="workflow-detail-edge">
      <span>{edge.edgeId}</span>
      <strong>
        {edge.fromNodeId} to {edge.toNodeId}
      </strong>
      <p>{edge.conditionSummary}</p>
    </article>
  );
}

function WorkflowDefinitionBlockedActionPreviewCard({
  preview,
}: {
  preview: WorkflowDefinitionBlockedActionPreview;
}) {
  return (
    <article className="workflow-detail-blocked-action">
      <div className="workflow-detail-row-main">
        <div>
          <p className="eyebrow">{preview.toolRef}</p>
          <h5>{preview.toolActionId}</h5>
        </div>
        <StatusBadge tone="bad">{preview.blockedState}</StatusBadge>
      </div>
      <dl className="workflow-detail-node-meta">
        <div>
          <dt>Action</dt>
          <dd>{preview.actionKind}</dd>
        </div>
        <div>
          <dt>Risk</dt>
          <dd>{preview.riskLevel}</dd>
        </div>
        <div>
          <dt>Confirmation</dt>
          <dd>{preview.requiresConfirmation ? "required later" : "not required"}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{preview.auditRef}</dd>
        </div>
      </dl>
      <p>{preview.policyReason}</p>
    </article>
  );
}

function WorkflowDraftDesignerPanel({
  designer,
  selectedDraft,
  selectedDraftId,
  savedDraftConsumerState,
  draftEditDirty,
  onSelectDraft,
  onUpdateDraftLabel,
  onUpdateDraftSummary,
  onUpdateNodeLabel,
  onUpdateEdgeCondition,
  onAddNode,
  onMoveNode,
  onRemoveNode,
  onResetDraftEdits,
  onValidateDraft,
  onSaveDraft,
  onReadDraft,
}: {
  designer: WorkflowDraftDesignerViewModel;
  selectedDraft: WorkflowDraftDesignerDraft;
  selectedDraftId: string;
  savedDraftConsumerState: WorkflowSavedDraftConsumerState;
  draftEditDirty: boolean;
  onSelectDraft: (draftId: string) => void;
  onUpdateDraftLabel: (label: string) => void;
  onUpdateDraftSummary: (summary: string) => void;
  onUpdateNodeLabel: (nodeId: string, label: string) => void;
  onUpdateEdgeCondition: (edgeId: string, conditionSummary: string) => void;
  onAddNode: (nodeType: WorkflowDraftDesignerNode["nodeType"]) => void;
  onMoveNode: (nodeId: string, direction: WorkflowDraftNodeMoveDirection) => void;
  onRemoveNode: (nodeId: string) => void;
  onResetDraftEdits: () => void;
  onValidateDraft: () => void;
  onSaveDraft: () => void;
  onReadDraft: () => void;
}) {
  const canCallDevConsumer = savedDraftConsumerState.mode === "dev_saved_draft_http";
  const operationPending = ["saving", "validating", "reading"].includes(savedDraftConsumerState.status);
  const editStateLabel = draftEditDirty ? "unsaved local" : selectedDraft.localOnlyInteraction;
  return (
    <div
      className="workflow-draft-designer"
      id="workflow-draft-designer"
      aria-label="Workflow draft designer offline surface"
    >
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">Workflow Draft Designer</p>
          <h4>{selectedDraft.label}</h4>
        </div>
        <StatusBadge tone={designer.canRenderDraftDesigner ? "good" : "bad"}>
          {designer.canRenderDraftDesigner ? "offline designer ready" : "blocked"}
        </StatusBadge>
      </div>

      <div className="workflow-draft-template-grid" aria-label="Workflow draft templates">
        {designer.templates.map((template) => (
          <WorkflowDraftTemplateButton
            key={template.draftId}
            template={template}
            selected={template.draftId === selectedDraftId}
            onSelectDraft={onSelectDraft}
          />
        ))}
      </div>

      <div className="workflow-draft-summary-grid" aria-label="Selected workflow draft summary">
        <article className="workflow-draft-card">
          <span>Draft</span>
          <strong>{selectedDraft.draftId}</strong>
          <p>{selectedDraft.summary}</p>
        </article>
        <article className="workflow-draft-card">
          <span>Route</span>
          <strong>{selectedDraft.routeMetadata.draftRouteId}</strong>
          <p>{selectedDraft.routeMetadata.routePath}</p>
        </article>
        <article className="workflow-draft-card">
          <span>Source</span>
          <strong>{selectedDraft.routeMetadata.sourceRouteId}</strong>
          <p>{selectedDraft.workflowDefinitionId}</p>
        </article>
        <article className="workflow-draft-card">
          <span>Request</span>
          <strong>{selectedDraft.routeMetadata.requestId}</strong>
          <p>{selectedDraft.routeMetadata.auditRef}</p>
        </article>
      </div>

      <div className="workflow-draft-edit-grid" aria-label="Workflow draft local editing">
        <label className="workflow-draft-edit-field">
          <span>Draft name</span>
          <input
            type="text"
            value={selectedDraft.label}
            maxLength={160}
            disabled={operationPending}
            onChange={(event) => onUpdateDraftLabel(event.currentTarget.value)}
          />
        </label>
        <label className="workflow-draft-edit-field wide">
          <span>Draft summary</span>
          <textarea
            value={selectedDraft.summary}
            maxLength={4000}
            rows={3}
            disabled={operationPending}
            onChange={(event) => onUpdateDraftSummary(event.currentTarget.value)}
          />
        </label>
        <article className="workflow-draft-card workflow-draft-edit-state">
          <span>Local edit</span>
          <strong>{editStateLabel}</strong>
          <p>{draftEditDirty ? "Local draft changes are ready for validation or save." : selectedDraft.templateRef}</p>
          <button type="button" disabled={!draftEditDirty || operationPending} onClick={onResetDraftEdits}>
            Reset
          </button>
        </article>
      </div>

      <div className="workflow-draft-consumer-grid" aria-label="Saved workflow draft dev consumer">
        <article className="workflow-draft-card">
          <span>Saved state</span>
          <strong>{savedDraftConsumerState.sourceLabel}</strong>
          <p>{savedDraftConsumerState.summary}</p>
        </article>
        <article className="workflow-draft-card">
          <span>Version</span>
          <strong>{String(savedDraftConsumerState.currentDraftVersion)}</strong>
          <p>
            {savedDraftConsumerState.conflictDraftVersion === null
              ? savedDraftConsumerState.auditRef
              : `Conflict current version ${savedDraftConsumerState.conflictDraftVersion}`}
          </p>
        </article>
        <article className="workflow-draft-card">
          <span>Failure</span>
          <strong>{savedDraftConsumerState.failureCode ?? "none"}</strong>
          <p>{savedDraftConsumerState.requestId}</p>
        </article>
        <article className="workflow-draft-card workflow-draft-consumer-actions">
          <span>Dev consumer</span>
          <StatusBadge tone={workflowSavedDraftConsumerTone(savedDraftConsumerState.status)}>
            {savedDraftConsumerState.status}
          </StatusBadge>
          <div className="workflow-draft-action-row" aria-label="Saved draft dev consumer actions">
            <button type="button" disabled={!canCallDevConsumer || operationPending} onClick={onValidateDraft}>
              Validate
            </button>
            <button type="button" disabled={!canCallDevConsumer || operationPending} onClick={onSaveDraft}>
              Save
            </button>
            <button type="button" disabled={!canCallDevConsumer || operationPending} onClick={onReadDraft}>
              Read
            </button>
          </div>
        </article>
      </div>

      <div className="workflow-draft-structure-controls" aria-label="Workflow draft structure editing">
        <article className="workflow-draft-card">
          <span>Structure editing</span>
          <strong>{selectedDraft.nodes.length} nodes / {selectedDraft.edges.length} edges</strong>
          <p>Local graph edits rebuild preview edges and keep protected lanes visible for validation.</p>
        </article>
        <div className="workflow-draft-add-node-grid" aria-label="Add workflow draft node">
          {WORKFLOW_DRAFT_NODE_TYPE_OPTIONS.map((option) => (
            <button
              key={option.nodeType}
              type="button"
              className="workflow-draft-node-type-button"
              disabled={operationPending}
              onClick={() => onAddNode(option.nodeType)}
            >
              <span>{option.lane}</span>
              <strong>{option.label}</strong>
              <small>{option.summary}</small>
            </button>
          ))}
        </div>
      </div>

      <div className="workflow-draft-node-grid" aria-label="Workflow draft nodes">
        {selectedDraft.nodes.map((node, nodeIndex) => (
          <WorkflowDraftNodeCard
            key={node.nodeId}
            node={node}
            nodeIndex={nodeIndex}
            nodeCount={selectedDraft.nodes.length}
            canDelete={canRemoveWorkflowDraftNode(selectedDraft, node.nodeId)}
            editingDisabled={operationPending}
            onUpdateLabel={onUpdateNodeLabel}
            onMoveNode={onMoveNode}
            onRemoveNode={onRemoveNode}
          />
        ))}
      </div>

      <div className="workflow-draft-edge-grid" aria-label="Workflow draft edges">
        {selectedDraft.edges.map((edge) => (
          <WorkflowDraftEdgeCard
            key={edge.edgeId}
            edge={edge}
            editingDisabled={operationPending}
            onUpdateCondition={onUpdateEdgeCondition}
          />
        ))}
      </div>

      <div className="workflow-draft-readiness-grid" aria-label="Workflow draft readiness">
        {selectedDraft.readiness.map((readiness) => (
          <WorkflowDraftReadinessCard key={readiness.checkId} readiness={readiness} />
        ))}
      </div>

      <div className="workflow-draft-risk-grid" aria-label="Workflow draft risk summary">
        {selectedDraft.risks.map((risk) => (
          <WorkflowDraftRiskCard key={risk.riskId} risk={risk} />
        ))}
      </div>

      <div className="workflow-draft-blocked-grid" aria-label="Workflow draft blocked capabilities">
        {selectedDraft.blockedCapabilities.map((capability) => (
          <WorkflowDraftBlockedCapabilityCard key={capability.capabilityId} capability={capability} />
        ))}
      </div>
    </div>
  );
}

function workflowSavedDraftConsumerTone(status: WorkflowSavedDraftConsumerState["status"]): "good" | "bad" | "neutral" {
  if (status === "saved_dev_record" || status === "validation_ready") {
    return "good";
  }
  if (status === "version_conflict" || status === "save_failed" || status === "read_failed" || status === "validation_failed") {
    return "bad";
  }
  return "neutral";
}

function cloneWorkflowDraftForEditing(draft: WorkflowDraftDesignerDraft): WorkflowDraftDesignerDraft {
  return {
    ...draft,
    nodes: draft.nodes.map((node) => ({ ...node })),
    edges: draft.edges.map((edge) => ({ ...edge })),
    readiness: draft.readiness.map((readiness) => ({ ...readiness })),
    risks: draft.risks.map((risk) => ({ ...risk })),
    blockedCapabilities: draft.blockedCapabilities.map((capability) => ({ ...capability })),
    routeMetadata: { ...draft.routeMetadata },
  };
}

function workspaceDraftCreatedConsumerState(
  config: ReturnType<typeof readWorkflowSavedDraftConsumerConfig>,
  draft: WorkflowDraftDesignerDraft,
): WorkflowSavedDraftConsumerState {
  const initialState = initialWorkflowSavedDraftConsumerState(config);
  return {
    ...initialState,
    status: "unsaved_local",
    sourceLabel: "workspace draft",
    summary:
      config.mode === "dev_saved_draft_http"
        ? `Workspace draft ${draft.draftId} is ready for validation or save through the dev-only saved draft route.`
        : `Workspace draft ${draft.draftId} is local only until the dev-only saved draft route is enabled.`,
    failureCode: null,
    currentDraftVersion: 0,
    conflictDraftVersion: null,
    auditRef: draft.routeMetadata.auditRef,
    requestId: draft.routeMetadata.requestId,
  };
}

function buildWorkspaceCreatedDraft(
  workflowDefinitionId: string,
  designer: WorkflowDraftDesignerViewModel,
  existingDrafts: WorkflowDraftDesignerDraft[],
): WorkflowDraftDesignerDraft | null {
  const template = designer.templates.find(
    (draftTemplate) => draftTemplate.workflowDefinitionId === workflowDefinitionId,
  );
  const baseDraft = template
    ? designer.drafts.find((draft) => draft.draftId === template.draftId)
    : designer.drafts.find((draft) => draft.workflowDefinitionId === workflowDefinitionId);
  if (!baseDraft) {
    return null;
  }
  const nextDraftNumber =
    existingDrafts.filter((draft) => draft.workflowDefinitionId === workflowDefinitionId).length + 1;
  const draftNumberLabel = String(nextDraftNumber).padStart(2, "0");
  const createdDraftId = `draft_${workflowDefinitionId}_workspace_${draftNumberLabel}`;
  return {
    ...cloneWorkflowDraftForEditing(baseDraft),
    draftId: createdDraftId,
    templateRef: baseDraft.draftId,
    label: `${baseDraft.label} workspace ${draftNumberLabel}`,
    summary: `Workspace-created draft derived from ${workflowDefinitionId}; edit locally, validate, and save through the dev-only saved draft route before review.`,
    localOnlyInteraction: "local_edit",
    routeMetadata: {
      ...baseDraft.routeMetadata,
      requestId: `${baseDraft.routeMetadata.requestId}_workspace_${draftNumberLabel}`,
      auditRef: `${baseDraft.routeMetadata.auditRef}_workspace_${draftNumberLabel}`,
    },
  };
}

function buildLocalWorkflowDraftNode(
  draft: WorkflowDraftDesignerDraft,
  nodeType: WorkflowDraftDesignerNode["nodeType"],
): WorkflowDraftDesignerNode {
  const option = workflowDraftNodeTypeOption(nodeType);
  const nodeNumber = nextWorkflowDraftNodeNumber(draft, nodeType);
  const nodeNumberLabel = String(nodeNumber).padStart(2, "0");
  const requiresConfirmation = nodeType === "condition" || nodeType === "http_tool";
  return {
    nodeId: uniqueWorkflowDraftNodeId(draft, nodeType, nodeNumber),
    label: `${option.label} ${nodeNumberLabel}`,
    nodeType,
    lane: option.lane,
    readiness: requiresConfirmation ? "review_required" : "ready",
    inputSummary: workflowDraftNodeInputSummary(option),
    outputSummary: workflowDraftNodeOutputSummary(option),
    riskLevel: requiresConfirmation ? "medium" : "low",
    requiresConfirmation,
    previewOnlyReason: "Local structure edit only; workflow execution remains blocked.",
  };
}

function workflowDraftWithStructureEdits(
  draft: WorkflowDraftDesignerDraft,
  nodes: WorkflowDraftDesignerNode[],
): WorkflowDraftDesignerDraft {
  return {
    ...draft,
    nodes,
    edges: rebuildWorkflowDraftEdges(nodes, draft.edges),
    localOnlyInteraction: "local_edit",
  };
}

function insertWorkflowDraftNode(
  nodes: WorkflowDraftDesignerNode[],
  nextNode: WorkflowDraftDesignerNode,
): WorkflowDraftDesignerNode[] {
  if (nextNode.lane === "output") {
    return [...nodes, nextNode];
  }
  const firstOutputIndex = nodes.findIndex((node) => node.lane === "output");
  if (firstOutputIndex === -1) {
    return [...nodes, nextNode];
  }
  return [...nodes.slice(0, firstOutputIndex), nextNode, ...nodes.slice(firstOutputIndex)];
}

function canMoveWorkflowDraftNode(
  draft: WorkflowDraftDesignerDraft,
  nodeId: string,
  direction: WorkflowDraftNodeMoveDirection,
): boolean {
  const nodeIndex = draft.nodes.findIndex((node) => node.nodeId === nodeId);
  if (nodeIndex === -1) {
    return false;
  }
  return direction === "up" ? nodeIndex > 0 : nodeIndex < draft.nodes.length - 1;
}

function moveWorkflowDraftNode(
  nodes: WorkflowDraftDesignerNode[],
  nodeId: string,
  direction: WorkflowDraftNodeMoveDirection,
): WorkflowDraftDesignerNode[] {
  const nodeIndex = nodes.findIndex((node) => node.nodeId === nodeId);
  const nextIndex = direction === "up" ? nodeIndex - 1 : nodeIndex + 1;
  if (nodeIndex === -1 || nextIndex < 0 || nextIndex >= nodes.length) {
    return nodes;
  }
  const reorderedNodes = [...nodes];
  const movedNode = reorderedNodes[nodeIndex]!;
  reorderedNodes[nodeIndex] = reorderedNodes[nextIndex]!;
  reorderedNodes[nextIndex] = movedNode;
  return reorderedNodes;
}

function canRemoveWorkflowDraftNode(draft: WorkflowDraftDesignerDraft, nodeId: string): boolean {
  const node = draft.nodes.find((candidate) => candidate.nodeId === nodeId);
  if (!node || draft.nodes.length <= 3) {
    return false;
  }
  const remainingNodes = draft.nodes.filter((candidate) => candidate.nodeId !== nodeId);
  if (!hasWorkflowDraftLane(remainingNodes, "context") || !hasWorkflowDraftLane(remainingNodes, "model")) {
    return false;
  }
  if (countWorkflowDraftLane(remainingNodes, "output") < 2) {
    return false;
  }
  if (
    node.lane === "policy" &&
    countWorkflowDraftLane(draft.nodes, "policy") === 1 &&
    hasWorkflowDraftLane(draft.nodes, "preview")
  ) {
    return false;
  }
  if (
    node.lane === "preview" &&
    countWorkflowDraftLane(draft.nodes, "preview") === 1 &&
    hasWorkflowDraftLane(draft.nodes, "policy")
  ) {
    return false;
  }
  return rebuildWorkflowDraftEdges(remainingNodes, draft.edges).length >= 3;
}

function rebuildWorkflowDraftEdges(
  nodes: WorkflowDraftDesignerNode[],
  previousEdges: WorkflowDraftDesignerEdge[],
): WorkflowDraftDesignerEdge[] {
  const rebuiltEdges = nodes.slice(1).map((node, index) =>
    buildWorkflowDraftEdge(nodes[index]!, node, previousEdges),
  );
  if (rebuiltEdges.some((edge) => edge.edgeKind === "audit")) {
    return rebuiltEdges;
  }
  const outputNodes = nodes.filter((node) => node.lane === "output");
  if (outputNodes.length < 2) {
    return rebuiltEdges;
  }
  return [
    ...rebuiltEdges,
    buildWorkflowDraftEdge(
      outputNodes[outputNodes.length - 2]!,
      outputNodes[outputNodes.length - 1]!,
      previousEdges,
      "audit",
    ),
  ];
}

function buildWorkflowDraftEdge(
  fromNode: WorkflowDraftDesignerNode,
  toNode: WorkflowDraftDesignerNode,
  previousEdges: WorkflowDraftDesignerEdge[],
  forcedEdgeKind?: WorkflowDraftDesignerEdge["edgeKind"],
): WorkflowDraftDesignerEdge {
  const previousEdge = previousEdges.find(
    (edge) => edge.fromNodeId === fromNode.nodeId && edge.toNodeId === toNode.nodeId,
  );
  const edgeKind = forcedEdgeKind ?? workflowDraftEdgeKindForConnection(fromNode, toNode);
  return {
    edgeId: previousEdge?.edgeId ?? workflowDraftEdgeId(fromNode.nodeId, toNode.nodeId, edgeKind),
    fromNodeId: fromNode.nodeId,
    toNodeId: toNode.nodeId,
    edgeKind,
    conditionSummary:
      previousEdge?.conditionSummary ?? workflowDraftEdgeConditionSummary(fromNode, toNode, edgeKind),
  };
}

function workflowDraftEdgeKindForConnection(
  fromNode: WorkflowDraftDesignerNode,
  toNode: WorkflowDraftDesignerNode,
): WorkflowDraftDesignerEdge["edgeKind"] {
  if (toNode.lane === "output" && (fromNode.lane === "output" || workflowDraftNodeLooksLikeAudit(toNode))) {
    return "audit";
  }
  if (toNode.lane === "preview" || fromNode.lane === "preview") {
    return "preview";
  }
  if (toNode.lane === "policy" || fromNode.lane === "policy") {
    return "policy";
  }
  return "context";
}

function workflowDraftEdgeConditionSummary(
  fromNode: WorkflowDraftDesignerNode,
  toNode: WorkflowDraftDesignerNode,
  edgeKind: WorkflowDraftDesignerEdge["edgeKind"],
): string {
  if (edgeKind === "audit") {
    return "Sanitized output metadata remains visible in the audit path after local graph editing.";
  }
  if (edgeKind === "preview") {
    return "Preview-only metadata flows forward while execution stays blocked.";
  }
  if (edgeKind === "policy") {
    return "Risk-bearing output remains behind policy and confirmation review markers.";
  }
  return `${fromNode.label} passes sanitized context to ${toNode.label}.`;
}

function uniqueWorkflowDraftNodeId(
  draft: WorkflowDraftDesignerDraft,
  nodeType: WorkflowDraftDesignerNode["nodeType"],
  initialNumber: number,
): string {
  const draftKey = workflowDraftSafeKey(draft.draftId, 32);
  let nodeNumber = initialNumber;
  let candidate = "";
  const existingNodeIds = new Set(draft.nodes.map((node) => node.nodeId));
  do {
    candidate = `node_${draftKey}_${nodeType}_${String(nodeNumber).padStart(2, "0")}`;
    nodeNumber += 1;
  } while (existingNodeIds.has(candidate));
  return candidate;
}

function workflowDraftEdgeId(
  fromNodeId: string,
  toNodeId: string,
  edgeKind: WorkflowDraftDesignerEdge["edgeKind"],
): string {
  return `edge_${workflowDraftSafeKey(fromNodeId, 36)}_to_${workflowDraftSafeKey(toNodeId, 36)}_${edgeKind}`;
}

function workflowDraftSafeKey(value: string, maxLength: number): string {
  const normalized = value.toLowerCase().replace(/[^a-z0-9]+/g, "_").replace(/^_+|_+$/g, "");
  return (normalized || "local").slice(0, maxLength);
}

function nextWorkflowDraftNodeNumber(
  draft: WorkflowDraftDesignerDraft,
  nodeType: WorkflowDraftDesignerNode["nodeType"],
): number {
  return draft.nodes.filter((node) => node.nodeType === nodeType).length + 1;
}

function workflowDraftNodeTypeOption(
  nodeType: WorkflowDraftDesignerNode["nodeType"],
): WorkflowDraftNodeTypeOption {
  return WORKFLOW_DRAFT_NODE_TYPE_OPTIONS.find((option) => option.nodeType === nodeType) ??
    WORKFLOW_DRAFT_NODE_TYPE_OPTIONS[0]!;
}

function workflowDraftNodeInputSummary(option: WorkflowDraftNodeTypeOption): string {
  if (option.nodeType === "prompt") {
    return "Tenant ref, application ref, selection summary, and diagnostic summary.";
  }
  if (option.nodeType === "llm") {
    return "Sanitized prompt context, answer contract, and provider profile reference.";
  }
  if (option.nodeType === "condition") {
    return "Candidate action shape, risk level, and confirmation policy marker.";
  }
  if (option.nodeType === "http_tool") {
    return "Sanitized candidate action payload without raw tool request body.";
  }
  return "Answer summary, risk summary, audit refs, and review context.";
}

function workflowDraftNodeOutputSummary(option: WorkflowDraftNodeTypeOption): string {
  if (option.nodeType === "prompt") {
    return "Sanitized context packet for advisory reasoning.";
  }
  if (option.nodeType === "llm") {
    return "Advisory answer, candidate actions, risk summary, and audit refs.";
  }
  if (option.nodeType === "condition") {
    return "Review-required branch metadata without execution unlock.";
  }
  if (option.nodeType === "http_tool") {
    return "Preview-only action metadata and audit reference.";
  }
  return "Read-only advisory output or sanitized audit projection.";
}

function hasWorkflowDraftLane(
  nodes: WorkflowDraftDesignerNode[],
  lane: WorkflowDraftDesignerNode["lane"],
): boolean {
  return nodes.some((node) => node.lane === lane);
}

function countWorkflowDraftLane(
  nodes: WorkflowDraftDesignerNode[],
  lane: WorkflowDraftDesignerNode["lane"],
): number {
  return nodes.filter((node) => node.lane === lane).length;
}

function workflowDraftNodeLooksLikeAudit(node: WorkflowDraftDesignerNode): boolean {
  return `${node.nodeId} ${node.label}`.toLowerCase().includes("audit");
}

function WorkflowDraftTemplateButton({
  template,
  selected,
  onSelectDraft,
}: {
  template: WorkflowDraftDesignerTemplate;
  selected: boolean;
  onSelectDraft: (draftId: string) => void;
}) {
  return (
    <button
      type="button"
      className={`workflow-draft-template-button${selected ? " selected" : ""}`}
      aria-pressed={selected}
      onClick={() => onSelectDraft(template.draftId)}
    >
      <span>{template.workflowKind}</span>
      <strong>{template.label}</strong>
      <p>{template.summary}</p>
      <small>
        {template.status} / risk {template.riskLevel} / nodes {template.nodeCount}
      </small>
    </button>
  );
}

function WorkflowDraftNodeCard({
  node,
  nodeIndex,
  nodeCount,
  canDelete,
  editingDisabled,
  onUpdateLabel,
  onMoveNode,
  onRemoveNode,
}: {
  node: WorkflowDraftDesignerNode;
  nodeIndex: number;
  nodeCount: number;
  canDelete: boolean;
  editingDisabled: boolean;
  onUpdateLabel: (nodeId: string, label: string) => void;
  onMoveNode: (nodeId: string, direction: WorkflowDraftNodeMoveDirection) => void;
  onRemoveNode: (nodeId: string) => void;
}) {
  return (
    <article className="workflow-draft-node">
      <div className="workflow-draft-row-main">
        <div>
          <p className="eyebrow">
            {node.lane} / {node.nodeType}
          </p>
          <input
            className="workflow-draft-node-label-input"
            type="text"
            value={node.label}
            maxLength={160}
            disabled={editingDisabled}
            aria-label={`Node label ${node.nodeId}`}
            onChange={(event) => onUpdateLabel(node.nodeId, event.currentTarget.value)}
          />
        </div>
        <StatusBadge tone={node.readiness === "blocked" ? "bad" : node.readiness === "ready" ? "good" : "neutral"}>
          {node.readiness}
        </StatusBadge>
      </div>
      <div className="workflow-draft-node-actions" aria-label={`Structure controls ${node.nodeId}`}>
        <button
          type="button"
          disabled={editingDisabled || nodeIndex === 0}
          onClick={() => onMoveNode(node.nodeId, "up")}
        >
          Up
        </button>
        <button
          type="button"
          disabled={editingDisabled || nodeIndex === nodeCount - 1}
          onClick={() => onMoveNode(node.nodeId, "down")}
        >
          Down
        </button>
        <button
          type="button"
          disabled={editingDisabled || !canDelete}
          onClick={() => onRemoveNode(node.nodeId)}
        >
          Remove
        </button>
      </div>
      <dl className="workflow-detail-node-meta">
        <div>
          <dt>Input</dt>
          <dd>{node.inputSummary}</dd>
        </div>
        <div>
          <dt>Output</dt>
          <dd>{node.outputSummary}</dd>
        </div>
        <div>
          <dt>Risk</dt>
          <dd>{node.riskLevel}</dd>
        </div>
        <div>
          <dt>Preview</dt>
          <dd>{node.previewOnlyReason}</dd>
        </div>
      </dl>
    </article>
  );
}

function WorkflowDraftEdgeCard({
  edge,
  editingDisabled,
  onUpdateCondition,
}: {
  edge: WorkflowDraftDesignerEdge;
  editingDisabled: boolean;
  onUpdateCondition: (edgeId: string, conditionSummary: string) => void;
}) {
  return (
    <article className="workflow-draft-edge">
      <span>{edge.edgeKind}</span>
      <strong>
        {edge.fromNodeId} to {edge.toNodeId}
      </strong>
      <textarea
        className="workflow-draft-edge-condition-input"
        value={edge.conditionSummary}
        maxLength={4000}
        rows={3}
        disabled={editingDisabled}
        aria-label={`Edge condition ${edge.edgeId}`}
        onChange={(event) => onUpdateCondition(edge.edgeId, event.currentTarget.value)}
      />
    </article>
  );
}

function WorkflowDraftReadinessCard({ readiness }: { readiness: WorkflowDraftDesignerReadiness }) {
  return (
    <article className="workflow-draft-readiness">
      <div className="workflow-draft-row-main">
        <div>
          <p className="eyebrow">{readiness.checkId}</p>
          <h5>{readiness.label}</h5>
        </div>
        <StatusBadge tone={readiness.status === "blocked" ? "bad" : readiness.status === "ready" ? "good" : "neutral"}>
          {readiness.status}
        </StatusBadge>
      </div>
      <p>{readiness.summary}</p>
    </article>
  );
}

function WorkflowDraftRiskCard({ risk }: { risk: WorkflowDraftDesignerRisk }) {
  return (
    <article className="workflow-draft-risk">
      <div className="workflow-draft-row-main">
        <div>
          <p className="eyebrow">{risk.riskId}</p>
          <h5>{risk.label}</h5>
        </div>
        <StatusBadge tone={risk.riskLevel === "high" ? "bad" : risk.riskLevel === "low" ? "good" : "neutral"}>
          {risk.riskLevel}
        </StatusBadge>
      </div>
      <p>{risk.summary}</p>
      <small>{risk.requiresConfirmation ? "future human review required" : "advisory only"}</small>
    </article>
  );
}

function WorkflowDraftBlockedCapabilityCard({
  capability,
}: {
  capability: WorkflowDraftDesignerBlockedCapability;
}) {
  return (
    <article className="workflow-draft-blocked-capability">
      <div className="workflow-draft-row-main">
        <div>
          <p className="eyebrow">{capability.capabilityId}</p>
          <h5>{capability.label}</h5>
        </div>
        <StatusBadge tone="bad">{capability.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Missing prerequisite</dt>
          <dd>{capability.missingPrerequisite}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{capability.auditRef}</dd>
        </div>
      </dl>
      <p>{capability.summary}</p>
    </article>
  );
}

function WorkflowDraftValidationInspectorPanel({
  inspector,
}: {
  inspector: WorkflowDraftValidationInspectorViewModel;
}) {
  return (
    <div
      className="workflow-draft-validation-inspector"
      id="workflow-draft-validation-inspector"
      aria-label="Workflow draft validation inspector offline surface"
    >
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">Draft Validation Inspector</p>
          <h4>{inspector.inspectedDraftId}</h4>
        </div>
        <StatusBadge tone={inspector.validationStatus === "blocked" ? "bad" : "neutral"}>
          {inspector.validationStatus}
        </StatusBadge>
      </div>

      <div className="workflow-draft-validation-summary-grid" aria-label="Workflow draft validation summary">
        {inspector.summary.map((summary) => (
          <WorkflowDraftValidationSummaryCard key={summary.label} summary={summary} />
        ))}
      </div>

      <div className="workflow-draft-structural-check-grid" aria-label="Workflow draft structural checks">
        {inspector.structuralChecks.map((check) => (
          <WorkflowDraftStructuralCheckCard key={check.checkId} check={check} />
        ))}
      </div>

      <div className="workflow-draft-contract-check-grid" aria-label="Workflow draft contract checks">
        {inspector.contractChecks.map((check) => (
          <WorkflowDraftContractCheckCard key={check.checkId} check={check} />
        ))}
      </div>

      <div
        className="workflow-draft-validation-blocked-grid"
        aria-label="Workflow draft validation blocked capability checks"
      >
        {inspector.blockedCapabilityChecks.map((check) => (
          <WorkflowDraftBlockedCapabilityCheckCard key={check.checkId} check={check} />
        ))}
      </div>

      <article className="workflow-draft-validation-card">
        <div className="workflow-draft-validation-row-main">
          <div>
            <p className="eyebrow">{inspector.auditMetadata.sourceRouteId}</p>
            <h5>{inspector.auditMetadata.draftRouteId}</h5>
          </div>
          <StatusBadge tone="neutral">offline</StatusBadge>
        </div>
        <dl className="workflow-run-guard-meta">
          <div>
            <dt>Request</dt>
            <dd>{inspector.auditMetadata.requestId}</dd>
          </div>
          <div>
            <dt>Audit</dt>
            <dd>{inspector.auditMetadata.auditRef}</dd>
          </div>
          <div>
            <dt>Draft</dt>
            <dd>{inspector.auditMetadata.inspectedDraftId}</dd>
          </div>
        </dl>
      </article>
    </div>
  );
}

function WorkflowDraftValidationSummaryCard({ summary }: { summary: WorkflowDraftValidationSummary }) {
  return (
    <article className="workflow-draft-validation-card">
      <span>{summary.label}</span>
      <strong>{summary.value}</strong>
      <p>{summary.summary}</p>
    </article>
  );
}

function WorkflowDraftStructuralCheckCard({ check }: { check: WorkflowDraftStructuralCheck }) {
  return (
    <article className="workflow-draft-structural-check">
      <div className="workflow-draft-validation-row-main">
        <div>
          <p className="eyebrow">{check.checkId}</p>
          <h5>{check.label}</h5>
        </div>
        <StatusBadge tone={check.status === "blocked" ? "bad" : check.status === "passed" ? "good" : "neutral"}>
          {check.status}
        </StatusBadge>
      </div>
      <p>{check.summary}</p>
      <div className="workflow-draft-validation-evidence" aria-label="Workflow draft structural check evidence">
        {check.evidenceRefs.map((evidenceRef) => (
          <code key={evidenceRef}>{evidenceRef}</code>
        ))}
      </div>
    </article>
  );
}

function WorkflowDraftContractCheckCard({ check }: { check: WorkflowDraftContractCheck }) {
  return (
    <article className="workflow-draft-contract-check">
      <div className="workflow-draft-validation-row-main">
        <div>
          <p className="eyebrow">{check.checkId}</p>
          <h5>{check.label}</h5>
        </div>
        <StatusBadge tone={check.status === "passed" ? "good" : "neutral"}>{check.status}</StatusBadge>
      </div>
      <p>{check.summary}</p>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Required</dt>
          <dd>{check.requiredFields.join(", ")}</dd>
        </div>
        <div>
          <dt>Present</dt>
          <dd>{check.presentFields.join(", ") || "none"}</dd>
        </div>
        <div>
          <dt>Missing</dt>
          <dd>{check.missingFields.join(", ") || "none"}</dd>
        </div>
      </dl>
    </article>
  );
}

function WorkflowDraftBlockedCapabilityCheckCard({
  check,
}: {
  check: WorkflowDraftBlockedCapabilityCheck;
}) {
  return (
    <article className="workflow-draft-validation-blocked-check">
      <div className="workflow-draft-validation-row-main">
        <div>
          <p className="eyebrow">{check.capabilityId}</p>
          <h5>{check.label}</h5>
        </div>
        <StatusBadge tone="bad">{check.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Missing prerequisite</dt>
          <dd>{check.missingPrerequisite}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{check.auditRef}</dd>
        </div>
      </dl>
      <p>{check.summary}</p>
    </article>
  );
}

function WorkflowExecutionPlanPreviewPanel({
  preview,
}: {
  preview: WorkflowExecutionPlanPreviewViewModel;
}) {
  return (
    <div
      className="workflow-execution-plan-preview"
      id="workflow-execution-plan-preview"
      aria-label="Workflow execution plan preview offline surface"
    >
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">Execution Plan Preview</p>
          <h4>{preview.selectedDraftId}</h4>
        </div>
        <StatusBadge tone={preview.canRenderExecutionPlanPreview ? "neutral" : "bad"}>
          {preview.canRenderExecutionPlanPreview ? "offline preview" : "blocked"}
        </StatusBadge>
      </div>

      <div className="workflow-execution-plan-summary-grid" aria-label="Workflow execution plan summary">
        {preview.summary.map((summary) => (
          <WorkflowExecutionPlanSummaryCard key={summary.label} summary={summary} />
        ))}
      </div>

      <div className="workflow-execution-plan-stage-grid" aria-label="Workflow execution plan stage order">
        {preview.stageOrder.map((stage) => (
          <WorkflowExecutionPlanStageCard key={stage.stageId} stage={stage} />
        ))}
      </div>

      <div className="workflow-execution-plan-node-grid" aria-label="Workflow execution plan node to stage mapping">
        {preview.nodeStageMappings.map((mapping) => (
          <WorkflowExecutionPlanNodeMappingCard key={mapping.nodeId} mapping={mapping} />
        ))}
      </div>

      <div className="workflow-execution-plan-provider-grid" aria-label="Workflow execution plan provider requirements">
        {preview.providerProfileRequirements.map((requirement) => (
          <WorkflowExecutionPlanProviderRequirementCard
            key={requirement.requirementId}
            requirement={requirement}
          />
        ))}
      </div>

      <div className="workflow-execution-plan-gate-grid" aria-label="Workflow execution plan confirmation and audit gates">
        {preview.confirmationAuditGates.map((gate) => (
          <WorkflowExecutionPlanGateCard key={gate.gateId} gate={gate} />
        ))}
      </div>

      <div className="workflow-execution-plan-blocked-grid" aria-label="Workflow execution plan blocked reasons">
        {preview.blockedPlanReasons.map((reason) => (
          <WorkflowExecutionPlanBlockedReasonCard key={reason.reasonId} reason={reason} />
        ))}
      </div>

      <article className="workflow-execution-plan-card">
        <div className="workflow-execution-plan-row-main">
          <div>
            <p className="eyebrow">{preview.auditMetadata.sourceRouteId}</p>
            <h5>{preview.auditMetadata.draftRouteId}</h5>
          </div>
          <StatusBadge tone="neutral">{preview.validationStatus}</StatusBadge>
        </div>
        <dl className="workflow-run-guard-meta">
          <div>
            <dt>Validation route</dt>
            <dd>{preview.auditMetadata.validationRouteId}</dd>
          </div>
          <div>
            <dt>Request</dt>
            <dd>{preview.auditMetadata.requestId}</dd>
          </div>
          <div>
            <dt>Audit</dt>
            <dd>{preview.auditMetadata.auditRef}</dd>
          </div>
          <div>
            <dt>Draft</dt>
            <dd>{preview.auditMetadata.selectedDraftId}</dd>
          </div>
        </dl>
      </article>
    </div>
  );
}

function WorkflowExecutionPlanSummaryCard({ summary }: { summary: WorkflowExecutionPlanSummary }) {
  return (
    <article className="workflow-execution-plan-card">
      <span>{summary.label}</span>
      <strong>{summary.value}</strong>
      <p>{summary.summary}</p>
    </article>
  );
}

function WorkflowExecutionPlanStageCard({ stage }: { stage: WorkflowExecutionPlanStage }) {
  return (
    <article className="workflow-execution-plan-stage">
      <div className="workflow-execution-plan-row-main">
        <div>
          <p className="eyebrow">
            {stage.order} / {stage.stageKind}
          </p>
          <h5>{stage.label}</h5>
        </div>
        <StatusBadge tone={stage.status === "blocked" ? "bad" : stage.status === "ready" ? "good" : "neutral"}>
          {stage.status}
        </StatusBadge>
      </div>
      <p>{stage.summary}</p>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Nodes</dt>
          <dd>{stage.nodeIds.join(", ") || "none"}</dd>
        </div>
        <div>
          <dt>Blocked reason</dt>
          <dd>{stage.blockedReason}</dd>
        </div>
      </dl>
    </article>
  );
}

function WorkflowExecutionPlanNodeMappingCard({ mapping }: { mapping: WorkflowExecutionPlanNodeMapping }) {
  return (
    <article className="workflow-execution-plan-node">
      <div className="workflow-execution-plan-row-main">
        <div>
          <p className="eyebrow">{mapping.stageId}</p>
          <h5>{mapping.label}</h5>
        </div>
        <StatusBadge tone={mapping.requiresConfirmation ? "bad" : "neutral"}>{mapping.executionMode}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Node</dt>
          <dd>{mapping.nodeId}</dd>
        </div>
        <div>
          <dt>Type</dt>
          <dd>{mapping.nodeType}</dd>
        </div>
        <div>
          <dt>Provider</dt>
          <dd>{mapping.providerProfileRef}</dd>
        </div>
        <div>
          <dt>Input</dt>
          <dd>{mapping.inputSummary}</dd>
        </div>
        <div>
          <dt>Output</dt>
          <dd>{mapping.outputSummary}</dd>
        </div>
      </dl>
    </article>
  );
}

function WorkflowExecutionPlanProviderRequirementCard({
  requirement,
}: {
  requirement: WorkflowExecutionPlanProviderRequirement;
}) {
  return (
    <article className="workflow-execution-plan-provider">
      <div className="workflow-execution-plan-row-main">
        <div>
          <p className="eyebrow">{requirement.requirementId}</p>
          <h5>{requirement.label}</h5>
        </div>
        <StatusBadge tone={requirement.status === "blocked" ? "bad" : "neutral"}>{requirement.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Profile</dt>
          <dd>{requirement.providerProfileRef}</dd>
        </div>
        <div>
          <dt>Nodes</dt>
          <dd>{requirement.nodeIds.join(", ") || "none"}</dd>
        </div>
        <div>
          <dt>Missing prerequisite</dt>
          <dd>{requirement.missingPrerequisite}</dd>
        </div>
      </dl>
      <p>{requirement.summary}</p>
    </article>
  );
}

function WorkflowExecutionPlanGateCard({ gate }: { gate: WorkflowExecutionPlanGate }) {
  return (
    <article className="workflow-execution-plan-gate">
      <div className="workflow-execution-plan-row-main">
        <div>
          <p className="eyebrow">{gate.gateKind}</p>
          <h5>{gate.label}</h5>
        </div>
        <StatusBadge tone={gate.status === "blocked" ? "bad" : "neutral"}>{gate.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Before stage</dt>
          <dd>{gate.requiredBeforeStageId}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{gate.auditRef}</dd>
        </div>
      </dl>
      <p>{gate.summary}</p>
    </article>
  );
}

function WorkflowExecutionPlanBlockedReasonCard({
  reason,
}: {
  reason: WorkflowExecutionPlanBlockedReason;
}) {
  return (
    <article className="workflow-execution-plan-blocked-reason">
      <div className="workflow-execution-plan-row-main">
        <div>
          <p className="eyebrow">{reason.blockedCapability}</p>
          <h5>{reason.label}</h5>
        </div>
        <StatusBadge tone="bad">{reason.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Missing prerequisite</dt>
          <dd>{reason.missingPrerequisite}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{reason.auditRef}</dd>
        </div>
      </dl>
      <p>{reason.summary}</p>
    </article>
  );
}

function WorkflowRuntimeReadinessInspectorPanel({
  readiness,
}: {
  readiness: WorkflowRuntimeReadinessInspectorViewModel;
}) {
  return (
    <div
      className="workflow-runtime-readiness-inspector"
      id="workflow-runtime-readiness-inspector"
      aria-label="Workflow runtime readiness inspector offline surface"
    >
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">Runtime Readiness Inspector</p>
          <h4>{readiness.selectedDraftId}</h4>
        </div>
        <StatusBadge tone={readiness.canRenderRuntimeReadinessInspector ? "bad" : "neutral"}>
          {readiness.canRenderRuntimeReadinessInspector ? "blocked readiness" : "missing evidence"}
        </StatusBadge>
      </div>

      <div className="workflow-runtime-readiness-summary-grid" aria-label="Workflow runtime readiness summary">
        {readiness.summary.map((summary) => (
          <WorkflowRuntimeReadinessSummaryCard key={summary.label} summary={summary} />
        ))}
      </div>

      <div className="workflow-runtime-readiness-prerequisite-grid" aria-label="Workflow runtime prerequisites">
        {readiness.runtimePrerequisites.map((prerequisite) => (
          <WorkflowRuntimeReadinessPrerequisiteCard
            key={prerequisite.prerequisiteId}
            prerequisite={prerequisite}
          />
        ))}
      </div>

      <div className="workflow-runtime-readiness-blocker-grid" aria-label="Workflow runtime readiness blockers">
        {readiness.readinessBlockers.map((blocker) => (
          <WorkflowRuntimeReadinessBlockerCard key={blocker.blockerId} blocker={blocker} />
        ))}
      </div>

      <div className="workflow-runtime-readiness-gate-grid" aria-label="Workflow runtime implementation gates">
        {readiness.implementationGates.map((gate) => (
          <WorkflowRuntimeReadinessGateCard key={gate.gateId} gate={gate} />
        ))}
      </div>

      <article className="workflow-runtime-readiness-card">
        <div className="workflow-runtime-readiness-row-main">
          <div>
            <p className="eyebrow">{readiness.auditMetadata.sourcePageId}</p>
            <h5>{readiness.auditMetadata.readinessRouteId}</h5>
          </div>
          <StatusBadge tone={readiness.forbiddenProjectionBlocked ? "bad" : "neutral"}>
            {readiness.forbiddenProjectionBlocked ? "guard active" : "metadata only"}
          </StatusBadge>
        </div>
        <dl className="workflow-run-guard-meta">
          <div>
            <dt>Plan route</dt>
            <dd>{readiness.auditMetadata.planRouteId}</dd>
          </div>
          <div>
            <dt>Request</dt>
            <dd>{readiness.auditMetadata.requestId}</dd>
          </div>
          <div>
            <dt>Audit</dt>
            <dd>{readiness.auditMetadata.auditRef}</dd>
          </div>
          <div>
            <dt>Draft</dt>
            <dd>{readiness.auditMetadata.selectedDraftId}</dd>
          </div>
        </dl>
      </article>
    </div>
  );
}

function WorkflowRuntimeReadinessSummaryCard({ summary }: { summary: WorkflowRuntimeReadinessSummary }) {
  return (
    <article className="workflow-runtime-readiness-card">
      <span>{summary.label}</span>
      <strong>{summary.value}</strong>
      <p>{summary.summary}</p>
    </article>
  );
}

function WorkflowRuntimeReadinessPrerequisiteCard({
  prerequisite,
}: {
  prerequisite: WorkflowRuntimeReadinessPrerequisite;
}) {
  return (
    <article className="workflow-runtime-readiness-prerequisite">
      <div className="workflow-runtime-readiness-row-main">
        <div>
          <p className="eyebrow">{prerequisite.area}</p>
          <h5>{prerequisite.label}</h5>
        </div>
        <StatusBadge tone={workflowRuntimeReadinessTone(prerequisite.status)}>{prerequisite.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Evidence</dt>
          <dd>{prerequisite.currentEvidence}</dd>
        </div>
        <div>
          <dt>Missing prerequisite</dt>
          <dd>{prerequisite.missingPrerequisite}</dd>
        </div>
        <div>
          <dt>Source refs</dt>
          <dd>{prerequisite.sourceRefs.join(", ")}</dd>
        </div>
      </dl>
      <p>{prerequisite.summary}</p>
    </article>
  );
}

function WorkflowRuntimeReadinessBlockerCard({ blocker }: { blocker: WorkflowRuntimeReadinessBlocker }) {
  return (
    <article className="workflow-runtime-readiness-blocker">
      <div className="workflow-runtime-readiness-row-main">
        <div>
          <p className="eyebrow">{blocker.area}</p>
          <h5>{blocker.label}</h5>
        </div>
        <StatusBadge tone="bad">{blocker.severity}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Source</dt>
          <dd>{blocker.sourceRef}</dd>
        </div>
        <div>
          <dt>Missing prerequisite</dt>
          <dd>{blocker.missingPrerequisite}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{blocker.auditRef}</dd>
        </div>
      </dl>
      <p>{blocker.summary}</p>
    </article>
  );
}

function WorkflowRuntimeReadinessGateCard({ gate }: { gate: WorkflowRuntimeReadinessGate }) {
  return (
    <article className="workflow-runtime-readiness-gate">
      <div className="workflow-runtime-readiness-row-main">
        <div>
          <p className="eyebrow">{gate.gateKind}</p>
          <h5>{gate.label}</h5>
        </div>
        <StatusBadge tone={workflowRuntimeReadinessTone(gate.status)}>{gate.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Required before</dt>
          <dd>{gate.requiredBefore}</dd>
        </div>
        <div>
          <dt>Evidence refs</dt>
          <dd>{gate.evidenceRefs.join(", ")}</dd>
        </div>
      </dl>
      <p>{gate.summary}</p>
    </article>
  );
}

function workflowRuntimeReadinessTone(status: WorkflowRuntimeReadinessStatus): "good" | "bad" | "neutral" {
  if (status === "blocked") {
    return "bad";
  }
  if (status === "satisfied") {
    return "good";
  }
  return "neutral";
}

function WorkflowDefinitionStatePreview({ state }: { state: WorkspaceWorkflowDefinitionsStatePreview }) {
  return (
    <article className="workflow-definition-state">
      <div>
        <strong>{state.label}</strong>
        <span>{state.status}</span>
      </div>
      <p>{state.summary}</p>
      <small>
        items {state.itemCount} / failure {state.failureCode}
      </small>
    </article>
  );
}

function UsageQuotaLimit({ limit }: { limit: WorkspaceUsageQuotaLimit }) {
  return (
    <article className="usage-quota-limit">
      <span>{limit.label}</span>
      <strong>{limit.used}</strong>
      <p>
        limit {limit.value} / {limit.detail}
      </p>
    </article>
  );
}

function UsageQuotaSnapshot({ snapshot }: { snapshot: WorkspaceUsageQuotaSnapshot }) {
  return (
    <article className="usage-quota-snapshot-card">
      <span>{snapshot.label}</span>
      <strong>{snapshot.value}</strong>
      <p>{snapshot.detail}</p>
    </article>
  );
}

function UsageQuotaStatePreview({ state }: { state: WorkspaceUsageQuotaStatePreview }) {
  return (
    <article className="usage-quota-state">
      <div>
        <strong>{state.label}</strong>
        <span>{state.status}</span>
      </div>
      <p>{state.summary}</p>
      <small>
        items {state.itemCount} / failure {state.failureCode}
      </small>
    </article>
  );
}

function ApiKeyMetric({ metric }: { metric: WorkspaceApiKeysMetric }) {
  return (
    <article className="api-key-metric">
      <span>{metric.label}</span>
      <strong>{metric.value}</strong>
      <p>{metric.detail}</p>
    </article>
  );
}

function ApiKeyRow({ apiKey }: { apiKey: WorkspaceApiKeyRow }) {
  return (
    <article className="api-key-row">
      <div className="api-key-row-main">
        <div>
          <p className="eyebrow">{apiKey.ownerSubjectRef}</p>
          <h4>{apiKey.apiKeyId}</h4>
        </div>
        <StatusBadge tone={apiKey.state === "active" ? "good" : "neutral"}>{apiKey.state}</StatusBadge>
      </div>
      <div className="api-key-scopes" aria-label="API key scopes">
        {apiKey.scopes.map((scope) => (
          <code key={scope}>{scope}</code>
        ))}
      </div>
      <dl className="api-key-row-meta">
        <div>
          <dt>Created</dt>
          <dd>{apiKey.createdAt}</dd>
        </div>
        <div>
          <dt>Expires</dt>
          <dd>{apiKey.expiresAt ?? "not set"}</dd>
        </div>
        <div>
          <dt>Last used</dt>
          <dd>{apiKey.lastUsedAt ?? "not recorded"}</dd>
        </div>
      </dl>
    </article>
  );
}

function ApiKeyStatePreview({ state }: { state: WorkspaceApiKeysStatePreview }) {
  return (
    <article className="api-key-state">
      <div>
        <strong>{state.label}</strong>
        <span>{state.status}</span>
      </div>
      <p>{state.summary}</p>
      <small>
        items {state.itemCount} / failure {state.failureCode}
      </small>
    </article>
  );
}

function WorkflowApplicationDetailPanel({ detail }: { detail: WorkflowApplicationDetailViewModel }) {
  return (
    <div
      className="workflow-application-detail"
      id="workflow-application-detail"
      aria-label="Workflow application detail read surface"
    >
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">Application Detail</p>
          <h4>{detail.displayName}</h4>
        </div>
        <StatusBadge tone={detail.canRenderApplicationDetail ? "good" : "bad"}>
          {detail.canRenderApplicationDetail ? "detail ready" : "blocked"}
        </StatusBadge>
      </div>

      <div className="workflow-application-summary-grid" aria-label="Workflow application identity">
        <article className="workflow-application-detail-card">
          <span>Application</span>
          <strong>{detail.applicationId}</strong>
          <p>{detail.displayName}</p>
        </article>
        <article className="workflow-application-detail-card">
          <span>Tenant and owner</span>
          <strong>{detail.tenantRef}</strong>
          <p>{detail.ownerSubjectRef}</p>
        </article>
        <article className="workflow-application-detail-card">
          <span>Application type</span>
          <strong>{detail.applicationType}</strong>
          <p>{detail.lifecycleStatus}</p>
        </article>
        <article className="workflow-application-detail-card">
          <span>Provider profile</span>
          <strong>{detail.providerProfileRef}</strong>
          <p>{detail.lastRunStatus}</p>
        </article>
      </div>

      <div className="workflow-application-risk-grid" aria-label="Workflow application route and risk metadata">
        <WorkflowApplicationRiskCard riskSummary={detail.riskSummary} />
        <WorkflowApplicationRouteMetadataCard routeMetadata={detail.routeMetadata} />
        <article className="workflow-application-detail-card">
          <span>Latest workflow</span>
          <strong>{detail.latestWorkflowDefinitionRef}</strong>
          <p>updated {detail.updatedAt}</p>
        </article>
        <article className="workflow-application-detail-card">
          <span>Latest run ref</span>
          <strong>{detail.latestRunRef}</strong>
          <p>read-only pointer; replay and resume stay blocked</p>
        </article>
      </div>

      <div className="workflow-application-blocked-grid" aria-label="Workflow application blocked capabilities">
        {detail.blockedCapabilities.map((capability) => (
          <WorkflowApplicationBlockedCapabilityCard key={capability.capabilityId} capability={capability} />
        ))}
      </div>
    </div>
  );
}

function WorkflowApplicationRiskCard({ riskSummary }: { riskSummary: WorkflowApplicationRiskSummary }) {
  return (
    <article className="workflow-application-detail-card">
      <span>{riskSummary.label}</span>
      <strong>{riskSummary.riskLevel}</strong>
      <p>{riskSummary.summary}</p>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Confirmation</dt>
          <dd>{riskSummary.requiresConfirmationCapable ? "capability marker only" : "not required by fixture"}</dd>
        </div>
        <div>
          <dt>Policy</dt>
          <dd>{riskSummary.policyRef}</dd>
        </div>
      </dl>
    </article>
  );
}

function WorkflowApplicationRouteMetadataCard({
  routeMetadata,
}: {
  routeMetadata: WorkflowApplicationRouteMetadata;
}) {
  return (
    <article className="workflow-application-detail-card">
      <span>Route metadata</span>
      <strong>{routeMetadata.draftRouteId}</strong>
      <p>{routeMetadata.routePath}</p>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Source route</dt>
          <dd>{routeMetadata.sourceRouteId}</dd>
        </div>
        <div>
          <dt>Request</dt>
          <dd>{routeMetadata.requestId}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{routeMetadata.auditRef}</dd>
        </div>
      </dl>
    </article>
  );
}

function WorkflowApplicationBlockedCapabilityCard({
  capability,
}: {
  capability: WorkflowApplicationBlockedCapabilityPreview;
}) {
  return (
    <article className="workflow-application-blocked-capability">
      <div className="workflow-application-row-main">
        <div>
          <p className="eyebrow">{capability.capabilityId}</p>
          <h5>{capability.label}</h5>
        </div>
        <StatusBadge tone="bad">{capability.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Missing prerequisite</dt>
          <dd>{capability.missingPrerequisite}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{capability.auditRef}</dd>
        </div>
      </dl>
      <p>{capability.reason}</p>
    </article>
  );
}

function ApplicationMetric({ metric }: { metric: WorkspaceApplicationsMetric }) {
  return (
    <article className="application-metric">
      <span>{metric.label}</span>
      <strong>{metric.value}</strong>
      <p>{metric.detail}</p>
    </article>
  );
}

function ApplicationRow({
  application,
  selected,
  onSelectApplication,
}: {
  application: WorkspaceApplicationRow;
  selected: boolean;
  onSelectApplication: (applicationRef: string) => void;
}) {
  return (
    <article
      className={`application-row selection-row${selected ? " selected" : ""}`}
      role="button"
      tabIndex={0}
      aria-pressed={selected}
      onClick={() => onSelectApplication(application.applicationRef)}
      onKeyDown={(event) => handleSelectionRowKeyDown(event, application.applicationRef, onSelectApplication)}
    >
      <div className="application-row-main">
        <div>
          <p className="eyebrow">{application.applicationKind}</p>
          <h4>{application.displayName}</h4>
        </div>
        <StatusBadge tone={selected ? "neutral" : application.lastRunStatus === "blocked" ? "bad" : "good"}>
          {selected ? "selected" : application.lastRunStatus}
        </StatusBadge>
      </div>
      <dl className="application-row-meta">
        <div>
          <dt>Application</dt>
          <dd>{application.applicationRef}</dd>
        </div>
        <div>
          <dt>Owner</dt>
          <dd>{application.ownerSubjectRef}</dd>
        </div>
        <div>
          <dt>Workflow</dt>
          <dd>{application.latestWorkflowDefinitionRef}</dd>
        </div>
        <div>
          <dt>Updated</dt>
          <dd>{application.updatedAt}</dd>
        </div>
      </dl>
    </article>
  );
}

function ApplicationStatePreview({ state }: { state: WorkspaceApplicationsStatePreview }) {
  return (
    <article className="application-state">
      <div>
        <strong>{state.label}</strong>
        <span>{state.status}</span>
      </div>
      <p>{state.summary}</p>
      <small>
        items {state.itemCount} / failure {state.failureCode}
      </small>
    </article>
  );
}

function TenantFact({ fact }: { fact: AdminTenantOverviewFact }) {
  return (
    <article className="tenant-fact">
      <span>{fact.label}</span>
      <strong>{fact.value}</strong>
      <p>{fact.detail}</p>
    </article>
  );
}

function TenantStatePreview({ state }: { state: AdminTenantOverviewStatePreview }) {
  return (
    <article className="tenant-state">
      <div>
        <strong>{state.label}</strong>
        <span>{state.status}</span>
      </div>
      <p>{state.summary}</p>
      <small>
        items {state.itemCount} / failure {state.failureCode}
      </small>
    </article>
  );
}

function Fact({ label, value }: { label: string; value: string }) {
  return (
    <div className="fact">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function LiveReadSourceStatus({ state, baseUrl }: { state: ControlPlaneReadDevLiveLoadState; baseUrl: string }) {
  const tone = state.status === "failed" ? "bad" : state.status === "ready" ? "good" : "neutral";
  return (
    <section className="live-read-source" aria-label="Read data source">
      <div>
        <p className="eyebrow">Read Source</p>
        <h3>{state.mode === "dev_live_http" ? "Dev live HTTP" : "Offline fixtures"}</h3>
        <p>{state.message}</p>
      </div>
      <dl>
        <div>
          <dt>Base URL</dt>
          <dd>{state.mode === "dev_live_http" ? baseUrl : "not used"}</dd>
        </div>
        <div>
          <dt>Auth</dt>
          <dd>{state.mode === "dev_live_http" ? "dev fake header" : "offline view model"}</dd>
        </div>
        <div>
          <dt>Database</dt>
          <dd>detached</dd>
        </div>
      </dl>
      <StatusBadge tone={tone}>{state.status}</StatusBadge>
    </section>
  );
}

function RouteCard({ route }: { route: ControlPlaneReadRouteCard }) {
  return (
    <article className="route-card">
      <div className="card-title-row">
        <h4>{route.label}</h4>
        <StatusBadge tone="neutral">{route.surface}</StatusBadge>
      </div>
      <p className="route-path">{route.path}</p>
      <dl className="route-meta">
        <div>
          <dt>Scope</dt>
          <dd>{route.requiredScope}</dd>
        </div>
        <div>
          <dt>Model</dt>
          <dd>{route.readModel}</dd>
        </div>
        <div>
          <dt>Pagination</dt>
          <dd>{route.pagination}</dd>
        </div>
      </dl>
    </article>
  );
}

function StatePreview({ state }: { state: ControlPlaneReadStatePreview }) {
  return (
    <article className="state-card">
      <div className="card-title-row">
        <h4>{state.label}</h4>
        <StatusBadge tone={state.tone}>{state.status}</StatusBadge>
      </div>
      <p>{state.summary}</p>
      <dl className="state-meta">
        <div>
          <dt>Items</dt>
          <dd>{state.itemCount}</dd>
        </div>
        <div>
          <dt>Failure</dt>
          <dd>{state.failureCode ?? "none"}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{state.auditRef}</dd>
        </div>
      </dl>
    </article>
  );
}

function StatusBadge({ children, tone }: { children: string; tone: "good" | "bad" | "neutral" }) {
  return <span className={`status-badge ${tone}`}>{children}</span>;
}
