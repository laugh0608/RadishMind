import assert from "node:assert/strict";
import test from "node:test";
import { readPromptPlanReference } from "../src/features/control-plane-read/promptEvaluationPlanReference.ts";
import type { PromptApplicationRuntimeConfig } from "../src/features/control-plane-read/promptApplicationRuntimeConsumer.ts";

const config: PromptApplicationRuntimeConfig = {
  mode: "dev_prompt_application_http", baseUrl: "http://platform.test", tenantRef: "tenant_scoped", workspaceId: "workspace_scoped", subjectRef: "subject_scoped", authMode: "dev_headers",
};
const applicationId = "app_aaaaaaaaaaaaaaaa";
const templateId = "ptpl_aaaaaaaaaaaaaaaa";
const digest = `sha256:${"a".repeat(64)}`;
const read = (isCurrent = () => true) => readPromptPlanReference(config, config, applicationId, isCurrent);
const response = (body: unknown) => new Response(JSON.stringify(body), { headers: { "Content-Type": "application/json" } });

test("Prompt reference reads assignment, exact source, then verifies assignment again using current scope", async () => {
  const requests: Array<{ url: string; init?: RequestInit }> = [];
  globalThis.fetch = async (url, init) => {
    requests.push({ url: String(url), init });
    return response(String(url).includes("/versions/") ? source() : assignment());
  };
  const result = await read();
  assert.equal(result.status, "ready");
  if (result.status !== "ready") return;
  assert.equal(result.reference.source.templateVersion, 1);
  assert.equal(result.reference.assignmentVersion, 4);
  assert.equal(result.reference.source.workspaceId, config.workspaceId);
  assert.equal(requests.length, 3);
  assert.ok(requests[0]!.url.endsWith(`/prompt-runtime-assignment?workspace_id=${config.workspaceId}`));
  assert.ok(requests[1]!.url.includes(`${templateId}/versions/1?workspace_id=${config.workspaceId}`));
  assert.equal(requests[0]!.url, requests[2]!.url);
  for (const request of requests) {
    const headers = new Headers(request.init?.headers);
    assert.equal(headers.get("X-RadishMind-Dev-Read-Tenant"), config.tenantRef);
    assert.equal(headers.get("X-RadishMind-Dev-Read-Subject"), config.subjectRef);
    assert.equal(request.init?.body, undefined);
    assert.equal(request.init?.cache, "no-store");
    assert.ok(!request.url.includes("/events"));
  }
  assert.equal(new Headers(requests[1]!.init?.headers).get("X-RadishMind-Dev-Read-Scopes"), "prompt_application_templates:read_source");
});

test("offline and mismatched reference scopes cannot start requests", async () => {
  globalThis.fetch = async () => { throw new Error("unexpected request"); };
  assert.deepEqual(await read(() => false), { status: "cancelled" });
  assert.deepEqual(await readPromptPlanReference({ ...config, mode: "offline" }, config, applicationId, () => true), { status: "unavailable", code: "prompt_plan_reference_offline" });
  for (const different of [{ tenantRef: "other" }, { workspaceId: "other" }, { subjectRef: "other" }, { authMode: "local_session_dev_test" as const }]) {
    assert.deepEqual(await readPromptPlanReference(config, { ...config, ...different }, applicationId, () => true), { status: "unavailable", code: "prompt_plan_reference_scope_mismatch" });
  }
});

test("reference reads stop after invalidation at every await boundary", async () => {
  for (const cancelAt of [1, 2, 3]) {
    let count = 0, active = true;
    globalThis.fetch = async url => {
      count += 1;
      if (count === cancelAt) active = false;
      return response(String(url).includes("/versions/") ? source() : assignment());
    };
    assert.deepEqual(await read(() => active), { status: "cancelled" });
    assert.equal(count, cancelAt);
  }
});

