import type { WorkflowSavedDraftConsumerState } from "./savedWorkflowDraftConsumer.ts";
import type {
  WorkflowDraftDesignerDraft,
  WorkflowDraftDesignerEdge,
  WorkflowDraftDesignerNode,
} from "./workflowDraftDesigner.ts";

const DEV_EXECUTOR_SOURCE = "dev-workflow-executor-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const DEFAULT_WORKSPACE_ID = "workspace_demo";
const DEFAULT_TENANT_REF = "tenant_demo";
const DEFAULT_SUBJECT_REF = "subject_demo_user";
const DEFAULT_SCOPES = "workflow_drafts:read,workflow_runs:execute,workflow_runs:read";
const EXECUTOR_V0_MAX_NODES = 16;
const EXECUTOR_V0_MAX_EDGES = 32;
const EXECUTOR_V0_MAX_LLM_NODES = 4;

export type WorkflowExecutorConsumerMode = "disabled" | "dev_workflow_executor_http";

export type WorkflowExecutorConsumerConfig = {
  mode: WorkflowExecutorConsumerMode;
  baseUrl: string;
  workspaceId: string;
  tenantRef: string;
  subjectRef: string;
};

export type WorkflowExecutorConsumerStatus =
  | "disabled"
  | "idle"
  | "starting"
  | "reading"
  | "succeeded"
  | "failed";

export type WorkflowRunNodeRecord = {
  nodeId: string;
  nodeType: string;
  label: string;
  status: "pending" | "running" | "succeeded" | "skipped" | "failed";
  startedAt: string;
  completedAt: string;
  durationMs: number;
  predecessorNodeIds: string[];
  providerRef: string;
  outputPreview: string;
  failureCode: string;
};

export type WorkflowRunRecord = {
  schemaVersion: "workflow_run_record.v0";
  runId: string;
  draftId: string;
  draftVersion: number;
  workspaceId: string;
  applicationId: string;
  status: "running" | "succeeded" | "failed" | "canceled";
  failureCode: string;
  failureSummary: string;
  startedAt: string;
  completedAt: string;
  inputBytes: number;
  conditionNodeIds: string[];
  requestedModel: string;
  selectedProvider: string;
  selectedProfile: string;
  selectedModel: string;
  upstreamModel: string;
  selectionSource: string;
  nodes: WorkflowRunNodeRecord[];
  output: string;
  requestId: string;
  auditRef: string;
  sideEffects: {
    providerCalls: number;
    toolCalls: number;
    confirmationCalls: number;
    businessWrites: number;
    replayWrites: number;
  };
};

export type WorkflowExecutorConsumerState = {
  status: WorkflowExecutorConsumerStatus;
  mode: WorkflowExecutorConsumerMode;
  summary: string;
  failureCode: string | null;
  failureSummary: string;
  requestId: string;
  auditRef: string;
  record: WorkflowRunRecord | null;
};

export type WorkflowExecutorEligibilityReason = {
  code: string;
  summary: string;
};

export type WorkflowExecutorEligibility = {
  eligible: boolean;
  savedDraftVersion: number;
  conditionNodeIds: string[];
  reasons: WorkflowExecutorEligibilityReason[];
};

type WorkflowRunEnvelopeDocument = {
  request_id: string;
  workspace_id: string;
  application_id: string;
  run: WorkflowRunRecordDocument | null;
  failure_code: string | null;
  failure_summary: string;
  audit_ref: string;
};

type WorkflowRunRecordDocument = {
  schema_version: "workflow_run_record.v0";
  run_id: string;
  draft_id: string;
  draft_version: number;
  workspace_id: string;
  application_id: string;
  status: "running" | "succeeded" | "failed" | "canceled";
  failure_code: string;
  failure_summary: string;
  started_at: string;
  completed_at: string;
  input_bytes: number;
  condition_node_ids: string[];
  requested_model: string;
  selected_provider: string;
  selected_profile: string;
  selected_model: string;
  upstream_model: string;
  selection_source: string;
  nodes: WorkflowRunNodeRecordDocument[];
  output: string;
  request_id: string;
  audit_ref: string;
  side_effects: WorkflowRunSideEffectsDocument;
};

