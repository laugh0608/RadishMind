import type { EvidenceProjectionCopy, EvidenceProjectionMessage } from "./evidenceProjectionCopy.ts";
import type { AdminAuditLogViewModel } from "./adminAuditLog";
import type { ControlPlaneReadShellViewModel } from "./readShell";
import type { WorkspaceApiKeysViewModel } from "./workspaceApiKeys";
import type { WorkspaceRunHistoryViewModel } from "./workspaceRunHistory";
import type { WorkspaceUsageQuotaViewModel } from "./workspaceUsageQuota";

export type ModelGatewayOverviewStatus = "ready" | "offline_only" | "review_required" | "blocked" | "locked";

export type ModelGatewayOverviewSource = {
  readShell: ControlPlaneReadShellViewModel;
  workspaceApiKeys: WorkspaceApiKeysViewModel;
  workspaceUsageQuota: WorkspaceUsageQuotaViewModel;
  workspaceRunHistory: WorkspaceRunHistoryViewModel;
  adminAuditLog: AdminAuditLogViewModel;
};

export type ModelGatewayOverviewMetric = EvidenceProjectionCopy & {
  metricId: string;
  label: string;
  value: string;
  status: ModelGatewayOverviewStatus;
  summary: string;
};

export type ModelGatewayApiSurface = EvidenceProjectionCopy & {
  surfaceId: string;
  label: string;
  path: string;
  status: ModelGatewayOverviewStatus;
  evidenceRef: string;
  summary: string;
};

export type ModelGatewayPolicyEvidence = EvidenceProjectionCopy & {
  evidenceId: string;
  label: string;
  value: string;
  status: ModelGatewayOverviewStatus;
  sourceRef: string;
  summary: string;
};

export type ModelGatewayTraceEvidence = EvidenceProjectionCopy & {
  traceId: string;
  runId: string;
  status: ModelGatewayOverviewStatus;
  cost: string;
  failureCode: string;
  auditRef: string;
  summary: string;
};

export type ModelGatewayBoundaryLock = EvidenceProjectionCopy & {
  lockId: string;
  label: string;
  status: ModelGatewayOverviewStatus;
  missingPrerequisite: string;
  summary: string;
};

export type ModelGatewayOverviewViewModel = EvidenceProjectionCopy & {
  pageId: "model-gateway-overview-offline";
  sourcePageIds: string[];
  overviewMode: "offline_read_only_gateway_evidence";
  tenantRef: string;
  requestId: string;
  auditRef: string;
  gatewayNarrative: string;
  metrics: ModelGatewayOverviewMetric[];
  apiSurfaces: ModelGatewayApiSurface[];
  policyEvidence: ModelGatewayPolicyEvidence[];
  traceEvidence: ModelGatewayTraceEvidence[];
  boundaryLocks: ModelGatewayBoundaryLock[];
  canRenderModelGatewayOverview: boolean;
  canInspectGatewayLocally: true;
  canConsumeDevReadEvidence: boolean;
  canRequestProductionGateway: false;
  canIssueApiKey: false;
  canValidateApiKey: false;
  canEnforceQuota: false;
  canApplyRateLimit: false;
  canWriteCostRecord: false;
  canResolveProductionSecret: false;
  canExecuteFallback: false;
  canAttachDatabase: false;
  canEnableRadishAuth: false;
};

