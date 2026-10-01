import {
  readLocalIdentityCSRFCookie,
  type LocalIdentityConsumerConfig,
} from "./localIdentityConsumer.ts";
import type { LocalIdentityRoleDefinition } from "./localIdentityAdministrationConsumer.ts";

const REFERENCE = /^[A-Za-z0-9][A-Za-z0-9_.:/-]{0,159}$/u;
const INVITATION_ID = /^wsi_[a-f0-9]{32}$/u;
const USER_ID = /^usr_[a-z0-9]{16,64}$/u;
const MEMBERSHIP_ID = /^mbr_[a-z0-9]{16,64}$/u;
const ASSIGNMENT_ID = /^rla_[a-z0-9]{16,64}$/u;
const DIGEST = /^sha256:[a-f0-9]{64}$/u;
const INVITATION_CODE = /^rmi_([a-f0-9]{32})\.[A-Za-z0-9_-]{42}[AEIMQUYcgkosw048]$/u;
const SECRET_IN_TEXT = /rmi_[a-f0-9]{32}\.[A-Za-z0-9_-]{43}/u;
export const INVITATION_ROLES = ["workspace_reader", "workspace_builder", "workspace_reviewer"] as const;
export const INVITATION_TTLS = ["1h", "24h", "72h", "7d"] as const;
const EFFECTIVE_STATES = ["pending", "claimed", "revoked", "expired"] as const;
const TTL_MILLISECONDS = { "1h": 3600000, "24h": 86400000, "72h": 259200000, "7d": 604800000 };
const ADMIN_ROUTE = "/v1/admin/local-identity/workspaces/{workspace_id}/invitations";
const CLAIM_ROUTE = "/v1/auth/workspace-invitations";

export type InvitationRole = typeof INVITATION_ROLES[number];
export type InvitationTTL = typeof INVITATION_TTLS[number];
export type InvitationEffectiveState = typeof EFFECTIVE_STATES[number];
export type InvitationClaimantConfig = LocalIdentityConsumerConfig & { tenantRef: string };
export type InvitationAdminConfig = InvitationClaimantConfig & { workspaceId: string };
export type InvitationScope = { tenantRef: string; workspaceId: string };

export type WorkspaceInvitation = InvitationScope & {
  invitationId: string;
  recordVersion: number;
  roleKey: InvitationRole;
  roleCatalogVersion: string;
  roleDefinitionDigest: string;
  ttlPolicy: InvitationTTL;
  lifecycleState: "pending" | "claimed" | "revoked";
  effectiveState: InvitationEffectiveState;
  expiresAt: string;
  createdAt: string;
  updatedAt: string;
  claimedAt?: string;
  claimedByUserId?: string;
  membershipId?: string;
  assignmentId?: string;
  revokedAt?: string;
};

export type WorkspaceInvitationPreview = InvitationScope & {
  invitationId: string;
  recordVersion: number;
  role: {
    roleKey: InvitationRole;
    displayName: string;
    summary: string;
    catalogVersion: string;
    definitionDigest: string;
  };
  effectiveState: InvitationEffectiveState;
  expiresAt: string;
};

export type WorkspaceInvitationPage = InvitationScope & {
  requestId: string;
  asOf: string;
  invitations: WorkspaceInvitation[];
  nextCursor: string;
};

