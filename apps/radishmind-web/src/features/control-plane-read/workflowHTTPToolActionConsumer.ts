import type { WorkflowSavedDraftConsumerState } from "./savedWorkflowDraftConsumer.ts";
import type {
  WorkflowDraftDesignerDraft,
  WorkflowDraftDesignerNode,
} from "./workflowDraftDesigner.ts";

const DEV_WORKFLOW_HTTP_TOOL_SOURCE = "dev-workflow-http-tool-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const DEFAULT_WORKSPACE_ID = "workspace_demo";
const DEFAULT_TENANT_REF = "tenant_demo";
const DEFAULT_SUBJECT_REF = "subject_demo_user";
const WORKFLOW_HTTP_TOOL_SCHEMA_VERSION = "workflow_http_tool_action_plan.v1";
const WORKFLOW_HTTP_TOOL_DECISION_SCHEMA_VERSION = "workflow_http_tool_confirmation_decision.v1";
const WORKFLOW_HTTP_TOOL_ID = "workflow.http.reviewed-json-read.v1";
const WORKFLOW_HTTP_TOOL_VERSION = 1;
const ACTION_PLAN_REFERENCE_STORAGE_KEY = "radishmind.workflow-http-tool-action-plan.v1";
const PLAN_SCOPE_GRANTS = ["workflow_drafts:read", "workflow_tool_actions:plan"] as const;
const READ_SCOPE_GRANTS = ["workflow_tool_actions:read"] as const;
const CONFIRM_SCOPE_GRANTS = ["workflow_tool_actions:confirm"] as const;
const EXECUTE_SCOPE_GRANTS = ["workflow_tool_actions:execute", "workflow_runs:execute", "workflow_drafts:read"] as const;
const DEFAULT_BATCH_A_SCOPE_GRANTS = [...PLAN_SCOPE_GRANTS, ...READ_SCOPE_GRANTS, ...CONFIRM_SCOPE_GRANTS];
const ACTION_PLAN_ENVELOPE_KEYS = [
  "request_id", "workspace_id", "application_id", "action_plan", "confirmation_decision",
  "failure_code", "failure_summary", "audit_ref",
] as const;
const ACTION_PLAN_KEYS = [
  "schema_version", "plan_id", "record_version", "tenant_ref", "workspace_id", "application_id",
  "draft_id", "draft_version", "node_id", "tool_id", "tool_version", "definition_digest", "profile_id",
  "profile_version", "profile_digest", "method", "target_policy_key", "public_arguments",
  "output_fields", "output_schema_digest", "credential_policy", "timeout_ms", "max_response_bytes", "max_output_bytes",
  "planned_by_actor_ref", "created_at", "expires_at", "tool_plan_digest", "status",
  "last_decision_by_actor_ref", "last_decision_at", "audit_ref",
] as const;
const CONFIRMATION_DECISION_KEYS = [
  "schema_version", "confirmation_id", "plan_id", "tenant_ref", "workspace_id", "application_id",
  "draft_id", "draft_version", "node_id", "tool_id", "tool_version", "tool_plan_digest", "outcome",
  "decided_by_actor_ref", "actor_source", "decided_at", "reason_code", "expected_record_version",
  "resulting_record_version", "audit_ref",
] as const;
const PUBLIC_ARGUMENT_KEYS = ["resource_key", "locale"] as const;
const PLAN_STATUSES = [
  "pending", "deferred", "approved", "rejected", "canceled", "expired", "invalidated", "consumed",
] as const;
const HUMAN_DECISIONS = ["approve", "reject", "defer", "cancel"] as const;
const DECISION_OUTCOMES = [...HUMAN_DECISIONS, "expire", "invalidate"] as const;
const CONFLICT_FAILURE_CODES = new Set([
  "workflow_tool_confirmation_stale",
  "workflow_tool_confirmation_mismatch",
]);
const FORBIDDEN_RESPONSE_FIELDS = new Set([
  "endpoint", "url", "scheme", "host", "port", "path", "ip", "dns", "header", "headers",
  "authorization", "cookie", "credential", "secret", "raw_query", "query_string", "raw_body",
  "raw_request", "raw_response", "stack", "stack_trace", "internal_error",
]);
const PLAN_ID_PATTERN = /^wtap_[a-z0-9]{16,64}$/u;
const CONFIRMATION_ID_PATTERN = /^wtcd_[a-z0-9]{16,64}$/u;
const SCOPED_ID_PATTERN = /^[A-Za-z0-9][A-Za-z0-9_.-]{2,119}$/u;
const REFERENCE_PATTERN = /^[A-Za-z0-9][A-Za-z0-9_.:/-]{2,159}$/u;
const PROFILE_ID_PATTERN = /^workflow_http_profile_[a-z0-9_]{1,80}$/u;
const TARGET_POLICY_KEY_PATTERN = /^[a-z][a-z0-9_.-]{2,79}$/u;
const OUTPUT_FIELDS = ["resource_key", "title", "summary", "updated_at"] as const;
const REASON_CODE_PATTERN = /^workflow_tool_confirmation_[a-z_]{3,64}$/u;
const SAFE_RESOURCE_KEY_PATTERN = /^(?![A-Za-z][A-Za-z0-9+.-]*:\/\/)[A-Za-z0-9][A-Za-z0-9._:/-]{0,159}$/u;
const SAFE_LOCALE_PATTERN = /^[A-Za-z]{2,8}(?:-[A-Za-z0-9]{1,8})*$/u;
const DIGEST_PATTERN = /^sha256:[a-f0-9]{64}$/u;

export type WorkflowHTTPToolActionConsumerMode = "disabled" | "dev_workflow_http_tool_http";

