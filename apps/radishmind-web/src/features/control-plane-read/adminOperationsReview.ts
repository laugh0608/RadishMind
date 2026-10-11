import type { EvidenceProjectionCopy, EvidenceProjectionMessage } from "./evidenceProjectionCopy.ts";
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

export type AdminOperationsReadinessRollup = EvidenceProjectionCopy & {
  rollupId: string;
  label: string;
  value: string;
  status: AdminOperationsReviewStatus;
  sourceRef: string;
  summary: string;
};

export type AdminOperationsEvidenceChecklistItem = EvidenceProjectionCopy & {
  checklistId: string;
  label: string;
  sourceSurface: "tenant" | "audit" | "gateway" | "production_ops" | "read_contract";
  routeOrPageId: string;
  requestId: string;
  auditRef: string;
  status: AdminOperationsReviewStatus;
  summary: string;
};

export type AdminOperationsRisk = EvidenceProjectionCopy & {
  riskId: string;
  label: string;
  sourceSurface: "audit" | "gateway" | "deployment" | "secret" | "auth_store";
  status: "review_required" | "blocked" | "locked";
  evidenceRef: string;
  summary: string;
  operatorQuestion: string;
};

export type AdminOperationsBoundaryLock = EvidenceProjectionCopy & {
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

export type AdminOperationsReviewViewModel = EvidenceProjectionCopy & {
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
    reviewNarrativeMessage: buildReviewNarrativeMessage(source, operationalRisks, boundaryLocks),
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
      labelMessage: { key: "adminOperationsEvidence_tenant_overview_label" as const },
      value: source.tenantOverview.canRenderTenant ? "rendered" : "blocked",
      valueMessage: source.tenantOverview.canRenderTenant ? { key: "adminOperationsEvidence_tenant_overview_value_yes" as const } : { key: "adminOperationsEvidence_tenant_overview_value_no" as const },
      status: source.tenantOverview.canRenderTenant ? "offline_only" : "blocked",
      sourceRef: source.tenantOverview.routeId,
      summary: "Tenant, plan, quota summary, deployment status, and audit summary refs are readable without tenant mutation.",
      summaryMessage: { key: "adminOperationsEvidence_tenant_overview_summary" as const },
    },
    {
      rollupId: "audit_log",
      label: "Audit log",
      labelMessage: { key: "adminOperationsEvidence_audit_log_label" as const },
      value: `${source.adminAuditLog.auditEvents.length} events`,
      valueMessage: { key: "adminOperationsEvidence_audit_log_value" as const, values: { auditEventsCount: source.adminAuditLog.auditEvents.length } },
      status: deniedAuditEvents > 0 ? "review_required" : "ready",
      sourceRef: source.adminAuditLog.routeId,
      summary: "Audit summaries expose allowed and denied decisions without raw payload export, edit, delete, or secret reveal.",
      summaryMessage: { key: "adminOperationsEvidence_audit_log_summary" as const },
    },
    {
      rollupId: "read_contract",
      label: "Read contract",
      labelMessage: { key: "adminOperationsEvidence_read_contract_label" as const },
      value: `${source.readShell.catalog.routes.length} routes`,
      valueMessage: { key: "adminOperationsEvidence_read_contract_value" as const, values: { routesCount: source.readShell.catalog.routes.length } },
      status: source.readShell.usesCanonicalRoutes ? "ready" : "blocked",
      sourceRef: "control-plane-read-formal-ui-readiness-close-v1",
      summary: "Admin review stays on the existing read-side route catalog and does not add a second product API.",
      summaryMessage: { key: "adminOperationsEvidence_read_contract_summary" as const },
    },
    {
      rollupId: "forbidden_projection_guard",
      label: "Forbidden projection guard",
      labelMessage: { key: "adminOperationsEvidence_forbidden_projection_guard_label" as const },
      value: source.readShell.forbiddenProjectionBlocked ? "active" : "missing",
      valueMessage: source.readShell.forbiddenProjectionBlocked ? { key: "adminOperationsEvidence_forbidden_projection_guard_value_yes" as const } : { key: "adminOperationsEvidence_forbidden_projection_guard_value_no" as const },
      status: source.readShell.forbiddenProjectionBlocked ? "locked" : "blocked",
      sourceRef: "control-plane-read-negative-contract-v1",
      summary: "Sensitive keys remain blocked before tenant, audit, gateway, or production evidence can render them.",
      summaryMessage: { key: "adminOperationsEvidence_forbidden_projection_guard_summary" as const },
    },
    {
      rollupId: "gateway_review",
      label: "Gateway review",
      labelMessage: { key: "adminOperationsEvidence_gateway_review_label" as const },
      value: source.modelGatewayEvidenceReview.canRenderEvidenceReview ? "rendered" : "blocked",
      valueMessage: source.modelGatewayEvidenceReview.canRenderEvidenceReview ? { key: "adminOperationsEvidence_gateway_review_value_yes" as const } : { key: "adminOperationsEvidence_gateway_review_value_no" as const },
      status: source.modelGatewayEvidenceReview.canRenderEvidenceReview ? "offline_only" : "blocked",
      sourceRef: source.modelGatewayEvidenceReview.pageId,
      summary: "Gateway evidence supplies provider/profile, route, usage, audit, readiness, and locked capability context.",
      summaryMessage: { key: "adminOperationsEvidence_gateway_review_summary" as const },
    },
    {
      rollupId: "gateway_risks",
      label: "Gateway risks",
      labelMessage: { key: "adminOperationsEvidence_gateway_risks_label" as const },
      value: String(gatewayRiskCount),
      status: gatewayRiskCount > 0 ? "review_required" : "ready",
      sourceRef: source.modelGatewayEvidenceReview.pageId,
      summary: "Route, usage, and audit risks remain review evidence and do not unlock production gateway execution.",
      summaryMessage: { key: "adminOperationsEvidence_gateway_risks_summary" as const },
    },
    {
      rollupId: "production_ops_static",
      label: "Production ops evidence",
      labelMessage: { key: "adminOperationsEvidence_production_ops_static_label" as const },
      value: "static boundary",
      valueMessage: { key: "adminOperationsEvidence_production_ops_static_value" as const },
      status: "offline_only",
      sourceRef: "production-ops-deployment-readiness-smoke",
      summary: "Production ops evidence is limited to static boundary, compose expansion, runbook, and record template refs.",
      summaryMessage: { key: "adminOperationsEvidence_production_ops_static_summary" as const },
    },
    {
      rollupId: "implementation_triggers",
      label: "Implementation triggers",
      labelMessage: { key: "adminOperationsEvidence_implementation_triggers_label" as const },
      value: "0 satisfied",
      valueMessage: { key: "adminOperationsEvidence_implementation_triggers_value" as const },
      status: "locked",
      sourceRef: "control-plane-read-implementation-trigger-review-v1",
      summary: "Schema artifact, store selector, production auth, and adapter smoke triggers remain unsatisfied.",
      summaryMessage: { key: "adminOperationsEvidence_implementation_triggers_summary" as const },
    },
    {
      rollupId: "locked_capabilities",
      label: "Locked capabilities",
      labelMessage: { key: "adminOperationsEvidence_locked_capabilities_label" as const },
      value: String(buildBoundaryLocks(source).length),
      status: "locked",
      sourceRef: "admin_operations_review",
      summary: "Tenant mutation, raw audit export, Radish auth, database, repository, secret resolver, production gateway, and runtime execution remain closed.",
      summaryMessage: { key: "adminOperationsEvidence_locked_capabilities_summary" as const },
    },
  ];
}

