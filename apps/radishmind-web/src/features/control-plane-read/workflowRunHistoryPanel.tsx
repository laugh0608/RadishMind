import { lazy, Suspense, useCallback, useEffect, useState } from "react";

import { readWorkflowExecutorConsumerConfig, startWorkflowDiagnosticDevRecord, type WorkflowRunDevFailureScenario } from "./workflowExecutorConsumer.ts";
import type { WorkflowRunRecord } from "./workflowRunRecordConsumer.ts";
import { readWorkflowRAGSnapshotConfig } from "./workflowRAGSnapshotConsumer.ts";
import {
  EMPTY_WORKFLOW_RUN_HISTORY_FILTER,
  initialWorkflowRunHistoryState,
  isWorkflowRunComparisonCompatible,
  isWorkflowRunComparisonEligible,
  listWorkflowRunHistory,
  readWorkflowRunHistoryDetail,
  type WorkflowRunHistoryFilter,
  type WorkflowRunHistorySummary,
} from "./workflowRunHistoryConsumer.ts";

const config = readWorkflowExecutorConsumerConfig();
const ragConfig = readWorkflowRAGSnapshotConfig();
const WorkflowRunComparisonPanel = lazy(() => import("./workflowRunComparisonPanel.tsx"));
const WorkflowEvaluationPanel = lazy(() => import("./workflowEvaluationPanel.tsx"));
const WorkflowEvaluationSuitePanel = lazy(() => import("./workflowEvaluationSuitePanel.tsx"));

