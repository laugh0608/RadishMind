import "../../i18n/operationsInboxResources.ts";
import "../../i18n/operationsOverviewResources.ts";
import "../../i18n/gatewayReviewResources.ts";
import { useTranslation } from "react-i18next";
import { useLocalePreference } from "../../i18n/LocaleProvider.tsx";
import { formatDisplayDate, formatDisplayNumber } from "../../i18n/formatters.ts";
import { gatewayReviewState } from "./gatewayReviewMessages.ts";
import type { ApplicationDevelopmentWorkspaceContext } from "./applicationDevelopmentWorkspace";
import type {
  ControlPlaneReadDevLiveConfig,
  ControlPlaneReadDevLiveLoadState,
} from "./devLiveReadConsumer";
import type {
  WorkspaceOperationsInboxCoverage,
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
  const { t } = useTranslation("gateway");
  const { locale } = useLocalePreference();
  const sourceModeLabel = sourceConfig.mode === "dev_live_http" ? t($ => $.operationsOverview.devTest) : t($ => $.operationsOverview.offline);
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
          <h2 id="workspace-overview-title">{t($ => $.operationsOverview.workspace)}</h2>
        </div>

        <article className="workspace-active-application" aria-label={t($ => $.operationsOverview.activeApplication)}>
          <span className="workspace-active-application-icon" aria-hidden="true">AI</span>
          <span className="workspace-active-application-copy">
            <small>{t($ => $.operationsOverview.activeApplication)}</small>
            <strong>{application.displayName || t($ => $.operationsOverview.noApplication)}</strong>
            <span>
              {application.status === "unavailable"
                ? t($ => $.operationsOverview.noContext)
                : t($ => $.operationsOverview.revision, { kind: application.applicationKind, version: application.recordVersion || t($ => $.operationsOverview.unavailable) })}
            </span>
          </span>
          <a
            href={application.status === "unavailable" ? "#workspace-applications" : "#application-development-workspace"}
            aria-label={application.status === "unavailable" ? t($ => $.operationsOverview.select) : t($ => $.operationsOverview.open, { name: application.displayName })}
          >
            <span aria-hidden="true">↗</span>
          </a>
        </article>
      </header>

      <div className="workspace-overview-grid">
        <article className="workspace-source-pulse">
          <div className="workspace-overview-card-heading">
            <h3>{t($ => $.operationsOverview.pulse)}</h3>
            <span className={`status-badge ${inbox.status === "ready" ? "good" : inbox.status === "blocked" ? "bad" : "neutral"}`}>
              {gatewayReviewState(t, inbox.status)}
            </span>
          </div>

          <div className="workspace-pulse-visual">
            <div className="workspace-pulse-total">
              <strong>{formatDisplayNumber(inbox.items.length, locale, { minimumIntegerDigits: 2 })}</strong>
              <span>{t($ => $.operationsOverview.attention)}</span>
              <small>{t($ => $.operationsOverview.loadedWindow)}</small>
            </div>
            <div className="workspace-pulse-bars" aria-label={t($ => $.operationsOverview.distribution)}>
              {inbox.coverage.map((coverage) => (
                <div
                  key={coverage.sourceId}
                  className={`workspace-pulse-bar ${coverage.status} load-${Math.min(4, coverage.itemCount)}`}
                  title={t($ => $.operationsOverview.coverage, { source: t($ => $.operationsInbox.sources[coverage.sourceId]), count: coverage.itemCount, status: gatewayReviewState(t, coverage.status) })}
                >
                  <span aria-hidden="true" />
                  <small>{t($ => $.operationsInbox.sources[coverage.sourceId])}</small>
                </div>
              ))}
            </div>
          </div>

          <dl className="workspace-pulse-metrics">
            <div><dt>{t($ => $.operationsOverview.readySources)}</dt><dd>{coverageCounts.ready}</dd></div>
            <div><dt>{t($ => $.operationsOverview.partialSources)}</dt><dd>{coverageCounts.partial}</dd></div>
            <div><dt>{t($ => $.operationsOverview.unavailable)}</dt><dd>{coverageCounts.blocked}</dd></div>
          </dl>
        </article>

        <article className="workspace-source-evidence">
          <div className="workspace-overview-card-heading">
            <h3>{t($ => $.operationsOverview.evidence)}</h3>
            <span className={`status-badge ${sourceTone}`}>{gatewayReviewState(t, sourceState.status)}</span>
          </div>

          <div className="workspace-evidence-distribution" aria-label={t($ => $.operationsOverview.evidenceDistribution)}>
            <div className="ready"><strong>{coverageCounts.ready}</strong><span>{t($ => $.operationsOverview.ready)}</span></div>
            <div className="partial"><strong>{coverageCounts.partial}</strong><span>{t($ => $.operationsOverview.partial)}</span></div>
            <div className="blocked"><strong>{coverageCounts.blocked}</strong><span>{t($ => $.operationsOverview.blocked)}</span></div>
          </div>

          <div className="workspace-evidence-matrix">
            {inbox.coverage.map((coverage) => (
              <div key={coverage.sourceId}>
                <span className={`workspace-source-pulse-dot ${coverage.status}`} aria-hidden="true" />
                <span title={t($ => $.operationsInbox.sources[coverage.sourceId])}>{t($ => $.operationsInbox.sources[coverage.sourceId])}</span>
                <strong className={coverage.status}>{gatewayReviewState(t, coverage.status)}</strong>
              </div>
            ))}
          </div>

          <dl className="workspace-source-context">
            <div><dt>{t($ => $.operationsOverview.source)}</dt><dd>{sourceModeLabel}</dd></div>
            <div><dt>{t($ => $.operationsOverview.workspace)}</dt><dd>{inbox.activeWorkspaceId}</dd></div>
            <div><dt>{t($ => $.operationsOverview.snapshot)}</dt><dd><time dateTime={inbox.referenceTime} title={inbox.referenceTime}>{formatDisplayDate(inbox.referenceTime, locale, { timeZone: "UTC" }) ? `${formatDisplayDate(inbox.referenceTime, locale, { timeZone: "UTC" })} UTC` : t($ => $.operationsOverview.unavailable)}</time></dd></div>
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
