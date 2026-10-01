import { useTranslation } from "react-i18next";
import { LanguageSelector } from "../i18n/LanguageSelector.tsx";
import { useEffect, useRef, useState } from "react";

import type {
  ControlPlaneReadDevLiveConfig,
  ControlPlaneReadDevLiveLoadState,
} from "../features/control-plane-read/devLiveReadConsumer";
import { applicationDevelopmentStageForHash } from "../features/control-plane-read/applicationDevelopmentWorkspace";
import { useLocalIdentity } from "../features/local-identity/localIdentityContext.ts";
import { availableIdentityWorkspaces } from "../features/local-identity/localIdentityWorkspaceAccess.ts";
import {
  adminControlPlanePrimaryHref,
  applicationApiAccessPrimaryHref,
  workflowDesignerPrimaryHref,
  workflowReviewPrimaryHref,
} from "./productNavigationRoute";

type ProductNavigationCounts = {
  inbox: number;
  applications: number;
  workflows: number;
  apiKeys: number;
};

type ProductNavigationProps = {
  activeWorkspaceId: string;
  apiKeysAnchor: "#workspace-api-keys";
  counts: ProductNavigationCounts;
  sourceConfig: ControlPlaneReadDevLiveConfig;
  sourceState: ControlPlaneReadDevLiveLoadState;
  onActiveWorkspaceSwitch: (candidate: string) => boolean;
};

const PRIMARY_LINKS = [
  { href: "#workspace-overview", label: "overview", icon: "overview", countKey: null },
  { href: "#workspace-operations-inbox", label: "inbox", icon: "inbox", countKey: "inbox" },
  { href: "#workspace-applications", label: "applications", icon: "application", countKey: "applications" },
  { href: "#workspace-workflow-definitions", label: "workflows", icon: "workflow", countKey: "workflows" },
  { href: "#workspace-api-keys", label: "apiKeys", icon: "key", countKey: "apiKeys" },
  { href: "#admin-control-plane", label: "admin", icon: "admin", countKey: null },
] as const;

const REVIEW_LINKS = [
  { href: "#workspace-run-history", label: "runHistory", icon: "history" },
  { href: "#model-gateway-evidence-review", label: "evidenceReview", icon: "evidence" },
] as const;

type NavigationIconName = typeof PRIMARY_LINKS[number]["icon"] | typeof REVIEW_LINKS[number]["icon"];

const SECONDARY_GROUPS = [
  {
    label: "application",
    links: [
      ["#application-api-integration", "apiIntegration"],
      ["#application-publish-review", "publishReview"],
      ["#application-interaction-session", "interaction"],
      ["#application-rag-invocation", "rag"],
      ["#workspace-usage-quota", "quota"],
    ],
  },
  {
    label: "gateway",
    links: [
      ["#model-gateway-playground", "playground"],
      ["#model-gateway-overview", "gatewayOverview"],
      ["#model-gateway-route-evidence", "routeEvidence"],
      ["#model-gateway-usage-audit-evidence", "usageEvidence"],
    ],
  },
  {
    label: "workflowReview",
    links: [
      ["#workflow-application-detail", "applicationDetail"],
      ["#workflow-draft-designer", "draftDesigner"],
      ["#workflow-http-tool-action-review", "toolReview"],
      ["#workflow-executor-v0", "executor"],
      ["#workflow-draft-validation-inspector", "draftValidation"],
      ["#workflow-execution-plan-preview", "runtimePlan"],
      ["#workflow-runtime-readiness-inspector", "runtimeReadiness"],
      ["#workflow-scenario-inspector", "scenarioInspector"],
      ["#workflow-workspace-review", "reviewWorkspace"],
      ["#workflow-review-handoff", "reviewHandoff"],
    ],
  },
  {
    label: "adminContract",
    links: [
      ["#admin-operations-review", "operationsReview"],
      ["#admin-provider-deployment-review", "providerDeployment"],
      ["#admin-tenant-overview", "tenantOverview"],
      ["#admin-audit-log", "auditLog"],
      ["#admin-gateway-request-quota", "requestQuota"],
      ["#routes", "routeCatalog"],
      ["#states", "sharedStates"],
      ["#guard", "outputGuard"],
    ],
  },
] as const;

