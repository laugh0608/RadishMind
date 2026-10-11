import type { EvidenceProjectionCopy, EvidenceProjectionMessage } from "./evidenceProjectionCopy.ts";
import type { AdminAuditLogViewModel } from "./adminAuditLog";
import type { AdminOperationsBoundaryLock, AdminOperationsReviewViewModel } from "./adminOperationsReview";
import type { AdminTenantOverviewViewModel } from "./adminTenantOverview";
import type { ModelGatewayEvidenceReviewStatus, ModelGatewayEvidenceReviewViewModel } from "./modelGatewayEvidenceReview";
import type {
  ModelGatewayProviderProfileEvidence,
  ModelGatewayRouteBindingEvidence,
  ModelGatewayRouteEvidenceViewModel,
  ModelGatewaySelectionCaseEvidence,
} from "./modelGatewayRouteEvidence";

export type AdminProviderDeploymentReviewStatus = ModelGatewayEvidenceReviewStatus;

export type AdminProviderDeploymentReviewSource = {
  tenantOverview: AdminTenantOverviewViewModel;
  adminAuditLog: AdminAuditLogViewModel;
  modelGatewayRouteEvidence: ModelGatewayRouteEvidenceViewModel;
  modelGatewayEvidenceReview: ModelGatewayEvidenceReviewViewModel;
  adminOperationsReview: AdminOperationsReviewViewModel;
};

export type AdminProviderProfileReadiness = EvidenceProjectionCopy & {
  readinessId: string;
  providerProfile: string;
  providerId: string;
  credentialState: string;
  deploymentMode: string;
  authMode: string;
  streaming: string;
  status: AdminProviderDeploymentReviewStatus;
  sourceRef: string;
  summary: string;
};

export type AdminModelRouteReadiness = EvidenceProjectionCopy & {
  routeId: string;
  path: string;
  selectedProfileId: string;
  selectionSource: string;
  status: AdminProviderDeploymentReviewStatus;
  sourceRef: string;
  summary: string;
};

export type AdminSecretDeploymentEvidence = EvidenceProjectionCopy & {
  evidenceId: string;
  label: string;
  evidenceKind: "secret" | "deployment" | "tenant" | "runtime";
  value: string;
  status: AdminProviderDeploymentReviewStatus;
  sourceRef: string;
  summary: string;
};

export type AdminProviderDeploymentOperatorRisk = EvidenceProjectionCopy & {
  riskId: string;
  label: string;
  riskSurface: "provider" | "route" | "secret" | "deployment" | "auth_store";
  status: "review_required" | "blocked" | "locked";
  evidenceRef: string;
  summary: string;
  operatorQuestion: string;
};

export type AdminProviderDeploymentLock = EvidenceProjectionCopy & {
  lockId: string;
  label: string;
  lockSurface: "provider" | "route" | "secret" | "deployment" | "auth" | "database" | "repository" | "gateway";
  status: "locked";
  missingPrerequisite: string;
  summary: string;
};

export type AdminProviderDeploymentReviewViewModel = EvidenceProjectionCopy & {
  pageId: "admin-provider-deployment-review-offline";
  sourcePageIds: string[];
  reviewMode: "offline_admin_provider_deployment_readiness";
  tenantRef: string;
  requestId: string;
  auditRef: string;
  deploymentStatusRef: string;
  reviewNarrative: string;
  providerProfiles: AdminProviderProfileReadiness[];
  modelRoutes: AdminModelRouteReadiness[];
  secretDeploymentEvidence: AdminSecretDeploymentEvidence[];
  operatorRisks: AdminProviderDeploymentOperatorRisk[];
  lockedCapabilities: AdminProviderDeploymentLock[];
  canRenderProviderDeploymentReview: boolean;
  canInspectProviderProfileEvidenceLocally: true;
  canRequestProductionGateway: false;
  canMutateProviderProfile: false;
  canChangeModelRoute: false;
  canResolveProductionSecret: false;
  canRunDeploymentPreflight: false;
  canEnableRadishAuth: false;
  canAttachDatabase: false;
  canImplementRepositoryAdapter: false;
  canStartWorkflowRuntime: false;
};

