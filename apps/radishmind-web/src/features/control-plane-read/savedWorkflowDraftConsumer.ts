import type { WorkflowDraftDesignerDraft } from "./workflowDraftDesigner";

const DEV_SAVED_DRAFT_SOURCE = "dev-saved-draft-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const DEFAULT_WORKSPACE_ID = "workspace_demo";
const DEFAULT_TENANT_REF = "tenant_demo";
const DEFAULT_SUBJECT_REF = "subject_demo_user";
const DEFAULT_SCOPES = "workflow_drafts:read,workflow_drafts:write";
const SAVED_DRAFT_SCHEMA_VERSION = "saved_workflow_draft.v1";

export type WorkflowSavedDraftConsumerMode = "sample_only" | "dev_saved_draft_http";

export type WorkflowSavedDraftConsumerConfig = {
  mode: WorkflowSavedDraftConsumerMode;
  baseUrl: string;
  workspaceId: string;
  tenantRef: string;
  subjectRef: string;
};

export type WorkflowSavedDraftConsumerStatus =
  | "sample"
  | "unsaved_local"
  | "saving"
  | "validating"
  | "reading"
  | "saved_dev_record"
  | "validation_ready"
  | "version_conflict"
  | "save_failed"
  | "read_failed"
  | "validation_failed";

export type WorkflowSavedDraftConsumerState = {
  status: WorkflowSavedDraftConsumerStatus;
  mode: WorkflowSavedDraftConsumerMode;
  sourceLabel: string;
  summary: string;
  failureCode: string | null;
  currentDraftVersion: number;
  conflictDraftVersion: number | null;
  auditRef: string;
  requestId: string;
};

type SavedWorkflowDraftEnvelope = {
  request_id: string;
  workspace_id: string;
  application_id: string;
  draft: SavedWorkflowDraftDocument | null;
  failure_code: string | null;
  current_draft_version: number;
  validation_summary: SavedWorkflowDraftValidationSummary;
  blocked_capabilities: SavedWorkflowDraftBlockedCapability[];
  audit_ref: string;
};

type SavedWorkflowDraftDocument = SavedWorkflowDraftPayload & {
  draft_version: number;
  sample_or_unsaved_draft_status: string;
};

type SavedWorkflowDraftPayload = {
  draft_id: string;
  workspace_id: string;
  application_id: string;
  source_definition_id: string;
  base_definition_version: number;
  schema_version: string;
  name: string;
  description: string;
  nodes: SavedWorkflowDraftNode[];
  edges: SavedWorkflowDraftEdge[];
  input_contract: SavedWorkflowDraftContract;
  output_contract: SavedWorkflowDraftContract;
  provider_refs: string[];
  tool_refs: string[];
  rag_refs: string[];
  requested_capabilities: string[];
};

type SavedWorkflowDraftNode = {
  node_id: string;
  node_type: string;
  label: string;
  input_contract_ref: string;
  output_contract_ref: string;
  provider_ref: string;
  tool_ref: string;
  rag_ref: string;
  risk_level: string;
  requires_confirmation: boolean;
};

type SavedWorkflowDraftEdge = {
  edge_id: string;
  from_node_id: string;
  to_node_id: string;
  condition_summary: string;
};

type SavedWorkflowDraftContract = {
  contract_id: string;
  required_fields: string[];
  summary: string;
};

type SavedWorkflowDraftValidationSummary = {
  validation_state: string;
  valid_for_review: boolean;
};

type SavedWorkflowDraftBlockedCapability = {
  capability_id: string;
};

export function readWorkflowSavedDraftConsumerConfig(): WorkflowSavedDraftConsumerConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  const source = env.VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_SOURCE?.trim();
  return {
    mode: source === DEV_SAVED_DRAFT_SOURCE ? "dev_saved_draft_http" : "sample_only",
    baseUrl: normalizeBaseUrl(
      env.VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_BASE_URL ??
        env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ??
        DEFAULT_BASE_URL,
    ),
    workspaceId: env.VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_WORKSPACE_ID?.trim() || DEFAULT_WORKSPACE_ID,
    tenantRef: env.VITE_RADISHMIND_DEV_READ_TENANT_REF?.trim() || DEFAULT_TENANT_REF,
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || DEFAULT_SUBJECT_REF,
  };
}

