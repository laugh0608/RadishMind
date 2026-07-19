import assert from "node:assert/strict";
import test from "node:test";

import {
  buildApplicationOperationsSnapshot,
  initialApplicationOperationsState,
  loadApplicationOperations,
} from "../src/features/control-plane-read/applicationOperationsConsumer.ts";
import type {
  GatewayRequestHistoryState,
  ModelGatewayRequestHistoryConfig,
} from "../src/features/control-plane-read/modelGatewayRequestHistoryConsumer.ts";
import type { WorkflowExecutorConsumerConfig } from "../src/features/control-plane-read/workflowExecutorConsumer.ts";
import type { WorkflowRunHistoryState } from "../src/features/control-plane-read/workflowRunHistoryConsumer.ts";

const offlineGateway: ModelGatewayRequestHistoryConfig = {
  mode: "offline",
  baseUrl: "http://127.0.0.1:7000",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  consumerRef: "consumer_web_dev",
  applicationId: "",
  subjectRef: "subject_web_dev",
};
const liveGateway: ModelGatewayRequestHistoryConfig = {
  ...offlineGateway,
  mode: "dev_gateway_request_history_http",
};
const offlineWorkflow: WorkflowExecutorConsumerConfig = {
  mode: "disabled",
  baseUrl: "http://127.0.0.1:7000",
  workspaceId: "workspace_demo",
  tenantRef: "tenant_demo",
  subjectRef: "subject_web_dev",
};
const liveWorkflow: WorkflowExecutorConsumerConfig = {
  ...offlineWorkflow,
  mode: "dev_workflow_executor_http",
};

