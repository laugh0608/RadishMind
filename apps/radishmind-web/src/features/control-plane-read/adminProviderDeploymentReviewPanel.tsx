import "../../i18n/adminDeploymentEvidenceResources.ts";
import { useTranslation } from "react-i18next";
import { evidenceProjectionText, evidenceProjectionStatus } from "./evidenceProjectionCopy.ts";
import { lazy, Suspense } from "react";

import type {
  AdminModelRouteReadiness,
  AdminProviderDeploymentLock,
  AdminProviderDeploymentOperatorRisk,
  AdminProviderDeploymentReviewStatus,
  AdminProviderDeploymentReviewViewModel,
  AdminProviderProfileReadiness,
  AdminSecretDeploymentEvidence,
} from "./adminProviderDeploymentReview";

const AdminProviderRouteWorkspacePanel = lazy(() =>
  import("./adminProviderRouteWorkspacePanel.tsx").then((module) => ({
    default: module.AdminProviderRouteWorkspacePanel,
  })),
);

type StatusBadgeTone = "good" | "bad" | "neutral";

export function AdminProviderDeploymentReviewPanel({
  review,
  includeControlledWorkspace = true,
}: {
  review: AdminProviderDeploymentReviewViewModel;
  includeControlledWorkspace?: boolean;
}) {
  const { t } = useTranslation("evaluation");
  return (
    <section
      className="surface-band admin-provider-deployment-review model-gateway-overview"
      id="admin-provider-deployment-review"
      aria-labelledby="admin-provider-deployment-review-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">{t($ => $.auxiliary.adminDeploymentEvidence_ui_Admin_Control_Plane)}</p>
          <h3 id="admin-provider-deployment-review-title">{t($ => $.auxiliary.adminDeploymentEvidence_ui_Provider_and_deployment_review)}</h3>
        </div>
        <StatusBadge tone={review.canRenderProviderDeploymentReview ? "neutral" : "bad"}>
          {review.canRenderProviderDeploymentReview ? t($ => $.auxiliary.adminDeploymentEvidence_ui_offline_review) : t($ => $.auxiliary.adminDeploymentEvidence_ui_blocked)}
        </StatusBadge>
      </div>

      <article className="model-gateway-overview-hero">
        <div>
          <p className="eyebrow">{evidenceProjectionStatus(t, review.reviewMode)}</p>
          <h4>{review.tenantRef}</h4>
          <p>{evidenceProjectionText(t, review, "reviewNarrative")}</p>
        </div>
        <dl className="model-gateway-overview-meta">
          <div>
            <dt>{t($ => $.auxiliary.adminDeploymentEvidence_ui_Deployment)}</dt>
            <dd>{review.deploymentStatusRef}</dd>
          </div>
          <div>
            <dt>{t($ => $.auxiliary.adminDeploymentEvidence_ui_Request)}</dt>
            <dd>{review.requestId}</dd>
          </div>
          <div>
            <dt>{t($ => $.auxiliary.adminDeploymentEvidence_ui_Audit)}</dt>
            <dd>{review.auditRef}</dd>
          </div>
          <div>
            <dt>{t($ => $.auxiliary.adminDeploymentEvidence_ui_Production_gateway)}</dt>
            <dd>{review.canRequestProductionGateway ? t($ => $.auxiliary.adminDeploymentEvidence_ui_enabled) : t($ => $.auxiliary.adminDeploymentEvidence_ui_locked)}</dd>
          </div>
        </dl>
      </article>

      {includeControlledWorkspace ? (
        <Suspense fallback={<div className="admin-provider-route-workspace"><p>{t($ => $.auxiliary.adminDeploymentEvidence_ui_Loading_Provider_route_workspace)}</p></div>}>
          <AdminProviderRouteWorkspacePanel />
        </Suspense>
      ) : null}

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.adminDeploymentEvidence_ui_Provider_Profile)}</p>
          <h4>{t($ => $.auxiliary.adminDeploymentEvidence_ui_Credential_state_deployment_mode_and_runtime_profile_evidence)}</h4>
        </div>
        <div className="model-gateway-route-evidence-profile-grid" aria-label={t($ => $.auxiliary.adminDeploymentEvidence_ui_Admin_provider_profile_readiness)}>
          {review.providerProfiles.map((profile) => (
            <ProviderProfileCard key={profile.readinessId} profile={profile} />
          ))}
        </div>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.adminDeploymentEvidence_ui_Model_Routes)}</p>
          <h4>{t($ => $.auxiliary.adminDeploymentEvidence_ui_Northbound_route_binding_and_selection_evidence)}</h4>
        </div>
        <div className="model-gateway-overview-policy-grid" aria-label={t($ => $.auxiliary.adminDeploymentEvidence_ui_Admin_model_route_readiness)}>
          {review.modelRoutes.map((route) => (
            <ModelRouteCard key={route.routeId} route={route} />
          ))}
        </div>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.adminDeploymentEvidence_ui_Secrets_and_Deployment)}</p>
          <h4>{t($ => $.auxiliary.adminDeploymentEvidence_ui_Static_readiness_evidence_and_unresolved_implementation_gates)}</h4>
        </div>
        <div className="model-gateway-overview-metric-grid" aria-label={t($ => $.auxiliary.adminDeploymentEvidence_ui_Admin_secret_and_deployment_evidence)}>
          {review.secretDeploymentEvidence.map((evidence) => (
            <SecretDeploymentCard key={evidence.evidenceId} evidence={evidence} />
          ))}
        </div>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.adminDeploymentEvidence_ui_Operator_Risks)}</p>
          <h4>{t($ => $.auxiliary.adminDeploymentEvidence_ui_Provider_route_secret_deployment_and_auth_store_risks)}</h4>
        </div>
        <div className="model-gateway-overview-trace-grid" aria-label={t($ => $.auxiliary.adminDeploymentEvidence_ui_Admin_provider_deployment_operator_risks)}>
          {review.operatorRisks.map((risk) => (
            <OperatorRiskCard key={risk.riskId} risk={risk} />
          ))}
        </div>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.adminDeploymentEvidence_ui_Boundary_Locks)}</p>
          <h4>{t($ => $.auxiliary.adminDeploymentEvidence_ui_Capabilities_that_remain_closed_before_implementation_work)}</h4>
        </div>
        <div className="model-gateway-overview-lock-grid" aria-label={t($ => $.auxiliary.adminDeploymentEvidence_ui_Admin_provider_deployment_boundary_locks)}>
          {review.lockedCapabilities.map((lock) => (
            <BoundaryLockCard key={lock.lockId} lock={lock} />
          ))}
        </div>
      </div>
    </section>
  );
}