type WorkflowRunNodeRecordDocument = {
  node_id: string;
  node_type: string;
  label: string;
  status: "pending" | "running" | "succeeded" | "skipped" | "failed";
  started_at: string;
  completed_at: string;
  duration_ms: number;
  predecessor_node_ids: string[];
  provider_ref: string;
  output_preview: string;
  failure_code: string;
};

type WorkflowRunSideEffectsDocument = {
  provider_calls: number;
  tool_calls: number;
  confirmation_calls: number;
  business_writes: number;
  replay_writes: number;
};

export function readWorkflowExecutorConsumerConfig(): WorkflowExecutorConsumerConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  const source = env.VITE_RADISHMIND_WORKFLOW_EXECUTOR_SOURCE?.trim();
  return {
    mode: source === DEV_EXECUTOR_SOURCE ? "dev_workflow_executor_http" : "disabled",
    baseUrl: normalizeBaseUrl(
      env.VITE_RADISHMIND_WORKFLOW_EXECUTOR_BASE_URL ??
        env.VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_BASE_URL ??
        env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ??
        DEFAULT_BASE_URL,
    ),
    workspaceId:
      env.VITE_RADISHMIND_WORKFLOW_EXECUTOR_WORKSPACE_ID?.trim() ||
      env.VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_WORKSPACE_ID?.trim() ||
      DEFAULT_WORKSPACE_ID,
    tenantRef: env.VITE_RADISHMIND_DEV_READ_TENANT_REF?.trim() || DEFAULT_TENANT_REF,
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || DEFAULT_SUBJECT_REF,
  };
}

export function initialWorkflowExecutorConsumerState(
  config: WorkflowExecutorConsumerConfig,
): WorkflowExecutorConsumerState {
  if (config.mode !== "dev_workflow_executor_http") {
    return {
      status: "disabled",
      mode: "disabled",
      summary: "Workflow Executor v0 stays disabled until its explicit dev consumer source is enabled.",
      failureCode: null,
      failureSummary: "",
      requestId: "workflow-executor-disabled",
      auditRef: "audit_workflow_executor_disabled",
      record: null,
    };
  }
  return {
    status: "idle",
    mode: "dev_workflow_executor_http",
    summary: "Save an eligible executor v0 draft before starting a development run.",
    failureCode: null,
    failureSummary: "",
    requestId: "workflow-executor-idle",
    auditRef: "audit_workflow_executor_idle",
    record: null,
  };
}

