import { useEffect, useMemo, useState } from "react";

import {
  initialApplicationConfigurationDraftListState,
  listApplicationConfigurationDrafts,
  readApplicationConfigurationDraftConfig,
  type ApplicationConfigurationBaseline,
  type ApplicationConfigurationDraftListState,
} from "./applicationConfigurationDraftConsumer.ts";
import { requestApplicationApiIntegrationDraftHandoff } from "./applicationApiIntegrationEvents.ts";
import {
  createApplicationPublishCandidate,
  initialApplicationPublishListState,
  initialApplicationPublishOperationState,
  listApplicationPublishCandidates,
  parseApplicationPublishEvidence,
  readApplicationPublishCandidate,
  readApplicationPublishCandidateConfig,
  reviewApplicationPublishCandidate,
  validateApplicationPublishReview,
  type ApplicationPublishCandidate,
  type ApplicationPublishCandidateListState,
  type ApplicationPublishDecision,
  type ApplicationPublishOperationState,
} from "./applicationPublishCandidateConsumer.ts";
import { requestGatewayRequestHistoryReview, requestModelGatewayPlaygroundHandoff } from "./modelGatewayPlaygroundEvents.ts";

const publishConfig = readApplicationPublishCandidateConfig();
const draftConfig = readApplicationConfigurationDraftConfig();

