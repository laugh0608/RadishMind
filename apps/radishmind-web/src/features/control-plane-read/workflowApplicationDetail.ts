import {
  CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS,
  CONTROL_PLANE_READ_ROUTE_DEFINITIONS,
  CONTROL_PLANE_READ_ROUTES,
  controlPlaneReadResponseHasForbiddenOutput,
} from "../../../../../contracts/typescript/control-plane-read-api";
import type { WorkspaceApplicationRow } from "./workspaceApplications";

export type WorkflowApplicationDetailSourceMetadata = {
  tenantRef: string;
  requestId: string;
  auditRef: string;
};

export type WorkflowApplicationRiskSummary = {
  label: string;
  riskLevel: "low" | "medium" | "high";
  requiresConfirmationCapable: boolean;
  summary: string;
  policyRef: string;
};

export type WorkflowApplicationRouteMetadata = {
  sourceRouteId: "application-summary-list-route";
  draftRouteId: "application-detail-read-draft";
  routePath: typeof CONTROL_PLANE_READ_ROUTES.applications;
  requestId: string;
  auditRef: string;
};

export type WorkflowApplicationBlockedCapabilityPreview = {
  capabilityId: string;
  label: string;
  status: "blocked";
  reason: string;
  missingPrerequisite: string;
  auditRef: string;
};

export type WorkflowApplicationDetailViewModel = {
  pageId: "workflow-application-detail-read";
  sourcePageId: "workspace-applications";
  sourceRouteId: "application-summary-list-route";
  draftRouteId: "application-detail-read-draft";
  routePath: typeof CONTROL_PLANE_READ_ROUTES.applications;
  requestId: string;
  auditRef: string;
  applicationId: string;
  tenantRef: string;
  ownerSubjectRef: string;
  applicationType: string;
  displayName: string;
  lifecycleStatus: string;
  providerProfileRef: string;
  latestWorkflowDefinitionRef: string;
  latestRunRef: string;
  lastRunStatus: string;
  updatedAt: string;
  riskSummary: WorkflowApplicationRiskSummary;
  routeMetadata: WorkflowApplicationRouteMetadata;
  blockedCapabilities: WorkflowApplicationBlockedCapabilityPreview[];
  forbiddenProjectionBlocked: boolean;
  canRenderApplicationDetail: boolean;
  canRequestLiveBackend: false;
  canMutate: false;
  canCreateApplication: false;
  canEditApplication: false;
  canDeleteApplication: false;
  canPublishApplication: false;
  canStartRun: false;
  canExecuteWorkflow: false;
  canSubmitConfirmationDecision: false;
  canWriteBusinessTruth: false;
  canReplayRun: false;
};

type WorkflowApplicationDetailEnrichment = {
  lifecycleStatus: string;
  providerProfileRef: string;
  latestRunRef: string;
  riskSummary: WorkflowApplicationRiskSummary;
};

const DEFAULT_APPLICATION: WorkspaceApplicationRow = {
  applicationRef: "app_flow_copilot",
  displayName: "RadishFlow Copilot",
  applicationKind: "workflow_copilot",
  ownerSubjectRef: "subject_platform_ops",
  latestWorkflowDefinitionRef: "wf_radishflow_copilot_latest",
  lastRunStatus: "succeeded",
  updatedAt: "2026-05-31T10:20:00Z",
};

const DEFAULT_SOURCE_METADATA: WorkflowApplicationDetailSourceMetadata = {
  tenantRef: "tenant_demo",
  requestId: "req_workspace_applications_demo",
  auditRef: "audit_workspace_applications_demo",
};

const APPLICATION_DETAIL_ENRICHMENT: Record<string, WorkflowApplicationDetailEnrichment> = {
  app_flow_copilot: {
    lifecycleStatus: "active",
    providerProfileRef: "profile:radishmind-default-workflow",
    latestRunRef: "run_radishflow_copilot_20260531_001",
    riskSummary: {
      label: "Workflow candidate action risk",
      riskLevel: "medium",
      requiresConfirmationCapable: true,
      summary:
        "The application can preview candidate RadishFlow actions, but execution remains blocked until confirmation and executor gates exist.",
      policyRef: "workflow_function_surface_boundary_v1",
    },
  },
  app_docs_assistant: {
    lifecycleStatus: "blocked_review",
    providerProfileRef: "profile:radishmind-docs-qa",
    latestRunRef: "run_radish_docs_qa_20260531_002",
    riskSummary: {
      label: "Docs QA advisory risk",
      riskLevel: "low",
      requiresConfirmationCapable: false,
      summary:
        "The application only renders advisory document answers and keeps raw prompt, token, and writeback payloads out of the read surface.",
      policyRef: "control_plane_read_workspace_applications_v1",
    },
  },
};