export function ProductNavigation({
  activeWorkspaceId,
  apiKeysAnchor,
  counts,
  sourceConfig,
  sourceState,
  onActiveWorkspaceSwitch,
}: ProductNavigationProps) {
  const { t } = useTranslation("shell");
  const [activeHash, setActiveHash] = useState(() => window.location.hash || "#workspace-overview");
  const mobileMenuRef = useRef<HTMLDetailsElement>(null);
  const identity = useLocalIdentity();
  const reportScope = identity?.onWorkspaceScopeChange;
  const workspaceChoices = identity ? availableIdentityWorkspaces(identity.profile)
    .filter((membership) => membership.tenantRef === sourceConfig.tenantRef).map((membership) => membership.workspaceId) : [];
  useEffect(() => { reportScope?.(sourceConfig.tenantRef, activeWorkspaceId); }, [reportScope, sourceConfig.tenantRef, activeWorkspaceId]);
  function switchWorkspace(candidate: string) {
    const accepted = onActiveWorkspaceSwitch(candidate);
    if (accepted) reportScope?.(sourceConfig.tenantRef, candidate.trim());
    return accepted;
  }

  useEffect(() => {
    const handleHashChange = () => {
      setActiveHash(window.location.hash || "#workspace-overview");
      mobileMenuRef.current?.removeAttribute("open");
    };
    window.addEventListener("hashchange", handleHashChange);
    return () => window.removeEventListener("hashchange", handleHashChange);
  }, []);

  return (
    <aside className="product-nav" aria-label={t($ => $.navigation)}>
      <div className="product-nav-mobile-bar">
        <ProductBrand />
        <details className="product-nav-mobile-menu" ref={mobileMenuRef}>
          <summary>{t($ => $.menu)}</summary>
          <div className="product-nav-mobile-menu-content">
            <WorkspaceSwitcher
              idPrefix="mobile"
              activeWorkspaceId={activeWorkspaceId}
              sourceState={sourceState}
              onActiveWorkspaceSwitch={switchWorkspace}
              workspaceChoices={workspaceChoices}
            />
            <LanguageSelector />
            <NavigationLinks activeHash={activeHash} apiKeysAnchor={apiKeysAnchor} counts={counts} />
          </div>
        </details>
      </div>

      <div className="product-nav-desktop">
        <ProductBrand />
        <WorkspaceSwitcher
          idPrefix="desktop"
          activeWorkspaceId={activeWorkspaceId}
          sourceState={sourceState}
          onActiveWorkspaceSwitch={switchWorkspace}
          workspaceChoices={workspaceChoices}
        />
        <NavigationLinks activeHash={activeHash} apiKeysAnchor={apiKeysAnchor} counts={counts} />
        <LanguageSelector />
        <div className="product-nav-environment">
          <span aria-hidden="true">△</span>
          <div>
            <strong>{sourceConfig.mode === "dev_live_http" ? t($ => $.development) : t($ => $.offline)}</strong>
            <small>{sourceState.status}</small>
          </div>
          <span className="product-nav-guard">{t($ => $.guarded)}</span>
        </div>
      </div>
    </aside>
  );
}

function ProductBrand() {
  const { t } = useTranslation("shell");
  return (
    <a className="product-brand" href="#workspace-overview" aria-label={t($ => $.workspaceOverview)}>
      <span className="product-brand-seal" aria-hidden="true">R</span>
      <span className="product-brand-copy">
        <strong>RadishMind</strong>
        <small>{t($ => $.workbench)}</small>
      </span>
    </a>
  );
}

