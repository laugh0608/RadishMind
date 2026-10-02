import { useTranslation } from "react-i18next";
import "../../i18n/workflowRAGPromotionResources.ts";
import { workflowRAGPromotionStatus, workflowRAGPromotionFailure, workflowRAGPromotionFeedback } from "./workflowRAGPromotionMessages.ts";
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
  onOpenConfigurationAttach: (candidateId: string) => void;
};

export default function WorkflowRAGPromotionPanel({ applicationId, applicationName, applicationActive, onEvidenceChange, onOpenConfigurationAttach }: Props) {
  const { t } = useTranslation("workflow");
  const [sourceLoading, setSourceLoading] = useState(false);
  const [reviewsLoading, setReviewsLoading] = useState(false);
  const [listLoading, setListLoading] = useState(false);
  const [pending, setPending] = useState<"create" | "read" | "decide" | null>(null);
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
    setSourceLoading(false);
    setReviewsLoading(false);
    setListLoading(false);
    setPending(null);
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
    setSourceLoading(true);
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
    setSourceLoading(false);
  }

  async function loadReviews(datasetId: string) {
    setSelectedDatasetId(datasetId);
    setSelectedReviewId("");
    if (!enabled || !datasetId) {
      setReviews(emptyReviews());
      return;
    }
    setReviewsLoading(true);
    setReviews((current) => ({ ...current, status: "empty", reviews: [], failureCode: "", summary: "Loading metadata-only candidate reviews." }));
    const next = await listWorkflowRAGCandidateReviews(evaluationConfig, applicationId, datasetId);
    setReviews(next);
    setReviewsLoading(false);
    const review = next.reviews.find((item) => item.candidateStatus === "passed" && (item.conclusion === "improved" || item.conclusion === "unchanged"));
    setSelectedReviewId(review?.reviewId ?? "");
  }

  async function refreshPromotions() {
    if (promotionConfig.mode !== "dev_workflow_rag_promotion_http" || !applicationId) return;
    setListLoading(true);
    setPromotions((current) => ({ ...current, status: "empty", summaries: [], failureCode: "", summary: "Loading knowledge promotion candidates." }));
    setPromotions(await listWorkflowRAGPromotionCandidates(promotionConfig, applicationId));
    setListLoading(false);
  }

  async function createCandidate() {
    if (!enabled || !selectedDataset || !eligibleReview || !selectedDraft || selectedDraft.validationState !== "valid") return;
    setOperation((current) => ({ ...current, status: "loaded", detail: null, failureCode: "", summary: "Creating a candidate from server-reloaded authority records." }));
    setPending("create");
    const result = await createWorkflowRAGPromotionCandidate(promotionConfig, applicationId, {
      datasetId: selectedDataset.datasetId, datasetVersion: selectedDataset.latestVersion, datasetDigest: selectedDataset.latestDigest,
    }, selectedReview.reviewId, { draftId: selectedDraft.draftId, draftVersion: selectedDraft.draftVersion });
    setPending(null);
    setOperation(result);
    if (result.detail) {
      setReason("");
      await refreshPromotions();
    }
  }

  async function openCandidate(candidateId: string) {
    setOperation((current) => ({ ...current, status: "loaded", detail: null, failureCode: "", summary: "Loading exact promotion evidence and current blockers." }));
    setPending("read");
    const result = await readWorkflowRAGPromotionCandidate(promotionConfig, applicationId, candidateId);
    setPending(null);
    setOperation(result);
    if (result.detail) setDecision(result.detail.candidate.candidateState === "approved" ? "cancel" : "approve");
  }

  async function submitDecision() {
    if (!enabled || !detail || !canDecide || reasonFailure) return;
    setPending("decide");
    const result = await decideWorkflowRAGPromotionCandidate(promotionConfig, applicationId, detail.candidate.candidateId, detail.candidate.recordVersion, decision, reason);
    setPending(null);
    setOperation(result.status === "record_version_conflict" ? { ...result, detail } : result);
    if (result.detail) {
      setReason("");
      await refreshPromotions();
    }
  }

  if (promotionConfig.mode === "offline") return <section className="workflow-rag-promotion-panel offline" id="workflow-rag-promotion-review" aria-label={t($ => $.ragPromotion.label)}><div className="section-heading compact-heading"><div><p className="eyebrow">{t($ => $.ragPromotion.eyebrow)}</p><h4>{t($ => $.ragPromotion.offlineTitle)}</h4></div><span className="status-badge neutral">{workflowRAGPromotionStatus(t, "offline")}</span></div><p>{t($ => $.ragPromotion.offlineNote)}</p></section>;

  return <section className="workflow-rag-promotion-panel" id="workflow-rag-promotion-review" aria-labelledby="workflow-rag-promotion-title">
    <div className="section-heading compact-heading"><div><p className="eyebrow">{t($ => $.ragPromotion.eyebrow)}</p><h4 id="workflow-rag-promotion-title">{t($ => $.ragPromotion.heading)}</h4></div><span className={`status-badge ${detail?.eligibility.eligible ? "good" : operation.failureCode ? "bad" : "neutral"}`}>{workflowRAGPromotionStatus(t, detail?.candidate.candidateState ?? operation.status)}</span></div>
    <div className="workflow-rag-promotion-scope"><article><span>{t($ => $.ragPromotion.application)}</span><strong>{applicationName || t($ => $.ragPromotion.noApplication)}</strong><code>{applicationId || t($ => $.ragPromotion.none)}</code></article><article><span>{t($ => $.ragPromotion.boundary)}</span><strong>{t($ => $.ragPromotion.threeSteps)}</strong><p>{t($ => $.ragPromotion.steps)}</p></article><article><span>{t($ => $.ragPromotion.automation)}</span><strong>{t($ => $.ragPromotion.disabled)}</strong><p>{t($ => $.ragPromotion.noMutation)}</p></article></div>

    {!applicationActive ? <p className="failure-summary">{t($ => $.ragPromotion.archived)}</p> : <div className="workflow-rag-promotion-layout">
      <article className="workflow-rag-promotion-create">
        <div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.ragPromotion.sourceEyebrow)}</p><h5>{t($ => $.ragPromotion.sourceTitle)}</h5></div><button type="button" onClick={() => void loadSources()} disabled={!enabled || sourceLoading}>{t($ => $.ragPromotion.loadSources)}</button></div>
        <label>{t($ => $.ragPromotion.dataset)}<select value={selectedDatasetId} onChange={(event) => void loadReviews(event.target.value)} disabled={!enabled || datasets.resources.length === 0}><option value="">{t($ => $.ragPromotion.selectDataset)}</option>{datasets.resources.map((item) => <option key={item.datasetId} value={item.datasetId}>{item.datasetId} · v{item.latestVersion}</option>)}</select></label>
        <label>{t($ => $.ragPromotion.review)}<select value={selectedReviewId} onChange={(event) => setSelectedReviewId(event.target.value)} disabled={!enabled || reviews.reviews.length === 0}><option value="">{t($ => $.ragPromotion.selectReview)}</option>{reviews.reviews.map((item) => <option key={item.reviewId} value={item.reviewId} disabled={item.candidateStatus !== "passed" || !["improved", "unchanged"].includes(item.conclusion)}>{item.reviewId} · {workflowRAGPromotionStatus(t, item.conclusion)} · {workflowRAGPromotionStatus(t, item.candidateStatus)}</option>)}</select></label>
        <label>{t($ => $.ragPromotion.draft)}<select value={selectedDraftId} onChange={(event) => setSelectedDraftId(event.target.value)} disabled={!enabled || drafts.summaries.length === 0}><option value="">{t($ => $.ragPromotion.selectDraft)}</option>{drafts.summaries.map((item) => <option key={item.draftId} value={item.draftId} disabled={item.validationState !== "valid"}>{item.draftId} · v{item.draftVersion} · {workflowRAGPromotionStatus(t, item.validationState)}</option>)}</select></label>
        {sourceLoading || reviewsLoading ? <p role="status">{sourceLoading ? t($ => $.ragPromotion.activity.sources) : t($ => $.ragPromotion.activity.reviews)}</p> : null}
        {[datasets.failureCode, reviews.failureCode, drafts.failureCode].filter(Boolean).map((code) => <p className="failure-summary" key={code}>{workflowRAGPromotionFailure(t, code)} <code>{code}</code></p>)}
        <button type="button" onClick={() => void createCandidate()} disabled={Boolean(pending) || sourceLoading || reviewsLoading || !enabled || !eligibleReview || !selectedDraft || selectedDraft.validationState !== "valid"}>{t($ => $.ragPromotion.create)}</button>
        <p className="boundary-note">{t($ => $.ragPromotion.sourceNote)}</p>
      </article>

      <article className="workflow-rag-promotion-decision">
        <div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.ragPromotion.decisionEyebrow)}</p><h5>{detail?.candidate.candidateId ?? t($ => $.ragPromotion.noCandidate)}</h5></div><span className="status-badge neutral">{t($ => $.ragPromotion.recordVersion, { version: detail?.candidate.recordVersion ?? 0 })}</span></div>
        <label>{t($ => $.ragPromotion.decision)}<select value={decision} onChange={(event) => setDecision(event.target.value as WorkflowRAGPromotionDecision)} disabled={!detail || Boolean(pending)}><option value="approve">{t($ => $.ragPromotion.approveOption)}</option><option value="reject">{t($ => $.ragPromotion.rejectOption)}</option><option value="defer">{t($ => $.ragPromotion.deferOption)}</option><option value="cancel">{t($ => $.ragPromotion.cancelOption)}</option></select></label>
        <label>{t($ => $.ragPromotion.reason)}<textarea rows={4} maxLength={500} value={reason} onChange={(event) => setReason(event.target.value)} placeholder={t($ => $.ragPromotion.reasonPlaceholder)} /></label>
        {reason && reasonFailure ? <p className="failure-summary">{reasonFailure === "workflow_rag_promotion_payload_invalid" ? t($ => $.ragPromotion.reasonInvalid) : workflowRAGPromotionFailure(t, reasonFailure)} <code>{reasonFailure}</code></p> : null}
        <button type="button" onClick={() => void submitDecision()} disabled={Boolean(pending) || !enabled || !detail || !canDecide || Boolean(reasonFailure)}>{t($ => $.ragPromotion.submit)}</button>
        {operation.failureCode ? <p className="failure-summary">{workflowRAGPromotionFailure(t, operation.failureCode)} <code>{operation.failureCode}</code></p> : null}
        <p className="boundary-note">{pending === "create" ? t($ => $.ragPromotion.activity.create) : pending === "read" ? t($ => $.ragPromotion.activity.read) : pending === "decide" ? t($ => $.ragPromotion.activity.decide) : workflowRAGPromotionFeedback(t, operation)}</p>
        {operation.status === "record_version_conflict" && detail ? <button type="button" onClick={() => void openCandidate(detail.candidate.candidateId)}>{t($ => $.ragPromotion.refreshRecord, { version: operation.currentRecordVersion })}</button> : null}
      </article>
    </div>}

    {detail ? <PromotionDetail detail={detail} applicationActive={applicationActive} onOpenConfigurationAttach={onOpenConfigurationAttach} /> : null}

    <article className="workflow-rag-promotion-saved"><div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.ragPromotion.savedEyebrow)}</p><h5>{listLoading ? t($ => $.ragPromotion.activity.list) : promotions.failureCode ? t($ => $.ragPromotion.listFailed) : promotions.summaries.length ? t($ => $.ragPromotion.candidates, { count: promotions.summaries.length }) : t($ => $.ragPromotion.emptyList)}</h5></div><button type="button" onClick={() => void refreshPromotions()} disabled={listLoading}>{t($ => $.ragPromotion.refresh)}</button></div>{promotions.failureCode ? <p className="failure-summary">{workflowRAGPromotionFailure(t, promotions.failureCode)} <code>{promotions.failureCode}</code></p> : null}<div className="workflow-rag-promotion-list">{promotions.summaries.map((item) => <button type="button" key={item.candidateId} onClick={() => void openCandidate(item.candidateId)} disabled={Boolean(pending)}><strong>{item.candidateId}</strong><span>{workflowRAGPromotionStatus(t, item.candidateState)} · {t($ => $.ragPromotion.recordVersion, { version: item.recordVersion })} · {workflowRAGPromotionStatus(t, item.eligibilityStatus)}</span><small>{item.dataset.datasetId} v{item.dataset.datasetVersion} · {t($ => $.ragPromotion.blockers, { count: item.blockerCount })}</small></button>)}</div></article>
    <p className="boundary-note">{t($ => $.ragPromotion.footer)}</p>
  </section>;
}

