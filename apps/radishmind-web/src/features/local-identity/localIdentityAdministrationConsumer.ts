import {
  LocalIdentityConsumerError,
  readLocalIdentityCSRFCookie,
  type LocalIdentityConsumerConfig,
} from "./localIdentityConsumer.ts";

const ACTIVE_TENANT_HEADER = "X-RadishMind-Active-Tenant";
const ACTIVE_WORKSPACE_HEADER = "X-RadishMind-Active-Workspace";
const CSRF_HEADER = "X-RadishMind-CSRF-Token";
const MEMBER_SUMMARY_SCHEMA = "local_identity_workspace_member_summary.v1";
const MEMBER_DETAIL_SCHEMA = "local_identity_workspace_member_detail.v1";
const MEMBERSHIP_SCHEMA = "local_identity_workspace_membership_view.v1";
const ASSIGNMENT_SCHEMA = "local_identity_workspace_role_assignment_view.v1";
const ROLE_CATALOG_SCHEMA = "local_identity_role_catalog.v1";

const USER_ID = /^usr_[a-z0-9]{16,64}$/u;
const MEMBERSHIP_ID = /^mbr_[a-z0-9]{16,64}$/u;
const ASSIGNMENT_ID = /^rla_[a-z0-9]{16,64}$/u;
const ROLE_KEY = /^[a-z][a-z0-9_]{2,63}$/u;
const DIGEST = /^sha256:[a-f0-9]{64}$/u;
const REFERENCE = /^[A-Za-z0-9][A-Za-z0-9_.:/-]{0,159}$/u;

const FORBIDDEN_RESPONSE_KEYS = new Set([
  "access_token",
  "audit_ref",
  "authorization",
  "cookie",
  "credential",
  "credential_digest",
  "email",
  "external_identities",
  "id_token",
  "issuer",
  "login_identifier",
  "password",
  "raw_claims",
  "refresh_token",
  "session",
  "subject",
  "token",
]);

const ADMIN_FAILURE_CODES = [
  "local_identity_admin_unavailable",
  "local_identity_admin_scope_mismatch",
  "local_identity_member_unavailable",
  "local_identity_member_cursor_invalid",
  "local_identity_role_catalog_mismatch",
  "local_identity_membership_conflict",
  "local_identity_role_assignment_conflict",
  "local_identity_self_membership_revoke_denied",
  "local_identity_last_admin_removal_denied",
  "local_identity_recent_authentication_required",
  "workspace_membership_denied",
  "workspace_permission_denied",
] as const;

const LOCAL_SESSION_FAILURE_CODES = [
  "LOCAL_IDENTITY_AUTHENTICATION_REQUIRED",
  "LOCAL_IDENTITY_CSRF_INVALID",
  "LOCAL_IDENTITY_HTTP_DISABLED",
  "LOCAL_IDENTITY_ORIGIN_FORBIDDEN",
  "LOCAL_IDENTITY_PAYLOAD_INVALID",
  "LOCAL_IDENTITY_SERVICE_UNAVAILABLE",
] as const;

export type LocalIdentityAdministrationConfig = LocalIdentityConsumerConfig & {
  tenantRef: string;
  workspaceId: string;
};

export type LocalIdentityMembershipState = "active" | "revoked";

export type LocalIdentityWorkspaceMemberSummary = {
  schemaVersion: typeof MEMBER_SUMMARY_SCHEMA;
  tenantRef: string;
  workspaceId: string;
  userId: string;
  displayName: string;
  accountLifecycleState: "active" | "disabled";
  membershipId: string;
  membershipLifecycleState: LocalIdentityMembershipState;
  membershipRecordVersion: number;
  membershipExpiresAt?: string;
  membershipEffective: boolean;
  roleKeys: string[];
  canManageLocalIdentity: boolean;
  roleCatalogDrift: boolean;
  updatedAt: string;
};

export type LocalIdentityWorkspaceMemberPage = {
  requestId: string;
  tenantRef: string;
  workspaceId: string;
  members: LocalIdentityWorkspaceMemberSummary[];
  nextCursor: string;
};

export type LocalIdentityWorkspaceMembershipView = {
  schemaVersion: typeof MEMBERSHIP_SCHEMA;
  membershipId: string;
  lifecycleState: LocalIdentityMembershipState;
  recordVersion: number;
  createdAt: string;
  updatedAt: string;
  expiresAt?: string;
  revokedAt?: string;
  effective: boolean;
};

