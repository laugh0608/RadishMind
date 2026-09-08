import type { LocalIdentityAccountProfile } from "./localIdentityConsumer.ts";

export function availableIdentityWorkspaces(profile: LocalIdentityAccountProfile, now = Date.now()) {
  const seen = new Set<string>();
  return profile.workspaceMemberships.filter((membership) => {
    const key = JSON.stringify([membership.tenantRef, membership.workspaceId]);
    if (membership.lifecycleState !== "active" || (membership.expiresAt && Date.parse(membership.expiresAt) <= now) || seen.has(key)) return false;
    seen.add(key);
    return true;
  });
}
