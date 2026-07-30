import type { ApplicationDevelopmentWorkspaceContext } from "./applicationDevelopmentWorkspace";
import type {
  ControlPlaneReadDevLiveConfig,
  ControlPlaneReadDevLiveLoadState,
} from "./devLiveReadConsumer";
import type { WorkspaceOperationsInboxViewModel } from "./workspaceOperationsInbox";

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
  const sourceLabel = sourceConfig.mode === "dev_live_http" ? "Development / test" : "Offline fixtures";
  const sourceTone = sourceState.status === "failed"
    ? "bad"
    : sourceState.status === "ready" || sourceState.status === "idle"
      ? "good"
      : "neutral";

  return (
    <section className="workspace-product-overview" aria-label="Workspace overview">
      <article className="workspace-source-pulse">
        <div className="workspace-overview-card-heading">
          <h3>Source coverage</h3>
          <span className={`status-badge ${sourceTone}`}>{sourceState.status}</span>
        </div>
        <div className="workspace-source-pulse-list">
          {inbox.coverage.map((coverage) => (
            <div key={coverage.sourceId}>
              <span
                className={`workspace-source-pulse-dot ${coverage.status}`}
                aria-label={coverage.status.replaceAll("_", " ")}
              />
              <span>{coverageLabel(coverage.sourceId)}</span>
              <strong>{coverage.itemCount}</strong>
            </div>
          ))}
        </div>
        <dl className="workspace-source-context">
          <div><dt>Source</dt><dd>{sourceLabel}</dd></div>
          <div><dt>Workspace</dt><dd>{inbox.activeWorkspaceId}</dd></div>
          <div><dt>Snapshot</dt><dd>{formatDateTime(inbox.referenceTime)}</dd></div>
        </dl>
      </article>

      <article className="workspace-active-application">
        <div className="workspace-overview-card-heading">
          <h3>Active application</h3>
          <span className={`status-badge ${application.status === "active" ? "good" : "neutral"}`}>
            {application.status}
          </span>
        </div>
        <strong className="workspace-active-application-name">{application.displayName}</strong>
        {application.status !== "unavailable" ? (
          <>
            <dl className="workspace-active-application-meta">
              <div><dt>Application</dt><dd>{application.applicationId}</dd></div>
              <div><dt>Kind</dt><dd>{application.applicationKind}</dd></div>
              <div><dt>Version</dt><dd>{application.recordVersion || "—"}</dd></div>
              <div><dt>Updated</dt><dd>{formatDateTime(application.updatedAt)}</dd></div>
            </dl>
            <a href="#application-development-workspace">Open application <span aria-hidden="true">→</span></a>
          </>
        ) : (
          <a href="#workspace-applications">Select application <span aria-hidden="true">→</span></a>
        )}
      </article>
    </section>
  );
}

function coverageLabel(sourceId: WorkspaceOperationsInboxViewModel["coverage"][number]["sourceId"]): string {
  switch (sourceId) {
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
