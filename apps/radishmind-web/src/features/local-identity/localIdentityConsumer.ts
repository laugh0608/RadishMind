const DEFAULT_LOCAL_IDENTITY_BASE_URL = "http://127.0.0.1:7000";
const LOCAL_IDENTITY_MODE = "local_identity_dev";
const BOOTSTRAP_CSRF = "bootstrap";
const CSRF_HEADER = "X-RadishMind-CSRF-Token";
const FORBIDDEN_RESPONSE_KEYS = new Set([
  "access_token",
  "audit_ref",
  "cookie",
  "credential_digest",
  "id_token",
  "issuer",
  "login_identifier",
  "password",
  "raw_claims",
  "refresh_token",
  "subject",
]);

export type LocalIdentityConsumerMode = "disabled" | "local_identity_dev";

export type LocalIdentityConsumerConfig = {
  mode: LocalIdentityConsumerMode;
  baseUrl: string;
};

export type LocalIdentityAccount = {
  userId: string;
  displayName: string;
  lifecycleState: "active" | "disabled";
};

export type LocalIdentitySession = {
  sessionId: string;
  authenticationMethod: "local_password" | "oidc";
  expiresAt: string;
};

export type LocalIdentityAuthentication = {
  account: LocalIdentityAccount;
  session: LocalIdentitySession;
  returnTo: string;
};

export type LocalIdentityExternalIdentity = {
  bindingId: string;
  providerRef: "radish_oidc";
  lifecycleState: "active" | "revoked";
  recordVersion: number;
  createdAt: string;
  updatedAt: string;
  revokedAt?: string;
  canRevoke: boolean;
};

export type LocalIdentityRoleAssignment = {
  assignmentId: string;
  tenantRef: string;
  workspaceId?: string;
  roleKey: string;
  permissionGrants: string[];
  lifecycleState: "active" | "revoked";
  recordVersion: number;
  expiresAt?: string;
};

export type LocalIdentityWorkspaceMembership = {
  membershipId: string;
  tenantRef: string;
  workspaceId: string;
  lifecycleState: "active" | "revoked";
  recordVersion: number;
  expiresAt?: string;
};

export type LocalIdentityAccountProfile = {
  account: LocalIdentityAccount;
  session: LocalIdentitySession;
  externalIdentities: LocalIdentityExternalIdentity[];
  roleAssignments: LocalIdentityRoleAssignment[];
  workspaceMemberships: LocalIdentityWorkspaceMembership[];
  capabilities: {
    oidcEnabled: boolean;
    recentAuthentication: boolean;
    hasActiveLocalCredential: boolean;
  };
};

export class LocalIdentityConsumerError extends Error {
  readonly status: number;
  readonly code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = "LocalIdentityConsumerError";
    this.status = status;
    this.code = code;
  }
}

export function readLocalIdentityConsumerConfig(): LocalIdentityConsumerConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  return localIdentityConsumerConfigFromEnvironment(env);
}

export function localIdentityConsumerConfigFromEnvironment(
  env: Record<string, string | undefined>,
): LocalIdentityConsumerConfig {
  return {
    mode: env.VITE_RADISHMIND_LOCAL_IDENTITY_MODE?.trim() === LOCAL_IDENTITY_MODE
      ? "local_identity_dev"
      : "disabled",
    baseUrl: normalizeBaseUrl(
      env.VITE_RADISHMIND_LOCAL_IDENTITY_BASE_URL ??
      env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ??
      DEFAULT_LOCAL_IDENTITY_BASE_URL,
    ),
  };
}

export function localIdentityReturnTarget(location: { pathname: string; search: string }): string {
  const target = `${location.pathname}${location.search}`;
  return target.startsWith("/") && !target.startsWith("//") ? target : "/";
}

