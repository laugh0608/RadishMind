import { useTranslation } from "react-i18next";
import { LanguageSelector } from "../../i18n/LanguageSelector.tsx";
import { identityRoleCopy } from "./localIdentityRoleMessages.ts";
import "../../i18n/identityInvitationResources.ts";
import { useEffect, useRef, useState, type FormEvent, type KeyboardEvent } from "react";
import type { LocalIdentityContextValue } from "./localIdentityContext.ts";
import { readControlPlaneReadDevLiveConfig } from "../control-plane-read/devLiveReadConsumer.ts";
import { WorkspaceInvitationError, claimWorkspaceInvitation, isInvitationCode, previewWorkspaceInvitation,
  workspaceInvitationFailure, type WorkspaceInvitation, type WorkspaceInvitationPreview } from "./workspaceInvitationConsumer.ts";
import { workspaceInvitationAuthorityKey } from "./workspaceInvitationState.ts";
import { availableIdentityWorkspaces } from "./localIdentityWorkspaceAccess.ts";
import { useWorkspaceInvitationRequests } from "./useWorkspaceInvitationRequests.ts";
import { InvitationFacts, InvitationFailureNotice, InvitationNotice, InvitationStateBadge } from "./workspaceInvitationView.tsx";

export function WorkspaceInvitationClaimPanel({ identity, onClose, onOpenSecurity, onLogout }: {
  identity: LocalIdentityContextValue; onClose: () => void; onOpenSecurity: () => void; onLogout: () => Promise<void>;
}) {
  const { t } = useTranslation("identity");
  const { profile } = identity;
  const config = { ...identity.config, tenantRef: readControlPlaneReadDevLiveConfig().tenantRef };
  const [code, setCode] = useState("");
  const [preview, setPreview] = useState<WorkspaceInvitationPreview | null>(null);
  const [confirmed, setConfirmed] = useState(false);
  const [busy, setBusy] = useState<"" | "preview" | "claim" | "refresh">("");
  const [failure, setFailure] = useState<WorkspaceInvitationError | null>(null);
  const [claimed, setClaimed] = useState<WorkspaceInvitation | null>(null);
  const input = useRef<HTMLInputElement>(null);
  const dialog = useRef<HTMLElement>(null);
  const authorityKey = workspaceInvitationAuthorityKey(profile, JSON.stringify([config, "claim-invitation"]));
  const requests = useWorkspaceInvitationRequests(authorityKey, identity.readAuthorityRevision, clearInteraction);
  const availableWorkspaces = availableIdentityWorkspaces(profile);

  useEffect(() => {
    const previousFocus = document.activeElement;
    input.current?.focus();
    return () => { if (previousFocus instanceof HTMLElement && previousFocus.isConnected) previousFocus.focus(); };
  }, []);

  function clearInteraction() {
    setCode(""); setPreview(null); setConfirmed(false); setFailure(null); setClaimed(null); setBusy("");
  }

  function changeCode() { requests.invalidate(); clearInteraction(); requestAnimationFrame(() => input.current?.focus()); }

  async function verifyCode(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (busy) return;
    const submittedCode = code.trim();
    const request = requests.begin();
    setPreview(null); setConfirmed(false); setFailure(null); setClaimed(null);
    if (!isInvitationCode(submittedCode)) {
      setCode(""); setFailure(new WorkspaceInvitationError("workspace_invitation_invalid")); return;
    }
    setBusy("preview"); setCode(submittedCode);
    try {
      const result = await previewWorkspaceInvitation(config, submittedCode, request.signal);
      if (!requests.isCurrent(request)) return;
      setPreview(result);
    } catch (error) {
      if (!requests.isCurrent(request)) return;
      setCode(""); setPreview(null); setConfirmed(false); setFailure(workspaceInvitationFailure(error));
    } finally { if (requests.isCurrent(request)) setBusy(""); }
  }

  async function confirmClaim() {
    if (!preview || !confirmed || busy || !profile.capabilities.recentAuthentication) return;
    const request = requests.begin();
    const invitationCode = code;
    setCode(""); setConfirmed(false); setBusy("claim"); setFailure(null);
    try {
      const result = await claimWorkspaceInvitation(config, { invitationCode, preview, userId: profile.account.userId, confirmed: true }, request.signal);
      if (!requests.isCurrent(request)) return;
      setPreview(null); setClaimed(result); setBusy("refresh");
      requests.broadcastChanged();
      await identity.refresh();
      if (requests.isCurrent(request)) setBusy("");
    } catch (error) {
      if (!requests.isCurrent(request)) return;
      setCode(""); setPreview(null); setConfirmed(false); setBusy(""); setFailure(workspaceInvitationFailure(error));
    }
  }

  function handleDialogKeyDown(event: KeyboardEvent<HTMLElement>) {
    if (event.key === "Escape") { event.preventDefault(); onClose(); return; }
    if (event.key !== "Tab" || !dialog.current) return;
    const focusable = Array.from(dialog.current.querySelectorAll<HTMLElement>("button:not(:disabled), input:not(:disabled), select:not(:disabled), a[href], [tabindex='0']"));
    const first = focusable[0]; const last = focusable.at(-1);
    if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last?.focus(); }
    else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first?.focus(); }
  }

  const step = claimed ? 3 : busy === "claim" ? 3 : preview ? 2 : busy === "preview" ? 1 : 0;

  return <section ref={dialog} className="workspace-invitation-claim" role="dialog" aria-modal="true" aria-labelledby="invitation-claim-title" onKeyDown={handleDialogKeyDown}>
    <header className="invitation-claim-header">
      <div className="invitation-brand"><span aria-hidden="true">R</span><strong>RadishMind</strong></div>
      <div className="invitation-actions"><LanguageSelector /><span className="invitation-state">{t($ => $.invitations.signedIn)}</span><button type="button" onClick={() => void onLogout()} disabled={Boolean(busy)}>{t($ => $.invitations.signOut)}</button>
        <button type="button" aria-label={t($ => $.invitations.closeClaim)} onClick={onClose}>×</button></div>
    </header>
    <div className="invitation-claim-body">
      <aside className="invitation-claim-navigation" aria-label={t($ => $.invitations.signedInTasks)}>
        <span className="invitation-rail-label">{t($ => $.invitations.signedInTasksHeading)}</span>
        <button type="button" onClick={onOpenSecurity}><strong>{t($ => $.invitations.account)}</strong><small>{t($ => $.invitations.localIdentity)}</small></button>
        <button type="button" onClick={onOpenSecurity}><strong>{t($ => $.invitations.sessions)}</strong><small>{t($ => $.invitations.selfService)}</small></button>
        <button type="button" aria-current="page"><strong>{t($ => $.invitations.claimInvitation)}</strong><small>{t($ => $.invitations.workspaceAccess)}</small></button>
        <InvitationNotice title={t($ => $.invitations.currentActor)} tone="success"><p>{profile.account.displayName}</p><p>{profile.account.lifecycleState === "active" ? t($ => $.invitations.activeAccount) : t($ => $.invitations.inactiveAccount)}</p>
          <p>{preview || claimed ? t($ => $.invitations.tenantVerified) : t($ => $.invitations.tenantDuringPreview)}</p><p>{profile.capabilities.recentAuthentication ? t($ => $.invitations.recentAuthVerified) : t($ => $.invitations.recentAuth)}</p></InvitationNotice>
        <InvitationNotice title={claimed ? t($ => $.invitations.accessCommitted) : t($ => $.invitations.noAccessYet)} tone="warning">{t($ => $.invitations.previewBoundary)}</InvitationNotice>
      </aside>
      <main className="invitation-claim-main">
        <header className="invitation-claim-title"><h2 id="invitation-claim-title">{t($ => $.invitations.claimTitle)}</h2><p>{t($ => $.invitations.claimHelp)}</p><span className="invitation-state">{t($ => $.invitations.noWorkspaceRequired)}</span></header>
        <ol className="invitation-progress invitation-claim-progress">{(["code", "preview", "confirm", "claim"] as const).map((stepKey, index) =>
          <li key={stepKey} aria-current={step === index ? "step" : undefined}><b>{step > index || claimed ? "✓" : index + 1}</b><span>{t($ => $.invitations.steps[stepKey])}</span></li>)}</ol>
        <div className="invitation-claim-workbench">
          <div className="invitation-claim-flow">
            {failure ? <InvitationFailureNotice failure={failure} /> : null}
            {claimed ? <section className="invitation-claim-success" role="status">
              <header className="invitation-section-header"><h3>{t($ => $.invitations.membershipCreated)}</h3><InvitationStateBadge state="claimed" /></header>
              <InvitationFacts invitation={claimed} />
              <p>{t($ => $.invitations.claimCommitted)}</p>
              <p>{busy === "refresh" ? t($ => $.invitations.refreshingWorkspaces) : t($ => $.invitations.workspacesRefreshed)}</p>
              {busy !== "refresh" ? <div className="invitation-available-workspaces"><h4>{t($ => $.invitations.availableWorkspaces)}</h4><ul>{availableWorkspaces.map((membership) =>
                <li key={membership.membershipId}>{membership.tenantRef} / <strong>{membership.workspaceId}</strong></li>)}</ul></div> : null}
              <div className="invitation-actions"><button type="button" onClick={onClose}>{t($ => $.invitations.done)}</button><button type="button" onClick={changeCode} disabled={Boolean(busy)}>{t($ => $.invitations.claimAnother)}</button></div>
            </section> : <>
              <section className="invitation-code-entry">
                <header className="invitation-section-header"><h3>{t($ => $.invitations.invitationCode)}</h3>{preview ? <button type="button" onClick={changeCode}>{t($ => $.invitations.changeCode)}</button> : null}</header>
                {preview ? <p>{t($ => $.invitations.verifiedHidden)}</p> : <form onSubmit={(event) => void verifyCode(event)} className="invitation-form">
                  <label><span className="invitation-visually-hidden">{t($ => $.invitations.invitationCode)}</span><input ref={input} type="password" aria-label={t($ => $.invitations.invitationCode)} placeholder={t($ => $.invitations.pasteCode)} autoComplete="off" autoCapitalize="none" spellCheck={false} maxLength={512} value={code}
                    disabled={Boolean(busy)} onChange={(event) => { setCode(event.target.value); setFailure(null); }} /></label>
                  <p>{t($ => $.invitations.codeHelp)}</p>
                  <button type="submit" className="invitation-primary" disabled={!code.trim() || Boolean(busy)}>{busy === "preview" ? t($ => $.invitations.verifying) : t($ => $.invitations.verifyPreview)}</button>
                </form>}
              </section>
              {preview ? <>
                <section className="invitation-preview"><header className="invitation-section-header"><h3>{t($ => $.invitations.serverPreview)}</h3><InvitationStateBadge state="pending" /></header><InvitationFacts invitation={preview} /><p>{identityRoleCopy(t, preview.role.roleKey).summary}</p></section>
                <section className="invitation-claim-confirmation"><h3>{t($ => $.invitations.confirmMembership)}</h3><p>{t($ => $.invitations.atomicSubmission)}</p>
                  {!profile.capabilities.recentAuthentication ? <InvitationNotice title={t($ => $.invitations.recentAuth)} tone="warning">{t($ => $.invitations.claimReauth)}</InvitationNotice> : null}
                  <label className="invitation-confirmation"><input type="checkbox" checked={confirmed} disabled={Boolean(busy) || !profile.capabilities.recentAuthentication} onChange={(event) => setConfirmed(event.target.checked)} />
                    <span>{t($ => $.invitations.claimConfirmation)}</span></label>
                  <p>{t($ => $.invitations.selectionUnchanged)}</p>
                  <button type="button" className="invitation-primary" disabled={!confirmed || Boolean(busy) || !profile.capabilities.recentAuthentication} onClick={() => void confirmClaim()}>{busy === "claim" ? t($ => $.invitations.claiming) : t($ => $.invitations.confirmClaim)}</button>
                </section>
              </> : null}
            </>}
          </div>
          <aside className="invitation-review-rail" aria-label={t($ => $.invitations.claimEffectsRecovery)}>
            <header className="invitation-section-header"><h3>{t($ => $.invitations.claimEffects)}</h3><span className="invitation-state">{t($ => $.invitations.atomic)}</span></header>
            <ol className="invitation-effects"><li><strong>WorkspaceMembership</strong><span>{t($ => $.invitations.newMembership)}</span></li><li><strong>LocalRoleAssignment</strong><span>{t($ => $.invitations.catalogGrants)}</span></li><li><strong>WorkspaceInvitation</strong><span>{t($ => $.invitations.pendingClaimed)}</span></li></ol>
            <InvitationNotice title={t($ => $.invitations.previewNoReservation)}>{t($ => $.invitations.previewStaleHelp)}</InvitationNotice>
            <InvitationNotice title={t($ => $.invitations.noPartialWrites)} tone="warning">{t($ => $.invitations.rollbackHelp)}</InvitationNotice>
            <InvitationNotice title={t($ => $.invitations.recovery)}>{t($ => $.invitations.recoveryHelp)}</InvitationNotice>
            <InvitationNotice title={t($ => $.invitations.invalidation)}>{t($ => $.invitations.invalidationHelp)}</InvitationNotice>
          </aside>
        </div>
      </main>
    </div>
  </section>;
}
