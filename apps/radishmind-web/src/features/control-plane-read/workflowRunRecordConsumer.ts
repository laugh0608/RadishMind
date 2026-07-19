export type WorkflowRunSchemaVersion =
  | "workflow_run_record.v0"
  | "workflow_run_record.v1"
  | "workflow_run_record.v2"
  | "workflow_run_record.v3"
  | "workflow_run_record.v4";

export type WorkflowRunStatus = "running" | "succeeded" | "failed" | "canceled" | "outcome_unknown";

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

export type WorkflowHTTPToolExecutionAttempt = {
  attemptId: string;
  nodeId: string;
  toolId: string;
  definitionDigest: string;
  profileId: string;
  profileDigest: string;
  toolPlanDigest: string;
  confirmationId: string;
  status: "claimed" | "succeeded" | "failed" | "outcome_unknown";
  claimedAt: string;
  completedAt: string;
  httpStatusClass: string;
  responseBytes: number;
  durationMs: number;
  outputProjection: Record<string, string | number | boolean | null>;
  failureCode: string;
};

export type WorkflowRunDiagnostic = {
  failureBoundary:
    | "draft_read"
    | "executor"
    | "gateway"
    | "provider"
    | "run_store"
    | "request"
    | "tool_policy"
    | "tool_confirmation"
    | "tool_transport"
    | "tool_response"
    | "tool_store"
    | "retrieval_policy"
    | "retrieval_store"
    | "retrieval_rank"
    | "retrieval_context"
    | "retrieval_citation"
    | "provider_selection"
    | "provider_call"
    | "";
  failureStage: string;
  failedNodeId: string;
  lastCompletedNodeId: string;
  terminalWriteState: "pending" | "stored";
  gatewayFailureCategory:
    | "none"
    | "queue_full"
    | "timeout"
    | "canceled"
    | "worker_crash"
    | "protocol"
    | "provider_failed"
    | "output_unavailable"
    | "unavailable";
  toolFailureCategory:
    | "none"
    | "policy"
    | "confirmation"
    | "transport"
    | "timeout"
    | "response_status"
    | "response_too_large"
    | "response_invalid"
    | "store"
    | "outcome_unknown";
  retrievalFailureCategory:
    | "none"
    | "scope"
    | "snapshot"
    | "profile"
    | "query"
    | "budget"
    | "no_evidence"
    | "provider"
    | "answer"
    | "citation"
    | "store";
  summary: string;
  recommendedReviewAction:
    | "review_draft"
    | "check_gateway_capacity"
    | "check_provider_configuration"
    | "check_run_store"
    | "start_new_run"
    | "check_tool_policy"
    | "review_tool_outcome"
    | "";
  observedAt: string;
};

export type WorkflowRAGRunSnapshotBinding = {
  snapshotId: string;
  snapshotVersion: number;
  snapshotDigest: string;
  ragRef: string;
};

export type WorkflowRAGRunSelectedFragment = {
  fragmentRef: string;
  contentDigest: string;
  rank: number;
  sourceType: "document" | "wiki" | "faq" | "forum" | "manual";
  isOfficial: boolean;
  excerptTruncated: boolean;
};

export type WorkflowRAGRunRetrievalAttempt = {
  nodeId: string;
  status: "not_started" | "succeeded" | "failed";
  profileId: "workflow.rag.lexical-ngram-dev.v1";
  profileVersion: 1;
  profileDigest: string;
  queryDigest: string;
  queryBytes: number;
  candidateCount: number;
  selectedFragments: WorkflowRAGRunSelectedFragment[];
  retrievalLatencyMs: number;
  contextBytes: number;
  citationRefs: string[];
};

export type WorkflowRAGFragmentPreview = {
  fragmentRef: string;
  preview: string;
  truncated: boolean;
};

export type WorkflowRAGApplicationRunAuthority = {
  assignmentId: string;
  assignmentVersion: number;
  assignmentDigest: string;
  publishCandidateId: string;
  publishReviewVersion: number;
  draftId: string;
  draftVersion: number;
  draftDigest: string;
  bindingId: string;
  bindingVersion: number;
  bindingDigest: string;
  datasetId: string;
  datasetVersion: number;
  datasetDigest: string;
  candidateReviewId: string;
  baselineSnapshot: WorkflowRAGRunSnapshotBinding;
  candidateSnapshot: WorkflowRAGRunSnapshotBinding;
  effectiveSnapshotRole: "candidate";
  profileId: "workflow.rag.lexical-ngram-dev.v1";
  profileVersion: 1;
  profileDigest: string;
  configuredProtocol: string;
  configuredModel: string;
};

export type WorkflowRunRecord = {
  schemaVersion: WorkflowRunSchemaVersion;
  recordVersion: number;
  runId: string;
  planId: string;
  confirmationId: string;
  tenantRef: string;
  draftId: string;
  draftVersion: number;
  draftDigest: string;
  executionKind: string;
  executionSourceKind: string;
  executionSourceId: string;
  executionSourceVersion: number;
  workspaceId: string;
  applicationId: string;
  status: WorkflowRunStatus;
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
  toolAttempt: WorkflowHTTPToolExecutionAttempt | null;
  ragSnapshot: WorkflowRAGRunSnapshotBinding | null;
  retrievalAttempt: WorkflowRAGRunRetrievalAttempt | null;
  retrievalFragmentPreviews: WorkflowRAGFragmentPreview[];
  ragApplicationAuthority: WorkflowRAGApplicationRunAuthority | null;
  output: string;
  requestId: string;
  auditRef: string;
  actorRef: string;
  sideEffects: {
    retrievalCalls: number;
    providerCalls: number;
    toolCalls: number;
    confirmationCalls: number;
    businessWrites: number;
    replayWrites: number;
  };
  diagnostic: WorkflowRunDiagnostic | null;
};