export function initialWorkflowSavedDraftConsumerState(
  config: WorkflowSavedDraftConsumerConfig,
): WorkflowSavedDraftConsumerState {
  if (config.mode !== "dev_saved_draft_http") {
    return {
      status: "sample",
      mode: "sample_only",
      sourceLabel: "sample",
      summary: "Offline sample draft is available for review without persistence.",
      failureCode: null,
      currentDraftVersion: 0,
      conflictDraftVersion: null,
      auditRef: "audit_workflow_saved_draft_sample",
      requestId: "workflow-saved-draft-sample",
    };
  }
  return {
    status: "unsaved_local",
    mode: "dev_saved_draft_http",
    sourceLabel: "unsaved local",
    summary: "Local draft can be validated or saved through the dev-only saved draft route.",
    failureCode: null,
    currentDraftVersion: 0,
    conflictDraftVersion: null,
    auditRef: "audit_workflow_saved_draft_unsaved",
    requestId: "workflow-saved-draft-unsaved",
  };
}

export async function saveWorkflowDraftDevRecord(
  draft: WorkflowDraftDesignerDraft,
  config: WorkflowSavedDraftConsumerConfig,
  expectedDraftVersion: number,
): Promise<WorkflowSavedDraftConsumerState> {
  const envelope = await requestSavedWorkflowDraftEnvelope("/v1/user-workspace/workflow-drafts", config, draft, {
    method: "POST",
    body: JSON.stringify({
      expected_draft_version: expectedDraftVersion,
      draft: toSavedWorkflowDraftPayload(draft, config),
    }),
  });
  return stateFromSavedWorkflowDraftEnvelope(envelope, "save");
}

export async function validateWorkflowDraftDevRecord(
  draft: WorkflowDraftDesignerDraft,
  config: WorkflowSavedDraftConsumerConfig,
): Promise<WorkflowSavedDraftConsumerState> {
  const envelope = await requestSavedWorkflowDraftEnvelope(
    "/v1/user-workspace/workflow-drafts/validate",
    config,
    draft,
    {
      method: "POST",
      body: JSON.stringify({ draft: toSavedWorkflowDraftPayload(draft, config) }),
    },
  );
  return stateFromSavedWorkflowDraftEnvelope(envelope, "validate");
}

export async function readWorkflowDraftDevRecord(
  draft: WorkflowDraftDesignerDraft,
  config: WorkflowSavedDraftConsumerConfig,
): Promise<WorkflowSavedDraftConsumerState> {
  const query = new URLSearchParams({
    workspace_id: config.workspaceId,
    application_id: draft.applicationRef,
  });
  const envelope = await requestSavedWorkflowDraftEnvelope(
    `/v1/user-workspace/workflow-drafts/${encodeURIComponent(draft.draftId)}?${query.toString()}`,
    config,
    draft,
    { method: "GET" },
  );
  return stateFromSavedWorkflowDraftEnvelope(envelope, "read");
}

export function nextWorkflowSavedDraftExpectedVersion(state: WorkflowSavedDraftConsumerState): number {
  return state.status === "saved_dev_record" || state.status === "version_conflict" ? state.currentDraftVersion : 0;
}

function stateFromSavedWorkflowDraftEnvelope(
  envelope: SavedWorkflowDraftEnvelope,
  operation: "save" | "read" | "validate",
): WorkflowSavedDraftConsumerState {
  const base = {
    mode: "dev_saved_draft_http" as const,
    failureCode: envelope.failure_code,
    currentDraftVersion: envelope.current_draft_version,
    conflictDraftVersion: null,
    auditRef: envelope.audit_ref,
    requestId: envelope.request_id,
  };
  if (envelope.failure_code) {
    if (envelope.failure_code === "draft_version_conflict") {
      return {
        ...base,
        status: "version_conflict",
        sourceLabel: "version conflict",
        conflictDraftVersion: envelope.current_draft_version,
        summary:
          "Saved draft version conflict. Local draft was kept unchanged; review the current dev record version before saving again.",
      };
    }
    return {
      ...base,
      status: operation === "read" ? "read_failed" : operation === "validate" ? "validation_failed" : "save_failed",
      sourceLabel: envelope.failure_code,
      summary: `Dev saved draft ${operation} failed with ${envelope.failure_code}.`,
    };
  }
  if (operation === "validate") {
    return {
      ...base,
      status: "validation_ready",
      sourceLabel: envelope.validation_summary.validation_state || "validated",
      summary: envelope.validation_summary.valid_for_review
        ? "Draft payload is valid for review through the dev-only validation route."
        : `Draft payload returned ${envelope.validation_summary.validation_state || "review findings"}.`,
    };
  }
  return {
    ...base,
    status: "saved_dev_record",
    sourceLabel: envelope.draft?.sample_or_unsaved_draft_status || "saved dev record",
    summary: `Draft is saved in the dev-only store at version ${envelope.current_draft_version}.`,
  };
}

