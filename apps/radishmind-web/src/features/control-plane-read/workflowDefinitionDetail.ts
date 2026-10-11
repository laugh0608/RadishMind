import type { WorkflowProjectionCopy } from "./workflowProjectionCopy.ts";
import {
  CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS,
  CONTROL_PLANE_READ_ROUTE_DEFINITIONS,
  CONTROL_PLANE_READ_ROUTES,
  controlPlaneReadResponseHasForbiddenOutput,
  type WorkflowDefinitionSummary,
} from "../../../../../contracts/typescript/control-plane-read-api";

export type WorkflowDefinitionDetailNode = WorkflowProjectionCopy & {
  nodeId: string;
  label: string;
  nodeType: "prompt" | "rag_retrieval" | "llm" | "http_tool" | "condition" | "output";
  inputSummary: string;
  outputSummary: string;
  riskLevel: "low" | "medium" | "high";
  requiresConfirmation: boolean;
};

export type WorkflowDefinitionDetailEdge = WorkflowProjectionCopy & {
  edgeId: string;
  fromNodeId: string;
  toNodeId: string;
  conditionSummary: string;
};

export type WorkflowDefinitionDetailSchemaSummary = WorkflowProjectionCopy & {
  label: string;
  summary: string;
  fields: string[];
};

export type WorkflowDefinitionBlockedActionPreview = WorkflowProjectionCopy & {
  toolActionId: string;
  nodeId: string;
  toolRef: string;
  actionKind: string;
  riskLevel: "low" | "medium" | "high";
  requiresConfirmation: true;
  policyReason: string;
  blockedState: "blocked_executor_not_available";
  auditRef: string;
};

export type WorkflowDefinitionDetailViewModel = {
  pageId: "workflow-definition-detail-read";
  sourcePageId: "workspace-workflow-definitions";
  sourceRouteId: "workflow-definition-summary-list-route";
  draftRouteId: "workflow-definition-detail-read-draft";
  routePath: typeof CONTROL_PLANE_READ_ROUTES.workflowDefinitions;
  requestId: string;
  auditRef: string;
  workflowDefinitionId: string;
  applicationRef: string;
  version: number;
  definitionStatus: string;
  riskLevel: string;
  requiresConfirmationCapable: boolean;
  nodes: WorkflowDefinitionDetailNode[];
  edges: WorkflowDefinitionDetailEdge[];
  inputSummary: WorkflowDefinitionDetailSchemaSummary;
  outputSummary: WorkflowDefinitionDetailSchemaSummary;
  blockedActionPreview: WorkflowDefinitionBlockedActionPreview;
  forbiddenProjectionBlocked: boolean;
  canRenderDefinitionDetail: boolean;
  canRequestLiveBackend: false;
  canMutate: false;
  canEditWorkflow: false;
  canRunWorkflow: false;
  canConfirmWorkflowAction: false;
  canWriteBusinessTruth: false;
  canReplayRun: false;
};

const DEFAULT_WORKFLOW_DEFINITION_SUMMARY: WorkflowDefinitionSummary = {
  workflow_definition_id: "wf_radishflow_copilot_latest",
  tenant_ref: "tenant_demo",
  application_ref: "app_flow_copilot",
  version: 3,
  definition_status: "published",
  node_count: 8,
  risk_level: "medium",
  requires_confirmation_capable: true,
  updated_at: "2026-05-31T10:25:00Z",
};

