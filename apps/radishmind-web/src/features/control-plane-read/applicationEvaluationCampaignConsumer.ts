import {
  decodeWorkflowRunComparison,
  type WorkflowRunComparison,
} from "./workflowRunComparisonConsumer.ts";
import {
  parseStructuredRuntimeInputContractDocument,
  type StructuredRuntimeInputContract,
  type StructuredRuntimeInputValues,
} from "./structuredRuntimeInput.ts";

const DEV_SOURCE = "dev-application-evaluation-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const PLAN_SCHEMA = "application_evaluation_plan.v1";
const PLAN_VERSION_SCHEMA = "application_evaluation_plan_version.v1";
const CAMPAIGN_SCHEMA = "application_evaluation_campaign.v1";
const STRUCTURED_PLAN_SCHEMA = "application_evaluation_plan.v2";
const STRUCTURED_PLAN_VERSION_SCHEMA = "application_evaluation_plan_version.v2";
const STRUCTURED_CAMPAIGN_SCHEMA = "application_evaluation_campaign.v2";
const IDENTIFIER = /^[A-Za-z0-9][A-Za-z0-9._:-]{2,255}$/u;
const WORKSPACE_IDENTIFIER = /^[A-Za-z0-9][A-Za-z0-9_.:/-]{2,159}$/u;
const DIGEST = /^sha256:[a-f0-9]{64}$/u;

const PLAN_KEYS = [
  "schema_version", "plan_id", "record_version", "latest_plan_version", "latest_plan_digest",
  "tenant_ref", "workspace_id", "environment", "application_id", "name", "execution_profile",
  "item_count", "lifecycle_state", "created_at", "updated_at", "created_by_actor_ref",
  "updated_by_actor_ref", "request_id", "audit_ref",
] as const;
const VERSION_KEYS = [
  "schema_version", "plan_id", "plan_version", "previous_plan_version", "plan_digest", "tenant_ref",
  "workspace_id", "environment", "application_id", "name", "execution_profile", "target", "items",
  "created_at", "created_by_actor_ref", "request_id", "audit_ref",
] as const;
const PLAN_ITEM_KEYS = [
  "item_key", "name", "expected_classification", "workflow_definition", "application_rag",
  "prompt_application", "agent_copilot",
] as const;
const CAMPAIGN_KEYS = [
  "schema_version", "campaign_id", "client_campaign_key", "record_version", "tenant_ref", "workspace_id",
  "environment", "application_id", "plan_id", "plan_version", "plan_digest", "execution_profile",
  "quota_api_key_id", "authority", "state", "current_item_index", "succeeded_items", "failed_items",
  "failure_code", "failure_summary", "items", "handoff", "created_at", "started_at", "completed_at",
  "created_by_actor_ref", "updated_by_actor_ref", "request_id", "audit_ref",
] as const;
const CAMPAIGN_ITEM_KEYS = [
  "item_key", "run_id", "state", "run_schema_version", "run_profile", "authority_digest",
  "failure_code", "failure_boundary", "started_at", "completed_at",
] as const;
const HANDOFF_KEYS = [
  "baseline_campaign_id", "candidate_campaign_id", "case_refs", "suite_id", "state", "audit_ref",
] as const;
const PAIR_REVIEW_KEYS = [
  "plan_id", "plan_name", "plan_version", "plan_digest", "execution_profile", "baseline_campaign_id",
  "candidate_campaign_id", "expected_matches", "expected_mismatches", "items", "existing_handoff",
] as const;
const PAIR_ITEM_KEYS = [
  "item_key", "name", "baseline_run_id", "candidate_run_id", "expected_classification",
  "actual_classification", "expectation_matched", "comparison",
] as const;

const FORBIDDEN_FIELDS = new Set([
  "authorization", "api_key", "api_key_token", "access_token", "refresh_token", "credential", "secret",
  "private_key", "cookie", "headers", "dsn", "endpoint", "raw_request", "raw_response",
  "provider_raw_response", "provider_raw_envelope", "answer", "output", "messages",
]);

export const APPLICATION_EVALUATION_PROFILES = [
  "workflow_definition_executor_v1",
  "workflow_definition_executor_v2",
  "application_rag_invocation_v1",
  "prompt_application_invocation_v1",
  "agent_copilot_suggestion_v1",
] as const;

export const APPLICATION_EVALUATION_FAILURE_CODES = [
  "application_evaluation_scope_denied",
  "application_evaluation_environment_denied",
  "application_evaluation_not_found",
  "application_evaluation_payload_invalid",
  "application_evaluation_secret_material_forbidden",
  "application_evaluation_profile_ineligible",
  "application_evaluation_version_conflict",
  "application_evaluation_archived",
  "application_evaluation_cursor_invalid",
  "application_evaluation_campaign_conflict",
  "application_evaluation_authority_changed",
  "application_evaluation_run_unavailable",
  "application_evaluation_quota_consumer_invalid",
  "application_evaluation_handoff_partial",
  "application_evaluation_store_unavailable",
  "application_evaluation_store_contract_mismatch",
  "application_evaluation_write_disabled",
] as const;

export type ApplicationEvaluationEnvironment = "development" | "test";
export type ApplicationEvaluationProfile = typeof APPLICATION_EVALUATION_PROFILES[number];
export type ApplicationEvaluationFailureCode = typeof APPLICATION_EVALUATION_FAILURE_CODES[number];
export type ApplicationEvaluationClassification =
  | "regression" | "improvement" | "changed" | "unchanged" | "inconclusive";
export type ApplicationEvaluationCampaignState = "pending" | "running" | "succeeded" | "failed" | "interrupted";
export type ApplicationEvaluationCampaignItemState = "pending" | "running" | "succeeded" | "failed";
export type JsonValue = string | number | boolean | null | JsonValue[] | { [key: string]: JsonValue };

export type ApplicationEvaluationConfig = {
  mode: "offline" | "dev_application_evaluation_http";
  authMode?: "dev_headers" | "local_session_dev_test";
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  environment: ApplicationEvaluationEnvironment;
  applicationId: string;
  subjectRef: string;
};

export type ApplicationEvaluationDefinitionTarget = {
  definitionId: string;
  expectedPointerVersion: number;
  expectedDefinitionVersion: number;
  expectedDefinitionDigest: string;
  inputContract: StructuredRuntimeInputContract | null;
};

export type ApplicationEvaluationPlanTarget = {
  workflowDefinition: ApplicationEvaluationDefinitionTarget | null;
};

export type ApplicationEvaluationPlanItem = {
  itemKey: string;
  name: string;
  expectedClassification: ApplicationEvaluationClassification;
  workflowDefinition: {
    inputText: string;
    conditionValues: Record<string, boolean>;
    model: string;
    temperature: number | null;
    inputs: StructuredRuntimeInputValues | null;
  } | null;
  applicationRAG: { input: string } | null;
  promptApplication: { variables: Record<string, JsonValue> } | null;
  agentCopilot: {
    task: string;
    locale: string;
    conversationId: string;
    artifacts: Array<Record<string, JsonValue>>;
    context: Record<string, JsonValue>;
  } | null;
};

export type ApplicationEvaluationPlan = {
  schemaVersion: typeof PLAN_SCHEMA | typeof STRUCTURED_PLAN_SCHEMA;
  planId: string;
  recordVersion: number;
  latestPlanVersion: number;
  latestPlanDigest: string;
  tenantRef: string;
  workspaceId: string;
  environment: ApplicationEvaluationEnvironment;
  applicationId: string;
  name: string;
  executionProfile: ApplicationEvaluationProfile;
  itemCount: number;
  lifecycleState: "active" | "archived";
  createdAt: string;
  updatedAt: string;
  createdByActorRef: string;
  updatedByActorRef: string;
  requestId: string;
  auditRef: string;
};