export default function WorkflowRunHistoryPanel({
  applicationId,
  refreshKey = 0,
  handoffRunId = "",
  handoffId = "",
  onHandoffConsumed,
}: {
  applicationId: string;
  refreshKey?: number;
  handoffRunId?: string;
  handoffId?: string;
  onHandoffConsumed?: (handoffId: string) => void;
}) {
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
  const [retrievalPreviewPending, setRetrievalPreviewPending] = useState(false);

  const load = useCallback(async (cursor = "", append = false) => {
    if (config.mode !== "dev_workflow_executor_http") return;
    if (!append) {
      setSelectedRunId("");
      setDetail(null);
      setRetrievalPreviewPending(false);
    }
    setHistory((current) => ({ ...current, status: "loading", failureCode: "", failureSummary: "" }));
    try {
      const next = await listWorkflowRunHistory(applicationId, config, filter, cursor, append ? history.runs : []);
      setHistory(next);
    } catch (error) {
      setHistory((current) => ({ ...current, status: "failed", runs: append ? current.runs : [], failureCode: "workflow_run_store_unavailable", failureSummary: error instanceof Error ? error.message : "Workflow run history is unavailable." }));
    }
  }, [applicationId, filter, history.runs]);

  useEffect(() => { void load(); }, [applicationId, refreshKey]); // filters are applied explicitly to avoid request churn while typing

  async function selectRun(run: WorkflowRunHistorySummary) {
    setSelectedRunId(run.runId);
    setRetrievalPreviewPending(false);
    try { setDetail(await readWorkflowRunHistoryDetail(run, applicationId, config)); }
    catch { setDetail(null); }
  }

  useEffect(() => {
    if (!handoffId || !handoffRunId || history.status === "loading") return;
    const run = history.runs.find((item) => item.runId === handoffRunId);
    if (run) {
      void selectRun(run);
    } else if (history.status === "ready" || history.status === "empty") {
      setDiagnosticGenerationState(`Run handoff ${handoffRunId} is outside the current owner page. Refresh or adjust owner filters without replaying the run.`);
    }
    onHandoffConsumed?.(handoffId);
  }, [handoffId, handoffRunId, history.runs, history.status, onHandoffConsumed]);

  async function loadRetrievalPreviews() {
    const run = history.runs.find((item) => item.runId === selectedRunId);
    if (!run || run.schemaVersion !== "workflow_run_record.v3" || !ragConfig.scopes.has("workflow_rag_snapshots:read")) return;
    setRetrievalPreviewPending(true);
    try { setDetail(await readWorkflowRunHistoryDetail(run, applicationId, config, true)); }
    finally { setRetrievalPreviewPending(false); }
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

  const forbiddenWrites = detail ? detail.sideEffects.businessWrites + detail.sideEffects.replayWrites : 0;
  const failedCount = history.runs.filter((run) => run.status === "failed").length;
  const canceledCount = history.runs.filter((run) => run.status === "canceled").length;
  const staleCount = history.runs.filter((run) => run.staleRunning).length;
  const gatewayCount = history.runs.filter((run) => run.failureBoundary === "gateway" || run.failureBoundary === "provider").length;
  const storeCount = history.runs.filter((run) => run.failureBoundary === "run_store").length;
  const comparisonRuns = history.runs.filter(isWorkflowRunComparisonEligible);
  const comparisonBaseline = comparisonRuns.find((run) => run.runId === baselineRunId);
  const comparisonCandidateRuns = comparisonRuns.filter((run) => isWorkflowRunComparisonCompatible(comparisonBaseline, run));
  const hasRetrievalRuns = history.runs.some((run) => run.schemaVersion === "workflow_run_record.v3" || run.schemaVersion === "workflow_run_record.v4");
  const hasDefinitionRuns = history.runs.some((run) => run.schemaVersion === "workflow_run_record.v5");
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
              <label>Status<select value={filter.status} onChange={(event) => setFilter({ ...filter, status: event.target.value as WorkflowRunHistoryFilter["status"] })}><option value="">All</option><option value="succeeded">Succeeded</option><option value="failed">Failed</option><option value="outcome_unknown">Outcome unknown</option><option value="canceled">Canceled</option><option value="running">Running</option></select></label>
              <label>Draft<input value={filter.draftId} onChange={(event) => setFilter({ ...filter, draftId: event.target.value })} placeholder="draft id" /></label>
              <label>Execution source<select value={filter.executionSourceKind} onChange={(event) => setFilter({ ...filter, executionSourceKind: event.target.value as WorkflowRunHistoryFilter["executionSourceKind"] })}><option value="">All</option><option value="workflow_draft">Saved Draft</option><option value="workflow_definition">Workflow Definition</option><option value="application_configuration_draft">Application RAG</option></select></label>
              <label>Source ID<input value={filter.executionSourceId} onChange={(event) => setFilter({ ...filter, executionSourceId: event.target.value })} placeholder="exact source id" /></label>
              <label>Source version<input type="number" min="1" value={filter.executionSourceVersion} onChange={(event) => setFilter({ ...filter, executionSourceVersion: event.target.value ? Number(event.target.value) : "" })} /></label>
              <label>Started from<input type="datetime-local" value={filter.startedFrom} onChange={(event) => setFilter({ ...filter, startedFrom: event.target.value })} /></label>
              <label>Started to<input type="datetime-local" value={filter.startedTo} onChange={(event) => setFilter({ ...filter, startedTo: event.target.value })} /></label>
              <label>Failure code<input value={filter.failureCode} onChange={(event) => setFilter({ ...filter, failureCode: event.target.value })} placeholder="workflow_run_…" /></label>
              <label>Boundary<select value={filter.failureBoundary} onChange={(event) => setFilter({ ...filter, failureBoundary: event.target.value as WorkflowRunHistoryFilter["failureBoundary"] })}><option value="">All</option><option value="executor">Executor</option><option value="gateway">Gateway</option><option value="provider">Provider</option><option value="run_store">Run store</option><option value="request">Request</option><option value="draft_read">Draft read</option><option value="tool_policy">Tool policy</option><option value="tool_confirmation">Tool confirmation</option><option value="tool_transport">Tool transport</option><option value="tool_response">Tool response</option><option value="tool_store">Tool store</option><option value="retrieval_policy">Retrieval policy</option><option value="retrieval_store">Retrieval store</option><option value="retrieval_rank">Retrieval rank</option><option value="retrieval_context">Retrieval context</option><option value="retrieval_citation">Retrieval citation</option><option value="provider_selection">Provider selection</option><option value="provider_call">Provider call</option></select></label>
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
            {history.runs.map((run) => <button type="button" className={`workflow-run-history-live-row ${selectedRunId === run.runId ? "is-selected" : ""}`} key={run.runId} onClick={() => void selectRun(run)}><span className="workflow-run-history-live-identity"><strong>{run.runId}</strong><small>{run.schemaVersion === "workflow_run_record.v4" || run.schemaVersion === "workflow_run_record.v5" ? `${run.executionSourceId} · version ${run.executionSourceVersion}` : `${run.draftId} · version ${run.draftVersion}`}</small></span><span><small>Status</small><strong>{run.status}{run.staleRunning ? " · stale" : ""}</strong><small>{run.schemaVersion}</small></span><span><small>Failure</small><strong>{run.failureBoundary || "none"}</strong><small>{run.retrievalFailureCategory || run.toolFailureCategory || run.gatewayFailureCategory || run.failureCode || "none"}</small></span><span><small>Controlled effects</small><strong>{run.schemaVersion === "workflow_run_record.v3" || run.schemaVersion === "workflow_run_record.v4" ? `${run.sideEffects.retrievalCalls} retrieval · ${run.sideEffects.providerCalls} provider` : run.schemaVersion === "workflow_run_record.v5" ? `${run.sideEffects.providerCalls} provider · 0 other` : `${run.sideEffects.toolCalls} tool · ${run.sideEffects.confirmationCalls} confirmation`}</strong><small>{run.schemaVersion === "workflow_run_record.v3" || run.schemaVersion === "workflow_run_record.v4" ? `${run.citationRefs.length} citation refs` : run.schemaVersion === "workflow_run_record.v5" ? run.executionProfile : run.toolAttemptStatus || "no tool attempt"}</small></span></button>)}
          </div>
          <div className="workflow-run-comparison-selector" aria-label="Workflow run comparison selection">
            <label>Baseline run<select value={baselineRunId} onChange={(event) => { setBaselineRunId(event.target.value); setCandidateRunId(""); setComparisonSelection(null); }}><option value="">Choose baseline</option>{comparisonRuns.map((run) => <option value={run.runId} key={`baseline-${run.runId}`}>{run.runId} · {run.status} · {run.schemaVersion === "workflow_run_record.v5" ? "Definition" : run.schemaVersion === "workflow_run_record.v4" ? "Application RAG" : run.schemaVersion === "workflow_run_record.v3" ? "Workflow RAG" : "standard"}</option>)}</select></label>
            <label>Candidate run<select value={candidateRunId} onChange={(event) => setCandidateRunId(event.target.value)}><option value="">Choose compatible candidate</option>{comparisonCandidateRuns.map((run) => <option value={run.runId} key={`candidate-${run.runId}`}>{run.runId} · {run.status}</option>)}</select></label>
            <button type="button" disabled={!baselineRunId || !candidateRunId || baselineRunId === candidateRunId} onClick={() => setComparisonSelection({ baseline: baselineRunId, candidate: candidateRunId })}>Compare runs</button>
          </div>
          {hasRetrievalRuns ? <p className="boundary-note"><code>workflow_rag_retrieval.v1</code> 与 <code>workflow_rag_application_invocation.v1</code> 分开审查；v4 只与同应用、同 input digest 的 v4 比较，并允许 assignment、binding 与 snapshot 发生受控变化。</p> : null}
          {hasDefinitionRuns ? <p className="boundary-note"><code>workflow_definition_executor.v1</code> 只读比较同一 definition lineage 的 v5 记录，不重新执行、不改变 activation pointer。</p> : null}
          {comparisonSelection ? <Suspense fallback={<p>Loading regression review…</p>}><WorkflowRunComparisonPanel applicationId={applicationId} baselineRunId={comparisonSelection.baseline} candidateRunId={comparisonSelection.candidate} config={config} /></Suspense> : null}
          <Suspense fallback={<p>Loading evaluation cases…</p>}><WorkflowEvaluationPanel applicationId={applicationId} runs={comparisonRuns} config={config} /></Suspense>
          <Suspense fallback={<p>Loading evaluation suites…</p>}><WorkflowEvaluationSuitePanel applicationId={applicationId} config={config} /></Suspense>
          {history.hasMore ? <button type="button" onClick={() => void load(history.nextCursor, true)} disabled={history.status === "loading"}>Load earlier runs</button> : null}
          {detail?.schemaVersion === "workflow_run_record.v3" || detail?.schemaVersion === "workflow_run_record.v4" ? (
            <WorkflowRAGRunHistoryEvidence
              detail={detail}
              canReadPreviews={detail.schemaVersion === "workflow_run_record.v3" && ragConfig.mode === "dev_workflow_rag_http" && ragConfig.scopes.has("workflow_rag_snapshots:read")}
              previewPending={retrievalPreviewPending}
              onLoadPreviews={loadRetrievalPreviews}
            />
          ) : null}
          {detail ? <article className="workflow-run-detail"><div className="card-title-row"><div><p className="eyebrow">Real run detail</p><h4>{detail.runId}</h4></div><span className={`status-badge ${detail.status === "succeeded" ? "status-good" : detail.status === "outcome_unknown" ? "status-neutral" : "status-bad"}`}>{detail.status}</span></div><p>{detail.output || detail.failureSummary || "No advisory output retained in this metadata-only record."}</p><dl className="tenant-meta"><div><dt>Input</dt><dd>{detail.inputBytes} bytes; raw text not retained</dd></div><div><dt>Provider calls</dt><dd>{detail.sideEffects.providerCalls}</dd></div><div><dt>Controlled effects</dt><dd>{detail.schemaVersion === "workflow_run_record.v3" || detail.schemaVersion === "workflow_run_record.v4" ? `${detail.sideEffects.retrievalCalls} retrieval · ${detail.sideEffects.providerCalls} provider` : detail.schemaVersion === "workflow_run_record.v5" ? `${detail.sideEffects.providerCalls} provider · 0 tool/retrieval/write` : `${detail.sideEffects.toolCalls} tool · ${detail.sideEffects.confirmationCalls} confirmation`}</dd></div><div><dt>Forbidden writes</dt><dd>{forbiddenWrites}</dd></div>{detail.schemaVersion === "workflow_run_record.v4" || detail.schemaVersion === "workflow_run_record.v5" ? <div><dt>Execution source</dt><dd>{detail.executionSourceKind} · {detail.executionSourceId} · v{detail.executionSourceVersion}</dd></div> : null}{detail.definitionAuthority ? <><div><dt>Definition authority</dt><dd>{detail.definitionAuthority.definitionId} · v{detail.definitionAuthority.definitionVersion} · pointer v{detail.definitionAuthority.activationPointerVersion}</dd></div><div><dt>Source draft provenance</dt><dd>{detail.definitionAuthority.sourceDraftId} · v{detail.definitionAuthority.sourceDraftVersion}</dd></div></> : null}{detail.planId ? <div><dt>Plan / confirmation</dt><dd>{detail.planId} · {detail.confirmationId}</dd></div> : null}{detail.toolAttempt ? <div><dt>Tool attempt</dt><dd>{detail.toolAttempt.attemptId} · {detail.toolAttempt.status}</dd></div> : null}</dl>{detail.diagnostic ? <div className="workflow-run-diagnostic-review"><p className="eyebrow">Structured failure review</p><h5>{detail.diagnostic.failureBoundary || "No failure"} · {detail.diagnostic.retrievalFailureCategory !== "none" ? detail.diagnostic.retrievalFailureCategory : detail.diagnostic.toolFailureCategory !== "none" ? detail.diagnostic.toolFailureCategory : detail.diagnostic.gatewayFailureCategory}</h5><p>{detail.diagnostic.summary || "The run completed without a structured failure."}</p><dl className="tenant-meta"><div><dt>Failed node</dt><dd>{detail.diagnostic.failedNodeId || "none"}</dd></div><div><dt>Last completed</dt><dd>{detail.diagnostic.lastCompletedNodeId || "none"}</dd></div><div><dt>Review action</dt><dd>{detail.diagnostic.recommendedReviewAction || "none"}</dd></div><div><dt>Terminal write</dt><dd>{detail.diagnostic.terminalWriteState}</dd></div></dl></div> : <p className="boundary-note">Legacy workflow_run_record.v0: structured diagnostic unavailable.</p>}{detail.status === "outcome_unknown" ? <p className="failure-summary">The tool outcome is uncertain. Retry and resume are disabled; review the durable attempt metadata before creating a new plan.</p> : null}<div className="workflow-run-reference-actions"><button type="button" onClick={() => void copyReference("request", detail.requestId)}>Copy request id</button><button type="button" onClick={() => void copyReference("audit", detail.auditRef)}>Copy audit ref</button><span>{copiedRef ? `${copiedRef} copied` : "References are metadata only."}</span></div><div className="workflow-run-history-node-list">{detail.nodes.map((node) => <div className={`workflow-run-history-node-row ${detail.diagnostic?.failedNodeId === node.nodeId ? "is-failed" : ""} ${detail.diagnostic?.lastCompletedNodeId === node.nodeId ? "is-last-completed" : ""}`} key={node.nodeId}><span><strong>{node.label}</strong><small>{node.nodeType}</small></span><span><small>Status</small><strong>{node.status}</strong></span><span><small>Duration</small><strong>{node.durationMs} ms</strong></span><p>{node.outputPreview}</p></div>)}</div><p className="boundary-note">Business writes and replay remain locked at 0. Tool and confirmation counts are allowed only for workflow_run_record.v2; retrieval and provider counts are bounded for v3/v4; v5 keeps all non-provider side effects at 0.</p></article> : null}
        </>
      )}
    </section>
  );
}

