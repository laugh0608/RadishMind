import type { StructuredRuntimeInputFieldError } from "./structuredRuntimeInput.ts";
import "../../i18n/workflowDraftResources.ts";
import { workflowDraftStatusLabel } from "./workflowDraftMessages.ts";
import { workflowPromotionMessage, type WorkflowPromotionMessage } from "./workflowOperationMessages.ts";
import "../../i18n/workflowPromotionResources.ts";
import { useTranslation } from "react-i18next";
import { useEffect, useMemo, useRef, useState } from "react";
import { useLocalIdentity, type LocalIdentityContextValue } from "../local-identity/localIdentityContext.ts";

import type { WorkflowDraftDesignerDraft } from "./workflowDraftDesigner.ts";
import {
  WorkflowDefinitionPromotionConflict,
  WorkflowDefinitionPromotionFailure,
  createWorkflowDefinitionCandidate,
  decideWorkflowDefinitionActivation,
  decideWorkflowDefinitionCandidate,
  deriveWorkflowDraftFromDefinitionVersion,
  evaluateWorkflowDefinitionCandidateCompatibility,
  listWorkflowDefinitionCandidates,
  readWorkflowDefinitionPromotionAuthority,
  readWorkflowDefinitionPromotionConfig,
  startWorkflowDefinitionRun,
  type WorkflowDefinitionPromotionAuthority,
  type WorkflowDefinitionCandidate,
} from "./workflowDefinitionPromotionConsumer.ts";
import type { ApplicationDevelopmentOwnerEvidence } from "./applicationDevelopmentReadiness.ts";
import WorkflowDefinitionHTTPToolRuntimePanel from "./workflowDefinitionHTTPToolRuntimePanel.tsx";
import StructuredRuntimeInputEditor from "./StructuredRuntimeInputEditor.tsx";
import {
  structuredRuntimeInputAuthorityKey,
  validateStructuredRuntimeInputDrafts,
  type StructuredRuntimeInputContract,
  type StructuredRuntimeInputDrafts,
} from "./structuredRuntimeInput.ts";

const config = readWorkflowDefinitionPromotionConfig();

type Props = {
  workspaceId: string;
  applicationId: string;
  activeDraft: WorkflowDraftDesignerDraft;
  savedDraftVersion: number;
  savedDraftLifecycleVersion: number;
  savedDraftLifecycleState: "active" | "archived" | "unknown";
  nextDerivedDraftNumber: number;
  onDerivedDraft: (draft: WorkflowDraftDesignerDraft) => void;
  onRunRecorded: (runId: string) => void;
  onOpenRun?: (runId: string) => void;
  onEvidenceChange?: (evidence: ApplicationDevelopmentOwnerEvidence) => void;
};

export default function WorkflowDefinitionPromotionPanel(props: Props) {
  const identity = useLocalIdentity();
  const scopeKey = JSON.stringify([props.workspaceId, props.applicationId,
    identity?.profile.account.userId, identity?.authorityRevision]);
  return <WorkflowDefinitionPromotionScope key={scopeKey} {...props} identity={identity} />;
}