export function buildWorkflowDefinitionDetailViewModel(
  workflowDefinition: WorkflowDefinitionSummary = DEFAULT_WORKFLOW_DEFINITION_SUMMARY,
): WorkflowDefinitionDetailViewModel {
  const route = CONTROL_PLANE_READ_ROUTE_DEFINITIONS["workflow-definition-summary-list-route"];
  const routePath = CONTROL_PLANE_READ_ROUTES.workflowDefinitions;
  const nodes = buildDetailNodes(workflowDefinition.application_ref);
  const edges = buildDetailEdges(workflowDefinition.application_ref);
  const forbiddenProjectionBlocked = controlPlaneReadResponseHasForbiddenOutput({
    detail: { [CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS[7]]: "blocked" },
  });

  return {
    pageId: "workflow-definition-detail-read",
    sourcePageId: "workspace-workflow-definitions",
    sourceRouteId: "workflow-definition-summary-list-route",
    draftRouteId: "workflow-definition-detail-read-draft",
    routePath,
    requestId: "req_workflow_definition_detail_demo",
    auditRef: "audit_workflow_definition_detail_demo",
    workflowDefinitionId: workflowDefinition.workflow_definition_id,
    applicationRef: workflowDefinition.application_ref,
    version: workflowDefinition.version,
    definitionStatus: workflowDefinition.definition_status,
    riskLevel: workflowDefinition.risk_level,
    requiresConfirmationCapable: workflowDefinition.requires_confirmation_capable,
    nodes,
    edges,
    inputSummary: {
      label: "Input summary",
      labelMessage: { key: "workflowDefinitionDetail_workflow_definition_detail_read_label" },
      summary: "Flowsheet context, selected objects, diagnostics, and operator intent are read as sanitized inputs.",
      summaryMessage: { key: "workflowDefinitionDetail_workflow_definition_detail_read_summary" },
      fields: ["tenant_ref", "application_ref", "selection_summary", "diagnostic_summary"],
    },
    outputSummary: {
      label: "Output summary",
      labelMessage: { key: "workflowDefinitionDetail_workflow_definition_detail_read_label_2" },
      summary: outputSummaryFor(workflowDefinition.application_ref),
      summaryMessage: { key: workflowDefinition.application_ref === "app_docs_assistant" ? "definitionDocsOutput" : "definitionFlowOutput" },
      fields: ["answer_summary", "candidate_actions", "risk_summary", "audit_refs"],
    },
    blockedActionPreview: buildBlockedActionPreview(workflowDefinition.application_ref),
    forbiddenProjectionBlocked,
    canRenderDefinitionDetail:
      route.path === routePath &&
      route.canMutate === false &&
      nodes.length === workflowDefinition.node_count,
    canRequestLiveBackend: false,
    canMutate: false,
    canEditWorkflow: false,
    canRunWorkflow: false,
    canConfirmWorkflowAction: false,
    canWriteBusinessTruth: false,
    canReplayRun: false,
  };
}

