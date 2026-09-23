import "../../i18n/runReviewResources.ts";
import { useLocalePreference } from "../../i18n/LocaleProvider.tsx";
import { formatDisplayDate } from "../../i18n/formatters.ts";
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation("evaluation");
  const [filter, setFilter] = useState<WorkflowRunHistoryFilter>(EMPTY_WORKFLOW_RUN_HISTORY_FILTER);
  const [history, setHistory] = useState(() => initialWorkflowRunHistoryState(config));
  const [detail, setDetail] = useState<WorkflowRunRecord | null>(null);
  const [selectedRunId, setSelectedRunId] = useState("");
  const [baselineRunId, setBaselineRunId] = useState("");
  const [candidateRunId, setCandidateRunId] = useState("");
  const [comparisonSelection, setComparisonSelection] = useState<{ baseline: string; candidate: string } | null>(null);
  const [diagnosticScenario, setDiagnosticScenario] = useState<WorkflowRunDevFailureScenario>("gateway_timeout");
  const [diagnosticState, setDiagnosticState] = useState("");
  const [handoffState, setHandoffState] = useState<{ kind: "handoffOffline" | "handoffLoading" | "handoffReady" | "handoffUnavailable" | "handoffFailed"; runId: string } | null>(null);
  const [copiedRef, setCopiedRef] = useState("");
  const [retrievalPreviewPending, setRetrievalPreviewPending] = useState(false);
  const requestGenerationRef = useRef(0);
  const initializedOwnerScopeRef = useRef("");
  const handledHandoffIdRef = useRef("");
  const historyRunsRef = useRef(history.runs);
  historyRunsRef.current = history.runs;

  const liveScope = config.mode === "dev_workflow_executor_http" &&
    workspaceId.trim().length > 0 && config.workspaceId === workspaceId.trim();
  const workspaceMismatch = config.mode === "dev_workflow_executor_http" && !liveScope;
  const ownerScopeKey = `${applicationId}\u0000${workspaceId}\u0000${refreshKey}\u0000${String(liveScope)}`;

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
    if (initializedOwnerScopeRef.current === ownerScopeKey) return;
    initializedOwnerScopeRef.current = ownerScopeKey;
    const generation = ++requestGenerationRef.current;
    setHistory(initialWorkflowRunHistoryState(config));
    setDetail(null);
    setSelectedRunId("");
    setBaselineRunId("");
    setCandidateRunId("");
    setComparisonSelection(null);
    setDiagnosticState("");
    setHandoffState(null);
    setCopiedRef("");
    setRetrievalPreviewPending(false);
    if (liveScope) void loadHistory("", false, generation);
  }, [applicationId, workspaceId, refreshKey, liveScope, ownerScopeKey]); // Filters are applied explicitly.

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
      setHandoffState({ kind: "handoffOffline", runId: handoffRunId });
      onHandoffConsumed?.(handoffId);
      return;
    }
    const generation = requestGenerationRef.current;
    setSelectedRunId(handoffRunId);
    setDetail(null);
    setHandoffState({ kind: "handoffLoading", runId: handoffRunId });
    void readWorkflowRunHistoryDetail({ runId: handoffRunId }, applicationId, config)
      .then((record) => {
        if (requestGenerationRef.current !== generation) return;
        setDetail(record);
        setHandoffState(record
          ? { kind: "handoffReady", runId: handoffRunId }
          : { kind: "handoffUnavailable", runId: handoffRunId });
      })
      .catch(() => {
        if (requestGenerationRef.current !== generation) return;
        setDetail(null);
        setHandoffState({ kind: "handoffFailed", runId: handoffRunId });
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
        title={t($ => $.runReview.workflowReviewScopeMismatch)}
        copy={`Application Workspace is ${workspaceId || "unavailable"}; Run History is configured for ${config.workspaceId}. Zero owner requests were sent.`}
      />
    );
  }

  if (config.mode !== "dev_workflow_executor_http") {
    return (
      <OwnerBoundary
        title={t($ => $.runReview.workflowReviewIsOffline)}
        copy="The offline fixture is evidence only and is not projected as current-application history, comparison, evaluation, or release review. Zero owner requests were sent."
      />
    );
  }

  return (
    <section className="workflow-review-owner-surface" aria-label={t($ => $.runReview.ownerRegion)}>
      {handoffState ? <p className="workflow-review-owner-message" role="status">{t($ => $.runReview[handoffState.kind], { runId: handoffState.runId })}</p> : null}
      <header className="workflow-review-owner-heading">
        <div>
          <p className="eyebrow">{t($ => $.runReview.surfaces[activeSurface].eyebrow)}</p>
          <h4>{t($ => $.runReview.surfaces[activeSurface].title)}</h4>
          <p>{t($ => $.runReview.surfaces[activeSurface].description)}</p>
        </div>
        <div className="workflow-review-owner-badges">
          <span className="status-badge neutral">{t($ => $.runReview.durableDevTest)}</span>
          <span className={`status-badge ${applicationActive ? "good" : "neutral"}`}>
            {applicationActive ? t($ => $.runReview.active) : t($ => $.runReview.archivedReadOnly)}
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
        <Suspense fallback={<p>{t($ => $.runReview.loadingEvaluationCases)}</p>}>
          <WorkflowEvaluationPanel
            applicationId={applicationId}
            runs={comparisonRuns}
            config={config}
            readOnly={!applicationActive}
          />
        </Suspense>
      ) : null}

      {activeSurface === "release" ? (
        <Suspense fallback={<p>{t($ => $.runReview.loadingEvaluationSuites)}</p>}>
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
  const { t } = useTranslation("evaluation");
  const { locale } = useLocalePreference();
  return (
    <div className="workflow-review-runs-layout">
      <aside className="workflow-review-run-list-pane">
        <div className="workflow-review-window-summary">
          <div><span>{t($ => $.runReview.currentWindow)}</span><strong>{history.runs.length}</strong></div>
          <div><span>{t($ => $.runReview.failed)}</span><strong>{failedCount}</strong></div>
          <div><span>{t($ => $.runReview.uncertain)}</span><strong>{uncertainCount}</strong></div>
        </div>
        <details className="workflow-review-disclosure">
          <summary><span>{t($ => $.runReview.filters)}</span><small>{t($ => $.runReview.exactScopeAndCursor)}</small></summary>
          <RunFilters filter={filter} onChange={onFilterChange} onApply={onApplyFilters} loading={history.status === "loading"} />
        </details>
        {config.diagnosticsDevEnabled ? (
          <details className="workflow-review-disclosure">
            <summary><span>{t($ => $.runReview.diagnosticRun)}</span><small>{t($ => $.runReview.explicitDevTestGate)}</small></summary>
            <fieldset disabled={!applicationActive}>
              <label>{t($ => $.runReview.scenario)}<select value={diagnosticScenario} onChange={(event) => onDiagnosticScenarioChange(event.target.value as WorkflowRunDevFailureScenario)}><option value="gateway_timeout">{t($ => $.runReview.gatewayTimeout)}</option><option value="gateway_queue_full">{t($ => $.runReview.gatewayQueueFull)}</option><option value="gateway_worker_crash">{t($ => $.runReview.gatewayWorkerCrash)}</option><option value="gateway_protocol_failure">{t($ => $.runReview.gatewayProtocol)}</option><option value="provider_failed">{t($ => $.runReview.providerFailed)}</option><option value="output_unavailable">{t($ => $.runReview.outputUnavailable)}</option><option value="request_canceled">{t($ => $.runReview.requestCanceled)}</option><option value="run_store_unavailable">{t($ => $.runReview.storeUnavailable)}</option><option value="terminal_write_conflict">{t($ => $.runReview.terminalWriteConflict)}</option><option value="budget_exceeded">{t($ => $.runReview.budgetExceeded)}</option><option value="stale_running">{t($ => $.runReview.staleRunning)}</option></select></label>
              <button type="button" onClick={onGenerateDiagnostic}>{t($ => $.runReview.generateExactDiagnostic)}</button>
            </fieldset>
            <p>{applicationActive ? diagnosticState || t($ => $.runReview.usesTheExactDraftFilterArbitraryFailurePayloads) : t($ => $.runReview.archivedApplicationsCannotCreateDiagnosticRuns)}</p>
          </details>
        ) : null}
        {history.failureCode ? <p className="failure-summary">{history.failureCode}: {t($ => $.runReview.historyFailed)}</p> : null}
        <div className="workflow-review-run-list" aria-label={t($ => $.runReview.exactWorkflowRunRecords)}>
          {history.runs.map((run) => (
            <button
              type="button"
              className={`workflow-review-run-row ${selectedRunId === run.runId ? "is-selected" : ""}`}
              key={run.runId}
              aria-pressed={selectedRunId === run.runId}
              onClick={() => onSelectRun(run)}
            >
              <span><strong>{run.runId}</strong><small>{t($ => $.runReview.sources[runSource(run)])} · {run.executionSourceId || run.ragRef || run.draftId}</small></span>
              <span className={`workflow-review-run-status status-${t($ => $.runReview.statuses[run.status])}`}>{t($ => $.runReview.statuses[run.status])}{run.staleRunning ? t($ => $.runReview.stale) : ""}</span>
              <small>{formatDisplayDate(run.startedAt, locale) ?? t($ => $.runReview.unknownTime)}</small>
            </button>
          ))}
          {history.status === "loading" && history.runs.length === 0 ? <p role="status">{t($ => $.runReview.loadingTheCurrentRunWindow)}</p> : null}
          {history.status === "empty" ? <p>{t($ => $.runReview.noRunRecordsMatchTheExactFilters)}</p> : null}
        </div>
        {history.hasMore ? <button type="button" className="workflow-review-load-earlier" onClick={onLoadEarlier} disabled={history.status === "loading"}>{t($ => $.runReview.loadEarlierRuns)}</button> : null}
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
            <h5>{t($ => $.runReview.selectAnExactRun)}</h5>
            <p>{t($ => $.runReview.detailRemainsMetadataOnlySelectingARecordNever)}</p>
          </div>
        )}
      </main>
    </div>
  );
}

