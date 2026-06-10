import type { AdminAuditLogViewModel } from "./adminAuditLog";
import type { ModelGatewayOverviewStatus, ModelGatewayOverviewViewModel } from "./modelGatewayOverview";
import type { ModelGatewayRouteEvidenceViewModel } from "./modelGatewayRouteEvidence";
import type { WorkspaceApiKeysViewModel } from "./workspaceApiKeys";
import type { WorkspaceRunHistoryViewModel } from "./workspaceRunHistory";
import type { WorkspaceUsageQuotaViewModel } from "./workspaceUsageQuota";

export type ModelGatewayUsageAuditEvidenceStatus = ModelGatewayOverviewStatus;

export type ModelGatewayUsageAuditEvidenceSource = {
  overview: ModelGatewayOverviewViewModel;
  routeEvidence: ModelGatewayRouteEvidenceViewModel;
  workspaceApiKeys: WorkspaceApiKeysViewModel;
  workspaceUsageQuota: WorkspaceUsageQuotaViewModel;
  workspaceRunHistory: WorkspaceRunHistoryViewModel;
  adminAuditLog: AdminAuditLogViewModel;
};

export type ModelGatewayUsageAuditRollup = {
  rollupId: string;
  label: string;
  value: string;
  status: ModelGatewayUsageAuditEvidenceStatus;
  sourceRef: string;
  summary: string;
};

export type ModelGatewayKeyScopeEvidence = {
  keyId: string;
  ownerSubjectRef: string;
  scopes: string[];
  state: string;
  lastUsedAt: string;
  status: ModelGatewayUsageAuditEvidenceStatus;
  sourceRef: string;
  summary: string;
};

export type ModelGatewayQuotaCostEvidence = {
  evidenceId: string;
  label: string;
  limit: string;
  used: string;
  usage: string;
  status: ModelGatewayUsageAuditEvidenceStatus;
  sourceRef: string;
  summary: string;
};

export type ModelGatewayTraceAuditEvidence = {
  traceId: string;
  runId: string;
  applicationRef: string;
  workflowDefinitionId: string;
  runStatus: string;
  failureCode: string;
  estimatedCost: string;
  auditRef: string;
  status: ModelGatewayUsageAuditEvidenceStatus;
  summary: string;
};

export type ModelGatewayAuditCorrelationEvidence = {
  auditRef: string;
  actorSubjectRef: string;
  eventKind: string;
  resourceRef: string;
  decision: string;
  failureCode: string;
  traceId: string;
  status: ModelGatewayUsageAuditEvidenceStatus;
  summary: string;
};

export type ModelGatewayUsageAuditGuard = {
  guardId: string;
  label: string;
  status: ModelGatewayUsageAuditEvidenceStatus;
  sourceRef: string;
  missingPrerequisite: string;
  summary: string;
};

export type ModelGatewayUsageAuditEvidenceViewModel = {
  pageId: "model-gateway-usage-audit-evidence-offline";
  sourcePageIds: string[];
  evidenceMode: "offline_usage_trace_audit_evidence";
  overviewPageId: ModelGatewayOverviewViewModel["pageId"];
  routeEvidencePageId: ModelGatewayRouteEvidenceViewModel["pageId"];
  tenantRef: string;
  requestId: string;
  auditRef: string;
  evidenceNarrative: string;
  rollups: ModelGatewayUsageAuditRollup[];
  keyScopes: ModelGatewayKeyScopeEvidence[];
  quotaCosts: ModelGatewayQuotaCostEvidence[];
  traceAudit: ModelGatewayTraceAuditEvidence[];
  auditCorrelations: ModelGatewayAuditCorrelationEvidence[];
  guards: ModelGatewayUsageAuditGuard[];
  canRenderUsageAuditEvidence: boolean;
  canRevealApiKey: false;
  canValidateApiKey: false;
  canEnforceQuota: false;
  canApplyRateLimit: false;
  canWriteCostRecord: false;
  canExportRawAuditPayload: false;
  canReplayRun: false;
  canUseProductionGateway: false;
};

