import {
  CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS,
  controlPlaneReadResponseHasForbiddenOutput,
} from "../../../../../contracts/typescript/control-plane-read-api";
import type {
  WorkflowSurfaceOverviewSource,
  WorkflowSurfaceOverviewStatus,
} from "./workflowSurfaceOverview";

export type WorkflowScenarioStatus = WorkflowSurfaceOverviewStatus;

export type WorkflowScenarioKind =
  | "radishflow_candidate_action"
  | "radish_docs_answer"
  | "workflow_advisory_review";

export type WorkflowScenarioInputField = {
  fieldId: string;
  label: string;
  sourceRef: string;
  required: boolean;
  summary: string;
};

export type WorkflowScenarioExpectedOutput = {
  outputId: string;
  label: string;
  status: WorkflowScenarioStatus;
  summary: string;
};

export type WorkflowScenario = {
  scenarioId: string;
  label: string;
  scenarioKind: WorkflowScenarioKind;
  applicationRef: string;
  workflowDefinitionId: string;
  draftId: string;
  runId: string;
  intent: string;
  triggerSummary: string;
  inputContract: WorkflowScenarioInputField[];
  expectedOutputs: WorkflowScenarioExpectedOutput[];
  riskLevel: "low" | "medium" | "high";
  requiresConfirmation: boolean;
  validationFocus: string;
  sourceRefs: string[];
};

export type WorkflowScenarioSummary = {
  label: string;
  value: string;
  status: WorkflowScenarioStatus;
  summary: string;
};

export type WorkflowScenarioRelation = {
  relationId: string;
  label: string;
  sourceRef: string;
  targetRef: string;
  status: WorkflowScenarioStatus;
  summary: string;
  auditRef: string;
};

export type WorkflowScenarioBlockedReason = {
  reasonId: string;
  label: string;
  sourceSurface: "application" | "definition" | "run" | "draft" | "plan" | "readiness";
  status: "blocked";
  missingPrerequisite: string;
  summary: string;
  auditRef: string;
};

export type WorkflowScenarioStopLine = {
  stopLineId: string;
  label: string;
  status: "locked";
  summary: string;
};

export type WorkflowScenarioInspectorViewModel = {
  pageId: "workflow-scenario-inspector-offline";
  sourcePageIds: string[];
  scenarioMode: "offline_read_only_advisory";
  applicationId: string;
  workflowDefinitionId: string;
  selectedDraftId: string;
  latestRunId: string;
  selectedScenarioId: string;
  scenarios: WorkflowScenario[];
  selectedScenario: WorkflowScenario;
  summary: WorkflowScenarioSummary[];
  relationMap: WorkflowScenarioRelation[];
  blockedReasons: WorkflowScenarioBlockedReason[];
  stopLines: WorkflowScenarioStopLine[];
  forbiddenProjectionBlocked: boolean;
  canRenderScenarioInspector: boolean;
  canInspectScenarioLocally: true;
  canSwitchScenarioLocally: true;
  canRequestLiveBackend: false;
  canPersistScenario: false;
  canMutateBuilder: false;
  canPublishWorkflow: false;
  canStartRuntime: false;
  canSubmitConfirmationDecision: false;
  canWriteBusinessTruth: false;
  canReplayRun: false;
};

