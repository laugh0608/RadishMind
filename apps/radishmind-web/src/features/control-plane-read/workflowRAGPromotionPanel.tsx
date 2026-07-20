import { useEffect, useMemo, useState } from "react";

import {
  initialApplicationConfigurationDraftListState,
  listApplicationConfigurationDrafts,
  readApplicationConfigurationDraftConfig,
  type ApplicationConfigurationDraftListState,
} from "./applicationConfigurationDraftConsumer.ts";
import {
  listWorkflowRAGCandidateReviews,
  listWorkflowRAGEvaluationDatasets,
  readWorkflowRAGEvaluationConfig,
  type WorkflowRAGCandidateReviewListResult,
  type WorkflowRAGEvaluationListResult,
} from "./workflowRAGEvaluationDatasetConsumer.ts";
import {
  createWorkflowRAGPromotionCandidate,
  decideWorkflowRAGPromotionCandidate,
  initialWorkflowRAGPromotionListResult,
  initialWorkflowRAGPromotionOperationResult,
  listWorkflowRAGPromotionCandidates,
  readWorkflowRAGPromotionCandidate,
  readWorkflowRAGPromotionConfig,
  workflowRAGPromotionDecisionAllowed,
  type WorkflowRAGPromotionDecision,
  type WorkflowRAGPromotionDetail,
  type WorkflowRAGPromotionListResult,
  type WorkflowRAGPromotionOperationResult,
} from "./workflowRAGPromotionConsumer.ts";
import type { ApplicationDevelopmentOwnerEvidence } from "./applicationDevelopmentReadiness.ts";

const promotionConfig = readWorkflowRAGPromotionConfig();
const evaluationConfig = readWorkflowRAGEvaluationConfig();
const draftConfig = readApplicationConfigurationDraftConfig();

type Props = {
  applicationId: string;
  applicationName: string;
  applicationActive: boolean;
  onEvidenceChange?: (evidence: ApplicationDevelopmentOwnerEvidence) => void;
};

