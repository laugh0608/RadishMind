import type { WorkflowExecutorConsumerConfig } from "./workflowExecutorConsumer.ts";
import {
  parseWorkflowRunRecordDocument,
  type WorkflowRAGRunSelectedFragment,
  type WorkflowRunRecord,
  type WorkflowRunSchemaVersion,
  type WorkflowRunStatus,
} from "./workflowRunRecordConsumer.ts";

export type WorkflowRunHistoryFilter = {
  status: "" | WorkflowRunStatus;
  draftId: string;
  executionSourceKind: "" | "workflow_draft" | "application_configuration_draft" | "workflow_definition";
  executionSourceId: string;
  executionSourceVersion: number | "";
  startedFrom: string;
  startedTo: string;
  failureCode: string;
  failureBoundary: "" | "draft_read" | "executor" | "gateway" | "provider" | "run_store" | "request" |
    "tool_policy" | "tool_confirmation" | "tool_transport" | "tool_response" | "tool_store" |
    "retrieval_policy" | "retrieval_store" | "retrieval_rank" | "retrieval_context" | "retrieval_citation" |
    "provider_selection" | "provider_call";
  provider: string;
  model: string;
  staleRunning: "" | "true" | "false";
};

export type WorkflowRunHistorySummary = {
  schemaVersion: WorkflowRunSchemaVersion;
  runId: string;
  planId: string;
  confirmationId: string;
  toolAttemptStatus: "" | "claimed" | "succeeded" | "failed" | "outcome_unknown";
  draftId: string;
  draftVersion: number;
  draftDigest: string;
  executionKind: string;
  executionSourceKind: string;
  executionSourceId: string;
  executionSourceVersion: number;
  executionProfile: string;
  definitionDigest: string;
  activationPointerVersion: number;
  sourceDraftId: string;
  sourceDraftVersion: number;
  sourceDraftDigest: string;
  runtimeAssignmentId: string;
  runtimeAssignmentVersion: number;
  publishCandidateId: string;
  publishReviewVersion: number;
  effectiveSnapshotRole: string;
  status: WorkflowRunStatus;
  failureCode: string;
  startedAt: string;
  completedAt: string;
  durationMs: number;
  selectedProvider: string;
  selectedProfile: string;
  selectedModel: string;
  requestId: string;
  auditRef: string;
  staleRunning: boolean;
  failureBoundary: string;
  failedNodeId: string;
  lastCompletedNodeId: string;
  gatewayFailureCategory: string;
  toolFailureCategory: string;
  snapshotId: string;
  snapshotVersion: number;
  snapshotDigest: string;
  ragRef: string;
  retrievalNodeId: string;
  retrievalAttemptStatus: string;
  retrievalProfileId: string;
  retrievalProfileVersion: number;
  retrievalProfileDigest: string;
  queryDigest: string;
  queryBytes: number;
  candidateCount: number;
  selectedFragments: WorkflowRAGRunSelectedFragment[];
  citationRefs: string[];
  retrievalLatencyMs: number;
  retrievalContextBytes: number;
  retrievalFailureCategory: string;
  recommendedReviewAction: string;
  sideEffects: { retrievalCalls: number; providerCalls: number; toolCalls: number; confirmationCalls: number; businessWrites: number; replayWrites: number };
};

export type WorkflowRunHistoryState = {
  status: "offline" | "loading" | "ready" | "empty" | "failed";
  runs: WorkflowRunHistorySummary[];
  nextCursor: string;
  hasMore: boolean;
  requestId: string;
  auditRef: string;
  failureCode: string;
  failureSummary: string;
};

export function isWorkflowRunComparisonEligible(run: WorkflowRunHistorySummary): boolean {
  return run.schemaVersion !== "workflow_run_record.v2";
}