export type LocalIdentityWorkspaceRoleAssignmentView = {
  schemaVersion: typeof ASSIGNMENT_SCHEMA;
  assignmentId: string;
  scope: "tenant" | "workspace";
  workspaceId?: string;
  roleKey: string;
  roleCatalogVersion?: string;
  roleDefinitionDigest?: string;
  permissionGrants: string[];
  lifecycleState: LocalIdentityMembershipState;
  recordVersion: number;
  createdAt: string;
  updatedAt: string;
  expiresAt?: string;
  revokedAt?: string;
  effective: boolean;
  catalogDrift: boolean;
  canManageLocalIdentity: boolean;
};

export type LocalIdentityWorkspaceMemberDetail = {
  schemaVersion: typeof MEMBER_DETAIL_SCHEMA;
  tenantRef: string;
  workspaceId: string;
  userId: string;
  displayName: string;
  accountLifecycleState: "active" | "disabled";
  accountRecordVersion: number;
  memberships: LocalIdentityWorkspaceMembershipView[];
  roleAssignments: LocalIdentityWorkspaceRoleAssignmentView[];
  canManageLocalIdentity: boolean;
};

export type LocalIdentityRoleDefinition = {
  catalogVersion: string;
  roleKey: string;
  displayName: string;
  summary: string;
  permissionGrants: string[];
  definitionDigest: string;
  canManageLocalIdentity: boolean;
};

export type LocalIdentityRoleCatalog = {
  schemaVersion: typeof ROLE_CATALOG_SCHEMA;
  catalogVersion: string;
  definitionDigest: string;
  roles: LocalIdentityRoleDefinition[];
};

export type LocalIdentityMembershipMutationResult = {
  requestId: string;
  tenantRef: string;
  workspaceId: string;
  membership: LocalIdentityWorkspaceMembershipView;
  revokedRoleAssignments: LocalIdentityWorkspaceRoleAssignmentView[];
};

export type LocalIdentityRoleAssignmentMutationResult = {
  requestId: string;
  tenantRef: string;
  workspaceId: string;
  roleAssignment: LocalIdentityWorkspaceRoleAssignmentView;
};

export type LocalIdentityAdministrationFailureKind =
  | "denied"
  | "unavailable"
  | "stale_conflict"
  | "catalog_drift"
  | "last_admin"
  | "recent_authentication"
  | "invalid_response"
  | "failed";

export class LocalIdentityAdministrationError extends Error {
  readonly status: number;
  readonly code: string;
  readonly recovery: string;

  constructor(status: number, code: string, message: string, recovery = "") {
    super(message);
    this.name = "LocalIdentityAdministrationError";
    this.status = status;
    this.code = code;
    this.recovery = recovery;
  }
}

export function localIdentityAdministrationFailureKind(
  error: unknown,
): LocalIdentityAdministrationFailureKind {
  if (!(error instanceof LocalIdentityAdministrationError)) return "failed";
  if (error.code === "local_identity_response_invalid") return "invalid_response";
  if ([
    "local_identity_admin_scope_mismatch",
    "workspace_membership_denied",
    "workspace_permission_denied",
    "LOCAL_IDENTITY_AUTHENTICATION_REQUIRED",
  ].includes(error.code)) return "denied";
  if ([
    "local_identity_admin_unavailable",
    "LOCAL_IDENTITY_HTTP_DISABLED",
    "LOCAL_IDENTITY_SERVICE_UNAVAILABLE",
  ].includes(error.code)) return "unavailable";
  if ([
    "local_identity_membership_conflict",
    "local_identity_role_assignment_conflict",
  ].includes(error.code)) return "stale_conflict";
  if (error.code === "local_identity_role_catalog_mismatch") return "catalog_drift";
  if ([
    "local_identity_self_membership_revoke_denied",
    "local_identity_last_admin_removal_denied",
  ].includes(error.code)) return "last_admin";
  if (error.code === "local_identity_recent_authentication_required") return "recent_authentication";
  return "failed";
}

export function isValidLocalIdentityAdministrationUserId(value: string): boolean {
  return USER_ID.test(value.trim());
}

export function normalizeLocalIdentityAdministrationExpiration(value: string): string | undefined | null {
  const normalized = value.trim();
  if (!normalized) return undefined;
  if (!isTimestamp(normalized)) return null;
  const expiration = new Date(normalized);
  if (expiration.getTime() <= Date.now()) return null;
  return expiration.toISOString();
}

