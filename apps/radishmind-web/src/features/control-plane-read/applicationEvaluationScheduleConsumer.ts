import {
  APPLICATION_EVALUATION_FAILURE_CODES,
  type ApplicationEvaluationConfig,
  type ApplicationEvaluationFailureCode,
  type ApplicationEvaluationPlan,
  type ApplicationEvaluationPlanVersion,
} from "./applicationEvaluationCampaignConsumer.ts";

const SCHEDULE_SCHEMA = "application_evaluation_schedule.v1";
const SCHEDULE_VERSION_SCHEMA = "application_evaluation_schedule_version.v1";
const OCCURRENCE_SCHEMA = "application_evaluation_schedule_occurrence.v1";
const PROMPT_PROFILE = "prompt_application_invocation_v1";
const SCHEDULE_ID = /^aesch_[a-z2-7]{16}$/u;
const PLAN_ID = /^aeplan_[a-z2-7]{16}$/u;
const API_KEY_ID = /^key_[a-z2-7]{16}$/u;
const CAMPAIGN_ID = /^aecamp_[a-z2-7]{16}$/u;
const CAMPAIGN_KEY = /^scheduled_campaign_[a-f0-9]{24}$/u;
const DIGEST = /^sha256:[a-f0-9]{64}$/u;
const REFERENCE = /^[A-Za-z0-9][A-Za-z0-9_.:/-]{2,159}$/u;
const SCOPED_ID = /^[A-Za-z0-9][A-Za-z0-9_.:-]{2,159}$/u;
const UTC_TIMESTAMP = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d{1,9})?Z$/u;

const SCHEDULE_KEYS = [
  "schema_version", "schedule_id", "record_version", "latest_schedule_version", "latest_schedule_digest",
  "tenant_ref", "workspace_id", "environment", "application_id", "plan_id", "plan_version", "plan_digest",
  "execution_profile", "quota_api_key_id", "authorization_model", "system_actor_ref", "delegated_by_user_ref",
  "lifecycle_state", "next_due_at", "created_at", "updated_at", "created_by_actor_ref", "updated_by_actor_ref",
  "request_id", "audit_ref",
] as const;
const VERSION_KEYS = [
  "schema_version", "schedule_id", "schedule_version", "previous_schedule_version", "schedule_digest",
  "tenant_ref", "workspace_id", "environment", "application_id", "plan_id", "plan_version", "plan_digest",
  "execution_profile", "quota_api_key_id", "schedule", "item_count", "max_provider_attempts",
  "missed_window_policy", "overlap_policy", "authorization", "created_at", "created_by_actor_ref", "request_id", "audit_ref",
] as const;
const OCCURRENCE_KEYS = [
  "schema_version", "record_version", "tenant_ref", "workspace_id", "environment", "application_id",
  "schedule_id", "schedule_version", "schedule_digest", "scheduled_for_utc", "state", "client_campaign_key",
  "campaign_id", "system_actor_ref", "delegated_by_user_ref", "claimed_at", "failure_code",
  "created_at", "updated_at", "completed_at", "request_id", "audit_ref",
] as const;
const FORBIDDEN_FIELDS = new Set([
  "api_key", "api_key_token", "access_token", "refresh_token", "credential", "secret", "private_key",
  "cookie", "headers", "authorization_header", "dsn", "endpoint", "raw_request", "raw_response",
  "provider_raw_response", "provider_raw_envelope", "fixture", "input", "answer", "output", "messages",
]);
const OCCURRENCE_FAILURE_CODES = new Set([
  "application_evaluation_schedule_authorization_unavailable",
  "application_evaluation_schedule_membership_denied",
  "application_evaluation_schedule_plan_changed",
  "application_evaluation_schedule_authority_changed",
  "application_evaluation_schedule_quota_consumer_invalid",
  "application_evaluation_schedule_quota_denied",
  "application_evaluation_schedule_overlap_blocked",
  "application_evaluation_schedule_missed_window",
  "application_evaluation_schedule_claim_conflict",
  "application_evaluation_schedule_campaign_failed",
  "application_evaluation_schedule_campaign_interrupted",
  "application_evaluation_schedule_store_unavailable",
  "application_evaluation_schedule_store_contract_mismatch",
]);

