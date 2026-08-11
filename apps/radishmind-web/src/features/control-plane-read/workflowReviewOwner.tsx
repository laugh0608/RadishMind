import { lazy, Suspense, useCallback, useEffect, useRef, useState } from "react";

import type { WorkflowReviewSurface } from "./applicationDevelopmentWorkspace.ts";
import { startWorkflowDiagnosticDevRecord, type WorkflowRunDevFailureScenario } from "./workflowExecutorConsumer.ts";
import { readWorkflowRAGSnapshotConfig } from "./workflowRAGSnapshotConsumer.ts";
import type { WorkflowRunRecord } from "./workflowRunRecordConsumer.ts";
import {
  EMPTY_WORKFLOW_RUN_HISTORY_FILTER,
  initialWorkflowRunHistoryState,
  isWorkflowRunComparisonCompatible,
  isWorkflowRunComparisonEligible,
  listWorkflowRunHistory,
  readWorkflowRunHistoryConfig,
  readWorkflowRunHistoryDetail,
  type WorkflowRunHistoryFilter,
  type WorkflowRunHistorySummary,
} from "./workflowRunHistoryConsumer.ts";

const WorkflowRunComparisonPanel = lazy(() => import("./workflowRunComparisonPanel.tsx"));
const WorkflowEvaluationPanel = lazy(() => import("./workflowEvaluationPanel.tsx"));
const WorkflowEvaluationSuitePanel = lazy(() => import("./workflowEvaluationSuitePanel.tsx"));

const config = readWorkflowRunHistoryConfig();
const ragConfig = readWorkflowRAGSnapshotConfig();

type Props = {
  applicationId: string;
  workspaceId: string;
  applicationActive: boolean;
  activeSurface: WorkflowReviewSurface;
  refreshKey?: number;
  handoffRunId?: string;
  handoffId?: string;
  onHandoffConsumed?: (handoffId: string) => void;
};