export function isWorkflowRunComparisonCompatible(
  baseline: WorkflowRunHistorySummary | undefined,
  candidate: WorkflowRunHistorySummary,
): boolean {
  if (!baseline || !isWorkflowRunComparisonEligible(baseline) || !isWorkflowRunComparisonEligible(candidate) || baseline.runId === candidate.runId) return false;
  const baselineApplicationRAG = baseline.schemaVersion === "workflow_run_record.v4";
  const candidateApplicationRAG = candidate.schemaVersion === "workflow_run_record.v4";
  if (baselineApplicationRAG !== candidateApplicationRAG) return false;
  if (baselineApplicationRAG) {
    return baseline.queryDigest === candidate.queryDigest && baseline.queryBytes === candidate.queryBytes &&
      baseline.executionKind === candidate.executionKind && baseline.executionSourceKind === candidate.executionSourceKind;
  }
  const baselineDefinition = baseline.schemaVersion === "workflow_run_record.v5";
  const candidateDefinition = candidate.schemaVersion === "workflow_run_record.v5";
  if (baselineDefinition !== candidateDefinition) return false;
  if (baselineDefinition) return baseline.executionSourceId === candidate.executionSourceId &&
    baseline.executionProfile === "workflow_definition_executor_v1" && candidate.executionProfile === "workflow_definition_executor_v1";
  const baselineRetrieval = baseline.schemaVersion === "workflow_run_record.v3";
  const candidateRetrieval = candidate.schemaVersion === "workflow_run_record.v3";
  if (baselineRetrieval !== candidateRetrieval) return false;
  if (!baselineRetrieval) return true;
  return baseline.snapshotId === candidate.snapshotId && baseline.snapshotVersion === candidate.snapshotVersion &&
    baseline.snapshotDigest === candidate.snapshotDigest && baseline.ragRef === candidate.ragRef &&
    baseline.retrievalProfileId === candidate.retrievalProfileId && baseline.retrievalProfileVersion === candidate.retrievalProfileVersion &&
    baseline.retrievalProfileDigest === candidate.retrievalProfileDigest && baseline.queryDigest === candidate.queryDigest &&
    baseline.queryBytes === candidate.queryBytes && baseline.retrievalNodeId === candidate.retrievalNodeId;
}

type RunHistoryEnvelope = {
  request_id: string; workspace_id: string; application_id: string; runs: RunSummaryDocument[];
  next_cursor: string; has_more: boolean; failure_code: string | null; failure_summary: string; audit_ref: string;
};
type RunSummaryDocument = {
  schema_version: WorkflowRunSchemaVersion; record_version: number; run_id: string;
  plan_id?: string; confirmation_id?: string; tool_attempt_status?: string; draft_id: string;
  draft_version: number; workspace_id: string; application_id: string;
  draft_digest?: string;
  execution_kind?: string; execution_source_kind?: string; execution_source_id?: string; execution_source_version?: number;
  execution_profile?: string; definition_digest?: string; activation_pointer_version?: number;
  source_draft_id?: string; source_draft_version?: number; source_draft_digest?: string;
  runtime_assignment_id?: string; runtime_assignment_version?: number; publish_candidate_id?: string;
  publish_review_version?: number; effective_snapshot_role?: string;
  status: WorkflowRunStatus; failure_code: string;
  started_at: string; completed_at: string; duration_ms: number; selected_provider: string;
  selected_profile: string; selected_model: string; request_id: string; audit_ref: string; stale_running: boolean;
  failure_boundary?: string; failed_node_id?: string; last_completed_node_id?: string;
  gateway_failure_category?: string; tool_failure_category?: string; recommended_review_action?: string;
  snapshot_id?: string; snapshot_version?: number; snapshot_digest?: string; rag_ref?: string;
  retrieval_node_id?: string; retrieval_attempt_status?: string; retrieval_profile_id?: string;
  retrieval_profile_version?: number; retrieval_profile_digest?: string; query_digest?: string; query_bytes?: number;
  candidate_count?: number; selected_fragments?: Array<{ fragment_ref: string; content_digest: string; rank: number; source_type: string; is_official: boolean; excerpt_truncated: boolean }>;
  citation_refs?: string[]; retrieval_latency_ms?: number; retrieval_context_bytes?: number; retrieval_failure_category?: string;
  side_effects: { retrieval_calls?: number; provider_calls: number; tool_calls: number; confirmation_calls: number; business_writes: number; replay_writes: number };
};