export async function listLocalIdentityWorkspaceMembers(
  config: LocalIdentityAdministrationConfig,
  query: { membershipState: LocalIdentityMembershipState; limit?: number; cursor?: string },
  signal?: AbortSignal,
): Promise<LocalIdentityWorkspaceMemberPage> {
  assertConfig(config);
  const limit = query.limit ?? 50;
  if ((query.membershipState !== "active" && query.membershipState !== "revoked") ||
    !Number.isInteger(limit) || limit < 1 || limit > 100 ||
    (query.cursor !== undefined && !isValidCursor(query.cursor))) {
    throw inputError("Workspace member query is invalid.");
  }
  const search = new URLSearchParams({ membership_state: query.membershipState, limit: String(limit) });
  if (query.cursor) search.set("cursor", query.cursor);
  const path = `${workspacePath(config)}/members?${search}`;
  const { response, value } = await requestAdministration(config, path, "GET", undefined, signal);
  if (response.status !== 200 || !isExactRecord(value, [
    "request_id", "tenant_ref", "workspace_id", "members",
  ], ["next_cursor"]) || !matchesEnvelopeScope(value, config) || !isReference(value.request_id) ||
    !Array.isArray(value.members) || value.members.length > limit ||
    (value.next_cursor !== undefined && !isValidCursor(value.next_cursor))) {
    throw invalidResponse(response.status);
  }
  return {
    requestId: value.request_id,
    tenantRef: value.tenant_ref,
    workspaceId: value.workspace_id,
    members: value.members.map((member) => parseMemberSummary(member, config)),
    nextCursor: value.next_cursor ?? "",
  };
}

export async function readLocalIdentityWorkspaceMember(
  config: LocalIdentityAdministrationConfig,
  userId: string,
  signal?: AbortSignal,
): Promise<{ requestId: string; member: LocalIdentityWorkspaceMemberDetail }> {
  assertConfig(config);
  const normalizedUserId = userId.trim();
  if (!USER_ID.test(normalizedUserId)) throw inputError("Workspace member ID is invalid.");
  const path = `${workspacePath(config)}/members/${encodeURIComponent(normalizedUserId)}`;
  const { response, value } = await requestAdministration(config, path, "GET", undefined, signal);
  if (response.status !== 200 || !isExactRecord(value, [
    "request_id", "tenant_ref", "workspace_id", "member",
  ]) || !matchesEnvelopeScope(value, config) || !isReference(value.request_id)) {
    throw invalidResponse(response.status);
  }
  const member = parseMemberDetail(value.member, config);
  if (member.userId !== normalizedUserId) throw invalidResponse(response.status);
  return { requestId: value.request_id, member };
}

export async function readLocalIdentityRoleCatalog(
  config: LocalIdentityAdministrationConfig,
  signal?: AbortSignal,
): Promise<{ requestId: string; catalog: LocalIdentityRoleCatalog }> {
  assertConfig(config);
  const { response, value } = await requestAdministration(
    config,
    "/v1/admin/local-identity/role-catalog",
    "GET",
    undefined,
    signal,
  );
  if (response.status !== 200 || !isExactRecord(value, [
    "request_id", "tenant_ref", "workspace_id", "catalog",
  ]) || !matchesEnvelopeScope(value, config) || !isReference(value.request_id)) {
    throw invalidResponse(response.status);
  }
  return { requestId: value.request_id, catalog: parseRoleCatalog(value.catalog) };
}

export async function createLocalIdentityWorkspaceMembership(
  config: LocalIdentityAdministrationConfig,
  input: { userId: string; expiresAt?: string },
): Promise<LocalIdentityMembershipMutationResult> {
  const userId = input.userId.trim();
  const expiresAt = normalizeLocalIdentityAdministrationExpiration(input.expiresAt ?? "");
  if (!USER_ID.test(userId) || expiresAt === null) {
    throw inputError("Workspace membership candidate is invalid.");
  }
  const result = await mutateMembership(config, `${workspacePath(config)}/memberships`, {
    user_id: userId,
    ...(expiresAt ? { expires_at: expiresAt } : {}),
    confirmed: true,
  }, 201);
  if (result.membership.lifecycleState !== "active" || !result.membership.effective ||
    result.revokedRoleAssignments.length !== 0) throw invalidResponse(201);
  return result;
}

export async function revokeLocalIdentityWorkspaceMembership(
  config: LocalIdentityAdministrationConfig,
  input: { membershipId: string; expectedRecordVersion: number },
): Promise<LocalIdentityMembershipMutationResult> {
  const membershipId = input.membershipId.trim();
  if (!MEMBERSHIP_ID.test(membershipId) || !isPositiveInteger(input.expectedRecordVersion)) {
    throw inputError("Workspace membership revocation is invalid.");
  }
  const result = await mutateMembership(
    config,
    `${workspacePath(config)}/memberships/${encodeURIComponent(membershipId)}/revoke`,
    { expected_record_version: input.expectedRecordVersion, confirmed: true },
    200,
  );
  if (result.membership.membershipId !== membershipId || result.membership.lifecycleState !== "revoked" ||
    result.membership.effective || result.revokedRoleAssignments.some((item) => item.lifecycleState !== "revoked")) {
    throw invalidResponse(200);
  }
  return result;
}

