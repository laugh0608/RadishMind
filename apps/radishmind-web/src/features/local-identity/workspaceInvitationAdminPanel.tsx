import { formatDisplayNumber } from "../../i18n/formatters.ts";
import { invitationDate } from "./workspaceInvitationMessages.ts";
import { identityRoleCopy } from "./localIdentityRoleMessages.ts";
import { useTranslation } from "react-i18next";
import { useLocalePreference } from "../../i18n/LocaleProvider.tsx";
import "../../i18n/identityInvitationResources.ts";
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
import { InvitationFacts, InvitationFailureNotice, InvitationNotice, InvitationStateBadge, invitationShortId } from "./workspaceInvitationView.tsx";

type DirectoryState = { status: "loading" } | { status: "ready"; page: WorkspaceInvitationPage } | { status: "failed"; failure: WorkspaceInvitationError };
type Review = { kind: "none" } | { kind: "create" } | { kind: "revoke"; invitation: WorkspaceInvitation };
type Handoff = { invitation: WorkspaceInvitation; invitationCode: string };

export function WorkspaceInvitationAdminPanel({ tenantRef, workspaceId }: { tenantRef: string; workspaceId: string }) {
  const { t } = useTranslation("identity");
  const identity = useLocalIdentity();
  if (!identity) return <InvitationNotice title={t($ => $.invitations.unavailable)}>{t($ => $.invitations.signInRequired)}</InvitationNotice>;
  const authorityKey = workspaceInvitationAuthorityKey(identity.profile, JSON.stringify([identity.config, tenantRef, workspaceId, "admin-invitations"]));
  return <WorkspaceInvitationAdministration key={`${authorityKey}:${identity.authorityRevision}`} identity={identity} authorityKey={authorityKey} tenantRef={tenantRef} workspaceId={workspaceId} />;
}