export const EMPTY_WORKFLOW_RUN_HISTORY_FILTER: WorkflowRunHistoryFilter = { status: "", draftId: "", executionSourceKind: "", executionSourceId: "", executionSourceVersion: "", startedFrom: "", startedTo: "", failureCode: "", failureBoundary: "", provider: "", model: "", staleRunning: "" };

export function initialWorkflowRunHistoryState(config: WorkflowExecutorConsumerConfig): WorkflowRunHistoryState {
  return config.mode === "dev_workflow_executor_http"
    ? { status: "loading", runs: [], nextCursor: "", hasMore: false, requestId: "workflow-run-history-loading", auditRef: "audit_workflow_run_history_loading", failureCode: "", failureSummary: "" }
    : { status: "offline", runs: [], nextCursor: "", hasMore: false, requestId: "workflow-run-history-offline", auditRef: "audit_workflow_run_history_offline", failureCode: "", failureSummary: "Offline sample mode does not request workflow run history." };
}

export async function listWorkflowRunHistory(
  applicationId: string, config: WorkflowExecutorConsumerConfig, filter: WorkflowRunHistoryFilter,
  cursor = "", previousRuns: WorkflowRunHistorySummary[] = [],
): Promise<WorkflowRunHistoryState> {
  if (config.mode !== "dev_workflow_executor_http") return initialWorkflowRunHistoryState(config);
  const query = new URLSearchParams({ workspace_id: config.workspaceId, application_id: applicationId, limit: "25" });
  if (cursor) query.set("cursor", cursor);
  if (filter.status) query.set("status", filter.status);
  if (filter.draftId.trim()) query.set("draft_id", filter.draftId.trim());
  if (filter.executionSourceKind) query.set("execution_source_kind", filter.executionSourceKind);
  if (filter.executionSourceId.trim()) query.set("execution_source_id", filter.executionSourceId.trim());
  if (filter.executionSourceVersion !== "") query.set("execution_source_version", String(filter.executionSourceVersion));
  if (filter.startedFrom) query.set("started_from", new Date(filter.startedFrom).toISOString());
  if (filter.startedTo) query.set("started_to", new Date(filter.startedTo).toISOString());
  if (filter.failureCode.trim()) query.set("failure_code", filter.failureCode.trim());
  if (filter.failureBoundary) query.set("failure_boundary", filter.failureBoundary);
  if (filter.provider.trim()) query.set("provider", filter.provider.trim());
  if (filter.model.trim()) query.set("model", filter.model.trim());
  if (filter.staleRunning) query.set("stale_running", filter.staleRunning);
  const response = await fetch(`${config.baseUrl}/v1/user-workspace/workflow-runs?${query}`, {
    headers: workflowRunHistoryHeaders(config, applicationId),
  });
  const body: unknown = await response.json();
  if (!response.ok || !isRunHistoryEnvelope(body)) throw new Error(`workflow run history route failed with HTTP ${response.status}`);
  if (body.failure_code) return { status: "failed", runs: [], nextCursor: "", hasMore: false, requestId: body.request_id, auditRef: body.audit_ref, failureCode: body.failure_code, failureSummary: body.failure_summary };
  const runs = body.runs.map(toSummary);
  return { status: previousRuns.length + runs.length ? "ready" : "empty", runs: [...previousRuns, ...runs], nextCursor: body.next_cursor, hasMore: body.has_more, requestId: body.request_id, auditRef: body.audit_ref, failureCode: "", failureSummary: "" };
}