export function buildModelGatewayOverviewViewModel(
  source: ModelGatewayOverviewSource,
): ModelGatewayOverviewViewModel {
  const metrics = buildMetrics(source);
  const apiSurfaces = buildApiSurfaces(source);
  const policyEvidence = buildPolicyEvidence(source);
  const traceEvidence = buildTraceEvidence(source);
  const boundaryLocks = buildBoundaryLocks();

  return {
    pageId: "model-gateway-overview-offline",
    sourcePageIds: [
      "control-plane-read-shared-shell",
      source.workspaceApiKeys.pageId,
      source.workspaceUsageQuota.pageId,
      source.workspaceRunHistory.pageId,
      source.adminAuditLog.pageId,
      "provider-runtime-health-v1",
      "gateway-api-key-quota-readiness",
    ],
    overviewMode: "offline_read_only_gateway_evidence",
    tenantRef: source.workspaceUsageQuota.collection.tenantRef,
    requestId: source.workspaceRunHistory.requestId,
    auditRef: source.adminAuditLog.auditRef,
    gatewayNarrative: buildGatewayNarrative(source),
    gatewayNarrativeMessage: buildGatewayNarrativeMessage(source),
    metrics,
    apiSurfaces,
    policyEvidence,
    traceEvidence,
    boundaryLocks,
    canRenderModelGatewayOverview:
      source.readShell.usesCanonicalRoutes &&
      source.workspaceApiKeys.canRenderApiKeys &&
      source.workspaceUsageQuota.canRenderQuota &&
      source.workspaceRunHistory.canRenderRuns &&
      source.adminAuditLog.canRenderAuditLog &&
      metrics.length >= 6 &&
      apiSurfaces.length >= 5 &&
      policyEvidence.length >= 6 &&
      boundaryLocks.length >= 6,
    canInspectGatewayLocally: true,
    canConsumeDevReadEvidence:
      source.workspaceApiKeys.canRequestLiveBackend ||
      source.workspaceUsageQuota.canRequestLiveBackend ||
      source.workspaceRunHistory.canRequestLiveBackend ||
      source.adminAuditLog.canRequestLiveBackend,
    canRequestProductionGateway: false,
    canIssueApiKey: false,
    canValidateApiKey: false,
    canEnforceQuota: false,
    canApplyRateLimit: false,
    canWriteCostRecord: false,
    canResolveProductionSecret: false,
    canExecuteFallback: false,
    canAttachDatabase: false,
    canEnableRadishAuth: false,
  };
}

function buildGatewayNarrative(source: ModelGatewayOverviewSource): string {
  const activeKeys = source.workspaceApiKeys.apiKeys.filter((apiKey) => apiKey.state === "active").length;
  return `${source.readShell.catalog.routes.length} read-side routes, ${source.workspaceApiKeys.apiKeys.length} API key summaries, ${activeKeys} active key reference, ${source.workspaceRunHistory.runs.length} trace-bearing run records, quota usage, and audit evidence are grouped as offline model gateway evidence.`;
}

