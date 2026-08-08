import { lazy, Suspense, useEffect, useMemo, useRef, useState, type KeyboardEvent } from "react";

import {
  buildAdminTenantOverviewViewModel,
} from "../features/control-plane-read/adminTenantOverview";
import {
  buildAdminAuditLogViewModel,
} from "../features/control-plane-read/adminAuditLog";
import { buildAdminOperationsReviewViewModel } from "../features/control-plane-read/adminOperationsReview";
import { buildAdminProviderDeploymentReviewViewModel } from "../features/control-plane-read/adminProviderDeploymentReview";
import {
  initialControlPlaneReadDevLiveLoadState,
  loadControlPlaneReadDevLiveCollections,
  normalizeActiveWorkspaceId,
  readControlPlaneReadDevLiveConfig,
  type ControlPlaneReadDevLiveLoadState,
} from "../features/control-plane-read/devLiveReadConsumer";
import {
  archiveWorkflowDraftDevRecord,
  continueLocalWorkflowDraftAfterVersionConflict,
  emptyWorkflowSavedDraftLibraryFilters,
  initialWorkflowSavedDraftLifecycleOperationState,
  initialWorkflowSavedDraftConsumerState,
  initialWorkflowSavedDraftListState,
  listWorkflowDraftDevRecords,
  mergeWorkflowSavedDraftListPage,
  nextWorkflowSavedDraftExpectedVersion,
  openWorkflowDraftDevRecord,
  readWorkflowDraftDevRecord,
  readWorkflowSavedDraftConsumerConfig,
  saveWorkflowDraftDevRecord,
  unarchiveWorkflowDraftDevRecord,
  validateWorkflowDraftDevRecord,
  workflowSavedDraftRequestIsCurrent,
  workflowSavedDraftConflictRequiresResolution,
  type WorkflowSavedDraftLibraryFilters,
  type WorkflowSavedDraftLifecycleOperationState,
  type WorkflowSavedDraftLifecycleState,
  type WorkflowSavedDraftListState,
  type WorkflowSavedDraftSummary,
  type WorkflowSavedDraftConsumerState,
  type WorkflowSavedDraftConflictReviewSummary,
} from "../features/control-plane-read/savedWorkflowDraftConsumer";
import type { WorkflowSavedDraftRevisionRestoreResult } from "../features/control-plane-read/workflowSavedDraftRevisionConsumer";
import {
  buildWorkflowExecutorV0Draft,
  evaluateWorkflowExecutorEligibility,
  initialWorkflowExecutorConsumerState,
  readWorkflowExecutorConsumerConfig,
  readWorkflowRunDevRecord,
  startWorkflowRunDevRecord,
  type WorkflowExecutorConsumerState,
} from "../features/control-plane-read/workflowExecutorConsumer";
import { WorkflowExecutorPanel } from "../features/control-plane-read/workflowExecutorPanel";
import {
  createWorkflowHTTPToolActionPlan,
  decideWorkflowHTTPToolActionPlan,
  evaluateWorkflowHTTPToolActionEligibility,
  initialWorkflowHTTPToolActionConsumerState,
  readWorkflowHTTPToolActionConsumerConfig,
  readWorkflowHTTPToolActionPlan,
  readWorkflowHTTPToolActionPlanReference,
  rememberWorkflowHTTPToolActionPlanReference,
  workflowHTTPToolActionPermissions,
  type WorkflowHTTPToolActionConsumerState,
  type WorkflowHTTPToolHumanDecision,
  type WorkflowHTTPToolPublicArguments,
} from "../features/control-plane-read/workflowHTTPToolActionConsumer";
import { WorkflowHTTPToolActionPanel } from "../features/control-plane-read/workflowHTTPToolActionPanel";
import {
  executeWorkflowHTTPToolActionPlan,
  initialWorkflowHTTPToolExecutionState,
  type WorkflowHTTPToolExecutionState,
} from "../features/control-plane-read/workflowHTTPToolExecutionConsumer.ts";
import { WorkflowHTTPToolExecutionPanel } from "../features/control-plane-read/workflowHTTPToolExecutionPanel.tsx";
import { buildModelGatewayOverviewViewModel } from "../features/control-plane-read/modelGatewayOverview";
import { ModelGatewayOverviewPanel } from "../features/control-plane-read/modelGatewayOverviewPanel";
import { buildModelGatewayRouteEvidenceViewModel } from "../features/control-plane-read/modelGatewayRouteEvidence";
import { ModelGatewayRouteEvidencePanel } from "../features/control-plane-read/modelGatewayRouteEvidencePanel";
import { buildModelGatewayUsageAuditEvidenceViewModel } from "../features/control-plane-read/modelGatewayUsageAuditEvidence";
import { ModelGatewayUsageAuditEvidencePanel } from "../features/control-plane-read/modelGatewayUsageAuditEvidencePanel";
import { buildModelGatewayEvidenceReviewViewModel } from "../features/control-plane-read/modelGatewayEvidenceReview";
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
  readApplicationCatalogConfig,
  type ApplicationCatalogRecord,
} from "../features/control-plane-read/applicationCatalogConsumer";
import type { ApplicationCatalogSnapshot } from "../features/control-plane-read/applicationCatalogPanel";
import { buildApplicationDevelopmentWorkspaceContext } from "../features/control-plane-read/applicationDevelopmentWorkspace";
import {
  type WorkflowApplicationBlockedCapabilityPreview,
  type WorkflowApplicationDetailViewModel,
  type WorkflowApplicationRiskSummary,
  type WorkflowApplicationRouteMetadata,
} from "../features/control-plane-read/workflowApplicationDetail";
import {
  buildWorkspaceApiKeysViewModel,
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
  type WorkflowDraftDesignerLayout,
  type WorkflowDraftDesignerNode,
  type WorkflowDraftDesignerReadiness,
  type WorkflowDraftDesignerRisk,
  type WorkflowDraftDesignerTemplate,
  type WorkflowDraftDesignerViewModel,
} from "../features/control-plane-read/workflowDraftDesigner";
import {
  buildDerivedWorkflowDraft,
  canDeriveSavedWorkflowDraft,
  cloneWorkflowDraftForEditing,
} from "../features/control-plane-read/workflowSavedDraftDerivation";
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
import {
  type WorkflowSurfaceOverviewBlockedCapability,
  type WorkflowSurfaceOverviewMetric,
  type WorkflowSurfaceOverviewRelation,
  type WorkflowSurfaceOverviewStatus,
  type WorkflowSurfaceOverviewStopLine,
  type WorkflowSurfaceOverviewViewModel,
} from "../features/control-plane-read/workflowSurfaceOverview";
import { WorkflowWorkspaceReviewPanel } from "../features/control-plane-read/workflowWorkspaceReviewPanel";
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
  buildWorkspaceOperationsInboxViewModel,
  type WorkspaceOperationsInboxItem,
} from "../features/control-plane-read/workspaceOperationsInbox";
import { WorkspaceOperationsInboxPanel } from "../features/control-plane-read/workspaceOperationsInboxPanel";
import { WorkspaceProductOverviewPanel } from "../features/control-plane-read/workspaceProductOverviewPanel";
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
import { ProductNavigation } from "./ProductNavigation";
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
const applicationCatalogConfig = readApplicationCatalogConfig();
const savedDraftConsumerConfig = readWorkflowSavedDraftConsumerConfig();
const workflowExecutorConsumerConfig = readWorkflowExecutorConsumerConfig();
const workflowHTTPToolActionConsumerConfig = readWorkflowHTTPToolActionConsumerConfig();
const workflowHTTPToolPermissions = workflowHTTPToolActionPermissions(workflowHTTPToolActionConsumerConfig);
const WorkflowDraftDesignerPanel = lazy(() =>
  import("../features/control-plane-read/workflowDraftDesignerPanel").then((module) => ({
    default: module.WorkflowDraftDesignerPanel,
  })),
);
const WorkflowSavedDraftRevisionPanel = lazy(() =>
  import("../features/control-plane-read/workflowSavedDraftRevisionPanel").then((module) => ({
    default: module.WorkflowSavedDraftRevisionPanel,
  })),
);
const WorkflowUserWorkspaceHomePanel = lazy(() =>
  import("../features/control-plane-read/workflowUserWorkspaceHomePanel").then((module) => ({
    default: module.WorkflowUserWorkspaceHomePanel,
  })),
);
const AdminControlPlaneWorkspace = lazy(() => import("../features/control-plane-read/adminControlPlaneWorkspace"));
const ModelGatewayEvidenceReviewPanel = lazy(() => import("../features/control-plane-read/modelGatewayEvidenceReviewPanel").then((module) => ({ default: module.ModelGatewayEvidenceReviewPanel })));
const ApplicationCatalogPanel = lazy(() => import("../features/control-plane-read/applicationCatalogPanel").then((module) => ({ default: module.ApplicationCatalogPanel })));
const ApplicationDevelopmentWorkspacePanel = lazy(() => import("../features/control-plane-read/applicationDevelopmentWorkspacePanel"));
const ApplicationDevelopmentWorkspaceSurface = lazy(() => import("../features/control-plane-read/applicationDevelopmentWorkspaceSurface"));
const ApplicationRuntimeReviewWorkspace = lazy(() => import("../features/control-plane-read/applicationRuntimeReviewWorkspace"));
const WorkflowReviewWorkspace = lazy(() => import("../features/control-plane-read/workflowReviewWorkspace"));
const WorkflowRAGExecutionPanel = lazy(() => import("../features/control-plane-read/workflowRAGExecutionPanel"));
const WorkflowReviewHandoffPanel = lazy(() => import("../features/control-plane-read/workflowReviewHandoffPanel").then((module) => ({ default: module.WorkflowReviewHandoffPanel })));
const DEFAULT_WORKFLOW_EXECUTOR_INPUT = "请根据当前工作流草案生成一条仅供人工审查的建议，并明确说明任何不确定性。";

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
    nodeType: "rag_retrieval",
    lane: "retrieval",
    label: "RAG Retrieval",
    summary: "Binds one exact immutable application knowledge snapshot version.",
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
  const [activeWorkspaceId, setActiveWorkspaceId] = useState(
    () => normalizeActiveWorkspaceId(devLiveConfig.workspaceId ?? "") ?? "workspace_demo",
  );
  const activeDevLiveConfig = useMemo(
    () => ({ ...devLiveConfig, workspaceId: activeWorkspaceId }),
    [activeWorkspaceId],
  );
  const activeSavedDraftConsumerConfig = useMemo(
    () => ({ ...savedDraftConsumerConfig, workspaceId: activeWorkspaceId }),
    [activeWorkspaceId],
  );
  const [devLiveState, setDevLiveState] = useState<ControlPlaneReadDevLiveLoadState>(() =>
    initialControlPlaneReadDevLiveLoadState(activeDevLiveConfig),
  );
  const [selectedApplicationRef, setSelectedApplicationRef] = useState<string | null>(null);
  const [applicationCatalogSnapshot, setApplicationCatalogSnapshot] = useState<ApplicationCatalogSnapshot | null>(null);
  const [selectedWorkflowDefinitionId, setSelectedWorkflowDefinitionId] = useState<string | null>(null);
  const [selectedRunId, setSelectedRunId] = useState<string | null>(null);
  const [selectedWorkflowDraftId, setSelectedWorkflowDraftId] = useState<string | null>(null);
  const [selectedWorkflowScenarioId, setSelectedWorkflowScenarioId] = useState<string | null>(null);
  const [savedDraftConsumerState, setSavedDraftConsumerState] = useState<WorkflowSavedDraftConsumerState>(() =>
    initialWorkflowSavedDraftConsumerState(activeSavedDraftConsumerConfig),
  );
  const [savedDraftLibraryLifecycle, setSavedDraftLibraryLifecycle] =
    useState<WorkflowSavedDraftLifecycleState>("active");
  const [savedDraftLibraryFilters, setSavedDraftLibraryFilters] =
    useState<WorkflowSavedDraftLibraryFilters>(() => emptyWorkflowSavedDraftLibraryFilters());
  const [savedDraftListStates, setSavedDraftListStates] = useState<
    Record<WorkflowSavedDraftLifecycleState, WorkflowSavedDraftListState>
  >(() => ({
    active: initialWorkflowSavedDraftListState(activeSavedDraftConsumerConfig, "", "active"),
    archived: initialWorkflowSavedDraftListState(activeSavedDraftConsumerConfig, "", "archived"),
  }));
  const savedDraftListRequestGenerationRef = useRef<Record<WorkflowSavedDraftLifecycleState, number>>({
    active: 0,
    archived: 0,
  });
  const savedDraftLifecycleOperationGenerationRef = useRef(0);
  const savedDraftOpenRequestGenerationRef = useRef(0);
  const [savedDraftLifecycleOperation, setSavedDraftLifecycleOperation] =
    useState<WorkflowSavedDraftLifecycleOperationState>(() =>
      initialWorkflowSavedDraftLifecycleOperationState()
    );
  const savedDraftListState = savedDraftListStates[savedDraftLibraryLifecycle];
  const activeSavedDraftListState = savedDraftListStates.active;
  const pendingSavedDraftConsumerStateRef = useRef<{
    draftId: string;
    state: WorkflowSavedDraftConsumerState;
  } | null>(null);
  const [workspaceCreatedDrafts, setWorkspaceCreatedDrafts] = useState<WorkflowDraftDesignerDraft[]>([]);
  const [editableWorkflowDraft, setEditableWorkflowDraft] = useState<WorkflowDraftDesignerDraft | null>(null);
  const [workflowDraftEditDirty, setWorkflowDraftEditDirty] = useState(false);
  const [workflowExecutorState, setWorkflowExecutorState] = useState<WorkflowExecutorConsumerState>(() =>
    initialWorkflowExecutorConsumerState(workflowExecutorConsumerConfig),
  );
  const [workflowExecutorInput, setWorkflowExecutorInput] = useState(DEFAULT_WORKFLOW_EXECUTOR_INPUT);
  const [workflowExecutorModel, setWorkflowExecutorModel] = useState("");
  const [workflowExecutorConditionValues, setWorkflowExecutorConditionValues] = useState<Record<string, boolean>>({});
  const [workflowHTTPToolActionState, setWorkflowHTTPToolActionState] = useState<WorkflowHTTPToolActionConsumerState>(() =>
    initialWorkflowHTTPToolActionConsumerState(workflowHTTPToolActionConsumerConfig),
  );
  const [workflowHTTPToolResourceKey, setWorkflowHTTPToolResourceKey] = useState("");
  const [workflowHTTPToolLocale, setWorkflowHTTPToolLocale] = useState("");
  const [workflowHTTPToolExecutionState, setWorkflowHTTPToolExecutionState] = useState<WorkflowHTTPToolExecutionState>(() =>
    initialWorkflowHTTPToolExecutionState(workflowHTTPToolActionConsumerConfig),
  );
  const [workflowHTTPToolExecutionInput, setWorkflowHTTPToolExecutionInput] = useState("Review the approved resource and return a bounded advisory summary.");
  const [workflowHTTPToolExecutionModel, setWorkflowHTTPToolExecutionModel] = useState("");
  const [workflowRAGOperationPending, setWorkflowRAGOperationPending] = useState(false);
  const [workflowRunHistoryRefreshKey, setWorkflowRunHistoryRefreshKey] = useState(0);
  const workflowExecutorOperationPending =
    workflowExecutorState.status === "starting" || workflowExecutorState.status === "reading";
  const workflowHTTPToolActionOperationPending =
    workflowHTTPToolActionState.status === "creating" ||
    workflowHTTPToolActionState.status === "reading" ||
    workflowHTTPToolActionState.status === "deciding";
  const workflowHTTPToolOperationPending = workflowHTTPToolActionOperationPending ||
    workflowHTTPToolExecutionState.status === "executing";

  useEffect(() => {
    if (activeDevLiveConfig.mode !== "dev_live_http") {
      return;
    }
    let cancelled = false;
    setDevLiveState({
      status: "loading",
      mode: "dev_live_http",
      message: `Loading workspace ${activeWorkspaceId} through the development/test read boundary.`,
    });
    loadControlPlaneReadDevLiveCollections(activeDevLiveConfig)
      .then((collections) => {
        if (cancelled) {
          return;
        }
        setDevLiveState({
          status: "ready",
          mode: "dev_live_http",
          message: `Workspace ${activeWorkspaceId} loaded sanitized read envelopes.`,
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
  }, [activeDevLiveConfig, activeWorkspaceId]);

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
  const workspaceReadApplications = useMemo(
    () => buildWorkspaceApplicationsViewModel(liveCollections["application-summary-list-route"]),
    [liveCollections],
  );
  const workspaceApplications = useMemo(
    () => buildWorkspaceApplicationsViewModel(
      liveCollections["application-summary-list-route"],
      applicationCatalogConfig.mode === "dev_application_catalog_http"
        ? {
            records: applicationCatalogSnapshot?.records ?? [],
            summary: applicationCatalogSnapshot?.summary ?? "Loading the authoritative application catalog.",
          }
        : undefined,
    ),
    [applicationCatalogSnapshot, liveCollections],
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
  const workspaceOperationsInbox = useMemo(
    () => buildWorkspaceOperationsInboxViewModel({
      activeWorkspaceId,
      referenceTime: new Date().toISOString(),
      sourceSnapshotReady:
        activeDevLiveConfig.mode !== "dev_live_http" || devLiveState.status === "ready",
      workspaceApplications: workspaceReadApplications,
      workspaceApiKeys,
      workspaceWorkflowDefinitions,
      workspaceRunHistory,
    }),
    [
      activeWorkspaceId,
      activeDevLiveConfig.mode,
      devLiveState.status,
      workspaceReadApplications,
      workspaceApiKeys,
      workspaceWorkflowDefinitions,
      workspaceRunHistory,
    ],
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
        savedDraftConsumerState,
        savedDraftListStatus: activeSavedDraftListState.status,
        savedDraftListFailureCode: activeSavedDraftListState.failureCode,
        savedDraftSummaries: activeSavedDraftListState.summaries,
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
      savedDraftConsumerState,
      activeSavedDraftListState.failureCode,
      activeSavedDraftListState.status,
      activeSavedDraftListState.summaries,
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
    savedDraftConflictReviewSummary,
    workflowReviewHandoff,
  } = workflowWorkspaceContext;
  const applicationCatalogLive = applicationCatalogConfig.mode === "dev_application_catalog_http";
  const selectedApplicationCatalogRecord = applicationCatalogSnapshot?.records.find(
    (record) => record.applicationId === selectedApplicationRef,
  ) ?? null;
  const applicationDevelopmentWorkspaceContext = useMemo(
    () => buildApplicationDevelopmentWorkspaceContext({
      applicationId: applicationCatalogLive
        ? selectedApplicationCatalogRecord?.applicationId ?? ""
        : selectedApplication.applicationRef,
      displayName: applicationCatalogLive
        ? selectedApplicationCatalogRecord?.displayName ?? ""
        : selectedApplication.displayName,
      applicationKind: applicationCatalogLive
        ? selectedApplicationCatalogRecord?.applicationKind ?? ""
        : selectedApplication.applicationKind,
      lifecycleState: applicationCatalogLive
        ? selectedApplicationCatalogRecord?.lifecycleState ?? ""
        : selectedApplication.lifecycleState ?? "active",
      recordVersion: applicationCatalogLive
        ? selectedApplicationCatalogRecord?.recordVersion ?? 0
        : selectedApplication.recordVersion ?? 0,
      updatedAt: applicationCatalogLive
        ? selectedApplicationCatalogRecord?.updatedAt ?? ""
        : selectedApplication.updatedAt,
      ownerSubjectRef: applicationCatalogLive
        ? selectedApplicationCatalogRecord?.ownerSubjectRef ?? ""
        : selectedApplication.ownerSubjectRef,
      workspaceId: applicationCatalogLive
        ? selectedApplicationCatalogRecord?.workspaceId ?? ""
        : selectedApplication.workspaceId ?? "",
      source: applicationCatalogLive ? "application_catalog" : "offline_read_model",
    }),
    [
      applicationCatalogLive,
      selectedApplication.applicationKind,
      selectedApplication.applicationRef,
      selectedApplication.displayName,
      selectedApplication.lifecycleState,
      selectedApplication.ownerSubjectRef,
      selectedApplication.recordVersion,
      selectedApplication.updatedAt,
      selectedApplication.workspaceId,
      selectedApplicationCatalogRecord,
    ],
  );
  const workflowScopedApplicationId = applicationDevelopmentWorkspaceContext.applicationId ||
    (applicationCatalogLive ? "" : activeWorkflowDraft.applicationRef);
  const savedDraftLibraryScopeKey =
    `${activeWorkspaceId}:${workflowScopedApplicationId}:${activeSavedDraftConsumerConfig.subjectRef}`;
  const savedDraftLibraryScopeKeyRef = useRef(savedDraftLibraryScopeKey);
  savedDraftLibraryScopeKeyRef.current = savedDraftLibraryScopeKey;
  const savedDraftConflictOpenSummary = useMemo(
    () =>
      activeSavedDraftListState.summaries.find(
        (summary) =>
          summary.draftId === activeWorkflowDraft.draftId &&
          summary.applicationRef === activeWorkflowDraft.applicationRef,
      ) ?? null,
    [activeSavedDraftListState.summaries, activeWorkflowDraft.applicationRef, activeWorkflowDraft.draftId],
  );
  const createdWorkspaceDraftCountsByDefinition = useMemo(
    () =>
      workspaceCreatedDrafts.reduce<Record<string, number>>((counts, draft) => {
        counts[draft.workflowDefinitionId] = (counts[draft.workflowDefinitionId] ?? 0) + 1;
        return counts;
      }, {}),
    [workspaceCreatedDrafts],
  );
  const workflowExecutorEligibility = useMemo(
    () => evaluateWorkflowExecutorEligibility(activeWorkflowDraft, savedDraftConsumerState, workflowDraftEditDirty),
    [activeWorkflowDraft, savedDraftConsumerState, workflowDraftEditDirty],
  );
  const workflowHTTPToolActionEligibility = useMemo(
    () => evaluateWorkflowHTTPToolActionEligibility(activeWorkflowDraft, savedDraftConsumerState, workflowDraftEditDirty),
    [activeWorkflowDraft, savedDraftConsumerState, workflowDraftEditDirty],
  );
  const activeWorkflowExecutorConditionValues = useMemo(
    () => Object.fromEntries(
      workflowExecutorEligibility.conditionNodeIds.map((nodeId) => [
        nodeId,
        workflowExecutorConditionValues[nodeId] ?? false,
      ]),
    ),
    [workflowExecutorConditionValues, workflowExecutorEligibility.conditionNodeIds],
  );

  useEffect(() => {
    setEditableWorkflowDraft(cloneWorkflowDraftForEditing(selectedWorkflowDraft));
    const pendingConsumerState = pendingSavedDraftConsumerStateRef.current;
    if (pendingConsumerState) {
      pendingSavedDraftConsumerStateRef.current = null;
      if (pendingConsumerState.draftId === selectedWorkflowDraft.draftId) {
        setSavedDraftConsumerState(pendingConsumerState.state);
        setWorkflowDraftEditDirty(false);
        return;
      }
    }
    if (selectedWorkflowDraft.localOnlyInteraction === "local_edit") {
      setWorkflowDraftEditDirty(true);
      setSavedDraftConsumerState(workspaceDraftCreatedConsumerState(activeSavedDraftConsumerConfig, selectedWorkflowDraft));
      return;
    }
    setSavedDraftConsumerState(initialWorkflowSavedDraftConsumerState(activeSavedDraftConsumerConfig));
    setWorkflowDraftEditDirty(false);
  }, [activeSavedDraftConsumerConfig, selectedWorkflowDraft.draftId]);

  useEffect(() => {
    setWorkflowExecutorState(initialWorkflowExecutorConsumerState(workflowExecutorConsumerConfig));
    setWorkflowExecutorConditionValues({});
  }, [selectedWorkflowDraft.draftId]);

  useEffect(() => {
    setWorkflowHTTPToolActionState(initialWorkflowHTTPToolActionConsumerState(workflowHTTPToolActionConsumerConfig));
    setWorkflowHTTPToolExecutionState(initialWorkflowHTTPToolExecutionState(workflowHTTPToolActionConsumerConfig));
    setWorkflowHTTPToolResourceKey("");
    setWorkflowHTTPToolLocale("");
    if (workflowHTTPToolActionConsumerConfig.mode !== "dev_workflow_http_tool_http") return;
    const reference = readWorkflowHTTPToolActionPlanReference();
    if (!reference || reference.workspaceId !== workflowHTTPToolActionConsumerConfig.workspaceId ||
      reference.applicationId !== selectedWorkflowDraft.applicationRef || reference.draftId !== selectedWorkflowDraft.draftId) return;

    let canceled = false;
    setWorkflowHTTPToolActionState((state) => ({
      ...state,
      status: "reading",
      summary: "Reloading the durable action plan selected before the page refresh.",
      failureCode: "",
    }));
    readWorkflowHTTPToolActionPlan(workflowHTTPToolActionConsumerConfig, reference).then((state) => {
      if (!canceled) setWorkflowHTTPToolActionState(state);
    });
    return () => {
      canceled = true;
    };
  }, [selectedWorkflowDraft.applicationRef, selectedWorkflowDraft.draftId]);

  const markWorkflowDraftLocallyEdited = () => {
    setWorkflowDraftEditDirty(true);
    setSavedDraftConsumerState((state) => {
      if (state.status === "version_conflict") {
        return {
          ...state,
          summary:
            "Local edits remain active, but the version conflict still requires explicit Continue local draft or Open saved draft before another dev route action.",
        };
      }
      if (state.status === "conflict_local_continued") {
        return {
          ...state,
          summary: `Local draft has unsaved edits after explicit conflict review; the next save will use saved version ${state.currentDraftVersion}.`,
        };
      }
      return {
        ...state,
        status: "unsaved_local",
        sourceLabel: "unsaved local",
        summary:
          state.mode === "dev_saved_draft_http"
            ? "Local draft has unsaved edits; validate or save through the dev-only saved draft route."
            : "Local draft has unsaved edits and remains in sample-only mode.",
        failureCode: null,
        conflictDraftVersion: null,
      };
    });
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
    handleWorkflowDraftNodePatch(nodeId, { label });
  };

  const handleWorkflowDraftNodeInputSummaryChange = (nodeId: string, inputSummary: string) => {
    handleWorkflowDraftNodePatch(nodeId, { inputSummary });
  };

  const handleWorkflowDraftNodeOutputSummaryChange = (nodeId: string, outputSummary: string) => {
    handleWorkflowDraftNodePatch(nodeId, { outputSummary });
  };

  const handleWorkflowDraftNodeProviderRefChange = (nodeId: string, providerRef: string) => {
    handleWorkflowDraftNodePatch(nodeId, { providerRef });
  };

  const handleWorkflowDraftNodeToolRefChange = (nodeId: string, toolRef: string) => {
    handleWorkflowDraftNodePatch(nodeId, { toolRef });
  };

  const handleWorkflowDraftNodeRagRefChange = (nodeId: string, ragRef: string) => {
    handleWorkflowDraftNodePatch(nodeId, { ragRef });
  };

  const handleWorkflowDraftNodeInputFieldsChange = (nodeId: string, inputFieldsText: string) => {
    handleWorkflowDraftNodePatch(nodeId, {
      inputContractFields: parseWorkflowDraftContractFields(inputFieldsText),
    });
  };

  const handleWorkflowDraftNodeOutputFieldsChange = (nodeId: string, outputFieldsText: string) => {
    handleWorkflowDraftNodePatch(nodeId, {
      outputContractFields: parseWorkflowDraftContractFields(outputFieldsText),
    });
  };

  const handleWorkflowDraftNodeOutputMappingChange = (nodeId: string, outputMappingSummary: string) => {
    handleWorkflowDraftNodePatch(nodeId, { outputMappingSummary });
  };

  const handleWorkflowDraftNodeDesignerPositionChange = (nodeId: string, x: number, y: number) => {
    setEditableWorkflowDraft((draft) => {
      const currentDraft = draft ?? cloneWorkflowDraftForEditing(selectedWorkflowDraft);
      if (!currentDraft.nodes.some((node) => node.nodeId === nodeId)) {
        return currentDraft;
      }
      return {
        ...currentDraft,
        localOnlyInteraction: "local_edit",
        designerLayout: workflowDraftLayoutWithNodePosition(currentDraft, nodeId, x, y),
      };
    });
    markWorkflowDraftLocallyEdited();
  };

  const handleWorkflowDraftNodePatch = (
    nodeId: string,
    patch: Partial<WorkflowDraftDesignerNode>,
  ) => {
    setEditableWorkflowDraft((draft) => {
      const currentDraft = draft ?? cloneWorkflowDraftForEditing(selectedWorkflowDraft);
      return {
        ...currentDraft,
        localOnlyInteraction: "local_edit",
        nodes: currentDraft.nodes.map((node) => (node.nodeId === nodeId ? { ...node, ...patch } : node)),
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
          edge.edgeId === edgeId
            ? {
                ...edge,
                conditionSummary: workflowDraftReviewableEdgeConditionSummary(
                  currentDraft,
                  edge,
                  conditionSummary,
                ),
              }
            : edge,
        ),
      };
    });
    markWorkflowDraftLocallyEdited();
  };

  const handleWorkflowDraftAddEdge = (fromNodeId: string, toNodeId: string): boolean => {
    if (!buildWorkflowDraftEdgeForConnection(activeWorkflowDraft, fromNodeId, toNodeId)) {
      return false;
    }
    setEditableWorkflowDraft((draft) => {
      const currentDraft = draft ?? cloneWorkflowDraftForEditing(selectedWorkflowDraft);
      const nextEdge = buildWorkflowDraftEdgeForConnection(currentDraft, fromNodeId, toNodeId);
      if (!nextEdge) {
        return currentDraft;
      }
      return {
        ...currentDraft,
        localOnlyInteraction: "local_edit",
        edges: [...currentDraft.edges, nextEdge],
      };
    });
    markWorkflowDraftLocallyEdited();
    return true;
  };

  const handleWorkflowDraftRemoveEdge = (edgeId: string): boolean => {
    if (!activeWorkflowDraft.edges.some((edge) => edge.edgeId === edgeId)) {
      return false;
    }
    setEditableWorkflowDraft((draft) => {
      const currentDraft = draft ?? cloneWorkflowDraftForEditing(selectedWorkflowDraft);
      if (!currentDraft.edges.some((edge) => edge.edgeId === edgeId)) {
        return currentDraft;
      }
      return {
        ...currentDraft,
        localOnlyInteraction: "local_edit",
        edges: currentDraft.edges.filter((edge) => edge.edgeId !== edgeId),
      };
    });
    markWorkflowDraftLocallyEdited();
    return true;
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
      setSavedDraftConsumerState(workspaceDraftCreatedConsumerState(activeSavedDraftConsumerConfig, selectedWorkflowDraft));
      return;
    }
    setWorkflowDraftEditDirty(false);
    setSavedDraftConsumerState(initialWorkflowSavedDraftConsumerState(activeSavedDraftConsumerConfig));
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
    if (workflowExecutorOperationPending) {
      return;
    }
    applyWorkflowSelectionPatch(
      selectionForApplication(applicationRef, {
        workspaceApplications,
        workspaceWorkflowDefinitions,
        workspaceRunHistory,
      }),
    );
  };
  const handleSelectApplicationCatalogRecord = (record: ApplicationCatalogRecord) => {
    if (workflowExecutorOperationPending) return;
    applyWorkflowSelectionPatch({
      applicationRef: record.applicationId,
      workflowDefinitionId: null,
      runId: null,
      draftId: null,
      scenarioId: null,
    });
  };
  const handleSelectWorkflowDefinition = (workflowDefinitionId: string) => {
    if (workflowExecutorOperationPending) {
      return;
    }
    applyWorkflowSelectionPatch(
      selectionForWorkflowDefinition(workflowDefinitionId, {
        workspaceWorkflowDefinitions,
        workspaceRunHistory,
      }),
    );
  };
  const handleSelectRun = (runId: string) => {
    if (workflowExecutorOperationPending) {
      return;
    }
    applyWorkflowSelectionPatch(selectionForRun(runId, { workspaceRunHistory }));
  };
  const handleOpenWorkspaceOperationsInboxItem = (item: WorkspaceOperationsInboxItem) => {
    if (item.sourceId === "applications" && item.applicationRef) {
      handleSelectApplication(item.applicationRef);
    } else if (item.sourceId === "workflow_definitions" && item.workflowDefinitionId) {
      handleSelectWorkflowDefinition(item.workflowDefinitionId);
    } else if (item.sourceId === "runs" && item.runId) {
      handleSelectRun(item.runId);
    }
  };
  const handleSelectWorkflowDraft = (draftId: string) => {
    if (workflowExecutorOperationPending) {
      return;
    }
    applyWorkflowSelectionPatch(selectionForDraft(draftId, workflowDraftDesigner, { workspaceRunHistory }));
  };
  const handleCreateWorkspaceDraftFromDefinition = (workflowDefinitionId: string) => {
    if (workflowExecutorOperationPending) {
      return;
    }
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
    setSavedDraftConsumerState(workspaceDraftCreatedConsumerState(activeSavedDraftConsumerConfig, createdDraft));
  };
  const handleCreateWorkflowExecutorDraft = () => {
    if (workflowExecutorOperationPending) {
      return;
    }
    const nextDraftNumber = workspaceCreatedDrafts.filter(
      (draft) =>
        draft.applicationRef === workflowScopedApplicationId &&
        draft.executionProfile === "executor_v0",
    ).length + 1;
    const createdDraft = buildWorkflowExecutorV0Draft(
      activeWorkflowDraft,
      nextDraftNumber,
      workflowScopedApplicationId,
    );
    setWorkspaceCreatedDrafts((drafts) => [...drafts, createdDraft]);
    applyWorkflowSelectionPatch({
      applicationRef: createdDraft.applicationRef,
      workflowDefinitionId: createdDraft.workflowDefinitionId,
      runId: null,
      draftId: createdDraft.draftId,
      scenarioId: null,
    });
    setEditableWorkflowDraft(cloneWorkflowDraftForEditing(createdDraft));
    setWorkflowDraftEditDirty(true);
    setSavedDraftConsumerState(workspaceDraftCreatedConsumerState(activeSavedDraftConsumerConfig, createdDraft));
    setWorkflowExecutorState(initialWorkflowExecutorConsumerState(workflowExecutorConsumerConfig));
    setWorkflowExecutorInput(DEFAULT_WORKFLOW_EXECUTOR_INPUT);
    setWorkflowExecutorModel("");
    setWorkflowExecutorConditionValues({});
  };
  const handleCreateWorkflowRAGDraft = (createdDraft: WorkflowDraftDesignerDraft) => {
    if (workflowRAGOperationPending) return;
    setWorkspaceCreatedDrafts((drafts) => [...drafts, createdDraft]);
    applyWorkflowSelectionPatch({ applicationRef: createdDraft.applicationRef, workflowDefinitionId: createdDraft.workflowDefinitionId, runId: null, draftId: createdDraft.draftId, scenarioId: null });
    setEditableWorkflowDraft(cloneWorkflowDraftForEditing(createdDraft));
    setWorkflowDraftEditDirty(true);
    setSavedDraftConsumerState(workspaceDraftCreatedConsumerState(activeSavedDraftConsumerConfig, createdDraft));
  };
  const handleCreateDefinitionDerivedDraft = (createdDraft: WorkflowDraftDesignerDraft) => {
    if (workflowExecutorOperationPending || workflowRAGOperationPending) return;
    setWorkspaceCreatedDrafts((drafts) => [...drafts, createdDraft]);
    applyWorkflowSelectionPatch({ applicationRef: createdDraft.applicationRef, workflowDefinitionId: createdDraft.workflowDefinitionId, runId: null, draftId: createdDraft.draftId, scenarioId: null });
    setEditableWorkflowDraft(cloneWorkflowDraftForEditing(createdDraft));
    setWorkflowDraftEditDirty(true);
    setSavedDraftConsumerState(workspaceDraftCreatedConsumerState(activeSavedDraftConsumerConfig, createdDraft));
  };
  const handleDeriveSavedWorkflowDraft = () => {
    const operationPending = workflowExecutorOperationPending || workflowRAGOperationPending;
    if (!canDeriveSavedWorkflowDraft(savedDraftConsumerState, workflowDraftEditDirty, operationPending)) {
      return;
    }
    const createdDraft = buildDerivedWorkflowDraft(
      activeWorkflowDraft,
      savedDraftConsumerState.currentDraftVersion,
      workflowDraftDesigner.drafts.map((draft) => draft.draftId),
    );
    const derivedConsumerState = workspaceDraftCreatedConsumerState(
      activeSavedDraftConsumerConfig,
      createdDraft,
      savedDraftConsumerState.currentLifecycleVersion,
    );
    pendingSavedDraftConsumerStateRef.current = {
      draftId: createdDraft.draftId,
      state: derivedConsumerState,
    };
    setWorkspaceCreatedDrafts((drafts) => [...drafts, createdDraft]);
    applyWorkflowSelectionPatch({
      applicationRef: createdDraft.applicationRef,
      workflowDefinitionId: createdDraft.workflowDefinitionId,
      runId: null,
      draftId: createdDraft.draftId,
      scenarioId: null,
    });
    setEditableWorkflowDraft(cloneWorkflowDraftForEditing(createdDraft));
    setWorkflowDraftEditDirty(true);
    setSavedDraftConsumerState(derivedConsumerState);
  };
  const refreshSavedWorkflowDraftList = (
    applicationRef: string,
    lifecycleState: WorkflowSavedDraftLifecycleState = savedDraftLibraryLifecycle,
    filters: WorkflowSavedDraftLibraryFilters = savedDraftLibraryFilters,
    append = false,
  ) => {
    const generation = savedDraftListRequestGenerationRef.current[lifecycleState] + 1;
    savedDraftListRequestGenerationRef.current[lifecycleState] = generation;
    const requestScopeKey = savedDraftLibraryScopeKeyRef.current;
    if (activeSavedDraftConsumerConfig.mode !== "dev_saved_draft_http" || !applicationRef) {
      setSavedDraftListStates((states) => ({
        ...states,
        [lifecycleState]: initialWorkflowSavedDraftListState(
          activeSavedDraftConsumerConfig,
          applicationRef,
          lifecycleState,
          filters,
        ),
      }));
      return;
    }
    const current = savedDraftListStates[lifecycleState];
    const cursor = append ? current.nextCursor : "";
    if (append && (!current.hasMore || current.status === "loading")) {
      return;
    }
    setSavedDraftListStates((states) => ({
      ...states,
      [lifecycleState]: {
        ...states[lifecycleState],
        status: "loading",
        mode: "dev_saved_draft_http",
        sourceLabel: cursor ? "loading more" : "loading",
        summary: cursor
          ? `Loading more ${lifecycleState} saved drafts.`
          : `Loading ${lifecycleState} saved drafts for the selected application.`,
        applicationRef,
        lifecycleState,
        filters,
        failureCode: null,
        ...(cursor ? {} : { summaries: [], nextCursor: "", hasMore: false }),
      },
    }));
    listWorkflowDraftDevRecords(applicationRef, activeSavedDraftConsumerConfig, {
      lifecycleState,
      filters,
      cursor,
      limit: 25,
    })
      .then((page) => {
        if (!workflowSavedDraftRequestIsCurrent(
          generation,
          savedDraftListRequestGenerationRef.current[lifecycleState],
          requestScopeKey,
          savedDraftLibraryScopeKeyRef.current,
        )) {
          return;
        }
        setSavedDraftListStates((states) => ({
          ...states,
          [lifecycleState]: cursor
            ? mergeWorkflowSavedDraftListPage(states[lifecycleState], page)
            : page,
        }));
      })
      .catch((error: unknown) => {
        if (!workflowSavedDraftRequestIsCurrent(
          generation,
          savedDraftListRequestGenerationRef.current[lifecycleState],
          requestScopeKey,
          savedDraftLibraryScopeKeyRef.current,
        )) {
          return;
        }
        setSavedDraftListStates((states) => ({
          ...states,
          [lifecycleState]: {
            ...states[lifecycleState],
            status: "list_failed",
            sourceLabel: "list_failed",
            summary: error instanceof Error ? error.message : "Saved draft list failed.",
            applicationRef,
            lifecycleState,
            filters,
            failureCode: "dev_saved_draft_list_failed",
            ...(cursor ? {} : { summaries: [], nextCursor: "", hasMore: false }),
          },
        }));
      });
  };
  useEffect(() => {
    savedDraftListRequestGenerationRef.current.active += 1;
    savedDraftListRequestGenerationRef.current.archived += 1;
    savedDraftLifecycleOperationGenerationRef.current += 1;
    savedDraftOpenRequestGenerationRef.current += 1;
    setSavedDraftLibraryLifecycle("active");
    setSavedDraftLibraryFilters(emptyWorkflowSavedDraftLibraryFilters());
    setSavedDraftLifecycleOperation(initialWorkflowSavedDraftLifecycleOperationState());
    setSavedDraftListStates({
      active: initialWorkflowSavedDraftListState(
        activeSavedDraftConsumerConfig,
        workflowScopedApplicationId,
        "active",
      ),
      archived: initialWorkflowSavedDraftListState(
        activeSavedDraftConsumerConfig,
        workflowScopedApplicationId,
        "archived",
      ),
    });
    refreshSavedWorkflowDraftList(
      workflowScopedApplicationId,
      "active",
      emptyWorkflowSavedDraftLibraryFilters(),
    );
  }, [
    activeSavedDraftConsumerConfig,
    applicationDevelopmentWorkspaceContext.generationKey,
    workflowScopedApplicationId,
  ]);
  const handleRefreshSavedWorkflowDraftList = () => {
    refreshSavedWorkflowDraftList(
      workflowScopedApplicationId,
      savedDraftLibraryLifecycle,
      savedDraftLibraryFilters,
    );
  };
  const handleSavedDraftLibraryLifecycleChange = (lifecycleState: WorkflowSavedDraftLifecycleState) => {
    savedDraftListRequestGenerationRef.current[lifecycleState] += 1;
    savedDraftLifecycleOperationGenerationRef.current += 1;
    savedDraftOpenRequestGenerationRef.current += 1;
    setSavedDraftLibraryLifecycle(lifecycleState);
    setSavedDraftLifecycleOperation(initialWorkflowSavedDraftLifecycleOperationState());
    setSavedDraftListStates((states) => ({
      ...states,
      [lifecycleState]: initialWorkflowSavedDraftListState(
        activeSavedDraftConsumerConfig,
        workflowScopedApplicationId,
        lifecycleState,
        savedDraftLibraryFilters,
      ),
    }));
    refreshSavedWorkflowDraftList(
      workflowScopedApplicationId,
      lifecycleState,
      savedDraftLibraryFilters,
    );
  };
  const handleSavedDraftLibraryFiltersChange = (filters: WorkflowSavedDraftLibraryFilters) => {
    savedDraftListRequestGenerationRef.current.active += 1;
    savedDraftListRequestGenerationRef.current.archived += 1;
    savedDraftLifecycleOperationGenerationRef.current += 1;
    savedDraftOpenRequestGenerationRef.current += 1;
    setSavedDraftLibraryFilters(filters);
    setSavedDraftLifecycleOperation(initialWorkflowSavedDraftLifecycleOperationState());
    setSavedDraftListStates({
      active: initialWorkflowSavedDraftListState(
        activeSavedDraftConsumerConfig,
        workflowScopedApplicationId,
        "active",
        filters,
      ),
      archived: initialWorkflowSavedDraftListState(
        activeSavedDraftConsumerConfig,
        workflowScopedApplicationId,
        "archived",
        filters,
      ),
    });
    refreshSavedWorkflowDraftList(workflowScopedApplicationId, savedDraftLibraryLifecycle, filters);
  };
  const handleLoadMoreSavedWorkflowDrafts = () => {
    refreshSavedWorkflowDraftList(
      workflowScopedApplicationId,
      savedDraftLibraryLifecycle,
      savedDraftLibraryFilters,
      true,
    );
  };
  const handleOpenSavedWorkflowDraft = (summary: WorkflowSavedDraftSummary) => {
    if (activeSavedDraftConsumerConfig.mode !== "dev_saved_draft_http") {
      return;
    }
    const requestGeneration = savedDraftOpenRequestGenerationRef.current + 1;
    savedDraftOpenRequestGenerationRef.current = requestGeneration;
    const requestScopeKey = savedDraftLibraryScopeKeyRef.current;
    setSavedDraftConsumerState((state) => ({
      ...state,
      status: "reading",
      summary: summary.lifecycleState === "archived"
        ? `Opening archived saved draft ${summary.draftId} for read-only review.`
        : `Opening saved draft ${summary.draftId} through the dev-only read route.`,
      failureCode: null,
      currentDraftVersion: summary.draftVersion,
      currentLifecycleVersion: summary.lifecycleVersion,
      currentLifecycleState: summary.lifecycleState,
      conflictDraftVersion: null,
    }));
    openWorkflowDraftDevRecord(summary, activeSavedDraftConsumerConfig)
      .then((result) => {
        if (!workflowSavedDraftRequestIsCurrent(
          requestGeneration,
          savedDraftOpenRequestGenerationRef.current,
          requestScopeKey,
          savedDraftLibraryScopeKeyRef.current,
        )) {
          return;
        }
        setSavedDraftConsumerState(result.state);
        if (!result.draft) {
          setSavedDraftListStates((states) => ({
            ...states,
            [summary.lifecycleState]: {
              ...states[summary.lifecycleState],
              status: "open_failed",
              sourceLabel: "open_failed",
              summary: result.state.summary,
              failureCode: result.state.failureCode ?? "dev_saved_draft_open_failed",
            },
          }));
          return;
        }
        const openedDraft = result.draft;
        pendingSavedDraftConsumerStateRef.current = {
          draftId: openedDraft.draftId,
          state: result.state,
        };
        const nextRun = workspaceRunHistory.runs.find(
          (run) =>
            run.applicationRef === openedDraft.applicationRef &&
            run.workflowDefinitionId === openedDraft.workflowDefinitionId,
        );
        setWorkspaceCreatedDrafts((drafts) => [
          ...drafts.filter((draft) => draft.draftId !== openedDraft.draftId),
          openedDraft,
        ]);
        applyWorkflowSelectionPatch({
          applicationRef: openedDraft.applicationRef,
          workflowDefinitionId: openedDraft.workflowDefinitionId,
          runId: nextRun?.runId ?? null,
          draftId: openedDraft.draftId,
          scenarioId: null,
        });
        setEditableWorkflowDraft(cloneWorkflowDraftForEditing(openedDraft));
        setWorkflowDraftEditDirty(false);
        window.location.hash = "#workflow-draft-designer";
      })
      .catch((error: unknown) => {
        if (!workflowSavedDraftRequestIsCurrent(
          requestGeneration,
          savedDraftOpenRequestGenerationRef.current,
          requestScopeKey,
          savedDraftLibraryScopeKeyRef.current,
        )) {
          return;
        }
        const message = error instanceof Error ? error.message : "Saved draft open failed.";
        setSavedDraftConsumerState((state) => ({
          ...state,
          status: "read_failed",
          sourceLabel: "open_failed",
          summary: message,
          failureCode: "dev_saved_draft_open_failed",
          conflictDraftVersion: null,
        }));
        setSavedDraftListStates((states) => ({
          ...states,
          [summary.lifecycleState]: {
            ...states[summary.lifecycleState],
            status: "open_failed",
            sourceLabel: "open_failed",
            summary: message,
            failureCode: "dev_saved_draft_open_failed",
          },
        }));
      });
  };
  const handleSavedWorkflowDraftLifecycleTransition = async (
    summary: WorkflowSavedDraftSummary,
    targetState: WorkflowSavedDraftLifecycleState,
  ) => {
    if (activeSavedDraftConsumerConfig.mode !== "dev_saved_draft_http") {
      return;
    }
    const operationGeneration = savedDraftLifecycleOperationGenerationRef.current + 1;
    savedDraftLifecycleOperationGenerationRef.current = operationGeneration;
    const operationScopeKey = savedDraftLibraryScopeKeyRef.current;
    if (
      targetState === "archived" &&
      activeWorkflowDraft.draftId === summary.draftId &&
      workflowDraftEditDirty
    ) {
      setSavedDraftLifecycleOperation({
        status: "failed",
        draftId: summary.draftId,
        targetState,
        currentDraftVersion: summary.draftVersion,
        currentLifecycleVersion: summary.lifecycleVersion,
        currentLifecycleState: summary.lifecycleState,
        failureCode: "draft_local_edits_pending",
        requestId: `saved-draft-${targetState}-${summary.draftId}`,
        auditRef: "not_sent",
        summary: "Save or reset local edits before archiving this exact saved draft version.",
      });
      return;
    }
    setSavedDraftLifecycleOperation({
      status: "transitioning",
      draftId: summary.draftId,
      targetState,
      currentDraftVersion: summary.draftVersion,
      currentLifecycleVersion: summary.lifecycleVersion,
      currentLifecycleState: summary.lifecycleState,
      failureCode: null,
      requestId: `saved-draft-${targetState}-${summary.draftId}`,
      auditRef: "pending",
      summary: `${targetState === "archived" ? "Archiving" : "Unarchiving"} ${summary.draftId}.`,
    });
    try {
      const result = targetState === "archived"
        ? await archiveWorkflowDraftDevRecord(summary, activeSavedDraftConsumerConfig)
        : await unarchiveWorkflowDraftDevRecord(summary, activeSavedDraftConsumerConfig);
      if (!workflowSavedDraftRequestIsCurrent(
        operationGeneration,
        savedDraftLifecycleOperationGenerationRef.current,
        operationScopeKey,
        savedDraftLibraryScopeKeyRef.current,
      )) {
        return;
      }
      setSavedDraftLifecycleOperation(result);
      if (result.status === "failed") {
        return;
      }
      if (activeWorkflowDraft.draftId === summary.draftId) {
        setSavedDraftConsumerState((state) => ({
          ...state,
          status: "saved_dev_record",
          sourceLabel: targetState === "active" ? "reopen required" : "archived read-only",
          currentDraftVersion: result.currentDraftVersion,
          currentLifecycleVersion: result.currentLifecycleVersion,
          currentLifecycleState: targetState === "active" ? "unknown" : result.currentLifecycleState,
          summary: targetState === "active"
            ? `${result.summary} The existing browser draft remains read-only until it is opened again.`
            : result.summary,
          failureCode: null,
          requestId: result.requestId,
          auditRef: result.auditRef,
        }));
        setWorkflowDraftEditDirty(false);
      }
      refreshSavedWorkflowDraftList(workflowScopedApplicationId, "active", savedDraftLibraryFilters);
      refreshSavedWorkflowDraftList(workflowScopedApplicationId, "archived", savedDraftLibraryFilters);
    } catch (error) {
      if (!workflowSavedDraftRequestIsCurrent(
        operationGeneration,
        savedDraftLifecycleOperationGenerationRef.current,
        operationScopeKey,
        savedDraftLibraryScopeKeyRef.current,
      )) {
        return;
      }
      setSavedDraftLifecycleOperation({
        status: "failed",
        draftId: summary.draftId,
        targetState,
        currentDraftVersion: summary.draftVersion,
        currentLifecycleVersion: summary.lifecycleVersion,
        currentLifecycleState: summary.lifecycleState,
        failureCode: "dev_saved_draft_lifecycle_request_failed",
        requestId: `saved-draft-${targetState}-${summary.draftId}`,
        auditRef: "unavailable",
        summary: error instanceof Error ? error.message : "Saved draft lifecycle request failed.",
      });
    }
  };
  const handleContinueLocalWorkflowDraftAfterConflict = () => {
    setSavedDraftConsumerState((state) =>
      continueLocalWorkflowDraftAfterVersionConflict(state, activeWorkflowDraft),
    );
    setWorkflowDraftEditDirty(true);
  };
  const handleOpenConflictSavedWorkflowDraft = () => {
    if (!savedDraftConflictOpenSummary) {
      return;
    }
    handleOpenSavedWorkflowDraft(savedDraftConflictOpenSummary);
  };
  const handleValidateWorkflowDraft = () => {
    if (
      activeSavedDraftConsumerConfig.mode !== "dev_saved_draft_http" ||
      (savedDraftConsumerState.currentDraftVersion > 0 &&
        savedDraftConsumerState.currentLifecycleState !== "active") ||
      workflowSavedDraftConflictRequiresResolution(savedDraftConsumerState)
    ) {
      return;
    }
    const currentDraftVersion = savedDraftConsumerState.currentDraftVersion;
    setSavedDraftConsumerState((state) => ({
      ...state,
      status: "validating",
      summary: "Validating local draft through the dev-only saved draft route.",
      failureCode: null,
      conflictDraftVersion: null,
    }));
    validateWorkflowDraftDevRecord(
      activeWorkflowDraft,
      activeSavedDraftConsumerConfig,
      currentDraftVersion,
      savedDraftConsumerState.currentLifecycleVersion,
      savedDraftConsumerState.currentLifecycleState,
    )
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
    if (
      activeSavedDraftConsumerConfig.mode !== "dev_saved_draft_http" ||
      (savedDraftConsumerState.currentDraftVersion > 0 &&
        savedDraftConsumerState.currentLifecycleState !== "active")
    ) {
      return;
    }
    const expectedDraftVersion = nextWorkflowSavedDraftExpectedVersion(savedDraftConsumerState);
    if (expectedDraftVersion === null) {
      return;
    }
    setSavedDraftConsumerState((state) => ({
      ...state,
      status: "saving",
      summary: "Saving local draft through the dev-only saved draft route.",
      failureCode: null,
      conflictDraftVersion: null,
    }));
    saveWorkflowDraftDevRecord(
      activeWorkflowDraft,
      activeSavedDraftConsumerConfig,
      expectedDraftVersion,
      savedDraftConsumerState.currentLifecycleVersion,
    )
      .then((nextState) => {
        setSavedDraftConsumerState(nextState);
        if (nextState.status === "version_conflict") {
          refreshSavedWorkflowDraftList(
            activeWorkflowDraft.applicationRef,
            "active",
            savedDraftLibraryFilters,
          );
          return;
        }
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
          refreshSavedWorkflowDraftList(
            activeWorkflowDraft.applicationRef,
            "active",
            savedDraftLibraryFilters,
          );
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
    if (
      activeSavedDraftConsumerConfig.mode !== "dev_saved_draft_http" ||
      workflowSavedDraftConflictRequiresResolution(savedDraftConsumerState)
    ) {
      return;
    }
    const currentDraftVersion = savedDraftConsumerState.currentDraftVersion;
    setSavedDraftConsumerState((state) => ({
      ...state,
      status: "reading",
      summary: "Reading local draft through the dev-only saved draft route.",
      failureCode: null,
      conflictDraftVersion: null,
    }));
    readWorkflowDraftDevRecord(activeWorkflowDraft, activeSavedDraftConsumerConfig, currentDraftVersion)
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
  const handleWorkflowDraftRevisionRestored = (
    restoredDraft: WorkflowDraftDesignerDraft,
    result: WorkflowSavedDraftRevisionRestoreResult,
  ) => {
    setWorkspaceCreatedDrafts((drafts) => [
      ...drafts.filter((draft) => draft.draftId !== restoredDraft.draftId),
      restoredDraft,
    ]);
    setEditableWorkflowDraft(cloneWorkflowDraftForEditing(restoredDraft));
    setWorkflowDraftEditDirty(false);
    setSavedDraftConsumerState({
      status: result.failureCode ? "save_failed" : "saved_dev_record",
      mode: "dev_saved_draft_http",
      sourceLabel: result.failureCode ?? "restored revision",
      summary: result.summary,
      failureCode: result.failureCode,
      currentDraftVersion: result.currentDraftVersion,
      currentLifecycleVersion: result.currentLifecycleVersion,
      currentLifecycleState: result.currentLifecycleState,
      conflictDraftVersion: null,
      auditRef: result.auditRef,
      requestId: result.requestId,
    });
    refreshSavedWorkflowDraftList(restoredDraft.applicationRef, "active", savedDraftLibraryFilters);
  };
  const handleWorkflowExecutorConditionValueChange = (nodeId: string, value: boolean) => {
    setWorkflowExecutorConditionValues((values) => ({ ...values, [nodeId]: value }));
  };
  const handleStartWorkflowRun = () => {
    if (
      workflowExecutorConsumerConfig.mode !== "dev_workflow_executor_http" ||
      !workflowExecutorEligibility.eligible ||
      workflowExecutorOperationPending ||
      workflowExecutorInput.trim().length === 0
    ) {
      return;
    }
    setWorkflowExecutorState((state) => ({
      ...state,
      status: "starting",
      summary: `Starting bounded run for saved draft ${activeWorkflowDraft.draftId} version ${workflowExecutorEligibility.savedDraftVersion}.`,
      failureCode: null,
      failureSummary: "",
      record: null,
    }));
    startWorkflowRunDevRecord(
      activeWorkflowDraft,
      workflowExecutorInput,
      activeWorkflowExecutorConditionValues,
      workflowExecutorConsumerConfig,
      { model: workflowExecutorModel },
    )
      .then(setWorkflowExecutorState)
      .catch((error: unknown) => {
        setWorkflowExecutorState((state) => ({
          ...state,
          status: "failed",
          summary: error instanceof Error ? error.message : "Workflow executor request failed.",
          failureCode: "dev_workflow_executor_consumer_failed",
          failureSummary: "The development executor route could not return a valid run envelope.",
          record: null,
        }));
      });
  };
  const handleReloadWorkflowRun = () => {
    const record = workflowExecutorState.record;
    if (!record || workflowExecutorOperationPending) {
      return;
    }
    setWorkflowExecutorState((state) => ({
      ...state,
      status: "reading",
      summary: `Reloading scoped run record ${record.runId}.`,
      failureCode: null,
      failureSummary: "",
    }));
    readWorkflowRunDevRecord(record, workflowExecutorConsumerConfig)
      .then(setWorkflowExecutorState)
      .catch((error: unknown) => {
        setWorkflowExecutorState((state) => ({
          ...state,
          status: "failed",
          summary: error instanceof Error ? error.message : "Workflow run record reload failed.",
          failureCode: "dev_workflow_run_read_failed",
          failureSummary: "The scoped development run record could not be reloaded.",
        }));
      });
  };
  const handleCreateWorkflowHTTPToolActionPlan = (publicArguments: WorkflowHTTPToolPublicArguments) => {
    if (workflowHTTPToolActionConsumerConfig.mode !== "dev_workflow_http_tool_http" ||
      !workflowHTTPToolActionEligibility.eligible || workflowHTTPToolActionOperationPending) return;
    setWorkflowHTTPToolActionState((state) => ({
      ...state,
      status: "creating",
      summary: `Creating an immutable plan from saved draft ${activeWorkflowDraft.draftId} version ${workflowHTTPToolActionEligibility.draftVersion}.`,
      failureCode: "",
      confirmationDecision: null,
    }));
    createWorkflowHTTPToolActionPlan(workflowHTTPToolActionConsumerConfig, {
      draftId: activeWorkflowDraft.draftId,
      applicationId: activeWorkflowDraft.applicationRef,
      draftVersion: workflowHTTPToolActionEligibility.draftVersion,
      nodeId: workflowHTTPToolActionEligibility.nodeId,
      publicArguments,
    }).then((state) => {
      if (state.actionPlan) rememberWorkflowHTTPToolActionPlanReference(state.actionPlan);
      setWorkflowHTTPToolActionState(state);
    });
  };
  const handleReloadWorkflowHTTPToolActionPlan = () => {
    const plan = workflowHTTPToolActionState.actionPlan;
    if (!plan || workflowHTTPToolActionOperationPending) return;
    setWorkflowHTTPToolActionState((state) => ({
      ...state,
      status: "reading",
      summary: `Reloading durable action plan ${plan.planId}.`,
      failureCode: "",
      confirmationDecision: null,
    }));
    readWorkflowHTTPToolActionPlan(workflowHTTPToolActionConsumerConfig, plan).then(setWorkflowHTTPToolActionState);
  };
  const handleWorkflowHTTPToolActionDecision = (decision: WorkflowHTTPToolHumanDecision) => {
    const plan = workflowHTTPToolActionState.actionPlan;
    if (!plan || workflowHTTPToolActionOperationPending) return;
    setWorkflowHTTPToolActionState((state) => ({
      ...state,
      status: "deciding",
      summary: `Recording ${decision} against durable plan version ${plan.recordVersion}.`,
      failureCode: "",
      confirmationDecision: null,
    }));
    decideWorkflowHTTPToolActionPlan(workflowHTTPToolActionConsumerConfig, plan, decision).then((state) => {
      if (state.actionPlan) rememberWorkflowHTTPToolActionPlanReference(state.actionPlan);
      setWorkflowHTTPToolActionState(state);
    });
  };

  const handleRunApprovedHTTPToolActionPlan = () => {
    const plan = workflowHTTPToolActionState.actionPlan;
    if (!plan || plan.status !== "approved" || workflowHTTPToolOperationPending || !workflowHTTPToolPermissions.execute.available) return;
    setWorkflowHTTPToolExecutionState((state) => ({
      ...state,
      status: "executing",
      summary: `Claiming approved plan ${plan.planId} version ${plan.recordVersion} for its single execution attempt.`,
      failureCode: "",
      actionPlan: plan,
      run: null,
    }));
    executeWorkflowHTTPToolActionPlan(workflowHTTPToolActionConsumerConfig, plan, {
      inputText: workflowHTTPToolExecutionInput,
      model: workflowHTTPToolExecutionModel,
    }).then((state) => {
      if (state.actionPlan) {
        rememberWorkflowHTTPToolActionPlanReference(state.actionPlan);
        setWorkflowHTTPToolActionState((current) => ({
          ...current,
          status: state.actionPlan?.status === "consumed" ? "ready" : current.status,
          summary: state.actionPlan?.status === "consumed"
            ? `Durable action plan ${state.actionPlan.planId} was consumed by its single execution attempt.`
            : current.summary,
          failureCode: state.failureCode,
          requestId: state.requestId,
          auditRef: state.auditRef,
          actionPlan: state.actionPlan,
          confirmationDecision: null,
        }));
      }
      setWorkflowHTTPToolExecutionState(state);
    });
  };

  const handleActiveWorkspaceSwitch = (candidate: string): boolean => {
    const normalized = normalizeActiveWorkspaceId(candidate);
    if (!normalized) {
      return false;
    }
    if (normalized === activeWorkspaceId) {
      return true;
    }
    setSelectedApplicationRef(null);
    setApplicationCatalogSnapshot(null);
    setSelectedWorkflowDefinitionId(null);
    setSelectedRunId(null);
    setSelectedWorkflowDraftId(null);
    setSelectedWorkflowScenarioId(null);
    setEditableWorkflowDraft(null);
    setWorkflowDraftEditDirty(false);
    pendingSavedDraftConsumerStateRef.current = null;
    savedDraftListRequestGenerationRef.current.active += 1;
    savedDraftListRequestGenerationRef.current.archived += 1;
    savedDraftLifecycleOperationGenerationRef.current += 1;
    savedDraftOpenRequestGenerationRef.current += 1;
    setActiveWorkspaceId(normalized);
    return true;
  };

  return (
    <main className="product-shell" data-rd-profile="workbench">
      <ProductNavigation
        activeWorkspaceId={activeWorkspaceId}
        apiKeysAnchor="#workspace-api-keys"
        counts={{
          inbox: workspaceOperationsInbox.items.length,
          applications: workspaceApplications.applications.length,
          workflows: workspaceWorkflowDefinitions.workflowDefinitions.length,
          apiKeys: workspaceApiKeys.apiKeys.length,
        }}
        sourceConfig={activeDevLiveConfig}
        sourceState={devLiveState}
        onActiveWorkspaceSwitch={handleActiveWorkspaceSwitch}
      />

      <section className="product-workspace" aria-label="Control plane read shell">
        <WorkspaceProductOverviewPanel
          application={applicationDevelopmentWorkspaceContext}
          inbox={workspaceOperationsInbox}
          sourceConfig={activeDevLiveConfig}
          sourceState={devLiveState}
        />
        <WorkspaceOperationsInboxPanel
          inbox={workspaceOperationsInbox}
          onOpenItem={handleOpenWorkspaceOperationsInboxItem}
        />
        <Suspense fallback={<section className="surface-band"><p>Loading Saved Draft Library…</p></section>}>
          <WorkflowUserWorkspaceHomePanel
            home={workflowUserWorkspaceHome}
            createdDraftCountsByWorkflowDefinition={createdWorkspaceDraftCountsByDefinition}
            savedDraftListState={savedDraftListState}
            libraryLifecycle={savedDraftLibraryLifecycle}
            libraryFilters={savedDraftLibraryFilters}
            lifecycleOperation={savedDraftLifecycleOperation}
            onCreateDraftForWorkflowDefinition={handleCreateWorkspaceDraftFromDefinition}
            onLibraryLifecycleChange={handleSavedDraftLibraryLifecycleChange}
            onLibraryFiltersChange={handleSavedDraftLibraryFiltersChange}
            onRefreshSavedDrafts={handleRefreshSavedWorkflowDraftList}
            onLoadMoreSavedDrafts={handleLoadMoreSavedWorkflowDrafts}
            onOpenSavedDraft={handleOpenSavedWorkflowDraft}
            onArchiveSavedDraft={(summary) => {
              void handleSavedWorkflowDraftLifecycleTransition(summary, "archived");
            }}
            onUnarchiveSavedDraft={(summary) => {
              void handleSavedWorkflowDraftLifecycleTransition(summary, "active");
            }}
          />
        </Suspense>
        <ModelGatewayOverviewPanel overview={modelGatewayOverview} />
        <ModelGatewayRouteEvidencePanel detail={modelGatewayRouteEvidence} />
        <ModelGatewayUsageAuditEvidencePanel evidence={modelGatewayUsageAuditEvidence} />
        <Suspense fallback={<section className="surface-band"><p>Loading review evidence…</p></section>}>
          <ModelGatewayEvidenceReviewPanel review={modelGatewayEvidenceReview} />
        </Suspense>
        <Suspense fallback={<section className="surface-band"><p>Loading Admin Control Plane…</p></section>}>
          <AdminControlPlaneWorkspace
            tenantOverview={tenantOverview}
            auditLog={adminAuditLog}
            operationsReview={adminOperationsReview}
            providerDeploymentReview={adminProviderDeploymentReview}
            sourceConfig={activeDevLiveConfig}
            sourceState={devLiveState}
          />
        </Suspense>
        <WorkflowWorkspaceReviewPanel review={workflowWorkspaceReview} />
        <WorkflowSurfaceOverviewPanel overview={workflowSurfaceOverview} />
        <WorkflowScenarioInspectorPanel
          inspector={workflowScenarioInspector}
          selectedScenarioId={workflowScenarioInspector.selectedScenarioId}
          onSelectScenario={setSelectedWorkflowScenarioId}
        />

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

          <Suspense fallback={<div className="application-catalog-panel"><p>Loading application catalog management…</p></div>}>
            <ApplicationCatalogPanel
              selectedApplicationId={selectedApplicationRef}
              onSelectRecord={handleSelectApplicationCatalogRecord}
              onSnapshotChange={setApplicationCatalogSnapshot}
            />
          </Suspense>

          {!applicationCatalogLive ? (
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
          ) : null}

          {applicationDevelopmentWorkspaceContext.status !== "unavailable" ? (
            <WorkflowApplicationDetailPanel detail={workflowApplicationDetail} />
          ) : null}

          <Suspense fallback={<div className="application-development-workspace"><p>Loading Application Development Workspace…</p></div>}>
            <ApplicationDevelopmentWorkspacePanel
              key={applicationDevelopmentWorkspaceContext.generationKey}
              context={applicationDevelopmentWorkspaceContext}
              renderStageSurfaces={(activeStage, surfaceKey, controls) => (
                <Suspense fallback={<div className="application-development-stage-surfaces"><p>Loading Application Development stage surfaces…</p></div>}>
                  <ApplicationDevelopmentWorkspaceSurface
                    key={surfaceKey}
                    context={applicationDevelopmentWorkspaceContext}
                    activeStage={activeStage}
                    surfaceKey={surfaceKey}
                    controls={controls}
                    offlineApiKeys={workspaceApiKeys}
                    suggestedDefinitionId={selectedWorkflowDefinitionId ?? ""}
                    activeWorkflowDraft={activeWorkflowDraft}
                    savedDraftVersion={savedDraftConsumerState.currentDraftVersion ?? 0}
                    savedDraftLifecycleVersion={savedDraftConsumerState.currentLifecycleVersion}
                    savedDraftLifecycleState={savedDraftConsumerState.currentLifecycleState}
                    nextDerivedDraftNumber={workspaceCreatedDrafts.filter(
                      (draft) => draft.applicationRef === workflowScopedApplicationId && (draft.baseDefinitionVersion ?? 0) > 0,
                    ).length + 1}
                    onDerivedDraft={handleCreateDefinitionDerivedDraft}
                    onRunRecorded={() => setWorkflowRunHistoryRefreshKey((key) => key + 1)}
                  />
                </Suspense>
              )}
              renderPersistentSurfaces={(surfaceKey, controls) => (
                <>
                  <Suspense fallback={<div className="application-runtime-review-loading"><p>Loading Application Runtime Review…</p></div>}>
                    <ApplicationRuntimeReviewWorkspace
                      context={applicationDevelopmentWorkspaceContext}
                      surfaceKey={surfaceKey}
                      controls={controls}
                    />
                  </Suspense>
                  <Suspense fallback={<div className="workflow-review-loading"><p>Loading Workflow Review…</p></div>}>
                    <WorkflowReviewWorkspace
                      context={applicationDevelopmentWorkspaceContext}
                      surfaceKey={surfaceKey}
                      controls={controls}
                      refreshKey={workflowRunHistoryRefreshKey}
                    />
                  </Suspense>
                </>
              )}
            />
          </Suspense>

          <div className="application-states" aria-label="Workspace application states">
            {workspaceApplications.statePreviews.map((state) => (
              <ApplicationStatePreview key={state.id} state={state} />
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
              <h3 id="workspace-usage-quota-title">Usage Quota</h3>
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
              <h3 id="workspace-workflow-definitions-title">Workflows</h3>
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
          <Suspense fallback={<section className="workflow-draft-designer"><p>正在加载 Workflow Designer Workbench…</p></section>}>
            <WorkflowDraftDesignerPanel
              designer={workflowDraftDesigner}
              selectedDraft={activeWorkflowDraft}
              validationInspector={activeWorkflowDraftValidationInspector}
              selectedDraftId={selectedWorkflowDraft.draftId}
              savedDraftConsumerState={savedDraftConsumerState}
              savedDraftConflictReviewSummary={savedDraftConflictReviewSummary}
              savedDraftConflictOpenSummary={savedDraftConflictOpenSummary}
              draftEditDirty={workflowDraftEditDirty}
              executorOperationPending={workflowExecutorOperationPending || workflowHTTPToolOperationPending || workflowRAGOperationPending}
              nodeTypeOptions={WORKFLOW_DRAFT_NODE_TYPE_OPTIONS}
              canRemoveNode={(nodeId) => canRemoveWorkflowDraftNode(activeWorkflowDraft, nodeId)}
              onSelectDraft={handleSelectWorkflowDraft}
              onUpdateDraftLabel={handleWorkflowDraftLabelChange}
              onUpdateDraftSummary={handleWorkflowDraftSummaryChange}
              onUpdateNodeLabel={handleWorkflowDraftNodeLabelChange}
              onUpdateNodeInputSummary={handleWorkflowDraftNodeInputSummaryChange}
              onUpdateNodeOutputSummary={handleWorkflowDraftNodeOutputSummaryChange}
              onUpdateNodeProviderRef={handleWorkflowDraftNodeProviderRefChange}
              onUpdateNodeToolRef={handleWorkflowDraftNodeToolRefChange}
              onUpdateNodeRagRef={handleWorkflowDraftNodeRagRefChange}
              onUpdateNodeInputFields={handleWorkflowDraftNodeInputFieldsChange}
              onUpdateNodeOutputFields={handleWorkflowDraftNodeOutputFieldsChange}
              onUpdateNodeOutputMapping={handleWorkflowDraftNodeOutputMappingChange}
              onUpdateNodeDesignerPosition={handleWorkflowDraftNodeDesignerPositionChange}
              onUpdateEdgeCondition={handleWorkflowDraftEdgeConditionChange}
              onAddEdge={handleWorkflowDraftAddEdge}
              onRemoveEdge={handleWorkflowDraftRemoveEdge}
              onAddNode={handleWorkflowDraftAddNode}
              onMoveNode={handleWorkflowDraftMoveNode}
              onRemoveNode={handleWorkflowDraftRemoveNode}
              onResetDraftEdits={handleWorkflowDraftEditReset}
              onContinueLocalDraftAfterConflict={handleContinueLocalWorkflowDraftAfterConflict}
              onOpenConflictSavedDraft={handleOpenConflictSavedWorkflowDraft}
              onDeriveSavedDraft={handleDeriveSavedWorkflowDraft}
              onValidateDraft={handleValidateWorkflowDraft}
              onSaveDraft={handleSaveWorkflowDraft}
              onReadDraft={handleReadWorkflowDraft}
            />
          </Suspense>
          <Suspense fallback={<section className="workflow-draft-revision-panel"><p>正在加载草案修订历史工作区…</p></section>}>
            <WorkflowSavedDraftRevisionPanel
              draft={activeWorkflowDraft}
              currentDraftVersion={savedDraftConsumerState.currentDraftVersion}
              currentLifecycleVersion={savedDraftConsumerState.currentLifecycleVersion}
              lifecycleState={savedDraftConsumerState.currentLifecycleState}
              config={activeSavedDraftConsumerConfig}
              dirty={workflowDraftEditDirty}
              disabled={
                workflowExecutorOperationPending ||
                workflowHTTPToolOperationPending ||
                workflowRAGOperationPending ||
                ["saving", "validating", "reading"].includes(savedDraftConsumerState.status)
              }
              onRestored={handleWorkflowDraftRevisionRestored}
            />
          </Suspense>
          <WorkflowHTTPToolActionPanel
            draft={activeWorkflowDraft}
            consumerState={workflowHTTPToolActionState}
            eligibility={workflowHTTPToolActionEligibility}
            permissions={workflowHTTPToolPermissions}
            resourceKey={workflowHTTPToolResourceKey}
            locale={workflowHTTPToolLocale}
            onResourceKeyChange={setWorkflowHTTPToolResourceKey}
            onLocaleChange={setWorkflowHTTPToolLocale}
            onCreatePlan={handleCreateWorkflowHTTPToolActionPlan}
            onReloadPlan={handleReloadWorkflowHTTPToolActionPlan}
            onDecision={handleWorkflowHTTPToolActionDecision}
          />
          <WorkflowHTTPToolExecutionPanel
            plan={workflowHTTPToolActionState.actionPlan}
            state={workflowHTTPToolExecutionState}
            permissions={workflowHTTPToolPermissions}
            inputText={workflowHTTPToolExecutionInput}
            model={workflowHTTPToolExecutionModel}
            onInputTextChange={setWorkflowHTTPToolExecutionInput}
            onModelChange={setWorkflowHTTPToolExecutionModel}
            onExecute={handleRunApprovedHTTPToolActionPlan}
          />
          <Suspense fallback={<section className="workflow-rag-execution-panel"><p>Loading Workflow RAG execution…</p></section>}>
            <WorkflowRAGExecutionPanel
              applicationRef={workflowScopedApplicationId}
              draft={activeWorkflowDraft}
              savedDraftState={savedDraftConsumerState}
              draftEditDirty={workflowDraftEditDirty}
              nextDraftNumber={workspaceCreatedDrafts.filter(
                (draft) => draft.applicationRef === workflowScopedApplicationId && draft.executionProfile === "rag_retrieval_v1",
              ).length + 1}
              onCreateDraft={handleCreateWorkflowRAGDraft}
              onBindRAGRef={handleWorkflowDraftNodeRagRefChange}
              onPendingChange={setWorkflowRAGOperationPending}
              onExecutionRecorded={() => setWorkflowRunHistoryRefreshKey((key) => key + 1)}
            />
          </Suspense>
          <WorkflowExecutorPanel
            draft={activeWorkflowDraft}
            consumerState={workflowExecutorState}
            eligibility={workflowExecutorEligibility}
            inputText={workflowExecutorInput}
            model={workflowExecutorModel}
            conditionValues={activeWorkflowExecutorConditionValues}
            onCreateExecutorDraft={handleCreateWorkflowExecutorDraft}
            onInputTextChange={setWorkflowExecutorInput}
            onModelChange={setWorkflowExecutorModel}
            onConditionValueChange={handleWorkflowExecutorConditionValueChange}
            onStartRun={handleStartWorkflowRun}
            onReloadRun={handleReloadWorkflowRun}
          />
          <WorkflowDraftValidationInspectorPanel inspector={activeWorkflowDraftValidationInspector} />
          <WorkflowExecutionPlanPreviewPanel preview={activeWorkflowExecutionPlanPreview} />
          <WorkflowRuntimeReadinessInspectorPanel readiness={activeWorkflowRuntimeReadinessInspector} />
          <Suspense fallback={<section className="surface-band"><p>Loading workflow review handoff…</p></section>}>
            <WorkflowReviewHandoffPanel handoff={workflowReviewHandoff} />
          </Suspense>

          <div className="workflow-definition-states" aria-label="Workspace workflow definition states">
            {workspaceWorkflowDefinitions.statePreviews.map((state) => (
              <WorkflowDefinitionStatePreview key={state.id} state={state} />
            ))}
          </div>
        </section>

        <section hidden aria-hidden="true"
          className="surface-band workspace-run-history"
          id="workspace-run-history-legacy"
          aria-labelledby="workspace-run-history-title"
        >
          {false && <>
          <div className="section-heading">
            <div>
              <p className="eyebrow">User Workspace</p>
              <h3 id="workspace-run-history-title">Run History</h3>
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
          </>}
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
          <p>{preview.requiresConfirmation ? "archived confirmation metadata" : "read-only metadata"}</p>
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
        <StatusBadge tone="neutral">archived legacy · read-only</StatusBadge>
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
          <p className="eyebrow">Archived Legacy Confirmation</p>
          <h4>{placeholder.confirmationPlaceholderId}</h4>
        </div>
        <StatusBadge tone="neutral">{placeholder.legacyContractStatus}</StatusBadge>
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
          <div>
            <dt>Legacy contract</dt>
            <dd>{placeholder.legacyContractStatus} · historical read only</dd>
          </div>
          <div>
            <dt>Superseded by</dt>
            <dd>{placeholder.supersededBy.join(", ")}</dd>
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
        <StatusBadge tone="neutral">{field.required ? "historical required field" : "historical optional field"}</StatusBadge>
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
          <dd>{preview.requiresConfirmation ? "legacy requirement only" : "not required"}</dd>
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

function workspaceDraftCreatedConsumerState(
  config: ReturnType<typeof readWorkflowSavedDraftConsumerConfig>,
  draft: WorkflowDraftDesignerDraft,
  sourceLifecycleVersion = 0,
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
    currentLifecycleVersion: sourceLifecycleVersion,
    currentLifecycleState: "active",
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
    providerRef: workflowDraftNodeProviderRef(option.nodeType),
    toolRef: option.nodeType === "http_tool" ? "tool:workflow-preview-readonly" : "",
    ragRef: "",
    inputContractFields: workflowDraftContractFieldsForNode(option.nodeType, "input"),
    outputContractFields: workflowDraftContractFieldsForNode(option.nodeType, "output"),
    outputMappingSummary: workflowDraftNodeOutputMappingSummary(option),
    riskLevel: requiresConfirmation ? "medium" : "low",
    requiresConfirmation,
    previewOnlyReason: "Local structure edit only; workflow execution remains blocked.",
  };
}

function parseWorkflowDraftContractFields(fieldsText: string): string[] {
  const seen = new Set<string>();
  return fieldsText
    .split(/[\n,]+/)
    .map((field) => workflowDraftSafeKey(field, 80))
    .filter((field) => {
      if (!field || seen.has(field)) {
        return false;
      }
      seen.add(field);
      return true;
    });
}

function workflowDraftWithStructureEdits(
  draft: WorkflowDraftDesignerDraft,
  nodes: WorkflowDraftDesignerNode[],
): WorkflowDraftDesignerDraft {
  return {
    ...draft,
    nodes,
    edges: rebuildWorkflowDraftEdges(nodes, draft.edges),
    designerLayout: workflowDraftLayoutForNodes(draft.designerLayout, nodes),
    localOnlyInteraction: "local_edit",
  };
}

function workflowDraftLayoutForNodes(
  layout: WorkflowDraftDesignerLayout,
  nodes: WorkflowDraftDesignerNode[],
): WorkflowDraftDesignerLayout {
  const nodeIds = new Set(nodes.map((node) => node.nodeId));
  return {
    source: "workflow_node_designer",
    persistence: "ui_only",
    nodePositions: layout.nodePositions.filter((position) => nodeIds.has(position.nodeId)),
  };
}

function workflowDraftLayoutWithNodePosition(
  draft: WorkflowDraftDesignerDraft,
  nodeId: string,
  x: number,
  y: number,
): WorkflowDraftDesignerLayout {
  const nodeIds = new Set(draft.nodes.map((node) => node.nodeId));
  const nextPosition = {
    nodeId,
    x: workflowDraftDesignerCoordinate(x),
    y: workflowDraftDesignerCoordinate(y),
  };
  const positions = draft.designerLayout.nodePositions
    .filter((position) => nodeIds.has(position.nodeId) && position.nodeId !== nodeId);
  return {
    source: "workflow_node_designer",
    persistence: "ui_only",
    nodePositions: [...positions, nextPosition],
  };
}

function workflowDraftDesignerCoordinate(value: number): number {
  if (!Number.isFinite(value)) {
    return 0;
  }
  return Math.max(-10000, Math.min(10000, Math.round(value)));
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
  if (countWorkflowDraftLane(remainingNodes, "output") < 1) {
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
      workflowDraftNonEmptyConditionSummary(
        previousEdge?.conditionSummary,
        workflowDraftEdgeConditionSummary(fromNode, toNode, edgeKind),
      ),
  };
}

function buildWorkflowDraftEdgeForConnection(
  draft: WorkflowDraftDesignerDraft,
  fromNodeId: string,
  toNodeId: string,
): WorkflowDraftDesignerEdge | null {
  if (fromNodeId === toNodeId) {
    return null;
  }
  const fromNode = draft.nodes.find((node) => node.nodeId === fromNodeId);
  const toNode = draft.nodes.find((node) => node.nodeId === toNodeId);
  if (!fromNode || !toNode) {
    return null;
  }
  if (draft.edges.some((edge) => edge.fromNodeId === fromNodeId && edge.toNodeId === toNodeId)) {
    return null;
  }
  return buildWorkflowDraftEdge(fromNode, toNode, draft.edges);
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

function workflowDraftReviewableEdgeConditionSummary(
  draft: WorkflowDraftDesignerDraft,
  edge: WorkflowDraftDesignerEdge,
  conditionSummary: string,
): string {
  const fromNode = draft.nodes.find((node) => node.nodeId === edge.fromNodeId);
  const toNode = draft.nodes.find((node) => node.nodeId === edge.toNodeId);
  const fallback =
    fromNode && toNode
      ? workflowDraftEdgeConditionSummary(fromNode, toNode, edge.edgeKind)
      : "Draft edge keeps a reviewable condition summary after local graph editing.";
  return workflowDraftNonEmptyConditionSummary(conditionSummary, fallback);
}

function workflowDraftNonEmptyConditionSummary(value: string | undefined, fallback: string): string {
  const normalized = value?.trim();
  return normalized ? normalized : fallback;
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

function workflowDraftNodeProviderRef(nodeType: WorkflowDraftDesignerNode["nodeType"]): string {
  if (nodeType === "llm") {
    return "profile:radishmind-default-workflow";
  }
  if (nodeType === "condition") {
    return "policy:confirmation-gated";
  }
  return "";
}

function workflowDraftContractFieldsForNode(
  nodeType: WorkflowDraftDesignerNode["nodeType"],
  contractKind: "input" | "output",
): string[] {
  if (contractKind === "input") {
    if (nodeType === "prompt") {
      return ["tenant_ref", "application_ref", "selection_summary", "diagnostic_summary"];
    }
    if (nodeType === "llm") {
      return ["prompt_context", "answer_contract", "provider_profile_ref"];
    }
    if (nodeType === "condition") {
      return ["candidate_action", "risk_level", "confirmation_policy"];
    }
    if (nodeType === "http_tool") {
      return ["candidate_action", "audit_refs"];
    }
    return ["answer_summary", "risk_summary", "audit_refs"];
  }
  if (nodeType === "prompt") {
    return ["prompt_context"];
  }
  if (nodeType === "llm") {
    return ["answer_summary", "candidate_actions", "risk_summary", "audit_refs"];
  }
  if (nodeType === "condition") {
    return ["policy_result", "requires_confirmation"];
  }
  if (nodeType === "http_tool") {
    return ["preview_action_metadata", "audit_refs"];
  }
  return ["answer_summary", "risk_summary", "audit_refs"];
}

function workflowDraftNodeOutputMappingSummary(option: WorkflowDraftNodeTypeOption): string {
  if (option.nodeType === "llm") {
    return "Map advisory answer, candidate actions, risk summary, and audit refs into reviewable output fields.";
  }
  if (option.nodeType === "condition") {
    return "Map policy result into review-required branch metadata without unlocking execution.";
  }
  if (option.nodeType === "http_tool") {
    return "Map preview-only action metadata into audit-visible candidate action fields.";
  }
  if (option.nodeType === "output") {
    return "Map advisory fields into the read-only workspace review surface.";
  }
  return "Map sanitized context fields into the next draft node contract.";
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
          <p className="eyebrow">Full-runtime Execution Plan Preview</p>
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
          <p className="eyebrow">Full-runtime Readiness Inspector</p>
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
