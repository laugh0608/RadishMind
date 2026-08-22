import {
  parseWorkflowHTTPToolActionPlanDocument,
  workflowHTTPToolActionPermissions,
  type WorkflowHTTPToolActionConsumerConfig,
  type WorkflowHTTPToolActionPlan,
} from "./workflowHTTPToolActionConsumer.ts";
import {
  parseWorkflowRunRecordDocument,
  type WorkflowRunRecord,
} from "./workflowRunRecordConsumer.ts";
import { readWorkflowRunHistoryDetail } from "./workflowRunHistoryConsumer.ts";

const EXECUTION_ENVELOPE_KEYS = [
  "request_id", "workspace_id", "application_id", "action_plan", "run",
  "failure_code", "failure_summary", "audit_ref",
] as const;
const EXECUTION_SCOPES = ["workflow_tool_actions:execute", "workflow_runs:execute", "workflow_drafts:read"] as const;
const DEFINITION_EXECUTION_SCOPES = ["workflow_tool_actions:execute", "workflow_runs:execute", "workflow_definitions:read"] as const;
const FORBIDDEN_EXECUTION_FIELDS = new Set([
  "endpoint", "url", "uri", "header", "headers", "authorization", "cookie", "credential", "secret",
  "raw_query", "query_string", "raw_body", "raw_request", "raw_response", "dns", "ip", "ip_address",
  "stack", "stack_trace", "stderr", "internal_error", "provider_raw_envelope",
]);

export type WorkflowHTTPToolExecutionState = {
  status: "disabled" | "idle" | "executing" | "succeeded" | "failed" | "outcome_unknown";
  summary: string;
  failureCode: string;
  requestId: string;
  auditRef: string;
  actionPlan: WorkflowHTTPToolActionPlan | null;
  run: WorkflowRunRecord | null;
};

type WorkflowHTTPToolExecutionEnvelopeDocument = {
  request_id: string;
  workspace_id: string;
  application_id: string;
  action_plan: unknown | null;
  run: unknown | null;
  failure_code: string | null;
  failure_summary: string;
  audit_ref: string;
};

export function initialWorkflowHTTPToolExecutionState(
  config: WorkflowHTTPToolActionConsumerConfig,
): WorkflowHTTPToolExecutionState {
  return config.mode === "dev_workflow_http_tool_http"
    ? {
        status: "idle",
        summary: "Approve a durable action plan before explicitly starting its single execution attempt.",
        failureCode: "",
        requestId: "workflow-http-tool-execution-idle",
        auditRef: "audit_workflow_http_tool_execution_idle",
        actionPlan: null,
        run: null,
      }
    : {
        status: "disabled",
        summary: "Workflow HTTP Tool execution is disabled; offline mode sends no requests.",
        failureCode: "",
        requestId: "workflow-http-tool-execution-disabled",
        auditRef: "audit_workflow_http_tool_execution_disabled",
        actionPlan: null,
        run: null,
      };
}

export async function executeWorkflowHTTPToolActionPlan(
  config: WorkflowHTTPToolActionConsumerConfig,
  plan: WorkflowHTTPToolActionPlan,
  input: { inputText: string; model: string; temperature?: number },
): Promise<WorkflowHTTPToolExecutionState> {
  if (config.mode !== "dev_workflow_http_tool_http") return initialWorkflowHTTPToolExecutionState(config);
  const permissions = workflowHTTPToolActionPermissions(config);
  const executionPermission = plan.sourceKind === "workflow_definition" ? permissions.definitionExecute : permissions.execute;
  if (!executionPermission.available) {
    return failureState("workflow_run_scope_denied", `Execution requires the source-specific grants: ${executionPermission.requiredGrants.join(", ")}.`, plan);
  }
  if (plan.status !== "approved" || plan.recordVersion < 1 || !input.inputText.trim() ||
    new TextEncoder().encode(input.inputText).byteLength > 8_192 || input.model.length > 256 ||
    (input.temperature !== undefined && (!Number.isFinite(input.temperature) || input.temperature < 0 || input.temperature > 2))) {
    return failureState("workflow_run_input_invalid", "The approved plan or bounded execution input is invalid.", plan);
  }

  const requestId = createRequestId("workflow-tool-execution");
  try {
    const response = await fetch(
      `${config.baseUrl}/v1/user-workspace/workflow-tool-action-plans/${encodeURIComponent(plan.planId)}/executions`,
      {
        method: "POST",
        headers: executionHeaders(config, plan.applicationId, requestId, plan.sourceKind),
        body: JSON.stringify({
          workspace_id: config.workspaceId,
          application_id: plan.applicationId,
          expected_record_version: plan.recordVersion,
          input_text: input.inputText,
          model: input.model.trim(),
          ...(input.temperature === undefined ? {} : { temperature: input.temperature }),
        }),
      },
    );
    const body: unknown = await response.json();
    if (!response.ok || !isExecutionEnvelope(body, config, plan.applicationId)) {
      return failureState("workflow_tool_store_contract_mismatch", "The execution route returned an invalid or unsafe envelope.", plan);
    }
    const consumedPlan = body.action_plan
      ? parseWorkflowHTTPToolActionPlanDocument(body.action_plan, config, plan.applicationId)
      : null;
    const run = body.run ? parseWorkflowRunRecordDocument(body.run) : null;
    if (body.failure_code && !run) {
      return {
        status: "failed",
        summary: body.failure_summary || "Workflow HTTP Tool execution was rejected before claim.",
        failureCode: body.failure_code,
        requestId: body.request_id,
        auditRef: body.audit_ref,
        actionPlan: consumedPlan ?? plan,
        run: null,
      };
    }
    const expectedRunSchema = plan.sourceKind === "workflow_definition" ? "workflow_run_record.v9" : "workflow_run_record.v2";
    if (!consumedPlan || !run || !consumedPlanMatchesApprovedSource(consumedPlan, plan) || consumedPlan.status !== "consumed" ||
      consumedPlan.recordVersion !== plan.recordVersion + 1 || run.schemaVersion !== expectedRunSchema ||
      run.planId !== plan.planId || run.workspaceId !== config.workspaceId || run.applicationId !== plan.applicationId ||
      run.sideEffects.toolCalls !== 1 || run.sideEffects.confirmationCalls !== 1 ||
      run.sideEffects.businessWrites !== 0 || run.sideEffects.replayWrites !== 0 || !workflowHTTPToolRunMatchesPlan(run, plan)) {
      return failureState("workflow_tool_store_contract_mismatch", "The execution result did not match the approved durable plan.", plan);
    }
    return completedState(consumedPlan, run, body.request_id, body.audit_ref, body.failure_code ?? run.failureCode, body.failure_summary);
  } catch {
    return failureState("workflow_tool_store_unavailable", "The execution route is unavailable; no automatic retry was attempted.", plan);
  }
}