export default function WorkflowReviewOwner({
  applicationId,
  workspaceId,
  applicationActive,
  activeSurface,
  refreshKey = 0,
  handoffRunId = "",
  handoffId = "",
  onHandoffConsumed,
}: Props) {
  const [filter, setFilter] = useState<WorkflowRunHistoryFilter>(EMPTY_WORKFLOW_RUN_HISTORY_FILTER);
  const [history, setHistory] = useState(() => initialWorkflowRunHistoryState(config));
  const [detail, setDetail] = useState<WorkflowRunRecord | null>(null);
  const [selectedRunId, setSelectedRunId] = useState("");
  const [baselineRunId, setBaselineRunId] = useState("");
  const [candidateRunId, setCandidateRunId] = useState("");
  const [comparisonSelection, setComparisonSelection] = useState<{ baseline: string; candidate: string } | null>(null);
  const [diagnosticScenario, setDiagnosticScenario] = useState<WorkflowRunDevFailureScenario>("gateway_timeout");
  const [diagnosticState, setDiagnosticState] = useState("");
  const [handoffState, setHandoffState] = useState("");
  const [copiedRef, setCopiedRef] = useState("");
  const [retrievalPreviewPending, setRetrievalPreviewPending] = useState(false);
  const requestGenerationRef = useRef(0);
  const handledHandoffIdRef = useRef("");
  const historyRunsRef = useRef(history.runs);
  historyRunsRef.current = history.runs;

  const liveScope = config.mode === "dev_workflow_executor_http" &&
    workspaceId.trim().length > 0 && config.workspaceId === workspaceId.trim();
  const workspaceMismatch = config.mode === "dev_workflow_executor_http" && !liveScope;

  const loadHistory = useCallback(async (cursor = "", append = false, expectedGeneration?: number) => {
    if (!liveScope) return;
    const generation = expectedGeneration ?? ++requestGenerationRef.current;
    const retainedRuns = append ? historyRunsRef.current : [];
    if (!append) {
      setSelectedRunId("");
      setDetail(null);
      setRetrievalPreviewPending(false);
    }
    setHistory((current) => ({ ...current, status: "loading", failureCode: "", failureSummary: "" }));
    try {
      const next = await listWorkflowRunHistory(applicationId, config, filter, cursor, retainedRuns);
      if (requestGenerationRef.current === generation) setHistory(next);
    } catch (error) {
      if (requestGenerationRef.current !== generation) return;
      setHistory((current) => ({
        ...current,
        status: "failed",
        runs: append ? current.runs : [],
        failureCode: "workflow_run_store_unavailable",
        failureSummary: error instanceof Error ? error.message : "Workflow run history is unavailable.",
      }));
    }
  }, [applicationId, filter, liveScope]);

  useEffect(() => {
    const generation = ++requestGenerationRef.current;
    setHistory(initialWorkflowRunHistoryState(config));
    setDetail(null);
    setSelectedRunId("");
    setBaselineRunId("");
    setCandidateRunId("");
    setComparisonSelection(null);
    setDiagnosticState("");
    setHandoffState("");
    setCopiedRef("");
    setRetrievalPreviewPending(false);
    if (liveScope) void loadHistory("", false, generation);
  }, [applicationId, workspaceId, refreshKey, liveScope]); // Filters are applied explicitly.

  async function selectRun(run: WorkflowRunHistorySummary) {
    const generation = requestGenerationRef.current;
    setSelectedRunId(run.runId);
    setDetail(null);
    setRetrievalPreviewPending(false);
    try {
      const next = await readWorkflowRunHistoryDetail(run, applicationId, config);
      if (requestGenerationRef.current === generation) setDetail(next);
    } catch {
      if (requestGenerationRef.current === generation) setDetail(null);
    }
  }

  useEffect(() => {
    if (!handoffId || !handoffRunId || handledHandoffIdRef.current === handoffId) return;
    handledHandoffIdRef.current = handoffId;
    if (!liveScope) {
      setHandoffState(`Run ${handoffRunId} was not loaded because its owner scope is unavailable. No request or replay was started.`);
      onHandoffConsumed?.(handoffId);
      return;
    }
    const generation = requestGenerationRef.current;
    setSelectedRunId(handoffRunId);
    setDetail(null);
    setHandoffState(`Loading exact run ${handoffRunId} from its owner.`);
    void readWorkflowRunHistoryDetail({ runId: handoffRunId }, applicationId, config)
      .then((record) => {
        if (requestGenerationRef.current !== generation) return;
        setDetail(record);
        setHandoffState(record
          ? `Exact run ${handoffRunId} was reloaded from Run History.`
          : `Run ${handoffRunId} is unavailable in the current Application scope. No replay was started.`);
      })
      .catch(() => {
        if (requestGenerationRef.current !== generation) return;
        setDetail(null);
        setHandoffState(`Run ${handoffRunId} could not be reloaded from its owner. No replay was started.`);
      })
      .finally(() => onHandoffConsumed?.(handoffId));
  }, [applicationId, handoffId, handoffRunId, liveScope, onHandoffConsumed]);

  const comparisonRuns = history.runs.filter(isWorkflowRunComparisonEligible);
  const comparisonBaseline = comparisonRuns.find((run) => run.runId === baselineRunId);
  const comparisonCandidateRuns = comparisonRuns.filter((run) => isWorkflowRunComparisonCompatible(comparisonBaseline, run));
  const failedCount = history.runs.filter((run) => run.status === "failed").length;
  const uncertainCount = history.runs.filter((run) => run.status === "outcome_unknown" || run.staleRunning).length;

  async function generateDiagnosticRun() {
    if (!applicationActive) return;
    if (!filter.draftId.trim()) {
      setDiagnosticState("Enter an exact saved draft id before generating a diagnostic run.");
      return;
    }
    const generation = requestGenerationRef.current;
    setDiagnosticState("Generating deterministic dev/test failure…");
    try {
      const state = await startWorkflowDiagnosticDevRecord(filter.draftId, applicationId, diagnosticScenario, config);
      if (requestGenerationRef.current !== generation) return;
      setDetail(state.record);
      setSelectedRunId(state.record?.runId ?? "");
      setDiagnosticState(state.record
        ? `${diagnosticScenario} recorded for review.`
        : `${state.failureCode ?? "workflow_run_unavailable"}: ${state.failureSummary}`);
      await loadHistory();
    } catch (error) {
      if (requestGenerationRef.current === generation) {
        setDiagnosticState(error instanceof Error ? error.message : "Diagnostic run generation failed.");
      }
    }
  }

  async function loadRetrievalPreviews() {
    const run = history.runs.find((item) => item.runId === selectedRunId);
    if (!run || run.schemaVersion !== "workflow_run_record.v3" || !ragConfig.scopes.has("workflow_rag_snapshots:read")) return;
    const generation = requestGenerationRef.current;
    setRetrievalPreviewPending(true);
    try {
      const next = await readWorkflowRunHistoryDetail(run, applicationId, config, true);
      if (requestGenerationRef.current === generation) setDetail(next);
    } finally {
      if (requestGenerationRef.current === generation) setRetrievalPreviewPending(false);
    }
  }

  if (workspaceMismatch) {
    return (
      <OwnerBoundary
        title="Workflow review scope mismatch"
        copy={`Application Workspace is ${workspaceId || "unavailable"}; Run History is configured for ${config.workspaceId}. Zero owner requests were sent.`}
      />
    );
  }

  if (config.mode !== "dev_workflow_executor_http") {
    return (
      <OwnerBoundary
        title="Workflow review is offline"
        copy="The offline fixture is evidence only and is not projected as current-application history, comparison, evaluation, or release review. Zero owner requests were sent."
      />
    );
  }

  return (
    <section className="workflow-review-owner-surface" aria-label={`${activeSurface} workflow review owner`}>
      {handoffState ? <p className="workflow-review-owner-message" role="status">{handoffState}</p> : null}
      <header className="workflow-review-owner-heading">
        <div>
          <p className="eyebrow">{surfaceEyebrow(activeSurface)}</p>
          <h4>{surfaceTitle(activeSurface)}</h4>
          <p>{surfaceDescription(activeSurface)}</p>
        </div>
        <div className="workflow-review-owner-badges">
          <span className="status-badge neutral">durable dev / test</span>
          <span className={`status-badge ${applicationActive ? "good" : "neutral"}`}>
            {applicationActive ? "active" : "archived · read only"}
          </span>
        </div>
      </header>

      {activeSurface === "runs" ? (
        <RunsSurface
          applicationId={applicationId}
          applicationActive={applicationActive}
          filter={filter}
          history={history}
          selectedRunId={selectedRunId}
          detail={detail}
          failedCount={failedCount}
          uncertainCount={uncertainCount}
          diagnosticScenario={diagnosticScenario}
          diagnosticState={diagnosticState}
          copiedRef={copiedRef}
          retrievalPreviewPending={retrievalPreviewPending}
          onFilterChange={setFilter}
          onApplyFilters={() => void loadHistory()}
          onSelectRun={(run) => void selectRun(run)}
          onLoadEarlier={() => void loadHistory(history.nextCursor, true)}
          onDiagnosticScenarioChange={setDiagnosticScenario}
          onGenerateDiagnostic={() => void generateDiagnosticRun()}
          onCopyReference={(label, value) => {
            void navigator.clipboard.writeText(value).then(() => setCopiedRef(label));
          }}
          onLoadRetrievalPreviews={() => void loadRetrievalPreviews()}
        />
      ) : null}

      {activeSurface === "comparison" ? (
        <ComparisonSurface
          applicationId={applicationId}
          runs={comparisonRuns}
          baselineRunId={baselineRunId}
          candidateRunId={candidateRunId}
          candidateRuns={comparisonCandidateRuns}
          selection={comparisonSelection}
          onBaselineChange={(value) => {
            setBaselineRunId(value);
            setCandidateRunId("");
            setComparisonSelection(null);
          }}
          onCandidateChange={setCandidateRunId}
          onCompare={() => setComparisonSelection({ baseline: baselineRunId, candidate: candidateRunId })}
        />
      ) : null}

      {activeSurface === "cases" ? (
        <Suspense fallback={<p>Loading evaluation cases…</p>}>
          <WorkflowEvaluationPanel
            applicationId={applicationId}
            runs={comparisonRuns}
            config={config}
            readOnly={!applicationActive}
          />
        </Suspense>
      ) : null}

      {activeSurface === "release" ? (
        <Suspense fallback={<p>Loading evaluation suites…</p>}>
          <WorkflowEvaluationSuitePanel
            applicationId={applicationId}
            config={config}
            readOnly={!applicationActive}
          />
        </Suspense>
      ) : null}
    </section>
  );
}

