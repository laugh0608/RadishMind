import type { ApplicationDevelopmentWorkspaceContext } from "./applicationDevelopmentWorkspace";
import type {
  ControlPlaneReadDevLiveConfig,
  ControlPlaneReadDevLiveLoadState,
} from "./devLiveReadConsumer";
import type {
  WorkspaceOperationsInboxCoverage,
  WorkspaceOperationsInboxCoverageStatus,
  WorkspaceOperationsInboxViewModel,
} from "./workspaceOperationsInbox";

export function WorkspaceProductOverviewPanel({
  application,
  inbox,
  sourceConfig,
  sourceState,
}: {
  application: ApplicationDevelopmentWorkspaceContext;
  inbox: WorkspaceOperationsInboxViewModel;
  sourceConfig: ControlPlaneReadDevLiveConfig;
  sourceState: ControlPlaneReadDevLiveLoadState;
}) {
  const sourceModeLabel = sourceConfig.mode === "dev_live_http" ? "Development / test" : "Offline fixtures";
  const sourceTone = sourceState.status === "failed"
    ? "bad"
    : sourceState.status === "ready" || sourceState.status === "idle"
      ? "good"
      : "neutral";
  const coverageCounts = summarizeCoverage(inbox.coverage);

  return (
    <section
      className="workspace-product-overview"
      id="workspace-overview"
      aria-labelledby="workspace-overview-title"
    >
      <header className="workspace-overview-hero">
        <div className="workspace-overview-title">
          <h2 id="workspace-overview-title">Workspace</h2>
        </div>

        <article className="workspace-active-application" aria-label="Active application">
          <span className="workspace-active-application-icon" aria-hidden="true">AI</span>
          <span className="workspace-active-application-copy">
            <small>Active application</small>
            <strong>{application.displayName || "No application selected"}</strong>
            <span>
              {application.status === "unavailable"
                ? "Application context unavailable"
                : `${application.applicationKind} · revision ${application.recordVersion || "unavailable"}`}
            </span>
          </span>
          <a
            href={application.status === "unavailable" ? "#workspace-applications" : "#application-development-workspace"}
            aria-label={application.status === "unavailable" ? "Select application" : `Open ${application.displayName}`}
          >
            <span aria-hidden="true">↗</span>
          </a>
        </article>
      </header>

      <div className="workspace-overview-grid">
        <article className="workspace-source-pulse">
          <div className="workspace-overview-card-heading">
            <h3>Workspace pulse</h3>
            <span className={`status-badge ${inbox.status === "ready" ? "good" : inbox.status === "blocked" ? "bad" : "neutral"}`}>
              {statusLabel(inbox.status)}
            </span>
          </div>

          <div className="workspace-pulse-visual">
            <div className="workspace-pulse-total">
              <strong>{String(inbox.items.length).padStart(2, "0")}</strong>
              <span>attention items</span>
              <small>within the loaded source window</small>
            </div>
            <div className="workspace-pulse-bars" aria-label="Loaded source distribution">
              {inbox.coverage.map((coverage) => (
                <div
                  key={coverage.sourceId}
                  className={`workspace-pulse-bar ${coverage.status} load-${Math.min(4, coverage.itemCount)}`}
                  title={`${coverage.label}: ${coverage.itemCount} items, ${coverageStatusLabel(coverage.status)}`}
                >
                  <span aria-hidden="true" />
                  <small>{matrixCoverageLabel(coverage)}</small>
                </div>
              ))}
            </div>
          </div>

          <dl className="workspace-pulse-metrics">
            <div><dt>Ready sources</dt><dd>{coverageCounts.ready}</dd></div>
            <div><dt>Partial sources</dt><dd>{coverageCounts.partial}</dd></div>
            <div><dt>Unavailable</dt><dd>{coverageCounts.blocked}</dd></div>
          </dl>
        </article>

        <article className="workspace-source-evidence">
          <div className="workspace-overview-card-heading">
            <h3>Source evidence</h3>
            <span className={`status-badge ${sourceTone}`}>{sourceState.status}</span>
          </div>

          <div className="workspace-evidence-distribution" aria-label="Source evidence distribution">
            <div className="ready"><strong>{coverageCounts.ready}</strong><span>Ready</span></div>
            <div className="partial"><strong>{coverageCounts.partial}</strong><span>Partial</span></div>
            <div className="blocked"><strong>{coverageCounts.blocked}</strong><span>Blocked</span></div>
          </div>

          <div className="workspace-evidence-matrix">
            {inbox.coverage.map((coverage) => (
              <div key={coverage.sourceId}>
                <span className={`workspace-source-pulse-dot ${coverage.status}`} aria-hidden="true" />
                <span title={coverage.label}>{matrixCoverageLabel(coverage)}</span>
                <strong className={coverage.status}>{coverageStatusLabel(coverage.status)}</strong>
              </div>
            ))}
          </div>

          <dl className="workspace-source-context">
            <div><dt>Source</dt><dd>{sourceModeLabel}</dd></div>
            <div><dt>Workspace</dt><dd>{inbox.activeWorkspaceId}</dd></div>
            <div><dt>Snapshot</dt><dd>{formatDateTime(inbox.referenceTime)}</dd></div>
          </dl>
        </article>
      </div>
    </section>
  );
}

function summarizeCoverage(coverage: WorkspaceOperationsInboxCoverage[]) {
  return coverage.reduce(
    (summary, item) => {
      if (item.status === "complete_window") {
        summary.ready += 1;
      } else if (item.status === "partial_window") {
        summary.partial += 1;
      } else {
        summary.blocked += 1;
      }
      return summary;
    },
    { ready: 0, partial: 0, blocked: 0 },
  );
}

function shortCoverageLabel(coverage: WorkspaceOperationsInboxCoverage): string {
  switch (coverage.sourceId) {
    case "api_keys":
      return "API keys";
    case "workflow_definitions":
      return "Workflows";
    case "runs":
      return "Runs";
    default:
      return "Applications";
  }
}

function matrixCoverageLabel(coverage: WorkspaceOperationsInboxCoverage): string {
  return coverage.sourceId === "applications" ? "Apps" : shortCoverageLabel(coverage);
}

function coverageStatusLabel(status: WorkspaceOperationsInboxCoverageStatus): string {
  switch (status) {
    case "complete_window":
      return "Ready";
    case "partial_window":
      return "Partial";
    default:
      return "Blocked";
  }
}

function statusLabel(status: WorkspaceOperationsInboxViewModel["status"]): string {
  return status.charAt(0).toUpperCase() + status.slice(1);
}

function formatDateTime(value: string): string {
  const timestamp = Date.parse(value);
  if (!Number.isFinite(timestamp)) {
    return value || "Not available";
  }
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(timestamp);
}
