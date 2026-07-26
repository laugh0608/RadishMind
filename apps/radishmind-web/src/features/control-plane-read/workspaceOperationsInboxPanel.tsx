import type {
  WorkspaceOperationsInboxCoverage,
  WorkspaceOperationsInboxItem,
  WorkspaceOperationsInboxSeverity,
  WorkspaceOperationsInboxViewModel,
} from "./workspaceOperationsInbox";

export function WorkspaceOperationsInboxPanel({
  inbox,
  onOpenItem,
}: {
  inbox: WorkspaceOperationsInboxViewModel;
  onOpenItem: (item: WorkspaceOperationsInboxItem) => void;
}) {
  return (
    <section
      className="surface-band workspace-operations-inbox"
      id="workspace-operations-inbox"
      aria-labelledby="workspace-operations-inbox-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">Workspace Operations Inbox</p>
          <h3 id="workspace-operations-inbox-title">Authorized attention queue</h3>
        </div>
        <span className={`status-badge ${inbox.status === "ready" ? "good" : inbox.status === "blocked" ? "bad" : "neutral"}`}>
          {inbox.status}
        </span>
      </div>

      <article className="workspace-operations-inbox-hero">
        <div>
          <p className="eyebrow">{inbox.activeWorkspaceId}</p>
          <h4>Current read windows</h4>
          <p>{inbox.summary}</p>
        </div>
        <dl className="workspace-operations-inbox-meta">
          <div><dt>Reference time</dt><dd>{inbox.referenceTime}</dd></div>
          <div><dt>Mutation</dt><dd>{inbox.canMutate ? "enabled" : "locked"}</dd></div>
          <div><dt>Quota</dt><dd>{inbox.canEnforceQuota ? "enabled" : "not projected"}</dd></div>
          <div><dt>Business truth</dt><dd>{inbox.canWriteBusinessTruth ? "writable" : "read only"}</dd></div>
        </dl>
      </article>

      <div className="workspace-operations-inbox-metrics" aria-label="Workspace operations inbox metrics">
        {inbox.metrics.map((metric) => (
          <article key={metric.label}>
            <span>{metric.label}</span>
            <strong>{metric.value}</strong>
            <p>{metric.detail}</p>
          </article>
        ))}
      </div>

      <div className="workspace-operations-inbox-coverage" aria-label="Workspace operations source coverage">
        {inbox.coverage.map((coverage) => <CoverageCard key={coverage.sourceId} coverage={coverage} />)}
      </div>

      {inbox.items.length > 0 ? (
        <div className="workspace-operations-inbox-items" aria-label="Workspace operations attention items">
          {inbox.items.map((item) => (
            <article key={item.itemId} className={`workspace-operations-inbox-item ${item.severity}`}>
              <div className="workspace-operations-inbox-row">
                <div>
                  <p className="eyebrow">{item.sourceId} · {item.reason}</p>
                  <h4>{item.title}</h4>
                </div>
                <span className={`status-badge ${severityTone(item.severity)}`}>{item.severity}</span>
              </div>
              <p>{item.summary}</p>
              <dl className="workspace-operations-inbox-meta">
                <div><dt>Resource</dt><dd>{item.resourceRef}</dd></div>
                <div><dt>Observed</dt><dd>{item.occurredAt}</dd></div>
                <div><dt>Request</dt><dd>{item.requestId}</dd></div>
                <div><dt>Audit</dt><dd>{item.auditRef}</dd></div>
              </dl>
              <a href={item.targetAnchor} onClick={() => onOpenItem(item)}>
                Open existing review surface
              </a>
            </article>
          ))}
        </div>
      ) : (
        <article className="workspace-operations-inbox-empty">
          <h4>{inbox.currentWindowHasNoAttentionItems ? "No attention items in the current complete windows" : "No resource items available"}</h4>
          <p>
            {inbox.currentWindowHasNoAttentionItems
              ? "This is a current-window statement, not a workspace health or SLA claim."
              : "Review source coverage before interpreting the empty queue."}
          </p>
        </article>
      )}
    </section>
  );
}

function CoverageCard({ coverage }: { coverage: WorkspaceOperationsInboxCoverage }) {
  return (
    <article>
      <div className="workspace-operations-inbox-row">
        <div><span>{coverage.label}</span><strong>{coverage.itemCount}</strong></div>
        <span className={`status-badge ${coverage.status === "unavailable" ? "bad" : "neutral"}`}>
          {coverage.status}
        </span>
      </div>
      <p>{coverage.summary}</p>
      <dl className="workspace-operations-inbox-meta">
        <div><dt>Failure</dt><dd>{coverage.failureCode}</dd></div>
        <div><dt>Audit</dt><dd>{coverage.auditRef}</dd></div>
      </dl>
    </article>
  );
}

function severityTone(severity: WorkspaceOperationsInboxSeverity): "good" | "bad" | "neutral" {
  return severity === "critical" || severity === "high" ? "bad" : "neutral";
}
