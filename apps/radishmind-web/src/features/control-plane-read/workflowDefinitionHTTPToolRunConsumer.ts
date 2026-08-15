import type {
  WorkflowDefinitionRunAuthority,
  WorkflowHTTPToolExecutionAttempt,
  WorkflowRunDiagnostic,
  WorkflowRunNodeRecord,
  WorkflowRunRecord,
  WorkflowRunStatus,
} from "./workflowRunRecordConsumer.ts";

const RECORD_KEYS = [
  "schema_version", "record_version", "run_id", "plan_id", "confirmation_id", "tenant_ref",
  "draft_id", "draft_version", "workspace_id", "application_id", "execution_kind",
  "execution_source_kind", "execution_source_id", "execution_source_version", "execution_profile",
  "input_digest", "definition_authority", "status", "failure_code", "failure_summary", "started_at",
  "completed_at", "input_bytes", "condition_node_ids", "requested_model", "selected_provider",
  "selected_profile", "selected_model", "upstream_model", "selection_source", "nodes", "tool_attempt",
  "output", "request_id", "audit_ref", "actor_ref", "side_effects", "diagnostic",
] as const;
const AUTHORITY_KEYS = [
  "definition_id", "definition_version", "definition_digest", "activation_pointer_version", "candidate_id",
  "candidate_review_version", "source_draft_id", "source_draft_version", "source_draft_digest",
  "application_record_version", "application_lifecycle",
] as const;
const NODE_KEYS = [
  "node_id", "node_type", "label", "status", "started_at", "completed_at", "duration_ms",
  "predecessor_node_ids", "provider_ref", "output_preview", "failure_code",
] as const;
const ATTEMPT_KEYS = [
  "attempt_id", "node_id", "tool_id", "definition_digest", "profile_id", "profile_digest",
  "tool_plan_digest", "confirmation_id", "status", "claimed_at", "completed_at", "http_status_class",
  "response_bytes", "duration_ms", "output_projection", "failure_code",
] as const;
const SIDE_EFFECT_KEYS = ["provider_calls", "tool_calls", "confirmation_calls", "business_writes", "replay_writes"] as const;
const DIAGNOSTIC_KEYS = [
  "failure_boundary", "failure_stage", "failed_node_id", "last_completed_node_id", "terminal_write_state",
  "gateway_failure_category", "tool_failure_category", "summary", "recommended_review_action", "observed_at",
] as const;
const OPTIONAL_DIAGNOSTIC_KEY = "retrieval_failure_category";
const DIGEST_PATTERN = /^sha256:[a-f0-9]{64}$/u;
const REFERENCE_PATTERN = /^[A-Za-z0-9][A-Za-z0-9_.:/-]{2,159}$/u;
const SCOPED_ID_PATTERN = /^[A-Za-z0-9][A-Za-z0-9_.-]{2,119}$/u;
const RUN_ID_PATTERN = /^run_[a-z0-9]{16,64}$/u;
const PLAN_ID_PATTERN = /^wtap_[a-z0-9]{16,64}$/u;
const CONFIRMATION_ID_PATTERN = /^wtcd_[a-z0-9]{16,64}$/u;
const ATTEMPT_ID_PATTERN = /^wtea_[a-z0-9]{16,64}$/u;
const TOOL_ID_PATTERN = /^workflow\.http\.[a-z0-9][a-z0-9._-]*\.v[0-9]+$/u;
const PROFILE_ID_PATTERN = /^workflow_http_profile_[a-z0-9_]{1,80}$/u;
const SAFE_PROJECTION_KEY_PATTERN = /^[a-z][a-z0-9_]{0,63}$/u;
const SENSITIVE_TEXT_PATTERN = /(?:[A-Za-z][A-Za-z0-9+.-]*:\/\/|authorization\s*[:=]|bearer\s+|api[_-]?key\s*[:=]|cookie\s*[:=]|x-radishmind-)/iu;