export const APPLICATION_EVALUATION_SCHEDULE_FAILURE_CODES = [
  ...APPLICATION_EVALUATION_FAILURE_CODES,
  "application_evaluation_schedule_authorization_unavailable",
  "application_evaluation_schedule_membership_denied",
  "application_evaluation_schedule_plan_changed",
  "application_evaluation_schedule_authority_changed",
  "application_evaluation_schedule_quota_consumer_invalid",
  "application_evaluation_schedule_quota_denied",
  "application_evaluation_schedule_overlap_blocked",
  "application_evaluation_schedule_missed_window",
  "application_evaluation_schedule_claim_conflict",
  "application_evaluation_schedule_campaign_failed",
  "application_evaluation_schedule_campaign_interrupted",
  "application_evaluation_schedule_store_unavailable",
  "application_evaluation_schedule_store_contract_mismatch",
  "application_evaluation_schedule_not_found",
  "application_evaluation_schedule_version_conflict",
] as const;

export type ApplicationEvaluationScheduleFailureCode =
  | ApplicationEvaluationFailureCode
  | Exclude<typeof APPLICATION_EVALUATION_SCHEDULE_FAILURE_CODES[number], ApplicationEvaluationFailureCode>;
export type ApplicationEvaluationScheduleState = "draft" | "active" | "paused" | "archived";
export type ApplicationEvaluationScheduleOccurrenceState =
  | "due" | "claimed" | "campaign_created" | "observing" | "succeeded" | "failed" | "interrupted" | "skipped";
export type ApplicationEvaluationScheduleAction = "activate" | "pause" | "resume" | "archive";

export type ApplicationEvaluationSchedule = {
  schemaVersion: typeof SCHEDULE_SCHEMA;
  scheduleId: string;
  recordVersion: number;
  latestScheduleVersion: number;
  latestScheduleDigest: string;
  tenantRef: string;
  workspaceId: string;
  environment: "development" | "test";
  applicationId: string;
  planId: string;
  planVersion: number;
  planDigest: string;
  executionProfile: typeof PROMPT_PROFILE;
  quotaAPIKeyId: string;
  authorizationModel: "system_actor_schedule_scoped_delegation_v1";
  systemActorRef: string;
  delegatedByUserRef: string;
  lifecycleState: ApplicationEvaluationScheduleState;
  nextDueAt: string | null;
  createdAt: string;
  updatedAt: string;
  createdByActorRef: string;
  updatedByActorRef: string;
  requestId: string;
  auditRef: string;
};

export type ApplicationEvaluationScheduleVersion = {
  schemaVersion: typeof SCHEDULE_VERSION_SCHEMA;
  scheduleId: string;
  scheduleVersion: number;
  previousScheduleVersion: number;
  scheduleDigest: string;
  tenantRef: string;
  workspaceId: string;
  environment: "development" | "test";
  applicationId: string;
  planId: string;
  planVersion: number;
  planDigest: string;
  executionProfile: typeof PROMPT_PROFILE;
  quotaAPIKeyId: string;
  schedule: { rule: "daily_utc"; hour: number; minute: number };
  itemCount: number;
  maxProviderAttempts: number;
  missedWindowPolicy: "record_only_no_catch_up";
  overlapPolicy: "skip_while_campaign_non_terminal";
  authorization: {
    model: "system_actor_schedule_scoped_delegation_v1";
    systemActorRef: string;
    delegatedByUserRef: string;
    requiredPermissions: ["application_evaluations:execute", "workflow_runs:execute"];
    revalidationPolicy: "every_occurrence";
    apiKeyOwnershipPolicy: "delegated_user_current_owner";
    revocationPolicy: "fail_closed_immediate";
  };
  createdAt: string;
  createdByActorRef: string;
  requestId: string;
  auditRef: string;
};

export type ApplicationEvaluationScheduleOccurrence = {
  schemaVersion: typeof OCCURRENCE_SCHEMA;
  recordVersion: number;
  tenantRef: string;
  workspaceId: string;
  environment: "development" | "test";
  applicationId: string;
  scheduleId: string;
  scheduleVersion: number;
  scheduleDigest: string;
  scheduledForUTC: string;
  state: ApplicationEvaluationScheduleOccurrenceState;
  clientCampaignKey: string;
  campaignId: string | null;
  systemActorRef: string;
  delegatedByUserRef: string;
  claimedAt: string | null;
  failureCode: ApplicationEvaluationScheduleFailureCode | null;
  createdAt: string;
  updatedAt: string;
  completedAt: string | null;
  requestId: string;
  auditRef: string;
};

type ScopeEnvelope = {
  requestId: string;
  tenantRef: string;
  workspaceId: string;
  environment: "development" | "test";
  applicationId: string;
  failureCode: ApplicationEvaluationScheduleFailureCode | null;
  failureSummary: string;
  auditRef: string;
};

export type ApplicationEvaluationScheduleEnvelope = ScopeEnvelope & {
  schedule: ApplicationEvaluationSchedule | null;
  version: ApplicationEvaluationScheduleVersion | null;
  occurrence: ApplicationEvaluationScheduleOccurrence | null;
  currentRecordVersion: number;
  currentState: string;
};

