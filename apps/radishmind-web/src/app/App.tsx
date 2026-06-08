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
import {
  initialControlPlaneReadDevLiveLoadState,
  loadControlPlaneReadDevLiveCollections,
  readControlPlaneReadDevLiveConfig,
  type ControlPlaneReadDevLiveLoadState,
} from "../features/control-plane-read/devLiveReadConsumer";
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
  buildWorkflowApplicationDetailViewModel,
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
  buildWorkflowDefinitionDetailViewModel,
  type WorkflowDefinitionBlockedActionPreview,
  type WorkflowDefinitionDetailEdge,
  type WorkflowDefinitionDetailNode,
  type WorkflowDefinitionDetailSchemaSummary,
  type WorkflowDefinitionDetailViewModel,
} from "../features/control-plane-read/workflowDefinitionDetail";
import {
  buildWorkflowDraftDesignerViewModel,
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
  buildWorkflowDraftValidationInspectorViewModel,
  type WorkflowDraftBlockedCapabilityCheck,
  type WorkflowDraftContractCheck,
  type WorkflowDraftStructuralCheck,
  type WorkflowDraftValidationInspectorViewModel,
  type WorkflowDraftValidationSummary,
} from "../features/control-plane-read/workflowDraftValidationInspector";
import {
  buildWorkflowExecutionPlanPreviewViewModel,
  type WorkflowExecutionPlanBlockedReason,
  type WorkflowExecutionPlanGate,
  type WorkflowExecutionPlanNodeMapping,
  type WorkflowExecutionPlanPreviewViewModel,
  type WorkflowExecutionPlanProviderRequirement,
  type WorkflowExecutionPlanStage,
  type WorkflowExecutionPlanSummary,
} from "../features/control-plane-read/workflowExecutionPlanPreview";
import {
  buildWorkflowRuntimeReadinessInspectorViewModel,
  type WorkflowRuntimeReadinessBlocker,
  type WorkflowRuntimeReadinessGate,
  type WorkflowRuntimeReadinessInspectorViewModel,
  type WorkflowRuntimeReadinessPrerequisite,
  type WorkflowRuntimeReadinessStatus,
  type WorkflowRuntimeReadinessSummary,
} from "../features/control-plane-read/workflowRuntimeReadinessInspector";
import {
  buildWorkflowSurfaceOverviewViewModel,
  type WorkflowSurfaceOverviewBlockedCapability,
  type WorkflowSurfaceOverviewMetric,
  type WorkflowSurfaceOverviewRelation,
  type WorkflowSurfaceOverviewStatus,
  type WorkflowSurfaceOverviewStopLine,
  type WorkflowSurfaceOverviewViewModel,
} from "../features/control-plane-read/workflowSurfaceOverview";
import {
  buildWorkflowWorkspaceReviewViewModel,
} from "../features/control-plane-read/workflowWorkspaceReview";
import { WorkflowWorkspaceReviewPanel } from "../features/control-plane-read/workflowWorkspaceReviewPanel";
import {
  buildWorkflowScenarioInspectorViewModel,
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
  buildWorkflowRunDetailViewModel,
  type WorkflowRunDetailGuardPreview,
  type WorkflowRunDetailSummary,
  type WorkflowRunDetailTimelineEvent,
  type WorkflowRunDetailViewModel,
} from "../features/control-plane-read/workflowRunDetail";
import {
  buildWorkflowBlockedActionPreviewViewModel,
  type WorkflowBlockedActionAuditStep,
  type WorkflowBlockedActionPreviewViewModel,
  type WorkflowBlockedActionRequirement,
  type WorkflowConfirmationPlaceholderPreview,
} from "../features/control-plane-read/workflowBlockedActionPreview";
import {
  buildWorkflowConfirmationPlaceholderViewModel,
  type WorkflowConfirmationDecisionField,
  type WorkflowConfirmationPlaceholderPrerequisite,
  type WorkflowConfirmationPlaceholderViewModel,
} from "../features/control-plane-read/workflowConfirmationPlaceholder";
import type {
  ControlPlaneReadCollectionViewModel,
  ControlPlaneReadRouteId,
  WorkflowDefinitionSummary,
} from "../../../../contracts/typescript/control-plane-read-api";

const shell = buildControlPlaneReadShellViewModel();
const devLiveConfig = readControlPlaneReadDevLiveConfig();

