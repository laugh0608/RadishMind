import type { WorkflowSavedDraftConsumerState } from "./savedWorkflowDraftConsumer.ts";
import type {
  WorkflowDraftDesignerDraft,
  WorkflowDraftDesignerEdge,
  WorkflowDraftDesignerNode,
} from "./workflowDraftDesigner.ts";
import {
  buildWorkflowRAGRequestHeaders,
  type WorkflowRAGSnapshotConfig,
} from "./workflowRAGSnapshotConsumer.ts";
import {
  parseWorkflowRunRecordDocument,
  type WorkflowRunRecord,
} from "./workflowRunRecordConsumer.ts";

const REQUIRED_EXECUTION_SCOPES = [
  "workflow_rag:execute",
  "workflow_runs:execute",
  "workflow_drafts:read",
  "workflow_rag_snapshots:read",
] as const;
const RAG_REF_PATTERN = /^workflow\.rag\.[a-z][a-z0-9_]{2,47}\.v[1-9][0-9]*$/u;
const SCOPE_ID_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._:-]{2,119}$/u;
const FRAGMENT_REF_PATTERN = /^[a-z][a-z0-9_]{2,63}$/u;
const REFERENCE_PATTERN = /^[A-Za-z0-9][A-Za-z0-9_.:/-]{2,159}$/u;
const ANSWER_KEYS = ["schema_version", "answer", "citations", "limitations", "confidence"] as const;
const CITATION_KEYS = ["fragment_ref", "claim_summary"] as const;
const ENVELOPE_KEYS = [
  "request_id", "workspace_id", "application_id", "run", "retrieval_answer",
  "failure_code", "failure_summary", "audit_ref",
] as const;

export type WorkflowRAGCitation = {
  fragmentRef: string;
  claimSummary: string;
};

export type WorkflowRAGAnswer = {
  schemaVersion: "workflow_rag_answer.v1";
  answer: string;
  citations: WorkflowRAGCitation[];
  limitations: string[];
  confidence: "low" | "medium" | "high";
};

export type WorkflowRAGExecutionState = {
  status: "offline" | "idle" | "executing" | "succeeded" | "failed" | "scope_denied";
  summary: string;
  failureCode: string;
  failureSummary: string;
  requestId: string;
  auditRef: string;
  record: WorkflowRunRecord | null;
  answer: WorkflowRAGAnswer | null;
};

export type WorkflowRAGExecutionEligibility = {
  eligible: boolean;
  draftVersion: number;
  retrievalNodeId: string;
  ragRef: string;
  reasons: Array<{ code: string; summary: string }>;
};

export type WorkflowRAGExecutionInput = {
  inputText: string;
  model: string;
  temperature: number | null;
};

type WorkflowRAGExecutionEnvelope = {
  request_id: string;
  workspace_id: string;
  application_id: string;
  run: unknown | null;
  retrieval_answer?: unknown;
  failure_code: string | null;
  failure_summary: string;
  audit_ref: string;
};

export function initialWorkflowRAGExecutionState(
  config: WorkflowRAGSnapshotConfig,
): WorkflowRAGExecutionState {
  if (config.mode !== "dev_workflow_rag_http") {
    return {
      status: "offline",
      summary: "Workflow RAG execution remains offline; zero execution requests are sent.",
      failureCode: "workflow_rag_execution_http_disabled",
      failureSummary: "",
      requestId: "workflow-rag-execution-offline",
      auditRef: "audit-workflow-rag-execution-offline",
      record: null,
      answer: null,
    };
  }
  return {
    status: "idle",
    summary: "Bind an exact active snapshot version, save the four-node draft, then start one retrieval execution.",
    failureCode: "",
    failureSummary: "",
    requestId: "workflow-rag-execution-idle",
    auditRef: "audit-workflow-rag-execution-idle",
    record: null,
    answer: null,
  };
}

