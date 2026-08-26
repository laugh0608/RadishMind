import {
  readLocalIdentityCSRFCookie,
  type LocalIdentityConsumerConfig,
} from "./localIdentityConsumer.ts";

const CSRF_HEADER = "X-RadishMind-CSRF-Token";
const SESSION_ID_PATTERN = /^ses_[a-z0-9]{16,64}$/u;
const SELF_SERVICE_FAILURE_BOUNDARY = "local_identity_self_service_security";
const LOCAL_IDENTITY_FAILURE_BOUNDARY = "local_identity";
const SELF_SERVICE_FAILURE_CODES = new Set([
  "local_identity_session_cursor_invalid",
  "local_identity_session_scope_denied",
  "local_identity_session_version_conflict",
  "local_identity_session_recent_authentication_required",
  "local_identity_session_bulk_revoke_conflict",
  "local_identity_credential_unavailable",
  "local_identity_credential_current_invalid",
  "local_identity_credential_policy_rejected",
  "local_identity_credential_reuse_denied",
  "local_identity_credential_rotation_conflict",
  "local_identity_service_unavailable",
]);
const LOCAL_IDENTITY_TRANSPORT_FAILURE_CODES = new Set([
  "LOCAL_IDENTITY_AUTHENTICATION_REQUIRED",
  "LOCAL_IDENTITY_CSRF_INVALID",
  "LOCAL_IDENTITY_ORIGIN_FORBIDDEN",
  "LOCAL_IDENTITY_PAYLOAD_INVALID",
  "LOCAL_IDENTITY_HTTP_DISABLED",
  "LOCAL_IDENTITY_SERVICE_UNAVAILABLE",
]);
const FORBIDDEN_RESPONSE_KEYS = new Set([
  "access_token",
  "audit_ref",
  "authentication_source_ref",
  "cookie",
  "credential_digest",
  "credential_id",
  "id_token",
  "ip",
  "issuer",
  "login_identifier",
  "password",
  "raw_claims",
  "refresh_token",
  "subject",
  "user_agent",
  "user_id",
]);

export type LocalIdentitySelfServiceSessionState = "active" | "expired" | "revoked";
export type LocalIdentitySelfServiceSessionFilter = LocalIdentitySelfServiceSessionState | "all";

export type LocalIdentitySelfServiceSessionSummary = {
  schemaVersion: "local_identity_self_service_session_summary.v1";
  sessionId: string;
  authenticationMethod: "local_password" | "oidc";
  effectiveState: LocalIdentitySelfServiceSessionState;
  recordVersion: number;
  currentSession: boolean;
  createdAt: string;
  lastVerifiedAt: string;
  expiresAt: string;
  revokedAt?: string;
};

export type LocalIdentitySelfServiceSessionPage = {
  sessions: LocalIdentitySelfServiceSessionSummary[];
  snapshotAt: string;
  nextCursor?: string;
};

export type LocalIdentitySelfServiceSessionRevocation = {
  schemaVersion: "local_identity_self_service_session_revocation.v1";
  session: LocalIdentitySelfServiceSessionSummary;
  currentSessionRevoked: boolean;
};

export type LocalIdentitySelfServiceBulkSessionRevocation = {
  schemaVersion: "local_identity_self_service_bulk_session_revocation.v1";
  revokedCount: number;
};

export type LocalIdentitySelfServiceCredentialRotation = {
  schemaVersion: "local_identity_self_service_credential_rotation.v1";
  policyVersion: string;
  revokedSessionCount: number;
  currentSessionRevoked: boolean;
};

export type LocalIdentitySelfServiceSecurityFailureKind =
  | "authentication_required"
  | "denied"
  | "recent_authentication"
  | "conflict"
  | "credential_unavailable"
  | "credential_invalid"
  | "credential_policy"
  | "unavailable"
  | "invalid_response"
  | "failed";

export class LocalIdentitySelfServiceSecurityError extends Error {
  readonly status: number;
  readonly code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = "LocalIdentitySelfServiceSecurityError";
    this.status = status;
    this.code = code;
  }
}