export type ApplicationEvaluationPlanVersion = {
  schemaVersion: typeof PLAN_VERSION_SCHEMA | typeof STRUCTURED_PLAN_VERSION_SCHEMA;
  planId: string;
  planVersion: number;
  previousPlanVersion: number;
  planDigest: string;
  tenantRef: string;
  workspaceId: string;
  environment: ApplicationEvaluationEnvironment;
  applicationId: string;
  name: string;
  executionProfile: ApplicationEvaluationProfile;
  target: ApplicationEvaluationPlanTarget;
  items: ApplicationEvaluationPlanItem[];
  createdAt: string;
  createdByActorRef: string;
  requestId: string;
  auditRef: string;
};

export type ApplicationEvaluationHandoff = {
  baselineCampaignId: string;
  candidateCampaignId: string;
  caseRefs: Array<{ caseId: string; version: number }>;
  suiteId: string;
  state: "partial" | "complete";
  auditRef: string;
};

export type ApplicationEvaluationCampaignItem = {
  itemKey: string;
  runId: string;
  state: ApplicationEvaluationCampaignItemState;
  runSchemaVersion: string;
  runProfile: string;
  authorityDigest: string;
  failureCode: string;
  failureBoundary: string;
  startedAt: string;
  completedAt: string;
};

export type ApplicationEvaluationCampaign = {
  schemaVersion: typeof CAMPAIGN_SCHEMA | typeof STRUCTURED_CAMPAIGN_SCHEMA;
  campaignId: string;
  clientCampaignKey: string;
  recordVersion: number;
  tenantRef: string;
  workspaceId: string;
  environment: ApplicationEvaluationEnvironment;
  applicationId: string;
  planId: string;
  planVersion: number;
  planDigest: string;
  executionProfile: ApplicationEvaluationProfile;
  quotaAPIKeyId: string;
  authority: { executionProfile: ApplicationEvaluationProfile; authorityDigest: string } | null;
  state: ApplicationEvaluationCampaignState;
  currentItemIndex: number;
  succeededItems: number;
  failedItems: number;
  failureCode: string;
  failureSummary: string;
  items: ApplicationEvaluationCampaignItem[];
  handoff: ApplicationEvaluationHandoff | null;
  createdAt: string;
  startedAt: string;
  completedAt: string;
  createdByActorRef: string;
  updatedByActorRef: string;
  requestId: string;
  auditRef: string;
};

export type ApplicationEvaluationPairItem = {
  itemKey: string;
  name: string;
  baselineRunId: string;
  candidateRunId: string;
  expectedClassification: ApplicationEvaluationClassification;
  actualClassification: ApplicationEvaluationClassification;
  expectationMatched: boolean;
  comparison: WorkflowRunComparison | null;
};

export type ApplicationEvaluationPairReview = {
  planId: string;
  planName: string;
  planVersion: number;
  planDigest: string;
  executionProfile: ApplicationEvaluationProfile;
  baselineCampaignId: string;
  candidateCampaignId: string;
  expectedMatches: number;
  expectedMismatches: number;
  items: ApplicationEvaluationPairItem[];
  existingHandoff: ApplicationEvaluationHandoff | null;
};

export type ApplicationEvaluationPlanEnvelope = ScopeEnvelope & {
  plan: ApplicationEvaluationPlan | null;
  version: ApplicationEvaluationPlanVersion | null;
  currentRecordVersion: number;
  currentState: string;
};

export type ApplicationEvaluationPlanListEnvelope = ScopeEnvelope & {
  plans: ApplicationEvaluationPlan[];
  nextCursor: string;
  hasMore: boolean;
};

export type ApplicationEvaluationCampaignEnvelope = ScopeEnvelope & {
  campaign: ApplicationEvaluationCampaign | null;
  idempotentReplay: boolean;
  currentRecordVersion: number;
  currentState: string;
};

export type ApplicationEvaluationCampaignListEnvelope = ScopeEnvelope & {
  campaigns: ApplicationEvaluationCampaign[];
  nextCursor: string;
  hasMore: boolean;
};

export type ApplicationEvaluationPairEnvelope = ScopeEnvelope & {
  review: ApplicationEvaluationPairReview | null;
  candidateCampaign: ApplicationEvaluationCampaign | null;
  handoff: ApplicationEvaluationHandoff | null;
  idempotentReplay: boolean;
  currentBaselineRecordVersion: number;
  currentCandidateRecordVersion: number;
};

export type ApplicationEvaluationPlanDraft = {
  name: string;
  executionProfile: ApplicationEvaluationProfile;
  target: ApplicationEvaluationPlanTarget;
  items: ApplicationEvaluationPlanItem[];
};

type ScopeEnvelope = {
  requestId: string;
  tenantRef: string;
  workspaceId: string;
  environment: ApplicationEvaluationEnvironment;
  applicationId: string;
  failureCode: ApplicationEvaluationFailureCode | null;
  failureSummary: string;
  auditRef: string;
};

type Document = Record<string, unknown>;

export function readApplicationEvaluationConfig(context: {
  workspaceId: string;
  applicationId: string;
}): ApplicationEvaluationConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  const configuredEnvironment = env.VITE_RADISHMIND_APPLICATION_EVALUATION_ENVIRONMENT?.trim();
  return {
    mode: env.VITE_RADISHMIND_APPLICATION_EVALUATION_SOURCE?.trim() === DEV_SOURCE
      ? "dev_application_evaluation_http"
      : "offline",
    baseUrl: normalizeBaseUrl(
      env.VITE_RADISHMIND_APPLICATION_EVALUATION_BASE_URL ??
      env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ??
      DEFAULT_BASE_URL,
    ),
    tenantRef: env.VITE_RADISHMIND_DEV_READ_TENANT_REF?.trim() || "tenant_demo",
    workspaceId: context.workspaceId.trim(),
    environment: configuredEnvironment === "development" ? "development" : "test",
    applicationId: context.applicationId.trim(),
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || "subject_demo_user",
    authMode: env.VITE_RADISHMIND_READ_AUTH_MODE?.trim() === "local_session_dev_test"
      ? "local_session_dev_test"
      : "dev_headers",
  };
}

export function applicationEvaluationPlanTemplate(
  profile: ApplicationEvaluationProfile,
): ApplicationEvaluationPlanDraft {
  const expectedClassification: ApplicationEvaluationClassification = "unchanged";
  const base = { itemKey: "case_01", name: "Representative case", expectedClassification };
  if (profile === "workflow_definition_executor_v1") {
    return {
      name: "Workflow definition regression",
      executionProfile: profile,
      target: { workflowDefinition: {
        definitionId: "wf_example",
        expectedPointerVersion: 1,
        expectedDefinitionVersion: 1,
        expectedDefinitionDigest: `sha256:${"0".repeat(64)}`,
        inputContract: null,
      } },
      items: [{ ...base, workflowDefinition: {
        inputText: "Review this representative input.", conditionValues: {}, model: "", temperature: null,
        inputs: null,
      }, applicationRAG: null, promptApplication: null, agentCopilot: null }],
    };
  }
  if (profile === "workflow_definition_executor_v2") {
    return {
      name: "Structured workflow definition regression",
      executionProfile: profile,
      target: { workflowDefinition: {
        definitionId: "wf_example",
        expectedPointerVersion: 1,
        expectedDefinitionVersion: 1,
        expectedDefinitionDigest: `sha256:${"0".repeat(64)}`,
        inputContract: {
          contractId: "contract_replace_with_exact_definition",
          fields: [{
            name: "representative_input", valueType: "string", required: true,
            label: "Representative input", description: "Replace this placeholder with the exact immutable Definition contract.",
          }],
          summary: "Replace with the exact immutable Definition input contract.",
          contractDigest: `sha256:${"0".repeat(64)}`,
        },
      } },
      items: [{ ...base, workflowDefinition: {
        inputText: "", conditionValues: {}, model: "", temperature: null,
        inputs: { representative_input: "Review this representative input." },
      }, applicationRAG: null, promptApplication: null, agentCopilot: null }],
    };
  }
  const target = { workflowDefinition: null };
  if (profile === "application_rag_invocation_v1") {
    return { name: "Application RAG regression", executionProfile: profile, target, items: [{
      ...base, workflowDefinition: null, applicationRAG: { input: "What does the reviewed knowledge snapshot say?" },
      promptApplication: null, agentCopilot: null,
    }] };
  }
  if (profile === "prompt_application_invocation_v1") {
    return { name: "Prompt application regression", executionProfile: profile, target, items: [{
      ...base, workflowDefinition: null, applicationRAG: null,
      promptApplication: { variables: { topic: "representative evaluation" } }, agentCopilot: null,
    }] };
  }
  return { name: "Agent suggestion regression", executionProfile: profile, target, items: [{
    ...base, workflowDefinition: null, applicationRAG: null, promptApplication: null,
    agentCopilot: { task: "Review the representative task.", locale: "en", conversationId: "", artifacts: [], context: {} },
  }] };
}

