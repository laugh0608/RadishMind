import { useTranslation } from "react-i18next";
import "../../i18n/workflowRAGExecutionResources.ts";
import "../../i18n/workflowDraftResources.ts";
import { workflowDraftStatusLabel } from "./workflowDraftMessages.ts";
import {
  workflowRAGConfidenceLabel,
  workflowRAGExecutionFailure,
  workflowRAGExecutionFeedback,
  workflowRAGExecutionReasonMessage,
  workflowRAGExecutionStatus,
} from "./workflowRAGExecutionMessages.ts";
import { useEffect, useMemo, useState } from "react";

import type { WorkflowSavedDraftConsumerState } from "./savedWorkflowDraftConsumer.ts";
import type { WorkflowDraftDesignerDraft } from "./workflowDraftDesigner.ts";
import {
  buildWorkflowRAGRetrievalDraft,
  evaluateWorkflowRAGExecutionEligibility,
  executeWorkflowRAGRetrieval,
  initialWorkflowRAGExecutionState,
  type WorkflowRAGExecutionEligibility,
  type WorkflowRAGExecutionState,
} from "./workflowRAGExecutionConsumer.ts";
import {
  listWorkflowRAGSnapshots,
  readWorkflowRAGSnapshot,
  readWorkflowRAGSnapshotConfig,
  type WorkflowRAGSnapshotOperationResult,
  type WorkflowRAGSnapshotResource,
} from "./workflowRAGSnapshotConsumer.ts";

const config = readWorkflowRAGSnapshotConfig();
const DEFAULT_WORKFLOW_RAG_INPUT = "根据当前应用知识快照，说明该工作流的主要使用边界，并引用实际证据。";

