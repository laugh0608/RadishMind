import { formatDisplayDate } from "../../i18n/formatters.ts";
import { resolveUiLocale } from "../../i18n/localePreference.ts";
import "../../i18n/gatewayUsageEvidenceResources.ts";
import { useTranslation } from "react-i18next";
import { evidenceProjectionText, evidenceProjectionStatus } from "./evidenceProjectionCopy.ts";
import type {
  ModelGatewayAuditCorrelationEvidence,
  ModelGatewayKeyScopeEvidence,
  ModelGatewayQuotaCostEvidence,
  ModelGatewayTraceAuditEvidence,
  ModelGatewayUsageAuditEvidenceStatus,
  ModelGatewayUsageAuditEvidenceViewModel,
  ModelGatewayUsageAuditGuard,
  ModelGatewayUsageAuditRollup,
} from "./modelGatewayUsageAuditEvidence";

type StatusBadgeTone = "good" | "bad" | "neutral";

export function ModelGatewayUsageAuditEvidencePanel({
  evidence,
}: {
  evidence: ModelGatewayUsageAuditEvidenceViewModel;
}) {
  const { t } = useTranslation("evaluation");
  return (
    <section
      className="surface-band model-gateway-usage-audit-evidence"
      id="model-gateway-usage-audit-evidence"
      aria-labelledby="model-gateway-usage-audit-evidence-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">{t($ => $.auxiliary.gatewayUsageEvidence_ui_Model_Gateway)}</p>
          <h3 id="model-gateway-usage-audit-evidence-title">{t($ => $.auxiliary.gatewayUsageEvidence_ui_Usage_trace_and_audit_evidence)}</h3>
        </div>
        <StatusBadge tone={evidence.canRenderUsageAuditEvidence ? "neutral" : "bad"}>
          {evidence.canRenderUsageAuditEvidence ? t($ => $.auxiliary.gatewayUsageEvidence_ui_offline_detail) : t($ => $.auxiliary.gatewayUsageEvidence_quota_missing_usage)}
        </StatusBadge>
      </div>

      <article className="model-gateway-usage-audit-evidence-hero">
        <div>
          <p className="eyebrow">{evidenceProjectionStatus(t, evidence.evidenceMode)}</p>
          <h4>{evidence.tenantRef}</h4>
          <p>{evidenceProjectionText(t, evidence, "evidenceNarrative")}</p>
        </div>
        <dl className="model-gateway-usage-audit-evidence-meta">
          <div>
            <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Overview)}</dt>
            <dd>{evidence.overviewPageId}</dd>
          </div>
          <div>
            <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Route_evidence)}</dt>
            <dd>{evidence.routeEvidencePageId}</dd>
          </div>
          <div>
            <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Request)}</dt>
            <dd>{evidence.requestId}</dd>
          </div>
          <div>
            <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Audit)}</dt>
            <dd>{evidence.auditRef}</dd>
          </div>
          <div>
            <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Quota_execution)}</dt>
            <dd>{evidence.canEnforceQuota ? t($ => $.auxiliary.gatewayUsageEvidence_ui_enabled) : t($ => $.auxiliary.gatewayUsageEvidence_ui_locked)}</dd>
          </div>
          <div>
            <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Replay)}</dt>
            <dd>{evidence.canReplayRun ? t($ => $.auxiliary.gatewayUsageEvidence_ui_enabled) : t($ => $.auxiliary.gatewayUsageEvidence_ui_locked)}</dd>
          </div>
        </dl>
      </article>

      <div className="model-gateway-usage-audit-evidence-rollup-grid" aria-label={t($ => $.auxiliary.gatewayUsageEvidence_ui_Model_gateway_usage_rollups)}>
        {evidence.rollups.map((rollup) => (
          <UsageRollupCard key={rollup.rollupId} rollup={rollup} />
        ))}
      </div>

      <div className="model-gateway-usage-audit-evidence-section">
        <div className="model-gateway-usage-audit-evidence-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.gatewayUsageEvidence_ui_Key_Scope)}</p>
          <h4>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Read_access_references)}</h4>
        </div>
        <div className="model-gateway-usage-audit-evidence-key-grid" aria-label={t($ => $.auxiliary.gatewayUsageEvidence_ui_Model_gateway_key_scope_evidence)}>
          {evidence.keyScopes.map((keyScope) => (
            <KeyScopeCard key={keyScope.keyId} keyScope={keyScope} />
          ))}
        </div>
      </div>

      <div className="model-gateway-usage-audit-evidence-section">
        <div className="model-gateway-usage-audit-evidence-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.gatewayUsageEvidence_ui_Quota_and_Cost)}</p>
          <h4>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Policy_snapshots)}</h4>
        </div>
        <div className="model-gateway-usage-audit-evidence-quota-grid" aria-label={t($ => $.auxiliary.gatewayUsageEvidence_ui_Model_gateway_quota_evidence)}>
          {evidence.quotaCosts.map((quotaCost) => (
            <QuotaCostCard key={quotaCost.evidenceId} quotaCost={quotaCost} />
          ))}
        </div>
      </div>

      <div className="model-gateway-usage-audit-evidence-section">
        <div className="model-gateway-usage-audit-evidence-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.gatewayUsageEvidence_ui_Trace_Evidence)}</p>
          <h4>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Run_records_and_failure_codes)}</h4>
        </div>
        <div className="model-gateway-usage-audit-evidence-trace-grid" aria-label={t($ => $.auxiliary.gatewayUsageEvidence_ui_Model_gateway_trace_evidence)}>
          {evidence.traceAudit.map((trace) => (
            <TraceAuditCard key={`${trace.traceId}:${trace.runId}`} trace={trace} />
          ))}
        </div>
      </div>

      <div className="model-gateway-usage-audit-evidence-section">
        <div className="model-gateway-usage-audit-evidence-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.gatewayUsageEvidence_ui_Audit_Correlation)}</p>
          <h4>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Decisions_and_resource_refs)}</h4>
        </div>
        <div className="model-gateway-usage-audit-evidence-audit-grid" aria-label={t($ => $.auxiliary.gatewayUsageEvidence_ui_Model_gateway_audit_evidence)}>
          {evidence.auditCorrelations.map((audit) => (
            <AuditCorrelationCard key={audit.auditRef} audit={audit} />
          ))}
        </div>
      </div>

      <div className="model-gateway-usage-audit-evidence-section">
        <div className="model-gateway-usage-audit-evidence-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.gatewayUsageEvidence_ui_Stop_Lines)}</p>
          <h4>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Locked_usage_capabilities)}</h4>
        </div>
        <div className="model-gateway-usage-audit-evidence-guard-grid" aria-label={t($ => $.auxiliary.gatewayUsageEvidence_ui_Model_gateway_usage_guards)}>
          {evidence.guards.map((guard) => (
            <UsageAuditGuardCard key={guard.guardId} guard={guard} />
          ))}
        </div>
      </div>
    </section>
  );
}