export default function WorkflowRAGPromotionPanel({ applicationId, applicationName, applicationActive, onEvidenceChange }: Props) {
  const [datasets, setDatasets] = useState<WorkflowRAGEvaluationListResult>(() => emptyDatasets());
  const [reviews, setReviews] = useState<WorkflowRAGCandidateReviewListResult>(() => emptyReviews());
  const [drafts, setDrafts] = useState<ApplicationConfigurationDraftListState>(() => initialApplicationConfigurationDraftListState(draftConfig));
  const [promotions, setPromotions] = useState<WorkflowRAGPromotionListResult>(() => initialWorkflowRAGPromotionListResult(promotionConfig));
  const [operation, setOperation] = useState<WorkflowRAGPromotionOperationResult>(() => initialWorkflowRAGPromotionOperationResult(promotionConfig));
  const [selectedDatasetId, setSelectedDatasetId] = useState("");
  const [selectedReviewId, setSelectedReviewId] = useState("");
  const [selectedDraftId, setSelectedDraftId] = useState("");
  const [decision, setDecision] = useState<WorkflowRAGPromotionDecision>("approve");
  const [reason, setReason] = useState("");

  useEffect(() => {
    setDatasets(emptyDatasets());
    setReviews(emptyReviews());
    setDrafts(initialApplicationConfigurationDraftListState(draftConfig));
    setPromotions(initialWorkflowRAGPromotionListResult(promotionConfig));
    setOperation(initialWorkflowRAGPromotionOperationResult(promotionConfig));
    setSelectedDatasetId("");
    setSelectedReviewId("");
    setSelectedDraftId("");
    setDecision("approve");
    setReason("");
  }, [applicationId]);

  const enabled = promotionConfig.mode === "dev_workflow_rag_promotion_http" && evaluationConfig.mode === "dev_workflow_rag_evaluation_http" && draftConfig.mode === "dev_application_draft_http" && applicationActive && Boolean(applicationId);
  const detail = operation.detail;
  const selectedDataset = datasets.resources.find((item) => item.datasetId === selectedDatasetId) ?? null;
  const selectedReview = reviews.reviews.find((item) => item.reviewId === selectedReviewId) ?? null;
  const selectedDraft = drafts.summaries.find((item) => item.draftId === selectedDraftId) ?? null;
  const eligibleReview = selectedReview && selectedDataset && selectedReview.datasetVersion === selectedDataset.latestVersion && selectedReview.datasetDigest === selectedDataset.latestDigest && selectedReview.candidateStatus === "passed" && (selectedReview.conclusion === "improved" || selectedReview.conclusion === "unchanged");
  const reasonFailure = useMemo(() => validateReason(reason), [reason]);
  const canDecide = Boolean(detail && workflowRAGPromotionDecisionAllowed(detail.candidate.candidateState, decision));

  useEffect(() => {
    if (!onEvidenceChange) return;
    const summaryWithBinding = promotions.summaries.find((item) => item.bindingRef && item.candidateState === "approved");
    const binding = detail?.binding ?? summaryWithBinding?.bindingRef ?? null;
    const ownerFailed = operation.status === "failed" || operation.status === "scope_denied" || operation.status === "record_version_conflict" || promotions.status === "failed";
    const blockers = detail?.eligibility.blockers ?? [];
    const available = Boolean(binding && detail?.eligibility.eligible !== false);
    const failureCode = operation.failureCode || promotions.failureCode;
    onEvidenceChange({
      contributionId: "rag_binding",
      status: ownerFailed || blockers.length ? "blocked" : available ? "available" : "incomplete",
      coverage: binding || ownerFailed ? "complete" : "none",
      evidenceRefs: binding ? [{ kind: "binding", id: binding.bindingId, version: binding.bindingVersion }] : [],
      missingEvidence: available ? [] : ["Approve an exact knowledge promotion candidate and immutable binding."],
      blockers: ownerFailed
        ? [{ code: failureCode || "rag_binding_owner_blocked", summary: operation.summary || promotions.summary }]
        : blockers.map((code) => ({ code, summary: "The RAG binding owner reports a current authority or eligibility blocker." })),
      failureCodes: ownerFailed && failureCode ? [failureCode] : [],
    });
  }, [detail, onEvidenceChange, operation, promotions]);

  async function loadSources() {
    if (!enabled) return;
    setDatasets((current) => ({ ...current, status: "empty", resources: [], failureCode: "", summary: "Loading active evaluation datasets." }));
    setDrafts((current) => ({ ...current, status: "loading", summaries: [], failureCode: "", summary: "Loading exact saved drafts." }));
    const [nextDatasets, nextDrafts] = await Promise.all([
      listWorkflowRAGEvaluationDatasets(evaluationConfig, applicationId, "active"),
      listApplicationConfigurationDrafts(draftConfig, applicationId),
    ]);
    setDatasets(nextDatasets);
    setDrafts(nextDrafts);
    const dataset = nextDatasets.resources[0];
    const draft = nextDrafts.summaries.find((item) => item.validationState === "valid");
    setSelectedDatasetId(dataset?.datasetId ?? "");
    setSelectedDraftId(draft?.draftId ?? "");
    setSelectedReviewId("");
    if (dataset) await loadReviews(dataset.datasetId);
  }

  async function loadReviews(datasetId: string) {
    setSelectedDatasetId(datasetId);
    setSelectedReviewId("");
    if (!enabled || !datasetId) {
      setReviews(emptyReviews());
      return;
    }
    setReviews((current) => ({ ...current, status: "empty", reviews: [], failureCode: "", summary: "Loading metadata-only candidate reviews." }));
    const next = await listWorkflowRAGCandidateReviews(evaluationConfig, applicationId, datasetId);
    setReviews(next);
    const review = next.reviews.find((item) => item.candidateStatus === "passed" && (item.conclusion === "improved" || item.conclusion === "unchanged"));
    setSelectedReviewId(review?.reviewId ?? "");
  }

  async function refreshPromotions() {
    if (promotionConfig.mode !== "dev_workflow_rag_promotion_http" || !applicationId) return;
    setPromotions((current) => ({ ...current, status: "empty", summaries: [], failureCode: "", summary: "Loading knowledge promotion candidates." }));
    setPromotions(await listWorkflowRAGPromotionCandidates(promotionConfig, applicationId));
  }

  async function createCandidate() {
    if (!enabled || !selectedDataset || !eligibleReview || !selectedDraft || selectedDraft.validationState !== "valid") return;
    setOperation((current) => ({ ...current, status: "loaded", detail: null, failureCode: "", summary: "Creating a candidate from server-reloaded authority records." }));
    const result = await createWorkflowRAGPromotionCandidate(promotionConfig, applicationId, {
      datasetId: selectedDataset.datasetId, datasetVersion: selectedDataset.latestVersion, datasetDigest: selectedDataset.latestDigest,
    }, selectedReview.reviewId, { draftId: selectedDraft.draftId, draftVersion: selectedDraft.draftVersion });
    setOperation(result);
    if (result.detail) {
      setReason("");
      await refreshPromotions();
    }
  }

  async function openCandidate(candidateId: string) {
    setOperation((current) => ({ ...current, status: "loaded", detail: null, failureCode: "", summary: "Loading exact promotion evidence and current blockers." }));
    const result = await readWorkflowRAGPromotionCandidate(promotionConfig, applicationId, candidateId);
    setOperation(result);
    if (result.detail) setDecision(result.detail.candidate.candidateState === "approved" ? "cancel" : "approve");
  }

  async function submitDecision() {
    if (!enabled || !detail || !canDecide || reasonFailure) return;
    const result = await decideWorkflowRAGPromotionCandidate(promotionConfig, applicationId, detail.candidate.candidateId, detail.candidate.recordVersion, decision, reason);
    setOperation(result.status === "record_version_conflict" ? { ...result, detail } : result);
    if (result.detail) {
      setReason("");
      await refreshPromotions();
    }
  }

  if (promotionConfig.mode === "offline") return <section className="workflow-rag-promotion-panel offline" aria-label="Workflow RAG knowledge promotion"><div className="section-heading compact-heading"><div><p className="eyebrow">Workflow RAG · Promotion</p><h4>知识基线晋级审查未启用</h4></div><span className="status-badge neutral">offline</span></div><p>Offline mode sends zero promotion requests. Dataset、草案、decision 与 binding 均不会在浏览器中伪造。</p></section>;

  return <section className="workflow-rag-promotion-panel" id="workflow-rag-promotion-review" aria-labelledby="workflow-rag-promotion-title">
    <div className="section-heading compact-heading"><div><p className="eyebrow">Workflow RAG · Promotion & binding</p><h4 id="workflow-rag-promotion-title">知识证据晋级、人工决定与配置绑定资格</h4></div><span className={`status-badge ${detail?.eligibility.eligible ? "good" : operation.failureCode ? "bad" : "neutral"}`}>{detail?.candidate.candidateState ?? operation.status}</span></div>
    <div className="workflow-rag-promotion-scope"><article><span>Application</span><strong>{applicationName || "No application selected"}</strong><code>{applicationId || "none"}</code></article><article><span>Boundary</span><strong>three explicit steps</strong><p>Approve → attach → publish review</p></article><article><span>Automation</span><strong>disabled</strong><p>No baseline, release, or publish mutation.</p></article></div>

    {!applicationActive ? <p className="failure-summary">Archived applications keep historical promotion evidence readable but cannot create or decide candidates.</p> : <div className="workflow-rag-promotion-layout">
      <article className="workflow-rag-promotion-create">
        <div className="application-api-card-heading"><div><p className="eyebrow">Exact source bindings</p><h5>Dataset review + source draft</h5></div><button type="button" onClick={() => void loadSources()} disabled={!enabled}>Load sources</button></div>
        <label>Active dataset<select value={selectedDatasetId} onChange={(event) => void loadReviews(event.target.value)} disabled={!enabled || datasets.resources.length === 0}><option value="">No active dataset selected</option>{datasets.resources.map((item) => <option key={item.datasetId} value={item.datasetId}>{item.datasetId} · v{item.latestVersion}</option>)}</select></label>
        <label>Eligible candidate review<select value={selectedReviewId} onChange={(event) => setSelectedReviewId(event.target.value)} disabled={!enabled || reviews.reviews.length === 0}><option value="">No eligible review selected</option>{reviews.reviews.map((item) => <option key={item.reviewId} value={item.reviewId} disabled={item.candidateStatus !== "passed" || !["improved", "unchanged"].includes(item.conclusion)}>{item.reviewId} · {item.conclusion} · {item.candidateStatus}</option>)}</select></label>
        <label>Valid source draft<select value={selectedDraftId} onChange={(event) => setSelectedDraftId(event.target.value)} disabled={!enabled || drafts.summaries.length === 0}><option value="">No saved draft selected</option>{drafts.summaries.map((item) => <option key={item.draftId} value={item.draftId} disabled={item.validationState !== "valid"}>{item.draftId} · v{item.draftVersion} · {item.validationState}</option>)}</select></label>
        {[datasets.failureCode, reviews.failureCode, drafts.failureCode].filter(Boolean).map((code) => <p className="failure-summary" key={code}>{code}</p>)}
        <button type="button" onClick={() => void createCandidate()} disabled={!enabled || !eligibleReview || !selectedDraft || selectedDraft.validationState !== "valid"}>Create promotion candidate</button>
        <p className="boundary-note">Only ids, versions, digests and review refs cross this boundary. The server reloads snapshots, lexical profile and current draft authority.</p>
      </article>

      <article className="workflow-rag-promotion-decision">
        <div className="application-api-card-heading"><div><p className="eyebrow">Append-only human decision</p><h5>{detail?.candidate.candidateId ?? "No candidate selected"}</h5></div><span className="status-badge neutral">record v{detail?.candidate.recordVersion ?? 0}</span></div>
        <label>Decision<select value={decision} onChange={(event) => setDecision(event.target.value as WorkflowRAGPromotionDecision)} disabled={!detail}><option value="approve">Approve binding eligibility</option><option value="reject">Reject candidate</option><option value="defer">Defer decision</option><option value="cancel">Cancel candidate or binding</option></select></label>
        <label>Sanitized reason<textarea rows={4} maxLength={500} value={reason} onChange={(event) => setReason(event.target.value)} placeholder="Explain the evidence decision without query, fragment, prompt, response, or credentials." /></label>
        {reason && reasonFailure ? <p className="failure-summary">{reasonFailure}</p> : null}
        <button type="button" onClick={() => void submitDecision()} disabled={!enabled || !detail || !canDecide || Boolean(reasonFailure)}>Record decision</button>
        {operation.failureCode ? <p className="failure-summary">{operation.failureCode}</p> : null}
        <p className="boundary-note">{operation.summary}</p>
        {operation.status === "record_version_conflict" && detail ? <button type="button" onClick={() => void openCandidate(detail.candidate.candidateId)}>Refresh current record v{operation.currentRecordVersion}</button> : null}
      </article>
    </div>}

    {detail ? <PromotionDetail detail={detail} /> : null}

    <article className="workflow-rag-promotion-saved"><div className="application-api-card-heading"><div><p className="eyebrow">Durable candidates</p><h5>{promotions.summary}</h5></div><button type="button" onClick={() => void refreshPromotions()}>Refresh candidates</button></div>{promotions.failureCode ? <p className="failure-summary">{promotions.failureCode}</p> : null}<div className="workflow-rag-promotion-list">{promotions.summaries.map((item) => <button type="button" key={item.candidateId} onClick={() => void openCandidate(item.candidateId)}><strong>{item.candidateId}</strong><span>{item.candidateState} · record v{item.recordVersion} · {item.eligibilityStatus}</span><small>{item.dataset.datasetId} v{item.dataset.datasetVersion} · {item.blockerCount} blocker(s)</small></button>)}</div></article>
    <p className="boundary-note">Approved means an immutable configuration binding is eligible. It never means attached, released, published, or production-ready.</p>
  </section>;
}