export function buildAdminProviderDeploymentReviewViewModel(
  source: AdminProviderDeploymentReviewSource,
): AdminProviderDeploymentReviewViewModel {
  const providerProfiles = buildProviderProfiles(source.modelGatewayRouteEvidence.profiles);
  const modelRoutes = buildModelRoutes(source.modelGatewayRouteEvidence.routeBindings);
  const secretDeploymentEvidence = buildSecretDeploymentEvidence(source);
  const operatorRisks = buildOperatorRisks(source);
  const lockedCapabilities = buildLockedCapabilities(source);
  const deploymentStatusRef = source.tenantOverview.tenant?.deployment_status_ref ?? "deployment_status_missing";

  return {
    pageId: "admin-provider-deployment-review-offline",
    sourcePageIds: [
      source.tenantOverview.pageId,
      source.adminAuditLog.pageId,
      source.modelGatewayRouteEvidence.pageId,
      source.modelGatewayEvidenceReview.pageId,
      source.adminOperationsReview.pageId,
      "provider-capability-matrix-v1",
      "provider-health-smoke-v1",
      "provider-selection-policy-v1",
      "provider-retry-fallback-policy-v1",
      "production-ops-config-secret-boundary",
      "production-secret-backend-contract",
      "production-secret-backend-implementation-readiness",
      "production-secret-reference-contract",
      "production-ops-deployment-readiness-smoke",
      "production-ops-container-smoke-record-template",
      "control-plane-read-implementation-trigger-review-v1",
    ],
    reviewMode: "offline_admin_provider_deployment_readiness",
    tenantRef: source.tenantOverview.collection.tenantRef,
    requestId: source.tenantOverview.requestId,
    auditRef: source.adminAuditLog.auditRef,
    deploymentStatusRef,
    reviewNarrative: buildReviewNarrative(providerProfiles, modelRoutes, operatorRisks, lockedCapabilities),
    reviewNarrativeMessage: buildReviewNarrativeMessage(providerProfiles, modelRoutes, operatorRisks, lockedCapabilities),
    providerProfiles,
    modelRoutes,
    secretDeploymentEvidence,
    operatorRisks,
    lockedCapabilities,
    canRenderProviderDeploymentReview:
      source.tenantOverview.canRenderTenant &&
      source.adminAuditLog.canRenderAuditLog &&
      source.modelGatewayRouteEvidence.canRenderRouteEvidenceDetail &&
      source.modelGatewayEvidenceReview.canRenderEvidenceReview &&
      source.adminOperationsReview.canRenderAdminOperationsReview &&
      providerProfiles.length >= 5 &&
      modelRoutes.length >= 5 &&
      secretDeploymentEvidence.length >= 7 &&
      operatorRisks.length >= 8 &&
      lockedCapabilities.length >= 10,
    canInspectProviderProfileEvidenceLocally: true,
    canRequestProductionGateway: false,
    canMutateProviderProfile: false,
    canChangeModelRoute: false,
    canResolveProductionSecret: false,
    canRunDeploymentPreflight: false,
    canEnableRadishAuth: false,
    canAttachDatabase: false,
    canImplementRepositoryAdapter: false,
    canStartWorkflowRuntime: false,
  };
}

function buildReviewNarrative(
  providerProfiles: AdminProviderProfileReadiness[],
  modelRoutes: AdminModelRouteReadiness[],
  risks: AdminProviderDeploymentOperatorRisk[],
  locks: AdminProviderDeploymentLock[],
): string {
  const blockedProfiles = providerProfiles.filter((profile) => profile.status === "blocked").length;
  const routeReviews = modelRoutes.filter((route) => route.status === "review_required").length;
  return `${providerProfiles.length} provider/profile records, ${blockedProfiles} blocked credential cases, ${modelRoutes.length} model route bindings, ${routeReviews} route review items, ${risks.length} operator risks, and ${locks.length} locked capabilities are grouped for admin provider and deployment review.`;
}