export async function assignLocalIdentityWorkspaceRole(
  config: LocalIdentityAdministrationConfig,
  input: {
    userId: string;
    roleKey: string;
    expectedCatalogVersion: string;
    expectedRoleDefinitionDigest: string;
    expiresAt?: string;
  },
): Promise<LocalIdentityRoleAssignmentMutationResult> {
  const normalized = {
    userId: input.userId.trim(),
    roleKey: input.roleKey.trim(),
    expectedCatalogVersion: input.expectedCatalogVersion.trim(),
    expectedRoleDefinitionDigest: input.expectedRoleDefinitionDigest.trim(),
    expiresAt: normalizeLocalIdentityAdministrationExpiration(input.expiresAt ?? ""),
  };
  if (!USER_ID.test(normalized.userId) || !ROLE_KEY.test(normalized.roleKey) ||
    !isReference(normalized.expectedCatalogVersion) || !DIGEST.test(normalized.expectedRoleDefinitionDigest) ||
    normalized.expiresAt === null) {
    throw inputError("Workspace role assignment candidate is invalid.");
  }
  const result = await mutateRoleAssignment(config, `${workspacePath(config)}/role-assignments`, {
    user_id: normalized.userId,
    role_key: normalized.roleKey,
    expected_catalog_version: normalized.expectedCatalogVersion,
    expected_role_definition_digest: normalized.expectedRoleDefinitionDigest,
    ...(normalized.expiresAt ? { expires_at: normalized.expiresAt } : {}),
    confirmed: true,
  }, 201);
  if (result.roleAssignment.roleKey !== normalized.roleKey ||
    result.roleAssignment.roleCatalogVersion !== normalized.expectedCatalogVersion ||
    result.roleAssignment.roleDefinitionDigest !== normalized.expectedRoleDefinitionDigest ||
    result.roleAssignment.lifecycleState !== "active" || !result.roleAssignment.effective) {
    throw invalidResponse(201);
  }
  return result;
}

export async function revokeLocalIdentityWorkspaceRole(
  config: LocalIdentityAdministrationConfig,
  input: { assignmentId: string; expectedRecordVersion: number },
): Promise<LocalIdentityRoleAssignmentMutationResult> {
  const assignmentId = input.assignmentId.trim();
  if (!ASSIGNMENT_ID.test(assignmentId) || !isPositiveInteger(input.expectedRecordVersion)) {
    throw inputError("Workspace role revocation is invalid.");
  }
  const result = await mutateRoleAssignment(
    config,
    `${workspacePath(config)}/role-assignments/${encodeURIComponent(assignmentId)}/revoke`,
    { expected_record_version: input.expectedRecordVersion, confirmed: true },
    200,
  );
  if (result.roleAssignment.assignmentId !== assignmentId ||
    result.roleAssignment.lifecycleState !== "revoked" || result.roleAssignment.effective) {
    throw invalidResponse(200);
  }
  return result;
}

async function mutateMembership(
  config: LocalIdentityAdministrationConfig,
  path: string,
  body: Record<string, unknown>,
  expectedStatus: 200 | 201,
): Promise<LocalIdentityMembershipMutationResult> {
  assertConfig(config);
  const { response, value } = await requestAdministration(config, path, "POST", body);
  if (response.status !== expectedStatus || !isExactRecord(value, [
    "request_id", "tenant_ref", "workspace_id", "membership",
  ], ["revoked_role_assignments"]) || !matchesEnvelopeScope(value, config) || !isReference(value.request_id) ||
    (value.revoked_role_assignments !== undefined && !Array.isArray(value.revoked_role_assignments))) {
    throw invalidResponse(response.status);
  }
  return {
    requestId: value.request_id,
    tenantRef: value.tenant_ref,
    workspaceId: value.workspace_id,
    membership: parseMembership(value.membership),
    revokedRoleAssignments: (value.revoked_role_assignments ?? []).map((item: unknown) =>
      parseRoleAssignment(item, config)
    ),
  };
}

