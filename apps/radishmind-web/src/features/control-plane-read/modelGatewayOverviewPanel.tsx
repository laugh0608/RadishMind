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
  return (
    <section
      className="surface-band model-gateway-overview"
      id="model-gateway-overview"
      aria-labelledby="model-gateway-overview-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">Model Gateway</p>
          <h3 id="model-gateway-overview-title">API distribution evidence</h3>
        </div>
        <StatusBadge tone={overview.canRenderModelGatewayOverview ? "neutral" : "bad"}>
          {overview.canRenderModelGatewayOverview ? "offline evidence" : "blocked"}
        </StatusBadge>
      </div>

      <article className="model-gateway-overview-hero">
        <div>
          <p className="eyebrow">{overview.overviewMode}</p>
          <h4>{overview.tenantRef}</h4>
          <p>{overview.gatewayNarrative}</p>
        </div>
        <dl className="model-gateway-overview-meta">
          <div>
            <dt>Request</dt>
            <dd>{overview.requestId}</dd>
          </div>
          <div>
            <dt>Audit</dt>
            <dd>{overview.auditRef}</dd>
          </div>
          <div>
            <dt>Dev evidence</dt>
            <dd>{overview.canConsumeDevReadEvidence ? "available" : "offline fixtures"}</dd>
          </div>
          <div>
            <dt>Production gateway</dt>
            <dd>{overview.canRequestProductionGateway ? "enabled" : "locked"}</dd>
          </div>
        </dl>
      </article>

      <div className="model-gateway-overview-metric-grid" aria-label="Model gateway overview metrics">
        {overview.metrics.map((metric) => (
          <ModelGatewayMetricCard key={metric.metricId} metric={metric} />
        ))}
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">Northbound API</p>
          <h4>Compatibility surfaces</h4>
        </div>
        <div className="model-gateway-overview-surface-grid" aria-label="Model gateway API surfaces">
          {overview.apiSurfaces.map((surface) => (
            <ModelGatewayApiSurfaceCard key={surface.surfaceId} surface={surface} />
          ))}
        </div>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">Policy Evidence</p>
          <h4>Keys, quota, trace, audit</h4>
        </div>
        <div className="model-gateway-overview-policy-grid" aria-label="Model gateway policy evidence">
          {overview.policyEvidence.map((evidence) => (
            <ModelGatewayPolicyEvidenceCard key={evidence.evidenceId} evidence={evidence} />
          ))}
        </div>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">Trace Evidence</p>
          <h4>Read-side usage records</h4>
        </div>
        <div className="model-gateway-overview-trace-grid" aria-label="Model gateway trace evidence">
          {overview.traceEvidence.map((trace) => (
            <ModelGatewayTraceEvidenceCard key={trace.traceId} trace={trace} />
          ))}
        </div>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">Stop Lines</p>
          <h4>Locked distribution capabilities</h4>
        </div>
        <div className="model-gateway-overview-lock-grid" aria-label="Model gateway boundary locks">
          {overview.boundaryLocks.map((lock) => (
            <ModelGatewayBoundaryLockCard key={lock.lockId} lock={lock} />
          ))}
        </div>
      </div>
    </section>
  );
}

function ModelGatewayMetricCard({ metric }: { metric: ModelGatewayOverviewMetric }) {
  return (
    <article className="model-gateway-overview-card">
      <span>{metric.label}</span>
      <strong>{metric.value}</strong>
      <StatusBadge tone={modelGatewayTone(metric.status)}>{metric.status}</StatusBadge>
      <p>{metric.summary}</p>
    </article>
  );
}

function ModelGatewayApiSurfaceCard({ surface }: { surface: ModelGatewayApiSurface }) {
  return (
    <article className="model-gateway-overview-surface">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{surface.surfaceId}</p>
          <h5>{surface.label}</h5>
        </div>
        <StatusBadge tone={modelGatewayTone(surface.status)}>{surface.status}</StatusBadge>
      </div>
      <p className="route-path">{surface.path}</p>
      <dl className="model-gateway-overview-meta">
        <div>
          <dt>Evidence</dt>
          <dd>{surface.evidenceRef}</dd>
        </div>
      </dl>
      <p>{surface.summary}</p>
    </article>
  );
}

function ModelGatewayPolicyEvidenceCard({ evidence }: { evidence: ModelGatewayPolicyEvidence }) {
  return (
    <article className="model-gateway-overview-policy">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{evidence.evidenceId}</p>
          <h5>{evidence.label}</h5>
        </div>
        <StatusBadge tone={modelGatewayTone(evidence.status)}>{evidence.status}</StatusBadge>
      </div>
      <strong>{evidence.value}</strong>
      <dl className="model-gateway-overview-meta">
        <div>
          <dt>Source</dt>
          <dd>{evidence.sourceRef}</dd>
        </div>
      </dl>
      <p>{evidence.summary}</p>
    </article>
  );
}

function ModelGatewayTraceEvidenceCard({ trace }: { trace: ModelGatewayTraceEvidence }) {
  return (
    <article className="model-gateway-overview-trace">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{trace.runId}</p>
          <h5>{trace.traceId}</h5>
        </div>
        <StatusBadge tone={modelGatewayTone(trace.status)}>{trace.status}</StatusBadge>
      </div>
      <dl className="model-gateway-overview-meta">
        <div>
          <dt>Cost</dt>
          <dd>{trace.cost}</dd>
        </div>
        <div>
          <dt>Failure</dt>
          <dd>{trace.failureCode}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{trace.auditRef}</dd>
        </div>
      </dl>
      <p>{trace.summary}</p>
    </article>
  );
}

function ModelGatewayBoundaryLockCard({ lock }: { lock: ModelGatewayBoundaryLock }) {
  return (
    <article className="model-gateway-overview-lock">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{lock.lockId}</p>
          <h5>{lock.label}</h5>
        </div>
        <StatusBadge tone="bad">{lock.status}</StatusBadge>
      </div>
      <dl className="model-gateway-overview-meta">
        <div>
          <dt>Missing prerequisite</dt>
          <dd>{lock.missingPrerequisite}</dd>
        </div>
      </dl>
      <p>{lock.summary}</p>
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