export function localIdentitySelfServiceSecurityFailureKind(
  error: unknown,
): LocalIdentitySelfServiceSecurityFailureKind {
  if (!(error instanceof LocalIdentitySelfServiceSecurityError)) return "failed";
  switch (error.code) {
    case "LOCAL_IDENTITY_AUTHENTICATION_REQUIRED":
      return "authentication_required";
    case "local_identity_session_scope_denied":
    case "LOCAL_IDENTITY_CSRF_INVALID":
    case "LOCAL_IDENTITY_ORIGIN_FORBIDDEN":
      return "denied";
    case "local_identity_session_recent_authentication_required":
      return "recent_authentication";
    case "local_identity_session_cursor_invalid":
    case "local_identity_session_version_conflict":
    case "local_identity_session_bulk_revoke_conflict":
    case "local_identity_credential_rotation_conflict":
      return "conflict";
    case "local_identity_credential_unavailable":
      return "credential_unavailable";
    case "local_identity_credential_current_invalid":
    case "local_identity_credential_reuse_denied":
      return "credential_invalid";
    case "local_identity_credential_policy_rejected":
    case "LOCAL_IDENTITY_PAYLOAD_INVALID":
      return "credential_policy";
    case "local_identity_service_unavailable":
    case "LOCAL_IDENTITY_HTTP_DISABLED":
    case "LOCAL_IDENTITY_SERVICE_UNAVAILABLE":
      return "unavailable";
    case "local_identity_response_invalid":
      return "invalid_response";
    default:
      return "failed";
  }
}

export async function listLocalIdentitySelfServiceSessions(
  config: LocalIdentityConsumerConfig,
  query: {
    state: LocalIdentitySelfServiceSessionFilter;
    limit?: number;
    cursor?: string;
  },
  signal?: AbortSignal,
): Promise<LocalIdentitySelfServiceSessionPage> {
  requireEnabled(config);
  const limit = query.limit ?? 50;
  if (!isSessionFilter(query.state) || !Number.isInteger(limit) || limit < 1 || limit > 100 ||
    (query.cursor !== undefined && !isCursor(query.cursor))) {
    throw inputError("Session directory query is invalid.");
  }
  const search = new URLSearchParams({ state: query.state, limit: String(limit) });
  if (query.cursor) search.set("cursor", query.cursor);
  const response = await fetch(`${config.baseUrl}/v1/auth/sessions?${search.toString()}`, {
    method: "GET",
    credentials: "include",
    cache: "no-store",
    headers: { Accept: "application/json" },
    signal,
  });
  const body = await readResponseJSON(response);
  if (!response.ok) throw errorFromResponse(response.status, body);
  return parseSessionPage(body);
}

export async function revokeLocalIdentitySelfServiceSession(
  config: LocalIdentityConsumerConfig,
  input: { sessionId: string; expectedRecordVersion: number },
  signal?: AbortSignal,
): Promise<LocalIdentitySelfServiceSessionRevocation> {
  if (!SESSION_ID_PATTERN.test(input.sessionId) || !isPositiveInteger(input.expectedRecordVersion)) {
    throw inputError("Exact session revocation input is invalid.");
  }
  const body = await selfServiceMutation(config, `/v1/auth/sessions/${encodeURIComponent(input.sessionId)}/revoke`, {
    expected_record_version: input.expectedRecordVersion,
    confirmed: true,
  }, signal);
  return parseSessionRevocation(body);
}

export async function revokeOtherLocalIdentitySelfServiceSessions(
  config: LocalIdentityConsumerConfig,
  signal?: AbortSignal,
): Promise<LocalIdentitySelfServiceBulkSessionRevocation> {
  return parseBulkSessionRevocation(await selfServiceMutation(
    config,
    "/v1/auth/sessions/revoke-others",
    { confirmed: true },
    signal,
  ));
}

export async function rotateLocalIdentitySelfServiceCredential(
  config: LocalIdentityConsumerConfig,
  input: { currentPassword: string; newPassword: string },
  signal?: AbortSignal,
): Promise<LocalIdentitySelfServiceCredentialRotation> {
  if (input.currentPassword.length < 1 || input.currentPassword.length > 1024 ||
    input.newPassword.length < 12 || input.newPassword.length > 1024) {
    throw inputError("Credential rotation input is invalid.");
  }
  return parseCredentialRotation(await selfServiceMutation(
    config,
    "/v1/auth/local/credential/rotate",
    {
      current_password: input.currentPassword,
      new_password: input.newPassword,
      session_impact_confirmed: true,
    },
    signal,
  ));
}

