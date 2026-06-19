import type { AdminAuditLogViewModel } from "./adminAuditLog";
import type { AdminTenantOverviewViewModel } from "./adminTenantOverview";
import type {
  ModelGatewayEvidenceReviewStatus,
  ModelGatewayEvidenceReviewViewModel,
} from "./modelGatewayEvidenceReview";
import type { ControlPlaneReadShellViewModel } from "./readShell";

export type AdminOperationsReviewStatus = ModelGatewayEvidenceReviewStatus;

export type AdminOperationsReviewSource = {
  readShell: ControlPlaneReadShellViewModel;
  tenantOverview: AdminTenantOverviewViewModel;
  adminAuditLog: AdminAuditLogViewModel;
  modelGatewayEvidenceReview: ModelGatewayEvidenceReviewViewModel;
};

export type AdminOperationsReadinessRollup = {
  rollupId: string;
  label: string;
  value: string;
  status: AdminOperationsReviewStatus;
  sourceRef: string;
  summary: string;
};

export type AdminOperationsEvidenceChecklistItem = {
  checklistId: string;
  label: string;
  sourceSurface: "tenant" | "audit" | "gateway" | "production_ops" | "read_contract";
  routeOrPageId: string;
  requestId: string;
  auditRef: string;
  status: AdminOperationsReviewStatus;
  summary: string;
};

export type AdminOperationsRisk = {
  riskId: string;
  label: string;
  sourceSurface: "audit" | "gateway" | "deployment" | "secret" | "auth_store";
  status: "review_required" | "blocked" | "locked";
  evidenceRef: string;
  summary: string;
  operatorQuestion: string;
};

export type AdminOperationsBoundaryLock = {
  lockId: string;
  label: string;
  sourceSurface:
    | "tenant"
    | "audit"
    | "auth"
    | "database"
    | "repository"
    | "secret"
    | "gateway"
    | "deployment"
    | "workflow";
  status: "locked";
  missingPrerequisite: string;
  summary: string;
};

export type AdminOperationsReviewViewModel = {
  pageId: "admin-operations-review-offline";
  sourcePageIds: string[];
  reviewMode: "offline_admin_operations_readiness";
  tenantRef: string;
  requestId: string;
  auditRef: string;
  reviewNarrative: string;
  readinessRollup: AdminOperationsReadinessRollup[];
  evidenceChecklist: AdminOperationsEvidenceChecklistItem[];
  operationalRisks: AdminOperationsRisk[];
  boundaryLocks: AdminOperationsBoundaryLock[];
  canRenderAdminOperationsReview: boolean;
  canInspectAdminEvidenceLocally: true;
  canRequestProductionBackend: false;
  canMutateTenant: false;
  canExportAuditPayload: false;
  canEnableRadishAuth: false;
  canAttachDatabase: false;
  canImplementRepositoryAdapter: false;
  canResolveProductionSecret: false;
  canIssueApiKey: false;
  canEnforceQuota: false;
  canWriteCostRecord: false;
  canRunDeploymentPreflight: false;
  canStartWorkflowRuntime: false;
  canWriteBusinessTruth: false;
  canReplayRun: false;
};

