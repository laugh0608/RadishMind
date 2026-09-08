import { useEffect, useRef, useState } from "react";
import { createInvitationRequestScope, isWorkspaceInvitationChangedEvent } from "./workspaceInvitationState.ts";

const INVITATION_CHANNEL = "radishmind-workspace-invitations-v1";

export function useWorkspaceInvitationRequests(
  authorityKey: string,
  readAuthorityRevision: () => number,
  onInvalidate: () => void,
) {
  const authority = useRef({ key: authorityKey, readRevision: readAuthorityRevision });
  authority.current = { key: authorityKey, readRevision: readAuthorityRevision };
  const clear = useRef(onInvalidate);
  const notifications = useRef<BroadcastChannel | null>(null);
  clear.current = onInvalidate;
  const [requests] = useState(() => createInvitationRequestScope(() =>
    `${authority.current.key}:${authority.current.readRevision()}`));

  useEffect(() => {
    requests.activate();
    const invalidate = () => { requests.invalidate(); clear.current(); };
    window.addEventListener("hashchange", invalidate);
    window.addEventListener("popstate", invalidate);
    window.addEventListener("pagehide", invalidate);
    const channel = typeof BroadcastChannel === "undefined" ? null : new BroadcastChannel(INVITATION_CHANNEL);
    notifications.current = channel;
    if (channel) channel.onmessage = (event: MessageEvent<unknown>) => {
      if (isWorkspaceInvitationChangedEvent(event.data)) invalidate();
    };
    return () => {
      requests.dispose();
      window.removeEventListener("hashchange", invalidate);
      window.removeEventListener("popstate", invalidate);
      window.removeEventListener("pagehide", invalidate);
      channel?.close();
      if (notifications.current === channel) notifications.current = null;
    };
  }, [requests]);
  return { ...requests, broadcastChanged: () => {
    // Reuse the listener's channel so the initiating panel keeps its own committed result.
    notifications.current?.postMessage({ kind: "invitations_changed", version: 1 });
  } };
}