export default function ApplicationPublishCandidatePanel({ baseline }: { baseline: ApplicationConfigurationBaseline }) {
  const [draftList, setDraftList] = useState<ApplicationConfigurationDraftListState>(() => initialApplicationConfigurationDraftListState(draftConfig));
  const [candidateList, setCandidateList] = useState<ApplicationPublishCandidateListState>(() => initialApplicationPublishListState(publishConfig));
  const [candidate, setCandidate] = useState<ApplicationPublishCandidate | null>(null);
  const [operation, setOperation] = useState<ApplicationPublishOperationState>(() => initialApplicationPublishOperationState(publishConfig));
  const [selectedDraftId, setSelectedDraftId] = useState("");
  const [candidateId, setCandidateId] = useState(() => newCandidateId(baseline.applicationId));
  const [evidenceText, setEvidenceText] = useState("");
  const [decision, setDecision] = useState<ApplicationPublishDecision>("approve");
  const [reviewReason, setReviewReason] = useState("");

  useEffect(() => {
    setDraftList(initialApplicationConfigurationDraftListState(draftConfig));
    setCandidateList(initialApplicationPublishListState(publishConfig));
    setCandidate(null);
    setOperation(initialApplicationPublishOperationState(publishConfig));
    setSelectedDraftId("");
    setCandidateId(newCandidateId(baseline.applicationId));
    setEvidenceText("");
    setDecision("approve");
    setReviewReason("");
  }, [baseline.applicationId]);

  const enabled = publishConfig.mode === "dev_application_publish_http" && draftConfig.mode === "dev_application_draft_http";
  const selectedDraft = draftList.summaries.find((summary) => summary.draftId === selectedDraftId) ?? null;
  const evidence = useMemo(() => parseApplicationPublishEvidence(evidenceText), [evidenceText]);
  const reviewFailure = useMemo(() => validateApplicationPublishReview(decision, reviewReason), [decision, reviewReason]);
  const canReview = candidate?.candidateState === "pending_review" || candidate?.candidateState === "approved" && decision === "withdraw";

  async function loadDrafts() {
    if (!enabled) return;
    setDraftList((current) => ({ ...current, status: "loading", summaries: [], failureCode: "", summary: "Loading saved valid application drafts." }));
    const next = await listApplicationConfigurationDrafts(draftConfig, baseline.applicationId);
    setDraftList(next);
    const firstValid = next.summaries.find((summary) => summary.validationState === "valid");
    setSelectedDraftId(firstValid?.draftId ?? "");
  }

  async function refreshCandidates() {
    if (!enabled) return;
    setCandidateList((current) => ({ ...current, status: "loading", summaries: [], failureCode: "", summary: "Loading publish candidates." }));
    setCandidateList(await listApplicationPublishCandidates(publishConfig, baseline.applicationId));
  }

  async function createCandidate() {
    if (!enabled || !selectedDraft || selectedDraft.validationState !== "valid" || evidence.failureCode) return;
    setOperation((current) => ({ ...current, status: "creating", summary: "Creating an immutable candidate from the server-side saved draft.", failureCode: "" }));
    const result = await createApplicationPublishCandidate(publishConfig, baseline.applicationId, candidateId, selectedDraft.draftId, selectedDraft.draftVersion, evidence.requestIds);
    setOperation(result.state);
    if (result.candidate) {
      setCandidate(result.candidate);
      setReviewReason("");
      await refreshCandidates();
    }
  }

  async function openCandidate(candidateRef: string) {
    setOperation((current) => ({ ...current, status: "loading", summary: "Loading the immutable publish candidate and current eligibility.", failureCode: "" }));
    const result = await readApplicationPublishCandidate(publishConfig, baseline.applicationId, candidateRef);
    setOperation(result.state);
    if (result.candidate) {
      setCandidate(result.candidate);
      setCandidateId(result.candidate.candidateId);
      setEvidenceText(result.candidate.evidenceRequestIds.join("\n"));
      setReviewReason("");
    }
  }

  async function submitReview() {
    if (!candidate || reviewFailure || !canReview) return;
    setOperation((current) => ({ ...current, status: "reviewing", summary: "Recording an append-only dev/test review decision.", failureCode: "" }));
    const result = await reviewApplicationPublishCandidate(publishConfig, baseline.applicationId, candidate.candidateId, candidate.reviewVersion, decision, reviewReason);
    setOperation(result.state);
    if (result.candidate) {
      setCandidate(result.candidate);
      setReviewReason("");
      await refreshCandidates();
    }
  }

  function openIntegration() {
    if (!candidate) return;
    requestApplicationApiIntegrationDraftHandoff(candidate.applicationId, candidate.configuration.defaultProtocol, candidate.configuration.defaultModel);
    window.location.hash = "application-api-integration";
  }

  function openPlayground() {
    if (!candidate) return;
    requestModelGatewayPlaygroundHandoff(candidate.applicationId, candidate.configuration.defaultProtocol, candidate.configuration.defaultModel);
    window.location.hash = "model-gateway-playground";
  }

  return (
    <section className="application-publish-workspace" id="application-publish-review" aria-labelledby="application-publish-title">
      <div className="section-heading compact-heading">
        <div><p className="eyebrow">Application Publish Governance</p><h4 id="application-publish-title">Candidate, review, drift, and promotion eligibility</h4></div>
        <span className={`status-badge ${candidate?.candidateState === "approved" ? "good" : operation.status.includes("conflict") || operation.status === "failed" ? "bad" : "neutral"}`}>{candidate?.candidateState ?? operation.status}</span>
      </div>

      <div className="application-publish-scope">
        <article><span>Application</span><strong>{baseline.displayName}</strong><code>{baseline.applicationId}</code></article>
        <article><span>Baseline</span><strong>{baseline.updatedAt}</strong><p>Control Plane read truth remains immutable.</p></article>
        <article><span>Promotion</span><strong>disabled</strong><p>Candidate approval never mutates the formal application.</p></article>
      </div>

      <div className="application-publish-layout">
        <article className="application-publish-create">
          <div className="application-api-card-heading"><div><p className="eyebrow">Candidate source</p><h5>Bind an exact saved draft version</h5></div><button type="button" onClick={() => void loadDrafts()} disabled={!enabled || draftList.status === "loading"}>Load saved drafts</button></div>
          <label>Saved valid draft<select value={selectedDraftId} onChange={(event) => setSelectedDraftId(event.target.value)} disabled={!enabled || draftList.summaries.length === 0}><option value="">No saved valid draft selected</option>{draftList.summaries.map((summary) => <option key={summary.draftId} value={summary.draftId} disabled={summary.validationState !== "valid"}>{summary.draftId} · v{summary.draftVersion} · {summary.validationState}</option>)}</select></label>
          <label>Candidate id<input value={candidateId} onChange={(event) => setCandidateId(event.target.value)} maxLength={160} /></label>
          <label>Sanitized Request History refs<textarea value={evidenceText} onChange={(event) => setEvidenceText(event.target.value)} rows={4} placeholder="One request_id per line; no prompts, responses, headers, or credentials." /></label>
          {evidence.failureCode ? <p className="failure-summary">{evidence.failureCode}</p> : <p className="boundary-note">{evidence.requestIds.length} normalized history reference(s); payloads stay in Request History.</p>}
          <button type="button" onClick={() => void createCandidate()} disabled={!enabled || !selectedDraft || selectedDraft.validationState !== "valid" || Boolean(evidence.failureCode) || operation.status === "creating"}>Create immutable candidate</button>
          <p className="boundary-note">The server reloads the saved draft and application baseline. Browser form content cannot replace the candidate snapshot.</p>
        </article>

        <article className="application-publish-review">
          <div className="application-api-card-heading"><div><p className="eyebrow">Review decision</p><h5>{candidate?.candidateId ?? "No candidate selected"}</h5></div><span className="status-badge neutral">review v{candidate?.reviewVersion ?? 0}</span></div>
          <label>Decision<select value={decision} onChange={(event) => setDecision(event.target.value as ApplicationPublishDecision)} disabled={!candidate}><option value="approve">Approve candidate</option><option value="reject">Reject candidate</option><option value="request_changes">Request changes</option><option value="withdraw">Withdraw candidate</option></select></label>
          <label>Review reason<textarea value={reviewReason} onChange={(event) => setReviewReason(event.target.value)} rows={4} maxLength={500} placeholder="Explain the bounded dev/test review decision without credentials or request content." /></label>
          {reviewFailure && reviewReason ? <p className="failure-summary">{reviewFailure}</p> : null}
          <button type="button" onClick={() => void submitReview()} disabled={!enabled || !candidate || !canReview || Boolean(reviewFailure) || operation.status === "reviewing"}>Record review decision</button>
          {operation.failureCode ? <p className="failure-summary">{operation.failureCode}</p> : null}
          <p className="boundary-note">{operation.summary}</p>
          {operation.status === "review_version_conflict" && candidate ? <button type="button" onClick={() => void openCandidate(candidate.candidateId)}>Restore current review version {operation.currentReviewVersion}</button> : null}
        </article>
      </div>

      {candidate ? <CandidateDetail candidate={candidate} baseline={baseline} onIntegration={openIntegration} onPlayground={openPlayground} onHistory={(requestId) => { requestGatewayRequestHistoryReview(requestId, candidate.applicationId); window.location.hash = "model-gateway-request-history"; }} /> : <p className="boundary-note">Create or open a candidate to review its immutable snapshot and blockers.</p>}

      <article className="application-publish-saved">
        <div className="application-api-card-heading"><div><p className="eyebrow">Saved dev/test candidates</p><h5>{candidateList.summary}</h5></div><button type="button" onClick={() => void refreshCandidates()} disabled={!enabled || candidateList.status === "loading"}>Refresh candidates</button></div>
        {candidateList.failureCode ? <p className="failure-summary">{candidateList.failureCode}</p> : null}
        <div className="application-publish-candidate-list">{candidateList.summaries.map((summary) => <button type="button" key={summary.candidateId} onClick={() => void openCandidate(summary.candidateId)}><strong>{summary.candidateId}</strong><span>{summary.candidateState} · review v{summary.reviewVersion}</span><small>draft v{summary.draftVersion} · {summary.promotionBlockers} blocker(s)</small></button>)}</div>
      </article>

      <p className="boundary-note">Offline mode performs no requests. Production auth, formal application repository, publish owner, promotion runtime, API key lifecycle, quota, billing, fallback, load balancing, Workflow tool, confirmation, writeback, replay, and resume remain disabled.</p>
    </section>
  );
}

