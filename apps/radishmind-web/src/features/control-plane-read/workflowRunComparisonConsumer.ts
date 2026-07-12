import type { WorkflowExecutorConsumerConfig } from "./workflowExecutorConsumer.ts";

export type WorkflowRunComparisonRun = {
  runId: string; schemaVersion: string; draftId: string; draftVersion: number;
  status: "running" | "succeeded" | "failed" | "canceled"; failureCode: string;
  failureBoundary: string; gatewayFailureCategory: string; selectedProvider: string;
  selectedProfile: string; selectedModel: string; durationMs: number; staleRunning: boolean;
  requestId: string; auditRef: string;
  sideEffects: { providerCalls: number; toolCalls: 0; confirmationCalls: 0; businessWrites: 0; replayWrites: 0 };
};

export type WorkflowRunNodeComparison = {
  nodeId: string; nodeType: string; change: "added" | "removed" | "changed" | "unchanged";
  baselineStatus: string; candidateStatus: string; baselineDurationMs: number;
  candidateDurationMs: number; durationDeltaMs: number;
};

export type WorkflowRunComparison = {
  schemaVersion: "workflow_run_comparison.v1";
  classification: "regression" | "improvement" | "changed" | "unchanged" | "inconclusive";
  comparisonState: "comparable" | "legacy_partial" | "running_inconclusive";
  baseline: WorkflowRunComparisonRun; candidate: WorkflowRunComparisonRun;
  draftChanged: boolean; providerChanged: boolean; modelChanged: boolean; statusChanged: boolean;
  failureChanged: boolean; durationDeltaMs: number; providerCallDelta: number;
  nodes: WorkflowRunNodeComparison[]; findings: Array<{ code: string; severity: "info" | "review_required" }>;
  recommendedReviewAction: string;
};

type ComparisonEnvelope = {
  request_id: string; workspace_id: string; application_id: string;
  comparison: ComparisonDocument | null; failure_code: string | null; failure_summary: string; audit_ref: string;
};

type ComparisonRunDocument = {
  run_id: string; schema_version: string; draft_id: string; draft_version: number; status: WorkflowRunComparisonRun["status"];
  failure_code: string; failure_boundary: string; gateway_failure_category: string; selected_provider: string;
  selected_profile: string; selected_model: string; duration_ms: number; stale_running: boolean; request_id: string; audit_ref: string;
  side_effects: { provider_calls: number; tool_calls: number; confirmation_calls: number; business_writes: number; replay_writes: number };
};

type ComparisonDocument = {
  schema_version: "workflow_run_comparison.v1"; classification: WorkflowRunComparison["classification"];
  comparison_state: WorkflowRunComparison["comparisonState"]; baseline: ComparisonRunDocument; candidate: ComparisonRunDocument;
  draft_changed: boolean; provider_changed: boolean; model_changed: boolean; status_changed: boolean; failure_changed: boolean;
  duration_delta_ms: number; provider_call_delta: number;
  nodes: Array<{ node_id: string; node_type: string; change: WorkflowRunNodeComparison["change"]; baseline_status: string; candidate_status: string; baseline_duration_ms: number; candidate_duration_ms: number; duration_delta_ms: number }>;
  findings: Array<{ code: string; severity: "info" | "review_required" }>; recommended_review_action: string;
};

export async function compareWorkflowRuns(applicationId: string, baselineRunId: string, candidateRunId: string, config: WorkflowExecutorConsumerConfig): Promise<WorkflowRunComparison> {
  if (config.mode !== "dev_workflow_executor_http") throw new Error("Workflow run comparison is unavailable in offline mode.");
  if (!baselineRunId || !candidateRunId || baselineRunId === candidateRunId) throw new Error("Choose two different workflow runs.");
  const query = new URLSearchParams({ baseline_run_id: baselineRunId, workspace_id: config.workspaceId, application_id: applicationId });
  const response = await fetch(`${config.baseUrl}/v1/user-workspace/workflow-runs/${encodeURIComponent(candidateRunId)}/comparison?${query}`, { headers: comparisonHeaders(config, applicationId) });
  const body: unknown = await response.json();
  if (!response.ok || !isComparisonEnvelope(body)) throw new Error(`workflow run comparison route failed with HTTP ${response.status}`);
  if (body.failure_code || !body.comparison) throw new Error(`${body.failure_code ?? "workflow_run_comparison_unavailable"}: ${body.failure_summary}`);
  return mapComparison(body.comparison);
}

