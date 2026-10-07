import { invitationDate, invitationFailureMessage } from "./workspaceInvitationMessages.ts";
import { useTranslation } from "react-i18next";
import { useLocalePreference } from "../../i18n/LocaleProvider.tsx";
import "../../i18n/identityInvitationResources.ts";
import type { ReactNode } from "react";
import type { InvitationEffectiveState, WorkspaceInvitation, WorkspaceInvitationError, WorkspaceInvitationPreview } from "./workspaceInvitationConsumer.ts";
import "./workspaceInvitation.css";

const STATE_MARKS = { pending: "◷", claimed: "✓", revoked: "×", expired: "◴" };

export function InvitationStateBadge({ state }: { state: InvitationEffectiveState }) {
  const { t } = useTranslation("identity");
  return <span className={`invitation-state is-${state}`}><span aria-hidden="true">{STATE_MARKS[state]}</span>{t($ => $.invitations.states[state])}</span>;
}

export function InvitationFailureNotice({ failure }: { failure: WorkspaceInvitationError }) {
  const { t } = useTranslation("identity");
  return <div className="invitation-notice is-error" role="alert"><strong>{failure.code === "workspace_invitation_not_claimable" ? t($ => $.invitations.notClaimableTitle) : t($ => $.invitations.cannotVerifyTitle)}</strong><p>{invitationFailureMessage(t, failure.code)}</p><code>{failure.code}</code></div>;
}

export function InvitationNotice({ title, children, tone = "neutral" }: { title: string; children: ReactNode; tone?: "neutral" | "warning" | "success" }) {
  return <div className={`invitation-notice is-${tone}`}><strong>{title}</strong><div>{children}</div></div>;
}

export function InvitationFacts({ invitation }: { invitation: WorkspaceInvitation | WorkspaceInvitationPreview }) {
  const { t } = useTranslation("identity");
  const { locale } = useLocalePreference();
  const roleKey = "role" in invitation ? invitation.role.roleKey : invitation.roleKey;
  const catalog = "role" in invitation ? invitation.role.catalogVersion : invitation.roleCatalogVersion;
  return <dl className="invitation-facts">
    <div className="invitation-scope-fact"><dt>{t($ => $.invitations.exactScope)}</dt><dd>{invitation.tenantRef} / {invitation.workspaceId}</dd></div>
    <div><dt>{t($ => $.invitations.roleDefinition)}</dt><dd>{roleKey}<small>{catalog}</small></dd></div>
    <div><dt>{t($ => $.invitations.expires)}</dt><dd><time dateTime={invitation.expiresAt} title={invitation.expiresAt}>{invitationDate(t, invitation.expiresAt, locale)}</time></dd></div>
    <div><dt>{t($ => $.invitations.record)}</dt><dd>v{invitation.recordVersion} · {t($ => $.invitations.states[invitation.effectiveState])}</dd></div>
  </dl>;
}

export function invitationShortId(value: string) { return `${value.slice(0, 10)}…${value.slice(-5)}`; }