export function buildAdminOperationsReviewViewModel(
  source: AdminOperationsReviewSource,
): AdminOperationsReviewViewModel {
  const readinessRollup = buildReadinessRollup(source);
  const evidenceChecklist = buildEvidenceChecklist(source);
  const operationalRisks = buildOperationalRisks(source);
  const boundaryLocks = buildBoundaryLocks(source);

  return {
    pageId: "admin-operations-review-offline",
    sourcePageIds: [
      source.tenantOverview.pageId,
      source.adminAuditLog.pageId,
      source.modelGatewayEvidenceReview.pageId,
      "control-plane-read-formal-ui-readiness-close-v1",
      "control-plane-read-auth-store-transition-preconditions-v1",
      "control-plane-read-implementation-trigger-review-v1",
      "production-ops-config-secret-boundary",
      "production-secret-backend-contract",
      "production-secret-backend-implementation-readiness",
      "production-secret-reference-contract",
      "production-ops-startup-supervisor-boundary",
      "production-ops-environment-isolation-boundary",
      "production-ops-deployment-readiness-smoke",
      "production-ops-container-smoke-record-template",
    ],
    reviewMode: "offline_admin_operations_readiness",
    tenantRef: source.tenantOverview.collection.tenantRef,
    requestId: source.tenantOverview.requestId,
    auditRef: source.adminAuditLog.auditRef,
    reviewNarrative: buildReviewNarrative(source, operationalRisks, boundaryLocks),
    readinessRollup,
    evidenceChecklist,
    operationalRisks,
    boundaryLocks,
    canRenderAdminOperationsReview:
      source.readShell.usesCanonicalRoutes &&
      source.tenantOverview.canRenderTenant &&
      source.adminAuditLog.canRenderAuditLog &&
      source.modelGatewayEvidenceReview.canRenderEvidenceReview &&
      readinessRollup.length >= 8 &&
      evidenceChecklist.length >= 14 &&
      operationalRisks.length >= 8 &&
      boundaryLocks.length >= 12,
    canInspectAdminEvidenceLocally: true,
    canRequestProductionBackend: false,
    canMutateTenant: false,
    canExportAuditPayload: false,
    canEnableRadishAuth: false,
    canAttachDatabase: false,
    canImplementRepositoryAdapter: false,
    canResolveProductionSecret: false,
    canIssueApiKey: false,
    canEnforceQuota: false,
    canWriteCostRecord: false,
    canRunDeploymentPreflight: false,
    canStartWorkflowRuntime: false,
    canWriteBusinessTruth: false,
    canReplayRun: false,
  };
}

function buildReviewNarrative(
  source: AdminOperationsReviewSource,
  risks: AdminOperationsRisk[],
  locks: AdminOperationsBoundaryLock[],
): string {
  const deniedAuditEvents = source.adminAuditLog.auditEvents.filter((event) => event.decision === "denied").length;
  return `${source.tenantOverview.pageId}, ${source.adminAuditLog.pageId}, and ${source.modelGatewayEvidenceReview.pageId} are grouped into one admin operations review with ${source.readShell.catalog.routes.length} read routes, ${deniedAuditEvents} denied audit event, ${risks.length} operational risks, and ${locks.length} locked capabilities.`;
}

