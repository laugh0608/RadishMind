import { useState } from "react";

import { useLocalIdentity } from "./localIdentityGateway.tsx";

export function AdminLocalIdentityOwner({ surface }: { surface: "user" | "role" }) {
  const identity = useLocalIdentity();
  const [pendingBindingId, setPendingBindingId] = useState("");
  const [busyBindingId, setBusyBindingId] = useState("");
  const [failure, setFailure] = useState("");

  if (!identity) return <DisabledLocalIdentityOwner surface={surface} />;
  const { profile } = identity;
  const userSurface = surface === "user";

  async function revokeBinding(bindingId: string, recordVersion: number) {
    if (pendingBindingId !== bindingId) {
      setPendingBindingId(bindingId);
      setFailure("");
      return;
    }
    setBusyBindingId(bindingId);
    setFailure("");
    try {
      await identity?.revokeExternalIdentity(bindingId, recordVersion);
      setPendingBindingId("");
    } catch (error) {
      setFailure(error instanceof Error ? error.message : "External identity could not be revoked.");
    } finally {
      setBusyBindingId("");
    }
  }

  return (
    <section className="admin-control-owner-surface admin-control-identity-owner local-identity-admin-owner" aria-labelledby="admin-identity-owner-title">
      <header>
        <div>
          <p className="eyebrow">{userSurface ? "User" : "Role"} · local identity owner</p>
          <h4 id="admin-identity-owner-title">{userSurface ? "Current local account" : "Current local grants"}</h4>
        </div>
        <span className="admin-control-status is-ready">development/test</span>
      </header>

      {userSurface ? (
        <>
          <div className="admin-control-identity-statement local-identity-admin-statement">
            <span aria-hidden="true">{profile.account.displayName.slice(0, 1).toUpperCase()}</span>
            <div>
              <strong>{profile.account.displayName}</strong>
              <p>{profile.account.userId} · {profile.account.lifecycleState}</p>
            </div>
          </div>
          <dl className="admin-control-owner-meta admin-control-identity-meta">
            <div><dt>Truth owner</dt><dd>RadishMind Local Identity Repository</dd></div>
            <div><dt>Session method</dt><dd>{profile.session.authenticationMethod}</dd></div>
            <div><dt>Local credential</dt><dd>{profile.capabilities.hasActiveLocalCredential ? "active" : "not present"}</dd></div>
            <div><dt>Recent authentication</dt><dd>{profile.capabilities.recentAuthentication ? "verified" : "required"}</dd></div>
          </dl>
          <div className="local-identity-admin-list" aria-label="External login methods">
            <header><strong>External login methods</strong><small>Sanitized provider references only</small></header>
            {profile.externalIdentities.length === 0 ? (
              <p className="local-identity-admin-empty">No external identity is bound to this local account.</p>
            ) : profile.externalIdentities.map((binding) => (
              <article key={binding.bindingId}>
                <div>
                  <strong>{binding.providerRef}</strong>
                  <small>{binding.bindingId} · version {binding.recordVersion}</small>
                </div>
                <span className={`status-badge ${binding.lifecycleState === "active" ? "good" : "neutral"}`}>
                  {binding.lifecycleState}
                </span>
                {binding.lifecycleState === "active" ? (
                  <button
                    type="button"
                    className={pendingBindingId === binding.bindingId ? "danger-action" : "secondary-action"}
                    disabled={!binding.canRevoke || busyBindingId !== "" || !profile.capabilities.recentAuthentication}
                    onClick={() => void revokeBinding(binding.bindingId, binding.recordVersion)}
                  >
                    {busyBindingId === binding.bindingId ? "Revoking…" :
                      !binding.canRevoke ? "Last login method" :
                      pendingBindingId === binding.bindingId ? "Confirm revoke" : "Revoke"}
                  </button>
                ) : null}
              </article>
            ))}
          </div>
          {failure ? <p className="local-identity-inline-error" role="alert">{failure}</p> : null}
          <div className="admin-control-blocked-actions" aria-label="Identity boundary actions">
            <button
              type="button"
              onClick={() => void identity.linkOIDC()}
              disabled={!profile.capabilities.oidcEnabled || !profile.capabilities.recentAuthentication}
            >
              Link reviewed Radish OIDC
            </button>
            <span>Account directory list not exposed</span>
            <span>Invite / onboarding closed</span>
            <span>Automated recovery closed</span>
          </div>
        </>
      ) : (
        <>
          <dl className="admin-control-owner-meta admin-control-identity-meta">
            <div><dt>Account</dt><dd>{profile.account.userId}</dd></div>
            <div><dt>Assignments</dt><dd>{profile.roleAssignments.length}</dd></div>
            <div><dt>Memberships</dt><dd>{profile.workspaceMemberships.length}</dd></div>
            <div><dt>Mutation owner</dt><dd>repository only; no Web mutation route</dd></div>
          </dl>
          <div className="local-identity-admin-split">
            <div className="local-identity-admin-list" aria-label="Local role assignments">
              <header><strong>Role assignments</strong><small>Local permission grants</small></header>
              {profile.roleAssignments.length === 0 ? (
                <p className="local-identity-admin-empty">No local role assignment is recorded for this account.</p>
              ) : profile.roleAssignments.map((assignment) => (
                <article key={assignment.assignmentId}>
                  <div>
                    <strong>{assignment.roleKey}</strong>
                    <small>{assignment.tenantRef} / {assignment.workspaceId || "tenant scope"}</small>
                  </div>
                  <span className={`status-badge ${assignment.lifecycleState === "active" ? "good" : "neutral"}`}>
                    {assignment.lifecycleState}
                  </span>
                  <ul>{assignment.permissionGrants.map((grant) => <li key={grant}>{grant}</li>)}</ul>
                </article>
              ))}
            </div>
            <div className="local-identity-admin-list" aria-label="Local workspace memberships">
              <header><strong>Workspace memberships</strong><small>Explicit scope records</small></header>
              {profile.workspaceMemberships.length === 0 ? (
                <p className="local-identity-admin-empty">No local workspace membership is recorded for this account.</p>
              ) : profile.workspaceMemberships.map((membership) => (
                <article key={membership.membershipId}>
                  <div>
                    <strong>{membership.workspaceId}</strong>
                    <small>{membership.tenantRef} · version {membership.recordVersion}</small>
                  </div>
                  <span className={`status-badge ${membership.lifecycleState === "active" ? "good" : "neutral"}`}>
                    {membership.lifecycleState}
                  </span>
                </article>
              ))}
            </div>
          </div>
          <div className="admin-control-blocked-actions" aria-label="Blocked authorization actions">
            <span>Assign or revoke role in Web</span>
            <span>Infer grants from role name</span>
            <span>Import upstream claims</span>
            <span>Declare production authorization</span>
          </div>
        </>
      )}

      <div className="boundary-note">
        This surface is a current-account projection, not a directory. OIDC issuer/subject, raw claims, credentials,
        cookie values and audit internals are intentionally absent.
      </div>
    </section>
  );
}

function DisabledLocalIdentityOwner({ surface }: { surface: "user" | "role" }) {
  return (
    <section className="admin-control-owner-surface admin-control-identity-owner" aria-labelledby="admin-identity-owner-title">
      <header>
        <div>
          <p className="eyebrow">{surface === "user" ? "User" : "Role"} · local identity boundary</p>
          <h4 id="admin-identity-owner-title">Local identity Web consumer disabled</h4>
        </div>
        <span className="admin-control-status is-blocked">not connected</span>
      </header>
      <div className="admin-control-identity-statement">
        <span aria-hidden="true">∅</span>
        <div>
          <strong>No current-account facts are available in offline mode</strong>
          <p>Enable the reviewed development/test local identity mode to consume the repository-owned account projection.</p>
        </div>
      </div>
      <div className="boundary-note">Offline fixtures and development headers cannot be promoted into user, membership or grant facts.</div>
    </section>
  );
}