export type WorkflowHTTPToolActionConsumerConfig = {
  mode: WorkflowHTTPToolActionConsumerMode;
  baseUrl: string;
  workspaceId: string;
  tenantRef: string;
  subjectRef: string;
  scopeGrants: string[];
};

export type WorkflowHTTPToolPermission = {
  operation: "plan" | "read" | "confirm" | "execute";
  requiredGrants: string[];
  available: boolean;
  phase: "batch_a" | "batch_c";
  summary: string;
};

export type WorkflowHTTPToolActionPermissions = {
  plan: WorkflowHTTPToolPermission;
  read: WorkflowHTTPToolPermission;
  confirm: WorkflowHTTPToolPermission;
  execute: WorkflowHTTPToolPermission;
};

export type WorkflowHTTPToolPublicArguments = {
  resourceKey: string;
  locale?: string;
};

export type WorkflowHTTPToolActionPlanStatus = typeof PLAN_STATUSES[number];
export type WorkflowHTTPToolHumanDecision = typeof HUMAN_DECISIONS[number];
export type WorkflowHTTPToolDecisionOutcome = typeof DECISION_OUTCOMES[number];

export type WorkflowHTTPToolActionPlan = {
  schemaVersion: typeof WORKFLOW_HTTP_TOOL_SCHEMA_VERSION;
  planId: string;
  recordVersion: number;
  tenantRef: string;
  workspaceId: string;
  applicationId: string;
  draftId: string;
  draftVersion: number;
  nodeId: string;
  toolId: typeof WORKFLOW_HTTP_TOOL_ID;
  toolVersion: typeof WORKFLOW_HTTP_TOOL_VERSION;
  definitionDigest: string;
  profileId: string;
  profileVersion: number;
  profileDigest: string;
  method: "GET";
  targetPolicyKey: string;
  publicArguments: WorkflowHTTPToolPublicArguments;
  outputFields: string[];
  outputSchemaDigest: string;
  credentialPolicy: "none";
  timeoutMs: number;
  maxResponseBytes: number;
  maxOutputBytes: number;
  plannedByActorRef: string;
  createdAt: string;
  expiresAt: string;
  toolPlanDigest: string;
  status: WorkflowHTTPToolActionPlanStatus;
  lastDecisionByActorRef: string | null;
  lastDecisionAt: string | null;
  auditRef: string;
};

export type WorkflowHTTPToolConfirmationDecision = {
  schemaVersion: typeof WORKFLOW_HTTP_TOOL_DECISION_SCHEMA_VERSION;
  confirmationId: string;
  planId: string;
  tenantRef: string;
  workspaceId: string;
  applicationId: string;
  draftId: string;
  draftVersion: number;
  nodeId: string;
  toolId: typeof WORKFLOW_HTTP_TOOL_ID;
  toolVersion: typeof WORKFLOW_HTTP_TOOL_VERSION;
  toolPlanDigest: string;
  outcome: WorkflowHTTPToolDecisionOutcome;
  decidedByActorRef: string;
  actorSource: "human" | "system";
  decidedAt: string;
  reasonCode: string;
  auditRef: string;
  expectedRecordVersion: number;
  resultingRecordVersion: number;
};

export type WorkflowHTTPToolActionConsumerStatus =
  | "disabled"
  | "idle"
  | "creating"
  | "reading"
  | "deciding"
  | "ready"
  | "conflict_refreshed"
  | "failed";

export type WorkflowHTTPToolActionConsumerState = {
  status: WorkflowHTTPToolActionConsumerStatus;
  mode: WorkflowHTTPToolActionConsumerMode;
  summary: string;
  failureCode: string;
  requestId: string;
  auditRef: string;
  actionPlan: WorkflowHTTPToolActionPlan | null;
  confirmationDecision: WorkflowHTTPToolConfirmationDecision | null;
};

export type WorkflowHTTPToolActionEligibilityReason = {
  code: string;
  summary: string;
};

export type WorkflowHTTPToolActionEligibility = {
  eligible: boolean;
  draftVersion: number;
  nodeId: string;
  toolId: string;
  reasons: WorkflowHTTPToolActionEligibilityReason[];
};

export type WorkflowHTTPToolPublicArgumentsValidation = {
  valid: boolean;
  failureCode: string;
  summary: string;
  value: WorkflowHTTPToolPublicArguments | null;
};

export type WorkflowHTTPToolActionPlanReference = {
  workspaceId: string;
  applicationId: string;
  draftId: string;
  planId: string;
};

type ActionPlanEnvelopeDocument = {
  request_id: string;
  workspace_id: string;
  application_id: string;
  action_plan: ActionPlanDocument | null;
  confirmation_decision: ConfirmationDecisionDocument | null;
  failure_code: string | null;
  failure_summary: string;
  audit_ref: string;
};

type ActionPlanDocument = {
  schema_version: string;
  plan_id: string;
  record_version: number;
  tenant_ref: string;
  workspace_id: string;
  application_id: string;
  draft_id: string;
  draft_version: number;
  node_id: string;
  tool_id: string;
  tool_version: number;
  definition_digest: string;
  profile_id: string;
  profile_version: number;
  profile_digest: string;
  method: string;
  target_policy_key: string;
  public_arguments: { resource_key: string; locale?: string };
  output_fields: string[];
  output_schema_digest: string;
  credential_policy: string;
  timeout_ms: number;
  max_response_bytes: number;
  max_output_bytes: number;
  planned_by_actor_ref: string;
  created_at: string;
  expires_at: string;
  tool_plan_digest: string;
  status: string;
  last_decision_by_actor_ref: string | null;
  last_decision_at: string | null;
  audit_ref: string;
};

