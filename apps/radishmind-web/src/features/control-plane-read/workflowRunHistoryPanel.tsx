import { workflowHistoryMessage, type WorkflowHistoryMessage } from "./workflowHistoryMessages.ts";
import "../../i18n/workflowDraftResources.ts";
import { workflowDraftStatusLabel } from "./workflowDraftMessages.ts";
import "../../i18n/workflowHistoryResources.ts";
import { useTranslation } from "react-i18next";
import { lazy, Suspense, useCallback, useEffect, useRef, useState } from "react";

import ActionSafetyReadPanel from "./ActionSafetyReadPanel.tsx";
import { startWorkflowDiagnosticDevRecord, type WorkflowRunDevFailureScenario } from "./workflowExecutorConsumer.ts";
import type { WorkflowRunRecord } from "./workflowRunRecordConsumer.ts";
import { readWorkflowRAGSnapshotConfig } from "./workflowRAGSnapshotConsumer.ts";
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

const config = readWorkflowRunHistoryConfig();
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
  const { t } = useTranslation("workflow");
  const [filter, setFilter] = useState<WorkflowRunHistoryFilter>(EMPTY_WORKFLOW_RUN_HISTORY_FILTER);
  const [history, setHistory] = useState(() => initialWorkflowRunHistoryState(config));
  const [detail, setDetail] = useState<WorkflowRunRecord | null>(null);
  const [selectedRunId, setSelectedRunId] = useState("");
  const [diagnosticScenario, setDiagnosticScenario] = useState<WorkflowRunDevFailureScenario>("gateway_timeout");
  const [diagnosticGenerationState, setDiagnosticGenerationState] = useState<WorkflowHistoryMessage | null>(null);
  const [copiedRef, setCopiedRef] = useState("");
  const [baselineRunId, setBaselineRunId] = useState("");
  const [candidateRunId, setCandidateRunId] = useState("");
  const [comparisonSelection, setComparisonSelection] = useState<{ baseline: string; candidate: string } | null>(null);
  const [retrievalPreviewPending, setRetrievalPreviewPending] = useState(false);
  const [handoffState, setHandoffState] = useState<WorkflowHistoryMessage | null>(null);
  const handledHandoffIdRef = useRef("");

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
    if (!handoffId || !handoffRunId || handledHandoffIdRef.current === handoffId) return;
    handledHandoffIdRef.current = handoffId;
    if (config.mode !== "dev_workflow_executor_http") {
      setHandoffState({ code: "handoffOffline", id: handoffRunId });
      onHandoffConsumed?.(handoffId);
      return;
    }
    setSelectedRunId(handoffRunId);
    setRetrievalPreviewPending(false);
    setHandoffState({ code: "handoffLoading", id: handoffRunId });
    void readWorkflowRunHistoryDetail({ runId: handoffRunId }, applicationId, config)
      .then((record) => {
        setDetail(record);
        setHandoffState({ code: record ? "handoffLoaded" : "handoffUnavailable", id: handoffRunId });
      })
      .catch(() => {
        setDetail(null);
        setHandoffState({ code: "handoffFailed", id: handoffRunId });
      })
      .finally(() => onHandoffConsumed?.(handoffId));
  }, [applicationId, handoffId, handoffRunId, onHandoffConsumed]);

  async function loadRetrievalPreviews() {
    const run = history.runs.find((item) => item.runId === selectedRunId);
    if (!run || run.schemaVersion !== "workflow_run_record.v3" || !ragConfig.scopes.has("workflow_rag_snapshots:read")) return;
    setRetrievalPreviewPending(true);
    try { setDetail(await readWorkflowRunHistoryDetail(run, applicationId, config, true)); }
    finally { setRetrievalPreviewPending(false); }
  }

  async function generateDiagnosticRun() {
    if (!filter.draftId.trim()) {
      setDiagnosticGenerationState({ code: "diagnosticNeedsDraft" });
      return;
    }
    setDiagnosticGenerationState({ code: "diagnosticWorking" });
    try {
      const state = await startWorkflowDiagnosticDevRecord(filter.draftId, applicationId, diagnosticScenario, config);
      setDetail(state.record);
      setSelectedRunId(state.record?.runId ?? "");
      setDiagnosticGenerationState(state.record ? { code: "diagnosticRecorded", scenario: diagnosticScenario } : { code: "diagnosticFailed", failureCode: state.failureCode ?? "workflow_run_unavailable" });
      await load();
    } catch {
      setDiagnosticGenerationState({ code: "diagnosticFailed", failureCode: "workflow_run_unavailable" });
    }
  }

  async function copyReference(label: string, value: string) {
    try {
      await navigator.clipboard.writeText(value);
      setCopiedRef(label);
    } catch {
      setCopiedRef("failed");
    }
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
  const hasDefinitionRuns = history.runs.some((run) => run.schemaVersion === "workflow_run_record.v5" || run.schemaVersion === "workflow_run_record.v9");
  const hasStructuredDefinitionRuns = history.runs.some((run) => run.schemaVersion === "workflow_run_record.v8");
  const hasPromptApplicationRuns = history.runs.some((run) => run.schemaVersion === "workflow_run_record.v6");
  const hasAgentCopilotRuns = history.runs.some((run) => run.schemaVersion === "workflow_run_record.v7");
  return (
    <section className="surface-band workspace-run-history" id="workspace-run-history" aria-labelledby="workspace-run-history-title">
      <div className="section-heading">
        <div><p className="eyebrow">{t($ => $.history.userWorkspace)}</p><h3 id="workspace-run-history-title">{t($ => $.history.runHistory)}</h3></div>
        <span className={`status-badge ${config.mode === "dev_workflow_executor_http" ? "status-good" : "status-neutral"}`}>{config.mode === "dev_workflow_executor_http" ? t($ => $.history.durable) : t($ => $.history.offlineSample)}</span>
      </div>
      {handoffState ? <p className="boundary-note" role="status">{workflowHistoryMessage(t, handoffState)}</p> : null}
      {config.mode !== "dev_workflow_executor_http" ? (
        <article className="run-history-route"><p className="eyebrow">{t($ => $.history.offlineMode)}</p><h4>{t($ => $.history.noLiveRunRequest)}</h4><p>{t($ => $.history.defaultOfflineModeKeepsTheHistoricalSampleSurfaceVisiblySeparate)}</p></article>
      ) : (
        <>
          <div className="run-history-summary">
            <article className="run-history-route">
              <p className="eyebrow">{t($ => $.history.realRunHistory)}</p><h4>/v1/user-workspace/workflow-runs</h4>
              <p className="route-path">{applicationId} · {workflowDraftStatusLabel(t, history.status)}</p>
              <dl className="tenant-meta"><div><dt>{t($ => $.history.records)}</dt><dd>{history.runs.length}</dd></div><div><dt>{t($ => $.history.failedCanceled)}</dt><dd>{failedCount} / {canceledCount}</dd></div><div><dt>{t($ => $.history.stale)}</dt><dd>{staleCount}</dd></div><div><dt>{t($ => $.history.gatewayStore)}</dt><dd>{gatewayCount} / {storeCount}</dd></div></dl>
            </article>
            <div className="run-history-metrics" aria-label={t($ => $.history.workflowRunHistoryFilters)}>
              <label>{t($ => $.history.status)}<select value={filter.status} onChange={(event) => setFilter({ ...filter, status: event.target.value as WorkflowRunHistoryFilter["status"] })}><option value="">{t($ => $.history.all)}</option><option value="succeeded">{t($ => $.history.succeeded)}</option><option value="failed">{t($ => $.history.failed)}</option><option value="outcome_unknown">{t($ => $.history.outcomeUnknown)}</option><option value="canceled">{t($ => $.history.canceled)}</option><option value="running">{t($ => $.history.running)}</option></select></label>
              <label>{t($ => $.history.draft)}<input value={filter.draftId} onChange={(event) => setFilter({ ...filter, draftId: event.target.value })} placeholder={t($ => $.history.draftId)} /></label>
              <label>{t($ => $.history.executionSource)}<select value={filter.executionSourceKind} onChange={(event) => setFilter({ ...filter, executionSourceKind: event.target.value as WorkflowRunHistoryFilter["executionSourceKind"] })}><option value="">{t($ => $.history.all)}</option><option value="workflow_draft">{t($ => $.history.savedDraft)}</option><option value="workflow_definition">{t($ => $.history.workflowDefinition)}</option><option value="application_configuration_draft">{t($ => $.history.applicationRAG)}</option><option value="prompt_application_template">{t($ => $.history.promptApplicationTemplate)}</option><option value="agent_copilot_profile">{t($ => $.history.agentCopilotProfile)}</option></select></label>
              <label>{t($ => $.history.sourceID)}<input value={filter.executionSourceId} onChange={(event) => setFilter({ ...filter, executionSourceId: event.target.value })} placeholder={t($ => $.history.exactSourceId)} /></label>
              <label>{t($ => $.history.sourceVersion)}<input type="number" min="1" value={filter.executionSourceVersion} onChange={(event) => setFilter({ ...filter, executionSourceVersion: event.target.value ? Number(event.target.value) : "" })} /></label>
              <label>{t($ => $.history.startedFrom)}<input type="datetime-local" value={filter.startedFrom} onChange={(event) => setFilter({ ...filter, startedFrom: event.target.value })} /></label>
              <label>{t($ => $.history.startedTo)}<input type="datetime-local" value={filter.startedTo} onChange={(event) => setFilter({ ...filter, startedTo: event.target.value })} /></label>
              <label>{t($ => $.history.failureCode)}<input value={filter.failureCode} onChange={(event) => setFilter({ ...filter, failureCode: event.target.value })} placeholder="workflow_run_…" /></label>
              <label>{t($ => $.history.boundary)}<select value={filter.failureBoundary} onChange={(event) => setFilter({ ...filter, failureBoundary: event.target.value as WorkflowRunHistoryFilter["failureBoundary"] })}><option value="">{t($ => $.history.all)}</option><option value="executor">{t($ => $.history.executor)}</option><option value="gateway">{t($ => $.history.gateway)}</option><option value="provider">{t($ => $.history.provider)}</option><option value="run_store">{t($ => $.history.runStore)}</option><option value="request">{t($ => $.history.request)}</option><option value="draft_read">{t($ => $.history.draftRead)}</option><option value="tool_policy">{t($ => $.history.toolPolicy)}</option><option value="tool_confirmation">{t($ => $.history.toolConfirmation)}</option><option value="tool_transport">{t($ => $.history.toolTransport)}</option><option value="tool_response">{t($ => $.history.toolResponse)}</option><option value="tool_store">{t($ => $.history.toolStore)}</option><option value="retrieval_policy">{t($ => $.history.retrievalPolicy)}</option><option value="retrieval_store">{t($ => $.history.retrievalStore)}</option><option value="retrieval_rank">{t($ => $.history.retrievalRank)}</option><option value="retrieval_context">{t($ => $.history.retrievalContext)}</option><option value="retrieval_citation">{t($ => $.history.retrievalCitation)}</option><option value="provider_selection">{t($ => $.history.providerSelection)}</option><option value="provider_call">{t($ => $.history.providerCall)}</option></select></label>
              <label>{t($ => $.history.provider)}<input value={filter.provider} onChange={(event) => setFilter({ ...filter, provider: event.target.value })} placeholder={t($ => $.history.exactProvider)} /></label>
              <label>{t($ => $.history.model)}<input value={filter.model} onChange={(event) => setFilter({ ...filter, model: event.target.value })} placeholder={t($ => $.history.exactModel)} /></label>
              <label>{t($ => $.history.staleRunning)}<select value={filter.staleRunning} onChange={(event) => setFilter({ ...filter, staleRunning: event.target.value as WorkflowRunHistoryFilter["staleRunning"] })}><option value="">{t($ => $.history.all)}</option><option value="true">{t($ => $.history.onlyStale)}</option><option value="false">{t($ => $.history.excludeStale)}</option></select></label>
              <button type="button" onClick={() => void load()} disabled={history.status === "loading"}>{t($ => $.history.applyFilters)}</button>
            </div>
          </div>
          {config.diagnosticsDevEnabled ? <div className="workflow-run-diagnostic-generator">
            <label>{t($ => $.history.devTestFailureScenario)}<select value={diagnosticScenario} onChange={(event) => setDiagnosticScenario(event.target.value as WorkflowRunDevFailureScenario)}><option value="gateway_timeout">{t($ => $.history.gatewayTimeout)}</option><option value="gateway_queue_full">{t($ => $.history.gatewayQueueFull)}</option><option value="gateway_worker_crash">{t($ => $.history.gatewayWorkerCrash)}</option><option value="gateway_protocol_failure">{t($ => $.history.gatewayProtocol)}</option><option value="provider_failed">{t($ => $.history.providerFailed)}</option><option value="output_unavailable">{t($ => $.history.outputUnavailable)}</option><option value="request_canceled">{t($ => $.history.requestCanceled)}</option><option value="run_store_unavailable">{t($ => $.history.storeUnavailable)}</option><option value="terminal_write_conflict">{t($ => $.history.terminalWriteConflict)}</option><option value="budget_exceeded">{t($ => $.history.budgetExceeded)}</option><option value="stale_running">{t($ => $.history.staleRunning)}</option></select></label>
            <button type="button" onClick={() => void generateDiagnosticRun()}>{t($ => $.history.generateDiagnosticRun)}</button>
            <p>{diagnosticGenerationState ? workflowHistoryMessage(t, diagnosticGenerationState) : t($ => $.history.diagnosticBoundary)}</p>
          </div> : null}
          {history.failureCode ? <p className="failure-summary">{history.failureCode}: {t($ => $.history.historyFailed)}</p> : null}
          <div className="workflow-run-history-live-list" aria-label={t($ => $.history.realWorkflowRunRecords)}>
            {history.runs.map((run) => <button type="button" className={`workflow-run-history-live-row ${selectedRunId === run.runId ? "is-selected" : ""}`} key={run.runId} onClick={() => void selectRun(run)}><span className="workflow-run-history-live-identity"><strong>{run.runId}</strong><small>{run.schemaVersion === "workflow_run_record.v4" || run.schemaVersion === "workflow_run_record.v5" || run.schemaVersion === "workflow_run_record.v6" || run.schemaVersion === "workflow_run_record.v7" || run.schemaVersion === "workflow_run_record.v8" || run.schemaVersion === "workflow_run_record.v9" ? t($ => $.history.sourceVersion, { id: run.executionSourceId, version: run.executionSourceVersion }) : t($ => $.history.sourceVersion, { id: run.draftId, version: run.draftVersion })}</small></span><span><small>{t($ => $.history.status)}</small><strong>{workflowDraftStatusLabel(t, run.status)}{run.staleRunning ? ` · ${workflowDraftStatusLabel(t, "stale")}` : ""}</strong><small>{run.schemaVersion}</small></span><span><small>{t($ => $.history.failure)}</small><strong>{run.failureBoundary || workflowDraftStatusLabel(t, "none")}</strong><small>{run.retrievalFailureCategory || run.toolFailureCategory || run.gatewayFailureCategory || run.failureCode || workflowDraftStatusLabel(t, "none")}</small></span><span><small>{t($ => $.history.controlledEffects)}</small><strong>{run.schemaVersion === "workflow_run_record.v3" || run.schemaVersion === "workflow_run_record.v4" ? t($ => $.history.retrievalEffects, { retrieval: run.sideEffects.retrievalCalls, provider: run.sideEffects.providerCalls }) : run.schemaVersion === "workflow_run_record.v5" || run.schemaVersion === "workflow_run_record.v6" || run.schemaVersion === "workflow_run_record.v7" || run.schemaVersion === "workflow_run_record.v8" ? t($ => $.history.providerEffects, { provider: run.sideEffects.providerCalls }) : t($ => $.history.toolEffects, { tool: run.sideEffects.toolCalls, confirmation: run.sideEffects.confirmationCalls })}</strong><small>{run.schemaVersion === "workflow_run_record.v3" || run.schemaVersion === "workflow_run_record.v4" ? t($ => $.history.citationCount, { count: run.citationRefs.length }) : run.schemaVersion === "workflow_run_record.v5" || run.schemaVersion === "workflow_run_record.v6" || run.schemaVersion === "workflow_run_record.v7" || run.schemaVersion === "workflow_run_record.v8" ? run.executionProfile : run.toolAttemptStatus || t($ => $.history.noTool)}</small></span></button>)}
          </div>
          <div className="workflow-run-comparison-selector" aria-label={t($ => $.history.workflowRunComparisonSelection)}>
            <label>{t($ => $.history.baselineRun)}<select value={baselineRunId} onChange={(event) => { setBaselineRunId(event.target.value); setCandidateRunId(""); setComparisonSelection(null); }}><option value="">{t($ => $.history.chooseBaseline)}</option>{comparisonRuns.map((run) => <option value={run.runId} key={`baseline-${run.runId}`}>{run.runId} · {workflowDraftStatusLabel(t, run.status)} · {run.schemaVersion === "workflow_run_record.v8" ? t($ => $.history.structuredDefinition) : run.schemaVersion === "workflow_run_record.v7" ? t($ => $.history.agentCopilot) : run.schemaVersion === "workflow_run_record.v6" ? t($ => $.history.promptApplication) : run.schemaVersion === "workflow_run_record.v5" ? t($ => $.history.definitionAuthority) : run.schemaVersion === "workflow_run_record.v4" ? t($ => $.history.applicationRAG) : run.schemaVersion === "workflow_run_record.v3" ? t($ => $.history.workflowRag) : t($ => $.history.standard)}</option>)}</select></label>
            <label>{t($ => $.history.candidateRun)}<select value={candidateRunId} onChange={(event) => setCandidateRunId(event.target.value)}><option value="">{t($ => $.history.chooseCompatibleCandidate)}</option>{comparisonCandidateRuns.map((run) => <option value={run.runId} key={`candidate-${run.runId}`}>{run.runId} · {workflowDraftStatusLabel(t, run.status)}</option>)}</select></label>
            <button type="button" disabled={!baselineRunId || !candidateRunId || baselineRunId === candidateRunId} onClick={() => setComparisonSelection({ baseline: baselineRunId, candidate: candidateRunId })}>{t($ => $.history.compareRuns)}</button>
          </div>
          {hasRetrievalRuns ? <p className="boundary-note"><code>workflow_rag_retrieval.v1</code>{" "}{t($ => $.history.and)}{" "}<code>workflow_rag_application_invocation.v1</code>{" "}{t($ => $.history.areReviewedSeparatelyCompareV4OnlyWithV4FromThe)}</p> : null}
          {hasDefinitionRuns ? <p className="boundary-note"><code>workflow_definition_executor.v1</code>{" "}{t($ => $.history.comparesV5RecordsFromTheSameDefinitionLineageWithoutExecution)}</p> : null}
          {hasStructuredDefinitionRuns ? <p className="boundary-note"><code>workflow_definition_executor.v2</code>{" "}{t($ => $.history.comparesV8MetadataForTheSameDefinitionAndInputContract)}</p> : null}
          {hasPromptApplicationRuns ? <p className="boundary-note"><code>prompt_application_invocation_v1</code>{" "}{t($ => $.history.comparesV6MetadataFromTheSamePromptTemplateLineageWithout)}</p> : null}
          {hasAgentCopilotRuns ? <p className="boundary-note"><code>agent_copilot_suggestion_v1</code>{" "}{t($ => $.history.comparesV7MetadataForTheSameProfileProjectAndTask)}</p> : null}
          {comparisonSelection ? <Suspense fallback={<p>{t($ => $.history.loadingRegressionReview)}</p>}><WorkflowRunComparisonPanel applicationId={applicationId} baselineRunId={comparisonSelection.baseline} candidateRunId={comparisonSelection.candidate} config={config} /></Suspense> : null}
          <Suspense fallback={<p>{t($ => $.history.loadingEvaluationCases)}</p>}><WorkflowEvaluationPanel applicationId={applicationId} runs={comparisonRuns} config={config} /></Suspense>
          <Suspense fallback={<p>{t($ => $.history.loadingEvaluationSuites)}</p>}><WorkflowEvaluationSuitePanel applicationId={applicationId} config={config} /></Suspense>
          {history.hasMore ? <button type="button" onClick={() => void load(history.nextCursor, true)} disabled={history.status === "loading"}>{t($ => $.history.loadEarlierRuns)}</button> : null}
          {detail?.schemaVersion === "workflow_run_record.v3" || detail?.schemaVersion === "workflow_run_record.v4" ? (
            <WorkflowRAGRunHistoryEvidence
              detail={detail}
              canReadPreviews={detail.schemaVersion === "workflow_run_record.v3" && ragConfig.mode === "dev_workflow_rag_http" && ragConfig.scopes.has("workflow_rag_snapshots:read")}
              previewPending={retrievalPreviewPending}
              onLoadPreviews={loadRetrievalPreviews}
            />
          ) : null}
          <ActionSafetyReadPanel projection={detail?.actionSafety ?? null} title={t($ => $.history.durableRunSafety)} />
          {detail ? <article className="workflow-run-detail"><div className="card-title-row"><div><p className="eyebrow">{t($ => $.history.realRunDetail)}</p><h4>{detail.runId}</h4></div><span className={`status-badge ${detail.status === "succeeded" ? "status-good" : detail.status === "outcome_unknown" ? "status-neutral" : "status-bad"}`}>{workflowDraftStatusLabel(t, detail.status)}</span></div><p>{detail.output || detail.failureSummary || t($ => $.history.noOutput)}</p><dl className="tenant-meta"><div><dt>{t($ => $.history.input)}</dt><dd>{detail.inputBytes}{t($ => $.history.bytesRawTextNotRetained)}</dd></div><div><dt>{t($ => $.history.providerCalls)}</dt><dd>{detail.sideEffects.providerCalls}</dd></div><div><dt>{t($ => $.history.controlledEffects)}</dt><dd>{detail.schemaVersion === "workflow_run_record.v3" || detail.schemaVersion === "workflow_run_record.v4" ? t($ => $.history.retrievalEffects, { retrieval: detail.sideEffects.retrievalCalls, provider: detail.sideEffects.providerCalls }) : detail.schemaVersion === "workflow_run_record.v5" || detail.schemaVersion === "workflow_run_record.v6" || detail.schemaVersion === "workflow_run_record.v7" || detail.schemaVersion === "workflow_run_record.v8" ? t($ => $.history.providerEffects, { provider: detail.sideEffects.providerCalls }) : t($ => $.history.toolEffects, { tool: detail.sideEffects.toolCalls, confirmation: detail.sideEffects.confirmationCalls })}</dd></div><div><dt>{t($ => $.history.forbiddenWrites)}</dt><dd>{forbiddenWrites}</dd></div>{detail.schemaVersion === "workflow_run_record.v4" || detail.schemaVersion === "workflow_run_record.v5" || detail.schemaVersion === "workflow_run_record.v6" || detail.schemaVersion === "workflow_run_record.v7" || detail.schemaVersion === "workflow_run_record.v8" || detail.schemaVersion === "workflow_run_record.v9" ? <div><dt>{t($ => $.history.executionSource)}</dt><dd>{detail.executionSourceKind} · {detail.executionSourceId} · v{detail.executionSourceVersion}</dd></div> : null}{detail.schemaVersion === "workflow_run_record.v8" ? <><div><dt>{t($ => $.history.inputContract)}</dt><dd>{detail.inputContractId} · {detail.inputContractDigest}</dd></div><div><dt>{t($ => $.history.inputFields)}</dt><dd>{detail.inputFields?.map((field) => `${field.name}:${field.valueType}`).join(" · ") || workflowDraftStatusLabel(t, "none")}</dd></div></> : null}{detail.agentCopilotAuthority ? <><div><dt>{t($ => $.history.agentAuthority)}</dt><dd>{detail.agentCopilotAuthority.assignmentId} · v{detail.agentCopilotAuthority.assignmentVersion}</dd></div><div><dt>{t($ => $.history.profile)}</dt><dd>{detail.agentCopilotAuthority.profileId} · v{detail.agentCopilotAuthority.profileVersion}</dd></div><div><dt>{t($ => $.history.projectTask)}</dt><dd>{detail.agentProject} · {detail.agentTask} · {detail.agentLocale}</dd></div><div><dt>{t($ => $.history.responseMetadata)}</dt><dd>{detail.agentResponseStatus} · {detail.agentActionCount}{t($ => $.history.actionsRisk)}{detail.agentRiskLevel}{t($ => $.history.confirmation)}{String(detail.agentRequiresConfirmation)}</dd></div></> : null}{detail.promptApplicationAuthority ? <><div><dt>{t($ => $.history.runtimeAssignment)}</dt><dd>{detail.promptApplicationAuthority.assignmentId} · v{detail.promptApplicationAuthority.assignmentVersion}</dd></div><div><dt>{t($ => $.history.publishReview)}</dt><dd>{detail.promptApplicationAuthority.publishCandidateId} · v{detail.promptApplicationAuthority.publishReviewVersion}</dd></div><div><dt>{t($ => $.history.promptAuthority)}</dt><dd>{detail.promptApplicationAuthority.templateId} · v{detail.promptApplicationAuthority.templateVersion}</dd></div><div><dt>{t($ => $.history.protocolUsage)}</dt><dd>{detail.requestedProtocol} → {detail.selectedProtocol} · {detail.promptUsage?.state ?? workflowDraftStatusLabel(t, "unavailable")}</dd></div></> : null}{detail.definitionAuthority ? <><div><dt>{t($ => $.history.definitionAuthority)}</dt><dd>{detail.definitionAuthority.definitionId} · v{detail.definitionAuthority.definitionVersion}{t($ => $.history.pointerV)}{detail.definitionAuthority.activationPointerVersion}</dd></div><div><dt>{t($ => $.history.sourceDraftProvenance)}</dt><dd>{detail.definitionAuthority.sourceDraftId} · v{detail.definitionAuthority.sourceDraftVersion}</dd></div></> : null}{detail.planId ? <div><dt>{t($ => $.history.planConfirmation)}</dt><dd>{detail.planId} · {detail.confirmationId}</dd></div> : null}{detail.toolAttempt ? <div><dt>{t($ => $.history.toolAttempt)}</dt><dd>{detail.toolAttempt.attemptId} · {detail.toolAttempt.status}</dd></div> : null}</dl>{detail.diagnostic ? <div className="workflow-run-diagnostic-review"><p className="eyebrow">{t($ => $.history.structuredFailureReview)}</p><h5>{detail.diagnostic.failureBoundary || t($ => $.history.noFailure)} · {detail.diagnostic.retrievalFailureCategory !== "none" ? detail.diagnostic.retrievalFailureCategory : detail.diagnostic.toolFailureCategory !== "none" ? detail.diagnostic.toolFailureCategory : detail.diagnostic.gatewayFailureCategory}</h5><p>{detail.diagnostic.summary || t($ => $.history.completedDiagnostic)}</p><dl className="tenant-meta"><div><dt>{t($ => $.history.failedNode)}</dt><dd>{detail.diagnostic.failedNodeId || workflowDraftStatusLabel(t, "none")}</dd></div><div><dt>{t($ => $.history.lastCompleted)}</dt><dd>{detail.diagnostic.lastCompletedNodeId || workflowDraftStatusLabel(t, "none")}</dd></div><div><dt>{t($ => $.history.reviewAction)}</dt><dd>{detail.diagnostic.recommendedReviewAction || workflowDraftStatusLabel(t, "none")}</dd></div><div><dt>{t($ => $.history.terminalWrite)}</dt><dd>{detail.diagnostic.terminalWriteState}</dd></div></dl></div> : <p className="boundary-note">{detail.schemaVersion === "workflow_run_record.v0" ? t($ => $.history.legacyDiagnostic) : t($ => $.history.noStructuredFailure)}</p>}{detail.status === "outcome_unknown" ? <p className="failure-summary">{t($ => $.history.theProviderOutcomeIsUncertainRetryAndResumeAreDisabled)}</p> : null}<div className="workflow-run-reference-actions"><button type="button" onClick={() => void copyReference("request", detail.requestId)}>{t($ => $.history.copyRequestId)}</button><button type="button" onClick={() => void copyReference("audit", detail.auditRef)}>{t($ => $.history.copyAuditRef)}</button><span>{copiedRef === "failed" ? t($ => $.history.copyFailed) : copiedRef ? t($ => $.history.copied, { reference: copiedRef === "request" ? t($ => $.history.request) : t($ => $.history.audit) }) : t($ => $.history.referencesOnly)}</span></div><div className="workflow-run-history-node-list">{detail.nodes.map((node) => <div className={`workflow-run-history-live-row ${detail.diagnostic?.failedNodeId === node.nodeId ? "is-failed" : ""} ${detail.diagnostic?.lastCompletedNodeId === node.nodeId ? "is-last-completed" : ""}`} key={node.nodeId}><span><strong>{node.label}</strong><small>{workflowDraftStatusLabel(t, node.nodeType)}</small></span><span><small>{t($ => $.history.status)}</small><strong>{workflowDraftStatusLabel(t, node.status)}</strong></span><span><small>{t($ => $.history.duration)}</small><strong>{node.durationMs} ms</strong></span><p>{node.outputPreview}</p></div>)}</div><p className="boundary-note">{t($ => $.history.businessWritesAndReplayRemainLockedAt0ToolAnd)}</p></article> : null}
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
  const { t } = useTranslation("workflow");
  const snapshot = detail.ragSnapshot;
  const attempt = detail.retrievalAttempt;
  const authority = detail.ragApplicationAuthority;
  if (!snapshot || !attempt) return null;
  return (
    <article className="workflow-run-detail workflow-rag-run-history-evidence" aria-label={t($ => $.history.ragEvidence, { schema: detail.schemaVersion })}>
      <div className="card-title-row"><div><p className="eyebrow">{t($ => $.history.metadataOnly)}{detail.schemaVersion}</p><h4>{snapshot.ragRef}</h4></div><span className="status-badge neutral">{workflowDraftStatusLabel(t, attempt.status)}</span></div>
      {authority ? <dl className="tenant-meta">
        <div><dt>{t($ => $.history.runtimeAssignment)}</dt><dd>{authority.assignmentId} · v{authority.assignmentVersion}</dd></div>
        <div><dt>{t($ => $.history.publishCandidate)}</dt><dd>{authority.publishCandidateId}{t($ => $.history.reviewV)}{authority.publishReviewVersion}</dd></div>
        <div><dt>{t($ => $.history.applicationDraft)}</dt><dd>{authority.draftId} · v{authority.draftVersion} · {authority.draftDigest}</dd></div>
        <div><dt>{t($ => $.history.binding)}</dt><dd>{authority.bindingId} · v{authority.bindingVersion} · {authority.bindingDigest}</dd></div>
        <div><dt>{t($ => $.history.datasetReview)}</dt><dd>{authority.datasetId} · v{authority.datasetVersion} · {authority.candidateReviewId}</dd></div>
        <div><dt>{t($ => $.history.configuredRoute)}</dt><dd>{authority.configuredProtocol} · {authority.configuredModel}</dd></div>
      </dl> : null}
      <dl className="tenant-meta">
        <div><dt>{authority ? t($ => $.history.executionSource) : t($ => $.history.draft)}</dt><dd>{authority ? `${detail.executionSourceId} · v${detail.executionSourceVersion}` : `v${detail.draftVersion} · ${detail.draftDigest}`}</dd></div>
        <div><dt>{t($ => $.history.snapshot)}</dt><dd>{snapshot.snapshotId} · v{snapshot.snapshotVersion} · {snapshot.snapshotDigest}</dd></div>
        <div><dt>{t($ => $.history.profile)}</dt><dd>{attempt.profileId} · v{attempt.profileVersion} · {attempt.profileDigest}</dd></div>
        <div><dt>{t($ => $.history.query)}</dt><dd>{attempt.queryBytes}{t($ => $.history.bytes)}{attempt.queryDigest}</dd></div>
        <div><dt>{t($ => $.history.retrieval)}</dt><dd>{attempt.candidateCount}{t($ => $.history.candidates)}{attempt.selectedFragments.length}{t($ => $.history.selected)}{attempt.retrievalLatencyMs} ms</dd></div>
        <div><dt>{t($ => $.history.context)}</dt><dd>{attempt.contextBytes}{t($ => $.history.bytes)}{attempt.citationRefs.length}{t($ => $.history.citationRefs)}</dd></div>
      </dl>
      <div className="workflow-run-history-node-list">
        {attempt.selectedFragments.map((fragment) => <div className="workflow-run-history-node-row" key={fragment.fragmentRef}><span><strong>#{fragment.rank} · {fragment.fragmentRef}</strong><small>{fragment.sourceType}{fragment.isOfficial ? ` · ${t($ => $.history.official)}` : ""}</small></span><span><small>{t($ => $.history.digest)}</small><code>{fragment.contentDigest}</code></span><span><small>{t($ => $.history.cited)}</small><strong>{attempt.citationRefs.includes(fragment.fragmentRef) ? t($ => $.history.yes) : t($ => $.history.no)}</strong></span></div>)}
      </div>
      <div className="workflow-run-reference-actions"><button type="button" disabled={!canReadPreviews || previewPending} onClick={() => void onLoadPreviews()}>{previewPending ? t($ => $.history.previewsLoading) : t($ => $.history.readPreviews)}</button><span>{canReadPreviews ? t($ => $.history.previewBoundary) : authority ? t($ => $.history.previewDisabled) : t($ => $.history.previewPermission)}</span></div>
      {detail.retrievalFragmentPreviews.map((preview) => <blockquote key={preview.fragmentRef}><strong>{preview.fragmentRef}</strong><p>{preview.preview}</p><small>{preview.truncated ? t($ => $.history.previewTruncated) : t($ => $.history.previewComplete)}</small></blockquote>)}
      <p className="boundary-note">{t($ => $.history.runRecordsDoNotContainRawQueriesFullFragmentsPrompt)}</p>
    </article>
  );
}
