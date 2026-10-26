import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { useLocalIdentity } from "../local-identity/localIdentityContext.ts";
import { availableIdentityWorkspaces } from "../local-identity/localIdentityWorkspaceAccess.ts";
import { workbenchConfiguration } from "../../runtime/workbenchConfiguration.ts";

export function LocalWorkbenchBoundary({ children }: { children: ReactNode }) {
  const { t } = useTranslation("shell");
  const identity = useLocalIdentity();
  const config = workbenchConfiguration();
  if (!config) return <>{children}</>;
  const hasMembership = identity && availableIdentityWorkspaces(identity.profile).some(membership =>
    membership.tenantRef === config.tenant_ref && membership.workspaceId === config.workspace_id);
  if (hasMembership) return <>{children}</>;
  return <main className="surface-band workbench-first-entry" aria-labelledby="workbench-first-entry-title">
    <h1 id="workbench-first-entry-title">{t($ => $.localPersistence.joinTitle)}</h1>
    <p>{t($ => $.localPersistence.joinDescription)}</p>
    <dl>
      <div><dt>{t($ => $.localPersistence.account)}</dt><dd><code>{identity?.profile.account.userId ?? "—"}</code></dd></div>
      <div><dt>{t($ => $.localPersistence.scope)}</dt><dd>{config.tenant_ref} / {config.workspace_id}</dd></div>
    </dl>
    <p>{t($ => $.localPersistence.bootstrapSteps)}</p>
    <pre><code>./start.sh workbench bootstrap-admin --config var/workbench/platform.json --user-id {identity?.profile.account.userId ?? "<user_id>"} --audit-ref initial-workspace-admin</code></pre>
    <p>{t($ => $.localPersistence.bootstrapVariants)}</p>
    <button type="button" onClick={() => { void identity?.refresh(); }}>{t($ => $.localPersistence.refreshMembership)}</button>
  </main>;
}

export function LocalWorkbenchStatus() {
  const { t } = useTranslation("shell");
  const config = workbenchConfiguration();
  if (!config) return null;
  return <div className="workbench-runtime-status" role="status">
    <strong>{t($ => $.localPersistence.mode)}</strong>
    <span>{config.provider_mode === "mock" ? t($ => $.localPersistence.mock) : t($ => $.localPersistence.remote)}</span>
    <span>{t($ => $.localPersistence.retention)}</span>
  </div>;
}
