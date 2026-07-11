import type { WorkflowExecutorConsumerConfig, WorkflowRunRecord } from "./workflowExecutorConsumer.ts";
import { readWorkflowRunDevRecordByID } from "./workflowExecutorConsumer.ts";

export type WorkflowRunHistoryFilter = {
  status: "" | "running" | "succeeded" | "failed" | "canceled";
  draftId: string;
  startedFrom: string;
  startedTo: string;
};

export type WorkflowRunHistorySummary = {
  runId: string;
  draftId: string;
  draftVersion: number;
  status: "running" | "succeeded" | "failed" | "canceled";
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
  sideEffects: { providerCalls: number; toolCalls: 0; confirmationCalls: 0; businessWrites: 0; replayWrites: 0 };
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

type RunHistoryEnvelope = {
  request_id: string; workspace_id: string; application_id: string; runs: RunSummaryDocument[];
  next_cursor: string; has_more: boolean; failure_code: string | null; failure_summary: string; audit_ref: string;
};
type RunSummaryDocument = {
  schema_version: "workflow_run_record.v0"; record_version: number; run_id: string; draft_id: string;
  draft_version: number; workspace_id: string; application_id: string;
  status: "running" | "succeeded" | "failed" | "canceled"; failure_code: string;
  started_at: string; completed_at: string; duration_ms: number; selected_provider: string;
  selected_profile: string; selected_model: string; request_id: string; audit_ref: string; stale_running: boolean;
  side_effects: { provider_calls: number; tool_calls: number; confirmation_calls: number; business_writes: number; replay_writes: number };
};

export const EMPTY_WORKFLOW_RUN_HISTORY_FILTER: WorkflowRunHistoryFilter = { status: "", draftId: "", startedFrom: "", startedTo: "" };

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
  if (filter.startedFrom) query.set("started_from", new Date(filter.startedFrom).toISOString());
  if (filter.startedTo) query.set("started_to", new Date(filter.startedTo).toISOString());
  const response = await fetch(`${config.baseUrl}/v1/user-workspace/workflow-runs?${query}`, {
    headers: workflowRunHistoryHeaders(config, applicationId),
  });
  const body: unknown = await response.json();
  if (!response.ok || !isRunHistoryEnvelope(body)) throw new Error(`workflow run history route failed with HTTP ${response.status}`);
  if (body.failure_code) return { status: "failed", runs: [], nextCursor: "", hasMore: false, requestId: body.request_id, auditRef: body.audit_ref, failureCode: body.failure_code, failureSummary: body.failure_summary };
  const runs = body.runs.map(toSummary);
  return { status: previousRuns.length + runs.length ? "ready" : "empty", runs: [...previousRuns, ...runs], nextCursor: body.next_cursor, hasMore: body.has_more, requestId: body.request_id, auditRef: body.audit_ref, failureCode: "", failureSummary: "" };
}

export async function readWorkflowRunHistoryDetail(run: WorkflowRunHistorySummary, applicationId: string, config: WorkflowExecutorConsumerConfig): Promise<WorkflowRunRecord | null> {
  const state = await readWorkflowRunDevRecordByID(run.runId, applicationId, config);
  return state.record;
}

function workflowRunHistoryHeaders(config: WorkflowExecutorConsumerConfig, applicationId: string): HeadersInit {
  return { Accept: "application/json", "X-Request-Id": `dev-workflow-run-history-${applicationId}`, "X-RadishMind-Dev-Read-Identity": "dev-workflow-run-history-consumer", "X-RadishMind-Dev-Read-Tenant": config.tenantRef, "X-RadishMind-Dev-Read-Subject": config.subjectRef, "X-RadishMind-Dev-Read-Scopes": "workflow_runs:read", "X-RadishMind-Dev-Read-Audit": "audit_dev_workflow_run_history_consumer", "X-RadishMind-Dev-Workflow-Workspace": config.workspaceId, "X-RadishMind-Dev-Workflow-Application": applicationId };
}

function toSummary(value: RunSummaryDocument): WorkflowRunHistorySummary {
  if (value.side_effects.tool_calls || value.side_effects.confirmation_calls || value.side_effects.business_writes || value.side_effects.replay_writes) throw new Error("workflow run history contains a forbidden side effect count");
  return { runId: value.run_id, draftId: value.draft_id, draftVersion: value.draft_version, status: value.status, failureCode: value.failure_code, startedAt: value.started_at, completedAt: value.completed_at, durationMs: value.duration_ms, selectedProvider: value.selected_provider, selectedProfile: value.selected_profile, selectedModel: value.selected_model, requestId: value.request_id, auditRef: value.audit_ref, staleRunning: value.stale_running, sideEffects: { providerCalls: value.side_effects.provider_calls, toolCalls: 0, confirmationCalls: 0, businessWrites: 0, replayWrites: 0 } };
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
  for (const forbidden of ["input_text", "condition_values", "credential", "endpoint", "provider_raw_envelope"]) if (forbidden in raw) return false;
  return item.schema_version === "workflow_run_record.v0" && typeof item.record_version === "number" && typeof item.run_id === "string" && typeof item.draft_id === "string" && typeof item.draft_version === "number" && ["running", "succeeded", "failed", "canceled"].includes(item.status ?? "") && typeof item.started_at === "string" && typeof item.side_effects === "object";
}