function WorkflowRAGRunHistoryEvidence({
  detail,
  canReadPreviews,
  previewPending,
  onLoadPreviews,
}: {
  detail: WorkflowRunRecord;
  canReadPreviews: boolean;
  previewPending: boolean;
  onLoadPreviews: () => Promise<void>;
}) {
  const snapshot = detail.ragSnapshot;
  const attempt = detail.retrievalAttempt;
  const authority = detail.ragApplicationAuthority;
  if (!snapshot || !attempt) return null;
  return (
    <article className="workflow-run-detail workflow-rag-run-history-evidence" aria-label={`Workflow RAG run ${detail.schemaVersion} evidence`}>
      <div className="card-title-row"><div><p className="eyebrow">Metadata-only {detail.schemaVersion}</p><h4>{snapshot.ragRef}</h4></div><span className="status-badge neutral">{attempt.status}</span></div>
      {authority ? <dl className="tenant-meta">
        <div><dt>Runtime assignment</dt><dd>{authority.assignmentId} · v{authority.assignmentVersion}</dd></div>
        <div><dt>Publish candidate</dt><dd>{authority.publishCandidateId} · review v{authority.publishReviewVersion}</dd></div>
        <div><dt>Application draft</dt><dd>{authority.draftId} · v{authority.draftVersion} · {authority.draftDigest}</dd></div>
        <div><dt>Binding</dt><dd>{authority.bindingId} · v{authority.bindingVersion} · {authority.bindingDigest}</dd></div>
        <div><dt>Dataset / review</dt><dd>{authority.datasetId} · v{authority.datasetVersion} · {authority.candidateReviewId}</dd></div>
        <div><dt>Configured route</dt><dd>{authority.configuredProtocol} · {authority.configuredModel}</dd></div>
      </dl> : null}
      <dl className="tenant-meta">
        <div><dt>{authority ? "Execution source" : "Draft"}</dt><dd>{authority ? `${detail.executionSourceId} · v${detail.executionSourceVersion}` : `v${detail.draftVersion} · ${detail.draftDigest}`}</dd></div>
        <div><dt>Snapshot</dt><dd>{snapshot.snapshotId} · v{snapshot.snapshotVersion} · {snapshot.snapshotDigest}</dd></div>
        <div><dt>Profile</dt><dd>{attempt.profileId} · v{attempt.profileVersion} · {attempt.profileDigest}</dd></div>
        <div><dt>Query</dt><dd>{attempt.queryBytes} bytes · {attempt.queryDigest}</dd></div>
        <div><dt>Retrieval</dt><dd>{attempt.candidateCount} candidates · {attempt.selectedFragments.length} selected · {attempt.retrievalLatencyMs} ms</dd></div>
        <div><dt>Context</dt><dd>{attempt.contextBytes} bytes · {attempt.citationRefs.length} citation refs</dd></div>
      </dl>
      <div className="workflow-run-history-node-list">
        {attempt.selectedFragments.map((fragment) => <div className="workflow-run-history-node-row" key={fragment.fragmentRef}><span><strong>#{fragment.rank} · {fragment.fragmentRef}</strong><small>{fragment.sourceType}{fragment.isOfficial ? " · official" : ""}</small></span><span><small>Digest</small><code>{fragment.contentDigest}</code></span><span><small>Cited</small><strong>{attempt.citationRefs.includes(fragment.fragmentRef) ? "yes" : "no"}</strong></span></div>)}
      </div>
      <div className="workflow-run-reference-actions"><button type="button" disabled={!canReadPreviews || previewPending} onClick={() => void onLoadPreviews()}>{previewPending ? "Loading previews…" : "Read authorized fragment previews"}</button><span>{canReadPreviews ? "Preview reads the exact immutable snapshot; maximum 512 characters per fragment." : authority ? "Application invocation history remains metadata-only; fragment preview requests stay disabled." : "workflow_rag_snapshots:read is required; zero preview requests sent."}</span></div>
      {detail.retrievalFragmentPreviews.map((preview) => <blockquote key={preview.fragmentRef}><strong>{preview.fragmentRef}</strong><p>{preview.preview}</p><small>{preview.truncated ? "truncated at 512 characters" : "complete authorized preview"}</small></blockquote>)}
      <p className="boundary-note">原始 query、完整 fragment、prompt packet、credential、模型原始响应和 answer 均不来自 run record。</p>
    </article>
  );
}