export function buildWorkflowExecutorV0Draft(
  source: WorkflowDraftDesignerDraft,
  draftNumber: number,
): WorkflowDraftDesignerDraft {
  const numberLabel = String(Math.max(1, draftNumber)).padStart(2, "0");
  const applicationKey = safeWorkflowExecutorKey(source.applicationRef, 42);
  const draftId = `draft_${applicationKey}_executor_v0_${numberLabel}`;
  const providerRef = source.providerProfileRef.trim() || "provider:mock";
  const nodes: WorkflowDraftDesignerNode[] = [
    {
      nodeId: "node_executor_prompt",
      nodeType: "prompt",
      label: "Prepare advisory prompt",
      lane: "context",
      readiness: "ready",
      inputSummary: "Use the bounded user input as advisory workflow context.",
      outputSummary: "Bounded prompt packet for the model node.",
      providerRef: "",
      toolRef: "",
      ragRef: "",
      inputContractFields: ["input_text"],
      outputContractFields: ["prompt_context"],
      outputMappingSummary: "Map the bounded run input into prompt context without retaining raw input in the run record.",
      riskLevel: "low",
      requiresConfirmation: false,
      previewOnlyReason: "Executor v0 accepts bounded input without tool, RAG, confirmation, or writeback access.",
    },
    {
      nodeId: "node_executor_model",
      nodeType: "llm",
      label: "Generate advisory response",
      lane: "model",
      readiness: "ready",
      inputSummary: "Generate a reviewable advisory answer from the active prompt packet.",
      outputSummary: "Advisory model response without an external action.",
      providerRef,
      toolRef: "",
      ragRef: "",
      inputContractFields: ["prompt_context"],
      outputContractFields: ["answer_summary"],
      outputMappingSummary: "Map the Gateway response into the final advisory answer field.",
      riskLevel: "low",
      requiresConfirmation: false,
      previewOnlyReason: "The node may call the configured Gateway only after the saved draft passes executor v0 checks.",
    },
    {
      nodeId: "node_executor_output",
      nodeType: "output",
      label: "Return advisory output",
      lane: "output",
      readiness: "ready",
      inputSummary: "Return the active model response for user review.",
      outputSummary: "Bounded advisory workflow output.",
      providerRef: "",
      toolRef: "",
      ragRef: "",
      inputContractFields: ["answer_summary"],
      outputContractFields: ["answer_summary"],
      outputMappingSummary: "Expose the bounded advisory answer without writing business truth.",
      riskLevel: "low",
      requiresConfirmation: false,
      previewOnlyReason: "Output is stored only in the bounded development run record.",
    },
  ];
  const edges: WorkflowDraftDesignerEdge[] = [
    {
      edgeId: "edge_executor_prompt_model",
      fromNodeId: "node_executor_prompt",
      toNodeId: "node_executor_model",
      edgeKind: "context",
      conditionSummary: "Bounded prompt context flows to the configured Gateway model node.",
    },
    {
      edgeId: "edge_executor_model_output",
      fromNodeId: "node_executor_model",
      toNodeId: "node_executor_output",
      edgeKind: "context",
      conditionSummary: "The advisory model response flows to the bounded output node.",
    },
  ];
  return {
    ...source,
    draftId,
    templateRef: source.draftId,
    label: `${source.label} executor v0 ${numberLabel}`,
    summary:
      "Executor v0 draft for a bounded Prompt → LLM → Output run through the existing Gateway, with tool, RAG, confirmation, writeback, replay, and production enablement disabled.",
    providerProfileRef: providerRef,
    nodes,
    edges,
    designerLayout: {
      source: "workflow_node_designer",
      persistence: "ui_only",
      nodePositions: [
        { nodeId: "node_executor_prompt", x: 0, y: 0 },
        { nodeId: "node_executor_model", x: 320, y: 0 },
        { nodeId: "node_executor_output", x: 640, y: 0 },
      ],
    },
    readiness: [
      {
        checkId: "executor_v0_graph",
        label: "Executor v0 graph",
        status: "ready",
        summary: "The draft contains one Prompt, one LLM, one Output, and an acyclic path.",
      },
      {
        checkId: "executor_v0_gateway",
        label: "Gateway boundary",
        status: "ready",
        summary: "The model node reuses the configured Gateway without a second provider adapter.",
      },
      {
        checkId: "executor_v0_side_effects",
        label: "Side-effect boundary",
        status: "ready",
        summary: "Tool, RAG, confirmation, business write, replay, and durable run storage remain disabled.",
      },
    ],
    risks: [
      {
        riskId: "executor_v0_advisory_output",
        label: "Advisory output",
        riskLevel: "low",
        requiresConfirmation: false,
        summary: "The run returns reviewable text and cannot apply an external action.",
      },
    ],
    blockedCapabilities: [
      workflowExecutorBlockedCapability(draftId, "tool_executor", "Tool execution"),
      workflowExecutorBlockedCapability(draftId, "confirmation_commit", "Confirmation commit"),
      workflowExecutorBlockedCapability(draftId, "business_writeback", "Business writeback"),
      workflowExecutorBlockedCapability(draftId, "run_replay", "Run replay / resume"),
    ],
    routeMetadata: {
      ...source.routeMetadata,
      requestId: `req_${draftId}`,
      auditRef: `audit_${draftId}`,
    },
    localOnlyInteraction: "local_edit",
    executionProfile: "executor_v0",
  };
}

