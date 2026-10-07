import { useTranslation } from "react-i18next";
import type { TFunction } from "i18next";
import { useLocalePreference } from "../../i18n/LocaleProvider.tsx";
import type { UiLocale } from "../../i18n/localePreference.ts";
import { formatDisplayNumber } from "../../i18n/formatters.ts";
import type { adminShell } from "../../i18n/locales/en-US/adminShell.ts";
import "../../i18n/adminShellResources.ts";
import { lazy, Suspense, useEffect, useMemo, useState } from "react";

import {
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
const WorkspaceInvitationAdminPanel = lazy(() =>
  import("../local-identity/workspaceInvitationAdminPanel.tsx").then((module) => ({ default: module.WorkspaceInvitationAdminPanel })),
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

const AdminTenantOwner = lazy(() => import("./adminControlPlaneReadPanel.tsx").then(module => ({ default: module.AdminTenantOwner })));
const AdminAuditOwner = lazy(() => import("./adminControlPlaneReadPanel.tsx").then(module => ({ default: module.AdminAuditOwner })));

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
  const { t } = useTranslation("admin");
  const { locale } = useLocalePreference();
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
  const invitationsActive = activeSurface === "invitations";

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
          <p className="eyebrow">{quotaActive ? t($ => $.quotaHeader) : pricingActive ? t($ => $.pricingHeader) : t($ => $.controlPlane)}</p>
          <h3 id="admin-control-plane-title">
            {invitationsActive ? t($ => $.invitations) : quotaActive ? t($ => $.quotaTitle) : pricingActive ? t($ => $.pricingTitle) : t($ => $.administration)}
          </h3>
          <p>
            {invitationsActive
              ? t($ => $.invitationHelp)
              : quotaActive
              ? t($ => $.quotaHelp)
              : pricingActive
              ? t($ => $.pricingHelp)
              : t($ => $.adminHelp)}
          </p>
        </div>
        <dl>
          <div><dt>{t($ => $.tenant)}</dt><dd>{sourceConfig.tenantRef}</dd></div>
          <div><dt>{t($ => $.workspace)}</dt><dd>{sourceConfig.workspaceId ?? t($ => $.unavailable)}</dd></div>
          <div><dt>{t($ => $.authSource)}</dt><dd>{t($ => $.auth[adminAuthSourceLabel(sourceConfig, localIdentity !== null)])}</dd></div>
          <div><dt>{t($ => $.environment)}</dt><dd>{quotaActive ? quotaConfig.environment : pricingActive ? pricingConfig.environment : providerRouteConfig.environment}</dd></div>
        </dl>
      </header>

      <div className="admin-control-plane-layout">
        <nav className="admin-control-plane-path" aria-label={t($ => $.resourceNavigation)}>
          <header><span>{t($ => $.resourcePath)}</span><strong>{t($ => $.singleOwner)}</strong></header>
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
                <span><strong>{t($ => $.tasks[task.surface].label)}</strong><small>{t($ => $.tasks[task.surface].scope)}</small></span>
                <em className={status.tone}>{resourceStatusLabel(t, status, locale)}</em>
              </a>
            );
          })}
          <p className="admin-control-plane-boundary">
            <span aria-hidden="true">!</span>
            {t($ => $.boundary)}
          </p>
        </nav>

        <main className="admin-control-plane-owner" data-owner={activeSurface ?? "inactive"}>
          {activeSurface === "tenant" ? <Suspense fallback={<div className="admin-control-plane-loading">{t($ => $.loadingRead)}</div>}><AdminTenantOwner overview={tenantOverview} /></Suspense> : null}
          {invitationsActive ? <Suspense fallback={<div className="admin-control-plane-loading">{t($ => $.loadingInvitations)}</div>}>
            <WorkspaceInvitationAdminPanel tenantRef={sourceConfig.tenantRef} workspaceId={sourceConfig.workspaceId ?? ""} />
          </Suspense> : null}
          {activeSurface === "user" || activeSurface === "role" ? (
            <Suspense fallback={<div className="admin-control-plane-loading">{t($ => $.loadingIdentity)}</div>}>
              <AdminLocalIdentityOwner
                surface={activeSurface}
                tenantRef={sourceConfig.tenantRef}
                workspaceId={sourceConfig.workspaceId ?? ""}
              />
            </Suspense>
          ) : null}
          {activeSurface === "audit" ? (
            <Suspense fallback={<div className="admin-control-plane-loading">{t($ => $.loadingRead)}</div>}><AdminAuditOwner auditLog={auditLog} sourceConfig={sourceConfig} /></Suspense>
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
        <summary>{t($ => $.supportingEvidence)}</summary>
        <Suspense fallback={<div className="admin-control-plane-loading">{t($ => $.loadingEvidence)}</div>}>
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

type ResourceStatus = { kind: keyof typeof adminShell.status; count?: number; code?: string; tone: "neutral" | "blocked" | "ready" };

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
    ? { kind: "authenticated", tone: "ready" }
    : tenantOverview.collection.failureCode
    ? { kind: "failure", code: tenantOverview.collection.failureCode, tone: "blocked" }
    : { kind: "offlineEvidence", tone: "neutral" };
  const auditStatus: ResourceStatus = liveReady && auditLog.canRenderAuditLog
    ? { kind: auditLog.nextCursor ? "auditCursor" : "auditPage", count: auditLog.auditEvents.length, tone: "ready" }
    : auditLog.collection.failureCode
    ? { kind: "failure", code: auditLog.collection.failureCode, tone: "blocked" }
    : { kind: "offlineWindow", tone: "neutral" };
  const routeStatus: ResourceStatus = routeConfig.mode === "dev_admin_provider_route_http"
    ? { kind: "devControl", tone: "ready" }
    : { kind: "offline", tone: "neutral" };
  return {
    tenant: tenantStatus,
    user: localIdentityReady ? { kind: "members", tone: "ready" } : { kind: "offline", tone: "neutral" },
    role: localIdentityReady ? { kind: "roles", tone: "ready" } : { kind: "offline", tone: "neutral" },
    invitations: localIdentityReady ? { kind: "claim", tone: "ready" } : { kind: "offline", tone: "neutral" },
    audit: auditStatus,
    provider: routeStatus,
    profile: routeStatus,
    route: routeStatus,
    quota: quotaMode === "dev_admin_gateway_request_quota_http"
      ? { kind: "quota", tone: "ready" }
      : { kind: "offline", tone: "neutral" },
    pricing: pricingMode === "dev_admin_gateway_model_pricing_http"
      ? { kind: "pricing", tone: "ready" }
      : { kind: "offline", tone: "neutral" },
  };
}

function resourceStatusLabel(t: TFunction<"admin">, status: ResourceStatus, locale: UiLocale): string {
  if (status.kind === "failure") return t($ => $.status.failure, { code: status.code ?? "" });
  if (status.kind === "auditCursor" || status.kind === "auditPage") {
    const key = status.kind;
    return t($ => $.status[key], { countText: formatDisplayNumber(status.count, locale) ?? t($ => $.unavailable) });
  }
  const key = status.kind;
  return t($ => $.status[key]);
}

function adminAuthSourceLabel(config: ControlPlaneReadDevLiveConfig, localIdentityReady: boolean): keyof typeof adminShell.auth {
  if (localIdentityReady) return "localSession";
  if (config.mode !== "dev_live_http") return "offlineFixtures";
  if (config.authMode === "radish_oidc_integration_test") return "oidcTest";
  if (config.authMode === "signed_test_token") return "signedTest";
  return "devHeaders";
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
  const { t } = useTranslation("admin");
  return (
    <section className="admin-control-owner-surface admin-control-provider-owner" aria-labelledby="admin-provider-route-owner-title">
      <header>
        <div>
          <p className="eyebrow">{t($ => $.developmentOwner, { surface: t($ => $.tasks[surface].label) })}</p>
          <h4 id="admin-provider-route-owner-title">{t($ => $.owners[surface].title)}</h4>
          <p>{t($ => $.owners[surface].help)}</p>
        </div>
        <StatusPill tone={config.mode === "dev_admin_provider_route_http" ? "ready" : "neutral"}>
          {config.mode === "dev_admin_provider_route_http" ? t($ => $.control) : t($ => $.offlineOwner)}
        </StatusPill>
      </header>
      <Suspense fallback={<div className="admin-control-plane-loading">{t($ => $.loadingProvider)}</div>}>
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
  const { t } = useTranslation("admin");
  return (
    <section className="admin-control-owner-surface admin-control-quota-owner" aria-labelledby="admin-quota-owner-title">
      <header>
        <div>
          <p className="eyebrow">{t($ => $.quotaOwner)}</p>
          <h4 id="admin-quota-owner-title">{t($ => $.quotaOwnerTitle)}</h4>
          <p>{t($ => $.quotaOwnerHelp)}</p>
        </div>
        <StatusPill tone={live ? "ready" : "neutral"}>{live ? t($ => $.control) : t($ => $.offlineOwner)}</StatusPill>
      </header>
      <Suspense fallback={<div className="admin-control-plane-loading">{t($ => $.loadingQuota)}</div>}>
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
  const { t } = useTranslation("admin");
  return (
    <section className="admin-control-owner-surface admin-control-pricing-owner" aria-labelledby="admin-pricing-owner-title">
      <header>
        <div>
          <p className="eyebrow">{t($ => $.pricingOwner)}</p>
          <h4 id="admin-pricing-owner-title">{t($ => $.pricingOwnerTitle)}</h4>
          <p>{t($ => $.pricingOwnerHelp)}</p>
        </div>
        <StatusPill tone={live ? "ready" : "neutral"}>{live ? t($ => $.control) : t($ => $.offlineOwner)}</StatusPill>
      </header>
      <Suspense fallback={<div className="admin-control-plane-loading">{t($ => $.loadingPricing)}</div>}>
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