export async function restoreWorkflowHTTPToolExecutionState(
  config: WorkflowHTTPToolActionConsumerConfig,
  plan: WorkflowHTTPToolActionPlan,
  runId: string,
): Promise<WorkflowHTTPToolExecutionState> {
  if (config.mode !== "dev_workflow_http_tool_http" || plan.status !== "consumed" || !/^run_[a-z0-9]{16,64}$/u.test(runId)) {
    return failureState("workflow_tool_store_contract_mismatch", "The durable execution reference is invalid.", plan);
  }
  try {
    const run = await readWorkflowRunHistoryDetail({ runId }, plan.applicationId, {
      mode: "dev_workflow_executor_http",
      baseUrl: config.baseUrl,
      workspaceId: config.workspaceId,
      tenantRef: config.tenantRef,
      subjectRef: config.subjectRef,
      diagnosticsDevEnabled: false,
    });
    if (!run || !workflowHTTPToolRunMatchesPlan(run, plan)) {
      return failureState("workflow_tool_store_contract_mismatch", "The durable run no longer matches the consumed plan authority.", plan);
    }
    return completedState(plan, run, run.requestId, run.auditRef, run.failureCode, run.failureSummary);
  } catch {
    return failureState("workflow_tool_store_unavailable", "The durable run could not be restored; no retry was attempted.", plan);
  }
}

function isExecutionEnvelope(
  value: unknown,
  config: WorkflowHTTPToolActionConsumerConfig,
  applicationId: string,
): value is WorkflowHTTPToolExecutionEnvelopeDocument {
  if (!isRecord(value) || !hasExactKeys(value, EXECUTION_ENVELOPE_KEYS) || containsForbiddenExecutionField(value)) return false;
  return typeof value.request_id === "string" && value.workspace_id === config.workspaceId &&
    value.application_id === applicationId && (value.action_plan === null || isRecord(value.action_plan)) &&
    (value.run === null || isRecord(value.run)) && (value.failure_code === null || typeof value.failure_code === "string") &&
    typeof value.failure_summary === "string" && typeof value.audit_ref === "string" &&
    !(value.failure_code === null && value.run === null);
}

function executionHeaders(
  config: WorkflowHTTPToolActionConsumerConfig,
  applicationId: string,
  requestId: string,
  sourceKind: WorkflowHTTPToolActionPlan["sourceKind"],
): HeadersInit {
  const scopes = sourceKind === "workflow_definition" ? DEFINITION_EXECUTION_SCOPES : EXECUTION_SCOPES;
  return {
    Accept: "application/json",
    "Content-Type": "application/json",
    "X-Request-Id": requestId,
    "X-RadishMind-Dev-Read-Identity": "dev-workflow-http-tool-execution-consumer",
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Read-Scopes": scopes.join(","),
    "X-RadishMind-Dev-Read-Audit": "audit_dev_workflow_http_tool_execution_consumer",
    "X-RadishMind-Dev-Workflow-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Workflow-Application": applicationId,
    "X-RadishMind-Active-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Read-Membership-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Read-Membership-Permissions": scopes.join(","),
  };
}