export function buildWorkflowScenarioInspectorViewModel(
  source: WorkflowSurfaceOverviewSource,
  selectedScenarioId: string | null = null,
): WorkflowScenarioInspectorViewModel {
  const scenarios = buildScenarios(source);
  const selectedScenario =
    scenarios.find((scenario) => scenario.scenarioId === selectedScenarioId) ?? scenarios[0]!;
  const blockedReasons = buildBlockedReasons(source, selectedScenario);
  const relationMap = buildRelationMap(source, selectedScenario, blockedReasons);
  const forbiddenProjectionBlocked = controlPlaneReadResponseHasForbiddenOutput({
    scenario: { [CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS[7]]: "blocked" },
  });

  return {
    pageId: "workflow-scenario-inspector-offline",
    sourcePageIds: [
      source.applicationDetail.pageId,
      source.definitionDetail.pageId,
      source.runDetail.pageId,
      "workflow-draft-designer-offline",
      source.validationInspector.pageId,
      source.executionPlanPreview.pageId,
      source.runtimeReadinessInspector.pageId,
    ],
    scenarioMode: "offline_read_only_advisory",
    applicationId: source.applicationDetail.applicationId,
    workflowDefinitionId: source.definitionDetail.workflowDefinitionId,
    selectedDraftId: source.selectedDraft.draftId,
    latestRunId: source.runDetail.runId,
    selectedScenarioId: selectedScenario.scenarioId,
    scenarios,
    selectedScenario,
    summary: buildSummary(source, selectedScenario, blockedReasons),
    relationMap,
    blockedReasons,
    stopLines: buildStopLines(),
    forbiddenProjectionBlocked,
    canRenderScenarioInspector:
      source.applicationDetail.canRenderApplicationDetail &&
      source.definitionDetail.canRenderDefinitionDetail &&
      source.runDetail.canRenderRunDetail &&
      source.validationInspector.canRenderDraftValidationInspector &&
      source.executionPlanPreview.canRenderExecutionPlanPreview &&
      source.runtimeReadinessInspector.canRenderRuntimeReadinessInspector &&
      scenarios.length >= 2 &&
      selectedScenario.inputContract.length >= 4 &&
      selectedScenario.expectedOutputs.length >= 4 &&
      relationMap.length >= 5 &&
      blockedReasons.length >= 4,
    canInspectScenarioLocally: true,
    canSwitchScenarioLocally: true,
    canRequestLiveBackend: false,
    canPersistScenario: false,
    canMutateBuilder: false,
    canPublishWorkflow: false,
    canStartRuntime: false,
    canSubmitConfirmationDecision: false,
    canWriteBusinessTruth: false,
    canReplayRun: false,
  };
}

function buildScenarios(source: WorkflowSurfaceOverviewSource): WorkflowScenario[] {
  if (source.applicationDetail.applicationId === "app_docs_assistant") {
    return buildDocsScenarios(source);
  }
  return buildRadishFlowScenarios(source);
}

