import type { WorkflowProjectionCopy } from "./workflowProjectionCopy.ts";
import {
  CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS,
  CONTROL_PLANE_READ_ROUTE_DEFINITIONS,
  CONTROL_PLANE_READ_ROUTES,
  controlPlaneReadResponseHasForbiddenOutput,
} from "../../../../../contracts/typescript/control-plane-read-api";
import {
  buildWorkflowExecutionPlanPreviewViewModel,
  type WorkflowExecutionPlanBlockedReason,
  type WorkflowExecutionPlanGate,
  type WorkflowExecutionPlanPreviewViewModel,
  type WorkflowExecutionPlanProviderRequirement,
} from "./workflowExecutionPlanPreview";

export type WorkflowRuntimeReadinessStatus = "satisfied" | "review_required" | "blocked";

export type WorkflowRuntimeReadinessArea =
  | "executor"
  | "provider"
  | "confirmation"
  | "store"
  | "audit"
  | "writeback"
  | "replay"
  | "auth_store"
  | "publish";

export type WorkflowRuntimeReadinessSummary = WorkflowProjectionCopy & {
  label: string;
  value: string;
  status: WorkflowRuntimeReadinessStatus;
  summary: string;
};

export type WorkflowRuntimeReadinessPrerequisite = WorkflowProjectionCopy & {
  prerequisiteId: string;
  label: string;
  area: WorkflowRuntimeReadinessArea;
  status: WorkflowRuntimeReadinessStatus;
  currentEvidence: string;
  missingPrerequisite: string;
  sourceRefs: string[];
  summary: string;
};

export type WorkflowRuntimeReadinessBlocker = WorkflowProjectionCopy & {
  blockerId: string;
  label: string;
  area: WorkflowRuntimeReadinessArea;
  severity: "critical" | "high";
  status: "blocked";
  sourceRef: string;
  missingPrerequisite: string;
  summary: string;
  auditRef: string;
};

export type WorkflowRuntimeReadinessGate = WorkflowProjectionCopy & {
  gateId: string;
  label: string;
  gateKind: "implementation_trigger" | "runtime" | "store" | "auth" | "policy";
  status: WorkflowRuntimeReadinessStatus;
  requiredBefore: "runtime_start" | "publish" | "writeback" | "replay" | "live_backend";
  evidenceRefs: string[];
  summary: string;
};

export type WorkflowRuntimeReadinessAuditMetadata = {
  sourcePageId: "workflow-execution-plan-preview-offline";
  sourceRouteId: "workflow-definition-summary-list-route";
  planRouteId: "workflow-execution-plan-preview-offline-draft";
  readinessRouteId: "workflow-runtime-readiness-inspector-offline-draft";
  requestId: string;
  auditRef: string;
  selectedDraftId: string;
};

export type WorkflowRuntimeReadinessInspectorViewModel = WorkflowProjectionCopy & {
  pageId: "workflow-runtime-readiness-inspector-offline";
  sourcePageId: "workflow-execution-plan-preview-offline";
  sourceRouteId: "workflow-definition-summary-list-route";
  planRouteId: "workflow-execution-plan-preview-offline-draft";
  readinessRouteId: "workflow-runtime-readiness-inspector-offline-draft";
  routePath: typeof CONTROL_PLANE_READ_ROUTES.workflowDefinitions;
  requestId: string;
  auditRef: string;
  selectedDraftId: string;
  summary: WorkflowRuntimeReadinessSummary[];
  runtimePrerequisites: WorkflowRuntimeReadinessPrerequisite[];
  readinessBlockers: WorkflowRuntimeReadinessBlocker[];
  implementationGates: WorkflowRuntimeReadinessGate[];
  auditMetadata: WorkflowRuntimeReadinessAuditMetadata;
  forbiddenProjectionBlocked: boolean;
  canRenderRuntimeReadinessInspector: boolean;
  canInspectReadinessLocally: true;
  canRequestLiveBackend: false;
  canPersistReadinessResult: false;
  canStartWorkflowRuntime: false;
  canExecuteWorkflow: false;
  canSubmitConfirmationDecision: false;
  canWriteBusinessTruth: false;
  canReplayRun: false;
  canEnableProductionAuth: false;
  canAttachDatabase: false;
  canImplementRepositoryAdapter: false;
};

