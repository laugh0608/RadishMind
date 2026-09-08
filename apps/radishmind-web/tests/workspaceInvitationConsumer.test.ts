import assert from "node:assert/strict";
import { randomBytes } from "node:crypto";
import test, { type TestContext } from "node:test";

import {
  WorkspaceInvitationError,
  claimWorkspaceInvitation,
  createWorkspaceInvitation,
  isInvitationCode,
  listWorkspaceInvitations,
  previewWorkspaceInvitation,
  revokeWorkspaceInvitation,
  workspaceInvitationFailure,
  type InvitationAdminConfig,
} from "../src/features/local-identity/workspaceInvitationConsumer.ts";

const config: InvitationAdminConfig = {
  mode: "local_identity_dev", baseUrl: "http://platform.test", tenantRef: "tenant_demo", workspaceId: "workspace_demo",
};
const role = {
  roleKey: "workspace_builder", catalogVersion: "local_identity_builtin_roles_v1", definitionDigest: `sha256:${"a".repeat(64)}`,
  displayName: "Workspace builder", summary: "Create applications and workflows.", permissionGrants: ["application:write"], canManageLocalIdentity: false,
};
const userId = `usr_${"b".repeat(32)}`;
const createdAt = "2026-09-08T02:00:00Z";
const expiresAt = "2026-09-09T02:00:00Z";
const changedAt = "2026-09-08T02:01:00Z";
const asOf = "2026-09-08T02:02:00Z";

function invitationId() { return `wsi_${"a".repeat(32)}`; }
function invitationCode() { return `rmi_${invitationId().slice(4)}.${randomBytes(32).toString("base64url")}`; }
function record(state: "pending" | "claimed" | "revoked" = "pending"): Record<string, unknown> {
  return {
    schema_version: "workspace_invitation.v1", invitation_id: invitationId(), record_version: state === "pending" ? 1 : 2,
    tenant_ref: config.tenantRef, workspace_id: config.workspaceId, role_key: role.roleKey,
    role_catalog_version: role.catalogVersion, role_definition_digest: role.definitionDigest, ttl_policy: "24h",
    lifecycle_state: state, effective_state: state, expires_at: expiresAt, created_at: createdAt,
    updated_at: state === "pending" ? createdAt : changedAt, created_by_actor_ref: `user:${userId}`,
    created_request_ref: "request:create", created_audit_ref: "audit:create", updated_request_ref: "request:update", updated_audit_ref: "audit:update",
    ...(state === "claimed" ? { claimed_at: changedAt, claimed_by_user_id: userId, membership_id: `mbr_${"c".repeat(32)}`, assignment_id: `rla_${"d".repeat(32)}` } : {}),
    ...(state === "revoked" ? { revoked_at: changedAt, revoked_by_actor_ref: `user:${userId}` } : {}),
  };
}
function page() {
  return { schema_version: "workspace_invitation_page.v1", request_id: "request:list", tenant_ref: config.tenantRef,
    workspace_id: config.workspaceId, as_of: asOf, invitations: [record()] };
}
function preview() {
  return { schema_version: "workspace_invitation_preview.v1", request_id: "request:preview", invitation_id: invitationId(), record_version: 1,
    tenant_ref: config.tenantRef, workspace_id: config.workspaceId, effective_state: "pending", expires_at: expiresAt,
    role: { role_key: role.roleKey, catalog_version: role.catalogVersion, definition_digest: role.definitionDigest, display_name: role.displayName, summary: role.summary } };
}
function claim() {
  return { schema_version: "workspace_invitation_mutation.v1", request_id: "request:claim", invitation: record("claimed"),
    membership: { membership_id: `mbr_${"c".repeat(32)}`, user_id: userId, tenant_ref: config.tenantRef, workspace_id: config.workspaceId, record_version: 1 },
    role_assignment: { assignment_id: `rla_${"d".repeat(32)}`, user_id: userId, tenant_ref: config.tenantRef,
      workspace_id: config.workspaceId, record_version: 1, role_key: role.roleKey } };
}
function json(value: unknown, status = 200) { return new Response(JSON.stringify(value), { status, headers: { "Content-Type": "application/json; charset=utf-8" } }); }
function installCSRF(t: TestContext) {
  const descriptor = Object.getOwnPropertyDescriptor(globalThis, "document");
  Object.defineProperty(globalThis, "document", { configurable: true, value: { cookie: "radishmind_csrf_dev=csrf-proof" } });
  t.after(() => { if (descriptor) Object.defineProperty(globalThis, "document", descriptor); else Reflect.deleteProperty(globalThis, "document"); });
}
function rejectedResponse(error: unknown): boolean { return error instanceof WorkspaceInvitationError && error.code === "workspace_invitation_response_invalid"; }