export function buildWorkflowRAGRetrievalDraft(
  source: WorkflowDraftDesignerDraft,
  draftNumber: number,
  applicationRef = source.applicationRef,
): WorkflowDraftDesignerDraft {
  const numberLabel = String(Math.max(1, draftNumber)).padStart(2, "0");
  const applicationKey = safeKey(applicationRef, 42);
  const draftId = `draft_${applicationKey}_rag_v1_${numberLabel}`;
  const providerRef = source.providerProfileRef.trim() || "provider:mock";
  const nodes: WorkflowDraftDesignerNode[] = [
    ragDraftNode("node_rag_prompt", "prompt", "Prepare retrieval query", "context", "Use the bounded user question as the exact retrieval query.", "Bounded query context.", "", "", ["input_text"], ["query_context"]),
    ragDraftNode("node_rag_retrieval", "rag_retrieval", "Retrieve immutable evidence", "retrieval", "Read only the selected immutable application snapshot.", "Ranked fragment references and bounded context.", "", "", ["query_context"], ["retrieval_context", "fragment_refs"]),
    ragDraftNode("node_rag_model", "llm", "Answer with citations", "model", "Use only the selected retrieval evidence.", "workflow_rag_answer.v1 advisory response.", providerRef, "", ["retrieval_context", "fragment_refs"], ["answer", "citations", "limitations", "confidence"]),
    ragDraftNode("node_rag_output", "output", "Return cited answer", "output", "Return the validated structured answer for this synchronous execution.", "Reviewable answer and citations without business writeback.", "", "", ["answer", "citations", "limitations", "confidence"], ["answer", "citations", "limitations", "confidence"]),
  ];
  const edges: WorkflowDraftDesignerEdge[] = [
    ragDraftEdge("edge_rag_prompt_retrieval", "node_rag_prompt", "node_rag_retrieval"),
    ragDraftEdge("edge_rag_retrieval_model", "node_rag_retrieval", "node_rag_model"),
    ragDraftEdge("edge_rag_model_output", "node_rag_model", "node_rag_output"),
  ];
  return {
    ...source,
    applicationRef,
    draftId,
    templateRef: source.draftId,
    label: `${source.label} RAG retrieval ${numberLabel}`,
    summary: "Version-pinned application knowledge retrieval followed by one Gateway model call and strict citation validation.",
    providerProfileRef: providerRef,
    nodes,
    edges,
    designerLayout: {
      source: "workflow_node_designer",
      persistence: "ui_only",
      nodePositions: [
        { nodeId: "node_rag_prompt", x: 0, y: 0 },
        { nodeId: "node_rag_retrieval", x: 300, y: 0 },
        { nodeId: "node_rag_model", x: 600, y: 0 },
        { nodeId: "node_rag_output", x: 900, y: 0 },
      ],
    },
    readiness: [
      { checkId: "rag_v1_graph", label: "RAG execution graph", status: "ready", summary: "The draft contains Prompt → RAG Retrieval → LLM → Output only." },
      { checkId: "rag_v1_snapshot", label: "Exact snapshot binding", status: "review_required", summary: "Choose and save one exact active application snapshot version before execution." },
      { checkId: "rag_v1_side_effects", label: "Side-effect boundary", status: "ready", summary: "One retrieval and one provider call are allowed; tool, confirmation, writeback, replay, and resume remain disabled." },
    ],
    risks: [{ riskId: "rag_v1_advisory_answer", label: "Evidence-bound advisory answer", riskLevel: "low", requiresConfirmation: false, summary: "The answer must cite selected immutable snapshot fragments and cannot write business truth." }],
    blockedCapabilities: [],
    routeMetadata: { ...source.routeMetadata, requestId: `req_${draftId}`, auditRef: `audit_${draftId}` },
    localOnlyInteraction: "local_edit",
    executionProfile: "rag_retrieval_v1",
  };
}

