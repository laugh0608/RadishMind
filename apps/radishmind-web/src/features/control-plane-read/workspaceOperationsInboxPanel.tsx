import "../../i18n/operationsInboxResources.ts";
import "../../i18n/gatewayReviewResources.ts";
import { useTranslation } from "react-i18next";
import { useLocalePreference } from "../../i18n/LocaleProvider.tsx";
import { formatDisplayDate } from "../../i18n/formatters.ts";
import { gatewayReviewState } from "./gatewayReviewMessages.ts";
import { inboxItemCopy } from "./workspaceOperationsInboxMessages.ts";
import { type MouseEvent, useEffect, useState } from "react";

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
  const { t } = useTranslation("gateway");
  const { locale } = useLocalePreference();
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
          <h3 id="workspace-operations-inbox-title">{t($ => $.operationsInbox.heading)}</h3>
        </div>
        <span className={`status-badge ${inbox.status === "ready" ? "good" : inbox.status === "blocked" ? "bad" : "neutral"}`}>
          {gatewayReviewState(t, inbox.status)}
        </span>
      </div>

      <div className="workspace-operations-layout">
        <div className="workspace-operations-list">
          {inbox.items.length > 0 ? (
            <div className="workspace-operations-inbox-items" aria-label={t($ => $.operationsInbox.items)}>
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
                      <strong>{inboxItemCopy(t, item).title}</strong>
                      <span>{inboxItemCopy(t, item).summary}</span>
                      <small>{t($ => $.operationsInbox.sources[item.sourceId])} · <time dateTime={item.occurredAt} title={item.occurredAt}>{formatDisplayDate(item.occurredAt, locale, { timeZone: "UTC" }) ? `${formatDisplayDate(item.occurredAt, locale, { timeZone: "UTC" })} UTC` : t($ => $.operationsInbox.unavailable)}</time></small>
                    </span>
                    <span className={`status-badge ${severityTone(item.severity)}`}>{gatewayReviewState(t, item.severity)}</span>
                    <span className="workspace-operations-item-chevron" aria-hidden="true">›</span>
                  </button>
                );
              })}
            </div>
          ) : (
            <article className="workspace-operations-inbox-empty">
              <span className="workspace-operations-empty-mark" aria-hidden="true">✓</span>
              <h4>{inbox.currentWindowHasNoAttentionItems ? t($ => $.operationsInbox.noAttention) : t($ => $.operationsInbox.noResources)}</h4>
              <p>
                {inbox.currentWindowHasNoAttentionItems
                  ? t($ => $.operationsInbox.complete)
                  : t($ => $.operationsInbox.checkCoverage)}
              </p>
            </article>
          )}

          <footer className="workspace-operations-inbox-footer">
            <span>{t($ => $.operationsInbox.sourceCount, { count: inbox.coverage.length })}</span>
            {selectedItem ? (
              <a
                className="workspace-operations-mobile-open"
                href={selectedItem.targetAnchor}
                onClick={(event) => openInboxItem(event, selectedItem, onOpenItem)}
              >
                {t($ => $.operationsInbox.openSelected)} <span aria-hidden="true">↗</span>
              </a>
            ) : (
              <a href="#workspace-applications">{t($ => $.operationsInbox.viewWorkspace)} <span aria-hidden="true">→</span></a>
            )}
          </footer>
        </div>

        <aside className="workspace-attention-detail" aria-live="polite">
          {selectedItem ? (
            <>
              <header className="workspace-attention-detail-heading">
                <div>
                  <h4>{inboxItemCopy(t, selectedItem).title}</h4>
                </div>
                <span className={`status-badge ${severityTone(selectedItem.severity)}`}>
                  {gatewayReviewState(t, selectedItem.severity)}
                </span>
              </header>

              <p className="workspace-attention-summary">{inboxItemCopy(t, selectedItem).summary}</p>

              <div className="workspace-attention-readonly">
                <span aria-hidden="true">!</span>
                <p>{t($ => $.operationsInbox.readOnlyEvidence)}</p>
              </div>

              <div className="workspace-attention-resource">
                <span className={`workspace-operations-item-mark ${selectedItem.sourceId}`} aria-hidden="true">
                  {sourceSymbol(selectedItem.sourceId)}
                </span>
                <span>
                  <small>{t($ => $.operationsInbox.sources[selectedItem.sourceId])}</small>
                  <strong>{selectedItem.resourceRef}</strong>
                </span>
              </div>

              <section className="workspace-attention-evidence" aria-labelledby="workspace-attention-evidence-title">
                <h5 id="workspace-attention-evidence-title">{t($ => $.operationsInbox.evidencePath)}</h5>
                <div className="workspace-attention-evidence-path">
                  <EvidenceStep
                    index="01"
                    label={t($ => $.operationsInbox.projection)}
                    value={selectedCoverage ? gatewayReviewState(t, selectedCoverage.status) : t($ => $.operationsInbox.unavailable)}
                    tone={selectedCoverage?.status === "complete_window" ? "good" : selectedCoverage?.status === "unavailable" ? "bad" : "neutral"}
                  />
                  <EvidenceStep
                    index="02"
                    label={t($ => $.operationsInbox.severity)}
                    value={gatewayReviewState(t, selectedItem.severity)}
                    tone={severityTone(selectedItem.severity)}
                  />
                  <EvidenceStep index="03" label={t($ => $.operationsInbox.authority)} value={t($ => $.operationsInbox.readOnly)} tone="neutral" />
                </div>
              </section>

              <dl className="workspace-attention-boundary">
                <div><dt>{t($ => $.operationsInbox.mutation)}</dt><dd>{inbox.canMutate ? t($ => $.operationsInbox.enabled) : t($ => $.operationsInbox.locked)}</dd></div>
                <div><dt>{t($ => $.operationsInbox.remediation)}</dt><dd>{inbox.canRemediate ? t($ => $.operationsInbox.enabled) : t($ => $.operationsInbox.reviewOnly)}</dd></div>
                <div><dt>{t($ => $.operationsInbox.businessTruth)}</dt><dd>{inbox.canWriteBusinessTruth ? t($ => $.operationsInbox.writable) : t($ => $.operationsInbox.readOnly)}</dd></div>
              </dl>

              <a
                className="workspace-attention-open"
                href={selectedItem.targetAnchor}
                onClick={(event) => openInboxItem(event, selectedItem, onOpenItem)}
              >
                {t($ => $.operationsInbox.openEvidence)} <span aria-hidden="true">↗</span>
              </a>
            </>
          ) : (
            <div className="workspace-attention-detail-empty">
              <span aria-hidden="true">◇</span>
              <p>{t($ => $.operationsInbox.select)}</p>
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

function severityTone(severity: WorkspaceOperationsInboxSeverity): "good" | "bad" | "neutral" {
  return severity === "critical" || severity === "high" ? "bad" : severity === "info" ? "good" : "neutral";
}
