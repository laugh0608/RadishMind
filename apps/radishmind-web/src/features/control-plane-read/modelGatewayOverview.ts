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

export type ModelGatewayOverviewMetric = {
  metricId: string;
  label: string;
  value: string;
  status: ModelGatewayOverviewStatus;
  summary: string;
};

export type ModelGatewayApiSurface = {
  surfaceId: string;
  label: string;
  path: string;
  status: ModelGatewayOverviewStatus;
  evidenceRef: string;
  summary: string;
};

export type ModelGatewayPolicyEvidence = {
  evidenceId: string;
  label: string;
  value: string;
  status: ModelGatewayOverviewStatus;
  sourceRef: string;
  summary: string;
};

export type ModelGatewayTraceEvidence = {
  traceId: string;
  runId: string;
  status: ModelGatewayOverviewStatus;
  cost: string;
  failureCode: string;
  auditRef: string;
  summary: string;
};

export type ModelGatewayBoundaryLock = {
  lockId: string;
  label: string;
  status: ModelGatewayOverviewStatus;
  missingPrerequisite: string;
  summary: string;
};

export type ModelGatewayOverviewViewModel = {
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
      value: "5",
      status: "offline_only",
      summary: "Models, chat completions, responses, messages, and model lookup stay visible as compatibility evidence.",
    },
    {
      metricId: "read_routes",
      label: "Read routes",
      value: String(source.readShell.catalog.routes.length),
      status: source.readShell.usesCanonicalRoutes ? "ready" : "blocked",
      summary: "Gateway overview reuses the shared read catalog instead of creating a second product contract.",
    },
    {
      metricId: "api_keys",
      label: "API keys",
      value: `${activeKeys}/${source.workspaceApiKeys.apiKeys.length}`,
      status: source.workspaceApiKeys.canRenderApiKeys ? "offline_only" : "blocked",
      summary: "Key summaries are displayed without key value, hash, validation, issuance, or lifecycle mutation.",
    },
    {
      metricId: "quota",
      label: "Quota",
      value: quotaUsageLabel(source.workspaceUsageQuota),
      status: source.workspaceUsageQuota.canRenderQuota ? "offline_only" : "blocked",
      summary: "Quota policy and usage snapshot are evidence only; enforcement and rate limit remain locked.",
    },
    {
      metricId: "trace_cost",
      label: "Trace and cost",
      value: `${source.workspaceRunHistory.runs.length} / ${failedRuns} failed`,
      status: failedRuns > 0 ? "review_required" : "ready",
      summary: "Run trace ids, failure codes, and estimated cost summaries are readable without replay or cost writeback.",
    },
    {
      metricId: "audit_events",
      label: "Audit events",
      value: String(source.adminAuditLog.auditEvents.length),
      status: source.adminAuditLog.canRenderAuditLog ? "ready" : "blocked",
      summary: "Audit summaries anchor read access and denied decisions without raw payload export.",
    },
    {
      metricId: "locked_capabilities",
      label: "Locked capabilities",
      value: String(buildBoundaryLocks().length),
      status: "locked",
      summary: "Production auth, API key issuance, quota enforcement, cost record writes, secret resolver, and fallback execution stay disabled.",
    },
  ];
}