export type ApplicationEvaluationScheduleListEnvelope = ScopeEnvelope & {
  schedules: ApplicationEvaluationSchedule[];
  nextCursor: string;
  hasMore: boolean;
};

type Document = Record<string, unknown>;

export async function listApplicationEvaluationSchedules(
  config: ApplicationEvaluationConfig,
  signal?: AbortSignal,
  lifecycleState: ApplicationEvaluationScheduleState = "active",
): Promise<ApplicationEvaluationScheduleListEnvelope> {
  if (config.mode === "offline") return offlineScheduleList(config);
  if (!isScheduleState(lifecycleState)) throw new Error("Schedule lifecycle filter is invalid.");
  const query = new URLSearchParams({
    workspace_id: config.workspaceId,
    environment: config.environment,
    lifecycle_state: lifecycleState,
    limit: "100",
  });
  const envelope = await request(
    config,
    `evaluation-schedules?${query}`,
    "GET",
    null,
    ["application_evaluations:read"],
    decodeScheduleListEnvelope,
    signal,
  );
  if (envelope.schedules.some((schedule) => schedule.lifecycleState !== lifecycleState)) {
    throw new Error("Schedule lifecycle projection does not match its requested owner.");
  }
  return envelope;
}

export async function readApplicationEvaluationSchedule(
  config: ApplicationEvaluationConfig,
  scheduleId: string,
  signal?: AbortSignal,
): Promise<ApplicationEvaluationScheduleEnvelope> {
  assertScheduleId(scheduleId);
  return request(config, `evaluation-schedules/${scheduleId}?${scopeQuery(config)}`, "GET", null,
    ["application_evaluations:read"], decodeScheduleEnvelope, signal);
}

export async function readApplicationEvaluationScheduleVersion(
  config: ApplicationEvaluationConfig,
  scheduleId: string,
  version: number,
  signal?: AbortSignal,
): Promise<ApplicationEvaluationScheduleEnvelope> {
  assertScheduleId(scheduleId);
  if (!positiveInteger(version)) throw new Error("Schedule version is invalid.");
  return request(config, `evaluation-schedules/${scheduleId}/versions/${version}?${scopeQuery(config)}`, "GET", null,
    ["application_evaluations:read"], decodeScheduleEnvelope, signal);
}

export async function readApplicationEvaluationScheduleOccurrence(
  config: ApplicationEvaluationConfig,
  scheduleId: string,
  scheduleVersion: number,
  scheduledForUTC: string,
  signal?: AbortSignal,
): Promise<ApplicationEvaluationScheduleEnvelope> {
  assertScheduleId(scheduleId);
  if (!positiveInteger(scheduleVersion) || !validTimestamp(scheduledForUTC)) {
    throw new Error("Exact schedule version and canonical UTC occurrence time are required.");
  }
  const path = `evaluation-schedules/${scheduleId}/occurrences/${scheduleVersion}/${encodeURIComponent(scheduledForUTC)}`;
  return request(config, `${path}?${scopeQuery(config)}`, "GET", null, ["application_evaluations:read"], decodeScheduleEnvelope, signal);
}

export async function createApplicationEvaluationSchedule(
  config: ApplicationEvaluationConfig,
  input: { plan: ApplicationEvaluationPlan; version: ApplicationEvaluationPlanVersion; quotaAPIKeyId: string; hour: number; minute: number },
  signal?: AbortSignal,
): Promise<ApplicationEvaluationScheduleEnvelope> {
  assertDefinitionInput(input);
  return request(config, "evaluation-schedules", "POST", definitionBody(config, input), scheduleMutationScopes(), decodeScheduleEnvelope, signal);
}

export async function reviseApplicationEvaluationSchedule(
  config: ApplicationEvaluationConfig,
  schedule: ApplicationEvaluationSchedule,
  input: { plan: ApplicationEvaluationPlan; version: ApplicationEvaluationPlanVersion; quotaAPIKeyId: string; hour: number; minute: number },
  signal?: AbortSignal,
): Promise<ApplicationEvaluationScheduleEnvelope> {
  assertScheduleId(schedule.scheduleId);
  assertDefinitionInput(input);
  return request(config, `evaluation-schedules/${schedule.scheduleId}/revisions`, "POST", {
    ...definitionBody(config, input), expected_version: schedule.recordVersion,
  }, scheduleMutationScopes(), decodeScheduleEnvelope, signal);
}

