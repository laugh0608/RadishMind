import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { ReactNode } from "react";
import {
  DEFAULT_PLATFORM_BASE_URL,
  buildPlatformLocalSmokeEndpoint,
  buildPlatformOverviewEndpoint,
  getPlatformOverviewFailureSurface,
  getPlatformOverviewDiagnostics,
  loadPlatformOverview,
  resolvePlatformBaseUrl,
  type PlatformOverviewLoadState,
  type PlatformOverviewReadyState,
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

const STOP_LINE_DETAILS: Record<string, { reason: string; evidence: string }> = {
  real_executor_enabled: {
    reason: "Tool execution is intentionally blocked until executor and confirmation gates are ready.",
    evidence: "Session/tooling action status stays blocked and local-smoke confirms no side effects.",
  },
  durable_store_enabled: {
    reason: "Durable session, checkpoint, audit, and result storage are not implemented in P3.",
    evidence: "Overview and local-smoke both report stop-lines enforced before any store is exposed.",
  },
  confirmation_flow_connected: {
    reason: "Upper-layer confirmation is still a future path, so the console cannot approve actions.",
    evidence: "Confirmation path remains future_upper_layer_confirmation_flow.",
  },
  materialized_result_reader: {
    reason: "Tool result materialization is not available from the metadata-only shell.",
    evidence: "Session/tooling routes expose metadata and blocked responses only.",
  },
  long_term_memory_enabled: {
    reason: "Long-term memory is outside the current local product shell.",
    evidence: "No durable memory route or replay source is exposed in overview.",
  },
  business_truth_write_enabled: {
    reason: "RadishMind remains advisory-only and must not write upstream truth sources.",
    evidence: "Audit boundary reports advisory only and writes_business_truth=false.",
  },
  automatic_replay_enabled: {
    reason: "Automatic replay is blocked until durable execution and recovery policies exist.",
    evidence: "Overview stop-lines and local-smoke readiness keep replay disabled.",
  },
  production_secret_backend_ready: {
    reason: "Production secret handling is not part of the local development shell.",
    evidence: "P3 short-close still marks production secret backend as not satisfied.",
  },
};

const initialBaseUrl = resolvePlatformBaseUrl();
const DEV_LAUNCHER_COMMANDS = [
  "pwsh ./scripts/run-radishmind-console-dev.ps1 -VerifyOnly",
  "./scripts/run-radishmind-console-dev.sh --verify-only",
];
const DEV_DIAGNOSTIC_HINTS = [
  "Port conflict: keep backend on 7000 and frontend on 4000, or stop the other local process.",
  "CORS / preflight: platform only allows http://127.0.0.1:4000 and http://localhost:4000.",
  "Unsafe port: ERR_UNSAFE_PORT means the browser blocked the port before RadishMind received the request.",
  "Contract mismatch: run the overview consumer smoke against the configured Platform URL.",
  "Local-smoke mismatch: run the local-smoke readiness check against the configured Platform URL.",
];

type ModelInventoryDetail = {
  id: string;
  kind: "configured_default" | "provider_profile" | "provider_registry" | "unknown";
  provider: string;
  profile: string;
};

type StopLineItem = {
  id: string;
  label: string;
};

export function App() {
  const [baseUrl, setBaseUrl] = useState(initialBaseUrl);
  const [loadState, setLoadState] = useState<PlatformOverviewLoadState>({
    status: "idle",
    endpoint: buildPlatformOverviewEndpoint(initialBaseUrl),
  });
  const lastReadyState = useRef<PlatformOverviewReadyState | null>(null);

  const refreshOverview = useCallback(async () => {
    const endpoint = buildPlatformOverviewEndpoint(baseUrl);
    const previousReadyState = lastReadyState.current ?? undefined;
    setLoadState({ status: "loading", endpoint, previous: previousReadyState });
    try {
      const readyState = await loadPlatformOverview(baseUrl);
      lastReadyState.current = readyState;
      setLoadState(readyState);
    } catch (error) {
      setLoadState({
        status: "error",
        endpoint,
        message: error instanceof Error ? error.message : "overview request failed",
        failureSurface: getPlatformOverviewFailureSurface(error),
        diagnostics: getPlatformOverviewDiagnostics(error),
        previous: previousReadyState,
      });
    }
  }, [baseUrl]);

  useEffect(() => {
    void refreshOverview();
  }, [refreshOverview]);

  const readyState = latestReadyState(loadState);
  const viewModel = readyState?.viewModel ?? null;
  const readinessViewModel = readyState?.readinessViewModel ?? null;
  const overview = readyState?.overview ?? null;
  const localSmoke = readyState?.localSmoke ?? null;
  const localSmokeEndpoint = readyState?.localSmokeEndpoint ?? buildPlatformLocalSmokeEndpoint(baseUrl);
  const showingStaleOverview = loadState.status === "error" && readyState !== null;
  const localSmokeFailure = loadState.status === "error" && loadState.failureSurface === "platform_local_smoke";
  const stopLineItems = useMemo(
    () =>
      viewModel?.stopLines.blockedCapabilityIds.map((capabilityId) => ({
        id: capabilityId,
        label: STOP_LINE_LABELS[capabilityId] ?? capabilityId,
      })) ?? [],
    [viewModel],
  );
  const modelInventoryDetails = useMemo(
    () =>
      viewModel
        ? buildModelInventoryDetails(
            viewModel.modelInventory.selectableModelIds,
            viewModel.modelInventory.defaultProvider,
            viewModel.modelInventory.defaultProfile,
            viewModel.modelInventory.defaultModel,
          )
        : [],
    [viewModel],
  );

  return (
    <main className="ops-shell">
      <NavigationRail loadState={loadState} readyState={readyState} />
      <section className="ops-workspace" aria-label="RadishMind Console">
        <ConsoleHeader
          baseUrl={baseUrl}
          loadState={loadState}
          onBaseUrlChange={setBaseUrl}
          onRefresh={refreshOverview}
        />
        <GlobalStatusBar
          loadState={loadState}
          readyState={readyState}
          localSmokeEndpoint={localSmokeEndpoint}
          showingStaleOverview={showingStaleOverview}
        />

        {loadState.status === "error" ? (
          <FailureNotice loadState={loadState} localSmokeFailure={localSmokeFailure} />
        ) : null}

        <OverviewMetrics
          loadState={loadState}
          viewModel={viewModel}
          readinessViewModel={readinessViewModel}
          stopLineItems={stopLineItems}
        />

        <div className="ops-surface-grid">
          <div className="ops-primary-column">
            <ServiceStatusPanel viewModel={viewModel} />
            <ProviderProfilePanel viewModel={viewModel} modelInventoryDetails={modelInventoryDetails} />
            <SessionToolingPanel viewModel={viewModel} readinessViewModel={readinessViewModel} />
            <DevDiagnosticsPanel
              baseUrl={baseUrl}
              loadState={loadState}
              readyState={readyState}
              readinessViewModel={readinessViewModel}
              localSmokeEndpoint={localSmokeEndpoint}
            />
          </div>

          <aside className="ops-secondary-column" aria-label="Local readiness and stop-line details">
            <Panel title="Local Readiness" id="readiness" className="panel-priority">
              <LocalReadinessContent readinessViewModel={readinessViewModel} localSmoke={localSmoke} />
            </Panel>
            <Panel title="Stop Lines" className="panel-priority">
              <StopLinesContent
                stopLineItems={stopLineItems}
                viewModel={viewModel}
                readinessViewModel={readinessViewModel}
                overview={overview}
              />
            </Panel>
            {overview ? <AuditBoundary overview={overview} /> : <EmptyPanel title="Audit Boundary" count={3} />}
          </aside>
        </div>
      </section>
    </main>
  );
}

function NavigationRail({
  loadState,
  readyState,
}: {
  loadState: PlatformOverviewLoadState;
  readyState: PlatformOverviewReadyState | null;
}) {
  return (
    <aside className="navigation-rail" aria-label="Console navigation">
      <div>
        <p className="eyebrow">RadishMind</p>
        <h1>Console</h1>
        <p className="rail-summary">Local ops surface for runtime status, readiness, diagnostics, and stop-lines.</p>
      </div>
      <nav className="rail-nav" aria-label="Ops surface sections">
        <a href="#overview" className="active">
          Overview
        </a>
        <a href="#readiness">Local Readiness</a>
        <a href="#provider-profile">Provider/Profile</a>
        <a href="#session-tooling">Session & Tooling</a>
        <a href="#stop-lines">Stop-line Details</a>
      </nav>
      <div className="rail-note">
        <StatusPill tone={loadState.status === "ready" ? "good" : loadState.status === "error" ? "bad" : "neutral"}>
          {loadState.status === "error" && readyState ? "stale" : loadState.status}
        </StatusPill>
        <p>Read-only shell. Executor, confirmation, writeback, durable store, and replay stay locked.</p>
      </div>
    </aside>
  );
}

function ConsoleHeader({
  baseUrl,
  loadState,
  onBaseUrlChange,
  onRefresh,
}: {
  baseUrl: string;
  loadState: PlatformOverviewLoadState;
  onBaseUrlChange: (value: string) => void;
  onRefresh: () => void;
}) {
  return (
    <header className="workspace-header" id="overview">
      <div>
        <p className="eyebrow">P3 Local Product Shell / Ops Surface</p>
        <h2>Runtime overview</h2>
        <p>Read platform overview and local-smoke readiness without enabling any execution path.</p>
      </div>
      <div className="connection-controls">
        <label className="base-url-field">
          <span>Platform URL</span>
          <input
            value={baseUrl}
            onChange={(event) => onBaseUrlChange(event.target.value)}
            placeholder={DEFAULT_PLATFORM_BASE_URL}
          />
        </label>
        <button type="button" onClick={onRefresh} disabled={loadState.status === "loading"}>
          {loadState.status === "loading" ? "Refreshing" : "Refresh"}
        </button>
      </div>
    </header>
  );
}

function GlobalStatusBar({
  loadState,
  readyState,
  localSmokeEndpoint,
  showingStaleOverview,
}: {
  loadState: PlatformOverviewLoadState;
  readyState: PlatformOverviewReadyState | null;
  localSmokeEndpoint: string;
  showingStaleOverview: boolean;
}) {
  return (
    <section className="status-strip" aria-live="polite">
      <StatusPill tone={loadState.status === "ready" ? "good" : loadState.status === "error" ? "bad" : "neutral"}>
        {loadState.status}
      </StatusPill>
      <span className="endpoint">{loadState.endpoint}</span>
      <span className="endpoint">{localSmokeEndpoint}</span>
      {readyState ? <span>Loaded {formatTimestamp(readyState.loadedAt)}</span> : null}
      {loadState.status === "loading" && readyState ? <span>Refreshing; showing last overview</span> : null}
      {showingStaleOverview ? <span>Connection failed; showing last overview</span> : null}
    </section>
  );
}

function FailureNotice({
  loadState,
  localSmokeFailure,
}: {
  loadState: Extract<PlatformOverviewLoadState, { status: "error" }>;
  localSmokeFailure: boolean;
}) {
  return (
    <section className="notice" role="alert">
      <h2>{localSmokeFailure ? "Local-smoke readiness unavailable" : "Platform service unavailable"}</h2>
      {localSmokeFailure ? (
        <p>
          Overview was readable; local-smoke readiness failed, so the console keeps the last read-only view when
          available.
        </p>
      ) : null}
      <p>{loadState.message}</p>
      <ul className="diagnostic-list">
        {loadState.diagnostics.map((diagnostic) => (
          <li key={diagnostic}>{diagnostic}</li>
        ))}
      </ul>
    </section>
  );
}

function OverviewMetrics({
  loadState,
  viewModel,
  readinessViewModel,
  stopLineItems,
}: {
  loadState: PlatformOverviewLoadState;
  viewModel: PlatformOverviewReadyState["viewModel"] | null;
  readinessViewModel: PlatformOverviewReadyState["readinessViewModel"] | null;
  stopLineItems: StopLineItem[];
}) {
  return (
    <section className="metric-card-grid" aria-label="Readiness summary">
      <MetricCard label="Service" value={viewModel?.serviceStatus.status ?? "unknown"} detail="local read-only shell" />
      <MetricCard
        label="Boundary"
        value={viewModel?.sessionTooling.actionStatusLabel ?? "blocked"}
        detail={readinessViewModel?.blockedActionNoSideEffects ? "no side effects" : "metadata only"}
        tone="warning"
      />
      <MetricCard
        label="Readiness"
        value={readinessViewModel?.status ?? loadState.status}
        detail={readinessViewModel?.localConsoleReady ? "console ready" : "not ready"}
      />
      <MetricCard label="Stop lines" value={`${stopLineItems.length} held`} detail="executor, store, replay locked" />
    </section>
  );
}

function ServiceStatusPanel({ viewModel }: { viewModel: PlatformOverviewReadyState["viewModel"] | null }) {
  return (
    <Panel title="Service Status">
      {viewModel ? (
        <dl className="metric-list">
          <Metric label="Service" value={viewModel.serviceStatus.serviceName} />
          <Metric label="Version" value={viewModel.serviceStatus.version} />
          <Metric label="Stage" value={viewModel.serviceStatus.stage} />
          <Metric label="Mode" value={viewModel.serviceStatus.mode} />
          <Metric label="Overview route" value={viewModel.serviceStatus.overviewRoute} />
          <Metric label="Console health" value={viewModel.serviceStatus.healthyForLocalConsole ? "ready" : "not ready"} />
        </dl>
      ) : (
        <SkeletonRows count={5} />
      )}
    </Panel>
  );
}

function ProviderProfilePanel({
  viewModel,
  modelInventoryDetails,
}: {
  viewModel: PlatformOverviewReadyState["viewModel"] | null;
  modelInventoryDetails: ModelInventoryDetail[];
}) {
  return (
    <Panel title="Model Inventory" id="provider-profile">
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
          {viewModel.modelInventory.status !== "ok" ? (
            <p className="inline-warning">Provider inventory is unavailable; check platform diagnostics.</p>
          ) : null}
          <p className="section-label">Selectable model IDs</p>
          <TokenList items={viewModel.modelInventory.selectableModelIds} emptyLabel="No selectable models" />
          <p className="section-label">Active profile chain</p>
          <TokenList items={viewModel.modelInventory.activeProfileChain} emptyLabel="No active profile chain" />
          <div className="inventory-details" aria-label="Provider/Profile Details">
            <p className="section-label">Provider/Profile Details</p>
            <dl className="compact-metric-list">
              <Metric label="Inventory kind" value={viewModel.modelInventory.inventoryKind} />
              <Metric label="Models route" value={viewModel.modelInventory.modelsRoute} />
              <Metric label="Detail route" value={viewModel.modelInventory.detailRoute} />
              <Metric
                label="Selector boundary"
                value={
                  viewModel.modelInventory.canShowProfileSelector
                    ? "display only; no health check or credential readiness"
                    : "hidden until inventory is readable"
                }
              />
            </dl>
            <ul className="inventory-detail-list">
              {modelInventoryDetails.map((item) => (
                <li key={item.id}>
                  <div>
                    <strong>{item.id}</strong>
                    <dl className="compact-metric-list nested">
                      <Metric label="Source" value={formatModelInventoryKind(item.kind)} />
                      <Metric label="Provider" value={item.provider} />
                      <Metric label="Profile" value={item.profile} />
                    </dl>
                  </div>
                  <StatusPill tone="neutral">read only</StatusPill>
                </li>
              ))}
            </ul>
          </div>
        </>
      ) : (
        <SkeletonRows count={4} />
      )}
    </Panel>
  );
}