export function evaluateWorkflowExecutorEligibility(
  draft: WorkflowDraftDesignerDraft,
  savedDraftState: WorkflowSavedDraftConsumerState,
  draftEditDirty: boolean,
): WorkflowExecutorEligibility {
  const graphEligibility = evaluateWorkflowExecutorGraphEligibility(draft);
  const reasons: WorkflowExecutorEligibilityReason[] = [...graphEligibility.reasons];
  if (savedDraftState.currentDraftVersion <= 0 ||
    !["saved_dev_record", "validation_ready"].includes(savedDraftState.status)) {
    reasons.push({
      code: "saved_draft_version_unavailable",
      summary: "The active executor draft must be saved before it can run.",
    });
  }
  if (draftEditDirty) {
    reasons.push({
      code: "unsaved_local_changes",
      summary: "Save the current local changes so the executor reads the exact visible draft version.",
    });
  }

  return {
    eligible: reasons.length === 0,
    savedDraftVersion: savedDraftState.currentDraftVersion,
    conditionNodeIds: graphEligibility.conditionNodeIds,
    reasons,
  };
}

export function evaluateWorkflowExecutorGraphEligibility(
  draft: WorkflowDraftDesignerDraft,
): Pick<WorkflowExecutorEligibility, "eligible" | "conditionNodeIds" | "reasons"> {
  const reasons: WorkflowExecutorEligibilityReason[] = [];
  if (draft.executionProfile !== "executor_v0") {
    reasons.push({
      code: "executor_profile_missing",
      summary: "Create a dedicated executor v0 draft instead of running the current review-only graph.",
    });
  }
  reasons.push(...workflowExecutorGraphReasons(draft));
  return {
    eligible: reasons.length === 0,
    conditionNodeIds: draft.nodes
      .filter((node) => node.nodeType === "condition")
      .map((node) => node.nodeId),
    reasons,
  };
}

export async function startWorkflowRunDevRecord(
  draft: WorkflowDraftDesignerDraft,
  inputText: string,
  conditionValues: Record<string, boolean>,
  config: WorkflowExecutorConsumerConfig,
  options: { model?: string; temperature?: number } = {},
): Promise<WorkflowExecutorConsumerState> {
  const envelope = await requestWorkflowRunEnvelope(
    `/v1/user-workspace/workflow-drafts/${encodeURIComponent(draft.draftId)}/runs`,
    draft.applicationRef,
    `dev-workflow-run-${draft.draftId}`,
    config,
    {
      method: "POST",
      body: JSON.stringify({
        workspace_id: config.workspaceId,
        application_id: draft.applicationRef,
        input_text: inputText,
        condition_values: conditionValues,
        model: options.model?.trim() ?? "",
        ...(options.temperature === undefined ? {} : { temperature: options.temperature }),
      }),
    },
  );
  return workflowExecutorStateFromEnvelope(envelope, "start");
}

export async function readWorkflowRunDevRecord(
  record: WorkflowRunRecord,
  config: WorkflowExecutorConsumerConfig,
): Promise<WorkflowExecutorConsumerState> {
  const query = new URLSearchParams({
    workspace_id: config.workspaceId,
    application_id: record.applicationId,
  });
  const envelope = await requestWorkflowRunEnvelope(
    `/v1/user-workspace/workflow-runs/${encodeURIComponent(record.runId)}?${query.toString()}`,
    record.applicationId,
    `dev-workflow-run-read-${record.runId}`,
    config,
    { method: "GET" },
  );
  return workflowExecutorStateFromEnvelope(envelope, "read");
}

