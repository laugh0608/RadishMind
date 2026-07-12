import { useCallback, useEffect, useMemo, useState } from "react";

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

export default function ModelGatewayRequestHistoryPanel() {
  const [applicationId, setApplicationId] = useState(baseConfig.applicationId);
  const [filter, setFilter] = useState<GatewayRequestHistoryFilter>(EMPTY_GATEWAY_REQUEST_HISTORY_FILTER);
  const [history, setHistory] = useState(() => initialGatewayRequestHistoryState(baseConfig));
  const [selectedRequestId, setSelectedRequestId] = useState("");
  const [detail, setDetail] = useState<GatewayRequestHistoryDetail | null>(null);
  const [detailFailure, setDetailFailure] = useState("");
  const config = useMemo(() => ({ ...baseConfig, applicationId }), [applicationId]);

  const load = useCallback(async (cursor = "", append = false) => {
    if (config.mode !== "dev_gateway_request_history_http") return;
    setHistory((current) => ({ ...current, status: "loading", failureCode: "", failureSummary: "" }));
    try {
      const next = await listGatewayRequestHistory(config, filter, cursor, append ? history.requests : []);
      setHistory(next);
    } catch (error) {
      setHistory((current) => ({
        ...current,
        status: "failed",
        requests: append ? current.requests : [],
        failureCode: "gateway_request_store_unavailable",
        failureSummary: error instanceof Error ? error.message : "Gateway request history is unavailable.",
      }));
    }
  }, [config, filter, history.requests]);

  useEffect(() => { void load(); }, []);

  useEffect(() => {
    function reviewPlaygroundRequest(event: Event) {
      const requestId = (event as CustomEvent<ModelGatewayRequestReviewEventDetail>).detail?.requestId?.trim();
      const nextApplicationId = (event as CustomEvent<ModelGatewayRequestReviewEventDetail>).detail?.applicationId?.trim();
      if (!requestId || !nextApplicationId || baseConfig.mode !== "dev_gateway_request_history_http") return;
      const reviewConfig = { ...baseConfig, applicationId: nextApplicationId };
      setApplicationId(nextApplicationId);
      setFilter(EMPTY_GATEWAY_REQUEST_HISTORY_FILTER);
      setSelectedRequestId(requestId);
      setDetail(null);
      setDetailFailure("");
      setHistory((current) => ({ ...current, status: "loading", failureCode: "", failureSummary: "" }));
      void Promise.all([
        listGatewayRequestHistory(reviewConfig, EMPTY_GATEWAY_REQUEST_HISTORY_FILTER),
        readGatewayRequestHistoryDetail(reviewConfig, requestId),
      ]).then(([nextHistory, nextDetail]) => {
        setHistory(nextHistory);
        setDetail(nextDetail);
      }).catch((error: unknown) => {
        setDetailFailure(error instanceof Error ? error.message : "Gateway request history handoff failed.");
        setHistory((current) => ({ ...current, status: "failed", failureCode: "gateway_request_store_unavailable" }));
      });
    }
    window.addEventListener(MODEL_GATEWAY_REQUEST_REVIEW_EVENT, reviewPlaygroundRequest);
    return () => window.removeEventListener(MODEL_GATEWAY_REQUEST_REVIEW_EVENT, reviewPlaygroundRequest);
  }, []);

  async function selectRequest(request: GatewayRequestHistorySummary) {
    setSelectedRequestId(request.requestId);
    setDetail(null);
    setDetailFailure("");
    try {
      setDetail(await readGatewayRequestHistoryDetail(config, request.requestId));
    } catch (error) {
      setDetailFailure(error instanceof Error ? error.message : "Gateway request detail is unavailable.");
    }
  }

  const failedCount = history.requests.filter((request) => request.status === "failed").length;
  const canceledCount = history.requests.filter((request) => request.status === "canceled").length;
  const usageReportedCount = history.requests.filter((request) => request.usageAvailability === "reported").length;
  const staleCount = history.requests.filter((request) => request.staleStarted).length;

  return (
    <div className="gateway-request-history" id="model-gateway-request-history">
      <div className="model-gateway-overview-subheading gateway-request-history-heading">
        <div>
          <p className="eyebrow">Real Request History</p>
          <h4>Usage, timing, and stable failure review</h4>
        </div>
        <span className={`status-badge ${config.mode === "dev_gateway_request_history_http" ? "good" : "neutral"}`}>
          {config.mode === "dev_gateway_request_history_http" ? (history.requests[0]?.storeMode ?? "dev/test") : "offline evidence"}
        </span>
      </div>

      {config.mode !== "dev_gateway_request_history_http" ? (
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
              >
                <span><strong>{request.route}</strong><small>{request.protocol} · {request.stream ? "stream" : "unary"}</small></span>
                <span><small>Provider / model</small><strong>{request.selectedProvider || "unavailable"}</strong><small>{request.selectedProfile || "no profile"} · {request.selectedModel || "unavailable"}</small></span>
                <span><small>Status / failure</small><strong>{request.status}{request.staleStarted ? " · stale" : ""}</strong><small>{request.failureBoundary || "no failure"}</small></span>
                <span><small>Usage / duration</small><strong>{request.usageAvailability}</strong><small>{request.durationMs} ms · provider {request.providerDurationAvailable ? `${request.providerDurationMs} ms` : "unavailable"}</small></span>
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
        <div><dt>Timing</dt><dd>total {detail.durationMs} ms · gateway {detail.gatewayDurationAvailable ? `${detail.gatewayDurationMs} ms` : "unavailable"} · provider {detail.providerDurationAvailable ? `${detail.providerDurationMs} ms` : "unavailable"}</dd></div>
        <div><dt>Usage</dt><dd>{detail.usageAvailability}{detail.usageAvailability === "reported" ? ` · ${detail.inputTokens} in / ${detail.outputTokens} out / ${detail.totalTokens} total` : ""}{detail.usageSource ? ` · ${detail.usageSource}` : ""}</dd></div>
        <div><dt>HTTP / failure</dt><dd>{detail.httpStatusCode || "unavailable"} · {detail.failureBoundary || "no failure"} · {detail.failureCode || "none"}</dd></div>
        <div><dt>Started / completed</dt><dd>{formatTimestamp(detail.startedAt)} / {detail.completedAt ? formatTimestamp(detail.completedAt) : "not completed"}</dd></div>
        <div><dt>Request / audit</dt><dd>{detail.requestId} / {detail.auditRef}</dd></div>
        <div><dt>Record</dt><dd>{detail.schemaVersion} · version {detail.recordVersion} · {detail.storeMode}{detail.staleStarted ? " · stale started" : ""}</dd></div>
      </dl>
      <p className="boundary-note">Raw input, output, credentials, endpoints, provider envelopes, retry/fallback, billing writes, tools, confirmation, business writes, replay, and resume are not retained or exposed.</p>
    </article>
  );
}

function formatTimestamp(value: string): string {
  if (!value) return "unavailable";
  const timestamp = new Date(value);
  return Number.isNaN(timestamp.valueOf()) ? value : timestamp.toLocaleString();
}