const RUN_ID_PATTERN = /^run_[a-z0-9]{8,64}$/u;
const PLAN_ID_PATTERN = /^wtap_[a-z0-9]{16,64}$/u;
const CONFIRMATION_ID_PATTERN = /^wtcd_[a-z0-9]{16,64}$/u;
const ATTEMPT_ID_PATTERN = /^wtea_[a-z0-9]{16,64}$/u;
const DIGEST_PATTERN = /^sha256:[a-f0-9]{64}$/u;
const PROFILE_ID_PATTERN = /^workflow_http_profile_[a-z0-9_]{1,80}$/u;
const REFERENCE_PATTERN = /^[A-Za-z0-9][A-Za-z0-9_.:/-]{2,159}$/u;
const FORBIDDEN_KEYS = new Set([
  "input_text", "condition_values", "credential", "endpoint", "url", "uri", "header", "headers",
  "authorization", "cookie", "raw_query", "raw_request", "raw_response", "provider_raw_envelope",
  "stderr", "stack", "stack_trace", "sql", "dns", "ip_address", "internal_error",
]);
const TOOL_PROJECTION_KEYS = new Set(["resource_key", "title", "summary", "updated_at"]);
const RAG_RECORD_KEYS = new Set([
  "schema_version", "record_version", "run_id", "tenant_ref", "workspace_id", "application_id",
  "draft_id", "draft_version", "draft_digest", "status", "failure_code", "failure_summary", "started_at",
  "completed_at", "snapshot", "retrieval_attempt", "answer", "selected_provider", "selected_model",
  "request_id", "audit_ref", "actor_ref", "side_effects", "diagnostic",
]);
const APPLICATION_RAG_RECORD_KEYS = new Set([
  "schema_version", "record_version", "run_id", "tenant_ref", "workspace_id", "application_id",
  "execution_kind", "execution_source_kind", "execution_source_id", "execution_source_version", "status",
  "failure_code", "failure_summary", "started_at", "completed_at", "input_digest", "input_bytes",
  "authority", "snapshot", "retrieval_attempt", "answer", "selected_provider", "selected_model",
  "request_id", "audit_ref", "actor_ref", "side_effects", "diagnostic",
]);

export function parseWorkflowRunRecordDocument(value: unknown): WorkflowRunRecord | null {
  if (!isRecord(value) || containsForbiddenWorkflowRunField(value)) return null;
  const schemaVersion = value.schema_version;
  if (schemaVersion !== "workflow_run_record.v0" && schemaVersion !== "workflow_run_record.v1" &&
    schemaVersion !== "workflow_run_record.v2" && schemaVersion !== "workflow_run_record.v3" &&
    schemaVersion !== "workflow_run_record.v4") return null;

  if (schemaVersion === "workflow_run_record.v3") return parseRAGRunRecord(value);
  if (schemaVersion === "workflow_run_record.v4") return parseApplicationRAGRunRecord(value);

  const common = parseCommonRecordFields(value, schemaVersion);
  if (!common) return null;
  if (schemaVersion === "workflow_run_record.v2") return parseToolRunRecord(value, common);
  if (common.status === "outcome_unknown" || common.sideEffects.toolCalls !== 0 ||
    common.sideEffects.confirmationCalls !== 0 || common.sideEffects.businessWrites !== 0 ||
    common.sideEffects.replayWrites !== 0 || value.tool_attempt !== undefined ||
    value.plan_id !== undefined || value.confirmation_id !== undefined) return null;
  const diagnostic = schemaVersion === "workflow_run_record.v0"
    ? null
    : parseWorkflowRunDiagnostic(value.diagnostic, false);
  if (schemaVersion === "workflow_run_record.v1" && !diagnostic) return null;
  return {
    ...common,
    schemaVersion,
    planId: "",
    confirmationId: "",
    tenantRef: optionalString(value.tenant_ref),
    conditionNodeIds: isStringArray(value.condition_node_ids) ? [...value.condition_node_ids] : [],
    toolAttempt: null,
    ragSnapshot: null,
    retrievalAttempt: null,
    retrievalFragmentPreviews: [],
    ragApplicationAuthority: null,
    actorRef: optionalString(value.actor_ref),
    diagnostic,
  };
}

function parseToolRunRecord(
  value: Record<string, unknown>,
  common: Omit<WorkflowRunRecord, "schemaVersion" | "planId" | "confirmationId" | "tenantRef" | "conditionNodeIds" | "toolAttempt" | "actorRef" | "diagnostic">,
): WorkflowRunRecord | null {
  if (!isPatternString(value.plan_id, PLAN_ID_PATTERN) ||
    !isPatternString(value.confirmation_id, CONFIRMATION_ID_PATTERN) ||
    !isPatternString(value.tenant_ref, REFERENCE_PATTERN) ||
    !isPatternString(value.actor_ref, REFERENCE_PATTERN) ||
    value.condition_node_ids !== undefined || common.sideEffects.toolCalls !== 1 ||
    common.sideEffects.confirmationCalls !== 1 || common.sideEffects.businessWrites !== 0 ||
    common.sideEffects.replayWrites !== 0) return null;
  const toolAttempt = parseToolAttempt(value.tool_attempt, value.confirmation_id);
  const diagnostic = parseWorkflowRunDiagnostic(value.diagnostic, true);
  if (!toolAttempt || !diagnostic || !toolRunStateMatches(common.status, common.failureCode, toolAttempt)) return null;
  return {
    ...common,
    schemaVersion: "workflow_run_record.v2",
    planId: value.plan_id,
    confirmationId: value.confirmation_id,
    tenantRef: value.tenant_ref,
    conditionNodeIds: [],
    toolAttempt,
    ragSnapshot: null,
    retrievalAttempt: null,
    retrievalFragmentPreviews: [],
    ragApplicationAuthority: null,
    actorRef: value.actor_ref,
    diagnostic,
  };
}

