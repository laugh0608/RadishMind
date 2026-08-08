import { useEffect, useRef, useState } from "react";

import type {
  ControlPlaneReadDevLiveConfig,
  ControlPlaneReadDevLiveLoadState,
} from "../features/control-plane-read/devLiveReadConsumer";
import { applicationDevelopmentStageForHash } from "../features/control-plane-read/applicationDevelopmentWorkspace";

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
  { href: "#workspace-overview", label: "Overview", icon: "overview", countKey: null },
  { href: "#workspace-operations-inbox", label: "Operations inbox", icon: "inbox", countKey: "inbox" },
  { href: "#workspace-applications", label: "Applications", icon: "application", countKey: "applications" },
  { href: "#workspace-workflow-definitions", label: "Workflows", icon: "workflow", countKey: "workflows" },
  { href: "#workspace-api-keys", label: "API & keys", icon: "key", countKey: "apiKeys" },
] as const;

const REVIEW_LINKS = [
  { href: "#workspace-run-history", label: "Run history", icon: "history" },
  { href: "#model-gateway-evidence-review", label: "Evidence review", icon: "evidence" },
] as const;

type NavigationIconName = typeof PRIMARY_LINKS[number]["icon"] | typeof REVIEW_LINKS[number]["icon"];

const SECONDARY_GROUPS = [
  {
    label: "Application",
    links: [
      ["#application-api-integration", "API integration"],
      ["#application-publish-review", "Publish review"],
      ["#application-interaction-session", "Application interaction"],
      ["#application-rag-invocation", "Application RAG"],
      ["#workspace-usage-quota", "Usage quota"],
    ],
  },
  {
    label: "Model gateway",
    links: [
      ["#model-gateway-playground", "Playground"],
      ["#model-gateway-overview", "Gateway overview"],
      ["#model-gateway-route-evidence", "Route evidence"],
      ["#model-gateway-usage-audit-evidence", "Usage evidence"],
    ],
  },
  {
    label: "Workflow review",
    links: [
      ["#workflow-application-detail", "Application detail"],
      ["#workflow-draft-designer", "Draft designer"],
      ["#workflow-http-tool-action-review", "HTTP tool review"],
      ["#workflow-executor-v0", "Executor v0"],
      ["#workflow-draft-validation-inspector", "Draft validation"],
      ["#workflow-execution-plan-preview", "Full-runtime plan"],
      ["#workflow-runtime-readiness-inspector", "Runtime readiness"],
      ["#workflow-scenario-inspector", "Scenario inspector"],
      ["#workflow-workspace-review", "Review workspace"],
      ["#workflow-review-handoff", "Review handoff"],
    ],
  },
  {
    label: "Admin & contract",
    links: [
      ["#admin-operations-review", "Operations review"],
      ["#admin-provider-deployment-review", "Provider deployment"],
      ["#admin-tenant-overview", "Tenant overview"],
      ["#admin-audit-log", "Audit log"],
      ["#routes", "Route catalog"],
      ["#states", "Shared states"],
      ["#guard", "Output guard"],
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
  const [activeHash, setActiveHash] = useState(() => window.location.hash || "#workspace-overview");
  const mobileMenuRef = useRef<HTMLDetailsElement>(null);

  useEffect(() => {
    const handleHashChange = () => {
      setActiveHash(window.location.hash || "#workspace-overview");
      mobileMenuRef.current?.removeAttribute("open");
    };
    window.addEventListener("hashchange", handleHashChange);
    return () => window.removeEventListener("hashchange", handleHashChange);
  }, []);

  return (
    <aside className="product-nav" aria-label="Product navigation">
      <div className="product-nav-mobile-bar">
        <ProductBrand />
        <details className="product-nav-mobile-menu" ref={mobileMenuRef}>
          <summary>Menu</summary>
          <div className="product-nav-mobile-menu-content">
            <WorkspaceSwitcher
              idPrefix="mobile"
              activeWorkspaceId={activeWorkspaceId}
              sourceState={sourceState}
              onActiveWorkspaceSwitch={onActiveWorkspaceSwitch}
            />
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
          onActiveWorkspaceSwitch={onActiveWorkspaceSwitch}
        />
        <NavigationLinks activeHash={activeHash} apiKeysAnchor={apiKeysAnchor} counts={counts} />
        <div className="product-nav-environment">
          <span aria-hidden="true">△</span>
          <div>
            <strong>{sourceConfig.mode === "dev_live_http" ? "Development / test" : "Offline fixtures"}</strong>
            <small>{sourceState.status}</small>
          </div>
          <span className="product-nav-guard">Guarded</span>
        </div>
      </div>
    </aside>
  );
}

function ProductBrand() {
  return (
    <a className="product-brand" href="#workspace-overview" aria-label="RadishMind workspace overview">
      <span className="product-brand-seal" aria-hidden="true">R</span>
      <span className="product-brand-copy">
        <strong>RadishMind</strong>
        <small>Workbench</small>
      </span>
    </a>
  );
}

function WorkspaceSwitcher({
  idPrefix,
  activeWorkspaceId,
  sourceState,
  onActiveWorkspaceSwitch,
}: {
  idPrefix: string;
  activeWorkspaceId: string;
  sourceState: ControlPlaneReadDevLiveLoadState;
  onActiveWorkspaceSwitch: (candidate: string) => boolean;
}) {
  const [workspaceDraft, setWorkspaceDraft] = useState(activeWorkspaceId);
  const [workspaceFailure, setWorkspaceFailure] = useState("");

  useEffect(() => {
    setWorkspaceDraft(activeWorkspaceId);
    setWorkspaceFailure("");
  }, [activeWorkspaceId]);

  const inputId = `${idPrefix}-active-workspace-input`;
  const workspaceSwitchEnabled = sourceState.mode === "dev_live_http";

  return (
    <form
      className="product-workspace-switcher"
      aria-label="Active workspace selector"
      onSubmit={(event) => {
        event.preventDefault();
        if (!onActiveWorkspaceSwitch(workspaceDraft)) {
          setWorkspaceFailure("Workspace reference is invalid.");
          return;
        }
        setWorkspaceFailure("");
      }}
    >
      <label htmlFor={inputId}>Workspace</label>
      <div className="product-workspace-switcher-control">
        <span className="product-workspace-avatar" aria-hidden="true">WS</span>
        <span className="product-workspace-copy">
          <input
            id={inputId}
            value={workspaceDraft}
            onChange={(event) => setWorkspaceDraft(event.target.value)}
            autoComplete="off"
            spellCheck={false}
            disabled={!workspaceSwitchEnabled}
          />
          <small>Developer workspace</small>
        </span>
        {workspaceSwitchEnabled ? (
          <button
            type="submit"
            aria-label="Switch workspace"
            disabled={sourceState.status === "loading" || workspaceDraft.trim() === activeWorkspaceId}
          >
            ↕
          </button>
        ) : null}
      </div>
      {workspaceFailure ? <small className="product-workspace-failure" role="alert">{workspaceFailure}</small> : null}
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
  const activePrimaryHref = primaryNavigationHref(activeHash, apiKeysAnchor);
  return (
    <nav className="product-nav-links" aria-label="Workspace sections">
      <div className="product-nav-group">
        <span className="product-nav-group-label">Workspace</span>
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
                <span>{link.label}</span>
              </span>
              {count !== null ? <small>{count}</small> : null}
            </a>
          );
        })}
      </div>

      <div className="product-nav-group">
        <span className="product-nav-group-label">Review</span>
        {REVIEW_LINKS.map((link) => (
          <a
            key={link.href}
            href={link.href}
            aria-current={activeHash === link.href ? "page" : undefined}
          >
            <span className="product-nav-link-main">
              <span className="product-nav-icon"><NavigationIcon name={link.icon} /></span>
              <span>{link.label}</span>
            </span>
          </a>
        ))}
      </div>

      <details className="product-nav-more">
        <summary>More surfaces</summary>
        <div>
          {SECONDARY_GROUPS.map((group) => (
            <section key={group.label}>
              <span>{group.label}</span>
              {group.links.map(([href, label]) => (
                <a key={href} href={href}>{label}</a>
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
  if (activeHash === "#application-development-workspace" || applicationDevelopmentStageForHash(activeHash)) {
    return "#workspace-applications";
  }
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
      ) : (
        <>
          <circle cx="11" cy="11" r="6" />
          <path d="m16 16 4 4M8.5 11h5M11 8.5v5" />
        </>
      )}
    </svg>
  );
}