type ConfirmationDecisionDocument = {
  schema_version: string;
  confirmation_id: string;
  plan_id: string;
  tenant_ref: string;
  workspace_id: string;
  application_id: string;
  draft_id: string;
  draft_version: number;
  node_id: string;
  tool_id: string;
  tool_version: number;
  tool_plan_digest: string;
  outcome: string;
  decided_by_actor_ref: string;
  actor_source: string;
  decided_at: string;
  reason_code: string;
  expected_record_version: number;
  resulting_record_version: number;
  audit_ref: string;
};

export function readWorkflowHTTPToolActionConsumerConfig(): WorkflowHTTPToolActionConsumerConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  const scopeGrants = parseScopeGrants(env.VITE_RADISHMIND_WORKFLOW_HTTP_TOOL_SCOPE_GRANTS);
  return {
    mode: env.VITE_RADISHMIND_WORKFLOW_HTTP_TOOL_SOURCE?.trim() === DEV_WORKFLOW_HTTP_TOOL_SOURCE
      ? "dev_workflow_http_tool_http"
      : "disabled",
    baseUrl: normalizeBaseUrl(
      env.VITE_RADISHMIND_WORKFLOW_HTTP_TOOL_BASE_URL ??
      env.VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_BASE_URL ??
      env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ??
      DEFAULT_BASE_URL,
    ),
    workspaceId:
      env.VITE_RADISHMIND_WORKFLOW_HTTP_TOOL_WORKSPACE_ID?.trim() ||
      env.VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_WORKSPACE_ID?.trim() ||
      DEFAULT_WORKSPACE_ID,
    tenantRef: env.VITE_RADISHMIND_DEV_READ_TENANT_REF?.trim() || DEFAULT_TENANT_REF,
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || DEFAULT_SUBJECT_REF,
    scopeGrants,
  };
}

export function workflowHTTPToolActionPermissions(
  config: WorkflowHTTPToolActionConsumerConfig,
): WorkflowHTTPToolActionPermissions {
  const enabled = config.mode === "dev_workflow_http_tool_http";
  const permission = (
    operation: "plan" | "read" | "confirm",
    requiredGrants: readonly string[],
  ): WorkflowHTTPToolPermission => {
    const available = enabled && requiredGrants.every((grant) => config.scopeGrants.includes(grant));
    return {
      operation,
      requiredGrants: [...requiredGrants],
      available,
      phase: "batch_a",
      summary: available
        ? "Available in the configured Batch A development grant set."
        : "Unavailable until every required Batch A grant is configured.",
    };
  };
  return {
    plan: permission("plan", PLAN_SCOPE_GRANTS),
    read: permission("read", READ_SCOPE_GRANTS),
    confirm: permission("confirm", CONFIRM_SCOPE_GRANTS),
    execute: {
      operation: "execute",
      requiredGrants: [...EXECUTE_SCOPE_GRANTS],
      available: enabled && EXECUTE_SCOPE_GRANTS.every((grant) => config.scopeGrants.includes(grant)),
      phase: "batch_c",
      summary: enabled && EXECUTE_SCOPE_GRANTS.every((grant) => config.scopeGrants.includes(grant))
        ? "Available through the separately gated Batch C execution route."
        : "Unavailable until every Batch C execution grant is configured.",
    },
  };
}

export function initialWorkflowHTTPToolActionConsumerState(
  config: WorkflowHTTPToolActionConsumerConfig,
): WorkflowHTTPToolActionConsumerState {
  if (config.mode === "disabled") {
    return {
      status: "disabled",
      mode: "disabled",
      summary: "Workflow HTTP Tool action planning is disabled; offline mode sends no requests.",
      failureCode: "",
      requestId: "workflow-http-tool-action-disabled",
      auditRef: "audit_workflow_http_tool_action_disabled",
      actionPlan: null,
      confirmationDecision: null,
    };
  }
  return {
    status: "idle",
    mode: config.mode,
    summary: "Save an exact clean HTTP Tool draft before creating an immutable action plan.",
    failureCode: "",
    requestId: "workflow-http-tool-action-idle",
    auditRef: "audit_workflow_http_tool_action_idle",
    actionPlan: null,
    confirmationDecision: null,
  };
}

export function rememberWorkflowHTTPToolActionPlanReference(plan: WorkflowHTTPToolActionPlan): void {
  if (typeof globalThis.sessionStorage === "undefined") return;
  const reference: WorkflowHTTPToolActionPlanReference = {
    workspaceId: plan.workspaceId,
    applicationId: plan.applicationId,
    draftId: plan.draftId,
    planId: plan.planId,
  };
  try {
    globalThis.sessionStorage.setItem(ACTION_PLAN_REFERENCE_STORAGE_KEY, JSON.stringify(reference));
  } catch {
    // The durable resource remains authoritative when browser storage is unavailable.
  }
}

export function readWorkflowHTTPToolActionPlanReference(): WorkflowHTTPToolActionPlanReference | null {
  if (typeof globalThis.sessionStorage === "undefined") return null;
  try {
    const raw = globalThis.sessionStorage.getItem(ACTION_PLAN_REFERENCE_STORAGE_KEY);
    if (!raw) return null;
    const value: unknown = JSON.parse(raw);
    if (!isRecord(value) || !hasExactKeys(value, ["workspaceId", "applicationId", "draftId", "planId"]) ||
      !isScopedId(value.workspaceId) || !isScopedId(value.applicationId) ||
      !isScopedId(value.draftId) || !isPlanId(value.planId)) return null;
    return {
      workspaceId: value.workspaceId,
      applicationId: value.applicationId,
      draftId: value.draftId,
      planId: value.planId,
    };
  } catch {
    return null;
  }
}

