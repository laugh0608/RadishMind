import "../../i18n/requestHistoryResources.ts";
import "../../i18n/gatewayReviewResources.ts";
import { useTranslation } from "react-i18next";
import { useLocalePreference } from "../../i18n/LocaleProvider.tsx";
import { formatDisplayDate, formatDisplayNumber, formatMicroUSD } from "../../i18n/formatters.ts";
import { gatewayReviewState, gatewayHistoryFailure } from "./gatewayReviewMessages.ts";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import {
  EMPTY_GATEWAY_REQUEST_HISTORY_FILTER,
  GatewayRequestHistoryError,
  initialGatewayRequestHistoryState,
  listGatewayRequestHistory,
  readGatewayRequestHistoryDetail,
  readModelGatewayRequestHistoryConfig,
  type GatewayRequestHistoryDetail,
  type GatewayRequestHistoryFilter,
  type GatewayRequestHistorySummary,
} from "./modelGatewayRequestHistoryConsumer.ts";
import { MODEL_GATEWAY_REQUEST_REVIEW_EVENT, type ModelGatewayRequestReviewEventDetail } from "./modelGatewayPlaygroundEvents.ts";

const baseConfig = readModelGatewayRequestHistoryConfig();

export default function ModelGatewayRequestHistoryPanel({
  selectedApplicationId,
  workspaceId,
  active,
}: {
  selectedApplicationId: string;
  workspaceId: string;
  active: boolean;
}) {
  const { t } = useTranslation("gateway");
  const { locale } = useLocalePreference();
  const number = (value: number) => formatDisplayNumber(value, locale) ?? t($ => $.review.states.unavailable);
  const cost = (value: number | null) => value === null ? t($ => $.requestHistory.unavailable) : formatMicroUSD(value, locale) ?? t($ => $.review.states.unavailable);
  const date = (value: string) => formatDisplayDate(value, locale) ?? t($ => $.requestHistory.unavailable);
  const [reviewScope, setReviewScope] = useState({
    applicationId: selectedApplicationId.trim(),
    consumerRef: baseConfig.consumerRef,
  });
  const [filter, setFilter] = useState<GatewayRequestHistoryFilter>(EMPTY_GATEWAY_REQUEST_HISTORY_FILTER);
  const [history, setHistory] = useState(() => initialGatewayRequestHistoryState(baseConfig));
  const [selectedRequestId, setSelectedRequestId] = useState("");
  const [detail, setDetail] = useState<GatewayRequestHistoryDetail | null>(null);
  const [detailFailure, setDetailFailure] = useState("");
  const config = useMemo(() => ({ ...baseConfig, ...reviewScope }), [reviewScope]);
  const scopeGeneration = useRef(0);
  const handoffInFlight = useRef(false);
  const previousActive = useRef(active);
  const historyRequests = useRef(history.requests);
  historyRequests.current = history.requests;
  const workspaceScopeMatches = Boolean(workspaceId.trim()) && baseConfig.workspaceId === workspaceId.trim();

  const load = useCallback(async (cursor = "", append = false) => {
    if (!active || !workspaceScopeMatches || config.mode !== "dev_gateway_request_history_http") return;
    const generation = scopeGeneration.current;
    setHistory((current) => ({ ...current, status: "loading", failureCode: "", failureSummary: "" }));
    try {
      const next = await listGatewayRequestHistory(config, filter, cursor, append ? historyRequests.current : []);
      if (scopeGeneration.current !== generation) return;
      setHistory(next);
    } catch (error) {
      if (scopeGeneration.current !== generation) return;
      setHistory((current) => ({
        ...current,
        status: "failed",
        requests: append ? current.requests : [],
        failureCode: error instanceof GatewayRequestHistoryError ? error.code : "gateway_request_store_unavailable",
        failureSummary: "",
      }));
    }
  }, [active, config, filter, workspaceScopeMatches]);

  useEffect(() => {
    scopeGeneration.current += 1;
    handoffInFlight.current = false;
    const applicationId = selectedApplicationId.trim();
    const nextConfig = { ...baseConfig, applicationId, consumerRef: baseConfig.consumerRef };
    setReviewScope({ applicationId, consumerRef: baseConfig.consumerRef });
    setFilter(EMPTY_GATEWAY_REQUEST_HISTORY_FILTER);
    setSelectedRequestId("");
    setDetail(null);
    setDetailFailure("");
    setHistory(initialGatewayRequestHistoryState(nextConfig));
    if (active && workspaceScopeMatches && nextConfig.mode === "dev_gateway_request_history_http") {
      const generation = scopeGeneration.current;
      void listGatewayRequestHistory(nextConfig, EMPTY_GATEWAY_REQUEST_HISTORY_FILTER).then((next) => {
        if (scopeGeneration.current === generation) setHistory(next);
      }).catch((error: unknown) => {
        if (scopeGeneration.current !== generation) return;
        setHistory((current) => ({
          ...current,
          status: "failed",
          requests: [],
          failureCode: error instanceof GatewayRequestHistoryError ? error.code : "gateway_request_store_unavailable",
          failureSummary: "",
        }));
      });
    }
  }, [selectedApplicationId, workspaceId, workspaceScopeMatches]);

  useEffect(() => {
    const becameActive = active && !previousActive.current;
    previousActive.current = active;
    if (becameActive && !handoffInFlight.current) void load();
  }, [active, load]);

  useEffect(() => {
    function reviewPlaygroundRequest(event: Event) {
      const requestId = (event as CustomEvent<ModelGatewayRequestReviewEventDetail>).detail?.requestId?.trim();
      const nextApplicationId = (event as CustomEvent<ModelGatewayRequestReviewEventDetail>).detail?.applicationId?.trim();
      const nextConsumerRef = (event as CustomEvent<ModelGatewayRequestReviewEventDetail>).detail?.consumerRef?.trim() || baseConfig.consumerRef;
      if (!requestId || !nextApplicationId || nextApplicationId !== selectedApplicationId.trim() ||
        !workspaceScopeMatches || baseConfig.mode !== "dev_gateway_request_history_http") return;
      const reviewConfig = { ...baseConfig, applicationId: nextApplicationId, consumerRef: nextConsumerRef };
      const generation = scopeGeneration.current + 1;
      scopeGeneration.current = generation;
      handoffInFlight.current = true;
      setReviewScope({ applicationId: nextApplicationId, consumerRef: nextConsumerRef });
      setFilter(EMPTY_GATEWAY_REQUEST_HISTORY_FILTER);
      setSelectedRequestId(requestId);
      setDetail(null);
      setDetailFailure("");
      setHistory((current) => ({ ...current, status: "loading", failureCode: "", failureSummary: "" }));
      void Promise.all([
        listGatewayRequestHistory(reviewConfig, EMPTY_GATEWAY_REQUEST_HISTORY_FILTER),
        readGatewayRequestHistoryDetail(reviewConfig, requestId),
      ]).then(([nextHistory, nextDetail]) => {
        if (scopeGeneration.current !== generation) return;
        handoffInFlight.current = false;
        setHistory(nextHistory);
        setDetail(nextDetail);
      }).catch((error: unknown) => {
        if (scopeGeneration.current !== generation) return;
        handoffInFlight.current = false;
        setDetailFailure(error instanceof GatewayRequestHistoryError ? error.code : "gateway_request_store_unavailable");
        setHistory((current) => ({ ...current, status: "failed", failureCode: "gateway_request_store_unavailable" }));
      });
    }
    window.addEventListener(MODEL_GATEWAY_REQUEST_REVIEW_EVENT, reviewPlaygroundRequest);
    return () => window.removeEventListener(MODEL_GATEWAY_REQUEST_REVIEW_EVENT, reviewPlaygroundRequest);
  }, [selectedApplicationId, workspaceScopeMatches]);

  async function selectRequest(request: GatewayRequestHistorySummary) {
    if (!active || !workspaceScopeMatches) return;
    const generation = scopeGeneration.current;
    setSelectedRequestId(request.requestId);
    setDetail(null);
    setDetailFailure("");
    try {
      const nextDetail = await readGatewayRequestHistoryDetail(config, request.requestId);
      if (scopeGeneration.current === generation) setDetail(nextDetail);
    } catch (error) {
      if (scopeGeneration.current !== generation) return;
      setDetailFailure(error instanceof GatewayRequestHistoryError ? error.code : "gateway_request_store_unavailable");
    }
  }

  const failedCount = history.requests.filter((request) => request.status === "failed").length;
  const canceledCount = history.requests.filter((request) => request.status === "canceled").length;
  const usageReportedCount = history.requests.filter((request) => request.usageAvailability === "reported").length;
  const estimatedCostCount = history.requests.filter((request) => request.costEstimate.availability === "estimated").length;
  const fallbackUsedCount = history.requests.filter((request) => request.fallbackUsed).length;
  const partialCostCount = history.requests.filter((request) => request.attemptCostSummary?.coverage === "partial").length;
  const loadedCostMicros = history.requests.reduce(
    (total, request) => total + (request.attemptCostSummary?.knownCostMicros ?? request.costEstimate.estimatedCostMicros ?? 0),
    0,
  );
  const staleCount = history.requests.filter((request) => request.staleStarted).length;

  return (
    <div className="gateway-request-history" id="model-gateway-request-history">
      <div className="model-gateway-overview-subheading gateway-request-history-heading">
        <div>
          <p className="eyebrow">{t($ => $.requestHistory.title)}</p>
          <h4>{t($ => $.requestHistory.subtitle)}</h4>
          <p>{config.applicationId || t($ => $.requestHistory.applicationUnavailable)} · {config.workspaceId} · {config.consumerRef}</p>
        </div>
        <span className={`status-badge ${config.mode === "dev_gateway_request_history_http" && workspaceScopeMatches ? "good" : "neutral"}`}>
          {config.mode === "dev_gateway_request_history_http" && workspaceScopeMatches ? (history.requests[0]?.storeMode ?? t($ => $.requestHistory.devTest)) : t($ => $.requestHistory.readOnly)}
        </span>
      </div>

      {config.mode === "dev_gateway_request_history_http" && !workspaceScopeMatches ? (
        <article className="model-gateway-overview-trace gateway-request-history-blocked" role="alert">
          <p className="eyebrow">{t($ => $.requestHistory.workspaceBoundary)}</p>
          <h5>{t($ => $.requestHistory.scopeMismatch)}</h5>
          <p>{t($ => $.requestHistory.scopeExplanation, { source: baseConfig.workspaceId, workspace: workspaceId || t($ => $.requestHistory.unavailable) })}</p>
        </article>
      ) : config.mode !== "dev_gateway_request_history_http" ? (
        <article className="model-gateway-overview-trace">
          <p className="eyebrow">{t($ => $.requestHistory.offlineEvidence)}</p>
          <h5>{t($ => $.requestHistory.noLive)}</h5>
          <p>{t($ => $.requestHistory.enableSource)}</p>
        </article>
      ) : (
        <>
          <div className="gateway-request-history-summary">
            <article className="model-gateway-overview-trace">
              <p className="eyebrow">{t($ => $.requestHistory.scopedAPI)}</p>
              <h5>/v1/model-gateway/requests</h5>
              <p>{config.workspaceId} · {config.applicationId || t($ => $.requestHistory.unbound)} · {config.consumerRef} · {gatewayReviewState(t, history.status)}</p>
              <dl className="model-gateway-overview-meta">
                <div><dt>{t($ => $.requestHistory.records)}</dt><dd>{history.requests.length}</dd></div>
                <div><dt>{t($ => $.requestHistory.failedCanceled)}</dt><dd>{failedCount} / {canceledCount}</dd></div>
                <div><dt>{t($ => $.requestHistory.usageReported)}</dt><dd>{usageReportedCount}</dd></div>
                <div><dt>{t($ => $.requestHistory.costEstimated)}</dt><dd>{estimatedCostCount} · {cost(loadedCostMicros)}</dd></div>
                <div><dt>{t($ => $.requestHistory.fallbackUsed)}</dt><dd>{fallbackUsedCount}</dd></div>
                <div><dt>{t($ => $.requestHistory.partialCost)}</dt><dd>{partialCostCount}</dd></div>
                <div><dt>{t($ => $.requestHistory.window)}</dt><dd>{history.hasMore ? t($ => $.requestHistory.partialWindow) : t($ => $.requestHistory.completeWindow)}</dd></div>
                <div><dt>{t($ => $.requestHistory.staleStarted)}</dt><dd>{staleCount}</dd></div>
              </dl>
            </article>
            <GatewayRequestHistoryFilters filter={filter} onChange={setFilter} onApply={() => void load()} loading={history.status === "loading"} />
          </div>

          {history.failureCode ? <p className="failure-summary">{history.failureCode}: {gatewayHistoryFailure(t, history.failureCode)}</p> : null}
          {history.status === "empty" ? <p className="boundary-note">{t($ => $.requestHistory.empty)}</p> : null}

          <div className="gateway-request-history-list" aria-label={t($ => $.requestHistory.requestRecords)}>
            {history.requests.map((request) => (
              <button
                type="button"
                className={`gateway-request-history-row ${selectedRequestId === request.requestId ? "is-selected" : ""}`}
                key={request.requestId}
                onClick={() => void selectRequest(request)}
                aria-pressed={selectedRequestId === request.requestId}
                data-status={request.status}
              >
                <span><strong>{request.route}</strong><small>{request.protocol} · {request.stream ? t($ => $.requestHistory.stream) : t($ => $.requestHistory.unary)}</small></span>
                <span><small>{t($ => $.requestHistory.providerModel)}</small><strong>{request.selectedProvider || t($ => $.requestHistory.unavailable)}</strong><small>{request.selectedProfile || t($ => $.requestHistory.noProfile)} · {request.selectedModel || t($ => $.requestHistory.unavailable)}{request.providerRouteGeneration ? t($ => $.requestHistory.generation, { version: request.providerRouteGeneration }) : ""}</small></span>
                <span>
                  <small>{t($ => $.requestHistory.attemptLineage)}</small>
                  <strong>{request.schemaVersion === "gateway_request_record.v3" ? t($ => $.requestHistory.attemptSummary, { count: request.attemptCount, usage: request.fallbackUsed ? t($ => $.requestHistory.used) : t($ => $.requestHistory.notUsed) }) : t($ => $.requestHistory.legacyAttempt)}</strong>
                  <small>{request.terminalProvider ? t($ => $.requestHistory.terminal, { provider: request.terminalProvider, profile: request.terminalProfile }) : t($ => $.requestHistory.terminalUnavailable)}</small>
                </span>
                <span><small>{t($ => $.requestHistory.statusFailure)}</small><strong className={`gateway-request-history-status ${request.status}`}>{gatewayReviewState(t, request.status)}{request.staleStarted ? t($ => $.requestHistory.stale) : ""}</strong><small>{request.failureBoundary || t($ => $.requestHistory.noFailure)}</small></span>
                <span>
                  <small>{t($ => $.requestHistory.usageDuration)}</small>
                  <strong>{request.usageAvailability === "reported" ? t($ => $.requestHistory.tokens, { displayCount: number(request.totalTokens) }) : gatewayReviewState(t, request.usageAvailability)}</strong>
                  <small>{request.usageAvailability === "reported" ? t($ => $.requestHistory.tokenDetail, { input: number(request.inputTokens), output: number(request.outputTokens), source: request.usageSource }) : t($ => $.requestHistory.usageUnavailable)} · {t($ => $.requestHistory.durationDetail, { total: number(request.durationMs), provider: request.providerDurationAvailable ? `${number(request.providerDurationMs)} ms` : t($ => $.requestHistory.unavailable) })}</small>
                </span>
                <span>
                  <small>{t($ => $.requestHistory.costSnapshot)}</small>
                  <strong>{request.attemptCostSummary ? `${cost(request.attemptCostSummary.knownCostMicros)} · ${gatewayReviewState(t, request.attemptCostSummary.coverage)}` : request.costEstimate.availability === "estimated" ? cost(request.costEstimate.estimatedCostMicros) : gatewayReviewState(t, request.costEstimate.availability)}</strong>
                  <small>{request.attemptCostSummary ? t($ => $.requestHistory.costCounts, { estimated: number(request.attemptCostSummary.estimatedAttemptCount), unknown: number(request.attemptCostSummary.unknownAttemptCount) }) : request.costEstimate.availability === "estimated" ? t($ => $.requestHistory.immutablePolicy, { version: request.costEstimate.pricingPolicyVersion ?? t($ => $.requestHistory.unavailable) }) : request.costEstimate.reason}</small>
                </span>
                <span><small>{t($ => $.requestHistory.started)}</small><strong>{<time dateTime={request.startedAt} title={request.startedAt}>{date(request.startedAt)}</time>}</strong></span>
              </button>
            ))}
          </div>

          {history.hasMore ? (
            <button type="button" onClick={() => void load(history.nextCursor, true)} disabled={history.status === "loading"}>{t($ => $.requestHistory.loadEarlier)}</button>
          ) : null}
          {detailFailure ? <p className="failure-summary">{detailFailure}: {gatewayHistoryFailure(t, detailFailure)}</p> : null}
          {detail ? <GatewayRequestDetail detail={detail} /> : null}
        </>
      )}
    </div>
  );
}

