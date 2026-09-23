import "../../i18n/publishResources.ts";
import { useTranslation } from "react-i18next";
import { useEffect, useMemo, useRef, useState } from "react";

import {
  findExactValidApplicationConfigurationDraft,
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
  type ApplicationPublishAgentCopilotProfileRef,
  type ApplicationPublishOperationState,
  type ApplicationPublishPromptTemplateRef,
} from "./applicationPublishCandidateConsumer.ts";
import { requestGatewayRequestHistoryReview, requestModelGatewayPlaygroundHandoff } from "./modelGatewayPlaygroundEvents.ts";
import PromptApplicationRuntimePanel from "./promptApplicationRuntimePanel.tsx";
import {
  readAgentCopilotProfileConfig,
  readAgentCopilotProfileVersion,
  type AgentCopilotProfileVersion,
} from "./agentCopilotProfileConsumer.ts";
import {
  readPromptTemplateConfig,
  readPromptTemplateVersion,
  type PromptTemplateVersion,
} from "./promptApplicationTemplateConsumer.ts";
import { WorkflowRAGRuntimeAssignmentPanel } from "./workflowRAGApplicationRuntimePanel.tsx";
import type { ApplicationDevelopmentOwnerEvidence } from "./applicationDevelopmentReadiness.ts";

const publishConfig = readApplicationPublishCandidateConfig();
const draftConfig = readApplicationConfigurationDraftConfig();
const promptTemplateConfig = readPromptTemplateConfig();
const agentProfileConfig = readAgentCopilotProfileConfig();

