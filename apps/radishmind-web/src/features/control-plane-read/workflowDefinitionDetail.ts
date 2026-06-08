import {
  CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS,
  CONTROL_PLANE_READ_ROUTE_DEFINITIONS,
  CONTROL_PLANE_READ_ROUTES,
  controlPlaneReadResponseHasForbiddenOutput,
  type WorkflowDefinitionSummary,
} from "../../../../../contracts/typescript/control-plane-read-api";

export type WorkflowDefinitionDetailNode = {
  nodeId: string;
  label: string;
  nodeType: "prompt" | "llm" | "http_tool" | "condition" | "output";
  inputSummary: string;
  outputSummary: string;
  riskLevel: "low" | "medium" | "high";
  requiresConfirmation: boolean;
};

export type WorkflowDefinitionDetailEdge = {
  edgeId: string;
  fromNodeId: string;
  toNodeId: string;
  conditionSummary: string;
};

export type WorkflowDefinitionDetailSchemaSummary = {
  label: string;
  summary: string;
  fields: string[];
};

export type WorkflowDefinitionBlockedActionPreview = {
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
      summary: "Flowsheet context, selected objects, diagnostics, and operator intent are read as sanitized inputs.",
      fields: ["tenant_ref", "application_ref", "selection_summary", "diagnostic_summary"],
    },
    outputSummary: {
      label: "Output summary",
      summary: outputSummaryFor(workflowDefinition.application_ref),
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
        nodeType: "prompt",
        inputSummary: "Tenant, application, question, selected documents, and citation policy.",
        outputSummary: "Sanitized retrieval prompt context for local inspection.",
        riskLevel: "low",
        requiresConfirmation: false,
      },
      {
        nodeId: "node_docs_model_reasoning",
        label: "Docs model reasoning",
        nodeType: "llm",
        inputSummary: "Question context, retrieval summary, and answer contract.",
        outputSummary: "Advisory answer, evidence gap markers, and citation summary.",
        riskLevel: "low",
        requiresConfirmation: false,
      },
      {
        nodeId: "node_docs_output_review",
        label: "Answer review",
        nodeType: "output",
        inputSummary: "Answer summary, citation markers, and policy notes.",
        outputSummary: "Read-only answer surface for workspace review.",
        riskLevel: "low",
        requiresConfirmation: false,
      },
      {
        nodeId: "node_docs_audit_projection",
        label: "Docs audit projection",
        nodeType: "output",
        inputSummary: "Trace id, request id, route ref, and sanitized audit refs.",
        outputSummary: "Audit metadata without raw prompt, token, or document payloads.",
        riskLevel: "low",
        requiresConfirmation: false,
      },
    ];
  }

  return [
    {
      nodeId: "node_collect_context",
      label: "Collect context",
      nodeType: "prompt",
      inputSummary: "Tenant, application, selected flowsheet objects, and diagnostic summary.",
      outputSummary: "Sanitized prompt context for the model call.",
      riskLevel: "low",
      requiresConfirmation: false,
    },
    {
      nodeId: "node_model_reasoning",
      label: "Model reasoning",
      nodeType: "llm",
      inputSummary: "Prompt context and canonical response contract.",
      outputSummary: "Advisory answer, candidate edit summary, citations, and risk markers.",
      riskLevel: "medium",
      requiresConfirmation: false,
    },
    {
      nodeId: "node_policy_gate",
      label: "Policy gate",
      nodeType: "condition",
      inputSummary: "Candidate action shape, risk level, and writeback boundary.",
      outputSummary: "Blocked state when confirmation or executor prerequisites are missing.",
      riskLevel: "medium",
      requiresConfirmation: true,
    },
    {
      nodeId: "node_tool_preview",
      label: "Tool action preview",
      nodeType: "http_tool",
      inputSummary: "Sanitized candidate action payload without raw tool request body.",
      outputSummary: "Preview-only action metadata and audit reference.",
      riskLevel: "medium",
      requiresConfirmation: true,
    },
    {
      nodeId: "node_human_confirmation_placeholder",
      label: "Confirmation placeholder",
      nodeType: "condition",
      inputSummary: "Future confirmation decision shape and disabled reason.",
      outputSummary: "No decision is submitted in this read-only surface.",
      riskLevel: "high",
      requiresConfirmation: true,
    },
    {
      nodeId: "node_advisory_response",
      label: "Advisory response",
      nodeType: "output",
      inputSummary: "Answer summary, blocked action preview, and risk explanation.",
      outputSummary: "Read-only response surface for user review.",
      riskLevel: "low",
      requiresConfirmation: false,
    },
    {
      nodeId: "node_audit_projection",
      label: "Audit projection",
      nodeType: "output",
      inputSummary: "Trace id, route ref, request id, and sanitized audit refs.",
      outputSummary: "Audit metadata without raw secret, token, prompt, or tool payload.",
      riskLevel: "low",
      requiresConfirmation: false,
    },
    {
      nodeId: "node_stop_line_summary",
      label: "Stop-line summary",
      nodeType: "output",
      inputSummary: "Executor, confirmation, writeback, replay, database, and OIDC stop lines.",
      outputSummary: "Visible disabled capability summary.",
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
      },
      {
        edgeId: "edge_docs_model_to_review",
        fromNodeId: "node_docs_model_reasoning",
        toNodeId: "node_docs_output_review",
        conditionSummary: "The answer remains advisory and citation-aware before any user review.",
      },
      {
        edgeId: "edge_docs_review_to_audit",
        fromNodeId: "node_docs_output_review",
        toNodeId: "node_docs_audit_projection",
        conditionSummary: "Sanitized answer metadata is projected into the audit surface.",
      },
      {
        edgeId: "edge_docs_audit_to_review",
        fromNodeId: "node_docs_audit_projection",
        toNodeId: "node_docs_output_review",
        conditionSummary: "Audit refs stay visible beside the read-only answer summary.",
      },
    ];
  }

  return [
    {
      edgeId: "edge_context_to_model",
      fromNodeId: "node_collect_context",
      toNodeId: "node_model_reasoning",
      conditionSummary: "Context is sanitized before model reasoning.",
    },
    {
      edgeId: "edge_model_to_policy",
      fromNodeId: "node_model_reasoning",
      toNodeId: "node_policy_gate",
      conditionSummary: "Candidate actions are routed through policy before preview.",
    },
    {
      edgeId: "edge_policy_to_tool_preview",
      fromNodeId: "node_policy_gate",
      toNodeId: "node_tool_preview",
      conditionSummary: "Only preview metadata is rendered; tool execution stays blocked.",
    },
    {
      edgeId: "edge_policy_to_confirmation_placeholder",
      fromNodeId: "node_policy_gate",
      toNodeId: "node_human_confirmation_placeholder",
      conditionSummary: "High-risk actions expose a disabled confirmation placeholder.",
    },
    {
      edgeId: "edge_preview_to_response",
      fromNodeId: "node_tool_preview",
      toNodeId: "node_advisory_response",
      conditionSummary: "Blocked action metadata is included in the advisory output.",
    },
    {
      edgeId: "edge_response_to_audit",
      fromNodeId: "node_advisory_response",
      toNodeId: "node_audit_projection",
      conditionSummary: "Sanitized audit refs remain visible for review.",
    },
    {
      edgeId: "edge_audit_to_stop_line",
      fromNodeId: "node_audit_projection",
      toNodeId: "node_stop_line_summary",
      conditionSummary: "Unavailable runtime capabilities stay explicit.",
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
    blockedState: "blocked_executor_not_available",
    auditRef: "audit_tool_action_preview_demo",
  };
}
