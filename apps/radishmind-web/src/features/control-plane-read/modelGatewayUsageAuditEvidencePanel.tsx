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
  return (
    <section
      className="surface-band model-gateway-usage-audit-evidence"
      id="model-gateway-usage-audit-evidence"
      aria-labelledby="model-gateway-usage-audit-evidence-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">Model Gateway</p>
          <h3 id="model-gateway-usage-audit-evidence-title">Usage, trace, and audit evidence</h3>
        </div>
        <StatusBadge tone={evidence.canRenderUsageAuditEvidence ? "neutral" : "bad"}>
          {evidence.canRenderUsageAuditEvidence ? "offline detail" : "blocked"}
        </StatusBadge>
      </div>

      <article className="model-gateway-usage-audit-evidence-hero">
        <div>
          <p className="eyebrow">{evidence.evidenceMode}</p>
          <h4>{evidence.tenantRef}</h4>
          <p>{evidence.evidenceNarrative}</p>
        </div>
        <dl className="model-gateway-usage-audit-evidence-meta">
          <div>
            <dt>Overview</dt>
            <dd>{evidence.overviewPageId}</dd>
          </div>
          <div>
            <dt>Route evidence</dt>
            <dd>{evidence.routeEvidencePageId}</dd>
          </div>
          <div>
            <dt>Request</dt>
            <dd>{evidence.requestId}</dd>
          </div>
          <div>
            <dt>Audit</dt>
            <dd>{evidence.auditRef}</dd>
          </div>
          <div>
            <dt>Quota execution</dt>
            <dd>{evidence.canEnforceQuota ? "enabled" : "locked"}</dd>
          </div>
          <div>
            <dt>Replay</dt>
            <dd>{evidence.canReplayRun ? "enabled" : "locked"}</dd>
          </div>
        </dl>
      </article>

      <div className="model-gateway-usage-audit-evidence-rollup-grid" aria-label="Model gateway usage rollups">
        {evidence.rollups.map((rollup) => (
          <UsageRollupCard key={rollup.rollupId} rollup={rollup} />
        ))}
      </div>

      <div className="model-gateway-usage-audit-evidence-section">
        <div className="model-gateway-usage-audit-evidence-subheading">
          <p className="eyebrow">Key Scope</p>
          <h4>Read access references</h4>
        </div>
        <div className="model-gateway-usage-audit-evidence-key-grid" aria-label="Model gateway key scope evidence">
          {evidence.keyScopes.map((keyScope) => (
            <KeyScopeCard key={keyScope.keyId} keyScope={keyScope} />
          ))}
        </div>
      </div>

      <div className="model-gateway-usage-audit-evidence-section">
        <div className="model-gateway-usage-audit-evidence-subheading">
          <p className="eyebrow">Quota and Cost</p>
          <h4>Policy snapshots</h4>
        </div>
        <div className="model-gateway-usage-audit-evidence-quota-grid" aria-label="Model gateway quota evidence">
          {evidence.quotaCosts.map((quotaCost) => (
            <QuotaCostCard key={quotaCost.evidenceId} quotaCost={quotaCost} />
          ))}
        </div>
      </div>

      <div className="model-gateway-usage-audit-evidence-section">
        <div className="model-gateway-usage-audit-evidence-subheading">
          <p className="eyebrow">Trace Evidence</p>
          <h4>Run records and failure codes</h4>
        </div>
        <div className="model-gateway-usage-audit-evidence-trace-grid" aria-label="Model gateway trace evidence">
          {evidence.traceAudit.map((trace) => (
            <TraceAuditCard key={trace.traceId} trace={trace} />
          ))}
        </div>
      </div>

      <div className="model-gateway-usage-audit-evidence-section">
        <div className="model-gateway-usage-audit-evidence-subheading">
          <p className="eyebrow">Audit Correlation</p>
          <h4>Decisions and resource refs</h4>
        </div>
        <div className="model-gateway-usage-audit-evidence-audit-grid" aria-label="Model gateway audit evidence">
          {evidence.auditCorrelations.map((audit) => (
            <AuditCorrelationCard key={audit.auditRef} audit={audit} />
          ))}
        </div>
      </div>

      <div className="model-gateway-usage-audit-evidence-section">
        <div className="model-gateway-usage-audit-evidence-subheading">
          <p className="eyebrow">Stop Lines</p>
          <h4>Locked usage capabilities</h4>
        </div>
        <div className="model-gateway-usage-audit-evidence-guard-grid" aria-label="Model gateway usage guards">
          {evidence.guards.map((guard) => (
            <UsageAuditGuardCard key={guard.guardId} guard={guard} />
          ))}
        </div>
      </div>
    </section>
  );
}

function UsageRollupCard({ rollup }: { rollup: ModelGatewayUsageAuditRollup }) {
  return (
    <article className="model-gateway-usage-audit-evidence-card">
      <span>{rollup.label}</span>
      <strong>{rollup.value}</strong>
      <StatusBadge tone={usageAuditTone(rollup.status)}>{rollup.status}</StatusBadge>
      <p className="model-gateway-usage-audit-evidence-source">Source: {rollup.sourceRef}</p>
      <p>{rollup.summary}</p>
    </article>
  );
}

