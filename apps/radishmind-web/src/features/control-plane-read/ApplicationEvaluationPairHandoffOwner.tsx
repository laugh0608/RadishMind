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
  const review = envelope?.review ?? null;
  const handoff = envelope?.handoff ?? review?.existingHandoff ?? null;
  return (
    <section className="application-evaluation-pair-owner" id={task === "pair" ? "application-evaluation-pair" : "application-evaluation-handoff"}>
      <header>
        <div><p className="eyebrow">Current owner · {task === "pair" ? "Pair Review" : "Handoff"}</p><h4>{task === "pair" ? "Exact campaign pair" : "Case and Suite evidence handoff"}</h4><p>{task === "pair" ? "Compare two succeeded campaigns from the same immutable plan version." : "Materialize exact Case versions first, then one Suite; partial evidence remains append-only."}</p></div>
        <span className={`status-badge ${handoff?.state === "complete" ? "good" : "neutral"}`}>{handoff?.state ?? (review ? "preview ready" : "selection required")}</span>
      </header>
      <div className="application-evaluation-pair-selectors">
        <label><span>Baseline campaign</span><select value={baselineCampaignId} onChange={(event) => onBaselineChange(event.target.value)} disabled={operationPending}><option value="">Choose baseline</option>{campaigns.map((campaign) => <option key={campaign.campaignId} value={campaign.campaignId}>{campaign.clientCampaignKey} · {campaign.campaignId}</option>)}</select></label>
        <label><span>Candidate campaign</span><select value={candidateCampaignId} onChange={(event) => onCandidateChange(event.target.value)} disabled={operationPending}><option value="">Choose candidate</option>{campaigns.map((campaign) => <option key={campaign.campaignId} value={campaign.campaignId}>{campaign.clientCampaignKey} · {campaign.campaignId}</option>)}</select></label>
        <button type="button" onClick={onPreview} disabled={operationPending || !baselineCampaignId || !candidateCampaignId || baselineCampaignId === candidateCampaignId}>Preview exact pair</button>
      </div>
      {review ? (
        <>
          <div className="application-evaluation-pair-summary"><div><span>Plan</span><strong>{review.planName} · v{review.planVersion}</strong></div><div><span>Expected matches</span><strong>{review.expectedMatches}</strong></div><div><span>Mismatches</span><strong>{review.expectedMismatches}</strong></div><div><span>Profile</span><strong>{review.executionProfile}</strong></div></div>
          <div className="application-evaluation-pair-items">
            {review.items.map((item) => <article key={item.itemKey} className={item.expectationMatched ? "is-match" : "is-mismatch"}><header><strong>{item.name}</strong><span>{item.expectationMatched ? "expected" : "mismatch"}</span></header><p><code>{item.baselineRunId}</code><span>→</span><code>{item.candidateRunId}</code></p><dl><div><dt>Expected</dt><dd>{item.expectedClassification}</dd></div><div><dt>Actual</dt><dd>{item.actualClassification}</dd></div><div><dt>Comparison</dt><dd>{item.comparison?.comparisonState ?? "unavailable"}</dd></div></dl></article>)}
          </div>
          {task === "handoff" ? (
            handoff ? <HandoffEvidence handoff={handoff} /> : !confirmed ? (
              <div className="application-evaluation-handoff-gate"><div><span>EXPLICIT HANDOFF</span><strong>Materialize exact Case versions and one Suite</strong><p>No candidate is approved, activated, released or deployed by this action.</p></div><button type="button" onClick={onReviewHandoff} disabled={writeBlocked}>Review handoff</button></div>
            ) : (
              <div className="application-evaluation-confirmation"><span>CONFIRM EVIDENCE MATERIALIZATION</span><strong>{review.items.length} exact campaign pairs</strong><p>Expected baseline record version {envelope?.currentBaselineRecordVersion}; candidate version {envelope?.currentCandidateRecordVersion}. Every completed Case is checkpointed before Suite creation.</p><p>Partial success remains durable and is not deleted or automatically retried.</p><div><button type="button" className="secondary-action" onClick={onCancelHandoff}>Cancel</button><button type="button" onClick={onConfirmHandoff} disabled={writeBlocked}>Confirm handoff</button></div></div>
            )
          ) : <p className="application-evaluation-pair-boundary">Pair Preview is read-only. Open Handoff for explicit Case / Suite evidence materialization.</p>}
        </>
      ) : <div className="application-evaluation-pair-empty"><strong>{campaigns.length < 2 ? "Two succeeded campaigns are required" : "Preview the selected pair"}</strong><span>Only exact Run references are compared. No execution, replay or release occurs.</span></div>}
    </section>
  );
}

function HandoffEvidence({ handoff }: { handoff: NonNullable<ApplicationEvaluationPairEnvelope["handoff"]> }) {
  return <div className={`application-evaluation-handoff-evidence ${handoff.state}`}><header><div><span>HANDOFF {handoff.state.toUpperCase()}</span><strong>{handoff.caseRefs.length} exact Case refs</strong></div><em>{handoff.suiteId || "Suite pending"}</em></header><div>{handoff.caseRefs.map((ref) => <code key={`${ref.caseId}:${ref.version}`}>{ref.caseId} · v{ref.version}</code>)}</div><p>Candidate campaign remains the single durable handoff anchor. Existing evidence is append-only.</p></div>;
}