function RunFilters({ filter, onChange, onApply, loading }: { filter: WorkflowRunHistoryFilter; onChange: (filter: WorkflowRunHistoryFilter) => void; onApply: () => void; loading: boolean }) {
  const { t } = useTranslation("evaluation");
  return (
    <div className="workflow-review-filter-grid">
      <label>{t($ => $.runReview.status)}<select value={filter.status} onChange={(event) => onChange({ ...filter, status: event.target.value as WorkflowRunHistoryFilter["status"] })}><option value="">{t($ => $.runReview.all)}</option><option value="succeeded">{t($ => $.runReview.succeeded)}</option><option value="failed">{t($ => $.runReview.failed)}</option><option value="outcome_unknown">{t($ => $.runReview.outcomeUnknown)}</option><option value="canceled">{t($ => $.runReview.canceled)}</option><option value="running">{t($ => $.runReview.running)}</option></select></label>
      <label>{t($ => $.runReview.draft)}<input value={filter.draftId} onChange={(event) => onChange({ ...filter, draftId: event.target.value })} placeholder={t($ => $.runReview.exactDraftId)} /></label>
      <label>{t($ => $.runReview.executionSource)}<select value={filter.executionSourceKind} onChange={(event) => onChange({ ...filter, executionSourceKind: event.target.value as WorkflowRunHistoryFilter["executionSourceKind"] })}><option value="">{t($ => $.runReview.all)}</option><option value="workflow_draft">{t($ => $.runReview.savedDraft)}</option><option value="workflow_definition">{t($ => $.runReview.workflowDefinition)}</option><option value="application_configuration_draft">{t($ => $.runReview.applicationRag)}</option><option value="prompt_application_template">{t($ => $.runReview.promptApplication)}</option><option value="agent_copilot_profile">{t($ => $.runReview.agentCopilot)}</option></select></label>
      <label>{t($ => $.runReview.sourceId)}<input value={filter.executionSourceId} onChange={(event) => onChange({ ...filter, executionSourceId: event.target.value })} placeholder={t($ => $.runReview.exactSourceId)} /></label>
      <label>{t($ => $.runReview.sourceVersion)}<input type="number" min="1" value={filter.executionSourceVersion} onChange={(event) => onChange({ ...filter, executionSourceVersion: event.target.value ? Number(event.target.value) : "" })} /></label>
      <label>{t($ => $.runReview.startedFrom)}<input type="datetime-local" value={filter.startedFrom} onChange={(event) => onChange({ ...filter, startedFrom: event.target.value })} /></label>
      <label>{t($ => $.runReview.startedTo)}<input type="datetime-local" value={filter.startedTo} onChange={(event) => onChange({ ...filter, startedTo: event.target.value })} /></label>
      <label>{t($ => $.runReview.failureCode)}<input value={filter.failureCode} onChange={(event) => onChange({ ...filter, failureCode: event.target.value })} placeholder="workflow_run_…" /></label>
      <label>{t($ => $.runReview.boundary)}<select value={filter.failureBoundary} onChange={(event) => onChange({ ...filter, failureBoundary: event.target.value as WorkflowRunHistoryFilter["failureBoundary"] })}><option value="">{t($ => $.runReview.all)}</option><option value="executor">{t($ => $.runReview.executor)}</option><option value="gateway">{t($ => $.runReview.gateway)}</option><option value="provider">{t($ => $.runReview.provider)}</option><option value="run_store">{t($ => $.runReview.runStore)}</option><option value="request">{t($ => $.runReview.request)}</option><option value="draft_read">{t($ => $.runReview.draftRead)}</option><option value="tool_policy">{t($ => $.runReview.toolPolicy)}</option><option value="tool_confirmation">{t($ => $.runReview.toolConfirmation)}</option><option value="tool_transport">{t($ => $.runReview.toolTransport)}</option><option value="tool_response">{t($ => $.runReview.toolResponse)}</option><option value="tool_store">{t($ => $.runReview.toolStore)}</option><option value="retrieval_policy">{t($ => $.runReview.retrievalPolicy)}</option><option value="retrieval_store">{t($ => $.runReview.retrievalStore)}</option><option value="retrieval_rank">{t($ => $.runReview.retrievalRank)}</option><option value="retrieval_context">{t($ => $.runReview.retrievalContext)}</option><option value="retrieval_citation">{t($ => $.runReview.retrievalCitation)}</option><option value="provider_selection">{t($ => $.runReview.providerSelection)}</option><option value="provider_call">{t($ => $.runReview.providerCall)}</option></select></label>
      <label>{t($ => $.runReview.provider)}<input value={filter.provider} onChange={(event) => onChange({ ...filter, provider: event.target.value })} placeholder={t($ => $.runReview.exactProvider)} /></label>
      <label>{t($ => $.runReview.model)}<input value={filter.model} onChange={(event) => onChange({ ...filter, model: event.target.value })} placeholder={t($ => $.runReview.exactModel)} /></label>
      <label>{t($ => $.runReview.staleRunning)}<select value={filter.staleRunning} onChange={(event) => onChange({ ...filter, staleRunning: event.target.value as WorkflowRunHistoryFilter["staleRunning"] })}><option value="">{t($ => $.runReview.all)}</option><option value="true">{t($ => $.runReview.onlyStale)}</option><option value="false">{t($ => $.runReview.excludeStale)}</option></select></label>
      <button type="button" onClick={onApply} disabled={loading}>{t($ => $.runReview.applyExactFilters)}</button>
    </div>
  );
}

