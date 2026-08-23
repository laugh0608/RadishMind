import assert from "node:assert/strict";
import test from "node:test";

import {
  LocalIdentityAdministrationError,
  assignLocalIdentityWorkspaceRole,
  createLocalIdentityWorkspaceMembership,
  listLocalIdentityWorkspaceMembers,
  localIdentityAdministrationFailureKind,
  readLocalIdentityRoleCatalog,
  readLocalIdentityWorkspaceMember,
  revokeLocalIdentityWorkspaceMembership,
  revokeLocalIdentityWorkspaceRole,
  type LocalIdentityAdministrationConfig,
} from "../src/features/local-identity/localIdentityAdministrationConsumer.ts";

const config: LocalIdentityAdministrationConfig = {
  mode: "local_identity_dev",
  baseUrl: "http://platform.test",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
};

test("local identity administration consumer connects all seven strict routes with cookie, scope, CSRF, CAS, and confirmation", async () => {
  const originalFetch = globalThis.fetch;
  const originalDocument = Object.getOwnPropertyDescriptor(globalThis, "document");
  const captured: Array<{
    url: URL;
    method: string;
    headers: Headers;
    credentials?: RequestCredentials;
    cache?: RequestCache;
    body?: Record<string, unknown>;
  }> = [];
  Object.defineProperty(globalThis, "document", {
    configurable: true,
    value: { cookie: "radishmind_csrf_dev=csrf-proof" },
  });
  globalThis.fetch = async (input, init) => {
    const url = new URL(String(input));
    const method = init?.method ?? "GET";
    const body = init?.body ? JSON.parse(String(init.body)) : undefined;
    captured.push({
      url,
      method,
      headers: new Headers(init?.headers),
      credentials: init?.credentials,
      cache: init?.cache,
      body,
    });
    if (method === "GET" && url.pathname.endsWith("/members")) return jsonResponse(memberListDocument());
    if (method === "GET" && url.pathname.includes("/members/usr_")) return jsonResponse(memberDetailDocument());
    if (method === "GET" && url.pathname.endsWith("/role-catalog")) return jsonResponse(roleCatalogDocument());
    if (method === "POST" && url.pathname.endsWith("/memberships")) {
      return jsonResponse(membershipMutationDocument(activeMembership()), 201);
    }
    if (method === "POST" && url.pathname.endsWith("/memberships/mbr_aaaaaaaaaaaaaaaa/revoke")) {
      return jsonResponse(membershipMutationDocument(revokedMembership(), [revokedAssignment()]));
    }
    if (method === "POST" && url.pathname.endsWith("/role-assignments")) {
      return jsonResponse(roleMutationDocument(activeAssignment()), 201);
    }
    if (method === "POST" && url.pathname.endsWith("/role-assignments/rla_aaaaaaaaaaaaaaaa/revoke")) {
      return jsonResponse(roleMutationDocument(revokedAssignment()));
    }
    throw new Error(`unexpected request ${method} ${url.pathname}`);
  };
  try {
    const page = await listLocalIdentityWorkspaceMembers(config, { membershipState: "active", limit: 100 });
    const detail = await readLocalIdentityWorkspaceMember(config, "usr_bbbbbbbbbbbbbbbb");
    const catalog = await readLocalIdentityRoleCatalog(config);
    const createdMembership = await createLocalIdentityWorkspaceMembership(config, {
      userId: "usr_bbbbbbbbbbbbbbbb",
      expiresAt: "2999-01-01T00:00:00+08:00",
    });
    const revokedMembershipResult = await revokeLocalIdentityWorkspaceMembership(config, {
      membershipId: "mbr_aaaaaaaaaaaaaaaa",
      expectedRecordVersion: 1,
    });
    const assignedRole = await assignLocalIdentityWorkspaceRole(config, {
      userId: "usr_bbbbbbbbbbbbbbbb",
      roleKey: "workspace_reader",
      expectedCatalogVersion: "local_identity_builtin_roles_v1",
      expectedRoleDefinitionDigest: digest("b"),
      permissionGrants: ["forbidden:client_grant"],
    } as Parameters<typeof assignLocalIdentityWorkspaceRole>[1] & { permissionGrants: string[] });
    const revokedRole = await revokeLocalIdentityWorkspaceRole(config, {
      assignmentId: "rla_aaaaaaaaaaaaaaaa",
      expectedRecordVersion: 1,
    });

    assert.equal(page.members[0]?.displayName, "Member One");
    assert.equal(detail.member.userId, "usr_bbbbbbbbbbbbbbbb");
    assert.equal(catalog.catalog.roles.length, 4);
    assert.equal(createdMembership.membership.effective, true);
    assert.equal(revokedMembershipResult.revokedRoleAssignments.length, 1);
    assert.equal(assignedRole.roleAssignment.roleKey, "workspace_reader");
    assert.equal(revokedRole.roleAssignment.lifecycleState, "revoked");
    assert.equal(captured.length, 7);

    for (const request of captured) {
      assert.equal(request.credentials, "include");
      assert.equal(request.cache, "no-store");
      assert.equal(request.headers.get("X-RadishMind-Active-Workspace"), "workspace_demo");
      assert.equal(request.headers.get("Authorization"), null);
      assert.equal(request.headers.get("X-RadishMind-Dev-Read-Identity"), null);
      if (request.method === "POST") {
        assert.equal(request.headers.get("X-RadishMind-CSRF-Token"), "csrf-proof");
        assert.equal(request.body?.confirmed, true);
      } else {
        assert.equal(request.headers.get("X-RadishMind-CSRF-Token"), null);
      }
    }
    assert.equal(captured[0]?.url.searchParams.get("membership_state"), "active");
    assert.equal(captured[0]?.url.searchParams.get("limit"), "100");
    assert.deepEqual(captured[3]?.body, {
      user_id: "usr_bbbbbbbbbbbbbbbb",
      expires_at: "2998-12-31T16:00:00.000Z",
      confirmed: true,
    });
    assert.deepEqual(captured[4]?.body, { expected_record_version: 1, confirmed: true });
    assert.deepEqual(captured[5]?.body, {
      user_id: "usr_bbbbbbbbbbbbbbbb",
      role_key: "workspace_reader",
      expected_catalog_version: "local_identity_builtin_roles_v1",
      expected_role_definition_digest: digest("b"),
      confirmed: true,
    });
    assert.equal("permission_grants" in (captured[5]?.body ?? {}), false);
    assert.deepEqual(captured[6]?.body, { expected_record_version: 1, confirmed: true });
  } finally {
    globalThis.fetch = originalFetch;
    if (originalDocument) Object.defineProperty(globalThis, "document", originalDocument);
    else Reflect.deleteProperty(globalThis, "document");
  }
});