function buildApiSurfaces(source: ModelGatewayOverviewSource): ModelGatewayApiSurface[] {
  const auditRef = source.adminAuditLog.auditRef;
  return [
    {
      surfaceId: "models_inventory",
      label: "Models inventory",
      path: "/v1/models",
      status: "offline_only",
      evidenceRef: "provider-runtime-docs-refresh",
      summary: "Provider-qualified profile inventory remains a read/discovery surface, not a provisioning API.",
    },
    {
      surfaceId: "model_lookup",
      label: "Model lookup",
      path: "/v1/models/{id}",
      status: "offline_only",
      evidenceRef: "provider-selection-policy-v1",
      summary: "Lookup evidence keeps provider/model selection explicit and does not enable implicit fallback.",
    },
    {
      surfaceId: "chat_completions",
      label: "Chat completions",
      path: "/v1/chat/completions",
      status: "offline_only",
      evidenceRef: auditRef,
      summary: "Compatibility route evidence stays tied to canonical gateway envelope and request metadata.",
    },
    {
      surfaceId: "responses",
      label: "Responses",
      path: "/v1/responses",
      status: "offline_only",
      evidenceRef: auditRef,
      summary: "Responses compatibility is grouped as gateway distribution evidence without new business truth.",
    },
    {
      surfaceId: "messages",
      label: "Messages",
      path: "/v1/messages",
      status: "offline_only",
      evidenceRef: auditRef,
      summary: "Messages compatibility remains a northbound adapter surface, not a separate product protocol.",
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
      value: `${activeKeys}/${source.workspaceApiKeys.apiKeys.length} active`,
      status: source.workspaceApiKeys.canRenderApiKeys ? "offline_only" : "blocked",
      sourceRef: source.workspaceApiKeys.routeId,
      summary: "Only key id, owner, scope, state, and timestamps render; raw key material is forbidden.",
    },
    {
      evidenceId: "scope_projection",
      label: "Scope projection",
      value: scopes.length > 0 ? scopes.join(", ") : "none",
      status: scopes.length > 0 ? "ready" : "review_required",
      sourceRef: source.workspaceApiKeys.requestId,
      summary: "Scope rows explain intended read access without validating production bearer tokens.",
    },
    {
      evidenceId: "quota_policy",
      label: "Quota policy",
      value: quota
        ? `${quota.request_limit.toLocaleString("en-US")} req / ${quota.token_limit.toLocaleString("en-US")} tokens`
        : "missing",
      status: quota ? "offline_only" : "blocked",
      sourceRef: source.workspaceUsageQuota.routeId,
      summary: "Quota limits and usage snapshot are visible, while quota enforcement and rate limit are still not implemented.",
    },
    {
      evidenceId: "cost_snapshot",
      label: "Cost snapshot",
      value: quota ? `$${quota.usage_snapshot.estimated_cost.toFixed(2)} used` : "missing",
      status: quota ? "offline_only" : "blocked",
      sourceRef: source.workspaceUsageQuota.requestId,
      summary: "Estimated cost is a read-side snapshot, not a cost record writer or durable cost calculation.",
    },
    {
      evidenceId: "trace_records",
      label: "Trace records",
      value: String(source.workspaceRunHistory.runs.length),
      status: source.workspaceRunHistory.canRenderRuns ? "ready" : "blocked",
      sourceRef: source.workspaceRunHistory.routeId,
      summary: "Run trace ids and failure codes give request evidence without result materialization or replay.",
    },
    {
      evidenceId: "audit_summary",
      label: "Audit summary",
      value: source.adminAuditLog.auditRef,
      status: source.adminAuditLog.canRenderAuditLog ? "ready" : "blocked",
      sourceRef: source.adminAuditLog.routeId,
      summary: "Gateway evidence remains audit-linked without raw payload export, mutation, or secret reveal.",
    },
    {
      evidenceId: "provider_policy",
      label: "Provider policy",
      value: "caller-managed retry / fallback disabled",
      status: "locked",
      sourceRef: "provider-retry-fallback-policy-v1",
      summary: "Selection and retry/fallback policy are documented as evidence; execution remains disabled until a separate task opens it.",
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
  }));
}

function buildBoundaryLocks(): ModelGatewayBoundaryLock[] {
  return [
    {
      lockId: "production_gateway_auth",
      label: "Production gateway auth",
      status: "locked",
      missingPrerequisite: "Radish auth and production API consumer",
      summary: "Gateway overview does not validate production tokens or replace Radish identity truth.",
    },
    {
      lockId: "api_key_lifecycle",
      label: "API key lifecycle",
      status: "locked",
      missingPrerequisite: "key issuance, hashing, storage, and validation contract",
      summary: "The UI shows key summaries only; it cannot create, rotate, reveal, hash, or validate keys.",
    },
    {
      lockId: "quota_enforcement",
      label: "Quota enforcement",
      status: "locked",
      missingPrerequisite: "gateway quota policy execution and rate limit",
      summary: "Quota fields are visible as policy evidence and do not block live requests.",
    },
    {
      lockId: "cost_record",
      label: "Cost record",
      status: "locked",
      missingPrerequisite: "durable cost ledger and pricing policy",
      summary: "Estimated cost is displayed from read-side run and quota summaries only.",
    },
    {
      lockId: "secret_resolver",
      label: "Secret resolver",
      status: "locked",
      missingPrerequisite: "production secret backend implementation",
      summary: "Provider credentials remain reference-only; no secret value is stored or rendered.",
    },
    {
      lockId: "retry_fallback_execution",
      label: "Retry and fallback execution",
      status: "locked",
      missingPrerequisite: "explicit retry/fallback execution task",
      summary: "Current policy is caller-managed retry and disabled fallback, with no implicit provider switching.",
    },
    {
      lockId: "database_auth_repository",
      label: "Database, auth, repository",
      status: "locked",
      missingPrerequisite: "implementation trigger review satisfied",
      summary: "Read-side fake store evidence does not unlock database, Radish auth, repository adapter, or production API consumer.",
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
