import "../../i18n/gatewayOverviewEvidenceResources.ts";
import { useTranslation } from "react-i18next";
import { evidenceProjectionText, evidenceProjectionStatus } from "./evidenceProjectionCopy.ts";
import type {
  ModelGatewayApiSurface,
  ModelGatewayBoundaryLock,
  ModelGatewayOverviewMetric,
  ModelGatewayOverviewStatus,
  ModelGatewayOverviewViewModel,
  ModelGatewayPolicyEvidence,
  ModelGatewayTraceEvidence,
} from "./modelGatewayOverview";

type StatusBadgeTone = "good" | "bad" | "neutral";

export function ModelGatewayOverviewPanel({
  overview,
}: {
  overview: ModelGatewayOverviewViewModel;
}) {
  const { t } = useTranslation("evaluation");
  return (
    <section
      className="surface-band model-gateway-overview"
      id="model-gateway-overview"
      aria-labelledby="model-gateway-overview-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">{t($ => $.auxiliary.gatewayOverviewEvidence_ui_Model_Gateway)}</p>
          <h3 id="model-gateway-overview-title">{t($ => $.auxiliary.gatewayOverviewEvidence_ui_API_distribution_evidence)}</h3>
        </div>
        <StatusBadge tone={overview.canRenderModelGatewayOverview ? "neutral" : "bad"}>
          {overview.canRenderModelGatewayOverview ? t($ => $.auxiliary.gatewayOverviewEvidence_ui_offline_evidence) : t($ => $.auxiliary.gatewayOverviewEvidence_ui_blocked)}
        </StatusBadge>
      </div>

      <article className="model-gateway-overview-hero">
        <div>
          <p className="eyebrow">{evidenceProjectionStatus(t, overview.overviewMode)}</p>
          <h4>{overview.tenantRef}</h4>
          <p>{evidenceProjectionText(t, overview, "gatewayNarrative")}</p>
        </div>
        <dl className="model-gateway-overview-meta">
          <div>
            <dt>{t($ => $.auxiliary.gatewayOverviewEvidence_ui_Request)}</dt>
            <dd>{overview.requestId}</dd>
          </div>
          <div>
            <dt>{t($ => $.auxiliary.gatewayOverviewEvidence_ui_Audit)}</dt>
            <dd>{overview.auditRef}</dd>
          </div>
          <div>
            <dt>{t($ => $.auxiliary.gatewayOverviewEvidence_ui_Dev_evidence)}</dt>
            <dd>{overview.canConsumeDevReadEvidence ? t($ => $.auxiliary.gatewayOverviewEvidence_ui_available) : t($ => $.auxiliary.gatewayOverviewEvidence_ui_offline_fixtures)}</dd>
          </div>
          <div>
            <dt>{t($ => $.auxiliary.gatewayOverviewEvidence_ui_Production_gateway)}</dt>
            <dd>{overview.canRequestProductionGateway ? t($ => $.auxiliary.gatewayOverviewEvidence_ui_enabled) : t($ => $.auxiliary.gatewayOverviewEvidence_ui_locked)}</dd>
          </div>
        </dl>
      </article>

      <div className="model-gateway-overview-metric-grid" aria-label={t($ => $.auxiliary.gatewayOverviewEvidence_ui_Model_gateway_overview_metrics)}>
        {overview.metrics.map((metric) => (
          <ModelGatewayMetricCard key={metric.metricId} metric={metric} />
        ))}
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.gatewayOverviewEvidence_ui_Northbound_API)}</p>
          <h4>{t($ => $.auxiliary.gatewayOverviewEvidence_ui_Compatibility_surfaces)}</h4>
        </div>
        <div className="model-gateway-overview-surface-grid" aria-label={t($ => $.auxiliary.gatewayOverviewEvidence_ui_Model_gateway_API_surfaces)}>
          {overview.apiSurfaces.map((surface) => (
            <ModelGatewayApiSurfaceCard key={surface.surfaceId} surface={surface} />
          ))}
        </div>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.gatewayOverviewEvidence_ui_Policy_Evidence)}</p>
          <h4>{t($ => $.auxiliary.gatewayOverviewEvidence_ui_Keys_quota_trace_audit)}</h4>
        </div>
        <div className="model-gateway-overview-policy-grid" aria-label={t($ => $.auxiliary.gatewayOverviewEvidence_ui_Model_gateway_policy_evidence)}>
          {overview.policyEvidence.map((evidence) => (
            <ModelGatewayPolicyEvidenceCard key={evidence.evidenceId} evidence={evidence} />
          ))}
        </div>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.gatewayOverviewEvidence_ui_Trace_Evidence)}</p>
          <h4>{t($ => $.auxiliary.gatewayOverviewEvidence_ui_Read_side_usage_records)}</h4>
        </div>
        <div className="model-gateway-overview-trace-grid" aria-label={t($ => $.auxiliary.gatewayOverviewEvidence_ui_Model_gateway_trace_evidence)}>
          {overview.traceEvidence.map((trace) => (
            <ModelGatewayTraceEvidenceCard key={`${trace.traceId}:${trace.runId}`} trace={trace} />
          ))}
        </div>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.gatewayOverviewEvidence_ui_Stop_Lines)}</p>
          <h4>{t($ => $.auxiliary.gatewayOverviewEvidence_ui_Locked_distribution_capabilities)}</h4>
        </div>
        <div className="model-gateway-overview-lock-grid" aria-label={t($ => $.auxiliary.gatewayOverviewEvidence_ui_Model_gateway_boundary_locks)}>
          {overview.boundaryLocks.map((lock) => (
            <ModelGatewayBoundaryLockCard key={lock.lockId} lock={lock} />
          ))}
        </div>
      </div>
    </section>
  );
}

