import {
  EMPTY_GATEWAY_REQUEST_HISTORY_FILTER,
  initialGatewayRequestHistoryState,
  listGatewayRequestHistory,
  type GatewayRequestHistoryState,
  type GatewayRequestHistorySummary,
  type ModelGatewayRequestHistoryConfig,
} from "./modelGatewayRequestHistoryConsumer.ts";
import type { WorkflowExecutorConsumerConfig } from "./workflowExecutorConsumer.ts";
import {
  EMPTY_WORKFLOW_RUN_HISTORY_FILTER,
  initialWorkflowRunHistoryState,
  listWorkflowRunHistory,
  type WorkflowRunHistoryState,
  type WorkflowRunHistorySummary,
} from "./workflowRunHistoryConsumer.ts";

export type ApplicationOperationsStatus =
  | "application_unavailable"
  | "offline"
  | "loading"
  | "ready"
  | "empty"
  | "partial_failure"
  | "failed";

export type ApplicationOperationsTimelineEntry = {
  source: "gateway_request" | "workflow_run";
  recordId: string;
  startedAt: string;
  completedAt: string;
  status: string;
  operation: string;
  contract: string;
  provider: string;
  profile: string;
  model: string;
  durationMs: number;
  failureCode: string;
  failureBoundary: string;
  requestId: string;
  auditRef: string;
  usageAvailability: "" | "reported" | "not_reported" | "not_applicable";
  providerCalls: number;
  retrievalCalls: number;
  toolCalls: number;
  confirmationCalls: number;
  businessWrites: number;
  replayWrites: number;
};

export type ApplicationOperationsMetrics = {
  gatewayLoaded: number;
  gatewayStarted: number;
  gatewaySucceeded: number;
  gatewayFailed: number;
  gatewayCanceled: number;
  gatewayUsageReported: number;
  gatewayUsageNotReported: number;
  gatewayUsageNotApplicable: number;
  workflowLoaded: number;
  workflowRunning: number;
  workflowSucceeded: number;
  workflowFailed: number;
  workflowCanceled: number;
  workflowOutcomeUnknown: number;
  workflowProviderCalls: number;
  workflowRetrievalCalls: number;
  workflowToolCalls: number;
  workflowConfirmationCalls: number;
  workflowBusinessWrites: number;
  workflowReplayWrites: number;
};

export type ApplicationOperationsState = {
  status: ApplicationOperationsStatus;
  applicationId: string;
  gateway: GatewayRequestHistoryState;
  workflow: WorkflowRunHistoryState;
  metrics: ApplicationOperationsMetrics;
  timeline: ApplicationOperationsTimelineEntry[];
  loadedWindowComplete: boolean;
  failureSummary: string;
};

export function initialApplicationOperationsState(
  applicationId: string,
  gatewayConfig: ModelGatewayRequestHistoryConfig,
  workflowConfig: WorkflowExecutorConsumerConfig,
): ApplicationOperationsState {
  const scopedApplicationId = applicationId.trim();
  const gateway = initialGatewayRequestHistoryState({ ...gatewayConfig, applicationId: scopedApplicationId });
  const workflow = initialWorkflowRunHistoryState(workflowConfig);
  if (!validApplicationId(scopedApplicationId)) {
    return assembleApplicationOperationsState("application_unavailable", "", gateway, workflow);
  }
  return assembleApplicationOperationsState(undefined, scopedApplicationId, gateway, workflow);
}

export async function loadApplicationOperations(
  applicationId: string,
  gatewayConfig: ModelGatewayRequestHistoryConfig,
  workflowConfig: WorkflowExecutorConsumerConfig,
): Promise<ApplicationOperationsState> {
  const scopedApplicationId = applicationId.trim();
  if (!validApplicationId(scopedApplicationId)) {
    return initialApplicationOperationsState("", gatewayConfig, workflowConfig);
  }

  const scopedGatewayConfig = { ...gatewayConfig, applicationId: scopedApplicationId };
  const [gateway, workflow] = await Promise.all([
    loadGatewayChannel(scopedGatewayConfig),
    loadWorkflowChannel(scopedApplicationId, workflowConfig),
  ]);
  return assembleApplicationOperationsState(undefined, scopedApplicationId, gateway, workflow);
}