export async function transitionApplicationEvaluationSchedule(
  config: ApplicationEvaluationConfig,
  schedule: ApplicationEvaluationSchedule,
  action: ApplicationEvaluationScheduleAction,
  signal?: AbortSignal,
): Promise<ApplicationEvaluationScheduleEnvelope> {
  assertScheduleId(schedule.scheduleId);
  if (!positiveInteger(schedule.recordVersion)) throw new Error("Schedule record version is invalid.");
  return request(config, `evaluation-schedules/${schedule.scheduleId}/${action}`, "POST", {
    workspace_id: config.workspaceId,
    environment: config.environment,
    expected_version: schedule.recordVersion,
    acknowledge_provider_consumption: action === "activate" || action === "resume",
    acknowledge_no_future_occurrences: action === "archive",
  }, scheduleMutationScopes(), decodeScheduleEnvelope, signal);
}

function definitionBody(
  config: ApplicationEvaluationConfig,
  input: { plan: ApplicationEvaluationPlan; version: ApplicationEvaluationPlanVersion; quotaAPIKeyId: string; hour: number; minute: number },
): Document {
  return {
    workspace_id: config.workspaceId,
    environment: config.environment,
    plan_id: input.plan.planId,
    plan_version: input.version.planVersion,
    plan_digest: input.version.planDigest,
    expected_plan_record_version: input.plan.recordVersion,
    quota_api_key_id: input.quotaAPIKeyId.trim(),
    schedule: { rule: "daily_utc", hour: input.hour, minute: input.minute },
    acknowledge_provider_consumption: true,
  };
}

function assertDefinitionInput(input: {
  plan: ApplicationEvaluationPlan;
  version: ApplicationEvaluationPlanVersion;
  quotaAPIKeyId: string;
  hour: number;
  minute: number;
}) {
  if (!PLAN_ID.test(input.plan.planId) || input.plan.executionProfile !== PROMPT_PROFILE ||
    input.version.executionProfile !== PROMPT_PROFILE || input.plan.planId !== input.version.planId ||
    input.plan.latestPlanVersion !== input.version.planVersion || input.plan.latestPlanDigest !== input.version.planDigest ||
    !DIGEST.test(input.version.planDigest) || !API_KEY_ID.test(input.quotaAPIKeyId.trim()) ||
    !Number.isInteger(input.hour) || input.hour < 0 || input.hour > 23 ||
    !Number.isInteger(input.minute) || input.minute < 0 || input.minute > 59) {
    throw new Error("Schedule requires one exact active Prompt plan, actor-owned API key id, and daily UTC time.");
  }
}

async function request<T>(
  config: ApplicationEvaluationConfig,
  suffix: string,
  method: "GET" | "POST",
  body: Document | null,
  permissions: string[],
  decoder: (value: unknown, config: ApplicationEvaluationConfig) => T,
  signal?: AbortSignal,
): Promise<T> {
  assertConfig(config);
  if (config.mode === "offline") throw new Error("Application evaluation schedule HTTP is disabled.");
  const requestId = `application-evaluation-schedule-${Date.now().toString(36)}`;
  const response = await fetch(`${config.baseUrl}/v1/user-workspace/applications/${encodeURIComponent(config.applicationId)}/${suffix}`, {
    method,
    credentials: config.authMode === "local_session_dev_test" ? "include" : "omit",
    cache: "no-store",
    headers: evaluationHeaders(config, requestId, permissions, body !== null),
    body: body === null ? undefined : JSON.stringify(body),
    signal,
  });
  let value: unknown;
  try {
    value = await response.json();
  } catch {
    throw new Error(`Application evaluation schedule returned invalid HTTP ${response.status} JSON.`);
  }
  assertNoForbiddenFields(value);
  try {
    return decoder(value, config);
  } catch (error) {
    const detail = error instanceof Error ? error.message : "strict validation failed";
    throw new Error(`Application evaluation schedule returned an invalid HTTP ${response.status} envelope: ${detail}`);
  }
}

function evaluationHeaders(
  config: ApplicationEvaluationConfig,
  requestId: string,
  permissions: string[],
  includeBody: boolean,
): Record<string, string> {
  const scopes = [...new Set(permissions)].sort().join(",");
  if (config.authMode === "local_session_dev_test") {
    return {
      Accept: "application/json",
      ...(includeBody ? { "Content-Type": "application/json" } : {}),
      "X-Request-Id": requestId,
      "X-RadishMind-Active-Tenant": config.tenantRef,
      "X-RadishMind-Active-Workspace": config.workspaceId,
      "X-RadishMind-Dev-Workflow-Workspace": config.workspaceId,
      "X-RadishMind-Dev-Workflow-Application": config.applicationId,
      "X-RadishMind-Dev-Application-Evaluation-Environment": config.environment,
    };
  }
  return {
    Accept: "application/json",
    ...(includeBody ? { "Content-Type": "application/json" } : {}),
    "X-Request-Id": requestId,
    "X-RadishMind-Dev-Read-Identity": "radishmind-web-application-evaluation-schedule",
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Read-Scopes": scopes,
    "X-RadishMind-Dev-Read-Audit": `audit_${requestId}`,
    "X-RadishMind-Active-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Read-Membership-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Read-Membership-Permissions": scopes,
    "X-RadishMind-Dev-Workflow-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Workflow-Application": config.applicationId,
    "X-RadishMind-Dev-Application-Evaluation-Environment": config.environment,
  };
}

