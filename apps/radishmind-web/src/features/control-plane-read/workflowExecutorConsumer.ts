import type { WorkflowSavedDraftConsumerState } from "./savedWorkflowDraftConsumer.ts";
import type {
  WorkflowDraftDesignerDraft,
  WorkflowDraftDesignerEdge,
  WorkflowDraftDesignerNode,
} from "./workflowDraftDesigner.ts";
import {
  parseWorkflowRunRecordDocument,
  type WorkflowRunRecord,
} from "./workflowRunRecordConsumer.ts";

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
  diagnosticsDevEnabled?: boolean;
};

export type WorkflowExecutorConsumerStatus =
  | "disabled"
  | "idle"
  | "starting"
  | "reading"
  | "succeeded"
  | "failed";

export type WorkflowRunDevFailureScenario = "gateway_timeout" | "gateway_queue_full" | "gateway_worker_crash" |
  "gateway_protocol_failure" | "provider_failed" | "output_unavailable" | "request_canceled" |
  "run_store_unavailable" | "terminal_write_conflict" | "budget_exceeded" | "stale_running";

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
  run: unknown | null;
  failure_code: string | null;
  failure_summary: string;
  audit_ref: string;
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
    diagnosticsDevEnabled: env.VITE_RADISHMIND_WORKFLOW_DIAGNOSTICS_DEV?.trim() === "true",
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
  return readWorkflowRunDevRecordByID(record.runId, record.applicationId, config);
}

export async function startWorkflowDiagnosticDevRecord(
  draftId: string,
  applicationId: string,
  scenario: WorkflowRunDevFailureScenario,
  config: WorkflowExecutorConsumerConfig,
): Promise<WorkflowExecutorConsumerState> {
  const envelope = await requestWorkflowRunEnvelope(
    `/v1/user-workspace/workflow-drafts/${encodeURIComponent(draftId.trim())}/runs`,
    applicationId,
    `dev-workflow-diagnostic-${scenario}`,
    config,
    {
      method: "POST",
      body: JSON.stringify({
        workspace_id: config.workspaceId,
        application_id: applicationId,
        input_text: "Review the deterministic workflow diagnostics path.",
        condition_values: {},
        model: "",
        dev_failure_scenario: scenario,
      }),
    },
  );
  return workflowExecutorStateFromEnvelope(envelope, "start");
}

export async function readWorkflowRunDevRecordByID(
  runId: string,
  applicationId: string,
  config: WorkflowExecutorConsumerConfig,
): Promise<WorkflowExecutorConsumerState> {
  const query = new URLSearchParams({
    workspace_id: config.workspaceId,
    application_id: applicationId,
  });
  const envelope = await requestWorkflowRunEnvelope(
    `/v1/user-workspace/workflow-runs/${encodeURIComponent(runId)}?${query.toString()}`,
    applicationId,
    `dev-workflow-run-read-${runId}`,
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
  const record = envelope.run ? parseWorkflowRunRecordDocument(envelope.run) : null;
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
    (candidate.run === null || parseWorkflowRunRecordDocument(candidate.run) !== null);
}

function normalizeBaseUrl(value: string): string {
  return value.trim().replace(/\/+$/, "");
}

function safeWorkflowExecutorKey(value: string, maxLength: number): string {
  const normalized = value.toLowerCase().replace(/[^a-z0-9]+/g, "_").replace(/^_+|_+$/g, "");
  return (normalized || "workflow").slice(0, maxLength);
}
