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
      summary: `Executing exact saved draft ${draft.draftId} version ${eligibility.draftVersion}.`,
      failureCode: "",
      failureSummary: "",
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
        <div><p className="eyebrow">Workflow RAG · Draft binding and execution</p><h4 id="workflow-rag-execution-title">精确知识版本与显式执行</h4></div>
        <span className={`status-badge ${executionState.status === "succeeded" ? "good" : executionState.status === "failed" ? "bad" : "neutral"}`}>{executionState.status}</span>
      </div>

      <div className="workflow-rag-scope-grid">
        <article><span>Draft profile</span><strong>{draft.executionProfile ?? "review_only"}</strong><small>{draft.draftId}</small></article>
        <article><span>Saved version</span><strong>{savedDraftState.currentDraftVersion || "not saved"}</strong><small>{draftEditDirty ? "unsaved local changes" : savedDraftState.status}</small></article>
        <article><span>Exact RAG ref</span><strong>{retrievalNode?.ragRef || "not bound"}</strong><small>{bindingMatchesDraft ? exactSnapshot?.record?.snapshotDigest : "select and read an exact active version"}</small></article>
        <article><span>Execution boundary</span><strong>1 retrieval · 1 provider</strong><small>0 tool · 0 confirmation · 0 write · 0 replay</small></article>
      </div>

      <div className="workflow-rag-toolbar">
        <button type="button" disabled={executionPending} onClick={createDraft}>创建 RAG retrieval v1 草案</button>
        <button type="button" disabled={!canReadSnapshots || snapshotStatus === "loading" || executionPending} onClick={() => void loadSnapshots()}>刷新 active snapshots</button>
        <span>{snapshotStatus}{snapshotFailure ? ` · ${snapshotFailure}` : ""}</span>
      </div>

      <div className="workflow-rag-binding-grid" aria-label="Draft Designer exact RAG snapshot binding">
        <label>
          <span>Active snapshot</span>
          <select value={selectedSnapshotId} disabled={!retrievalNode || executionPending || snapshotStatus === "loading"} onChange={(event) => {
            const resource = snapshots.find((snapshot) => snapshot.snapshotId === event.currentTarget.value);
            if (resource) void bindExactVersion(resource.snapshotId, resource.latestVersion);
          }}>
            <option value="">选择 active snapshot</option>
            {snapshots.map((snapshot) => <option value={snapshot.snapshotId} key={snapshot.snapshotId}>{snapshot.displayName} · {snapshot.snapshotKey} · latest v{snapshot.latestVersion}</option>)}
          </select>
        </label>
        <label>
          <span>Exact version</span>
          <select value={selectedVersion || ""} disabled={!selectedResource || executionPending || snapshotStatus === "reading"} onChange={(event) => void bindExactVersion(selectedSnapshotId, Number(event.currentTarget.value))}>
            <option value="">选择精确版本</option>
            {selectedResource ? Array.from({ length: selectedResource.latestVersion }, (_, index) => index + 1).map((version) => <option value={version} key={version}>v{version}</option>) : null}
          </select>
        </label>
      </div>
      <p className="boundary-note">该选择器只写入精确 `rag_ref`；fragment、排名、citation 白名单、snapshot digest 与 profile digest 均由服务端重读。</p>

      <div className="workflow-rag-execution-form">
        <label><span>Question</span><textarea rows={4} maxLength={4096} value={inputText} disabled={executionPending} onChange={(event) => setInputText(event.currentTarget.value)} /></label>
        <div className="workflow-rag-binding-grid">
          <label><span>Model override</span><input value={model} maxLength={256} disabled={executionPending} onChange={(event) => setModel(event.currentTarget.value)} placeholder="configured default" /></label>
          <label><span>Temperature</span><input type="number" min="0" max="2" step="0.1" value={temperature} disabled={executionPending} onChange={(event) => setTemperature(event.currentTarget.value)} placeholder="configured default" /></label>
        </div>
        <div className="workflow-rag-actions"><button type="button" disabled={executeDisabled} onClick={() => void execute()}>{executionPending ? "执行中…" : "启动一次 retrieval execution"}</button><span>{inputBytes} / 4096 bytes</span></div>
      </div>

      {!eligibility.eligible ? <div className="workflow-rag-eligibility" aria-label="RAG execution eligibility blockers">{eligibility.reasons.map((item) => <p key={`${item.code}-${item.summary}`}><code>{item.code}</code> · {item.summary}</p>)}</div> : null}
      {executionState.failureCode ? <p className="failure-summary"><code>{executionState.failureCode}</code> · {executionState.failureSummary || executionState.summary}</p> : null}
      {executionState.answer ? <article className="workflow-rag-answer"><div className="card-title-row"><div><p className="eyebrow">Transient workflow_rag_answer.v1</p><h5>{executionState.answer.confidence} confidence</h5></div><span className="status-badge good">validated</span></div><p>{executionState.answer.answer}</p><ul>{executionState.answer.citations.map((citation) => <li key={citation.fragmentRef}><code>{citation.fragmentRef}</code> · {citation.claimSummary}</li>)}</ul>{executionState.answer.limitations.length ? <div>{executionState.answer.limitations.map((limitation) => <p key={limitation}>限制：{limitation}</p>)}</div> : null}<small>回答只保留在当前 execution 响应状态；Run History 仅保存 citation refs 与检索元数据。</small></article> : null}
    </section>
  );
}

function parseRAGRef(value: string): { snapshotKey: string; version: number } | null {
  const match = /^workflow\.rag\.([a-z][a-z0-9_]{2,47})\.v([1-9][0-9]*)$/u.exec(value.trim());
  return match ? { snapshotKey: match[1]!, version: Number(match[2]) } : null;
}