test("application operations stays offline and rejects a missing application without fetching", async () => {
  const originalFetch = globalThis.fetch;
  let calls = 0;
  globalThis.fetch = async () => {
    calls += 1;
    throw new Error("unexpected fetch");
  };
  try {
    assert.equal(initialApplicationOperationsState("app_demo", offlineGateway, offlineWorkflow).status, "offline");
    assert.equal((await loadApplicationOperations("app_demo", offlineGateway, offlineWorkflow)).status, "offline");
    assert.equal((await loadApplicationOperations("", liveGateway, liveWorkflow)).status, "application_unavailable");
    assert.equal(calls, 0);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("application operations keeps channel attribution separate and sorts one loaded window", () => {
  const snapshot = buildApplicationOperationsSnapshot("app_demo", gatewayReadyState(), workflowReadyState());
  assert.equal(snapshot.status, "ready");
  assert.equal(snapshot.loadedWindowComplete, false);
  assert.equal(snapshot.metrics.gatewayLoaded, 2);
  assert.equal(snapshot.metrics.gatewaySucceeded, 1);
  assert.equal(snapshot.metrics.gatewayFailed, 1);
  assert.equal(snapshot.metrics.gatewayUsageReported, 1);
  assert.equal(snapshot.metrics.gatewayUsageNotReported, 1);
  assert.equal(snapshot.metrics.workflowLoaded, 2);
  assert.equal(snapshot.metrics.workflowSucceeded, 1);
  assert.equal(snapshot.metrics.workflowFailed, 1);
  assert.equal(snapshot.metrics.workflowProviderCalls, 2);
  assert.equal(snapshot.metrics.workflowRetrievalCalls, 1);
  assert.equal(snapshot.metrics.workflowToolCalls, 1);
  assert.deepEqual(snapshot.timeline.map((entry) => entry.source), [
    "workflow_run",
    "gateway_request",
    "workflow_run",
    "gateway_request",
  ]);
  assert.equal(snapshot.timeline.filter((entry) => entry.requestId === "shared_request").length, 2);
});

test("application operations marks coverage complete only for successful exhausted channels", () => {
  const gateway = gatewayReadyState();
  gateway.nextCursor = "";
  gateway.hasMore = false;
  const workflow = workflowReadyState();
  const complete = buildApplicationOperationsSnapshot("app_demo", gateway, workflow);
  assert.equal(complete.loadedWindowComplete, true);

  gateway.status = "failed";
  gateway.requests = [];
  const partial = buildApplicationOperationsSnapshot("app_demo", gateway, workflow);
  assert.equal(partial.status, "partial_failure");
  assert.equal(partial.loadedWindowComplete, false);
  assert.equal(initialApplicationOperationsState("app_demo", offlineGateway, offlineWorkflow).loadedWindowComplete, false);
});

test("application operations preserves workflow observations when Gateway history fails", async () => {
  const originalFetch = globalThis.fetch;
  const paths: string[] = [];
  globalThis.fetch = async (input) => {
    const url = new URL(String(input));
    paths.push(`${url.pathname}?${url.searchParams}`);
    assert.equal(url.searchParams.get("application_id"), "app_demo");
    if (url.pathname === "/v1/model-gateway/requests") {
      return new Response("{}", { status: 503, headers: { "Content-Type": "application/json" } });
    }
    return workflowListResponse();
  };
  try {
    const state = await loadApplicationOperations("app_demo", liveGateway, liveWorkflow);
    assert.equal(paths.length, 2);
    assert.equal(state.status, "partial_failure");
    assert.equal(state.gateway.status, "failed");
    assert.equal(state.workflow.status, "ready");
    assert.equal(state.timeline.length, 1);
    assert.equal(state.timeline[0]?.source, "workflow_run");
    assert.equal(state.failureSummary, "Gateway request observations are unavailable for this application.");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("application operations reports failure only when every enabled source fails", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => new Response("{}", {
    status: 503,
    headers: { "Content-Type": "application/json" },
  });
  try {
    const state = await loadApplicationOperations("app_demo", liveGateway, liveWorkflow);
    assert.equal(state.status, "failed");
    assert.equal(state.timeline.length, 0);
    assert.match(state.failureSummary, /Gateway request observations/);
    assert.match(state.failureSummary, /Workflow run observations/);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

function gatewayReadyState(): GatewayRequestHistoryState {
  return {
    status: "ready",
    requests: [
      {
        schemaVersion: "gateway_request_record.v1",
        recordVersion: 2,
        storeMode: "sqlite_dev",
        requestId: "shared_request",
        auditRef: "audit_gateway_new",
        route: "/v1/responses",
        protocol: "openai-responses",
        stream: false,
        status: "succeeded",
        startedAt: "2026-07-19T10:03:00Z",
        completedAt: "2026-07-19T10:03:01Z",
        durationMs: 1000,
        providerDurationMs: 900,
        providerDurationAvailable: true,
        selectionSource: "explicit_model",
        selectedProvider: "mock",
        selectedProfile: "profile_mock",
        selectedModel: "mock-model",
        httpStatusCode: 200,
        failureCode: "",
        failureBoundary: "none",
        usageAvailability: "reported",
        staleStarted: false,
      },
      {
        schemaVersion: "gateway_request_record.v1",
        recordVersion: 2,
        storeMode: "sqlite_dev",
        requestId: "gateway_failed",
        auditRef: "audit_gateway_old",
        route: "/v1/messages",
        protocol: "anthropic-messages",
        stream: false,
        status: "failed",
        startedAt: "2026-07-19T10:00:00Z",
        completedAt: "2026-07-19T10:00:01Z",
        durationMs: 1000,
        providerDurationMs: 0,
        providerDurationAvailable: false,
        selectionSource: "explicit_model",
        selectedProvider: "mock",
        selectedProfile: "profile_mock",
        selectedModel: "mock-model",
        httpStatusCode: 503,
        failureCode: "provider_failed",
        failureBoundary: "provider",
        usageAvailability: "not_reported",
        staleStarted: false,
      },
    ],
    nextCursor: "gateway_cursor",
    hasMore: true,
    requestId: "gateway_list",
    auditRef: "audit_gateway_list",
    failureCode: "",
    failureSummary: "",
  };
}

function workflowReadyState(): WorkflowRunHistoryState {
  return {
    status: "ready",
    runs: [
      workflowSummary("workflow_new", "shared_request", "succeeded", "2026-07-19T10:04:00Z", {
        retrievalCalls: 1,
        providerCalls: 1,
        toolCalls: 0,
        confirmationCalls: 0,
        businessWrites: 0,
        replayWrites: 0,
      }),
      workflowSummary("workflow_failed", "workflow_request", "failed", "2026-07-19T10:02:00Z", {
        retrievalCalls: 0,
        providerCalls: 1,
        toolCalls: 1,
        confirmationCalls: 1,
        businessWrites: 0,
        replayWrites: 0,
      }),
    ],
    nextCursor: "",
    hasMore: false,
    requestId: "workflow_list",
    auditRef: "audit_workflow_list",
    failureCode: "",
    failureSummary: "",
  };
}

function workflowSummary(
  runId: string,
  requestId: string,
  status: "succeeded" | "failed",
  startedAt: string,
  sideEffects: {
    retrievalCalls: number;
    providerCalls: number;
    toolCalls: number;
    confirmationCalls: number;
    businessWrites: number;
    replayWrites: number;
  },
) {
  return {
    schemaVersion: "workflow_run_record.v4" as const,
    runId,
    planId: "",
    confirmationId: "",
    toolAttemptStatus: "" as const,
    draftId: "",
    draftVersion: 0,
    draftDigest: "",
    executionKind: "application_rag_invocation",
    executionSourceKind: "application_configuration_draft",
    executionSourceId: "draft_app",
    executionSourceVersion: 2,
    runtimeAssignmentId: "assignment_app",
    runtimeAssignmentVersion: 1,
    publishCandidateId: "candidate_app",
    publishReviewVersion: 1,
    effectiveSnapshotRole: "candidate",
    status,
    failureCode: status === "failed" ? "workflow_run_gateway_failed" : "",
    startedAt,
    completedAt: startedAt,
    durationMs: 100,
    selectedProvider: "mock",
    selectedProfile: "profile_mock",
    selectedModel: "mock-model",
    requestId,
    auditRef: `audit_${runId}`,
    staleRunning: false,
    failureBoundary: status === "failed" ? "gateway" : "none",
    failedNodeId: "",
    lastCompletedNodeId: "",
    gatewayFailureCategory: status === "failed" ? "provider_failed" : "none",
    toolFailureCategory: "",
    snapshotId: "snapshot_app",
    snapshotVersion: 1,
    snapshotDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    ragRef: "rag_ref_app",
    retrievalNodeId: "application_rag_retrieval",
    retrievalAttemptStatus: "succeeded",
    retrievalProfileId: "profile_lexical",
    retrievalProfileVersion: 1,
    retrievalProfileDigest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
    queryDigest: "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
    queryBytes: 12,
    candidateCount: 1,
    selectedFragments: [],
    citationRefs: [],
    retrievalLatencyMs: 2,
    retrievalContextBytes: 32,
    retrievalFailureCategory: "none",
    recommendedReviewAction: "none",
    sideEffects,
  };
}

function workflowListResponse(): Response {
  return new Response(JSON.stringify({
    request_id: "workflow_list",
    workspace_id: "workspace_demo",
    application_id: "app_demo",
    runs: [{
      schema_version: "workflow_run_record.v0",
      record_version: 2,
      run_id: "workflow_live",
      draft_id: "draft_live",
      draft_version: 1,
      workspace_id: "workspace_demo",
      application_id: "app_demo",
      status: "succeeded",
      failure_code: "",
      started_at: "2026-07-19T10:00:00Z",
      completed_at: "2026-07-19T10:00:01Z",
      duration_ms: 1000,
      selected_provider: "mock",
      selected_profile: "profile_mock",
      selected_model: "mock-model",
      request_id: "workflow_request",
      audit_ref: "workflow_audit",
      stale_running: false,
      side_effects: {
        provider_calls: 1,
        tool_calls: 0,
        confirmation_calls: 0,
        business_writes: 0,
        replay_writes: 0,
      },
    }],
    next_cursor: "",
    has_more: false,
    failure_code: null,
    failure_summary: "",
    audit_ref: "workflow_list_audit",
  }), { status: 200, headers: { "Content-Type": "application/json" } });
}
