import { useEffect, useMemo, useState } from "react";

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
  buildWorkspaceRunHistoryViewModel,
  type WorkspaceRunHistoryMetric,
  type WorkspaceRunHistoryStatePreview,
  type WorkspaceRunRecordRow,
} from "../features/control-plane-read/workspaceRunHistory";
import type {
  ControlPlaneReadCollectionViewModel,
  ControlPlaneReadRouteId,
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
  const workflowDefinitionDetail = useMemo(() => buildWorkflowDefinitionDetailViewModel(), []);
  const workspaceRunHistory = useMemo(
    () => buildWorkspaceRunHistoryViewModel(liveCollections["run-record-summary-list-route"]),
    [liveCollections],
  );

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
          <a href="#workspace-api-keys">API Keys</a>
          <a href="#workspace-usage-quota">Usage Quota</a>
          <a href="#workspace-workflow-definitions">Workflows</a>
          <a href="#workspace-run-history">Run History</a>
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
            <Fact label="Key page" value={workspaceApiKeys.canRenderApiKeys ? "ready" : "blocked"} />
            <Fact label="Quota page" value={workspaceUsageQuota.canRenderQuota ? "ready" : "blocked"} />
            <Fact
              label="Workflow page"
              value={workspaceWorkflowDefinitions.canRenderWorkflowDefinitions ? "ready" : "blocked"}
            />
            <Fact label="Run page" value={workspaceRunHistory.canRenderRuns ? "ready" : "blocked"} />
          </div>
        </header>

        <LiveReadSourceStatus state={devLiveState} baseUrl={devLiveConfig.baseUrl} />

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
              <ApplicationRow key={application.applicationRef} application={application} />
            ))}
          </div>

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
              />
            ))}
          </div>

          <WorkflowDefinitionDetailPanel detail={workflowDefinitionDetail} />

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
              <RunRecordRow key={run.runId} run={run} />
            ))}
          </div>

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

function RunRecordRow({ run }: { run: WorkspaceRunRecordRow }) {
  return (
    <article className="run-record-row">
      <div className="run-record-row-main">
        <div>
          <p className="eyebrow">{run.applicationRef}</p>
          <h4>{run.runId}</h4>
        </div>
        <StatusBadge tone={run.status === "failed" ? "bad" : "good"}>{run.status}</StatusBadge>
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

function WorkflowDefinitionRow({ workflowDefinition }: { workflowDefinition: WorkspaceWorkflowDefinitionRow }) {
  return (
    <article className="workflow-definition-row">
      <div className="workflow-definition-row-main">
        <div>
          <p className="eyebrow">{workflowDefinition.applicationRef}</p>
          <h4>{workflowDefinition.workflowDefinitionId}</h4>
        </div>
        <StatusBadge tone={workflowDefinition.definitionStatus === "published" ? "good" : "neutral"}>
          {workflowDefinition.definitionStatus}
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

function ApplicationMetric({ metric }: { metric: WorkspaceApplicationsMetric }) {
  return (
    <article className="application-metric">
      <span>{metric.label}</span>
      <strong>{metric.value}</strong>
      <p>{metric.detail}</p>
    </article>
  );
}

function ApplicationRow({ application }: { application: WorkspaceApplicationRow }) {
  return (
    <article className="application-row">
      <div className="application-row-main">
        <div>
          <p className="eyebrow">{application.applicationKind}</p>
          <h4>{application.displayName}</h4>
        </div>
        <StatusBadge tone={application.lastRunStatus === "blocked" ? "bad" : "good"}>
          {application.lastRunStatus}
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