test("administration reads reject forbidden identity fields, scope drift, unknown keys, and non-canonical arrays", async () => {
  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async () => jsonResponse({ ...memberListDocument(), email: "hidden@example.test" });
    await assert.rejects(
      () => listLocalIdentityWorkspaceMembers(config, { membershipState: "active" }),
      responseInvalid,
    );

    globalThis.fetch = async () => jsonResponse({ ...memberListDocument(), workspace_id: "workspace_other" });
    await assert.rejects(
      () => listLocalIdentityWorkspaceMembers(config, { membershipState: "active" }),
      responseInvalid,
    );

    const unknownMemberField = memberListDocument();
    unknownMemberField.members[0] = { ...unknownMemberField.members[0], debug: "forbidden" };
    globalThis.fetch = async () => jsonResponse(unknownMemberField);
    await assert.rejects(
      () => listLocalIdentityWorkspaceMembers(config, { membershipState: "active" }),
      responseInvalid,
    );

    const nonCanonicalRoleKeys = memberListDocument();
    nonCanonicalRoleKeys.members[0] = {
      ...nonCanonicalRoleKeys.members[0],
      role_keys: ["workspace_reader", "workspace_admin"],
    };
    globalThis.fetch = async () => jsonResponse(nonCanonicalRoleKeys);
    await assert.rejects(
      () => listLocalIdentityWorkspaceMembers(config, { membershipState: "active" }),
      responseInvalid,
    );
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("stable administration errors preserve only reviewed recovery metadata and classify key UI states", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => jsonResponse(administrationErrorDocument(
    "local_identity_role_catalog_mismatch",
    "refresh_role_catalog",
  ), 409);
  try {
    await assert.rejects(
      () => readLocalIdentityRoleCatalog(config),
      (error: unknown) => error instanceof LocalIdentityAdministrationError &&
        error.code === "local_identity_role_catalog_mismatch" &&
        error.recovery === "refresh_role_catalog" &&
        localIdentityAdministrationFailureKind(error) === "catalog_drift",
    );
  } finally {
    globalThis.fetch = originalFetch;
  }

  const cases: Array<[string, ReturnType<typeof localIdentityAdministrationFailureKind>]> = [
    ["workspace_permission_denied", "denied"],
    ["local_identity_admin_unavailable", "unavailable"],
    ["local_identity_membership_conflict", "stale_conflict"],
    ["local_identity_role_catalog_mismatch", "catalog_drift"],
    ["local_identity_last_admin_removal_denied", "last_admin"],
    ["local_identity_recent_authentication_required", "recent_authentication"],
    ["local_identity_response_invalid", "invalid_response"],
  ];
  for (const [code, expected] of cases) {
    assert.equal(localIdentityAdministrationFailureKind(
      new LocalIdentityAdministrationError(409, code, "stable failure"),
    ), expected);
  }
});

test("invalid mutation inputs fail before issuing a request", async () => {
  const originalFetch = globalThis.fetch;
  let calls = 0;
  globalThis.fetch = async () => {
    calls += 1;
    throw new Error("unexpected fetch");
  };
  try {
    await assert.rejects(
      () => createLocalIdentityWorkspaceMembership(config, { userId: "email@example.test" }),
      /candidate is invalid/u,
    );
    await assert.rejects(
      () => createLocalIdentityWorkspaceMembership(config, {
        userId: "usr_bbbbbbbbbbbbbbbb",
        expiresAt: "2000-01-01T00:00:00Z",
      }),
      /candidate is invalid/u,
    );
    await assert.rejects(
      () => revokeLocalIdentityWorkspaceRole(config, {
        assignmentId: "rla_invalid",
        expectedRecordVersion: 0,
      }),
      /revocation is invalid/u,
    );
    assert.equal(calls, 0);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

function memberListDocument() {
  return {
    request_id: "req_member_list",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    members: [{
      schema_version: "local_identity_workspace_member_summary.v1",
      tenant_ref: "tenant_demo",
      workspace_id: "workspace_demo",
      user_id: "usr_bbbbbbbbbbbbbbbb",
      display_name: "Member One",
      account_lifecycle_state: "active",
      membership_id: "mbr_aaaaaaaaaaaaaaaa",
      membership_lifecycle_state: "active",
      membership_record_version: 1,
      membership_effective: true,
      role_keys: ["workspace_reader"],
      can_manage_local_identity: false,
      role_catalog_drift: false,
      updated_at: "2026-08-23T12:00:00Z",
    }],
    next_cursor: "cursor_opaque_page_2",
  };
}

function memberDetailDocument() {
  return {
    request_id: "req_member_detail",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    member: {
      schema_version: "local_identity_workspace_member_detail.v1",
      tenant_ref: "tenant_demo",
      workspace_id: "workspace_demo",
      user_id: "usr_bbbbbbbbbbbbbbbb",
      display_name: "Member One",
      account_lifecycle_state: "active",
      account_record_version: 2,
      memberships: [activeMembership()],
      role_assignments: [activeAssignment()],
      can_manage_local_identity: false,
    },
  };
}

function roleCatalogDocument() {
  return {
    request_id: "req_role_catalog",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    catalog: {
      schema_version: "local_identity_role_catalog.v1",
      catalog_version: "local_identity_builtin_roles_v1",
      definition_digest: digest("f"),
      roles: [
        roleDefinition("workspace_admin", digest("a"), true, ["local_identity_members:read"]),
        roleDefinition("workspace_builder", digest("c"), false, ["applications:read", "applications:write"]),
        roleDefinition("workspace_reader", digest("b"), false, ["applications:read"]),
        roleDefinition("workspace_reviewer", digest("d"), false, ["applications:read", "workflow_definitions:review"]),
      ],
    },
  };
}

function roleDefinition(roleKey: string, definitionDigest: string, canManage: boolean, grants: string[]) {
  return {
    catalog_version: "local_identity_builtin_roles_v1",
    role_key: roleKey,
    display_name: roleKey.replaceAll("_", " "),
    summary: `Canonical ${roleKey} definition.`,
    permission_grants: grants,
    definition_digest: definitionDigest,
    can_manage_local_identity: canManage,
  };
}

function activeMembership() {
  return {
    schema_version: "local_identity_workspace_membership_view.v1",
    membership_id: "mbr_aaaaaaaaaaaaaaaa",
    lifecycle_state: "active",
    record_version: 1,
    created_at: "2026-08-23T11:00:00Z",
    updated_at: "2026-08-23T12:00:00Z",
    effective: true,
  };
}

function revokedMembership() {
  return {
    ...activeMembership(),
    lifecycle_state: "revoked",
    record_version: 2,
    updated_at: "2026-08-23T13:00:00Z",
    revoked_at: "2026-08-23T13:00:00Z",
    effective: false,
  };
}

function activeAssignment() {
  return {
    schema_version: "local_identity_workspace_role_assignment_view.v1",
    assignment_id: "rla_aaaaaaaaaaaaaaaa",
    scope: "workspace",
    workspace_id: "workspace_demo",
    role_key: "workspace_reader",
    role_catalog_version: "local_identity_builtin_roles_v1",
    role_definition_digest: digest("b"),
    permission_grants: ["applications:read"],
    lifecycle_state: "active",
    record_version: 1,
    created_at: "2026-08-23T11:30:00Z",
    updated_at: "2026-08-23T12:00:00Z",
    effective: true,
    catalog_drift: false,
    can_manage_local_identity: false,
  };
}

function revokedAssignment() {
  return {
    ...activeAssignment(),
    lifecycle_state: "revoked",
    record_version: 2,
    updated_at: "2026-08-23T13:00:00Z",
    revoked_at: "2026-08-23T13:00:00Z",
    effective: false,
  };
}

function membershipMutationDocument(membership: ReturnType<typeof activeMembership>, revokedRoleAssignments?: unknown[]) {
  return {
    request_id: "req_membership_mutation",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    membership,
    ...(revokedRoleAssignments ? { revoked_role_assignments: revokedRoleAssignments } : {}),
  };
}

function roleMutationDocument(roleAssignment: ReturnType<typeof activeAssignment>) {
  return {
    request_id: "req_role_mutation",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    role_assignment: roleAssignment,
  };
}

function administrationErrorDocument(code: string, recovery: string) {
  return {
    error: {
      message: "the role catalog changed before this request",
      type: "invalid_request_error",
      code,
      request_id: "req_catalog_conflict",
      route: "/v1/admin/local-identity/role-catalog",
      failure_boundary: "local_identity_administration",
      metadata: { recovery },
    },
  };
}

function digest(character: string): string {
  return `sha256:${character.repeat(64)}`;
}

function jsonResponse(value: unknown, status = 200): Response {
  return new Response(JSON.stringify(value), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function responseInvalid(error: unknown): boolean {
  return error instanceof LocalIdentityAdministrationError && error.code === "local_identity_response_invalid";
}