function buildRadishFlowScenarios(source: WorkflowSurfaceOverviewSource): WorkflowScenario[] {
  return [
    {
      scenarioId: "scenario_radishflow_reconnect_stream_outlet",
      label: "Reconnect stream outlet",
      scenarioKind: "radishflow_candidate_action",
      applicationRef: source.applicationDetail.applicationId,
      workflowDefinitionId: source.definitionDetail.workflowDefinitionId,
      draftId: source.selectedDraft.draftId,
      runId: source.runDetail.runId,
      intent:
        "Explain a RadishFlow diagnostic gap and prepare a candidate stream reconnect action for human review.",
      triggerSummary:
        "A selected outlet stream is disconnected after a simulation diagnostic review and needs an advisory repair candidate.",
      inputContract: [
        buildInput("tenant_ref", "Tenant", source.applicationDetail.tenantRef, "Tenant scope for the advisory surface."),
        buildInput(
          "application_ref",
          "Application",
          source.applicationDetail.applicationId,
          "Current workflow application selected in the workspace.",
        ),
        buildInput(
          "selection_summary",
          "Selection",
          "radishflow.selected_stream.outlet",
          "Selected flowsheet object, stream endpoint, and local diagnostic focus.",
        ),
        buildInput(
          "diagnostic_summary",
          "Diagnostic",
          source.runDetail.failureCode,
          "Sanitized run or diagnostic summary used to explain why a reconnect candidate is suggested.",
        ),
        buildInput(
          "action_policy",
          "Action policy",
          source.definitionDetail.blockedActionPreview.toolRef,
          "Candidate action schema and confirmation requirement remain visible but blocked.",
        ),
      ],
      expectedOutputs: [
        buildOutput("answer_summary", "Answer summary", "review_required", "A short explanation of the disconnected outlet and candidate repair path."),
        buildOutput("candidate_action", "Candidate action", "blocked", "A structured reconnect suggestion that cannot execute from this UI."),
        buildOutput("risk_summary", "Risk summary", "blocked", "Medium risk action remains behind confirmation and executor gates."),
        buildOutput("audit_refs", "Audit refs", "offline_only", "Request, run, and policy audit refs stay visible for review."),
      ],
      riskLevel: "medium",
      requiresConfirmation: true,
      validationFocus: "Policy gate, action preview shape, confirmation requirement, writeback boundary.",
      sourceRefs: [
        source.applicationDetail.applicationId,
        source.definitionDetail.workflowDefinitionId,
        source.selectedDraft.draftId,
        source.runDetail.runId,
      ],
    },
    {
      scenarioId: "scenario_radishflow_cross_object_warning",
      label: "Cross-object warning review",
      scenarioKind: "workflow_advisory_review",
      applicationRef: source.applicationDetail.applicationId,
      workflowDefinitionId: source.definitionDetail.workflowDefinitionId,
      draftId: source.selectedDraft.draftId,
      runId: source.runDetail.runId,
      intent:
        "Review a cross-object warning and keep the response ordered by evidence, risk, and allowed candidate scope.",
      triggerSummary:
        "A user selects more than one flowsheet object and needs a single primary actionable target with warnings.",
      inputContract: [
        buildInput("tenant_ref", "Tenant", source.applicationDetail.tenantRef, "Tenant scope for the advisory surface."),
        buildInput(
          "application_ref",
          "Application",
          source.applicationDetail.applicationId,
          "Current workflow application selected in the workspace.",
        ),
        buildInput(
          "selection_summary",
          "Selection",
          "radishflow.multi_object_selection",
          "Selected unit and stream summaries used to choose the primary focus.",
        ),
        buildInput(
          "warning_policy",
          "Warning policy",
          "cross_object_warning_citation_ordering",
          "Risk and citation ordering constraints used before previewing any candidate action.",
        ),
        buildInput(
          "audit_context",
          "Audit context",
          source.runDetail.traceId,
          "Trace id and sanitized audit refs for later review.",
        ),
      ],
      expectedOutputs: [
        buildOutput("answer_summary", "Answer summary", "review_required", "A ranked explanation of the warning and primary target."),
        buildOutput("candidate_action", "Candidate action", "blocked", "At most one advisory candidate action remains preview-only."),
        buildOutput("risk_summary", "Risk summary", "blocked", "Risk labels keep execution and writeback unavailable."),
        buildOutput("audit_refs", "Audit refs", "offline_only", "Audit refs stay attached to the read-only warning review."),
      ],
      riskLevel: "medium",
      requiresConfirmation: true,
      validationFocus: "Primary target selection, citation order, policy gate, and blocked action preview.",
      sourceRefs: [
        source.applicationDetail.applicationId,
        source.definitionDetail.workflowDefinitionId,
        source.selectedDraft.draftId,
        source.runDetail.runId,
      ],
    },
  ];
}

