import "../../i18n/usageQuotaResources.ts";
import "../../i18n/gatewayReviewResources.ts";
import { useTranslation } from "react-i18next";
import { useLocalePreference } from "../../i18n/LocaleProvider.tsx";
import { formatDisplayNumber } from "../../i18n/formatters.ts";
import { gatewayReviewState } from "./gatewayReviewMessages.ts";
import type { WorkspaceUsageQuotaViewModel } from "./workspaceUsageQuota.ts";

export default function WorkspaceUsageQuotaPanel({ view }: { view: WorkspaceUsageQuotaViewModel }) {
  const { t } = useTranslation("gateway");
  const { locale } = useLocalePreference();
  const number = (value: number) => formatDisplayNumber(value, locale) ?? t($ => $.usageQuota.unavailable);
  const quota = view.quota;
  const limits = quota ? [
    { id: "requests" as const, used: quota.usage_snapshot.request_count, limit: quota.request_limit },
    { id: "tokens" as const, used: quota.usage_snapshot.token_count, limit: quota.token_limit },
    { id: "cost" as const, used: quota.usage_snapshot.estimated_cost, limit: quota.cost_limit },
  ] : [];
  const period = quota?.period === "monthly" ? t($ => $.usageQuota.monthly) : quota?.period ?? t($ => $.usageQuota.unavailable);
  return (
    <section className="surface-band workspace-usage-quota" id="workspace-usage-quota" aria-labelledby="workspace-usage-quota-title">
      <div className="section-heading">
        <div><p className="eyebrow">{t($ => $.usageQuota.workspace)}</p><h3 id="workspace-usage-quota-title">{t($ => $.usageQuota.title)}</h3></div>
        <span className={`status-badge ${view.canRenderQuota ? "good" : "bad"}`}>{view.canRenderQuota ? t($ => $.usageQuota.readOnlyReady) : t($ => $.usageQuota.blocked)}</span>
      </div>
      <div className="usage-quota-summary">
        <article className="usage-quota-route">
          <div className="card-title-row"><div><p className="eyebrow">{t($ => $.usageQuota.route)}</p><h4>{view.routeId}</h4></div><span className="status-badge neutral">{view.requiredScope}</span></div>
          <p className="route-path">{view.routePath}</p>
          <dl className="tenant-meta">
            <div><dt>{t($ => $.usageQuota.model)}</dt><dd>{view.readModel}</dd></div>
            <div><dt>{t($ => $.usageQuota.period)}</dt><dd>{period}</dd></div>
            <div><dt>{t($ => $.usageQuota.request)}</dt><dd>{view.requestId}</dd></div>
            <div><dt>{t($ => $.usageQuota.audit)}</dt><dd>{view.auditRef}</dd></div>
          </dl>
        </article>
        <div className="usage-quota-snapshot" aria-label={t($ => $.usageQuota.snapshot)}>
          {quota ? <>
            <article className="usage-quota-snapshot-card"><span>{t($ => $.usageQuota.quota)}</span><strong>{quota.quota_id}</strong><p>{quota.tenant_ref}</p></article>
            <article className="usage-quota-snapshot-card"><span>{t($ => $.usageQuota.period)}</span><strong>{period}</strong><p>{t($ => $.usageQuota.window)}</p></article>
            <article className="usage-quota-snapshot-card"><span>{t($ => $.usageQuota.failure)}</span><strong>{quota.over_quota_failure_code}</strong><p>{t($ => $.usageQuota.noEnforcement)}</p></article>
            <article className="usage-quota-snapshot-card"><span>{t($ => $.usageQuota.audit)}</span><strong>{t($ => $.usageQuota.available)}</strong><p>{t($ => $.usageQuota.references)}</p></article>
          </> : <p>{t($ => $.usageQuota.noPolicy)}</p>}
        </div>
      </div>
      <div className="usage-quota-limits" aria-label={t($ => $.usageQuota.limits)}>
        {limits.map(item => <article className="usage-quota-limit" key={item.id}>
          <span>{t($ => $.usageQuota.measures[item.id])}</span><strong>{number(item.used)}</strong>
          <p>{t($ => $.usageQuota.limit, { limit: number(item.limit), percent: item.limit > 0 ? formatDisplayNumber(item.used / item.limit, locale, { style: "percent", minimumFractionDigits: 1, maximumFractionDigits: 1 }) ?? t($ => $.usageQuota.unavailable) : t($ => $.usageQuota.unavailable) })}</p>
          {item.id === "cost" ? <p>{t($ => $.usageQuota.noCurrency)}</p> : null}
        </article>)}
      </div>
      <div className="usage-quota-failure"><span>{t($ => $.usageQuota.overQuota)}</span><strong>{view.overQuotaFailureCode}</strong><p>{t($ => $.usageQuota.boundary)}</p></div>
      <div className="usage-quota-states" aria-label={t($ => $.usageQuota.states)}>
        {view.statePreviews.map(state => <article className="usage-quota-state" key={state.id}>
          <div><strong>{t($ => $.usageQuota.previews[state.id].label)}</strong><span>{gatewayReviewState(t, state.status)}</span></div>
          <p>{t($ => $.usageQuota.previews[state.id].summary)}</p>
          <small>{t($ => $.usageQuota.failureCount, { count: state.itemCount, code: state.failureCode })}</small>
        </article>)}
      </div>
    </section>
  );
}