const FAILURE_GUIDANCE: Record<string, string> = {
  workspace_invitation_cursor_invalid: "邀请目录已变化，请重新加载列表。",
  workspace_invitation_role_ineligible: "此角色不能用于邀请，请重新选择角色。",
  workspace_invitation_role_catalog_mismatch: "角色定义已变化，请重新加载角色目录并创建新邀请。",
  workspace_invitation_version_conflict: "邀请已被其他操作修改，请重新读取状态。",
  workspace_invitation_transition_invalid: "此邀请已不能执行该操作，请刷新目录。",
  workspace_invitation_invalid: "邀请码无效，请重新输入。",
  workspace_invitation_not_claimable: "此邀请已不能认领，请向管理员申请新的邀请。",
  workspace_invitation_account_ineligible: "当前账户不符合认领条件，请使用同一租户下的有效账户。",
  workspace_invitation_membership_conflict: "当前账户已有或存在冲突的工作区访问记录，请联系管理员处理。",
  workspace_invitation_admin_unavailable: "邀请管理暂时不可用，请稍后重新加载。",
  workspace_invitation_store_unavailable: "邀请存储暂时不可用，请恢复服务后重新读取状态。",
  local_identity_recent_authentication_required: "需要近期认证，请退出后重新登录，再重新审查操作。",
  local_identity_admin_scope_mismatch: "管理范围已变化，请重新选择有效的工作区。",
  local_identity_admin_unavailable: "本地身份管理暂时不可用。",
  workspace_membership_denied: "当前账户没有此工作区的有效成员资格。",
  workspace_permission_denied: "当前账户没有执行此操作的权限。",
  LOCAL_IDENTITY_AUTHENTICATION_REQUIRED: "登录已失效，请重新登录。",
  LOCAL_IDENTITY_CSRF_INVALID: "请求凭据已失效，请重新登录后重试。",
  LOCAL_IDENTITY_HTTP_DISABLED: "本地身份服务未启用。",
  LOCAL_IDENTITY_ORIGIN_FORBIDDEN: "当前页面来源不被身份服务允许。",
  LOCAL_IDENTITY_PAYLOAD_INVALID: "请求不符合邀请接口要求，请重新审查操作。",
  LOCAL_IDENTITY_SERVICE_UNAVAILABLE: "本地身份服务暂时不可用。",
};

export class WorkspaceInvitationError extends Error {
  readonly code: string;
  readonly status: number;
  constructor(code: string, status = 0) {
    super(FAILURE_GUIDANCE[code] ?? (code === "workspace_invitation_response_invalid"
      ? "邀请响应未通过校验，请重新读取状态。"
      : code === "workspace_invitation_input_invalid"
      ? "邀请输入不完整或不符合允许范围。"
      : "无法确认邀请操作结果，请恢复连接后重新读取状态。"));
    this.name = "WorkspaceInvitationError";
    this.code = code;
    this.status = status;
  }
}

export function workspaceInvitationFailure(error: unknown): WorkspaceInvitationError {
  return error instanceof WorkspaceInvitationError ? error : new WorkspaceInvitationError("workspace_invitation_unavailable");
}

export function isInvitationCode(value: string): boolean {
  return INVITATION_CODE.test(value);
}

export async function listWorkspaceInvitations(
  config: InvitationAdminConfig,
  query: { effectiveState?: InvitationEffectiveState; limit?: number; cursor?: string } = {},
  signal?: AbortSignal,
): Promise<WorkspaceInvitationPage> {
  const limit = query.limit ?? 50;
  const effectiveState = query.effectiveState ?? "pending";
  requireInput(Number.isInteger(limit) && limit >= 1 && limit <= 100 &&
    (query.effectiveState === undefined || EFFECTIVE_STATES.includes(query.effectiveState)) &&
    (query.cursor === undefined || /^[A-Za-z0-9_-]{1,8192}$/u.test(query.cursor)));
  const params = new URLSearchParams({ limit: String(limit), effective_state: effectiveState });
  if (query.cursor) params.set("cursor", query.cursor);
  const value = await requestInvitation(config, `${adminPath(config)}?${params}`, ADMIN_ROUTE, 200, undefined, signal);
  const page = exactRecord(value, ["request_id", "tenant_ref", "workspace_id", "schema_version", "as_of", "invitations"], ["next_cursor"]);
  schema(page, "workspace_invitation_page.v1");
  const scope = parseScope(page);
  requireResponse(sameScope(scope, config) && Array.isArray(page.invitations) && page.invitations.length <= limit);
  const asOf = timestamp(page.as_of);
  const invitations = (page.invitations as unknown[]).map((entry) => parseInvitation(entry, scope));
  const ids = new Set<string>();
  for (const invitation of invitations) {
    requireResponse(!ids.has(invitation.invitationId) &&
      invitation.effectiveState === effectiveState);
    if (invitation.lifecycleState === "pending") {
      requireResponse(invitation.effectiveState ===
        (Date.parse(invitation.expiresAt) <= Date.parse(asOf) ? "expired" : "pending"));
    }
    ids.add(invitation.invitationId);
  }
  const nextCursor = page.next_cursor === undefined ? "" : matching(page.next_cursor, /^[A-Za-z0-9_-]{1,8192}$/u);
  if (nextCursor) validateCursor(nextCursor, scope, effectiveState, limit, asOf);
  return { ...scope, requestId: matching(page.request_id, REFERENCE), asOf, invitations, nextCursor };
}