function parseCommonRecordFields(
  value: Record<string, unknown>,
  schemaVersion: WorkflowRunSchemaVersion,
): Omit<WorkflowRunRecord, "schemaVersion" | "planId" | "confirmationId" | "tenantRef" | "conditionNodeIds" | "toolAttempt" | "actorRef" | "diagnostic"> | null {
  const status = value.status;
  const failureCode = nullableString(value.failure_code);
  const completedAt = nullableString(value.completed_at);
  const nodes = Array.isArray(value.nodes) ? value.nodes.map(parseWorkflowRunNode) : [];
  const sideEffects = parseSideEffects(value.side_effects);
  if (!isPatternString(value.run_id, RUN_ID_PATTERN) || !isPositiveInteger(value.record_version) ||
    !isPatternString(value.draft_id, REFERENCE_PATTERN) || !isPositiveInteger(value.draft_version) ||
    !isPatternString(value.workspace_id, REFERENCE_PATTERN) || !isPatternString(value.application_id, REFERENCE_PATTERN) ||
    !isWorkflowRunStatus(status) || (schemaVersion !== "workflow_run_record.v2" && status === "outcome_unknown") ||
    failureCode === null || typeof value.failure_summary !== "string" || !isTimestamp(value.started_at) ||
    completedAt === null || !isNonNegativeInteger(value.input_bytes) || typeof value.requested_model !== "string" ||
    typeof value.selected_provider !== "string" || typeof value.selected_profile !== "string" ||
    typeof value.selected_model !== "string" || typeof value.upstream_model !== "string" ||
    typeof value.selection_source !== "string" || !Array.isArray(value.nodes) || nodes.some((node) => node === null) ||
    typeof value.output !== "string" || !isPatternString(value.request_id, REFERENCE_PATTERN) ||
    !isPatternString(value.audit_ref, REFERENCE_PATTERN) || !sideEffects) return null;
  return {
    recordVersion: value.record_version,
    runId: value.run_id,
    draftId: value.draft_id,
    draftVersion: value.draft_version,
    draftDigest: "",
    executionKind: "workflow_execution",
    executionSourceKind: "workflow_draft",
    executionSourceId: value.draft_id,
    executionSourceVersion: value.draft_version,
    workspaceId: value.workspace_id,
    applicationId: value.application_id,
    status,
    failureCode,
    failureSummary: value.failure_summary,
    startedAt: value.started_at,
    completedAt,
    inputBytes: value.input_bytes,
    requestedModel: value.requested_model,
    selectedProvider: value.selected_provider,
    selectedProfile: value.selected_profile,
    selectedModel: value.selected_model,
    upstreamModel: value.upstream_model,
    selectionSource: value.selection_source,
    nodes: nodes as WorkflowRunNodeRecord[],
    ragSnapshot: null,
    retrievalAttempt: null,
    retrievalFragmentPreviews: [],
    ragApplicationAuthority: null,
    output: value.output,
    requestId: value.request_id,
    auditRef: value.audit_ref,
    sideEffects,
  };
}