const DEFAULT_EXECUTION_PLAN_PREVIEW = buildWorkflowExecutionPlanPreviewViewModel();

export function buildWorkflowRuntimeReadinessInspectorViewModel(
  executionPlanPreview: WorkflowExecutionPlanPreviewViewModel = DEFAULT_EXECUTION_PLAN_PREVIEW,
): WorkflowRuntimeReadinessInspectorViewModel {
  const route = CONTROL_PLANE_READ_ROUTE_DEFINITIONS["workflow-definition-summary-list-route"];
  const routePath = CONTROL_PLANE_READ_ROUTES.workflowDefinitions;
  const runtimePrerequisites = buildRuntimePrerequisites(executionPlanPreview);
  const readinessBlockers = buildReadinessBlockers(runtimePrerequisites, executionPlanPreview);
  const implementationGates = buildImplementationGates(runtimePrerequisites);
  const forbiddenProjectionBlocked = controlPlaneReadResponseHasForbiddenOutput({
    runtimeReadinessInspector: { [CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS[8]]: "blocked" },
  });

  return {
    pageId: "workflow-runtime-readiness-inspector-offline",
    sourcePageId: "workflow-execution-plan-preview-offline",
    sourceRouteId: "workflow-definition-summary-list-route",
    planRouteId: "workflow-execution-plan-preview-offline-draft",
    readinessRouteId: "workflow-runtime-readiness-inspector-offline-draft",
    routePath,
    requestId: executionPlanPreview.requestId,
    auditRef: executionPlanPreview.auditRef,
    selectedDraftId: executionPlanPreview.selectedDraftId,
    summary: buildSummary(executionPlanPreview, runtimePrerequisites, readinessBlockers, implementationGates),
    runtimePrerequisites,
    readinessBlockers,
    implementationGates,
    auditMetadata: {
      sourcePageId: "workflow-execution-plan-preview-offline",
      sourceRouteId: "workflow-definition-summary-list-route",
      planRouteId: "workflow-execution-plan-preview-offline-draft",
      readinessRouteId: "workflow-runtime-readiness-inspector-offline-draft",
      requestId: executionPlanPreview.requestId,
      auditRef: executionPlanPreview.auditRef,
      selectedDraftId: executionPlanPreview.selectedDraftId,
    },
    forbiddenProjectionBlocked,
    canRenderRuntimeReadinessInspector:
      route.path === routePath &&
      route.canMutate === false &&
      executionPlanPreview.canRenderExecutionPlanPreview &&
      runtimePrerequisites.length >= 9 &&
      readinessBlockers.length >= 7 &&
      implementationGates.length >= 5,
    canInspectReadinessLocally: true,
    canRequestLiveBackend: false,
    canPersistReadinessResult: false,
    canStartWorkflowRuntime: false,
    canExecuteWorkflow: false,
    canSubmitConfirmationDecision: false,
    canWriteBusinessTruth: false,
    canReplayRun: false,
    canEnableProductionAuth: false,
    canAttachDatabase: false,
    canImplementRepositoryAdapter: false,
  };
}