function WorkspaceSwitcher({
  idPrefix,
  activeWorkspaceId,
  sourceState,
  onActiveWorkspaceSwitch,
  workspaceChoices,
}: {
  idPrefix: string;
  activeWorkspaceId: string;
  sourceState: ControlPlaneReadDevLiveLoadState;
  onActiveWorkspaceSwitch: (candidate: string) => boolean;
  workspaceChoices: string[];
}) {
  const { t } = useTranslation("shell");
  const [workspaceDraft, setWorkspaceDraft] = useState(activeWorkspaceId);
  const [workspaceFailure, setWorkspaceFailure] = useState(false);

  useEffect(() => {
    setWorkspaceDraft(activeWorkspaceId);
    setWorkspaceFailure(false);
  }, [activeWorkspaceId]);

  const inputId = `${idPrefix}-active-workspace-input`;
  const workspaceSwitchEnabled = sourceState.mode === "dev_live_http";

  return (
    <form
      className="product-workspace-switcher"
      aria-label={t($ => $.workspaceSelector)}
      onSubmit={(event) => {
        event.preventDefault();
        if (!onActiveWorkspaceSwitch(workspaceDraft)) {
          setWorkspaceFailure(true);
          return;
        }
        setWorkspaceFailure(false);
      }}
    >
      <label htmlFor={inputId}>{t($ => $.workspace)}</label>
      <div className="product-workspace-switcher-control">
        <span className="product-workspace-avatar" aria-hidden="true">WS</span>
        <span className="product-workspace-copy">
          <input
            id={inputId}
            list={`${idPrefix}-available-workspaces`}
            value={workspaceDraft}
            onChange={(event) => setWorkspaceDraft(event.target.value)}
            autoComplete="off"
            spellCheck={false}
            disabled={!workspaceSwitchEnabled}
          />
          <datalist id={`${idPrefix}-available-workspaces`}>
            {workspaceChoices.map((workspaceId) => <option key={workspaceId} value={workspaceId} />)}
          </datalist>
          <small>{t($ => $.developerWorkspace)}</small>
        </span>
        {workspaceSwitchEnabled ? (
          <button
            type="submit"
            aria-label={t($ => $.switchWorkspace)}
            disabled={sourceState.status === "loading" || workspaceDraft.trim() === activeWorkspaceId}
          >
            ↕
          </button>
        ) : null}
      </div>
      {workspaceFailure ? <small className="product-workspace-failure" role="alert">{t($ => $.invalidWorkspace)}</small> : null}
    </form>
  );
}

function NavigationLinks({
  activeHash,
  apiKeysAnchor,
  counts,
}: {
  activeHash: string;
  apiKeysAnchor: ProductNavigationProps["apiKeysAnchor"];
  counts: ProductNavigationCounts;
}) {
  const { t } = useTranslation("shell");
  const activePrimaryHref = primaryNavigationHref(activeHash, apiKeysAnchor);
  const activeReviewHref = workflowReviewPrimaryHref(activeHash) ?? activeHash;
  return (
    <nav className="product-nav-links" aria-label={t($ => $.sections)}>
      <div className="product-nav-group">
        <span className="product-nav-group-label">{t($ => $.workspace)}</span>
        {PRIMARY_LINKS.map((link) => {
          const href = link.countKey === "apiKeys" ? apiKeysAnchor : link.href;
          const count = link.countKey ? counts[link.countKey] : null;
          return (
            <a
              key={href}
              href={href}
              aria-current={activePrimaryHref === href ? "page" : undefined}
            >
              <span className="product-nav-link-main">
                <span className="product-nav-icon"><NavigationIcon name={link.icon} /></span>
                <span>{t($ => $.links[link.label])}</span>
              </span>
              {count !== null ? <small>{count}</small> : null}
            </a>
          );
        })}
      </div>

      <div className="product-nav-group">
        <span className="product-nav-group-label">{t($ => $.review)}</span>
        {REVIEW_LINKS.map((link) => (
          <a
            key={link.href}
            href={link.href}
            aria-current={activeReviewHref === link.href ? "page" : undefined}
          >
            <span className="product-nav-link-main">
              <span className="product-nav-icon"><NavigationIcon name={link.icon} /></span>
              <span>{t($ => $.links[link.label])}</span>
            </span>
          </a>
        ))}
      </div>

      <details className="product-nav-more">
        <summary>{t($ => $.moreSurfaces)}</summary>
        <div>
          {SECONDARY_GROUPS.map((group) => (
            <section key={t($ => $.links[group.label])}>
              <span>{t($ => $.links[group.label])}</span>
              {group.links.map(([href, label]) => (
                <a key={href} href={href}>{t($ => $.links[label])}</a>
              ))}
            </section>
          ))}
        </div>
      </details>
    </nav>
  );
}

