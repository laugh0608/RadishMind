import { CONTROL_PLANE_READ_ROUTES } from "../../../../../contracts/typescript/control-plane-read-api.ts";
import type {
  WorkflowDraftDesignerDraft,
  WorkflowDraftDesignerLayout,
  WorkflowDraftDesignerLayoutPosition,
  WorkflowDraftDesignerNode,
} from "./workflowDraftDesigner.ts";
import {
  nextWorkflowSavedDraftExpectedVersion,
  resolveWorkflowSavedDraftVersion,
  workflowSavedDraftConflictRequiresResolution,
} from "./savedWorkflowDraftLifecycle.ts";
import { workflowSavedDraftRequestedCapabilities } from "./workflowSavedDraftCapabilityPolicy.ts";

export { nextWorkflowSavedDraftExpectedVersion, workflowSavedDraftConflictRequiresResolution };

const DEV_SAVED_DRAFT_SOURCE = "dev-saved-draft-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const DEFAULT_WORKSPACE_ID = "workspace_demo";
const DEFAULT_TENANT_REF = "tenant_demo";
const DEFAULT_SUBJECT_REF = "subject_demo_user";
const DEFAULT_SCOPES = "workflow_drafts:read,workflow_drafts:write";
const SAVED_DRAFT_SCHEMA_VERSION = "saved_workflow_draft.v1";
const DESIGNER_LAYOUT_VERSION = "designer_layout_v1";
const DESIGNER_LAYOUT_SOURCE = "workflow_node_designer";
const DESIGNER_LAYOUT_PERSISTENCE = "saved_draft_metadata";
const EXECUTOR_V0_METADATA_VERSION = "workflow_executor_v0";
const MAX_DESIGNER_LAYOUT_COORDINATE = 10000;

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
  | "conflict_local_continued"
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

export type WorkflowSavedDraftListStatus =
  | "sample"
  | "loading"
  | "ready"
  | "empty"
  | "list_failed"
  | "restore_failed";

export type WorkflowSavedDraftSummary = {
  draftId: string;
  workspaceId: string;
  applicationRef: string;
  workflowDefinitionId: string;
  draftVersion: number;
  draftStatus: string;
  name: string;
  description: string;
  updatedAt: string;
  updatedByActorRef: string;
  nodeCount: number;
  edgeCount: number;
  blockedCapabilityCount: number;
  validationState: string;
  validForReview: boolean;
  sampleOrUnsavedDraftStatus: string;
};

export type WorkflowSavedDraftListState = {
  status: WorkflowSavedDraftListStatus;
  mode: WorkflowSavedDraftConsumerMode;
  sourceLabel: string;
  summary: string;
  applicationRef: string;
  failureCode: string | null;
  auditRef: string;
  requestId: string;
  summaries: WorkflowSavedDraftSummary[];
};

export type WorkflowSavedDraftConflictMetadataState =
  | "loaded"
  | "refreshing"
  | "empty"
  | "failed"
  | "disabled"
  | "missing";

export type WorkflowSavedDraftRestoreResult = {
  state: WorkflowSavedDraftConsumerState;
  draft: WorkflowDraftDesignerDraft | null;
};