export default function ApplicationPublishCandidatePanel({
  baseline,
  readOnly = false,
  onEvidenceChange,
  handoffDraftId = "",
  handoffId = "",
  onHandoffConsumed,
  includeRuntimeAssignment = true,
  onSelectedCandidateChange,
}: {
  baseline: ApplicationConfigurationBaseline;
  readOnly?: boolean;
  onEvidenceChange?: (evidence: ApplicationDevelopmentOwnerEvidence) => void;
  handoffDraftId?: string;
  handoffId?: string;
  onHandoffConsumed?: (handoffId: string) => void;
  includeRuntimeAssignment?: boolean;
  onSelectedCandidateChange?: (candidate: ApplicationPublishCandidate | null) => void;
}) {
  const { t } = useTranslation("applications");
  const [draftList, setDraftList] = useState<ApplicationConfigurationDraftListState>(() => initialApplicationConfigurationDraftListState(draftConfig));
  const [candidateList, setCandidateList] = useState<ApplicationPublishCandidateListState>(() => initialApplicationPublishListState(publishConfig));
  const [candidate, setCandidate] = useState<ApplicationPublishCandidate | null>(null);
  const [operation, setOperation] = useState<ApplicationPublishOperationState>(() => initialApplicationPublishOperationState(publishConfig));
  const [selectedDraftId, setSelectedDraftId] = useState("");
  const [candidateId, setCandidateId] = useState(() => newCandidateId(baseline.applicationId));
  const [evidenceText, setEvidenceText] = useState("");
  const [decision, setDecision] = useState<ApplicationPublishDecision>("approve");
  const [reviewReason, setReviewReason] = useState("");
  const [handoffState, setHandoffState] = useState<{ kind: "offline" | "loading" | "reloaded" | "unavailable" | "failed"; draftId: string } | null>(null);
  const handledHandoffIdRef = useRef("");

  const candidateStateLabel = (state: string) => {
    switch (state) {
      case "pending_review": return t($ => $.publish.pendingReviewStatus);
      case "approved": return t($ => $.publish.approvedStatus);
      case "rejected": return t($ => $.publish.rejectedStatus);
      case "changes_requested": return t($ => $.publish.changesRequestedStatus);
      case "withdrawn": return t($ => $.publish.withdrawnStatus);
      default: return t($ => $.publish.unknownStateNotice);
    }
  };
  const operationStatusLabel = () => {
    switch (operation.status) {
      case "offline": return t($ => $.publish.offlineStatus);
      case "idle": return t($ => $.publish.idleStatus);
      case "loading": return t($ => $.publish.loadingStatus);
      case "creating": return t($ => $.publish.creatingStatus);
      case "created": return t($ => $.publish.createdStatus);
      case "loaded": return t($ => $.publish.loadedStatus);
      case "reviewing": return t($ => $.publish.reviewingStatus);
      case "reviewed": return t($ => $.publish.reviewedStatus);
      case "review_version_conflict": return t($ => $.publish.reviewVersionConflictStatus);
      case "immutable_conflict": return t($ => $.publish.immutableConflictStatus);
      case "scope_denied": return t($ => $.publish.scopeDeniedStatus);
      case "failed": return t($ => $.publish.failedStatus);
    }
  };
  const operationMessage = () => {
    if (operation.failureCode) return operation.status === "review_version_conflict"
      ? t($ => $.publish.reviewVersionConflictNotice) : operation.status === "scope_denied"
        ? t($ => $.publish.scopeDeniedNotice) : t($ => $.publish.publishOperationFailed);
    switch (operation.status) {
      case "offline": return t($ => $.publish.offlineReviewNotice);
      case "idle": return t($ => $.publish.selectSavedValidDraftNotice);
      case "loading": return t($ => $.publish.loadingCandidateAndEligibility);
      case "creating": return t($ => $.publish.creatingCandidateFromSavedDraft);
      case "created": return t($ => $.publish.candidateCreatedNotice);
      case "loaded": return t($ => $.publish.candidateLoadedNotice);
      case "reviewing": return t($ => $.publish.recordingAppendOnlyReview);
      case "reviewed": return t($ => $.publish.reviewRecordedNotice);
      default: return t($ => $.publish.publishOperationFailed);
    }
  };
  const candidateListMessage = candidateList.status === "offline" ? t($ => $.publish.offlineCandidateListNotice)
    : candidateList.status === "idle" ? t($ => $.publish.loadCandidatesNotice)
      : candidateList.status === "loading" ? t($ => $.publish.loadingPublishCandidates)
        : candidateList.status === "failed" ? t($ => $.publish.candidateListFailedNotice)
          : candidateList.status === "empty" ? t($ => $.publish.noPublishCandidates)
            : t($ => $.publish.loadedCandidateCount, { count: candidateList.summaries.length });

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
    setHandoffState(null);
    handledHandoffIdRef.current = "";
  }, [baseline.applicationId]);

  useEffect(() => {
    onSelectedCandidateChange?.(candidate);
  }, [candidate, onSelectedCandidateChange]);

  const enabled = publishConfig.mode === "dev_application_publish_http" && draftConfig.mode === "dev_application_draft_http";
  const mutationEnabled = enabled && !readOnly;
  const selectedDraft = draftList.summaries.find((summary) => summary.draftId === selectedDraftId) ?? null;
  const evidence = useMemo(() => parseApplicationPublishEvidence(evidenceText), [evidenceText]);
  const reviewFailure = useMemo(() => validateApplicationPublishReview(decision, reviewReason), [decision, reviewReason]);
  const canReview = candidate?.candidateState === "pending_review" || candidate?.candidateState === "approved" && decision === "withdraw";

  useEffect(() => {
    if (!onEvidenceChange) return;
    const ownerFailed = operation.status === "failed" || operation.status === "scope_denied" || operation.status.endsWith("conflict");
    const candidateReviewed = candidate?.candidateState === "approved";
    const candidateBlocked = Boolean(candidate && candidate.promotionEligibility.blockers.length > 0);
    const candidateRef = candidate
      ? [{
        kind: "candidate" as const,
        id: candidate.candidateId,
        ...(candidate.reviewVersion > 0 ? { version: candidate.reviewVersion } : {}),
      }]
      : [];
    onEvidenceChange({
      contributionId: "publish_candidate",
      status: ownerFailed || candidateBlocked ? "blocked" : candidateReviewed ? "available" : "incomplete",
      coverage: candidate ? "complete" : ownerFailed ? "complete" : "none",
      evidenceRefs: candidateRef,
      missingEvidence: candidateReviewed ? [] : ["Approve an immutable Application publish candidate through its owner."],
      blockers: ownerFailed
        ? [{ code: operation.failureCode || "application_publish_candidate_blocked", summary: operation.summary }]
        : candidate?.promotionEligibility.blockers ?? [],
      failureCodes: ownerFailed && operation.failureCode ? [operation.failureCode] : [],
    });
  }, [candidate, onEvidenceChange, operation]);

  useEffect(() => {
    if (readOnly && enabled) void refreshCandidates();
  }, [baseline.applicationId, readOnly]);

  async function loadDrafts(preferredDraftId = ""): Promise<string> {
    if (!enabled) return "";
    setDraftList((current) => ({ ...current, status: "loading", summaries: [], failureCode: "", summary: "Loading saved valid application drafts." }));
    const next = await listApplicationConfigurationDrafts(draftConfig, baseline.applicationId);
    setDraftList(next);
    const firstValid = preferredDraftId
      ? findExactValidApplicationConfigurationDraft(next.summaries, baseline.applicationId, preferredDraftId)
      : next.summaries.find((summary) => summary.validationState === "valid");
    setSelectedDraftId(firstValid?.draftId ?? "");
    return firstValid?.draftId ?? "";
  }

  useEffect(() => {
    if (!handoffId || !handoffDraftId || handledHandoffIdRef.current === handoffId) return;
    handledHandoffIdRef.current = handoffId;
    if (!enabled) {
      setHandoffState({ kind: "offline", draftId: handoffDraftId });
      onHandoffConsumed?.(handoffId);
      return;
    }
    setHandoffState({ kind: "loading", draftId: handoffDraftId });
    void loadDrafts(handoffDraftId)
      .then((selectedId) => setHandoffState(selectedId === handoffDraftId
        ? { kind: "reloaded", draftId: handoffDraftId }
        : { kind: "unavailable", draftId: handoffDraftId }))
      .catch(() => setHandoffState({ kind: "failed", draftId: handoffDraftId }))
      .finally(() => onHandoffConsumed?.(handoffId));
  }, [baseline.applicationId, enabled, handoffDraftId, handoffId, onHandoffConsumed]);

  async function refreshCandidates() {
    if (!enabled) return;
    setCandidateList((current) => ({ ...current, status: "loading", summaries: [], failureCode: "", summary: "Loading publish candidates." }));
    setCandidateList(await listApplicationPublishCandidates(publishConfig, baseline.applicationId));
  }

  async function createCandidate() {
    if (!mutationEnabled || !selectedDraft || selectedDraft.validationState !== "valid" || evidence.failureCode) return;
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
    if (readOnly || !candidate || reviewFailure || !canReview) return;
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
        <div><p className="eyebrow">{t($ => $.publish.publishGovernanceTitle)}</p><h4 id="application-publish-title">{t($ => $.publish.publishGovernanceSubtitle)}</h4></div>
        <span className={`status-badge ${readOnly || candidate?.candidateState === "approved" ? "good" : operation.status.includes("conflict") || operation.status === "failed" ? "bad" : "neutral"}`}>{readOnly ? t($ => $.publish.archivedReadOnly) : candidate ? candidateStateLabel(candidate.candidateState) : operationStatusLabel()}</span>
      </div>

      <div className="application-publish-scope">
        <article><span>{t($ => $.publish.application)}</span><strong>{baseline.displayName}</strong><code>{baseline.applicationId}</code></article>
        <article><span>{t($ => $.publish.baseline)}</span><strong>{baseline.updatedAt}</strong><p>{t($ => $.publish.controlPlaneTruthImmutable)}</p></article>
        <article><span>{t($ => $.publish.promotion)}</span><strong>{t($ => $.publish.disabled)}</strong><p>{t($ => $.publish.candidateApprovalDoesNotMutateApplication)}</p></article>
      </div>
      {handoffState ? <p className="boundary-note" role="status">{handoffState.kind === "offline" ? t($ => $.publish.draftHandoffOffline, { draftId: handoffState.draftId }) : handoffState.kind === "loading" ? t($ => $.publish.loadingExactDraft, { draftId: handoffState.draftId }) : handoffState.kind === "reloaded" ? t($ => $.publish.exactDraftReloaded, { draftId: handoffState.draftId }) : handoffState.kind === "unavailable" ? t($ => $.publish.draftUnavailableNoFallback, { draftId: handoffState.draftId }) : t($ => $.publish.draftReloadFailedNoFallback, { draftId: handoffState.draftId })}</p> : null}

      {readOnly ? <p className="boundary-note">{t($ => $.publish.archivedCandidateReadOnlyNotice)}</p> : <div className="application-publish-layout">
        <article className="application-publish-create">
          <div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.publish.candidateSource)}</p><h5>{t($ => $.publish.bindExactSavedDraftVersion)}</h5></div><button type="button" onClick={() => void loadDrafts()} disabled={!enabled || draftList.status === "loading"}>{t($ => $.publish.loadSavedDrafts)}</button></div>
          <label>{t($ => $.publish.savedValidDraft)}<select value={selectedDraftId} onChange={(event) => setSelectedDraftId(event.target.value)} disabled={!enabled || draftList.summaries.length === 0}><option value="">{t($ => $.publish.noSavedValidDraftSelected)}</option>{draftList.summaries.map((summary) => <option key={summary.draftId} value={summary.draftId} disabled={summary.validationState !== "valid"}>{summary.draftId} · v{summary.draftVersion} · {summary.validationState === "valid" ? t($ => $.publish.validStatus) : t($ => $.publish.invalidStatus)}{summary.workflowRAGBindingRef ? t($ => $.publish.ragBound) : ""}{summary.promptTemplateRef ? t($ => $.publish.templateVersionReference, { version: summary.promptTemplateRef.templateVersion }) : ""}{summary.agentCopilotProfileRef ? t($ => $.publish.profileVersionReference, { version: summary.agentCopilotProfileRef.profileVersion }) : ""}</option>)}</select></label>
          {selectedDraft?.workflowRAGBindingRef ? <div className="application-publish-binding"><strong>{t($ => $.publish.exactDraftBinding)}</strong><code>{selectedDraft.workflowRAGBindingRef.bindingId} · v{selectedDraft.workflowRAGBindingRef.bindingVersion}</code><code>{selectedDraft.workflowRAGBindingRef.bindingDigest}</code></div> : null}
          {selectedDraft?.promptTemplateRef ? <div className="application-publish-binding"><strong>{t($ => $.publish.exactPromptTemplateReference)}</strong><code>{selectedDraft.promptTemplateRef.templateId} · v{selectedDraft.promptTemplateRef.templateVersion}</code><code>{selectedDraft.promptTemplateRef.templateDigest}</code></div> : null}
          {selectedDraft?.agentCopilotProfileRef ? <div className="application-publish-binding"><strong>{t($ => $.publish.exactAgentProfileReference)}</strong><code>{selectedDraft.agentCopilotProfileRef.profileId} · v{selectedDraft.agentCopilotProfileRef.profileVersion}</code><code>{selectedDraft.agentCopilotProfileRef.profileDigest}</code><code>{selectedDraft.agentCopilotProfileRef.policyDigest}</code></div> : null}
          {!selectedDraft?.workflowRAGBindingRef && !selectedDraft?.promptTemplateRef && !selectedDraft?.agentCopilotProfileRef ? <p className="boundary-note">{t($ => $.publish.draftNoRelatedReferences)}</p> : null}
          <label>{t($ => $.publish.candidateIdLabel)}<input value={candidateId} onChange={(event) => setCandidateId(event.target.value)} maxLength={160} /></label>
          <label>{t($ => $.publish.sanitizedRequestHistoryReferences)}<textarea value={evidenceText} onChange={(event) => setEvidenceText(event.target.value)} rows={4} placeholder={t($ => $.publish.requestIdsPlaceholder)} /></label>
          {evidence.failureCode ? <p className="failure-summary">{t($ => $.publish.requestEvidenceInvalid)} <code>{evidence.failureCode}</code></p> : <p className="boundary-note">{t($ => $.publish.normalizedHistoryReferences, { count: evidence.requestIds.length })}</p>}
          <button type="button" onClick={() => void createCandidate()} disabled={!mutationEnabled || !selectedDraft || selectedDraft.validationState !== "valid" || Boolean(evidence.failureCode) || operation.status === "creating"}>{t($ => $.publish.createImmutableCandidate)}</button>
          <p className="boundary-note">{t($ => $.publish.serverReloadsCandidateSources)}</p>
        </article>

        <article className="application-publish-review">
          <div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.publish.reviewDecision)}</p><h5>{candidate?.candidateId ?? t($ => $.publish.noCandidateSelected)}</h5></div><span className="status-badge neutral">{t($ => $.publish.reviewVersionLabel, { version: candidate?.reviewVersion ?? 0 })}</span></div>
          <label>{t($ => $.publish.decision)}<select value={decision} onChange={(event) => setDecision(event.target.value as ApplicationPublishDecision)} disabled={!candidate}><option value="approve">{t($ => $.publish.approveCandidate)}</option><option value="reject">{t($ => $.publish.rejectCandidate)}</option><option value="request_changes">{t($ => $.publish.requestChanges)}</option><option value="withdraw">{t($ => $.publish.withdrawCandidate)}</option></select></label>
          <label>{t($ => $.publish.reviewReason)}<textarea value={reviewReason} onChange={(event) => setReviewReason(event.target.value)} rows={4} maxLength={500} placeholder={t($ => $.publish.reviewReasonPlaceholder)} /></label>
          {reviewFailure && reviewReason ? <p className="failure-summary">{t($ => $.publish.reviewReasonInvalid)} <code>{reviewFailure}</code></p> : null}
          <button type="button" onClick={() => void submitReview()} disabled={!enabled || !candidate || !canReview || Boolean(reviewFailure) || operation.status === "reviewing"}>{t($ => $.publish.recordReviewDecision)}</button>
          {operation.failureCode ? <p className="failure-summary">{operation.failureCode}</p> : null}
          <p className="boundary-note">{operationMessage()}</p>
          {operation.status === "review_version_conflict" && candidate ? <button type="button" onClick={() => void openCandidate(candidate.candidateId)}>{t($ => $.publish.restoreReviewVersionWithNumber, { version: operation.currentReviewVersion })}</button> : null}
        </article>
      </div>}

      {candidate ? <>
        <CandidateDetail candidate={candidate} baseline={baseline} readOnly={readOnly} onIntegration={openIntegration} onPlayground={openPlayground} onHistory={(requestId) => { requestGatewayRequestHistoryReview(requestId, candidate.applicationId); window.location.hash = "model-gateway-request-history"; }} />
        {includeRuntimeAssignment && candidate.schemaVersion === "application_publish_candidate.v3" ? (
          <PromptApplicationRuntimePanel
            applicationId={candidate.applicationId}
            publishCandidateId={candidate.candidateId}
            candidateApproved={candidate.candidateState === "approved"}
            readOnly={readOnly}
            onEvidenceChange={onEvidenceChange}
          />
        ) : includeRuntimeAssignment && candidate.schemaVersion !== "application_publish_candidate.v4" ? (
          <WorkflowRAGRuntimeAssignmentPanel
            applicationId={candidate.applicationId}
            publishCandidateId={candidate.candidateId}
            candidateApproved={candidate.candidateState === "approved"}
            readOnly={readOnly}
            onEvidenceChange={onEvidenceChange}
          />
        ) : null}
      </> : <p className="boundary-note">{readOnly ? t($ => $.publish.openExistingCandidateInstruction) : t($ => $.publish.createOrOpenCandidateInstruction)}</p>}

      <article className="application-publish-saved">
        <div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.publish.savedDevelopmentCandidates)}</p><h5>{candidateListMessage}</h5></div><button type="button" onClick={() => void refreshCandidates()} disabled={!enabled || candidateList.status === "loading"}>{t($ => $.publish.refreshCandidates)}</button></div>
        {candidateList.failureCode ? <p className="failure-summary">{candidateList.failureCode}</p> : null}
        <div className="application-publish-candidate-list">{candidateList.summaries.map((summary) => <button type="button" key={summary.candidateId} onClick={() => void openCandidate(summary.candidateId)}><strong>{summary.candidateId}</strong><span>{candidateStateLabel(summary.candidateState)} · {t($ => $.publish.reviewVersionLabel, { version: summary.reviewVersion })}{summary.workflowRAGBindingRef ? t($ => $.publish.ragBound) : ""}{summary.promptTemplateRef ? t($ => $.publish.templateVersionReference, { version: summary.promptTemplateRef.templateVersion }) : ""}{summary.agentCopilotProfileRef ? t($ => $.publish.profileVersionReference, { version: summary.agentCopilotProfileRef.profileVersion }) : ""}</span><small>{t($ => $.publish.draftVersionAndBlockers, { version: summary.draftVersion, count: summary.promotionBlockers })}</small></button>)}</div>
      </article>

      <p className="boundary-note">{t($ => $.publish.offlineModeBoundaryNotice)}</p>
    </section>
  );
}