function GatewayRequestHistoryFilters({
  filter,
  onChange,
  onApply,
  loading,
}: {
  filter: GatewayRequestHistoryFilter;
  onChange: (filter: GatewayRequestHistoryFilter) => void;
  onApply: () => void;
  loading: boolean;
}) {
  const { t } = useTranslation("gateway");
  return (
    <details className="gateway-request-history-filter-disclosure">
      <summary>{t($ => $.requestHistory.exactFilters)} <span>{t($ => $.requestHistory.filterHint)}</span></summary>
      <div className="gateway-request-history-filters" aria-label={t($ => $.requestHistory.filters)}>
        <label>{t($ => $.requestHistory.route)}<input value={filter.route} onChange={(event) => onChange({ ...filter, route: event.target.value })} placeholder={t($ => $.requestHistory.exactRoute)} /></label>
        <label>{t($ => $.requestHistory.protocol)}<select value={filter.protocol} onChange={(event) => onChange({ ...filter, protocol: event.target.value as GatewayRequestHistoryFilter["protocol"] })}><option value="">{t($ => $.requestHistory.all)}</option><option value="openai-chat-completions">Chat Completions</option><option value="openai-responses">Responses</option><option value="anthropic-messages">Messages</option></select></label>
        <label>{t($ => $.requestHistory.provider)}<input value={filter.provider} onChange={(event) => onChange({ ...filter, provider: event.target.value })} placeholder={t($ => $.requestHistory.exactProvider)} /></label>
        <label>{t($ => $.requestHistory.profile)}<input value={filter.profile} onChange={(event) => onChange({ ...filter, profile: event.target.value })} placeholder={t($ => $.requestHistory.exactProfile)} /></label>
        <label>{t($ => $.requestHistory.model)}<input value={filter.model} onChange={(event) => onChange({ ...filter, model: event.target.value })} placeholder={t($ => $.requestHistory.exactModel)} /></label>
        <label>{t($ => $.requestHistory.status)}<select value={filter.status} onChange={(event) => onChange({ ...filter, status: event.target.value as GatewayRequestHistoryFilter["status"] })}><option value="">{t($ => $.requestHistory.all)}</option><option value="started">{t($ => $.requestHistory.started)}</option><option value="succeeded">{t($ => $.requestHistory.succeeded)}</option><option value="failed">{t($ => $.requestHistory.failed)}</option><option value="canceled">{t($ => $.requestHistory.canceled)}</option></select></label>
        <label>{t($ => $.requestHistory.failureBoundary)}<input value={filter.failureBoundary} onChange={(event) => onChange({ ...filter, failureBoundary: event.target.value })} placeholder={t($ => $.requestHistory.exactBoundary)} /></label>
        <label>{t($ => $.requestHistory.usage)}<select value={filter.usageAvailability} onChange={(event) => onChange({ ...filter, usageAvailability: event.target.value as GatewayRequestHistoryFilter["usageAvailability"] })}><option value="">{t($ => $.requestHistory.all)}</option><option value="reported">{t($ => $.requestHistory.reported)}</option><option value="not_reported">{t($ => $.requestHistory.notReported)}</option><option value="not_applicable">{t($ => $.requestHistory.notApplicable)}</option></select></label>
        <label>{t($ => $.requestHistory.fallbackUsed)}<select value={filter.fallbackUsed} onChange={(event) => onChange({ ...filter, fallbackUsed: event.target.value as GatewayRequestHistoryFilter["fallbackUsed"] })}><option value="">{t($ => $.requestHistory.all)}</option><option value="true">{t($ => $.requestHistory.used)}</option><option value="false">{t($ => $.requestHistory.notUsed)}</option></select></label>
        <label>{t($ => $.requestHistory.terminalProvider)}<input value={filter.terminalProvider} onChange={(event) => onChange({ ...filter, terminalProvider: event.target.value })} placeholder={t($ => $.requestHistory.exactTerminalProvider)} /></label>
        <label>{t($ => $.requestHistory.terminalProfile)}<input value={filter.terminalProfile} onChange={(event) => onChange({ ...filter, terminalProfile: event.target.value })} placeholder={t($ => $.requestHistory.exactTerminalProfile)} /></label>
        <label>{t($ => $.requestHistory.startedFrom)}<input type="datetime-local" value={filter.startedFrom} onChange={(event) => onChange({ ...filter, startedFrom: event.target.value })} /></label>
        <label>{t($ => $.requestHistory.startedTo)}<input type="datetime-local" value={filter.startedTo} onChange={(event) => onChange({ ...filter, startedTo: event.target.value })} /></label>
        <button type="button" onClick={onApply} disabled={loading}>{t($ => $.requestHistory.apply)}</button>
      </div>
    </details>
  );
}

