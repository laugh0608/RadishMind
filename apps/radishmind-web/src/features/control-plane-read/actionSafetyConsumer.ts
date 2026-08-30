const DIGEST_PATTERN = /^sha256:[a-f0-9]{64}$/u;
const REFERENCE_PATTERN = /^[A-Za-z0-9][A-Za-z0-9_.:/-]{0,159}$/u;
const VERSION_PATTERN = /^[a-z][a-z0-9_.-]*\.v[0-9]+$/u;
const TOKEN_PATTERN = /^[a-z][a-z0-9._-]{0,119}$/u;
const TIMESTAMP_PATTERN = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d{1,9})?Z$/u;

const LEVELS = [
  "answer_only", "proposal_only", "handoff_ready", "tool_callable",
  "write_blocked", "write_allowed_by_policy",
] as const;
const BLOCKERS = [
  "action_safety_scope_denied",
  "action_safety_payload_invalid",
  "action_safety_source_unavailable",
  "action_safety_source_changed",
  "action_safety_policy_unavailable",
  "action_safety_policy_changed",
  "action_safety_level_escalation_denied",
  "action_safety_confirmation_required",
  "action_safety_confirmation_changed",
  "action_safety_tool_authority_unavailable",
  "action_safety_write_blocked",
  "action_safety_store_contract_mismatch",
] as const;
const OWNER_KINDS = [
  "agent_copilot_response",
  "agent_copilot_runtime_assignment",
  "workflow_http_tool_action_plan",
  "workflow_run",
] as const;
const SOURCE_KINDS = ["copilot_response", "agent_copilot_proposed_action", "workflow_http_tool_action"] as const;
const TARGET_KINDS = [
  "none", "human_review_owner", "workflow_http_tool", "business_truth", "shell", "code",
  "sandbox", "agent_loop", "connector_mutation", "automatic_apply",
] as const;
const WRITE_TARGETS = new Set([
  "business_truth", "shell", "code", "sandbox", "agent_loop", "connector_mutation", "automatic_apply",
]);
const FORBIDDEN_KEYS = new Set([
  "input", "answer", "prompt", "context", "artifact_content", "candidate_body", "tool_arguments",
  "public_arguments", "url", "uri", "endpoint", "header", "headers", "authorization", "cookie",
  "credential", "secret", "token", "dsn", "raw_request", "raw_response", "provider_raw_response",
  "business_payload", "actor_ref", "request_id", "audit_ref",
]);

const TOP_RECORDED_KEYS = [
  "schema_version", "status", "owner", "projection_version", "projection_digest", "decisions",
] as const;
const TOP_LEGACY_KEYS = ["schema_version", "status", "owner", "decisions"] as const;
const OWNER_KEYS = ["kind", "id", "version"] as const;
const DECISION_KEYS = [
  "decision_id", "decision_digest", "scope", "source", "project", "task", "action_kind", "risk_level",
  "target_kind", "method", "requested_level", "maximum_allowed_level", "effective_level",
  "requires_confirmation", "confirmation_state", "writes_business_truth", "side_effect_budget",
  "blockers", "policy", "created_at",
] as const;
const SCOPE_KEYS = ["tenant_ref", "workspace_id", "environment", "application_id"] as const;
const SOURCE_KEYS = ["kind", "schema_version", "id", "version", "digest"] as const;
const POLICY_KEYS = ["version", "digest"] as const;
const BUDGET_KEYS = [
  "provider_calls", "handoff_refs", "tool_network_calls", "confirmation_consumptions",
  "business_writes", "replay_writes",
] as const;
const OBSERVED_KEYS = [
  "provider_calls", "tool_calls", "confirmation_calls", "business_writes", "replay_writes",
] as const;

export type ActionSafetyLevel = typeof LEVELS[number];
export type ActionSafetyFailureCode = typeof BLOCKERS[number];
export type ActionSafetyOwnerKind = typeof OWNER_KINDS[number];