export function parseWorkflowDefinitionHTTPToolRunRecord(value: Record<string, unknown>): WorkflowRunRecord | null {
  if (!hasExactKeys(value, RECORD_KEYS) || value.schema_version !== "workflow_run_record.v9" ||
    !isPositiveInteger(value.record_version) || !isPatternString(value.run_id, RUN_ID_PATTERN) ||
    !isPatternString(value.plan_id, PLAN_ID_PATTERN) || !isPatternString(value.confirmation_id, CONFIRMATION_ID_PATTERN) ||
    !isPatternString(value.tenant_ref, REFERENCE_PATTERN) || value.draft_id !== "" || value.draft_version !== 0 ||
    !isPatternString(value.workspace_id, SCOPED_ID_PATTERN) || !isPatternString(value.application_id, SCOPED_ID_PATTERN) ||
    value.execution_kind !== "workflow_definition_http_tool_execution" ||
    value.execution_source_kind !== "workflow_definition" || !isPatternString(value.execution_source_id, SCOPED_ID_PATTERN) ||
    !isPositiveInteger(value.execution_source_version) || value.execution_profile !== "workflow_definition_http_tool_v1" ||
    !isPatternString(value.input_digest, DIGEST_PATTERN) || !isRunStatus(value.status) ||
    typeof value.failure_code !== "string" || value.failure_code.length > 120 ||
    !isSafeString(value.failure_summary, 256) || !isTimestamp(value.started_at) ||
    typeof value.completed_at !== "string" || value.completed_at.length > 64 ||
    !isBoundedInteger(value.input_bytes, 1, 8192) || !Array.isArray(value.condition_node_ids) || value.condition_node_ids.length !== 0 ||
    !isSafeString(value.requested_model, 256) || !isSafeString(value.selected_provider, 256) ||
    !isSafeString(value.selected_profile, 256) || !isSafeString(value.selected_model, 256) ||
    !isSafeString(value.upstream_model, 256) || !isSafeString(value.selection_source, 256) || value.output !== "" ||
    !isPatternString(value.request_id, REFERENCE_PATTERN) || !isPatternString(value.audit_ref, REFERENCE_PATTERN) ||
    !isPatternString(value.actor_ref, REFERENCE_PATTERN) || !Array.isArray(value.nodes) || value.nodes.length < 4 || value.nodes.length > 16) {
    return null;
  }

  const authority = parseAuthority(value.definition_authority);
  const nodes = value.nodes.map(parseNode);
  const attempt = parseAttempt(value.tool_attempt, value.confirmation_id);
  const sideEffects = parseSideEffects(value.side_effects);
  const diagnostic = parseDiagnostic(value.diagnostic);
  if (!authority || nodes.some((node) => !node) || !attempt || !sideEffects || !diagnostic ||
    authority.definitionId !== value.execution_source_id || authority.definitionVersion !== value.execution_source_version ||
    !runStateMatches(value.status, value.failure_code, value.completed_at, attempt)) return null;

  return {
    schemaVersion: "workflow_run_record.v9", recordVersion: value.record_version, runId: value.run_id,
    planId: value.plan_id, confirmationId: value.confirmation_id, tenantRef: value.tenant_ref,
    draftId: "", draftVersion: 0, draftDigest: "", executionKind: value.execution_kind,
    executionSourceKind: value.execution_source_kind, executionSourceId: value.execution_source_id,
    executionSourceVersion: value.execution_source_version, executionProfile: value.execution_profile,
    inputDigest: value.input_digest, definitionAuthority: authority, workspaceId: value.workspace_id,
    applicationId: value.application_id, status: value.status, failureCode: value.failure_code,
    failureSummary: value.failure_summary, startedAt: value.started_at, completedAt: value.completed_at,
    inputBytes: value.input_bytes, conditionNodeIds: [], requestedModel: value.requested_model,
    selectedProvider: value.selected_provider, selectedProfile: value.selected_profile,
    selectedModel: value.selected_model, upstreamModel: value.upstream_model, selectionSource: value.selection_source,
    nodes: nodes as WorkflowRunNodeRecord[], toolAttempt: attempt, ragSnapshot: null, retrievalAttempt: null,
    retrievalFragmentPreviews: [], ragApplicationAuthority: null, output: "", requestId: value.request_id,
    auditRef: value.audit_ref, actorRef: value.actor_ref, sideEffects, diagnostic,
  };
}

function parseAuthority(value: unknown): WorkflowDefinitionRunAuthority | null {
  if (!isRecord(value) || !hasExactKeys(value, AUTHORITY_KEYS) || !isPatternString(value.definition_id, SCOPED_ID_PATTERN) ||
    !isPositiveInteger(value.definition_version) || !isPatternString(value.definition_digest, DIGEST_PATTERN) ||
    !isPositiveInteger(value.activation_pointer_version) || !isPatternString(value.candidate_id, SCOPED_ID_PATTERN) ||
    !isPositiveInteger(value.candidate_review_version) || !isPatternString(value.source_draft_id, SCOPED_ID_PATTERN) ||
    !isPositiveInteger(value.source_draft_version) || !isPatternString(value.source_draft_digest, DIGEST_PATTERN) ||
    !isPositiveInteger(value.application_record_version) || value.application_lifecycle !== "active") return null;
  return {
    definitionId: value.definition_id, definitionVersion: value.definition_version,
    definitionDigest: value.definition_digest, activationPointerVersion: value.activation_pointer_version,
    candidateId: value.candidate_id, candidateReviewVersion: value.candidate_review_version,
    sourceDraftId: value.source_draft_id, sourceDraftVersion: value.source_draft_version,
    sourceDraftDigest: value.source_draft_digest, applicationRecordVersion: value.application_record_version,
    applicationLifecycle: "active",
  };
}

