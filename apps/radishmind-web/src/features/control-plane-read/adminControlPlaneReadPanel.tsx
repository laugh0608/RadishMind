import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useLocalePreference } from "../../i18n/LocaleProvider.tsx";
import { formatDisplayNumber } from "../../i18n/formatters.ts";
import { adminDisplayDate } from "./adminManagementFormatters.ts";
import { buildAdminAuditLogViewModel, type AdminAuditEventRow, type AdminAuditLogViewModel } from "./adminAuditLog.ts";
import type { AdminTenantOverviewViewModel } from "./adminTenantOverview.ts";
import { loadControlPlaneReadDevLiveCollection, type ControlPlaneReadDevLiveConfig } from "./devLiveReadConsumer.ts";
import "../../i18n/adminReadResources.ts";

export function AdminTenantOwner({ overview }: { overview: AdminTenantOverviewViewModel }) {
  const { t } = useTranslation("admin");
  return (
    <section className="admin-control-owner-surface admin-control-tenant-owner" aria-labelledby="admin-tenant-owner-title">
      <header>
        <div>
          <p className="eyebrow">{t($ => $.read.tenantEyebrow, { scope: overview.requiredScope })}</p>
          <h4 id="admin-tenant-owner-title">{overview.tenant?.tenant_display_name ?? t($ => $.read.tenantUnavailable)}</h4>
        </div>
        <StatusPill tone={overview.canRenderTenant ? "ready" : "blocked"}>
          {overview.canRenderTenant ? t($ => $.read.readOnly) : t($ => $.read.blocked)}
        </StatusPill>
      </header>
      <div className="admin-control-tenant-layout">
        <dl className="admin-control-owner-meta">
          <div><dt>{t($ => $.read.tenantRef)}</dt><dd>{overview.collection.tenantRef}</dd></div>
          <div><dt>{t($ => $.read.route)}</dt><dd>{overview.routePath}</dd></div>
          <div><dt>{t($ => $.read.request)}</dt><dd>{overview.requestId}</dd></div>
          <div><dt>{t($ => $.read.audit)}</dt><dd>{overview.auditRef}</dd></div>
        </dl>
        <div className="admin-control-fact-list" aria-label={t($ => $.read.tenantFacts)}>
          {overview.facts.map((fact) => (
            <article key={fact.kind}>
              <span>{t($ => $.read.facts[fact.kind])}</span><strong title={fact.value}>{fact.kind === "tenantState" && fact.value === "active" ? t($ => $.read.states.active) : fact.value}</strong><small>{fact.kind === "tenantState" ? fact.detail : t($ => $.read.factDetails[fact.kind as Exclude<typeof fact.kind, "tenantState">])}</small>
            </article>
          ))}
        </div>
      </div>
      <BoundaryNotice>
        {t($ => $.read.tenantBoundary)}</BoundaryNotice>
    </section>
  );
}

export function AdminAuditOwner({
  auditLog,
  sourceConfig,
}: {
  auditLog: AdminAuditLogViewModel;
  sourceConfig: ControlPlaneReadDevLiveConfig;
}) {
  const { t } = useTranslation("admin");
  const { locale } = useLocalePreference();
  const [pages, setPages] = useState<AdminAuditLogViewModel[]>([auditLog]);
  const [pageIndex, setPageIndex] = useState(0);
  const [selectedAuditRef, setSelectedAuditRef] = useState(auditLog.auditEvents[0]?.auditRef ?? "");
  const [paginationStatus, setPaginationStatus] = useState<"idle" | "loading" | "failed">("idle");

  useEffect(() => {
    setPages([auditLog]);
    setPageIndex(0);
    setSelectedAuditRef(auditLog.auditEvents[0]?.auditRef ?? "");
    setPaginationStatus("idle");
  }, [auditLog, sourceConfig.baseUrl, sourceConfig.tenantRef]);

  const page = pages[pageIndex] ?? auditLog;
  const selectedEvent = page.auditEvents.find((event) => event.auditRef === selectedAuditRef) ??
    page.auditEvents[0] ?? null;

  useEffect(() => {
    if (!page.auditEvents.some((event) => event.auditRef === selectedAuditRef)) {
      setSelectedAuditRef(page.auditEvents[0]?.auditRef ?? "");
    }
  }, [page, selectedAuditRef]);

  async function loadNextPage() {
    if (!page.nextCursor || sourceConfig.mode !== "dev_live_http") return;
    setPaginationStatus("loading");
    try {
      const collection = await loadControlPlaneReadDevLiveCollection(
        sourceConfig,
        "audit-summary-list-route",
        { cursor: page.nextCursor, limit: 50, sort: "recorded_at_desc" },
      );
      const nextPage = buildAdminAuditLogViewModel(collection);
      setPages((current) => [...current.slice(0, pageIndex + 1), nextPage]);
      setPageIndex((current) => current + 1);
      setSelectedAuditRef(nextPage.auditEvents[0]?.auditRef ?? "");
      setPaginationStatus("idle");
    } catch {
      setPaginationStatus("failed");
    }
  }

  return (
    <section className="admin-control-owner-surface admin-control-audit-owner" aria-labelledby="admin-audit-owner-title">
      <header>
        <div>
          <p className="eyebrow">{t($ => $.read.auditEyebrow, { scope: page.requiredScope })}</p>
          <h4 id="admin-audit-owner-title">{t($ => $.read.auditTitle)}</h4>
        </div>
        <StatusPill tone={page.canRenderAuditLog ? "ready" : "blocked"}>
          {page.canRenderAuditLog ? t($ => $.read.page, { number: formatDisplayNumber(pageIndex + 1, locale) ?? t($ => $.read.states.unknown) }) : t($ => $.read.states[page.collection.statusLabel])}
        </StatusPill>
      </header>
      <div className="admin-control-audit-toolbar">
        <span>{t($ => $.read.windowRecords, { countText: formatDisplayNumber(page.auditEvents.length, locale) ?? t($ => $.read.states.unknown) })}</span>
        <span>{page.nextCursor ? t($ => $.read.nextCursor) : t($ => $.read.endPages)}</span>
        <div>
          <button type="button" className="secondary-action" disabled={pageIndex === 0 || paginationStatus === "loading"} onClick={() => setPageIndex((current) => current - 1)}>
            {t($ => $.read.previousPage)}</button>
          <button
            type="button"
            className="secondary-action"
            disabled={!page.nextCursor || sourceConfig.mode !== "dev_live_http" || paginationStatus === "loading"}
            onClick={() => void loadNextPage()}
          >
            {paginationStatus === "loading" ? t($ => $.read.loading) : t($ => $.read.nextPage)}
          </button>
        </div>
      </div>
      {paginationStatus === "failed" ? (
        <p className="admin-control-audit-failure" role="alert">{t($ => $.read.nextPageFailed)}</p>
      ) : null}
      <div className="admin-control-audit-layout">
        <div className="admin-control-audit-list" aria-label={t($ => $.read.auditEvents)}>
          {page.auditEvents.length ? page.auditEvents.map((event) => (
            <AuditSelectionRow
              key={event.auditRef}
              event={event}
              selected={event.auditRef === selectedEvent?.auditRef}
              onSelect={setSelectedAuditRef}
            />
          )) : (
            <div className="admin-control-audit-empty">{t($ => $.read.noRecords)}</div>
          )}
        </div>
        <AuditReadOnlyDetail event={selectedEvent} requestId={page.requestId} />
      </div>
      <BoundaryNotice>
        {t($ => $.read.auditBoundary)}</BoundaryNotice>
    </section>
  );
}