function decodeScheduleListEnvelope(value: unknown, config: ApplicationEvaluationConfig): ApplicationEvaluationScheduleListEnvelope {
  const keys = ["request_id", "tenant_ref", "workspace_id", "environment", "application_id", "schedules", "next_cursor", "has_more", "failure_code", "failure_summary", "audit_ref"];
  if (!isExactDocument(value, keys) || !Array.isArray(value.schedules) ||
    !value.schedules.every((schedule) => isScheduleDocument(schedule, config)) || typeof value.next_cursor !== "string" ||
    typeof value.has_more !== "boolean") throw new Error("schedule list contract mismatch");
  return { ...mapScope(value, config), schedules: value.schedules.map(mapSchedule), nextCursor: value.next_cursor, hasMore: value.has_more };
}

function decodeScheduleEnvelope(value: unknown, config: ApplicationEvaluationConfig): ApplicationEvaluationScheduleEnvelope {
  const keys = ["request_id", "tenant_ref", "workspace_id", "environment", "application_id", "schedule", "version", "occurrence", "failure_code", "failure_summary", "current_record_version", "current_state", "audit_ref"];
  if (!isExactDocument(value, keys) || !(value.schedule === null || isScheduleDocument(value.schedule, config)) ||
    !(value.version === null || isVersionDocument(value.version, config)) ||
    !(value.occurrence === null || isOccurrenceDocument(value.occurrence, config)) ||
    !nonNegativeInteger(value.current_record_version) || typeof value.current_state !== "string") {
    throw new Error("schedule envelope contract mismatch");
  }
  const scope = mapScope(value, config);
  if (scope.failureCode === null && value.schedule === null && value.version === null && value.occurrence === null) {
    throw new Error("successful schedule envelope is empty");
  }
  return {
    ...scope,
    schedule: value.schedule === null ? null : mapSchedule(value.schedule),
    version: value.version === null ? null : mapVersion(value.version),
    occurrence: value.occurrence === null ? null : mapOccurrence(value.occurrence),
    currentRecordVersion: Number(value.current_record_version),
    currentState: String(value.current_state),
  };
}

function mapScope(value: Document, config: ApplicationEvaluationConfig): ScopeEnvelope {
  const failureCode = nullableFailureCode(value.failure_code);
  if (failureCode === undefined || typeof value.request_id !== "string" || typeof value.failure_summary !== "string" ||
    typeof value.audit_ref !== "string" || (value.environment !== config.environment && !(failureCode && value.environment === "")) ||
    (value.application_id !== config.applicationId && !(failureCode && value.application_id === "")) ||
    !matchesScope(value.tenant_ref, config.tenantRef, failureCode) || !matchesScope(value.workspace_id, config.workspaceId, failureCode)) {
    throw new Error("scope or failure contract mismatch");
  }
  return {
    requestId: value.request_id,
    tenantRef: value.tenant_ref === "" ? config.tenantRef : String(value.tenant_ref),
    workspaceId: value.workspace_id === "" ? config.workspaceId : String(value.workspace_id),
    environment: config.environment,
    applicationId: config.applicationId,
    failureCode,
    failureSummary: value.failure_summary,
    auditRef: value.audit_ref,
  };
}

function isScheduleDocument(value: unknown, config: ApplicationEvaluationConfig): value is Document {
  if (!isExactDocument(value, SCHEDULE_KEYS)) return false;
  const state = value.lifecycle_state;
  const nextDueValid = state === "active" ? validTimestamp(value.next_due_at) : value.next_due_at === null;
  return value.schema_version === SCHEDULE_SCHEMA && typeof value.schedule_id === "string" && SCHEDULE_ID.test(value.schedule_id) &&
    positiveInteger(value.record_version) && positiveInteger(value.latest_schedule_version) && digest(value.latest_schedule_digest) &&
    scopeDocumentMatches(value, config) && typeof value.plan_id === "string" && PLAN_ID.test(value.plan_id) &&
    positiveInteger(value.plan_version) && digest(value.plan_digest) && value.execution_profile === PROMPT_PROFILE &&
    typeof value.quota_api_key_id === "string" && API_KEY_ID.test(value.quota_api_key_id) &&
    value.authorization_model === "system_actor_schedule_scoped_delegation_v1" && reference(value.system_actor_ref) &&
    reference(value.delegated_by_user_ref) && isScheduleState(state) && nextDueValid && validTimestamp(value.created_at) &&
    validTimestamp(value.updated_at) && reference(value.created_by_actor_ref) && reference(value.updated_by_actor_ref) &&
    reference(value.request_id) && reference(value.audit_ref);
}