test("invitation consumer connects five cookie routes with scope, explicit confirmation, CAS and exact result relationships", async (t) => {
  installCSRF(t);
  const code = invitationCode();
  const requests: Array<{ url: URL; options?: RequestInit; body?: Record<string, unknown> }> = [];
  t.mock.method(globalThis, "fetch", async (input: RequestInfo | URL, options?: RequestInit) => {
    const url = new URL(String(input));
    const body = options?.body ? JSON.parse(String(options.body)) : undefined;
    requests.push({ url, options, body });
    if (url.pathname.endsWith("/preview")) return json(preview());
    if (url.pathname.endsWith("/claim")) return json(claim());
    if (url.pathname.endsWith("/revoke")) return json({ request_id: "request:revoke", schema_version: "workspace_invitation_mutation.v1", invitation: record("revoked") });
    if (options?.method === "POST") return json({ request_id: "request:create", schema_version: "workspace_invitation_creation.v1", invitation: record(), invitation_code: code }, 201);
    return json(page());
  });
  const listed = await listWorkspaceInvitations(config, { effectiveState: "pending", limit: 1 });
  const created = await createWorkspaceInvitation(config, { role, ttlPolicy: "24h", confirmed: true });
  assert.ok(created.invitationCode === code);
  await revokeWorkspaceInvitation(config, { invitation: listed.invitations[0], confirmed: true });
  const checked = await previewWorkspaceInvitation(config, code);
  const claimed = await claimWorkspaceInvitation(config, { invitationCode: code, preview: checked, userId, confirmed: true });
  assert.equal(claimed.claimedByUserId, userId);
  assert.equal(requests.length, 5);
  for (const request of requests) {
    assert.equal(request.options?.credentials, "include");
    assert.equal(request.options?.cache, "no-store");
    assert.equal(request.options?.redirect, "error");
    assert.ok(!request.url.href.includes(code));
    const headers = new Headers(request.options?.headers);
    const admin = request.url.pathname.startsWith("/v1/admin/");
    assert.equal(headers.get("X-RadishMind-Active-Workspace"), admin ? config.workspaceId : null);
    assert.equal(headers.get("X-RadishMind-Active-Tenant"), config.tenantRef);
    if (request.options?.method === "POST") assert.equal(headers.get("X-RadishMind-CSRF-Token"), "csrf-proof");
    assert.equal(headers.get("Authorization"), null);
  }
  assert.deepEqual(requests[1].body, { role_key: role.roleKey, expected_catalog_version: role.catalogVersion,
    expected_role_definition_digest: role.definitionDigest, ttl_policy: "24h", confirmed: true });
  assert.deepEqual(requests[2].body, { expected_record_version: 1, confirmed: true });
  assert.ok(requests[3].body?.invitation_code === code && Object.keys(requests[3].body).length === 1);
  assert.ok(requests[4].body?.invitation_code === code && Object.keys(requests[4].body).length === 3);
  assert.equal(requests[4].body?.confirmed, true);
  assert.equal(requests[4].body?.expected_record_version, 1);
  assert.ok(!JSON.stringify(listed).includes(code) && !JSON.stringify(checked).includes(code) && !JSON.stringify(claimed).includes(code));
});