function buildEvidenceChecklist(source: AdminOperationsReviewSource): AdminOperationsEvidenceChecklistItem[] {
  const refs = {
    requestId: source.tenantOverview.requestId,
    auditRef: source.adminAuditLog.auditRef,
  };

  return [
    checklistItem("tenant_summary_route", "Tenant summary route", "tenant", source.tenantOverview.routeId, refs, "offline_only", "Tenant summary route anchors the admin review without creating tenant mutation paths.", { labelMessage: { key: "adminOperationsEvidence_checklist_tenant_summary_route_label" as const }, summaryMessage: { key: "adminOperationsEvidence_checklist_tenant_summary_route_summary" as const } }),
    checklistItem("tenant_state_preview", "Tenant states", "tenant", source.tenantOverview.pageId, refs, "offline_only", "Ready, empty, denied, stale, and forbidden projection states remain visible as read-side evidence.", { labelMessage: { key: "adminOperationsEvidence_checklist_tenant_state_preview_label" as const }, summaryMessage: { key: "adminOperationsEvidence_checklist_tenant_state_preview_summary" as const } }),
    checklistItem("audit_summary_route", "Audit summary route", "audit", source.adminAuditLog.routeId, refs, "offline_only", "Audit rows keep actor, decision, failure code, trace, and cursor evidence sanitized.", { labelMessage: { key: "adminOperationsEvidence_checklist_audit_summary_route_label" as const }, summaryMessage: { key: "adminOperationsEvidence_checklist_audit_summary_route_summary" as const } }),
    checklistItem("audit_denial_review", "Audit denial review", "audit", source.adminAuditLog.pageId, refs, "review_required", "Denied audit events are surfaced for governance review without payload export or mutation.", { labelMessage: { key: "adminOperationsEvidence_checklist_audit_denial_review_label" as const }, summaryMessage: { key: "adminOperationsEvidence_checklist_audit_denial_review_summary" as const } }),
    checklistItem("read_route_catalog", "Read route catalog", "read_contract", "control-plane-read-shared-shell-v1", refs, "ready", "The admin review consumes existing canonical read route metadata.", { labelMessage: { key: "adminOperationsEvidence_checklist_read_route_catalog_label" as const }, summaryMessage: { key: "adminOperationsEvidence_checklist_read_route_catalog_summary" as const } }),
    checklistItem("forbidden_output_guard", "Forbidden output guard", "read_contract", "control-plane-read-negative-contract-v1", refs, "locked", "Forbidden output keys stay blocked across tenant, audit, and operations summaries.", { labelMessage: { key: "adminOperationsEvidence_checklist_forbidden_output_guard_label" as const }, summaryMessage: { key: "adminOperationsEvidence_checklist_forbidden_output_guard_summary" as const } }),
    checklistItem("gateway_evidence_review", "Gateway evidence review", "gateway", source.modelGatewayEvidenceReview.pageId, refs, "offline_only", "Gateway review contributes provider/profile, route, usage, audit, risk, and lock evidence.", { labelMessage: { key: "adminOperationsEvidence_checklist_gateway_evidence_review_label" as const }, summaryMessage: { key: "adminOperationsEvidence_checklist_gateway_evidence_review_summary" as const } }),
    checklistItem("gateway_key_quota", "Gateway key and quota", "gateway", "gateway-api-key-quota-readiness", refs, "locked", "API key lifecycle, quota execution, rate limit, and cost ledger are readable only as readiness evidence.", { labelMessage: { key: "adminOperationsEvidence_checklist_gateway_key_quota_label" as const }, summaryMessage: { key: "adminOperationsEvidence_checklist_gateway_key_quota_summary" as const } }),
    checklistItem("production_secret_boundary", "Secret boundary", "production_ops", "production-ops-config-secret-boundary", refs, "locked", "Config and secret boundary evidence keeps secret values out of committed and rendered output.", { labelMessage: { key: "adminOperationsEvidence_checklist_production_secret_boundary_label" as const }, summaryMessage: { key: "adminOperationsEvidence_checklist_production_secret_boundary_summary" as const } }),
    checklistItem("secret_backend_contract", "Secret backend contract", "production_ops", "production-secret-backend-contract", refs, "locked", "External secret backend contract exists, but implementation and resolver are still absent.", { labelMessage: { key: "adminOperationsEvidence_checklist_secret_backend_contract_label" as const }, summaryMessage: { key: "adminOperationsEvidence_checklist_secret_backend_contract_summary" as const } }),
    checklistItem("secret_reference_schema", "Secret ref schema", "production_ops", "production-secret-reference-contract", refs, "locked", "Secret refs are schema evidence only and do not resolve production credentials.", { labelMessage: { key: "adminOperationsEvidence_checklist_secret_reference_schema_label" as const }, summaryMessage: { key: "adminOperationsEvidence_checklist_secret_reference_schema_summary" as const } }),
    checklistItem("startup_supervisor", "Startup supervisor", "production_ops", "production-ops-startup-supervisor-boundary", refs, "locked", "Supervisor boundary is documented without process manager implementation.", { labelMessage: { key: "adminOperationsEvidence_checklist_startup_supervisor_label" as const }, summaryMessage: { key: "adminOperationsEvidence_checklist_startup_supervisor_summary" as const } }),
    checklistItem("environment_isolation", "Environment isolation", "production_ops", "production-ops-environment-isolation-boundary", refs, "locked", "Environment isolation remains static boundary evidence, not production deployment readiness.", { labelMessage: { key: "adminOperationsEvidence_checklist_environment_isolation_label" as const }, summaryMessage: { key: "adminOperationsEvidence_checklist_environment_isolation_summary" as const } }),
    checklistItem("deployment_static_smoke", "Deployment static smoke", "production_ops", "production-ops-deployment-readiness-smoke", refs, "offline_only", "Deployment readiness only covers static compose expansion and does not run test or production smoke.", { labelMessage: { key: "adminOperationsEvidence_checklist_deployment_static_smoke_label" as const }, summaryMessage: { key: "adminOperationsEvidence_checklist_deployment_static_smoke_summary" as const } }),
    checklistItem("container_smoke_record", "Container smoke record", "production_ops", "production-ops-container-smoke-record-template", refs, "offline_only", "Container smoke record template and local mock run are audit evidence, not a production preflight.", { labelMessage: { key: "adminOperationsEvidence_checklist_container_smoke_record_label" as const }, summaryMessage: { key: "adminOperationsEvidence_checklist_container_smoke_record_summary" as const } }),
    checklistItem("auth_store_transition", "Auth and store transition", "read_contract", "control-plane-read-auth-store-transition-preconditions-v1", refs, "locked", "Auth middleware and read store repository transition conditions remain unsatisfied.", { labelMessage: { key: "adminOperationsEvidence_checklist_auth_store_transition_label" as const }, summaryMessage: { key: "adminOperationsEvidence_checklist_auth_store_transition_summary" as const } }),
    checklistItem("implementation_trigger_review", "Implementation trigger review", "read_contract", "control-plane-read-implementation-trigger-review-v1", refs, "locked", "Current implementation trigger review keeps repository, selector, auth, and adapter work closed.", { labelMessage: { key: "adminOperationsEvidence_checklist_implementation_trigger_review_label" as const }, summaryMessage: { key: "adminOperationsEvidence_checklist_implementation_trigger_review_summary" as const } }),
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
  copy: EvidenceProjectionCopy,
): AdminOperationsEvidenceChecklistItem {
  return {
    ...copy,
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
      summaryMessage: { key: "adminOperationsEvidence_buildOperationalRisks_summary" as const, values: { actorSubjectRef: event.actorSubjectRef, resourceRef: event.resourceRef, failureCode: event.failureCode } },
      operatorQuestion: "Which scope or tenant binding should be reviewed before production auth is introduced?",
      operatorQuestionMessage: { key: "adminOperationsEvidence_buildOperationalRisks_operatorQuestion" as const },
    }));
  const gatewayRisks: AdminOperationsRisk[] = [
    {
      riskId: "gateway_route_risks",
      label: "Gateway route risks",
      labelMessage: { key: "adminOperationsEvidence_gateway_route_risks_label" as const },
      sourceSurface: "gateway",
      status: source.modelGatewayEvidenceReview.routeRisks.length > 0 ? "review_required" : "locked",
      evidenceRef: source.modelGatewayEvidenceReview.pageId,
      summary: `${source.modelGatewayEvidenceReview.routeRisks.length} route risks remain visible for provider/profile review.`,
      summaryMessage: { key: "adminOperationsEvidence_gateway_route_risks_summary" as const, values: { routeRisksCount: source.modelGatewayEvidenceReview.routeRisks.length } },
      operatorQuestion: "Which provider/profile mismatch or selection case should block execution first?",
      operatorQuestionMessage: { key: "adminOperationsEvidence_gateway_route_risks_operatorQuestion" as const },
    },
    {
      riskId: "gateway_usage_risks",
      label: "Gateway usage risks",
      labelMessage: { key: "adminOperationsEvidence_gateway_usage_risks_label" as const },
      sourceSurface: "gateway",
      status: source.modelGatewayEvidenceReview.usageRisks.length > 0 ? "review_required" : "locked",
      evidenceRef: source.modelGatewayEvidenceReview.pageId,
      summary: `${source.modelGatewayEvidenceReview.usageRisks.length} usage risks remain visible without enforcement or cost writes.`,
      summaryMessage: { key: "adminOperationsEvidence_gateway_usage_risks_summary" as const, values: { usageRisksCount: source.modelGatewayEvidenceReview.usageRisks.length } },
      operatorQuestion: "Which usage, quota, or cost prerequisite is absent before production traffic?",
      operatorQuestionMessage: { key: "adminOperationsEvidence_gateway_usage_risks_operatorQuestion" as const },
    },
    {
      riskId: "gateway_audit_risks",
      label: "Gateway audit risks",
      labelMessage: { key: "adminOperationsEvidence_gateway_audit_risks_label" as const },
      sourceSurface: "gateway",
      status: source.modelGatewayEvidenceReview.auditRisks.length > 0 ? "review_required" : "locked",
      evidenceRef: source.modelGatewayEvidenceReview.auditRef,
      summary: `${source.modelGatewayEvidenceReview.auditRisks.length} gateway audit risks remain governance evidence only.`,
      summaryMessage: { key: "adminOperationsEvidence_gateway_audit_risks_summary" as const, values: { auditRisksCount: source.modelGatewayEvidenceReview.auditRisks.length } },
      operatorQuestion: "Which audit decision or trace should be reviewed before any production route is enabled?",
      operatorQuestionMessage: { key: "adminOperationsEvidence_gateway_audit_risks_operatorQuestion" as const },
    },
  ];
  const operationsRisks: AdminOperationsRisk[] = [
    {
      riskId: "production_secret_backend_missing",
      label: "Production secret backend",
      labelMessage: { key: "adminOperationsEvidence_production_secret_backend_missing_label" as const },
      sourceSurface: "secret",
      status: "locked",
      evidenceRef: "production-secret-backend-implementation-readiness",
      summary: "Secret backend implementation readiness exists, but resolver injection and backend implementation are not present.",
      summaryMessage: { key: "adminOperationsEvidence_production_secret_backend_missing_summary" as const },
      operatorQuestion: "Which backend, profile binding, rotation, and audit policy must be selected before resolver work starts?",
      operatorQuestionMessage: { key: "adminOperationsEvidence_production_secret_backend_missing_operatorQuestion" as const },
    },
    {
      riskId: "process_supervisor_missing",
      label: "Process supervisor",
      labelMessage: { key: "adminOperationsEvidence_process_supervisor_missing_label" as const },
      sourceSurface: "deployment",
      status: "locked",
      evidenceRef: "production-ops-startup-supervisor-boundary",
      summary: "Startup and supervisor boundary is documented, while a real process supervisor is still absent.",
      summaryMessage: { key: "adminOperationsEvidence_process_supervisor_missing_summary" as const },
      operatorQuestion: "Which supervisor and restart policy should be chosen before production packaging?",
      operatorQuestionMessage: { key: "adminOperationsEvidence_process_supervisor_missing_operatorQuestion" as const },
    },
    {
      riskId: "environment_isolation_static",
      label: "Environment isolation",
      labelMessage: { key: "adminOperationsEvidence_environment_isolation_static_label" as const },
      sourceSurface: "deployment",
      status: "locked",
      evidenceRef: "production-ops-environment-isolation-boundary",
      summary: "Environment isolation is static evidence and does not prove test or production environment separation.",
      summaryMessage: { key: "adminOperationsEvidence_environment_isolation_static_summary" as const },
      operatorQuestion: "Which test and production smoke records are needed before deployment readiness can advance?",
      operatorQuestionMessage: { key: "adminOperationsEvidence_environment_isolation_static_operatorQuestion" as const },
    },
    {
      riskId: "implementation_trigger_unsatisfied",
      label: "Implementation trigger",
      labelMessage: { key: "adminOperationsEvidence_implementation_trigger_unsatisfied_label" as const },
      sourceSurface: "auth_store",
      status: "locked",
      evidenceRef: "control-plane-read-implementation-trigger-review-v1",
      summary: "Schema artifact, store selector, production auth, and adapter smoke triggers are all not satisfied.",
      summaryMessage: { key: "adminOperationsEvidence_implementation_trigger_unsatisfied_summary" as const },
      operatorQuestion: "Which trigger evidence should be completed before any repository or auth implementation begins?",
      operatorQuestionMessage: { key: "adminOperationsEvidence_implementation_trigger_unsatisfied_operatorQuestion" as const },
    },
    {
      riskId: "radish_auth_not_ready",
      label: "Radish auth",
      labelMessage: { key: "adminOperationsEvidence_radish_auth_not_ready_label" as const },
      sourceSurface: "auth_store",
      status: "locked",
      evidenceRef: "control-plane-read-production-auth-readiness-v1",
      summary: "Production auth readiness defines future issuer, claim, tenant, and scope requirements without token validation.",
      summaryMessage: { key: "adminOperationsEvidence_radish_auth_not_ready_summary" as const },
      operatorQuestion: "Which issuer discovery and claim mapping evidence is required before auth middleware work?",
      operatorQuestionMessage: { key: "adminOperationsEvidence_radish_auth_not_ready_operatorQuestion" as const },
    },
  ];

  return [...deniedAuditRisks, ...gatewayRisks, ...operationsRisks].slice(0, 12);
}