export function buildApplicationOperationsSnapshot(
  applicationId: string,
  gateway: GatewayRequestHistoryState,
  workflow: WorkflowRunHistoryState,
): ApplicationOperationsState {
  return assembleApplicationOperationsState(undefined, applicationId.trim(), gateway, workflow);
}

async function loadGatewayChannel(config: ModelGatewayRequestHistoryConfig): Promise<GatewayRequestHistoryState> {
  try {
    return await listGatewayRequestHistory(config, EMPTY_GATEWAY_REQUEST_HISTORY_FILTER);
  } catch {
    return {
      status: "failed",
      requests: [],
      nextCursor: "",
      hasMore: false,
      requestId: "application-operations-gateway-failed",
      auditRef: "audit_application_operations_gateway_failed",
      failureCode: "application_operations_gateway_unavailable",
      failureSummary: "Gateway request observations are unavailable for this application.",
    };
  }
}

async function loadWorkflowChannel(
  applicationId: string,
  config: WorkflowExecutorConsumerConfig,
): Promise<WorkflowRunHistoryState> {
  try {
    return await listWorkflowRunHistory(applicationId, config, EMPTY_WORKFLOW_RUN_HISTORY_FILTER);
  } catch {
    return {
      status: "failed",
      runs: [],
      nextCursor: "",
      hasMore: false,
      requestId: "application-operations-workflow-failed",
      auditRef: "audit_application_operations_workflow_failed",
      failureCode: "application_operations_workflow_unavailable",
      failureSummary: "Workflow run observations are unavailable for this application.",
    };
  }
}

function assembleApplicationOperationsState(
  forcedStatus: ApplicationOperationsStatus | undefined,
  applicationId: string,
  gateway: GatewayRequestHistoryState,
  workflow: WorkflowRunHistoryState,
): ApplicationOperationsState {
  const metrics = buildMetrics(gateway.requests, workflow.runs);
  const timeline = [
    ...gateway.requests.map(gatewayTimelineEntry),
    ...workflow.runs.map(workflowTimelineEntry),
  ].sort(compareTimelineEntries);
  const status = forcedStatus ?? resolveStatus(gateway, workflow, timeline.length);
  const failures = [gateway, workflow]
    .filter((channel) => channel.status === "failed")
    .map((channel) => channel.failureSummary)
    .filter(Boolean);
  return {
    status,
    applicationId,
    gateway,
    workflow,
    metrics,
    timeline,
    loadedWindowComplete: loadedWindowIsComplete(gateway, workflow),
    failureSummary: failures.join(" "),
  };
}

function loadedWindowIsComplete(
  gateway: GatewayRequestHistoryState,
  workflow: WorkflowRunHistoryState,
): boolean {
  const enabled = [gateway, workflow].filter((channel) => channel.status !== "offline");
  return enabled.length > 0 && enabled.every((channel) =>
    (channel.status === "ready" || channel.status === "empty") && !channel.hasMore
  );
}

function resolveStatus(
  gateway: GatewayRequestHistoryState,
  workflow: WorkflowRunHistoryState,
  entryCount: number,
): ApplicationOperationsStatus {
  const channels = [gateway, workflow];
  const enabled = channels.filter((channel) => channel.status !== "offline");
  if (enabled.length === 0) return "offline";
  if (enabled.some((channel) => channel.status === "loading")) return "loading";
  const failed = enabled.filter((channel) => channel.status === "failed").length;
  if (failed === enabled.length) return "failed";
  if (failed > 0) return "partial_failure";
  return entryCount > 0 ? "ready" : "empty";
}