export function buildModelGatewayUsageAuditEvidenceViewModel(
  source: ModelGatewayUsageAuditEvidenceSource,
): ModelGatewayUsageAuditEvidenceViewModel {
  const rollups = buildRollups(source);
  const keyScopes = buildKeyScopes(source);
  const quotaCosts = buildQuotaCosts(source);
  const traceAudit = buildTraceAudit(source);
  const auditCorrelations = buildAuditCorrelations(source);
  const guards = buildGuards(source);

  return {
    pageId: "model-gateway-usage-audit-evidence-offline",
    sourcePageIds: [
      source.overview.pageId,
      source.routeEvidence.pageId,
      source.workspaceApiKeys.pageId,
      source.workspaceUsageQuota.pageId,
      source.workspaceRunHistory.pageId,
      source.adminAuditLog.pageId,
      "gateway-api-key-quota-readiness",
      "provider-retry-fallback-policy-v1",
      "control-plane-read-formal-ui-readiness-close-v1",
    ],
    evidenceMode: "offline_usage_trace_audit_evidence",
    overviewPageId: source.overview.pageId,
    routeEvidencePageId: source.routeEvidence.pageId,
    tenantRef: source.workspaceUsageQuota.collection.tenantRef,
    requestId: source.workspaceRunHistory.requestId,
    auditRef: source.adminAuditLog.auditRef,
    evidenceNarrative: buildNarrative(source, rollups, traceAudit, auditCorrelations),
    rollups,
    keyScopes,
    quotaCosts,
    traceAudit,
    auditCorrelations,
    guards,
    canRenderUsageAuditEvidence:
      source.overview.canRenderModelGatewayOverview &&
      source.routeEvidence.canRenderRouteEvidenceDetail &&
      source.workspaceApiKeys.canRenderApiKeys &&
      source.workspaceUsageQuota.canRenderQuota &&
      source.workspaceRunHistory.canRenderRuns &&
      source.adminAuditLog.canRenderAuditLog &&
      rollups.length >= 6 &&
      keyScopes.length >= 2 &&
      quotaCosts.length >= 3 &&
      traceAudit.length >= 2 &&
      auditCorrelations.length >= 2 &&
      guards.length >= 6,
    canRevealApiKey: false,
    canValidateApiKey: false,
    canEnforceQuota: false,
    canApplyRateLimit: false,
    canWriteCostRecord: false,
    canExportRawAuditPayload: false,
    canReplayRun: false,
    canUseProductionGateway: false,
  };
}

function buildNarrative(
  source: ModelGatewayUsageAuditEvidenceSource,
  rollups: ModelGatewayUsageAuditRollup[],
  traceAudit: ModelGatewayTraceAuditEvidence[],
  auditCorrelations: ModelGatewayAuditCorrelationEvidence[],
): string {
  const activeKeys = source.workspaceApiKeys.apiKeys.filter((apiKey) => apiKey.state === "active").length;
  const failedRuns = traceAudit.filter((run) => run.failureCode !== "none").length;
  const deniedAuditEvents = auditCorrelations.filter((event) => event.decision === "denied").length;
  return `${rollups.length} usage rollups, ${source.workspaceApiKeys.apiKeys.length} key summaries, ${activeKeys} active key reference, ${traceAudit.length} trace records, ${failedRuns} failed run, and ${deniedAuditEvents} denied audit event are grouped for offline gateway evidence.`;
}