export type ActionSafetySideEffectBudget = {
  providerCalls: number;
  handoffRefs: number;
  toolNetworkCalls: number;
  confirmationConsumptions: number;
  businessWrites: number;
  replayWrites: number;
};

export type ActionSafetyObservedSideEffects = {
  retrievalCalls: number;
  providerCalls: number;
  toolCalls: number;
  confirmationCalls: number;
  businessWrites: number;
  replayWrites: number;
};

export type ActionSafetyDecision = {
  decisionId: string;
  decisionDigest: string;
  tenantRef: string;
  workspaceId: string;
  environment: "development" | "test";
  applicationId: string;
  sourceKind: typeof SOURCE_KINDS[number];
  sourceSchemaVersion: string;
  sourceId: string;
  sourceVersion: number;
  sourceDigest: string;
  project: string;
  task: string;
  actionKind: string;
  riskLevel: "low" | "medium" | "high";
  targetKind: typeof TARGET_KINDS[number];
  method: "none" | "GET";
  requestedLevel: ActionSafetyLevel;
  maximumAllowedLevel: ActionSafetyLevel;
  effectiveLevel: ActionSafetyLevel;
  requiresConfirmation: boolean;
  confirmationState: "not_required" | "pending" | "approved" | "rejected" | "changed";
  writesBusinessTruth: boolean;
  sideEffectBudget: ActionSafetySideEffectBudget;
  blockers: ActionSafetyFailureCode[];
  policyVersion: string;
  policyDigest: string;
  createdAt: string;
};

export type ActionSafetyReadProjection = {
  schemaVersion: "action_safety_read_projection.v1";
  status: "recorded" | "not_recorded_legacy";
  owner: { kind: ActionSafetyOwnerKind; id: string; version: number; digest: string };
  projectionVersion: string;
  projectionDigest: string;
  decisions: ActionSafetyDecision[];
  observedSideEffects: ActionSafetyObservedSideEffects | null;
};

export type ActionSafetyExpectedScope = {
  tenantRef: string;
  workspaceId: string;
  applicationId: string;
  ownerKinds?: readonly ActionSafetyOwnerKind[];
  ownerId?: string;
  ownerVersion?: number;
};

type Document = Record<string, unknown>;

export function parseActionSafetyReadProjection(
  value: unknown,
  expected: ActionSafetyExpectedScope,
): ActionSafetyReadProjection | null {
  if (!isRecord(value) || containsForbiddenKey(value) ||
    value.schema_version !== "action_safety_read_projection.v1" ||
    (value.status !== "recorded" && value.status !== "not_recorded_legacy") ||
    !isOwner(value.owner, expected)) return null;

  const recorded = value.status === "recorded";
  const expectedTopKeys = recorded ? TOP_RECORDED_KEYS : TOP_LEGACY_KEYS;
  if (!hasExactKeys(value, expectedTopKeys, recorded ? ["observed_side_effects"] : [])) return null;
  const decisions = Array.isArray(value.decisions)
    ? value.decisions.map((decision) => parseDecision(decision, expected))
    : [];
  if (!Array.isArray(value.decisions) || decisions.some((decision) => decision === null)) return null;
  const parsedDecisions = decisions as ActionSafetyDecision[];
  if (new Set(parsedDecisions.map((decision) => decision.decisionId)).size !== parsedDecisions.length) return null;

  const owner = value.owner as Document;
  const ownerDigest = typeof owner.digest === "string" ? owner.digest : "";
  if (ownerDigest && !DIGEST_PATTERN.test(ownerDigest)) return null;
  if (!recorded) {
    if (parsedDecisions.length !== 0) return null;
    return {
      schemaVersion: "action_safety_read_projection.v1",
      status: "not_recorded_legacy",
      owner: { kind: owner.kind as ActionSafetyOwnerKind, id: String(owner.id), version: Number(owner.version), digest: ownerDigest },
      projectionVersion: "",
      projectionDigest: "",
      decisions: [],
      observedSideEffects: null,
    };
  }
  if (typeof value.projection_version !== "string" || !VERSION_PATTERN.test(value.projection_version) ||
    typeof value.projection_digest !== "string" || !DIGEST_PATTERN.test(value.projection_digest) ||
    parsedDecisions.length === 0) return null;
  const observed = value.observed_side_effects === undefined
    ? null
    : parseObservedSideEffects(value.observed_side_effects);
  if (value.observed_side_effects !== undefined && observed === null) return null;
  return {
    schemaVersion: "action_safety_read_projection.v1",
    status: "recorded",
    owner: { kind: owner.kind as ActionSafetyOwnerKind, id: String(owner.id), version: Number(owner.version), digest: ownerDigest },
    projectionVersion: value.projection_version,
    projectionDigest: value.projection_digest,
    decisions: parsedDecisions,
    observedSideEffects: observed,
  };
}

