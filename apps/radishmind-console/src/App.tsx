import { useCallback, useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import {
  DEFAULT_PLATFORM_BASE_URL,
  buildPlatformOverviewEndpoint,
  loadPlatformOverview,
  resolvePlatformBaseUrl,
  type PlatformOverviewLoadState,
} from "./platformOverviewClient";

const STOP_LINE_LABELS: Record<string, string> = {
  real_executor_enabled: "Real executor",
  durable_store_enabled: "Durable store",
  confirmation_flow_connected: "Confirmation flow",
  materialized_result_reader: "Materialized result reader",
  long_term_memory_enabled: "Long-term memory",
  business_truth_write_enabled: "Business truth write",
  automatic_replay_enabled: "Automatic replay",
  production_secret_backend_ready: "Production secret backend",
};

const initialBaseUrl = resolvePlatformBaseUrl();

export function App() {
  const [baseUrl, setBaseUrl] = useState(initialBaseUrl);
  const [loadState, setLoadState] = useState<PlatformOverviewLoadState>({
    status: "idle",
    endpoint: buildPlatformOverviewEndpoint(initialBaseUrl),
  });

  const refreshOverview = useCallback(async () => {
    const endpoint = buildPlatformOverviewEndpoint(baseUrl);
    setLoadState({ status: "loading", endpoint });
    try {
      setLoadState(await loadPlatformOverview(baseUrl));
    } catch (error) {
      setLoadState({
        status: "error",
        endpoint,
        message: error instanceof Error ? error.message : "overview request failed",
      });
    }
  }, [baseUrl]);

  useEffect(() => {
    void refreshOverview();
  }, [refreshOverview]);

  const viewModel = loadState.status === "ready" ? loadState.viewModel : null;
  const overview = loadState.status === "ready" ? loadState.overview : null;
  const stopLineItems = useMemo(
    () =>
      viewModel?.stopLines.blockedCapabilityIds.map((capabilityId) => ({
        id: capabilityId,
        label: STOP_LINE_LABELS[capabilityId] ?? capabilityId,
      })) ?? [],
    [viewModel],
  );

  return (
    <main className="console-shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">RadishMind</p>
          <h1>RadishMind Console</h1>
        </div>
        <div className="connection-controls">
          <label className="base-url-field">
            <span>Platform URL</span>
            <input
              value={baseUrl}
              onChange={(event) => setBaseUrl(event.target.value)}
              placeholder={DEFAULT_PLATFORM_BASE_URL}
            />
          </label>
          <button type="button" onClick={refreshOverview} disabled={loadState.status === "loading"}>
            {loadState.status === "loading" ? "Refreshing" : "Refresh"}
          </button>
        </div>
      </header>

      <section className="status-strip" aria-live="polite">
        <StatusPill tone={loadState.status === "ready" ? "good" : loadState.status === "error" ? "bad" : "neutral"}>
          {loadState.status}
        </StatusPill>
        <span className="endpoint">{loadState.endpoint}</span>
        {loadState.status === "ready" ? <span>Loaded {formatTimestamp(loadState.loadedAt)}</span> : null}
      </section>

      {loadState.status === "error" ? (
        <section className="notice" role="alert">
          <h2>Platform service unavailable</h2>
          <p>{loadState.message}</p>
        </section>
      ) : null}

      <section className="dashboard-grid">
        <Panel title="Service Status">
          {viewModel ? (
            <dl className="metric-list">
              <Metric label="Service" value={viewModel.serviceStatus.serviceName} />
              <Metric label="Version" value={viewModel.serviceStatus.version} />
              <Metric label="Stage" value={viewModel.serviceStatus.stage} />
              <Metric label="Mode" value={viewModel.serviceStatus.mode} />
              <Metric
                label="Console health"
                value={viewModel.serviceStatus.healthyForLocalConsole ? "ready" : "not ready"}
              />
            </dl>
          ) : (
            <SkeletonRows count={5} />
          )}
        </Panel>

        <Panel title="Model Inventory">
          {viewModel ? (
            <>
              <div className="summary-row">
                <SummaryItem label="Models" value={viewModel.modelInventory.modelCount} />
                <SummaryItem label="Providers" value={viewModel.modelInventory.providerCount} />
                <SummaryItem label="Profiles" value={viewModel.modelInventory.profileCount} />
              </div>
              <dl className="metric-list">
                <Metric label="Status" value={viewModel.modelInventory.status} />
                <Metric label="Default provider" value={viewModel.modelInventory.defaultProvider || "unset"} />
                <Metric label="Default profile" value={viewModel.modelInventory.defaultProfile || "unset"} />
                <Metric label="Default model" value={viewModel.modelInventory.defaultModel || "unset"} />
              </dl>
              <TokenList items={viewModel.modelInventory.selectableModelIds} emptyLabel="No selectable models" />
            </>
          ) : (
            <SkeletonRows count={4} />
          )}
        </Panel>

        <Panel title="Session And Tooling">
          {viewModel ? (
            <>
              <div className="summary-row">
                <SummaryItem label="Tools" value={viewModel.sessionTooling.toolCount} />
                <SummaryItem label="Action" value={viewModel.sessionTooling.actionStatusLabel} />
              </div>
              <dl className="metric-list">
                <Metric label="Session metadata" value={viewModel.sessionTooling.sessionMetadataRoute} />
                <Metric label="Tools metadata" value={viewModel.sessionTooling.toolsMetadataRoute} />
                <Metric label="Blocked action" value={viewModel.sessionTooling.blockedActionRoute} />
                <Metric label="Execution" value={viewModel.sessionTooling.executionEnabled ? "enabled" : "disabled"} />
              </dl>
            </>
          ) : (
            <SkeletonRows count={4} />
          )}
        </Panel>

        <Panel title="Stop Lines">
          {viewModel ? (
            <>
              <div className="summary-row">
                <SummaryItem
                  label="Enforced"
                  value={viewModel.stopLines.allStopLinesEnforced ? "yes" : "no"}
                />
                <SummaryItem label="Blocked" value={stopLineItems.length} />
              </div>
              <ul className="stop-line-list">
                {stopLineItems.map((item) => (
                  <li key={item.id}>
                    <span>{item.label}</span>
                    <StatusPill tone="neutral">disabled</StatusPill>
                  </li>
                ))}
              </ul>
            </>
          ) : (
            <SkeletonRows count={6} />
          )}
        </Panel>
      </section>

      {overview ? (
        <section className="audit-band">
          <div>
            <h2>Audit Boundary</h2>
            <p>{overview.audit.notes.join(" ")}</p>
          </div>
          <StatusPill tone={overview.audit.writes_business_truth ? "bad" : "good"}>
            {overview.audit.writes_business_truth ? "write enabled" : "advisory only"}
          </StatusPill>
        </section>
      ) : null}
    </main>
  );
}

function Panel({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="panel">
      <h2>{title}</h2>
      {children}
    </section>
  );
}

function Metric({ label, value }: { label: string; value: string | number | boolean }) {
  return (
    <>
      <dt>{label}</dt>
      <dd>{String(value)}</dd>
    </>
  );
}

function SummaryItem({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="summary-item">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function StatusPill({ tone, children }: { tone: "good" | "bad" | "neutral"; children: ReactNode }) {
  return <span className={`status-pill status-pill-${tone}`}>{children}</span>;
}

function TokenList({ items, emptyLabel }: { items: string[]; emptyLabel: string }) {
  if (items.length === 0) {
    return <p className="empty-state">{emptyLabel}</p>;
  }

  return (
    <ul className="token-list">
      {items.map((item) => (
        <li key={item}>{item}</li>
      ))}
    </ul>
  );
}

function SkeletonRows({ count }: { count: number }) {
  return (
    <div className="skeleton-stack" aria-hidden="true">
      {Array.from({ length: count }).map((_, index) => (
        <span key={index} className="skeleton-row" />
      ))}
    </div>
  );
}

function formatTimestamp(value: string): string {
  return new Intl.DateTimeFormat(undefined, {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  }).format(new Date(value));
}