function buildReadinessRollup(source: AdminOperationsReviewSource): AdminOperationsReadinessRollup[] {
  const deniedAuditEvents = source.adminAuditLog.auditEvents.filter((event) => event.decision === "denied").length;
  const gatewayRiskCount =
    source.modelGatewayEvidenceReview.routeRisks.length +
    source.modelGatewayEvidenceReview.usageRisks.length +
    source.modelGatewayEvidenceReview.auditRisks.length;

  return [
    {
      rollupId: "tenant_overview",
      label: "Tenant overview",
      value: source.tenantOverview.canRenderTenant ? "rendered" : "blocked",
      status: source.tenantOverview.canRenderTenant ? "offline_only" : "blocked",
      sourceRef: source.tenantOverview.routeId,
      summary: "Tenant, plan, quota summary, deployment status, and audit summary refs are readable without tenant mutation.",
    },
    {
      rollupId: "audit_log",
      label: "Audit log",
      value: `${source.adminAuditLog.auditEvents.length} events`,
      status: deniedAuditEvents > 0 ? "review_required" : "ready",
      sourceRef: source.adminAuditLog.routeId,
      summary: "Audit summaries expose allowed and denied decisions without raw payload export, edit, delete, or secret reveal.",
    },
    {
      rollupId: "read_contract",
      label: "Read contract",
      value: `${source.readShell.catalog.routes.length} routes`,
      status: source.readShell.usesCanonicalRoutes ? "ready" : "blocked",
      sourceRef: "control-plane-read-formal-ui-readiness-close-v1",
      summary: "Admin review stays on the existing read-side route catalog and does not add a second product API.",
    },
    {
      rollupId: "forbidden_projection_guard",
      label: "Forbidden projection guard",
      value: source.readShell.forbiddenProjectionBlocked ? "active" : "missing",
      status: source.readShell.forbiddenProjectionBlocked ? "locked" : "blocked",
      sourceRef: "control-plane-read-negative-contract-v1",
      summary: "Sensitive keys remain blocked before tenant, audit, gateway, or production evidence can render them.",
    },
    {
      rollupId: "gateway_review",
      label: "Gateway review",
      value: source.modelGatewayEvidenceReview.canRenderEvidenceReview ? "rendered" : "blocked",
      status: source.modelGatewayEvidenceReview.canRenderEvidenceReview ? "offline_only" : "blocked",
      sourceRef: source.modelGatewayEvidenceReview.pageId,
      summary: "Gateway evidence supplies provider/profile, route, usage, audit, readiness, and locked capability context.",
    },
    {
      rollupId: "gateway_risks",
      label: "Gateway risks",
      value: String(gatewayRiskCount),
      status: gatewayRiskCount > 0 ? "review_required" : "ready",
      sourceRef: source.modelGatewayEvidenceReview.pageId,
      summary: "Route, usage, and audit risks remain review evidence and do not unlock production gateway execution.",
    },
    {
      rollupId: "production_ops_static",
      label: "Production ops evidence",
      value: "static boundary",
      status: "offline_only",
      sourceRef: "production-ops-deployment-readiness-smoke",
      summary: "Production ops evidence is limited to static boundary, compose expansion, runbook, and record template refs.",
    },
    {
      rollupId: "implementation_triggers",
      label: "Implementation triggers",
      value: "0 satisfied",
      status: "locked",
      sourceRef: "control-plane-read-implementation-trigger-review-v1",
      summary: "Schema artifact, store selector, production auth, and adapter smoke triggers remain unsatisfied.",
    },
    {
      rollupId: "locked_capabilities",
      label: "Locked capabilities",
      value: String(buildBoundaryLocks(source).length),
      status: "locked",
      sourceRef: "admin_operations_review",
      summary: "Tenant mutation, raw audit export, Radish auth, database, repository, secret resolver, production gateway, and runtime execution remain closed.",
    },
  ];
}