function parseRAGRunRecord(value: Record<string, unknown>): WorkflowRunRecord | null {
  if (Object.keys(value).length !== RAG_RECORD_KEYS.size || Object.keys(value).some((key) => !RAG_RECORD_KEYS.has(key)) ||
    !isPatternString(value.run_id, RUN_ID_PATTERN) || !isPositiveInteger(value.record_version) ||
    !isPatternString(value.tenant_ref, REFERENCE_PATTERN) || !isPatternString(value.workspace_id, REFERENCE_PATTERN) ||
    !isPatternString(value.application_id, REFERENCE_PATTERN) || !isPatternString(value.draft_id, REFERENCE_PATTERN) ||
    !isPositiveInteger(value.draft_version) || !isPatternString(value.draft_digest, DIGEST_PATTERN) ||
    !isRAGRunStatus(value.status) || nullableString(value.failure_code) === null || typeof value.failure_summary !== "string" ||
    value.failure_summary.length > 256 || !isTimestamp(value.started_at) || nullableString(value.completed_at) === null ||
    value.answer !== null || typeof value.selected_provider !== "string" || !value.selected_provider.trim() ||
    typeof value.selected_model !== "string" || !value.selected_model.trim() ||
    !isPatternString(value.request_id, REFERENCE_PATTERN) || !isPatternString(value.audit_ref, REFERENCE_PATTERN) ||
    !isPatternString(value.actor_ref, REFERENCE_PATTERN)) return null;
  const snapshot = parseRAGSnapshotBinding(value.snapshot);
  const retrievalAttempt = parseRAGRetrievalAttempt(value.retrieval_attempt);
  const sideEffects = parseRAGSideEffects(value.side_effects);
  const diagnostic = parseRAGDiagnostic(value.diagnostic, value.status, value.started_at);
  if (!snapshot || !retrievalAttempt || !sideEffects || !diagnostic) return null;
  const completedAt = nullableString(value.completed_at)!;
  const failureCode = nullableString(value.failure_code)!;
  if ((value.status === "running" && (completedAt || failureCode)) ||
    (value.status !== "running" && !completedAt) ||
    (value.status === "succeeded" && (failureCode || retrievalAttempt.status !== "succeeded" ||
      retrievalAttempt.citationRefs.length < 1 || sideEffects.retrievalCalls !== 1 || sideEffects.providerCalls !== 1)) ||
    ((value.status === "failed" || value.status === "canceled") && !failureCode) ||
    (retrievalAttempt.status === "not_started" ? sideEffects.retrievalCalls !== 0 : sideEffects.retrievalCalls !== 1)) return null;
  return {
    schemaVersion: "workflow_run_record.v3",
    recordVersion: value.record_version,
    runId: value.run_id,
    planId: "",
    confirmationId: "",
    tenantRef: value.tenant_ref,
    draftId: value.draft_id,
    draftVersion: value.draft_version,
    draftDigest: value.draft_digest,
    executionKind: "workflow_rag_execution",
    executionSourceKind: "workflow_draft",
    executionSourceId: value.draft_id,
    executionSourceVersion: value.draft_version,
    workspaceId: value.workspace_id,
    applicationId: value.application_id,
    status: value.status,
    failureCode,
    failureSummary: value.failure_summary,
    startedAt: value.started_at,
    completedAt,
    inputBytes: retrievalAttempt.queryBytes,
    conditionNodeIds: [],
    requestedModel: value.selected_model,
    selectedProvider: value.selected_provider,
    selectedProfile: retrievalAttempt.profileId,
    selectedModel: value.selected_model,
    upstreamModel: value.selected_model,
    selectionSource: "workflow_rag_execution_v1",
    nodes: [],
    toolAttempt: null,
    ragSnapshot: snapshot,
    retrievalAttempt,
    retrievalFragmentPreviews: [],
    ragApplicationAuthority: null,
    output: "",
    requestId: value.request_id,
    auditRef: value.audit_ref,
    actorRef: value.actor_ref,
    sideEffects,
    diagnostic,
  };
}

function parseApplicationRAGRunRecord(value: Record<string, unknown>): WorkflowRunRecord | null {
  if (Object.keys(value).length !== APPLICATION_RAG_RECORD_KEYS.size ||
    Object.keys(value).some((key) => !APPLICATION_RAG_RECORD_KEYS.has(key)) ||
    !isPatternString(value.run_id, /^run_[a-z0-9]{16,64}$/u) || !isPositiveInteger(value.record_version) ||
    !isPatternString(value.tenant_ref, REFERENCE_PATTERN) || !isPatternString(value.workspace_id, REFERENCE_PATTERN) ||
    !isPatternString(value.application_id, REFERENCE_PATTERN) || value.execution_kind !== "application_rag_invocation" ||
    value.execution_source_kind !== "application_configuration_draft" ||
    !isPatternString(value.execution_source_id, REFERENCE_PATTERN) || !isPositiveInteger(value.execution_source_version) ||
    !isRAGRunStatus(value.status) || nullableString(value.failure_code) === null ||
    typeof value.failure_summary !== "string" || value.failure_summary.length > 256 ||
    !isTimestamp(value.started_at) || nullableString(value.completed_at) === null ||
    !isPatternString(value.input_digest, DIGEST_PATTERN) || !isBoundedInteger(value.input_bytes, 1, 4096) ||
    value.answer !== null || typeof value.selected_provider !== "string" || !value.selected_provider.trim() ||
    value.selected_provider.includes("://") || typeof value.selected_model !== "string" || !value.selected_model.trim() ||
    value.selected_model.includes("://") || !isPatternString(value.request_id, REFERENCE_PATTERN) ||
    !isPatternString(value.audit_ref, REFERENCE_PATTERN) || !isPatternString(value.actor_ref, REFERENCE_PATTERN)) return null;
  const authority = parseApplicationRAGAuthority(value.authority);
  const snapshot = parseRAGSnapshotBinding(value.snapshot);
  const retrievalAttempt = parseRAGRetrievalAttempt(value.retrieval_attempt);
  const sideEffects = parseRAGSideEffects(value.side_effects);
  const diagnostic = parseRAGDiagnostic(value.diagnostic, value.status, value.started_at);
  if (!authority || !snapshot || !retrievalAttempt || !sideEffects || !diagnostic ||
    value.execution_source_id !== authority.draftId || value.execution_source_version !== authority.draftVersion ||
    snapshot.snapshotId !== authority.candidateSnapshot.snapshotId ||
    snapshot.snapshotVersion !== authority.candidateSnapshot.snapshotVersion ||
    snapshot.snapshotDigest !== authority.candidateSnapshot.snapshotDigest || snapshot.ragRef !== authority.candidateSnapshot.ragRef ||
    retrievalAttempt.profileId !== authority.profileId || retrievalAttempt.profileVersion !== authority.profileVersion ||
    retrievalAttempt.profileDigest !== authority.profileDigest || retrievalAttempt.queryDigest !== value.input_digest ||
    retrievalAttempt.queryBytes !== value.input_bytes) return null;
  const completedAt = nullableString(value.completed_at)!;
  const failureCode = nullableString(value.failure_code)!;
  if ((value.status === "running" && (completedAt || failureCode)) ||
    (value.status !== "running" && !completedAt) ||
    (value.status === "succeeded" && (failureCode || retrievalAttempt.status !== "succeeded" ||
      retrievalAttempt.citationRefs.length < 1 || sideEffects.retrievalCalls !== 1 || sideEffects.providerCalls !== 1)) ||
    ((value.status === "failed" || value.status === "canceled") && !failureCode) ||
    (retrievalAttempt.status === "not_started" ? sideEffects.retrievalCalls !== 0 : sideEffects.retrievalCalls !== 1)) return null;
  return {
    schemaVersion: "workflow_run_record.v4",
    recordVersion: value.record_version,
    runId: value.run_id,
    planId: "",
    confirmationId: "",
    tenantRef: value.tenant_ref,
    draftId: "",
    draftVersion: 0,
    draftDigest: authority.draftDigest,
    executionKind: value.execution_kind,
    executionSourceKind: value.execution_source_kind,
    executionSourceId: value.execution_source_id,
    executionSourceVersion: value.execution_source_version,
    workspaceId: value.workspace_id,
    applicationId: value.application_id,
    status: value.status,
    failureCode,
    failureSummary: value.failure_summary,
    startedAt: value.started_at,
    completedAt,
    inputBytes: value.input_bytes,
    conditionNodeIds: [],
    requestedModel: authority.configuredModel,
    selectedProvider: value.selected_provider,
    selectedProfile: authority.profileId,
    selectedModel: value.selected_model,
    upstreamModel: value.selected_model,
    selectionSource: "workflow_rag_application_invocation.v1",
    nodes: [],
    toolAttempt: null,
    ragSnapshot: snapshot,
    retrievalAttempt,
    retrievalFragmentPreviews: [],
    ragApplicationAuthority: authority,
    output: "",
    requestId: value.request_id,
    auditRef: value.audit_ref,
    actorRef: value.actor_ref,
    sideEffects,
    diagnostic,
  };
}