function buildSummary(
  executionPlanPreview: WorkflowExecutionPlanPreviewViewModel,
  runtimePrerequisites: WorkflowRuntimeReadinessPrerequisite[],
  readinessBlockers: WorkflowRuntimeReadinessBlocker[],
  implementationGates: WorkflowRuntimeReadinessGate[],
): WorkflowRuntimeReadinessSummary[] {
  return [
    {
      label: "Selected draft",
      labelMessage: { key: "readinessDraft" },
      value: executionPlanPreview.selectedDraftId,
      status: "review_required",
      summary: "Runtime readiness is inspected from the selected offline execution plan preview.",
      summaryMessage: { key: "readinessDraftSummary" },
    },
    {
      label: "Prerequisites",
      labelMessage: { key: "prerequisites" },
      value: String(runtimePrerequisites.length),
      status: "blocked",
      summary: "Prerequisites are grouped by executor, provider, confirmation, store, audit, writeback, replay, auth/store, and publish gates.",
      summaryMessage: { key: "readinessPrerequisitesSummary" },
    },
    {
      label: "Blockers",
      labelMessage: { key: "blockers" },
      value: String(readinessBlockers.length),
      status: "blocked",
      summary: "Runtime cannot start while executor, durable store, confirmation, auth/store, writeback, and replay blockers remain open.",
      summaryMessage: { key: "readinessBlockedSummary" },
    },
    {
      label: "Implementation gates",
      labelMessage: { key: "implementationGates" },
      value: String(implementationGates.length),
      status: "blocked",
      summary: "All future implementation triggers are displayed as blocked or review-only evidence.",
      summaryMessage: { key: "readinessGatesSummary" },
    },
  ];
}