function WorkflowDefinitionPromotionScope({ workspaceId, applicationId, activeDraft, savedDraftVersion, savedDraftLifecycleVersion, savedDraftLifecycleState, nextDerivedDraftNumber, onDerivedDraft, onRunRecorded, onOpenRun, onEvidenceChange, identity }: Props & { identity: LocalIdentityContextValue | null }) {
  const { t } = useTranslation("workflow");
  const requestEpoch = useRef(0);
  const pendingRequest = useRef(false);
  const liveConfig = useMemo(() => ({ ...config, workspaceId }), [workspaceId]);
  const [candidates, setCandidates] = useState<WorkflowDefinitionCandidate[]>([]);
  const [selectedCandidateId, setSelectedCandidateId] = useState("");
  const selectedCandidate = candidates.find((candidate) => candidate.candidateId === selectedCandidateId) ?? null;
  const [authority, setAuthority] = useState<WorkflowDefinitionPromotionAuthority | null>(null);
  const [authorityState, setAuthorityState] = useState<"empty" | "loading" | "ready" | "stale" | "failed">("empty");
  const authorityReady = authorityState === "ready" && authority !== null && authority.definitionId === selectedCandidate?.definitionId;
  const versions = authorityReady ? authority.versions : [];
  const activation = authorityReady ? authority.activation : null;
  const [candidateId, setCandidateId] = useState(defaultCandidateId(activeDraft.draftId, savedDraftVersion));
  const [definitionId, setDefinitionId] = useState(activeDraft.workflowDefinitionId || defaultDefinitionId(activeDraft.draftId));
  const [reviewDecision, setReviewDecision] = useState<"approve" | "reject">("approve");
  const [activationDecision, setActivationDecision] = useState<"activate" | "replace" | "deactivate">("activate");
  const [selectedVersion, setSelectedVersion] = useState(1);
  const [reason, setReason] = useState("Reviewed immutable workflow definition evidence.");
  const [inputText, setInputText] = useState("");
  const [structuredInputDrafts, setStructuredInputDrafts] = useState<StructuredRuntimeInputDrafts>({});
  const [structuredInputErrors, setStructuredInputErrors] = useState<Record<string, StructuredRuntimeInputFieldError>>({});
  const [model, setModel] = useState("");
  const [conditionValues, setConditionValues] = useState<Record<string, boolean>>({});
  const [advisoryOutput, setAdvisoryOutput] = useState("");
  const [lastRunId, setLastRunId] = useState("");
  const [pending, setPending] = useState("");
  const [notice, setNotice] = useState<WorkflowPromotionMessage | null>(null);
  const [failure, setFailure] = useState<WorkflowPromotionMessage | null>(null);
  const candidateCompatibility = useMemo(() => evaluateWorkflowDefinitionCandidateCompatibility(activeDraft), [activeDraft]);
  const canCreateCandidate = candidateCompatibility.compatible && savedDraftVersion > 0
    && savedDraftLifecycleVersion > 0 && savedDraftLifecycleState === "active";
  const activeVersion = activation?.state === "active"
    ? versions.find((version) => version.version === activation.activeVersion && version.definitionDigest === activation.activeDefinitionDigest) ?? null
    : null;
  const selectedDefinitionVersion = versions.find((version) => version.version === selectedVersion);
  const canDecideActivation = authorityReady && (activationDecision === "deactivate"
    ? activation?.state === "active"
    : Boolean(selectedDefinitionVersion?.activationEligible) && (activationDecision === "activate"
      ? activation?.state !== "active"
      : activation?.state === "active" && selectedVersion !== activation.activeVersion));
  const structuredInputContract = activeVersion?.snapshot.executionProfile === "workflow_definition_executor_v2"
    ? activeVersion.snapshot.inputContract as StructuredRuntimeInputContract : null;
  const definitionHTTPToolActive = activeVersion?.snapshot.executionProfile === "workflow_definition_http_tool_v1";
  const renderedEpoch = requestEpoch.current;

  function isCurrent(epoch: number) {
    return requestEpoch.current === epoch && (!identity || identity.readAuthorityRevision() === identity.authorityRevision);
  }

  function clearRuntimeInput() {
    setInputText("");
    setModel("");
    setStructuredInputDrafts({});
    setStructuredInputErrors({});
    setConditionValues({});
    setAdvisoryOutput("");
    setLastRunId("");
  }

  function applyAuthority(next: WorkflowDefinitionPromotionAuthority, epoch: number, preferredVersion?: number) {
    if (!isCurrent(epoch)) return;
    setAuthority(next);
    setAuthorityState("ready");
    setSelectedVersion(next.versions.some((version) => version.version === preferredVersion)
      ? preferredVersion! : next.activation?.activeVersion || next.versions.at(-1)?.version || 1);
    setActivationDecision((decision) => next.activation?.state === "active"
      ? decision === "activate" ? "replace" : decision : "activate");
  }

  useEffect(() => {
    if (!onEvidenceChange) return;
    const ownerFailed = Boolean(failure) || authorityState === "failed" || authorityState === "stale";
    onEvidenceChange({
      contributionId: "workflow_definition",
      status: ownerFailed ? "blocked" : activeVersion ? "available" : "incomplete",
      coverage: activeVersion || ownerFailed ? "complete" : "none",
      evidenceRefs: activeVersion ? [{ kind: "definition", id: activeVersion.definitionId, version: activeVersion.version }] : [],
      missingEvidence: activeVersion ? [] : ["Approve and activate an immutable Workflow Definition version."],
      blockers: ownerFailed ? [{ code: "workflow_definition_owner_failure", summary: failure && "failureCode" in failure ? failure.failureCode : failure?.code ?? "workflow_definition_owner_failure" }] : [],
      failureCodes: ownerFailed ? ["workflow_definition_owner_failure"] : [],
    });
  }, [activeVersion, authorityState, failure, onEvidenceChange]);

  useEffect(() => {
    onEvidenceChange?.({
      contributionId: "controlled_run",
      status: lastRunId ? "available" : "incomplete",
      coverage: lastRunId ? "complete" : "none",
      evidenceRefs: lastRunId ? [{ kind: "run", id: lastRunId }] : [],
      missingEvidence: lastRunId ? [] : ["Run the exact active Workflow Definition version."],
      blockers: [], failureCodes: [],
    });
  }, [lastRunId, onEvidenceChange]);

  useEffect(() => {
    if (liveConfig.mode === "offline" || !applicationId) return;
    void runOperation("loading", async (epoch) => { await refresh("", "", epoch); });
    return () => { requestEpoch.current += 1; pendingRequest.current = false; };
  }, [applicationId, liveConfig]);

  useEffect(() => {
    setCandidateId(defaultCandidateId(activeDraft.draftId, savedDraftVersion));
    setDefinitionId(activeDraft.workflowDefinitionId || defaultDefinitionId(activeDraft.draftId));
  }, [activeDraft.draftId, activeDraft.workflowDefinitionId, savedDraftVersion]);

  useEffect(() => {
    clearRuntimeInput();
    setConditionValues(Object.fromEntries((activeVersion?.snapshot.nodes ?? [])
      .filter((node) => node.nodeType === "condition").map((node) => [node.nodeId, false])));
  }, [activeVersion?.definitionId, activeVersion?.version, activation?.pointerVersion,
    structuredInputContract ? structuredRuntimeInputAuthorityKey(structuredInputContract) : "legacy"]);

  // Each read and mutation, including its refresh and finalizer, belongs to one selection epoch.
  async function runOperation(name: string, operation: (epoch: number) => Promise<void>, replaceSelection = false) {
    if ((!replaceSelection && pendingRequest.current) || (identity && identity.readAuthorityRevision() !== identity.authorityRevision)) return;
    const epoch = ++requestEpoch.current;
    pendingRequest.current = true;
    setPending(name);
    setFailure(null);
    setNotice(null);
    try {
      await operation(epoch);
    } catch (error: unknown) {
      if (!isCurrent(epoch)) return;
      setAuthority(null);
      setAuthorityState(error instanceof WorkflowDefinitionPromotionConflict ? "stale" : "failed");
      setFailure(promotionFailure(error));
    } finally {
      if (isCurrent(epoch)) { pendingRequest.current = false; setPending(""); }
    }
  }

  async function selectCandidate(candidate: WorkflowDefinitionCandidate) {
    if (candidate.candidateId === selectedCandidateId) return;
    await runOperation("authority", async (epoch) => {
      setSelectedCandidateId(candidate.candidateId);
      setAuthority(null);
      setAuthorityState("loading");
      setReason("Reviewed immutable workflow definition evidence.");
      setReviewDecision("approve");
      clearRuntimeInput();
      applyAuthority(await readWorkflowDefinitionPromotionAuthority(liveConfig, applicationId, candidate.definitionId), epoch);
    }, true);
  }

  async function refresh(candidate: string, definition: string, epoch: number, preferredVersion?: number) {
    setAuthorityState("loading");
    const [nextCandidates, knownAuthority] = await Promise.all([
      listWorkflowDefinitionCandidates(liveConfig, applicationId),
      definition ? readWorkflowDefinitionPromotionAuthority(liveConfig, applicationId, definition) : Promise.resolve(null),
    ]);
    if (!isCurrent(epoch)) return;
    const selected = candidate ? nextCandidates.find((item) => item.candidateId === candidate) : nextCandidates[0];
    if (candidate && (!selected || selected.definitionId !== definition)) {
      throw new WorkflowDefinitionPromotionFailure("workflow_definition_selection_unavailable", "The selected candidate is no longer available in this lineage.");
    }
    const nextAuthority = knownAuthority ?? (selected ? await readWorkflowDefinitionPromotionAuthority(liveConfig, applicationId, selected.definitionId) : null);
    if (!isCurrent(epoch)) return;
    setCandidates(nextCandidates);
    setSelectedCandidateId(selected?.candidateId ?? "");
    if (nextAuthority) applyAuthority(nextAuthority, epoch, preferredVersion);
    else { setAuthority(null); setAuthorityState("empty"); }
  }

  async function refreshSelection() {
    await runOperation("refresh", async (epoch) => {
      await refresh(selectedCandidate?.candidateId ?? "", selectedCandidate?.definitionId ?? "", epoch, selectedVersion);
      if (isCurrent(epoch)) setNotice({ code: "refreshed" });
    });
  }

  async function createCandidate() {
    if (!canCreateCandidate) return;
    await runOperation("create", async (epoch) => {
      const created = await createWorkflowDefinitionCandidate(liveConfig, applicationId, {
        candidateId, definitionId, draftId: activeDraft.draftId, expectedDraftVersion: savedDraftVersion,
        expectedLifecycleVersion: savedDraftLifecycleVersion, executionProfile: candidateCompatibility.executionProfile,
      });
      if (!isCurrent(epoch)) return;
      await refresh(created.candidateId, created.definitionId, epoch);
      if (isCurrent(epoch)) setNotice({ code: "created", id: created.candidateId, version: created.sourceDraftVersion });
    });
  }

  async function decideCandidate() {
    if (!authorityReady || !selectedCandidate || selectedCandidate.state !== "pending") return;
    await runOperation("review", async (epoch) => {
      await decideWorkflowDefinitionCandidate(liveConfig, applicationId, selectedCandidate.candidateId, { expectedReviewVersion: selectedCandidate.reviewVersion, decision: reviewDecision, reason });
      if (!isCurrent(epoch)) return;
      await refresh(selectedCandidate.candidateId, selectedCandidate.definitionId, epoch, selectedVersion);
      if (isCurrent(epoch)) setNotice({ code: "reviewed", decision: reviewDecision });
    });
  }

  async function decideActivation() {
    if (!canDecideActivation || !selectedCandidate) return;
    await runOperation("activation", async (epoch) => {
      await decideWorkflowDefinitionActivation(liveConfig, applicationId, selectedCandidate.definitionId, { expectedPointerVersion: activation?.pointerVersion ?? 0, decision: activationDecision, version: activationDecision === "deactivate" ? 0 : selectedVersion, reason });
      if (!isCurrent(epoch)) return;
      await refresh(selectedCandidate.candidateId, selectedCandidate.definitionId, epoch, selectedVersion);
      if (isCurrent(epoch)) setNotice({ code: "activated", decision: activationDecision, version: (activation?.pointerVersion ?? 0) + 1 });
    });
  }

  async function startRun() {
    if (!authorityReady || !activeVersion || !activation || activation.state !== "active" || pendingRequest.current) return;
    const structuredValidation = structuredInputContract ? validateStructuredRuntimeInputDrafts(structuredInputContract, structuredInputDrafts) : null;
    if (structuredValidation && !structuredValidation.ok) {
      setStructuredInputErrors(structuredValidation.fieldErrors);
      setFailure({ code: "inputInvalid", failureCode: structuredValidation.failureCode });
      return;
    }
    if (structuredInputContract) { setStructuredInputDrafts({}); setStructuredInputErrors({}); }
    setAdvisoryOutput("");
    setLastRunId("");
    await runOperation("run", async (epoch) => {
      const exact = { definitionId: activeVersion.definitionId, expectedPointerVersion: activation.pointerVersion, expectedDefinitionVersion: activeVersion.version, expectedDefinitionDigest: activeVersion.definitionDigest, conditionValues, model };
      const result = activeVersion.snapshot.executionProfile === "workflow_definition_executor_v2"
        ? await startWorkflowDefinitionRun(liveConfig, applicationId, { ...exact, executionProfile: "workflow_definition_executor_v2", inputs: structuredValidation!.inputs })
        : await startWorkflowDefinitionRun(liveConfig, applicationId, { ...exact, executionProfile: "workflow_definition_executor_v1", inputText });
      if (!isCurrent(epoch)) return;
      setLastRunId(result.record.runId);
      setAdvisoryOutput(result.advisoryOutput);
      setInputText("");
      setNotice({ code: "runFinished", schema: result.record.schemaVersion, id: result.record.runId });
      onRunRecorded(result.record.runId);
    });
  }

  if (liveConfig.mode === "offline") {
    return <section className="workflow-definition-promotion-panel offline" id="workflow-definition-promotion"><div className="section-heading compact-heading"><div><p className="eyebrow">{t($ => $.promotion.workflowDefinitionPromotion)}</p><h4>{t($ => $.promotion.immutableVersionPromotionIsDisabled)}</h4></div><span className="status-badge neutral">{t($ => $.promotion.offlineZeroRequests)}</span></div><p>{t($ => $.promotion.enableTheUnifiedLocalProductProfileToCreateCandidatesReview)}</p></section>;
  }

  return <section className="workflow-definition-promotion-panel" id="workflow-definition-promotion" aria-labelledby="workflow-definition-promotion-title">
    <div className="section-heading compact-heading"><div><p className="eyebrow">{t($ => $.promotion.workflowDefinitionControlledRuntime)}</p><h4 id="workflow-definition-promotion-title">{t($ => $.promotion.immutableVersionPromotionAndExactExecution)}</h4></div><span className={`status-badge ${activation?.state === "active" ? "status-good" : "status-neutral"}`}>{workflowDraftStatusLabel(t, activation?.state ?? "inactive")}</span></div>
    <p className="boundary-note">{t($ => $.promotion.savedDraftRemainsEditableCandidatesVersionsActivationAndV5V8)}</p>
    {failure ? <p className="workflow-definition-failure" role="alert">{workflowPromotionMessage(t, failure)}</p> : null}
    {notice ? <p className="workflow-definition-notice" aria-live="polite">{workflowPromotionMessage(t, notice)}</p> : null}
    <div className="workflow-definition-refresh">
      <button type="button" disabled={Boolean(pending)} onClick={() => void refreshSelection()}>{t($ => $.promotion.reloadAuthority)}</button>
      {authorityState === "loading" ? <p role="status">{t($ => $.promotion.loadingAuthority)}</p> : null}
      {authorityState === "stale" || authorityState === "failed" ? <p>{t($ => $.promotion.authorityUnavailable)}</p> : null}
    </div>
    <div className="workflow-definition-promotion-grid">
      <article>
        <p className="eyebrow">{t($ => $.promotion.step1Candidate)}</p><h5>{t($ => $.promotion.createACandidateFromASavedDraft)}</h5>
        <label>{t($ => $.promotion.candidateID)}<input value={candidateId} onChange={(event) => setCandidateId(event.currentTarget.value)} /></label>
        <label>{t($ => $.promotion.definitionID)}<input value={definitionId} onChange={(event) => setDefinitionId(event.currentTarget.value)} /></label>
        <dl><div><dt>{t($ => $.promotion.draft)}</dt><dd>{t($ => $.promotion.draftVersion, { id: activeDraft.draftId, content: savedDraftVersion, lifecycle: savedDraftLifecycleVersion, state: workflowDraftStatusLabel(t, savedDraftLifecycleState) })}</dd></div><div><dt>{t($ => $.promotion.provenance)}</dt><dd>{t($ => $.promotion.baseVersion, { lineage: activeDraft.workflowDefinitionId || t($ => $.promotion.newLineage), version: activeDraft.baseDefinitionVersion ?? 0 })}</dd></div></dl>
        {!candidateCompatibility.compatible ? <div className="workflow-definition-candidate-handoff"><strong>{t($ => $.promotion.thisDraftCannotBecomeADefinitionCandidate)}</strong><p>{workflowPromotionMessage(t, { code: "incompatible", compatibility: candidateCompatibility })}</p>{candidateCompatibility.handoffAnchor ? <a href={`#${candidateCompatibility.handoffAnchor}`}>{t($ => $.promotion.openWorkflowRAGPromotion)}</a> : null}</div> : null}
        <button type="button" disabled={Boolean(pending) || !canCreateCandidate} onClick={() => void createCandidate()}>{t($ => $.promotion.createPromotionCandidate)}</button>
        <div className="workflow-definition-list">{candidates.map((candidate) => <button type="button" className={candidate.candidateId === selectedCandidate?.candidateId ? "selected" : ""} key={candidate.candidateId} onClick={() => void selectCandidate(candidate)}><strong>{candidate.candidateId}</strong><span>{t($ => $.promotion.candidateState, { state: workflowDraftStatusLabel(t, candidate.state), version: candidate.reviewVersion })}</span></button>)}</div>
      </article>
      <article>
        <p className="eyebrow">{t($ => $.promotion.step2Review)}</p><h5>{t($ => $.promotion.humanReviewAndImmutableVersions)}</h5>
        {selectedCandidate ? <><dl><div><dt>{t($ => $.promotion.definition)}</dt><dd>{selectedCandidate.definitionId}</dd></div><div><dt>{t($ => $.promotion.profile)}</dt><dd>{selectedCandidate.snapshot.executionProfile}</dd></div><div><dt>{t($ => $.promotion.digest)}</dt><dd><code>{shortDigest(selectedCandidate.definitionDigest)}</code></dd></div><div><dt>{t($ => $.promotion.eligibility)}</dt><dd>{selectedCandidate.activationEligible ? workflowDraftStatusLabel(t, "eligible") : selectedCandidate.eligibilityBlockers.join(", ")}</dd></div></dl>
          <label>{t($ => $.promotion.decision)}<select value={reviewDecision} onChange={(event) => setReviewDecision(event.currentTarget.value as "approve" | "reject")}><option value="approve">{t($ => $.promotion.approve)}</option><option value="reject">{t($ => $.promotion.reject)}</option></select></label>
          <label>{t($ => $.promotion.reason)}<textarea value={reason} onChange={(event) => setReason(event.currentTarget.value)} /></label>
          <button type="button" disabled={Boolean(pending) || !authorityReady || selectedCandidate.state !== "pending"} onClick={() => void decideCandidate()}>{t($ => $.promotion.appendReview, { version: selectedCandidate.reviewVersion + 1 })}</button>
          <div className="workflow-definition-evidence">{selectedCandidate.reviews.map((review) => <p key={review.reviewVersion}><strong>v{review.reviewVersion} · {workflowDraftStatusLabel(t, review.decision)}</strong><span>{review.reason}</span></p>)}</div>
        </> : <p>{t($ => $.promotion.noPromotionCandidatesExistForThisApplication)}</p>}
      </article>
      <article>
        <p className="eyebrow">{t($ => $.promotion.step3Activation)}</p><h5>{t($ => $.promotion.versionHistoryAndPointerCAS)}</h5>
        <div className="workflow-definition-list">{versions.map((version) => <button type="button" className={selectedVersion === version.version ? "selected" : ""} key={version.version} disabled={Boolean(pending)} onClick={() => setSelectedVersion(version.version)}><strong>v{version.version}</strong><span>{shortDigest(version.definitionDigest)} · {workflowDraftStatusLabel(t, version.activationEligible ? "eligible" : "blocked")}</span></button>)}</div>
        <label>{t($ => $.promotion.decision)}<select value={activationDecision} onChange={(event) => setActivationDecision(event.currentTarget.value as "activate" | "replace" | "deactivate")}><option value="activate" disabled={activation?.state === "active"}>{t($ => $.promotion.activate)}</option><option value="replace" disabled={activation?.state !== "active"}>{t($ => $.promotion.replace)}</option><option value="deactivate" disabled={activation?.state !== "active"}>{t($ => $.promotion.deactivate)}</option></select></label>
        <button type="button" disabled={Boolean(pending) || !canDecideActivation} onClick={() => void decideActivation()}>{t($ => $.promotion.pointerDecision, { decision: workflowDraftStatusLabel(t, activationDecision), version: activation?.pointerVersion ?? 0 })}</button>
        <p>{authorityReady ? t($ => $.promotion.currentPointer, { state: workflowDraftStatusLabel(t, activation?.state ?? "inactive"), active: activation?.activeVersion ?? 0, pointer: activation?.pointerVersion ?? 0 }) : t($ => $.promotion.noCurrentAuthority)}</p>
        {versions.some((version) => version.version === selectedVersion) ? <button type="button" disabled={Boolean(pending)} onClick={() => onDerivedDraft(deriveWorkflowDraftFromDefinitionVersion(versions.find((version) => version.version === selectedVersion)!, applicationId, nextDerivedDraftNumber))}>{t($ => $.promotion.deriveVersion, { version: selectedVersion })}</button> : null}
      </article>
      <article>
        <p className="eyebrow">{t($ => $.promotion.step4DefinitionBoundRun)}</p><h5>{t($ => $.promotion.runOnlyFromTheExactActiveVersion)}</h5>
        <dl><div><dt>{t($ => $.promotion.profile)}</dt><dd>{activeVersion?.snapshot.executionProfile ?? t($ => $.promotion.noActiveProfile)}</dd></div><div><dt>{t($ => $.promotion.authority)}</dt><dd>{activeVersion ? `${activeVersion.definitionId} · v${activeVersion.version}` : t($ => $.promotion.noActiveAuthority)}</dd></div></dl>
        {definitionHTTPToolActive ? <p className="boundary-note">{t($ => $.promotion.theActiveDefinitionUsesTheControlledHTTPToolProfileFollow)}</p> : <>
          {structuredInputContract ? <StructuredRuntimeInputEditor contract={structuredInputContract} drafts={structuredInputDrafts} fieldErrors={structuredInputErrors} disabled={Boolean(pending)} onChange={(drafts) => { setStructuredInputDrafts(drafts); setStructuredInputErrors({}); }} /> : <label>{t($ => $.promotion.oneTimeInput)}<textarea disabled={Boolean(pending) || !activeVersion} value={inputText} onChange={(event) => setInputText(event.currentTarget.value)} /></label>}
          {activeVersion?.snapshot.nodes.filter((node) => node.nodeType === "condition").map((node) => <label className="workflow-definition-condition" key={node.nodeId}><input type="checkbox" disabled={Boolean(pending)} checked={conditionValues[node.nodeId] ?? false} onChange={(event) => setConditionValues((values) => ({ ...values, [node.nodeId]: event.currentTarget.checked }))} />{node.label} · {node.nodeId}</label>)}
          <label>{t($ => $.promotion.modelOptional)}<input disabled={Boolean(pending) || !activeVersion} value={model} onChange={(event) => setModel(event.currentTarget.value)} /></label>
          <button type="button" disabled={Boolean(pending) || !activeVersion || (!structuredInputContract && !inputText.trim())} onClick={() => void startRun()}>{t($ => $.promotion.startExactVersionRun)}</button>
          {lastRunId ? <div className="workflow-definition-run-result"><strong>{lastRunId}</strong><p>{advisoryOutput || t($ => $.promotion.noOutput)}</p><button type="button" onClick={() => onOpenRun?.(lastRunId)}>{t($ => $.promotion.openRunHistory)}</button><button type="button" onClick={() => { setAdvisoryOutput(""); setLastRunId(""); }}>{t($ => $.promotion.clearOneTimeResult)}</button></div> : null}
        </>}
      </article>
    </div>
    {definitionHTTPToolActive && activeVersion && activation?.state === "active" ? <WorkflowDefinitionHTTPToolRuntimePanel key={`${activeVersion.definitionId}:${activeVersion.version}:${activation.pointerVersion}`} workspaceId={workspaceId} applicationId={applicationId} version={activeVersion} activationPointerVersion={activation.pointerVersion} onRunRecorded={(runId) => { if (isCurrent(renderedEpoch)) { setLastRunId(runId); onRunRecorded(runId); } }} onOpenRun={onOpenRun} /> : null}
  </section>;
}

function defaultCandidateId(draftId: string, version: number): string { return `wdrc_${safePart(draftId)}_v${Math.max(1, version)}`.slice(0, 150); }
function defaultDefinitionId(draftId: string): string { return `wdef_${safePart(draftId)}`.slice(0, 150); }
function safePart(value: string): string { return value.toLowerCase().replace(/[^a-z0-9]+/gu, "_").replace(/^_+|_+$/gu, "").slice(0, 64) || "workflow"; }
function shortDigest(value: string): string { return value.length > 20 ? `${value.slice(0, 19)}…` : value; }
function promotionFailure(error: unknown): WorkflowPromotionMessage {
  return error instanceof WorkflowDefinitionPromotionConflict
    ? { code: "conflict", failureCode: error.failureCode, review: error.currentReviewVersion, pointer: error.currentPointerVersion }
    : { code: "failed", failureCode: error instanceof WorkflowDefinitionPromotionFailure ? error.failureCode : "workflow_definition_operation_failed" };
}