function buildRollups(source: ModelGatewayUsageAuditEvidenceSource): ModelGatewayUsageAuditRollup[] {
  const quota = source.workspaceUsageQuota.quota;
  const activeKeys = source.workspaceApiKeys.apiKeys.filter((apiKey) => apiKey.state === "active").length;
  const rotationRequiredKeys = source.workspaceApiKeys.apiKeys.filter((apiKey) => apiKey.state === "rotation_required").length;
  const failedRuns = source.workspaceRunHistory.runs.filter((run) => run.failureCode !== "none").length;
  const deniedAuditEvents = source.adminAuditLog.auditEvents.filter((event) => event.decision === "denied").length;

  return [
    {
      rollupId: "key_scope_inventory",
      label: "Key scope inventory",
      value: `${activeKeys}/${source.workspaceApiKeys.apiKeys.length} active`,
      status: source.workspaceApiKeys.canRenderApiKeys ? "offline_only" : "blocked",
      sourceRef: source.workspaceApiKeys.routeId,
      summary: "Key ids, owners, scopes, state, and timestamps are visible without raw key value, hash, issuance, or validation.",
    },
    {
      rollupId: "rotation_review",
      label: "Rotation review",
      value: String(rotationRequiredKeys),
      status: rotationRequiredKeys > 0 ? "review_required" : "ready",
      sourceRef: source.workspaceApiKeys.requestId,
      summary: "Rotation-required keys are surfaced as operator review evidence, not lifecycle mutation capability.",
    },
    {
      rollupId: "request_usage",
      label: "Request usage",
      value: quota ? `${quota.usage_snapshot.request_count}/${quota.request_limit}` : "missing",
      status: quota ? "offline_only" : "blocked",
      sourceRef: source.workspaceUsageQuota.routeId,
      summary: "Request usage is a read-side quota snapshot and does not enforce gateway traffic.",
    },
    {
      rollupId: "token_usage",
      label: "Token usage",
      value: quota ? `${quota.usage_snapshot.token_count}/${quota.token_limit}` : "missing",
      status: quota ? "offline_only" : "blocked",
      sourceRef: source.workspaceUsageQuota.requestId,
      summary: "Token usage is displayed as a policy snapshot without live provider accounting or durable writeback.",
    },
    {
      rollupId: "trace_failures",
      label: "Trace failures",
      value: `${failedRuns}/${source.workspaceRunHistory.runs.length}`,
      status: failedRuns > 0 ? "review_required" : "ready",
      sourceRef: source.workspaceRunHistory.routeId,
      summary: "Run failure codes and trace ids remain readable without replay, resume, or materialized result access.",
    },
    {
      rollupId: "audit_denials",
      label: "Audit denials",
      value: String(deniedAuditEvents),
      status: deniedAuditEvents > 0 ? "review_required" : "ready",
      sourceRef: source.adminAuditLog.routeId,
      summary: "Denied decisions are audit evidence only; raw payload export and audit record mutation remain locked.",
    },
    {
      rollupId: "estimated_cost",
      label: "Estimated cost",
      value: quota ? `$${quota.usage_snapshot.estimated_cost.toFixed(2)}` : "missing",
      status: quota ? "offline_only" : "blocked",
      sourceRef: source.workspaceUsageQuota.auditRef,
      summary: "Estimated cost is a read-side snapshot and cannot create or modify a cost record.",
    },
  ];
}

function buildKeyScopes(source: ModelGatewayUsageAuditEvidenceSource): ModelGatewayKeyScopeEvidence[] {
  return source.workspaceApiKeys.apiKeys.map((apiKey) => ({
    keyId: apiKey.apiKeyId,
    ownerSubjectRef: apiKey.ownerSubjectRef,
    scopes: apiKey.scopes,
    state: apiKey.state,
    lastUsedAt: apiKey.lastUsedAt ?? "not used",
    status: apiKey.state === "active" ? "offline_only" : "review_required",
    sourceRef: source.workspaceApiKeys.auditRef,
    summary:
      apiKey.state === "active"
        ? "Active key summary can explain intended read access without validating or revealing key material."
        : "Rotation-required key summary is visible for review without create, rotate, revoke, hash, or reveal operations.",
  }));
}

function buildQuotaCosts(source: ModelGatewayUsageAuditEvidenceSource): ModelGatewayQuotaCostEvidence[] {
  const quota = source.workspaceUsageQuota.quota;
  if (!quota) {
    return [
      {
        evidenceId: "quota_missing",
        label: "Quota policy",
        limit: "missing",
        used: "missing",
        usage: "blocked",
        status: "blocked",
        sourceRef: source.workspaceUsageQuota.routeId,
        summary: "Quota summary is unavailable, so gateway usage evidence remains blocked.",
      },
    ];
  }
  return [
    {
      evidenceId: "request_limit",
      label: "Requests",
      limit: quota.request_limit.toLocaleString("en-US"),
      used: quota.usage_snapshot.request_count.toLocaleString("en-US"),
      usage: `${usagePercent(quota.usage_snapshot.request_count, quota.request_limit)} used`,
      status: "offline_only",
      sourceRef: source.workspaceUsageQuota.routeId,
      summary: "Request count supports usage review but does not apply rate limit or reject live gateway calls.",
    },
    {
      evidenceId: "token_limit",
      label: "Tokens",
      limit: quota.token_limit.toLocaleString("en-US"),
      used: quota.usage_snapshot.token_count.toLocaleString("en-US"),
      usage: `${usagePercent(quota.usage_snapshot.token_count, quota.token_limit)} used`,
      status: "offline_only",
      sourceRef: source.workspaceUsageQuota.requestId,
      summary: "Token count is an offline usage snapshot and is not a live provider metering system.",
    },
    {
      evidenceId: "cost_limit",
      label: "Cost",
      limit: `$${quota.cost_limit.toFixed(2)}`,
      used: `$${quota.usage_snapshot.estimated_cost.toFixed(2)}`,
      usage: `${usagePercent(quota.usage_snapshot.estimated_cost, quota.cost_limit)} used`,
      status: "offline_only",
      sourceRef: source.workspaceUsageQuota.auditRef,
      summary: "Cost evidence remains estimated and read-only; durable pricing policy and cost record writes stay locked.",
    },
    {
      evidenceId: "over_quota_failure",
      label: "Failure code",
      limit: quota.over_quota_failure_code,
      used: "not enforced",
      usage: "policy evidence",
      status: "locked",
      sourceRef: source.workspaceUsageQuota.overQuotaFailureCode,
      summary: "The failure code is documented for future gateway policy, but this surface cannot enforce quota.",
    },
  ];
}