function isVersionDocument(value: unknown, config: ApplicationEvaluationConfig): value is Document {
  if (!isExactDocument(value, VERSION_KEYS)) return false;
  return value.schema_version === SCHEDULE_VERSION_SCHEMA && typeof value.schedule_id === "string" && SCHEDULE_ID.test(value.schedule_id) &&
    positiveInteger(value.schedule_version) && nonNegativeInteger(value.previous_schedule_version) && digest(value.schedule_digest) &&
    scopeDocumentMatches(value, config) && typeof value.plan_id === "string" && PLAN_ID.test(value.plan_id) &&
    positiveInteger(value.plan_version) && digest(value.plan_digest) && value.execution_profile === PROMPT_PROFILE &&
    typeof value.quota_api_key_id === "string" && API_KEY_ID.test(value.quota_api_key_id) && isDailyUTC(value.schedule) &&
    positiveInteger(value.item_count) && Number(value.item_count) <= 20 && value.max_provider_attempts === value.item_count &&
    value.missed_window_policy === "record_only_no_catch_up" && value.overlap_policy === "skip_while_campaign_non_terminal" &&
    isAuthorization(value.authorization) && validTimestamp(value.created_at) && reference(value.created_by_actor_ref) &&
    reference(value.request_id) && reference(value.audit_ref);
}

function isOccurrenceDocument(value: unknown, config: ApplicationEvaluationConfig): value is Document {
  if (!isExactDocument(value, OCCURRENCE_KEYS)) return false;
  if (value.schema_version !== OCCURRENCE_SCHEMA || !positiveInteger(value.record_version) || !scopeDocumentMatches(value, config) ||
    typeof value.schedule_id !== "string" || !SCHEDULE_ID.test(value.schedule_id) || !positiveInteger(value.schedule_version) ||
    !digest(value.schedule_digest) || !validTimestamp(value.scheduled_for_utc) || !isOccurrenceState(value.state) ||
    typeof value.client_campaign_key !== "string" || !CAMPAIGN_KEY.test(value.client_campaign_key) ||
    !(value.campaign_id === null || typeof value.campaign_id === "string" && CAMPAIGN_ID.test(value.campaign_id)) ||
    !reference(value.system_actor_ref) || !reference(value.delegated_by_user_ref) ||
    !(value.claimed_at === null || validTimestamp(value.claimed_at)) || nullableOccurrenceFailureCode(value.failure_code) === undefined ||
    !validTimestamp(value.created_at) || !validTimestamp(value.updated_at) ||
    !(value.completed_at === null || validTimestamp(value.completed_at)) || !reference(value.request_id) || !reference(value.audit_ref)) return false;
  const campaign = value.campaign_id;
  const claimed = value.claimed_at;
  const failure = value.failure_code;
  const completed = value.completed_at;
  switch (value.state) {
  case "due": return value.record_version === 1 && campaign === null && claimed === null && failure === null && completed === null;
  case "claimed": return campaign === null && claimed !== null && failure === null && completed === null;
  case "campaign_created":
  case "observing": return campaign !== null && claimed !== null && failure === null && completed === null;
  case "succeeded": return campaign !== null && claimed !== null && failure === null && completed !== null;
  case "failed":
  case "interrupted": return claimed !== null && failure !== null && completed !== null;
  case "skipped": return campaign === null && claimed !== null &&
      (failure === "application_evaluation_schedule_overlap_blocked" || failure === "application_evaluation_schedule_missed_window") && completed !== null;
  }
}

function isDailyUTC(value: unknown): value is Document {
  return isExactDocument(value, ["rule", "hour", "minute"]) && value.rule === "daily_utc" &&
    integerInRange(value.hour, 0, 23) && integerInRange(value.minute, 0, 59);
}

function isAuthorization(value: unknown): value is Document {
  return isExactDocument(value, ["model", "system_actor_ref", "delegated_by_user_ref", "required_permissions", "revalidation_policy", "api_key_ownership_policy", "revocation_policy"]) &&
    value.model === "system_actor_schedule_scoped_delegation_v1" && reference(value.system_actor_ref) &&
    reference(value.delegated_by_user_ref) && Array.isArray(value.required_permissions) && value.required_permissions.length === 2 &&
    value.required_permissions[0] === "application_evaluations:execute" && value.required_permissions[1] === "workflow_runs:execute" &&
    value.revalidation_policy === "every_occurrence" && value.api_key_ownership_policy === "delegated_user_current_owner" &&
    value.revocation_policy === "fail_closed_immediate";
}

