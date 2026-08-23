import { lazy, Suspense, useEffect, useMemo, useState } from "react";

import {
  buildAdminAuditLogViewModel,
  type AdminAuditEventRow,
  type AdminAuditLogViewModel,
} from "./adminAuditLog.ts";
import {
  ADMIN_CONTROL_PLANE_RESOURCE_TASKS,
  adminControlPlaneSurfaceForHash,
  type AdminControlPlaneSurface,
} from "./adminControlPlaneRoute.ts";
import type { AdminOperationsReviewViewModel } from "./adminOperationsReview.ts";
import type { AdminProviderDeploymentReviewViewModel } from "./adminProviderDeploymentReview.ts";
import {
  readAdminProviderRouteConfig,
  type AdminProviderRouteConfig,
} from "./adminProviderRouteConsumer.ts";
import { readAdminGatewayRequestQuotaConfig } from "./adminGatewayRequestQuotaConsumer.ts";
import { readAdminGatewayModelPricingConfig } from "./adminGatewayModelPricingConsumer.ts";
import type { AdminTenantOverviewViewModel } from "./adminTenantOverview.ts";
import {
  loadControlPlaneReadDevLiveCollection,
  type ControlPlaneReadDevLiveConfig,
  type ControlPlaneReadDevLiveLoadState,
} from "./devLiveReadConsumer.ts";
import type { WorkspaceApplicationRow } from "./workspaceApplications.ts";
import { useLocalIdentity } from "../local-identity/localIdentityGateway.tsx";

const AdminLocalIdentityOwner = lazy(() =>
  import("../local-identity/adminLocalIdentityOwner.tsx").then((module) => ({
    default: module.AdminLocalIdentityOwner,
  })),
);

const AdminOperationsReviewPanel = lazy(() =>
  import("./adminOperationsReviewPanel.tsx").then((module) => ({
    default: module.AdminOperationsReviewPanel,
  })),
);
const AdminProviderDeploymentReviewPanel = lazy(() =>
  import("./adminProviderDeploymentReviewPanel.tsx").then((module) => ({
    default: module.AdminProviderDeploymentReviewPanel,
  })),
);
const AdminProviderRouteWorkspacePanel = lazy(() =>
  import("./adminProviderRouteWorkspacePanel.tsx").then((module) => ({
    default: module.AdminProviderRouteWorkspacePanel,
  })),
);
const AdminGatewayRequestQuotaPanel = lazy(() =>
  import("./adminGatewayRequestQuotaPanel.tsx").then((module) => ({
    default: module.AdminGatewayRequestQuotaPanel,
  })),
);
const AdminGatewayModelPricingPanel = lazy(() =>
  import("./adminGatewayModelPricingPanel.tsx").then((module) => ({
    default: module.AdminGatewayModelPricingPanel,
  })),
);

const providerRouteConfig = readAdminProviderRouteConfig();