function SessionToolingPanel({
  viewModel,
  readinessViewModel,
}: {
  viewModel: PlatformOverviewReadyState["viewModel"] | null;
  readinessViewModel: PlatformOverviewReadyState["readinessViewModel"] | null;
}) {
  return (
    <Panel title="Session And Tooling" id="session-tooling">
      {viewModel ? (
        <>
          <div className="summary-row">
            <SummaryItem label="Tools" value={viewModel.sessionTooling.toolCount} />
            <SummaryItem label="Action" value={viewModel.sessionTooling.actionStatusLabel} />
            <SummaryItem label="Metadata" value={viewModel.sessionTooling.metadataOnly ? "only" : "unknown"} />
          </div>
          <dl className="metric-list">
            <Metric label="Session metadata" value={viewModel.sessionTooling.sessionMetadataRoute} />
            <Metric label="Tools metadata" value={viewModel.sessionTooling.toolsMetadataRoute} />
            <Metric label="Blocked action" value={viewModel.sessionTooling.blockedActionRoute} />
            <Metric label="Execution" value={viewModel.sessionTooling.executionEnabled ? "enabled" : "disabled"} />
            <Metric label="Confirmation path" value={viewModel.sessionTooling.requiresConfirmationPath} />
          </dl>
          <div className="blocked-action-detail" aria-label="Blocked Action Detail">
            <div>
              <p className="section-label">Blocked Action Detail</p>
              <p>
                The current `POST /v1/tools/actions` contract is represented as a blocked, metadata-only response. No
                execute, confirm, writeback, apply, or replay control is rendered here.
              </p>
            </div>
            <StatusPill tone={readinessViewModel?.blockedActionNoSideEffects ? "neutral" : "bad"}>
              {readinessViewModel?.blockedActionNoSideEffects ? "blocked" : "unknown"}
            </StatusPill>
          </div>
        </>
      ) : (
        <SkeletonRows count={4} />
      )}
    </Panel>
  );
}