export async function probeLocalIdentitySession(
  config: LocalIdentityConsumerConfig,
  signal?: AbortSignal,
): Promise<LocalIdentityAuthentication | null> {
  requireEnabled(config);
  const response = await fetch(`${config.baseUrl}/v1/auth/session`, {
    method: "GET",
    credentials: "include",
    cache: "no-store",
    headers: { Accept: "application/json" },
    signal,
  });
  if (response.status === 401) return null;
  const body = await readResponseJSON(response);
  if (!response.ok) throw localIdentityErrorFromResponse(response.status, body);
  return parseAuthentication(body);
}

export async function readLocalIdentityAccountProfile(
  config: LocalIdentityConsumerConfig,
  signal?: AbortSignal,
): Promise<LocalIdentityAccountProfile> {
  requireEnabled(config);
  const response = await fetch(`${config.baseUrl}/v1/auth/account`, {
    method: "GET",
    credentials: "include",
    cache: "no-store",
    headers: { Accept: "application/json" },
    signal,
  });
  const body = await readResponseJSON(response);
  if (!response.ok) throw localIdentityErrorFromResponse(response.status, body);
  return parseAccountProfile(body);
}

export async function authenticateLocalIdentity(
  config: LocalIdentityConsumerConfig,
  input: {
    intent: "login" | "register";
    loginIdentifier: string;
    password: string;
    displayName?: string;
    returnTo?: string;
  },
): Promise<LocalIdentityAuthentication> {
  requireEnabled(config);
  const path = input.intent === "register" ? "/v1/auth/local/register" : "/v1/auth/local/login";
  const payload = input.intent === "register"
    ? {
        login_identifier: input.loginIdentifier,
        display_name: input.displayName ?? "",
        password: input.password,
        return_to: input.returnTo ?? "/",
      }
    : {
        login_identifier: input.loginIdentifier,
        password: input.password,
        return_to: input.returnTo ?? "/",
      };
  const response = await localIdentityJSONRequest(config, path, payload, BOOTSTRAP_CSRF);
  const body = await readResponseJSON(response);
  if (!response.ok) throw localIdentityErrorFromResponse(response.status, body);
  return parseAuthentication(body);
}

export async function startLocalIdentityOIDC(
  config: LocalIdentityConsumerConfig,
  intent: "login" | "link",
  returnTo = "/",
): Promise<{ authorizationUrl: string; expiresAt: string }> {
  requireEnabled(config);
  const csrf = intent === "login" ? BOOTSTRAP_CSRF : readLocalIdentityCSRFCookie();
  const response = await localIdentityJSONRequest(config, "/v1/auth/oidc/start", {
    intent,
    return_to: returnTo,
  }, csrf);
  const body = await readResponseJSON(response);
  if (!response.ok) throw localIdentityErrorFromResponse(response.status, body);
  if (containsForbiddenResponse(body) || !isRecord(body) ||
    !hasExactKeys(body, ["authorization_url", "expires_at"]) ||
    typeof body.authorization_url !== "string" || !isValidDate(body.expires_at)) {
    throw invalidResponse();
  }
  const authorizationURL = new URL(body.authorization_url);
  if (authorizationURL.protocol !== "https:" &&
    !(authorizationURL.protocol === "http:" && isLoopbackHost(authorizationURL.hostname))) {
    throw invalidResponse();
  }
  return { authorizationUrl: authorizationURL.toString(), expiresAt: body.expires_at as string };
}

export async function logoutLocalIdentity(config: LocalIdentityConsumerConfig): Promise<void> {
  requireEnabled(config);
  const response = await localIdentityJSONRequest(
    config,
    "/v1/auth/logout",
    {},
    readLocalIdentityCSRFCookie(),
  );
  if (response.status === 204) return;
  throw localIdentityErrorFromResponse(response.status, await readResponseJSON(response));
}