async function mutateRoleAssignment(
  config: LocalIdentityAdministrationConfig,
  path: string,
  body: Record<string, unknown>,
  expectedStatus: 200 | 201,
): Promise<LocalIdentityRoleAssignmentMutationResult> {
  assertConfig(config);
  const { response, value } = await requestAdministration(config, path, "POST", body);
  if (response.status !== expectedStatus || !isExactRecord(value, [
    "request_id", "tenant_ref", "workspace_id", "role_assignment",
  ]) || !matchesEnvelopeScope(value, config) || !isReference(value.request_id)) {
    throw invalidResponse(response.status);
  }
  return {
    requestId: value.request_id,
    tenantRef: value.tenant_ref,
    workspaceId: value.workspace_id,
    roleAssignment: parseRoleAssignment(value.role_assignment, config),
  };
}

async function requestAdministration(
  config: LocalIdentityAdministrationConfig,
  path: string,
  method: "GET" | "POST",
  body?: Record<string, unknown>,
  signal?: AbortSignal,
): Promise<{ response: Response; value: Record<string, unknown> }> {
  assertConfig(config);
  const response = await fetch(`${config.baseUrl}${path}`, {
    method,
    credentials: "include",
    cache: "no-store",
    headers: {
      Accept: "application/json",
      [ACTIVE_TENANT_HEADER]: config.tenantRef,
      [ACTIVE_WORKSPACE_HEADER]: config.workspaceId,
      ...(method === "POST" ? {
        "Content-Type": "application/json",
        [CSRF_HEADER]: readLocalIdentityCSRFCookie(),
      } : {}),
    },
    body: body === undefined ? undefined : JSON.stringify(body),
    signal,
  });
  const value = await readResponseRecord(response);
  if (!response.ok) throw parseAdministrationError(response.status, value);
  if (containsForbiddenResponse(value)) throw invalidResponse(response.status);
  return { response, value };
}

async function readResponseRecord(response: Response): Promise<Record<string, unknown>> {
  if (!(response.headers.get("Content-Type") ?? "").toLowerCase().includes("application/json")) {
    throw invalidResponse(response.status);
  }
  try {
    const value: unknown = await response.json();
    if (!isRecord(value)) throw invalidResponse(response.status);
    return value;
  } catch (error) {
    if (error instanceof LocalIdentityAdministrationError) throw error;
    throw invalidResponse(response.status);
  }
}

function parseAdministrationError(status: number, value: Record<string, unknown>): LocalIdentityAdministrationError {
  if (containsForbiddenResponse(value) || !isExactRecord(value, ["error"]) || !isRecord(value.error)) {
    return invalidResponse(status);
  }
  const error = value.error;
  if (!isBoundedText(error.message, 512) || !isReference(error.type) || !isReference(error.request_id) ||
    typeof error.route !== "string" || !error.route.startsWith("/v1/admin/local-identity/") ||
    typeof error.code !== "string") return invalidResponse(status);
  if (error.failure_boundary === "local_identity_administration") {
    if (!isExactRecord(error, [
      "message", "type", "code", "request_id", "route", "failure_boundary", "metadata",
    ]) || !isAdminFailureCode(error.code) || !isExactRecord(error.metadata, ["recovery"]) ||
      !isReference(error.metadata.recovery)) return invalidResponse(status);
    return new LocalIdentityAdministrationError(status, error.code, error.message, error.metadata.recovery);
  }
  if (error.failure_boundary === "local_identity") {
    if (!isExactRecord(error, [
      "message", "type", "code", "request_id", "route", "failure_boundary",
    ]) || !isLocalSessionFailureCode(error.code)) return invalidResponse(status);
    return new LocalIdentityAdministrationError(status, error.code, error.message);
  }
  return invalidResponse(status);
}

