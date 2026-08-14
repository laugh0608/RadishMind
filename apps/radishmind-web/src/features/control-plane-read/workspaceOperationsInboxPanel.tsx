import { type MouseEvent, useEffect, useState } from "react";

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
  const [selectedItemId, setSelectedItemId] = useState(inbox.items[0]?.itemId ?? "");

  useEffect(() => {
    setSelectedItemId((currentItemId) => (
      inbox.items.some((item) => item.itemId === currentItemId)
        ? currentItemId
        : inbox.items[0]?.itemId ?? ""
    ));
  }, [inbox.items]);

  const selectedItem = inbox.items.find((item) => item.itemId === selectedItemId) ?? inbox.items[0] ?? null;
  const selectedCoverage = selectedItem
    ? inbox.coverage.find((coverage) => coverage.sourceId === selectedItem.sourceId) ?? null
    : null;

  return (
    <section
      className="surface-band workspace-operations-inbox"
      id="workspace-operations-inbox"
      aria-labelledby="workspace-operations-inbox-title"
    >
      <div className="workspace-operations-inbox-heading">
        <div>
          <h3 id="workspace-operations-inbox-title">Operations inbox</h3>
        </div>
        <span className={`status-badge ${inbox.status === "ready" ? "good" : inbox.status === "blocked" ? "bad" : "neutral"}`}>
          {statusLabel(inbox.status)}
        </span>
      </div>

      <div className="workspace-operations-layout">
        <div className="workspace-operations-list">
          {inbox.items.length > 0 ? (
            <div className="workspace-operations-inbox-items" aria-label="Workspace operations attention items">
              {inbox.items.map((item) => {
                const selected = item.itemId === selectedItem?.itemId;
                return (
                  <button
                    key={item.itemId}
                    className={`workspace-operations-inbox-item ${item.severity}${selected ? " is-selected" : ""}`}
                    type="button"
                    aria-pressed={selected}
                    onClick={() => setSelectedItemId(item.itemId)}
                  >
                    <span className={`workspace-operations-item-mark ${item.sourceId}`} aria-hidden="true">
                      {sourceSymbol(item.sourceId)}
                    </span>
                    <span className="workspace-operations-inbox-item-copy">
                      <strong>{item.title}</strong>
                      <span>{item.summary}</span>
                      <small>{sourceLabel(item.sourceId)} · {formatObservedTime(item.occurredAt)}</small>
                    </span>
                    <span className={`status-badge ${severityTone(item.severity)}`}>{severityLabel(item.severity)}</span>
                    <span className="workspace-operations-item-chevron" aria-hidden="true">›</span>
                  </button>
                );
              })}
            </div>
          ) : (
            <article className="workspace-operations-inbox-empty">
              <span className="workspace-operations-empty-mark" aria-hidden="true">✓</span>
              <h4>{inbox.currentWindowHasNoAttentionItems ? "No attention items" : "No resource items available"}</h4>
              <p>
                {inbox.currentWindowHasNoAttentionItems
                  ? "The current source windows are complete."
                  : "Review source coverage before interpreting the empty queue."}
              </p>
            </article>
          )}

          <footer className="workspace-operations-inbox-footer">
            <span>Loaded from {inbox.coverage.length} authorized source projections</span>
            {selectedItem ? (
              <a
                className="workspace-operations-mobile-open"
                href={selectedItem.targetAnchor}
                onClick={(event) => openInboxItem(event, selectedItem, onOpenItem)}
              >
                Open selected <span aria-hidden="true">↗</span>
              </a>
            ) : (
              <a href="#workspace-applications">View workspace <span aria-hidden="true">→</span></a>
            )}
          </footer>
        </div>

        <aside className="workspace-attention-detail" aria-live="polite">
          {selectedItem ? (
            <>
              <header className="workspace-attention-detail-heading">
                <div>
                  <h4>{selectedItem.title}</h4>
                </div>
                <span className={`status-badge ${severityTone(selectedItem.severity)}`}>
                  {severityLabel(selectedItem.severity)}
                </span>
              </header>

              <p className="workspace-attention-summary">{selectedItem.summary}</p>

              <div className="workspace-attention-readonly">
                <span aria-hidden="true">!</span>
                <p>Read-only evidence. Review continues in the owning surface; no action is executed here.</p>
              </div>

              <div className="workspace-attention-resource">
                <span className={`workspace-operations-item-mark ${selectedItem.sourceId}`} aria-hidden="true">
                  {sourceSymbol(selectedItem.sourceId)}
                </span>
                <span>
                  <small>{sourceLabel(selectedItem.sourceId)}</small>
                  <strong>{selectedItem.resourceRef}</strong>
                </span>
              </div>

              <section className="workspace-attention-evidence" aria-labelledby="workspace-attention-evidence-title">
                <h5 id="workspace-attention-evidence-title">Evidence path</h5>
                <div className="workspace-attention-evidence-path">
                  <EvidenceStep
                    index="01"
                    label="Source projection"
                    value={selectedCoverage ? coverageStatusLabel(selectedCoverage) : "Unavailable"}
                    tone={selectedCoverage?.status === "complete_window" ? "good" : selectedCoverage?.status === "unavailable" ? "bad" : "neutral"}
                  />
                  <EvidenceStep
                    index="02"
                    label="Attention severity"
                    value={severityLabel(selectedItem.severity)}
                    tone={severityTone(selectedItem.severity)}
                  />
                  <EvidenceStep index="03" label="Authority" value="Read only" tone="neutral" />
                </div>
              </section>

              <dl className="workspace-attention-boundary">
                <div><dt>Mutation</dt><dd>{inbox.canMutate ? "Enabled" : "Locked"}</dd></div>
                <div><dt>Remediation</dt><dd>{inbox.canRemediate ? "Enabled" : "Review only"}</dd></div>
                <div><dt>Business truth</dt><dd>{inbox.canWriteBusinessTruth ? "Writable" : "Read only"}</dd></div>
              </dl>

              <a
                className="workspace-attention-open"
                href={selectedItem.targetAnchor}
                onClick={(event) => openInboxItem(event, selectedItem, onOpenItem)}
              >
                Open evidence <span aria-hidden="true">↗</span>
              </a>
            </>
          ) : (
            <div className="workspace-attention-detail-empty">
              <span aria-hidden="true">◇</span>
              <p>Select an attention item to inspect its evidence path.</p>
            </div>
          )}
        </aside>
      </div>
    </section>
  );
}

