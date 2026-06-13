import type {
  AdminModelRouteReadiness,
  AdminProviderDeploymentLock,
  AdminProviderDeploymentOperatorRisk,
  AdminProviderDeploymentReviewStatus,
  AdminProviderDeploymentReviewViewModel,
  AdminProviderProfileReadiness,
  AdminSecretDeploymentEvidence,
} from "./adminProviderDeploymentReview";

type StatusBadgeTone = "good" | "bad" | "neutral";

export function AdminProviderDeploymentReviewPanel({
  review,
}: {
  review: AdminProviderDeploymentReviewViewModel;
}) {
  return (
    <section
      className="surface-band admin-provider-deployment-review model-gateway-overview"
      id="admin-provider-deployment-review"
      aria-labelledby="admin-provider-deployment-review-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">Admin Control Plane</p>
          <h3 id="admin-provider-deployment-review-title">Provider and deployment review</h3>
        </div>
        <StatusBadge tone={review.canRenderProviderDeploymentReview ? "neutral" : "bad"}>
          {review.canRenderProviderDeploymentReview ? "offline review" : "blocked"}
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
            <dt>Deployment</dt>
            <dd>{review.deploymentStatusRef}</dd>
          </div>
          <div>
            <dt>Request</dt>
            <dd>{review.requestId}</dd>
          </div>
          <div>
            <dt>Audit</dt>
            <dd>{review.auditRef}</dd>
          </div>
          <div>
            <dt>Production gateway</dt>
            <dd>{review.canRequestProductionGateway ? "enabled" : "locked"}</dd>
          </div>
        </dl>
      </article>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">Provider/Profile</p>
          <h4>Credential state, deployment mode, and runtime profile evidence</h4>
        </div>
        <div className="model-gateway-route-evidence-profile-grid" aria-label="Admin provider profile readiness">
          {review.providerProfiles.map((profile) => (
            <ProviderProfileCard key={profile.readinessId} profile={profile} />
          ))}
        </div>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">Model Routes</p>
          <h4>Northbound route binding and selection evidence</h4>
        </div>
        <div className="model-gateway-overview-policy-grid" aria-label="Admin model route readiness">
          {review.modelRoutes.map((route) => (
            <ModelRouteCard key={route.routeId} route={route} />
          ))}
        </div>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">Secrets and Deployment</p>
          <h4>Static readiness evidence and unresolved implementation gates</h4>
        </div>
        <div className="model-gateway-overview-metric-grid" aria-label="Admin secret and deployment evidence">
          {review.secretDeploymentEvidence.map((evidence) => (
            <SecretDeploymentCard key={evidence.evidenceId} evidence={evidence} />
          ))}
        </div>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">Operator Risks</p>
          <h4>Provider, route, secret, deployment, and auth/store risks</h4>
        </div>
        <div className="model-gateway-overview-trace-grid" aria-label="Admin provider deployment operator risks">
          {review.operatorRisks.map((risk) => (
            <OperatorRiskCard key={risk.riskId} risk={risk} />
          ))}
        </div>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">Boundary Locks</p>
          <h4>Capabilities that remain closed before implementation work</h4>
        </div>
        <div className="model-gateway-overview-lock-grid" aria-label="Admin provider deployment boundary locks">
          {review.lockedCapabilities.map((lock) => (
            <BoundaryLockCard key={lock.lockId} lock={lock} />
          ))}
        </div>
      </div>
    </section>
  );
}

function ProviderProfileCard({ profile }: { profile: AdminProviderProfileReadiness }) {
  return (
    <article className="model-gateway-route-evidence-card">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{profile.providerId}</p>
          <h5>{profile.providerProfile}</h5>
        </div>
        <StatusBadge tone={adminProviderDeploymentTone(profile.status)}>{profile.status}</StatusBadge>
      </div>
      <dl className="model-gateway-overview-meta">
        <div>
          <dt>Credential</dt>
          <dd>{profile.credentialState}</dd>
        </div>
        <div>
          <dt>Deployment</dt>
          <dd>{profile.deploymentMode}</dd>
        </div>
        <div>
          <dt>Auth</dt>
          <dd>{profile.authMode}</dd>
        </div>
        <div>
          <dt>Streaming</dt>
          <dd>{profile.streaming}</dd>
        </div>
      </dl>
      <p className="model-gateway-usage-audit-evidence-source">Source: {profile.sourceRef}</p>
      <p>{profile.summary}</p>
    </article>
  );
}

function ModelRouteCard({ route }: { route: AdminModelRouteReadiness }) {
  return (
    <article className="model-gateway-overview-policy">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{route.selectionSource}</p>
          <h5>{route.path}</h5>
        </div>
        <StatusBadge tone={adminProviderDeploymentTone(route.status)}>{route.status}</StatusBadge>
      </div>
      <dl className="model-gateway-overview-meta">
        <div>
          <dt>Route</dt>
          <dd>{route.routeId}</dd>
        </div>
        <div>
          <dt>Profile</dt>
          <dd>{route.selectedProfileId}</dd>
        </div>
        <div>
          <dt>Source</dt>
          <dd>{route.sourceRef}</dd>
        </div>
      </dl>
      <p>{route.summary}</p>
    </article>
  );
}

function SecretDeploymentCard({ evidence }: { evidence: AdminSecretDeploymentEvidence }) {
  return (
    <article className="model-gateway-overview-card">
      <span>{evidence.evidenceKind}</span>
      <strong>{evidence.label}</strong>
      <StatusBadge tone={adminProviderDeploymentTone(evidence.status)}>{evidence.status}</StatusBadge>
      <p className="model-gateway-usage-audit-evidence-source">Value: {evidence.value}</p>
      <p className="model-gateway-usage-audit-evidence-source">Source: {evidence.sourceRef}</p>
      <p>{evidence.summary}</p>
    </article>
  );
}

function OperatorRiskCard({ risk }: { risk: AdminProviderDeploymentOperatorRisk }) {
  return (
    <article className="model-gateway-overview-trace">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{risk.riskSurface}</p>
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

function BoundaryLockCard({ lock }: { lock: AdminProviderDeploymentLock }) {
  return (
    <article className="model-gateway-overview-lock">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{lock.lockSurface}</p>
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

function adminProviderDeploymentTone(status: AdminProviderDeploymentReviewStatus): StatusBadgeTone {
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
