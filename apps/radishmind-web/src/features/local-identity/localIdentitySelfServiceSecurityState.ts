import {
  LocalIdentitySelfServiceSecurityError,
  type LocalIdentitySelfServiceSessionSummary,
} from "./localIdentitySelfServiceSecurityConsumer.ts";
import type { LocalIdentityAccountProfile } from "./localIdentityConsumer.ts";

export type LocalIdentitySelfServiceSecurityScope = {
  generation: number;
  userId: string;
  currentSessionId: string;
};

export type LocalIdentitySelfServiceSecurityProjection = {
  currentSession: LocalIdentitySelfServiceSessionSummary | null;
  otherActiveSessions: LocalIdentitySelfServiceSessionSummary[];
  endedSessions: LocalIdentitySelfServiceSessionSummary[];
  activeLocalPasswordSessions: LocalIdentitySelfServiceSessionSummary[];
};

export function localIdentitySelfServiceSecurityScope(
  profile: LocalIdentityAccountProfile,
  generation: number,
): LocalIdentitySelfServiceSecurityScope {
  return {
    generation,
    userId: profile.account.userId,
    currentSessionId: profile.session.sessionId,
  };
}

export function localIdentitySelfServiceSecurityScopeKey(
  profile: LocalIdentityAccountProfile,
  generation: number,
): string {
  const scope = localIdentitySelfServiceSecurityScope(profile, generation);
  return `${scope.userId}:${scope.currentSessionId}:${scope.generation}`;
}

export function localIdentitySelfServiceSecurityResponseMatchesScope(
  expected: LocalIdentitySelfServiceSecurityScope,
  observed: LocalIdentitySelfServiceSecurityScope,
): boolean {
  return expected.generation === observed.generation && expected.userId === observed.userId &&
    expected.currentSessionId === observed.currentSessionId;
}

export function mergeLocalIdentitySelfServiceSessions(
  current: LocalIdentitySelfServiceSessionSummary[],
  incoming: LocalIdentitySelfServiceSessionSummary[],
): LocalIdentitySelfServiceSessionSummary[] {
  const identifiers = new Set(current.map((session) => session.sessionId));
  if (incoming.some((session) => identifiers.has(session.sessionId)) ||
    new Set(incoming.map((session) => session.sessionId)).size !== incoming.length) {
    throw invalidState("Session pagination repeated a canonical session.");
  }
  return [...current, ...incoming];
}

export function projectLocalIdentitySelfServiceSessions(
  sessions: LocalIdentitySelfServiceSessionSummary[],
  currentSessionId: string,
): LocalIdentitySelfServiceSecurityProjection {
  const declaredCurrent = sessions.filter((session) => session.currentSession);
  if (declaredCurrent.length > 1 || declaredCurrent.some((session) => session.sessionId !== currentSessionId) ||
    sessions.some((session) => session.sessionId === currentSessionId && !session.currentSession)) {
    throw invalidState("Session directory current actor projection is inconsistent.");
  }
  return {
    currentSession: declaredCurrent[0] ?? null,
    otherActiveSessions: sessions.filter((session) =>
      !session.currentSession && session.effectiveState === "active"
    ),
    endedSessions: sessions.filter((session) => session.effectiveState !== "active"),
    activeLocalPasswordSessions: sessions.filter((session) =>
      session.effectiveState === "active" && session.authenticationMethod === "local_password"
    ),
  };
}

function invalidState(message: string): LocalIdentitySelfServiceSecurityError {
  return new LocalIdentitySelfServiceSecurityError(0, "local_identity_response_invalid", message);
}