function RunsSurface({
  applicationId,
  applicationActive,
  filter,
  history,
  selectedRunId,
  detail,
  failedCount,
  uncertainCount,
  diagnosticScenario,
  diagnosticState,
  copiedRef,
  retrievalPreviewPending,
  onFilterChange,
  onApplyFilters,
  onSelectRun,
  onLoadEarlier,
  onDiagnosticScenarioChange,
  onGenerateDiagnostic,
  onCopyReference,
  onLoadRetrievalPreviews,
}: {
  applicationId: string;
  applicationActive: boolean;
  filter: WorkflowRunHistoryFilter;
  history: ReturnType<typeof initialWorkflowRunHistoryState>;
  selectedRunId: string;
  detail: WorkflowRunRecord | null;
  failedCount: number;
  uncertainCount: number;
  diagnosticScenario: WorkflowRunDevFailureScenario;
  diagnosticState: string;
  copiedRef: string;
  retrievalPreviewPending: boolean;
  onFilterChange: (filter: WorkflowRunHistoryFilter) => void;
  onApplyFilters: () => void;
  onSelectRun: (run: WorkflowRunHistorySummary) => void;
  onLoadEarlier: () => void;
  onDiagnosticScenarioChange: (scenario: WorkflowRunDevFailureScenario) => void;
  onGenerateDiagnostic: () => void;
  onCopyReference: (label: string, value: string) => void;
  onLoadRetrievalPreviews: () => void;
}) {
  return (
    <div className="workflow-review-runs-layout">
      <aside className="workflow-review-run-list-pane">
        <div className="workflow-review-window-summary">
          <div><span>Current window</span><strong>{history.runs.length}</strong></div>
          <div><span>Failed</span><strong>{failedCount}</strong></div>
          <div><span>Uncertain</span><strong>{uncertainCount}</strong></div>
        </div>
        <details className="workflow-review-disclosure">
          <summary><span>Filters</span><small>Exact scope and cursor</small></summary>
          <RunFilters filter={filter} onChange={onFilterChange} onApply={onApplyFilters} loading={history.status === "loading"} />
        </details>
        {config.diagnosticsDevEnabled ? (
          <details className="workflow-review-disclosure">
            <summary><span>Diagnostic run</span><small>Explicit dev / test gate</small></summary>
            <fieldset disabled={!applicationActive}>
              <label>Scenario<select value={diagnosticScenario} onChange={(event) => onDiagnosticScenarioChange(event.target.value as WorkflowRunDevFailureScenario)}><option value="gateway_timeout">Gateway timeout</option><option value="gateway_queue_full">Gateway queue full</option><option value="gateway_worker_crash">Gateway worker crash</option><option value="gateway_protocol_failure">Gateway protocol</option><option value="provider_failed">Provider failed</option><option value="output_unavailable">Output unavailable</option><option value="request_canceled">Request canceled</option><option value="run_store_unavailable">Store unavailable</option><option value="terminal_write_conflict">Terminal write conflict</option><option value="budget_exceeded">Budget exceeded</option><option value="stale_running">Stale running</option></select></label>
              <button type="button" onClick={onGenerateDiagnostic}>Generate exact diagnostic</button>
            </fieldset>
            <p>{applicationActive ? diagnosticState || "Uses the exact Draft filter; arbitrary failure payloads are rejected." : "Archived applications cannot create diagnostic runs."}</p>
          </details>
        ) : null}
        {history.failureCode ? <p className="failure-summary">{history.failureCode}: {history.failureSummary}</p> : null}
        <div className="workflow-review-run-list" aria-label="Exact workflow run records">
          {history.runs.map((run) => (
            <button
              type="button"
              className={`workflow-review-run-row ${selectedRunId === run.runId ? "is-selected" : ""}`}
              key={run.runId}
              aria-pressed={selectedRunId === run.runId}
              onClick={() => onSelectRun(run)}
            >
              <span><strong>{run.runId}</strong><small>{runSource(run)}</small></span>
              <span className={`workflow-review-run-status status-${run.status}`}>{run.status}{run.staleRunning ? " · stale" : ""}</span>
              <small>{run.startedAt || run.schemaVersion}</small>
            </button>
          ))}
          {history.status === "loading" && history.runs.length === 0 ? <p role="status">Loading the current run window…</p> : null}
          {history.status === "empty" ? <p>No run records match the exact filters.</p> : null}
        </div>
        {history.hasMore ? <button type="button" className="workflow-review-load-earlier" onClick={onLoadEarlier} disabled={history.status === "loading"}>Load earlier runs</button> : null}
      </aside>
      <main className="workflow-review-run-detail-pane">
        {detail ? (
          <RunDetail
            applicationId={applicationId}
            detail={detail}
            copiedRef={copiedRef}
            retrievalPreviewPending={retrievalPreviewPending}
            onCopyReference={onCopyReference}
            onLoadRetrievalPreviews={onLoadRetrievalPreviews}
          />
        ) : (
          <div className="workflow-review-empty-detail">
            <span aria-hidden="true">◎</span>
            <h5>Select an exact run</h5>
            <p>Detail remains metadata-only. Selecting a record never replays, resumes, retries, or changes application state.</p>
          </div>
        )}
      </main>
    </div>
  );
}

