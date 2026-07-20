import { useEffect, useMemo, useRef, useState } from "react";

import type { WorkflowDraftDesignerDraft } from "./workflowDraftDesigner.ts";
import {
  WorkflowDefinitionPromotionConflict,
  createWorkflowDefinitionCandidate,
  decideWorkflowDefinitionActivation,
  decideWorkflowDefinitionCandidate,
  deriveWorkflowDraftFromDefinitionVersion,
  listWorkflowDefinitionCandidates,
  listWorkflowDefinitionVersions,
  readWorkflowDefinitionActivation,
  readWorkflowDefinitionPromotionConfig,
  startWorkflowDefinitionRun,
  type WorkflowDefinitionActivation,
  type WorkflowDefinitionCandidate,
  type WorkflowDefinitionVersion,
} from "./workflowDefinitionPromotionConsumer.ts";
import type { ApplicationDevelopmentOwnerEvidence } from "./applicationDevelopmentReadiness.ts";

const config = readWorkflowDefinitionPromotionConfig();

type Props = {
  applicationId: string;
  activeDraft: WorkflowDraftDesignerDraft;
  savedDraftVersion: number;
  nextDerivedDraftNumber: number;
  onDerivedDraft: (draft: WorkflowDraftDesignerDraft) => void;
  onRunRecorded: (runId: string) => void;
  onOpenRun?: (runId: string) => void;
  onEvidenceChange?: (evidence: ApplicationDevelopmentOwnerEvidence) => void;
};