function ModelGatewayMetricCard({ metric }: { metric: ModelGatewayOverviewMetric }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-overview-card">
      <span>{evidenceProjectionText(t, metric, "label")}</span>
      <strong>{evidenceProjectionText(t, metric, "value")}</strong>
      <StatusBadge tone={modelGatewayTone(metric.status)}>{evidenceProjectionStatus(t, metric.status)}</StatusBadge>
      <p>{evidenceProjectionText(t, metric, "summary")}</p>
    </article>
  );
}

function ModelGatewayApiSurfaceCard({ surface }: { surface: ModelGatewayApiSurface }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-overview-surface">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{surface.surfaceId}</p>
          <h5>{evidenceProjectionText(t, surface, "label")}</h5>
        </div>
        <StatusBadge tone={modelGatewayTone(surface.status)}>{evidenceProjectionStatus(t, surface.status)}</StatusBadge>
      </div>
      <p className="route-path">{surface.path}</p>
      <dl className="model-gateway-overview-meta">
        <div>
          <dt>{t($ => $.auxiliary.gatewayOverviewEvidence_ui_Evidence)}</dt>
          <dd>{surface.evidenceRef}</dd>
        </div>
      </dl>
      <p>{evidenceProjectionText(t, surface, "summary")}</p>
    </article>
  );
}

function ModelGatewayPolicyEvidenceCard({ evidence }: { evidence: ModelGatewayPolicyEvidence }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-overview-policy">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{evidence.evidenceId}</p>
          <h5>{evidenceProjectionText(t, evidence, "label")}</h5>
        </div>
        <StatusBadge tone={modelGatewayTone(evidence.status)}>{evidenceProjectionStatus(t, evidence.status)}</StatusBadge>
      </div>
      <strong>{evidenceProjectionText(t, evidence, "value")}</strong>
      <dl className="model-gateway-overview-meta">
        <div>
          <dt>{t($ => $.auxiliary.gatewayOverviewEvidence_ui_Source)}</dt>
          <dd>{evidence.sourceRef}</dd>
        </div>
      </dl>
      <p>{evidenceProjectionText(t, evidence, "summary")}</p>
    </article>
  );
}

function ModelGatewayTraceEvidenceCard({ trace }: { trace: ModelGatewayTraceEvidence }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-overview-trace">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{trace.runId}</p>
          <h5>{trace.traceId}</h5>
        </div>
        <StatusBadge tone={modelGatewayTone(trace.status)}>{evidenceProjectionStatus(t, trace.status)}</StatusBadge>
      </div>
      <dl className="model-gateway-overview-meta">
        <div>
          <dt>{t($ => $.auxiliary.gatewayOverviewEvidence_ui_Cost)}</dt>
          <dd>{trace.cost}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayOverviewEvidence_ui_Failure)}</dt>
          <dd>{trace.failureCode}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayOverviewEvidence_ui_Audit)}</dt>
          <dd>{trace.auditRef}</dd>
        </div>
      </dl>
      <p>{evidenceProjectionText(t, trace, "summary")}</p>
    </article>
  );
}

function ModelGatewayBoundaryLockCard({ lock }: { lock: ModelGatewayBoundaryLock }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-overview-lock">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{lock.lockId}</p>
          <h5>{evidenceProjectionText(t, lock, "label")}</h5>
        </div>
        <StatusBadge tone="bad">{evidenceProjectionStatus(t, lock.status)}</StatusBadge>
      </div>
      <dl className="model-gateway-overview-meta">
        <div>
          <dt>{t($ => $.auxiliary.gatewayOverviewEvidence_ui_Missing_prerequisite)}</dt>
          <dd>{evidenceProjectionText(t, lock, "missingPrerequisite")}</dd>
        </div>
      </dl>
      <p>{evidenceProjectionText(t, lock, "summary")}</p>
    </article>
  );
}

function modelGatewayTone(status: ModelGatewayOverviewStatus): StatusBadgeTone {
  if (status === "ready") {
    return "good";
  }
  if (status === "blocked" || status === "locked" || status === "review_required") {
    return "bad";
  }
  return "neutral";
}

function StatusBadge({ children, tone }: { children: string; tone: StatusBadgeTone }) {
  return <span className={`status-badge ${tone}`}>{children}</span>;
}