function EvidenceStep({
  index,
  label,
  value,
  tone,
}: {
  index: string;
  label: string;
  value: string;
  tone: "good" | "bad" | "neutral";
}) {
  return (
    <div className="workspace-attention-evidence-step">
      <span>{index}</span>
      <strong>{label}</strong>
      <em className={tone}>{value}</em>
    </div>
  );
}

function openInboxItem(
  event: MouseEvent<HTMLAnchorElement>,
  item: WorkspaceOperationsInboxItem,
  onOpenItem: (item: WorkspaceOperationsInboxItem) => void,
) {
  if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) {
    return;
  }
  event.preventDefault();
  onOpenItem(item);
  window.location.hash = item.targetAnchor;
  scrollToInboxTargetWhenRendered(item.targetAnchor);
}

function scrollToInboxTargetWhenRendered(targetAnchor: WorkspaceOperationsInboxItem["targetAnchor"]) {
  const scrollToTarget = () => {
    const target = document.querySelector<HTMLElement>(targetAnchor);
    if (!target) {
      return false;
    }
    window.requestAnimationFrame(() => {
      const scrollMarginTop = Number.parseFloat(window.getComputedStyle(target).scrollMarginTop) || 0;
      window.scrollTo({
        top: target.getBoundingClientRect().top + window.scrollY - scrollMarginTop,
        behavior: "auto",
      });
    });
    return true;
  };

  if (scrollToTarget()) {
    return;
  }

  const observer = new MutationObserver(() => {
    if (scrollToTarget()) {
      observer.disconnect();
    }
  });
  observer.observe(document.body, { childList: true, subtree: true });
  window.setTimeout(() => observer.disconnect(), 2_000);
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

function sourceSymbol(sourceId: WorkspaceOperationsInboxItem["sourceId"]): string {
  switch (sourceId) {
    case "api_keys":
      return "◇";
    case "workflow_definitions":
      return "⌁";
    case "runs":
      return "↶";
    default:
      return "▣";
  }
}

function coverageStatusLabel(coverage: WorkspaceOperationsInboxCoverage): string {
  switch (coverage.status) {
    case "complete_window":
      return "Ready";
    case "partial_window":
      return "Partial";
    default:
      return "Blocked";
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
  return severity === "critical" || severity === "high" ? "bad" : severity === "info" ? "good" : "neutral";
}

function severityLabel(severity: WorkspaceOperationsInboxSeverity): string {
  return severity.charAt(0).toUpperCase() + severity.slice(1);
}

function statusLabel(status: WorkspaceOperationsInboxViewModel["status"]): string {
  return status.charAt(0).toUpperCase() + status.slice(1);
}
