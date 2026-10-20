import "../../i18n/gatewayReviewEvidenceResources.ts";
import { useTranslation } from "react-i18next";
import { evidenceProjectionText, evidenceProjectionStatus } from "./evidenceProjectionCopy.ts";
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
  const { t } = useTranslation("evaluation");
  return (
    <section
      className="surface-band model-gateway-evidence-review model-gateway-overview"
      id="model-gateway-evidence-review"
      aria-labelledby="model-gateway-evidence-review-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">{t($ => $.auxiliary.gatewayReviewEvidence_ui_Model_Gateway)}</p>
          <h3 id="model-gateway-evidence-review-title">{t($ => $.auxiliary.gatewayReviewEvidence_ui_Evidence_review_and_readiness)}</h3>
        </div>
        <StatusBadge tone={review.canRenderEvidenceReview ? "neutral" : "bad"}>
          {review.canRenderEvidenceReview ? t($ => $.auxiliary.gatewayReviewEvidence_ui_offline_review) : t($ => $.auxiliary.gatewayReviewEvidence_overview_ready_value_no)}
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
            <dt>{t($ => $.auxiliary.gatewayReviewEvidence_ui_Request)}</dt>
            <dd>{review.requestId}</dd>
          </div>
          <div>
            <dt>{t($ => $.auxiliary.gatewayReviewEvidence_ui_Audit)}</dt>
            <dd>{review.auditRef}</dd>
          </div>
          <div>
            <dt>{t($ => $.auxiliary.gatewayReviewEvidence_ui_Local_review)}</dt>
            <dd>{review.canInspectGatewayEvidenceLocally ? t($ => $.auxiliary.gatewayReviewEvidence_ui_enabled) : t($ => $.auxiliary.gatewayReviewEvidence_overview_ready_value_no)}</dd>
          </div>
          <div>
            <dt>{t($ => $.auxiliary.gatewayReviewEvidence_ui_Production_gateway)}</dt>
            <dd>{review.canRequestProductionGateway ? t($ => $.auxiliary.gatewayReviewEvidence_ui_enabled) : t($ => $.auxiliary.gatewayReviewEvidence_ui_locked)}</dd>
          </div>
        </dl>
      </article>

      <div className="model-gateway-overview-section model-gateway-runtime-review-link">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.gatewayReviewEvidence_ui_Current_request_evidence)}</p>
          <h4>{t($ => $.auxiliary.gatewayReviewEvidence_ui_Open_the_application_scoped_runtime_review)}</h4>
        </div>
        <p>{t($ => $.auxiliary.gatewayReviewEvidence_ui_Real_request_history_now_stays_with_its_selected)}</p>
        <a href="#model-gateway-request-history">{t($ => $.auxiliary.gatewayReviewEvidence_ui_Open_sanitized_request_history)}<span aria-hidden="true">→</span></a>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.gatewayReviewEvidence_ui_Readiness_Rollup)}</p>
          <h4>{t($ => $.auxiliary.gatewayReviewEvidence_ui_What_is_readable_and_what_still_needs_review)}</h4>
        </div>
        <div className="model-gateway-overview-metric-grid" aria-label={t($ => $.auxiliary.gatewayReviewEvidence_ui_Model_gateway_readiness_rollup)}>
          {review.readinessRollup.map((rollup) => (
            <ReadinessRollupCard key={rollup.rollupId} rollup={rollup} />
          ))}
        </div>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.gatewayReviewEvidence_ui_Evidence_Checklist)}</p>
          <h4>{t($ => $.auxiliary.gatewayReviewEvidence_ui_Overview_route_usage_audit_and_lock_coverage)}</h4>
        </div>
        <div className="model-gateway-overview-policy-grid" aria-label={t($ => $.auxiliary.gatewayReviewEvidence_ui_Model_gateway_evidence_checklist)}>
          {review.evidenceChecklist.map((item) => (
            <EvidenceChecklistCard key={item.checklistId} item={item} />
          ))}
        </div>
      </div>

      <RiskSection title={t($ => $.auxiliary.gatewayReviewEvidence_ui_Route_risks)} label={t($ => $.auxiliary.gatewayReviewEvidence_ui_Route)} risks={review.routeRisks} />
      <RiskSection title={t($ => $.auxiliary.gatewayReviewEvidence_ui_Usage_risks)} label={t($ => $.auxiliary.gatewayReviewEvidence_ui_Usage)} risks={review.usageRisks} />
      <RiskSection title={t($ => $.auxiliary.gatewayReviewEvidence_ui_Audit_risks)} label={t($ => $.auxiliary.gatewayReviewEvidence_ui_Audit)} risks={review.auditRisks} />

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.gatewayReviewEvidence_ui_Locked_Capabilities)}</p>
          <h4>{t($ => $.auxiliary.gatewayReviewEvidence_ui_Capabilities_that_remain_closed)}</h4>
        </div>
        <div className="model-gateway-overview-lock-grid" aria-label={t($ => $.auxiliary.gatewayReviewEvidence_ui_Model_gateway_review_locked_capabilities)}>
          {review.lockedCapabilities.map((capability) => (
            <LockedCapabilityCard key={capability.capabilityId} capability={capability} />
          ))}
        </div>
      </div>
    </section>
  );
}

