import assert from "node:assert/strict";
import test from "node:test";

import {
  archiveApplicationCatalogRecord,
  createApplicationCatalogRecord,
  listApplicationCatalogRecords,
  readApplicationCatalogRecord,
  updateApplicationCatalogRecord,
  validateApplicationCatalogMutableFields,
  type ApplicationCatalogConfig,
} from "../src/features/control-plane-read/applicationCatalogConsumer.ts";
import { buildWorkspaceApplicationsViewModel } from "../src/features/control-plane-read/workspaceApplications.ts";

const config: ApplicationCatalogConfig = {
  mode: "dev_application_catalog_http",
  baseUrl: "http://platform.test",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  subjectRef: "subject_demo_user",
  authMode: "dev_headers",
};

test("application catalog offline mode sends zero requests", async () => {
  const originalFetch = globalThis.fetch;
  let calls = 0;
  globalThis.fetch = async () => { calls += 1; throw new Error("offline mode must not fetch"); };
  const offline = { ...config, mode: "offline" as const };
  try {
    await listApplicationCatalogRecords(offline, "active");
    await createApplicationCatalogRecord(offline, fields());
    await readApplicationCatalogRecord(offline, "app_aaaaaaaaaaaaaaaa");
    await updateApplicationCatalogRecord(offline, "app_aaaaaaaaaaaaaaaa", 1, fields());
    await archiveApplicationCatalogRecord(offline, "app_aaaaaaaaaaaaaaaa", 1);
    assert.equal(calls, 0);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("create sends only mutable fields with exact write scope", async () => {
  const originalFetch = globalThis.fetch;
  let captured: { url: string; method: string; headers: Headers; body: unknown } | null = null;
  globalThis.fetch = async (input, init) => {
    captured = {
      url: String(input),
      method: String(init?.method),
      headers: new Headers(init?.headers),
      body: JSON.parse(String(init?.body)),
    };
    return jsonResponse(operationEnvelope());
  };
  try {
    const result = await createApplicationCatalogRecord(config, fields());
    assert.equal(result.status, "created");
    assert.equal(result.record?.recordVersion, 1);
    assert.equal(captured?.url, "http://platform.test/v1/user-workspace/applications");
    assert.equal(captured?.method, "POST");
    assert.equal(captured?.headers.get("X-RadishMind-Dev-Read-Scopes"), "applications:write,applications:read");
    assert.deepEqual(captured?.body, {
      workspace_id: "workspace_demo",
      display_name: "Catalog App",
      description: "Created by a strict Web consumer.",
      application_kind: "agent",
    });
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("list strictly accepts scoped active and archived projections", async () => {
  const originalFetch = globalThis.fetch;
  const requested: string[] = [];
  globalThis.fetch = async (input) => {
    requested.push(String(input));
    const archived = String(input).includes("lifecycle_state=archived");
    return jsonResponse(listEnvelope(archived ? "archived" : "active"));
  };
  try {
    const active = await listApplicationCatalogRecords(config, "active");
    const archived = await listApplicationCatalogRecords(config, "archived");
    assert.equal(active.status, "ready");
    assert.equal(active.records[0]?.lifecycleState, "active");
    assert.equal(archived.records[0]?.lifecycleState, "archived");
    assert.equal(requested.every((url) => url.includes("workspace_id=workspace_demo")), true);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("update preserves a server CAS conflict and archive uses the archive scope", async () => {
  const originalFetch = globalThis.fetch;
  let calls = 0;
  globalThis.fetch = async (_input, init) => {
    calls += 1;
    if (calls === 1) {
      return jsonResponse({
        ...operationEnvelope(),
        record: null,
        failure_code: "application_catalog_version_conflict",
        current_record_version: 3,
        current_lifecycle_state: "active",
      });
    }
    const headers = new Headers(init?.headers);
    assert.equal(headers.get("X-RadishMind-Dev-Read-Scopes"), "applications:archive,applications:read");
    return jsonResponse(operationEnvelope({ lifecycle: "archived", version: 4 }));
  };
  try {
    const conflict = await updateApplicationCatalogRecord(config, "app_aaaaaaaaaaaaaaaa", 2, fields("Local edit"));
    assert.equal(conflict.status, "version_conflict");
    assert.equal(conflict.currentRecordVersion, 3);
    assert.equal(conflict.record, null);
    const archived = await archiveApplicationCatalogRecord(config, "app_aaaaaaaaaaaaaaaa", 3);
    assert.equal(archived.status, "archived");
    assert.equal(archived.record?.archivedAt, "2026-07-13T16:05:00Z");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("strict response validation rejects unknown fields, scope drift, and secret material", async () => {
  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async () => jsonResponse({ ...operationEnvelope(), debug: "unexpected" });
    assert.equal((await readApplicationCatalogRecord(config, "app_aaaaaaaaaaaaaaaa")).failureCode, "application_catalog_store_unavailable");

    const drift = operationEnvelope();
    drift.record.workspace_id = "workspace_other";
    globalThis.fetch = async () => jsonResponse(drift);
    assert.equal((await readApplicationCatalogRecord(config, "app_aaaaaaaaaaaaaaaa")).failureCode, "application_catalog_store_unavailable");

    const secret = operationEnvelope();
    secret.record.description = "Authorization: Bearer hidden";
    globalThis.fetch = async () => jsonResponse(secret);
    assert.equal((await readApplicationCatalogRecord(config, "app_aaaaaaaaaaaaaaaa")).failureCode, "application_catalog_store_unavailable");

    const listUnknown = { ...listEnvelope("active"), raw_request: "forbidden" };
    globalThis.fetch = async () => jsonResponse(listUnknown);
    assert.equal((await listApplicationCatalogRecords(config, "active")).failureCode, "application_catalog_store_unavailable");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("mutable field validation rejects invalid lengths and sensitive material", () => {
  assert.equal(validateApplicationCatalogMutableFields(fields()), "");
  assert.equal(validateApplicationCatalogMutableFields(fields("x")), "application_catalog_payload_invalid");
  assert.equal(validateApplicationCatalogMutableFields({ ...fields(), description: "postgresql://secret" }), "application_catalog_secret_material_forbidden");
});

test("an enabled empty catalog remains the only workspace application truth source", () => {
  const view = buildWorkspaceApplicationsViewModel(undefined, {
    records: [],
    summary: "The authoritative catalog is empty.",
  });

  assert.equal(view.applications.length, 0);
  assert.equal(view.metrics[0]?.value, "0");
  assert.equal(view.requestId, "request_application_catalog_snapshot");
  assert.equal(view.auditRef, "audit_application_catalog_snapshot");
});

function fields(displayName = "Catalog App") {
  return { displayName, description: "Created by a strict Web consumer.", applicationKind: "agent" as const };
}

function operationEnvelope(options: { lifecycle?: "active" | "archived"; version?: number } = {}) {
  const lifecycle = options.lifecycle ?? "active";
  const version = options.version ?? 1;
  return {
    request_id: "application-catalog-request-0001",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    record: {
      schema_version: "application_catalog_record.v1",
      application_id: "app_aaaaaaaaaaaaaaaa",
      tenant_ref: "tenant_demo",
      workspace_id: "workspace_demo",
      owner_subject_ref: "subject_demo_user",
      display_name: "Catalog App",
      description: "Created by a strict Web consumer.",
      application_kind: "agent",
      lifecycle_state: lifecycle,
      record_version: version,
      created_at: "2026-07-13T16:00:00Z",
      updated_at: lifecycle === "archived" ? "2026-07-13T16:05:00Z" : "2026-07-13T16:00:00Z",
      archived_at: lifecycle === "archived" ? "2026-07-13T16:05:00Z" : null,
      created_by_actor_ref: "subject_demo_user",
      updated_by_actor_ref: "subject_demo_user",
      request_id: "application-catalog-request-0001",
      audit_ref: "audit-application-catalog-request-0001",
    },
    failure_code: null,
    current_record_version: version,
    current_lifecycle_state: lifecycle,
    audit_ref: "audit-application-catalog-request-0001",
  };
}

function listEnvelope(lifecycle: "active" | "archived") {
  return {
    request_id: `application-catalog-list-${lifecycle}`,
    tenant_ref: "tenant_demo",
    items: [{
      application_ref: "app_aaaaaaaaaaaaaaaa",
      tenant_ref: "tenant_demo",
      workspace_id: "workspace_demo",
      application_kind: "agent",
      display_name: "Catalog App",
      description: "Created by a strict Web consumer.",
      owner_subject_ref: "subject_demo_user",
      latest_workflow_definition_ref: "",
      last_run_status: "not_available",
      lifecycle_state: lifecycle,
      record_version: lifecycle === "active" ? 1 : 4,
      created_at: "2026-07-13T16:00:00Z",
      updated_at: lifecycle === "active" ? "2026-07-13T16:00:00Z" : "2026-07-13T16:05:00Z",
      archived_at: lifecycle === "active" ? null : "2026-07-13T16:05:00Z",
    }],
    next_cursor: null,
    failure_code: null,
    audit_ref: `audit-application-catalog-list-${lifecycle}`,
  };
}

function jsonResponse(document: unknown): Response {
  return new Response(JSON.stringify(document), { status: 200, headers: { "Content-Type": "application/json" } });
}