function parseNode(value: unknown): WorkflowRunNodeRecord | null {
  if (!isRecord(value) || !hasExactKeys(value, NODE_KEYS) || !isPatternString(value.node_id, SCOPED_ID_PATTERN) ||
    typeof value.node_type !== "string" || !["prompt", "http_tool", "llm", "output"].includes(value.node_type) ||
    !isSafeString(value.label, 256) || typeof value.status !== "string" ||
    !["pending", "running", "succeeded", "skipped", "failed"].includes(value.status) ||
    typeof value.started_at !== "string" || value.started_at.length > 64 || typeof value.completed_at !== "string" ||
    value.completed_at.length > 64 || !isNonNegativeInteger(value.duration_ms) || !isScopedIdArray(value.predecessor_node_ids) ||
    !isSafeString(value.provider_ref, 256) || value.output_preview !== "" || typeof value.failure_code !== "string" ||
    value.failure_code.length > 120) return null;
  return {
    nodeId: value.node_id, nodeType: value.node_type, label: value.label,
    status: value.status as WorkflowRunNodeRecord["status"], startedAt: value.started_at,
    completedAt: value.completed_at, durationMs: value.duration_ms,
    predecessorNodeIds: [...value.predecessor_node_ids], providerRef: value.provider_ref,
    outputPreview: "", failureCode: value.failure_code,
  };
}

function parseAttempt(value: unknown, confirmationId: string): WorkflowHTTPToolExecutionAttempt | null {
  if (!isRecord(value) || !hasExactKeys(value, ATTEMPT_KEYS) || !isPatternString(value.attempt_id, ATTEMPT_ID_PATTERN) ||
    !isPatternString(value.node_id, SCOPED_ID_PATTERN) || !isPatternString(value.tool_id, TOOL_ID_PATTERN) ||
    !isPatternString(value.definition_digest, DIGEST_PATTERN) || !isPatternString(value.profile_id, PROFILE_ID_PATTERN) ||
    !isPatternString(value.profile_digest, DIGEST_PATTERN) || !isPatternString(value.tool_plan_digest, DIGEST_PATTERN) ||
    value.confirmation_id !== confirmationId || typeof value.status !== "string" ||
    !["claimed", "succeeded", "failed", "outcome_unknown"].includes(value.status) ||
    !isTimestamp(value.claimed_at) || typeof value.completed_at !== "string" || value.completed_at.length > 64 ||
    typeof value.http_status_class !== "string" || !["", "2xx", "3xx", "4xx", "5xx"].includes(value.http_status_class) ||
    !isBoundedInteger(value.response_bytes, 0, 65_536) || !isBoundedInteger(value.duration_ms, 0, 30_000) ||
    !isSafeProjection(value.output_projection) || typeof value.failure_code !== "string" || value.failure_code.length > 120) return null;
  return {
    attemptId: value.attempt_id, nodeId: value.node_id, toolId: value.tool_id,
    definitionDigest: value.definition_digest, profileId: value.profile_id, profileDigest: value.profile_digest,
    toolPlanDigest: value.tool_plan_digest, confirmationId, status: value.status as WorkflowHTTPToolExecutionAttempt["status"],
    claimedAt: value.claimed_at, completedAt: value.completed_at, httpStatusClass: value.http_status_class,
    responseBytes: value.response_bytes, durationMs: value.duration_ms,
    outputProjection: { ...value.output_projection }, failureCode: value.failure_code,
  };
}

function parseSideEffects(value: unknown): WorkflowRunRecord["sideEffects"] | null {
  if (!isRecord(value) || !hasExactKeys(value, SIDE_EFFECT_KEYS) || !isBoundedInteger(value.provider_calls, 0, 4) ||
    value.tool_calls !== 1 || value.confirmation_calls !== 1 || value.business_writes !== 0 || value.replay_writes !== 0) return null;
  return { retrievalCalls: 0, providerCalls: value.provider_calls, toolCalls: 1, confirmationCalls: 1, businessWrites: 0, replayWrites: 0 };
}