function buildTraceAudit(source: ModelGatewayUsageAuditEvidenceSource): ModelGatewayTraceAuditEvidence[] {
  return source.workspaceRunHistory.runs.map((run) => ({
    traceId: run.traceId,
    runId: run.runId,
    applicationRef: run.applicationRef,
    workflowDefinitionId: run.workflowDefinitionId,
    runStatus: run.status,
    failureCode: run.failureCode,
    estimatedCost: run.estimatedCost,
    auditRef: source.workspaceRunHistory.auditRef,
    status: run.failureCode === "none" ? "ready" : "review_required",
    summary:
      run.failureCode === "none"
        ? "Successful run evidence can be inspected without result materialization or replay controls."
        : "Failed run evidence preserves failure classification for review without retry, replay, or writeback.",
  }));
}

function buildAuditCorrelations(
  source: ModelGatewayUsageAuditEvidenceSource,
): ModelGatewayAuditCorrelationEvidence[] {
  return source.adminAuditLog.auditEvents.map((event) => ({
    auditRef: event.auditRef,
    actorSubjectRef: event.actorSubjectRef,
    eventKind: event.eventKind,
    resourceRef: event.resourceRef,
    decision: event.decision,
    failureCode: event.failureCode,
    traceId: event.traceId,
    status: event.decision === "denied" || event.failureCode !== "none" ? "review_required" : "ready",
    summary:
      event.decision === "denied"
        ? "Denied audit event remains visible for governance review without raw payload export."
        : "Allowed audit event anchors read access evidence without audit record mutation.",
  }));
}

function buildGuards(source: ModelGatewayUsageAuditEvidenceSource): ModelGatewayUsageAuditGuard[] {
  return [
    {
      guardId: "raw_key_hidden",
      label: "Raw key hidden",
      status: "locked",
      sourceRef: source.workspaceApiKeys.routeId,
      missingPrerequisite: "API key issuance, hashing, storage, reveal, and validation contract",
      summary: "This detail surface shows key summaries only and cannot reveal, create, rotate, revoke, hash, or validate keys.",
    },
    {
      guardId: "quota_execution_locked",
      label: "Quota execution locked",
      status: "locked",
      sourceRef: source.workspaceUsageQuota.routeId,
      missingPrerequisite: "gateway quota policy execution and rate limit",
      summary: "Quota limits and usage are readable policy evidence and do not block or throttle live requests.",
    },
    {
      guardId: "cost_record_locked",
      label: "Cost record locked",
      status: "locked",
      sourceRef: source.workspaceUsageQuota.auditRef,
      missingPrerequisite: "durable pricing policy and cost record writer",
      summary: "Estimated cost is a snapshot; no durable cost record is created or updated by this page.",
    },
    {
      guardId: "trace_replay_locked",
      label: "Replay locked",
      status: "locked",
      sourceRef: source.workspaceRunHistory.routeId,
      missingPrerequisite: "executor, durable run store, materialized result reader, and replay policy",
      summary: "Trace and run evidence can be inspected, but replay, resume, result materialization, and writeback stay unavailable.",
    },
    {
      guardId: "audit_payload_locked",
      label: "Raw audit payload locked",
      status: "locked",
      sourceRef: source.adminAuditLog.routeId,
      missingPrerequisite: "audit export policy and sensitive projection review",
      summary: "Audit summaries are visible without raw payload export, secret reveal, edit, delete, or durable audit mutation.",
    },
    {
      guardId: "gateway_dependency",
      label: "Gateway evidence dependency",
      status:
        source.overview.canRenderModelGatewayOverview && source.routeEvidence.canRenderRouteEvidenceDetail
          ? "ready"
          : "blocked",
      sourceRef: source.routeEvidence.pageId,
      missingPrerequisite: "overview and route evidence must render first",
      summary: "Usage, trace, and audit evidence stays subordinate to the model gateway evidence chain.",
    },
  ];
}

function usagePercent(used: number, limit: number): string {
  if (limit <= 0) {
    return "0.0%";
  }
  return `${((used / limit) * 100).toFixed(1)}%`;
}
