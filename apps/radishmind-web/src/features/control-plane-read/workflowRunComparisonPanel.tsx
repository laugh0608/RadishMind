import { useEffect, useState } from "react";

import type { WorkflowExecutorConsumerConfig } from "./workflowExecutorConsumer.ts";
import { compareWorkflowRuns, type WorkflowRunComparison } from "./workflowRunComparisonConsumer.ts";

export default function WorkflowRunComparisonPanel({ applicationId, baselineRunId, candidateRunId, config }: { applicationId: string; baselineRunId: string; candidateRunId: string; config: WorkflowExecutorConsumerConfig }) {
  const [comparison, setComparison] = useState<WorkflowRunComparison | null>(null);
  const [failure, setFailure] = useState("");
  useEffect(() => {
    let active = true;
    setComparison(null); setFailure("");
    void compareWorkflowRuns(applicationId, baselineRunId, candidateRunId, config).then((value) => { if (active) setComparison(value); }).catch((error) => { if (active) setFailure(error instanceof Error ? error.message : "Workflow run comparison failed."); });
    return () => { active = false; };
  }, [applicationId, baselineRunId, candidateRunId, config]);
  if (failure) return <article className="workflow-run-comparison"><p className="failure-summary">{failure}</p></article>;
  if (!comparison) return <article className="workflow-run-comparison"><p>Comparing durable run records…</p></article>;
  const forbidden = comparison.baseline.sideEffects.toolCalls + comparison.baseline.sideEffects.confirmationCalls + comparison.baseline.sideEffects.businessWrites + comparison.baseline.sideEffects.replayWrites + comparison.candidate.sideEffects.toolCalls + comparison.candidate.sideEffects.confirmationCalls + comparison.candidate.sideEffects.businessWrites + comparison.candidate.sideEffects.replayWrites;
  return <article className="workflow-run-comparison" aria-label="Workflow run regression comparison">
    <div className="card-title-row"><div><p className="eyebrow">Regression review</p><h4>{comparison.classification}</h4></div><span className={`status-badge ${comparison.classification === "regression" ? "status-warning" : "status-good"}`}>{comparison.comparisonState}</span></div>
    <p className="route-path">{comparison.baseline.runId} → {comparison.candidate.runId}</p>
    <dl className="tenant-meta"><div><dt>Status</dt><dd>{comparison.baseline.status} → {comparison.candidate.status}</dd></div><div><dt>Duration delta</dt><dd>{comparison.durationDeltaMs} ms</dd></div><div><dt>Provider calls delta</dt><dd>{comparison.providerCallDelta}</dd></div><div><dt>Forbidden side effects</dt><dd>{forbidden}</dd></div></dl>
    <div className="workflow-run-comparison-findings">{comparison.findings.map((finding) => <span className={`status-badge ${finding.severity === "review_required" ? "status-warning" : "status-neutral"}`} key={finding.code}>{finding.code}</span>)}</div>
    {comparison.retrieval ? <section className="workflow-run-comparison-retrieval" aria-label="Retrieval evidence comparison">
      <div className="card-title-row"><div><p className="eyebrow">{comparison.retrieval.runProfile}</p><h5>Immutable retrieval binding</h5></div><span className="status-badge status-neutral">metadata only</span></div>
      <dl className="tenant-meta">
        <div><dt>Snapshot</dt><dd>{comparison.retrieval.snapshotId} · v{comparison.retrieval.snapshotVersion}</dd></div>
        <div><dt>Profile</dt><dd>{comparison.retrieval.profileId} · v{comparison.retrieval.profileVersion}</dd></div>
        <div><dt>Query</dt><dd>{comparison.retrieval.queryBytes} bytes · {comparison.retrieval.queryDigest}</dd></div>
        <div><dt>Attempt</dt><dd>{comparison.retrieval.baselineAttemptStatus} → {comparison.retrieval.candidateAttemptStatus}</dd></div>
        <div><dt>Candidates</dt><dd>{comparison.retrieval.baselineCandidateCount} → {comparison.retrieval.candidateCandidateCount} ({comparison.retrieval.candidateCountDelta >= 0 ? "+" : ""}{comparison.retrieval.candidateCountDelta})</dd></div>
        <div><dt>Selected / citations</dt><dd>{comparison.retrieval.baselineSelectedCount} → {comparison.retrieval.candidateSelectedCount} / {comparison.retrieval.baselineCitationCount} → {comparison.retrieval.candidateCitationCount}</dd></div>
        <div><dt>Context delta</dt><dd>{comparison.retrieval.contextBytesDelta} bytes</dd></div>
        <div><dt>Retrieval latency</dt><dd>{comparison.retrieval.latencyDeltaMs} ms</dd></div>
      </dl>
      <div className="workflow-run-comparison-findings">
        {comparison.retrieval.citationRemovedRefs.map((ref) => <span className="status-badge status-warning" key={`removed-${ref}`}>citation removed · {ref}</span>)}
        {comparison.retrieval.citationAddedRefs.map((ref) => <span className="status-badge status-neutral" key={`added-${ref}`}>citation added · {ref}</span>)}
      </div>
      <div className="workflow-run-history-node-list">{comparison.retrieval.fragments.map((fragment) => <div className={`workflow-run-history-node-row ${fragment.change === "unchanged" ? "" : "is-failed"}`} key={fragment.fragmentRef}><span><strong>{fragment.fragmentRef}</strong><small>{fragment.sourceType}{fragment.isOfficial ? " · official" : ""}</small></span><span><small>Rank</small><strong>{fragment.baselineRank || "absent"} → {fragment.candidateRank || "absent"}</strong></span><span><small>Change</small><strong>{fragment.change}</strong></span><code>{fragment.contentDigest}</code></div>)}</div>
      <p className="boundary-note">Snapshot {comparison.retrieval.snapshotDigest} · profile {comparison.retrieval.profileDigest}. Query、fragment 正文、prompt packet、answer 和模型原始响应均不在比较响应中。</p>
    </section> : null}
    <div className="workflow-run-history-node-list">{comparison.nodes.map((node) => <div className={`workflow-run-history-node-row ${node.change === "changed" || node.change === "added" || node.change === "removed" ? "is-failed" : ""}`} key={node.nodeId}><span><strong>{node.nodeId}</strong><small>{node.nodeType} · {node.change}</small></span><span><small>Status</small><strong>{node.baselineStatus || "absent"} → {node.candidateStatus || "absent"}</strong></span><span><small>Duration delta</small><strong>{node.durationDeltaMs} ms</strong></span></div>)}</div>
    <p className="boundary-note">Review action: {comparison.recommendedReviewAction || "none"}. Comparison is read-only; replay and resume remain unavailable.</p>
  </article>;
}