export async function readWorkflowRunHistoryDetail(
  run: WorkflowRunHistorySummary,
  applicationId: string,
  config: WorkflowExecutorConsumerConfig,
  includeRetrievalFragmentPreviews = false,
): Promise<WorkflowRunRecord | null> {
  if (config.mode !== "dev_workflow_executor_http") return null;
  const query = new URLSearchParams({ workspace_id: config.workspaceId, application_id: applicationId });
  if (includeRetrievalFragmentPreviews) query.set("include_retrieval_fragment_previews", "true");
  const response = await fetch(`${config.baseUrl}/v1/user-workspace/workflow-runs/${encodeURIComponent(run.runId)}?${query}`, {
    headers: workflowRunHistoryHeaders(config, applicationId, includeRetrievalFragmentPreviews),
  });
  const body: unknown = await response.json();
  if (!response.ok || !isRunDetailEnvelope(body, config, applicationId)) throw new Error(`workflow run detail route failed with HTTP ${response.status}`);
  if (body.failure_code || body.run === null) return null;
  const record = parseWorkflowRunRecordDocument(body.run);
  if (!record || record.runId !== run.runId || record.applicationId !== applicationId || record.workspaceId !== config.workspaceId) {
    throw new Error("workflow run detail contains an incompatible record");
  }
  const previews = body.retrieval_fragment_previews ?? [];
  if (!isRAGFragmentPreviews(previews, record, includeRetrievalFragmentPreviews)) {
    throw new Error("workflow run detail contains incompatible retrieval previews");
  }
  return { ...record, retrievalFragmentPreviews: previews.map((preview) => ({ fragmentRef: preview.fragment_ref, preview: preview.preview, truncated: preview.truncated })) };
}

function workflowRunHistoryHeaders(config: WorkflowExecutorConsumerConfig, applicationId: string, includeRetrievalFragmentPreviews = false): HeadersInit {
  return { Accept: "application/json", "X-Request-Id": `dev-workflow-run-history-${applicationId}`, "X-RadishMind-Dev-Read-Identity": "dev-workflow-run-history-consumer", "X-RadishMind-Dev-Read-Tenant": config.tenantRef, "X-RadishMind-Dev-Read-Subject": config.subjectRef, "X-RadishMind-Dev-Read-Scopes": includeRetrievalFragmentPreviews ? "workflow_runs:read,workflow_rag_snapshots:read" : "workflow_runs:read", "X-RadishMind-Dev-Read-Audit": "audit_dev_workflow_run_history_consumer", "X-RadishMind-Dev-Workflow-Workspace": config.workspaceId, "X-RadishMind-Dev-Workflow-Application": applicationId };
}