function buildDocsScenarios(source: WorkflowSurfaceOverviewSource): WorkflowScenario[] {
  return [
    {
      scenarioId: "scenario_radish_docs_evidence_gap",
      label: "Docs evidence gap",
      scenarioKind: "radish_docs_answer",
      applicationRef: source.applicationDetail.applicationId,
      workflowDefinitionId: source.definitionDetail.workflowDefinitionId,
      draftId: source.selectedDraft.draftId,
      runId: source.runDetail.runId,
      intent:
        "Answer a product documentation question while marking missing evidence instead of inventing a confident answer.",
      triggerSummary:
        "A docs question has partial evidence, so the workflow should produce an advisory answer with explicit gap markers.",
      inputContract: [
        buildInput("tenant_ref", "Tenant", source.applicationDetail.tenantRef, "Tenant scope for the advisory surface."),
        buildInput(
          "application_ref",
          "Application",
          source.applicationDetail.applicationId,
          "Current docs assistant application selected in the workspace.",
        ),
        buildInput(
          "question",
          "Question",
          "radish.docs.question.evidence_gap",
          "Sanitized user question and document selection summary.",
        ),
        buildInput(
          "citation_policy",
          "Citation policy",
          "docs_citation_required",
          "Citation and evidence gap policy used before rendering an answer.",
        ),
        buildInput(
          "retrieval_summary",
          "Retrieval summary",
          source.runDetail.failureCode,
          "Read store status and sanitized document context for local inspection.",
        ),
      ],
      expectedOutputs: [
        buildOutput("answer_summary", "Answer summary", "review_required", "A concise advisory answer with visible uncertainty."),
        buildOutput("evidence_gap", "Evidence gap", "review_required", "Missing evidence markers stay explicit beside the answer."),
        buildOutput("citation_summary", "Citation summary", "offline_only", "Citation refs are shown as sanitized metadata only."),
        buildOutput("publish_preview", "Publish preview", "blocked", "Publishing or writing answer state remains unavailable."),
      ],
      riskLevel: "low",
      requiresConfirmation: false,
      validationFocus: "Question context, evidence gap markers, citation summary, and blocked publish/writeback path.",
      sourceRefs: [
        source.applicationDetail.applicationId,
        source.definitionDetail.workflowDefinitionId,
        source.selectedDraft.draftId,
        source.runDetail.runId,
      ],
    },
    {
      scenarioId: "scenario_radish_docs_faq_forum_conflict",
      label: "FAQ and forum conflict",
      scenarioKind: "radish_docs_answer",
      applicationRef: source.applicationDetail.applicationId,
      workflowDefinitionId: source.definitionDetail.workflowDefinitionId,
      draftId: source.selectedDraft.draftId,
      runId: source.runDetail.runId,
      intent:
        "Resolve conflicting FAQ and forum signals by explaining the conflict and keeping the answer advisory.",
      triggerSummary:
        "The available docs and forum summaries disagree, so the response should identify the conflict rather than hide it.",
      inputContract: [
        buildInput("tenant_ref", "Tenant", source.applicationDetail.tenantRef, "Tenant scope for the advisory surface."),
        buildInput(
          "application_ref",
          "Application",
          source.applicationDetail.applicationId,
          "Current docs assistant application selected in the workspace.",
        ),
        buildInput(
          "question",
          "Question",
          "radish.docs.question.faq_forum_conflict",
          "Sanitized user question with conflicting source summaries.",
        ),
        buildInput(
          "conflict_policy",
          "Conflict policy",
          "docs_faq_forum_conflict_review",
          "Policy requiring the response to surface conflict instead of selecting a hidden winner.",
        ),
        buildInput(
          "audit_context",
          "Audit context",
          source.runDetail.traceId,
          "Trace id and sanitized audit refs for local review.",
        ),
      ],
      expectedOutputs: [
        buildOutput("answer_summary", "Answer summary", "review_required", "A direct answer framed with conflict context."),
        buildOutput("conflict_summary", "Conflict summary", "review_required", "FAQ/forum disagreement is surfaced for the user."),
        buildOutput("citation_summary", "Citation summary", "offline_only", "Citation refs remain sanitized and read-only."),
        buildOutput("publish_preview", "Publish preview", "blocked", "Publishing or writing answer state remains unavailable."),
      ],
      riskLevel: "low",
      requiresConfirmation: false,
      validationFocus: "Conflict handling, citation visibility, advisory answer boundary, and blocked publish/writeback path.",
      sourceRefs: [
        source.applicationDetail.applicationId,
        source.definitionDetail.workflowDefinitionId,
        source.selectedDraft.draftId,
        source.runDetail.runId,
      ],
    },
  ];
}

