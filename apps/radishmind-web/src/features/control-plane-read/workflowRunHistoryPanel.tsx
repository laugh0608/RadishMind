import { useCallback, useEffect, useState } from "react";

import { readWorkflowExecutorConsumerConfig, type WorkflowRunRecord } from "./workflowExecutorConsumer.ts";
import {
  EMPTY_WORKFLOW_RUN_HISTORY_FILTER,
  initialWorkflowRunHistoryState,
  listWorkflowRunHistory,
  readWorkflowRunHistoryDetail,
  type WorkflowRunHistoryFilter,
  type WorkflowRunHistorySummary,
} from "./workflowRunHistoryConsumer.ts";

const config = readWorkflowExecutorConsumerConfig();

export default function WorkflowRunHistoryPanel({ applicationId }: { applicationId: string }) {
  const [filter, setFilter] = useState<WorkflowRunHistoryFilter>(EMPTY_WORKFLOW_RUN_HISTORY_FILTER);
  const [history, setHistory] = useState(() => initialWorkflowRunHistoryState(config));
  const [detail, setDetail] = useState<WorkflowRunRecord | null>(null);
  const [selectedRunId, setSelectedRunId] = useState("");

  const load = useCallback(async (cursor = "", append = false) => {
    if (config.mode !== "dev_workflow_executor_http") return;
    setHistory((current) => ({ ...current, status: "loading", failureCode: "", failureSummary: "" }));
    try {
      const next = await listWorkflowRunHistory(applicationId, config, filter, cursor, append ? history.runs : []);
      setHistory(next);
    } catch (error) {
      setHistory((current) => ({ ...current, status: "failed", runs: append ? current.runs : [], failureCode: "workflow_run_store_unavailable", failureSummary: error instanceof Error ? error.message : "Workflow run history is unavailable." }));
    }
  }, [applicationId, filter, history.runs]);

  useEffect(() => { void load(); }, [applicationId]); // filters are applied explicitly to avoid request churn while typing

  async function selectRun(run: WorkflowRunHistorySummary) {
    setSelectedRunId(run.runId);
    try { setDetail(await readWorkflowRunHistoryDetail(run, applicationId, config)); }
    catch { setDetail(null); }
  }

  const forbiddenSideEffects = detail ? detail.sideEffects.toolCalls + detail.sideEffects.confirmationCalls + detail.sideEffects.businessWrites + detail.sideEffects.replayWrites : 0;
  return (
    <section className="surface-band workspace-run-history" id="workspace-run-history" aria-labelledby="workspace-run-history-title">
      <div className="section-heading">
        <div><p className="eyebrow">User Workspace</p><h3 id="workspace-run-history-title">Run history</h3></div>
        <span className={`status-badge ${config.mode === "dev_workflow_executor_http" ? "status-good" : "status-neutral"}`}>{config.mode === "dev_workflow_executor_http" ? "durable dev/test" : "offline sample"}</span>
      </div>
      {config.mode !== "dev_workflow_executor_http" ? (
        <article className="run-history-route"><p className="eyebrow">Offline mode</p><h4>No live run request</h4><p>Default offline mode keeps the historical sample surface visibly separate. Enable the explicit executor dev source to read `/v1/user-workspace/workflow-runs`.</p></article>
      ) : (
        <>
          <div className="run-history-summary">
            <article className="run-history-route">
              <p className="eyebrow">Real executor history</p><h4>/v1/user-workspace/workflow-runs</h4>
              <p className="route-path">{applicationId} · {history.status}</p>
              <dl className="tenant-meta"><div><dt>Request</dt><dd>{history.requestId}</dd></div><div><dt>Audit</dt><dd>{history.auditRef}</dd></div><div><dt>Cursor</dt><dd>{history.hasMore ? "available" : "none"}</dd></div><div><dt>Records</dt><dd>{history.runs.length}</dd></div></dl>
            </article>
            <div className="run-history-metrics" aria-label="Workflow run history filters">
              <label>Status<select value={filter.status} onChange={(event) => setFilter({ ...filter, status: event.target.value as WorkflowRunHistoryFilter["status"] })}><option value="">All</option><option value="succeeded">Succeeded</option><option value="failed">Failed</option><option value="canceled">Canceled</option><option value="running">Running</option></select></label>
              <label>Draft<input value={filter.draftId} onChange={(event) => setFilter({ ...filter, draftId: event.target.value })} placeholder="draft id" /></label>
              <label>Started from<input type="datetime-local" value={filter.startedFrom} onChange={(event) => setFilter({ ...filter, startedFrom: event.target.value })} /></label>
              <label>Started to<input type="datetime-local" value={filter.startedTo} onChange={(event) => setFilter({ ...filter, startedTo: event.target.value })} /></label>
              <button type="button" onClick={() => void load()} disabled={history.status === "loading"}>Apply filters</button>
            </div>
          </div>
          {history.failureCode ? <p className="failure-summary">{history.failureCode}: {history.failureSummary}</p> : null}
          <div className="workflow-run-history-live-list" aria-label="Real workflow run records">
            {history.runs.map((run) => <button type="button" className={`workflow-run-history-live-row ${selectedRunId === run.runId ? "is-selected" : ""}`} key={run.runId} onClick={() => void selectRun(run)}><span className="workflow-run-history-live-identity"><strong>{run.runId}</strong><small>{run.draftId} · version {run.draftVersion}</small></span><span><small>Status</small><strong>{run.status}{run.staleRunning ? " · stale" : ""}</strong></span><span><small>Started</small><strong>{run.startedAt}</strong></span><span><small>Provider</small><strong>{run.selectedProvider || "unavailable"}</strong></span></button>)}
          </div>
          {history.hasMore ? <button type="button" onClick={() => void load(history.nextCursor, true)} disabled={history.status === "loading"}>Load earlier runs</button> : null}
          {detail ? <article className="workflow-run-detail"><div className="card-title-row"><div><p className="eyebrow">Real run detail</p><h4>{detail.runId}</h4></div><span className="status-badge status-good">{detail.status}</span></div><p>{detail.output || detail.failureSummary || "No advisory output recorded."}</p><dl className="tenant-meta"><div><dt>Input</dt><dd>{detail.inputBytes} bytes; raw text not retained</dd></div><div><dt>Conditions</dt><dd>{detail.conditionNodeIds.join(", ") || "none"}; values not retained</dd></div><div><dt>Provider calls</dt><dd>{detail.sideEffects.providerCalls}</dd></div><div><dt>Forbidden side effects</dt><dd>{forbiddenSideEffects}</dd></div></dl><div className="workflow-run-history-node-list">{detail.nodes.map((node) => <div className="workflow-run-history-node-row" key={node.nodeId}><span><strong>{node.label}</strong><small>{node.nodeType}</small></span><span><small>Status</small><strong>{node.status}</strong></span><span><small>Duration</small><strong>{node.durationMs} ms</strong></span><p>{node.outputPreview}</p></div>)}</div><p className="boundary-note">tool / confirmation / business write / replay remain locked at 0. Replay and resume are unavailable.</p></article> : null}
        </>
      )}
    </section>
  );
}
