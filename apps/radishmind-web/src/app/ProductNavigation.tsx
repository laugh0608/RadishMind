import { useEffect, useRef, useState } from "react";

import type {
  ControlPlaneReadDevLiveConfig,
  ControlPlaneReadDevLiveLoadState,
} from "../features/control-plane-read/devLiveReadConsumer";

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
  { href: "#workspace-overview", label: "Overview", countKey: null },
  { href: "#workspace-operations-inbox", label: "Operations inbox", countKey: "inbox" },
  { href: "#workspace-applications", label: "Applications", countKey: "applications" },
  { href: "#workspace-workflow-definitions", label: "Workflows", countKey: "workflows" },
  { href: "#workspace-api-keys", label: "API & keys", countKey: "apiKeys" },
] as const;

const REVIEW_LINKS = [
  { href: "#workspace-run-history", label: "Run history" },
  { href: "#model-gateway-evidence-review", label: "Evidence review" },
] as const;

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
      <span className="product-brand-seal" aria-hidden="true">RM</span>
      <strong>RadishMind</strong>
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
      <div>
        <input
          id={inputId}
          value={workspaceDraft}
          onChange={(event) => setWorkspaceDraft(event.target.value)}
          autoComplete="off"
          spellCheck={false}
          disabled={!workspaceSwitchEnabled}
        />
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
      {workspaceFailure ? <small role="alert">{workspaceFailure}</small> : null}
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
              aria-current={activeHash === href ? "page" : undefined}
            >
              <span>{link.label}</span>
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
            <span>{link.label}</span>
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
