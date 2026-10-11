import { useTranslation } from "react-i18next";
import { lazy, Suspense, useEffect, useMemo, useState, type KeyboardEvent } from "react";

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
import { readWorkflowSavedDraftConsumerConfig } from "../features/control-plane-read/savedWorkflowDraftConsumer";
import { useWorkflowDraftWorkspace } from "../features/control-plane-read/useWorkflowDraftWorkspace";
import {
  evaluateWorkflowExecutorEligibility,
  initialWorkflowExecutorConsumerState,
  readWorkflowExecutorConsumerConfig,
  readWorkflowRunDevRecord,
  startWorkflowRunDevRecord,
  type WorkflowExecutorConsumerState,
} from "../features/control-plane-read/workflowExecutorConsumer";
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
import {
  executeWorkflowHTTPToolActionPlan,
  initialWorkflowHTTPToolExecutionState,
  restoreWorkflowHTTPToolExecutionState,
  type WorkflowHTTPToolExecutionState,
} from "../features/control-plane-read/workflowHTTPToolExecutionConsumer.ts";
import { buildModelGatewayOverviewViewModel } from "../features/control-plane-read/modelGatewayOverview";
import { buildModelGatewayRouteEvidenceViewModel } from "../features/control-plane-read/modelGatewayRouteEvidence";
import { buildModelGatewayUsageAuditEvidenceViewModel } from "../features/control-plane-read/modelGatewayUsageAuditEvidence";
import { buildModelGatewayEvidenceReviewViewModel } from "../features/control-plane-read/modelGatewayEvidenceReview";
import {
  buildControlPlaneReadShellViewModel,
} from "../features/control-plane-read/readShell";
import {
  buildWorkspaceApplicationsViewModel,
  type WorkspaceApplicationsMetric,
} from "../features/control-plane-read/workspaceApplications";
import {
  readApplicationCatalogConfig,
  type ApplicationCatalogRecord,
} from "../features/control-plane-read/applicationCatalogConsumer";
import type { ApplicationCatalogSnapshot } from "../features/control-plane-read/applicationCatalogPanel";
import {
  applicationDevelopmentSurfaceKind,
  buildApplicationDevelopmentWorkspaceContext,
} from "../features/control-plane-read/applicationDevelopmentWorkspace";
import { promptAgentTypeWorkspaceCanonicalHashForTypeSwitch } from "../features/control-plane-read/promptAgentTypeWorkspaceModel";
import {
  buildWorkspaceApiKeysViewModel,
} from "../features/control-plane-read/workspaceApiKeys";
import {
  buildWorkspaceUsageQuotaViewModel,
} from "../features/control-plane-read/workspaceUsageQuota";
import {
  buildWorkspaceWorkflowDefinitionsViewModel,
} from "../features/control-plane-read/workspaceWorkflowDefinitions";
import {
  selectWorkflowWorkspaceApplication,
  selectionForApplication,
  selectionForDraft,
  selectionForRun,
  selectionForWorkflowDefinition,
} from "../features/control-plane-read/workflowWorkspaceContext";
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

const WorkflowSurfaceOverviewPanel = lazy(() => import("../features/control-plane-read/workflowReadEvidencePanels.tsx").then(module => ({ default: module.WorkflowSurfaceOverviewPanel })));
const WorkflowScenarioInspectorPanel = lazy(() => import("../features/control-plane-read/workflowScenarioEvidencePanel.tsx").then(module => ({ default: module.WorkflowScenarioInspectorPanel })));
const WorkflowApplicationDetailPanel = lazy(() => import("../features/control-plane-read/workflowReadEvidencePanels.tsx").then(module => ({ default: module.WorkflowApplicationDetailPanel })));
const WorkflowDefinitionDetailPanel = lazy(() => import("../features/control-plane-read/workflowReadEvidencePanels.tsx").then(module => ({ default: module.WorkflowDefinitionDetailPanel })));
const WorkflowDefinitionMetric = lazy(() => import("../features/control-plane-read/workflowReadEvidencePanels.tsx").then(module => ({ default: module.WorkflowDefinitionMetric })));
const WorkflowDefinitionRow = lazy(() => import("../features/control-plane-read/workflowReadEvidencePanels.tsx").then(module => ({ default: module.WorkflowDefinitionRow })));
const WorkflowDefinitionStatePreview = lazy(() => import("../features/control-plane-read/workflowReadEvidencePanels.tsx").then(module => ({ default: module.WorkflowDefinitionStatePreview })));
const ApplicationRow = lazy(() => import("../features/control-plane-read/workflowReadEvidencePanels.tsx").then(module => ({ default: module.ApplicationRow })));
const ApplicationStatePreview = lazy(() => import("../features/control-plane-read/workflowReadEvidencePanels.tsx").then(module => ({ default: module.ApplicationStatePreview })));
const RouteCard = lazy(() => import("../features/control-plane-read/workflowReadEvidencePanels.tsx").then(module => ({ default: module.RouteCard })));
const StatePreview = lazy(() => import("../features/control-plane-read/workflowReadEvidencePanels.tsx").then(module => ({ default: module.StatePreview })));