export function evaluateWorkflowHTTPToolActionEligibility(
  draft: WorkflowDraftDesignerDraft,
  savedDraftState: WorkflowSavedDraftConsumerState,
  draftEditDirty: boolean,
): WorkflowHTTPToolActionEligibility {
  const reasons: WorkflowHTTPToolActionEligibilityReason[] = [];
  const toolNodes = draft.nodes.filter((node) => node.nodeType === "http_tool");
  const toolNode = toolNodes.length === 1 ? toolNodes[0] : null;

  if (savedDraftState.status !== "saved_dev_record" || savedDraftState.currentDraftVersion < 1) {
    reasons.push({
      code: "workflow_tool_saved_draft_required",
      summary: "The selected draft must have an exact durable saved version.",
    });
  }
  if (draftEditDirty) {
    reasons.push({
      code: "workflow_tool_unsaved_local_changes",
      summary: "Save or discard local draft changes before creating a plan.",
    });
  }
  if (!toolNode || toolNodes.length !== 1) {
    reasons.push({
      code: "workflow_tool_node_count_invalid",
      summary: "The first version requires exactly one HTTP Tool node.",
    });
  } else {
    if (toolNode.toolRef !== WORKFLOW_HTTP_TOOL_ID) {
      reasons.push({
        code: "workflow_tool_exact_version_required",
        summary: `The HTTP Tool node must reference ${WORKFLOW_HTTP_TOOL_ID} exactly.`,
      });
    }
    if (!toolNode.requiresConfirmation || toolNode.riskLevel !== "medium") {
      reasons.push({
        code: "workflow_tool_confirmation_boundary_invalid",
        summary: "The HTTP Tool node must retain medium risk and mandatory confirmation.",
      });
    }
  }
  if (!isEligibleHTTPToolGraph(draft)) {
    reasons.push({
      code: "workflow_tool_graph_ineligible",
      summary: "The saved graph must be a single prompt → HTTP Tool → one or more LLM → output chain.",
    });
  }

  return {
    eligible: reasons.length === 0,
    draftVersion: savedDraftState.currentDraftVersion,
    nodeId: toolNode?.nodeId ?? "",
    toolId: toolNode?.toolRef ?? "",
    reasons,
  };
}

export function validateWorkflowHTTPToolPublicArguments(
  input: WorkflowHTTPToolPublicArguments,
): WorkflowHTTPToolPublicArgumentsValidation {
  const resourceKey = input.resourceKey.trim();
  const locale = input.locale?.trim() ?? "";
  if (!SAFE_RESOURCE_KEY_PATTERN.test(resourceKey)) {
    return invalidArguments("resource_key must be a stable 1–160 character public resource identifier.");
  }
  if (locale && !SAFE_LOCALE_PATTERN.test(locale)) {
    return invalidArguments("locale must be a short language tag such as zh-CN or en.");
  }
  if (containsSensitiveText(resourceKey) || containsSensitiveText(locale)) {
    return invalidArguments("Public arguments contain disallowed transport or secret material.");
  }
  return {
    valid: true,
    failureCode: "",
    summary: "Public arguments are valid for server-side definition validation.",
    value: locale ? { resourceKey, locale } : { resourceKey },
  };
}

export async function createWorkflowHTTPToolActionPlan(
  config: WorkflowHTTPToolActionConsumerConfig,
  input: {
    draftId: string;
    applicationId: string;
    draftVersion: number;
    nodeId: string;
    publicArguments: WorkflowHTTPToolPublicArguments;
  },
): Promise<WorkflowHTTPToolActionConsumerState> {
  if (config.mode === "disabled") return initialWorkflowHTTPToolActionConsumerState(config);
  if (!workflowHTTPToolActionPermissions(config).plan.available) {
    return failedState(config, "workflow_tool_action_scope_denied", "Action-plan creation requires workflow_drafts:read and workflow_tool_actions:plan.");
  }
  const validation = validateWorkflowHTTPToolPublicArguments(input.publicArguments);
  if (!isScopedId(input.draftId) || !isScopedId(input.applicationId) || !isScopedId(input.nodeId) ||
    !isPositiveInteger(input.draftVersion) || !validation.valid || !validation.value) {
    return failedState(config, validation.failureCode || "workflow_tool_arguments_invalid", validation.summary || "Action plan input is invalid.");
  }
  const requestId = createRequestId("workflow-tool-plan-create");
  const publicArguments = {
    resource_key: validation.value.resourceKey,
    ...(validation.value.locale ? { locale: validation.value.locale } : {}),
  };
  return requestActionPlan(
    config,
    input.applicationId,
    requestId,
    "plan",
    `/v1/user-workspace/workflow-drafts/${encodeURIComponent(input.draftId)}/tool-action-plans`,
    {
      method: "POST",
      body: JSON.stringify({
        workspace_id: config.workspaceId,
        application_id: input.applicationId,
        draft_version: input.draftVersion,
        node_id: input.nodeId,
        public_arguments: publicArguments,
      }),
    },
    "create",
  );
}

export async function readWorkflowHTTPToolActionPlan(
  config: WorkflowHTTPToolActionConsumerConfig,
  plan: Pick<WorkflowHTTPToolActionPlan, "planId" | "applicationId"> & Partial<Pick<WorkflowHTTPToolActionPlan, "draftId">>,
): Promise<WorkflowHTTPToolActionConsumerState> {
  if (config.mode === "disabled") return initialWorkflowHTTPToolActionConsumerState(config);
  if (!workflowHTTPToolActionPermissions(config).read.available) {
    return failedState(config, "workflow_tool_action_scope_denied", "Action-plan detail requires workflow_tool_actions:read.");
  }
  if (!isPlanId(plan.planId) || !isScopedId(plan.applicationId)) {
    return failedState(config, "workflow_tool_arguments_invalid", "Action plan scope is invalid.");
  }
  const requestId = createRequestId("workflow-tool-plan-read");
  const query = new URLSearchParams({
    workspace_id: config.workspaceId,
    application_id: plan.applicationId,
  });
  const result = await requestActionPlan(
    config,
    plan.applicationId,
    requestId,
    "read",
    `/v1/user-workspace/workflow-tool-action-plans/${encodeURIComponent(plan.planId)}?${query}`,
    { method: "GET" },
    "read",
  );
  if (plan.draftId && result.actionPlan && result.actionPlan.draftId !== plan.draftId) {
    return failedState(config, "workflow_tool_store_contract_mismatch", "The durable action plan does not belong to the selected saved draft.");
  }
  return result;
}