function KeyScopeCard({ keyScope }: { keyScope: ModelGatewayKeyScopeEvidence }) {
  return (
    <article className="model-gateway-usage-audit-evidence-card">
      <div className="model-gateway-usage-audit-evidence-row-main">
        <div>
          <p className="eyebrow">{keyScope.ownerSubjectRef}</p>
          <h5>{keyScope.keyId}</h5>
        </div>
        <StatusBadge tone={usageAuditTone(keyScope.status)}>{keyScope.status}</StatusBadge>
      </div>
      <dl className="model-gateway-usage-audit-evidence-meta">
        <div>
          <dt>State</dt>
          <dd>{keyScope.state}</dd>
        </div>
        <div>
          <dt>Last used</dt>
          <dd>{keyScope.lastUsedAt}</dd>
        </div>
        <div>
          <dt>Source</dt>
          <dd>{keyScope.sourceRef}</dd>
        </div>
      </dl>
      <ChipList label="Scopes" items={keyScope.scopes} />
      <p>{keyScope.summary}</p>
    </article>
  );
}

function QuotaCostCard({ quotaCost }: { quotaCost: ModelGatewayQuotaCostEvidence }) {
  return (
    <article className="model-gateway-usage-audit-evidence-card">
      <div className="model-gateway-usage-audit-evidence-row-main">
        <div>
          <p className="eyebrow">{quotaCost.evidenceId}</p>
          <h5>{quotaCost.label}</h5>
        </div>
        <StatusBadge tone={usageAuditTone(quotaCost.status)}>{quotaCost.status}</StatusBadge>
      </div>
      <dl className="model-gateway-usage-audit-evidence-meta">
        <div>
          <dt>Limit</dt>
          <dd>{quotaCost.limit}</dd>
        </div>
        <div>
          <dt>Used</dt>
          <dd>{quotaCost.used}</dd>
        </div>
        <div>
          <dt>Usage</dt>
          <dd>{quotaCost.usage}</dd>
        </div>
        <div>
          <dt>Source</dt>
          <dd>{quotaCost.sourceRef}</dd>
        </div>
      </dl>
      <p>{quotaCost.summary}</p>
    </article>
  );
}

function TraceAuditCard({ trace }: { trace: ModelGatewayTraceAuditEvidence }) {
  return (
    <article className="model-gateway-usage-audit-evidence-card">
      <div className="model-gateway-usage-audit-evidence-row-main">
        <div>
          <p className="eyebrow">{trace.runId}</p>
          <h5>{trace.traceId}</h5>
        </div>
        <StatusBadge tone={usageAuditTone(trace.status)}>{trace.status}</StatusBadge>
      </div>
      <dl className="model-gateway-usage-audit-evidence-meta">
        <div>
          <dt>Application</dt>
          <dd>{trace.applicationRef}</dd>
        </div>
        <div>
          <dt>Workflow</dt>
          <dd>{trace.workflowDefinitionId}</dd>
        </div>
        <div>
          <dt>Run status</dt>
          <dd>{trace.runStatus}</dd>
        </div>
        <div>
          <dt>Failure</dt>
          <dd>{trace.failureCode}</dd>
        </div>
        <div>
          <dt>Cost</dt>
          <dd>{trace.estimatedCost}</dd>
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

function AuditCorrelationCard({ audit }: { audit: ModelGatewayAuditCorrelationEvidence }) {
  return (
    <article className="model-gateway-usage-audit-evidence-card">
      <div className="model-gateway-usage-audit-evidence-row-main">
        <div>
          <p className="eyebrow">{audit.eventKind}</p>
          <h5>{audit.auditRef}</h5>
        </div>
        <StatusBadge tone={usageAuditTone(audit.status)}>{audit.status}</StatusBadge>
      </div>
      <dl className="model-gateway-usage-audit-evidence-meta">
        <div>
          <dt>Actor</dt>
          <dd>{audit.actorSubjectRef}</dd>
        </div>
        <div>
          <dt>Resource</dt>
          <dd>{audit.resourceRef}</dd>
        </div>
        <div>
          <dt>Decision</dt>
          <dd>{audit.decision}</dd>
        </div>
        <div>
          <dt>Failure</dt>
          <dd>{audit.failureCode}</dd>
        </div>
        <div>
          <dt>Trace</dt>
          <dd>{audit.traceId}</dd>
        </div>
      </dl>
      <p>{audit.summary}</p>
    </article>
  );
}

function UsageAuditGuardCard({ guard }: { guard: ModelGatewayUsageAuditGuard }) {
  return (
    <article className="model-gateway-usage-audit-evidence-guard">
      <div className="model-gateway-usage-audit-evidence-row-main">
        <div>
          <p className="eyebrow">{guard.guardId}</p>
          <h5>{guard.label}</h5>
        </div>
        <StatusBadge tone={usageAuditTone(guard.status)}>{guard.status}</StatusBadge>
      </div>
      <dl className="model-gateway-usage-audit-evidence-meta">
        <div>
          <dt>Source</dt>
          <dd>{guard.sourceRef}</dd>
        </div>
        <div>
          <dt>Missing prerequisite</dt>
          <dd>{guard.missingPrerequisite}</dd>
        </div>
      </dl>
      <p>{guard.summary}</p>
    </article>
  );
}

function ChipList({ items, label }: { items: string[]; label: string }) {
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