test("invitation consumer rejects malformed scope, policy, confirmation and code without sending a request", async (t) => {
  const fetch = t.mock.method(globalThis, "fetch", async () => json({}));
  const inputError = (error: unknown) => error instanceof WorkspaceInvitationError && error.code === "workspace_invitation_input_invalid";
  for (const bad of ["", "https://platform.test/?credential=hidden", "http://user:password@platform.test", "file:///tmp/test", "http://platform.test/other"]) {
    await assert.rejects(listWorkspaceInvitations({ ...config, baseUrl: bad }), inputError);
  }
  await assert.rejects(listWorkspaceInvitations({ ...config, mode: "disabled" }), inputError);
  await assert.rejects(listWorkspaceInvitations({ ...config, workspaceId: "bad space" }), inputError);
  await assert.rejects(listWorkspaceInvitations(config, { limit: 101 }), inputError);
  await assert.rejects(listWorkspaceInvitations(config, { cursor: "contains space" }), inputError);
  await assert.rejects(createWorkspaceInvitation(config, { role: { ...role, roleKey: "workspace_admin" }, ttlPolicy: "24h", confirmed: true }), inputError);
  await assert.rejects(createWorkspaceInvitation(config, { role: { ...role, canManageLocalIdentity: true }, ttlPolicy: "24h", confirmed: true }), inputError);
  await assert.rejects(createWorkspaceInvitation(config, { role, ttlPolicy: "24h", confirmed: false as true }), inputError);
  await assert.rejects(createWorkspaceInvitation(config, { role, ttlPolicy: "forever" as "24h", confirmed: true }), inputError);
  await assert.rejects(previewWorkspaceInvitation(config, "malformed"), inputError);
  assert.equal(fetch.mock.callCount(), 0);
  for (let index = 0; index < 64; index += 1) assert.ok(isInvitationCode(invitationCode()));
});

test("directory rejects unknown fields, secrets, mismatched scope and impossible canonical transitions", async (t) => {
  const mutations: Array<(value: ReturnType<typeof page>) => void> = [
    (value) => { Object.assign(value, { invitation_code: invitationCode() }); },
    (value) => { value.workspace_id = "another_workspace"; },
    (value) => { value.invitations[0].tenant_ref = "another_tenant"; },
    (value) => { value.invitations[0].secret_digest = role.definitionDigest; },
    (value) => { value.invitations[0].role_key = "workspace_admin"; },
    (value) => { value.invitations[0].record_version = 2; },
    (value) => { value.invitations[0].claimed_at = changedAt; },
    (value) => { value.invitations[0].effective_state = "claimed"; },
    (value) => { value.invitations[0].effective_state = "expired"; },
    (value) => { value.invitations[0].ttl_policy = "forever"; },
    (value) => { value.invitations[0].expires_at = changedAt; },
    (value) => { value.invitations.push(record()); },
    (value) => { value.invitations[0] = record("claimed"); delete value.invitations[0].membership_id; },
    (value) => { value.invitations[0] = record("revoked"); value.invitations[0].revoked_at = createdAt; },
  ];
  let response = page();
  t.mock.method(globalThis, "fetch", async () => json(response));
  for (const mutate of mutations) {
    response = page(); mutate(response);
    await assert.rejects(listWorkspaceInvitations(config), rejectedResponse);
  }
});

