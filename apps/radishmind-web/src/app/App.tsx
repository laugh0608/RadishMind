import {
  buildControlPlaneReadShellViewModel,
  type ControlPlaneReadRouteCard,
  type ControlPlaneReadStatePreview,
} from "../features/control-plane-read/readShell";

const shell = buildControlPlaneReadShellViewModel();

export function App() {
  return (
    <main className="product-shell">
      <aside className="product-nav" aria-label="Product areas">
        <div>
          <p className="eyebrow">RadishMind</p>
          <h1>Control Plane</h1>
          <p className="nav-summary">Read-only product surface for tenant, workspace, usage, workflow, and audit views.</p>
        </div>
        <nav className="nav-links" aria-label="Read shell sections">
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
          </div>
        </header>

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

function Fact({ label, value }: { label: string; value: string }) {
  return (
    <div className="fact">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
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
