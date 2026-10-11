import type { EvidenceProjectionCopy, EvidenceProjectionMessage } from "./evidenceProjectionCopy.ts";
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

export type ModelGatewayUsageAuditRollup = EvidenceProjectionCopy & {
  rollupId: string;
  label: string;
  value: string;
  status: ModelGatewayUsageAuditEvidenceStatus;
  sourceRef: string;
  summary: string;
};

export type ModelGatewayKeyScopeEvidence = EvidenceProjectionCopy & {
  keyId: string;
  ownerSubjectRef: string;
  scopes: string[];
  state: string;
  lastUsedAt: string;
  status: ModelGatewayUsageAuditEvidenceStatus;
  sourceRef: string;
  summary: string;
};

export type ModelGatewayQuotaCostEvidence = EvidenceProjectionCopy & {
  evidenceId: string;
  label: string;
  limit: string;
  used: string;
  usage: string;
  status: ModelGatewayUsageAuditEvidenceStatus;
  sourceRef: string;
  summary: string;
};

export type ModelGatewayTraceAuditEvidence = EvidenceProjectionCopy & {
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

export type ModelGatewayAuditCorrelationEvidence = EvidenceProjectionCopy & {
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

export type ModelGatewayUsageAuditGuard = EvidenceProjectionCopy & {
  guardId: string;
  label: string;
  status: ModelGatewayUsageAuditEvidenceStatus;
  sourceRef: string;
  missingPrerequisite: string;
  summary: string;
};

export type ModelGatewayUsageAuditEvidenceViewModel = EvidenceProjectionCopy & {
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
    evidenceNarrativeMessage: buildNarrativeMessage(source, rollups, traceAudit, auditCorrelations),
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
      labelMessage: { key: "gatewayUsageEvidence_key_scope_inventory_label" as const },
      value: `${activeKeys}/${source.workspaceApiKeys.apiKeys.length} active`,
      valueMessage: { key: "gatewayUsageEvidence_key_scope_inventory_value" as const, values: { activeKeys: activeKeys, apiKeysCount: source.workspaceApiKeys.apiKeys.length } },
      status: source.workspaceApiKeys.canRenderApiKeys ? "offline_only" : "blocked",
      sourceRef: source.workspaceApiKeys.routeId,
      summary: "Key ids, owners, scopes, state, and timestamps are visible without raw key value, hash, issuance, or validation.",
      summaryMessage: { key: "gatewayUsageEvidence_key_scope_inventory_summary" as const },
    },
    {
      rollupId: "rotation_review",
      label: "Rotation review",
      labelMessage: { key: "gatewayUsageEvidence_rotation_review_label" as const },
      value: String(rotationRequiredKeys),
      status: rotationRequiredKeys > 0 ? "review_required" : "ready",
      sourceRef: source.workspaceApiKeys.requestId,
      summary: "Rotation-required keys are surfaced as operator review evidence, not lifecycle mutation capability.",
      summaryMessage: { key: "gatewayUsageEvidence_rotation_review_summary" as const },
    },
    {
      rollupId: "request_usage",
      label: "Request usage",
      labelMessage: { key: "gatewayUsageEvidence_request_usage_label" as const },
      value: quota ? `${quota.usage_snapshot.request_count}/${quota.request_limit}` : "missing",
      valueMessage: quota ? { key: "gatewayUsageEvidence_request_usage_value_yes" as const, values: { count: quota.usage_snapshot.request_count, limit: quota.request_limit } } : { key: "gatewayUsageEvidence_request_usage_value_no" as const },
      status: quota ? "offline_only" : "blocked",
      sourceRef: source.workspaceUsageQuota.routeId,
      summary: "Request usage is a read-side quota snapshot and does not enforce gateway traffic.",
      summaryMessage: { key: "gatewayUsageEvidence_request_usage_summary" as const },
    },
    {
      rollupId: "token_usage",
      label: "Token usage",
      labelMessage: { key: "gatewayUsageEvidence_token_usage_label" as const },
      value: quota ? `${quota.usage_snapshot.token_count}/${quota.token_limit}` : "missing",
      valueMessage: quota ? { key: "gatewayUsageEvidence_token_usage_value_yes" as const, values: { count: quota.usage_snapshot.token_count, limit: quota.token_limit } } : { key: "gatewayUsageEvidence_token_usage_value_no" as const },
      status: quota ? "offline_only" : "blocked",
      sourceRef: source.workspaceUsageQuota.requestId,
      summary: "Token usage is displayed as a policy snapshot without live provider accounting or durable writeback.",
      summaryMessage: { key: "gatewayUsageEvidence_token_usage_summary" as const },
    },
    {
      rollupId: "trace_failures",
      label: "Trace failures",
      labelMessage: { key: "gatewayUsageEvidence_trace_failures_label" as const },
      value: `${failedRuns}/${source.workspaceRunHistory.runs.length}`,
      valueMessage: { key: "gatewayUsageEvidence_trace_failures_value" as const, values: { failedRuns: failedRuns, runsCount: source.workspaceRunHistory.runs.length } },
      status: failedRuns > 0 ? "review_required" : "ready",
      sourceRef: source.workspaceRunHistory.routeId,
      summary: "Run failure codes and trace ids remain readable without replay, resume, or materialized result access.",
      summaryMessage: { key: "gatewayUsageEvidence_trace_failures_summary" as const },
    },
    {
      rollupId: "audit_denials",
      label: "Audit denials",
      labelMessage: { key: "gatewayUsageEvidence_audit_denials_label" as const },
      value: String(deniedAuditEvents),
      status: deniedAuditEvents > 0 ? "review_required" : "ready",
      sourceRef: source.adminAuditLog.routeId,
      summary: "Denied decisions are audit evidence only; raw payload export and audit record mutation remain locked.",
      summaryMessage: { key: "gatewayUsageEvidence_audit_denials_summary" as const },
    },
    {
      rollupId: "estimated_cost",
      label: "Estimated cost",
      labelMessage: { key: "gatewayUsageEvidence_estimated_cost_label" as const },
      value: quota ? `$${quota.usage_snapshot.estimated_cost.toFixed(2)}` : "missing",
      valueMessage: quota ? { key: "gatewayUsageEvidence_estimated_cost_value_yes" as const, values: { cost: { number: quota.usage_snapshot.estimated_cost, style: "money" } } } : { key: "gatewayUsageEvidence_estimated_cost_value_no" as const },
      status: quota ? "offline_only" : "blocked",
      sourceRef: source.workspaceUsageQuota.auditRef,
      summary: "Estimated cost is a read-side snapshot and cannot create or modify a cost record.",
      summaryMessage: { key: "gatewayUsageEvidence_estimated_cost_summary" as const },
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
    summaryMessage: apiKey.state === "active" ? { key: "gatewayUsageEvidence_buildKeyScopes_summary_yes" as const } : { key: "gatewayUsageEvidence_buildKeyScopes_summary_no" as const },
  }));
}