function buildSummary(
  source: WorkflowSurfaceOverviewSource,
  scenario: WorkflowScenario,
  blockedReasons: WorkflowScenarioBlockedReason[],
): WorkflowScenarioSummary[] {
  return [
    {
      label: "Scenario",
      value: scenario.label,
      status: "offline_only",
      summary: scenario.intent,
    },
    {
      label: "Application",
      value: source.applicationDetail.displayName,
      status: source.applicationDetail.canRenderApplicationDetail ? "ready" : "blocked",
      summary: `${scenario.applicationRef} uses ${source.applicationDetail.providerProfileRef}.`,
    },
    {
      label: "Validation",
      value: source.validationInspector.validationStatus,
      status: validationStatusToScenarioStatus(source.validationInspector.validationStatus),
      summary: scenario.validationFocus,
    },
    {
      label: "Runtime",
      value: `${blockedReasons.length} blockers`,
      status: "blocked",
      summary: "Scenario remains explainable, but runtime execution, persistence, writeback, and replay stay unavailable.",
    },
  ];
}

function buildRelationMap(
  source: WorkflowSurfaceOverviewSource,
  scenario: WorkflowScenario,
  blockedReasons: WorkflowScenarioBlockedReason[],
): WorkflowScenarioRelation[] {
  return [
    {
      relationId: "scenario_to_application",
      label: "Scenario to application",
      sourceRef: scenario.scenarioId,
      targetRef: scenario.applicationRef,
      status: scenario.applicationRef === source.applicationDetail.applicationId ? "ready" : "review_required",
      summary: "The scenario follows the currently selected application context.",
      auditRef: source.applicationDetail.auditRef,
    },
    {
      relationId: "scenario_to_definition",
      label: "Scenario to definition",
      sourceRef: scenario.scenarioId,
      targetRef: scenario.workflowDefinitionId,
      status: scenario.workflowDefinitionId === source.definitionDetail.workflowDefinitionId ? "ready" : "review_required",
      summary: "Definition detail supplies the node, edge, risk, and blocked action shape.",
      auditRef: source.definitionDetail.auditRef,
    },
    {
      relationId: "scenario_to_draft",
      label: "Scenario to draft",
      sourceRef: scenario.scenarioId,
      targetRef: scenario.draftId,
      status: "offline_only",
      summary: "The selected offline draft supplies input, output, and validation evidence for the scenario.",
      auditRef: source.selectedDraft.routeMetadata.auditRef,
    },
    {
      relationId: "scenario_to_plan",
      label: "Scenario to plan",
      sourceRef: scenario.scenarioId,
      targetRef: source.executionPlanPreview.draftRouteId,
      status: source.executionPlanPreview.canRenderExecutionPlanPreview ? "offline_only" : "blocked",
      summary: "Execution plan preview explains future stage order without creating a persistent plan.",
      auditRef: source.executionPlanPreview.auditRef,
    },
    {
      relationId: "scenario_to_runtime_readiness",
      label: "Scenario to runtime readiness",
      sourceRef: scenario.scenarioId,
      targetRef: `${blockedReasons.length} blocked reasons`,
      status: "blocked",
      summary: "Readiness blockers explain why the scenario cannot start, confirm, write back, or replay.",
      auditRef: source.runtimeReadinessInspector.auditRef,
    },
  ];
}