export function serializeApplicationEvaluationTarget(target: ApplicationEvaluationPlanTarget): string {
  return JSON.stringify(encodeTarget(target), null, 2);
}

export function serializeApplicationEvaluationItems(items: ApplicationEvaluationPlanItem[]): string {
  return JSON.stringify(items.map(encodePlanItem), null, 2);
}

export function parseApplicationEvaluationPlanDraft(
  name: string,
  executionProfile: ApplicationEvaluationProfile,
  targetJSON: string,
  itemsJSON: string,
): ApplicationEvaluationPlanDraft {
  let targetValue: unknown;
  let itemsValue: unknown;
  try {
    targetValue = JSON.parse(targetJSON);
    itemsValue = JSON.parse(itemsJSON);
  } catch {
    throw new Error("Plan target and items must be valid JSON.");
  }
  assertNoForbiddenFields(targetValue);
  assertNoForbiddenFields(itemsValue);
  if (!isTargetDocument(targetValue) || !Array.isArray(itemsValue) || itemsValue.length < 1 || itemsValue.length > 20 ||
    !itemsValue.every(isPlanItemDocument)) {
    throw new Error("Plan target or item contract failed strict validation.");
  }
  const draft = {
    name: name.trim(),
    executionProfile,
    target: mapTarget(targetValue),
    items: itemsValue.map(mapPlanItem),
  };
  if (!validPlanDraft(draft)) throw new Error("Plan profile, target, or fixture shape is inconsistent.");
  return draft;
}

export async function listApplicationEvaluationPlans(
  config: ApplicationEvaluationConfig,
): Promise<ApplicationEvaluationPlanListEnvelope> {
  if (config.mode === "offline") return offlinePlanList(config);
  const query = new URLSearchParams({
    workspace_id: config.workspaceId,
    environment: config.environment,
    lifecycle_state: "active",
    limit: "100",
  });
  return request(config, `evaluation-plans?${query}`, "GET", null, ["application_evaluations:read"], decodePlanListEnvelope);
}

export async function readApplicationEvaluationPlanVersion(
  config: ApplicationEvaluationConfig,
  planId: string,
  version: number,
): Promise<ApplicationEvaluationPlanEnvelope> {
  assertReference(planId, "plan");
  if (!positiveInteger(version)) throw new Error("Plan version is invalid.");
  const query = scopeQuery(config);
  return request(
    config,
    `evaluation-plans/${encodeURIComponent(planId)}/versions/${version}?${query}`,
    "GET",
    null,
    ["application_evaluations:read"],
    decodePlanEnvelope,
  );
}

export async function createApplicationEvaluationPlan(
  config: ApplicationEvaluationConfig,
  draft: ApplicationEvaluationPlanDraft,
): Promise<ApplicationEvaluationPlanEnvelope> {
  if (!validPlanDraft(draft)) throw new Error("Application evaluation plan draft is invalid.");
  return request(config, "evaluation-plans", "POST", {
    workspace_id: config.workspaceId,
    environment: config.environment,
    name: draft.name,
    execution_profile: draft.executionProfile,
    target: encodeTarget(draft.target),
    items: draft.items.map(encodePlanItem),
  }, ["application_evaluations:write"], decodePlanEnvelope);
}

export async function reviseApplicationEvaluationPlan(
  config: ApplicationEvaluationConfig,
  planId: string,
  expectedVersion: number,
  draft: ApplicationEvaluationPlanDraft,
): Promise<ApplicationEvaluationPlanEnvelope> {
  assertReference(planId, "plan");
  if (!positiveInteger(expectedVersion) || !validPlanDraft(draft)) {
    throw new Error("Application evaluation plan revision is invalid.");
  }
  return request(config, `evaluation-plans/${encodeURIComponent(planId)}/revisions`, "POST", {
    workspace_id: config.workspaceId,
    environment: config.environment,
    expected_version: expectedVersion,
    name: draft.name,
    execution_profile: draft.executionProfile,
    target: encodeTarget(draft.target),
    items: draft.items.map(encodePlanItem),
  }, ["application_evaluations:write"], decodePlanEnvelope);
}

export async function archiveApplicationEvaluationPlan(
  config: ApplicationEvaluationConfig,
  planId: string,
  expectedVersion: number,
): Promise<ApplicationEvaluationPlanEnvelope> {
  assertReference(planId, "plan");
  if (!positiveInteger(expectedVersion)) throw new Error("Plan archive version is invalid.");
  return request(config, `evaluation-plans/${encodeURIComponent(planId)}/archive`, "POST", {
    workspace_id: config.workspaceId,
    environment: config.environment,
    expected_version: expectedVersion,
    acknowledge_no_new_campaigns: true,
  }, ["application_evaluations:write"], decodePlanEnvelope);
}

export async function listApplicationEvaluationCampaigns(
  config: ApplicationEvaluationConfig,
  planId = "",
): Promise<ApplicationEvaluationCampaignListEnvelope> {
  if (config.mode === "offline") return offlineCampaignList(config);
  const query = new URLSearchParams({ workspace_id: config.workspaceId, environment: config.environment, limit: "100" });
  if (planId) query.set("plan_id", planId);
  return request(config, `evaluation-campaigns?${query}`, "GET", null, ["application_evaluations:read"], decodeCampaignListEnvelope);
}

export async function executeApplicationEvaluationCampaign(
  config: ApplicationEvaluationConfig,
  input: {
    plan: ApplicationEvaluationPlan;
    version: ApplicationEvaluationPlanVersion;
    clientCampaignKey: string;
    quotaAPIKeyId: string;
  },
): Promise<ApplicationEvaluationCampaignEnvelope> {
  if (input.plan.planId !== input.version.planId || input.plan.latestPlanVersion !== input.version.planVersion ||
    input.plan.latestPlanDigest !== input.version.planDigest || !IDENTIFIER.test(input.clientCampaignKey.trim()) ||
    !IDENTIFIER.test(input.quotaAPIKeyId.trim())) {
    throw new Error("Campaign authority or quota consumer is invalid.");
  }
  return request(config, "evaluation-campaigns", "POST", {
    workspace_id: config.workspaceId,
    environment: config.environment,
    plan_id: input.plan.planId,
    plan_version: input.version.planVersion,
    plan_digest: input.version.planDigest,
    expected_plan_record_version: input.plan.recordVersion,
    client_campaign_key: input.clientCampaignKey.trim(),
    quota_api_key_id: input.quotaAPIKeyId.trim(),
    acknowledge_sequential_execution: true,
    acknowledge_quota_consumption: true,
  }, ["application_evaluations:execute", "workflow_runs:execute", "workflow_definitions:read"], decodeCampaignEnvelope);
}

