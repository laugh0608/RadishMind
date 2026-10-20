import "../../i18n/operationsResources.ts";
import "../../i18n/gatewayReviewResources.ts";
import { useTranslation } from "react-i18next";
import type { TFunction } from "i18next";
import { useLocalePreference } from "../../i18n/LocaleProvider.tsx";
import { formatDisplayDate, formatDisplayNumber, formatMicroUSD } from "../../i18n/formatters.ts";
import { gatewayReviewState } from "./gatewayReviewMessages.ts";
import { useEffect, useState } from "react";

import {
  initialApplicationOperationsState,
  loadApplicationOperations,
  type ApplicationOperationsState,
  type ApplicationOperationsTimelineEntry,
} from "./applicationOperationsConsumer.ts";
import type { ApplicationDevelopmentOwnerEvidence } from "./applicationDevelopmentReadiness.ts";
import { readModelGatewayRequestHistoryConfig } from "./modelGatewayRequestHistoryConsumer.ts";
import { readWorkflowExecutorConsumerConfig } from "./workflowExecutorConsumer.ts";

const gatewayConfig = readModelGatewayRequestHistoryConfig();
const workflowConfig = readWorkflowExecutorConsumerConfig();

export default function ApplicationOperationsPanel({
  applicationId,
  applicationName,
  workspaceId,
  active,
  onEvidenceChange,
  onOpenGatewayRequest,
  onOpenWorkflowRun,
}: {
  applicationId: string;
  applicationName: string;
  workspaceId: string;
  active: boolean;
  onEvidenceChange?: (evidence: ApplicationDevelopmentOwnerEvidence) => void;
  onOpenGatewayRequest?: (requestId: string, consumerRef: string) => void;
  onOpenWorkflowRun?: (runId: string) => void;
}) {
  const { t } = useTranslation("gateway");
  const { locale } = useLocalePreference();
  const number = (value: number) => formatDisplayNumber(value, locale) ?? t($ => $.review.states.unavailable);
  const [refreshKey, setRefreshKey] = useState(0);
  const [state, setState] = useState<ApplicationOperationsState>(() =>
    initialApplicationOperationsState(applicationId, gatewayConfig, workflowConfig)
  );
  const workspaceScopeFailureSummary = [
    gatewayConfig.mode === "dev_gateway_request_history_http" && gatewayConfig.workspaceId !== workspaceId.trim()
      ? `Gateway source is configured for ${gatewayConfig.workspaceId}`
      : "",
    workflowConfig.mode === "dev_workflow_executor_http" && workflowConfig.workspaceId !== workspaceId.trim()
      ? `Workflow source is configured for ${workflowConfig.workspaceId}`
      : "",
  ].filter(Boolean).join(". ");
  const workspaceScopeMatches = Boolean(workspaceId.trim()) && !workspaceScopeFailureSummary;

  useEffect(() => {
    setState(initialApplicationOperationsState(applicationId, gatewayConfig, workflowConfig));
  }, [applicationId, workspaceId]);

  useEffect(() => {
    let cancelled = false;
    if (!active || !workspaceScopeMatches) return () => { cancelled = true; };
    setState(initialApplicationOperationsState(applicationId, gatewayConfig, workflowConfig));
    void loadApplicationOperations(applicationId, gatewayConfig, workflowConfig).then((nextState) => {
      if (!cancelled) setState(nextState);
    });
    return () => {
      cancelled = true;
    };
  }, [active, applicationId, refreshKey, workspaceScopeMatches]);

  const metrics = state.metrics;
  const refreshDisabled = !active || !workspaceScopeMatches || state.status === "loading" || state.status === "offline" ||
    state.status === "application_unavailable";

  useEffect(() => {
    if (!onEvidenceChange) return;
    const failureCodes = [state.gateway.failureCode, state.workflow.failureCode].filter(Boolean);
    const blocked = !workspaceScopeMatches || state.status === "failed" || state.status === "application_unavailable";
    const partialFailure = state.status === "partial_failure";
    const available = state.status === "ready" && state.loadedWindowComplete;
    const evidenceRefs = [
      state.gateway.requestId ? { kind: "request" as const, id: state.gateway.requestId } : null,
      state.workflow.requestId ? { kind: "request" as const, id: state.workflow.requestId } : null,
    ].filter((ref): ref is { kind: "request"; id: string } => Boolean(ref));
    onEvidenceChange({
      contributionId: "operations_coverage",
      status: blocked ? "blocked" : partialFailure ? "partial_failure" : available ? "available" : "incomplete",
      coverage: available ? "complete" : evidenceRefs.length ? "partial" : blocked ? "complete" : "none",
      evidenceRefs,
      missingEvidence: available ? [] : [workspaceScopeFailureSummary || state.failureSummary || "Load complete Gateway and Workflow operations windows."],
      blockers: blocked || partialFailure ? [{
        code: !workspaceScopeMatches ? "application_operations_workspace_scope_mismatch" : failureCodes[0] || (partialFailure ? "application_operations_partial_failure" : "application_operations_unavailable"),
        summary: workspaceScopeFailureSummary || state.failureSummary || "Application operations evidence is unavailable.",
      }] : [],
      failureCodes: !workspaceScopeMatches ? ["application_operations_workspace_scope_mismatch"] : failureCodes,
    });
  }, [onEvidenceChange, state, workspaceScopeFailureSummary, workspaceScopeMatches]);

  return (
    <section className="surface-band application-operations" id="application-operations" aria-labelledby="application-operations-title">
      <div className="section-heading">
        <div>
          <p className="eyebrow">{t($ => $.operations.eyebrow)}</p>
          <h3 id="application-operations-title">{t($ => $.operations.heading)}</h3>
          <p>
            {applicationName || t($ => $.operations.noApplication)} · <code>{state.applicationId || t($ => $.operations.applicationUnavailable)}</code> · {t($ => $.operations.consumer, { consumer: gatewayConfig.consumerRef })}
          </p>
        </div>
        <div className="application-operations-actions">
          <StatusBadge status={state.status} />
          <button type="button" className="secondary-action" disabled={refreshDisabled} onClick={() => setRefreshKey((key) => key + 1)}>
            {t($ => $.operations.refresh)}
          </button>
        </div>
      </div>

      {!workspaceScopeMatches ? (
        <p className="application-operations-failure" role="alert">
          {t($ => $.operations.workspaceBoundary, { gateway: gatewayConfig.workspaceId, workflow: workflowConfig.workspaceId, workspace: workspaceId || t($ => $.operations.unavailable) })}
        </p>
      ) : null}

      <div className="application-operations-coverage" aria-label={t($ => $.operations.sourceCoverage)}>
        <ChannelCoverage
          label={t($ => $.operations.requests)}
          status={state.gateway.status}
          loaded={metrics.gatewayLoaded}
          hasMore={state.gateway.hasMore}
          requestId={state.gateway.requestId}
          auditRef={state.gateway.auditRef}
          failureCode={state.gateway.failureCode}
        />
        <ChannelCoverage
          label={t($ => $.operations.runs)}
          status={state.workflow.status}
          loaded={metrics.workflowLoaded}
          hasMore={state.workflow.hasMore}
          requestId={state.workflow.requestId}
          auditRef={state.workflow.auditRef}
          failureCode={state.workflow.failureCode}
        />
      </div>

      {state.status === "failed" || state.status === "partial_failure" ? <p className="application-operations-failure" role="alert">{t($ => $.operations.sourceFailure)}</p> : null}

      <div className="application-operations-metrics" aria-label={t($ => $.operations.attribution)}>
        <MetricCard
          label={t($ => $.operations.gatewayStatus)}
          value={t($ => $.operations.succeededCount, { displayCount: number(metrics.gatewaySucceeded) })}
          detail={t($ => $.operations.gatewayCounts, { failed: number(metrics.gatewayFailed), canceled: number(metrics.gatewayCanceled), started: number(metrics.gatewayStarted) })}
        />
        <MetricCard
          label={t($ => $.operations.gatewayUsage)}
          value={t($ => $.operations.reportedCount, { displayCount: number(metrics.gatewayUsageReported) })}
          detail={t($ => $.operations.usageCounts, { missing: number(metrics.gatewayUsageNotReported), na: number(metrics.gatewayUsageNotApplicable) })}
        />
        <MetricCard
          label={t($ => $.operations.tokens)}
          value={t($ => $.operations.totalCount, { displayCount: number(metrics.gatewayTotalTokens) })}
          detail={t($ => $.operations.tokenCounts, { input: number(metrics.gatewayInputTokens), output: number(metrics.gatewayOutputTokens) })}
        />
        <MetricCard
          label={t($ => $.operations.estimate)}
          value={formatMicroUSD(metrics.gatewayEstimatedCostMicros, locale) ?? t($ => $.operations.unavailable)}
          detail={t($ => $.operations.estimateCounts, { estimated: number(metrics.gatewayCostEstimated), partial: number(metrics.gatewayCostPartial), window: state.gateway.hasMore ? t($ => $.operations.partialWindow) : t($ => $.operations.completeWindows) })}
        />
        <MetricCard
          label={t($ => $.operations.workflowStatus)}
          value={t($ => $.operations.succeededCount, { displayCount: number(metrics.workflowSucceeded) })}
          detail={t($ => $.operations.workflowCounts, { failed: number(metrics.workflowFailed), canceled: number(metrics.workflowCanceled), running: number(metrics.workflowRunning), unknown: number(metrics.workflowOutcomeUnknown) })}
        />
        <MetricCard
          label={t($ => $.operations.workflowCalls)}
          value={t($ => $.operations.observedCalls, { provider: number(metrics.workflowProviderCalls), retrieval: number(metrics.workflowRetrievalCalls) })}
          detail={t($ => $.operations.otherCalls, { tool: number(metrics.workflowToolCalls), confirmation: number(metrics.workflowConfirmationCalls) })}
        />
      </div>

      <div className="application-operations-cost-coverage" aria-label={t($ => $.operations.costCoverage)}>
        <p className="eyebrow">{t($ => $.operations.costWindow)}</p>
        <div>
          <CostCoverage label={t($ => $.operations.estimated)} value={metrics.gatewayCostEstimated} tone="ready" />
          <CostCoverage label={t($ => $.operations.partialAttempts)} value={metrics.gatewayCostPartial} tone="attention" />
          <CostCoverage label={t($ => $.operations.usageMissing)} value={metrics.gatewayCostUsageNotReported} tone="neutral" />
          <CostCoverage label={t($ => $.operations.priceMissing)} value={metrics.gatewayCostPriceNotConfigured} tone="attention" />
          <CostCoverage label={t($ => $.operations.priceUnavailable)} value={metrics.gatewayCostPriceUnavailable} tone="blocked" />
          <CostCoverage label={t($ => $.operations.notApplicable)} value={metrics.gatewayCostNotApplicable} tone="neutral" />
          <CostCoverage label={t($ => $.operations.legacy)} value={metrics.gatewayCostLegacyNotCaptured} tone="neutral" />
        </div>
        <p>{state.gateway.hasMore ? t($ => $.operations.hasMore) : t($ => $.operations.completeGateway)}</p>
      </div>

      {(metrics.workflowBusinessWrites > 0 || metrics.workflowReplayWrites > 0) && (
        <p className="application-operations-stop-line" role="alert">
          {t($ => $.operations.stopLine, { business: number(metrics.workflowBusinessWrites), replay: number(metrics.workflowReplayWrites) })}
        </p>
      )}

      <div className="application-operations-timeline-heading">
        <div>
          <p className="eyebrow">{t($ => $.operations.timeline)}</p>
          <h4>{t($ => $.operations.independentRecords)}</h4>
        </div>
        <span>{state.loadedWindowComplete ? t($ => $.operations.completeWindows) : t($ => $.operations.moreAvailable)}</span>
      </div>

      {state.timeline.length > 0 ? (
        <ol className="application-operations-timeline">
          {state.timeline.map((entry) => (
            <TimelineEntry
              key={`${entry.source}:${entry.recordId}`}
              entry={entry}
              onOpenGatewayRequest={onOpenGatewayRequest}
              onOpenWorkflowRun={onOpenWorkflowRun}
            />
          ))}
        </ol>
      ) : (
        <p className="application-operations-empty">
          {emptyMessage(t, state)}
        </p>
      )}

      <p className="boundary-note">
        {t($ => $.operations.boundary)}
      </p>
    </section>
  );
}