export default function WorkflowRAGExecutionPanel({
  applicationRef,
  draft,
  savedDraftState,
  draftEditDirty,
  nextDraftNumber,
  onCreateDraft,
  onBindRAGRef,
  onPendingChange,
  onExecutionRecorded,
}: {
  applicationRef: string;
  draft: WorkflowDraftDesignerDraft;
  savedDraftState: WorkflowSavedDraftConsumerState;
  draftEditDirty: boolean;
  nextDraftNumber: number;
  onCreateDraft: (draft: WorkflowDraftDesignerDraft) => void;
  onBindRAGRef: (nodeId: string, ragRef: string) => void;
  onPendingChange: (pending: boolean) => void;
  onExecutionRecorded: () => void;
}) {
  const { t } = useTranslation("workflow");
  const [executionState, setExecutionState] = useState<WorkflowRAGExecutionState>(() => initialWorkflowRAGExecutionState(config));
  const [inputText, setInputText] = useState(DEFAULT_WORKFLOW_RAG_INPUT);
  const [model, setModel] = useState("");
  const [temperature, setTemperature] = useState("");
  const [snapshots, setSnapshots] = useState<WorkflowRAGSnapshotResource[]>([]);
  const [snapshotStatus, setSnapshotStatus] = useState("idle");
  const [snapshotFailure, setSnapshotFailure] = useState("");
  const [selectedSnapshotId, setSelectedSnapshotId] = useState("");
  const [selectedVersion, setSelectedVersion] = useState(0);
  const [exactSnapshot, setExactSnapshot] = useState<WorkflowRAGSnapshotOperationResult | null>(null);
  const retrievalNode = useMemo(() => draft.nodes.find((node) => node.nodeType === "rag_retrieval") ?? null, [draft.nodes]);
  const selectedResource = snapshots.find((snapshot) => snapshot.snapshotId === selectedSnapshotId) ?? null;
  const executionPending = executionState.status === "executing";
  const eligibility: WorkflowRAGExecutionEligibility = useMemo(
    () => evaluateWorkflowRAGExecutionEligibility(draft, savedDraftState, draftEditDirty, config),
    [draft, savedDraftState, draftEditDirty],
  );
  const canReadSnapshots = config.mode === "dev_workflow_rag_http" && config.scopes.has("workflow_rag_snapshots:read");

  const loadSnapshots = async () => {
    if (!canReadSnapshots || !applicationRef.trim()) return;
    setSnapshotStatus("loading");
    setSnapshotFailure("");
    const result = await listWorkflowRAGSnapshots(config, applicationRef, "active");
    setSnapshotStatus(result.status);
    setSnapshotFailure(result.failureCode);
    setSnapshots(result.records);
    const current = parseRAGRef(retrievalNode?.ragRef ?? "");
    const matching = result.records.find((snapshot) => snapshot.snapshotKey === current?.snapshotKey);
    if (matching && current) {
      setSelectedSnapshotId(matching.snapshotId);
      setSelectedVersion(current.version);
      const exact = await readWorkflowRAGSnapshot(config, applicationRef, matching.snapshotId, current.version);
      setExactSnapshot(exact.record?.ragRef === retrievalNode?.ragRef ? exact : null);
      if (exact.failureCode) setSnapshotFailure(exact.failureCode);
      return;
    }
    setSelectedSnapshotId("");
    setSelectedVersion(0);
    setExactSnapshot(null);
  };

  useEffect(() => {
    setSnapshots([]);
    setSnapshotStatus("idle");
    setSnapshotFailure("");
    setExactSnapshot(null);
    void loadSnapshots();
  }, [applicationRef, draft.draftId, canReadSnapshots]);

  useEffect(() => {
    setExecutionState(initialWorkflowRAGExecutionState(config));
    setInputText(DEFAULT_WORKFLOW_RAG_INPUT);
    setModel("");
    setTemperature("");
  }, [draft.draftId]);

  useEffect(() => {
    onPendingChange(executionPending);
    return () => onPendingChange(false);
  }, [executionPending, onPendingChange]);

  async function bindExactVersion(snapshotId: string, version: number) {
    const resource = snapshots.find((snapshot) => snapshot.snapshotId === snapshotId);
    if (!resource || version < 1 || version > resource.latestVersion || !retrievalNode) return;
    setSelectedSnapshotId(snapshotId);
    setSelectedVersion(version);
    setSnapshotStatus("reading");
    const result = await readWorkflowRAGSnapshot(config, applicationRef, snapshotId, version);
    setSnapshotStatus(result.status);
    setSnapshotFailure(result.failureCode);
    setExactSnapshot(result);
    if (result.record?.lifecycleState === "active" && result.record.ragRef === `workflow.rag.${resource.snapshotKey}.v${version}`) {
      onBindRAGRef(retrievalNode.nodeId, result.record.ragRef);
    }
  }

  const inputBytes = new TextEncoder().encode(inputText.trim()).length;
  const executeDisabled = !eligibility.eligible || executionPending || inputBytes < 1 || inputBytes > 4096;
  const bindingMatchesDraft = Boolean(exactSnapshot?.record && exactSnapshot.record.ragRef === retrievalNode?.ragRef);
  const createDraft = () => {
    if (executionPending) return;
    onCreateDraft(buildWorkflowRAGRetrievalDraft(draft, nextDraftNumber, applicationRef));
  };
  const execute = async () => {
    if (!eligibility.eligible || executionPending) return;
    const parsedTemperature = temperature.trim() === "" ? null : Number(temperature);
    setExecutionState((state) => ({
      ...state,
      status: "executing",
      message: "executing",
      failureCode: "",
      record: null,
      answer: null,
    }));
    const state = await executeWorkflowRAGRetrieval(config, draft, eligibility, {
      inputText,
      model,
      temperature: parsedTemperature,
    });
    setExecutionState(state);
    if (state.record) onExecutionRecorded();
  };
  return (
    <section className="workflow-rag-execution-panel" id="workflow-rag-execution" aria-labelledby="workflow-rag-execution-title">
      <div className="section-heading compact-heading">
        <div><p className="eyebrow">{t($ => $.ragExecution.eyebrow)}</p><h4 id="workflow-rag-execution-title">{t($ => $.ragExecution.heading)}</h4></div>
        <span className={`status-badge ${executionState.status === "succeeded" ? "good" : executionState.status === "failed" ? "bad" : "neutral"}`}>{workflowRAGExecutionStatus(t, executionState.status)}</span>
      </div>

      <div className="workflow-rag-scope-grid">
        <article><span>{t($ => $.ragExecution.draftProfile)}</span><strong>{draft.executionProfile ?? "review_only"}</strong><small>{draft.draftId}</small></article>
        <article><span>{t($ => $.ragExecution.savedVersion)}</span><strong>{savedDraftState.currentDraftVersion || t($ => $.ragExecution.notSaved)}</strong><small>{draftEditDirty ? t($ => $.ragExecution.unsaved) : workflowDraftStatusLabel(t, savedDraftState.status)}</small></article>
        <article><span>{t($ => $.ragExecution.exactRef)}</span><strong>{retrievalNode?.ragRef || t($ => $.ragExecution.notBound)}</strong><small>{bindingMatchesDraft ? exactSnapshot?.record?.snapshotDigest : t($ => $.ragExecution.selectRead)}</small></article>
        <article><span>{t($ => $.ragExecution.boundary)}</span><strong>{t($ => $.ragExecution.calls)}</strong><small>{t($ => $.ragExecution.sideEffects)}</small></article>
      </div>

      <div className="workflow-rag-toolbar">
        <button type="button" disabled={executionPending} onClick={createDraft}>{t($ => $.ragExecution.createDraft)}</button>
        <button type="button" disabled={!canReadSnapshots || snapshotStatus === "loading" || executionPending} onClick={() => void loadSnapshots()}>{t($ => $.ragExecution.refresh)}</button>
        <span role="status">{workflowRAGExecutionStatus(t, snapshotStatus)}</span>
      </div>

      {snapshotStatus === "empty" ? <p>{t($ => $.ragExecution.emptySnapshots)}</p> : null}
      {snapshotFailure ? <p className="failure-summary"><code>{snapshotFailure}</code> · {workflowRAGExecutionFailure(t, snapshotFailure)}</p> : null}
      <div className="workflow-rag-binding-grid" aria-label={t($ => $.ragExecution.bindingLabel)}>
        <label>
          <span>{t($ => $.ragExecution.snapshot)}</span>
          <select value={selectedSnapshotId} disabled={!retrievalNode || executionPending || snapshotStatus === "loading"} onChange={(event) => {
            const resource = snapshots.find((snapshot) => snapshot.snapshotId === event.currentTarget.value);
            if (resource) void bindExactVersion(resource.snapshotId, resource.latestVersion);
          }}>
            <option value="">{t($ => $.ragExecution.selectSnapshot)}</option>
            {snapshots.map((snapshot) => <option value={snapshot.snapshotId} key={snapshot.snapshotId}>{t($ => $.ragExecution.snapshotOption, { name: snapshot.displayName, key: snapshot.snapshotKey, version: snapshot.latestVersion })}</option>)}
          </select>
        </label>
        <label>
          <span>{t($ => $.ragExecution.version)}</span>
          <select value={selectedVersion || ""} disabled={!selectedResource || executionPending || snapshotStatus === "reading"} onChange={(event) => void bindExactVersion(selectedSnapshotId, Number(event.currentTarget.value))}>
            <option value="">{t($ => $.ragExecution.selectVersion)}</option>
            {selectedResource ? Array.from({ length: selectedResource.latestVersion }, (_, index) => index + 1).map((version) => <option value={version} key={version}>v{version}</option>) : null}
          </select>
        </label>
      </div>
      <p className="boundary-note">{t($ => $.ragExecution.bindingNote)}</p>

      <div className="workflow-rag-execution-form">
        <label><span>{t($ => $.ragExecution.question)}</span><textarea rows={4} maxLength={4096} value={inputText} disabled={executionPending} onChange={(event) => setInputText(event.currentTarget.value)} /></label>
        <div className="workflow-rag-binding-grid">
          <label><span>{t($ => $.ragExecution.model)}</span><input value={model} maxLength={256} disabled={executionPending} onChange={(event) => setModel(event.currentTarget.value)} placeholder={t($ => $.ragExecution.defaultValue)} /></label>
          <label><span>{t($ => $.ragExecution.temperature)}</span><input type="number" min="0" max="2" step="0.1" value={temperature} disabled={executionPending} onChange={(event) => setTemperature(event.currentTarget.value)} placeholder={t($ => $.ragExecution.defaultValue)} /></label>
        </div>
        <div className="workflow-rag-actions"><button type="button" disabled={executeDisabled} onClick={() => void execute()}>{executionPending ? t($ => $.ragExecution.executing) : t($ => $.ragExecution.execute)}</button><span>{t($ => $.ragExecution.bytes, { count: inputBytes })}</span></div>
      </div>

      {!eligibility.eligible ? <div className="workflow-rag-eligibility" aria-label={t($ => $.ragExecution.blockers)}>{eligibility.reasons.map((item) => <p key={item.code === "rag_execution_scope_denied" ? `${item.code}-${item.scope}` : item.code}><code>{item.code}</code> · {workflowRAGExecutionReasonMessage(t, item)}</p>)}</div> : null}
      <p role="status">{workflowRAGExecutionFeedback(t, executionState.message)}</p>
      {executionState.failureCode ? <p className="failure-summary"><code>{executionState.failureCode}</code> · {workflowRAGExecutionFailure(t, executionState.failureCode)}</p> : null}
      {executionState.message === "failed" || executionState.message === "succeeded" ? <p className="boundary-note">{t($ => $.ragExecution.request)}: <code>{executionState.requestId}</code> · {t($ => $.ragExecution.audit)}: <code>{executionState.auditRef}</code></p> : null}
      {executionState.answer ? <article className="workflow-rag-answer"><div className="card-title-row"><div><p className="eyebrow">{t($ => $.ragExecution.transientAnswer)}</p><h5>{t($ => $.ragExecution.confidence, { level: workflowRAGConfidenceLabel(t, executionState.answer.confidence) })}</h5></div><span className="status-badge good">{t($ => $.ragExecution.validated)}</span></div><p>{executionState.answer.answer}</p><ul>{executionState.answer.citations.map((citation) => <li key={citation.fragmentRef}><code>{citation.fragmentRef}</code> · {citation.claimSummary}</li>)}</ul>{executionState.answer.limitations.length ? <div>{executionState.answer.limitations.map((limitation) => <p key={limitation}>{t($ => $.ragExecution.limitation, { text: limitation })}</p>)}</div> : null}<small>{t($ => $.ragExecution.answerNote)}</small></article> : null}
    </section>
  );
}

function parseRAGRef(value: string): { snapshotKey: string; version: number } | null {
  const match = /^workflow\.rag\.([a-z][a-z0-9_]{2,47})\.v([1-9][0-9]*)$/u.exec(value.trim());
  return match ? { snapshotKey: match[1]!, version: Number(match[2]) } : null;
}