type ControlPlaneReadCollectionsByRoute = Partial<
  Record<ControlPlaneReadRouteId, ControlPlaneReadCollectionViewModel>
>;

export function App() {
  const [devLiveState, setDevLiveState] = useState<ControlPlaneReadDevLiveLoadState>(() =>
    initialControlPlaneReadDevLiveLoadState(devLiveConfig),
  );
  const [selectedApplicationRef, setSelectedApplicationRef] = useState<string | null>(null);
  const [selectedWorkflowDefinitionId, setSelectedWorkflowDefinitionId] = useState<string | null>(null);
  const [selectedRunId, setSelectedRunId] = useState<string | null>(null);
  const [selectedWorkflowDraftId, setSelectedWorkflowDraftId] = useState<string | null>(null);
  const [selectedWorkflowScenarioId, setSelectedWorkflowScenarioId] = useState<string | null>(null);

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
  const selectedApplication = useMemo<WorkspaceApplicationRow>(() => {
    const targetApplicationRef = selectedApplicationRef ?? workspaceApplications.applications[0]?.applicationRef;
    return (
      workspaceApplications.applications.find(
        (application) => application.applicationRef === targetApplicationRef,
      ) ?? workspaceApplications.applications[0]!
    );
  }, [selectedApplicationRef, workspaceApplications]);
  const workflowApplicationDetail = useMemo(
    () =>
      buildWorkflowApplicationDetailViewModel(selectedApplication, {
        tenantRef: workspaceApplications.collection.tenantRef,
        requestId: workspaceApplications.requestId,
        auditRef: workspaceApplications.auditRef,
      }),
    [selectedApplication, workspaceApplications],
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
  const workflowDefinitionsForSelectedApplication = useMemo(() => {
    const filteredDefinitions = workspaceWorkflowDefinitions.workflowDefinitions.filter(
      (workflowDefinition) => workflowDefinition.applicationRef === selectedApplication.applicationRef,
    );
    return filteredDefinitions.length > 0 ? filteredDefinitions : workspaceWorkflowDefinitions.workflowDefinitions;
  }, [selectedApplication, workspaceWorkflowDefinitions]);
  const selectedWorkflowDefinition = useMemo<WorkspaceWorkflowDefinitionRow>(() => {
    const targetWorkflowDefinitionId =
      selectedWorkflowDefinitionId ??
      selectedApplication.latestWorkflowDefinitionRef ??
      workflowDefinitionsForSelectedApplication[0]?.workflowDefinitionId;
    return (
      workflowDefinitionsForSelectedApplication.find(
        (workflowDefinition) => workflowDefinition.workflowDefinitionId === targetWorkflowDefinitionId,
      ) ??
      workflowDefinitionsForSelectedApplication.find(
        (workflowDefinition) =>
          workflowDefinition.workflowDefinitionId === selectedApplication.latestWorkflowDefinitionRef,
      ) ??
      workflowDefinitionsForSelectedApplication[0]!
    );
  }, [selectedApplication, selectedWorkflowDefinitionId, workflowDefinitionsForSelectedApplication]);
  const workflowDefinitionDetail = useMemo(
    () =>
      buildWorkflowDefinitionDetailViewModel(
        toWorkflowDefinitionSummary(selectedWorkflowDefinition, workspaceWorkflowDefinitions.collection.tenantRef),
      ),
    [selectedWorkflowDefinition, workspaceWorkflowDefinitions],
  );
  const workspaceRunHistory = useMemo(
    () => buildWorkspaceRunHistoryViewModel(liveCollections["run-record-summary-list-route"]),
    [liveCollections],
  );
  const runsForSelectedContext = useMemo(() => {
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
  }, [selectedApplication, selectedWorkflowDefinition, workspaceRunHistory]);
  const selectedRun = useMemo<WorkspaceRunRecordRow>(() => {
    const targetRunId = selectedRunId ?? runsForSelectedContext[0]?.runId;
    return runsForSelectedContext.find((run) => run.runId === targetRunId) ?? runsForSelectedContext[0]!;
  }, [runsForSelectedContext, selectedRunId]);
  const workflowRunDetail = useMemo(
    () => buildWorkflowRunDetailViewModel(selectedRun),
    [selectedRun],
  );
  const workflowBlockedActionPreview = useMemo(
    () =>
      buildWorkflowBlockedActionPreviewViewModel(
        workflowDefinitionDetail.blockedActionPreview,
        workflowRunDetail.blockedReplayPreview,
      ),
    [workflowDefinitionDetail, workflowRunDetail],
  );
  const workflowConfirmationPlaceholder = useMemo(
    () => buildWorkflowConfirmationPlaceholderViewModel(workflowBlockedActionPreview),
    [workflowBlockedActionPreview],
  );
  const workflowDraftDesigner = useMemo(
    () =>
      buildWorkflowDraftDesignerViewModel({
        workflowDefinitions: workspaceWorkflowDefinitions.workflowDefinitions,
        detailNodes: workflowDefinitionDetail.nodes,
        detailEdges: workflowDefinitionDetail.edges,
        blockedActionPreview: workflowDefinitionDetail.blockedActionPreview,
        confirmationPlaceholder: workflowConfirmationPlaceholder,
        sourceRequestId: workspaceWorkflowDefinitions.requestId,
        sourceAuditRef: workspaceWorkflowDefinitions.auditRef,
      }),
    [workspaceWorkflowDefinitions, workflowDefinitionDetail, workflowConfirmationPlaceholder],
  );
  const selectedWorkflowDraft = useMemo<WorkflowDraftDesignerDraft>(() => {
    const definitionDraft = workflowDraftDesigner.drafts.find(
      (draft) => draft.workflowDefinitionId === selectedWorkflowDefinition.workflowDefinitionId,
    );
    const targetDraftId = selectedWorkflowDraftId ?? definitionDraft?.draftId ?? workflowDraftDesigner.defaultDraftId;
    return (
      workflowDraftDesigner.drafts.find((draft) => draft.draftId === targetDraftId) ??
      definitionDraft ??
      workflowDraftDesigner.drafts[0]!
    );
  }, [selectedWorkflowDefinition, selectedWorkflowDraftId, workflowDraftDesigner]);
  const workflowDraftValidationInspector = useMemo(
    () => buildWorkflowDraftValidationInspectorViewModel(selectedWorkflowDraft),
    [selectedWorkflowDraft],
  );
  const workflowExecutionPlanPreview = useMemo(
    () => buildWorkflowExecutionPlanPreviewViewModel(selectedWorkflowDraft, workflowDraftValidationInspector),
    [selectedWorkflowDraft, workflowDraftValidationInspector],
  );
  const workflowRuntimeReadinessInspector = useMemo(
    () => buildWorkflowRuntimeReadinessInspectorViewModel(workflowExecutionPlanPreview),
    [workflowExecutionPlanPreview],
  );
  const workflowSurfaceOverview = useMemo(
    () =>
      buildWorkflowSurfaceOverviewViewModel({
        applicationDetail: workflowApplicationDetail,
        definitionDetail: workflowDefinitionDetail,
        runDetail: workflowRunDetail,
        selectedDraft: selectedWorkflowDraft,
        validationInspector: workflowDraftValidationInspector,
        executionPlanPreview: workflowExecutionPlanPreview,
        runtimeReadinessInspector: workflowRuntimeReadinessInspector,
      }),
    [
      workflowApplicationDetail,
      workflowDefinitionDetail,
      workflowRunDetail,
      selectedWorkflowDraft,
      workflowDraftValidationInspector,
      workflowExecutionPlanPreview,
      workflowRuntimeReadinessInspector,
    ],
  );
  const workflowScenarioInspector = useMemo(
    () =>
      buildWorkflowScenarioInspectorViewModel(
        {
          applicationDetail: workflowApplicationDetail,
          definitionDetail: workflowDefinitionDetail,
          runDetail: workflowRunDetail,
          selectedDraft: selectedWorkflowDraft,
          validationInspector: workflowDraftValidationInspector,
          executionPlanPreview: workflowExecutionPlanPreview,
          runtimeReadinessInspector: workflowRuntimeReadinessInspector,
        },
        selectedWorkflowScenarioId,
      ),
    [
      workflowApplicationDetail,
      workflowDefinitionDetail,
      workflowRunDetail,
      selectedWorkflowDraft,
      workflowDraftValidationInspector,
      workflowExecutionPlanPreview,
      workflowRuntimeReadinessInspector,
      selectedWorkflowScenarioId,
    ],
  );
  const workflowWorkspaceReview = useMemo(
    () =>
      buildWorkflowWorkspaceReviewViewModel({
        applicationDetail: workflowApplicationDetail,
        definitionDetail: workflowDefinitionDetail,
        runDetail: workflowRunDetail,
        selectedDraft: selectedWorkflowDraft,
        validationInspector: workflowDraftValidationInspector,
        executionPlanPreview: workflowExecutionPlanPreview,
        runtimeReadinessInspector: workflowRuntimeReadinessInspector,
        surfaceOverview: workflowSurfaceOverview,
        scenarioInspector: workflowScenarioInspector,
      }),
    [
      workflowApplicationDetail,
      workflowDefinitionDetail,
      workflowRunDetail,
      selectedWorkflowDraft,
      workflowDraftValidationInspector,
      workflowExecutionPlanPreview,
      workflowRuntimeReadinessInspector,
      workflowSurfaceOverview,
      workflowScenarioInspector,
    ],
  );
  const handleSelectApplication = (applicationRef: string) => {
    const nextApplication = workspaceApplications.applications.find(
      (application) => application.applicationRef === applicationRef,
    );
    const nextDefinition =
      workspaceWorkflowDefinitions.workflowDefinitions.find(
        (workflowDefinition) =>
          workflowDefinition.applicationRef === applicationRef &&
          workflowDefinition.workflowDefinitionId === nextApplication?.latestWorkflowDefinitionRef,
      ) ??
      workspaceWorkflowDefinitions.workflowDefinitions.find(
        (workflowDefinition) => workflowDefinition.applicationRef === applicationRef,
      );
    const nextRun =
      workspaceRunHistory.runs.find(
        (run) =>
          run.applicationRef === applicationRef &&
          (!nextDefinition || run.workflowDefinitionId === nextDefinition.workflowDefinitionId),
      ) ?? workspaceRunHistory.runs.find((run) => run.applicationRef === applicationRef);

    setSelectedApplicationRef(applicationRef);
    setSelectedWorkflowDefinitionId(nextDefinition?.workflowDefinitionId ?? null);
    setSelectedRunId(nextRun?.runId ?? null);
    setSelectedWorkflowDraftId(null);
    setSelectedWorkflowScenarioId(null);
  };
  const handleSelectWorkflowDefinition = (workflowDefinitionId: string) => {
    const nextDefinition = workspaceWorkflowDefinitions.workflowDefinitions.find(
      (workflowDefinition) => workflowDefinition.workflowDefinitionId === workflowDefinitionId,
    );
    const nextRun = workspaceRunHistory.runs.find(
      (run) =>
        run.workflowDefinitionId === workflowDefinitionId &&
        (!nextDefinition || run.applicationRef === nextDefinition.applicationRef),
    );

    setSelectedWorkflowDefinitionId(workflowDefinitionId);
    if (nextDefinition) {
      setSelectedApplicationRef(nextDefinition.applicationRef);
    }
    setSelectedRunId(nextRun?.runId ?? null);
    setSelectedWorkflowDraftId(null);
    setSelectedWorkflowScenarioId(null);
  };
  const handleSelectRun = (runId: string) => {
    const nextRun = workspaceRunHistory.runs.find((run) => run.runId === runId);

    setSelectedRunId(runId);
    if (nextRun) {
      setSelectedApplicationRef(nextRun.applicationRef);
      setSelectedWorkflowDefinitionId(nextRun.workflowDefinitionId);
      setSelectedWorkflowDraftId(null);
      setSelectedWorkflowScenarioId(null);
    }
  };
  const handleSelectWorkflowDraft = (draftId: string) => {
    const nextDraft = workflowDraftDesigner.drafts.find((draft) => draft.draftId === draftId);
    const nextRun = workspaceRunHistory.runs.find(
      (run) =>
        run.applicationRef === nextDraft?.applicationRef &&
        run.workflowDefinitionId === nextDraft?.workflowDefinitionId,
    );

    setSelectedWorkflowDraftId(draftId);
    if (nextDraft) {
      setSelectedApplicationRef(nextDraft.applicationRef);
      setSelectedWorkflowDefinitionId(nextDraft.workflowDefinitionId);
      setSelectedRunId(nextRun?.runId ?? null);
      setSelectedWorkflowScenarioId(null);
    }
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
          <a href="#admin-tenant-overview">Tenant Overview</a>
          <a href="#admin-audit-log">Audit Log</a>
          <a href="#workspace-applications">Applications</a>
          <a href="#workflow-application-detail">Application Detail</a>
          <a href="#workflow-workspace-review">Workflow Review</a>
          <a href="#workflow-surface-overview">Workflow Overview</a>
          <a href="#workflow-scenario-inspector">Scenario Inspector</a>
          <a href="#workspace-api-keys">API Keys</a>
          <a href="#workspace-usage-quota">Usage Quota</a>
          <a href="#workspace-workflow-definitions">Workflows</a>
          <a href="#workflow-draft-designer">Draft Designer</a>
          <a href="#workflow-draft-validation-inspector">Draft Validation</a>
          <a href="#workflow-execution-plan-preview">Execution Plan</a>
          <a href="#workflow-runtime-readiness-inspector">Runtime Readiness</a>
          <a href="#workspace-run-history">Run History</a>
          <a href="#workflow-blocked-action-preview">Blocked Action</a>
          <a href="#workflow-confirmation-placeholder">Confirmation</a>
          <a href="#routes">Route Catalog</a>
          <a href="#states">Shared States</a>
          <a href="#guard">Output Guard</a>
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
              value={workflowDraftValidationInspector.validationStatus}
            />
            <Fact
              label="Plan"
              value={workflowExecutionPlanPreview.canRenderExecutionPlanPreview ? "preview" : "blocked"}
            />
            <Fact
              label="Runtime"
              value={workflowRuntimeReadinessInspector.canRenderRuntimeReadinessInspector ? "blocked" : "missing"}
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
          </div>
        </header>

        <LiveReadSourceStatus state={devLiveState} baseUrl={devLiveConfig.baseUrl} />
        <WorkflowWorkspaceReviewPanel review={workflowWorkspaceReview} />
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
            <p>Displayed as read-side metadata only; enforcement, rate limit and billing ledger remain outside this page.</p>
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
                onSelectWorkflowDefinition={handleSelectWorkflowDefinition}
              />
            ))}
          </div>

          <WorkflowDefinitionDetailPanel detail={workflowDefinitionDetail} />
          <WorkflowDraftDesignerPanel
            designer={workflowDraftDesigner}
            selectedDraft={selectedWorkflowDraft}
            selectedDraftId={selectedWorkflowDraft.draftId}
            onSelectDraft={handleSelectWorkflowDraft}
          />
          <WorkflowDraftValidationInspectorPanel inspector={workflowDraftValidationInspector} />
          <WorkflowExecutionPlanPreviewPanel preview={workflowExecutionPlanPreview} />
          <WorkflowRuntimeReadinessInspectorPanel readiness={workflowRuntimeReadinessInspector} />

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
  onSelectWorkflowDefinition,
}: {
  workflowDefinition: WorkspaceWorkflowDefinitionRow;
  selected: boolean;
  onSelectWorkflowDefinition: (workflowDefinitionId: string) => void;
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
  onSelectDraft,
}: {
  designer: WorkflowDraftDesignerViewModel;
  selectedDraft: WorkflowDraftDesignerDraft;
  selectedDraftId: string;
  onSelectDraft: (draftId: string) => void;
}) {
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

      <div className="workflow-draft-node-grid" aria-label="Workflow draft nodes">
        {selectedDraft.nodes.map((node) => (
          <WorkflowDraftNodeCard key={node.nodeId} node={node} />
        ))}
      </div>

      <div className="workflow-draft-edge-grid" aria-label="Workflow draft edges">
        {selectedDraft.edges.map((edge) => (
          <WorkflowDraftEdgeCard key={edge.edgeId} edge={edge} />
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

function WorkflowDraftNodeCard({ node }: { node: WorkflowDraftDesignerNode }) {
  return (
    <article className="workflow-draft-node">
      <div className="workflow-draft-row-main">
        <div>
          <p className="eyebrow">
            {node.lane} / {node.nodeType}
          </p>
          <h5>{node.label}</h5>
        </div>
        <StatusBadge tone={node.readiness === "blocked" ? "bad" : node.readiness === "ready" ? "good" : "neutral"}>
          {node.readiness}
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
        <div>
          <dt>Preview</dt>
          <dd>{node.previewOnlyReason}</dd>
        </div>
      </dl>
    </article>
  );
}

function WorkflowDraftEdgeCard({ edge }: { edge: WorkflowDraftDesignerEdge }) {
  return (
    <article className="workflow-draft-edge">
      <span>{edge.edgeKind}</span>
      <strong>
        {edge.fromNodeId} to {edge.toNodeId}
      </strong>
      <p>{edge.conditionSummary}</p>
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
