import type { LocalIdentityAccountProfile } from "./localIdentityConsumer.ts";
import { WorkspaceInvitationError, type WorkspaceInvitationPage } from "./workspaceInvitationConsumer.ts";

export function workspaceInvitationAuthorityKey(profile: LocalIdentityAccountProfile, scope: string): string {
  return JSON.stringify([scope, profile.account.userId, profile.account.lifecycleState, profile.session.sessionId,
    profile.session.authenticationMethod, profile.session.expiresAt, profile.capabilities.recentAuthentication]);
}

export type InvitationRequest = { generation: number; authority: string; signal: AbortSignal };

// Only request metadata lives here. Codes, review details and confirmations belong to components.
export function createInvitationRequestScope(readAuthority: () => string) {
  let generation = 0;
  let active = true;
  let controller: AbortController | null = null;
  function invalidate() {
    generation += 1;
    controller?.abort();
    controller = null;
  }
  return {
    activate() { invalidate(); active = true; },
    invalidate,
    dispose() { invalidate(); active = false; },
    begin(): InvitationRequest {
      invalidate();
      controller = new AbortController();
      if (!active) controller.abort();
      return { generation, authority: readAuthority(), signal: controller.signal };
    },
    isCurrent(request: InvitationRequest): boolean {
      return active && !request.signal.aborted && request.generation === generation && request.authority === readAuthority();
    },
  };
}

export function mergeWorkspaceInvitationPages(current: WorkspaceInvitationPage, incoming: WorkspaceInvitationPage): WorkspaceInvitationPage {
  const ids = new Set(current.invitations.map((row) => row.invitationId));
  if (current.tenantRef !== incoming.tenantRef || current.workspaceId !== incoming.workspaceId ||
    current.asOf !== incoming.asOf || incoming.invitations.some((row) => ids.has(row.invitationId)) ||
    new Set(incoming.invitations.map((row) => row.invitationId)).size !== incoming.invitations.length ||
    (incoming.nextCursor && incoming.nextCursor === current.nextCursor)) {
    throw new WorkspaceInvitationError("workspace_invitation_response_invalid");
  }
  return { ...incoming, invitations: [...current.invitations, ...incoming.invitations] };
}

export function isWorkspaceInvitationChangedEvent(value: unknown): boolean {
  if (!value || typeof value !== "object" || Array.isArray(value)) return false;
  const message = value as Record<string, unknown>;
  return Object.keys(message).length === 2 && message.kind === "invitations_changed" && message.version === 1;
}