export function primaryActionSafetyBlocker(projection: ActionSafetyReadProjection | null): string {
  return projection?.decisions.flatMap((decision) => decision.blockers)[0] ?? "";
}

function parseDecision(value: unknown, expected: ActionSafetyExpectedScope): ActionSafetyDecision | null {
  if (!isRecord(value) || !hasExactKeys(value, DECISION_KEYS) ||
    typeof value.decision_id !== "string" || !/^asd_[a-z2-7]{16}$/u.test(value.decision_id) ||
    typeof value.decision_digest !== "string" || !DIGEST_PATTERN.test(value.decision_digest) ||
    !isRecord(value.scope) || !hasExactKeys(value.scope, SCOPE_KEYS) ||
    value.scope.tenant_ref !== expected.tenantRef || value.scope.workspace_id !== expected.workspaceId ||
    value.scope.application_id !== expected.applicationId ||
    (value.scope.environment !== "development" && value.scope.environment !== "test") ||
    !isRecord(value.source) || !hasExactKeys(value.source, SOURCE_KEYS) ||
    !SOURCE_KINDS.includes(value.source.kind as typeof SOURCE_KINDS[number]) ||
    typeof value.source.schema_version !== "string" || !VERSION_PATTERN.test(value.source.schema_version) ||
    typeof value.source.id !== "string" || !REFERENCE_PATTERN.test(value.source.id) ||
    !Number.isInteger(value.source.version) || Number(value.source.version) < 1 ||
    typeof value.source.digest !== "string" || !DIGEST_PATTERN.test(value.source.digest) ||
    typeof value.project !== "string" || !TOKEN_PATTERN.test(value.project) ||
    typeof value.task !== "string" || !TOKEN_PATTERN.test(value.task) ||
    typeof value.action_kind !== "string" || !TOKEN_PATTERN.test(value.action_kind) ||
    !["low", "medium", "high"].includes(String(value.risk_level)) ||
    !TARGET_KINDS.includes(value.target_kind as typeof TARGET_KINDS[number]) ||
    (value.method !== "none" && value.method !== "GET") ||
    !LEVELS.includes(value.requested_level as ActionSafetyLevel) ||
    !LEVELS.includes(value.maximum_allowed_level as ActionSafetyLevel) ||
    !LEVELS.includes(value.effective_level as ActionSafetyLevel) ||
    typeof value.requires_confirmation !== "boolean" ||
    !["not_required", "pending", "approved", "rejected", "changed"].includes(String(value.confirmation_state)) ||
    typeof value.writes_business_truth !== "boolean" || !isRecord(value.side_effect_budget) ||
    !Array.isArray(value.blockers) || !canonicalBlockers(value.blockers) ||
    !isRecord(value.policy) || !hasExactKeys(value.policy, POLICY_KEYS) ||
    typeof value.policy.version !== "string" || !VERSION_PATTERN.test(value.policy.version) ||
    typeof value.policy.digest !== "string" || !DIGEST_PATTERN.test(value.policy.digest) ||
    typeof value.created_at !== "string" || !TIMESTAMP_PATTERN.test(value.created_at)) return null;

  const budget = parseBudget(value.side_effect_budget);
  if (!budget) return null;
  const requested = value.requested_level as ActionSafetyLevel;
  const maximum = value.maximum_allowed_level as ActionSafetyLevel;
  const effective = value.effective_level as ActionSafetyLevel;
  const target = value.target_kind as typeof TARGET_KINDS[number];
  const writeRequest = WRITE_TARGETS.has(target) || value.writes_business_truth || value.method !== "none" && value.method !== "GET";
  const expectedEffective = transition(requested, maximum, writeRequest);
  if (!expectedEffective || expectedEffective !== effective || maximum === "write_blocked" ||
    maximum === "write_allowed_by_policy" || effective === "write_allowed_by_policy" ||
    !budgetMatchesLevel(budget, effective)) return null;
  const blockers = value.blockers as ActionSafetyFailureCode[];
  const requires = value.requires_confirmation;
  const confirmation = String(value.confirmation_state) as ActionSafetyDecision["confirmationState"];
  if ((confirmation === "not_required") !== !requires ||
    blockers.includes("action_safety_write_blocked") !== writeRequest ||
    blockers.includes("action_safety_level_escalation_denied") !== (!writeRequest && effective !== requested) ||
    (requested === "tool_callable" && confirmation !== "approved") !== blockers.includes("action_safety_confirmation_required") ||
    effective === "tool_callable" && (!requires || confirmation !== "approved" || blockers.length !== 0)) return null;

  return {
    decisionId: value.decision_id,
    decisionDigest: value.decision_digest,
    tenantRef: String(value.scope.tenant_ref), workspaceId: String(value.scope.workspace_id),
    environment: value.scope.environment, applicationId: String(value.scope.application_id),
    sourceKind: value.source.kind as ActionSafetyDecision["sourceKind"],
    sourceSchemaVersion: value.source.schema_version, sourceId: value.source.id,
    sourceVersion: Number(value.source.version), sourceDigest: value.source.digest,
    project: value.project, task: value.task, actionKind: value.action_kind,
    riskLevel: value.risk_level as ActionSafetyDecision["riskLevel"], targetKind: target,
    method: value.method, requestedLevel: requested, maximumAllowedLevel: maximum, effectiveLevel: effective,
    requiresConfirmation: requires, confirmationState: confirmation,
    writesBusinessTruth: value.writes_business_truth, sideEffectBudget: budget,
    blockers, policyVersion: value.policy.version, policyDigest: value.policy.digest, createdAt: value.created_at,
  };
}