export async function reconcileApplicationEvaluationCampaign(
  config: ApplicationEvaluationConfig,
  campaign: ApplicationEvaluationCampaign,
): Promise<ApplicationEvaluationCampaignEnvelope> {
  if (campaign.state !== "interrupted" && campaign.state !== "running") {
    throw new Error("Only interrupted or running campaigns can be reconciled.");
  }
  return request(config, `evaluation-campaigns/${encodeURIComponent(campaign.campaignId)}/reconcile`, "POST", {
    workspace_id: config.workspaceId,
    environment: config.environment,
    expected_version: campaign.recordVersion,
  }, ["application_evaluations:execute", "workflow_runs:execute"], decodeCampaignEnvelope);
}

export async function previewApplicationEvaluationPair(
  config: ApplicationEvaluationConfig,
  baselineCampaignId: string,
  candidateCampaignId: string,
): Promise<ApplicationEvaluationPairEnvelope> {
  assertPair(baselineCampaignId, candidateCampaignId);
  return request(config, "evaluation-campaign-pairs/preview", "POST", {
    workspace_id: config.workspaceId,
    environment: config.environment,
    baseline_campaign_id: baselineCampaignId,
    candidate_campaign_id: candidateCampaignId,
  }, ["application_evaluations:read", "workflow_runs:read"], decodePairEnvelope);
}

export async function materializeApplicationEvaluationHandoff(
  config: ApplicationEvaluationConfig,
  baseline: ApplicationEvaluationCampaign,
  candidate: ApplicationEvaluationCampaign,
): Promise<ApplicationEvaluationPairEnvelope> {
  assertPair(baseline.campaignId, candidate.campaignId);
  return request(config, "evaluation-campaign-pairs/handoff", "POST", {
    workspace_id: config.workspaceId,
    environment: config.environment,
    baseline_campaign_id: baseline.campaignId,
    candidate_campaign_id: candidate.campaignId,
    expected_baseline_campaign_record_version: baseline.recordVersion,
    expected_candidate_campaign_record_version: candidate.recordVersion,
    acknowledge_evidence_materializing: true,
  }, ["application_evaluations:read", "workflow_runs:read", "workflow_evaluations:write"], decodePairEnvelope);
}