function primaryNavigationHref(
  activeHash: string,
  apiKeysAnchor: ProductNavigationProps["apiKeysAnchor"],
): string | null {
  const exactPrimary = PRIMARY_LINKS.map((link) => link.countKey === "apiKeys" ? apiKeysAnchor : link.href)
    .find((href) => href === activeHash);
  if (exactPrimary) return exactPrimary;
  const applicationApiAccessHref = applicationApiAccessPrimaryHref(activeHash);
  if (applicationApiAccessHref) return applicationApiAccessHref;
  const workflowReviewHref = workflowReviewPrimaryHref(activeHash);
  if (workflowReviewHref) return workflowReviewHref;
  const adminHref = adminControlPlanePrimaryHref(activeHash);
  if (adminHref) return adminHref;
  if (activeHash === "#application-development-workspace" || applicationDevelopmentStageForHash(activeHash)) {
    return "#workspace-applications";
  }
  const workflowHref = workflowDesignerPrimaryHref(activeHash);
  if (workflowHref) return workflowHref;
  return null;
}

function NavigationIcon({ name }: { name: NavigationIconName }) {
  return (
    <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
      {name === "overview" ? (
        <>
          <rect x="3" y="3" width="7" height="7" rx="1.5" />
          <rect x="14" y="3" width="7" height="7" rx="1.5" />
          <rect x="3" y="14" width="7" height="7" rx="1.5" />
          <rect x="14" y="14" width="7" height="7" rx="1.5" />
        </>
      ) : name === "inbox" ? (
        <>
          <path d="M4 5.5h16v13H4z" />
          <path d="M4 14h4l1.5 2h5l1.5-2h4" />
        </>
      ) : name === "application" ? (
        <>
          <rect x="4" y="3.5" width="16" height="17" rx="2.5" />
          <path d="M4 8.5h16M8 6h.01" />
        </>
      ) : name === "workflow" ? (
        <>
          <rect x="3" y="3" width="6" height="6" rx="1.5" />
          <rect x="15" y="15" width="6" height="6" rx="1.5" />
          <path d="M9 6h3a3 3 0 0 1 3 3v6M12 15h3" />
        </>
      ) : name === "key" ? (
        <>
          <circle cx="8" cy="12" r="4" />
          <path d="M12 12h9M17 12v3M20 12v2" />
        </>
      ) : name === "history" ? (
        <>
          <path d="M4 5v5h5" />
          <path d="M5.5 16.5A8 8 0 1 0 5 8.5L4 10" />
          <path d="M12 8v4l3 2" />
        </>
      ) : name === "admin" ? (
        <>
          <path d="M4 9.5 12 4l8 5.5" />
          <path d="M6 10v8.5h12V10M9 18.5v-5h6v5" />
        </>
      ) : (
        <>
          <circle cx="11" cy="11" r="6" />
          <path d="m16 16 4 4M8.5 11h5M11 8.5v5" />
        </>
      )}
    </svg>
  );
}