export type WorkflowSavedDraftConflictReviewSummary = {
  reviewId: "saved_draft_conflict_review";
  status: "needs_review" | "local_draft_continued";
  failureCode: "draft_version_conflict";
  draftId: string;
  applicationRef: string;
  workspaceId: string;
  localDraftLabel: string;
  localNodeCount: number;
  localEdgeCount: number;
  savedDraftVersion: number;
  savedUpdatedAt: string;
  savedUpdatedByActorRef: string;
  savedValidationState: string;
  savedValidForReview: boolean | null;
  savedBlockedCapabilityCount: number | null;
  savedMetadataLoaded: boolean;
  savedMetadataState: WorkflowSavedDraftConflictMetadataState;
  restoreActionState: "restore_available" | "restore_requires_saved_list";
  restoreUnavailableReason: string | null;
  localDraftPreservationSummary: string;
  nextReviewerStep: string;
  requestId: string;
  auditRef: string;
  summary: string;
  reviewerQuestion: string;
  canContinueLocalDraft: boolean;
  canRestoreFromSavedDraft: boolean;
  canAutoOverwriteLocalDraft: false;
  canAutoMergeDraft: false;
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

type SavedWorkflowDraftListEnvelope = {
  request_id: string;
  workspace_id: string;
  application_id: string;
  draft_summaries: SavedWorkflowDraftSummaryDocument[];
  failure_code: string | null;
  audit_ref: string;
};

type SavedWorkflowDraftDocument = SavedWorkflowDraftPayload & {
  draft_version: number;
  draft_status: string;
  created_at: string;
  updated_at: string;
  created_by_actor_ref: string;
  updated_by_actor_ref: string;
  validation_summary: SavedWorkflowDraftValidationSummary;
  blocked_capability_summary: SavedWorkflowDraftBlockedCapability[];
  request_audit_metadata: SavedWorkflowDraftAuditMetadata;
  sample_or_unsaved_draft_status: string;
};

type SavedWorkflowDraftSummaryDocument = {
  draft_id: string;
  workspace_id: string;
  application_id: string;
  source_definition_id: string;
  draft_version: number;
  schema_version: string;
  draft_status: string;
  name: string;
  description: string;
  updated_at: string;
  updated_by_actor_ref: string;
  node_count: number;
  edge_count: number;
  blocked_capability_count: number;
  validation_state: string;
  valid_for_review: boolean;
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
  additional_fields?: SavedWorkflowDraftAdditionalFields;
};

type SavedWorkflowDraftAdditionalFields = {
  designer_layout_v1?: SavedWorkflowDraftDesignerLayoutV1;
  executor_v0?: SavedWorkflowDraftExecutorV0Metadata;
  rag_retrieval_v1?: SavedWorkflowDraftRAGRetrievalV1Metadata;
} & Record<string, unknown>;

type SavedWorkflowDraftExecutorV0Metadata = {
  version: typeof EXECUTOR_V0_METADATA_VERSION;
  side_effect_policy: "no_external_side_effects";
};

type SavedWorkflowDraftRAGRetrievalV1Metadata = {
  version: 1;
  execution_route: "retrieval_executions";
  side_effect_policy: "retrieval_and_provider_once";
};

type SavedWorkflowDraftDesignerLayoutV1 = {
  layout_version: typeof DESIGNER_LAYOUT_VERSION;
  source: typeof DESIGNER_LAYOUT_SOURCE;
  persistence: typeof DESIGNER_LAYOUT_PERSISTENCE;
  nodes: SavedWorkflowDraftDesignerLayoutNode[];
};

type SavedWorkflowDraftDesignerLayoutNode = {
  node_id: string;
  x: number;
  y: number;
  pinned: false;
};

type SavedWorkflowDraftNode = {
  node_id: string;
  node_type: string;
  label: string;
  input_summary?: string;
  output_summary?: string;
  input_contract_ref: string;
  output_contract_ref: string;
  input_contract_fields?: string[];
  output_contract_fields?: string[];
  output_mapping_summary?: string;
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
  findings?: SavedWorkflowDraftFinding[];
};

type SavedWorkflowDraftFinding = {
  code: string;
  severity: string;
  field: string;
  summary: string;
  evidence_id: string;
};

type SavedWorkflowDraftBlockedCapability = {
  capability_id: string;
  missing_prerequisite?: string;
  summary?: string;
};

type SavedWorkflowDraftAuditMetadata = {
  request_id: string;
  audit_ref: string;
  actor_ref: string;
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

export function initialWorkflowSavedDraftListState(
  config: WorkflowSavedDraftConsumerConfig,
  applicationRef = "",
): WorkflowSavedDraftListState {
  if (config.mode !== "dev_saved_draft_http") {
    return {
      status: "sample",
      mode: "sample_only",
      sourceLabel: "sample-only",
      summary: "Saved draft list stays disabled until the dev-only saved draft route is enabled.",
      applicationRef,
      failureCode: null,
      auditRef: "audit_workflow_saved_draft_list_sample",
      requestId: "workflow-saved-draft-list-sample",
      summaries: [],
    };
  }
  return {
    status: "empty",
    mode: "dev_saved_draft_http",
    sourceLabel: "not loaded",
    summary: "Saved draft list can be loaded from the dev-only list route for the selected application.",
    applicationRef,
    failureCode: null,
    auditRef: "audit_workflow_saved_draft_list_idle",
    requestId: "workflow-saved-draft-list-idle",
    summaries: [],
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
  return stateFromSavedWorkflowDraftEnvelope(envelope, "save", expectedDraftVersion);
}

export async function validateWorkflowDraftDevRecord(
  draft: WorkflowDraftDesignerDraft,
  config: WorkflowSavedDraftConsumerConfig,
  currentDraftVersion = 0,
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
  return stateFromSavedWorkflowDraftEnvelope(envelope, "validate", currentDraftVersion);
}

export async function readWorkflowDraftDevRecord(
  draft: WorkflowDraftDesignerDraft,
  config: WorkflowSavedDraftConsumerConfig,
  currentDraftVersion = 0,
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
  return stateFromSavedWorkflowDraftEnvelope(envelope, "read", currentDraftVersion);
}

export async function listWorkflowDraftDevRecords(
  applicationRef: string,
  config: WorkflowSavedDraftConsumerConfig,
): Promise<WorkflowSavedDraftListState> {
  const query = new URLSearchParams({
    workspace_id: config.workspaceId,
    application_id: applicationRef,
  });
  const envelope = await requestSavedWorkflowDraftListEnvelope(
    `/v1/user-workspace/workflow-drafts?${query.toString()}`,
    config,
    applicationRef,
  );
  return listStateFromSavedWorkflowDraftEnvelope(envelope);
}

export async function restoreWorkflowDraftDevRecord(
  summary: WorkflowSavedDraftSummary,
  config: WorkflowSavedDraftConsumerConfig,
): Promise<WorkflowSavedDraftRestoreResult> {
  const query = new URLSearchParams({
    workspace_id: config.workspaceId,
    application_id: summary.applicationRef,
  });
  const envelope = await requestSavedWorkflowDraftEnvelopeForApplication(
    `/v1/user-workspace/workflow-drafts/${encodeURIComponent(summary.draftId)}?${query.toString()}`,
    config,
    summary.applicationRef,
    `dev-saved-draft-restore-${summary.draftId}`,
    { method: "GET" },
  );
  const state = stateFromSavedWorkflowDraftEnvelope(envelope, "read", summary.draftVersion);
  if (envelope.failure_code || !envelope.draft) {
    return { state, draft: null };
  }
  return { state, draft: workflowDraftFromSavedWorkflowDraftDocument(envelope.draft) };
}

export function continueLocalWorkflowDraftAfterVersionConflict(
  state: WorkflowSavedDraftConsumerState,
  draft: WorkflowDraftDesignerDraft,
): WorkflowSavedDraftConsumerState {
  if (state.status !== "version_conflict") {
    return state;
  }
  const savedDraftVersion = state.conflictDraftVersion ?? state.currentDraftVersion;
  return {
    ...state,
    status: "conflict_local_continued",
    sourceLabel: "local draft continued",
    summary: `Local draft ${draft.draftId} remains active after explicit conflict review; retry save against saved version ${savedDraftVersion} or restore the saved draft explicitly.`,
    failureCode: "draft_version_conflict",
    currentDraftVersion: savedDraftVersion,
    conflictDraftVersion: savedDraftVersion,
  };
}

export function buildWorkflowSavedDraftConflictReviewSummary(
  state: WorkflowSavedDraftConsumerState,
  draft: WorkflowDraftDesignerDraft,
  savedDraftSummaries: WorkflowSavedDraftSummary[],
  savedDraftListStatus?: WorkflowSavedDraftListStatus,
  savedDraftListFailureCode?: string | null,
): WorkflowSavedDraftConflictReviewSummary | null {
  if (state.status !== "version_conflict" && state.status !== "conflict_local_continued") {
    return null;
  }
  const savedSummary = savedDraftSummaries.find(
    (summary) => summary.draftId === draft.draftId && summary.applicationRef === draft.applicationRef,
  );
  const savedDraftVersion = state.conflictDraftVersion ?? state.currentDraftVersion;
  const status =
    state.status === "conflict_local_continued" ? "local_draft_continued" : "needs_review";
  const savedMetadataLoaded = savedSummary !== undefined;
  const savedMetadataState = savedMetadataLoaded
    ? "loaded"
    : conflictMetadataStateFromSavedDraftListStatus(savedDraftListStatus);
  const restoreActionState = savedMetadataLoaded ? "restore_available" : "restore_requires_saved_list";
  const restoreUnavailableReason = savedMetadataLoaded
    ? null
    : conflictRestoreUnavailableReason(savedMetadataState, savedDraftListFailureCode);
  const savedMetadataLabel = savedSummary
    ? `saved version ${savedSummary.draftVersion} updated at ${savedSummary.updatedAt} by ${savedSummary.updatedByActorRef}`
    : `saved version ${savedDraftVersion} metadata is not loaded in the sanitized saved draft list`;
  const localDraftPreservationSummary =
    status === "local_draft_continued"
      ? `Local draft ${draft.draftId} remains active after explicit continue; the next save retries against saved version ${savedDraftVersion}.`
      : `Local draft ${draft.draftId} remains active and unchanged; no saved version has been restored or merged.`;
  const nextReviewerStep =
    status === "local_draft_continued"
      ? "Review the local edits before the next save attempt; restoring the saved version remains an explicit separate action."
      : savedMetadataLoaded
        ? "Choose continue local draft to keep the current browser edits, or restore saved version to replace the active draft explicitly."
        : conflictNextReviewerStep(savedMetadataState);
  return {
    reviewId: "saved_draft_conflict_review",
    status,
    failureCode: "draft_version_conflict",
    draftId: draft.draftId,
    applicationRef: draft.applicationRef,
    workspaceId: savedSummary?.workspaceId ?? "dev_saved_draft_workspace",
    localDraftLabel: draft.label,
    localNodeCount: draft.nodes.length,
    localEdgeCount: draft.edges.length,
    savedDraftVersion: savedSummary?.draftVersion ?? savedDraftVersion,
    savedUpdatedAt: savedSummary?.updatedAt ?? "not_loaded",
    savedUpdatedByActorRef: savedSummary?.updatedByActorRef ?? "not_loaded",
    savedValidationState: savedSummary?.validationState ?? "not_loaded",
    savedValidForReview: savedSummary?.validForReview ?? null,
    savedBlockedCapabilityCount: savedSummary?.blockedCapabilityCount ?? null,
    savedMetadataLoaded,
    savedMetadataState,
    restoreActionState,
    restoreUnavailableReason,
    localDraftPreservationSummary,
    nextReviewerStep,
    requestId: state.requestId,
    auditRef: state.auditRef,
    summary: `Version conflict review keeps local draft ${draft.draftId} active with ${draft.nodes.length} nodes and ${draft.edges.length} edges while ${savedMetadataLabel} remains the remote saved draft source.`,
    reviewerQuestion:
      "Should the reviewer keep editing the local draft or explicitly restore the saved draft before another save attempt?",
    canContinueLocalDraft: true,
    canRestoreFromSavedDraft: savedMetadataLoaded,
    canAutoOverwriteLocalDraft: false,
    canAutoMergeDraft: false,
  };
}

function conflictMetadataStateFromSavedDraftListStatus(
  status: WorkflowSavedDraftListStatus | undefined,
): WorkflowSavedDraftConflictMetadataState {
  switch (status) {
    case "sample":
      return "disabled";
    case "loading":
      return "refreshing";
    case "empty":
      return "empty";
    case "list_failed":
    case "restore_failed":
      return "failed";
    case "ready":
      return "missing";
    default:
      return "refreshing";
  }
}

function conflictRestoreUnavailableReason(
  metadataState: WorkflowSavedDraftConflictMetadataState,
  failureCode: string | null | undefined,
): string {
  switch (metadataState) {
    case "disabled":
      return "Saved draft list is disabled outside the dev-only saved draft route; restore stays unavailable.";
    case "empty":
      return "Saved draft list returned no sanitized summary for this draft; restore stays disabled and the local draft remains active.";
    case "failed":
      return `Saved version metadata refresh failed with ${failureCode ?? "unknown_failure"}; restore stays disabled and no sample fallback was used.`;
    case "missing":
      return "Saved draft list is ready but does not include this draft summary; restore stays disabled until matching sanitized metadata is available.";
    case "refreshing":
    case "loaded":
      return "Saved version metadata is refreshing from the dev-only saved draft list before restore is enabled.";
  }
}

function conflictNextReviewerStep(metadataState: WorkflowSavedDraftConflictMetadataState): string {
  switch (metadataState) {
    case "disabled":
      return "Enable the dev-only saved draft route before trying to restore; local browser edits stay active.";
    case "empty":
    case "missing":
      return "Refresh saved drafts or keep editing locally; restore remains blocked until a matching sanitized summary is available.";
    case "failed":
      return "Resolve the saved draft list failure before restore; keep reviewing the local draft without auto overwrite or auto merge.";
    case "refreshing":
    case "loaded":
      return "Keep reviewing the local draft while the saved draft list refreshes; restore stays disabled until sanitized saved metadata is available.";
  }
}

function listStateFromSavedWorkflowDraftEnvelope(
  envelope: SavedWorkflowDraftListEnvelope,
): WorkflowSavedDraftListState {
  const base = {
    mode: "dev_saved_draft_http" as const,
    applicationRef: envelope.application_id,
    failureCode: envelope.failure_code,
    auditRef: envelope.audit_ref,
    requestId: envelope.request_id,
    summaries: envelope.draft_summaries.map(toWorkflowSavedDraftSummary),
  };
  if (envelope.failure_code) {
    return {
      ...base,
      status: "list_failed",
      sourceLabel: envelope.failure_code,
      summary: `Saved draft list failed with ${envelope.failure_code}; no sample fallback was used.`,
    };
  }
  if (envelope.draft_summaries.length === 0) {
    return {
      ...base,
      status: "empty",
      sourceLabel: "empty",
      summary: "No saved dev draft records were returned for the selected application.",
    };
  }
  return {
    ...base,
    status: "ready",
    sourceLabel: "saved dev drafts",
    summary: `${envelope.draft_summaries.length} saved dev draft summaries are available for restore.`,
  };
}

function stateFromSavedWorkflowDraftEnvelope(
  envelope: SavedWorkflowDraftEnvelope,
  operation: "save" | "read" | "validate",
  previousDraftVersion: number,
): WorkflowSavedDraftConsumerState {
  const base = {
    mode: "dev_saved_draft_http" as const,
    failureCode: envelope.failure_code,
    currentDraftVersion: resolveWorkflowSavedDraftVersion({
      operation,
      failureCode: envelope.failure_code,
      envelopeDraftVersion: envelope.current_draft_version,
      previousDraftVersion,
    }),
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
  return requestSavedWorkflowDraftEnvelopeForApplication(
    path,
    config,
    draft.applicationRef,
    `dev-saved-draft-${draft.draftId}`,
    init,
  );
}

async function requestSavedWorkflowDraftEnvelopeForApplication(
  path: string,
  config: WorkflowSavedDraftConsumerConfig,
  applicationRef: string,
  requestId: string,
  init: RequestInit,
): Promise<SavedWorkflowDraftEnvelope> {
  if (config.mode !== "dev_saved_draft_http") {
    throw new Error("saved draft dev HTTP source is disabled");
  }
  const response = await fetch(`${config.baseUrl}${path}`, {
    ...init,
    headers: savedWorkflowDraftHeadersForApplication(config, applicationRef, requestId),
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

async function requestSavedWorkflowDraftListEnvelope(
  path: string,
  config: WorkflowSavedDraftConsumerConfig,
  applicationRef: string,
): Promise<SavedWorkflowDraftListEnvelope> {
  if (config.mode !== "dev_saved_draft_http") {
    throw new Error("saved draft dev HTTP source is disabled");
  }
  const response = await fetch(`${config.baseUrl}${path}`, {
    method: "GET",
    headers: savedWorkflowDraftHeadersForApplication(
      config,
      applicationRef,
      `dev-saved-draft-list-${applicationRef}`,
    ),
  });
  const body: unknown = await response.json();
  if (!response.ok) {
    throw new Error(`saved draft list route returned HTTP ${response.status}`);
  }
  if (!isSavedWorkflowDraftListEnvelope(body)) {
    throw new Error("saved draft list route returned an unexpected envelope");
  }
  return body;
}

function savedWorkflowDraftHeadersForApplication(
  config: WorkflowSavedDraftConsumerConfig,
  applicationRef: string,
  requestId: string,
): HeadersInit {
  return {
    Accept: "application/json",
    "Content-Type": "application/json",
    "X-Request-Id": requestId,
    "X-RadishMind-Dev-Read-Identity": "dev-saved-draft-consumer",
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Read-Scopes": DEFAULT_SCOPES,
    "X-RadishMind-Dev-Read-Audit": "audit_dev_saved_draft_consumer",
    "X-RadishMind-Dev-Workflow-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Workflow-Application": applicationRef,
  };
}

function toSavedWorkflowDraftPayload(
  draft: WorkflowDraftDesignerDraft,
  config: WorkflowSavedDraftConsumerConfig,
): SavedWorkflowDraftPayload {
  const additionalFields = toSavedWorkflowDraftAdditionalFields(draft);
  return {
    draft_id: draft.draftId,
    workspace_id: config.workspaceId,
    application_id: draft.applicationRef,
    source_definition_id: draft.workflowDefinitionId,
    base_definition_version: draft.baseDefinitionVersion ?? 0,
    schema_version: SAVED_DRAFT_SCHEMA_VERSION,
    name: draft.label,
    description: draft.summary,
    nodes: draft.nodes.map((node) => {
      const nodeType = String(node.nodeType);
      const providerRef = workflowDraftNodeProviderRef(draft, node);
      const toolRef = workflowDraftNodeToolRef(node);
      const ragRef = workflowDraftNodeRagRef(node);
      return {
        node_id: node.nodeId,
        node_type: nodeType,
        label: node.label,
        input_summary: node.inputSummary,
        output_summary: node.outputSummary,
        input_contract_ref: "contract_input",
        output_contract_ref: nodeType === "output" ? "contract_output" : "contract_intermediate",
        input_contract_fields: normalizedWorkflowDraftFields(node.inputContractFields),
        output_contract_fields: normalizedWorkflowDraftFields(node.outputContractFields),
        output_mapping_summary: node.outputMappingSummary,
        provider_ref: providerRef,
        tool_ref: toolRef,
        rag_ref: ragRef,
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
      required_fields: workflowDraftRequiredFields(draft.nodes, "input"),
      summary: "Workspace-scoped draft input contract.",
    },
    output_contract: {
      contract_id: "contract_output",
      required_fields: workflowDraftRequiredFields(draft.nodes, "output"),
      summary: "Reviewable advisory output contract.",
    },
    provider_refs: workflowDraftProviderRefs(draft),
    tool_refs: workflowDraftToolRefs(draft),
    rag_refs: workflowDraftRagRefs(draft),
    requested_capabilities: workflowSavedDraftRequestedCapabilities(draft),
    ...(additionalFields ? { additional_fields: additionalFields } : {}),
  };
}

function toSavedWorkflowDraftAdditionalFields(
  draft: WorkflowDraftDesignerDraft,
): SavedWorkflowDraftAdditionalFields | undefined {
  const layoutNodes = savedWorkflowDraftDesignerLayoutNodes(draft);
  const additionalFields: SavedWorkflowDraftAdditionalFields = {};
  if (layoutNodes.length > 0) {
    additionalFields.designer_layout_v1 = {
      layout_version: DESIGNER_LAYOUT_VERSION,
      source: DESIGNER_LAYOUT_SOURCE,
      persistence: DESIGNER_LAYOUT_PERSISTENCE,
      nodes: layoutNodes,
    };
  }
  if (draft.executionProfile === "executor_v0") {
    additionalFields.executor_v0 = {
      version: EXECUTOR_V0_METADATA_VERSION,
      side_effect_policy: "no_external_side_effects",
    };
  }
  if (draft.executionProfile === "rag_retrieval_v1") {
    additionalFields.rag_retrieval_v1 = {
      version: 1,
      execution_route: "retrieval_executions",
      side_effect_policy: "retrieval_and_provider_once",
    };
  }
  return Object.keys(additionalFields).length > 0 ? additionalFields : undefined;
}

function savedWorkflowDraftDesignerLayoutNodes(
  draft: WorkflowDraftDesignerDraft,
): SavedWorkflowDraftDesignerLayoutNode[] {
  const nodeIds = new Set(draft.nodes.map((node) => node.nodeId));
  const positionsByNodeId = new Map<string, SavedWorkflowDraftDesignerLayoutNode>();
  for (const position of draft.designerLayout.nodePositions) {
    if (!nodeIds.has(position.nodeId)) {
      continue;
    }
    positionsByNodeId.set(position.nodeId, {
      node_id: position.nodeId,
      x: workflowDraftDesignerLayoutCoordinate(position.x),
      y: workflowDraftDesignerLayoutCoordinate(position.y),
      pinned: false,
    });
  }
  return draft.nodes
    .map((node) => positionsByNodeId.get(node.nodeId))
    .filter((position): position is SavedWorkflowDraftDesignerLayoutNode => Boolean(position));
}

function toWorkflowSavedDraftSummary(document: SavedWorkflowDraftSummaryDocument): WorkflowSavedDraftSummary {
  return {
    draftId: document.draft_id,
    workspaceId: document.workspace_id,
    applicationRef: document.application_id,
    workflowDefinitionId: document.source_definition_id,
    draftVersion: document.draft_version,
    draftStatus: document.draft_status,
    name: document.name,
    description: document.description,
    updatedAt: document.updated_at,
    updatedByActorRef: document.updated_by_actor_ref,
    nodeCount: document.node_count,
    edgeCount: document.edge_count,
    blockedCapabilityCount: document.blocked_capability_count,
    validationState: document.validation_state,
    validForReview: document.valid_for_review,
    sampleOrUnsavedDraftStatus: document.sample_or_unsaved_draft_status,
  };
}

function workflowDraftFromSavedWorkflowDraftDocument(
  document: SavedWorkflowDraftDocument,
): WorkflowDraftDesignerDraft {
  const validationState = document.validation_summary?.validation_state || document.draft_status || "invalid_draft";
  const blockedCapabilities = document.blocked_capability_summary ?? [];
  const nodes = document.nodes.map(toWorkflowDraftDesignerNode);
  const isRAGRetrievalV1 = isSavedWorkflowDraftRAGRetrievalV1Metadata(
    document.additional_fields?.rag_retrieval_v1,
  );
  return {
    draftId: document.draft_id,
    templateRef: document.source_definition_id || document.draft_id,
    label: document.name,
    applicationRef: document.application_id,
    workflowDefinitionId: document.source_definition_id,
    baseDefinitionVersion: document.base_definition_version,
    providerProfileRef: document.provider_refs[0] ?? "",
    summary: document.description,
    nodes,
    edges: document.edges.map((edge) => ({
      edgeId: edge.edge_id,
      fromNodeId: edge.from_node_id,
      toNodeId: edge.to_node_id,
      edgeKind: "context",
      conditionSummary: isRAGRetrievalV1
        ? edge.condition_summary
        : edge.condition_summary || "Saved draft edge restored from dev record.",
    })),
    designerLayout: workflowDraftDesignerLayoutFromSavedDraftAdditionalFields(document.additional_fields, nodes),
    readiness: [
      {
        checkId: "saved_draft_restore_validation",
        label: "Saved draft validation",
        status: document.validation_summary?.valid_for_review ? "ready" : "review_required",
        summary: `Restored dev saved draft version ${document.draft_version} with ${validationState}.`,
      },
      {
        checkId: "saved_draft_restore_scope",
        label: "Scope guard",
        status: "ready",
        summary: "Restore used workspace and application scoped dev-only read route.",
      },
    ],
    risks: buildRestoredWorkflowDraftRisks(document),
    blockedCapabilities: blockedCapabilities.map((capability) => ({
      capabilityId: capability.capability_id,
      label: capability.capability_id,
      status: "blocked",
      missingPrerequisite: capability.missing_prerequisite ?? "independent workflow runtime target",
      summary: capability.summary ?? "Capability remains blocked after saved draft restore.",
      auditRef: document.request_audit_metadata?.audit_ref ?? "audit_saved_draft_restore",
    })),
    routeMetadata: {
      sourceRouteId: "workflow-definition-summary-list-route",
      draftRouteId: "workflow-draft-designer-offline-draft",
      routePath: CONTROL_PLANE_READ_ROUTES.workflowDefinitions,
      requestId: document.request_audit_metadata?.request_id ?? `restore_${document.draft_id}`,
      auditRef: document.request_audit_metadata?.audit_ref ?? "audit_saved_draft_restore",
    },
    localOnlyInteraction: "inspect_only",
    executionProfile: isSavedWorkflowDraftExecutorV0Metadata(document.additional_fields?.executor_v0)
      ? "executor_v0"
      : isRAGRetrievalV1
        ? "rag_retrieval_v1"
        : "review_only",
  };
}

function isSavedWorkflowDraftExecutorV0Metadata(value: unknown): value is SavedWorkflowDraftExecutorV0Metadata {
  if (!value || typeof value !== "object") {
    return false;
  }
  const candidate = value as Partial<SavedWorkflowDraftExecutorV0Metadata>;
  return candidate.version === EXECUTOR_V0_METADATA_VERSION &&
    candidate.side_effect_policy === "no_external_side_effects";
}

function isSavedWorkflowDraftRAGRetrievalV1Metadata(
  value: unknown,
): value is SavedWorkflowDraftRAGRetrievalV1Metadata {
  if (!value || typeof value !== "object") {
    return false;
  }
  const candidate = value as Partial<SavedWorkflowDraftRAGRetrievalV1Metadata>;
  return candidate.version === 1 &&
    candidate.execution_route === "retrieval_executions" &&
    candidate.side_effect_policy === "retrieval_and_provider_once";
}

function workflowDraftDesignerLayoutFromSavedDraftAdditionalFields(
  additionalFields: SavedWorkflowDraftAdditionalFields | undefined,
  nodes: WorkflowDraftDesignerNode[],
): WorkflowDraftDesignerLayout {
  const nodePositions = workflowDraftDesignerLayoutPositionsFromSavedDraft(
    additionalFields?.designer_layout_v1,
    nodes,
  );
  return {
    source: DESIGNER_LAYOUT_SOURCE,
    persistence: nodePositions.length > 0 ? DESIGNER_LAYOUT_PERSISTENCE : "ui_only",
    nodePositions,
  };
}

function workflowDraftDesignerLayoutPositionsFromSavedDraft(
  value: unknown,
  nodes: WorkflowDraftDesignerNode[],
): WorkflowDraftDesignerLayoutPosition[] {
  if (!isSavedWorkflowDraftDesignerLayoutV1(value)) {
    return [];
  }
  const nodeIds = new Set(nodes.map((node) => node.nodeId));
  const positionsByNodeId = new Map<string, WorkflowDraftDesignerLayoutPosition>();
  for (const position of value.nodes) {
    if (!nodeIds.has(position.node_id)) {
      continue;
    }
    positionsByNodeId.set(position.node_id, {
      nodeId: position.node_id,
      x: workflowDraftDesignerLayoutCoordinate(position.x),
      y: workflowDraftDesignerLayoutCoordinate(position.y),
    });
  }
  return nodes
    .map((node) => positionsByNodeId.get(node.nodeId))
    .filter((position): position is WorkflowDraftDesignerLayoutPosition => Boolean(position));
}

function isSavedWorkflowDraftDesignerLayoutV1(value: unknown): value is SavedWorkflowDraftDesignerLayoutV1 {
  if (!value || typeof value !== "object") {
    return false;
  }
  const candidate = value as Partial<SavedWorkflowDraftDesignerLayoutV1>;
  return candidate.layout_version === DESIGNER_LAYOUT_VERSION &&
    candidate.source === DESIGNER_LAYOUT_SOURCE &&
    candidate.persistence === DESIGNER_LAYOUT_PERSISTENCE &&
    Array.isArray(candidate.nodes) &&
    candidate.nodes.every(isSavedWorkflowDraftDesignerLayoutNode);
}

function isSavedWorkflowDraftDesignerLayoutNode(value: unknown): value is SavedWorkflowDraftDesignerLayoutNode {
  if (!value || typeof value !== "object") {
    return false;
  }
  const candidate = value as Partial<SavedWorkflowDraftDesignerLayoutNode>;
  return typeof candidate.node_id === "string" &&
    candidate.node_id.trim().length > 0 &&
    typeof candidate.x === "number" &&
    typeof candidate.y === "number" &&
    Number.isFinite(candidate.x) &&
    Number.isFinite(candidate.y) &&
    candidate.pinned === false;
}

function workflowDraftDesignerLayoutCoordinate(value: number): number {
  if (!Number.isFinite(value)) {
    return 0;
  }
  return Math.max(-MAX_DESIGNER_LAYOUT_COORDINATE, Math.min(MAX_DESIGNER_LAYOUT_COORDINATE, Math.round(value)));
}

function toWorkflowDraftDesignerNode(node: SavedWorkflowDraftNode): WorkflowDraftDesignerNode {
  const nodeType = toWorkflowDraftDesignerNodeType(node.node_type);
  return {
    nodeId: node.node_id,
    nodeType,
    label: node.label,
    lane: workflowDraftLaneForNodeType(nodeType),
    readiness: node.requires_confirmation || node.risk_level === "high" ? "review_required" : "ready",
    inputSummary: node.input_summary || node.input_contract_ref || "saved draft input",
    outputSummary: node.output_summary || node.output_contract_ref || "saved draft output",
    providerRef: node.provider_ref ?? "",
    toolRef: node.tool_ref ?? "",
    ragRef: node.rag_ref ?? "",
    inputContractFields: normalizedWorkflowDraftFields(node.input_contract_fields ?? [node.input_contract_ref]),
    outputContractFields: normalizedWorkflowDraftFields(node.output_contract_fields ?? [node.output_contract_ref]),
    outputMappingSummary: node.output_mapping_summary || "Saved draft output mapping restored from dev record.",
    riskLevel: toWorkflowDraftRiskLevel(node.risk_level),
    requiresConfirmation: node.requires_confirmation,
    previewOnlyReason: "Restored from dev-only saved draft record; execution remains blocked.",
  };
}

function workflowDraftNodeProviderRef(
  draft: WorkflowDraftDesignerDraft,
  node: WorkflowDraftDesignerNode,
): string {
  const providerRef = node.providerRef.trim();
  if (providerRef) {
    return providerRef;
  }
  if (node.nodeType === "llm") {
    return draft.providerProfileRef;
  }
  if (node.nodeType === "condition") {
    return "policy:confirmation-gated";
  }
  return "";
}

function workflowDraftNodeToolRef(node: WorkflowDraftDesignerNode): string {
  const toolRef = node.toolRef.trim();
  if (toolRef) {
    return toolRef;
  }
  return node.nodeType === "http_tool" ? "tool:workflow-preview-readonly" : "";
}

function workflowDraftNodeRagRef(node: WorkflowDraftDesignerNode): string {
  return node.ragRef.trim();
}

function workflowDraftProviderRefs(draft: WorkflowDraftDesignerDraft): string[] {
  const modelProviderRefs = draft.nodes
    .filter((node) => node.nodeType === "llm")
    .map((node) => workflowDraftNodeProviderRef(draft, node));
  return uniqueWorkflowDraftRefs([draft.providerProfileRef, ...modelProviderRefs]);
}

function workflowDraftToolRefs(draft: WorkflowDraftDesignerDraft): string[] {
  return uniqueWorkflowDraftRefs(draft.nodes.map(workflowDraftNodeToolRef));
}

function workflowDraftRagRefs(draft: WorkflowDraftDesignerDraft): string[] {
  return uniqueWorkflowDraftRefs(draft.nodes.map(workflowDraftNodeRagRef));
}

function workflowDraftRequiredFields(
  nodes: WorkflowDraftDesignerNode[],
  contractKind: "input" | "output",
): string[] {
  const fields = nodes.flatMap((node) =>
    contractKind === "input" ? node.inputContractFields : node.outputContractFields,
  );
  const normalized = normalizedWorkflowDraftFields(fields);
  if (normalized.length > 0) {
    return normalized;
  }
  return contractKind === "input"
    ? ["workspace_id", "application_id", "selection_summary"]
    : ["answer_summary", "risk_summary", "audit_refs"];
}

function normalizedWorkflowDraftFields(fields: string[]): string[] {
  return uniqueWorkflowDraftRefs(fields.map((field) => field.toLowerCase().replace(/[^a-z0-9]+/g, "_").replace(/^_+|_+$/g, "")));
}

function uniqueWorkflowDraftRefs(values: string[]): string[] {
  const seen = new Set<string>();
  return values
    .map((value) => value.trim())
    .filter((value) => {
      if (!value || seen.has(value)) {
        return false;
      }
      seen.add(value);
      return true;
    });
}

function toWorkflowDraftDesignerNodeType(nodeType: string): WorkflowDraftDesignerNode["nodeType"] {
  switch (nodeType) {
    case "prompt":
    case "rag_retrieval":
    case "llm":
    case "http_tool":
    case "condition":
    case "output":
      return nodeType;
    default:
      return "prompt";
  }
}

function workflowDraftLaneForNodeType(nodeType: WorkflowDraftDesignerNode["nodeType"]): WorkflowDraftDesignerNode["lane"] {
  switch (nodeType) {
    case "rag_retrieval":
      return "retrieval";
    case "llm":
      return "model";
    case "condition":
      return "policy";
    case "http_tool":
      return "preview";
    case "output":
      return "output";
    default:
      return "context";
  }
}

function toWorkflowDraftRiskLevel(riskLevel: string): WorkflowDraftDesignerNode["riskLevel"] {
  if (riskLevel === "high" || riskLevel === "medium" || riskLevel === "low") {
    return riskLevel;
  }
  return "low";
}

function buildRestoredWorkflowDraftRisks(document: SavedWorkflowDraftDocument): WorkflowDraftDesignerDraft["risks"] {
  const riskNodes = document.nodes.filter((node) => node.requires_confirmation || node.risk_level === "high");
  if (riskNodes.length === 0) {
    return [
      {
        riskId: "saved_draft_restore_review",
        label: "Saved draft review",
        riskLevel: "low",
        requiresConfirmation: false,
        summary: "Restored draft keeps advisory-only review state and does not unlock execution.",
      },
    ];
  }
  return riskNodes.map((node) => ({
    riskId: `risk_${node.node_id}`,
    label: node.label,
    riskLevel: toWorkflowDraftRiskLevel(node.risk_level),
    requiresConfirmation: node.requires_confirmation,
    summary: "Risk marker restored for review; confirmation decision storage remains unavailable.",
  }));
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

function isSavedWorkflowDraftListEnvelope(value: unknown): value is SavedWorkflowDraftListEnvelope {
  if (!value || typeof value !== "object") {
    return false;
  }
  const candidate = value as Partial<SavedWorkflowDraftListEnvelope>;
  return typeof candidate.request_id === "string" &&
    typeof candidate.workspace_id === "string" &&
    typeof candidate.application_id === "string" &&
    Array.isArray(candidate.draft_summaries) &&
    typeof candidate.audit_ref === "string" &&
    (typeof candidate.failure_code === "string" || candidate.failure_code === null);
}

function normalizeBaseUrl(baseUrl: string): string {
  const normalized = baseUrl.trim() || DEFAULT_BASE_URL;
  return normalized.endsWith("/") ? normalized.slice(0, -1) : normalized;
}
