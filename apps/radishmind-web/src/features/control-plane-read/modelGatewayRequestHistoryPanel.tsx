import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import {
  EMPTY_GATEWAY_REQUEST_HISTORY_FILTER,
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
        failureCode: "gateway_request_store_unavailable",
        failureSummary: error instanceof Error ? error.message : "Gateway request history is unavailable.",
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
          failureCode: "gateway_request_store_unavailable",
          failureSummary: error instanceof Error ? error.message : "Gateway request history is unavailable.",
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
        setDetailFailure(error instanceof Error ? error.message : "Gateway request history handoff failed.");
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
      setDetailFailure(error instanceof Error ? error.message : "Gateway request detail is unavailable.");
    }
  }

  const failedCount = history.requests.filter((request) => request.status === "failed").length;
  const canceledCount = history.requests.filter((request) => request.status === "canceled").length;
  const usageReportedCount = history.requests.filter((request) => request.usageAvailability === "reported").length;
  const estimatedCostCount = history.requests.filter((request) => request.costEstimate.availability === "estimated").length;
  const loadedCostMicros = history.requests.reduce(
    (total, request) => total + (request.costEstimate.estimatedCostMicros ?? 0),
    0,
  );
  const staleCount = history.requests.filter((request) => request.staleStarted).length;

  return (
    <div className="gateway-request-history" id="model-gateway-request-history">
      <div className="model-gateway-overview-subheading gateway-request-history-heading">
        <div>
          <p className="eyebrow">Real Request History</p>
          <h4>Usage, cost snapshot, and stable failure review</h4>
          <p>{config.applicationId || "application unavailable"} · {config.workspaceId} · {config.consumerRef}</p>
        </div>
        <span className={`status-badge ${config.mode === "dev_gateway_request_history_http" && workspaceScopeMatches ? "good" : "neutral"}`}>
          {config.mode === "dev_gateway_request_history_http" && workspaceScopeMatches ? (history.requests[0]?.storeMode ?? "dev/test") : "read only"}
        </span>
      </div>

      {config.mode === "dev_gateway_request_history_http" && !workspaceScopeMatches ? (
        <article className="model-gateway-overview-trace gateway-request-history-blocked" role="alert">
          <p className="eyebrow">Workspace boundary</p>
          <h5>Request history source scope does not match</h5>
          <p>Configured source <code>{baseConfig.workspaceId}</code> cannot be read in Application Workspace <code>{workspaceId || "unavailable"}</code>. No history request is sent.</p>
        </article>
      ) : config.mode !== "dev_gateway_request_history_http" ? (
        <article className="model-gateway-overview-trace">
          <p className="eyebrow">Offline evidence</p>
          <h5>No live Gateway history request</h5>
          <p>Enable the explicit Gateway request-history source to read sanitized records. Existing quota, cost, workflow, and audit evidence is not used as a fallback.</p>
        </article>
      ) : (
        <>
          <div className="gateway-request-history-summary">
            <article className="model-gateway-overview-trace">
              <p className="eyebrow">Scoped dev/test API</p>
              <h5>/v1/model-gateway/requests</h5>
              <p>{config.workspaceId} · {config.applicationId || "unbound"} · {config.consumerRef} · {history.status}</p>
              <dl className="model-gateway-overview-meta">
                <div><dt>Records</dt><dd>{history.requests.length}</dd></div>
                <div><dt>Failed / canceled</dt><dd>{failedCount} / {canceledCount}</dd></div>
                <div><dt>Usage reported</dt><dd>{usageReportedCount}</dd></div>
                <div><dt>Cost estimated</dt><dd>{estimatedCostCount} · {formatCostMicros(loadedCostMicros)}</dd></div>
                <div><dt>Window</dt><dd>{history.hasMore ? "partial · has_more" : "loaded complete"}</dd></div>
                <div><dt>Stale started</dt><dd>{staleCount}</dd></div>
              </dl>
            </article>
            <GatewayRequestHistoryFilters filter={filter} onChange={setFilter} onApply={() => void load()} loading={history.status === "loading"} />
          </div>

          {history.failureCode ? <p className="failure-summary">{history.failureCode}: {history.failureSummary}</p> : null}
          {history.status === "empty" ? <p className="boundary-note">No request records match the current caller scope and filters.</p> : null}

          <div className="gateway-request-history-list" aria-label="Real Gateway request records">
            {history.requests.map((request) => (
              <button
                type="button"
                className={`gateway-request-history-row ${selectedRequestId === request.requestId ? "is-selected" : ""}`}
                key={request.requestId}
                onClick={() => void selectRequest(request)}
                aria-pressed={selectedRequestId === request.requestId}
                data-status={request.status}
              >
                <span><strong>{request.route}</strong><small>{request.protocol} · {request.stream ? "stream" : "unary"}</small></span>
                <span><small>Provider / model</small><strong>{request.selectedProvider || "unavailable"}</strong><small>{request.selectedProfile || "no profile"} · {request.selectedModel || "unavailable"}{request.providerRouteGeneration ? ` · generation ${request.providerRouteGeneration}` : ""}</small></span>
                <span><small>Status / failure</small><strong className={`gateway-request-history-status ${request.status}`}>{request.status}{request.staleStarted ? " · stale" : ""}</strong><small>{request.failureBoundary || "no failure"}</small></span>
                <span>
                  <small>Usage / duration</small>
                  <strong>{request.usageAvailability === "reported" ? `${request.totalTokens} tokens` : request.usageAvailability}</strong>
                  <small>{request.usageAvailability === "reported" ? `${request.inputTokens} in · ${request.outputTokens} out · ${request.usageSource}` : "Provider usage unavailable"} · {request.durationMs} ms · provider {request.providerDurationAvailable ? `${request.providerDurationMs} ms` : "unavailable"}</small>
                </span>
                <span>
                  <small>Cost snapshot</small>
                  <strong>{costEstimateLabel(request.costEstimate)}</strong>
                  <small>{request.costEstimate.availability === "estimated" ? `policy v${request.costEstimate.pricingPolicyVersion} · immutable` : request.costEstimate.reason}</small>
                </span>
                <span><small>Started</small><strong>{formatTimestamp(request.startedAt)}</strong></span>
              </button>
            ))}
          </div>

          {history.hasMore ? (
            <button type="button" onClick={() => void load(history.nextCursor, true)} disabled={history.status === "loading"}>Load earlier requests</button>
          ) : null}
          {detailFailure ? <p className="failure-summary">{detailFailure}</p> : null}
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
  return (
    <details className="gateway-request-history-filter-disclosure">
      <summary>Exact filters <span>route · protocol · status · time</span></summary>
      <div className="gateway-request-history-filters" aria-label="Gateway request history filters">
        <label>Route<input value={filter.route} onChange={(event) => onChange({ ...filter, route: event.target.value })} placeholder="exact route" /></label>
        <label>Protocol<select value={filter.protocol} onChange={(event) => onChange({ ...filter, protocol: event.target.value as GatewayRequestHistoryFilter["protocol"] })}><option value="">All</option><option value="openai-chat-completions">Chat Completions</option><option value="openai-responses">Responses</option><option value="anthropic-messages">Messages</option></select></label>
        <label>Provider<input value={filter.provider} onChange={(event) => onChange({ ...filter, provider: event.target.value })} placeholder="exact provider" /></label>
        <label>Profile<input value={filter.profile} onChange={(event) => onChange({ ...filter, profile: event.target.value })} placeholder="exact profile" /></label>
        <label>Model<input value={filter.model} onChange={(event) => onChange({ ...filter, model: event.target.value })} placeholder="exact model" /></label>
        <label>Status<select value={filter.status} onChange={(event) => onChange({ ...filter, status: event.target.value as GatewayRequestHistoryFilter["status"] })}><option value="">All</option><option value="started">Started</option><option value="succeeded">Succeeded</option><option value="failed">Failed</option><option value="canceled">Canceled</option></select></label>
        <label>Failure boundary<input value={filter.failureBoundary} onChange={(event) => onChange({ ...filter, failureBoundary: event.target.value })} placeholder="exact boundary" /></label>
        <label>Usage<select value={filter.usageAvailability} onChange={(event) => onChange({ ...filter, usageAvailability: event.target.value as GatewayRequestHistoryFilter["usageAvailability"] })}><option value="">All</option><option value="reported">Reported</option><option value="not_reported">Not reported</option><option value="not_applicable">Not applicable</option></select></label>
        <label>Started from<input type="datetime-local" value={filter.startedFrom} onChange={(event) => onChange({ ...filter, startedFrom: event.target.value })} /></label>
        <label>Started to<input type="datetime-local" value={filter.startedTo} onChange={(event) => onChange({ ...filter, startedTo: event.target.value })} /></label>
        <button type="button" onClick={onApply} disabled={loading}>Apply filters</button>
      </div>
    </details>
  );
}

function GatewayRequestDetail({ detail }: { detail: GatewayRequestHistoryDetail }) {
  return (
    <article className="gateway-request-history-detail">
      <div className="model-gateway-overview-row-main">
        <div><p className="eyebrow">Sanitized request detail</p><h5>{detail.requestId}</h5></div>
        <span className={`status-badge ${detail.status === "succeeded" ? "good" : detail.status === "started" ? "neutral" : "bad"}`}>{detail.status}</span>
      </div>
      <dl className="gateway-request-history-detail-grid">
        <div><dt>Caller scope</dt><dd>{detail.tenantRef} / {detail.workspaceId} / {detail.consumerRef}</dd></div>
        <div><dt>Application / subject</dt><dd>{detail.applicationId || "unbound"} / {detail.subjectRef}</dd></div>
        <div><dt>Selection</dt><dd>{detail.selectionSource || "unavailable"} · {detail.selectedProvider || "unavailable"} / {detail.selectedProfile || "no profile"} / {detail.selectedModel || "unavailable"}</dd></div>
        <div><dt>Provider route snapshot</dt><dd>{detail.providerRouteGeneration ? `${detail.providerRouteConfigurationId} · generation ${detail.providerRouteGeneration} · ${shortDigest(detail.providerRouteSnapshotDigest)}` : "static configuration / unavailable"}</dd></div>
        <div><dt>Timing</dt><dd>total {detail.durationMs} ms · gateway {detail.gatewayDurationAvailable ? `${detail.gatewayDurationMs} ms` : "unavailable"} · provider {detail.providerDurationAvailable ? `${detail.providerDurationMs} ms` : "unavailable"}</dd></div>
        <div><dt>Usage</dt><dd>{detail.usageAvailability}{detail.usageAvailability === "reported" ? ` · ${detail.inputTokens} in / ${detail.outputTokens} out / ${detail.totalTokens} total` : ""}{detail.usageSource ? ` · ${detail.usageSource}` : ""}</dd></div>
        <div><dt>Cost availability</dt><dd>{detail.costEstimate.availability}{detail.costEstimate.availability === "estimated" ? ` · ${formatCostMicros(detail.costEstimate.estimatedCostMicros ?? 0)}` : ` · ${detail.costEstimate.reason}`}</dd></div>
        <div><dt>Pricing snapshot</dt><dd>{detail.costEstimate.availability === "estimated" ? `${detail.costEstimate.pricingPolicyId} · v${detail.costEstimate.pricingPolicyVersion} · ${shortDigest(detail.costEstimate.pricingPolicyDigest)}` : "not captured for estimation"}</dd></div>
        <div><dt>Snapshot rates</dt><dd>{detail.costEstimate.availability === "estimated" ? `${formatCostMicros(detail.costEstimate.inputPriceMicrosPerTokenUnit ?? 0)} input / ${formatCostMicros(detail.costEstimate.outputPriceMicrosPerTokenUnit ?? 0)} output · USD / 1M · ${detail.costEstimate.roundingMode}` : "not applicable"}</dd></div>
        <div><dt>HTTP / failure</dt><dd>{detail.httpStatusCode || "unavailable"} · {detail.failureBoundary || "no failure"} · {detail.failureCode || "none"}</dd></div>
        <div><dt>Started / completed</dt><dd>{formatTimestamp(detail.startedAt)} / {detail.completedAt ? formatTimestamp(detail.completedAt) : "not completed"}</dd></div>
        <div><dt>Request / audit</dt><dd>{detail.requestId} / {detail.auditRef}</dd></div>
        <div><dt>Record</dt><dd>{detail.schemaVersion} · version {detail.recordVersion} · {detail.storeMode}{detail.staleStarted ? " · stale started" : ""}</dd></div>
      </dl>
      <p className="boundary-note">The amount is a request-local development/test estimate, not a Provider invoice or billing write. Historical records are never recalculated. Raw input, output, credentials, endpoints, and provider envelopes are not retained or exposed.</p>
    </article>
  );
}

function formatTimestamp(value: string): string {
  if (!value) return "unavailable";
  const timestamp = new Date(value);
  return Number.isNaN(timestamp.valueOf()) ? value : timestamp.toLocaleString();
}

function shortDigest(value: string): string {
  return value.length > 24 ? `${value.slice(0, 16)}…${value.slice(-8)}` : value;
}

function costEstimateLabel(estimate: GatewayRequestHistorySummary["costEstimate"]): string {
  return estimate.availability === "estimated"
    ? formatCostMicros(estimate.estimatedCostMicros ?? 0)
    : estimate.availability.replaceAll("_", " ");
}

function formatCostMicros(value: number): string {
  return `$${(value / 1_000_000).toFixed(6)}`;
}