function buildBlockedReasons(
  source: WorkflowSurfaceOverviewSource,
  scenario: WorkflowScenario,
): WorkflowScenarioBlockedReason[] {
  const applicationReason = source.applicationDetail.blockedCapabilities[0];
  const draftReason = source.selectedDraft.blockedCapabilities[0];
  const planReason = source.executionPlanPreview.blockedPlanReasons.find((reason) =>
    scenario.requiresConfirmation ? reason.reasonId === "blocked_runtime" : reason.reasonId === "blocked_publish",
  ) ?? source.executionPlanPreview.blockedPlanReasons[0];
  const readinessReasons = source.runtimeReadinessInspector.readinessBlockers.slice(0, 3);

  return [
    applicationReason
      ? {
          reasonId: `application_${applicationReason.capabilityId}`,
          label: applicationReason.label,
          sourceSurface: "application",
          status: "blocked",
          missingPrerequisite: applicationReason.missingPrerequisite,
          summary: applicationReason.reason,
          auditRef: applicationReason.auditRef,
        }
      : undefined,
    draftReason
      ? {
          reasonId: `draft_${draftReason.capabilityId}`,
          label: draftReason.label,
          sourceSurface: "draft",
          status: "blocked",
          missingPrerequisite: draftReason.missingPrerequisite,
          summary: draftReason.summary,
          auditRef: draftReason.auditRef,
        }
      : undefined,
    planReason
      ? {
          reasonId: `plan_${planReason.reasonId}`,
          label: planReason.label,
          sourceSurface: "plan",
          status: "blocked",
          missingPrerequisite: planReason.missingPrerequisite,
          summary: planReason.summary,
          auditRef: planReason.auditRef,
        }
      : undefined,
    ...readinessReasons.map((reason) => ({
      reasonId: `readiness_${reason.blockerId}`,
      label: reason.label,
      sourceSurface: "readiness" as const,
      status: "blocked" as const,
      missingPrerequisite: reason.missingPrerequisite,
      summary: reason.summary,
      auditRef: reason.auditRef,
    })),
    {
      reasonId: `run_${source.runDetail.blockedReplayPreview.guardId}`,
      label: source.runDetail.blockedReplayPreview.label,
      sourceSurface: "run",
      status: "blocked",
      missingPrerequisite: source.runDetail.blockedReplayPreview.missingPrerequisite,
      summary: source.runDetail.blockedReplayPreview.reason,
      auditRef: source.runDetail.blockedReplayPreview.auditRef,
    },
  ].filter((reason): reason is WorkflowScenarioBlockedReason => reason !== undefined);
}

function buildStopLines(): WorkflowScenarioStopLine[] {
  return [
    {
      stopLineId: "offline_scenario_only",
      label: "Offline scenario only",
      status: "locked",
      summary: "Scenario selection uses fixture-derived view models and never requests a live backend.",
    },
    {
      stopLineId: "no_scenario_persistence",
      label: "No persistence",
      status: "locked",
      summary: "Scenario selection changes browser view state only and does not save drafts, plans, readiness, or results.",
    },
    {
      stopLineId: "no_runtime_start",
      label: "No runtime start",
      status: "locked",
      summary: "The inspector explains intent and blockers; it cannot start workflow, node, tool, or agent execution.",
    },
    {
      stopLineId: "no_confirmation_writeback_replay",
      label: "No confirmation, writeback, replay",
      status: "locked",
      summary: "No approve, reject, defer, execution unlock, business writeback, replay, or resume control exists.",
    },
  ];
}

function buildInput(
  fieldId: string,
  label: string,
  sourceRef: string,
  summary: string,
): WorkflowScenarioInputField {
  return {
    fieldId,
    label,
    sourceRef,
    required: true,
    summary,
  };
}

function buildOutput(
  outputId: string,
  label: string,
  status: WorkflowScenarioStatus,
  summary: string,
): WorkflowScenarioExpectedOutput {
  return {
    outputId,
    label,
    status,
    summary,
  };
}

function validationStatusToScenarioStatus(status: string): WorkflowScenarioStatus {
  if (status === "passed") {
    return "ready";
  }
  if (status === "blocked") {
    return "blocked";
  }
  return "review_required";
}