function RunFilters({ filter, onChange, onApply, loading }: { filter: WorkflowRunHistoryFilter; onChange: (filter: WorkflowRunHistoryFilter) => void; onApply: () => void; loading: boolean }) {
  return (
    <div className="workflow-review-filter-grid">
      <label>Status<select value={filter.status} onChange={(event) => onChange({ ...filter, status: event.target.value as WorkflowRunHistoryFilter["status"] })}><option value="">All</option><option value="succeeded">Succeeded</option><option value="failed">Failed</option><option value="outcome_unknown">Outcome unknown</option><option value="canceled">Canceled</option><option value="running">Running</option></select></label>
      <label>Draft<input value={filter.draftId} onChange={(event) => onChange({ ...filter, draftId: event.target.value })} placeholder="exact draft id" /></label>
      <label>Execution source<select value={filter.executionSourceKind} onChange={(event) => onChange({ ...filter, executionSourceKind: event.target.value as WorkflowRunHistoryFilter["executionSourceKind"] })}><option value="">All</option><option value="workflow_draft">Saved Draft</option><option value="workflow_definition">Workflow Definition</option><option value="application_configuration_draft">Application RAG</option><option value="prompt_application_template">Prompt Application</option><option value="agent_copilot_profile">Agent Copilot</option></select></label>
      <label>Source ID<input value={filter.executionSourceId} onChange={(event) => onChange({ ...filter, executionSourceId: event.target.value })} placeholder="exact source id" /></label>
      <label>Source version<input type="number" min="1" value={filter.executionSourceVersion} onChange={(event) => onChange({ ...filter, executionSourceVersion: event.target.value ? Number(event.target.value) : "" })} /></label>
      <label>Started from<input type="datetime-local" value={filter.startedFrom} onChange={(event) => onChange({ ...filter, startedFrom: event.target.value })} /></label>
      <label>Started to<input type="datetime-local" value={filter.startedTo} onChange={(event) => onChange({ ...filter, startedTo: event.target.value })} /></label>
      <label>Failure code<input value={filter.failureCode} onChange={(event) => onChange({ ...filter, failureCode: event.target.value })} placeholder="workflow_run_…" /></label>
      <label>Boundary<select value={filter.failureBoundary} onChange={(event) => onChange({ ...filter, failureBoundary: event.target.value as WorkflowRunHistoryFilter["failureBoundary"] })}><option value="">All</option><option value="executor">Executor</option><option value="gateway">Gateway</option><option value="provider">Provider</option><option value="run_store">Run store</option><option value="request">Request</option><option value="draft_read">Draft read</option><option value="tool_policy">Tool policy</option><option value="tool_confirmation">Tool confirmation</option><option value="tool_transport">Tool transport</option><option value="tool_response">Tool response</option><option value="tool_store">Tool store</option><option value="retrieval_policy">Retrieval policy</option><option value="retrieval_store">Retrieval store</option><option value="retrieval_rank">Retrieval rank</option><option value="retrieval_context">Retrieval context</option><option value="retrieval_citation">Retrieval citation</option><option value="provider_selection">Provider selection</option><option value="provider_call">Provider call</option></select></label>
      <label>Provider<input value={filter.provider} onChange={(event) => onChange({ ...filter, provider: event.target.value })} placeholder="exact provider" /></label>
      <label>Model<input value={filter.model} onChange={(event) => onChange({ ...filter, model: event.target.value })} placeholder="exact model" /></label>
      <label>Stale running<select value={filter.staleRunning} onChange={(event) => onChange({ ...filter, staleRunning: event.target.value as WorkflowRunHistoryFilter["staleRunning"] })}><option value="">All</option><option value="true">Only stale</option><option value="false">Exclude stale</option></select></label>
      <button type="button" onClick={onApply} disabled={loading}>Apply exact filters</button>
    </div>
  );
}