function RunDetail({ detail, copiedRef, retrievalPreviewPending, onCopyReference, onLoadRetrievalPreviews }: { applicationId: string; detail: WorkflowRunRecord; copiedRef: string; retrievalPreviewPending: boolean; onCopyReference: (label: string, value: string) => void; onLoadRetrievalPreviews: () => void }) {
  const { t } = useTranslation("evaluation");
  const { locale } = useLocalePreference();
  const forbiddenWrites = detail.sideEffects.businessWrites + detail.sideEffects.replayWrites;
  const canReadPreviews = detail.schemaVersion === "workflow_run_record.v3" && ragConfig.mode === "dev_workflow_rag_http" && ragConfig.scopes.has("workflow_rag_snapshots:read");
  return (
    <article className="workflow-review-run-detail">
      <header>
        <div><p className="eyebrow">{t($ => $.runReview.exactRunDetail)}</p><h5>{detail.runId}</h5><code>{detail.schemaVersion}</code></div>
        <span className={`status-badge ${detail.status === "succeeded" ? "good" : detail.status === "failed" ? "bad" : "neutral"}`}>{t($ => $.runReview.statuses[detail.status])}</span>
      </header>
      <p>{detail.output || t($ => $.runReview.noAdvisoryOutputIsRetainedInThisMetadata)}</p>
      {detail.failureSummary ? <details><summary>{t($ => $.runReview.originalDiagnostic)}</summary><p>{detail.failureSummary}</p></details> : null}
      <dl className="workflow-review-detail-facts">
        <div><dt>{t($ => $.runReview.input)}</dt><dd>{t($ => $.runReview.inputByteCount, { count: detail.inputBytes })}</dd></div>
        <div><dt>{t($ => $.runReview.recordedWindow)}</dt><dd>{formatDisplayDate(detail.startedAt, locale) ?? t($ => $.runReview.unknownTime)} → {formatDisplayDate(detail.completedAt, locale) ?? t($ => $.runReview.open)}</dd></div>
        <div><dt>{t($ => $.runReview.providerCalls)}</dt><dd>{detail.sideEffects.providerCalls}</dd></div>
        <div><dt>{t($ => $.runReview.forbiddenWrites)}</dt><dd>{forbiddenWrites}</dd></div>
        {detail.executionSourceId ? <div><dt>{t($ => $.runReview.executionSource)}</dt><dd>{detail.executionSourceKind} · {detail.executionSourceId} · v{detail.executionSourceVersion}</dd></div> : null}
        {detail.schemaVersion === "workflow_run_record.v8" ? <><div><dt>{t($ => $.runReview.inputContract)}</dt><dd>{detail.inputContractId} · {detail.inputContractDigest}</dd></div><div><dt>{t($ => $.runReview.inputFields)}</dt><dd>{detail.inputFields?.map((field) => `${field.name}:${field.valueType}`).join(" · ") || t($ => $.runReview.none)}</dd></div></> : null}
        <div><dt>{t($ => $.runReview.failureBoundary)}</dt><dd>{detail.diagnostic?.failureBoundary || t($ => $.runReview.none)}</dd></div>
      </dl>
      {detail.diagnostic ? (
        <section className="workflow-review-diagnostic">
          <div><strong>{detail.diagnostic.failureBoundary || t($ => $.runReview.noFailure)}</strong><span>{detail.diagnostic.summary}</span></div>
          <dl><div><dt>{t($ => $.runReview.failedNode)}</dt><dd>{detail.diagnostic.failedNodeId || t($ => $.runReview.none)}</dd></div><div><dt>{t($ => $.runReview.lastCompleted)}</dt><dd>{detail.diagnostic.lastCompletedNodeId || t($ => $.runReview.none)}</dd></div><div><dt>{t($ => $.runReview.reviewAction)}</dt><dd>{detail.diagnostic.recommendedReviewAction || t($ => $.runReview.none)}</dd></div></dl>
        </section>
      ) : null}
      <details className="workflow-review-disclosure">
        <summary><span>{t($ => $.runReview.nodeEvidence)}</span><small>{t($ => $.runReview.nodeCount, { count: detail.nodes.length })}</small></summary>
        <div className="workflow-review-node-list">
          {detail.nodes.map((node) => <div key={node.nodeId}><span><strong>{node.label}</strong><small>{node.nodeType}</small></span><span className={`status-${node.status}`}>{node.status}</span><small>{node.durationMs} ms</small></div>)}
        </div>
      </details>
      {detail.retrievalAttempt ? (
        <details className="workflow-review-disclosure">
          <summary><span>{t($ => $.runReview.retrievalEvidence)}</span><small>{detail.retrievalAttempt.selectedFragments.length} {t($ => $.runReview.selectedRefs)}</small></summary>
          <dl className="workflow-review-detail-facts"><div><dt>{t($ => $.runReview.snapshot)}</dt><dd>{detail.ragSnapshot?.snapshotId} · v{detail.ragSnapshot?.snapshotVersion}</dd></div><div><dt>{t($ => $.runReview.query)}</dt><dd>{detail.retrievalAttempt.queryBytes} bytes · {detail.retrievalAttempt.queryDigest}</dd></div><div><dt>{t($ => $.runReview.context)}</dt><dd>{detail.retrievalAttempt.contextBytes} bytes · {detail.retrievalAttempt.citationRefs.length} {t($ => $.runReview.citations)}</dd></div></dl>
          <button type="button" disabled={!canReadPreviews || retrievalPreviewPending} onClick={onLoadRetrievalPreviews}>{retrievalPreviewPending ? t($ => $.runReview.loadingPreviews) : t($ => $.runReview.readAuthorizedFragmentPreviews)}</button>
          {detail.retrievalFragmentPreviews.map((preview) => <blockquote key={preview.fragmentRef}><strong>{preview.fragmentRef}</strong><p>{preview.preview}</p></blockquote>)}
        </details>
      ) : null}
      <div className="workflow-review-reference-actions">
        <button type="button" onClick={() => onCopyReference("request", detail.requestId)}>{t($ => $.runReview.copyRequestId)}</button>
        <button type="button" onClick={() => onCopyReference("audit", detail.auditRef)}>{t($ => $.runReview.copyAuditRef)}</button>
        <span>{copiedRef ? t($ => $.runReview.referenceCopied, { reference: copiedRef }) : t($ => $.runReview.metadataReferencesOnly)}</span>
      </div>
      <p className="workflow-review-stop-line">{t($ => $.runReview.businessWritesAndReplayRemainLockedAt0)}</p>
    </article>
  );
}