export default function WorkflowDefinitionPromotionPanel({ applicationId, activeDraft, savedDraftVersion, nextDerivedDraftNumber, onDerivedDraft, onRunRecorded, onOpenRun, onEvidenceChange }: Props) {
  const requestEpoch = useRef(0);
  const [candidates, setCandidates] = useState<WorkflowDefinitionCandidate[]>([]);
  const [selectedCandidateId, setSelectedCandidateId] = useState("");
  const selectedCandidate = candidates.find((candidate) => candidate.candidateId === selectedCandidateId) ?? candidates[0] ?? null;
  const [versions, setVersions] = useState<WorkflowDefinitionVersion[]>([]);
  const [activation, setActivation] = useState<WorkflowDefinitionActivation | null>(null);
  const [candidateId, setCandidateId] = useState("");
  const [definitionId, setDefinitionId] = useState("");
  const [reviewDecision, setReviewDecision] = useState<"approve" | "reject">("approve");
  const [activationDecision, setActivationDecision] = useState<"activate" | "replace" | "deactivate">("activate");
  const [selectedVersion, setSelectedVersion] = useState(1);
  const [reason, setReason] = useState("Reviewed immutable workflow definition evidence.");
  const [inputText, setInputText] = useState("Generate a bounded advisory response from the exact active workflow definition.");
  const [model, setModel] = useState("");
  const [conditionValues, setConditionValues] = useState<Record<string, boolean>>({});
  const [advisoryOutput, setAdvisoryOutput] = useState("");
  const [lastRunId, setLastRunId] = useState("");
  const [pending, setPending] = useState("");
  const [notice, setNotice] = useState("");
  const [failure, setFailure] = useState("");

  const activeVersion = useMemo(
    () => versions.find((version) => activation?.state === "active" && version.version === activation.activeVersion) ?? null,
    [activation, versions],
  );

  useEffect(() => {
    if (!onEvidenceChange) return;
    const ownerFailed = Boolean(failure);
    const active = activation?.state === "active" && Boolean(activeVersion);
    onEvidenceChange({
      contributionId: "workflow_definition",
      status: ownerFailed ? "blocked" : active ? "available" : "incomplete",
      coverage: activeVersion || ownerFailed ? "complete" : "none",
      evidenceRefs: activeVersion ? [{ kind: "definition", id: activeVersion.definitionId, version: activeVersion.version }] : [],
      missingEvidence: active ? [] : ["Approve and activate an immutable Workflow Definition version."],
      blockers: ownerFailed ? [{ code: "workflow_definition_owner_failure", summary: failure }] : [],
      failureCodes: ownerFailed ? ["workflow_definition_owner_failure"] : [],
    });
  }, [activation?.state, activeVersion, failure, onEvidenceChange]);

  useEffect(() => {
    if (!onEvidenceChange || !lastRunId) return;
    onEvidenceChange({
      contributionId: "controlled_run",
      status: "available",
      coverage: "complete",
      evidenceRefs: [{ kind: "run", id: lastRunId }],
      missingEvidence: [],
      blockers: [],
      failureCodes: [],
    });
  }, [lastRunId, onEvidenceChange]);

  useEffect(() => {
    requestEpoch.current += 1;
    const epoch = requestEpoch.current;
    setCandidates([]);
    setSelectedCandidateId("");
    setVersions([]);
    setActivation(null);
    setCandidateId(defaultCandidateId(activeDraft.draftId, savedDraftVersion));
    setDefinitionId(activeDraft.workflowDefinitionId || defaultDefinitionId(activeDraft.draftId));
    setConditionValues({});
    setAdvisoryOutput("");
    setLastRunId("");
    setFailure("");
    setNotice("");
    if (config.mode === "offline" || !applicationId) return;
    setPending("loading");
    listWorkflowDefinitionCandidates(config, applicationId)
      .then((items) => {
        if (requestEpoch.current !== epoch) return;
        setCandidates(items);
        setSelectedCandidateId(items[0]?.candidateId ?? "");
      })
      .catch((error: unknown) => { if (requestEpoch.current === epoch) setFailure(message(error)); })
      .finally(() => { if (requestEpoch.current === epoch) setPending(""); });
  }, [applicationId]);

  useEffect(() => {
    setCandidateId(defaultCandidateId(activeDraft.draftId, savedDraftVersion));
    setDefinitionId(activeDraft.workflowDefinitionId || defaultDefinitionId(activeDraft.draftId));
  }, [activeDraft.draftId, activeDraft.workflowDefinitionId, savedDraftVersion]);

  useEffect(() => {
    const definition = selectedCandidate?.definitionId ?? "";
    if (config.mode === "offline" || !definition) {
      setVersions([]);
      setActivation(null);
      return;
    }
    const epoch = requestEpoch.current;
    setPending("authority");
    Promise.all([
      listWorkflowDefinitionVersions(config, applicationId, definition),
      readWorkflowDefinitionActivation(config, applicationId, definition),
    ]).then(([nextVersions, nextActivation]) => {
      if (requestEpoch.current !== epoch) return;
      setVersions(nextVersions);
      setActivation(nextActivation);
      setSelectedVersion(nextActivation?.activeVersion || nextVersions.at(-1)?.version || 1);
    }).catch((error: unknown) => { if (requestEpoch.current === epoch) setFailure(message(error)); })
      .finally(() => { if (requestEpoch.current === epoch) setPending(""); });
  }, [applicationId, selectedCandidate?.definitionId]);

  useEffect(() => {
    setConditionValues(Object.fromEntries(
      (activeVersion?.snapshot.nodes ?? [])
        .filter((node) => node.nodeType === "condition")
        .map((node) => [node.nodeId, false]),
    ));
  }, [activeVersion?.definitionId, activeVersion?.version]);

  async function refresh(definition = selectedCandidate?.definitionId ?? "") {
    const nextCandidates = await listWorkflowDefinitionCandidates(config, applicationId);
    setCandidates(nextCandidates);
    if (definition) {
      setVersions(await listWorkflowDefinitionVersions(config, applicationId, definition));
      setActivation(await readWorkflowDefinitionActivation(config, applicationId, definition));
    }
  }

  async function createCandidate() {
    if (savedDraftVersion < 1) {
      setFailure("必须先保存并校验当前草案，才能创建晋级候选。");
      return;
    }
    await runOperation("create", async () => {
      const created = await createWorkflowDefinitionCandidate(config, applicationId, { candidateId, definitionId, draftId: activeDraft.draftId, expectedDraftVersion: savedDraftVersion });
      await refresh(created.definitionId);
      setSelectedCandidateId(created.candidateId);
      setNotice(`候选 ${created.candidateId} 已从精确草案 v${created.sourceDraftVersion} 创建。`);
    });
  }

  async function decideCandidate() {
    if (!selectedCandidate) return;
    await runOperation("review", async () => {
      await decideWorkflowDefinitionCandidate(config, applicationId, selectedCandidate.candidateId, { expectedReviewVersion: selectedCandidate.reviewVersion, decision: reviewDecision, reason });
      await refresh(selectedCandidate.definitionId);
      setNotice(`${reviewDecision} 已追加到候选审查历史；批准不会自动激活。`);
    });
  }

  async function decideActivation() {
    if (!selectedCandidate) return;
    await runOperation("activation", async () => {
      const next = await decideWorkflowDefinitionActivation(config, applicationId, selectedCandidate.definitionId, { expectedPointerVersion: activation?.pointerVersion ?? 0, decision: activationDecision, version: activationDecision === "deactivate" ? 0 : selectedVersion, reason });
      setActivation(next);
      setNotice(`${activationDecision} 已通过 pointer CAS 写入 v${next.pointerVersion}。`);
    });
  }

  async function startRun() {
    if (!activeVersion || !activation || activation.state !== "active") return;
    const epoch = requestEpoch.current;
    await runOperation("run", async () => {
      const result = await startWorkflowDefinitionRun(config, applicationId, { definitionId: activeVersion.definitionId, expectedPointerVersion: activation.pointerVersion, expectedDefinitionVersion: activeVersion.version, expectedDefinitionDigest: activeVersion.definitionDigest, inputText, conditionValues, model });
      if (requestEpoch.current !== epoch) return;
      setLastRunId(result.record.runId);
      setAdvisoryOutput(result.advisoryOutput);
      setInputText("");
      setNotice(`v5 运行 ${result.record.runId} 已完成；输入已从页面状态清除。`);
      onRunRecorded(result.record.runId);
    });
  }

  async function runOperation(name: string, operation: () => Promise<void>) {
    setPending(name);
    setFailure("");
    setNotice("");
    try {
      await operation();
    } catch (error: unknown) {
      if (error instanceof WorkflowDefinitionPromotionConflict) {
        setFailure(`${error.failureCode}：当前 review v${error.currentReviewVersion} / pointer v${error.currentPointerVersion}，请刷新后重试。`);
      } else {
        setFailure(message(error));
      }
    } finally {
      setPending("");
    }
  }

  if (config.mode === "offline") {
    return <section className="workflow-definition-promotion-panel offline" id="workflow-definition-promotion"><div className="section-heading compact-heading"><div><p className="eyebrow">Workflow Definition · Promotion</p><h4>不可变版本晋级未启用</h4></div><span className="status-badge neutral">offline · zero requests</span></div><p>启用统一本地产品档后才能创建 candidate、人工审查、激活和运行；离线模式不会发起请求。</p></section>;
  }

  return <section className="workflow-definition-promotion-panel" id="workflow-definition-promotion" aria-labelledby="workflow-definition-promotion-title">
    <div className="section-heading compact-heading"><div><p className="eyebrow">Workflow Definition · Controlled runtime</p><h4 id="workflow-definition-promotion-title">不可变版本晋级与精确运行</h4></div><span className={`status-badge ${activation?.state === "active" ? "status-good" : "status-neutral"}`}>{activation?.state ?? "inactive"}</span></div>
    <p className="boundary-note">Saved Draft 保持可编辑；candidate、version、activation 与 v5 run 各自保存精确证据。批准不自动激活，激活不自动执行。</p>
    {failure ? <p className="workflow-definition-failure" role="alert">{failure}</p> : null}
    {notice ? <p className="workflow-definition-notice" aria-live="polite">{notice}</p> : null}
    <div className="workflow-definition-promotion-grid">
      <article>
        <p className="eyebrow">1 · Candidate</p><h5>从已保存草案创建候选</h5>
        <label>Candidate ID<input value={candidateId} onChange={(event) => setCandidateId(event.currentTarget.value)} /></label>
        <label>Definition ID<input value={definitionId} onChange={(event) => setDefinitionId(event.currentTarget.value)} /></label>
        <dl><div><dt>Draft</dt><dd>{activeDraft.draftId} · v{savedDraftVersion}</dd></div><div><dt>Provenance</dt><dd>{activeDraft.workflowDefinitionId || "new lineage"} · base v{activeDraft.baseDefinitionVersion ?? 0}</dd></div></dl>
        <button type="button" disabled={Boolean(pending) || savedDraftVersion < 1} onClick={() => void createCandidate()}>创建晋级候选</button>
        <div className="workflow-definition-list">{candidates.map((candidate) => <button type="button" className={candidate.candidateId === selectedCandidate?.candidateId ? "selected" : ""} key={candidate.candidateId} onClick={() => setSelectedCandidateId(candidate.candidateId)}><strong>{candidate.candidateId}</strong><span>{candidate.state} · review v{candidate.reviewVersion}</span></button>)}</div>
      </article>
      <article>
        <p className="eyebrow">2 · Review</p><h5>人工审查与不可变版本</h5>
        {selectedCandidate ? <><dl><div><dt>Definition</dt><dd>{selectedCandidate.definitionId}</dd></div><div><dt>Digest</dt><dd><code>{shortDigest(selectedCandidate.definitionDigest)}</code></dd></div><div><dt>Eligibility</dt><dd>{selectedCandidate.activationEligible ? "eligible" : selectedCandidate.eligibilityBlockers.join(", ")}</dd></div></dl>
          <label>Decision<select value={reviewDecision} onChange={(event) => setReviewDecision(event.currentTarget.value as "approve" | "reject")}><option value="approve">Approve</option><option value="reject">Reject</option></select></label>
          <label>Reason<textarea value={reason} onChange={(event) => setReason(event.currentTarget.value)} /></label>
          <button type="button" disabled={Boolean(pending) || selectedCandidate.state !== "pending"} onClick={() => void decideCandidate()}>追加 review v{selectedCandidate.reviewVersion + 1}</button>
          <div className="workflow-definition-evidence">{selectedCandidate.reviews.map((review) => <p key={review.reviewVersion}><strong>v{review.reviewVersion} · {review.decision}</strong><span>{review.reason}</span></p>)}</div>
        </> : <p>当前应用暂无晋级候选。</p>}
      </article>
      <article>
        <p className="eyebrow">3 · Activation</p><h5>版本历史与 pointer CAS</h5>
        <div className="workflow-definition-list">{versions.map((version) => <button type="button" className={selectedVersion === version.version ? "selected" : ""} key={version.version} onClick={() => setSelectedVersion(version.version)}><strong>v{version.version}</strong><span>{shortDigest(version.definitionDigest)} · {version.activationEligible ? "eligible" : "blocked"}</span></button>)}</div>
        <label>Decision<select value={activationDecision} onChange={(event) => setActivationDecision(event.currentTarget.value as "activate" | "replace" | "deactivate")}><option value="activate">Activate</option><option value="replace">Replace</option><option value="deactivate">Deactivate</option></select></label>
        <button type="button" disabled={Boolean(pending) || !selectedCandidate || versions.length === 0} onClick={() => void decideActivation()}>{activationDecision} · expected pointer v{activation?.pointerVersion ?? 0}</button>
        <p>Current: {activation?.state ?? "inactive"} · active v{activation?.activeVersion ?? 0} · pointer v{activation?.pointerVersion ?? 0}</p>
        {versions.find((version) => version.version === selectedVersion) ? <button type="button" disabled={Boolean(pending)} onClick={() => onDerivedDraft(deriveWorkflowDraftFromDefinitionVersion(versions.find((version) => version.version === selectedVersion)!, applicationId, nextDerivedDraftNumber))}>从 v{selectedVersion} 派生新草案</button> : null}
      </article>
      <article>
        <p className="eyebrow">4 · Definition-bound run</p><h5>仅从 exact active version 运行</h5>
        <dl><div><dt>Profile</dt><dd>workflow_definition_executor_v1</dd></div><div><dt>Authority</dt><dd>{activeVersion ? `${activeVersion.definitionId} · v${activeVersion.version}` : "no active authority"}</dd></div></dl>
        <label>一次性输入<textarea value={inputText} onChange={(event) => setInputText(event.currentTarget.value)} /></label>
        {activeVersion?.snapshot.nodes.filter((node) => node.nodeType === "condition").map((node) => <label className="workflow-definition-condition" key={node.nodeId}><input type="checkbox" checked={conditionValues[node.nodeId] ?? false} onChange={(event) => setConditionValues((values) => ({ ...values, [node.nodeId]: event.currentTarget.checked }))} />{node.label} · {node.nodeId}</label>)}
        <label>Model（可留空）<input value={model} onChange={(event) => setModel(event.currentTarget.value)} /></label>
        <button type="button" disabled={Boolean(pending) || !activeVersion || !inputText.trim()} onClick={() => void startRun()}>启动精确版本运行</button>
        {lastRunId ? <div className="workflow-definition-run-result"><strong>{lastRunId}</strong><p>{advisoryOutput || "运行已完成；无可展示 advisory output。"}</p><button type="button" onClick={() => onOpenRun?.(lastRunId)}>打开 Run History</button><button type="button" onClick={() => { setAdvisoryOutput(""); setLastRunId(""); }}>清除一次性结果</button></div> : null}
      </article>
    </div>
  </section>;
}

function defaultCandidateId(draftId: string, version: number): string { return `wdrc_${safePart(draftId)}_v${Math.max(1, version)}`.slice(0, 150); }
function defaultDefinitionId(draftId: string): string { return `wdef_${safePart(draftId)}`.slice(0, 150); }
function safePart(value: string): string { return value.toLowerCase().replace(/[^a-z0-9]+/gu, "_").replace(/^_+|_+$/gu, "").slice(0, 64) || "workflow"; }
function shortDigest(value: string): string { return value.length > 20 ? `${value.slice(0, 19)}…` : value; }
function message(error: unknown): string { return error instanceof Error ? error.message : "Workflow definition promotion failed."; }