export function buildWorkflowApplicationDetailViewModel(
  application: WorkspaceApplicationRow = DEFAULT_APPLICATION,
  sourceMetadata: WorkflowApplicationDetailSourceMetadata = DEFAULT_SOURCE_METADATA,
): WorkflowApplicationDetailViewModel {
  const route = CONTROL_PLANE_READ_ROUTE_DEFINITIONS["application-summary-list-route"];
  const routePath = CONTROL_PLANE_READ_ROUTES.applications;
  const enrichment = APPLICATION_DETAIL_ENRICHMENT[application.applicationRef] ?? buildFallbackEnrichment(application);
  const blockedCapabilities = buildBlockedCapabilities(application, enrichment);
  const forbiddenProjectionBlocked = controlPlaneReadResponseHasForbiddenOutput({
    detail: { [CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS[8]]: "blocked" },
  });

  return {
    pageId: "workflow-application-detail-read",
    sourcePageId: "workspace-applications",
    sourceRouteId: "application-summary-list-route",
    draftRouteId: "application-detail-read-draft",
    routePath,
    requestId: sourceMetadata.requestId,
    auditRef: sourceMetadata.auditRef,
    applicationId: application.applicationRef,
    tenantRef: sourceMetadata.tenantRef,
    ownerSubjectRef: application.ownerSubjectRef,
    applicationType: application.applicationKind,
    displayName: application.displayName,
    lifecycleStatus: enrichment.lifecycleStatus,
    providerProfileRef: enrichment.providerProfileRef,
    latestWorkflowDefinitionRef: application.latestWorkflowDefinitionRef,
    latestRunRef: enrichment.latestRunRef,
    lastRunStatus: application.lastRunStatus,
    updatedAt: application.updatedAt,
    riskSummary: enrichment.riskSummary,
    routeMetadata: {
      sourceRouteId: "application-summary-list-route",
      draftRouteId: "application-detail-read-draft",
      routePath,
      requestId: sourceMetadata.requestId,
      auditRef: sourceMetadata.auditRef,
    },
    blockedCapabilities,
    forbiddenProjectionBlocked,
    canRenderApplicationDetail:
      route.path === routePath &&
      route.canMutate === false &&
      application.applicationRef.length > 0 &&
      blockedCapabilities.length >= 4,
    canRequestLiveBackend: false,
    canMutate: false,
    canCreateApplication: false,
    canEditApplication: false,
    canDeleteApplication: false,
    canPublishApplication: false,
    canStartRun: false,
    canExecuteWorkflow: false,
    canSubmitConfirmationDecision: false,
    canWriteBusinessTruth: false,
    canReplayRun: false,
  };
}

function buildFallbackEnrichment(application: WorkspaceApplicationRow): WorkflowApplicationDetailEnrichment {
  return {
    lifecycleStatus: application.lastRunStatus === "blocked" ? "blocked_review" : "active",
    providerProfileRef: `profile:${application.applicationKind}`,
    latestRunRef: `run_ref_for_${application.applicationRef}`,
    riskSummary: {
      label: "Read-only application risk",
      riskLevel: application.lastRunStatus === "blocked" ? "medium" : "low",
      requiresConfirmationCapable: application.applicationKind.includes("workflow"),
      summary:
        "Application detail is derived from the summary row and committed offline metadata; execution and writeback stay blocked.",
      policyRef: "workflow_application_detail_read_v1",
    },
  };
}

function buildBlockedCapabilities(
  application: WorkspaceApplicationRow,
  enrichment: WorkflowApplicationDetailEnrichment,
): WorkflowApplicationBlockedCapabilityPreview[] {
  return [
    {
      capabilityId: "blocked_application_mutation",
      label: "Application mutation",
      status: "blocked",
      reason: "Create, edit, delete, and publish flows are outside the current read-side surface.",
      missingPrerequisite: "application lifecycle mutation task card",
      auditRef: "audit_application_detail_mutation_blocked_demo",
    },
    {
      capabilityId: "blocked_workflow_execution",
      label: "Workflow execution",
      status: "blocked",
      reason: `${application.displayName} can show latest workflow references, but no workflow executor is connected.`,
      missingPrerequisite: "workflow executor implementation gate",
      auditRef: "audit_application_detail_executor_blocked_demo",
    },
    {
      capabilityId: "blocked_confirmation_decision",
      label: "Confirmation decision",
      status: "blocked",
      reason: enrichment.riskSummary.requiresConfirmationCapable
        ? "Human review is visible as a capability marker only; no decision submission path is available."
        : "This application does not require confirmation today, and the confirmation path is still not implemented.",
      missingPrerequisite: "confirmation flow implementation gate",
      auditRef: "audit_application_detail_confirmation_blocked_demo",
    },
    {
      capabilityId: "blocked_business_writeback",
      label: "Business writeback",
      status: "blocked",
      reason: "Upstream truth sources remain protected; this surface only renders sanitized route, run, and audit metadata.",
      missingPrerequisite: "business writeback policy and adapter gate",
      auditRef: "audit_application_detail_writeback_blocked_demo",
    },
    {
      capabilityId: "blocked_run_replay",
      label: "Run replay",
      status: "blocked",
      reason: "Latest run references are pointers only; replay and resume stay disabled.",
      missingPrerequisite: "durable run store and replay gate",
      auditRef: "audit_application_detail_replay_blocked_demo",
    },
  ];
}