function buildBoundaryLocks(source: AdminOperationsReviewSource): AdminOperationsBoundaryLock[] {
  return [
    {
      lockId: "tenant_mutation",
      label: "Tenant mutation",
      labelMessage: { key: "adminOperationsEvidence_tenant_mutation_label" as const },
      sourceSurface: "tenant",
      status: "locked",
      missingPrerequisite: "tenant write model, mutation API, auth policy, and audit write contract",
      missingPrerequisiteMessage: { key: "adminOperationsEvidence_tenant_mutation_missingPrerequisite" as const },
      summary: "Tenant overview is a read-side summary and cannot create, update, archive, or migrate a tenant.",
      summaryMessage: { key: "adminOperationsEvidence_tenant_mutation_summary" as const },
    },
    {
      lockId: "raw_audit_export",
      label: "Raw audit export",
      labelMessage: { key: "adminOperationsEvidence_raw_audit_export_label" as const },
      sourceSurface: "audit",
      status: "locked",
      missingPrerequisite: "audit export policy and sensitive projection review",
      missingPrerequisiteMessage: { key: "adminOperationsEvidence_raw_audit_export_missingPrerequisite" as const },
      summary: "Audit rows remain sanitized summaries and cannot reveal payloads, edit records, or delete records.",
      summaryMessage: { key: "adminOperationsEvidence_raw_audit_export_summary" as const },
    },
    {
      lockId: "radish_auth",
      label: "Radish auth",
      labelMessage: { key: "adminOperationsEvidence_radish_auth_label" as const },
      sourceSurface: "auth",
      status: "locked",
      missingPrerequisite: "Radish auth issuer discovery, token validation, claim mapping, and auth middleware",
      missingPrerequisiteMessage: { key: "adminOperationsEvidence_radish_auth_missingPrerequisite" as const },
      summary: "Admin review does not validate production tokens or replace Radish identity truth.",
      summaryMessage: { key: "adminOperationsEvidence_radish_auth_summary" as const },
    },
    {
      lockId: "database_attachment",
      label: "Database attachment",
      labelMessage: { key: "adminOperationsEvidence_database_attachment_label" as const },
      sourceSurface: "database",
      status: "locked",
      missingPrerequisite: "schema artifact manifest, migrations, read-only role smoke, and durable adapter smoke",
      missingPrerequisiteMessage: { key: "adminOperationsEvidence_database_attachment_missingPrerequisite" as const },
      summary: "Read-side data remains fixture/fake-store backed until database implementation triggers are satisfied.",
      summaryMessage: { key: "adminOperationsEvidence_database_attachment_summary" as const },
    },
    {
      lockId: "repository_adapter",
      label: "Repository adapter",
      labelMessage: { key: "adminOperationsEvidence_repository_adapter_label" as const },
      sourceSurface: "repository",
      status: "locked",
      missingPrerequisite: "repository interface, adapter implementation, store selector, and adapter smoke evidence",
      missingPrerequisiteMessage: { key: "adminOperationsEvidence_repository_adapter_missingPrerequisite" as const },
      summary: "Repository adapter work stays blocked by implementation trigger review.",
      summaryMessage: { key: "adminOperationsEvidence_repository_adapter_summary" as const },
    },
    {
      lockId: "production_secret_resolver",
      label: "Secret resolver",
      labelMessage: { key: "adminOperationsEvidence_production_secret_resolver_label" as const },
      sourceSurface: "secret",
      status: "locked",
      missingPrerequisite: "production secret backend implementation and resolver injection policy",
      missingPrerequisiteMessage: { key: "adminOperationsEvidence_production_secret_resolver_missingPrerequisite" as const },
      summary: "Secret references are allowed as schema evidence, while secret values and resolver execution stay unavailable.",
      summaryMessage: { key: "adminOperationsEvidence_production_secret_resolver_summary" as const },
    },
    {
      lockId: "api_key_lifecycle",
      label: "API key lifecycle",
      labelMessage: { key: "adminOperationsEvidence_api_key_lifecycle_label" as const },
      sourceSurface: "gateway",
      status: "locked",
      missingPrerequisite: "key issuance, hashing, storage, validation, rotation, and revocation contract",
      missingPrerequisiteMessage: { key: "adminOperationsEvidence_api_key_lifecycle_missingPrerequisite" as const },
      summary: "Admin review can inspect key summaries through gateway evidence but cannot create, reveal, hash, or validate keys.",
      summaryMessage: { key: "adminOperationsEvidence_api_key_lifecycle_summary" as const },
    },
    {
      lockId: "quota_enforcement",
      label: "Quota enforcement",
      labelMessage: { key: "adminOperationsEvidence_quota_enforcement_label" as const },
      sourceSurface: "gateway",
      status: "locked",
      missingPrerequisite: "gateway quota policy execution, rate limit, and billing policy",
      missingPrerequisiteMessage: { key: "adminOperationsEvidence_quota_enforcement_missingPrerequisite" as const },
      summary: "Quota is rendered as policy evidence and does not throttle traffic or reject requests.",
      summaryMessage: { key: "adminOperationsEvidence_quota_enforcement_summary" as const },
    },
    {
      lockId: "cost_record_write",
      label: "Cost record write",
      labelMessage: { key: "adminOperationsEvidence_cost_record_write_label" as const },
      sourceSurface: "gateway",
      status: "locked",
      missingPrerequisite: "durable pricing policy and cost ledger",
      missingPrerequisiteMessage: { key: "adminOperationsEvidence_cost_record_write_missingPrerequisite" as const },
      summary: "Estimated cost remains read-side evidence and cannot create or update cost records.",
      summaryMessage: { key: "adminOperationsEvidence_cost_record_write_summary" as const },
    },
    {
      lockId: "production_gateway",
      label: "Production gateway",
      labelMessage: { key: "adminOperationsEvidence_production_gateway_label" as const },
      sourceSurface: "gateway",
      status: "locked",
      missingPrerequisite: "production API consumer, auth, secret resolver, quota enforcement, and provider execution policy",
      missingPrerequisiteMessage: { key: "adminOperationsEvidence_production_gateway_missingPrerequisite" as const },
      summary: "Gateway evidence remains offline and cannot route production requests.",
      summaryMessage: { key: "adminOperationsEvidence_production_gateway_summary" as const },
    },
    {
      lockId: "deployment_preflight",
      label: "Deployment preflight",
      labelMessage: { key: "adminOperationsEvidence_deployment_preflight_label" as const },
      sourceSurface: "deployment",
      status: "locked",
      missingPrerequisite: "explicit test or production preflight window and smoke record",
      missingPrerequisiteMessage: { key: "adminOperationsEvidence_deployment_preflight_missingPrerequisite" as const },
      summary: "Deployment evidence is static unless an explicit run window creates a new smoke record.",
      summaryMessage: { key: "adminOperationsEvidence_deployment_preflight_summary" as const },
    },
    {
      lockId: "workflow_runtime_execution",
      label: "Workflow runtime",
      labelMessage: { key: "adminOperationsEvidence_workflow_runtime_execution_label" as const },
      sourceSurface: "workflow",
      status: "locked",
      missingPrerequisite: "runtime runner, durable run/result store, confirmation decision store, writeback policy, and replay policy",
      missingPrerequisiteMessage: { key: "adminOperationsEvidence_workflow_runtime_execution_missingPrerequisite" as const },
      summary: "Workflow runtime, confirmation, writeback, replay, and business truth mutation remain unavailable.",
      summaryMessage: { key: "adminOperationsEvidence_workflow_runtime_execution_summary" as const },
    },
    ...source.modelGatewayEvidenceReview.lockedCapabilities.slice(0, 4).map((capability) => ({
      lockId: `gateway_${capability.capabilityId}`,
      label: capability.label,
      labelMessage: capability.labelMessage,
      sourceSurface: "gateway" as const,
      status: "locked" as const,
      missingPrerequisite: capability.missingPrerequisite,
      missingPrerequisiteMessage: capability.missingPrerequisiteMessage,
      summary: capability.summary,
      summaryMessage: capability.summaryMessage,
    })),
  ];
}

function buildReviewNarrativeMessage(
  source: AdminOperationsReviewSource,
  risks: AdminOperationsRisk[],
  locks: AdminOperationsBoundaryLock[],
): EvidenceProjectionMessage {
  const deniedAuditEvents = source.adminAuditLog.auditEvents.filter((event) => event.decision === "denied").length;
  return { key: "adminOperationsEvidence_reviewNarrative" as const, values: { tenantOverviewPageId: source.tenantOverview.pageId, adminAuditLogPageId: source.adminAuditLog.pageId, modelGatewayEvidenceReviewPageId: source.modelGatewayEvidenceReview.pageId, routesCount: source.readShell.catalog.routes.length, deniedAuditEvents: deniedAuditEvents, risksCount: risks.length, locksCount: locks.length } };
}