function buildQuotaCosts(source: ModelGatewayUsageAuditEvidenceSource): ModelGatewayQuotaCostEvidence[] {
  const quota = source.workspaceUsageQuota.quota;
  if (!quota) {
    return [
      {
        evidenceId: "quota_missing",
        label: "Quota policy",
        labelMessage: { key: "gatewayUsageEvidence_quota_missing_label" as const },
        limit: "missing",
        limitMessage: { key: "gatewayUsageEvidence_quota_missing_limit" as const },
        used: "missing",
        usedMessage: { key: "gatewayUsageEvidence_quota_missing_used" as const },
        usage: "blocked",
        usageMessage: { key: "gatewayUsageEvidence_quota_missing_usage" as const },
        status: "blocked",
        sourceRef: source.workspaceUsageQuota.routeId,
        summary: "Quota summary is unavailable, so gateway usage evidence remains blocked.",
        summaryMessage: { key: "gatewayUsageEvidence_quota_missing_summary" as const },
      },
    ];
  }
  return [
    {
      evidenceId: "request_limit",
      label: "Requests",
      labelMessage: { key: "gatewayUsageEvidence_request_limit_label" as const },
      limit: quota.request_limit.toLocaleString("en-US"),
      limitMessage: { key: "gatewayUsageEvidence_number", values: { number: quota.request_limit } },
      used: quota.usage_snapshot.request_count.toLocaleString("en-US"),
      usedMessage: { key: "gatewayUsageEvidence_number", values: { number: quota.usage_snapshot.request_count } },
      usage: `${usagePercent(quota.usage_snapshot.request_count, quota.request_limit)} used`,
      usageMessage: { key: "gatewayUsageEvidence_request_limit_usage" as const, values: { limit: { number: quota.request_limit > 0 ? quota.usage_snapshot.request_count / quota.request_limit : 0, style: "percent" } } },
      status: "offline_only",
      sourceRef: source.workspaceUsageQuota.routeId,
      summary: "Request count supports usage review but does not apply rate limit or reject live gateway calls.",
      summaryMessage: { key: "gatewayUsageEvidence_request_limit_summary" as const },
    },
    {
      evidenceId: "token_limit",
      label: "Tokens",
      labelMessage: { key: "gatewayUsageEvidence_token_limit_label" as const },
      limit: quota.token_limit.toLocaleString("en-US"),
      limitMessage: { key: "gatewayUsageEvidence_number", values: { number: quota.token_limit } },
      used: quota.usage_snapshot.token_count.toLocaleString("en-US"),
      usedMessage: { key: "gatewayUsageEvidence_number", values: { number: quota.usage_snapshot.token_count } },
      usage: `${usagePercent(quota.usage_snapshot.token_count, quota.token_limit)} used`,
      usageMessage: { key: "gatewayUsageEvidence_token_limit_usage" as const, values: { limit: { number: quota.token_limit > 0 ? quota.usage_snapshot.token_count / quota.token_limit : 0, style: "percent" } } },
      status: "offline_only",
      sourceRef: source.workspaceUsageQuota.requestId,
      summary: "Token count is an offline usage snapshot and is not a live provider metering system.",
      summaryMessage: { key: "gatewayUsageEvidence_token_limit_summary" as const },
    },
    {
      evidenceId: "cost_limit",
      label: "Cost",
      labelMessage: { key: "gatewayUsageEvidence_cost_limit_label" as const },
      limit: `$${quota.cost_limit.toFixed(2)}`,
      limitMessage: { key: "gatewayUsageEvidence_cost_limit_limit" as const, values: { limit: { number: quota.cost_limit, style: "money" } } },
      used: `$${quota.usage_snapshot.estimated_cost.toFixed(2)}`,
      usedMessage: { key: "gatewayUsageEvidence_cost_limit_used" as const, values: { cost: { number: quota.usage_snapshot.estimated_cost, style: "money" } } },
      usage: `${usagePercent(quota.usage_snapshot.estimated_cost, quota.cost_limit)} used`,
      usageMessage: { key: "gatewayUsageEvidence_cost_limit_usage" as const, values: { limit: { number: quota.cost_limit > 0 ? quota.usage_snapshot.estimated_cost / quota.cost_limit : 0, style: "percent" } } },
      status: "offline_only",
      sourceRef: source.workspaceUsageQuota.auditRef,
      summary: "Cost evidence remains estimated and read-only; durable pricing policy and cost record writes stay locked.",
      summaryMessage: { key: "gatewayUsageEvidence_cost_limit_summary" as const },
    },
    {
      evidenceId: "over_quota_failure",
      label: "Failure code",
      labelMessage: { key: "gatewayUsageEvidence_over_quota_failure_label" as const },
      limit: quota.over_quota_failure_code,
      used: "not enforced",
      usedMessage: { key: "gatewayUsageEvidence_over_quota_failure_used" as const },
      usage: "policy evidence",
      usageMessage: { key: "gatewayUsageEvidence_over_quota_failure_usage" as const },
      status: "locked",
      sourceRef: source.workspaceUsageQuota.overQuotaFailureCode,
      summary: "The failure code is documented for future gateway policy, but this surface cannot enforce quota.",
      summaryMessage: { key: "gatewayUsageEvidence_over_quota_failure_summary" as const },
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
    summaryMessage: run.failureCode === "none" ? { key: "gatewayUsageEvidence_buildTraceAudit_summary_yes" as const } : { key: "gatewayUsageEvidence_buildTraceAudit_summary_no" as const },
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
    summaryMessage: event.decision === "denied" ? { key: "gatewayUsageEvidence_buildAuditCorrelations_summary_yes" as const } : { key: "gatewayUsageEvidence_buildAuditCorrelations_summary_no" as const },
  }));
}