function workflowExecutorGraphReasons(
  draft: WorkflowDraftDesignerDraft,
): WorkflowExecutorEligibilityReason[] {
  const reasons: WorkflowExecutorEligibilityReason[] = [];
  if (draft.nodes.length < 3 || draft.nodes.length > EXECUTOR_V0_MAX_NODES ||
    draft.edges.length < 2 || draft.edges.length > EXECUTOR_V0_MAX_EDGES) {
    reasons.push({
      code: "executor_graph_budget",
      summary: "Executor v0 accepts 3–16 nodes and 2–32 edges.",
    });
  }
  const nodesById = new Map<string, WorkflowDraftDesignerNode>();
  for (const node of draft.nodes) {
    if (!node.nodeId.trim() || nodesById.has(node.nodeId)) {
      reasons.push({ code: "executor_node_id_invalid", summary: "Every executor node needs a unique non-empty id." });
      continue;
    }
    nodesById.set(node.nodeId, node);
    if (!["prompt", "llm", "condition", "output"].includes(node.nodeType)) {
      reasons.push({
        code: "executor_node_type_blocked",
        summary: `Node ${node.nodeId} uses blocked type ${node.nodeType}.`,
      });
    }
    if (node.toolRef.trim() || node.ragRef.trim() || node.requiresConfirmation || node.riskLevel !== "low") {
      reasons.push({
        code: "executor_node_risk_blocked",
        summary: `Node ${node.nodeId} has a tool, RAG, confirmation, or non-low-risk boundary.`,
      });
    }
  }
  const promptCount = draft.nodes.filter((node) => node.nodeType === "prompt").length;
  const outputCount = draft.nodes.filter((node) => node.nodeType === "output").length;
  const llmCount = draft.nodes.filter((node) => node.nodeType === "llm").length;
  if (promptCount !== 1 || outputCount !== 1 || llmCount < 1 || llmCount > EXECUTOR_V0_MAX_LLM_NODES) {
    reasons.push({
      code: "executor_node_roles_invalid",
      summary: "Executor v0 requires one Prompt, one Output, and one to four LLM nodes.",
    });
  }
  reasons.push(...workflowExecutorEdgeAndReachabilityReasons(draft, nodesById));
  return uniqueWorkflowExecutorReasons(reasons);
}

function workflowExecutorEdgeAndReachabilityReasons(
  draft: WorkflowDraftDesignerDraft,
  nodesById: Map<string, WorkflowDraftDesignerNode>,
): WorkflowExecutorEligibilityReason[] {
  const reasons: WorkflowExecutorEligibilityReason[] = [];
  const indegree = new Map<string, number>();
  const outgoing = new Map<string, WorkflowDraftDesignerEdge[]>();
  const incoming = new Map<string, WorkflowDraftDesignerEdge[]>();
  const edgeIds = new Set<string>();
  const edgePairs = new Set<string>();
  for (const nodeId of nodesById.keys()) {
    indegree.set(nodeId, 0);
  }
  for (const edge of draft.edges) {
    const pair = `${edge.fromNodeId}\u0000${edge.toNodeId}`;
    if (!edge.edgeId.trim() || edgeIds.has(edge.edgeId) || edgePairs.has(pair) ||
      edge.fromNodeId === edge.toNodeId || !nodesById.has(edge.fromNodeId) || !nodesById.has(edge.toNodeId)) {
      reasons.push({ code: "executor_edge_invalid", summary: "Executor graph contains a duplicate or invalid edge." });
      continue;
    }
    edgeIds.add(edge.edgeId);
    edgePairs.add(pair);
    const sourceNode = nodesById.get(edge.fromNodeId)!;
    const conditionRoute = workflowExecutorConditionRoute(edge.conditionSummary);
    if (sourceNode.nodeType === "condition" && conditionRoute === null) {
      reasons.push({
        code: "executor_condition_route_invalid",
        summary: `Condition edge ${edge.edgeId} must use when:true, when:false, or always.`,
      });
    }
    if (sourceNode.nodeType !== "condition" && conditionRoute !== null) {
      reasons.push({
        code: "executor_condition_source_invalid",
        summary: `Edge ${edge.edgeId} uses conditional routing without a condition source node.`,
      });
    }
    indegree.set(edge.toNodeId, (indegree.get(edge.toNodeId) ?? 0) + 1);
    outgoing.set(edge.fromNodeId, [...(outgoing.get(edge.fromNodeId) ?? []), edge]);
    incoming.set(edge.toNodeId, [...(incoming.get(edge.toNodeId) ?? []), edge]);
  }
  if (reasons.some((reason) => reason.code === "executor_edge_invalid")) {
    return reasons;
  }
  const roots = [...nodesById.keys()].filter((nodeId) => (indegree.get(nodeId) ?? 0) === 0);
  const terminals = [...nodesById.keys()].filter((nodeId) => (outgoing.get(nodeId) ?? []).length === 0);
  if (roots.length !== 1 || nodesById.get(roots[0] ?? "")?.nodeType !== "prompt" ||
    terminals.length !== 1 || nodesById.get(terminals[0] ?? "")?.nodeType !== "output") {
    reasons.push({
      code: "executor_root_terminal_invalid",
      summary: "Executor graph must have one Prompt root and one Output terminal.",
    });
    return reasons;
  }
  const order = workflowExecutorTopologicalOrder(indegree, outgoing, draft.nodes.map((node) => node.nodeId));
  if (order.length !== nodesById.size) {
    reasons.push({ code: "executor_cycle", summary: "Executor v0 does not allow graph cycles." });
    return reasons;
  }
  if (workflowExecutorReachableCount(roots[0]!, outgoing, "toNodeId") !== nodesById.size ||
    workflowExecutorReachableCount(terminals[0]!, incoming, "fromNodeId") !== nodesById.size) {
    reasons.push({
      code: "executor_unreachable_node",
      summary: "Every executor node must be on a path from Prompt to Output.",
    });
  }
  return reasons;
}

