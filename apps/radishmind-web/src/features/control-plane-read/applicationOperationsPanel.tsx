import { useEffect, useState } from "react";

import {
  initialApplicationOperationsState,
  loadApplicationOperations,
  type ApplicationOperationsState,
  type ApplicationOperationsTimelineEntry,
} from "./applicationOperationsConsumer.ts";
import { readModelGatewayRequestHistoryConfig } from "./modelGatewayRequestHistoryConsumer.ts";
import { readWorkflowExecutorConsumerConfig } from "./workflowExecutorConsumer.ts";

const gatewayConfig = readModelGatewayRequestHistoryConfig();
const workflowConfig = readWorkflowExecutorConsumerConfig();

export default function ApplicationOperationsPanel({
  applicationId,
  applicationName,
}: {
  applicationId: string;
  applicationName: string;
}) {
  const [refreshKey, setRefreshKey] = useState(0);
  const [state, setState] = useState<ApplicationOperationsState>(() =>
    initialApplicationOperationsState(applicationId, gatewayConfig, workflowConfig)
  );

  useEffect(() => {
    let cancelled = false;
    setState(initialApplicationOperationsState(applicationId, gatewayConfig, workflowConfig));
    void loadApplicationOperations(applicationId, gatewayConfig, workflowConfig).then((nextState) => {
      if (!cancelled) setState(nextState);
    });
    return () => {
      cancelled = true;
    };
  }, [applicationId, refreshKey]);

  const metrics = state.metrics;
  const refreshDisabled = state.status === "loading" || state.status === "offline" ||
    state.status === "application_unavailable";

  return (
    <section className="surface-band application-operations" id="application-operations" aria-labelledby="application-operations-title">
      <div className="section-heading">
        <div>
          <p className="eyebrow">User Workspace · Application Operations</p>
          <h3 id="application-operations-title">运行观测与用量归因</h3>
          <p>
            {applicationName || "No selected application"} · <code>{state.applicationId || "application unavailable"}</code>
          </p>
        </div>
        <div className="application-operations-actions">
          <StatusBadge status={state.status} />
          <button type="button" className="secondary-action" disabled={refreshDisabled} onClick={() => setRefreshKey((key) => key + 1)}>
            Refresh observations
          </button>
        </div>
      </div>

      <div className="application-operations-coverage" aria-label="Application operations source coverage">
        <ChannelCoverage
          label="Gateway requests"
          status={state.gateway.status}
          loaded={metrics.gatewayLoaded}
          hasMore={state.gateway.hasMore}
          requestId={state.gateway.requestId}
          auditRef={state.gateway.auditRef}
          failureCode={state.gateway.failureCode}
        />
        <ChannelCoverage
          label="Workflow runs"
          status={state.workflow.status}
          loaded={metrics.workflowLoaded}
          hasMore={state.workflow.hasMore}
          requestId={state.workflow.requestId}
          auditRef={state.workflow.auditRef}
          failureCode={state.workflow.failureCode}
        />
      </div>

      {state.failureSummary && <p className="application-operations-failure" role="alert">{state.failureSummary}</p>}

      <div className="application-operations-metrics" aria-label="Application operations attribution summary">
        <MetricCard
          label="Gateway status"
          value={`${metrics.gatewaySucceeded} succeeded`}
          detail={`${metrics.gatewayFailed} failed · ${metrics.gatewayCanceled} canceled · ${metrics.gatewayStarted} started`}
        />
        <MetricCard
          label="Gateway usage availability"
          value={`${metrics.gatewayUsageReported} reported`}
          detail={`${metrics.gatewayUsageNotReported} not reported · ${metrics.gatewayUsageNotApplicable} not applicable`}
        />
        <MetricCard
          label="Workflow status"
          value={`${metrics.workflowSucceeded} succeeded`}
          detail={`${metrics.workflowFailed} failed · ${metrics.workflowCanceled} canceled · ${metrics.workflowRunning} running · ${metrics.workflowOutcomeUnknown} unknown`}
        />
        <MetricCard
          label="Workflow observed calls"
          value={`${metrics.workflowProviderCalls} provider · ${metrics.workflowRetrievalCalls} retrieval`}
          detail={`${metrics.workflowToolCalls} tool · ${metrics.workflowConfirmationCalls} confirmation`}
        />
      </div>

      {(metrics.workflowBusinessWrites > 0 || metrics.workflowReplayWrites > 0) && (
        <p className="application-operations-stop-line" role="alert">
          Stop-line violation observed: {metrics.workflowBusinessWrites} business writes and {metrics.workflowReplayWrites} replay writes.
        </p>
      )}

      <div className="application-operations-timeline-heading">
        <div>
          <p className="eyebrow">Loaded-window timeline</p>
          <h4>Gateway 与 Workflow 独立观测记录</h4>
        </div>
        <span>{state.loadedWindowComplete ? "current windows complete" : "more records available"}</span>
      </div>

      {state.timeline.length > 0 ? (
        <ol className="application-operations-timeline">
          {state.timeline.map((entry) => <TimelineEntry key={`${entry.source}:${entry.recordId}`} entry={entry} />)}
        </ol>
      ) : (
        <p className="application-operations-empty">
          {emptyMessage(state)}
        </p>
      )}

      <p className="boundary-note">
        两个通道只在应用作用域下并列展示，不建立一对一关联。当前数字只覆盖已加载窗口，不是全量 usage、token、cost、quota 或 billing；输入、回答、凭据和 provider 原始材料不会进入该视图。
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
  return (
    <article>
      <div className="card-title-row"><h4>{label}</h4><span>{status}</span></div>
      <p><strong>{loaded}</strong> loaded · {hasMore ? "more available" : "window complete"}</p>
      <dl>
        <div><dt>Request</dt><dd>{requestId}</dd></div>
        <div><dt>Audit</dt><dd>{auditRef}</dd></div>
        <div><dt>Failure</dt><dd>{failureCode || "none"}</dd></div>
      </dl>
    </article>
  );
}

function MetricCard({ label, value, detail }: { label: string; value: string; detail: string }) {
  return <article><p>{label}</p><strong>{value}</strong><span>{detail}</span></article>;
}

function TimelineEntry({ entry }: { entry: ApplicationOperationsTimelineEntry }) {
  return (
    <li>
      <div className="application-operations-timeline-marker" aria-hidden="true" />
      <article>
        <div className="card-title-row">
          <div><p className="eyebrow">{entry.source.replace("_", " ")}</p><h4>{entry.operation || "unavailable"}</h4></div>
          <span>{entry.status}</span>
        </div>
        <p><code>{entry.recordId}</code> · {formatTimestamp(entry.startedAt)} · {entry.durationMs} ms</p>
        <dl>
          <div><dt>Contract</dt><dd>{entry.contract || "unavailable"}</dd></div>
          <div><dt>Route</dt><dd>{entry.provider || "unavailable"} / {entry.profile || "default"} / {entry.model || "unavailable"}</dd></div>
          <div><dt>Failure</dt><dd>{entry.failureCode || "none"} · {entry.failureBoundary || "none"}</dd></div>
          <div><dt>Request / audit</dt><dd>{entry.requestId} · {entry.auditRef}</dd></div>
          {entry.source === "gateway_request" ? (
            <div><dt>Usage</dt><dd>{entry.usageAvailability}</dd></div>
          ) : (
            <div><dt>Calls</dt><dd>{entry.providerCalls} provider · {entry.retrievalCalls} retrieval · {entry.toolCalls} tool</dd></div>
          )}
        </dl>
      </article>
    </li>
  );
}

function StatusBadge({ status }: { status: ApplicationOperationsState["status"] }) {
  const tone = status === "ready" || status === "empty" ? "good" :
    status === "failed" || status === "application_unavailable" ? "bad" : "neutral";
  return <span className={`status-badge ${tone}`}>{status.replaceAll("_", " ")}</span>;
}

function emptyMessage(state: ApplicationOperationsState): string {
  if (state.status === "offline") return "Offline mode keeps both observation sources at zero requests.";
  if (state.status === "application_unavailable") return "Select an application before loading operations.";
  if (state.status === "loading") return "Loading application-scoped observations…";
  if (state.status === "failed") return "Both enabled observation sources failed closed.";
  return "No records are available in the current observation windows.";
}

function formatTimestamp(value: string): string {
  if (!value) return "unavailable";
  const timestamp = new Date(value);
  return Number.isNaN(timestamp.valueOf()) ? value : timestamp.toLocaleString();
}