function buildDetailNodes(applicationRef: string): WorkflowDefinitionDetailNode[] {
  if (applicationRef === "app_docs_assistant") {
    return [
      {
        nodeId: "node_docs_collect_context",
        label: "Collect docs context",
      labelMessage: { key: "workflowDefinitionDetail_node_docs_collect_context_label" },
        nodeType: "prompt",
        inputSummary: "Tenant, application, question, selected documents, and citation policy.",
      inputSummaryMessage: { key: "workflowDefinitionDetail_node_docs_collect_context_inputSummary" as const },
        outputSummary: "Sanitized retrieval prompt context for local inspection.",
      outputSummaryMessage: { key: "workflowDefinitionDetail_node_docs_collect_context_outputSummary" as const },
        riskLevel: "low",
        requiresConfirmation: false,
      },
      {
        nodeId: "node_docs_model_reasoning",
        label: "Docs model reasoning",
      labelMessage: { key: "workflowDefinitionDetail_node_docs_model_reasoning_label" },
        nodeType: "llm",
        inputSummary: "Question context, retrieval summary, and answer contract.",
      inputSummaryMessage: { key: "workflowDefinitionDetail_node_docs_model_reasoning_inputSummary" as const },
        outputSummary: "Advisory answer, evidence gap markers, and citation summary.",
      outputSummaryMessage: { key: "workflowDefinitionDetail_node_docs_model_reasoning_outputSummary" as const },
        riskLevel: "low",
        requiresConfirmation: false,
      },
      {
        nodeId: "node_docs_output_review",
        label: "Answer review",
      labelMessage: { key: "workflowDefinitionDetail_node_docs_output_review_label" },
        nodeType: "output",
        inputSummary: "Answer summary, citation markers, and policy notes.",
      inputSummaryMessage: { key: "workflowDefinitionDetail_node_docs_output_review_inputSummary" as const },
        outputSummary: "Read-only answer surface for workspace review.",
      outputSummaryMessage: { key: "workflowDefinitionDetail_node_docs_output_review_outputSummary" as const },
        riskLevel: "low",
        requiresConfirmation: false,
      },
      {
        nodeId: "node_docs_audit_projection",
        label: "Docs audit projection",
      labelMessage: { key: "workflowDefinitionDetail_node_docs_audit_projection_label" },
        nodeType: "output",
        inputSummary: "Trace id, request id, route ref, and sanitized audit refs.",
      inputSummaryMessage: { key: "workflowDefinitionDetail_node_docs_audit_projection_inputSummary" as const },
        outputSummary: "Audit metadata without raw prompt, token, or document payloads.",
      outputSummaryMessage: { key: "workflowDefinitionDetail_node_docs_audit_projection_outputSummary" as const },
        riskLevel: "low",
        requiresConfirmation: false,
      },
    ];
  }

  return [
    {
      nodeId: "node_collect_context",
      label: "Collect context",
      labelMessage: { key: "workflowDefinitionDetail_node_collect_context_label" },
      nodeType: "prompt",
      inputSummary: "Tenant, application, selected flowsheet objects, and diagnostic summary.",
      inputSummaryMessage: { key: "workflowDefinitionDetail_node_collect_context_inputSummary" as const },
      outputSummary: "Sanitized prompt context for the model call.",
      outputSummaryMessage: { key: "workflowDefinitionDetail_node_collect_context_outputSummary" as const },
      riskLevel: "low",
      requiresConfirmation: false,
    },
    {
      nodeId: "node_model_reasoning",
      label: "Model reasoning",
      labelMessage: { key: "workflowDefinitionDetail_node_model_reasoning_label" },
      nodeType: "llm",
      inputSummary: "Prompt context and canonical response contract.",
      inputSummaryMessage: { key: "workflowDefinitionDetail_node_model_reasoning_inputSummary" as const },
      outputSummary: "Advisory answer, candidate edit summary, citations, and risk markers.",
      outputSummaryMessage: { key: "workflowDefinitionDetail_node_model_reasoning_outputSummary" as const },
      riskLevel: "medium",
      requiresConfirmation: false,
    },
    {
      nodeId: "node_policy_gate",
      label: "Policy gate",
      labelMessage: { key: "workflowDefinitionDetail_node_policy_gate_label" },
      nodeType: "condition",
      inputSummary: "Candidate action shape, risk level, and writeback boundary.",
      inputSummaryMessage: { key: "workflowDefinitionDetail_node_policy_gate_inputSummary" as const },
      outputSummary: "Blocked state when confirmation or executor prerequisites are missing.",
      outputSummaryMessage: { key: "workflowDefinitionDetail_node_policy_gate_outputSummary" as const },
      riskLevel: "medium",
      requiresConfirmation: true,
    },
    {
      nodeId: "node_tool_preview",
      label: "Tool action preview",
      labelMessage: { key: "workflowDefinitionDetail_node_tool_preview_label" },
      nodeType: "http_tool",
      inputSummary: "Sanitized candidate action payload without raw tool request body.",
      inputSummaryMessage: { key: "workflowDefinitionDetail_node_tool_preview_inputSummary" as const },
      outputSummary: "Preview-only action metadata and audit reference.",
      outputSummaryMessage: { key: "workflowDefinitionDetail_node_tool_preview_outputSummary" as const },
      riskLevel: "medium",
      requiresConfirmation: true,
    },
    {
      nodeId: "node_human_confirmation_placeholder",
      label: "Confirmation placeholder",
      labelMessage: { key: "workflowDefinitionDetail_node_human_confirmation_placeholder_label" },
      nodeType: "condition",
      inputSummary: "Future confirmation decision shape and disabled reason.",
      inputSummaryMessage: { key: "workflowDefinitionDetail_node_human_confirmation_placeholder_inputSummary" as const },
      outputSummary: "No decision is submitted in this read-only surface.",
      outputSummaryMessage: { key: "workflowDefinitionDetail_node_human_confirmation_placeholder_outputSummary" as const },
      riskLevel: "high",
      requiresConfirmation: true,
    },
    {
      nodeId: "node_advisory_response",
      label: "Advisory response",
      labelMessage: { key: "workflowDefinitionDetail_node_advisory_response_label" },
      nodeType: "output",
      inputSummary: "Answer summary, blocked action preview, and risk explanation.",
      inputSummaryMessage: { key: "workflowDefinitionDetail_node_advisory_response_inputSummary" as const },
      outputSummary: "Read-only response surface for user review.",
      outputSummaryMessage: { key: "workflowDefinitionDetail_node_advisory_response_outputSummary" as const },
      riskLevel: "low",
      requiresConfirmation: false,
    },
    {
      nodeId: "node_audit_projection",
      label: "Audit projection",
      labelMessage: { key: "workflowDefinitionDetail_node_audit_projection_label" },
      nodeType: "output",
      inputSummary: "Trace id, route ref, request id, and sanitized audit refs.",
      inputSummaryMessage: { key: "workflowDefinitionDetail_node_audit_projection_inputSummary" as const },
      outputSummary: "Audit metadata without raw secret, token, prompt, or tool payload.",
      outputSummaryMessage: { key: "workflowDefinitionDetail_node_audit_projection_outputSummary" as const },
      riskLevel: "low",
      requiresConfirmation: false,
    },
    {
      nodeId: "node_stop_line_summary",
      label: "Stop-line summary",
      labelMessage: { key: "workflowDefinitionDetail_node_stop_line_summary_label" },
      nodeType: "output",
      inputSummary: "Executor, confirmation, writeback, replay, database, and auth stop lines.",
      inputSummaryMessage: { key: "workflowDefinitionDetail_node_stop_line_summary_inputSummary" as const },
      outputSummary: "Visible disabled capability summary.",
      outputSummaryMessage: { key: "workflowDefinitionDetail_node_stop_line_summary_outputSummary" as const },
      riskLevel: "low",
      requiresConfirmation: false,
    },
  ];
}