function workflowExecutorTopologicalOrder(
  indegree: Map<string, number>,
  outgoing: Map<string, WorkflowDraftDesignerEdge[]>,
  nodeOrder: string[],
): string[] {
  const position = new Map(nodeOrder.map((nodeId, index) => [nodeId, index]));
  const remaining = new Map(indegree);
  const ready = nodeOrder.filter((nodeId) => (remaining.get(nodeId) ?? 0) === 0);
  const order: string[] = [];
  while (ready.length > 0) {
    ready.sort((left, right) => (position.get(left) ?? 0) - (position.get(right) ?? 0));
    const nodeId = ready.shift()!;
    order.push(nodeId);
    for (const edge of outgoing.get(nodeId) ?? []) {
      const nextDegree = (remaining.get(edge.toNodeId) ?? 0) - 1;
      remaining.set(edge.toNodeId, nextDegree);
      if (nextDegree === 0) {
        ready.push(edge.toNodeId);
      }
    }
  }
  return order;
}

function workflowExecutorReachableCount(
  startNodeId: string,
  edgesByNode: Map<string, WorkflowDraftDesignerEdge[]>,
  targetField: "toNodeId" | "fromNodeId",
): number {
  const visited = new Set([startNodeId]);
  const queue = [startNodeId];
  while (queue.length > 0) {
    const nodeId = queue.shift()!;
    for (const edge of edgesByNode.get(nodeId) ?? []) {
      const target = edge[targetField];
      if (!visited.has(target)) {
        visited.add(target);
        queue.push(target);
      }
    }
  }
  return visited.size;
}

function workflowExecutorConditionRoute(summary: string): "true" | "false" | "always" | null {
  switch (summary.trim().toLowerCase()) {
    case "when:true":
      return "true";
    case "when:false":
      return "false";
    case "always":
      return "always";
    default:
      return null;
  }
}

function uniqueWorkflowExecutorReasons(
  reasons: WorkflowExecutorEligibilityReason[],
): WorkflowExecutorEligibilityReason[] {
  const seen = new Set<string>();
  return reasons.filter((reason) => {
    const key = `${reason.code}\u0000${reason.summary}`;
    if (seen.has(key)) {
      return false;
    }
    seen.add(key);
    return true;
  });
}

function workflowExecutorBlockedCapability(
  draftId: string,
  capabilityId: string,
  label: string,
): WorkflowDraftDesignerDraft["blockedCapabilities"][number] {
  return {
    capabilityId,
    label,
    status: "blocked",
    missingPrerequisite: "independent high-risk workflow feature design",
    summary: `${label} remains outside Workflow Executor v0.`,
    auditRef: `audit_${draftId}_${capabilityId}_blocked`,
  };
}

async function requestWorkflowRunEnvelope(
  path: string,
  applicationRef: string,
  requestId: string,
  config: WorkflowExecutorConsumerConfig,
  init: RequestInit,
): Promise<WorkflowRunEnvelopeDocument> {
  if (config.mode !== "dev_workflow_executor_http") {
    throw new Error("workflow executor dev HTTP source is disabled");
  }
  const response = await fetch(`${config.baseUrl}${path}`, {
    ...init,
    headers: workflowExecutorHeaders(config, applicationRef, requestId),
  });
  const body: unknown = await response.json();
  if (!response.ok) {
    throw new Error(`workflow executor route returned HTTP ${response.status}`);
  }
  if (!isWorkflowRunEnvelopeDocument(body)) {
    throw new Error("workflow executor route returned an unexpected envelope");
  }
  return body;
}