export default function AdminControlPlaneWorkspace({
  tenantOverview,
  auditLog,
  operationsReview,
  providerDeploymentReview,
  sourceConfig,
  sourceState,
  applications,
  selectedApplicationId,
  selectedApplicationDisplayName,
  onSelectApplication,
}: {
  tenantOverview: AdminTenantOverviewViewModel;
  auditLog: AdminAuditLogViewModel;
  operationsReview: AdminOperationsReviewViewModel;
  providerDeploymentReview: AdminProviderDeploymentReviewViewModel;
  sourceConfig: ControlPlaneReadDevLiveConfig;
  sourceState: ControlPlaneReadDevLiveLoadState;
  applications: WorkspaceApplicationRow[];
  selectedApplicationId: string;
  selectedApplicationDisplayName: string;
  onSelectApplication: (applicationId: string) => void;
}) {
  const localIdentity = useLocalIdentity();
  const [activeSurface, setActiveSurface] = useState<AdminControlPlaneSurface | null>(() =>
    adminControlPlaneSurfaceForHash(window.location.hash),
  );
  const [supportingEvidenceOpen, setSupportingEvidenceOpen] = useState(() =>
    window.location.hash === "#admin-operations-review" ||
    window.location.hash === "#admin-provider-deployment-review",
  );

  useEffect(() => {
    function synchronizeSurface() {
      const nextSurface = adminControlPlaneSurfaceForHash(window.location.hash);
      setActiveSurface(nextSurface);
      if (
        window.location.hash === "#admin-operations-review" ||
        window.location.hash === "#admin-provider-deployment-review"
      ) {
        setSupportingEvidenceOpen(true);
      }
      if (nextSurface) {
        window.requestAnimationFrame(() => {
          document.querySelector<HTMLElement>(".admin-control-plane-workspace")
            ?.scrollIntoView({ block: "start" });
        });
      }
    }
    synchronizeSurface();
    window.addEventListener("hashchange", synchronizeSurface);
    return () => window.removeEventListener("hashchange", synchronizeSurface);
  }, []);

  useEffect(() => {
    if (!activeSurface) return;
    let secondFrame = 0;
    const firstFrame = window.requestAnimationFrame(() => {
      secondFrame = window.requestAnimationFrame(() => {
        document.querySelector<HTMLElement>(".admin-control-plane-workspace")
          ?.scrollIntoView({ block: "start" });
      });
    });
    return () => {
      window.cancelAnimationFrame(firstFrame);
      if (secondFrame) window.cancelAnimationFrame(secondFrame);
    };
  }, [activeSurface]);

  const activeTask = ADMIN_CONTROL_PLANE_RESOURCE_TASKS.find((task) => task.surface === activeSurface);
  const quotaConfig = useMemo(
    () => readAdminGatewayRequestQuotaConfig({
      tenantRef: sourceConfig.tenantRef,
      workspaceId: sourceConfig.workspaceId ?? "",
      applicationId: selectedApplicationId,
    }),
    [selectedApplicationId, sourceConfig.tenantRef, sourceConfig.workspaceId],
  );
  const pricingConfig = useMemo(
    () => readAdminGatewayModelPricingConfig({
      tenantRef: sourceConfig.tenantRef,
      workspaceId: sourceConfig.workspaceId ?? "",
    }),
    [sourceConfig.tenantRef, sourceConfig.workspaceId],
  );
  const statusBySurface = useMemo(
    () => buildResourceStatuses(
      tenantOverview,
      auditLog,
      sourceConfig,
      sourceState,
      providerRouteConfig,
      quotaConfig.mode,
      pricingConfig.mode,
      localIdentity !== null,
    ),
    [auditLog, localIdentity, pricingConfig.mode, quotaConfig.mode, sourceConfig, sourceState, tenantOverview],
  );

  const pricingActive = activeSurface === "pricing";
  const quotaActive = activeSurface === "quota";

  return (
    <section
      className="admin-control-plane-workspace"
      id={activeTask?.anchor ?? "admin-control-plane"}
      hidden={activeSurface === null}
      aria-labelledby="admin-control-plane-title"
      data-active-surface={activeSurface ?? "inactive"}
    >
      <header className="admin-control-plane-heading">
        <div>
          <p className="eyebrow">{quotaActive ? "S9 · Admin Quota Admission" : pricingActive ? "S7 · Admin Pricing" : "S7 · Admin Control Plane"}</p>
          <h3 id="admin-control-plane-title">
            {quotaActive ? "Application request quota" : pricingActive ? "Provider model pricing" : "Administration and routing"}
          </h3>
          <p>
            {quotaActive
              ? "Maintain the exact development/test application policy and its quota-owner UTC usage."
              : pricingActive
              ? "Maintain one exact immutable USD pricing revision for future development/test request snapshots."
              : "Review scoped identity evidence, then use the explicit development/test configuration owner."}
          </p>
        </div>
        <dl>
          <div><dt>Tenant</dt><dd>{sourceConfig.tenantRef}</dd></div>
          <div><dt>Workspace</dt><dd>{sourceConfig.workspaceId ?? "unavailable"}</dd></div>
          <div><dt>Auth source</dt><dd>{adminAuthSourceLabel(sourceConfig, localIdentity !== null)}</dd></div>
          <div><dt>Environment</dt><dd>{quotaActive ? quotaConfig.environment : pricingActive ? pricingConfig.environment : providerRouteConfig.environment}</dd></div>
        </dl>
      </header>

      <div className="admin-control-plane-layout">
        <nav className="admin-control-plane-path" aria-label="Admin Control Plane resources">
          <header><span>Resource path</span><strong>One owner at a time</strong></header>
          {ADMIN_CONTROL_PLANE_RESOURCE_TASKS.map((task) => {
            const selected = task.surface === activeSurface;
            const status = statusBySurface[task.surface];
            return (
              <a
                key={task.surface}
                className={`admin-control-plane-task ${selected ? "is-selected" : ""}`}
                href={`#${task.anchor}`}
                aria-current={selected ? "step" : undefined}
              >
                <i aria-hidden="true" />
                <b>{task.number}</b>
                <span><strong>{task.label}</strong><small>{task.scope}</small></span>
                <em className={status.tone}>{status.label}</em>
              </a>
            );
          })}
          <p className="admin-control-plane-boundary">
            <span aria-hidden="true">!</span>
            User and Role consume only the exact workspace member directory and server-owned built-in role catalog.
            Global account search, email lookup, invitations, custom roles, production IAM and bootstrap HTTP stay closed.
          </p>
        </nav>

        <main className="admin-control-plane-owner" data-owner={activeSurface ?? "inactive"}>
          {activeSurface === "tenant" ? <AdminTenantOwner overview={tenantOverview} /> : null}
          {activeSurface === "user" || activeSurface === "role" ? (
            <Suspense fallback={<div className="admin-control-plane-loading">Loading local identity administration…</div>}>
              <AdminLocalIdentityOwner
                surface={activeSurface}
                tenantRef={sourceConfig.tenantRef}
                workspaceId={sourceConfig.workspaceId ?? ""}
              />
            </Suspense>
          ) : null}
          {activeSurface === "audit" ? (
            <AdminAuditOwner auditLog={auditLog} sourceConfig={sourceConfig} />
          ) : null}
          {activeSurface === "provider" || activeSurface === "profile" || activeSurface === "route" ? (
            <AdminProviderRouteOwner
              surface={activeSurface}
              config={providerRouteConfig}
              selectedApplicationId={selectedApplicationId}
            />
          ) : null}
          {activeSurface === "quota" ? (
            <AdminQuotaOwner
              tenantRef={sourceConfig.tenantRef}
              workspaceId={sourceConfig.workspaceId ?? ""}
              selectedApplicationId={selectedApplicationId}
              selectedApplicationDisplayName={selectedApplicationDisplayName}
              applications={applications}
              onSelectApplication={onSelectApplication}
              live={quotaConfig.mode === "dev_admin_gateway_request_quota_http"}
            />
          ) : null}
          {activeSurface === "pricing" ? (
            <AdminPricingOwner
              tenantRef={sourceConfig.tenantRef}
              workspaceId={sourceConfig.workspaceId ?? ""}
              live={pricingConfig.mode === "dev_admin_gateway_model_pricing_http"}
            />
          ) : null}
        </main>
      </div>

      <details
        className="admin-control-plane-supporting"
        open={supportingEvidenceOpen}
        onToggle={(event) => setSupportingEvidenceOpen(event.currentTarget.open)}
      >
        <summary>Supporting readiness and deployment evidence</summary>
        <Suspense fallback={<div className="admin-control-plane-loading">Loading supporting evidence…</div>}>
          <AdminOperationsReviewPanel review={operationsReview} />
          <AdminProviderDeploymentReviewPanel
            review={providerDeploymentReview}
            includeControlledWorkspace={false}
          />
        </Suspense>
      </details>
    </section>
  );
}