function ChannelCoverage({
  label,
  status,
  loaded,
  hasMore,
  requestId,
  auditRef,
  failureCode,
}: {
  label: string;
  status: string;
  loaded: number;
  hasMore: boolean;
  requestId: string;
  auditRef: string;
  failureCode: string;
}) {
  const { t } = useTranslation("gateway");
  const { locale } = useLocalePreference();
  const number = (value: number) => formatDisplayNumber(value, locale) ?? t($ => $.review.states.unavailable);
  return (
    <article>
      <div className="card-title-row"><h4>{label}</h4><span>{gatewayReviewState(t, status)}</span></div>
      <p>{t($ => $.operations.loaded, { displayCount: number(loaded), window: hasMore ? t($ => $.operations.moreAvailable) : t($ => $.operations.completeWindows) })}</p>
      <dl>
        <div><dt>{t($ => $.operations.request)}</dt><dd>{requestId}</dd></div>
        <div><dt>{t($ => $.operations.audit)}</dt><dd>{auditRef}</dd></div>
        <div><dt>{t($ => $.operations.failure)}</dt><dd>{failureCode || t($ => $.operations.none)}</dd></div>
      </dl>
    </article>
  );
}

function MetricCard({ label, value, detail }: { label: string; value: string; detail: string }) {
  return <article><p>{label}</p><strong>{value}</strong><span>{detail}</span></article>;
}