export async function revokeLocalIdentityExternalIdentity(
  config: LocalIdentityConsumerConfig,
  bindingId: string,
  expectedRecordVersion: number,
): Promise<void> {
  requireEnabled(config);
  if (!/^xid_[a-z0-9]{16,64}$/u.test(bindingId) ||
    !Number.isInteger(expectedRecordVersion) || expectedRecordVersion < 1) {
    throw new LocalIdentityConsumerError(0, "local_identity_input_invalid", "External identity input is invalid.");
  }
  const response = await localIdentityJSONRequest(
    config,
    `/v1/auth/external-identities/${encodeURIComponent(bindingId)}/revoke`,
    { expected_record_version: expectedRecordVersion },
    readLocalIdentityCSRFCookie(),
  );
  if (response.status === 204) return;
  throw localIdentityErrorFromResponse(response.status, await readResponseJSON(response));
}

export function readLocalIdentityCSRFCookie(cookieSource = globalThis.document?.cookie ?? ""): string {
  const values = cookieSource.split(";").map((entry) => entry.trim()).filter(Boolean);
  const candidates = values.flatMap((entry) => {
    const separator = entry.indexOf("=");
    if (separator <= 0) return [];
    const name = entry.slice(0, separator);
    if (name !== "radishmind_csrf_dev" && name !== "__Host-radishmind_csrf") return [];
    return [decodeURIComponent(entry.slice(separator + 1))];
  }).filter(Boolean);
  if (candidates.length !== 1) {
    throw new LocalIdentityConsumerError(0, "local_identity_csrf_unavailable", "The local session CSRF proof is unavailable.");
  }
  return candidates[0];
}

async function localIdentityJSONRequest(
  config: LocalIdentityConsumerConfig,
  path: string,
  body: unknown,
  csrf: string,
): Promise<Response> {
  return fetch(`${config.baseUrl}${path}`, {
    method: "POST",
    credentials: "include",
    cache: "no-store",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
      [CSRF_HEADER]: csrf,
    },
    body: JSON.stringify(body),
  });
}

function parseAuthentication(value: unknown): LocalIdentityAuthentication {
  if (containsForbiddenResponse(value) || !isRecord(value) ||
    !(hasExactKeys(value, ["account", "session"]) || hasExactKeys(value, ["account", "session", "return_to"]))) {
    throw invalidResponse();
  }
  return {
    account: parseAccount(value.account),
    session: parseSession(value.session),
    returnTo: typeof value.return_to === "string" && value.return_to.startsWith("/") ? value.return_to : "/",
  };
}

function parseAccountProfile(value: unknown): LocalIdentityAccountProfile {
  if (containsForbiddenResponse(value) || !isRecord(value) || !hasExactKeys(value, [
    "account", "session", "external_identities", "role_assignments", "workspace_memberships", "capabilities",
  ]) || !Array.isArray(value.external_identities) || !Array.isArray(value.role_assignments) ||
    !Array.isArray(value.workspace_memberships) || !isRecord(value.capabilities) ||
    !hasExactKeys(value.capabilities, ["oidc_enabled", "recent_authentication", "has_active_local_credential"]) ||
    typeof value.capabilities.oidc_enabled !== "boolean" ||
    typeof value.capabilities.recent_authentication !== "boolean" ||
    typeof value.capabilities.has_active_local_credential !== "boolean") {
    throw invalidResponse();
  }
  return {
    account: parseAccount(value.account),
    session: parseSession(value.session),
    externalIdentities: value.external_identities.map(parseExternalIdentity),
    roleAssignments: value.role_assignments.map(parseRoleAssignment),
    workspaceMemberships: value.workspace_memberships.map(parseWorkspaceMembership),
    capabilities: {
      oidcEnabled: value.capabilities.oidc_enabled,
      recentAuthentication: value.capabilities.recent_authentication,
      hasActiveLocalCredential: value.capabilities.has_active_local_credential,
    },
  };
}

function parseAccount(value: unknown): LocalIdentityAccount {
  if (!isRecord(value) || !hasExactKeys(value, ["user_id", "display_name", "lifecycle_state"]) ||
    !/^usr_[a-z0-9]{16,64}$/u.test(String(value.user_id)) || typeof value.display_name !== "string" ||
    (value.lifecycle_state !== "active" && value.lifecycle_state !== "disabled")) throw invalidResponse();
  return { userId: String(value.user_id), displayName: value.display_name, lifecycleState: value.lifecycle_state };
}