type ResourceStatus = { label: string; tone: "neutral" | "blocked" | "ready" };

function buildResourceStatuses(
  tenantOverview: AdminTenantOverviewViewModel,
  auditLog: AdminAuditLogViewModel,
  sourceConfig: ControlPlaneReadDevLiveConfig,
  sourceState: ControlPlaneReadDevLiveLoadState,
  routeConfig: AdminProviderRouteConfig,
  quotaMode: "offline" | "dev_admin_gateway_request_quota_http",
  pricingMode: "offline" | "dev_admin_gateway_model_pricing_http",
  localIdentityReady: boolean,
): Record<AdminControlPlaneSurface, ResourceStatus> {
  const liveReady = sourceConfig.mode === "dev_live_http" && sourceState.status === "ready";
  const tenantStatus: ResourceStatus = liveReady && tenantOverview.canRenderTenant
    ? { label: "authenticated", tone: "ready" }
    : tenantOverview.collection.failureCode
    ? { label: tenantOverview.collection.failureCode, tone: "blocked" }
    : { label: "offline evidence", tone: "neutral" };
  const auditStatus: ResourceStatus = liveReady && auditLog.canRenderAuditLog
    ? { label: `${auditLog.auditEvents.length} / ${auditLog.nextCursor ? "cursor" : "page"}`, tone: "ready" }
    : auditLog.collection.failureCode
    ? { label: auditLog.collection.failureCode, tone: "blocked" }
    : { label: "offline window", tone: "neutral" };
  const routeStatus: ResourceStatus = routeConfig.mode === "dev_admin_provider_route_http"
    ? { label: "dev/test control", tone: "ready" }
    : { label: "offline", tone: "neutral" };
  return {
    tenant: tenantStatus,
    user: localIdentityReady ? { label: "member directory", tone: "ready" } : { label: "offline", tone: "neutral" },
    role: localIdentityReady ? { label: "built-in catalog", tone: "ready" } : { label: "offline", tone: "neutral" },
    audit: auditStatus,
    provider: routeStatus,
    profile: routeStatus,
    route: routeStatus,
    quota: quotaMode === "dev_admin_gateway_request_quota_http"
      ? { label: "UTC daily CAS", tone: "ready" }
      : { label: "offline", tone: "neutral" },
    pricing: pricingMode === "dev_admin_gateway_model_pricing_http"
      ? { label: "USD / 1M CAS", tone: "ready" }
      : { label: "offline", tone: "neutral" },
  };
}

