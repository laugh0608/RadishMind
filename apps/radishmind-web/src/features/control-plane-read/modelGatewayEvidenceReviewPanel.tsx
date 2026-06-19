import type {
  ModelGatewayEvidenceChecklistItem,
  ModelGatewayEvidenceLockedCapability,
  ModelGatewayEvidenceReadinessRollup,
  ModelGatewayEvidenceReviewStatus,
  ModelGatewayEvidenceReviewViewModel,
  ModelGatewayEvidenceRisk,
} from "./modelGatewayEvidenceReview";

type StatusBadgeTone = "good" | "bad" | "neutral";

export function ModelGatewayEvidenceReviewPanel({
  review,
}: {
  review: ModelGatewayEvidenceReviewViewModel;
}) {
  return (
    <section
      className="surface-band model-gateway-evidence-review model-gateway-overview"
      id="model-gateway-evidence-review"
      aria-labelledby="model-gateway-evidence-review-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">Model Gateway</p>
          <h3 id="model-gateway-evidence-review-title">Evidence review and readiness</h3>
        </div>
        <StatusBadge tone={review.canRenderEvidenceReview ? "neutral" : "bad"}>
          {review.canRenderEvidenceReview ? "offline review" : "blocked"}
        </StatusBadge>
      </div>

      <article className="model-gateway-overview-hero">
        <div>
          <p className="eyebrow">{review.reviewMode}</p>
          <h4>{review.tenantRef}</h4>
          <p>{review.reviewNarrative}</p>
        </div>
        <dl className="model-gateway-overview-meta">
          <div>
            <dt>Request</dt>
            <dd>{review.requestId}</dd>
          </div>
          <div>
            <dt>Audit</dt>
            <dd>{review.auditRef}</dd>
          </div>
          <div>
            <dt>Local review</dt>
            <dd>{review.canInspectGatewayEvidenceLocally ? "enabled" : "blocked"}</dd>
          </div>
          <div>
            <dt>Production gateway</dt>
            <dd>{review.canRequestProductionGateway ? "enabled" : "locked"}</dd>
          </div>
        </dl>
      </article>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">Readiness Rollup</p>
          <h4>What is readable and what still needs review</h4>
        </div>
        <div className="model-gateway-overview-metric-grid" aria-label="Model gateway readiness rollup">
          {review.readinessRollup.map((rollup) => (
            <ReadinessRollupCard key={rollup.rollupId} rollup={rollup} />
          ))}
        </div>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">Evidence Checklist</p>
          <h4>Overview, route, usage, audit, and lock coverage</h4>
        </div>
        <div className="model-gateway-overview-policy-grid" aria-label="Model gateway evidence checklist">
          {review.evidenceChecklist.map((item) => (
            <EvidenceChecklistCard key={item.checklistId} item={item} />
          ))}
        </div>
      </div>

      <RiskSection title="Route risks" label="Route" risks={review.routeRisks} />
      <RiskSection title="Usage risks" label="Usage" risks={review.usageRisks} />
      <RiskSection title="Audit risks" label="Audit" risks={review.auditRisks} />

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">Locked Capabilities</p>
          <h4>Capabilities that remain closed</h4>
        </div>
        <div className="model-gateway-overview-lock-grid" aria-label="Model gateway review locked capabilities">
          {review.lockedCapabilities.map((capability) => (
            <LockedCapabilityCard key={capability.capabilityId} capability={capability} />
          ))}
        </div>
      </div>
    </section>
  );
}

function ReadinessRollupCard({ rollup }: { rollup: ModelGatewayEvidenceReadinessRollup }) {
  return (
    <article className="model-gateway-overview-card">
      <span>{rollup.label}</span>
      <strong>{rollup.value}</strong>
      <StatusBadge tone={evidenceReviewTone(rollup.status)}>{rollup.status}</StatusBadge>
      <p className="model-gateway-usage-audit-evidence-source">Source: {rollup.sourceRef}</p>
      <p>{rollup.summary}</p>
    </article>
  );
}

function EvidenceChecklistCard({ item }: { item: ModelGatewayEvidenceChecklistItem }) {
  return (
    <article className="model-gateway-overview-policy">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{item.sourceSurface}</p>
          <h5>{item.label}</h5>
        </div>
        <StatusBadge tone={evidenceReviewTone(item.status)}>{item.status}</StatusBadge>
      </div>
      <dl className="model-gateway-overview-meta">
        <div>
          <dt>Route/page</dt>
          <dd>{item.routeOrPageId}</dd>
        </div>
        <div>
          <dt>Request</dt>
          <dd>{item.requestId}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{item.auditRef}</dd>
        </div>
      </dl>
      <p>{item.summary}</p>
    </article>
  );
}

function RiskSection({
  title,
  label,
  risks,
}: {
  title: string;
  label: string;
  risks: ModelGatewayEvidenceRisk[];
}) {
  return (
    <div className="model-gateway-overview-section">
      <div className="model-gateway-overview-subheading">
        <p className="eyebrow">{label}</p>
        <h4>{title}</h4>
      </div>
      <div className="model-gateway-overview-trace-grid" aria-label={`Model gateway ${label.toLowerCase()} risks`}>
        {risks.map((risk) => (
          <RiskCard key={risk.riskId} risk={risk} />
        ))}
      </div>
    </div>
  );
}

function RiskCard({ risk }: { risk: ModelGatewayEvidenceRisk }) {
  return (
    <article className="model-gateway-overview-trace">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{risk.sourceSurface}</p>
          <h5>{risk.label}</h5>
        </div>
        <StatusBadge tone="bad">{risk.status}</StatusBadge>
      </div>
      <dl className="model-gateway-overview-meta">
        <div>
          <dt>Evidence</dt>
          <dd>{risk.evidenceRef}</dd>
        </div>
      </dl>
      <p>{risk.summary}</p>
      <p>{risk.reviewQuestion}</p>
    </article>
  );
}

function LockedCapabilityCard({
  capability,
}: {
  capability: ModelGatewayEvidenceLockedCapability;
}) {
  return (
    <article className="model-gateway-overview-lock">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{capability.sourceSurface}</p>
          <h5>{capability.label}</h5>
        </div>
        <StatusBadge tone="bad">{capability.status}</StatusBadge>
      </div>
      <dl className="model-gateway-overview-meta">
        <div>
          <dt>Source</dt>
          <dd>{capability.sourceRef}</dd>
        </div>
        <div>
          <dt>Missing prerequisite</dt>
          <dd>{capability.missingPrerequisite}</dd>
        </div>
      </dl>
      <p>{capability.summary}</p>
    </article>
  );
}

function evidenceReviewTone(status: ModelGatewayEvidenceReviewStatus): StatusBadgeTone {
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