function PromotionDetail({ detail }: { detail: WorkflowRAGPromotionDetail }) {
  const evidence = detail.candidate.evidence;
  return <div className="workflow-rag-promotion-detail">
    <article><div className="application-api-card-heading"><div><p className="eyebrow">Immutable evidence</p><h5>{evidence.dataset.datasetId} · v{evidence.dataset.datasetVersion}</h5></div><span className="status-badge neutral">{evidence.candidateReviewId}</span></div><EvidenceRow label="Dataset digest" value={evidence.dataset.datasetDigest} /><EvidenceRow label="Baseline snapshot" value={`${evidence.baselineSnapshot.snapshotId} v${evidence.baselineSnapshot.snapshotVersion} · ${evidence.baselineSnapshot.ragRef}`} /><EvidenceRow label="Baseline digest" value={evidence.baselineSnapshot.snapshotDigest} /><EvidenceRow label="Candidate snapshot" value={`${evidence.candidateSnapshot.snapshotId} v${evidence.candidateSnapshot.snapshotVersion} · ${evidence.candidateSnapshot.ragRef}`} /><EvidenceRow label="Candidate digest" value={evidence.candidateSnapshot.snapshotDigest} /><EvidenceRow label="Lexical profile" value={`${evidence.profile.profileId} v${evidence.profile.profileVersion} · ${evidence.profile.profileDigest}`} /><EvidenceRow label="Source draft" value={`${evidence.sourceDraft.draftId} v${evidence.sourceDraft.draftVersion} · ${evidence.sourceDraft.draftDigest}`} /></article>
    <article><div className="application-api-card-heading"><div><p className="eyebrow">Dynamic eligibility</p><h5>{detail.eligibility.status}</h5></div><span className={`status-badge ${detail.eligibility.eligible ? "good" : "bad"}`}>{detail.eligibility.blockers.length} blockers</span></div>{detail.eligibility.blockers.length ? <ul>{detail.eligibility.blockers.map((code) => <li key={code}><code>{code}</code></li>)}</ul> : <p>All current authorities still match the approved binding.</p>}{detail.binding ? <div className="workflow-rag-binding-card"><strong>Immutable binding ready for explicit attach</strong><code>{detail.binding.bindingId} · v{detail.binding.bindingVersion}</code><code>{detail.binding.bindingDigest}</code><a href="#application-configuration-draft">Open configuration draft attach step</a></div> : <p className="boundary-note">No binding exists until an explicit approve decision succeeds.</p>}</article>
    <article><div className="application-api-card-heading"><div><p className="eyebrow">Decision history</p><h5>{detail.decisions.length} append-only records</h5></div></div>{detail.decisions.length ? detail.decisions.map((item) => <div className="workflow-rag-decision-record" key={item.decisionId}><strong>{item.decision} · v{item.beforeRecordVersion} → v{item.afterRecordVersion}</strong><span>{item.actorRef} · {item.occurredAt}</span><p>{item.reason}</p></div>) : <p className="boundary-note">No human decision recorded.</p>}</article>
  </div>;
}

function EvidenceRow({ label, value }: { label: string; value: string }) { return <div className="workflow-rag-evidence-row"><strong>{label}</strong><code>{value}</code></div>; }
function emptyDatasets(): WorkflowRAGEvaluationListResult { return { status: "empty", resources: [], nextCursor: "", failureCode: "", summary: "Load active evaluation datasets." }; }
function emptyReviews(): WorkflowRAGCandidateReviewListResult { return { status: "empty", reviews: [], nextCursor: "", failureCode: "", summary: "Select a dataset to load candidate reviews." }; }
function validateReason(reason: string): string { const value = reason.trim(); if (value.length < 4 || value.length > 500) return "workflow_rag_promotion_payload_invalid"; return /authorization:|bearer\s|api[_-]?key\s*[:=]|x-radishmind-dev-|cookie:|password\s*=|secret\s*=|token\s*=|sk-[a-z0-9]|(?:postgres(?:ql)?|mysql|mongodb):\/\//iu.test(value) ? "workflow_rag_promotion_secret_material_forbidden" : ""; }