function UsageRollupCard({ rollup }: { rollup: ModelGatewayUsageAuditRollup }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-usage-audit-evidence-card">
      <span>{evidenceProjectionText(t, rollup, "label")}</span>
      <strong>{evidenceProjectionText(t, rollup, "value")}</strong>
      <StatusBadge tone={usageAuditTone(rollup.status)}>{evidenceProjectionStatus(t, rollup.status)}</StatusBadge>
      <p className="model-gateway-usage-audit-evidence-source">{t($ => $.auxiliary.gatewayUsageEvidence_ui_Source)}{rollup.sourceRef}</p>
      <p>{evidenceProjectionText(t, rollup, "summary")}</p>
    </article>
  );
}

function KeyScopeCard({ keyScope }: { keyScope: ModelGatewayKeyScopeEvidence }) {
  const { t, i18n } = useTranslation("evaluation");
  return (
    <article className="model-gateway-usage-audit-evidence-card">
      <div className="model-gateway-usage-audit-evidence-row-main">
        <div>
          <p className="eyebrow">{keyScope.ownerSubjectRef}</p>
          <h5>{keyScope.keyId}</h5>
        </div>
        <StatusBadge tone={usageAuditTone(keyScope.status)}>{evidenceProjectionStatus(t, keyScope.status)}</StatusBadge>
      </div>
      <dl className="model-gateway-usage-audit-evidence-meta">
        <div>
          <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_State)}</dt>
          <dd>{evidenceProjectionStatus(t, keyScope.state)}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Last_used)}</dt>
          <dd title={keyScope.lastUsedAt}>{formatDisplayDate(keyScope.lastUsedAt, resolveUiLocale(i18n.resolvedLanguage, []), { timeZone: "UTC" }) ?? evidenceProjectionStatus(t, keyScope.lastUsedAt)}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Source_2)}</dt>
          <dd>{keyScope.sourceRef}</dd>
        </div>
      </dl>
      <ChipList label={t($ => $.auxiliary.gatewayUsageEvidence_ui_Scopes)} items={keyScope.scopes} />
      <p>{evidenceProjectionText(t, keyScope, "summary")}</p>
    </article>
  );
}