export async function decideWorkflowHTTPToolActionPlan(
  config: WorkflowHTTPToolActionConsumerConfig,
  plan: WorkflowHTTPToolActionPlan,
  decision: WorkflowHTTPToolHumanDecision,
): Promise<WorkflowHTTPToolActionConsumerState> {
  if (config.mode === "disabled") return initialWorkflowHTTPToolActionConsumerState(config);
  if (!workflowHTTPToolActionPermissions(config).confirm.available) {
    return failedState(config, "workflow_tool_action_scope_denied", "Human decisions require workflow_tool_actions:confirm.", plan);
  }
  if (!HUMAN_DECISIONS.includes(decision) || !decisionAllowed(plan.status, decision)) {
    return failedState(config, "workflow_tool_confirmation_stale", "The selected decision is not allowed for the current durable plan state.", plan);
  }
  const requestId = createRequestId("workflow-tool-plan-decision");
  const result = await requestActionPlan(
    config,
    plan.applicationId,
    requestId,
    "confirm",
    `/v1/user-workspace/workflow-tool-action-plans/${encodeURIComponent(plan.planId)}/decisions`,
    {
      method: "POST",
      body: JSON.stringify({
        workspace_id: config.workspaceId,
        application_id: plan.applicationId,
        expected_record_version: plan.recordVersion,
        decision,
      }),
    },
    "decision",
  );
  if (!CONFLICT_FAILURE_CODES.has(result.failureCode)) return result;

  const refreshed = await readWorkflowHTTPToolActionPlan(config, plan);
  if (!refreshed.actionPlan) {
    return {
      ...refreshed,
      failureCode: result.failureCode,
      summary: "The CAS decision conflicted and the durable plan could not be refreshed.",
    };
  }
  return {
    ...refreshed,
    status: "conflict_refreshed",
    failureCode: result.failureCode,
    confirmationDecision: null,
    summary: `The decision was not stored because the plan changed. Durable version ${refreshed.actionPlan.recordVersion} is now displayed; review it before retrying.`,
  };
}

function isEligibleHTTPToolGraph(draft: WorkflowDraftDesignerDraft): boolean {
  if (draft.nodes.length < 4 || draft.edges.length !== draft.nodes.length - 1) return false;
  const byId = new Map(draft.nodes.map((node) => [node.nodeId, node]));
  const outgoing = new Map<string, string[]>();
  const incoming = new Map<string, string[]>();
  for (const edge of draft.edges) {
    if (!byId.has(edge.fromNodeId) || !byId.has(edge.toNodeId)) return false;
    outgoing.set(edge.fromNodeId, [...(outgoing.get(edge.fromNodeId) ?? []), edge.toNodeId]);
    incoming.set(edge.toNodeId, [...(incoming.get(edge.toNodeId) ?? []), edge.fromNodeId]);
  }
  const roots = draft.nodes.filter((node) => (incoming.get(node.nodeId) ?? []).length === 0);
  const terminals = draft.nodes.filter((node) => (outgoing.get(node.nodeId) ?? []).length === 0);
  if (roots.length !== 1 || roots[0]?.nodeType !== "prompt" || terminals.length !== 1 || terminals[0]?.nodeType !== "output") return false;

  const chain: WorkflowDraftDesignerNode[] = [];
  const visited = new Set<string>();
  let current: WorkflowDraftDesignerNode | undefined = roots[0];
  while (current && !visited.has(current.nodeId)) {
    chain.push(current);
    visited.add(current.nodeId);
    const nextIds: string[] = outgoing.get(current.nodeId) ?? [];
    if (nextIds.length > 1) return false;
    current = nextIds.length === 1 ? byId.get(nextIds[0]!) : undefined;
  }
  if (chain.length !== draft.nodes.length || visited.size !== draft.nodes.length) return false;
  const types = chain.map((node) => node.nodeType);
  return types[0] === "prompt" && types[1] === "http_tool" && types.at(-1) === "output" &&
    types.slice(2, -1).length >= 1 && types.slice(2, -1).every((nodeType) => nodeType === "llm");
}

function decisionAllowed(status: WorkflowHTTPToolActionPlanStatus, decision: WorkflowHTTPToolHumanDecision): boolean {
  if (status === "pending") return true;
  if (status === "deferred") return decision !== "defer";
  return status === "approved" && decision === "cancel";
}

async function requestActionPlan(
  config: WorkflowHTTPToolActionConsumerConfig,
  applicationId: string,
  requestId: string,
  scope: "plan" | "read" | "confirm",
  path: string,
  init: RequestInit,
  operation: "create" | "read" | "decision",
): Promise<WorkflowHTTPToolActionConsumerState> {
  try {
    const response = await fetch(`${config.baseUrl}${path}`, {
      ...init,
      headers: workflowHTTPToolActionHeaders(config, applicationId, requestId, scope),
    });
    const body: unknown = await response.json();
    if (!isActionPlanEnvelope(body, config, applicationId) || (!response.ok && !body.failure_code)) {
      return failedState(config, "workflow_tool_store_contract_mismatch", "The action plan route returned an invalid or unsafe envelope.");
    }
    if (!body.failure_code && ((operation === "decision" && body.confirmation_decision === null) ||
      (operation === "create" && body.confirmation_decision !== null))) {
      return failedState(config, "workflow_tool_store_contract_mismatch", "The action plan route returned a decision shape that does not match the operation.");
    }
    return stateFromEnvelope(config, body, operation);
  } catch {
    return failedState(config, "workflow_tool_store_unavailable", "The durable action plan route is unavailable.");
  }
}

