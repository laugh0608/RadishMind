import "../../i18n/runEvidenceResources.ts";
import { useTranslation } from "react-i18next";
import { useLocalePreference } from "../../i18n/LocaleProvider.tsx";
import { formatDisplayNumber } from "../../i18n/formatters.ts";
import { runEvidenceState, runEvidenceAction, runEvidenceFinding, runEvidenceNodeType, runEvidenceSourceType, runComparisonFailure } from "./runEvidenceMessages.ts";
import { useEffect, useState } from "react";

import type { WorkflowExecutorConsumerConfig } from "./workflowExecutorConsumer.ts";
import { compareWorkflowRuns, WorkflowRunComparisonError, type WorkflowRunComparison } from "./workflowRunComparisonConsumer.ts";

export default function WorkflowRunComparisonPanel({ applicationId, baselineRunId, candidateRunId, config }: { applicationId: string; baselineRunId: string; candidateRunId: string; config: WorkflowExecutorConsumerConfig }) {
  const { t } = useTranslation("evaluation");
  const { locale } = useLocalePreference();
  const number = (value: number) => formatDisplayNumber(value, locale) ?? t($ => $.evidence.unknown);
  const milliseconds = (value: number) => t($ => $.evidence.milliseconds, { value: number(value) });
  const bytes = (value: number) => t($ => $.evidence.bytes, { value: number(value) });
  const [retry, setRetry] = useState(0);
  const [comparison, setComparison] = useState<WorkflowRunComparison | null>(null);
  const [failure, setFailure] = useState("");
  useEffect(() => {
    let active = true;
    setComparison(null); setFailure("");
    void compareWorkflowRuns(applicationId, baselineRunId, candidateRunId, config).then((value) => { if (active) setComparison(value); }).catch((error) => { if (active) setFailure(error instanceof WorkflowRunComparisonError ? error.code : "unknown"); });
    return () => { active = false; };
  }, [applicationId, baselineRunId, candidateRunId, config, retry]);
  if (failure) return <article className="workflow-run-comparison" aria-label={t($ => $.evidence.comparisonRegion)}><p className="failure-summary" role="alert">{runComparisonFailure(t, failure)} <code>{failure}</code></p><button type="button" onClick={() => setRetry(value => value + 1)}>{t($ => $.evidence.retryComparison)}</button></article>;
  if (!comparison) return <article className="workflow-run-comparison" aria-label={t($ => $.evidence.comparisonRegion)}><p role="status">{t($ => $.evidence.loading)}</p></article>;
  const forbidden = comparison.baseline.sideEffects.toolCalls + comparison.baseline.sideEffects.confirmationCalls + comparison.baseline.sideEffects.businessWrites + comparison.baseline.sideEffects.replayWrites + comparison.candidate.sideEffects.toolCalls + comparison.candidate.sideEffects.confirmationCalls + comparison.candidate.sideEffects.businessWrites + comparison.candidate.sideEffects.replayWrites;
  return <article className="workflow-run-comparison" aria-label={t($ => $.evidence.comparisonRegion)}>
    <div className="card-title-row"><div><p className="eyebrow">{t($ => $.evidence.regressionReview)}</p><h4>{t($ => $.evidence.classifications[comparison.classification])}</h4></div><span className={`status-badge ${comparison.classification === "regression" || comparison.classification === "inconclusive" ? "status-warning" : comparison.classification === "improvement" ? "status-good" : "status-neutral"}`}>{t($ => $.evidence.comparisonStates[comparison.comparisonState])}</span></div>
    <p className="route-path">{comparison.baseline.runId} → {comparison.candidate.runId}</p>
    <dl className="tenant-meta"><div><dt>{t($ => $.evidence.status)}</dt><dd>{runEvidenceState(t, comparison.baseline.status)} → {runEvidenceState(t, comparison.candidate.status)}</dd></div><div><dt>{t($ => $.evidence.durationDelta)}</dt><dd>{milliseconds(comparison.durationDeltaMs)}</dd></div><div><dt>{t($ => $.evidence.providerDelta)}</dt><dd>{number(comparison.providerCallDelta)}</dd></div><div><dt>{t($ => $.evidence.forbiddenEffects)}</dt><dd>{number(forbidden)}</dd></div></dl>
    <div className="workflow-run-comparison-findings">{comparison.findings.map((finding) => <span className={`status-badge ${finding.severity === "review_required" ? "status-warning" : "status-neutral"}`} key={finding.code}>{runEvidenceFinding(t, finding.code)} <code>{finding.code}</code></span>)}</div>
    {comparison.retrieval ? <section className="workflow-run-comparison-retrieval" aria-label={t($ => $.evidence.retrievalRegion)}>
      <div className="card-title-row"><div><p className="eyebrow">{comparison.retrieval.runProfile}</p><h5>{comparison.retrieval.baselineAuthority ? t($ => $.evidence.ragAuthority) : t($ => $.evidence.binding)}</h5></div><span className="status-badge status-neutral">{t($ => $.evidence.metadataOnly)}</span></div>
      {comparison.retrieval.baselineAuthority && comparison.retrieval.candidateAuthority ? <div className="workflow-rag-authority-comparison">
        {[{ id: "baseline", label: t($ => $.evidence.baselineAuthority), value: comparison.retrieval.baselineAuthority }, { id: "candidate", label: t($ => $.evidence.candidateAuthority), value: comparison.retrieval.candidateAuthority }].map(({ id, label, value }) => <section key={id}>
          <h6>{label}</h6>
          <dl className="tenant-meta">
            <div><dt>{t($ => $.evidence.assignment)}</dt><dd>{value.assignmentId} · v{value.assignmentVersion}</dd></div>
            <div><dt>{t($ => $.evidence.publishCandidate)}</dt><dd>{value.publishCandidateId} · {t($ => $.evidence.reviewVersion, { version: value.publishReviewVersion })}</dd></div>
            <div><dt>{t($ => $.evidence.draftBinding)}</dt><dd>{value.draftId} · v{value.draftVersion} / {value.bindingId}</dd></div>
            <div><dt>{t($ => $.evidence.snapshot)}</dt><dd>{value.snapshotId} · v{value.snapshotVersion}</dd></div>
            <div><dt>{t($ => $.evidence.profile)}</dt><dd>{value.profileId} · v{value.profileVersion}</dd></div>
            <div><dt>{t($ => $.evidence.contract)}</dt><dd>{value.configuredProtocol} · {value.configuredModel}</dd></div>
          </dl>
        </section>)}
      </div> : null}
      <dl className="tenant-meta">
        {!comparison.retrieval.baselineAuthority ? <div><dt>{t($ => $.evidence.snapshot)}</dt><dd>{comparison.retrieval.snapshotId} · v{comparison.retrieval.snapshotVersion}</dd></div> : null}
        {!comparison.retrieval.baselineAuthority ? <div><dt>{t($ => $.evidence.profile)}</dt><dd>{comparison.retrieval.profileId} · v{comparison.retrieval.profileVersion}</dd></div> : null}
        <div><dt>{t($ => $.evidence.query)}</dt><dd>{bytes(comparison.retrieval.queryBytes)} · {comparison.retrieval.queryDigest}</dd></div>
        <div><dt>{t($ => $.evidence.attempt)}</dt><dd>{runEvidenceState(t, comparison.retrieval.baselineAttemptStatus)} → {runEvidenceState(t, comparison.retrieval.candidateAttemptStatus)}</dd></div>
        <div><dt>{t($ => $.evidence.candidates)}</dt><dd>{number(comparison.retrieval.baselineCandidateCount)} → {number(comparison.retrieval.candidateCandidateCount)} ({comparison.retrieval.candidateCountDelta >= 0 ? "+" : ""}{number(comparison.retrieval.candidateCountDelta)})</dd></div>
        <div><dt>{t($ => $.evidence.selectedCitations)}</dt><dd>{number(comparison.retrieval.baselineSelectedCount)} → {number(comparison.retrieval.candidateSelectedCount)} / {number(comparison.retrieval.baselineCitationCount)} → {number(comparison.retrieval.candidateCitationCount)}</dd></div>
        <div><dt>{t($ => $.evidence.contextDelta)}</dt><dd>{bytes(comparison.retrieval.contextBytesDelta)}</dd></div>
        <div><dt>{t($ => $.evidence.retrievalLatency)}</dt><dd>{milliseconds(comparison.retrieval.latencyDeltaMs)}</dd></div>
      </dl>
      <div className="workflow-run-comparison-findings">
        {comparison.retrieval.citationRemovedRefs.map((ref) => <span className="status-badge status-warning" key={`removed-${ref}`}>{t($ => $.evidence.citationRemoved, { reference: ref })}</span>)}
        {comparison.retrieval.citationAddedRefs.map((ref) => <span className="status-badge status-neutral" key={`added-${ref}`}>{t($ => $.evidence.citationAdded, { reference: ref })}</span>)}
      </div>
      <div className="workflow-run-history-node-list">{comparison.retrieval.fragments.map((fragment) => <div className={`workflow-run-history-node-row ${fragment.change === "unchanged" ? "" : "is-failed"}`} key={fragment.fragmentRef}><span><strong>{fragment.fragmentRef}</strong><small>{runEvidenceSourceType(t, fragment.sourceType)}{fragment.isOfficial ? ` · ${t($ => $.evidence.official)}` : ""}</small></span><span><small>{t($ => $.evidence.rank)}</small><strong>{fragment.baselineRank ? number(fragment.baselineRank) : t($ => $.evidence.absent)} → {fragment.candidateRank ? number(fragment.candidateRank) : t($ => $.evidence.absent)}</strong></span><span><small>{t($ => $.evidence.change)}</small><strong>{t($ => $.evidence.changes[fragment.change])}</strong></span><code>{fragment.contentDigest}</code></div>)}</div>
      <p className="boundary-note">{comparison.retrieval.baselineAuthority ? t($ => $.evidence.authorityNote) : t($ => $.evidence.snapshotNote, { snapshot: comparison.retrieval.snapshotDigest, profile: comparison.retrieval.profileDigest })} {t($ => $.evidence.privacyNote)}</p>
    </section> : null}
    <div className="workflow-run-history-node-list">{comparison.nodes.map((node) => <div className={`workflow-run-history-node-row ${node.change === "changed" || node.change === "added" || node.change === "removed" ? "is-failed" : ""}`} key={node.nodeId}><span><strong>{node.nodeId}</strong><small>{runEvidenceNodeType(t, node.nodeType)} · {t($ => $.evidence.changes[node.change])}</small></span><span><small>{t($ => $.evidence.status)}</small><strong>{runEvidenceState(t, node.baselineStatus)} → {runEvidenceState(t, node.candidateStatus)}</strong></span><span><small>{t($ => $.evidence.durationDelta)}</small><strong>{milliseconds(node.durationDeltaMs)}</strong></span></div>)}</div>
    <p className="boundary-note">{t($ => $.evidence.reviewAction)}: {runEvidenceAction(t, comparison.recommendedReviewAction)} <code>{comparison.recommendedReviewAction}</code>. {t($ => $.evidence.readOnly)}</p>
  </article>;
}