function buildRuntimePrerequisites(
  executionPlanPreview: WorkflowExecutionPlanPreviewViewModel,
): WorkflowRuntimeReadinessPrerequisite[] {
  const runtimeReason = findBlockedReason(executionPlanPreview.blockedPlanReasons, "blocked_runtime");
  const publishReason = findBlockedReason(executionPlanPreview.blockedPlanReasons, "blocked_publish");
  const writebackReason = findBlockedReason(executionPlanPreview.blockedPlanReasons, "blocked_writeback");
  const replayReason = findBlockedReason(executionPlanPreview.blockedPlanReasons, "blocked_replay");
  const providerRequirement = findProviderRequirement(
    executionPlanPreview.providerProfileRequirements,
    "provider_profile_model_stage",
  );
  const toolRequirement = findProviderRequirement(
    executionPlanPreview.providerProfileRequirements,
    "tool_adapter_preview_stage",
  );
  const confirmationGate = findExecutionPlanGate(
    executionPlanPreview.confirmationAuditGates,
    "gate_confirmation_before_tool",
  );
  const auditGate = findExecutionPlanGate(executionPlanPreview.confirmationAuditGates, "gate_audit_projection");

  return [
    {
      prerequisiteId: "runtime_executor_implementation",
      label: "Workflow executor",
      labelMessage: { key: "runtime_executor_implementation_label" },
      area: "executor",
      status: "blocked",
      currentEvidence: "execution plan preview has stage order but no executor binding",
      currentEvidenceMessage: { key: "runtime_executor_implementation_currentEvidence" },
      missingPrerequisite: runtimeReason.missingPrerequisite,
      missingPrerequisiteMessage: runtimeReason.missingPrerequisiteMessage,
      sourceRefs: ["stage_model_reasoning", "gate_runtime_execution", runtimeReason.reasonId],
      summary: "A future runtime must provide workflow, node, and tool executor ownership before any run can start.",
      summaryMessage: { key: "runtime_executor_implementation_summary" },
    },
    {
      prerequisiteId: "provider_profile_binding",
      label: "Provider profile binding",
      labelMessage: { key: "provider_profile_binding_label" },
      area: "provider",
      status: toolRequirement.status === "blocked" ? "blocked" : "review_required",
      currentEvidence: providerRequirement.providerProfileRef,
      missingPrerequisite: toolRequirement.missingPrerequisite,
      missingPrerequisiteMessage: toolRequirement.missingPrerequisiteMessage,
      sourceRefs: [providerRequirement.requirementId, toolRequirement.requirementId, "stage_tool_preview"],
      summary: "Model and tool provider requirements are visible, but no runtime provider binding or tool adapter call is enabled.",
      summaryMessage: { key: "provider_profile_binding_summary" },
    },
    {
      prerequisiteId: "confirmation_decision_store",
      label: "Confirmation decision store",
      labelMessage: { key: "confirmation_decision_store_label" },
      area: "confirmation",
      status: "blocked",
      currentEvidence: confirmationGate.gateId,
      missingPrerequisite: "confirmation decision store and execution unlock implementation gate",
      missingPrerequisiteMessage: { key: "confirmation_decision_store_missingPrerequisite" },
      sourceRefs: [confirmationGate.gateId, "policy_confirmation_profile", "blocked_confirmation_decision"],
      summary: "Human review metadata exists, but decisions cannot be submitted, persisted, or used to unlock execution.",
      summaryMessage: { key: "confirmation_decision_store_summary" },
    },
    {
      prerequisiteId: "durable_run_result_store",
      label: "Durable run and result store",
      labelMessage: { key: "durable_run_result_store_label" },
      area: "store",
      status: "blocked",
      currentEvidence: "run detail and execution plan preview are offline projections",
      currentEvidenceMessage: { key: "durable_run_result_store_currentEvidence" },
      missingPrerequisite: "durable run store, durable result store, and materialized result reader implementation gate",
      missingPrerequisiteMessage: { key: "durable_run_result_store_missingPrerequisite" },
      sourceRefs: ["stage_audit_projection", "gate_runtime_execution", replayReason.reasonId],
      summary: "Runtime start, result materialization, and replay need durable run/result storage before they can be enabled.",
      summaryMessage: { key: "durable_run_result_store_summary" },
    },
    {
      prerequisiteId: "audit_policy_projection",
      label: "Audit policy",
      labelMessage: { key: "audit_policy_projection_label" },
      area: "audit",
      status: "review_required",
      currentEvidence: auditGate.gateId,
      missingPrerequisite: "durable audit store and raw payload redaction policy",
      missingPrerequisiteMessage: { key: "audit_policy_projection_missingPrerequisite" },
      sourceRefs: [auditGate.gateId, "stage_audit_projection"],
      summary: "Request and audit refs are visible, but audit policy remains metadata until durable stores and redaction evidence exist.",
      summaryMessage: { key: "audit_policy_projection_summary" },
    },
    {
      prerequisiteId: "writeback_policy_gate",
      label: "Business writeback policy",
      labelMessage: { key: "writeback_policy_gate_label" },
      area: "writeback",
      status: "blocked",
      currentEvidence: writebackReason.reasonId,
      missingPrerequisite: writebackReason.missingPrerequisite,
      missingPrerequisiteMessage: writebackReason.missingPrerequisiteMessage,
      sourceRefs: [writebackReason.reasonId, "gate_confirmation_before_tool"],
      summary: "Candidate actions stay advisory until writeback policy, confirmation, and upstream ownership gates are implemented.",
      summaryMessage: { key: "writeback_policy_gate_summary" },
    },
    {
      prerequisiteId: "replay_policy_gate",
      label: "Replay policy",
      labelMessage: { key: "replay_policy_gate_label" },
      area: "replay",
      status: "blocked",
      currentEvidence: replayReason.reasonId,
      missingPrerequisite: replayReason.missingPrerequisite,
      missingPrerequisiteMessage: replayReason.missingPrerequisiteMessage,
      sourceRefs: [replayReason.reasonId, "gate_runtime_execution"],
      summary: "Replay and resume remain unavailable without durable run history, executor ownership, and replay policy evidence.",
      summaryMessage: { key: "replay_policy_gate_summary" },
    },
    {
      prerequisiteId: "auth_db_repository_gate",
      label: "Auth, database, and repository gate",
      labelMessage: { key: "auth_db_repository_gate_label" },
      area: "auth_store",
      status: "blocked",
      currentEvidence: "control-plane-read-implementation-trigger-review-v1 has no satisfied trigger",
      currentEvidenceMessage: { key: "auth_db_repository_gate_currentEvidence" },
      missingPrerequisite: "Radish auth, database schema artifact, store selector, and repository adapter gates",
      missingPrerequisiteMessage: { key: "auth_db_repository_gate_missingPrerequisite" },
      sourceRefs: [
        "control-plane-read-production-auth-readiness-v1",
        "control-plane-read-schema-artifact-manifest-readiness-v1",
        "control-plane-read-implementation-trigger-review-v1",
      ],
      summary: "Future live runtime work must not bypass the existing auth/db/repository implementation trigger review.",
      summaryMessage: { key: "auth_db_repository_gate_summary" },
    },
    {
      prerequisiteId: "workflow_publish_lifecycle_gate",
      label: "Workflow publish lifecycle",
      labelMessage: { key: "workflow_publish_lifecycle_gate_label" },
      area: "publish",
      status: "blocked",
      currentEvidence: publishReason.reasonId,
      missingPrerequisite: publishReason.missingPrerequisite,
      missingPrerequisiteMessage: publishReason.missingPrerequisiteMessage,
      sourceRefs: [publishReason.reasonId, "workflow-execution-plan-preview-offline-draft"],
      summary: "The selected draft can be inspected locally, but version publish and lifecycle mutation remain disabled.",
      summaryMessage: { key: "workflow_publish_lifecycle_gate_summary" },
    },
  ];
}

