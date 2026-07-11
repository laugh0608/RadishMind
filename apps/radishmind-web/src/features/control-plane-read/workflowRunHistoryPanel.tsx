import { lazy, Suspense, useCallback, useEffect, useState } from "react";

import { readWorkflowExecutorConsumerConfig, startWorkflowDiagnosticDevRecord, type WorkflowRunDevFailureScenario, type WorkflowRunRecord } from "./workflowExecutorConsumer.ts";
import {
  EMPTY_WORKFLOW_RUN_HISTORY_FILTER,
  initialWorkflowRunHistoryState,
  listWorkflowRunHistory,
  readWorkflowRunHistoryDetail,
  type WorkflowRunHistoryFilter,
  type WorkflowRunHistorySummary,
} from "./workflowRunHistoryConsumer.ts";

const config = readWorkflowExecutorConsumerConfig();
const WorkflowRunComparisonPanel = lazy(() => import("./workflowRunComparisonPanel.tsx"));

export default function WorkflowRunHistoryPanel({ applicationId }: { applicationId: string }) {
  const [filter, setFilter] = useState<WorkflowRunHistoryFilter>(EMPTY_WORKFLOW_RUN_HISTORY_FILTER);
  const [history, setHistory] = useState(() => initialWorkflowRunHistoryState(config));
  const [detail, setDetail] = useState<WorkflowRunRecord | null>(null);
  const [selectedRunId, setSelectedRunId] = useState("");
  const [diagnosticScenario, setDiagnosticScenario] = useState<WorkflowRunDevFailureScenario>("gateway_timeout");
  const [diagnosticGenerationState, setDiagnosticGenerationState] = useState("");
  const [copiedRef, setCopiedRef] = useState("");
  const [baselineRunId, setBaselineRunId] = useState("");
  const [candidateRunId, setCandidateRunId] = useState("");
  const [comparisonSelection, setComparisonSelection] = useState<{ baseline: string; candidate: string } | null>(null);

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

  async function generateDiagnosticRun() {
    if (!filter.draftId.trim()) {
      setDiagnosticGenerationState("Enter an exact saved draft id before generating a diagnostic run.");
      return;
    }
    setDiagnosticGenerationState("Generating deterministic dev/test failure…");
    try {
      const state = await startWorkflowDiagnosticDevRecord(filter.draftId, applicationId, diagnosticScenario, config);
      setDetail(state.record);
      setSelectedRunId(state.record?.runId ?? "");
      setDiagnosticGenerationState(state.record ? `${diagnosticScenario} recorded for review.` : `${state.failureCode ?? "workflow_run_unavailable"}: ${state.failureSummary}`);
      await load();
    } catch (error) {
      setDiagnosticGenerationState(error instanceof Error ? error.message : "Diagnostic run generation failed.");
    }
  }

  async function copyReference(label: string, value: string) {
    await navigator.clipboard.writeText(value);
    setCopiedRef(label);
  }

  const forbiddenSideEffects = detail ? detail.sideEffects.toolCalls + detail.sideEffects.confirmationCalls + detail.sideEffects.businessWrites + detail.sideEffects.replayWrites : 0;
  const failedCount = history.runs.filter((run) => run.status === "failed").length;
  const canceledCount = history.runs.filter((run) => run.status === "canceled").length;
  const staleCount = history.runs.filter((run) => run.staleRunning).length;
  const gatewayCount = history.runs.filter((run) => run.failureBoundary === "gateway" || run.failureBoundary === "provider").length;
  const storeCount = history.runs.filter((run) => run.failureBoundary === "run_store").length;
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
              <dl className="tenant-meta"><div><dt>Records</dt><dd>{history.runs.length}</dd></div><div><dt>Failed / canceled</dt><dd>{failedCount} / {canceledCount}</dd></div><div><dt>Stale</dt><dd>{staleCount}</dd></div><div><dt>Gateway / store</dt><dd>{gatewayCount} / {storeCount}</dd></div></dl>
            </article>
            <div className="run-history-metrics" aria-label="Workflow run history filters">
              <label>Status<select value={filter.status} onChange={(event) => setFilter({ ...filter, status: event.target.value as WorkflowRunHistoryFilter["status"] })}><option value="">All</option><option value="succeeded">Succeeded</option><option value="failed">Failed</option><option value="canceled">Canceled</option><option value="running">Running</option></select></label>
              <label>Draft<input value={filter.draftId} onChange={(event) => setFilter({ ...filter, draftId: event.target.value })} placeholder="draft id" /></label>
              <label>Started from<input type="datetime-local" value={filter.startedFrom} onChange={(event) => setFilter({ ...filter, startedFrom: event.target.value })} /></label>
              <label>Started to<input type="datetime-local" value={filter.startedTo} onChange={(event) => setFilter({ ...filter, startedTo: event.target.value })} /></label>
              <label>Failure code<input value={filter.failureCode} onChange={(event) => setFilter({ ...filter, failureCode: event.target.value })} placeholder="workflow_run_…" /></label>
              <label>Boundary<select value={filter.failureBoundary} onChange={(event) => setFilter({ ...filter, failureBoundary: event.target.value as WorkflowRunHistoryFilter["failureBoundary"] })}><option value="">All</option><option value="executor">Executor</option><option value="gateway">Gateway</option><option value="provider">Provider</option><option value="run_store">Run store</option><option value="request">Request</option><option value="draft_read">Draft read</option></select></label>
              <label>Provider<input value={filter.provider} onChange={(event) => setFilter({ ...filter, provider: event.target.value })} placeholder="exact provider" /></label>
              <label>Model<input value={filter.model} onChange={(event) => setFilter({ ...filter, model: event.target.value })} placeholder="exact model" /></label>
              <label>Stale running<select value={filter.staleRunning} onChange={(event) => setFilter({ ...filter, staleRunning: event.target.value as WorkflowRunHistoryFilter["staleRunning"] })}><option value="">All</option><option value="true">Only stale</option><option value="false">Exclude stale</option></select></label>
              <button type="button" onClick={() => void load()} disabled={history.status === "loading"}>Apply filters</button>
            </div>
          </div>
          {config.diagnosticsDevEnabled ? <div className="workflow-run-diagnostic-generator">
            <label>Dev/test failure scenario<select value={diagnosticScenario} onChange={(event) => setDiagnosticScenario(event.target.value as WorkflowRunDevFailureScenario)}><option value="gateway_timeout">Gateway timeout</option><option value="gateway_queue_full">Gateway queue full</option><option value="gateway_worker_crash">Gateway worker crash</option><option value="gateway_protocol_failure">Gateway protocol</option><option value="provider_failed">Provider failed</option><option value="output_unavailable">Output unavailable</option><option value="request_canceled">Request canceled</option><option value="run_store_unavailable">Store unavailable</option><option value="terminal_write_conflict">Terminal write conflict</option><option value="budget_exceeded">Budget exceeded</option><option value="stale_running">Stale running</option></select></label>
            <button type="button" onClick={() => void generateDiagnosticRun()}>Generate diagnostic run</button>
            <p>{diagnosticGenerationState || "Uses the exact Draft filter and the server-side diagnostics gate; no arbitrary failure payload is accepted."}</p>
          </div> : null}
          {history.failureCode ? <p className="failure-summary">{history.failureCode}: {history.failureSummary}</p> : null}
          <div className="workflow-run-history-live-list" aria-label="Real workflow run records">
            {history.runs.map((run) => <button type="button" className={`workflow-run-history-live-row ${selectedRunId === run.runId ? "is-selected" : ""}`} key={run.runId} onClick={() => void selectRun(run)}><span className="workflow-run-history-live-identity"><strong>{run.runId}</strong><small>{run.draftId} · version {run.draftVersion}</small></span><span><small>Status</small><strong>{run.status}{run.staleRunning ? " · stale" : ""}</strong></span><span><small>Failure</small><strong>{run.failureBoundary || "none"}</strong><small>{run.gatewayFailureCategory || run.failureCode || "none"}</small></span><span><small>Provider</small><strong>{run.selectedProvider || "unavailable"}</strong></span></button>)}
          </div>
          <div className="workflow-run-comparison-selector" aria-label="Workflow run comparison selection">
            <label>Baseline run<select value={baselineRunId} onChange={(event) => setBaselineRunId(event.target.value)}><option value="">Choose baseline</option>{history.runs.map((run) => <option value={run.runId} key={`baseline-${run.runId}`}>{run.runId} · {run.status}</option>)}</select></label>
            <label>Candidate run<select value={candidateRunId} onChange={(event) => setCandidateRunId(event.target.value)}><option value="">Choose candidate</option>{history.runs.map((run) => <option value={run.runId} key={`candidate-${run.runId}`}>{run.runId} · {run.status}</option>)}</select></label>
            <button type="button" disabled={!baselineRunId || !candidateRunId || baselineRunId === candidateRunId} onClick={() => setComparisonSelection({ baseline: baselineRunId, candidate: candidateRunId })}>Compare runs</button>
          </div>
          {comparisonSelection ? <Suspense fallback={<p>Loading regression review…</p>}><WorkflowRunComparisonPanel applicationId={applicationId} baselineRunId={comparisonSelection.baseline} candidateRunId={comparisonSelection.candidate} config={config} /></Suspense> : null}
          {history.hasMore ? <button type="button" onClick={() => void load(history.nextCursor, true)} disabled={history.status === "loading"}>Load earlier runs</button> : null}
          {detail ? <article className="workflow-run-detail"><div className="card-title-row"><div><p className="eyebrow">Real run detail</p><h4>{detail.runId}</h4></div><span className="status-badge status-good">{detail.status}</span></div><p>{detail.output || detail.failureSummary || "No advisory output recorded."}</p><dl className="tenant-meta"><div><dt>Input</dt><dd>{detail.inputBytes} bytes; raw text not retained</dd></div><div><dt>Conditions</dt><dd>{detail.conditionNodeIds.join(", ") || "none"}; values not retained</dd></div><div><dt>Provider calls</dt><dd>{detail.sideEffects.providerCalls}</dd></div><div><dt>Forbidden side effects</dt><dd>{forbiddenSideEffects}</dd></div></dl>{detail.diagnostic ? <div className="workflow-run-diagnostic-review"><p className="eyebrow">Structured failure review</p><h5>{detail.diagnostic.failureBoundary || "No failure"} · {detail.diagnostic.gatewayFailureCategory}</h5><p>{detail.diagnostic.summary || "The run completed without a structured failure."}</p><dl className="tenant-meta"><div><dt>Failed node</dt><dd>{detail.diagnostic.failedNodeId || "none"}</dd></div><div><dt>Last completed</dt><dd>{detail.diagnostic.lastCompletedNodeId || "none"}</dd></div><div><dt>Review action</dt><dd>{detail.diagnostic.recommendedReviewAction || "none"}</dd></div><div><dt>Terminal write</dt><dd>{detail.diagnostic.terminalWriteState}</dd></div></dl></div> : <p className="boundary-note">Legacy workflow_run_record.v0: structured diagnostic unavailable.</p>}<div className="workflow-run-reference-actions"><button type="button" onClick={() => void copyReference("request", detail.requestId)}>Copy request id</button><button type="button" onClick={() => void copyReference("audit", detail.auditRef)}>Copy audit ref</button><span>{copiedRef ? `${copiedRef} copied` : "References are metadata only."}</span></div><div className="workflow-run-history-node-list">{detail.nodes.map((node) => <div className={`workflow-run-history-node-row ${detail.diagnostic?.failedNodeId === node.nodeId ? "is-failed" : ""} ${detail.diagnostic?.lastCompletedNodeId === node.nodeId ? "is-last-completed" : ""}`} key={node.nodeId}><span><strong>{node.label}</strong><small>{node.nodeType}</small></span><span><small>Status</small><strong>{node.status}</strong></span><span><small>Duration</small><strong>{node.durationMs} ms</strong></span><p>{node.outputPreview}</p></div>)}</div><p className="boundary-note">tool / confirmation / business write / replay remain locked at 0. Replay and resume are unavailable.</p></article> : null}
        </>
      )}
    </section>
  );
}