export async function createWorkspaceInvitation(
  config: InvitationAdminConfig,
  input: { role: LocalIdentityRoleDefinition; ttlPolicy: InvitationTTL; confirmed: true },
  signal?: AbortSignal,
): Promise<{ invitation: WorkspaceInvitation; invitationCode: string }> {
  requireInput(input.confirmed === true && INVITATION_ROLES.includes(input.role.roleKey as InvitationRole) &&
    !input.role.canManageLocalIdentity && INVITATION_TTLS.includes(input.ttlPolicy) &&
    REFERENCE.test(input.role.catalogVersion) && DIGEST.test(input.role.definitionDigest));
  const value = await requestInvitation(config, adminPath(config), ADMIN_ROUTE, 201, {
    role_key: input.role.roleKey, expected_catalog_version: input.role.catalogVersion,
    expected_role_definition_digest: input.role.definitionDigest, ttl_policy: input.ttlPolicy, confirmed: true,
  }, signal);
  const creation = exactRecord(value, ["request_id", "schema_version", "invitation", "invitation_code"]);
  schema(creation, "workspace_invitation_creation.v1");
  matching(creation.request_id, REFERENCE);
  const invitation = parseInvitation(creation.invitation, config);
  const invitationCode = matching(creation.invitation_code, INVITATION_CODE);
  requireResponse(invitation.effectiveState === "pending" && invitation.recordVersion === 1 &&
    invitation.roleKey === input.role.roleKey && invitation.ttlPolicy === input.ttlPolicy &&
    invitation.roleCatalogVersion === input.role.catalogVersion &&
    invitation.roleDefinitionDigest === input.role.definitionDigest && codeMatchesInvitation(invitationCode, invitation.invitationId));
  return { invitation, invitationCode };
}

export async function revokeWorkspaceInvitation(
  config: InvitationAdminConfig,
  input: { invitation: WorkspaceInvitation; confirmed: true },
  signal?: AbortSignal,
): Promise<WorkspaceInvitation> {
  const expected = input.invitation;
  requireInput(input.confirmed === true && sameScope(expected, config) && INVITATION_ID.test(expected.invitationId) &&
    positiveInteger(expected.recordVersion) && expected.lifecycleState === "pending");
  const value = await requestInvitation(config, `${adminPath(config)}/${expected.invitationId}/revoke`,
    `${ADMIN_ROUTE}/{invitation_id}/revoke`, 200, { expected_record_version: expected.recordVersion, confirmed: true }, signal);
  const mutation = exactRecord(value, ["request_id", "schema_version", "invitation"]);
  schema(mutation, "workspace_invitation_mutation.v1");
  matching(mutation.request_id, REFERENCE);
  const invitation = parseInvitation(mutation.invitation, config);
  requireResponse(invitation.effectiveState === "revoked" && invitation.invitationId === expected.invitationId &&
    invitation.recordVersion === expected.recordVersion + 1 && invitation.roleKey === expected.roleKey &&
    invitation.roleCatalogVersion === expected.roleCatalogVersion &&
    invitation.roleDefinitionDigest === expected.roleDefinitionDigest && invitation.expiresAt === expected.expiresAt);
  return invitation;
}

export async function previewWorkspaceInvitation(
  config: InvitationClaimantConfig,
  invitationCode: string,
  signal?: AbortSignal,
): Promise<WorkspaceInvitationPreview> {
  requireInput(isInvitationCode(invitationCode));
  const value = await requestInvitation(config, `${CLAIM_ROUTE}/preview`, `${CLAIM_ROUTE}/preview`, 200,
    { invitation_code: invitationCode }, signal);
  const preview = exactRecord(value, ["request_id", "schema_version", "invitation_id", "record_version", "tenant_ref",
    "workspace_id", "role", "effective_state", "expires_at"]);
  schema(preview, "workspace_invitation_preview.v1");
  matching(preview.request_id, REFERENCE);
  // Verified terminal invitations return the stable not_claimable error, never a successful preview.
  requireResponse(preview.effective_state === "pending" && preview.record_version === 1 && preview.tenant_ref === config.tenantRef);
  const invitationId = matching(preview.invitation_id, INVITATION_ID);
  requireResponse(codeMatchesInvitation(invitationCode, invitationId));
  const role = exactRecord(preview.role, ["role_key", "display_name", "summary", "catalog_version", "definition_digest"]);
  return {
    ...parseScope(preview), invitationId, recordVersion: version(preview.record_version),
    role: { roleKey: enumeration(role.role_key, INVITATION_ROLES), displayName: safeText(role.display_name, 120),
      summary: safeText(role.summary, 1024), catalogVersion: matching(role.catalog_version, REFERENCE),
      definitionDigest: matching(role.definition_digest, DIGEST) },
    effectiveState: enumeration(preview.effective_state, EFFECTIVE_STATES), expiresAt: timestamp(preview.expires_at),
  };
}