function workflowHTTPToolActionHeaders(
  config: WorkflowHTTPToolActionConsumerConfig,
  applicationId: string,
  requestId: string,
  scope: "plan" | "read" | "confirm",
): HeadersInit {
  const scopes = scope === "plan" ? PLAN_SCOPE_GRANTS : scope === "confirm" ? CONFIRM_SCOPE_GRANTS : READ_SCOPE_GRANTS;
  return {
    Accept: "application/json",
    "Content-Type": "application/json",
    "X-Request-Id": requestId,
    "X-RadishMind-Dev-Read-Identity": "dev-workflow-http-tool-action-consumer",
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Read-Scopes": scopes.join(","),
    "X-RadishMind-Dev-Read-Audit": "audit_dev_workflow_http_tool_action_consumer",
    "X-RadishMind-Dev-Workflow-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Workflow-Application": applicationId,
  };
}

function stateFromEnvelope(
  config: WorkflowHTTPToolActionConsumerConfig,
  envelope: ActionPlanEnvelopeDocument,
  operation: "create" | "read" | "decision",
): WorkflowHTTPToolActionConsumerState {
  const plan = envelope.action_plan ? mapActionPlan(envelope.action_plan) : null;
  const confirmationDecision = envelope.confirmation_decision
    ? mapConfirmationDecision(envelope.confirmation_decision)
    : null;
  if (envelope.failure_code || !plan) {
    return {
      status: "failed",
      mode: config.mode,
      summary: envelope.failure_summary || "The durable action plan operation failed without changing network or run state.",
      failureCode: envelope.failure_code ?? "workflow_tool_store_contract_mismatch",
      requestId: envelope.request_id,
      auditRef: envelope.audit_ref,
      actionPlan: plan,
      confirmationDecision: null,
    };
  }
  const summary = operation === "create"
    ? `Immutable action plan ${plan.planId} is ready for human review; no network request or workflow run was started.`
    : operation === "read"
      ? `Durable action plan ${plan.planId} version ${plan.recordVersion} was reloaded.`
      : `Decision ${confirmationDecision?.outcome ?? "unknown"} was recorded at plan version ${plan.recordVersion}; approval does not start a network request or workflow run.`;
  return {
    status: "ready",
    mode: config.mode,
    summary,
    failureCode: "",
    requestId: envelope.request_id,
    auditRef: envelope.audit_ref,
    actionPlan: plan,
    confirmationDecision,
  };
}

function isActionPlanEnvelope(
  value: unknown,
  config: WorkflowHTTPToolActionConsumerConfig,
  applicationId: string,
): value is ActionPlanEnvelopeDocument {
  if (!isRecord(value) || !hasExactKeys(value, ACTION_PLAN_ENVELOPE_KEYS) || containsForbiddenResponse(value)) return false;
  if (!isNonEmptyString(value.request_id) || value.workspace_id !== config.workspaceId || value.application_id !== applicationId ||
    !(value.failure_code === null || isNonEmptyString(value.failure_code)) || typeof value.failure_summary !== "string" ||
    !isNonEmptyString(value.audit_ref)) return false;
  if (!(value.action_plan === null || isActionPlanDocument(value.action_plan, config, applicationId))) return false;
  if (!(value.confirmation_decision === null || isConfirmationDecisionDocument(value.confirmation_decision, config, applicationId))) return false;
  if (value.failure_code === null && value.action_plan === null) return false;
  if (value.failure_code !== null && (value.action_plan !== null || value.confirmation_decision !== null)) return false;
  if (value.confirmation_decision !== null && value.action_plan === null) return false;
  return value.confirmation_decision === null ||
    (value.action_plan !== null && confirmationDecisionMatchesPlan(value.confirmation_decision, value.action_plan));
}

function confirmationDecisionMatchesPlan(
  decision: ConfirmationDecisionDocument,
  plan: ActionPlanDocument,
): boolean {
  const expectedStatus: Record<WorkflowHTTPToolDecisionOutcome, WorkflowHTTPToolActionPlanStatus> = {
    approve: "approved",
    reject: "rejected",
    defer: "deferred",
    cancel: "canceled",
    expire: "expired",
    invalidate: "invalidated",
  };
  return decision.plan_id === plan.plan_id && decision.tenant_ref === plan.tenant_ref &&
    decision.workspace_id === plan.workspace_id && decision.application_id === plan.application_id &&
    decision.draft_id === plan.draft_id && decision.draft_version === plan.draft_version &&
    decision.node_id === plan.node_id && decision.tool_id === plan.tool_id && decision.tool_version === plan.tool_version &&
    decision.tool_plan_digest === plan.tool_plan_digest && decision.resulting_record_version === plan.record_version &&
    plan.status === expectedStatus[decision.outcome as WorkflowHTTPToolDecisionOutcome] &&
    plan.last_decision_by_actor_ref === decision.decided_by_actor_ref && plan.last_decision_at === decision.decided_at;
}