function PromotionDetail({
  detail,
  applicationActive,
  onOpenConfigurationAttach,
}: {
  detail: WorkflowRAGPromotionDetail;
  applicationActive: boolean;
  onOpenConfigurationAttach: (candidateId: string) => void;
}) {
  const { t } = useTranslation("workflow");
  const evidence = detail.candidate.evidence;
  return <div className="workflow-rag-promotion-detail">
    <article><div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.ragPromotion.evidence)}</p><h5>{evidence.dataset.datasetId} · v{evidence.dataset.datasetVersion}</h5></div><span className="status-badge neutral">{evidence.candidateReviewId}</span></div><EvidenceRow label={t($ => $.ragPromotion.datasetDigest)} value={evidence.dataset.datasetDigest} /><EvidenceRow label={t($ => $.ragPromotion.baselineSnapshot)} value={`${evidence.baselineSnapshot.snapshotId} v${evidence.baselineSnapshot.snapshotVersion} · ${evidence.baselineSnapshot.ragRef}`} /><EvidenceRow label={t($ => $.ragPromotion.baselineDigest)} value={evidence.baselineSnapshot.snapshotDigest} /><EvidenceRow label={t($ => $.ragPromotion.candidateSnapshot)} value={`${evidence.candidateSnapshot.snapshotId} v${evidence.candidateSnapshot.snapshotVersion} · ${evidence.candidateSnapshot.ragRef}`} /><EvidenceRow label={t($ => $.ragPromotion.candidateDigest)} value={evidence.candidateSnapshot.snapshotDigest} /><EvidenceRow label={t($ => $.ragPromotion.profile)} value={`${evidence.profile.profileId} v${evidence.profile.profileVersion} · ${evidence.profile.profileDigest}`} /><EvidenceRow label={t($ => $.ragPromotion.sourceDraft)} value={`${evidence.sourceDraft.draftId} v${evidence.sourceDraft.draftVersion} · ${evidence.sourceDraft.draftDigest}`} /></article>
    <article><div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.ragPromotion.eligibility)}</p><h5>{workflowRAGPromotionStatus(t, detail.eligibility.status)}</h5></div><span className={`status-badge ${detail.eligibility.eligible ? "good" : "bad"}`}>{t($ => $.ragPromotion.blockers, { count: detail.eligibility.blockers.length })}</span></div>{detail.eligibility.blockers.length ? <ul>{detail.eligibility.blockers.map((code) => <li key={code}>{workflowRAGPromotionFailure(t, code)} <code>{code}</code></li>)}</ul> : <p>{t($ => $.ragPromotion.authoritiesMatch)}</p>}{detail.binding ? <div className="workflow-rag-binding-card"><strong>{t($ => $.ragPromotion.bindingReady)}</strong><code>{detail.binding.bindingId} · v{detail.binding.bindingVersion}</code><code>{detail.binding.bindingDigest}</code><button type="button" disabled={!applicationActive || !detail.eligibility.eligible} onClick={() => onOpenConfigurationAttach(detail.candidate.candidateId)}>{t($ => $.ragPromotion.attach)}</button></div> : <p className="boundary-note">{t($ => $.ragPromotion.noBinding)}</p>}</article>
    <article><div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.ragPromotion.history)}</p><h5>{t($ => $.ragPromotion.records, { count: detail.decisions.length })}</h5></div></div>{detail.decisions.length ? detail.decisions.map((item) => <div className="workflow-rag-decision-record" key={item.decisionId}><strong>{workflowRAGPromotionStatus(t, item.decision)} · v{item.beforeRecordVersion} → v{item.afterRecordVersion}</strong><span>{item.actorRef} · {item.occurredAt}</span><p>{item.reason}</p></div>) : <p className="boundary-note">{t($ => $.ragPromotion.noDecisions)}</p>}</article>
  </div>;
}

function EvidenceRow({ label, value }: { label: string; value: string }) { return <div className="workflow-rag-evidence-row"><strong>{label}</strong><code>{value}</code></div>; }
function emptyDatasets(): WorkflowRAGEvaluationListResult { return { status: "empty", resources: [], nextCursor: "", failureCode: "", summary: "Load active evaluation datasets." }; }
function emptyReviews(): WorkflowRAGCandidateReviewListResult { return { status: "empty", reviews: [], nextCursor: "", failureCode: "", summary: "Select a dataset to load candidate reviews." }; }
function validateReason(reason: string): string { const value = reason.trim(); if (value.length < 4 || value.length > 500) return "workflow_rag_promotion_payload_invalid"; return /authorization:|bearer\s|api[_-]?key\s*[:=]|x-radishmind-dev-|cookie:|password\s*=|secret\s*=|token\s*=|sk-[a-z0-9]|(?:postgres(?:ql)?|mysql|mongodb):\/\//iu.test(value) ? "workflow_rag_promotion_secret_material_forbidden" : ""; }
