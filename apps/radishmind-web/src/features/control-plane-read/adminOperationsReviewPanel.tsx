import "../../i18n/adminOperationsEvidenceResources.ts";
import { useTranslation } from "react-i18next";
import { evidenceProjectionText, evidenceProjectionStatus } from "./evidenceProjectionCopy.ts";
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
  const { t } = useTranslation("evaluation");
  return (
    <section
      className="surface-band admin-operations-review model-gateway-overview"
      id="admin-operations-review"
      aria-labelledby="admin-operations-review-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">{t($ => $.auxiliary.adminOperationsEvidence_ui_Admin_Control_Plane)}</p>
          <h3 id="admin-operations-review-title">{t($ => $.auxiliary.adminOperationsEvidence_ui_Operations_review_and_readiness)}</h3>
        </div>
        <StatusBadge tone={review.canRenderAdminOperationsReview ? "neutral" : "bad"}>
          {review.canRenderAdminOperationsReview ? t($ => $.auxiliary.adminOperationsEvidence_ui_offline_review) : t($ => $.auxiliary.adminOperationsEvidence_tenant_overview_value_no)}
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
            <dt>{t($ => $.auxiliary.adminOperationsEvidence_ui_Request)}</dt>
            <dd>{review.requestId}</dd>
          </div>
          <div>
            <dt>{t($ => $.auxiliary.adminOperationsEvidence_ui_Audit)}</dt>
            <dd>{review.auditRef}</dd>
          </div>
          <div>
            <dt>{t($ => $.auxiliary.adminOperationsEvidence_ui_Local_review)}</dt>
            <dd>{review.canInspectAdminEvidenceLocally ? t($ => $.auxiliary.adminOperationsEvidence_ui_enabled) : t($ => $.auxiliary.adminOperationsEvidence_tenant_overview_value_no)}</dd>
          </div>
          <div>
            <dt>{t($ => $.auxiliary.adminOperationsEvidence_ui_Production_backend)}</dt>
            <dd>{review.canRequestProductionBackend ? t($ => $.auxiliary.adminOperationsEvidence_ui_enabled) : t($ => $.auxiliary.adminOperationsEvidence_ui_locked)}</dd>
          </div>
        </dl>
      </article>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.adminOperationsEvidence_ui_Readiness_Rollup)}</p>
          <h4>{t($ => $.auxiliary.adminOperationsEvidence_ui_Admin_evidence_and_production_boundaries)}</h4>
        </div>
        <div className="model-gateway-overview-metric-grid" aria-label={t($ => $.auxiliary.adminOperationsEvidence_ui_Admin_operations_readiness_rollup)}>
          {review.readinessRollup.map((rollup) => (
            <ReadinessRollupCard key={rollup.rollupId} rollup={rollup} />
          ))}
        </div>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.adminOperationsEvidence_ui_Evidence_Checklist)}</p>
          <h4>{t($ => $.auxiliary.adminOperationsEvidence_ui_Tenant_audit_gateway_read_contract_and_production_ops)}</h4>
        </div>
        <div className="model-gateway-overview-policy-grid" aria-label={t($ => $.auxiliary.adminOperationsEvidence_ui_Admin_operations_evidence_checklist)}>
          {review.evidenceChecklist.map((item) => (
            <EvidenceChecklistCard key={item.checklistId} item={item} />
          ))}
        </div>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.adminOperationsEvidence_ui_Operational_Risks)}</p>
          <h4>{t($ => $.auxiliary.adminOperationsEvidence_ui_What_needs_operator_review_before_production_work)}</h4>
        </div>
        <div className="model-gateway-overview-trace-grid" aria-label={t($ => $.auxiliary.adminOperationsEvidence_ui_Admin_operations_risks)}>
          {review.operationalRisks.map((risk) => (
            <OperationalRiskCard key={risk.riskId} risk={risk} />
          ))}
        </div>
      </div>

      <div className="model-gateway-overview-section">
        <div className="model-gateway-overview-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.adminOperationsEvidence_ui_Boundary_Locks)}</p>
          <h4>{t($ => $.auxiliary.adminOperationsEvidence_ui_Capabilities_that_stay_closed)}</h4>
        </div>
        <div className="model-gateway-overview-lock-grid" aria-label={t($ => $.auxiliary.adminOperationsEvidence_ui_Admin_operations_boundary_locks)}>
          {review.boundaryLocks.map((lock) => (
            <BoundaryLockCard key={lock.lockId} lock={lock} />
          ))}
        </div>
      </div>
    </section>
  );
}

function ReadinessRollupCard({ rollup }: { rollup: AdminOperationsReadinessRollup }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-overview-card">
      <span>{evidenceProjectionText(t, rollup, "label")}</span>
      <strong>{evidenceProjectionText(t, rollup, "value")}</strong>
      <StatusBadge tone={adminOperationsTone(rollup.status)}>{evidenceProjectionStatus(t, rollup.status)}</StatusBadge>
      <p className="model-gateway-usage-audit-evidence-source">{t($ => $.auxiliary.adminOperationsEvidence_ui_Source)}{rollup.sourceRef}</p>
      <p>{evidenceProjectionText(t, rollup, "summary")}</p>
    </article>
  );
}

function EvidenceChecklistCard({ item }: { item: AdminOperationsEvidenceChecklistItem }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-overview-policy">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{evidenceProjectionStatus(t, item.sourceSurface)}</p>
          <h5>{evidenceProjectionText(t, item, "label")}</h5>
        </div>
        <StatusBadge tone={adminOperationsTone(item.status)}>{evidenceProjectionStatus(t, item.status)}</StatusBadge>
      </div>
      <dl className="model-gateway-overview-meta">
        <div>
          <dt>{t($ => $.auxiliary.adminOperationsEvidence_ui_Route_page)}</dt>
          <dd>{item.routeOrPageId}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.adminOperationsEvidence_ui_Request)}</dt>
          <dd>{item.requestId}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.adminOperationsEvidence_ui_Audit)}</dt>
          <dd>{item.auditRef}</dd>
        </div>
      </dl>
      <p>{evidenceProjectionText(t, item, "summary")}</p>
    </article>
  );
}

function OperationalRiskCard({ risk }: { risk: AdminOperationsRisk }) {
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
          <dt>{t($ => $.auxiliary.adminOperationsEvidence_ui_Evidence)}</dt>
          <dd>{risk.evidenceRef}</dd>
        </div>
      </dl>
      <p>{evidenceProjectionText(t, risk, "summary")}</p>
      <p>{evidenceProjectionText(t, risk, "operatorQuestion")}</p>
    </article>
  );
}

function BoundaryLockCard({ lock }: { lock: AdminOperationsBoundaryLock }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-overview-lock">
      <div className="model-gateway-overview-row-main">
        <div>
          <p className="eyebrow">{evidenceProjectionStatus(t, lock.sourceSurface)}</p>
          <h5>{evidenceProjectionText(t, lock, "label")}</h5>
        </div>
        <StatusBadge tone="bad">{evidenceProjectionStatus(t, lock.status)}</StatusBadge>
      </div>
      <dl className="model-gateway-overview-meta">
        <div>
          <dt>{t($ => $.auxiliary.adminOperationsEvidence_ui_Missing_prerequisite)}</dt>
          <dd>{evidenceProjectionText(t, lock, "missingPrerequisite")}</dd>
        </div>
      </dl>
      <p>{evidenceProjectionText(t, lock, "summary")}</p>
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