function parseApplicationRAGAuthority(value: unknown): WorkflowRAGApplicationRunAuthority | null {
  const keys = new Set([
    "assignment_id", "assignment_version", "assignment_digest", "publish_candidate_id", "publish_review_version",
    "publish_candidate_state", "draft_id", "draft_version", "draft_digest", "binding_ref", "dataset",
    "candidate_review_id", "baseline_snapshot", "candidate_snapshot", "effective_snapshot_role", "profile",
    "configured_protocol", "configured_model",
  ]);
  if (!isRecord(value) || Object.keys(value).length !== keys.size || Object.keys(value).some((key) => !keys.has(key)) ||
    !isPatternString(value.assignment_id, /^wragra_[a-z2-7]{16}$/u) || !isPositiveInteger(value.assignment_version) ||
    !isPatternString(value.assignment_digest, DIGEST_PATTERN) || !isPatternString(value.publish_candidate_id, REFERENCE_PATTERN) ||
    !isPositiveInteger(value.publish_review_version) || value.publish_candidate_state !== "approved" ||
    !isPatternString(value.draft_id, REFERENCE_PATTERN) || !isPositiveInteger(value.draft_version) ||
    !isPatternString(value.draft_digest, DIGEST_PATTERN) || !isRecord(value.binding_ref) ||
    Object.keys(value.binding_ref).length !== 3 || !isPatternString(value.binding_ref.binding_id, /^wragb_[a-z2-7]{16}$/u) ||
    value.binding_ref.binding_version !== 1 || !isPatternString(value.binding_ref.binding_digest, DIGEST_PATTERN) ||
    !isRecord(value.dataset) || Object.keys(value.dataset).length !== 3 ||
    !isPatternString(value.dataset.dataset_id, /^wragd_[a-z2-7]{16}$/u) || !isPositiveInteger(value.dataset.dataset_version) ||
    !isPatternString(value.dataset.dataset_digest, DIGEST_PATTERN) ||
    !isPatternString(value.candidate_review_id, /^wragr_[a-z2-7]{16}$/u) || value.effective_snapshot_role !== "candidate" ||
    !isRecord(value.profile) || Object.keys(value.profile).length !== 3 ||
    value.profile.profile_id !== "workflow.rag.lexical-ngram-dev.v1" || value.profile.profile_version !== 1 ||
    !isPatternString(value.profile.profile_digest, DIGEST_PATTERN) || typeof value.configured_protocol !== "string" ||
    value.configured_protocol.includes("://") || typeof value.configured_model !== "string" || value.configured_model.includes("://")) return null;
  const baselineSnapshot = parseRAGSnapshotBinding(value.baseline_snapshot);
  const candidateSnapshot = parseRAGSnapshotBinding(value.candidate_snapshot);
  if (!baselineSnapshot || !candidateSnapshot) return null;
  return {
    assignmentId: value.assignment_id,
    assignmentVersion: value.assignment_version,
    assignmentDigest: value.assignment_digest,
    publishCandidateId: value.publish_candidate_id,
    publishReviewVersion: value.publish_review_version,
    draftId: value.draft_id,
    draftVersion: value.draft_version,
    draftDigest: value.draft_digest,
    bindingId: value.binding_ref.binding_id,
    bindingVersion: 1,
    bindingDigest: value.binding_ref.binding_digest,
    datasetId: value.dataset.dataset_id,
    datasetVersion: value.dataset.dataset_version,
    datasetDigest: value.dataset.dataset_digest,
    candidateReviewId: value.candidate_review_id,
    baselineSnapshot,
    candidateSnapshot,
    effectiveSnapshotRole: "candidate",
    profileId: "workflow.rag.lexical-ngram-dev.v1",
    profileVersion: 1,
    profileDigest: value.profile.profile_digest,
    configuredProtocol: value.configured_protocol,
    configuredModel: value.configured_model,
  };
}