function parseMemberSummary(
  value: unknown,
  config: LocalIdentityAdministrationConfig,
): LocalIdentityWorkspaceMemberSummary {
  if (!isExactRecord(value, [
    "schema_version", "tenant_ref", "workspace_id", "user_id", "display_name",
    "account_lifecycle_state", "membership_id", "membership_lifecycle_state",
    "membership_record_version", "membership_effective", "role_keys",
    "can_manage_local_identity", "role_catalog_drift", "updated_at",
  ], ["membership_expires_at"]) || value.schema_version !== MEMBER_SUMMARY_SCHEMA ||
    value.tenant_ref !== config.tenantRef || value.workspace_id !== config.workspaceId ||
    !USER_ID.test(String(value.user_id)) || !isDisplayName(value.display_name) ||
    !isLifecycle(value.account_lifecycle_state) || !MEMBERSHIP_ID.test(String(value.membership_id)) ||
    !isMembershipState(value.membership_lifecycle_state) || !isPositiveInteger(value.membership_record_version) ||
    typeof value.membership_effective !== "boolean" || !isCanonicalRoleKeys(value.role_keys) ||
    typeof value.can_manage_local_identity !== "boolean" || typeof value.role_catalog_drift !== "boolean" ||
    !isTimestamp(value.updated_at) || (value.membership_expires_at !== undefined && !isTimestamp(value.membership_expires_at))) {
    throw invalidResponse();
  }
  return {
    schemaVersion: MEMBER_SUMMARY_SCHEMA,
    tenantRef: value.tenant_ref,
    workspaceId: value.workspace_id,
    userId: value.user_id,
    displayName: value.display_name,
    accountLifecycleState: value.account_lifecycle_state,
    membershipId: value.membership_id,
    membershipLifecycleState: value.membership_lifecycle_state,
    membershipRecordVersion: value.membership_record_version,
    ...(value.membership_expires_at === undefined ? {} : { membershipExpiresAt: value.membership_expires_at }),
    membershipEffective: value.membership_effective,
    roleKeys: [...value.role_keys],
    canManageLocalIdentity: value.can_manage_local_identity,
    roleCatalogDrift: value.role_catalog_drift,
    updatedAt: value.updated_at,
  };
}

function parseMemberDetail(
  value: unknown,
  config: LocalIdentityAdministrationConfig,
): LocalIdentityWorkspaceMemberDetail {
  if (!isExactRecord(value, [
    "schema_version", "tenant_ref", "workspace_id", "user_id", "display_name",
    "account_lifecycle_state", "account_record_version", "memberships", "role_assignments",
    "can_manage_local_identity",
  ]) || value.schema_version !== MEMBER_DETAIL_SCHEMA || value.tenant_ref !== config.tenantRef ||
    value.workspace_id !== config.workspaceId || !USER_ID.test(String(value.user_id)) ||
    !isDisplayName(value.display_name) || !isLifecycle(value.account_lifecycle_state) ||
    !isPositiveInteger(value.account_record_version) || !Array.isArray(value.memberships) ||
    !Array.isArray(value.role_assignments) || typeof value.can_manage_local_identity !== "boolean") {
    throw invalidResponse();
  }
  return {
    schemaVersion: MEMBER_DETAIL_SCHEMA,
    tenantRef: value.tenant_ref,
    workspaceId: value.workspace_id,
    userId: value.user_id,
    displayName: value.display_name,
    accountLifecycleState: value.account_lifecycle_state,
    accountRecordVersion: value.account_record_version,
    memberships: value.memberships.map(parseMembership),
    roleAssignments: value.role_assignments.map((item) => parseRoleAssignment(item, config)),
    canManageLocalIdentity: value.can_manage_local_identity,
  };
}

function parseMembership(value: unknown): LocalIdentityWorkspaceMembershipView {
  if (!isExactRecord(value, [
    "schema_version", "membership_id", "lifecycle_state", "record_version", "created_at",
    "updated_at", "effective",
  ], ["expires_at", "revoked_at"]) || value.schema_version !== MEMBERSHIP_SCHEMA ||
    !MEMBERSHIP_ID.test(String(value.membership_id)) || !isMembershipState(value.lifecycle_state) ||
    !isPositiveInteger(value.record_version) || !isTimestamp(value.created_at) || !isTimestamp(value.updated_at) ||
    (value.expires_at !== undefined && !isTimestamp(value.expires_at)) ||
    (value.revoked_at !== undefined && !isTimestamp(value.revoked_at)) || typeof value.effective !== "boolean") {
    throw invalidResponse();
  }
  return {
    schemaVersion: MEMBERSHIP_SCHEMA,
    membershipId: value.membership_id,
    lifecycleState: value.lifecycle_state,
    recordVersion: value.record_version,
    createdAt: value.created_at,
    updatedAt: value.updated_at,
    ...(value.expires_at === undefined ? {} : { expiresAt: value.expires_at }),
    ...(value.revoked_at === undefined ? {} : { revokedAt: value.revoked_at }),
    effective: value.effective,
  };
}