function parseSession(value: unknown): LocalIdentitySession {
  if (!isRecord(value) || !hasExactKeys(value, ["session_id", "authentication_method", "expires_at"]) ||
    !/^ses_[a-z0-9]{16,64}$/u.test(String(value.session_id)) ||
    (value.authentication_method !== "local_password" && value.authentication_method !== "oidc") ||
    !isValidDate(value.expires_at)) throw invalidResponse();
  return {
    sessionId: String(value.session_id),
    authenticationMethod: value.authentication_method,
    expiresAt: value.expires_at as string,
  };
}

function parseExternalIdentity(value: unknown): LocalIdentityExternalIdentity {
  if (!isRecord(value) || !hasExactKeys(value, [
    "binding_id", "provider_ref", "lifecycle_state", "record_version", "created_at", "updated_at", "can_revoke",
  ], ["revoked_at"]) || !/^xid_[a-z0-9]{16,64}$/u.test(String(value.binding_id)) ||
    value.provider_ref !== "radish_oidc" || (value.lifecycle_state !== "active" && value.lifecycle_state !== "revoked") ||
    !isPositiveInteger(value.record_version) || !isValidDate(value.created_at) || !isValidDate(value.updated_at) ||
    (value.revoked_at !== undefined && !isValidDate(value.revoked_at)) || typeof value.can_revoke !== "boolean") {
    throw invalidResponse();
  }
  return {
    bindingId: String(value.binding_id), providerRef: value.provider_ref, lifecycleState: value.lifecycle_state,
    recordVersion: value.record_version, createdAt: value.created_at as string, updatedAt: value.updated_at as string,
    ...(value.revoked_at === undefined ? {} : { revokedAt: value.revoked_at as string }), canRevoke: value.can_revoke,
  };
}

function parseRoleAssignment(value: unknown): LocalIdentityRoleAssignment {
  if (!isRecord(value) || !hasExactKeys(value, [
    "assignment_id", "tenant_ref", "role_key", "permission_grants", "lifecycle_state", "record_version",
  ], ["workspace_id", "expires_at"]) || !/^rla_[a-z0-9]{16,64}$/u.test(String(value.assignment_id)) ||
    !isReference(value.tenant_ref) || (value.workspace_id !== undefined && !isReference(value.workspace_id)) ||
    !/^[a-z][a-z0-9_]{2,63}$/u.test(String(value.role_key)) || !isStringArray(value.permission_grants) ||
    (value.lifecycle_state !== "active" && value.lifecycle_state !== "revoked") ||
    !isPositiveInteger(value.record_version) || (value.expires_at !== undefined && !isValidDate(value.expires_at))) {
    throw invalidResponse();
  }
  return {
    assignmentId: String(value.assignment_id), tenantRef: String(value.tenant_ref),
    ...(value.workspace_id === undefined ? {} : { workspaceId: String(value.workspace_id) }),
    roleKey: String(value.role_key), permissionGrants: [...value.permission_grants as string[]],
    lifecycleState: value.lifecycle_state, recordVersion: value.record_version,
    ...(value.expires_at === undefined ? {} : { expiresAt: value.expires_at as string }),
  };
}

function parseWorkspaceMembership(value: unknown): LocalIdentityWorkspaceMembership {
  if (!isRecord(value) || !hasExactKeys(value, [
    "membership_id", "tenant_ref", "workspace_id", "lifecycle_state", "record_version",
  ], ["expires_at"]) || !/^mbr_[a-z0-9]{16,64}$/u.test(String(value.membership_id)) ||
    !isReference(value.tenant_ref) || !isReference(value.workspace_id) ||
    (value.lifecycle_state !== "active" && value.lifecycle_state !== "revoked") ||
    !isPositiveInteger(value.record_version) || (value.expires_at !== undefined && !isValidDate(value.expires_at))) {
    throw invalidResponse();
  }
  return {
    membershipId: String(value.membership_id), tenantRef: String(value.tenant_ref),
    workspaceId: String(value.workspace_id), lifecycleState: value.lifecycle_state,
    recordVersion: value.record_version,
    ...(value.expires_at === undefined ? {} : { expiresAt: value.expires_at as string }),
  };
}