function LocalReadinessContent({
  readinessViewModel,
  localSmoke,
}: {
  readinessViewModel: PlatformOverviewReadyState["readinessViewModel"] | null;
  localSmoke: PlatformOverviewReadyState["localSmoke"] | null;
}) {
  if (!readinessViewModel || !localSmoke) {
    return <SkeletonRows count={5} />;
  }

  return (
    <>
      <div className="summary-row sidebar-summary">
        <SummaryItem label="Readiness" value={readinessViewModel.status} />
        <SummaryItem label="Console" value={readinessViewModel.localConsoleReady ? "ready" : "not ready"} />
        <SummaryItem label="Mode" value={readinessViewModel.readOnly ? "read only" : "write capable"} />
      </div>
      <ul className="readiness-list">
        <ReadinessItem label="Healthz" ready={readinessViewModel.healthzOk} />
        <ReadinessItem label="Overview contract" ready={readinessViewModel.overviewContractReadable} />
        <ReadinessItem label="Model inventory" ready={readinessViewModel.modelInventoryReadable} />
        <ReadinessItem label="Session/tooling metadata" ready={readinessViewModel.sessionToolingMetadataReadable} />
        <ReadinessItem label="Blocked action no side effects" ready={readinessViewModel.blockedActionNoSideEffects} />
        <ReadinessItem label="Stop lines enforced" ready={readinessViewModel.allStopLinesEnforced} />
      </ul>
      <p className="section-label">Local CORS origins</p>
      <TokenList items={readinessViewModel.allowedCorsOrigins} emptyLabel="No local origins declared" />
      <p className="section-label">Failure hints</p>
      <TokenList
        items={localSmoke.failure_hints.map((hint) => `${hint.code}: ${hint.message}`)}
        emptyLabel="No failure hints"
      />
    </>
  );
}