function comparisonHeaders(config: WorkflowExecutorConsumerConfig, applicationId: string): HeadersInit {
  return { Accept: "application/json", "X-Request-Id": `dev-workflow-run-comparison-${applicationId}`, "X-RadishMind-Dev-Read-Identity": "dev-workflow-run-comparison-consumer", "X-RadishMind-Dev-Read-Tenant": config.tenantRef, "X-RadishMind-Dev-Read-Subject": config.subjectRef, "X-RadishMind-Dev-Read-Scopes": "workflow_runs:read", "X-RadishMind-Dev-Read-Audit": "audit_dev_workflow_run_comparison_consumer", "X-RadishMind-Dev-Workflow-Workspace": config.workspaceId, "X-RadishMind-Dev-Workflow-Application": applicationId };
}

function mapRun(value: ComparisonRunDocument): WorkflowRunComparisonRun {
  const sideEffects = value.side_effects;
  if (sideEffects.tool_calls || sideEffects.confirmation_calls || sideEffects.business_writes || sideEffects.replay_writes) throw new Error("workflow run comparison contains a forbidden side effect count");
  return { runId: value.run_id, schemaVersion: value.schema_version, draftId: value.draft_id, draftVersion: value.draft_version, status: value.status, failureCode: value.failure_code, failureBoundary: value.failure_boundary, gatewayFailureCategory: value.gateway_failure_category, selectedProvider: value.selected_provider, selectedProfile: value.selected_profile, selectedModel: value.selected_model, durationMs: value.duration_ms, staleRunning: value.stale_running, requestId: value.request_id, auditRef: value.audit_ref, sideEffects: { providerCalls: sideEffects.provider_calls, toolCalls: 0, confirmationCalls: 0, businessWrites: 0, replayWrites: 0 } };
}

function mapComparison(value: ComparisonDocument): WorkflowRunComparison {
  return { schemaVersion: value.schema_version, classification: value.classification, comparisonState: value.comparison_state, baseline: mapRun(value.baseline), candidate: mapRun(value.candidate), draftChanged: value.draft_changed, providerChanged: value.provider_changed, modelChanged: value.model_changed, statusChanged: value.status_changed, failureChanged: value.failure_changed, durationDeltaMs: value.duration_delta_ms, providerCallDelta: value.provider_call_delta, nodes: value.nodes.map((node) => ({ nodeId: node.node_id, nodeType: node.node_type, change: node.change, baselineStatus: node.baseline_status, candidateStatus: node.candidate_status, baselineDurationMs: node.baseline_duration_ms, candidateDurationMs: node.candidate_duration_ms, durationDeltaMs: node.duration_delta_ms })), findings: value.findings, recommendedReviewAction: value.recommended_review_action };
}

function isComparisonEnvelope(value: unknown): value is ComparisonEnvelope {
  if (!value || typeof value !== "object") return false;
  if (hasForbiddenComparisonKey(value)) return false;
  const item = value as Partial<ComparisonEnvelope>;
  return typeof item.request_id === "string" && typeof item.application_id === "string" && (item.failure_code === null || typeof item.failure_code === "string") && typeof item.failure_summary === "string" && typeof item.audit_ref === "string" && (item.comparison === null || isComparisonDocument(item.comparison));
}

function hasForbiddenComparisonKey(value: unknown): boolean {
  if (Array.isArray(value)) return value.some(hasForbiddenComparisonKey);
  if (!value || typeof value !== "object") return false;
  const forbidden = new Set(["input_text", "input_bytes", "condition_values", "condition_node_ids", "credential", "endpoint", "provider_raw_envelope", "output", "output_preview", "actor_ref"]);
  return Object.entries(value as Record<string, unknown>).some(([key, child]) => forbidden.has(key) || hasForbiddenComparisonKey(child));
}

function isComparisonDocument(value: unknown): value is ComparisonDocument {
  if (!value || typeof value !== "object") return false;
  const item = value as Partial<ComparisonDocument>;
  return item.schema_version === "workflow_run_comparison.v1" && ["regression", "improvement", "changed", "unchanged", "inconclusive"].includes(item.classification ?? "") && ["comparable", "legacy_partial", "running_inconclusive"].includes(item.comparison_state ?? "") && isComparisonRun(item.baseline) && isComparisonRun(item.candidate) && Array.isArray(item.nodes) && Array.isArray(item.findings);
}

function isComparisonRun(value: unknown): value is ComparisonRunDocument {
  if (!value || typeof value !== "object") return false;
  const item = value as Partial<ComparisonRunDocument>;
  return typeof item.run_id === "string" && typeof item.draft_id === "string" && typeof item.draft_version === "number" && ["running", "succeeded", "failed", "canceled"].includes(item.status ?? "") && typeof item.duration_ms === "number" && typeof item.side_effects === "object";
}