function CandidateDetail({ candidate, baseline, readOnly, onIntegration, onPlayground, onHistory }: { candidate: ApplicationPublishCandidate; baseline: ApplicationConfigurationBaseline; readOnly: boolean; onIntegration: () => void; onPlayground: () => void; onHistory: (requestId: string) => void }) {
  const { t } = useTranslation("applications");
  const decisionLabel = (decision: string) => decision === "approve" ? t($ => $.publish.approveCandidate)
    : decision === "reject" ? t($ => $.publish.rejectCandidate)
      : decision === "request_changes" ? t($ => $.publish.requestChanges)
        : decision === "withdraw" ? t($ => $.publish.withdrawCandidate)
          : t($ => $.publish.unknownStateNotice);
  const blockerMessage = (code: string) => {
    switch (code) {
      case "publish_review_required": return t($ => $.publish.reviewRequiredBlocker);
      case "publish_review_rejected": return t($ => $.publish.reviewRejectedBlocker);
      case "publish_changes_requested": return t($ => $.publish.changesRequestedBlocker);
      case "publish_candidate_withdrawn": return t($ => $.publish.candidateWithdrawnBlocker);
      case "promotion_disabled": return t($ => $.publish.promotionDisabledBlocker);
      case "publish_candidate_superseded": return t($ => $.publish.candidateSupersededBlocker);
      case "publish_candidate_draft_changed": return t($ => $.publish.draftChangedBlocker);
      case "application_base_revision_changed": return t($ => $.publish.baselineChangedBlocker);
      default: return t($ => $.publish.unknownPromotionBlocker);
    }
  };
  const comparison = [
    { field: "display_name", before: baseline.displayName, after: candidate.configuration.displayName },
    { field: "application_kind", before: baseline.applicationKind, after: candidate.configuration.applicationKind },
    { field: "default_protocol", before: t($ => $.publish.notConfiguredInReadModel), after: candidate.configuration.defaultProtocol },
    { field: "default_model", before: t($ => $.publish.notConfiguredInReadModel), after: candidate.configuration.defaultModel },
  ];
  return <div className="application-publish-detail">
    <article className="application-publish-snapshot"><div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.publish.immutableSnapshot)}</p><h5>{candidate.draftId} · v{candidate.draftVersion}</h5></div><span className="status-badge neutral">{candidate.schemaVersion}</span></div><code className="application-publish-digest">{candidate.draftDigest}</code>{candidate.configuration.workflowRAGBindingRef ? <div className="application-publish-binding"><strong>{t($ => $.publish.exactImmutableRagBinding)}</strong><code>{candidate.configuration.workflowRAGBindingRef.bindingId} · v{candidate.configuration.workflowRAGBindingRef.bindingVersion}</code><code>{candidate.configuration.workflowRAGBindingRef.bindingDigest}</code></div> : null}{candidate.configuration.promptTemplateRef ? <PromptTemplateSourceReview applicationId={candidate.applicationId} templateRef={candidate.configuration.promptTemplateRef} /> : null}{candidate.configuration.agentCopilotProfileRef ? <AgentCopilotProfileSourceReview applicationId={candidate.applicationId} profileRef={candidate.configuration.agentCopilotProfileRef} /> : null}{!candidate.configuration.workflowRAGBindingRef && !candidate.configuration.promptTemplateRef && !candidate.configuration.agentCopilotProfileRef ? <p className="boundary-note">{t($ => $.publish.snapshotNoRelatedReferences)}</p> : null}<p>{candidate.configuration.description || t($ => $.publish.noPublicDescription)}</p><div className="application-publish-comparison">{comparison.map((item) => <div className={item.before === item.after ? "unchanged" : "changed"} key={item.field}><strong>{item.field}</strong><span>{item.before}</span><span>→</span><span>{item.after}</span></div>)}</div>{!readOnly && candidate.configuration.applicationKind !== "agent" ? <div className="application-draft-handoff"><button type="button" onClick={onIntegration}>{t($ => $.publish.openApiIntegration)}</button><button type="button" onClick={onPlayground}>{t($ => $.publish.testInPlayground)}</button></div> : null}</article>
    <article className="application-publish-eligibility"><div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.publish.promotionEligibility)}</p><h5>{candidate.promotionEligibility.eligible ? t($ => $.publish.eligiblePromotionStatus) : t($ => $.publish.promotionBlockedStatus)}</h5></div><span className={`status-badge ${candidate.promotionEligibility.eligible ? "good" : "bad"}`}>{t($ => $.publish.blockerCount, { count: candidate.promotionEligibility.blockers.length })}</span></div>{candidate.promotionEligibility.blockers.length ? <ul>{candidate.promotionEligibility.blockers.map((blocker) => <li key={blocker.code}><strong>{blocker.code}</strong><p>{blockerMessage(blocker.code)}</p></li>)}</ul> : <p className="boundary-note">{t($ => $.publish.eligibleForExplicitRuntimeDecision)}</p>}</article>
    <article className="application-publish-evidence"><div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.publish.requestHistoryReferences)}</p><h5>{t($ => $.publish.sanitizedReferenceCount, { count: candidate.evidenceRequestIds.length })}</h5></div></div>{candidate.evidenceRequestIds.length ? candidate.evidenceRequestIds.map((requestId) => <button type="button" key={requestId} onClick={() => onHistory(requestId)}><code>{requestId}</code><span>{t($ => $.publish.openExactHistoryDetail)}</span></button>) : <p className="boundary-note">{t($ => $.publish.noGatewayHistoryReferences)}</p>}</article>
    <article className="application-publish-review-log"><div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.publish.appendOnlyReviewLog)}</p><h5>{t($ => $.publish.reviewDecisionCount, { count: candidate.reviews.length })}</h5></div></div>{candidate.reviews.length ? candidate.reviews.map((review) => <div key={review.reviewVersion}><strong>v{review.reviewVersion} · {decisionLabel(review.decision)}</strong><span>{review.reviewerRef} · {review.reviewedAt}</span><p>{review.reason}</p></div>) : <p className="boundary-note">{t($ => $.publish.noReviewDecisionRecorded)}</p>}</article>
  </div>;
}