test("directory defaults to pending and binds canonical state pagination to scope, filter, limit and snapshot", async (t) => {
  const response: ReturnType<typeof page> & { next_cursor?: string } = page();
  const cursor = { schema_version: "workspace_invitation_cursor.v1", tenant_ref: config.tenantRef, workspace_id: config.workspaceId,
    effective_state: "pending", limit: 50, as_of: asOf, updated_at: createdAt, invitation_id: invitationId(), binding_digest: role.definitionDigest };
  const encode = () => Buffer.from(JSON.stringify(cursor)).toString("base64url");
  response.next_cursor = encode();
  const fetch = t.mock.method(globalThis, "fetch", async () => json(response));
  const result = await listWorkspaceInvitations(config);
  assert.equal(new URL(String(fetch.mock.calls[0].arguments[0])).searchParams.get("effective_state"), "pending");
  assert.equal(result.invitations[0].effectiveState, "pending");
  assert.equal(result.nextCursor, response.next_cursor);
  for (const key of ["tenant_ref", "workspace_id", "effective_state", "as_of"] as const) {
    const original = cursor[key]; cursor[key] = "different"; response.next_cursor = encode();
    await assert.rejects(listWorkspaceInvitations(config), rejectedResponse);
    cursor[key] = original;
  }
  response.next_cursor = "invalid";
  await assert.rejects(listWorkspaceInvitations(config), rejectedResponse);
  delete response.next_cursor;
  for (const state of ["claimed", "revoked", "expired"] as const) {
    response.invitations = [state === "expired" ? { ...record(), effective_state: "expired" } : record(state)];
    response.as_of = expiresAt;
    await assert.rejects(listWorkspaceInvitations(config), rejectedResponse);
    const filtered = await listWorkspaceInvitations(config, { effectiveState: state });
    assert.equal(filtered.invitations[0].effectiveState, state);
  }
});

test("creation rejects mismatched locator, status, schema and catalog instead of retaining one-time material", async (t) => {
  installCSRF(t);
  const code = invitationCode();
  let response: Record<string, unknown>;
  let status = 201;
  t.mock.method(globalThis, "fetch", async () => json(response, status));
  const make = () => ({ request_id: "request:create", schema_version: "workspace_invitation_creation.v1", invitation: record(), invitation_code: code });
  const mutations: Array<(value: ReturnType<typeof make>) => void> = [
    (value) => { value.invitation.invitation_id = `wsi_${"e".repeat(32)}`; },
    (value) => { value.invitation.role_key = "workspace_reader"; },
    (value) => { value.invitation.role_catalog_version = "drifted_catalog"; },
    (value) => { value.invitation.role_definition_digest = `sha256:${"b".repeat(64)}`; },
    (value) => { value.schema_version = "workspace_invitation_creation.v2"; },
    (value) => { Object.assign(value, { secret: code }); },
  ];
  for (const mutate of mutations) {
    const value = make(); mutate(value); response = value;
    await assert.rejects(createWorkspaceInvitation(config, { role, ttlPolicy: "24h", confirmed: true }), (error: unknown) =>
      rejectedResponse(error) && !String(error).includes(code));
  }
  response = make(); status = 200;
  await assert.rejects(createWorkspaceInvitation(config, { role, ttlPolicy: "24h", confirmed: true }), rejectedResponse);
});

test("claim rejects partial ownership, actor substitution, stale versions and role/scope drift", async (t) => {
  installCSRF(t);
  const code = invitationCode();
  let response: Record<string, unknown> = preview();
  const fetch = t.mock.method(globalThis, "fetch", async () => json(response));
  const checked = await previewWorkspaceInvitation(config, code);
  const input = { invitationCode: code, preview: checked, userId, confirmed: true as const };
  const mutations: Array<(value: ReturnType<typeof claim>) => void> = [
    (value) => { Reflect.deleteProperty(value, "membership"); },
    (value) => { value.membership.user_id = `usr_${"e".repeat(32)}`; },
    (value) => { value.role_assignment.workspace_id = "another_workspace"; },
    (value) => { value.role_assignment.role_key = "workspace_reader"; },
    (value) => { value.membership.record_version = 2; },
    (value) => { value.invitation.record_version = 1; },
    (value) => { value.invitation.claimed_by_user_id = `usr_${"e".repeat(32)}`; },
    (value) => { value.invitation.role_definition_digest = `sha256:${"b".repeat(64)}`; },
    (value) => { value.invitation.membership_id = `mbr_${"e".repeat(32)}`; },
    (value) => { Object.assign(value.role_assignment, { permission_grants: ["admin:all"] }); },
  ];
  for (const mutate of mutations) {
    const value = claim(); mutate(value); response = value;
    await assert.rejects(claimWorkspaceInvitation(config, input), rejectedResponse);
  }
  const before = fetch.mock.callCount();
  await assert.rejects(claimWorkspaceInvitation(config, { ...input, confirmed: false as true }));
  await assert.rejects(claimWorkspaceInvitation(config, { ...input, preview: { ...checked, effectiveState: "expired" } }));
  assert.equal(fetch.mock.callCount(), before);
});