function RunDetail({ detail, copiedRef, retrievalPreviewPending, onCopyReference, onLoadRetrievalPreviews }: { applicationId: string; detail: WorkflowRunRecord; copiedRef: string; retrievalPreviewPending: boolean; onCopyReference: (label: string, value: string) => void; onLoadRetrievalPreviews: () => void }) {
  const forbiddenWrites = detail.sideEffects.businessWrites + detail.sideEffects.replayWrites;
  const canReadPreviews = detail.schemaVersion === "workflow_run_record.v3" && ragConfig.mode === "dev_workflow_rag_http" && ragConfig.scopes.has("workflow_rag_snapshots:read");
  return (
    <article className="workflow-review-run-detail">
      <header>
        <div><p className="eyebrow">Exact run detail</p><h5>{detail.runId}</h5><code>{detail.schemaVersion}</code></div>
        <span className={`status-badge ${detail.status === "succeeded" ? "good" : detail.status === "failed" ? "bad" : "neutral"}`}>{detail.status}</span>
      </header>
      <p>{detail.output || detail.failureSummary || "No advisory output is retained in this metadata-only record."}</p>
      <dl className="workflow-review-detail-facts">
        <div><dt>Input</dt><dd>{detail.inputBytes} bytes; raw text not retained</dd></div>
        <div><dt>Recorded window</dt><dd>{detail.startedAt} → {detail.completedAt || "open"}</dd></div>
        <div><dt>Provider calls</dt><dd>{detail.sideEffects.providerCalls}</dd></div>
        <div><dt>Forbidden writes</dt><dd>{forbiddenWrites}</dd></div>
        {detail.executionSourceId ? <div><dt>Execution source</dt><dd>{detail.executionSourceKind} · {detail.executionSourceId} · v{detail.executionSourceVersion}</dd></div> : null}
        {detail.schemaVersion === "workflow_run_record.v8" ? <><div><dt>Input contract</dt><dd>{detail.inputContractId} · {detail.inputContractDigest}</dd></div><div><dt>Input fields</dt><dd>{detail.inputFields?.map((field) => `${field.name}:${field.valueType}`).join(" · ") || "none"}</dd></div></> : null}
        <div><dt>Failure boundary</dt><dd>{detail.diagnostic?.failureBoundary || "none"}</dd></div>
      </dl>
      {detail.diagnostic ? (
        <section className="workflow-review-diagnostic">
          <div><strong>{detail.diagnostic.failureBoundary || "No failure"}</strong><span>{detail.diagnostic.summary}</span></div>
          <dl><div><dt>Failed node</dt><dd>{detail.diagnostic.failedNodeId || "none"}</dd></div><div><dt>Last completed</dt><dd>{detail.diagnostic.lastCompletedNodeId || "none"}</dd></div><div><dt>Review action</dt><dd>{detail.diagnostic.recommendedReviewAction || "none"}</dd></div></dl>
        </section>
      ) : null}
      <details className="workflow-review-disclosure">
        <summary><span>Node evidence</span><small>{detail.nodes.length} nodes</small></summary>
        <div className="workflow-review-node-list">
          {detail.nodes.map((node) => <div key={node.nodeId}><span><strong>{node.label}</strong><small>{node.nodeType}</small></span><span className={`status-${node.status}`}>{node.status}</span><small>{node.durationMs} ms</small></div>)}
        </div>
      </details>
      {detail.retrievalAttempt ? (
        <details className="workflow-review-disclosure">
          <summary><span>Retrieval evidence</span><small>{detail.retrievalAttempt.selectedFragments.length} selected refs</small></summary>
          <dl className="workflow-review-detail-facts"><div><dt>Snapshot</dt><dd>{detail.ragSnapshot?.snapshotId} · v{detail.ragSnapshot?.snapshotVersion}</dd></div><div><dt>Query</dt><dd>{detail.retrievalAttempt.queryBytes} bytes · {detail.retrievalAttempt.queryDigest}</dd></div><div><dt>Context</dt><dd>{detail.retrievalAttempt.contextBytes} bytes · {detail.retrievalAttempt.citationRefs.length} citations</dd></div></dl>
          <button type="button" disabled={!canReadPreviews || retrievalPreviewPending} onClick={onLoadRetrievalPreviews}>{retrievalPreviewPending ? "Loading previews…" : "Read authorized fragment previews"}</button>
          {detail.retrievalFragmentPreviews.map((preview) => <blockquote key={preview.fragmentRef}><strong>{preview.fragmentRef}</strong><p>{preview.preview}</p></blockquote>)}
        </details>
      ) : null}
      <div className="workflow-review-reference-actions">
        <button type="button" onClick={() => onCopyReference("request", detail.requestId)}>Copy request id</button>
        <button type="button" onClick={() => onCopyReference("audit", detail.auditRef)}>Copy audit ref</button>
        <span>{copiedRef ? `${copiedRef} copied` : "Metadata references only."}</span>
      </div>
      <p className="workflow-review-stop-line">Business writes and replay remain locked at 0. This detail cannot retry, resume, execute, publish, or deploy.</p>
    </article>
  );
}