function ProviderProfileCard({ profile }: { profile: AdminProviderProfileReadiness }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-route-evidence-card">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{profile.providerId}</p>
          <h5>{profile.providerProfile}</h5>
        </div>
        <StatusBadge tone={adminProviderDeploymentTone(profile.status)}>{evidenceProjectionStatus(t, profile.status)}</StatusBadge>
      </div>
      <dl className="model-gateway-overview-meta">
        <div>
          <dt>{t($ => $.auxiliary.adminDeploymentEvidence_ui_Credential)}</dt>
          <dd>{evidenceProjectionStatus(t, profile.credentialState)}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.adminDeploymentEvidence_ui_Deployment)}</dt>
          <dd>{evidenceProjectionStatus(t, profile.deploymentMode)}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.adminDeploymentEvidence_ui_Auth)}</dt>
          <dd>{evidenceProjectionStatus(t, profile.authMode)}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.adminDeploymentEvidence_ui_Streaming)}</dt>
          <dd>{evidenceProjectionStatus(t, profile.streaming)}</dd>
        </div>
      </dl>
      <p className="model-gateway-usage-audit-evidence-source">{t($ => $.auxiliary.adminDeploymentEvidence_ui_Source)}{profile.sourceRef}</p>
      <p>{evidenceProjectionText(t, profile, "summary")}</p>
    </article>
  );
}

function ModelRouteCard({ route }: { route: AdminModelRouteReadiness }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-overview-policy">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{evidenceProjectionStatus(t, route.selectionSource)}</p>
          <h5>{route.path}</h5>
        </div>
        <StatusBadge tone={adminProviderDeploymentTone(route.status)}>{evidenceProjectionStatus(t, route.status)}</StatusBadge>
      </div>
      <dl className="model-gateway-overview-meta">
        <div>
          <dt>{t($ => $.auxiliary.adminDeploymentEvidence_ui_Route)}</dt>
          <dd>{route.routeId}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.adminDeploymentEvidence_ui_Profile)}</dt>
          <dd>{route.selectedProfileId}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.adminDeploymentEvidence_ui_Source_2)}</dt>
          <dd>{route.sourceRef}</dd>
        </div>
      </dl>
      <p>{evidenceProjectionText(t, route, "summary")}</p>
    </article>
  );
}

function SecretDeploymentCard({ evidence }: { evidence: AdminSecretDeploymentEvidence }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-overview-card">
      <span>{evidenceProjectionStatus(t, evidence.evidenceKind)}</span>
      <strong>{evidenceProjectionText(t, evidence, "label")}</strong>
      <StatusBadge tone={adminProviderDeploymentTone(evidence.status)}>{evidenceProjectionStatus(t, evidence.status)}</StatusBadge>
      <p className="model-gateway-usage-audit-evidence-source">{t($ => $.auxiliary.adminDeploymentEvidence_ui_Value)}{evidenceProjectionText(t, evidence, "value")}</p>
      <p className="model-gateway-usage-audit-evidence-source">{t($ => $.auxiliary.adminDeploymentEvidence_ui_Source)}{evidence.sourceRef}</p>
      <p>{evidenceProjectionText(t, evidence, "summary")}</p>
    </article>
  );
}

function OperatorRiskCard({ risk }: { risk: AdminProviderDeploymentOperatorRisk }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-overview-trace">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{evidenceProjectionStatus(t, risk.riskSurface)}</p>
          <h5>{evidenceProjectionText(t, risk, "label")}</h5>
        </div>
        <StatusBadge tone="bad">{evidenceProjectionStatus(t, risk.status)}</StatusBadge>
      </div>
      <dl className="model-gateway-overview-meta">
        <div>
          <dt>{t($ => $.auxiliary.adminDeploymentEvidence_ui_Evidence)}</dt>
          <dd>{risk.evidenceRef}</dd>
        </div>
      </dl>
      <p>{evidenceProjectionText(t, risk, "summary")}</p>
      <p>{evidenceProjectionText(t, risk, "operatorQuestion")}</p>
    </article>
  );
}

function BoundaryLockCard({ lock }: { lock: AdminProviderDeploymentLock }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-overview-lock">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{evidenceProjectionStatus(t, lock.lockSurface)}</p>
          <h5>{evidenceProjectionText(t, lock, "label")}</h5>
        </div>
        <StatusBadge tone="bad">{evidenceProjectionStatus(t, lock.status)}</StatusBadge>
      </div>
      <dl className="model-gateway-overview-meta">
        <div>
          <dt>{t($ => $.auxiliary.adminDeploymentEvidence_ui_Missing_prerequisite)}</dt>
          <dd>{evidenceProjectionText(t, lock, "missingPrerequisite")}</dd>
        </div>
      </dl>
      <p>{evidenceProjectionText(t, lock, "summary")}</p>
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