export function evaluateWorkflowRAGExecutionEligibility(
  draft: WorkflowDraftDesignerDraft,
  savedDraftState: WorkflowSavedDraftConsumerState,
  draftEditDirty: boolean,
  config: WorkflowRAGSnapshotConfig,
): WorkflowRAGExecutionEligibility {
  const reasons: WorkflowRAGExecutionEligibility["reasons"] = [];
  if (config.mode !== "dev_workflow_rag_http") reasons.push(reason("rag_execution_offline", "RAG execution source is not enabled."));
  for (const scope of REQUIRED_EXECUTION_SCOPES) {
    if (!config.scopes.has(scope)) reasons.push(reason("rag_execution_scope_denied", `Missing exact scope ${scope}.`));
  }
  if (draft.executionProfile !== "rag_retrieval_v1") reasons.push(reason("rag_execution_profile_required", "Create or restore a RAG retrieval v1 draft."));
  if (draftEditDirty) reasons.push(reason("unsaved_local_changes", "Save the current exact draft before execution."));
  if (savedDraftState.currentDraftVersion < 1 || !["saved_dev_record", "validation_ready"].includes(savedDraftState.status)) {
    reasons.push(reason("saved_draft_version_unavailable", "An exact valid saved draft version is required."));
  }
  if (draft.blockedCapabilities.length) reasons.push(reason("blocked_capabilities_present", "The saved draft must have zero blocked capabilities."));

  const prompt = draft.nodes.filter((node) => node.nodeType === "prompt");
  const retrieval = draft.nodes.filter((node) => node.nodeType === "rag_retrieval");
  const llm = draft.nodes.filter((node) => node.nodeType === "llm");
  const output = draft.nodes.filter((node) => node.nodeType === "output");
  const exactNodeSet = draft.nodes.length === 4 && prompt.length === 1 && retrieval.length === 1 && llm.length === 1 && output.length === 1;
  if (!exactNodeSet) reasons.push(reason("rag_execution_topology_invalid", "The graph must contain exactly one Prompt, RAG Retrieval, LLM, and Output node."));
  const retrievalNode = retrieval[0];
  const ragRef = retrievalNode?.ragRef.trim() ?? "";
  if (!RAG_REF_PATTERN.test(ragRef)) reasons.push(reason("rag_ref_invalid", "Select an exact workflow.rag.<snapshot>.v<version> reference."));
  if (draft.nodes.some((node) => node.toolRef.trim() || node.requiresConfirmation || (node !== retrievalNode && node.ragRef.trim())) || retrievalNode?.riskLevel !== "low") {
    reasons.push(reason("rag_execution_node_boundary_invalid", "RAG v1 nodes cannot contain tools, confirmations, extra RAG refs, or elevated retrieval risk."));
  }
  if (exactNodeSet) {
    const expected = new Set([
      `${prompt[0]!.nodeId}\u0000${retrieval[0]!.nodeId}`,
      `${retrieval[0]!.nodeId}\u0000${llm[0]!.nodeId}`,
      `${llm[0]!.nodeId}\u0000${output[0]!.nodeId}`,
    ]);
    const actual = new Set(draft.edges.map((edge) => `${edge.fromNodeId}\u0000${edge.toNodeId}`));
    if (draft.edges.length !== 3 || draft.edges.some((edge) => edge.conditionSummary.trim()) || actual.size !== expected.size || [...actual].some((edge) => !expected.has(edge))) {
      reasons.push(reason("rag_execution_edges_invalid", "The graph must use the three direct edges with empty conditions."));
    }
  }
  return {
    eligible: reasons.length === 0,
    draftVersion: savedDraftState.currentDraftVersion,
    retrievalNodeId: retrievalNode?.nodeId ?? "",
    ragRef,
    reasons,
  };
}

export function validateWorkflowRAGExecutionInput(input: WorkflowRAGExecutionInput): string {
  const queryBytes = new TextEncoder().encode(input.inputText.trim()).length;
  if (!input.inputText.trim() || queryBytes > 4096) return "workflow_rag_query_invalid";
  if (input.model.length > 256 || input.model.includes("://") || containsSecretMaterial(input.model)) return "workflow_rag_draft_ineligible";
  if (input.temperature !== null && (!Number.isFinite(input.temperature) || input.temperature < 0 || input.temperature > 2)) return "workflow_rag_draft_ineligible";
  return "";
}