function buildGuards(source: ModelGatewayUsageAuditEvidenceSource): ModelGatewayUsageAuditGuard[] {
  return [
    {
      guardId: "raw_key_hidden",
      label: "Raw key hidden",
      labelMessage: { key: "gatewayUsageEvidence_raw_key_hidden_label" as const },
      status: "locked",
      sourceRef: source.workspaceApiKeys.routeId,
      missingPrerequisite: "API key issuance, hashing, storage, reveal, and validation contract",
      missingPrerequisiteMessage: { key: "gatewayUsageEvidence_raw_key_hidden_missingPrerequisite" as const },
      summary: "This detail surface shows key summaries only and cannot reveal, create, rotate, revoke, hash, or validate keys.",
      summaryMessage: { key: "gatewayUsageEvidence_raw_key_hidden_summary" as const },
    },
    {
      guardId: "quota_execution_locked",
      label: "Quota execution locked",
      labelMessage: { key: "gatewayUsageEvidence_quota_execution_locked_label" as const },
      status: "locked",
      sourceRef: source.workspaceUsageQuota.routeId,
      missingPrerequisite: "gateway quota policy execution and rate limit",
      missingPrerequisiteMessage: { key: "gatewayUsageEvidence_quota_execution_locked_missingPrerequisite" as const },
      summary: "Quota limits and usage are readable policy evidence and do not block or throttle live requests.",
      summaryMessage: { key: "gatewayUsageEvidence_quota_execution_locked_summary" as const },
    },
    {
      guardId: "cost_record_locked",
      label: "Cost record locked",
      labelMessage: { key: "gatewayUsageEvidence_cost_record_locked_label" as const },
      status: "locked",
      sourceRef: source.workspaceUsageQuota.auditRef,
      missingPrerequisite: "durable pricing policy and cost record writer",
      missingPrerequisiteMessage: { key: "gatewayUsageEvidence_cost_record_locked_missingPrerequisite" as const },
      summary: "Estimated cost is a snapshot; no durable cost record is created or updated by this page.",
      summaryMessage: { key: "gatewayUsageEvidence_cost_record_locked_summary" as const },
    },
    {
      guardId: "trace_replay_locked",
      label: "Replay locked",
      labelMessage: { key: "gatewayUsageEvidence_trace_replay_locked_label" as const },
      status: "locked",
      sourceRef: source.workspaceRunHistory.routeId,
      missingPrerequisite: "executor, durable run store, materialized result reader, and replay policy",
      missingPrerequisiteMessage: { key: "gatewayUsageEvidence_trace_replay_locked_missingPrerequisite" as const },
      summary: "Trace and run evidence can be inspected, but replay, resume, result materialization, and writeback stay unavailable.",
      summaryMessage: { key: "gatewayUsageEvidence_trace_replay_locked_summary" as const },
    },
    {
      guardId: "audit_payload_locked",
      label: "Raw audit payload locked",
      labelMessage: { key: "gatewayUsageEvidence_audit_payload_locked_label" as const },
      status: "locked",
      sourceRef: source.adminAuditLog.routeId,
      missingPrerequisite: "audit export policy and sensitive projection review",
      missingPrerequisiteMessage: { key: "gatewayUsageEvidence_audit_payload_locked_missingPrerequisite" as const },
      summary: "Audit summaries are visible without raw payload export, secret reveal, edit, delete, or durable audit mutation.",
      summaryMessage: { key: "gatewayUsageEvidence_audit_payload_locked_summary" as const },
    },
    {
      guardId: "gateway_dependency",
      label: "Gateway evidence dependency",
      labelMessage: { key: "gatewayUsageEvidence_gateway_dependency_label" as const },
      status:
        source.overview.canRenderModelGatewayOverview && source.routeEvidence.canRenderRouteEvidenceDetail
          ? "ready"
          : "blocked",
      sourceRef: source.routeEvidence.pageId,
      missingPrerequisite: "overview and route evidence must render first",
      missingPrerequisiteMessage: { key: "gatewayUsageEvidence_gateway_dependency_missingPrerequisite" as const },
      summary: "Usage, trace, and audit evidence stays subordinate to the model gateway evidence chain.",
      summaryMessage: { key: "gatewayUsageEvidence_gateway_dependency_summary" as const },
    },
  ];
}

function usagePercent(used: number, limit: number): string {
  if (limit <= 0) {
    return "0.0%";
  }
  return `${((used / limit) * 100).toFixed(1)}%`;
}

function buildNarrativeMessage(
  source: ModelGatewayUsageAuditEvidenceSource,
  rollups: ModelGatewayUsageAuditRollup[],
  traceAudit: ModelGatewayTraceAuditEvidence[],
  auditCorrelations: ModelGatewayAuditCorrelationEvidence[],
): EvidenceProjectionMessage {
  const activeKeys = source.workspaceApiKeys.apiKeys.filter((apiKey) => apiKey.state === "active").length;
  const failedRuns = traceAudit.filter((run) => run.failureCode !== "none").length;
  const deniedAuditEvents = auditCorrelations.filter((event) => event.decision === "denied").length;
  return { key: "gatewayUsageEvidence_evidenceNarrative" as const, values: { rollupsCount: rollups.length, apiKeysCount: source.workspaceApiKeys.apiKeys.length, activeKeys: activeKeys, traceAuditCount: traceAudit.length, failedRuns: failedRuns, deniedAuditEvents: deniedAuditEvents } };
}