function buildProviderProfiles(
  profiles: ModelGatewayProviderProfileEvidence[],
): AdminProviderProfileReadiness[] {
  return profiles.map((profile) => ({
    readinessId: `profile_${profile.profileId}`,
    providerProfile: profile.profileId,
    providerId: profile.providerId,
    credentialState: profile.credentialState,
    deploymentMode: profile.deploymentMode,
    authMode: profile.authMode,
    streaming: profile.streaming,
    status: profile.status,
    sourceRef: profile.evidenceRef,
    summary: profile.summary,
    summaryMessage: profile.summaryMessage,
  }));
}

function buildModelRoutes(routeBindings: ModelGatewayRouteBindingEvidence[]): AdminModelRouteReadiness[] {
  return routeBindings.map((route) => ({
    routeId: route.routeId,
    path: route.path,
    selectedProfileId: route.selectedProfileId,
    selectionSource: route.selectionSource,
    status: route.status,
    sourceRef: route.evidenceRef,
    summary: route.summary,
    summaryMessage: route.summaryMessage,
  }));
}

function buildSecretDeploymentEvidence(
  source: AdminProviderDeploymentReviewSource,
): AdminSecretDeploymentEvidence[] {
  const deploymentStatusRef = source.tenantOverview.tenant?.deployment_status_ref ?? "deployment_status_missing";
  const missingCredentials = source.modelGatewayRouteEvidence.profiles.filter(
    (profile) => profile.credentialState === "missing",
  ).length;
  const optionalCredentials = source.modelGatewayRouteEvidence.profiles.filter(
    (profile) => profile.credentialState === "optional_missing",
  ).length;

  return [
    {
      evidenceId: "tenant_deployment_status",
      label: "Tenant deployment status",
      labelMessage: { key: "adminDeploymentEvidence_tenant_deployment_status_label" as const },
      evidenceKind: "tenant",
      value: deploymentStatusRef,
      status: "offline_only",
      sourceRef: source.tenantOverview.pageId,
      summary: "Tenant deployment status is a read-side reference and does not launch a deployment preflight.",
      summaryMessage: { key: "adminDeploymentEvidence_tenant_deployment_status_summary" as const },
    },
    {
      evidenceId: "provider_credential_states",
      label: "Provider credential states",
      labelMessage: { key: "adminDeploymentEvidence_provider_credential_states_label" as const },
      evidenceKind: "runtime",
      value: `${missingCredentials} missing / ${optionalCredentials} optional`,
      valueMessage: { key: "adminDeploymentEvidence_provider_credential_states_value" as const, values: { missingCredentials: missingCredentials, optionalCredentials: optionalCredentials } },
      status: missingCredentials > 0 ? "review_required" : "offline_only",
      sourceRef: source.modelGatewayRouteEvidence.pageId,
      summary: "Credential states are rendered as configured, missing, optional_missing, or not_required without exposing values.",
      summaryMessage: { key: "adminDeploymentEvidence_provider_credential_states_summary" as const },
    },
    {
      evidenceId: "secret_backend_contract",
      label: "Secret backend contract",
      labelMessage: { key: "adminDeploymentEvidence_secret_backend_contract_label" as const },
      evidenceKind: "secret",
      value: "contract only",
      valueMessage: { key: "adminDeploymentEvidence_secret_backend_contract_value" as const },
      status: "locked",
      sourceRef: "production-secret-backend-contract",
      summary: "The external secret backend contract exists, while implementation and resolver injection remain absent.",
      summaryMessage: { key: "adminDeploymentEvidence_secret_backend_contract_summary" as const },
    },
    {
      evidenceId: "secret_backend_implementation",
      label: "Secret backend implementation",
      labelMessage: { key: "adminDeploymentEvidence_secret_backend_implementation_label" as const },
      evidenceKind: "secret",
      value: "not satisfied",
      valueMessage: { key: "adminDeploymentEvidence_secret_backend_implementation_value" as const },
      status: "locked",
      sourceRef: "production-secret-backend-implementation-readiness",
      summary: "Backend selection, profile binding, rotation, audit policy, and tests must be resolved before implementation.",
      summaryMessage: { key: "adminDeploymentEvidence_secret_backend_implementation_summary" as const },
    },
    {
      evidenceId: "secret_reference_schema",
      label: "Secret reference schema",
      labelMessage: { key: "adminDeploymentEvidence_secret_reference_schema_label" as const },
      evidenceKind: "secret",
      value: "schema evidence",
      valueMessage: { key: "adminDeploymentEvidence_secret_reference_schema_value" as const },
      status: "locked",
      sourceRef: "production-secret-reference-contract",
      summary: "Secret references can be checked as schema artifacts, but secret values and resolver execution stay unavailable.",
      summaryMessage: { key: "adminDeploymentEvidence_secret_reference_schema_summary" as const },
    },
    {
      evidenceId: "deployment_static_smoke",
      label: "Deployment static smoke",
      labelMessage: { key: "adminDeploymentEvidence_deployment_static_smoke_label" as const },
      evidenceKind: "deployment",
      value: "static only",
      valueMessage: { key: "adminDeploymentEvidence_deployment_static_smoke_value" as const },
      status: "offline_only",
      sourceRef: "production-ops-deployment-readiness-smoke",
      summary: "Deployment readiness covers compose expansion and static checks, not test environment smoke or production preflight.",
      summaryMessage: { key: "adminDeploymentEvidence_deployment_static_smoke_summary" as const },
    },
    {
      evidenceId: "container_smoke_record",
      label: "Container smoke record",
      labelMessage: { key: "adminDeploymentEvidence_container_smoke_record_label" as const },
      evidenceKind: "deployment",
      value: "local mock recorded",
      valueMessage: { key: "adminDeploymentEvidence_container_smoke_record_value" as const },
      status: "offline_only",
      sourceRef: "production-ops-container-smoke-record-template",
      summary: "The local mock container smoke record is audit evidence and cannot be treated as production readiness.",
      summaryMessage: { key: "adminDeploymentEvidence_container_smoke_record_summary" as const },
    },
    {
      evidenceId: "startup_supervisor_boundary",
      label: "Startup supervisor",
      labelMessage: { key: "adminDeploymentEvidence_startup_supervisor_boundary_label" as const },
      evidenceKind: "deployment",
      value: "boundary only",
      valueMessage: { key: "adminDeploymentEvidence_startup_supervisor_boundary_value" as const },
      status: "locked",
      sourceRef: "production-ops-startup-supervisor-boundary",
      summary: "Startup and supervisor evidence documents the boundary without process supervisor implementation.",
      summaryMessage: { key: "adminDeploymentEvidence_startup_supervisor_boundary_summary" as const },
    },
  ];
}