function StopLinesContent({
  stopLineItems,
  viewModel,
  readinessViewModel,
  overview,
}: {
  stopLineItems: StopLineItem[];
  viewModel: PlatformOverviewReadyState["viewModel"] | null;
  readinessViewModel: PlatformOverviewReadyState["readinessViewModel"] | null;
  overview: PlatformOverviewReadyState["overview"] | null;
}) {
  if (!viewModel) {
    return <SkeletonRows count={6} />;
  }

  return (
    <>
      <div className="summary-row sidebar-summary">
        <SummaryItem label="Enforced" value={viewModel.stopLines.allStopLinesEnforced ? "yes" : "no"} />
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
      <div className="stop-line-details" aria-label="Stop-line Details" id="stop-lines">
        <p className="section-label">Stop-line Details</p>
        <dl className="stop-line-evidence">
          <Metric label="Overview enforcement" value={viewModel.stopLines.allStopLinesEnforced ? "enforced" : "not enforced"} />
          <Metric label="Local-smoke enforcement" value={readinessViewModel?.allStopLinesEnforced ? "enforced" : "unknown"} />
          <Metric label="Blocked action" value={readinessViewModel?.blockedActionNoSideEffects ? "no side effects" : "unknown"} />
          <Metric label="Audit mode" value={overview?.audit.advisory_only ? "advisory only" : "unknown"} />
        </dl>
        <ul className="stop-line-detail-list">
          {stopLineItems.map((item) => {
            const details = STOP_LINE_DETAILS[item.id] ?? {
              reason: "Capability remains disabled by the current platform stop-lines.",
              evidence: "Overview stop-lines report this capability as blocked.",
            };
            return (
              <li key={`${item.id}-details`}>
                <div>
                  <strong>{item.label}</strong>
                  <p>{details.reason}</p>
                  <p className="evidence-note">{details.evidence}</p>
                </div>
                <StatusPill tone="neutral">blocked</StatusPill>
              </li>
            );
          })}
        </ul>
      </div>
    </>
  );
}