export async function claimWorkspaceInvitation(
  config: InvitationClaimantConfig,
  input: { invitationCode: string; preview: WorkspaceInvitationPreview; userId: string; confirmed: true },
  signal?: AbortSignal,
): Promise<WorkspaceInvitation> {
  const expected = input.preview;
  requireInput(input.confirmed === true && isInvitationCode(input.invitationCode) &&
    codeMatchesInvitation(input.invitationCode, expected.invitationId) && expected.effectiveState === "pending" &&
    positiveInteger(expected.recordVersion) && USER_ID.test(input.userId) && expected.tenantRef === config.tenantRef);
  const value = await requestInvitation(config, `${CLAIM_ROUTE}/claim`, `${CLAIM_ROUTE}/claim`, 200, {
    invitation_code: input.invitationCode, expected_record_version: expected.recordVersion, confirmed: true,
  }, signal);
  const mutation = exactRecord(value, ["request_id", "schema_version", "invitation", "membership", "role_assignment"]);
  schema(mutation, "workspace_invitation_mutation.v1");
  matching(mutation.request_id, REFERENCE);
  const invitation = parseInvitation(mutation.invitation, expected);
  requireResponse(invitation.effectiveState === "claimed" && invitation.invitationId === expected.invitationId &&
    invitation.recordVersion === expected.recordVersion + 1 && invitation.roleKey === expected.role.roleKey &&
    invitation.roleCatalogVersion === expected.role.catalogVersion &&
    invitation.roleDefinitionDigest === expected.role.definitionDigest && invitation.expiresAt === expected.expiresAt &&
    invitation.claimedByUserId === input.userId);
  const membership = exactRecord(mutation.membership, ["membership_id", "user_id", "tenant_ref", "workspace_id", "record_version"]);
  const assignment = exactRecord(mutation.role_assignment, ["assignment_id", "user_id", "tenant_ref", "workspace_id", "record_version", "role_key"]);
  requireResponse(sameScope(parseScope(membership), expected) && sameScope(parseScope(assignment), expected) &&
    membership.user_id === input.userId && assignment.user_id === input.userId && membership.record_version === 1 &&
    assignment.record_version === 1 && membership.membership_id === invitation.membershipId &&
    assignment.assignment_id === invitation.assignmentId && assignment.role_key === invitation.roleKey);
  return invitation;
}

async function requestInvitation(
  config: InvitationClaimantConfig,
  path: string,
  canonicalRoute: string,
  expectedStatus: number,
  body?: Record<string, unknown>,
  signal?: AbortSignal,
): Promise<Record<string, unknown>> {
  assertConfig(config);
  requireInput(typeof config.tenantRef === "string" && REFERENCE.test(config.tenantRef));
  const admin = canonicalRoute.startsWith("/v1/admin/");
  const scope = admin ? config as InvitationAdminConfig : null;
  if (scope) requireInput(REFERENCE.test(scope.tenantRef) && REFERENCE.test(scope.workspaceId));
  let response: Response;
  try {
    response = await fetch(`${config.baseUrl}${path}`, {
      method: body ? "POST" : "GET", credentials: "include", cache: "no-store", redirect: "error", signal,
      headers: { Accept: "application/json", "X-RadishMind-Active-Tenant": config.tenantRef,
        ...(scope ? { "X-RadishMind-Active-Workspace": scope.workspaceId } : {}),
        ...(body ? { "Content-Type": "application/json", "X-RadishMind-CSRF-Token": readLocalIdentityCSRFCookie() } : {}),
      },
      body: body ? JSON.stringify(body) : undefined,
    });
  } catch (error) {
    if (signal?.aborted) throw error;
    throw new WorkspaceInvitationError("workspace_invitation_unavailable");
  }
  requireResponse((response.headers.get("Content-Type") ?? "").toLowerCase().split(";")[0].trim() === "application/json");
  let value: unknown;
  try { value = await response.json(); } catch { throw new WorkspaceInvitationError("workspace_invitation_response_invalid", response.status); }
  requireResponse(isRecord(value));
  if (!response.ok) throw parseFailure(value as Record<string, unknown>, response.status, canonicalRoute, admin);
  requireResponse(response.status === expectedStatus);
  return value as Record<string, unknown>;
}