function buildEvidenceChecklist(source: AdminOperationsReviewSource): AdminOperationsEvidenceChecklistItem[] {
  const refs = {
    requestId: source.tenantOverview.requestId,
    auditRef: source.adminAuditLog.auditRef,
  };

  return [
    checklistItem("tenant_summary_route", "Tenant summary route", "tenant", source.tenantOverview.routeId, refs, "offline_only", "Tenant summary route anchors the admin review without creating tenant mutation paths."),
    checklistItem("tenant_state_preview", "Tenant states", "tenant", source.tenantOverview.pageId, refs, "offline_only", "Ready, empty, denied, stale, and forbidden projection states remain visible as read-side evidence."),
    checklistItem("audit_summary_route", "Audit summary route", "audit", source.adminAuditLog.routeId, refs, "offline_only", "Audit rows keep actor, decision, failure code, trace, and cursor evidence sanitized."),
    checklistItem("audit_denial_review", "Audit denial review", "audit", source.adminAuditLog.pageId, refs, "review_required", "Denied audit events are surfaced for governance review without payload export or mutation."),
    checklistItem("read_route_catalog", "Read route catalog", "read_contract", "control-plane-read-shared-shell-v1", refs, "ready", "The admin review consumes existing canonical read route metadata."),
    checklistItem("forbidden_output_guard", "Forbidden output guard", "read_contract", "control-plane-read-negative-contract-v1", refs, "locked", "Forbidden output keys stay blocked across tenant, audit, and operations summaries."),
    checklistItem("gateway_evidence_review", "Gateway evidence review", "gateway", source.modelGatewayEvidenceReview.pageId, refs, "offline_only", "Gateway review contributes provider/profile, route, usage, audit, risk, and lock evidence."),
    checklistItem("gateway_key_quota", "Gateway key and quota", "gateway", "gateway-api-key-quota-readiness", refs, "locked", "API key lifecycle, quota execution, rate limit, and cost ledger are readable only as readiness evidence."),
    checklistItem("production_secret_boundary", "Secret boundary", "production_ops", "production-ops-config-secret-boundary", refs, "locked", "Config and secret boundary evidence keeps secret values out of committed and rendered output."),
    checklistItem("secret_backend_contract", "Secret backend contract", "production_ops", "production-secret-backend-contract", refs, "locked", "External secret backend contract exists, but implementation and resolver are still absent."),
    checklistItem("secret_reference_schema", "Secret ref schema", "production_ops", "production-secret-reference-contract", refs, "locked", "Secret refs are schema evidence only and do not resolve production credentials."),
    checklistItem("startup_supervisor", "Startup supervisor", "production_ops", "production-ops-startup-supervisor-boundary", refs, "locked", "Supervisor boundary is documented without process manager implementation."),
    checklistItem("environment_isolation", "Environment isolation", "production_ops", "production-ops-environment-isolation-boundary", refs, "locked", "Environment isolation remains static boundary evidence, not production deployment readiness."),
    checklistItem("deployment_static_smoke", "Deployment static smoke", "production_ops", "production-ops-deployment-readiness-smoke", refs, "offline_only", "Deployment readiness only covers static compose expansion and does not run test or production smoke."),
    checklistItem("container_smoke_record", "Container smoke record", "production_ops", "production-ops-container-smoke-record-template", refs, "offline_only", "Container smoke record template and local mock run are audit evidence, not a production preflight."),
    checklistItem("auth_store_transition", "Auth and store transition", "read_contract", "control-plane-read-auth-store-transition-preconditions-v1", refs, "locked", "Auth middleware and read store repository transition conditions remain unsatisfied."),
    checklistItem("implementation_trigger_review", "Implementation trigger review", "read_contract", "control-plane-read-implementation-trigger-review-v1", refs, "locked", "Current implementation trigger review keeps repository, selector, auth, and adapter work closed."),
  ];
}

function checklistItem(
  checklistId: string,
  label: string,
  sourceSurface: AdminOperationsEvidenceChecklistItem["sourceSurface"],
  routeOrPageId: string,
  refs: Pick<AdminOperationsEvidenceChecklistItem, "requestId" | "auditRef">,
  status: AdminOperationsReviewStatus,
  summary: string,
): AdminOperationsEvidenceChecklistItem {
  return {
    checklistId,
    label,
    sourceSurface,
    routeOrPageId,
    requestId: refs.requestId,
    auditRef: refs.auditRef,
    status,
    summary,
  };
}