function DevDiagnosticsPanel({
  baseUrl,
  loadState,
  readyState,
  readinessViewModel,
  localSmokeEndpoint,
}: {
  baseUrl: string;
  loadState: PlatformOverviewLoadState;
  readyState: PlatformOverviewReadyState | null;
  readinessViewModel: PlatformOverviewReadyState["readinessViewModel"] | null;
  localSmokeEndpoint: string;
}) {
  return (
    <Panel title="Dev Diagnostics" className="diagnostics-band">
      <dl className="diagnostics-grid">
        <Metric label="Platform URL" value={baseUrl} />
        <Metric label="Overview endpoint" value={loadState.endpoint} />
        <Metric label="Local-smoke endpoint" value={localSmokeEndpoint} />
        <Metric label="Load status" value={loadState.status} />
        <Metric
          label="Failure surface"
          value={loadState.status === "error" ? formatFailureSurface(loadState.failureSurface) : "none"}
        />
        <Metric label="Last loaded" value={readyState ? formatTimestamp(readyState.loadedAt) : "not loaded"} />
        <Metric label="Service status" value={readyState?.viewModel.serviceStatus.status ?? "unknown"} />
        <Metric label="Console connection" value={readinessViewModel?.localConsoleReady ? "ready" : "not ready"} />
        <Metric label="Readiness status" value={readinessViewModel?.status ?? "unknown"} />
      </dl>
      <div className="diagnostics-columns">
        <div>
          <p className="section-label">Local probes</p>
          <ul className="command-list">
            {DEV_LAUNCHER_COMMANDS.map((command) => (
              <li key={command}>{command}</li>
            ))}
          </ul>
        </div>
        <div>
          <p className="section-label">Failure classes</p>
          <ul className="diagnostic-list compact-list">
            {DEV_DIAGNOSTIC_HINTS.map((hint) => (
              <li key={hint}>{hint}</li>
            ))}
          </ul>
        </div>
      </div>
    </Panel>
  );
}