function mapSchedule(value: Document): ApplicationEvaluationSchedule {
  return {
    schemaVersion: SCHEDULE_SCHEMA, scheduleId: String(value.schedule_id), recordVersion: Number(value.record_version),
    latestScheduleVersion: Number(value.latest_schedule_version), latestScheduleDigest: String(value.latest_schedule_digest),
    tenantRef: String(value.tenant_ref), workspaceId: String(value.workspace_id), environment: value.environment as "development" | "test",
    applicationId: String(value.application_id), planId: String(value.plan_id), planVersion: Number(value.plan_version),
    planDigest: String(value.plan_digest), executionProfile: PROMPT_PROFILE, quotaAPIKeyId: String(value.quota_api_key_id),
    authorizationModel: "system_actor_schedule_scoped_delegation_v1", systemActorRef: String(value.system_actor_ref),
    delegatedByUserRef: String(value.delegated_by_user_ref), lifecycleState: value.lifecycle_state as ApplicationEvaluationScheduleState,
    nextDueAt: value.next_due_at === null ? null : String(value.next_due_at), createdAt: String(value.created_at),
    updatedAt: String(value.updated_at), createdByActorRef: String(value.created_by_actor_ref),
    updatedByActorRef: String(value.updated_by_actor_ref), requestId: String(value.request_id), auditRef: String(value.audit_ref),
  };
}

function mapVersion(value: Document): ApplicationEvaluationScheduleVersion {
  const schedule = value.schedule as Document;
  const authorization = value.authorization as Document;
  return {
    schemaVersion: SCHEDULE_VERSION_SCHEMA, scheduleId: String(value.schedule_id), scheduleVersion: Number(value.schedule_version),
    previousScheduleVersion: Number(value.previous_schedule_version), scheduleDigest: String(value.schedule_digest),
    tenantRef: String(value.tenant_ref), workspaceId: String(value.workspace_id), environment: value.environment as "development" | "test",
    applicationId: String(value.application_id), planId: String(value.plan_id), planVersion: Number(value.plan_version),
    planDigest: String(value.plan_digest), executionProfile: PROMPT_PROFILE, quotaAPIKeyId: String(value.quota_api_key_id),
    schedule: { rule: "daily_utc", hour: Number(schedule.hour), minute: Number(schedule.minute) }, itemCount: Number(value.item_count),
    maxProviderAttempts: Number(value.max_provider_attempts), missedWindowPolicy: "record_only_no_catch_up",
    overlapPolicy: "skip_while_campaign_non_terminal",
    authorization: {
      model: "system_actor_schedule_scoped_delegation_v1", systemActorRef: String(authorization.system_actor_ref),
      delegatedByUserRef: String(authorization.delegated_by_user_ref),
      requiredPermissions: ["application_evaluations:execute", "workflow_runs:execute"], revalidationPolicy: "every_occurrence",
      apiKeyOwnershipPolicy: "delegated_user_current_owner", revocationPolicy: "fail_closed_immediate",
    },
    createdAt: String(value.created_at), createdByActorRef: String(value.created_by_actor_ref),
    requestId: String(value.request_id), auditRef: String(value.audit_ref),
  };
}

function mapOccurrence(value: Document): ApplicationEvaluationScheduleOccurrence {
  return {
    schemaVersion: OCCURRENCE_SCHEMA, recordVersion: Number(value.record_version), tenantRef: String(value.tenant_ref),
    workspaceId: String(value.workspace_id), environment: value.environment as "development" | "test",
    applicationId: String(value.application_id), scheduleId: String(value.schedule_id), scheduleVersion: Number(value.schedule_version),
    scheduleDigest: String(value.schedule_digest), scheduledForUTC: String(value.scheduled_for_utc),
    state: value.state as ApplicationEvaluationScheduleOccurrenceState, clientCampaignKey: String(value.client_campaign_key),
    campaignId: value.campaign_id === null ? null : String(value.campaign_id), systemActorRef: String(value.system_actor_ref),
    delegatedByUserRef: String(value.delegated_by_user_ref), claimedAt: value.claimed_at === null ? null : String(value.claimed_at),
    failureCode: nullableOccurrenceFailureCode(value.failure_code) ?? null, createdAt: String(value.created_at), updatedAt: String(value.updated_at),
    completedAt: value.completed_at === null ? null : String(value.completed_at), requestId: String(value.request_id), auditRef: String(value.audit_ref),
  };
}