function buildReadinessBlockers(
  runtimePrerequisites: WorkflowRuntimeReadinessPrerequisite[],
  executionPlanPreview: WorkflowExecutionPlanPreviewViewModel,
): WorkflowRuntimeReadinessBlocker[] {
  return runtimePrerequisites
    .filter((prerequisite) => prerequisite.status === "blocked")
    .map((prerequisite) => ({
      blockerId: `blocker_${prerequisite.prerequisiteId}`,
      label: prerequisite.label,
      labelMessage: prerequisite.labelMessage,
      area: prerequisite.area,
      severity: prerequisite.area === "executor" || prerequisite.area === "auth_store" ? "critical" : "high",
      status: "blocked",
      sourceRef: prerequisite.sourceRefs[0] ?? prerequisite.prerequisiteId,
      missingPrerequisite: prerequisite.missingPrerequisite,
      missingPrerequisiteMessage: prerequisite.missingPrerequisiteMessage,
      summary: prerequisite.summary,
      summaryMessage: prerequisite.summaryMessage,
      auditRef: auditRefForPrerequisite(prerequisite, executionPlanPreview),
    }));
}

function buildImplementationGates(
  runtimePrerequisites: WorkflowRuntimeReadinessPrerequisite[],
): WorkflowRuntimeReadinessGate[] {
  return [
    {
      gateId: "gate_runtime_implementation_trigger_review",
      label: "Implementation trigger review",
      labelMessage: { key: "gate_runtime_implementation_trigger_review_label" },
      gateKind: "implementation_trigger",
      status: "blocked",
      requiredBefore: "live_backend",
      evidenceRefs: ["control-plane-read-implementation-trigger-review-v1", "workflow-execution-plan-preview-offline-v1"],
      summary: "Current implementation trigger review remains not satisfied, so live backend and durable store work stay blocked.",
      summaryMessage: { key: "gate_runtime_implementation_trigger_review_summary" },
    },
    {
      gateId: "gate_executor_before_runtime_start",
      label: "Executor before runtime start",
      labelMessage: { key: "gate_executor_before_runtime_start_label" },
      gateKind: "runtime",
      status: statusFor(runtimePrerequisites, "runtime_executor_implementation"),
      requiredBefore: "runtime_start",
      evidenceRefs: ["runtime_executor_implementation", "durable_run_result_store"],
      summary: "Workflow runtime start requires executor ownership plus durable run/result storage.",
      summaryMessage: { key: "gate_executor_before_runtime_start_summary" },
    },
    {
      gateId: "gate_confirmation_store_before_tool_execution",
      label: "Confirmation store before tool execution",
      labelMessage: { key: "gate_confirmation_store_before_tool_execution_label" },
      gateKind: "store",
      status: statusFor(runtimePrerequisites, "confirmation_decision_store"),
      requiredBefore: "runtime_start",
      evidenceRefs: ["confirmation_decision_store", "provider_profile_binding"],
      summary: "Tool execution cannot be enabled until confirmation decisions can be stored and audited.",
      summaryMessage: { key: "gate_confirmation_store_before_tool_execution_summary" },
    },
    {
      gateId: "gate_auth_repository_before_live_backend",
      label: "Auth and repository before live backend",
      labelMessage: { key: "gate_auth_repository_before_live_backend_label" },
      gateKind: "auth",
      status: statusFor(runtimePrerequisites, "auth_db_repository_gate"),
      requiredBefore: "live_backend",
      evidenceRefs: ["auth_db_repository_gate"],
      summary: "Live backend mode must wait for Radish auth, database schema, store selector, and repository adapter gates.",
      summaryMessage: { key: "gate_auth_repository_before_live_backend_summary" },
    },
    {
      gateId: "gate_writeback_replay_policy_before_unblock",
      label: "Writeback and replay policy before unblock",
      labelMessage: { key: "gate_writeback_replay_policy_before_unblock_label" },
      gateKind: "policy",
      status: "blocked",
      requiredBefore: "writeback",
      evidenceRefs: ["writeback_policy_gate", "replay_policy_gate", "audit_policy_projection"],
      summary: "Writeback and replay need separate policy, confirmation, audit, and upstream ownership evidence before either can open.",
      summaryMessage: { key: "gate_writeback_replay_policy_before_unblock_summary" },
    },
  ];
}