function toSummary(value: RunSummaryDocument): WorkflowRunHistorySummary {
  const sideEffects = value.side_effects;
  const toolRecord = value.schema_version === "workflow_run_record.v2";
  const retrievalRecord = value.schema_version === "workflow_run_record.v3" || value.schema_version === "workflow_run_record.v4";
  if (sideEffects.business_writes || sideEffects.replay_writes ||
    (toolRecord && (sideEffects.tool_calls !== 1 || sideEffects.confirmation_calls !== 1)) ||
    (!toolRecord && (sideEffects.tool_calls !== 0 || sideEffects.confirmation_calls !== 0)) ||
    (retrievalRecord && ![0, 1].includes(sideEffects.retrieval_calls ?? -1)) ||
    (!retrievalRecord && (sideEffects.retrieval_calls ?? 0) !== 0)) {
    throw new Error("workflow run history contains an incompatible side effect count");
  }
  return {
    schemaVersion: value.schema_version, runId: value.run_id, planId: value.plan_id ?? "",
    confirmationId: value.confirmation_id ?? "", toolAttemptStatus: (value.tool_attempt_status ?? "") as WorkflowRunHistorySummary["toolAttemptStatus"],
    draftId: value.draft_id, draftVersion: value.draft_version, draftDigest: value.draft_digest ?? "", status: value.status, failureCode: value.failure_code,
    executionKind: value.execution_kind ?? "", executionSourceKind: value.execution_source_kind ?? "",
    executionSourceId: value.execution_source_id ?? "", executionSourceVersion: value.execution_source_version ?? 0,
    executionProfile: value.execution_profile ?? "", definitionDigest: value.definition_digest ?? "",
    activationPointerVersion: value.activation_pointer_version ?? 0, sourceDraftId: value.source_draft_id ?? "",
    sourceDraftVersion: value.source_draft_version ?? 0, sourceDraftDigest: value.source_draft_digest ?? "",
    runtimeAssignmentId: value.runtime_assignment_id ?? "", runtimeAssignmentVersion: value.runtime_assignment_version ?? 0,
    publishCandidateId: value.publish_candidate_id ?? "", publishReviewVersion: value.publish_review_version ?? 0,
    effectiveSnapshotRole: value.effective_snapshot_role ?? "",
    startedAt: value.started_at, completedAt: value.completed_at, durationMs: value.duration_ms,
    selectedProvider: value.selected_provider, selectedProfile: value.selected_profile, selectedModel: value.selected_model,
    requestId: value.request_id, auditRef: value.audit_ref, staleRunning: value.stale_running,
    failureBoundary: value.failure_boundary ?? "", failedNodeId: value.failed_node_id ?? "",
    lastCompletedNodeId: value.last_completed_node_id ?? "", gatewayFailureCategory: value.gateway_failure_category ?? "",
    toolFailureCategory: value.tool_failure_category ?? "", recommendedReviewAction: value.recommended_review_action ?? "",
    snapshotId: value.snapshot_id ?? "", snapshotVersion: value.snapshot_version ?? 0, snapshotDigest: value.snapshot_digest ?? "", ragRef: value.rag_ref ?? "",
    retrievalNodeId: value.retrieval_node_id ?? "", retrievalAttemptStatus: value.retrieval_attempt_status ?? "", retrievalProfileId: value.retrieval_profile_id ?? "", retrievalProfileVersion: value.retrieval_profile_version ?? 0, retrievalProfileDigest: value.retrieval_profile_digest ?? "", queryDigest: value.query_digest ?? "", queryBytes: value.query_bytes ?? 0, candidateCount: value.candidate_count ?? 0,
    selectedFragments: (value.selected_fragments ?? []).map((fragment) => ({ fragmentRef: fragment.fragment_ref, contentDigest: fragment.content_digest, rank: fragment.rank, sourceType: fragment.source_type as WorkflowRAGRunSelectedFragment["sourceType"], isOfficial: fragment.is_official, excerptTruncated: fragment.excerpt_truncated })),
    citationRefs: [...(value.citation_refs ?? [])], retrievalLatencyMs: value.retrieval_latency_ms ?? 0, retrievalContextBytes: value.retrieval_context_bytes ?? 0, retrievalFailureCategory: value.retrieval_failure_category ?? "",
    sideEffects: { retrievalCalls: sideEffects.retrieval_calls ?? 0, providerCalls: sideEffects.provider_calls, toolCalls: sideEffects.tool_calls,
      confirmationCalls: sideEffects.confirmation_calls, businessWrites: sideEffects.business_writes, replayWrites: sideEffects.replay_writes },
  };
}

function isRunHistoryEnvelope(value: unknown): value is RunHistoryEnvelope {
  if (!value || typeof value !== "object") return false;
  const candidate = value as Partial<RunHistoryEnvelope>;
  return typeof candidate.request_id === "string" && typeof candidate.application_id === "string" && Array.isArray(candidate.runs) && typeof candidate.next_cursor === "string" && typeof candidate.has_more === "boolean" && (candidate.failure_code === null || typeof candidate.failure_code === "string") && typeof candidate.failure_summary === "string" && typeof candidate.audit_ref === "string" && candidate.runs.every(isRunSummary);
}

