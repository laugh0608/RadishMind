import assert from "node:assert/strict";
import test from "node:test";

import {
  applicationResultArtifactResponseMatchesScope,
  listApplicationResultArtifacts,
  readApplicationResultArtifact,
  transitionApplicationResultArtifactLifecycle,
  type ApplicationResultArtifactConfig,
} from "../src/features/control-plane-read/applicationResultArtifactConsumer.ts";

const config: ApplicationResultArtifactConfig = {
  mode: "dev_application_session_http",
  baseUrl: "http://platform.test",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  subjectRef: "subject_demo_user",
};
const applicationId = "app_abcdefghijklmnop";
const sessionId = "appsess_abcdefghijklmnop";
const artifactId = "appres_abcdefghijklmnop";

test("result artifact consumer stays offline with zero requests", async () => {
  let requests = 0;
  globalThis.fetch = async () => { requests += 1; throw new Error("offline request"); };
  const offline = { ...config, mode: "offline" as const };
  assert.equal((await listApplicationResultArtifacts(offline, { applicationId, sessionId })).status, "offline");
  assert.equal((await readApplicationResultArtifact(offline, { applicationId, sessionId, artifactId })).status, "offline");
  assert.equal((await transitionApplicationResultArtifactLifecycle(offline, {
    applicationId, sessionId, artifactId, expectedLifecycleVersion: 1, targetState: "archived",
  })).status, "offline");
  assert.equal(requests, 0);
});

test("active artifact list is metadata-only and binds exact read membership scope", async () => {
  let requestURL = "";
  let headers = new Headers();
  globalThis.fetch = async (input, init) => {
    requestURL = String(input);
    headers = new Headers(init?.headers);
    return jsonResponse(listEnvelope([summaryDocument()]));
  };
  const result = await listApplicationResultArtifacts(config, {
    applicationId, sessionId, lifecycleState: "active", limit: 50,
  });
  assert.equal(result.status, "ready");
  assert.equal(result.items[0]?.artifactId, artifactId);
  assert.equal(result.items[0]?.lifecycleState, "active");
  assert.equal(Object.hasOwn(result.items[0] ?? {}, "content"), false);
  const url = new URL(requestURL);
  assert.equal(url.searchParams.get("lifecycle_state"), "active");
  assert.equal(url.searchParams.get("workspace_id"), "workspace_demo");
  assert.equal(headers.get("X-RadishMind-Dev-Read-Scopes"), "application_sessions:read");
  assert.equal(headers.get("X-RadishMind-Active-Workspace"), "workspace_demo");
  assert.equal(headers.get("X-RadishMind-Dev-Read-Membership-Permissions"), "application_sessions:read");
});

test("artifact list fails closed on content, owner, lifecycle, and unknown schema drift", async () => {
  const invalidItems = [
    { ...summaryDocument(), content: "must not enter list" },
    { ...summaryDocument(), owner_subject_ref: "subject_other" },
    { ...summaryDocument(), lifecycle_state: "archived", archived_at: null },
    { ...summaryDocument(), run_ref: { schema_version: "workflow_run_record.v4", run_id: "run_abcdefghijklmnop" } },
  ];
  for (const item of invalidItems) {
    globalThis.fetch = async () => jsonResponse(listEnvelope([item]));
    const result = await listApplicationResultArtifacts(config, { applicationId, sessionId });
    assert.equal(result.failureCode, "application_result_artifact_store_contract_mismatch");
  }
  globalThis.fetch = async () => jsonResponse({ ...listEnvelope([]), unexpected: true });
  assert.equal(
    (await listApplicationResultArtifacts(config, { applicationId, sessionId })).failureCode,
    "application_result_artifact_store_contract_mismatch",
  );
});