function parseDiagnostic(value: unknown): WorkflowRunDiagnostic | null {
  if (!isRecord(value) || !hasDiagnosticKeys(value) || typeof value.failure_boundary !== "string" ||
    !["", "draft_read", "executor", "gateway", "provider", "run_store", "request", "tool_policy", "tool_confirmation",
      "tool_transport", "tool_response", "tool_store", "authority", "output_contract"].includes(value.failure_boundary) ||
    !isSafeString(value.failure_stage, 64) || !isSafeString(value.failed_node_id, 160) ||
    !isSafeString(value.last_completed_node_id, 160) || (value.terminal_write_state !== "pending" && value.terminal_write_state !== "stored") ||
    !["none", "queue_full", "timeout", "canceled", "worker_crash", "protocol", "provider_failed", "output_unavailable", "unavailable", "quota"].includes(String(value.gateway_failure_category)) ||
    !["none", "policy", "confirmation", "transport", "timeout", "response_status", "response_too_large", "response_invalid", "store", "outcome_unknown"].includes(String(value.tool_failure_category)) ||
    !isSafeString(value.summary, 256) || !["", "review_draft", "check_gateway_capacity", "check_provider_configuration",
      "check_run_store", "start_new_run", "check_tool_policy", "review_tool_outcome", "review_authority",
      "review_output_contract", "review_cancellation", "review_run"].includes(String(value.recommended_review_action)) ||
    !isTimestamp(value.observed_at) || (value.retrieval_failure_category !== undefined && !isSafeString(value.retrieval_failure_category, 64))) return null;
  return {
    failureBoundary: value.failure_boundary as WorkflowRunDiagnostic["failureBoundary"], failureStage: value.failure_stage,
    failedNodeId: value.failed_node_id, lastCompletedNodeId: value.last_completed_node_id,
    terminalWriteState: value.terminal_write_state, gatewayFailureCategory: value.gateway_failure_category as WorkflowRunDiagnostic["gatewayFailureCategory"],
    toolFailureCategory: value.tool_failure_category as WorkflowRunDiagnostic["toolFailureCategory"], retrievalFailureCategory: "none",
    summary: value.summary, recommendedReviewAction: value.recommended_review_action as WorkflowRunDiagnostic["recommendedReviewAction"],
    observedAt: value.observed_at,
  };
}

function runStateMatches(status: WorkflowRunStatus, failureCode: string, completedAt: string, attempt: WorkflowHTTPToolExecutionAttempt): boolean {
  if (status === "running") return !completedAt && !failureCode && attempt.status === "claimed";
  if (status === "succeeded") return Boolean(completedAt) && !failureCode && attempt.status === "succeeded";
  if (status === "outcome_unknown") return Boolean(completedAt) && failureCode === "workflow_tool_outcome_unknown" && attempt.status === "outcome_unknown";
  return Boolean(completedAt) && Boolean(failureCode) && (attempt.status === "failed" || attempt.status === "succeeded");
}

function hasDiagnosticKeys(value: Record<string, unknown>): boolean {
  const keys = Object.keys(value);
  return keys.length >= DIAGNOSTIC_KEYS.length && keys.length <= DIAGNOSTIC_KEYS.length + 1 &&
    DIAGNOSTIC_KEYS.every((key) => Object.hasOwn(value, key)) &&
    keys.every((key) => DIAGNOSTIC_KEYS.includes(key as typeof DIAGNOSTIC_KEYS[number]) || key === OPTIONAL_DIAGNOSTIC_KEY);
}

function isSafeProjection(value: unknown): value is Record<string, string | number | boolean | null> {
  if (!isRecord(value) || Object.keys(value).length > 16 || Object.keys(value).some((key) => !SAFE_PROJECTION_KEY_PATTERN.test(key))) return false;
  return Object.values(value).every((item) => item === null || typeof item === "boolean" ||
    (typeof item === "number" && Number.isFinite(item)) || isSafeString(item, 4096));
}

function isScopedIdArray(value: unknown): value is string[] {
  return Array.isArray(value) && value.length <= 16 && value.every((item) => isPatternString(item, SCOPED_ID_PATTERN));
}

function isSafeString(value: unknown, maximum: number): value is string {
  return typeof value === "string" && value.length <= maximum && !SENSITIVE_TEXT_PATTERN.test(value);
}

function hasExactKeys(value: Record<string, unknown>, keys: readonly string[]): boolean {
  const expected = new Set(keys);
  return Object.keys(value).length === expected.size && Object.keys(value).every((key) => expected.has(key));
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
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

function isBoundedInteger(value: unknown, minimum: number, maximum: number): value is number {
  return typeof value === "number" && Number.isInteger(value) && value >= minimum && value <= maximum;
}

function isTimestamp(value: unknown): value is string {
  return typeof value === "string" && value.length > 0 && Number.isFinite(Date.parse(value));
}

function isRunStatus(value: unknown): value is WorkflowRunStatus {
  return value === "running" || value === "succeeded" || value === "failed" || value === "canceled" || value === "outcome_unknown";
}
