import assert from "node:assert/strict";
import test from "node:test";
import { reduceApplicationResultArtifactInspection as reduce, type ApplicationResultArtifactInspection } from "../src/features/control-plane-read/applicationResultArtifactInspection.ts";
import type { ApplicationResultArtifactReadResult, ApplicationResultArtifactSummary } from "../src/features/control-plane-read/applicationResultArtifactConsumer.ts";

const empty: ApplicationResultArtifactInspection = { generation: 0, summary: null, readResult: null };
function summary(id: string): ApplicationResultArtifactSummary {
  return {
    schemaVersion: "application_result_artifact_summary.v2", artifactId: id, recordVersion: 1,
    tenantRef: "tenant_demo", workspaceId: "workspace_demo", applicationId: "app_demo", ownerSubjectRef: "subject_demo_user",
    sessionId: "appsess_demo", turnId: `turn_${id}`, clientTurnKey: `key_${id}`, executionProfile: "prompt_application_invocation_v1",
    runRef: { schemaVersion: "workflow_run_record.v6", runId: `run_${id}` }, contentType: "application/json",
    contentBytes: 2, contentDigest: `sha256:${"a".repeat(64)}`, createdAt: "2026-10-15T00:00:00Z",
    lifecycleState: "active", lifecycleVersion: 1, archivedAt: null, lifecycleUpdatedAt: "2026-10-15T00:00:00Z",
  };
}
function read(item: ApplicationResultArtifactSummary): ApplicationResultArtifactReadResult {
  return {
    status: "ready", failureCode: "", requestId: "request_read", auditRef: "audit_read", summary: "",
    artifact: { ...item, schemaVersion: "application_result_artifact.v1", content: `Answer ${item.artifactId}`, createdByActorRef: item.ownerSubjectRef, requestId: "request_save", auditRef: "audit_save" },
    lifecycle: {
      schemaVersion: "application_result_artifact_lifecycle.v1", tenantRef: item.tenantRef, workspaceId: item.workspaceId,
      applicationId: item.applicationId, ownerSubjectRef: item.ownerSubjectRef, artifactId: item.artifactId,
      lifecycleState: "active", lifecycleVersion: 1, archivedAt: null, updatedAt: item.createdAt,
      updatedByActorRef: item.ownerSubjectRef, requestId: "request_read", auditRef: "audit_read",
    },
  };
}

test("a new saved result clears the previous answer and run until the exact result is read", () => {
  const a = summary("artifact_a"), b = summary("artifact_b");
  let state = reduce(empty, { type: "select", generation: 1, summary: a });
  state = reduce(state, { type: "read", generation: 1, result: read(a) });
  assert.equal(state.readResult?.artifact?.runRef.runId, "run_artifact_a");
  state = reduce(state, { type: "select", generation: 2, summary: b });
  assert.equal(state.summary?.artifactId, b.artifactId);
  assert.equal(state.readResult, null);
  assert.equal(reduce(state, { type: "read", generation: 1, result: read(a) }), state);
  assert.equal(reduce(state, { type: "read", generation: 2, result: read(a) }), state);
  state = reduce(state, { type: "read", generation: 2, result: read(b) });
  assert.equal(state.readResult?.artifact?.content, "Answer artifact_b");
  assert.equal(state.readResult?.artifact?.runRef.runId, "run_artifact_b");
});

test("request generations reject late reads even after selecting the same asset again", () => {
  const a = summary("artifact_a");
  let state = reduce(empty, { type: "select", generation: 1, summary: a });
  state = reduce(state, { type: "select", generation: 2, summary: summary("artifact_b") });
  state = reduce(state, { type: "select", generation: 3, summary: a });
  assert.equal(reduce(state, { type: "read", generation: 1, result: read(a) }), state);
  assert.equal(reduce(state, { type: "select", generation: 2, summary: null }), state);
  const failed: ApplicationResultArtifactReadResult = { ...read(a), status: "failed", artifact: null, lifecycle: null, failureCode: "application_result_artifact_store_unavailable" };
  state = reduce(state, { type: "read", generation: 3, result: failed });
  assert.equal(state.readResult?.failureCode, failed.failureCode);
  assert.equal(state.readResult?.artifact, null);
});

test("lifecycle changes only update the selected and explicitly read asset", () => {
  const a = summary("artifact_a"), b = summary("artifact_b");
  const archived = { ...read(b).lifecycle!, lifecycleState: "archived" as const, lifecycleVersion: 2, archivedAt: b.createdAt };
  let state = reduce(empty, { type: "select", generation: 1, summary: b });
  assert.equal(reduce(state, { type: "transition", generation: 1, lifecycle: archived }), state);
  assert.equal(reduce(state, { type: "read", generation: 1, result: { ...read(b), lifecycle: read(a).lifecycle } }), state);
  state = reduce(state, { type: "read", generation: 1, result: read(b) });
  state = reduce(state, { type: "begin", generation: 2 });
  assert.equal(reduce(state, { type: "transition", generation: 1, lifecycle: archived }), state);
  assert.equal(reduce(state, { type: "transition", generation: 2, lifecycle: { ...archived, artifactId: a.artifactId } }), state);
  state = reduce(state, { type: "transition", generation: 2, lifecycle: archived });
  assert.equal(state.summary?.lifecycleVersion, 2);
  assert.equal(state.readResult?.lifecycle?.lifecycleState, "archived");
  assert.equal(state.readResult?.artifact?.content, "Answer artifact_b");
});