function parseRoleAssignment(
  value: unknown,
  config: LocalIdentityAdministrationConfig,
): LocalIdentityWorkspaceRoleAssignmentView {
  if (!isExactRecord(value, [
    "schema_version", "assignment_id", "scope", "role_key", "permission_grants", "lifecycle_state",
    "record_version", "created_at", "updated_at", "effective", "catalog_drift", "can_manage_local_identity",
  ], [
    "workspace_id", "role_catalog_version", "role_definition_digest", "expires_at", "revoked_at",
  ]) || value.schema_version !== ASSIGNMENT_SCHEMA || !ASSIGNMENT_ID.test(String(value.assignment_id)) ||
    (value.scope !== "tenant" && value.scope !== "workspace") ||
    (value.scope === "workspace" && value.workspace_id !== config.workspaceId) ||
    (value.scope === "tenant" && value.workspace_id !== undefined) || !ROLE_KEY.test(String(value.role_key)) ||
    !isCanonicalReferences(value.permission_grants) || !isMembershipState(value.lifecycle_state) ||
    !isPositiveInteger(value.record_version) || !isTimestamp(value.created_at) || !isTimestamp(value.updated_at) ||
    (value.role_catalog_version !== undefined && !isReference(value.role_catalog_version)) ||
    (value.role_definition_digest !== undefined && !DIGEST.test(String(value.role_definition_digest))) ||
    (value.expires_at !== undefined && !isTimestamp(value.expires_at)) ||
    (value.revoked_at !== undefined && !isTimestamp(value.revoked_at)) || typeof value.effective !== "boolean" ||
    typeof value.catalog_drift !== "boolean" || typeof value.can_manage_local_identity !== "boolean") {
    throw invalidResponse();
  }
  return {
    schemaVersion: ASSIGNMENT_SCHEMA,
    assignmentId: value.assignment_id,
    scope: value.scope,
    ...(value.workspace_id === undefined ? {} : { workspaceId: value.workspace_id }),
    roleKey: value.role_key,
    ...(value.role_catalog_version === undefined ? {} : { roleCatalogVersion: value.role_catalog_version }),
    ...(value.role_definition_digest === undefined ? {} : { roleDefinitionDigest: value.role_definition_digest }),
    permissionGrants: [...value.permission_grants],
    lifecycleState: value.lifecycle_state,
    recordVersion: value.record_version,
    createdAt: value.created_at,
    updatedAt: value.updated_at,
    ...(value.expires_at === undefined ? {} : { expiresAt: value.expires_at }),
    ...(value.revoked_at === undefined ? {} : { revokedAt: value.revoked_at }),
    effective: value.effective,
    catalogDrift: value.catalog_drift,
    canManageLocalIdentity: value.can_manage_local_identity,
  };
}

function parseRoleCatalog(value: unknown): LocalIdentityRoleCatalog {
  if (!isExactRecord(value, ["schema_version", "catalog_version", "definition_digest", "roles"]) ||
    value.schema_version !== ROLE_CATALOG_SCHEMA || !isReference(value.catalog_version) ||
    !DIGEST.test(String(value.definition_digest)) || !Array.isArray(value.roles) || value.roles.length !== 4) {
    throw invalidResponse();
  }
  const roles = value.roles.map(parseRoleDefinition);
  if (!isCanonicalStringValues(roles.map((role) => role.roleKey)) ||
    roles.some((role) => role.catalogVersion !== value.catalog_version) ||
    roles.filter((role) => role.canManageLocalIdentity).length !== 1 ||
    roles.find((role) => role.canManageLocalIdentity)?.roleKey !== "workspace_admin") {
    throw invalidResponse();
  }
  return {
    schemaVersion: ROLE_CATALOG_SCHEMA,
    catalogVersion: value.catalog_version,
    definitionDigest: value.definition_digest,
    roles,
  };
}

function parseRoleDefinition(value: unknown): LocalIdentityRoleDefinition {
  if (!isExactRecord(value, [
    "catalog_version", "role_key", "display_name", "summary", "permission_grants",
    "definition_digest", "can_manage_local_identity",
  ]) || !isReference(value.catalog_version) || !ROLE_KEY.test(String(value.role_key)) ||
    !isBoundedText(value.display_name, 120) || !isBoundedText(value.summary, 512) ||
    !isCanonicalReferences(value.permission_grants) || !DIGEST.test(String(value.definition_digest)) ||
    typeof value.can_manage_local_identity !== "boolean") throw invalidResponse();
  return {
    catalogVersion: value.catalog_version,
    roleKey: value.role_key,
    displayName: value.display_name,
    summary: value.summary,
    permissionGrants: [...value.permission_grants],
    definitionDigest: value.definition_digest,
    canManageLocalIdentity: value.can_manage_local_identity,
  };
}

function workspacePath(config: LocalIdentityAdministrationConfig): string {
  return `/v1/admin/local-identity/workspaces/${encodeURIComponent(config.workspaceId)}`;
}