function buildOperatorRisks(source: AdminProviderDeploymentReviewSource): AdminProviderDeploymentOperatorRisk[] {
  const profileRisks = source.modelGatewayRouteEvidence.profiles
    .filter((profile) => profile.status === "blocked" || profile.status === "review_required")
    .map((profile) => riskFromProfile(profile));
  const routeRisks = source.modelGatewayRouteEvidence.selectionCases
    .filter((selectionCase) => selectionCase.status === "blocked" || selectionCase.status === "review_required")
    .map((selectionCase) => riskFromSelectionCase(selectionCase));
  const operationsRisks = source.adminOperationsReview.operationalRisks
    .filter(
      (risk) =>
        risk.sourceSurface === "secret" ||
        risk.sourceSurface === "deployment" ||
        risk.sourceSurface === "auth_store" ||
        risk.sourceSurface === "gateway",
    )
    .map((risk) => ({
      riskId: `operations_${risk.riskId}`,
      label: risk.label,
      labelMessage: risk.labelMessage,
      riskSurface: operationsRiskSurface(risk.sourceSurface),
      status: risk.status,
      evidenceRef: risk.evidenceRef,
      summary: risk.summary,
      summaryMessage: risk.summaryMessage,
      operatorQuestion: risk.operatorQuestion,
      operatorQuestionMessage: risk.operatorQuestionMessage,
    }));

  return [...profileRisks, ...routeRisks, ...operationsRisks].slice(0, 14);
}

function operationsRiskSurface(
  sourceSurface: AdminProviderDeploymentReviewSource["adminOperationsReview"]["operationalRisks"][number]["sourceSurface"],
): AdminProviderDeploymentOperatorRisk["riskSurface"] {
  if (sourceSurface === "gateway") {
    return "route";
  }
  if (sourceSurface === "secret" || sourceSurface === "deployment" || sourceSurface === "auth_store") {
    return sourceSurface;
  }
  return "route";
}

