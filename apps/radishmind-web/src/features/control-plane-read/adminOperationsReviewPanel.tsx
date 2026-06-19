import type {
  AdminOperationsBoundaryLock,
  AdminOperationsEvidenceChecklistItem,
  AdminOperationsReadinessRollup,
  AdminOperationsReviewStatus,
  AdminOperationsReviewViewModel,
  AdminOperationsRisk,
} from "./adminOperationsReview";

type StatusBadgeTone = "good" | "bad" | "neutral";

export function AdminOperationsReviewPanel({
  review,
}: {
  review: AdminOperationsReviewViewModel;
}) {
  return (
    <section
      className="surface-band admin-operations-review model-gateway-overview"
      id="admin-operations-review"
      aria-labelledby="admin-operations-review-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">Admin Control Plane</p>
          <h3 id="admin-operations-review-title">Operations review and readiness</h3>
        </div>
        <StatusBadge tone={review.canRenderAdminOperationsReview ? "neutral" : "bad"}>
          {review.canRenderAdminOperationsReview ? "offline review" : "blocked"}
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
            <dd>{review.canInspectAdminEvidenceLocally ? "enabled" : "blocked"}</dd>
          </div>
          <div>
            <dt>Production backend</dt>
            <dd>{review.canRequestProductionBackend ? "enabled" : "locked"}</dd>
          </div>
        </dl>
      </article>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">Readiness Rollup</p>
          <h4>Admin evidence and production boundaries</h4>
        </div>
        <div className="model-gateway-overview-metric-grid" aria-label="Admin operations readiness rollup">
          {review.readinessRollup.map((rollup) => (
            <ReadinessRollupCard key={rollup.rollupId} rollup={rollup} />
          ))}
        </div>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">Evidence Checklist</p>
          <h4>Tenant, audit, gateway, read contract, and production ops evidence</h4>
        </div>
        <div className="model-gateway-overview-policy-grid" aria-label="Admin operations evidence checklist">
          {review.evidenceChecklist.map((item) => (
            <EvidenceChecklistCard key={item.checklistId} item={item} />
          ))}
        </div>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">Operational Risks</p>
          <h4>What needs operator review before production work</h4>
        </div>
        <div className="model-gateway-overview-trace-grid" aria-label="Admin operations risks">
          {review.operationalRisks.map((risk) => (
            <OperationalRiskCard key={risk.riskId} risk={risk} />
          ))}
        </div>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">Boundary Locks</p>
          <h4>Capabilities that stay closed</h4>
        </div>
        <div className="model-gateway-overview-lock-grid" aria-label="Admin operations boundary locks">
          {review.boundaryLocks.map((lock) => (
            <BoundaryLockCard key={lock.lockId} lock={lock} />
          ))}
        </div>
      </div>
    </section>
  );
}

function ReadinessRollupCard({ rollup }: { rollup: AdminOperationsReadinessRollup }) {
  return (
    <article className="model-gateway-overview-card">
      <span>{rollup.label}</span>
      <strong>{rollup.value}</strong>
      <StatusBadge tone={adminOperationsTone(rollup.status)}>{rollup.status}</StatusBadge>
      <p className="model-gateway-usage-audit-evidence-source">Source: {rollup.sourceRef}</p>
      <p>{rollup.summary}</p>
    </article>
  );
}

function EvidenceChecklistCard({ item }: { item: AdminOperationsEvidenceChecklistItem }) {
  return (
    <article className="model-gateway-overview-policy">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{item.sourceSurface}</p>
          <h5>{item.label}</h5>
        </div>
        <StatusBadge tone={adminOperationsTone(item.status)}>{item.status}</StatusBadge>
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

function OperationalRiskCard({ risk }: { risk: AdminOperationsRisk }) {
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
      <p>{risk.operatorQuestion}</p>
    </article>
  );
}

function BoundaryLockCard({ lock }: { lock: AdminOperationsBoundaryLock }) {
  return (
    <article className="model-gateway-overview-lock">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{lock.sourceSurface}</p>
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

function adminOperationsTone(status: AdminOperationsReviewStatus): StatusBadgeTone {
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