function parseFailure(value: unknown, status: number, route: string, admin: boolean): WorkspaceInvitationError {
  const envelope = exactRecord(value, ["error"]);
  const error = exactRecord(envelope.error, ["message", "type", "code", "request_id", "route", "failure_boundary"], ["metadata"]);
  safeText(error.message, 512);
  matching(error.type, REFERENCE);
  matching(error.request_id, REFERENCE);
  const code = matching(error.code, REFERENCE);
  requireResponse(Object.hasOwn(FAILURE_GUIDANCE, code) && error.route === route);
  if (error.failure_boundary === "local_identity") {
    requireResponse(code.startsWith("LOCAL_IDENTITY_") && error.metadata === undefined);
  } else {
    requireResponse(!code.startsWith("LOCAL_IDENTITY_") &&
      (error.failure_boundary === (admin ? "workspace_invitation_administration" : "workspace_invitation_claim") ||
        (admin && error.failure_boundary === "local_identity_administration")));
    const metadata = exactRecord(error.metadata, ["recovery"]);
    matching(metadata.recovery, REFERENCE);
  }
  // Provider/server text is never retained by the UI, including unknown error payloads.
  return new WorkspaceInvitationError(code, status);
}

function parseInvitation(value: unknown, scope: InvitationScope): WorkspaceInvitation {
  const row = exactRecord(value, ["schema_version", "invitation_id", "record_version", "tenant_ref", "workspace_id", "role_key",
    "role_catalog_version", "role_definition_digest", "ttl_policy", "lifecycle_state", "effective_state", "expires_at",
    "created_at", "updated_at", "created_by_actor_ref", "created_request_ref", "created_audit_ref", "updated_request_ref", "updated_audit_ref"],
  ["claimed_at", "claimed_by_user_id", "membership_id", "assignment_id", "revoked_at", "revoked_by_actor_ref"]);
  schema(row, "workspace_invitation.v1");
  const parsedScope = parseScope(row);
  requireResponse(sameScope(parsedScope, scope));
  for (const key of ["created_by_actor_ref", "created_request_ref", "created_audit_ref", "updated_request_ref", "updated_audit_ref"]) matching(row[key], REFERENCE);
  const lifecycleState = enumeration(row.lifecycle_state, ["pending", "claimed", "revoked"] as const);
  const effectiveState = enumeration(row.effective_state, EFFECTIVE_STATES);
  requireResponse(lifecycleState === "pending" ? effectiveState === "pending" || effectiveState === "expired" : effectiveState === lifecycleState);
  const createdAt = timestamp(row.created_at);
  const updatedAt = timestamp(row.updated_at);
  const expiresAt = timestamp(row.expires_at);
  const ttlPolicy = enumeration(row.ttl_policy, INVITATION_TTLS);
  requireResponse(Date.parse(updatedAt) >= Date.parse(createdAt) &&
    Math.abs(Date.parse(expiresAt) - Date.parse(createdAt) - TTL_MILLISECONDS[ttlPolicy]) <= 1);
  const claimedKeys = ["claimed_at", "claimed_by_user_id", "membership_id", "assignment_id"];
  requireResponse(claimedKeys.every((key) => (row[key] !== undefined) === (lifecycleState === "claimed")) &&
    (row.revoked_at !== undefined) === (lifecycleState === "revoked") &&
    (row.revoked_by_actor_ref !== undefined) === (lifecycleState === "revoked"));
  const result: WorkspaceInvitation = {
    ...parsedScope, invitationId: matching(row.invitation_id, INVITATION_ID), recordVersion: version(row.record_version),
    roleKey: enumeration(row.role_key, INVITATION_ROLES), roleCatalogVersion: matching(row.role_catalog_version, REFERENCE),
    roleDefinitionDigest: matching(row.role_definition_digest, DIGEST), ttlPolicy, lifecycleState, effectiveState,
    createdAt, updatedAt, expiresAt,
  };
  if (lifecycleState === "claimed") {
    result.claimedAt = timestamp(row.claimed_at);
    result.claimedByUserId = matching(row.claimed_by_user_id, USER_ID);
    result.membershipId = matching(row.membership_id, MEMBERSHIP_ID);
    result.assignmentId = matching(row.assignment_id, ASSIGNMENT_ID);
    requireResponse(result.claimedAt === updatedAt && Date.parse(result.claimedAt) < Date.parse(expiresAt));
  }
  if (lifecycleState === "revoked") {
    result.revokedAt = timestamp(row.revoked_at);
    matching(row.revoked_by_actor_ref, REFERENCE);
    requireResponse(result.revokedAt === updatedAt);
  }
  requireResponse(result.recordVersion === (lifecycleState === "pending" ? 1 : 2));
  return result;
}