function offlineScheduleList(config: ApplicationEvaluationConfig): ApplicationEvaluationScheduleListEnvelope {
  return {
    requestId: "offline_application_evaluation_schedule", tenantRef: config.tenantRef, workspaceId: config.workspaceId,
    environment: config.environment, applicationId: config.applicationId, schedules: [], nextCursor: "", hasMore: false,
    failureCode: "application_evaluation_write_disabled", failureSummary: "Schedule HTTP requires explicit development/test opt-in.",
    auditRef: "offline_application_evaluation_schedule",
  };
}

function scheduleMutationScopes(): string[] {
  return ["application_evaluations:execute", "workflow_runs:execute"];
}

function assertConfig(config: ApplicationEvaluationConfig) {
  if (!REFERENCE.test(config.tenantRef) || !SCOPED_ID.test(config.workspaceId) || !SCOPED_ID.test(config.applicationId) ||
    !REFERENCE.test(config.subjectRef) || (config.environment !== "development" && config.environment !== "test") ||
    (config.authMode !== undefined && config.authMode !== "dev_headers" && config.authMode !== "local_session_dev_test")) {
    throw new Error("Application evaluation schedule scope is invalid.");
  }
}

function assertScheduleId(value: string) {
  if (!SCHEDULE_ID.test(value)) throw new Error("Application evaluation schedule reference is invalid.");
}

function assertNoForbiddenFields(value: unknown, path = "response") {
  if (Array.isArray(value)) {
    value.forEach((item, index) => assertNoForbiddenFields(item, `${path}[${index}]`));
    return;
  }
  if (!isDocument(value)) return;
  for (const [key, child] of Object.entries(value)) {
    if (FORBIDDEN_FIELDS.has(key.toLowerCase())) throw new Error(`forbidden field ${path}.${key}`);
    assertNoForbiddenFields(child, `${path}.${key}`);
  }
}

function scopeQuery(config: ApplicationEvaluationConfig): string {
  return new URLSearchParams({ workspace_id: config.workspaceId, environment: config.environment }).toString();
}

function scopeDocumentMatches(value: Document, config: ApplicationEvaluationConfig): boolean {
  return value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId &&
    value.environment === config.environment && value.application_id === config.applicationId;
}

function nullableFailureCode(value: unknown): ApplicationEvaluationScheduleFailureCode | null | undefined {
  if (value === null) return null;
  return typeof value === "string" && APPLICATION_EVALUATION_SCHEDULE_FAILURE_CODES.includes(value as ApplicationEvaluationScheduleFailureCode)
    ? value as ApplicationEvaluationScheduleFailureCode
    : undefined;
}

function nullableOccurrenceFailureCode(value: unknown): ApplicationEvaluationScheduleFailureCode | null | undefined {
  if (value === null) return null;
  return typeof value === "string" && OCCURRENCE_FAILURE_CODES.has(value)
    ? value as ApplicationEvaluationScheduleFailureCode
    : undefined;
}

function matchesScope(value: unknown, expected: string, failure: ApplicationEvaluationScheduleFailureCode | null): boolean {
  return value === expected || (failure !== null && value === "");
}

function isScheduleState(value: unknown): value is ApplicationEvaluationScheduleState {
  return value === "draft" || value === "active" || value === "paused" || value === "archived";
}

function isOccurrenceState(value: unknown): value is ApplicationEvaluationScheduleOccurrenceState {
  return value === "due" || value === "claimed" || value === "campaign_created" || value === "observing" ||
    value === "succeeded" || value === "failed" || value === "interrupted" || value === "skipped";
}

function digest(value: unknown): boolean {
  return typeof value === "string" && DIGEST.test(value);
}

function reference(value: unknown): boolean {
  return typeof value === "string" && REFERENCE.test(value);
}

function validTimestamp(value: unknown): value is string {
  return typeof value === "string" && UTC_TIMESTAMP.test(value) && !Number.isNaN(Date.parse(value));
}

function positiveInteger(value: unknown): boolean {
  return Number.isInteger(value) && Number(value) > 0;
}

function nonNegativeInteger(value: unknown): boolean {
  return Number.isInteger(value) && Number(value) >= 0;
}

function integerInRange(value: unknown, minimum: number, maximum: number): boolean {
  return Number.isInteger(value) && Number(value) >= minimum && Number(value) <= maximum;
}

function isDocument(value: unknown): value is Document {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function isExactDocument(value: unknown, keys: readonly string[]): value is Document {
  if (!isDocument(value)) return false;
  const actual = Object.keys(value).sort();
  const expected = [...keys].sort();
  return actual.length === expected.length && actual.every((key, index) => key === expected[index]);
}