function parseRAGSnapshotBinding(value: unknown): WorkflowRAGRunSnapshotBinding | null {
  if (!isRecord(value) || Object.keys(value).length !== 4 ||
    !isPatternString(value.snapshot_id, /^rags_[a-z2-7]{16}$/u) || !isPositiveInteger(value.snapshot_version) ||
    !isPatternString(value.snapshot_digest, DIGEST_PATTERN) ||
    !isPatternString(value.rag_ref, /^workflow\.rag\.[a-z][a-z0-9_]{2,47}\.v[1-9][0-9]*$/u)) return null;
  return { snapshotId: value.snapshot_id, snapshotVersion: value.snapshot_version, snapshotDigest: value.snapshot_digest, ragRef: value.rag_ref };
}

function parseRAGRetrievalAttempt(value: unknown): WorkflowRAGRunRetrievalAttempt | null {
  if (!isRecord(value) || Object.keys(value).length !== 12 || !isPatternString(value.node_id, REFERENCE_PATTERN) ||
    !isRAGAttemptStatus(value.status) || value.profile_id !== "workflow.rag.lexical-ngram-dev.v1" || value.profile_version !== 1 ||
    !isPatternString(value.profile_digest, DIGEST_PATTERN) || !isPatternString(value.query_digest, DIGEST_PATTERN) ||
    !isBoundedInteger(value.query_bytes, 0, 4096) || !isBoundedInteger(value.candidate_count, 0, 256) ||
    !Array.isArray(value.selected_fragments) || value.selected_fragments.length > 8 ||
    !isBoundedInteger(value.retrieval_latency_ms, 0, 2000) || !isBoundedInteger(value.context_bytes, 0, 32768) ||
    !Array.isArray(value.citation_refs) || value.citation_refs.length > 8 ||
    !value.citation_refs.every((ref) => typeof ref === "string")) return null;
  const selectedFragments = value.selected_fragments.map(parseRAGSelectedFragment);
  if (selectedFragments.some((fragment) => fragment === null) || selectedFragments.length > value.candidate_count) return null;
  const selected = selectedFragments as WorkflowRAGRunSelectedFragment[];
  const selectedRefs = new Set<string>();
  for (let index = 0; index < selected.length; index += 1) {
    const fragment = selected[index]!;
    if (fragment.rank !== index + 1 || selectedRefs.has(fragment.fragmentRef)) return null;
    selectedRefs.add(fragment.fragmentRef);
  }
  const citationRefs = value.citation_refs as string[];
  if (new Set(citationRefs).size !== citationRefs.length || citationRefs.some((ref) => !selectedRefs.has(ref))) return null;
  return {
    nodeId: value.node_id,
    status: value.status,
    profileId: value.profile_id,
    profileVersion: 1,
    profileDigest: value.profile_digest,
    queryDigest: value.query_digest,
    queryBytes: value.query_bytes,
    candidateCount: value.candidate_count,
    selectedFragments: selected,
    retrievalLatencyMs: value.retrieval_latency_ms,
    contextBytes: value.context_bytes,
    citationRefs: [...citationRefs],
  };
}

function parseRAGSelectedFragment(value: unknown): WorkflowRAGRunSelectedFragment | null {
  if (!isRecord(value) || Object.keys(value).length !== 6 ||
    !isPatternString(value.fragment_ref, /^[a-z][a-z0-9_]{2,63}$/u) ||
    !isPatternString(value.content_digest, DIGEST_PATTERN) || !isBoundedInteger(value.rank, 1, 8) ||
    !isRAGSourceType(value.source_type) || typeof value.is_official !== "boolean" ||
    typeof value.excerpt_truncated !== "boolean") return null;
  return { fragmentRef: value.fragment_ref, contentDigest: value.content_digest, rank: value.rank, sourceType: value.source_type, isOfficial: value.is_official, excerptTruncated: value.excerpt_truncated };
}

function parseRAGSideEffects(value: unknown): WorkflowRunRecord["sideEffects"] | null {
  if (!isRecord(value) || Object.keys(value).length !== 6 || !isBoundedInteger(value.retrieval_calls, 0, 1) ||
    !isBoundedInteger(value.provider_calls, 0, 1) || value.tool_calls !== 0 || value.confirmation_calls !== 0 ||
    value.business_writes !== 0 || value.replay_writes !== 0) return null;
  return { retrievalCalls: value.retrieval_calls, providerCalls: value.provider_calls, toolCalls: 0, confirmationCalls: 0, businessWrites: 0, replayWrites: 0 };
}