function isOwner(value: unknown, expected: ActionSafetyExpectedScope): value is Document {
  if (!isRecord(value) || !hasExactKeys(value, OWNER_KEYS, ["digest"]) ||
    !OWNER_KINDS.includes(value.kind as ActionSafetyOwnerKind) || typeof value.id !== "string" ||
    !REFERENCE_PATTERN.test(value.id) || !Number.isInteger(value.version) || Number(value.version) < 1 ||
    (value.digest !== undefined && (typeof value.digest !== "string" || !DIGEST_PATTERN.test(value.digest)))) return false;
  if (expected.ownerKinds && !expected.ownerKinds.includes(value.kind as ActionSafetyOwnerKind)) return false;
  return (!expected.ownerId || value.id === expected.ownerId) &&
    (expected.ownerVersion === undefined || value.version === expected.ownerVersion);
}

function parseBudget(value: Document): ActionSafetySideEffectBudget | null {
  if (!hasExactKeys(value, BUDGET_KEYS) || !BUDGET_KEYS.every((key) => isBoundedCount(value[key]))) return null;
  return {
    providerCalls: Number(value.provider_calls), handoffRefs: Number(value.handoff_refs),
    toolNetworkCalls: Number(value.tool_network_calls), confirmationConsumptions: Number(value.confirmation_consumptions),
    businessWrites: Number(value.business_writes), replayWrites: Number(value.replay_writes),
  };
}