function TimelineEntry({
  entry,
  onOpenGatewayRequest,
  onOpenWorkflowRun,
}: {
  entry: ApplicationOperationsTimelineEntry;
  onOpenGatewayRequest?: (requestId: string, consumerRef: string) => void;
  onOpenWorkflowRun?: (runId: string) => void;
}) {
  const { t } = useTranslation("gateway");
  const { locale } = useLocalePreference();
  const cost = (value: number | null) => value === null ? t($ => $.operations.unavailable) : formatMicroUSD(value, locale) ?? t($ => $.review.states.unavailable);
  return (
    <li data-source={entry.source} data-status={entry.status}>
      <div className="application-operations-timeline-marker" aria-hidden="true" />
      <article>
        <div className="card-title-row">
          <div><p className="eyebrow">{entry.source === "gateway_request" ? t($ => $.operations.requests) : t($ => $.operations.runs)}</p><h4>{entry.operation || t($ => $.operations.unavailable)}</h4></div>
          <span className={`application-operations-status ${entry.status}`}>{gatewayReviewState(t, entry.status)}</span>
        </div>
        <p><code>{entry.recordId}</code> · <time dateTime={entry.startedAt} title={entry.startedAt}>{formatDisplayDate(entry.startedAt, locale) ?? t($ => $.operations.unavailable)}</time> · {entry.durationMs} ms</p>
        <dl>
          <div><dt>{t($ => $.operations.contract)}</dt><dd>{entry.contract || t($ => $.operations.unavailable)}</dd></div>
          <div><dt>{t($ => $.operations.route)}</dt><dd>{entry.provider || t($ => $.operations.unavailable)} / {entry.profile || t($ => $.operations.default)} / {entry.model || t($ => $.operations.unavailable)}{entry.providerAttempts ? t($ => $.operations.entryAttempts, { count: entry.providerAttempts, fallback: entry.fallbackUsed ? t($ => $.operations.fallbackUsed) : t($ => $.operations.fallbackNotUsed) }) : ""}</dd></div>
          <div><dt>{t($ => $.operations.failure)}</dt><dd>{entry.failureCode || t($ => $.operations.none)} · {entry.failureBoundary || t($ => $.operations.none)}</dd></div>
          <div><dt>{t($ => $.operations.requestAudit)}</dt><dd>{entry.requestId} · {entry.auditRef}</dd></div>
          {entry.source === "gateway_request" ? (
            <>
              <div>
                <dt>{t($ => $.operations.usage)}</dt>
                <dd>
                  {entry.usageAvailability === "reported"
                    ? t($ => $.operations.entryTokens, { total: entry.totalTokens, input: entry.inputTokens, output: entry.outputTokens, source: entry.usageSource })
                    : gatewayReviewState(t, entry.usageAvailability ?? "unavailable")}
                </dd>
              </div>
              <div>
                <dt>{t($ => $.operations.cost)}</dt>
                <dd>{entry.attemptCostCoverage
                  ? t($ => $.operations.entryCost, { cost: cost(entry.estimatedCostMicros), coverage: gatewayReviewState(t, entry.attemptCostCoverage) })
                  : entry.costAvailability === "estimated"
                    ? t($ => $.operations.entryPolicy, { cost: cost(entry.estimatedCostMicros), version: entry.pricingPolicyVersion ?? t($ => $.operations.unavailable) })
                    : `${gatewayReviewState(t, entry.costAvailability ?? "unavailable")} · ${entry.costReason}`}</dd>
              </div>
            </>
          ) : (
            <div><dt>{t($ => $.operations.calls)}</dt><dd>{t($ => $.operations.entryCalls, { provider: entry.providerCalls, retrieval: entry.retrievalCalls, tool: entry.toolCalls })}</dd></div>
          )}
        </dl>
        {entry.source === "gateway_request" && onOpenGatewayRequest ? (
          <button type="button" className="secondary-action" onClick={() => onOpenGatewayRequest(entry.recordId, gatewayConfig.consumerRef)}>
            {t($ => $.operations.openRequest)}
          </button>
        ) : entry.source === "workflow_run" && onOpenWorkflowRun ? (
          <button type="button" className="secondary-action" onClick={() => onOpenWorkflowRun(entry.recordId)}>
            {t($ => $.operations.openRun)}
          </button>
        ) : null}
      </article>
    </li>
  );
}

function StatusBadge({ status }: { status: ApplicationOperationsState["status"] }) {
  const { t } = useTranslation("gateway");
  const tone = status === "ready" || status === "empty" ? "good" :
    status === "failed" || status === "application_unavailable" ? "bad" : "neutral";
  return <span className={`status-badge ${tone}`}>{gatewayReviewState(t, status)}</span>;
}

function CostCoverage({
  label,
  value,
  tone,
}: {
  label: string;
  value: number;
  tone: "ready" | "neutral" | "attention" | "blocked";
}) {
  return <span data-tone={tone}><strong>{value}</strong>{label}</span>;
}

function emptyMessage(t: TFunction<"gateway">, state: ApplicationOperationsState): string {
  if (state.status === "offline") return t($ => $.operations.emptyOffline);
  if (state.status === "application_unavailable") return t($ => $.operations.emptyApplication);
  if (state.status === "loading") return t($ => $.operations.emptyLoading);
  if (state.status === "failed") return t($ => $.operations.emptyFailed);
  return t($ => $.operations.empty);
}