function AuditBoundary({ overview }: { overview: PlatformOverviewReadyState["overview"] }) {
  return (
    <section className="audit-band">
      <div>
        <h2>Audit Boundary</h2>
        <ul className="audit-notes">
          {overview.audit.notes.map((note) => (
            <li key={note}>{note}</li>
          ))}
        </ul>
      </div>
      <StatusPill tone={overview.audit.writes_business_truth ? "bad" : "good"}>
        {overview.audit.writes_business_truth ? "write enabled" : "advisory only"}
      </StatusPill>
    </section>
  );
}

function EmptyPanel({ title, count }: { title: string; count: number }) {
  return (
    <Panel title={title}>
      <SkeletonRows count={count} />
    </Panel>
  );
}

function latestReadyState(loadState: PlatformOverviewLoadState): PlatformOverviewReadyState | null {
  if (loadState.status === "ready") {
    return loadState;
  }
  if (loadState.status === "loading" || loadState.status === "error") {
    return loadState.previous ?? null;
  }
  return null;
}

function Panel({
  title,
  children,
  id,
  className,
}: {
  title: string;
  children: ReactNode;
  id?: string;
  className?: string;
}) {
  return (
    <section className={["panel", className].filter(Boolean).join(" ")} id={id}>
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

function MetricCard({
  label,
  value,
  detail,
  tone = "default",
}: {
  label: string;
  value: string | number;
  detail: string;
  tone?: "default" | "warning";
}) {
  return (
    <section className={`metric-card metric-card-${tone}`}>
      <span>{label}</span>
      <strong>{value}</strong>
      <p>{detail}</p>
    </section>
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

function ReadinessItem({ label, ready }: { label: string; ready: boolean }) {
  return (
    <li>
      <span>{label}</span>
      <StatusPill tone={ready ? "good" : "bad"}>{ready ? "ready" : "not ready"}</StatusPill>
    </li>
  );
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

function buildModelInventoryDetails(
  selectableModelIds: string[],
  defaultProvider: string,
  defaultProfile: string,
  defaultModel: string,
): ModelInventoryDetail[] {
  return selectableModelIds.map((id) => parseSelectableModelId(id, defaultProvider, defaultProfile, defaultModel));
}

function parseSelectableModelId(
  id: string,
  defaultProvider: string,
  defaultProfile: string,
  defaultModel: string,
): ModelInventoryDetail {
  if (id === defaultModel) {
    return {
      id,
      kind: "configured_default",
      provider: defaultProvider || "unset",
      profile: defaultProfile || "unset",
    };
  }
  if (id.startsWith("provider:") && id.includes(":profile:")) {
    const [provider, profile] = id.replace(/^provider:/, "").split(":profile:", 2);
    return {
      id,
      kind: "provider_profile",
      provider: provider || "unset",
      profile: profile || "default",
    };
  }
  if (id.startsWith("profile:")) {
    return {
      id,
      kind: "provider_profile",
      provider: "openai-compatible",
      profile: id.replace(/^profile:/, "") || "default",
    };
  }
  if (id.startsWith("provider:")) {
    return {
      id,
      kind: "provider_registry",
      provider: id.replace(/^provider:/, "") || "unset",
      profile: "not profile-specific",
    };
  }
  return {
    id,
    kind: "unknown",
    provider: defaultProvider || "unset",
    profile: defaultProfile || "unset",
  };
}

function formatModelInventoryKind(kind: ModelInventoryDetail["kind"]): string {
  if (kind === "configured_default") {
    return "configured default";
  }
  if (kind === "provider_profile") {
    return "provider profile";
  }
  if (kind === "provider_registry") {
    return "provider registry";
  }
  return "unknown";
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

function formatFailureSurface(value: string): string {
  if (value === "platform_local_smoke") {
    return "local-smoke readiness";
  }
  if (value === "platform_overview") {
    return "platform overview";
  }
  return "unknown";
}