function parseRAGDiagnostic(value: unknown, status: WorkflowRunStatus, observedAt: string): WorkflowRunDiagnostic | null {
  if (!isRecord(value) || Object.keys(value).length !== 2 ||
    !isRAGFailureBoundary(value.failure_boundary) || !isRAGFailureCategory(value.retrieval_failure_category)) return null;
  return {
    failureBoundary: value.failure_boundary === "none" ? "" : value.failure_boundary,
    failureStage: "workflow_rag_execution",
    failedNodeId: "",
    lastCompletedNodeId: "",
    terminalWriteState: status === "running" ? "pending" : "stored",
    gatewayFailureCategory: "none",
    toolFailureCategory: "none",
    retrievalFailureCategory: value.retrieval_failure_category,
    summary: "",
    recommendedReviewAction: status === "failed" ? "start_new_run" : "",
    observedAt,
  };
}

function parseWorkflowRunNode(value: unknown): WorkflowRunNodeRecord | null {
  if (!isRecord(value)) return null;
  const startedAt = nullableString(value.started_at);
  const completedAt = nullableString(value.completed_at);
  const failureCode = nullableString(value.failure_code);
  if (!isPatternString(value.node_id, REFERENCE_PATTERN) || typeof value.node_type !== "string" ||
    typeof value.label !== "string" || !isWorkflowRunNodeStatus(value.status) || startedAt === null ||
    completedAt === null || !isNonNegativeInteger(value.duration_ms) || !isStringArray(value.predecessor_node_ids) ||
    typeof value.provider_ref !== "string" || typeof value.output_preview !== "string" || failureCode === null) return null;
  return {
    nodeId: value.node_id,
    nodeType: value.node_type,
    label: value.label,
    status: value.status,
    startedAt,
    completedAt,
    durationMs: value.duration_ms,
    predecessorNodeIds: [...value.predecessor_node_ids],
    providerRef: value.provider_ref,
    outputPreview: value.output_preview,
    failureCode,
  };
}

function parseToolAttempt(value: unknown, confirmationId: string): WorkflowHTTPToolExecutionAttempt | null {
  if (!isRecord(value) || !isPatternString(value.attempt_id, ATTEMPT_ID_PATTERN) ||
    !isPatternString(value.node_id, REFERENCE_PATTERN) || !isPatternString(value.tool_id, REFERENCE_PATTERN) ||
    !isPatternString(value.definition_digest, DIGEST_PATTERN) || !isPatternString(value.profile_id, PROFILE_ID_PATTERN) ||
    !isPatternString(value.profile_digest, DIGEST_PATTERN) || !isPatternString(value.tool_plan_digest, DIGEST_PATTERN) ||
    value.confirmation_id !== confirmationId || !isToolAttemptStatus(value.status) || !isTimestamp(value.claimed_at) ||
    nullableString(value.completed_at) === null || nullableString(value.http_status_class) === null ||
    !isNonNegativeInteger(value.response_bytes) || value.response_bytes > 65_536 ||
    !isNonNegativeInteger(value.duration_ms) || value.duration_ms > 30_000 ||
    !isSafeProjection(value.output_projection) || nullableString(value.failure_code) === null) return null;
  return {
    attemptId: value.attempt_id,
    nodeId: value.node_id,
    toolId: value.tool_id,
    definitionDigest: value.definition_digest,
    profileId: value.profile_id,
    profileDigest: value.profile_digest,
    toolPlanDigest: value.tool_plan_digest,
    confirmationId: value.confirmation_id,
    status: value.status,
    claimedAt: value.claimed_at,
    completedAt: nullableString(value.completed_at)!,
    httpStatusClass: nullableString(value.http_status_class)!,
    responseBytes: value.response_bytes,
    durationMs: value.duration_ms,
    outputProjection: { ...value.output_projection },
    failureCode: nullableString(value.failure_code)!,
  };
}

function parseWorkflowRunDiagnostic(value: unknown, toolRecord: boolean): WorkflowRunDiagnostic | null {
  if (!isRecord(value)) return null;
  const failureBoundary = nullableString(value.failure_boundary);
  const failedNodeId = nullableString(value.failed_node_id);
  const lastCompletedNodeId = nullableString(value.last_completed_node_id);
  const toolFailureCategory = toolRecord ? value.tool_failure_category : "none";
  if (failureBoundary === null || failedNodeId === null || lastCompletedNodeId === null ||
    !isFailureBoundary(failureBoundary) || typeof value.failure_stage !== "string" ||
    (value.terminal_write_state !== "pending" && value.terminal_write_state !== "stored") ||
    !isGatewayFailureCategory(value.gateway_failure_category) || !isToolFailureCategory(toolFailureCategory) ||
    typeof value.summary !== "string" || !isReviewAction(value.recommended_review_action) ||
    !isTimestamp(value.observed_at)) return null;
  return {
    failureBoundary,
    failureStage: value.failure_stage,
    failedNodeId,
    lastCompletedNodeId,
    terminalWriteState: value.terminal_write_state,
    gatewayFailureCategory: value.gateway_failure_category,
    toolFailureCategory,
    retrievalFailureCategory: "none",
    summary: value.summary,
    recommendedReviewAction: value.recommended_review_action,
    observedAt: value.observed_at,
  };
}

function parseSideEffects(value: unknown): WorkflowRunRecord["sideEffects"] | null {
  if (!isRecord(value)) return null;
  const counts = [value.provider_calls, value.tool_calls, value.confirmation_calls, value.business_writes, value.replay_writes];
  if (!counts.every(isNonNegativeInteger)) return null;
  return {
    retrievalCalls: 0,
    providerCalls: value.provider_calls as number,
    toolCalls: value.tool_calls as number,
    confirmationCalls: value.confirmation_calls as number,
    businessWrites: value.business_writes as number,
    replayWrites: value.replay_writes as number,
  };
}

