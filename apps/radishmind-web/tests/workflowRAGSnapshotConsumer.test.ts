import assert from "node:assert/strict";
import test from "node:test";

import {
  archiveWorkflowRAGSnapshot,
  createWorkflowRAGSnapshot,
  listWorkflowRAGSnapshots,
  readWorkflowRAGSnapshot,
  validateWorkflowRAGSnapshotWriteInput,
  versionWorkflowRAGSnapshot,
  type WorkflowRAGSnapshotConfig,
} from "../src/features/control-plane-read/workflowRAGSnapshotConsumer.ts";

const snapshotId = "rags_abcdefghijklmnop";
const config: WorkflowRAGSnapshotConfig = {
  mode: "dev_workflow_rag_http",
  baseUrl: "http://platform.test",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  subjectRef: "subject_demo_user",
  authMode: "dev_headers",
  scopes: new Set([
    "workflow_rag_snapshots:read",
    "workflow_rag_snapshots:write",
    "workflow_rag_snapshots:archive",
  ]),
};

test("offline and missing scopes send zero knowledge snapshot requests", async () => {
  const originalFetch = globalThis.fetch;
  let calls = 0;
  globalThis.fetch = async () => { calls += 1; throw new Error("boundary must not fetch"); };
  try {
    const offline = { ...config, mode: "offline" as const };
    assert.equal((await listWorkflowRAGSnapshots(offline, "app_flow_copilot", "active")).status, "offline");
    assert.equal((await createWorkflowRAGSnapshot(offline, "app_flow_copilot", writeInput())).status, "offline");
    assert.equal((await readWorkflowRAGSnapshot(offline, "app_flow_copilot", snapshotId, 1)).status, "offline");
    assert.equal((await versionWorkflowRAGSnapshot(offline, "app_flow_copilot", snapshotId, 1, writeInput())).status, "offline");
    assert.equal((await archiveWorkflowRAGSnapshot(offline, "app_flow_copilot", snapshotId, 1)).status, "offline");

    const noScopes = { ...config, scopes: new Set<never>() };
    assert.equal((await listWorkflowRAGSnapshots(noScopes, "app_flow_copilot", "active")).status, "scope_denied");
    assert.equal((await createWorkflowRAGSnapshot(noScopes, "app_flow_copilot", writeInput())).status, "scope_denied");
    assert.equal((await readWorkflowRAGSnapshot(noScopes, "app_flow_copilot", snapshotId, 1)).status, "scope_denied");
    assert.equal((await versionWorkflowRAGSnapshot(noScopes, "app_flow_copilot", snapshotId, 1, writeInput())).status, "scope_denied");
    assert.equal((await archiveWorkflowRAGSnapshot(noScopes, "app_flow_copilot", snapshotId, 1)).status, "scope_denied");
    assert.equal(calls, 0);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("create sends exact application scope and sanitized fragment input", async () => {
  const originalFetch = globalThis.fetch;
  let captured: { url: string; method: string; headers: Headers; body: unknown } | null = null;
  globalThis.fetch = async (input, init) => {
    captured = { url: String(input), method: String(init?.method), headers: new Headers(init?.headers), body: JSON.parse(String(init?.body)) };
    return jsonResponse(operationEnvelope());
  };
  try {
    const result = await createWorkflowRAGSnapshot(config, "app_flow_copilot", writeInput());
    assert.equal(result.status, "created");
    assert.equal(result.record?.ragRef, "workflow.rag.product_docs.v1");
    assert.equal(captured?.url, "http://platform.test/v1/user-workspace/workflow-retrieval-snapshots");
    assert.equal(captured?.method, "POST");
    assert.equal(captured?.headers.get("X-RadishMind-Dev-Read-Scopes"), "workflow_rag_snapshots:write");
    assert.equal(captured?.headers.get("X-RadishMind-Dev-Application"), "app_flow_copilot");
    assert.deepEqual(captured?.body, {
      workspace_id: "workspace_demo",
      application_id: "app_flow_copilot",
      snapshot_key: "product_docs",
      display_name: "产品文档",
      content_classification: "workspace_internal",
      fragments: [{ fragment_ref: "overview", source_type: "manual", source_ref: "manual_docs", page_slug: "overview", title: "产品概览", is_official: true, content: "RadishMind 提供结构化建议。" }],
    });
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("list accepts metadata only, supports cursor, and rejects content or unknown fields", async () => {
  const originalFetch = globalThis.fetch;
  const requested: string[] = [];
  try {
    globalThis.fetch = async (input) => { requested.push(String(input)); return jsonResponse(listEnvelope()); };
    const listed = await listWorkflowRAGSnapshots(config, "app_flow_copilot", "active", "previous_docs");
    assert.equal(listed.status, "ready");
    assert.equal(listed.records[0]?.latestVersion, 1);
    assert.match(requested[0]!, /cursor=previous_docs/u);
    assert.equal(JSON.stringify(listed).includes("结构化建议"), false);

    const withContent = listEnvelope() as ReturnType<typeof listEnvelope> & { content?: string };
    withContent.content = "正文不得出现在列表";
    globalThis.fetch = async () => jsonResponse(withContent);
    assert.equal((await listWorkflowRAGSnapshots(config, "app_flow_copilot", "active")).failureCode, "workflow_rag_store_unavailable");

    const withUnknownItem = listEnvelope();
    Object.assign(withUnknownItem.items[0]!, { debug: "unexpected" });
    globalThis.fetch = async () => jsonResponse(withUnknownItem);
    assert.equal((await listWorkflowRAGSnapshots(config, "app_flow_copilot", "active")).failureCode, "workflow_rag_store_unavailable");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("detail exposes exact authorized content but rejects scope drift and secret material", async () => {
  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async () => jsonResponse(operationEnvelope());
    const detail = await readWorkflowRAGSnapshot(config, "app_flow_copilot", snapshotId, 1);
    assert.equal(detail.status, "loaded");
    assert.equal(detail.record?.fragments[0]?.content, "RadishMind 提供结构化建议。");

    const drift = operationEnvelope();
    drift.record!.workspace_id = "workspace_other";
    globalThis.fetch = async () => jsonResponse(drift);
    assert.equal((await readWorkflowRAGSnapshot(config, "app_flow_copilot", snapshotId, 1)).failureCode, "workflow_rag_store_unavailable");

    const secret = operationEnvelope();
    secret.record!.fragments[0]!.content = "Authorization: Bearer hidden";
    secret.record!.fragments[0]!.content_bytes = new TextEncoder().encode(secret.record!.fragments[0]!.content).length;
    globalThis.fetch = async () => jsonResponse(secret);
    const rejected = await readWorkflowRAGSnapshot(config, "app_flow_copilot", snapshotId, 1);
    assert.equal(rejected.failureCode, "workflow_rag_store_unavailable");
    assert.equal(JSON.stringify(rejected).includes("hidden"), false);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("version uses full replacement CAS without snapshot_key and archive uses its own scope", async () => {
  const originalFetch = globalThis.fetch;
  const requests: Array<{ url: string; headers: Headers; body: Record<string, unknown> }> = [];
  globalThis.fetch = async (input, init) => {
    requests.push({ url: String(input), headers: new Headers(init?.headers), body: JSON.parse(String(init?.body)) });
    return jsonResponse(requests.length === 1 ? operationEnvelope({ version: 2 }) : operationEnvelope({ version: 2, lifecycle: "archived" }));
  };
  try {
    const versioned = await versionWorkflowRAGSnapshot(config, "app_flow_copilot", snapshotId, 1, writeInput());
    assert.equal(versioned.status, "versioned");
    assert.equal(requests[0]?.body.expected_latest_version, 1);
    assert.equal("snapshot_key" in requests[0]!.body, false);
    assert.equal(requests[0]?.headers.get("X-RadishMind-Dev-Read-Scopes"), "workflow_rag_snapshots:write");

    const archived = await archiveWorkflowRAGSnapshot(config, "app_flow_copilot", snapshotId, 2);
    assert.equal(archived.status, "archived");
    assert.deepEqual(requests[1]?.body, { workspace_id: "workspace_demo", application_id: "app_flow_copilot", expected_latest_version: 2 });
    assert.equal(requests[1]?.headers.get("X-RadishMind-Dev-Read-Scopes"), "workflow_rag_snapshots:archive");
    assert.equal(requests.some((request) => request.url.includes("retrieval-executions")), false);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("consumer preserves version conflict metadata and validates local budgets", async () => {
  const originalFetch = globalThis.fetch;
  const conflict = operationEnvelope();
  conflict.record = null;
  conflict.failure_code = "workflow_rag_snapshot_version_conflict";
  conflict.current_latest_version = 3;
  globalThis.fetch = async () => jsonResponse(conflict);
  try {
    const result = await versionWorkflowRAGSnapshot(config, "app_flow_copilot", snapshotId, 1, writeInput());
    assert.equal(result.status, "version_conflict");
    assert.equal(result.currentLatestVersion, 3);
  } finally {
    globalThis.fetch = originalFetch;
  }

  assert.equal(validateWorkflowRAGSnapshotWriteInput(writeInput()), "");
  assert.equal(validateWorkflowRAGSnapshotWriteInput({ ...writeInput(), fragments: [{ ...writeInput().fragments[0]!, content: "api_key=secret" }] }), "workflow_rag_secret_material_forbidden");
  assert.equal(validateWorkflowRAGSnapshotWriteInput({ ...writeInput(), fragments: [{ ...writeInput().fragments[0]!, content: "x".repeat(8193) }] }), "workflow_rag_fragment_invalid");
});

function writeInput() {
  return {
    snapshotKey: "product_docs",
    displayName: "产品文档",
    contentClassification: "workspace_internal" as const,
    fragments: [{ fragmentRef: "overview", sourceType: "manual" as const, sourceRef: "manual_docs", pageSlug: "overview", title: "产品概览", isOfficial: true, content: "RadishMind 提供结构化建议。" }],
  };
}

function operationEnvelope(options: { version?: number; lifecycle?: "active" | "archived" } = {}) {
  const version = options.version ?? 1;
  const lifecycle = options.lifecycle ?? "active";
  return {
    request_id: "workflow-rag-request-0001",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: "app_flow_copilot",
    record: snapshotRecord(version, lifecycle),
    failure_code: null as string | null,
    current_latest_version: version,
    current_lifecycle_state: lifecycle,
    audit_ref: "audit-workflow-rag-request-0001",
  };
}

function snapshotRecord(version: number, lifecycle: "active" | "archived") {
  const content = "RadishMind 提供结构化建议。";
  return {
    schema_version: "workflow_rag_snapshot.v1",
    snapshot_id: snapshotId,
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: "app_flow_copilot",
    snapshot_key: "product_docs",
    rag_ref: `workflow.rag.product_docs.v${version}`,
    snapshot_version: version,
    display_name: "产品文档",
    lifecycle_state: lifecycle,
    content_classification: "workspace_internal",
    profile_ref: "workflow.rag.lexical-ngram-dev.v1",
    fragment_count: 1,
    total_content_bytes: new TextEncoder().encode(content).length,
    snapshot_digest: `sha256:${"a".repeat(64)}`,
    created_at: "2026-07-17T08:00:00Z",
    created_by_actor_ref: "subject_demo_user",
    request_id: "workflow-rag-request-0001",
    audit_ref: "audit-workflow-rag-request-0001",
    fragments: [{
      schema_version: "workflow_rag_fragment.v1",
      fragment_ref: "overview",
      source_type: "manual",
      source_ref: "manual_docs",
      page_slug: "overview",
      title: "产品概览",
      is_official: true,
      content,
      content_classification: "workspace_internal",
      content_bytes: new TextEncoder().encode(content).length,
      content_digest: `sha256:${"b".repeat(64)}`,
    }],
  };
}

function listEnvelope() {
  return {
    request_id: "workflow-rag-list-0001",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    application_id: "app_flow_copilot",
    items: [{
      snapshot_id: snapshotId,
      tenant_ref: "tenant_demo",
      workspace_id: "workspace_demo",
      application_id: "app_flow_copilot",
      snapshot_key: "product_docs",
      display_name: "产品文档",
      lifecycle_state: "active",
      latest_version: 1,
      latest_rag_ref: "workflow.rag.product_docs.v1",
      latest_digest: `sha256:${"a".repeat(64)}`,
      fragment_count: 1,
      total_content_bytes: 36,
      created_at: "2026-07-17T08:00:00Z",
      updated_at: "2026-07-17T08:00:00Z",
      archived_at: null,
    }],
    next_cursor: "product_docs",
    failure_code: null,
    audit_ref: "audit-workflow-rag-list-0001",
  };
}

function jsonResponse(document: unknown): Response {
  return new Response(JSON.stringify(document), { status: 200, headers: { "Content-Type": "application/json" } });
}