function consumedPlanMatchesApprovedSource(
  consumed: WorkflowHTTPToolActionPlan,
  approved: WorkflowHTTPToolActionPlan,
): boolean {
  if (consumed.planId !== approved.planId || consumed.schemaVersion !== approved.schemaVersion ||
    consumed.sourceKind !== approved.sourceKind || consumed.toolPlanDigest !== approved.toolPlanDigest ||
    consumed.nodeId !== approved.nodeId || consumed.toolId !== approved.toolId) return false;
  return approved.sourceKind === "workflow_definition"
    ? consumed.workflowDefinitionId === approved.workflowDefinitionId &&
      consumed.workflowDefinitionVersion === approved.workflowDefinitionVersion &&
      consumed.workflowDefinitionDigest === approved.workflowDefinitionDigest &&
      consumed.activationPointerVersion === approved.activationPointerVersion
    : consumed.draftId === approved.draftId && consumed.draftVersion === approved.draftVersion;
}

function workflowHTTPToolRunMatchesPlan(run: WorkflowRunRecord, plan: WorkflowHTTPToolActionPlan): boolean {
  if (run.planId !== plan.planId || run.workspaceId !== plan.workspaceId || run.applicationId !== plan.applicationId ||
    run.sideEffects.toolCalls !== 1 || run.sideEffects.confirmationCalls !== 1 ||
    run.sideEffects.businessWrites !== 0 || run.sideEffects.replayWrites !== 0) return false;
  if (plan.sourceKind === "saved_workflow_draft") {
    return run.schemaVersion === "workflow_run_record.v2" && run.draftId === plan.draftId && run.draftVersion === plan.draftVersion;
  }
  const authority = run.definitionAuthority;
  const attempt = run.toolAttempt;
  return run.schemaVersion === "workflow_run_record.v9" && run.executionKind === "workflow_definition_http_tool_execution" &&
    run.executionSourceKind === "workflow_definition" && run.executionSourceId === plan.workflowDefinitionId &&
    run.executionSourceVersion === plan.workflowDefinitionVersion && run.executionProfile === "workflow_definition_http_tool_v1" &&
    run.output === "" && Boolean(authority) && authority!.definitionId === plan.workflowDefinitionId &&
    authority!.definitionVersion === plan.workflowDefinitionVersion &&
    authority!.definitionDigest === plan.workflowDefinitionDigest &&
    authority!.activationPointerVersion === plan.activationPointerVersion && Boolean(attempt) &&
    attempt!.nodeId === plan.nodeId && attempt!.toolId === plan.toolId &&
    attempt!.definitionDigest === plan.definitionDigest && attempt!.profileId === plan.profileId &&
    attempt!.profileDigest === plan.profileDigest && attempt!.toolPlanDigest === plan.toolPlanDigest;
}

function completedState(
  plan: WorkflowHTTPToolActionPlan,
  run: WorkflowRunRecord,
  requestId: string,
  auditRef: string,
  failureCode: string,
  failureSummary: string,
): WorkflowHTTPToolExecutionState {
  const outcomeUnknown = run.status === "outcome_unknown";
  const failed = Boolean(failureCode) || run.status === "failed" || run.status === "canceled";
  return {
    status: outcomeUnknown ? "outcome_unknown" : failed ? "failed" : "succeeded",
    summary: outcomeUnknown
      ? "The single attempt may have reached the remote service; manual review is required and retry remains disabled."
      : failed
        ? failureSummary || run.failureSummary || "The claimed execution ended in a stable failure state."
        : `Workflow run ${run.runId} completed after one confirmed tool attempt.`,
    failureCode: failureCode || run.failureCode,
    requestId,
    auditRef,
    actionPlan: plan,
    run,
  };
}

function failureState(
  failureCode: string,
  summary: string,
  actionPlan: WorkflowHTTPToolActionPlan | null,
): WorkflowHTTPToolExecutionState {
  return { status: "failed", summary, failureCode, requestId: "", auditRef: actionPlan?.auditRef ?? "", actionPlan, run: null };
}

function containsForbiddenExecutionField(value: unknown): boolean {
  if (Array.isArray(value)) return value.some(containsForbiddenExecutionField);
  if (!isRecord(value)) return false;
  return Object.entries(value).some(([key, nested]) =>
    FORBIDDEN_EXECUTION_FIELDS.has(key.toLowerCase()) || containsForbiddenExecutionField(nested));
}

function hasExactKeys(value: Record<string, unknown>, allowed: readonly string[]): boolean {
  const expected = new Set(allowed);
  return Object.keys(value).length === allowed.length && Object.keys(value).every((key) => expected.has(key));
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function createRequestId(prefix: string): string {
  const nonce = globalThis.crypto?.randomUUID?.() ?? Math.random().toString(16).slice(2);
  return `${prefix}-${Date.now()}-${nonce.replaceAll("-", "").slice(0, 12)}`;
}