function workflowExecutorHeaders(
  config: WorkflowExecutorConsumerConfig,
  applicationRef: string,
  requestId: string,
): HeadersInit {
  return {
    Accept: "application/json",
    "Content-Type": "application/json",
    "X-Request-Id": requestId,
    "X-RadishMind-Dev-Read-Identity": "dev-workflow-executor-consumer",
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Read-Scopes": DEFAULT_SCOPES,
    "X-RadishMind-Dev-Read-Audit": "audit_dev_workflow_executor_consumer",
    "X-RadishMind-Dev-Workflow-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Workflow-Application": applicationRef,
  };
}

function workflowExecutorStateFromEnvelope(
  envelope: WorkflowRunEnvelopeDocument,
  operation: "start" | "read",
): WorkflowExecutorConsumerState {
  const record = envelope.run ? workflowRunRecordFromDocument(envelope.run) : null;
  if (envelope.failure_code || !record) {
    return {
      status: "failed",
      mode: "dev_workflow_executor_http",
      summary: envelope.failure_summary || `Workflow run ${operation} failed.`,
      failureCode: envelope.failure_code ?? record?.failureCode ?? "workflow_run_unavailable",
      failureSummary: envelope.failure_summary || record?.failureSummary || "Workflow run record is unavailable.",
      requestId: envelope.request_id,
      auditRef: envelope.audit_ref,
      record,
    };
  }
  if (record.status !== "succeeded") {
    return {
      status: "failed",
      mode: "dev_workflow_executor_http",
      summary: record.failureSummary || `Workflow run ended with ${record.status}.`,
      failureCode: record.failureCode || `workflow_run_${record.status}`,
      failureSummary: record.failureSummary,
      requestId: envelope.request_id,
      auditRef: envelope.audit_ref,
      record,
    };
  }
  return {
    status: "succeeded",
    mode: "dev_workflow_executor_http",
    summary: operation === "read"
      ? `Workflow run ${record.runId} was reloaded from the scoped development run store.`
      : `Workflow run ${record.runId} completed with bounded advisory output.`,
    failureCode: null,
    failureSummary: "",
    requestId: envelope.request_id,
    auditRef: envelope.audit_ref,
    record,
  };
}

function workflowRunRecordFromDocument(document: WorkflowRunRecordDocument): WorkflowRunRecord {
  return {
    schemaVersion: document.schema_version,
    runId: document.run_id,
    draftId: document.draft_id,
    draftVersion: document.draft_version,
    workspaceId: document.workspace_id,
    applicationId: document.application_id,
    status: document.status,
    failureCode: document.failure_code,
    failureSummary: document.failure_summary,
    startedAt: document.started_at,
    completedAt: document.completed_at,
    inputBytes: document.input_bytes,
    conditionNodeIds: [...document.condition_node_ids],
    requestedModel: document.requested_model,
    selectedProvider: document.selected_provider,
    selectedProfile: document.selected_profile,
    selectedModel: document.selected_model,
    upstreamModel: document.upstream_model,
    selectionSource: document.selection_source,
    nodes: document.nodes.map((node) => ({
      nodeId: node.node_id,
      nodeType: node.node_type,
      label: node.label,
      status: node.status,
      startedAt: node.started_at,
      completedAt: node.completed_at,
      durationMs: node.duration_ms,
      predecessorNodeIds: [...node.predecessor_node_ids],
      providerRef: node.provider_ref,
      outputPreview: node.output_preview,
      failureCode: node.failure_code,
    })),
    output: document.output,
    requestId: document.request_id,
    auditRef: document.audit_ref,
    sideEffects: {
      providerCalls: document.side_effects.provider_calls,
      toolCalls: document.side_effects.tool_calls,
      confirmationCalls: document.side_effects.confirmation_calls,
      businessWrites: document.side_effects.business_writes,
      replayWrites: document.side_effects.replay_writes,
    },
  };
}