function isRunSummary(value: unknown): value is RunSummaryDocument {
  if (!value || typeof value !== "object") return false;
  const item = value as Partial<RunSummaryDocument>;
  const raw = value as Record<string, unknown>;
  if (containsForbiddenHistoryField(raw)) return false;
  const schemaValid = item.schema_version === "workflow_run_record.v0" || item.schema_version === "workflow_run_record.v1" || item.schema_version === "workflow_run_record.v2" || item.schema_version === "workflow_run_record.v3" || item.schema_version === "workflow_run_record.v4" || item.schema_version === "workflow_run_record.v5";
  const statusValid = ["running", "succeeded", "failed", "canceled", "outcome_unknown"].includes(item.status ?? "") &&
    (item.status !== "outcome_unknown" || item.schema_version === "workflow_run_record.v2");
  const toolMetadataValid = item.schema_version === "workflow_run_record.v2"
    ? typeof item.plan_id === "string" && typeof item.confirmation_id === "string" &&
      ["claimed", "succeeded", "failed", "outcome_unknown"].includes(item.tool_attempt_status ?? "")
    : [item.plan_id, item.confirmation_id, item.tool_attempt_status].every((field) => field === undefined || field === "");
  const retrievalMetadataValid = item.schema_version === "workflow_run_record.v3" || item.schema_version === "workflow_run_record.v4" ? isRAGRunSummary(item) : true;
  const executionMetadataValid = item.schema_version === "workflow_run_record.v5"
    ? item.draft_id === "" && item.draft_version === 0 && item.execution_kind === "workflow_definition_execution" &&
      item.execution_source_kind === "workflow_definition" && typeof item.execution_source_id === "string" &&
      Number.isInteger(item.execution_source_version) && item.execution_profile === "workflow_definition_executor_v1" &&
      DIGEST_PATTERN.test(item.definition_digest ?? "") && Number.isInteger(item.activation_pointer_version) &&
      typeof item.source_draft_id === "string" && Number.isInteger(item.source_draft_version) &&
      DIGEST_PATTERN.test(item.source_draft_digest ?? "")
    : item.schema_version === "workflow_run_record.v4"
    ? item.execution_kind === "application_rag_invocation" && item.execution_source_kind === "application_configuration_draft" &&
      typeof item.execution_source_id === "string" && Number.isInteger(item.execution_source_version) &&
      typeof item.runtime_assignment_id === "string" && Number.isInteger(item.runtime_assignment_version) &&
      typeof item.publish_candidate_id === "string" && Number.isInteger(item.publish_review_version) && item.effective_snapshot_role === "candidate"
    : (!item.execution_source_kind || (item.execution_source_kind === "workflow_draft" &&
      item.execution_source_id === item.draft_id && item.execution_source_version === item.draft_version &&
      typeof item.execution_kind === "string" && item.execution_kind.length > 0)) &&
      [item.runtime_assignment_id, item.runtime_assignment_version, item.publish_candidate_id,
        item.publish_review_version, item.effective_snapshot_role].every((field) => field === undefined || field === "" || field === 0);
  return schemaValid && statusValid && toolMetadataValid && retrievalMetadataValid && typeof item.record_version === "number" &&
    executionMetadataValid &&
    typeof item.run_id === "string" && typeof item.draft_id === "string" && typeof item.draft_version === "number" &&
    typeof item.started_at === "string" && typeof item.side_effects === "object" &&
    [item.failure_boundary, item.failed_node_id, item.last_completed_node_id, item.gateway_failure_category,
      item.tool_failure_category, item.recommended_review_action].every((field) => field === undefined || typeof field === "string");
}

const DIGEST_PATTERN = /^sha256:[a-f0-9]{64}$/u;

function containsForbiddenHistoryField(value: unknown): boolean {
  if (Array.isArray(value)) return value.some(containsForbiddenHistoryField);
  if (!value || typeof value !== "object") return false;
  const forbidden = new Set([
    "input_text", "condition_values", "query", "raw_query", "prompt", "prompt_packet", "messages",
    "fragment_content", "content", "credential", "authorization", "cookie", "endpoint", "raw_response",
    "provider_raw_envelope",
  ]);
  return Object.entries(value as Record<string, unknown>).some(
    ([key, nested]) => forbidden.has(key.toLowerCase()) || containsForbiddenHistoryField(nested),
  );
}