function ComparisonSurface({ applicationId, runs, baselineRunId, candidateRunId, candidateRuns, selection, onBaselineChange, onCandidateChange, onCompare }: { applicationId: string; runs: WorkflowRunHistorySummary[]; baselineRunId: string; candidateRunId: string; candidateRuns: WorkflowRunHistorySummary[]; selection: { baseline: string; candidate: string } | null; onBaselineChange: (value: string) => void; onCandidateChange: (value: string) => void; onCompare: () => void }) {
  return (
    <div className="workflow-review-comparison-layout">
      <aside className="workflow-review-pair-rail">
        <div><p className="eyebrow">Exact run pair</p><h5>Choose compatible evidence</h5><p>Compatibility follows the current run profile and immutable authority fields.</p></div>
        <label>Baseline<RunSelect value={baselineRunId} runs={runs} onChange={onBaselineChange} /></label>
        <label>Candidate<RunSelect value={candidateRunId} runs={candidateRuns} onChange={onCandidateChange} /></label>
        <button type="button" disabled={!baselineRunId || !candidateRunId || baselineRunId === candidateRunId} onClick={onCompare}>Compare exact runs</button>
        <p className="workflow-review-stop-line">Instantaneous read-only view. It is not persisted and does not rerun either workflow.</p>
      </aside>
      <main className="workflow-review-comparison-detail">
        {selection ? (
          <Suspense fallback={<p>Comparing durable run records…</p>}>
            <WorkflowRunComparisonPanel applicationId={applicationId} baselineRunId={selection.baseline} candidateRunId={selection.candidate} config={config} />
          </Suspense>
        ) : (
          <div className="workflow-review-empty-detail"><span aria-hidden="true">⇄</span><h5>Choose a compatible pair</h5><p>Status, findings, retrieval authority, and node deltas will remain in one dominant comparison surface.</p></div>
        )}
      </main>
    </div>
  );
}