function assertConfig(config: LocalIdentityAdministrationConfig): void {
  if (config.mode !== "local_identity_dev") {
    throw new LocalIdentityAdministrationError(
      0,
      "LOCAL_IDENTITY_HTTP_DISABLED",
      "The local identity administration consumer is disabled.",
    );
  }
  if (!isReference(config.tenantRef) || !isReference(config.workspaceId) || !isSafeBaseUrl(config.baseUrl)) {
    throw inputError("Local identity administration scope is invalid.");
  }
}

function matchesEnvelopeScope(
  value: Record<string, unknown>,
  config: LocalIdentityAdministrationConfig,
): value is Record<string, unknown> & { tenant_ref: string; workspace_id: string } {
  return value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId;
}

function containsForbiddenResponse(value: unknown): boolean {
  if (Array.isArray(value)) return value.some(containsForbiddenResponse);
  if (!isRecord(value)) return false;
  return Object.entries(value).some(([key, child]) =>
    FORBIDDEN_RESPONSE_KEYS.has(key.toLowerCase()) || containsForbiddenResponse(child)
  );
}

function isExactRecord(
  value: unknown,
  required: readonly string[],
  optional: readonly string[] = [],
): value is Record<string, any> {
  if (!isRecord(value)) return false;
  const keys = Object.keys(value);
  return required.every((key) => keys.includes(key)) &&
    keys.every((key) => required.includes(key) || optional.includes(key));
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function isReference(value: unknown): value is string {
  return typeof value === "string" && REFERENCE.test(value);
}

function isDisplayName(value: unknown): value is string {
  return isBoundedText(value, 120) && !/[\0\r\n]/u.test(value);
}

function isBoundedText(value: unknown, maximum: number): value is string {
  return typeof value === "string" && value.trim().length > 0 && value.length <= maximum;
}

function isTimestamp(value: unknown): value is string {
  return typeof value === "string" && value.length <= 64 && !Number.isNaN(Date.parse(value));
}

function isPositiveInteger(value: unknown): value is number {
  return Number.isInteger(value) && Number(value) > 0;
}

function isLifecycle(value: unknown): value is "active" | "disabled" {
  return value === "active" || value === "disabled";
}

function isMembershipState(value: unknown): value is LocalIdentityMembershipState {
  return value === "active" || value === "revoked";
}

function isCanonicalRoleKeys(value: unknown): value is string[] {
  return Array.isArray(value) && value.length <= 32 && value.every((entry) =>
    typeof entry === "string" && ROLE_KEY.test(entry)
  ) && isCanonicalStringValues(value);
}

function isCanonicalReferences(value: unknown): value is string[] {
  return Array.isArray(value) && value.length <= 128 && value.every(isReference) &&
    isCanonicalStringValues(value);
}

function isCanonicalStringValues(values: string[]): boolean {
  return values.every((value, index) => index === 0 || values[index - 1] < value);
}

function isValidCursor(value: unknown): value is string {
  return typeof value === "string" && value.length > 0 && value.length <= 8192 && !/[\0\r\n]/u.test(value);
}

function isAdminFailureCode(value: string): value is typeof ADMIN_FAILURE_CODES[number] {
  return ADMIN_FAILURE_CODES.includes(value as typeof ADMIN_FAILURE_CODES[number]);
}

function isLocalSessionFailureCode(value: string): value is typeof LOCAL_SESSION_FAILURE_CODES[number] {
  return LOCAL_SESSION_FAILURE_CODES.includes(value as typeof LOCAL_SESSION_FAILURE_CODES[number]);
}

function isSafeBaseUrl(value: string): boolean {
  try {
    const url = new URL(value);
    return (url.protocol === "https:" || url.protocol === "http:") && !url.username && !url.password &&
      !url.search && !url.hash;
  } catch {
    return false;
  }
}

function inputError(message: string): LocalIdentityAdministrationError {
  return new LocalIdentityAdministrationError(0, "local_identity_input_invalid", message);
}

function invalidResponse(status = 0): LocalIdentityAdministrationError {
  return new LocalIdentityAdministrationError(
    status,
    "local_identity_response_invalid",
    "The local identity administration response is invalid.",
  );
}

export function toLocalIdentityAdministrationError(error: unknown): LocalIdentityAdministrationError {
  if (error instanceof LocalIdentityAdministrationError) return error;
  if (error instanceof LocalIdentityConsumerError) {
    return new LocalIdentityAdministrationError(error.status, error.code, error.message);
  }
  return new LocalIdentityAdministrationError(
    0,
    "local_identity_admin_unavailable",
    "Local identity administration could not be verified.",
    "retry_after_service_recovery",
  );
}