async function request<T>(
  config: ApplicationEvaluationConfig,
  suffix: string,
  method: "GET" | "POST",
  body: Record<string, unknown> | null,
  permissions: string[],
  decoder: (value: unknown, config: ApplicationEvaluationConfig) => T,
): Promise<T> {
  assertConfig(config);
  if (config.mode === "offline") throw new Error("Application evaluation HTTP is disabled.");
  const requestId = `application-evaluation-${Date.now().toString(36)}`;
  const response = await fetch(
    `${config.baseUrl}/v1/user-workspace/applications/${encodeURIComponent(config.applicationId)}/${suffix}`,
    {
      method,
      credentials: config.authMode === "local_session_dev_test" ? "include" : "omit",
      cache: "no-store",
      headers: evaluationHeaders(config, requestId, permissions, body !== null),
      body: body === null ? undefined : JSON.stringify(body),
    },
  );
  let value: unknown;
  try {
    value = await response.json();
  } catch {
    throw new Error(`Application evaluation returned invalid HTTP ${response.status} JSON.`);
  }
  assertNoForbiddenFields(value);
  try {
    return decoder(value, config);
  } catch (error) {
    const detail = error instanceof Error ? error.message : "strict validation failed";
    throw new Error(`Application evaluation returned an invalid HTTP ${response.status} envelope: ${detail}`);
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
    "X-RadishMind-Dev-Read-Identity": "radishmind-web-application-evaluation",
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

function decodePlanListEnvelope(value: unknown, config: ApplicationEvaluationConfig): ApplicationEvaluationPlanListEnvelope {
  const keys = ["request_id", "tenant_ref", "workspace_id", "environment", "application_id", "plans", "next_cursor", "has_more", "failure_code", "failure_summary", "audit_ref"];
  if (!isExactDocument(value, keys) || !Array.isArray(value.plans) || !value.plans.every((plan) => isPlanDocument(plan, config)) ||
    typeof value.next_cursor !== "string" || typeof value.has_more !== "boolean") throw new Error("plan list contract mismatch");
  return { ...mapScope(value, config), plans: value.plans.map(mapPlan), nextCursor: value.next_cursor, hasMore: value.has_more };
}

function decodePlanEnvelope(value: unknown, config: ApplicationEvaluationConfig): ApplicationEvaluationPlanEnvelope {
  const keys = ["request_id", "tenant_ref", "workspace_id", "environment", "application_id", "plan", "version", "failure_code", "failure_summary", "current_record_version", "current_state", "audit_ref"];
  if (!isExactDocument(value, keys) || !(value.plan === null || isPlanDocument(value.plan, config)) ||
    !(value.version === null || isVersionDocument(value.version, config)) || !nonNegativeInteger(value.current_record_version) ||
    typeof value.current_state !== "string") throw new Error("plan envelope contract mismatch");
  const scope = mapScope(value, config);
  if (scope.failureCode === null && value.plan === null && value.version === null) throw new Error("successful plan envelope is empty");
  return {
    ...scope,
    plan: value.plan === null ? null : mapPlan(value.plan),
    version: value.version === null ? null : mapVersion(value.version),
    currentRecordVersion: value.current_record_version,
    currentState: value.current_state,
  };
}

function decodeCampaignListEnvelope(value: unknown, config: ApplicationEvaluationConfig): ApplicationEvaluationCampaignListEnvelope {
  const keys = ["request_id", "tenant_ref", "workspace_id", "environment", "application_id", "campaigns", "next_cursor", "has_more", "failure_code", "failure_summary", "audit_ref"];
  if (!isExactDocument(value, keys) || !Array.isArray(value.campaigns) ||
    !value.campaigns.every((campaign) => isCampaignDocument(campaign, config)) || typeof value.next_cursor !== "string" ||
    typeof value.has_more !== "boolean") throw new Error("campaign list contract mismatch");
  return { ...mapScope(value, config), campaigns: value.campaigns.map(mapCampaign), nextCursor: value.next_cursor, hasMore: value.has_more };
}

function decodeCampaignEnvelope(value: unknown, config: ApplicationEvaluationConfig): ApplicationEvaluationCampaignEnvelope {
  const keys = ["request_id", "tenant_ref", "workspace_id", "environment", "application_id", "campaign", "idempotent_replay", "failure_code", "failure_summary", "current_record_version", "current_state", "audit_ref"];
  if (!isExactDocument(value, keys) || !(value.campaign === null || isCampaignDocument(value.campaign, config)) ||
    typeof value.idempotent_replay !== "boolean" || !nonNegativeInteger(value.current_record_version) ||
    typeof value.current_state !== "string") throw new Error("campaign envelope contract mismatch");
  const scope = mapScope(value, config);
  if (scope.failureCode === null && value.campaign === null) throw new Error("successful campaign envelope is empty");
  return {
    ...scope,
    campaign: value.campaign === null ? null : mapCampaign(value.campaign),
    idempotentReplay: value.idempotent_replay,
    currentRecordVersion: value.current_record_version,
    currentState: value.current_state,
  };
}

function decodePairEnvelope(value: unknown, config: ApplicationEvaluationConfig): ApplicationEvaluationPairEnvelope {
  const keys = ["request_id", "tenant_ref", "workspace_id", "environment", "application_id", "review", "candidate_campaign", "handoff", "idempotent_replay", "failure_code", "failure_summary", "current_baseline_record_version", "current_candidate_record_version", "audit_ref"];
  if (!isExactDocument(value, keys) || !(value.review === null || isPairReviewDocument(value.review, config)) ||
    !(value.candidate_campaign === null || isCampaignDocument(value.candidate_campaign, config)) ||
    !(value.handoff === null || isHandoffDocument(value.handoff)) || typeof value.idempotent_replay !== "boolean" ||
    !nonNegativeInteger(value.current_baseline_record_version) || !nonNegativeInteger(value.current_candidate_record_version)) {
    throw new Error("pair envelope contract mismatch");
  }
  const scope = mapScope(value, config);
  return {
    ...scope,
    review: value.review === null ? null : mapPairReview(value.review),
    candidateCampaign: value.candidate_campaign === null ? null : mapCampaign(value.candidate_campaign),
    handoff: value.handoff === null ? null : mapHandoff(value.handoff),
    idempotentReplay: value.idempotent_replay,
    currentBaselineRecordVersion: value.current_baseline_record_version,
    currentCandidateRecordVersion: value.current_candidate_record_version,
  };
}

function mapScope(value: Document, config: ApplicationEvaluationConfig): ScopeEnvelope {
  const failureCode = nullableFailureCode(value.failure_code);
  if (failureCode === undefined || typeof value.request_id !== "string" || typeof value.failure_summary !== "string" ||
    typeof value.audit_ref !== "string" || value.environment !== config.environment || value.application_id !== config.applicationId ||
    !matchesScope(value.tenant_ref, config.tenantRef, failureCode) || !matchesScope(value.workspace_id, config.workspaceId, failureCode)) {
    throw new Error("scope or failure contract mismatch");
  }
  return {
    requestId: value.request_id,
    tenantRef: String(value.tenant_ref),
    workspaceId: String(value.workspace_id),
    environment: config.environment,
    applicationId: config.applicationId,
    failureCode,
    failureSummary: value.failure_summary,
    auditRef: value.audit_ref,
  };
}

function isPlanDocument(value: unknown, config: ApplicationEvaluationConfig): value is Document {
  if (!isExactDocument(value, PLAN_KEYS)) return false;
  return value.schema_version === planSchemaForProfile(value.execution_profile) && typeof value.plan_id === "string" && positiveInteger(value.record_version) &&
    positiveInteger(value.latest_plan_version) && digest(value.latest_plan_digest) && scopeDocumentMatches(value, config) &&
    nonEmptyString(value.name, 160) && isProfile(value.execution_profile) && nonNegativeInteger(value.item_count) &&
    Number(value.item_count) <= 20 && (value.lifecycle_state === "active" || value.lifecycle_state === "archived") &&
    validTimestamp(value.created_at) && validTimestamp(value.updated_at) && nonEmptyString(value.created_by_actor_ref) &&
    nonEmptyString(value.updated_by_actor_ref) && nonEmptyString(value.request_id) && nonEmptyString(value.audit_ref);
}

function isVersionDocument(value: unknown, config: ApplicationEvaluationConfig): value is Document {
  if (!isExactDocument(value, VERSION_KEYS)) return false;
  return value.schema_version === planVersionSchemaForProfile(value.execution_profile) && typeof value.plan_id === "string" && positiveInteger(value.plan_version) &&
    nonNegativeInteger(value.previous_plan_version) && digest(value.plan_digest) && scopeDocumentMatches(value, config) &&
    nonEmptyString(value.name, 160) && isProfile(value.execution_profile) && isTargetDocument(value.target) &&
    Array.isArray(value.items) && value.items.length > 0 && value.items.length <= 20 && value.items.every(isPlanItemDocument) &&
    validTimestamp(value.created_at) && nonEmptyString(value.created_by_actor_ref) && nonEmptyString(value.request_id) &&
    nonEmptyString(value.audit_ref) && validPlanDraft({ name: value.name, executionProfile: value.execution_profile,
      target: mapTarget(value.target), items: value.items.map(mapPlanItem) });
}

function isTargetDocument(value: unknown): value is Document {
  if (!isExactDocument(value, ["workflow_definition"])) return false;
  if (value.workflow_definition === null) return true;
  const target = value.workflow_definition;
  if (!isDocument(target)) return false;
  const keys = Object.hasOwn(target, "input_contract")
    ? ["definition_id", "expected_pointer_version", "expected_definition_version", "expected_definition_digest", "input_contract"]
    : ["definition_id", "expected_pointer_version", "expected_definition_version", "expected_definition_digest"];
  return isExactDocument(target, keys) &&
    nonEmptyString(target.definition_id) && positiveInteger(target.expected_pointer_version) &&
    positiveInteger(target.expected_definition_version) && digest(target.expected_definition_digest) &&
    (!Object.hasOwn(target, "input_contract") || parseStructuredRuntimeInputContractDocument(target.input_contract) !== null);
}

function isPlanItemDocument(value: unknown): value is Document {
  if (!isExactDocument(value, PLAN_ITEM_KEYS) || !nonEmptyString(value.item_key, 64) || !nonEmptyString(value.name, 160) ||
    !isClassification(value.expected_classification)) return false;
  const fixtures = [value.workflow_definition, value.application_rag, value.prompt_application, value.agent_copilot];
  if (fixtures.filter((fixture) => fixture !== null).length !== 1) return false;
  const workflowValid = value.workflow_definition === null ||
    (isExactDocument(value.workflow_definition, ["input_text", "condition_values", "model", "temperature"]) &&
      nonEmptyString(value.workflow_definition.input_text, 8000) && isBooleanRecord(value.workflow_definition.condition_values) &&
      typeof value.workflow_definition.model === "string" && (value.workflow_definition.temperature === null || finiteNumber(value.workflow_definition.temperature))) ||
    (isExactDocument(value.workflow_definition, ["inputs"]) && isStructuredInputRecord(value.workflow_definition.inputs));
  const ragValid = value.application_rag === null || (isExactDocument(value.application_rag, ["input"]) && nonEmptyString(value.application_rag.input, 8000));
  const promptValid = value.prompt_application === null || (isExactDocument(value.prompt_application, ["variables"]) && isJsonRecord(value.prompt_application.variables));
  const agentValid = value.agent_copilot === null || (isExactDocument(value.agent_copilot, ["task", "locale", "conversation_id", "artifacts", "context"]) &&
    nonEmptyString(value.agent_copilot.task, 8000) && nonEmptyString(value.agent_copilot.locale, 64) &&
    typeof value.agent_copilot.conversation_id === "string" && Array.isArray(value.agent_copilot.artifacts) &&
    value.agent_copilot.artifacts.every(isArtifactDocument) && isJsonRecord(value.agent_copilot.context));
  return workflowValid && ragValid && promptValid && agentValid;
}

function isArtifactDocument(value: unknown): value is Document {
  if (!isDocument(value)) return false;
  const required = ["kind", "role", "name", "mime_type"];
  const allowed = new Set([...required, "uri", "content", "metadata"]);
  return Object.keys(value).every((key) => allowed.has(key)) && required.every((key) => nonEmptyString(value[key], 255)) &&
    (value.uri === undefined || typeof value.uri === "string") && (value.content === undefined || isJsonValue(value.content)) &&
    (value.metadata === undefined || isJsonRecord(value.metadata));
}

function isCampaignDocument(value: unknown, config: ApplicationEvaluationConfig): value is Document {
  if (!isExactDocument(value, CAMPAIGN_KEYS)) return false;
  const itemsValid = Array.isArray(value.items) && value.items.length > 0 && value.items.length <= 20 &&
    value.items.every(isCampaignItemDocument);
  return value.schema_version === campaignSchemaForProfile(value.execution_profile) && nonEmptyString(value.campaign_id) && nonEmptyString(value.client_campaign_key) &&
    positiveInteger(value.record_version) && scopeDocumentMatches(value, config) && nonEmptyString(value.plan_id) &&
    positiveInteger(value.plan_version) && digest(value.plan_digest) && isProfile(value.execution_profile) &&
    nonEmptyString(value.quota_api_key_id) && (value.authority === null || isAuthorityDocument(value.authority)) &&
    isCampaignState(value.state) && nonNegativeInteger(value.current_item_index) && nonNegativeInteger(value.succeeded_items) &&
    nonNegativeInteger(value.failed_items) && typeof value.failure_code === "string" && typeof value.failure_summary === "string" &&
    itemsValid && (value.handoff === null || isHandoffDocument(value.handoff)) && validTimestamp(value.created_at) &&
    validTimestampOrEmpty(value.started_at) && validTimestampOrEmpty(value.completed_at) && nonEmptyString(value.created_by_actor_ref) &&
    nonEmptyString(value.updated_by_actor_ref) && nonEmptyString(value.request_id) && nonEmptyString(value.audit_ref);
}

function isAuthorityDocument(value: unknown): value is Document {
  return isExactDocument(value, ["execution_profile", "authority_digest", "snapshot"]) && isProfile(value.execution_profile) &&
    digest(value.authority_digest) && isJsonValue(value.snapshot);
}

function isCampaignItemDocument(value: unknown): value is Document {
  if (!isExactDocument(value, CAMPAIGN_ITEM_KEYS)) return false;
  return nonEmptyString(value.item_key, 64) && typeof value.run_id === "string" && isCampaignItemState(value.state) &&
    typeof value.run_schema_version === "string" && typeof value.run_profile === "string" &&
    (value.authority_digest === "" || digest(value.authority_digest)) && typeof value.failure_code === "string" &&
    typeof value.failure_boundary === "string" && validTimestampOrEmpty(value.started_at) && validTimestampOrEmpty(value.completed_at);
}

function isHandoffDocument(value: unknown): value is Document {
  return isExactDocument(value, HANDOFF_KEYS) && nonEmptyString(value.baseline_campaign_id) &&
    nonEmptyString(value.candidate_campaign_id) && Array.isArray(value.case_refs) && value.case_refs.every((ref) =>
      isExactDocument(ref, ["case_id", "version"]) && nonEmptyString(ref.case_id) && positiveInteger(ref.version)) &&
    typeof value.suite_id === "string" && (value.state === "partial" || value.state === "complete") &&
    nonEmptyString(value.audit_ref);
}

function isPairReviewDocument(value: unknown, config: ApplicationEvaluationConfig): value is Document {
  if (!isExactDocument(value, PAIR_REVIEW_KEYS)) return false;
  return nonEmptyString(value.plan_id) && nonEmptyString(value.plan_name) && positiveInteger(value.plan_version) &&
    digest(value.plan_digest) && isProfile(value.execution_profile) && nonEmptyString(value.baseline_campaign_id) &&
    nonEmptyString(value.candidate_campaign_id) && nonNegativeInteger(value.expected_matches) &&
    nonNegativeInteger(value.expected_mismatches) && Array.isArray(value.items) && value.items.length > 0 &&
    value.items.every(isPairItemDocument) && (value.existing_handoff === null || isHandoffDocument(value.existing_handoff)) &&
    value.items.every((item) => {
      const comparison = item.comparison;
      if (comparison === null) return true;
      try {
        decodeWorkflowRunComparison(comparison);
        return true;
      } catch {
        return false;
      }
    }) && config.applicationId.length > 0;
}

function isPairItemDocument(value: unknown): value is Document {
  return isExactDocument(value, PAIR_ITEM_KEYS) && nonEmptyString(value.item_key, 64) && nonEmptyString(value.name, 160) &&
    nonEmptyString(value.baseline_run_id) && nonEmptyString(value.candidate_run_id) &&
    isClassification(value.expected_classification) && isClassification(value.actual_classification) &&
    typeof value.expectation_matched === "boolean" && (value.comparison === null || isDocument(value.comparison));
}

function mapPlan(value: Document): ApplicationEvaluationPlan {
  return {
    schemaVersion: value.schema_version as ApplicationEvaluationPlan["schemaVersion"],
    planId: String(value.plan_id),
    recordVersion: Number(value.record_version),
    latestPlanVersion: Number(value.latest_plan_version),
    latestPlanDigest: String(value.latest_plan_digest),
    tenantRef: String(value.tenant_ref),
    workspaceId: String(value.workspace_id),
    environment: value.environment as ApplicationEvaluationEnvironment,
    applicationId: String(value.application_id),
    name: String(value.name),
    executionProfile: value.execution_profile as ApplicationEvaluationProfile,
    itemCount: Number(value.item_count),
    lifecycleState: value.lifecycle_state as "active" | "archived",
    createdAt: String(value.created_at),
    updatedAt: String(value.updated_at),
    createdByActorRef: String(value.created_by_actor_ref),
    updatedByActorRef: String(value.updated_by_actor_ref),
    requestId: String(value.request_id),
    auditRef: String(value.audit_ref),
  };
}

function mapVersion(value: Document): ApplicationEvaluationPlanVersion {
  return {
    schemaVersion: value.schema_version as ApplicationEvaluationPlanVersion["schemaVersion"],
    planId: String(value.plan_id),
    planVersion: Number(value.plan_version),
    previousPlanVersion: Number(value.previous_plan_version),
    planDigest: String(value.plan_digest),
    tenantRef: String(value.tenant_ref),
    workspaceId: String(value.workspace_id),
    environment: value.environment as ApplicationEvaluationEnvironment,
    applicationId: String(value.application_id),
    name: String(value.name),
    executionProfile: value.execution_profile as ApplicationEvaluationProfile,
    target: mapTarget(value.target as Document),
    items: (value.items as Document[]).map(mapPlanItem),
    createdAt: String(value.created_at),
    createdByActorRef: String(value.created_by_actor_ref),
    requestId: String(value.request_id),
    auditRef: String(value.audit_ref),
  };
}

function mapTarget(value: Document): ApplicationEvaluationPlanTarget {
  if (value.workflow_definition === null) return { workflowDefinition: null };
  const definition = value.workflow_definition as Document;
  return { workflowDefinition: {
    definitionId: String(definition.definition_id),
    expectedPointerVersion: Number(definition.expected_pointer_version),
    expectedDefinitionVersion: Number(definition.expected_definition_version),
    expectedDefinitionDigest: String(definition.expected_definition_digest),
    inputContract: Object.hasOwn(definition, "input_contract")
      ? parseStructuredRuntimeInputContractDocument(definition.input_contract)
      : null,
  } };
}

function mapPlanItem(value: Document): ApplicationEvaluationPlanItem {
  const workflow = value.workflow_definition as Document | null;
  const rag = value.application_rag as Document | null;
  const prompt = value.prompt_application as Document | null;
  const agent = value.agent_copilot as Document | null;
  return {
    itemKey: String(value.item_key),
    name: String(value.name),
    expectedClassification: value.expected_classification as ApplicationEvaluationClassification,
    workflowDefinition: workflow ? Object.hasOwn(workflow, "inputs") ? {
      inputText: "", conditionValues: {}, model: "", temperature: null,
      inputs: workflow.inputs as StructuredRuntimeInputValues,
    } : {
      inputText: String(workflow.input_text), conditionValues: workflow.condition_values as Record<string, boolean>,
      model: String(workflow.model), temperature: workflow.temperature === null ? null : Number(workflow.temperature),
      inputs: null,
    } : null,
    applicationRAG: rag ? { input: String(rag.input) } : null,
    promptApplication: prompt ? { variables: prompt.variables as Record<string, JsonValue> } : null,
    agentCopilot: agent ? {
      task: String(agent.task), locale: String(agent.locale), conversationId: String(agent.conversation_id),
      artifacts: agent.artifacts as Array<Record<string, JsonValue>>, context: agent.context as Record<string, JsonValue>,
    } : null,
  };
}

function mapCampaign(value: Document): ApplicationEvaluationCampaign {
  const authority = value.authority as Document | null;
  return {
    schemaVersion: value.schema_version as ApplicationEvaluationCampaign["schemaVersion"],
    campaignId: String(value.campaign_id),
    clientCampaignKey: String(value.client_campaign_key),
    recordVersion: Number(value.record_version),
    tenantRef: String(value.tenant_ref),
    workspaceId: String(value.workspace_id),
    environment: value.environment as ApplicationEvaluationEnvironment,
    applicationId: String(value.application_id),
    planId: String(value.plan_id),
    planVersion: Number(value.plan_version),
    planDigest: String(value.plan_digest),
    executionProfile: value.execution_profile as ApplicationEvaluationProfile,
    quotaAPIKeyId: String(value.quota_api_key_id),
    authority: authority ? {
      executionProfile: authority.execution_profile as ApplicationEvaluationProfile,
      authorityDigest: String(authority.authority_digest),
    } : null,
    state: value.state as ApplicationEvaluationCampaignState,
    currentItemIndex: Number(value.current_item_index),
    succeededItems: Number(value.succeeded_items),
    failedItems: Number(value.failed_items),
    failureCode: String(value.failure_code),
    failureSummary: String(value.failure_summary),
    items: (value.items as Document[]).map(mapCampaignItem),
    handoff: value.handoff === null ? null : mapHandoff(value.handoff as Document),
    createdAt: String(value.created_at),
    startedAt: String(value.started_at),
    completedAt: String(value.completed_at),
    createdByActorRef: String(value.created_by_actor_ref),
    updatedByActorRef: String(value.updated_by_actor_ref),
    requestId: String(value.request_id),
    auditRef: String(value.audit_ref),
  };
}

function mapCampaignItem(value: Document): ApplicationEvaluationCampaignItem {
  return {
    itemKey: String(value.item_key), runId: String(value.run_id), state: value.state as ApplicationEvaluationCampaignItemState,
    runSchemaVersion: String(value.run_schema_version), runProfile: String(value.run_profile),
    authorityDigest: String(value.authority_digest), failureCode: String(value.failure_code),
    failureBoundary: String(value.failure_boundary), startedAt: String(value.started_at), completedAt: String(value.completed_at),
  };
}

function mapHandoff(value: Document): ApplicationEvaluationHandoff {
  return {
    baselineCampaignId: String(value.baseline_campaign_id),
    candidateCampaignId: String(value.candidate_campaign_id),
    caseRefs: (value.case_refs as Document[]).map((ref) => ({ caseId: String(ref.case_id), version: Number(ref.version) })),
    suiteId: String(value.suite_id),
    state: value.state as "partial" | "complete",
    auditRef: String(value.audit_ref),
  };
}

function mapPairReview(value: Document): ApplicationEvaluationPairReview {
  return {
    planId: String(value.plan_id), planName: String(value.plan_name), planVersion: Number(value.plan_version),
    planDigest: String(value.plan_digest), executionProfile: value.execution_profile as ApplicationEvaluationProfile,
    baselineCampaignId: String(value.baseline_campaign_id), candidateCampaignId: String(value.candidate_campaign_id),
    expectedMatches: Number(value.expected_matches), expectedMismatches: Number(value.expected_mismatches),
    items: (value.items as Document[]).map((item) => ({
      itemKey: String(item.item_key), name: String(item.name), baselineRunId: String(item.baseline_run_id),
      candidateRunId: String(item.candidate_run_id),
      expectedClassification: item.expected_classification as ApplicationEvaluationClassification,
      actualClassification: item.actual_classification as ApplicationEvaluationClassification,
      expectationMatched: Boolean(item.expectation_matched),
      comparison: item.comparison === null ? null : decodeWorkflowRunComparison(item.comparison),
    })),
    existingHandoff: value.existing_handoff === null ? null : mapHandoff(value.existing_handoff as Document),
  };
}

function encodeTarget(value: ApplicationEvaluationPlanTarget): Document {
  if (!value.workflowDefinition) return { workflow_definition: null };
  const definition: Document = {
    definition_id: value.workflowDefinition.definitionId,
    expected_pointer_version: value.workflowDefinition.expectedPointerVersion,
    expected_definition_version: value.workflowDefinition.expectedDefinitionVersion,
    expected_definition_digest: value.workflowDefinition.expectedDefinitionDigest,
  };
  if (value.workflowDefinition.inputContract) {
    definition.input_contract = encodeStructuredInputContract(value.workflowDefinition.inputContract);
  }
  return { workflow_definition: definition };
}

function encodePlanItem(value: ApplicationEvaluationPlanItem): Document {
  return {
    item_key: value.itemKey,
    name: value.name,
    expected_classification: value.expectedClassification,
    workflow_definition: value.workflowDefinition ? value.workflowDefinition.inputs !== null ? {
      inputs: value.workflowDefinition.inputs,
    } : {
      input_text: value.workflowDefinition.inputText,
      condition_values: value.workflowDefinition.conditionValues,
      model: value.workflowDefinition.model,
      temperature: value.workflowDefinition.temperature,
    } : null,
    application_rag: value.applicationRAG ? { input: value.applicationRAG.input } : null,
    prompt_application: value.promptApplication ? { variables: value.promptApplication.variables } : null,
    agent_copilot: value.agentCopilot ? {
      task: value.agentCopilot.task,
      locale: value.agentCopilot.locale,
      conversation_id: value.agentCopilot.conversationId,
      artifacts: value.agentCopilot.artifacts,
      context: value.agentCopilot.context,
    } : null,
  };
}

function validPlanDraft(value: ApplicationEvaluationPlanDraft): boolean {
  if (!nonEmptyString(value.name, 160) || !isProfile(value.executionProfile) || value.items.length < 1 || value.items.length > 20) return false;
  const workflowTarget = value.target.workflowDefinition;
  if (value.executionProfile === "workflow_definition_executor_v1") {
    if (!workflowTarget || !DIGEST.test(workflowTarget.expectedDefinitionDigest) ||
      !positiveInteger(workflowTarget.expectedPointerVersion) || !positiveInteger(workflowTarget.expectedDefinitionVersion) || workflowTarget.inputContract !== null) return false;
  } else if (value.executionProfile === "workflow_definition_executor_v2") {
    if (!workflowTarget || !DIGEST.test(workflowTarget.expectedDefinitionDigest) ||
      !positiveInteger(workflowTarget.expectedPointerVersion) || !positiveInteger(workflowTarget.expectedDefinitionVersion) ||
      workflowTarget.inputContract === null) return false;
  } else if (workflowTarget !== null) return false;
  return value.items.every((item) => {
    const fixtures = [item.workflowDefinition, item.applicationRAG, item.promptApplication, item.agentCopilot];
    const selected = fixtures.findIndex((fixture) => fixture !== null);
    const expected = expectedFixtureIndex(value.executionProfile);
    return selected === expected && fixtures.filter((fixture) => fixture !== null).length === 1 &&
      nonEmptyString(item.itemKey, 64) && nonEmptyString(item.name, 160) && isClassification(item.expectedClassification) &&
      (value.executionProfile !== "workflow_definition_executor_v2" || isStructuredInputRecord(item.workflowDefinition?.inputs));
  });
}

function planSchemaForProfile(profile: unknown): string {
  return profile === "workflow_definition_executor_v2" ? STRUCTURED_PLAN_SCHEMA : PLAN_SCHEMA;
}

function planVersionSchemaForProfile(profile: unknown): string {
  return profile === "workflow_definition_executor_v2" ? STRUCTURED_PLAN_VERSION_SCHEMA : PLAN_VERSION_SCHEMA;
}

function campaignSchemaForProfile(profile: unknown): string {
  return profile === "workflow_definition_executor_v2" ? STRUCTURED_CAMPAIGN_SCHEMA : CAMPAIGN_SCHEMA;
}

function expectedFixtureIndex(profile: ApplicationEvaluationProfile): number {
  if (profile === "workflow_definition_executor_v1" || profile === "workflow_definition_executor_v2") return 0;
  if (profile === "application_rag_invocation_v1") return 1;
  if (profile === "prompt_application_invocation_v1") return 2;
  return 3;
}

function encodeStructuredInputContract(contract: StructuredRuntimeInputContract): Document {
  return {
    contract_id: contract.contractId,
    fields: contract.fields.map((field) => ({
      name: field.name, value_type: field.valueType, required: field.required,
      label: field.label, description: field.description,
    })),
    summary: contract.summary,
    contract_digest: contract.contractDigest,
  };
}

function isStructuredInputRecord(value: unknown): value is Record<string, string | number | boolean> {
  return isDocument(value) && Object.values(value).every((item) =>
    typeof item === "string" || typeof item === "boolean" || typeof item === "number" && Number.isFinite(item));
}

function offlinePlanList(config: ApplicationEvaluationConfig): ApplicationEvaluationPlanListEnvelope {
  return { ...offlineScope(config), plans: [], nextCursor: "", hasMore: false };
}

function offlineCampaignList(config: ApplicationEvaluationConfig): ApplicationEvaluationCampaignListEnvelope {
  return { ...offlineScope(config), campaigns: [], nextCursor: "", hasMore: false };
}

function offlineScope(config: ApplicationEvaluationConfig): ScopeEnvelope {
  return {
    requestId: "offline_application_evaluation",
    tenantRef: config.tenantRef,
    workspaceId: config.workspaceId,
    environment: config.environment,
    applicationId: config.applicationId,
    failureCode: "application_evaluation_write_disabled",
    failureSummary: "Application evaluation requires explicit development HTTP opt-in.",
    auditRef: "offline_application_evaluation",
  };
}

function assertConfig(config: ApplicationEvaluationConfig) {
  if (!IDENTIFIER.test(config.tenantRef) || !WORKSPACE_IDENTIFIER.test(config.workspaceId) ||
    !IDENTIFIER.test(config.applicationId) || !IDENTIFIER.test(config.subjectRef) ||
    (config.environment !== "development" && config.environment !== "test") ||
    (config.authMode !== undefined && config.authMode !== "dev_headers" && config.authMode !== "local_session_dev_test")) {
    throw new Error("Application evaluation scope is invalid.");
  }
}

function assertPair(baselineCampaignId: string, candidateCampaignId: string) {
  if (!IDENTIFIER.test(baselineCampaignId) || !IDENTIFIER.test(candidateCampaignId) || baselineCampaignId === candidateCampaignId) {
    throw new Error("Choose two different evaluation campaigns.");
  }
}

function assertReference(value: string, label: string) {
  if (!IDENTIFIER.test(value)) throw new Error(`Application evaluation ${label} reference is invalid.`);
}

function assertNoForbiddenFields(value: unknown, path = "response") {
  if (Array.isArray(value)) {
    value.forEach((item, index) => assertNoForbiddenFields(item, `${path}[${index}]`));
    return;
  }
  if (!isDocument(value)) return;
  for (const [key, child] of Object.entries(value)) {
    if (FORBIDDEN_FIELDS.has(key.toLowerCase())) {
      throw new Error(`forbidden field ${path}.${key}`);
    }
    assertNoForbiddenFields(child, `${path}.${key}`);
  }
}

function scopeDocumentMatches(value: Document, config: ApplicationEvaluationConfig): boolean {
  return value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId &&
    value.environment === config.environment && value.application_id === config.applicationId;
}

function scopeQuery(config: ApplicationEvaluationConfig): string {
  return new URLSearchParams({ workspace_id: config.workspaceId, environment: config.environment }).toString();
}

function matchesScope(value: unknown, expected: string, failure: ApplicationEvaluationFailureCode | null): boolean {
  return value === expected || (failure !== null && value === "");
}

function nullableFailureCode(value: unknown): ApplicationEvaluationFailureCode | null | undefined {
  if (value === null) return null;
  return typeof value === "string" && APPLICATION_EVALUATION_FAILURE_CODES.includes(value as ApplicationEvaluationFailureCode)
    ? value as ApplicationEvaluationFailureCode
    : undefined;
}

function isProfile(value: unknown): value is ApplicationEvaluationProfile {
  return typeof value === "string" && APPLICATION_EVALUATION_PROFILES.includes(value as ApplicationEvaluationProfile);
}

function isClassification(value: unknown): value is ApplicationEvaluationClassification {
  return typeof value === "string" && ["regression", "improvement", "changed", "unchanged", "inconclusive"].includes(value);
}

function isCampaignState(value: unknown): value is ApplicationEvaluationCampaignState {
  return typeof value === "string" && ["pending", "running", "succeeded", "failed", "interrupted"].includes(value);
}

function isCampaignItemState(value: unknown): value is ApplicationEvaluationCampaignItemState {
  return typeof value === "string" && ["pending", "running", "succeeded", "failed"].includes(value);
}

function isBooleanRecord(value: unknown): value is Record<string, boolean> {
  return isDocument(value) && Object.values(value).every((item) => typeof item === "boolean");
}

function isJsonRecord(value: unknown): value is Record<string, JsonValue> {
  return isDocument(value) && Object.values(value).every(isJsonValue);
}

function isJsonValue(value: unknown): value is JsonValue {
  if (value === null || typeof value === "string" || typeof value === "boolean") return true;
  if (typeof value === "number") return Number.isFinite(value);
  if (Array.isArray(value)) return value.every(isJsonValue);
  return isJsonRecord(value);
}

function isExactDocument(value: unknown, keys: readonly string[]): value is Document {
  if (!isDocument(value)) return false;
  const actual = Object.keys(value).sort();
  const expected = [...keys].sort();
  return actual.length === expected.length && actual.every((key, index) => key === expected[index]);
}

function isDocument(value: unknown): value is Document {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function digest(value: unknown): value is string {
  return typeof value === "string" && DIGEST.test(value);
}

function nonEmptyString(value: unknown, maximum = 255): value is string {
  return typeof value === "string" && value.trim().length > 0 && value.length <= maximum;
}

function positiveInteger(value: unknown): boolean {
  return Number.isSafeInteger(value) && Number(value) > 0;
}

function nonNegativeInteger(value: unknown): value is number {
  return Number.isSafeInteger(value) && Number(value) >= 0;
}

function finiteNumber(value: unknown): boolean {
  return typeof value === "number" && Number.isFinite(value);
}

function validTimestamp(value: unknown): value is string {
  return typeof value === "string" && value.length <= 64 && !Number.isNaN(Date.parse(value));
}

function validTimestampOrEmpty(value: unknown): value is string {
  return value === "" || validTimestamp(value);
}

function normalizeBaseUrl(value: string): string {
  const normalized = value.trim() || DEFAULT_BASE_URL;
  return normalized.endsWith("/") ? normalized.slice(0, -1) : normalized;
}