test("missing, revoked, denied, malformed and drifting authorities never supply a usable reference", async () => {
  for (const failure of ["prompt_runtime_assignment_not_found", "prompt_runtime_scope_denied"]) {
    globalThis.fetch = async () => response({ ...assignment(), assignment: null, failure_code: failure });
    assert.deepEqual(await read(), { status: "unavailable", code: failure });
  }
  globalThis.fetch = async () => response({ ...assignment(), assignment: { ...assignment().assignment, state: "revoked", revoked_at: "2026-10-15T10:00:00Z" }, current_state: "revoked" });
  assert.equal((await read()).status, "unavailable");
  globalThis.fetch = async () => { throw new Error("offline"); };
  assert.equal((await read()).status, "unavailable");
  for (const change of ["version", "state"]) {
    let calls = 0;
    globalThis.fetch = async url => {
      calls += 1;
      if (String(url).includes("/versions/")) return response(source());
      const body = assignment();
      if (calls === 3) {
        if (change === "version") { body.assignment.assignment_version += 1; body.current_assignment_version += 1; }
        else { body.assignment.state = "revoked"; body.assignment.revoked_at = "2026-10-15T10:00:00Z"; body.current_state = "revoked"; }
      }
      return response(body);
    };
    assert.deepEqual(await read(), { status: "unavailable", code: "prompt_plan_reference_changed" });
  }
});

test("reference source must match exact template identity, digest and tenant and preserve numeric defaults", async () => {
  for (const difference of [{ template_id: "ptpl_bbbbbbbbbbbbbbbb" }, { template_version: 2 }, { template_digest: `sha256:${"b".repeat(64)}` }, { tenant_ref: "tenant_other" }]) {
    globalThis.fetch = async url => response(String(url).includes("/versions/") ? { ...source(), version: { ...source().version, ...difference } } : assignment());
    assert.equal((await read()).status, "unavailable");
  }
  globalThis.fetch = async url => response(String(url).includes("/versions/") ? { ...source(), version: null, failure_code: "prompt_template_scope_denied" } : assignment());
  assert.deepEqual(await read(), { status: "unavailable", code: "prompt_template_scope_denied" });
  globalThis.fetch = async url => String(url).includes("/versions/")
    ? new Response(JSON.stringify(source()).replace('"default_value":0.1', '"default_value":0.10000000000000001'))
    : response(assignment());
  assert.equal((await read()).status, "unavailable");
});

function assignment() {
  return {
    request_id: "ref-read", tenant_ref: config.tenantRef, workspace_id: config.workspaceId, application_id: applicationId,
    assignment: {
      schema_version: "prompt_application_runtime_assignment.v1", assignment_id: "ptra_aaaaaaaaaaaaaaaa", tenant_ref: config.tenantRef,
      workspace_id: config.workspaceId, application_id: applicationId, owner_subject_ref: config.subjectRef, assignment_version: 4, state: "active",
      candidate_id: "candidate-1", candidate_review_version: 1, draft_id: "draft-1", draft_version: 1, draft_digest: digest,
      prompt_template_ref: { template_id: templateId, template_version: 1, template_digest: digest }, assignment_digest: digest,
      activated_at: "2026-10-15T09:00:00Z", updated_at: "2026-10-15T09:00:00Z", revoked_at: null as string | null,
      activated_by_actor_ref: config.subjectRef, updated_by_actor_ref: config.subjectRef, request_id: "ref-read", audit_ref: "audit-ref-read",
    },
    events: [], failure_code: null, current_assignment_version: 4, current_state: "active", audit_ref: "audit-ref-read",
  };
}

function source() {
  return {
    request_id: "ref-source", workspace_id: config.workspaceId, application_id: applicationId, draft: null,
    version: {
      schema_version: "prompt_application_template_version.v1", template_id: templateId, template_version: 1, source_draft_version: 1,
      tenant_ref: config.tenantRef, workspace_id: config.workspaceId, application_id: applicationId, owner_subject_ref: config.subjectRef,
      template_name: "Reference", description: "", messages: [{ role: "user", content: "{{question}} {{weight}}" }],
      variables: [{ name: "question", type: "string", required: true, description: "" }, { name: "weight", type: "number", required: false, description: "", default_value: 0.1 }],
      output_contract: { kind: "text", allow_empty: false, max_bytes: 1024 }, template_digest: digest,
      created_at: "2026-10-15T09:00:00Z", created_by_actor_ref: config.subjectRef, request_id: "ref-source", audit_ref: "audit-ref-source",
    },
    failure_code: null, current_draft_version: 1, current_template_version: 1, validation_summary: { state: "valid", is_valid: true, findings: [] }, audit_ref: "audit-ref-source",
  };
}