const WorkspaceUsageQuotaPanel = lazy(() => import("../features/control-plane-read/workspaceUsageQuotaPanel.tsx"));
const WorkspaceOperationsInboxPanel = lazy(() => import("../features/control-plane-read/workspaceOperationsInboxPanel").then(module => ({ default: module.WorkspaceOperationsInboxPanel })));
const WorkspaceProductOverviewPanel = lazy(() => import("../features/control-plane-read/workspaceProductOverviewPanel").then(module => ({ default: module.WorkspaceProductOverviewPanel })));
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

const LocalIdentityGateway = lazy(() =>
  import("../features/local-identity/localIdentityGateway.tsx").then((module) => ({
    default: module.LocalIdentityGateway,
  })),
);

const localIdentityGatewayEnabled = (import.meta.env as Record<string, string | undefined>)
  .VITE_RADISHMIND_LOCAL_IDENTITY_MODE?.trim() === "local_identity_dev";
const AdminControlPlaneWorkspace = lazy(() => import("../features/control-plane-read/adminControlPlaneWorkspace"));
const ModelGatewayOverviewPanel = lazy(() => import("../features/control-plane-read/modelGatewayOverviewPanel").then(module => ({ default: module.ModelGatewayOverviewPanel })));
const ModelGatewayRouteEvidencePanel = lazy(() => import("../features/control-plane-read/modelGatewayRouteEvidencePanel").then(module => ({ default: module.ModelGatewayRouteEvidencePanel })));
const ModelGatewayUsageAuditEvidencePanel = lazy(() => import("../features/control-plane-read/modelGatewayUsageAuditEvidencePanel").then(module => ({ default: module.ModelGatewayUsageAuditEvidencePanel })));
const WorkflowWorkspaceReviewPanel = lazy(() => import("../features/control-plane-read/workflowWorkspaceReviewPanel").then(module => ({ default: module.WorkflowWorkspaceReviewPanel })));
const ModelGatewayEvidenceReviewPanel = lazy(() => import("../features/control-plane-read/modelGatewayEvidenceReviewPanel").then((module) => ({ default: module.ModelGatewayEvidenceReviewPanel })));
const ApplicationCatalogPanel = lazy(() => import("../features/control-plane-read/applicationCatalogPanel").then((module) => ({ default: module.ApplicationCatalogPanel })));
const ApplicationDevelopmentWorkspacePanel = lazy(() => import("../features/control-plane-read/applicationDevelopmentWorkspacePanel"));
const ApplicationDevelopmentWorkspaceSurface = lazy(() => import("../features/control-plane-read/applicationDevelopmentWorkspaceSurface"));
const ApplicationRuntimeReviewWorkspace = lazy(() => import("../features/control-plane-read/applicationRuntimeReviewWorkspace"));
const WorkflowReviewWorkspace = lazy(() => import("../features/control-plane-read/workflowReviewWorkspace"));
const WorkflowHTTPToolActionPanel = lazy(() => import("../features/control-plane-read/workflowHTTPToolActionPanel").then(module => ({ default: module.WorkflowHTTPToolActionPanel })));
const WorkflowHTTPToolExecutionPanel = lazy(() => import("../features/control-plane-read/workflowHTTPToolExecutionPanel").then(module => ({ default: module.WorkflowHTTPToolExecutionPanel })));
const WorkflowRAGExecutionPanel = lazy(() => import("../features/control-plane-read/workflowRAGExecutionPanel"));
const WorkflowExecutorPanel = lazy(() => import("../features/control-plane-read/workflowExecutorPanel").then(module => ({ default: module.WorkflowExecutorPanel })));
const WorkflowDraftValidationInspectorPanel = lazy(() => import("../features/control-plane-read/workflowInspectionPanels").then(module => ({ default: module.WorkflowDraftValidationInspectorPanel })));
const WorkflowExecutionPlanPreviewPanel = lazy(() => import("../features/control-plane-read/workflowInspectionPanels").then(module => ({ default: module.WorkflowExecutionPlanPreviewPanel })));
const WorkflowRuntimeReadinessInspectorPanel = lazy(() => import("../features/control-plane-read/workflowInspectionPanels").then(module => ({ default: module.WorkflowRuntimeReadinessInspectorPanel })));
const WorkflowReviewHandoffPanel = lazy(() => import("../features/control-plane-read/workflowReviewHandoffPanel").then((module) => ({ default: module.WorkflowReviewHandoffPanel })));
const DEFAULT_WORKFLOW_EXECUTOR_INPUT = "请根据当前工作流草案生成一条仅供人工审查的建议，并明确说明任何不确定性。";