function CandidateDetail({ candidate, baseline, onIntegration, onPlayground, onHistory }: { candidate: ApplicationPublishCandidate; baseline: ApplicationConfigurationBaseline; onIntegration: () => void; onPlayground: () => void; onHistory: (requestId: string) => void }) {
  const comparison = [
    { field: "display_name", before: baseline.displayName, after: candidate.configuration.displayName },
    { field: "application_kind", before: baseline.applicationKind, after: candidate.configuration.applicationKind },
    { field: "default_protocol", before: "not configured in read model", after: candidate.configuration.defaultProtocol },
    { field: "default_model", before: "not configured in read model", after: candidate.configuration.defaultModel },
  ];
  return <div className="application-publish-detail">
    <article className="application-publish-snapshot"><div className="application-api-card-heading"><div><p className="eyebrow">Immutable snapshot</p><h5>{candidate.draftId} · v{candidate.draftVersion}</h5></div><span className="status-badge neutral">{candidate.schemaVersion}</span></div><code className="application-publish-digest">{candidate.draftDigest}</code><p>{candidate.configuration.description || "No public description."}</p><div className="application-publish-comparison">{comparison.map((item) => <div className={item.before === item.after ? "unchanged" : "changed"} key={item.field}><strong>{item.field}</strong><span>{item.before}</span><span>→</span><span>{item.after}</span></div>)}</div><div className="application-draft-handoff"><button type="button" onClick={onIntegration}>Open API Integration</button><button type="button" onClick={onPlayground}>Test in Playground</button></div></article>
    <article className="application-publish-eligibility"><div className="application-api-card-heading"><div><p className="eyebrow">Promotion eligibility</p><h5>{candidate.promotionEligibility.status}</h5></div><span className="status-badge bad">{candidate.promotionEligibility.blockers.length} blockers</span></div><ul>{candidate.promotionEligibility.blockers.map((blocker) => <li key={blocker.code}><strong>{blocker.code}</strong><p>{blocker.summary}</p></li>)}</ul></article>
    <article className="application-publish-evidence"><div className="application-api-card-heading"><div><p className="eyebrow">Request History references</p><h5>{candidate.evidenceRequestIds.length} sanitized refs</h5></div></div>{candidate.evidenceRequestIds.length ? candidate.evidenceRequestIds.map((requestId) => <button type="button" key={requestId} onClick={() => onHistory(requestId)}><code>{requestId}</code><span>Open exact history detail</span></button>) : <p className="boundary-note">No Gateway request references were attached.</p>}</article>
    <article className="application-publish-review-log"><div className="application-api-card-heading"><div><p className="eyebrow">Append-only review log</p><h5>{candidate.reviews.length} decisions</h5></div></div>{candidate.reviews.length ? candidate.reviews.map((review) => <div key={review.reviewVersion}><strong>v{review.reviewVersion} · {review.decision}</strong><span>{review.reviewerRef} · {review.reviewedAt}</span><p>{review.reason}</p></div>) : <p className="boundary-note">No review decision has been recorded.</p>}</article>
  </div>;
}

function newCandidateId(applicationId: string): string {
  const suffix = (globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random()}`).replaceAll("-", "").slice(0, 10);
  return `publish-${applicationId}-${suffix}`;
}