function isWorkflowRunEnvelopeDocument(value: unknown): value is WorkflowRunEnvelopeDocument {
  if (!value || typeof value !== "object") {
    return false;
  }
  const candidate = value as Partial<WorkflowRunEnvelopeDocument>;
  return typeof candidate.request_id === "string" &&
    typeof candidate.workspace_id === "string" &&
    typeof candidate.application_id === "string" &&
    (candidate.failure_code === null || typeof candidate.failure_code === "string") &&
    typeof candidate.failure_summary === "string" &&
    typeof candidate.audit_ref === "string" &&
    (candidate.run === null || isWorkflowRunRecordDocument(candidate.run));
}

function isWorkflowRunRecordDocument(value: unknown): value is WorkflowRunRecordDocument {
  if (!value || typeof value !== "object") {
    return false;
  }
  const candidate = value as Partial<WorkflowRunRecordDocument>;
  return candidate.schema_version === "workflow_run_record.v0" &&
    typeof candidate.run_id === "string" &&
    typeof candidate.draft_id === "string" &&
    typeof candidate.draft_version === "number" &&
    typeof candidate.workspace_id === "string" &&
    typeof candidate.application_id === "string" &&
    isWorkflowRunStatus(candidate.status) &&
    typeof candidate.failure_code === "string" &&
    typeof candidate.failure_summary === "string" &&
    typeof candidate.started_at === "string" &&
    typeof candidate.completed_at === "string" &&
    typeof candidate.input_bytes === "number" &&
    isStringArray(candidate.condition_node_ids) &&
    typeof candidate.requested_model === "string" &&
    typeof candidate.selected_provider === "string" &&
    typeof candidate.selected_profile === "string" &&
    typeof candidate.selected_model === "string" &&
    typeof candidate.upstream_model === "string" &&
    typeof candidate.selection_source === "string" &&
    Array.isArray(candidate.nodes) && candidate.nodes.every(isWorkflowRunNodeRecordDocument) &&
    typeof candidate.output === "string" &&
    typeof candidate.request_id === "string" &&
    typeof candidate.audit_ref === "string" &&
    isWorkflowRunSideEffectsDocument(candidate.side_effects);
}

function isWorkflowRunNodeRecordDocument(value: unknown): value is WorkflowRunNodeRecordDocument {
  if (!value || typeof value !== "object") {
    return false;
  }
  const candidate = value as Partial<WorkflowRunNodeRecordDocument>;
  return typeof candidate.node_id === "string" &&
    typeof candidate.node_type === "string" &&
    typeof candidate.label === "string" &&
    isWorkflowRunNodeStatus(candidate.status) &&
    typeof candidate.started_at === "string" &&
    typeof candidate.completed_at === "string" &&
    typeof candidate.duration_ms === "number" &&
    isStringArray(candidate.predecessor_node_ids) &&
    typeof candidate.provider_ref === "string" &&
    typeof candidate.output_preview === "string" &&
    typeof candidate.failure_code === "string";
}

function isWorkflowRunSideEffectsDocument(value: unknown): value is WorkflowRunSideEffectsDocument {
  if (!value || typeof value !== "object") {
    return false;
  }
  const candidate = value as Partial<WorkflowRunSideEffectsDocument>;
  return [
    candidate.provider_calls,
    candidate.tool_calls,
    candidate.confirmation_calls,
    candidate.business_writes,
    candidate.replay_writes,
  ].every((count) => typeof count === "number" && Number.isInteger(count) && count >= 0);
}

function isWorkflowRunStatus(value: unknown): value is WorkflowRunRecordDocument["status"] {
  return value === "running" || value === "succeeded" || value === "failed" || value === "canceled";
}

function isWorkflowRunNodeStatus(value: unknown): value is WorkflowRunNodeRecordDocument["status"] {
  return value === "pending" || value === "running" || value === "succeeded" || value === "skipped" || value === "failed";
}

function isStringArray(value: unknown): value is string[] {
  return Array.isArray(value) && value.every((item) => typeof item === "string");
}

function normalizeBaseUrl(value: string): string {
  return value.trim().replace(/\/+$/, "");
}

function safeWorkflowExecutorKey(value: string, maxLength: number): string {
  const normalized = value.toLowerCase().replace(/[^a-z0-9]+/g, "_").replace(/^_+|_+$/g, "");
  return (normalized || "workflow").slice(0, maxLength);
}