function isActionPlanDocument(
  value: unknown,
  config: WorkflowHTTPToolActionConsumerConfig,
  applicationId: string,
): value is ActionPlanDocument {
  if (!isRecord(value) || !hasExactKeys(value, ACTION_PLAN_KEYS)) return false;
  if (value.schema_version !== WORKFLOW_HTTP_TOOL_SCHEMA_VERSION || !isPlanId(value.plan_id) ||
    !isPositiveInteger(value.record_version) || value.tenant_ref !== config.tenantRef ||
    value.workspace_id !== config.workspaceId || value.application_id !== applicationId ||
    !isScopedId(value.draft_id) || !isPositiveInteger(value.draft_version) || !isScopedId(value.node_id) ||
    value.tool_id !== WORKFLOW_HTTP_TOOL_ID || value.tool_version !== WORKFLOW_HTTP_TOOL_VERSION ||
    !isDigest(value.definition_digest) || !isProfileId(value.profile_id) ||
    !isPositiveInteger(value.profile_version) || !isDigest(value.profile_digest) || value.method !== "GET" ||
    !isTargetPolicyKey(value.target_policy_key) || !isPublicArgumentsDocument(value.public_arguments) ||
    !isOutputFields(value.output_fields) || !isDigest(value.output_schema_digest) ||
    value.credential_policy !== "none" || value.timeout_ms !== 5000 ||
    value.max_response_bytes !== 65536 || value.max_output_bytes !== 16384 || !isReference(value.planned_by_actor_ref) ||
    !isTimestamp(value.created_at) || !isTimestamp(value.expires_at) || !isDigest(value.tool_plan_digest) ||
    !PLAN_STATUSES.includes(value.status as WorkflowHTTPToolActionPlanStatus) || !isNullableSafeIdentifier(value.last_decision_by_actor_ref) ||
    !isNullableTimestamp(value.last_decision_at) || !isReference(value.audit_ref)) return false;
  if ((value.last_decision_by_actor_ref === null) !== (value.last_decision_at === null) ||
    (value.status === "pending") !== (value.last_decision_by_actor_ref === null)) return false;
  return new Date(value.expires_at).getTime() > new Date(value.created_at).getTime();
}

export function parseWorkflowHTTPToolActionPlanDocument(
  value: unknown,
  config: WorkflowHTTPToolActionConsumerConfig,
  applicationId: string,
): WorkflowHTTPToolActionPlan | null {
  return isActionPlanDocument(value, config, applicationId) ? mapActionPlan(value) : null;
}

function isConfirmationDecisionDocument(
  value: unknown,
  config: WorkflowHTTPToolActionConsumerConfig,
  applicationId: string,
): value is ConfirmationDecisionDocument {
  return isRecord(value) && hasExactKeys(value, CONFIRMATION_DECISION_KEYS) &&
    value.schema_version === WORKFLOW_HTTP_TOOL_DECISION_SCHEMA_VERSION && isConfirmationId(value.confirmation_id) &&
    isPlanId(value.plan_id) && value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId &&
    value.application_id === applicationId && isScopedId(value.draft_id) && isPositiveInteger(value.draft_version) &&
    isScopedId(value.node_id) && value.tool_id === WORKFLOW_HTTP_TOOL_ID &&
    value.tool_version === WORKFLOW_HTTP_TOOL_VERSION && isDigest(value.tool_plan_digest) &&
    DECISION_OUTCOMES.includes(value.outcome as WorkflowHTTPToolDecisionOutcome) && isReference(value.decided_by_actor_ref) &&
    (value.actor_source === "human" || value.actor_source === "system") && isTimestamp(value.decided_at) &&
    typeof value.reason_code === "string" && REASON_CODE_PATTERN.test(value.reason_code) && isPositiveInteger(value.expected_record_version) &&
    isPositiveInteger(value.resulting_record_version) && value.resulting_record_version === value.expected_record_version + 1 &&
    isReference(value.audit_ref) &&
    (value.actor_source === "human" ? HUMAN_DECISIONS.includes(value.outcome as WorkflowHTTPToolHumanDecision) :
      value.outcome === "expire" || value.outcome === "invalidate");
}

function isPublicArgumentsDocument(value: unknown): value is ActionPlanDocument["public_arguments"] {
  if (!isRecord(value) || !hasOnlyKeys(value, PUBLIC_ARGUMENT_KEYS) || typeof value.resource_key !== "string" ||
    !(value.locale === undefined || typeof value.locale === "string")) return false;
  return validateWorkflowHTTPToolPublicArguments({
    resourceKey: value.resource_key,
    ...(value.locale ? { locale: value.locale } : {}),
  }).valid;
}

function isOutputFields(value: unknown): value is string[] {
  return Array.isArray(value) && value.length === OUTPUT_FIELDS.length &&
    value.every((field, index) => field === OUTPUT_FIELDS[index]);
}

function mapActionPlan(document: ActionPlanDocument): WorkflowHTTPToolActionPlan {
  return {
    schemaVersion: WORKFLOW_HTTP_TOOL_SCHEMA_VERSION,
    planId: document.plan_id,
    recordVersion: document.record_version,
    tenantRef: document.tenant_ref,
    workspaceId: document.workspace_id,
    applicationId: document.application_id,
    draftId: document.draft_id,
    draftVersion: document.draft_version,
    nodeId: document.node_id,
    toolId: WORKFLOW_HTTP_TOOL_ID,
    toolVersion: WORKFLOW_HTTP_TOOL_VERSION,
    definitionDigest: document.definition_digest,
    profileId: document.profile_id,
    profileVersion: document.profile_version,
    profileDigest: document.profile_digest,
    method: "GET",
    targetPolicyKey: document.target_policy_key,
    publicArguments: {
      resourceKey: document.public_arguments.resource_key,
      ...(document.public_arguments.locale ? { locale: document.public_arguments.locale } : {}),
    },
    outputFields: [...document.output_fields],
    outputSchemaDigest: document.output_schema_digest,
    credentialPolicy: "none",
    timeoutMs: document.timeout_ms,
    maxResponseBytes: document.max_response_bytes,
    maxOutputBytes: document.max_output_bytes,
    plannedByActorRef: document.planned_by_actor_ref,
    createdAt: document.created_at,
    expiresAt: document.expires_at,
    toolPlanDigest: document.tool_plan_digest,
    status: document.status as WorkflowHTTPToolActionPlanStatus,
    lastDecisionByActorRef: document.last_decision_by_actor_ref,
    lastDecisionAt: document.last_decision_at,
    auditRef: document.audit_ref,
  };
}