function buildDetailEdges(applicationRef: string): WorkflowDefinitionDetailEdge[] {
  if (applicationRef === "app_docs_assistant") {
    return [
      {
        edgeId: "edge_docs_context_to_model",
        fromNodeId: "node_docs_collect_context",
        toNodeId: "node_docs_model_reasoning",
        conditionSummary: "Question and document summaries are sanitized before model reasoning.",
      conditionSummaryMessage: { key: "edge_edge_docs_context_to_model" },
      },
      {
        edgeId: "edge_docs_model_to_review",
        fromNodeId: "node_docs_model_reasoning",
        toNodeId: "node_docs_output_review",
        conditionSummary: "The answer remains advisory and citation-aware before any user review.",
      conditionSummaryMessage: { key: "edge_edge_docs_model_to_review" },
      },
      {
        edgeId: "edge_docs_review_to_audit",
        fromNodeId: "node_docs_output_review",
        toNodeId: "node_docs_audit_projection",
        conditionSummary: "Sanitized answer metadata is projected into the audit surface.",
      conditionSummaryMessage: { key: "edge_edge_docs_review_to_audit" },
      },
      {
        edgeId: "edge_docs_audit_to_review",
        fromNodeId: "node_docs_audit_projection",
        toNodeId: "node_docs_output_review",
        conditionSummary: "Audit refs stay visible beside the read-only answer summary.",
      conditionSummaryMessage: { key: "edge_edge_docs_audit_to_review" },
      },
    ];
  }

  return [
    {
      edgeId: "edge_context_to_model",
      fromNodeId: "node_collect_context",
      toNodeId: "node_model_reasoning",
      conditionSummary: "Context is sanitized before model reasoning.",
      conditionSummaryMessage: { key: "edge_edge_context_to_model" },
    },
    {
      edgeId: "edge_model_to_policy",
      fromNodeId: "node_model_reasoning",
      toNodeId: "node_policy_gate",
      conditionSummary: "Candidate actions are routed through policy before preview.",
      conditionSummaryMessage: { key: "edge_edge_model_to_policy" },
    },
    {
      edgeId: "edge_policy_to_tool_preview",
      fromNodeId: "node_policy_gate",
      toNodeId: "node_tool_preview",
      conditionSummary: "Only preview metadata is rendered; tool execution stays blocked.",
      conditionSummaryMessage: { key: "edge_edge_policy_to_tool_preview" },
    },
    {
      edgeId: "edge_policy_to_confirmation_placeholder",
      fromNodeId: "node_policy_gate",
      toNodeId: "node_human_confirmation_placeholder",
      conditionSummary: "Historical high-risk actions expose an archived, read-only confirmation placeholder.",
      conditionSummaryMessage: { key: "edge_edge_policy_to_confirmation_placeholder" },
    },
    {
      edgeId: "edge_preview_to_response",
      fromNodeId: "node_tool_preview",
      toNodeId: "node_advisory_response",
      conditionSummary: "Blocked action metadata is included in the advisory output.",
      conditionSummaryMessage: { key: "edge_edge_preview_to_response" },
    },
    {
      edgeId: "edge_response_to_audit",
      fromNodeId: "node_advisory_response",
      toNodeId: "node_audit_projection",
      conditionSummary: "Sanitized audit refs remain visible for review.",
      conditionSummaryMessage: { key: "edge_edge_response_to_audit" },
    },
    {
      edgeId: "edge_audit_to_stop_line",
      fromNodeId: "node_audit_projection",
      toNodeId: "node_stop_line_summary",
      conditionSummary: "Unavailable runtime capabilities stay explicit.",
      conditionSummaryMessage: { key: "edge_edge_audit_to_stop_line" },
    },
  ];
}

