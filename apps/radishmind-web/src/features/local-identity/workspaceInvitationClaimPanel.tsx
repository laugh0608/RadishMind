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
    const focusable = Array.from(dialog.current.querySelectorAll<HTMLElement>("button:not(:disabled), input:not(:disabled), a[href], [tabindex='0']"));
    const first = focusable[0]; const last = focusable.at(-1);
    if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last?.focus(); }
    else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first?.focus(); }
  }

  const step = claimed ? 3 : busy === "claim" ? 3 : preview ? 2 : busy === "preview" ? 1 : 0;

  return <section ref={dialog} className="workspace-invitation-claim" role="dialog" aria-modal="true" aria-labelledby="invitation-claim-title" onKeyDown={handleDialogKeyDown}>
    <header className="invitation-claim-header">
      <div className="invitation-brand"><span aria-hidden="true">R</span><strong>RadishMind</strong></div>
      <div className="invitation-actions"><span className="invitation-state">SIGNED IN · DEV / TEST</span><button type="button" onClick={() => void onLogout()} disabled={Boolean(busy)}>Sign out</button>
        <button type="button" aria-label="Close invitation claim" onClick={onClose}>×</button></div>
    </header>
    <div className="invitation-claim-body">
      <aside className="invitation-claim-navigation" aria-label="Signed-in tasks">
        <span className="invitation-rail-label">SIGNED-IN TASKS</span>
        <button type="button" onClick={onOpenSecurity}><strong>Account</strong><small>Local identity</small></button>
        <button type="button" onClick={onOpenSecurity}><strong>Sessions</strong><small>Self-service security</small></button>
        <button type="button" aria-current="page"><strong>Claim invitation</strong><small>Workspace access</small></button>
        <InvitationNotice title="Current actor" tone="success"><p>{profile.account.displayName}</p><p>{profile.account.lifecycleState === "active" ? "✓ Active local account" : "Account inactive"}</p>
          <p>{preview || claimed ? "✓ Same tenant verified by server" : "Tenant verified during preview"}</p><p>{profile.capabilities.recentAuthentication ? "✓ Recent authentication" : "Recent authentication required"}</p></InvitationNotice>
        <InvitationNotice title={claimed ? "ACCESS COMMITTED" : "NO ACCESS YET"} tone="warning">Code possession and preview are not membership. Only a committed atomic claim grants workspace access.</InvitationNotice>
      </aside>
      <main className="invitation-claim-main">
        <header className="invitation-claim-title"><h2 id="invitation-claim-title">Claim workspace invitation</h2><p>Review the server preview, then explicitly confirm one atomic claim.</p><span className="invitation-state">NO ACTIVE WORKSPACE REQUIRED</span></header>
        <ol className="invitation-progress invitation-claim-progress">{["Code", "Preview", "Confirm", "Claim"].map((label, index) =>
          <li key={label} aria-current={step === index ? "step" : undefined}><b>{step > index || claimed ? "✓" : index + 1}</b><span>{label}</span></li>)}</ol>
        <div className="invitation-claim-workbench">
          <div className="invitation-claim-flow">
            {failure ? <InvitationFailureNotice failure={failure} /> : null}
            {claimed ? <section className="invitation-claim-success" role="status">
              <header className="invitation-section-header"><h3>Workspace membership created</h3><InvitationStateBadge state="claimed" /></header>
              <InvitationFacts invitation={claimed} />
              <p>The full claim committed. The invitation code has been cleared.</p>
              <p>{busy === "refresh" ? "Reloading available workspaces…" : "Available workspaces reloaded. Choose a workspace explicitly when you return; your active selection has not changed."}</p>
              {busy !== "refresh" ? <div className="invitation-available-workspaces"><h4>Available workspaces</h4><ul>{availableWorkspaces.map((membership) =>
                <li key={membership.membershipId}>{membership.tenantRef} / <strong>{membership.workspaceId}</strong></li>)}</ul></div> : null}
              <div className="invitation-actions"><button type="button" onClick={onClose}>Done</button><button type="button" onClick={changeCode} disabled={Boolean(busy)}>Claim another invitation</button></div>
            </section> : <>
              <section className="invitation-code-entry">
                <header className="invitation-section-header"><h3>Invitation code</h3>{preview ? <button type="button" onClick={changeCode}>Change code</button> : null}</header>
                {preview ? <p>Verified code hidden · retained only in this component memory</p> : <form onSubmit={(event) => void verifyCode(event)} className="invitation-form">
                  <label><span className="invitation-visually-hidden">Invitation code</span><input ref={input} type="password" aria-label="Invitation code" placeholder="Paste invitation code" autoComplete="off" autoCapitalize="none" spellCheck={false} maxLength={512} value={code}
                    disabled={Boolean(busy)} onChange={(event) => { setCode(event.target.value); setFailure(null); }} /></label>
                  <p>Use the code shared by your administrator. Invalid input does not reveal whether an invitation exists.</p>
                  <button type="submit" className="invitation-primary" disabled={!code.trim() || Boolean(busy)}>{busy === "preview" ? "Verifying…" : "Verify and preview"}</button>
                </form>}
              </section>
              {preview ? <>
                <section className="invitation-preview"><header className="invitation-section-header"><h3>Server preview</h3><InvitationStateBadge state="pending" /></header><InvitationFacts invitation={preview} /><p>{preview.role.summary}</p></section>
                <section className="invitation-claim-confirmation"><h3>Confirm membership creation</h3><p>One submission · atomic commit · no partial access</p>
                  {!profile.capabilities.recentAuthentication ? <InvitationNotice title="Recent authentication required" tone="warning">Sign out and authenticate again, then re-enter the invitation code.</InvitationNotice> : null}
                  <label className="invitation-confirmation"><input type="checkbox" checked={confirmed} disabled={Boolean(busy) || !profile.capabilities.recentAuthentication} onChange={(event) => setConfirmed(event.target.checked)} />
                    <span>I understand that a new membership and catalog-derived role assignment are created only if the full claim commits.</span></label>
                  <p>Success reloads available workspaces but does not select or navigate to the new workspace.</p>
                  <button type="button" className="invitation-primary" disabled={!confirmed || Boolean(busy) || !profile.capabilities.recentAuthentication} onClick={() => void confirmClaim()}>{busy === "claim" ? "Claiming…" : "Confirm & claim workspace"}</button>
                </section>
              </> : null}
            </>}
          </div>
          <aside className="invitation-review-rail" aria-label="Claim effects and recovery">
            <header className="invitation-section-header"><h3>Claim effects</h3><span className="invitation-state">ATOMIC</span></header>
            <ol className="invitation-effects"><li><strong>WorkspaceMembership</strong><span>New active membership</span></li><li><strong>LocalRoleAssignment</strong><span>Catalog-derived grants</span></li><li><strong>WorkspaceInvitation</strong><span>Pending → claimed</span></li></ol>
            <InvitationNotice title="Preview does not reserve access">Later revocation, expiry, role drift or another claim makes this confirmation stale. The server checks the exact version again.</InvitationNotice>
            <InvitationNotice title="Zero partial writes" tone="warning">Account, tenant, catalog, membership, assignment or store failure rolls back every write.</InvitationNotice>
            <InvitationNotice title="Recovery">Invalid: re-enter or ask the administrator for a new code. Claimed, revoked, expired or drifted: a new invitation is required.</InvitationNotice>
            <InvitationNotice title="State invalidation">Actor, tenant, workspace, session or route changes clear the code, preview, confirmation and late responses.</InvitationNotice>
          </aside>
        </div>
      </main>
    </div>
  </section>;
}