function adminAuthSourceLabel(config: ControlPlaneReadDevLiveConfig, localIdentityReady: boolean): string {
  if (localIdentityReady) return "local Web session";
  if (config.mode !== "dev_live_http") return "offline fixtures";
  if (config.authMode === "radish_oidc_integration_test") return "OIDC integration test";
  if (config.authMode === "signed_test_token") return "signed test token";
  return "development headers";
}

function AdminTenantOwner({ overview }: { overview: AdminTenantOverviewViewModel }) {
  return (
    <section className="admin-control-owner-surface admin-control-tenant-owner" aria-labelledby="admin-tenant-owner-title">
      <header>
        <div>
          <p className="eyebrow">Tenant · {overview.requiredScope}</p>
          <h4 id="admin-tenant-owner-title">{overview.tenant?.tenant_display_name ?? "Tenant summary unavailable"}</h4>
        </div>
        <StatusPill tone={overview.canRenderTenant ? "ready" : "blocked"}>
          {overview.canRenderTenant ? "read only" : "blocked"}
        </StatusPill>
      </header>
      <div className="admin-control-tenant-layout">
        <dl className="admin-control-owner-meta">
          <div><dt>Tenant ref</dt><dd>{overview.collection.tenantRef}</dd></div>
          <div><dt>Route</dt><dd>{overview.routePath}</dd></div>
          <div><dt>Request</dt><dd>{overview.requestId}</dd></div>
          <div><dt>Audit</dt><dd>{overview.auditRef}</dd></div>
        </dl>
        <div className="admin-control-fact-list" aria-label="Tenant summary references">
          {overview.facts.map((fact) => (
            <article key={fact.label}>
              <span>{fact.label}</span><strong>{fact.value}</strong><small>{fact.detail}</small>
            </article>
          ))}
        </div>
      </div>
      <BoundaryNotice>
        This is a sanitized summary projection. It cannot create a tenant, edit membership, change plan or quota,
        reveal a raw tenant record, or establish production authorization.
      </BoundaryNotice>
    </section>
  );
}