function parseObservedSideEffects(value: unknown): ActionSafetyObservedSideEffects | null {
  if (!isRecord(value) || !hasExactKeys(value, OBSERVED_KEYS, ["retrieval_calls"]) ||
    !OBSERVED_KEYS.every((key) => isBoundedCount(value[key])) ||
    (value.retrieval_calls !== undefined && !isBoundedCount(value.retrieval_calls)) ||
    value.business_writes !== 0 || value.replay_writes !== 0) return null;
  return {
    retrievalCalls: Number(value.retrieval_calls ?? 0), providerCalls: Number(value.provider_calls),
    toolCalls: Number(value.tool_calls), confirmationCalls: Number(value.confirmation_calls),
    businessWrites: 0, replayWrites: 0,
  };
}

function transition(requested: ActionSafetyLevel, maximum: ActionSafetyLevel, writeRequest: boolean): ActionSafetyLevel | "" {
  if (writeRequest) return requested === "write_allowed_by_policy" ? "write_blocked" : "";
  const matrix: Record<Exclude<ActionSafetyLevel, "write_blocked" | "write_allowed_by_policy">, ActionSafetyLevel[]> = {
    answer_only: ["answer_only", "answer_only", "answer_only", "answer_only"],
    proposal_only: ["answer_only", "proposal_only", "proposal_only", "proposal_only"],
    handoff_ready: ["answer_only", "proposal_only", "handoff_ready", "handoff_ready"],
    tool_callable: ["answer_only", "proposal_only", "handoff_ready", "tool_callable"],
  };
  const maximumIndex = ["answer_only", "proposal_only", "handoff_ready", "tool_callable"].indexOf(maximum);
  if (!(requested in matrix) || maximumIndex < 0) return "";
  return matrix[requested as keyof typeof matrix][maximumIndex];
}

function budgetMatchesLevel(budget: ActionSafetySideEffectBudget, level: ActionSafetyLevel): boolean {
  const expected = level === "answer_only" || level === "proposal_only"
    ? [1, 0, 0, 0, 0, 0]
    : level === "handoff_ready"
      ? [0, 1, 0, 0, 0, 0]
      : level === "tool_callable"
        ? [0, 0, 1, 1, 0, 0]
        : [0, 0, 0, 0, 0, 0];
  return [budget.providerCalls, budget.handoffRefs, budget.toolNetworkCalls, budget.confirmationConsumptions,
    budget.businessWrites, budget.replayWrites].every((value, index) => value === expected[index]);
}

function canonicalBlockers(values: unknown[]): values is ActionSafetyFailureCode[] {
  let last = -1;
  for (const value of values) {
    const index = BLOCKERS.indexOf(value as ActionSafetyFailureCode);
    if (index <= last) return false;
    last = index;
  }
  return true;
}

function isBoundedCount(value: unknown): boolean {
  return Number.isInteger(value) && Number(value) >= 0 && Number(value) <= 1;
}

function containsForbiddenKey(value: unknown): boolean {
  if (Array.isArray(value)) return value.some(containsForbiddenKey);
  if (!isRecord(value)) return false;
  return Object.entries(value).some(([key, nested]) => FORBIDDEN_KEYS.has(key.toLowerCase()) || containsForbiddenKey(nested));
}

function hasExactKeys(value: Document, required: readonly string[], optional: readonly string[] = []): boolean {
  const requiredSet = new Set(required);
  const allowedSet = new Set([...required, ...optional]);
  const keys = Object.keys(value);
  return required.every((key) => Object.hasOwn(value, key)) && keys.every((key) => allowedSet.has(key)) &&
    keys.length >= requiredSet.size && keys.length <= allowedSet.size;
}

function isRecord(value: unknown): value is Document {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}