test("exact artifact read returns content only with matching lifecycle and provenance", async () => {
  let headers = new Headers();
  globalThis.fetch = async (_input, init) => {
    headers = new Headers(init?.headers);
    return jsonResponse(readEnvelope());
  };
  const result = await readApplicationResultArtifact(config, { applicationId, sessionId, artifactId });
  assert.equal(result.status, "ready");
  assert.equal(result.artifact?.content, "Saved result");
  assert.equal(result.artifact?.runRef.schemaVersion, "workflow_run_record.v5");
  assert.equal(result.lifecycle?.lifecycleState, "active");
  assert.equal(headers.get("X-RadishMind-Dev-Read-Membership-Permissions"), "application_sessions:read");

  const mismatched = readEnvelope();
  mismatched.artifact.content_bytes = 999;
  globalThis.fetch = async () => jsonResponse(mismatched);
  assert.equal(
    (await readApplicationResultArtifact(config, { applicationId, sessionId, artifactId })).failureCode,
    "application_result_artifact_store_contract_mismatch",
  );
});

test("archive uses independent permission, expected lifecycle CAS, and strict event", async () => {
  let body: Record<string, unknown> = {};
  let headers = new Headers();
  globalThis.fetch = async (_input, init) => {
    body = JSON.parse(String(init?.body));
    headers = new Headers(init?.headers);
    return jsonResponse(transitionEnvelope());
  };
  const result = await transitionApplicationResultArtifactLifecycle(config, {
    applicationId, sessionId, artifactId, expectedLifecycleVersion: 1, targetState: "archived",
  });
  assert.equal(result.status, "ready");
  assert.equal(result.lifecycle?.lifecycleState, "archived");
  assert.equal(result.event?.transitionKind, "archived");
  assert.deepEqual(body, {
    workspace_id: "workspace_demo",
    application_id: applicationId,
    expected_lifecycle_version: 1,
  });
  assert.equal(
    headers.get("X-RadishMind-Dev-Read-Scopes"),
    "application_sessions:read,application_result_artifacts:archive",
  );
  assert.equal(
    headers.get("X-RadishMind-Dev-Read-Membership-Permissions"),
    "application_sessions:read,application_result_artifacts:archive",
  );
});

test("stale lifecycle conflict exposes only sanitized current state and version", async () => {
  globalThis.fetch = async () => jsonResponse({
    request_id: "artifact-transition-conflict",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: applicationId,
    session_id: sessionId,
    lifecycle: null,
    event: null,
    failure_code: "application_result_artifact_lifecycle_version_conflict",
    current_lifecycle_version: 2,
    current_lifecycle_state: "archived",
    audit_ref: "audit-artifact-transition-conflict",
  }, 409);
  const result = await transitionApplicationResultArtifactLifecycle(config, {
    applicationId, sessionId, artifactId, expectedLifecycleVersion: 1, targetState: "active",
  });
  assert.equal(result.status, "version_conflict");
  assert.equal(result.currentLifecycleVersion, 2);
  assert.equal(result.currentLifecycleState, "archived");
  assert.equal(result.lifecycle, null);
  assert.equal(result.event, null);
});

test("artifact response scope rejects application, session, filter, artifact, and generation drift", () => {
  const expected = { generation: 4, applicationId, sessionId, lifecycleState: "active" as const, artifactId };
  assert.equal(applicationResultArtifactResponseMatchesScope(expected, { ...expected }), true);
  assert.equal(applicationResultArtifactResponseMatchesScope(expected, { ...expected, generation: 5 }), false);
  assert.equal(applicationResultArtifactResponseMatchesScope(expected, { ...expected, applicationId: "app_ponmlkjihgfedcba" }), false);
  assert.equal(applicationResultArtifactResponseMatchesScope(expected, { ...expected, sessionId: "appsess_ponmlkjihgfedcba" }), false);
  assert.equal(applicationResultArtifactResponseMatchesScope(expected, { ...expected, lifecycleState: "archived" }), false);
  assert.equal(applicationResultArtifactResponseMatchesScope(expected, { ...expected, artifactId: "appres_ponmlkjihgfedcba" }), false);
});