function PromptTemplateSourceReview({
  applicationId,
  templateRef,
}: {
  applicationId: string;
  templateRef: ApplicationPublishPromptTemplateRef;
}) {
  const { t } = useTranslation("applications");
  const [source, setSource] = useState<PromptTemplateVersion | null>(null);
  const [status, setStatus] = useState<"idle" | "loading" | "verified" | "failed">("idle");
  const [failureCode, setFailureCode] = useState("");

  useEffect(() => {
    setSource(null);
    setStatus("idle");
    setFailureCode("");
  }, [applicationId, templateRef.templateId, templateRef.templateVersion, templateRef.templateDigest]);

  async function reviewExactSource() {
    setStatus("loading");
    setFailureCode("");
    const result = await readPromptTemplateVersion(
      promptTemplateConfig,
      applicationId,
      templateRef.templateId,
      templateRef.templateVersion,
    );
    if (!result.version || result.version.templateDigest !== templateRef.templateDigest) {
      setSource(null);
      setStatus("failed");
      setFailureCode(result.failureCode || "prompt_template_candidate_ref_mismatch");
      return;
    }
    setSource(result.version);
    setStatus("verified");
  }

  return (
    <div className="application-publish-binding prompt-template-source-review">
      <div className="application-api-card-heading">
        <div><strong>{t($ => $.publish.exactImmutablePromptTemplate)}</strong><code>{templateRef.templateId} · v{templateRef.templateVersion}</code></div>
        <button type="button" onClick={() => void reviewExactSource()} disabled={status === "loading"}>
          {status === "loading" ? t($ => $.publish.readingSource) : t($ => $.publish.readExactSource)}
        </button>
      </div>
      <code>{templateRef.templateDigest}</code>
      {failureCode ? <p className="failure-summary">{failureCode}</p> : null}
      {source ? (
        <div className="prompt-template-source">
          <strong>{source.templateName}</strong>
          <p>{source.description || t($ => $.publish.noTemplateDescription)}</p>
          {source.messages.map((message, index) => (
            <div key={`${message.role}-${index}`}>
              <code>{message.role}</code>
              <pre>{message.content}</pre>
            </div>
          ))}
          <small>
            {t($ => $.publish.variablesLabel)}{source.variables.map((variable) => `${variable.name}:${variable.type}${variable.required ? "!" : ""}`).join(", ") || t($ => $.publish.none)}
          </small>
          <pre aria-label={t($ => $.publish.reviewOutputContract)}>{JSON.stringify(source.outputContract, null, 2)}</pre>
        </div>
      ) : <p className="boundary-note">{t($ => $.publish.readTemplateSourceBeforeReview)}</p>}
    </div>
  );
}