function buildMetrics(
  requests: GatewayRequestHistorySummary[],
  runs: WorkflowRunHistorySummary[],
): ApplicationOperationsMetrics {
  const metrics: ApplicationOperationsMetrics = {
    gatewayLoaded: requests.length,
    gatewayStarted: 0,
    gatewaySucceeded: 0,
    gatewayFailed: 0,
    gatewayCanceled: 0,
    gatewayUsageReported: 0,
    gatewayUsageNotReported: 0,
    gatewayUsageNotApplicable: 0,
    workflowLoaded: runs.length,
    workflowRunning: 0,
    workflowSucceeded: 0,
    workflowFailed: 0,
    workflowCanceled: 0,
    workflowOutcomeUnknown: 0,
    workflowProviderCalls: 0,
    workflowRetrievalCalls: 0,
    workflowToolCalls: 0,
    workflowConfirmationCalls: 0,
    workflowBusinessWrites: 0,
    workflowReplayWrites: 0,
  };
  for (const request of requests) {
    if (request.status === "started") metrics.gatewayStarted += 1;
    if (request.status === "succeeded") metrics.gatewaySucceeded += 1;
    if (request.status === "failed") metrics.gatewayFailed += 1;
    if (request.status === "canceled") metrics.gatewayCanceled += 1;
    if (request.usageAvailability === "reported") metrics.gatewayUsageReported += 1;
    if (request.usageAvailability === "not_reported") metrics.gatewayUsageNotReported += 1;
    if (request.usageAvailability === "not_applicable") metrics.gatewayUsageNotApplicable += 1;
  }
  for (const run of runs) {
    if (run.status === "running") metrics.workflowRunning += 1;
    if (run.status === "succeeded") metrics.workflowSucceeded += 1;
    if (run.status === "failed") metrics.workflowFailed += 1;
    if (run.status === "canceled") metrics.workflowCanceled += 1;
    if (run.status === "outcome_unknown") metrics.workflowOutcomeUnknown += 1;
    metrics.workflowProviderCalls += run.sideEffects.providerCalls;
    metrics.workflowRetrievalCalls += run.sideEffects.retrievalCalls;
    metrics.workflowToolCalls += run.sideEffects.toolCalls;
    metrics.workflowConfirmationCalls += run.sideEffects.confirmationCalls;
    metrics.workflowBusinessWrites += run.sideEffects.businessWrites;
    metrics.workflowReplayWrites += run.sideEffects.replayWrites;
  }
  return metrics;
}

function gatewayTimelineEntry(request: GatewayRequestHistorySummary): ApplicationOperationsTimelineEntry {
  return {
    source: "gateway_request",
    recordId: request.requestId,
    startedAt: request.startedAt,
    completedAt: request.completedAt,
    status: request.status,
    operation: request.route,
    contract: request.protocol,
    provider: request.selectedProvider,
    profile: request.selectedProfile,
    model: request.selectedModel,
    durationMs: request.durationMs,
    failureCode: request.failureCode,
    failureBoundary: request.failureBoundary,
    requestId: request.requestId,
    auditRef: request.auditRef,
    usageAvailability: request.usageAvailability,
    providerCalls: 0,
    retrievalCalls: 0,
    toolCalls: 0,
    confirmationCalls: 0,
    businessWrites: 0,
    replayWrites: 0,
  };
}

function workflowTimelineEntry(run: WorkflowRunHistorySummary): ApplicationOperationsTimelineEntry {
  return {
    source: "workflow_run",
    recordId: run.runId,
    startedAt: run.startedAt,
    completedAt: run.completedAt,
    status: run.status,
    operation: run.executionKind || run.executionSourceKind || "workflow_execution",
    contract: run.schemaVersion,
    provider: run.selectedProvider,
    profile: run.selectedProfile,
    model: run.selectedModel,
    durationMs: run.durationMs,
    failureCode: run.failureCode,
    failureBoundary: run.failureBoundary,
    requestId: run.requestId,
    auditRef: run.auditRef,
    usageAvailability: "",
    providerCalls: run.sideEffects.providerCalls,
    retrievalCalls: run.sideEffects.retrievalCalls,
    toolCalls: run.sideEffects.toolCalls,
    confirmationCalls: run.sideEffects.confirmationCalls,
    businessWrites: run.sideEffects.businessWrites,
    replayWrites: run.sideEffects.replayWrites,
  };
}

function compareTimelineEntries(left: ApplicationOperationsTimelineEntry, right: ApplicationOperationsTimelineEntry): number {
  const timeDifference = timestampValue(right.startedAt) - timestampValue(left.startedAt);
  if (timeDifference !== 0) return timeDifference;
  const sourceDifference = left.source.localeCompare(right.source);
  return sourceDifference !== 0 ? sourceDifference : left.recordId.localeCompare(right.recordId);
}

function timestampValue(value: string): number {
  const timestamp = Date.parse(value);
  return Number.isFinite(timestamp) ? timestamp : 0;
}

function validApplicationId(value: string): boolean {
  return /^[A-Za-z0-9][A-Za-z0-9._:-]{0,159}$/.test(value);
}