function summaryDocument() {
  return {
    schema_version: "application_result_artifact_summary.v2",
    artifact_id: artifactId,
    record_version: 1,
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: applicationId,
    owner_subject_ref: "subject_demo_user",
    session_id: sessionId,
    turn_id: "appturn_abcdefghijklmnop",
    client_turn_key: "web_turn_001",
    execution_profile: "workflow_definition_executor_v1",
    run_ref: { schema_version: "workflow_run_record.v5", run_id: "run_abcdefghijklmnop" },
    content_type: "text/markdown",
    content_bytes: 12,
    content_digest: `sha256:${"a".repeat(64)}`,
    created_at: "2026-08-17T03:00:00Z",
    lifecycle_state: "active",
    lifecycle_version: 1,
    archived_at: null,
    lifecycle_updated_at: "2026-08-17T03:00:00Z",
  };
}

function artifactDocument() {
  const summary = summaryDocument();
  return {
    schema_version: "application_result_artifact.v1",
    artifact_id: summary.artifact_id,
    record_version: 1,
    tenant_ref: summary.tenant_ref,
    workspace_id: summary.workspace_id,
    application_id: summary.application_id,
    owner_subject_ref: summary.owner_subject_ref,
    session_id: summary.session_id,
    turn_id: summary.turn_id,
    client_turn_key: summary.client_turn_key,
    execution_profile: summary.execution_profile,
    run_ref: summary.run_ref,
    content_type: summary.content_type,
    content: "Saved result",
    content_bytes: new TextEncoder().encode("Saved result").length,
    content_digest: summary.content_digest,
    created_at: summary.created_at,
    created_by_actor_ref: "subject_demo_user",
    request_id: "artifact-turn-request",
    audit_ref: "audit-artifact-turn-request",
  };
}

function lifecycleDocument(state: "active" | "archived" = "active", version = 1) {
  return {
    schema_version: "application_result_artifact_lifecycle.v1",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: applicationId,
    owner_subject_ref: "subject_demo_user",
    artifact_id: artifactId,
    lifecycle_state: state,
    lifecycle_version: version,
    archived_at: state === "archived" ? "2026-08-17T03:05:00Z" : null,
    updated_at: state === "archived" ? "2026-08-17T03:05:00Z" : "2026-08-17T03:00:00Z",
    updated_by_actor_ref: "subject_demo_user",
    request_id: "artifact-lifecycle-request",
    audit_ref: "audit-artifact-lifecycle-request",
  };
}

function listEnvelope(items: unknown[]) {
  return {
    request_id: "artifact-list-request",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: applicationId,
    session_id: sessionId,
    items,
    next_cursor: null,
    failure_code: null,
    audit_ref: "audit-artifact-list-request",
  };
}

function readEnvelope() {
  return {
    request_id: "artifact-read-request",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: applicationId,
    session_id: sessionId,
    artifact: artifactDocument(),
    lifecycle: lifecycleDocument(),
    failure_code: null,
    audit_ref: "audit-artifact-read-request",
  };
}

function transitionEnvelope() {
  return {
    request_id: "artifact-transition-request",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: applicationId,
    session_id: sessionId,
    lifecycle: lifecycleDocument("archived", 2),
    event: {
      schema_version: "application_result_artifact_lifecycle_event.v1",
      tenant_ref: "tenant_demo",
      workspace_id: "workspace_demo",
      application_id: applicationId,
      owner_subject_ref: "subject_demo_user",
      artifact_id: artifactId,
      lifecycle_version: 2,
      from_state: "active",
      to_state: "archived",
      transition_kind: "archived",
      occurred_at: "2026-08-17T03:05:00Z",
      actor_ref: "subject_demo_user",
      request_id: "artifact-transition-request",
      audit_ref: "audit-artifact-transition-request",
    },
    failure_code: null,
    current_lifecycle_version: 2,
    current_lifecycle_state: "archived",
    audit_ref: "audit-artifact-transition-request",
  };
}

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}