type ControlPlaneReadCollectionsByRoute = Partial<
  Record<ControlPlaneReadRouteId, ControlPlaneReadCollectionViewModel>
>;

export function App() {
  if (!localIdentityGatewayEnabled) return <ProductApp />;
  return (
    <Suspense fallback={null}>
      <LocalIdentityGateway>
        <ProductApp />
      </LocalIdentityGateway>
    </Suspense>
  );
}

function ProductApp() {
  const { t } = useTranslation("shell");
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
  const selectedApplication = selectWorkflowWorkspaceApplication(workspaceApplications, selectedApplicationRef);
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
    (applicationCatalogLive ? "" : selectedApplication.applicationRef);
  const draftWorkspace = useWorkflowDraftWorkspace({
    source: {
      workspaceApplications, workspaceApiKeys, workspaceUsageQuota, workspaceWorkflowDefinitions, workspaceRunHistory,
      selection: { applicationRef: selectedApplicationRef, workflowDefinitionId: selectedWorkflowDefinitionId,
        runId: selectedRunId, draftId: selectedWorkflowDraftId, scenarioId: selectedWorkflowScenarioId },
    },
    config: activeSavedDraftConsumerConfig, applicationId: workflowScopedApplicationId,
    generationKey: applicationDevelopmentWorkspaceContext.generationKey,
    workflowExecutorOperationPending, workflowRAGOperationPending,
    onSelect: applyWorkflowSelectionPatch,
    onOpenDesigner: () => { window.location.hash = "#workflow-draft-designer"; },
  });
  const {
    workflowWorkspaceContext, savedDraftConsumerState, workflowDraftEditDirty, savedDraftConflictOpenSummary,
    createdWorkspaceDraftCountsByDefinition, editWorkflowDraft, handleWorkflowDraftEditReset,
    handleCreateWorkspaceDraftFromDefinition, handleCreateWorkflowRAGDraft, handleCreateDefinitionDerivedDraft,
    handleOpenTemplateDerivedDraft, handleDeriveSavedWorkflowDraft, handleOpenSavedWorkflowDraft,
    handleSavedWorkflowDraftLifecycleTransition, handleContinueLocalWorkflowDraftAfterConflict,
    handleOpenConflictSavedWorkflowDraft, handleValidateWorkflowDraft, handleSaveWorkflowDraft,
    handleReadWorkflowDraft, handleWorkflowDraftRevisionRestored,
  } = draftWorkspace;
  const {
    savedDraftListState, savedDraftLibraryLifecycle, savedDraftLibraryFilters, savedDraftLifecycleOperation,
    handleSavedDraftLibraryLifecycleChange, handleSavedDraftLibraryFiltersChange,
    handleRefreshSavedWorkflowDraftList, handleLoadMoreSavedWorkflowDrafts,
  } = draftWorkspace.library;
  const {
    selectedWorkflowDefinition, selectedRun, selectedWorkflowDraft, activeWorkflowDraft,
    workflowApplicationDetail, workflowDefinitionDetail, workflowRunDetail, workflowBlockedActionPreview,
    workflowConfirmationPlaceholder, workflowDraftDesigner,
    workflowDraftValidationInspector: activeWorkflowDraftValidationInspector,
    workflowExecutionPlanPreview: activeWorkflowExecutionPlanPreview,
    workflowRuntimeReadinessInspector: activeWorkflowRuntimeReadinessInspector,
    workflowSurfaceOverview, workflowScenarioInspector, workflowWorkspaceReview,
    workflowUserWorkspaceHome, savedDraftConflictReviewSummary, workflowReviewHandoff,
  } = workflowWorkspaceContext;
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
      reference.applicationId !== selectedWorkflowDraft.applicationRef ||
      reference.sourceKind !== "saved_workflow_draft" || reference.sourceId !== selectedWorkflowDraft.draftId) return;

    let canceled = false;
    setWorkflowHTTPToolActionState((state) => ({
      ...state,
      status: "reading",
      summary: "Reloading the durable action plan selected before the page refresh.",
      failureCode: "",
    }));
    readWorkflowHTTPToolActionPlan(workflowHTTPToolActionConsumerConfig, reference).then(async (state) => {
      if (canceled) return;
      setWorkflowHTTPToolActionState(state);
      if (state.actionPlan?.status !== "consumed" || !reference.runId) return;
      setWorkflowHTTPToolExecutionState((current) => ({
        ...current,
        status: "restoring",
        summary: "Reloading the exact durable v2 run; no execution request is sent.",
      }));
      const restored = await restoreWorkflowHTTPToolExecutionState(
        workflowHTTPToolActionConsumerConfig,
        state.actionPlan,
        reference.runId,
      );
      if (!canceled) setWorkflowHTTPToolExecutionState(restored);
    });
    return () => {
      canceled = true;
    };
  }, [selectedWorkflowDraft.applicationRef, selectedWorkflowDraft.draftId]);

  const canonicalizeApplicationTypeWorkspaceHash = (nextApplicationKind: string) => {
    const canonicalHash = promptAgentTypeWorkspaceCanonicalHashForTypeSwitch(
      window.location.hash,
      applicationDevelopmentWorkspaceContext.surfaceKind,
      applicationDevelopmentSurfaceKind(nextApplicationKind),
    );
    if (!canonicalHash) return;

    const oldURL = window.location.href;
    window.history.replaceState(window.history.state, "", canonicalHash);
    window.dispatchEvent(new HashChangeEvent("hashchange", { oldURL, newURL: window.location.href }));
  };
  function applyWorkflowSelectionPatch({
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
  }) {
    if (
      applicationRef !== selectedApplicationRef ||
      workflowDefinitionId !== selectedWorkflowDefinitionId ||
      draftId !== selectedWorkflowDraftId
    ) {
      draftWorkspace.invalidateSelection();
    }
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
    const nextApplication = workspaceApplications.applications.find(
      (application) => application.applicationRef === applicationRef,
    );
    if (nextApplication) canonicalizeApplicationTypeWorkspaceHash(nextApplication.applicationKind);
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
    canonicalizeApplicationTypeWorkspaceHash(record.applicationKind);
    applyWorkflowSelectionPatch({
      applicationRef: record.applicationId,
      workflowDefinitionId: null,
      runId: null,
      draftId: null,
      scenarioId: null,
    });
  };
  const handleSelectAdminQuotaApplication = (applicationId: string) => {
    if (applicationCatalogLive) {
      const record = applicationCatalogSnapshot?.records.find((item) => item.applicationId === applicationId);
      if (record) handleSelectApplicationCatalogRecord(record);
      return;
    }
    handleSelectApplication(applicationId);
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
  const handleCreateWorkflowExecutorDraft = () => {
    if (!draftWorkspace.handleCreateWorkflowExecutorDraft()) return;
    setWorkflowExecutorState(initialWorkflowExecutorConsumerState(workflowExecutorConsumerConfig));
    setWorkflowExecutorInput(DEFAULT_WORKFLOW_EXECUTOR_INPUT);
    setWorkflowExecutorModel("");
    setWorkflowExecutorConditionValues({});
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
        rememberWorkflowHTTPToolActionPlanReference(state.actionPlan, state.run?.runId ?? "");
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
    draftWorkspace.invalidateSelection();
    setSelectedApplicationRef(null);
    setApplicationCatalogSnapshot(null);
    setSelectedWorkflowDefinitionId(null);
    setSelectedRunId(null);
    setSelectedWorkflowDraftId(null);
    setSelectedWorkflowScenarioId(null);
    draftWorkspace.clearWorkspace();
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

      <section className="product-workspace" aria-label={t($ => $.appShell.controlPlaneShell)}>
        <Suspense fallback={null}>
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
        </Suspense>
        <Suspense fallback={<section className="surface-band"><p>{t($ => $.appShell.loadingSavedDraftLibrary)}</p></section>}>
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
        <Suspense fallback={<section className="surface-band"><p>{t($ => $.appShell.evidenceLoading)}</p></section>}>
          <ModelGatewayOverviewPanel overview={modelGatewayOverview} />
        </Suspense>
        <Suspense fallback={<section className="surface-band"><p>{t($ => $.appShell.evidenceLoading)}</p></section>}>
          <ModelGatewayRouteEvidencePanel detail={modelGatewayRouteEvidence} />
        </Suspense>
        <Suspense fallback={<section className="surface-band"><p>{t($ => $.appShell.evidenceLoading)}</p></section>}>
          <ModelGatewayUsageAuditEvidencePanel evidence={modelGatewayUsageAuditEvidence} />
        </Suspense>
        <Suspense fallback={<section className="surface-band"><p>{t($ => $.appShell.loadingreviewevidence)}</p></section>}>
          <ModelGatewayEvidenceReviewPanel review={modelGatewayEvidenceReview} />
        </Suspense>
        <Suspense fallback={<section className="surface-band"><p>{t($ => $.appShell.loadingAdminControlPlane)}</p></section>}>
          <AdminControlPlaneWorkspace
            tenantOverview={tenantOverview}
            auditLog={adminAuditLog}
            operationsReview={adminOperationsReview}
            providerDeploymentReview={adminProviderDeploymentReview}
            sourceConfig={activeDevLiveConfig}
            sourceState={devLiveState}
            applications={workspaceApplications.applications}
            selectedApplicationId={applicationDevelopmentWorkspaceContext.applicationId}
            selectedApplicationDisplayName={applicationDevelopmentWorkspaceContext.displayName}
            onSelectApplication={handleSelectAdminQuotaApplication}
          />
        </Suspense>
        <Suspense fallback={<section className="surface-band"><p>{t($ => $.appShell.evidenceLoading)}</p></section>}>
          <WorkflowWorkspaceReviewPanel review={workflowWorkspaceReview} />
        </Suspense>
        <Suspense fallback={null}><WorkflowSurfaceOverviewPanel overview={workflowSurfaceOverview} /></Suspense>
        <Suspense fallback={null}><WorkflowScenarioInspectorPanel
          inspector={workflowScenarioInspector}
          selectedScenarioId={workflowScenarioInspector.selectedScenarioId}
          onSelectScenario={setSelectedWorkflowScenarioId}
        /></Suspense>

        <section
          className="surface-band workspace-applications"
          id="workspace-applications"
          aria-labelledby="workspace-applications-title"
        >
          <div className="section-heading">
            <div>
              <p className="eyebrow">{t($ => $.appShell.userWorkspace)}</p>
              <h3 id="workspace-applications-title">{t($ => $.appShell.applicationsTitle)}</h3>
            </div>
            <StatusBadge tone={workspaceApplications.canRenderApplications ? "good" : "bad"}>
              {workspaceApplications.canRenderApplications ? t($ => $.appShell.readOnlyReady) : t($ => $.appShell.blocked)}
            </StatusBadge>
          </div>

          <div className="applications-summary">
            <article className="applications-route">
              <div className="card-title-row">
                <div>
                  <p className="eyebrow">{t($ => $.appShell.applicationSummaryRoute)}</p>
                  <h4>{workspaceApplications.routeId}</h4>
                </div>
                <StatusBadge tone="neutral">{workspaceApplications.requiredScope}</StatusBadge>
              </div>
              <p className="route-path">{workspaceApplications.routePath}</p>
              <dl className="tenant-meta">
                <div>
                  <dt>{t($ => $.appShell.model)}</dt>
                  <dd>{workspaceApplications.readModel}</dd>
                </div>
                <div>
                  <dt>{t($ => $.appShell.request)}</dt>
                  <dd>{workspaceApplications.requestId}</dd>
                </div>
                <div>
                  <dt>{t($ => $.appShell.nextCursor)}</dt>
                  <dd>{workspaceApplications.nextCursor ?? t($ => $.appShell.none)}</dd>
                </div>
                <div>
                  <dt>{t($ => $.appShell.audit)}</dt>
                  <dd>{workspaceApplications.auditRef}</dd>
                </div>
              </dl>
            </article>

            <div className="applications-metrics" aria-label={t($ => $.appShell.applicationMetrics)}>
              {workspaceApplications.metrics.map((metric) => (
                <ApplicationMetric key={metric.id} metric={metric} />
              ))}
            </div>
          </div>

          <Suspense fallback={<div className="application-catalog-panel"><p>{t($ => $.appShell.loadingCatalog)}</p></div>}>
            <ApplicationCatalogPanel
              key={activeWorkspaceId}
              workspaceId={activeWorkspaceId}
              selectedApplicationId={selectedApplicationRef}
              onSelectRecord={handleSelectApplicationCatalogRecord}
              onSnapshotChange={setApplicationCatalogSnapshot}
            />
          </Suspense>

          {!applicationCatalogLive ? (
            <div className="application-list" aria-label={t($ => $.appShell.workspaceApplications)}>
              {workspaceApplications.applications.map((application) => (
                <Suspense key={application.applicationRef} fallback={null}><ApplicationRow
                  key={application.applicationRef}
                  application={application}
                  selected={application.applicationRef === selectedApplication.applicationRef}
                  onSelectApplication={handleSelectApplication}
                /></Suspense>
              ))}
            </div>
          ) : null}

          {applicationDevelopmentWorkspaceContext.status !== "unavailable" ? (
            <Suspense fallback={null}><WorkflowApplicationDetailPanel detail={workflowApplicationDetail} /></Suspense>
          ) : null}

          <Suspense fallback={<div className="application-development-workspace"><p>{t($ => $.appShell.loadingWorkspace)}</p></div>}>
            <ApplicationDevelopmentWorkspacePanel
              key={applicationDevelopmentWorkspaceContext.generationKey}
              context={applicationDevelopmentWorkspaceContext}
              renderStageSurfaces={(activeStage, surfaceKey, controls) => (
                <Suspense fallback={<div className="application-development-stage-surfaces"><p>{t($ => $.appShell.loadingStage)}</p></div>}>
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
                    nextDerivedDraftNumber={draftWorkspace.nextDefinitionDerivedDraftNumber}
                    onDerivedDraft={handleCreateDefinitionDerivedDraft}
                    onTemplateDerivedDraft={handleOpenTemplateDerivedDraft}
                    onRunRecorded={() => setWorkflowRunHistoryRefreshKey((key) => key + 1)}
                  />
                </Suspense>
              )}
              renderPersistentSurfaces={(surfaceKey, controls) => (
                <>
                  <Suspense fallback={<div className="application-runtime-review-loading"><p>{t($ => $.appShell.loadingApplicationRuntimeReview)}</p></div>}>
                    <ApplicationRuntimeReviewWorkspace
                      context={applicationDevelopmentWorkspaceContext}
                      surfaceKey={surfaceKey}
                      controls={controls}
                    />
                  </Suspense>
                  <Suspense fallback={<div className="workflow-review-loading"><p>{t($ => $.appShell.loadingWorkflowReview)}</p></div>}>
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

          <div className="application-states" aria-label={t($ => $.appShell.applicationStates)}>
            {workspaceApplications.statePreviews.map((state) => (
              <Suspense key={state.id} fallback={null}><ApplicationStatePreview key={state.id} state={state} /></Suspense>
            ))}
          </div>
        </section>

        <Suspense fallback={null}>
          <WorkspaceUsageQuotaPanel view={workspaceUsageQuota} />
        </Suspense>

        <section
          className="surface-band workspace-workflow-definitions"
          id="workspace-workflow-definitions"
          aria-labelledby="workspace-workflow-definitions-title"
        >
          <div className="section-heading">
            <div>
              <p className="eyebrow">{t($ => $.appShell.userWorkspace)}</p>
              <h3 id="workspace-workflow-definitions-title">{t($ => $.appShell.workflows)}</h3>
            </div>
            <StatusBadge tone={workspaceWorkflowDefinitions.canRenderWorkflowDefinitions ? "good" : "bad"}>
              {workspaceWorkflowDefinitions.canRenderWorkflowDefinitions ? t($ => $.appShell.readOnlyReady) : t($ => $.appShell.blocked)}
            </StatusBadge>
          </div>

          <div className="workflow-definitions-summary">
            <article className="workflow-definitions-route">
              <div className="card-title-row">
                <div>
                  <p className="eyebrow">{t($ => $.appShell.definitionSummaryRoute)}</p>
                  <h4>{workspaceWorkflowDefinitions.routeId}</h4>
                </div>
                <StatusBadge tone="neutral">{workspaceWorkflowDefinitions.requiredScope}</StatusBadge>
              </div>
              <p className="route-path">{workspaceWorkflowDefinitions.routePath}</p>
              <dl className="tenant-meta">
                <div>
                  <dt>{t($ => $.appShell.model)}</dt>
                  <dd>{workspaceWorkflowDefinitions.readModel}</dd>
                </div>
                <div>
                  <dt>{t($ => $.appShell.request)}</dt>
                  <dd>{workspaceWorkflowDefinitions.requestId}</dd>
                </div>
                <div>
                  <dt>{t($ => $.appShell.nextCursor)}</dt>
                  <dd>{workspaceWorkflowDefinitions.nextCursor ?? t($ => $.appShell.none)}</dd>
                </div>
                <div>
                  <dt>{t($ => $.appShell.audit)}</dt>
                  <dd>{workspaceWorkflowDefinitions.auditRef}</dd>
                </div>
              </dl>
            </article>

            <div className="workflow-definitions-metrics" aria-label={t($ => $.appShell.definitionMetrics)}>
              {workspaceWorkflowDefinitions.metrics.map((metric, index) => (
                <Suspense key={index} fallback={null}><WorkflowDefinitionMetric key={index} metric={metric} /></Suspense>
              ))}
            </div>
          </div>

          <div className="workflow-definition-list" aria-label={t($ => $.appShell.definitions)}>
            {workspaceWorkflowDefinitions.workflowDefinitions.map((workflowDefinition) => (
              <Suspense key={workflowDefinition.workflowDefinitionId} fallback={null}><WorkflowDefinitionRow
                key={workflowDefinition.workflowDefinitionId}
                workflowDefinition={workflowDefinition}
                selected={workflowDefinition.workflowDefinitionId === selectedWorkflowDefinition.workflowDefinitionId}
                createdDraftCount={
                  createdWorkspaceDraftCountsByDefinition[workflowDefinition.workflowDefinitionId] ?? 0
                }
                onSelectWorkflowDefinition={handleSelectWorkflowDefinition}
                onCreateDraftForWorkflowDefinition={handleCreateWorkspaceDraftFromDefinition}
              /></Suspense>
            ))}
          </div>

          <Suspense fallback={null}><WorkflowDefinitionDetailPanel detail={workflowDefinitionDetail} /></Suspense>
          <Suspense fallback={<section className="workflow-draft-designer"><p>{t($ => $.appShell.loadingDesigner)}</p></section>}>
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
              onSelectDraft={handleSelectWorkflowDraft}
              onEditDraft={editWorkflowDraft}
              onResetDraftEdits={handleWorkflowDraftEditReset}
              onContinueLocalDraftAfterConflict={handleContinueLocalWorkflowDraftAfterConflict}
              onOpenConflictSavedDraft={handleOpenConflictSavedWorkflowDraft}
              onDeriveSavedDraft={handleDeriveSavedWorkflowDraft}
              onValidateDraft={handleValidateWorkflowDraft}
              onSaveDraft={handleSaveWorkflowDraft}
              onReadDraft={handleReadWorkflowDraft}
            />
          </Suspense>
          <Suspense fallback={<section className="workflow-draft-revision-panel"><p>{t($ => $.appShell.loadingRevisions)}</p></section>}>
            <WorkflowSavedDraftRevisionPanel
              key={`${draftWorkspace.editorScopeKey}:${savedDraftConsumerState.currentDraftVersion}:${savedDraftConsumerState.currentLifecycleVersion}`}
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
          <Suspense fallback={<section className="workflow-executor-v0"><p>{t($ => $.appShell.httpToolLoading)}</p></section>}>
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
          </Suspense>
          <Suspense fallback={<section className="workflow-rag-execution-panel"><p>{t($ => $.appShell.ragExecutionLoading)}</p></section>}>
            <WorkflowRAGExecutionPanel
              applicationRef={workflowScopedApplicationId}
              draft={activeWorkflowDraft}
              savedDraftState={savedDraftConsumerState}
              draftEditDirty={workflowDraftEditDirty}
              nextDraftNumber={draftWorkspace.nextRAGDraftNumber}
              onCreateDraft={handleCreateWorkflowRAGDraft}
              onBindRAGRef={draftWorkspace.bindRAGRef}
              onPendingChange={setWorkflowRAGOperationPending}
              onExecutionRecorded={() => setWorkflowRunHistoryRefreshKey((key) => key + 1)}
            />
          </Suspense>
          <Suspense fallback={null}>
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
          </Suspense>
          <Suspense fallback={null}>
          <WorkflowDraftValidationInspectorPanel inspector={activeWorkflowDraftValidationInspector} />
          <WorkflowExecutionPlanPreviewPanel preview={activeWorkflowExecutionPlanPreview} />
          <WorkflowRuntimeReadinessInspectorPanel readiness={activeWorkflowRuntimeReadinessInspector} />
          </Suspense>
          <Suspense fallback={<section className="surface-band"><p>{t($ => $.appShell.workflowReviewLoading)}</p></section>}>
            <WorkflowReviewHandoffPanel handoff={workflowReviewHandoff} />
          </Suspense>

          <div className="workflow-definition-states" aria-label={t($ => $.appShell.definitionStates)}>
            {workspaceWorkflowDefinitions.statePreviews.map((state) => (
              <Suspense key={state.id} fallback={null}><WorkflowDefinitionStatePreview key={state.id} state={state} /></Suspense>
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
              <p className="eyebrow">{t($ => $.appShell.userWorkspace)}</p>
              <h3 id="workspace-run-history-title">Run History</h3>
            </div>
            <StatusBadge tone={workspaceRunHistory.canRenderRuns ? "good" : "bad"}>
              {workspaceRunHistory.canRenderRuns ? t($ => $.appShell.readOnlyReady) : t($ => $.appShell.blocked)}
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
                  <dt>{t($ => $.appShell.model)}</dt>
                  <dd>{workspaceRunHistory.readModel}</dd>
                </div>
                <div>
                  <dt>{t($ => $.appShell.request)}</dt>
                  <dd>{workspaceRunHistory.requestId}</dd>
                </div>
                <div>
                  <dt>{t($ => $.appShell.nextCursor)}</dt>
                  <dd>{workspaceRunHistory.nextCursor ?? t($ => $.appShell.none)}</dd>
                </div>
                <div>
                  <dt>{t($ => $.appShell.audit)}</dt>
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
              <p className="eyebrow">{t($ => $.appShell.contractBinding)}</p>
              <h3 id="routes-title">{t($ => $.appShell.routeCatalog)}</h3>
            </div>
            <StatusBadge tone="neutral">{t($ => $.appShell.offlineContract)}</StatusBadge>
          </div>
          <div className="route-grid">
            {shell.routeCards.map((route) => (
              <Suspense key={route.routeId} fallback={null}><RouteCard key={route.routeId} route={route} /></Suspense>
            ))}
          </div>
        </section>

        <section className="surface-band" id="states" aria-labelledby="states-title">
          <div className="section-heading">
            <div>
              <p className="eyebrow">{t($ => $.appShell.stateModel)}</p>
              <h3 id="states-title">{t($ => $.appShell.sharedStates)}</h3>
            </div>
            <StatusBadge tone="good">{t($ => $.appShell.readOnlyReady)}</StatusBadge>
          </div>
          <div className="state-grid">
            {shell.statePreviews.map((state) => (
              <Suspense key={state.id} fallback={null}><StatePreview key={state.id} state={state} /></Suspense>
            ))}
          </div>
        </section>

        <section className="surface-band guard-band" id="guard" aria-labelledby="guard-title">
          <div className="section-heading">
            <div>
              <p className="eyebrow">{t($ => $.appShell.sensitiveOutput)}</p>
              <h3 id="guard-title">{t($ => $.appShell.outputGuard)}</h3>
            </div>
            <StatusBadge tone={shell.forbiddenProjectionBlocked ? "bad" : "good"}>
              {shell.forbiddenProjectionBlocked ? t($ => $.appShell.blocked) : t($ => $.appShell.clear)}
            </StatusBadge>
          </div>
          <div className="guard-layout">
            <div>
              <p className="metric-value">{shell.forbiddenOutputKeys.length}</p>
              <p className="metric-label">{t($ => $.appShell.blockedKeys)}</p>
            </div>
            <div className="guard-list" aria-label={t($ => $.appShell.forbiddenKeys)}>
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

function ApplicationMetric({ metric }: { metric: WorkspaceApplicationsMetric }) {
  const { t } = useTranslation("shell");
  const value = metric.id === "cursor" ? t($ => $.appShell.metrics.cursorValues[metric.value === "available" ? "available" : "none"]) : metric.value;
  const detail = metric.id === "blockedRuns" ? t($ => $.appShell.metrics.blockedDetail)
    : metric.id === "audit" ? t($ => $.appShell.metrics.auditDetail)
    : metric.id === "cursor" && metric.value === "none" ? t($ => $.appShell.metrics.singlePage) : metric.detail;
  return (
    <article className="application-metric">
      <span>{t($ => $.appShell.metrics.labels[metric.id])}</span>
      <strong>{value}</strong>
      <p>{detail}</p>
    </article>
  );
}

function StatusBadge({ children, tone }: { children: string; tone: "good" | "bad" | "neutral" }) {
  return <span className={`status-badge ${tone}`}>{children}</span>;
}