function AdminAuditOwner({
  auditLog,
  sourceConfig,
}: {
  auditLog: AdminAuditLogViewModel;
  sourceConfig: ControlPlaneReadDevLiveConfig;
}) {
  const [pages, setPages] = useState<AdminAuditLogViewModel[]>([auditLog]);
  const [pageIndex, setPageIndex] = useState(0);
  const [selectedAuditRef, setSelectedAuditRef] = useState(auditLog.auditEvents[0]?.auditRef ?? "");
  const [paginationStatus, setPaginationStatus] = useState<"idle" | "loading" | "failed">("idle");

  useEffect(() => {
    setPages([auditLog]);
    setPageIndex(0);
    setSelectedAuditRef(auditLog.auditEvents[0]?.auditRef ?? "");
    setPaginationStatus("idle");
  }, [auditLog, sourceConfig.baseUrl, sourceConfig.tenantRef]);

  const page = pages[pageIndex] ?? auditLog;
  const selectedEvent = page.auditEvents.find((event) => event.auditRef === selectedAuditRef) ??
    page.auditEvents[0] ?? null;

  useEffect(() => {
    if (!page.auditEvents.some((event) => event.auditRef === selectedAuditRef)) {
      setSelectedAuditRef(page.auditEvents[0]?.auditRef ?? "");
    }
  }, [page, selectedAuditRef]);

  async function loadNextPage() {
    if (!page.nextCursor || sourceConfig.mode !== "dev_live_http") return;
    setPaginationStatus("loading");
    try {
      const collection = await loadControlPlaneReadDevLiveCollection(
        sourceConfig,
        "audit-summary-list-route",
        { cursor: page.nextCursor, limit: 50, sort: "recorded_at_desc" },
      );
      const nextPage = buildAdminAuditLogViewModel(collection);
      setPages((current) => [...current.slice(0, pageIndex + 1), nextPage]);
      setPageIndex((current) => current + 1);
      setSelectedAuditRef(nextPage.auditEvents[0]?.auditRef ?? "");
      setPaginationStatus("idle");
    } catch {
      setPaginationStatus("failed");
    }
  }

  return (
    <section className="admin-control-owner-surface admin-control-audit-owner" aria-labelledby="admin-audit-owner-title">
      <header>
        <div>
          <p className="eyebrow">Audit · {page.requiredScope}</p>
          <h4 id="admin-audit-owner-title">Sanitized audit window</h4>
        </div>
        <StatusPill tone={page.canRenderAuditLog ? "ready" : "blocked"}>
          {page.canRenderAuditLog ? `page ${pageIndex + 1}` : page.collection.statusLabel}
        </StatusPill>
      </header>
      <div className="admin-control-audit-toolbar">
        <span>{page.auditEvents.length} records in the current window</span>
        <span>{page.nextCursor ? "next cursor available" : "end of loaded pages"}</span>
        <div>
          <button type="button" className="secondary-action" disabled={pageIndex === 0 || paginationStatus === "loading"} onClick={() => setPageIndex((current) => current - 1)}>
            Previous page
          </button>
          <button
            type="button"
            className="secondary-action"
            disabled={!page.nextCursor || sourceConfig.mode !== "dev_live_http" || paginationStatus === "loading"}
            onClick={() => void loadNextPage()}
          >
            {paginationStatus === "loading" ? "Loading…" : "Next page"}
          </button>
        </div>
      </div>
      {paginationStatus === "failed" ? (
        <p className="admin-control-audit-failure" role="alert">The next sanitized audit page is unavailable. The current window is unchanged.</p>
      ) : null}
      <div className="admin-control-audit-layout">
        <div className="admin-control-audit-list" aria-label="Audit events in the current cursor window">
          {page.auditEvents.length ? page.auditEvents.map((event) => (
            <AuditSelectionRow
              key={event.auditRef}
              event={event}
              selected={event.auditRef === selectedEvent?.auditRef}
              onSelect={setSelectedAuditRef}
            />
          )) : (
            <div className="admin-control-audit-empty">No sanitized audit record exists in this window.</div>
          )}
        </div>
        <AuditReadOnlyDetail event={selectedEvent} requestId={page.requestId} />
      </div>
      <BoundaryNotice>
        Cursor pages are not a snapshot-isolated full history. Audit rows cannot be edited, deleted, exported as raw
        payload, or used to reveal token, claim, membership, secret, SQL, or provider material.
      </BoundaryNotice>
    </section>
  );
}