async function selfServiceMutation(
  config: LocalIdentityConsumerConfig,
  path: string,
  body: unknown,
  signal?: AbortSignal,
): Promise<unknown> {
  requireEnabled(config);
  const response = await fetch(`${config.baseUrl}${path}`, {
    method: "POST",
    credentials: "include",
    cache: "no-store",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
      [CSRF_HEADER]: readLocalIdentityCSRFCookie(),
    },
    body: JSON.stringify(body),
    signal,
  });
  const value = await readResponseJSON(response);
  if (!response.ok) throw errorFromResponse(response.status, value);
  return value;
}

function parseSessionPage(value: unknown): LocalIdentitySelfServiceSessionPage {
  if (containsForbiddenResponse(value) || !isRecord(value) ||
    !hasExactKeys(value, ["sessions", "snapshot_at"], ["next_cursor"]) ||
    !Array.isArray(value.sessions) || value.sessions.length > 100 || !isValidDate(value.snapshot_at) ||
    (value.next_cursor !== undefined && !isCursor(value.next_cursor))) {
    throw invalidResponse();
  }
  const sessions = value.sessions.map(parseSessionSummary);
  if (new Set(sessions.map((session) => session.sessionId)).size !== sessions.length ||
    sessions.filter((session) => session.currentSession).length > 1) {
    throw invalidResponse();
  }
  return {
    sessions,
    snapshotAt: value.snapshot_at,
    ...(value.next_cursor === undefined ? {} : { nextCursor: value.next_cursor }),
  };
}

function parseSessionSummary(value: unknown): LocalIdentitySelfServiceSessionSummary {
  if (!isRecord(value) || !hasExactKeys(value, [
    "schema_version", "session_id", "authentication_method", "effective_state", "record_version",
    "current_session", "created_at", "last_verified_at", "expires_at",
  ], ["revoked_at"]) || value.schema_version !== "local_identity_self_service_session_summary.v1" ||
    !SESSION_ID_PATTERN.test(String(value.session_id)) ||
    (value.authentication_method !== "local_password" && value.authentication_method !== "oidc") ||
    !isSessionState(value.effective_state) || !isPositiveInteger(value.record_version) ||
    typeof value.current_session !== "boolean" || !isValidDate(value.created_at) ||
    !isValidDate(value.last_verified_at) || !isValidDate(value.expires_at) ||
    (value.revoked_at !== undefined && !isValidDate(value.revoked_at)) ||
    (value.effective_state === "revoked") !== (value.revoked_at !== undefined)) {
    throw invalidResponse();
  }
  return {
    schemaVersion: value.schema_version,
    sessionId: String(value.session_id),
    authenticationMethod: value.authentication_method,
    effectiveState: value.effective_state,
    recordVersion: value.record_version,
    currentSession: value.current_session,
    createdAt: value.created_at,
    lastVerifiedAt: value.last_verified_at,
    expiresAt: value.expires_at,
    ...(value.revoked_at === undefined ? {} : { revokedAt: value.revoked_at }),
  };
}

function parseSessionRevocation(value: unknown): LocalIdentitySelfServiceSessionRevocation {
  if (containsForbiddenResponse(value) || !isRecord(value) || !hasExactKeys(value, [
    "schema_version", "session", "current_session_revoked",
  ]) || value.schema_version !== "local_identity_self_service_session_revocation.v1" ||
    typeof value.current_session_revoked !== "boolean") {
    throw invalidResponse();
  }
  const session = parseSessionSummary(value.session);
  if (session.effectiveState !== "revoked" || session.currentSession !== value.current_session_revoked) {
    throw invalidResponse();
  }
  return {
    schemaVersion: value.schema_version,
    session,
    currentSessionRevoked: value.current_session_revoked,
  };
}

function parseBulkSessionRevocation(value: unknown): LocalIdentitySelfServiceBulkSessionRevocation {
  if (containsForbiddenResponse(value) || !isRecord(value) || !hasExactKeys(value, [
    "schema_version", "revoked_count",
  ]) || value.schema_version !== "local_identity_self_service_bulk_session_revocation.v1" ||
    !isNonNegativeInteger(value.revoked_count)) {
    throw invalidResponse();
  }
  return { schemaVersion: value.schema_version, revokedCount: value.revoked_count };
}