function AuditSelectionRow({
  event,
  selected,
  onSelect,
}: {
  event: AdminAuditEventRow;
  selected: boolean;
  onSelect: (auditRef: string) => void;
}) {
  const { t } = useTranslation("admin");
  const { locale } = useLocalePreference();
  return (
    <button
      type="button"
      className={`admin-control-audit-row ${selected ? "is-selected" : ""}`}
      onClick={() => onSelect(event.auditRef)}
      aria-pressed={selected}
    >
      <i aria-hidden="true" />
      <span>
        <strong>{event.eventKind}</strong>
        <small>{event.auditRef}</small>
        <small>{event.resourceRef} · <time dateTime={event.recordedAt} title={event.recordedAt}>{adminDisplayDate(event.recordedAt, locale)}</time></small>
      </span>
      <StatusPill tone={event.decision === "denied" ? "blocked" : "ready"}>{event.decision === "allowed" || event.decision === "denied" ? t($ => $.read.states[event.decision as "allowed" | "denied"]) : event.decision}</StatusPill>
    </button>
  );
}

function AuditReadOnlyDetail({ event, requestId }: { event: AdminAuditEventRow | null; requestId: string }) {
  const { t } = useTranslation("admin");
  const { locale } = useLocalePreference();
  return (
    <aside className="admin-control-audit-detail" aria-label={t($ => $.read.detailLabel)}>
      <header>
        <div><p className="eyebrow">{t($ => $.read.detailTitle)}</p><h5>{event?.auditRef ?? t($ => $.read.noSelection)}</h5></div>
        <StatusPill tone="neutral">{t($ => $.read.metadataOnly)}</StatusPill>
      </header>
      {event ? (
        <dl className="admin-control-owner-meta">
          <div><dt>{t($ => $.read.actor)}</dt><dd>{event.actorSubjectRef}</dd></div>
          <div><dt>{t($ => $.read.resource)}</dt><dd>{event.resourceRef}</dd></div>
          <div><dt>{t($ => $.read.decision)}</dt><dd>{event.decision === "allowed" || event.decision === "denied" ? t($ => $.read.states[event.decision as "allowed" | "denied"]) : event.decision}</dd></div>
          <div><dt>{t($ => $.read.failure)}</dt><dd>{event.failureCode === "none" ? t($ => $.read.states.none) : event.failureCode}</dd></div>
          <div><dt>{t($ => $.read.trace)}</dt><dd>{event.traceId}</dd></div>
          <div><dt>{t($ => $.read.recorded)}</dt><dd><time dateTime={event.recordedAt} title={event.recordedAt}>{adminDisplayDate(event.recordedAt, locale)}</time></dd></div>
          <div><dt>{t($ => $.read.request)}</dt><dd>{requestId}</dd></div>
        </dl>
      ) : <p className="admin-control-audit-empty">{t($ => $.read.emptyWindow)}</p>}
      <div className="admin-control-readonly-actions">
        <span>{t($ => $.read.editBlocked)}</span><span>{t($ => $.read.deleteBlocked)}</span><span>{t($ => $.read.exportBlocked)}</span>
      </div>
    </aside>
  );
}

function StatusPill({
  tone,
  children,
}: {
  tone: "ready" | "blocked" | "neutral";
  children: string;
}) {
  return <span className={`admin-control-status is-${tone}`}>{children}</span>;
}

function BoundaryNotice({ children }: { children: string }) {
  return <p className="admin-control-boundary-notice"><span aria-hidden="true">!</span>{children}</p>;
}
