import { useTranslation } from "react-i18next";
import "../../i18n/evaluationExecutionResources.ts";
import { campaignState } from "./applicationEvaluationMessages.ts";
import type { ApplicationEvaluationPlan, ApplicationEvaluationPlanVersion, ApplicationEvaluationCampaign } from "./applicationEvaluationCampaignConsumer.ts";
export default function CampaignOwner({
  plan,
  version,
  campaigns,
  selectedCampaign,
  quotaAPIKeyId,
  clientCampaignKey,
  confirmed,
  writeBlocked,
  operationPending,
  onSelectCampaign,
  onNewCampaign,
  onQuotaAPIKeyIdChange,
  onClientCampaignKeyChange,
  onReview,
  onCancel,
  onConfirm,
  onReconcile,
  onOpenHandoff,
}: {
  plan: ApplicationEvaluationPlan | null;
  version: ApplicationEvaluationPlanVersion | null;
  campaigns: ApplicationEvaluationCampaign[];
  selectedCampaign: ApplicationEvaluationCampaign | null;
  quotaAPIKeyId: string;
  clientCampaignKey: string;
  confirmed: boolean;
  writeBlocked: boolean;
  operationPending: boolean;
  onSelectCampaign: (id: string) => void;
  onNewCampaign: () => void;
  onQuotaAPIKeyIdChange: (value: string) => void;
  onClientCampaignKeyChange: (value: string) => void;
  onReview: () => void;
  onCancel: () => void;
  onConfirm: () => void;
  onReconcile: () => void;
  onOpenHandoff: () => void;
}) {
  const { t } = useTranslation("evaluation");
  const total = selectedCampaign?.items.length ?? version?.items.length ?? 0;
  const completed = (selectedCampaign?.succeededItems ?? 0) + (selectedCampaign?.failedItems ?? 0);
  const terminalCampaigns = campaigns.filter((campaign) => campaign.state === "succeeded");
  return (
    <section className="application-evaluation-campaign-owner">
      <div className="application-evaluation-campaign-context">
        <div>
          <span>{t($ => $.campaign.selectedCampaign)}</span>
          <label>
            <select
              value={selectedCampaign?.campaignId ?? ""}
              onChange={(event) => onSelectCampaign(event.target.value)}
              aria-label={t($ => $.campaign.selectedCampaignLabel)}
              disabled={operationPending}
            >
              <option value="">{t($ => $.campaign.newCampaignDraft)}</option>
              {campaigns.map((campaign) => (
                <option key={campaign.campaignId} value={campaign.campaignId}>
                  {campaign.clientCampaignKey} · v{campaign.recordVersion}
                </option>
              ))}
            </select>
          </label>
        </div>
        <div className="application-evaluation-campaign-context-meta">
          <span>{t($ => $.campaign.campaignCount, { count: campaigns.length })}</span>
          <small>{selectedCampaign ? shortRef(selectedCampaign.campaignId) : t($ => $.campaign.noExecution)}</small>
          <div>
            <span className={`status-badge ${selectedCampaign?.state === "succeeded" ? "good" : "neutral"}`}>
              {campaignState(t, selectedCampaign?.state ?? "not_started")}
            </span>
            <button type="button" className="secondary-action" aria-label={t($ => $.campaign.newCampaign)} onClick={onNewCampaign} disabled={writeBlocked || !plan || !version}>{t($ => $.campaign.new)}</button>
          </div>
        </div>
      </div>
      <div className="application-evaluation-campaign-surface">
        <div className="application-evaluation-campaign-main">
          <header>
            <div>
              <p className="eyebrow">{t($ => $.campaign.campaignOwner)}</p>
              <h4>{selectedCampaign?.clientCampaignKey ?? t($ => $.campaign.sequentialExecution)}</h4>
              <p>{t($ => $.campaign.campaignIntro)}</p>
            </div>
            <span className={`status-badge ${selectedCampaign?.state === "succeeded" ? "good" : "neutral"}`}>
              {campaignState(t, selectedCampaign?.state ?? "draft")}
            </span>
          </header>
          <dl className="application-evaluation-authority">
            <div><dt>{t($ => $.campaign.profile)}</dt><dd>{version?.executionProfile ?? "—"}</dd></div>
            <div><dt>{t($ => $.campaign.plan)}</dt><dd>{version ? `v${version.planVersion} · ${shortRef(version.planDigest)}` : "—"}</dd></div>
            <div><dt>{t($ => $.campaign.quotaConsumer)}</dt><dd>{selectedCampaign?.quotaAPIKeyId || t($ => $.campaign.required)}</dd></div>
            <div><dt>{t($ => $.campaign.authority)}</dt><dd>{shortRef(selectedCampaign?.authority?.authorityDigest ?? "")}</dd></div>
          </dl>
          {selectedCampaign ? (
            <>
              <div className="application-evaluation-progress">
                <div><span>{t($ => $.campaign.progress)}</span><strong>{completed} / {total}</strong><i style={{ width: `${total ? Math.min(100, completed / total * 100) : 0}%` }} /></div>
                <dl><div><dt>{t($ => $.campaign.succeeded)}</dt><dd>{selectedCampaign.succeededItems}</dd></div><div><dt>{t($ => $.campaign.failed)}</dt><dd>{selectedCampaign.failedItems}</dd></div><div><dt>{t($ => $.campaign.retryReplay)}</dt><dd>0</dd></div></dl>
              </div>
              <CampaignItems campaign={selectedCampaign} />
              {selectedCampaign.state === "interrupted" || selectedCampaign.state === "running" ? (
                <button type="button" className="secondary-action" onClick={onReconcile} disabled={writeBlocked}>{t($ => $.campaign.reconcile)}</button>
              ) : null}
            </>
          ) : (
            <div className="application-evaluation-campaign-create">
              <label><span>{t($ => $.campaign.apiKeyInput)}</span><input value={quotaAPIKeyId} onChange={(event) => onQuotaAPIKeyIdChange(event.target.value)} placeholder="key_…" disabled={operationPending} /></label>
              <label><span>{t($ => $.campaign.campaignKey)}</span><input value={clientCampaignKey} onChange={(event) => onClientCampaignKeyChange(event.target.value)} disabled={operationPending} /></label>
              {!confirmed ? <button type="button" onClick={onReview} disabled={writeBlocked || !plan || !version || !quotaAPIKeyId.trim()}>{t($ => $.campaign.reviewExecution)}</button> : (
                <div className="application-evaluation-confirmation">
                  <span>{t($ => $.campaign.confirmExecution)}</span>
                  <strong>{version?.name}</strong>
                  <p>{t($ => $.campaign.executionConfirmation, { count: version?.items.length ?? 0, version: version?.planVersion ?? 0, keyId: quotaAPIKeyId })}</p>
                  <p>{t($ => $.campaign.executionBoundary)}</p>
                  <div><button type="button" className="secondary-action" onClick={onCancel} disabled={operationPending}>{t($ => $.campaign.cancel)}</button><button type="button" onClick={onConfirm} disabled={writeBlocked}>{t($ => $.campaign.confirmCampaign)}</button></div>
                </div>
              )}
            </div>
          )}
        </div>
        <aside className="application-evaluation-campaign-handoff-rail" aria-label={t($ => $.campaign.handoffReadiness)}>
          <header>
            <span>{t($ => $.campaign.evidenceHandoff)}</span>
            <strong>{terminalCampaigns.length >= 2 ? t($ => $.campaign.pairReady) : t($ => $.campaign.waiting)}</strong>
          </header>
          <strong>{terminalCampaigns.length} / 2</strong>
          <span>{t($ => $.campaign.succeededCampaigns)}</span>
          <dl>
            <div><dt>{t($ => $.campaign.selected)}</dt><dd>{campaignState(t, selectedCampaign?.state === "succeeded" ? "checkpointed" : "not_terminal")}</dd></div>
            <div><dt>{t($ => $.campaign.pairReview)}</dt><dd>{campaignState(t, terminalCampaigns.length >= 2 ? "available" : "blocked")}</dd></div>
            <div><dt>{t($ => $.campaign.caseSuite)}</dt><dd>{t($ => $.campaign.explicitOnly)}</dd></div>
          </dl>
          <p>{t($ => $.campaign.handoffBoundary)}</p>
          <button type="button" onClick={onOpenHandoff} disabled={terminalCampaigns.length < 2}>{t($ => $.campaign.reviewHandoff)}</button>
        </aside>
      </div>
    </section>
  );
}

function CampaignItems({ campaign }: { campaign: ApplicationEvaluationCampaign }) {
  const { t } = useTranslation("evaluation");
  return (
    <div className="application-evaluation-items">
      <div className="application-evaluation-item-header"><span>{t($ => $.campaign.item)}</span><span>{t($ => $.campaign.runRef)}</span><span>{t($ => $.campaign.state)}</span><span>{t($ => $.campaign.itemBoundary)}</span></div>
      {campaign.items.map((item) => (
        <div key={item.itemKey} className="application-evaluation-item">
          <strong>{item.itemKey}</strong><code>{item.runId || t($ => $.campaign.pendingRun)}</code><em className={item.state}>{campaignState(t, item.state)}</em><small>{item.failureBoundary || (item.state === "succeeded" ? t($ => $.campaign.durableTerminal) : t($ => $.campaign.zeroSideEffects))}</small>
        </div>
      ))}
    </div>
  );
}


function shortRef(value: string): string {
  return value.length <= 22 ? value || "—" : `${value.slice(0, 11)}…${value.slice(-8)}`;
}