function validateCursor(cursor: string, scope: InvitationScope, effectiveState: string, limit: number, asOf: string) {
  let decoded: unknown;
  try { decoded = JSON.parse(atob(cursor.replace(/-/gu, "+").replace(/_/gu, "/"))); } catch { requireResponse(false); }
  const value = exactRecord(decoded, ["schema_version", "tenant_ref", "workspace_id", "effective_state", "limit", "as_of", "updated_at", "invitation_id", "binding_digest"]);
  schema(value, "workspace_invitation_cursor.v1");
  requireResponse(sameScope(parseScope(value), scope) && value.effective_state === effectiveState && value.limit === limit && value.as_of === asOf);
  timestamp(value.updated_at);
  matching(value.invitation_id, INVITATION_ID);
  matching(value.binding_digest, DIGEST);
}

function assertConfig(config: LocalIdentityConsumerConfig) {
  let url: URL;
  try { url = new URL(config.baseUrl); } catch { throw new WorkspaceInvitationError("workspace_invitation_input_invalid"); }
  requireInput(config.mode === "local_identity_dev" && ["http:", "https:"].includes(url.protocol) &&
    !url.username && !url.password && !url.search && !url.hash && url.pathname === "/");
}

function adminPath(config: InvitationAdminConfig): string {
  requireInput(typeof config.tenantRef === "string" && typeof config.workspaceId === "string" &&
    REFERENCE.test(config.tenantRef) && REFERENCE.test(config.workspaceId));
  return `/v1/admin/local-identity/workspaces/${encodeURIComponent(config.workspaceId)}/invitations`;
}

function codeMatchesInvitation(code: string, invitationId: string): boolean {
  return INVITATION_CODE.exec(code)?.[1] === invitationId.slice(4);
}
function sameScope(left: InvitationScope, right: InvitationScope): boolean {
  return left.tenantRef === right.tenantRef && left.workspaceId === right.workspaceId;
}
function parseScope(value: Record<string, unknown>): InvitationScope {
  return { tenantRef: matching(value.tenant_ref, REFERENCE), workspaceId: matching(value.workspace_id, REFERENCE) };
}
function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}
function exactRecord(value: unknown, required: string[], optional: string[] = []): Record<string, unknown> {
  requireResponse(isRecord(value));
  const record = value as Record<string, unknown>;
  requireResponse(required.every((key) => Object.hasOwn(record, key)) && Object.keys(record).every((key) => required.includes(key) || optional.includes(key)));
  return record;
}
function schema(value: Record<string, unknown>, expected: string) { requireResponse(value.schema_version === expected); }
function matching(value: unknown, pattern: RegExp): string {
  requireResponse(typeof value === "string" && pattern.test(value));
  return value as string;
}
function safeText(value: unknown, maximum: number): string {
  requireResponse(typeof value === "string" && value.trim().length > 0 && value.length <= maximum && !/[\0\r\n]/u.test(value) && !SECRET_IN_TEXT.test(value));
  return value as string;
}
function timestamp(value: unknown): string {
  requireResponse(typeof value === "string" && /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{1,9})?Z$/u.test(value) && Number.isFinite(Date.parse(value)));
  return value as string;
}
function enumeration<const T extends readonly string[]>(value: unknown, options: T): T[number] {
  requireResponse(typeof value === "string" && options.includes(value));
  return value as T[number];
}
function positiveInteger(value: unknown): value is number { return Number.isSafeInteger(value) && Number(value) > 0; }
function version(value: unknown): number { requireResponse(positiveInteger(value)); return value as number; }
function requireInput(valid: boolean): asserts valid {
  if (!valid) throw new WorkspaceInvitationError("workspace_invitation_input_invalid");
}
function requireResponse(valid: boolean): asserts valid {
  if (!valid) throw new WorkspaceInvitationError("workspace_invitation_response_invalid");
}