function RunSelect({ value, runs, onChange }: { value: string; runs: WorkflowRunHistorySummary[]; onChange: (value: string) => void }) {
  return <select value={value} onChange={(event) => onChange(event.target.value)}><option value="">Choose exact run</option>{runs.map((run) => <option key={run.runId} value={run.runId}>{run.runId} · {run.status} · {runSource(run)}</option>)}</select>;
}

function OwnerBoundary({ title, copy }: { title: string; copy: string }) {
  return (
    <article className="workflow-review-owner-boundary" role="status">
      <div>
        <p className="eyebrow">Fail-closed owner</p>
        <h4>{title}</h4>
        <p>{copy}</p>
      </div>
      <dl>
        <div><dt>Projection</dt><dd>separate offline evidence</dd></div>
        <div><dt>Owner requests</dt><dd>0</dd></div>
        <div><dt>Mutation</dt><dd>unavailable</dd></div>
      </dl>
      <ul>
        <li>No current-application run list or exact detail is inferred.</li>
        <li>No comparison, evaluation result, suite decision, or release authority is fabricated.</li>
        <li>An explicit dev/test source with matching workspace scope is required.</li>
      </ul>
      <strong>No current-application capability is inferred from offline evidence.</strong>
    </article>
  );
}

function runSource(run: WorkflowRunHistorySummary): string {
  if (run.schemaVersion === "workflow_run_record.v8") return `Structured Workflow Definition · ${run.executionSourceId}`;
  if (run.schemaVersion === "workflow_run_record.v7") return `Agent Copilot · ${run.executionSourceId}`;
  if (run.schemaVersion === "workflow_run_record.v6") return `Prompt Application · ${run.executionSourceId}`;
  if (run.schemaVersion === "workflow_run_record.v5") return `Workflow Definition · ${run.executionSourceId}`;
  if (run.schemaVersion === "workflow_run_record.v4") return `Application RAG · ${run.executionSourceId}`;
  if (run.schemaVersion === "workflow_run_record.v3") return `Workflow RAG · ${run.ragRef}`;
  return `Saved Draft · ${run.draftId}`;
}

function surfaceEyebrow(surface: WorkflowReviewSurface): string {
  if (surface === "runs") return "01 · Durable run evidence";
  if (surface === "comparison") return "02 · Regression comparison";
  if (surface === "cases") return "03 · Versioned evaluation cases";
  return "04 · Human release review";
}

function surfaceTitle(surface: WorkflowReviewSurface): string {
  if (surface === "runs") return "Locate and inspect runs";
  if (surface === "comparison") return "Compare compatible runs";
  if (surface === "cases") return "Evaluation cases";
  return "Evaluation suites and decisions";
}

function surfaceDescription(surface: WorkflowReviewSurface): string {
  if (surface === "runs") return "Read the current keyset-paged window and one exact metadata-only detail.";
  if (surface === "comparison") return "Inspect instantaneous deltas without persisting a comparison or re-executing either run.";
  if (surface === "cases") return "Bind exact run references and expected classifications into versioned review evidence.";
  return "Review exact case versions and append a digest-bound human decision; approval is not release authority.";
}