function buildOperationalRisks(source: AdminOperationsReviewSource): AdminOperationsRisk[] {
  const deniedAuditRisks = source.adminAuditLog.auditEvents
    .filter((event) => event.decision === "denied")
    .map((event) => ({
      riskId: `audit_${event.auditRef}`,
      label: event.eventKind,
      sourceSurface: "audit" as const,
      status: "review_required" as const,
      evidenceRef: event.auditRef,
      summary: `${event.actorSubjectRef} was denied against ${event.resourceRef} with ${event.failureCode}.`,
      operatorQuestion: "Which scope or tenant binding should be reviewed before production auth is introduced?",
    }));
  const gatewayRisks: AdminOperationsRisk[] = [
    {
      riskId: "gateway_route_risks",
      label: "Gateway route risks",
      sourceSurface: "gateway",
      status: source.modelGatewayEvidenceReview.routeRisks.length > 0 ? "review_required" : "locked",
      evidenceRef: source.modelGatewayEvidenceReview.pageId,
      summary: `${source.modelGatewayEvidenceReview.routeRisks.length} route risks remain visible for provider/profile review.`,
      operatorQuestion: "Which provider/profile mismatch or selection case should block execution first?",
    },
    {
      riskId: "gateway_usage_risks",
      label: "Gateway usage risks",
      sourceSurface: "gateway",
      status: source.modelGatewayEvidenceReview.usageRisks.length > 0 ? "review_required" : "locked",
      evidenceRef: source.modelGatewayEvidenceReview.pageId,
      summary: `${source.modelGatewayEvidenceReview.usageRisks.length} usage risks remain visible without enforcement or cost writes.`,
      operatorQuestion: "Which usage, quota, or cost prerequisite is absent before production traffic?",
    },
    {
      riskId: "gateway_audit_risks",
      label: "Gateway audit risks",
      sourceSurface: "gateway",
      status: source.modelGatewayEvidenceReview.auditRisks.length > 0 ? "review_required" : "locked",
      evidenceRef: source.modelGatewayEvidenceReview.auditRef,
      summary: `${source.modelGatewayEvidenceReview.auditRisks.length} gateway audit risks remain governance evidence only.`,
      operatorQuestion: "Which audit decision or trace should be reviewed before any production route is enabled?",
    },
  ];
  const operationsRisks: AdminOperationsRisk[] = [
    {
      riskId: "production_secret_backend_missing",
      label: "Production secret backend",
      sourceSurface: "secret",
      status: "locked",
      evidenceRef: "production-secret-backend-implementation-readiness",
      summary: "Secret backend implementation readiness exists, but resolver injection and backend implementation are not present.",
      operatorQuestion: "Which backend, profile binding, rotation, and audit policy must be selected before resolver work starts?",
    },
    {
      riskId: "process_supervisor_missing",
      label: "Process supervisor",
      sourceSurface: "deployment",
      status: "locked",
      evidenceRef: "production-ops-startup-supervisor-boundary",
      summary: "Startup and supervisor boundary is documented, while a real process supervisor is still absent.",
      operatorQuestion: "Which supervisor and restart policy should be chosen before production packaging?",
    },
    {
      riskId: "environment_isolation_static",
      label: "Environment isolation",
      sourceSurface: "deployment",
      status: "locked",
      evidenceRef: "production-ops-environment-isolation-boundary",
      summary: "Environment isolation is static evidence and does not prove test or production environment separation.",
      operatorQuestion: "Which test and production smoke records are needed before deployment readiness can advance?",
    },
    {
      riskId: "implementation_trigger_unsatisfied",
      label: "Implementation trigger",
      sourceSurface: "auth_store",
      status: "locked",
      evidenceRef: "control-plane-read-implementation-trigger-review-v1",
      summary: "Schema artifact, store selector, production auth, and adapter smoke triggers are all not satisfied.",
      operatorQuestion: "Which trigger evidence should be completed before any repository or auth implementation begins?",
    },
    {
      riskId: "radish_auth_not_ready",
      label: "Radish auth",
      sourceSurface: "auth_store",
      status: "locked",
      evidenceRef: "control-plane-read-production-auth-readiness-v1",
      summary: "Production auth readiness defines future issuer, claim, tenant, and scope requirements without token validation.",
      operatorQuestion: "Which issuer discovery and claim mapping evidence is required before auth middleware work?",
    },
  ];

  return [...deniedAuditRisks, ...gatewayRisks, ...operationsRisks].slice(0, 12);
}

