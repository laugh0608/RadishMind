import type { ReactNode } from "react";
import type { InvitationEffectiveState, WorkspaceInvitation, WorkspaceInvitationError, WorkspaceInvitationPreview } from "./workspaceInvitationConsumer.ts";
import "./workspaceInvitation.css";

const STATE_MARKS = { pending: "◷", claimed: "✓", revoked: "×", expired: "◴" };

export function InvitationStateBadge({ state }: { state: InvitationEffectiveState }) {
  return <span className={`invitation-state is-${state}`}><span aria-hidden="true">{STATE_MARKS[state]}</span>{state.toUpperCase()}</span>;
}

export function InvitationFailureNotice({ failure }: { failure: WorkspaceInvitationError }) {
  return <div className="invitation-notice is-error" role="alert"><strong>{failure.code === "workspace_invitation_not_claimable" ? "Invitation no longer claimable" : "Invitation cannot be verified"}</strong><p>{failure.message}</p><code>{failure.code}</code></div>;
}

export function InvitationNotice({ title, children, tone = "neutral" }: { title: string; children: ReactNode; tone?: "neutral" | "warning" | "success" }) {
  return <div className={`invitation-notice is-${tone}`}><strong>{title}</strong><div>{children}</div></div>;
}

export function InvitationFacts({ invitation }: { invitation: WorkspaceInvitation | WorkspaceInvitationPreview }) {
  const roleKey = "role" in invitation ? invitation.role.roleKey : invitation.roleKey;
  const catalog = "role" in invitation ? invitation.role.catalogVersion : invitation.roleCatalogVersion;
  return <dl className="invitation-facts">
    <div className="invitation-scope-fact"><dt>Exact scope</dt><dd>{invitation.tenantRef} / {invitation.workspaceId}</dd></div>
    <div><dt>Role definition</dt><dd>{roleKey}<small>{catalog}</small></dd></div>
    <div><dt>Expires</dt><dd><time dateTime={invitation.expiresAt}>{invitationDate(invitation.expiresAt)}</time></dd></div>
    <div><dt>Record</dt><dd>v{invitation.recordVersion} · {invitation.effectiveState}</dd></div>
  </dl>;
}

export function invitationDate(value: string) {
  return new Intl.DateTimeFormat("en-GB", { dateStyle: "medium", timeStyle: "short", timeZone: "UTC" }).format(new Date(value)) + " UTC";
}

export function invitationShortId(value: string) { return `${value.slice(0, 10)}…${value.slice(-5)}`; }