test("preview rejects successful terminal projections and mismatched scope evidence without exposing the target", async (t) => {
  installCSRF(t);
  const code = invitationCode();
  let response = preview();
  t.mock.method(globalThis, "fetch", async () => json(response));
  for (const state of ["claimed", "revoked", "expired"]) {
    response.effective_state = state;
    await assert.rejects(previewWorkspaceInvitation(config, code), rejectedResponse);
  }
  response = preview(); response.invitation_id = `wsi_${"b".repeat(32)}`;
  await assert.rejects(previewWorkspaceInvitation(config, code), rejectedResponse);
  response = preview(); response.role.summary = code;
  await assert.rejects(previewWorkspaceInvitation(config, code), (error: unknown) => rejectedResponse(error) && !String(error).includes(code));
});

test("failed responses and transport errors never retain raw text, credentials, scope or retry mutations", async (t) => {
  installCSRF(t);
  const code = invitationCode();
  const base = { error: { message: "safe server text", type: "invalid_request_error", code: "workspace_invitation_invalid",
    request_id: "request:failure", route: "/v1/auth/workspace-invitations/preview", failure_boundary: "workspace_invitation_claim",
    metadata: { recovery: "reenter_invitation_code" } } };
  let response: unknown = base;
  const fetch = t.mock.method(globalThis, "fetch", async () => json(response, 400));
  await assert.rejects(previewWorkspaceInvitation(config, code), (error: unknown) => error instanceof WorkspaceInvitationError &&
    error.code === "workspace_invitation_invalid" && error.message !== base.error.message);
  for (const change of [
    { message: code }, { code: "unknown_server_code" }, { route: "/another/route" },
    { failure_boundary: "another_owner" }, { metadata: { recovery: "reenter_invitation_code", secret: code } },
    { invitation: { tenant_ref: "private_tenant" } },
  ]) {
    response = { error: { ...base.error, ...change } };
    await assert.rejects(previewWorkspaceInvitation(config, code), (error: unknown) => rejectedResponse(error) && !String(error).includes(code));
  }
  response = { error: { ...base.error, code: "LOCAL_IDENTITY_AUTHENTICATION_REQUIRED", failure_boundary: "local_identity" } };
  Reflect.deleteProperty((response as typeof base).error, "metadata");
  await assert.rejects(previewWorkspaceInvitation(config, code), (error: unknown) => error instanceof WorkspaceInvitationError &&
    error.code === "LOCAL_IDENTITY_AUTHENTICATION_REQUIRED");
  const before = fetch.mock.callCount();
  fetch.mock.mockImplementation(async () => { throw new Error(code); });
  await assert.rejects(previewWorkspaceInvitation(config, code), (error: unknown) => error instanceof WorkspaceInvitationError && !String(error).includes(code));
  assert.equal(fetch.mock.callCount(), before + 1);
  assert.ok(!workspaceInvitationFailure(new Error(code)).message.includes(code));
});

test("invitation transport rejects non-JSON and aborts without substituting cached or mock data", async (t) => {
  installCSRF(t);
  const code = invitationCode();
  const fetch = t.mock.method(globalThis, "fetch", async () => new Response("not JSON", { headers: { "Content-Type": "text/html" } }));
  await assert.rejects(previewWorkspaceInvitation(config, code), rejectedResponse);
  fetch.mock.mockImplementation(async () => new Response("{", { headers: { "Content-Type": "application/json" } }));
  await assert.rejects(previewWorkspaceInvitation(config, code), rejectedResponse);
  const controller = new AbortController();
  fetch.mock.mockImplementation(async (_input: unknown, options?: RequestInit) => {
    assert.equal(options?.signal, controller.signal);
    controller.abort(); throw new DOMException("Aborted", "AbortError");
  });
  await assert.rejects(previewWorkspaceInvitation(config, code, controller.signal), { name: "AbortError" });
});