function buildMetrics(source: ModelGatewayOverviewSource): ModelGatewayOverviewMetric[] {
  const activeKeys = source.workspaceApiKeys.apiKeys.filter((apiKey) => apiKey.state === "active").length;
  const failedRuns = source.workspaceRunHistory.runs.filter((run) => run.failureCode !== "none").length;

  return [
    {
      metricId: "northbound_surfaces",
      label: "Northbound surfaces",
      labelMessage: { key: "gatewayOverviewEvidence_northbound_surfaces_label" as const },
      value: "5",
      status: "offline_only",
      summary: "Models, chat completions, responses, messages, and model lookup stay visible as compatibility evidence.",
      summaryMessage: { key: "gatewayOverviewEvidence_northbound_surfaces_summary" as const },
    },
    {
      metricId: "read_routes",
      label: "Read routes",
      labelMessage: { key: "gatewayOverviewEvidence_read_routes_label" as const },
      value: String(source.readShell.catalog.routes.length),
      status: source.readShell.usesCanonicalRoutes ? "ready" : "blocked",
      summary: "Gateway overview reuses the shared read catalog instead of creating a second product contract.",
      summaryMessage: { key: "gatewayOverviewEvidence_read_routes_summary" as const },
    },
    {
      metricId: "api_keys",
      label: "API keys",
      labelMessage: { key: "gatewayOverviewEvidence_api_keys_label" as const },
      value: `${activeKeys}/${source.workspaceApiKeys.apiKeys.length}`,
      valueMessage: { key: "gatewayOverviewEvidence_api_keys_value" as const, values: { activeKeys: activeKeys, apiKeysCount: source.workspaceApiKeys.apiKeys.length } },
      status: source.workspaceApiKeys.canRenderApiKeys ? "offline_only" : "blocked",
      summary: "Key summaries are displayed without key value, hash, validation, issuance, or lifecycle mutation.",
      summaryMessage: { key: "gatewayOverviewEvidence_api_keys_summary" as const },
    },
    {
      metricId: "quota",
      label: "Quota",
      labelMessage: { key: "gatewayOverviewEvidence_quota_label" as const },
      value: quotaUsageLabel(source.workspaceUsageQuota),
      valueMessage: source.workspaceUsageQuota.quota
        ? { key: "gatewayOverviewEvidence_quota_counts", values: { used: source.workspaceUsageQuota.quota.usage_snapshot.request_count, limit: source.workspaceUsageQuota.quota.request_limit } }
        : { key: "gatewayOverviewEvidence_quota_policy_value_no" },
      status: source.workspaceUsageQuota.canRenderQuota ? "offline_only" : "blocked",
      summary: "Quota policy and usage snapshot are evidence only; enforcement and rate limit remain locked.",
      summaryMessage: { key: "gatewayOverviewEvidence_quota_summary" as const },
    },
    {
      metricId: "trace_cost",
      label: "Trace and cost",
      labelMessage: { key: "gatewayOverviewEvidence_trace_cost_label" as const },
      value: `${source.workspaceRunHistory.runs.length} / ${failedRuns} failed`,
      valueMessage: { key: "gatewayOverviewEvidence_trace_cost_value" as const, values: { runsCount: source.workspaceRunHistory.runs.length, failedRuns: failedRuns } },
      status: failedRuns > 0 ? "review_required" : "ready",
      summary: "Run trace ids, failure codes, and estimated cost summaries are readable without replay or cost writeback.",
      summaryMessage: { key: "gatewayOverviewEvidence_trace_cost_summary" as const },
    },
    {
      metricId: "audit_events",
      label: "Audit events",
      labelMessage: { key: "gatewayOverviewEvidence_audit_events_label" as const },
      value: String(source.adminAuditLog.auditEvents.length),
      status: source.adminAuditLog.canRenderAuditLog ? "ready" : "blocked",
      summary: "Audit summaries anchor read access and denied decisions without raw payload export.",
      summaryMessage: { key: "gatewayOverviewEvidence_audit_events_summary" as const },
    },
    {
      metricId: "locked_capabilities",
      label: "Locked capabilities",
      labelMessage: { key: "gatewayOverviewEvidence_locked_capabilities_label" as const },
      value: String(buildBoundaryLocks().length),
      status: "locked",
      summary: "Production auth, API key issuance, quota enforcement, cost record writes, secret resolver, and fallback execution stay disabled.",
      summaryMessage: { key: "gatewayOverviewEvidence_locked_capabilities_summary" as const },
    },
  ];
}

function buildApiSurfaces(source: ModelGatewayOverviewSource): ModelGatewayApiSurface[] {
  const auditRef = source.adminAuditLog.auditRef;
  return [
    {
      surfaceId: "models_inventory",
      label: "Models inventory",
      labelMessage: { key: "gatewayOverviewEvidence_models_inventory_label" as const },
      path: "/v1/models",
      status: "offline_only",
      evidenceRef: "provider-runtime-docs-refresh",
      summary: "Provider-qualified profile inventory remains a read/discovery surface, not a provisioning API.",
      summaryMessage: { key: "gatewayOverviewEvidence_models_inventory_summary" as const },
    },
    {
      surfaceId: "model_lookup",
      label: "Model lookup",
      labelMessage: { key: "gatewayOverviewEvidence_model_lookup_label" as const },
      path: "/v1/models/{id}",
      status: "offline_only",
      evidenceRef: "provider-selection-policy-v1",
      summary: "Lookup evidence keeps provider/model selection explicit and does not enable implicit fallback.",
      summaryMessage: { key: "gatewayOverviewEvidence_model_lookup_summary" as const },
    },
    {
      surfaceId: "chat_completions",
      label: "Chat completions",
      labelMessage: { key: "gatewayOverviewEvidence_chat_completions_label" as const },
      path: "/v1/chat/completions",
      status: "offline_only",
      evidenceRef: auditRef,
      summary: "Compatibility route evidence stays tied to canonical gateway envelope and request metadata.",
      summaryMessage: { key: "gatewayOverviewEvidence_chat_completions_summary" as const },
    },
    {
      surfaceId: "responses",
      label: "Responses",
      labelMessage: { key: "gatewayOverviewEvidence_responses_label" as const },
      path: "/v1/responses",
      status: "offline_only",
      evidenceRef: auditRef,
      summary: "Responses compatibility is grouped as gateway distribution evidence without new business truth.",
      summaryMessage: { key: "gatewayOverviewEvidence_responses_summary" as const },
    },
    {
      surfaceId: "messages",
      label: "Messages",
      labelMessage: { key: "gatewayOverviewEvidence_messages_label" as const },
      path: "/v1/messages",
      status: "offline_only",
      evidenceRef: auditRef,
      summary: "Messages compatibility remains a northbound adapter surface, not a separate product protocol.",
      summaryMessage: { key: "gatewayOverviewEvidence_messages_summary" as const },
    },
  ];
}

