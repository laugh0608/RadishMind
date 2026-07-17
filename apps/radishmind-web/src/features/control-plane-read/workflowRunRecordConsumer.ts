export type WorkflowRunSchemaVersion =
  | "workflow_run_record.v0"
  | "workflow_run_record.v1"
  | "workflow_run_record.v2";

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

export type WorkflowRunRecord = {
  schemaVersion: WorkflowRunSchemaVersion;
  recordVersion: number;
  runId: string;
  planId: string;
  confirmationId: string;
  tenantRef: string;
  draftId: string;
  draftVersion: number;
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
  output: string;
  requestId: string;
  auditRef: string;
  actorRef: string;
  sideEffects: {
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

export function parseWorkflowRunRecordDocument(value: unknown): WorkflowRunRecord | null {
  if (!isRecord(value) || containsForbiddenWorkflowRunField(value)) return null;
  const schemaVersion = value.schema_version;
  if (schemaVersion !== "workflow_run_record.v0" && schemaVersion !== "workflow_run_record.v1" &&
    schemaVersion !== "workflow_run_record.v2") return null;

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
    output: value.output,
    requestId: value.request_id,
    auditRef: value.audit_ref,
    sideEffects,
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
  return ["", "draft_read", "executor", "gateway", "provider", "run_store", "request", "tool_policy", "tool_confirmation", "tool_transport", "tool_response", "tool_store"].includes(value);
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