function ComparisonSurface({ applicationId, runs, baselineRunId, candidateRunId, candidateRuns, selection, onBaselineChange, onCandidateChange, onCompare }: { applicationId: string; runs: WorkflowRunHistorySummary[]; baselineRunId: string; candidateRunId: string; candidateRuns: WorkflowRunHistorySummary[]; selection: { baseline: string; candidate: string } | null; onBaselineChange: (value: string) => void; onCandidateChange: (value: string) => void; onCompare: () => void }) {
  const { t } = useTranslation("evaluation");
  return (
    <div className="workflow-review-comparison-layout">
      <aside className="workflow-review-pair-rail">
        <div><p className="eyebrow">{t($ => $.runReview.exactRunPair)}</p><h5>{t($ => $.runReview.chooseCompatibleEvidence)}</h5><p>{t($ => $.runReview.compatibilityFollowsTheCurrentRunProfileAndImmutable)}</p></div>
        <label>{t($ => $.runReview.baseline)}<RunSelect value={baselineRunId} runs={runs} onChange={onBaselineChange} /></label>
        <label>{t($ => $.runReview.candidate)}<RunSelect value={candidateRunId} runs={candidateRuns} onChange={onCandidateChange} /></label>
        <button type="button" disabled={!baselineRunId || !candidateRunId || baselineRunId === candidateRunId} onClick={onCompare}>{t($ => $.runReview.compareExactRuns)}</button>
        <p className="workflow-review-stop-line">{t($ => $.runReview.instantaneousReadOnlyViewItIsNotPersisted)}</p>
      </aside>
      <main className="workflow-review-comparison-detail">
        {selection ? (
          <Suspense fallback={<p>{t($ => $.runReview.comparingDurableRunRecords)}</p>}>
            <WorkflowRunComparisonPanel applicationId={applicationId} baselineRunId={selection.baseline} candidateRunId={selection.candidate} config={config} />
          </Suspense>
        ) : (
          <div className="workflow-review-empty-detail"><span aria-hidden="true">⇄</span><h5>{t($ => $.runReview.chooseACompatiblePair)}</h5><p>{t($ => $.runReview.statusFindingsRetrievalAuthorityAndNodeDeltasWill)}</p></div>
        )}
      </main>
    </div>
  );
}