async function requestSavedWorkflowDraftEnvelope(
  path: string,
  config: WorkflowSavedDraftConsumerConfig,
  draft: WorkflowDraftDesignerDraft,
  init: RequestInit,
): Promise<SavedWorkflowDraftEnvelope> {
  if (config.mode !== "dev_saved_draft_http") {
    throw new Error("saved draft dev HTTP source is disabled");
  }
  const response = await fetch(`${config.baseUrl}${path}`, {
    ...init,
    headers: savedWorkflowDraftHeaders(config, draft),
  });
  const body: unknown = await response.json();
  if (!response.ok) {
    throw new Error(`saved draft route returned HTTP ${response.status}`);
  }
  if (!isSavedWorkflowDraftEnvelope(body)) {
    throw new Error("saved draft route returned an unexpected envelope");
  }
  return body;
}

function savedWorkflowDraftHeaders(
  config: WorkflowSavedDraftConsumerConfig,
  draft: WorkflowDraftDesignerDraft,
): HeadersInit {
  return {
    Accept: "application/json",
    "Content-Type": "application/json",
    "X-Request-Id": `dev-saved-draft-${draft.draftId}`,
    "X-RadishMind-Dev-Read-Identity": "dev-saved-draft-consumer",
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Read-Scopes": DEFAULT_SCOPES,
    "X-RadishMind-Dev-Read-Audit": "audit_dev_saved_draft_consumer",
    "X-RadishMind-Dev-Workflow-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Workflow-Application": draft.applicationRef,
  };
}

function toSavedWorkflowDraftPayload(
  draft: WorkflowDraftDesignerDraft,
  config: WorkflowSavedDraftConsumerConfig,
): SavedWorkflowDraftPayload {
  return {
    draft_id: draft.draftId,
    workspace_id: config.workspaceId,
    application_id: draft.applicationRef,
    source_definition_id: draft.workflowDefinitionId,
    base_definition_version: 0,
    schema_version: SAVED_DRAFT_SCHEMA_VERSION,
    name: draft.label,
    description: draft.summary,
    nodes: draft.nodes.map((node) => {
      const nodeType = String(node.nodeType);
      return {
        node_id: node.nodeId,
        node_type: nodeType,
        label: node.label,
        input_contract_ref: "contract_input",
        output_contract_ref: nodeType === "output" ? "contract_output" : "contract_intermediate",
        provider_ref: nodeType === "llm" ? draft.providerProfileRef : "",
        tool_ref: nodeType === "http_tool" ? "tool:workflow-preview-readonly" : "",
        rag_ref: nodeType === "rag_retrieval" ? "rag:workflow-docs" : "",
        risk_level: node.riskLevel,
        requires_confirmation: node.requiresConfirmation,
      };
    }),
    edges: draft.edges.map((edge) => ({
      edge_id: edge.edgeId,
      from_node_id: edge.fromNodeId,
      to_node_id: edge.toNodeId,
      condition_summary: edge.conditionSummary,
    })),
    input_contract: {
      contract_id: "contract_input",
      required_fields: ["workspace_id", "application_id", "selection_summary"],
      summary: "Workspace-scoped draft input contract.",
    },
    output_contract: {
      contract_id: "contract_output",
      required_fields: ["answer_summary", "risk_summary", "audit_refs"],
      summary: "Reviewable advisory output contract.",
    },
    provider_refs: [draft.providerProfileRef],
    tool_refs: ["tool:workflow-preview-readonly"],
    rag_refs: [],
    requested_capabilities: ["publish", "run", "confirmation_decision", "writeback", "replay"],
  };
}

function isSavedWorkflowDraftEnvelope(value: unknown): value is SavedWorkflowDraftEnvelope {
  if (!value || typeof value !== "object") {
    return false;
  }
  const candidate = value as Partial<SavedWorkflowDraftEnvelope>;
  return typeof candidate.request_id === "string" &&
    typeof candidate.workspace_id === "string" &&
    typeof candidate.application_id === "string" &&
    typeof candidate.current_draft_version === "number" &&
    typeof candidate.audit_ref === "string" &&
    (typeof candidate.failure_code === "string" || candidate.failure_code === null) &&
    !!candidate.validation_summary &&
    typeof candidate.validation_summary === "object";
}

function normalizeBaseUrl(baseUrl: string): string {
  const normalized = baseUrl.trim() || DEFAULT_BASE_URL;
  return normalized.endsWith("/") ? normalized.slice(0, -1) : normalized;
}