function buildPolicyEvidence(source: ModelGatewayOverviewSource): ModelGatewayPolicyEvidence[] {
  const quota = source.workspaceUsageQuota.quota;
  const scopes = Array.from(new Set(source.workspaceApiKeys.apiKeys.flatMap((apiKey) => apiKey.scopes))).sort();
  const activeKeys = source.workspaceApiKeys.apiKeys.filter((apiKey) => apiKey.state === "active").length;

  return [
    {
      evidenceId: "api_key_summary",
      label: "API key summary",
      labelMessage: { key: "gatewayOverviewEvidence_api_key_summary_label" as const },
      value: `${activeKeys}/${source.workspaceApiKeys.apiKeys.length} active`,
      valueMessage: { key: "gatewayOverviewEvidence_api_key_summary_value" as const, values: { activeKeys: activeKeys, apiKeysCount: source.workspaceApiKeys.apiKeys.length } },
      status: source.workspaceApiKeys.canRenderApiKeys ? "offline_only" : "blocked",
      sourceRef: source.workspaceApiKeys.routeId,
      summary: "Only key id, owner, scope, state, and timestamps render; raw key material is forbidden.",
      summaryMessage: { key: "gatewayOverviewEvidence_api_key_summary_summary" as const },
    },
    {
      evidenceId: "scope_projection",
      label: "Scope projection",
      labelMessage: { key: "gatewayOverviewEvidence_scope_projection_label" as const },
      value: scopes.length > 0 ? scopes.join(", ") : "none",
      valueMessage: scopes.length > 0 ? undefined : { key: "gatewayOverviewEvidence_scope_projection_value_no" as const },
      status: scopes.length > 0 ? "ready" : "review_required",
      sourceRef: source.workspaceApiKeys.requestId,
      summary: "Scope rows explain intended read access without validating production bearer tokens.",
      summaryMessage: { key: "gatewayOverviewEvidence_scope_projection_summary" as const },
    },
    {
      evidenceId: "quota_policy",
      label: "Quota policy",
      labelMessage: { key: "gatewayOverviewEvidence_quota_policy_label" as const },
      value: quota
        ? `${quota.request_limit.toLocaleString("en-US")} req / ${quota.token_limit.toLocaleString("en-US")} tokens`
        : "missing",
    valueMessage: quota ? { key: "gatewayOverviewEvidence_quota_policy_value_yes" as const, values: { limit: quota.request_limit, limit2: quota.token_limit } } : { key: "gatewayOverviewEvidence_quota_policy_value_no" as const },
      status: quota ? "offline_only" : "blocked",
      sourceRef: source.workspaceUsageQuota.routeId,
      summary: "Quota limits and usage snapshot are visible, while quota enforcement and rate limit are still not implemented.",
      summaryMessage: { key: "gatewayOverviewEvidence_quota_policy_summary" as const },
    },
    {
      evidenceId: "cost_snapshot",
      label: "Cost snapshot",
      labelMessage: { key: "gatewayOverviewEvidence_cost_snapshot_label" as const },
      value: quota ? `$${quota.usage_snapshot.estimated_cost.toFixed(2)} used` : "missing",
      valueMessage: quota ? { key: "gatewayOverviewEvidence_cost_snapshot_value_yes" as const, values: { cost: { number: quota.usage_snapshot.estimated_cost, style: "money" } } } : { key: "gatewayOverviewEvidence_cost_snapshot_value_no" as const },
      status: quota ? "offline_only" : "blocked",
      sourceRef: source.workspaceUsageQuota.requestId,
      summary: "Estimated cost is a read-side snapshot, not a cost record writer or durable cost calculation.",
      summaryMessage: { key: "gatewayOverviewEvidence_cost_snapshot_summary" as const },
    },
    {
      evidenceId: "trace_records",
      label: "Trace records",
      labelMessage: { key: "gatewayOverviewEvidence_trace_records_label" as const },
      value: String(source.workspaceRunHistory.runs.length),
      status: source.workspaceRunHistory.canRenderRuns ? "ready" : "blocked",
      sourceRef: source.workspaceRunHistory.routeId,
      summary: "Run trace ids and failure codes give request evidence without result materialization or replay.",
      summaryMessage: { key: "gatewayOverviewEvidence_trace_records_summary" as const },
    },
    {
      evidenceId: "audit_summary",
      label: "Audit summary",
      labelMessage: { key: "gatewayOverviewEvidence_audit_summary_label" as const },
      value: source.adminAuditLog.auditRef,
      status: source.adminAuditLog.canRenderAuditLog ? "ready" : "blocked",
      sourceRef: source.adminAuditLog.routeId,
      summary: "Gateway evidence remains audit-linked without raw payload export, mutation, or secret reveal.",
      summaryMessage: { key: "gatewayOverviewEvidence_audit_summary_summary" as const },
    },
    {
      evidenceId: "provider_policy",
      label: "Provider policy",
      labelMessage: { key: "gatewayOverviewEvidence_provider_policy_label" as const },
      value: "caller-managed retry / fallback disabled",
      valueMessage: { key: "gatewayOverviewEvidence_provider_policy_value" as const },
      status: "locked",
      sourceRef: "provider-retry-fallback-policy-v1",
      summary: "Selection and retry/fallback policy are documented as evidence; execution remains disabled until a separate task opens it.",
      summaryMessage: { key: "gatewayOverviewEvidence_provider_policy_summary" as const },
    },
  ];
}