function AuditSelectionRow({
  event,
  selected,
  onSelect,
}: {
  event: AdminAuditEventRow;
  selected: boolean;
  onSelect: (auditRef: string) => void;
}) {
  return (
    <button
      type="button"
      className={`admin-control-audit-row ${selected ? "is-selected" : ""}`}
      onClick={() => onSelect(event.auditRef)}
      aria-pressed={selected}
    >
      <i aria-hidden="true" />
      <span>
        <strong>{event.eventKind}</strong>
        <small>{event.auditRef}</small>
        <small>{event.resourceRef} · {event.recordedAt}</small>
      </span>
      <StatusPill tone={event.decision === "denied" ? "blocked" : "ready"}>{event.decision}</StatusPill>
    </button>
  );
}

function AuditReadOnlyDetail({ event, requestId }: { event: AdminAuditEventRow | null; requestId: string }) {
  return (
    <aside className="admin-control-audit-detail" aria-label="Selected audit record read-only detail">
      <header>
        <div><p className="eyebrow">Read-only detail</p><h5>{event?.auditRef ?? "No event selected"}</h5></div>
        <StatusPill tone="neutral">metadata only</StatusPill>
      </header>
      {event ? (
        <dl className="admin-control-owner-meta">
          <div><dt>Actor subject</dt><dd>{event.actorSubjectRef}</dd></div>
          <div><dt>Resource</dt><dd>{event.resourceRef}</dd></div>
          <div><dt>Decision</dt><dd>{event.decision}</dd></div>
          <div><dt>Failure</dt><dd>{event.failureCode}</dd></div>
          <div><dt>Trace</dt><dd>{event.traceId}</dd></div>
          <div><dt>Recorded</dt><dd>{event.recordedAt}</dd></div>
          <div><dt>Request</dt><dd>{requestId}</dd></div>
        </dl>
      ) : <p className="admin-control-audit-empty">The current cursor window is empty.</p>}
      <div className="admin-control-readonly-actions">
        <span>Edit blocked</span><span>Delete blocked</span><span>Raw export blocked</span>
      </div>
    </aside>
  );
}