async function readResponseJSON(response: Response): Promise<unknown> {
  const contentType = response.headers.get("Content-Type") ?? "";
  if (!contentType.toLowerCase().includes("application/json")) throw invalidResponse();
  try {
    return await response.json() as unknown;
  } catch {
    throw invalidResponse();
  }
}

function localIdentityErrorFromResponse(status: number, value: unknown): LocalIdentityConsumerError {
  if (containsForbiddenResponse(value) || !isRecord(value) || !hasExactKeys(value, ["error"]) ||
    !isRecord(value.error) || !hasExactKeys(value.error, [
      "message", "type", "code", "request_id", "route", "failure_boundary",
    ]) || typeof value.error.message !== "string" || value.error.message.length > 512 ||
    typeof value.error.type !== "string" || !/^[a-z][a-z0-9_]{2,63}$/u.test(value.error.type) ||
    typeof value.error.code !== "string" || !/^LOCAL_IDENTITY_[A-Z0-9_]{3,96}$/u.test(value.error.code) ||
    typeof value.error.request_id !== "string" || !isReference(value.error.request_id) ||
    typeof value.error.route !== "string" || !value.error.route.startsWith("/v1/auth/") ||
    value.error.failure_boundary !== "local_identity") {
    return invalidResponse(status);
  }
  return new LocalIdentityConsumerError(status, value.error.code, value.error.message);
}

function invalidResponse(status = 0): LocalIdentityConsumerError {
  return new LocalIdentityConsumerError(status, "local_identity_response_invalid", "The local identity response is invalid.");
}

function requireEnabled(config: LocalIdentityConsumerConfig): void {
  if (config.mode !== "local_identity_dev") {
    throw new LocalIdentityConsumerError(0, "local_identity_http_disabled", "The local identity Web consumer is disabled.");
  }
}

function containsForbiddenResponse(value: unknown): boolean {
  if (Array.isArray(value)) return value.some(containsForbiddenResponse);
  if (!isRecord(value)) return false;
  return Object.entries(value).some(([key, child]) =>
    FORBIDDEN_RESPONSE_KEYS.has(key.toLowerCase()) || containsForbiddenResponse(child)
  );
}

function hasExactKeys(
  value: Record<string, unknown>,
  required: string[],
  optional: string[] = [],
): boolean {
  const keys = Object.keys(value);
  return required.every((key) => keys.includes(key)) &&
    keys.every((key) => required.includes(key) || optional.includes(key));
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function isPositiveInteger(value: unknown): value is number {
  return Number.isInteger(value) && Number(value) > 0;
}

function isValidDate(value: unknown): value is string {
  return typeof value === "string" && value.length <= 64 && !Number.isNaN(Date.parse(value));
}

function isReference(value: unknown): value is string {
  return typeof value === "string" && /^[A-Za-z0-9][A-Za-z0-9_.:/-]{0,159}$/u.test(value);
}

function isStringArray(value: unknown): value is string[] {
  return Array.isArray(value) && value.length <= 128 && value.every((entry) => isReference(entry));
}

function isLoopbackHost(hostname: string): boolean {
  return hostname === "localhost" || hostname === "127.0.0.1" || hostname === "[::1]" || hostname === "::1";
}

function normalizeBaseUrl(raw: string): string {
  const value = raw.trim().replace(/\/+$/u, "");
  try {
    const url = new URL(value);
    if ((url.protocol !== "http:" && url.protocol !== "https:") || url.username || url.password ||
      url.search || url.hash || (url.protocol === "http:" && !isLoopbackHost(url.hostname))) {
      return DEFAULT_LOCAL_IDENTITY_BASE_URL;
    }
    return url.toString().replace(/\/+$/u, "");
  } catch {
    return DEFAULT_LOCAL_IDENTITY_BASE_URL;
  }
}