function GatewayRequestDetail({ detail }: { detail: GatewayRequestHistoryDetail }) {
  const { t } = useTranslation("gateway");
  const { locale } = useLocalePreference();
  const number = (value: number) => formatDisplayNumber(value, locale) ?? t($ => $.review.states.unavailable);
  const cost = (value: number | null) => value === null ? t($ => $.requestHistory.unavailable) : formatMicroUSD(value, locale) ?? t($ => $.review.states.unavailable);
  const date = (value: string) => formatDisplayDate(value, locale) ?? t($ => $.requestHistory.unavailable);
  return (
    <article className="gateway-request-history-detail">
      <div className="model-gateway-overview-row-main">
        <div><p className="eyebrow">{t($ => $.requestHistory.detail)}</p><h5>{detail.requestId}</h5></div>
        <span className={`status-badge ${detail.status === "succeeded" ? "good" : detail.status === "started" ? "neutral" : "bad"}`}>{gatewayReviewState(t, detail.status)}</span>
      </div>
      <dl className="gateway-request-history-detail-grid">
        <div><dt>{t($ => $.requestHistory.callerScope)}</dt><dd>{detail.tenantRef} / {detail.workspaceId} / {detail.consumerRef}</dd></div>
        <div><dt>{t($ => $.requestHistory.applicationSubject)}</dt><dd>{detail.applicationId || t($ => $.requestHistory.unbound)} / {detail.subjectRef}</dd></div>
        <div><dt>{t($ => $.requestHistory.selection)}</dt><dd>{detail.selectionSource || t($ => $.requestHistory.unavailable)} · {detail.selectedProvider || t($ => $.requestHistory.unavailable)} / {detail.selectedProfile || t($ => $.requestHistory.noProfile)} / {detail.selectedModel || t($ => $.requestHistory.unavailable)}</dd></div>
        <div><dt>{t($ => $.requestHistory.routeSnapshot)}</dt><dd>{detail.providerRouteGeneration ? t($ => $.requestHistory.routeLineage, { id: detail.providerRouteConfigurationId, version: detail.providerRouteGeneration, digest: shortDigest(detail.providerRouteSnapshotDigest) }) : t($ => $.requestHistory.staticConfiguration)}</dd></div>
        <div><dt>{t($ => $.requestHistory.timing)}</dt><dd>{t($ => $.requestHistory.timingDetail, { total: number(detail.durationMs), gateway: detail.gatewayDurationAvailable ? `${number(detail.gatewayDurationMs)} ms` : t($ => $.requestHistory.unavailable), provider: detail.providerDurationAvailable ? `${number(detail.providerDurationMs)} ms` : t($ => $.requestHistory.unavailable) })}</dd></div>
        <div><dt>{t($ => $.requestHistory.usage)}</dt><dd>{gatewayReviewState(t, detail.usageAvailability)}{detail.usageAvailability === "reported" ? t($ => $.requestHistory.usageDetail, { input: number(detail.inputTokens), output: number(detail.outputTokens), total: number(detail.totalTokens) }) : ""}{detail.usageSource ? ` · ${detail.usageSource}` : ""}</dd></div>
        <div><dt>{t($ => $.requestHistory.costAvailability)}</dt><dd>{gatewayReviewState(t, detail.costEstimate.availability)}{detail.costEstimate.availability === "estimated" ? ` · ${cost(detail.costEstimate.estimatedCostMicros)}` : ` · ${detail.costEstimate.reason}`}</dd></div>
        <div><dt>{t($ => $.requestHistory.pricingSnapshot)}</dt><dd>{detail.costEstimate.availability === "estimated" ? `${detail.costEstimate.pricingPolicyId} · v${detail.costEstimate.pricingPolicyVersion} · ${shortDigest(detail.costEstimate.pricingPolicyDigest)}` : t($ => $.requestHistory.notCaptured)}</dd></div>
        <div><dt>{t($ => $.requestHistory.rates)}</dt><dd>{detail.costEstimate.availability === "estimated" ? t($ => $.requestHistory.rateDetail, { input: cost(detail.costEstimate.inputPriceMicrosPerTokenUnit), output: cost(detail.costEstimate.outputPriceMicrosPerTokenUnit), rounding: detail.costEstimate.roundingMode }) : t($ => $.requestHistory.notApplicableValue)}</dd></div>
        <div><dt>{t($ => $.requestHistory.httpFailure)}</dt><dd>{detail.httpStatusCode || t($ => $.requestHistory.unavailable)} · {detail.failureBoundary || t($ => $.requestHistory.noFailure)} · {detail.failureCode || t($ => $.requestHistory.none)}</dd></div>
        <div><dt>{t($ => $.requestHistory.startedCompleted)}</dt><dd>{<time dateTime={detail.startedAt} title={detail.startedAt}>{date(detail.startedAt)}</time>} / {detail.completedAt ? <time dateTime={detail.completedAt} title={detail.completedAt}>{date(detail.completedAt)}</time> : t($ => $.requestHistory.notCompleted)}</dd></div>
        <div><dt>{t($ => $.requestHistory.requestAudit)}</dt><dd>{detail.requestId} / {detail.auditRef}</dd></div>
        <div><dt>{t($ => $.requestHistory.record)}</dt><dd>{detail.schemaVersion} · {t($ => $.requestHistory.version, { version: detail.recordVersion })} · {detail.storeMode}{detail.staleStarted ? t($ => $.requestHistory.staleSuffix) : ""}</dd></div>
      </dl>
      {detail.schemaVersion === "gateway_request_record.v3" && detail.attemptPlan ? (
        <section className="gateway-request-attempt-lineage" aria-label={t($ => $.requestHistory.providerAttemptLineage)}>
          <div className="model-gateway-overview-row-main">
            <div><p className="eyebrow">{t($ => $.requestHistory.durableLineage)}</p><h6>{gatewayReviewState(t, detail.attemptPlan.executionMode)}</h6></div>
            <span className={`status-badge ${detail.fallbackUsed ? "good" : "neutral"}`}>
              {gatewayReviewState(t, detail.attemptPhase)} · {t($ => $.requestHistory.attempts, { displayCount: number(detail.attemptCount) })}
            </span>
          </div>
          <p className="boundary-note">
            {t($ => $.requestHistory.planDetail, { mode: gatewayReviewState(t, detail.attemptPlan.fallbackMode), allowed: detail.fallbackAllowed ? t($ => $.requestHistory.allowed) : t($ => $.requestHistory.disabled), provider: detail.terminalProvider || t($ => $.requestHistory.unavailable), profile: detail.terminalProfile || t($ => $.requestHistory.unavailable) })}
          </p>
          <div className="gateway-request-attempt-list">
            {detail.providerAttempts.map((attempt) => (
              <article key={attempt.attemptId} data-status={attempt.status}>
                <div className="model-gateway-overview-row-main">
                  <div><p className="eyebrow">{attempt.ordinal === 1 ? t($ => $.requestHistory.primary) : t($ => $.requestHistory.backup)} · {t($ => $.requestHistory.attemptOrdinal, { attemptNumber: attempt.ordinal })}</p><h6>{attempt.providerId} / {attempt.runtimeProfile}</h6></div>
                  <span className={`status-badge ${attempt.status === "succeeded" ? "good" : attempt.status === "running" ? "neutral" : "bad"}`}>{gatewayReviewState(t, attempt.status)}</span>
                </div>
                <dl className="model-gateway-overview-meta">
                  <div><dt>{t($ => $.requestHistory.attemptId)}</dt><dd>{attempt.attemptId}</dd></div>
                  <div><dt>{t($ => $.requestHistory.configuredProfile)}</dt><dd>{attempt.configuredProfileId}</dd></div>
                  <div><dt>{t($ => $.requestHistory.quota)}</dt><dd>{attempt.quotaRejectionCode || attempt.quotaAdmissionId || t($ => $.requestHistory.notAdmitted)}</dd></div>
                  <div><dt>{t($ => $.requestHistory.duration)}</dt><dd>{attempt.durationMs} ms</dd></div>
                  <div><dt>{t($ => $.requestHistory.usage)}</dt><dd>{gatewayReviewState(t, attempt.usageAvailability)}{attempt.usageAvailability === "reported" ? ` · ${number(attempt.totalTokens)} token` : ""}</dd></div>
                  <div><dt>{t($ => $.requestHistory.cost)}</dt><dd>{attempt.costEstimate.availability === "estimated" ? cost(attempt.costEstimate.estimatedCostMicros) : gatewayReviewState(t, attempt.costEstimate.availability)}</dd></div>
                  <div><dt>{t($ => $.requestHistory.failure)}</dt><dd>{attempt.failure ? `${attempt.failure.failureClass} · ${attempt.failure.fallbackDisposition} · ${attempt.failure.outcome}` : attempt.failureBoundary || t($ => $.requestHistory.none)}</dd></div>
                  <div><dt>{t($ => $.requestHistory.completed)}</dt><dd>{attempt.completedAt ? <time dateTime={attempt.completedAt} title={attempt.completedAt}>{date(attempt.completedAt)}</time> : t($ => $.requestHistory.running)}</dd></div>
                </dl>
              </article>
            ))}
          </div>
          <p className="boundary-note">
            {t($ => $.requestHistory.knownCost, { cost: cost(detail.attemptCostSummary?.knownCostMicros ?? null), coverage: gatewayReviewState(t, detail.attemptCostSummary?.coverage ?? "none"), estimated: detail.attemptCostSummary?.estimatedAttemptCount ?? t($ => $.requestHistory.unavailable), unknown: detail.attemptCostSummary?.unknownAttemptCount ?? t($ => $.requestHistory.unavailable) })}
          </p>
        </section>
      ) : (
        <p className="boundary-note">{t($ => $.requestHistory.legacyBoundary)}</p>
      )}
      <p className="boundary-note">{t($ => $.requestHistory.privacyBoundary)}</p>
    </article>
  );
}

function shortDigest(value: string): string {
  return value.length > 24 ? `${value.slice(0, 16)}…${value.slice(-8)}` : value;
}
