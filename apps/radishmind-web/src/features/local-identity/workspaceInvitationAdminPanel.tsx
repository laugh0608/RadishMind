import { useEffect, useState } from "react";
import { useLocalIdentity, type LocalIdentityContextValue } from "./localIdentityContext.ts";
import { readLocalIdentityRoleCatalog, type LocalIdentityRoleDefinition } from "./localIdentityAdministrationConsumer.ts";
import {
  INVITATION_ROLES, INVITATION_TTLS, createWorkspaceInvitation, listWorkspaceInvitations, revokeWorkspaceInvitation,
  workspaceInvitationFailure, type InvitationAdminConfig, type InvitationEffectiveState, type InvitationTTL,
  type WorkspaceInvitation, type WorkspaceInvitationError, type WorkspaceInvitationPage,
} from "./workspaceInvitationConsumer.ts";
import { mergeWorkspaceInvitationPages, workspaceInvitationAuthorityKey } from "./workspaceInvitationState.ts";
import { useWorkspaceInvitationRequests } from "./useWorkspaceInvitationRequests.ts";
import { InvitationFacts, InvitationFailureNotice, InvitationNotice, InvitationStateBadge, invitationDate, invitationShortId } from "./workspaceInvitationView.tsx";

type DirectoryState = { status: "loading" } | { status: "ready"; page: WorkspaceInvitationPage } | { status: "failed"; failure: WorkspaceInvitationError };
type Review = { kind: "none" } | { kind: "create" } | { kind: "revoke"; invitation: WorkspaceInvitation };
type Handoff = { invitation: WorkspaceInvitation; invitationCode: string };
const TTL_LABELS: Record<InvitationTTL, string> = { "1h": "1 hour", "24h": "24 hours", "72h": "72 hours", "7d": "7 days" };

export function WorkspaceInvitationAdminPanel({ tenantRef, workspaceId }: { tenantRef: string; workspaceId: string }) {
  const identity = useLocalIdentity();
  if (!identity) return <InvitationNotice title="Invitations unavailable">Sign in with the local development/test identity service to manage workspace invitations.</InvitationNotice>;
  const authorityKey = workspaceInvitationAuthorityKey(identity.profile, JSON.stringify([identity.config, tenantRef, workspaceId, "admin-invitations"]));
  return <WorkspaceInvitationAdministration key={`${authorityKey}:${identity.authorityRevision}`} identity={identity} authorityKey={authorityKey} tenantRef={tenantRef} workspaceId={workspaceId} />;
}