function RunSelect({ value, runs, onChange }: { value: string; runs: WorkflowRunHistorySummary[]; onChange: (value: string) => void }) {
  const { t } = useTranslation("evaluation");
  return <select value={value} onChange={(event) => onChange(event.target.value)}><option value="">{t($ => $.runReview.chooseExactRun)}</option>{runs.map((run) => <option key={run.runId} value={run.runId}>{run.runId} · {t($ => $.runReview.statuses[run.status])} · {t($ => $.runReview.sources[runSource(run)])}</option>)}</select>;
}

function OwnerBoundary({ title, copy }: { title: string; copy: string }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="workflow-review-owner-boundary" role="status">
      <div>
        <p className="eyebrow">{t($ => $.runReview.failClosedOwner)}</p>
        <h4>{title}</h4>
        <p>{copy}</p>
      </div>
      <dl>
        <div><dt>{t($ => $.runReview.projection)}</dt><dd>{t($ => $.runReview.separateOfflineEvidence)}</dd></div>
        <div><dt>{t($ => $.runReview.ownerRequests)}</dt><dd>0</dd></div>
        <div><dt>{t($ => $.runReview.mutation)}</dt><dd>{t($ => $.runReview.unavailable)}</dd></div>
      </dl>
      <ul>
        <li>{t($ => $.runReview.noCurrentApplicationRunListOrExactDetail)}</li>
        <li>{t($ => $.runReview.noComparisonEvaluationResultSuiteDecisionOrRelease)}</li>
        <li>{t($ => $.runReview.anExplicitDevTestSourceWithMatchingWorkspace)}</li>
      </ul>
      <strong>{t($ => $.runReview.noCurrentApplicationCapabilityIsInferredFromOffline)}</strong>
    </article>
  );
}

function runSource(run: WorkflowRunHistorySummary) {
  switch (run.schemaVersion) {
    case "workflow_run_record.v9": return "httpTool";
    case "workflow_run_record.v8": return "structuredWorkflow";
    case "workflow_run_record.v7": return "agent";
    case "workflow_run_record.v6": return "prompt";
    case "workflow_run_record.v5": return "workflow";
    case "workflow_run_record.v4": return "applicationRag";
    case "workflow_run_record.v3": return "workflowRag";
    default: return "savedDraft";
  }
}