function QuotaCostCard({ quotaCost }: { quotaCost: ModelGatewayQuotaCostEvidence }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-usage-audit-evidence-card">
      <div className="model-gateway-usage-audit-evidence-row-main">
        <div>
          <p className="eyebrow">{quotaCost.evidenceId}</p>
          <h5>{evidenceProjectionText(t, quotaCost, "label")}</h5>
        </div>
        <StatusBadge tone={usageAuditTone(quotaCost.status)}>{evidenceProjectionStatus(t, quotaCost.status)}</StatusBadge>
      </div>
      <dl className="model-gateway-usage-audit-evidence-meta">
        <div>
          <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Limit)}</dt>
          <dd>{evidenceProjectionText(t, quotaCost, "limit")}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Used)}</dt>
          <dd>{evidenceProjectionText(t, quotaCost, "used")}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Usage)}</dt>
          <dd>{evidenceProjectionText(t, quotaCost, "usage")}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Source_2)}</dt>
          <dd>{quotaCost.sourceRef}</dd>
        </div>
      </dl>
      <p>{evidenceProjectionText(t, quotaCost, "summary")}</p>
    </article>
  );
}

function TraceAuditCard({ trace }: { trace: ModelGatewayTraceAuditEvidence }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-usage-audit-evidence-card">
      <div className="model-gateway-usage-audit-evidence-row-main">
        <div>
          <p className="eyebrow">{trace.runId}</p>
          <h5>{trace.traceId}</h5>
        </div>
        <StatusBadge tone={usageAuditTone(trace.status)}>{evidenceProjectionStatus(t, trace.status)}</StatusBadge>
      </div>
      <dl className="model-gateway-usage-audit-evidence-meta">
        <div>
          <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Application)}</dt>
          <dd>{trace.applicationRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Workflow)}</dt>
          <dd>{trace.workflowDefinitionId}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Run_status)}</dt>
          <dd>{evidenceProjectionStatus(t, trace.runStatus)}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Failure)}</dt>
          <dd>{trace.failureCode}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayUsageEvidence_cost_limit_label)}</dt>
          <dd>{trace.estimatedCost}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Audit)}</dt>
          <dd>{trace.auditRef}</dd>
        </div>
      </dl>
      <p>{evidenceProjectionText(t, trace, "summary")}</p>
    </article>
  );
}

function AuditCorrelationCard({ audit }: { audit: ModelGatewayAuditCorrelationEvidence }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-usage-audit-evidence-card">
      <div className="model-gateway-usage-audit-evidence-row-main">
        <div>
          <p className="eyebrow">{audit.eventKind}</p>
          <h5>{audit.auditRef}</h5>
        </div>
        <StatusBadge tone={usageAuditTone(audit.status)}>{evidenceProjectionStatus(t, audit.status)}</StatusBadge>
      </div>
      <dl className="model-gateway-usage-audit-evidence-meta">
        <div>
          <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Actor)}</dt>
          <dd>{audit.actorSubjectRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Resource)}</dt>
          <dd>{audit.resourceRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Decision)}</dt>
          <dd>{evidenceProjectionStatus(t, audit.decision)}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Failure)}</dt>
          <dd>{audit.failureCode}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Trace)}</dt>
          <dd>{audit.traceId}</dd>
        </div>
      </dl>
      <p>{evidenceProjectionText(t, audit, "summary")}</p>
    </article>
  );
}

function UsageAuditGuardCard({ guard }: { guard: ModelGatewayUsageAuditGuard }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-usage-audit-evidence-guard">
      <div className="model-gateway-usage-audit-evidence-row-main">
        <div>
          <p className="eyebrow">{guard.guardId}</p>
          <h5>{evidenceProjectionText(t, guard, "label")}</h5>
        </div>
        <StatusBadge tone={usageAuditTone(guard.status)}>{evidenceProjectionStatus(t, guard.status)}</StatusBadge>
      </div>
      <dl className="model-gateway-usage-audit-evidence-meta">
        <div>
          <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Source_2)}</dt>
          <dd>{guard.sourceRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayUsageEvidence_ui_Missing_prerequisite)}</dt>
          <dd>{evidenceProjectionText(t, guard, "missingPrerequisite")}</dd>
        </div>
      </dl>
      <p>{evidenceProjectionText(t, guard, "summary")}</p>
    </article>
  );
}

function ChipList({ items, label }: { items: string[]; label: string }) {
  const { t } = useTranslation("evaluation");
  return (
    <div className="model-gateway-usage-audit-evidence-chip-row" aria-label={label}>
      <span>{label}</span>
      <div>
        {items.map((item) => (
          <code key={item}>{item}</code>
        ))}
      </div>
    </div>
  );
}

function usageAuditTone(status: ModelGatewayUsageAuditEvidenceStatus): StatusBadgeTone {
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