function findBlockedReason(
  blockedReasons: WorkflowExecutionPlanBlockedReason[],
  reasonId: string,
): WorkflowExecutionPlanBlockedReason {
  return (
    blockedReasons.find((reason) => reason.reasonId === reasonId) ?? {
      reasonId,
      label: reasonId,
      blockedCapability: "runtime",
      status: "blocked",
      missingPrerequisite: "future implementation gate",
      missingPrerequisiteMessage: { key: "findBlockedReason_missingPrerequisite" },
      summary: "Execution plan blocker is not present in the source preview.",
      summaryMessage: { key: "findBlockedReason_summary" },
      auditRef: `audit_runtime_readiness_${reasonId}`,
    }
  );
}

function findProviderRequirement(
  requirements: WorkflowExecutionPlanProviderRequirement[],
  requirementId: string,
): WorkflowExecutionPlanProviderRequirement {
  return (
    requirements.find((requirement) => requirement.requirementId === requirementId) ?? {
      requirementId,
      label: requirementId,
      providerProfileRef: "provider:unknown",
      nodeIds: [],
      status: "blocked",
      missingPrerequisite: "future provider implementation gate",
      missingPrerequisiteMessage: { key: "findProviderRequirement_missingPrerequisite" },
      summary: "Provider requirement is not present in the source preview.",
      summaryMessage: { key: "findProviderRequirement_summary" },
    }
  );
}

function findExecutionPlanGate(gates: WorkflowExecutionPlanGate[], gateId: string): WorkflowExecutionPlanGate {
  return (
    gates.find((gate) => gate.gateId === gateId) ?? {
      gateId,
      label: gateId,
      gateKind: "runtime",
      status: "blocked",
      requiredBeforeStageId: "stage_model_reasoning",
      summary: "Execution plan gate is not present in the source preview.",
      summaryMessage: { key: "missingSourceGate" },
      auditRef: `audit_runtime_readiness_${gateId}`,
    }
  );
}

function statusFor(
  runtimePrerequisites: WorkflowRuntimeReadinessPrerequisite[],
  prerequisiteId: string,
): WorkflowRuntimeReadinessStatus {
  return (
    runtimePrerequisites.find((prerequisite) => prerequisite.prerequisiteId === prerequisiteId)?.status ?? "blocked"
  );
}

function auditRefForPrerequisite(
  prerequisite: WorkflowRuntimeReadinessPrerequisite,
  executionPlanPreview: WorkflowExecutionPlanPreviewViewModel,
): string {
  const blockedReason = executionPlanPreview.blockedPlanReasons.find((reason) =>
    prerequisite.sourceRefs.includes(reason.reasonId),
  );
  return blockedReason?.auditRef ?? `audit_runtime_readiness_${prerequisite.prerequisiteId}`;
}