function toolRunStateMatches(
  status: WorkflowRunStatus,
  failureCode: string,
  attempt: WorkflowHTTPToolExecutionAttempt,
): boolean {
  if (status === "running") return attempt.status === "claimed" && !failureCode;
  if (status === "succeeded") return attempt.status === "succeeded" && !failureCode;
  if (status === "outcome_unknown") return attempt.status === "outcome_unknown" && failureCode === "workflow_tool_outcome_unknown";
  if (status === "failed" || status === "canceled") {
    return Boolean(failureCode) && (attempt.status === "failed" || attempt.status === "succeeded");
  }
  return false;
}

function isSafeProjection(value: unknown): value is Record<string, string | number | boolean | null> {
  if (!isRecord(value) || Object.keys(value).some((key) => !TOOL_PROJECTION_KEYS.has(key))) return false;
  return Object.values(value).every((item) => item === null || typeof item === "string" ||
    typeof item === "number" || typeof item === "boolean");
}

function containsForbiddenWorkflowRunField(value: unknown): boolean {
  if (Array.isArray(value)) return value.some(containsForbiddenWorkflowRunField);
  if (!isRecord(value)) return false;
  return Object.entries(value).some(([key, nested]) => FORBIDDEN_KEYS.has(key.toLowerCase()) || containsForbiddenWorkflowRunField(nested));
}

function nullableString(value: unknown): string | null {
  return value === null ? "" : typeof value === "string" ? value : null;
}

function optionalString(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function isTimestamp(value: unknown): value is string {
  return typeof value === "string" && value.length >= 20 && Number.isFinite(Date.parse(value));
}

function isPatternString(value: unknown, pattern: RegExp): value is string {
  return typeof value === "string" && pattern.test(value);
}

function isPositiveInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value > 0;
}

function isNonNegativeInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value >= 0;
}

function isStringArray(value: unknown): value is string[] {
  return Array.isArray(value) && value.every((item) => typeof item === "string");
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function isWorkflowRunStatus(value: unknown): value is WorkflowRunStatus {
  return value === "running" || value === "succeeded" || value === "failed" || value === "canceled" || value === "outcome_unknown";
}

function isWorkflowRunNodeStatus(value: unknown): value is WorkflowRunNodeRecord["status"] {
  return value === "pending" || value === "running" || value === "succeeded" || value === "skipped" || value === "failed";
}

function isToolAttemptStatus(value: unknown): value is WorkflowHTTPToolExecutionAttempt["status"] {
  return value === "claimed" || value === "succeeded" || value === "failed" || value === "outcome_unknown";
}

function isFailureBoundary(value: string): value is WorkflowRunDiagnostic["failureBoundary"] {
  return ["", "draft_read", "executor", "gateway", "provider", "run_store", "request", "tool_policy", "tool_confirmation", "tool_transport", "tool_response", "tool_store", "retrieval_policy", "retrieval_store", "retrieval_rank", "retrieval_context", "retrieval_citation", "provider_selection", "provider_call"].includes(value);
}

function isRAGRunStatus(value: unknown): value is Exclude<WorkflowRunStatus, "outcome_unknown"> {
  return value === "running" || value === "succeeded" || value === "failed" || value === "canceled";
}

function isRAGAttemptStatus(value: unknown): value is WorkflowRAGRunRetrievalAttempt["status"] {
  return value === "not_started" || value === "succeeded" || value === "failed";
}

function isRAGSourceType(value: unknown): value is WorkflowRAGRunSelectedFragment["sourceType"] {
  return value === "document" || value === "wiki" || value === "faq" || value === "forum" || value === "manual";
}

function isRAGFailureBoundary(
  value: unknown,
): value is "none" | "retrieval_policy" | "retrieval_store" | "retrieval_rank" | "retrieval_context" | "retrieval_citation" | "provider_selection" | "provider_call" | "run_store" {
  return ["none", "retrieval_policy", "retrieval_store", "retrieval_rank", "retrieval_context", "retrieval_citation", "provider_selection", "provider_call", "run_store"].includes(String(value));
}

function isRAGFailureCategory(value: unknown): value is WorkflowRunDiagnostic["retrievalFailureCategory"] {
  return ["none", "scope", "snapshot", "profile", "query", "budget", "no_evidence", "provider", "answer", "citation", "store"].includes(String(value));
}

function isBoundedInteger(value: unknown, minimum: number, maximum: number): value is number {
  return typeof value === "number" && Number.isInteger(value) && value >= minimum && value <= maximum;
}

function isGatewayFailureCategory(value: unknown): value is WorkflowRunDiagnostic["gatewayFailureCategory"] {
  return ["none", "queue_full", "timeout", "canceled", "worker_crash", "protocol", "provider_failed", "output_unavailable", "unavailable"].includes(String(value));
}

function isToolFailureCategory(value: unknown): value is WorkflowRunDiagnostic["toolFailureCategory"] {
  return ["none", "policy", "confirmation", "transport", "timeout", "response_status", "response_too_large", "response_invalid", "store", "outcome_unknown"].includes(String(value));
}

function isReviewAction(value: unknown): value is WorkflowRunDiagnostic["recommendedReviewAction"] {
  return ["", "review_draft", "check_gateway_capacity", "check_provider_configuration", "check_run_store", "start_new_run", "check_tool_policy", "review_tool_outcome"].includes(String(value));
}