function mapConfirmationDecision(document: ConfirmationDecisionDocument): WorkflowHTTPToolConfirmationDecision {
  return {
    schemaVersion: WORKFLOW_HTTP_TOOL_DECISION_SCHEMA_VERSION,
    confirmationId: document.confirmation_id,
    planId: document.plan_id,
    tenantRef: document.tenant_ref,
    workspaceId: document.workspace_id,
    applicationId: document.application_id,
    draftId: document.draft_id,
    draftVersion: document.draft_version,
    nodeId: document.node_id,
    toolId: WORKFLOW_HTTP_TOOL_ID,
    toolVersion: WORKFLOW_HTTP_TOOL_VERSION,
    toolPlanDigest: document.tool_plan_digest,
    outcome: document.outcome as WorkflowHTTPToolDecisionOutcome,
    decidedByActorRef: document.decided_by_actor_ref,
    actorSource: document.actor_source as "human" | "system",
    decidedAt: document.decided_at,
    reasonCode: document.reason_code,
    auditRef: document.audit_ref,
    expectedRecordVersion: document.expected_record_version,
    resultingRecordVersion: document.resulting_record_version,
  };
}

function failedState(
  config: WorkflowHTTPToolActionConsumerConfig,
  failureCode: string,
  summary: string,
  actionPlan: WorkflowHTTPToolActionPlan | null = null,
): WorkflowHTTPToolActionConsumerState {
  return {
    status: "failed",
    mode: config.mode,
    summary,
    failureCode,
    requestId: "",
    auditRef: actionPlan?.auditRef ?? "",
    actionPlan,
    confirmationDecision: null,
  };
}

function invalidArguments(summary: string): WorkflowHTTPToolPublicArgumentsValidation {
  return { valid: false, failureCode: "workflow_tool_arguments_invalid", summary, value: null };
}

function containsForbiddenResponse(value: unknown): boolean {
  if (Array.isArray(value)) return value.some(containsForbiddenResponse);
  if (typeof value === "string") return containsSensitiveText(value);
  if (!isRecord(value)) return false;
  return Object.entries(value).some(([key, nested]) =>
    FORBIDDEN_RESPONSE_FIELDS.has(key.toLowerCase()) || containsForbiddenResponse(nested));
}

function containsSensitiveText(value: string): boolean {
  return /https?:\/\/|authorization\s*[:=]|bearer\s+|api[_-]?key\s*[:=]|cookie\s*[:=]|raw[_-]?query|[?&][A-Za-z0-9._-]+=|x-radishmind-/iu.test(value);
}

function hasExactKeys(value: Record<string, unknown>, allowed: readonly string[]): boolean {
  return Object.keys(value).length === allowed.length && hasOnlyKeys(value, allowed);
}

function hasOnlyKeys(value: Record<string, unknown>, allowed: readonly string[]): boolean {
  const allowedSet = new Set(allowed);
  return Object.keys(value).every((key) => allowedSet.has(key));
}

function isPlanId(value: unknown): value is string {
  return typeof value === "string" && PLAN_ID_PATTERN.test(value);
}

function isConfirmationId(value: unknown): value is string {
  return typeof value === "string" && CONFIRMATION_ID_PATTERN.test(value);
}

function isScopedId(value: unknown): value is string {
  return typeof value === "string" && SCOPED_ID_PATTERN.test(value);
}

function isReference(value: unknown): value is string {
  return typeof value === "string" && REFERENCE_PATTERN.test(value);
}

function isProfileId(value: unknown): value is string {
  return typeof value === "string" && PROFILE_ID_PATTERN.test(value);
}

function isTargetPolicyKey(value: unknown): value is string {
  return typeof value === "string" && TARGET_POLICY_KEY_PATTERN.test(value);
}

function isNonEmptyString(value: unknown): value is string {
  return typeof value === "string" && value.trim().length > 0;
}

function isNullableSafeIdentifier(value: unknown): value is string | null {
  return value === null || isReference(value);
}

function isDigest(value: unknown): value is string {
  return typeof value === "string" && DIGEST_PATTERN.test(value);
}

function isTimestamp(value: unknown): value is string {
  return typeof value === "string" && value.length >= 20 && Number.isFinite(Date.parse(value));
}

function isNullableTimestamp(value: unknown): value is string | null {
  return value === null || isTimestamp(value);
}

function isPositiveInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value > 0;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function parseScopeGrants(value: string | undefined): string[] {
  if (value === undefined) return [...DEFAULT_BATCH_A_SCOPE_GRANTS];
  return [...new Set(value.split(/[\s,]+/u).map((grant) => grant.trim()).filter(Boolean))];
}

function normalizeBaseUrl(value: string): string {
  return value.trim().replace(/\/+$/u, "");
}

function createRequestId(prefix: string): string {
  const nonce = globalThis.crypto?.randomUUID?.() ?? Math.random().toString(16).slice(2);
  return `${prefix}-${Date.now()}-${nonce.replaceAll("-", "").slice(0, 12)}`;
}