function WorkspaceInvitationAdministration({ identity, authorityKey, tenantRef, workspaceId }: {
  identity: LocalIdentityContextValue; authorityKey: string; tenantRef: string; workspaceId: string;
}) {
  const { t } = useTranslation("identity");
  const { locale } = useLocalePreference();
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
  const [notice, setNotice] = useState<"" | "revokedNotice" | "clearedNotice">("");
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
      setNotice("revokedNotice");
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

  return <section className="workspace-invitation-admin" aria-label={t($ => $.invitations.workspaceInvitations)}>
    <div className="invitation-admin-workbench">
      <main className="invitation-directory">
        <header className="invitation-section-header"><div><h4>{t($ => $.invitations.directory)}</h4><p>{directory.status === "ready"
          ? t($ => $.invitations.directorySummary, { countText: (formatDisplayNumber(directory.page.invitations.length, locale) ?? t($ => $.invitations.unknown)), date: invitationDate(t, directory.page.asOf, locale) }) : t($ => $.invitations.scopeIntent)}</p></div>
          <div className="invitation-actions"><button type="button" onClick={() => void loadDirectory()} disabled={Boolean(busy)}>{t($ => $.invitations.refresh)}</button>
            <button type="button" className="invitation-primary" onClick={() => void startCreate()} disabled={Boolean(busy) || !canMutate || directory.status !== "ready"}>{t($ => $.invitations.create)}</button></div>
        </header>
        <div className="invitation-filters" aria-label={t($ => $.invitations.stateFilter)}>
          {(["pending", "claimed", "expired", "revoked"] as const).map((state) => <button key={state} type="button" aria-pressed={filter === state}
            disabled={Boolean(busy) || directory.status === "loading"} onClick={() => { requests.invalidate(); clearInteraction(); setSelectedId(""); setFilter(state); }}>
            {t($ => $.invitations.states[state])}</button>)}
        </div>
        {!canMutate ? <InvitationNotice title={t($ => $.invitations.recentAuth)} tone="warning">{t($ => $.invitations.adminReauth)}</InvitationNotice> : null}
        {failure ? <InvitationFailureNotice failure={failure} /> : null}
        {notice ? <p role="status" className="invitation-feedback">{t($ => $.invitations[notice])}</p> : null}
        {directory.status === "loading" ? <p className="invitation-empty" role="status">{t($ => $.invitations.readingDirectory)}</p>
          : directory.status === "failed" ? <InvitationFailureNotice failure={directory.failure} />
          : <>
            {directory.page.invitations.length === 0 ? <p className="invitation-empty">{t($ => $.invitations.empty)}</p> : <ul className="invitation-rows">
              {directory.page.invitations.map((invitation) => <li key={invitation.invitationId}>
                <button type="button" className={selectedId === invitation.invitationId ? "is-selected" : ""} aria-pressed={selectedId === invitation.invitationId}
                  disabled={Boolean(busy)} onClick={() => { requests.invalidate(); clearInteraction(); setSelectedId(invitation.invitationId); }}>
                  <div><code title={invitation.invitationId}>{invitationShortId(invitation.invitationId)}</code><InvitationStateBadge state={invitation.effectiveState} /></div>
                  <strong>{invitation.roleKey}</strong><span>{t($ => $.invitations.expiryVersion, { date: invitationDate(t, invitation.expiresAt, locale), version: invitation.recordVersion })}</span>
                  <small>{selectedId === invitation.invitationId ? t($ => $.invitations.selected) : t($ => $.invitations.view)}</small>
                </button>
              </li>)}
            </ul>}
            {directory.page.nextCursor ? <button type="button" onClick={() => void loadMore()} disabled={Boolean(busy)}>{busy === "page" ? t($ => $.invitations.loading) : t($ => $.invitations.loadMore)}</button> : null}
          </>}
        {selected ? <section className="invitation-selected" aria-label={t($ => $.invitations.selectedInvitation)}>
          <header className="invitation-section-header"><h4>{t($ => $.invitations.selectedState, { state: t($ => $.invitations.states[selected.effectiveState]) })}</h4>
            {selected.lifecycleState === "pending" ? <button type="button" className="invitation-danger" disabled={Boolean(busy) || !canMutate} onClick={() => startRevoke(selected)}>{t($ => $.invitations.reviewRevoke)}</button> : null}</header>
          <InvitationFacts invitation={selected} />
          {selected.claimedByUserId ? <dl className="invitation-facts"><div><dt>{t($ => $.invitations.claimedBy)}</dt><dd>{selected.claimedByUserId}</dd></div><div><dt>{t($ => $.invitations.membership)}</dt><dd>{selected.membershipId}</dd></div></dl> : null}
        </section> : null}
      </main>

      <aside className="invitation-review-rail" aria-label={t($ => $.invitations.reviewAndHandoff)}>
        {handoff ? <>
          <header className="invitation-section-header"><h4>{t($ => $.invitations.oneTimeCode)}</h4><span className="invitation-state">{t($ => $.invitations.visibleOnce)}</span></header>
          <ol className="invitation-progress"><li>{t($ => $.invitations.reviewComplete)}</li><li>{t($ => $.invitations.createComplete)}</li><li aria-current="step">{t($ => $.invitations.handoff)}</li></ol>
          <InvitationFacts invitation={handoff.invitation} />
          <label className="invitation-code-handoff"><span>{t($ => $.invitations.codeVisibleOnce)}</span><textarea aria-label={t($ => $.invitations.oneTimeCodeLabel)} readOnly value={handoff.invitationCode} autoComplete="off" spellCheck={false} rows={3} />
            <button type="button" className="invitation-primary" onClick={() => void copyCode()} disabled={directory.status === "loading"}>{copyStatus === "copied" ? t($ => $.invitations.copied) : t($ => $.invitations.copyCode)}</button></label>
          {copyStatus === "failed" ? <p role="alert">{t($ => $.invitations.copyFailed)}</p> : <span className="invitation-feedback" role="status">{copyStatus === "copied" ? t($ => $.invitations.copySuccess) : ""}</span>}
          <InvitationNotice title={t($ => $.invitations.unrecoverable)} tone="warning">{t($ => $.invitations.clearWarning)}</InvitationNotice>
          <div className="invitation-actions"><button type="button" onClick={() => startRevoke(handoff.invitation)}>{t($ => $.invitations.revokeRecreate)}</button>
            <button type="button" className="invitation-success" onClick={() => { cancelInteraction(); setNotice("clearedNotice"); }}>{t($ => $.invitations.doneClear)}</button></div>
        </> : review.kind === "create" ? <>
          <header className="invitation-section-header"><h4>{t($ => $.invitations.roleTtlReview)}</h4><button type="button" onClick={cancelInteraction}>{t($ => $.invitations.cancel)}</button></header>
          {busy === "catalog" ? <p role="status">{t($ => $.invitations.readingCatalog)}</p> : <form onSubmit={(event) => { event.preventDefault(); void createInvitation(); }} className="invitation-form">
            <dl className="invitation-facts"><div><dt>{t($ => $.invitations.workspace)}</dt><dd>{tenantRef} / {workspaceId}</dd></div></dl>
            <label>{t($ => $.invitations.role)}<select value={selectedRole} disabled={Boolean(busy)} onChange={(event) => { setSelectedRole(event.target.value); setConfirmed(false); }} required><option value="">{t($ => $.invitations.selectRole)}</option>
              {roles.map((entry) => <option key={entry.roleKey} value={entry.roleKey}>{identityRoleCopy(t, entry.roleKey).name}</option>)}</select></label>
            <label>{t($ => $.invitations.ttl)}<select value={ttlPolicy} disabled={Boolean(busy)} onChange={(event) => { setTTLPolicy(event.target.value as InvitationTTL); setConfirmed(false); }}>
              {INVITATION_TTLS.map((value) => <option key={value} value={value}>{t($ => $.invitations.ttls[value])}</option>)}</select></label>
            {role ? <div className="invitation-reviewed-role"><strong>{role.roleKey}</strong><p>{identityRoleCopy(t, role.roleKey).summary}</p><small>{role.catalogVersion}</small><code>{role.definitionDigest}</code></div> : null}
            <label className="invitation-confirmation"><input type="checkbox" checked={confirmed} disabled={!role || Boolean(busy)} onChange={(event) => setConfirmed(event.target.checked)} />
              <span>{t($ => $.invitations.createConfirmation)}</span></label>
            <button type="submit" className="invitation-primary" disabled={!confirmed || !role || Boolean(busy) || !canMutate}>{busy === "create" ? t($ => $.invitations.creating) : t($ => $.invitations.confirmCreate)}</button>
          </form>}
        </> : review.kind === "revoke" ? <>
          <header className="invitation-section-header"><h4>{t($ => $.invitations.reviewRevoke)}</h4><button type="button" onClick={cancelInteraction}>{t($ => $.invitations.cancel)}</button></header>
          <code>{review.invitation.invitationId}</code><InvitationFacts invitation={review.invitation} />
          <InvitationNotice title={t($ => $.invitations.withdraw)} tone="warning">{t($ => $.invitations.withdrawHelp)}</InvitationNotice>
          <label className="invitation-confirmation"><input type="checkbox" checked={confirmed} disabled={Boolean(busy)} onChange={(event) => setConfirmed(event.target.checked)} /><span>{t($ => $.invitations.revokeConfirmation, { version: review.invitation.recordVersion })}</span></label>
          <button type="button" className="invitation-danger" disabled={!confirmed || Boolean(busy) || !canMutate} onClick={() => void revokeInvitation()}>{busy === "revoke" ? t($ => $.invitations.revoking) : t($ => $.invitations.confirmRevoke)}</button>
        </> : <InvitationNotice title={t($ => $.invitations.oneTimeCode)}>{t($ => $.invitations.chooseHelp)}</InvitationNotice>}
        <InvitationNotice title={t($ => $.invitations.accessAfterClaim)}>{t($ => $.invitations.accessAfterClaimHelp)}</InvitationNotice>
      </aside>
    </div>
  </section>;
}