function isRAGRunSummary(item: Partial<RunSummaryDocument>): boolean {
  const digest = /^sha256:[a-f0-9]{64}$/u;
  const ref = /^[a-z][a-z0-9_]{2,63}$/u;
  const selected = item.selected_fragments;
  if ((item.schema_version === "workflow_run_record.v3" && !digest.test(item.draft_digest ?? "")) ||
    !/^rags_[a-z2-7]{16}$/u.test(item.snapshot_id ?? "") ||
    !Number.isInteger(item.snapshot_version) || (item.snapshot_version ?? 0) < 1 || !digest.test(item.snapshot_digest ?? "") ||
    !/^workflow\.rag\.[a-z][a-z0-9_]{2,47}\.v[1-9][0-9]*$/u.test(item.rag_ref ?? "") ||
    typeof item.retrieval_node_id !== "string" || !["not_started", "succeeded", "failed"].includes(item.retrieval_attempt_status ?? "") ||
    item.retrieval_profile_id !== "workflow.rag.lexical-ngram-dev.v1" || item.retrieval_profile_version !== 1 ||
    !digest.test(item.retrieval_profile_digest ?? "") || !digest.test(item.query_digest ?? "") ||
    !Number.isInteger(item.query_bytes) || (item.query_bytes ?? -1) < 0 || (item.query_bytes ?? 0) > 4096 ||
    !Number.isInteger(item.candidate_count) || (item.candidate_count ?? -1) < 0 || !Array.isArray(selected) || selected.length > 8 ||
    !Array.isArray(item.citation_refs) || !Number.isInteger(item.retrieval_latency_ms) || !Number.isInteger(item.retrieval_context_bytes) ||
    typeof item.retrieval_failure_category !== "string") return false;
  const selectedRefs = new Set<string>();
  for (let index = 0; index < selected.length; index += 1) {
    const fragment = selected[index]!;
    if (!ref.test(fragment.fragment_ref) || !digest.test(fragment.content_digest) || fragment.rank !== index + 1 ||
      !["document", "wiki", "faq", "forum", "manual"].includes(fragment.source_type) || typeof fragment.is_official !== "boolean" ||
      typeof fragment.excerpt_truncated !== "boolean" || selectedRefs.has(fragment.fragment_ref)) return false;
    selectedRefs.add(fragment.fragment_ref);
  }
  return new Set(item.citation_refs).size === item.citation_refs.length && item.citation_refs.every((citation) => selectedRefs.has(citation));
}

type RunDetailEnvelope = { request_id: string; workspace_id: string; application_id: string; run: unknown | null; failure_code: string | null; failure_summary: string; audit_ref: string; retrieval_fragment_previews?: Array<{ fragment_ref: string; preview: string; truncated: boolean }> };

function isRunDetailEnvelope(value: unknown, config: WorkflowExecutorConsumerConfig, applicationId: string): value is RunDetailEnvelope {
  if (!value || typeof value !== "object" || Array.isArray(value)) return false;
  const item = value as Partial<RunDetailEnvelope>;
  const raw = value as Record<string, unknown>;
  const allowed = new Set(["request_id", "workspace_id", "application_id", "run", "failure_code", "failure_summary", "audit_ref", "retrieval_fragment_previews"]);
  return Object.keys(raw).every((key) => allowed.has(key)) && typeof item.request_id === "string" && item.workspace_id === config.workspaceId &&
    item.application_id === applicationId && (item.run === null || typeof item.run === "object") &&
    (item.failure_code === null || typeof item.failure_code === "string") && typeof item.failure_summary === "string" &&
    typeof item.audit_ref === "string" && (item.retrieval_fragment_previews === undefined || Array.isArray(item.retrieval_fragment_previews));
}

function isRAGFragmentPreviews(value: unknown, record: WorkflowRunRecord, requested: boolean): value is Array<{ fragment_ref: string; preview: string; truncated: boolean }> {
  if (!Array.isArray(value) || (!requested && value.length > 0) || (record.schemaVersion !== "workflow_run_record.v3" && value.length > 0)) return false;
  const selectedRefs = new Set(record.retrievalAttempt?.selectedFragments.map((fragment) => fragment.fragmentRef) ?? []);
  const seen = new Set<string>();
  return value.every((preview) => {
    if (!preview || typeof preview !== "object" || Array.isArray(preview)) return false;
    const item = preview as Record<string, unknown>;
    const fragmentRef = item.fragment_ref;
    if (Object.keys(item).length !== 3 || typeof fragmentRef !== "string" || !selectedRefs.has(fragmentRef) || seen.has(fragmentRef) ||
      typeof item.preview !== "string" || [...item.preview].length > 512 || typeof item.truncated !== "boolean") return false;
    seen.add(fragmentRef);
    return true;
  });
}