export async function executeWorkflowRAGRetrieval(
  config: WorkflowRAGSnapshotConfig,
  draft: WorkflowDraftDesignerDraft,
  eligibility: WorkflowRAGExecutionEligibility,
  input: WorkflowRAGExecutionInput,
): Promise<WorkflowRAGExecutionState> {
  const inputFailure = validateWorkflowRAGExecutionInput(input);
  if (!eligibility.eligible || inputFailure || !SCOPE_ID_PATTERN.test(draft.applicationRef) || !SCOPE_ID_PATTERN.test(draft.draftId)) {
    return localExecutionFailure(inputFailure || "workflow_rag_draft_ineligible");
  }
  const body = {
    workspace_id: config.workspaceId,
    application_id: draft.applicationRef,
    draft_version: eligibility.draftVersion,
    input_text: input.inputText.trim(),
    model: input.model.trim(),
    temperature: input.temperature,
  };
  try {
    const response = await fetch(`${config.baseUrl}/v1/user-workspace/workflow-drafts/${encodeURIComponent(draft.draftId)}/retrieval-executions`, {
      method: "POST",
      headers: {
        ...buildWorkflowRAGRequestHeaders(config, draft.applicationRef, REQUIRED_EXECUTION_SCOPES, "execute"),
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    });
    const value: unknown = await response.json();
    if (!isExecutionEnvelope(value, config, draft.applicationRef)) throw new Error("workflow RAG execution returned an unexpected envelope");
    const record = value.run === null ? null : parseWorkflowRunRecordDocument(value.run);
    if (value.run !== null && (!record || record.schemaVersion !== "workflow_run_record.v3" || record.draftId !== draft.draftId || record.draftVersion !== eligibility.draftVersion)) {
      throw new Error("workflow RAG execution returned an invalid run v3 record");
    }
    const answer = value.retrieval_answer === undefined || value.retrieval_answer === null
      ? null
      : parseWorkflowRAGAnswer(value.retrieval_answer, record);
    if (value.retrieval_answer !== undefined && value.retrieval_answer !== null && !answer) {
      throw new Error("workflow RAG execution returned an invalid structured answer");
    }
    if (value.failure_code === null) {
      if (!response.ok || !record || record.status !== "succeeded" || !answer) throw new Error("workflow RAG execution success evidence is incomplete");
      return { status: "succeeded", summary: "One lexical retrieval and one Gateway call completed with validated citations.", failureCode: "", failureSummary: "", requestId: value.request_id, auditRef: value.audit_ref, record, answer };
    }
    if (answer || (record && record.status === "succeeded")) throw new Error("workflow RAG execution failure envelope contains incompatible success evidence");
    return { status: value.failure_code === "workflow_rag_snapshot_scope_denied" ? "scope_denied" : "failed", summary: value.failure_summary || "Workflow RAG execution failed without retry or fallback.", failureCode: value.failure_code, failureSummary: value.failure_summary, requestId: value.request_id, auditRef: value.audit_ref, record, answer: null };
  } catch (error) {
    return { ...localExecutionFailure("workflow_rag_store_unavailable"), summary: error instanceof Error ? error.message : "Workflow RAG execution failed without fallback." };
  }
}

function ragDraftNode(
  nodeId: string,
  nodeType: WorkflowDraftDesignerNode["nodeType"],
  label: string,
  lane: WorkflowDraftDesignerNode["lane"],
  inputSummary: string,
  outputSummary: string,
  providerRef: string,
  ragRef: string,
  inputContractFields: string[],
  outputContractFields: string[],
): WorkflowDraftDesignerNode {
  return { nodeId, nodeType, label, lane, readiness: "ready", inputSummary, outputSummary, providerRef, toolRef: "", ragRef, inputContractFields, outputContractFields, outputMappingSummary: `Map ${outputContractFields.join(", ")} to the next bounded RAG v1 stage.`, riskLevel: "low", requiresConfirmation: false, previewOnlyReason: "This node is executable only through the independent development retrieval route." };
}

function ragDraftEdge(edgeId: string, fromNodeId: string, toNodeId: string): WorkflowDraftDesignerEdge {
  return { edgeId, fromNodeId, toNodeId, edgeKind: "context", conditionSummary: "" };
}

function parseWorkflowRAGAnswer(value: unknown, record: WorkflowRunRecord | null): WorkflowRAGAnswer | null {
  if (!isRecord(value) || !hasOnlyKeys(value, ANSWER_KEYS) || value.schema_version !== "workflow_rag_answer.v1" ||
    typeof value.answer !== "string" || !value.answer.trim() || value.answer.length > 16_384 || containsSecretMaterial(value.answer) ||
    !Array.isArray(value.citations) || value.citations.length < 1 || value.citations.length > 8 ||
    !Array.isArray(value.limitations) || value.limitations.length > 8 ||
    !value.limitations.every((item) => typeof item === "string" && item.trim().length > 0 && item.length <= 512 && !containsSecretMaterial(item)) ||
    (value.confidence !== "low" && value.confidence !== "medium" && value.confidence !== "high")) return null;
  const selectedRefs = new Set(record?.retrievalAttempt?.selectedFragments.map((fragment) => fragment.fragmentRef) ?? []);
  const seen = new Set<string>();
  const citations: WorkflowRAGCitation[] = [];
  for (const rawCitation of value.citations) {
    if (!isRecord(rawCitation) || !hasOnlyKeys(rawCitation, CITATION_KEYS) || !FRAGMENT_REF_PATTERN.test(String(rawCitation.fragment_ref)) ||
      !selectedRefs.has(String(rawCitation.fragment_ref)) || seen.has(String(rawCitation.fragment_ref)) ||
      typeof rawCitation.claim_summary !== "string" || !rawCitation.claim_summary.trim() || rawCitation.claim_summary.length > 512 || containsSecretMaterial(rawCitation.claim_summary)) return null;
    seen.add(rawCitation.fragment_ref as string);
    citations.push({ fragmentRef: rawCitation.fragment_ref as string, claimSummary: rawCitation.claim_summary });
  }
  return { schemaVersion: "workflow_rag_answer.v1", answer: value.answer, citations, limitations: [...value.limitations] as string[], confidence: value.confidence };
}

function isExecutionEnvelope(value: unknown, config: WorkflowRAGSnapshotConfig, applicationId: string): value is WorkflowRAGExecutionEnvelope {
  if (!isRecord(value) || !hasOnlyKnownKeys(value, ENVELOPE_KEYS) || containsForbiddenResponseField(value)) return false;
  return isReference(value.request_id) && value.workspace_id === config.workspaceId && value.application_id === applicationId &&
    (value.run === null || isRecord(value.run)) && (value.failure_code === null || typeof value.failure_code === "string") &&
    typeof value.failure_summary === "string" && isReference(value.audit_ref);
}

function localExecutionFailure(failureCode: string): WorkflowRAGExecutionState {
  return { status: "failed", summary: "Workflow RAG execution input was rejected before any request.", failureCode, failureSummary: "", requestId: "workflow-rag-execution-local-failure", auditRef: "audit-workflow-rag-execution-local-failure", record: null, answer: null };
}

function reason(code: string, summary: string) { return { code, summary }; }
function safeKey(value: string, maxLength: number): string { return value.toLowerCase().replace(/[^a-z0-9]+/gu, "_").replace(/^_+|_+$/gu, "").slice(0, maxLength) || "application"; }
function isRecord(value: unknown): value is Record<string, unknown> { return Boolean(value) && typeof value === "object" && !Array.isArray(value); }
function hasOnlyKeys(value: Record<string, unknown>, allowed: readonly string[]): boolean { const expected = new Set(allowed); return Object.keys(value).length === allowed.length && Object.keys(value).every((key) => expected.has(key)); }
function hasOnlyKnownKeys(value: Record<string, unknown>, allowed: readonly string[]): boolean { const expected = new Set(allowed); return Object.keys(value).every((key) => expected.has(key)); }
function isReference(value: unknown): value is string { return typeof value === "string" && REFERENCE_PATTERN.test(value); }
function containsSecretMaterial(value: string): boolean { return /authorization:|bearer\s|api[_-]?key\s*[:=]|cookie:|password\s*=|secret\s*=|token\s*=|sk-[a-z0-9]|-----begin private key-----|(?:postgres(?:ql)?|mysql|mongodb):\/\//iu.test(value); }
function containsForbiddenResponseField(value: unknown): boolean { if (Array.isArray(value)) return value.some(containsForbiddenResponseField); if (!isRecord(value)) return false; const forbidden = new Set(["input_text", "query", "raw_query", "prompt", "prompt_packet", "messages", "fragment_content", "content", "credential", "authorization", "raw_response", "provider_raw_envelope"]); return Object.entries(value).some(([key, nested]) => forbidden.has(key.toLowerCase()) || containsForbiddenResponseField(nested)); }