function ReadinessRollupCard({ rollup }: { rollup: ModelGatewayEvidenceReadinessRollup }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-overview-card">
      <span>{evidenceProjectionText(t, rollup, "label")}</span>
      <strong>{evidenceProjectionText(t, rollup, "value")}</strong>
      <StatusBadge tone={evidenceReviewTone(rollup.status)}>{evidenceProjectionStatus(t, rollup.status)}</StatusBadge>
      <p className="model-gateway-usage-audit-evidence-source">{t($ => $.auxiliary.gatewayReviewEvidence_ui_Source)}{rollup.sourceRef}</p>
      <p>{evidenceProjectionText(t, rollup, "summary")}</p>
    </article>
  );
}

function EvidenceChecklistCard({ item }: { item: ModelGatewayEvidenceChecklistItem }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-overview-policy">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{evidenceProjectionStatus(t, item.sourceSurface)}</p>
          <h5>{evidenceProjectionText(t, item, "label")}</h5>
        </div>
        <StatusBadge tone={evidenceReviewTone(item.status)}>{evidenceProjectionStatus(t, item.status)}</StatusBadge>
      </div>
      <dl className="model-gateway-overview-meta">
        <div>
          <dt>{t($ => $.auxiliary.gatewayReviewEvidence_ui_Route_page)}</dt>
          <dd>{item.routeOrPageId}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayReviewEvidence_ui_Request)}</dt>
          <dd>{item.requestId}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayReviewEvidence_ui_Audit)}</dt>
          <dd>{item.auditRef}</dd>
        </div>
      </dl>
      <p>{evidenceProjectionText(t, item, "summary")}</p>
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
  const { t } = useTranslation("evaluation");
  return (
    <div className="model-gateway-overview-section">
      <div className="model-gateway-overview-subheading">
        <p className="eyebrow">{label}</p>
        <h4>{title}</h4>
      </div>
      <div className="model-gateway-overview-trace-grid" aria-label={t($ => $.auxiliary.gatewayReviewEvidence_ui_Model_gateway_value_risks, { value: label.toLowerCase() })}>
        {risks.map((risk) => (
          <RiskCard key={risk.riskId} risk={risk} />
        ))}
      </div>
    </div>
  );
}

function RiskCard({ risk }: { risk: ModelGatewayEvidenceRisk }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-overview-trace">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{evidenceProjectionStatus(t, risk.sourceSurface)}</p>
          <h5>{evidenceProjectionText(t, risk, "label")}</h5>
        </div>
        <StatusBadge tone="bad">{evidenceProjectionStatus(t, risk.status)}</StatusBadge>
      </div>
      <dl className="model-gateway-overview-meta">
        <div>
          <dt>{t($ => $.auxiliary.gatewayReviewEvidence_ui_Evidence)}</dt>
          <dd>{risk.evidenceRef}</dd>
        </div>
      </dl>
      <p>{evidenceProjectionText(t, risk, "summary")}</p>
      <p>{evidenceProjectionText(t, risk, "reviewQuestion")}</p>
    </article>
  );
}

function LockedCapabilityCard({
  capability,
}: {
  capability: ModelGatewayEvidenceLockedCapability;
}) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-overview-lock">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{evidenceProjectionStatus(t, capability.sourceSurface)}</p>
          <h5>{evidenceProjectionText(t, capability, "label")}</h5>
        </div>
        <StatusBadge tone="bad">{evidenceProjectionStatus(t, capability.status)}</StatusBadge>
      </div>
      <dl className="model-gateway-overview-meta">
        <div>
          <dt>{t($ => $.auxiliary.gatewayReviewEvidence_ui_Source_2)}</dt>
          <dd>{capability.sourceRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayReviewEvidence_ui_Missing_prerequisite)}</dt>
          <dd>{evidenceProjectionText(t, capability, "missingPrerequisite")}</dd>
        </div>
      </dl>
      <p>{evidenceProjectionText(t, capability, "summary")}</p>
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