function AgentCopilotProfileSourceReview({
  applicationId,
  profileRef,
}: {
  applicationId: string;
  profileRef: ApplicationPublishAgentCopilotProfileRef;
}) {
  const { t } = useTranslation("applications");
  const [source, setSource] = useState<AgentCopilotProfileVersion | null>(null);
  const [status, setStatus] = useState<"idle" | "loading" | "verified" | "failed">("idle");
  const [failureCode, setFailureCode] = useState("");

  useEffect(() => {
    setSource(null);
    setStatus("idle");
    setFailureCode("");
  }, [applicationId, profileRef.policyDigest, profileRef.profileDigest, profileRef.profileId, profileRef.profileVersion]);

  async function reviewExactSource() {
    setStatus("loading");
    setFailureCode("");
    const result = await readAgentCopilotProfileVersion(
      agentProfileConfig,
      applicationId,
      profileRef.profileId,
      profileRef.profileVersion,
    );
    if (!result.version ||
        result.version.profileDigest !== profileRef.profileDigest ||
        result.version.policyDigest !== profileRef.policyDigest) {
      setSource(null);
      setStatus("failed");
      setFailureCode(result.failureCode || "agent_copilot_candidate_profile_ref_mismatch");
      return;
    }
    setSource(result.version);
    setStatus("verified");
  }

  return (
    <div className="application-publish-binding prompt-template-source-review">
      <div className="application-api-card-heading">
        <div><strong>{t($ => $.publish.exactImmutableAgentProfile)}</strong><code>{profileRef.profileId} · v{profileRef.profileVersion}</code></div>
        <button type="button" onClick={() => void reviewExactSource()} disabled={status === "loading"}>
          {status === "loading" ? t($ => $.publish.readingSource) : t($ => $.publish.readExactSource)}
        </button>
      </div>
      <code>{profileRef.profileDigest}</code>
      <code>{profileRef.policyDigest}</code>
      {failureCode ? <p className="failure-summary">{failureCode}</p> : null}
      {source ? (
        <dl className="tenant-meta">
          <div><dt>{t($ => $.publish.project)}</dt><dd>{source.project}</dd></div>
          <div><dt>{t($ => $.publish.tasks)}</dt><dd>{source.allowedTasks.join(", ")}</dd></div>
          <div><dt>{t($ => $.publish.locale)}</dt><dd>{source.defaultLocale}</dd></div>
          <div><dt>{t($ => $.publish.safety)}</dt><dd>{source.riskPolicy.mode} {t($ => $.publish.confirmationRequired)}</dd></div>
        </dl>
      ) : <p className="boundary-note">{t($ => $.publish.readProfileSourceBeforeReview)}</p>}
    </div>
  );
}

function newCandidateId(applicationId: string): string {
  const suffix = (globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random()}`).replaceAll("-", "").slice(0, 10);
  return `publish-${applicationId}-${suffix}`;
}
