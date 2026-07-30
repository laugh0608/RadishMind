import type {
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
      <div className="workspace-operations-inbox-heading">
        <h3 id="workspace-operations-inbox-title">Operations inbox</h3>
        <div>
          <span className="workspace-inbox-count">{inbox.items.length}</span>
          <span className={`status-badge ${inbox.status === "ready" ? "good" : inbox.status === "blocked" ? "bad" : "neutral"}`}>
            {inbox.status}
          </span>
        </div>
      </div>

      <div className="workspace-operations-layout">
        <div className="workspace-operations-list">
          <div className="workspace-operations-column-labels" aria-hidden="true">
            <span>Attention item</span>
            <span>Source</span>
            <span>Observed</span>
            <span>State</span>
          </div>
          {inbox.items.length > 0 ? (
            <div className="workspace-operations-inbox-items" aria-label="Workspace operations attention items">
              {inbox.items.map((item) => (
                <article key={item.itemId} className={`workspace-operations-inbox-item ${item.severity}`}>
                  <div className="workspace-operations-inbox-item-main">
                    <span className="workspace-operations-item-mark" aria-hidden="true">◇</span>
                    <div>
                      <h4>{item.title}</h4>
                      <small>{item.resourceRef}</small>
                    </div>
                  </div>
                  <span className="workspace-operations-source">{sourceLabel(item.sourceId)}</span>
                  <time dateTime={item.occurredAt}>{formatObservedTime(item.occurredAt)}</time>
                  <div className="workspace-operations-inbox-action">
                    <span className={`status-badge ${severityTone(item.severity)}`}>{item.severity}</span>
                    <a
                      href={item.targetAnchor}
                      aria-label={`Open ${item.title}`}
                      onClick={() => onOpenItem(item)}
                    >
                      <span aria-hidden="true">↗</span>
                    </a>
                  </div>
                </article>
              ))}
            </div>
          ) : (
            <article className="workspace-operations-inbox-empty">
              <h4>{inbox.currentWindowHasNoAttentionItems ? "No attention items" : "No resource items available"}</h4>
              <p>
                {inbox.currentWindowHasNoAttentionItems
                  ? "The current source windows are complete."
                  : "Review source coverage before interpreting the empty queue."}
              </p>
            </article>
          )}
          <footer className="workspace-operations-inbox-footer">
            <span>{inbox.items.length} attention items · current loaded window</span>
            <a href="#workspace-applications">View workspace <span aria-hidden="true">→</span></a>
          </footer>
        </div>

        <aside className="workspace-operations-aside">
          <section className="workspace-continue-work">
            <h4>Continue work</h4>
            {inbox.items.slice(0, 2).map((item) => (
              <a
                key={item.itemId}
                href={item.targetAnchor}
                onClick={() => onOpenItem(item)}
              >
                <span className="workspace-operations-item-mark" aria-hidden="true">◇</span>
                <span>
                  <strong>{item.title}</strong>
                  <small>{sourceLabel(item.sourceId)} · {formatObservedTime(item.occurredAt)}</small>
                </span>
                <span className={`status-badge ${severityTone(item.severity)}`}>{item.severity}</span>
              </a>
            ))}
            {inbox.items.length === 0 ? <p>No current items.</p> : null}
          </section>

          <section className="workspace-write-boundary">
            <div className="workspace-write-boundary-heading">
              <span aria-hidden="true">◇</span>
              <span className="status-badge neutral">Guarded</span>
            </div>
            <h4>Write boundary</h4>
            <dl>
              <div><dt>Mutation</dt><dd>{inbox.canMutate ? "Enabled" : "Locked"}</dd></div>
              <div><dt>Remediation</dt><dd>{inbox.canRemediate ? "Enabled" : "Review only"}</dd></div>
              <div><dt>Business truth</dt><dd>{inbox.canWriteBusinessTruth ? "Writable" : "Read only"}</dd></div>
            </dl>
          </section>
        </aside>
      </div>
    </section>
  );
}

function sourceLabel(sourceId: WorkspaceOperationsInboxItem["sourceId"]): string {
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

function formatObservedTime(value: string): string {
  const timestamp = Date.parse(value);
  if (!Number.isFinite(timestamp)) {
    return value || "Not available";
  }
  const elapsedMilliseconds = Math.max(0, Date.now() - timestamp);
  const minutes = Math.floor(elapsedMilliseconds / 60_000);
  if (minutes < 1) {
    return "Now";
  }
  if (minutes < 60) {
    return `${minutes} min`;
  }
  const hours = Math.floor(minutes / 60);
  if (hours < 24) {
    return `${hours} h`;
  }
  return `${Math.floor(hours / 24)} d`;
}

function severityTone(severity: WorkspaceOperationsInboxSeverity): "good" | "bad" | "neutral" {
  return severity === "critical" || severity === "high" ? "bad" : "neutral";
}