function parseCredentialRotation(value: unknown): LocalIdentitySelfServiceCredentialRotation {
  if (containsForbiddenResponse(value) || !isRecord(value) || !hasExactKeys(value, [
    "schema_version", "policy_version", "revoked_session_count", "current_session_revoked",
  ]) || value.schema_version !== "local_identity_self_service_credential_rotation.v1" ||
    !isReference(value.policy_version) || !isNonNegativeInteger(value.revoked_session_count) ||
    typeof value.current_session_revoked !== "boolean" ||
    (value.current_session_revoked && value.revoked_session_count < 1)) {
    throw invalidResponse();
  }
  return {
    schemaVersion: value.schema_version,
    policyVersion: value.policy_version,
    revokedSessionCount: value.revoked_session_count,
    currentSessionRevoked: value.current_session_revoked,
  };
}

async function readResponseJSON(response: Response): Promise<unknown> {
  if (!(response.headers.get("Content-Type") ?? "").toLowerCase().includes("application/json")) {
    throw invalidResponse(response.status);
  }
  try {
    return await response.json() as unknown;
  } catch {
    throw invalidResponse(response.status);
  }
}

function errorFromResponse(status: number, value: unknown): LocalIdentitySelfServiceSecurityError {
  if (containsForbiddenResponse(value) || !isRecord(value) || !hasExactKeys(value, ["error"]) ||
    !isRecord(value.error) || !hasExactKeys(value.error, [
      "message", "type", "code", "request_id", "route", "failure_boundary",
    ]) || typeof value.error.message !== "string" || value.error.message.length > 512 ||
    typeof value.error.type !== "string" || !/^[a-z][a-z0-9_]{2,63}$/u.test(value.error.type) ||
    typeof value.error.code !== "string" || typeof value.error.request_id !== "string" ||
    !isReference(value.error.request_id) || typeof value.error.route !== "string" ||
    !value.error.route.startsWith("/v1/auth/")) {
    return invalidResponse(status);
  }
  const selfServiceFailure = value.error.failure_boundary === SELF_SERVICE_FAILURE_BOUNDARY &&
    SELF_SERVICE_FAILURE_CODES.has(value.error.code);
  const transportFailure = value.error.failure_boundary === LOCAL_IDENTITY_FAILURE_BOUNDARY &&
    LOCAL_IDENTITY_TRANSPORT_FAILURE_CODES.has(value.error.code);
  if (!selfServiceFailure && !transportFailure) return invalidResponse(status);
  return new LocalIdentitySelfServiceSecurityError(status, value.error.code, value.error.message);
}

function requireEnabled(config: LocalIdentityConsumerConfig): void {
  if (config.mode !== "local_identity_dev") {
    throw new LocalIdentitySelfServiceSecurityError(
      0,
      "LOCAL_IDENTITY_HTTP_DISABLED",
      "The local identity self-service security consumer is disabled.",
    );
  }
}

function inputError(message: string): LocalIdentitySelfServiceSecurityError {
  return new LocalIdentitySelfServiceSecurityError(0, "local_identity_input_invalid", message);
}

function invalidResponse(status = 0): LocalIdentitySelfServiceSecurityError {
  return new LocalIdentitySelfServiceSecurityError(
    status,
    "local_identity_response_invalid",
    "The local identity self-service security response is invalid.",
  );
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

function isSessionState(value: unknown): value is LocalIdentitySelfServiceSessionState {
  return value === "active" || value === "expired" || value === "revoked";
}

function isSessionFilter(value: unknown): value is LocalIdentitySelfServiceSessionFilter {
  return value === "all" || isSessionState(value);
}

function isPositiveInteger(value: unknown): value is number {
  return Number.isInteger(value) && Number(value) > 0;
}

function isNonNegativeInteger(value: unknown): value is number {
  return Number.isInteger(value) && Number(value) >= 0;
}

function isValidDate(value: unknown): value is string {
  return typeof value === "string" && value.length <= 64 && !Number.isNaN(Date.parse(value));
}

function isCursor(value: unknown): value is string {
  return typeof value === "string" && value.length >= 16 && value.length <= 8192 &&
    /^[A-Za-z0-9_-]+$/u.test(value);
}

function isReference(value: unknown): value is string {
  return typeof value === "string" && /^[A-Za-z0-9][A-Za-z0-9_.:/-]{0,159}$/u.test(value);
}
