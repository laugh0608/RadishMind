import type { APIKeyRecord, APIKeyScope } from "./apiKeyLifecycleConsumer.ts";

export type APIKeyRotationPhase = "replacement_pending" | "verification_pending" | "verified";

export type APIKeyRotationSession = {
  applicationId: string;
  ownerSubjectRef: string;
  sourceApiKeyId: string;
  sourceRecordVersion: number;
  sourceDisplayName: string;
  scopes: APIKeyScope[];
  replacementApiKeyId: string;
  replacementCreatedAt: string;
  replacementLastUsedAt: string | null;
  phase: APIKeyRotationPhase;
};

export class APIKeyRotationSessionError extends Error {
  readonly failureCode: string;

  constructor(failureCode: string) {
    super(failureCode);
    this.name = "APIKeyRotationSessionError";
    this.failureCode = failureCode;
  }
}

let activeSession: APIKeyRotationSession | null = null;

export function beginAPIKeyRotationSession(source: APIKeyRecord): APIKeyRotationSession {
  if (source.lifecycleState !== "active" || source.effectiveState !== "active") {
    throw new APIKeyRotationSessionError("api_key_rotation_source_inactive");
  }
  if (activeSession) {
    throw new APIKeyRotationSessionError("api_key_rotation_session_active");
  }
  activeSession = {
    applicationId: source.applicationId,
    ownerSubjectRef: source.ownerSubjectRef,
    sourceApiKeyId: source.apiKeyId,
    sourceRecordVersion: source.recordVersion,
    sourceDisplayName: source.displayName,
    scopes: normalizeScopes(source.scopes),
    replacementApiKeyId: "",
    replacementCreatedAt: "",
    replacementLastUsedAt: null,
    phase: "replacement_pending",
  };
  return cloneSession(activeSession);
}

export function synchronizeAPIKeyRotationApplication(applicationId: string): APIKeyRotationSession | null {
  if (activeSession && activeSession.applicationId !== applicationId.trim()) activeSession = null;
  return activeSession ? cloneSession(activeSession) : null;
}

export function readAPIKeyRotationSession(applicationId: string): APIKeyRotationSession | null {
  if (!activeSession || activeSession.applicationId !== applicationId.trim()) return null;
  return cloneSession(activeSession);
}

export function recordAPIKeyRotationReplacement(
  applicationId: string,
  replacement: APIKeyRecord,
): APIKeyRotationSession {
  const session = requireSession(applicationId);
  if (session.phase !== "replacement_pending") {
    throw new APIKeyRotationSessionError("api_key_rotation_transition_invalid");
  }
  validateReplacement(session, replacement);
  activeSession = {
    ...session,
    replacementApiKeyId: replacement.apiKeyId,
    replacementCreatedAt: replacement.createdAt,
    replacementLastUsedAt: replacement.lastUsedAt,
    phase: replacement.lastUsedAt ? "verified" : "verification_pending",
  };
  return cloneSession(activeSession);
}

export function refreshAPIKeyRotationVerification(
  applicationId: string,
  replacement: APIKeyRecord,
): APIKeyRotationSession {
  const session = requireSession(applicationId);
  if (session.phase === "replacement_pending") {
    throw new APIKeyRotationSessionError("api_key_rotation_replacement_missing");
  }
  validateReplacement(session, replacement);
  activeSession = {
    ...session,
    replacementLastUsedAt: replacement.lastUsedAt,
    phase: replacement.lastUsedAt ? "verified" : "verification_pending",
  };
  return cloneSession(activeSession);
}

export function currentAPIKeyRotationSourceVersion(
  applicationId: string,
  source: APIKeyRecord,
): number {
  const session = requireSession(applicationId);
  if (session.phase !== "verified") {
    throw new APIKeyRotationSessionError("api_key_rotation_replacement_unverified");
  }
  if (source.apiKeyId !== session.sourceApiKeyId || source.applicationId !== session.applicationId ||
    source.ownerSubjectRef !== session.ownerSubjectRef || !sameScopes(source.scopes, session.scopes)) {
    throw new APIKeyRotationSessionError("api_key_rotation_source_mismatch");
  }
  if (source.lifecycleState !== "active" || source.effectiveState !== "active") {
    throw new APIKeyRotationSessionError("api_key_rotation_source_inactive");
  }
  return source.recordVersion;
}

export function completeAPIKeyRotationSession(
  applicationId: string,
  revokedSource: APIKeyRecord,
): void {
  const session = requireSession(applicationId);
  if (session.phase !== "verified" || revokedSource.apiKeyId !== session.sourceApiKeyId ||
    revokedSource.applicationId !== session.applicationId || revokedSource.ownerSubjectRef !== session.ownerSubjectRef ||
    !sameScopes(revokedSource.scopes, session.scopes) ||
    revokedSource.lifecycleState !== "revoked" || revokedSource.effectiveState !== "revoked") {
    throw new APIKeyRotationSessionError("api_key_rotation_completion_invalid");
  }
  activeSession = null;
}

export function cancelAPIKeyRotationSession(applicationId: string): void {
  if (activeSession?.applicationId === applicationId.trim()) activeSession = null;
}

function requireSession(applicationId: string): APIKeyRotationSession {
  if (!activeSession || activeSession.applicationId !== applicationId.trim()) {
    throw new APIKeyRotationSessionError("api_key_rotation_session_missing");
  }
  return activeSession;
}

function validateReplacement(session: APIKeyRotationSession, replacement: APIKeyRecord): void {
  if (replacement.apiKeyId === session.sourceApiKeyId || replacement.applicationId !== session.applicationId ||
    replacement.ownerSubjectRef !== session.ownerSubjectRef || !sameScopes(replacement.scopes, session.scopes)) {
    throw new APIKeyRotationSessionError("api_key_rotation_replacement_mismatch");
  }
  if (replacement.lifecycleState !== "active" || replacement.effectiveState !== "active") {
    throw new APIKeyRotationSessionError("api_key_rotation_replacement_inactive");
  }
}

function sameScopes(left: APIKeyScope[], right: APIKeyScope[]): boolean {
  const normalizedLeft = normalizeScopes(left);
  const normalizedRight = normalizeScopes(right);
  return normalizedLeft.length === normalizedRight.length &&
    normalizedLeft.every((scope, index) => scope === normalizedRight[index]);
}

function normalizeScopes(scopes: APIKeyScope[]): APIKeyScope[] {
  return [...new Set(scopes)].sort();
}

function cloneSession(session: APIKeyRotationSession): APIKeyRotationSession {
  return { ...session, scopes: [...session.scopes] };
}