function outputSummaryFor(applicationRef: string): string {
  if (applicationRef === "app_docs_assistant") {
    return "The workflow returns advisory document answers, evidence gap markers, citation summaries, and audit references.";
  }
  return "The workflow returns advisory explanations, candidate actions, risk labels, and audit references.";
}

function buildBlockedActionPreview(applicationRef: string): WorkflowDefinitionBlockedActionPreview {
  if (applicationRef === "app_docs_assistant") {
    return {
      toolActionId: "tool_action_preview_docs_answer_publish",
      nodeId: "node_docs_output_review",
      toolRef: "radish.docs.answer_publish",
      actionKind: "publish_docs_answer",
      riskLevel: "low",
      requiresConfirmation: true,
      policyReason: "Docs answers remain advisory; publishing or writing answer state requires a future policy and writeback gate.",
      policyReasonMessage: { key: "tool_action_preview_docs_answer_publish_policyReason" },
      blockedState: "blocked_executor_not_available",
      auditRef: "audit_docs_answer_publish_blocked_demo",
    };
  }
  return {
    toolActionId: "tool_action_preview_reconnect_stream",
    nodeId: "node_policy_gate",
    toolRef: "radishflow.candidate_action",
    actionKind: "suggest_flowsheet_edit",
    riskLevel: "medium",
    requiresConfirmation: true,
    policyReason: "Candidate action remains blocked until a future confirmation flow and executor are implemented.",
    policyReasonMessage: { key: "tool_action_preview_reconnect_stream_policyReason" },
    blockedState: "blocked_executor_not_available",
    auditRef: "audit_tool_action_preview_demo",
  };
}