function AdminProviderRouteOwner({
  surface,
  config,
  selectedApplicationId,
}: {
  surface: "provider" | "profile" | "route";
  config: AdminProviderRouteConfig;
  selectedApplicationId: string;
}) {
  const copy = {
    provider: ["Provider inventory boundary", "Reference existing runtime inventory without copying endpoint or credential material."],
    profile: ["Provider Profile assignments", "Bind stable profile references and capabilities inside one development/test configuration."],
    route: ["Versioned model routes", "Review immutable candidates before an explicit generation switch changes later Gateway requests."],
  }[surface];
  return (
    <section className="admin-control-owner-surface admin-control-provider-owner" aria-labelledby="admin-provider-route-owner-title">
      <header>
        <div>
          <p className="eyebrow">{surface} · development / test</p>
          <h4 id="admin-provider-route-owner-title">{copy[0]}</h4>
          <p>{copy[1]}</p>
        </div>
        <StatusPill tone={config.mode === "dev_admin_provider_route_http" ? "ready" : "neutral"}>
          {config.mode === "dev_admin_provider_route_http" ? "dev/test control" : "offline"}
        </StatusPill>
      </header>
      <Suspense fallback={<div className="admin-control-plane-loading">Loading controlled configuration owner…</div>}>
        <AdminProviderRouteWorkspacePanel focus={surface} applicationId={selectedApplicationId} />
      </Suspense>
    </section>
  );
}

function AdminQuotaOwner({
  tenantRef,
  workspaceId,
  selectedApplicationId,
  selectedApplicationDisplayName,
  applications,
  onSelectApplication,
  live,
}: {
  tenantRef: string;
  workspaceId: string;
  selectedApplicationId: string;
  selectedApplicationDisplayName: string;
  applications: WorkspaceApplicationRow[];
  onSelectApplication: (applicationId: string) => void;
  live: boolean;
}) {
  return (
    <section className="admin-control-owner-surface admin-control-quota-owner" aria-labelledby="admin-quota-owner-title">
      <header>
        <div>
          <p className="eyebrow">Quota · development / test</p>
          <h4 id="admin-quota-owner-title">Application request quota admission</h4>
          <p>Read the current UTC owner, then review and confirm one expected-version policy update.</p>
        </div>
        <StatusPill tone={live ? "ready" : "neutral"}>{live ? "dev/test control" : "offline"}</StatusPill>
      </header>
      <Suspense fallback={<div className="admin-control-plane-loading">Loading application quota owner…</div>}>
        <AdminGatewayRequestQuotaPanel
          tenantRef={tenantRef}
          workspaceId={workspaceId}
          selectedApplicationId={selectedApplicationId}
          selectedApplicationDisplayName={selectedApplicationDisplayName}
          applications={applications}
          onSelectApplication={onSelectApplication}
        />
      </Suspense>
    </section>
  );
}

function AdminPricingOwner({
  tenantRef,
  workspaceId,
  live,
}: {
  tenantRef: string;
  workspaceId: string;
  live: boolean;
}) {
  return (
    <section className="admin-control-owner-surface admin-control-pricing-owner" aria-labelledby="admin-pricing-owner-title">
      <header>
        <div>
          <p className="eyebrow">Pricing · development / test</p>
          <h4 id="admin-pricing-owner-title">Immutable model pricing revisions</h4>
          <p>Read one exact Provider / Profile / Model owner, then review and confirm a CAS revision for future requests.</p>
        </div>
        <StatusPill tone={live ? "ready" : "neutral"}>{live ? "dev/test control" : "offline"}</StatusPill>
      </header>
      <Suspense fallback={<div className="admin-control-plane-loading">Loading model pricing owner…</div>}>
        <AdminGatewayModelPricingPanel tenantRef={tenantRef} workspaceId={workspaceId} />
      </Suspense>
    </section>
  );
}

function StatusPill({
  tone,
  children,
}: {
  tone: "ready" | "blocked" | "neutral";
  children: string;
}) {
  return <span className={`admin-control-status is-${tone}`}>{children}</span>;
}

function BoundaryNotice({ children }: { children: string }) {
  return <p className="admin-control-boundary-notice"><span aria-hidden="true">!</span>{children}</p>;
}
