import "../../i18n/evaluationPairResources.ts";
import { useTranslation } from "react-i18next";
import { campaignState } from "./applicationEvaluationMessages.ts";
import type {
  ApplicationEvaluationCampaign,
  ApplicationEvaluationPairEnvelope,
} from "./applicationEvaluationCampaignConsumer.ts";

export default function ApplicationEvaluationPairHandoffOwner({
  task,
  campaigns,
  baselineCampaignId,
  candidateCampaignId,
  envelope,
  confirmed,
  writeBlocked,
  operationPending,
  onBaselineChange,
  onCandidateChange,
  onPreview,
  onReviewHandoff,
  onCancelHandoff,
  onConfirmHandoff,
}: {
  task: "pair" | "handoff";
  campaigns: ApplicationEvaluationCampaign[];
  baselineCampaignId: string;
  candidateCampaignId: string;
  envelope: ApplicationEvaluationPairEnvelope | null;
  confirmed: boolean;
  writeBlocked: boolean;
  operationPending: boolean;
  onBaselineChange: (value: string) => void;
  onCandidateChange: (value: string) => void;
  onPreview: () => void;
  onReviewHandoff: () => void;
  onCancelHandoff: () => void;
  onConfirmHandoff: () => void;
}) {
  const { t } = useTranslation("evaluation");
  const review = envelope?.review ?? null;
  const handoff = envelope?.handoff ?? review?.existingHandoff ?? null;
  return (
    <section className="application-evaluation-pair-owner" id={task === "pair" ? "application-evaluation-pair" : "application-evaluation-handoff"}>
      <header>
        <div><p className="eyebrow">{t($ => $.campaign.pairOwner, { task: task === "pair" ? t($ => $.campaign.pair) : t($ => $.campaign.handoff) })}</p><h4>{task === "pair" ? t($ => $.campaign.exactPair) : t($ => $.campaign.handoffTitle)}</h4><p>{task === "pair" ? t($ => $.campaign.pairIntro) : t($ => $.campaign.handoffIntro)}</p></div>
        <span className={`status-badge ${handoff?.state === "complete" ? "good" : "neutral"}`}>{campaignState(t, handoff?.state ?? (review ? "preview_ready" : "selection_required"))}</span>
      </header>
      <div className="application-evaluation-pair-selectors">
        <label><span>{t($ => $.campaign.baselineCampaign)}</span><select value={baselineCampaignId} onChange={(event) => onBaselineChange(event.target.value)} disabled={operationPending}><option value="">{t($ => $.campaign.chooseBaseline)}</option>{campaigns.map((campaign) => <option key={campaign.campaignId} value={campaign.campaignId}>{campaign.clientCampaignKey} · {campaign.campaignId}</option>)}</select></label>
        <label><span>{t($ => $.campaign.candidateCampaign)}</span><select value={candidateCampaignId} onChange={(event) => onCandidateChange(event.target.value)} disabled={operationPending}><option value="">{t($ => $.campaign.chooseCandidate)}</option>{campaigns.map((campaign) => <option key={campaign.campaignId} value={campaign.campaignId}>{campaign.clientCampaignKey} · {campaign.campaignId}</option>)}</select></label>
        <button type="button" onClick={onPreview} disabled={operationPending || !baselineCampaignId || !candidateCampaignId || baselineCampaignId === candidateCampaignId}>{t($ => $.campaign.previewPair)}</button>
      </div>
      {review ? (
        <>
          <div className="application-evaluation-pair-summary"><div><span>{t($ => $.campaign.plan)}</span><strong>{review.planName} · v{review.planVersion}</strong></div><div><span>{t($ => $.campaign.expectedMatches)}</span><strong>{review.expectedMatches}</strong></div><div><span>{t($ => $.campaign.mismatches)}</span><strong>{review.expectedMismatches}</strong></div><div><span>{t($ => $.campaign.profile)}</span><strong>{review.executionProfile}</strong></div></div>
          <div className="application-evaluation-pair-items">
            {review.items.map((item) => <article key={item.itemKey} className={item.expectationMatched ? "is-match" : "is-mismatch"}><header><strong>{item.name}</strong><span>{campaignState(t, item.expectationMatched ? "expected" : "mismatch")}</span></header><p><code>{item.baselineRunId}</code><span>→</span><code>{item.candidateRunId}</code></p><dl><div><dt>{t($ => $.campaign.expected)}</dt><dd>{campaignState(t, item.expectedClassification)}</dd></div><div><dt>{t($ => $.campaign.actual)}</dt><dd>{campaignState(t, item.actualClassification)}</dd></div><div><dt>{t($ => $.campaign.comparison)}</dt><dd>{campaignState(t, item.comparison?.comparisonState ?? "unavailable")}</dd></div></dl></article>)}
          </div>
          {task === "handoff" ? (
            handoff ? <HandoffEvidence handoff={handoff} /> : !confirmed ? (
              <div className="application-evaluation-handoff-gate"><div><span>{t($ => $.campaign.explicitHandoff)}</span><strong>{t($ => $.campaign.materialize)}</strong><p>{t($ => $.campaign.materializeBoundary)}</p></div><button type="button" onClick={onReviewHandoff} disabled={writeBlocked}>{t($ => $.campaign.reviewHandoff)}</button></div>
            ) : (
              <div className="application-evaluation-confirmation"><span>{t($ => $.campaign.confirmMaterialize)}</span><strong>{t($ => $.campaign.pairCount, { count: review.items.length })}</strong><p>{t($ => $.campaign.pairConfirmation, { baseline: envelope?.currentBaselineRecordVersion ?? 0, candidate: envelope?.currentCandidateRecordVersion ?? 0 })}</p><p>{t($ => $.campaign.partialBoundary)}</p><div><button type="button" className="secondary-action" onClick={onCancelHandoff} disabled={operationPending}>{t($ => $.campaign.cancel)}</button><button type="button" onClick={onConfirmHandoff} disabled={writeBlocked}>{t($ => $.campaign.confirmHandoff)}</button></div></div>
            )
          ) : <p className="application-evaluation-pair-boundary">{t($ => $.campaign.pairBoundary)}</p>}
        </>
      ) : <div className="application-evaluation-pair-empty"><strong>{campaigns.length < 2 ? t($ => $.campaign.twoCampaigns) : t($ => $.campaign.previewSelection)}</strong><span>{t($ => $.campaign.pairEmptyBoundary)}</span></div>}
    </section>
  );
}

function HandoffEvidence({ handoff }: { handoff: NonNullable<ApplicationEvaluationPairEnvelope["handoff"]> }) {
  const { t } = useTranslation("evaluation");
  return <div className={`application-evaluation-handoff-evidence ${handoff.state}`}><header><div><span>{t($ => $.campaign.handoffState, { state: campaignState(t, handoff.state) })}</span><strong>{t($ => $.campaign.caseRefCount, { count: handoff.caseRefs.length })}</strong></div><em>{handoff.suiteId || t($ => $.campaign.suitePending)}</em></header><div>{handoff.caseRefs.map((ref) => <code key={`${ref.caseId}:${ref.version}`}>{ref.caseId} · v{ref.version}</code>)}</div><p>{t($ => $.campaign.handoffAnchor)}</p></div>;
}