function WorkspaceInvitationAdministration({ identity, authorityKey, tenantRef, workspaceId }: {
  identity: LocalIdentityContextValue; authorityKey: string; tenantRef: string; workspaceId: string;
}) {
  const config: InvitationAdminConfig = { ...identity.config, tenantRef, workspaceId };
  const [directory, setDirectory] = useState<DirectoryState>({ status: "loading" });
  const [filter, setFilter] = useState<InvitationEffectiveState>("pending");
  const [reload, setReload] = useState(0);
  const [selectedId, setSelectedId] = useState("");
  const [review, setReview] = useState<Review>({ kind: "none" });
  const [roles, setRoles] = useState<LocalIdentityRoleDefinition[]>([]);
  const [selectedRole, setSelectedRole] = useState("");
  const [ttlPolicy, setTTLPolicy] = useState<InvitationTTL>("24h");
  const [confirmed, setConfirmed] = useState(false);
  const [handoff, setHandoff] = useState<Handoff | null>(null);
  const [copyStatus, setCopyStatus] = useState<"idle" | "copied" | "failed">("idle");
  const [busy, setBusy] = useState<"" | "catalog" | "create" | "revoke" | "page">("");
  const [failure, setFailure] = useState<WorkspaceInvitationError | null>(null);
  const [notice, setNotice] = useState("");
  const requests = useWorkspaceInvitationRequests(authorityKey, identity.readAuthorityRevision, () => {
    clearInteraction(); setSelectedId(""); setDirectory({ status: "loading" }); setReload((value) => value + 1);
  });
  const selected = directory.status === "ready" ? directory.page.invitations.find((row) => row.invitationId === selectedId) : undefined;
  const role = roles.find((entry) => entry.roleKey === selectedRole);
  const canMutate = identity.profile.capabilities.recentAuthentication && identity.profile.account.lifecycleState === "active";

  useEffect(() => { void loadDirectory(); }, [filter, reload]);

  function clearInteraction() {
    setHandoff(null); setCopyStatus("idle"); setReview({ kind: "none" }); setRoles([]); setSelectedRole("");
    setTTLPolicy("24h"); setConfirmed(false); setBusy(""); setFailure(null); setNotice("");
  }

  async function loadDirectory(preserveHandoff = false) {
    const request = requests.begin();
    if (!preserveHandoff) clearInteraction();
    setSelectedId(""); setDirectory({ status: "loading" });
    try {
      const page = await listWorkspaceInvitations(config, { effectiveState: filter, limit: 50 }, request.signal);
      if (!requests.isCurrent(request)) return;
      setDirectory({ status: "ready", page });
    } catch (error) {
      if (!requests.isCurrent(request)) return;
      clearInteraction(); setDirectory({ status: "failed", failure: workspaceInvitationFailure(error) });
    }
  }

  async function loadMore() {
    if (directory.status !== "ready" || !directory.page.nextCursor || busy) return;
    const current = directory.page;
    const request = requests.begin();
    setHandoff(null); setReview({ kind: "none" }); setConfirmed(false); setFailure(null); setBusy("page");
    try {
      const incoming = await listWorkspaceInvitations(config, { effectiveState: filter, limit: 50, cursor: current.nextCursor }, request.signal);
      if (!requests.isCurrent(request)) return;
      setDirectory({ status: "ready", page: mergeWorkspaceInvitationPages(current, incoming) });
    } catch (error) {
      if (!requests.isCurrent(request)) return;
      setSelectedId(""); setDirectory({ status: "failed", failure: workspaceInvitationFailure(error) });
    } finally { if (requests.isCurrent(request)) setBusy(""); }
  }

  async function startCreate() {
    const request = requests.begin();
    clearInteraction(); setReview({ kind: "create" }); setBusy("catalog");
    try {
      const result = await readLocalIdentityRoleCatalog(config, request.signal);
      if (!requests.isCurrent(request)) return;
      setRoles(result.catalog.roles.filter((entry) => INVITATION_ROLES.some((key) => key === entry.roleKey) && !entry.canManageLocalIdentity));
    } catch (error) {
      if (!requests.isCurrent(request)) return;
      setReview({ kind: "none" }); setFailure(workspaceInvitationFailure(error));
    } finally { if (requests.isCurrent(request)) setBusy(""); }
  }

  async function createInvitation() {
    if (!role || !confirmed || busy || !canMutate) return;
    const request = requests.begin();
    setBusy("create"); setFailure(null); setConfirmed(false);
    try {
      const result = await createWorkspaceInvitation(config, { role, ttlPolicy, confirmed: true }, request.signal);
      if (!requests.isCurrent(request)) return;
      clearInteraction(); setSelectedId(""); setHandoff(result);
      requests.broadcastChanged();
      void loadDirectory(true);
    } catch (error) {
      if (!requests.isCurrent(request)) return;
      clearInteraction(); setSelectedId(""); setFailure(workspaceInvitationFailure(error));
    } finally { if (requests.isCurrent(request)) setBusy(""); }
  }

  function startRevoke(invitation: WorkspaceInvitation) {
    requests.invalidate(); clearInteraction(); setReview({ kind: "revoke", invitation });
  }

  async function revokeInvitation() {
    if (review.kind !== "revoke" || !confirmed || busy || !canMutate) return;
    const request = requests.begin();
    setBusy("revoke"); setFailure(null); setConfirmed(false);
    try {
      await revokeWorkspaceInvitation(config, { invitation: review.invitation, confirmed: true }, request.signal);
      if (!requests.isCurrent(request)) return;
      clearInteraction(); setSelectedId("");
      requests.broadcastChanged();
      setNotice("Invitation revoked. Create a new invitation if access is still needed.");
      void loadDirectory(true);
    } catch (error) {
      if (!requests.isCurrent(request)) return;
      clearInteraction(); setSelectedId(""); setFailure(workspaceInvitationFailure(error));
    } finally { if (requests.isCurrent(request)) setBusy(""); }
  }

  async function copyCode() {
    if (!handoff) return;
    const request = requests.begin();
    try {
      await navigator.clipboard.writeText(handoff.invitationCode);
      if (requests.isCurrent(request)) setCopyStatus("copied");
    } catch { if (requests.isCurrent(request)) setCopyStatus("failed"); }
  }

  function cancelInteraction() {
    requests.invalidate(); clearInteraction(); setSelectedId("");
    if (directory.status === "loading") void loadDirectory();
  }

  return <section className="workspace-invitation-admin" aria-label="Workspace invitations">
    <div className="invitation-admin-workbench">
      <main className="invitation-directory">
        <header className="invitation-section-header"><div><h4>Invitation directory</h4><p>{directory.status === "ready"
          ? `${directory.page.invitations.length} loaded · as of ${invitationDate(directory.page.asOf)}` : "Workspace-scoped access intent"}</p></div>
          <div className="invitation-actions"><button type="button" onClick={() => void loadDirectory()} disabled={Boolean(busy)}>Refresh</button>
            <button type="button" className="invitation-primary" onClick={() => void startCreate()} disabled={Boolean(busy) || !canMutate || directory.status !== "ready"}>Create invitation</button></div>
        </header>
        <div className="invitation-filters" aria-label="Invitation state filter">
          {(["pending", "claimed", "expired", "revoked"] as const).map((state) => <button key={state} type="button" aria-pressed={filter === state}
            disabled={Boolean(busy) || directory.status === "loading"} onClick={() => { requests.invalidate(); clearInteraction(); setSelectedId(""); setFilter(state); }}>
            {state[0].toUpperCase() + state.slice(1)}</button>)}
        </div>
        {!canMutate ? <InvitationNotice title="Recent authentication required" tone="warning">Sign out and authenticate again before creating or revoking invitations.</InvitationNotice> : null}
        {failure ? <InvitationFailureNotice failure={failure} /> : null}
        {notice ? <p role="status" className="invitation-feedback">{notice}</p> : null}
        {directory.status === "loading" ? <p className="invitation-empty" role="status">Reading invitation directory…</p>
          : directory.status === "failed" ? <InvitationFailureNotice failure={directory.failure} />
          : <>
            {directory.page.invitations.length === 0 ? <p className="invitation-empty">No invitations in this state.</p> : <ul className="invitation-rows">
              {directory.page.invitations.map((invitation) => <li key={invitation.invitationId}>
                <button type="button" className={selectedId === invitation.invitationId ? "is-selected" : ""} aria-pressed={selectedId === invitation.invitationId}
                  disabled={Boolean(busy)} onClick={() => { requests.invalidate(); clearInteraction(); setSelectedId(invitation.invitationId); }}>
                  <div><code title={invitation.invitationId}>{invitationShortId(invitation.invitationId)}</code><InvitationStateBadge state={invitation.effectiveState} /></div>
                  <strong>{invitation.roleKey}</strong><span>Expires {invitationDate(invitation.expiresAt)} · v{invitation.recordVersion}</span>
                  <small>{selectedId === invitation.invitationId ? "SELECTED" : "View"}</small>
                </button>
              </li>)}
            </ul>}
            {directory.page.nextCursor ? <button type="button" onClick={() => void loadMore()} disabled={Boolean(busy)}>{busy === "page" ? "Loading…" : "Load more invitations"}</button> : null}
          </>}
        {selected ? <section className="invitation-selected" aria-label="Selected invitation">
          <header className="invitation-section-header"><h4>Selected {selected.effectiveState} invitation</h4>
            {selected.lifecycleState === "pending" ? <button type="button" className="invitation-danger" disabled={Boolean(busy) || !canMutate} onClick={() => startRevoke(selected)}>Review revoke</button> : null}</header>
          <InvitationFacts invitation={selected} />
          {selected.claimedByUserId ? <dl className="invitation-facts"><div><dt>Claimed by</dt><dd>{selected.claimedByUserId}</dd></div><div><dt>Membership</dt><dd>{selected.membershipId}</dd></div></dl> : null}
        </section> : null}
      </main>

      <aside className="invitation-review-rail" aria-label="Invitation review and one-time handoff">
        {handoff ? <>
          <header className="invitation-section-header"><h4>One-time code</h4><span className="invitation-state">VISIBLE ONCE</span></header>
          <ol className="invitation-progress"><li>✓ Review</li><li>✓ Create</li><li aria-current="step">Handoff</li></ol>
          <InvitationFacts invitation={handoff.invitation} />
          <label className="invitation-code-handoff"><span>CODE · VISIBLE ONCE</span><textarea aria-label="One-time invitation code" readOnly value={handoff.invitationCode} autoComplete="off" spellCheck={false} rows={3} />
            <button type="button" className="invitation-primary" onClick={() => void copyCode()} disabled={directory.status === "loading"}>{copyStatus === "copied" ? "Copied" : "Copy code"}</button></label>
          {copyStatus === "failed" ? <p role="alert">Copy failed. Select the visible code and copy it manually.</p> : <span className="invitation-feedback" role="status">{copyStatus === "copied" ? "Code copied to clipboard." : ""}</span>}
          <InvitationNotice title="Cannot be recovered" tone="warning">Leaving, switching actor or workspace, or choosing Done clears the code. Lists and refresh never return it.</InvitationNotice>
          <div className="invitation-actions"><button type="button" onClick={() => startRevoke(handoff.invitation)}>Revoke &amp; recreate</button>
            <button type="button" className="invitation-success" onClick={() => { cancelInteraction(); setNotice("Code cleared. Only invitation metadata remains available."); }}>Done &amp; clear code</button></div>
        </> : review.kind === "create" ? <>
          <header className="invitation-section-header"><h4>Role &amp; TTL review</h4><button type="button" onClick={cancelInteraction}>Cancel</button></header>
          {busy === "catalog" ? <p role="status">Reading the current role catalog…</p> : <form onSubmit={(event) => { event.preventDefault(); void createInvitation(); }} className="invitation-form">
            <dl className="invitation-facts"><div><dt>Workspace</dt><dd>{tenantRef} / {workspaceId}</dd></div></dl>
            <label>Role<select value={selectedRole} disabled={Boolean(busy)} onChange={(event) => { setSelectedRole(event.target.value); setConfirmed(false); }} required><option value="">Select a role</option>
              {roles.map((entry) => <option key={entry.roleKey} value={entry.roleKey}>{entry.displayName}</option>)}</select></label>
            <label>TTL<select value={ttlPolicy} disabled={Boolean(busy)} onChange={(event) => { setTTLPolicy(event.target.value as InvitationTTL); setConfirmed(false); }}>
              {INVITATION_TTLS.map((value) => <option key={value} value={value}>{TTL_LABELS[value]}</option>)}</select></label>
            {role ? <div className="invitation-reviewed-role"><strong>{role.roleKey}</strong><p>{role.summary}</p><small>{role.catalogVersion}</small><code>{role.definitionDigest}</code></div> : null}
            <label className="invitation-confirmation"><input type="checkbox" checked={confirmed} disabled={!role || Boolean(busy)} onChange={(event) => setConfirmed(event.target.checked)} />
              <span>I confirm this exact workspace, role and TTL. Access begins only after a signed-in claimant completes the claim.</span></label>
            <button type="submit" className="invitation-primary" disabled={!confirmed || !role || Boolean(busy) || !canMutate}>{busy === "create" ? "Creating…" : "Confirm & create invitation"}</button>
          </form>}
        </> : review.kind === "revoke" ? <>
          <header className="invitation-section-header"><h4>Review revoke</h4><button type="button" onClick={cancelInteraction}>Cancel</button></header>
          <code>{review.invitation.invitationId}</code><InvitationFacts invitation={review.invitation} />
          <InvitationNotice title="Withdraw this invitation" tone="warning">Revocation prevents future claim. To offer access again, create a new invitation after this operation completes.</InvitationNotice>
          <label className="invitation-confirmation"><input type="checkbox" checked={confirmed} disabled={Boolean(busy)} onChange={(event) => setConfirmed(event.target.checked)} /><span>I confirm revocation of this exact invitation at version {review.invitation.recordVersion}.</span></label>
          <button type="button" className="invitation-danger" disabled={!confirmed || Boolean(busy) || !canMutate} onClick={() => void revokeInvitation()}>{busy === "revoke" ? "Revoking…" : "Confirm revoke"}</button>
        </> : <InvitationNotice title="One-time code">Choose Create invitation to review a role and TTL, or select an existing record to review its state. Codes are shown only once after creation.</InvitationNotice>}
        <InvitationNotice title="Access begins after claim">Preview and code possession grant nothing. Membership and the catalog-derived role assignment take effect only after the full claim commits.</InvitationNotice>
      </aside>
    </div>
  </section>;
}