function riskFromProfile(profile: ModelGatewayProviderProfileEvidence): AdminProviderDeploymentOperatorRisk {
  return {
    riskId: `profile_${profile.profileId}`,
    label: profile.profileName,
    riskSurface: "provider",
    status: profile.status === "blocked" ? "blocked" : "review_required",
    evidenceRef: profile.evidenceRef,
    summary: profile.summary,
    summaryMessage: profile.summaryMessage,
    operatorQuestion: "Which credential source, provider profile binding, or secret reference must be reviewed first?",
    operatorQuestionMessage: { key: "adminDeploymentEvidence_riskFromProfile_operatorQuestion" as const },
  };
}

function riskFromSelectionCase(
  selectionCase: ModelGatewaySelectionCaseEvidence,
): AdminProviderDeploymentOperatorRisk {
  return {
    riskId: `selection_${selectionCase.caseId}`,
    label: selectionCase.label,
    labelMessage: selectionCase.labelMessage,
    riskSurface: "route",
    status: selectionCase.status === "blocked" ? "blocked" : "review_required",
    evidenceRef: selectionCase.evidenceRef,
    summary: selectionCase.summary,
    summaryMessage: selectionCase.summaryMessage,
    operatorQuestion: "Should this route input be rejected explicitly before any production provider routing is enabled?",
    operatorQuestionMessage: { key: "adminDeploymentEvidence_riskFromSelectionCase_operatorQuestion" as const },
  };
}

function buildLockedCapabilities(source: AdminProviderDeploymentReviewSource): AdminProviderDeploymentLock[] {
  const adminLocks = source.adminOperationsReview.boundaryLocks
    .filter((lock) =>
      ["secret", "gateway", "deployment", "auth", "database", "repository"].includes(lock.sourceSurface),
    )
    .map((lock) => lockFromAdminLock(lock));
  const gatewayLocks = source.modelGatewayEvidenceReview.lockedCapabilities
    .filter((capability) => capability.sourceSurface === "route" || capability.sourceSurface === "usage")
    .slice(0, 6)
    .map((capability) => ({
      lockId: `gateway_${capability.capabilityId}`,
      label: capability.label,
      labelMessage: capability.labelMessage,
      lockSurface: capability.sourceSurface === "route" ? ("route" as const) : ("gateway" as const),
      status: "locked" as const,
      missingPrerequisite: capability.missingPrerequisite,
      missingPrerequisiteMessage: capability.missingPrerequisiteMessage,
      summary: capability.summary,
      summaryMessage: capability.summaryMessage,
    }));

  return [...adminLocks, ...gatewayLocks].slice(0, 18);
}

function lockFromAdminLock(lock: AdminOperationsBoundaryLock): AdminProviderDeploymentLock {
  const lockSurface = lock.sourceSurface === "workflow" || lock.sourceSurface === "tenant" || lock.sourceSurface === "audit"
    ? "gateway"
    : lock.sourceSurface;
  return {
    lockId: `admin_${lock.lockId}`,
    label: lock.label,
    labelMessage: lock.labelMessage,
    lockSurface,
    status: "locked",
    missingPrerequisite: lock.missingPrerequisite,
    missingPrerequisiteMessage: lock.missingPrerequisiteMessage,
    summary: lock.summary,
    summaryMessage: lock.summaryMessage,
  };
}

function buildReviewNarrativeMessage(
  providerProfiles: AdminProviderProfileReadiness[],
  modelRoutes: AdminModelRouteReadiness[],
  risks: AdminProviderDeploymentOperatorRisk[],
  locks: AdminProviderDeploymentLock[],
): EvidenceProjectionMessage {
  const blockedProfiles = providerProfiles.filter((profile) => profile.status === "blocked").length;
  const routeReviews = modelRoutes.filter((route) => route.status === "review_required").length;
  return { key: "adminDeploymentEvidence_reviewNarrative" as const, values: { providerProfilesCount: providerProfiles.length, blockedProfiles: blockedProfiles, modelRoutesCount: modelRoutes.length, routeReviews: routeReviews, risksCount: risks.length, locksCount: locks.length } };
}