function buildBoundaryLocks(source: AdminOperationsReviewSource): AdminOperationsBoundaryLock[] {
  return [
    {
      lockId: "tenant_mutation",
      label: "Tenant mutation",
      sourceSurface: "tenant",
      status: "locked",
      missingPrerequisite: "tenant write model, mutation API, auth policy, and audit write contract",
      summary: "Tenant overview is a read-side summary and cannot create, update, archive, or migrate a tenant.",
    },
    {
      lockId: "raw_audit_export",
      label: "Raw audit export",
      sourceSurface: "audit",
      status: "locked",
      missingPrerequisite: "audit export policy and sensitive projection review",
      summary: "Audit rows remain sanitized summaries and cannot reveal payloads, edit records, or delete records.",
    },
    {
      lockId: "radish_auth",
      label: "Radish auth",
      sourceSurface: "auth",
      status: "locked",
      missingPrerequisite: "Radish auth issuer discovery, token validation, claim mapping, and auth middleware",
      summary: "Admin review does not validate production tokens or replace Radish identity truth.",
    },
    {
      lockId: "database_attachment",
      label: "Database attachment",
      sourceSurface: "database",
      status: "locked",
      missingPrerequisite: "schema artifact manifest, migrations, read-only role smoke, and durable adapter smoke",
      summary: "Read-side data remains fixture/fake-store backed until database implementation triggers are satisfied.",
    },
    {
      lockId: "repository_adapter",
      label: "Repository adapter",
      sourceSurface: "repository",
      status: "locked",
      missingPrerequisite: "repository interface, adapter implementation, store selector, and adapter smoke evidence",
      summary: "Repository adapter work stays blocked by implementation trigger review.",
    },
    {
      lockId: "production_secret_resolver",
      label: "Secret resolver",
      sourceSurface: "secret",
      status: "locked",
      missingPrerequisite: "production secret backend implementation and resolver injection policy",
      summary: "Secret references are allowed as schema evidence, while secret values and resolver execution stay unavailable.",
    },
    {
      lockId: "api_key_lifecycle",
      label: "API key lifecycle",
      sourceSurface: "gateway",
      status: "locked",
      missingPrerequisite: "key issuance, hashing, storage, validation, rotation, and revocation contract",
      summary: "Admin review can inspect key summaries through gateway evidence but cannot create, reveal, hash, or validate keys.",
    },
    {
      lockId: "quota_enforcement",
      label: "Quota enforcement",
      sourceSurface: "gateway",
      status: "locked",
      missingPrerequisite: "gateway quota policy execution, rate limit, and billing policy",
      summary: "Quota is rendered as policy evidence and does not throttle traffic or reject requests.",
    },
    {
      lockId: "cost_record_write",
      label: "Cost record write",
      sourceSurface: "gateway",
      status: "locked",
      missingPrerequisite: "durable pricing policy and cost ledger",
      summary: "Estimated cost remains read-side evidence and cannot create or update cost records.",
    },
    {
      lockId: "production_gateway",
      label: "Production gateway",
      sourceSurface: "gateway",
      status: "locked",
      missingPrerequisite: "production API consumer, auth, secret resolver, quota enforcement, and provider execution policy",
      summary: "Gateway evidence remains offline and cannot route production requests.",
    },
    {
      lockId: "deployment_preflight",
      label: "Deployment preflight",
      sourceSurface: "deployment",
      status: "locked",
      missingPrerequisite: "explicit test or production preflight window and smoke record",
      summary: "Deployment evidence is static unless an explicit run window creates a new smoke record.",
    },
    {
      lockId: "workflow_runtime_execution",
      label: "Workflow runtime",
      sourceSurface: "workflow",
      status: "locked",
      missingPrerequisite: "runtime runner, durable run/result store, confirmation decision store, writeback policy, and replay policy",
      summary: "Workflow runtime, confirmation, writeback, replay, and business truth mutation remain unavailable.",
    },
    ...source.modelGatewayEvidenceReview.lockedCapabilities.slice(0, 4).map((capability) => ({
      lockId: `gateway_${capability.capabilityId}`,
      label: capability.label,
      sourceSurface: "gateway" as const,
      status: "locked" as const,
      missingPrerequisite: capability.missingPrerequisite,
      summary: capability.summary,
    })),
  ];
}