function buildTraceEvidence(source: ModelGatewayOverviewSource): ModelGatewayTraceEvidence[] {
  return source.workspaceRunHistory.runs.map((run) => ({
    traceId: run.traceId,
    runId: run.runId,
    status: run.failureCode === "none" ? "ready" : "review_required",
    cost: run.estimatedCost,
    failureCode: run.failureCode,
    auditRef: source.workspaceRunHistory.auditRef,
    summary: `${run.applicationRef} uses ${run.workflowDefinitionId}; status ${run.status} is read-only gateway usage evidence.`,
    summaryMessage: { key: "gatewayOverviewEvidence_buildTraceEvidence_summary" as const, values: { applicationRef: run.applicationRef, workflowDefinitionId: run.workflowDefinitionId, status: run.status } },
  }));
}

function buildBoundaryLocks(): ModelGatewayBoundaryLock[] {
  return [
    {
      lockId: "production_gateway_auth",
      label: "Production gateway auth",
      labelMessage: { key: "gatewayOverviewEvidence_production_gateway_auth_label" as const },
      status: "locked",
      missingPrerequisite: "Radish auth and production API consumer",
      missingPrerequisiteMessage: { key: "gatewayOverviewEvidence_production_gateway_auth_missingPrerequisite" as const },
      summary: "Gateway overview does not validate production tokens or replace Radish identity truth.",
      summaryMessage: { key: "gatewayOverviewEvidence_production_gateway_auth_summary" as const },
    },
    {
      lockId: "api_key_lifecycle",
      label: "API key lifecycle",
      labelMessage: { key: "gatewayOverviewEvidence_api_key_lifecycle_label" as const },
      status: "locked",
      missingPrerequisite: "key issuance, hashing, storage, and validation contract",
      missingPrerequisiteMessage: { key: "gatewayOverviewEvidence_api_key_lifecycle_missingPrerequisite" as const },
      summary: "The UI shows key summaries only; it cannot create, rotate, reveal, hash, or validate keys.",
      summaryMessage: { key: "gatewayOverviewEvidence_api_key_lifecycle_summary" as const },
    },
    {
      lockId: "quota_enforcement",
      label: "Quota enforcement",
      labelMessage: { key: "gatewayOverviewEvidence_quota_enforcement_label" as const },
      status: "locked",
      missingPrerequisite: "gateway quota policy execution and rate limit",
      missingPrerequisiteMessage: { key: "gatewayOverviewEvidence_quota_enforcement_missingPrerequisite" as const },
      summary: "Quota fields are visible as policy evidence and do not block live requests.",
      summaryMessage: { key: "gatewayOverviewEvidence_quota_enforcement_summary" as const },
    },
    {
      lockId: "cost_record",
      label: "Cost record",
      labelMessage: { key: "gatewayOverviewEvidence_cost_record_label" as const },
      status: "locked",
      missingPrerequisite: "durable cost ledger and pricing policy",
      missingPrerequisiteMessage: { key: "gatewayOverviewEvidence_cost_record_missingPrerequisite" as const },
      summary: "Estimated cost is displayed from read-side run and quota summaries only.",
      summaryMessage: { key: "gatewayOverviewEvidence_cost_record_summary" as const },
    },
    {
      lockId: "secret_resolver",
      label: "Secret resolver",
      labelMessage: { key: "gatewayOverviewEvidence_secret_resolver_label" as const },
      status: "locked",
      missingPrerequisite: "production secret backend implementation",
      missingPrerequisiteMessage: { key: "gatewayOverviewEvidence_secret_resolver_missingPrerequisite" as const },
      summary: "Provider credentials remain reference-only; no secret value is stored or rendered.",
      summaryMessage: { key: "gatewayOverviewEvidence_secret_resolver_summary" as const },
    },
    {
      lockId: "retry_fallback_execution",
      label: "Retry and fallback execution",
      labelMessage: { key: "gatewayOverviewEvidence_retry_fallback_execution_label" as const },
      status: "locked",
      missingPrerequisite: "explicit retry/fallback execution task",
      missingPrerequisiteMessage: { key: "gatewayOverviewEvidence_retry_fallback_execution_missingPrerequisite" as const },
      summary: "Current policy is caller-managed retry and disabled fallback, with no implicit provider switching.",
      summaryMessage: { key: "gatewayOverviewEvidence_retry_fallback_execution_summary" as const },
    },
    {
      lockId: "database_auth_repository",
      label: "Database, auth, repository",
      labelMessage: { key: "gatewayOverviewEvidence_database_auth_repository_label" as const },
      status: "locked",
      missingPrerequisite: "implementation trigger review satisfied",
      missingPrerequisiteMessage: { key: "gatewayOverviewEvidence_database_auth_repository_missingPrerequisite" as const },
      summary: "Read-side fake store evidence does not unlock database, Radish auth, repository adapter, or production API consumer.",
      summaryMessage: { key: "gatewayOverviewEvidence_database_auth_repository_summary" as const },
    },
  ];
}

function quotaUsageLabel(source: WorkspaceUsageQuotaViewModel): string {
  const quota = source.quota;
  if (!quota) {
    return "missing";
  }
  return `${quota.usage_snapshot.request_count.toLocaleString("en-US")} / ${quota.request_limit.toLocaleString(
    "en-US",
  )}`;
}

function buildGatewayNarrativeMessage(source: ModelGatewayOverviewSource): EvidenceProjectionMessage {
  const activeKeys = source.workspaceApiKeys.apiKeys.filter((apiKey) => apiKey.state === "active").length;
  return { key: "gatewayOverviewEvidence_gatewayNarrative" as const, values: { routesCount: source.readShell.catalog.routes.length, apiKeysCount: source.workspaceApiKeys.apiKeys.length, activeKeys: activeKeys, runsCount: source.workspaceRunHistory.runs.length } };
}
